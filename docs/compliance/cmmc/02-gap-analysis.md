# CMMC Gap Analysis

**Document ID:** CMMC-GAP-001
**Version:** 1.0
**Created:** 2026-03-09
**Target:** CMMC Level 2 (NIST 800-171 Rev 3)

---

## Executive Summary

This gap analysis identifies the delta between Aegis Platform's current compliance posture (SOC 2 Type II + ISO 27001) and CMMC Level 2 requirements. CMMC Level 2 maps directly to NIST SP 800-171, which contains 97 active requirements in Rev 3 (expanded and reorganized from the 110 requirements in Rev 2).

The analysis shows that approximately **47% of controls are fully implemented**, with an additional **26% partially addressed** through existing SOC 2 and ISO 27001 programs. Physical and environmental controls (11%) are inherited from AWS as the IaaS provider. The remaining **16%** represent gaps requiring new implementation or assessment.

### Gap Summary by Priority

| Priority | Gap Count | Effort | Timeline |
|----------|-----------|--------|----------|
| Critical | 4 | High | Must address before C3PAO assessment |
| High | 6 | Medium-High | Address in Phase 2 |
| Medium | 8 | Medium | Address in Phase 3 |
| Low | 7 | Low | Minor documentation or inherited |

### Status Summary (97 Active Controls)

| Status | Count | Percentage |
|--------|-------|------------|
| Implemented | 46 | 47% |
| Partial | 25 | 26% |
| Inherited/N/A | 11 | 11% |
| Not Implemented | 6 | 6% |
| Not Assessed | 9 | 9% |

---

## Methodology

### Approach

1. **Control Extraction:** Parsed the OSCAL NIST 800-171 Rev 3 catalog (`docs/oscal/nist_sp_800_171_rev3_catalog.json`) to extract all 97 active controls (33 withdrawn controls excluded).

2. **Cross-Reference Mapping:** Mapped each 800-171 control to its parent NIST 800-53 control, then cross-referenced against:
   - FedRAMP control mapping (`docs/compliance/fedramp/01-control-mapping.csv`) for SOC 2 and ISO 27001 overlap
   - Customer control implementation statements (`docs/compliance/customer-docs/control-implementation-statements.md`) for 17 implemented controls
   - FedRAMP gap analysis (`docs/compliance/fedramp/02-gap-analysis.md`) for shared gaps

3. **Status Assessment:** Each control assessed as:
   - **Implemented:** Fully addressed by existing controls with evidence
   - **Partial:** Partially addressed; needs CMMC-specific documentation or enhancements
   - **Not Implemented:** Not currently addressed; needs new implementation
   - **Not Assessed:** Requires further analysis to determine status
   - **Inherited/N/A:** Inherited from AWS IaaS or not applicable to cloud-hosted platform

4. **Gap Prioritization:** Gaps prioritized based on:
   - Impact on SPRS score (weight of control)
   - Criticality for CUI protection
   - Effort required for remediation
   - Cross-framework benefit (shared with FedRAMP)

---

## Critical Gaps (Blockers for Level 2 Certification)

### CMMC-GAP-001: FIPS 140-2/3 Cryptographic Modules

| Attribute | Value |
|-----------|-------|
| **Priority** | Critical |
| **Controls Affected** | 03.13.10 (Cryptographic Key Establishment), 03.13.11 (Cryptographic Protection) |
| **CMMC Domain** | System and Communications Protection (SC) |
| **Current State** | Standard cryptographic libraries (Go stdlib, OpenSSL) |
| **Required State** | FIPS 140-2/3 validated cryptographic modules for all CUI processing |
| **Impact** | Certification blocker -- DoD mandate for CUI protection |

**Gap Details:**
- Go standard library crypto is NOT FIPS validated
- Container base images use standard OpenSSL (not FIPS-validated)
- Keycloak default configuration not FIPS-compliant
- TLS certificates generated with non-FIPS tools (step-ca)
- No cryptographic module inventory document

