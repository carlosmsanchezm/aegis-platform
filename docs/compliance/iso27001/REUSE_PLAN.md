# ISO 27001 Reuse Plan from SOC 2 Foundation

**Document ID:** ISMS-REUSE-001
**Version:** 1.0
**Created:** 2026-01-17
**Author:** AI Agent + Carlos Sanchez

---

## 1. Executive Summary

This document identifies what SOC 2 artifacts can be reused for ISO 27001:2022, what ISO requires that SOC 2 didn't cover, and the gap list with owners and evidence pointers.

**Bottom Line:**
- ~85% of SOC 2 controls map to ISO 27001 Annex A
- 6 existing policies are directly reusable
- 4+ months of retroactive evidence applies to both standards
- 13 gaps require new documentation or controls
- **CRITICAL:** No internal audit or management review has been conducted yet - cannot claim "audit ready"

---

## 2. SOC 2 Assets Inventory (What We Have)

### 2.1 Documentation

| Asset | Location | ISO 27001 Reuse |
|-------|----------|-----------------|
| 00-scope-and-systems.md | `soc2/` | ✅ Maps to Clause 4.3 ISMS Scope |
| 01-control-narratives.md | `soc2/` | ✅ Maps to Annex A implementation |
| control-matrix.csv | `soc2/` | ✅ Foundation for SoA |
| policy-register.md | `soc2/` | ✅ Reuse for policy tracking |
| evidence-map.md | `soc2/` | ✅ Extend for ISO evidence |
| AGENT_CONTEXT.md | `soc2/` | ✅ Update with ISO context |

### 2.2 Policies (All Directly Reusable)

| Policy | SOC 2 ID | ISO 27001 Controls | Notes |
|--------|----------|-------------------|-------|
| Information Security Policy | ISP-001 | A.5.1, Clause 5.2 | Add ISO clause references |
| Access Control Policy | ACP-001 | A.5.15-18, A.8.2-5 | Complete |
| Change Management Policy | CMP-001 | A.8.32 | Complete |
| Incident Response Policy | IRP-001 | A.5.24-28 | Complete |
| Risk Management Policy | RMP-001 | Clause 6.1.2, A.5.7 | Complete |
| Vendor Management Policy | VMP-001 | A.5.19-23 | Complete |

### 2.3 Procedures (All Reusable)

| Procedure | Location | ISO 27001 Use |
|-----------|----------|---------------|
| employee-onboarding-checklist.md | `soc2/procedures/` | A.6.2, A.5.16 |
| employee-offboarding-checklist.md | `soc2/procedures/` | A.6.5, A.5.11 |
| quarterly-access-review.md | `soc2/procedures/` | A.5.18 |

### 2.4 Automation Scripts (Reusable)

| Script | SOC 2 Controls | ISO 27001 Controls |
|--------|----------------|-------------------|
| export_github_security_baseline.sh | CC6.1, CC6.7, CC7.1, CC8.1 | A.5.15-18, A.8.4, A.8.8, A.8.32 |
| export_aws_security_baseline.sh | CC6.1, CC6.6, CC7.2 | A.5.17, A.8.5, A.8.15 |
| export_ci_reports.sh | CC7.1, CC8.1 | A.8.8, A.8.25 |

### 2.5 Retroactive Evidence (Sept 2025 - Jan 2026)

| Evidence Type | Count | SOC 2 Controls | ISO 27001 Controls |
|---------------|-------|----------------|-------------------|
| Git commits | 725 (465+99+161) | CC8.1 | A.8.32 |
| Pull requests | 58+ | CC6.1, CC8.1 | A.8.4, A.8.32 |
| CI/CD runs | 200+ | CC7.2, CC8.1 | A.8.25, A.8.29 |
| Dependabot alerts fixed | 55 (11+25+19) | CC7.1 | A.8.8 |
| AWS security exports | 13 files | CC6.1, CC6.6, CC7.2 | A.5.17, A.8.5, A.8.15 |
| GitHub security exports | 10 files | CC6.1, CC6.7, CC8.1 | A.5.15-18, A.8.4 |

