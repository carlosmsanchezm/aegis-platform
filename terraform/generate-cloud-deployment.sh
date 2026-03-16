#!/bin/bash
# Generate Helm values and optionally deploy to Kubernetes
#
# For complete deployment documentation, see: DEPLOYMENT_AND_TESTING_GUIDE.md

set -e

NON_INTERACTIVE=0
K8S_NAMESPACE=${K8S_NAMESPACE:-aegis-system}
HELM_RELEASE=${HELM_RELEASE:-aegis}
SKIP_DNS_UPDATE=${SKIP_DNS_UPDATE:-${SKIP_ROUTE53_UPDATE:-0}}

# Helper to split an image reference into repository and tag components
parse_image_ref() {
  local ref="$1"
  local repo=""
  local tag="$ref"
  if [[ -n "$ref" && "$ref" == *:* ]]; then
    repo="${ref%:*}"
    tag="${ref##*:}"
  fi
  printf '%s|%s' "$repo" "$tag"
}

require_image_ref() {
  local var_name="$1"
  local value="$2"
  if [[ -z "${value}" ]]; then
    echo "❌ ${var_name} must be set to a prebuilt image reference (for example: 123456789012.dkr.ecr.us-east-1.amazonaws.com/aegis/ui:abc1234)" >&2
    exit 1
  fi
  if [[ "${value}" != *:* ]]; then
    echo "❌ ${var_name} must include an explicit tag in repo:tag form: ${value}" >&2
    exit 1
  fi
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

verify_ecr_image() {
  local image_ref="$1"
  local repo tag repo_name
  IFS='|' read -r repo tag <<< "$(parse_image_ref "${image_ref}")"
  repo_name="$(ecr_repo_name "${repo}")"
  if [[ -z "${repo_name}" || -z "${tag}" ]]; then
    echo "❌ Unable to validate ECR image reference: ${image_ref}" >&2
    exit 1
  fi
  if ! aws ecr describe-images \
    --repository-name "${repo_name}" \
    --image-ids imageTag="${tag}" \
    --region "${AWS_REGION}" \
    --profile "${AWS_PROFILE}" >/dev/null 2>&1; then
    echo "❌ ECR image not found: ${image_ref}" >&2
    exit 1
  fi
}

cloudflare_record_id() {
  local record_name="$1"
  curl -sS "https://api.cloudflare.com/client/v4/zones/${CF_ZONE_ID}/dns_records?name=${record_name}&type=CNAME" \
    -H "Authorization: Bearer ${CF_API_TOKEN}" | jq -r '.result[0].id // empty'
}

update_cloudflare_cname() {
  local record_name="$1"
  local lb_host="$2"
  local record_id
  record_id="$(cloudflare_record_id "${record_name}")"
  if [[ -z "${record_id}" ]]; then
    echo "❌ Cloudflare record not found: ${record_name}" >&2
    exit 1
  fi
  curl -sS -X PATCH "https://api.cloudflare.com/client/v4/zones/${CF_ZONE_ID}/dns_records/${record_id}" \
    -H "Authorization: Bearer ${CF_API_TOKEN}" \
    -H "Content-Type: application/json" \
    --data "{\"content\":\"${lb_host}\"}" >/dev/null
}

wait_for_lb_hostname() {
  local description="$1"
  local resource="$2"
  local namespace="$3"
  local outvar="$4"
  local value=""
  echo "   Waiting for ${description} Load Balancer..."
  for _ in {1..60}; do
    value=$(kubectl get "${resource}" -n "${namespace}" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
    if [[ -n "${value}" ]]; then
      echo "   ✅ ${description} Load Balancer ready: ${value}"
      printf -v "${outvar}" '%s' "${value}"
      return 0
    fi
    echo -n "."
    sleep 2
  done
  echo ""
  echo "❌ Timed out waiting for ${description} Load Balancer" >&2
  exit 1
}

wait_for_lb_dns() {
  local hostname="$1"
  echo "   Waiting for DNS propagation for ${hostname}..."
  for _ in {1..30}; do
    if dig +short "${hostname}" | grep -q '^[0-9]'; then
      echo "   ✅ DNS resolved: $(dig +short "${hostname}" | head -1)"
      return 0
    fi
    echo -n "."
    sleep 2
  done
  echo ""
  echo "❌ Timed out waiting for DNS resolution for ${hostname}" >&2
  exit 1
}

wait_for_public_dns() {
  local hostname="$1"
  echo "   Waiting for public DNS propagation for ${hostname}..."
  for _ in {1..30}; do
    if [[ -n "$(dig @1.1.1.1 +short "${hostname}")" ]]; then
      echo "   ✅ Public DNS resolved: $(dig @1.1.1.1 +short "${hostname}" | head -1)"
      return 0
    fi
    echo -n "."
    sleep 2
  done
  echo ""
  echo "❌ Timed out waiting for public DNS resolution for ${hostname}" >&2
  exit 1
}

# Derive common resource names based on the Helm release (aligned with Helm's truncation logic)
RELEASE_BASENAME=$(printf '%s' "${HELM_RELEASE}" | cut -c1-63)
RELEASE_BASENAME=${RELEASE_BASENAME%-}
PLATFORM_API_RELEASE_NAME="${RELEASE_BASENAME}-platform-api"
PROXY_RELEASE_NAME="${RELEASE_BASENAME}-proxy"
SPOKE_HELM_RELEASE=${SPOKE_HELM_RELEASE:-${HELM_RELEASE}-spoke}
SPOKE_NAMESPACE=${SPOKE_NAMESPACE:-${K8S_NAMESPACE}}
KEYCLOAK_SERVICE_NAME="${RELEASE_BASENAME}-keycloak-service"
KEYCLOAK_INTERNAL_HOST="${KEYCLOAK_SERVICE_NAME}.${K8S_NAMESPACE}.svc.cluster.local"

usage() {
cat <<'EOF'
Usage: ./generate-cloud-deployment.sh [--non-interactive]

Options:
  --non-interactive  Run without interactive prompts for CI/CD
  -h, --help         Show this help message

All cloud deployments terminate TLS inside the platform-api and proxy pods; no additional flags are required.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tls)
      echo "⚠️  --tls is deprecated; TLS is always enforced for cloud deployments."
      shift
      ;;
    --non-interactive)
      NON_INTERACTIVE=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 1
      ;;
  esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/../charts"
AWS_PROFILE="${AWS_PROFILE:-aegis-new}"
AWS_REGION="${AWS_REGION:-us-east-1}"
PLATFORM_API_IMAGE_TAG="${PLATFORM_API_IMAGE_TAG:-}"
PROXY_IMAGE_TAG="${PROXY_IMAGE_TAG:-}"
K8S_AGENT_IMAGE_TAG="${K8S_AGENT_IMAGE_TAG:-}"
UI_IMAGE_TAG="${UI_IMAGE_TAG:-}"
TLS_CERT_PATH=/tmp/proxy-cert.pem
TLS_KEY_PATH=/tmp/proxy-key.pem
TLS_CA_CERT_PATH=/tmp/proxy-ca.pem
TLS_CA_KEY_PATH=/tmp/proxy-ca-key.pem
TLS_CERT_CSR_PATH=/tmp/proxy-cert.csr
TLS_CERT_EXT_PATH=/tmp/proxy-cert-ext.cnf
TLS_CERT_CHAIN_PATH=/tmp/proxy-cert-chain.pem
KEYCLOAK_CERT_PATH=/tmp/keycloak-cert.pem
KEYCLOAK_KEY_PATH=/tmp/keycloak-key.pem
KEYCLOAK_CERT_CSR_PATH=/tmp/keycloak-cert.csr
KEYCLOAK_CERT_EXT_PATH=/tmp/keycloak-cert-ext.cnf
KEYCLOAK_CERT_CHAIN_PATH=/tmp/keycloak-cert-chain.pem
PLATFORM_API_SERVICE_ACCOUNT=${PLATFORM_API_SERVICE_ACCOUNT:-aegis-platform-api}
LEGACY_UI_RELEASE=${LEGACY_UI_RELEASE:-aegis-ui}
REMOVE_LEGACY_UI_RELEASE=${REMOVE_LEGACY_UI_RELEASE:-1}
CA_BUNDLE="${HOME}/aegis-platform-api-ca.crt"
OVERRIDE_FILE=""
TLS_OVERRIDE_FILE=""
MIGRATIONS_UP_FILE=""
KEYCLOAK_ADMIN_USERNAME=${KEYCLOAK_ADMIN_USERNAME:-admin}
KEYCLOAK_ADMIN_PASSWORD=${KEYCLOAK_ADMIN_PASSWORD:-PreviewAdmin123!}
KEYCLOAK_DB_USERNAME=${KEYCLOAK_DB_USERNAME:-keycloak}
KEYCLOAK_DB_PASSWORD=${KEYCLOAK_DB_PASSWORD:-PreviewKeycloakDb123!}
KEYCLOAK_BACKSTAGE_CLIENT_SECRET=${KEYCLOAK_BACKSTAGE_CLIENT_SECRET:-preview-backstage-client-secret}
KEYCLOAK_ADMIN_SECRET_NAME=${KEYCLOAK_ADMIN_SECRET_NAME:-keycloak-admin-secret}
KEYCLOAK_DB_SECRET_NAME=${KEYCLOAK_DB_SECRET_NAME:-keycloak-db-secret}
KEYCLOAK_CLIENT_SECRET_NAME=${KEYCLOAK_CLIENT_SECRET_NAME:-keycloak-backstage-client-secret}
KEYCLOAK_TLS_SECRET_NAME=${KEYCLOAK_TLS_SECRET_NAME:-keycloak-tls}
KEYCLOAK_INTERNAL_PORT=${KEYCLOAK_INTERNAL_PORT:-8443}
SPOKE_OIDC_CLIENT_SECRET=${SPOKE_OIDC_CLIENT_SECRET:-$(cd "${SCRIPT_DIR}" && terraform output -raw spoke_oidc_client_secret 2>/dev/null || echo "rEC99sBBWQAbRgg0xRQFBsMC8rt6pZOB")}
PLATFORM_API_PUBLIC_HOST=${PLATFORM_API_PUBLIC_HOST:-platform-api.aegis-platform.tech}
KEYCLOAK_PUBLIC_HOST=${KEYCLOAK_PUBLIC_HOST:-keycloak.aegis-platform.tech}
PROXY_PUBLIC_HOST=${PROXY_PUBLIC_HOST:-proxy.aegis-platform.tech}
UI_PUBLIC_HOST=${UI_PUBLIC_HOST:-ui.aegis-platform.tech}
CF_ZONE_ID=""
CF_API_TOKEN="${CLOUDFLARE_API_TOKEN:-}"

IFS='|' read -r PLATFORM_API_IMAGE_REPO PLATFORM_API_IMAGE_TAG_VALUE <<< "$(parse_image_ref "${PLATFORM_API_IMAGE_TAG}")"
IFS='|' read -r PROXY_IMAGE_REPO PROXY_IMAGE_TAG_VALUE <<< "$(parse_image_ref "${PROXY_IMAGE_TAG}")"
IFS='|' read -r K8S_AGENT_IMAGE_REPO K8S_AGENT_IMAGE_TAG_VALUE <<< "$(parse_image_ref "${K8S_AGENT_IMAGE_TAG}")"
IFS='|' read -r UI_IMAGE_REPO UI_IMAGE_TAG_VALUE <<< "$(parse_image_ref "${UI_IMAGE_TAG}")"

require_image_ref "PLATFORM_API_IMAGE_TAG" "${PLATFORM_API_IMAGE_TAG}"
require_image_ref "PROXY_IMAGE_TAG" "${PROXY_IMAGE_TAG}"
require_image_ref "K8S_AGENT_IMAGE_TAG" "${K8S_AGENT_IMAGE_TAG}"
require_image_ref "UI_IMAGE_TAG" "${UI_IMAGE_TAG}"

cleanup() {
  rm -f "${OVERRIDE_FILE}" "${TLS_OVERRIDE_FILE}" \
    "${TLS_CERT_CSR_PATH}" "${TLS_CERT_EXT_PATH}" "${TLS_CA_KEY_PATH}" "${TLS_CA_CERT_PATH}" "${TLS_CERT_CHAIN_PATH}" "${TLS_CA_CERT_PATH}.srl" \
    "${KEYCLOAK_CERT_CSR_PATH}" "${KEYCLOAK_CERT_EXT_PATH}" "${KEYCLOAK_KEY_PATH}" "${KEYCLOAK_CERT_PATH}" "${KEYCLOAK_CERT_CHAIN_PATH}" \
    "${MIGRATIONS_UP_FILE}"
}
trap cleanup EXIT

echo "🔐 Enforcing TLS for platform-api and proxy"
rm -f "${TLS_CERT_PATH}" "${TLS_KEY_PATH}" "${TLS_CA_CERT_PATH}" "${TLS_CA_KEY_PATH}" "${TLS_CERT_CSR_PATH}" "${TLS_CERT_CHAIN_PATH}" "${TLS_CA_CERT_PATH}.srl" \
  "${KEYCLOAK_CERT_PATH}" "${KEYCLOAK_KEY_PATH}" "${KEYCLOAK_CERT_CSR_PATH}" "${KEYCLOAK_CERT_CHAIN_PATH}" "${KEYCLOAK_CERT_EXT_PATH}"

echo "🚀 Generating Helm values from Terraform outputs..."
echo "📖 See DEPLOYMENT_AND_TESTING_GUIDE.md for complete deployment guide"
echo ""

# Run Terraform commands from the Terraform root regardless of caller cwd.
cd "${SCRIPT_DIR}"

# Check if terraform is initialized
if [ ! -d "${SCRIPT_DIR}/.terraform" ]; then
    echo "❌ Terraform not initialized. Run 'terraform init' first."
    exit 1
fi

# Check if terraform state exists
if [ ! -f "${SCRIPT_DIR}/terraform.tfstate" ]; then
    echo "❌ No Terraform state found. Run 'terraform apply' first."
    exit 1
fi

if [[ -n "${SKIP_ROUTE53_UPDATE:-}" ]]; then
  echo "⚠️  SKIP_ROUTE53_UPDATE is deprecated; use SKIP_DNS_UPDATE instead."
fi

CF_ZONE_ID=$(terraform output -raw cloudflare_zone_id 2>/dev/null || echo "")
if [[ "${SKIP_DNS_UPDATE}" != "1" ]]; then
  if [[ -z "${CF_API_TOKEN}" ]]; then
    echo "❌ CLOUDFLARE_API_TOKEN must be set for cloud deployment." >&2
    exit 1
  fi
  if [[ -z "${CF_ZONE_ID}" ]]; then
    echo "❌ Cloudflare zone output not available. Run 'terraform apply' first." >&2
    exit 1
  fi
  for record in \
    "${PLATFORM_API_PUBLIC_HOST}" \
    "${KEYCLOAK_PUBLIC_HOST}" \
    "${PROXY_PUBLIC_HOST}" \
    "${UI_PUBLIC_HOST}"; do
    if [[ -z "$(cloudflare_record_id "${record}")" ]]; then
      echo "❌ Required Cloudflare record not found: ${record}" >&2
      exit 1
    fi
  done
fi

echo "🔎 Verifying required images exist in ECR..."
verify_ecr_image "${PLATFORM_API_IMAGE_TAG}"
verify_ecr_image "${PROXY_IMAGE_TAG}"
verify_ecr_image "${K8S_AGENT_IMAGE_TAG}"
verify_ecr_image "${UI_IMAGE_TAG}"
echo "   ✅ All required images found"

# Generate aegis-services values
echo "📝 Generating aegis-services values..."
terraform output -raw helm_values_aegis_services > "${OUTPUT_DIR}/aegis-services/values-cloud-generated.yaml"
echo "   ✅ Created: ${OUTPUT_DIR}/aegis-services/values-cloud-generated.yaml"

# Generate aegis-spoke values
echo "📝 Generating aegis-spoke values..."
terraform output -raw helm_values_aegis_spoke > "${OUTPUT_DIR}/aegis-spoke/values-cloud-generated.yaml"
echo "   ✅ Created: ${OUTPUT_DIR}/aegis-spoke/values-cloud-generated.yaml"

# Display secret creation commands
echo ""
echo "🔐 Kubernetes Secrets Creation Commands:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
terraform output -raw k8s_secret_commands
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo ""
echo "📊 Quick Reference:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Cluster Name:     $(terraform output -raw cluster_name)"
echo "Region:           $(terraform output -raw aws_region)"
echo "ECR Registry:     $(terraform output -raw ecr_registry_url)"
echo "RDS Endpoint:     $(terraform output -raw rds_endpoint)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
if [[ $NON_INTERACTIVE -eq 0 ]]; then
    echo ""
    echo "🚀 Deploy to Kubernetes?"
    echo ""
    read -p "Do you want to deploy now? (y/n): " -n 1 -r
    echo ""

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo ""
        echo "⏭️  Skipping deployment. You can deploy later with:"
        echo ""
        echo "   # Configure kubectl"
        echo "   $(terraform output -raw kubectl_config_command)"
        echo ""
        echo "   # Create secrets"
        echo "   kubectl create secret generic aegis-platform-secrets \\"
        echo "     --from-literal=db-password=\"\$(terraform output -raw db_password_secret_value)\" \\"
        echo "     --from-literal=proxy-jwt-secret=\"\$(terraform output -raw jwt_secret_value)\" \\"
        echo "     --namespace ${K8S_NAMESPACE} --create-namespace"
        echo ""
        echo "   # Deploy"
        echo "   cd ${OUTPUT_DIR}"
        echo "   helm upgrade --install ${HELM_RELEASE} ./aegis-services \\"
        echo "     -f ./aegis-services/values/common.yaml \\"
        echo "     -f ./aegis-services/values/cloud.yaml \\"
        echo "     -f ./aegis-services/values-cloud-generated.yaml \\"
        echo "     -f <your-overrides.yaml> \\"
        echo "     --namespace ${K8S_NAMESPACE} --create-namespace"
        echo ""
        echo "   # Example overrides file (include secrets and image tags):"
        echo "   cat > overrides.yaml <<'EOF'"
        echo "   platformApi:"
        echo "     image:"
        echo "       tag: ${PLATFORM_API_IMAGE_TAG}"
        echo "     env:"
        echo "       DATABASE_URL: ${DB_URL}"
        echo "     secrets:"
        echo "       db-password: \$(terraform output -raw db_password_secret_value)"
        echo "       proxy-jwt-secret: \$(terraform output -raw jwt_secret_value)"
        echo "   proxy:"
        echo "     image:"
        echo "       tag: ${PROXY_IMAGE_TAG}"
        echo "     jwtSecret: \$(terraform output -raw jwt_secret_value)"
        echo "   EOF"
        echo ""
        echo "✨ Done!"
        echo ""
        echo "ℹ️  TLS is enforced automatically; rerun this script anytime you want to regenerate certificates"
        exit 0
    fi
else
    echo ""
    echo "🤖 Non-interactive mode enabled; proceeding with automated deployment"
fi

echo ""
echo "📋 Deployment Steps (namespace: ${K8S_NAMESPACE}, release: ${HELM_RELEASE}):"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Step 1: Configure kubectl
echo ""
echo "1️⃣  Configuring kubectl..."
KUBECTL_CMD=$(terraform output -raw kubectl_config_command)
echo "   Running: ${KUBECTL_CMD}"
eval "${KUBECTL_CMD}"
echo "   ✅ kubectl configured"

# Surface the fully qualified context so operators can switch back easily
AWS_REGION=${AWS_REGION:-$(terraform output -raw aws_region 2>/dev/null || echo "")}
AWS_ACCOUNT_ID=$(terraform output -raw aws_account_id 2>/dev/null || echo "")
CLUSTER_NAME=$(terraform output -raw cluster_name 2>/dev/null || echo "")
if [[ -n "${AWS_REGION}" && -n "${AWS_ACCOUNT_ID}" && -n "${CLUSTER_NAME}" ]]; then
  KUBE_CONTEXT="arn:aws:eks:${AWS_REGION}:${AWS_ACCOUNT_ID}:cluster/${CLUSTER_NAME}"
  echo "   ℹ️  Switch kubectl context with:"
  echo "      kubectl config use-context ${KUBE_CONTEXT}"
fi

# Step 2: Create namespace and secrets
echo ""
echo "2️⃣  Creating namespace and Kubernetes secrets..."
DB_PASSWORD=$(terraform output -raw db_password_secret_value)
JWT_SECRET=$(terraform output -raw jwt_secret_value)

# Create namespace if it doesn't exist
kubectl create namespace "${K8S_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

# Create or update secrets
kubectl create secret generic aegis-platform-secrets \
  --from-literal=db-password="${DB_PASSWORD}" \
  --from-literal=proxy-jwt-secret="${JWT_SECRET}" \
  --namespace "${K8S_NAMESPACE}" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "   ℹ️  Ensuring Keycloak secrets exist"
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

echo "   ℹ️  Resetting Keycloak PVCs to pick up storage class overrides"
kubectl delete pvc -n "${K8S_NAMESPACE}" -l app.kubernetes.io/component=keycloak-postgres --ignore-not-found --wait=false >/dev/null 2>&1 || true
kubectl wait --for=delete pvc -n "${K8S_NAMESPACE}" -l app.kubernetes.io/component=keycloak-postgres --timeout=120s >/dev/null 2>&1 || echo "   ⚠️  Keycloak PVC deletion still in progress; continuing"
echo "   ℹ️  Resetting Keycloak StatefulSets to allow PVC recreation"
kubectl delete statefulset "${HELM_RELEASE}-keycloak" -n "${K8S_NAMESPACE}" --ignore-not-found --wait=false >/dev/null 2>&1 || true
kubectl delete statefulset "${HELM_RELEASE}-keycloak-db" -n "${K8S_NAMESPACE}" --ignore-not-found --wait=false >/dev/null 2>&1 || true
kubectl wait --for=delete "statefulset/${HELM_RELEASE}-keycloak" -n "${K8S_NAMESPACE}" --timeout=120s >/dev/null 2>&1 || echo "   ⚠️  Keycloak StatefulSet deletion still in progress; continuing"
kubectl wait --for=delete "statefulset/${HELM_RELEASE}-keycloak-db" -n "${K8S_NAMESPACE}" --timeout=120s >/dev/null 2>&1 || echo "   ⚠️  Keycloak DB StatefulSet deletion still in progress; continuing"

echo "   ℹ️  Skipping aegis-kubeconfigs secret (managed by Helm)"

echo "   ✅ Namespace and secrets created"

# Ensure the platform API ServiceAccount exists with Helm ownership metadata so
# pre-deploy jobs (like migrations) can run before Helm installs the chart.
echo "   ℹ️  Ensuring ServiceAccount ${PLATFORM_API_SERVICE_ACCOUNT} exists"
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
echo "   ✅ ServiceAccount ready"

# Step 3: Run database migrations manually (to avoid public image pull issues)
echo ""
echo "3️⃣  Running database migrations..."

# URL-encode the DB password for DATABASE_URL
DB_PASSWORD_RAW=$(terraform output -raw db_password_secret_value)
DB_PASSWORD_ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.quote('${DB_PASSWORD_RAW}', safe=''))")
DB_ENDPOINT=$(terraform output -raw rds_endpoint 2>/dev/null || echo "")
DB_NAME=$(terraform output -raw rds_database_name 2>/dev/null || echo "aegis_platform")
DB_USER_FALLBACK="aegis_platform"
if terraform output -raw rds_username >/tmp/rds_user 2>/dev/null; then
  DB_USER=$(cat /tmp/rds_user)
  rm -f /tmp/rds_user