**Remediation:**
1. Replace Go crypto with BoringCrypto (FIPS-validated) using `GOEXPERIMENT=boringcrypto`
2. Rebuild all Go services (platform-api, proxy, k8s-agent) with FIPS-enabled crypto
3. Use FIPS-validated base images (Red Hat UBI FIPS or similar)
4. Configure Keycloak for FIPS mode
5. Document cryptographic module inventory
6. Validate all cipher suites are FIPS-approved

**Effort:** High (2-4 weeks development + testing)
**Cost:** $0 (BoringCrypto is free) + engineering time
**FedRAMP Overlap:** Shared gap (FED-GAP-001). Single remediation benefits both programs.
**Reuse Potential:** Low (new implementation required)

---

### CMMC-GAP-002: System Security Plan (SSP)

| Attribute | Value |
|-----------|-------|
| **Priority** | Critical |
| **Controls Affected** | 03.15.01 (Policy and Procedures), 03.15.02 (System Security Plan) |
| **CMMC Domain** | Planning (PL) |
| **Current State** | Policies and procedures exist but not in SSP format |
| **Required State** | Complete SSP documenting all 97 control implementations |
| **Impact** | Certification blocker -- C3PAO requires SSP as primary assessment artifact |

**Gap Details:**
- No consolidated SSP document exists
- Existing policies not mapped to 800-171 requirements
- No OSCAL machine-readable version
- Control implementation statements exist for only 17 of 97 controls
- Missing system boundary description per CMMC scoping guidance

**Remediation:**
1. Create SSP main document using CMMC/NIST 800-171 template
2. Document system boundary and CUI data flows
3. Write implementation descriptions for all 97 controls
4. Map existing evidence to each control requirement
5. Create POA&M for unmet requirements
6. Convert to OSCAL format (recommended but not required for CMMC)

**Effort:** High (4-8 weeks documentation)
**Cost:** $0-$20K (consultant optional for review)
**FedRAMP Overlap:** Shared gap (FED-GAP-002). SSP structure can serve both CMMC and FedRAMP.
**Reuse Potential:** Medium (existing policies provide foundation)

---

### CMMC-GAP-003: Supply Chain Risk Management Program

| Attribute | Value |
|-----------|-------|
| **Priority** | Critical |
| **Controls Affected** | 03.17.01 (SCRM Plan), 03.17.02 (Acquisition Strategies), 03.17.03 (Supply Chain Requirements) |
| **CMMC Domain** | Supply Chain Risk Management (SR) |
| **Current State** | Vendor management policy exists; SBOM generated; no formal SCRM plan |
| **Required State** | Complete SCRM program per SP 800-161 |
| **Impact** | Three controls scored as Not Implemented or Partial |

**Gap Details:**
- No formal Supply Chain Risk Management Plan (SCRMP) document
- SBOM exists but not integrated into supply chain risk assessment
- No component authenticity verification process
- No supply chain risk assessment methodology
- No formal acquisition security requirements

**Remediation:**
1. Create SCRMP following SP 800-161 guidance
2. Document supply chain risk assessment methodology
3. Integrate SBOM generation into CI/CD pipeline with automated checks
4. Create supplier assessment checklist and scoring criteria
5. Document acquisition security requirements for third-party components
6. Establish supply chain incident notification process

**Effort:** Medium (2-3 weeks documentation + process)
**Cost:** $0 (internal effort)
**FedRAMP Overlap:** Shared gap (FED-GAP-006). Single SCRMP serves both programs.
**Reuse Potential:** Medium (existing vendor policy and SBOM provide foundation)

---

### CMMC-GAP-004: Personnel Screening / Background Checks

| Attribute | Value |
|-----------|-------|
| **Priority** | Critical |
| **Controls Affected** | 03.09.01 (Personnel Screening) |
| **CMMC Domain** | Personnel Security (PS) |
| **Current State** | No background checks conducted |
| **Required State** | Personnel screening commensurate with CUI access level |
| **Impact** | Fundamental CUI protection requirement |

**Gap Details:**
- No background check process established
- No position risk designations defined
- No personnel screening policy document
- Solo founder -- limited staff but still requires documented process
- No contractor screening requirements

**Remediation:**
1. Define position risk designations for all roles with CUI access
2. Implement background check process (minimum: criminal history, identity verification)
3. Document screening requirements for contractors and third parties
4. Establish re-screening schedule (e.g., every 5 years)
5. Maintain screening records

