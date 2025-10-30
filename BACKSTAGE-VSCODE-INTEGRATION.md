# Backstage + VS Code Extension Integration Guide

## Overview

This document explains how Backstage (Aegis Platform UI) integrates with the VS Code extension to provide a complete workspace management and development experience.

---

## Architecture: Backstage as Control Plane + VS Code as Dev Environment

```
┌─────────────────────────────────────────────────────────────┐
│  User's Local Machine                                       │
│                                                              │
│  ┌──────────────────┐         ┌─────────────────────┐      │
│  │  Web Browser     │         │  VS Code (Insiders) │      │
│  │  (Backstage UI)  │         │  + Aegis Extension  │      │
│  └────────┬─────────┘         └──────────┬──────────┘      │
│           │                              │                  │
└───────────┼──────────────────────────────┼──────────────────┘
            │                              │
            │ HTTPS                        │ WSS + JWT
            ▼                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Cloud (AWS/EKS)                                            │
│                                                              │
│  ┌──────────────────────┐    ┌─────────────────────────┐   │
│  │  Backstage Service   │    │  AWS NLB (public)       │   │
│  │  (port 7007)         │    │  → aegis-auth-proxy     │   │
│  └──────────┬───────────┘    └──────────┬──────────────┘   │
│             │                           │                   │
│             │ gRPC                      │ WS→TCP tunnel     │
│             ▼                           ▼                   │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Platform-API (gRPC)                                 │  │
│  │  - CreateWorkspace                                   │  │
│  │  - CreateConnectionSession (returns JWT + proxy URL)│  │
│  │  - ListWorkspaces                                    │  │
│  └──────────┬───────────────────────────────────────────┘  │
│             │                                               │
│             │ creates AegisWorkload CRs                     │
│             ▼                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Workspace Pods (REH on :11111, SSH on :2222)       │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## Backstage's Role in This Flow

### 1. Workspace Management UI (Primary Use Case)

Backstage provides the **web UI** for workspace lifecycle management:

#### **Before VS Code Connection:**
- **Browse templates** - View available workspace configurations (GPU/CPU flavors)
- **Create workspaces** - Select image, resources, mount volumes
- **View status** - See workspace state (starting/running/stopped/failed)
- **Manage lifecycle** - Stop, restart, delete workspaces
- **Configure settings** - Set environment variables, resource limits

#### **During/After VS Code Connection:**
- **Monitor activity** - See which workspaces are actively connected
- **Resource usage** - View CPU/GPU/memory consumption
- **Session logs** - Connection history, duration, data transferred
- **Team collaboration** - Share workspaces, manage permissions
- **Billing/quotas** - Track costs, enforce limits

---

## User Workflows

### **Workflow A: Backstage → VS Code (Recommended)**

This is the primary flow for most users:

```
Step 1: User opens Backstage in web browser
  ↓
Step 2: User clicks "Create Workspace"
        → Selects: GPU flavor, container image, volumes
        → Backstage calls platform-api.CreateWorkload()
        → Workspace pod starts in EKS
  ↓
Step 3: Workspace status updates in Backstage UI
        → "Starting..." → "Running" (shows ready indicator)
  ↓
Step 4: User clicks "Connect with VS Code" button
        → Backstage calls platform-api.CreateConnectionSession()
        → Returns connection details (JWT token + proxy URL)
        → Opens vscode:// deep link OR copies connection command
  ↓
Step 5: VS Code launches automatically (via deep link)
        → Aegis extension activates
        → Extension uses connection details from deep link
        → Opens Remote window via WSS tunnel to workspace
  ↓
Step 6: User codes in VS Code Remote window
        → Backstage shows "Connected" status
        → Monitors resource usage in dashboard
```

**Benefits:**
- ✅ Guided workspace creation (pick from templates)
- ✅ Visual status feedback (loading indicators, errors)
- ✅ One-click connection
- ✅ Centralized management (see all workspaces in one place)

---

### **Workflow B: VS Code Only (Power Users)**

Advanced users can bypass Backstage for faster access:

```
Step 1: User opens VS Code
  ↓
