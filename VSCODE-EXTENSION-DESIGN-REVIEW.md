# VS Code Extension Design Review & Assessment

## Overview

This document contains a comprehensive review of the proposed VS Code extension design for Aegis, which aims to eliminate SSH dependency and provide a native Remote window experience using VS Code's RemoteAuthorityResolver API.

---

## Assessment Summary

This is an **excellent, well-researched design** that makes the right architectural call for eliminating SSH friction while leveraging existing Aegis infrastructure.

**Verdict:** **Proceed with this design** - it's 95% correct for the Aegis environment.

---

## ✅ What's Right About This Design

### 1. Correct Approach: RemoteAuthorityResolver vs SSH

**Why this is the right choice:**
- Using VS Code's **proposed RemoteAuthorityResolver API** eliminates SSH dependency
- Provides a **true Remote window** with native file system, terminal, debugging, etc. - not a hacked-together VFS
- Matches how GitHub Codespaces, Gitpod, and other modern remote development tools work
- Uses the **real REH (Remote Extension Host) protocol** that VS Code already implements

**The API choice is critical:**
- **RemoteAuthorityResolver** (Proposed API) - **Recommended** ✅
  - Opens native Remote window
  - Full VS Code feature support
  - Standard approach used by Codespaces/Gitpod
  - Purpose-built for this use case

- **Custom FileSystemProvider + Pseudoterminal** (Fallback) - Not recommended ❌
  - Doesn't open true Remote window
  - Must emulate file operations and terminal
  - Many edge cases to handle
  - Limited VS Code features

### 2. Clean Integration Path with Existing Infrastructure

The design **correctly identifies** how to swap in production proxy infrastructure:

**Current (MVP):**
```
VS Code Extension → WSS (wss://127.0.0.1:7001) → Node Proxy → REH (TCP :11111)
```

**Production swap:**
```
VS Code Extension → WSS (wss://your-nlb.aws.com:8080/proxy/wl-123)
                  → aegis-auth-proxy (Go)
                  → Workspace Pod (REH on :11111)
```

**Existing infrastructure already supports this:**

Your **existing `aegis-auth-proxy`** already provides:
- ✅ WebSocket → TCP bridging
- ✅ JWT validation (Bearer token in header)
- ✅ JTI single-use tracking
- ✅ mTLS client cert verification
- ✅ Inactivity timeouts
- ✅ Audit logging

**This is huge** - you're not rebuilding infrastructure; you're just adding a new client (VS Code extension) to talk to your existing proxy.

### 3. Phased Rollout is Smart

The three-phase approach is pragmatic:

**Phase 0 — Skeleton & Harness**
- Prove byte streaming works with echo server
- Validate WebSocket → TCP forwarding
- Test ping/pong and connection lifecycle

**Phase 1 — Remote Editing & Terminal (MVP)**
- Replace echo server with actual REH
- Open VS Code Remote window
- Enable file editing and integrated terminal

**Phase 2 — Stability & DevX**
- Auto-reconnect with exponential backoff
- Better logs/metrics (latency, rx/tx, reconnect count)
- Config UI and clean shutdown

This approach gets value early and iterates fast.

---

## ⚠️ Gaps vs Aegis Cloud Environment

While the design is fundamentally sound, there are implementation details specific to the Aegis environment that need to be addressed:

### 1. JWT Token Management (Critical for Production)

**What's missing:**
The design mentions JWT but doesn't detail **how the extension gets tokens from platform-api**.

**Current Aegis flow:**
```typescript
// Extension needs to call CreateConnectionSession before connecting
const session = await platformAPI.createConnectionSession({
  workloadId: 'wl-123',
  client: 'vscode'
});

// session contains:
{
  proxyUrl: "wss://a9604fc12...elb.amazonaws.com:8080/proxy/wl-123",
  token: "eyJhbGciOiJIUzI1NiIs...",  // JWT (5min expiry, one-time use)
  sshConfig: "...",                    // for SSH fallback
  vscodeUri: "vscode-remote://aegis+wl-123/home/project"
}
```

