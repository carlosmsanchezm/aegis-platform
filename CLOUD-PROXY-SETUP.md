# Cloud Proxy SSH Connection - Complete Setup Guide

This guide documents the complete setup and testing of the Aegis cloud proxy infrastructure, which enables secure SSH connections from local machines to workspace pods running in remote Kubernetes clusters through the public internet.

## Architecture Overview

```
Local Machine → aegis-connect (CLI) → Internet → AWS NLB → aegis-auth-proxy (EKS) → Workspace Pod (EKS)
     ↓                                                           ↓
  ~/.ssh/config                                          JWT Token Auth
```

## Components

1. **Platform-API**: Manages workloads and creates connection sessions with JWT tokens
2. **PostgreSQL Database**: Stores workloads, connection sessions, and tokens
3. **aegis-auth-proxy**: HTTPS proxy server that authenticates JWT tokens and forwards SSH connections
4. **AWS Network Load Balancer (NLB)**: Exposes proxy to public internet
5. **aegis-connect**: Local CLI tool that establishes CONNECT tunnels through the proxy
6. **Workspace Pods**: Kubernetes pods running SSH servers

## Prerequisites

- Kubernetes cluster with Aegis services deployed
- AWS RDS PostgreSQL database
- kubectl configured with cluster access
- grpcurl installed (`brew install grpcurl`)
- sshpass installed (`brew install hudochenkov/sshpass/sshpass`)
- jq installed (`brew install jq`)

## Step 1: Database Setup

### 1.1 Ensure Database Schema is Complete

The workloads table needs all required columns:

```bash
kubectl run psql-migration --rm -i --restart=Never --image=postgres:15 \
  --env="PGPASSWORD=<your-db-password>" -- \
  psql -h <your-rds-endpoint> -U aegis_api -d aegis <<'EOF'
-- Verify workloads table has all columns
\d workloads

-- Add missing columns if needed (these should already exist)
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS ui_status TEXT;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS url TEXT;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS message TEXT;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS kind TEXT;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS hints_resource_name TEXT;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS hints_gpu_count INTEGER;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS hints_cpu_request TEXT;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS hints_mem_request TEXT;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS workspace_json JSONB;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS training_json JSONB;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS placed_at TIMESTAMPTZ;
EOF
```

### 1.2 Create Connection Sessions Tables

```bash
kubectl run psql-migration --rm -i --restart=Never --image=postgres:15 \
  --env="PGPASSWORD=<your-db-password>" -- \
  psql -h <your-rds-endpoint> -U aegis_api -d aegis <<'EOF'
CREATE TABLE IF NOT EXISTS connection_sessions (
    session_id TEXT PRIMARY KEY,
    workload_id TEXT NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
    subject TEXT NOT NULL,
    client TEXT NOT NULL,
    jti TEXT NOT NULL UNIQUE,
    token TEXT NOT NULL,
    ssh_user TEXT NOT NULL,
    ssh_host_alias TEXT NOT NULL,
    internal_host TEXT NOT NULL,
    port INTEGER NOT NULL,
    ssh_config TEXT NOT NULL,
    proxy_url TEXT NOT NULL,
    vscode_uri TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    one_time BOOLEAN NOT NULL DEFAULT TRUE,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS sessions_by_workload ON connection_sessions(workload_id);
CREATE INDEX IF NOT EXISTS sessions_by_expires ON connection_sessions(expires_at);

CREATE TABLE IF NOT EXISTS session_jtis (
    jti TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES connection_sessions(session_id) ON DELETE CASCADE,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL
);
EOF
```

### 1.3 Verify Platform-API is Using PostgreSQL Backend

```bash
# Check environment variables
kubectl get deployment -n aegis-services aegis-services-aegis-services-platform-api \
  -o jsonpath='{.spec.template.spec.containers[0].env[*]}' | jq

# Should show:
# - AEGIS_STORE_BACKEND: "postgres"
# - DB_HOST: "<rds-endpoint>"
# - DB_NAME: "aegis"
# - DB_USER: "aegis_api"
# - DB_PASSWORD: (from secret)

# Check logs to confirm PostgreSQL is being used
kubectl logs -n aegis-services deployment/aegis-services-aegis-services-platform-api | grep "store"
# Should show: "using PostgreSQL store"
```

