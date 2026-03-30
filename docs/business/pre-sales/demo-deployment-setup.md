# Aegis Demo Deployment Setup

**Purpose:** Pre-configure everything needed so that the demo shows a clean `helm install` deploying the full Aegis platform on an existing EKS cluster. This documents what the `generate-cloud-deployment.sh` script normally handles automatically.

**Last Updated:** 2026-03-30

---

## Prerequisites (done once, before demo day)

### 1. AWS Infrastructure (via Terraform)

These resources must exist before the helm install:

| Resource | Purpose | How to verify |
|----------|---------|---------------|
| EKS cluster (`aegis-hub-prod`) | Hub control plane | `aws eks describe-cluster --name aegis-hub-prod` |
| IRSA role (`aegis-platform-api`) | Platform-api AWS access | `aws iam get-role --role-name aegis-platform-api` |
| ECR repos (platform-api, proxy, k8s-agent, ui, workspace-vscode, vscode-reh-init) | Container images | `aws ecr describe-repositories` |
| RDS or in-cluster Postgres | Platform database | Verify connectivity |
| Cloudflare DNS (4 CNAMEs) | Public endpoints | `dig ui.aegis-platform.tech` |
| S3 bucket for Pulumi state | Spoke provisioning state | `aws s3 ls s3://aegis-pulumi-state-195714074609/` |
| Project IAM role (`aegis-project-live-demo`) | Spoke cluster provisioning | `aws iam get-role --role-name aegis-project-live-demo` |

All of these are created by `cd terraform && terraform apply`.

### 2. Container Images (pushed to ECR)

Build and push from current HEAD:
```bash
make push-cloud-images
```

Or individually:
```bash
# From aegis-platform repo
AWS_PROFILE=aegis-new make push-cloud-images \
  TARGET_ARCH=amd64 \
  CLOUD_PLATFORM_API_IMAGE=195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/platform-api:demo \
  CLOUD_PROXY_IMAGE=195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/proxy:demo \
  CLOUD_K8S_AGENT_IMAGE=195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/k8s-agent:demo \
  CLOUD_UI_IMAGE=195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/ui:demo
```

Also ensure the REH init image exists:
```bash
# Check
aws ecr describe-images --repository-name aegis/vscode-reh-init --region us-east-1 --query 'imageDetails[0].imageTags'

# If missing, build from workspace-images/vscode-reh-init/
```

### 3. kubectl Context

```bash
export AWS_PROFILE=aegis-new
aws eks update-kubeconfig --name aegis-hub-prod --region us-east-1
kubectl get nodes  # verify access
```

---

## Pre-Helm Setup (Steps 1-5 from deploy script)

These steps prepare the cluster for the helm install. Run them once before the demo.

### Step 1: Create Namespace and Secrets

```bash
export K8S_NAMESPACE=aegis-system
export PKI_NAMESPACE=aegis-pki

# Namespace
kubectl create namespace ${K8S_NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# Database password (in-cluster Postgres)
kubectl create secret generic aegis-platform-secrets \
  -n ${K8S_NAMESPACE} \
  --from-literal=db-password="$(openssl rand -hex 16)" \
  --from-literal=proxy-jwt-secret="$(openssl rand -hex 32)" \
  --dry-run=client -o yaml | kubectl apply -f -

# Keycloak secrets (admin, db, client, tls)
KEYCLOAK_ADMIN_PASS="$(openssl rand -base64 16)"
KEYCLOAK_DB_PASS="$(openssl rand -hex 16)"
BACKSTAGE_CLIENT_SECRET="$(openssl rand -hex 32)"
SPOKE_CLIENT_SECRET="$(openssl rand -hex 32)"

kubectl create secret generic aegis-keycloak-admin \
  -n ${K8S_NAMESPACE} \
  --from-literal=admin-password="${KEYCLOAK_ADMIN_PASS}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic aegis-keycloak-db \
  -n ${K8S_NAMESPACE} \
  --from-literal=password="${KEYCLOAK_DB_PASS}" \
  --from-literal=postgres-password="${KEYCLOAK_DB_PASS}" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic aegis-keycloak-client \
  -n ${K8S_NAMESPACE} \
  --from-literal=clientSecret="${BACKSTAGE_CLIENT_SECRET}" \
  --from-literal=spokeClientSecret="${SPOKE_CLIENT_SECRET}" \
  --dry-run=client -o yaml | kubectl apply -f -
```

### Step 2: Install Internal PKI

```bash
# Create step-ca NLB first (spokes need external access)
kubectl create namespace ${PKI_NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: step-certificates-nlb
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
EOF

# Wait for NLB hostname
echo "Waiting for step-ca NLB..."
until STEP_CA_NLB=$(kubectl get svc step-certificates-nlb -n ${PKI_NAMESPACE} \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null) && [[ -n "$STEP_CA_NLB" ]]; do
  sleep 5
done
echo "Step-CA NLB: ${STEP_CA_NLB}"

# Install step-ca with NLB hostname in cert SANs
TRUST_BUNDLE_NAMESPACES="${K8S_NAMESPACE}" \
PKI_NAMESPACE="${PKI_NAMESPACE}" \
CERT_MANAGER_NAMESPACE=cert-manager \
STEP_CA_DB_PERSISTENT=false \
STEP_CA_REINSTALL_ON_MISMATCH=true \
STEP_CA_EXTERNAL_DNS_NAMES="${STEP_CA_NLB}" \
  scripts/install-internal-pki.sh
```

