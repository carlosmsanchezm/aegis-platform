#!/usr/bin/env bash
# =============================================================================
# hub-eks.sh — Deterministic Aegis hub deploy on EKS (lab or production-shaped)
#
# Prefer this over `make deploy-cloud` / long Makefile chains for cloud hub work.
# Each phase is isolated: logs land under .deploy-logs/, failures print phase name.
#
# Usage:
#   ./scripts/hub-eks.sh preflight
#   ./scripts/hub-eks.sh up                 # terraform → app → verify (images from CI/ECR)
#   ./scripts/hub-eks.sh up --build-images  # rare: build on this machine (needs Docker)
#   ./scripts/hub-eks.sh terraform|images|app|verify|status|down
#
# Images are NEVER built by default. Build on GitLab CI (or run `images` /
# `--build-images` only when you explicitly want a local Docker build).
#
# Env (see docs/HUB-EKS-DEPLOY.md):
#   AWS_PROFILE, AWS_REGION, AWS_ECR_REGISTRY, IMAGE_FLAVOR (public|ironbank),
#   SKIP_DNS_UPDATE, CLOUD_IMAGE_TAG
# =============================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# Defaults — lab account; override via env
export AWS_PROFILE="${AWS_PROFILE:-aegis-lab}"
export AWS_REGION="${AWS_REGION:-us-east-1}"
export AWS_ECR_REGISTRY="${AWS_ECR_REGISTRY:-471147325433.dkr.ecr.us-east-1.amazonaws.com}"
export IMAGE_FLAVOR="${IMAGE_FLAVOR:-public}"
export SKIP_DNS_UPDATE="${SKIP_DNS_UPDATE:-1}"
export SKIP_MIGRATION_PLACEHOLDER="${SKIP_MIGRATION_PLACEHOLDER:-1}"

CLOUD_IMAGE_TAG="${CLOUD_IMAGE_TAG:-$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo latest)}"
export CLOUD_IMAGE_TAG

PLATFORM_API_IMAGE="${PLATFORM_API_IMAGE:-${AWS_ECR_REGISTRY}/aegis/platform-api:${CLOUD_IMAGE_TAG}}"
PROXY_IMAGE="${PROXY_IMAGE:-${AWS_ECR_REGISTRY}/aegis/proxy:${CLOUD_IMAGE_TAG}}"
K8S_AGENT_IMAGE="${K8S_AGENT_IMAGE:-${AWS_ECR_REGISTRY}/aegis/k8s-agent:${CLOUD_IMAGE_TAG}}"
WORKSPACE_IMAGE="${WORKSPACE_IMAGE:-${AWS_ECR_REGISTRY}/aegis/workspace-vscode:${CLOUD_IMAGE_TAG}}"
UI_IMAGE="${UI_IMAGE:-${AWS_ECR_REGISTRY}/aegis/ui:${CLOUD_IMAGE_TAG}}"

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

