#!/bin/bash
# spoke-connectivity.sh - Ensures all spoke cluster connectivity requirements are met
#
# This script:
# 1. Validates NLB endpoint configuration is correct
# 2. Starts SSH tunnel if not running
# 3. Starts port-forward if not running
# 4. Validates spoke agents can reach platform-api
#
# Usage:
#   ./scripts/spoke-connectivity.sh          # Check and start everything
#   ./scripts/spoke-connectivity.sh check    # Only check, don't start anything
#   ./scripts/spoke-connectivity.sh stop     # Stop tunnel and port-forward

set -e

# Configuration
RELAY_HOST="67.202.30.169"
RELAY_USER="ec2-user"
RELAY_KEY="$HOME/.ssh/aegis-relay"
NLB_NAME="aegis-dev-relay-nlb"
AWS_REGION="us-east-1"
GRPC_PORT="8081"
KEYCLOAK_PORT="8443"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Get the actual NLB DNS name from AWS
get_actual_nlb_dns() {
    aws elbv2 describe-load-balancers \
        --names "$NLB_NAME" \
        --region "$AWS_REGION" \
        --query 'LoadBalancers[0].DNSName' \
        --output text 2>/dev/null
}

# Get the configured endpoint from platform-api
get_configured_endpoint() {
    kubectl get deploy -n aegis-system aegis-services-platform-api \
        -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="AEGIS_PLATFORM_API_ENDPOINT")].value}' 2>/dev/null
}

# Get the endpoint from values file
get_values_file_endpoint() {
    local values_file="$1"
    if [[ -f "$values_file" ]]; then
        grep "AEGIS_CP_GRPC:" "$values_file" 2>/dev/null | head -1 | sed 's/.*"\(.*\)".*/\1/' | cut -d: -f1
    fi
}

# Check if SSH tunnel is running
is_ssh_tunnel_running() {
    pgrep -f "ssh.*$RELAY_HOST.*8081" > /dev/null 2>&1
}

# Check if port-forward is running
is_port_forward_running() {
    lsof -i ":$GRPC_PORT" 2>/dev/null | grep -q kubectl
}

# Start SSH tunnel
start_ssh_tunnel() {
    if [[ ! -f "$RELAY_KEY" ]]; then
        log_error "SSH key not found: $RELAY_KEY"
        log_error "Please ensure the aegis-relay SSH key is in place"
        return 1
    fi

    log_info "Starting SSH tunnel to $RELAY_HOST..."
    ssh -i "$RELAY_KEY" \
        -R "0.0.0.0:$GRPC_PORT:localhost:$GRPC_PORT" \
        -R "0.0.0.0:$KEYCLOAK_PORT:localhost:$KEYCLOAK_PORT" \
        -N \
        -o ServerAliveInterval=30 \
        -o ServerAliveCountMax=3 \
        -o StrictHostKeyChecking=no \
        -o UserKnownHostsFile=/dev/null \
        -o ExitOnForwardFailure=yes \
        "$RELAY_USER@$RELAY_HOST" &

    local tunnel_pid=$!
    echo "$tunnel_pid" > /tmp/aegis-ssh-tunnel.pid
    sleep 2

    if kill -0 "$tunnel_pid" 2>/dev/null; then
        log_info "SSH tunnel started (PID: $tunnel_pid)"
        return 0
    else
        log_error "SSH tunnel failed to start"
        return 1
    fi
}

# Start port-forward
start_port_forward() {
    log_info "Starting port-forward for platform-api..."
    kubectl port-forward -n aegis-system svc/aegis-services-platform-api "$GRPC_PORT:$GRPC_PORT" --address 0.0.0.0 &

    local pf_pid=$!
    echo "$pf_pid" > /tmp/aegis-port-forward.pid
    sleep 2

    if kill -0 "$pf_pid" 2>/dev/null; then
        log_info "Port-forward started (PID: $pf_pid)"
        return 0
    else
        log_error "Port-forward failed to start"
        return 1
    fi
}

