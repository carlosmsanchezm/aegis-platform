# Aegis Platform — Competitive Analysis

**Purpose:** Product competitive analysis for positioning Aegis in the DoD/regulated AI/ML platform market.

**Last Updated:** 2026-03-14

---

## Market Position

Aegis is a **self-hosted, multi-cluster GPU control plane** for regulated environments. It is not SaaS, not a monolithic application, and not a data science workbench. It is a Kubernetes-native platform that deploys inside the customer's authorization boundary via Helm charts.

**Unique niche:** No other product combines all of these in a single platform:
- Multi-cluster hub-and-spoke GPU scheduling
- Outbound-only spoke connectivity (zero inbound firewall rules on spokes)
- Integrated secure remote workspaces (VS Code Remote via WebSocket tunnels)
- CAC/PIV-native OIDC authentication
- Deployable in days via Helm

---

## Competitive Matrix

| Capability | Aegis | Run:ai (NVIDIA) | Domino Data Lab | HPE MLDE | Kubeflow | Platform One |
|---|---|---|---|---|---|---|
| **On-prem / air-gapped** | Yes | Yes | Yes (Isolated Edition) | Yes | Yes | Yes |
| **Multi-cluster GPU scheduling** | Yes (hub-spoke) | No (single cluster) | No | No | No | No |
| **Secure remote dev (VS Code)** | Yes (WebSocket tunnel) | No | Partial (in-browser) | No | Partial (in-browser) | No |
| **Outbound-only spoke** | Yes | No | No | No | No | N/A |
| **CAC/PIV OIDC** | Yes (Keycloak) | No | Partial | Partial | Partial | Yes (Keycloak) |
| **Authorization (IL-level)** | IL-4/5 ready | No published posture | FedRAMP High | Partial (HPE umbrella) | No (but in Iron Bank) | cATO pathway |
| **Deploy timeline** | Days | 1-2 weeks | 4-8 weeks | 1-2 weeks | 4-8 weeks | 1-2 weeks |
| **FinOps / budget per project** | Yes (HARD/SOFT) | Partial (GPU quota) | No | No | No | No |
| **Fractional GPU / MIG** | No (Kueue-based) | Yes | No | Yes | No | N/A |
| **ML pipelines** | No (bring your own) | No | Yes | No | Yes | No |
| **Experiment tracking** | No (bring your own) | No | Yes | Yes | Limited | No |
| **Estimated annual cost** | $50K-200K | $100K-500K | $200K-1M+ | Bundled with HW | Free (high ops cost) | Free |

---

## Competitor Deep Dives

### Run:ai (NVIDIA)

**What they do:** GPU orchestration platform — fractional GPU sharing, GPU pooling, topology-aware scheduling, MIG support. Acquired by NVIDIA in 2024.

**On-prem / air-gapped:** Yes. Helm-based on Kubernetes. Air-gapped requires manual image mirroring.

**Where Run:ai wins:**
- GPU scheduling sophistication — fractional GPUs, GPU pooling, topology-aware scheduling, MIG support
- NVIDIA's sales force and brand recognition in GPU hardware
- Deeper GPU telemetry and utilization metrics
- Bundled with DGX purchases

**Where Aegis wins:**
- Multi-cluster hub-and-spoke (Run:ai is single-cluster)
- Secure remote workspaces (Sovran VS Code extension)
- Outbound-only spoke connectivity
- CAC/PIV-native OIDC with Keycloak
- Built-in FinOps with per-project/queue budgets
- Faster compliance path for CUI environments

**Compliance:** No FedRAMP. No published NIST 800-171 posture. They rely on deploying within a customer's existing ATO.

**Pricing:** Enterprise licensing per GPU — typically $2,000-5,000/GPU/year.

**DoD presence:** NVIDIA has broad DoD GPU hardware contracts. Run:ai specifically has limited confirmed DoD deployments as of early 2026 — the NVIDIA acquisition is still being integrated into federal sales.

**Key risk:** If NVIDIA bundles Run:ai free with DGX purchases and adds multi-cluster features, it becomes the default GPU orchestrator for customers already buying NVIDIA hardware.

---

### Domino Data Lab

**What they do:** Enterprise data science workbench — Jupyter/RStudio/VS Code workspaces, experiment tracking, model registry, collaboration features.

