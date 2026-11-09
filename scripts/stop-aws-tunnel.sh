#!/bin/bash
# Stop AWS SSH tunnel and port-forwards

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }

log_info "Stopping AWS tunnel and port-forwards..."

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

# Kill any remaining processes
pkill -f "ssh.*aegis-relay" 2>/dev/null || true
pkill -f "kubectl port-forward.*8081:8081" 2>/dev/null || true
pkill -f "kubectl port-forward.*8443:8443" 2>/dev/null || true

log_info "All tunnels stopped."
