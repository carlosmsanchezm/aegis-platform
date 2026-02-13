#!/usr/bin/env bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TUNNEL_NAME="${TUNNEL_NAME:-aegis-platform}"
DOMAIN="${DOMAIN:-aegis-platform.tech}"
PLATFORM_HOSTNAME="${PLATFORM_HOSTNAME:-remote}"
KEYCLOAK_HOSTNAME="${KEYCLOAK_HOSTNAME:-keycloak}"
NAMESPACE="${NAMESPACE:-aegis-system}"
KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"

# Derived values
PLATFORM_FQDN="${PLATFORM_HOSTNAME}.${DOMAIN}"
KEYCLOAK_FQDN="${KEYCLOAK_HOSTNAME}.${DOMAIN}"
TUNNEL_CONFIG_FILE="$PWD/cloudflared-config.yaml"

# Usage
usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Setup Cloudflare Tunnel for Aegis Platform and Keycloak

OPTIONS:
    -t, --tunnel-name NAME      Tunnel name (default: aegis-platform)
    -d, --domain DOMAIN         Domain name (default: aegis-platform.tech)
    -p, --platform-host HOST    Platform hostname (default: remote)
    -k, --keycloak-host HOST    Keycloak hostname (default: keycloak)
    -n, --namespace NAMESPACE   Kubernetes namespace (default: aegis-system)
    -c, --clean                 Clean up existing tunnel and recreate
    -h, --help                  Show this help message

ENVIRONMENT VARIABLES:
    TUNNEL_NAME                 Override tunnel name
    DOMAIN                      Override domain
    PLATFORM_HOSTNAME           Override platform hostname
    KEYCLOAK_HOSTNAME           Override keycloak hostname
    NAMESPACE                   Override Kubernetes namespace
    KUBECONFIG                  Path to kubeconfig (default: ~/.kube/config)

EXAMPLES:
    # Setup with defaults
    $0

    # Setup with custom domain
    $0 --domain example.com

    # Clean up and recreate
    $0 --clean

    # Custom hostnames
    $0 --platform-host api --keycloak-host auth

EOF
}

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*" >&2
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    local missing=0

    if ! command -v cloudflared &> /dev/null; then
        log_error "cloudflared not found. Install: brew install cloudflared"
        missing=1
    fi

    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not found. Install: brew install kubectl"
        missing=1
    fi

    if ! command -v jq &> /dev/null; then
        log_error "jq not found. Install: brew install jq"
        missing=1
    fi

    if [[ $missing -eq 1 ]]; then
        exit 1
    fi

    # Check kubectl access
    if ! kubectl get nodes &> /dev/null; then
        log_error "Cannot access Kubernetes cluster. Check KUBECONFIG or context."
        exit 1
    fi

    # Check if cloudflared is logged in
    if [[ ! -f "$HOME/.cloudflared/cert.pem" ]]; then
        log_warn "Cloudflared not logged in. Running login..."
        cloudflared tunnel login
        if [[ ! -f "$HOME/.cloudflared/cert.pem" ]]; then
            log_error "Login failed. Please authenticate with Cloudflare."
            exit 1
        fi
    fi

    log_success "Prerequisites OK"
}

# Clean up existing tunnel
cleanup_tunnel() {
    log_info "Cleaning up existing tunnel configuration..."

    # Delete Kubernetes resources
    kubectl delete deployment cloudflared -n "$NAMESPACE" --ignore-not-found
    kubectl delete configmap cloudflared-config -n "$NAMESPACE" --ignore-not-found
    kubectl delete secret cloudflared-credentials -n "$NAMESPACE" --ignore-not-found
    kubectl delete secret cloudflared-tunnel-token -n "$NAMESPACE" --ignore-not-found

    # Delete local config file
    rm -f "$TUNNEL_CONFIG_FILE"

    # Note: We don't delete the tunnel itself or DNS records to avoid breaking production
    log_warn "Tunnel '$TUNNEL_NAME' and DNS records NOT deleted (manual cleanup required)"
    log_warn "To delete: cloudflared tunnel delete $TUNNEL_NAME"

    log_success "Cleanup complete"
}

# Get or create tunnel
get_or_create_tunnel() {
    log_info "Checking for existing tunnel '$TUNNEL_NAME'..."

    local tunnel_id
    tunnel_id=$(cloudflared tunnel list --output json 2>/dev/null | jq -r ".[] | select(.name == \"$TUNNEL_NAME\") | .id" || echo "")

    if [[ -n "$tunnel_id" ]]; then
        log_success "Tunnel '$TUNNEL_NAME' exists (ID: $tunnel_id)"
        echo "$tunnel_id"
        return 0
    fi

    log_info "Creating tunnel '$TUNNEL_NAME'..."
    cloudflared tunnel create "$TUNNEL_NAME" 2>&1 | tee /tmp/tunnel-create.log

    tunnel_id=$(cloudflared tunnel list --output json 2>/dev/null | jq -r ".[] | select(.name == \"$TUNNEL_NAME\") | .id")

    if [[ -z "$tunnel_id" ]]; then
        log_error "Failed to create tunnel"
        exit 1
    fi

    log_success "Tunnel created (ID: $tunnel_id)"
    echo "$tunnel_id"
}

