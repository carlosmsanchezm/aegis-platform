# Aegis MVP UX Analysis: Why Local Setup is Required & Path to One-Click Experience

## What We Validated Today

✅ **Core functionality works end-to-end:**
- Users can SSH into remote workspaces (CPU or GPU) in the cloud
- Connection is secure (JWT authentication + TLS)
- VS Code Remote-SSH works perfectly once configured
- Same flow works for GPU workspaces - just change the workload ID

## Current User Experience (Manual Setup Required)

### What Users Must Do Locally

1. **Install aegis-connect binary** (`~/.local/bin/aegis-connect`)
2. **Generate SSH key pair** (`~/.ssh/aegis_workspace_key`)
3. **Add SSH key to workspace** (one-time per workspace)
4. **Before each connection:**
   - Run `aegis-refresh-ssh` to get fresh JWT token
   - Update `~/.ssh/config` with new token
   - Connect via VS Code Remote-SSH

### Why This Setup is Required

#### 1. **aegis-connect Must Be Local (SSH ProxyCommand Requirement)**

```
VS Code → SSH Client → ProxyCommand: aegis-connect (LOCAL) → HTTPS → Cloud Proxy → Workspace
```

**Technical Reason:**
- SSH's `ProxyCommand` directive runs a **local** program to establish the connection
- This is how SSH's architecture works - it's not Aegis-specific
- VS Code's Remote-SSH extension uses native SSH, which reads `~/.ssh/config`

**Analogy:**
- **Direct SSH:** You connect directly to a server
- **SSH Jump Host:** SSH tunnels through another SSH server (still requires local SSH config)
- **Aegis:** SSH tunnels through HTTPS proxy with JWT auth (requires local helper binary)

#### 2. **SSH Config Must Be Local**

- `~/.ssh/config` is read by the **local** SSH client
- VS Code Remote-SSH uses the native SSH client on your Mac
- No way around this - it's how SSH works

#### 3. **JWT Tokens Are One-Time Use**

**Why one-time tokens:**
- Security: prevents token replay attacks
- Each connection gets a fresh, short-lived token
- Tokens expire in 5 minutes

**Impact:**
- User must refresh token before each VS Code connection
- Our `aegis-refresh-ssh` helper makes this 1 command instead of 3

## Current Friction Points

### 🔴 High Friction
1. **Initial setup** - Multiple manual steps (install binary, generate keys, add keys)
2. **Token refresh** - Must run command before each connection
3. **Port-forward required** - `kubectl port-forward` to platform-api must be running

### 🟡 Medium Friction
4. **SSH config pollution** - `~/.ssh/config` gets entries for each workspace
5. **Error messages unclear** - "Connection timeout" could mean token expired, proxy down, etc.

### 🟢 Low Friction (acceptable)
6. **aegis-connect binary** - Must be installed locally, but this is one-time
7. **SSH keys** - Standard practice, familiar to developers

## Why This Matters

**Current state: ~8 manual steps before first connection**
**Goal: 1-click connection**

Gap: We're far from the Jupyter/Colab experience where you just "click to connect"

## Proposed MVP UX Enhancements

### Phase 1: Reduce Setup (Quick Wins)

#### 1.1 Aegis Desktop App (Electron/Tauri)
**What it does:**
- Bundles aegis-connect binary (no manual install)
- Auto-manages `~/.ssh/config` entries
- Shows workspace list with "Connect" buttons
- Auto-refreshes JWT tokens in background
- One-time auth with Aegis platform

**User flow:**
```
1. Download Aegis Desktop App
2. Login with Aegis account
3. Click "Connect" next to workspace
4. VS Code opens automatically
```

**Pros:**
- 1-click after initial app install
- Familiar pattern (like Docker Desktop, Cursor IDE)
- Can handle SSH config management transparently
- Can bundle aegis-connect binary for all platforms

**Cons:**
- Requires building/maintaining desktop app
- Still depends on VS Code + Remote-SSH extension

**Effort:** Medium (2-3 weeks for MVP)

---

