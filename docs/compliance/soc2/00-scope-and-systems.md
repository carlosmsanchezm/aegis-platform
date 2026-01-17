# SOC 2 Type II Scope and Systems Inventory

**Document ID:** SOC2-SCOPE-001
**Version:** 1.0
**Created:** 2025-01-17
**Status:** ACTIVE

## Executive Summary

This document defines the scope of Aegis Technologies' SOC 2 Type II audit and inventories all systems that process, store, or transmit data relevant to the Trust Services Criteria.

## Scope Decision

### Selected Scope: **Model B (Self-Hosted Software)**

Aegis currently operates as a **self-hosted software vendor** (Model B). Customers deploy and operate Aegis in their own environments. The SOC 2 scope covers Aegis's corporate systems and software development lifecycle:

| Scope Area | Included | Rationale |
|------------|----------|-----------|
| Corporate Systems | ✅ Yes | Required for all SOC 2 audits |
| SDLC & Development | ✅ Yes | Software is the product |
| Production/Service Operations | ❌ No | Not operating customer environments (Model B) |
| Support Systems | ✅ Yes | Customer interactions are in scope |

### Trust Services Criteria Selection

| Criteria | Include? | Rationale |
|----------|----------|-----------|
| **Security** | ✅ Required | Always required |
| Availability | ❌ No | Not operating production services for customers |
| Confidentiality | ⬜ Defer | Consider for Year 2 if handling classified data |
| Processing Integrity | ⬜ No | Not a financial/transactional system |
| Privacy | ⬜ No | Not a consumer-facing service |

> **Note:** Security-only scope is appropriate for Model B (software vendor). Availability would be added when/if Aegis begins operating managed services (Model C).

---

## Company Profile

| Attribute | Value |
|-----------|-------|
| Company Name | Aegis Technologies (DBA Aegis) |
| Primary Product | Aegis Platform - Multi-cluster GPU workload orchestration |
| Service Model | Model B (self-hosted software) |
| Team Size | 1-5 people (founder-led) |
| Primary Cloud | AWS GovCloud (for development/testing) |
| GitHub Owner | `carlosmsanchezm` (personal account, transitioning to org) |
| Corporate Domain | TBD (planned: Google Workspace) |

---

## Systems Inventory

### 1. Identity & Access Management

| System | Purpose | Data Types | SOC 2 Relevance | Evidence Location |
|--------|---------|------------|-----------------|-------------------|
| **Google Account (Personal)** | Current corporate identity | Auth credentials | CC6.1, CC6.6 | Google Admin Console |
| **Google Workspace** (Planned) | Corporate IdP, email, SSO | Employee identities | CC6.1, CC6.2, CC6.6 | Admin Console |
| **GitHub** | Code repository access | Developer access | CC6.1, CC6.7 | GitHub Audit Log |
| **AWS IAM** | Cloud infrastructure access | Service & user access | CC6.1, CC6.7 | CloudTrail, IAM |
| **Keycloak** (Product) | Customer authentication | Customer identities | CC6.1 | Keycloak Logs |

**Current State:**
- ⚠️ No corporate IdP established
- ⚠️ GitHub uses personal account (`carlosmsanchezm`)
- ⚠️ AWS access via personal credentials
- 🔴 **GAP:** Need to establish company domain + Google Workspace/Cloud Identity

### 2. Source Code & Development

| System | Purpose | Data Types | SOC 2 Relevance | Evidence Location |
|--------|---------|------------|-----------------|-------------------|
| **GitHub** (carlosmsanchezm) | Source code repository | Code, configs, secrets (encrypted) | CC8.1 | GitHub API |
| **GitHub Actions** | CI/CD pipelines | Build artifacts, test results | CC8.1, CC7.1 | Workflow logs |
| **Go/Node.js** | Development languages | Dependencies | CC7.1 | SBOM |
| **Docker/Container Registry** | Container images | Built artifacts | CC7.1 | ECR/Registry |

**Repositories in Scope:**

| Repository | Purpose | Commits | Controls |
|------------|---------|---------|----------|
| `carlosmsanchezm/aegis-platform` | Platform backend, K8s agents, proxy | 465 | ✅ Branch protection, ✅ Dependabot, ✅ CODEOWNERS |
| `carlosmsanchezm/aegis-ui` | Frontend UI | 99 | ✅ Branch protection, ✅ Dependabot |
| `carlosmsanchezm/sovran` | Infrastructure/IaC | 161 | ✅ Branch protection, ✅ Dependabot |

