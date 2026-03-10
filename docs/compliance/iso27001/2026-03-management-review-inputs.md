# Management Review Inputs -- March 2026

**Document ID:** ISMS-MR-2026-03
**ISO 27001 Reference:** Clause 9.3
**Review Frequency:** Semi-annual (minimum annual per ISO 27001)

---

## Management Review Meeting

**Suggested Date:** 2026-03-15
**Attendees:** Carlos Sanchez (Top Management / Information Security Manager)
**Location:** Remote / Virtual
**Duration:** ~1 hour
**Internal Audit Date:** 2026-03-10 to 2026-03-11

> **Note:** This is the FIRST management review since ISMS establishment on 2026-01-17. All data
> covers the period January 17 -- March 10, 2026 (~8 weeks of ISMS operation). Because this is the
> initial review, there is no prior review baseline for trend comparison.

---

## 1. Required Inputs (Clause 9.3.2)

### 1.1 Status of Actions from Previous Reviews

**First review -- N/A.** This is the initial management review. No prior actions exist.

| Action Item | Source | Due Date | Status | Notes |
|-------------|--------|----------|--------|-------|
| N/A | First review | -- | -- | No previous management review conducted |

---

### 1.2 Changes in External and Internal Issues

#### External Changes

| Issue | Change | ISMS Impact |
|-------|--------|-------------|
| Regulatory | FedRAMP free-tier prep initiated (SSP sections 1-3, appendices E/F/L completed); CMMC program structure added | Expanded compliance scope in planning; no immediate ISMS change required but influences risk appetite |
| Customer requirements | Pre-revenue startup; no external customers yet | No customer-driven security requirements at this time; ISMS is proactive |
| Threat landscape | AI-assisted attack tooling maturing; supply chain attacks increasing (OpenTelemetry Go SDK vuln CVE active) | RISK-001 (supply chain) and RISK-008 (phishing) remain relevant; Dependabot alerts require ongoing vigilance |
| Technology | Platform buildout active: EKS spoke provisioning, gRPC proxy, internal PKI (step-ca), Keycloak OIDC | Expanded attack surface from new infrastructure components; controls in CI/CD pipeline mitigate |
| Market conditions | Government/defense market interest driving FedRAMP and CMMC preparation | Positive alignment -- compliance investment has dual-use value (ISO 27001 + GovCloud readiness) |

#### Internal Changes

| Issue | Change | ISMS Impact |
|-------|--------|-------------|
| Organizational structure | Solo founder; no change from ISMS establishment | All roles remain held by Carlos Sanchez; compensating controls documented |
| Personnel | No new hires | Onboarding/offboarding procedures (RISK-005) remain untested but documented |
| Products/services | Aegis Platform v1.0 pilot release in progress; hub-and-spoke architecture operational | Production workloads introduce real operational risk; need to ensure controls cover production |
| ISMS establishment | ISMS established 2026-01-17; evidence vault operational since 2026-01-17 (CA-001 closed) | ISMS is now operating; evidence collection running monthly |
| Technology stack | CI enforcement active on all 3 repos; branch protection with required status checks | Positive: automated security controls reduce manual oversight burden |
| Evidence collection | Evidence vault: 472 total files (27 from 2025 retroactive, 140 non-placeholder from 2026); monthly collection running Jan-Mar 2026 | Compliance evidence generation is operational and consistent |

---

### 1.3 Feedback on Information Security Performance

#### 1.3.1 Nonconformities and Corrective Actions

| CA ID | Description | Root Cause | Status | Effectiveness |
|-------|-------------|------------|--------|---------------|
| CA-001 | Evidence vault not operational -- collection scripts existed but vault not created/populated | New ISMS implementation; scripts created but not executed; vault directory not initialized | **Closed** (2026-01-17) | Effective -- vault now contains 472 files across 6 months; monthly collection running |
| CA-002 | aegis-ui unit tests disabled in CI due to @backstage/cli Node 20 incompatibility | @backstage/cli jest config uses `promisify(glob)` incompatible with glob versions that removed callback API | **Open** (due 2026-04-15) | Pending -- lint, typecheck, and build still enforced in CI; test gap accepted as low risk |

