# =============================================================================
# Aegis VS Code REH Init Container — Iron Bank Dockerfile
# =============================================================================
# Contains a pre-downloaded, pinned VS Code Remote Extension Host binary.
# Used as an init container that copies the REH binary to a shared emptyDir
# volume before the workspace container starts.
# =============================================================================

ARG BASE_REGISTRY=registry1.dso.mil
ARG BASE_IMAGE=ironbank/redhat/ubi/ubi9-minimal
ARG BASE_TAG=9.7

FROM ${BASE_REGISTRY}/${BASE_IMAGE}:${BASE_TAG}

LABEL org.opencontainers.image.title="aegis-vscode-reh-init" \
      org.opencontainers.image.description="Init container that injects VS Code Remote Extension Host into Aegis workspace pods" \
      org.opencontainers.image.vendor="Aegis Platform" \
      org.opencontainers.image.licenses="Proprietary" \
      mil.dso.ironbank.image.type="commercial" \
      mil.dso.ironbank.product.name="aegis-vscode-reh-init"

# Install tar for extraction (available in UBI9 base repos)
RUN microdnf update -y && \
    microdnf install -y --nodocs tar gzip && \
    microdnf clean all

# VS Code REH tarball — pre-fetched by Iron Bank pipeline from hardening_manifest
COPY vscode-server-linux-x64.tar.gz /tmp/

# Extract to /reh/bin/current (same path the workspace entrypoint expects)
RUN mkdir -p /reh/bin/current && \
    tar -xzf /tmp/vscode-server-linux-x64.tar.gz -C /reh/bin/current --strip-components=1 && \
    rm -f /tmp/vscode-server-linux-x64.tar.gz && \
    chmod -R 755 /reh

RUN find / -xdev -perm /6000 -type f -exec chmod a-s {} \; 2>/dev/null || true

USER 65534

# When used as init container, the command is overridden to:
#   cp -r /reh/. /shared/reh/
# Uses -r instead of -a because non-root users cannot preserve timestamps.
ENTRYPOINT ["cp", "-r", "/reh/.", "/shared/reh/"]