# Get tunnel credentials
get_tunnel_credentials() {
    local tunnel_id="$1"
    local cred_file="$HOME/.cloudflared/${tunnel_id}.json"

    if [[ ! -f "$cred_file" ]]; then
        log_error "Credentials file not found: $cred_file"
        exit 1
    fi

    cat "$cred_file"
}

# Setup DNS routes
setup_dns_routes() {
    local tunnel_name="$1"

    log_info "Setting up DNS routes..."

    # Platform API route
    log_info "Creating DNS route for $PLATFORM_FQDN..."
    if cloudflared tunnel route dns "$tunnel_name" "$PLATFORM_HOSTNAME" 2>&1 | grep -q "already configured\|Added CNAME"; then
        log_success "DNS route for $PLATFORM_FQDN configured"
    else
        log_error "Failed to create DNS route for $PLATFORM_FQDN"
        exit 1
    fi

    # Keycloak route
    log_info "Creating DNS route for $KEYCLOAK_FQDN..."
    if cloudflared tunnel route dns "$tunnel_name" "$KEYCLOAK_HOSTNAME" 2>&1 | grep -q "already configured\|Added CNAME"; then
        log_success "DNS route for $KEYCLOAK_FQDN configured"
    else
        log_error "Failed to create DNS route for $KEYCLOAK_FQDN"
        exit 1
    fi
}

# Get Kubernetes service details
get_k8s_services() {
    log_info "Detecting Kubernetes services..."

    # Platform API
    local platform_svc
    platform_svc=$(kubectl get svc -n aegis-system -o name 2>/dev/null | grep platform-api | head -1 | cut -d/ -f2)
    if [[ -z "$platform_svc" ]]; then
        log_error "Platform API service not found in aegis-system namespace"
        exit 1
    fi
    log_success "Platform API: $platform_svc.aegis-system.svc.cluster.local:8081"

    # Keycloak
    local keycloak_svc
    keycloak_svc=$(kubectl get svc -n keycloak -o name 2>/dev/null | grep keycloak-service | head -1 | cut -d/ -f2)
    if [[ -z "$keycloak_svc" ]]; then
        log_error "Keycloak service not found in keycloak namespace"
        exit 1
    fi
    log_success "Keycloak: $keycloak_svc.keycloak.svc.cluster.local:8443"

    echo "$platform_svc" "$keycloak_svc"
}

# Create Kubernetes resources
create_k8s_resources() {
    local tunnel_id="$1"
    local tunnel_credentials="$2"
    local platform_svc="$3"
    local keycloak_svc="$4"

    log_info "Creating Kubernetes resources..."

    # Create namespace if it doesn't exist
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f - > /dev/null

    cat > "$TUNNEL_CONFIG_FILE" <<EOF
---
apiVersion: v1
kind: Secret
metadata:
  name: cloudflared-credentials
  namespace: $NAMESPACE
type: Opaque
stringData:
  credentials.json: |
    $tunnel_credentials
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudflared-config
  namespace: $NAMESPACE
data:
  config.yaml: |
    tunnel: $tunnel_id
    credentials-file: /etc/cloudflared/creds/credentials.json

    ingress:
      - hostname: $PLATFORM_FQDN
        service: https://${platform_svc}.aegis-system.svc.cluster.local:8081
        originRequest:
          noTLSVerify: true
          httpHostHeader: platform-api-grpc.localtest.me
          originServerName: platform-api-grpc.localtest.me
          http2Origin: true
      - hostname: $KEYCLOAK_FQDN
        service: https://${keycloak_svc}.keycloak.svc.cluster.local:8443
        originRequest:
          noTLSVerify: true
          httpHostHeader: keycloak.localtest.me
          originServerName: keycloak.localtest.me
      - service: http_status:404
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudflared
  namespace: $NAMESPACE
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cloudflared
  template:
    metadata:
      labels:
        app: cloudflared
    spec:
      containers:
      - name: cloudflared
        image: cloudflare/cloudflared:latest
        args:
        - tunnel
        - --config
        - /etc/cloudflared/config/config.yaml
        - run
        volumeMounts:
        - name: config
          mountPath: /etc/cloudflared/config
          readOnly: true
        - name: creds
          mountPath: /etc/cloudflared/creds
          readOnly: true
        resources:
          limits:
            cpu: 100m
            memory: 128Mi
          requests:
            cpu: 50m
            memory: 64Mi
      volumes:
      - name: config
        configMap:
          name: cloudflared-config
      - name: creds
        secret:
          secretName: cloudflared-credentials
EOF

    log_info "Applying Kubernetes manifest..."
    kubectl apply -f "$TUNNEL_CONFIG_FILE"

    log_success "Kubernetes resources created: $TUNNEL_CONFIG_FILE"
}

# Wait for deployment
wait_for_deployment() {
    log_info "Waiting for cloudflared deployment to be ready..."

    if kubectl rollout status deployment/cloudflared -n "$NAMESPACE" --timeout=60s; then
        log_success "Cloudflared deployment ready"
    else
        log_error "Deployment failed to become ready"
        exit 1
    fi
}