**Current State:**
- ✅ GitHub Actions CI/CD exists
- ✅ Branch protection enabled on all repos
- ✅ CODEOWNERS file (aegis-platform)
- ✅ Dependabot enabled on all repos
- ⚠️ Secret scanning requires GitHub Advanced Security
- ⚠️ Vulnerabilities to remediate: aegis-ui (25), sovran (19)

### 3. Cloud Infrastructure

| System | Purpose | Data Types | SOC 2 Relevance | Evidence Location |
|--------|---------|------------|-----------------|-------------------|
| **AWS GovCloud** | Development/test infrastructure | Dev workloads, CI/CD | CC6.1, CC7.2 | CloudTrail |
| **EKS** | Kubernetes clusters (dev/test) | Workload orchestration | CC6.7 | K8s audit logs |
| **RDS** | Database (dev/test) | Application data | CC6.1 | RDS logs |
| **S3** | Object storage | Terraform state, artifacts | CC6.1 | S3 access logs |
| **ECR** | Container registry | Docker images | CC7.1 | ECR access logs |

**AWS Resources Identified (from Terraform):**
- S3: `aegis-platform-tf-state-bucket` (Terraform state)
- DynamoDB: `aegis-terraform-locks`
- EKS: Aegis clusters (prod)
- Region: `us-east-1` (GovCloud)

**Current State:**
- ✅ Terraform-managed infrastructure
- ✅ S3 state bucket encrypted
- ⚠️ CloudTrail status unknown
- ⚠️ AWS Config status unknown
- 🔴 **GAP:** Verify CloudTrail, Config, GuardDuty, Security Hub

### 4. Monitoring & Logging

| System | Purpose | Data Types | SOC 2 Relevance | Evidence Location |
|--------|---------|------------|-----------------|-------------------|
| **Prometheus** | Metrics collection | System metrics | CC7.2 | Prometheus/Grafana |
| **Platform API Logs** | Application audit logs | Auth, authz events | AU-2, AU-3 | Stdout/SIEM |
| **CloudTrail** | AWS audit logs | API calls | CC7.2, AU-2 | S3/CloudWatch |
| **GitHub Audit Log** | Repository activity | Code changes, access | CC8.1 | GitHub API |

**Current State:**
- ✅ Prometheus mentioned in architecture
- ⚠️ Structured logging partially implemented
- 🔴 **GAP:** Centralized log aggregation, alerting, retention policy

### 5. Ticketing & Documentation

| System | Purpose | Data Types | SOC 2 Relevance | Evidence Location |
|--------|---------|------------|-----------------|-------------------|
| **Jira** | Task tracking | Work items, incidents | CC4.2, CC7.4 | Jira API |
| **GitHub Issues** | Bug tracking | Defects, features | CC7.4 | GitHub API |
| **Markdown Docs** | Policies, procedures | Compliance docs | CC2.2 | Git repo |

**Current State:**
- ✅ Jira integration mentioned in AGENTS.md
- ✅ Documentation in repo
- ⚠️ No formal incident tracking system identified

### 6. Product Components (Shipped to Customers)

> **Note:** These systems are developed by Aegis but deployed/operated by customers in their own environments (Model B). They are in scope for SDLC controls but NOT for operational/availability controls.

| System | Purpose | SOC 2 Relevance | Notes |
|--------|---------|-----------------|-------|
| **Aegis Platform API** | Workload orchestration | CC8.1 (secure dev) | Customer-operated |
| **Backstage UI** | Developer portal | CC8.1 (secure dev) | Customer-operated |
| **Keycloak** | Customer auth | CC8.1 (secure dev) | Customer-operated |
| **Workspace Proxy** | Secure access | CC8.1 (secure dev) | Customer-operated |

**Security Features Built Into Product:**
- ✅ mTLS between services
- ✅ OIDC authentication
- ✅ Short-lived tokens (configurable)
- ✅ MFA enforcement (customer-configurable)
- ✅ FIPS-capable cryptography

### 7. Endpoint & Personnel

| System | Purpose | Data Types | SOC 2 Relevance | Evidence Location |
|--------|---------|------------|-----------------|-------------------|
| **macOS** (assumed) | Developer workstation | Code, credentials | CC6.6 | MDM (none) |
| **1Password/similar** | Password management | Credentials | CC6.6 | PM audit log |

**Current State:**
- ⚠️ No MDM/endpoint management
- ⚠️ No centralized password manager policy
- 🔴 **GAP:** Endpoint security, password manager

---

## Data Classification

