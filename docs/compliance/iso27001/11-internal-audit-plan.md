# Internal Audit Plan - March 2026

**Document ID:** ISMS-IAP-2026-01
**Audit Period:** March 10-11, 2026
**Prepared By:** Carlos Sanchez
**Audit Type:** Full ISMS Internal Audit (First Audit) — Self-Assessment

---

## 1. Audit Objective

To verify that Aegis Technologies' Information Security Management System:

1. Conforms to the requirements of ISO/IEC 27001:2022
2. Conforms to the organization's own ISMS requirements
3. Is effectively implemented and maintained
4. Provides input for the first management review

---

## 2. Audit Scope

### 2.1 Organizational Scope
- Aegis Technologies (solo founder operation)
- All ISMS-related activities

### 2.2 System Scope
- GitHub repositories (aegis-platform, aegis-ui, sovran)
- AWS GovCloud development infrastructure
- Developer workstation
- Corporate systems (email, documentation)

### 2.3 Process Scope
- All ISO 27001:2022 mandatory clauses (4-10)
- Selected Annex A controls (sample-based)

### 2.4 Exclusions
- Customer-deployed environments (Model B - out of ISMS scope)
- Physical security controls (no premises - A.7.1-7.4, A.7.6, A.7.11-12, A.7.14)

---

## 3. Audit Criteria

- ISO/IEC 27001:2022 Clauses 4-10
- ISO/IEC 27001:2022 Annex A (applicable controls per SoA)
- Aegis ISMS policies and procedures
- SOC 2 Trust Services Criteria (for overlap controls)

---

## 4. Audit Team

| Role | Name | Organization | Qualification |
|------|------|--------------|---------------|
| Lead Auditor | Carlos Sanchez | Aegis Technologies | Founder/ISM (Self-Assessment) |
| AI-Assisted Review | Claude (AI Agent) | Anthropic | Systematic evidence verification |

**Note:** This audit is conducted as a documented self-assessment, acceptable for startups building toward certification. The AI agent provides systematic, objective evidence verification. An external auditor will be engaged for the Stage 1/Stage 2 certification audit.

---

## 5. Audit Schedule

| Date | Time | Activity | Clause/Area | Auditee |
|------|------|----------|-------------|---------|
| Day 1 | 09:00-09:30 | Opening meeting | - | Carlos Sanchez |
| Day 1 | 09:30-11:00 | Documentation review | Clauses 4-5 | Carlos Sanchez |
| Day 1 | 11:00-12:00 | Risk management | Clause 6, A.5.7 | Carlos Sanchez |
| Day 1 | 13:00-14:30 | Support & operations | Clauses 7-8 | Carlos Sanchez |
| Day 1 | 14:30-16:00 | Control testing - Access | A.5.15-18, A.8.2-5 | Carlos Sanchez |
| Day 2 | 09:00-10:30 | Control testing - Development | A.8.25-32 | Carlos Sanchez |
| Day 2 | 10:30-12:00 | Control testing - Vulnerabilities | A.8.7-8 | Carlos Sanchez |
| Day 2 | 13:00-14:00 | Performance evaluation | Clause 9 | Carlos Sanchez |
| Day 2 | 14:00-15:00 | Improvement | Clause 10 | Carlos Sanchez |
| Day 2 | 15:00-16:00 | Closing meeting | - | Carlos Sanchez |

---

## 6. Audit Resources

### 6.1 Documents to Prepare

| Document | Location | Prepared? |
|----------|----------|-----------|
| ISMS Scope | `iso27001/00-isms-scope-and-context.md` | ✅ |
| Leadership & Roles | `iso27001/01-leadership-policy-roles.md` | ✅ |
| Risk Methodology | `iso27001/02-risk-methodology.md` | ✅ |
| Risk Register | `iso27001/03-risk-register.csv` | ✅ |
| Risk Treatment Plan | `iso27001/04-risk-treatment-plan.csv` | ✅ |
| Statement of Applicability | `iso27001/05-statement-of-applicability.csv` | ✅ |
| All Policies | `soc2/policies/` | ✅ |
| Procedures | `soc2/procedures/` | ✅ |
| Internal Audit Program | `iso27001/07-internal-audit-program.md` | ✅ |
| Management Review Template | `iso27001/08-management-review-template.md` | ✅ |
| Corrective Actions Log | `iso27001/09-corrective-actions-log.csv` | ✅ |
| ISMS Operating Plan | `iso27001/10-isms-operating-plan.md` | ✅ |

