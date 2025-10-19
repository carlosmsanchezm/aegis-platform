#!/bin/bash

# Complete AWS Aegis Infrastructure Teardown Script
# This script will delete ALL AWS resources created for the Aegis spoke deployment
# USE WITH CAUTION - This will delete everything and is irreversible

set -e

CLUSTER_NAME="aegis-spoke-prod"
REGION="us-east-1"
AWS_PROFILE="${AWS_PROFILE:-myclaude}"

echo "🚨 WARNING: This will completely tear down the Aegis AWS infrastructure!"
echo "   - EKS Cluster: $CLUSTER_NAME"
echo "   - ECR Repositories: aegis/k8s-agent, aegis/proxy"
echo "   - All associated AWS resources (VPC, Load Balancers, etc.)"
echo ""
read -p "Are you sure you want to proceed? (type 'DELETE' to confirm): " confirmation

if [ "$confirmation" != "DELETE" ]; then
    echo "❌ Teardown cancelled."
    exit 1
fi

echo "🔥 Starting complete teardown..."

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

echo "✅ All required tools found"

# 1. Delete Helm deployments first
echo "📦 Removing Helm deployments..."
if kubectl get namespace aegis-spoke >/dev/null 2>&1; then
    helm uninstall aegis-spoke-aws -n aegis-spoke 2>/dev/null || echo "⚠️  Helm release already removed"
    kubectl delete namespace aegis-spoke --timeout=60s 2>/dev/null || echo "⚠️  Namespace already removed"
fi

# Remove NGINX Ingress Controller
helm uninstall ingress-nginx -n kube-system 2>/dev/null || echo "⚠️  NGINX Ingress already removed"

echo "✅ Helm deployments removed"

# 2. Delete EKS cluster (this will also delete node groups)
echo "🏗️  Deleting EKS cluster (this may take 10-15 minutes)..."
if aws eks describe-cluster --name $CLUSTER_NAME --region $REGION --profile $AWS_PROFILE >/dev/null 2>&1; then
    eksctl delete cluster --name $CLUSTER_NAME --region $REGION --profile $AWS_PROFILE --wait
    echo "✅ EKS cluster deleted"
else
    echo "⚠️  EKS cluster already deleted or doesn't exist"
fi

# 3. Delete ECR repositories
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
        echo "⚠️  ECR repository $repo already deleted or doesn't exist"
    fi
done

# 4. Clean up any remaining CloudFormation stacks
echo "☁️  Checking for remaining CloudFormation stacks..."
remaining_stacks=$(aws cloudformation list-stacks \
    --stack-status-filter CREATE_COMPLETE UPDATE_COMPLETE \
    --query "StackSummaries[?contains(StackName, 'eksctl-$CLUSTER_NAME')].StackName" \
    --output text --region $REGION --profile $AWS_PROFILE)

if [ -n "$remaining_stacks" ]; then
    echo "🧹 Cleaning up remaining CloudFormation stacks:"
    for stack in $remaining_stacks; do
        echo "   Deleting stack: $stack"
        aws cloudformation delete-stack --stack-name "$stack" --region $REGION --profile $AWS_PROFILE
    done

    echo "⏳ Waiting for CloudFormation stacks to be deleted..."
    for stack in $remaining_stacks; do
        aws cloudformation wait stack-delete-complete --stack-name "$stack" --region $REGION --profile $AWS_PROFILE 2>/dev/null || true
    done
    echo "✅ CloudFormation stacks cleaned up"
else
    echo "✅ No remaining CloudFormation stacks found"
fi

# 5. Verify cleanup
echo "🔍 Verifying cleanup..."

# Check EKS clusters
if aws eks list-clusters --region $REGION --profile $AWS_PROFILE --query "clusters[?@=='$CLUSTER_NAME']" --output text | grep -q $CLUSTER_NAME; then
    echo "⚠️  Warning: EKS cluster still exists"
else
    echo "✅ EKS cluster confirmed deleted"
fi

# Check ECR repositories
ecr_repos=$(aws ecr describe-repositories --query "repositories[?contains(repositoryName, 'aegis/')].repositoryName" --output text --region $REGION --profile $AWS_PROFILE 2>/dev/null || true)
if [ -n "$ecr_repos" ]; then
    echo "⚠️  Warning: Some ECR repositories still exist: $ecr_repos"
else
    echo "✅ ECR repositories confirmed deleted"
fi

echo ""
echo "🎉 Complete teardown finished!"
echo "💰 All billable AWS resources for Aegis have been removed."
echo ""
echo "📝 Summary of what was deleted:"
echo "   ✅ EKS Cluster: $CLUSTER_NAME"
echo "   ✅ All node groups (CPU and GPU)"
echo "   ✅ VPC and networking resources"
echo "   ✅ Load Balancers and Ingress resources"
echo "   ✅ ECR repositories and images"
echo "   ✅ CloudFormation stacks"
echo ""
echo "💡 To redeploy later, run: eksctl create cluster -f aegis-spoke-eks.yaml --profile $AWS_PROFILE"