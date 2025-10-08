#!/usr/bin/env bash
set -Eeuo pipefail

trap 'rc=$?; echo "[e2e-platform-api] failed at line ${LINENO}" >&2; exit ${rc}' ERR

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing dependency: $1" >&2; exit 2; }
}

need grpcurl
need jq
need kubectl

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

ADDR="${AEGIS_GRPC_ADDR:-127.0.0.1:10081}"
TLS="${GRPC_TLS:-0}"
PLATFORM_NAMESPACE="${AEGIS_PLATFORM_NAMESPACE:-aegis-system}"
PLATFORM_SERVICE="${AEGIS_PLATFORM_SERVICE:-aegis-services-platform-api}"

GRPC_ARGS=(-import-path "$REPO_ROOT/proto" -proto aegis/v1/platform.proto)
if [[ "$TLS" == "1" ]]; then
  [[ -n "${GRPC_CA:-}" ]] && GRPC_ARGS+=(-cacert "$GRPC_CA")
  if [[ -n "${GRPC_CERT:-}" && -n "${GRPC_KEY:-}" ]]; then
    GRPC_ARGS+=(-cert "$GRPC_CERT" -key "$GRPC_KEY")
  fi
else
  GRPC_ARGS+=(-plaintext)
fi

EPOCH="$(date +%s)"
PROJECT_ID="p-e2e-${EPOCH}"
QUEUE="default"
FLAVOR="a10-mig-1g"
CLUSTER_ID="dev-local"
TMP_LIST="$(mktemp)"
FORWARD_LOG=""
FORWARD_PID=""

cleanup() {
  if [[ -n "$FORWARD_PID" ]]; then
    kill "$FORWARD_PID" >/dev/null 2>&1 || true
    wait "$FORWARD_PID" >/dev/null 2>&1 || true
  fi
  [[ -n "$FORWARD_LOG" ]] && rm -f "$FORWARD_LOG"
  rm -f "$TMP_LIST"
}

trap cleanup EXIT

say() { printf '==> %s\n' "$*"; }

WORKLOAD_NAMESPACE="${AEGIS_WORKLOAD_NAMESPACE:-aegis-workloads-local}"

ensure_namespace() {
  kubectl get namespace "$1" >/dev/null 2>&1 && return 0
  kubectl create namespace "$1" >/dev/null 2>&1
}

port_is_ready() {
  local host="$1" port="$2"
  (exec 3<>"/dev/tcp/${host}/${port}" && exec 3<&- && exec 3>&-) >/dev/null 2>&1
}

maybe_port_forward() {
  local addr="$1"
  local host="${addr%:*}"
  local port="${addr##*:}"

  case "$host" in
    127.0.0.1|localhost|::1) : ;;
    *) return 0 ;;
  esac

  if port_is_ready "$host" "$port"; then
    return 0
  fi

  FORWARD_LOG="$(mktemp)"
  kubectl -n "${PLATFORM_NAMESPACE}" port-forward svc/"${PLATFORM_SERVICE}" "${port}:8081" \
    >"$FORWARD_LOG" 2>&1 &
  FORWARD_PID=$!

  for _ in {1..20}; do
    sleep 0.5
    if port_is_ready "$host" "$port"; then
      return 0
    fi
    if ! kill -0 "$FORWARD_PID" >/dev/null 2>&1; then
      echo "kubectl port-forward exited prematurely" >&2
      cat "$FORWARD_LOG" >&2
      exit 1
    fi
  done

  echo "timed out waiting for port-forward on ${addr}" >&2
  cat "$FORWARD_LOG" >&2
  exit 1
}

ensure_namespace "$WORKLOAD_NAMESPACE"
maybe_port_forward "$ADDR"

say "UpsertFlavor"
grpcurl "${GRPC_ARGS[@]}" -d @ "$ADDR" aegis.v1.AegisPlatform/UpsertFlavor <<JSON >/dev/null
{ "flavor": { "name": "$FLAVOR", "resourceName": "nvidia.com/mig-1g.10gb", "gpuCount": 1, "priceUsdPerGpuHour": 0.5 } }
JSON

