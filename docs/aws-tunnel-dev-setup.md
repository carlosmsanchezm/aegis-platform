# AWS Tunnel Development Setup

This guide covers connecting remote AWS EKS spoke clusters to your local Docker Desktop platform during development.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Local (Docker Desktop)                                                  │
│  ┌──────────────┐     ┌─────────────┐     ┌─────────────────────────┐   │
│  │ Platform-API │◄────┤ Port-Forward├─────┤ SSH Reverse Tunnel      │   │
│  │ (gRPC:8081)  │     │ (8081)      │     │ to AWS Relay EC2       │   │
│  └──────────────┘     └─────────────┘     └───────────┬─────────────┘   │
│                                                        │                 │
│  ┌──────────────┐     ┌─────────────────────────────────────────────┐   │
│  │ Keycloak     │◄────┤ Cloudflare Tunnel (keycloak.aegis-...)      │   │
│  │ (HTTPS:8443) │     │ (HTTPS works fine through Cloudflare)       │   │
│  └──────────────┘     └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ AWS NLB (TCP passthrough)
                                    │ Port 8081 → SSH Tunnel → Local
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  AWS EKS Cluster                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │ aegis-spoke (k8s-agent)                                           │   │
│  │  - gRPC → NLB:8081 → SSH tunnel → local platform-api             │   │
│  │  - OIDC → Cloudflare → local Keycloak                            │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

## Why This Architecture?

| Traffic | Route | Reason |
|---------|-------|--------|
| gRPC (platform-api) | AWS NLB → SSH tunnel | Cloudflare doesn't handle gRPC streaming properly |
| OIDC (Keycloak) | Cloudflare tunnel | HTTPS works fine; NLB port 8443 often conflicts |

## Prerequisites

1. **AWS Relay Infrastructure** deployed:
   ```bash
   cd terraform/pulumi-stack && terraform apply
   ```

2. **SSH Key** for relay access:
   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/aegis-relay -N ""
   # Add public key to terraform.tfvars: relay_ssh_public_key = "..."
   ```

3. **Cloudflare Tunnel** running (for Keycloak):
   ```bash
   kubectl get pods -n aegis-system -l app=cloudflared
   ```

4. **Remote cluster kubeconfig**:
   ```bash
   # Create kubeconfig using AWS CLI (not aegis-eks-token)
   # See: /tmp/remote-kubeconfig-aegis.yaml
   ```

## Setup Steps

### Step 1: Start the AWS Tunnel

```bash
./scripts/start-aws-tunnel.sh
```

This script:
1. Starts `kubectl port-forward` to local platform-api (8081)
2. Establishes SSH reverse tunnel to AWS relay EC2
3. **Generates `charts/aegis-spoke/values-aws-relay.yaml`** with correct endpoints

Keep this running in a terminal.

### Step 2: Deploy/Update Spoke Agent

```bash
KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml helm upgrade aegis-spoke \
  ./charts/aegis-spoke -n aegis-system \
  --reset-values \
  -f charts/aegis-spoke/values-aws-relay.yaml \
  --set k8sAgent.env.AEGIS_CLUSTER_ID=<cluster-id> \
  --set k8sAgent.env.AEGIS_REGION=us-east-1 \
  --set k8sAgent.env.AEGIS_PROVIDER=aws
```

**Important:** Always use `--reset-values` to clear any stale Cloudflare endpoints from previous deployments.

### Step 3: Verify Connection

```bash
# Check k8s-agent logs
KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml kubectl logs \
  -l app.kubernetes.io/component=k8s-agent -n aegis-system --tail=20

# Check platform-api for cluster registration
kubectl --context docker-desktop logs \
  -l app.kubernetes.io/component=platform-api -n aegis-system --tail=20 \
  | grep "cluster registered"
