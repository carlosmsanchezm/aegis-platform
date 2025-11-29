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
RUN go mod download

# Copy the entire project structure
WORKDIR /app
COPY . .

# Build the platform-api service binary with optimized flags
WORKDIR /app/services/platform-api
RUN GOMAXPROCS=1 CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o platform-api .

# Download and install Pulumi CLI
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

# Copy and extract pre-downloaded Pulumi plugins to avoid runtime downloads/rate limits
# Run services/platform-api/pulumi-plugins/download-plugins.sh first to create the bundle
COPY services/platform-api/pulumi-plugins/pulumi-plugins-bundle.tar.gz /tmp/
RUN mkdir -p /root/.pulumi && \
    tar -xzf /tmp/pulumi-plugins-bundle.tar.gz -C /root/.pulumi && \
    rm /tmp/pulumi-plugins-bundle.tar.gz

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
    chmod +x /usr/local/bin/kubectl

WORKDIR /root/

RUN mkdir -p /tmp/pulumi-backend && chmod 0777 /tmp/pulumi-backend

# Copy the binary from builder stage
COPY --from=builder --chown=1000:1000 /app/services/platform-api/platform-api ./platform-api

# Copy migrations if they exist
COPY --from=builder /app/services/platform-api/migrations ./migrations

# Copy Helm charts needed by Pulumi (local chart installs)
COPY --from=builder /app/charts ./charts

# Copy pulumi CLI
COPY --from=builder /tmp/pulumi-cli /usr/local/bin/pulumi
# Copy pulumi plugins to avoid runtime downloads/rate limits
COPY --from=builder /root/.pulumi /root/.pulumi

# Expose ports
EXPOSE 8080 8081

# Run the binary
CMD ["./platform-api"]
