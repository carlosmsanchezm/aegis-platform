# FedRAMP System Security Plan (SSP) Outline

**Document ID:** FEDRAMP-SSP-OUTLINE-001
**Version:** 1.0
**Created:** 2026-01-17
**Target Baseline:** FedRAMP Low / LI-SaaS

---

## Purpose

This document outlines the structure and content requirements for the Aegis Platform System Security Plan (SSP). The SSP is the foundational document for FedRAMP authorization and describes how security controls are implemented.

---

## SSP Main Document Structure

### Section 1: System Identification

| Subsection | Content | Source |
|------------|---------|--------|
| 1.1 System Name and Identifier | Aegis Platform | New |
| 1.2 System Categorization | LOW (FIPS 199) | 00-scope-and-impact-assessment.md |
| 1.3 System Owner | Aegis Technologies | New |
| 1.4 Authorizing Official | TBD (Agency) | Pending |
| 1.5 Other Key Contacts | Carlos Sanchez (ISSO) | New |
| 1.6 Assignment of Security Responsibility | Roles and responsibilities | ISO 27001 01-leadership-policy-roles.md |

### Section 2: System Description

| Subsection | Content | Source |
|------------|---------|--------|
| 2.1 System Function/Purpose | GPU workload orchestration | 00-scope-and-impact-assessment.md |
| 2.2 Information Processed | Minimal PII (login only) | 00-scope-and-impact-assessment.md |
| 2.3 System Architecture | High-level architecture | Architecture docs |
| 2.4 Network Architecture | Network diagram | New (boundary diagram) |
| 2.5 Data Flows | Authentication, workload submission | 00-scope-and-impact-assessment.md |
| 2.6 Ports, Protocols, Services | PPS table | New |

### Section 3: System Environment

| Subsection | Content | Source |
|------------|---------|--------|
| 3.1 Authorization Boundary | Boundary diagram | New |
| 3.2 Deployed Location | AWS GovCloud (customer) | New |
| 3.3 System Interconnections | External connections table | New |
| 3.4 Leveraged Authorizations | AWS GovCloud FedRAMP | New |

### Section 4: Control Implementation

| Subsection | Content | Source |
|------------|---------|--------|
| 4.1 Control Implementation Summary | Overview | New |
| 4.2 Appendix A Reference | Full control details | Appendix A |

---

## Required Appendices

### Appendix A: Control Implementation Statements

**Purpose:** Detailed description of how each security control is implemented.

**Format:** FedRAMP-provided template with control-by-control implementation.

**Content for Each Control:**
- Control ID and Title
- Implementation Status (Implemented, Partially Implemented, Planned, N/A)
- Responsible Role
- Implementation Details
- Customer Responsibility (if applicable)

**Reuse from Existing:**
- SOC 2 control-matrix.csv → Map to 800-53
- ISO 27001 statement-of-applicability.csv → Cross-reference
- 01-control-mapping.csv → Gap analysis

**Effort:** HIGH - Most time-consuming appendix

---

### Appendix B: Related Acronyms

**Purpose:** Define acronyms used throughout the SSP.

**Status:** ✅ Can be generated from existing docs

**Effort:** LOW

---

### Appendix C: Security Policies and Procedures

**Purpose:** Reference policies supporting control implementation.

**Existing Policies (Reusable):**
| Policy | ID | SOC 2/ISO Source |
|--------|-----|------------------|
| Information Security Policy | ISP-001 | soc2/policies/ |
| Access Control Policy | ACP-001 | soc2/policies/ |
| Change Management Policy | CMP-001 | soc2/policies/ |
| Incident Response Policy | IRP-001 | soc2/policies/ |
| Risk Management Policy | RMP-001 | soc2/policies/ |
| Vendor Management Policy | VMP-001 | soc2/policies/ |
| Acceptable Use Policy | AUP-001 | iso27001/policies/ |

**New Policies Needed:**
- System and Communications Protection Policy
- System and Information Integrity Policy
- Contingency Planning Policy
- Supply Chain Risk Management Policy

**Effort:** MEDIUM - Modify existing + create new

---

### Appendix D: User Guide

**Purpose:** Instructions for system users on security features.

**Content:**
- Account management procedures
- Password requirements
- MFA setup
- Acceptable use reminders
- Reporting security incidents

**Effort:** MEDIUM

---

### Appendix E: Digital Identity Worksheet

**Purpose:** Determine Identity Assurance Level (IAL), Authenticator Assurance Level (AAL), and Federation Assurance Level (FAL) per SP 800-63.

