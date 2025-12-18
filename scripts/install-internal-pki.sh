#!/usr/bin/env bash
# Install and configure Aegis internal PKI (step-ca + cert-manager + step-issuer).
#
# This script is intended to be safe to re-run. It installs:
#   - cert-manager (Jetstack Helm chart) into CERT_MANAGER_NAMESPACE
#   - step-certificates (smallstep/step-certificates Helm chart) into PKI_NAMESPACE
#   - step-issuer (smallstep/step-issuer Helm chart) into CERT_MANAGER_NAMESPACE
# and then creates a StepClusterIssuer that cert-manager can use to mint service
# certificates. It also creates a root CA trust bundle Secret in target namespaces.
#
# Prereqs:
#   - kubectl configured for the target cluster
#   - helm, jq available in PATH
#
# Environment variables:
#   PKI_NAMESPACE            (default: aegis-pki)
#   CERT_MANAGER_NAMESPACE   (default: cert-manager)
#   STEP_CA_RELEASE          (default: step-certificates)
#   STEP_ISSUER_RELEASE      (default: step-issuer)
#   STEP_CA_PROVISIONER_NAME (default: aegis)
#   STEP_CA_NAME             (default: Aegis Internal CA)
#   STEP_CA_SERVICE_PORT     (default: auto; falls back to 443)
#   STEP_CA_MAX_TLS_CERT_DURATION     (default: 2160h)
#   STEP_CA_DEFAULT_TLS_CERT_DURATION (default: 2160h)
#   STEP_CA_MIN_TLS_CERT_DURATION     (default: 5m)
#   STEP_CA_DB_PERSISTENT             (default: true)
#   STEP_CA_REINSTALL_ON_MISMATCH     (default: false)
#   STEP_CLUSTER_ISSUER_NAME (default: aegis-internal)
#   TRUST_BUNDLE_SECRET_NAME (default: aegis-trust-bundle)
#   TRUST_BUNDLE_NAMESPACES  (default: aegis-system,keycloak)
#
# Notes:
# - step-certificates bootstrap job creates encrypted private keys in configmaps.
#   For production environments, consider using external KMS/HSM or existingSecrets.
set -Eeuo pipefail

log() {
  printf '[%s] %s\n' "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" "$1"
}

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    printf 'Missing required command: %s\n' "$cmd" >&2
    exit 1
  fi
}

require_cmd helm
require_cmd kubectl
require_cmd jq
require_cmd base64

PKI_NAMESPACE="${PKI_NAMESPACE:-aegis-pki}"
CERT_MANAGER_NAMESPACE="${CERT_MANAGER_NAMESPACE:-cert-manager}"

STEP_CA_RELEASE="${STEP_CA_RELEASE:-step-certificates}"
STEP_ISSUER_RELEASE="${STEP_ISSUER_RELEASE:-step-issuer}"

STEP_CA_PROVISIONER_NAME="${STEP_CA_PROVISIONER_NAME:-aegis}"
STEP_CA_NAME="${STEP_CA_NAME:-Aegis Internal CA}"
STEP_CLUSTER_ISSUER_NAME="${STEP_CLUSTER_ISSUER_NAME:-aegis-internal}"
STEP_CA_SERVICE_PORT="${STEP_CA_SERVICE_PORT:-}"

STEP_CA_MAX_TLS_CERT_DURATION="${STEP_CA_MAX_TLS_CERT_DURATION:-2160h}"
STEP_CA_DEFAULT_TLS_CERT_DURATION="${STEP_CA_DEFAULT_TLS_CERT_DURATION:-2160h}"
STEP_CA_MIN_TLS_CERT_DURATION="${STEP_CA_MIN_TLS_CERT_DURATION:-5m}"
STEP_CA_DB_PERSISTENT="${STEP_CA_DB_PERSISTENT:-true}"
STEP_CA_REINSTALL_ON_MISMATCH="${STEP_CA_REINSTALL_ON_MISMATCH:-false}"

