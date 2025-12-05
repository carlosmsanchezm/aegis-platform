# syntax=docker/dockerfile:1.7
# Build stage
FROM golang:1.24-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG PULUMI_VERSION=3.153.1

WORKDIR /app

# Copy go.mod and go.sum first for better layer caching
COPY go.work* ./
COPY proto/aegis/go.* ./proto/aegis/
COPY services/platform-api/go.* ./services/platform-api/
COPY services/proxy/go.* ./services/proxy/
COPY agents/k8s-agent/go.* ./agents/k8s-agent/
COPY pkg/workspace/go.* ./pkg/workspace/

# Download dependencies (cached layer)
WORKDIR /app/services/platform-api
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    GOMODCACHE=/go/pkg/mod \
    go mod download

# Download and install Pulumi CLI early so it stays cached across code changes
WORKDIR /tmp
RUN apk add --no-cache curl tar && \
    ARCH="${TARGETARCH:-amd64}" && \
    case "$ARCH" in \
      amd64) PULUMI_ARCH="x64" ;; \
      *) PULUMI_ARCH="$ARCH" ;; \
    esac && \
    curl --retry 5 --retry-delay 5 -fsSL "https://get.pulumi.com/releases/sdk/pulumi-v${PULUMI_VERSION}-linux-${PULUMI_ARCH}.tar.gz" -o pulumi.tar.gz && \
    tar -xzf pulumi.tar.gz && \
    mv pulumi/pulumi /tmp/pulumi-cli && \
    chmod +x /tmp/pulumi-cli && \
    rm -rf pulumi pulumi.tar.gz

# Optionally copy and extract pre-downloaded Pulumi plugins to avoid runtime downloads/rate limits.
# If the bundle is missing, skip and fall back to on-demand plugin download.
RUN mkdir -p /root/.pulumi
COPY services/platform-api/pulumi-plugins/pulumi-plugins-bundle.tar.gz /tmp/pulumi-plugins-bundle.tar.gz 2>/dev/null || true
RUN if [ -f /tmp/pulumi-plugins-bundle.tar.gz ]; then \
      tar -xzf /tmp/pulumi-plugins-bundle.tar.gz -C /root/.pulumi && rm /tmp/pulumi-plugins-bundle.tar.gz; \
    else \
      echo "Pulumi plugin bundle not found; plugins will be fetched at runtime"; \
    fi

# Copy the entire project structure
WORKDIR /app
COPY --link . .

# Build the platform-api service binary with optimized flags
WORKDIR /app/services/platform-api
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    GOMODCACHE=/go/pkg/mod \
    GOCACHE=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o platform-api .

# Final stage
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.5

ARG TARGETARCH
ARG AWSCLI_VERSION=2.17.10
ARG KUBECTL_VERSION=1.30.4

RUN if command -v microdnf >/dev/null 2>&1; then \
      PKG_CMD="microdnf"; \
    elif command -v dnf >/dev/null 2>&1; then \
      PKG_CMD="dnf"; \
    else \
      echo "No supported package manager found (expected microdnf or dnf)" >&2; \
      exit 1; \
    fi; \
    "$PKG_CMD" update -y && \
    "$PKG_CMD" install -y ca-certificates tzdata crypto-policies tar gzip unzip && \
    "$PKG_CMD" clean all && \
    if command -v update-crypto-policies >/dev/null 2>&1; then update-crypto-policies --set FIPS; fi

RUN set -euo pipefail; \
    ARCH="${TARGETARCH:-$(uname -m)}"; \
    case "$ARCH" in \
      amd64|x86_64) AWS_ARCH="x86_64"; KUBE_ARCH="amd64" ;; \
      arm64|aarch64) AWS_ARCH="aarch64"; KUBE_ARCH="arm64" ;; \
      *) echo "Unsupported architecture: ${ARCH}" >&2; exit 1 ;; \
    esac; \
    curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-${AWS_ARCH}-${AWSCLI_VERSION}.zip" -o /tmp/awscliv2.zip && \
    unzip -q /tmp/awscliv2.zip -d /tmp && \
    /tmp/aws/install --bin-dir /usr/local/bin --install-dir /usr/local/aws-cli && \
    rm -rf /tmp/aws /tmp/awscliv2.zip && \
    curl -fsSLo /usr/local/bin/kubectl "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/${KUBE_ARCH}/kubectl" && \
    chmod +x /usr/local/bin/kubectl && \
    curl -fsSLo /usr/local/bin/jq https://github.com/jqlang/jq/releases/download/jq-1.7/jq-linux-${KUBE_ARCH} && \
    chmod +x /usr/local/bin/jq

WORKDIR /root/

RUN mkdir -p /tmp/pulumi-backend && chmod 0777 /tmp/pulumi-backend

# Copy the binary from builder stage
COPY --from=builder --chown=1000:1000 /app/services/platform-api/platform-api ./platform-api

# Copy migrations if they exist
COPY --from=builder /app/services/platform-api/migrations ./migrations

# Copy Helm charts needed by Pulumi (local chart installs)
COPY --from=builder /app/charts ./charts

# Copy observability values used by the Pulumi installer
COPY --from=builder /app/services/platform-api/config/observability ./services/platform-api/config/observability

# Copy pulumi CLI
COPY --from=builder /tmp/pulumi-cli /usr/local/bin/pulumi
# Copy pulumi plugins to avoid runtime downloads/rate limits
COPY --from=builder /root/.pulumi /root/.pulumi

# Helper to assume role with external ID and fetch EKS exec credential
RUN cat > /usr/local/bin/aegis-eks-token <<'EOF' && chmod +x /usr/local/bin/aegis-eks-token
#!/usr/bin/env bash
set -euo pipefail

: "${AEGIS_EKS_CLUSTER_NAME:?cluster name required}"
: "${AEGIS_EKS_REGION:?region required}"
: "${AEGIS_EKS_ROLE_ARN:?role arn required}"
: "${AEGIS_EKS_EXTERNAL_ID:?external id required}"

assumed=$(aws sts assume-role \
  --role-arn "${AEGIS_EKS_ROLE_ARN}" \
  --role-session-name "aegis-eks-token" \
  --external-id "${AEGIS_EKS_EXTERNAL_ID}")

export AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY
export AWS_SESSION_TOKEN

AWS_ACCESS_KEY_ID=$(echo "${assumed}" | jq -r .Credentials.AccessKeyId)
AWS_SECRET_ACCESS_KEY=$(echo "${assumed}" | jq -r .Credentials.SecretAccessKey)
AWS_SESSION_TOKEN=$(echo "${assumed}" | jq -r .Credentials.SessionToken)

aws eks get-token \
  --cluster-name "${AEGIS_EKS_CLUSTER_NAME}" \
  --region "${AEGIS_EKS_REGION}" \
  --output json
EOF

# Expose ports
EXPOSE 8080 8081

# Run the binary
CMD ["./platform-api"]