Step 2: Aegis extension TreeView shows workspace list
        → Extension calls platform-api.ListWorkspaces()
        → Shows: workspace ID, status, GPU count
  ↓
Step 3: User clicks "Connect" on existing workspace
        → Extension calls platform-api.CreateConnectionSession()
        → Gets JWT token + proxy URL
        → Opens Remote window via WSS tunnel
  ↓
Step 4: (Optional) Create new workspace from VS Code
        → Command: "Aegis: Create Workspace"
        → Quick picker: Select flavor/image
        → Extension calls platform-api.CreateWorkload()
        → Auto-connects when ready
```

**Benefits:**
- ✅ Faster for repeat connections (no browser needed)
- ✅ Stay in IDE (developer-focused workflow)
- ✅ Keyboard shortcuts (e.g., Cmd+Shift+P → "Aegis: Connect")

**Trade-offs:**
- ⚠️ Less visual feedback (status bar vs. full UI)
- ⚠️ Limited management features (can't see team workspaces, billing, etc.)

---

## Deployment Options

### **Option A: Backstage in Cloud (Recommended for Production)**

Deploy Backstage as a service in EKS, accessible via public URL.

#### **Kubernetes Deployment**

```yaml
# backstage-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backstage
  namespace: aegis-services
spec:
  replicas: 2
  selector:
    matchLabels:
      app: backstage
  template:
    metadata:
      labels:
        app: backstage
    spec:
      containers:
      - name: backstage
        image: your-org/aegis-backstage:latest
        ports:
        - containerPort: 7007
        env:
        - name: PLATFORM_API_ENDPOINT
          value: "aegis-services-platform-api.aegis-services.svc.cluster.local:8081"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: backstage-secrets
              key: postgres-url
        - name: AUTH_GOOGLE_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: backstage-auth
              key: google-client-id
---
apiVersion: v1
kind: Service
metadata:
  name: backstage
  namespace: aegis-services
spec:
  type: LoadBalancer  # or use Ingress
  ports:
  - port: 443
    targetPort: 7007
    protocol: TCP
  selector:
    app: backstage
```

#### **Ingress (for HTTPS with custom domain)**

```yaml
# backstage-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: backstage
  namespace: aegis-services
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - backstage.yourcompany.com
    secretName: backstage-tls
  rules:
  - host: backstage.yourcompany.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: backstage
            port:
              number: 7007
```

#### **Access URL**
```
https://backstage.yourcompany.com
```

#### **Pros:**
- ✅ Accessible from anywhere (team members, remote workers)
- ✅ Multi-user with SSO (Google, Okta, GitHub, etc.)
- ✅ Persistent workspace catalog (PostgreSQL backend)
- ✅ Team collaboration (shared workspaces, permissions)
- ✅ Centralized monitoring (audit logs, usage metrics)
- ✅ Production-ready (HA, auto-scaling, backups)

#### **Cons:**
- ⚠️ Requires hosting infrastructure (EKS resources)
- ⚠️ Needs monitoring/alerting setup
- ⚠️ Auth configuration required

#### **When to use:**
- **Production deployments**
- **Multi-user teams** (5+ developers)
- **Enterprise environments** with SSO requirements

---

### **Option B: Backstage Local (Good for Testing/Development)**

Run Backstage on your local machine for rapid development.

#### **Setup**

```bash
# Clone and set up Backstage
cd aegis-platform/backstage
yarn install

# Configure platform-api endpoint
cat > app-config.local.yaml << EOF
backend:
  baseUrl: http://localhost:7007
  database:
    client: sqlite3
    connection: ':memory:'

aegis:
  platformAPI:
    endpoint: localhost:8081  # port-forwarded
EOF

# Start Backstage dev server
yarn dev  # runs on http://localhost:3000
```

#### **Port-forward to platform-api** (in another terminal)

```bash
kubectl port-forward -n aegis-services \
  svc/aegis-services-platform-api 8081:8081
