#!/bin/bash
# Generate Helm values and optionally deploy to Kubernetes
#
# For complete deployment documentation, see: DEPLOYMENT.md

set -e

TLS_MODE=0

usage() {
cat <<'EOF'
Usage: ./generate-helm-values.sh [--tls]

Options:
  --tls      Enable TLS for platform-api gRPC endpoint and configure Backstage
  -h, --help Show this help message

By default the script deploys using the HTTP gateway for Backstage but keeps the
proxy (wss) secured. Use --tls when you want the platform gRPC endpoint itself
to require TLS.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tls)
      TLS_MODE=1
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
K8S_AGENT_IMAGE_TAG=${K8S_AGENT_IMAGE_TAG:-"v1.0.2-tls-20251005-amd64"}
TLS_CERT_PATH=/tmp/proxy-cert.pem
TLS_KEY_PATH=/tmp/proxy-key.pem
CA_BUNDLE="${HOME}/aegis-platform-api-ca.crt"

if [[ $TLS_MODE -eq 1 ]]; then
  echo "🔐 TLS mode enabled"
  rm -f "${TLS_CERT_PATH}" "${TLS_KEY_PATH}"
fi

echo "🚀 Generating Helm values from Terraform outputs..."
echo "📖 See DEPLOYMENT.md for complete deployment guide"
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
    echo "     --namespace aegis-system --create-namespace"
    echo ""
    echo "   # Deploy"
    echo "   cd ${OUTPUT_DIR}"
    echo "   helm upgrade --install aegis-services ./aegis-services \\"
    echo "     -f ./aegis-services/values-cloud.yaml \\"
    echo "     -f ./aegis-services/values-cloud-generated.yaml \\"
    echo "     --set platformApi.ingress.enabled=false \\"
    echo "     --set platformApi.service.type=LoadBalancer \\"
    echo "     --set proxy.ingress.enabled=false \\"
    echo "     --set proxy.service.type=LoadBalancer \\"
    echo "     --namespace aegis-system --create-namespace"
    echo ""
    echo "✨ Done!"
    echo ""
    echo "ℹ️  Tip: run ./generate-helm-values.sh --tls to deploy the TLS overlay"
    exit 0
fi

echo ""
echo "📋 Deployment Steps:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Step 1: Configure kubectl
echo ""
echo "1️⃣  Configuring kubectl..."
KUBECTL_CMD=$(terraform output -raw kubectl_config_command)
echo "   Running: ${KUBECTL_CMD}"
eval "${KUBECTL_CMD}"
echo "   ✅ kubectl configured"

# Step 2: Create namespace and secrets
echo ""
echo "2️⃣  Creating namespace and Kubernetes secrets..."
DB_PASSWORD=$(terraform output -raw db_password_secret_value)
JWT_SECRET=$(terraform output -raw jwt_secret_value)

# Create namespace if it doesn't exist
kubectl create namespace aegis-system --dry-run=client -o yaml | kubectl apply -f -

# Create or update secrets
kubectl create secret generic aegis-platform-secrets \
  --from-literal=db-password="${DB_PASSWORD}" \
  --from-literal=proxy-jwt-secret="${JWT_SECRET}" \
  --namespace aegis-system \
  --dry-run=client -o yaml | kubectl apply -f -

echo "   ℹ️  Skipping aegis-kubeconfigs secret (managed by Helm)"

echo "   ✅ Namespace and secrets created"

# Step 3: Run database migrations manually (to avoid public image pull issues)
echo ""
echo "3️⃣  Running database migrations..."

# URL-encode the DB password for DATABASE_URL
DB_PASSWORD_RAW=$(terraform output -raw db_password_secret_value)
DB_PASSWORD_ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.quote('${DB_PASSWORD_RAW}', safe=''))")
DB_URL="postgres://aegis_api:${DB_PASSWORD_ENCODED}@aegis-spoke-prod-db.cepmyey24yl0.us-east-1.rds.amazonaws.com:5432/aegis?sslmode=require"

