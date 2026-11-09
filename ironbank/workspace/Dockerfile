# =============================================================================
# Aegis Workspace — Iron Bank Dockerfile
# Base: Iron Bank NVIDIA CUDA on UBI9
# VS Code REH binary is NOT in this image — injected via init container
# Scripts are in the Repo1 build context under scripts/
# =============================================================================

ARG BASE_REGISTRY=registry1.dso.mil
ARG BASE_IMAGE=ironbank/opensource/nvidia/cuda
ARG BASE_TAG=12.6

FROM ${BASE_REGISTRY}/${BASE_IMAGE}:${BASE_TAG}

LABEL org.opencontainers.image.title="aegis-workspace" \
      org.opencontainers.image.description="Aegis GPU development workspace with CUDA runtime. VS Code server injected via init container." \
      org.opencontainers.image.vendor="Aegis Platform" \
      org.opencontainers.image.licenses="Proprietary" \
      mil.dso.ironbank.image.type="commercial" \
      mil.dso.ironbank.product.name="aegis-workspace"

ENV LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8 \
    WORKSPACE_ROOT=/home/aegis/work \
    VSCODE_SERVER_PORT=11111 \
    VSCODE_QUALITY=stable \
    NVIDIA_VISIBLE_DEVICES=all \
    NVIDIA_DRIVER_CAPABILITIES=compute,utility

USER root

RUN dnf upgrade -y --nodocs && \
    dnf install -y --nodocs \
      bash \
      git \
      ca-certificates \
      procps-ng \
      shadow-utils && \
    dnf clean all && \
    rm -rf /var/cache/dnf /var/cache/yum

COPY tini-amd64 /usr/local/bin/tini
RUN chmod +x /usr/local/bin/tini

RUN groupadd -g 1000 aegis && \
    useradd -u 1000 -g 1000 -m -d /home/aegis -s /bin/bash aegis && \
    mkdir -p /home/aegis/work /reh && \
    chown -R 1000:1000 /home/aegis /reh

COPY scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
COPY scripts/start-reh.sh /usr/local/bin/start-reh.sh
RUN chmod +x /usr/local/bin/entrypoint.sh /usr/local/bin/start-reh.sh

RUN find / -xdev -perm /6000 -type f -exec chmod a-s {} \; 2>/dev/null || true

USER 1000:1000
WORKDIR /home/aegis/work

EXPOSE 11111

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s \
    CMD pgrep -f "code-server" > /dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/tini", "-g", "--"]
CMD ["/usr/local/bin/entrypoint.sh"]
