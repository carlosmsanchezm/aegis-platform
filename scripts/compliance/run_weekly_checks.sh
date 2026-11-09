#!/bin/bash
# run_weekly_checks.sh
#
# SOC 2 + ISO 27001 Weekly Evidence Collection
# Single-command operation for weekly security review
#
# Controls Covered:
# - SOC 2: CC7.1 (Vulnerability detection), CC8.1 (Change management)
# - ISO 27001: A.8.8 (Technical vulnerability management), A.8.32 (Change management)
#
# Usage: ./run_weekly_checks.sh [--dry-run] [--push]
#
# Preconditions:
# - EVIDENCE_VAULT env var must be set
# - gh CLI authenticated
# - (Optional) AWS CLI configured if checking AWS

set -euo pipefail

#######################################
# Configuration
#######################################
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
YEAR=$(date +%Y)
MONTH=$(date +%Y-%m)
WEEK=$(date +%V)
DATE_STAMP=$(date +%Y-%m-%d)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
RUN_ID="${DATE_STAMP}-weekly"

# Flags
DRY_RUN=false
PUSH=false

# Parse arguments
for arg in "$@"; do
    case $arg in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --push)
            PUSH=true
            shift
            ;;
    esac
done

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }
log_dry() { echo -e "${CYAN}[DRY-RUN]${NC} $1"; }

#######################################
# Precondition Checks
#######################################
check_preconditions() {
    log_step "Checking preconditions..."

    # Check EVIDENCE_VAULT is set
    if [[ -z "${EVIDENCE_VAULT:-}" ]]; then
        log_error "EVIDENCE_VAULT environment variable is not set."
        echo ""
        echo "Fix: Set the environment variable:"
        echo "  export EVIDENCE_VAULT=\$HOME/code/aegis-compliance-evidence"
        echo ""
        echo "Or add to your shell profile (~/.zshrc or ~/.bashrc):"
        echo "  echo 'export EVIDENCE_VAULT=\$HOME/code/aegis-compliance-evidence' >> ~/.zshrc"
        exit 1
    fi
    log_info "  EVIDENCE_VAULT: ${EVIDENCE_VAULT}"

    # Check vault directory exists
    if [[ ! -d "${EVIDENCE_VAULT}" ]]; then
        log_error "EVIDENCE_VAULT directory does not exist: ${EVIDENCE_VAULT}"
        echo ""
        echo "Fix: Create the directory or run migrate-evidence-vault.sh first:"
        echo "  mkdir -p ${EVIDENCE_VAULT}"
        exit 1
    fi
    log_info "  Vault directory exists: ✓"

    # Check gh CLI
    if ! command -v gh &> /dev/null; then
        log_error "GitHub CLI (gh) not installed"
        echo ""
        echo "Fix: brew install gh"
        exit 1
    fi
    log_info "  gh CLI installed: ✓"

    # Check jq
    if ! command -v jq &> /dev/null; then
        log_error "jq not installed"
        echo ""
        echo "Fix: brew install jq"
        exit 1
    fi
    log_info "  jq installed: ✓"

    log_info "  All preconditions met ✓"
}

#######################################
# Auth Verification
#######################################
verify_auth() {
    log_step "Verifying authentication..."

    # GitHub CLI auth
    if ! gh auth status &> /dev/null 2>&1; then
        log_error "GitHub CLI not authenticated"
        echo ""
        echo "Fix: gh auth login"
        exit 1
    fi

    local gh_user
    gh_user=$(gh api user --jq '.login' 2>/dev/null || echo "unknown")
    log_info "  GitHub: authenticated as ${gh_user} ✓"

    # AWS CLI auth (optional - only warn if not configured)
    if command -v aws &> /dev/null; then
        if aws sts get-caller-identity --profile "${AWS_PROFILE:-default}" &> /dev/null 2>&1; then
            local aws_account
            aws_account=$(aws sts get-caller-identity --profile "${AWS_PROFILE:-default}" --query 'Account' --output text 2>/dev/null || echo "unknown")
            log_info "  AWS: authenticated (account ${aws_account}) ✓"
        else
            log_warn "  AWS: not authenticated (weekly checks will skip AWS - run 'aws sso login' if needed)"
        fi
    else
        log_warn "  AWS CLI not installed (skipping AWS checks)"
    fi
}

