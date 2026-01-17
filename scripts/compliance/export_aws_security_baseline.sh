#!/bin/bash
# export_aws_security_baseline.sh
#
# SOC 2 Evidence Collection Script - AWS Security Baseline
# Exports security-relevant configuration from AWS account
#
# Evidence Maps to:
# - CC6.1: Logical access security (IAM)
# - CC6.6: Authentication (MFA, password policy)
# - CC7.2: System monitoring (CloudTrail, Config)
# - A1.3: Backup/recovery (backup status)
#
# Usage: ./export_aws_security_baseline.sh [output_dir]
#
# Prerequisites:
# - AWS CLI installed and configured
# - Appropriate IAM permissions for read operations
# - jq installed

set -euo pipefail

# Configuration
AWS_PROFILE="${AWS_PROFILE:-default}"
AWS_REGION="${AWS_REGION:-us-east-1}"
OUTPUT_DIR="${1:-./compliance-evidence/soc2/$(date +%Y)/$(date +%Y-%m)/access-reviews}"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
DATE_STAMP=$(date +"%Y-%m-%d")

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check prerequisites
check_prereqs() {
    local missing=()

    if ! command -v aws &> /dev/null; then
        missing+=("aws CLI")
    fi

    if ! command -v jq &> /dev/null; then
        missing+=("jq")
    fi

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing prerequisites: ${missing[*]}"
        echo "Install with:"
        echo "  brew install awscli jq"
        exit 1
    fi

    # Check AWS credentials
    if ! aws sts get-caller-identity --profile "${AWS_PROFILE}" &> /dev/null; then
        log_error "AWS credentials not configured or expired"
        echo "Run: aws configure --profile ${AWS_PROFILE}"
        echo "Or: aws sso login --profile ${AWS_PROFILE}"
        exit 1
    fi
}

# Create output directory
setup_output() {
    mkdir -p "${OUTPUT_DIR}"
    log_info "Output directory: ${OUTPUT_DIR}"
}

# Export account info
export_account_info() {
    log_info "Exporting account information..."

    aws sts get-caller-identity --profile "${AWS_PROFILE}" | jq '{
        account_id: .Account,
        user_arn: .Arn,
        export_timestamp: "'"${TIMESTAMP}"'"
    }' > "${OUTPUT_DIR}/${DATE_STAMP}_account_info.json"

    # Get account alias if set
    aws iam list-account-aliases --profile "${AWS_PROFILE}" 2>/dev/null | jq '{
        aliases: .AccountAliases
    }' >> "${OUTPUT_DIR}/${DATE_STAMP}_account_info.json" || true

    log_info "  ✓ Account info exported"
}

# Export IAM password policy
export_password_policy() {
    log_info "Exporting IAM password policy..."

    aws iam get-account-password-policy --profile "${AWS_PROFILE}" 2>/dev/null | jq '.PasswordPolicy' \
        > "${OUTPUT_DIR}/${DATE_STAMP}_iam_password_policy.json" || {
        log_warn "  No custom password policy configured"
        echo '{"status": "DEFAULT_POLICY", "note": "Using AWS default password policy"}' \
            > "${OUTPUT_DIR}/${DATE_STAMP}_iam_password_policy.json"
    }

    log_info "  ✓ Password policy exported"
}

