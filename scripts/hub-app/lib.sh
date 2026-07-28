#!/usr/bin/env bash
# Shared helpers for hub-app deploy phases.
# shellcheck disable=SC2034

set -euo pipefail

hub_log()  { printf '[%s] %s\n' "$(date -u +%H:%M:%S)" "$*"; }
hub_die()  { hub_log "ERROR: $*"; exit 1; }
hub_hr()   { hub_log "────────────────────────────────────────"; }

parse_image_ref() {
  local ref="$1"
  local repo="" tag="$ref"
  if [[ -n "$ref" && "$ref" == *:* ]]; then
    repo="${ref%:*}"
    tag="${ref##*:}"
  fi
  printf '%s|%s' "$repo" "$tag"
}

require_image_ref() {
  local var_name="$1" value="$2"
  [[ -n "${value}" ]] || hub_die "${var_name} must be set to repo:tag (e.g. ghcr.io/carlosmsanchezm/aegis/ui:abc1234)"
  [[ "${value}" == *:* ]] || hub_die "${var_name} must include an explicit tag: ${value}"
}

ecr_repo_name() {
  local repo="$1"
  local trimmed="${repo#*.amazonaws.com/}"
  if [[ "${trimmed}" == "${repo}" ]]; then
    printf '%s' ""
    return
  fi
  printf '%s' "${trimmed}"
}

# Verify image exists in GHCR (preferred) or legacy ECR. No local Docker.
verify_registry_image() {
  local image_ref="$1"
  local repo tag code
  IFS='|' read -r repo tag <<< "$(parse_image_ref "${image_ref}")"
  [[ -n "${repo}" && -n "${tag}" ]] || hub_die "Unable to parse image: ${image_ref}"

  if [[ "$repo" == *.dkr.ecr.*.amazonaws.com/* || "$repo" == *".amazonaws.com/"* ]]; then
    local repo_name
    repo_name="$(ecr_repo_name "${repo}")"
    [[ -n "${repo_name}" ]] || hub_die "Unable to parse ECR image: ${image_ref}"
    aws ecr describe-images \
      --repository-name "${repo_name}" \
      --image-ids "imageTag=${tag}" \
      --region "${AWS_REGION}" \
      --profile "${AWS_PROFILE}" >/dev/null 2>&1 \
      || hub_die "ECR image not found: ${image_ref}"
    hub_log "  OK ECR ${image_ref}"
    return 0
  fi

  if [[ "$repo" == ghcr.io/* ]]; then
    if registry_image_exists "${image_ref}"; then
      hub_log "  OK GHCR ${image_ref}"
      return 0
    fi
    hub_die "GHCR image not found or not authorized: ${image_ref}. Run Actions 'Build images (GHCR)' or set GHCR_PULL_TOKEN / make package public. See docs/CI-GHCR-IMAGES.md"
  fi

  hub_log "  OK (skipped remote check) ${image_ref}"
}

# Soft check (no die): 0 if present, 1 if missing.
registry_image_exists() {
  local image_ref="$1"
  local repo tag code
  IFS='|' read -r repo tag <<< "$(parse_image_ref "${image_ref}")"
  [[ -n "${repo}" && -n "${tag}" ]] || return 1
  if [[ "$repo" == ghcr.io/* ]]; then
    local path="${repo#ghcr.io/}"
    local auth=()
    if [[ -n "${GHCR_PULL_TOKEN:-}" ]]; then
      auth=(-H "Authorization: Bearer ${GHCR_PULL_TOKEN}")
    elif [[ -n "${GITHUB_TOKEN:-}" ]]; then
      auth=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
    fi
    code="$(curl -sS -o /dev/null -w '%{http_code}' \
      -H "Accept: application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json" \
      "${auth[@]}" \
      "https://ghcr.io/v2/${path}/manifests/${tag}" 2>/dev/null || echo 000)"
    [[ "$code" == "200" ]]
    return $?
  fi
  if [[ "$repo" == *.dkr.ecr.*.amazonaws.com/* ]]; then
    local repo_name
    repo_name="$(ecr_repo_name "${repo}")"
    aws ecr describe-images \
      --repository-name "${repo_name}" \
      --image-ids "imageTag=${tag}" \
      --region "${AWS_REGION}" \
      --profile "${AWS_PROFILE}" >/dev/null 2>&1
    return $?
  fi
  return 0
}

# Back-compat
verify_ecr_image() { verify_registry_image "$@"; }

wait_for_lb_hostname() {
  local description="$1" resource="$2" namespace="$3" outvar="$4"
  local value=""
  hub_log "Waiting for ${description} LoadBalancer..."
  for _ in $(seq 1 90); do
    value=$(kubectl get "${resource}" -n "${namespace}" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
    if [[ -n "${value}" ]]; then
      hub_log "  ${description} LB ready: ${value}"
      printf -v "${outvar}" '%s' "${value}"
      return 0
    fi
    sleep 2
  done
  hub_die "Timed out waiting for ${description} LoadBalancer (${resource})"
}

wait_for_lb_dns() {
  local hostname="$1"
  hub_log "Waiting for LB DNS ${hostname}..."
  for _ in $(seq 1 45); do
    if dig +short "${hostname}" 2>/dev/null | grep -qE '^[0-9]'; then
      hub_log "  DNS ok: $(dig +short "${hostname}" | head -1)"
      return 0
    fi
    sleep 2
  done
  hub_log "WARN: LB DNS not resolved yet for ${hostname} (continuing)"
}

cloudflare_record_id() {
  local record_name="$1"
  curl -sS "https://api.cloudflare.com/client/v4/zones/${CF_ZONE_ID}/dns_records?name=${record_name}&type=CNAME" \
    -H "Authorization: Bearer ${CF_API_TOKEN}" | jq -r '.result[0].id // empty'
}

update_cloudflare_cname() {
  local record_name="$1" lb_host="$2" record_id
  record_id="$(cloudflare_record_id "${record_name}")"
  [[ -n "${record_id}" ]] || hub_die "Cloudflare record not found: ${record_name}"
  curl -sS -X PATCH "https://api.cloudflare.com/client/v4/zones/${CF_ZONE_ID}/dns_records/${record_id}" \
    -H "Authorization: Bearer ${CF_API_TOKEN}" \
    -H "Content-Type: application/json" \
    --data "{\"content\":\"${lb_host}\"}" >/dev/null
  hub_log "  CF ${record_name} → ${lb_host}"
}

tf() {
  (cd "${TF_DIR}" && AWS_PROFILE="${AWS_PROFILE}" terraform "$@")
}

tf_out() {
  tf output -raw "$1" 2>/dev/null || echo ""
}
