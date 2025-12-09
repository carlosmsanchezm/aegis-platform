#!/bin/bash
# Start ngrok tunnels for local development with remote AWS spoke clusters
# This exposes the local platform-api (gRPC) and keycloak (HTTPS) to the internet
# so that AWS-based spoke clusters can connect to them.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configuration
NGROK_CONFIG_FILE="${PROJECT_ROOT}/ngrok-aegis.yml"
VALUES_OUTPUT_FILE="${PROJECT_ROOT}/charts/aegis-spoke/values-ngrok.yaml"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check prerequisites
check_prerequisites() {
    if ! command -v ngrok &> /dev/null; then
        log_error "ngrok is not installed. Install it with: brew install ngrok"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        log_error "jq is not installed. Install it with: brew install jq"
        exit 1
    fi

    # Check if ngrok is authenticated
    if ! ngrok config check &> /dev/null; then
        log_error "ngrok is not configured. Run: ngrok config add-authtoken <your-token>"
        exit 1
    fi
}

# Get local service ports via kubectl port-forward
get_local_ports() {
    # Platform API gRPC port (exposed via ingress on 443, but we'll use port-forward)
    PLATFORM_API_PORT=8181
    KEYCLOAK_PORT=8543

    log_info "Will use local ports: platform-api=$PLATFORM_API_PORT, keycloak=$KEYCLOAK_PORT"
}

# Create ngrok configuration file
create_ngrok_config() {
    log_info "Creating ngrok configuration at $NGROK_CONFIG_FILE"

    cat > "$NGROK_CONFIG_FILE" << 'EOF'
version: "3"
tunnels:
  platform-api:
    proto: http
    addr: 8181
    schemes:
      - https
    inspect: true
    # gRPC requires HTTP/2, ngrok handles this automatically
  keycloak:
    proto: http
    addr: 8543
    schemes:
      - https
    inspect: true
EOF

    log_info "ngrok configuration created"
}

# Start port-forwards to local kubernetes services
start_port_forwards() {
    log_info "Starting kubectl port-forwards..."

    # Kill any existing port-forwards
    pkill -f "kubectl port-forward.*8181:8081" 2>/dev/null || true
    pkill -f "kubectl port-forward.*8543:8443" 2>/dev/null || true
    sleep 1

    # Ensure we're using docker-desktop context
    kubectl config use-context docker-desktop

    # Start platform-api port-forward (gRPC on 8081)
    kubectl port-forward svc/aegis-services-platform-api -n aegis-system 8181:8081 &
    PF_PLATFORM_PID=$!
    log_info "Platform API port-forward started (PID: $PF_PLATFORM_PID)"

    # Start keycloak port-forward (HTTPS on 8443)
    kubectl port-forward svc/aegis-services-keycloak-service -n keycloak 8543:8443 &
    PF_KEYCLOAK_PID=$!
    log_info "Keycloak port-forward started (PID: $PF_KEYCLOAK_PID)"

    # Save PIDs for cleanup
    echo "$PF_PLATFORM_PID" > /tmp/aegis-pf-platform.pid
    echo "$PF_KEYCLOAK_PID" > /tmp/aegis-pf-keycloak.pid

    sleep 3

    # Verify port-forwards are working
    if ! curl -sk https://localhost:8181 &>/dev/null; then
        log_warn "Platform API port-forward may not be ready yet"
    fi
    if ! curl -sk https://localhost:8543 &>/dev/null; then
        log_warn "Keycloak port-forward may not be ready yet"
    fi
}

# Start ngrok tunnels
start_ngrok() {
    log_info "Starting ngrok tunnels..."

    # Kill any existing ngrok processes
    pkill -f "ngrok start" 2>/dev/null || true
    sleep 1

    # Start ngrok with our config
    ngrok start --config "$NGROK_CONFIG_FILE" --all &
    NGROK_PID=$!
    echo "$NGROK_PID" > /tmp/aegis-ngrok.pid

    log_info "ngrok started (PID: $NGROK_PID)"

    # Wait for tunnels to be established
    log_info "Waiting for tunnels to be established..."
    sleep 5
}