```

#### **Access URL**
```
http://localhost:3000
```

#### **Pros:**
- ✅ Fast iteration during development
- ✅ No cloud resources needed (cost-effective)
- ✅ Easy debugging (browser DevTools)
- ✅ Quick config changes (no deploy cycle)
- ✅ Works offline (once workspace list cached)

#### **Cons:**
- ⚠️ Only accessible to you (single-user)
- ⚠️ Needs port-forwards to platform-api
- ⚠️ In-memory database (data lost on restart)
- ⚠️ Not suitable for team collaboration

#### **When to use:**
- **Local development** of Backstage plugins
- **Testing** new features before deploying
- **Demo/prototyping** without cloud setup

---

## Integration Points: Backstage ↔ VS Code Extension

### 1. "Connect with VS Code" Button in Backstage

Add a button to each workspace card in the Backstage UI that launches VS Code.

#### **Backstage Component (React)**

```typescript
// backstage/plugins/aegis/src/components/WorkspaceCard.tsx
import React from 'react';
import { Button, Card, CardContent, Typography } from '@material-ui/core';
import { useApi, configApiRef } from '@backstage/core-plugin-api';

export const WorkspaceCard = ({ workspace }) => {
  const config = useApi(configApiRef);
  const platformAPI = config.getString('aegis.platformAPI.endpoint');

  const handleConnectVSCode = async () => {
    try {
      // Call platform-api to get connection session
      const response = await fetch(`http://${platformAPI}/api/sessions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}` // token from Backstage identity API
        },
        body: JSON.stringify({
          workload_id: workspace.id,
          client: 'vscode'
        })
      });

      const session = await response.json();

      // Option A: Deep link (opens VS Code directly)
      const deepLink = `vscode://aegis.aegis-remote/connect?` +
        `wid=${workspace.id}&` +
        `token=${encodeURIComponent(session.token)}&` +
        `proxy=${encodeURIComponent(session.proxyUrl)}`;

      window.location.href = deepLink;

      // Option B: Copy command to clipboard
      const command = `code-insiders --folder-uri vscode-remote://aegis+${workspace.id}/home/project`;
      await navigator.clipboard.writeText(command);

      // Show notification
      showNotification('VS Code connection ready! Click the notification to open.');

    } catch (error) {
      showError(`Failed to create connection: ${error.message}`);
    }
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h6">{workspace.id}</Typography>
        <Typography color="textSecondary">
          Status: {workspace.status}
        </Typography>
        <Typography color="textSecondary">
          GPU: {workspace.hints?.gpuCount || 0}x
        </Typography>

        <Button
          variant="contained"
          color="primary"
          onClick={handleConnectVSCode}
          disabled={workspace.status !== 'running'}
        >
          Connect with VS Code
        </Button>
      </CardContent>
    </Card>
  );
};
```

#### **VS Code Extension Deep Link Handler**

```typescript
// extension/src/extension.ts
import * as vscode from 'vscode';

export function activate(ctx: vscode.ExtensionContext) {
  // Register URI handler for deep links
  ctx.subscriptions.push(
    vscode.window.registerUriHandler({
      handleUri(uri: vscode.Uri) {
        // Parse: vscode://aegis.aegis-remote/connect?wid=...&token=...&proxy=...
        const params = new URLSearchParams(uri.query);
        const wid = params.get('wid');
        const token = params.get('token');
        const proxyUrl = params.get('proxy');

        if (wid && token && proxyUrl) {
          // Store connection details for resolver to use
          ctx.globalState.update(`connection.${wid}`, { token, proxyUrl });

          // Open remote window
          const remoteUri = vscode.Uri.parse(`vscode-remote://aegis+${wid}/home/project`);
          vscode.commands.executeCommand('vscode.openFolder', remoteUri, { forceNewWindow: true });
        }
      }
    })
  );
}
```

### 2. Workspace List API Integration

The VS Code extension can call Backstage's backend API to get workspace list (alternatively, can call platform-api directly).

#### **Option A: Call Backstage API**

```typescript
// extension/src/api/backstage-client.ts
export class BackstageClient {
  private baseUrl: string;

  constructor(baseUrl: string = 'https://backstage.yourcompany.com') {
    this.baseUrl = baseUrl;
  }