**Summary:**
- Total CAs identified: 2
- CAs closed: 1 (CA-001)
- CAs open: 1 (CA-002 -- due 2026-04-15, extended from 2026-03-01 because upstream Backstage CLI fix not yet released)

**Internal audit findings** (2026-03-10/11): To be incorporated after audit completion. The audit is scheduled to cover Clauses 4-10 and Annex A controls, conducted by an external auditor for independence.

---

#### 1.3.2 Monitoring and Measurement Results

| KPI/Metric | Target | Actual | Status | Notes |
|------------|--------|--------|--------|-------|
| Open critical/high vulnerabilities (aegis-platform) | <5 | **1 high** (OTel Go SDK PATH hijacking) | Met | 11 alerts fixed historically; 3 total open (1 high, 1 medium, 1 low) |
| Open critical/high vulnerabilities (aegis-ui) | <5 | **9** (2 critical, 7 high) | **Not Met** | fast-xml-parser entity injection (critical); mkdocs, minimatch, Koa (high); upstream Backstage dependencies |
| Open critical/high vulnerabilities (sovran) | <5 | **26 high** | **Not Met** | VS Code extension; all high-severity; primarily transitive dependencies |
| Open vulnerabilities (all repos total) | <15 | **46** (2 critical, 34 high, 5 medium, 5 low) | **Not Met** | Majority in aegis-ui (13) and sovran (30); aegis-platform well-managed (3) |
| Security incidents (P1/P2) | <2/quarter | **0** | Met | Zero security incidents Jan-Mar 2026 per monthly incident logs |
| Policy acknowledgment rate | 100% | **100%** (1/1 personnel) | Met | Solo founder; all 7 policies approved and signed 2026-01-17 |
| MFA coverage (GitHub) | 100% | **100%** | Met | MFA enabled on GitHub account (carlosmsanchezm) |
| MFA coverage (AWS) | 100% | **Needs verification** | Review | AWS account shows 0 IAM users (root-only); root MFA status needs confirmation (evidence shows `account_mfa_enabled: 0` -- potential gap) |
| Access review completion | 100% | **Quarterly structure created** | Partial | Q4 2025, Q1-Q4 2025 directories exist with .gitkeep; no formal access review documents completed yet |
| Training completion | 100% | **100%** (1/1 personnel) | Met | Solo founder; security awareness, secure coding, incident response, ISO 27001 awareness all marked complete |
| Overdue risk treatments | 0 | **10** | **Not Met** | Multiple treatments past due date (see Section 1.5 for details) |
| Evidence collection months | 3/3 | **3/3** | Met | Jan (50 files), Feb (20 files), Mar (69 files) evidence collected |

**Key concerns requiring management decision:**
1. Vulnerability counts in aegis-ui (13 open) and sovran (30 open) exceed targets -- most are transitive dependencies
2. AWS root MFA evidence shows `account_mfa_enabled: 0` -- needs immediate investigation
3. Formal quarterly access reviews not yet executed (directory structure exists but no review documents)
4. 10 risk treatments overdue (see Section 1.5)

---

#### 1.3.3 Audit Results

**Internal Audits:**

| Audit ID | Date | Scope | Major NCs | Minor NCs | Observations |
|----------|------|-------|-----------|-----------|--------------|
| IA-2026-01 | 2026-03-10/11 | Full ISMS (Clauses 4-10, Annex A) | TBD (audit in progress) | TBD | TBD |

> The internal audit is being conducted on 2026-03-10/11 by an external auditor. Findings will be
> added to `09-corrective-actions-log.csv` and incorporated into this review when the audit report
> is finalized.

**External Audits:**

| Audit | Date | Result | Findings |
|-------|------|--------|----------|
| SOC 2 Type II | Not yet conducted | Planned for 2026 | N/A |
| ISO 27001 Stage 1 | Not yet conducted | Target: 2026-03-15 | N/A |
| ISO 27001 Stage 2 | Not yet conducted | Target: 2026-04-15 | N/A |

---

#### 1.3.4 Fulfillment of Information Security Objectives

