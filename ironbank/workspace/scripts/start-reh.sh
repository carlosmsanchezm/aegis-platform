#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# VS Code Remote Extension Host (REH) Launcher — Iron Bank Version
# =============================================================================
# Starts the pre-installed VS Code server binary from the shared volume
# (injected by the vscode-reh-init container). No runtime downloads.
# =============================================================================

log() {
  printf '[start-reh] %s\n' "$*" >&2
}

trap 'log "received termination signal"; exit 0' TERM INT

QUALITY="${VSCODE_QUALITY:-stable}"

# Setup workspace directory
WORKSPACE_ROOT="${WORKSPACE_ROOT:-/home/aegis/work}"
mkdir -p "${WORKSPACE_ROOT}"
if [[ ! -f "${WORKSPACE_ROOT}/README.md" ]]; then
  cat <<'EOF' > "${WORKSPACE_ROOT}/README.md"
# Aegis Workspace

This folder is provisioned by the Aegis workspace image.
EOF
fi

REH_DIR="${VSCODE_SERVER_BASE:-/reh}"

# Create connection token file
printf 'hello' > "${REH_DIR}/token"

# Determine server binary name
SERVER_BIN="code-server"
if [[ "${QUALITY}" == "insider" || "${QUALITY}" == "insiders" ]]; then
  SERVER_BIN="code-server-insiders"
fi

PORT="${VSCODE_SERVER_PORT:-11111}"

log "Starting VS Code server on port ${PORT}"
log "Workspace ready at ${WORKSPACE_ROOT}"
exec "${REH_DIR}/bin/current/bin/${SERVER_BIN}" \
  --host 0.0.0.0 \
  --port "${PORT}" \
  --telemetry-level off \
  --connection-token-file "${REH_DIR}/token" \
  --accept-server-license-terms \
  --disable-telemetry \
  "$@"
