# ISMS Operating Plan for Solo Founder

**Document ID:** ISMS-OPS-001
**Version:** 1.0
**Effective Date:** 2026-01-17
**Owner:** Carlos Sanchez
**Classification:** Internal

---

## 1. Purpose

This document provides a practical operating schedule for maintaining the ISO 27001 ISMS as a solo founder. It defines exactly what to do, when to do it, and where to store evidence.

---

## 2. Operating Schedule Summary

| Frequency | Time Required | Key Activities |
|-----------|---------------|----------------|
| Weekly | 30 min | Dependabot review, security alerts |
| Monthly | 2-3 hours | Evidence exports, access review prep, incident check |
| Quarterly | 4-6 hours | Access review, risk register review, policy check |
| Semi-annual | 8-12 hours | Internal audit, management review |
| Annual | 4-6 hours | Full policy review, vendor assessments, certification prep |

**Total Annual Effort:** ~80-100 hours (~2 hours/week average)

---

## 3. Weekly Tasks (30 minutes)

### Week Checklist

| Task | Steps | Evidence Location |
|------|-------|-------------------|
| **Review Dependabot alerts** | 1. Check each repo for alerts 2. Triage: critical/high = fix this week 3. Document any deferrals | GitHub Security tab |
| **Check GitHub security alerts** | 1. Review Code scanning alerts 2. Review Secret scanning alerts | GitHub Security tab |
| **Monitor for incidents** | 1. Check email for security notifications 2. Review AWS CloudTrail for anomalies | CloudTrail console |

### Commands to Run

```bash
# Quick Dependabot status check (run locally)
for repo in aegis-platform aegis-ui sovran; do
  echo "=== $repo ==="
  gh api repos/carlosmsanchezm/$repo/dependabot/alerts --jq '[.[] | select(.state=="open")] | length'
done
```

---

## 4. Monthly Tasks (2-3 hours)

### First Week of Each Month

| # | Task | Steps | Evidence Location | Time |
|---|------|-------|-------------------|------|
| 1 | **Export GitHub security baseline** | Run `scripts/compliance/export_github_security_baseline.sh` | `aegis-compliance-evidence/soc2/YYYY/YYYY-MM/access-reviews/` | 15 min |
| 2 | **Export AWS security baseline** | Run `scripts/compliance/export_aws_security_baseline.sh` | `aegis-compliance-evidence/soc2/YYYY/YYYY-MM/access-reviews/` | 15 min |
| 3 | **Export CI/CD reports** | Run `scripts/compliance/export_ci_reports.sh` | `aegis-compliance-evidence/soc2/YYYY/YYYY-MM/ci-cd-security/` | 15 min |
| 4 | **Check for security incidents** | Review logs for month, document any incidents or "no incidents" | `aegis-compliance-evidence/soc2/YYYY/YYYY-MM/incident-response/` | 30 min |
| 5 | **Verify Dependabot remediation** | Confirm all critical/high vulns addressed | `aegis-compliance-evidence/soc2/YYYY/YYYY-MM/vuln-management/` | 15 min |
| 6 | **Review vendor security advisories** | Check AWS, GitHub security bulletins | Notes in evidence folder | 15 min |
| 7 | **Update compliance metrics** | Log: vulns open, access review status, incidents | Monthly metrics log | 15 min |

### Monthly Evidence Export Script

```bash
#!/bin/bash
# monthly-evidence-export.sh
# Run on 1st of each month

YEAR=$(date +%Y)
MONTH=$(date +%Y-%m)
VAULT_BASE="$HOME/aegis-compliance-evidence"

# Create month directories
mkdir -p "$VAULT_BASE/soc2/$YEAR/$MONTH"/{access-reviews,ci-cd-security,vuln-management,incident-response,logging-monitoring}

# Export evidence
cd /path/to/aegis-platform-observability-integration

echo "=== GitHub Security Baseline ==="
./scripts/compliance/export_github_security_baseline.sh "$VAULT_BASE/soc2/$YEAR/$MONTH/access-reviews"

echo "=== AWS Security Baseline ==="
./scripts/compliance/export_aws_security_baseline.sh "$VAULT_BASE/soc2/$YEAR/$MONTH/access-reviews"

echo "=== CI/CD Reports ==="
./scripts/compliance/export_ci_reports.sh "$VAULT_BASE/soc2/$YEAR/$MONTH/ci-cd-security"

echo "=== Monthly export complete ==="
ls -la "$VAULT_BASE/soc2/$YEAR/$MONTH/"
```

