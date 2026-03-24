# CMMC Readiness Status

**Last Updated:** 2026-03-23
**Status:** FOUNDATION BUILDING
**Document Owner:** Carlos Sanchez, Founder/CEO

---

## Executive Summary

Aegis Technologies is building its Cybersecurity Maturity Model Certification (CMMC) readiness program to enable Department of Defense (DoD) customers and defense industrial base (DIB) contractors to use Aegis Platform for multi-cluster GPU workload orchestration. This document tracks CMMC readiness and builds on the existing SOC 2 Type II and ISO 27001 compliance programs.

| Metric | Status |
|--------|--------|
| **Target Level** | CMMC Level 2 (110 practices per NIST 800-171 Rev 2; 97 active requirements per Rev 3) |
| **Current Phase** | Foundation Building |
| **Foundation Compliance** | SOC 2 Type II + ISO 27001 |
| **Estimated Control Reuse** | ~40% from existing SOC 2 + ISO 27001 programs |
| **SPRS Score** | 38/110 preliminary (see 03-sprs-score-worksheet.md) |
| **C3PAO Selected** | No |
| **Target Certification** | Level 1 self-assessment by Q3 2026; Level 2 when DoD pipeline materializes |

---

## Critical Status Warning

**DO NOT claim "CMMC Certified" or "CMMC Level 2 Assessed"**

The following mandatory prerequisites have NOT been completed:

| Requirement | Status |
|-------------|--------|
| SPRS Score Calculated and Submitted | Not Completed |
| System Security Plan (SSP) Finalized | Not Completed |
| POA&M for Unmet Requirements | Not Completed |
| C3PAO Selected | Not Selected |
| C3PAO Assessment Conducted | Not Conducted |
| CMMC Level 1 Self-Assessment | Not Completed |
| CMMC Level 2 Certification Assessment | Not Conducted |
| SPRS Score Posted to SPRS Portal | Not Posted |

**Current State:** Building foundational CMMC program documentation and leveraging SOC 2 + ISO 27001 controls for maximum reuse.

---

## CMMC Level Overview

### Level 1: Foundational (17 Practices)
- **Scope:** Federal Contract Information (FCI) protection
- **Assessment:** Annual self-assessment
- **Requirements:** 17 practices from 6 NIST 800-171 domains
- **Cost:** Minimal (internal effort only)
- **Timeline:** 2-4 weeks once documentation is prepared

### Level 2: Advanced (110 Practices / 97 Active in Rev 3)
- **Scope:** Controlled Unclassified Information (CUI) protection
- **Assessment:** C3PAO assessment (third-party) for critical programs; self-assessment for non-critical
- **Requirements:** All NIST 800-171 Rev 2 requirements (aligned to Rev 3 catalog: 97 active controls)
- **Cost:** $50K-$150K for C3PAO assessment
- **Timeline:** 6-12 months from decision to pursue

### Level 3: Expert (110+ Practices + NIST 800-172)
- **Scope:** Advanced persistent threat (APT) protection
- **Assessment:** Government-led (DIBCAC)
- **Not currently targeted**

---

## Existing Assets and Foundation

### Compliance Programs

| Program | Status | CMMC Reuse Value |
|---------|--------|------------------|
| **SOC 2 Type II** | Active | ~30% direct control overlap (AC, AU, CM, IR families) |
| **ISO 27001** | Pending internal audit | ~35% control overlap (risk management, ISMS structure) |
| **FedRAMP Low evaluation** | Assessment phase | ~50% overlap with CMMC Level 2 (both map to 800-53) |

### Documentation Assets

