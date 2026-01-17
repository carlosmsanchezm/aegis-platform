#!/bin/bash
# setup_evidence_vault.sh
#
# Creates and initializes the SOC 2 evidence vault repository
#
# Prerequisites:
# - gh CLI installed and authenticated
# - git configured with user.name and user.email
#
# Usage: ./setup_evidence_vault.sh

set -euo pipefail

# Configuration
REPO_NAME="aegis-compliance-evidence"
REPO_OWNER="${GITHUB_OWNER:-carlosmsanchezm}"
CURRENT_YEAR=$(date +%Y)
CURRENT_MONTH=$(date +%Y-%m)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
VAULT_SETUP_DIR="${REPO_ROOT}/docs/compliance/soc2/evidence-vault-setup"

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
    log_info "Checking prerequisites..."

    if ! command -v gh &> /dev/null; then
        log_error "gh CLI not found. Install with: brew install gh"
        exit 1
    fi

    if ! gh auth status &> /dev/null; then
        log_error "gh CLI not authenticated. Run: gh auth login"
        exit 1
    fi

    if ! command -v git &> /dev/null; then
        log_error "git not found"
        exit 1
    fi

    log_info "  ✓ Prerequisites OK"
}

# Create the repository
create_repo() {
    log_info "Creating private repository: ${REPO_OWNER}/${REPO_NAME}..."

    # Check if repo already exists
    if gh repo view "${REPO_OWNER}/${REPO_NAME}" &> /dev/null; then
        log_warn "Repository already exists. Skipping creation."
        return 0
    fi

    gh repo create "${REPO_NAME}" \
        --private \
        --description "SOC 2 Type II Evidence Vault - CONFIDENTIAL" \
        --disable-wiki \
        --disable-issues

    log_info "  ✓ Repository created"
}

# Configure repository settings
configure_repo() {
    log_info "Configuring repository settings..."

    # Disable GitHub Pages
    gh api -X PUT "/repos/${REPO_OWNER}/${REPO_NAME}" \
        -f has_pages=false \
        -f has_wiki=false \
        -f has_issues=false \
        -f allow_squash_merge=true \
        -f allow_merge_commit=false \
        -f allow_rebase_merge=false \
        -f delete_branch_on_merge=true \
        > /dev/null 2>&1 || log_warn "  Some settings may not have applied"

    log_info "  ✓ Repository configured"
}

# Set up branch protection (solo-founder safe)
setup_branch_protection() {
    log_info "Setting up branch protection..."

    # Create initial commit first (branch protection needs a branch to exist)
    # We'll do this after cloning

    log_info "  Branch protection will be configured after initial commit"
}

