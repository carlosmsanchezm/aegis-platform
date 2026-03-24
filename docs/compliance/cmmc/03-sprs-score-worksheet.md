# SPRS Score Worksheet

**Document ID:** CMMC-SPRS-001
**Version:** 1.0
**Created:** 2026-03-09
**Assessment Date:** 2026-03-09
**Assessor:** Carlos Sanchez (Self-Assessment)
**Assessment Type:** Preliminary / Pre-Submission

---

## Summary

| Metric | Value |
|--------|-------|
| **Total NIST 800-171 Rev 3 Active Controls** | 97 |
| **Maximum Possible SPRS Score** | 110 |
| **Minimum Possible SPRS Score** | -203 |
| **Current Estimated SPRS Score** | **38** |
| **Controls Fully Met (Score = 5)** | 46 |
| **Controls Partially Met (Score = 3)** | 25 |
| **Controls Planned/Documented (Score = 1)** | 6 |
| **Controls Not Met (Score = 0)** | 9 |
| **Controls Inherited/N/A (Score = 5)** | 11 |
| **Controls Requiring POA&M** | 15 (all scoring below 5) |

### Score Breakdown by Category

| Category | Count | Points Contributed | Points Lost |
|----------|-------|--------------------|-------------|
| Fully Implemented (5/5) | 46 | 0 lost | 0 |
| Inherited/N/A (5/5) | 11 | 0 lost | 0 |
| Partially Met (3/5) | 25 | -2 each = -50 | -50 |
| Planned/Documented (1/5) | 6 | -4 each = -24 | -24 |
| Not Met (0/5) | 9 | varies by weight | varies |
| **Subtotal points lost** | | | **see detail below** |

### Score Calculation

The SPRS score starts at 110 and subtracts points for each unmet requirement. Each of the 97 active NIST 800-171 requirements has a point value (weight) of 1, 3, or 5 points. When a requirement is not met, its point value is subtracted from the score. Partial implementations reduce the subtraction.

**Scoring methodology used in this worksheet:**
- **5 = Fully Implemented:** No points subtracted
- **3 = Partially Implemented:** Weight subtracted at 40% (weight * 0.4, rounded)
- **1 = Planned/Documented:** Weight subtracted at 80% (weight * 0.8, rounded)
- **0 = Not Implemented:** Full weight subtracted

---

## Detailed Control Scoring

### Family 03.01: Access Control (AC)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.01.01 | Account Management | 5 | 5 | 0 | Implemented | No | Keycloak account management, RBAC, audit logging |
| 03.01.02 | Access Enforcement | 5 | 5 | 0 | Implemented | No | RBAC via Keycloak, namespace isolation |
| 03.01.03 | Information Flow Enforcement | 5 | 3 | 2 | Partial | Yes | Network policies exist; need CUI flow documentation |
| 03.01.04 | Separation of Duties | 1 | 3 | 0 | Partial | Yes | Solo founder; compensating controls needed |
| 03.01.05 | Least Privilege | 5 | 5 | 0 | Implemented | No | RBAC with minimal permissions |
| 03.01.06 | Least Privilege - Privileged Accounts | 5 | 5 | 0 | Implemented | No | Separate admin/user roles |
| 03.01.07 | Least Privilege - Privileged Functions | 3 | 5 | 0 | Implemented | No | Privileged functions audited |
| 03.01.08 | Unsuccessful Logon Attempts | 3 | 5 | 0 | Implemented | No | Keycloak brute force detection |
| 03.01.09 | System Use Notification | 1 | 0 | 1 | Not Implemented | Yes | Login banner not implemented |
| 03.01.10 | Device Lock | 1 | 3 | 0 | Partial | Yes | Session timeout exists; endpoint policy needed |
| 03.01.11 | Session Termination | 3 | 5 | 0 | Implemented | No | Configurable session timeout |
| 03.01.12 | Remote Access | 5 | 5 | 0 | Implemented | No | TLS/mTLS for all remote access |
| 03.01.16 | Wireless Access | 1 | 5 | 0 | Inherited/N/A | No | Cloud-hosted; no wireless infrastructure |
| 03.01.18 | Access Control for Mobile Devices | 1 | 3 | 0 | Partial | Yes | Need mobile device policy |
| 03.01.20 | Use of External Systems | 1 | 3 | 0 | Partial | Yes | Need external system policy |
| 03.01.22 | Publicly Accessible Content | 1 | 5 | 0 | Implemented | No | No public content hosting |

