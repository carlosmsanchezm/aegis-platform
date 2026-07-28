#!/usr/bin/env bash
# =============================================================================
# deploy-app.sh — Phased Helm deploy of Aegis hub onto an existing EKS cluster
#
# Replaces the monolithic generate-cloud-deployment.sh as the preferred entrypoint.
# That file is now a thin wrapper for backward compatibility.
#
# Prerequisites: terraform apply done; images in ECR; kubectl can reach cluster
#
#   AWS_PROFILE=aegis-lab \
#   PLATFORM_API_IMAGE_TAG=...:tag PROXY_IMAGE_TAG=... K8S_AGENT_IMAGE_TAG=... UI_IMAGE_TAG=... \
#   SKIP_DNS_UPDATE=1 \
#   ./scripts/hub-app/deploy-app.sh
#
# Env:
#   SKIP_DNS_UPDATE=1     private/lab (default for hub-eks): no Cloudflare; skip public DNS checks
#   REUSE_EXISTING=1       default: do NOT wipe namespace on re-run
#   SKIP_MIGRATION_PLACEHOLDER=1  in-cluster Postgres migrations after helm (lab default)
# =============================================================================
set -euo pipefail

HUB_APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${HUB_APP_DIR}/lib.sh"

ROOT="$(cd "${HUB_APP_DIR}/../.." && pwd)"
TF_DIR="${ROOT}/terraform"
CHARTS_DIR="${ROOT}/charts"
PKI_SCRIPT="${ROOT}/scripts/install-internal-pki.sh"

export AWS_PROFILE="${AWS_PROFILE:-aegis-lab}"
export AWS_REGION="${AWS_REGION:-us-east-1}"
export SKIP_DNS_UPDATE="${SKIP_DNS_UPDATE:-${SKIP_ROUTE53_UPDATE:-1}}"
export SKIP_MIGRATION_PLACEHOLDER="${SKIP_MIGRATION_PLACEHOLDER:-1}"
export REUSE_EXISTING="${REUSE_EXISTING:-1}"
export REMOVE_LEGACY_UI_RELEASE="${REMOVE_LEGACY_UI_RELEASE:-1}"

K8S_NAMESPACE="${K8S_NAMESPACE:-aegis-system}"
HELM_RELEASE="${HELM_RELEASE:-aegis}"
SPOKE_NAMESPACE="${SPOKE_NAMESPACE:-${K8S_NAMESPACE}}"
SPOKE_HELM_RELEASE="${SPOKE_HELM_RELEASE:-${HELM_RELEASE}-spoke}"
LEGACY_UI_RELEASE="${LEGACY_UI_RELEASE:-aegis-ui}"
PKI_NAMESPACE="${PKI_NAMESPACE:-aegis-pki}"
CERT_MANAGER_NAMESPACE="${CERT_MANAGER_NAMESPACE:-cert-manager}"
PLATFORM_API_SERVICE_ACCOUNT="${PLATFORM_API_SERVICE_ACCOUNT:-aegis-platform-api}"

PLATFORM_API_IMAGE_TAG="${PLATFORM_API_IMAGE_TAG:-}"
PROXY_IMAGE_TAG="${PROXY_IMAGE_TAG:-}"
K8S_AGENT_IMAGE_TAG="${K8S_AGENT_IMAGE_TAG:-}"
UI_IMAGE_TAG="${UI_IMAGE_TAG:-}"

KEYCLOAK_ADMIN_USERNAME="${KEYCLOAK_ADMIN_USERNAME:-admin}"
KEYCLOAK_ADMIN_PASSWORD="${KEYCLOAK_ADMIN_PASSWORD:-PreviewAdmin123!}"
KEYCLOAK_DB_USERNAME="${KEYCLOAK_DB_USERNAME:-keycloak}"
KEYCLOAK_DB_PASSWORD="${KEYCLOAK_DB_PASSWORD:-PreviewKeycloakDb123!}"
KEYCLOAK_BACKSTAGE_CLIENT_SECRET="${KEYCLOAK_BACKSTAGE_CLIENT_SECRET:-preview-backstage-client-secret}"
KEYCLOAK_ADMIN_SECRET_NAME="${KEYCLOAK_ADMIN_SECRET_NAME:-keycloak-admin-secret}"
KEYCLOAK_DB_SECRET_NAME="${KEYCLOAK_DB_SECRET_NAME:-keycloak-db-secret}"
KEYCLOAK_CLIENT_SECRET_NAME="${KEYCLOAK_CLIENT_SECRET_NAME:-keycloak-backstage-client-secret}"
KEYCLOAK_TLS_SECRET_NAME="${KEYCLOAK_TLS_SECRET_NAME:-keycloak-tls}"

PLATFORM_API_PUBLIC_HOST="${PLATFORM_API_PUBLIC_HOST:-platform-api.aegis-platform.tech}"
KEYCLOAK_PUBLIC_HOST="${KEYCLOAK_PUBLIC_HOST:-keycloak.aegis-platform.tech}"
PROXY_PUBLIC_HOST="${PROXY_PUBLIC_HOST:-proxy.aegis-platform.tech}"
UI_PUBLIC_HOST="${UI_PUBLIC_HOST:-ui.aegis-platform.tech}"
CF_API_TOKEN="${CLOUDFLARE_API_TOKEN:-}"
CF_ZONE_ID=""
CA_BUNDLE="${HOME}/aegis-platform-api-ca.crt"

