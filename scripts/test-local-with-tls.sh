#!/usr/bin/env bash
# Runs the full local test suite twice: once against the standard local
# deployment (no TLS) and once against the TLS-enabled deployment. This wraps
# existing make targets / scripts so the swarm can trigger the whole flow with
# a single command (AEGIS_TEST_CMD).

set -Eeuo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

log() {
  printf '\n[%s] %s\n' "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" "$1"
}

run_test_all_local() {
  local phase="$1"
  shift
  log "Running ./scripts/test-all-local.sh (${phase})"
  "$@"
}

HTTP_ARGS=("$@")
TLS_ARGS=("$@")
if [[ -n "${HTTP_TEST_ARGS:-}" ]]; then
  read -r -a HTTP_ARGS <<<"${HTTP_TEST_ARGS}"
fi
if [[ -n "${TLS_TEST_ARGS:-}" ]]; then
  read -r -a TLS_ARGS <<<"${TLS_TEST_ARGS}"
fi

HTTP_RUN_E2E_PLATFORM="${HTTP_RUN_E2E_PLATFORM:-0}"
HTTP_RUN_E2E_OPERATOR="${HTTP_RUN_E2E_OPERATOR:-0}"
TLS_RUN_E2E_PLATFORM="${TLS_RUN_E2E_PLATFORM:-1}"
TLS_RUN_E2E_OPERATOR="${TLS_RUN_E2E_OPERATOR:-1}"
TLS_SKIP_VERIFY="${TLS_SKIP_VERIFY:-1}"
TLS_CA_FILE="${TLS_CA:-$HOME/aegis-platform-api-ca.crt}"

log "Deploying local stack without TLS"
make deploy-local

if (( ${#HTTP_ARGS[@]} )); then
  run_test_all_local "http" \
    env \
      RUN_E2E_PLATFORM="$HTTP_RUN_E2E_PLATFORM" \
      RUN_E2E_OPERATOR="$HTTP_RUN_E2E_OPERATOR" \
      ./scripts/test-all-local.sh "${HTTP_ARGS[@]}"
else
  run_test_all_local "http" \
    env \
      RUN_E2E_PLATFORM="$HTTP_RUN_E2E_PLATFORM" \
      RUN_E2E_OPERATOR="$HTTP_RUN_E2E_OPERATOR" \
      ./scripts/test-all-local.sh
fi

log "Deploying local stack with TLS"
make deploy-local-tls

if [[ ! -s "$TLS_CA_FILE" ]]; then
  printf 'TLS CA bundle not found at %s\n' "$TLS_CA_FILE" >&2
  exit 1
fi

if (( ${#TLS_ARGS[@]} )); then
  run_test_all_local "tls" \
    env \
      GRPC_TLS=1 \
      GRPC_TLS_SKIP_VERIFY="$TLS_SKIP_VERIFY" \
      GRPC_CA="$TLS_CA_FILE" \
      AEGIS_GRPC_ADDR="platform-api-grpc.localtest.me:443" \
      AEGIS_PROXY_HOSTNAME="proxy.localtest.me" \
      RUN_E2E_PLATFORM="$TLS_RUN_E2E_PLATFORM" \
      RUN_E2E_OPERATOR="$TLS_RUN_E2E_OPERATOR" \
      ./scripts/test-all-local.sh "${TLS_ARGS[@]}"
else
  run_test_all_local "tls" \
    env \
      GRPC_TLS=1 \
      GRPC_TLS_SKIP_VERIFY="$TLS_SKIP_VERIFY" \
      GRPC_CA="$TLS_CA_FILE" \
      AEGIS_GRPC_ADDR="platform-api-grpc.localtest.me:443" \
      AEGIS_PROXY_HOSTNAME="proxy.localtest.me" \
      RUN_E2E_PLATFORM="$TLS_RUN_E2E_PLATFORM" \
      RUN_E2E_OPERATOR="$TLS_RUN_E2E_OPERATOR" \
      ./scripts/test-all-local.sh
fi

log "Local HTTP + TLS test run complete"
