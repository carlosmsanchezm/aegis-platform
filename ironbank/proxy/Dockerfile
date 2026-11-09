# =============================================================================
# Aegis Proxy — Iron Bank Dockerfile (Tarball Pattern)
# Source code is pre-fetched by the Iron Bank pipeline as a tarball declared
# in hardening_manifest.yaml. No source code lives in this repository.
# =============================================================================

ARG BASE_REGISTRY=registry1.dso.mil
ARG BASE_IMAGE=ironbank/redhat/ubi/ubi9-minimal
ARG BASE_TAG=9.7
ARG BUILDER_REGISTRY=registry1.dso.mil
ARG BUILDER_IMAGE=ironbank/google/golang/ubi9/golang-1.24
ARG BUILDER_TAG=1.24.13

# --- Build Stage ---
FROM ${BUILDER_REGISTRY}/${BUILDER_IMAGE}:${BUILDER_TAG} AS builder

ARG TARGETOS=linux
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

WORKDIR /app

# Extract monorepo source tarball (pre-fetched by IB pipeline from hardening_manifest)
COPY aegis-platform-src.tar.gz /tmp/
RUN tar -xzf /tmp/aegis-platform-src.tar.gz --strip-components=1 -C /app && \
    rm /tmp/aegis-platform-src.tar.gz

# Build the proxy binary
RUN CGO_ENABLED=$(if [ "${FIPS_ENABLED}" = "true" ]; then echo 1; else echo 0; fi) \
    GOEXPERIMENT=$(if [ "${FIPS_ENABLED}" = "true" ]; then echo boringcrypto; fi) \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o /aegis-auth-proxy ./services/proxy/cmd/aegis-auth-proxy

# --- Final Stage: Iron Bank UBI9-minimal ---
FROM ${BASE_REGISTRY}/${BASE_IMAGE}:${BASE_TAG}

LABEL org.opencontainers.image.title="aegis-proxy" \
      org.opencontainers.image.description="Aegis Platform reverse proxy for secure WebSocket workspace tunneling" \
      org.opencontainers.image.vendor="Aegis Platform" \
      org.opencontainers.image.licenses="Proprietary" \
      mil.dso.ironbank.image.type="commercial" \
      mil.dso.ironbank.product.name="aegis-proxy"

RUN microdnf update -y && \
    microdnf install -y --nodocs \
      ca-certificates \
      shadow-utils && \
    microdnf clean all

# Create non-root user (replaces distroless nonroot user)
RUN groupadd -g 65532 nonroot && \
    useradd -u 65532 -g 65532 -s /sbin/nologin nonroot

COPY --from=builder --chown=root:root /aegis-auth-proxy /aegis-auth-proxy

RUN find / -xdev -perm /6000 -type f -exec chmod a-s {} \; 2>/dev/null || true

EXPOSE 8080

USER 65532:65532

HEALTHCHECK NONE

ENTRYPOINT ["/aegis-auth-proxy"]
