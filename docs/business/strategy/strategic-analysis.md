# Aegis Platform: Strategic Positioning Analysis

## Context

Carlos wants a strategic analysis of how Aegis Platform fits within the AI industry ecosystem described in a video about Perplexity Computer's launch and the "middleware trap." The video outlines how non-hyperscaler companies must find durable structural positions or risk being squeezed between model providers and hyperscaler infrastructure. Carlos wants to evaluate whether Aegis is correctly positioned for DoD/government/AI markets and whether it's doing what it should, including a growth path toward full MLOps.

**Critical framing correction:** Aegis is self-hosted software deployed INSIDE the customer's own ATO'd environment. The customer owns the ATO, not Aegis. This means:
- No FedRAMP authorization needed (Aegis is not a SaaS/cloud service)
- Aegis needs to be STIG-compatible and FIPS-capable
- Two deployment models: (1) customer's AWS GovCloud, (2) customer's bare metal with RKE2/k3s (Platform One)
- This is analogous to how Platform One's Big Bang, Red Hat OpenShift, or Rancher operate in DoD

---

## 1. Where Aegis Sits in the Video's AI Stack

The video describes: Model providers (bottom) -> Orchestration/middleware (middle) -> Distribution (top) -> Cloud/infra (hovering).

**Aegis is NOT in any of these layers.** Aegis is GPU infrastructure management software -- it manages the physical compute that organizations use to train and run models. It doesn't consume model APIs, doesn't orchestrate AI models, doesn't distribute AI to users. It schedules GPU workloads across Kubernetes clusters while enforcing security and compliance boundaries.

**Aegis is not in the middleware trap.** The trap applies to companies building on models they don't control and serving customers that model providers sell to directly. Aegis has no dependency on any model provider. Its value -- "schedule GPU workloads across regulated multi-cluster environments" -- is not something OpenAI, Anthropic, or Google are building or interested in building. Model providers sell API access; they don't manage classified GPU infrastructure for DoD programs.

---

## 2. Which Durable Positions Aegis Occupies

The video identifies four structural positions that survive the hyperscaler squeeze:

### Position 4 (STRONGEST): Trust and Verification Layer

> "The gap between agents doing real work and we can prove what agents did is wide and growing."

**Aegis is already building this for GPU compute.** The audit event system captures who submitted what workload, to which cluster, when, from where, with what outcome. Budget enforcement rejects workloads exceeding program limits. Policy domains restrict workloads by region and data classification. Session management tracks every VS Code connection with subject, duration, and token lifecycle.

As AI agents become autonomous -- launching their own training runs, requesting GPU resources, deploying models on classified data -- the question "who authorized this, what data was accessed, which compliance boundary contained it, and what did it cost?" becomes existential for defense customers. **Aegis already answers these questions for human-initiated workloads. Extending to agent-initiated workloads is the natural evolution.**

This position is reinforced by the self-hosted deployment model: the trust and verification layer runs INSIDE the customer's security boundary, not in a third-party cloud. For DoD customers, this is non-negotiable.

### Position 2 (STRONG): Deep Workflow Integration

For a DoD program running Aegis, ripping it out means rebuilding: identity integration (Keycloak/OIDC), network security (outbound-only spokes, mTLS), compliance evidence collection, budget governance, and the developer experience (VS Code via Sovran). The more institutional knowledge gets encoded in policy domains, placement rules, and compliance configs, the higher the switching cost.

The VS Code connection via Sovran is particularly sticky -- once developers are trained on local VS Code connecting to remote GPU workspaces instead of browser-through-VM, they won't go back.

### Position 3 (MODERATE): Context Advantage

Aegis accumulates operational context about GPU workload patterns, cluster health, cost structures, and compliance state. This context is moderately defensible -- it's the customer's operational telemetry, not proprietary data, but the institutional knowledge encoded in placement policies and budget rules creates switching costs.

---

## 3. The Kubeflow Comparison -- Why It Matters

**Kubeflow in classified environments today:** Browser-based Jupyter through a VM in a locked-down network. Developer sits in SCIF, uses hardened thin client, VPNs into cluster, uses browser IDE. Poor developer experience, no local extensions, high latency.

**Aegis + Sovran:** VS Code runs locally with full extension ecosystem. Connects through WebSocket tunnels via Aegis proxy to remote GPU workspace pods. Only keystrokes and screen updates traverse the tunnel -- no data egress. JWT single-use tokens, mTLS, SC-10 session timeouts, full audit trail.

