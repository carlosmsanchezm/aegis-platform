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

### OIDC token fetch timeout

**Cause:** Keycloak Cloudflare tunnel not running or slow.

**Fix:** Check cloudflared pod:
```bash
kubectl --context docker-desktop logs -n aegis-system deployment/cloudflared
```

### Port 8443 already in use

**Note:** This is expected. The script tries to forward Keycloak via NLB but we use Cloudflare instead. The error can be ignored.

## Files Reference

| File | Purpose |
|------|---------|
| `scripts/start-aws-tunnel.sh` | Starts tunnel, generates values file |
| `scripts/stop-aws-tunnel.sh` | Stops tunnel and port-forwards |
| `charts/aegis-spoke/values-aws-relay.yaml` | Auto-generated helm values |
| `/tmp/remote-kubeconfig-aegis.yaml` | Remote EKS cluster access |

## Enabling Spoke Proxy

To test the multi-cluster spoke proxy architecture, add these to the helm command:

```bash
--set proxy.enabled=true \
--set proxy.ingress.hostname="spoke-proxy.<node-ip>.nip.io" \
--set-file proxy.tls.cert=/path/to/cert.pem \
--set-file proxy.tls.key=/path/to/key.pem
```

See the spoke proxy architecture documentation for details.
