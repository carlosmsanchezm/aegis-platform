# =============================================================================
# Aegis K8s Agent — Iron Bank Compliant Dockerfile
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

ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

COPY proto/aegis/go.mod proto/aegis/go.sum ./proto/aegis/
COPY agents/k8s-agent/go.mod agents/k8s-agent/go.sum ./agents/k8s-agent/
COPY pkg/workspace/go.mod pkg/workspace/go.sum ./pkg/workspace/

WORKDIR /workspace/agents/k8s-agent
RUN go mod download

COPY proto/aegis ../../proto/aegis
COPY pkg/workspace ../../pkg/workspace
COPY agents/k8s-agent/cmd ./cmd
COPY agents/k8s-agent/api ./api
COPY agents/k8s-agent/internal ./internal

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -a -ldflags="-s -w" -o manager cmd/main.go

# --- Final Stage: Iron Bank distroless ---
FROM ${BASE_REGISTRY}/${BASE_IMAGE}:${BASE_TAG}

LABEL org.opencontainers.image.title="aegis-k8s-agent" \
      org.opencontainers.image.description="Aegis Platform Kubernetes operator agent for spoke cluster workload reconciliation" \
      org.opencontainers.image.vendor="Aegis Platform" \
      org.opencontainers.image.licenses="Proprietary" \
      mil.dso.ironbank.image.type="opensource" \
      mil.dso.ironbank.product.name="aegis-k8s-agent"

WORKDIR /
COPY --from=builder --chown=root:root /workspace/agents/k8s-agent/manager .

USER 65532:65532

HEALTHCHECK NONE

ENTRYPOINT ["/manager"]