**Effort:** Medium (process + vendor setup)
**Cost:** $100-$500/person + vendor setup
**FedRAMP Overlap:** Shared gap (FED-GAP-007).
**Reuse Potential:** Low (new process required)

---

## High Priority Gaps

### CMMC-GAP-005: Incident Response Testing

| Attribute | Value |
|-----------|-------|
| **Priority** | High |
| **Controls Affected** | 03.06.03 (Incident Response Testing) |
| **CMMC Domain** | Incident Response (IR) |
| **Current State** | Incident Response Plan exists but has not been tested |
| **Required State** | Annual incident response testing (tabletop or functional) |

**Remediation:**
1. Schedule tabletop exercise with realistic CUI breach scenario
2. Create test scenario documentation
3. Conduct exercise and record observations
4. Document lessons learned and remediation actions
5. Update IRP based on findings
6. Schedule recurring annual test

**Effort:** Low (1-2 days)
**Cost:** $0-$5K (external facilitator optional)
**FedRAMP Overlap:** Shared gap (FED-GAP-012).
**Reuse Potential:** High (existing IRP provides foundation)

---

### CMMC-GAP-006: System Use Notification (Login Banner)

| Attribute | Value |
|-----------|-------|
| **Priority** | High |
| **Controls Affected** | 03.01.09 (System Use Notification) |
| **CMMC Domain** | Access Control (AC) |
| **Current State** | No login banner displayed |
| **Required State** | System use notification before granting access |

**Remediation:**
1. Implement login banner in Keycloak login page
2. Implement consent acknowledgment before access
3. Use DoD-approved banner text for CUI systems
4. Test banner display across all access methods
5. Document banner configuration

**Effort:** Low (1-2 days development)
**Cost:** $0
**FedRAMP Overlap:** Shared gap (FED-GAP-010).
**Reuse Potential:** High (single implementation serves all frameworks)

---

### CMMC-GAP-007: Security Awareness Training Records

| Attribute | Value |
|-----------|-------|
| **Priority** | High |
| **Controls Affected** | 03.02.01 (Literacy Training and Awareness), 03.02.02 (Role-Based Training) |
| **CMMC Domain** | Awareness and Training (AT) |
| **Current State** | Informal security awareness; no formal training records |
| **Required State** | Documented security awareness training with completion records |

**Remediation:**
1. Select or create CUI-specific security awareness training
2. Implement role-based training for system administrators and developers
3. Track training completion dates and maintain records
4. Require annual refresher training
5. Document training content and delivery method

**Effort:** Low-Medium (1-2 weeks)
**Cost:** $0-$1K (free DoD CUI training available via CDSE)
**FedRAMP Overlap:** Shared gap (FED-GAP-013).
**Reuse Potential:** Medium (existing SOC 2 awareness program can be extended)

---

### CMMC-GAP-008: Rules of Behavior

| Attribute | Value |
|-----------|-------|
| **Priority** | High |
| **Controls Affected** | 03.15.03 (Rules of Behavior) |
| **CMMC Domain** | Planning (PL) |
| **Current State** | Acceptable use policy exists but not formalized as Rules of Behavior |
| **Required State** | Formal Rules of Behavior with user acknowledgment |

**Remediation:**
1. Create Rules of Behavior document covering CUI handling
2. Define user responsibilities for CUI protection
3. Implement acknowledgment tracking (digital signature)
4. Include in onboarding process
5. Require annual re-acknowledgment

**Effort:** Low (2-3 days)
**Cost:** $0
**FedRAMP Overlap:** Shared gap (FED-GAP-011).
**Reuse Potential:** High (existing acceptable use policy provides foundation)

---

### CMMC-GAP-009: Audit Record Review Process

| Attribute | Value |
|-----------|-------|
| **Priority** | High |
| **Controls Affected** | 03.03.04 (Response to Audit Failures), 03.03.05 (Audit Review and Reporting), 03.03.06 (Audit Reduction and Report Generation) |
| **CMMC Domain** | Audit and Accountability (AU) |
| **Current State** | Monthly review process exists; no alerting on audit failures; no automated reduction |
| **Required State** | Automated alerting on audit failures; regular review with documented findings |

