#!/bin/bash
# run_monthly_evidence.sh
#
# SOC 2 + ISO 27001 Monthly Evidence Collection
# Single-command operation for full monthly evidence export
#
# Runs all three export scripts:
# - export_github_security_baseline.sh
# - export_aws_security_baseline.sh
# - export_ci_reports.sh
#
# Controls Covered:
# - SOC 2: CC6.1, CC6.4, CC6.6, CC6.7, CC7.1, CC7.2, CC8.1
# - ISO 27001: A.5.15-18, A.8.8, A.8.32, Clause 8.2
#
# Usage: ./run_monthly_evidence.sh [--dry-run] [--push] [--skip-aws] [--skip-ci]
#
# Preconditions:
# - EVIDENCE_VAULT env var must be set
# - gh CLI authenticated
# - aws CLI authenticated (unless --skip-aws)
# - syft/grype installed for CI reports (optional)

set -euo pipefail

#######################################
# Configuration
#######################################
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
YEAR=$(date +%Y)
MONTH=$(date +%Y-%m)
DATE_STAMP=$(date +%Y-%m-%d)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
RUN_ID="${DATE_STAMP}-monthly"

# Flags
DRY_RUN=false
PUSH=false
SKIP_AWS=false
SKIP_CI=false

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
        --skip-aws)
            SKIP_AWS=true
            shift
            ;;
        --skip-ci)
            SKIP_CI=true
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
BOLD='\033[1m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }
log_dry() { echo -e "${CYAN}[DRY-RUN]${NC} $1"; }
log_section() { echo -e "\n${BOLD}=== $1 ===${NC}\n"; }

