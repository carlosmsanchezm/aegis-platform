# Aegis Platform vs. P1 Centralized Cluster Management (CCM) — Gap Analysis & Roadmap Enhancement

**Purpose:** Map Aegis Platform's current capabilities against the 25 requirements from the USAF Platform One CCM prototype OTA (FA8307-26-9-B005) to identify what the government values in cluster management platforms and what Aegis should add to its roadmap.

**Sources:** OTA Agreement (44pp), VivSoft ENBUILD Proposal Response (17pp), Traceability Matrix (25 requirements), Aegis Platform codebase.

**Last Updated:** 2026-03-14

---

## 1. Executive Summary

### What the CCM Contract Is

Platform One (P1) — USAF's DevSecOps platform serving Iron Bank, Big Bang, and Party Bus value streams — issued an Other Transaction Agreement (OTA) for a **Centralized Cluster Management** prototype. The contract:

- **Value:** $254,800 (prototype phase)
- **Period:** 120 days
- **Awarded to:** VivSoft Technologies (product: ENBUILD Accelerator)
- **Goal:** Centralize the configuration, deployment, management, and monitoring of P1's numerous Kubernetes clusters across AWS, Azure, GCP, and on-prem — replacing the current decentralized, team-by-team approach

### Why This Matters for Aegis

This contract validates **exactly what the DoD is willing to pay for** in cluster management. The 25 requirements represent what P1's engineering leadership considers essential. Even though Aegis is positioned as a GPU workload control plane rather than a pure cluster lifecycle tool, many requirements overlap directly with what Aegis does or could do.

### Coverage Snapshot

