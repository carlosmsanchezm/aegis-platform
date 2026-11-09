#!/bin/bash
# export_ci_reports.sh
#
# SOC 2 Evidence Collection Script - CI/CD Security Reports
# Exports SBOM, dependency scans, and container scan summaries
#
# Evidence Maps to:
# - CC7.1: Vulnerability detection
# - CC8.1: Change management (CI/CD)
# - CC5.2: Technology controls
#
# Usage: ./export_ci_reports.sh [output_dir]
#
# Prerequisites:
# - syft (for SBOM generation)
# - grype or trivy (for vulnerability scanning)
# - Access to container images (if scanning containers)

set -euo pipefail

# Configuration
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"

# Evidence vault location (use env var or argument or default)
EVIDENCE_VAULT="${EVIDENCE_VAULT:-$HOME/code/aegis-compliance-evidence}"
OUTPUT_DIR="${1:-$EVIDENCE_VAULT/soc2/$(date +%Y)/$(date +%Y-%m)/vuln-management}"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
DATE_STAMP=$(date +"%Y-%m-%d")

# Project configuration
GO_MODULES="${GO_MODULES:-services/platform-api,services/proxy,agents/k8s-agent}"
NODE_PROJECTS="${NODE_PROJECTS:-}"  # Add paths to Node.js projects if any

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

    local tools_available=()
    local tools_missing=()

    # Check for SBOM tools
    if command -v syft &> /dev/null; then
        tools_available+=("syft")
    elif command -v cyclonedx-gomod &> /dev/null; then
        tools_available+=("cyclonedx-gomod")
    else
        tools_missing+=("syft or cyclonedx-gomod (SBOM)")
    fi

    # Check for vulnerability scanners
    if command -v grype &> /dev/null; then
        tools_available+=("grype")
    elif command -v trivy &> /dev/null; then
        tools_available+=("trivy")
    else
        tools_missing+=("grype or trivy (vuln scan)")
    fi

    # Check for jq
    if ! command -v jq &> /dev/null; then
        tools_missing+=("jq")
    fi

    log_info "  Available: ${tools_available[*]:-none}"

    if [[ ${#tools_missing[@]} -gt 0 ]]; then
        log_warn "  Missing (optional): ${tools_missing[*]}"
        echo ""
        echo "Install with:"
        echo "  brew install syft grype jq"
        echo "  # or"
        echo "  brew install trivy jq"
        echo ""
    fi
}

# Create output directory
setup_output() {
    mkdir -p "${OUTPUT_DIR}/sbom"
    mkdir -p "${OUTPUT_DIR}/vuln-scans"
    log_info "Output directory: ${OUTPUT_DIR}"
}

# Generate SBOM for Go module
generate_go_sbom() {
    local module_path=$1
    local module_name
    module_name=$(basename "${module_path}")

    local full_path="${REPO_ROOT}/${module_path}"

    if [[ ! -f "${full_path}/go.mod" ]]; then
        log_warn "  No go.mod at ${full_path}, skipping"
        return 0
    fi

    log_info "  Generating SBOM for ${module_name}..."

    local sbom_file="${OUTPUT_DIR}/sbom/${DATE_STAMP}_${module_name}_sbom.json"

    if command -v syft &> /dev/null; then
        syft dir:"${full_path}" -o cyclonedx-json="${sbom_file}" 2>/dev/null || {
            log_warn "    syft failed, trying cyclonedx-gomod..."
            (cd "${full_path}" && cyclonedx-gomod mod -json -output "${sbom_file}" 2>/dev/null) || {
                log_warn "    SBOM generation failed for ${module_name}"
                return 0
            }
        }
    elif command -v cyclonedx-gomod &> /dev/null; then
        (cd "${full_path}" && cyclonedx-gomod mod -json -output "${sbom_file}" 2>/dev/null) || {
            log_warn "    SBOM generation failed for ${module_name}"
            return 0
        }
    fi

    if [[ -f "${sbom_file}" ]]; then
        local component_count
        component_count=$(jq '.components | length' "${sbom_file}" 2>/dev/null || echo "0")
        log_info "    ✓ ${module_name}: ${component_count} components"
    fi
}

# Generate SBOM for Node.js project
generate_node_sbom() {
    local project_path=$1
    local project_name
    project_name=$(basename "${project_path}")

    local full_path="${REPO_ROOT}/${project_path}"

    if [[ ! -f "${full_path}/package.json" ]]; then
        log_warn "  No package.json at ${full_path}, skipping"
        return 0
    fi

    log_info "  Generating SBOM for ${project_name}..."

    local sbom_file="${OUTPUT_DIR}/sbom/${DATE_STAMP}_${project_name}_sbom.json"

    if command -v syft &> /dev/null; then
        syft dir:"${full_path}" -o cyclonedx-json="${sbom_file}" 2>/dev/null || {
            log_warn "    SBOM generation failed for ${project_name}"
            return 0
        }
    fi

    if [[ -f "${sbom_file}" ]]; then
        local component_count
        component_count=$(jq '.components | length' "${sbom_file}" 2>/dev/null || echo "0")
        log_info "    ✓ ${project_name}: ${component_count} components"
    fi
}

# Scan SBOM for vulnerabilities
scan_sbom() {
    local sbom_file=$1
    local base_name
    base_name=$(basename "${sbom_file}" .json)

    log_info "  Scanning ${base_name}..."

    local scan_file="${OUTPUT_DIR}/vuln-scans/${base_name}_vulns.json"

    if command -v grype &> /dev/null; then
        grype sbom:"${sbom_file}" -o json > "${scan_file}" 2>/dev/null || {
            log_warn "    Grype scan failed"
            return 0
        }

        # Generate summary
        if [[ -f "${scan_file}" ]]; then
            jq '{
                scan_timestamp: "'"${TIMESTAMP}"'",
                source: "'"${base_name}"'",
                total_vulnerabilities: (.matches | length),
                by_severity: (.matches | group_by(.vulnerability.severity) | map({(.[0].vulnerability.severity // "Unknown"): length}) | add),
                critical_vulns: [.matches[] | select(.vulnerability.severity == "Critical") | {id: .vulnerability.id, package: .artifact.name, version: .artifact.version, fix: .vulnerability.fix.versions[0]}],
                high_vulns: [.matches[] | select(.vulnerability.severity == "High") | {id: .vulnerability.id, package: .artifact.name, version: .artifact.version, fix: .vulnerability.fix.versions[0]}] | .[0:10]
            }' "${scan_file}" > "${OUTPUT_DIR}/vuln-scans/${base_name}_summary.json"

            local total_vulns critical_count high_count
            total_vulns=$(jq '.matches | length' "${scan_file}")
            critical_count=$(jq '[.matches[] | select(.vulnerability.severity == "Critical")] | length' "${scan_file}")
            high_count=$(jq '[.matches[] | select(.vulnerability.severity == "High")] | length' "${scan_file}")

            log_info "    ✓ ${total_vulns} vulns (${critical_count} critical, ${high_count} high)"
        fi

    elif command -v trivy &> /dev/null; then
        trivy sbom "${sbom_file}" --format json --output "${scan_file}" 2>/dev/null || {
            log_warn "    Trivy scan failed"
            return 0
        }

        if [[ -f "${scan_file}" ]]; then
            log_info "    ✓ Scan complete (see ${scan_file})"
        fi
    else
        log_warn "    No vulnerability scanner available"
    fi
}

# Export go.sum for dependency verification
export_go_checksums() {
    log_info "Exporting Go module checksums..."

    for module_path in ${GO_MODULES//,/ }; do
        local full_path="${REPO_ROOT}/${module_path}"
        local module_name
        module_name=$(basename "${module_path}")

        if [[ -f "${full_path}/go.sum" ]]; then
            cp "${full_path}/go.sum" "${OUTPUT_DIR}/sbom/${DATE_STAMP}_${module_name}_go.sum"
            log_info "  ✓ ${module_name}/go.sum"
        fi
    done
}

# Export GitHub Actions workflow security
export_workflow_security() {
    log_info "Analyzing GitHub Actions workflows..."

    local workflows_dir="${REPO_ROOT}/.github/workflows"
    local output_file="${OUTPUT_DIR}/${DATE_STAMP}_workflow_analysis.json"

    if [[ ! -d "${workflows_dir}" ]]; then
        log_warn "  No workflows directory found"
        echo '{"status": "NO_WORKFLOWS"}' > "${output_file}"
        return 0
    fi

    # Analyze each workflow
    echo "[" > "${output_file}"
    local first=true

    for workflow in "${workflows_dir}"/*.yml "${workflows_dir}"/*.yaml; do
        [[ -f "${workflow}" ]] || continue

        if [[ "${first}" != "true" ]]; then
            echo "," >> "${output_file}"
        fi
        first=false

        local workflow_name
        workflow_name=$(basename "${workflow}")

        # Check for security-relevant patterns
        local uses_secrets=false
        local pins_actions=false
        local has_permissions=false

        grep -q '\${{ secrets\.' "${workflow}" 2>/dev/null && uses_secrets=true
        grep -qE '@[a-f0-9]{40}' "${workflow}" 2>/dev/null && pins_actions=true
        grep -q 'permissions:' "${workflow}" 2>/dev/null && has_permissions=true

        cat >> "${output_file}" << EOF
  {
    "workflow": "${workflow_name}",
    "uses_secrets": ${uses_secrets},
    "pins_action_versions": ${pins_actions},
    "defines_permissions": ${has_permissions}
  }
EOF
    done

    echo "]" >> "${output_file}"

    log_info "  ✓ Workflow analysis complete"
}

# Generate summary report
generate_summary() {
    log_info "Generating summary report..."

    local summary_file="${OUTPUT_DIR}/${DATE_STAMP}_ci_security_summary.json"

    # Count SBOMs and scans
    local sbom_count vuln_scan_count total_vulns critical_vulns

    sbom_count=$(find "${OUTPUT_DIR}/sbom" -name "*_sbom.json" 2>/dev/null | wc -l | tr -d ' ')
    vuln_scan_count=$(find "${OUTPUT_DIR}/vuln-scans" -name "*_vulns.json" 2>/dev/null | wc -l | tr -d ' ')

    # Aggregate vulnerability counts
    total_vulns=0
    critical_vulns=0

    for summary in "${OUTPUT_DIR}"/vuln-scans/*_summary.json; do
        [[ -f "${summary}" ]] || continue
        local file_total file_critical
        file_total=$(jq '.total_vulnerabilities // 0' "${summary}" 2>/dev/null || echo "0")
        file_critical=$(jq '.by_severity.Critical // 0' "${summary}" 2>/dev/null || echo "0")
        total_vulns=$((total_vulns + file_total))
        critical_vulns=$((critical_vulns + file_critical))
    done

    cat > "${summary_file}" << EOF
{
    "export_timestamp": "${TIMESTAMP}",
    "repository_root": "${REPO_ROOT}",
    "evidence_type": "ci_security_reports",
    "soc2_controls": ["CC7.1", "CC8.1", "CC5.2"],
    "summary": {
        "sboms_generated": ${sbom_count},
        "vulnerability_scans": ${vuln_scan_count},
        "total_vulnerabilities": ${total_vulns},
        "critical_vulnerabilities": ${critical_vulns}
    },
    "modules_scanned": [$(echo "${GO_MODULES}" | tr ',' '\n' | sed 's/.*/"&"/' | tr '\n' ',' | sed 's/,$//')],
    "recommendations": [
        $([ "${critical_vulns}" -gt 0 ] && echo '"Address critical vulnerabilities within 7 days",' || echo "")
        $([ "${sbom_count}" -eq 0 ] && echo '"Install syft/cyclonedx-gomod to generate SBOMs",' || echo "")
        "Review vulnerability summaries for remediation priorities"
    ]
}
EOF

    log_info "  ✓ Summary: ${summary_file}"
}

# Main
main() {
    log_info "=== CI/CD Security Reports Export ==="
    log_info "Timestamp: ${TIMESTAMP}"
    log_info "Repository: ${REPO_ROOT}"

    check_prereqs
    setup_output

    # Generate SBOMs
    log_info "Generating SBOMs..."
    for module in ${GO_MODULES//,/ }; do
        generate_go_sbom "${module}"
    done

    for project in ${NODE_PROJECTS//,/ }; do
        [[ -n "${project}" ]] && generate_node_sbom "${project}"
    done

    # Export checksums
    export_go_checksums

    # Scan SBOMs for vulnerabilities
    log_info "Scanning for vulnerabilities..."
    for sbom in "${OUTPUT_DIR}"/sbom/*_sbom.json; do
        [[ -f "${sbom}" ]] && scan_sbom "${sbom}"
    done

    # Analyze workflows
    export_workflow_security

    # Generate summary
    generate_summary

    echo ""
    log_info "=== Export Complete ==="
    log_info "Output: ${OUTPUT_DIR}"
    log_info ""
    log_info "SBOMs: ${OUTPUT_DIR}/sbom/"
    log_info "Vulnerability Scans: ${OUTPUT_DIR}/vuln-scans/"
    log_info "Summary: ${OUTPUT_DIR}/${DATE_STAMP}_ci_security_summary.json"

    echo ""
    log_info "Next steps:"
    echo "  1. Review vulnerability summaries"
    echo "  2. Create tickets for critical/high vulnerabilities"
    echo "  3. Archive reports in evidence vault"
}

main "$@"
