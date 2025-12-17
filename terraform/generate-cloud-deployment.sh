#!/bin/bash
# Generate Helm values and optionally deploy to Kubernetes
#
# For complete deployment documentation, see: DEPLOYMENT_AND_TESTING_GUIDE.md

set -e

NON_INTERACTIVE=0
K8S_NAMESPACE=${K8S_NAMESPACE:-aegis-system}
HELM_RELEASE=${HELM_RELEASE:-aegis}
SKIP_ROUTE53_UPDATE=${SKIP_ROUTE53_UPDATE:-0}

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
PLATFORM_API_IMAGE_TAG=${PLATFORM_API_IMAGE_TAG:-"v1.0.6-tls2"}
PROXY_IMAGE_TAG=${PROXY_IMAGE_TAG:-"no-client-cert"}
K8S_AGENT_IMAGE_TAG=${K8S_AGENT_IMAGE_TAG:-"v1.0.2-tls-20251005-amd64"}
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

IFS='|' read -r PLATFORM_API_IMAGE_REPO PLATFORM_API_IMAGE_TAG_VALUE <<< "$(parse_image_ref "${PLATFORM_API_IMAGE_TAG}")"
IFS='|' read -r PROXY_IMAGE_REPO PROXY_IMAGE_TAG_VALUE <<< "$(parse_image_ref "${PROXY_IMAGE_TAG}")"
IFS='|' read -r K8S_AGENT_IMAGE_REPO K8S_AGENT_IMAGE_TAG_VALUE <<< "$(parse_image_ref "${K8S_AGENT_IMAGE_TAG}")"

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