  async listWorkspaces(user: string): Promise<Workspace[]> {
    const token = await this.getAuthToken(); // from VS Code SecretStorage

    const response = await fetch(`${this.baseUrl}/api/aegis/workspaces`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    return response.json();
  }

  async createConnectionSession(wid: string): Promise<ConnectionSession> {
    const token = await this.getAuthToken();

    const response = await fetch(`${this.baseUrl}/api/aegis/workspaces/${wid}/connect`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    return response.json();
  }
}
```

**Pros:**
- ✅ Centralized API (one source of truth)
- ✅ Backstage can add middleware (rate limiting, caching)
- ✅ Audit logs in one place

**Cons:**
- ⚠️ Backstage must be running for extension to work
- ⚠️ Extra network hop (extension → Backstage → platform-api)

#### **Option B: Call Platform-API Directly (Recommended)**

```typescript
// extension/src/api/platform-api-client.ts
import * as grpc from '@grpc/grpc-js';
import { AegisPlatformClient } from './generated/aegis_grpc_pb';

export class PlatformAPIClient {
  private client: AegisPlatformClient;

  constructor(endpoint: string = 'platform-api.yourcompany.com:443') {
    const credentials = grpc.credentials.createSsl(); // mTLS in production

    this.client = new AegisPlatformClient(endpoint, credentials);
  }

  async listWorkspaces(user: string): Promise<Workspace[]> {
    const request = new ListWorkloadsRequest();
    request.setProjectId('p-demo');

    return new Promise(async (resolve, reject) => {
      const metadata = new grpc.Metadata();
      const token = await this.identityApi.getIdToken();
      metadata.add('authorization', `Bearer ${token}`);

      this.client.listWorkloads(request, metadata, (err, response) => {
        if (err) reject(err);
        else resolve(response.getWorkloadsList());
      });
    });
  }

  async createConnectionSession(wid: string): Promise<ConnectionSession> {
    const request = new CreateConnectionSessionRequest();
    request.setWorkloadId(wid);
    request.setClient('vscode');

    return new Promise((resolve, reject) => {
      this.client.createConnectionSession(request, (err, response) => {
        if (err) reject(err);
        else resolve(response);
      });
    });
  }
}
```

**Pros:**
- ✅ Extension works standalone (no Backstage dependency)
- ✅ Faster (fewer network hops)
- ✅ Simpler architecture

**Cons:**
- ⚠️ Extension must handle auth directly

**Recommendation:** Use **Option B** (direct platform-api calls). This keeps Backstage **optional** - users can use the VS Code extension without Backstage running.

---

## Backstage Features for Aegis Workspaces

Here are key Backstage pages/plugins to build:

### 1. **Workspace Catalog Page**

```
URL: /aegis/workspaces

┌────────────────────────────────────────────────────────────┐
│ Aegis Workspaces                     [+ Create Workspace]  │
├────────────────────────────────────────────────────────────┤
│                                                            │
│ Filters: [All] [Running] [Stopped]   Search: [______]     │
│                                                            │
│ ┌────────────────────────────────────┐                    │
│ │ wl-my-gpu-workspace       [Running]│                    │
│ │ GPU: 1x NVIDIA A100                │                    │
│ │ Created: 2 hours ago               │                    │
│ │ CPU: 32% | Memory: 16GB/32GB       │                    │
│ │                                    │                    │
│ │ [Connect VS Code] [Stop] [Logs]    │                    │
│ └────────────────────────────────────┘                    │
│                                                            │
│ ┌────────────────────────────────────┐                    │
│ │ wl-training-job          [Stopped] │                    │
│ │ CPU: 4 cores, 16GB RAM             │                    │
│ │ Created: 1 day ago                 │                    │
│ │                                    │                    │
│ │ [Start] [Delete]                   │                    │
│ └────────────────────────────────────┘                    │
└────────────────────────────────────────────────────────────┘
```

### 2. **Create Workspace Wizard**

```
URL: /aegis/workspaces/new

┌────────────────────────────────────────────────────────────┐
│ Create New Workspace                     Step 1 of 3       │
├────────────────────────────────────────────────────────────┤
│                                                            │
│ Choose Template:                                           │
│                                                            │
│ ○ Python ML (GPU)                                          │
│   Pre-configured for PyTorch/TensorFlow with GPU support   │
│                                                            │
│ ○ PyTorch Training (Multi-GPU)                             │
│   High-performance training with 4x A100 GPUs              │
│                                                            │
│ ● General Dev (CPU)                                        │
│   Lightweight development environment                      │
│                                                            │
│                            [Cancel] [Next: Configure →]    │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│ Create New Workspace                     Step 2 of 3       │
├────────────────────────────────────────────────────────────┤
│                                                            │
│ Configure Resources:                                       │
│                                                            │
│ Workspace ID: [wl-my-workspace_________]                   │
│                                                            │
│ GPU: [0 ▼] x [NVIDIA A100 ▼]                               │
│                                                            │
│ CPU: [4 ▼] cores                                           │
│                                                            │
│ Memory: [16 ▼] GB                                          │
│                                                            │
│ Container Image:                                           │
│ [carlosmsanchez/aegis-workspace-vscode:latest_______]      │
│                                                            │
│ Estimated cost: $0.50/hour                                 │
│                                                            │
│                            [← Back] [Next: Review →]       │
└────────────────────────────────────────────────────────────┘
```

### 3. **Workspace Details Page**

```
URL: /aegis/workspaces/wl-my-gpu-workspace

┌────────────────────────────────────────────────────────────┐
│ wl-my-gpu-workspace                           [Stop] [⋮]   │
├────────────────────────────────────────────────────────────┤
│                                                            │
│ Overview                                                   │
│ ├─ Status: ● Running (uptime: 3h 24m)                     │
│ ├─ Cost: $4.50/hour                                        │
│ └─ Cluster: aegis-spoke-prod (us-east-1)                   │
│                                                            │
│ Resources                                                  │
│ ├─ GPU: 1x NVIDIA A100                                     │
│ │   └─ Utilization: [██████████░░░░░] 85%                 │
│ ├─ CPU: 8 cores                                            │
│ │   └─ Utilization: [████░░░░░░░░░░░] 32%                 │
│ └─ Memory: 16GB / 32GB                                     │
│     └─ Usage: [████████░░░░░░░░] 50%                      │
│                                                            │
│ Connections                                                │
│ ├─ VS Code: ● Connected (user@company.com)                │
│ │   └─ Connected 2h 15m ago                               │
│ └─ SSH: Last used 1h ago                                   │
│                                                            │
│ Actions                                                    │
│ [Connect VS Code] [Open SSH Terminal] [View Logs]         │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### 4. **Team Workspaces (Advanced Feature)**

```
URL: /aegis/team/ml-team

┌────────────────────────────────────────────────────────────┐
│ ML Team Workspaces                                         │
├────────────────────────────────────────────────────────────┤
│                                                            │
│ Shared Workspaces:                                         │
│                                                            │
│ ├─ wl-shared-dataset              [Read-only]             │
│ │   └─ 3 members can access                               │
│ │                                                          │
│ └─ wl-model-training              [Active]                │
│     └─ Currently used by: alice@company.com               │
│                                                            │
│ Team Members:                                              │
│ ├─ alice@company.com (Owner)                               │
│ ├─ bob@company.com (Contributor)                           │
│ └─ carol@company.com (Viewer)                              │
│                                                            │
│ Usage This Month:                                          │
│ ├─ GPU hours: 340 / 500                                    │
│ ├─ Cost: $1,250 / $2,000 budget                            │
│ └─ Active workspaces: 5 / 10 limit                         │
└────────────────────────────────────────────────────────────┘
```

---

## Backstage Backend Implementation

### **Backstage Plugin Backend Router**

```typescript
// backstage/plugins/aegis-backend/src/service/router.ts
import { createRouter } from '@backstage/backend-common';
import { AegisPlatformClient } from './grpc-client';
import { Logger } from 'winston';

export interface RouterOptions {
  logger: Logger;
  platformAPIEndpoint: string;
}

export async function createRouter(options: RouterOptions) {
  const { logger, platformAPIEndpoint } = options;
  const router = Router();
  const platformAPI = new AegisPlatformClient(platformAPIEndpoint);

  // List workspaces
  router.get('/workspaces', async (req, res) => {
    try {
      const user = req.user?.entity?.metadata?.name; // from Backstage auth
      const workspaces = await platformAPI.listWorkspaces({ user });

      res.json({
        items: workspaces.map(w => ({
          id: w.id,
          status: w.status,
          flavor: w.workspace.flavor,
          gpuCount: w.hints?.gpuCount || 0,
          createdAt: w.createdAt,
          updatedAt: w.updatedAt
        }))
      });
    } catch (error) {
      logger.error('Failed to list workspaces', error);
      res.status(500).json({ error: error.message });
    }
  });

  // Create workspace
  router.post('/workspaces', async (req, res) => {
    try {
      const { flavor, image, gpu_count, cpu_request, memory_request } = req.body;
      const user = req.user?.entity?.metadata?.name;

      const workspace = await platformAPI.createWorkload({
        projectId: 'p-demo',
        workspace: {
          flavor,
          image,
          interactive: true
        },
        hints: {
          gpuCount: gpu_count,
          cpuRequest: cpu_request,
          memoryRequest: memory_request
        }
      });

      logger.info(`Created workspace ${workspace.id} for user ${user}`);
      res.status(201).json(workspace);
    } catch (error) {
      logger.error('Failed to create workspace', error);
      res.status(500).json({ error: error.message });
    }
  });

  // Get connection session (for "Connect VS Code" button)
  router.post('/workspaces/:wid/connect', async (req, res) => {
    try {
      const { wid } = req.params;
      const user = req.user?.entity?.metadata?.name;

      const session = await platformAPI.createConnectionSession({
        workloadId: wid,
        client: 'vscode'
      });

      logger.info(`Created connection session for ${wid} (user: ${user})`);

      // Return multiple formats for flexibility
      res.json({
        // Deep link (opens VS Code directly)
        deepLink: `vscode://aegis.aegis-remote/connect?wid=${wid}&token=${encodeURIComponent(session.token)}&proxy=${encodeURIComponent(session.proxyUrl)}`,

        // Connection details (for manual setup)
        proxyUrl: session.proxyUrl,
        token: session.token,
        sshConfig: session.sshConfig,

        // VS Code Remote URI
        vscodeUri: `vscode-remote://aegis+${wid}/home/project`
      });
    } catch (error) {
      logger.error(`Failed to create connection session for ${wid}`, error);
      res.status(500).json({ error: error.message });
    }
  });

  // Stop workspace
  router.post('/workspaces/:wid/stop', async (req, res) => {
    try {
      const { wid } = req.params;
      await platformAPI.stopWorkload({ workloadId: wid });
      res.json({ message: 'Workspace stopped' });
    } catch (error) {
      logger.error(`Failed to stop workspace ${wid}`, error);
      res.status(500).json({ error: error.message });
    }
  });

  // Delete workspace
  router.delete('/workspaces/:wid', async (req, res) => {
    try {
      const { wid } = req.params;
      await platformAPI.deleteWorkload({ workloadId: wid });
      res.json({ message: 'Workspace deleted' });
    } catch (error) {
      logger.error(`Failed to delete workspace ${wid}`, error);
      res.status(500).json({ error: error.message });
    }
  });

  return router;
}
```

---

## Recommended Setup for Testing the VS Code Extension

### **Phase 1: Local Development (Just You) - Week 1-2**

Run everything locally for rapid iteration:

```bash
# Terminal 1: Port-forward to platform-api in cloud
kubectl port-forward -n aegis-services \
  svc/aegis-services-platform-api 8081:8081

