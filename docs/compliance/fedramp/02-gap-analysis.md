# FedRAMP Gap Analysis

**Document ID:** FEDRAMP-GAP-001
**Version:** 1.0
**Created:** 2026-01-17
**Baseline:** FedRAMP Low / LI-SaaS / 20x Low

---

## Executive Summary

This gap analysis identifies the delta between Aegis's current compliance posture (SOC 2 + ISO 27001) and FedRAMP Low/LI-SaaS requirements. The analysis shows approximately **25-30% of controls can be directly reused** from existing programs, with significant gaps in FedRAMP-specific requirements.

### Gap Summary by Priority

| Priority | Gaps | Effort | Timeline |
|----------|------|--------|----------|
| 🔴 Critical | 6 | High | Must address before 3PAO |
| 🟡 High | 8 | Medium-High | Address in Phase 2 |
| 🟢 Medium | 12 | Medium | Address in Phase 3 |
| ⚪ Low | 15+ | Low | Minor documentation |

---

## Critical Gaps (Blockers)

### FED-GAP-001: FIPS 140-2/3 Cryptographic Modules

| Attribute | Value |
|-----------|-------|
| **Priority** | 🔴 Critical |
| **Controls Affected** | IA-7, SC-13, SC-12 |
| **Current State** | Standard cryptographic libraries (Go stdlib, OpenSSL) |
| **Required State** | FIPS 140-2/3 validated cryptographic modules |
| **Impact** | Authorization blocker - federal mandate |

**Gap Details:**
- Go standard library crypto is NOT FIPS validated
- Container base images use standard OpenSSL
- Keycloak default configuration not FIPS-compliant
- TLS certificates generated with non-FIPS tools

**Remediation:**
1. Replace Go crypto with BoringCrypto (FIPS-validated)
2. Rebuild Go services with `GOEXPERIMENT=boringcrypto`
3. Use FIPS-validated base images (e.g., Red Hat UBI FIPS)
4. Configure Keycloak FIPS mode
5. Document cryptographic module inventory

**Effort:** High (2-4 weeks development, testing)
**Cost:** $0 (BoringCrypto is free) + engineering time

---

### FED-GAP-002: System Security Plan (SSP)

| Attribute | Value |
|-----------|-------|
| **Priority** | 🔴 Critical |
| **Controls Affected** | PL-2, PL-10, PL-11 |
| **Current State** | Policies and procedures exist but not in SSP format |
| **Required State** | Complete FedRAMP SSP with all appendices |
| **Impact** | Cannot submit for authorization without SSP |

**Gap Details:**
- No SSP document exists
- Existing policies not mapped to 800-53 controls
- No OSCAL machine-readable format
- Missing required appendices (17+ appendices)

**Remediation:**
1. Create SSP main document using FedRAMP template
2. Complete Appendix A (control implementations)
3. Create all required appendices (see SSP outline)
4. Convert to OSCAL format
5. Internal review and quality check

**Effort:** High (4-8 weeks documentation)
**Cost:** $0-$20K (consultant optional)

---

### FED-GAP-003: Agency Sponsor / Authorization Path

| Attribute | Value |
|-----------|-------|
| **Priority** | 🔴 Critical |
| **Controls Affected** | CA-6 |
| **Current State** | No federal agency relationship |
| **Required State** | Agency sponsor or Program Authorization slot |
| **Impact** | Cannot proceed without authorization path |

**Gap Details:**
- No existing federal customers
- No agency sponsor identified
- No FedRAMP Connect submission
- Unknown demand for federal authorization

**Remediation:**
1. Identify target federal agencies (DoD, DOE, national labs)
2. Engage federal sales/BD efforts
3. Secure letter of intent from potential customer
4. Submit to FedRAMP Connect (if agency path)
5. Alternative: Apply for Program Authorization (competitive)

**Effort:** High (ongoing business development)
**Cost:** Variable (sales/BD)

---

### FED-GAP-004: Third-Party Assessment Organization (3PAO)

| Attribute | Value |
|-----------|-------|
| **Priority** | 🔴 Critical |
| **Controls Affected** | CA-2 |
| **Current State** | No 3PAO selected |
| **Required State** | FedRAMP-accredited 3PAO engaged |
| **Impact** | Cannot complete authorization without 3PAO |

**Gap Details:**
- No 3PAO research conducted
- No budget allocated for assessment
- No timeline for engagement