```

### Step 4: Stop Tunnel (When Done)

```bash
./scripts/stop-aws-tunnel.sh
```

## Troubleshooting

### "server closed the stream without sending trailers"

**Cause:** k8s-agent is using Cloudflare endpoint for gRPC instead of NLB.

**Fix:** Re-run helm upgrade with `--reset-values`:
```bash
KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml helm upgrade aegis-spoke \
  ./charts/aegis-spoke -n aegis-system \
  --reset-values \
  -f charts/aegis-spoke/values-aws-relay.yaml \
  --set k8sAgent.env.AEGIS_CLUSTER_ID=<cluster-id> \
  --set k8sAgent.env.AEGIS_REGION=us-east-1 \
  --set k8sAgent.env.AEGIS_PROVIDER=aws
```

### Pulumi keeps overwriting helm values with Cloudflare

**Cause:** Platform-api has `AEGIS_PLATFORM_API_ENDPOINT` set to Cloudflare. The Pulumi runner (runner.go line 997) **overrides** `AEGIS_CP_GRPC` in the values file with this env var.

**Fix:** Update platform-api's env var to use NLB:
```bash
kubectl set env deploy/aegis-services-platform-api -n aegis-system \
  AEGIS_PLATFORM_API_ENDPOINT="aegis-dev-relay-nlb-353e7f4c0ac5c4be.elb.us-east-1.amazonaws.com:8081"
```

Or update `charts/aegis-services/values/local-tls.yaml`:
```yaml
env:
  AEGIS_PLATFORM_API_ENDPOINT: "aegis-dev-relay-nlb-353e7f4c0ac5c4be.elb.us-east-1.amazonaws.com:8081"
```

### Helm upgrade stuck in "pending-upgrade"

**Cause:** Pulumi is running a helm operation.

**Fix:** Wait or rollback:
```bash
KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml helm rollback aegis-spoke -n aegis-system
```

### OIDC token fetch timeout

**Cause:** Keycloak Cloudflare tunnel not running or slow.

**Fix:** Check cloudflared pod:
```bash
kubectl --context docker-desktop logs -n aegis-system deployment/cloudflared
```

### Port 8443 already in use

**Note:** This is expected. The script tries to forward Keycloak via NLB but we use Cloudflare instead. The error can be ignored.

### Connection reset by peer on NLB

**Cause:** SSH tunnel or port-forward died.

**Fix:** Restart the tunnel:
```bash
./scripts/stop-aws-tunnel.sh
./scripts/start-aws-tunnel.sh
```

## Files Reference

| File | Purpose |
|------|---------|
| `scripts/start-aws-tunnel.sh` | Starts tunnel, generates values file |
| `scripts/stop-aws-tunnel.sh` | Stops tunnel and port-forwards |
| `charts/aegis-spoke/values-aws-relay.yaml` | Auto-generated helm values |
| `/tmp/remote-kubeconfig-aegis.yaml` | Remote EKS cluster access |

## Enabling Spoke Proxy (Multi-Cluster Workspace Connectivity)

The spoke proxy allows VS Code extension to connect to workspaces running in remote clusters.

### Architecture

```
VS Code Extension → Platform-API (local) → returns spoke proxy URL
                           ↓
VS Code Extension → Spoke Proxy (AWS) → Workspace Pod (AWS)
```

### Setup

1. **Generate self-signed TLS certs** (one-time):
   ```bash
   openssl req -x509 -newkey rsa:2048 \
     -keyout /tmp/proxy-key.pem -out /tmp/proxy-cert.pem \
     -days 365 -nodes -subj "/CN=spoke-proxy.aegis.local"
   ```

2. **Get node external IP**:
   ```bash
   KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml kubectl get nodes \
     -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}'
   # Example: 107.22.50.137
   ```

3. **Deploy spoke with proxy enabled**:
   ```bash
   KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml helm upgrade aegis-spoke \
     ./charts/aegis-spoke -n aegis-system \
     --reset-values \
     -f charts/aegis-spoke/values-aws-relay.yaml \
     --set k8sAgent.env.AEGIS_CLUSTER_ID=db-1-us-east-1-atlas-train-govcloud \
     --set k8sAgent.env.AEGIS_REGION=us-east-1 \
     --set k8sAgent.env.AEGIS_PROVIDER=aws \
     --set proxy.enabled=true \
     --set proxy.url="wss://spoke-proxy.107.22.50.137.nip.io:31484" \
     --set proxy.ingress.hostname="spoke-proxy.107.22.50.137.nip.io" \
     --set-file proxy.tls.cert=/tmp/proxy-cert.pem \
     --set-file proxy.tls.key=/tmp/proxy-key.pem
   ```

4. **Verify heartbeats include proxy URL**:
   ```bash
   KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml kubectl logs \
     -l app.kubernetes.io/component=k8s-agent -n aegis-system --tail=20 \
     | grep "proxy_url"
   # Should show: proxy_url":"wss://spoke-proxy.107.22.50.137.nip.io:31484
   ```

### Key Values

| Value | Purpose | Example |
|-------|---------|---------|
| `proxy.enabled` | Enable spoke proxy deployment | `true` |
| `proxy.url` | Full URL for heartbeat (with port) | `wss://spoke-proxy.107.22.50.137.nip.io:31484` |
| `proxy.ingress.hostname` | Hostname only (no port) | `spoke-proxy.107.22.50.137.nip.io` |
| `proxy.service.nodePort` | Fixed NodePort | `31484` |