# Terminal 2: Run local Backstage (optional, for UI testing)
cd aegis-platform/backstage
yarn install
yarn dev  # http://localhost:3000

# Terminal 3: Run VS Code extension in debug mode
cd aegis-vscode-remote/extension
npm install
npm run watch
# Press F5 in VS Code Insiders to launch Extension Dev Host
```

**What you're testing:**
- ✅ VS Code extension connects to cloud workspaces
- ✅ Backstage UI shows workspace status
- ✅ Both talk to same platform-api
- ✅ "Connect VS Code" button in Backstage works

---

### **Phase 2: Team Testing (Cloud Backstage + Extension) - Week 3-4**

Deploy Backstage to EKS for multi-user testing:

```bash
# Deploy Backstage to EKS (one-time)
kubectl apply -f backstage-deployment.yaml
kubectl apply -f backstage-ingress.yaml

# Access via public URL
# https://backstage-test.yourcompany.com

# Team members:
# 1. Install VS Code Insiders
# 2. Install Aegis extension from .vsix file:
#    code-insiders --install-extension aegis-remote-0.0.1.vsix
# 3. Open Backstage, create workspace
# 4. Click "Connect with VS Code" or use extension TreeView
```

**What you're testing:**
- ✅ Multi-user scenarios (5-10 team members)
- ✅ SSO/auth integration (Google, Okta, etc.)
- ✅ Real-world network conditions
- ✅ Workspace sharing & permissions
- ✅ Load testing (concurrent connections)

---

## Summary: Where to Run What

| Component | Local Dev | Team Testing | Production |
|-----------|-----------|--------------|------------|
| **VS Code Extension** | Local (F5 debug) | User's machine | User's machine |
| **Backstage UI** | `localhost:3000` | EKS (https://backstage-test...) | EKS (https://backstage...) |
| **Platform-API** | Port-forward (8081) | EKS | EKS |
| **Aegis-Auth-Proxy** | Port-forward or NLB | EKS + NLB | EKS + NLB |
| **Workspaces** | EKS | EKS | EKS |

---

## Integration Timeline

### **Week 1-2: Local Development**
- **Goal:** Get basic integration working locally
- **Tasks:**
  - Run Backstage locally (`yarn dev`)
  - Add "Connect VS Code" button to workspace cards
  - Test deep link opens VS Code extension
  - Extension connects using connection details from deep link

### **Week 3-4: Cloud Deployment**
- **Goal:** Deploy Backstage to EKS for team testing
- **Tasks:**
  - Deploy Backstage service + ingress
  - Configure SSO (Google/Okta)
  - Test multi-user access
  - Package extension as `.vsix` for team install

### **Week 5-6: Polish & Features**
- **Goal:** Production-ready integration
- **Tasks:**
  - Add workspace templates (ML, training, dev)
  - Resource usage monitoring (CPU/GPU graphs)
  - Connection status indicators
  - Team workspace sharing

---

## Key Insights

### **Complementary, Not Competing**

Backstage and VS Code extension serve different purposes:

| Aspect | Backstage | VS Code Extension |
|--------|-----------|-------------------|
| **Purpose** | Workspace management & team collaboration | Direct connection to workspaces |
| **Interface** | Web UI (browser) | Native IDE (VS Code) |
| **Users** | All team members (devs, PMs, admins) | Developers only |
| **Features** | Create, monitor, share, billing | Connect, code, debug, terminal |
| **Access** | Any device with browser | Desktop/laptop with VS Code |

### **Users Can Choose Their Workflow**

- **Casual users:** Use Backstage only (click "Connect VS Code" button)
- **Power users:** Use extension only (faster, keyboard-driven)
- **Managers/admins:** Use Backstage for monitoring & cost tracking
- **Hybrid:** Create in Backstage, connect via extension

### **Extension Should Work Standalone**

**Recommendation:** Make the VS Code extension **independent** of Backstage:
- Call platform-api directly (not through Backstage)
- TreeView shows workspace list without Backstage
- Users can create workspaces from extension (optional)

This keeps Backstage **optional** - nice to have, but not required.

---

## Conclusion

Backstage provides the **control plane** for Aegis workspaces:
- **Web UI** for lifecycle management
- **Team collaboration** features
- **Monitoring & billing** dashboards

VS Code extension provides the **developer experience**:
- **Native IDE integration**
- **Fast, one-click connections**
- **Familiar workflow** for developers

Together, they create a **complete platform** for cloud-based GPU development.