RELEASE_BASENAME=$(printf '%s' "${HELM_RELEASE}" | cut -c1-63)
RELEASE_BASENAME=${RELEASE_BASENAME%-}
PLATFORM_API_RELEASE_NAME="${RELEASE_BASENAME}-platform-api"
PROXY_RELEASE_NAME="${RELEASE_BASENAME}-proxy"

# -----------------------------------------------------------------------------
phase_preflight() {
  hub_hr
  hub_log "PHASE app/preflight"
  [[ -d "${TF_DIR}/.terraform" ]] || hub_die "Terraform not initialized in ${TF_DIR}"
  # Accept local state file OR remote backend (no local tfstate file)
  if [[ ! -f "${TF_DIR}/terraform.tfstate" ]] && ! tf state list >/dev/null 2>&1; then
    hub_die "No Terraform state. Run hub-eks terraform / terraform apply first."
  fi

  require_image_ref "PLATFORM_API_IMAGE_TAG" "${PLATFORM_API_IMAGE_TAG}"
  require_image_ref "PROXY_IMAGE_TAG" "${PROXY_IMAGE_TAG}"
  require_image_ref "K8S_AGENT_IMAGE_TAG" "${K8S_AGENT_IMAGE_TAG}"
  require_image_ref "UI_IMAGE_TAG" "${UI_IMAGE_TAG}"

  CF_ZONE_ID="$(tf_out cloudflare_zone_id)"
  if [[ "${SKIP_DNS_UPDATE}" != "1" ]]; then
    [[ -n "${CF_API_TOKEN}" ]] || hub_die "CLOUDFLARE_API_TOKEN required when SKIP_DNS_UPDATE!=1"
    [[ -n "${CF_ZONE_ID}" ]] || hub_die "Cloudflare zone empty — enable_cloudflare=true in terraform or set SKIP_DNS_UPDATE=1"
  else
    hub_log "SKIP_DNS_UPDATE=1 — private/lab mode (no Cloudflare required)"
  fi

  hub_log "Verifying registry images (GHCR preferred)..."
  verify_registry_image "${PLATFORM_API_IMAGE_TAG}"
  verify_registry_image "${PROXY_IMAGE_TAG}"
  verify_registry_image "${K8S_AGENT_IMAGE_TAG}"
  # UI may be published by aegis-ui CI; warn only if missing
  if registry_image_exists "${UI_IMAGE_TAG}"; then
    hub_log "  OK GHCR ${UI_IMAGE_TAG}"
  else
    hub_log "WARN: UI image not found yet: ${UI_IMAGE_TAG} (hub may start without UI)"
  fi

  IFS='|' read -r PLATFORM_API_IMAGE_REPO PLATFORM_API_IMAGE_TAG_VALUE <<< "$(parse_image_ref "${PLATFORM_API_IMAGE_TAG}")"
  IFS='|' read -r PROXY_IMAGE_REPO PROXY_IMAGE_TAG_VALUE <<< "$(parse_image_ref "${PROXY_IMAGE_TAG}")"
  IFS='|' read -r K8S_AGENT_IMAGE_REPO K8S_AGENT_IMAGE_TAG_VALUE <<< "$(parse_image_ref "${K8S_AGENT_IMAGE_TAG}")"
  IFS='|' read -r UI_IMAGE_REPO UI_IMAGE_TAG_VALUE <<< "$(parse_image_ref "${UI_IMAGE_TAG}")"

  export PLATFORM_API_IMAGE_REPO PLATFORM_API_IMAGE_TAG_VALUE
  export PROXY_IMAGE_REPO PROXY_IMAGE_TAG_VALUE
  export K8S_AGENT_IMAGE_REPO K8S_AGENT_IMAGE_TAG_VALUE
  export UI_IMAGE_REPO UI_IMAGE_TAG_VALUE
}

phase_kubeconfig() {
  hub_hr
  hub_log "PHASE app/kubeconfig"
  local cmd
  cmd="$(tf_out kubectl_config_command)"
  [[ -n "$cmd" ]] || hub_die "kubectl_config_command empty"
  hub_log "Running: ${cmd}"
  eval "${cmd}"
  kubectl get nodes -o wide
}