**Remediation:**
1. Implement alerting for audit logging process failures (CloudWatch/Prometheus)
2. Create audit review schedule and documented review process
3. Implement log aggregation and reduction tools (ELK/CloudWatch Insights)
4. Generate periodic audit summary reports
5. Document audit review findings and actions taken

**Effort:** Medium (1-2 weeks)
**Cost:** $0-$5K/year (tool costs)
**Reuse Potential:** High (existing logging infrastructure provides foundation)

---

### CMMC-GAP-010: Cryptographic Key Management Documentation

| Attribute | Value |
|-----------|-------|
| **Priority** | High |
| **Controls Affected** | 03.13.10 (Cryptographic Key Establishment and Management) |
| **CMMC Domain** | System and Communications Protection (SC) |
| **Current State** | step-ca and cert-manager deployed; key management not fully documented |
| **Required State** | Documented key management plan with lifecycle procedures |

**Remediation:**
1. Document cryptographic key inventory (TLS certs, JWT signing keys, mTLS CAs)
2. Document key generation, distribution, storage, and destruction procedures
3. Define key rotation schedules and automate where possible
4. Document key recovery procedures
5. Create crypto module inventory with FIPS validation status

**Effort:** Medium (1-2 weeks documentation)
**Cost:** $0
**FedRAMP Overlap:** Related to FED-GAP-001.
**Reuse Potential:** Medium (existing PKI infrastructure provides foundation)

---

## Medium Priority Gaps

| ID | Gap | Controls | Current State | Required State | Effort | Reuse |
|----|-----|----------|---------------|----------------|--------|-------|
| CMMC-GAP-011 | Separation of Duties | 03.01.04 | Solo founder | Documented compensating controls | Low | Medium |
| CMMC-GAP-012 | Information Flow Enforcement Documentation | 03.01.03 | Network policies in place | Documented network segmentation with CUI flow diagrams | Medium | Medium |
| CMMC-GAP-013 | Device Lock Policy | 03.01.10 | Configurable session timeout | Documented endpoint lock policy for CUI access devices | Low | Medium |
| CMMC-GAP-014 | Mobile Device Policy | 03.01.18 | No formal policy | Written mobile device management policy for CUI access | Low | Medium |
| CMMC-GAP-015 | External System Policy | 03.01.20 | Partial policy | Formal external system access agreements | Low | Medium |
| CMMC-GAP-016 | System Component Inventory | 03.04.10, 03.04.11 | Partial inventory via IaC | Formal component inventory with CUI boundary marking | Medium | Medium |
| CMMC-GAP-017 | Malicious Code Protection | 03.14.02 | Cloud workload scanning | Documented anti-malware strategy for containerized workloads | Medium | Medium |
| CMMC-GAP-018 | Information Management and Retention | 03.14.08 | Partial data handling docs | CUI-specific data retention and disposal policy | Low | Medium |

---

## Low Priority Gaps (Documentation or Not Assessed)

| ID | Gap | Controls | Notes | Effort |
|----|-----|----------|-------|--------|
| CMMC-GAP-019 | Device Identification and Authentication | 03.05.02 | Needs assessment; Kubernetes node authentication may satisfy | Low |
| CMMC-GAP-020 | Maintenance Tools | 03.07.04 | Cloud-hosted; limited applicability | Low |
| CMMC-GAP-021 | Media Storage | 03.08.01 | Digital media only; cloud storage | Low |
| CMMC-GAP-022 | Media Marking | 03.08.04 | Digital-only; CUI marking in metadata | Low |
| CMMC-GAP-023 | Media Transport | 03.08.05 | TLS for all transport; needs documentation | Low |
| CMMC-GAP-024 | Mobile Code | 03.13.13 | Web application security; needs assessment | Low |
| CMMC-GAP-025 | Session Authenticity | 03.13.15 | TLS/mTLS in place; needs formal documentation | Low |

---

## Inherited/Not Applicable Controls (No Action Required)

The following 11 controls are inherited from AWS as the IaaS provider or not applicable to the cloud-hosted Aegis Platform:

