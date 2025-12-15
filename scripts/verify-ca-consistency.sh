#!/bin/bash
# verify-ca-consistency.sh
#
# Verifies that the CA certificate and Keycloak TLS certificate in local-tls.yaml
# are consistent (Keycloak cert is signed by the CA).
#
# Run this before any helm upgrade to prevent CA mismatch issues.
# Usage: ./scripts/verify-ca-consistency.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VALUES_FILE="$REPO_ROOT/charts/aegis-services/values/local-tls.yaml"

echo "=== Aegis CA Consistency Verification ==="
echo ""

if [ ! -f "$VALUES_FILE" ]; then
  echo "ERROR: Values file not found: $VALUES_FILE"
  exit 1
fi

# Extract CA cert from values file
# The CA cert is in the oidc.caBundle.data section under platformApi.auth
echo "Extracting CA certificate from values file..."
CA_CERT=$(sed -n '/auth:/,/^  env:/p' "$VALUES_FILE" | \
  sed -n '/data: |/,/-----END CERTIFICATE-----/p' | \
  grep -v "data:" | sed 's/^[[:space:]]*//')

if [ -z "$CA_CERT" ]; then
  echo "ERROR: Could not extract CA certificate from values file"
  exit 1
fi

# Extract Keycloak cert from values file
# The Keycloak cert is under keycloak.tls.secret.cert
echo "Extracting Keycloak TLS certificate from values file..."
KC_CERT=$(sed -n '/^keycloak:/,/^[a-zA-Z]/p' "$VALUES_FILE" | \
  sed -n '/cert: |/,/-----END CERTIFICATE-----/p' | head -30 | \
  grep -v "cert:" | sed 's/^[[:space:]]*//')

if [ -z "$KC_CERT" ]; then
  echo "ERROR: Could not extract Keycloak certificate from values file"
  exit 1
fi

# Get CA subject
CA_SUBJECT=$(echo "$CA_CERT" | openssl x509 -noout -subject 2>/dev/null)
CA_FINGERPRINT=$(echo "$CA_CERT" | openssl x509 -noout -fingerprint -sha256 2>/dev/null)

# Get Keycloak cert issuer
KC_ISSUER=$(echo "$KC_CERT" | openssl x509 -noout -issuer 2>/dev/null)
KC_SUBJECT=$(echo "$KC_CERT" | openssl x509 -noout -subject 2>/dev/null)

echo ""
echo "CA Certificate:"
echo "  Subject: $CA_SUBJECT"
echo "  Fingerprint: $CA_FINGERPRINT"
echo ""
echo "Keycloak TLS Certificate:"
echo "  Subject: $KC_SUBJECT"
echo "  Issuer: $KC_ISSUER"
echo ""

# Verify consistency
# Both should contain "O=Aegis Dev, CN=Aegis Local Root CA" or similar matching pattern
CA_CN=$(echo "$CA_SUBJECT" | sed 's/.*CN *= *//' | sed 's/,.*//')
KC_ISSUER_CN=$(echo "$KC_ISSUER" | sed 's/.*CN *= *//' | sed 's/,.*//')

if [ "$CA_CN" != "$KC_ISSUER_CN" ]; then
  echo "=============================================="
  echo "ERROR: CA MISMATCH DETECTED!"
  echo "=============================================="
  echo ""
  echo "CA Subject CN: $CA_CN"
  echo "Keycloak Issuer CN: $KC_ISSUER_CN"
  echo ""
  echo "The Keycloak TLS certificate was NOT signed by the CA in the values file."
  echo "This will cause 'certificate signed by unknown authority' errors."
  echo ""
  echo "To fix this, run:"
  echo "  ./scripts/regenerate-matching-certs.sh"
  echo ""
  exit 1
fi

# Verify the cert chain
echo "Verifying certificate chain..."
VERIFY_RESULT=$(echo "$KC_CERT" | openssl verify -CAfile <(echo "$CA_CERT") 2>&1)

if echo "$VERIFY_RESULT" | grep -q "OK"; then
  echo ""
  echo "=============================================="
  echo "SUCCESS: CA consistency verified!"
  echo "=============================================="
  echo ""
  echo "The Keycloak TLS certificate is properly signed by the CA."
  echo "Safe to proceed with helm upgrade."
  exit 0
else
  echo ""
  echo "=============================================="
  echo "ERROR: Certificate chain verification failed!"
  echo "=============================================="
  echo "$VERIFY_RESULT"
  echo ""
  echo "To fix this, run:"
  echo "  ./scripts/regenerate-matching-certs.sh"
  exit 1
fi