## Step 2: Proxy Service Setup

### 2.1 Change Proxy Service to LoadBalancer Type

Edit the helm values:

```yaml
# charts/aegis-services/values-test-cloud.yaml
proxy:
  service:
    type: LoadBalancer  # Changed from ClusterIP
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
      service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
```

Deploy the change:

```bash
helm upgrade aegis-services ./charts/aegis-services \
  -n aegis-services \
  -f charts/aegis-services/values-test-cloud.yaml
```

### 2.2 Get the NLB DNS Name

```bash
kubectl get svc -n aegis-services aegis-services-aegis-services-proxy

# Output will show EXTERNAL-IP like:
# a9604fc12441d48e6bedf773e3f63996-98644661.us-east-1.elb.amazonaws.com
```

### 2.3 Configure Platform-API with Proxy URL

Get the proxy JWT secret:

```bash
PROXY_SECRET=$(kubectl get secret -n aegis-services aegis-services-aegis-services-proxy-secret \
  -o jsonpath='{.data.jwt-secret}' | base64 -d)
echo $PROXY_SECRET
```

Update platform-api deployment:

```bash
kubectl set env deployment/aegis-services-aegis-services-platform-api \
  -n aegis-services \
  AEGIS_PROXY_BASE_URL="https://<nlb-dns-name>:8080" \
  AEGIS_PROXY_JWT_SECRET="$PROXY_SECRET"
```

Wait for platform-api to restart:

```bash
kubectl rollout status deployment/aegis-services-aegis-services-platform-api -n aegis-services
```

## Step 3: K8s-Agent Configuration

### 3.1 Set Bootstrap Image for Init Containers

The k8s-agent uses an init container for SSH setup. Set a multi-arch bootstrap image:

```bash
kubectl set env deployment/aegis-spoke-aegis-spoke-k8s-agent \
  -n aegis-spoke \
  AEGIS_SSH_BOOTSTRAP_IMAGE=busybox:1.36
```

Wait for k8s-agent to restart:

```bash
kubectl rollout status deployment/aegis-spoke-aegis-spoke-k8s-agent -n aegis-spoke
```

## Step 4: Build and Install aegis-connect CLI Tool

### 4.1 Modify for InsecureSkipVerify (Testing Only)

For testing with self-signed certificates, modify the TLS config:

```go
// services/proxy/cmd/aegis-proxy-client/main.go
// Line 42 (approximately):
conn, err = tls.Dial("tcp", host, &tls.Config{
    MinVersion: tls.VersionTLS12,
    InsecureSkipVerify: true  // Add this line for testing
})
```

### 4.2 Build and Install

```bash
cd services/proxy
go build -o ~/.local/bin/aegis-connect ./cmd/aegis-proxy-client
chmod +x ~/.local/bin/aegis-connect

# Verify installation
which aegis-connect
aegis-connect --help
```

## Step 5: Create SSH-Enabled Workspace

### 5.1 Create Project, Flavor, Queue, and Budget (if needed)

```bash
# Set up port-forward to platform-api
kubectl port-forward -n aegis-services svc/aegis-services-aegis-services-platform-api 8081:8081 &

# Create project
grpcurl -plaintext -d '{
  "project":{"id":"p-demo","displayName":"Demo","ownerGroup":"eng"}
}' localhost:8081 aegis.v1.AegisPlatform/CreateProject

# Create flavor
grpcurl -plaintext -d '{
  "flavor": {
    "name": "nano",
    "cpu_cores_request": "100m",
    "memory_request": "512Mi",
    "price_usd_per_gpu_hour": 0
  }
}' localhost:8081 aegis.v1.AegisPlatform/UpsertFlavor

# Create queue
grpcurl -plaintext -d '{
  "queue":{
    "name":"default",
    "projectId":"p-demo",
    "allowedFlavors":["nano"]
  }
}' localhost:8081 aegis.v1.AegisPlatform/UpsertQueue

# Create budget
grpcurl -plaintext -d '{
  "budget":{"projectId":"p-demo","queue":"default","limitUsd":100,"policyMode":"SOFT"}
}' localhost:8081 aegis.v1.AegisPlatform/UpsertBudget
```