TRUST_BUNDLE_SECRET_NAME="${TRUST_BUNDLE_SECRET_NAME:-aegis-trust-bundle}"
TRUST_BUNDLE_NAMESPACES="${TRUST_BUNDLE_NAMESPACES:-aegis-system,keycloak}"

add_helm_repos() {
  # Idempotent (helm repo add returns non-zero if already exists in some versions).
  helm repo add jetstack https://charts.jetstack.io >/dev/null 2>&1 || true
  helm repo add smallstep https://smallstep.github.io/helm-charts >/dev/null 2>&1 || true
  helm repo update >/dev/null
}

ensure_namespace() {
  local ns="$1"
  kubectl get namespace "$ns" >/dev/null 2>&1 || kubectl create namespace "$ns" >/dev/null
}

install_cert_manager() {
  log "Installing/upgrading cert-manager in namespace ${CERT_MANAGER_NAMESPACE}"
  ensure_namespace "$CERT_MANAGER_NAMESPACE"
  helm upgrade --install cert-manager jetstack/cert-manager \
    --namespace "$CERT_MANAGER_NAMESPACE" \
    --set installCRDs=true \
    --wait \
    --timeout 10m >/dev/null
}

install_step_ca() {
  log "Installing/upgrading step-certificates (step-ca) in namespace ${PKI_NAMESPACE}"
  ensure_namespace "$PKI_NAMESPACE"

  local desired_persistent="${STEP_CA_DB_PERSISTENT}"
  local db_persistent="${desired_persistent}"
  if kubectl get statefulset "$STEP_CA_RELEASE" -n "$PKI_NAMESPACE" >/dev/null 2>&1; then
    local vct
    vct="$(kubectl get statefulset "$STEP_CA_RELEASE" -n "$PKI_NAMESPACE" -o jsonpath='{.spec.volumeClaimTemplates[0].metadata.name}' 2>/dev/null || true)"
    if [[ -n "$vct" ]]; then
      if [[ "$desired_persistent" != "true" ]]; then
        if [[ "$STEP_CA_REINSTALL_ON_MISMATCH" == "true" ]]; then
          log "step-ca is installed with a persistent DB but STEP_CA_DB_PERSISTENT=false; reinstalling to avoid immutable StatefulSet updates"
          helm uninstall "$STEP_CA_RELEASE" -n "$PKI_NAMESPACE" --wait --timeout 10m >/dev/null 2>&1 || true
          kubectl delete pvc -n "$PKI_NAMESPACE" -l "app.kubernetes.io/instance=${STEP_CA_RELEASE}" --ignore-not-found >/dev/null 2>&1 || true
        else
          log "step-ca already uses a persistent DB; preserving ca.db.persistent=true to avoid immutable StatefulSet updates"
          db_persistent="true"
        fi
      else
        db_persistent="true"
      fi
    else
      if [[ "$desired_persistent" != "false" ]]; then
        if [[ "$STEP_CA_REINSTALL_ON_MISMATCH" == "true" ]]; then
          log "step-ca is installed with an ephemeral DB but STEP_CA_DB_PERSISTENT=true; reinstalling to avoid immutable StatefulSet updates"
          helm uninstall "$STEP_CA_RELEASE" -n "$PKI_NAMESPACE" --wait --timeout 10m >/dev/null 2>&1 || true
        else
          log "step-ca already uses an ephemeral DB; preserving ca.db.persistent=false to avoid immutable StatefulSet updates"
          db_persistent="false"
        fi
      else
        db_persistent="false"
      fi
    fi
  fi

  # The chart's bootstrap job is named after the release; keep release stable.
  helm upgrade --install "$STEP_CA_RELEASE" smallstep/step-certificates \
    --namespace "$PKI_NAMESPACE" \
    --set image.repository=smallstep/step-ca \
    --set bootstrap.image.repository=smallstep/step-ca-bootstrap \
    --set "ca.name=${STEP_CA_NAME}" \
    --set "ca.provisioner.name=${STEP_CA_PROVISIONER_NAME}" \
    --set "ca.db.persistent=${db_persistent}" \
    --wait \
    --timeout 10m >/dev/null

  # Bootstrap job only runs on install, but we still wait for it if present.
  if kubectl get job "$STEP_CA_RELEASE" -n "$PKI_NAMESPACE" >/dev/null 2>&1; then
    log "Waiting for step-ca bootstrap job to complete"
    kubectl wait --for=condition=complete "job/${STEP_CA_RELEASE}" -n "$PKI_NAMESPACE" --timeout=10m >/dev/null
  fi
}

