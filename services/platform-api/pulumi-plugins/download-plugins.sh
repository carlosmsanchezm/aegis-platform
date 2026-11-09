#!/bin/bash
# Script to download Pulumi plugins for pre-baking into Docker image
# Run this script when GitHub/Pulumi rate limits have cleared

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Plugin versions (must match Dockerfile and Go code)
AWS_VERSION="5.43.0"
K8S_VERSION="4.24.1"
ARCH="${ARCH:-amd64}"  # Target architecture for Docker image

echo "Downloading Pulumi plugins for linux-${ARCH}..."
echo "AWS plugin: v${AWS_VERSION}"
echo "Kubernetes plugin: v${K8S_VERSION}"
echo ""

# Optional: Set GITHUB_TOKEN if available to increase rate limits
if [ -n "$GITHUB_TOKEN" ]; then
    echo "Using GITHUB_TOKEN for authenticated requests"
    export GITHUB_TOKEN
fi

# Method 1: Try using Pulumi CLI in a Docker container
echo "Attempting to download plugins using Pulumi Docker image..."
if docker run --rm \
    --platform linux/${ARCH} \
    -v "$(pwd):/output" \
    -e GITHUB_TOKEN="${GITHUB_TOKEN:-}" \
    --entrypoint="" \
    pulumi/pulumi:3.153.1 \
    sh -c "
        set -e
        PULUMI_HOME=/tmp/.pulumi
        echo 'Installing AWS plugin v${AWS_VERSION}...'
        pulumi plugin install resource aws ${AWS_VERSION}
        echo 'Installing Kubernetes plugin v${K8S_VERSION}...'
        pulumi plugin install resource kubernetes ${K8S_VERSION}
        echo 'Creating tarball...'
        tar -czf /output/pulumi-plugins-bundle.tar.gz -C /tmp/.pulumi plugins
        echo 'Done!'
    " 2>&1; then
    echo "✓ Docker method succeeded"
else
    echo "✗ Docker method failed, trying alternative method..."
    echo ""

    # Method 2: Try direct download (may work if CDN rate limit is per-method)
    echo "Attempting direct download from Pulumi CDN..."
    mkdir -p .pulumi/plugins

    if curl -fsSL --retry 3 --retry-delay 10 \
        -o ".pulumi/plugins/pulumi-resource-aws-v${AWS_VERSION}-linux-${ARCH}.tar.gz" \
        "https://get.pulumi.com/releases/plugins/pulumi-resource-aws-v${AWS_VERSION}-linux-${ARCH}.tar.gz" && \
       curl -fsSL --retry 3 --retry-delay 10 \
        -o ".pulumi/plugins/pulumi-resource-kubernetes-v${K8S_VERSION}-linux-${ARCH}.tar.gz" \
        "https://get.pulumi.com/releases/plugins/pulumi-resource-kubernetes-v${K8S_VERSION}-linux-${ARCH}.tar.gz"; then

        echo "✓ Direct download succeeded, extracting..."
        mkdir -p .pulumi/plugins/resource-aws-v${AWS_VERSION}
        mkdir -p .pulumi/plugins/resource-kubernetes-v${K8S_VERSION}

        tar -xzf ".pulumi/plugins/pulumi-resource-aws-v${AWS_VERSION}-linux-${ARCH}.tar.gz" \
            -C ".pulumi/plugins/resource-aws-v${AWS_VERSION}"
        tar -xzf ".pulumi/plugins/pulumi-resource-kubernetes-v${K8S_VERSION}-linux-${ARCH}.tar.gz" \
            -C ".pulumi/plugins/resource-kubernetes-v${K8S_VERSION}"

        rm .pulumi/plugins/*.tar.gz
        tar -czf pulumi-plugins-bundle.tar.gz -C .pulumi plugins
        rm -rf .pulumi
    else
        echo "✗ Direct download also failed"
    fi
fi

if [ -f "pulumi-plugins-bundle.tar.gz" ]; then
    echo ""
    echo "✓ Successfully downloaded plugins to pulumi-plugins-bundle.tar.gz"
    echo ""
    echo "Extracting plugins for verification..."
    mkdir -p .pulumi-verify
    tar -xzf pulumi-plugins-bundle.tar.gz -C .pulumi-verify
    echo "Plugins downloaded:"
    ls -lh .pulumi-verify/plugins/
    rm -rf .pulumi-verify
    echo ""
    echo "You can now build the Docker image with: make build-platform"
else
    echo ""
    echo "✗ Failed to download plugins"
    echo ""
    echo "This is likely due to GitHub/Pulumi rate limits."
    echo "You can:"
    echo "  1. Wait ~1 hour and try again"
    echo "  2. Set GITHUB_TOKEN environment variable and retry:"
    echo "     export GITHUB_TOKEN=your_github_token"
    echo "     ./download-plugins.sh"
    echo "  3. Download manually from a different machine/network"
    exit 1
fi