### 6.2 Evidence to Prepare

| Evidence Type | Location | Status |
|---------------|----------|--------|
| Monthly evidence exports | `aegis-compliance-evidence/soc2/2025/` | ⚠️ Verify |
| Access review records | `aegis-compliance-evidence/soc2/YYYY/QX/` | ⚠️ Verify |
| Vulnerability reports | GitHub Dependabot | ✅ Available |
| Change records (PRs) | GitHub | ✅ Available |
| Training records | `aegis-compliance-evidence/soc2/YYYY/training/` | ⚠️ Verify |
| Incident logs | `aegis-compliance-evidence/soc2/YYYY/MM/incident-response/` | ⚠️ Verify |

---

## 7. Control Sample Selection

### 7.1 Sampling Approach

Given the solo founder operation and limited transaction volume:
- **100% review** for: policies, procedures, risk register
- **Sample-based** for: evidence (3-month sample), PRs (10 samples), access reviews (all quarters available)

### 7.2 Sample Selection List

| Control Area | Sample Type | Sample Size | Selection Method |
|--------------|-------------|-------------|------------------|
| Access reviews | Quarterly reviews | All available | 100% (limited data) |
| Change management | Pull requests | 10 PRs | Random across repos |
| Vulnerability management | Dependabot alerts | 20 alerts | Random closed alerts |
| Incident response | Incident logs | All | 100% (expected: none) |
| Training | Training records | All | 100% (solo founder) |
| Monthly evidence | Evidence exports | 3 months | Sept, Nov 2025, Jan 2026 |

### 7.3 Pre-Selected Samples

#### Pull Request Samples (10)

| # | Repo | PR Number | Selection Basis |
|---|------|-----------|-----------------|
| 1 | aegis-platform | [Latest] | Most recent |
| 2 | aegis-platform | [Random] | Random selection |
| 3 | aegis-platform | [Random] | Random selection |
| 4 | aegis-platform | [Random] | Random selection |
| 5 | aegis-ui | [Latest] | Most recent |
| 6 | aegis-ui | [Random] | Random selection |
| 7 | aegis-ui | [Random] | Random selection |
| 8 | sovran | [Latest] | Most recent |
| 9 | sovran | [Random] | Random selection |
| 10 | sovran | [Random] | Random selection |

**Note:** Actual PR numbers to be populated before audit by running:
```bash
gh pr list --repo carlosmsanchezm/aegis-platform --state merged --limit 20
```

#### Dependabot Alert Samples (20)

| # | Repo | Alert ID | Severity | Resolution |
|---|------|----------|----------|------------|
| 1-7 | aegis-platform | [Random 7] | Mixed | Fixed |
| 8-17 | aegis-ui | [Random 10] | Mixed | Fixed |
| 18-20 | sovran | [Random 3] | Mixed | Fixed |

**Note:** Select from 55 total fixed alerts

#### Evidence Export Samples (3 months)

| Month | Evidence Types to Review |
|-------|--------------------------|
| September 2025 | Retroactive git history, PRs, CI runs |
| November 2025 | Monthly exports (if created) |
| January 2026 | Current month exports |

---

## 8. Audit Checklist

See: `iso27001/12-internal-audit-checklist.md`

---

## 9. Communication Plan

| Milestone | Date | Deliverable |
|-----------|------|-------------|
| Audit plan issued | 2026-01-17 | This document (original) |
| Audit plan updated | 2026-03-10 | Updated for self-assessment |
| Pre-audit preparation | 2026-03-09 | Evidence gathered (452 evidence files in vault) |
| Audit execution | 2026-03-10-11 | AI-assisted systematic evidence review |
| Final report | 2026-03-11 | Completed audit report |
| CA submission | 2026-03-11 | Corrective action plans updated |
| Management review | 2026-03-15 | Management review using audit inputs |

---

## 10. Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Auditee / Lead Auditor | Carlos Sanchez | /s/ Carlos Sanchez | 2026-03-10 |
| AI-Assisted Review | Claude (AI Agent) | /s/ AI-Assisted | 2026-03-10 |

---

## 11. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 2.0 | 2026-03-10 | Claude Opus 4.6 | Updated for self-assessment; revised dates to Mar 10-11; AI-assisted systematic review |
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial audit plan |
