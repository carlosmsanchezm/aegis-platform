#!/usr/bin/env bash
# =============================================================================
# Iron Bank Build Verification + SBOM + Grype Scan
# =============================================================================
# Builds all 6 ironbank/ Dockerfiles using the real v1.0.0 tarball,
# generates CycloneDX SBOMs via syft, and runs Grype vulnerability scans.
#
# Usage:
#   ./scripts/ironbank-build-and-scan.sh
#   ./scripts/ironbank-build-and-scan.sh --skip-build    # SBOM+scan only
#   ./scripts/ironbank-build-and-scan.sh --only platform-api k8s-agent
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
UI_ROOT="${REPO_ROOT}/../aegis-ui"
ARTIFACTS_DIR="${REPO_ROOT}/ironbank-artifacts"

# Image tag for local builds
TAG="ib-v1.0.0"

# VS Code REH pinned version
VSCODE_REH_COMMIT="ce099c1ed25d9eb3076c11e4a280f3eb52b4fbeb"

# Source tarballs (download once, reuse for all Go images)
PLATFORM_TARBALL="/tmp/aegis-platform-src.tar.gz"
UI_TARBALL="/tmp/aegis-ui-src.tar.gz"
PLATFORM_TARBALL_URL="https://github.com/carlosmsanchezm/aegis-platform/archive/refs/tags/v1.0.0.tar.gz"
UI_TARBALL_URL="https://github.com/carlosmsanchezm/aegis-ui/archive/refs/tags/v1.0.0.tar.gz"

# Public base image overrides (can't access registry1.dso.mil locally)
IB_GO_ARGS=(
  --build-arg BUILDER_REGISTRY=docker.io
  --build-arg BUILDER_IMAGE=library/golang
  --build-arg BUILDER_TAG=1.24
  --build-arg BASE_REGISTRY=registry.access.redhat.com
  --build-arg BASE_IMAGE=ubi9/ubi-minimal
  --build-arg BASE_TAG=9.7
)

IB_UBI9_ARGS=(
  --build-arg BASE_REGISTRY=registry.access.redhat.com
  --build-arg BASE_IMAGE=ubi9/ubi-minimal
  --build-arg BASE_TAG=9.7
)

IB_CUDA_ARGS=(
  --build-arg BASE_REGISTRY=docker.io
  --build-arg BASE_IMAGE=nvidia/cuda
  --build-arg BASE_TAG=12.6.1-runtime-ubi9
)

IB_NODE_ARGS=(
  --build-arg BASE_REGISTRY=docker.io
  --build-arg BASE_IMAGE=library/node
  --build-arg BASE_TAG=22-bookworm-slim
  --build-arg BUILDER_REGISTRY=docker.io
  --build-arg BUILDER_IMAGE=library/node
  --build-arg BUILDER_TAG=22-bookworm-slim
)

# ---------------------------------------------------------------------------
# Colors
# ---------------------------------------------------------------------------
if [ -t 1 ]; then
  RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
  BLUE='\033[0;34m'; BOLD='\033[1m'; RESET='\033[0m'
else
  RED='' GREEN='' YELLOW='' BLUE='' BOLD='' RESET=''
fi

