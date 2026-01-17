#!/bin/bash
# generate-sbom.sh - Generate SBOMs for all Aegis components
#
# Prerequisites:
#   npm install -g @cyclonedx/cdxgen
#   go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest
#   curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
#
# Usage:
#   ./generate-sbom.sh [--container-images]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION="${VERSION:-$(git describe --tags --always 2>/dev/null || echo 'dev')}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check prerequisites
check_prereqs() {
    local missing=()

    if ! command -v cyclonedx-gomod &> /dev/null; then
        missing+=("cyclonedx-gomod")
    fi

    if ! command -v cdxgen &> /dev/null; then
        missing+=("cdxgen (@cyclonedx/cdxgen)")
    fi

    if [[ "${1:-}" == "--container-images" ]] && ! command -v syft &> /dev/null; then
        missing+=("syft")
    fi

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing prerequisites: ${missing[*]}"
        echo ""
        echo "Install with:"
        echo "  npm install -g @cyclonedx/cdxgen"
        echo "  go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest"
        echo "  curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin"
        exit 1
    fi
}

# Generate SBOM for Go project
generate_go_sbom() {
    local project_path=$1
    local output_name=$2

    if [[ ! -f "${project_path}/go.mod" ]]; then
        log_warn "No go.mod found at ${project_path}, skipping"
        return 0
    fi

    log_info "Generating SBOM for ${output_name} (Go)..."

    (
        cd "${project_path}"
        cyclonedx-gomod mod \
            -json \
            -output "${OUTPUT_DIR}/${output_name}-sbom.json" \
            2>/dev/null || {
                log_warn "cyclonedx-gomod failed for ${output_name}, trying alternative..."
                # Fallback to cdxgen for Go
                cdxgen \
                    -o "${OUTPUT_DIR}/${output_name}-sbom.json" \
                    --format json \
                    . 2>/dev/null || log_error "Failed to generate SBOM for ${output_name}"
            }
    )

    if [[ -f "${OUTPUT_DIR}/${output_name}-sbom.json" ]]; then
        log_info "  ✓ ${output_name}-sbom.json"
    fi
}

# Generate SBOM for Node.js project
generate_node_sbom() {
    local project_path=$1
    local output_name=$2

    if [[ ! -f "${project_path}/package.json" ]]; then
        log_warn "No package.json found at ${project_path}, skipping"
        return 0
    fi

    log_info "Generating SBOM for ${output_name} (Node.js)..."

    (
        cd "${project_path}"
        cdxgen \
            -o "${OUTPUT_DIR}/${output_name}-sbom.json" \
            --format json \
            . 2>/dev/null || log_error "Failed to generate SBOM for ${output_name}"
    )

    if [[ -f "${OUTPUT_DIR}/${output_name}-sbom.json" ]]; then
        log_info "  ✓ ${output_name}-sbom.json"
    fi
}

# Generate SBOM for container image
generate_container_sbom() {
    local image=$1
    local output_name=$2

    log_info "Generating SBOM for ${image} (container)..."

    syft "${image}" \
        -o cyclonedx-json="${OUTPUT_DIR}/${output_name}-container-sbom.json" \
        2>/dev/null || {
            log_warn "Failed to generate container SBOM for ${image}"
            return 0
        }

    if [[ -f "${OUTPUT_DIR}/${output_name}-container-sbom.json" ]]; then
        log_info "  ✓ ${output_name}-container-sbom.json"
    fi
}

# Main
main() {
    log_info "Aegis SBOM Generator"
    log_info "Version: ${VERSION}"
    log_info "Timestamp: ${TIMESTAMP}"
    log_info "Output directory: ${OUTPUT_DIR}"
    echo ""

    check_prereqs "${1:-}"

    # Create output directory
    mkdir -p "${OUTPUT_DIR}"

    # Generate SBOMs for Go services
    log_info "=== Go Services ==="
    generate_go_sbom "${REPO_ROOT}/services/platform-api" "platform-api"
    generate_go_sbom "${REPO_ROOT}/services/k8s-agent" "k8s-agent"
    generate_go_sbom "${REPO_ROOT}/services/spoke-proxy" "spoke-proxy"

    # Generate SBOM for Node.js services
    log_info ""
    log_info "=== Node.js Services ==="
    # Adjust path based on your actual Backstage location
    if [[ -d "${REPO_ROOT}/packages/backend" ]]; then
        generate_node_sbom "${REPO_ROOT}/packages/backend" "backstage"
    elif [[ -d "${REPO_ROOT}/services/backstage" ]]; then
        generate_node_sbom "${REPO_ROOT}/services/backstage" "backstage"
    else
        log_warn "Backstage project not found, skipping"
    fi

    # Generate container image SBOMs if requested
    if [[ "${1:-}" == "--container-images" ]]; then
        log_info ""
        log_info "=== Container Images ==="
        log_warn "Container image SBOMs require images to be built and accessible"
        log_warn "Set REGISTRY environment variable or update script with your registry"

        REGISTRY="${REGISTRY:-your-registry}"

        # Uncomment and adjust as needed
        # generate_container_sbom "${REGISTRY}/aegis-platform-api:${VERSION}" "platform-api"
        # generate_container_sbom "${REGISTRY}/aegis-backstage:${VERSION}" "backstage"
        # generate_container_sbom "${REGISTRY}/aegis-k8s-agent:${VERSION}" "k8s-agent"
        # generate_container_sbom "${REGISTRY}/aegis-spoke-proxy:${VERSION}" "spoke-proxy"
    fi

    # Summary
    echo ""
    log_info "=== Summary ==="
    local count=$(find "${OUTPUT_DIR}" -name "*-sbom.json" -newer "${OUTPUT_DIR}/README.md" 2>/dev/null | wc -l | tr -d ' ')
    log_info "Generated ${count} SBOM(s)"

    if [[ -n "$(find "${OUTPUT_DIR}" -name "*-sbom.json" -newer "${OUTPUT_DIR}/README.md" 2>/dev/null)" ]]; then
        log_info "Files:"
        find "${OUTPUT_DIR}" -name "*-sbom.json" -newer "${OUTPUT_DIR}/README.md" -exec basename {} \; | sed 's/^/  - /'
    fi

    echo ""
    log_info "To scan for vulnerabilities:"
    echo "  grype sbom:${OUTPUT_DIR}/platform-api-sbom.json"
    echo "  trivy sbom ${OUTPUT_DIR}/platform-api-sbom.json"
}

main "$@"
