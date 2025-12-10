#!/bin/bash
# Stop ngrok tunnels and port-forwards

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }

log_info "Stopping ngrok tunnels and port-forwards..."

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

# Kill any remaining processes
pkill -f "ngrok start" 2>/dev/null || true
pkill -f "kubectl port-forward.*8181:8081" 2>/dev/null || true
pkill -f "kubectl port-forward.*8543:8443" 2>/dev/null || true

log_info "All tunnels stopped."