# Clone and initialize the repository
init_repo() {
    log_info "Cloning repository..."

    local temp_dir=$(mktemp -d)
    cd "${temp_dir}"

    gh repo clone "${REPO_OWNER}/${REPO_NAME}" .

    log_info "  ✓ Repository cloned to ${temp_dir}"

    # Copy setup files
    log_info "Copying vault setup files..."

    if [[ -d "${VAULT_SETUP_DIR}" ]]; then
        cp "${VAULT_SETUP_DIR}/README.md" ./README.md
        cp "${VAULT_SETUP_DIR}/.gitignore" ./.gitignore
        cp "${VAULT_SETUP_DIR}/EVIDENCE_INTAKE_CHECKLIST.md" ./EVIDENCE_INTAKE_CHECKLIST.md
        log_info "  ✓ Setup files copied"
    else
        log_warn "  Vault setup directory not found, creating basic files..."
        echo "# Aegis SOC 2 Evidence Vault" > README.md
        echo "*.pem" > .gitignore
        echo "*.key" >> .gitignore
        echo ".env*" >> .gitignore
    fi

    # Create folder structure
    log_info "Creating folder structure..."

    # Current year structure
    mkdir -p "soc2/${CURRENT_YEAR}/governance/org-chart"
    mkdir -p "soc2/${CURRENT_YEAR}/governance/policies"
    mkdir -p "soc2/${CURRENT_YEAR}/governance/risk-assessment"
    mkdir -p "soc2/${CURRENT_YEAR}/vendor-management"
    mkdir -p "soc2/${CURRENT_YEAR}/training-policy-ack"
    mkdir -p "soc2/${CURRENT_YEAR}/backups-dr"

    # Quarterly folders
    for q in Q1 Q2 Q3 Q4; do
        mkdir -p "soc2/${CURRENT_YEAR}/${q}/access-reviews"
    done

    # Monthly folders for current year
    for month in 01 02 03 04 05 06 07 08 09 10 11 12; do
        local month_dir="soc2/${CURRENT_YEAR}/${CURRENT_YEAR}-${month}"
        mkdir -p "${month_dir}/access-reviews"
        mkdir -p "${month_dir}/change-management"
        mkdir -p "${month_dir}/ci-cd-security"
        mkdir -p "${month_dir}/vuln-management"
        mkdir -p "${month_dir}/logging-monitoring"
        mkdir -p "${month_dir}/incident-response"
        mkdir -p "${month_dir}/communication"
    done

    # Templates
    mkdir -p "soc2/templates"

    # Add .gitkeep files to empty directories
    find soc2 -type d -empty -exec touch {}/.gitkeep \;

    log_info "  ✓ Folder structure created"

    # Create template files
    log_info "Creating template files..."

    cat > "soc2/templates/access-review-template.md" << 'EOF'
# Quarterly Access Review - QX YYYY

**Review Period:** YYYY-MM-DD to YYYY-MM-DD
**Reviewer:** [Name]
**Completion Date:** YYYY-MM-DD

## Systems Reviewed

- [ ] GitHub
- [ ] AWS IAM
- [ ] Google Workspace (if applicable)
- [ ] Other: ___

## Summary

| System | Total Users | Users Certified | Users Removed | Notes |
|--------|-------------|-----------------|---------------|-------|
| GitHub | | | | |
| AWS | | | | |

## Certification

I certify that I have reviewed all access for my direct reports and confirmed that:
- Access is appropriate for current job functions
- Terminated users have been removed
- Unnecessary access has been revoked

**Manager Signature:** _______________
**Date:** _______________
EOF

    cat > "soc2/templates/incident-report-template.md" << 'EOF'
# Security Incident Report

**Incident ID:** INC-YYYY-###
**Severity:** [Critical/High/Medium/Low]
**Status:** [Open/Investigating/Contained/Resolved/Closed]

## Timeline

| Time (UTC) | Event |
|-----------|-------|
| YYYY-MM-DD HH:MM | Incident detected |
| | |

## Summary

[Brief description of what happened]

## Impact

- Systems affected:
- Data affected:
- Users affected:

## Root Cause

[Description of root cause - complete after investigation]

## Remediation

- [ ] Immediate containment actions
- [ ] Root cause addressed
- [ ] Preventive measures implemented

## Lessons Learned

[Complete after incident closure]
EOF

    log_info "  ✓ Templates created"

    # Initial commit
    log_info "Creating initial commit..."
    git add -A
    git commit -m "Initialize SOC 2 evidence vault structure

- Add folder structure for ${CURRENT_YEAR}
- Add README with naming conventions and policies
- Add .gitignore for secrets protection
- Add evidence intake checklist
- Add templates for access reviews and incidents"

    git push origin main

    log_info "  ✓ Initial commit pushed"

    # Now set up branch protection
    log_info "Setting up branch protection on main..."

    gh api -X PUT "/repos/${REPO_OWNER}/${REPO_NAME}/branches/main/protection" \
        -f required_status_checks='null' \
        -F enforce_admins=false \
        -f required_pull_request_reviews='{"required_approving_review_count":0,"dismiss_stale_reviews":false}' \
        -f restrictions='null' \
        -F required_linear_history=false \
        -F allow_force_pushes=false \
        -F allow_deletions=false \
        > /dev/null 2>&1 || log_warn "  Branch protection may need manual configuration"

    log_info "  ✓ Branch protection configured (solo-founder safe: no required reviewers)"

    # Print location
    echo ""
    log_info "=== Evidence Vault Created ==="
    log_info "Repository: https://github.com/${REPO_OWNER}/${REPO_NAME}"
    log_info "Local clone: ${temp_dir}"
    echo ""
    log_info "Next steps:"
    echo "  1. Move the cloned repo to your preferred location:"
    echo "     mv ${temp_dir} ~/code/${REPO_NAME}"
    echo ""
    echo "  2. Run evidence collection:"
    echo "     cd ~/code/aegis-platform"
    echo "     ./scripts/compliance/export_github_security_baseline.sh ~/code/${REPO_NAME}/soc2/${CURRENT_YEAR}/${CURRENT_MONTH}/ci-cd-security"
    echo "     AWS_PROFILE=myclaude ./scripts/compliance/export_aws_security_baseline.sh ~/code/${REPO_NAME}/soc2/${CURRENT_YEAR}/${CURRENT_MONTH}/access-reviews"
    echo ""
    echo "  3. Commit evidence to vault:"
    echo "     cd ~/code/${REPO_NAME}"
    echo "     git add -A"
    echo "     git commit -m 'SOC2 evidence: baseline exports ${CURRENT_MONTH}'"
    echo "     git push"

    # Return the temp directory path
    echo "${temp_dir}"
}

# Main
main() {
    log_info "=== Aegis SOC 2 Evidence Vault Setup ==="
    echo ""

    check_prereqs
    create_repo
    configure_repo
    init_repo

    echo ""
    log_info "=== Setup Complete ==="

    # Security reminder
    echo ""
    log_warn "🚨 SECURITY REMINDER 🚨"
    echo "If you shared your GitHub PAT in chat, rotate it now:"
    echo "  https://github.com/settings/tokens"
}

main "$@"