**Conclusion:** All retroactive evidence is directly applicable to ISO 27001 Annex A controls.

---

## 3. ISO 27001 Requirements Not Covered by SOC 2

### 3.1 ISMS Governance (Clauses 4-10)

| Clause | Requirement | SOC 2 Equivalent | Gap Status |
|--------|-------------|------------------|------------|
| 4.1 | Context of organization | Informal | 📋 Need formal statement |
| 4.2 | Interested parties | Informal | 📋 Need formal list |
| 4.3 | ISMS scope | Partial (00-scope-and-systems.md) | ✅ Adapt existing |
| 4.4 | ISMS | N/A | 📋 Need explicit declaration |
| 5.1 | Leadership commitment | CC1.1-1.2 | ✅ Covered |
| 5.2 | Information security policy | ISP-001 | ✅ Add clause references |
| 5.3 | Organizational roles | CC1.3 | ✅ Covered |
| 6.1.1 | Actions to address risks | CC3.1-3.4 | ✅ Covered |
| 6.1.2 | Risk assessment | CC3.2 | ✅ Formalize methodology |
| 6.1.3 | Risk treatment | CC3.2 | ✅ Add treatment plan |
| **6.1.3d** | **Statement of Applicability** | **N/A** | ✅ Created |
| 6.2 | Objectives | CC3.1 | 📋 Need formal objectives |
| 6.3 | Planning of changes | CC3.4, CC8.1 | ✅ Covered |
| 7.1 | Resources | Implicit | 📋 Document resources |
| 7.2 | Competence | CC1.4 | ✅ Covered |
| 7.3 | Awareness | CC2.1 | ✅ Covered |
| 7.4 | Communication | CC2.1-2.3 | ✅ Covered |
| 7.5 | Documented information | All | ✅ Covered |
| 8.1 | Operational planning | CC5.1-5.3 | ✅ Covered |
| 8.2 | Risk assessment execution | CC3.2 | ✅ Risk register exists |
| 8.3 | Risk treatment execution | CC3.2 | ✅ Treatment documented |
| **9.1** | **Monitoring, measurement** | **CC4.1** | 🔄 Formalize metrics |
| **9.2** | **Internal audit** | **N/A** | ❌ NOT DONE - CRITICAL |
| **9.3** | **Management review** | **N/A** | ❌ NOT DONE - CRITICAL |
| 10.1 | Nonconformity & corrective action | CC4.2 | 📋 Need formal process |
| 10.2 | Continual improvement | Implicit | 📋 Need explicit process |

### 3.2 Annex A Controls Not Covered by SOC 2

| Control | Title | Gap Description | Priority |
|---------|-------|-----------------|----------|
| A.5.5 | Contact with authorities | No documented emergency contacts | High |
| A.5.6 | Contact with special interest groups | No documented security community engagement | Low |
| A.5.10 | Acceptable use | No formal policy | High |
| A.5.13 | Information labelling | No classification headers on documents | Medium |
| A.5.32 | Intellectual property | License compliance not formalized | Low |
| A.5.34 | Privacy/PII | Privacy impact assessment not done | Low |
| A.6.1 | Screening | No background check process | Medium |
| A.6.2 | Employment terms | Template needs ISO clauses | Medium |
| A.6.4 | Disciplinary process | Not documented | Low |
| A.8.10 | Information deletion | Data retention/deletion not documented | Medium |
| A.8.30 | Outsourced development | No contractor security requirements | Low |

---

## 4. Gap List with Owners and Evidence Pointers

### 4.1 Critical Gaps (Must Fix Before Claiming "Audit Ready")