### Testing with VS Code Extension

1. Open VS Code with the Aegis extension
2. Sign in (uses Cloudflare → Keycloak)
3. Select a workspace in the remote cluster
4. Connect - the extension will receive the spoke proxy URL from platform-api

The platform-api's `buildSessionContext()` checks the cluster's proxy URL and returns the spoke proxy URL for workspaces in that cluster.

### Spoke Proxy Troubleshooting

**"exec format error" on proxy pod**
- Cause: Proxy image architecture mismatch (arm64 vs amd64)
- Fix: Rebuild with `make build-proxy` (creates multi-arch image)

**Ingress hostname validation error with port**
- Cause: K8s ingress hostnames can't include ports
- Fix: Use `proxy.url` for full URL with port, `proxy.ingress.hostname` for hostname only

**Workspace DNS resolution fails**
- Cause: Hub proxy can't resolve remote cluster DNS
- Fix: Enable spoke proxy - it runs in the same cluster as workspaces

---

## Full Cluster Reconnection Guide

Use this checklist when spinning up the EKS cluster after it's been stopped.

### Prerequisites
- AWS profile `aegis` configured with role `arn:aws:iam::567751785679:role/aegis-platform`
- Docker Desktop running with Kubernetes enabled
- Local aegis-services deployed

### Step 1: Start Local Services

```bash
# Ensure Docker Desktop K8s is running
kubectl get pods -n aegis-system

# Start the AWS tunnel (establishes NLB connectivity)
./scripts/start-aws-tunnel.sh
```

### Step 2: Generate EKS Kubeconfig

```bash
# Get cluster credentials
aws eks update-kubeconfig \
  --name db-1-us-east-1-atlas-train-govcloud \
  --region us-east-1 \
  --profile aegis \
  --role-arn arn:aws:iam::567751785679:role/aegis-platform \
  --kubeconfig /tmp/remote-kubeconfig-aegis.yaml

# Test access
KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml kubectl get nodes
```

### Step 3: Add Kubeconfig to Platform-API Secret (CRITICAL)

> **The secret key MUST match the `AEGIS_CLUSTER_ID` exactly!**

```bash
# Define cluster ID (used in both spoke and kubeconfig secret)
CLUSTER_ID="db-1-us-east-1-atlas-train-govcloud"

# Create kubeconfig for platform-api (uses aws CLI for token)
cat > /tmp/eks-kubeconfig.yaml << 'EOF'
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: <BASE64_CA_DATA_FROM_EKS>
    server: https://<EKS_API_ENDPOINT>
  name: eks-cluster
contexts:
- context:
    cluster: eks-cluster
    user: eks-cluster
  name: eks-cluster
current-context: eks-cluster
kind: Config
users:
- name: eks-cluster
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - --region
      - us-east-1
      - eks
      - get-token
      - --cluster-name
      - db-1-us-east-1-atlas-train-govcloud
      - --role-arn
      - arn:aws:iam::567751785679:role/aegis-platform
      - --output
      - json
      command: aws
EOF

# Patch the secret (key name = CLUSTER_ID)
kubectl patch secret aegis-kubeconfigs -n aegis-system --type='json' -p='[
  {"op": "add", "path": "/data/'$CLUSTER_ID'", "value": "'$(base64 -i /tmp/eks-kubeconfig.yaml | tr -d '\n')'"}
]'

# Restart platform-api to pick up new kubeconfig
kubectl rollout restart deployment aegis-services-platform-api -n aegis-system
```