**AC Subtotal:** 16 controls, 3 points lost

---

### Family 03.02: Awareness and Training (AT)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.02.01 | Literacy Training and Awareness | 3 | 3 | 1 | Partial | Yes | Informal awareness; needs formal program |
| 03.02.02 | Role-Based Training | 3 | 3 | 1 | Partial | Yes | Need specialized training records |

**AT Subtotal:** 2 controls, 2 points lost

---

### Family 03.03: Audit and Accountability (AU)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.03.01 | Event Logging | 5 | 5 | 0 | Implemented | No | Structured logging for all security events |
| 03.03.02 | Audit Record Content | 3 | 5 | 0 | Implemented | No | Rich audit log format with all required fields |
| 03.03.03 | Audit Record Generation | 3 | 5 | 0 | Implemented | No | Application-level audit logging |
| 03.03.04 | Response to Audit Logging Process Failures | 3 | 3 | 1 | Partial | Yes | Need alerting on audit failures |
| 03.03.05 | Audit Record Review, Analysis, and Reporting | 3 | 3 | 1 | Partial | Yes | Monthly review exists; needs formalization |
| 03.03.06 | Audit Record Reduction and Report Generation | 1 | 3 | 0 | Partial | Yes | Need automated report generation |
| 03.03.07 | Time Stamps | 1 | 5 | 0 | Implemented | No | NTP synchronized, UTC ISO 8601 |
| 03.03.08 | Protection of Audit Information | 3 | 5 | 0 | Implemented | No | Immutable cloud log storage |

**AU Subtotal:** 8 controls, 2 points lost

---

### Family 03.04: Configuration Management (CM)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.04.01 | Baseline Configuration | 5 | 5 | 0 | Implemented | No | IaC baselines (Terraform, Helm) |
| 03.04.02 | Configuration Settings | 3 | 5 | 0 | Implemented | No | Documented in values files |
| 03.04.03 | Configuration Change Control | 5 | 5 | 0 | Implemented | No | GitOps with PR review and approval |
| 03.04.04 | Impact Analyses | 3 | 5 | 0 | Implemented | No | PR review process with CI checks |
| 03.04.05 | Access Restrictions for Change | 3 | 5 | 0 | Implemented | No | Branch protection, CODEOWNERS |
| 03.04.06 | Least Functionality | 3 | 5 | 0 | Implemented | No | Minimal container images, distroless |
| 03.04.08 | Authorized Software - Allow by Exception | 3 | 5 | 0 | Implemented | No | Container image allow-listing |
| 03.04.10 | System Component Inventory | 3 | 3 | 1 | Partial | Yes | IaC provides partial inventory; needs formal doc |
| 03.04.11 | Information Location | 1 | 3 | 0 | Partial | Yes | Need CUI data location documentation |
| 03.04.12 | System and Component Configuration for High-Risk Areas | 3 | 5 | 0 | Implemented | No | Hardened configurations |

**CM Subtotal:** 10 controls, 1 point lost

---

### Family 03.05: Identification and Authentication (IA)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.05.01 | User Identification and Authentication | 5 | 5 | 0 | Implemented | No | OIDC via Keycloak, unique user IDs |
| 03.05.02 | Device Identification and Authentication | 3 | 1 | 2 | Not Assessed | Yes | Needs assessment; K8s node auth may satisfy |
| 03.05.03 | Multi-Factor Authentication | 5 | 5 | 0 | Implemented | No | MFA via Keycloak, AMR claim validation |
| 03.05.04 | Replay-Resistant Authentication | 3 | 5 | 0 | Implemented | No | OIDC token-based; replay-resistant by design |
| 03.05.05 | Identifier Management | 3 | 5 | 0 | Implemented | No | Unique identifiers, lifecycle management |
| 03.05.07 | Password Management | 3 | 5 | 0 | Implemented | No | Keycloak password policy enforcement |
| 03.05.11 | Authentication Feedback | 1 | 5 | 0 | Implemented | No | Masked password feedback |
| 03.05.12 | Authenticator Management | 5 | 5 | 0 | Implemented | No | Password policy, token expiration |

