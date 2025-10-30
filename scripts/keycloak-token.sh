#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: keycloak-token.sh

Fetches an access token from the configured Keycloak realm and prints it to stdout.

Environment variables:
  KEYCLOAK_TOKEN_URL        Full token endpoint URL. If unset, derived from
                            KEYCLOAK_BASE_URL and KEYCLOAK_REALM.
  KEYCLOAK_BASE_URL         Base URL such as https://keycloak.localtest.me
  KEYCLOAK_REALM            Realm name (default: aegis) when deriving URL.
  KEYCLOAK_CLIENT_ID        OAuth client identifier (required).
  KEYCLOAK_CLIENT_SECRET    Client secret for confidential clients.
  KEYCLOAK_USERNAME         Username for Resource Owner Password flow.
  KEYCLOAK_PASSWORD         Password for Resource Owner Password flow.
  KEYCLOAK_SCOPE            Optional space-delimited scopes to request.
  KEYCLOAK_GRANT_TYPE       Override grant type (password or client_credentials).
  KEYCLOAK_CA_CERT          Path to CA bundle when Keycloak uses a custom certificate.
  KEYCLOAK_SKIP_TLS_VERIFY  Set to 1 to disable TLS verification (not recommended).

The script attempts the Resource Owner Password grant when username/password
values are supplied. Otherwise it falls back to the Client Credentials grant
when a client secret is available.
EOF
}

if [[ "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "✖ Required dependency missing: $1" >&2
    exit 1
  fi
}

need curl
need jq

KEYCLOAK_REALM=${KEYCLOAK_REALM:-aegis}
TOKEN_URL=${KEYCLOAK_TOKEN_URL:-}
if [[ -z "${TOKEN_URL}" ]]; then
  if [[ -n "${KEYCLOAK_BASE_URL:-}" ]]; then
    base=${KEYCLOAK_BASE_URL%/}
    TOKEN_URL="${base}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token"
  else
    echo "✖ KEYCLOAK_TOKEN_URL or KEYCLOAK_BASE_URL must be set" >&2
    exit 1
  fi
fi

CLIENT_ID=${KEYCLOAK_CLIENT_ID:-}
if [[ -z "${CLIENT_ID}" ]]; then
  echo "✖ KEYCLOAK_CLIENT_ID must be provided" >&2
  exit 1
fi

CLIENT_SECRET=${KEYCLOAK_CLIENT_SECRET:-}
USERNAME=${KEYCLOAK_USERNAME:-}
PASSWORD=${KEYCLOAK_PASSWORD:-}
SCOPE=${KEYCLOAK_SCOPE:-}
GRANT_TYPE=${KEYCLOAK_GRANT_TYPE:-}

if [[ -z "${GRANT_TYPE}" ]]; then
  if [[ -n "${USERNAME}" || -n "${PASSWORD}" ]]; then
    GRANT_TYPE=password
  else
    GRANT_TYPE=client_credentials
  fi
fi

declare -a CURL_ARGS=("-sS" "--fail" "--request" "POST" "${TOKEN_URL}")
if [[ -n "${KEYCLOAK_CA_CERT:-}" && -f "${KEYCLOAK_CA_CERT}" ]]; then
  CURL_ARGS+=("--cacert" "${KEYCLOAK_CA_CERT}")
elif [[ -n "${KEYCLOAK_CA_BUNDLE:-}" && -f "${KEYCLOAK_CA_BUNDLE}" ]]; then
  CURL_ARGS+=("--cacert" "${KEYCLOAK_CA_BUNDLE}")
fi
if [[ "${KEYCLOAK_SKIP_TLS_VERIFY:-0}" == "1" ]]; then
  CURL_ARGS+=("--insecure")
fi
declare -a FORM_DATA

case "${GRANT_TYPE}" in
  password)
    if [[ -z "${USERNAME}" || -z "${PASSWORD}" ]]; then
      echo "✖ KEYCLOAK_USERNAME and KEYCLOAK_PASSWORD are required for password grant" >&2
      exit 1
    fi
    FORM_DATA+=("grant_type=password" "username=${USERNAME}" "password=${PASSWORD}" "client_id=${CLIENT_ID}")
    if [[ -n "${CLIENT_SECRET}" ]]; then
      FORM_DATA+=("client_secret=${CLIENT_SECRET}")
    fi
    ;;
  client_credentials)
    if [[ -z "${CLIENT_SECRET}" ]]; then
      echo "✖ KEYCLOAK_CLIENT_SECRET is required for client_credentials grant" >&2
      exit 1
    fi
    FORM_DATA+=("grant_type=client_credentials" "client_id=${CLIENT_ID}" "client_secret=${CLIENT_SECRET}")
    ;;
  *)
    echo "✖ Unsupported KEYCLOAK_GRANT_TYPE: ${GRANT_TYPE}" >&2
    exit 1
    ;;
 esac

if [[ -n "${SCOPE}" ]]; then
  FORM_DATA+=("scope=${SCOPE}")
fi

for entry in "${FORM_DATA[@]}"; do
  CURL_ARGS+=("--data" "${entry}")
done

TMP_BODY=$(mktemp)
trap 'rm -f "${TMP_BODY}"' EXIT

HTTP_STATUS=$(curl "${CURL_ARGS[@]}" -w '%{http_code}' -o "${TMP_BODY}" || true)
if [[ ! ${HTTP_STATUS} =~ ^[0-9]{3}$ ]]; then
  echo "✖ Failed to reach Keycloak token endpoint" >&2
  cat "${TMP_BODY}" >&2 || true
  exit 1
fi

if [[ "${HTTP_STATUS}" -ge 400 ]]; then
  echo "✖ Keycloak token request failed (HTTP ${HTTP_STATUS})" >&2
  cat "${TMP_BODY}" >&2 || true
  exit 1
fi

token=$(jq -r '.access_token // empty' "${TMP_BODY}")
if [[ -z "${token}" ]]; then
  echo "✖ Response did not contain an access_token" >&2
  cat "${TMP_BODY}" >&2
  exit 1
fi

echo "${token}"
