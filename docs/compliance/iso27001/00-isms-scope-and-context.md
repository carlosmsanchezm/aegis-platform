# ISMS Scope and Context

**Document ID:** ISMS-SCOPE-001
**Version:** 1.1
**Effective Date:** 2026-01-17
**Next Review:** 2027-01-17
**Owner:** Carlos Sanchez (Information Security Manager)
**Classification:** Internal

---

## 1. Purpose

This document defines the scope of the Information Security Management System (ISMS) and establishes the context of the organization as required by ISO/IEC 27001:2022 Clauses 4.1, 4.2, and 4.3.

---

## 2. Context of the Organization (Clause 4.1)

### 2.1 External Context

| Factor | Description | Impact on ISMS |
|--------|-------------|----------------|
| **Regulatory** | SOC 2, FedRAMP ecosystem, FIPS 140-2 | Security controls must support customer compliance |
| **Market** | Government/Defense contractors, enterprise AI/ML | High security expectations |
| **Technology** | Cloud-native, Kubernetes, GPU workloads | Secure SDLC, supply chain security |
| **Threat landscape** | Supply chain attacks, credential theft | Dependency scanning, MFA enforcement |

### 2.2 Internal Context

| Factor | Description | Impact on ISMS |
|--------|-------------|----------------|
| **Size** | Solo founder (Carlos Sanchez) | Compensating controls for segregation of duties |
| **Structure** | Fully remote | Physical security N/A, endpoint focus |
| **Resources** | Limited budget | Prioritize automation, cloud-native security |
| **Culture** | Security-first, engineering-driven | Strong technical controls |

---

## 3. Interested Parties (Clause 4.2)

| Interested Party | Needs and Expectations | ISMS Requirements |
|------------------|----------------------|-------------------|
| **Customers** | Secure software, compliance evidence | SOC 2 + ISO 27001 certification |
| **Regulators** | Compliance with standards | Documented ISMS, audit evidence |
| **GitHub** | Secure code hosting | Branch protection, access control |
| **AWS** | Proper use of cloud services | IAM policies, encryption, logging |
| **Employees (future)** | Secure work environment | Policies, training, clear procedures |
| **Founder** | Business success, reputation | Balanced security/usability |

---

## 4. ISMS Scope (Clause 4.3)

### 4.1 Scope Statement

The ISMS applies to:

> **The secure development, delivery, and maintenance of the Aegis Platform - a multi-cluster GPU workload orchestration system - including all corporate systems, software development lifecycle processes, and supporting infrastructure operated by Aegis Technologies.**

### 4.2 Service Model

| Model | Description | ISMS Scope |
|-------|-------------|------------|
| **Model B (Current)** | Self-hosted software vendor | ✅ SDLC, corporate systems |
| Model C (Future) | Managed service | Will expand scope |

### 4.3 Systems in Scope

| System | Purpose | Owner | Location |
|--------|---------|-------|----------|
| GitHub (3 repos) | Source control, CI/CD | carlosmsanchezm | github.com |
| AWS GovCloud | Development infrastructure | Carlos Sanchez | us-east-1 |
| Developer workstations | Development | Carlos Sanchez | Remote |
| Google Workspace (planned) | Corporate identity | Carlos Sanchez | Cloud |

### 4.4 Repositories in Scope

| Repository | Purpose | Commits | Security Controls |
|------------|---------|---------|-------------------|
| carlosmsanchezm/aegis-platform | Platform backend | 465 | Branch protection, Dependabot, CODEOWNERS |
| carlosmsanchezm/aegis-ui | Frontend UI | 99 | Branch protection, Dependabot |
| carlosmsanchezm/sovran | Infrastructure/IaC | 161 | Branch protection, Dependabot |

### 4.5 Out of Scope

| Item | Reason |
|------|--------|
| Customer-deployed instances | Model B - customers operate their own |
| Customer data | Processed in customer environments |
| Physical facilities | Fully remote, cloud-hosted |
| Production operations | Model B - no managed service |

### 4.6 Interfaces and Dependencies

| Interface | Description | Security Responsibility |
|-----------|-------------|------------------------|
| GitHub → AWS | CI/CD deployment | Aegis (OIDC tokens, IAM roles) |
| Developer → GitHub | Code commits | Aegis (MFA, SSH keys) |
| Customers → Aegis releases | Software delivery | Aegis (signed releases, SBOM) |

---

## 5. Information Security Management System (Clause 4.4)