| Control | Title | Rationale |
|---------|-------|-----------|
| 03.01.16 | Wireless Access | Cloud-hosted; no wireless infrastructure managed by Aegis |
| 03.07.05 | Nonlocal Maintenance | IaaS inherited; AWS manages hardware maintenance |
| 03.07.06 | Maintenance Personnel | IaaS inherited; AWS manages maintenance personnel |
| 03.08.02 | Media Access | Cloud-hosted; no physical media managed by Aegis |
| 03.08.03 | Media Sanitization | IaaS inherited; AWS handles media sanitization |
| 03.10.01 | Physical Access Authorizations | IaaS inherited; AWS data center security |
| 03.10.02 | Monitoring Physical Access | IaaS inherited; AWS data center monitoring |
| 03.10.06 | Alternate Work Site | Remote-first; covered by endpoint and remote access policies |
| 03.10.07 | Physical Access Control | IaaS inherited; AWS data center access control |
| 03.10.08 | Access Control for Transmission | IaaS inherited; AWS manages physical transmission security |
| 03.13.12 | Collaborative Computing Devices | No collaborative computing devices in scope |

---

## Reusable Assets from SOC 2 / ISO 27001

### High Reuse Potential (directly applicable to CMMC)

| Asset | Source Framework | CMMC Controls Served | Modification Needed |
|-------|-----------------|----------------------|---------------------|
| Information Security Policy | ISO 27001 | 03.15.01 | Add CUI-specific language |
| Access Control Policy | SOC 2 | 03.01.01-03.01.22 | Add 800-171 references |
| Change Management Policy | SOC 2 | 03.04.01-03.04.06 | Minimal |
| Incident Response Plan | ISO 27001 | 03.06.01-03.06.05 | Add CUI breach procedures |
| Risk Register (17 risks) | ISO 27001 | 03.11.01, 03.11.04 | Add supply chain risks |
| Risk Treatment Plan | ISO 27001 | 03.11.04 | Minimal |
| Evidence Collection Scripts | SOC 2 | 03.12.03 | Add CUI-specific evidence |
| Corrective Actions Log | ISO 27001 | 03.12.02 | Rename to POA&M format |
| RBAC Implementation | SOC 2/Aegis | 03.01.02, 03.01.05-07 | Document for SSP |
| Keycloak OIDC/MFA | Aegis | 03.05.01, 03.05.03, 03.05.04 | Document for SSP |
| Structured Audit Logging | Aegis | 03.03.01-03.03.08 | Document for SSP |
| GitOps Change Control | Aegis | 03.04.03-03.04.05 | Document for SSP |

### Medium Reuse Potential (needs adaptation)

| Asset | Source | CMMC Controls | Modifications Needed |
|-------|--------|---------------|----------------------|
| Quarterly Access Review | SOC 2 | 03.01.01 | Add CUI-specific review criteria |
| Vulnerability Scanning | SOC 2 | 03.11.02 | Add CUI scope documentation |
| Employee Onboarding | SOC 2 | 03.09.01, 03.09.02 | Add background check step |
| Employee Offboarding | SOC 2 | 03.09.02 | Add CUI access termination SLA |
| SBOM | SOC 2 | 03.17.01-03 | Integrate into SCRMP |
| Network Segmentation | Aegis | 03.01.03, 03.13.01, 03.13.06 | Document CUI data flows |

### Low Reuse Potential (mostly new work)

| Asset | Source | Gap | New Work |
|-------|--------|-----|----------|
| Architecture Docs | Internal | CUI boundary diagram | Draw CMMC scoping boundary |
| Policies | Various | 800-171 format | Reformat with 800-171 references |
| Evidence | SOC 2 | CMMC format | Adapt evidence collection for CMMC |

---

## FedRAMP Overlap Analysis

Many CMMC gaps are shared with the FedRAMP program. Addressing these once eliminates duplicate effort:

