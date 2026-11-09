#!/bin/bash
# generate-spoke-proxy-ca.sh - Generate a CA for signing spoke-proxy certificates
#
# This creates a CA that will be used by the platform-api's Pulumi runner to sign
# spoke-proxy TLS certificates. The CA cert should be added to the VS Code extension's
# trust bundle so all spoke-proxy certs are automatically trusted.
#
# Usage: ./scripts/generate-spoke-proxy-ca.sh [--install]
#   --install: Also create/update the Kubernetes secret

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CA_DIR="${SCRIPT_DIR}/../.pki/spoke-proxy-ca"
INSTALL_SECRET=false

if [[ "${1:-}" == "--install" ]]; then
    INSTALL_SECRET=true
fi

mkdir -p "$CA_DIR"

# Generate CA if it doesn't exist
if [[ ! -f "$CA_DIR/ca.crt" ]] || [[ ! -f "$CA_DIR/ca.key" ]]; then
    echo "Generating spoke-proxy CA..."

    # Generate CA private key
    openssl genrsa -out "$CA_DIR/ca.key" 4096

    # Generate CA certificate
    openssl req -x509 -new -nodes \
        -key "$CA_DIR/ca.key" \
        -sha256 \
        -days 3650 \
        -out "$CA_DIR/ca.crt" \
        -subj "/O=Aegis Platform/CN=Aegis Spoke-Proxy CA" \
        -addext "basicConstraints=critical,CA:TRUE,pathlen:0" \
        -addext "keyUsage=critical,keyCertSign,cRLSign"

    echo "CA generated at: $CA_DIR/"
else
    echo "CA already exists at: $CA_DIR/"
fi

# Display CA info
echo ""
echo "CA Certificate:"
openssl x509 -in "$CA_DIR/ca.crt" -noout -subject -dates

# Create/update Kubernetes secret
if $INSTALL_SECRET; then
    echo ""
    echo "Installing CA secret to Kubernetes..."

    kubectl create secret generic spoke-proxy-ca \
        --namespace aegis-system \
        --from-file=ca.crt="$CA_DIR/ca.crt" \
        --from-file=ca.key="$CA_DIR/ca.key" \
        --dry-run=client -o yaml | kubectl apply -f -

    echo "Secret 'spoke-proxy-ca' created/updated in aegis-system namespace"
fi

echo ""
echo "=== Next Steps ==="
echo ""
echo "1. Add the CA to your trust bundle:"
echo "   cat $CA_DIR/ca.crt >> ~/aegis-local-trust.pem"
echo ""
echo "2. Configure platform-api to use the CA (add to helm values or deployment):"
echo "   AEGIS_SPOKE_PROXY_CA_CERT=/etc/spoke-proxy-ca/ca.crt"
echo "   AEGIS_SPOKE_PROXY_CA_KEY=/etc/spoke-proxy-ca/ca.key"
echo ""
echo "3. Mount the secret in platform-api deployment:"
echo "   volumeMounts:"
echo "     - name: spoke-proxy-ca"
echo "       mountPath: /etc/spoke-proxy-ca"
echo "       readOnly: true"
echo "   volumes:"
echo "     - name: spoke-proxy-ca"
echo "       secret:"
echo "         secretName: spoke-proxy-ca"