| Gap ID | Description | ISO Clause | Owner | Evidence Location | Status |
|--------|-------------|------------|-------|-------------------|--------|
| GAP-001 | No internal audit conducted | Clause 9.2 | Carlos Sanchez | `iso27001/2026/internal-audits/` | ❌ NOT DONE |
| GAP-002 | No management review conducted | Clause 9.3 | Carlos Sanchez | `iso27001/2026/management-reviews/` | ❌ NOT DONE |
| GAP-003 | No acceptable use policy | A.5.10 | Carlos Sanchez | `iso27001/policies/` | ✅ Created |
| GAP-004 | No formal corrective action process | Clause 10.1 | Carlos Sanchez | `iso27001/09-corrective-actions-log.csv` | 📋 TODO |

### 4.2 High Priority Gaps (Complete within 30 days)

| Gap ID | Description | ISO Clause | Owner | Evidence Location | Due Date |
|--------|-------------|------------|-------|-------------------|----------|
| GAP-005 | No documented authority contacts | A.5.5 | Carlos Sanchez | `iso27001/00-isms-scope-and-context.md` | 2026-02-01 |
| GAP-006 | Risk register needs CSV format | Clause 8.2 | Carlos Sanchez | `iso27001/03-risk-register.csv` | 2026-02-01 |
| GAP-007 | Risk treatment plan not formal | Clause 8.3 | Carlos Sanchez | `iso27001/04-risk-treatment-plan.csv` | 2026-02-01 |
| GAP-008 | SoA needs CSV format | Clause 6.1.3d | Carlos Sanchez | `iso27001/05-statement-of-applicability.csv` | 2026-02-01 |

### 4.3 Medium Priority Gaps (Complete within 90 days)

| Gap ID | Description | ISO Clause | Owner | Evidence Location | Due Date |
|--------|-------------|------------|-------|-------------------|----------|
| GAP-009 | No formal asset register | A.5.9 | Carlos Sanchez | `iso27001/asset-register.csv` | 2026-03-01 |
| GAP-010 | Document classification headers | A.5.13 | Carlos Sanchez | All documents | 2026-03-01 |
| GAP-011 | Background check process | A.6.1 | Carlos Sanchez | `soc2/procedures/` | 2026-03-01 |
| GAP-012 | Employment agreement template | A.6.2 | Carlos Sanchez | HR templates | 2026-03-01 |
| GAP-013 | Data deletion procedures | A.8.10 | Carlos Sanchez | Data retention policy | 2026-03-01 |

### 4.4 Low Priority Gaps (Complete within 180 days)

| Gap ID | Description | ISO Clause | Owner | Evidence Location | Due Date |
|--------|-------------|------------|-------|-------------------|----------|
| GAP-014 | Disciplinary process | A.6.4 | Carlos Sanchez | Employee handbook | 2026-06-01 |
| GAP-015 | Outsourced development security | A.8.30 | Carlos Sanchez | Contractor requirements | 2026-06-01 |
| GAP-016 | Privacy impact assessment | A.5.34 | Carlos Sanchez | Privacy documentation | 2026-06-01 |

---

## 5. Evidence Vault Alignment

### 5.1 Shared Evidence Structure

```
aegis-compliance-evidence/
├── soc2/                          # SOC 2 evidence (also serves ISO)
│   └── 2025/
│       ├── retroactive/           # Historical evidence → ISO A.8.32, A.8.25
│       └── 2025-01/               # Monthly evidence → Multiple controls
│           ├── access-reviews/    # → A.5.15-18
│           ├── ci-cd-security/    # → A.8.4, A.8.25, A.8.29
│           └── vuln-management/   # → A.8.8
└── iso27001/                      # ISO-specific evidence
    └── 2026/
        ├── internal-audits/       # Clause 9.2 (NO EVIDENCE YET)
        ├── management-reviews/    # Clause 9.3 (NO EVIDENCE YET)
        ├── risk-assessments/      # Clause 8.2
        └── corrective-actions/    # Clause 10.1
```

### 5.2 Evidence Reuse Summary