| Classification | Definition | Examples | Controls |
|----------------|------------|----------|----------|
| **Confidential** | Customer data, credentials | Customer workloads, auth tokens | Encryption, access logging, MFA |
| **Internal** | Company operational data | Configs, metrics, internal docs | Access control, audit |
| **Public** | Publicly available | Public docs, OSS code | No special controls |

---

## Third-Party Services

| Vendor | Service | Data Shared | Assessment Status |
|--------|---------|-------------|-------------------|
| GitHub | Code hosting | Source code, CI logs | ✅ SOC 2 Type II available |
| AWS | Cloud infrastructure | All workloads | ✅ SOC 2/FedRAMP |
| Google | Identity (planned) | Employee identities | ✅ SOC 2 Type II available |
| Atlassian | Jira | Task data | ✅ SOC 2 Type II available |
| Cloudflare | Tunnel (if used) | Traffic metadata | ✅ SOC 2 available |

---

## Boundaries and Exclusions

### In Scope

1. Corporate systems (identity, email, collaboration)
2. All personnel with access to in-scope systems
3. Software development lifecycle (code, CI/CD, testing)
4. Development and test environments
5. Customer support processes

### Out of Scope

1. Customer-owned infrastructure where Aegis is deployed (Model B)
2. Customer data processed within customer environments
3. Production operations (Aegis does not operate production services)
4. Third-party SaaS that does not process Aegis data
5. Personal devices not used for work (if any)

---

## Gap Summary

| ID | Gap | Priority | Remediation |
|----|-----|----------|-------------|
| GAP-001 | No corporate IdP | 🔴 Critical | Stand up Google Workspace |
| GAP-002 | GitHub on personal account | 🔴 Critical | Create GitHub org, transfer repos |
| GAP-003 | No branch protection | 🟡 High | Enable on all repos |
| GAP-004 | No secret scanning | 🟡 High | Enable GitHub secret scanning |
| GAP-005 | No Dependabot | 🟡 High | Enable Dependabot alerts |
| GAP-006 | No endpoint management | 🟡 High | Evaluate MDM (Kandji/Jamf/Fleet) |
| GAP-007 | No centralized logging | 🟡 High | Set up CloudWatch/SIEM |
| GAP-008 | CloudTrail status unknown | 🟡 High | Verify and configure |
| GAP-009 | No formal access reviews | 🟡 High | Establish quarterly process |
| GAP-010 | No CODEOWNERS | 🟢 Medium | Add to repos |

---

## Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| CEO/Founder | Carlos Sanchez | _____________ | ______ |
| Security Lead | Carlos Sanchez | _____________ | ______ |

---

## Next Steps

1. **Immediate:** Create evidence vault structure
2. **Week 1:** Implement GitHub security controls (branch protection, scanning)
3. **Week 2:** Stand up Google Workspace + corporate domain
4. **Week 3:** AWS security baseline verification
5. **Month 1:** Complete control matrix implementation

---

## Future Scope: Model C + Availability

When Aegis begins operating managed services for customers (Model C), the SOC 2 scope will expand:

### Planned Scope Expansion

| Scope Area | Current | Future (Model C) |
|------------|---------|------------------|
| Trust Services - Security | ✅ In scope | ✅ In scope |
| Trust Services - Availability | ❌ Out of scope | ✅ Add to scope |
| Production Operations | ❌ Out of scope | ✅ Add to scope |
| Customer Environment Access | ❌ N/A | ✅ Add controls |

### Additional Controls Required for Model C

| Control Area | Additional Requirements |
|--------------|------------------------|
| **Availability (A1)** | SLOs, uptime monitoring, capacity planning |
| **Personnel** | Background checks, clearances (if DoD) |
| **Operations** | 24/7 on-call, incident response, change windows |
| **Access** | Customer environment access controls, PAM |
| **BC/DR** | Customer-specific recovery procedures |

### Trigger for Scope Expansion

Model C scope expansion should be initiated when:
- [ ] First paying managed-service customer contract signed
- [ ] Production environment operated on customer's behalf
- [ ] Personnel granted access to customer environments

### Timeline Estimate

| Phase | Duration | Activities |
|-------|----------|------------|
| Pre-Model C | Current | Security-only SOC 2, build foundation |
| Model C Prep | 3-6 months | Add Availability controls, operational procedures |
| Model C Audit | 6+ months | Extended audit period with Availability |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.1 | 2025-01-17 | Claude (SOC2 Engineer) | Narrowed scope to Model B only, Security TSC only |
| 1.0 | 2025-01-17 | Claude (SOC2 Engineer) | Initial scope determination |
