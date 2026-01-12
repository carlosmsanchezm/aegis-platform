# Spoke Proxy Automation - Architecture & Implementation Guide

This document describes the automated spoke-proxy provisioning system implemented to enable remote workspace connectivity from VS Code to EKS clusters.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Component Details](#component-details)
4. [Data Flow](#data-flow)
5. [Configuration](#configuration)
6. [Current Limitations](#current-limitations)
7. [Troubleshooting](#troubleshooting)
8. [Future Work for Scalability & Multi-tenancy](#future-work-for-scalability--multi-tenancy)

---

## Overview

### Problem Statement

When a user deploys a workspace on a remote EKS cluster, the VS Code extension needs to connect to that workspace via WebSocket. This requires:

1. A publicly accessible proxy endpoint on the EKS cluster
2. TLS encryption for the WebSocket connection
3. Client-side trust of the proxy's TLS certificate

Previously, all of these required manual configuration after cluster provisioning.

### Solution

The system now automates:

| Component | Before | After |
|-----------|--------|-------|
| Security Group | Manual AWS console rule | Pulumi creates rule automatically |
| Spoke Proxy | Disabled by default | Enabled with NodePort service |
| TLS Certificate | Manual `openssl` generation | Auto-generated per cluster |
| Node IP Discovery | Manual lookup | k8s-agent auto-discovers via AWS metadata |
| proxy_url Registration | Manual DB update | k8s-agent registers in heartbeat |
| Client Trust Bundle | Manual `openssl` export | Manual - add cert to `~/aegis-local-trust.pem` |

---

## Architecture

### High-Level Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              CLUSTER PROVISIONING                                    │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐               │
│  │  Platform API    │    │  Pulumi Runner   │    │  AWS EKS         │               │
│  │  (Provisioning   │───►│                  │───►│                  │               │
│  │   Controller)    │    │  1. Create VPC   │    │  - Cluster       │               │
│  └──────────────────┘    │  2. Create EKS   │    │  - Node Groups   │               │
│                          │  3. Security     │    │  - Security      │               │
│                          │     Groups       │    │    Groups        │               │
│                          │  4. Generate TLS │    └──────────────────┘               │
│                          │     Certificate  │              │                        │
│                          │  5. Deploy Helm  │              │                        │
│                          │     (aegis-spoke)│              ▼                        │
│                          └──────────────────┘    ┌──────────────────┐               │
│                                                  │  aegis-spoke     │               │
│                                                  │  Helm Release    │               │
│                                                  │  - k8s-agent     │               │
│                                                  │  - spoke-proxy   │               │
│                                                  │  - TLS Secret    │               │
│                                                  └──────────────────┘               │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              CLUSTER REGISTRATION                                    │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐               │
│  │  k8s-agent       │    │  Platform API    │    │  PostgreSQL      │               │
│  │                  │───►│  (gRPC)          │───►│                  │               │
│  │  1. Discover     │    │                  │    │  clusters table  │               │
│  │     node IP      │    │  RegisterCluster │    │  - proxy_url     │               │
│  │  2. Construct    │    │  Heartbeat       │    │                  │               │
│  │     proxy URL    │    │                  │    │                  │               │
│  │  3. Send in      │    └──────────────────┘    └──────────────────┘               │
│  │     heartbeat    │                                                               │
│  └──────────────────┘                                                               │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT CONNECTION                                       │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐               │
│  │  VS Code         │    │  Platform API    │    │  spoke-proxy     │               │
│  │  Extension       │    │  (HTTP)          │    │  (EKS NodePort)  │               │
│  │                  │    │                  │    │                  │               │
│  │  1. User trusts  │    │                  │    │                  │               │
│  │     cert locally │    │                  │    │                  │               │
│  │                  │    │                  │    │                  │               │
│  │  2. Connect via  │────────────────────────────►  WebSocket       │               │
│  │     WebSocket    │    │                  │    │  wss://spoke-    │               │
│  │                  │    │                  │    │  proxy.IP.nip.io │               │
│  └──────────────────┘    └──────────────────┘    └──────────────────┘               │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### Network Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                   INTERNET                                           │
└───────────────────────────────────────┬─────────────────────────────────────────────┘
                                        │
                                        │ Port 31484 (NodePort)
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              AWS EKS Cluster                                         │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │                         Node Security Group                                  │    │
│  │   Inbound Rules:                                                            │    │
│  │   - TCP 31484 from 0.0.0.0/0 (spoke-proxy NodePort)                        │    │
│  │   - TCP 443 from Cluster SG (control plane)                                 │    │
│  │   - All from self (node-to-node)                                            │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
│                                        │                                             │
│                                        ▼                                             │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │                              Node (EC2)                                      │    │
│  │   Public IP: x.x.x.x                                                        │    │
│  │                                                                              │    │
│  │   ┌─────────────────────┐    ┌─────────────────────┐                        │    │
│  │   │   spoke-proxy Pod   │    │   k8s-agent Pod     │                        │    │
│  │   │                     │    │                     │                        │    │
│  │   │   Port 8085 ◄───────│────│   Discovers IP      │                        │    │
│  │   │   TLS enabled       │    │   Sends heartbeat   │                        │    │
│  │   │                     │    │                     │                        │    │
│  │   └─────────────────────┘    └─────────────────────┘                        │    │
│  │            ▲                                                                 │    │
│  │            │ NodePort 31484                                                  │    │
│  │   ┌────────┴────────────┐                                                    │    │
│  │   │   spoke-proxy-svc   │                                                    │    │
│  │   │   Type: NodePort    │                                                    │    │
│  │   │   Port: 443         │                                                    │    │
│  │   │   NodePort: 31484   │                                                    │    │
│  │   └─────────────────────┘                                                    │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘

DNS: spoke-proxy.{NODE_PUBLIC_IP}.nip.io:31484
Example: wss://spoke-proxy.3.80.219.92.nip.io:31484
```

---

## Component Details

### 1. Pulumi Runner (`runner.go`)

**Location:** `services/platform-api/internal/provisioning/pulumi/aws/runner.go`

**Responsibilities:**

| Function | Description |
|----------|-------------|
| `createSecurityGroups()` | Creates node security group with NodePort 31484 rule |
| `installSpokeHelmChart()` | Deploys aegis-spoke with proxy enabled |
| `generateSpokeProxyCert()` | Generates self-signed TLS certificate with *.nip.io wildcard |
| `getSpokeProxyNodePort()` | Returns configurable NodePort (env or default 31484) |

**Key Configuration:**

```go
const (
    defaultSpokeProxyNodePort = 31484  // Configurable via AEGIS_SPOKE_PROXY_NODEPORT
)

// Security group rule
awsec2.NewSecurityGroupRule("spoke-proxy-nodeport", {
    Type:     "ingress",
    Protocol: "tcp",
    FromPort: nodePort,  // 31484
    ToPort:   nodePort,
    CidrBlocks: ["0.0.0.0/0"],
})

// Helm values for spoke-proxy
proxyValues := pulumi.Map{
    "enabled": pulumi.Bool(true),
    "service": pulumi.Map{
        "type":     "NodePort",
        "nodePort": nodePort,
        "port":     443,
    },
}
```

**TLS Certificate Generation:**

```go
func generateSpokeProxyCert() (certPEM, keyPEM string, err error)
```

- Creates 2048-bit RSA key
- Self-signed CA certificate (valid 1 year)
- DNS SANs: `*.nip.io`, `spoke-proxy.*.nip.io`, `localhost`
- Used for all spoke proxies (wildcard approach)

### 2. k8s-Agent (`aegisworkload_controller.go`)

**Location:** `agents/k8s-agent/internal/controller/aegisworkload_controller.go`

**Responsibilities:**

| Function | Description |
|----------|-------------|
| `buildProxyURL()` | Auto-discovers node public IP, constructs nip.io URL |
| `discoverNodePublicIP()` | Queries AWS metadata or external services for public IP |

**Node IP Discovery:**

```go
func discoverNodePublicIP() string {
    // 1. Try AWS EC2 metadata service (IMDSv1)
    if ip := fetchURL("http://169.254.169.254/latest/meta-data/public-ipv4", 2*time.Second); ip != "" {
        return ip
    }
    // 2. Fallback to external services
    services := []string{
        "https://api.ipify.org",
        "https://checkip.amazonaws.com",
        "https://ifconfig.me/ip",
    }
    // ... try each service
}
```

### 3. Heartbeat Client (`client.go`)

**Location:** `agents/k8s-agent/internal/cpclient/client.go`

```go
// HeartbeatLoop sends periodic heartbeats with proxy URL
func (c *Client) HeartbeatLoop(ctx context.Context, logger *zap.Logger,
    clusterID string, flavors []*aegis.Flavor, proxyURL string) {

    c.api.Heartbeat(ctx, &aegis.ClusterHeartbeat{
        ClusterId:        clusterID,
        TtfGpuSecondsP50: ttf,
        AvailableFlavors: flavors,
        ProxyUrl:         proxyURL,   // e.g., "wss://spoke-proxy.3.80.219.92.nip.io:31484"
    })
}
```

### 4. Platform API Store

**Location:** `services/platform-api/internal/store/`

**ClusterInfo Struct:**

```go
type ClusterInfo struct {
    ID                 string
    ProjectID          string
    Provider           string
    Region             string
    Labels             map[string]string
    AvailableFlavorSet map[string]bool
    TTFGSecondsP50     float64
    LastHeartbeat      time.Time
    ProxyURL           string  // e.g., "wss://spoke-proxy.3.80.219.92.nip.io:31484"
    CreatedAt          time.Time
    DeletedAt          *time.Time
}
```

### 5. Proto Definitions

**Location:** `proto/aegis/v1/platform.proto`

```protobuf
message ClusterRegisterRequest {
  string cluster_id = 1;
  string provider = 2;
  string region = 3;
  string il_level = 4;
  map<string,string> labels = 5;
  string proxy_url = 6;   // spoke proxy URL
}

message ClusterHeartbeat {
  string cluster_id = 1;
  double ttf_gpu_seconds_p50 = 2;
  repeated Flavor available_flavors = 3;
  string proxy_url = 4;   // spoke proxy URL
}
```

---

## Data Flow

### Provisioning Flow

```
1. User creates ProjectInfra CR
   └─► ProjectInfra Controller detects new CR

2. Pulumi Runner starts provisioning
   ├─► Creates VPC, subnets, security groups
   │   └─► Security group includes NodePort 31484 rule
   ├─► Creates EKS cluster and node groups
   ├─► Generates TLS certificate (generateSpokeProxyCert)
   └─► Deploys aegis-spoke Helm chart
       ├─► proxy.enabled: true
       ├─► proxy.service.type: NodePort
       ├─► proxy.service.nodePort: 31484
       └─► proxy.tls.cert/key: <generated cert>

3. k8s-agent pod starts
   ├─► Discovers node public IP (AWS metadata)
   ├─► Constructs proxy URL: wss://spoke-proxy.{IP}.nip.io:31484
   └─► Registers with platform-api (includes proxy_url)

4. Platform-api stores in database
   └─► clusters table: proxy_url column
```

### Connection Flow

```
1. User clicks "Connect" on workspace in UI
   └─► UI calls platform-api to create connection session

2. Platform-api creates session
   ├─► Looks up cluster's proxy_url from database
   ├─► Generates JWT token for workspace access
   └─► Returns session with proxy_url

3. VS Code extension receives session
   ├─► Uses locally-trusted certificate (~/.aegis-local-trust.pem)
   └─► Connects to wss://spoke-proxy.{IP}.nip.io:31484/proxy/{workload-id}

4. spoke-proxy validates JWT and proxies to workspace pod
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `AEGIS_SPOKE_PROXY_NODEPORT` | 31484 | NodePort for spoke-proxy service |
| `POD_NAMESPACE` | aegis-system | Namespace where k8s-agent runs |
| `AEGIS_PROXY_URL` | (auto) | Override proxy URL (bypasses auto-discovery) |
| `AEGIS_PROXY_INGRESS_HOST` | (auto) | Override proxy hostname |

### Helm Values (aegis-spoke)

```yaml
proxy:
  enabled: true
  service:
    type: NodePort
    port: 443
    nodePort: 31484
  tls:
    terminateAtIngress: false
    cert: ""  # Provided by Pulumi
    key: ""   # Provided by Pulumi
```

### Client Trust Bundle Setup

After cluster provisioning, users must add the spoke-proxy certificate to their local trust bundle:

```bash
# 1. Get the certificate from the EKS cluster
kubectl get secret aegis-spoke-proxy-tls -n aegis-system -o jsonpath='{.data.tls\.crt}' | base64 -d >> ~/aegis-local-trust.pem

# 2. Or export from the running proxy
openssl s_client -connect spoke-proxy.X.X.X.X.nip.io:31484 \
  -servername spoke-proxy.X.X.X.X.nip.io < /dev/null 2>/dev/null | \
  openssl x509 >> ~/aegis-local-trust.pem
```

The VS Code extension reads this file to trust the spoke-proxy's self-signed certificate.

---

## Current Limitations

### 1. Single Node IP Assumption

**Issue:** The system discovers ONE node's public IP and uses it for the proxy URL. If:
- The spoke-proxy pod moves to a different node
- The node is replaced (auto-scaling, spot instance termination)

...the proxy URL becomes invalid.

**Impact:** Connection failures after node changes.

**Workaround:** Manually update proxy_url in database or redeploy spoke with new IP.

### 2. Shared Wildcard Certificate

**Issue:** All clusters use the same *.nip.io wildcard certificate generated by Pulumi.

**Impact:**
- If one cluster's certificate key is compromised, all are affected
- No certificate isolation between tenants

**Workaround:** None currently. See Future Work section.

### 3. NodePort Collision Risk

**Issue:** Fixed NodePort 31484 for all clusters.

**Impact:** If you deploy multiple spokes on the SAME cluster (unusual), they'd conflict.

**Workaround:** Configure different port via `AEGIS_SPOKE_PROXY_NODEPORT`.

### 4. No Certificate Rotation

**Issue:** Certificates are valid for 1 year with no automatic rotation.

**Impact:** Connections will fail after 1 year without manual intervention.

**Workaround:** Manually regenerate and redeploy.

### 5. nip.io Dependency

**Issue:** Uses nip.io for dynamic DNS (IP-based hostnames).

**Impact:** If nip.io service is down, DNS resolution fails.

**Workaround:** Use fixed DNS with NLB (see Future Work).

### 6. Manual Trust Bundle Distribution

**Issue:** Users must manually add the certificate to `~/aegis-local-trust.pem`.

**Impact:** Extra setup step for each new cluster.

**Workaround:** Use Let's Encrypt with real domain (see Future Work).

---

## Troubleshooting

### Common Issues

#### 1. "cluster has no proxy URL configured"

**Cause:** k8s-agent hasn't registered the cluster yet, or heartbeat isn't reaching platform-api.

**Check:**
```bash
# Check k8s-agent logs on EKS
kubectl logs -n aegis-system deploy/aegis-spoke-k8s-agent

# Check database
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U aegis_platform -d aegis_platform -c \
  "SELECT id, proxy_url FROM clusters WHERE id = 'your-cluster-id';"
```

**Fix:**
```sql
-- Manually update if needed
UPDATE clusters SET proxy_url = 'wss://spoke-proxy.YOUR_IP.nip.io:31484'
WHERE id = 'your-cluster-id';
```

#### 2. Connection timeout to spoke-proxy

**Cause:** Security group doesn't allow NodePort traffic.

**Check:**
```bash
# Test connectivity
nc -zv spoke-proxy.X.X.X.X.nip.io 31484

# Check security group rules in AWS console
# Look for inbound TCP 31484 from 0.0.0.0/0
```

**Fix:**
```bash
aws ec2 authorize-security-group-ingress \
  --group-id sg-XXXXX \
  --protocol tcp --port 31484 --cidr 0.0.0.0/0
```

#### 3. TLS certificate error

**Cause:** Certificate not in trust bundle or doesn't match hostname.

**Check:**
```bash
# Verify cert SANs
openssl s_client -connect spoke-proxy.X.X.X.X.nip.io:31484 \
  -servername spoke-proxy.X.X.X.X.nip.io < /dev/null 2>/dev/null | \
  openssl x509 -noout -ext subjectAltName
```

**Fix:**
```bash
# Get cert from cluster and add to trust bundle
kubectl get secret aegis-spoke-proxy-tls -n aegis-system \
  -o jsonpath='{.data.tls\.crt}' | base64 -d >> ~/aegis-local-trust.pem
```

#### 4. k8s-agent can't discover node IP

**Cause:** AWS metadata service not accessible or external services blocked.

**Check:**
```bash
# From k8s-agent pod
curl -s http://169.254.169.254/latest/meta-data/public-ipv4
```

**Fix:** Use `AEGIS_PROXY_URL` environment variable to override auto-discovery.

---

## Future Work for Scalability & Multi-tenancy

### Priority 1: Stable Endpoint (High Priority)

**Current:** NodePort on node's public IP (unstable)

**Recommended:** Network Load Balancer (NLB) per cluster

```
┌─────────────────────────────────────────────────────────────────┐
│  CURRENT (Unstable)                                             │
│                                                                 │
│  Client ──► Node Public IP:31484 ──► spoke-proxy pod           │
│             (changes on node failure)                           │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  RECOMMENDED (Stable)                                           │
│                                                                 │
│  Client ──► NLB DNS ──► Target Group ──► spoke-proxy pod       │
│             (stable, auto-updates)                              │
└─────────────────────────────────────────────────────────────────┘
```

**Implementation:**
1. Create NLB in Pulumi during cluster provisioning
2. Use NLB DNS name instead of node IP
3. Eliminates nip.io dependency

### Priority 2: Per-Cluster Certificates (High Priority)

**Current:** Shared wildcard *.nip.io certificate

**Recommended:** Unique certificate per cluster

```go
// In Pulumi runner
func generateSpokeProxyCert(clusterID string) (certPEM, keyPEM string, err error) {
    template := x509.Certificate{
        Subject: pkix.Name{
            CommonName: fmt.Sprintf("spoke-proxy.%s.aegis.local", clusterID),
        },
        DNSNames: []string{
            fmt.Sprintf("spoke-proxy.%s.aegis.local", clusterID),
            "*.nip.io",  // Keep for compatibility
        },
    }
}
```

**Benefits:**
- Certificate isolation per tenant
- Compromised cert affects only one cluster
- Better audit trail

### Priority 3: Certificate Rotation (Medium Priority)

**Current:** 1-year certificates, no rotation

**Recommended:** cert-manager with automatic renewal

```yaml
# Use existing cert-manager template (already in aegis-platform)
proxy:
  tls:
    certManager:
      enabled: true
      issuerRef:
        name: aegis-internal
        kind: StepClusterIssuer
      duration: 2160h      # 90 days
      renewBefore: 168h    # Renew 7 days before expiry
```

**Implementation:**
1. Install step-ca and step-issuer (script exists: `scripts/install-internal-pki.sh`)
2. Enable `proxy.tls.certManager.enabled: true` in Pulumi
3. cert-manager auto-renews before expiry

### Priority 4: Let's Encrypt for Public Trust (Medium Priority)

**Current:** Self-signed certs require manual trust bundle setup

**Recommended:** Let's Encrypt with real domain

```
┌─────────────────────────────────────────────────────────────────┐
│  CURRENT                                                        │
│                                                                 │
│  1. Pulumi generates self-signed cert                           │
│  2. User manually exports cert to ~/aegis-local-trust.pem       │
│  3. VS Code extension trusts the cert                           │
│  4. Connect                                                     │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  WITH LET'S ENCRYPT                                             │
│                                                                 │
│  1. cert-manager requests Let's Encrypt cert                    │
│  2. Publicly trusted - no trust bundle needed                   │
│  3. Client connects directly                                    │
└─────────────────────────────────────────────────────────────────┘
```

**Requirements:**
- Real domain (e.g., `*.spoke.aegis-platform.com`)
- DNS challenge or HTTP challenge for Let's Encrypt
- cert-manager with ACME issuer

### Priority 5: DaemonSet for spoke-proxy (Low Priority)

**Current:** Single spoke-proxy pod, can move between nodes

**Recommended:** DaemonSet ensures proxy on every node

**Benefits:**
- Any node IP works for connection
- Better availability
- Load distribution

### Priority 6: Multi-Region HA (Future)

**Current:** Single cluster per region

**Future:** Multiple clusters with global load balancing

```
┌─────────────────────────────────────────────────────────────────┐
│  FUTURE: Global Load Balancing                                  │
│                                                                 │
│  Client ──► Route 53 (latency-based) ──┬──► us-east-1 NLB      │
│                                         ├──► us-west-2 NLB      │
│                                         └──► eu-west-1 NLB      │
└─────────────────────────────────────────────────────────────────┘
```

---

## Summary: What to Expect During Testing

### Working Automatically

1. Security group rule for NodePort 31484
2. spoke-proxy enabled with NodePort service
3. TLS certificate generated and deployed
4. k8s-agent discovers node IP and registers proxy_url

### Manual Steps Still Required

1. **Trust bundle on client machine:**
   ```bash
   kubectl get secret aegis-spoke-proxy-tls -n aegis-system \
     -o jsonpath='{.data.tls\.crt}' | base64 -d >> ~/aegis-local-trust.pem
   ```

2. **If node IP changes:** Update proxy_url in database or redeploy

3. **If security group rule missing:** Add manually via AWS console/CLI

### Potential Issues to Watch

| Issue | Symptom | Quick Fix |
|-------|---------|-----------|
| Node IP changed | Connection timeout | Update proxy_url in DB |
| Security group missing | Connection refused | Add NodePort rule |
| k8s-agent not running | No proxy_url in DB | Check pod logs |
| Cert not trusted | TLS handshake failure | Add cert to trust bundle |
| nip.io down | DNS resolution failure | Use IP directly (temp) |

---

## Files Modified

### Modified Files

| File | Changes |
|------|---------|
| `proto/aegis/v1/platform.proto` | proxy_url in ClusterRegisterRequest and ClusterHeartbeat |
| `services/platform-api/internal/store/cluster.go` | ProxyURL in ClusterInfo struct |
| `services/platform-api/internal/store/postgres/clusters.go` | proxy_url in all queries |
| `services/platform-api/internal/provisioning/pulumi/aws/runner.go` | Security group rule, TLS cert generation, proxy helm values |
| `agents/k8s-agent/internal/controller/aegisworkload_controller.go` | IP discovery, proxy URL construction |
| `agents/k8s-agent/internal/cpclient/client.go` | proxyURL in HeartbeatLoop |
| `charts/aegis-spoke/values.yaml` | nodePort configuration |