# Stop everything
stop_all() {
    log_info "Stopping spoke connectivity services..."

    # Kill SSH tunnel
    if [[ -f /tmp/aegis-ssh-tunnel.pid ]]; then
        local pid=$(cat /tmp/aegis-ssh-tunnel.pid)
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null
            log_info "Stopped SSH tunnel (PID: $pid)"
        fi
        rm -f /tmp/aegis-ssh-tunnel.pid
    fi

    # Also kill by pattern
    pkill -f "ssh.*$RELAY_HOST.*8081" 2>/dev/null || true

    # Kill port-forward
    if [[ -f /tmp/aegis-port-forward.pid ]]; then
        local pid=$(cat /tmp/aegis-port-forward.pid)
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null
            log_info "Stopped port-forward (PID: $pid)"
        fi
        rm -f /tmp/aegis-port-forward.pid
    fi

    # Also kill by pattern
    pkill -f "kubectl.*port-forward.*platform-api.*8081" 2>/dev/null || true

    log_info "All spoke connectivity services stopped"
}

# Validate NLB endpoint configuration
validate_nlb_endpoint() {
    log_info "Validating NLB endpoint configuration..."

    local actual_dns=$(get_actual_nlb_dns)
    if [[ -z "$actual_dns" ]]; then
        log_error "Could not fetch NLB DNS from AWS. Check AWS credentials."
        return 1
    fi

    local configured=$(get_configured_endpoint)
    local configured_host=$(echo "$configured" | cut -d: -f1)

    if [[ "$configured_host" != "$actual_dns" ]]; then
        log_error "NLB endpoint MISMATCH detected!"
        log_error "  Actual NLB DNS:    $actual_dns"
        log_error "  Configured in k8s: $configured_host"
        log_warn ""
        log_warn "Fix with:"
        log_warn "  kubectl set env deployment/aegis-services-platform-api -n aegis-system \\"
        log_warn "    AEGIS_PLATFORM_API_ENDPOINT=${actual_dns}:${GRPC_PORT}"
        return 1
    fi

    log_info "NLB endpoint configuration is correct: $actual_dns"

    # Also check values file
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local repo_root="$(dirname "$script_dir")"
    local values_file="$repo_root/charts/aegis-spoke/values-aws-relay.yaml"

    if [[ -f "$values_file" ]]; then
        local values_dns=$(get_values_file_endpoint "$values_file")
        if [[ -n "$values_dns" && "$values_dns" != "$actual_dns" ]]; then
            log_warn "Values file has outdated NLB DNS:"
            log_warn "  File: $values_file"
            log_warn "  Has:  $values_dns"
            log_warn "  Should be: $actual_dns"
            log_warn ""
            log_warn "Note: This is usually OK because inline values override the file during Pulumi deployment."
        fi
    fi

    return 0
}

# Check NLB target health
check_nlb_health() {
    log_info "Checking NLB target health..."

    local target_group_arn=$(aws elbv2 describe-target-groups \
        --names aegis-dev-relay-grpc \
        --region "$AWS_REGION" \
        --query 'TargetGroups[0].TargetGroupArn' \
        --output text 2>/dev/null)

    if [[ -z "$target_group_arn" || "$target_group_arn" == "None" ]]; then
        log_warn "Could not find target group. Skipping health check."
        return 0
    fi

    local health=$(aws elbv2 describe-target-health \
        --target-group-arn "$target_group_arn" \
        --region "$AWS_REGION" \
        --query 'TargetHealthDescriptions[0].TargetHealth.State' \
        --output text 2>/dev/null)

    if [[ "$health" == "healthy" ]]; then
        log_info "NLB target is healthy"
        return 0
    else
        log_error "NLB target is NOT healthy: $health"
        return 1
    fi
}