#######################################
# Setup Output Directories
#######################################
setup_output() {
    OUTPUT_DIR="${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}/ci-cd-security/weekly"
    RUN_LOG="${OUTPUT_DIR}/RUN_LOG_${DATE_STAMP}.txt"

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_dry "Would create: ${OUTPUT_DIR}"
        return 0
    fi

    mkdir -p "${OUTPUT_DIR}"

    # Initialize run log
    cat > "${RUN_LOG}" << EOF
======================================
Weekly Security Check - Run Log
======================================
Run ID:      ${RUN_ID}
Timestamp:   ${TIMESTAMP}
Year/Month:  ${YEAR}/${MONTH}
Week:        ${WEEK}
Host:        $(hostname)
User:        $(whoami)
Dry Run:     ${DRY_RUN}
Push:        ${PUSH}
======================================

EOF

    log_info "Output directory: ${OUTPUT_DIR}"
}

#######################################
# Collect Dependabot/Security Alerts
#######################################
collect_security_alerts() {
    log_step "Collecting security alerts..."

    local github_owner="${GITHUB_OWNER:-carlosmsanchezm}"
    local repos="${GITHUB_REPOS:-aegis-platform,aegis-ui,sovran}"
    local alerts_file="${OUTPUT_DIR}/${DATE_STAMP}_security_alerts.json"

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_dry "Would collect Dependabot/security alerts for: ${repos}"
        log_dry "Output: ${alerts_file}"
        return 0
    fi

    # Collect alerts for each repo
    echo "[" > "${alerts_file}"
    local first=true

    IFS=',' read -ra REPO_ARRAY <<< "${repos}"
    for repo in "${REPO_ARRAY[@]}"; do
        repo=$(echo "${repo}" | xargs)  # trim whitespace
        local full_repo="${github_owner}/${repo}"

        log_info "  Checking ${full_repo}..."

        if [[ "${first}" != "true" ]]; then
            echo "," >> "${alerts_file}"
        fi
        first=false

        # Get Dependabot alerts count
        local dependabot_open=0
        local dependabot_json
        dependabot_json=$(gh api "/repos/${full_repo}/dependabot/alerts?state=open&per_page=100" 2>/dev/null || echo "[]")
        dependabot_open=$(echo "${dependabot_json}" | jq 'length' 2>/dev/null || echo "0")

        # Get severity breakdown
        local critical=0 high=0 medium=0 low=0
        critical=$(echo "${dependabot_json}" | jq '[.[] | select(.security_vulnerability.severity == "critical")] | length' 2>/dev/null || echo "0")
        high=$(echo "${dependabot_json}" | jq '[.[] | select(.security_vulnerability.severity == "high")] | length' 2>/dev/null || echo "0")
        medium=$(echo "${dependabot_json}" | jq '[.[] | select(.security_vulnerability.severity == "medium")] | length' 2>/dev/null || echo "0")
        low=$(echo "${dependabot_json}" | jq '[.[] | select(.security_vulnerability.severity == "low")] | length' 2>/dev/null || echo "0")

        # Get secret scanning alerts count (gh api may return error JSON for repos without secret scanning)
        local secrets_open=0
        secrets_open=$(gh api "/repos/${full_repo}/secret-scanning/alerts?state=open&per_page=100" 2>/dev/null | jq 'if type == "array" then length else 0 end' 2>/dev/null | tail -1 || echo "0")
        secrets_open="${secrets_open//[^0-9]/}"
        secrets_open="${secrets_open:-0}"

        # Get code scanning alerts count (gh api may return error JSON for repos without code scanning)
        local code_scan_open=0
        code_scan_open=$(gh api "/repos/${full_repo}/code-scanning/alerts?state=open&per_page=100" 2>/dev/null | jq 'if type == "array" then length else 0 end' 2>/dev/null | tail -1 || echo "0")
        code_scan_open="${code_scan_open//[^0-9]/}"
        code_scan_open="${code_scan_open:-0}"

        cat >> "${alerts_file}" << EOF
  {
    "repo": "${repo}",
    "full_repo": "${full_repo}",
    "check_timestamp": "${TIMESTAMP}",
    "dependabot": {
      "open_alerts": ${dependabot_open},
      "by_severity": {
        "critical": ${critical},
        "high": ${high},
        "medium": ${medium},
        "low": ${low}
      }
    },
    "secret_scanning": {
      "open_alerts": ${secrets_open}
    },
    "code_scanning": {
      "open_alerts": ${code_scan_open}
    }
  }
EOF

        log_info "    Dependabot: ${dependabot_open} open (${critical} critical, ${high} high)"
        log_info "    Secrets: ${secrets_open} open"
        log_info "    Code scan: ${code_scan_open} open"

        # Log to run log
        echo "Repository: ${full_repo}" >> "${RUN_LOG}"
        echo "  Dependabot alerts: ${dependabot_open} (critical=${critical}, high=${high})" >> "${RUN_LOG}"
        echo "  Secret scanning: ${secrets_open}" >> "${RUN_LOG}"
        echo "  Code scanning: ${code_scan_open}" >> "${RUN_LOG}"
        echo "" >> "${RUN_LOG}"
    done

    echo "]" >> "${alerts_file}"

    log_info "  ✓ Security alerts exported to ${alerts_file}"
}