# Confirm required tags already exist in ECR (used when not building locally).
verify_ecr_images() {
  local img repo tag
  log "Verifying required image tags exist in ECR (CLOUD_IMAGE_TAG=${CLOUD_IMAGE_TAG})..."
  for img in "$PLATFORM_API_IMAGE" "$PROXY_IMAGE" "$K8S_AGENT_IMAGE"; do
    repo="${img#*.amazonaws.com/}"
    repo="${repo%:*}"
    tag="${img##*:}"
    if ! aws ecr describe-images \
      --repository-name "$repo" \
      --image-ids "imageTag=${tag}" \
      --region "$AWS_REGION" \
      --profile "$AWS_PROFILE" \
      --query 'imageDetails[0].imageSizeInBytes' \
      --output text >/dev/null 2>&1; then
      die "ECR missing ${img}. Build on GitLab CI (job build-images), set CLOUD_IMAGE_TAG to that pipeline SHA, then re-run. See docs/CI-GITLAB-IMAGES.md"
    fi
    log "  OK ${img}"
  done
  # UI is optional when CI skipped it
  local ui_repo="${UI_IMAGE#*.amazonaws.com/}"
  ui_repo="${ui_repo%:*}"
  if aws ecr describe-images \
    --repository-name "$ui_repo" \
    --image-ids "imageTag=${UI_IMAGE##*:}" \
    --region "$AWS_REGION" --profile "$AWS_PROFILE" \
    --query 'imageDetails[0].imageSizeInBytes' --output text >/dev/null 2>&1; then
    log "  OK ${UI_IMAGE}"
  else
    log "  (optional UI tag not in ECR: ${UI_IMAGE} — hub may run without Backstage UI)"
  fi
}

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
      || die "Docker daemon not running. Prefer GitLab CI builds (docs/CI-GITLAB-IMAGES.md); only use --build-images / 'images' with Docker."
    log "BUILD_IMAGES=1 — will build/push on this machine"
  else
    log "Images from ECR/CI (default) — Docker not required on this machine"
  fi
  local acct
  acct="$(aws sts get-caller-identity --profile "$AWS_PROFILE" --query Account --output text 2>/dev/null)" \
    || die "AWS credentials failed for profile=${AWS_PROFILE}"
  log "AWS account=${acct} profile=${AWS_PROFILE} region=${AWS_REGION}"
  log "IMAGE_FLAVOR=${IMAGE_FLAVOR}  CLOUD_IMAGE_TAG=${CLOUD_IMAGE_TAG}"
  log "ECR=${AWS_ECR_REGISTRY}"
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
  log "ECR login..."
  aws ecr get-login-password --region "$AWS_REGION" --profile "$AWS_PROFILE" \
    | docker login --username AWS --password-stdin "$AWS_ECR_REGISTRY"

  log "Building/pushing images (IMAGE_FLAVOR=${IMAGE_FLAVOR})..."
  # Use make targets that already encode IMAGE_FLAVOR — but one phase = one log
  make -C "$ROOT" print-image-flavor IMAGE_FLAVOR="$IMAGE_FLAVOR"
  make -C "$ROOT" \
    AWS_PROFILE="$AWS_PROFILE" \
    AWS_REGION="$AWS_REGION" \
    AWS_ECR_REGISTRY="$AWS_ECR_REGISTRY" \
    IMAGE_FLAVOR="$IMAGE_FLAVOR" \
    CLOUD_IMAGE_TAG="$CLOUD_IMAGE_TAG" \
    PUSH_LATEST=0 \
    push-cloud-images

  log "Verifying images in ECR..."
  local img repo tag
  for img in "$PLATFORM_API_IMAGE" "$PROXY_IMAGE" "$K8S_AGENT_IMAGE" "$UI_IMAGE"; do
    repo="${img#*.amazonaws.com/}"
    repo="${repo%:*}"
    tag="${img##*:}"
    aws ecr describe-images \
      --repository-name "$repo" \
      --image-ids "imageTag=${tag}" \
      --region "$AWS_REGION" \
      --profile "$AWS_PROFILE" \
      --query 'imageDetails[0].imageSizeInBytes' \
      --output text >/dev/null \
      || die "ECR missing after push: ${img}"
    log "  OK ${img}"
  done
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
  $0 images                 EXPLICIT local build/push (needs Docker; prefer CI)
  $0 app                    helm deploy via generate-cloud-deployment.sh
  $0 verify                 rollout + healthz via port-forward
  $0 up                     preflight → terraform → app → verify  (ECR/CI images)
  $0 down                   helm uninstall + terraform destroy
  $0 status                 cluster / pods snapshot

Default: images are NEVER built on this machine. Build on GitLab CI, then:

  export CLOUD_IMAGE_TAG=<CI_COMMIT_SHORT_SHA>
  $0 up

Options (after command):
  --build-images            with 'up': also build/push on this machine (needs Docker)
  --skip-images             no-op (kept for old scripts; images already skipped by default)
  --skip-terraform          with 'up': skip terraform
  --skip-app                with 'up': skip helm
  --skip-verify             with 'up': skip health check

Environment:
  AWS_PROFILE          default: aegis-lab
  AWS_REGION           default: us-east-1
  AWS_ECR_REGISTRY      default: 471147325433.dkr.ecr.us-east-1.amazonaws.com
  IMAGE_FLAVOR         public (default) | ironbank
  CLOUD_IMAGE_TAG      default: git short SHA (must match CI-built tag)
  SKIP_DNS_UPDATE       default: 1

Logs: ${LOG_DIR}/
See also: docs/CI-GITLAB-IMAGES.md
EOF
}

# -----------------------------------------------------------------------------
main() {
  local cmd="${1:-}"
  shift || true

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --build-images)   BUILD_IMAGES=1 ;;
      --skip-images)    BUILD_IMAGES=0 ;;  # no-op default; kept so old scripts don't break
      --skip-terraform) SKIP_TERRAFORM=1 ;;
      --skip-app)       SKIP_APP=1 ;;
      --skip-verify)    SKIP_VERIFY=1 ;;
      -h|--help)        usage; exit 0 ;;
      *) die "Unknown option: $1" ;;
    esac
    shift
  done

  # `images` command always means local build
  if [[ "${cmd}" == "images" ]]; then
    BUILD_IMAGES=1
  fi

  log "hub-eks run_id=${RUN_ID} cmd=${cmd:-help} BUILD_IMAGES=${BUILD_IMAGES}"
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
      run_phase preflight phase_preflight
      run_phase images phase_images
      ;;
    app)
      run_phase preflight phase_preflight
      verify_ecr_images
      run_phase app phase_app
      ;;
    verify)
      run_phase verify phase_verify
      ;;
    up)
      run_phase preflight phase_preflight
      [[ "$SKIP_TERRAFORM" == "1" ]] || run_phase terraform phase_terraform
      if [[ "$BUILD_IMAGES" == "1" ]]; then
        run_phase images phase_images
      else
        verify_ecr_images
      fi
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
