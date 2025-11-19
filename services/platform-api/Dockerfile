# Build stage
FROM golang:1.24-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG PULUMI_VERSION=3.131.0

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
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o platform-api .

WORKDIR /tmp
RUN apk add --no-cache curl tar && \
    ARCH="${TARGETARCH:-amd64}" && \
    case "$ARCH" in \
      amd64) PULUMI_ARCH="x64" ;; \
      *) PULUMI_ARCH="$ARCH" ;; \
    esac && \
    curl -fsSL "https://get.pulumi.com/releases/sdk/pulumi-v${PULUMI_VERSION}-linux-${PULUMI_ARCH}.tar.gz" -o pulumi.tar.gz && \
    tar -xzf pulumi.tar.gz && \
    mv pulumi/pulumi /tmp/pulumi-cli && \
    chmod +x /tmp/pulumi-cli && \
    rm -rf pulumi pulumi.tar.gz

# Final stage
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.5

RUN if command -v microdnf >/dev/null 2>&1; then \
      PKG_CMD="microdnf"; \
    elif command -v dnf >/dev/null 2>&1; then \
      PKG_CMD="dnf"; \
    else \
      echo "No supported package manager found (expected microdnf or dnf)" >&2; \
      exit 1; \
    fi; \
    "$PKG_CMD" update -y && \
    "$PKG_CMD" install -y ca-certificates tzdata crypto-policies && \
    "$PKG_CMD" clean all && \
    if command -v update-crypto-policies >/dev/null 2>&1; then update-crypto-policies --set FIPS; fi

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder --chown=1000:1000 /app/services/platform-api/platform-api ./platform-api

# Copy migrations if they exist
COPY --from=builder /app/services/platform-api/migrations ./migrations

# Copy pulumi CLI
COPY --from=builder /tmp/pulumi-cli /usr/local/bin/pulumi

# Expose ports
EXPOSE 8080 8081

# Run the binary
CMD ["./platform-api"]
