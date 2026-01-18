# FedRAMP Scope and Impact Level Assessment

**Document ID:** FEDRAMP-SCOPE-001
**Version:** 1.0
**Created:** 2026-01-17
**Status:** DRAFT - PENDING REVIEW

---

## Executive Summary

This document determines the appropriate FedRAMP impact level and authorization scope for Aegis Platform based on FIPS 199 categorization and the types of federal data the system may process.

## Cloud Service Offering (CSO) Description

### System Name
**Aegis Platform** - Multi-Cluster GPU Workload Orchestration

### System Description
Aegis Platform is a self-hosted software solution that enables organizations to orchestrate GPU workloads across multiple Kubernetes clusters. The platform provides:

- Unified control plane for multi-cluster GPU management
- Workload scheduling and placement optimization
- Developer portal (Backstage-based) for self-service
- Secure workspace proxying for remote development
- Integration with existing identity providers (OIDC/SAML)

### Service Model

| Attribute | Value |
|-----------|-------|
| **Cloud Service Model** | Software-as-a-Service (SaaS) / Platform-as-a-Service (PaaS) |
| **Deployment Model** | Customer self-hosted (Model B) or Aegis-operated (Future Model C) |
| **Target Customers** | Federal agencies, DoD, national laboratories |
| **Primary Use Case** | ML/AI infrastructure, HPC workload management |

---

## FIPS 199 Security Categorization

### Information Types Processed

Based on NIST SP 800-60 guidance, identify information types the system may process:

| Information Type | C | I | A | Rationale |
|------------------|---|---|---|-----------|
| **System Credentials** | M | M | L | User authentication data (email, username) |
| **Workload Metadata** | L | M | L | Job names, resource requests, scheduling info |
| **Infrastructure Config** | M | H | M | Cluster configs, network settings |
| **Audit Logs** | L | M | M | System activity, user actions |
| **Application Code** | L | M | L | Deployed workloads (customer responsibility) |

*C = Confidentiality, I = Integrity, A = Availability*
*L = Low, M = Moderate, H = High*

### High-Water Mark Calculation

| Security Objective | Maximum Impact | Justification |
|--------------------|----------------|---------------|
| **Confidentiality** | MODERATE | Credentials, infrastructure configs |
| **Integrity** | HIGH | Infrastructure config modification |
| **Availability** | MODERATE | GPU workload continuity |

### Preliminary Categorization

```
SC (Aegis Platform) = {(Confidentiality, MODERATE), (Integrity, HIGH), (Availability, MODERATE)}
```

**Initial Assessment:** MODERATE (due to HIGH integrity for infrastructure configuration)

### Justification for LOW Impact Override

However, Aegis Platform qualifies for LOW impact categorization under specific conditions:

| Condition | Met? | Explanation |
|-----------|------|-------------|
| Platform handles minimal PII | ✅ Yes | Only username, email for authentication |
| Customer controls data in their boundary | ✅ Yes | Aegis doesn't access customer workloads |
| Infrastructure config is customer-controlled | ✅ Yes | Customer operates in their environment |
| No CUI processing by Aegis | ✅ Yes | CUI in customer workloads, not Aegis |

**Revised Categorization for Model B (Self-Hosted):**

```
SC (Aegis Platform - Model B) = {(Confidentiality, LOW), (Integrity, LOW), (Availability, LOW)}
```

**Result:** LOW Impact → Eligible for **LI-SaaS or FedRAMP 20x Low**

---

## LI-SaaS Eligibility Assessment

FedRAMP LI-SaaS has specific eligibility criteria. Assess Aegis Platform:

### Required Criteria

| Criterion | Requirement | Aegis Status | Notes |
|-----------|-------------|--------------|-------|
| 1 | Low impact per FIPS 199 | ✅ Eligible | Model B categorization |
| 2 | Cloud-based | ✅ Yes | Kubernetes-native |
| 3 | Does not process PII beyond login | ✅ Yes | Username, email only |
| 4 | Does not store government data | ⚠️ Depends | Customer workload metadata |
| 5 | Hosted on FedRAMP-authorized IaaS/PaaS | ⚠️ Depends | Customer responsibility |
| 6 | Multi-tenant or single-tenant SaaS | ✅ Yes | Customer-isolated |

### Conditional Criteria

| Criterion | Applies? | Impact |
|-----------|----------|--------|
| External system connections | ✅ Yes | +20 conditional controls |
| Privileged user access | ✅ Yes | +conditional controls |
| Underlying IaaS not FedRAMP-authorized | Depends | May need to assess IaaS controls |

### LI-SaaS Control Count Estimate

| Control Type | Count |
|--------------|-------|
| Base controls requiring documentation | 45 |
| Conditional controls (if applicable) | +10-20 |
| Controls requiring attestation only | 75-95 |
| **Total Addressable** | ~125-140 |

---

## FedRAMP 20x Low Eligibility Assessment

FedRAMP 20x is the new streamlined authorization path:

### Key Security Indicators (KSIs) Alignment

| KSI Category | Aegis Readiness | Evidence |
|--------------|-----------------|----------|
| **Access Control** | ✅ Strong | OIDC/SAML, MFA, RBAC |
| **Awareness & Training** | ⚠️ Partial | Solo founder, needs formalization |
| **Audit & Accountability** | ✅ Strong | Structured logging, audit trails |
| **Configuration Management** | ✅ Strong | IaC, GitOps, immutable infra |
| **Identification & Auth** | ✅ Strong | Keycloak, MFA support |
| **Incident Response** | ✅ Documented | ISO 27001 IRP |
| **Risk Assessment** | ✅ Strong | ISO 27001 risk register |
| **System & Info Protection** | ⚠️ Gap | FIPS crypto needed |
| **Supply Chain** | ⚠️ Gap | SCRMP needed |

