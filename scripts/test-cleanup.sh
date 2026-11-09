#!/bin/bash

# Cleanup script for testing environments

echo "🧹 Cleaning up test environments..."

# Kill any local port forwards
pkill -f "kubectl port-forward" || true

# Remove local test namespaces
kubectl delete namespace aegis-services-local --timeout=60s 2>/dev/null || true
kubectl delete namespace aegis-services-local-pg --timeout=60s 2>/dev/null || true

# Remove local Docker containers
docker stop aegis-local-db 2>/dev/null || true
docker rm aegis-local-db 2>/dev/null || true

# Clean up temporary files
rm -f ../charts/aegis-services/values-test-cloud.yaml
rm -f ../charts/aegis-spoke/values-test-cloud.yaml

echo "✅ Test cleanup complete"