| Evidence Category | Source | Reusable for ISO? | Action Required |
|-------------------|--------|-------------------|-----------------|
| Access reviews | SOC 2 vault | ✅ Yes | None |
| Vulnerability reports | SOC 2 vault | ✅ Yes | None |
| Change records (PRs) | GitHub + vault | ✅ Yes | None |
| Training records | SOC 2 vault | ✅ Yes | None |
| Incident records | SOC 2 vault | ✅ Yes | None |
| Vendor assessments | SOC 2 vault | ✅ Yes | None |
| Policies | SOC 2 docs | ✅ Yes | Add ISO references |
| Internal audit | N/A | ❌ No | **MUST CREATE** |
| Management review | N/A | ❌ No | **MUST CREATE** |

---

## 6. Existing ISO 27001 Documents (Already Created)

The following documents were created in a previous session:

| Document | Status | Needs Update? |
|----------|--------|---------------|
| 00-isms-scope.md | ✅ Created | Rename + expand to 00-isms-scope-and-context.md |
| 01-information-security-policy.md | ✅ Created | ✅ Good |
| 02-statement-of-applicability.md | ✅ Created | Convert to CSV per user request |
| 03-risk-assessment-methodology.md | ✅ Created | Rename to 02-risk-methodology.md |
| 04-risk-register.md | ✅ Created | Convert to CSV per user request |
| 07-internal-audit-program.md | ✅ Created | ✅ Good (program exists, audit NOT conducted) |
| 08-management-review-template.md | ✅ Created | ✅ Good (template exists, review NOT conducted) |
| 09-soc2-iso27001-mapping.md | ✅ Created | Keep as reference |
| policies/acceptable-use-policy.md | ✅ Created | ✅ Good |
| STATUS.md | ✅ Created | Update to reflect actual status |

---

## 7. Phase 1 Deliverables Checklist

Per user instructions, the following files are required:

| File | Status | Format |
|------|--------|--------|
| 00-isms-scope-and-context.md | 🔄 Adapt from existing | Markdown |
| 01-leadership-policy-roles.md | 📋 Create | Markdown |
| 02-risk-methodology.md | 🔄 Rename existing 03-* | Markdown |
| 03-risk-register.csv | 📋 Convert from MD | CSV |
| 04-risk-treatment-plan.csv | 📋 Create | CSV |
| 05-statement-of-applicability.csv | 📋 Convert from MD | CSV |
| 07-internal-audit-program.md | ✅ Exists | Markdown |
| internal-audit-report-template.md | 📋 Create | Markdown |
| 08-management-review-template.md | ✅ Exists | Markdown |
| 09-corrective-actions-log.csv | 📋 Create | CSV |
| STATUS.md | 🔄 Update | Markdown |

---

## 8. Key Warnings

### ⚠️ DO NOT CLAIM "CERTIFICATION READY" OR "AUDIT READY"

The following critical requirements have NOT been met:

1. **Clause 9.2 Internal Audit**: No internal audit has been conducted
2. **Clause 9.3 Management Review**: No management review has been conducted
3. **Clause 10.1 Corrective Actions**: No evidence of corrective action process operating

**Current Status:** ISMS documentation created, but ISMS NOT YET OPERATING

### ⚠️ Model B Only

- Scope limited to SDLC and corporate systems
- Customers deploy and operate their own environments
- Availability TSC is OUT OF SCOPE

### ⚠️ Solo Founder Limitations

- Segregation of duties has compensating controls only
- Internal audit requires EXTERNAL auditor
- Management review is self-review (must be documented formally)

---

## 9. Recommended Next Steps

### Immediate (This Session)

1. ✅ Complete this REUSE_PLAN.md
2. 📋 Create Phase 1 files with correct formats (CSV where specified)
3. 📋 Create soc2-crosswalk.csv (Phase 2)
4. 📋 Update STATUS.md to reflect honest state

### Before Claiming "Audit Ready"

1. ❌ Conduct internal audit (external auditor required)
2. ❌ Conduct management review (founder self-review, documented)
3. ❌ Document at least one corrective action cycle

---

## 10. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | AI Agent + Carlos Sanchez | Initial REUSE_PLAN |
