# AFWERX Open Topic SBIR Phase I — Proposal Draft

> **Status**: DRAFT — Not for submission. Pending SAM.gov UEI validation.
> **Target**: AFWERX SBIR Open Topic (rolling deadline)
> **Phase I Award**: $50,000–$75,000 for 3–6 months
>
> **Also applicable to** (same core content, reframed per solicitation):
> - **Kessel Run DevSecOps CSO** (FA8730-22-S-C001) — verified active on SAM.gov, open until Nov 29, 2027. No SAM.gov needed for submission.
> - **AFLCMC Data Operations CSO** (FA8600-23-S-C056) — verified active, open until Feb 8, 2028.
> - **JADC2 CSO** (FA8612-21-S-C001) — verified active, open until Nov 9, 2026.
> - **DIU CSO** — always open at diu.mil/work-with-us. No SAM.gov needed for submission.

---

## 1. Cover Page Information

| Field | Value |
|-------|-------|
| Company Name | Aegis Platform LLC |
| Address | Virginia, USA |
| EIN | 41-4974971 |
| UEI | Pending (SAM.gov validation in progress) |
| CAGE Code | Pending (assigned during SAM.gov registration) |
| NAICS Code (Primary) | 511210 — Software Publishers |
| Additional NAICS | 518210, 541511, 541512, 541519, 541715 |
| PSC Codes | D302, D306, D307, D310, 7030 |
| Principal Investigator | Carlos Sanchez |
| PI Phone | [To be filled] |
| PI Email | [To be filled] |
| Company Size | Small Business, 1 employee |
| Business Type | LLC |

---

## 2. Technical Abstract (250 words max)

**Title: Aegis — Secure GPU Workspace Platform for DoD AI/ML Programs**

Department of Defense data scientists face a critical infrastructure bottleneck: accessing GPU hardware inside Authority to Operate (ATO) boundaries requires 3–6 months of infrastructure assembly spanning Transit Gateway, Network Firewall, AWS Workspaces, and Kubeflow across multiple accounts. The resulting environment delivers degraded developer experience — browser-based Jupyter notebooks through remote desktop sessions with high latency and no real debugging capability.

Aegis Platform eliminates this bottleneck with a self-hosted, compliance-first GPU control plane that deploys inside existing ATO boundaries in under one week. Using a hub-and-spoke Kubernetes architecture, Aegis schedules GPU workloads across distributed clusters while enforcing NIST 800-53/800-171 compliance controls, budget limits, and multi-tenant isolation. Data scientists launch workspaces through a web interface, selecting GPU size and container image, and connect via native VS Code through the Sovran extension — achieving p95 time-to-first-GPU under 90 seconds.

The platform ships with built-in security: Keycloak OIDC authentication with MFA enforcement, fail-closed RBAC, single-use JWT session tokens with 5-minute TTL, structured audit logging, default-deny network policies, and internal PKI via step-ca. Budget enforcement prevents GPU cost overruns with HARD (reject) and SOFT (warn) policy modes. OSCAL-formatted evidence export accelerates ATO documentation.

Aegis is a working prototype deployed on AWS EKS with demonstrated multi-cluster GPU scheduling, workspace launching, budget enforcement, and compliance controls. Iron Bank container image submission is in progress at repo1.dso.mil. Phase I will validate Aegis with Air Force data science teams on a partner cluster.

---

## 3. Technical Approach

### 3.1 Problem Statement

DoD AI/ML programs face two compounding problems:

**Infrastructure deployment timeline**: Standing up a CUI-compliant ML environment currently requires 3–6 months of infrastructure engineering across 4+ AWS accounts. This includes Transit Gateway for network connectivity, Network Firewall for traffic inspection, AWS Workspaces for user access, Active Directory for identity, and Kubeflow for notebook/pipeline management. Each component requires separate ATO review, and the assembled system requires 1-2 full-time platform engineers to maintain.

**Degraded developer experience**: Once deployed, data scientists RDP into a Workspaces VM, open a browser, navigate to Kubeflow, and use Jupyter notebooks on a remote desktop. This eliminates local IDE extensions, introduces latency on every keystroke, prevents real debugging (breakpoints, variable inspection), and starts with a 5-10 minute VM boot. The result is measurably lower productivity compared to commercial ML environments.

These problems directly impact DoD AI readiness. Programs that cannot give data scientists GPU access quickly cannot train models quickly, and programs with poor developer experience lose talent to commercial organizations.

