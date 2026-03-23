# =============================================================================
# Aegis Platform API — Iron Bank Compliant Dockerfile
# Pattern: Go binary compiled in IB golang builder, deployed on IB UBI9-minimal
# Reference: repo1.dso.mil/dsop/hashicorp/vault-k8s
# =============================================================================
#
# For Iron Bank builds: all resources (AWS CLI, kubectl, jq, Pulumi) are
# pre-fetched by the Iron Bank pipeline from hardening_manifest.yaml.
# For local builds: override BUILDER_REGISTRY/BASE_REGISTRY with public
# equivalents and stage dependency files in the build context.
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
ARG PULUMI_VERSION=3.226.0
ARG FIPS_ENABLED=true

USER 0

# Install C compiler for CGO (FIPS/BoringCrypto requires CGO_ENABLED=1).
# IB golang builder already has gcc; this handles public golang:slim images.
RUN if ! command -v gcc >/dev/null 2>&1; then \
      if command -v microdnf >/dev/null 2>&1; then microdnf install -y gcc && microdnf clean all; \
      elif command -v dnf >/dev/null 2>&1; then dnf install -y gcc && dnf clean all; \
      elif command -v apt-get >/dev/null 2>&1; then apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev && rm -rf /var/lib/apt/lists/*; \
      fi; \
    fi

WORKDIR /app

COPY go.work* ./
COPY proto/aegis/go.* ./proto/aegis/
COPY services/platform-api/go.* ./services/platform-api/
COPY services/proxy/go.* ./services/proxy/
COPY agents/k8s-agent/go.* ./agents/k8s-agent/
COPY pkg/workspace/go.* ./pkg/workspace/

WORKDIR /app/services/platform-api
RUN GOCACHE=/tmp/go-cache go mod download

# Install Pulumi CLI from pre-fetched resource
WORKDIR /tmp
COPY pulumi-linux-x64.tar.gz /tmp/pulumi.tar.gz
RUN tar -xzf pulumi.tar.gz && \
    mv pulumi/pulumi /tmp/pulumi-cli && \
    chmod +x /tmp/pulumi-cli && \
    rm -rf pulumi pulumi.tar.gz

# Optionally extract pre-downloaded Pulumi plugins
RUN mkdir -p /root/.pulumi
COPY pulumi-plugins-bundle.tar.gz* /tmp/
RUN if [ -f /tmp/pulumi-plugins-bundle.tar.gz ]; then \
      tar -xzf /tmp/pulumi-plugins-bundle.tar.gz -C /root/.pulumi; \
    fi

WORKDIR /app
COPY . .

WORKDIR /app/services/platform-api
RUN GOCACHE=/tmp/go-cache \
    CGO_ENABLED=$(if [ "${FIPS_ENABLED}" = "true" ]; then echo 1; else echo 0; fi) \
    GOEXPERIMENT=$(if [ "${FIPS_ENABLED}" = "true" ]; then echo boringcrypto; fi) \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o platform-api .

# --- Final Stage: Iron Bank UBI9-minimal ---
FROM ${BASE_REGISTRY}/${BASE_IMAGE}:${BASE_TAG}

LABEL org.opencontainers.image.title="aegis-platform-api" \
      org.opencontainers.image.description="Aegis Platform API — gRPC+REST control plane for multi-cluster GPU workload orchestration" \
      org.opencontainers.image.vendor="Aegis Platform" \
      org.opencontainers.image.licenses="Proprietary" \
      mil.dso.ironbank.image.type="opensource" \
      mil.dso.ironbank.product.name="aegis-platform-api"

ARG TARGETARCH=amd64

RUN microdnf update -y && \
    microdnf install -y --nodocs \
      ca-certificates \
      tzdata \
      tar \
      gzip \
      unzip \
      shadow-utils \
      bash && \
    microdnf clean all

# Install pre-fetched runtime tools
COPY awscli.zip /tmp/awscliv2.zip
RUN unzip -q /tmp/awscliv2.zip -d /tmp && \
    /tmp/aws/install --bin-dir /usr/local/bin --install-dir /usr/local/aws-cli && \
    rm -rf /tmp/aws /tmp/awscliv2.zip

COPY kubectl /usr/local/bin/kubectl
RUN chmod +x /usr/local/bin/kubectl

COPY jq /usr/local/bin/jq
RUN chmod +x /usr/local/bin/jq

# Create non-root user
RUN groupadd -g 1000 aegis && \
    useradd -u 1000 -g 1000 -m -d /home/aegis -s /bin/bash aegis

WORKDIR /home/aegis

RUN mkdir -p /tmp/pulumi-backend && chmod 0777 /tmp/pulumi-backend

COPY --from=builder --chown=1000:1000 /app/services/platform-api/platform-api ./platform-api
COPY --from=builder --chown=1000:1000 /app/services/platform-api/migrations ./migrations
COPY --from=builder --chown=1000:1000 /app/charts ./charts
COPY --from=builder --chown=1000:1000 /app/services/platform-api/config/observability ./services/platform-api/config/observability
COPY --from=builder /tmp/pulumi-cli /usr/local/bin/pulumi
COPY --from=builder /root/.pulumi /home/aegis/.pulumi
RUN chown -R 1000:1000 /home/aegis/.pulumi

# EKS token helper script
RUN printf '#!/usr/bin/env bash\nset -euo pipefail\n: "${AEGIS_EKS_CLUSTER_NAME:?cluster name required}"\n: "${AEGIS_EKS_REGION:?region required}"\n: "${AEGIS_EKS_ROLE_ARN:?role arn required}"\n: "${AEGIS_EKS_EXTERNAL_ID:?external id required}"\nassumed=$(aws sts assume-role --role-arn "${AEGIS_EKS_ROLE_ARN}" --role-session-name "aegis-eks-token" --external-id "${AEGIS_EKS_EXTERNAL_ID}")\nexport AWS_ACCESS_KEY_ID=$(echo "${assumed}" | jq -r .Credentials.AccessKeyId)\nexport AWS_SECRET_ACCESS_KEY=$(echo "${assumed}" | jq -r .Credentials.SecretAccessKey)\nexport AWS_SESSION_TOKEN=$(echo "${assumed}" | jq -r .Credentials.SessionToken)\naws eks get-token --cluster-name "${AEGIS_EKS_CLUSTER_NAME}" --region "${AEGIS_EKS_REGION}" --output json\n' \
  > /usr/local/bin/aegis-eks-token && chmod +x /usr/local/bin/aegis-eks-token

# VS Code extension distribution: setup script always bundled, VSIX included if
# present in build context (staged by Makefile from the sovran repo).
RUN mkdir -p /data/extension && chown 1000:1000 /data/extension
COPY --chown=1000:1000 scripts/aegis-setup-vscode.sh /data/extension/aegis-setup-vscode.sh
COPY --chown=1000:1000 aegis-remote.vsi[x] /data/extension/aegis-remote.vsix

EXPOSE 8080 8081

USER 1000:1000

CMD ["./platform-api"]