**Required Analysis:**
| Level | Aegis Requirement | Justification |
|-------|-------------------|---------------|
| IAL | 1 (Self-asserted) | Minimal PII, no proofing required |
| AAL | 2 (MFA) | Sensitive system access |
| FAL | 2 (Encrypted assertions) | SAML/OIDC federation |

**Effort:** MEDIUM

---

### Appendix F: Rules of Behavior

**Purpose:** User agreement on acceptable system use.

**Content:**
- Government resources statement
- Acceptable use rules
- Prohibited activities
- Monitoring notice
- Acknowledgment signature block

**Reuse:** Acceptable Use Policy + FedRAMP template

**Effort:** LOW

---

### Appendix G: Information System Contingency Plan (ISCP)

**Purpose:** Procedures for system recovery after disruption.

**Content:**
- Contingency planning roles
- Activation procedures
- Recovery procedures
- Reconstitution procedures
- Testing and training
- Plan maintenance

**RTO/RPO for Aegis:**
| Tier | RTO | RPO | Components |
|------|-----|-----|------------|
| Critical | 4 hours | 1 hour | Platform API, Auth |
| Important | 8 hours | 4 hours | UI, Proxy |
| Supportive | 24 hours | 24 hours | Logging, Monitoring |

**Effort:** HIGH

---

### Appendix H: Configuration Management Plan

**Purpose:** Document configuration management processes.

**Content:**
- Configuration management roles
- Baseline configuration
- Change control process (GitOps)
- Configuration monitoring
- Deviation handling

**Reuse:** Change Management Policy + GitOps docs

**Effort:** MEDIUM

---

### Appendix I: Incident Response Plan

**Purpose:** Procedures for detecting, responding to, and recovering from incidents.

**Content:**
- IR team roles and responsibilities
- Incident categories and priorities
- Detection and analysis procedures
- Containment and eradication
- Recovery procedures
- Post-incident activities
- Reporting requirements (FedRAMP-specific)

**Reuse:** ISO 27001 Incident Response Policy

**FedRAMP Additions:**
- US-CERT reporting requirements
- Agency notification procedures
- FedRAMP PMO notification

**Effort:** MEDIUM

---

### Appendix J: Control Implementation Summary (CIS) Workbook

**Purpose:** Summary view of all controls with implementation status.

**Format:** FedRAMP CIS template (Excel/CSV)

**Content:**
| Column | Description |
|--------|-------------|
| Control ID | NIST 800-53 control |
| Control Title | Control name |
| Implementation Status | Implemented/Partial/Planned/NA |
| Responsible Entity | CSP/Customer/Shared/Inherited |
| Implementation Description | Brief summary |

**Effort:** MEDIUM (generate from Appendix A)

---

### Appendix K: FIPS 140 Cryptographic Modules

**Purpose:** Document validated cryptographic modules in use.

**Required Content:**
| Module | Version | FIPS Certificate | Use |
|--------|---------|------------------|-----|
| BoringCrypto | TBD | TBD | Go services TLS |
| AWS KMS | N/A | AWS-validated | Key management |
| Keycloak FIPS | TBD | TBD | Token signing |

**Current Gap:** No FIPS modules in use

**Effort:** HIGH (implementation required)

---

### Appendix L: Separation of Duties Matrix

**Purpose:** Document role-based access and separation of duties.

**Format:**
| Function | Role 1 | Role 2 | Role 3 |
|----------|--------|--------|--------|
| Code Commit | Dev | - | - |
| Code Review | - | Reviewer | - |
| Deploy | - | - | Ops |

**Solo Founder Note:** Document compensating controls where separation not possible.

**Effort:** MEDIUM

---

### Appendix M: Integrated Inventory Workbook

**Purpose:** Complete system component inventory.

**Content:**
- Hardware inventory (cloud instances)
- Software inventory (applications, libraries)
- Network inventory (VPCs, subnets)
- Data inventory

**Reuse:** SBOM + Terraform state + architecture docs

**Effort:** MEDIUM

---

### Appendix N: Continuous Monitoring Plan

**Purpose:** Document ongoing security monitoring activities.

**Content:**
| Activity | Frequency | Tool | Deliverable |
|----------|-----------|------|-------------|
| Vulnerability Scanning | Monthly | Dependabot/Grype | ConMon report |
| Configuration Assessment | Monthly | Terraform | Drift report |
| Access Review | Quarterly | Manual | Access review |
| Security Control Assessment | Annual | 3PAO | SAR |
| Penetration Testing | Annual | 3PAO/vendor | Pen test report |