| Objective | Target | Actual | Status |
|-----------|--------|--------|--------|
| Protect customer trust / No breaches | No breaches | 0 security incidents in operating period | **Met** |
| Maintain compliance / SOC 2 + ISO 27001 | Certification by 2026-05-01 | ISMS operational; internal audit in progress; 81% control implementation | **On Track** |
| Enable secure development / Security in CI/CD | CI enforcement on all repos | Branch protection + required status checks on all 3 repos since 2026-01-17 | **Met** |
| Manage risks / No critical open risks | No critical unmitigated risks | 0 critical inherent risks; 4 high risks all have treatments | **Met** (inherent risk managed) |
| Continuous improvement / Annual review complete | Review conducted | This is the first review; CA-001 closed demonstrating improvement cycle | **Met** (first cycle) |
| Secure development / Critical vulns | 0 open > 7 days | 1 high in aegis-platform; 2 critical in aegis-ui (upstream dependency) | **Not Met** (aegis-ui) |
| Access control / MFA coverage | 100% | GitHub 100%; AWS root needs verification | **Partially Met** |

---

### 1.4 Feedback from Interested Parties

| Stakeholder | Feedback | Action Required |
|-------------|----------|-----------------|
| Customers | No external customers yet (pre-revenue startup) | None currently; build security posture proactively |
| Auditors | Internal audit in progress (2026-03-10/11); no prior audit feedback | Incorporate findings when available |
| Regulators | No regulatory interactions; FedRAMP/CMMC preparation proactive | Continue compliance preparation |
| Partners | No formal partnerships; AWS and GitHub are key vendors | Annual vendor review scheduled for Q2 2026 |
| Pilot users | Pilot deployment planned for v1.0; no security feedback yet | Establish feedback channel for pilot |

---

### 1.5 Risk Assessment and Treatment Status

#### Risk Register Summary

| Risk Level (Inherent) | Count | Change from Last Review |
|------------------------|-------|------------------------|
| Critical | 0 | N/A (first review) |
| High | 4 | N/A (first review) |
| Medium | 12 | N/A (first review) |
| Low | 3 | N/A (first review) |
| **Total** | **19** | **Established 2026-01-17** |

| Risk Level (Residual, after treatment) | Count |
|-----------------------------------------|-------|
| Critical | 0 |
| High | 0 |
| Medium | 5 (RISK-001, RISK-002, RISK-008, RISK-014, + RISK-009 partial) |
| Low | 14 |

#### Top 5 Risks (by inherent risk score)

| Rank | Risk ID | Description | Inherent Score | Residual Score | Treatment Status |
|------|---------|-------------|----------------|----------------|------------------|
| 1 | RISK-001 | Supply chain vulnerability in dependencies | 15 (High) | 8 (Medium) | **Implemented** -- Dependabot active, 55 alerts fixed historically; 3 open in aegis-platform |
| 2 | RISK-008 | Phishing attack on founder (credential theft) | 12 (High) | 6 (Medium) | **In Progress** -- MFA active; FIDO2 hardware keys planned but not yet purchased |
| 3 | RISK-014 | Container image vulnerabilities in CI/CD | 12 (High) | 6 (Medium) | **Planned** -- Trivy scanning not yet added to CI; base image updates manual |
| 4 | RISK-002 | Solo founder single point of failure | 10 (High) | 8 (Medium) | **In Progress** -- Runbooks in git; password manager configured; key person insurance not yet evaluated |
| 5 | RISK-003 | No corporate identity provider | 9 (Medium) | 2 (Low) | **Planned** -- Google Workspace not yet set up; using personal accounts with MFA |

#### Risk Treatment Plan Progress

| Status | Count | Risks |
|--------|-------|-------|
| Implemented | 1 | RISK-001 |
| In Progress | 8 | RISK-002, RISK-007, RISK-008, RISK-009, RISK-011, RISK-013, RISK-015, RISK-018 |
| Planned | 7 | RISK-003, RISK-004, RISK-010, RISK-012, RISK-014, RISK-016, RISK-017 |
| Accepted | 2 | RISK-006, RISK-019 |
| Procedures Ready | 1 | RISK-005 |

#### Overdue Risk Treatments (due date passed, not yet completed)

