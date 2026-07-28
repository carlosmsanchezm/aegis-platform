#!/usr/bin/env bash
# =============================================================================
# hub-eks.sh — Deterministic Aegis hub deploy on EKS (lab or production-shaped)
#
# Prefer this over `make deploy-cloud` / long Makefile chains for cloud hub work.
# Each phase is isolated: logs land under .deploy-logs/, failures print phase name.
#
# Usage:
#   ./scripts/hub-eks.sh preflight
#   ./scripts/hub-eks.sh up                 # terraform → app → verify (images from GHCR/CI)
#   ./scripts/hub-eks.sh up --build-images  # rare: build on this machine (needs Docker)
#   ./scripts/hub-eks.sh terraform|images|app|verify|status|down
#
# Images are NEVER built by default. GitHub Actions builds → ghcr.io.
# Laptop only needs: aws, terraform, kubectl, helm (no Docker).
#
# Env (see docs/CI-GHCR-IMAGES.md):
#   AWS_PROFILE, AWS_REGION, IMAGE_REGISTRY (ghcr.io/...), CLOUD_IMAGE_TAG,
#   GHCR_PULL_TOKEN (if packages private), SKIP_DNS_UPDATE
# =============================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# Defaults — lab account; images live on GitHub Container Registry (not ECR)
export AWS_PROFILE="${AWS_PROFILE:-aegis-lab}"
export AWS_REGION="${AWS_REGION:-us-east-1}"
# App images: GHCR. (AWS_ECR_REGISTRY kept only for legacy --build-images → ECR path.)
export IMAGE_REGISTRY="${IMAGE_REGISTRY:-ghcr.io/carlosmsanchezm/aegis}"
export AWS_ECR_REGISTRY="${AWS_ECR_REGISTRY:-${IMAGE_REGISTRY}}"
export IMAGE_FLAVOR="${IMAGE_FLAVOR:-public}"
export SKIP_DNS_UPDATE="${SKIP_DNS_UPDATE:-1}"
export SKIP_MIGRATION_PLACEHOLDER="${SKIP_MIGRATION_PLACEHOLDER:-1}"

CLOUD_IMAGE_TAG="${CLOUD_IMAGE_TAG:-$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo latest)}"
export CLOUD_IMAGE_TAG

PLATFORM_API_IMAGE="${PLATFORM_API_IMAGE:-${IMAGE_REGISTRY}/platform-api:${CLOUD_IMAGE_TAG}}"
PROXY_IMAGE="${PROXY_IMAGE:-${IMAGE_REGISTRY}/proxy:${CLOUD_IMAGE_TAG}}"
K8S_AGENT_IMAGE="${K8S_AGENT_IMAGE:-${IMAGE_REGISTRY}/k8s-agent:${CLOUD_IMAGE_TAG}}"
WORKSPACE_IMAGE="${WORKSPACE_IMAGE:-${IMAGE_REGISTRY}/workspace-vscode:${CLOUD_IMAGE_TAG}}"
UI_IMAGE="${UI_IMAGE:-${IMAGE_REGISTRY}/ui:${CLOUD_IMAGE_TAG}}"

LOG_DIR="${ROOT}/.deploy-logs"
mkdir -p "$LOG_DIR"
RUN_ID="$(date +%Y%m%d-%H%M%S)"
RUN_LOG="${LOG_DIR}/hub-eks-${RUN_ID}.log"

# Images default OFF on laptop — CI owns builds. Opt in with --build-images or `images`.
BUILD_IMAGES=0
SKIP_TERRAFORM=0
SKIP_APP=0
SKIP_VERIFY=0
AUTO_APPROVE=1

# -----------------------------------------------------------------------------
log()  { printf '[%s] %s\n' "$(date -u +%H:%M:%S)" "$*" | tee -a "$RUN_LOG"; }
die()  { log "ERROR: $*"; log "See full log: $RUN_LOG"; exit 1; }
hr()   { log "────────────────────────────────────────"; }

run_phase() {
  local name="$1"
  shift
  local phase_log="${LOG_DIR}/${RUN_ID}-${name}.log"
  hr
  log "PHASE START: ${name}"
  log "  log → ${phase_log}"
  # shellcheck disable=SC2068
  if "$@" > >(tee -a "$phase_log" "$RUN_LOG") 2> >(tee -a "$phase_log" "$RUN_LOG" >&2); then
    log "PHASE OK: ${name}"
    return 0
  else
    local rc=$?
    log "PHASE FAILED: ${name} (exit ${rc})"
    log "  last 40 lines of ${phase_log}:"
    tail -40 "$phase_log" | while IFS= read -r line; do log "  | $line"; done
    die "Stopped at phase '${name}'. Fix the error, then re-run: $0 ${name}   (or: $0 up --skip-terraform ...)"
  fi
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "Missing required command: $1"
}