### 20x Automation Requirements

| Requirement | Status | Notes |
|-------------|--------|-------|
| Machine-readable documentation | ⚠️ Partial | Need OSCAL conversion |
| Automated evidence collection | ✅ Yes | Monthly evidence scripts |
| Continuous validation | ⚠️ Partial | Need real-time monitoring |
| Version-controlled configs | ✅ Yes | GitOps-based |

---

## Authorization Boundary

### System Boundary Diagram (Textual)

```
┌──────────────────────────────────────────────────────────────────────┐
│                        AUTHORIZATION BOUNDARY                         │
│                        (FedRAMP CSO: Aegis Platform)                  │
│                                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                    AEGIS PLATFORM COMPONENTS                     │  │
│  │                                                                   │  │
│  │   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐           │  │
│  │   │  Platform   │   │  Backstage  │   │  Workspace  │           │  │
│  │   │    API      │   │     UI      │   │    Proxy    │           │  │
│  │   └──────┬──────┘   └──────┬──────┘   └──────┬──────┘           │  │
│  │          │                 │                 │                   │  │
│  │   ┌──────┴─────────────────┴─────────────────┴──────┐           │  │
│  │   │              Kubernetes Control Plane            │           │  │
│  │   └──────────────────────┬───────────────────────────┘           │  │
│  │                          │                                        │  │
│  │   ┌──────────────────────┴───────────────────────────┐           │  │
│  │   │              Identity Provider (Keycloak)         │           │  │
│  │   └──────────────────────────────────────────────────┘           │  │
│  │                                                                   │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                    SUPPORTING INFRASTRUCTURE                      │  │
│  │   (Inherited from FedRAMP-Authorized IaaS - e.g., AWS GovCloud)  │  │
│  │                                                                   │  │
│  │   • Compute (EKS, EC2)     • Network (VPC, ALB)                  │  │
│  │   • Storage (S3, EBS)      • Database (RDS)                      │  │
│  │   • Logging (CloudWatch)   • Monitoring (CloudTrail)             │  │
│  │                                                                   │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                        │
└──────────────────────────────────────────────────────────────────────┘

                         EXTERNAL CONNECTIONS

    ┌─────────────┐         ┌─────────────┐         ┌─────────────┐
    │   GitHub    │         │    AWS      │         │  Customer   │
    │  (CI/CD)    │←───────→│  GovCloud   │←───────→│   IdP       │
    │             │  TLS    │  IaaS       │  SAML   │  (External) │
    └─────────────┘         └─────────────┘         └─────────────┘
```

### Boundary Components

| Component | In Boundary? | Responsibility |
|-----------|--------------|----------------|
| Platform API | ✅ Yes | Aegis |
| Backstage UI | ✅ Yes | Aegis |
| Workspace Proxy | ✅ Yes | Aegis |
| Keycloak | ✅ Yes | Aegis |
| K8s Control Plane | ✅ Yes | Aegis config, IaaS operation |
| AWS GovCloud | ⚠️ Inherited | AWS (FedRAMP-authorized) |
| GitHub Actions | ⚠️ External | GitHub (SOC 2/FedRAMP) |
| Customer IdP | ❌ External | Customer |
| Customer Workloads | ❌ External | Customer |

---

## Data Flow Categories

### Data Flow 1: User Authentication

```
User → Backstage UI → Keycloak → Customer IdP (SAML/OIDC) → JWT → API
```

| Data Element | Classification | Protection |
|--------------|----------------|------------|
| Username | Low | TLS in transit |
| Email | Low | TLS in transit |
| Password | N/A | Handled by external IdP |
| JWT Token | Moderate | TLS, short-lived, signed |

### Data Flow 2: Workload Submission

```
User → API → Kubernetes → GPU Node → Workload Execution
```

| Data Element | Classification | Protection |
|--------------|----------------|------------|
| Workload spec | Low | TLS, RBAC |
| Resource requests | Low | TLS |
| Container image ref | Low | TLS |
| Execution logs | Customer-controlled | Customer boundary |

### Data Flow 3: Administrative Actions

```
Admin → API → Configuration Change → Audit Log
```

| Data Element | Classification | Protection |
|--------------|----------------|------------|
| Config changes | Moderate | TLS, audit logging |
| Admin credentials | Moderate | MFA, short sessions |
| Audit logs | Moderate | Immutable storage |

---

## Impact Level Recommendation

### Recommendation: LOW Impact (LI-SaaS or 20x Low)

| Factor | Assessment |
|--------|------------|
| Data sensitivity | Low (minimal PII, no CUI) |
| Customer deployment | Self-hosted (Model B) |
| Aegis data access | Platform only, no customer workloads |
| Existing compliance | SOC 2 + ISO 27001 foundation |
| Time to market | Fastest path to federal customers |

### Conditions for Moderate

Upgrade to FedRAMP Moderate if:
- [ ] Customer requires Aegis to process CUI
- [ ] Aegis operates managed services (Model C)
- [ ] DoD customer requires IL4+ authorization
- [ ] Sensitive PII beyond login credentials

---

## Next Steps

1. **Finalize scope document** - Review with stakeholders
2. **Create boundary diagram** - Visual representation
3. **Begin SSP Section 1-3** - System identification, scope, info types
4. **FIPS crypto assessment** - Inventory current cryptography
5. **Engage potential agency sponsor** - Start federal customer conversations

---

## Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| CEO/Founder | Carlos Sanchez | _____________ | ______ |
| Authorizing Official (AO) | TBD (Agency) | _____________ | ______ |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial FIPS 199 assessment |