| Risk ID | Description | Due Date | Status | Days Overdue |
|---------|-------------|----------|--------|--------------|
| RISK-007 | Credential exposure in code | 2026-02-15 | In Progress | 23 |
| RISK-008 | Phishing attack on founder (FIDO2 keys) | 2026-02-15 | In Progress | 23 |
| RISK-011 | Unpatched developer workstation | 2026-02-15 | In Progress | 23 |
| RISK-014 | Container vulnerabilities (Trivy in CI) | 2026-02-28 | Planned | 10 |
| RISK-002 | Solo founder SPOF | 2026-03-01 | In Progress | 9 |
| RISK-003 | No corporate identity provider | 2026-03-01 | Planned | 9 |
| RISK-004 | GitHub personal account limitations | 2026-03-01 | Planned | 9 |
| RISK-010 | Laptop theft/loss exposure | 2026-03-01 | Planned | 9 |
| RISK-013 | Insufficient logging | 2026-03-01 | In Progress | 9 |
| RISK-016 | Excessive API permissions | 2026-03-01 | Planned | 9 |

> **10 of 17 active risk treatments are overdue.** This reflects the startup reality of prioritizing
> product development alongside compliance buildout. Management decision required on re-prioritization
> and revised due dates.

#### Risk Acceptances in Effect

| Risk ID | Description | Residual Score | Accepted By | Date |
|---------|-------------|----------------|-------------|------|
| RISK-006 | AWS GovCloud outage | 4 (Low) | Carlos Sanchez | 2026-01-17 |
| RISK-019 | sovran coverage threshold disabled | 2 (Low) | Carlos Sanchez | 2026-01-17 |

---

### 1.6 Opportunities for Improvement

| # | Opportunity | Source | Priority | Recommendation |
|---|-------------|--------|----------|----------------|
| 1 | Automate vulnerability remediation for transitive dependencies | KPI gap: 46 open Dependabot alerts | High | Evaluate Dependabot auto-merge for patch/minor updates; schedule dedicated remediation sprint for aegis-ui and sovran |
| 2 | Implement container scanning in CI | RISK-014 overdue; no Trivy in pipeline | High | Add Trivy to GitHub Actions for all repos; block merges with critical/high vulnerabilities |
| 3 | Formalize quarterly access reviews | KPI gap: access reviews not yet executed | Medium | Complete first Q1 2026 access review before March 31; document in `soc2/2026/Q1/access-reviews/` |
| 4 | Verify and enable AWS root MFA | Evidence gap: `account_mfa_enabled: 0` | High | Confirm root MFA status; enable if not active; re-export evidence |
| 5 | Consolidate risk treatment due dates | 10 overdue treatments | Medium | Review all due dates in management review; set realistic revised dates based on pilot release timeline |
| 6 | Establish formal change management log | Audit preparation | Low | Track infrastructure changes beyond Git commits; useful for Stage 2 audit evidence |
| 7 | Create vendor security assessment | RISK-017 treatment planned | Low | Collect AWS and GitHub SOC 2 reports; document vendor review before Q2 2026 |
| 8 | Migrate to GitHub organization | RISK-004 treatment | Medium | Create org, transfer repos; improves access control, audit logging, and Advanced Security features |

---

## 2. Review Discussion (to be completed during review)

### 2.1 ISMS Effectiveness Assessment

**Overall ISMS Effectiveness:** *(to be determined by Top Management)*

- [ ] Effective
- [ ] Partially Effective
- [ ] Ineffective

**Pre-populated assessment for discussion:**

The ISMS has been operating for approximately 8 weeks since establishment. Key indicators of effectiveness:

**Positive indicators:**
- Zero security incidents since establishment
- Evidence collection operational and consistent (3 months of monthly evidence)
- CA-001 corrective action cycle completed successfully (opened and closed)
- CI enforcement active on all repositories with required status checks
- Branch protection preventing unreviewed merges
- 7 policies approved and in effect
- 66 of 81 applicable controls implemented (81%)
- Risk register comprehensive with 19 identified risks