**IA Subtotal:** 8 controls, 2 points lost

---

### Family 03.06: Incident Response (IR)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.06.01 | Incident Handling | 5 | 5 | 0 | Implemented | No | Incident handling procedure documented |
| 03.06.02 | Incident Monitoring, Reporting, and Response Assistance | 3 | 3 | 1 | Partial | Yes | Monitoring exists; reporting process partial |
| 03.06.03 | Incident Response Testing | 3 | 0 | 3 | Not Implemented | Yes | No IR testing conducted |
| 03.06.04 | Incident Response Training | 3 | 3 | 1 | Partial | Yes | Solo founder; needs documentation |
| 03.06.05 | Incident Response Plan | 5 | 5 | 0 | Implemented | No | IRP documented and approved |

**IR Subtotal:** 5 controls, 5 points lost

---

### Family 03.07: Maintenance (MA)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.07.04 | Maintenance Tools | 1 | 1 | 1 | Not Assessed | Yes | Cloud-hosted; limited applicability; needs assessment |
| 03.07.05 | Nonlocal Maintenance | 3 | 5 | 0 | Inherited/N/A | No | IaaS inherited from AWS |
| 03.07.06 | Maintenance Personnel | 1 | 5 | 0 | Inherited/N/A | No | IaaS inherited from AWS |

**MA Subtotal:** 3 controls, 1 point lost

---

### Family 03.08: Media Protection (MP)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.08.01 | Media Storage | 3 | 1 | 2 | Not Assessed | Yes | Digital media only; needs policy |
| 03.08.02 | Media Access | 1 | 5 | 0 | Inherited/N/A | No | Cloud; no physical media |
| 03.08.03 | Media Sanitization | 3 | 5 | 0 | Inherited/N/A | No | IaaS inherited from AWS |
| 03.08.04 | Media Marking | 1 | 1 | 1 | Not Assessed | Yes | Digital-only; needs CUI marking policy |
| 03.08.05 | Media Transport | 1 | 1 | 1 | Not Assessed | Yes | TLS in place; needs formal documentation |
| 03.08.07 | Media Use | 1 | 5 | 0 | Implemented | No | No removable media policy |
| 03.08.09 | System Backup - Cryptographic Protection | 3 | 5 | 0 | Implemented | No | Encrypted cloud backups |

**MP Subtotal:** 7 controls, 4 points lost

---

### Family 03.09: Personnel Security (PS)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.09.01 | Personnel Screening | 5 | 0 | 5 | Not Implemented | Yes | No background checks conducted |
| 03.09.02 | Personnel Termination and Transfer | 3 | 3 | 1 | Partial | Yes | Offboarding exists; solo founder limitation |

**PS Subtotal:** 2 controls, 6 points lost

---

### Family 03.10: Physical Protection (PE)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.10.01 | Physical Access Authorizations | 3 | 5 | 0 | Inherited/N/A | No | AWS data center security |
| 03.10.02 | Monitoring Physical Access | 1 | 5 | 0 | Inherited/N/A | No | AWS data center monitoring |
| 03.10.06 | Alternate Work Site | 1 | 5 | 0 | Inherited/N/A | No | Remote-first; VPN/TLS policies |
| 03.10.07 | Physical Access Control | 5 | 5 | 0 | Inherited/N/A | No | AWS data center access control |
| 03.10.08 | Access Control for Transmission | 3 | 5 | 0 | Inherited/N/A | No | AWS physical transmission security |

**PE Subtotal:** 5 controls, 0 points lost

---