**Remediation:**
1. Research FedRAMP-accredited 3PAOs
2. Request quotes (3-5 3PAOs)
3. Select based on cost, timeline, experience
4. Engage for readiness assessment (optional)
5. Engage for full assessment

**Effort:** Medium (vendor selection)
**Cost:** $40K-$100K (assessment fees)

---

### FED-GAP-005: Continuous Monitoring Program

| Attribute | Value |
|-----------|-------|
| **Priority** | 🔴 Critical |
| **Controls Affected** | CA-7, SI-2, SI-4, RA-5 |
| **Current State** | Monthly evidence collection, ad-hoc monitoring |
| **Required State** | FedRAMP ConMon with monthly reporting |
| **Impact** | Required for ongoing authorization |

**Gap Details:**
- Monthly scans exist but not in FedRAMP format
- No ConMon plan document
- No monthly ConMon deliverables defined
- No POA&M SLA tracking (30/90/180 days)

**Remediation:**
1. Create Continuous Monitoring Plan (SSP Appendix N)
2. Implement monthly vulnerability scanning with FedRAMP format
3. Create ConMon reporting templates
4. Implement POA&M SLA tracking
5. Define monthly/annual deliverable schedule

**Effort:** Medium (process + tooling)
**Cost:** $5K-$20K/year (scanning tools)

---

### FED-GAP-006: Supply Chain Risk Management Plan

| Attribute | Value |
|-----------|-------|
| **Priority** | 🔴 Critical |
| **Controls Affected** | SR-1, SR-2, SR-3 |
| **Current State** | Vendor management policy exists |
| **Required State** | Full SCRMP per SP 800-161 |
| **Impact** | FedRAMP Rev 5 requirement |

**Gap Details:**
- No formal SCRMP document
- SBOM exists but not integrated into SCRMP
- No supply chain risk assessment process
- No component authenticity verification

**Remediation:**
1. Create SCRMP following SP 800-161 guidance
2. Document supply chain risk assessment process
3. Integrate SBOM into SCRMP
4. Document component verification process
5. Create supplier assessment checklist

**Effort:** Medium (documentation)
**Cost:** $0 (internal effort)

---

## High Priority Gaps

### FED-GAP-007: Personnel Security - Background Checks

| Attribute | Value |
|-----------|-------|
| **Priority** | 🟡 High |
| **Controls Affected** | PS-2, PS-3 |
| **Current State** | No background checks conducted |
| **Required State** | Background checks for personnel with system access |

**Remediation:**
1. Define position risk designations
2. Implement background check process
3. Document clearance requirements (if DoD)
4. Maintain personnel screening records

**Effort:** Medium
**Cost:** $100-$500/person

---

### FED-GAP-008: Digital Identity Assessment

| Attribute | Value |
|-----------|-------|
| **Priority** | 🟡 High |
| **Controls Affected** | IA-2, IA-8, IA-12 |
| **Current State** | No SP 800-63 assessment |
| **Required State** | IAL/AAL/FAL determination documented |

**Remediation:**
1. Complete Digital Identity Worksheet (SSP Appendix E)
2. Determine required assurance levels
3. Document authentication mechanisms
4. Verify compliance with selected levels

**Effort:** Medium
**Cost:** $0

---

### FED-GAP-009: Contingency Plan and Testing

| Attribute | Value |
|-----------|-------|
| **Priority** | 🟡 High |
| **Controls Affected** | CP-2, CP-3, CP-4 |
| **Current State** | Partial - cloud backups exist |
| **Required State** | Documented CP with annual testing |

**Remediation:**
1. Create Information System Contingency Plan (ISCP)
2. Define RTOs and RPOs
3. Conduct CP testing (tabletop or functional)
4. Document test results
5. Train personnel on CP procedures

**Effort:** Medium
**Cost:** $0-$5K (tabletop exercise)

---

### FED-GAP-010: System Use Notification (Login Banner)

| Attribute | Value |
|-----------|-------|
| **Priority** | 🟡 High |
| **Controls Affected** | AC-8 |
| **Current State** | No login banner |
| **Required State** | Government-approved login banner |

**Remediation:**
1. Implement login banner in Backstage UI
2. Implement login banner in Keycloak
3. Use FedRAMP-approved banner text
4. Test banner display

