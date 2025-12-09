#!/bin/bash
# Establish SSH reverse tunnel to AWS relay for aegis-platform development
#
# Prerequisites:
#   1. Deploy the relay infrastructure:
#      cd terraform/pulumi-stack && terraform apply
#
#   2. Generate SSH key if not exists:
#      ssh-keygen -t ed25519 -f ~/.ssh/aegis-relay -N ""
#
#   3. Add the public key to terraform.tfvars:
#      relay_ssh_public_key = "ssh-ed25519 AAAA... your-email"
#
# Usage:
#   ./scripts/start-aws-tunnel.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TERRAFORM_DIR="${PROJECT_ROOT}/terraform/pulumi-stack"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

SSH_KEY="${SSH_KEY:-$HOME/.ssh/aegis-relay}"

# Check prerequisites
check_prerequisites() {
    if [ ! -f "$SSH_KEY" ]; then
        log_error "SSH key not found at $SSH_KEY"
        log_info "Generate one with: ssh-keygen -t ed25519 -f $SSH_KEY -N \"\""
        exit 1
    fi

    if ! command -v terraform &> /dev/null; then
        log_error "terraform is not installed"
        exit 1
    fi
}

# Get relay IP from terraform output
get_relay_ip() {
    log_info "Getting relay IP from Terraform..."

    cd "$TERRAFORM_DIR"

    RELAY_IP=$(terraform output -raw relay_public_ip 2>/dev/null || echo "")

    if [ -z "$RELAY_IP" ]; then
        log_error "Could not get relay IP. Make sure terraform apply has been run."
        log_info "Run: cd terraform/pulumi-stack && terraform apply"
        exit 1
    fi

    NLB_DNS=$(terraform output -raw relay_nlb_dns 2>/dev/null || echo "")

    log_info "Relay Public IP: $RELAY_IP"
    log_info "NLB DNS: $NLB_DNS"
}

# Start port-forwards to local kubernetes services
start_port_forwards() {
    log_info "Starting kubectl port-forwards to local services..."

    # Ensure we're using docker-desktop context
    kubectl config use-context docker-desktop

    # Kill any existing port-forwards on these ports
    pkill -f "kubectl port-forward.*8081:8081" 2>/dev/null || true
    pkill -f "kubectl port-forward.*8443:8443" 2>/dev/null || true
    sleep 1

    # Start platform-api port-forward (gRPC on 8081)
    kubectl port-forward svc/aegis-services-platform-api -n aegis-system 8081:8081 &
    PF_PLATFORM_PID=$!
    echo "$PF_PLATFORM_PID" > /tmp/aegis-aws-pf-platform.pid
    log_info "Platform API port-forward started (localhost:8081)"

    # Start keycloak port-forward (HTTPS on 8443)
    kubectl port-forward svc/aegis-services-keycloak-service -n keycloak 8443:8443 &
    PF_KEYCLOAK_PID=$!
    echo "$PF_KEYCLOAK_PID" > /tmp/aegis-aws-pf-keycloak.pid
    log_info "Keycloak port-forward started (localhost:8443)"

    sleep 2
}

# Establish SSH tunnel
start_ssh_tunnel() {
    log_info "Establishing SSH reverse tunnel to AWS relay..."

    # Kill any existing tunnel
    pkill -f "ssh.*aegis-relay.*$RELAY_IP" 2>/dev/null || true
    sleep 1

    # Establish SSH tunnel with reverse port forwarding
    # -R 0.0.0.0:8081:localhost:8081 - Forward remote 8081 to local 8081 (platform-api gRPC)
    # -R 0.0.0.0:8443:localhost:8443 - Forward remote 8443 to local 8443 (keycloak HTTPS)
    # -N - Don't execute remote command
    # -o ServerAliveInterval=30 - Keep connection alive
    # -o StrictHostKeyChecking=no - Auto-accept host key (dev only!)

    ssh -i "$SSH_KEY" \
        -R 0.0.0.0:8081:localhost:8081 \
        -R 0.0.0.0:8443:localhost:8443 \
        -N \
        -o ServerAliveInterval=30 \
        -o ServerAliveCountMax=3 \
        -o StrictHostKeyChecking=no \
        -o UserKnownHostsFile=/dev/null \
        ec2-user@"$RELAY_IP" &

    SSH_PID=$!
    echo "$SSH_PID" > /tmp/aegis-aws-ssh.pid

    log_info "SSH tunnel established (PID: $SSH_PID)"

    # Wait and verify tunnel is working
    sleep 3
    if ! kill -0 "$SSH_PID" 2>/dev/null; then
        log_error "SSH tunnel failed to start"
        exit 1
    fi
}

