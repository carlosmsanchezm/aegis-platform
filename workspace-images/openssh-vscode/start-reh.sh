#!/usr/bin/env bash
set -euo pipefail

log() {
  printf '[start-reh] %s\n' "$*" >&2
}

trap 'log "received termination signal"; exit 0' TERM INT

QUALITY="${VSCODE_QUALITY:-stable}"
COMMIT="${VSCODE_COMMIT:-${VSCODE_SERVER_COMMIT:-}}"

if [[ $# -gt 0 && "${1:-}" =~ ^[0-9a-f]{40}$ ]]; then
  [[ -zMMIT" ]] && COMMIT="$1"
  shift
fi

if [[ -z "$COMMIT" && -n "${VSCODE_SERVER_URL:-}" ]]; then
  if [[ "${VSCODE_SERVER_URL}" =~ commit:?([0-9a-f]{40}) ]]; then
    COMMIT="${BASH_REMATCH[1]}"
  fi
fi

if [[ -z "$COMMIT" ]]; then
  log "Set VSCODE_COMMIT (40 hex chars) to match the local VS Code client."
  exit 1
fi

REH_DIR="${VSCODE_SERVER_BASE:-/reh}"
mkdir -p "${REH_DIR}/bin/current" "${REH_DIR}/workspace"

WORKSPACE_ROOT="${WORKSPACE_ROOT:-/home/project}"
if [[ -e "${WORKSPACE_ROOT}" && ! -d "${WORKSPACE_ROOT}" ]]; then
  rm -f "${WORKSPACE_ROOT}"
fi
mkdir -p "${WORKSPACE_ROOT}"
if [[ ! -f "${WORKSPACE_ROOT}/README.md" ]]; then
  cat <<'EOF' > "${WORKSPACE_ROOT}/README.md"
# Aegis Sample Workspace

This folder is provisioned by the Aegis workspace image.

Feel free to add files here to validate remote editing via VS Code.
EOF
fi

ARCH="$(uname -m)"
case "${ARCH}" in
  aarch64|arm64) ARCH_SUFFIX="arm64" ;;
  *) ARCH_SUFFIX="x64" ;;
esac

BASE_URL="https://update.code.visualstudio.com"
TMP_TAR="$(mktemp /tmp/vscode-server-XXXXXX.tar.gz)"

DOWNLOAD_URL="${VSCODE_SERVER_URL:-}"
if [[ -z "${DOWNLOAD_URL}" ]]; then
  DOWNLOAD_URL="https://vscode.download.prss.microsoft.com/dbazure/download/${QUALITY}/${COMMIT}/vscode-server-linux-${ARCH_SUFFIX}.tar.gz"
fi
API_URL="${BASE_URL}/commit/${COMMIT}/server-linux-${ARCH_SUFFIX}/${QUALITY}"
LATEST_URL="${BASE_URL}/latest/server-linux-${ARCH_SUFFIX}/${QUALITY}"

MARK_FILE="${REH_DIR}/bin/current/.commit"
CURRENT_MARK=""
if [[ -f "${MARK_FILE}" ]]; then
  CURRENT_MARK="$(<"${MARK_FILE}")"
fi

if [[ "${CURRENT_MARK}" != "${COMMIT}" || ! -x "${REH_DIR}/bin/current/bin/code-server" ]]; then
  log "Downloading VS Code server commit ${COMMIT} (${QUALITY}, ${ARCH_SUFFIX})"
  if ! curl -fsSL "${DOWNLOAD_URL}" -o "${TMP_TAR}"; then
    if ! curl -fsSL "${API_URL}" -o "${TMP_TAR}"; then
      log "Falling back to latest ${QUALITY} server for ${ARCH_SUFFIX}"
      curl -fsSL "${LATEST_URL}" -o "${TMP_TAR}"
    fi
  fi

  rm -rf "${REH_DIR}/bin/current"
  mkdir -p "${REH_DIR}/bin/current"
  tar -xzf "${TMP_TAR}" -C "${REH_DIR}/bin/current" --strip-components=1
  printf '%s' "${COMMIT}" > "${MARK_FILE}"
fi

rm -f "${TMP_TAR}"

printf 'hello' > "${REH_DIR}/token"

SERVER_BIN="code-server"
if [[ "${QUALITY}" == "insider" || "${QUALITY}" == "insiders" ]]; then
  SERVER_BIN="code-server-insiders"
fi

PORT="${VSCODE_SERVER_PORT:-11111}"

exec "${REH_DIR}/bin/current/bin/${SERVER_BIN}" \
  --host 0.0.0.0 \
  --port "${PORT}" \
  --telemetry-level off \
  --connection-token-file "${REH_DIR}/token" \
  --accept-server-license-terms \
  --disable-telemetry \
  "$@"