**On-prem / air-gapped:** Yes. Domino Enterprise deploys on K8s. "Isolated Edition" for disconnected environments.

**Where Domino wins:**
- FedRAMP High authorized (Domino Gov Cloud) — massive procurement advantage
- Full data science workbench features (experiment tracking, model registry, collaboration)
- Established federal sales team with existing DoD contracts
- Broader feature set for data scientist workflows

**Where Aegis wins:**
- Multi-cluster GPU orchestration (Domino is single-cluster)
- Deployment speed (days vs 4-8 weeks with professional services)
- Cost (likely 3-5x cheaper)
- Outbound-only spoke architecture for multi-site deployments
- Purpose-built for GPU orchestration rather than being a full workbench

**Compliance:** FedRAMP High. NIST 800-53 mapped. Dedicated government offering.

**Pricing:** $50,000-200,000+/year for seats. $50,000-100,000+ for professional services deployment.

**DoD presence:** Confirmed federal/DoD customers through FedRAMP authorization.

**Key insight:** Domino is the "safe enterprise choice" for DoD data science teams. Aegis competes best when the customer's primary need is multi-cluster GPU orchestration across sites, not a full data science workbench. Domino's weakness: heavy, expensive, slow to deploy, single-cluster.

---

### HPE Machine Learning Development Environment (Determined AI)

**What they do:** ML training platform — distributed training (PyTorch, TensorFlow), experiment management, hyperparameter search, GPU scheduling with fair-share queues. Acquired by HPE in 2022.

**On-prem / air-gapped:** Yes. Helm chart for K8s or standalone on bare metal. HPE bundles with HPC/AI hardware.

**Where HPE MLDE wins:**
- Deeper distributed training features (fault-tolerant, elastic training, advanced hyperparameter search)
- HPE hardware integration (Cray, ProLiant GPU servers)
- HPE federal sales channel (extensive DoD HPC contracts)
- Bundled with hardware purchases

**Where Aegis wins:**
- Multi-cluster/multi-site control plane (MLDE is single-cluster)
- Secure remote workspaces (Sovran)
- Outbound-only spoke architecture
- Cloud-native K8s focus (not tied to specific hardware)
- FinOps and budget enforcement
- Lighter deployment footprint

**Compliance:** HPE has FedRAMP offerings for other products. MLDE inherits customer's ATO on-prem.

**Pricing:** Bundled with HPE hardware or licensed standalone. Enterprise pricing.

**DoD presence:** HPE has extensive DoD HPC contracts (Cray supercomputers at DoD labs). MLDE likely deployed alongside but specific references are limited.

**Key insight:** HPE MLDE is the closest direct competitor in on-prem GPU orchestration on K8s. The differentiation is Aegis's multi-cluster architecture and secure remote development vs HPE MLDE's deeper ML training primitives. If the customer already has HPE hardware, MLDE may come bundled.

---

### Kubeflow (Open Source)

**What they do:** Open source ML platform on Kubernetes — Jupyter notebooks, training operators (TFJob, PyTorchJob, MPI), Katib hyperparameter tuning, Kubeflow Pipelines for ML workflow DAGs.

**On-prem / air-gapped:** Yes. Runs on any K8s. Air-gapped with image mirroring. Included in Platform One's Iron Bank.

**Where Kubeflow wins:**
- Free (open source)
- Broader ML pipeline capabilities (Kubeflow Pipelines for DAG orchestration)
- Large ecosystem and community
- Already in Iron Bank (pre-validated for DoD)
- Training operators for distributed training

**Where Aegis wins:**
- Dramatically simpler deployment (days vs 4-8 weeks for full Kubeflow)
- Integrated secure remote workspaces (not browser-based code-server)
- Multi-cluster hub-and-spoke (Kubeflow is single-cluster)
- CAC/PIV OIDC, built-in RBAC, audit logging
- Outbound-only connectivity
- Cohesive UX vs Kubeflow's collection of 20+ loosely coupled components
- FinOps and budget enforcement
- Dramatically lower maintenance burden

**Compliance:** None inherent. Customer builds their own. However, Kubeflow components are in Iron Bank.

**Pricing:** Free. Operational cost is high due to deployment and maintenance complexity.

**DoD presence:** Used in some DoD ML programs, often in reduced form (cherry-picked training operators, not the full platform). Many DoD teams have attempted and abandoned full Kubeflow deployments.