**Security is actually improved, not weakened.** The outbound-only spoke architecture means no inbound ports on the classified cluster. Every session is logged with subject, duration, and bytes transferred. This is more auditable than browser-through-VPN.

**Critical nuance:** Aegis is NOT a Kubeflow replacement in the MLOps sense. It replaces HOW you access and manage GPU compute, not the ML tooling. Kubeflow, MLflow, or W&B can run ON TOP of Aegis-managed workspaces. The correct positioning: "Aegis replaces the compute access and scheduling layer in classified environments. Bring your own ML tools."

---

## 4. The Regulatory Moat -- What the Video Misses

The video's framework is built for commercial markets. It significantly underweights regulatory moats:

### Self-Hosted Software Inside Customer ATO = Fastest Path to DoD

Aegis doesn't need its own FedRAMP authorization. It deploys inside the customer's already-authorized environment. The customer's ATO covers the deployment. This is the same model as:
- Platform One's Big Bang (DoD's preferred software factory)
- Red Hat OpenShift (used across DoD)
- Rancher/RKE2 (Platform One's default K8s distribution)

What Aegis DOES need:
- **STIG compatibility** -- Hardened container images, security contexts, documented STIG checklist
- **FIPS-capable cryptography** -- BoringCrypto for Go services, FIPS OpenSSL for Keycloak
- **Platform One / Big Bang integration** -- Helm charts that deploy cleanly alongside Big Bang's Istio service mesh, Kiali, monitoring stack
- **RKE2/k3s support** -- The codebase is currently EKS-focused; bare metal with RKE2 is the other deployment model

### Why This Position Is Structurally Defensible

**Hyperscalers can't easily serve this market:**
- AWS GovCloud sells infrastructure, not compliance-focused GPU control planes
- Building a DoD-specific GPU scheduler with audit logging and policy domains is too small a TAM for AWS/Azure/GCP to justify
- The hyperscalers are Aegis's infrastructure providers, not competitors

**Procurement lock-in:** Government procurement cycles are 12-36 months. Once deployed and integrated into a program's workflow, switching requires new procurement action, new security review, new ATO amendment, new budget authorization.

**Personnel moat:** Selling to DoD/IC often requires cleared staff. Building a cleared team takes years and creates barriers that commercial competitors can't easily cross.

**Integration depth moat:** Every DoD program has unique compliance requirements, network topologies, and approval workflows. The more Aegis is customized for a specific program, the harder it is to rip out.

---

## 5. Gap Between Architecture and Target Deployment Models

### Current state: Very AWS EKS-focused

The codebase has deep AWS dependencies:
- Pulumi-based EKS provisioning (`services/platform-api/internal/provisioning/pulumi/aws/`)
- EKS token authentication (`services/platform-api/cmd/aegis-eks-token/`)
- AWS NLB for load balancing
- ECR for container images
- Route53 for DNS
- RDS for PostgreSQL

### Target Model 1: Customer's AWS GovCloud -- MOSTLY READY

The existing EKS automation largely works. Gaps:
- Needs FIPS-enabled container images
- Needs STIG-hardened base images (UBI9 FIPS or Iron Bank equivalents)
- AWS GovCloud API endpoints differ slightly from commercial AWS
- IAM policies need GovCloud partition (`arn:aws-us-gov:` vs `arn:aws:`)

### Target Model 2: Customer's Bare Metal with RKE2/k3s -- SIGNIFICANT GAPS

| Capability | Current (EKS) | Needed (RKE2/k3s) |
|------------|---------------|-------------------|
| K8s provisioning | Pulumi creates EKS clusters | Not applicable -- customer provides cluster |
| Load balancing | AWS NLB | MetalLB or customer's F5/HAProxy |
| Container registry | ECR | Harbor (Platform One standard) or customer registry |
| Database | RDS PostgreSQL | Self-hosted PostgreSQL or customer's DB |
| DNS | Route53 | CoreDNS internal only, or customer DNS |
| Certificate management | cert-manager + step-ca | Same (this works on any K8s) |
| K8s auth | EKS OIDC, EKS token | Standard kubeconfig / service account tokens |
| Storage | EBS CSI driver | Longhorn (RKE2 default) or customer storage |

**The good news:** The core architecture (hub-spoke, gRPC, mTLS, Kueue, audit logging) is K8s-native and should work on any conformant cluster. The gaps are in the infrastructure automation and cloud-specific integrations, not the core product.

**Key work needed for RKE2/k3s:**
1. Abstract away AWS-specific provisioning (or accept `agent_only` import for bare metal)
2. Test Helm charts on RKE2 (Istio sidecar injection compatibility with Big Bang)
3. Support Harbor as container registry
4. Support self-hosted PostgreSQL (already exists as in-memory fallback; needs tested PostgreSQL-on-K8s path)
5. Remove AWS NLB assumptions from spoke proxy service definitions

---

## 6. Growth Path

### Phase 1 (Now): GPU Control Plane for DoD -- Nail the Beachhead

What exists: workload scheduling, multi-cluster placement, workspace access, budget enforcement, audit logging. This is the product. Harden it for RKE2 and GovCloud. Get the first customer.

For ML tooling (experiment tracking, model registry, pipelines): **integrate, don't build.** Pre-integrate MLflow into Aegis-managed workspaces as a Helm addon. "Aegis + MLflow on GovCloud/RKE2" where MLflow runs inside the compliance boundary. This creates value without building from scratch.

### Phase 2 (Next): Expand Within the ML/AI Workload Domain

Stay in the lane of GPU/ML workload management but go deeper:
- Better training job orchestration (distributed training, hyperparameter sweeps)
- Model deployment pipelines (promote from training cluster to serving cluster)
- Integrated experiment tracking (if MLflow integration proves insufficient)
- Multi-cluster inference serving with compliance boundaries

This keeps Aegis focused on what it uniquely solves (multi-cluster GPU scheduling with compliance) while expanding value within that domain.

### Phase 3 (Later): Regulated Enterprise Expansion

Same product, different market. Finance, healthcare, and defense contractors all need multi-cluster GPU scheduling with audit, budgets, and compliance. The product doesn't change -- the sales channel does.

**Do not build Phase 2 until Phase 1 has a paying customer.** Revenue validates the beachhead before expansion.

---

## 7. Strategic Verdict: Is Aegis Doing What It Should?

### What Aegis Is Doing Right

1. **Architecture is genuinely designed for the target market.** Hub-and-spoke with outbound-only agents, mTLS, session management, Kueue integration, policy domain enforcement. Real security engineering, not compliance theater.

2. **VS Code remote workspace via Sovran is a genuine differentiator.** Browser-in-VM is the status quo. This is better for developers AND more auditable.

3. **Self-hosted deployment model is the right choice for DoD.** No FedRAMP overhead. Fits into customer's existing ATO. Fastest path to market.

4. **Multi-cluster GPU scheduling with policy enforcement fills a real gap.** Nobody else does this for regulated environments.

### What Needs to Change (Priority Order)

**P1: FIPS cryptography (2-4 weeks, $0 cost).**
Rebuild Go services with `GOEXPERIMENT=boringcrypto`, use FIPS-validated base images (UBI9 FIPS or Iron Bank), configure Keycloak FIPS mode. Without FIPS, nothing in the DoD space is credible. The gap analysis already documents exact steps.

**P2: RKE2/k3s compatibility (the bare metal deployment model).**
The current codebase is deeply AWS-focused. For the bare metal model to be real:
- Test Helm charts on RKE2 (especially with Istio/Big Bang service mesh)
- Support `agent_only` cluster import as the primary path (customer provides cluster, Aegis installs agent)
- Abstract cloud-specific LB/registry/DNS assumptions into configurable Helm values
- Test with self-hosted PostgreSQL on K8s

**P3: Finish SOC 2 audit.**
It's nearly done. Complete it. Even though DoD customers may not require SOC 2, it demonstrates security maturity and covers enterprise/commercial customers in the pipeline.

**P4: Platform One ecosystem integration.**
Big Bang is the DoD's preferred software factory. Being a "Big Bang addon" (Helm chart that drops into a Big Bang deployment with Istio, monitoring, etc.) is the fastest distribution channel into DoD programs. This is the government equivalent of being in an app store.

**P5: Deepen ML workload capabilities.**
Once the core is production-solid, expand within the domain: distributed training orchestration, model deployment pipelines across clusters, MLflow integration as a Helm addon. Stay in the GPU/ML workload lane -- go deeper, not wider.

### What Aegis Should Stop Doing

1. **Stop claiming capabilities that don't exist in code.** Defense buyers verify. Trust > marketing.
2. **Stop building toward FedRAMP.** The self-hosted model eliminates this requirement entirely. Redirect compliance effort toward STIG/FIPS.
3. **Stop building standalone MLOps features.** Integrate MLflow as a Helm addon. Don't reinvent experiment tracking.
4. **Stop spreading thin across cloud providers.** AWS GovCloud + RKE2 bare metal. That's it for now.

---

## 8. Business Model: Where's the Money?

### Fully Proprietary -- No Open Source

The entire platform is proprietary. The target market (DoD programs, defense primes) discovers tools through Platform One evaluations, DIU scouts, prime contractor partnerships, and conference demos -- not GitHub. The compliance features ARE the product, so there's no meaningful free tier. Fork risk from defense primes is eliminated.

### Pricing Models

**Option A: Per-GPU-node annual subscription (RECOMMENDED)**

| Tier | Includes | Indicative Price |
|------|----------|-----------------|
| **Standard** | Multi-cluster scheduling, workspace access, basic audit, mTLS, Sovran | $500-1,000/GPU-node/year |
| **Enterprise** | + FIPS, STIG profiles, budget enforcement, FinOps, SSO/OIDC, advanced audit, OSCAL export, policy domains | $2,000-5,000/GPU-node/year |
| **Mission Critical** | + Forward-deployed support, custom compliance configs, SLA, priority patches, training | $5,000-10,000/GPU-node/year |

Why per-GPU-node works:
- Aligns with value: more GPUs managed = more value = more revenue
- Easy for government to budget (they know how many GPU nodes they have)
- Scales naturally with customer's infrastructure growth
- A 100-node GPU cluster on Enterprise tier = $200K-500K/year -- meaningful revenue

**Option B: Per-deployment annual license**

| Tier | Includes | Indicative Price |
|------|----------|-----------------|
| **Single Cluster** | Hub + 1 spoke | $75K-150K/year |
| **Multi-Cluster** (up to 5 spokes) | Hub + 5 spokes | $250K-500K/year |
| **Enterprise** (unlimited) | Hub + unlimited spokes + support | $500K-1M/year |

Why per-deployment works:
- Simpler for government budgeting (fixed annual cost)
- Matches how government buys software (annual license + support contract)
- Doesn't require metering infrastructure on customer's air-gapped cluster

**Recommendation: Start with Option B (per-deployment).** It's simpler to sell, easier for government to budget, and doesn't require metering on air-gapped systems. Move to per-GPU-node pricing when you have enough customers to justify the metering complexity.

### Revenue Channels to First Dollar

**Channel 1: SBIR / STTR (6-18 months to revenue)**
- Phase I: $50K-250K to prototype and demonstrate (6 months)
- Phase II: $500K-1.5M to develop and pilot (2 years)
- Phase III: Production contracts (unlimited, competitive)
- Best topics: AI/ML infrastructure, secure computing, DevSecOps
- Target agencies: AFRL (Air Force Research Lab), DARPA, CDAO (Chief Digital & AI Office)
- Pros: Non-dilutive funding, validates with real DoD users, Phase III can become ongoing revenue
- Cons: Long application cycles, competitive, Phase I is small

**Channel 2: DIU (Defense Innovation Unit) (3-12 months)**
- DIU scouts commercial tech for DoD adoption
- Prototype agreements (OTA): $1M-5M, 12-24 months
- Production follow-on: can become ongoing procurement
- Pros: Faster than traditional procurement, designed for commercial tech companies
- Cons: Competitive, requires demonstrable product maturity

**Channel 3: Direct to DoD Programs via Platform One (6-12 months)**
- Platform One evaluates and adopts DevSecOps tools for DoD
- Getting listed in Platform One's "Big Bang" or approved tools list = distribution to every DoD program using P1
- Path: submit for evaluation, demonstrate on RKE2, pass security review
- Pros: Instant distribution to hundreds of DoD programs
- Cons: Requires RKE2 compatibility (currently a gap), P1 has high bar for approval

**Channel 4: Through Defense Primes (6-18 months)**
- Lockheed, Raytheon, Booz Allen, SAIC, etc. need AI infrastructure tools for their programs
- OEM/embed licensing: prime buys Aegis license, includes in their solution to DoD customer
- Pros: Prime handles procurement, customer relationship, and compliance. You just deliver software.
- Cons: Lower margins (prime takes 30-50%), less direct customer relationship, prime may demand exclusivity or source code escrow

**Channel 5: Direct Enterprise Sales (3-12 months)**
- Defense contractors and cleared facilities that aren't DoD but handle classified data
- SpaceX, Palantir, Anduril, Shield AI, and dozens of smaller AI defense companies
- These companies buy commercial software with corporate credit cards or simple procurement
- Pros: Fastest path to revenue, no government procurement overhead
- Cons: Smaller deal size, may be more price-sensitive

**Recommended first revenue path:**

1. **Immediate (0-3 months):** Target Channel 5 -- small/mid defense contractors and cleared facilities. These companies buy fast, have GPU clusters, and need compliance. A $75K-150K annual license is within single-approval budget limits for most.

2. **Parallel (0-6 months):** Apply for relevant SBIRs (Channel 1) and engage DIU (Channel 2). These are non-dilutive and validate the product with real DoD users.

3. **Medium-term (6-12 months):** Pursue Platform One evaluation (Channel 3) once RKE2 compatibility is achieved. This is the distribution multiplier.

4. **Longer-term (12+ months):** Engage defense primes (Channel 4) once you have at least one reference deployment. Primes want proven technology, not prototypes.

### Where Is the TAM?

- ~1,500 DoD programs with software/IT components
- ~200+ that use or plan to use AI/ML workloads
- ~50+ defense contractors with their own GPU clusters for classified work
- ~30+ IC agencies with AI/ML programs

Conservative: 50 customers x $250K/year = $12.5M ARR
Moderate: 150 customers x $400K/year = $60M ARR
Aggressive: 300 customers x $500K/year = $150M ARR

These numbers assume multi-cluster deployments at Enterprise tier. The defense AI market is growing 25-30% annually as CDAO pushes AI adoption across DoD.

### The Money Question Answered

The money is in the compliance premium. Every piece of Aegis's value -- scheduling, workspaces, audit, budgets -- exists because the customer operates in a regulated environment. They can't just spin up SageMaker or Vertex AI. They need something that deploys inside their ATO boundary, on their RKE2 clusters or GovCloud accounts, with FIPS crypto, STIG-hardened images, and full audit trails.

**That compliance requirement is why they pay $250K-$500K/year instead of $0 for open-source Kubernetes tooling.** The compliance tax is the business model. Everything Aegis builds should reinforce the compliance value: deeper audit, better STIG profiles, tighter policy enforcement, easier evidence collection. That's what customers pay for, and it's what competitors can't easily replicate without the same investment in understanding regulated environments.

---

## 9. The Thesis

**Aegis is a self-hosted GPU control plane that deploys inside customer ATO boundaries, providing secure multi-cluster workload scheduling with audit, policy enforcement, and budget governance for DoD/government AI programs. The business model is the compliance premium: customers pay $250K-500K/year because they can't use commercial cloud ML platforms in classified environments, and Aegis is the only thing that gives their data scientists GPU access with the security and audit their compliance officers require.**

It occupies the two most durable positions from the video's framework:
- **Trust/Verification Layer**: Every workload submission, placement decision, budget check, and workspace session is logged with who/what/when/where/outcome. In environments where proving what happened is a compliance requirement, this is not optional infrastructure -- it's mandatory.
- **Deep Workflow Integration**: Identity, network security, compliance evidence, budget governance, and developer toolchain (VS Code via Sovran) -- the more institutional knowledge gets encoded, the higher the switching cost.

The **regulatory moat** reinforces both positions in ways the video's commercial framework doesn't address: self-hosted inside customer ATO (no FedRAMP dependency), STIG/FIPS compatibility, procurement lock-in (12-36 month cycles), personnel moat (cleared staff), and integration depth moat (every DoD program's unique compliance config encoded in Aegis).

The **VS Code/Sovran connection** is the user-facing differentiator that creates switching costs at the developer level. Once ML engineers are trained on local VS Code connecting to remote classified GPU workspaces, reverting to browser-in-VM is not acceptable.

**The gaps are concrete and closable:** (1) FIPS crypto implementation, (2) RKE2/bare metal compatibility. Close those two and the beachhead is real. Growth comes from going deeper in the ML workload domain (distributed training, model pipelines, MLflow integration) and expanding to regulated enterprise (finance, healthcare).
