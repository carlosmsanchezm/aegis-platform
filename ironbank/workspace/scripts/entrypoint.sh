#!/usr/bin/env bash
set -euo pipefail

log() {
  printf '[entrypoint] %s\n' "$*" >&2
}

# Ensure workspace directory exists
WORKSPACE_ROOT="${WORKSPACE_ROOT:-/home/aegis/work}"
mkdir -p "${WORKSPACE_ROOT}"

# Verify VS Code REH binary exists (injected by init container via shared volume)
REH_DIR="${VSCODE_SERVER_BASE:-/reh}"
if [[ ! -d "${REH_DIR}/bin/current" ]]; then
  log "ERROR: VS Code REH not found at ${REH_DIR}/bin/current"
  log "The vscode-reh-init container may not have run. Check init container logs."
  exit 1
fi

if [[ ! -x "${REH_DIR}/bin/current/bin/code-server" ]]; then
  log "ERROR: code-server binary not executable at ${REH_DIR}/bin/current/bin/code-server"
  log "Init container may have failed to extract the VS Code server tarball."
  exit 1
fi

log "VS Code REH found at ${REH_DIR}/bin/current"
log "Starting workspace..."

exec /usr/local/bin/start-reh.sh "$@"