**Areas requiring attention:**
- 10 of 17 active risk treatments overdue
- 46 open Dependabot vulnerabilities across 3 repos (2 critical in aegis-ui)
- Quarterly access reviews not yet formally executed
- AWS root MFA evidence gap
- No formal training acknowledgment records beyond founder self-certification

**Suggested rating:** Partially Effective -- the ISMS framework is sound and operating, but treatment execution has lagged behind planning due to product development priorities.

---

### 2.2 Resource Adequacy

**Are resources adequate for ISMS operation?** *(to be determined by Top Management)*

- [ ] Yes
- [ ] No

**Pre-populated resource assessment:**

| Resource | Current State | Adequacy | Action Needed |
|----------|--------------|----------|---------------|
| Personnel | 1 (founder) -- all roles | Adequate for current scope but strained | Prioritization of compliance tasks vs. product development |
| Budget | ~$20-30K/year allocated for compliance | Adequate for certification | Internal audit ($2-5K), Stage 1+2 ($8-15K) within budget |
| Tools | GitHub, AWS, Dependabot, scripts | Adequate | Consider adding Trivy (free), FIDO2 keys ($50-100) |
| Time | ~10 hrs/month ISMS maintenance | Strained | 10 overdue treatments suggest time allocation insufficient; recommend dedicated compliance sprint |

---

### 2.3 Policy Review

**Is the Information Security Policy still appropriate?** *(to be determined by Top Management)*

- [ ] Yes
- [ ] No

**Pre-populated assessment:** All 7 policies were approved on 2026-01-17 and have been in effect for ~8 weeks. No regulatory changes, organizational changes, or incidents have occurred that would necessitate policy updates. The policies remain appropriate for the current organizational context.

| Policy | ID | Status | Change Needed? |
|--------|-----|--------|----------------|
| Information Security Policy | ISP-001 | Current | No |
| Access Control Policy | ACP-001 | Current | No |
| Change Management Policy | CMP-001 | Current | No |
| Incident Response Policy | IRP-001 | Current | No |
| Risk Management Policy | RMP-001 | Current | No |
| Vendor Management Policy | VMP-001 | Current | No |
| Acceptable Use Policy | ISMS-POL-AUP-001 | Current | No |

---

## 3. Required Outputs (Clause 9.3.3)

> Sections 3.1 through 3.4 are to be completed by Top Management during the review meeting.
> Pre-populated recommendations are provided as starting points for discussion.

### 3.1 Decisions on Continual Improvement Opportunities

| # | Improvement | Priority | Owner | Recommended Due Date |
|---|-------------|----------|-------|----------------------|
| 1 | Remediate aegis-ui critical Dependabot alerts (fast-xml-parser) | High | Carlos Sanchez | 2026-03-22 |
| 2 | Add Trivy container scanning to CI pipeline (RISK-014) | High | Carlos Sanchez | 2026-04-15 |
| 3 | Verify and enable AWS root MFA | High | Carlos Sanchez | 2026-03-17 |
| 4 | Complete Q1 2026 formal access review | Medium | Carlos Sanchez | 2026-03-31 |
| 5 | Purchase and register FIDO2 security keys (RISK-008) | Medium | Carlos Sanchez | 2026-04-01 |
| 6 | Create GitHub organization and transfer repos (RISK-004) | Medium | Carlos Sanchez | 2026-06-01 |
| 7 | Conduct sovran dependency remediation sprint | Low | Carlos Sanchez | 2026-05-01 |

---

### 3.2 Decisions on Changes to the ISMS

| Change | Scope | Impact | Approval |
|--------|-------|--------|----------|
| Revise risk treatment due dates for 10 overdue items | Risk treatment plan | Administrative -- aligns plan with reality | [ ] Approved  [ ] Rejected |
| Add RISK-020: AWS root MFA not confirmed | Risk register | New risk if root MFA is indeed disabled | [ ] Approved  [ ] Rejected |
| Incorporate internal audit findings into corrective actions | CA log | New CAs from audit findings | [ ] Approved  [ ] Rejected |

---

### 3.3 Resource Decisions