# Get tunnel URLs from ngrok API
get_tunnel_urls() {
    log_info "Fetching tunnel URLs from ngrok API..."

    local max_retries=10
    local retry=0

    while [ $retry -lt $max_retries ]; do
        TUNNELS_JSON=$(curl -s http://localhost:4040/api/tunnels 2>/dev/null)

        if [ -n "$TUNNELS_JSON" ] && [ "$(echo "$TUNNELS_JSON" | jq -r '.tunnels | length')" -gt 0 ]; then
            break
        fi

        retry=$((retry + 1))
        log_warn "Waiting for ngrok API... (attempt $retry/$max_retries)"
        sleep 2
    done

    if [ $retry -eq $max_retries ]; then
        log_error "Failed to get tunnel URLs from ngrok API"
        exit 1
    fi

    # Extract URLs
    PLATFORM_API_URL=$(echo "$TUNNELS_JSON" | jq -r '.tunnels[] | select(.name=="platform-api") | .public_url' | sed 's|https://||')
    KEYCLOAK_URL=$(echo "$TUNNELS_JSON" | jq -r '.tunnels[] | select(.name=="keycloak") | .public_url' | sed 's|https://||')

    if [ -z "$PLATFORM_API_URL" ] || [ -z "$KEYCLOAK_URL" ]; then
        log_error "Failed to extract tunnel URLs"
        echo "Tunnels JSON: $TUNNELS_JSON"
        exit 1
    fi

    log_info "Platform API URL: https://$PLATFORM_API_URL"
    log_info "Keycloak URL: https://$KEYCLOAK_URL"
}

# Generate Helm values file
generate_helm_values() {
    log_info "Generating Helm values file at $VALUES_OUTPUT_FILE"

    cat > "$VALUES_OUTPUT_FILE" << EOF
# Auto-generated values for ngrok-based connectivity
# Generated at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
#
# Platform API ngrok URL: https://${PLATFORM_API_URL}
# Keycloak ngrok URL: https://${KEYCLOAK_URL}
#
# Usage:
#   helm upgrade aegis-spoke ./charts/aegis-spoke -n aegis-system \\
#     -f charts/aegis-spoke/values.yaml \\
#     -f charts/aegis-spoke/values-ngrok.yaml \\
#     --set k8sAgent.env.AEGIS_CLUSTER_ID=<your-cluster-id> \\
#     --set k8sAgent.env.AEGIS_REGION=<your-region> \\
#     --set k8sAgent.env.AEGIS_PROVIDER=aws

k8sAgent:
  image:
    pullPolicy: Always
    tag: "dev"
  env:
    # gRPC to platform-api via ngrok (supports HTTP/2 and gRPC natively)
    AEGIS_CP_GRPC: "${PLATFORM_API_URL}:443"
    AEGIS_CP_GRPC_INSECURE: "false"
    AEGIS_CP_GRPC_SKIP_VERIFY: "false"
    AEGIS_CP_GRPC_SERVER_NAME: "${PLATFORM_API_URL}"
    AEGIS_FLAVORS: "cpu-small,t4-1gpu"
    AEGIS_DEFAULT_IMAGE: "docker.io/carlosmsanchez/aegis-workspace-vscode:latest"

    # OIDC client credentials (Keycloak via ngrok)
    AEGIS_CP_OIDC_TOKEN_URL: "https://${KEYCLOAK_URL}/realms/aegis/protocol/openid-connect/token"
    AEGIS_CP_OIDC_CLIENT_ID: "spoke-agent"
    AEGIS_CP_OIDC_CLIENT_SECRET: "rEC99sBBWQAbRgg0xRQFBsMC8rt6pZOB"
    AEGIS_CP_OIDC_AUDIENCE: "aegis-platform"
    AEGIS_CP_OIDC_SKIP_TLS_VERIFY: "false"

    # No custom CA needed - ngrok uses valid public TLS certificates
    AEGIS_PLATFORM_CA_B64: ""
    AEGIS_CP_OIDC_CA_B64: ""

proxy:
  enabled: false
EOF

    log_info "Helm values file generated successfully"
}

# Print usage instructions
print_instructions() {
    echo ""
    echo "=============================================="
    echo -e "${GREEN}ngrok tunnels are now running!${NC}"
    echo "=============================================="
    echo ""
    echo "Tunnel URLs:"
    echo "  Platform API (gRPC): https://${PLATFORM_API_URL}"
    echo "  Keycloak (HTTPS):    https://${KEYCLOAK_URL}"
    echo ""
    echo "ngrok Web Interface: http://localhost:4040"
    echo ""
    echo "To update the spoke agent in your AWS cluster, run:"
    echo ""
    echo "  # First, update kubeconfig for your AWS cluster"
    echo "  aws eks update-kubeconfig --name <cluster-name> --region us-east-1 --profile aegis"
    echo ""
    echo "  # Then upgrade the helm release"
    echo "  helm upgrade aegis-spoke ./charts/aegis-spoke -n aegis-system \\"
    echo "    -f charts/aegis-spoke/values.yaml \\"
    echo "    -f charts/aegis-spoke/values-ngrok.yaml \\"
    echo "    --set k8sAgent.env.AEGIS_CLUSTER_ID=<cluster-id> \\"
    echo "    --set k8sAgent.env.AEGIS_REGION=us-east-1 \\"
    echo "    --set k8sAgent.env.AEGIS_PROVIDER=aws"
    echo ""
    echo "Or use the convenience command:"
    echo "  ./scripts/update-spoke-ngrok.sh <cluster-name>"
    echo ""
    echo "To stop tunnels, run:"
    echo "  ./scripts/stop-ngrok-tunnels.sh"
    echo ""
    echo "=============================================="
}

# Main
main() {
    log_info "Starting ngrok tunnels for aegis-platform development"

    check_prerequisites
    get_local_ports
    create_ngrok_config
    start_port_forwards
    start_ngrok
    get_tunnel_urls
    generate_helm_values
    print_instructions

    log_info "Setup complete! Press Ctrl+C to stop."

    # Keep script running and handle cleanup on exit
    trap cleanup EXIT

    # Wait for user interrupt
    while true; do
        sleep 60
        # Check if processes are still running
        if ! kill -0 $NGROK_PID 2>/dev/null; then
            log_error "ngrok process died unexpectedly"
            exit 1
        fi
    done
}

cleanup() {
    log_info "Cleaning up..."

    # Kill ngrok
    if [ -f /tmp/aegis-ngrok.pid ]; then
        kill $(cat /tmp/aegis-ngrok.pid) 2>/dev/null || true
        rm /tmp/aegis-ngrok.pid
    fi

    # Kill port-forwards
    if [ -f /tmp/aegis-pf-platform.pid ]; then
        kill $(cat /tmp/aegis-pf-platform.pid) 2>/dev/null || true
        rm /tmp/aegis-pf-platform.pid
    fi
    if [ -f /tmp/aegis-pf-keycloak.pid ]; then
        kill $(cat /tmp/aegis-pf-keycloak.pid) 2>/dev/null || true
        rm /tmp/aegis-pf-keycloak.pid
    fi

    pkill -f "ngrok start" 2>/dev/null || true
    pkill -f "kubectl port-forward.*8181:8081" 2>/dev/null || true
    pkill -f "kubectl port-forward.*8543:8443" 2>/dev/null || true

    log_info "Cleanup complete"
}

# Handle arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [--help]"
        echo ""
        echo "Start ngrok tunnels to expose local aegis-platform services"
        echo "for remote AWS spoke cluster connectivity."
        exit 0
        ;;
    *)
        main
        ;;
esac