**Code change needed in `resolver.ts`:**

```typescript
// BEFORE makeManagedConnection, add token fetch:
async function getConnectionSession(wid: string): Promise<ConnectionSession> {
  // Call platform-api via gRPC or REST
  const response = await fetch('http://localhost:8081/api/sessions', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${await getAccessToken()}`,
    },
    body: JSON.stringify({ workload_id: wid, client: 'vscode' })
  });
  return response.json();
}

// In resolver.resolve():
const session = await getConnectionSession(wid);

// Pass JWT in WebSocket connection:
const ws = new WebSocket(session.proxyUrl, {
  headers: {
    'Authorization': `Bearer ${session.token}`,
    'X-Aegis-Workload': wid
  },
  rejectUnauthorized: true // production should verify certs
});
```

**Token refresh challenge:**
- JWT tokens are **one-time use** with **5-minute expiry**
- VS Code may disconnect/reconnect (user switches windows, network hiccup)
- Each reconnect needs a **fresh token**

**Solution options:**

**Option A: Request new token on each reconnect**
```typescript
// VS Code calls resolve() again on reconnect
// Extension fetches fresh token from platform-api
```
- ✅ Most secure (new token each time)
- ⚠️ Requires platform-api call on every reconnect
- ⚠️ User sees "Connecting..." briefly

**Option B: Use refresh tokens**
```typescript
// Initial session includes refresh token (8hr expiry)
{
  token: "eyJ...",           // access token (5min)
  refreshToken: "eyJ...",    // refresh token (8hr)
}

// Extension refreshes access token in background
async function refreshAccessToken(refreshToken: string): Promise<string> {
  const response = await platformAPI.refreshSession({ refreshToken });
  return response.token;
}
```
- ✅ Seamless reconnects (no user-visible delay)
- ✅ Fewer platform-api calls
- ⚠️ More complex token management
- ⚠️ Slightly less secure (longer-lived refresh token)

**Recommendation:** Start with **Option A** (simpler), add **Option B** in Phase 2 if reconnects are too slow.

### 2. Workspace REH Bootstrap (Critical for Production)

**The gap:**
The mock workspace uses **VSCodium REH**. Aegis **production workspaces** need to run REH on startup.

**Current Aegis workspace setup:**
- SSH server runs on port 2222 (already working)
- Need to **add** REH server on port 11111

**Two approaches:**

#### **Approach A: Extend existing workspace images**

Add REH to workspace images (e.g., `carlosmsanchez/aegis-workspace-vscode:latest`):

```dockerfile
# Workspace Dockerfile
FROM ubuntu:22.04

# Install SSH (existing)
RUN apt-get update && apt-get install -y openssh-server

# Add REH (new)
ARG VSCODE_COMMIT=0f0d87fa9e96c856c5212fc86db137ac0d783365
RUN curl -sSL "https://update.code.visualstudio.com/commit:${VSCODE_COMMIT}/server-linux-x64/stable" \
  | tar -xz -C /opt/vscode-server

# Entrypoint runs both SSH and REH
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
```

**entrypoint.sh:**
```bash
#!/bin/bash
# Start SSH server
/usr/sbin/sshd -D -p 2222 &

# Start REH server
/opt/vscode-server/bin/code-server \
  --host 0.0.0.0 \
  --port 11111 \
  --connection-token "${AEGIS_REH_TOKEN:-hello}" \
  --accept-server-license-terms \
  --without-connection-secret \
  &

# Wait for both processes
wait
```

**k8s-agent changes (workspace.go):**
```go
// Add REH port to workspace pod spec
spec.Template.Spec.Containers[0].Ports = append(
  spec.Template.Spec.Containers[0].Ports,
  corev1.ContainerPort{ContainerPort: 11111, Protocol: corev1.ProtocolTCP},
)

