#!/bin/bash

# Selective AWS Aegis Infrastructure Teardown Script
# This script provides options to selectively tear down resources to minimize costs
# while optionally keeping ECR repositories for faster redeployment

set -e

CLUSTER_NAME="aegis-spoke-prod"
REGION="us-east-1"
AWS_PROFILE="${AWS_PROFILE:-aegis-new}"

echo "🔧 Selective AWS Aegis Infrastructure Teardown"
echo "This script allows you to selectively remove costly resources while keeping others for faster redeployment."
echo ""

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check required tools
for tool in aws eksctl kubectl; do
    if ! command_exists $tool; then
        echo "❌ Error: $tool is not installed"
        exit 1
    fi
done

# Function to prompt for yes/no
ask_yes_no() {
    local prompt="$1"
    local response
    while true; do
        read -p "$prompt (y/n): " response
        case $response in
            [Yy]* ) return 0;;
            [Nn]* ) return 1;;
            * ) echo "Please answer yes (y) or no (n).";;
        esac
    done
}

echo "💰 Cost Optimization Options:"
echo "   1. Scale down GPU nodes (saves ~$1-2/hour)"
echo "   2. Scale down CPU nodes (saves ~$0.05/hour)"
echo "   3. Delete entire EKS cluster (saves all compute costs)"
echo "   4. Keep ECR repositories (for faster redeployment)"
echo ""

# Option 1: Scale down GPU nodes
if ask_yes_no "🎮 Scale down GPU nodes to 0? (Recommended for cost savings)"; then
    echo "📉 Scaling down GPU node group..."
    if aws eks describe-nodegroup --cluster-name $CLUSTER_NAME --nodegroup-name gpu-workers-g4 --region $REGION --profile $AWS_PROFILE >/dev/null 2>&1; then
        aws eks update-nodegroup-config \
            --cluster-name $CLUSTER_NAME \
            --nodegroup-name gpu-workers-g4 \
            --scaling-config minSize=0,maxSize=2,desiredSize=0 \
            --region $REGION --profile $AWS_PROFILE
        echo "✅ GPU nodes scaled to 0 (will scale up automatically when needed)"
    else
        echo "⚠️  GPU node group not found"
    fi
fi

# Option 2: Scale down CPU nodes
if ask_yes_no "💻 Scale down CPU nodes to 0? (Will affect k8s-agent connectivity)"; then
    echo "📉 Scaling down CPU node group..."
    if aws eks describe-nodegroup --cluster-name $CLUSTER_NAME --nodegroup-name cpu-workers --region $REGION --profile $AWS_PROFILE >/dev/null 2>&1; then
        aws eks update-nodegroup-config \
            --cluster-name $CLUSTER_NAME \
            --nodegroup-name cpu-workers \
            --scaling-config minSize=0,maxSize=1,desiredSize=0 \
            --region $REGION --profile $AWS_PROFILE
        echo "✅ CPU nodes scaled to 0"
        echo "⚠️  Note: k8s-agent will be unavailable until nodes scale back up"
    else
        echo "⚠️  CPU node group not found"
    fi
fi

# Option 3: Delete entire cluster
if ask_yes_no "🏗️  Delete entire EKS cluster? (Maximum cost savings)"; then
    echo "🚨 WARNING: This will delete the entire EKS cluster!"
    if ask_yes_no "Are you absolutely sure?"; then
        # Remove Helm deployments first
        echo "📦 Removing Helm deployments..."
        if kubectl get namespace aegis-spoke >/dev/null 2>&1; then
            helm uninstall aegis-spoke-aws -n aegis-spoke 2>/dev/null || echo "⚠️  Helm release already removed"
            kubectl delete namespace aegis-spoke --timeout=60s 2>/dev/null || echo "⚠️  Namespace already removed"
        fi

        helm uninstall ingress-nginx -n kube-system 2>/dev/null || echo "⚠️  NGINX Ingress already removed"

        echo "🏗️  Deleting EKS cluster..."
        if aws eks describe-cluster --name $CLUSTER_NAME --region $REGION --profile $AWS_PROFILE >/dev/null 2>&1; then
            eksctl delete cluster --name $CLUSTER_NAME --region $REGION --profile $AWS_PROFILE --wait
            echo "✅ EKS cluster deleted"
        else
            echo "⚠️  EKS cluster already deleted"
        fi
    else
        echo "❌ Cluster deletion cancelled"
    fi
