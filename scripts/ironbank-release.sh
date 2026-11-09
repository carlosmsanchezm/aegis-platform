#!/usr/bin/env bash
# =============================================================================
# Iron Bank Release Helper
# =============================================================================
# Creates a tagged release, downloads tarballs, computes sha256 hashes,
# and updates all hardening_manifest.yaml files in ironbank/ and aegis-ui/ironbank/.
#
# Usage:
#   ./scripts/ironbank-release.sh v1.0.0
#   ./scripts/ironbank-release.sh v1.0.0 --dry-run
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
UI_ROOT="${REPO_ROOT}/../aegis-ui"

PLATFORM_REPO="carlosmsanchezm/aegis-platform"
UI_REPO="carlosmsanchezm/aegis-ui"

# ---------------------------------------------------------------------------
# Color helpers
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

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
VERSION="${1:-}"
DRY_RUN=false

if [ -z "$VERSION" ]; then
  die "Usage: $0 <version-tag> [--dry-run]"
fi

if [ "${2:-}" = "--dry-run" ]; then
  DRY_RUN=true
  warn "Dry run mode — no tags will be created or pushed"
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  die "Version must match pattern vX.Y.Z (e.g., v1.0.0). Got: $VERSION"
fi

VERSION_NUM="${VERSION#v}"  # Strip 'v' prefix for manifest tags

printf "\n${BOLD}Iron Bank Release: ${VERSION}${RESET}\n\n"

# ---------------------------------------------------------------------------
# Step 1: Create and push tags
# ---------------------------------------------------------------------------
info "Step 1: Creating tags..."

if [ "$DRY_RUN" = false ]; then
  cd "$REPO_ROOT"
  if git rev-parse "$VERSION" >/dev/null 2>&1; then
    warn "Tag $VERSION already exists in aegis-platform"
  else
    git tag "$VERSION"
    git push origin "$VERSION"
    ok "Created and pushed tag $VERSION for aegis-platform"
  fi

  if [ -d "$UI_ROOT" ]; then
    cd "$UI_ROOT"
    if git rev-parse "$VERSION" >/dev/null 2>&1; then
      warn "Tag $VERSION already exists in aegis-ui"
    else
      git tag "$VERSION"
      git push origin "$VERSION"
      ok "Created and pushed tag $VERSION for aegis-ui"
    fi
  else
    warn "aegis-ui repo not found at $UI_ROOT — skipping UI tag"
  fi
else
  info "[dry-run] Would create tag $VERSION in both repos"
fi

# ---------------------------------------------------------------------------
# Step 2: Download tarballs and compute hashes
# ---------------------------------------------------------------------------
info "Step 2: Downloading tarballs and computing sha256 hashes..."

PLATFORM_URL="https://github.com/${PLATFORM_REPO}/archive/refs/tags/${VERSION}.tar.gz"
UI_URL="https://github.com/${UI_REPO}/archive/refs/tags/${VERSION}.tar.gz"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL -o "${TMPDIR}/aegis-platform-src.tar.gz" "$PLATFORM_URL" \
  || die "Failed to download platform tarball from $PLATFORM_URL"
PLATFORM_HASH="$(shasum -a 256 "${TMPDIR}/aegis-platform-src.tar.gz" | awk '{print $1}')"
ok "aegis-platform hash: $PLATFORM_HASH"

curl -fsSL -o "${TMPDIR}/aegis-ui-src.tar.gz" "$UI_URL" \
  || die "Failed to download UI tarball from $UI_URL"
UI_HASH="$(shasum -a 256 "${TMPDIR}/aegis-ui-src.tar.gz" | awk '{print $1}')"
ok "aegis-ui hash:       $UI_HASH"

# ---------------------------------------------------------------------------
# Step 3: Update hardening_manifest.yaml files
# ---------------------------------------------------------------------------
info "Step 3: Updating hardening_manifest.yaml files..."

update_manifest() {
  local file="$1"
  local old_hash="$2"
  local new_hash="$3"
  local old_url_version="$4"
  local new_url_version="$5"
  local old_tag="$6"
  local new_tag="$7"

  if [ ! -f "$file" ]; then
    warn "File not found: $file"
    return
  fi

  if [ "$DRY_RUN" = false ]; then
    # Update hash (handles both placeholder and real hashes)
    sed -i '' "s|value: .*${old_hash}.*|value: ${new_hash}|g" "$file"
    # Update URL version tag
    sed -i '' "s|/tags/${old_url_version}.tar.gz|/tags/${new_url_version}.tar.gz|g" "$file"
    # Update image version tag
    sed -i '' "s|\"${old_tag}\"|\"${new_tag}\"|g" "$file"
    ok "Updated: $file"
  else
    info "[dry-run] Would update: $file"
  fi
}