# Verify tunnel connectivity
verify_tunnel() {
    log_info "Verifying tunnel connectivity..."

    sleep 5  # Give DNS a moment to propagate

    # Check platform API
    log_info "Testing $PLATFORM_FQDN..."
    if curl -I -s --max-time 10 "https://$PLATFORM_FQDN" | head -1 | grep -q "HTTP"; then
        log_success "$PLATFORM_FQDN is reachable"
    else
        log_warn "$PLATFORM_FQDN may not be reachable yet (DNS propagation can take a few minutes)"
    fi

    # Check Keycloak
    log_info "Testing $KEYCLOAK_FQDN..."
    if curl -I -s --max-time 10 "https://$KEYCLOAK_FQDN" | head -1 | grep -q "HTTP"; then
        log_success "$KEYCLOAK_FQDN is reachable"
    else
        log_warn "$KEYCLOAK_FQDN may not be reachable yet (DNS propagation can take a few minutes)"
    fi

    # Check tunnel status
    log_info "Checking tunnel status..."
    cloudflared tunnel info "$TUNNEL_NAME" | grep -q "active connection" && \
        log_success "Tunnel has active connections" || \
        log_warn "No active connections reported (check cloudflared logs)"
}

# Display summary
display_summary() {
    local tunnel_id="$1"

    cat <<EOF

${GREEN}═══════════════════════════════════════════════════════════════${NC}
${GREEN}  Cloudflare Tunnel Setup Complete!${NC}
${GREEN}═══════════════════════════════════════════════════════════════${NC}

${BLUE}Tunnel Information:${NC}
  Name:       $TUNNEL_NAME
  ID:         $tunnel_id
  Config:     $TUNNEL_CONFIG_FILE

${BLUE}Public Endpoints:${NC}
  Platform:   https://$PLATFORM_FQDN
  Keycloak:   https://$KEYCLOAK_FQDN

${BLUE}Kubernetes Resources:${NC}
  Namespace:  $NAMESPACE
  Deployment: cloudflared
  ConfigMap:  cloudflared-config
  Secret:     cloudflared-credentials

${BLUE}Spoke Configuration (AWS/Remote):${NC}
  AEGIS_CP_GRPC: $PLATFORM_FQDN:443
  AEGIS_CP_GRPC_INSECURE: false
  AEGIS_CP_GRPC_SKIP_VERIFY: true
  AEGIS_CP_GRPC_SERVER_NAME: $PLATFORM_FQDN
  AEGIS_CP_OIDC_TOKEN_URL: https://$KEYCLOAK_FQDN/realms/aegis/protocol/openid-connect/token

${BLUE}Useful Commands:${NC}
  # Check tunnel status
  cloudflared tunnel info $TUNNEL_NAME

  # View cloudflared logs
  kubectl logs -n $NAMESPACE deployment/cloudflared -f

  # Restart cloudflared
  kubectl rollout restart deployment/cloudflared -n $NAMESPACE

  # Test endpoints
  curl -I https://$PLATFORM_FQDN
  curl -I https://$KEYCLOAK_FQDN

  # Delete tunnel (WARNING: breaks connectivity)
  kubectl delete deployment cloudflared -n $NAMESPACE
  cloudflared tunnel delete $TUNNEL_NAME

${GREEN}═══════════════════════════════════════════════════════════════${NC}

EOF
}

# Main function
main() {
    local clean_mode=0

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--tunnel-name)
                TUNNEL_NAME="$2"
                shift 2
                ;;
            -d|--domain)
                DOMAIN="$2"
                shift 2
                ;;
            -p|--platform-host)
                PLATFORM_HOSTNAME="$2"
                shift 2
                ;;
            -k|--keycloak-host)
                KEYCLOAK_HOSTNAME="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -c|--clean)
                clean_mode=1
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    # Update derived values
    PLATFORM_FQDN="${PLATFORM_HOSTNAME}.${DOMAIN}"
    KEYCLOAK_FQDN="${KEYCLOAK_HOSTNAME}.${DOMAIN}"

    log_info "Starting Cloudflare Tunnel setup..."
    log_info "Tunnel: $TUNNEL_NAME"
    log_info "Domain: $DOMAIN"
    log_info "Platform: $PLATFORM_FQDN"
    log_info "Keycloak: $KEYCLOAK_FQDN"
    echo ""

    check_prerequisites

    if [[ $clean_mode -eq 1 ]]; then
        cleanup_tunnel
        echo ""
    fi

    tunnel_id=$(get_or_create_tunnel)
    tunnel_credentials=$(get_tunnel_credentials "$tunnel_id")

    setup_dns_routes "$TUNNEL_NAME"

    read -r platform_svc keycloak_svc <<< "$(get_k8s_services)"

    create_k8s_resources "$tunnel_id" "$tunnel_credentials" "$platform_svc" "$keycloak_svc"

    wait_for_deployment

    verify_tunnel

    display_summary "$tunnel_id"
}

# Run main
main "$@"