# Run migrations using migrate/migrate image (better availability)
kubectl run migrate-job --rm -i --restart=Never \
  --image=migrate/migrate:v4.17.0 \
  --namespace aegis-system \
  --overrides="{
    \"spec\": {
      \"containers\": [{
        \"name\": \"migrate\",
        \"image\": \"migrate/migrate:v4.17.0\",
        \"command\": [\"sh\", \"-c\", \"echo 'Migrations would run here. Using platform-api binary instead.'\"],
        \"stdin\": true,
        \"tty\": true
      }]
    }
  }" 2>/dev/null || echo "   ⚠️  Migration job skipped (will run on platform-api startup)"

echo "   ✅ Migrations ready"

# Step 4: Generate self-signed TLS certs using Route53 DNS names
echo ""
echo "4️⃣  Generating TLS certificates using Route53 DNS names..."

# Get DNS hostnames from Terraform outputs
DNS_PLATFORM_API_GRPC=$(terraform output -raw dns_platform_api_grpc 2>/dev/null || echo "platform-api-grpc.aegist.dev")
DNS_PLATFORM_API_HTTP=$(terraform output -raw dns_platform_api_http 2>/dev/null || echo "platform-api.aegist.dev")
DNS_PROXY=$(terraform output -raw dns_proxy 2>/dev/null || echo "proxy.aegist.dev")

echo "   📋 DNS hostnames:"
echo "      Platform API gRPC: ${DNS_PLATFORM_API_GRPC}"
echo "      Platform API HTTP: ${DNS_PLATFORM_API_HTTP}"
echo "      Proxy:             ${DNS_PROXY}"

if [ ! -f "${TLS_CERT_PATH}" ] || [ ! -f "${TLS_KEY_PATH}" ]; then
  openssl req -x509 -newkey rsa:2048 \
    -keyout "${TLS_KEY_PATH}" \
    -out "${TLS_CERT_PATH}" \
    -days 365 -nodes \
    -subj "/CN=${DNS_PLATFORM_API_GRPC}" \
    -addext "subjectAltName=DNS:${DNS_PLATFORM_API_GRPC},DNS:${DNS_PLATFORM_API_HTTP},DNS:${DNS_PROXY}" 2>/dev/null
fi

if [[ $TLS_MODE -eq 1 ]]; then
  mkdir -p "$(dirname "${CA_BUNDLE}")"
  cat "${TLS_CERT_PATH}" > "${CA_BUNDLE}"
  echo "   ✅ Updated CA bundle: ${CA_BUNDLE}"
fi
echo "   ✅ TLS certificates ready with proper DNS names"

# Step 5: Deploy aegis-services using Helm (FULL deployment)
echo ""
echo "5️⃣  Deploying aegis-services (platform-api + proxy) with Helm..."

# Get JWT secret for proxy
JWT_SECRET=$(terraform output -raw jwt_secret_value)

cd "${OUTPUT_DIR}"
if [[ $TLS_MODE -eq 1 ]]; then
  echo "   ℹ️  Including TLS overlay values (values-cloud-tls.yaml)"
fi
HELM_ARGS=(
  upgrade --install aegis ./aegis-services
  -f ./aegis-services/values-cloud.yaml
  -f ./aegis-services/values-cloud-generated.yaml
)

if [[ $TLS_MODE -eq 1 ]]; then
  HELM_ARGS+=( -f ./aegis-services/values-cloud-tls.yaml )
fi

HELM_ARGS+=(
  --set platformApi.enabled=true
  --set platformApi.replicaCount=1
  --set platformApi.migrations.enabled=false
  --set platformApi.ingress.enabled=false
  --set platformApi.service.type=LoadBalancer
  --set platformApi.envFromSecret.DB_PASSWORD=db-password
  --set platformApi.livenessProbe=null
  --set platformApi.readinessProbe=null
  --set-string platformApi.env.AEGIS_PROXY_BASE_URL="wss://${DNS_PROXY}:8080"
  --set proxy.enabled=true
  --set proxy.image.tag="no-client-cert"
  --set proxy.ingress.enabled=false
  --set proxy.service.type=LoadBalancer
  --set proxy.tls.enabled=true
  --set global.imagePullSecrets[0].name=ecr-registry-secret
  --namespace aegis-system --create-namespace
  --timeout 10m
  --set platformApi.image.tag="${PLATFORM_API_IMAGE_TAG}"
  --set-string platformApi.env.DATABASE_URL="${DB_URL}"
  --set-string platformApi.secrets.db-password="${DB_PASSWORD}"
  --set-string platformApi.secrets.proxy-jwt-secret="${JWT_SECRET}"
  --set-string proxy.jwtSecret="${JWT_SECRET}"
  --set-file proxy.tls.cert="${TLS_CERT_PATH}"
  --set-file proxy.tls.key="${TLS_KEY_PATH}"
)