| Resource Request | Justification | Decision |
|------------------|---------------|----------|
| External internal auditor engagement ($2-5K) | Required for Clause 9.2 independence (solo founder) | [ ] Approved  [ ] Rejected |
| FIDO2 hardware security keys ($50-100) | RISK-008 treatment; phishing-resistant MFA | [ ] Approved  [ ] Rejected |
| Dedicated compliance sprint (1 week) | Address 10 overdue risk treatments + access review | [ ] Approved  [ ] Rejected |
| ISO 27001 certification audit ($8-15K) | Stage 1 + Stage 2 audit fees | [ ] Approved  [ ] Rejected |

---

### 3.4 Action Items

| # | Action | Owner | Recommended Due Date | Priority |
|---|--------|-------|----------------------|----------|
| 1 | Verify AWS root account MFA status and enable if disabled | Carlos Sanchez | 2026-03-17 | Critical |
| 2 | Remediate 2 critical Dependabot alerts in aegis-ui (fast-xml-parser) | Carlos Sanchez | 2026-03-22 | High |
| 3 | Complete Q1 2026 quarterly access review with formal documentation | Carlos Sanchez | 2026-03-31 | High |
| 4 | Review and revise all overdue risk treatment due dates | Carlos Sanchez | 2026-03-17 | High |
| 5 | Incorporate internal audit findings into corrective actions log | Carlos Sanchez | 2026-03-22 (1 week after audit) | High |
| 6 | Add Trivy container scanning to GitHub Actions CI pipeline | Carlos Sanchez | 2026-04-15 | Medium |
| 7 | Purchase FIDO2 hardware security keys and register on critical accounts | Carlos Sanchez | 2026-04-01 | Medium |
| 8 | Evaluate and remediate sovran Dependabot alerts (30 open) | Carlos Sanchez | 2026-05-01 | Low |
| 9 | Schedule ISO 27001 Stage 1 audit with certification body | Carlos Sanchez | 2026-04-01 | High |
| 10 | Update STATUS.md to reflect audit and management review completion | Carlos Sanchez | 2026-03-17 | Low |

---

## 4. Next Review

**Next Management Review Date:** 2026-09-15 (semi-annual per operating plan)

**Focus areas for next review:**
1. Internal audit findings closure and corrective action effectiveness
2. ISO 27001 certification audit results (Stage 1 and Stage 2)
3. Vulnerability management trend -- whether Dependabot alert counts decreased
4. Risk treatment plan completion rate (target: all overdue items resolved)
5. First customer/pilot feedback on security posture (if applicable)

---

## 5. Approval

This management review has been conducted in accordance with ISO 27001:2022 Clause 9.3.

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Top Management | Carlos Sanchez | | |
| Information Security Manager | Carlos Sanchez | | |

---

## Appendices

### Appendix A: KPI Dashboard Summary

| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| Open critical vulns (all repos) | 0 | 2 | Not Met |
| Open high vulns (all repos) | <5 | 34 | Not Met |
| Open vulns total (all repos) | <15 | 46 | Not Met |
| Open vulns (aegis-platform only) | <5 | 3 | Met |
| Security incidents (Q1) | <2 | 0 | Met |
| MFA coverage (GitHub) | 100% | 100% | Met |
| MFA coverage (AWS root) | 100% | Unconfirmed | Review |
| Policy acknowledgment | 100% | 100% | Met |
| Access reviews completed | Q4 2025 + Q1 2026 | Structure only | Not Met |
| Training completion | 100% | 100% | Met |
| Evidence months collected | 3/3 | 3/3 | Met |
| Overdue risk treatments | 0 | 10 | Not Met |
| Controls implemented | 81/81 | 66/81 (81%) | Partial |
| Corrective actions closed | All | 1/2 (CA-002 open, due Apr 15) | On Track |

### Appendix B: Risk Register Summary

**Total risks:** 19
**Inherent risk distribution:** 0 Critical, 4 High, 12 Medium, 3 Low
**Residual risk distribution:** 0 Critical, 0 High, 5 Medium, 14 Low
**Treatments implemented:** 1 | **In progress:** 8 | **Planned:** 7 | **Accepted:** 2 | **Ready:** 1
**Overdue treatments:** 10

### Appendix C: Audit Reports