#### 1.2 Browser Extension + Native Messaging
**What it does:**
- Chrome/Firefox extension talks to local helper via native messaging
- Handles token refresh + SSH config updates
- "Connect" button in web UI triggers local setup

**User flow:**
```
1. Install browser extension (one-time)
2. Click "Connect to VS Code" in Aegis web UI
3. Extension handles token refresh + launches VS Code
```

**Pros:**
- Lightweight (no full desktop app)
- Integrates with existing web UI
- Standard pattern (like 1Password, Bitwarden)

**Cons:**
- Still requires native helper binary
- Browser extension complexity
- Native messaging can be finicky

**Effort:** Low-Medium (1-2 weeks)

---

#### 1.3 CLI Wrapper (Fastest to Ship)
**What we have:**
- `aegis-refresh-ssh` helper script (✅ already built today!)

**Enhance it:**
```bash
# Instead of:
aegis-refresh-ssh
code --remote ssh-remote+aegis-w-wl-ssh-workspace

# Make it:
aegis connect wl-ssh-workspace
# → Auto-refreshes token, updates config, launches VS Code
```

**User flow:**
```
1. Install aegis CLI: brew install aegis-cli
2. Login once: aegis login
3. Connect: aegis connect <workspace-id>
```

**Pros:**
- Ships in days (extend what we built today)
- Familiar to developer audience
- Zero new UI to build

**Cons:**
- Still requires CLI install
- Not as friendly as GUI

**Effort:** Very Low (2-3 days)

---

### Phase 2: Eliminate SSH Dependency (Bigger Lift)

#### 2.1 VS Code Extension (Direct TCP Tunnel)
**What it does:**
- Native VS Code extension that connects directly to Aegis proxy
- Bypasses SSH entirely - uses VS Code's remote protocol over HTTPS
- Extension handles JWT token refresh automatically

**User flow:**
```
1. Install "Aegis" extension in VS Code
2. Click "Connect to Aegis Workspace" in sidebar
3. Select workspace from list
4. Connected (no SSH, no config files)
```

**How it works:**
```
VS Code → Aegis Extension → HTTPS (with JWT) → aegis-auth-proxy → VS Code Server
```

**Pros:**
- No SSH config files
- No local binary (extension bundles connection logic)
- True 1-click experience
- Auto token refresh in background
- Extension can show workspace list, logs, etc.

**Cons:**
- Must build VS Code extension
- Must implement VS Code remote protocol
- Competes with official Remote-SSH extension

**Effort:** High (4-6 weeks)

**Reference:** GitHub Codespaces, Gitpod use this approach

---

#### 2.2 Web-Based IDE (JupyterLab/Code-Server)
**What it does:**
- Run VS Code in browser (code-server) or JupyterLab
- Direct HTTPS connection, no SSH needed
- Works on tablets, Chromebooks, etc.

**User flow:**
```
1. Click workspace in web UI
2. IDE opens in browser tab
3. Start coding
```

**Pros:**
- Zero local setup
- Works on any device with browser
- Familiar (like Google Colab, Repl.it)

**Cons:**
- Not native VS Code (code-server has limitations)
- Some extensions don't work in browser
- Network latency more noticeable

**Effort:** Medium (3-4 weeks to integrate)

---

### Phase 3: Eliminate Token Refresh

#### 3.1 Long-Lived Session Tokens
**Current:** One-time JWT tokens (5min expiry)
**Proposed:** Session tokens that allow multiple connections

**Changes needed:**
```go
// Instead of one_time: true
type ConnectionSession struct {
    Token       string
    ExpiresAt   time.Time // 8 hours
    MaxUses     int       // e.g., 10 connections
    UsedCount   int
}
```

**Pros:**
- Users don't need to refresh token between connections
- Still secure (time-limited + use-limited)

**Cons:**
- Slightly less secure than one-time tokens
- Need to track usage in database

**Effort:** Low (1-2 days)

---

#### 3.2 Persistent Connection with Reconnect
**What it does:**
- Keep TCP tunnel open in background
- VS Code reconnects through same tunnel
- Token refreshes automatically before expiry

