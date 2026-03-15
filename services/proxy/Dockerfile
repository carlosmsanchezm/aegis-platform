# =============================================================================
# Aegis Proxy — Iron Bank Compliant Dockerfile
# Pattern: Go binary compiled in IB golang builder, deployed on IB distroless
# Reference: repo1.dso.mil/dsop/hashicorp/vault-csi-provider
# =============================================================================

ARG BASE_REGISTRY=registry1.dso.mil
ARG BASE_IMAGE=ironbank/google/distroless/static
ARG BASE_TAG=nonroot
ARG BUILDER_REGISTRY=registry1.dso.mil
ARG BUILDER_IMAGE=ironbank/google/golang/ubi9/golang-1.24
ARG BUILDER_TAG=1.24.13

# --- Build Stage ---
FROM ${BUILDER_REGISTRY}/${BUILDER_IMAGE}:${BUILDER_TAG} AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /app

COPY go.work go.work.sum ./
COPY services/proxy/go.mod services/proxy/go.sum ./services/proxy/
COPY services/platform-api/go.mod services/platform-api/go.sum ./services/platform-api/
COPY agents/k8s-agent/go.mod agents/k8s-agent/go.sum ./agents/k8s-agent/
COPY proto/aegis/go.mod proto/aegis/go.sum ./proto/aegis/
COPY pkg/workspace/go.mod pkg/workspace/go.sum ./pkg/workspace/

RUN go work sync

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /aegis-auth-proxy ./services/proxy/cmd/aegis-auth-proxy

# --- Final Stage: Iron Bank distroless ---
FROM ${BASE_REGISTRY}/${BASE_IMAGE}:${BASE_TAG}

LABEL org.opencontainers.image.title="aegis-proxy" \
      org.opencontainers.image.description="Aegis Platform reverse proxy for secure WebSocket workspace tunneling" \
      org.opencontainers.image.vendor="Aegis Platform" \
      org.opencontainers.image.licenses="Proprietary" \
      mil.dso.ironbank.image.type="opensource" \
      mil.dso.ironbank.product.name="aegis-proxy"

COPY --from=builder --chown=root:root /aegis-auth-proxy /aegis-auth-proxy

EXPOSE 8080

USER nonroot:nonroot

HEALTHCHECK NONE

ENTRYPOINT ["/aegis-auth-proxy"]