// Expose REH port in Service
spec.Spec.Ports = append(spec.Spec.Ports, corev1.ServicePort{
  Name:       "reh",
  Port:       11111,
  TargetPort: intstr.FromInt(11111),
  Protocol:   corev1.ProtocolTCP,
})
```

#### **Approach B: REH as sidecar container**

Keep workspace images unchanged, add REH as a sidecar:

```go
// In k8s-agent workspace.go BuildWorkspaceJob()
spec.Template.Spec.Containers = append(spec.Template.Spec.Containers, corev1.Container{
  Name:  "vscode-server",
  Image: "ghcr.io/coder/code-server:latest", // or custom image
  Ports: []corev1.ContainerPort{
    {ContainerPort: 11111, Protocol: corev1.ProtocolTCP},
  },
  Env: []corev1.EnvVar{
    {Name: "PASSWORD", Value: os.Getenv("AEGIS_REH_TOKEN")},
  },
  VolumeMounts: []corev1.VolumeMount{
    {Name: "workspace", MountPath: "/home/coder/project"},
  },
})
```

**Recommendation:** Use **Approach A** (extend images). It's cleaner and gives you full control over REH configuration.

### 3. REH Connection Token Handling

VS Code REH uses a **connection token** for authentication:
```bash
code-server --connection-token-file /path/to/token
```

The extension must send this token in the **first handshake message** after WebSocket connects.

**The challenge:**
- Aegis **aegis-auth-proxy** validates JWT tokens
- REH **also** expects a connection token
- These are **two separate authentication layers**

**VS Code REH handshake flow:**
```
1. WebSocket opens (JWT validated by aegis-auth-proxy)
2. Client sends first message: {"type":"auth","token":"<REH_CONNECTION_TOKEN>"}
3. REH validates connection token
4. Byte stream starts (REH protocol)
```

**Solution options:**

#### **Option A: Pass REH token in JWT claims**

```go
// platform-api CreateConnectionSession
rehToken := generateRandomToken(32) // random 32-char string

claims := jwt.MapClaims{
  "sub":       user,
  "wid":       workloadId,
  "reh_token": rehToken,  // NEW: embed REH token in JWT
  "dest":      serviceHost,
  "one_time":  true,
  "exp":       time.Now().Add(5 * time.Minute).Unix(),
}

tokenString, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)

// Store REH token in database for workspace
db.Exec("UPDATE workloads SET reh_token = ? WHERE id = ?", rehToken, workloadId)

return &pb.ConnectionSession{
  Token:     tokenString,
  ProxyUrl:  proxyUrl,
  RehToken:  rehToken, // also return separately for extension
}
```

**aegis-auth-proxy extracts REH token:**
```go
// In proxy CONNECT handler after JWT validation
claims := token.Claims.(jwt.MapClaims)
rehToken := claims["reh_token"].(string)

// After establishing TCP connection to REH, inject auth message
authMsg := fmt.Sprintf(`{"type":"auth","token":"%s"}`, rehToken)
targetConn.Write([]byte(authMsg + "\n"))

// Then start byte pump (bidirectional forwarding)
```

#### **Option B: Disable REH token (dev only)**

For **local development only**:
```bash
code-server --without-connection-token  # NOT for production
```

**Recommendation:** Use **Option A** for production (more secure). Use **Option B** for MVP/local testing to move faster.

### 4. Workspace Discovery (TreeView Needs Real Data)

The mock TreeView hardcodes `w-1234`. Production needs dynamic workspace list.

**Current design:**
```typescript
export class WorkspacesProvider {
  getChildren() {
    return [new vscode.TreeItem('w-1234')]; // hardcoded
  }
}
```

**Production version:**
```typescript
export class WorkspacesProvider {
  private platformAPI: AegisPlatformClient;