**Monthly ConMon Deliverables:**
- Vulnerability scan results
- POA&M updates
- Significant change report
- Incident report (if any)

**Effort:** MEDIUM

---

### Appendix O: Plan of Action and Milestones (POA&M)

**Purpose:** Track remediation of identified weaknesses.

**Format:** FedRAMP POA&M template

**Content:**
| Column | Description |
|--------|-------------|
| POA&M ID | Unique identifier |
| Weakness | Description of finding |
| Source | Assessment, scan, etc. |
| Control ID | Related NIST control |
| Risk Level | Low/Moderate/High |
| Remediation | Planned actions |
| Resources | Required resources |
| Scheduled Completion | Target date |
| Status | Open/Closed |
| Milestone Changes | History |

**SLA Requirements:**
- High: 30 days
- Moderate: 90 days
- Low: 180 days

**Reuse:** ISO 27001 09-corrective-actions-log.csv (adapt format)

**Effort:** LOW (process exists)

---

### Appendix P: Supply Chain Risk Management Plan (SCRMP)

**Purpose:** Document supply chain risk management per SP 800-161.

**Content:**
- Supply chain risk assessment process
- Supplier evaluation criteria
- Component verification procedures
- SBOM management
- Vulnerability response for dependencies
- Supplier incident notification

**Reuse:** Vendor Management Policy + SBOM

**Effort:** MEDIUM

---

### Appendix Q: Cryptographic Inventory

**Purpose:** Detailed inventory of all cryptographic implementations.

**Content:**
| Component | Algorithm | Key Size | Purpose | FIPS Module |
|-----------|-----------|----------|---------|-------------|
| TLS | TLS 1.2/1.3 | 256-bit | Transport | BoringCrypto |
| mTLS | X.509 | 2048-bit RSA | Service auth | BoringCrypto |
| Token Signing | RS256 | 2048-bit RSA | JWT | Keycloak FIPS |
| Storage | AES-256-GCM | 256-bit | Data at rest | AWS KMS |

**Effort:** MEDIUM

---

## SSP Development Timeline

### Phase 1: Foundation (Weeks 1-4)

| Week | Task | Deliverable |
|------|------|-------------|
| 1 | SSP Sections 1-2 | System identification, description |
| 2 | SSP Section 3 | Environment, boundary |
| 3 | Appendix E (Digital Identity) | IAL/AAL/FAL assessment |
| 4 | Appendix H (Config Mgmt) | CM plan |

### Phase 2: Controls (Weeks 5-10)

| Week | Task | Deliverable |
|------|------|-------------|
| 5-6 | Appendix A (Controls - AC, AT, AU) | Access, Audit families |
| 7-8 | Appendix A (Controls - CA, CM, CP) | Assessment, Config, Contingency |
| 9-10 | Appendix A (Controls - IA, IR, MA) | Identity, Incident, Maintenance |

### Phase 3: Appendices (Weeks 11-16)

| Week | Task | Deliverable |
|------|------|-------------|
| 11 | Appendix A (Controls - remaining) | All control families |
| 12 | Appendix G (ISCP) | Contingency plan |
| 13 | Appendix I (IRP) | Incident response |
| 14 | Appendix N (ConMon) | Continuous monitoring |
| 15 | Appendix P (SCRMP) | Supply chain |
| 16 | Appendix K, Q (Crypto) | Cryptographic docs |

### Phase 4: Review (Weeks 17-20)

| Week | Task | Deliverable |
|------|------|-------------|
| 17-18 | Internal review | Quality check |
| 19 | OSCAL conversion | Machine-readable |
| 20 | Readiness assessment prep | Final SSP package |

---

## Templates and Resources

### FedRAMP Official Templates

| Template | URL |
|----------|-----|
| SSP Template | https://www.fedramp.gov/resources/templates/ |
| POA&M Template | https://www.fedramp.gov/resources/templates/ |
| CIS Workbook | https://www.fedramp.gov/resources/templates/ |
| Inventory Workbook | https://www.fedramp.gov/resources/templates/ |

### Aegis Source Documents

| Document | Location | Use |
|----------|----------|-----|
| SOC 2 Control Matrix | soc2/control-matrix.csv | Appendix A input |
| ISO 27001 SoA | iso27001/05-statement-of-applicability.csv | Appendix A input |
| Risk Register | iso27001/03-risk-register.csv | RA-3 evidence |
| Policies | soc2/policies/ | Appendix C |
| Architecture | Various | Section 2 |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial SSP outline |
