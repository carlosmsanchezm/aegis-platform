#!/usr/bin/env bash
# Local test runner that mirrors the GitHub Actions workflow
# Run all unit, integration, and E2E tests locally

set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
RUN_LINT="${RUN_LINT:-1}"
RUN_GO_TESTS="${RUN_GO_TESTS:-1}"
RUN_FRONTEND_TESTS="${RUN_FRONTEND_TESTS:-1}"
RUN_E2E_PLATFORM="${RUN_E2E_PLATFORM:-0}"  # Requires deployed environment
RUN_E2E_OPERATOR="${RUN_E2E_OPERATOR:-0}"  # Requires cluster access
USE_EXISTING_CLUSTER="${USE_EXISTING_CLUSTER:-true}"  # For operator E2E

# Track failures
FAILED_TESTS=()

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

check_dependencies() {
  section "Checking dependencies"

  local missing_deps=()

  # Go
  if ! command -v go &> /dev/null; then
    missing_deps+=("go")
  else
    log "Go version: $(go version)"
  fi

  # Node/Yarn
  if ! command -v node &> /dev/null; then
    missing_deps+=("node")
  else
    log "Node version: $(node --version)"
  fi

  if ! command -v yarn &> /dev/null; then
    missing_deps+=("yarn")
  else
    log "Yarn version: $(yarn --version)"
  fi

  # kubectl (for E2E tests)
  if [[ "${RUN_E2E_PLATFORM}" == "1" ]] || [[ "${RUN_E2E_OPERATOR}" == "1" ]]; then
    if ! command -v kubectl &> /dev/null; then
      missing_deps+=("kubectl")
    else
      log "kubectl version: $(kubectl version --client --short 2>/dev/null || echo 'unknown')"
    fi
  fi

  # grpcurl (for platform E2E)
  if [[ "${RUN_E2E_PLATFORM}" == "1" ]]; then
    if ! command -v grpcurl &> /dev/null; then
      missing_deps+=("grpcurl")
    else
      log "grpcurl version: $(grpcurl --version | head -1)"
    fi

    if ! command -v jq &> /dev/null; then
      missing_deps+=("jq")
    fi
  fi

  if [ ${#missing_deps[@]} -ne 0 ]; then
    error "Missing required dependencies: ${missing_deps[*]}"
    echo ""
    echo "Install instructions:"
    for dep in "${missing_deps[@]}"; do
      case "$dep" in
        go)
          echo "  - Go: https://golang.org/doc/install"
          ;;
        node)
          echo "  - Node.js: https://nodejs.org/"
          ;;
        yarn)
          echo "  - Yarn: npm install -g yarn"
          ;;
        kubectl)
          echo "  - kubectl: https://kubernetes.io/docs/tasks/tools/"
          ;;
        grpcurl)
          echo "  - grpcurl: brew install grpcurl (Mac) or go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
          ;;
        jq)
          echo "  - jq: brew install jq (Mac) or apt-get install jq (Linux)"
          ;;
      esac
    done
    exit 1
  fi

  success "All required dependencies are installed"
}

run_go_unit_tests() {
  section "Running Go Unit & Integration Tests"

  log "Syncing Go workspace..."
  go work sync

  # Test k8s-agent (uses envtest, excludes e2e)
  log "Testing agents/k8s-agent..."
  if (cd agents/k8s-agent && make test); then
    success "k8s-agent tests passed"
  else
    error "k8s-agent tests failed"
    FAILED_TESTS+=("k8s-agent")
  fi

  # Test each Go module
  for module in pkg/workspace services/platform-api services/proxy; do
    log "Testing ${module}..."
    if (cd "${module}" && go test ./...); then
      success "${module} tests passed"
    else
      error "${module} tests failed"
      FAILED_TESTS+=("${module}")
    fi
  done
}

run_frontend_tests() {
  section "Running Frontend Tests"

  if [[ ! -d "${REPO_ROOT}/aegis-platform" || ! -f "${REPO_ROOT}/aegis-platform/package.json" ]]; then
    warning "Frontend sources not found; skipping frontend tests"
    return
  fi

  cd "${REPO_ROOT}/aegis-platform"

  log "Installing frontend dependencies (if needed)..."
  if [ ! -d "node_modules" ]; then
    yarn install --immutable
  fi

  log "Running frontend unit tests..."
  if yarn test --watchAll=false; then
    success "Frontend tests passed"
  else
    error "Frontend tests failed"
    FAILED_TESTS+=("frontend")
  fi

  cd "${REPO_ROOT}"
}