### Step 3: Extract PKI Credentials

```bash
# These will be passed to the helm install
export STEP_CA_URL="https://${STEP_CA_NLB}"

export STEP_PROV_KID=$(kubectl exec step-certificates-0 -n ${PKI_NAMESPACE} -- \
  cat /home/step/config/ca.json 2>/dev/null \
  | grep -o '"kid"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 \
  | sed 's/.*"kid"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')

export STEP_PROV_PASS=$(kubectl get secret step-certificates-provisioner-password \
  -n ${PKI_NAMESPACE} -o jsonpath='{.data.password}' | base64 --decode)

echo "KID: ${STEP_PROV_KID:0:12}..."
echo "URL: ${STEP_CA_URL}"

# Verify step-ca is reachable
curl -sk "${STEP_CA_URL}/health"  # should return {"status":"ok"}
```

### Step 4: Apply CRDs

```bash
kubectl apply -f agents/k8s-agent/config/crd/bases/
```

### Step 5: Refresh CA Bundle

```bash
kubectl get secret aegis-trust-bundle -n ${K8S_NAMESPACE} \
  -o "jsonpath={.data.ca\.crt}" | base64 --decode > /tmp/aegis-ca.pem

# Copy to your local machine for VS Code extension
cp /tmp/aegis-ca.pem ~/aegis-cloud-ca.pem
```

---

## The Demo Helm Install

This is what you show on camera. Everything above is pre-configured.

### Values Override File

Create `demo-values.yaml`:

```bash
ECR_REGISTRY="195714074609.dkr.ecr.us-east-1.amazonaws.com"
IMAGE_TAG="demo"  # or git SHA
DNS_UI="ui.aegis-platform.tech"
DNS_KEYCLOAK="keycloak.aegis-platform.tech"
DNS_PLATFORM_API="platform-api.aegis-platform.tech"
DNS_PROXY="proxy.aegis-platform.tech"
JWT_SECRET=$(kubectl get secret aegis-platform-secrets -n aegis-system \
  -o jsonpath='{.data.proxy-jwt-secret}' | base64 --decode)
DB_PASSWORD=$(kubectl get secret aegis-platform-secrets -n aegis-system \
  -o jsonpath='{.data.db-password}' | base64 --decode)
BACKEND_SECRET=$(openssl rand -hex 32)

cat > demo-values.yaml <<YAML
ingressController:
  enabled: true
ingress-nginx:
  controller:
    service:
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
        service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
platformApi:
  image:
    repository: ${ECR_REGISTRY}/aegis/platform-api
    tag: "${IMAGE_TAG}"
  env:
    DATABASE_URL: "postgres://aegis_platform:${DB_PASSWORD}@platform-postgres.aegis-system.svc.cluster.local:5432/aegis_platform?sslmode=disable"
    OIDC_ISSUER_URL: "https://${DNS_KEYCLOAK}/realms/aegis"
    OIDC_AUDIENCE: "backstage,aegis-platform"
    OIDC_JWKS_URL: "https://${DNS_KEYCLOAK}/realms/aegis/protocol/openid-connect/certs"
    AEGIS_PROXY_BASE_URL: "wss://${DNS_PROXY}:8080"
    AEGIS_PLATFORM_API_ENDPOINT: "${DNS_PLATFORM_API}:8081"
    AEGIS_PLATFORM_API_GRPC_INSECURE: "false"
    AEGIS_SPOKE_IMAGE_REPO: "${ECR_REGISTRY}/aegis/k8s-agent"
    AEGIS_SPOKE_IMAGE_TAG: "${IMAGE_TAG}"
    AEGIS_SPOKE_OIDC_TOKEN_URL: "https://${DNS_KEYCLOAK}/realms/aegis/protocol/openid-connect/token"
    AEGIS_SPOKE_OIDC_CLIENT_ID: "spoke-agent"
    AEGIS_SPOKE_OIDC_CLIENT_SECRET: "$(kubectl get secret aegis-keycloak-client -n aegis-system -o jsonpath='{.data.spokeClientSecret}' | base64 --decode)"
    AEGIS_SPOKE_OIDC_AUDIENCE: "aegis-platform"
    AEGIS_SPOKE_VALUES_FILE: "/home/aegis/charts/aegis-spoke/values-cloud-remote.yaml"
    AEGIS_DISCOVERY_GRPC_ENDPOINT: "${DNS_PLATFORM_API}:8081"
    # cert-manager / step-ca for spoke TLS
    AEGIS_CERT_MANAGER_ENABLED: "true"
    AEGIS_STEP_CA_URL: "${STEP_CA_URL}"
    AEGIS_CLUSTER_ISSUER_NAME: "aegis-internal"
    AEGIS_STEP_CA_ROOT_CA_FILE: "/etc/aegis-platform-api/oidc/ca.crt"
    AEGIS_STEP_PROVISIONER_KID: "${STEP_PROV_KID}"
    AEGIS_STEP_PROVISIONER_PASSWORD: "${STEP_PROV_PASS}"
    # VS Code REH init image
    AEGIS_VSCODE_REH_INIT_IMAGE: "${ECR_REGISTRY}/aegis/vscode-reh-init:latest"
  secrets:
    db-password: "${DB_PASSWORD}"
    proxy-jwt-secret: "${JWT_SECRET}"
  tls:
    certManager:
      enabled: true
      dnsNames:
        - "${DNS_PLATFORM_API}"
proxy:
  publicHost: "${DNS_PROXY}"
  image:
    repository: ${ECR_REGISTRY}/aegis/proxy
    tag: "${IMAGE_TAG}"
  jwtSecret: "${JWT_SECRET}"
  tls:
    enabled: true
    certManager:
      enabled: true
      dnsNames:
        - "${DNS_PROXY}"
backstage:
  enabled: true
  replicaCount: 1
  service:
    type: ClusterIP
    port: 7007
  ingress:
    enabled: true
    className: ingress-nginx
    annotations:
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
    hosts:
      - host: "${DNS_UI}"
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: aegis-backstage-tls
        hosts:
          - "${DNS_UI}"
  tls:
    certManager:
      enabled: true
      dnsNames:
        - "${DNS_UI}"
  image:
    repository: ${ECR_REGISTRY}/aegis/ui
    tag: "${IMAGE_TAG}"
  appConfig:
    appBaseUrl: "https://${DNS_UI}"
    backendBaseUrl: "https://${DNS_UI}"
  postgres:
    host: "platform-postgres.aegis-system.svc.cluster.local"
    port: "5432"
    user: "aegis_platform"
  keycloak:
    baseUrl: "https://${DNS_KEYCLOAK}"
    realm: "aegis"
    clientId: "backstage"
    clientSecret:
      secretName: aegis-keycloak-client
      key: clientSecret
  secrets:
    create: true
    backendSecret: "${BACKEND_SECRET}"
keycloak:
  enabled: true
  forceRender: true
  namespace: aegis-system
  hostname:
    hostname: "https://${DNS_KEYCLOAK}"
    admin: "https://${DNS_KEYCLOAK}"
    strict: false
  ingress:
    enabled: false
  customIngress:
    enabled: true
    className: ingress-nginx
    annotations:
      nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
  http:
    httpEnabled: false
  admin:
    secret:
      name: aegis-keycloak-admin
      create: false
  database:
    secret:
      name: aegis-keycloak-db
      create: false
  realm:
    client:
      secret:
        name: aegis-keycloak-client
        create: false
  tls:
    secret:
      name: aegis-keycloak-tls
      create: false
    certManager:
      enabled: true
      dnsNames:
        - "${DNS_KEYCLOAK}"
YAML

echo "demo-values.yaml generated"
```