  async getChildren(): Promise<vscode.TreeItem[]> {
    try {
      // Call platform-api ListWorkloads
      const workspaces = await this.platformAPI.listWorkloads({
        projectId: 'p-demo',
        filter: { interactive: true } // only show interactive workspaces
      });

      return workspaces.map(w => {
        const item = new vscode.TreeItem(
          w.id,
          vscode.TreeItemCollapsibleState.None
        );

        // Set icon based on status
        item.iconPath = new vscode.ThemeIcon(
          w.status === 'running' ? 'vm-running' : 'vm-outline'
        );

        // Show GPU info in description
        item.description = w.hints?.gpuCount > 0
          ? `${w.hints.gpuCount}x GPU`
          : 'CPU';

        // Tooltip with details
        item.tooltip = `Status: ${w.status}\nFlavor: ${w.workspace.flavor}`;

        // Click to connect
        item.command = {
          command: 'aegis.connect',
          title: 'Connect',
          arguments: [w.id]
        };

        return item;
      });
    } catch (error) {
      out.appendLine(`[workspaces] Error listing: ${error}`);
      return [];
    }
  }

  // Refresh every 10 seconds to show status changes
  startAutoRefresh() {
    setInterval(() => this.refresh(), 10000);
  }
}
```

**Platform-API call:**
```typescript
// extension/src/api-client.ts
import * as grpc from '@grpc/grpc-js';
import { AegisPlatformClient } from './generated/aegis_grpc_pb';

export class PlatformAPIClient {
  private client: AegisPlatformClient;

  constructor(endpoint: string = 'localhost:8081') {
    this.client = new AegisPlatformClient(
      endpoint,
      grpc.credentials.createInsecure() // use mTLS in production
    );
  }

  async listWorkloads(req: ListWorkloadsRequest): Promise<Workload[]> {
    return new Promise((resolve, reject) => {
      this.client.listWorkloads(req, (err, response) => {
        if (err) reject(err);
        else resolve(response.workloadsList);
      });
    });
  }