run_frontend_lint() {
  section "Running Frontend Linting"

  if [[ ! -d "${REPO_ROOT}/aegis-platform" || ! -f "${REPO_ROOT}/aegis-platform/package.json" ]]; then
    warning "Frontend sources not found; skipping frontend lint"
    return
  fi

  cd "${REPO_ROOT}/aegis-platform"

  log "Running prettier check..."
  if yarn prettier:check; then
    success "Prettier check passed"
  else
    error "Prettier check failed"
    FAILED_TESTS+=("prettier")
  fi

  log "Running eslint..."
  if yarn lint:all; then
    success "ESLint passed"
  else
    warning "ESLint found issues"
    # Don't fail on lint warnings by default
  fi

  cd "${REPO_ROOT}"
}

run_go_lint() {
  section "Running Go Linting & Formatting"

  log "Checking Go formatting..."
  if (cd agents/k8s-agent && go fmt ./...); then
    success "k8s-agent formatting check passed"
  else
    error "k8s-agent formatting check failed"
    FAILED_TESTS+=("k8s-agent-fmt")
  fi

  log "Running go vet..."
  if (cd agents/k8s-agent && go vet ./...); then
    success "Go vet passed"
  else
    error "Go vet failed"
    FAILED_TESTS+=("go-vet")
  fi

  # Check if golangci-lint is available
  if command -v golangci-lint &> /dev/null || [ -f "agents/k8s-agent/bin/golangci-lint" ]; then
    log "Running golangci-lint..."
    if (cd agents/k8s-agent && make lint); then
      success "golangci-lint passed"
    else
      warning "golangci-lint found issues"
      # Don't fail on lint warnings by default
    fi
  else
    warning "golangci-lint not installed, skipping (optional)"
  fi
}

run_platform_e2e() {
  section "Running Platform API E2E Tests"

  warning "Platform API E2E tests require a deployed Aegis environment"
  echo "Prerequisites:"
  echo "  1. Aegis services deployed (via 'make deploy-local' or remote environment)"
  echo "  2. kubectl configured to access the cluster"
  echo "  3. Port-forwarding or DNS configured for services"
  echo ""

  # Check environment variables
  if [ -z "${AEGIS_GRPC_ADDR:-}" ]; then
    log "AEGIS_GRPC_ADDR not set, defaulting to 127.0.0.1:10081 (local port-forward)"
    export AEGIS_GRPC_ADDR="127.0.0.1:10081"
  fi

  log "Running E2E tests against: ${AEGIS_GRPC_ADDR}"
  log "Platform namespace: ${AEGIS_PLATFORM_NAMESPACE:-aegis-system}"

  if "${SCRIPT_DIR}/e2e-platform-api.sh"; then
    success "Platform API E2E tests passed"
  else
    error "Platform API E2E tests failed"
    FAILED_TESTS+=("platform-e2e")
  fi
}

run_operator_e2e() {
  section "Running K8s Operator E2E Tests"

  if [[ "${USE_EXISTING_CLUSTER}" == "true" ]]; then
    warning "Using existing cluster: $(kubectl config current-context)"
    log "The operator will be deployed to the current cluster for testing"
  else
    log "A new Kind cluster will be created for testing"
  fi

  export USE_EXISTING_CLUSTER
  export KIND_CLUSTER="${KIND_CLUSTER:-aegis-e2e}"

  log "Running operator E2E tests..."
  if (cd agents/k8s-agent && make test-e2e); then
    success "Operator E2E tests passed"
  else
    error "Operator E2E tests failed"
    FAILED_TESTS+=("operator-e2e")
  fi
}

