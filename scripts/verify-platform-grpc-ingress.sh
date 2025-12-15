#!/usr/bin/env bash
#
# verify-platform-grpc-ingress.sh
#
# Validates that the gRPC ingress is configured with the correct upstream protocol
# (GRPC vs GRPCS) based on whether the platform-api gRPC port (8081) is serving TLS.
#
# This prevents the common failure mode where ingress is configured with:
#   nginx.ingress.kubernetes.io/backend-protocol: "GRPC"
# while platform-api is serving TLS on 8081, causing 502s for gRPC clients (e.g. VS Code extension).
#
# Usage:
#   ./scripts/verify-platform-grpc-ingress.sh
#
set -euo pipefail

NS="${NS:-aegis-system}"
INGRESS_NAME="${INGRESS_NAME:-aegis-services-platform-api}"
SERVICE_NAME="${SERVICE_NAME:-aegis-services-platform-api}"
SERVICE_GRPC_PORT="${SERVICE_GRPC_PORT:-8081}"

INGRESS_HOST="${INGRESS_HOST:-platform-api-grpc.localtest.me}"
INGRESS_PORT="${INGRESS_PORT:-443}"

CA_FILE_DEFAULT="$HOME/aegis-local-trust.pem"
CA_FILE="${CA_FILE:-$CA_FILE_DEFAULT}"

log() { printf '%s\n' "$*"; }
fail() { printf '%s\n' "$*" >&2; exit 1; }

if ! command -v kubectl >/dev/null 2>&1; then
  fail "kubectl not found; cannot verify ingress configuration"
fi
if ! command -v openssl >/dev/null 2>&1; then
  fail "openssl not found; cannot verify backend TLS"
fi

backend_protocol="$(
  kubectl get ingress -n "$NS" "$INGRESS_NAME" \
    -o jsonpath='{.metadata.annotations.nginx\.ingress\.kubernetes\.io/backend-protocol}' 2>/dev/null || true
)"
backend_protocol="${backend_protocol:-<missing>}"

local_port="$(
  python3 - <<'PY'
import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PY
)"

pf_log="$(mktemp)"
pf_pid=""
cleanup() {
  if [[ -n "${pf_pid}" ]]; then
    kill "${pf_pid}" >/dev/null 2>&1 || true
    wait "${pf_pid}" >/dev/null 2>&1 || true
  fi
  rm -f "$pf_log"
}
trap cleanup EXIT

kubectl port-forward -n "$NS" "svc/${SERVICE_NAME}" "${local_port}:${SERVICE_GRPC_PORT}" --address 127.0.0.1 >"$pf_log" 2>&1 &
pf_pid="$!"

ready=0
for _ in $(seq 1 50); do
  if nc -z 127.0.0.1 "$local_port" >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 0.1
done
if [[ "$ready" -ne 1 ]]; then
  log "kubectl port-forward did not become ready; last logs:"
  tail -20 "$pf_log" || true
  fail "failed to port-forward svc/${SERVICE_NAME}:${SERVICE_GRPC_PORT}"
fi

backend_tls=0
if echo | openssl s_client -servername "$INGRESS_HOST" -connect "127.0.0.1:${local_port}" 2>/dev/null | openssl x509 -noout -subject >/dev/null 2>&1; then
  backend_tls=1
fi

expected_protocol="GRPC"
if [[ "$backend_tls" -eq 1 ]]; then
  expected_protocol="GRPCS"
fi

log "Platform API gRPC backend TLS: $([[ "$backend_tls" -eq 1 ]] && echo yes || echo no)"
log "Ingress annotation backend-protocol: ${backend_protocol}"
log "Expected backend-protocol: ${expected_protocol}"

if [[ "${backend_protocol}" != "${expected_protocol}" ]]; then
  fail "$(
    cat <<EOF
ERROR: gRPC ingress is misconfigured.

The platform-api gRPC service on ${SERVICE_GRPC_PORT} is $([[ "$backend_tls" -eq 1 ]] && echo "TLS-enabled" || echo "plaintext"), but the ingress is configured as:
  nginx.ingress.kubernetes.io/backend-protocol: ${backend_protocol}

Fix (quick):
  kubectl annotate ingress -n ${NS} ${INGRESS_NAME} nginx.ingress.kubernetes.io/backend-protocol=${expected_protocol} --overwrite

Fix (durable):
  redeploy using the Helm local TLS overlay so this annotation is rendered correctly.
EOF
  )"
fi

if command -v grpcurl >/dev/null 2>&1 && [[ -f "$CA_FILE" ]]; then
  endpoint="${INGRESS_HOST}:${INGRESS_PORT}"
  log ""
  log "Verifying end-to-end gRPC through ingress (${endpoint})..."
  out="$(grpcurl -cacert "$CA_FILE" -authority "$INGRESS_HOST" "$endpoint" list 2>&1 || true)"

  if echo "$out" | grep -q "unexpected HTTP status code received from server: 502"; then
    fail "ERROR: ingress returned 502 for gRPC (upstream protocol mismatch)."
  fi
  if echo "$out" | grep -qi "Failed to dial target host"; then
    fail "ERROR: grpcurl could not dial ${endpoint} (network/DNS/ingress issue)."
  fi

  # Expected in local dev: auth required, so we should at least get a gRPC status back.
  if echo "$out" | grep -q "Unauthenticated" || echo "$out" | grep -q "does not support the reflection API" || echo "$out" | grep -q "Unimplemented"; then
    log "✅ gRPC ingress reachable (expected auth/reflection error)"
  else
    log "NOTE: grpcurl output did not match expected patterns; raw output:"
    printf '%s\n' "$out"
  fi
else
  log ""
  log "Skipping end-to-end grpcurl check (grpcurl not installed or CA file missing: $CA_FILE)"
fi

log ""
log "✅ Platform API gRPC ingress looks correctly configured"
