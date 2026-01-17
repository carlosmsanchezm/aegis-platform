#!/bin/bash
# export_github_security_baseline.sh
#
# SOC 2 Evidence Collection Script - GitHub Security Baseline
# Exports security-relevant configuration from GitHub org/repos
#
# Evidence Maps to:
# - CC6.1: Logical access security
# - CC6.7: Role-based access
# - CC7.1: Vulnerability detection (Dependabot, secret scanning)
# - CC8.1: Change management (branch protections)
#
# Usage: ./export_github_security_baseline.sh [output_dir]
#
# Prerequisites:
# - gh CLI installed and authenticated
# - jq installed
# - Access to the GitHub org/repos

set -euo pipefail

# Configuration
GITHUB_OWNER="${GITHUB_OWNER:-carlosmsanchezm}"
REPOS="${GITHUB_REPOS:-aegis-platform,aegis-ui,sovran}"  # Comma-separated list

# Evidence vault location (use env var or argument or default)
EVIDENCE_VAULT="${EVIDENCE_VAULT:-$HOME/code/aegis-compliance-evidence}"
OUTPUT_DIR="${1:-$EVIDENCE_VAULT/soc2/$(date +%Y)/$(date +%Y-%m)/ci-cd-security}"
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

    if ! command -v gh &> /dev/null; then
        missing+=("gh CLI")
    fi

    if ! command -v jq &> /dev/null; then
        missing+=("jq")
    fi

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing prerequisites: ${missing[*]}"
        echo "Install with:"
        echo "  brew install gh jq"
        exit 1
    fi

    # Check gh auth
    if ! gh auth status &> /dev/null; then
        log_error "GitHub CLI not authenticated. Run: gh auth login"
        exit 1
    fi
}

# Create output directory
setup_output() {
    mkdir -p "${OUTPUT_DIR}"
    log_info "Output directory: ${OUTPUT_DIR}"
}

# Export org-level security settings (if org exists)
export_org_settings() {
    log_info "Checking for organization settings..."

    # Try to get org settings - may fail if personal account
    if gh api "/orgs/${GITHUB_OWNER}" &> /dev/null 2>&1; then
        log_info "Exporting organization settings..."

        gh api "/orgs/${GITHUB_OWNER}" | jq '{
            login: .login,
            two_factor_requirement_enabled: .two_factor_requirement_enabled,
            default_repository_permission: .default_repository_permission,
            members_can_create_repositories: .members_can_create_repositories,
            members_can_create_public_repositories: .members_can_create_public_repositories,
            members_can_create_private_repositories: .members_can_create_private_repositories,
            web_commit_signoff_required: .web_commit_signoff_required
        }' > "${OUTPUT_DIR}/${DATE_STAMP}_org_settings.json"

        # Export org members
        gh api "/orgs/${GITHUB_OWNER}/members" --paginate | jq '[.[] | {login: .login, role: "member"}]' \
            > "${OUTPUT_DIR}/${DATE_STAMP}_org_members.json" 2>/dev/null || true

        log_info "  ✓ Organization settings exported"
    else
        log_warn "Not an organization or no org access - skipping org-level exports"
        echo '{"note": "Personal account - no org-level settings"}' > "${OUTPUT_DIR}/${DATE_STAMP}_org_settings.json"
    fi
}