**Key insight:** Kubeflow is the "obvious open source choice" but its complexity is legendary. Many DoD teams have failed Kubeflow deployments because the full platform requires assembling Istio, KNative, Argo, and 15+ other components. Aegis's pitch: "You get the GPU orchestration and developer experience you actually need, without the 6-month Kubeflow deployment project."

---

### Palantir Foundry / AIP

**What they do:** Data fusion and decision-support platform. Not primarily a GPU orchestrator — it's a data integration, ontology, and analytics platform with AI capabilities.

**On-prem / air-gapped:** Yes. Deployed in classified environments up to TS/SCI. Palantir Apollo manages deployments.

**Where Palantir wins:** IL-5/IL-6 authorized, FedRAMP High, massive DoD incumbency ($1B+ annual gov revenue), embedded forward-deployed engineers, comprehensive data integration.

**Where Aegis wins:** Cost (orders of magnitude cheaper), deployment speed, focused GPU orchestration, developer experience, self-service (no embedded engineers required).

**Pricing:** $5M-50M+/year. Not in the same market segment.

**Key insight:** Palantir is not a direct competitor — it's a data fusion platform that competes for the same budget dollars. Aegis's pitch: "You don't need a $10M platform to let your ML engineers use GPUs across sites. Aegis does the GPU orchestration piece for a fraction of the cost and deploys in days."

---

## Complementary Tools (Not Competitors)

These tools serve different purposes and should be positioned as part of the Aegis ecosystem, not as competition:

| Tool | Category | Relationship to Aegis |
|------|----------|----------------------|
| **Platform One / Big Bang** | DevSecOps platform | Aegis deploys on top as the GPU/ML layer. Big Bang provides CI/CD, monitoring, service mesh. Aegis provides GPU orchestration and secure workspaces. |
| **W&B / MLflow** | Experiment tracking | Runs inside Aegis workspaces. Aegis orchestrates compute; W&B/MLflow tracks results. |
| **Ray / KubeRay** | Distributed compute | Ray clusters can be managed as Aegis workloads. Aegis provides the multi-cluster control plane; Ray handles distributed training. |
| **Seldon Core** | Model serving | Aegis schedules training; Seldon serves the models. Different lifecycle phases. |
| **Game Warden (Second Front)** | DoD hosting PaaS | Aegis could deploy on Game Warden to inherit its IL-4/5 ATO. Potential partner/channel. |

---

## Strategic Recommendations

### 1. Own the Multi-Cluster GPU Niche

No competitor does multi-cluster hub-and-spoke GPU scheduling with outbound-only spoke connectivity. This is Aegis's unique territory. All messaging should lead with this.

### 2. Get Into Iron Bank

Publishing Aegis containers to Iron Bank (registry1.dso.mil) would be the single highest-impact go-to-market action for DoD adoption. It converts Aegis from "vendor software requiring full ATO review" to "pre-validated component deployable in any Big Bang environment."

### 3. Position as Big Bang Add-On

Aegis fills the gap Big Bang has: no GPU orchestration, no ML workloads, no secure remote development. Position as: "Deploy Aegis on your Big Bang environment to add GPU workload orchestration and secure ML workspaces." This makes Aegis the natural next step for USAF programs already running Big Bang.

### 4. Monitor Run:ai Closely

If NVIDIA integrates Run:ai deeply and adds multi-cluster features, it becomes the default GPU orchestrator for DGX customers. Watch for: multi-cluster announcements, federal compliance certifications, GPU bundling deals.

### 5. Integrate MLOps Tools as Add-Ons

Customers will ask about experiment tracking and model registries. Rather than building these, pre-integrate MLflow and TensorBoard as optional Helm add-ons that run alongside Aegis workloads. Position: "Aegis is the GPU control plane; bring your own MLOps tools — or use our pre-integrated MLflow deployment."

---

## Pricing Strategy & Value Analysis

### What Aegis Displaces (Customer's Cost Without Aegis)

Organizations building CUI-compliant ML environments from infrastructure primitives (AWS Workspaces, Transit Gateway, Network Firewall, Microsoft AD, Kubeflow on EKS) incur the following annual steady-state costs — not counting the 3-6 month initial build:

