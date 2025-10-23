#!/usr/bin/env bash
set -euo pipefail

GRPC_ADDR="${GRPC_ADDR:-localhost:10081}"
if [[ $# -gt 0 && "${1}" != "" ]]; then
  WORKSPACE_IMAGE="${1}"
  shift
else
  WORKSPACE_IMAGE="${WORKSPACE_IMAGE:-aegis-workspace:latest}"
fi
WORKSPACE_NAMESPACE="${WORKSPACE_NAMESPACE:-aegis-workloads-local}"
AEGIS_USER_HEADER="${AEGIS_WORKSPACE_USER:-testuser@test.com}"
QUALITY="${VSCODE_QUALITY:-stable}"
COMMIT="${VSCODE_COMMIT:-}"
TIMEOUT_SECONDS="${WORKSPACE_TIMEOUT:-600}"
REH_READY_STRING="${REH_READY_STRING:-Extension host agent listening}"
GRPC_TLS="${GRPC_TLS:-0}"
GRPC_CA="${GRPC_CA:-}"
GRPC_TLS_SERVER_NAME="${GRPC_TLS_SERVER_NAME:-}"
OUTPUT_DIR=".aegis"
SESSION_FILE="${OUTPUT_DIR}/workspace-session.json"

usage() {
  cat <<EOF
Usage: $(basename "$0") [image]

Environment variables:
  GRPC_ADDR              Platform API address (default: localhost:10081)
  WORKSPACE_IMAGE        Container image to launch (default: aegis-workspace:latest or positional argument)
  VSCODE_COMMIT          VS Code commit hash expected by the client (required)
  VSCODE_QUALITY         VS Code channel to fetch (default: stable)
  WORKSPACE_NAMESPACE    Namespace for workloads (default: aegis-workloads-local)
  WORKSPACE_TIMEOUT      Seconds to wait for pod readiness (default: 600)
  AEGIS_WORKSPACE_USER   Value for the x-aegis-user header (default: testuser@test.com)
  GRPC_TLS               Set to 1 to enable TLS (default: 0)
  GRPC_CA                Path to CA bundle when GRPC_TLS=1
  GRPC_TLS_SERVER_NAME   Expected server name (SNI) when GRPC_TLS=1
EOF
}

find_workspace_pod() {
  local candidate=""
  candidate=$(kubectl get pods -n "${WORKSPACE_NAMESPACE}" -l "${LABEL_SELECTOR}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
  if [[ -n "${candidate}" ]]; then
    echo "${candidate}"
    return 0
  fi

  candidate=$(kubectl get pods -n "${WORKSPACE_NAMESPACE}" -l "aegis.yourorg.dev/workspace-id=${WORKLOAD_ID}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
  if [[ -n "${candidate}" ]]; then
    echo "${candidate}"
    return 0
  fi

  local pod_json=""
  if pod_json=$(kubectl get pods -n "${WORKSPACE_NAMESPACE}" -o json 2>/dev/null); then
    candidate=$(jq -r --arg id "${WORKLOAD_ID}" '[
        .items[]
        | select(
            (.metadata.labels["aegis.workload/id"] // "") == $id
            or (.metadata.labels["aegis.yourorg.dev/workspace-id"] // "") == $id
            or ((.metadata.name // "") | contains($id))
          )
        | .metadata.name
      ][0] // ""' <<<"${pod_json}" 2>/dev/null || true)
    if [[ -n "${candidate}" ]]; then
      echo "${candidate}"
      return 0
    fi
  fi

  return 1
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
  echo "✖ Unable to determine VSCODE_COMMIT automatically. Set VSCODE_COMMIT to your VS Code build hash." >&2
  exit 1
fi

if [[ ! "${COMMIT}" =~ ^[0-9a-f]{40}$ ]]; then
  echo "✖ VSCODE_COMMIT (${COMMIT}) must be a 40-character hexadecimal hash." >&2
  exit 1
fi

GRPCURL_OPTS=()
if [[ "${GRPC_TLS}" == "1" ]]; then
  if [[ -z "${GRPC_CA}" ]]; then
    echo "✖ GRPC_TLS=1 requires GRPC_CA to be set (path to CA bundle)." >&2
    exit 1
  fi
  if [[ -z "${GRPC_TLS_SERVER_NAME}" ]]; then
    echo "✖ GRPC_TLS=1 requires GRPC_TLS_SERVER_NAME to be set." >&2
    exit 1
  fi
  GRPCURL_OPTS=(-cacert "${GRPC_CA}" -authority "${GRPC_TLS_SERVER_NAME}")
else
  GRPCURL_OPTS=(-plaintext)
fi

for dep in grpcurl jq kubectl awk sed curl; do
  if ! command -v "${dep}" >/dev/null 2>&1; then
    echo "✖ ${dep} is required." >&2
    exit 1
  fi
done

ensure_project_queue() {
  grpcurl "${GRPCURL_OPTS[@]}" -d '{"project":{"id":"p-demo"}}' "${GRPC_ADDR}" aegis.v1.AegisPlatform/CreateProject >/dev/null || true
  grpcurl "${GRPCURL_OPTS[@]}" -d '{"flavor":{"name":"cpu-small","gpuCount":0}}' "${GRPC_ADDR}" aegis.v1.AegisPlatform/UpsertFlavor >/dev/null || true
  grpcurl "${GRPCURL_OPTS[@]}" -d '{
    "queue":{
      "name":"default",
      "projectId":"p-demo",
      "allowedFlavors":["cpu-small"]
    }
  }' "${GRPC_ADDR}" aegis.v1.AegisPlatform/UpsertQueue >/dev/null || true
}

ensure_project_queue

echo "→ Submitting workspace with image ${WORKSPACE_IMAGE}"
SUBMIT_JSON=$(
  grpcurl "${GRPCURL_OPTS[@]}" \
    -H "x-aegis-user: ${AEGIS_USER_HEADER}" \
    -d "{
      \"workload\": {
        \"projectId\": \"p-demo\",
        \"queue\": \"default\",
        \"workspace\": {
          \"flavor\": \"cpu-small\",
          \"interactive\": true,
          \"image\": \"${WORKSPACE_IMAGE}\",
          \"env\": {
            \"VSCODE_QUALITY\": \"${QUALITY}\",
            \"VSCODE_COMMIT\": \"${COMMIT}\"
          }
        }
      }
    }" "${GRPC_ADDR}" aegis.v1.AegisPlatform/SubmitWorkload
)

WORKLOAD_ID=$(echo "${SUBMIT_JSON}" | jq -r '.id')
if [[ -z "${WORKLOAD_ID}" || "${WORKLOAD_ID}" == "null" ]]; then
  echo "✖ Failed to extract workload ID from SubmitWorkload response." >&2
  echo "${SUBMIT_JSON}" >&2
  exit 1
fi
echo "→ Workload ID: ${WORKLOAD_ID}"

LABEL_SELECTOR="aegis.workload/id=${WORKLOAD_ID}"
DEADLINE=$((SECONDS + TIMEOUT_SECONDS))
POD_NAME=""
UNSCHEDULABLE_MSG=""

echo "→ Waiting for workspace pod to become Ready..."
while (( SECONDS < DEADLINE )); do
  if POD_CANDIDATE=$(find_workspace_pod); then
    POD_NAME="${POD_CANDIDATE}"
    POD_JSON=$(kubectl get pod "${POD_NAME}" -n "${WORKSPACE_NAMESPACE}" -o json 2>/dev/null || true)
    if [[ -n "${POD_JSON}" && "${POD_JSON}" != "null" ]]; then
      PHASE=$(echo "${POD_JSON}" | jq -r '.status.phase // ""')
      READY=$(echo "${POD_JSON}" | jq -r '.status.containerStatuses[0].ready // "false"')
      SCHEDULED_STATUS=$(echo "${POD_JSON}" | jq -r '.status.conditions[]? | select(.type=="PodScheduled") | .status // empty')
      if [[ "${SCHEDULED_STATUS}" == "False" ]]; then
        REASON=$(echo "${POD_JSON}" | jq -r '.status.conditions[]? | select(.type=="PodScheduled") | .reason // empty')
        MESSAGE=$(echo "${POD_JSON}" | jq -r '.status.conditions[]? | select(.type=="PodScheduled") | .message // empty')
        if [[ "${REASON}" == "Unschedulable" ]]; then
          UNSCHEDULABLE_MSG="${MESSAGE}"
          break
        fi
      fi
      if [[ "${PHASE}" == "Running" && "${READY}" == "true" ]]; then
        echo "→ Pod ${POD_NAME} is Ready."
        break
      fi
    fi
  else
    POD_NAME=""
  fi
  sleep 3
done

if [[ -z "${POD_NAME}" ]]; then
  if [[ -n "${UNSCHEDULABLE_MSG}" ]]; then
    echo "✖ Workspace pod is unschedulable: ${UNSCHEDULABLE_MSG}" >&2
  else
    echo "✖ Workspace pod did not schedule." >&2
  fi
  exit 1
fi

PHASE=$(kubectl get pod "${POD_NAME}" -n "${WORKSPACE_NAMESPACE}" -o jsonpath='{.status.phase}')
if [[ "${PHASE}" != "Running" ]]; then
  if [[ -n "${UNSCHEDULABLE_MSG}" ]]; then
    echo "✖ Workspace pod is unschedulable: ${UNSCHEDULABLE_MSG}" >&2
  else
    echo "✖ Workspace pod failed to reach Running phase (current: ${PHASE})." >&2
  fi
  kubectl describe pod "${POD_NAME}" -n "${WORKSPACE_NAMESPACE}" >&2 || true
  exit 1
fi

echo "→ Waiting for Remote Extension Host log '${REH_READY_STRING}'..."
REH_DEADLINE=$((SECONDS + 120))
REH_READY=0
while (( SECONDS < REH_DEADLINE )); do
  if kubectl logs "${POD_NAME}" -n "${WORKSPACE_NAMESPACE}" --tail=400 | grep -q "${REH_READY_STRING}"; then
    REH_READY=1
    break
  fi
  sleep 6
done

if [[ "${REH_READY}" != "1" ]]; then
  echo "✖ VS Code Remote Extension Host did not report readiness (${REH_READY_STRING})." >&2
  kubectl logs "${POD_NAME}" -n "${WORKSPACE_NAMESPACE}" >&2 || true
  exit 1
fi
echo "✔ Remote Extension Host reported readiness"

echo "→ Probing VS Code server on pod ${POD_NAME}:11111/version"
if ! kubectl exec "${POD_NAME}" -n "${WORKSPACE_NAMESPACE}" -- curl -sf --max-time 5 http://127.0.0.1:11111/version >/dev/null; then
  echo "✖ VS Code server did not respond to in-cluster curl http://127.0.0.1:11111/version" >&2
  exit 1
fi
echo "✔ VS Code server responded to in-cluster curl"

ASSIGNED_CLUSTER_ID=$(echo "${SUBMIT_JSON}" | jq -r '.clusterId // empty')
if [[ -z "${ASSIGNED_CLUSTER_ID}" ]]; then
  WORKLOAD_STATE=$(
    grpcurl "${GRPCURL_OPTS[@]}" \
      -H "x-aegis-user: ${AEGIS_USER_HEADER}" \
      -d "{\"id\":\"${WORKLOAD_ID}\"}" \
      "${GRPC_ADDR}" aegis.v1.AegisPlatform/GetWorkload || true
  )
  if [[ -n "${WORKLOAD_STATE}" && "${WORKLOAD_STATE}" != "null" ]]; then
    ASSIGNED_CLUSTER_ID=$(echo "${WORKLOAD_STATE}" | jq -r '.clusterId // empty')
  fi
fi

if [[ -n "${ASSIGNED_CLUSTER_ID}" ]]; then
  echo "→ Marking workload ${WORKLOAD_ID} as RUNNING via StartWorkload (${ASSIGNED_CLUSTER_ID})"
  grpcurl "${GRPCURL_OPTS[@]}" \
    -H "x-aegis-user: ${AEGIS_USER_HEADER}" \
    -d "{\"id\":\"${WORKLOAD_ID}\",\"clusterId\":\"${ASSIGNED_CLUSTER_ID}\"}" \
    "${GRPC_ADDR}" aegis.v1.AegisPlatform/StartWorkload >/dev/null || true
fi

echo "→ Requesting connection session..."
SESSION_JSON=$(
  grpcurl "${GRPCURL_OPTS[@]}" \
    -H "x-aegis-user: ${AEGIS_USER_HEADER}" \
    -d "{\"workload_id\":\"${WORKLOAD_ID}\",\"client\":\"vscode\"}" \
    "${GRPC_ADDR}" aegis.v1.AegisPlatform/CreateConnectionSession
)

VSCODE_URI=$(echo "${SESSION_JSON}" | jq -r '.vscodeUri')
SSH_CONFIG=$(echo "${SESSION_JSON}" | jq -r '.sshConfig')

if [[ -z "${VSCODE_URI}" || "${VSCODE_URI}" == "null" ]]; then
  echo "✖ CreateConnectionSession did not return a vscode URI." >&2
  echo "${SESSION_JSON}" >&2
  exit 1
fi

echo "✔ VS Code deep link: ${VSCODE_URI}"
echo
echo "SSH config snippet:"
echo "${SSH_CONFIG}"

mkdir -p "${OUTPUT_DIR}"
echo "${SESSION_JSON}" | jq --arg workload_id "${WORKLOAD_ID}" --arg pod "${POD_NAME}" --arg image "${WORKSPACE_IMAGE}" --arg commit "${COMMIT}" '
  . + {
    workloadId: $workload_id,
    podName: $pod,
    workspaceImage: $image,
    vscodeCommit: $commit
  }
' | jq > "${SESSION_FILE}"

echo
echo "ℹ︎ Workspace details saved to ${SESSION_FILE}"
echo "   Workload: ${WORKLOAD_ID}"
echo "   Pod:      ${POD_NAME}"

if [[ "${CLEANUP:-0}" == "1" ]]; then
  echo "→ CLEANUP=1 detected; tearing down workspace ${WORKLOAD_ID}"
  kubectl delete aegisworkload "${WORKLOAD_ID}" -n "${WORKSPACE_NAMESPACE}" >/dev/null 2>&1 || true
else
  cat <<EOF

Next steps:
  1. Open Backstage or the VS Code extension and connect to workspace ${WORKLOAD_ID}.
  2. Use the SSH snippet above (or the saved JSON) if you need manual SSH access.
  3. When finished testing, clean up with:
       kubectl delete aegisworkload ${WORKLOAD_ID} -n ${WORKSPACE_NAMESPACE}

EOF
fi