  async createConnectionSession(req: CreateConnectionSessionRequest): Promise<ConnectionSession> {
    return new Promise((resolve, reject) => {
      this.client.createConnectionSession(req, (err, response) => {
        if (err) reject(err);
        else resolve(response);
      });
    });
  }
}
```

### 5. Multi-Cluster Support

Aegis has **aegis-spoke** clusters in different regions/environments. The extension needs to connect to the correct proxy URL.

**Current architecture:**
```
aegis-spoke-prod (us-east-1)  → NLB: a9604fc1...elb.amazonaws.com:8080
aegis-spoke-dev  (us-west-2)  → NLB: b7805gd2...elb.amazonaws.com:8080
aegis-spoke-eu   (eu-west-1)  → NLB: c8906he3...elb.amazonaws.com:8080
```

**Good news:** Platform-API already handles this in `CreateConnectionSession`:

```go
// platform-api already returns cluster-specific proxy URL
func (s *Server) CreateConnectionSession(ctx context.Context, req *pb.CreateConnectionSessionRequest) (*pb.ConnectionSession, error) {
  workload := s.getWorkload(req.WorkloadId)

  // Get proxy URL for workload's cluster
  proxyHost := s.getProxyURLForCluster(workload.ClusterID) // e.g., "a9604fc1...elb.amazonaws.com:8080"

  return &pb.ConnectionSession{
    ProxyUrl: fmt.Sprintf("wss://%s/proxy/%s", proxyHost, req.WorkloadId),
    Token:    jwt,
  }, nil
}
```

**Extension just uses the returned URL:**
```typescript
// resolver.ts
const session = await getConnectionSession(wid);
const ws = new WebSocket(session.proxyUrl, { // uses correct cluster URL
  headers: { 'Authorization': `Bearer ${session.token}` }
});
```

**No changes needed** if platform-api returns the correct proxy URL for each cluster.

---

## 📊 Comparison: This Approach vs Current SSH Approach

| Aspect | SSH + aegis-connect (current) | VS Code Extension (proposed) |
|--------|------------------------------|------------------------------|
| **Initial setup** | Manual (8 steps) | 1-click in VS Code |
| **User friction** | Medium (CLI commands) | Low (GUI buttons) |
| **Token refresh** | Manual (`aegis-refresh-ssh`) | Automatic in extension |
| **SSH config** | Pollutes `~/.ssh/config` | No SSH config needed |
| **Dependencies** | SSH client + aegis-connect binary | Just VS Code Insiders |
| **Terminal** | Native SSH terminal | VS Code integrated terminal |
| **File editing** | VS Code Remote-SSH | Native Remote window |
| **Debugging** | Full support | Full support |
| **Port forwarding** | SSH port forwarding | VS Code port forwarding |
| **Works on Windows** | Yes (with OpenSSH) | Yes (native) |
| **Works on iPad/Chromebook** | No | **Future:** Yes (web-based) |
| **Multi-user support** | Each user needs setup | Extension installed once |
| **Backstage integration** | Copy SSH config manually | Deep links / API integration |

**UX Improvement Summary:**
- **Setup time:** 10 minutes → 30 seconds
- **Commands per connection:** 2 (refresh + connect) → 1 (click)
- **Config files touched:** 2 (`~/.ssh/config` + install binary) → 0

**Verdict:** The VS Code extension approach is **significantly better UX** and **architecturally cleaner**.

---

## 🚀 Recommended Implementation Plan

### Timeline: 5-8 weeks to production-ready extension

#### **Week 1-2: Phase 0-1 (Local Harness)**
**Goal:** Get local MVP working end-to-end

Tasks:
- [ ] Scaffold extension with Proposed API setup
- [ ] Build Node/TS WSS→TCP proxy
- [ ] Create mock workspace Docker image with REH
- [ ] Test echo server (Phase 0)
- [ ] Test REH connection with file editing + terminal (Phase 1)

**Success criteria:**
- ✅ Remote window opens in VS Code
- ✅ Can edit/save files
- ✅ Integrated terminal works

#### **Week 3-4: Production Integration**
**Goal:** Connect extension to real Aegis infrastructure

Tasks:
- [ ] Add token provider that calls `CreateConnectionSession`
- [ ] Use real `proxyUrl` from platform-api (not hardcoded localhost)
- [ ] Add REH to workspace images (Dockerfile changes)
- [ ] Update k8s-agent to expose port 11111 in Services
- [ ] Handle REH connection token (embed in JWT claims)
- [ ] Test with real workspace in EKS cluster

**Success criteria:**
- ✅ Extension connects to cloud workspaces (not just localhost)
- ✅ JWT auth works through aegis-auth-proxy
- ✅ REH connection token validated

#### **Week 5-6: Polish & Features**
**Goal:** Make it production-ready

Tasks:
- [ ] Dynamic workspace list (TreeView calls `ListWorkloads`)
- [ ] Status indicators (running/stopped/starting)
- [ ] Error handling & user-friendly messages
- [ ] Auto-reconnect with exponential backoff (Phase 2)
- [ ] Logs/metrics (bytes tx/rx, latency, reconnect count)
- [ ] Extension settings UI (proxy URL, timeout, etc.)

**Success criteria:**
- ✅ TreeView shows all user's workspaces
- ✅ Reconnect works after network hiccup
- ✅ Clear error messages for common failures

#### **Week 7-8: Beta Testing**
**Goal:** Internal team validation

Tasks:
- [ ] Package extension as `.vsix`
- [ ] Internal rollout (5-10 users)
- [ ] Collect feedback on bugs/UX issues
- [ ] Fix critical issues
- [ ] Write user documentation
- [ ] Create demo video

**Success criteria:**
- ✅ 5+ internal users successfully connect to workspaces
- ✅ No critical bugs reported
- ✅ Documentation complete

---

## 🔒 Security Considerations

The proposed design is **security-conscious** and aligns with Aegis's existing security model:

### ✅ Existing Security (Preserved)
- **JWT authentication** (existing, carried forward)
- **JTI single-use tokens** (existing, prevents replay attacks)
- **TLS encryption** (WSS protocol)
- **mTLS client certificates** (existing proxy supports this)
- **5-minute token expiry** (existing)
- **Audit logging** (existing in aegis-auth-proxy)

### 🆕 New Security Considerations

#### **1. REH Connection Token**
- REH has its own connection token (separate from JWT)
- **Recommendation:** Generate random 32-char token, embed in JWT claims
- Store in workspace metadata for validation

#### **2. Token Refresh Pattern**
Current: One-time JWT tokens (very secure but requires manual refresh)

**Enhancement options:**

**Option A: Keep one-time tokens (most secure)**
- User must explicitly reconnect after token used
- Extension shows "Token expired - Click to reconnect"
- ✅ Maximum security
- ⚠️ More friction on reconnects

**Option B: Add refresh tokens (balanced)**
```go
type ConnectionSession {
  AccessToken  string // 5min expiry, one-time use
  RefreshToken string // 8hr expiry, multi-use (rate-limited)
}
```
- Extension refreshes access token in background
- User doesn't see reconnect prompts
- ✅ Better UX
- ⚠️ Slightly less secure (longer-lived refresh token)

**Option C: Long-lived session tokens (least secure)**
```go
type ConnectionSession {
  SessionToken string // 8hr expiry, multi-use
}
```
- ✅ Simplest implementation
- ❌ No replay attack protection
- ❌ Not recommended for production

**Recommendation:** Start with **Option A** (MVP), add **Option B** if reconnect friction is too high in beta testing.

#### **3. Extension Distribution Security**
When publishing to VS Code Marketplace:
- **Sign extension** with code signing certificate
- **Verify package integrity** (checksums)
- **Review process** (VS Code team reviews submissions)
- **Update mechanism** (auto-updates for security patches)

#### **4. Secrets Management in Extension**
- **Never hardcode** JWT secrets or API keys
- **Use VS Code SecretStorage** for user tokens:
```typescript
// Store user auth token securely
await context.secrets.store('aegis.userToken', token);