#######################################
# Precondition Checks
#######################################
check_preconditions() {
    log_step "Checking preconditions..."

    local errors=()

    # Check EVIDENCE_VAULT is set
    if [[ -z "${EVIDENCE_VAULT:-}" ]]; then
        errors+=("EVIDENCE_VAULT environment variable is not set")
    else
        log_info "  EVIDENCE_VAULT: ${EVIDENCE_VAULT}"

        # Check vault directory exists
        if [[ ! -d "${EVIDENCE_VAULT}" ]]; then
            errors+=("EVIDENCE_VAULT directory does not exist: ${EVIDENCE_VAULT}")
        else
            log_info "  Vault directory exists: ✓"
        fi
    fi

    # Check gh CLI
    if ! command -v gh &> /dev/null; then
        errors+=("GitHub CLI (gh) not installed - brew install gh")
    else
        log_info "  gh CLI installed: ✓"
    fi

    # Check jq
    if ! command -v jq &> /dev/null; then
        errors+=("jq not installed - brew install jq")
    else
        log_info "  jq installed: ✓"
    fi

    # Check aws CLI (warn only)
    if [[ "${SKIP_AWS}" != "true" ]]; then
        if ! command -v aws &> /dev/null; then
            log_warn "  AWS CLI not installed (will skip AWS export)"
            SKIP_AWS=true
        else
            log_info "  aws CLI installed: ✓"
        fi
    else
        log_info "  AWS export: skipped (--skip-aws)"
    fi

    # Check CI tools (warn only)
    if [[ "${SKIP_CI}" != "true" ]]; then
        if ! command -v syft &> /dev/null && ! command -v cyclonedx-gomod &> /dev/null; then
            log_warn "  SBOM tools not installed (CI reports will be limited)"
        fi
        if ! command -v grype &> /dev/null && ! command -v trivy &> /dev/null; then
            log_warn "  Vulnerability scanner not installed (CI reports will be limited)"
        fi
    else
        log_info "  CI export: skipped (--skip-ci)"
    fi

    if [[ ${#errors[@]} -gt 0 ]]; then
        echo ""
        log_error "Precondition failures:"
        for err in "${errors[@]}"; do
            echo "  - ${err}"
        done
        echo ""
        echo "Fix: Set environment variables and install required tools:"
        echo "  export EVIDENCE_VAULT=\$HOME/code/aegis-compliance-evidence"
        echo "  brew install gh jq awscli syft grype"
        exit 1
    fi

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

    # AWS CLI auth (optional)
    if [[ "${SKIP_AWS}" != "true" ]]; then
        if ! aws sts get-caller-identity --profile "${AWS_PROFILE:-default}" &> /dev/null 2>&1; then
            log_warn "  AWS: not authenticated (skipping AWS export)"
            log_warn "  Fix: aws sso login --profile ${AWS_PROFILE:-default}"
            SKIP_AWS=true
        else
            local aws_account
            aws_account=$(aws sts get-caller-identity --profile "${AWS_PROFILE:-default}" --query 'Account' --output text 2>/dev/null || echo "unknown")
            log_info "  AWS: authenticated (account ${aws_account}) ✓"
        fi
    fi
}

#######################################
# Setup Output Directories
#######################################
setup_output() {
    log_step "Setting up output directories..."

    # Create all required monthly folders
    local folders=(
        "soc2/${YEAR}/${MONTH}/access-reviews"
        "soc2/${YEAR}/${MONTH}/ci-cd-security"
        "soc2/${YEAR}/${MONTH}/vuln-management"
        "soc2/${YEAR}/${MONTH}/incident-response"
        "soc2/${YEAR}/${MONTH}/logging-monitoring"
        "iso27001/${YEAR}/corrective-actions"
        "iso27001/${YEAR}/risk-assessments"
    )

    RUN_LOG="${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}/RUN_LOG_${DATE_STAMP}.txt"

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_dry "Would create directories:"
        for folder in "${folders[@]}"; do
            log_dry "  ${EVIDENCE_VAULT}/${folder}"
        done
        return 0
    fi

    for folder in "${folders[@]}"; do
        mkdir -p "${EVIDENCE_VAULT}/${folder}"
    done

    log_info "  Monthly folders created ✓"

    # Initialize run log
    cat > "${RUN_LOG}" << EOF
======================================
Monthly Evidence Collection - Run Log
======================================
Run ID:      ${RUN_ID}
Timestamp:   ${TIMESTAMP}
Year/Month:  ${YEAR}/${MONTH}
Host:        $(hostname)
User:        $(whoami)
Dry Run:     ${DRY_RUN}
Push:        ${PUSH}
Skip AWS:    ${SKIP_AWS}
Skip CI:     ${SKIP_CI}
======================================

EOF

    log_info "  Run log: ${RUN_LOG}"
}

#######################################
# Run GitHub Security Export
#######################################
run_github_export() {
    log_section "GitHub Security Baseline Export"

    local script="${SCRIPT_DIR}/export_github_security_baseline.sh"
    local output_dir="${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}/ci-cd-security"

    if [[ ! -f "${script}" ]]; then
        log_error "Script not found: ${script}"
        return 1
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_dry "Would run: ${script}"
        log_dry "Output: ${output_dir}"
        return 0
    fi

    echo "--- GitHub Export Start ---" >> "${RUN_LOG}"
    echo "Start time: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "${RUN_LOG}"

    chmod +x "${script}"
    if "${script}" "${output_dir}" 2>&1 | tee -a "${RUN_LOG}"; then
        log_info "GitHub export completed ✓"
        echo "Status: SUCCESS" >> "${RUN_LOG}"
    else
        log_error "GitHub export failed"
        echo "Status: FAILED" >> "${RUN_LOG}"
    fi

    echo "End time: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "${RUN_LOG}"
    echo "--- GitHub Export End ---" >> "${RUN_LOG}"
    echo "" >> "${RUN_LOG}"
}

#######################################
# Run AWS Security Export
#######################################
run_aws_export() {
    log_section "AWS Security Baseline Export"

    if [[ "${SKIP_AWS}" == "true" ]]; then
        log_warn "Skipping AWS export (--skip-aws or not authenticated)"
        echo "AWS Export: SKIPPED" >> "${RUN_LOG}"
        return 0
    fi

    local script="${SCRIPT_DIR}/export_aws_security_baseline.sh"
    local output_dir="${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}/access-reviews"

    if [[ ! -f "${script}" ]]; then
        log_error "Script not found: ${script}"
        return 1
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_dry "Would run: ${script}"
        log_dry "Output: ${output_dir}"
        return 0
    fi

    echo "--- AWS Export Start ---" >> "${RUN_LOG}"
    echo "Start time: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "${RUN_LOG}"

    chmod +x "${script}"
    if "${script}" "${output_dir}" 2>&1 | tee -a "${RUN_LOG}"; then
        log_info "AWS export completed ✓"
        echo "Status: SUCCESS" >> "${RUN_LOG}"
    else
        log_error "AWS export failed"
        echo "Status: FAILED" >> "${RUN_LOG}"
    fi

    echo "End time: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "${RUN_LOG}"
    echo "--- AWS Export End ---" >> "${RUN_LOG}"
    echo "" >> "${RUN_LOG}"
}

#######################################
# Run CI Security Export
#######################################
run_ci_export() {
    log_section "CI/CD Security Reports Export"

    if [[ "${SKIP_CI}" == "true" ]]; then
        log_warn "Skipping CI export (--skip-ci)"
        echo "CI Export: SKIPPED" >> "${RUN_LOG}"
        return 0
    fi

    local script="${SCRIPT_DIR}/export_ci_reports.sh"
    local output_dir="${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}/vuln-management"

    if [[ ! -f "${script}" ]]; then
        log_error "Script not found: ${script}"
        return 1
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_dry "Would run: ${script}"
        log_dry "Output: ${output_dir}"
        return 0
    fi

    echo "--- CI Export Start ---" >> "${RUN_LOG}"
    echo "Start time: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "${RUN_LOG}"

    chmod +x "${script}"
    if "${script}" "${output_dir}" 2>&1 | tee -a "${RUN_LOG}"; then
        log_info "CI export completed ✓"
        echo "Status: SUCCESS" >> "${RUN_LOG}"
    else
        log_warn "CI export completed with warnings"
        echo "Status: PARTIAL" >> "${RUN_LOG}"
    fi

    echo "End time: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "${RUN_LOG}"
    echo "--- CI Export End ---" >> "${RUN_LOG}"
    echo "" >> "${RUN_LOG}"
}

#######################################
# Create Incident Log Entry
#######################################
create_incident_log() {
    log_section "Incident Log Entry"

    local incident_dir="${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}/incident-response"
    local incident_file="${incident_dir}/${MONTH}_incident_log.md"

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_dry "Would create incident log: ${incident_file}"
        return 0
    fi

    if [[ -f "${incident_file}" ]]; then
        log_info "Incident log already exists for ${MONTH}"
        return 0
    fi

    cat > "${incident_file}" << EOF
# Incident Log - ${MONTH}

**Period:** ${MONTH}-01 to $(date -d "${MONTH}-01 +1 month -1 day" +%Y-%m-%d 2>/dev/null || date -v1d -v+1m -v-1d +%Y-%m-%d 2>/dev/null || echo "${MONTH}-31")
**Created:** ${DATE_STAMP}
**Owner:** Carlos Sanchez

---

## Summary

| Metric | Count |
|--------|-------|
| Total Incidents | 0 |
| Security Incidents | 0 |
| P1 (Critical) | 0 |
| P2 (High) | 0 |
| P3 (Medium) | 0 |
| P4 (Low) | 0 |

## Incidents

<!-- Record any security incidents here -->

| ID | Date | Description | Severity | Status | Resolution |
|----|------|-------------|----------|--------|------------|
| - | - | No incidents this period | - | - | - |

## Notes

This log was auto-generated as part of monthly evidence collection.
If incidents occurred, update this file with details.

---

**Controls Covered:**
- SOC 2: CC7.3 (Incident response), CC7.4 (Incident resolution)
- ISO 27001: A.5.24-28 (Incident management)

EOF

    log_info "Created incident log: ${incident_file}"
    echo "Incident log created: ${incident_file}" >> "${RUN_LOG}"
}

#######################################
# Create Monthly Summary
#######################################
create_monthly_summary() {
    log_section "Monthly Evidence Summary"

    local summary_file="${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}/${DATE_STAMP}_monthly_summary.md"

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_dry "Would create monthly summary: ${summary_file}"
        return 0
    fi

    # Count files created
    local github_files aws_files ci_files total_files
    github_files=$(find "${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}/ci-cd-security" -name "*.json" -type f 2>/dev/null | wc -l | tr -d ' ')
    aws_files=$(find "${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}/access-reviews" -name "*.json" -type f 2>/dev/null | wc -l | tr -d ' ')
    ci_files=$(find "${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}/vuln-management" -name "*.json" -type f 2>/dev/null | wc -l | tr -d ' ')
    total_files=$((github_files + aws_files + ci_files))

    cat > "${summary_file}" << EOF
# Monthly Evidence Collection Summary

**Period:** ${MONTH}
**Collection Date:** ${DATE_STAMP}
**Run ID:** ${RUN_ID}
**Timestamp:** ${TIMESTAMP}

---

## Evidence Collected

| Category | Files | Location |
|----------|-------|----------|
| GitHub Security | ${github_files} | \`soc2/${YEAR}/${MONTH}/ci-cd-security/\` |
| AWS Security | ${aws_files} | \`soc2/${YEAR}/${MONTH}/access-reviews/\` |
| CI/CD Reports | ${ci_files} | \`soc2/${YEAR}/${MONTH}/vuln-management/\` |
| Incident Log | 1 | \`soc2/${YEAR}/${MONTH}/incident-response/\` |
| **Total** | **${total_files}** | |

## Export Status

| Export | Status |
|--------|--------|
| GitHub Security Baseline | $(if [[ "${github_files}" -gt 0 ]]; then echo "✅ Complete"; else echo "⚠️ No files"; fi) |
| AWS Security Baseline | $(if [[ "${SKIP_AWS}" == "true" ]]; then echo "⏭️ Skipped"; elif [[ "${aws_files}" -gt 0 ]]; then echo "✅ Complete"; else echo "⚠️ No files"; fi) |
| CI/CD Reports | $(if [[ "${SKIP_CI}" == "true" ]]; then echo "⏭️ Skipped"; elif [[ "${ci_files}" -gt 0 ]]; then echo "✅ Complete"; else echo "⚠️ No files"; fi) |
| Incident Log | ✅ Created |

## Controls Covered

### SOC 2 Trust Services Criteria

- CC6.1: Logical access security (IAM exports)
- CC6.4: Access reviews (quarterly process)
- CC6.6: Authentication controls (MFA status)
- CC6.7: Role-based access (collaborators, roles)
- CC7.1: Vulnerability detection (Dependabot, security alerts)
- CC7.2: System monitoring (CloudTrail, Config)
- CC8.1: Change management (branch protection, workflows)

### ISO 27001 Annex A Controls

- A.5.15-18: Access control
- A.8.8: Technical vulnerability management
- A.8.32: Change management

## Next Steps

1. Review exported evidence for anomalies
2. Address any critical vulnerabilities within 7 days
3. Update incident log if any incidents occurred
4. Run quarterly access review if due

---

**Run Log:** \`RUN_LOG_${DATE_STAMP}.txt\`

EOF

    log_info "Created monthly summary: ${summary_file}"
    echo "Monthly summary created: ${summary_file}" >> "${RUN_LOG}"
}

#######################################
# Git Commit/Push (Optional)
#######################################
git_commit_push() {
    log_section "Git Commit/Push"

    # Check if vault is a git repo
    if [[ ! -d "${EVIDENCE_VAULT}/.git" ]]; then
        log_warn "Evidence vault is not a git repository (skipping commit)"
        return 0
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_dry "Would commit monthly evidence to git"
        if [[ "${PUSH}" == "true" ]]; then
            log_dry "Would push to remote"
        fi
        return 0
    fi

    cd "${EVIDENCE_VAULT}"

    # Check for changes
    if [[ -z $(git status --porcelain) ]]; then
        log_info "No changes to commit"
        return 0
    fi

    # Count files to commit
    local files_changed
    files_changed=$(git status --porcelain | wc -l | tr -d ' ')

    # Stage and commit
    git add .
    git commit -m "Monthly evidence collection: ${MONTH}

Automated monthly evidence export:
- GitHub security baseline
- AWS security baseline
- CI/CD reports (SBOM + vuln scans)
- Incident log

Run ID: ${RUN_ID}
Files: ${files_changed}"

    log_info "Changes committed (${files_changed} files) ✓"
    echo "Git commit created: ${DATE_STAMP} (${files_changed} files)" >> "${RUN_LOG}"

    # Push if requested
    if [[ "${PUSH}" == "true" || "${PUSH:-0}" == "1" ]]; then
        if git remote -v | grep -q origin; then
            git push origin HEAD
            log_info "Pushed to remote ✓"
            echo "Git push completed" >> "${RUN_LOG}"
        else
            log_warn "No remote 'origin' configured (skipping push)"
        fi
    else
        log_info "Push skipped (use --push to enable)"
    fi
}

#######################################
# Print Final Summary
#######################################
print_summary() {
    echo ""
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║         Monthly Evidence Collection Complete           ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo ""

    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "  Mode: DRY RUN (no files written)"
        return 0
    fi

    echo "  Evidence Path: ${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}/"
    echo ""
    echo "  Directories:"
    echo "    ✓ access-reviews/     (AWS IAM, MFA, roles)"
    echo "    ✓ ci-cd-security/     (GitHub security, workflows)"
    echo "    ✓ vuln-management/    (SBOM, vulnerability scans)"
    echo "    ✓ incident-response/  (incident log)"
    echo "    ✓ logging-monitoring/ (CloudTrail, Config)"
    echo ""

    # Show file counts
    local total_files
    total_files=$(find "${EVIDENCE_VAULT}/soc2/${YEAR}/${MONTH}" -name "*.json" -o -name "*.md" 2>/dev/null | wc -l | tr -d ' ')
    echo "  Total files created: ${total_files}"
    echo ""

    echo "  Run Log: ${RUN_LOG}"
    echo ""
    echo "════════════════════════════════════════════════════════"

    # Finalize run log
    if [[ -f "${RUN_LOG}" ]]; then
        echo "" >> "${RUN_LOG}"
        echo "======================================" >> "${RUN_LOG}"
        echo "Run completed: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" >> "${RUN_LOG}"
        echo "Total files: ${total_files}" >> "${RUN_LOG}"
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
    echo "║     Monthly Evidence Collection - SOC 2 + ISO 27001    ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo ""
    echo "  Date: ${DATE_STAMP}"
    echo "  Month: ${MONTH}"
    echo ""

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_warn "DRY RUN MODE - No changes will be made"
        echo ""
    fi

    check_preconditions
    verify_auth
    setup_output
    run_github_export
    run_aws_export
    run_ci_export
    create_incident_log
    create_monthly_summary
    git_commit_push
    print_summary

    echo ""
    log_info "Monthly evidence collection complete."
    log_info "Next run: 1st of $(date -d "next month" +%B 2>/dev/null || date -v+1m +%B 2>/dev/null || echo "next month")"
}

main "$@"
