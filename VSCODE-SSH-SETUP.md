# VS Code Remote-SSH Setup for Aegis

## Quick Answer to Your Questions

### Why does aegis-connect need to be local?

**aegis-connect MUST be on your local machine** because it's part of the SSH connection process:

```
VS Code (your Mac) → SSH client (your Mac) → ProxyCommand: aegis-connect (your Mac) → Internet → Cloud Proxy
```

When you connect to a remote host via SSH, the `ProxyCommand` in `~/.ssh/config` runs **locally** to establish the connection tunnel. It's like having a local "adapter" that speaks HTTPS to the cloud proxy.

Think of it as:
- **SSH**: Direct connection to server
- **SSH + Jump Host**: SSH tunnels through another SSH server
- **SSH + aegis-connect**: SSH tunnels through HTTPS proxy (with JWT auth)

aegis-connect is installed at: `~/.local/bin/aegis-connect`

## Current Setup Status

✅ **What's Working:**
- Cloud proxy exposed via AWS NLB
- Platform-API creating JWT tokens
- aegis-connect installed locally
- Workspace pod running with SSH on port 2222
- SSH connections work through the cloud proxy

⚠️ **Current Issue:**
- SSH key authentication needs to be set up OR
- VS Code will prompt for password (`aegis123`) on each connection

## Option 1: Use Password Authentication (Easiest)

### Step 1: Get Fresh Connection Session

```bash
# Port-forward to platform-api (if not already running)
kubectl port-forward -n aegis-services svc/aegis-services-aegis-services-platform-api 8081:8081 &

# Create connection session
grpcurl -plaintext -H "x-aegis-user: testuser@test.com" \
  -d '{"workload_id":"wl-ssh-workspace","client":"vscode"}' \
  localhost:8081 aegis.v1.AegisPlatform/CreateConnectionSession | \
  jq -r '.sshConfig' | sed 's/User aegis-[a-z0-9]*/User aegis/' >> ~/.ssh/config
```

### Step 2: Connect with VS Code

1. Open VS Code
2. Install "Remote - SSH" extension if you haven't
3. Press `Cmd+Shift+P`
4. Type: `Remote-SSH: Connect to Host...`
5. Select: `aegis-w-wl-ssh-workspace`
6. When prompted for password, enter: `aegis123`
7. Wait for VS Code Server to install (~1 minute)
8. ✅ You're connected!

**Note:** VS Code will prompt for password each time you connect because JWT tokens are one-time use.

## Option 2: Use SSH Keys (Passwordless)

### Step 1: Generate SSH Key (if not already done)

```bash
ssh-keygen -t ed25519 -f ~/.ssh/aegis_workspace_key -N ""
```

### Step 2: Add Key to Workspace

```bash
# Get fresh session
grpcurl -plaintext -H "x-aegis-user: testuser@test.com" \
  -d '{"workload_id":"wl-ssh-workspace","client":"vscode"}' \
  localhost:8081 aegis.v1.AegisPlatform/CreateConnectionSession | \
  jq -r '.sshConfig' | sed 's/User aegis-[a-z0-9]*/User aegis/' > /tmp/temp-ssh.config

# Add public key to workspace
sshpass -p 'aegis123' ssh -F /tmp/temp-ssh.config -o StrictHostKeyChecking=no aegis-w-wl-ssh-workspace \
  "mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys" \
  < ~/.ssh/aegis_workspace_key.pub
```

### Step 3: Update SSH Config with Key

```bash
# Get new session and add key reference
grpcurl -plaintext -H "x-aegis-user: testuser@test.com" \
  -d '{"workload_id":"wl-ssh-workspace","client":"vscode"}' \
  localhost:8081 aegis.v1.AegisPlatform/CreateConnectionSession | \
  jq -r '.sshConfig' | sed 's/User aegis-[a-z0-9]*/User aegis/' > /tmp/final-ssh.config

# Add key reference
echo "  IdentityFile ~/.ssh/aegis_workspace_key" >> /tmp/final-ssh.config

# Add to SSH config
cat /tmp/final-ssh.config >> ~/.ssh/config
```

### Step 4: Test and Connect

```bash
# Test SSH works
ssh aegis-w-wl-ssh-workspace 'echo "SUCCESS"'

# If that works, open VS Code and connect (no password needed!)
```

## Simplified Connection Workflow (Recommended)

We've created a helper script to make reconnections easy:

### One-Time Setup

```bash
# Generate SSH key (if not already done)
ssh-keygen -t ed25519 -f ~/.ssh/aegis_workspace_key -N ""

# Add public key to workspace (do this once per workspace)
aegis-refresh-ssh
sshpass -p 'aegis123' ssh aegis-w-wl-ssh-workspace \
  "mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys" \
  < ~/.ssh/aegis_workspace_key.pub
```