### Step 4: Deploy Spoke Agent on EKS

```bash
CLUSTER_ID="db-1-us-east-1-atlas-train-govcloud"

KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml helm upgrade aegis-spoke \
  ./charts/aegis-spoke -n aegis-system --create-namespace \
  --reset-values \
  -f charts/aegis-spoke/values-aws-relay.yaml \
  --set k8sAgent.env.AEGIS_CLUSTER_ID=$CLUSTER_ID \
  --set k8sAgent.env.AEGIS_REGION=us-east-1 \
  --set k8sAgent.env.AEGIS_PROVIDER=aws
```

### Step 5: Enable Spoke Proxy (for VS Code connectivity)

```bash
# Get node external IP
NODE_IP=$(KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml kubectl get nodes \
  -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
echo "Node IP: $NODE_IP"

# Generate self-signed TLS certs (if needed)
openssl req -x509 -newkey rsa:2048 \
  -keyout /tmp/proxy-key.pem -out /tmp/proxy-cert.pem \
  -days 365 -nodes -subj "/CN=spoke-proxy.${NODE_IP}.nip.io"

# Deploy with proxy enabled
KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml helm upgrade aegis-spoke \
  ./charts/aegis-spoke -n aegis-system \
  --reset-values \
  -f charts/aegis-spoke/values-aws-relay.yaml \
  --set k8sAgent.env.AEGIS_CLUSTER_ID=$CLUSTER_ID \
  --set k8sAgent.env.AEGIS_REGION=us-east-1 \
  --set k8sAgent.env.AEGIS_PROVIDER=aws \
  --set proxy.enabled=true \
  --set proxy.service.nodePort=31484 \
  --set proxy.url="wss://spoke-proxy.${NODE_IP}.nip.io:31484" \
  --set proxy.ingress.hostname="spoke-proxy.${NODE_IP}.nip.io" \
  --set-file proxy.tls.cert=/tmp/proxy-cert.pem \
  --set-file proxy.tls.key=/tmp/proxy-key.pem
```

### Step 6: Configure VS Code Extension Trust

The VS Code extension needs the local CA certificate:

```bash
# Location of local trust cert
ls -la ~/aegis-local-trust.pem

# VS Code extension settings should have:
# "security.caPath": "/Users/carlossanchez/aegis-local-trust.pem"
```

### Verification

```bash
# Check spoke agent is registering
kubectl logs -n aegis-system -l app.kubernetes.io/name=platform-api --tail=20 | grep "cluster registered"

# Check heartbeats include proxy URL
KUBECONFIG=/tmp/remote-kubeconfig-aegis.yaml kubectl logs \
  -l app.kubernetes.io/component=k8s-agent -n aegis-system --tail=20 | grep proxy

# Check kubeconfig secret has correct key
kubectl get secret aegis-kubeconfigs -n aegis-system -o json | jq '.data | keys'
# Should include: "db-1-us-east-1-atlas-train-govcloud"
```

### Quick Reference

| Component | Value |
|-----------|-------|
| Cluster ID | `db-1-us-east-1-atlas-train-govcloud` |
| AWS Profile | `aegis` |
| IAM Role | `arn:aws:iam::567751785679:role/aegis-platform` |
| EKS Region | `us-east-1` |
| Spoke Proxy Port | `31484` |
| Local Trust Cert | `~/aegis-local-trust.pem` |

### Critical Rule

**The kubeconfig secret key MUST exactly match the `AEGIS_CLUSTER_ID`:**
- Spoke: `--set k8sAgent.env.AEGIS_CLUSTER_ID=db-1-us-east-1-atlas-train-govcloud`
- Secret key: `/data/db-1-us-east-1-atlas-train-govcloud`

If these don't match, you'll get `failed to resolve cluster client`.