### 3.2 Proposed Solution — Aegis Platform Architecture

Aegis replaces the multi-month infrastructure assembly with a Helm chart deployment and replaces the degraded browser IDE with native VS Code on remote GPU hardware.

**Hub-and-Spoke Architecture:**

The **hub cluster** runs three services:
- **Platform API** (Go, gRPC + REST gateway): Central orchestration service handling workload placement, budget enforcement, cluster management, and session token minting. Placement algorithm selects optimal spoke cluster based on GPU flavor availability, heartbeat freshness, and time-to-first-GPU metrics.
- **Keycloak** (OIDC + MFA): Enterprise identity provider with realm auto-provisioning, TOTP and WebAuthn MFA enforcement, PKCE authentication flows, and CAC/PIV readiness.
- **Proxy** (Go, WebSocket): Authenticated reverse proxy for VS Code workspace connections. Validates JWT bearer tokens with JTI single-use enforcement, audience restriction, and 5-minute TTL.

**Spoke clusters** run a lightweight agent (Go, controller-runtime) that:
- Heartbeats outbound to the hub every 10 seconds (no inbound firewall rules required)
- Reports GPU capacity, node health, and workload status
- Reconciles AegisWorkload CRDs into Kubernetes Jobs with Kueue queue annotations
- Manages workspace pod lifecycle (creation, monitoring, cleanup)

**Developer workflow:**
1. Data scientist opens Backstage web UI or VS Code with Sovran extension
2. Authenticates via Keycloak OIDC with MFA
3. Selects GPU size (flavor), container image, and queue
4. Platform API runs placement algorithm → selects spoke → creates AegisWorkload CRD
5. Spoke agent reconciles CRD → Kubernetes Job → workspace pod starts
6. Data scientist clicks "Connect" in VS Code → Sovran opens WebSocket tunnel through proxy
7. **Result: Native VS Code on GPU hardware in under 90 seconds**

### 3.3 What Has Already Been Built

Aegis is a working prototype, not a concept. The following capabilities are implemented and tested:

| Capability | Implementation Status |
|-----------|----------------------|
| Multi-cluster hub-and-spoke orchestration | **Built** — gRPC communication, heartbeat, stale detection |
| GPU placement algorithm | **Built** — spread + lowest-TTFG strategies, policy constraint filtering |
| Budget enforcement | **Built** — per-queue HARD/SOFT modes, transactional reservation (SELECT FOR UPDATE) |
| Keycloak OIDC + PKCE + MFA | **Built** — realm auto-provisioned, TOTP + WebAuthn |
| Fail-closed RBAC | **Built** — empty bindings deny all; scoped by project/queue/client |
| Single-use JWT session tokens | **Built** — HS256, JTI, audience-restricted, 5-min TTL |
| Proxy JWT validation | **Built** — bearer extraction, JTI reuse detection, audience validation |
| Structured audit logging | **Built** — PostgreSQL events: type, subject, resource, action, outcome, source_ip |
| Network policies | **Built** — default-deny egress + per-service allowlists (hardening profile gated) |
| VS Code extension (Sovran) | **Built** — PKCE auth, WebSocket tunnel, 85% TTL auto-renewal, workspace discovery |
| Backstage UI (23 pages) | **Built** — workloads, clusters, projects, budgets, metrics, logs, alerts |
| Internal PKI | **Built** — step-ca + cert-manager, no internet dependency |
| AWS EKS provisioning | **Built** — Pulumi Automation API for VPC, EKS, node pools, IAM |
| Cluster import | **Built** — kubeconfig, assume-role, endpoint/CA methods |
| Multi-tenancy | **Built** — per-project namespaces, AWS credentials, cluster-project binding |
| Cost estimation | **Built** — GPU-hours × price from flavor, Prometheus metrics |
| Prometheus budget metrics | **Built** — actual/reserved/denied/overrun counters |

### 3.4 Technical Differentiators vs. Kubeflow

| Area | Aegis | Kubeflow |
|------|-------|---------|
| Multi-cluster scheduling | Built-in, TTFG-optimized placement | Single cluster only |
| Budget enforcement | HARD/SOFT modes with Prometheus metrics | None |
| Developer experience | Native VS Code via WebSocket tunnel | Browser-only Jupyter |
| Session security | Single-use JWT, 5-min TTL, JTI enforcement | Long-lived browser sessions |
| Deployment time | ~1 week (Helm chart) | 12–20+ weeks to equivalent security posture |
| Customer engineering (Year 1) | ~1.05 FTE | ~9.5 FTE |
| 3-year customer engineering TCO | ~$600K | ~$4.6M |

