# Internal Audit Report - February 2026 (Pre-filled Template)

**Document ID:** ISMS-IAR-2026-01
**Audit Date:** February 17-21, 2026
**Report Date:** [TO BE COMPLETED]
**Lead Auditor:** [External Auditor TBD]
**Classification:** Confidential

---

## 1. Executive Summary

### 1.1 Audit Objective
To verify conformance of the Aegis Technologies ISMS to ISO/IEC 27001:2022 requirements and provide input for the first management review.

### 1.2 Audit Scope
- ISO/IEC 27001:2022 Clauses 4-10
- Selected Annex A controls (per sampling plan)
- Systems: GitHub (3 repos), AWS GovCloud, developer workstation
- Exclusions: Customer environments (Model B), physical security (no premises)

### 1.3 Audit Criteria
- ISO/IEC 27001:2022
- Aegis ISMS policies and procedures
- Statement of Applicability (93 controls)

### 1.4 Summary of Findings

| Category | Count |
|----------|-------|
| Major Nonconformities | [#] |
| Minor Nonconformities | [#] |
| Observations/OFIs | [#] |
| Positive Findings | [#] |

### 1.5 Overall Audit Opinion
[TO BE COMPLETED BY AUDITOR]

---

## 2. Audit Details

### 2.1 Audit Team

| Role | Name | Organization |
|------|------|--------------|
| Lead Auditor | [TBD] | [External Audit Firm] |
| Auditee | Carlos Sanchez | Aegis Technologies |

### 2.2 Auditee Representatives

| Name | Role | Areas Covered |
|------|------|---------------|
| Carlos Sanchez | Founder/Information Security Manager | All areas |

### 2.3 Audit Schedule

| Date | Time | Area/Clause | What to Check |
|------|------|-------------|---------------|
| Day 1 | 09:00-09:30 | Opening meeting | Confirm scope, logistics |
| Day 1 | 09:30-11:00 | Clause 4-5 (Context, Leadership) | `00-isms-scope-and-context.md`, `01-leadership-policy-roles.md`, policies |
| Day 1 | 11:00-12:00 | Clause 6 (Risk) | `02-risk-methodology.md`, `03-risk-register.csv`, `04-risk-treatment-plan.csv`, `05-statement-of-applicability.csv` |
| Day 1 | 13:00-14:30 | Clause 7-8 (Support, Operations) | Training records, evidence vault, procedures |
| Day 1 | 14:30-16:00 | Access controls | A.5.15-18, A.8.2-5: GitHub/AWS access exports, MFA status |
| Day 2 | 09:00-10:30 | Development controls | A.8.25-32: 10 PR samples, branch protection, CODEOWNERS |
| Day 2 | 10:30-12:00 | Vulnerability management | A.8.7-8: 20 Dependabot alert samples, container scanning |
| Day 2 | 13:00-14:00 | Clause 9 (Performance) | Audit programme, management review (template only - first audit) |
| Day 2 | 14:00-15:00 | Clause 10 (Improvement) | `09-corrective-actions-log.csv`, improvement evidence |
| Day 2 | 15:00-16:00 | Closing meeting | Review findings, agree classifications |

### 2.4 Documents Reviewed

| Document | Location | Reviewed? |
|----------|----------|-----------|
| ISMS Scope and Context | `iso27001/00-isms-scope-and-context.md` | ☐ |
| Leadership, Policy, Roles | `iso27001/01-leadership-policy-roles.md` | ☐ |
| Risk Methodology | `iso27001/02-risk-methodology.md` | ☐ |
| Risk Register (17 risks) | `iso27001/03-risk-register.csv` | ☐ |
| Risk Treatment Plan | `iso27001/04-risk-treatment-plan.csv` | ☐ |
| Statement of Applicability (93 controls) | `iso27001/05-statement-of-applicability.csv` | ☐ |
| Information Security Policy | `soc2/policies/information-security-policy.md` | ☐ |
| Access Control Policy | `soc2/policies/access-control-policy.md` | ☐ |
| Change Management Policy | `soc2/policies/change-management-policy.md` | ☐ |
| Incident Response Policy | `soc2/policies/incident-response-policy.md` | ☐ |
| Risk Management Policy | `soc2/policies/risk-management-policy.md` | ☐ |
| Vendor Management Policy | `soc2/policies/vendor-management-policy.md` | ☐ |
| Acceptable Use Policy | `iso27001/policies/acceptable-use-policy.md` | ☐ |
| Internal Audit Program | `iso27001/07-internal-audit-program.md` | ☐ |
| Management Review Template | `iso27001/08-management-review-template.md` | ☐ |
| ISMS Operating Plan | `iso27001/10-isms-operating-plan.md` | ☐ |
| Onboarding Checklist | `soc2/procedures/employee-onboarding-checklist.md` | ☐ |
| Offboarding Checklist | `soc2/procedures/employee-offboarding-checklist.md` | ☐ |
| Access Review Procedure | `soc2/procedures/quarterly-access-review.md` | ☐ |

---

## 3. Key Areas to Examine

### 3.1 High-Risk Areas (Focus Attention)

| Area | Risk | What to Verify | Where to Look |
|------|------|----------------|---------------|
| Evidence vault operation | RISK-015 | Is vault populated? Are exports running? | `aegis-compliance-evidence/` |
| Management review | Clause 9.3 | First review conducted? (may be pending) | `iso27001/management-reviews/` |
| Corrective actions | Clause 10.1 | At least one CA cycle complete? | `09-corrective-actions-log.csv` |
| Solo founder controls | Multiple | Compensating controls documented? | Scope doc, procedures |
| Access reviews | A.5.18 | Quarterly reviews occurring? | Evidence vault |

### 3.2 Evidence Sampling Plan

#### Pull Request Samples (10 total)

**Commands to select samples:**
```bash
# Get recent PRs for sampling
gh pr list --repo carlosmsanchezm/aegis-platform --state merged --limit 20 --json number,title,mergedAt
gh pr list --repo carlosmsanchezm/aegis-ui --state merged --limit 10 --json number,title,mergedAt
gh pr list --repo carlosmsanchezm/sovran --state merged --limit 10 --json number,title,mergedAt
```

**For each PR, verify:**
- [ ] Branch protection was enforced
- [ ] PR review was required (or documented exception for solo founder)
- [ ] CI checks passed before merge
- [ ] Appropriate reviewer approved
- [ ] Merge to protected branch followed process

| # | Repo | PR # | Verified? | Findings |
|---|------|------|-----------|----------|
| 1 | aegis-platform | | ☐ | |
| 2 | aegis-platform | | ☐ | |
| 3 | aegis-platform | | ☐ | |
| 4 | aegis-platform | | ☐ | |
| 5 | aegis-ui | | ☐ | |
| 6 | aegis-ui | | ☐ | |
| 7 | aegis-ui | | ☐ | |
| 8 | sovran | | ☐ | |
| 9 | sovran | | ☐ | |
| 10 | sovran | | ☐ | |

#### Dependabot Alert Samples (20 total)

**Commands to get closed alerts:**
```bash
# List resolved Dependabot alerts
gh api repos/carlosmsanchezm/aegis-platform/dependabot/alerts --jq '[.[] | select(.state=="fixed")] | .[0:10]'
gh api repos/carlosmsanchezm/aegis-ui/dependabot/alerts --jq '[.[] | select(.state=="fixed")] | .[0:10]'
gh api repos/carlosmsanchezm/sovran/dependabot/alerts --jq '[.[] | select(.state=="fixed")] | .[0:5]'
```

**For each alert, verify:**
- [ ] Alert was addressed within SLA (7 days critical, 30 days high)
- [ ] Fix was appropriate (update, not dismiss)
- [ ] No recurring issues

| # | Repo | Alert | Severity | Days to Fix | Verified? |
|---|------|-------|----------|-------------|-----------|
| 1-11 | aegis-platform | [11 alerts] | Mixed | | ☐ |
| 12-36 | aegis-ui | [25 alerts] | Mixed | | ☐ |
| 37-55 | sovran | [19 alerts] | Mixed | | ☐ |

#### Monthly Evidence Samples (3 months)

**Verify evidence exists for:**

| Month | Evidence Type | Location | Exists? |
|-------|---------------|----------|---------|
| Sept 2025 | Git history export | `soc2/2025/retroactive/` | ☐ |
| Sept 2025 | PR export | `soc2/2025/retroactive/` | ☐ |
| Nov 2025 | Monthly exports | `soc2/2025/2025-11/` | ☐ |
| Nov 2025 | Incident log | `soc2/2025/2025-11/incident-response/` | ☐ |
| Jan 2026 | Access review exports | `soc2/2026/2026-01/access-reviews/` | ☐ |
| Jan 2026 | Vuln management | `soc2/2026/2026-01/vuln-management/` | ☐ |

---

## 4. Pre-Identified Potential Findings

Based on ISMS status review, the following areas may result in findings:

### 4.1 Expected Nonconformities (if not addressed before audit)

| Potential NC | Clause | Risk | Mitigation |
|--------------|--------|------|------------|
| Management review not conducted | Clause 9.3 | High | Schedule before audit |
| Evidence vault incomplete | Clause 7.5, A.5.33 | Medium | Run evidence scripts |
| No corrective action cycle | Clause 10.1 | Medium | CA-001 opened for vault |

### 4.2 Expected Observations

| Potential OFI | Clause | Recommendation |
|---------------|--------|----------------|
| Solo founder lacks segregation | A.5.3 | Documented compensating controls |
| Google Workspace not yet implemented | A.5.16 | Plan exists, timeline documented |
| No container scanning yet | A.8.7 | RISK-014 tracked, planned |

---

## 5. Clause-by-Clause Summary

| Clause | Title | Status | Key Evidence | Notes |
|--------|-------|--------|--------------|-------|
| 4.1 | Context - external/internal | ☐ | `00-isms-scope-and-context.md` S2 | |
| 4.2 | Interested parties | ☐ | `00-isms-scope-and-context.md` S3 | |
| 4.3 | ISMS scope | ☐ | `00-isms-scope-and-context.md` S4 | Model B only |
| 4.4 | ISMS | ☐ | Full document set | |
| 5.1 | Leadership commitment | ☐ | `01-leadership-policy-roles.md` | Solo founder signs all |
| 5.2 | Policy | ☐ | 7 policies approved | |
| 5.3 | Roles | ☐ | `01-leadership-policy-roles.md` S4 | Solo founder = all roles |
| 6.1 | Risk actions | ☐ | Risk register, SoA, treatment plan | 17 risks, 93 controls |
| 6.2 | Objectives | ☐ | `01-leadership-policy-roles.md` S3.2 | 4 objectives |
| 6.3 | Change planning | ☐ | Change management policy | |
| 7.1 | Resources | ☐ | `01-leadership-policy-roles.md` S6 | |
| 7.2 | Competence | ☐ | Training records | Solo founder trained |
| 7.3 | Awareness | ☐ | Training records | |
| 7.4 | Communication | ☐ | `01-leadership-policy-roles.md` S5 | |
| 7.5 | Documented information | ☐ | Full doc set, evidence vault | |
| 8.1 | Operational planning | ☐ | `10-isms-operating-plan.md` | |
| 8.2 | Risk assessment | ☐ | Risk register | 17 risks assessed |
| 8.3 | Risk treatment | ☐ | Treatment plan | All risks have treatment |
| 9.1 | Monitoring | ☐ | Monthly evidence | |
| 9.2 | Internal audit | ☐ | This audit | First audit |
| 9.3 | Management review | ☐ | **PENDING** | Schedule after audit |
| 10.1 | Improvement | ☐ | CA log | CA-001 opened |
| 10.2 | Nonconformity/CA | ☐ | CA log | |

---

## 6. Annex A Sample Results

| Control | Title | Status | Evidence | Verified? |
|---------|-------|--------|----------|-----------|
| A.5.1 | Policies | Implemented | 7 policies | ☐ |
| A.5.7 | Threat intelligence | Implemented | Dependabot | ☐ |
| A.5.15 | Access control | Implemented | ACP-001 | ☐ |
| A.5.17 | Authentication | Implemented | MFA exports | ☐ |
| A.5.18 | Access rights | Implemented | Quarterly reviews | ☐ |
| A.5.21 | Supply chain | Implemented | SBOM, Dependabot | ☐ |
| A.5.24 | Incident planning | Implemented | IRP-001 | ☐ |
| A.8.4 | Source code access | Implemented | Branch protection | ☐ |
| A.8.5 | Secure auth | Implemented | MFA | ☐ |
| A.8.8 | Vulnerability mgmt | Implemented | 55 vulns fixed | ☐ |
| A.8.28 | Secure coding | Implemented | PR reviews | ☐ |
| A.8.32 | Change management | Implemented | PR workflow | ☐ |

---

## 7. Detailed Findings

### 7.1 Major Nonconformities

[TO BE COMPLETED BY AUDITOR]

### 7.2 Minor Nonconformities

[TO BE COMPLETED BY AUDITOR]

### 7.3 Observations / Opportunities for Improvement

[TO BE COMPLETED BY AUDITOR]

### 7.4 Positive Findings

**Expected positive findings based on preparation:**

1. **Comprehensive documentation**: Full ISMS document set established
2. **Automated security**: Dependabot, branch protection, CI/CD security
3. **Vulnerability management**: 55 vulnerabilities identified and fixed
4. **SOC 2 foundation**: Strong control overlap with ISO 27001
5. **Risk management**: 17 risks identified with treatment plans

---

## 8. Corrective Action Requirements

| Finding ID | Type | Clause | Due Date | Owner |
|------------|------|--------|----------|-------|
| | Major NC | | 30 days | Carlos Sanchez |
| | Minor NC | | 90 days | Carlos Sanchez |

---

## 9. Audit Conclusion

[TO BE COMPLETED BY AUDITOR]

---

## 10. Signatures

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Lead Auditor | [TBD] | _____________ | [TBD] |
| Auditee | Carlos Sanchez | _____________ | [TBD] |

---

## Appendix A: Evidence Collection Commands

```bash
# === GitHub Security Baseline ===
./scripts/compliance/export_github_security_baseline.sh

# === AWS Security Baseline ===
./scripts/compliance/export_aws_security_baseline.sh

# === PR Samples ===
gh pr list --repo carlosmsanchezm/aegis-platform --state merged --limit 20 --json number,title,mergedAt,reviewDecision

# === Dependabot Alerts ===
gh api repos/carlosmsanchezm/aegis-platform/dependabot/alerts --jq 'length'

# === Branch Protection ===
gh api repos/carlosmsanchezm/aegis-platform/branches/main/protection

# === MFA Status (GitHub) ===
# Requires org admin - check settings manually

# === Training Records ===
ls -la aegis-compliance-evidence/soc2/*/training-policy-ack/
```

---

## Appendix B: Key File Locations Reference

| Category | File | Path |
|----------|------|------|
| ISMS Scope | Context | `docs/compliance/iso27001/00-isms-scope-and-context.md` |
| ISMS | Leadership | `docs/compliance/iso27001/01-leadership-policy-roles.md` |
| Risk | Methodology | `docs/compliance/iso27001/02-risk-methodology.md` |
| Risk | Register | `docs/compliance/iso27001/03-risk-register.csv` |
| Risk | Treatment | `docs/compliance/iso27001/04-risk-treatment-plan.csv` |
| Controls | SoA | `docs/compliance/iso27001/05-statement-of-applicability.csv` |
| Audit | Program | `docs/compliance/iso27001/07-internal-audit-program.md` |
| Review | Template | `docs/compliance/iso27001/08-management-review-template.md` |
| Improvement | CA Log | `docs/compliance/iso27001/09-corrective-actions-log.csv` |
| Operations | Plan | `docs/compliance/iso27001/10-isms-operating-plan.md` |
| Policies | All | `docs/compliance/soc2/policies/` |
| Procedures | All | `docs/compliance/soc2/procedures/` |
| Evidence | Vault | `aegis-compliance-evidence/` |