// Retrieve when needed
const token = await context.secrets.get('aegis.userToken');
```

---

## 🎯 Integration with Existing Aegis Components

### **What needs to change in Aegis codebase:**

#### **1. Platform-API (Minor changes)**
```go
// services/platform-api/internal/server/server.go

// ADD: Return REH token in CreateConnectionSession
func (s *Server) CreateConnectionSession(ctx context.Context, req *pb.CreateConnectionSessionRequest) (*pb.ConnectionSession, error) {
  // ... existing JWT creation code ...

  // NEW: Generate REH connection token
  rehToken := generateRandomToken(32)

  // NEW: Embed in JWT claims
  claims["reh_token"] = rehToken

  // NEW: Store for later validation
  s.db.Exec("UPDATE workloads SET reh_token = ? WHERE id = ?", rehToken, workloadId)

  return &pb.ConnectionSession{
    Token:     tokenString,
    ProxyUrl:  proxyUrl,
    RehToken:  rehToken, // NEW field
  }, nil
}
```

#### **2. K8s-Agent (Workspace builder changes)**
```go
// agents/k8s-agent/internal/workload/builders/workspace.go

// ADD: REH port to container spec
func BuildWorkspaceJob(opts WorkspaceOptions) *batchv1.Job {
  container := corev1.Container{
    Name:  "workspace",
    Image: opts.Image,
    Ports: []corev1.ContainerPort{
      {ContainerPort: 2222, Protocol: corev1.ProtocolTCP},  // SSH (existing)
      {ContainerPort: 11111, Protocol: corev1.ProtocolTCP}, // REH (NEW)
    },
    Env: []corev1.EnvVar{
      {Name: "AEGIS_REH_TOKEN", Value: opts.REHToken}, // NEW
    },
  }

  // ADD: REH port to Service
  service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
    Name:       "reh",
    Port:       11111,
    TargetPort: intstr.FromInt(11111),
  })

  return job
}
```

#### **3. Aegis-Auth-Proxy (No changes needed!)**
The existing proxy already supports:
- ✅ WebSocket connections
- ✅ JWT validation
- ✅ TCP forwarding

**Only if using REH token in JWT claims:**
```go
// services/proxy/internal/handlers/connect.go

