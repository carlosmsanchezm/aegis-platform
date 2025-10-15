#!/bin/bash
# Build and push all Aegis Docker images to ECR
#
# Usage: ./scripts/build-and-push-images.sh [--profile PROFILE]

set -e

AWS_PROFILE=${AWS_PROFILE:-myclaude}
AWS_REGION=${AWS_REGION:-us-east-1}
ECR_REGISTRY="567751785679.dkr.ecr.${AWS_REGION}.amazonaws.com"

# Image tags (should match terraform/generate-helm-values.sh)
PLATFORM_API_TAG="v1.0.6-tls2"
PROXY_TAG="no-client-cert"
K8S_AGENT_TAG="v1.0.2-tls-20251005-amd64"

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --profile)
      AWS_PROFILE="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

echo "🚀 Building and pushing Aegis images to ECR"
echo "   Registry: ${ECR_REGISTRY}"
echo "   Profile:  ${AWS_PROFILE}"
echo ""

# Step 1: Login to ECR
echo "1️⃣  Logging in to ECR..."
aws ecr get-login-password --region ${AWS_REGION} --profile ${AWS_PROFILE} | \
  docker login --username AWS --password-stdin ${ECR_REGISTRY}
echo "   ✅ Logged in to ECR"
echo ""

# Get repo root
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

# Step 2: Build and push platform-api
echo "2️⃣  Building platform-api..."
docker build --platform linux/amd64 -f services/platform-api/Dockerfile -t ${ECR_REGISTRY}/aegis/platform-api:${PLATFORM_API_TAG} .
echo "   ✅ Built platform-api:${PLATFORM_API_TAG}"

echo "   Pushing platform-api..."
docker push ${ECR_REGISTRY}/aegis/platform-api:${PLATFORM_API_TAG}
echo "   ✅ Pushed platform-api:${PLATFORM_API_TAG}"
echo ""

# Step 3: Build and push proxy
echo "3️⃣  Building proxy..."
docker build --platform linux/amd64 -f services/proxy/Dockerfile -t ${ECR_REGISTRY}/aegis/proxy:${PROXY_TAG} .
echo "   ✅ Built proxy:${PROXY_TAG}"

echo "   Pushing proxy..."
docker push ${ECR_REGISTRY}/aegis/proxy:${PROXY_TAG}
echo "   ✅ Pushed proxy:${PROXY_TAG}"
echo ""

# Step 4: Build and push k8s-agent
echo "4️⃣  Building k8s-agent..."
docker build --platform linux/amd64 -f agents/k8s-agent/Dockerfile -t ${ECR_REGISTRY}/aegis/k8s-agent:${K8S_AGENT_TAG} .
echo "   ✅ Built k8s-agent:${K8S_AGENT_TAG}"

echo "   Pushing k8s-agent..."
docker push ${ECR_REGISTRY}/aegis/k8s-agent:${K8S_AGENT_TAG}
echo "   ✅ Pushed k8s-agent:${K8S_AGENT_TAG}"
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 All images built and pushed successfully!"
echo ""
echo "📦 Images in ECR:"
echo "   ${ECR_REGISTRY}/aegis/platform-api:${PLATFORM_API_TAG}"
echo "   ${ECR_REGISTRY}/aegis/proxy:${PROXY_TAG}"
echo "   ${ECR_REGISTRY}/aegis/k8s-agent:${K8S_AGENT_TAG}"
echo ""
echo "✅ Verify images:"
echo "   aws ecr list-images --repository-name aegis/platform-api --region ${AWS_REGION} --profile ${AWS_PROFILE}"
echo "   aws ecr list-images --repository-name aegis/proxy --region ${AWS_REGION} --profile ${AWS_PROFILE}"
echo "   aws ecr list-images --repository-name aegis/k8s-agent --region ${AWS_REGION} --profile ${AWS_PROFILE}"
echo ""
echo "🚀 Deploy with:"
echo "   cd terraform && ./generate-helm-values.sh"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
