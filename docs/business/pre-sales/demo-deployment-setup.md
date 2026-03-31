# Aegis Demo Deployment Setup

**Purpose:** Prepare the environment so the demo shows a single `helm install` deploying Aegis on an existing EKS cluster. Uses the deploy script for initial setup, then tears down only the helm release so the on-camera install is clean.

**Last Updated:** 2026-03-30

---

## Overview

```
Phase 1: Full deploy via script      (before demo, off-camera)
Phase 2: Capture values + uninstall  (before demo, off-camera)
Phase 3: helm install + DNS update   (on-camera demo)
```

---

## Phase 1: Full Deploy (off-camera)

Run the standard cloud deployment. This sets up everything: Terraform, secrets, PKI, DNS, helm release.

```bash
export AWS_PROFILE=aegis-new
export CLOUDFLARE_API_TOKEN="<your-token>"

# Build and push images from current HEAD
make push-cloud-images TARGET_ARCH=amd64

# Full deploy (terraform + secrets + PKI + helm + DNS)
make deploy-cloud
```

Verify everything works:
```bash
# UI loads
curl -sk https://ui.aegis-platform.tech | grep -o "<title>.*</title>"

# Platform-api discovery
curl -sk http://platform-api.aegis-platform.tech:8080/api/v1/discovery | python3 -m json.tool | head -10

# Step-CA health
STEP_CA_NLB=$(kubectl get svc step-certificates-nlb -n aegis-pki \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
curl -sk "https://${STEP_CA_NLB}/health"
```

---

## Phase 2: Capture Values and Uninstall (off-camera)

### 2a. Save the helm values

The deploy script generates a temporary override file. We need to recreate it with all the dynamic values that were computed during Phase 1.

```bash
#!/bin/bash
# Run this BEFORE helm uninstall — it reads from the running deployment
set -euo pipefail

export AWS_PROFILE=aegis-new
K8S_NAMESPACE=aegis-system
PKI_NAMESPACE=aegis-pki

# --- Read live values from the running deployment ---
DNS_PLATFORM_API="platform-api.aegis-platform.tech"
DNS_KEYCLOAK="keycloak.aegis-platform.tech"
DNS_PROXY="proxy.aegis-platform.tech"
DNS_UI="ui.aegis-platform.tech"
ECR_REGISTRY="195714074609.dkr.ecr.us-east-1.amazonaws.com"

# Image tag from running platform-api
IMAGE_TAG=$(kubectl get deployment aegis-platform-api -n ${K8S_NAMESPACE} \
  -o jsonpath='{.spec.template.spec.containers[0].image}' | awk -F: '{print $2}')

# Secrets
JWT_SECRET=$(kubectl get secret aegis-platform-secrets -n ${K8S_NAMESPACE} \
  -o jsonpath='{.data.proxy-jwt-secret}' | base64 --decode)
DB_PASSWORD=$(kubectl get secret aegis-platform-secrets -n ${K8S_NAMESPACE} \
  -o jsonpath='{.data.db-password}' | base64 --decode)
SPOKE_CLIENT_SECRET=$(kubectl get secret aegis-keycloak-client -n ${K8S_NAMESPACE} \
  -o jsonpath='{.data.spokeClientSecret}' | base64 --decode)

# PKI credentials
STEP_CA_NLB=$(kubectl get svc step-certificates-nlb -n ${PKI_NAMESPACE} \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
STEP_CA_URL="https://${STEP_CA_NLB}"
STEP_PROV_KID=$(kubectl exec step-certificates-0 -n ${PKI_NAMESPACE} -- \
  cat /home/step/config/ca.json 2>/dev/null \
  | grep -o '"kid"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 \
  | sed 's/.*"kid"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
STEP_PROV_PASS=$(kubectl get secret step-certificates-provisioner-password \
  -n ${PKI_NAMESPACE} -o jsonpath='{.data.password}' | base64 --decode)

BACKEND_SECRET=$(openssl rand -hex 32)

echo "Image tag:    ${IMAGE_TAG}"
echo "Step-CA URL:  ${STEP_CA_URL}"
echo "Prov KID:     ${STEP_PROV_KID:0:12}..."

# --- Generate demo-values.yaml ---
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
    pullPolicy: Always
  env:
    DATABASE_URL: "postgres://aegis_platform:${DB_PASSWORD}@platform-postgres.${K8S_NAMESPACE}.svc.cluster.local:5432/aegis_platform?sslmode=disable"
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
    AEGIS_SPOKE_OIDC_CLIENT_SECRET: "${SPOKE_CLIENT_SECRET}"
    AEGIS_SPOKE_OIDC_AUDIENCE: "aegis-platform"
    AEGIS_SPOKE_VALUES_FILE: "/home/aegis/charts/aegis-spoke/values-cloud-remote.yaml"
    AEGIS_DISCOVERY_GRPC_ENDPOINT: "${DNS_PLATFORM_API}:8081"
    AEGIS_CERT_MANAGER_ENABLED: "true"
    AEGIS_STEP_CA_URL: "${STEP_CA_URL}"
    AEGIS_CLUSTER_ISSUER_NAME: "aegis-internal"
    AEGIS_STEP_CA_ROOT_CA_FILE: "/etc/aegis-platform-api/oidc/ca.crt"
    AEGIS_STEP_PROVISIONER_KID: "${STEP_PROV_KID}"
    AEGIS_STEP_PROVISIONER_PASSWORD: "${STEP_PROV_PASS}"
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
    pullPolicy: Always
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
    pullPolicy: Always
  appConfig:
    appBaseUrl: "https://${DNS_UI}"
    backendBaseUrl: "https://${DNS_UI}"
  postgres:
    host: "platform-postgres.${K8S_NAMESPACE}.svc.cluster.local"
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
  namespace: ${K8S_NAMESPACE}
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

echo ""
echo "demo-values.yaml generated ($(wc -l < demo-values.yaml) lines)"
echo "Verify: grep AEGIS_STEP_CA_URL demo-values.yaml"
```

