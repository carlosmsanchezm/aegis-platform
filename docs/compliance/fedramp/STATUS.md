# FedRAMP Readiness Status

**Last Updated:** 2026-03-23
**Status:** 🔴 PRE-ASSESSMENT - RESEARCH & PLANNING
**Document Owner:** Carlos Sanchez, Founder/CEO

---

## Executive Summary

Aegis Technologies is evaluating FedRAMP authorization to enable federal government customers to use Aegis Platform for GPU workload orchestration. This document tracks FedRAMP readiness status and builds on existing SOC 2 Type II and ISO 27001 compliance programs.

| Metric | Status |
|--------|--------|
| **Target Baseline** | LI-SaaS or FedRAMP 20x Low |
| **Current Phase** | Pre-Assessment (Research & Planning) |
| **Foundation Compliance** | SOC 2 ✅ + ISO 27001 ✅ |
| **Estimated Reuse** | ~25-30% of controls from existing programs |
| **Target Authorization** | Q4 2026 or Q1 2027 |

---

## ⚠️ Critical Status Warning

**DO NOT claim "FedRAMP Ready" or "FedRAMP In Process"**

The following mandatory steps have NOT been completed:

| Requirement | Status |
|-------------|--------|
| Impact Level Determination (FIPS 199) | ❌ **NOT COMPLETED** |
| Agency Sponsor Identified | ❌ **NOT IDENTIFIED** |
| 3PAO Selected | ❌ **NOT SELECTED** |
| System Security Plan (SSP) | ❌ **NOT WRITTEN** |
| Readiness Assessment | ❌ **NOT CONDUCTED** |
| FedRAMP Connect Submission | ❌ **NOT SUBMITTED** |

**Current State:** Evaluating FedRAMP authorization paths and leveraging existing SOC 2 + ISO 27001 controls.

---

## Authorization Path Options

### Recommended: FedRAMP 20x Low or LI-SaaS

Based on Aegis's current profile (Model B - self-hosted software, minimal PII), the most efficient paths are:

| Path | Controls | Timeline | Cost | Best For |
|------|----------|----------|------|----------|
| **FedRAMP 20x Low** | ~51 KSIs | 3-6 months | $50K-100K | Cloud-native services |
| **LI-SaaS** | 45-65 | 6-9 months | $75K-150K | Simple SaaS, minimal PII |
| **FedRAMP Low** | 156 | 9-12 months | $100K-250K | Traditional path |
| **FedRAMP Moderate** | ~325 | 12-18 months | $250K-500K | CUI, sensitive data |

**Recommended Path:** FedRAMP 20x Low (when available) or LI-SaaS

**Rationale:**
- Aegis Platform handles only minimal PII (usernames, emails for auth)
- Cloud-native architecture fits 20x automation requirements
- Existing SOC 2 + ISO 27001 provides strong foundation
- Faster time-to-market for federal customers

---

## Aegis Platform: Federal Use Case

### Target Federal Agencies

| Agency Type | Use Case | Data Classification |
|-------------|----------|---------------------|
| DoD/IC Research | GPU workload orchestration | CUI possible (→ Moderate) |
| Civilian Agencies | ML/AI infrastructure | Low Impact |
| National Labs | HPC/GPU management | Low Impact |

### Platform Deployment Model

| Aspect | Current (Model B) | Federal Deployment |
|--------|-------------------|-------------------|
| Deployment | Customer self-hosted | Customer self-hosted in GovCloud |
| Data Location | Customer environment | Customer's FedRAMP boundary |
| Aegis Responsibility | Software development | Software + optional support |
| Customer Responsibility | Operations, security monitoring | All operational controls |

**Key Decision:** Is Aegis seeking authorization for:
1. **Software only** (customer operates in their boundary) → Simpler
2. **Managed service** (Aegis operates for customer) → More complex

---

## Foundation: Existing Compliance Programs

### SOC 2 Type II (✅ Complete)

| Aspect | Status | FedRAMP Reuse |
|--------|--------|---------------|
| Security TSC | ✅ Implemented | ~30% control overlap |
| Policies (6) | ✅ Approved | ✅ Reuse with tailoring |
| Access Reviews | ✅ Quarterly | ✅ Meets FedRAMP AC family |
| Vulnerability Mgmt | ✅ Continuous | ✅ Meets RA/SI families |
| Change Management | ✅ CI/CD | ✅ Meets CM/SA families |
| Evidence Vault | ✅ Operational | ✅ Adapt format for OSCAL |

### ISO 27001 (⏳ Pending Internal Audit)

