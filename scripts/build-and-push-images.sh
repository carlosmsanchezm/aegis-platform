#!/bin/bash
# Build and push all Aegis Docker images to ECR
#
# Usage: ./scripts/build-and-push-images.sh [--profile PROFILE]
#
# Builds 6 images: platform-api, proxy, k8s-agent, workspace-vscode,
# vscode-reh-init, and UI (from aegis-ui repo).
#
# The Iron Bank Dockerfiles use ARG-based base image references.
# For ECR builds (default), public equivalents are used via --build-arg.
# For Iron Bank builds, omit build-args and use registry1.dso.mil auth.

set -e

AWS_PROFILE=${AWS_PROFILE:-aegis-new}
AWS_REGION=${AWS_REGION:-us-east-1}
ECR_REGISTRY="195714074609.dkr.ecr.${AWS_REGION}.amazonaws.com"

# Image tags -- default to git SHA; override via env vars if needed
IMAGE_TAG="${IMAGE_TAG:-$(git rev-parse --short HEAD)}"
PLATFORM_API_TAG="${PLATFORM_API_TAG:-${IMAGE_TAG}}"
PROXY_TAG="${PROXY_TAG:-${IMAGE_TAG}}"
K8S_AGENT_TAG="${K8S_AGENT_TAG:-${IMAGE_TAG}}"
WORKSPACE_TAG="${WORKSPACE_TAG:-${IMAGE_TAG}}"
VSCODE_REH_INIT_TAG="${VSCODE_REH_INIT_TAG:-${IMAGE_TAG}}"
UI_TAG="${UI_TAG:-${IMAGE_TAG}}"
AEGIS_UI_DIR="${AEGIS_UI_DIR:-$HOME/code/aegis-ui}"

# VS Code REH pinned version (update when upgrading VS Code server)
VSCODE_REH_COMMIT="${VSCODE_REH_COMMIT:-ce099c1ed25d9eb3076c11e4a280f3eb52b4fbeb}"

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --profile)
      AWS_PROFILE="$2"
      shift 2
      ;;
    --tag)
      IMAGE_TAG="$2"
      PLATFORM_API_TAG="${IMAGE_TAG}"
      PROXY_TAG="${IMAGE_TAG}"
      K8S_AGENT_TAG="${IMAGE_TAG}"
      WORKSPACE_TAG="${IMAGE_TAG}"
      VSCODE_REH_INIT_TAG="${IMAGE_TAG}"
      UI_TAG="${IMAGE_TAG}"
      shift 2
      ;;
    --skip-ui)
      SKIP_UI=1
      shift
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--profile PROFILE] [--tag TAG] [--skip-ui]"
      exit 1
      ;;
  esac
done

echo "🚀 Building and pushing Aegis images to ECR"
echo "   Registry: ${ECR_REGISTRY}"
echo "   Profile:  ${AWS_PROFILE}"
echo "   Tag:      ${IMAGE_TAG}"
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

# Iron Bank build-args for public base image equivalents
IB_ARGS_GO_DISTROLESS=(
  --build-arg BUILDER_REGISTRY=docker.io
  --build-arg BUILDER_IMAGE=library/golang
  --build-arg BUILDER_TAG=1.24
  --build-arg BASE_REGISTRY=gcr.io
  --build-arg BASE_IMAGE=distroless/static
  --build-arg BASE_TAG=nonroot
)

IB_ARGS_GO_UBI9=(
  --build-arg BUILDER_REGISTRY=docker.io
  --build-arg BUILDER_IMAGE=library/golang
  --build-arg BUILDER_TAG=1.24
  --build-arg BASE_REGISTRY=registry.access.redhat.com
  --build-arg BASE_IMAGE=ubi9/ubi-minimal
  --build-arg BASE_TAG=9.7
)

IB_ARGS_UBI9=(
  --build-arg BASE_REGISTRY=registry.access.redhat.com
  --build-arg BASE_IMAGE=ubi9/ubi-minimal
  --build-arg BASE_TAG=9.7
)

IB_ARGS_CUDA=(
  --build-arg BASE_REGISTRY=docker.io
  --build-arg BASE_IMAGE=nvidia/cuda
  --build-arg BASE_TAG=12.6.1-runtime-ubi9
)

# Step 2: Build and push platform-api
echo "2️⃣  Building platform-api..."
# Stage dependencies for Iron Bank Dockerfile (COPY instead of curl)
echo "   Staging dependencies..."
curl -fsSL "https://dl.k8s.io/release/v1.33.0/bin/linux/amd64/kubectl" -o kubectl
curl -fsSL "https://github.com/jqlang/jq/releases/download/jq-1.8.0/jq-linux-amd64" -o jq
curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-x86_64-2.17.10.zip" -o awscli.zip
curl -fsSL "https://get.pulumi.com/releases/sdk/pulumi-v3.226.0-linux-x64.tar.gz" -o pulumi-linux-x64.tar.gz

docker build --platform linux/amd64 "${IB_ARGS_GO_UBI9[@]}" \
  -f services/platform-api/Dockerfile \
  -t ${ECR_REGISTRY}/aegis/platform-api:${PLATFORM_API_TAG} .
rm -f kubectl jq awscli.zip pulumi-linux-x64.tar.gz
echo "   ✅ Built platform-api:${PLATFORM_API_TAG}"

echo "   Pushing platform-api..."
docker push ${ECR_REGISTRY}/aegis/platform-api:${PLATFORM_API_TAG}
echo "   ✅ Pushed"
echo ""