phase_secrets() {
  hub_hr
  hub_log "PHASE app/secrets"
  DB_PASSWORD="$(tf_out db_password_secret_value)"
  JWT_SECRET="$(tf_out jwt_secret_value)"
  SPOKE_OIDC_CLIENT_SECRET="$(tf_out spoke_oidc_client_secret)"
  [[ -n "${SPOKE_OIDC_CLIENT_SECRET}" ]] || SPOKE_OIDC_CLIENT_SECRET="rEC99sBBWQAbRgg0xRQFBsMC8rt6pZOB"
  export DB_PASSWORD JWT_SECRET SPOKE_OIDC_CLIENT_SECRET

  kubectl create namespace "${K8S_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

  if [[ "${REUSE_EXISTING}" != "1" ]]; then
    hub_log "REUSE_EXISTING!=1 — uninstalling prior release (namespace kept)"
    helm uninstall "${HELM_RELEASE}" -n "${K8S_NAMESPACE}" 2>/dev/null || true
    helm uninstall "${SPOKE_HELM_RELEASE}" -n "${K8S_NAMESPACE}" 2>/dev/null || true
  else
    hub_log "REUSE_EXISTING=1 — keeping existing helm release for upgrade"
  fi

  if helm status "${LEGACY_UI_RELEASE}" -n "${K8S_NAMESPACE}" >/dev/null 2>&1; then
    if [[ "${REMOVE_LEGACY_UI_RELEASE}" == "1" ]]; then
      helm uninstall "${LEGACY_UI_RELEASE}" -n "${K8S_NAMESPACE}" 2>/dev/null || true
    else
      hub_die "Legacy UI release ${LEGACY_UI_RELEASE} exists; set REMOVE_LEGACY_UI_RELEASE=1"
    fi
  fi

  kubectl create secret generic aegis-platform-secrets \
    --from-literal=db-password="${DB_PASSWORD}" \
    --from-literal=proxy-jwt-secret="${JWT_SECRET}" \
    --namespace "${K8S_NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -

  kubectl create secret generic "${KEYCLOAK_ADMIN_SECRET_NAME}" \
    --from-literal=username="${KEYCLOAK_ADMIN_USERNAME}" \
    --from-literal=password="${KEYCLOAK_ADMIN_PASSWORD}" \
    --namespace "${K8S_NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -

  kubectl create secret generic "${KEYCLOAK_DB_SECRET_NAME}" \
    --from-literal=username="${KEYCLOAK_DB_USERNAME}" \
    --from-literal=password="${KEYCLOAK_DB_PASSWORD}" \
    --namespace "${K8S_NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -

  kubectl create secret generic "${KEYCLOAK_CLIENT_SECRET_NAME}" \
    --from-literal=clientSecret="${KEYCLOAK_BACKSTAGE_CLIENT_SECRET}" \
    --namespace "${K8S_NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -

  # GHCR pull secret for EKS nodes (private packages). Prefer public packages.
  if [[ -n "${GHCR_PULL_TOKEN:-${GITHUB_TOKEN:-}}" ]]; then
    local gh_user="${GHCR_USERNAME:-${GITHUB_USER:-carlosmsanchezm}}"
    local gh_tok="${GHCR_PULL_TOKEN:-${GITHUB_TOKEN}}"
    kubectl create secret docker-registry ghcr-registry-secret \
      --docker-server=ghcr.io \
      --docker-username="${gh_user}" \
      --docker-password="${gh_tok}" \
      --namespace "${K8S_NAMESPACE}" \
      --dry-run=client -o yaml | kubectl apply -f -
    hub_log "Created/updated ghcr-registry-secret (imagePullSecrets)"
  else
    hub_log "No GHCR_PULL_TOKEN — assuming public GHCR packages (or set token for private)"
  fi

  # Soft reset Keycloak storage so storage class changes apply (idempotent)
  kubectl delete pvc -n "${K8S_NAMESPACE}" -l app.kubernetes.io/component=keycloak-postgres --ignore-not-found --wait=false >/dev/null 2>&1 || true
  kubectl delete statefulset "${HELM_RELEASE}-keycloak" -n "${K8S_NAMESPACE}" --ignore-not-found --wait=false >/dev/null 2>&1 || true
  kubectl delete statefulset "${HELM_RELEASE}-keycloak-db" -n "${K8S_NAMESPACE}" --ignore-not-found --wait=false >/dev/null 2>&1 || true

  cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${PLATFORM_API_SERVICE_ACCOUNT}
  namespace: ${K8S_NAMESPACE}
  labels:
    app.kubernetes.io/name: aegis-services
    app.kubernetes.io/instance: ${HELM_RELEASE}
    app.kubernetes.io/component: platform-api
    app.kubernetes.io/managed-by: Helm
  annotations:
    meta.helm.sh/release-name: ${HELM_RELEASE}
    meta.helm.sh/release-namespace: ${K8S_NAMESPACE}
EOF
  hub_log "Secrets + ServiceAccount ready"
}

phase_values() {
  hub_hr
  hub_log "PHASE app/helm-values"
  mkdir -p "${CHARTS_DIR}/aegis-services" "${CHARTS_DIR}/aegis-spoke"
  tf output -raw helm_values_aegis_services > "${CHARTS_DIR}/aegis-services/values-cloud-generated.yaml"
  tf output -raw helm_values_aegis_spoke > "${CHARTS_DIR}/aegis-spoke/values-cloud-generated.yaml"
  hub_log "Wrote values-cloud-generated.yaml for services + spoke"
}