### 5.2 Create Workspace with SSH

Create an AegisWorkload CR directly:

```bash
kubectl apply -f - <<'EOF'
apiVersion: aegis.yourorg.dev/v1alpha1
kind: AegisWorkload
metadata:
  name: wl-ssh-workspace
  namespace: default
spec:
  projectId: p-demo
  queue: default
  workspace:
    flavor: nano
    image: ubuntu:22.04
    interactive: true
    ports:
      - 2222
    command:
      - /bin/bash
      - -c
      - |
        apt-get update &&
        DEBIAN_FRONTEND=noninteractive apt-get install -y openssh-server &&
        mkdir -p /run/sshd &&
        useradd -m -s /bin/bash aegis &&
        echo 'aegis:aegis123' | chpasswd &&
        sed -i 's/#Port 22/Port 2222/' /etc/ssh/sshd_config &&
        sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config &&
        sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config &&
        /usr/sbin/sshd -D -p 2222
  hints:
    cpuCoresRequest: "100m"
    memoryRequest: "512Mi"
EOF
```

### 5.3 Wait for Workspace to Start

```bash
# Watch pod status
kubectl get pods -n default -w | grep wl-ssh-workspace

# Check logs to ensure SSH is running (wait for apt install to complete)
kubectl logs -n default -f $(kubectl get pods -n default -l aegis.workload/id=wl-ssh-workspace -o name)

# Verify SSH is running (after ~2-3 minutes)
kubectl exec -n default $(kubectl get pods -n default -l aegis.workload/id=wl-ssh-workspace -o name) \
  -- ps aux | grep sshd
```

### 5.4 Add Workspace to Database

Since we created the CR directly, add it to the database:

```bash
kubectl run psql-query --rm -i --restart=Never --image=postgres:15 \
  --env="PGPASSWORD=<your-db-password>" -- \
  psql -h <your-rds-endpoint> -U aegis_api -d aegis <<'EOF'
INSERT INTO workloads (id, project_id, queue, cluster_id, status, kind, workspace_json, created_at, updated_at)
VALUES (
  'wl-ssh-workspace',
  'p-demo',
  'default',
  'aegis-spoke-prod',
  'RUNNING',
  'workspace',
  '{"flavor":"nano","image":"ubuntu:22.04","interactive":true,"ports":[2222]}',
  now(),
  now()
)
ON CONFLICT (id) DO UPDATE SET status='RUNNING', updated_at=now();
EOF
```

## Step 6: Test SSH Connection

### 6.1 Verify In-Cluster Connectivity

First test SSH works within the cluster:

```bash
kubectl run test-ssh --rm -i --restart=Never --image=alpine -- \
  sh -c "apk add --no-cache openssh-client sshpass && \
  sshpass -p 'aegis123' ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
  aegis@aegis-w-wl-ssh-workspace.default.svc.cluster.local -p 2222 \
  'echo SSH_CONNECTION_SUCCESS'"

# Should output: SSH_CONNECTION_SUCCESS
```

### 6.2 Create Connection Session

```bash
TOKEN=$(./scripts/keycloak-token.sh)

grpcurl -plaintext -H "authorization: Bearer ${TOKEN}" \
  -d '{"workload_id":"wl-ssh-workspace","client":"cli"}' \
  localhost:8081 aegis.v1.AegisPlatform/CreateConnectionSession
```

This returns a JSON response with:
- `sessionId`: Unique session identifier
- `token`: JWT token for authentication
- `sshConfig`: Ready-to-use SSH config
- `proxyUrl`: HTTPS URL to the proxy
- `expiresAtUtc`: Token expiration time
- `oneTime`: Boolean (tokens are one-time use)

### 6.3 Save SSH Config

```bash
# Get the SSH config and save it
grpcurl -plaintext -H "authorization: Bearer ${TOKEN}" \
  -d '{"workload_id":"wl-ssh-workspace","client":"cli"}' \
  localhost:8081 aegis.v1.AegisPlatform/CreateConnectionSession \
  | jq -r '.sshConfig' > ~/.ssh/aegis-workspace-config

# Important: Change the user from aegis-d3b42d2c to aegis (the actual SSH user)
sed -i '' 's/User aegis-[a-z0-9]*/User aegis/' ~/.ssh/aegis-workspace-config
```