# Generate aegis-services values
echo "📝 Generating aegis-services values..."
cd "${SCRIPT_DIR}"
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
DB_ENDPOINT=$(terraform output -raw rds_endpoint)
if [[ "${DB_ENDPOINT}" == *:* ]]; then
  DB_HOST=${DB_ENDPOINT%%:*}
  DB_PORT=${DB_ENDPOINT##*:}
else
  DB_HOST=${DB_ENDPOINT}
  DB_PORT=$(terraform output -raw rds_port 2>/dev/null || echo "5432")
fi
DB_NAME=$(terraform output -raw rds_database_name 2>/dev/null || echo "aegis")
DB_USER_FALLBACK="aegis_api"
if terraform output -raw rds_username >/tmp/rds_user 2>/dev/null; then
  DB_USER=$(cat /tmp/rds_user)
  rm -f /tmp/rds_user
else
  DB_USER=${DB_USER_FALLBACK}
fi
DB_URL="postgres://${DB_USER}:${DB_PASSWORD_ENCODED}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require"

if [[ -z "${AWS_REGION:-}" ]]; then
  AWS_REGION=$(terraform output -raw aws_region 2>/dev/null || echo "")
fi

if [[ -z "${AWS_REGION}" ]]; then
  echo "❌ Unable to determine AWS region for migrations"
  exit 1
fi

RDS_CA_BUNDLE_URL="https://truststore.pki.rds.amazonaws.com/${AWS_REGION}/${AWS_REGION}-bundle.pem"
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
else
  JOB_ANNOTATIONS_BLOCK=""
  POD_ANNOTATIONS_BLOCK=""
fi


if [[ "${SKIP_MIGRATION_PLACEHOLDER:-0}" == "1" ]]; then
  echo "   ⚠️  Skipping migration placeholder (handled externally)"
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

# Step 4: Ensure internal PKI (step-ca + cert-manager) and refresh CA bundle
echo ""
echo "4️⃣  Ensuring internal PKI (step-ca + cert-manager)..."

# Get DNS hostnames from Terraform outputs (used for Certificate SANs)
DNS_PLATFORM_API_GRPC=$(terraform output -raw dns_platform_api_grpc 2>/dev/null || echo "platform-api-grpc.aegist.dev")
DNS_PLATFORM_API_HTTP=$(terraform output -raw dns_platform_api_http 2>/dev/null || echo "platform-api.aegist.dev")
DNS_PROXY=$(terraform output -raw dns_proxy 2>/dev/null || echo "proxy.aegist.dev")

echo "   📋 DNS hostnames:"
echo "      Platform API gRPC: ${DNS_PLATFORM_API_GRPC}"
echo "      Platform API HTTP: ${DNS_PLATFORM_API_HTTP}"
echo "      Proxy:             ${DNS_PROXY}"

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
echo "6️⃣  Deploying aegis-services (platform-api + proxy) with Helm..."

# Get JWT secret for proxy
JWT_SECRET=$(terraform output -raw jwt_secret_value)

OVERRIDE_FILE=$(mktemp)
{
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
  echo "    OIDC_ISSUER_URL: \"https://${KEYCLOAK_INTERNAL_HOST}:${KEYCLOAK_INTERNAL_PORT}/realms/aegis\""
  echo "    OIDC_AUDIENCE: \"backstage\""
  echo "    OIDC_JWKS_URL: \"https://${KEYCLOAK_INTERNAL_HOST}:${KEYCLOAK_INTERNAL_PORT}/realms/aegis/protocol/openid-connect/certs\""
  if [[ -n "${DNS_PROXY}" ]]; then
    echo "    AEGIS_PROXY_BASE_URL: \"wss://${DNS_PROXY}:8080\""
  fi
  echo "  secrets:"
  echo "    db-password: \"${DB_PASSWORD}\""
  echo "    proxy-jwt-secret: \"${JWT_SECRET}\""
  echo "  tls:"
  echo "    certManager:"
  echo "      enabled: true"
  echo "      dnsNames:"
  echo "        - \"${DNS_PLATFORM_API_GRPC}\""
  echo "        - \"${DNS_PLATFORM_API_HTTP}\""
  echo "  auth:"
  echo "    oidc:"
  echo "      caBundle:"
  echo "        create: false"
  echo "        secretName: aegis-trust-bundle"
  echo "        key: ca.crt"
  echo "        fileName: ca.crt"
  echo "        mountPath: /etc/aegis-platform-api/oidc"
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
  echo "keycloak:"
  echo "  enabled: true"
  echo "  forceRender: true"
  echo "  namespace: ${K8S_NAMESPACE}"
  echo "  hostname:"
    echo "    hostname: \"https://${KEYCLOAK_INTERNAL_HOST}:${KEYCLOAK_INTERNAL_PORT}\""
  echo "    admin: \"https://${KEYCLOAK_INTERNAL_HOST}:${KEYCLOAK_INTERNAL_PORT}\""
  echo "    strict: false"
  echo "  ingress:"
  echo "    enabled: false"
  echo "  customIngress:"
  echo "    enabled: false"
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

helm "${HELM_ARGS[@]}"


echo "   ✅ aegis-services deployed"
echo "   🔄 Restarting workloads to pick up latest configuration"
kubectl rollout restart "deployment/${HELM_RELEASE}-platform-api" -n "${K8S_NAMESPACE}" >/dev/null
kubectl rollout restart "deployment/${HELM_RELEASE}-proxy" -n "${K8S_NAMESPACE}" >/dev/null

echo "   ⏳ Waiting for deployments to become ready"
kubectl rollout status "deployment/${HELM_RELEASE}-platform-api" -n "${K8S_NAMESPACE}" --timeout=5m
kubectl rollout status "deployment/${HELM_RELEASE}-proxy" -n "${K8S_NAMESPACE}" --timeout=5m
echo "   ⏳ Waiting for Keycloak components to become ready"
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/component=keycloak -n "${K8S_NAMESPACE}" --timeout=5m >/dev/null 2>&1 || \
  echo "   ⚠️  Keycloak pods not ready within timeout"
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/component=keycloak-postgres -n "${K8S_NAMESPACE}" --timeout=5m >/dev/null 2>&1 || \
  echo "   ⚠️  Keycloak Postgres pods not ready within timeout"

# Step 7: Wait for Load Balancers
echo ""
echo "7️⃣  Waiting for Load Balancers to provision (this takes ~2 minutes)..."

echo "   Waiting for platform-api Load Balancer..."
for i in {1..60}; do
  PLATFORM_API_LB=$(kubectl get svc "${PLATFORM_API_RELEASE_NAME}" -n "${K8S_NAMESPACE}" -o jsonpath='{.status.loadBalancer.i
ngress[0].hostname}' 2>/dev/null || echo "")
  if [ -n "${PLATFORM_API_LB}" ]; then
    echo "   ✅ Platform API Load Balancer ready: ${PLATFORM_API_LB}"
    break
  fi
  echo -n "."
  sleep 2
done

if [ -n "${PLATFORM_API_LB}" ]; then
  echo "   Waiting for DNS propagation for ${PLATFORM_API_LB}..."
  for i in {1..30}; do
    if dig +short "${PLATFORM_API_LB}" | grep -q '^[0-9]'; then
      echo "   ✅ DNS resolved: $(dig +short ${PLATFORM_API_LB} | head -1)"
      break
    fi
    echo -n "."
    sleep 2
  done
fi

echo ""
echo "   Waiting for proxy Load Balancer..."
for i in {1..60}; do
  PROXY_LB=$(kubectl get svc "${PROXY_RELEASE_NAME}" -n "${K8S_NAMESPACE}" -o jsonpath='{.status.loadBalancer.ingress[0].host
name}' 2>/dev/null || echo "")
  if [ -n "${PROXY_LB}" ]; then
    echo "   ✅ Proxy Load Balancer ready: ${PROXY_LB}"
    break
  fi
  echo -n "."
  sleep 2
done

if [ -n "${PROXY_LB}" ]; then
  echo "   Waiting for DNS propagation for ${PROXY_LB}..."
  for i in {1..30}; do
    if dig +short "${PROXY_LB}" | grep -q '^[0-9]'; then
      echo "   ✅ DNS resolved: $(dig +short ${PROXY_LB} | head -1)"
      break
    fi
    echo -n "."
    sleep 2
  done
fi

if [ -n "${PROXY_LB}" ]; then
  echo ""
  echo "   ℹ️  Proxy endpoint configured via Helm: wss://${DNS_PROXY}:8080"
  echo "   (No kubectl set env needed - Helm manages AEGIS_PROXY_BASE_URL)"
fi

# Update Route53 DNS records with actual LoadBalancer hostnames
echo ""
echo "   🌐 Updating Route53 DNS records with LoadBalancer hostnames..."
if [[ "${SKIP_ROUTE53_UPDATE}" == "1" ]]; then
  echo "   ⚠️  SKIP_ROUTE53_UPDATE=1, skipping Route53 changes"
elif [ -n "${PLATFORM_API_LB}" ] && [ -n "${PROXY_LB}" ]; then
  cd "${SCRIPT_DIR}"
  terraform apply -auto-approve \
    -var="platform_api_lb_hostname=${PLATFORM_API_LB}" \
    -var="proxy_lb_hostname=${PROXY_LB}" \
    -target=aws_route53_record.platform_api_grpc \
    -target=aws_route53_record.platform_api_http \
    -target=aws_route53_record.proxy >/dev/null 2>&1 \
    && echo "   ✅ Route53 records updated" \
    || echo "   ⚠️  Route53 update failed; update manually"

  echo "   📋 DNS Records:"
  echo "      platform-api-grpc.aegist.dev → ${PLATFORM_API_LB}"
  echo "      platform-api.aegist.dev      → ${PLATFORM_API_LB}"
  echo "      proxy.aegist.dev             → ${PROXY_LB}"

  # Update /etc/hosts for local DNS resolution
  echo ""
  echo "   🖥️  Updating /etc/hosts for local DNS resolution..."
  # Get ALL IPs from Network Load Balancers (NLBs have multiple IPs)
  PLATFORM_API_IPS=$(dig +short "${PLATFORM_API_LB}" | grep '^[0-9]' | tr '\n' ' ')
  PROXY_IPS=$(dig +short "${PROXY_LB}" | grep '^[0-9]' | tr '\n' ' ')
  # Use first IP for /etc/hosts entry
  PLATFORM_API_IP=$(echo "${PLATFORM_API_IPS}" | awk '{print $1}')
  PROXY_IP=$(echo "${PROXY_IPS}" | awk '{print $1}')

  if [ -n "${PLATFORM_API_IP}" ] && [ -n "${PROXY_IP}" ]; then
    # Remove old aegist.dev entries
    sudo sed -i.bak '/aegist\.dev/d' /etc/hosts 2>/dev/null || true

    # Add new entries (using first IP from NLB)
    echo "${PLATFORM_API_IP} platform-api-grpc.aegist.dev platform-api.aegist.dev" | sudo tee -a /etc/hosts >/dev/null
    echo "${PROXY_IP} proxy.aegist.dev" | sudo tee -a /etc/hosts >/dev/null

    echo "   ✅ /etc/hosts updated:"
    echo "      ${PLATFORM_API_IP} → platform-api-grpc.aegist.dev, platform-api.aegist.dev"
    echo "      ${PROXY_IP} → proxy.aegist.dev"
    echo "   📝 Note: NLB IPs (all): platform-api=${PLATFORM_API_IPS}, proxy=${PROXY_IPS}"
  else
    echo "   ⚠️  Could not resolve LoadBalancer IPs; /etc/hosts not updated"
  fi

  cd "${OUTPUT_DIR}"
else
  echo "   ⚠️  LoadBalancers not ready; skip Route53 update"
fi

# Step 7: Update Backstage configuration
echo ""
echo "7️⃣  Updating Backstage configuration..."

if [ -n "${PLATFORM_API_LB}" ]; then
  APP_TARGET="http://${DNS_PLATFORM_API_HTTP}:8080"
  SECURE_FLAG=false

  # Generate Backstage proxy configuration
  BACKSTAGE_CONFIG=$(cat <<EOF_BACKSTAGE
# Backstage override configuration for your cloud deployment

proxy:
  endpoints:
    '/aegis':
      target: '${APP_TARGET}'
      changeOrigin: true
      credentials: forward
      secure: ${SECURE_FLAG}
      allowedHeaders:
        - authorization
        - Authorization
EOF_BACKSTAGE
)

  # Write to both cloud config files for convenience
  echo "${BACKSTAGE_CONFIG}" > "${OUTPUT_DIR}/../aegis-platform/app-config.cloud.yaml"
  echo "   ✅ Updated: aegis-platform/app-config.cloud.yaml"

  echo "${BACKSTAGE_CONFIG}" > "${OUTPUT_DIR}/../aegis-platform/app-config.cloud-tls.yaml"
  echo "   ✅ Updated: aegis-platform/app-config.cloud-tls.yaml"

  echo "${BACKSTAGE_CONFIG}" > "${OUTPUT_DIR}/../aegis-platform/app-config.local.yaml"
  echo "   ✅ Updated: aegis-platform/app-config.local.yaml"
fi

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

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🎉 Deployment Complete!"
echo ""
echo "📊 Service URLs:"
echo "   Platform API (HTTP gateway): http://${DNS_PLATFORM_API_HTTP}:8080"
echo "   Platform API (gRPC/TLS):     ${DNS_PLATFORM_API_GRPC}:8081"
echo "   Proxy (WSS):                 wss://${DNS_PROXY}:8080"
echo ""
echo "🌐 LoadBalancer Endpoints:"
echo "   Platform API: ${PLATFORM_API_LB}"
echo "   Proxy:        ${PROXY_LB}"
echo ""
echo "📝 Check deployment status:"
echo "   kubectl get pods -n ${K8S_NAMESPACE}"
echo "   kubectl get svc -n ${K8S_NAMESPACE}"
echo ""
echo "🔍 View logs:"
echo "   kubectl logs -n ${K8S_NAMESPACE} -l app.kubernetes.io/component=platform-api -f"
echo ""
echo "🔐 Test gRPC with grpcurl:"
echo "   export GRPC_HOST=${DNS_PLATFORM_API_GRPC}"
echo "   export CA_BUNDLE=${CA_BUNDLE}"
echo ""
echo "   grpcurl -cacert \"\$CA_BUNDLE\" \\"
echo "     -d '{\"project\":{\"id\":\"p-demo\",\"displayName\":\"Demo\",\"ownerGroup\":\"eng\"}}' \\"
echo "     \${GRPC_HOST}:8081 aegis.v1.AegisPlatform/CreateProject"
echo ""
echo "📱 VSCode Extension Configuration:"
echo "   grpcEndpoint: \"${DNS_PLATFORM_API_GRPC}:8081\""
echo "   caPath: \"${CA_BUNDLE}\""
echo ""
echo "🚀 Next steps:"
echo "   1. Start Backstage UI: cd aegis-platform && yarn dev:cloud-tls"
echo "   2. Access at: http://localhost:3000"
echo "   3. Backstage proxy will use the configuration generated above"
echo ""
echo "🔐 TLS mode: gRPC clients must trust ${CA_BUNDLE}"
echo ""
echo "✨ Done!"