| Status | Count | Percentage |
|--------|-------|-----------|
| **STRONG** (Aegis does this well today) | 6 | 24% |
| **PARTIAL** (building blocks exist, gaps remain) | 8 | 32% |
| **GAP** (Aegis doesn't do this) | 6 | 24% |
| **N/A** (process/contract, not platform code) | 5 | 20% |

**Bottom line:** Aegis covers ~56% of the technical requirements at STRONG or PARTIAL level. The major gaps are in **multi-cloud provisioning** (Azure/GCP), **Big Bang lifecycle management**, **SIEM export**, and **SLA frameworks**. These are high-value additions that would make Aegis competitive for CCM-like contracts.

---

## 2. Requirement-by-Requirement Mapping

### SOO 3.1 — Initial Deployment

#### CCM-01: Deploy all solution tools
**SOO 3.1.1** — "Demonstrate deployment of all tools associated with the solution."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | PARTIAL |
| **What Aegis Has** | Helm charts for hub (`aegis-services`) and spoke (`aegis-spoke`). Terraform for EKS infrastructure. `scripts/deploy.sh` for automated deployment. Backstage UI for management. |
| **What's Missing** | No self-service "click to deploy the entire platform" UI. No catalog-driven bootstrapping. Deployment requires kubectl + Helm CLI knowledge. |
| **What ENBUILD Does** | Self-service catalog with Docker Compose bootstrap and Helm chart for cluster deployment. CLI tool (`enbuild create`). |
| **Recommendation** | **ENHANCE** — Add a deployment wizard or CLI bootstrapper. Low priority since Aegis targets operators who are comfortable with Helm, but a self-service installer raises the product polish bar. |

---

#### CCM-02: Same-account GovCloud Big Bang deployment
**SOO 3.1.2** — "Demonstrate deployment of Big Bang Core and Addon Components in the same AWS GovCloud account."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | GAP |
| **What Aegis Has** | Aegis deploys its own services via Helm but does not orchestrate Big Bang deployment. |
| **What's Missing** | No Big Bang awareness. No ability to deploy, configure, or manage Big Bang Core or Addons. |
| **What ENBUILD Does** | Pre-loaded Big Bang catalog items. FluxCD-driven deployment. Helm chart synchronization for Big Bang components. |
| **Recommendation** | **ADD (Tier 1)** — Big Bang integration is table stakes for P1 contracts. Aegis should be able to deploy Big Bang as a "stack" alongside its own services. This could be implemented as a catalog item in the UI or as a Helm umbrella chart. |

---

#### CCM-03: Cross-account GovCloud Big Bang deployment
**SOO 3.1.2** — "Demonstrate deployment within a different AWS GovCloud account."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | PARTIAL |
| **What Aegis Has** | Cross-account AWS role assumption for spoke clusters (`ProjectAwsCredentials` with `assume_role_arn`). Cluster import via assume-role method. |
| **What's Missing** | No Big Bang deployment orchestration in cross-account scenarios. The cross-account capability exists for workload scheduling, not cluster provisioning. |
| **Recommendation** | **ADD (Tier 1)** — Extend existing cross-account credential infrastructure to support full cluster+Big Bang provisioning, not just workload placement. |

---

#### CCM-04: Separate Azure account Big Bang deployment
**SOO 3.1.2** — "Demonstrate deployment within an Azure account separate from the proposed management solution."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | GAP |
| **What Aegis Has** | AWS-only. No Azure provider support. |
| **What's Missing** | Azure cluster provisioning, AKS support, Azure identity/credential management. |
| **What ENBUILD Does** | Azure landing zone templates, AKS catalog items. |
| **Recommendation** | **ADD (Tier 1)** — Multi-cloud is a hard requirement in every DoD cluster management RFP. Azure AKS support (at minimum cluster import, ideally provisioning) should be on the roadmap. |

---

#### CCM-05: Managed distro + third-party distro support
**SOO 3.1.2** — "Demonstrate use of the cloud service providers' managed Kubernetes distribution (AWS/EKS) as well as a third-party Kubernetes distribution."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | PARTIAL |
| **What Aegis Has** | EKS fully supported and tested. Helm charts are distribution-agnostic (pure Kubernetes). Cluster import supports kubeconfig from any distro. |
| **What's Missing** | RKE2/k3s not yet tested. No dedicated provisioning templates for non-EKS distros. |
| **What ENBUILD Does** | Catalog items for EKS, AKS, GKE, RKE2, OKE. |
| **Recommendation** | **ENHANCE** — Test and document RKE2 and k3s support (already on existing roadmap). Add Terraform modules for RKE2 provisioning. |

---

### SOO 3.1.3 — Cloud Cost Oversight

#### CCM-06: Cloud-spend oversight and alerts
**SOO 3.1.3** — "AWS CloudWatch metrics configured to monitor cloud spend in real-time. Alerts at 25%, 50%, 75%, 90% thresholds with SMS notifications."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | STRONG |
| **What Aegis Has** | Budget enforcement with HARD (reject workload) and SOFT (warn) modes. Per-project, per-queue budget limits with transactional locking. Prometheus metrics (`aegis_budget_actual_usd`, `aegis_budget_denied_total`). FinOps dashboards in UI. |
| **What's Missing** | No CloudWatch billing alarm integration. No SMS notifications. Budgets are workload-level, not cloud-infrastructure-level (no EC2/EBS cost tracking). |
| **What ENBUILD Does** | Kubecost URL placeholder. CloudWatch flow-log plumbing. Prometheus-driven cost monitoring. |
| **Recommendation** | **ENHANCE** — Add CloudWatch billing alarm integration (Terraform module) and threshold-based alerting (email/SMS/webhook). Aegis's budget enforcement is already superior to what ENBUILD offers — extending it to cloud infrastructure costs would be a strong differentiator. |

---

### SOO 3.2 — Application Integration

#### CCM-07: Non-Big Bang app integrated with mesh, logging, monitoring, SSO
**SOO 3.2.1** — "Deploy a non-Big Bang, containerized application alongside Big Bang and demonstrate automated integration to the service mesh, logging and monitoring, SSO capabilities."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | PARTIAL |
| **What Aegis Has** | Deploys containerized workloads (workspace and training pods) to spoke clusters. Keycloak SSO. Structured logging. Prometheus monitoring. |
| **What's Missing** | No service mesh integration (Istio VirtualService, DestinationRule). No automated Loki/Prometheus sidecar injection. No Big Bang awareness. |
| **Recommendation** | **ADD (Tier 2)** — Add service mesh integration as an optional spoke configuration. When deploying workloads to a Big Bang cluster, auto-generate Istio resources and logging sidecars. |

---

### SOO 3.3 — Manage Existing Resources

#### CCM-08: Ingest existing clusters and cloud resources
**SOO 3.3.1** — "Ingest existing P1 kubernetes clusters and cloud resources into the solution to establish management."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | STRONG |
| **What Aegis Has** | `ImportCluster` RPC stores auth material (kubeconfig, assume-role ARN, or endpoint+CA) in PostgreSQL + K8s Secrets. Spoke agent auto-registers via heartbeat. Project-to-cluster binding with auto-derive from cluster ID format. Multi-tenant isolation. |
| **What's Missing** | Import doesn't validate connectivity — cluster stays `pending_agent` until spoke agent registers and heartbeats. No cloud resource discovery beyond clusters (VPCs, IAM, SGs not tracked). No "brownfield" baseline analysis of existing cluster state. |
| **What ENBUILD Does** | Repo URL lookup, stack name matching. Focused on linking clusters to GitOps repos for lifecycle management (different purpose than Aegis's workload scheduling import). RTM notes ENBUILD's brownfield import is also minimal. |
| **Recommendation** | **ENHANCE** — Add connectivity validation at import time (test the kubeconfig/role before storing). Add cloud resource discovery as optional enrichment. Both platforms have room to grow here. |

---

### SOO 3.4 — Updates and Upgrades

#### CCM-09: Upgrade Big Bang clusters
**SOO 3.4.1** — "Perform updates and upgrades to currently deployed Big Bang clusters including security patches, performance improvements, and functional upgrades."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | GAP |
| **What Aegis Has** | Helm-based rollback for Aegis's own services. No Big Bang lifecycle management. |
| **What's Missing** | No Big Bang version tracking, upgrade orchestration, or rollback capability. |
| **What ENBUILD Does** | FluxCD Helm chart synchronization. Version bump via Git. Blue-green/canary upgrade strategies. |
| **Recommendation** | **ADD (Tier 1)** — If targeting P1 contracts, Big Bang lifecycle management is mandatory. Implement a "cluster stack" abstraction that tracks deployed Big Bang version and enables Git-driven upgrades. |

---

#### CCM-10: Upgrade infrastructure and Kubernetes distributions
**SOO 3.4.2** — "Perform updates and upgrades to underlying cloud infrastructure and kubernetes distributions."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | PARTIAL |
| **What Aegis Has** | Pulumi-based EKS provisioning (can update cluster version). Terraform IaC for infrastructure. |
| **What's Missing** | No automated upgrade orchestration (node pool drain, staged rollout, validation, rollback). No UI-driven upgrade workflow. |
| **What ENBUILD Does** | Terraform provider/module version updates via catalog. Blue-green node pool strategies. |
| **Recommendation** | **ADD (Tier 2)** — Implement automated K8s version upgrade workflow: drain → upgrade control plane → upgrade node pools → validate → rollback on failure. |

---

### SOO 3.5 — Logging and Monitoring

#### CCM-11: Ship logs to P1 SIEM
**SOO 3.5.1** — "Aggregate and ship logs generated by the solution and managed resources to the P1 Security Information and Event Management (SIEM) solution."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | PARTIAL |
| **What Aegis Has** | Structured JSON audit logs to stdout. Prometheus metrics endpoints. Log format is SIEM-compatible (structured JSON with AU-3 fields). |
| **What's Missing** | No built-in log exporter/forwarder (Fluent Bit, Fluentd). No SIEM destination configuration in Helm chart. Customer must configure their own log pipeline. |
| **What ENBUILD Does** | Full Loki stack (Loki + Promtail + Grafana) deployed via Big Bang. Configurable exporters to P1 SIEM. |
| **Recommendation** | **ADD (Tier 1)** — Bundle Fluent Bit as an optional sidecar or DaemonSet in the Helm chart with configurable SIEM destinations (Splunk HEC, ELK, CloudWatch, Loki). This is a common DoD requirement and relatively low-effort. |

---

#### CCM-12: Health and liveliness monitoring
**SOO 3.5.2** — "Demonstrate health and liveliness monitoring capabilities of the solution and managed resources."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | STRONG |
| **What Aegis Has** | K8s-agent heartbeat (10s interval, 45s stale detection). Two-phase stale cluster detection with grace period. Prometheus metrics for workload placement, budget, auth failures. Health endpoints (`/readyz`). Kubernetes liveness/readiness probes on all pods. |
| **What's Missing** | No Grafana dashboards bundled. No fleet-wide health dashboard in the UI (cluster status is shown but not deep health metrics). |
| **Recommendation** | **ENHANCE** — Add pre-built Grafana dashboards (or integrate them into the Backstage UI) showing fleet health, per-cluster metrics, and alerting status. |

---

#### CCM-13: Holistic single-pane ecosystem view
**SOO 3.5.3** — "Provide a holistic view of the entire cluster ecosystem through a single interface."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | STRONG |
| **What Aegis Has** | Backstage UI with 23+ pages covering clusters, workloads, projects, budgets, metrics, logs, and admin. Cluster list with status, region, capacity. Workload lifecycle view. FinOps dashboards. |
| **What's Missing** | No deep cluster internals view (pods, services, nodes within a spoke). The UI shows Aegis-managed resources, not raw Kubernetes resources. |
| **What ENBUILD Does** | Headlamp integration for K8s resource browsing. Grafana dashboards embedded. |
| **Recommendation** | **ENHANCE** — Add a cluster detail view showing node count, pod status, resource utilization from the spoke. Could be sourced from heartbeat data or a lightweight metrics API on the agent. |

---

### SOO 3.6 — Service Delivery Plan

#### CCM-14: Roles and responsibilities model
**SOO 3.6.1** — "Provide a service delivery model defining roles and responsibility differentiations."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | PARTIAL |
| **What Aegis Has** | RBAC with predefined roles (workspace-admin, etc.). Project-level access boundaries. Keycloak groups/roles. |
| **What's Missing** | No formal operational responsibility model document. RBAC is technical access control, not an org-level service delivery model. |
| **Recommendation** | **PROCESS-ONLY** — This is a documentation deliverable, not a platform feature. Create a service delivery model template as a compliance document. |

---

#### CCM-15: SLAs and metrics
**SOO 3.6.2** — "Describe proposed Service Level Agreements and any other relevant metrics."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | GAP |
| **What Aegis Has** | Prometheus metrics exist (workload queue time, auth failures). Health monitoring exists. No SLA definition, tracking, or reporting framework. |
| **What's Missing** | No SLA definition model. No SLA compliance tracking or breach alerting. No availability/uptime measurement. |
| **What ENBUILD Does** | Proposed SLA targets: <30min cluster provisioning, 99.9% platform uptime, <1hr incident response, <72hr change management. Dashboards for SLA tracking. |
| **Recommendation** | **ADD (Tier 1)** — Implement an SLA framework: define SLA targets per cluster/project, track compliance via Prometheus, surface SLA dashboards in UI, alert on breaches. This is valuable for any enterprise customer, not just P1. |

---

#### CCM-16: External-customer and cost-recovery roadmap
**SOO 3.6.3** — "Provide a notional service delivery model for external customers including cost recovery analysis."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | PARTIAL |
| **What Aegis Has** | Multi-tenancy with per-project namespaces and AWS credentials. Budget enforcement. Project-level isolation. |
| **What's Missing** | No metering/chargeback model. No per-tenant cost attribution beyond budget limits. No tenant onboarding self-service. |
| **What ENBUILD Does** | Basic auth/RBAC scaffolding. Prometheus-based metering. |
| **Recommendation** | **ENHANCE** — Add per-project cost metering and reporting. Aegis already tracks GPU hours and budget — extend to generate chargeback reports per project/tenant. |

---

### SOO 3.7 — Procurement

#### CCM-17: Containerized and Iron Bank-friendly posture
**SOO 3.7.2** — "Containerized solutions should leverage the Iron Bank repository during prototyping and must leverage Iron Bank for production."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | PARTIAL |
| **What Aegis Has** | All services are containerized. Multi-stage Docker builds. Platform-api uses UBI9-minimal (Red Hat). Proxy and k8s-agent use Google distroless. |
| **What's Missing** | No images published to Iron Bank (registry1.dso.mil). Not all images use Iron Bank-approved base images. No SBOM generation. No vulnerability scanning in CI/CD. |
| **What ENBUILD Does** | Backend image published to Iron Bank. Big Bang Flux images swapped to Iron Bank sources. |
| **Recommendation** | **ADD (Tier 1)** — Publish Aegis container images to Iron Bank. Switch all base images to Iron Bank-approved UBI9 or equivalent. Add SBOM generation (Syft/CycloneDX) and vulnerability scanning (Trivy) to CI/CD. This is a prerequisite for any P1 production deployment. |

---

#### CCM-18: Deliver IaC to P1-owned GitLab
**SOO 4.0** — "All associated IaC will be delivered to P1 owned and maintained GitLab source code repository."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | GAP |
| **What Aegis Has** | GitHub-based. CI/CD on GitHub Actions. No GitLab integration. |
| **What's Missing** | No GitLab CI/CD pipeline. No GitLab runner configuration. No ability to push IaC to a customer-owned GitLab. |
| **Recommendation** | **ADD (Tier 1)** — Add GitLab CI/CD pipeline support alongside GitHub Actions. Many DoD organizations (P1, Army, Navy) use GitLab. A `.gitlab-ci.yml` with equivalent build/push/deploy stages is essential. |

---

### Contract/Process Requirements

#### CCM-19: Kickoff, MSRs, Service Delivery Plan, Production Implementation Plan
**SOO 4.0** — "Deliverables: Kick-off Meeting, Monthly Status Reports, Infrastructure as Code, Service Delivery Plan, Production Implementation Plan."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | N/A (Process) |
| **Recommendation** | **PROCESS-ONLY** — These are contract deliverables, not platform features. Create templates for MSRs, Service Delivery Plans, and Production Implementation Plans that can be adapted per contract. |

---

#### CCM-20: P1 zero-trust access and public-release controls
**SOO 6.3/7.0** — "Performer must be able to install software specific to accessing P1 through its zero-trust solution (AppGate)."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | GAP |
| **What Aegis Has** | No AppGate integration. Keycloak handles auth but not network-layer zero-trust. |
| **Recommendation** | **SKIP** — AppGate is P1-specific. Don't build for it unless pursuing a P1 contract directly. However, ensure Aegis works behind generic zero-trust solutions (BeyondCorp, Zscaler, etc.). |

---

#### CCM-21: Provide licensed software
**SOO 3.7.1** — "Offerors are responsible for providing all licensed software."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | N/A (Process) |
| **What Aegis Has** | Open-source Go services. Keycloak (open-source). Backstage (open-source). No proprietary dependencies requiring separate licensing. |
| **Recommendation** | **PROCESS-ONLY** — Maintain a license declaration document. Add SBOM generation to formalize this. |

---

#### CCM-22: Performer-furnished personnel tooling
**SOO 6.1/6.3** — "Performer shall furnish, manage, and provide all hardware and software for their personnel."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | N/A (Process) |
| **Recommendation** | **PROCESS-ONLY** — Standard contract requirement. Not a platform feature. |

---

#### CCM-23: Personnel screening, OPSEC, CUI safeguarding
**SOO 7.0** — "Tier 1 background investigation for CAC, OPSEC coordination, NIST 800-171 for CUI."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | N/A (Process) |
| **What Aegis Has** | Sovran extension has secure mode for CUI handling (RAM disk, FDE verification, ephemeral tokens). |
| **Recommendation** | **PROCESS-ONLY** — Personnel screening is an organizational requirement. The CUI safeguarding in Sovran's secure mode is a platform differentiator worth highlighting. |

---

#### CCM-24: Unified authentication and logging strategy
**SOO 3.2.2** — "Both applications can operate cohesively with unified authentication and logging strategies."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | STRONG |
| **What Aegis Has** | Keycloak OIDC for all components (UI, API, extension). Structured JSON audit logging across all services. Same auth token validated everywhere. Backstage auth-logging middleware captures sign-in/sign-out events. |
| **What's Missing** | Auth logging for managed workloads (inside workspace pods) is not centralized. |
| **Recommendation** | **Current strength.** Aegis's unified auth + logging is one of its best features. Highlight this in proposals. |

---

#### CCM-25: Government data rights
**SOO 5.0** — "Intellectual Property developed under the prototype will be covered under Unlimited Rights."

| Dimension | Assessment |
|-----------|-----------|
| **Aegis Status** | N/A (Process) |
| **Recommendation** | **PROCESS-ONLY** — Legal/IP consideration, not a platform feature. Note: OTA prototypes typically grant Unlimited Data Rights to the government for all work product. Plan IP strategy accordingly. |

---

## 3. Coverage Summary

```
STRONG  (6):  CCM-06 (Cost oversight), CCM-08 (Cluster import), CCM-12 (Health monitoring),
              CCM-13 (Single-pane view), CCM-24 (Unified auth/logging), CCM-14 (RBAC)

PARTIAL (8):  CCM-01 (Deploy tools), CCM-03 (Cross-account), CCM-05 (Multi-distro),
              CCM-07 (App integration), CCM-10 (K8s upgrades), CCM-11 (SIEM logs),
              CCM-16 (Cost recovery), CCM-17 (Iron Bank)

GAP     (6):  CCM-02 (Big Bang same-acct), CCM-04 (Azure), CCM-09 (Big Bang upgrades),
              CCM-15 (SLAs), CCM-18 (GitLab), CCM-20 (AppGate zero-trust)

N/A     (5):  CCM-19 (Deliverables), CCM-21 (Licensing), CCM-22 (Personnel tooling),
              CCM-23 (Personnel/OPSEC), CCM-25 (Data rights)
```

---

## 4. Strategic Recommendations — What to Add to Aegis Roadmap

### Tier 1: Build These (Required for DoD workload orchestration market)

These are needed regardless of whether Aegis expands into cluster lifecycle management. They appear in nearly every DoD RFP and directly enable deployment into regulated environments.

| Enhancement | CCM Req | Effort | Impact | Why |
|-------------|---------|--------|--------|-----|
| **Iron Bank image publishing + SBOM + scanning** | CCM-17 | 2-3 weeks | Critical | Hard blocker. Cannot deploy to any P1 production cluster without Iron Bank images. |
| **FIPS 140-3 cryptography** | (existing roadmap) | 2-4 weeks | Critical | Hard requirement for all DoD customers. Already planned. |
| **SIEM log export (Fluent Bit sidecar)** | CCM-11 | 1-2 weeks | High | Every DoD org requires SIEM integration. Low effort, high signal. |
| **GitLab CI/CD support** | CCM-18 | 1-2 weeks | High | P1, Army, Navy all use GitLab. Essential for DoD market access. |
| **SLA framework (definition + tracking + dashboards)** | CCM-15 | 2-3 weeks | High | Shows operational maturity. Reusable across all customers. |

### Tier 2: Valuable Extensions (Enhance competitiveness, build when ready)

These extend existing Aegis capabilities and make it more attractive to DoD evaluators, but aren't hard blockers for initial contracts.

| Enhancement | CCM Req | Effort | Impact |
|-------------|---------|--------|--------|
| Azure AKS cluster support (import at minimum) | CCM-04/05 | 3-4 weeks | Medium-High |
| Cloud billing alarm integration (CloudWatch thresholds + alerts) | CCM-06 | 1-2 weeks | Medium |
| Cluster detail view in UI (node/pod/resource metrics from spoke) | CCM-13 | 1-2 weeks | Medium |
| Per-project cost metering and chargeback reports | CCM-16 | 2-3 weeks | Medium |
| Service mesh integration (Istio-aware workload deployment) | CCM-07 | 2-3 weeks | Medium |
| Import connectivity validation (test kubeconfig/role at import time) | CCM-08 | 1 week | Medium |

### Tier 3: Cluster Lifecycle Features (Only if expanding scope beyond workload orchestration)

These are central to the CCM contract but outside Aegis's current focus. Build only if strategically deciding to expand into cluster lifecycle management.

| Enhancement | CCM Req | Effort | Impact |
|-------------|---------|--------|--------|
| Big Bang deployment integration | CCM-02/03/09 | 4-6 weeks | High (if pursuing P1 cluster mgmt contracts) |
| GitOps-driven cluster lifecycle (FluxCD/ArgoCD) | CCM-09/10 | 3-4 weeks | High (cluster mgmt only) |
| Automated K8s version upgrade orchestration | CCM-10 | 2-3 weeks | Medium (cluster mgmt only) |
| Multi-cloud IaC catalog (Terraform/Crossplane modules) | CCM-02/03 | 3-4 weeks | Medium (cluster mgmt only) |
| Self-service deployment wizard/CLI bootstrapper | CCM-01 | 2-3 weeks | Medium |

### Tier 3: Process/Documentation

| Enhancement | CCM Req | Effort | Notes |
|-------------|---------|--------|-------|
| Service delivery plan template | CCM-14/19 | 1 week | Document template, not code |
| Licensed software declaration + SBOM process | CCM-21 | 1 week | Depends on Tier 1 SBOM tooling |
| Personnel/OPSEC/CUI compliance package | CCM-22/23 | 1 week | Leverage Sovran secure mode docs |
| Data rights handling guidance | CCM-25 | 1 week | Legal template |
| Monthly status report template | CCM-19 | 2 days | Simple document |

---

## 5. What Aegis Already Does Better Than ENBUILD

Based on the VivSoft proposal and the RTM's gap analysis of ENBUILD, Aegis has clear advantages in several areas:

| Capability | Aegis | ENBUILD (from RTM gaps) |
|-----------|-------|------------------------|
| **Workload scheduling + GPU placement** | Built-in placement algorithm with TTFG, spread strategies, policy constraints | Not applicable — ENBUILD is cluster lifecycle only |
| **Interactive workspaces (VS Code Remote)** | Full WebSocket tunnel with session tokens, idle timeouts, one-time-use JTI | Not applicable |
| **Budget enforcement** | HARD (reject) and SOFT (warn) modes, transactional locking, per-queue/per-project | Kubecost URL placeholder only (RTM: "Cost visibility is not the same as spend alerting") |
| **Audit logging** | AU-3 compliant structured JSON on every auth/authz/workload event | Not assessed as a strength |
| **Fail-closed RBAC** | Empty policy = deny all. Project + queue scoping. | Basic auth/RBAC scaffolding (RTM: "Missing" for external customer model) |
| **Session management** | 15-min idle timeout, 5-min max token TTL, one-time-use, JTI enforcement | Not assessed |
| **Internal PKI** | step-ca + cert-manager, 90-day rotation, no internet dependency | Not mentioned |
| **Cluster onboarding** | Stores auth material (kubeconfig/assume-role/endpoint+CA), spoke agent auto-registers via heartbeat, project-to-cluster binding. Designed for workload scheduling. | Repo URL lookup to link clusters to GitOps repos. Designed for lifecycle management. Different problem — not directly comparable. |
| **CUI secure mode** | RAM disk, FDE verification, in-memory-only tokens, log redaction | Not applicable |

**Key talking point:** "ENBUILD is a cluster lifecycle automation tool. Aegis is a **secure workload orchestration platform** with cluster management capabilities. We do everything ENBUILD does for cluster import and monitoring, plus workload scheduling, budget enforcement, interactive workspaces, and deep security architecture that ENBUILD doesn't attempt."

---

## 6. Proposed Enhanced Roadmap

Merging the existing 7-gap roadmap with CCM-derived requirements:

### Phase 1: Security Foundation — DoD Hard Requirements (Weeks 1-6)
*Must-have for any DoD deployment. Existing roadmap + CCM Tier 1.*

| Item | Source | Effort |
|------|--------|--------|
| FIPS 140-3 cryptography | Existing Gap 1 | 2-4 weeks |
| Iron Bank image publishing + SBOM + scanning | CCM-17 | 2-3 weeks |
| Full mTLS spoke-to-hub | Existing Gap 5 | 1-2 weeks |
| PIV/CAC enforcement | Existing Gap 6 | 1 week |

### Phase 2: DoD Market Access (Weeks 6-10)
*Enable Aegis to be delivered into DoD environments. CCM Tier 1.*

| Item | Source | Effort |
|------|--------|--------|
| SIEM log export (Fluent Bit sidecar in Helm) | CCM-11 | 1-2 weeks |
| GitLab CI/CD pipeline (alongside GitHub Actions) | CCM-18 | 1-2 weeks |
| Audit log viewer (wire existing UI to API) | Existing Gap 2 | 3-5 days |

### Phase 3: Operational Maturity (Weeks 10-14)
*Differentiators that show enterprise readiness. CCM Tier 1 + Tier 2.*

| Item | Source | Effort |
|------|--------|--------|
| SLA framework (define, track, dashboard, alert) | CCM-15 | 2-3 weeks |
| Cloud billing alarm integration | CCM-06 | 1-2 weeks |
| Per-project cost metering + chargeback reports | CCM-16 | 2-3 weeks |

### Phase 4: Platform Polish (Weeks 14-18)
*Extend existing capabilities. CCM Tier 2 + existing roadmap.*

| Item | Source | Effort |
|------|--------|--------|
| Azure AKS cluster import support | CCM-04/05 | 3-4 weeks |
| Cluster detail view in UI (node/pod metrics) | CCM-13 | 1-2 weeks |
| Import connectivity validation | CCM-08 | 1 week |
| Kueue fair queuing (Helm exposure) | Existing Gap 3 | 1 day |
| PyTorch distributed training | Existing Gap 7 | 1 week |

### Parallel Track: Process Documentation
*Not dependent on code. Run alongside any phase.*

| Item | Source | Effort |
|------|--------|--------|
| Service delivery plan template | CCM-14/19 | 1 week |
| Compliance documentation productization | Existing plan | 3 weeks |
| License declaration + SBOM process docs | CCM-21/25 | 1 week |

### Deferred: Cluster Lifecycle Management
*Only pursue if strategically expanding Aegis's scope.*

| Item | Source | Effort |
|------|--------|--------|
| Big Bang deployment integration | CCM-02/03/09 | 4-6 weeks |
| GitOps-driven cluster lifecycle | CCM-09/10 | 3-4 weeks |
| Multi-cloud IaC catalog | CCM-02/03 | 3-4 weeks |
| K8s version upgrade orchestration | CCM-10 | 2-3 weeks |

---

## Appendix: CCM Contract Key Numbers

| Metric | Value |
|--------|-------|
| Contract value (prototype) | $254,800 |
| Period of performance | 120 days |
| Authority | 10 U.S.C. 4022 (OTA) |
| Follow-on potential | Production OTA under 10 U.S.C. 4022(f) |
| ENBUILD license (prototype) | $50,000 |
| ENBUILD services (prototype) | $204,800 (4 FTEs x 320hrs x $160 avg) |
| Production pricing (per cluster/year) | Small: $180K, Medium: $210K, Large: $240K |
| Production SLA targets | <30min provisioning, 99.9% uptime, <1hr incident response |