### Every Time You Want to Connect

```bash
# 1. Refresh the JWT token (takes 1 second)
aegis-refresh-ssh

# 2. Open VS Code and connect to: aegis-w-wl-ssh-workspace
```

That's it! The helper script handles token refresh and SSH config updates automatically.

### For GPU Workspaces

```bash
aegis-refresh-ssh wl-gpu-workspace
```

## Manual Connection Workflow (Advanced)

If you prefer not to use the helper script:

1. **Create new JWT token** (they're one-time use):
   ```bash
   grpcurl -plaintext -H "x-aegis-user: testuser@test.com" \
     -d '{"workload_id":"wl-ssh-workspace","client":"vscode"}' \
     localhost:8081 aegis.v1.AegisPlatform/CreateConnectionSession | \
     jq -r '.sshConfig' | sed 's/User aegis-[a-z0-9]*/User aegis/' > /tmp/new-session.txt
   ```

2. **Update SSH config**:
   ```bash
   # Remove old entry
   grep -v "aegis-w-wl-ssh-workspace" ~/.ssh/config > ~/.ssh/config.tmp
   cat ~/.ssh/config.tmp > ~/.ssh/config
   rm ~/.ssh/config.tmp

   # Add new entry
   cat /tmp/new-session.txt >> ~/.ssh/config

   # Add key if using key-based auth
   echo "  IdentityFile ~/.ssh/aegis_workspace_key" >> ~/.ssh/config
   ```

3. **Connect with VS Code**

## What Happens When You Connect

```
1. VS Code triggers: ssh aegis-w-wl-ssh-workspace

2. SSH reads ~/.ssh/config and sees:
   ProxyCommand aegis-connect --proxy=https://... --token=...

3. SSH launches aegis-connect locally with the JWT token

4. aegis-connect:
   - Opens HTTPS connection to AWS NLB
   - NLB forwards to aegis-auth-proxy pod
   - Sends HTTP CONNECT request with JWT in Authorization header

5. aegis-auth-proxy:
   - Validates JWT signature
   - Checks expiration
   - Marks token as "used" (one-time tokens)
   - Dials workspace Service: aegis-w-wl-ssh-workspace.default.svc.cluster.local:2222

6. TCP tunnel established:
   - All SSH traffic flows through the HTTPS tunnel
   - aegis-connect becomes transparent pipe

7. SSH completes handshake with workspace pod

8. VS Code installs/launches VS Code Server on workspace

9. You can now code remotely! 🎉
```

## Troubleshooting

### "Connection timeout"
- Token expired (get new session)
- Proxy pod not running
- NLB not accessible

### "Permission denied"
- Wrong username (should be `aegis` not `aegis-d3b42d2c`)
- Wrong password (`aegis123`)
- SSH key not added to workspace

### "access denied" from proxy
- JWT token already used (one-time tokens)
- Get new connection session

### VS Code can't find aegis-connect
- Ensure it's installed: `which aegis-connect`
- Should be at: `/Users/carlossanchez/.local/bin/aegis-connect`
- Make sure `~/.local/bin` is in your PATH

## Testing Commands

```bash
# Test workspace is running
kubectl get pods -n default | grep wl-ssh-workspace

# Test direct SSH (within cluster)
kubectl exec -it <workspace-pod> -- /bin/bash

# Test SSH through proxy (from your Mac)
ssh aegis-w-wl-ssh-workspace 'hostname'

# Check proxy logs
kubectl logs -n aegis-services deployment/aegis-services-aegis-services-proxy --tail=20
```

## For GPU Workspaces

Same process, just use different workload ID:

```bash
grpcurl -plaintext -H "x-aegis-user: testuser@test.com" \
  -d '{"workload_id":"wl-gpu-workspace","client":"vscode"}' \
  localhost:8081 aegis.v1.AegisPlatform/CreateConnectionSession
```

Then connect with VS Code and run in the terminal:
```bash
nvidia-smi
python3 -c "import torch; print(torch.cuda.is_available())"
```

## Architecture Summary

**Local (Your Mac):**
- VS Code with Remote-SSH extension
- SSH client (uses ~/.ssh/config)
- aegis-connect (at ~/.local/bin/aegis-connect)

**Cloud (AWS/EKS):**
- AWS Network Load Balancer (public internet → cluster)
- aegis-auth-proxy pod (JWT validation, HTTPS→TCP proxying)
- Workspace pod (SSH server on port 2222, your code/GPU)

**Flow:**
```
Your Mac → aegis-connect → Internet → AWS NLB → aegis-auth-proxy → Workspace Pod
  (VS Code)      (JWT)                   (TLS)    (TCP tunnel)     (SSH:2222)
```
