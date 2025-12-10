#!/bin/bash
# Update the aegis-spoke deployment in an AWS EKS cluster to use AWS relay endpoints
# Usage: ./scripts/update-spoke-aws.sh <cluster-name> [region]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Arguments
CLUSTER_NAME="${1:-}"
REGION="${2:-us-east-1}"
AWS_PROFILE="${AWS_PROFILE:-aegis}"

if [ -z "$CLUSTER_NAME" ]; then
    log_error "Usage: $0 <cluster-name> [region]"
    log_error "Example: $0 db-1-us-east-1-atlas-train-govcloud us-east-1"
    exit 1
fi

# Check if AWS relay values file exists
VALUES_AWS="${PROJECT_ROOT}/charts/aegis-spoke/values-aws-relay.yaml"
if [ ! -f "$VALUES_AWS" ]; then
    log_error "AWS relay values file not found at $VALUES_AWS"
    log_error "Run ./scripts/start-aws-tunnel.sh first to generate it"
    exit 1
fi

# Extract cluster ID from the cluster name
CLUSTER_ID="${CLUSTER_NAME}"

log_info "Updating spoke agent in cluster: $CLUSTER_NAME"
log_info "Region: $REGION"
log_info "AWS Profile: $AWS_PROFILE"
log_info "Cluster ID: $CLUSTER_ID"

# Update kubeconfig
log_info "Updating kubeconfig for EKS cluster..."
aws eks update-kubeconfig \
    --name "$CLUSTER_NAME" \
    --region "$REGION" \
    --profile "$AWS_PROFILE" \
    --alias "eks-$CLUSTER_NAME" \
    --role-arn "arn:aws:iam::567751785679:role/aegis-platform"

# Use the new context
kubectl config use-context "eks-$CLUSTER_NAME"

# Verify connection
log_info "Verifying cluster connection..."
kubectl get nodes --no-headers | head -1

# Extract project ID from cluster ID (format: {projectId}-{region}-{clusterName})
# e.g., "db-1-us-east-1-atlas-train-govcloud" -> "db-1"
PROJECT_ID=""
for pattern in "-us-east-" "-us-west-" "-eu-west-" "-eu-central-" "-ap-" "-sa-east-" "-ca-central-" "-me-south-" "-af-south-"; do
    if [[ "$CLUSTER_ID" == *"$pattern"* ]]; then
        PROJECT_ID="${CLUSTER_ID%%$pattern*}"
        break
    fi
done
if [ -z "$PROJECT_ID" ]; then
    # Fallback: use first part before hyphen
    PROJECT_ID="${CLUSTER_ID%%-*}"
fi
log_info "Project ID: $PROJECT_ID"

# Upgrade helm release
log_info "Upgrading aegis-spoke helm release..."
helm upgrade aegis-spoke "${PROJECT_ROOT}/charts/aegis-spoke" \
    -n aegis-system \
    -f "${PROJECT_ROOT}/charts/aegis-spoke/values.yaml" \
    -f "$VALUES_AWS" \
    --set k8sAgent.env.AEGIS_CLUSTER_ID="$CLUSTER_ID" \
    --set k8sAgent.env.AEGIS_REGION="$REGION" \
    --set k8sAgent.env.AEGIS_PROVIDER=aws \
    --set k8sAgent.env.AEGIS_PROJECT_ID="$PROJECT_ID" \
    --wait \
    --timeout 5m

# Restart the deployment to pick up new environment variables
log_info "Restarting aegis-spoke-k8s-agent deployment..."
kubectl rollout restart deployment/aegis-spoke-k8s-agent -n aegis-system
kubectl rollout status deployment/aegis-spoke-k8s-agent -n aegis-system --timeout=120s

# Show logs
log_info "Showing recent logs from spoke agent..."
sleep 5
kubectl logs deployment/aegis-spoke-k8s-agent -n aegis-system --tail=20

log_info "Spoke agent updated successfully!"
log_info ""
log_info "To monitor logs:"
log_info "  kubectl logs -f deployment/aegis-spoke-k8s-agent -n aegis-system"