### 6.4 Test SSH Connection Through Cloud Proxy

```bash
# Test with sshpass (password: aegis123)
sshpass -p 'aegis123' ssh -F ~/.ssh/aegis-workspace-config \
  -o StrictHostKeyChecking=no \
  aegis-w-wl-ssh-workspace \
  'echo "=== SSH SUCCESS ===" && hostname && whoami'
```

**Expected Output:**
```
=== SSH SUCCESS ===
aegis-wl-ssh-workspace-xxxxx
aegis
```

### 6.5 Run Additional Commands

```bash
# Check system info
sshpass -p 'aegis123' ssh -F ~/.ssh/aegis-workspace-config \
  aegis-w-wl-ssh-workspace \
  'uname -a && df -h / && free -h'

# Install packages
sshpass -p 'aegis123' ssh -F ~/.ssh/aegis-workspace-config \
  aegis-w-wl-ssh-workspace \
  'sudo apt-get update && sudo apt-get install -y python3 pip'
```

**Note:** Each SSH command requires a new connection session because tokens are one-time use. For interactive sessions, you would use a non-expiring token or implement token renewal.

## Step 7: Verify Proxy Logs

Check that the proxy is logging successful connections:

```bash
kubectl logs -n aegis-services deployment/aegis-services-aegis-services-proxy --tail=20

# Look for entries like:
# {"event":"session.start",...,"mode":"connect"}
# {"event":"session.stop",...,"duration":0.875,"bytes_tx":4013,"bytes_rx":4650}
```

## Architecture Details

### Connection Flow

1. **Client requests connection**: User calls `CreateConnectionSession` API
2. **Platform-API generates JWT**: Contains workload ID, destination, cluster, expiry
3. **Client initiates SSH**: Uses `aegis-connect` as ProxyCommand
4. **aegis-connect establishes CONNECT tunnel**:
   - Opens TLS connection to NLB
   - Sends HTTP CONNECT request with JWT in Authorization header
5. **Proxy validates JWT**:
   - Checks signature against shared secret
   - Verifies expiry and audience
   - Marks one-time token as used
6. **Proxy dials workspace Service**:
   - Connects to `aegis-w-<workload-id>.default.svc.cluster.local:<port>`
7. **Bidirectional proxy**: All SSH traffic flows through established tunnel
8. **Session tracking**: Proxy logs bytes transferred and duration

### Security Features

- **JWT Token Authentication**: All connections require valid JWT tokens
- **One-Time Use Tokens**: Tokens are invalidated after first use
- **Token Expiration**: Default 5-minute TTL for connection establishment
- **TLS Encryption**: All traffic encrypted between client and proxy
- **Network Isolation**: Workspace pods only accessible through authenticated proxy
- **Audit Logging**: All connection attempts and data transfer logged

### Database Schema

**workloads table:**
- Stores workload metadata (project, cluster, status, etc.)
- `workspace_json` contains interactive workspace configuration
- `training_json` contains training job configuration

**connection_sessions table:**
- Stores active connection sessions
- Links to workload via `workload_id`
- Contains JWT token and SSH configuration
- Tracks usage and revocation status

**session_jtis table:**
- Tracks JWT token IDs (JTI) to enforce one-time use
- Prevents token replay attacks

## Troubleshooting

### Issue: "exec format error" in Workspace Pod

**Cause:** Image architecture mismatch (ARM64 vs AMD64)

**Solution:** Use multi-arch images or specify architecture:
```yaml
image: lscr.io/linuxserver/openssh-server:amd64-latest
```

### Issue: "connection refused" on SSH Port

**Cause:** SSH server not running or wrong port

**Solution:**
1. Check SSH process: `kubectl exec <pod> -- ps aux | grep sshd`
2. Verify port matches workspace spec (2222)
3. Check logs: `kubectl logs <pod>`

### Issue: "access denied" from Proxy

**Cause:** JWT token already used (one-time tokens)