| Aspect | Status | FedRAMP Reuse |
|--------|--------|---------------|
| ISMS Documentation | ✅ Complete | ✅ Maps to SSP sections |
| Risk Register (17 risks) | ✅ Complete | ✅ Input to RA-3 |
| Statement of Applicability | ✅ 93 controls | ✅ Cross-reference to 800-53 |
| Corrective Actions | ✅ CA-001 closed | ✅ Demonstrates POA&M process |
| Internal Audit | ❌ Pending | Required for ISO, helpful for FedRAMP |

### Control Reuse Estimate

| Control Family | SOC 2 Coverage | ISO 27001 Coverage | FedRAMP Gap |
|----------------|----------------|--------------------| ------------|
| Access Control (AC) | 70% | 80% | 20-30% |
| Audit (AU) | 60% | 70% | 30-40% |
| Config Mgmt (CM) | 80% | 75% | 20-25% |
| Incident Response (IR) | 70% | 80% | 20-30% |
| Risk Assessment (RA) | 50% | 90% | 10-20% |
| System & Info (SI) | 60% | 70% | 30-40% |
| **Total Reuse** | | | **~25-30%** |

---

## FedRAMP-Specific Requirements

### What's New (Not Covered by SOC 2/ISO 27001)

| Requirement | Description | Effort |
|-------------|-------------|--------|
| **FIPS 140-2/3** | Cryptographic module validation | High |
| **OSCAL Format** | Machine-readable SSP | Medium |
| **FedRAMP Templates** | Specific document formats | Medium |
| **Continuous Monitoring** | ConMon reporting (monthly) | Medium |
| **3PAO Assessment** | Independent security assessment | High ($$$) |
| **Agency Sponsorship** | Federal agency to sponsor ATO | High |
| **POA&M Management** | Formal remediation tracking | Low (have CA log) |
| **Digital Identity (SP 800-63)** | IAL/AAL/FAL determination | Medium |

### FIPS 140-2/3 Cryptography

**Current State Analysis:**

| Component | Cryptography Used | FIPS Status |
|-----------|-------------------|-------------|
| Platform API | TLS 1.2/1.3 (Go stdlib) | ⚠️ Not validated |
| mTLS | X.509 certificates | ⚠️ Not validated |
| Keycloak | TLS, token signing | ⚠️ Check configuration |
| AWS Services | AWS-provided encryption | ✅ AWS is FIPS validated |
| Container Images | Alpine/Debian crypto | ⚠️ Not validated |

**Required Actions:**
1. Use FIPS-validated crypto libraries (e.g., BoringCrypto for Go)
2. Configure Keycloak for FIPS mode
3. Use FIPS-validated base images
4. Document cryptographic inventory

---

## Gap Analysis Summary

### High Priority Gaps

| ID | Gap | Current State | Required | Effort |
|----|-----|---------------|----------|--------|
| FED-001 | No FIPS 140 crypto | Standard crypto | FIPS-validated modules | High |
| FED-002 | No SSP document | Policies only | Full SSP + appendices | High |
| FED-003 | No agency sponsor | N/A | Federal agency relationship | High |
| FED-004 | No 3PAO | N/A | Accredited assessor | High ($$$) |
| FED-005 | No OSCAL package | Markdown/CSV | Machine-readable | Medium |
| FED-006 | No digital identity assessment | Basic auth | SP 800-63 DIW | Medium |

### Medium Priority Gaps

| ID | Gap | Current State | Required | Effort |
|----|-----|---------------|----------|--------|
| FED-007 | No supply chain plan | Vendor policy | SCRMP per SP 800-161 | Medium |
| FED-008 | No security inbox | General email | Dedicated inbox + SLA | Low |
| FED-009 | Limited personnel security | Onboard/offboard | Background checks | Medium |
| FED-010 | No boundary diagram | Architecture docs | FedRAMP-specific diagram | Medium |

### Inheritable Controls (from AWS GovCloud)

If Aegis or customers deploy on AWS GovCloud, many controls are inherited:

| Control Family | Inheritable | Examples |
|----------------|-------------|----------|
| Physical (PE) | 100% | Data center security |
| Media (MP) | 90% | Media destruction |
| Environmental (PE) | 100% | Fire, HVAC, power |
| Maintenance (MA) | 80% | Hardware maintenance |

---

## Active Control Evidence (Pre-Assessment)

Even though FedRAMP assessment has not begun, existing compliance operations provide evidence for several FedRAMP control families. This evidence will be referenced in the SSP when written.

### SI-2: Flaw Remediation — Active