**Effort:** Low
**Cost:** $0

---

### FED-GAP-011: Rules of Behavior

| Attribute | Value |
|-----------|-------|
| **Priority** | 🟡 High |
| **Controls Affected** | PL-4, PS-6 |
| **Current State** | Acceptable use policy exists |
| **Required State** | FedRAMP-formatted Rules of Behavior |

**Remediation:**
1. Create Rules of Behavior document
2. Define user responsibilities
3. Implement acknowledgment tracking
4. Include in onboarding process

**Effort:** Low
**Cost:** $0

---

### FED-GAP-012: Incident Response Testing

| Attribute | Value |
|-----------|-------|
| **Priority** | 🟡 High |
| **Controls Affected** | IR-3 |
| **Current State** | IRP exists but not tested |
| **Required State** | Annual IR testing (tabletop/functional) |

**Remediation:**
1. Schedule tabletop exercise
2. Create test scenario
3. Conduct exercise
4. Document lessons learned
5. Update IRP based on findings

**Effort:** Low
**Cost:** $0-$5K

---

### FED-GAP-013: Training Records

| Attribute | Value |
|-----------|-------|
| **Priority** | 🟡 High |
| **Controls Affected** | AT-2, AT-3, AT-4 |
| **Current State** | No formal training records |
| **Required State** | Documented security awareness training |

**Remediation:**
1. Implement security awareness training
2. Track training completion
3. Require annual refresher
4. Document role-based training

**Effort:** Low
**Cost:** $0-$1K (free training available)

---

### FED-GAP-014: Separation of Duties

| Attribute | Value |
|-----------|-------|
| **Priority** | 🟡 High |
| **Controls Affected** | AC-5 |
| **Current State** | Solo founder - single person |
| **Required State** | Separation of duties or compensating controls |

**Remediation:**
1. Document compensating controls
2. Implement automated controls where possible
3. Use external review (auditor, consultant)
4. Document in SSP with rationale

**Effort:** Medium
**Cost:** $0

---

## Medium Priority Gaps

| ID | Gap | Controls | Effort |
|----|-----|----------|--------|
| FED-GAP-015 | Security Inbox | CA-7 | Low |
| FED-GAP-016 | Boundary Diagram | PL-2 | Medium |
| FED-GAP-017 | System Component Inventory | CM-8 | Medium |
| FED-GAP-018 | Interconnection Agreements | CA-3 | Medium |
| FED-GAP-019 | External System Policy | AC-20 | Low |
| FED-GAP-020 | Mobile Device Policy | AC-19 | Low |
| FED-GAP-021 | Media Protection Policy | MP-1 | Low |
| FED-GAP-022 | Cryptographic Key Mgmt Docs | SC-12 | Medium |
| FED-GAP-023 | Audit Log Retention Policy | AU-11 | Low |
| FED-GAP-024 | Audit Failure Alerting | AU-5 | Medium |
| FED-GAP-025 | Personnel Sanctions Process | PS-8 | Low |
| FED-GAP-026 | Access Agreements | PS-6 | Low |

---

## Reusable Assets from SOC 2 / ISO 27001

### High Reuse Potential (>80%)

| Asset | Source | FedRAMP Use |
|-------|--------|-------------|
| Information Security Policy | ISO 27001 | SC-1, SI-1, PL-1 |
| Access Control Policy | SOC 2 | AC-1 |
| Change Management Policy | SOC 2 | CM-1 |
| Incident Response Policy | ISO 27001 | IR-1 |
| Risk Management Policy | ISO 27001 | RA-1 |
| Vendor Management Policy | SOC 2 | SA-1, SR-1 |
| Risk Register | ISO 27001 | RA-3 |
| Risk Treatment Plan | ISO 27001 | RA-7 |
| Evidence Collection Scripts | SOC 2 | CA-7 |
| Corrective Actions Log | ISO 27001 | CA-5 |

### Medium Reuse Potential (50-80%)

| Asset | Source | FedRAMP Use | Modifications |
|-------|--------|-------------|---------------|
| Quarterly Access Review | SOC 2 | AC-2 | Add FedRAMP fields |
| Employee Onboarding | SOC 2 | PS-3, PS-4 | Add background check |
| Employee Offboarding | SOC 2 | PS-4 | Add access termination SLA |
| Vulnerability Scans | SOC 2 | RA-5 | FedRAMP scan format |
| SBOM | SOC 2 | SR-11 | Integrate into SCRMP |

