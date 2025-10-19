FROM debian:12-slim

ARG TARGETARCH

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    PUID=1000 \
    PGID=1000 \
    USER_NAME=aegis \
    USER_PASSWORD=aegis123 \
    PASSWORD_ACCESS=true \
    SSH_PORT=2222 \
    WORKSPACE_ROOT=/home/project \
    VSCODE_QUALITY=insider

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      openssh-server \
      curl \
      ca-certificates \
      sudo \
      tini \
      bash \
      jq \
      xz-utils \
      git \
 && rm -rf /var/lib/apt/lists/* \
 && mkdir -p /var/run/sshd "${WORKSPACE_ROOT}" \
 && useradd -m -s /bin/bash "${USER_NAME}" \
 && echo "${USER_NAME} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/${USER_NAME} \
 && chmod 0440 /etc/sudoers.d/${USER_NAME}

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
COPY start-reh.sh /usr/local/bin/start-reh.sh

RUN chmod +x /usr/local/bin/entrypoint.sh /usr/local/bin/start-reh.sh

EXPOSE 2222 11111

ENTRYPOINT ["/usr/bin/tini","--","/usr/local/bin/entrypoint.sh"]