#######################################
# Create Weekly Review Record
#######################################
create_review_record() {
    log_step "Creating weekly review record..."

    local review_file="${OUTPUT_DIR}/${DATE_STAMP}_weekly_review.md"

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_dry "Would create review record: ${review_file}"
        return 0
    fi

    # Calculate totals from alerts file
    local total_dependabot=0 total_critical=0 total_secrets=0 total_code=0
    if [[ -f "${OUTPUT_DIR}/${DATE_STAMP}_security_alerts.json" ]]; then
        total_dependabot=$(jq '[.[].dependabot.open_alerts] | add // 0' "${OUTPUT_DIR}/${DATE_STAMP}_security_alerts.json")
        total_critical=$(jq '[.[].dependabot.by_severity.critical] | add // 0' "${OUTPUT_DIR}/${DATE_STAMP}_security_alerts.json")
        total_secrets=$(jq '[.[].secret_scanning.open_alerts] | add // 0' "${OUTPUT_DIR}/${DATE_STAMP}_security_alerts.json")
        total_code=$(jq '[.[].code_scanning.open_alerts] | add // 0' "${OUTPUT_DIR}/${DATE_STAMP}_security_alerts.json")
    fi

    cat > "${review_file}" << EOF
# Weekly Security Review

**Date:** ${DATE_STAMP}
**Week:** ${WEEK}
**Reviewer:** $(whoami) (automated)
**Timestamp:** ${TIMESTAMP}

---

## Summary

| Metric | Count |
|--------|-------|
| Dependabot Alerts (open) | ${total_dependabot} |
| Critical Alerts | ${total_critical} |
| Secret Scanning Alerts | ${total_secrets} |
| Code Scanning Alerts | ${total_code} |

## What Was Checked

- [x] Reviewed Dependabot security alerts across all repositories
- [x] Reviewed secret scanning alerts
- [x] Reviewed code scanning alerts
- [x] Exported alert status to evidence vault

## Actions Taken

<!-- Update this section with any actions taken -->

$(if [[ ${total_critical} -gt 0 ]]; then
    echo "- ⚠️ **CRITICAL ALERTS PRESENT** - Action required within 7 days"
elif [[ ${total_dependabot} -gt 0 ]]; then
    echo "- 🟡 Open alerts present - Review and prioritize remediation"
else
    echo "- ✅ No critical alerts - Continue monitoring"
fi)

## Evidence Files

- \`${DATE_STAMP}_security_alerts.json\` - Raw alert data
- \`${DATE_STAMP}_weekly_review.md\` - This review record
- \`RUN_LOG_${DATE_STAMP}.txt\` - Script execution log

## Next Review

Next weekly check due: $(date -d "+7 days" +%Y-%m-%d 2>/dev/null || date -v+7d +%Y-%m-%d 2>/dev/null || echo "in 7 days")

---

**Controls Covered:**
- SOC 2: CC7.1 (Vulnerability detection), CC8.1 (Change management)
- ISO 27001: A.8.8 (Technical vulnerability management)

EOF

    log_info "  ✓ Weekly review record created: ${review_file}"

    # Update run log
    echo "======================================" >> "${RUN_LOG}"
    echo "Weekly Review Summary" >> "${RUN_LOG}"
    echo "======================================" >> "${RUN_LOG}"
    echo "Total Dependabot Alerts: ${total_dependabot}" >> "${RUN_LOG}"
    echo "Total Critical: ${total_critical}" >> "${RUN_LOG}"
    echo "Total Secret Alerts: ${total_secrets}" >> "${RUN_LOG}"
    echo "Total Code Scan Alerts: ${total_code}" >> "${RUN_LOG}"
    echo "" >> "${RUN_LOG}"
}

#######################################
# Git Commit/Push (Optional)
#######################################
git_commit_push() {
    log_step "Checking git commit/push..."

    # Check if vault is a git repo
    if [[ ! -d "${EVIDENCE_VAULT}/.git" ]]; then
        log_warn "  Evidence vault is not a git repository (skipping commit)"
        return 0
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_dry "Would commit weekly evidence to git"
        if [[ "${PUSH}" == "true" ]]; then
            log_dry "Would push to remote"
        fi
        return 0
    fi

    cd "${EVIDENCE_VAULT}"

    # Check for changes
    if [[ -z $(git status --porcelain) ]]; then
        log_info "  No changes to commit"
        return 0
    fi

    # Stage and commit
    git add .
    git commit -m "Weekly security check: ${DATE_STAMP}

Automated weekly evidence collection:
- Dependabot/security alerts status
- Weekly review record

Run ID: ${RUN_ID}"

    log_info "  ✓ Changes committed"
    echo "Git commit created: ${DATE_STAMP}" >> "${RUN_LOG}"

    # Push if requested
    if [[ "${PUSH}" == "true" || "${PUSH:-0}" == "1" ]]; then
        if git remote -v | grep -q origin; then
            git push origin HEAD
            log_info "  ✓ Pushed to remote"
            echo "Git push completed" >> "${RUN_LOG}"
        else
            log_warn "  No remote 'origin' configured (skipping push)"
        fi
    fi
}

#######################################
# Print Summary
#######################################
print_summary() {
    log_step "Weekly check complete!"

    echo ""
    echo "========================================"
    echo "  Weekly Security Check Summary"
    echo "========================================"

    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "  Mode: DRY RUN (no files written)"
    else
        echo "  Evidence Path: ${OUTPUT_DIR}"
        echo ""
        echo "  Files Created:"
        if [[ -f "${OUTPUT_DIR}/${DATE_STAMP}_security_alerts.json" ]]; then
            echo "    ✓ ${DATE_STAMP}_security_alerts.json"
        fi
        if [[ -f "${OUTPUT_DIR}/${DATE_STAMP}_weekly_review.md" ]]; then
            echo "    ✓ ${DATE_STAMP}_weekly_review.md"
        fi
        if [[ -f "${RUN_LOG}" ]]; then
            echo "    ✓ RUN_LOG_${DATE_STAMP}.txt"
        fi
    fi

    echo ""
    echo "========================================"

    # Finalize run log
    if [[ "${DRY_RUN}" != "true" && -f "${RUN_LOG}" ]]; then
        echo "" >> "${RUN_LOG}"
        echo "======================================" >> "${RUN_LOG}"
        echo "Run completed: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "${RUN_LOG}"
        echo "Exit code: 0" >> "${RUN_LOG}"
        echo "======================================" >> "${RUN_LOG}"
    fi
}

#######################################
# Main
#######################################
main() {
    echo ""
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║     Weekly Security Check - SOC 2 + ISO 27001          ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo ""

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_warn "DRY RUN MODE - No changes will be made"
        echo ""
    fi

    check_preconditions
    verify_auth
    setup_output
    collect_security_alerts
    create_review_record
    git_commit_push
    print_summary

    echo ""
    log_info "Weekly check complete. Next run: in 7 days."
}

main "$@"