phase_db_url() {
  hub_hr
  hub_log "PHASE app/db-url"
  local DB_PASSWORD_RAW DB_PASSWORD_ENCODED DB_ENDPOINT DB_NAME DB_USER DB_HOST DB_PORT DB_SSLMODE
  DB_PASSWORD_RAW="$(tf_out db_password_secret_value)"
  DB_PASSWORD_ENCODED="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "${DB_PASSWORD_RAW}")"
  DB_ENDPOINT="$(tf_out rds_endpoint)"
  DB_NAME="$(tf_out rds_database_name)"
  [[ -n "${DB_NAME}" ]] || DB_NAME="aegis_platform"
  DB_USER="$(tf_out rds_username)"
  [[ -n "${DB_USER}" ]] || DB_USER="aegis_platform"

  if [[ -n "${DB_ENDPOINT}" ]]; then
    if [[ "${DB_ENDPOINT}" == *:* ]]; then
      DB_HOST=${DB_ENDPOINT%%:*}
      DB_PORT=${DB_ENDPOINT##*:}
    else
      DB_HOST=${DB_ENDPOINT}
      DB_PORT="$(tf_out rds_port)"
      [[ -n "${DB_PORT}" ]] || DB_PORT="5432"
    fi
    DB_SSLMODE="require"
    hub_log "Using RDS ${DB_HOST}:${DB_PORT}"
  else
    DB_HOST="platform-postgres.${K8S_NAMESPACE}.svc.cluster.local"
    DB_PORT="5432"
    DB_SSLMODE="disable"
    export SKIP_MIGRATION_PLACEHOLDER=1
    hub_log "No RDS — in-cluster Postgres (migrations after helm)"
  fi
  DB_URL="postgres://${DB_USER}:${DB_PASSWORD_ENCODED}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
  export DB_URL DB_PASSWORD="${DB_PASSWORD_RAW}"
}

phase_pki() {
  hub_hr
  hub_log "PHASE app/pki"
  DNS_PLATFORM_API="$(tf_out dns_platform_api)"
  DNS_KEYCLOAK="$(tf_out dns_keycloak)"
  DNS_PROXY="$(tf_out dns_proxy)"
  DNS_UI="$(tf_out dns_ui)"
  [[ -n "${DNS_PLATFORM_API}" ]] || DNS_PLATFORM_API="${PLATFORM_API_PUBLIC_HOST}"
  [[ -n "${DNS_KEYCLOAK}" ]] || DNS_KEYCLOAK="${KEYCLOAK_PUBLIC_HOST}"
  [[ -n "${DNS_PROXY}" ]] || DNS_PROXY="${PROXY_PUBLIC_HOST}"
  [[ -n "${DNS_UI}" ]] || DNS_UI="${UI_PUBLIC_HOST}"
  export DNS_PLATFORM_API DNS_KEYCLOAK DNS_PROXY DNS_UI
  hub_log "Hosts: api=${DNS_PLATFORM_API} kc=${DNS_KEYCLOAK} proxy=${DNS_PROXY} ui=${DNS_UI}"

  [[ -x "${PKI_SCRIPT}" ]] || hub_die "PKI installer missing: ${PKI_SCRIPT}"

  local STEP_CA_SVC_NLB="step-certificates-nlb"
  kubectl create namespace "${PKI_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f - >/dev/null
  if ! kubectl get svc "${STEP_CA_SVC_NLB}" -n "${PKI_NAMESPACE}" >/dev/null 2>&1; then
    hub_log "Creating step-ca NLB service..."
    kubectl apply -f - <<EOSVC
apiVersion: v1
kind: Service
metadata:
  name: ${STEP_CA_SVC_NLB}
  namespace: ${PKI_NAMESPACE}
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
    service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
spec:
  type: LoadBalancer
  selector:
    app.kubernetes.io/name: step-certificates
    app.kubernetes.io/instance: step-certificates
  ports:
    - name: https
      port: 443
      targetPort: 9000
      protocol: TCP
EOSVC
  fi

  STEP_CA_NLB_HOST=""
  for _ in $(seq 1 60); do
    STEP_CA_NLB_HOST=$(kubectl get svc "${STEP_CA_SVC_NLB}" -n "${PKI_NAMESPACE}" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || true)
    [[ -n "${STEP_CA_NLB_HOST}" ]] && break
    sleep 5
  done
  if [[ -n "${STEP_CA_NLB_HOST}" ]]; then
    export STEP_CA_EXTERNAL_URL="https://${STEP_CA_NLB_HOST}"
    hub_log "Step-CA NLB: ${STEP_CA_EXTERNAL_URL}"
  else
    hub_log "WARN: Step-CA NLB not ready; spoke external CA may be limited"
    export STEP_CA_EXTERNAL_URL=""
  fi

  TRUST_BUNDLE_NAMESPACES="${K8S_NAMESPACE},${SPOKE_NAMESPACE}" \
  PKI_NAMESPACE="${PKI_NAMESPACE}" \
  CERT_MANAGER_NAMESPACE="${CERT_MANAGER_NAMESPACE}" \
  STEP_CA_DB_PERSISTENT=false \
  STEP_CA_REINSTALL_ON_MISMATCH=true \
  STEP_CA_EXTERNAL_DNS_NAMES="${STEP_CA_NLB_HOST}" \
    "${PKI_SCRIPT}"

  mkdir -p "$(dirname "${CA_BUNDLE}")"
  kubectl get secret aegis-trust-bundle -n "${K8S_NAMESPACE}" -o "jsonpath={.data.ca\\.crt}" | base64 --decode > "${CA_BUNDLE}"
  hub_log "CA bundle → ${CA_BUNDLE}"

  STEP_PROV_KID=$(kubectl exec step-certificates-0 -n "${PKI_NAMESPACE}" -- \
    cat /home/step/config/ca.json 2>/dev/null \
    | grep -o '"kid"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 \
    | sed 's/.*"kid"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' || echo "")
  STEP_PROV_PASS=$(kubectl get secret step-certificates-provisioner-password \
    -n "${PKI_NAMESPACE}" -o jsonpath='{.data.password}' 2>/dev/null \
    | base64 --decode || echo "")
  export STEP_PROV_KID STEP_PROV_PASS
}

phase_crds() {
  hub_hr
  hub_log "PHASE app/crds"
  local CRD_BASE_DIR="${ROOT}/agents/k8s-agent/config/crd/bases"
  if [[ -d "${CRD_BASE_DIR}" ]]; then
    kubectl apply -f "${CRD_BASE_DIR}" >/dev/null
    hub_log "CRDs applied from ${CRD_BASE_DIR}"
  else
    hub_log "WARN: no CRD dir ${CRD_BASE_DIR}"
  fi
}

phase_helm() {
  hub_hr
  hub_log "PHASE app/helm"
  local JWT_SECRET BACKEND_SECRET OVERRIDE_FILE
  JWT_SECRET="$(tf_out jwt_secret_value)"
  BACKEND_SECRET="$(openssl rand -hex 32)"
  OVERRIDE_FILE="$(mktemp)"

  {
    echo "global:"
    echo "  imagePullSecrets:"
    if kubectl get secret ghcr-registry-secret -n "${K8S_NAMESPACE}" >/dev/null 2>&1; then
      echo "    - name: ghcr-registry-secret"
    else
      echo "    []"
    fi
    echo "ingressController:"
    echo "  enabled: true"
    echo "ingress-nginx:"
    echo "  controller:"
    echo "    service:"
    echo "      annotations:"
    echo "        service.beta.kubernetes.io/aws-load-balancer-type: \"nlb\""
    echo "        service.beta.kubernetes.io/aws-load-balancer-scheme: \"internet-facing\""
    echo "platformApi:"
    echo "  image:"
    [[ -n "${PLATFORM_API_IMAGE_REPO}" ]] && echo "    repository: ${PLATFORM_API_IMAGE_REPO}"
    [[ -n "${PLATFORM_API_IMAGE_TAG_VALUE}" ]] && echo "    tag: \"${PLATFORM_API_IMAGE_TAG_VALUE}\""
    echo "  imagePullPolicy: Always"
    echo "  env:"
    echo "    DATABASE_URL: \"${DB_URL}\""
    echo "    OIDC_ISSUER_URL: \"https://${DNS_KEYCLOAK}/realms/aegis\""
    echo "    OIDC_AUDIENCE: \"backstage,aegis-platform\""
    echo "    OIDC_JWKS_URL: \"https://${DNS_KEYCLOAK}/realms/aegis/protocol/openid-connect/certs\""
    echo "    AEGIS_PROXY_BASE_URL: \"wss://${DNS_PROXY}:8080\""
    echo "    AEGIS_PLATFORM_API_ENDPOINT: \"${DNS_PLATFORM_API}:8081\""
    echo "    AEGIS_PLATFORM_API_GRPC_INSECURE: \"false\""
    [[ -n "${K8S_AGENT_IMAGE_REPO}" ]] && echo "    AEGIS_SPOKE_IMAGE_REPO: \"${K8S_AGENT_IMAGE_REPO}\""
    [[ -n "${K8S_AGENT_IMAGE_TAG_VALUE}" ]] && echo "    AEGIS_SPOKE_IMAGE_TAG: \"${K8S_AGENT_IMAGE_TAG_VALUE}\""
    echo "    AEGIS_SPOKE_OIDC_TOKEN_URL: \"https://${DNS_KEYCLOAK}/realms/aegis/protocol/openid-connect/token\""
    echo "    AEGIS_SPOKE_OIDC_CLIENT_ID: \"spoke-agent\""
    echo "    AEGIS_SPOKE_OIDC_CLIENT_SECRET: \"${SPOKE_OIDC_CLIENT_SECRET}\""
    echo "    AEGIS_SPOKE_OIDC_AUDIENCE: \"aegis-platform\""
    echo "    AEGIS_SPOKE_VALUES_FILE: \"/home/aegis/charts/aegis-spoke/values-cloud-remote.yaml\""
    if [[ -n "${STEP_CA_EXTERNAL_URL:-}" ]]; then
      echo "    AEGIS_CERT_MANAGER_ENABLED: \"true\""
      echo "    AEGIS_STEP_CA_URL: \"${STEP_CA_EXTERNAL_URL}\""
      echo "    AEGIS_CLUSTER_ISSUER_NAME: \"aegis-internal\""
      echo "    AEGIS_STEP_CA_ROOT_CA_FILE: \"/etc/aegis-platform-api/oidc/ca.crt\""
    fi
    [[ -n "${STEP_PROV_KID:-}" ]] && echo "    AEGIS_STEP_PROVISIONER_KID: \"${STEP_PROV_KID}\""
    [[ -n "${STEP_PROV_PASS:-}" ]] && echo "    AEGIS_STEP_PROVISIONER_PASSWORD: \"${STEP_PROV_PASS}\""
    if [[ -n "${K8S_AGENT_IMAGE_REPO}" ]]; then
      echo "    AEGIS_VSCODE_REH_INIT_IMAGE: \"${K8S_AGENT_IMAGE_REPO%/k8s-agent}/vscode-reh-init:latest\""
    fi
    echo "  secrets:"
    echo "    db-password: \"${DB_PASSWORD}\""
    echo "    proxy-jwt-secret: \"${JWT_SECRET}\""
    echo "  tls:"
    echo "    certManager:"
    echo "      enabled: true"
    echo "      dnsNames:"
    echo "        - \"${DNS_PLATFORM_API}\""
    echo "proxy:"
    echo "  publicHost: \"${DNS_PROXY}\""
    echo "  image:"
    [[ -n "${PROXY_IMAGE_REPO}" ]] && echo "    repository: ${PROXY_IMAGE_REPO}"
    [[ -n "${PROXY_IMAGE_TAG_VALUE}" ]] && echo "    tag: \"${PROXY_IMAGE_TAG_VALUE}\""
    echo "  jwtSecret: \"${JWT_SECRET}\""
    echo "  tls:"
    echo "    enabled: true"
    echo "    certManager:"
    echo "      enabled: true"
    echo "      dnsNames:"
    echo "        - \"${DNS_PROXY}\""
    echo "backstage:"
    echo "  enabled: true"
    echo "  replicaCount: 1"
    echo "  service:"
    echo "    type: ClusterIP"
    echo "    port: 7007"
    echo "  ingress:"
    echo "    enabled: true"
    echo "    className: ingress-nginx"
    echo "    annotations:"
    echo "      nginx.ingress.kubernetes.io/ssl-redirect: \"true\""
    echo "    hosts:"
    echo "      - host: \"${DNS_UI}\""
    echo "        paths:"
    echo "          - path: /"
    echo "            pathType: Prefix"
    echo "    tls:"
    echo "      - secretName: ${HELM_RELEASE}-backstage-tls"
    echo "        hosts:"
    echo "          - \"${DNS_UI}\""
    echo "  tls:"
    echo "    certManager:"
    echo "      enabled: true"
    echo "      dnsNames:"
    echo "        - \"${DNS_UI}\""
    echo "  image:"
    [[ -n "${UI_IMAGE_REPO}" ]] && echo "    repository: ${UI_IMAGE_REPO}"
    [[ -n "${UI_IMAGE_TAG_VALUE}" ]] && echo "    tag: \"${UI_IMAGE_TAG_VALUE}\""
    echo "  appConfig:"
    echo "    appBaseUrl: \"https://${DNS_UI}\""
    echo "    backendBaseUrl: \"https://${DNS_UI}\""
    echo "  postgres:"
    echo "    host: \"platform-postgres.${K8S_NAMESPACE}.svc.cluster.local\""
    echo "    port: \"5432\""
    echo "    user: \"aegis_platform\""
    echo "  keycloak:"
    echo "    baseUrl: \"https://${DNS_KEYCLOAK}\""
    echo "    realm: \"aegis\""
    echo "    clientId: \"backstage\""
    echo "    clientSecret:"
    echo "      secretName: ${KEYCLOAK_CLIENT_SECRET_NAME}"
    echo "      key: clientSecret"
    echo "  secrets:"
    echo "    create: true"
    echo "    backendSecret: \"${BACKEND_SECRET}\""
    echo "keycloak:"
    echo "  enabled: true"
    echo "  forceRender: true"
    echo "  namespace: ${K8S_NAMESPACE}"
    echo "  hostname:"
    echo "    hostname: \"https://${DNS_KEYCLOAK}\""
    echo "    admin: \"https://${DNS_KEYCLOAK}\""
    echo "    strict: false"
    echo "  ingress:"
    echo "    enabled: false"
    echo "  customIngress:"
    echo "    enabled: true"
    echo "    className: ingress-nginx"
    echo "    annotations:"
    echo "      nginx.ingress.kubernetes.io/backend-protocol: \"HTTPS\""
    echo "      nginx.ingress.kubernetes.io/ssl-redirect: \"true\""
    echo "  http:"
    echo "    httpEnabled: false"
    echo "  admin:"
    echo "    secret:"
    echo "      name: ${KEYCLOAK_ADMIN_SECRET_NAME}"
    echo "      create: false"
    echo "  database:"
    echo "    secret:"
    echo "      name: ${KEYCLOAK_DB_SECRET_NAME}"
    echo "      create: false"
    echo "  realm:"
    echo "    client:"
    echo "      secret:"
    echo "        name: ${KEYCLOAK_CLIENT_SECRET_NAME}"
    echo "        create: false"
    echo "  tls:"
    echo "    secret:"
    echo "      name: ${KEYCLOAK_TLS_SECRET_NAME}"
    echo "      create: false"
    echo "    certManager:"
    echo "      enabled: true"
    echo "      dnsNames:"
    echo "        - \"${DNS_KEYCLOAK}\""
  } > "${OVERRIDE_FILE}"

  hub_log "helm upgrade --install ${HELM_RELEASE} (timeout 15m)..."
  (
    cd "${CHARTS_DIR}"
    helm upgrade --install "${HELM_RELEASE}" ./aegis-services \
      -f ./aegis-services/values/common.yaml \
      -f ./aegis-services/values/cloud.yaml \
      -f ./aegis-services/values-cloud-generated.yaml \
      -f "${OVERRIDE_FILE}" \
      --namespace "${K8S_NAMESPACE}" --create-namespace \
      --timeout 15m \
      --set "platformApi.env.AEGIS_DISCOVERY_GRPC_ENDPOINT=${DNS_PLATFORM_API}:8081"
  )
  rm -f "${OVERRIDE_FILE}"

  hub_log "Waiting for platform-api + proxy rollouts..."
  kubectl rollout status "deployment/${HELM_RELEASE}-platform-api" -n "${K8S_NAMESPACE}" --timeout=8m
  kubectl rollout status "deployment/${HELM_RELEASE}-proxy" -n "${K8S_NAMESPACE}" --timeout=8m
  hub_log "Waiting for Keycloak..."
  kubectl rollout status "statefulset/${HELM_RELEASE}-keycloak-db" -n "${K8S_NAMESPACE}" --timeout=12m \
    || hub_die "Keycloak Postgres not ready"
  kubectl rollout status "statefulset/${HELM_RELEASE}-keycloak" -n "${K8S_NAMESPACE}" --timeout=12m \
    || hub_die "Keycloak not ready"
  hub_log "Helm hub release ready"
}

phase_migrations_incluster() {
  hub_hr
  hub_log "PHASE app/migrations-incluster"
  if [[ "${SKIP_MIGRATION_PLACEHOLDER:-0}" != "1" ]]; then
    hub_log "SKIP_MIGRATION_PLACEHOLDER!=1 — assuming RDS path already migrated or chart handles it"
    return 0
  fi
  local MIGRATIONS_DIR="${ROOT}/services/platform-api/migrations"
  [[ -f "${MIGRATIONS_DIR}/0001_init.sql" ]] || { hub_log "No migrations; skip"; return 0; }

  kubectl -n "${K8S_NAMESPACE}" wait --for=condition=ready pod -l app.kubernetes.io/name=platform-postgres --timeout=180s 2>/dev/null || true
  local PG_POD
  PG_POD=$(kubectl -n "${K8S_NAMESPACE}" get pods -l app.kubernetes.io/name=platform-postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
  if [[ -z "${PG_POD}" ]]; then
    hub_log "WARN: platform-postgres pod not found; skip migrations"
    return 0
  fi
  local ALL_UP
  ALL_UP=$(mktemp)
  for mig in "${MIGRATIONS_DIR}"/*.sql; do
    awk '/^--[[:space:]]+\+migrate[[:space:]]+Down/{exit} {print}' "$mig" >> "$ALL_UP"
    echo "" >> "$ALL_UP"
  done
  kubectl cp "$ALL_UP" "${K8S_NAMESPACE}/${PG_POD}:/tmp/all_migrations.sql"
  kubectl exec -n "${K8S_NAMESPACE}" "${PG_POD}" -- \
    psql -U aegis_platform -d aegis_platform -v ON_ERROR_STOP=1 -f /tmp/all_migrations.sql
  rm -f "$ALL_UP"
  hub_log "In-cluster migrations applied"
}

phase_lbs_and_dns() {
  hub_hr
  hub_log "PHASE app/lbs-dns"
  local PLATFORM_API_LB="" PROXY_LB="" PUBLIC_INGRESS_LB=""
  wait_for_lb_hostname "platform-api" "svc/${PLATFORM_API_RELEASE_NAME}" "${K8S_NAMESPACE}" PLATFORM_API_LB
  wait_for_lb_dns "${PLATFORM_API_LB}"
  wait_for_lb_hostname "proxy" "svc/${PROXY_RELEASE_NAME}" "${K8S_NAMESPACE}" PROXY_LB
  wait_for_lb_dns "${PROXY_LB}"

  local INGRESS_CONTROLLER_SERVICE
  INGRESS_CONTROLLER_SERVICE=$(kubectl get svc -n "${K8S_NAMESPACE}" \
    -l app.kubernetes.io/component=controller,app.kubernetes.io/name=ingress-nginx \
    -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
  [[ -n "${INGRESS_CONTROLLER_SERVICE}" ]] || hub_die "ingress-nginx controller service not found"
  wait_for_lb_hostname "public ingress" "svc/${INGRESS_CONTROLLER_SERVICE}" "${K8S_NAMESPACE}" PUBLIC_INGRESS_LB
  wait_for_lb_dns "${PUBLIC_INGRESS_LB}"

  export PLATFORM_API_LB PROXY_LB PUBLIC_INGRESS_LB

  if [[ "${SKIP_DNS_UPDATE}" == "1" ]]; then
    hub_log "SKIP_DNS_UPDATE=1 — skipping Cloudflare; use LB hostnames or port-forward:"
    hub_log "  platform-api LB: ${PLATFORM_API_LB}"
    hub_log "  proxy LB:        ${PROXY_LB}"
    hub_log "  ingress LB:      ${PUBLIC_INGRESS_LB}"
  else
    update_cloudflare_cname "${DNS_PLATFORM_API}" "${PLATFORM_API_LB}"
    update_cloudflare_cname "${DNS_PROXY}" "${PROXY_LB}"
    update_cloudflare_cname "${DNS_UI}" "${PUBLIC_INGRESS_LB}"
    update_cloudflare_cname "${DNS_KEYCLOAK}" "${PUBLIC_INGRESS_LB}"
    hub_log "Restarting Backstage after DNS update..."
    kubectl rollout restart "deployment/${HELM_RELEASE}-backstage" -n "${K8S_NAMESPACE}" >/dev/null
    kubectl rollout status "deployment/${HELM_RELEASE}-backstage" -n "${K8S_NAMESPACE}" --timeout=10m
  fi
}

phase_verify_app() {
  hub_hr
  hub_log "PHASE app/verify"
  kubectl -n "${K8S_NAMESPACE}" get pods,svc -o wide

  # Always verify in-cluster health via port-forward (works private + public)
  local svc pf_pid code
  svc="$(kubectl -n "${K8S_NAMESPACE}" get svc -o name | grep platform-api | head -1 | sed 's|service/||')"
  [[ -n "$svc" ]] || hub_die "platform-api service missing"
  kubectl -n "${K8S_NAMESPACE}" port-forward "svc/${svc}" 18080:8080 >/tmp/aegis-hub-pf.log 2>&1 &
  pf_pid=$!
  sleep 4
  code="$(curl -sk -o /tmp/aegis-hub-health.out -w '%{http_code}' "https://127.0.0.1:18080/healthz" 2>/dev/null \
    || curl -s -o /tmp/aegis-hub-health.out -w '%{http_code}' "http://127.0.0.1:18080/healthz" 2>/dev/null \
    || echo 000)"
  kill "$pf_pid" 2>/dev/null || true
  wait "$pf_pid" 2>/dev/null || true
  hub_log "port-forward healthz HTTP ${code}: $(head -c 120 /tmp/aegis-hub-health.out 2>/dev/null || true)"

  if [[ "${SKIP_DNS_UPDATE}" == "1" ]]; then
    hub_log "Private mode: skipping public DNS endpoint checks"
    if [[ "$code" != "200" && "$code" != "204" ]]; then
      hub_log "WARN: healthz not 200 — check logs; some builds only expose gRPC"
      kubectl -n "${K8S_NAMESPACE}" logs -l app.kubernetes.io/component=platform-api --tail=40 2>/dev/null || true
    fi
    return 0
  fi

  # Public path only
  for host in "${DNS_PLATFORM_API}" "${DNS_PROXY}" "${DNS_UI}" "${DNS_KEYCLOAK}"; do
    hub_log "Checking public DNS ${host}..."
    for _ in $(seq 1 30); do
      [[ -n "$(dig @1.1.1.1 +short "${host}" 2>/dev/null)" ]] && break
      sleep 2
    done
  done
  curl -fsS --max-time 20 "https://${DNS_KEYCLOAK}/realms/aegis/.well-known/openid-configuration" >/dev/null \
    || hub_log "WARN: Keycloak discovery not ready yet"
  hub_log "Public verification finished"
}

# -----------------------------------------------------------------------------
main() {
  hub_log "hub-app deploy-app start profile=${AWS_PROFILE} skip_dns=${SKIP_DNS_UPDATE} reuse=${REUSE_EXISTING}"
  phase_preflight
  phase_kubeconfig
  phase_values
  phase_secrets
  phase_db_url
  phase_pki
  phase_crds
  phase_helm
  phase_migrations_incluster
  phase_lbs_and_dns
  phase_verify_app
  hub_hr
  hub_log "APP DEPLOY COMPLETE"
  hub_log "  kubectl get pods -n ${K8S_NAMESPACE}"
  hub_log "  CA: ${CA_BUNDLE}"
  if [[ "${SKIP_DNS_UPDATE}" == "1" ]]; then
    hub_log "  Private: use LB hostnames above or port-forward for demo"
  else
    hub_log "  UI https://${DNS_UI}"
  fi
}

main "$@"