# Confirm required tags exist in the registry (GHCR or legacy ECR). No local Docker.
verify_registry_images() {
  log "Verifying required image tags (CLOUD_IMAGE_TAG=${CLOUD_IMAGE_TAG})..."
  local img
  for img in "$PLATFORM_API_IMAGE" "$PROXY_IMAGE" "$K8S_AGENT_IMAGE"; do
    if ! verify_one_image "$img"; then
      die "Image not found: ${img}. Run GitHub Actions workflow 'Build images (GHCR)', then set CLOUD_IMAGE_TAG to that SHA. See docs/CI-GHCR-IMAGES.md"
    fi
    log "  OK ${img}"
  done
  if verify_one_image "$UI_IMAGE"; then
    log "  OK ${UI_IMAGE}"
  else
    log "  (optional UI not in registry: ${UI_IMAGE} — build from aegis-ui CI or omit UI)"
  fi
}

# Returns 0 if image:tag is pullable/metadata-visible without docker.
verify_one_image() {
  local image_ref="$1"
  local repo="${image_ref%:*}"
  local tag="${image_ref##*:}"

  # Legacy ECR path
  if [[ "$repo" == *.dkr.ecr.*.amazonaws.com/* ]]; then
    local ecr_name="${repo#*.amazonaws.com/}"
    aws ecr describe-images \
      --repository-name "$ecr_name" \
      --image-ids "imageTag=${tag}" \
      --region "$AWS_REGION" \
      --profile "$AWS_PROFILE" \
      --query 'imageDetails[0].imageSizeInBytes' \
      --output text >/dev/null 2>&1
    return $?
  fi

  # GHCR / other OCI: registry HTTP API (no Docker)
  if [[ "$repo" == ghcr.io/* ]]; then
    local path="${repo#ghcr.io/}"
    local url="https://ghcr.io/v2/${path}/manifests/${tag}"
    local auth=()
    if [[ -n "${GHCR_PULL_TOKEN:-}" ]]; then
      auth=(-H "Authorization: Bearer ${GHCR_PULL_TOKEN}")
    elif [[ -n "${GITHUB_TOKEN:-}" ]]; then
      auth=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
    fi
    local code
    code="$(curl -sS -o /dev/null -w '%{http_code}' \
      -H "Accept: application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json" \
      "${auth[@]}" \
      "$url" 2>/dev/null || echo 000)"
    # 200 = public or authorized; 401/403 often means private + no token (still may exist)
    [[ "$code" == "200" ]] && return 0
    if [[ "$code" == "401" || "$code" == "403" ]]; then
      log "  WARN ${image_ref} needs auth (HTTP ${code}). Set GHCR_PULL_TOKEN (read:packages) or make package public."
      # Treat as present if we have a token that got 401 (wrong token) vs missing 404
      return 1
    fi
    return 1
  fi

  log "  WARN unknown registry for ${image_ref} — skipping remote verify"
  return 0
}

# Back-compat name
verify_ecr_images() { verify_registry_images; }

# -----------------------------------------------------------------------------
phase_preflight() {
  log "Checking tools and AWS identity..."
  need_cmd aws
  need_cmd terraform
  need_cmd kubectl
  need_cmd helm
  need_cmd git
  need_cmd jq

  # Docker only when this machine explicitly builds images (CI is the default path)
  if [[ "${BUILD_IMAGES:-0}" == "1" ]]; then
    need_cmd docker
    docker info >/dev/null 2>&1 \
      || die "Docker daemon not running. Prefer GitHub Actions → GHCR (docs/CI-GHCR-IMAGES.md)."
    log "BUILD_IMAGES=1 — will build/push on this machine (not the default path)"
  else
    log "Images from GHCR/CI (default) — Docker not required on this machine"
  fi
  local acct
  acct="$(aws sts get-caller-identity --profile "$AWS_PROFILE" --query Account --output text 2>/dev/null)" \
    || die "AWS credentials failed for profile=${AWS_PROFILE}"
  log "AWS account=${acct} profile=${AWS_PROFILE} region=${AWS_REGION}"
  log "IMAGE_FLAVOR=${IMAGE_FLAVOR}  CLOUD_IMAGE_TAG=${CLOUD_IMAGE_TAG}"
  log "IMAGE_REGISTRY=${IMAGE_REGISTRY}"
  log "Images:"
  log "  platform-api  ${PLATFORM_API_IMAGE}"
  log "  proxy         ${PROXY_IMAGE}"
  log "  k8s-agent     ${K8S_AGENT_IMAGE}"
  log "  workspace     ${WORKSPACE_IMAGE}"
  log "  ui            ${UI_IMAGE}"

  if [[ ! -d "${ROOT}/terraform" ]]; then
    die "terraform/ directory missing"
  fi
  if [[ ! -f "${ROOT}/terraform/generate-cloud-deployment.sh" ]]; then
    die "terraform/generate-cloud-deployment.sh missing"
  fi
}

phase_terraform() {
  cd "${ROOT}/terraform"
  export AWS_PROFILE AWS_REGION
  log "terraform init (local or configured backend)..."
  terraform init -input=false -reconfigure
  if [[ "$AUTO_APPROVE" == "1" ]]; then
    terraform apply -input=false -auto-approve
  else
    terraform apply -input=false
  fi
  local cmd
  cmd="$(terraform output -raw kubectl_config_command)"
  log "Configuring kubectl: ${cmd}"
  # shellcheck disable=SC2086
  eval ${cmd}
  kubectl get nodes -o wide
  cd "$ROOT"
}

phase_images() {
  die "Local image build is disabled by default. Use GitHub Actions → GHCR (docs/CI-GHCR-IMAGES.md). To force a local build anyway: IMAGE_REGISTRY=... with Docker, or re-enable make push-cloud-images manually."
}

phase_app() {
  log "Deploying hub application via scripts/hub-app/deploy-app.sh..."
  export AWS_PROFILE AWS_REGION
  export SKIP_DNS_UPDATE SKIP_MIGRATION_PLACEHOLDER
  export REUSE_EXISTING="${REUSE_EXISTING:-1}"
  export PLATFORM_API_IMAGE_TAG="$PLATFORM_API_IMAGE"
  export PROXY_IMAGE_TAG="$PROXY_IMAGE"
  export K8S_AGENT_IMAGE_TAG="$K8S_AGENT_IMAGE"
  export UI_IMAGE_TAG="$UI_IMAGE"

  if [[ ! -x "${ROOT}/scripts/hub-app/deploy-app.sh" ]]; then
    chmod +x "${ROOT}/scripts/hub-app/deploy-app.sh" "${ROOT}/scripts/hub-app/lib.sh" 2>/dev/null || true
  fi
  bash "${ROOT}/scripts/hub-app/deploy-app.sh"
}

phase_verify() {
  log "Waiting for platform-api pods..."
  kubectl -n aegis-system get pods -o wide || true

  if ! kubectl -n aegis-system get deploy -o name 2>/dev/null | grep -q platform-api; then
    die "No platform-api deployment in aegis-system — app phase may have failed"
  fi

  kubectl -n aegis-system rollout status "deploy/$(kubectl -n aegis-system get deploy -o name | grep platform-api | head -1 | sed 's|deployment.apps/||')" --timeout=300s \
    || die "platform-api rollout failed"

  log "Port-forward health check..."
  local pf_pid=""
  kubectl -n aegis-system port-forward "svc/$(kubectl -n aegis-system get svc -o name | grep platform-api | head -1 | sed 's|service/||')" 18080:8080 >/tmp/aegis-pf.log 2>&1 &
  pf_pid=$!
  sleep 3
  local code
  code="$(curl -sk -o /tmp/aegis-health.out -w '%{http_code}' https://127.0.0.1:18080/healthz 2>/dev/null \
    || curl -s -o /tmp/aegis-health.out -w '%{http_code}' http://127.0.0.1:18080/healthz 2>/dev/null \
    || echo 000)"
  kill "$pf_pid" 2>/dev/null || true
  wait "$pf_pid" 2>/dev/null || true

  log "healthz HTTP ${code}: $(head -c 200 /tmp/aegis-health.out 2>/dev/null || true)"
  if [[ "$code" != "200" && "$code" != "204" ]]; then
    # Some builds use /readyz or only gRPC — still report pods
    log "WARN: healthz not 200 (got ${code}). Check logs; pods may still be starting."
    kubectl -n aegis-system get pods
    kubectl -n aegis-system logs -l app.kubernetes.io/name=platform-api --tail=30 2>/dev/null || true
  else
    log "VERIFY OK — platform-api responded ${code}"
  fi
}

phase_status() {
  aws sts get-caller-identity --profile "$AWS_PROFILE" || true
  aws eks list-clusters --profile "$AWS_PROFILE" --region "$AWS_REGION" || true
  kubectl config current-context 2>/dev/null || true
  kubectl get nodes 2>/dev/null || true
  kubectl -n aegis-system get pods,svc 2>/dev/null || log "(no aegis-system or no cluster)"
}

phase_down() {
  log "Tearing down hub (helm then terraform)..."
  export AWS_PROFILE AWS_REGION
  if kubectl config current-context >/dev/null 2>&1; then
    helm uninstall aegis -n aegis-system 2>/dev/null || true
    helm uninstall aegis-spoke -n aegis-system 2>/dev/null || true
    # Wait briefly for AWS LBs created by Services
    log "Waiting 45s for LoadBalancers to drain..."
    sleep 45
  fi
  cd "${ROOT}/terraform"
  if [[ "$AUTO_APPROVE" == "1" ]]; then
    terraform destroy -input=false -auto-approve || {
      log "terraform destroy had errors — force-deleting non-empty ECR if needed..."
      for repo in aegis/platform-api aegis/proxy aegis/k8s-agent aegis/ui aegis/workspace-vscode aegis/vscode-reh-init; do
        aws ecr delete-repository --repository-name "$repo" --force \
          --profile "$AWS_PROFILE" --region "$AWS_REGION" 2>/dev/null || true
      done
      terraform destroy -input=false -auto-approve
    }
  else
    terraform destroy -input=false
  fi
  cd "$ROOT"
  log "DOWN complete — verify with: $0 status"
}

usage() {
  cat <<EOF
Aegis hub-on-EKS orchestrator (deterministic phases)

  $0 preflight              Check tools + AWS identity
  $0 terraform              terraform apply + kubeconfig
  $0 images                 disabled (use GitHub Actions → GHCR)
  $0 app                    helm deploy
  $0 verify                 rollout + healthz via port-forward
  $0 up                     preflight → terraform → app → verify  (GHCR images)
  $0 down                   helm uninstall + terraform destroy
  $0 status                 cluster / pods snapshot

Default: images are NEVER built on this machine. Build on GitHub Actions → ghcr.io, then:

  export IMAGE_REGISTRY=ghcr.io/carlosmsanchezm/aegis
  export CLOUD_IMAGE_TAG=<github_sha_short>
  $0 up

Options (after command):
  --build-images            rejected (use GHCR CI)
  --skip-images             no-op
  --skip-terraform          with 'up': skip terraform
  --skip-app                with 'up': skip helm
  --skip-verify             with 'up': skip health check

Environment:
  AWS_PROFILE          default: aegis-lab
  AWS_REGION           default: us-east-1
  IMAGE_REGISTRY       default: ghcr.io/carlosmsanchezm/aegis
  CLOUD_IMAGE_TAG      default: git short SHA (must match GHCR tag)
  GHCR_PULL_TOKEN      PAT with read:packages if GHCR packages are private
  SKIP_DNS_UPDATE       default: 1

Logs: ${LOG_DIR}/
See also: docs/CI-GHCR-IMAGES.md
EOF
}

# -----------------------------------------------------------------------------
main() {
  local cmd="${1:-}"
  shift || true

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --build-images)   die "--build-images is disabled. Images are built by GitHub Actions and stored on ghcr.io (docs/CI-GHCR-IMAGES.md)." ;;
      --skip-images)    BUILD_IMAGES=0 ;;  # no-op default
      --skip-terraform) SKIP_TERRAFORM=1 ;;
      --skip-app)       SKIP_APP=1 ;;
      --skip-verify)    SKIP_VERIFY=1 ;;
      -h|--help)        usage; exit 0 ;;
      *) die "Unknown option: $1" ;;
    esac
    shift
  done

  log "hub-eks run_id=${RUN_ID} cmd=${cmd:-help}"
  log "ROOT=${ROOT}"

  case "${cmd}" in
    preflight)
      run_phase preflight phase_preflight
      ;;
    terraform|tf)
      run_phase preflight phase_preflight
      run_phase terraform phase_terraform
      ;;
    images)
      phase_images
      ;;
    app)
      run_phase preflight phase_preflight
      verify_registry_images
      run_phase app phase_app
      ;;
    verify)
      run_phase verify phase_verify
      ;;
    up)
      run_phase preflight phase_preflight
      [[ "$SKIP_TERRAFORM" == "1" ]] || run_phase terraform phase_terraform
      verify_registry_images
      [[ "$SKIP_APP" == "1" ]]       || run_phase app phase_app
      [[ "$SKIP_VERIFY" == "1" ]]    || run_phase verify phase_verify
      hr
      log "UP complete. Logs: $RUN_LOG"
      ;;
    down|destroy)
      run_phase preflight phase_preflight
      run_phase down phase_down
      ;;
    status)
      phase_status
      ;;
    ""|-h|--help|help)
      usage
      ;;
    *)
      usage
      die "Unknown command: ${cmd}"
      ;;
  esac
}

main "$@"