### Family 03.11: Risk Assessment (RA)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.11.01 | Risk Assessment | 5 | 5 | 0 | Implemented | No | ISO 27001 risk register with 17 risks |
| 03.11.02 | Vulnerability Monitoring and Scanning | 5 | 5 | 0 | Implemented | No | Dependabot, container scanning |
| 03.11.04 | Risk Response | 3 | 5 | 0 | Implemented | No | Risk treatment plan documented |

**RA Subtotal:** 3 controls, 0 points lost

---

### Family 03.12: Security Assessment and Monitoring (CA)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.12.01 | Security Assessment | 3 | 3 | 1 | Partial | Yes | ISO internal audit pending |
| 03.12.02 | Plan of Action and Milestones | 3 | 5 | 0 | Implemented | No | Corrective actions log as POA&M |
| 03.12.03 | Continuous Monitoring | 5 | 5 | 0 | Implemented | No | Monthly evidence collection scripts |
| 03.12.05 | Information Exchange | 1 | 3 | 0 | Partial | Yes | Need interconnection agreements |

**CA Subtotal:** 4 controls, 1 point lost

---

### Family 03.13: System and Communications Protection (SC)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.13.01 | Boundary Protection | 5 | 5 | 0 | Implemented | No | Network segmentation, NetworkPolicy |
| 03.13.04 | Information in Shared System Resources | 3 | 1 | 2 | Not Assessed | Yes | Container isolation; needs formal doc |
| 03.13.06 | Network Communications - Deny by Default | 5 | 5 | 0 | Implemented | No | NetworkPolicy deny-by-default |
| 03.13.08 | Transmission and Storage Confidentiality | 5 | 5 | 0 | Implemented | No | TLS everywhere, encrypted storage |
| 03.13.09 | Network Disconnect | 1 | 1 | 1 | Not Assessed | Yes | Session timeout exists; needs documentation |
| 03.13.10 | Cryptographic Key Establishment and Management | 5 | 3 | 2 | Partial | Yes | step-ca deployed; key mgmt docs needed |
| 03.13.11 | Cryptographic Protection | 5 | 0 | 5 | Not Implemented | Yes | FIPS 140 gap -- critical |
| 03.13.12 | Collaborative Computing Devices | 1 | 5 | 0 | Inherited/N/A | No | Not applicable |
| 03.13.13 | Mobile Code | 1 | 1 | 1 | Not Assessed | Yes | Web app; needs assessment |
| 03.13.15 | Session Authenticity | 3 | 1 | 2 | Not Assessed | Yes | TLS provides session auth; needs doc |

**SC Subtotal:** 10 controls, 13 points lost

---

### Family 03.14: System and Information Integrity (SI)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.14.01 | Flaw Remediation | 5 | 5 | 0 | Implemented | No | Vulnerability remediation process |
| 03.14.02 | Malicious Code Protection | 5 | 3 | 2 | Partial | Yes | Container scanning; needs strategy doc |
| 03.14.03 | Security Alerts, Advisories, and Directives | 3 | 5 | 0 | Implemented | No | Dependabot alerts, security advisories |
| 03.14.06 | System Monitoring | 5 | 5 | 0 | Implemented | No | Prometheus, structured logging |
| 03.14.08 | Information Management and Retention | 1 | 3 | 0 | Partial | Yes | Data handling exists; needs CUI policy |

**SI Subtotal:** 5 controls, 2 points lost

---

### Family 03.15: Planning (PL)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.15.01 | Policy and Procedures | 5 | 5 | 0 | Implemented | No | Security program documentation |
| 03.15.02 | System Security Plan | 5 | 0 | 5 | Not Implemented | Yes | SSP not written |
| 03.15.03 | Rules of Behavior | 1 | 3 | 0 | Partial | Yes | AUP exists; needs formalization |

**PL Subtotal:** 3 controls, 5 points lost

---

### Family 03.16: System and Services Acquisition (SA)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.16.01 | Security Engineering Principles | 3 | 5 | 0 | Implemented | No | Security by design, threat modeling |
| 03.16.02 | Unsupported System Components | 3 | 3 | 1 | Partial | Yes | Need EOL tracking process |
| 03.16.03 | External System Services | 3 | 3 | 1 | Partial | Yes | Third-party inventory partial |