# Export repository settings
export_repo_settings() {
    local repo=$1
    local full_repo="${GITHUB_OWNER}/${repo}"

    log_info "Exporting settings for ${full_repo}..."

    local repo_dir="${OUTPUT_DIR}/${repo}"
    mkdir -p "${repo_dir}"

    # Basic repo info
    gh api "/repos/${full_repo}" | jq '{
        name: .name,
        full_name: .full_name,
        private: .private,
        visibility: .visibility,
        default_branch: .default_branch,
        allow_forking: .allow_forking,
        delete_branch_on_merge: .delete_branch_on_merge,
        allow_merge_commit: .allow_merge_commit,
        allow_squash_merge: .allow_squash_merge,
        allow_rebase_merge: .allow_rebase_merge,
        allow_auto_merge: .allow_auto_merge,
        archived: .archived,
        has_issues: .has_issues,
        has_wiki: .has_wiki,
        security_and_analysis: .security_and_analysis
    }' > "${repo_dir}/${DATE_STAMP}_repo_settings.json" 2>/dev/null || {
        log_warn "  Could not export repo settings for ${repo}"
    }

    # Branch protection rules
    log_info "  Checking branch protection..."
    local default_branch
    default_branch=$(gh api "/repos/${full_repo}" --jq '.default_branch' 2>/dev/null || echo "main")

    gh api "/repos/${full_repo}/branches/${default_branch}/protection" 2>/dev/null | jq '{
        required_status_checks: .required_status_checks,
        enforce_admins: .enforce_admins,
        required_pull_request_reviews: .required_pull_request_reviews,
        restrictions: .restrictions,
        required_linear_history: .required_linear_history,
        allow_force_pushes: .allow_force_pushes,
        allow_deletions: .allow_deletions,
        required_conversation_resolution: .required_conversation_resolution,
        required_signatures: .required_signatures
    }' > "${repo_dir}/${DATE_STAMP}_branch_protection_${default_branch}.json" 2>/dev/null || {
        log_warn "  No branch protection configured for ${default_branch}"
        echo '{"status": "NOT_CONFIGURED", "branch": "'"${default_branch}"'"}' > "${repo_dir}/${DATE_STAMP}_branch_protection_${default_branch}.json"
    }

    # Collaborators/access
    log_info "  Exporting collaborators..."
    gh api "/repos/${full_repo}/collaborators" --paginate 2>/dev/null | jq '[.[] | {
        login: .login,
        permissions: .permissions,
        role_name: .role_name
    }]' > "${repo_dir}/${DATE_STAMP}_collaborators.json" 2>/dev/null || {
        log_warn "  Could not export collaborators"
        echo '[]' > "${repo_dir}/${DATE_STAMP}_collaborators.json"
    }

    # Dependabot alerts summary (count only - no sensitive details)
    log_info "  Checking Dependabot alerts..."
    gh api "/repos/${full_repo}/dependabot/alerts?state=open&per_page=100" 2>/dev/null | jq '{
        total_open: length,
        by_severity: (group_by(.security_vulnerability.severity) | map({(.[0].security_vulnerability.severity // "unknown"): length}) | add)
    }' > "${repo_dir}/${DATE_STAMP}_dependabot_summary.json" 2>/dev/null || {
        log_warn "  Dependabot alerts not accessible (may not be enabled)"
        echo '{"status": "NOT_ENABLED_OR_NO_ACCESS"}' > "${repo_dir}/${DATE_STAMP}_dependabot_summary.json"
    }

    # Secret scanning alerts summary
    log_info "  Checking secret scanning..."
    gh api "/repos/${full_repo}/secret-scanning/alerts?state=open&per_page=100" 2>/dev/null | jq '{
        total_open: length,
        by_secret_type: (group_by(.secret_type) | map({(.[0].secret_type // "unknown"): length}) | add)
    }' > "${repo_dir}/${DATE_STAMP}_secret_scanning_summary.json" 2>/dev/null || {
        log_warn "  Secret scanning not accessible (may not be enabled)"
        echo '{"status": "NOT_ENABLED_OR_NO_ACCESS"}' > "${repo_dir}/${DATE_STAMP}_secret_scanning_summary.json"
    }

    # Code scanning (CodeQL) status
    log_info "  Checking code scanning..."
    gh api "/repos/${full_repo}/code-scanning/alerts?state=open&per_page=100" 2>/dev/null | jq '{
        total_open: length,
        by_severity: (group_by(.rule.severity) | map({(.[0].rule.severity // "unknown"): length}) | add)
    }' > "${repo_dir}/${DATE_STAMP}_code_scanning_summary.json" 2>/dev/null || {
        log_warn "  Code scanning not accessible (may not be enabled)"
        echo '{"status": "NOT_ENABLED_OR_NO_ACCESS"}' > "${repo_dir}/${DATE_STAMP}_code_scanning_summary.json"
    }

    # Workflows (CI/CD)
    log_info "  Exporting workflow configurations..."
    gh api "/repos/${full_repo}/actions/workflows" 2>/dev/null | jq '[.workflows[] | {
        name: .name,
        path: .path,
        state: .state,
        created_at: .created_at,
        updated_at: .updated_at
    }]' > "${repo_dir}/${DATE_STAMP}_workflows.json" 2>/dev/null || {
        log_warn "  Could not export workflows"
        echo '[]' > "${repo_dir}/${DATE_STAMP}_workflows.json"
    }

    # Recent releases
    log_info "  Exporting recent releases..."
    gh api "/repos/${full_repo}/releases?per_page=10" 2>/dev/null | jq '[.[] | {
        tag_name: .tag_name,
        name: .name,
        published_at: .published_at,
        author: .author.login,
        prerelease: .prerelease,
        draft: .draft
    }]' > "${repo_dir}/${DATE_STAMP}_releases.json" 2>/dev/null || {
        echo '[]' > "${repo_dir}/${DATE_STAMP}_releases.json"
    }

    log_info "  ✓ ${repo} export complete"
}

# Generate summary report
generate_summary() {
    log_info "Generating summary report..."

    local summary_file="${OUTPUT_DIR}/${DATE_STAMP}_security_baseline_summary.json"

    cat > "${summary_file}" << EOF
{
    "export_timestamp": "${TIMESTAMP}",
    "github_owner": "${GITHUB_OWNER}",
    "repos_exported": [$(echo "${REPOS}" | tr ',' '\n' | sed 's/.*/"&"/' | tr '\n' ',' | sed 's/,$//')],
    "evidence_type": "github_security_baseline",
    "soc2_controls": ["CC6.1", "CC6.7", "CC7.1", "CC8.1"],
    "files_generated": $(find "${OUTPUT_DIR}" -name "*.json" -newer "${OUTPUT_DIR}" 2>/dev/null | wc -l | tr -d ' '),
    "notes": "Review individual repo folders for detailed exports"
}
EOF

    log_info "  ✓ Summary: ${summary_file}"
}

# Main
main() {
    log_info "=== GitHub Security Baseline Export ==="
    log_info "Timestamp: ${TIMESTAMP}"
    log_info "Owner: ${GITHUB_OWNER}"
    log_info "Repos: ${REPOS}"

    check_prereqs
    setup_output
    export_org_settings

    # Process each repo
    IFS=',' read -ra REPO_ARRAY <<< "${REPOS}"
    for repo in "${REPO_ARRAY[@]}"; do
        repo=$(echo "${repo}" | xargs)  # trim whitespace
        export_repo_settings "${repo}"
    done

    generate_summary

    echo ""
    log_info "=== Export Complete ==="
    log_info "Output: ${OUTPUT_DIR}"
    log_info "Files generated:"
    find "${OUTPUT_DIR}" -name "*.json" -exec basename {} \; | sort | uniq -c | head -20

    echo ""
    log_info "Next steps:"
    echo "  1. Review exports for completeness"
    echo "  2. Copy to evidence vault: compliance-evidence/soc2/..."
    echo "  3. Address any NOT_CONFIGURED or NOT_ENABLED findings"
}

main "$@"
