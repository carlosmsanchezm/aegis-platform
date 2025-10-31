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
port_forward_pid=""
port_forward_log=""
port_forward_port=""
port_forward_host=""
keycloak_host_header=""
keycloak_remote_port=""
keycloak_port_forward_log=""

cleanup() {
  local status=$?
  if [[ ${#cleanup_files[@]} -gt 0 ]]; then
    rm -f "${cleanup_files[@]}" 2>/dev/null || true
  fi
  if [[ -n "${port_forward_pid}" ]]; then
    kill "${port_forward_pid}" >/dev/null 2>&1 || true
    wait "${port_forward_pid}" >/dev/null 2>&1 || true
  fi
  if [[ -n "${port_forward_log}" ]]; then
    rm -f "${port_forward_log}" 2>/dev/null || true
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
  local auto_username="${USERNAME:-cloud@test.com}"
  local auto_password="${PASSWORD:-password}"
  local admin_secret="${KEYCLOAK_ADMIN_SECRET_NAME:-keycloak-admin-secret}"
  local admin_user_key="${KEYCLOAK_ADMIN_USERNAME_KEY:-username}"
  local admin_pass_key="${KEYCLOAK_ADMIN_PASSWORD_KEY:-password}"
  local base_url="${KEYCLOAK_BASE_URL%/}"
  local realm="${KEYCLOAK_REALM:-aegis}"
  local ca_path=""
  local curl_base=(-sS --fail)

  if [[ -n "${USERNAME:-}" && -n "${PASSWORD:-}" ]]; then
    return
  fi

  if [[ -z "${base_url}" ]]; then
    base_url="https://keycloak.localtest.me"
  fi

  if [[ -n "${KEYCLOAK_CA_CERT:-}" && -f "${KEYCLOAK_CA_CERT}" ]]; then
    ca_path="${KEYCLOAK_CA_CERT}"
  elif [[ -n "${KEYCLOAK_CA_BUNDLE:-}" && -f "${KEYCLOAK_CA_BUNDLE}" ]]; then
    ca_path="${KEYCLOAK_CA_BUNDLE}"
  elif [[ -f "${HOME}/keycloak.localtest.me.crt" ]]; then
    ca_path="${HOME}/keycloak.localtest.me.crt"
  fi
  if [[ -n "${ca_path}" ]]; then
    curl_base+=(--cacert "${ca_path}")
  fi

  if [[ -n "${keycloak_host_header}" ]]; then
    curl_base+=(-H "Host: ${keycloak_host_header}")
  fi

  if [[ "${KEYCLOAK_SKIP_TLS_VERIFY:-0}" == "1" ]]; then
    curl_base+=("--insecure")
  fi

  if [[ -z "${ns}" ]] || ! command -v kubectl >/dev/null 2>&1; then
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
  admin_token=$(curl "${curl_base[@]}" \
    -X POST "${base_url}/realms/master/protocol/openid-connect/token" \
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
  user_json=$(curl "${curl_base[@]}" \
    -H "Authorization: Bearer ${admin_token}" \
    "${base_url}/admin/realms/${realm}/users?search=${encoded_username}&exact=true" \
    2>/dev/null || true)
  user_id=$(echo "${user_json}" | jq -r --arg username "${auto_username}" 'map(select(.username==$username)) | .[0].id // empty' 2>/dev/null || true)

  if [[ -z "${user_id}" || "${user_id}" == "null" ]]; then
    local payload
    payload=$(printf '{"username":"%s","email":"%s","firstName":"cloud","lastName":"user","enabled":true,"emailVerified":true,"credentials":[{"type":"password","value":"%s","temporary":false}]}' "${auto_username}" "${auto_username}" "${auto_password}")
    curl "${curl_base[@]}" \
      -H "Authorization: Bearer ${admin_token}" \
      -H "Content-Type: application/json" \
      -X POST "${base_url}/admin/realms/${realm}/users" \
      -d "${payload}" >/dev/null 2>&1 || true
    user_json=$(curl "${curl_base[@]}" \
      -H "Authorization: Bearer ${admin_token}" \
      "${base_url}/admin/realms/${realm}/users?search=${encoded_username}&exact=true" \
      2>/dev/null || true)
    user_id=$(echo "${user_json}" | jq -r --arg username "${auto_username}" 'map(select(.username==$username)) | .[0].id // empty' 2>/dev/null || true)
  else
    local payload
    payload=$(printf '{"type":"password","value":"%s","temporary":false}' "${auto_password}")
    curl "${curl_base[@]}" \
      -H "Authorization: Bearer ${admin_token}" \
      -H "Content-Type: application/json" \
      -X PUT "${base_url}/admin/realms/${realm}/users/${user_id}/reset-password" \
      -d "${payload}" >/dev/null 2>&1 || true
  fi

  if [[ -n "${user_id}" && "${user_id}" != "null" ]]; then
    local role_json
    role_json=$(curl "${curl_base[@]}" \
      -H "Authorization: Bearer ${admin_token}" \
      "${base_url}/admin/realms/${realm}/roles/workspace-admin" \
      2>/dev/null || true)
    if [[ -n "${role_json}" && "${role_json}" != "null" ]]; then
      curl "${curl_base[@]}" \
        -H "Authorization: Bearer ${admin_token}" \
        -H "Content-Type: application/json" \
        -X POST "${base_url}/admin/realms/${realm}/users/${user_id}/role-mappings/realm" \
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

}

maybe_port_forward() {
  if [[ -n "${port_forward_pid}" ]]; then
    return
  fi
  if [[ -z "${KEYCLOAK_BASE_URL:-}" ]]; then
    return
  fi
  if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
    printf 'keycloak-token: base_raw=%q\n' "${KEYCLOAK_BASE_URL}" >&2
  fi
  if ! command -v python3 >/dev/null 2>&1; then
    return
  fi
  if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
    echo "keycloak-token: maybe_port_forward base=${KEYCLOAK_BASE_URL}" >&2
  fi

  local parsed scheme host port remote_port
  if ! parsed=$(python3 - "$KEYCLOAK_BASE_URL" <<'PY'
from urllib.parse import urlparse
import sys
raw = sys.argv[1]
url_str = raw.strip()
url = urlparse(url_str)
scheme = url.scheme or ''
host = url.hostname or ''
port = url.port or 0
if port == 0:
    if scheme == 'https':
        port = 8443
    elif scheme == 'http':
        port = 8080
remote = port
print(url_str)
print(scheme)
print(host)
print(port)
print(remote)
PY
  ); then
    return
  fi
  if [[ -z "${parsed}" ]]; then
    return
  fi
  read -r KEYCLOAK_BASE_URL scheme host port remote_port <<<"${parsed}"
  if [[ -z "${scheme}" || -z "${host}" ]]; then
    local base_without_scheme="${KEYCLOAK_BASE_URL#*://}"
    if [[ "${base_without_scheme}" != "${KEYCLOAK_BASE_URL}" ]]; then
      local authority="${base_without_scheme%%/*}"
      if [[ -z "${scheme}" ]]; then
        scheme="${KEYCLOAK_BASE_URL%%://*}"
      fi
      if [[ "${authority}" == *":"* ]]; then
        host="${authority%%:*}"
        port="${authority##*:}"
      else
        host="${authority}"
      fi
      if [[ -z "${remote_port}" || "${remote_port}" == "0" ]]; then
        if [[ "${scheme}" == "https" ]]; then
          remote_port=8443
        elif [[ "${scheme}" == "http" ]]; then
          remote_port=8080
        fi
      fi
    fi
  fi
  if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
    echo "keycloak-token: parsed scheme=${scheme} host=${host} port=${port}" >&2
  fi
  if [[ -z "${host}" ]]; then
    if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
      echo "keycloak-token: aborting port-forward due to empty host" >&2
    fi
    return
  fi
  if [[ "${KEYCLOAK_PORT_FORWARD:-0}" != "1" && "${host}" != *".svc."* && "${host}" != *".cluster.local" ]]; then
    return
  fi
  if ! command -v kubectl >/dev/null 2>&1; then
    return
  fi

  local ns="${KEYCLOAK_NAMESPACE:-}"
  local svc="${KEYCLOAK_SERVICE_NAME:-}"
  if [[ -z "${svc}" ]]; then
    svc="${host%%.*}"
  fi
  if [[ -z "${ns}" && "${host}" == *"."* ]]; then
    local remainder="${host#${svc}.}"
    if [[ "${remainder}" != "${host}" ]]; then
      ns="${remainder%%.*}"
    fi
  fi
  ns=${ns:-keycloak}
  KEYCLOAK_NAMESPACE="${ns}"
  KEYCLOAK_SERVICE_NAME="${svc}"
  if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
    echo "keycloak-token: kube targets ns=${ns} svc=${svc}" >&2
  fi
  if [[ -z "${svc}" ]]; then
    svc=$(kubectl get svc -n "${ns}" -l app.kubernetes.io/component=keycloak -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
  fi
  if [[ -z "${svc}" ]]; then
    if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
      echo "keycloak-token: unable to determine keycloak service in namespace ${ns}" >&2
    fi
    return
  fi

  local remote="${remote_port:-8443}"
  local local_port="${KEYCLOAK_PORT_FORWARD_PORT:-}"
  if [[ -z "${local_port}" || "${local_port}" == "0" ]]; then
    local_port="${remote}"
  fi

  if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
    echo "keycloak-token: port-forward target ns=${ns} svc=${svc} remote=${remote} local=${local_port}" >&2
  fi

  port_forward_log=$(mktemp)
  keycloak_port_forward_log="${port_forward_log}"
  kubectl -n "${ns}" port-forward "svc/${svc}" "${local_port}:${remote}" --address 127.0.0.1 >"${port_forward_log}" 2>&1 &
  port_forward_pid=$!
  if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
    echo "keycloak-token: port-forwarding svc/${svc} in ${ns} on ${local_port}:${remote}" >&2
  fi
  for _ in {1..50}; do
    if command -v nc >/dev/null 2>&1; then
      if nc -z 127.0.0.1 "${local_port}" >/dev/null 2>&1; then
        port_forward_port="${local_port}"
        port_forward_host="127.0.0.1"
        if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
          echo "keycloak-token: port-forward ready host=${port_forward_host} port=${port_forward_port}" >&2
        fi
        break
      fi
    else
      if python3 - "${local_port}" <<'PY'
import socket
import sys
port = int(sys.argv[1])
with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
    sock.settimeout(0.1)
    try:
        sock.connect(("127.0.0.1", port))
    except OSError:
        sys.exit(1)
PY
      then
        port_forward_port="${local_port}"
        port_forward_host="127.0.0.1"
        if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
          echo "keycloak-token: port-forward ready host=${port_forward_host} port=${port_forward_port}" >&2
        fi
        break
      fi
    fi
    if ! kill -0 "${port_forward_pid}" >/dev/null 2>&1; then
      cat "${port_forward_log}" >&2 || true
      port_forward_pid=""
      return
    fi
    sleep 0.2
  done
  if [[ -z "${port_forward_port}" ]]; then
    if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
      echo "keycloak-token: port-forward failed; details:" >&2
      cat "${port_forward_log}" >&2 || true
    fi
    echo "✖ Failed to establish port-forward to Keycloak" >&2
    return
  fi

  keycloak_host_header="${host}"
  keycloak_remote_port="${remote}"
  KEYCLOAK_SKIP_TLS_VERIFY="${KEYCLOAK_SKIP_TLS_VERIFY:-1}"
}

detect_with_kubectl

if [[ -z "${KEYCLOAK_BASE_URL:-}" ]]; then
  ns_hint="${KEYCLOAK_NAMESPACE:-}"
  svc_hint="${KEYCLOAK_SERVICE_NAME:-}"
  host_hint=""
  if [[ -z "${ns_hint}" ]]; then
    if [[ -n "${PREVIEW_NAMESPACE:-}" ]]; then
      ns_hint="${PREVIEW_NAMESPACE}"
    elif [[ -n "${AEGIS_PLATFORM_NAMESPACE:-}" ]]; then
      ns_hint="${AEGIS_PLATFORM_NAMESPACE}"
    fi
  fi
  if [[ -z "${svc_hint}" ]]; then
    if [[ -n "${PREVIEW_RELEASE:-}" ]]; then
      svc_hint="${PREVIEW_RELEASE}-keycloak-service"
      host_hint="${PREVIEW_RELEASE}-keycloak"
    elif [[ -n "${KEYCLOAK_SERVICE_BASENAME:-}" ]]; then
      svc_hint="${KEYCLOAK_SERVICE_BASENAME}"
    fi
  elif [[ -n "${PREVIEW_RELEASE:-}" && "${svc_hint}" == "${PREVIEW_RELEASE}-keycloak" ]]; then
    host_hint="${PREVIEW_RELEASE}-keycloak"
  fi
  if [[ -n "${svc_hint}" ]]; then
    KEYCLOAK_SERVICE_NAME="${svc_hint}"
  fi
  if [[ -n "${ns_hint}" ]]; then
    KEYCLOAK_NAMESPACE="${ns_hint}"
  fi
  if [[ -z "${host_hint}" && -n "${KEYCLOAK_SERVICE_NAME:-}" && -n "${KEYCLOAK_NAMESPACE:-}" ]]; then
    host_hint="${KEYCLOAK_SERVICE_NAME}.${KEYCLOAK_NAMESPACE}.svc.cluster.local"
  elif [[ -n "${host_hint}" && -n "${KEYCLOAK_NAMESPACE:-}" ]]; then
    host_hint="${host_hint}.${KEYCLOAK_NAMESPACE}.svc.cluster.local"
  fi
  if [[ -z "${KEYCLOAK_BASE_URL:-}" && -n "${host_hint}" ]]; then
    port_hint="${KEYCLOAK_INTERNAL_PORT:-${KEYCLOAK_PORT:-8443}}"
    KEYCLOAK_BASE_URL="https://${host_hint}:${port_hint}"
  fi
fi

maybe_port_forward

ensure_automation_user "${KEYCLOAK_NAMESPACE:-keycloak}" || true

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

if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
  echo "keycloak-token: resolved host=${port_forward_host:-} port=${port_forward_port:-} base=${KEYCLOAK_BASE_URL}" >&2
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
if [[ -n "${port_forward_host}" && -n "${keycloak_host_header}" && -n "${keycloak_remote_port}" ]]; then
  CURL_ARGS+=(--resolve "${keycloak_host_header}:${keycloak_remote_port}:${port_forward_host}")
  CURL_ARGS+=(--connect-to "${keycloak_host_header}:${keycloak_remote_port}:${port_forward_host}:${keycloak_remote_port}")
  CURL_ARGS+=("--tlsv1.2")
fi
if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
  CURL_ARGS+=("-v")
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
if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
  printf 'keycloak-token: curl args -> ' >&2
  printf '%q ' "${CURL_ARGS[@]}" >&2
  printf '\n' >&2
fi
trap 'rm -f "${TMP_BODY}"' EXIT

HTTP_STATUS=$(curl "${CURL_ARGS[@]}" -w '%{http_code}' -o "${TMP_BODY}" || true)
if [[ "${KEYCLOAK_DEBUG:-0}" == "1" ]]; then
  echo "keycloak-token: http_status=${HTTP_STATUS}" >&2
  echo "keycloak-token: response_body=$(cat "${TMP_BODY}")" >&2
  if [[ -n "${keycloak_port_forward_log}" && -f "${keycloak_port_forward_log}" ]]; then
    echo "keycloak-token: port-forward log:" >&2
    cat "${keycloak_port_forward_log}" >&2
  fi
fi
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
