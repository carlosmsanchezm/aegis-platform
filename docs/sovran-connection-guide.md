# Connecting to Aegis Workspaces with VS Code

## Prerequisites

- VS Code (v1.99+)
- Aegis Remote extension (Sovran) installed
- Access to an Aegis Platform deployment
- Keycloak account with workspace access

## Quick Start

### Step 1: Install the Extension

Install the Aegis Remote extension from a `.vsix` file provided by your platform administrator:

```bash
code --install-extension aegis-remote-*.vsix
```

### Step 2: Configure

Add **one setting** to your VS Code settings (`Cmd+,` → search "aegis"):

```json
{
  "aegisRemote.platform.url": "your-platform.example.com"
}
```

Replace `your-platform.example.com` with the platform URL provided by your administrator.

That's it. Everything else — gRPC endpoint, authentication, CA trust — is auto-discovered from the platform.

### Step 3: Connect

1. Open the **Aegis Workspaces** panel in the VS Code sidebar
2. Sign in when prompted (Keycloak SSO + MFA)
3. Your projects and workspaces appear automatically
4. Click a running workspace to connect
5. A remote VS Code window opens with terminal, file explorer, and full IDE access

## How It Works

```
VS Code (Sovran Extension)
  │
  ├─ Fetch /api/v1/discovery ──► Platform API (public metadata, no auth)
  │   Returns: gRPC endpoint, Keycloak URL, PKI CA URL
  │
  ├─ Fetch root CA ──► Platform PKI endpoint (TLS bootstrap)
  │
  ├─ OIDC/PKCE login ──► Keycloak (browser popup, MFA)
  │
  ├─ ListProjects + ListWorkloads ──► Platform API (gRPC, JWT)
  │
  ├─ CreateConnectionSession ──► Platform API (JWT proxy ticket)
  │
  └─ WebSocket Secure ──► Proxy ──► Workspace Pod (port 11111)
      └─ VS Code Remote Authority protocol
```

## Security

- All connections are TLS-encrypted
- Authentication via Keycloak OIDC with PKCE (phishing-resistant MFA supported)
- Session tokens are short-lived (5-minute max TTL) and one-time-use
- Proxy validates JWT signature, audience, expiry, and JTI on every connection
- Idle sessions terminate after 15 minutes
- Full audit trail of all authentication and connection events

## Advanced Configuration

For environments where auto-discovery is not available (air-gapped, custom DNS), configure individual settings:

```json
{
  "aegisRemote.platform.grpcEndpoint": "platform-api.internal:8081",
  "aegisRemote.auth.authority": "https://keycloak.internal/realms/aegis",
  "aegisRemote.auth.clientId": "vscode-extension",
  "aegisRemote.security.caPath": "/path/to/ca-bundle.pem"
}
```

## Troubleshooting

**"No workspaces found"**
- Verify you have access to at least one project in the Aegis UI
- Check that workspaces are in RUNNING status

**"Unable to get local issuer certificate"**
- The CA trust auto-fetch may have failed. Set `aegisRemote.security.caPath` to your platform's CA certificate file.

**"workspace is not interactive"**
- The workspace was not submitted as interactive. Re-submit from the Aegis UI.

**Connection timeout**
- Verify the platform URL is reachable: `curl https://your-platform/api/v1/discovery`
- Check firewall rules allow outbound HTTPS (443) and gRPC (8081)
