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

keycloak_admin_token() {
  if [[ -n "${_KEYCLOAK_ADMIN_TOKEN:-}" ]]; then
    printf '%s' "$_KEYCLOAK_ADMIN_TOKEN"
    return
  fi
  local token attempt
  for attempt in {1..12}; do
    token=$(
      curl -sS --fail --cacert "$HOME/keycloak.localtest.me.crt" \
        -X POST "https://keycloak.localtest.me/realms/master/protocol/openid-connect/token" \
        -d "grant_type=password" \
        -d "client_id=admin-cli" \
        -d "username=admin" \
        -d "password=SuperSecureAdmin123!" \
        | jq -r '.access_token' 2>/dev/null || true
    )
    if [[ -n "$token" && "$token" != "null" ]]; then
      _KEYCLOAK_ADMIN_TOKEN="$token"
      printf '%s' "$token"
      return
    fi
    sleep 5
  done
}

ensure_backstage_direct_access() {
  local admin_token client_uuid
  admin_token=$(keycloak_admin_token)
  if [[ -z "$admin_token" || "$admin_token" == "null" ]]; then
    log "Skipped Backstage direct access update (admin token unavailable)"
    return
  fi
  client_uuid=$(
    curl -sS --fail --cacert "$HOME/keycloak.localtest.me.crt" \
      -H "Authorization: Bearer $admin_token" \
      "https://keycloak.localtest.me/admin/realms/aegis/clients?clientId=backstage" \
      | jq -r '.[0].id' 2>/dev/null || true
  )
  if [[ -z "$client_uuid" || "$client_uuid" == "null" ]]; then
    log "Skipped Backstage direct access update (client id not found)"
    return
  fi
  curl -sS --fail --cacert "$HOME/keycloak.localtest.me.crt" \
    -H "Authorization: Bearer $admin_token" \
    -H "Content-Type: application/json" \
    -X PUT "https://keycloak.localtest.me/admin/realms/aegis/clients/${client_uuid}" \
    -d '{"directAccessGrantsEnabled":true}' >/dev/null 2>&1 || true
}

ensure_automation_user() {
  local admin_token user_json user_id role_json
  admin_token=$(keycloak_admin_token)
  if [[ -z "$admin_token" || "$admin_token" == "null" ]]; then
    log "Skipped automation user sync (admin token unavailable)"
    return
  fi
  local encoded_username="cloud%40test.com"
  local desired_username="cloud@test.com"
  user_json=$(
    curl -sS --fail --cacert "$HOME/keycloak.localtest.me.crt" \
      -H "Authorization: Bearer $admin_token" \
      "https://keycloak.localtest.me/admin/realms/aegis/users?search=${encoded_username}&exact=true" \
      2>/dev/null || true
  )
  user_id=$(echo "$user_json" | jq -r --arg username "$desired_username" 'map(select(.username==$username)) | .[0].id // empty' 2>/dev/null || true)
  if [[ -z "$user_id" || "$user_id" == "null" ]]; then
    curl -sS --fail --cacert "$HOME/keycloak.localtest.me.crt" \
      -H "Authorization: Bearer $admin_token" \
      -H "Content-Type: application/json" \
      -X POST "https://keycloak.localtest.me/admin/realms/aegis/users" \
      -d '{
        "username": "cloud@test.com",
        "email": "cloud@test.com",
        "firstName": "cloud",
        "lastName": "user",
        "enabled": true,
        "emailVerified": true,
        "credentials": [
          {"type":"password","value":"password","temporary": false}
        ]
      }' >/dev/null 2>&1 || true
    user_json=$(
      curl -sS --fail --cacert "$HOME/keycloak.localtest.me.crt" \
        -H "Authorization: Bearer $admin_token" \
        "https://keycloak.localtest.me/admin/realms/aegis/users?search=${encoded_username}&exact=true" \
        2>/dev/null || true
    )
    user_id=$(echo "$user_json" | jq -r --arg username "$desired_username" 'map(select(.username==$username)) | .[0].id // empty' 2>/dev/null || true)
  else
    curl -sS --fail --cacert "$HOME/keycloak.localtest.me.crt" \
      -H "Authorization: Bearer $admin_token" \
      -H "Content-Type: application/json" \
      -X PUT "https://keycloak.localtest.me/admin/realms/aegis/users/${user_id}/reset-password" \
      -d '{"type":"password","value":"password","temporary":false}' >/dev/null 2>&1 || true
  fi
  if [[ -z "$user_id" || "$user_id" == "null" ]]; then
    log "Failed to reconcile automation user in Keycloak"
    return
  fi
  role_json=$(
    curl -sS --fail --cacert "$HOME/keycloak.localtest.me.crt" \
      -H "Authorization: Bearer $admin_token" \
      "https://keycloak.localtest.me/admin/realms/aegis/roles/workspace-admin" \
      2>/dev/null || true
  )
  if [[ -n "$role_json" ]]; then
    curl -sS --fail --cacert "$HOME/keycloak.localtest.me.crt" \
      -H "Authorization: Bearer $admin_token" \
      -H "Content-Type: application/json" \
      -X POST "https://keycloak.localtest.me/admin/realms/aegis/users/${user_id}/role-mappings/realm" \
      -d "[${role_json}]" >/dev/null 2>&1 || true
  fi
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
# Reset Keycloak realm import so Helm applies updated realm configuration
# This is important because Keycloak only imports realms on initial creation, not on updates
log "Forcing Keycloak realm reimport to pick up configuration changes"
kubectl delete keycloakrealmimports.k8s.keycloak.org/aegis-services-keycloak-realm -n keycloak --ignore-not-found >/dev/null 2>&1 || true
kubectl delete secret/aegis-services-keycloak-aegis-realm -n keycloak --ignore-not-found >/dev/null 2>&1 || true
kubectl delete job.batch/aegis-services-keycloak-realm -n keycloak --ignore-not-found >/dev/null 2>&1 || true
make deploy-local-tls
# Wait for realm import to complete
log "Waiting for Keycloak realm import to complete"
sleep 10
kubectl wait --for=condition=complete job -l app.kubernetes.io/component=realm -n keycloak --timeout=120s >/dev/null 2>&1 || true

ensure_automation_user
ensure_backstage_direct_access

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
      KEYCLOAK_BASE_URL="https://keycloak.localtest.me" \
      KEYCLOAK_REALM="aegis" \
      KEYCLOAK_CLIENT_ID="backstage" \
      KEYCLOAK_CLIENT_SECRET="local-backstage-client-secret" \
      KEYCLOAK_INSECURE=1 \
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
      KEYCLOAK_BASE_URL="https://keycloak.localtest.me" \
      KEYCLOAK_REALM="aegis" \
      KEYCLOAK_CLIENT_ID="backstage" \
      KEYCLOAK_CLIENT_SECRET="local-backstage-client-secret" \
      KEYCLOAK_INSECURE=1 \
      ./scripts/test-all-local.sh
fi

log "Local HTTP + TLS test run complete"