### Low Reuse Potential (<50%)

| Asset | Source | FedRAMP Use | Gap |
|-------|--------|-------------|-----|
| Architecture Docs | Internal | PL-2 | Need boundary diagram |
| Policies | Various | Multiple | Need OSCAL format |
| Evidence | SOC 2 | CA-7 | Need ConMon format |

---

## Gap Remediation Roadmap

### Recommended Strategy: Prepare Now, Pay When Needed

Since you don't have federal customers yet, split work into **free prep** vs **paid execution**:

---

### Phase 0: Free Preparation (Now - Ongoing, Cost: $0)

Do this while waiting for federal customer interest:

| Task | Gap | Cost | Status |
|------|-----|------|--------|
| Keep SOC 2 + ISO 27001 current | Foundation | $0 | ✅ Ongoing |
| Run weekly/monthly evidence scripts | FED-GAP-005 | $0 | ✅ Automated |
| FIPS crypto inventory | FED-GAP-001 | $0 | ❌ Todo |
| Research BoringCrypto implementation | FED-GAP-001 | $0 | ❌ Todo |
| Draft SSP outline (sections 1-3) | FED-GAP-002 | $0 | ❌ Todo |
| Track FedRAMP 20x announcements | Strategy | $0 | ❌ Todo |
| Login banner implementation | FED-GAP-010 | $0 | ❌ Easy win |

**Why do these now:**
- Zero cost - just your time when convenient
- Removes biggest blockers (FIPS, SSP foundation)
- Evidence history builds credibility

---

### Phase 1: When Federal Customer Shows Interest

| Week | Task | Gap | Cost |
|------|------|-----|------|
| 1 | Confirm customer can sponsor | FED-GAP-003 | $0 |
| 1-2 | Implement FIPS crypto | FED-GAP-001 | $0 (dev time) |
| 2-4 | Complete SSP with customer input | FED-GAP-002 | $0 |
| 3-4 | Get 3PAO/assessor quotes | FED-GAP-004 | $0 |

### Phase 2: Authorization Sprint (Month 2-4)

| Week | Task | Gap | Cost |
|------|------|-----|------|
| 5-8 | Complete all SSP appendices | FED-GAP-002 | $0 |
| 6-8 | Digital Identity + ConMon | FED-GAP-005, 008 | $0 |
| 8-10 | Assessor engagement | FED-GAP-004 | $15K-$40K |
| 10-12 | Remediate findings | Various | $0 |

### Phase 3: Authorization (Month 4-6)

| Week | Task | Cost |
|------|------|------|
| 13-16 | Agency review | $0 |
| 16-20 | Final authorization | $0 |
| 20+ | FedRAMP Marketplace listing | $0 |

**Total Timeline:** 6-9 months from customer interest
**Total Cost:** $15K-$40K (assessor only)

---

## Gap Status Tracking

| ID | Gap | Status | Owner | Due |
|----|-----|--------|-------|-----|
| FED-GAP-001 | FIPS Crypto | ❌ Not Started | Carlos | TBD |
| FED-GAP-002 | SSP | ❌ Not Started | Carlos | TBD |
| FED-GAP-003 | Agency Sponsor | ❌ Not Started | Carlos | TBD |
| FED-GAP-004 | 3PAO | ❌ Not Started | Carlos | TBD |
| FED-GAP-005 | ConMon | ❌ Not Started | Carlos | TBD |
| FED-GAP-006 | SCRMP | ❌ Not Started | Carlos | TBD |
| FED-GAP-007 | Background Checks | ❌ Not Started | Carlos | TBD |
| FED-GAP-008 | Digital Identity | ❌ Not Started | Carlos | TBD |
| FED-GAP-009 | Contingency Plan | ❌ Not Started | Carlos | TBD |
| FED-GAP-010 | Login Banner | ❌ Not Started | Carlos | TBD |
| FED-GAP-011 | Rules of Behavior | ❌ Not Started | Carlos | TBD |
| FED-GAP-012 | IR Testing | ❌ Not Started | Carlos | TBD |
| FED-GAP-013 | Training Records | ❌ Not Started | Carlos | TBD |
| FED-GAP-014 | Separation of Duties | ❌ Not Started | Carlos | TBD |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial gap analysis |