| Asset | Location | CMMC Use |
|-------|----------|----------|
| Control Implementation Statements (40 controls) | `customer-docs/control-implementation-statements.md` | Direct input to SSP and CMMC practice descriptions |
| OSCAL NIST 800-171 Rev 3 Catalog | `docs/oscal/nist_sp_800_171_rev3_catalog.json` | Machine-readable control catalog for automated mapping |
| FedRAMP Control Mapping (145 controls) | `docs/compliance/fedramp/01-control-mapping.csv` | Cross-reference for 800-53 to 800-171 mapping |
| FedRAMP Gap Analysis | `docs/compliance/fedramp/02-gap-analysis.md` | Shared gaps (FIPS crypto, SSP, supply chain) |
| Risk Register (17 risks) | ISO 27001 program | Input to RA family controls |
| Quarterly Access Reviews | SOC 2 program | Evidence for AC family controls |
| Evidence Collection Scripts | `scripts/compliance/` | Automated evidence generation for continuous monitoring |
| Corrective Actions Log | ISO 27001 program | POA&M format and process |

### Agent Tooling for Evidence

| Tool | Purpose | CMMC Relevance |
|------|---------|----------------|
| `mcp__aegis__evidence_write_json` | Write structured compliance evidence | Automated evidence collection for CMMC practices |
| `mcp__aegis__run_tests` | Run automated tests | Continuous validation of technical controls |
| `scripts/compliance/run_weekly_checks.sh` | Weekly compliance checks | Regular evidence generation |
| `scripts/compliance/setup_evidence_vault.sh` | Evidence vault setup | Organized evidence repository |

---

## SPRS Score

**Current SPRS Score:** 38/110 (preliminary, per 03-sprs-score-worksheet.md)

The SPRS (Supplier Performance Risk System) score represents the organization's self-assessed implementation status against NIST 800-171 requirements. The score will be calculated in the `03-sprs-score-worksheet.md` document.

| Metric | Value |
|--------|-------|
| **Maximum Possible Score** | 110 |
| **Minimum Possible Score** | -203 |
| **Passing Threshold** | No formal minimum, but scores below 0 indicate significant gaps |
| **Estimated Score Range** | 40-60 (based on preliminary control mapping) |
| **Score Submission** | Required to be posted in SPRS portal before contract award |

**Note:** The SPRS score must be submitted to the DoD SPRS portal and is valid for 3 years. A Plan of Action and Milestones (POA&M) must accompany any score below 110.

---

## C3PAO Assessment

**C3PAO Status:** Not selected

### Selection Criteria (When Ready)

| Criterion | Requirement |
|-----------|-------------|
| CMMC Accreditation Body (Cyber AB) authorized | Mandatory |
| Experience with cloud-native / Kubernetes platforms | Preferred |
| Experience with SaaS/PaaS assessments | Preferred |
| DoD supply chain assessment experience | Preferred |
| Geographic proximity or remote capability | Preferred |
| Cost within budget | Mandatory |

### Estimated Assessment Costs

| Assessment Type | Cost Range | Timeline | Notes |
|-----------------|-----------|----------|-------|
| **Level 1 Self-Assessment** | $0 (internal) | 2-4 weeks | Annual self-assessment, results posted to SPRS |
| **Level 2 Self-Assessment** | $5K-$15K (internal + consultant review) | 4-8 weeks | For non-critical CUI programs |
| **Level 2 C3PAO Assessment** | $50K-$150K | 3-6 months | Required for critical CUI programs |
| **Level 2 C3PAO Readiness Review** | $15K-$30K | 2-4 weeks | Optional pre-assessment gap check |

---

## Recommended Strategy

### Phase 1: Level 1 Self-Assessment (Recommended Now)

Pursue CMMC Level 1 self-assessment immediately. This establishes CMMC compliance baseline with minimal investment and qualifies Aegis for FCI-only DoD contracts.

**Rationale:**
- Level 1 requires only 17 basic practices (all likely already met)
- Self-assessment only -- no C3PAO cost
- Demonstrates compliance commitment to DoD prospects
- Creates foundation for Level 2 pursuit
- Can be completed in 2-4 weeks

### Phase 2: Level 2 Preparation (When DoD Pipeline Materializes)

Begin Level 2 preparation when a concrete DoD customer opportunity requires CUI handling.

