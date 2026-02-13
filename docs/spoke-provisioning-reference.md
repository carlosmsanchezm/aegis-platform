# Spoke Cluster Provisioning Reference

This document provides a complete reference for provisioning AWS EKS spoke clusters from the Aegis platform. It covers all required components, configurations, and connectivity requirements.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Prerequisites Checklist](#prerequisites-checklist)
3. [Component Configuration](#component-configuration)
4. [Connectivity Flow](#connectivity-flow)
5. [Environment Variables Reference](#environment-variables-reference)
6. [Helm Chart Configuration](#helm-chart-configuration)
7. [Verification Commands](#verification-commands)
8. [Troubleshooting](#troubleshooting)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            LOCAL DEVELOPMENT MACHINE                             │
│                                                                                  │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │                     Docker Desktop Kubernetes                            │   │
│  │                                                                          │   │
│  │  ┌─────────────────────┐     ┌──────────────────────────────────────┐  │   │
│  │  │ platform-api        │     │ Pulumi Provisioner                    │  │   │
│  │  │ (aegis-system)      │     │ (runs inside platform-api)            │  │   │
│  │  │                     │     │                                        │  │   │
│  │  │ Ports:              │     │ 1. Creates EKS cluster                 │  │   │
│  │  │   8080 (HTTP)       │     │ 2. Deploys aegis-spoke helm chart      │  │   │
│  │  │   8081 (gRPC)       │     │ 3. Stores kubeconfig in secret         │  │   │
│  │  └─────────┬───────────┘     └──────────────────────────────────────┘  │   │
│  │            │                                                            │   │
│  └────────────┼────────────────────────────────────────────────────────────┘   │
│               │                                                                  │
│  ┌────────────┴────────────┐                                                    │
│  │ Port-forwards           │                                                    │
│  │  localhost:8081 ──────────────────────────────────────┐                      │
│  │  localhost:10080 (HTTP) │                              │                      │
│  │  localhost:10081 (gRPC) │                              │                      │
│  └─────────────────────────┘                              │                      │
│                                                           │                      │
│  ┌────────────────────────────────────────────────────────┼─────────────────┐   │
│  │ SSH Tunnel                                             │                  │   │
│  │ ssh -R 0.0.0.0:8081:localhost:8081 ec2-user@relay ─────┘                  │   │
│  │                                                                           │   │
│  └───────────────────────────────────┬───────────────────────────────────────┘   │
│                                      │                                           │
└──────────────────────────────────────┼───────────────────────────────────────────┘
                                       │
                                       ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                                    AWS VPC                                        │
│                                                                                   │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │ aegis-dev-relay EC2                                                          │ │
│  │ (i-08cde33f38cb9a911)                                                        │ │
│  │                                                                               │ │
│  │ Public IP: 67.202.30.169                                                     │ │
│  │ Receives SSH tunnel, listens on 0.0.0.0:8081                                 │ │
│  └──────────────────────────────────────┬──────────────────────────────────────┘ │
│                                         │                                         │
│  ┌──────────────────────────────────────┴──────────────────────────────────────┐ │
│  │ NLB: aegis-dev-relay-nlb                                                     │ │
│  │ DNS: aegis-dev-relay-nlb-12123e67db2b2af4.elb.us-east-1.amazonaws.com       │ │
│  │ Type: Internal (only accessible within VPC)                                  │ │
│  │ Forwards port 8081 TCP to relay EC2                                          │ │
│  └──────────────────────────────────────┬──────────────────────────────────────┘ │
│                                         │                                         │
│  ┌──────────────────────────────────────┴──────────────────────────────────────┐ │
│  │ EKS Spoke Cluster                                                            │ │
│  │                                                                               │ │
│  │  ┌─────────────────────────────────────────────────────────────────────┐    │ │
│  │  │ aegis-spoke (helm release)                                           │    │ │
│  │  │                                                                       │    │ │
│  │  │  ┌─────────────────┐    ┌─────────────────┐                         │    │ │
│  │  │  │ k8s-agent       │    │ spoke-proxy     │                         │    │ │
│  │  │  │                 │    │ (optional)      │                         │    │ │
│  │  │  │ Connects to NLB │    │                 │                         │    │ │
│  │  │  │ for gRPC        │    │ For workspace   │                         │    │ │
│  │  │  │                 │    │ connections     │                         │    │ │
│  │  │  └─────────────────┘    └─────────────────┘                         │    │ │
│  │  └─────────────────────────────────────────────────────────────────────┘    │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                   │
└───────────────────────────────────────────────────────────────────────────────────┘
```

### Data Flow Summary

1. **Provisioning Request**: UI → platform-api → Pulumi
2. **EKS Creation**: Pulumi → AWS EKS API
3. **Spoke Deployment**: Pulumi → helm install aegis-spoke on EKS
4. **Agent Registration**: k8s-agent → NLB → SSH tunnel → localhost:8081 → platform-api
5. **Heartbeats**: Same path as registration, every 10 seconds

---

## Prerequisites Checklist

Before provisioning a spoke cluster, verify ALL of these components:

### 1. AWS Credentials
```bash
# Verify AWS credentials work from platform-api pod
kubectl --context docker-desktop exec -n aegis-system deploy/aegis-services-platform-api -- aws sts get-caller-identity

# Expected: Returns IAM user/role details
```

### 2. Pulumi S3 Backend
```bash
# Verify Pulumi can access S3 state bucket
kubectl --context docker-desktop exec -n aegis-system deploy/aegis-services-platform-api -- aws s3 ls s3://aegis-pulumi-state-dev/aegis/pulumi/

# Expected: Lists .pulumi/ directory
```

### 3. SSH Tunnel to Relay
```bash
# Check SSH tunnel is running
ps aux | grep "ssh.*aegis-relay\|ssh.*8081" | grep -v grep

# Expected output example:
# ssh -R 0.0.0.0:8081:localhost:8081 -N ec2-user@67.202.30.169
```

### 4. Relay EC2 Instance
```bash
# Check relay instance is running
aws ec2 describe-instances \
  --filters "Name=tag:Name,Values=*aegis*relay*" \
  --region us-east-1 \
  --query 'Reservations[*].Instances[*].{ID:InstanceId,State:State.Name,PublicIP:PublicIpAddress}' \
  --output table

# Expected: State = running
```

### 5. NLB Target Health
```bash
# Check NLB targets are healthy
aws elbv2 describe-target-health \
  --target-group-arn $(aws elbv2 describe-target-groups \
    --load-balancer-arn $(aws elbv2 describe-load-balancers \
      --names aegis-dev-relay-nlb \
      --region us-east-1 \
      --query 'LoadBalancers[0].LoadBalancerArn' \
      --output text) \
    --region us-east-1 \
    --query 'TargetGroups[0].TargetGroupArn' \
    --output text) \
  --region us-east-1

# Expected: State = healthy
```

### 6. Port-forwards
```bash
# Check required port-forwards are running
ps aux | grep "port-forward" | grep -v grep

# Required port-forwards:
# - 8081:8081   (for SSH tunnel relay connection)
# - 10080:8080  (for local HTTP access)
# - 10081:8081  (for local gRPC access)
```

### 7. Cloudflared Tunnel (for Keycloak)
```bash
# Check cloudflared is running
kubectl --context docker-desktop get pods -n aegis-system | grep cloudflared

# Expected: 1/1 Running
```

---

## Component Configuration

### Platform-API Environment Variables

These environment variables control spoke provisioning behavior:

| Variable | Description | Current Value |
|----------|-------------|---------------|
| `AEGIS_PLATFORM_API_ENDPOINT` | gRPC endpoint for spoke agents | `aegis-dev-relay-nlb-*.elb.us-east-1.amazonaws.com:8081` |
| `AEGIS_PLATFORM_API_GRPC_INSECURE` | Skip TLS for gRPC | `true` |
| `AEGIS_SPOKE_VALUES_FILE` | Helm values file path | `/root/charts/aegis-spoke/values-aws-relay.yaml` |
| `AEGIS_SPOKE_CHART` | Path to helm chart | (auto-detected: `/root/charts/aegis-spoke`) |
| `AWS_REGION` | Default AWS region | `us-east-1` |
| `PULUMI_BACKEND_URL` | Pulumi state storage | `s3://aegis-pulumi-state-dev/aegis/pulumi` |

### Verify Current Configuration
```bash
kubectl --context docker-desktop exec -n aegis-system deploy/aegis-services-platform-api -- env | grep -E "^AEGIS_|^AWS_|^PULUMI_" | sort
```

---

## Connectivity Flow

### Spoke Agent → Platform-API (gRPC)

```
k8s-agent (EKS)
    │
    │ gRPC connection to AEGIS_CP_GRPC
    ▼
NLB (internal): aegis-dev-relay-nlb-*.elb.us-east-1.amazonaws.com:8081
    │
    │ TCP passthrough
    ▼
Relay EC2 (67.202.30.169:8081)
    │
    │ SSH reverse tunnel
    ▼
localhost:8081 (your machine)
    │
    │ kubectl port-forward
    ▼
platform-api:8081 (Docker Desktop)
```

### Spoke Agent → Keycloak (OIDC)

```
k8s-agent (EKS)
    │
    │ HTTPS to AEGIS_CP_OIDC_TOKEN_URL
    ▼
Cloudflare Tunnel: keycloak.aegis-platform.tech
    │
    │ Tunnel to local cluster
    ▼
keycloak service (Docker Desktop)
```

---

## Environment Variables Reference

### Platform-API (Provisioner)

| Variable | Purpose | Example |
|----------|---------|---------|
| `AEGIS_PLATFORM_API_ENDPOINT` | Endpoint spoke agents connect to | `aegis-dev-relay-nlb-*.amazonaws.com:8081` |
| `AEGIS_PLATFORM_API_CA_FILE` | CA cert for TLS (optional) | `/path/to/ca.crt` |
| `AEGIS_PLATFORM_API_GRPC_INSECURE` | Skip TLS verification | `true` |
| `AEGIS_SPOKE_CHART` | Helm chart path | `/root/charts/aegis-spoke` |
| `AEGIS_SPOKE_VALUES_FILE` | Base values file | `/root/charts/aegis-spoke/values-aws-relay.yaml` |
| `AEGIS_SPOKE_REPO` | Helm repository URL (if using remote chart) | `https://charts.example.com` |
| `AEGIS_SPOKE_CHART_VERSION` | Chart version (if using remote chart) | `1.0.0` |
| `AEGIS_SPOKE_IMAGE_REPO` | Override k8s-agent image repo | `carlosmsanchez/aegis-k8s-agent` |
| `AEGIS_SPOKE_IMAGE_TAG` | Override k8s-agent image tag | `dev` |
| `AEGIS_SKIP_SPOKE_HELM` | Skip helm deployment (debug) | `false` |
| `PULUMI_BACKEND_URL` | Pulumi state backend | `s3://bucket/path` |
| `PULUMI_CONFIG_PASSPHRASE` | Pulumi encryption passphrase | `aegis-local-passphrase` |
| `AWS_ACCESS_KEY_ID` | AWS credentials | (from secret) |
| `AWS_SECRET_ACCESS_KEY` | AWS credentials | (from secret) |
| `AWS_REGION` | Default region | `us-east-1` |

### Spoke Agent (k8s-agent)

These are set via Helm values and/or Pulumi overrides:

| Variable | Purpose | Set By |
|----------|---------|--------|
| `AEGIS_CLUSTER_ID` | Unique cluster identifier | Pulumi (dynamic) |
| `AEGIS_REGION` | AWS region | Pulumi (dynamic) |
| `AEGIS_PROVIDER` | Cloud provider | Pulumi (`aws`) |
| `AEGIS_CP_GRPC` | Platform-API gRPC endpoint | Pulumi (from `AEGIS_PLATFORM_API_ENDPOINT`) |
| `AEGIS_CP_GRPC_INSECURE` | Skip TLS for gRPC | Values file |
| `AEGIS_CP_GRPC_SKIP_VERIFY` | Skip cert verification | Values file |
| `AEGIS_CP_OIDC_TOKEN_URL` | Keycloak token endpoint | Values file |
| `AEGIS_CP_OIDC_CLIENT_ID` | OIDC client ID | Values file (`spoke-agent`) |
| `AEGIS_CP_OIDC_CLIENT_SECRET` | OIDC client secret | Values file |
| `AEGIS_CP_OIDC_AUDIENCE` | Expected token audience | Values file (`aegis-platform`) |
| `AEGIS_FLAVORS` | Advertised compute flavors | Values file |
| `AEGIS_DEFAULT_IMAGE` | Default workspace image | Values file |

---

## Helm Chart Configuration

### Chart Location

The aegis-spoke chart is embedded in the platform-api Docker image:

```
/root/charts/aegis-spoke/
├── Chart.yaml
├── templates/
│   ├── k8s-agent-deployment.yaml
│   ├── k8s-agent-rbac.yaml
│   ├── proxy-deployment.yaml (optional)
│   └── ...
├── values.yaml                 # Default values
├── values-aws-relay.yaml       # AWS relay configuration
├── values-local.yaml           # Local development
└── values-local-tls.yaml       # Local with TLS
```

### Values Precedence

When Pulumi deploys the spoke, values are merged in this order (later overrides earlier):

1. `values.yaml` (chart defaults)
2. `values-aws-relay.yaml` (via `AEGIS_SPOKE_VALUES_FILE`)
3. Pulumi runtime values (dynamic overrides)

### Critical Runtime Overrides

Pulumi automatically sets these values at deployment time:

```yaml
k8sAgent:
  env:
    AEGIS_CLUSTER_ID: "<cluster-id>"           # From infra spec
    AEGIS_REGION: "<region>"                   # From infra spec
    AEGIS_PROVIDER: "aws"                      # Hardcoded
    AEGIS_CP_GRPC: "<nlb-endpoint>:8081"       # From AEGIS_PLATFORM_API_ENDPOINT
    AEGIS_PLATFORM_API_ENDPOINT: "<endpoint>"  # From env var
```

### values-aws-relay.yaml Structure

```yaml
k8sAgent:
  image:
    pullPolicy: Always
    tag: "dev"
  env:
    # gRPC endpoint (overridden by Pulumi with correct NLB)
    AEGIS_CP_GRPC: "aegis-dev-relay-nlb-*.amazonaws.com:8081"
    AEGIS_CP_GRPC_INSECURE: "false"
    AEGIS_CP_GRPC_SKIP_VERIFY: "true"

    # Compute flavors advertised by this cluster
    AEGIS_FLAVORS: "cpu-small,cpu-medium,cpu-large,gpu-standard,gpu-large"

    # Default workspace image
    AEGIS_DEFAULT_IMAGE: "docker.io/carlosmsanchez/aegis-workspace-vscode:latest"

    # OIDC configuration (via Cloudflare)
    AEGIS_CP_OIDC_TOKEN_URL: "https://keycloak.aegis-platform.tech/realms/aegis/protocol/openid-connect/token"
    AEGIS_CP_OIDC_CLIENT_ID: "spoke-agent"
    AEGIS_CP_OIDC_CLIENT_SECRET: "<secret>"
    AEGIS_CP_OIDC_AUDIENCE: "aegis-platform"
    AEGIS_CP_OIDC_SKIP_TLS_VERIFY: "true"

proxy:
  enabled: false  # Enable for workspace connections
```

---

## Verification Commands

### Quick Health Check (All-in-One)

```bash
echo "=== Spoke Provisioning Readiness Check ==="

echo -n "AWS Credentials: "
kubectl --context docker-desktop exec -n aegis-system deploy/aegis-services-platform-api -- aws sts get-caller-identity >/dev/null 2>&1 && echo "OK" || echo "FAILED"

echo -n "Pulumi Backend: "
kubectl --context docker-desktop exec -n aegis-system deploy/aegis-services-platform-api -- aws s3 ls s3://aegis-pulumi-state-dev/aegis/pulumi/ >/dev/null 2>&1 && echo "OK" || echo "FAILED"

echo -n "SSH Tunnel: "
ps aux | grep -q "ssh.*8081.*relay\|ssh.*aegis-relay" && echo "OK" || echo "NOT RUNNING"

echo -n "Port-forward 8081: "
lsof -i :8081 >/dev/null 2>&1 && echo "OK" || echo "NOT RUNNING"

echo -n "Port-forward 10080: "
lsof -i :10080 >/dev/null 2>&1 && echo "OK" || echo "NOT RUNNING"

echo -n "NLB Health: "
aws elbv2 describe-target-health \
  --target-group-arn $(aws elbv2 describe-target-groups \
    --load-balancer-arn $(aws elbv2 describe-load-balancers \
      --names aegis-dev-relay-nlb --region us-east-1 \
      --query 'LoadBalancers[0].LoadBalancerArn' --output text) \
    --region us-east-1 --query 'TargetGroups[0].TargetGroupArn' --output text) \
  --region us-east-1 --query 'TargetHealthDescriptions[0].TargetHealth.State' --output text 2>/dev/null

echo -n "Cloudflared: "
kubectl --context docker-desktop get pods -n aegis-system -l app=cloudflared -o jsonpath='{.items[0].status.phase}' 2>/dev/null || echo "NOT FOUND"
```

### Check Spoke Agent Logs (After Provisioning)

```bash
# Get cluster kubeconfig
CLUSTER_ID="<your-cluster-id>"

# Check spoke agent logs via platform-api
kubectl --context docker-desktop exec -n aegis-system deploy/aegis-services-platform-api -- \
  kubectl --kubeconfig /tmp/kubeconfigs/${CLUSTER_ID}.kubeconfig \
  -n aegis-system logs deploy/aegis-spoke-k8s-agent --tail=50
```

### Check Cluster Registration in Database

```bash
kubectl --context docker-desktop exec platform-postgres-0 -n aegis-system -- \
  env PGPASSWORD=platform_local_pass psql -U aegis_platform -d aegis_platform \
  -c "SELECT id, project_id, last_heartbeat, proxy_url FROM clusters ORDER BY created_at DESC LIMIT 5;"
```

---

## Troubleshooting

### Symptom: Cluster Stuck in "Provisioning"

**Check Pulumi logs:**
```bash
kubectl --context docker-desktop logs deploy/aegis-services-platform-api -n aegis-system --tail=200 | grep -E "(pulumi|provision|error)"
```

**Check provisioning database:**
```bash
kubectl --context docker-desktop exec platform-postgres-0 -n aegis-system -- \
  env PGPASSWORD=platform_local_pass psql -U aegis_platform -d aegis_platform \
  -c "SELECT job_id, phase, cluster_id FROM provisioning_runs ORDER BY started_at DESC LIMIT 5;"
```

### Symptom: Spoke Agent Not Registering

**Common causes:**

1. **SSH tunnel not running**
   ```bash
   # Check tunnel
   ps aux | grep "ssh.*8081" | grep -v grep

   # Restart tunnel
   ssh -i ~/.ssh/aegis-relay -o StrictHostKeyChecking=no \
     -R 0.0.0.0:8081:localhost:8081 -N ec2-user@67.202.30.169 &
   ```

2. **Port-forward not running**
   ```bash
   # Check port-forward
   lsof -i :8081

   # Start port-forward
   kubectl --context docker-desktop -n aegis-system \
     port-forward svc/aegis-services-platform-api 8081:8081 &
   ```

3. **NLB target unhealthy**
   ```bash
   # Check NLB health (see Verification Commands section)
   ```

4. **Wrong NLB endpoint in spoke**
   ```bash
   # Check what endpoint the spoke is using
   kubectl --kubeconfig /tmp/kubeconfigs/<cluster-id>.kubeconfig \
     -n aegis-system get deploy aegis-spoke-k8s-agent \
     -o jsonpath='{.spec.template.spec.containers[0].env}' | jq '.[] | select(.name | contains("GRPC"))'
   ```

### Symptom: "ResourceInUseException: Cluster already exists"

**Cause:** Pulumi was interrupted mid-provisioning (e.g., pod restart).

**Fix:** The platform-api now includes automatic recovery logic. If it still fails:

```bash
# 1. Delete the failed ProjectInfra
kubectl --context docker-desktop delete projectinfra <infra-name> -n aegis-system

# 2. Clean up database entries
kubectl --context docker-desktop exec platform-postgres-0 -n aegis-system -- \
  env PGPASSWORD=platform_local_pass psql -U aegis_platform -d aegis_platform \
  -c "DELETE FROM provisioning_runs WHERE job_id = '<job-id>';"

# 3. If cluster exists in AWS but not in Pulumi state, delete from AWS
aws eks delete-cluster --name <cluster-name> --region us-east-1

# 4. Retry provisioning via UI
```

### Symptom: OIDC Authentication Failures

**Check Keycloak accessibility:**
```bash
curl -sk https://keycloak.aegis-platform.tech/realms/aegis/.well-known/openid-configuration | head -5
```

**Check spoke-agent client exists in Keycloak:**
- Go to Keycloak admin console
- Verify `spoke-agent` client exists in `aegis` realm
- Verify client secret matches `values-aws-relay.yaml`

---

## Quick Start Commands

### Start All Required Components

```bash
# 1. Start port-forwards
pkill -f "kubectl.*port-forward" 2>/dev/null || true
kubectl --context docker-desktop -n aegis-system port-forward svc/aegis-services-platform-api 8081:8081 10080:8080 10081:8081 &
kubectl --context docker-desktop -n aegis-system port-forward svc/aegis-services-proxy 10085:8085 &
kubectl --context docker-desktop -n keycloak port-forward svc/aegis-services-keycloak-service 10443:8443 &

# 2. Start SSH tunnel (if not running)
if ! ps aux | grep -q "ssh.*aegis-relay.*8081"; then
  ssh -i ~/.ssh/aegis-relay -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
    -R 0.0.0.0:8081:localhost:8081 -N ec2-user@67.202.30.169 &
fi

# 3. Verify
sleep 3
echo "Port 8081: $(lsof -i :8081 >/dev/null 2>&1 && echo 'OK' || echo 'FAILED')"
echo "Port 10080: $(lsof -i :10080 >/dev/null 2>&1 && echo 'OK' || echo 'FAILED')"
```

### Stop All Components

```bash
pkill -f "kubectl.*port-forward"
pkill -f "ssh.*aegis-relay"
```

---

## Important Notes

1. **Don't restart platform-api during provisioning** - EKS cluster creation takes 15-20 minutes. Restarting the pod mid-provisioning causes state sync issues.

2. **NLB is internal** - The relay NLB is only accessible from within the AWS VPC. Your local machine connects via SSH tunnel, not directly.

3. **Values file vs runtime overrides** - The `values-aws-relay.yaml` may have stale NLB endpoints. Pulumi runtime values (from `AEGIS_PLATFORM_API_ENDPOINT`) always take precedence.

4. **Keycloak via Cloudflare, gRPC via NLB** - Cloudflare tunnels don't support gRPC streaming properly. That's why we use the NLB+SSH tunnel for gRPC and Cloudflare only for Keycloak OIDC.

---

*Last updated: January 2026*
