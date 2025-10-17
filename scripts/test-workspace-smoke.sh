#!/usr/bin/env bash
set -euo pipefail

IMAGE="${1:-${WORKSPACE_IMAGE:-aegis-workspace:latest}}"
QUALITY="${VSCODE_QUALITY:-insider}"
COMMIT="${VSCODE_COMMIT:-}"
TIMEOUT_SECONDS="${SMOKE_TIMEOUT:-120}"

usage() {
  cat <<EOF
Usage: $(basename "$0") [image]

Environment variables:
  WORKSPACE_IMAGE   Override the image name (default: aegis-workspace:latest)
  VSCODE_COMMIT     Required VS Code commit hash (40 hex characters)
  VSCODE_QUALITY    Channel to fetch (default: insider)
  SMOKE_TIMEOUT     Seconds to wait for the server to start (default: 120)
EOF
}

if [[ "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ -z "${COMMIT}" ]]; then
  for bin in code-insiders code; do
    if command -v "${bin}" >/dev/null 2>&1; then
      COMMIT="$(${bin} --version | sed -n '2p' | awk '{print $NF}')"
      break
    fi
  done
fi

if [[ -z "${COMMIT}" ]]; then
  echo "✖ VSCODE_COMMIT is required (40-character commit hash)." >&2
  exit 1
fi

if [[ ! "${COMMIT}" =~ ^[0-9a-f]{40}$ ]]; then
  echo "✖ VSCODE_COMMIT must be a 40-character hexadecimal hash." >&2
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "✖ docker is required." >&2
  exit 1
fi

CONTAINER="workspace-smoke-$$"
cleanup() {
  docker rm -f "${CONTAINER}" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

echo "→ Starting ${IMAGE} with commit ${COMMIT}"
docker run --rm -d \
  --name "${CONTAINER}" \
  -e VSCODE_COMMIT="${COMMIT}" \
  -e VSCODE_QUALITY="${QUALITY}" \
  "${IMAGE}" >/dev/null

echo "→ Waiting for VS Code server to advertise readiness..."
if ! timeout "${TIMEOUT_SECONDS}" docker logs -f "${CONTAINER}" 2>&1 | grep -m1 "Extension host agent listening"; then
  echo "✖ Server did not become ready within ${TIMEOUT_SECONDS}s." >&2
  exit 1
fi

echo "✔ VS Code server started successfully."