**SA Subtotal:** 3 controls, 2 points lost

---

### Family 03.17: Supply Chain Risk Management (SR)

| Control | Title | Weight | Score | Points Lost | Status | POA&M Required | Notes |
|---------|-------|--------|-------|-------------|--------|----------------|-------|
| 03.17.01 | Supply Chain Risk Management Plan | 5 | 0 | 5 | Not Implemented | Yes | SCRMP not created |
| 03.17.02 | Acquisition Strategies, Tools, and Methods | 3 | 3 | 1 | Partial | Yes | Need formal acquisition security strategy |
| 03.17.03 | Supply Chain Requirements and Processes | 3 | 3 | 1 | Partial | Yes | Vendor policy exists; needs SCRM alignment |

**SR Subtotal:** 3 controls, 7 points lost

---

## SPRS Score Calculation

### Points Lost by Family

| Family | Controls | Points Lost | Key Gaps |
|--------|----------|-------------|----------|
| AC (Access Control) | 16 | 3 | Login banner, policies |
| AT (Awareness and Training) | 2 | 2 | Training program |
| AU (Audit and Accountability) | 8 | 2 | Audit alerting, review process |
| CM (Configuration Management) | 10 | 1 | Component inventory |
| IA (Identification and Authentication) | 8 | 2 | Device authentication |
| IR (Incident Response) | 5 | 5 | IR testing |
| MA (Maintenance) | 3 | 1 | Maintenance tools assessment |
| MP (Media Protection) | 7 | 4 | Media policies |
| PS (Personnel Security) | 2 | 6 | Background checks |
| PE (Physical Protection) | 5 | 0 | All inherited |
| RA (Risk Assessment) | 3 | 0 | All implemented |
| CA (Security Assessment) | 4 | 1 | Security assessment |
| SC (System and Comms Protection) | 10 | 13 | FIPS crypto, key mgmt |
| SI (System and Info Integrity) | 5 | 2 | Malicious code protection |
| PL (Planning) | 3 | 5 | SSP |
| SA (System and Services Acquisition) | 3 | 2 | EOL tracking, third-party |
| SR (Supply Chain Risk Management) | 3 | 7 | SCRMP |
| **TOTAL** | **97** | **56** | |

### Final Score

```
SPRS Score = 110 - Total Points Lost
SPRS Score = 110 - 56
SPRS Score = 54 (estimated, subject to weighted recalculation)
```

**Note:** The SPRS scoring methodology applies specific point values (1, 3, or 5) per requirement, and the actual formula subtracts the full weight when a requirement is not met and a proportional amount for partial implementations. Using a simplified proportional model:

| Score Level | Formula | Count | Points Lost |
|-------------|---------|-------|-------------|
| Score = 5 (Met) | 0 per control | 57 (46 impl + 11 inherited) | 0 |
| Score = 3 (Partial) | Weight * 0.4 | 25 | ~28 |
| Score = 1 (Planned) | Weight * 0.8 | 6 | ~14 |
| Score = 0 (Not Met) | Full weight | 9 | ~30 |
| **Total points lost** | | | **~72** |

**Adjusted SPRS Score: 110 - 72 = 38**

The more conservative estimate of **38** accounts for the weighted scoring methodology. The actual score will be finalized after formal assessment of all "Not Assessed" controls and determination of precise weights.

---

## Controls Requiring POA&M

The following controls scored below 5 and require a Plan of Action and Milestones entry:

### Critical POA&M Items (Score = 0)

| Control | Title | Weight | Current Score | Target Score | Target Date | Remediation |
|---------|-------|--------|---------------|--------------|-------------|-------------|
| 03.01.09 | System Use Notification | 1 | 0 | 5 | Phase 1 | Implement login banner in Keycloak |
| 03.06.03 | Incident Response Testing | 3 | 0 | 5 | Phase 1 | Conduct tabletop exercise |
| 03.09.01 | Personnel Screening | 5 | 0 | 5 | Phase 3 | Implement background check process |
| 03.13.11 | Cryptographic Protection | 5 | 0 | 5 | Phase 2 | Implement FIPS 140 (BoringCrypto) |
| 03.15.02 | System Security Plan | 5 | 0 | 5 | Phase 3 | Create complete SSP |
| 03.17.01 | Supply Chain Risk Management Plan | 5 | 0 | 5 | Phase 3 | Create SCRMP |