else
  DB_USER=${DB_USER_FALLBACK}
fi
DB_SSLMODE="require"
USING_RDS=1
if [[ -n "${DB_ENDPOINT}" ]]; then
  if [[ "${DB_ENDPOINT}" == *:* ]]; then
    DB_HOST=${DB_ENDPOINT%%:*}
    DB_PORT=${DB_ENDPOINT##*:}
  else
    DB_HOST=${DB_ENDPOINT}
    DB_PORT=$(terraform output -raw rds_port 2>/dev/null || echo "5432")
  fi
else
  USING_RDS=0
  DB_HOST="platform-postgres.${K8S_NAMESPACE}.svc.cluster.local"
  DB_PORT="5432"
  DB_SSLMODE="disable"
  SKIP_MIGRATION_PLACEHOLDER=1
  echo "   ℹ️  No RDS endpoint configured; using in-cluster Postgres"
fi
DB_URL="postgres://${DB_USER}:${DB_PASSWORD_ENCODED}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

if [[ -z "${AWS_REGION:-}" ]]; then
  AWS_REGION=$(terraform output -raw aws_region 2>/dev/null || echo "")
fi

if [[ -z "${AWS_REGION}" ]]; then
  echo "❌ Unable to determine AWS region for migrations"
  exit 1
fi

RDS_CA_BUNDLE_URL="https://truststore.pki.rds.amazonaws.com/${AWS_REGION}/${AWS_REGION}-bundle.pem"
JOB_ANNOTATIONS_BLOCK=""
POD_ANNOTATIONS_BLOCK=""
if [[ "${USING_RDS}" == "1" ]]; then
  NODE_SECURITY_GROUP="$(terraform output -raw node_security_group_id 2>/dev/null || echo "")"
  CLUSTER_SECURITY_GROUP="$(terraform output -raw cluster_security_group_id 2>/dev/null || echo "")"
  EKS_CLUSTER_MANAGED_SECURITY_GROUP="$(terraform output -raw eks_cluster_security_group_id 2>/dev/null || echo "")"
  SECURITY_GROUPS_LIST=()
  if [[ -n "${NODE_SECURITY_GROUP}" ]]; then
    SECURITY_GROUPS_LIST+=("${NODE_SECURITY_GROUP}")
  fi
  if [[ -n "${CLUSTER_SECURITY_GROUP}" ]]; then
    SECURITY_GROUPS_LIST+=("${CLUSTER_SECURITY_GROUP}")
  fi
  if [[ -n "${EKS_CLUSTER_MANAGED_SECURITY_GROUP}" ]]; then
    SECURITY_GROUPS_LIST+=("${EKS_CLUSTER_MANAGED_SECURITY_GROUP}")
  fi

  if (( ${#SECURITY_GROUPS_LIST[@]} > 0 )); then
    SECURITY_GROUPS=$(IFS=','; echo "${SECURITY_GROUPS_LIST[*]}")
    JOB_ANNOTATIONS_BLOCK=$(cat <<EOF
  annotations:
    vpc.amazonaws.com/security-groups: "${SECURITY_GROUPS}"
EOF
)
    POD_ANNOTATIONS_BLOCK=$(cat <<EOF
      annotations:
        vpc.amazonaws.com/security-groups: "${SECURITY_GROUPS}"
EOF
)
  fi
fi


if [[ "${SKIP_MIGRATION_PLACEHOLDER:-0}" == "1" ]]; then
  echo "   ℹ️  Using in-cluster Postgres — migrations will run after Helm install (Step 6b)"
else
# Run migrations using an in-cluster Job so RDS schema exists before tests
echo "   ⚙️  Applying database schema via Kubernetes Job"
MIGRATION_CONFIGMAP="${HELM_RELEASE}-migrations"
MIGRATION_JOB="${HELM_RELEASE}-migrate"
MIGRATIONS_DIR="${SCRIPT_DIR}/../services/platform-api/migrations"

if [ ! -f "${MIGRATIONS_DIR}/0001_init.sql" ]; then
  echo "❌ Migration file not found at ${MIGRATIONS_DIR}/0001_init.sql"
  exit 1
fi
MIGRATIONS_UP_FILE=$(mktemp)
awk '/^--[[:space:]]+\+migrate[[:space:]]+Down/{exit} {print}' "${MIGRATIONS_DIR}/0001_init.sql" > "${MIGRATIONS_UP_FILE}"
if [[ ! -s "${MIGRATIONS_UP_FILE}" ]]; then
  echo "❌ Failed to extract migration up statements"
  exit 1
fi

kubectl -n "${K8S_NAMESPACE}" create configmap "${MIGRATION_CONFIGMAP}" \
  --from-file=0001_init.sql="${MIGRATIONS_UP_FILE}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl delete job "${MIGRATION_JOB}" -n "${K8S_NAMESPACE}" --ignore-not-found >/dev/null 2>&1 || true

cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: ${MIGRATION_JOB}
  namespace: ${K8S_NAMESPACE}
  labels:
    app.kubernetes.io/name: aegis-services
    app.kubernetes.io/instance: ${HELM_RELEASE}
    app.kubernetes.io/component: platform-api
    app: aegis-platform-api
${JOB_ANNOTATIONS_BLOCK}
spec:
  ttlSecondsAfterFinished: 600
  backoffLimit: 1
  template:
    metadata:
      labels:
        app.kubernetes.io/name: aegis-services
        app.kubernetes.io/instance: ${HELM_RELEASE}
        app.kubernetes.io/component: platform-api
        app: aegis-platform-api
${POD_ANNOTATIONS_BLOCK}
    spec:
      serviceAccountName: "${PLATFORM_API_SERVICE_ACCOUNT}"
      restartPolicy: Never
      containers:
      - name: migrate
        image: postgres:16-alpine
        env:
        - name: PGPASSWORD
          valueFrom:
            secretKeyRef:
              name: aegis-platform-secrets
              key: db-password
        - name: DB_HOST
          value: "${DB_HOST}"
        - name: DB_PORT
          value: "${DB_PORT}"
        - name: DB_NAME
          value: "${DB_NAME}"
        - name: DB_USER
          value: "${DB_USER}"
        - name: AWS_REGION
          value: "${AWS_REGION}"
        - name: RDS_CA_BUNDLE_URL
          value: "${RDS_CA_BUNDLE_URL}"
        command:
        - /bin/sh
        - -c
        - |
          set -euo pipefail
          apk add --no-cache ca-certificates curl >/dev/null 2>&1
          if [ -z "${RDS_CA_BUNDLE_URL:-}" ]; then
            echo "Missing RDS_CA_BUNDLE_URL" >&2
            exit 1
          fi
          curl -fsSL "${RDS_CA_BUNDLE_URL}" -o /tmp/rds.pem
          psql "host=${DB_HOST} port=${DB_PORT} sslmode=verify-full sslrootcert=/tmp/rds.pem user=${DB_USER} dbname=${DB_NAME}" -v ON_ERROR_STOP=1 -f /migrations/0001_init.sql
        volumeMounts:
        - name: migrations
          mountPath: /migrations
      volumes:
      - name: migrations
        configMap:
          name: ${MIGRATION_CONFIGMAP}
EOF

if ! kubectl -n "${K8S_NAMESPACE}" wait --for=condition=complete "job/${MIGRATION_JOB}" --timeout=5m; then
  echo "❌ Migration job failed. Logs:"
  kubectl logs job/"${MIGRATION_JOB}" -n "${K8S_NAMESPACE}" || true
  echo "ℹ️  Leaving ${MIGRATION_JOB} and configmap ${MIGRATION_CONFIGMAP} in place for troubleshooting"
  exit 1
fi

kubectl logs job/"${MIGRATION_JOB}" -n "${K8S_NAMESPACE}" || true
kubectl delete job "${MIGRATION_JOB}" -n "${K8S_NAMESPACE}" --ignore-not-found >/dev/null 2>&1 || true
kubectl delete configmap "${MIGRATION_CONFIGMAP}" -n "${K8S_NAMESPACE}" --ignore-not-found >/dev/null 2>&1 || true
echo "   ✅ Migrations applied"
fi

# Optional namespace reset (cleans out previous preview if requested)
RESET_NAMESPACE="${RESET_NAMESPACE:-0}"
REUSE_EXISTING_DEPLOYMENT="${REUSE_EXISTING_DEPLOYMENT:-0}"
if kubectl get namespace "${K8S_NAMESPACE}" >/dev/null 2>&1; then
  if [[ "${RESET_NAMESPACE}" == "1" || "${REUSE_EXISTING_DEPLOYMENT}" != "1" ]]; then
    echo ""
    echo "4️⃣  Cleaning previous ${K8S_NAMESPACE} namespace"
    helm uninstall "${HELM_RELEASE}" -n "${K8S_NAMESPACE}" >/dev/null 2>&1 || true
    helm uninstall "${SPOKE_HELM_RELEASE}" -n "${K8S_NAMESPACE}" >/dev/null 2>&1 || true
    kubectl delete namespace "${K8S_NAMESPACE}" --ignore-not-found --wait >/dev/null 2>&1 || true
    echo "   🔁 Recreating namespace ${K8S_NAMESPACE} and required secrets"
    kubectl create namespace "${K8S_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -
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
  else
    echo ""
    echo "4️⃣  Reusing existing namespace ${K8S_NAMESPACE} (skipping reset)"
  fi
fi

if helm status "${LEGACY_UI_RELEASE}" -n "${K8S_NAMESPACE}" >/dev/null 2>&1; then
  if [[ "${REMOVE_LEGACY_UI_RELEASE}" == "1" ]]; then
    echo "   ℹ️  Removing legacy UI release ${LEGACY_UI_RELEASE} to avoid conflicting cloud UI state"
    helm uninstall "${LEGACY_UI_RELEASE}" -n "${K8S_NAMESPACE}" >/dev/null 2>&1 || true
  else
    echo "❌ Legacy UI release ${LEGACY_UI_RELEASE} exists in namespace ${K8S_NAMESPACE}. Remove it or set REMOVE_LEGACY_UI_RELEASE=1." >&2
    exit 1
  fi
fi

# Step 4: Ensure internal PKI (step-ca + cert-manager) and refresh CA bundle
echo ""
echo "4️⃣  Ensuring internal PKI (step-ca + cert-manager)..."

# Get canonical public DNS hostnames (used for certificates and deployment verification)
DNS_PLATFORM_API=$(terraform output -raw dns_platform_api 2>/dev/null || echo "${PLATFORM_API_PUBLIC_HOST}")
DNS_KEYCLOAK=$(terraform output -raw dns_keycloak 2>/dev/null || echo "${KEYCLOAK_PUBLIC_HOST}")
DNS_PROXY=$(terraform output -raw dns_proxy 2>/dev/null || echo "${PROXY_PUBLIC_HOST}")
DNS_UI=$(terraform output -raw dns_ui 2>/dev/null || echo "${UI_PUBLIC_HOST}")

echo "   📋 DNS hostnames:"
echo "      Platform API:      ${DNS_PLATFORM_API}"
echo "      Keycloak:          ${DNS_KEYCLOAK}"
echo "      Proxy:             ${DNS_PROXY}"
echo "      UI:                ${DNS_UI}"

PKI_NAMESPACE=${PKI_NAMESPACE:-aegis-pki}
CERT_MANAGER_NAMESPACE=${CERT_MANAGER_NAMESPACE:-cert-manager}

PKI_SCRIPT="${SCRIPT_DIR}/../scripts/install-internal-pki.sh"
if [[ ! -x "${PKI_SCRIPT}" ]]; then
  echo "❌ Internal PKI installer not found or not executable: ${PKI_SCRIPT}" >&2
  exit 1
fi

TRUST_BUNDLE_NAMESPACES="${K8S_NAMESPACE},${SPOKE_NAMESPACE}" \
PKI_NAMESPACE="${PKI_NAMESPACE}" \
CERT_MANAGER_NAMESPACE="${CERT_MANAGER_NAMESPACE}" \
STEP_CA_DB_PERSISTENT=false \
STEP_CA_REINSTALL_ON_MISMATCH=true \
  "${PKI_SCRIPT}"

mkdir -p "$(dirname "${CA_BUNDLE}")"
kubectl get secret aegis-trust-bundle -n "${K8S_NAMESPACE}" -o "jsonpath={.data.ca\\.crt}" | base64 --decode > "${CA_BUNDLE}"
echo "   ✅ Updated CA bundle: ${CA_BUNDLE}"

# Step 5: Ensure CRDs are present before Helm upgrades
echo ""
echo "5️⃣  Applying CRDs (k8s-agent) before Helm upgrade..."
CRD_BASE_DIR="${SCRIPT_DIR}/../agents/k8s-agent/config/crd/bases"
if [ -d "${CRD_BASE_DIR}" ]; then
  kubectl apply -f "${CRD_BASE_DIR}" >/dev/null
fi
echo "   ✅ CRDs synced"

# Step 6: Deploy aegis-services using Helm (FULL deployment)
echo ""
echo "6️⃣  Deploying aegis-services (platform-api + proxy + keycloak + backstage) with Helm..."

# Get JWT secret for proxy
JWT_SECRET=$(terraform output -raw jwt_secret_value)
BACKEND_SECRET=$(openssl rand -hex 32)

OVERRIDE_FILE=$(mktemp)
{
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
  if [[ -n "${PLATFORM_API_IMAGE_REPO}" ]]; then
    echo "    repository: ${PLATFORM_API_IMAGE_REPO}"
  fi
  if [[ -n "${PLATFORM_API_IMAGE_TAG_VALUE}" ]]; then
    echo "    tag: \"${PLATFORM_API_IMAGE_TAG_VALUE}\""
  fi
  echo "  env:"
  echo "    DATABASE_URL: \"${DB_URL}\""
  echo "    OIDC_ISSUER_URL: \"https://${DNS_KEYCLOAK}/realms/aegis\""
  echo "    OIDC_AUDIENCE: \"backstage,aegis-platform\""
  echo "    OIDC_JWKS_URL: \"https://${DNS_KEYCLOAK}/realms/aegis/protocol/openid-connect/certs\""
  if [[ -n "${DNS_PROXY}" ]]; then
    echo "    AEGIS_PROXY_BASE_URL: \"wss://${DNS_PROXY}:8080\""
  fi
  # Spoke provisioning env vars: image tag and dynamic endpoints
  echo "    AEGIS_PLATFORM_API_ENDPOINT: \"${DNS_PLATFORM_API}:8081\""
  echo "    AEGIS_PLATFORM_API_GRPC_INSECURE: \"false\""
  if [[ -n "${K8S_AGENT_IMAGE_REPO}" ]]; then
    echo "    AEGIS_SPOKE_IMAGE_REPO: \"${K8S_AGENT_IMAGE_REPO}\""
  fi
  if [[ -n "${K8S_AGENT_IMAGE_TAG_VALUE}" ]]; then
    echo "    AEGIS_SPOKE_IMAGE_TAG: \"${K8S_AGENT_IMAGE_TAG_VALUE}\""
  fi
  echo "    AEGIS_SPOKE_OIDC_TOKEN_URL: \"https://${DNS_KEYCLOAK}/realms/aegis/protocol/openid-connect/token\""
  echo "    AEGIS_SPOKE_OIDC_CLIENT_ID: \"spoke-agent\""
  echo "    AEGIS_SPOKE_OIDC_CLIENT_SECRET: \"${SPOKE_OIDC_CLIENT_SECRET}\""
  echo "    AEGIS_SPOKE_OIDC_AUDIENCE: \"aegis-platform\""
  echo "    AEGIS_SPOKE_VALUES_FILE: \"/root/charts/aegis-spoke/values-cloud-remote.yaml\""
  echo "  secrets:"
  echo "    db-password: \"${DB_PASSWORD}\""
  echo "    proxy-jwt-secret: \"${JWT_SECRET}\""
  echo "  tls:"
  echo "    certManager:"
  echo "      enabled: true"
  echo "      dnsNames:"
  echo "        - \"${DNS_PLATFORM_API}\""
  echo "proxy:"
  if [[ -n "${DNS_PROXY}" ]]; then
    echo "  publicHost: \"${DNS_PROXY}\""
  fi
  echo "  image:"
  if [[ -n "${PROXY_IMAGE_REPO}" ]]; then
    echo "    repository: ${PROXY_IMAGE_REPO}"
  fi
  if [[ -n "${PROXY_IMAGE_TAG_VALUE}" ]]; then
    echo "    tag: \"${PROXY_IMAGE_TAG_VALUE}\""
  fi
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
  if [[ -n "${UI_IMAGE_REPO}" ]]; then
    echo "    repository: ${UI_IMAGE_REPO}"
  fi
  if [[ -n "${UI_IMAGE_TAG_VALUE}" ]]; then
    echo "    tag: \"${UI_IMAGE_TAG_VALUE}\""
  fi
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

cd "${OUTPUT_DIR}"
HELM_ARGS=(
  upgrade --install "${HELM_RELEASE}" ./aegis-services
  -f ./aegis-services/values/common.yaml
  -f ./aegis-services/values/cloud.yaml
  -f ./aegis-services/values-cloud-generated.yaml
  -f "${OVERRIDE_FILE}"
)

HELM_ARGS+=(
  --namespace "${K8S_NAMESPACE}" --create-namespace
  --timeout 10m
)

# Set discovery endpoint for VS Code extension auto-configuration
if [[ -n "${DNS_PLATFORM_API}" ]]; then
  HELM_ARGS+=( --set "platformApi.env.AEGIS_DISCOVERY_GRPC_ENDPOINT=${DNS_PLATFORM_API}:8081" )
fi

helm "${HELM_ARGS[@]}"


echo "   ✅ aegis-services deployed"
echo "   ⏳ Waiting for deployments to become ready"
kubectl rollout status "deployment/${HELM_RELEASE}-platform-api" -n "${K8S_NAMESPACE}" --timeout=5m
kubectl rollout status "deployment/${HELM_RELEASE}-proxy" -n "${K8S_NAMESPACE}" --timeout=5m
echo "   ⏳ Waiting for Keycloak components to become ready"
kubectl rollout status "statefulset/${HELM_RELEASE}-keycloak-db" -n "${K8S_NAMESPACE}" --timeout=10m || \
  { echo "❌ Keycloak Postgres StatefulSet not ready within timeout" >&2; exit 1; }
kubectl rollout status "statefulset/${HELM_RELEASE}-keycloak" -n "${K8S_NAMESPACE}" --timeout=10m || \
  { echo "❌ Keycloak StatefulSet not ready within timeout" >&2; exit 1; }

# Step 6a: Verify spoke provisioning env vars in platform-api deployment
echo ""
echo "   🔍 Verifying platform-api spoke provisioning configuration..."

verify_pod_env() {
  local var_name="$1"
  local required="$2"
  local value
  value=$(kubectl -n "${K8S_NAMESPACE}" get deployment "${HELM_RELEASE}-platform-api" \
    -o jsonpath="{.spec.template.spec.containers[0].env[?(@.name==\"${var_name}\")].value}" 2>/dev/null || echo "")
  if [[ -z "${value}" ]]; then
    if [[ "${required}" == "required" ]]; then
      echo "   ❌ FATAL: platform-api missing required env var: ${var_name}" >&2
      return 1
    else
      echo "   ⚠️  Optional env var not set: ${var_name}"
    fi
  else
    local display="${value:0:50}"
    [[ ${#value} -gt 50 ]] && display="${display}..."
    echo "   ✅ ${var_name} = ${display}"
  fi
}

SPOKE_ENV_OK=true
verify_pod_env "AEGIS_SPOKE_IMAGE_REPO" "required"      || SPOKE_ENV_OK=false
verify_pod_env "AEGIS_PLATFORM_API_ENDPOINT" "required"  || SPOKE_ENV_OK=false
verify_pod_env "AEGIS_SPOKE_VALUES_FILE" "required"      || SPOKE_ENV_OK=false
verify_pod_env "AEGIS_SPOKE_OIDC_TOKEN_URL" "optional"
verify_pod_env "AEGIS_SPOKE_IMAGE_TAG" "optional"

if [[ "${SPOKE_ENV_OK}" != "true" ]]; then
  echo "" >&2
  echo "   ❌ FATAL: platform-api is missing critical spoke provisioning env vars." >&2
  echo "   Spoke cluster provisioning will fail silently with wrong defaults." >&2
  echo "   Fix Helm values (cloud.yaml or deploy script override) and re-deploy." >&2
  exit 1
fi
echo "   ✅ Spoke provisioning config verified"

# Step 6b: Run in-cluster Postgres migrations (when SKIP_MIGRATION_PLACEHOLDER=1)
if [[ "${SKIP_MIGRATION_PLACEHOLDER:-0}" == "1" ]]; then
  echo ""
  echo "   ⚙️  Running in-cluster Postgres migrations..."

  MIGRATIONS_DIR="${SCRIPT_DIR}/../services/platform-api/migrations"
  if [ ! -f "${MIGRATIONS_DIR}/0001_init.sql" ]; then
    echo "   ⚠️  No migration files found; skipping"
  else
    # Wait for postgres pod
    echo "   ⏳ Waiting for platform-postgres to be ready..."
    kubectl -n "${K8S_NAMESPACE}" wait --for=condition=ready pod -l app.kubernetes.io/name=platform-postgres --timeout=120s 2>/dev/null || true
    PG_POD=$(kubectl -n "${K8S_NAMESPACE}" get pods -l app.kubernetes.io/name=platform-postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

    if [[ -n "${PG_POD}" ]]; then
      # Extract Up-only migration SQL
      ALL_UP=$(mktemp)
      for mig in "${MIGRATIONS_DIR}"/*.sql; do
        awk '/^--[[:space:]]+\+migrate[[:space:]]+Down/{exit} {print}' "$mig" >> "$ALL_UP"
        echo "" >> "$ALL_UP"
      done

      kubectl cp "$ALL_UP" "${K8S_NAMESPACE}/${PG_POD}:/tmp/all_migrations.sql"
      kubectl exec -n "${K8S_NAMESPACE}" "${PG_POD}" -- \
        psql -U aegis_platform -d aegis_platform -v ON_ERROR_STOP=1 -f /tmp/all_migrations.sql
      rm -f "$ALL_UP"
      echo "   ✅ In-cluster migrations applied"
    else
      echo "   ⚠️  platform-postgres pod not found; skipping migrations"
    fi
  fi
fi

# Step 7: Wait for Load Balancers
echo ""
echo "7️⃣  Waiting for Load Balancers to provision (this takes ~2 minutes)..."
PLATFORM_API_LB=""
PROXY_LB=""
PUBLIC_INGRESS_LB=""

wait_for_lb_hostname "platform-api" "svc/${PLATFORM_API_RELEASE_NAME}" "${K8S_NAMESPACE}" PLATFORM_API_LB
wait_for_lb_dns "${PLATFORM_API_LB}"
echo ""
wait_for_lb_hostname "proxy" "svc/${PROXY_RELEASE_NAME}" "${K8S_NAMESPACE}" PROXY_LB
wait_for_lb_dns "${PROXY_LB}"

INGRESS_CONTROLLER_SERVICE=$(kubectl get svc -n "${K8S_NAMESPACE}" \
  -l app.kubernetes.io/component=controller,app.kubernetes.io/name=ingress-nginx \
  -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
if [[ -z "${INGRESS_CONTROLLER_SERVICE}" ]]; then
  echo "❌ ingress-nginx controller service not found in namespace ${K8S_NAMESPACE}" >&2
  exit 1
fi
echo ""
wait_for_lb_hostname "public ingress" "svc/${INGRESS_CONTROLLER_SERVICE}" "${K8S_NAMESPACE}" PUBLIC_INGRESS_LB
wait_for_lb_dns "${PUBLIC_INGRESS_LB}"

echo ""
echo "   ℹ️  Proxy endpoint configured via Helm: wss://${DNS_PROXY}:8080"

echo ""
echo "   🌐 Updating Cloudflare DNS records with LoadBalancer hostnames..."
if [[ "${SKIP_DNS_UPDATE}" == "1" ]]; then
  echo "   ⚠️  SKIP_DNS_UPDATE=1, skipping DNS changes"
else
  update_cloudflare_cname "${DNS_PLATFORM_API}" "${PLATFORM_API_LB}"
  update_cloudflare_cname "${DNS_PROXY}" "${PROXY_LB}"
  update_cloudflare_cname "${DNS_UI}" "${PUBLIC_INGRESS_LB}"
  update_cloudflare_cname "${DNS_KEYCLOAK}" "${PUBLIC_INGRESS_LB}"
  echo "   ✅ Cloudflare DNS updated"
  echo "   📋 DNS Records:"
  echo "      ${DNS_PLATFORM_API} → ${PLATFORM_API_LB}"
  echo "      ${DNS_PROXY} → ${PROXY_LB}"
  echo "      ${DNS_UI} → ${PUBLIC_INGRESS_LB}"
  echo "      ${DNS_KEYCLOAK} → ${PUBLIC_INGRESS_LB}"
fi

echo ""
echo "   🔄 Restarting Backstage after public DNS update"
kubectl rollout restart "deployment/${HELM_RELEASE}-backstage" -n "${K8S_NAMESPACE}" >/dev/null
kubectl rollout status "deployment/${HELM_RELEASE}-backstage" -n "${K8S_NAMESPACE}" --timeout=10m

# Step 8: Deploy k8s-agent (aegis-spoke)
echo ""
echo "8️⃣  Deploying k8s-agent with Helm..."

cd "${OUTPUT_DIR}"

# Build helm arguments for aegis-spoke
SPOKE_HELM_ARGS=(
  upgrade --install "${SPOKE_HELM_RELEASE}" ./aegis-spoke
  -f ./aegis-spoke/values-cloud-generated.yaml
  -f ./aegis-spoke/values-cloud-tls.yaml
)
echo "   ℹ️  Configuring k8s-agent to require TLS when dialing the hub"

SPOKE_HELM_ARGS+=(
  --set k8sAgent.enabled=true
  --set k8sAgent.replicaCount=1
  --set k8sAgent.trustBundle.enabled=true
  --set k8sAgent.trustBundle.secretName=aegis-trust-bundle
  --set proxy.enabled=false
  --set "k8sAgent.env.AEGIS_CP_OIDC_TOKEN_URL=https://${KEYCLOAK_INTERNAL_HOST}:${KEYCLOAK_INTERNAL_PORT}/realms/aegis/protocol/openid-connect/token"
  --set "k8sAgent.env.AEGIS_CP_OIDC_CLIENT_ID=spoke-agent"
  --set "k8sAgent.env.AEGIS_CP_OIDC_CLIENT_SECRET=${SPOKE_OIDC_CLIENT_SECRET}"
  --set "k8sAgent.env.AEGIS_CP_OIDC_AUDIENCE=aegis-platform"
  --namespace "${SPOKE_NAMESPACE}"
  --create-namespace
  --timeout 5m
)

if [[ -n "${K8S_AGENT_IMAGE_REPO}" ]]; then
  SPOKE_HELM_ARGS+=( --set k8sAgent.image.repository=${K8S_AGENT_IMAGE_REPO} )
fi
if [[ -n "${K8S_AGENT_IMAGE_TAG_VALUE}" ]]; then
  SPOKE_HELM_ARGS+=( --set k8sAgent.image.tag=${K8S_AGENT_IMAGE_TAG_VALUE} )
fi

# VS Code REH init container — uses same repo prefix and tag as k8s-agent
if [[ -n "${K8S_AGENT_IMAGE_REPO}" && -n "${K8S_AGENT_IMAGE_TAG_VALUE}" ]]; then
  VSCODE_REH_INIT_REPO="${K8S_AGENT_IMAGE_REPO/aegis\/k8s-agent/aegis\/vscode-reh-init}"
  SPOKE_HELM_ARGS+=( --set "k8sAgent.env.AEGIS_VSCODE_REH_INIT_IMAGE=${VSCODE_REH_INIT_REPO}:${K8S_AGENT_IMAGE_TAG_VALUE}" )
  echo "   ℹ️  VS Code REH init container: ${VSCODE_REH_INIT_REPO}:${K8S_AGENT_IMAGE_TAG_VALUE}"
fi

echo "   🔧 Ensuring existing CRDs are adoptable by Helm..."
SPOKE_CRDS=( \
  aegisworkloads.aegis.yourorg.dev \
  aegisworkloadslices.aegis.yourorg.dev \
  workspaceclasses.aegis.yourorg.dev \
  workspaceclusters.aegis.yourorg.dev \
  workspacegpuprofiles.aegis.yourorg.dev \
  workspacenetworks.aegis.yourorg.dev \
  workspaces.aegis.yourorg.dev \
  workspacestorages.aegis.yourorg.dev \
)
for crd in "${SPOKE_CRDS[@]}"; do
  if kubectl get crd "${crd}" >/dev/null 2>&1; then
    existing_rel="$(kubectl get crd "${crd}" -o jsonpath='{.metadata.annotations.meta\.helm\.sh/release-name}' 2>/dev/null || true)"
    existing_ns="$(kubectl get crd "${crd}" -o jsonpath='{.metadata.annotations.meta\.helm\.sh/release-namespace}' 2>/dev/null || true)"
    if [[ -z "${existing_rel}" && -z "${existing_ns}" ]]; then
      kubectl label crd "${crd}" app.kubernetes.io/managed-by=Helm --overwrite >/dev/null
      kubectl annotate crd "${crd}" \
        meta.helm.sh/release-name="${SPOKE_HELM_RELEASE}" \
        meta.helm.sh/release-namespace="${SPOKE_NAMESPACE}" \
        --overwrite >/dev/null
    fi
  fi
done

helm "${SPOKE_HELM_ARGS[@]}"

echo "   ✅ k8s-agent deployed"

# Step 8b: Register in-cluster kubeconfig for co-located spoke
echo ""
echo "8️⃣ b Registering in-cluster kubeconfig for spoke cluster..."
SPOKE_CLUSTER_ID="$(kubectl get deploy -n "${SPOKE_NAMESPACE}" -l app.kubernetes.io/name=k8s-agent \
  -o jsonpath='{.items[0].spec.template.spec.containers[0].env[?(@.name=="AEGIS_CLUSTER_ID")].value}' 2>/dev/null || true)"

if [[ -n "${SPOKE_CLUSTER_ID}" ]]; then
  IN_CLUSTER_KC=$(cat <<EOKC
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    server: https://kubernetes.default.svc
  name: ${SPOKE_CLUSTER_ID}
contexts:
- context:
    cluster: ${SPOKE_CLUSTER_ID}
    user: ${SPOKE_CLUSTER_ID}
  name: ${SPOKE_CLUSTER_ID}
current-context: ${SPOKE_CLUSTER_ID}
kind: Config
users:
- name: ${SPOKE_CLUSTER_ID}
  user:
    tokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
EOKC
)
  KC_B64=$(echo "${IN_CLUSTER_KC}" | base64 | tr -d '\n')
  KC_KEY="${SPOKE_CLUSTER_ID}.kubeconfig"

  # Create or update the kubeconfigs secret
  if kubectl -n "${K8S_NAMESPACE}" get secret aegis-kubeconfigs >/dev/null 2>&1; then
    kubectl -n "${K8S_NAMESPACE}" patch secret aegis-kubeconfigs --type=json \
      -p "[{\"op\":\"add\",\"path\":\"/data/${KC_KEY}\",\"value\":\"${KC_B64}\"}]" >/dev/null 2>&1 || true
  else
    kubectl -n "${K8S_NAMESPACE}" create secret generic aegis-kubeconfigs \
      --from-literal="${KC_KEY}=$(echo "${IN_CLUSTER_KC}")" >/dev/null
  fi
  echo "   ✅ Kubeconfig registered for cluster '${SPOKE_CLUSTER_ID}'"

  # Update database with kubeconfig_secret_ref (idempotent)
  POSTGRES_POD="$(kubectl -n "${K8S_NAMESPACE}" get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
  if [[ -n "${POSTGRES_POD}" ]]; then
    kubectl -n "${K8S_NAMESPACE}" exec "${POSTGRES_POD}" -- \
      psql -U aegis_platform -d aegis_platform -c \
      "UPDATE clusters SET kubeconfig_secret_ref = '${K8S_NAMESPACE}/aegis-kubeconfigs:${KC_KEY}', project_id = COALESCE(NULLIF(project_id, ''), 'default') WHERE id = '${SPOKE_CLUSTER_ID}' AND deleted_at IS NULL;" \
      >/dev/null 2>&1 || echo "   ⚠️  DB update skipped (cluster may not be registered yet — will auto-resolve on next heartbeat)"
  fi
else
  echo "   ⚠️  Could not determine AEGIS_CLUSTER_ID from spoke deployment — skipping kubeconfig registration"
fi

# Step 8c: Bootstrap default project with policy regions
echo ""
echo "8️⃣ c Bootstrapping default project..."
BOOTSTRAP_TOKEN=$(curl -sk -X POST "https://${KEYCLOAK_INTERNAL_HOST}:${KEYCLOAK_INTERNAL_PORT}/realms/aegis/protocol/openid-connect/token" \
  -d "grant_type=client_credentials" \
  -d "client_id=spoke-agent" \
  -d "client_secret=${SPOKE_OIDC_CLIENT_SECRET}" 2>/dev/null \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))" 2>/dev/null || true)

if [[ -n "${BOOTSTRAP_TOKEN}" ]]; then
  # Port-forward to platform-api for gRPC call
  kubectl -n "${K8S_NAMESPACE}" port-forward svc/"${HELM_RELEASE}-platform-api" 18081:8081 &
  PF_PID=$!
  sleep 3

  DEPLOY_REGION="$(cd "${SCRIPT_DIR}" && terraform output -raw aws_region 2>/dev/null || echo "us-east-1")"
  grpcurl -insecure -H "authorization: Bearer ${BOOTSTRAP_TOKEN}" \
    -d "{\"project\":{\"id\":\"default\",\"display_name\":\"Default Project\",\"policy\":{\"regions\":[\"${DEPLOY_REGION}\"]}}}" \
    localhost:18081 aegis.v1.AegisPlatform/CreateProject >/dev/null 2>&1 && \
    echo "   ✅ Default project created with region '${DEPLOY_REGION}'" || \
    echo "   ⚠️  Project bootstrap skipped (grpcurl may not be installed or platform-api not ready)"

  kill ${PF_PID} 2>/dev/null || true
  wait ${PF_PID} 2>/dev/null || true
else
  echo "   ⚠️  Could not obtain Keycloak token — project bootstrap skipped"
fi

# Step 9: Verify canonical public endpoints
echo ""
echo "9️⃣  Verifying canonical public endpoints..."

for host in "${DNS_PLATFORM_API}" "${DNS_PROXY}" "${DNS_UI}" "${DNS_KEYCLOAK}"; do
  wait_for_public_dns "${host}"
done

curl -fsS --max-time 20 "http://${DNS_PLATFORM_API}:8080/healthz" >/dev/null || {
  echo "❌ Platform API health check failed: http://${DNS_PLATFORM_API}:8080/healthz" >&2
  exit 1
}
curl -fsS --max-time 20 "http://${DNS_PROXY}:8080/healthz" >/dev/null || {
  echo "❌ Proxy health check failed: http://${DNS_PROXY}:8080/healthz" >&2
  exit 1
}
curl -fsS --max-time 20 "https://${DNS_UI}/healthcheck" >/dev/null || {
  echo "❌ Backstage health check failed: https://${DNS_UI}/healthcheck" >&2
  exit 1
}
curl -fsS --max-time 20 "https://${DNS_UI}/" >/dev/null || {
  echo "❌ Backstage root endpoint failed: https://${DNS_UI}/" >&2
  exit 1
}
curl -fsS --max-time 20 "https://${DNS_KEYCLOAK}/realms/aegis/.well-known/openid-configuration" >/dev/null || {
  echo "❌ Keycloak discovery failed: https://${DNS_KEYCLOAK}/realms/aegis/.well-known/openid-configuration" >&2
  exit 1
}
echo "   ✅ Canonical public endpoints verified"

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🎉 Deployment Complete!"
echo ""
echo "📊 Service URLs:"
echo "   Platform API (HTTP gateway): http://${DNS_PLATFORM_API}:8080"
echo "   Platform API (gRPC/TLS):     ${DNS_PLATFORM_API}:8081"
echo "   Keycloak:                    https://${DNS_KEYCLOAK}"
echo "   Proxy (WSS):                 wss://${DNS_PROXY}:8080"
echo "   Backstage UI:                https://${DNS_UI}"
echo ""
echo "🌐 LoadBalancer Endpoints:"
echo "   Platform API: ${PLATFORM_API_LB}"
echo "   Proxy:        ${PROXY_LB}"
echo "   Public Ingress (UI/Keycloak): ${PUBLIC_INGRESS_LB}"
echo ""
echo "📝 Check deployment status:"
echo "   kubectl get pods -n ${K8S_NAMESPACE}"
echo "   kubectl get svc -n ${K8S_NAMESPACE}"
echo ""
echo "🔍 View logs:"
echo "   kubectl logs -n ${K8S_NAMESPACE} -l app.kubernetes.io/component=platform-api -f"
echo ""
echo "🔐 Test gRPC with grpcurl:"
echo "   export GRPC_HOST=${DNS_PLATFORM_API}"
echo "   export CA_BUNDLE=${CA_BUNDLE}"
echo ""
echo "   grpcurl -cacert \"\$CA_BUNDLE\" \\"
echo "     -d '{\"project\":{\"id\":\"p-demo\",\"displayName\":\"Demo\",\"ownerGroup\":\"eng\"}}' \\"
echo "     \${GRPC_HOST}:8081 aegis.v1.AegisPlatform/CreateProject"
echo ""
echo "📱 VSCode Extension Configuration:"
echo "   grpcEndpoint: \"${DNS_PLATFORM_API}:8081\""
echo "   caPath: \"${CA_BUNDLE}\""
echo ""
echo "🚀 Next steps:"
echo "   1. Open Backstage at: https://${DNS_UI}"
echo "   2. Verify login redirects to: https://${DNS_KEYCLOAK}"
echo "   3. Test UI workflows against platform-api/proxy on aegis-platform.tech"
echo ""
echo "🔐 TLS mode: gRPC clients must trust ${CA_BUNDLE}"
echo ""
echo "✨ Done!"