# Main status check
check_status() {
    local all_ok=true

    echo ""
    echo "=========================================="
    echo "  Aegis Spoke Connectivity Status"
    echo "=========================================="
    echo ""

    # Check SSH tunnel
    echo -n "SSH Tunnel:      "
    if is_ssh_tunnel_running; then
        echo -e "${GREEN}RUNNING${NC}"
    else
        echo -e "${RED}NOT RUNNING${NC}"
        all_ok=false
    fi

    # Check port-forward
    echo -n "Port-forward:    "
    if is_port_forward_running; then
        echo -e "${GREEN}RUNNING${NC}"
    else
        echo -e "${RED}NOT RUNNING${NC}"
        all_ok=false
    fi

    # Check NLB endpoint
    echo -n "NLB Endpoint:    "
    if validate_nlb_endpoint > /dev/null 2>&1; then
        echo -e "${GREEN}CONFIGURED CORRECTLY${NC}"
    else
        echo -e "${RED}MISCONFIGURED${NC}"
        all_ok=false
    fi

    # Check NLB health
    echo -n "NLB Health:      "
    if check_nlb_health > /dev/null 2>&1; then
        echo -e "${GREEN}HEALTHY${NC}"
    else
        echo -e "${YELLOW}UNKNOWN/UNHEALTHY${NC}"
    fi

    # Check for registered clusters
    echo -n "Clusters in DB:  "
    local cluster_count=$(kubectl exec -n aegis-system platform-postgres-0 -- \
        psql -U aegis_platform -d aegis_platform -t -c "SELECT COUNT(*) FROM clusters;" 2>/dev/null | tr -d ' ')
    if [[ "$cluster_count" -gt 0 ]]; then
        echo -e "${GREEN}$cluster_count registered${NC}"
    else
        echo -e "${YELLOW}None${NC}"
    fi

    # Check AegisCluster CRDs
    echo -n "AegisCluster CRDs: "
    local crd_count=$(kubectl get aegisclusters -A --no-headers 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$crd_count" -gt 0 ]]; then
        echo -e "${GREEN}$crd_count found${NC}"
    else
        echo -e "${YELLOW}None${NC}"
    fi

    echo ""

    if $all_ok; then
        log_info "All connectivity requirements are met!"
        return 0
    else
        log_warn "Some connectivity requirements are not met."
        return 1
    fi
}

# Ensure everything is running
ensure_connectivity() {
    local needs_restart=false

    # Validate NLB endpoint first
    if ! validate_nlb_endpoint; then
        log_error "Please fix the NLB endpoint configuration first."
        exit 1
    fi

    # Check and start SSH tunnel
    if ! is_ssh_tunnel_running; then
        log_warn "SSH tunnel not running"
        start_ssh_tunnel || exit 1
        needs_restart=true
    else
        log_info "SSH tunnel is already running"
    fi

    # Check and start port-forward
    if ! is_port_forward_running; then
        log_warn "Port-forward not running"
        start_port_forward || exit 1
        needs_restart=true
    else
        log_info "Port-forward is already running"
    fi

    # Final status
    echo ""
    check_status
}

# Fix NLB endpoint automatically
fix_nlb_endpoint() {
    log_info "Auto-fixing NLB endpoint configuration..."

    local actual_dns=$(get_actual_nlb_dns)
    if [[ -z "$actual_dns" ]]; then
        log_error "Could not fetch NLB DNS from AWS"
        return 1
    fi

    log_info "Setting AEGIS_PLATFORM_API_ENDPOINT to ${actual_dns}:${GRPC_PORT}"
    kubectl set env deployment/aegis-services-platform-api -n aegis-system \
        "AEGIS_PLATFORM_API_ENDPOINT=${actual_dns}:${GRPC_PORT}"

    log_info "Waiting for platform-api to restart..."
    kubectl rollout status deployment/aegis-services-platform-api -n aegis-system --timeout=120s

    log_info "NLB endpoint fixed!"
}

# Show help
show_help() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  (none)    Check and ensure all connectivity requirements are met"
    echo "  check     Only check status, don't start anything"
    echo "  stop      Stop SSH tunnel and port-forward"
    echo "  fix-nlb   Auto-fix NLB endpoint configuration"
    echo "  help      Show this help message"
    echo ""
    echo "This script manages the connectivity between AWS EKS spoke clusters"
    echo "and your local development hub (platform-api in kind cluster)."
    echo ""
    echo "Components managed:"
    echo "  - SSH reverse tunnel to aegis-dev-relay EC2 instance"
    echo "  - kubectl port-forward from localhost to platform-api"
    echo "  - NLB endpoint configuration validation"
}

# Main
case "${1:-}" in
    check)
        check_status
        ;;
    stop)
        stop_all
        ;;
    fix-nlb)
        fix_nlb_endpoint
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        ensure_connectivity
        ;;
esac