# Iron Bank manifests in aegis-platform/ironbank/
for manifest in \
  "${REPO_ROOT}/ironbank/platform-api/hardening_manifest.yaml" \
  "${REPO_ROOT}/ironbank/k8s-agent/hardening_manifest.yaml" \
  "${REPO_ROOT}/ironbank/proxy/hardening_manifest.yaml"; do

  if [ -f "$manifest" ]; then
    if [ "$DRY_RUN" = false ]; then
      # Update platform source hash
      sed -i '' "s|value: .*|value: ${PLATFORM_HASH}|" "$manifest"
      # Update URL to new version
      sed -i '' "s|/tags/v[0-9]*\.[0-9]*\.[0-9]*.tar.gz|/tags/${VERSION}.tar.gz|g" "$manifest"
      # Update version tag
      sed -i '' "s|  - \"[0-9]*\.[0-9]*\.[0-9]*\"|  - \"${VERSION_NUM}\"|" "$manifest"
      sed -i '' "s|org.opencontainers.image.version: \"[0-9]*\.[0-9]*\.[0-9]*\"|org.opencontainers.image.version: \"${VERSION_NUM}\"|" "$manifest"
      ok "Updated: $manifest"
    else
      info "[dry-run] Would update: $manifest"
    fi
  fi
done

# Workspace and vscode-reh-init manifests (no source tarball to update, just version)
for manifest in \
  "${REPO_ROOT}/ironbank/workspace/hardening_manifest.yaml" \
  "${REPO_ROOT}/ironbank/vscode-reh-init/hardening_manifest.yaml"; do

  if [ -f "$manifest" ] && [ "$DRY_RUN" = false ]; then
    sed -i '' "s|  - \"[0-9]*\.[0-9]*\.[0-9]*\"|  - \"${VERSION_NUM}\"|" "$manifest"
    sed -i '' "s|org.opencontainers.image.version: \"[0-9]*\.[0-9]*\.[0-9]*\"|org.opencontainers.image.version: \"${VERSION_NUM}\"|" "$manifest"
    ok "Updated: $manifest"
  fi
done

# UI manifest
UI_MANIFEST="${UI_ROOT}/ironbank/hardening_manifest.yaml"
if [ -f "$UI_MANIFEST" ]; then
  if [ "$DRY_RUN" = false ]; then
    sed -i '' "s|value: .*|value: ${UI_HASH}|" "$UI_MANIFEST"
    sed -i '' "s|/tags/v[0-9]*\.[0-9]*\.[0-9]*.tar.gz|/tags/${VERSION}.tar.gz|g" "$UI_MANIFEST"
    sed -i '' "s|  - \"[0-9]*\.[0-9]*\.[0-9]*\"|  - \"${VERSION_NUM}\"|" "$UI_MANIFEST"
    sed -i '' "s|org.opencontainers.image.version: \"[0-9]*\.[0-9]*\.[0-9]*\"|org.opencontainers.image.version: \"${VERSION_NUM}\"|" "$UI_MANIFEST"
    ok "Updated: $UI_MANIFEST"
  else
    info "[dry-run] Would update: $UI_MANIFEST"
  fi
fi

# Also update the existing in-place manifests
for manifest in \
  "${REPO_ROOT}/services/platform-api/hardening_manifest.yaml" \
  "${REPO_ROOT}/agents/k8s-agent/hardening_manifest.yaml" \
  "${REPO_ROOT}/services/proxy/hardening_manifest.yaml"; do

  if [ -f "$manifest" ] && [ "$DRY_RUN" = false ]; then
    sed -i '' "s|PLACEHOLDER_REPLACE_WITH_ACTUAL_HASH_AT_RELEASE|${PLATFORM_HASH}|g" "$manifest"
    sed -i '' "s|/tags/v[0-9]*\.[0-9]*\.[0-9]*.tar.gz|/tags/${VERSION}.tar.gz|g" "$manifest"
    ok "Updated in-place: $manifest"
  fi
done

UI_INPLACE="${UI_ROOT}/packages/backend/hardening_manifest.yaml"
if [ -f "$UI_INPLACE" ] && [ "$DRY_RUN" = false ]; then
  sed -i '' "s|PLACEHOLDER_REPLACE_WITH_ACTUAL_HASH_AT_RELEASE|${UI_HASH}|g" "$UI_INPLACE"
  sed -i '' "s|/tags/v[0-9]*\.[0-9]*\.[0-9]*.tar.gz|/tags/${VERSION}.tar.gz|g" "$UI_INPLACE"
  ok "Updated in-place: $UI_INPLACE"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
printf "\n${BOLD}═══════════════════════════════════════════${RESET}\n"
printf "${BOLD}  Iron Bank Release Summary: ${VERSION}${RESET}\n"
printf "${BOLD}═══════════════════════════════════════════${RESET}\n\n"

printf "  Platform tarball URL:\n"
printf "    ${PLATFORM_URL}\n"
printf "  Platform sha256:\n"
printf "    ${PLATFORM_HASH}\n\n"

printf "  UI tarball URL:\n"
printf "    ${UI_URL}\n"
printf "  UI sha256:\n"
printf "    ${UI_HASH}\n\n"

printf "  Updated manifests:\n"
printf "    ironbank/platform-api/hardening_manifest.yaml\n"
printf "    ironbank/k8s-agent/hardening_manifest.yaml\n"
printf "    ironbank/proxy/hardening_manifest.yaml\n"
printf "    ironbank/workspace/hardening_manifest.yaml\n"
printf "    ironbank/vscode-reh-init/hardening_manifest.yaml\n"
printf "    aegis-ui/ironbank/hardening_manifest.yaml\n\n"

ok "Release ${VERSION} prepared successfully."
printf "\n  Next steps:\n"
printf "    1. Review the updated manifests\n"
printf "    2. Push artifacts to Repo1 development branches\n"
printf "    3. Submit the Getting Started form if not done\n\n"
