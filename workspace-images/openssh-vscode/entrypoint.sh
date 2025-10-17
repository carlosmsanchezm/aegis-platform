#!/usr/bin/env bash
set -euo pipefail

log() {
  printf '[entrypoint] %s\n' "$*" >&2
}

# Adjust UID/GID if requested.
if [[ -n "${PGID:-}" ]]; then
  if ! getent group "${PGID}" >/dev/null 2>&1; then
    log "Creating group ${PGID}"
    groupmod -o -g "${PGID}" "${USER_NAME}" 2>/dev/null || groupadd -g "${PGID}" "${USER_NAME}"
  fi
fi

if [[ -n "${PUID:-}" ]]; then
  log "Setting ${USER_NAME} uid=${PUID}"
  usermod -o -u "${PUID}" "${USER_NAME}" 2>/dev/null || true
fi

if [[ -n "${PGID:-}" ]]; then
  log "Ensuring ${USER_NAME} gid=${PGID}"
  usermod -g "${PGID}" "${USER_NAME}" 2>/dev/null || true
fi

# Configure workspace ownership.
mkdir -p "${WORKSPACE_ROOT}"
chown -R "${USER_NAME}:${USER_NAME}" "${WORKSPACE_ROOT}"

# Configure SSH.
SSH_CONF="/etc/ssh/sshd_config"
sed -i "s/^#*Port .*/Port ${SSH_PORT}/" "${SSH_CONF}"
sed -i "s/^#*UsePAM .*/UsePAM no/" "${SSH_CONF}"
if [[ "${PASSWORD_ACCESS,,}" == "true" ]]; then
  sed -i "s/^#*PasswordAuthentication .*/PasswordAuthentication yes/" "${SSH_CONF}"
  echo "${USER_NAME}:${USER_PASSWORD}" | chpasswd
else
  sed -i "s/^#*PasswordAuthentication .*/PasswordAuthentication no/" "${SSH_CONF}"
fi

sed -i "s/^#*PermitRootLogin .*/PermitRootLogin prohibit-password/" "${SSH_CONF}"
echo "AllowUsers ${USER_NAME}" >> "${SSH_CONF}"

mkdir -p /var/run/sshd
ssh-keygen -A

/usr/sbin/sshd -D &
log "sshd started on port ${SSH_PORT}"

exec /usr/local/bin/start-reh.sh "$@"
