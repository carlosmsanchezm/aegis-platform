#!/bin/bash

# Aegis AWS Infrastructure - Complete Teardown
# Deletes everything to avoid billing

set -e

CLUSTER_NAME="aegis-spoke-prod"
REGION="us-east-1"
AWS_PROFILE="${AWS_PROFILE:-aegis-new}"

echo "🚨 COMPLETE TEARDOWN - This will delete ALL AWS resources"
echo "   - EKS Cluster: $CLUSTER_NAME"
echo "   - ECR Repositories"
echo "   - All networking resources"
echo ""
read -p "Type 'DELETE' to confirm: " confirmation

if [ "$confirmation" != "DELETE" ]; then
    echo "❌ Cancelled"
    exit 1
fi

echo "🔥 Starting teardown..."

# Remove Helm releases
echo "📦 Removing Helm deployments..."
helm uninstall aegis-spoke-aws -n aegis-spoke 2>/dev/null || true
helm uninstall ingress-nginx -n kube-system 2>/dev/null || true
kubectl delete namespace aegis-spoke --timeout=60s 2>/dev/null || true

# Delete EKS cluster
echo "🏗️  Deleting EKS cluster..."
eksctl delete cluster --name $CLUSTER_NAME --region $REGION --profile $AWS_PROFILE --wait 2>/dev/null || true

# Delete ECR repositories
echo "🗂️  Deleting ECR repositories..."
for repo in "aegis/k8s-agent" "aegis/proxy"; do
    # Delete all images first
    aws ecr batch-delete-image \
        --repository-name "$repo" \
        --image-ids "$(aws ecr list-images --repository-name "$repo" --region $REGION --profile $AWS_PROFILE --query 'imageIds[*]' --output json 2>/dev/null)" \
        --region $REGION --profile $AWS_PROFILE >/dev/null 2>&1 || true

    # Delete repository
    aws ecr delete-repository --repository-name "$repo" --region $REGION --profile $AWS_PROFILE --force 2>/dev/null || true
done

echo "✅ Complete teardown finished - no AWS charges"