# Export IAM users (no secrets - just metadata)
export_iam_users() {
    log_info "Exporting IAM users..."

    aws iam list-users --profile "${AWS_PROFILE}" | jq '[.Users[] | {
        user_name: .UserName,
        user_id: .UserId,
        arn: .Arn,
        create_date: .CreateDate,
        password_last_used: .PasswordLastUsed
    }]' > "${OUTPUT_DIR}/${DATE_STAMP}_iam_users.json"

    # Get MFA status for each user
    log_info "  Checking MFA status for users..."
    local users_file="${OUTPUT_DIR}/${DATE_STAMP}_iam_users.json"
    local mfa_file="${OUTPUT_DIR}/${DATE_STAMP}_iam_mfa_status.json"

    echo "[" > "${mfa_file}"
    local first=true

    for username in $(jq -r '.[].user_name' "${users_file}"); do
        if [[ "${first}" != "true" ]]; then
            echo "," >> "${mfa_file}"
        fi
        first=false

        local mfa_devices
        mfa_devices=$(aws iam list-mfa-devices --user-name "${username}" --profile "${AWS_PROFILE}" 2>/dev/null | jq '.MFADevices | length')

        echo "  {\"user_name\": \"${username}\", \"mfa_enabled\": $([ "${mfa_devices}" -gt 0 ] && echo "true" || echo "false"), \"mfa_device_count\": ${mfa_devices}}" >> "${mfa_file}"
    done

    echo "]" >> "${mfa_file}"

    log_info "  ✓ IAM users and MFA status exported"
}

# Export IAM roles
export_iam_roles() {
    log_info "Exporting IAM roles..."

    aws iam list-roles --profile "${AWS_PROFILE}" | jq '[.Roles[] | {
        role_name: .RoleName,
        role_id: .RoleId,
        arn: .Arn,
        create_date: .CreateDate,
        max_session_duration: .MaxSessionDuration,
        path: .Path
    }] | [.[] | select(.path != "/aws-service-role/")]' \
        > "${OUTPUT_DIR}/${DATE_STAMP}_iam_roles.json"

    log_info "  ✓ IAM roles exported"
}

# Export access key age report
export_access_key_report() {
    log_info "Generating access key age report..."

    local report_file="${OUTPUT_DIR}/${DATE_STAMP}_access_key_age.json"
    echo "[" > "${report_file}"
    local first=true

    for username in $(aws iam list-users --profile "${AWS_PROFILE}" --query 'Users[].UserName' --output text); do
        local keys
        keys=$(aws iam list-access-keys --user-name "${username}" --profile "${AWS_PROFILE}" 2>/dev/null)

        for key_id in $(echo "${keys}" | jq -r '.AccessKeyMetadata[].AccessKeyId'); do
            if [[ "${first}" != "true" ]]; then
                echo "," >> "${report_file}"
            fi
            first=false

            local key_info
            key_info=$(echo "${keys}" | jq ".AccessKeyMetadata[] | select(.AccessKeyId == \"${key_id}\")")

            local create_date
            create_date=$(echo "${key_info}" | jq -r '.CreateDate')

            # Calculate age in days
            local age_days=0
            if [[ -n "${create_date}" && "${create_date}" != "null" ]]; then
                local create_epoch
                create_epoch=$(date -d "${create_date}" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%S" "${create_date%%+*}" +%s 2>/dev/null || echo "0")
                local now_epoch
                now_epoch=$(date +%s)
                age_days=$(( (now_epoch - create_epoch) / 86400 ))
            fi

            echo "  {\"user_name\": \"${username}\", \"access_key_id\": \"${key_id}\", \"status\": $(echo "${key_info}" | jq '.Status'), \"create_date\": \"${create_date}\", \"age_days\": ${age_days}}" >> "${report_file}"
        done
    done

    echo "]" >> "${report_file}"

    log_info "  ✓ Access key age report exported"
}

# Check root account MFA (requires specific permissions)
export_root_mfa_status() {
    log_info "Checking root account MFA status..."

    # This requires iam:GetAccountSummary permission
    aws iam get-account-summary --profile "${AWS_PROFILE}" 2>/dev/null | jq '{
        account_mfa_enabled: .SummaryMap.AccountMFAEnabled,
        users: .SummaryMap.Users,
        groups: .SummaryMap.Groups,
        roles: .SummaryMap.Roles,
        mfa_devices: .SummaryMap.MFADevices,
        access_keys_per_user_quota: .SummaryMap.AccessKeysPerUserQuota
    }' > "${OUTPUT_DIR}/${DATE_STAMP}_account_summary.json" || {
        log_warn "  Could not get account summary (permissions issue)"
        echo '{"status": "NO_ACCESS"}' > "${OUTPUT_DIR}/${DATE_STAMP}_account_summary.json"
    }

    log_info "  ✓ Account summary exported"
}

