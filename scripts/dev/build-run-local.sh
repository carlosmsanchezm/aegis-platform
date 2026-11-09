#!/usr/bin/env bash
set -euo pipefail
pushd services/platform-api
go run ./... &
API_PID=$!
popd
sleep 1
if [[ -z "${KUBECONFIG:-}" ]]; then
  echo "KUBECONFIG must be set with access to the target cluster" >&2
  kill ${API_PID}
  exit 1
fi

AEGIS_CP_GRPC=localhost:8081 AEGIS_CLUSTER_ID=dev-eks KUBECONFIG=${KUBECONFIG} AEGIS_DISABLE_KUEUE=1 AEGIS_FLAVORS="cpu-small" \
  go run ./agents/k8s-agent/cmd &
OP_PID=$!
echo "API:$API_PID OPERATOR:$OP_PID"
wait