**Rationale:**
- Level 2 C3PAO assessment costs $50K-$150K
- ~40% of controls reusable from SOC 2 + ISO 27001
- Shared work with FedRAMP program (FIPS crypto, SSP, supply chain plan)
- No business justification for upfront investment without DoD revenue pipeline
- 6-12 month preparation timeline is manageable once triggered by customer demand

---

## Timeline

### Phase 0: Foundation Building (Current -- Q2 2026)

| Task | Owner | Target | Status |
|------|-------|--------|--------|
| Complete CMMC control mapping (01-control-mapping.csv) | Carlos | 2026-03-09 | Complete |
| Complete gap analysis (02-gap-analysis.md) | Carlos | 2026-03-09 | Complete |
| Complete SPRS score worksheet (03-sprs-score-worksheet.md) | Carlos | 2026-03-09 | Complete |
| Cross-reference with FedRAMP program | Carlos | 2026-03-15 | In Progress |
| Identify shared gaps with FedRAMP (FIPS crypto, SSP, SCRMP) | Carlos | 2026-03-31 | Not Started |

### Phase 1: Level 1 Self-Assessment (Q2-Q3 2026)

| Task | Owner | Target | Status |
|------|-------|--------|--------|
| Verify all 17 Level 1 practices are met | Carlos | 2026-04-15 | Not Started |
| Document Level 1 practice implementations | Carlos | 2026-04-30 | Not Started |
| Complete Level 1 self-assessment | Carlos | 2026-05-15 | Not Started |
| Submit SPRS score (Level 1) to SPRS portal | Carlos | 2026-05-31 | Not Started |
| Annual Level 1 re-assessment process | Carlos | Ongoing | Not Started |

### Phase 2: Level 2 Readiness (Trigger: DoD Customer Pipeline)

| Task | Owner | Target | Status |
|------|-------|--------|--------|
| Finalize SSP with CMMC-specific sections | Carlos | T+4 weeks | Not Started |
| Implement FIPS 140-2/3 cryptographic modules | Carlos | T+8 weeks | Not Started |
| Complete supply chain risk management plan | Carlos | T+6 weeks | Not Started |
| Address all POA&M items from gap analysis | Carlos | T+12 weeks | Not Started |
| Conduct internal readiness review | Carlos | T+14 weeks | Not Started |
| Select and engage C3PAO | Carlos | T+16 weeks | Not Started |
| C3PAO readiness assessment (optional) | C3PAO | T+20 weeks | Not Started |
| Remediate C3PAO findings | Carlos | T+24 weeks | Not Started |
| C3PAO certification assessment | C3PAO | T+28 weeks | Not Started |
| Submit SPRS score (Level 2) to SPRS portal | Carlos | T+30 weeks | Not Started |

*T = Trigger date (DoD customer opportunity confirmed)*

### Phase 3: Continuous Compliance (Post-Certification)

| Task | Frequency | Owner |
|------|-----------|-------|
| SPRS score review and update | Annual | Carlos |
| POA&M progress review | Monthly | Carlos |
| Evidence collection | Continuous (automated) | Scripts |
| Control effectiveness validation | Quarterly | Carlos |
| Level 2 re-assessment | Triennial (C3PAO) | C3PAO |

---

## Active Control Evidence

### SI Family — Flaw Remediation (SI.L2-3.14.1)

**2026-03-23:** Weekly automated scan identified 55 Dependabot alerts. All 35 critical and high severity vulnerabilities remediated same-day across 3 repositories.

| Practice | Description | Evidence |
|----------|-------------|----------|
| **SI.L2-3.14.1** | Identify, report, and correct system flaws in a timely manner | 35 critical+high vulns identified and patched within 24 hours |
| **SI.L2-3.14.2** | Provide protection from malicious code at designated locations | Dependabot + automated scanning active on all repos |
| **SI.L2-3.14.3** | Monitor system security alerts/advisories | Weekly automated checks via `run_weekly_checks.sh` |
| **RA.L2-3.11.2** | Remediate vulnerabilities in accordance with risk assessments | Critical/high remediated same-day; medium/low scheduled for next cycle |

