#!/usr/bin/env bash
# Backward-compatible entrypoint for hub Helm deploy.
# Implementation lives in scripts/hub-app/deploy-app.sh (phased, fail-fast).
#
# Full history of the previous monolith: scripts/hub-app/generate-cloud-deployment.sh.orig
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DEPLOY_APP="${ROOT}/scripts/hub-app/deploy-app.sh"

# Strip legacy flags for compatibility with old callers
ARGS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --non-interactive)
      # always non-interactive in new script
      shift
      ;;
    --tls)
      echo "⚠️  --tls is deprecated; TLS is always enforced for cloud deployments." >&2
      shift
      ;;
    -h|--help)
      cat <<'EOF'
Usage: ./terraform/generate-cloud-deployment.sh [--non-interactive]

Deploys Aegis hub (platform-api, proxy, Keycloak, Backstage) onto the EKS
cluster described by terraform outputs.

Preferred entrypoint for full lab/prod hub lifecycle:
  ./scripts/hub-eks.sh up

Required env:
  PLATFORM_API_IMAGE_TAG, PROXY_IMAGE_TAG, K8S_AGENT_IMAGE_TAG, UI_IMAGE_TAG
  AWS_PROFILE (default: aegis-lab)

Optional:
  SKIP_DNS_UPDATE=1   (default for lab private deploys)
  REUSE_EXISTING=1     (default: helm upgrade without namespace wipe)
  IMAGE tags as full ECR repo:tag
EOF
      exit 0
      ;;
    *)
      ARGS+=("$1")
      shift
      ;;
  esac
done

if [[ ! -x "${DEPLOY_APP}" ]]; then
  chmod +x "${DEPLOY_APP}" 2>/dev/null || true
fi
if [[ ! -f "${DEPLOY_APP}" ]]; then
  echo "❌ Missing ${DEPLOY_APP}" >&2
  exit 1
fi

# Defaults aligned with hub-eks lab
export AWS_PROFILE="${AWS_PROFILE:-aegis-lab}"
export SKIP_DNS_UPDATE="${SKIP_DNS_UPDATE:-1}"
export REUSE_EXISTING="${REUSE_EXISTING:-1}"
export SKIP_MIGRATION_PLACEHOLDER="${SKIP_MIGRATION_PLACEHOLDER:-1}"

exec bash "${DEPLOY_APP}" "${ARGS[@]+"${ARGS[@]}"}"