# Export CloudTrail status
export_cloudtrail_status() {
    log_info "Checking CloudTrail configuration..."

    aws cloudtrail describe-trails --profile "${AWS_PROFILE}" --region "${AWS_REGION}" 2>/dev/null | jq '[.trailList[] | {
        name: .Name,
        s3_bucket: .S3BucketName,
        is_multi_region: .IsMultiRegionTrail,
        is_organization_trail: .IsOrganizationTrail,
        include_global_service_events: .IncludeGlobalServiceEvents,
        log_file_validation_enabled: .LogFileValidationEnabled,
        kms_key_id: .KmsKeyId,
        home_region: .HomeRegion
    }]' > "${OUTPUT_DIR}/${DATE_STAMP}_cloudtrail_config.json" || {
        log_warn "  Could not get CloudTrail config"
        echo '[]' > "${OUTPUT_DIR}/${DATE_STAMP}_cloudtrail_config.json"
    }

    # Check trail status
    for trail in $(aws cloudtrail describe-trails --profile "${AWS_PROFILE}" --region "${AWS_REGION}" --query 'trailList[].Name' --output text 2>/dev/null); do
        aws cloudtrail get-trail-status --name "${trail}" --profile "${AWS_PROFILE}" --region "${AWS_REGION}" 2>/dev/null | jq '{
            trail_name: "'"${trail}"'",
            is_logging: .IsLogging,
            latest_delivery_time: .LatestDeliveryTime,
            latest_delivery_error: .LatestDeliveryError,
            start_logging_time: .StartLoggingTime
        }' >> "${OUTPUT_DIR}/${DATE_STAMP}_cloudtrail_status.json" || true
    done

    log_info "  ✓ CloudTrail configuration exported"
}

# Export AWS Config status
export_config_status() {
    log_info "Checking AWS Config status..."

    aws configservice describe-configuration-recorders --profile "${AWS_PROFILE}" --region "${AWS_REGION}" 2>/dev/null | jq '.ConfigurationRecorders' \
        > "${OUTPUT_DIR}/${DATE_STAMP}_config_recorders.json" || {
        log_warn "  AWS Config not enabled or no access"
        echo '[]' > "${OUTPUT_DIR}/${DATE_STAMP}_config_recorders.json"
    }

    aws configservice describe-configuration-recorder-status --profile "${AWS_PROFILE}" --region "${AWS_REGION}" 2>/dev/null | jq '.ConfigurationRecordersStatus' \
        >> "${OUTPUT_DIR}/${DATE_STAMP}_config_status.json" || true

    log_info "  ✓ AWS Config status exported"
}

# Export Security Hub status (if enabled)
export_security_hub_status() {
    log_info "Checking Security Hub status..."

    aws securityhub describe-hub --profile "${AWS_PROFILE}" --region "${AWS_REGION}" 2>/dev/null | jq '{
        hub_arn: .HubArn,
        subscribed_at: .SubscribedAt,
        auto_enable_controls: .AutoEnableControls
    }' > "${OUTPUT_DIR}/${DATE_STAMP}_security_hub.json" || {
        log_warn "  Security Hub not enabled"
        echo '{"status": "NOT_ENABLED"}' > "${OUTPUT_DIR}/${DATE_STAMP}_security_hub.json"
    }

    log_info "  ✓ Security Hub status exported"
}

# Export GuardDuty status
export_guardduty_status() {
    log_info "Checking GuardDuty status..."

    local detector_ids
    detector_ids=$(aws guardduty list-detectors --profile "${AWS_PROFILE}" --region "${AWS_REGION}" --query 'DetectorIds' --output text 2>/dev/null || echo "")

    if [[ -n "${detector_ids}" && "${detector_ids}" != "None" ]]; then
        for detector_id in ${detector_ids}; do
            aws guardduty get-detector --detector-id "${detector_id}" --profile "${AWS_PROFILE}" --region "${AWS_REGION}" 2>/dev/null | jq '{
                detector_id: "'"${detector_id}"'",
                status: .Status,
                finding_publishing_frequency: .FindingPublishingFrequency,
                service_role: .ServiceRole
            }' > "${OUTPUT_DIR}/${DATE_STAMP}_guardduty.json"
        done
    else
        log_warn "  GuardDuty not enabled"
        echo '{"status": "NOT_ENABLED"}' > "${OUTPUT_DIR}/${DATE_STAMP}_guardduty.json"
    fi

    log_info "  ✓ GuardDuty status exported"
}

