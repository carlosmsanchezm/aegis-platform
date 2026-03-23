# Aegis Platform — Demo Script

## Pre-Demo Checklist

| Item | Status | Notes |
|---|---|---|
| Hub running on EKS with Backstage UI accessible | Verify | `ui.aegis-platform.tech` or port-forward |
| At least 1 spoke cluster registered and heartbeating | Verify | Check Clusters page in UI shows "Ready" |
| Workspace pod launches successfully | Verify | Test submit → RUNNING before demo |
| Sovran extension installed in VS Code Insiders | Verify | Built and installed VSIX |
| Keycloak has a demo user account | Verify | User can sign in via OIDC |
| Project with budget and queue configured | Verify | Workloads page shows available project |

---

## The 15-Minute Demo

### Act 1: The Problem (2 minutes)

**Talk, don't show.** Set up the pain:

> "Today, standing up a CUI-compliant ML environment takes 3-6 months. You're assembling Transit Gateway, Network Firewall, AWS Workspaces, Microsoft AD, and Kubeflow across 4+ AWS accounts — just so a data scientist can train a model on a GPU.
>
> And the experience is terrible. Your engineers RDP into a Workspaces VM, open a browser, navigate to Kubeflow, and use Jupyter notebooks on a remote desktop. Slow, laggy, no real debugging. Every session starts with a 5-10 minute VM boot.
>
> We built this architecture ourselves. We lived it. And we thought — why does it take 6 months and 5 engineers to give a data scientist a GPU?"

---

### Act 2: The Platform (3 minutes)

**Show the Backstage UI.** Walk through quickly:

**Clusters page:**
> "Here's our control plane. This spoke cluster registered itself automatically — zero firewall rules opened. The spoke calls outbound to the hub. It's been heartbeating every 10 seconds reporting its health and available GPU capacity."

**Projects page:**
> "Each project gets its own isolated namespace, budget, and queue. Project A can't see Project B's workloads. Budgets enforce cost limits — hard mode rejects workloads when the budget is exceeded."

**Workloads page:**
> "This is where engineers launch workspaces. They pick a GPU size — we call them flavors — and the platform handles the rest."

---

### Act 3: The Money Shot — Launch a Workspace (5 minutes)

This is where you win or lose the demo.

1. **Submit a workspace** — Click "Launch Workspace" in the UI. Pick a flavor (GPU size). Submit.
2. **Show the status transition** — Watch it go from QUEUED → PLACED → RUNNING (should take under a minute)
3. **Open VS Code** with Sovran installed
4. **Click "Connect"** — show the workspace appear in the sidebar, click it
5. **30 seconds later** — you're in a native VS Code window connected to the workspace pod

**In the terminal, run:**
```bash
nvidia-smi          # Show GPU is attached (if GPU nodes available)
python --version    # Show Python is ready
pip list | head -20 # Show pre-installed packages
```

**Key line:**
> "That was 30 seconds. No VM boot. No browser. No Kubeflow navigation. The data scientist picks a GPU size, clicks Launch, and they're in VS Code on that GPU. Full terminal, full debugging, port forwarding — everything they're used to locally, but on GPU hardware."

---

### Act 4: Security (3 minutes)

**While still connected to the workspace:**

> "Let me show you what's happening underneath."

**Authentication:**
> "Every user authenticates through Keycloak with OIDC and MFA. CAC/PIV ready. The connection to this workspace uses a single-use token — 5-minute TTL, can't be replayed. If someone intercepts it, it's already expired or already used."

**Audit trail:**
> "Every action is logged — who launched what workspace, when they connected, from which IP, what they did. Structured JSON, SIEM-compatible."

**Network security:**
> "No inbound firewall rules were opened on the spoke network. The spoke initiates all connections outbound. If you wanted to deploy this at 5 different sites, you'd install the spoke Helm chart 5 times. No network change requests, no firewall approvals, no months of waiting."

**Endpoint protection (if showing secure mode):**
> "In secure mode, all VS Code data lives on an encrypted RAM disk. When the session ends, the RAM disk is destroyed. No CUI persists on the engineer's workstation."