| Cost Category | Annual Estimate | Basis |
|---|---|---|
| Platform engineering to build & maintain (TGW, NFW, AD, Workspaces, Kubeflow) | $600K-1.2M | 3-5 senior DevOps/platform engineers at $200K+ loaded |
| Kubeflow-specific maintenance | $200K-400K | 1-2 dedicated engineers (Kubeflow is notoriously brittle, 20+ components) |
| Developer productivity loss from VM hop | $300K-600K | 50 engineers × ~7 hrs/month wasted on Workspaces boot, browser nav, latency × $100/hr |
| AWS infrastructure overhead (Workspaces, TGW, NFW, NAT) | $50K-100K | Networking/access layer only — does not include GPU compute costs |
| **Total displaced cost** | **$1.2M-2.3M/year** | Steady-state operational cost |

Initial build cost adds another $500K-1M+ (3-6 months of engineering across 4+ AWS accounts, TGW routing, NFW rules, AD integration, Kubeflow deployment, DNS, and compliance documentation).

### Competitor Pricing Reference

| Competitor | Annual Range | Model |
|---|---|---|
| Run:ai (NVIDIA) | $100K-500K | Per-GPU licensing ($2K-5K/GPU/year) |
| Domino Data Lab | $200K-1M+ | Per-seat + professional services ($50K-100K for deployment) |
| HPE MLDE (Determined AI) | $200K-500K | Standalone license or bundled with HPE hardware |
| Palantir Foundry/AIP | $5M-50M+ | Enterprise contract with embedded forward-deployed engineers |
| Kubeflow | $0 (free) | Open source — but $400K-800K/year in engineering to operate |

### Recommended Aegis Pricing

**Mid-size deployment** (50-200 ML engineers, 3-10 GPU spoke clusters):

| Component | Annual Price |
|---|---|
| Platform license (hub + unlimited spokes + Sovran extension) | $350K-500K |
| Deployment & onboarding (year 1, can amortize or bill separately) | $50K-100K |
| Support & maintenance (SLA, upgrades, quarterly training) | $100K-200K |
| **Total** | **$500K-800K/year** |

**Large deployment** (200+ engineers, 10+ clusters, multiple IL levels): **$800K-1.5M/year**

### Why This Pricing Works

- **2-3x ROI** — Customer saves $1.2M-2.3M/year and pays $500K-800K
- **Undercuts Domino** at the high end while offering multi-cluster GPU scheduling (which Domino can't do)
- **10x cheaper than Palantir** for the GPU orchestration piece
- **Competitive with Run:ai** but includes secure workspaces and compliance posture they lack
- **High margin** — Aegis deploys in days via Helm; support cost per customer is low relative to license revenue

### Pricing Anchor for Customer Conversations

> "It took your team 6 months and 5 engineers to build what Aegis replaces. That's $500K+ just in build cost, then $800K+/year to maintain. We give you a better platform for half that, deployed in a week."

### Contract Vehicle Fit

For DoD contracts, $500K-800K/year lands well within these vehicles:
- **SBIR Phase III** — Sole-source, no ceiling, any agency can award
- **DIU OTA (Other Transaction Authority)** — Rapid prototyping/production, streamlined acquisition
- **GWACs** — SEWP, ITES, or 8(a) STARS III for IT services
- **BPA/IDIQ** — Blanket purchase agreements with task orders per deployment
- **Direct program office funding** — Under simplified acquisition threshold ($250K) per CLIN, or above with competitive justification

At $500K-800K, a program office can often approve without full competitive acquisition if Aegis is on the right vehicle or justified as sole-source (unique multi-cluster GPU orchestration capability).

---

## Threats to Monitor

| Threat | Likelihood | Impact | Mitigation |
|--------|-----------|--------|------------|
| NVIDIA bundles Run:ai free with DGX | Medium | High | Lead with multi-cluster and secure workspaces — capabilities Run:ai lacks |
| Kubeflow community builds easy installer | Low | Medium | Aegis integrates security/compliance that Kubeflow will never have natively |
| Domino drops price for DoD | Low | Medium | Compete on deployment speed and multi-cluster, not features or price |
| HPE MLDE adds multi-cluster | Medium | Medium | Move faster on distributed training features |
| New DoD-focused ML startup | Medium | Medium | Iron Bank inclusion and DoD customer references build switching cost |
