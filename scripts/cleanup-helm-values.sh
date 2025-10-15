#!/usr/bin/env bash
# Cleanup redundant Helm values files
# Based on analysis in charts/CURRENT-DEPLOYMENT-ANALYSIS.md

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHARTS_DIR="${SCRIPT_DIR}/../charts"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
  echo -e "${BLUE}==>${NC} $*"
}

success() {
  echo -e "${GREEN}✓${NC} $*"
}

error() {
  echo -e "${RED}✗${NC} $*"
}

warning() {
  echo -e "${YELLOW}!${NC} $*"
}

section() {
  echo ""
  echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
  echo -e "${BLUE}║${NC} $*"
  echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
  echo ""
}

section "Aegis Helm Values Cleanup"

log "This script will delete redundant Helm values files that are:"
echo "  - Not used by terraform/generate-helm-values.sh"
echo "  - Duplicates of files in values/ subdirectory"
echo "  - Old test artifacts"
echo ""

warning "Files to be deleted:"
echo ""
echo "aegis-services:"
echo "  - values.yaml                (388 lines of unused FedRAMP config)"
echo "  - values-local.yaml          (duplicate of values/local.yaml)"
echo "  - values-cloud.yaml          (duplicate of values/cloud.yaml)"
echo "  - values-test-cloud.yaml     (old test file)"
echo ""
echo "aegis-spoke:"
echo "  - values-cloud.yaml          (never used by deployment script)"
echo "  - values-test-cloud.yaml     (old test file)"
echo ""

read -p "Continue with cleanup? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  log "Cleanup cancelled"
  exit 0
fi

section "Deleting files"

cd "${CHARTS_DIR}"

# aegis-services cleanup
if [ -f "aegis-services/values.yaml" ]; then
  rm -v "aegis-services/values.yaml"
  success "Deleted aegis-services/values.yaml"
else
  warning "File not found: aegis-services/values.yaml"
fi

if [ -f "aegis-services/values-local.yaml" ]; then
  rm -v "aegis-services/values-local.yaml"
  success "Deleted aegis-services/values-local.yaml"
else
  warning "File not found: aegis-services/values-local.yaml"
fi

if [ -f "aegis-services/values-cloud.yaml" ]; then
  rm -v "aegis-services/values-cloud.yaml"
  success "Deleted aegis-services/values-cloud.yaml"
else
  warning "File not found: aegis-services/values-cloud.yaml"
fi

if [ -f "aegis-services/values-test-cloud.yaml" ]; then
  rm -v "aegis-services/values-test-cloud.yaml"
  success "Deleted aegis-services/values-test-cloud.yaml"
else
  warning "File not found: aegis-services/values-test-cloud.yaml"
fi

# aegis-spoke cleanup
if [ -f "aegis-spoke/values-cloud.yaml" ]; then
  rm -v "aegis-spoke/values-cloud.yaml"
  success "Deleted aegis-spoke/values-cloud.yaml"
else
  warning "File not found: aegis-spoke/values-cloud.yaml"
fi

if [ -f "aegis-spoke/values-test-cloud.yaml" ]; then
  rm -v "aegis-spoke/values-test-cloud.yaml"
  success "Deleted aegis-spoke/values-test-cloud.yaml"
else
  warning "File not found: aegis-spoke/values-test-cloud.yaml"
fi

section "Cleanup complete"

log "Remaining values files:"
echo ""
echo "aegis-services:"
find aegis-services -maxdepth 2 -name "*.yaml" -o -name "values" -type d | grep -E "(values|yaml)" | sort
echo ""
echo "aegis-spoke:"
find aegis-spoke -maxdepth 1 -name "*.yaml" | sort
echo ""

success "Cleanup successful!"
echo ""
log "Next steps:"
echo "  1. Review remaining files above"
echo "  2. Test cloud deployment: cd terraform && ./generate-helm-values.sh --tls --non-interactive"
echo "  3. Commit changes if everything works"