| CMMC Gap | FedRAMP Gap | Shared Work | Estimated Savings |
|----------|------------|-------------|-------------------|
| CMMC-GAP-001 (FIPS Crypto) | FED-GAP-001 | Identical remediation | 100% -- implement once |
| CMMC-GAP-002 (SSP) | FED-GAP-002 | SSP structure, 80% content overlap | 60-70% |
| CMMC-GAP-003 (SCRMP) | FED-GAP-006 | Identical document | 100% |
| CMMC-GAP-004 (Background Checks) | FED-GAP-007 | Identical process | 100% |
| CMMC-GAP-005 (IR Testing) | FED-GAP-012 | Identical exercise | 100% |
| CMMC-GAP-006 (Login Banner) | FED-GAP-010 | Identical implementation | 100% |
| CMMC-GAP-007 (Training) | FED-GAP-013 | Identical program | 100% |
| CMMC-GAP-008 (Rules of Behavior) | FED-GAP-011 | Identical document | 100% |

**Estimated total effort savings from FedRAMP overlap:** 40-50% reduction in CMMC-specific work.

---

## Gap Remediation Roadmap

### Recommended Strategy: Level 1 Now, Level 2 When Triggered

#### Phase 1: Quick Wins from SOC 2 / ISO 27001 Reuse (Weeks 1-4)

Address gaps that can be closed by documenting existing implementations.

| Week | Task | Gap IDs | Cost |
|------|------|---------|------|
| 1 | Implement login banner in Keycloak | CMMC-GAP-006 | $0 |
| 1-2 | Create Rules of Behavior document | CMMC-GAP-008 | $0 |
| 2 | Document separation of duties compensating controls | CMMC-GAP-011 | $0 |
| 2-3 | Formalize mobile device and external system policies | CMMC-GAP-014, 015 | $0 |
| 3-4 | Document device lock and endpoint policies | CMMC-GAP-013 | $0 |
| 4 | Conduct incident response tabletop exercise | CMMC-GAP-005 | $0-$5K |
| 4 | Complete Level 1 self-assessment | -- | $0 |

**Phase 1 Cost:** $0-$5K
**Controls Addressed:** ~10

#### Phase 2: Technical Gaps (Weeks 5-12)

Address gaps requiring technical implementation or significant documentation.

| Week | Task | Gap IDs | Cost |
|------|------|---------|------|
| 5-6 | Implement security awareness training program (CDSE CUI training) | CMMC-GAP-007 | $0-$1K |
| 6-8 | Implement FIPS 140-2/3 cryptographic modules (BoringCrypto) | CMMC-GAP-001 | $0 (dev time) |
| 7-8 | Document cryptographic key management procedures | CMMC-GAP-010 | $0 |
| 8-9 | Implement audit failure alerting and review process | CMMC-GAP-009 | $0-$5K |
| 9-10 | Create system component inventory with CUI boundary | CMMC-GAP-016 | $0 |
| 10-11 | Document information flow enforcement with CUI data flow diagrams | CMMC-GAP-012 | $0 |
| 11-12 | Document malicious code protection strategy for containers | CMMC-GAP-017 | $0 |

**Phase 2 Cost:** $0-$6K + engineering time
**Controls Addressed:** ~15

#### Phase 3: Process Gaps (Weeks 13-20)

Address gaps requiring new organizational processes.

| Week | Task | Gap IDs | Cost |
|------|------|---------|------|
| 13-14 | Implement personnel screening / background check process | CMMC-GAP-004 | $100-$500/person |
| 14-16 | Create Supply Chain Risk Management Plan | CMMC-GAP-003 | $0 |
| 16-18 | Create System Security Plan (SSP) | CMMC-GAP-002 | $0-$20K |
| 18-19 | Create CUI data retention and disposal policy | CMMC-GAP-018 | $0 |
| 19-20 | Assess and document remaining "Not Assessed" controls | CMMC-GAP-019-025 | $0 |
| 20 | Calculate and submit SPRS score | -- | $0 |

**Phase 3 Cost:** $500-$21K
**Controls Addressed:** ~15

#### Phase 4: C3PAO Assessment (Weeks 21-32, Trigger: DoD Customer)

| Week | Task | Cost |
|------|------|------|
| 21-22 | Select C3PAO (get 3-5 quotes) | $0 |
| 23-24 | C3PAO readiness review (optional) | $15K-$30K |
| 25-28 | Remediate readiness findings | $0 |
| 29-32 | C3PAO Level 2 certification assessment | $50K-$150K |
| 32 | Submit SPRS score update | $0 |

