# =============================================================================
# Aegis K8s Agent — Iron Bank Compliant Dockerfile
# Pattern: Go binary compiled in IB golang builder, deployed on IB UBI9-minimal
# Reference: repo1.dso.mil/dsop/hashicorp/vault-k8s
# =============================================================================

ARG BASE_REGISTRY=registry1.dso.mil
ARG BASE_IMAGE=ironbank/redhat/ubi/ubi9-minimal
ARG BASE_TAG=9.7
ARG BUILDER_REGISTRY=registry1.dso.mil
ARG BUILDER_IMAGE=ironbank/google/golang/ubi9/golang-1.24
ARG BUILDER_TAG=1.24.13

# --- Build Stage ---
FROM ${BUILDER_REGISTRY}/${BUILDER_IMAGE}:${BUILDER_TAG} AS builder

ARG TARGETOS
ARG TARGETARCH
ARG FIPS_ENABLED=true

# Install C compiler for CGO (FIPS/BoringCrypto requires CGO_ENABLED=1).
# IB golang builder already has gcc; this handles public golang:slim images.
RUN if ! command -v gcc >/dev/null 2>&1; then \
      if command -v microdnf >/dev/null 2>&1; then microdnf install -y gcc && microdnf clean all; \
      elif command -v dnf >/dev/null 2>&1; then dnf install -y gcc && dnf clean all; \
      elif command -v apt-get >/dev/null 2>&1; then apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev && rm -rf /var/lib/apt/lists/*; \
      fi; \
    fi

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

RUN CGO_ENABLED=$(if [ "${FIPS_ENABLED}" = "true" ]; then echo 1; else echo 0; fi) \
    GOEXPERIMENT=$(if [ "${FIPS_ENABLED}" = "true" ]; then echo boringcrypto; fi) \
    GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -a -ldflags="-s -w" -o manager cmd/main.go

# --- Final Stage: Iron Bank UBI9-minimal ---
FROM ${BASE_REGISTRY}/${BASE_IMAGE}:${BASE_TAG}

LABEL org.opencontainers.image.title="aegis-k8s-agent" \
      org.opencontainers.image.description="Aegis Platform Kubernetes operator agent for spoke cluster workload reconciliation" \
      org.opencontainers.image.vendor="Aegis Platform" \
      org.opencontainers.image.licenses="Proprietary" \
      mil.dso.ironbank.image.type="commercial" \
      mil.dso.ironbank.product.name="aegis-k8s-agent"

RUN microdnf update -y && \
    microdnf install -y --nodocs \
      ca-certificates \
      shadow-utils && \
    microdnf clean all

# Create non-root user (replaces distroless nonroot user)
RUN groupadd -g 65532 nonroot && \
    useradd -u 65532 -g 65532 -s /sbin/nologin nonroot

WORKDIR /
COPY --from=builder --chown=root:root /workspace/agents/k8s-agent/manager .

RUN find / -xdev -perm /6000 -type f -exec chmod a-s {} \; 2>/dev/null || true

USER 65532:65532

HEALTHCHECK NONE

ENTRYPOINT ["/manager"]