- Internal audit report (IA-2026-01): In progress, 2026-03-10/11
- No external audit reports yet

### Appendix D: Incident Summary

| Month | Incidents | P1 | P2 | P3 | P4 |
|-------|-----------|----|----|----|----|
| January 2026 | 0 | 0 | 0 | 0 | 0 |
| February 2026 | 0 | 0 | 0 | 0 | 0 |
| March 2026 (to date) | 0 | 0 | 0 | 0 | 0 |
| **Total** | **0** | **0** | **0** | **0** | **0** |

### Appendix E: Evidence Vault Statistics

| Metric | Value |
|--------|-------|
| Total files in vault | 472 |
| 2025 retroactive evidence files | 27 |
| 2026 evidence files (non-placeholder) | 140 |
| January 2026 evidence files | 50 |
| February 2026 evidence files | 20 |
| March 2026 evidence files | 69 |
| Monthly incident logs present | 3/3 (Jan, Feb, Mar) |
| Dependabot evidence exports | Present for all 3 repos, all 3 months |
| AWS security baselines | Present for Jan and Mar |
| Branch protection evidence | Present for all 3 repos |
| SBOM evidence | Present (3 Go modules, Mar 2026) |

### Appendix F: Dependabot Alert Detail (as of 2026-03-10)

**aegis-platform (3 open, 11 fixed):**
| # | Severity | Summary |
|---|----------|---------|
| 14 | High | OpenTelemetry Go SDK Vulnerable to Arbitrary Code Execution via PATH Hijacking |
| 13 | Low | CIRCL incorrect calculation in secp384r1 CombinedMult |
| 12 | Medium | go-git improperly verifies data integrity values for .idx and .pack files |

**aegis-ui (13 open):**
| Severity | Count | Key Issues |
|----------|-------|------------|
| Critical | 2 | fast-xml-parser entity encoding bypass via regex injection in DOCTYPE |
| High | 7 | fast-xml-parser DoS, TechDocs Mkdocs arbitrary code execution, minimatch ReDoS, Koa Host Header Injection |
| Low | 4 | Various low-severity issues |

**sovran (30 open):**
| Severity | Count | Key Issues |
|----------|-------|------------|
| High | 26 | Primarily transitive dependencies in VS Code extension ecosystem |
| Medium | 4 | Various medium-severity issues |

### Appendix G: Recent Development Activity

**Last 5 merged PRs (aegis-platform):**
| PR | Title | Merged |
|----|-------|--------|
| #60 | fix(ci): add envtest binaries for controller tests | 2026-01-18 |
| #58 | MVP-58: internal PKI via step-ca + cert-manager | 2025-12-18 |
| #57 | MVP-65: Suspend workspaces on idle timeout | 2025-12-16 |
| #56 | MVP-66: Import existing cluster API | 2025-12-15 |
| #55 | MVP-59: Hub-to-spoke workload garbage collection | 2025-12-15 |

**Recent commits on release/v1.0-pilot (unpushed):**
- `057a0f0` fix: 8 deployment fixes for zero-intervention deploy-cloud-full
- `85ff9c7` fix: resolve 5 supervisor findings from buildout review
- `bb22086` Add CMMC program structure; complete FedRAMP free prep
- `86597ed` compliance: execute supervisor review action items (Mar 9-15)
- `02f7e6c` feat: pilot-ready P0+P1 implementation (audit system, compliance enforcement, production hardening)

---

## Solo Founder Note

> As a solo-founder company, Carlos Sanchez fulfills both the "Top Management" and "Information
> Security Manager" roles. This management review is conducted as a formal, documented self-review
> to satisfy ISO 27001:2022 Clause 9.3 requirements. The review is conducted with the same rigor as
> if multiple stakeholders were involved, and all decisions are documented for audit evidence.
>
> This is the FIRST management review. The ISMS was established on 2026-01-17 and has been operating
> for approximately 8 weeks. The review covers all required inputs per Clause 9.3.2 and provides
> pre-populated data to enable efficient decision-making by Top Management.

---

## Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-10 | AI Agent (Claude Opus 4.6) | Initial management review inputs with real data from evidence vault, GitHub, risk register |