### High-Priority POA&M Items (Score = 1)

| Control | Title | Weight | Current Score | Target Score | Target Date | Remediation |
|---------|-------|--------|---------------|--------------|-------------|-------------|
| 03.05.02 | Device Identification and Authentication | 3 | 1 | 5 | Phase 3 | Assess K8s node auth; document |
| 03.07.04 | Maintenance Tools | 1 | 1 | 5 | Phase 3 | Assess cloud maintenance tools |
| 03.08.01 | Media Storage | 3 | 1 | 5 | Phase 3 | Document digital media storage policy |
| 03.08.04 | Media Marking | 1 | 1 | 5 | Phase 3 | Document CUI marking for digital media |
| 03.08.05 | Media Transport | 1 | 1 | 5 | Phase 3 | Document TLS-based transport security |
| 03.13.04 | Information in Shared System Resources | 3 | 1 | 5 | Phase 3 | Document container isolation |
| 03.13.09 | Network Disconnect | 1 | 1 | 5 | Phase 3 | Document session timeout for network |
| 03.13.13 | Mobile Code | 1 | 1 | 5 | Phase 3 | Assess web application code security |
| 03.13.15 | Session Authenticity | 3 | 1 | 5 | Phase 3 | Document TLS session authenticity |

### Medium-Priority POA&M Items (Score = 3)

| Control | Title | Weight | Current Score | Target Score | Target Date | Remediation |
|---------|-------|--------|---------------|--------------|-------------|-------------|
| 03.01.03 | Information Flow Enforcement | 5 | 3 | 5 | Phase 2 | Document CUI data flow diagrams |
| 03.01.04 | Separation of Duties | 1 | 3 | 5 | Phase 1 | Document compensating controls |
| 03.01.10 | Device Lock | 1 | 3 | 5 | Phase 1 | Document endpoint lock policy |
| 03.01.18 | Access Control for Mobile Devices | 1 | 3 | 5 | Phase 1 | Create mobile device policy |
| 03.01.20 | Use of External Systems | 1 | 3 | 5 | Phase 1 | Create external system agreements |
| 03.02.01 | Literacy Training and Awareness | 3 | 3 | 5 | Phase 2 | Implement formal training program |
| 03.02.02 | Role-Based Training | 3 | 3 | 5 | Phase 2 | Document role-based training |
| 03.03.04 | Response to Audit Logging Process Failures | 3 | 3 | 5 | Phase 2 | Implement audit failure alerting |
| 03.03.05 | Audit Record Review, Analysis, and Reporting | 3 | 3 | 5 | Phase 2 | Formalize review process |
| 03.03.06 | Audit Record Reduction and Report Generation | 1 | 3 | 5 | Phase 2 | Automate report generation |
| 03.04.10 | System Component Inventory | 3 | 3 | 5 | Phase 2 | Create formal component inventory |
| 03.04.11 | Information Location | 1 | 3 | 5 | Phase 2 | Document CUI data locations |
| 03.06.02 | Incident Monitoring, Reporting, and Response Assistance | 3 | 3 | 5 | Phase 2 | Formalize incident reporting |
| 03.06.04 | Incident Response Training | 3 | 3 | 5 | Phase 2 | Document IR training |
| 03.09.02 | Personnel Termination and Transfer | 3 | 3 | 5 | Phase 1 | Formalize transfer procedures |
| 03.12.01 | Security Assessment | 3 | 3 | 5 | Phase 2 | Complete internal audit |
| 03.12.05 | Information Exchange | 1 | 3 | 5 | Phase 2 | Create interconnection agreements |
| 03.13.10 | Cryptographic Key Establishment and Management | 5 | 3 | 5 | Phase 2 | Document key management procedures |
| 03.14.02 | Malicious Code Protection | 5 | 3 | 5 | Phase 2 | Document container security strategy |
| 03.14.08 | Information Management and Retention | 1 | 3 | 5 | Phase 3 | Create CUI retention policy |
| 03.15.03 | Rules of Behavior | 1 | 3 | 5 | Phase 1 | Formalize Rules of Behavior |
| 03.16.02 | Unsupported System Components | 3 | 3 | 5 | Phase 2 | Implement EOL tracking |
| 03.16.03 | External System Services | 3 | 3 | 5 | Phase 2 | Complete third-party inventory |
| 03.17.02 | Acquisition Strategies, Tools, and Methods | 3 | 3 | 5 | Phase 3 | Formalize acquisition strategy |
| 03.17.03 | Supply Chain Requirements and Processes | 3 | 3 | 5 | Phase 3 | Align vendor policy with SCRM |