---

## 4. Phase I Work Plan — 6 Months, $50K–$75K

### Objective

Deploy Aegis on an Air Force partner cluster, validate p95 <90s time-to-first-GPU with AF data scientists, generate OSCAL evidence packages, and document ATO support artifacts.

### Month 1–2: Partner Engagement & Environment Setup

- Identify AF data science team partner (via AFWERX connection or direct outreach)
- Deploy Aegis hub on partner's GovCloud EKS or RKE2 cluster
- Configure Keycloak with partner's identity provider (OIDC federation)
- Deploy spoke agent on partner's GPU-equipped cluster
- Validate end-to-end workspace launch in partner environment

**Milestone 1**: Aegis deployed and operational on AF partner cluster. First workspace launched successfully.

### Month 2–4: User Validation & Performance Measurement

- Onboard 5-10 AF data scientists onto the platform
- Measure time-to-first-GPU across 100+ workspace launches
- Collect user feedback on VS Code workspace experience vs. existing tools
- Document deployment playbook for AF environments (GovCloud + RKE2)
- Iterate on workspace images based on user requirements (PyTorch, TensorFlow, custom libraries)

**Milestone 2**: p95 time-to-first-GPU <90 seconds validated with AF data scientists. User feedback documented.

### Month 4–6: Compliance Documentation & ATO Support Package

- Generate OSCAL component definition for Aegis
- Map all built-in controls to NIST 800-53 Rev 5 moderate baseline
- Create DISA Kubernetes STIG compliance checklist for Aegis deployment
- Generate SBOMs for all container images (CycloneDX format)
- Run vulnerability scans and remediate critical/high findings
- Package complete ATO support documentation:
  - Customer Responsibility Matrix
  - Control Implementation Statements
  - Hardening Guide
  - STIG Compliance Checklist
  - SBOM
  - Vulnerability scan results

**Milestone 3**: Complete ATO support package delivered to partner ISSM. OSCAL evidence export demonstrated.

### Phase I Deliverables

1. Technical report: deployment results, performance metrics, user feedback analysis
2. ATO support package: NIST 800-53 control mapping, STIG checklist, OSCAL component definition, SBOMs
3. Deployment playbook: step-by-step guide for AF GovCloud and RKE2 environments
4. Phase II proposal: expanded deployment plan, additional AF programs, production hardening roadmap

### Budget Estimate

| Category | Amount |
|----------|--------|
| PI labor (6 months, part-time) | $35,000–$50,000 |
| Cloud infrastructure (GovCloud EKS for testing) | $5,000–$10,000 |
| Travel (partner site visits, 2 trips) | $3,000–$5,000 |
| Security scanning tools and licenses | $2,000–$5,000 |
| Compliance documentation tools | $1,000–$2,000 |
| **Total** | **$46,000–$72,000** |

---

## 5. Phase II Vision — $500K–$1.7M over 24 Months

### Phase II Objectives

1. **Production hardening**: FIPS 140-3 cryptography (BoringCrypto for Go, UBI9 FIPS base images), mTLS spoke-to-hub communication, IL-level data classification enforcement
2. **Iron Bank certification**: Get all Aegis container images approved in Iron Bank (registry1.dso.mil) for DoD-wide distribution
3. **Platform One integration**: Deploy Aegis as a Big Bang addon on RKE2 with Istio service mesh compatibility
4. **Multi-program deployment**: Deploy to 3-5 additional AF programs with different GPU configurations, classification levels, and compliance requirements
5. **Advanced ML workload support**: Kueue fair queuing integration, PyTorch distributed training, model deployment pipelines across clusters
6. **Enterprise features**: PIV/CAC MFA enforcement, IL-level policy domains, cost-aware placement, audit log viewer and SIEM integration

### Phase II Success Criteria

- Aegis deployed in production at 3+ AF programs
- Iron Bank-approved container images available to all DoD programs
- FIPS 140-3 validated cryptography in all Aegis services
- p95 time-to-first-GPU <90s maintained at scale (50+ concurrent users per cluster)
- ATO support package used in 2+ successful ATO assessments
- User satisfaction score >8/10 from AF data scientists

### Phase III Potential

Phase III SBIR contracts are unlimited value and can be sole-sourced. A successful Phase II positions Aegis for:
- Production deployment across Air Force-wide AI/ML programs
- Expansion to other services (Navy, Army, Space Force)
- Integration into Platform One's standard tool offering
- Enterprise license agreements with defense contractors
- Potential annual contract value: $1M–$10M+ depending on scope

