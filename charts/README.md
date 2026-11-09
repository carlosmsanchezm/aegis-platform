# Aegis Helm Charts

This directory contains Helm charts for deploying Aegis services.

For deployment commands and step-by-step guides, see [AGENT_DEPLOYMENT_GUIDE.md](../AGENT_DEPLOYMENT_GUIDE.md).

## Charts

### aegis-services
Control plane services (platform-api + proxy)

### aegis-spoke
Spoke/agent components (k8s-agent) that run in workload clusters

---

## Values File Structure

### aegis-services

```
aegis-services/
├── Chart.yaml
├── values/
│   ├── common.yaml              # Base defaults (all environments)
│   ├── local.yaml               # Local Docker Desktop config (no TLS)
│   ├── local-tls.yaml           # Local TLS overlay (pod TLS + cert-manager)
│   ├── local-pki.yaml           # Internal PKI trust bundle overlay
│   └── cloud.yaml               # Cloud base config (LoadBalancer + pod TLS)
└── values-cloud-generated.yaml  # Auto-generated (gitignored)
```

### aegis-spoke

```
aegis-spoke/
├── Chart.yaml
├── values.yaml                  # Base spoke defaults
├── values-local.yaml            # Local spoke config
├── values-local-tls.yaml        # Local TLS connection setting
├── values-local-pki.yaml        # Internal PKI trust bundle overlay
├── values-cloud-tls.yaml        # Cloud TLS connection setting
└── values-cloud-generated.yaml  # Auto-generated (gitignored)
```

---

## File Purposes

### aegis-services/values/common.yaml
**Size:** 8.5KB
**Purpose:** Shared defaults across all environments
**Contains:** Image repos, resource limits, security contexts, probes

### aegis-services/values/local.yaml
**Size:** 6.3KB
**Purpose:** Local Docker Desktop deployment
**Contains:** `tls.enabled: false`, ClusterIP services, debug logging, relaxed limits

### aegis-services/values/cloud.yaml
**Size:** 6.7KB
**Purpose:** Cloud base configuration (LoadBalancer + TLS to pods)
**Contains:** LoadBalancer services, production replicas, TLS-enabled proxy, kubeconfig setup. Uses in-cluster Postgres by default; RDS is optional (see `docs/PRODUCTION_DEPLOYMENT.md`).

### aegis-services/values-cloud-generated.yaml
**Auto-generated** by `terraform output -raw helm_values_aegis_services`
**Contains:** DB credentials, JWT secret, DNS names, LoadBalancer annotations

### aegis-spoke/values.yaml
**Size:** 2.2KB
**Purpose:** Base spoke configuration
**Contains:** k8s-agent defaults, RBAC, resources

### aegis-spoke/values-local.yaml
**Size:** 379B
**Purpose:** Local spoke overrides
**Contains:** Cluster ID, platform-api endpoint (in-cluster), region, flavors

### aegis-spoke/values-cloud-tls.yaml
**Size:** 249B
**Purpose:** Enable secure gRPC connection
**Contains:** TLS client settings for the spoke agent (insecure=false, skip-verify toggles)

### aegis-spoke/values-cloud-generated.yaml
**Auto-generated** by `terraform output -raw helm_values_aegis_spoke`
**Contains:** Platform-api gRPC endpoint (with DNS), cluster ID, region, flavors, CA cert

---

## Layering & Merge Order

Helm merges values files from left to right, with later files overriding earlier ones.

### Local TLS Deployment

```
Layer 1: values/common.yaml          (base defaults)
Layer 2: values/local.yaml           (local config, ClusterIP)
Layer 3: values/local-tls.yaml       (TLS overlay)
Layer 4: values/local-pki.yaml       (PKI trust bundles)
         ↓
    Final Config: TLS enabled; cert-manager manages TLS secrets
```

### Cloud Deployment