---

## Score Improvement Projections

### After Phase 1 (Quick Wins)

| Action | Controls | Score Impact |
|--------|----------|-------------|
| Login banner | 03.01.09 | +1 |
| IR tabletop exercise | 03.06.03 | +3 |
| Rules of Behavior | 03.15.03 | +1 |
| Compensating controls (SoD) | 03.01.04 | +1 |
| Device lock policy | 03.01.10 | +1 |
| Mobile device policy | 03.01.18 | +1 |
| External system policy | 03.01.20 | +1 |
| Personnel transfer docs | 03.09.02 | +1 |
| **Phase 1 Total** | **8 controls** | **+10 points (est. score: 48)** |

### After Phase 2 (Technical Gaps)

| Action | Controls | Score Impact |
|--------|----------|-------------|
| FIPS crypto | 03.13.11 | +5 |
| Training program | 03.02.01, 03.02.02 | +4 |
| Key mgmt docs | 03.13.10 | +2 |
| Audit alerting/review | 03.03.04, 03.03.05, 03.03.06 | +3 |
| Component inventory | 03.04.10, 03.04.11 | +2 |
| CUI data flow docs | 03.01.03 | +2 |
| Container security docs | 03.14.02 | +2 |
| Other technical items | Various | +8 |
| **Phase 2 Total** | **~15 controls** | **+28 points (est. score: 76)** |

### After Phase 3 (Process Gaps)

| Action | Controls | Score Impact |
|--------|----------|-------------|
| Background checks | 03.09.01 | +5 |
| SCRMP | 03.17.01 | +5 |
| SSP | 03.15.02 | +5 |
| Remaining assessments | Various | +19 |
| **Phase 3 Total** | **~15 controls** | **+34 points (est. score: 110)** |

### Score Progression Summary

| Phase | Estimated Score | Controls Resolved | Timeline |
|-------|-----------------|-------------------|----------|
| Current | 38 | -- | Now |
| After Phase 1 | 48 | 8 | +4 weeks |
| After Phase 2 | 76 | +15 | +12 weeks |
| After Phase 3 | 110 | +15 | +20 weeks |

---

## Submission Requirements

### SPRS Portal Submission

Before any DoD contract award, the SPRS score must be posted to the SPRS portal:

| Field | Value |
|-------|-------|
| **System Name** | Aegis Platform |
| **SPRS Score** | TBD (pending formal calculation) |
| **Assessment Date** | TBD |
| **Assessment Type** | Self-Assessment (Basic) |
| **Scope** | All NIST 800-171 requirements |
| **POA&M Included** | Yes (for all non-5 scores) |
| **Score Validity** | 3 years from assessment date |

### Prerequisites for Submission

- [ ] Complete formal self-assessment of all 97 controls
- [ ] Determine precise weights for each control per DoD Assessment Methodology
- [ ] Calculate final SPRS score
- [ ] Create POA&M for all unmet requirements with remediation timelines
- [ ] Obtain management approval of score
- [ ] Submit to SPRS portal (https://www.sprs.csd.disa.mil/)

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-09 | Carlos Sanchez | Initial SPRS score worksheet (preliminary) |
