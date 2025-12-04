#!/usr/bin/env bash

set -euo pipefail

# Lightweight Keycloak token helper for local/preview E2E tests.
# - Prefers env overrides (KEYCLOAK_* vars)
# - Falls back to local defaults (keycloak.localtest.me, realm=aegis, backstage client)
# - Supports password or client_credentials grants
# - Keeps existing Keycloak/Helm state untouched

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "✖ missing dependency: $1" >&2; exit 2; }
}

need curl
need jq

BASE_URL="${KEYCLOAK_BASE_URL:-https://keycloak.localtest.me}"
REALM="${KEYCLOAK_REALM:-aegis}"
TOKEN_URL="${KEYCLOAK_TOKEN_URL:-${BASE_URL%/}/realms/${REALM}/protocol/openid-connect/token}"
CLIENT_ID="${KEYCLOAK_CLIENT_ID:-backstage}"
CLIENT_SECRET="${KEYCLOAK_CLIENT_SECRET:-}"
USERNAME="${KEYCLOAK_USERNAME:-}"
PASSWORD="${KEYCLOAK_PASSWORD:-}"
GRANT_TYPE="${KEYCLOAK_GRANT_TYPE:-}"
SCOPE="${KEYCLOAK_SCOPE:-openid profile email offline_access}"
AUDIENCE="${KEYCLOAK_AUDIENCE:-}"
CA_CERT="${KEYCLOAK_CA_CERT:-${HOME}/keycloak.localtest.me.crt}"
INSECURE="${KEYCLOAK_INSECURE:-0}"

# If grant type not set, infer from presence of username/password.
if [[ -z "${GRANT_TYPE}" ]]; then
  if [[ -n "${USERNAME}" && -n "${PASSWORD}" ]]; then
    GRANT_TYPE="password"
  else
    GRANT_TYPE="client_credentials"
  fi
fi

# If client secret not provided, try to read the local secret created by the chart.
detect_client_secret() {
  local secret ns key data
  [[ -n "${CLIENT_SECRET}" ]] && return
  command -v kubectl >/dev/null 2>&1 || return
  secret="${KEYCLOAK_CLIENT_SECRET_NAME:-keycloak-backstage-client-secret}"
  ns="${KEYCLOAK_NAMESPACE:-keycloak}"
  for key in clientSecret CLIENT_SECRET secret; do
    if data=$(kubectl get secret "${secret}" -n "${ns}" -o "jsonpath={.data.${key}}" 2>/dev/null); then
      if [[ -n "${data}" ]]; then
        CLIENT_SECRET="$(printf '%s' "${data}" | base64 --decode)"
        return
      fi
    fi
  done
  # fallback to known local default if still empty
  CLIENT_SECRET="${CLIENT_SECRET:-local-backstage-client-secret}"
}

if [[ "${GRANT_TYPE}" == "client_credentials" ]]; then
  detect_client_secret
  if [[ -z "${CLIENT_SECRET}" ]]; then
    echo "✖ KEYCLOAK_CLIENT_SECRET is required for client_credentials grant" >&2
    exit 1
  fi
fi

CURL_ARGS=(-sS --fail --request POST "${TOKEN_URL}")
if [[ -n "${CA_CERT}" && -f "${CA_CERT}" ]]; then
  CURL_ARGS+=(--cacert "${CA_CERT}")
fi
if [[ "${INSECURE}" == "1" ]]; then
  CURL_ARGS+=(--insecure)
fi

declare -a FORM_DATA
case "${GRANT_TYPE}" in
  password)
    if [[ -z "${USERNAME}" || -z "${PASSWORD}" ]]; then
      echo "✖ KEYCLOAK_USERNAME and KEYCLOAK_PASSWORD are required for password grant" >&2
      exit 1
    fi
    FORM_DATA+=(
      "grant_type=password"
      "client_id=${CLIENT_ID}"
      "username=${USERNAME}"
      "password=${PASSWORD}"
    )
    [[ -n "${CLIENT_SECRET}" ]] && FORM_DATA+=("client_secret=${CLIENT_SECRET}")
    ;;
  client_credentials)
    FORM_DATA+=(
      "grant_type=client_credentials"
      "client_id=${CLIENT_ID}"
      "client_secret=${CLIENT_SECRET}"
    )
    ;;
  *)
    echo "✖ Invalid grant type: ${GRANT_TYPE}" >&2
    exit 1
    ;;
esac

[[ -n "${SCOPE}" ]] && FORM_DATA+=("scope=${SCOPE}")
[[ -n "${AUDIENCE}" ]] && FORM_DATA+=("audience=${AUDIENCE}")

for entry in "${FORM_DATA[@]}"; do
  CURL_ARGS+=(--data "${entry}")
done

resp="$(curl "${CURL_ARGS[@]}")"
token="$(printf '%s' "${resp}" | jq -r '.access_token // empty')"
if [[ -z "${token}" ]]; then
  echo "✖ Failed to obtain access token from Keycloak" >&2
  printf 'Response: %s\n' "${resp}" >&2
  exit 1
fi

printf '%s\n' "${token}"