**Solution:** Create a new connection session for each SSH command

### Issue: "workload not found" when Creating Session

**Cause:** Workload not in database

**Solution:** Insert workload into database (see Step 5.4)

### Issue: TLS Certificate Verification Failed

**Cause:** Self-signed certificates on proxy

**Solution:** Use `InsecureSkipVerify: true` in aegis-connect (testing only)

For production, use proper certificates:
- AWS Certificate Manager (ACM) certificates on NLB
- cert-manager for in-cluster TLS
- Update aegis-connect to trust CA

## Testing GPU Workspaces

To test with GPU-enabled workspaces:

### 1. Create GPU Flavor

```bash
grpcurl -plaintext -d '{
  "flavor": {
    "name": "a10-1gpu",
    "chip": "nvidia-a10",
    "gpu_count": 1,
    "memory_gib": 24,
    "resource_name": "nvidia.com/gpu",
    "cpu_cores_request": "4",
    "memory_request": "16Gi",
    "price_usd_per_gpu_hour": 1.50
  }
}' localhost:8081 aegis.v1.AegisPlatform/UpsertFlavor
```

### 2. Create GPU Workspace

```bash
kubectl apply -f - <<'EOF'
apiVersion: aegis.yourorg.dev/v1alpha1
kind: AegisWorkload
metadata:
  name: wl-gpu-workspace
  namespace: default
spec:
  projectId: p-demo
  queue: default
  workspace:
    flavor: a10-1gpu
    image: nvidia/cuda:12.2.0-devel-ubuntu22.04
    interactive: true
    ports:
      - 2222
    command:
      - /bin/bash
      - -c
      - |
        apt-get update &&
        DEBIAN_FRONTEND=noninteractive apt-get install -y openssh-server &&
        mkdir -p /run/sshd &&
        useradd -m -s /bin/bash aegis &&
        echo 'aegis:aegis123' | chpasswd &&
        sed -i 's/#Port 22/Port 2222/' /etc/ssh/sshd_config &&
        sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config &&
        sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config &&
        /usr/sbin/sshd -D -p 2222
  hints:
    resourceName: "nvidia.com/gpu"
    gpuCount: 1
    cpuCoresRequest: "4"
    memoryRequest: "16Gi"
EOF
```

### 3. Verify GPU Access

After connecting via SSH:

```bash
sshpass -p 'aegis123' ssh -F ~/.ssh/aegis-gpu-workspace-config \
  aegis-w-wl-gpu-workspace \
  'nvidia-smi'
```

## Next Steps

1. **Key-Based Authentication**: Generate SSH keys for passwordless access
2. **VS Code Remote**: Test VS Code Remote-SSH extension
3. **Production Certificates**: Replace self-signed certs with ACM/cert-manager
4. **Token Renewal**: Implement long-lived sessions with token refresh
5. **Multi-Cluster**: Test workspaces across different clusters
6. **Port Forwarding**: Test SSH tunnel for additional ports (Jupyter, TensorBoard, etc.)

## Reference Files

- Platform-API: `services/platform-api/internal/server/server.go`
- Proxy Server: `services/proxy/internal/server/server.go`
- Proxy Client: `services/proxy/cmd/aegis-proxy-client/main.go`
- K8s Agent: `agents/k8s-agent/internal/controller/aegisworkload_controller.go`
- Workspace Builder: `agents/k8s-agent/internal/workload/builders/workspace.go`
- Helm Values: `charts/aegis-services/values-test-cloud.yaml`

## Success Criteria Checklist

- [ ] Database has all required tables and columns
- [ ] Platform-API using PostgreSQL backend
- [ ] Proxy service exposed via LoadBalancer with NLB DNS
- [ ] Platform-API configured with proxy URL and JWT secret
- [ ] aegis-connect CLI tool built and installed
- [ ] Workspace pod running with SSH server on port 2222
- [ ] In-cluster SSH connectivity verified
- [ ] Connection session created with valid JWT token
- [ ] SSH connection successful through cloud proxy
- [ ] Proxy logs show successful session start/stop
- [ ] Commands execute successfully in workspace

All checkboxes completed = **Production Ready! 🎉**