fi

# Option 4: Keep ECR repositories
keep_ecr=true
if ask_yes_no "🗂️  Delete ECR repositories? (Will require rebuilding images)"; then
    keep_ecr=false
    echo "🗂️  Deleting ECR repositories..."
    for repo in "aegis/k8s-agent" "aegis/proxy"; do
        if aws ecr describe-repositories --repository-names "$repo" --region $REGION --profile $AWS_PROFILE >/dev/null 2>&1; then
            # Force delete all images first
            aws ecr batch-delete-image \
                --repository-name "$repo" \
                --image-ids "$(aws ecr list-images --repository-name "$repo" --region $REGION --profile $AWS_PROFILE --query 'imageIds[*]' --output json)" \
                --region $REGION --profile $AWS_PROFILE >/dev/null 2>&1 || true

            # Delete the repository
            aws ecr delete-repository --repository-name "$repo" --region $REGION --profile $AWS_PROFILE --force
            echo "✅ Deleted ECR repository: $repo"
        else
            echo "⚠️  ECR repository $repo already deleted"
        fi
    done
fi

echo ""
echo "🎉 Selective teardown completed!"
echo ""
echo "💰 Cost Impact:"
if aws eks describe-cluster --name $CLUSTER_NAME --region $REGION --profile $AWS_PROFILE >/dev/null 2>&1; then
    # Check current node counts
    gpu_desired=$(aws eks describe-nodegroup --cluster-name $CLUSTER_NAME --nodegroup-name gpu-workers-g4 --region $REGION --profile $AWS_PROFILE --query 'nodegroup.scalingConfig.desiredSize' --output text 2>/dev/null || echo "0")
    cpu_desired=$(aws eks describe-nodegroup --cluster-name $CLUSTER_NAME --nodegroup-name cpu-workers --region $REGION --profile $AWS_PROFILE --query 'nodegroup.scalingConfig.desiredSize' --output text 2>/dev/null || echo "0")

    echo "   📊 Current configuration:"
    echo "      - EKS Control Plane: ~$0.10/hour"
    echo "      - CPU nodes ($cpu_desired): ~$(echo "$cpu_desired * 0.05" | bc)/hour"
    echo "      - GPU nodes ($gpu_desired): ~$(echo "$gpu_desired * 1.5" | bc)/hour"

    total_hourly=$(echo "0.10 + $cpu_desired * 0.05 + $gpu_desired * 1.5" | bc)
    echo "      - Total estimated: ~\$${total_hourly}/hour"
else
    echo "   ✅ EKS cluster deleted - no ongoing compute costs"
fi

if [ "$keep_ecr" = true ]; then
    echo "   📦 ECR repositories kept (minimal storage cost ~$0.10/GB/month)"
else
    echo "   📦 ECR repositories deleted"
fi

echo ""
echo "🚀 Quick redeploy commands:"
if aws eks describe-cluster --name $CLUSTER_NAME --region $REGION --profile $AWS_PROFILE >/dev/null 2>&1; then
    echo "   Scale up CPU nodes:  aws eks update-nodegroup-config --cluster-name $CLUSTER_NAME --nodegroup-name cpu-workers --scaling-config desiredSize=1 --region $REGION --profile $AWS_PROFILE"
    echo "   Scale up GPU nodes:  aws eks update-nodegroup-config --cluster-name $CLUSTER_NAME --nodegroup-name gpu-workers-g4 --scaling-config desiredSize=1 --region $REGION --profile $AWS_PROFILE"
    echo "   Redeploy Aegis:      helm install aegis-spoke-aws charts/aegis-spoke -n aegis-spoke -f charts/aegis-spoke/values-cloud.yaml --set-file proxy.tls.cert=/tmp/tls.crt --set-file proxy.tls.key=/tmp/tls.key"
else
    echo "   Recreate cluster:    eksctl create cluster -f aegis-spoke-eks.yaml --profile $AWS_PROFILE"
    echo "   Then redeploy:       ./scripts/deploy-to-aws.sh"
fi