### Monthly Incident Log Entry

Create `aegis-compliance-evidence/soc2/YYYY/YYYY-MM/incident-response/YYYY-MM-incident-log.md`:

```markdown
# Incident Log - [YYYY-MM]

## Summary
- **Security Incidents:** [0 | N - describe]
- **Near Misses:** [0 | N - describe]
- **Dependabot Alerts Resolved:** [N]
- **Access Changes:** [N new, N removed, N modified]

## Incidents
[None this month | Incident details...]

## Observations
[Any security-relevant observations or concerns]

## Prepared By
Carlos Sanchez - [Date]
```

---

## 5. Quarterly Tasks (4-6 hours)

### Quarterly Schedule

| Quarter | Months | Access Review Due | Risk Review Due |
|---------|--------|-------------------|-----------------|
| Q1 | Jan-Mar | March 31 | March 31 |
| Q2 | Apr-Jun | June 30 | June 30 |
| Q3 | Jul-Sep | September 30 | September 30 |
| Q4 | Oct-Dec | December 31 | December 31 |

### Q1/Q2/Q3/Q4 Checklist

| # | Task | Steps | Evidence Location | Time |
|---|------|-------|-------------------|------|
| 1 | **Quarterly Access Review** | Follow `soc2/procedures/quarterly-access-review.md` | `aegis-compliance-evidence/soc2/YYYY/QX/access-reviews/` | 2 hrs |
| 2 | **Risk Register Review** | Review all 17 risks, update status, add new risks | `iso27001/03-risk-register.csv` + meeting notes | 1 hr |
| 3 | **Policy Quick Check** | Confirm policies still accurate, note any needed updates | Checklist (see below) | 30 min |
| 4 | **Vendor Review** | Check vendor SOC 2 report expiry, any security incidents | Vendor tracking spreadsheet | 30 min |
| 5 | **Training Verification** | Confirm security awareness current | Training records | 15 min |
| 6 | **Corrective Action Status** | Review open CAs, close completed ones | `iso27001/09-corrective-actions-log.csv` | 30 min |

### Quarterly Access Review Procedure

```markdown
## Access Review Steps

1. Export current access lists:
   - GitHub: `gh api orgs/{org}/members` or repo collaborators
   - AWS: `aws iam list-users` + `aws iam list-roles`

2. For each user/service account, verify:
   - [ ] Still needed (active project/role)
   - [ ] Permissions appropriate (least privilege)
   - [ ] MFA enabled
   - [ ] Last activity reasonable

3. Document findings in:
   `aegis-compliance-evidence/soc2/YYYY/QX/access-reviews/YYYY-QX-access-review.md`

4. For solo founder:
   - Self-certification is acceptable
   - Document: "Access reviewed by Carlos Sanchez on [date]"
   - Sign with commit
```

### Policy Quick Check Template

```markdown
# Policy Quick Check - Q[X] [YYYY]

| Policy | Current | Still Accurate? | Changes Needed? |
|--------|---------|-----------------|-----------------|
| ISP-001 Information Security | v1.0 | ☐ Yes ☐ No | |
| ACP-001 Access Control | v1.0 | ☐ Yes ☐ No | |
| CMP-001 Change Management | v1.0 | ☐ Yes ☐ No | |
| IRP-001 Incident Response | v1.0 | ☐ Yes ☐ No | |
| RMP-001 Risk Management | v1.0 | ☐ Yes ☐ No | |
| VMP-001 Vendor Management | v1.0 | ☐ Yes ☐ No | |
| AUP-001 Acceptable Use | v1.0 | ☐ Yes ☐ No | |

**Reviewed By:** Carlos Sanchez
**Date:** [YYYY-MM-DD]
**Next Full Review:** [Annual date]
```