---

### Act 5: The Close — By Persona (2 minutes)

> **"What you just saw deploys in a week.**
>
> **Your ML engineers and data scientists** get a native VS Code IDE on GPU hardware — not a remote desktop, not browser Jupyter. They connect in 30 seconds. They bring their own tools — PyTorch, TensorFlow, MLflow — whatever they already use. They pick a GPU size, click Launch, and start working. Custom images let teams pre-install their specific toolchains.
>
> **Your platform engineers** get a GPU workspace platform that deploys from Helm charts. No Transit Gateway, no Network Firewall, no Workspaces VMs to manage. Need a GPU cluster? Aegis provisions one through the UI. Adding a new site is a Helm install — no network team involvement.
>
> **Your security and compliance teams** get OIDC with MFA and PIV/CAC, fail-closed RBAC, single-use session tokens, structured audit logging on every action, and namespace isolation between projects. Zero inbound ports on spoke networks. Endpoint protection with RAM-disk isolation for CUI. Every security control is in the code path, not bolted on.
>
> **Your program office** gets a platform that lives inside your ATO boundary — no separate FedRAMP authorization needed. The same Helm charts deploy to dev, staging, and your high-side. One week to first GPU, not six months.
>
> What does your current GPU infrastructure setup timeline look like?"

---

## FAQ — Prepared Answers

### "Where's MLflow / TensorBoard / Kubeflow Pipelines?"

> "Aegis is the GPU workspace platform — scheduling, security, budgets, compliance. Your ML tools run inside the workspace. Engineers `pip install mlflow` or use a custom image with everything pre-installed. If you have an MLflow tracking server, the workspace connects to it. We don't replace your ML tools — we replace the 6 months of infrastructure work to make them accessible securely."

### "Can engineers customize the workspace image?"

> "Yes. The workload spec accepts any container image. Your team builds a Docker image with their tools — PyTorch, MLflow, TensorBoard, their specific libraries — pushes it to your registry, and selects it when launching a workspace. Different projects can use different images. We also provide base images with common ML stacks."

### "What about scheduling? How do you handle GPU contention?"

> "When an engineer picks a GPU size and launches a workspace, Aegis finds a cluster with that GPU available and schedules the workspace there. If you have 20 engineers competing for GPUs, Kueue handles fair queuing — no one team hogs all the resources. Budget enforcement adds a financial limit on top. The engineer doesn't think about any of this — they just pick a GPU size and get a workspace."

### "Do we need multiple clusters?"

> "No. One cluster with GPU nodes handles all your workspaces. Multi-cluster is there when you grow — multiple sites, different regions, dev vs prod separation. But a single cluster is the normal starting point."

### "How do engineers get their data in and out?"

> "The workspace pod can mount persistent storage — your datasets survive workspace restarts. Models and artifacts push to your existing registries (ECR, S3, etc.). The workspace has full network access within its namespace for connecting to databases, object stores, and ML tracking servers."

### "Is this FIPS compliant?"

> "The architecture is designed for FIPS — standard algorithms, configurable cipher suites. The remaining work is switching to FIPS-validated implementations (BoringCrypto for Go, UBI9 FIPS base images), which is a build configuration change, not an architecture change. Concrete plan, 2-4 weeks of engineering."

### "How is this different from just running Kubeflow?"

> "Kubeflow is 20+ components — Istio, KNative, Argo, various operators — that take 4-8 weeks to deploy and require dedicated platform engineers to maintain. Many DoD teams have tried and abandoned full Kubeflow deployments. Aegis gives you the GPU workspace access you actually need, with security and compliance built in, deployable in a week."

### "Can this run on our high-side / in a SCIF?"

> "Yes. Everything runs inside the air-gapped network — hub, spoke, user workstations. No internet dependency at runtime. Updates delivered as container images via content drop. The engineer sits at their workstation inside the facility, opens VS Code, and connects to a workspace over the internal network. Same workflow, same platform, same Helm charts — just air-gapped."