print_summary() {
  echo ""
  echo "╔════════════════════════════════════════════════════════════════╗"
  echo "║                        TEST SUMMARY                            ║"
  echo "╚════════════════════════════════════════════════════════════════╝"
  echo ""

  if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
    success "All tests passed! 🎉"
    return 0
  else
    error "Some tests failed:"
    for test in "${FAILED_TESTS[@]}"; do
      echo "  - ${test}"
    done
    return 1
  fi
}

print_usage() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Run all Aegis tests locally (unit, integration, and E2E).

OPTIONS:
  --skip-lint              Skip linting checks
  --skip-go-tests          Skip Go unit/integration tests
  --skip-frontend-tests    Skip frontend tests
  --with-platform-e2e      Run platform API E2E tests (requires deployed env)
  --with-operator-e2e      Run k8s-agent E2E tests (requires cluster access)
  --use-kind               Create a new Kind cluster for operator E2E
  --help                   Show this help message

ENVIRONMENT VARIABLES:
  RUN_LINT                 Run linting (default: 1)
  RUN_GO_TESTS            Run Go tests (default: 1)
  RUN_FRONTEND_TESTS      Run frontend tests (default: 1)
  RUN_E2E_PLATFORM        Run platform E2E (default: 0)
  RUN_E2E_OPERATOR        Run operator E2E (default: 0)
  USE_EXISTING_CLUSTER    Use current kubectl context (default: true)
  AEGIS_GRPC_ADDR         Platform API address for E2E tests
  AEGIS_PLATFORM_NAMESPACE Platform namespace (default: aegis-system)

EXAMPLES:
  # Run all unit and integration tests (default)
  ./scripts/test-all-local.sh

  # Run tests including platform E2E (requires deployed environment)
  ./scripts/test-all-local.sh --with-platform-e2e

  # Run operator E2E tests using current cluster
  ./scripts/test-all-local.sh --with-operator-e2e

  # Run operator E2E tests in a new Kind cluster
  ./scripts/test-all-local.sh --with-operator-e2e --use-kind

  # Run everything (requires deployed environment + cluster)
  ./scripts/test-all-local.sh --with-platform-e2e --with-operator-e2e

  # Skip frontend tests, run only Go tests
  ./scripts/test-all-local.sh --skip-frontend-tests

  # Match GitHub Actions CI (lint + unit tests only)
  ./scripts/test-all-local.sh

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --skip-lint)
      RUN_LINT=0
      shift
      ;;
    --skip-go-tests)
      RUN_GO_TESTS=0
      shift
      ;;
    --skip-frontend-tests)
      RUN_FRONTEND_TESTS=0
      shift
      ;;
    --with-platform-e2e)
      RUN_E2E_PLATFORM=1
      shift
      ;;
    --with-operator-e2e)
      RUN_E2E_OPERATOR=1
      shift
      ;;
    --use-kind)
      USE_EXISTING_CLUSTER=false
      shift
      ;;
    --help|-h)
      print_usage
      exit 0
      ;;
    *)
      error "Unknown option: $1"
      print_usage
      exit 1
      ;;
  esac
done

# Main execution
main() {
  section "Aegis Local Test Runner"
  log "Repository: ${REPO_ROOT}"
  log "Configuration:"
  echo "  - Lint: ${RUN_LINT}"
  echo "  - Go tests: ${RUN_GO_TESTS}"
  echo "  - Frontend tests: ${RUN_FRONTEND_TESTS}"
  echo "  - Platform E2E: ${RUN_E2E_PLATFORM}"
  echo "  - Operator E2E: ${RUN_E2E_OPERATOR}"

  check_dependencies

  # Run tests in order (matching GitHub Actions workflow)
  if [[ "${RUN_LINT}" == "1" ]]; then
    run_go_lint
    run_frontend_lint
  fi

  if [[ "${RUN_GO_TESTS}" == "1" ]]; then
    run_go_unit_tests
  fi

  if [[ "${RUN_FRONTEND_TESTS}" == "1" ]]; then
    run_frontend_tests
  fi

  if [[ "${RUN_E2E_PLATFORM}" == "1" ]]; then
    run_platform_e2e
  fi

  if [[ "${RUN_E2E_OPERATOR}" == "1" ]]; then
    run_operator_e2e
  fi

  print_summary
}

# Run main and capture exit status
if main; then
  exit 0
else
  exit 1
fi