# Generate Helm values file
generate_helm_values() {
    VALUES_FILE="${PROJECT_ROOT}/charts/aegis-spoke/values-aws-relay.yaml"

    log_info "Generating Helm values at $VALUES_FILE"

    cat > "$VALUES_FILE" << EOF
# Auto-generated values for AWS relay-based connectivity
# Generated at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
#
# NLB DNS: ${NLB_DNS}
# Platform API: ${NLB_DNS}:8081
# Keycloak: ${NLB_DNS}:8443
#
# Usage:
#   helm upgrade aegis-spoke ./charts/aegis-spoke -n aegis-system \\
#     -f charts/aegis-spoke/values.yaml \\
#     -f charts/aegis-spoke/values-aws-relay.yaml \\
#     --set k8sAgent.env.AEGIS_CLUSTER_ID=<cluster-id> \\
#     --set k8sAgent.env.AEGIS_REGION=us-east-1 \\
#     --set k8sAgent.env.AEGIS_PROVIDER=aws

k8sAgent:
  image:
    pullPolicy: Always
    tag: "dev"
  env:
    # gRPC to platform-api via AWS NLB (TCP passthrough, preserves HTTP/2)
    AEGIS_CP_GRPC: "${NLB_DNS}:8081"
    AEGIS_CP_GRPC_INSECURE: "false"
    AEGIS_CP_GRPC_SKIP_VERIFY: "true"
    AEGIS_CP_GRPC_SERVER_NAME: ""
    AEGIS_FLAVORS: "cpu-small,t4-1gpu"
    AEGIS_DEFAULT_IMAGE: "docker.io/carlosmsanchez/aegis-workspace-vscode:latest"

    # OIDC client credentials (Keycloak via AWS NLB)
    AEGIS_CP_OIDC_TOKEN_URL: "https://${NLB_DNS}:8443/realms/aegis/protocol/openid-connect/token"
    AEGIS_CP_OIDC_CLIENT_ID: "spoke-agent"
    AEGIS_CP_OIDC_CLIENT_SECRET: "rEC99sBBWQAbRgg0xRQFBsMC8rt6pZOB"
    AEGIS_CP_OIDC_AUDIENCE: "aegis-platform"
    AEGIS_CP_OIDC_SKIP_TLS_VERIFY: "true"

    # Skip CA verification (self-signed certs through tunnel)
    AEGIS_PLATFORM_CA_B64: ""
    AEGIS_CP_OIDC_CA_B64: ""

proxy:
  enabled: false
EOF

    log_info "Helm values file generated"
}

# Print instructions
print_instructions() {
    echo ""
    echo "=============================================="
    echo -e "${GREEN}AWS tunnel is now running!${NC}"
    echo "=============================================="
    echo ""
    echo "Endpoints available in AWS VPC:"
    echo "  Platform API (gRPC): ${NLB_DNS}:8081"
    echo "  Keycloak (HTTPS):    ${NLB_DNS}:8443"
    echo ""
    echo "To update the spoke agent in your AWS cluster, run:"
    echo ""
    echo "  ./scripts/update-spoke-aws.sh <cluster-name>"
    echo ""
    echo "Or manually:"
    echo ""
    echo "  aws eks update-kubeconfig --name <cluster-name> --region us-east-1 --profile aegis"
    echo ""
    echo "  helm upgrade aegis-spoke ./charts/aegis-spoke -n aegis-system \\"
    echo "    -f charts/aegis-spoke/values.yaml \\"
    echo "    -f charts/aegis-spoke/values-aws-relay.yaml \\"
    echo "    --set k8sAgent.env.AEGIS_CLUSTER_ID=<cluster-id> \\"
    echo "    --set k8sAgent.env.AEGIS_REGION=us-east-1 \\"
    echo "    --set k8sAgent.env.AEGIS_PROVIDER=aws"
    echo ""
    echo "To stop the tunnel, run:"
    echo "  ./scripts/stop-aws-tunnel.sh"
    echo ""
    echo "=============================================="
}

# Cleanup handler
cleanup() {
    log_info "Cleaning up..."

    # Kill SSH tunnel
    if [ -f /tmp/aegis-aws-ssh.pid ]; then
        kill $(cat /tmp/aegis-aws-ssh.pid) 2>/dev/null || true
        rm /tmp/aegis-aws-ssh.pid
    fi

    # Kill port-forwards
    if [ -f /tmp/aegis-aws-pf-platform.pid ]; then
        kill $(cat /tmp/aegis-aws-pf-platform.pid) 2>/dev/null || true
        rm /tmp/aegis-aws-pf-platform.pid
    fi
    if [ -f /tmp/aegis-aws-pf-keycloak.pid ]; then
        kill $(cat /tmp/aegis-aws-pf-keycloak.pid) 2>/dev/null || true
        rm /tmp/aegis-aws-pf-keycloak.pid
    fi

    log_info "Cleanup complete"
}

# Main
main() {
    log_info "Starting AWS tunnel for aegis-platform development"

    check_prerequisites
    get_relay_ip
    start_port_forwards
    start_ssh_tunnel
    generate_helm_values
    print_instructions

    log_info "Setup complete! Press Ctrl+C to stop."

    trap cleanup EXIT

    # Keep running
    while true; do
        sleep 60
        # Check if SSH tunnel is still running
        if [ -f /tmp/aegis-aws-ssh.pid ] && ! kill -0 $(cat /tmp/aegis-aws-ssh.pid) 2>/dev/null; then
            log_warn "SSH tunnel died, restarting..."
            start_ssh_tunnel
        fi
    done
}

case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [--help]"
        echo ""
        echo "Establish SSH reverse tunnel to AWS relay for aegis-platform development."
        echo ""
        echo "Environment variables:"
        echo "  SSH_KEY - Path to SSH private key (default: ~/.ssh/aegis-relay)"
        exit 0
        ;;
    *)
        main
        ;;
esac