// OPTIONAL: Extract REH token from JWT and inject in first message
func (h *ConnectHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
  claims := r.Context().Value("claims").(jwt.MapClaims)
  rehToken, _ := claims["reh_token"].(string)

  // After establishing TCP connection to workspace:11111
  if rehToken != "" {
    authMsg := fmt.Sprintf(`{"type":"auth","token":"%s"}`, rehToken)
    targetConn.Write([]byte(authMsg + "\n"))
  }

  // Continue with existing byte pump logic...
}
```

#### **4. Workspace Images (Dockerfile updates)**
```dockerfile
# Base workspace image changes
FROM ubuntu:22.04

# Existing SSH setup
RUN apt-get update && apt-get install -y openssh-server

# NEW: Add VS Code Remote Extension Host
ARG VSCODE_COMMIT=latest
RUN curl -sSL "https://update.code.visualstudio.com/commit:${VSCODE_COMMIT}/server-linux-x64/stable" \
  | tar -xz -C /opt/vscode-server

# NEW: Entrypoint runs both SSH and REH
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

# entrypoint.sh:
# - Start sshd on :2222
# - Start code-server on :11111
# - Use $AEGIS_REH_TOKEN if provided
```

---

## Summary: Design Validation

### ✅ Strengths
1. **Architecturally sound** - Uses VS Code's official Proposed API for remote connections
2. **Leverages existing infrastructure** - Minimal changes to aegis-auth-proxy, platform-api
3. **Phased approach** - MVP first, polish later
4. **Security-conscious** - Maintains JWT auth, adds REH token layer
5. **Better UX** - Eliminates SSH config management, token refresh scripts

### ⚠️ Implementation Details to Address
1. **Token provider** - Extension needs to call `CreateConnectionSession`
2. **REH bootstrap** - Add code-server to workspace images
3. **REH connection token** - Embed in JWT claims or disable for MVP
4. **Workspace list** - TreeView calls `ListWorkloads` API
5. **Multi-cluster** - Use proxy URL from platform-api (already supported)

### 📈 Expected Outcomes
- **Setup time:** 10 minutes → 30 seconds
- **User friction:** High (manual CLI) → Low (1-click GUI)
- **Maintenance:** Custom scripts → Standard VS Code extension
- **Team adoption:** Slow (learning curve) → Fast (familiar IDE)

### 🚦 Go/No-Go Decision
**Recommendation: GO** ✅

This design is 95% correct for Aegis. The core architecture is sound, and the gaps are straightforward implementation details. Proceed with Phase 0-1 implementation.

---

## References

### VS Code Proposed APIs
- [RemoteAuthorityResolver API](https://raw.githubusercontent.com/microsoft/vscode/main/src/vscode-dts/vscode.proposed.resolvers.d.ts)
- [Using Proposed APIs (requires VS Code Insiders)](https://code.visualstudio.com/api/advanced-topics/using-proposed-api)
- [Tunnels API](https://raw.githubusercontent.com/microsoft/vscode/main/src/vscode-dts/vscode.proposed.tunnels.d.ts)

### Similar Implementations
- [GitHub Codespaces](https://github.com/features/codespaces) - Uses RemoteAuthorityResolver
- [Gitpod](https://www.gitpod.io/) - Similar WebSocket tunnel approach
- [Remote OSS Extension](https://marketplace.visualstudio.com/items?itemName=codeforx.remote-oss) - Community example

### Aegis Documentation
- [VSCODE-SSH-SETUP.md](/Users/carlossanchez/code/aegis/VSCODE-SSH-SETUP.md) - Current SSH approach
- [MVP-UX-ANALYSIS.md](/Users/carlossanchez/code/aegis/MVP-UX-ANALYSIS.md) - UX improvement roadmap
- [CLOUD-PROXY-SETUP.md](/Users/carlossanchez/code/aegis/CLOUD-PROXY-SETUP.md) - Proxy infrastructure details
