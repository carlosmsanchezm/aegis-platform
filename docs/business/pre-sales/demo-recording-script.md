# Plan: Aegis Platform Demo Recording Script

## Context

We need to record an 8-12 minute demo of the Aegis platform targeting **DoD/defense engineering leaders** — CTOs, engineering directors, and lead engineers navigating NIST/CMMC frameworks who need secure developer tooling for their teams. The demo should show the full platform capability while emphasizing security architecture and developer experience.

The deployment is live at:
- UI: `https://ui.aegis-platform.tech`
- Platform API: `platform-api.aegis-platform.tech:8081`
- Keycloak: `https://keycloak.aegis-platform.tech`
- Proxy: `wss://proxy.aegis-platform.tech:8080`

## Demo Script (8-12 minutes)

### Act 1: The Problem (60 seconds)
**[Browser — title slide or simple text page]**

Narrate the problem:
- "In defense and intelligence environments, data scientists and engineers need GPU-accelerated development environments"
- "Today this means weeks of security reviews, manual cluster provisioning, and VPN configurations"
- "Aegis eliminates this. One platform, one command, secure by default."

### Act 2: One-Shot Infrastructure (90 seconds)
**[Terminal — show the deploy script output]**

- Show the terminal output from `generate-cloud-deployment.sh` (already captured)
- Highlight key moments:
  - Internal PKI auto-provisioned (step-ca + cert-manager)
  - Helm deploys all services in one pass
  - Keycloak realm auto-imported with OIDC clients
  - Spoke agent registers, project bootstrapped
  - DNS records updated automatically
- Key message: "From zero to a fully operational, TLS-encrypted, OIDC-authenticated platform in one command"

### Act 3: Security Architecture (2 minutes)
**[Browser — Keycloak admin + platform-api discovery endpoint]**

1. Open `https://keycloak.aegis-platform.tech` — show the Keycloak admin console
   - Show the `aegis` realm with OIDC clients (backstage, spoke-agent, vscode-extension)
   - Show role-based access control (workspace-admin role)
   - Key message: "Every component authenticates via OIDC — no shared secrets, no API keys"

2. Open `http://platform-api.aegis-platform.tech:8080/api/v1/discovery`
   - Show the auto-discovery response: gRPC endpoint, auth authority, inline root CA
   - Key message: "Clients auto-discover the entire security chain — zero manual certificate distribution"

3. Narrate the trust chain:
   - "Internal PKI issues short-lived TLS certificates via step-ca"
   - "Every service-to-service call is mTLS encrypted"
   - "The proxy validates JWT tokens on every WebSocket connection — no SSH keys to manage"

### Act 4: The UI — Project & Workspace Management (2 minutes)
**[Browser — Backstage UI at ui.aegis-platform.tech]**

1. Navigate to `https://ui.aegis-platform.tech`
   - Login via Keycloak (OIDC redirect)
   - Show the Backstage dashboard

2. Show the workspace creation flow:
   - Navigate to workspace wizard
   - Show project selection (project `default` already exists)
   - Show cluster selection (cluster `default-us-east-1-prod` with flavor `cpu-small`)
   - Create a workspace — select flavor, name it
   - Show the workspace transitioning: PLACED → RUNNING

3. Key messages:
   - "Projects provide tenant isolation — each project has its own clusters, budgets, and access policies"
   - "Workload placement is automatic — the platform matches the requested GPU profile to available clusters"
   - "The entire workspace lifecycle is managed — creation, suspension, resumption, termination"

### Act 5: Zero-Config Developer Experience (2.5 minutes)
**[VS Code — Sovran extension]**

1. Open VS Code with the Sovran extension
   - Show the single setting: `aegisRemote.platform.url: "platform-api.aegis-platform.tech:8080"`
   - Key message: "One URL. That's the only configuration a developer needs."

2. Show the extension activating:
   - Discovery fetches platform config automatically
   - Root CA obtained inline — no manual cert installation
   - Login prompt → Keycloak OIDC → browser redirect → token cached

3. Show the workspace tree view:
   - Workspaces listed with status indicators (Running = green)
   - Click "Connect" on the running workspace
   - VS Code opens a remote window connected to the workspace

4. In the remote workspace:
   - Show terminal access (full Linux environment)
   - Show file explorer (workspace root)
   - Key message: "The developer is now inside a GPU-accelerated environment on an EKS cluster, with full IDE capabilities, accessed through a zero-trust proxy — no VPN, no SSH keys, no manual configuration"

### Act 6: Close (30 seconds)
**[Browser — back to UI or title slide]**

- "Aegis turns weeks of security review and infrastructure setup into minutes"
- "NIST 800-53 controls are built in — not bolted on"
- "One command to deploy. One URL to connect. Zero trust throughout."

---

## Implementation

### Approach
I drive Chrome through the entire UI flow — navigating, clicking, reading pages — and provide feedback on what each screen shows, what's demo-worthy, and what needs work. This is a **scouting run** to design the demo together before you record it.

You handle the VS Code / Sovran extension demo separately.

### Browser walkthrough sequence

**Step 0: Pre-flight**
- Flush DNS cache
- Submit a workspace via grpcurl so there's a RUNNING workspace to show
- Verify all pods healthy

**Step 1: Deploy output (30s of demo)**
- Show the saved terminal output from `generate-cloud-deployment.sh`
- Highlight: PKI, Keycloak, DNS, all green checkmarks
- "One command, full security stack"

**Step 2: Discovery endpoint**
- Navigate to `http://platform-api.aegis-platform.tech:8080/api/v1/discovery`
- Show the JSON: gRPC endpoint, auth authority, inline root CA
- "Zero-config client bootstrap — this is all a developer needs"

**Step 3: Keycloak admin**
- Navigate to `https://keycloak.aegis-platform.tech`
- Login as admin
- Show: aegis realm, OIDC clients, roles, security settings
- "Every component authenticates via OIDC — zero shared secrets"

**Step 4: Backstage UI**
- Navigate to `https://ui.aegis-platform.tech`
- Login via Keycloak redirect
- Explore: dashboard, workspace wizard, project views, cluster views
- Create a workspace through the UI wizard
- "This is what your engineers see — one click to a GPU environment"

**Step 5: Workspace running**
- Show the workspace in RUNNING state
- Show connection details
- Hand off to you for VS Code demo

### What I'll report back
After walking through each screen:
- What's there and what's demo-ready
- What's missing or needs polish
- Recommended demo flow with specific narration points
- Screenshots/GIFs of key moments

### Pre-requisites
- Deployment is live (confirmed — all pods running)
- DNS resolves (flush cache first)
- At least one workspace submitted and RUNNING
- Keycloak admin credentials available
