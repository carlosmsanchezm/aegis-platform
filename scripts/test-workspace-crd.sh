#!/usr/bin/env bash
set -euo pipefail

WORKSPACE_PROFILE="${WORKSPACE_PROFILE:-vscode-python}"
WORKSPACE_PROJECT="${WORKSPACE_PROJECT:-p-demo}"
WORKSPACE_QUEUE="${WORKSPACE_QUEUE:-default}"
WORKSPACE_NAMESPACE="${WORKSPACE_NAMESPACE:-default}"
WORKSPACE_NAME="${WORKSPACE_NAME:-ws-$(date +%s)}"
TIMEOUT_SECONDS="${WORKSPACE_TIMEOUT:-240}"
OUTPUT_DIR=".aegis"

usage() {
  cat <<EOF
Usage: $(basename "$0")

Environment variables:
  WORKSPACE_PROFILE      Persona/template to launch (default: vscode-python)
  WORKSPACE_PROJECT      Project identifier (default: p-demo)
  WORKSPACE_QUEUE        Queue name (default: default)
  WORKSPACE_NAMESPACE    Namespace for the Workspace resource (default: default)
  WORKSPACE_TIMEOUT      Seconds to wait for pod readiness (default: 240)
  WORKSPACE_NAME         Name for the Workspace resource (default: ws-<timestamp>)
EOF
}

if [[ "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

for dep in kubectl jq; do
  if ! command -v "${dep}" >/dev/null 2>&1; then
    echo "✖ ${dep} is required." >&2
    exit 1
  fi
done

mkdir -p "${OUTPUT_DIR}"

cat <<EOF | kubectl apply -f - >/dev/null
apiVersion: aegis.yourorg.dev/v1alpha2
kind: Workspace
metadata:
  name: ${WORKSPACE_NAME}
  namespace: ${WORKSPACE_NAMESPACE}
spec:
  projectRef: ${WORKSPACE_PROJECT}
  queue: ${WORKSPACE_QUEUE}
  profileRef: ${WORKSPACE_PROFILE}
EOF

echo "→ Created Workspace ${WORKSPACE_NAMESPACE}/${WORKSPACE_NAME}"

echo "→ Waiting for Workspace to expose workload reference..."
DEADLINE=$((SECONDS + TIMEOUT_SECONDS))
WORKLOAD_ID=""
while (( SECONDS < DEADLINE )); do
  WORKLOAD_ID=$(kubectl get workspace "${WORKSPACE_NAME}" -n "${WORKSPACE_NAMESPACE}" -o jsonpath='{.status.workloadRef.name}' 2>/dev/null || true)
  if [[ -n "${WORKLOAD_ID}" ]]; then
    break
  fi
  sleep 3
done

if [[ -z "${WORKLOAD_ID}" ]]; then
  echo "✖ Workspace did not produce an AegisWorkload reference." >&2
  kubectl describe workspace "${WORKSPACE_NAME}" -n "${WORKSPACE_NAMESPACE}" >&2 || true
  exit 1
fi

echo "→ Workspace references workload ${WORKLOAD_ID}"

LABEL_SELECTOR="aegis.yourorg.dev/workspace=${WORKSPACE_NAME}"
POD_NAME=""

echo "→ Waiting for workspace pod to become Ready..."
DEADLINE=$((SECONDS + TIMEOUT_SECONDS))
while (( SECONDS < DEADLINE )); do
  POD_NAME=$(kubectl get pods -n "${WORKSPACE_NAMESPACE}" -l "${LABEL_SELECTOR}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
  if [[ -n "${POD_NAME}" ]]; then
    PHASE=$(kubectl get pod "${POD_NAME}" -n "${WORKSPACE_NAMESPACE}" -o jsonpath='{.status.phase}')
    READY=$(kubectl get pod "${POD_NAME}" -n "${WORKSPACE_NAMESPACE}" -o jsonpath='{.status.containerStatuses[0].ready}' 2>/dev/null || echo "false")
    if [[ "${PHASE}" == "Running" && "${READY}" == "true" ]]; then
      echo "→ Pod ${POD_NAME} is Ready."
      break
    fi
  fi
  sleep 3
done

if [[ -z "${POD_NAME}" ]]; then
  echo "✖ Workspace pod did not schedule." >&2
  exit 1
fi

PHASE=$(kubectl get pod "${POD_NAME}" -n "${WORKSPACE_NAMESPACE}" -o jsonpath='{.status.phase}')
if [[ "${PHASE}" != "Running" ]]; then
  echo "✖ Workspace pod failed to reach Running phase (current: ${PHASE})." >&2
  kubectl describe pod "${POD_NAME}" -n "${WORKSPACE_NAMESPACE}" >&2 || true
  exit 1
fi

echo "✔ Workspace pod ${POD_NAME} is Running"

kubectl get workspace "${WORKSPACE_NAME}" -n "${WORKSPACE_NAMESPACE}" -o yaml | tee "${OUTPUT_DIR}/workspace-${WORKSPACE_NAME}.yaml" >/dev/null
kubectl get aegisworkload "${WORKLOAD_ID}" -n "${WORKSPACE_NAMESPACE}" -o yaml | tee "${OUTPUT_DIR}/aegisworkload-${WORKLOAD_ID}.yaml" >/dev/null

cat <<EOF

Next steps:
  1. Connect via VS Code extension to workload ${WORKLOAD_ID}.
  2. Inspect saved manifests under ${OUTPUT_DIR}/.
  3. When finished testing, clean up with:
       kubectl delete workspace ${WORKSPACE_NAME} -n ${WORKSPACE_NAMESPACE}
EOF