---

## 6. Company Background

### Principal Investigator — Carlos Sanchez

- 5+ years of defense infrastructure engineering experience
- Built multi-month AWS GovCloud CUI ML environment from scratch: Transit Gateway, Network Firewall, AWS Workspaces, Active Directory, Kubeflow, GPU clusters — the exact architecture Aegis replaces
- Hands-on experience with the pain points Aegis solves: spent months assembling what should take days
- Deep Kubernetes expertise: controller-runtime operators, Helm chart authoring, multi-cluster management
- Go systems programming: gRPC services, WebSocket proxies, JWT/OIDC authentication
- Active Platform One access with repo1.dso.mil account
- P1 CAC and Air Force clearance access from existing contract

### Company — Aegis Platform LLC

- Virginia-based LLC
- EIN: 41-4974971
- SAM.gov registration: submitted, pending UEI validation
- Self-certified small business under NAICS 511210
- Iron Bank onboarding: submitted to repo1.dso.mil

### Existing Codebase

Aegis is not a concept — it is a working platform:

| Component | Language | Lines of Code | Status |
|-----------|----------|--------------|--------|
| Platform API | Go | 10,000+ | Production-ready |
| K8s Agent (Operator) | Go | 3,000+ | Production-ready |
| Proxy | Go | 2,000+ | Production-ready |
| Sovran (VS Code Extension) | TypeScript | 3,000+ | Production-ready |
| Aegis UI (Backstage) | TypeScript/React | 15,000+ | Production-ready |
| Helm Charts (hub + spoke) | YAML | 3,000+ | Production-ready |
| Terraform (AWS infra) | HCL | 2,000+ | Production-ready |
| Compliance Documentation | Markdown | 5,000+ | Comprehensive |

---

## 7. Prior Work & Related Efforts

### Working EKS Deployment

Aegis is currently deployed on AWS EKS with:
- Hub cluster running platform-api, proxy, Keycloak, Backstage UI, PostgreSQL
- Spoke cluster with k8s-agent heartbeating to hub
- End-to-end workspace launch and VS Code connection verified
- Budget enforcement, audit logging, and RBAC operational
- Internal PKI (step-ca + cert-manager) providing certificate infrastructure

### Compliance Documentation

Extensive compliance documentation exists:
- NIST 800-53 Rev 5 control implementation statements
- NIST 800-171 R3 compliance mapping
- Customer Responsibility Matrix (CRM)
- Hardening Guide
- OSCAL component definition (in progress)
- Session management controls (AC-11, AC-12)
- Audit event system design

### Iron Bank Submission

Container images submitted to repo1.dso.mil for Iron Bank evaluation:
- Built on approved base images
- Non-root execution with read-only filesystems
- SBOM generation capability
- CVE scanning and remediation process established

### Architecture Documentation

Comprehensive architecture documentation covering:
- Hub-and-spoke networking with outbound-only spoke connectivity
- gRPC service definitions (Protocol Buffers)
- Authentication and session management flows
- Budget enforcement transaction model
- Placement algorithm design
- Multi-tenant isolation model

---

## Appendix A: Relevant NAICS and PSC Codes

| Code | Description | Relevance |
|------|-------------|-----------|
| **NAICS 511210** | Software Publishers | Primary — Aegis is commercial software |
| NAICS 518210 | Data Processing, Hosting, and Related Services | GPU workload hosting/scheduling |
| NAICS 541511 | Custom Computer Programming Services | Platform development |
| NAICS 541512 | Computer Systems Design Services | System architecture |
| NAICS 541519 | Other Computer Related Services | Compliance automation |
| NAICS 541715 | R&D in Physical/Engineering/Life Sciences | SBIR R&D alignment |
| **PSC D302** | ADP Systems Development Services | Platform development |
| PSC D306 | ADP Systems Analysis Services | Architecture design |
| PSC D307 | Automated Information Systems Design/Integration | System integration |
| PSC D310 | ADP Backup and Security Services | Security controls |
| PSC 7030 | ADP Software | Software product |

---

*This document is a DRAFT proposal outline. Do not submit until: (1) SAM.gov UEI validation completes, (2) SBIR.gov company profile is created, (3) AFWERX account is registered, (4) Current AFWERX Open Topic submission requirements are verified on afwerx.com/sbir-sttr.*