**Remediation details:**
- aegis-platform: gRPC authorization bypass (critical), OTel PATH hijacking (high) — Go module upgrades
- aegis-ui: fast-xml-parser regex injection (critical), 10 high-severity npm deps — yarn resolutions
- sovran: 20 high-severity npm + Python deps — npm overrides + pip constraint upgrades

**Evidence location:** `$EVIDENCE_VAULT/soc2/2026/2026-03/vuln-management/2026-03-23_vulnerability_remediation_report.md`

**SPRS impact:** This remediation activity demonstrates active implementation of SI and RA family controls, supporting the current SPRS score and providing evidence for future score improvements.

---

## Cross-Framework Synergies

### Shared Work with FedRAMP

The following gaps are shared between CMMC Level 2 and FedRAMP Low programs. Addressing them once benefits both:

| Gap | CMMC Impact | FedRAMP Impact | Effort |
|-----|-------------|----------------|--------|
| FIPS 140-2/3 Cryptographic Modules | SC family (03.13.10, 03.13.11) | IA-7, SC-13 | High |
| System Security Plan | PL family (03.15.01, 03.15.02) | PL-2 | High |
| Supply Chain Risk Management Plan | SR family (03.17.01-03) | SR-1, SR-2 | Medium |
| Contingency Plan Testing | Not directly in 800-171 | CP-4 | Medium |
| Personnel Background Checks | PS family (03.09.01) | PS-3 | Medium |
| Login Banner | AC family (03.01.09) | AC-8 | Low |
| Training Records | AT family (03.02.01, 03.02.02) | AT-4 | Low |

### SOC 2 Control Reuse

| SOC 2 TSC | CMMC Domains Covered | Reuse Level |
|-----------|----------------------|-------------|
| CC6 (Logical and Physical Access) | AC, IA | High |
| CC7 (System Operations) | AU, SI | High |
| CC8 (Change Management) | CM, SA | High |
| CC4 (Monitoring) | CA, RA | Medium |
| CC1 (Control Environment) | AT, PL | Medium |
| CC9 (Risk Mitigation) | SR | Medium |

### ISO 27001 Control Reuse

| ISO 27001 Clause/Annex | CMMC Domains Covered | Reuse Level |
|-------------------------|----------------------|-------------|
| A.5 (Organizational) | PL, AT, PS | High |
| A.6 (People) | PS, AT | High |
| A.7 (Physical) | PE, MA | N/A (cloud-inherited) |
| A.8 (Technological) | AC, AU, CM, IA, SC, SI | High |
| Clause 6 (Planning) | RA, PL | High |
| Clause 8 (Operation) | CM, IR | High |
| Clause 9 (Evaluation) | CA | Medium |

---

## Related Documents

| Document | Location |
|----------|----------|
| CMMC Control Mapping | `./01-control-mapping.csv` |
| CMMC Gap Analysis | `./02-gap-analysis.md` |
| SPRS Score Worksheet | `./03-sprs-score-worksheet.md` |
| FedRAMP Status | `../fedramp/STATUS.md` |
| SOC 2 Status | `../soc2/STATUS.md` |
| FedRAMP Control Mapping | `../fedramp/01-control-mapping.csv` |
| FedRAMP Gap Analysis | `../fedramp/02-gap-analysis.md` |
| Control Implementation Statements | `../customer-docs/control-implementation-statements.md` |
| OSCAL NIST 800-171 Rev 3 Catalog | `../../oscal/nist_sp_800_171_rev3_catalog.json` |

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.1 | 2026-03-23 | Claude Opus 4.6 | Added active control evidence section: SI family flaw remediation (35 critical+high vulns fixed same-day) |
| 1.0 | 2026-03-09 | Carlos Sanchez | Initial CMMC readiness status document |
