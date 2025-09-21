#!/usr/bin/env bash
set -euo pipefail
pushd services/platform-api
go run ./... &
API_PID=$!
popd
sleep 1
AEGIS_CP_GRPC=localhost:8081 AEGIS_CLUSTER_ID=dev-eks AEGIS_REGION=us-east AEGIS_PROVIDER=DEV \
  go run ./agents/k8s-agent/cmd/agent &
AGENT1=$!
AEGIS_CP_GRPC=localhost:8081 AEGIS_CLUSTER_ID=dev-gke AEGIS_REGION=us-central AEGIS_PROVIDER=DEV \
  go run ./agents/k8s-agent/cmd/agent &
AGENT2=$!
echo "API:$API_PID AGENTS:$AGENT1,$AGENT2"
wait
