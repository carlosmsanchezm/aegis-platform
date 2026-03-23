# Aegis Platform — Shared Responsibility Matrix

Aegis is an application-layer platform that deploys inside the customer's existing network and ATO boundary. This document defines who owns what.

**The rule:** Aegis owns application-layer security controls. The customer owns infrastructure-layer controls. Aegis does not manage, configure, or monitor the customer's network boundary.

---

## Responsibility Matrix

| Area | Customer Owns | Aegis Provides |
|---|---|---|
| **Network boundary** (VPC, firewall, egress filtering, squid proxy, Transit Gateway) | Full ownership — design, deploy, maintain, monitor | Not involved. Aegis runs inside the boundary. |
| **Kubernetes cluster** | Provision and maintain cluster (EKS, RKE2, OpenShift, Big Bang) | Spoke Helm chart + optional Pulumi provisioner for EKS |
| **DNS and load balancing** | Route53, NLB/ALB, DNS resolution | Ingress configuration via Helm values |
| **TLS at boundary** | TLS termination at load balancer or ingress | mTLS between all internal services (platform-api, proxy, k8s-agent) |
| **Container registry** | ECR, Harbor, Iron Bank access | Iron Bank-compatible container images (UBI9, distroless) |
| **Identity provider** | AD/LDAP federation if needed | Keycloak (OIDC/SAML, MFA, CAC/PIV support) |
| **Authentication** | Configure Keycloak realm and MFA policies | Enforce OIDC PKCE, validate JWT, check MFA claims |
| **Authorization** | Define project/role bindings, assign users to projects | Enforce RBAC (fail-closed), namespace isolation, budget limits |
| **GPU scheduling & budgets** | Set budget limits per project/queue | Placement algorithm, Kueue fair queuing, budget enforcement |
| **Audit logging** | Ingest logs into SIEM (Splunk, ELK, CloudWatch) | Generate structured JSON audit events (AU-3 compliant) |
| **Compliance documentation** | Own the SSP, POA&M, assessment package | Provide control implementation statements for application controls |
| **Endpoint protection** (developer workstations) | MDM policy (FDE, clipboard restrictions, EDR) | Sovran secure mode (RAM-disk, ephemeral tokens, session cleanup) |
| **Continuous monitoring** | Nessus, Qualys, Prisma Cloud, network scanning | Application health metrics, Prometheus endpoints |
| **Data encryption at rest** | EBS encryption, RDS encryption, KMS key management | Delegates to infrastructure — no application-level at-rest encryption |
| **Data encryption in transit** | TLS at load balancer, VPN tunnels | TLS 1.2+ on all connections, mTLS between services, FIPS cipher suites |

---

## NIST Control Ownership

### NIST 800-53 Rev 5 / NIST 800-171 R3 / CMMC Level 2

| Control | Name | Owner | How |
|---|---|---|---|
| **SC-7** | Boundary Protection | **Customer** | VPC security groups, Network Firewall, squid proxy, deny-all egress |
| **SC-8** | Transmission Confidentiality | **Shared** | Customer: TLS at load balancer. Aegis: mTLS between services, TLS 1.2+ on WebSocket. |
| **SC-13** | Cryptographic Protection | **Shared** | Customer: FIPS on infrastructure (OpenSSL, KMS). Aegis: FIPS in application (BoringCrypto, cipher suite config). |
| **SC-28** | Protection at Rest | **Customer** | EBS/RDS encryption with KMS. Aegis stores workload metadata in PostgreSQL — encrypted by customer's infrastructure. |
| **AC-2** | Account Management | **Aegis** | Keycloak OIDC with automated provisioning, MFA enforcement, session lifecycle |
| **AC-3** | Access Enforcement | **Aegis** | RBAC per project with fail-closed policy. No bindings configured = all requests denied. |
| **AC-11** | Session Lock | **Aegis** | Configurable inactivity timeout, auto-disconnect in secure mode |
| **AC-12** | Session Termination | **Aegis** | Single-use JWT tokens (5-min TTL, JTI tracking), forced session revoke |
| **AU-2** | Audit Events | **Aegis** | All security-relevant events logged: auth, workspace connect/disconnect, workload lifecycle |
| **AU-3** | Content of Audit Records | **Aegis** | Structured JSON: timestamp, subject, action, resource, outcome, source IP |
| **MP-7** | Media Protection | **Shared** | Customer: MDM policy on workstations (FDE, clipboard). Aegis: Sovran RAM-disk, ephemeral tokens, session cleanup. |
| **IA-2** | Identification & Authentication | **Aegis** | OIDC PKCE with Keycloak, MFA enforcement, CAC/PIV support via WebAuthn |
| **CM-7** | Least Functionality | **Aegis** | Minimal container images (distroless/UBI9), non-root, read-only rootfs, dropped capabilities |

---

## Deployment Prerequisites

Before deploying Aegis, the customer's environment must have:

### Required

- [ ] **Kubernetes cluster** — EKS, RKE2, OpenShift, or Big Bang (CNCF-conformant)
- [ ] **Egress filtering** — Deny-all by default. Allow required endpoints (container registry, OIDC issuer, DNS).
- [ ] **TLS certificates** — For ingress, or cert-manager installed for automatic provisioning
- [ ] **DNS resolution** — Internal service DNS must work (CoreDNS or equivalent)
- [ ] **Container registry** — Accessible from cluster nodes (ECR, Harbor, Iron Bank at registry1.dso.mil)
- [ ] **Storage class** — For workspace persistent volumes (`gp3` on EKS, `managed-premium` on AKS, etc.)
- [ ] **kubectl access** — Cluster admin access for Helm install

### Optional but recommended

- [ ] **VPC endpoints** — For AWS services (S3, ECR, STS) if running in air-gapped or restricted egress environment
- [ ] **MDM policy** — Full disk encryption, clipboard restrictions, EDR on developer workstations
- [ ] **SIEM endpoint** — For ingesting Aegis audit logs (Splunk HEC, Elasticsearch, CloudWatch Logs)
- [ ] **cert-manager + step-ca** — For internal PKI (Aegis provides install script: `scripts/install-internal-pki.sh`)

### What Aegis deploys (via Helm)

- **Hub cluster**: platform-api, proxy, Keycloak, Backstage UI, PostgreSQL, ingress-nginx
- **Spoke cluster(s)**: k8s-agent, spoke-proxy, AegisWorkload CRDs
- **PKI** (optional): cert-manager, step-ca, step-issuer (via external install script)

---

## For ISSMs: How to reference Aegis in your SSP

Aegis provides control implementation statements for all application-layer controls listed above. Your SSP should:

1. **Inherit** Aegis control statements for AC-2, AC-3, AC-11, AC-12, AU-2, AU-3, IA-2, CM-7
2. **Document as shared** for SC-8, SC-13, MP-7 — specify the customer's infrastructure controls and Aegis's application controls
3. **Document as customer-owned** for SC-7, SC-28 — these are infrastructure controls Aegis does not touch

Control implementation statements are available in `docs/compliance/customer-docs/control-implementation-statements.md` and can be exported in OSCAL format.