**2026-03-23:** Automated vulnerability scanning identified 55 Dependabot alerts. All 35 critical and high severity vulnerabilities were remediated same-day.

| Control | Requirement | Evidence |
|---------|-------------|----------|
| **SI-2** | Identify, report, and correct information system flaws | 35 critical+high vulns patched within 24 hours (commits: `a2708aa`, `96b163d`, `de15aa1`) |
| **SI-2(2)** | Employ automated mechanisms to determine flaw remediation status | Dependabot + weekly `run_weekly_checks.sh` automation |
| **RA-5** | Scan for vulnerabilities and remediate | Weekly automated scanning, same-day critical remediation demonstrated |
| **RA-5(2)** | Update vulnerabilities to be scanned | Dependabot auto-updates vulnerability database |
| **CM-3** | Configuration change control | All fixes via version-controlled commits with CI build verification |
| **SA-11** | Developer security testing | Build verification (`go build`, `yarn tsc`, `yarn build:backend`, `npm run build`, `uv sync`) after all patches |

**SSP sections this will support:** SI-2, RA-5, CM-3, SA-11
**Evidence location:** `$EVIDENCE_VAULT/soc2/2026/2026-03/vuln-management/2026-03-23_vulnerability_remediation_report.md`

---

## Roadmap to FedRAMP Ready

### Phase 1: Assessment (Current - Q1 2026)

| Task | Owner | Due | Status |
|------|-------|-----|--------|
| Complete FedRAMP research | Carlos | 2026-01-17 | ✅ Done |
| Determine target baseline (Low vs Moderate) | Carlos | 2026-01-31 | 🔄 In Progress |
| Identify potential agency sponsors | Carlos | 2026-02-15 | ❌ Not Started |
| Budget for 3PAO assessment | Carlos | 2026-02-28 | ❌ Not Started |
| FIPS crypto inventory | Carlos | 2026-02-28 | ❌ Not Started |

### Phase 2: Foundation (Q2 2026)

| Task | Owner | Due | Status |
|------|-------|-----|--------|
| Begin SSP outline | Carlos | 2026-04-01 | ❌ Not Started |
| Complete digital identity worksheet | Carlos | 2026-04-15 | ❌ Not Started |
| Implement FIPS crypto | Carlos | 2026-05-01 | ❌ Not Started |
| Create boundary diagram | Carlos | 2026-05-15 | ❌ Not Started |
| Draft supply chain plan | Carlos | 2026-05-31 | ❌ Not Started |

### Phase 3: Documentation (Q3 2026)

| Task | Owner | Due | Status |
|------|-------|-----|--------|
| Complete SSP Appendix A (controls) | Carlos | 2026-07-01 | ❌ Not Started |
| Complete all SSP appendices | Carlos | 2026-08-01 | ❌ Not Started |
| Convert to OSCAL format | Carlos | 2026-08-15 | ❌ Not Started |
| Internal review | Carlos | 2026-09-01 | ❌ Not Started |

### Phase 4: Assessment (Q4 2026)

| Task | Owner | Due | Status |
|------|-------|-----|--------|
| Select 3PAO | Carlos | 2026-09-15 | ❌ Not Started |
| Readiness assessment | 3PAO | 2026-10-15 | ❌ Not Started |
| Remediate findings | Carlos | 2026-11-15 | ❌ Not Started |
| Full assessment | 3PAO | 2026-12-15 | ❌ Not Started |

### Phase 5: Authorization (Q1 2027)

| Task | Owner | Due | Status |
|------|-------|-----|--------|
| Submit package to agency/PMO | Carlos | 2027-01-15 | ❌ Not Started |
| Address PMO questions | Carlos | 2027-02-01 | ❌ Not Started |
| Receive ATO | Agency | 2027-03-01 | ❌ Not Started |
| List on FedRAMP Marketplace | PMO | 2027-03-15 | ❌ Not Started |

---

## Cost Estimates

### Recommended: Wait for 20x + Agency Sponsor

| Phase | Cost | Notes |
|-------|------|-------|
| **Preparation (Now)** | **$0** | Keep SOC 2 + ISO 27001 current |
| **FIPS Crypto Implementation** | **$0** | BoringCrypto is free (dev time only) |
| **When Customer Appears** | **$20K-$40K** | Agency authorization path |

### Path Comparison

| Path | Cost | Timeline | When to Use |
|------|------|----------|-------------|
| **Agency Authorization** | $20K-$40K | 6-9 months | Have federal customer sponsor |
| **FedRAMP 20x Low** | $15K-$40K | 3-6 months | 20x is public (mid-2026), no sponsor |
| **LI-SaaS** | $40K-$60K | 6-9 months | Need auth now, no sponsor |
| **Traditional Low** | $100K-$200K | 12-18 months | ❌ Avoid - too expensive |