Aegis Technologies has established, implemented, and maintains an ISMS in accordance with ISO/IEC 27001:2022 to:

1. Protect the confidentiality, integrity, and availability of information
2. Meet customer requirements for security assurance
3. Comply with regulatory and contractual obligations
4. Continually improve information security performance

### 5.1 ISMS Components

| Component | Document | Status |
|-----------|----------|--------|
| Context & Scope | This document | ✅ Complete |
| Leadership & Roles | 01-leadership-policy-roles.md | ✅ Complete |
| Risk Assessment | 02-risk-methodology.md + 03-risk-register.csv | ✅ Complete |
| Risk Treatment | 04-risk-treatment-plan.csv | ✅ Complete |
| Statement of Applicability | 05-statement-of-applicability.csv | ✅ Complete |
| Policies | soc2/policies/ + iso27001/policies/ | ✅ Complete |
| Internal Audit | 07-internal-audit-program.md | ⚠️ Program exists, audit NOT conducted |
| Management Review | 08-management-review-template.md | ⚠️ Template exists, review NOT conducted |
| Corrective Actions | 09-corrective-actions-log.csv | ⚠️ Log exists, no entries yet |

---

## 6. Contact with Authorities (A.5.5)

### 6.1 Emergency Contacts

| Authority | When to Contact | Contact Method |
|-----------|-----------------|----------------|
| FBI IC3 | Cyber crime, data breach | ic3.gov |
| CISA | Critical infrastructure threat | cisa.gov/report |
| AWS Security | AWS account compromise | AWS Support (Enterprise) |
| GitHub Security | Repository compromise | security@github.com |
| Legal Counsel | Regulatory notification | TBD |

### 6.2 Regulatory Contacts

| Regulator | Jurisdiction | Notification Trigger |
|-----------|--------------|---------------------|
| FTC | US - consumer data | PII breach affecting consumers |
| State AG | US - state residents | Breach affecting state residents |
| Sector regulators | DoD/Gov customers | Per contract requirements |

### 6.3 Security Community (A.5.6)

| Organization | Purpose | Engagement |
|--------------|---------|------------|
| CISA KEV | Known exploited vulnerabilities | Monitor feed |
| NVD/CVE | Vulnerability database | Dependabot integration |
| MITRE ATT&CK | Threat intelligence | Reference for IR |
| Cloud Security Alliance | Best practices | Guidance |

---

## 7. Solo Founder Considerations

As a solo-founder company, the following adaptations apply:

| ISO Requirement | Standard Approach | Aegis Adaptation | Compensating Control |
|-----------------|-------------------|------------------|---------------------|
| Segregation of duties | Different people for dev/review | Single person | PR audit trail, CloudTrail logging |
| Internal audit | Internal audit team | External auditor | Independence maintained |
| Management review | Board/exec committee | Founder self-review | Formal documentation required |
| Access reviews | Manager certifies reports | Self-certification | Documented quarterly |

---

## 8. Scope Boundaries Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                        ISMS SCOPE BOUNDARY                          │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │   GitHub     │  │  AWS GovCloud│  │  Developer   │             │
│  │   (3 repos)  │  │  (Dev/Test)  │  │  Workstation │             │
│  └──────────────┘  └──────────────┘  └──────────────┘             │
│         │                 │                 │                      │
│         └────────┬────────┴────────┬────────┘                      │
│                  │                 │                               │
│           ┌──────▼──────┐   ┌──────▼──────┐                       │
│           │   CI/CD     │   │  Corporate   │                       │
│           │  Pipeline   │   │   Systems    │                       │
│           └──────┬──────┘   └──────────────┘                       │
│                  │                                                  │
│           ┌──────▼──────┐                                          │
│           │   Software  │                                          │
│           │  Releases   │                                          │
│           └──────┬──────┘                                          │
│                  │                                                  │
└──────────────────┼──────────────────────────────────────────────────┘
                   │
           ════════╧════════  SCOPE BOUNDARY  ════════
                   │
           ┌───────▼───────┐
           │   CUSTOMER    │  ← OUT OF SCOPE
           │  ENVIRONMENT  │
           └───────────────┘
```

---

## 9. Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Information Security Manager | Carlos Sanchez | /s/ Carlos Sanchez | 2026-01-17 |
| Top Management | Carlos Sanchez | /s/ Carlos Sanchez | 2026-01-17 |

---

## 10. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.1 | 2026-01-17 | Carlos Sanchez | Added Clause 4.1, 4.2 content; authority contacts |
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial ISMS scope |