# Step 3: Build and push proxy
echo "3️⃣  Building proxy..."
docker build --platform linux/amd64 "${IB_ARGS_GO_DISTROLESS[@]}" \
  -f services/proxy/Dockerfile \
  -t ${ECR_REGISTRY}/aegis/proxy:${PROXY_TAG} .
echo "   ✅ Built proxy:${PROXY_TAG}"

echo "   Pushing proxy..."
docker push ${ECR_REGISTRY}/aegis/proxy:${PROXY_TAG}
echo "   ✅ Pushed"
echo ""

# Step 4: Build and push k8s-agent
echo "4️⃣  Building k8s-agent..."
docker build --platform linux/amd64 "${IB_ARGS_GO_DISTROLESS[@]}" \
  -f agents/k8s-agent/Dockerfile \
  -t ${ECR_REGISTRY}/aegis/k8s-agent:${K8S_AGENT_TAG} .
echo "   ✅ Built k8s-agent:${K8S_AGENT_TAG}"

echo "   Pushing k8s-agent..."
docker push ${ECR_REGISTRY}/aegis/k8s-agent:${K8S_AGENT_TAG}
echo "   ✅ Pushed"
echo ""

# Step 5: Build and push workspace
echo "5️⃣  Building workspace..."
echo "   Staging tini..."
curl -fsSL "https://github.com/krallin/tini/releases/download/v0.19.0/tini-amd64" \
  -o workspace-images/openssh-vscode/tini-amd64

docker build --platform linux/amd64 "${IB_ARGS_CUDA[@]}" \
  -f workspace-images/openssh-vscode/Dockerfile \
  -t ${ECR_REGISTRY}/aegis/workspace-vscode:${WORKSPACE_TAG} \
  workspace-images/openssh-vscode/
rm -f workspace-images/openssh-vscode/tini-amd64
echo "   ✅ Built workspace-vscode:${WORKSPACE_TAG}"

echo "   Pushing workspace-vscode..."
docker push ${ECR_REGISTRY}/aegis/workspace-vscode:${WORKSPACE_TAG}
echo "   ✅ Pushed"
echo ""

# Step 6: Build and push vscode-reh-init
echo "6️⃣  Building vscode-reh-init..."
echo "   Downloading VS Code server (commit: ${VSCODE_REH_COMMIT:0:12}...)..."
curl -fsSL "https://vscode.download.prss.microsoft.com/dbazure/download/stable/${VSCODE_REH_COMMIT}/vscode-server-linux-x64.tar.gz" \
  -o workspace-images/vscode-reh-init/vscode-server-linux-x64.tar.gz

docker build --platform linux/amd64 "${IB_ARGS_UBI9[@]}" \
  -f workspace-images/vscode-reh-init/Dockerfile \
  -t ${ECR_REGISTRY}/aegis/vscode-reh-init:${VSCODE_REH_INIT_TAG} \
  workspace-images/vscode-reh-init/
rm -f workspace-images/vscode-reh-init/vscode-server-linux-x64.tar.gz
echo "   ✅ Built vscode-reh-init:${VSCODE_REH_INIT_TAG}"

echo "   Pushing vscode-reh-init..."
docker push ${ECR_REGISTRY}/aegis/vscode-reh-init:${VSCODE_REH_INIT_TAG}
echo "   ✅ Pushed"
echo ""

# Step 7: Build and push UI
if [[ "${SKIP_UI:-0}" == "1" ]]; then
  echo "7️⃣  Skipping UI build (--skip-ui)"
elif [[ ! -d "${AEGIS_UI_DIR}" ]]; then
  echo "7️⃣  ⚠️  AEGIS_UI_DIR not found: ${AEGIS_UI_DIR} — skipping UI build"
else
  echo "7️⃣  Building UI..."
  docker build --platform linux/amd64 -f "${AEGIS_UI_DIR}/packages/backend/Dockerfile.cloud" \
    -t ${ECR_REGISTRY}/aegis/ui:${UI_TAG} "${AEGIS_UI_DIR}"
  echo "   ✅ Built ui:${UI_TAG}"

  echo "   Pushing ui..."
  docker push ${ECR_REGISTRY}/aegis/ui:${UI_TAG}
  echo "   ✅ Pushed"
fi
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 All images built and pushed successfully!"
echo ""
echo "📦 Images in ECR:"
echo "   ${ECR_REGISTRY}/aegis/platform-api:${PLATFORM_API_TAG}"
echo "   ${ECR_REGISTRY}/aegis/proxy:${PROXY_TAG}"
echo "   ${ECR_REGISTRY}/aegis/k8s-agent:${K8S_AGENT_TAG}"
echo "   ${ECR_REGISTRY}/aegis/workspace-vscode:${WORKSPACE_TAG}"
echo "   ${ECR_REGISTRY}/aegis/vscode-reh-init:${VSCODE_REH_INIT_TAG}"
if [[ "${SKIP_UI:-0}" != "1" && -d "${AEGIS_UI_DIR}" ]]; then
  echo "   ${ECR_REGISTRY}/aegis/ui:${UI_TAG}"
fi
echo ""
echo "🚀 Deploy with:"
echo "   cd terraform && \\"
echo "   PLATFORM_API_IMAGE_TAG=${ECR_REGISTRY}/aegis/platform-api:${PLATFORM_API_TAG} \\"
echo "   PROXY_IMAGE_TAG=${ECR_REGISTRY}/aegis/proxy:${PROXY_TAG} \\"
echo "   K8S_AGENT_IMAGE_TAG=${ECR_REGISTRY}/aegis/k8s-agent:${K8S_AGENT_TAG} \\"
echo "   UI_IMAGE_TAG=${ECR_REGISTRY}/aegis/ui:${UI_TAG} \\"
echo "   ./generate-cloud-deployment.sh"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
