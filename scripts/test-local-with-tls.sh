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

wait_for_helm_release() {
  local release="$1"
  local namespace="$2"
  local max_wait=${3:-120}
  local sleep_interval=${4:-5}

  if ! command -v helm >/dev/null 2>&1; then
    log "helm not found; skipping wait for release ${release}"
    return 0
  fi

  for ((i = 0; i < max_wait; i++)); do
    if ! status_output=$(helm status "$release" -n "$namespace" 2>&1); then
      # Release does not exist yet; nothing is currently in progress.
      return 0
    fi

    if ! grep -qi "pending" <<<"$status_output"; then
      return 0
    fi

    sleep "${sleep_interval}"
  done

  printf 'Timed out waiting for Helm release %s in %s to finish current operation\n' "$release" "$namespace" >&2
  return 1
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
TLS_AGENT_IMAGE="${K8S_AGENT_IMAGE:-carlosmsanchez/aegis-k8s-agent:dev}"

log "Deploying local stack with TLS"
wait_for_helm_release "aegis-services" "aegis-system" || exit 1
wait_for_helm_release "aegis-spoke" "aegis-system" || exit 1
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
      K8S_AGENT_E2E_IMAGE="$TLS_AGENT_IMAGE" \
      K8S_AGENT_E2E_SKIP_BUILD=1 \
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
      K8S_AGENT_E2E_IMAGE="$TLS_AGENT_IMAGE" \
      K8S_AGENT_E2E_SKIP_BUILD=1 \
      RUN_E2E_PLATFORM="$TLS_RUN_E2E_PLATFORM" \
      RUN_E2E_OPERATOR="$TLS_RUN_E2E_OPERATOR" \
      ./scripts/test-all-local.sh
fi

log "Local HTTP + TLS test run complete"