Save this as `scripts/generate-demo-values.sh` and run it.

### 2b. Uninstall the helm release

```bash
helm uninstall aegis -n aegis-system --wait
```

### 2c. Verify prerequisites survived

Everything the helm install needs should still be in place:

```bash
# Secrets (created manually, not by helm)
kubectl get secret aegis-platform-secrets -n aegis-system
kubectl get secret aegis-keycloak-admin -n aegis-system
kubectl get secret aegis-keycloak-db -n aegis-system
kubectl get secret aegis-keycloak-client -n aegis-system

# PKI (separate helm releases, untouched)
kubectl get pods -n aegis-pki          # step-ca running
kubectl get pods -n cert-manager       # cert-manager running
kubectl get stepclusterissuer          # aegis-internal Ready

# Trust bundle and PKI init job secrets (helm hooks, survive uninstall)
kubectl get secret aegis-trust-bundle -n aegis-system
kubectl get secret step-ca-credentials -n aegis-system
kubectl get secret aegis-platform-api-oidc-ca -n aegis-system

# CRDs
kubectl get crd aegisworkloads.aegis.io

# Step-CA NLB (created via kubectl, not helm)
kubectl get svc step-certificates-nlb -n aegis-pki

# demo-values.yaml exists
cat demo-values.yaml | head -5
```

If anything is missing, re-run `generate-cloud-deployment.sh` and start Phase 2 over.

---

## Phase 3: Demo (on-camera)

### The Helm Install

```bash
helm upgrade --install aegis ./charts/aegis-services \
  -f charts/aegis-services/values/common.yaml \
  -f charts/aegis-services/values/cloud.yaml \
  -f demo-values.yaml \
  --namespace aegis-system --create-namespace \
  --timeout 10m
```

This takes ~3-5 minutes. Services come up in this order:
1. ingress-nginx controller (NLB created)
2. Keycloak DB + Keycloak (StatefulSets)
3. Platform-api + Proxy (Deployments, NLBs created)
4. Backstage (waits for Keycloak)
5. PKI init job runs (copies trust bundle + provisioner creds)
6. TLS certificates issued by step-ca

### Update DNS (run in a second terminal while helm is installing)

The new NLBs have different hostnames. Update Cloudflare as soon as they appear:

```bash
#!/bin/bash
# Run this in a second terminal during the helm install
set -euo pipefail

export AWS_PROFILE=aegis-new
K8S_NAMESPACE=aegis-system
CF_API_TOKEN="${CLOUDFLARE_API_TOKEN}"
CF_ZONE_ID="$(cd terraform && terraform output -raw cloudflare_zone_id)"

update_dns() {
  local name="$1" target="$2"
  local record_id=$(curl -sS "https://api.cloudflare.com/client/v4/zones/${CF_ZONE_ID}/dns_records?name=${name}&type=CNAME" \
    -H "Authorization: Bearer ${CF_API_TOKEN}" | python3 -c "import sys,json; r=json.load(sys.stdin)['result']; print(r[0]['id'] if r else '')")
  if [[ -n "${record_id}" ]]; then
    curl -sS -X PATCH "https://api.cloudflare.com/client/v4/zones/${CF_ZONE_ID}/dns_records/${record_id}" \
      -H "Authorization: Bearer ${CF_API_TOKEN}" \
      -H "Content-Type: application/json" \
      --data "{\"content\":\"${target}\",\"proxied\":false}" | python3 -c "import sys,json; print('OK' if json.load(sys.stdin).get('success') else 'FAIL')"
    echo "  ${name} -> ${target}"
  fi
}

echo "Waiting for NLBs..."

# Wait for platform-api NLB
until PLATFORM_API_LB=$(kubectl get svc aegis-platform-api -n ${K8S_NAMESPACE} \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null) && [[ -n "$PLATFORM_API_LB" ]]; do sleep 5; done
echo "Platform-API NLB: ${PLATFORM_API_LB}"

# Wait for proxy NLB
until PROXY_LB=$(kubectl get svc aegis-proxy -n ${K8S_NAMESPACE} \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null) && [[ -n "$PROXY_LB" ]]; do sleep 5; done
echo "Proxy NLB: ${PROXY_LB}"

# Wait for ingress NLB
until INGRESS_LB=$(kubectl get svc -n ${K8S_NAMESPACE} \
  -l app.kubernetes.io/component=controller,app.kubernetes.io/name=ingress-nginx \
  -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}' 2>/dev/null) && [[ -n "$INGRESS_LB" ]]; do sleep 5; done
echo "Ingress NLB: ${INGRESS_LB}"

echo ""
echo "Updating DNS..."
update_dns "platform-api.aegis-platform.tech" "${PLATFORM_API_LB}"
update_dns "proxy.aegis-platform.tech" "${PROXY_LB}"
update_dns "ui.aegis-platform.tech" "${INGRESS_LB}"
update_dns "keycloak.aegis-platform.tech" "${INGRESS_LB}"
echo "DNS updated. Allow 30-60s for propagation."
```

### Post-Install Verification

Once helm completes and DNS propagates:

```bash
# All pods running
kubectl get pods -n aegis-system

# UI accessible
open https://ui.aegis-platform.tech

# Discovery endpoint (VS Code extension uses this)
curl -sk http://platform-api.aegis-platform.tech:8080/api/v1/discovery | python3 -m json.tool | head -5

# Step-CA still healthy
curl -sk "https://$(kubectl get svc step-certificates-nlb -n aegis-pki \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')/health"
```

---

## What Survives helm uninstall (and why)

| Resource | Survives? | Reason |
|----------|-----------|--------|
| aegis-system namespace | Yes | Not helm-managed |
| aegis-platform-secrets | Yes | Created by kubectl, not helm |
| Keycloak secrets (admin, db, client) | Yes | Created by kubectl |
| step-ca (aegis-pki namespace) | Yes | Separate helm release |
| cert-manager (cert-manager namespace) | Yes | Separate helm release |
| step-issuer | Yes | Separate helm release |
| StepClusterIssuer (aegis-internal) | Yes | Created by install-internal-pki.sh |
| aegis-trust-bundle | Yes | Created by install-internal-pki.sh |
| step-ca-credentials | Yes | Helm hook (not deleted on uninstall) |
| aegis-platform-api-oidc-ca | Yes | Helm hook (not deleted on uninstall) |
| CRDs (AegisWorkload) | Yes | Applied via kubectl |
| step-ca NLB service | Yes | Created via kubectl |
| Platform-api/Proxy NLBs | **No** | Created by helm, deleted on uninstall |
| Ingress-nginx NLB | **No** | Created by helm, deleted on uninstall |
| TLS certificates | **No** | Created by helm, re-issued on install |

---

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `ui.aegis-platform.tech` not loading after install | DNS points to old NLBs | Run the DNS update script |
| Keycloak CrashLoop | TLS cert not yet issued | Wait 30s, cert-manager is issuing from step-ca |
| Platform-api can't start | Missing secret | Check `kubectl get secret aegis-platform-secrets -n aegis-system` |
| StepClusterIssuer not Ready after uninstall | Shouldn't happen — it's not helm-managed | `kubectl describe stepclusterissuer aegis-internal` |
| VS Code "unable to get local issuer" | Local CA file stale | `kubectl exec step-certificates-0 -n aegis-pki -- cat /home/step/certs/root_ca.crt > ~/aegis-cloud-ca.pem` |
| VS Code "version mismatch" | REH image doesn't match VS Code | Rebuild vscode-reh-init matching your VS Code version |