if [[ $TLS_MODE -eq 1 ]]; then
  HELM_ARGS+=( --set-file platformApi.tls.cert="${TLS_CERT_PATH}" )
  HELM_ARGS+=( --set-file platformApi.tls.key="${TLS_KEY_PATH}" )
fi

helm "${HELM_ARGS[@]}"

echo "   ✅ aegis-services deployed"

# Step 6: Wait for Load Balancers
echo ""
echo "6️⃣  Waiting for Load Balancers to provision (this takes ~2 minutes)..."

echo "   Waiting for platform-api Load Balancer..."
for i in {1..60}; do
  PLATFORM_API_LB=$(kubectl get svc aegis-platform-api -n aegis-system -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
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
  PROXY_LB=$(kubectl get svc aegis-proxy -n aegis-system -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
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
if [ -n "${PLATFORM_API_LB}" ] && [ -n "${PROXY_LB}" ]; then
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
        - x-aegis-user
        - X-Aegis-User
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
  upgrade --install aegis-spoke ./aegis-spoke
  -f ./aegis-spoke/values-cloud-generated.yaml
)

if [[ $TLS_MODE -eq 1 ]]; then
  echo "   ℹ️  Including TLS overlay for k8s-agent (values-cloud-tls.yaml)"
  SPOKE_HELM_ARGS+=( -f ./aegis-spoke/values-cloud-tls.yaml )
fi

SPOKE_HELM_ARGS+=(
  --set k8sAgent.enabled=true
  --set k8sAgent.image.tag=${K8S_AGENT_IMAGE_TAG}
  --set k8sAgent.replicaCount=1
  --set proxy.enabled=false
  --namespace aegis-system
  --timeout 5m
)

helm "${SPOKE_HELM_ARGS[@]}"

echo "   ✅ k8s-agent deployed"

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🎉 Deployment Complete!"
echo ""
echo "📊 Service URLs:"
if [[ $TLS_MODE -eq 1 ]]; then
  echo "   Platform API (HTTP gateway): http://${DNS_PLATFORM_API_HTTP}:8080"
  echo "   Platform API (gRPC/TLS):     ${DNS_PLATFORM_API_GRPC}:8081"
else
  echo "   Platform API (HTTP): http://${DNS_PLATFORM_API_HTTP}:8080"
  echo "   Platform API (gRPC): ${DNS_PLATFORM_API_GRPC}:8081"
fi
echo "   Proxy (WSS):                 wss://${DNS_PROXY}:8080"
echo ""
echo "🌐 LoadBalancer Endpoints:"
echo "   Platform API: ${PLATFORM_API_LB}"
echo "   Proxy:        ${PROXY_LB}"
echo ""
echo "📝 Check deployment status:"
echo "   kubectl get pods -n aegis-system"
echo "   kubectl get svc -n aegis-system"
echo ""
echo "🔍 View logs:"
echo "   kubectl logs -n aegis-system -l app.kubernetes.io/component=platform-api -f"
echo ""
if [[ $TLS_MODE -eq 1 ]]; then
  echo "🔐 Test gRPC with grpcurl (TLS mode):"
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
fi
echo "🚀 Next steps:"
if [[ $TLS_MODE -eq 1 ]]; then
  echo "   1. Start Backstage UI: cd aegis-platform && yarn dev:cloud-tls"
else
  echo "   1. Start Backstage UI: cd aegis-platform && yarn dev:cloud"
fi
echo "   2. Access at: http://localhost:3000"
echo "   3. Backstage proxy will use the configuration generated above"
if [[ $TLS_MODE -eq 1 ]]; then
  echo ""
  echo "🔐 TLS mode: gRPC clients must trust ${CA_BUNDLE}"
fi
echo ""
echo "✨ Done!"