```
Layer 1: values/common.yaml          (base defaults)
Layer 2: values/cloud.yaml           (cloud config, LoadBalancer + TLS)
Layer 3: values-cloud-generated.yaml (terraform outputs)
Layer 4: OVERRIDE_FILE               (temp: secrets, images, cert-manager SANs)
         ↓
    Final Config: TLS enabled; cert-manager manages TLS secrets
```

**Key insight:** `cloud.yaml` and Terraform outputs describe the infrastructure; cert-manager mints and rotates certificates from the configured issuer.

### Helm Direct Usage

Both aegis-services AND aegis-spoke must be deployed together for a working stack.

**aegis-services (platform-api, proxy, keycloak, ingress):**

```bash
helm upgrade --install aegis-services charts/aegis-services \
  -f charts/aegis-services/values/common.yaml \
  -f charts/aegis-services/values/local.yaml \
  -f charts/aegis-services/values/local-pki.yaml \
  -f charts/aegis-services/values/local-tls.yaml \
  --set hardeningProfile=dev \
  --set platformApi.image.repository=carlosmsanchez/aegis-platform-api \
  --set platformApi.image.tag=dev \
  --namespace aegis-system --create-namespace \
  --wait --timeout 5m
```

**aegis-spoke (k8s-agent, spoke proxy):**

All four values files are required for TLS to work:

```bash
helm upgrade --install aegis-spoke charts/aegis-spoke \
  -f charts/aegis-spoke/values.yaml \
  -f charts/aegis-spoke/values-local.yaml \
  -f charts/aegis-spoke/values-local-tls.yaml \
  -f charts/aegis-spoke/values-local-pki.yaml \
  --set k8sAgent.image.repository=carlosmsanchez/aegis-k8s-agent \
  --set k8sAgent.image.tag=dev \
  --set k8sAgent.image.pullPolicy=Always \
  --set hardeningProfile=dev \
  --namespace aegis-system --create-namespace
```

WARNING: Omitting `values-local-tls.yaml` or `values-local-pki.yaml` will cause
the agent to fail with "error reading server preface: EOF" because it cannot
establish a TLS connection to the platform-api.

---

## Common Tasks

### Update Base Defaults
Edit `aegis-services/values/common.yaml`

### Update Local Config
Edit `aegis-services/values/local.yaml`

### Update Cloud Config
Edit `aegis-services/values/cloud.yaml`

### Update TLS Settings
Edit `platformApi.tls.certManager` / `proxy.tls.certManager` (and rerun the deployment script to update SANs/overrides as needed)

### Regenerate Cloud Values
```bash
cd terraform
terraform output -raw helm_values_aegis_services > ../charts/aegis-services/values-cloud-generated.yaml
terraform output -raw helm_values_aegis_spoke > ../charts/aegis-spoke/values-cloud-generated.yaml
```

---

## Troubleshooting

### Check What Values Are Actually Used

```bash
# Dry-run to see merged values
helm install --dry-run --debug aegis ./aegis-services \
  -f ./aegis-services/values/common.yaml \
  -f ./aegis-services/values/cloud.yaml \
  -f ./aegis-services/values-cloud-generated.yaml \
  -f overrides.yaml \
  --namespace aegis-system
```

### Verify TLS Status

```bash
# Platform API
kubectl get deployment aegis-services-platform-api -n aegis-system \
  -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="GRPC_TLS_ENABLED")].value}'

# Expected: "true" when TLS enabled
```

### Check Generated Values

```bash
cat charts/aegis-services/values-cloud-generated.yaml
cat charts/aegis-spoke/values-cloud-generated.yaml
```

---

## Migration from Old Structure

Previously, there were duplicate values files:
- ❌ `values.yaml` (388-line FedRAMP config, unused)
- ❌ `values-local.yaml` (duplicate of `values/local.yaml`)
- ❌ `values-cloud.yaml` (duplicate of `values/cloud.yaml`)

These have been cleaned up. All deployments continue to work with the streamlined structure.
