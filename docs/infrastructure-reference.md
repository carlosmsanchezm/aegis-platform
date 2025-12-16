# Aegis Platform Infrastructure Reference

**Last Updated:** December 14, 2025
**Purpose:** Quick reference for understanding the complete local development infrastructure setup

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Kubernetes Namespaces and Services](#kubernetes-namespaces-and-services)
3. [Network Endpoints and Ports](#network-endpoints-and-ports)
4. [AWS Spoke Cluster Connectivity](#aws-spoke-cluster-connectivity-critical) *(Critical for EKS clusters)*
5. [DNS and Hostname Configuration](#dns-and-hostname-configuration)
6. [TLS/Certificate Configuration](#tlscertificate-configuration)
7. [Cloudflare Tunnel Setup](#cloudflare-tunnel-setup)
8. [Authentication (Keycloak)](#authentication-keycloak)
9. [Database Configuration](#database-configuration)
10. [AWS Credentials Architecture](#aws-credentials-architecture)
11. [Key Environment Variables](#key-environment-variables)
12. [Important Files Reference](#important-files-reference)
13. [Deployment Commands](#deployment-commands)
14. [Verification Checklist](#verification-checklist)
15. [Common Issues and Solutions](#common-issues-and-solutions)

---

## Architecture Overview

```
                                    INTERNET
                                        |
                        +---------------+---------------+
                        |                               |
                  Cloudflare CDN                  AWS NLB (gRPC)
                        |                               |
            +-----------+-----------+                   |
            |                       |                   |
    keycloak.aegis-platform.tech   remote.aegis-platform.tech
            |                       |                   |
            +-------+-------+-------+                   |
                    |                                   |
            +-------v-------+                           |
            |  cloudflared  |                           |
            |  (aegis-system)|                          |
            +-------+-------+                           |
                    |                                   |
    +---------------+---------------+                   |
    |               |               |                   |
+---v---+     +-----v-----+   +-----v-----+             |
|Keycloak|    |Platform-API|   |   Proxy   |<-----------+
|(keycloak)|  |(aegis-system)| |(aegis-system)|
+---+---+     +-----+-----+   +-----+-----+
    |               |               |
    |         +-----v-----+         |
    |         | PostgreSQL|         |
    |         |(aegis-system)|       |
    |         +-----------+         |
    |                               |
+---v-------------------------------v---+
|        Local Kubernetes Cluster       |
|         (Docker Desktop)              |
+---------------------------------------+
                    ^
                    |
            +-------+-------+
            |   Backstage   |
            | (localhost:3000)|
            +---------------+
```

### Key Components

| Component | Namespace | Purpose |
|-----------|-----------|---------|
| Platform API | aegis-system | gRPC/HTTP API for infrastructure management |
| Proxy | aegis-system | WebSocket/HTTP proxy for workspaces |
| Keycloak | keycloak | OIDC authentication server |
| cloudflared | aegis-system | Cloudflare tunnel for external access |
| PostgreSQL (platform-api) | aegis-system | Platform API database |
| PostgreSQL (keycloak) | keycloak | Keycloak database |
| ingress-nginx | ingress-nginx | Kubernetes ingress controller |

---

## Kubernetes Namespaces and Services

### Namespaces

```bash
# Core namespaces
aegis-system     # Platform API, Proxy, cloudflared, platform-api PostgreSQL
keycloak         # Keycloak and its PostgreSQL
ingress-nginx    # Ingress controller
aegis-workloads  # User workloads (created by platform-api)
```

### Key Services

| Service | Namespace | Cluster DNS | Port(s) |
|---------|-----------|-------------|---------|
| aegis-services-platform-api | aegis-system | aegis-services-platform-api.aegis-system.svc.cluster.local | 8080 (HTTP), 8081 (gRPC) |
| aegis-services-proxy | aegis-system | aegis-services-proxy.aegis-system.svc.cluster.local | 8080, 8443 |
| aegis-services-keycloak-service | keycloak | aegis-services-keycloak-service.keycloak.svc.cluster.local | 8443 (HTTPS) |
| aegis-services-keycloak (ExternalName) | keycloak | aegis-services-keycloak.keycloak.svc.cluster.local | Points to -service |
| platform-api-postgres | aegis-system | platform-api-postgres.aegis-system.svc.cluster.local | 5432 |
| keycloak-postgres | keycloak | keycloak-postgres.keycloak.svc.cluster.local | 5432 |

**IMPORTANT:** The Keycloak TLS certificate has SANs for `aegis-services-keycloak.keycloak.svc.cluster.local` (WITHOUT `-service` suffix). Use this short name for internal TLS connections (e.g., OIDC_JWKS_URL).

---

## Network Endpoints and Ports

### Local Development Endpoints

| Service | Local URL | Protocol | Notes |
|---------|-----------|----------|-------|
| Backstage UI | http://localhost:3000 | HTTP | Development server |
| Backstage Backend | http://localhost:7008 | HTTP | API backend |
| Platform API (HTTP) | https://platform-api.localtest.me | HTTPS | Via ingress |
| Platform API (gRPC) | https://platform-api-grpc.localtest.me | gRPCS | Via ingress |
| Keycloak | https://keycloak.localtest.me | HTTPS | Via ingress |
| Proxy | https://proxy.localtest.me | HTTPS | Via ingress |

### External Endpoints (via Cloudflare Tunnel)

| Hostname | Backend Service | Protocol | Purpose |
|----------|-----------------|----------|---------|
| keycloak.aegis-platform.tech | keycloak:8443 | HTTPS | Public Keycloak access |
| remote.aegis-platform.tech | platform-api:8081 | h2c/gRPC | Remote platform API access |

### AWS Resources (Production/Remote)

| Resource | Endpoint | Purpose |
|----------|----------|---------|
| NLB (gRPC) | aegis-dev-relay-nlb-353e7f4c0ac5c4be.elb.us-east-1.amazonaws.com:8081 | gRPC streaming for remote workspaces |

---

## AWS Spoke Cluster Connectivity (CRITICAL)

This section documents the complete connectivity architecture for AWS EKS spoke clusters to communicate with the local hub platform-api. This is one of the most complex aspects of the local development setup.

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           AWS VPC (us-east-1)                                    │
│                                                                                  │
│  ┌────────────────────┐        ┌─────────────────────────────────────────────┐  │
│  │  EKS Spoke Cluster │        │  aegis-dev-relay EC2 Instance               │  │
│  │                    │        │  (i-08cde33f38cb9a911)                       │  │
│  │  ┌──────────────┐  │        │  Public IP: 67.202.30.169                    │  │
│  │  │aegis-spoke   │  │        │  Private IP: 172.31.2.242                    │  │
│  │  │k8s-agent     │──┼───────►│                                              │  │
│  │  │              │  │  gRPC  │  Listens on:                                 │  │
│  │  └──────────────┘  │        │    0.0.0.0:8081 (forwarded via SSH tunnel)   │  │
│  │                    │        │    0.0.0.0:8443 (forwarded via SSH tunnel)   │  │
│  └────────────────────┘        └──────────────────┬──────────────────────────┘  │
│           │                                       │                              │
│           │                                       │ NLB Target                   │
│           │                                       │                              │
│  ┌────────▼────────────────────────────────────────────────────────────────────┐│
│  │  NLB: aegis-dev-relay-nlb-353e7f4c0ac5c4be.elb.us-east-1.amazonaws.com      ││
│  │  Type: Internal (VPC only)                                                   ││
│  │  Port 8081 → Target Group → aegis-dev-relay:8081                            ││
│  │  Port 8443 → Target Group → aegis-dev-relay:8443 (Keycloak)                 ││
│  └──────────────────────────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       │ SSH Reverse Tunnel
                                       │ (Ports 8081, 8443)
                                       ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                        Local Development Machine                                  │
│                                                                                   │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  SSH Tunnel Process (must be running)                                        │ │
│  │  ssh -R 0.0.0.0:8081:localhost:8081 -R 0.0.0.0:8443:localhost:8443 \        │ │
│  │      -i ~/.ssh/aegis-relay ec2-user@67.202.30.169                           │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
│                    │                              │                               │
│                    ▼                              ▼                               │
│             localhost:8081                  localhost:8443                        │
│                    │                                                              │
│                    │ Port-forward (must be running)                               │
│                    ▼                                                              │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  kubectl port-forward svc/aegis-services-platform-api 8081:8081             │ │
│  │  (connects localhost:8081 → kind cluster platform-api)                       │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
│                    │                                                              │
│                    ▼                                                              │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  Kind Cluster (Docker Desktop)                                               │ │
│  │                                                                               │ │
│  │  ┌─────────────────────┐                                                     │ │
│  │  │ platform-api        │◄── Receives gRPC calls from spoke agent             │ │
│  │  │ (aegis-system)      │    - Register cluster                               │ │
│  │  │ Port: 8081          │    - Heartbeats                                     │ │
│  │  └─────────────────────┘    - Workload management                            │ │
│  │                                                                               │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────────────┘
```

### Critical Components

| Component | Location | Purpose | Must Be Running |
|-----------|----------|---------|-----------------|
| SSH Tunnel | Local machine | Forwards relay:8081/8443 → localhost | Yes |
| Port-forward | Local machine | Forwards localhost:8081 → kind platform-api | Yes |
| NLB | AWS VPC | Routes spoke traffic to relay instance | Yes (always on) |
| aegis-dev-relay | EC2 | Receives NLB traffic, forwards via SSH | Yes (always on) |
| aegis-spoke-k8s-agent | EKS | Registers cluster, sends heartbeats | Yes (auto-deployed) |

### Configuration Reference

#### Platform-API Environment Variable

```yaml
# Must match the ACTUAL NLB DNS name
AEGIS_PLATFORM_API_ENDPOINT: aegis-dev-relay-nlb-353e7f4c0ac5c4be.elb.us-east-1.amazonaws.com:8081
```

**CRITICAL:** This value is passed to spoke clusters during Pulumi provisioning. If it's wrong, spoke agents can't connect.

#### Spoke Agent Helm Values

The spoke agent receives its endpoint configuration from the platform-api during provisioning:

```yaml
# charts/aegis-spoke/values-aws-relay.yaml
k8sAgent:
  env:
    AEGIS_CP_GRPC: "aegis-dev-relay-nlb-353e7f4c0ac5c4be.elb.us-east-1.amazonaws.com:8081"
    AEGIS_CP_GRPC_INSECURE: "false"
    AEGIS_CP_GRPC_SKIP_VERIFY: "true"
```

**Note:** The inline values from `AEGIS_PLATFORM_API_ENDPOINT` override the values file during Pulumi deployment.

### Startup Checklist

Before working with spoke clusters, ensure ALL of these are running:

```bash
# 1. Check SSH tunnel is running
ps aux | grep "ssh.*aegis-relay" | grep -v grep
# Expected: ssh -R 0.0.0.0:8081:localhost:8081 -R 0.0.0.0:8443:localhost:8443 ...

# 2. Start SSH tunnel if not running
ssh -i ~/.ssh/aegis-relay \
    -R 0.0.0.0:8081:localhost:8081 \
    -R 0.0.0.0:8443:localhost:8443 \
    -N -o ServerAliveInterval=30 -o ServerAliveCountMax=3 \
    -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
    ec2-user@67.202.30.169 &

# 3. Check port-forward is running
lsof -i :8081
# Expected: kubectl process listening

# 4. Start port-forward if not running
kubectl port-forward -n aegis-system svc/aegis-services-platform-api 8081:8081 --address 0.0.0.0 &

# 5. Verify spoke can reach platform-api
kubectl exec -n aegis-system deploy/aegis-services-platform-api -- \
  kubectl --kubeconfig /tmp/kubeconfigs/<cluster-id>.kubeconfig \
  -n aegis-system logs deploy/aegis-spoke-k8s-agent --tail=5
# Expected: "cluster registered" or "heartbeat" messages, NOT connection errors
```

### How Cluster Registration Works

1. **Pulumi Provisioning**
   - Creates EKS cluster
   - Installs aegis-spoke helm chart with correct `AEGIS_CP_GRPC` endpoint
   - Stores kubeconfig in `aegis-kubeconfigs` secret

2. **Spoke Agent Startup**
   - Reads `AEGIS_CP_GRPC` endpoint from environment
   - Obtains OIDC token from Keycloak
   - Calls `Register()` gRPC method on platform-api

3. **Database Registration**
   - Platform-api creates row in `clusters` table
   - Sets `last_heartbeat` timestamp
   - Stores available flavors

4. **AegisCluster CRD Sync**
   - `AegisClusterStatusSync` controller (runs every 30s)
   - Reads clusters from database
   - Creates/updates AegisCluster CRDs
   - Phase transitions: Pending → HeartbeatOK

5. **UI Display**
   - `ListClusters` API reads from database + cluster cache
   - Returns clusters to UI

### Workload Garbage Collection (Orphan Workspace Cleanup)

When a workload is deleted from the hub database, the corresponding `Workspace` CRD on the spoke cluster can remain. The spoke `aegis-spoke-k8s-agent` runs a periodic garbage collection loop to detect and delete these orphaned Workspaces (preventing stray Jobs from triggering GPU autoscaling).

**Flow (runs every 30 seconds by default):**
1. Spoke calls `ListClusterWorkloadIDs` on platform-api
2. Spoke lists local `Workspace` CRDs and reads the `aegis.workload/id` label
3. Spoke deletes Workspaces whose workload ID is missing from the hub response

Deleting the `Workspace` CRD should cascade to `AegisWorkload` and the underlying Kubernetes `Job`.

#### GC Configuration (Spoke Agent)

Set these on the `aegis-spoke-k8s-agent` Deployment:

| Variable | Default | Purpose |
|----------|---------|---------|
| `AEGIS_WORKLOAD_GC_ENABLED` | `true` | Enable/disable the workload GC loop |
| `AEGIS_WORKLOAD_GC_INTERVAL` | `30s` | How often to sync with hub (`time.ParseDuration`) |
| `AEGIS_WORKLOAD_GC_MAX_DELETIONS` | `25` | Safety limit: max Workspaces deleted per run |
| `AEGIS_WORKLOAD_GC_DRY_RUN` | `false` | Log would-delete actions without deleting |

#### Troubleshooting Orphaned Workspaces

```bash
# List Workspaces on the spoke (look for ones with aegis.workload/id labels)
kubectl --kubeconfig=/tmp/eks-kubeconfig.yaml get workspace -A -L aegis.workload/id

# Check GC activity in the spoke agent logs
kubectl --kubeconfig=/tmp/eks-kubeconfig.yaml logs -n aegis-system deploy/aegis-spoke-k8s-agent --tail=200 | grep -i "workload gc\\|orphan"

# (Optional) Confirm the hub view of workload IDs for a cluster via the HTTP gateway
# Requires platform-api HTTP port-forward (8080) to be available.
curl -sS http://localhost:8080/api/v1/clusters/<cluster-id>/workloads/ids
```

### Workspace Session Management

Workspaces use a session-style lifecycle so that policy timeouts suspend execution instead of marking the workload as failed.

**State diagram**

```
RUNNING  --(idle timeout)-->  SUSPENDED  --(resume)-->  RUNNING
RUNNING  --(terminate)-->     TERMINATED
SUSPENDED --(terminate)-->    TERMINATED
```

**Compliance mapping**

- NIST 800-171 R3 3.1.10 / FedRAMP HIGH AC-11: session lock after inactivity → `SUSPENDED`
- NIST 800-171 R3 3.1.11 / FedRAMP HIGH AC-12: session termination → `TERMINATED`

**Implementation notes**

- The spoke agent enforces `maxDurationSeconds` by setting the underlying Kubernetes `Job.spec.suspend=true` and annotating the Job with `aegis.yourorg.dev/suspend-reason=idle_timeout`.
- Resuming clears the suspend annotation, sets `aegis.yourorg.dev/resumed-at=<RFC3339Nano>`, and sets `Job.spec.suspend=false`.
- Terminating deletes the `Workspace`/`AegisWorkload` resources on the spoke and marks the hub workload `TERMINATED`.

**Configuration**

- `Workspace.spec.maxDurationSeconds` (or queue default) controls when an active session is suspended.
- Database migration `0007_add_workload_session_management` adds `workloads.suspended_at`, `suspend_reason`, `resume_count`, `terminated_at`, and `terminate_reason`.

**API**

- `POST /api/v1/workloads/{id}/resume` (also available at `/v1/workloads/{id}/resume`)
- `POST /api/v1/workloads/{id}/terminate` with body `{"reason":"user_requested"}` (also available at `/v1/workloads/{id}/terminate`)

### Troubleshooting Decision Tree

```
Cluster not showing in UI?
│
├─► Check AegisCluster CRD exists
│   kubectl get aegisclusters -A
│   │
│   ├─► No CRDs → Check database
│   │   kubectl exec -n aegis-system platform-postgres-0 -- \
│   │     psql -U aegis_platform -d aegis_platform -c "SELECT id FROM clusters;"
│   │   │
│   │   ├─► No rows → Spoke agent not registering
│   │   │   └─► Go to "Spoke Agent Not Registering" section
│   │   │
│   │   └─► Has rows → Status sync controller issue
│   │       └─► Check platform-api logs for "clusterstatus" errors
│   │
│   └─► CRD exists but phase is wrong → Check heartbeats
│       kubectl get aegiscluster <name> -o yaml
│       └─► Check lastHeartbeat timestamp
│
└─► CRD exists with HeartbeatOK → API/UI issue
    └─► Check platform-api ListClusters endpoint
```

### Spoke Agent Not Registering

#### Symptom: DNS Resolution Failure

```
cluster registration failed: dial tcp: lookup aegis-dev-relay-nlb-XXX...amazonaws.com: no such host
```

**Cause:** Wrong NLB DNS name configured

**Fix:**
```bash
# 1. Find the ACTUAL NLB DNS name
aws elbv2 describe-load-balancers --names aegis-dev-relay-nlb --region us-east-1 \
  --query 'LoadBalancers[0].DNSName' --output text

# 2. Update platform-api with correct endpoint
kubectl set env deployment/aegis-services-platform-api -n aegis-system \
  AEGIS_PLATFORM_API_ENDPOINT=<CORRECT_NLB_DNS>:8081

# 3. Wait for Pulumi to update the spoke (happens automatically)
kubectl logs -n aegis-system deploy/aegis-services-platform-api --tail=100 | grep "release updating"
```

#### Symptom: Connection Reset / TLS Handshake Failure

```
cluster registration failed: transport: authentication handshake failed: read tcp ... connection reset by peer
```

**Cause:** SSH tunnel or port-forward not running

**Fix:**
```bash
# Check what's listening on 8081
lsof -i :8081
# If empty, nothing is listening!

# Start port-forward
kubectl port-forward -n aegis-system svc/aegis-services-platform-api 8081:8081 --address 0.0.0.0 &

# Verify SSH tunnel
ps aux | grep "ssh.*aegis-relay"
# If not running, start it (see Startup Checklist)
```

#### Symptom: Foreign Key Constraint Violation

```
postgres action failed: set_cluster_project: violates foreign key constraint "cluster_labels_cluster_id_fkey"
```

**Cause:** Cluster row doesn't exist because spoke agent never registered (cascading from connectivity issue)

**Fix:** Resolve the underlying connectivity issue first, then the FK constraint issue resolves automatically when the spoke registers.

### AWS NLB Reference

| Property | Value |
|----------|-------|
| Name | aegis-dev-relay-nlb |
| DNS | aegis-dev-relay-nlb-353e7f4c0ac5c4be.elb.us-east-1.amazonaws.com |
| Type | Network Load Balancer (internal) |
| VPC | vpc-025484f1a3008d8eb |
| Listeners | 8081 (gRPC), 8443 (Keycloak) |

#### Target Groups

| Target Group | Port | Target | Health Check |
|--------------|------|--------|--------------|
| aegis-dev-relay-grpc | 8081 | i-08cde33f38cb9a911 (aegis-dev-relay) | TCP 8081 |
| aegis-dev-relay-kc | 8443 | i-08cde33f38cb9a911 (aegis-dev-relay) | TCP 8443 |

#### Check NLB Health

```bash
# Check target health
aws elbv2 describe-target-health \
  --target-group-arn "arn:aws:elasticloadbalancing:us-east-1:567751785679:targetgroup/aegis-dev-relay-grpc/a0dad6f9b7dc5670" \
  --region us-east-1

# Expected: "State": "healthy"
```

### Connecting to Spoke Cluster

The platform-api stores kubeconfigs for spoke clusters:

```bash
# List available kubeconfigs
kubectl get secret -n aegis-system aegis-kubeconfigs -o jsonpath='{.data}' | jq -r 'keys[]'

# Execute commands on spoke cluster via platform-api
kubectl exec -n aegis-system deploy/aegis-services-platform-api -- \
  kubectl --kubeconfig /tmp/kubeconfigs/<cluster-id>.kubeconfig \
  get pods -A

# Example: Check spoke agent logs
kubectl exec -n aegis-system deploy/aegis-services-platform-api -- \
  kubectl --kubeconfig /tmp/kubeconfigs/end-us-east-1-atlas-train-govcloud.kubeconfig \
  -n aegis-system logs deploy/aegis-spoke-k8s-agent --tail=50
```

### Automation Tools (USE THESE FIRST!)

**Before debugging manually, use these Makefile targets:**

| Command | What It Does |
|---------|--------------|
| `make spoke-start` | **Primary setup** - Starts SSH tunnel + port-forwards in foreground (Ctrl+C to stop) |
| `make spoke-check` | Quick status check - Shows what's running/broken |
| `make spoke-stop` | Stop all spoke connectivity services |
| `make spoke-fix-nlb` | Auto-fix NLB endpoint if misconfigured |
| `make spoke-connectivity` | Ensure everything running (quick start if needed) |

#### Typical Workflow

```bash
# 1. Starting fresh - run in a dedicated terminal (stays running):
make spoke-start

# 2. In another terminal, verify everything works:
make spoke-check

# 3. If cluster not showing in UI, try:
make spoke-fix-nlb    # Fixes NLB endpoint if wrong
make spoke-check      # Verify again

# 4. When done for the day:
make spoke-stop
```

#### Example Output from `make spoke-check`

```
==========================================
  Aegis Spoke Connectivity Status
==========================================

SSH Tunnel:       RUNNING ✓
Port-forward:     RUNNING ✓
NLB Endpoint:     CONFIGURED CORRECTLY ✓
NLB Health:       HEALTHY ✓
Clusters in DB:   1 registered ✓
AegisCluster CRDs: 1 found ✓

All connectivity requirements are met!
```

If any item shows **NOT RUNNING** or **MISCONFIGURED**, run `make spoke-start` to fix it.

### Scripts Reference

| Script | Purpose |
|--------|---------|
| `scripts/start-aws-tunnel.sh` | Full tunnel setup with auto-restart (called by `make spoke-start`) |
| `scripts/stop-aws-tunnel.sh` | Stop the tunnel (called by `make spoke-stop`) |
| `scripts/spoke-connectivity.sh` | Status checks and fixes (called by `make spoke-check`, `make spoke-fix-nlb`) |

### Manual Commands (if automation fails)

Only use these if the Makefile targets don't work:

```bash
# Check SSH tunnel manually
ps aux | grep "ssh.*aegis-relay" | grep -v grep

# Check port-forward manually
lsof -i :8081

# Start SSH tunnel manually
ssh -i ~/.ssh/aegis-relay \
    -R 0.0.0.0:8081:localhost:8081 \
    -R 0.0.0.0:8443:localhost:8443 \
    -N -o ServerAliveInterval=30 \
    ec2-user@67.202.30.169 &

# Start port-forward manually
kubectl port-forward -n aegis-system svc/aegis-services-platform-api 8081:8081 &

# Check NLB health manually
aws elbv2 describe-target-health \
  --target-group-arn "arn:aws:elasticloadbalancing:us-east-1:567751785679:targetgroup/aegis-dev-relay-grpc/a0dad6f9b7dc5670" \
  --region us-east-1

# Check clusters in database
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U aegis_platform -d aegis_platform -c "SELECT id, last_heartbeat FROM clusters;"

# Check AegisCluster CRDs
kubectl get aegisclusters -A
```

---

## DNS and Hostname Configuration

### /etc/hosts Entries Required

```bash
# Aegis Local Development
127.0.0.1 platform-api.localtest.me platform-api-grpc.localtest.me proxy.localtest.me
127.0.0.1 keycloak.localtest.me
```

### Why .localtest.me?

- `.localtest.me` is a public domain that always resolves to 127.0.0.1
- This allows proper TLS certificate validation without custom DNS
- Certificates can have valid SANs for these hostnames
- No need for custom DNS server or dnsmasq

### External Domains (Cloudflare DNS)

| Domain | Points To |
|--------|-----------|
| *.aegis-platform.tech | Cloudflare Tunnel |
| *.aegist.dev | AWS resources (EKS, NLB) |

---

## TLS/Certificate Configuration

### Certificate Hierarchy

```
Aegis Local Root CA (O=Aegis Dev, CN=Aegis Local Root CA)
    |
    +-- keycloak.localtest.me certificate
    |       (SANs: keycloak.localtest.me, keycloak.aegis-platform.tech,
    |              aegis-services-keycloak.keycloak.svc.cluster.local)
    |
    +-- platform-api.localtest.me certificate
    |       (SANs: platform-api.localtest.me, platform-api-grpc.localtest.me, proxy.localtest.me)
    |
    +-- proxy.localtest.me certificate
```

### TLS Secrets in Kubernetes

| Secret Name | Namespace | Contains | Used By |
|-------------|-----------|----------|---------|
| aegis-services-platform-api-tls | aegis-system | Platform API cert/key | Platform API, Ingress |
| aegis-services-proxy-tls | aegis-system | Proxy cert/key | Proxy, Ingress |
| keycloak-tls | keycloak | Keycloak cert/key | Keycloak |
| aegis-platform-api-oidc-ca | aegis-system | CA certificate | Platform API (for OIDC/JWKS validation) |

### CA Certificate Consistency (CRITICAL)

**All certificates must be signed by the SAME CA.** This is the #1 cause of recurring authentication failures.

#### Why This Matters

Platform-API validates Keycloak tokens by fetching JWKS keys from Keycloak's internal endpoint:
```
Platform-API → https://aegis-services-keycloak.keycloak.svc.cluster.local:8443/realms/aegis/.../certs
```

This internal HTTPS connection requires Platform-API to trust Keycloak's TLS certificate. If the CA that signed Keycloak's cert doesn't match the CA in `aegis-platform-api-oidc-ca`, you get:
```
x509: certificate signed by unknown authority
```

#### The CA Used

- **Subject:** `O=Aegis Dev, CN=Aegis Local Root CA`
- **Valid for:** 10 years (generated December 2025)
- **Stored in:** `charts/aegis-services/values/local-tls.yaml`

#### Common Symptoms of CA Mismatch

1. `Authentication required. Please sign in with Keycloak and retry.` in Backstage
2. Platform-API logs show: `auth_failure`, `token is unverifiable`, `certificate signed by unknown authority`
3. Works after fresh deploy, breaks after Docker Desktop restart or helm upgrade

#### Prevention: Single Source of Truth

The `local-tls.yaml` values file contains BOTH:
- `platformApi.auth.oidc.caBundle.data` - The CA certificate
- `keycloak.tls.secret.cert` - Keycloak's TLS cert (signed by the CA above)

**NEVER regenerate these separately.** If you need new certs, regenerate BOTH from the same CA.

#### Verification Script

```bash
#!/bin/bash
# verify-ca-consistency.sh - Run this after any cert changes

echo "=== Checking CA Consistency ==="

# Get OIDC CA fingerprint
OIDC_CA=$(kubectl -n aegis-system get secret aegis-platform-api-oidc-ca \
  -o jsonpath='{.data.ca\.crt}' | base64 -d | openssl x509 -noout -fingerprint -sha256 2>/dev/null)

# Get Keycloak cert issuer fingerprint (need to extract CA from chain or check issuer)
KC_ISSUER=$(kubectl -n keycloak get secret keycloak-tls \
  -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -issuer 2>/dev/null)

# Get OIDC CA subject
OIDC_SUBJECT=$(kubectl -n aegis-system get secret aegis-platform-api-oidc-ca \
  -o jsonpath='{.data.ca\.crt}' | base64 -d | openssl x509 -noout -subject 2>/dev/null)

echo "OIDC CA: $OIDC_SUBJECT"
echo "OIDC CA Fingerprint: $OIDC_CA"
echo "Keycloak Cert Issuer: $KC_ISSUER"

# Check if they match
if [[ "$OIDC_SUBJECT" == *"Aegis Dev"* ]] && [[ "$KC_ISSUER" == *"Aegis Dev"* ]]; then
  echo "✅ CA consistency OK - both use 'O=Aegis Dev, CN=Aegis Local Root CA'"
else
  echo "❌ CA MISMATCH DETECTED - run regeneration script"
  exit 1
fi

# Test actual connectivity
echo ""
echo "=== Testing JWKS Fetch ==="
kubectl -n aegis-system exec deployment/aegis-services-platform-api -- \
  curl -s --cacert /etc/aegis-platform-api/oidc/ca.crt \
  https://aegis-services-keycloak.keycloak.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/certs \
  | head -c 100
echo "..."
echo "✅ JWKS fetch successful"
```

#### Regeneration Script (When Certs Are Broken)

```bash
#!/bin/bash
# regenerate-matching-certs.sh - Regenerates CA + Keycloak cert that match

set -e
cd /tmp

echo "=== Generating New CA ==="
openssl genrsa -out aegis-ca.key 4096
openssl req -x509 -new -nodes -key aegis-ca.key -sha256 -days 3650 \
  -subj "/O=Aegis Dev/CN=Aegis Local Root CA" -out aegis-ca.crt

echo "=== Generating Keycloak TLS Cert ==="
openssl genrsa -out keycloak.key 2048
openssl req -new -key keycloak.key -subj "/CN=keycloak.localtest.me" -out keycloak.csr

cat > keycloak-san.cnf << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
[req_distinguished_name]
[v3_req]
subjectAltName = @alt_names
[alt_names]
DNS.1 = keycloak.localtest.me
DNS.2 = aegis-services-keycloak.keycloak.svc.cluster.local
DNS.3 = aegis-services-keycloak
DNS.4 = keycloak.aegis-platform.tech
EOF

openssl x509 -req -in keycloak.csr -CA aegis-ca.crt -CAkey aegis-ca.key \
  -CAcreateserial -out keycloak.crt -days 1095 -sha256 \
  -extfile keycloak-san.cnf -extensions v3_req

echo "=== Verifying Chain ==="
openssl verify -CAfile aegis-ca.crt keycloak.crt

echo "=== Updating Kubernetes Secrets ==="
kubectl -n keycloak create secret tls keycloak-tls \
  --cert=keycloak.crt --key=keycloak.key --dry-run=client -o yaml | kubectl apply -f -

kubectl -n aegis-system create secret generic aegis-platform-api-oidc-ca \
  --from-file=ca.crt=aegis-ca.crt --dry-run=client -o yaml | kubectl apply -f -

echo "=== Restarting Services ==="
kubectl rollout restart statefulset/aegis-services-keycloak -n keycloak
kubectl rollout restart deployment/aegis-services-platform-api -n aegis-system

echo "=== Waiting for Rollout ==="
kubectl rollout status statefulset/aegis-services-keycloak -n keycloak --timeout=120s
kubectl rollout status deployment/aegis-services-platform-api -n aegis-system --timeout=120s

echo ""
echo "✅ Certs regenerated and applied!"
echo ""
echo "⚠️  IMPORTANT: Update local-tls.yaml with the new certs to make this permanent:"
echo "   - Copy /tmp/aegis-ca.crt content to platformApi.auth.oidc.caBundle.data"
echo "   - Copy /tmp/keycloak.crt content to keycloak.tls.secret.cert"
echo "   - Copy /tmp/keycloak.key content to keycloak.tls.secret.key"
```

#### Quick Fix (If You Just Need It Working Now)

```bash
# 1. Verify the mismatch
kubectl -n aegis-system exec deployment/aegis-services-platform-api -- \
  cat /etc/aegis-platform-api/oidc/ca.crt | openssl x509 -noout -subject
kubectl -n keycloak get secret keycloak-tls -o jsonpath='{.data.tls\.crt}' | \
  base64 -d | openssl x509 -noout -issuer

# 2. If they don't match, run the regeneration script above

# 3. Or as a temporary workaround (NOT RECOMMENDED for production):
# Set OIDC_SKIP_TLS_VERIFY=true in platform-api environment
```

### Who Handles TLS?

| Access Path | TLS Termination |
|-------------|-----------------|
| localhost:3000 (Backstage) | None (HTTP dev server) |
| *.localtest.me (local) | ingress-nginx terminates, re-encrypts to backends |
| *.aegis-platform.tech (external) | Cloudflare terminates, cloudflared connects with noTLSVerify |

### Local CA Trust

For local development, add the CA to your system trust store:

```bash
# Location of local CA (generated during setup)
~/aegis-local-trust.pem

# Run Backstage with CA trust
NODE_EXTRA_CA_CERTS=~/aegis-local-trust.pem yarn dev
```

### Client-Side TLS CA Configuration (VS Code Extension)

**CRITICAL:** When configuring TLS for Node.js HTTP and gRPC clients, both **undici** and **@grpc/grpc-js** libraries **REPLACE** system root CAs entirely when you provide a custom CA.

#### The Problem

If you only pass your custom CA to these libraries:
- Connections to services using your custom CA will work (e.g., `platform-api-grpc.localtest.me`)
- Connections to services using public CAs (Cloudflare, Let's Encrypt, etc.) will **FAIL** with `unable to get local issuer certificate`

Example failure scenario:
1. Extension authenticates via Keycloak at `keycloak.aegis-platform.tech` (Cloudflare certificate)
2. Only custom CA is loaded → Cloudflare's certificate chain cannot be verified
3. Token exchange fails: `fetch failed (unable to get local issuer certificate)`

#### The Solution: Shared TLS Utility

The VS Code extension includes `src/tls.ts` which **always** combines the environment's CA with system CAs.

**Extension Location:** `aegis-vscode-remote/extension/src/tls.ts`

```typescript
import * as tls from 'tls';

// For undici (returns string array)
export function getCombinedCAsArray(customCAPem: string | Buffer): string[] {
  const systemCAs = tls.rootCertificates;
  const customCerts = extractPemCertificates(customCAPem.toString());
  return [...systemCAs, ...customCerts];
}

// For @grpc/grpc-js (returns Buffer)
export function getCombinedCAsBuffer(customCAPem: string | Buffer): Buffer {
  const allCAs = getCombinedCAsArray(customCAPem);
  return Buffer.from(allCAs.join('\n'), 'utf8');
}
```

#### Usage Examples

**HTTP Client (undici):**
```typescript
import { getCombinedCAsArray } from './tls';

// CORRECT: Use the shared utility
const combinedCAs = getCombinedCAsArray(customCA);
const agent = new Agent({ connect: { ca: combinedCAs } });

// WRONG: Never do this - it replaces system CAs!
// const agent = new Agent({ connect: { ca: customCA } });
```

**gRPC Client (@grpc/grpc-js):**
```typescript
import { loadCombinedCAsBuffer } from './tls';

// CORRECT: Use the shared utility
const combinedCA = await loadCombinedCAsBuffer(security.caPath);
return grpc.credentials.createSsl(combinedCA);

// WRONG: Never do this - it replaces system CAs!
// const ca = await fs.readFile(caPath);
// return grpc.credentials.createSsl(ca);
```

#### Key Points

1. **`tls.rootCertificates`** - Node.js exposes system root CAs as an array of PEM strings
2. **PEM extraction regex** - Custom CA files may contain multiple certificates; extract each one
3. **Join with newlines** - For gRPC, combine all PEM strings into a single buffer
4. **Array format for undici** - undici accepts an array of PEM strings directly

#### Code Review Checklist for TLS

When reviewing PRs that touch TLS/certificate configuration, verify:

- [ ] **No direct CA passing** - Code does NOT pass custom CA directly to `grpc.credentials.createSsl()`, `undici Agent`, or `https.Agent`
- [ ] **Uses shared utility** - All CA configuration uses functions from `src/tls.ts`
- [ ] **Error handling** - CA file read errors are logged, not silently ignored

**Red flags to look for:**
```typescript
// BAD - These patterns will break connections to public CA services
grpc.credentials.createSsl(await fs.readFile(caPath))
new Agent({ connect: { ca: customCaBuffer } })
https.request({ ca: customCa })

// GOOD - Always use the shared utilities
grpc.credentials.createSsl(await loadCombinedCAsBuffer(caPath))
new Agent({ connect: { ca: getCombinedCAsArray(customCa) } })
```

#### Debugging TLS Issues

1. **Check channel state** - gRPC channels stuck in CONNECTING (state 1) usually indicate TLS handshake failure:
   ```
   IDLE = 0, CONNECTING = 1, READY = 2, TRANSIENT_FAILURE = 3, SHUTDOWN = 4
   ```

2. **Test with openssl**:
   ```bash
   openssl s_client -connect platform-api-grpc.localtest.me:443 \
     -CAfile ~/aegis-local-trust.pem \
     -servername platform-api-grpc.localtest.me
   ```

3. **Check extension logs** - Look for these patterns in Aegis Remote output:
   ```
   [tls] combining 8 custom CA cert(s) with 146 system CAs
   [platform] gRPC channel state changed: 1 -> 2  # Success!
   [platform] gRPC channel state changed: 0 -> 1  # Stuck here = TLS issue
   ```

#### Common TLS Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `unable to get local issuer certificate` | Custom CA replaces system CAs | Use shared tls.ts utilities |
| `DEADLINE_EXCEEDED: Waiting for LB pick` | gRPC channel stuck in CONNECTING | Fix TLS CA configuration |
| `self signed certificate in certificate chain` | Custom CA not loaded | Check `caPath` setting |

---

## Cloudflare Tunnel Setup

### Tunnel Information

| Property | Value |
|----------|-------|
| Tunnel ID | a21f6af6-e7b8-49d5-90be-08f0a8d9a829 |
| Account Tag | aff4ed64e72c1e215be1a4ba105e4ab1 |
| Config File | `/cloudflared-config.yaml` (repo root) |

### Tunnel Ingress Rules

```yaml
ingress:
  # Platform API (gRPC via h2c)
  - hostname: remote.aegis-platform.tech
    service: https://aegis-services-platform-api.aegis-system.svc.cluster.local:8081
    originRequest:
      noTLSVerify: true
      http2Origin: true
      disableChunkedEncoding: true

  # Keycloak (HTTPS)
  - hostname: keycloak.aegis-platform.tech
    service: https://aegis-services-keycloak-service.keycloak.svc.cluster.local:8443
    originRequest:
      noTLSVerify: true
      httpHostHeader: keycloak.aegis-platform.tech
      originServerName: keycloak.aegis-platform.tech

  # Catch-all 404
  - service: http_status:404
```

### Deploying cloudflared

```bash
# Manual deployment (if setup script fails)
kubectl apply -f cloudflared-config.yaml

# Verify tunnel is running
kubectl -n aegis-system get pods -l app=cloudflared
kubectl -n aegis-system logs -l app=cloudflared --tail=20

# Expected log output: "Connection ... registered" (4 connections)
```

### Why cloudflared Instead of Public Ingress?

- Local Docker Desktop cluster has no public IP
- Cloudflare tunnel creates outbound connection to Cloudflare edge
- Cloudflare handles TLS termination and DDoS protection
- No need to expose ports or configure port forwarding

---

## Authentication (Keycloak)

### OIDC Configuration

| Setting | Value |
|---------|-------|
| Realm | aegis |
| Issuer URL (local) | https://keycloak.localtest.me/realms/aegis |
| Issuer URL (external) | https://keycloak.aegis-platform.tech/realms/aegis |
| Client ID | backstage |
| Client Secret | local-backstage-client-secret |

### Keycloak Admin Access

```bash
# URL
https://keycloak.localtest.me/admin

# Credentials (from values/local-tls.yaml)
Username: admin
Password: REDACTED_KEYCLOAK_ADMIN_PASSWORD
```

### Authentication Flow

```
1. User visits localhost:3000 (Backstage)
2. Clicks "Sign In"
3. Redirected to keycloak.localtest.me/realms/aegis/protocol/openid-connect/auth
4. User authenticates with Keycloak
5. Redirected back to localhost:7008/api/auth/keycloak/handler/frame
6. Backstage receives tokens and creates session
7. User is authenticated
```

### Important: Issuer Consistency

The Keycloak issuer in the JWT **must match** what Backstage expects:
- If Backstage uses `keycloak.localtest.me`, Keycloak must be configured with that hostname
- Mixing `localtest.me` and `aegis-platform.tech` will cause "invalid issuer" errors

---

## Database Configuration

### Platform API PostgreSQL

| Setting | Value |
|---------|-------|
| Host | platform-api-postgres.aegis-system.svc.cluster.local |
| Port | 5432 |
| Database | platformapi |
| Username | platformapi |
| Password | platformapi-password |
| Storage | 1Gi PVC |

### Keycloak PostgreSQL

| Setting | Value |
|---------|-------|
| Host | keycloak-postgres.keycloak.svc.cluster.local |
| Port | 5432 |
| Database | keycloak |
| Username | keycloak |
| Password | REDACTED_KEYCLOAK_DB_PASSWORD |
| Storage | 5Gi PVC |

### Migrations

Platform API migrations run as a Helm post-install hook:
- Wait for PostgreSQL to be ready (init container)
- Run golang-migrate with SQL files from ConfigMap
- Migration files: `charts/aegis-services/files/platform-api/migrations/`

---

## AWS Credentials Architecture

**CRITICAL:** The platform uses a specific credential flow. Using incorrect credentials will cause Pulumi failures, EKS provisioning errors, or "InvalidAccessKeyId" errors.

### Credential Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│  aegis-pulumi-provisioner (IAM User)                                    │
│  Managed by: terraform/pulumi-stack/                                    │
│                                                                          │
│  Permissions:                                                            │
│  ├── S3 Access: s3://aegis-pulumi-state-dev/aegis/pulumi/*              │
│  │   (ListBucket, GetObject, PutObject, DeleteObject)                   │
│  │                                                                       │
│  └── sts:AssumeRole → aegis-platform role                               │
└────────────────────────────┬────────────────────────────────────────────┘
                             │
                             │ AssumeRole (automatic via Pulumi/SDK)
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  aegis-platform (IAM Role)                                              │
│  ARN: arn:aws:iam::567751785679:role/aegis-platform                     │
│  Policy: AdministratorAccess                                            │
│                                                                          │
│  Used for:                                                               │
│  ├── EKS cluster provisioning (via Pulumi)                              │
│  ├── GPU/EC2 workload provisioning                                      │
│  ├── Kubernetes API authentication (via aegis-eks-token)                │
│  └── CloudWatch/Observability APIs                                      │
└─────────────────────────────────────────────────────────────────────────┘
```

### The Single Source of Truth

**Always use terraform-managed credentials.** Never manually create or rotate AWS keys.

```bash
# Get the correct credentials (run from repo root)
cd terraform/pulumi-stack
terraform output pulumi_access_key_id
terraform output pulumi_secret_access_key
```

### Where Credentials Are Configured

| Location | Purpose | How to Update |
|----------|---------|---------------|
| `charts/aegis-services/values/local-tls.yaml` | Helm values (secrets section) | Edit file, redeploy |
| `aegis-services-platform-api-secret` (K8s) | Runtime secret in cluster | Patch secret, restart pod |
| `terraform/pulumi-stack/terraform.tfstate` | Terraform state (source of truth) | `terraform output` |

### Current Credentials Reference

| Item | Value |
|------|-------|
| IAM User | `aegis-pulumi-provisioner` |
| User ARN | `arn:aws:iam::567751785679:user/service/aegis-pulumi-provisioner` |
| Access Key ID | `REDACTED_AWS_ACCESS_KEY_ID` |
| Target Role ARN | `arn:aws:iam::567751785679:role/aegis-platform` |
| S3 Bucket | `aegis-pulumi-state-dev` |
| S3 Prefix | `aegis/pulumi` |

### How Different Operations Use Credentials

| Operation | Credential Path |
|-----------|-----------------|
| Pulumi S3 state access | Direct IAM user credentials (AWS_ACCESS_KEY_ID/SECRET) |
| EKS cluster provisioning | IAM user → AssumeRole → aegis-platform role |
| kubectl/EKS API auth | IAM user → AssumeRole → aegis-eks-token → EKS token |
| CloudWatch/Observability | IAM user → AssumeRole → aegis-platform role |

### Updating Credentials After Rotation

If credentials are rotated in AWS or terraform:

```bash
# 1. Get new credentials from terraform
cd terraform/pulumi-stack
NEW_KEY_ID=$(terraform output -raw pulumi_access_key_id)
NEW_SECRET=$(terraform output -raw pulumi_secret_access_key)

# 2. Update Helm values file
# Edit charts/aegis-services/values/local-tls.yaml:
#   secrets:
#     aws-access-key-id: "<NEW_KEY_ID>"
#     aws-secret-access-key: "<NEW_SECRET>"

# 3. Update Kubernetes secret directly (for immediate fix)
NEW_ACCESS_KEY_B64=$(echo -n "$NEW_KEY_ID" | base64)
NEW_SECRET_KEY_B64=$(echo -n "$NEW_SECRET" | base64)
kubectl -n aegis-system patch secret aegis-services-platform-api-secret \
  -p '{"data":{"aws-access-key-id":"'"$NEW_ACCESS_KEY_B64"'","aws-secret-access-key":"'"$NEW_SECRET_KEY_B64"'"}}'

# 4. Restart platform-api to pick up new credentials
kubectl -n aegis-system rollout restart deployment/aegis-services-platform-api

# 5. Verify
kubectl -n aegis-system exec deployment/aegis-services-platform-api -- \
  aws sts get-caller-identity
# Should show: arn:aws:iam::567751785679:user/service/aegis-pulumi-provisioner
```

### Verification Commands

```bash
# Check current credentials in pod
kubectl -n aegis-system exec deployment/aegis-services-platform-api -- \
  env | grep AWS_ACCESS_KEY

# Test S3 access (Pulumi state)
kubectl -n aegis-system exec deployment/aegis-services-platform-api -- \
  aws s3 ls s3://aegis-pulumi-state-dev/aegis/pulumi/.pulumi/

# Test AssumeRole capability
kubectl -n aegis-system exec deployment/aegis-services-platform-api -- \
  aws sts assume-role \
    --role-arn "arn:aws:iam::567751785679:role/aegis-platform" \
    --role-session-name test \
    --query 'AssumedRoleUser.Arn' --output text
# Should show: arn:aws:sts::567751785679:assumed-role/aegis-platform/test
```

### Common Credential Issues

#### InvalidAccessKeyId Error

**Symptom:**
```
InvalidAccessKeyId: The AWS Access Key Id you provided does not exist in our records
```

**Cause:** Wrong credentials in pod (not the terraform-managed ones)

**Fix:** Follow "Updating Credentials After Rotation" steps above

#### Access Denied on AssumeRole

**Symptom:**
```
AccessDenied: User is not authorized to perform: sts:AssumeRole
```

**Cause:** Using credentials that don't have AssumeRole permission (e.g., root account or different IAM user)

**Fix:** Ensure you're using the `aegis-pulumi-provisioner` user credentials

#### S3 Access Denied

**Symptom:**
```
AccessDenied: Access Denied for bucket aegis-pulumi-state-dev
```

**Cause:** Wrong credentials or IAM policy issue

**Fix:** Verify using terraform-managed credentials; check IAM policy in `terraform/pulumi-stack/main.tf`

### Terraform IAM Setup Reference

The IAM user and permissions are defined in `terraform/pulumi-stack/main.tf`:

```hcl
# IAM User
resource "aws_iam_user" "pulumi" {
  name = "aegis-pulumi-provisioner"
  path = "/service/"
}

# Permissions: AssumeRole + S3
data "aws_iam_policy_document" "pulumi" {
  # AssumeRole to aegis-platform
  statement {
    effect    = "Allow"
    actions   = ["sts:AssumeRole"]
    resources = ["arn:aws:iam::567751785679:role/aegis-platform"]
  }

  # S3 state bucket access
  statement {
    effect  = "Allow"
    actions = ["s3:ListBucket", "s3:GetObject", "s3:PutObject", "s3:DeleteObject", ...]
    resources = [
      "arn:aws:s3:::aegis-pulumi-state-dev",
      "arn:aws:s3:::aegis-pulumi-state-dev/aegis/pulumi/*",
      "arn:aws:s3:::aegis-pulumi-state-dev/.pulumi/*"
    ]
  }
}
```

---

## Key Environment Variables

### Platform API

| Variable | Value | Purpose |
|----------|-------|---------|
| AEGIS_PLATFORM_API_ENDPOINT | aegis-dev-relay-nlb-...elb.us-east-1.amazonaws.com:8081 | Remote gRPC endpoint |
| AEGIS_PROXY_BASE_URL | https://proxy.localtest.me | Proxy service URL |
| PULUMI_BACKEND_URL | s3://aegis-pulumi-state-dev/aegis/pulumi | Pulumi state storage |
| AWS_REGION | us-east-1 | Default AWS region |
| OIDC_ISSUER_URL | https://keycloak.aegis-platform.tech/realms/aegis | Token validation (must match token issuer) |
| OIDC_JWKS_URL | https://aegis-services-keycloak.keycloak.svc.cluster.local:8443/... | JWKS fetch (use short service name for cert SAN match) |

### Backstage (aegis-ui)

See `aegis-ui/app-config.local-dev.yaml` and `.env.development`:
```yaml
auth:
  providers:
    keycloak:
      development:
        clientId: backstage
        clientSecret: local-backstage-client-secret
        issuer: https://keycloak.aegis-platform.tech/realms/aegis
        metadataUrl: https://keycloak.aegis-platform.tech/realms/aegis/.well-known/openid-configuration
```

---

## Important Files Reference

### Helm Values

| File | Purpose |
|------|---------|
| `charts/aegis-services/values/common.yaml` | Shared defaults |
| `charts/aegis-services/values/local.yaml` | Local dev without TLS |
| `charts/aegis-services/values/local-tls.yaml` | Local dev with TLS (primary) |
| `charts/aegis-services/values/cloud.yaml` | Cloud/production settings |

### Kubernetes Manifests

| File | Purpose |
|------|---------|
| `cloudflared-config.yaml` | Cloudflare tunnel deployment (manual) |
| `charts/aegis-services/templates/keycloak/` | Keycloak CRDs and resources |
| `charts/aegis-services/templates/platform-api-*.yaml` | Platform API resources |
| `charts/aegis-services/crds/` | Custom Resource Definitions |

### Configuration Files

| File | Purpose |
|------|---------|
| `charts/aegis-services/files/keycloak/aegis-realm.json` | Keycloak realm definition |
| `charts/aegis-services/files/platform-api/migrations/` | Database migrations |

### Scripts

| File | Purpose |
|------|---------|
| `scripts/setup-cloudflare-tunnels.sh` | Cloudflare tunnel setup (may have bugs) |
| `scripts/generate-certs.sh` | Generate local TLS certificates |
| `scripts/verify-ca-consistency.sh` | **Pre-deploy check** - Validates CA and Keycloak certs match |
| `scripts/regenerate-matching-certs.sh` | **Recovery** - Regenerates matching CA + Keycloak certs |
| `Makefile` | Primary build/deploy automation |

---

## Debugging Tools

This section documents scripts and commands available for diagnosing and fixing common infrastructure issues.

### CA Certificate Verification

**When to use:** Before deploying, or when seeing "certificate signed by unknown authority" or "Authentication required" errors.

```bash
# Verify CA consistency (runs automatically before make deploy-local-tls)
./scripts/verify-ca-consistency.sh

# Or via Makefile
make verify-ca-consistency
```

**What it does:**
- Extracts CA certificate from `local-tls.yaml`
- Extracts Keycloak TLS certificate from `local-tls.yaml`
- Verifies the Keycloak cert was signed by the CA
- Fails with clear error if mismatch detected

### CA Certificate Regeneration

**When to use:** When CA verification fails and you need to regenerate matching certificates.

```bash
./scripts/regenerate-matching-certs.sh
```

**What it does:**
1. Generates new CA certificate (4096-bit RSA, 10-year validity)
2. Generates new Keycloak TLS certificate signed by that CA
3. Updates Kubernetes secrets (`keycloak-tls`, `aegis-platform-api-oidc-ca`)
4. Restarts Keycloak and Platform-API
5. Outputs instructions for updating `local-tls.yaml` to make permanent

**Important:** After running, update `local-tls.yaml` with the new certs to prevent recurrence on next deploy.

### Quick Diagnostic Commands

```bash
# Check if Platform-API can reach Keycloak JWKS (auth validation)
kubectl -n aegis-system exec deployment/aegis-services-platform-api -- \
  curl -s --cacert /etc/aegis-platform-api/oidc/ca.crt \
  https://aegis-services-keycloak.keycloak.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/certs \
  | head -c 100

# Check for auth failures in Platform-API logs
kubectl -n aegis-system logs deployment/aegis-services-platform-api --tail=50 | grep -i "auth_failure\|certificate"

# Verify CA subject matches Keycloak cert issuer
kubectl -n aegis-system exec deployment/aegis-services-platform-api -- \
  cat /etc/aegis-platform-api/oidc/ca.crt | openssl x509 -noout -subject
kubectl -n keycloak get secret keycloak-tls -o jsonpath='{.data.tls\.crt}' | \
  base64 -d | openssl x509 -noout -issuer

# Check Keycloak is responding
curl -k -s https://keycloak.localtest.me/realms/aegis/ | jq '.realm'

# Check Platform-API health
curl -k https://platform-api.localtest.me/healthz

# Check cloudflared tunnel status
kubectl -n aegis-system logs -l app=cloudflared --tail=10

# Check all pods status
kubectl get pods -A | grep -E "aegis-system|keycloak|ingress-nginx"
```

### Debugging Decision Tree

```
Authentication Error?
├── "certificate signed by unknown authority"
│   └── Run: ./scripts/verify-ca-consistency.sh
│       ├── PASS → Check OIDC_JWKS_URL uses correct service name
│       └── FAIL → Run: ./scripts/regenerate-matching-certs.sh
│
├── "Authentication required" in Backstage
│   └── Check Platform-API logs for auth_failure
│       ├── "certificate" error → CA mismatch (see above)
│       └── "invalid issuer" → Check OIDC_ISSUER_URL matches Keycloak
│
├── Keycloak 530/Connection Refused
│   └── kubectl apply -f cloudflared-config.yaml
│
└── "invalid bearer token"
    └── Clear browser cookies and re-login
```

---

## Deployment Commands

### Full Local Deployment

```bash
# Deploy everything with TLS
make deploy-local-tls

# Clean and redeploy (PRESERVES database data)
make clean-local && make deploy-local-tls

# Clean EVERYTHING including databases (use with caution!)
make clean-local-all && make deploy-local-tls
```

### Data Persistence Warning

| Command | Database Data | Use When |
|---------|--------------|----------|
| `make clean-local` | **Preserved** | Normal redeploy, fix issues |
| `make clean-local-all` | **Deleted** | Fresh start, schema changes |

**IMPORTANT:** `make clean-local` preserves PVCs (database data). Use `make clean-local-all` only when you intentionally want to delete all data.

### Individual Components

```bash
# Rebuild and push platform-api image
make build-platform-api push-platform-api

# Apply cloudflared manually
kubectl apply -f cloudflared-config.yaml

# Restart a deployment
kubectl -n aegis-system rollout restart deploy/aegis-services-platform-api
```

### Start Backstage

```bash
cd ~/code/aegis-ui
NODE_EXTRA_CA_CERTS=~/aegis-local-trust.pem yarn dev
```

---

## Verification Checklist

After deployment or Docker Desktop restart:

```bash
# 1. Check all pods are running
kubectl get pods -A | grep -E "aegis-system|keycloak|ingress-nginx"

# 2. Verify Keycloak is accessible
curl -k -s https://keycloak.localtest.me/realms/aegis/ | jq '.realm'
# Expected: "aegis"

# 3. Verify Platform API is accessible
curl -k https://platform-api.localtest.me/healthz
# Expected: OK or similar

# 4. Verify cloudflared is running
kubectl -n aegis-system get pods -l app=cloudflared
kubectl -n aegis-system logs -l app=cloudflared --tail=10
# Expected: "Connection ... registered"

# 5. Verify external access via Cloudflare
curl -s https://keycloak.aegis-platform.tech/realms/aegis/ | jq '.realm'
# Expected: "aegis"

# 6. Verify ingress controller
kubectl -n ingress-nginx get pods

# 7. Check PostgreSQL (platform-api)
kubectl -n aegis-system get pods -l app=platform-api-postgres

# 8. Check PostgreSQL (keycloak)
kubectl -n keycloak get pods -l app=keycloak-postgres
```

---

## Common Issues and Solutions

### After Docker Desktop Restart

**Symptom:** Services not accessible, pods in CrashLoopBackOff

**Solution:**
1. Wait 2-3 minutes for all pods to stabilize
2. Check `kubectl get pods -A`
3. If cloudflared is missing: `kubectl apply -f cloudflared-config.yaml`
4. If persistent issues: `make clean-local && make deploy-local-tls`

### Keycloak 530/Connection Refused

**Symptom:** Browser shows 530 error or ERR_CONNECTION_REFUSED for keycloak.aegis-platform.tech

**Cause:** cloudflared tunnel not deployed or not running

**Solution:**
```bash
kubectl apply -f cloudflared-config.yaml
kubectl -n aegis-system logs -l app=cloudflared --tail=20
```

### Invalid Issuer / Token Errors

**Symptom:** "Invalid issuer" or "Illegal token" errors

**Cause:** Mismatch between Keycloak hostname and Backstage OIDC config

**Solution:**
1. Verify Keycloak issuer:
   ```bash
   curl -s https://keycloak.aegis-platform.tech/realms/aegis/.well-known/openid-configuration | jq -r '.issuer'
   ```
2. Ensure Backstage config (`aegis-ui/.env.development` and `app-config.local-dev.yaml`) uses the same issuer
3. Clear browser cookies
4. Restart Backstage

### Invalid Bearer Token / CA Certificate Mismatch

**Symptom:** `{"error":"Unauthenticated","message":"invalid bearer token"}` from platform-api

**Cause:** Platform-api can't validate tokens because:
1. **JWKS URL uses wrong service name** - The TLS cert has SANs for `aegis-services-keycloak` (without `-service`)
2. **CA mismatch** - The `aegis-platform-api-oidc-ca` secret has a different CA than what signed the Keycloak cert

**To diagnose:**
```bash
# Check platform-api logs for specific error
kubectl -n aegis-system logs deployment/aegis-services-platform-api --tail=20 | grep -i "auth\|token\|cert"

# Common errors:
# - "certificate is valid for X, not Y" → Wrong service name in JWKS URL
# - "certificate signed by unknown authority" → CA mismatch
```

**Solution for service name issue:**
```bash
# JWKS URL must use the short service name that matches TLS cert SAN
kubectl -n aegis-system set env deployment/aegis-services-platform-api \
  OIDC_JWKS_URL="https://aegis-services-keycloak.keycloak.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/certs"
```

**Solution for CA mismatch:**

1. **Verify the mismatch:**
```bash
kubectl -n aegis-system exec deployment/aegis-services-platform-api -- \
  cat /etc/aegis-platform-api/oidc/ca.crt | openssl x509 -noout -subject
kubectl -n keycloak get secret keycloak-tls -o jsonpath='{.data.tls\.crt}' | \
  base64 -d | openssl x509 -noout -issuer
# These MUST both show: O=Aegis Dev, CN=Aegis Local Root CA
```

2. **Run the regeneration script** (see [CA Certificate Consistency](#ca-certificate-consistency-critical) section):
```bash
# The regeneration script generates matching CA + Keycloak cert
# and updates both secrets
```

3. **Make it permanent** by updating `charts/aegis-services/values/local-tls.yaml`:
   - Copy new CA to `platformApi.auth.oidc.caBundle.data`
   - Copy new cert to `keycloak.tls.secret.cert`
   - Copy new key to `keycloak.tls.secret.key`

**Root cause:** The values file contained certificates generated at different times with different CAs. When the cluster was redeployed, the mismatched certs were re-applied. The permanent fix is ensuring `local-tls.yaml` has matching certs.

### Image Not Updating After Build

**Symptom:** Code changes not reflected after `make build-platform-api`

**Cause:** Docker Desktop kind cluster image caching

**Solutions:**
1. Use specific tags instead of `latest`
2. Push to Docker Hub: `make push-platform-api`
3. Delete and recreate pod: `kubectl -n aegis-system delete pod -l app=aegis-services-platform-api`

### Migration Job Failed

**Symptom:** platform-api-migrations job in BackoffLimitExceeded

**Cause:** PostgreSQL not ready or migration SQL error

**Solution:**
```bash
# Check job logs
kubectl -n aegis-system logs job/aegis-services-platform-api-migrations

# If postgres issue, check postgres pod
kubectl -n aegis-system logs -l app=platform-api-postgres

# Delete failed job and redeploy
kubectl -n aegis-system delete job aegis-services-platform-api-migrations
make deploy-local-tls
```

### Self-Signed Certificate Errors

**Symptom:** ERR_CERT_AUTHORITY_INVALID in browser

**Note:** For external access via *.aegis-platform.tech, Cloudflare handles TLS. You should NOT see certificate errors for these domains.

For local *.localtest.me access:
1. Ensure `NODE_EXTRA_CA_CERTS=~/aegis-local-trust.pem` when running Backstage
2. Add CA to browser/system trust store if needed

### AWS InvalidAccessKeyId / Pulumi S3 Errors

**Symptom:**
```
InvalidAccessKeyId: The AWS Access Key Id you provided does not exist in our records
```
or
```
permanent failure: upsert pulumi stack: failed to select stack: exit status 255
```

**Cause:** Wrong AWS credentials in platform-api pod. The platform requires specific terraform-managed credentials.

**Solution:** See [AWS Credentials Architecture](#aws-credentials-architecture) section for full details. Quick fix:

```bash
# Get correct credentials from terraform
cd terraform/pulumi-stack
terraform output pulumi_access_key_id
terraform output pulumi_secret_access_key

# Update and restart (see AWS Credentials section for full steps)
```

---

## Quick Reference Card

```
LOCAL URLS:
  Backstage UI:     http://localhost:3000
  Backstage API:    http://localhost:7008
  Keycloak:         https://keycloak.localtest.me
  Platform API:     https://platform-api.localtest.me
  Platform gRPC:    https://platform-api-grpc.localtest.me

EXTERNAL URLS (via Cloudflare):
  Keycloak:         https://keycloak.aegis-platform.tech
  Platform API:     https://remote.aegis-platform.tech

NAMESPACES:
  aegis-system      Platform API, Proxy, cloudflared
  keycloak          Keycloak
  ingress-nginx     Ingress controller

KEY COMMANDS:
  make deploy-local-tls              Deploy everything
  make clean-local                   Clean all resources
  kubectl apply -f cloudflared-config.yaml   Deploy tunnel
  kubectl -n aegis-system logs -l app=cloudflared   Check tunnel

SPOKE CLUSTER CONNECTIVITY (must be running for EKS clusters):
  make spoke-start      Start tunnel (foreground, Ctrl+C to stop)
  make spoke-check      Check connectivity status
  make spoke-stop       Stop tunnel
  make spoke-fix-nlb    Fix NLB endpoint if misconfigured

  Cluster not showing in UI? Run: make spoke-check
  Then fix what's broken: make spoke-start or make spoke-fix-nlb

CREDENTIALS:
  Keycloak Admin:   admin / REDACTED_KEYCLOAK_ADMIN_PASSWORD
  Backstage OIDC:   backstage / local-backstage-client-secret

AWS CREDENTIALS (terraform-managed):
  IAM User:         aegis-pulumi-provisioner
  Access Key:       REDACTED_AWS_ACCESS_KEY_ID
  Get secret:       cd terraform/pulumi-stack && terraform output pulumi_secret_access_key
  Target Role:      arn:aws:iam::567751785679:role/aegis-platform
```

---

## Appendix: Cloudflare Tunnel Recovery

If the Cloudflare tunnel needs to be recreated from scratch:

1. **Tunnel ID:** a21f6af6-e7b8-49d5-90be-08f0a8d9a829
2. **Account Tag:** aff4ed64e72c1e215be1a4ba105e4ab1
3. **Credentials location:** `cloudflared-config.yaml` (stringData.credentials.json)

The tunnel configuration is fully captured in `/cloudflared-config.yaml`. Simply apply it:
```bash
kubectl apply -f cloudflared-config.yaml
```

If the tunnel itself needs to be recreated in Cloudflare dashboard:
1. Go to Cloudflare Zero Trust dashboard
2. Access > Tunnels
3. Create new tunnel or manage existing
4. Update credentials in `cloudflared-config.yaml`
5. Update DNS records for *.aegis-platform.tech to point to new tunnel

---

## Appendix: VS Code Remote Extension - EKS Spoke Proxy Connectivity

This section documents how to connect the local Aegis VS Code extension to remote workspaces running on AWS EKS spoke clusters.

### Architecture Overview

```
VS Code Extension (local)
    │
    ├─► gRPC to platform-api-grpc.localtest.me:443 (via ingress)
    │   └─► Gets proxy ticket with spoke-proxy URL
    │
    └─► WebSocket to wss://spoke-proxy.<PUBLIC_IP>.nip.io:<NODEPORT>
        └─► Spoke-proxy on EKS (NodePort, e.g., 31484)
            └─► Forwards to workspace pod in aegis-workloads-<project> namespace
```

### Issues and Fixes Reference

#### 1. Spoke Proxy URL Routing (platform-api)

**Problem:** Platform-api returns local proxy URL (`wss://proxy.localtest.me`) instead of EKS spoke-proxy URL.

**Root Cause:** Docker image didn't have the spoke proxy routing code.

**Fix:**
```bash
# Rebuild and push the platform-api image
cd /Users/carlossanchez/code/aegis-platform
make build-platform
docker push carlosmsanchez/aegis-platform-api:dev

# Restart the deployment
kubectl rollout restart deployment/aegis-services-platform-api -n aegis-system
```

**Code Location:** `services/platform-api/internal/server/server.go` (lines 1297-1305)

#### 2. TLS Certificate Hostname Mismatch

**Problem:** Certificate CN doesn't match connection hostname.

**Error:** `Hostname/IP does not match certificate's altnames`

**Fix:** Generate certificate with proper SANs (see provisioning script below).

#### 3. TLS Certificate Chain Verification

**Problem:** Node.js ws library can't verify self-signed certificate.

**Error:** `unable to verify the first certificate`

**Fix:** Add `basicConstraints = critical, CA:TRUE` to the certificate.

### New Cluster/Workspace Provisioning

When deploying a new EKS spoke cluster, follow these steps:

#### Step 1: Generate Spoke-Proxy TLS Certificate

```bash
#!/bin/bash
# Usage: ./generate-spoke-cert.sh <PUBLIC_IP> [NODEPORT]
# Example: ./generate-spoke-cert.sh 107.22.50.137 31484

PUBLIC_IP=$1
NODEPORT=${2:-31484}
HOSTNAME="spoke-proxy.${PUBLIC_IP}.nip.io"

cat > /tmp/spoke-proxy-san.cnf << EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_ca
prompt = no

[req_distinguished_name]
CN = spoke-proxy.aegis.local

[v3_ca]
basicConstraints = critical, CA:TRUE
keyUsage = critical, digitalSignature, keyEncipherment, keyCertSign
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = spoke-proxy.aegis.local
DNS.2 = ${HOSTNAME}
DNS.3 = *.nip.io
IP.1 = ${PUBLIC_IP}
EOF

openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /tmp/spoke-proxy.key \
  -out /tmp/spoke-proxy.crt \
  -subj "/CN=spoke-proxy.aegis.local" \
  -config /tmp/spoke-proxy-san.cnf

echo "Certificate generated:"
echo "  Cert: /tmp/spoke-proxy.crt"
echo "  Key:  /tmp/spoke-proxy.key"
echo "  Hostname: ${HOSTNAME}:${NODEPORT}"
```

#### Step 2: Deploy aegis-spoke Helm Chart

```bash
helm upgrade --install aegis-spoke ./charts/aegis-spoke \
  --namespace aegis-system \
  --create-namespace \
  --set k8sAgent.env.AEGIS_CLUSTER_ID="<CLUSTER_ID>" \
  --set k8sAgent.env.AEGIS_PROXY_INGRESS_HOST="spoke-proxy.<PUBLIC_IP>.nip.io:<NODEPORT>" \
  --set k8sAgent.env.AEGIS_FLAVORS="cpu-small,cpu-medium,cpu-large,gpu-standard,gpu-large" \
  --set proxy.service.type=NodePort \
  --set proxy.service.nodePort=31484 \
  --set-file proxy.tls.cert=/tmp/spoke-proxy.crt \
  --set-file proxy.tls.key=/tmp/spoke-proxy.key
```

#### Step 3: Verify Cluster Registration

The k8s-agent auto-registers the cluster. Verify the `proxy_url` is set:

```bash
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U platformapi -d platformapi -c \
  "SELECT id, proxy_url FROM clusters;"
```

If `proxy_url` is empty, manually set it:
```bash
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U platformapi -d platformapi -c \
  "UPDATE clusters SET proxy_url = 'wss://spoke-proxy.<IP>.nip.io:<PORT>' WHERE id = '<CLUSTER_ID>';"
```

#### Step 4: Update Client Trust Bundle

Add the spoke-proxy certificate to the local trust bundle:

```bash
cat /tmp/spoke-proxy.crt >> ~/aegis-local-trust.pem
```

Trust bundle contents (example):
```
~/aegis-local-trust.pem
├── platform-api.localtest.me cert (signed by Aegis Local Root CA)
├── Aegis Local Root CA (O=Aegis Dev, CN=Aegis Local Root CA)
├── keycloak.localtest.me cert
├── spoke-proxy-cluster-1.crt (self-signed with CA:TRUE)
├── spoke-proxy-cluster-2.crt (add as needed for each cluster)
└── ...
```

#### Step 5: Update EKS Spoke-Proxy TLS Secret (if regenerating)

```bash
# Delete old secret and create new one
kubectl --kubeconfig=/tmp/eks-kubeconfig.yaml -n aegis-system \
  delete secret aegis-spoke-proxy-tls

kubectl --kubeconfig=/tmp/eks-kubeconfig.yaml -n aegis-system \
  create secret tls aegis-spoke-proxy-tls \
  --cert=/tmp/spoke-proxy.crt \
  --key=/tmp/spoke-proxy.key

# Restart spoke-proxy to pick up new cert
kubectl --kubeconfig=/tmp/eks-kubeconfig.yaml -n aegis-system \
  rollout restart deployment/aegis-spoke-proxy
```

### VS Code Extension Settings

**File:** `~/Library/Application Support/Code/User/settings.json`

```json
{
  "aegisRemote.platform.grpcEndpoint": "platform-api-grpc.localtest.me:443",
  "aegisRemote.platform.projectId": "db-1",
  "aegisRemote.security.caPath": "/Users/<username>/aegis-local-trust.pem",
  "aegisRemote.security.rejectUnauthorized": true,
  "aegisRemote.auth.authority": "https://keycloak.aegis-platform.tech/realms/aegis",
  "aegisRemote.auth.clientId": "vscode-extension"
}
```

### Running the Extension in Development

```bash
"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code" \
  --extensionDevelopmentPath=/Users/carlossanchez/code/sovran/aegis-vscode-remote/extension \
  --enable-proposed-api aegis.aegis-remote
```

### Workload Reset (when stuck in FAILED state)

```bash
# Delete workload resources on EKS
kubectl --kubeconfig=/tmp/eks-kubeconfig.yaml \
  delete aegisworkload aegis-<WORKLOAD_ID> -n aegis-workloads-<project>
kubectl --kubeconfig=/tmp/eks-kubeconfig.yaml \
  delete job aegis-<WORKLOAD_ID> -n aegis-workloads-<project>

# Reset status in postgres
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U platformapi -d platformapi -c \
  "UPDATE workloads SET status = 'PENDING', cluster_id = NULL, placed_at = NULL, started_at = NULL WHERE id = '<WORKLOAD_ID>';"

# Manually assign to cluster if scheduler doesn't pick it up
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U platformapi -d platformapi -c \
  "UPDATE workloads SET cluster_id = '<CLUSTER_ID>', placed_at = NOW() WHERE id = '<WORKLOAD_ID>';"
```

### Verification Commands

```bash
# Check spoke-proxy cert SANs
echo | openssl s_client -connect spoke-proxy.<IP>.nip.io:<NODEPORT> \
  -servername spoke-proxy.<IP>.nip.io 2>/dev/null | \
  openssl x509 -text -noout | grep -A2 "Subject Alternative Name"

# Check workload status on EKS
kubectl --kubeconfig=/tmp/eks-kubeconfig.yaml \
  get aegisworkload -n aegis-workloads-<project>

# Check spoke-proxy pod
kubectl --kubeconfig=/tmp/eks-kubeconfig.yaml \
  get pods -n aegis-system | grep spoke-proxy

# Check cluster proxy_url in database
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U platformapi -d platformapi -c "SELECT id, proxy_url FROM clusters;"
```

### Key Requirements for Self-Signed Spoke-Proxy Certs

| Requirement | Value | Why |
|-------------|-------|-----|
| `basicConstraints` | `CA:TRUE` | Required for Node.js to trust as CA |
| `subjectAltName` | Include actual hostname/IP | TLS verification requires SAN match |
| `keyCertSign` in keyUsage | Required | Needed when CA:TRUE is set |

### Summary: What's Stored Where

| Item | Location | When to Update |
|------|----------|----------------|
| Spoke-proxy TLS cert/key | EKS secret `aegis-spoke-proxy-tls` | New cluster or IP change |
| Cluster `proxy_url` | PostgreSQL `clusters` table | Auto-registered by k8s-agent |
| Client trust bundle | `~/aegis-local-trust.pem` | Add cert for each new cluster |
| Helm values | CLI flags or values file | Each deployment |

---

*This document should be updated whenever infrastructure changes are made.*