### Risk Register Review Steps

1. Open `docs/compliance/iso27001/03-risk-register.csv`
2. For each risk:
   - Is status still accurate?
   - Has likelihood/impact changed?
   - Are treatments on track?
   - Any new controls implemented?
3. Consider new risks from:
   - Changes in technology/architecture
   - Security incidents or near-misses
   - Threat intelligence
   - Audit findings
4. Update CSV and commit with message: `chore(compliance): Q[X] risk register review`
5. Store meeting notes in `aegis-compliance-evidence/iso27001/YYYY/risk-assessments/YYYY-QX-risk-review.md`

---

## 6. Semi-Annual Tasks (8-12 hours)

### Schedule

| Period | Months | Internal Audit | Management Review |
|--------|--------|----------------|-------------------|
| H1 | Jan-Jun | February | March |
| H2 | Jul-Dec | August | September |

### 6.1 Internal Audit (6-8 hours)

**For solo founder: Use external auditor for independence**

#### Pre-Audit Preparation (2 hours)

1. Update `docs/compliance/iso27001/11-internal-audit-plan.md` with current scope
2. Gather evidence index for auditor
3. Confirm availability with auditor
4. Prepare audit checklist

#### During Audit (3-4 hours)

1. Opening meeting (15 min)
2. Documentation review (1-2 hours)
3. Evidence sampling (1-2 hours)
4. Closing meeting (30 min)

#### Post-Audit (1-2 hours)

1. Review draft report from auditor
2. Agree on findings and classifications
3. Enter findings in `09-corrective-actions-log.csv`
4. Store report in `aegis-compliance-evidence/iso27001/YYYY/internal-audits/`

### 6.2 Management Review (2-4 hours)

**For solo founder: Self-review with formal documentation**

#### Preparation (1 hour)

Gather inputs for management review (per Clause 9.3):

| Input | Source |
|-------|--------|
| Status of previous review actions | Previous management review |
| Changes affecting ISMS | Risk register, incident log |
| Nonconformity/corrective action status | 09-corrective-actions-log.csv |
| Monitoring/measurement results | Monthly metrics |
| Audit results | Internal audit report |
| Fulfillment of objectives | Objective tracking |
| Feedback from interested parties | Customer feedback, if any |
| Risk assessment results | 03-risk-register.csv |
| Opportunities for improvement | Accumulated suggestions |

#### Review (1-2 hours)

1. Complete `08-management-review-template.md`
2. Document decisions on:
   - Improvement opportunities
   - Changes to ISMS
   - Resource needs
3. Sign and date

#### Post-Review (30 min)

1. Store completed review in `aegis-compliance-evidence/iso27001/YYYY/management-reviews/`
2. Update STATUS.md
3. Create action items in tracking system
4. Commit changes

---

## 7. Annual Tasks (4-6 hours)

### Annual Schedule

| Task | Due | Time |
|------|-----|------|
| Full policy review | January | 2 hrs |
| Vendor security assessment | January | 1 hr |
| Training renewal | January | 1 hr |
| ISMS scope review | December (for next year) | 1 hr |
| Statement of Applicability review | December | 1 hr |

### Annual Policy Review Checklist

For each policy:

1. [ ] Review entire document
2. [ ] Verify alignment with current practices
3. [ ] Update any outdated references
4. [ ] Confirm regulatory changes addressed
5. [ ] Update version and approval date
6. [ ] Communicate changes to stakeholders

### Vendor Security Assessment