### The One-Liner (what you show on camera)

```bash
helm upgrade --install aegis ./charts/aegis-services \
  -f charts/aegis-services/values/common.yaml \
  -f charts/aegis-services/values/cloud.yaml \
  -f demo-values.yaml \
  --namespace aegis-system --create-namespace \
  --timeout 10m
```

---

## Post-Deploy Verification (off-camera or quick check)

```bash
# All pods running
kubectl get pods -n aegis-system

# Endpoints reachable
curl -sk https://ui.aegis-platform.tech | head -5
curl -sk https://keycloak.aegis-platform.tech/realms/aegis/.well-known/openid-configuration | head -5
curl -sk https://platform-api.aegis-platform.tech:8080/api/v1/discovery | python3 -m json.tool

# Step-CA accessible from spokes
curl -sk "${STEP_CA_URL}/health"  # {"status":"ok"}
```

---

## Teardown (reset for next demo run)

```bash
# Delete helm release (keeps secrets and PKI)
helm uninstall aegis -n aegis-system

# Full teardown including PKI
helm uninstall aegis -n aegis-system
helm uninstall step-certificates -n aegis-pki
helm uninstall step-issuer -n cert-manager
helm uninstall cert-manager -n cert-manager
kubectl delete namespace aegis-system aegis-pki cert-manager
```

---

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| Keycloak CrashLoop | TLS cert not issued | Check `kubectl get certificate -n aegis-system` — StepClusterIssuer must be Ready |
| Platform-api can't start | Missing DB password secret | Verify `aegis-platform-secrets` exists with `db-password` key |
| Spoke StepClusterIssuer "failed initialize" | step-ca NLB not reachable or wrong CA | Verify `curl -sk ${STEP_CA_URL}/health` returns ok |
| VS Code "version mismatch" | REH init image tag doesn't match local VS Code | Rebuild vscode-reh-init image matching your VS Code version |
| VS Code "unable to get local issuer" | Local CA file stale | Re-download: `kubectl exec step-certificates-0 -n aegis-pki -- cat /home/step/certs/root_ca.crt > ~/aegis-cloud-ca.pem` |