**Implementation:**
- Long-running local daemon keeps tunnel alive
- Refreshes JWT token every 4 minutes
- VS Code connects to local tunnel (localhost:XXXX)

**Pros:**
- Zero manual token refresh
- Fast reconnections

**Cons:**
- Background daemon required
- Must handle daemon lifecycle

**Effort:** Medium (2-3 weeks)

---

## Recommended Path Forward

### Ship in 2 Weeks:
1. ✅ **CLI wrapper** (`aegis connect <workspace-id>`) - Extends what we built today
2. ✅ **Long-lived session tokens** - Simple backend change
3. ✅ **Installer script** - One-line install: `curl -sSL aegis.io/install.sh | sh`

**Result:** User flow becomes:
```bash
# One-time setup
curl -sSL https://aegis.io/install.sh | sh
aegis login

# Every connection (one command)
aegis connect my-gpu-workspace
```

---

### Ship in 1-2 Months:
4. **Aegis Desktop App** - Electron app with workspace list + 1-click connect
5. **Auto SSH key injection** - When workspace starts, inject user's public key automatically

**Result:** User flow becomes:
```
1. Download Aegis app
2. Login
3. Click "Connect" button
4. VS Code opens
```

---

### Ship in 3-6 Months:
6. **VS Code Extension** - Native extension that bypasses SSH entirely
7. **Web IDE option** - For users who prefer browser-based experience

**Result:** True 1-click experience like GitHub Codespaces

---

## Why Local Setup Will Always Exist (For Power Users)

Even with best UX, some users will prefer manual SSH setup:
- ✅ Works with any SSH client (not just VS Code)
- ✅ Can customize SSH config extensively
- ✅ Works with SSH tunneling, port forwarding, etc.
- ✅ Familiar to DevOps/SRE users

**Solution:** Offer both paths:
- **Simple path:** Desktop app / VS Code extension (1-click)
- **Power user path:** Manual SSH setup (full control)

---

## Comparison to Competitors

| Feature | Aegis (current) | GitHub Codespaces | Gitpod | Modal Labs |
|---------|----------------|-------------------|---------|------------|
| Initial setup | Manual (8 steps) | 1-click | 1-click | CLI (2 steps) |
| Connection | SSH + helper | Browser or VS Code ext | Browser or VS Code ext | SSH + CLI |
| Token refresh | Manual | Automatic | Automatic | Automatic |
| GPU support | ✅ | ❌ | ❌ | ✅ |
| Self-hosted | ✅ | ❌ | ✅ | ❌ |

**Our competitive advantage:** GPU support + self-hosted
**Our UX gap:** Manual setup vs. 1-click

---

## Technical Constraints (Why We Can't Eliminate Everything)

### Why aegis-connect must be local:
- SSH's ProxyCommand architecture requires it
- VS Code Remote-SSH uses native SSH client
- **Workaround:** Build native VS Code extension (bypasses SSH)

### Why SSH config must be local:
- SSH client reads `~/.ssh/config` from filesystem
- No API to inject config programmatically
- **Workaround:** Desktop app manages config transparently OR VS Code extension

### Why token refresh is needed:
- JWT tokens expire for security
- One-time tokens prevent replay attacks
- **Workaround:** Long-lived session tokens (slightly less secure but acceptable)

---

## Summary: What This Tells Us

✅ **Core product works:** SSH to GPU workspaces through cloud proxy is proven
✅ **Security model is sound:** JWT auth + TLS + one-time tokens
✅ **Architecture is correct:** Local proxy helper is right approach for SSH

⚠️ **UX needs work:** Too many manual steps before first connection

**Path forward:**
1. **Short-term (2 weeks):** CLI wrapper + long-lived tokens → 1 command to connect
2. **Mid-term (2 months):** Desktop app → 1-click connect
3. **Long-term (6 months):** VS Code extension → Zero local setup

**For MVP:** Focus on CLI wrapper - gets us 80% of the way with 20% of the effort.