```markdown
# Annual Vendor Review - [YYYY]

| Vendor | Service | SOC 2 Report Date | Expiry | Issues | Status |
|--------|---------|-------------------|--------|--------|--------|
| AWS | Cloud infrastructure | [Date] | [Date] | None | ✅ |
| GitHub | Source control/CI | [Date] | [Date] | None | ✅ |
| Google (planned) | Identity | [Date] | [Date] | N/A | 📋 |

**Action Items:**
1. [List any needed actions]

**Reviewed By:** Carlos Sanchez
**Date:** [YYYY-MM-DD]
```

---

## 8. Evidence Storage Structure

```
aegis-compliance-evidence/
├── soc2/
│   ├── 2025/
│   │   ├── retroactive/           # Historical Sept-Dec 2025
│   │   ├── 2025-09/ through 2025-12/  # Monthly evidence
│   │   └── Q4/                    # Q4 2025 access review
│   └── 2026/
│       ├── 2026-01/               # January 2026
│       │   ├── access-reviews/
│       │   │   ├── github_security_baseline_YYYY-MM-DD.json
│       │   │   ├── aws_security_baseline_YYYY-MM-DD.json
│       │   │   └── mfa_status_YYYY-MM-DD.json
│       │   ├── ci-cd-security/
│       │   │   ├── github_actions_runs_YYYY-MM-DD.json
│       │   │   └── branch_protection_YYYY-MM-DD.json
│       │   ├── vuln-management/
│       │   │   └── dependabot_status_YYYY-MM-DD.json
│       │   ├── incident-response/
│       │   │   └── 2026-01-incident-log.md
│       │   └── logging-monitoring/
│       │       └── cloudtrail_status_YYYY-MM-DD.json
│       ├── 2026-02/ ... 2026-12/
│       ├── Q1/
│       │   └── access-reviews/
│       │       └── 2026-Q1-access-review.md
│       ├── Q2/, Q3/, Q4/
│       ├── governance/
│       │   ├── policy-review-2026.md
│       │   └── training-records/
│       ├── risk-management/
│       │   └── annual-risk-assessment-2026.md
│       └── vendor-management/
│           ├── aws-soc2-2026.pdf
│           └── github-soc2-2026.pdf
└── iso27001/
    └── 2026/
        ├── internal-audits/
        │   ├── 2026-02-internal-audit-report.md
        │   └── 2026-08-internal-audit-report.md
        ├── management-reviews/
        │   ├── 2026-03-management-review.md
        │   └── 2026-09-management-review.md
        ├── risk-assessments/
        │   ├── 2026-Q1-risk-review.md
        │   └── ...
        └── corrective-actions/
            └── [CA evidence files]
```

---

## 9. Quick Reference Card

### Weekly (30 min - Fridays)
- [ ] Review Dependabot alerts
- [ ] Check security notifications
- [ ] Note any concerns

### Monthly (2 hrs - 1st week)
- [ ] Run evidence export scripts
- [ ] Create incident log entry
- [ ] Update metrics

### Quarterly (4 hrs - last week of quarter)
- [ ] Access review
- [ ] Risk register review
- [ ] Policy quick check
- [ ] Vendor check

### Semi-Annual (10 hrs)
- [ ] Internal audit (external auditor)
- [ ] Management review

### Annual (4 hrs - January)
- [ ] Full policy review
- [ ] Vendor assessments
- [ ] Training renewal

---

## 10. Calendar Integration

Add these recurring calendar events:

| Event | Frequency | Duration | When |
|-------|-----------|----------|------|
| Security Review | Weekly | 30 min | Friday 3pm |
| Monthly Evidence Export | Monthly | 2 hrs | 1st Monday |
| Quarterly Access Review | Quarterly | 4 hrs | Last Friday of Mar/Jun/Sep/Dec |
| Internal Audit Prep | Semi-annual | 2 hrs | 1st week of Feb/Aug |
| Management Review | Semi-annual | 2 hrs | 2nd week of Mar/Sep |
| Annual Policy Review | Annual | 4 hrs | 2nd week of January |

---

## 11. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial ISMS operating plan |