**Phase 4 Cost:** $65K-$180K

---

## Cost Summary

| Phase | Cost Range | Timing |
|-------|-----------|--------|
| Phase 1: Quick Wins | $0-$5K | Do now |
| Phase 2: Technical Gaps | $0-$6K + dev time | Do now (shared with FedRAMP) |
| Phase 3: Process Gaps | $500-$21K | When DoD customer identified |
| Phase 4: C3PAO Assessment | $65K-$180K | When DoD contract requires Level 2 |
| **Total (through certification)** | **$65K-$212K** | **6-8 months from trigger** |

### Ongoing Annual Costs

| Item | Estimate | Notes |
|------|----------|-------|
| SPRS score maintenance | $0 | Internal effort |
| Annual self-assessment (Level 1) | $0 | Internal effort |
| Evidence collection | $0 | Automated scripts |
| Training program | $0-$1K | Free CDSE training |
| Triennial Level 2 re-assessment | $50K-$150K | Every 3 years (amortized: $17K-$50K/year) |
| **Total Annual** | **$17K-$51K** | |

---

## Gap Status Tracking

| ID | Gap | Priority | Status | Owner | Due |
|----|-----|----------|--------|-------|-----|
| CMMC-GAP-001 | FIPS Crypto | Critical | Not Started | Carlos | Phase 2 |
| CMMC-GAP-002 | System Security Plan | Critical | Not Started | Carlos | Phase 3 |
| CMMC-GAP-003 | Supply Chain RMP | Critical | Not Started | Carlos | Phase 3 |
| CMMC-GAP-004 | Personnel Screening | Critical | Not Started | Carlos | Phase 3 |
| CMMC-GAP-005 | IR Testing | High | Not Started | Carlos | Phase 1 |
| CMMC-GAP-006 | Login Banner | High | Not Started | Carlos | Phase 1 |
| CMMC-GAP-007 | Training Records | High | Not Started | Carlos | Phase 2 |
| CMMC-GAP-008 | Rules of Behavior | High | Not Started | Carlos | Phase 1 |
| CMMC-GAP-009 | Audit Review Process | High | Not Started | Carlos | Phase 2 |
| CMMC-GAP-010 | Key Mgmt Documentation | High | Not Started | Carlos | Phase 2 |
| CMMC-GAP-011 | Separation of Duties | Medium | Not Started | Carlos | Phase 1 |
| CMMC-GAP-012 | Info Flow Documentation | Medium | Not Started | Carlos | Phase 2 |
| CMMC-GAP-013 | Device Lock Policy | Medium | Not Started | Carlos | Phase 1 |
| CMMC-GAP-014 | Mobile Device Policy | Medium | Not Started | Carlos | Phase 1 |
| CMMC-GAP-015 | External System Policy | Medium | Not Started | Carlos | Phase 1 |
| CMMC-GAP-016 | Component Inventory | Medium | Not Started | Carlos | Phase 2 |
| CMMC-GAP-017 | Malicious Code Protection | Medium | Not Started | Carlos | Phase 2 |
| CMMC-GAP-018 | Data Retention Policy | Medium | Not Started | Carlos | Phase 3 |
| CMMC-GAP-019 | Device Authentication | Low | Not Assessed | Carlos | Phase 3 |
| CMMC-GAP-020 | Maintenance Tools | Low | Not Assessed | Carlos | Phase 3 |
| CMMC-GAP-021 | Media Storage | Low | Not Assessed | Carlos | Phase 3 |
| CMMC-GAP-022 | Media Marking | Low | Not Assessed | Carlos | Phase 3 |
| CMMC-GAP-023 | Media Transport | Low | Not Assessed | Carlos | Phase 3 |
| CMMC-GAP-024 | Mobile Code | Low | Not Assessed | Carlos | Phase 3 |
| CMMC-GAP-025 | Session Authenticity | Low | Not Assessed | Carlos | Phase 3 |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-09 | Carlos Sanchez | Initial CMMC gap analysis |