say "CreateProject $PROJECT_ID"
grpcurl "${GRPC_ARGS[@]}" -d @ "$ADDR" aegis.v1.AegisPlatform/CreateProject <<JSON >/dev/null
{ "project": { "id": "$PROJECT_ID", "displayName": "E2E", "ownerGroup": "aegis-dev" } }
JSON

say "UpsertQueue"
grpcurl "${GRPC_ARGS[@]}" -d @ "$ADDR" aegis.v1.AegisPlatform/UpsertQueue <<JSON >/dev/null
{ "queue": { "name": "$QUEUE", "projectId": "$PROJECT_ID", "priorityTier": "prod", "allowedFlavors": ["$FLAVOR"], "defaultMaxDurationSeconds": 600 } }
JSON

say "RegisterCluster"
grpcurl "${GRPC_ARGS[@]}" -d @ "$ADDR" aegis.v1.AegisPlatform/RegisterCluster <<JSON >/dev/null
{ "clusterId": "$CLUSTER_ID", "provider": "kind", "region": "us-local", "ilLevel": "il1", "labels": {"env": "dev"} }
JSON

say "Heartbeat"
grpcurl "${GRPC_ARGS[@]}" -d @ "$ADDR" aegis.v1.AegisPlatform/Heartbeat <<JSON >/dev/null
{ "clusterId": "$CLUSTER_ID", "ttfGpuSecondsP50": 5, "availableFlavors": [{"name": "$FLAVOR"}] }
JSON

say "SubmitWorkload"
SUBMIT_JSON="$(grpcurl "${GRPC_ARGS[@]}" -d @ "$ADDR" aegis.v1.AegisPlatform/SubmitWorkload <<JSON
{ "workload": { "projectId": "$PROJECT_ID", "queue": "$QUEUE", "workspace": { "flavor": "$FLAVOR", "image": "alpine:3.19", "command": ["sh","-c","echo hello; sleep 1"], "interactive": false } } }
JSON
)"

echo "$SUBMIT_JSON" | jq '.'
WID="$(echo "$SUBMIT_JSON" | jq -r '.id // empty')"
[[ -n "$WID" ]] || { echo "workload ID missing in response" >&2; exit 1; }

say "ListWorkloads"
grpcurl "${GRPC_ARGS[@]}" -d @ "$ADDR" aegis.v1.AegisPlatform/ListWorkloads <<JSON >"$TMP_LIST"
{ "projectId": "$PROJECT_ID" }
JSON
cat "$TMP_LIST" | jq '.'

grep -q "\"id\": \"$WID\"" "$TMP_LIST" || { echo "workload $WID not returned by ListWorkloads" >&2; exit 1; }

say "StartWorkload"
grpcurl "${GRPC_ARGS[@]}" -d @ "$ADDR" aegis.v1.AegisPlatform/StartWorkload <<JSON >/dev/null
{ "id": "$WID", "clusterId": "$CLUSTER_ID" }
JSON

say "GetWorkload"
GET_JSON="$(grpcurl "${GRPC_ARGS[@]}" -d @ "$ADDR" aegis.v1.AegisPlatform/GetWorkload <<JSON
{ "id": "$WID" }
JSON
)"
echo "$GET_JSON" | jq '.'
STATUS="$(echo "$GET_JSON" | jq -r '.status // empty')"
case "$STATUS" in
  RUNNING|PLACED|SUCCEEDED) ;;
  *) echo "unexpected workload status after start: $STATUS" >&2; exit 1 ;;
esac

say "AckWorkload"
grpcurl "${GRPC_ARGS[@]}" -d @ "$ADDR" aegis.v1.AegisPlatform/AckWorkload <<JSON >/dev/null
{ "id": "$WID", "status": "SUCCEEDED", "url": "k8s://default/job/$WID", "backend": "workspace" }
JSON

say "Verify terminal state"
FINAL_JSON="$(grpcurl "${GRPC_ARGS[@]}" -d @ "$ADDR" aegis.v1.AegisPlatform/GetWorkload <<JSON
{ "id": "$WID" }
JSON
)"
echo "$FINAL_JSON" | jq '.'
[[ "$(echo "$FINAL_JSON" | jq -r '.status')" == "SUCCEEDED" ]] || { echo "final status is not SUCCEEDED" >&2; exit 1; }

say "OK — e2e platform-api passed (project=$PROJECT_ID workload=$WID)"
