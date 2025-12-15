#!/bin/bash
# regenerate-matching-certs.sh
#
# Regenerates CA + Keycloak TLS certificates that match each other.
# Use this when CA mismatch issues are detected.
#
# This script:
# 1. Generates a new CA certificate
# 2. Generates a new Keycloak TLS certificate signed by that CA
# 3. Updates both Kubernetes secrets
# 4. Restarts affected services
# 5. Outputs commands to update local-tls.yaml for permanent fix
#
# Usage: ./scripts/regenerate-matching-certs.sh

set -e

echo "=== Aegis Certificate Regeneration ==="
echo ""
echo "This will generate new CA and Keycloak TLS certificates."
echo "Existing certificates will be replaced."
echo ""

cd /tmp

echo "=== Step 1: Generating New CA ==="
openssl genrsa -out aegis-ca.key 4096
openssl req -x509 -new -nodes -key aegis-ca.key -sha256 -days 3650 \
  -subj "/O=Aegis Dev/CN=Aegis Local Root CA" -out aegis-ca.crt

echo ""
echo "=== Step 2: Generating Keycloak TLS Cert ==="
openssl genrsa -out keycloak.key 2048
openssl req -new -key keycloak.key -subj "/CN=keycloak.localtest.me" -out keycloak.csr

cat > keycloak-san.cnf << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
[req_distinguished_name]
[v3_req]
subjectAltName = @alt_names
[alt_names]
DNS.1 = keycloak.localtest.me
DNS.2 = aegis-services-keycloak.keycloak.svc.cluster.local
DNS.3 = aegis-services-keycloak
DNS.4 = keycloak.aegis-platform.tech
EOF

openssl x509 -req -in keycloak.csr -CA aegis-ca.crt -CAkey aegis-ca.key \
  -CAcreateserial -out keycloak.crt -days 1095 -sha256 \
  -extfile keycloak-san.cnf -extensions v3_req

echo ""
echo "=== Step 3: Verifying Certificate Chain ==="
openssl verify -CAfile aegis-ca.crt keycloak.crt

echo ""
echo "=== Step 4: Updating Kubernetes Secrets ==="

# Update Keycloak TLS secret
kubectl -n keycloak create secret tls keycloak-tls \
  --cert=keycloak.crt --key=keycloak.key --dry-run=client -o yaml | kubectl apply -f -

# Update Platform-API OIDC CA secret
kubectl -n aegis-system create secret generic aegis-platform-api-oidc-ca \
  --from-file=ca.crt=aegis-ca.crt --dry-run=client -o yaml | kubectl apply -f -

echo ""
echo "=== Step 5: Restarting Services ==="
kubectl rollout restart statefulset/aegis-services-keycloak -n keycloak
kubectl rollout restart deployment/aegis-services-platform-api -n aegis-system

echo ""
echo "=== Step 6: Waiting for Rollout ==="
kubectl rollout status statefulset/aegis-services-keycloak -n keycloak --timeout=120s
kubectl rollout status deployment/aegis-services-platform-api -n aegis-system --timeout=120s

echo ""
echo "=============================================="
echo "SUCCESS: Certificates regenerated and applied!"
echo "=============================================="
echo ""
echo "The new certificates are now active in the cluster."
echo ""
echo "⚠️  IMPORTANT: To make this PERMANENT, update local-tls.yaml:"
echo ""
echo "1. Copy the CA certificate to platformApi.auth.oidc.caBundle.data:"
echo "   cat /tmp/aegis-ca.crt"
echo ""
echo "2. Copy the Keycloak cert to keycloak.tls.secret.cert:"
echo "   cat /tmp/keycloak.crt"
echo ""
echo "3. Copy the Keycloak key to keycloak.tls.secret.key:"
echo "   cat /tmp/keycloak.key"
echo ""
echo "4. Verify with: ./scripts/verify-ca-consistency.sh"
echo ""