### What's Actually Required

| Item | Cost | Can Avoid? |
|------|------|------------|
| Some form of assessment | $15K-$40K | No - required |
| FIPS crypto implementation | $0 (dev time) | No - federal mandate |
| SSP documentation | $0 (your time) | No - required |
| GRC tools | $0-$5K | Yes - can use spreadsheets |
| Consulting | $0-$20K | Yes - optional |

### Ongoing (Annual)

| Item | Estimate | Notes |
|------|----------|-------|
| Continuous Monitoring | $0-$5,000 | Your existing scripts + free tools |
| Annual Assessment | $10,000-$20,000 | Smaller scope after initial |
| **Total Annual** | **$10K-$25K** | |

---

## Key Decisions Required

### Decision 1: Target Baseline

| Option | Pros | Cons | Recommendation |
|--------|------|------|----------------|
| **FedRAMP 20x Low** | Fastest, cheapest, automation-friendly | New program, limited history | ✅ Primary |
| **LI-SaaS** | Established path, streamlined | More controls than 20x | Fallback |
| **FedRAMP Low** | Well-established | More documentation | If 20x unavailable |
| **FedRAMP Moderate** | Required for CUI | Expensive, complex | Only if customer requires |

**Recommendation:** Start with FedRAMP 20x Low (when available). Fall back to LI-SaaS if 20x timeline doesn't align.

### Decision 2: Authorization Path

| Option | Pros | Cons | Recommendation |
|--------|------|------|----------------|
| **Agency Authorization** | Faster, agency relationship | Tied to single agency | ✅ Preferred |
| **Program Authorization** | Not tied to agency | Competitive, limited slots | Alternative |

**Recommendation:** Pursue agency authorization with a target customer agency.

### Decision 3: Deployment Model

| Option | Pros | Cons | Recommendation |
|--------|------|------|----------------|
| **Software-only (Model B)** | Simpler, fewer controls | Limited scope | Current model |
| **Managed Service (Model C)** | More value, recurring revenue | More controls, operational burden | Future |

**Recommendation:** Start with Model B authorization, expand to Model C later.

---

## Document Inventory (Planned)

### System Security Plan (SSP)

| Document | Status | Priority |
|----------|--------|----------|
| SSP Main Document | ❌ Not Started | High |
| Appendix A: Control Implementation | ❌ Not Started | High |
| Appendix B: Related Acronyms | ❌ Not Started | Low |
| Appendix C: Security Policies | ✅ Reuse from SOC 2/ISO | Medium |
| Appendix D: User Guide | ❌ Not Started | Medium |
| Appendix E: Digital Identity Worksheet | ❌ Not Started | High |
| Appendix F: Rules of Behavior | ❌ Not Started | Medium |
| Appendix G: Info System Contingency Plan | ❌ Not Started | Medium |
| Appendix H: Configuration Management Plan | ❌ Not Started | High |
| Appendix I: Incident Response Plan | ✅ Reuse from ISO 27001 | Medium |
| Appendix J: CIS Workbook | ❌ Not Started | High |
| Appendix K: FIPS 140 Validation | ❌ Not Started | High |
| Appendix L: Separation of Duties Matrix | ❌ Not Started | Medium |
| Appendix M: Integrated Inventory | ❌ Not Started | High |
| Appendix N: Continuous Monitoring Plan | ❌ Not Started | High |
| Appendix O: POA&M | ✅ Reuse CA log format | Medium |
| Appendix P: Supply Chain Risk Mgmt Plan | ❌ Not Started | High |
| Appendix Q: Cryptographic Modules | ❌ Not Started | High |

---

## Related Documents

| Document | Location |
|----------|----------|
| SOC 2 Status | `../soc2/STATUS.md` |
| ISO 27001 Status | `../iso27001/STATUS.md` |
| Compliance Program Status | `../COMPLIANCE_PROGRAM_STATUS.md` |
| FedRAMP Gap Analysis | `./01-gap-analysis.md` |
| FedRAMP Control Mapping | `./02-control-mapping.csv` |
| FedRAMP Readiness Roadmap | `./03-readiness-roadmap.md` |
| SSP Outline | `./04-ssp-outline.md` |

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.1 | 2026-03-23 | Claude Opus 4.6 | Added active control evidence: SI-2/RA-5 flaw remediation (35 critical+high vulns fixed same-day). Updated SSP sections this supports. |
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial FedRAMP assessment |
