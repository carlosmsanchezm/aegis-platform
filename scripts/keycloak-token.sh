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

cleanup_files=()
cleanup() {
  local status=$?
  if [[ ${#cleanup_files[@]} -gt 0 ]]; then
    rm -f "${cleanup_files[@]}" 2>/dev/null || true
  fi
  exit $status
}

trap cleanup EXIT

KEYCLOAK_REALM=${KEYCLOAK_REALM:-aegis}
TOKEN_URL=${KEYCLOAK_TOKEN_URL:-}
CLIENT_ID=${KEYCLOAK_CLIENT_ID:-}
CLIENT_SECRET=${KEYCLOAK_CLIENT_SECRET:-}
USERNAME=${KEYCLOAK_USERNAME:-}
PASSWORD=${KEYCLOAK_PASSWORD:-}
SCOPE=${KEYCLOAK_SCOPE:-}
GRANT_TYPE=${KEYCLOAK_GRANT_TYPE:-}

ensure_automation_user() {
  local ns="$1"
  local auto_username="${USERNAME:-automation@test.com}"
  local auto_password="${PASSWORD:-Automation123!}"
  local admin_secret="${KEYCLOAK_ADMIN_SECRET_NAME:-keycloak-admin-secret}"
  local admin_user_key="${KEYCLOAK_ADMIN_USERNAME_KEY:-username}"
  local admin_pass_key="${KEYCLOAK_ADMIN_PASSWORD_KEY:-password}"

  if [[ -n "${USERNAME:-}" && -n "${PASSWORD:-}" ]]; then
    return
  fi

  if [[ -z "${KEYCLOAK_BASE_URL:-}" ]] || ! command -v kubectl >/dev/null 2>&1; then
    USERNAME=${USERNAME:-$auto_username}
    PASSWORD=${PASSWORD:-$auto_password}
    return
  fi

  local admin_username admin_password
  admin_username=$(kubectl get secret "${admin_secret}" -n "${ns}" -o "jsonpath={.data.${admin_user_key}}" 2>/dev/null | base64 --decode 2>/dev/null || true)
  admin_password=$(kubectl get secret "${admin_secret}" -n "${ns}" -o "jsonpath={.data.${admin_pass_key}}" 2>/dev/null | base64 --decode 2>/dev/null || true)
  if [[ -z "${admin_username}" || -z "${admin_password}" ]]; then
    USERNAME=${USERNAME:-$auto_username}
    PASSWORD=${PASSWORD:-$auto_password}
    return
  fi

  local admin_token
  admin_token=$(curl -sS --fail ${KEYCLOAK_CA_CERT:+--cacert "${KEYCLOAK_CA_CERT}"} \
    -X POST "${KEYCLOAK_BASE_URL}/realms/master/protocol/openid-connect/token" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" \
    -d "username=${admin_username}" \
    -d "password=${admin_password}" \
    | jq -r '.access_token' 2>/dev/null || true)

  if [[ -z "${admin_token}" || "${admin_token}" == "null" ]]; then
    USERNAME=${USERNAME:-$auto_username}
    PASSWORD=${PASSWORD:-$auto_password}
    return
  fi

  local encoded_username
  encoded_username=$(python3 -c 'import sys, urllib.parse; print(urllib.parse.quote(sys.argv[1], safe=""))' "${auto_username}" 2>/dev/null || true)
  if [[ -z "${encoded_username}" ]]; then
    encoded_username="${auto_username}"
  fi

  local user_json user_id
  user_json=$(curl -sS --fail ${KEYCLOAK_CA_CERT:+--cacert "${KEYCLOAK_CA_CERT}"} \
    -H "Authorization: Bearer ${admin_token}" \
    "${KEYCLOAK_BASE_URL}/admin/realms/${KEYCLOAK_REALM}/users?search=${encoded_username}&exact=true" \
    2>/dev/null || true)
  user_id=$(echo "${user_json}" | jq -r --arg username "${auto_username}" 'map(select(.username==$username)) | .[0].id // empty' 2>/dev/null || true)

  if [[ -z "${user_id}" || "${user_id}" == "null" ]]; then
    curl -sS --fail ${KEYCLOAK_CA_CERT:+--cacert "${KEYCLOAK_CA_CERT}"} \
      -H "Authorization: Bearer ${admin_token}" \
      -H "Content-Type: application/json" \
      -X POST "${KEYCLOAK_BASE_URL}/admin/realms/${KEYCLOAK_REALM}/users" \
      -d "{\"username\":\"${auto_username}\",\"email\":\"${auto_username}\",\"firstName\":\"automation\",\"lastName\":\"user\",\"enabled\":true,\"emailVerified\":true,\"credentials\":[{\"type\":\"password\",\"value\":\"${auto_password}\",\"temporary\":false}]}" >/dev/null 2>&1 || true
    user_json=$(curl -sS --fail ${KEYCLOAK_CA_CERT:+--cacert "${KEYCLOAK_CA_CERT}"} \
      -H "Authorization: Bearer ${admin_token}" \
      "${KEYCLOAK_BASE_URL}/admin/realms/${KEYCLOAK_REALM}/users?search=${encoded_username}&exact=true" \
      2>/dev/null || true)
    user_id=$(echo "${user_json}" | jq -r --arg username "${auto_username}" 'map(select(.username==$username)) | .[0].id // empty' 2>/dev/null || true)
  else
    curl -sS --fail ${KEYCLOAK_CA_CERT:+--cacert "${KEYCLOAK_CA_CERT}"} \
      -H "Authorization: Bearer ${admin_token}" \
      -H "Content-Type: application/json" \
      -X PUT "${KEYCLOAK_BASE_URL}/admin/realms/${KEYCLOAK_REALM}/users/${user_id}/reset-password" \
      -d "{\"type\":\"password\",\"value\":\"${auto_password}\",\"temporary\":false}" >/dev/null 2>&1 || true
  fi

  if [[ -n "${user_id}" && "${user_id}" != "null" ]]; then
    local role_json
    role_json=$(curl -sS --fail ${KEYCLOAK_CA_CERT:+--cacert "${KEYCLOAK_CA_CERT}"} \
      -H "Authorization: Bearer ${admin_token}" \
      "${KEYCLOAK_BASE_URL}/admin/realms/${KEYCLOAK_REALM}/roles/workspace-admin" \
      2>/dev/null || true)
    if [[ -n "${role_json}" && "${role_json}" != "null" ]]; then
      curl -sS --fail ${KEYCLOAK_CA_CERT:+--cacert "${KEYCLOAK_CA_CERT}"} \
        -H "Authorization: Bearer ${admin_token}" \
        -H "Content-Type: application/json" \
        -X POST "${KEYCLOAK_BASE_URL}/admin/realms/${KEYCLOAK_REALM}/users/${user_id}/role-mappings/realm" \
        -d "[${role_json}]" >/dev/null 2>&1 || true
    fi
  fi

  USERNAME=${USERNAME:-$auto_username}
  PASSWORD=${PASSWORD:-$auto_password}
}

detect_with_kubectl() {
  if ! command -v kubectl >/dev/null 2>&1; then
    ensure_automation_user "" || true
    return
  fi

  local kc_json
  kc_json=$(kubectl get keycloak -A -o json 2>/dev/null || true)
  if [[ -z "${kc_json}" || "${kc_json}" == "{}" ]]; then
    ensure_automation_user "" || true
    return
  fi

  local detected_ns detected_host tls_secret
  detected_ns=$(echo "${kc_json}" | jq -r '.items[0].metadata.namespace // empty' 2>/dev/null || true)
  detected_host=$(echo "${kc_json}" | jq -r '.items[0].spec.hostname.hostname // empty' 2>/dev/null || true)
  tls_secret=$(echo "${kc_json}" | jq -r '.items[0].spec.http.tlsSecret // empty' 2>/dev/null || true)

  if [[ -n "${detected_ns}" ]]; then
    KEYCLOAK_NAMESPACE="${KEYCLOAK_NAMESPACE:-${detected_ns}}"
  fi
  local ns="${KEYCLOAK_NAMESPACE:-keycloak}"

  if [[ -z "${KEYCLOAK_BASE_URL:-}" && -n "${detected_host}" ]]; then
    if [[ "${detected_host}" =~ ^https?:// ]]; then
      KEYCLOAK_BASE_URL="${detected_host}"
    else
      KEYCLOAK_BASE_URL="https://${detected_host}"
    fi
  elif [[ -z "${KEYCLOAK_BASE_URL:-}" ]]; then
    local host
    host=$(kubectl get keycloak -n "${ns}" -o jsonpath='{.items[0].spec.hostname.hostname}' 2>/dev/null || true)
    if [[ -n "${host}" ]]; then
      if [[ "${host}" =~ ^https?:// ]]; then
        KEYCLOAK_BASE_URL="${host}"
      else
        KEYCLOAK_BASE_URL="https://${host}"
      fi
    fi
  fi

  if [[ -z "${KEYCLOAK_BASE_URL:-}" ]]; then
    local ingress_json ingress_host ingress_ns ingress_tls
    ingress_json=$(kubectl get ingress -A -o json 2>/dev/null || true)
    if [[ -n "${ingress_json}" ]]; then
      ingress_host=$(echo "${ingress_json}" | jq -r '.items[] | select((.metadata.labels["app.kubernetes.io/component"] // "") == "keycloak" or (.metadata.labels["app.kubernetes.io/name"] // "") == "keycloak" or ((.metadata.name // "")|test("keycloak";"i"))) | .spec.rules[]?.host | select(. != null and . != "")' 2>/dev/null | head -n1 || true)
      if [[ -n "${ingress_host}" ]]; then
        ingress_ns=$(echo "${ingress_json}" | jq -r '.items[] | select((.metadata.labels["app.kubernetes.io/component"] // "") == "keycloak" or (.metadata.labels["app.kubernetes.io/name"] // "") == "keycloak" or ((.metadata.name // "")|test("keycloak";"i"))) | .metadata.namespace' 2>/dev/null | head -n1 || true)
        ingress_tls=$(echo "${ingress_json}" | jq -r '.items[] | select((.metadata.labels["app.kubernetes.io/component"] // "") == "keycloak" or (.metadata.labels["app.kubernetes.io/name"] // "") == "keycloak" or ((.metadata.name // "")|test("keycloak";"i"))) | .spec.tls[0].secretName // empty' 2>/dev/null | head -n1 || true)
        if [[ -n "${ingress_ns}" ]]; then
          KEYCLOAK_NAMESPACE="${ingress_ns}"
          ns="${ingress_ns}"
        fi
        if [[ -n "${ingress_host}" ]]; then
          if [[ "${ingress_host}" =~ ^https?:// ]]; then
            KEYCLOAK_BASE_URL="${ingress_host}"
          else
            KEYCLOAK_BASE_URL="https://${ingress_host}"
          fi
        fi
        if [[ -n "${ingress_tls}" && -z "${KEYCLOAK_TLS_SECRET_NAME:-}" ]]; then
          KEYCLOAK_TLS_SECRET_NAME="${ingress_tls}"
        fi
      fi
    fi
  fi

  if [[ -z "${KEYCLOAK_BASE_URL:-}" ]]; then
    local svc_json svc_host svc_ns
    svc_json=$(kubectl get svc -A -o json 2>/dev/null || true)
    if [[ -n "${svc_json}" ]]; then
      svc_host=$(echo "${svc_json}" | jq -r '.items[] | select((.metadata.labels["app.kubernetes.io/component"] // "") == "keycloak" or (.metadata.labels["app.kubernetes.io/name"] // "") == "keycloak" or ((.metadata.name // "")|test("keycloak";"i"))) | .status.loadBalancer.ingress[0].hostname // .status.loadBalancer.ingress[0].ip // empty' 2>/dev/null | head -n1 || true)
      if [[ -n "${svc_host}" ]]; then
        svc_ns=$(echo "${svc_json}" | jq -r '.items[] | select((.metadata.labels["app.kubernetes.io/component"] // "") == "keycloak" or (.metadata.labels["app.kubernetes.io/name"] // "") == "keycloak" or ((.metadata.name // "")|test("keycloak";"i"))) | .metadata.namespace' 2>/dev/null | head -n1 || true)
        if [[ -n "${svc_ns}" ]]; then
          KEYCLOAK_NAMESPACE="${svc_ns}"
          ns="${svc_ns}"
        fi
        if [[ "${svc_host}" =~ ^https?:// ]]; then
          KEYCLOAK_BASE_URL="${svc_host}"
        else
          KEYCLOAK_BASE_URL="https://${svc_host}"
        fi
      fi
    fi
  fi

  if [[ -z "${KEYCLOAK_BASE_URL:-}" ]]; then
    local ingress_json ingress_host ingress_ns ingress_tls
    ingress_json=$(kubectl get ingress -A -o json 2>/dev/null || true)
    if [[ -n "${ingress_json}" ]]; then
      ingress_host=$(echo "${ingress_json}" | jq -r '.items[] | select((.metadata.name // "")|test("keycloak";"i")) | .spec.rules[]?.host | select(. != null and . != "")' 2>/dev/null | head -n1 || true)
      if [[ -n "${ingress_host}" ]]; then
        ingress_ns=$(echo "${ingress_json}" | jq -r '.items[] | select((.metadata.name // "")|test("keycloak";"i")) | .metadata.namespace' 2>/dev/null | head -n1 || true)
        ingress_tls=$(echo "${ingress_json}" | jq -r '.items[] | select((.metadata.name // "")|test("keycloak";"i")) | .spec.tls[0].secretName // empty' 2>/dev/null | head -n1 || true)
        if [[ -n "${ingress_ns}" ]]; then
          KEYCLOAK_NAMESPACE="${ingress_ns}"
          ns="${ingress_ns}"
        fi
        if [[ -n "${ingress_host}" ]]; then
          if [[ "${ingress_host}" =~ ^https?:// ]]; then
            KEYCLOAK_BASE_URL="${ingress_host}"
          else
            KEYCLOAK_BASE_URL="https://${ingress_host}"
          fi
        fi
        if [[ -n "${ingress_tls}" && -z "${KEYCLOAK_TLS_SECRET_NAME:-}" ]]; then
          KEYCLOAK_TLS_SECRET_NAME="${ingress_tls}"
        fi
      fi
    fi
  fi

  if [[ -z "${KEYCLOAK_BASE_URL:-}" ]]; then
    local svc_json svc_host svc_ns
    svc_json=$(kubectl get svc -A -o json 2>/dev/null || true)
    if [[ -n "${svc_json}" ]]; then
      svc_host=$(echo "${svc_json}" | jq -r '.items[] | select((.metadata.name // "")|test("keycloak";"i")) | .status.loadBalancer.ingress[0].hostname // .status.loadBalancer.ingress[0].ip // empty' 2>/dev/null | head -n1 || true)
      if [[ -n "${svc_host}" ]]; then
        svc_ns=$(echo "${svc_json}" | jq -r '.items[] | select((.metadata.name // "")|test("keycloak";"i")) | .metadata.namespace' 2>/dev/null | head -n1 || true)
        if [[ -n "${svc_ns}" ]]; then
          KEYCLOAK_NAMESPACE="${svc_ns}"
          ns="${svc_ns}"
        fi
        if [[ "${svc_host}" =~ ^https?:// ]]; then
          KEYCLOAK_BASE_URL="${svc_host}"
        else
          KEYCLOAK_BASE_URL="https://${svc_host}"
        fi
      fi
    fi
  fi

  if [[ -n "${tls_secret}" ]]; then
    KEYCLOAK_TLS_SECRET_NAME="${KEYCLOAK_TLS_SECRET_NAME:-${tls_secret}}"
  fi

  if [[ -z "${KEYCLOAK_CA_CERT:-}" ]]; then
    local secret_name="${KEYCLOAK_TLS_SECRET_NAME:-keycloak-tls}"
    local ca_tmp
    ca_tmp=$(mktemp) || true
    if [[ -n "${ca_tmp}" ]] && kubectl get secret "${secret_name}" -n "${ns}" -o 'jsonpath={.data.tls\.crt}' 2>/dev/null | base64 --decode >"${ca_tmp}" 2>/dev/null; then
      KEYCLOAK_CA_CERT="${ca_tmp}"
      cleanup_files+=("${ca_tmp}")
    else
      [[ -n "${ca_tmp}" ]] && rm -f "${ca_tmp}" 2>/dev/null || true
    fi
  fi

  if [[ -z "${CLIENT_SECRET}" ]]; then
    local client_secret_name="${KEYCLOAK_CLIENT_SECRET_NAME:-keycloak-backstage-client-secret}"
    CLIENT_SECRET=$(kubectl get secret "${client_secret_name}" -n "${ns}" -o 'jsonpath={.data.clientSecret}' 2>/dev/null | base64 --decode 2>/dev/null || true)
  fi

  ensure_automation_user "${ns}" || true
}

detect_with_kubectl

if [[ -z "${TOKEN_URL}" ]]; then
  if [[ -n "${KEYCLOAK_BASE_URL:-}" ]]; then
    TOKEN_URL="${KEYCLOAK_BASE_URL%/}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token"
  else
    echo "✖ KEYCLOAK_TOKEN_URL or KEYCLOAK_BASE_URL must be set" >&2
    exit 1
  fi
fi

if [[ -z "${CLIENT_ID}" ]]; then
  CLIENT_ID="backstage"
fi


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