# Generate summary report
generate_summary() {
    log_info "Generating summary report..."

    local summary_file="${OUTPUT_DIR}/${DATE_STAMP}_aws_security_summary.json"

    # Count findings
    local users_count mfa_enabled_count root_mfa cloudtrail_enabled config_enabled

    users_count=$(jq 'length' "${OUTPUT_DIR}/${DATE_STAMP}_iam_users.json" 2>/dev/null || echo "0")
    mfa_enabled_count=$(jq '[.[] | select(.mfa_enabled == true)] | length' "${OUTPUT_DIR}/${DATE_STAMP}_iam_mfa_status.json" 2>/dev/null || echo "0")
    root_mfa=$(jq '.account_mfa_enabled' "${OUTPUT_DIR}/${DATE_STAMP}_account_summary.json" 2>/dev/null || echo "unknown")
    cloudtrail_enabled=$(jq 'length > 0' "${OUTPUT_DIR}/${DATE_STAMP}_cloudtrail_config.json" 2>/dev/null || echo "false")
    config_enabled=$(jq 'length > 0' "${OUTPUT_DIR}/${DATE_STAMP}_config_recorders.json" 2>/dev/null || echo "false")

    cat > "${summary_file}" << EOF
{
    "export_timestamp": "${TIMESTAMP}",
    "aws_region": "${AWS_REGION}",
    "evidence_type": "aws_security_baseline",
    "soc2_controls": ["CC6.1", "CC6.6", "CC7.2", "A1.3"],
    "summary": {
        "iam_users_count": ${users_count},
        "mfa_enabled_users": ${mfa_enabled_count},
        "root_mfa_enabled": ${root_mfa},
        "cloudtrail_enabled": ${cloudtrail_enabled},
        "aws_config_enabled": ${config_enabled}
    },
    "files_generated": $(find "${OUTPUT_DIR}" -name "${DATE_STAMP}*.json" | wc -l | tr -d ' '),
    "recommendations": [
        $([ "${root_mfa}" != "1" ] && echo '"Enable MFA on root account",' || echo "")
        $([ "${users_count}" -gt "${mfa_enabled_count}" ] && echo '"Enable MFA for all IAM users",' || echo "")
        $([ "${cloudtrail_enabled}" == "false" ] && echo '"Enable CloudTrail",' || echo "")
        $([ "${config_enabled}" == "false" ] && echo '"Enable AWS Config",' || echo "")
        "Review access key age report for keys > 90 days"
    ]
}
EOF

    log_info "  ✓ Summary: ${summary_file}"
}

# Main
main() {
    log_info "=== AWS Security Baseline Export ==="
    log_info "Timestamp: ${TIMESTAMP}"
    log_info "Profile: ${AWS_PROFILE}"
    log_info "Region: ${AWS_REGION}"

    check_prereqs
    setup_output
    export_account_info
    export_password_policy
    export_iam_users
    export_iam_roles
    export_access_key_report
    export_root_mfa_status
    export_cloudtrail_status
    export_config_status
    export_security_hub_status
    export_guardduty_status
    generate_summary

    echo ""
    log_info "=== Export Complete ==="
    log_info "Output: ${OUTPUT_DIR}"
    log_info "Files generated:"
    find "${OUTPUT_DIR}" -name "${DATE_STAMP}*.json" -exec basename {} \; | sort

    echo ""
    log_info "Review ${OUTPUT_DIR}/${DATE_STAMP}_aws_security_summary.json for recommendations"
}

main "$@"