ensure_step_ca_tls_claims() {
  wait_for_step_ca_config

  local cm_config="${STEP_CA_RELEASE}-config"
  local ca_json_raw ca_json_canon desired_json desired_json_canon escaped_json
  ca_json_raw="$(kubectl get configmap "$cm_config" -n "$PKI_NAMESPACE" -o jsonpath='{.data.ca\.json}')"
  ca_json_canon="$(printf '%s' "$ca_json_raw" | jq -c .)"

  desired_json="$(printf '%s' "$ca_json_raw" | jq \
    --arg min "$STEP_CA_MIN_TLS_CERT_DURATION" \
    --arg max "$STEP_CA_MAX_TLS_CERT_DURATION" \
    --arg def "$STEP_CA_DEFAULT_TLS_CERT_DURATION" \
    '
      .authority.claims |= (. // {})
      | .authority.claims.minTLSCertDuration = $min
      | .authority.claims.maxTLSCertDuration = $max
      | .authority.claims.defaultTLSCertDuration = $def
      | .authority.claims.disableRenewal = (.authority.claims.disableRenewal // false)
    ')"
  desired_json_canon="$(printf '%s' "$desired_json" | jq -c .)"

  if [[ "$desired_json_canon" == "$ca_json_canon" ]]; then
    return 0
  fi

  log "Configuring step-ca TLS claims (max=${STEP_CA_MAX_TLS_CERT_DURATION})"
  escaped_json="$(printf '%s' "$desired_json" | jq -Rs .)"
  kubectl patch configmap "$cm_config" -n "$PKI_NAMESPACE" --type merge \
    -p "{\"data\":{\"ca.json\":${escaped_json}}}" >/dev/null

  log "Restarting step-ca to apply config changes"
  kubectl rollout restart "statefulset/${STEP_CA_RELEASE}" -n "$PKI_NAMESPACE" >/dev/null
  kubectl rollout status "statefulset/${STEP_CA_RELEASE}" -n "$PKI_NAMESPACE" --timeout=5m >/dev/null
}

install_step_issuer() {
  log "Installing/upgrading step-issuer in namespace ${CERT_MANAGER_NAMESPACE}"
  ensure_namespace "$CERT_MANAGER_NAMESPACE"
  helm upgrade --install "$STEP_ISSUER_RELEASE" smallstep/step-issuer \
    --namespace "$CERT_MANAGER_NAMESPACE" \
    --set image.repository=smallstep/step-issuer \
    --wait \
    --timeout 10m >/dev/null
}

wait_for_step_ca_config() {
  local cm_config="${STEP_CA_RELEASE}-config"
  local cm_certs="${STEP_CA_RELEASE}-certs"

  log "Waiting for step-ca configmaps to be populated"
  local i ca_json root_ca
  for i in {1..120}; do
    ca_json="$(kubectl get configmap "$cm_config" -n "$PKI_NAMESPACE" -o jsonpath='{.data.ca\.json}' 2>/dev/null || true)"
    root_ca="$(kubectl get configmap "$cm_certs" -n "$PKI_NAMESPACE" -o jsonpath='{.data.root_ca\.crt}' 2>/dev/null || true)"
    if [[ -n "$ca_json" && -n "$root_ca" ]]; then
      return 0
    fi
    sleep 2
  done
  printf 'Timed out waiting for step-ca configmaps (%s, %s) in %s\n' "$cm_config" "$cm_certs" "$PKI_NAMESPACE" >&2
  return 1
}

create_step_cluster_issuer() {
  wait_for_step_ca_config

  local cm_config="${STEP_CA_RELEASE}-config"
  local cm_certs="${STEP_CA_RELEASE}-certs"
  local password_secret="${STEP_CA_RELEASE}-provisioner-password"

  log "Creating/updating StepClusterIssuer ${STEP_CLUSTER_ISSUER_NAME}"
  local ca_json root_ca kid ca_bundle_b64 step_ca_port step_ca_url
  ca_json="$(kubectl get configmap "$cm_config" -n "$PKI_NAMESPACE" -o jsonpath='{.data.ca\.json}')"
  root_ca="$(kubectl get configmap "$cm_certs" -n "$PKI_NAMESPACE" -o jsonpath='{.data.root_ca\.crt}')"

  kid="$(
    printf '%s' "$ca_json" \
      | jq -r --arg name "$STEP_CA_PROVISIONER_NAME" '
          .authority.provisioners
          | map(select(.type=="JWK" and .name==$name))
          | .[0].key.kid // empty
        '
  )"
  if [[ -z "$kid" ]]; then
    printf 'Unable to locate JWK provisioner kid for provisioner "%s"\n' "$STEP_CA_PROVISIONER_NAME" >&2
    return 1
  fi

  ca_bundle_b64="$(printf '%s' "$root_ca" | base64 | tr -d '\n')"
  step_ca_port="${STEP_CA_SERVICE_PORT}"
  if [[ -z "${step_ca_port}" ]]; then
    step_ca_port="$(kubectl get svc "${STEP_CA_RELEASE}" -n "${PKI_NAMESPACE}" -o jsonpath='{.spec.ports[0].port}' 2>/dev/null || true)"
  fi
  step_ca_port="${step_ca_port:-443}"
  step_ca_url="https://${STEP_CA_RELEASE}.${PKI_NAMESPACE}.svc.cluster.local:${step_ca_port}"

  cat <<EOF | kubectl apply -f - >/dev/null
apiVersion: certmanager.step.sm/v1beta1
kind: StepClusterIssuer
metadata:
  name: ${STEP_CLUSTER_ISSUER_NAME}
spec:
  url: ${step_ca_url}
  caBundle: ${ca_bundle_b64}
  provisioner:
    name: ${STEP_CA_PROVISIONER_NAME}
    kid: ${kid}
    passwordRef:
      name: ${password_secret}
      namespace: ${PKI_NAMESPACE}
      key: password
EOF
}

create_trust_bundle_secrets() {
  wait_for_step_ca_config

  local cm_certs="${STEP_CA_RELEASE}-certs"
  local root_ca
  root_ca="$(kubectl get configmap "$cm_certs" -n "$PKI_NAMESPACE" -o jsonpath='{.data.root_ca\.crt}')"

  local tmp
  tmp="$(mktemp)"
  printf '%s\n' "$root_ca" >"$tmp"

  local ns
  IFS=',' read -r -a namespaces <<<"$TRUST_BUNDLE_NAMESPACES"
  for ns in "${namespaces[@]}"; do
    ns="$(echo "$ns" | xargs)"
    [[ -z "$ns" ]] && continue
    ensure_namespace "$ns"
    log "Creating/updating trust bundle secret ${TRUST_BUNDLE_SECRET_NAME} in namespace ${ns}"
    kubectl create secret generic "$TRUST_BUNDLE_SECRET_NAME" \
      --namespace "$ns" \
      --from-file=ca.crt="$tmp" \
      --from-file=aegis-local-trust.pem="$tmp" \
      --dry-run=client -o yaml | kubectl apply -f - >/dev/null
  done

  rm -f "$tmp"
}

main() {
  add_helm_repos
  install_cert_manager
  install_step_ca
  ensure_step_ca_tls_claims
  install_step_issuer
  create_step_cluster_issuer
  create_trust_bundle_secrets
  log "Internal PKI ready: issuer=${STEP_CLUSTER_ISSUER_NAME}, step-ca=${STEP_CA_RELEASE}.${PKI_NAMESPACE}"
}

main "$@"