info()  { printf "${BLUE}[INFO]${RESET}  %s\n" "$*"; }
ok()    { printf "${GREEN}[OK]${RESET}    %s\n" "$*"; }
warn()  { printf "${YELLOW}[WARN]${RESET}  %s\n" "$*"; }
err()   { printf "${RED}[ERROR]${RESET} %s\n" "$*" >&2; }
die()   { err "$@"; exit 1; }
step()  { printf "\n${BOLD}==> %s${RESET}\n" "$*"; }

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
SKIP_BUILD=false
ONLY_IMAGES=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-build) SKIP_BUILD=true; shift ;;
    --only) shift; while [[ $# -gt 0 ]] && [[ "$1" != --* ]]; do ONLY_IMAGES+=("$1"); shift; done ;;
    *) die "Unknown option: $1" ;;
  esac
done

should_process() {
  local name="$1"
  if [ ${#ONLY_IMAGES[@]} -eq 0 ]; then return 0; fi
  for img in "${ONLY_IMAGES[@]}"; do
    if [ "$img" = "$name" ]; then return 0; fi
  done
  return 1
}

# ---------------------------------------------------------------------------
# Prerequisites
# ---------------------------------------------------------------------------
step "Checking prerequisites"

command -v docker >/dev/null 2>&1 || die "docker not found"
command -v syft >/dev/null 2>&1   || die "syft not found (brew install syft)"
command -v grype >/dev/null 2>&1  || die "grype not found (brew install grype)"

ok "docker, syft, grype available"

mkdir -p "$ARTIFACTS_DIR"

# ---------------------------------------------------------------------------
# Download source tarballs (once)
# ---------------------------------------------------------------------------
step "Downloading source tarballs"

if [ ! -f "$PLATFORM_TARBALL" ]; then
  info "Downloading aegis-platform v1.0.0 tarball..."
  curl -fsSL -o "$PLATFORM_TARBALL" "$PLATFORM_TARBALL_URL"
  ok "Downloaded: $PLATFORM_TARBALL"
else
  ok "Already cached: $PLATFORM_TARBALL"
fi

if [ ! -f "$UI_TARBALL" ]; then
  info "Downloading aegis-ui v1.0.0 tarball..."
  curl -fsSL -o "$UI_TARBALL" "$UI_TARBALL_URL"
  ok "Downloaded: $UI_TARBALL"
else
  ok "Already cached: $UI_TARBALL"
fi

# ---------------------------------------------------------------------------
# Helper: build + sbom + scan
# ---------------------------------------------------------------------------
build_and_scan() {
  local name="$1"
  local image_tag="aegis-ib-${name}:${TAG}"
  local build_ctx="$2"
  local dockerfile="${build_ctx}/Dockerfile"
  shift 2
  local build_args=("$@")

  step "Processing: ${name}"

  if [ "$SKIP_BUILD" = false ]; then
    info "Building ${image_tag}..."
    docker build --platform linux/amd64 \
      "${build_args[@]}" \
      --build-arg FIPS_ENABLED=false \
      -f "$dockerfile" \
      -t "$image_tag" \
      "$build_ctx" || { err "Build failed for ${name}"; return 1; }
    ok "Built: ${image_tag}"
  else
    info "Skipping build (--skip-build)"
  fi

  info "Generating SBOM (CycloneDX JSON)..."
  syft "$image_tag" -o cyclonedx-json > "${ARTIFACTS_DIR}/sbom-${name}.json" 2>/dev/null \
    || { err "SBOM generation failed for ${name}"; return 1; }
  ok "SBOM: ironbank-artifacts/sbom-${name}.json ($(wc -c < "${ARTIFACTS_DIR}/sbom-${name}.json" | tr -d ' ') bytes)"

  info "Running Grype vulnerability scan..."
  grype "$image_tag" -o json > "${ARTIFACTS_DIR}/scan-${name}.json" 2>/dev/null \
    || { err "Grype scan failed for ${name}"; return 1; }
  ok "Scan: ironbank-artifacts/scan-${name}.json ($(wc -c < "${ARTIFACTS_DIR}/scan-${name}.json" | tr -d ' ') bytes)"
}

# ---------------------------------------------------------------------------
# 1. platform-api
# ---------------------------------------------------------------------------
if should_process "platform-api"; then
  TMPCTX="$(mktemp -d)"
  trap "rm -rf $TMPCTX" EXIT

  cp -r "${REPO_ROOT}/ironbank/platform-api/"* "$TMPCTX/"
  cp "$PLATFORM_TARBALL" "${TMPCTX}/aegis-platform-src.tar.gz"

  # Stage runtime dependencies
  info "Staging platform-api dependencies..."
  curl -fsSL "https://dl.k8s.io/release/v1.33.0/bin/linux/amd64/kubectl" -o "${TMPCTX}/kubectl"
  curl -fsSL "https://github.com/jqlang/jq/releases/download/jq-1.8.0/jq-linux-amd64" -o "${TMPCTX}/jq"
  curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "${TMPCTX}/awscli.zip"
  curl -fsSL "https://get.pulumi.com/releases/sdk/pulumi-v3.226.0-linux-x64.tar.gz" -o "${TMPCTX}/pulumi-linux-x64.tar.gz"

  build_and_scan "platform-api" "$TMPCTX" "${IB_GO_ARGS[@]}"
  rm -rf "$TMPCTX"
  trap - EXIT
fi

# ---------------------------------------------------------------------------
# 2. k8s-agent
# ---------------------------------------------------------------------------
if should_process "k8s-agent"; then
  TMPCTX="$(mktemp -d)"

  cp -r "${REPO_ROOT}/ironbank/k8s-agent/"* "$TMPCTX/"
  cp "$PLATFORM_TARBALL" "${TMPCTX}/aegis-platform-src.tar.gz"

  build_and_scan "k8s-agent" "$TMPCTX" "${IB_GO_ARGS[@]}"
  rm -rf "$TMPCTX"
fi

# ---------------------------------------------------------------------------
# 3. proxy
# ---------------------------------------------------------------------------
if should_process "proxy"; then
  TMPCTX="$(mktemp -d)"

  cp -r "${REPO_ROOT}/ironbank/proxy/"* "$TMPCTX/"
  cp "$PLATFORM_TARBALL" "${TMPCTX}/aegis-platform-src.tar.gz"

  build_and_scan "proxy" "$TMPCTX" "${IB_GO_ARGS[@]}"
  rm -rf "$TMPCTX"
fi

# ---------------------------------------------------------------------------
# 4. workspace
# ---------------------------------------------------------------------------
if should_process "workspace"; then
  TMPCTX="$(mktemp -d)"

  cp -r "${REPO_ROOT}/ironbank/workspace/"* "$TMPCTX/"

  info "Staging tini..."
  curl -fsSL "https://github.com/krallin/tini/releases/download/v0.19.0/tini-amd64" \
    -o "${TMPCTX}/tini-amd64"

  build_and_scan "workspace" "$TMPCTX" "${IB_CUDA_ARGS[@]}"
  rm -rf "$TMPCTX"
fi

# ---------------------------------------------------------------------------
# 5. vscode-reh-init
# ---------------------------------------------------------------------------
if should_process "vscode-reh-init"; then
  TMPCTX="$(mktemp -d)"

  cp -r "${REPO_ROOT}/ironbank/vscode-reh-init/"* "$TMPCTX/"

  info "Staging VS Code server (commit: ${VSCODE_REH_COMMIT:0:12}...)..."
  curl -fsSL "https://vscode.download.prss.microsoft.com/dbazure/download/stable/${VSCODE_REH_COMMIT}/vscode-server-linux-x64.tar.gz" \
    -o "${TMPCTX}/vscode-server-linux-x64.tar.gz"

  build_and_scan "vscode-reh-init" "$TMPCTX" "${IB_UBI9_ARGS[@]}"
  rm -rf "$TMPCTX"
fi

# ---------------------------------------------------------------------------
# 6. ui
# ---------------------------------------------------------------------------
if should_process "ui"; then
  TMPCTX="$(mktemp -d)"

  cp -r "${UI_ROOT}/ironbank/"* "$TMPCTX/"
  cp "$UI_TARBALL" "${TMPCTX}/aegis-ui-src.tar.gz"

  build_and_scan "ui" "$TMPCTX" "${IB_NODE_ARGS[@]}"
  rm -rf "$TMPCTX"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
step "Summary"

printf "\n"
printf "  %-20s %-12s %-12s\n" "Image" "SBOM" "Scan"
printf "  %-20s %-12s %-12s\n" "-----" "----" "----"

for name in platform-api k8s-agent proxy workspace vscode-reh-init ui; do
  sbom_status="--"
  scan_status="--"
  if [ -f "${ARTIFACTS_DIR}/sbom-${name}.json" ] && [ -s "${ARTIFACTS_DIR}/sbom-${name}.json" ]; then
    sbom_status="OK"
  fi
  if [ -f "${ARTIFACTS_DIR}/scan-${name}.json" ] && [ -s "${ARTIFACTS_DIR}/scan-${name}.json" ]; then
    scan_status="OK"
  fi
  printf "  %-20s %-12s %-12s\n" "$name" "$sbom_status" "$scan_status"
done

printf "\n"
ok "All artifacts in: ironbank-artifacts/"
printf "\n"
