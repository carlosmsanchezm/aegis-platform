# Aegis SOC 2 Evidence Vault

**Repository:** `aegis-compliance-evidence` (PRIVATE)
**Classification:** CONFIDENTIAL - Internal Use Only
**Last Updated:** 2025-01-17

---

## Purpose

This repository serves as the centralized evidence vault for Aegis Technologies' SOC 2 Type II audit. It contains timestamped exports, screenshots, and documentation that demonstrate control effectiveness over the audit period.

## Scope

- **Trust Services Criteria:** Security (CC series only)
- **Service Model:** Model B (self-hosted software vendor)
- **Audit Period:** [START_DATE] to [END_DATE]

---

## Folder Structure

```
aegis-compliance-evidence/
├── README.md                          # This file
├── .gitignore                         # Prevents secrets from being committed
├── EVIDENCE_INTAKE_CHECKLIST.md       # Checklist for adding new evidence
│
└── soc2/
    ├── 2025/                          # Year folder
    │   ├── governance/                # Annual governance documents
    │   │   ├── org-chart/
    │   │   ├── policies/
    │   │   └── risk-assessment/
    │   ├── vendor-management/         # Vendor SOC 2 reports, assessments
    │   ├── training-policy-ack/       # Training records, policy acknowledgments
    │   ├── Q1/                        # Quarterly evidence
    │   │   └── access-reviews/
    │   ├── Q2/
    │   ├── Q3/
    │   ├── Q4/
    │   ├── 2025-01/                   # Monthly evidence folders
    │   │   ├── access-reviews/
    │   │   ├── change-management/
    │   │   ├── ci-cd-security/
    │   │   ├── vuln-management/
    │   │   ├── logging-monitoring/
    │   │   ├── incident-response/
    │   │   └── communication/
    │   ├── 2025-02/
    │   │   └── ...
    │   └── ...
    │
    └── templates/                     # Reusable templates
        ├── access-review-template.md
        ├── incident-report-template.md
        └── vendor-assessment-template.md
```

---

## Naming Conventions

### Folder Names
- **Year:** `YYYY/` (e.g., `2025/`)
- **Month:** `YYYY-MM/` (e.g., `2025-01/`)
- **Quarter:** `QX/` (e.g., `Q1/`)
- **Control Area:** lowercase with hyphens (e.g., `access-reviews/`, `vuln-management/`)

### File Names

**Format:** `YYYY-MM-DD_<system>_<control>_<description>.<ext>`

**Examples:**
```
2025-01-17_github_CC8.1_branch-protection-settings.json
2025-01-17_aws_CC6.6_mfa-status-report.json
2025-01-17_github_CC7.1_dependabot-alerts-summary.json
2025-01-15_quarterly_CC6.4_access-review-certification.pdf
2025-01-20_meeting_CC1.2_security-review-notes.md
```

**Components:**
| Component | Description | Example |
|-----------|-------------|---------|
| Date | ISO date of evidence collection | `2025-01-17` |
| System | Source system | `github`, `aws`, `keycloak`, `manual` |
| Control | SOC 2 control ID | `CC6.1`, `CC8.1` |
| Description | Brief description | `branch-protection-settings` |
| Extension | File type | `.json`, `.pdf`, `.png`, `.md` |

---

## What Counts as Acceptable Evidence

### ✅ Good Evidence

| Type | Examples | Format |
|------|----------|--------|
| **Configuration Exports** | IAM policies, branch protection settings, RBAC configs | JSON, YAML |
| **System Reports** | Access lists, vulnerability summaries, audit logs | JSON, CSV |
| **Screenshots** | Dashboard views, setting confirmations | PNG (with timestamp visible) |
| **Meeting Notes** | Security review minutes, decisions | Markdown |
| **Certifications** | Manager sign-offs, training completions | PDF, signed markdown |
| **Vendor Reports** | SOC 2 reports from vendors | PDF |

### ❌ Not Acceptable / Excluded

| Type | Reason | What to Do Instead |
|------|--------|-------------------|
| **Credentials/Secrets** | Security risk | Never commit; reference exists only |
| **API Keys/Tokens** | Security risk | Redact or exclude entirely |
| **Customer Data** | Privacy/confidentiality | Aggregate or anonymize |
| **PII** | Privacy regulations | Redact names/emails where possible |
| **Penetration Test Details** | Security risk | Store summary only, full report offline |
| **Exploit Code** | Security risk | Reference CVE only |

---

## 🚨 NO SECRETS POLICY 🚨

**This repository must NEVER contain:**

- ❌ Passwords or passphrases
- ❌ API keys or tokens
- ❌ Private keys or certificates
- ❌ AWS access keys or secret keys
- ❌ Database connection strings
- ❌ OAuth client secrets
- ❌ Any credential that could grant system access

**Before every commit:**
1. Run `git diff --staged` and review for secrets
2. Check JSON files for `password`, `secret`, `key`, `token` fields
3. Use `git secrets` or similar pre-commit hooks

**If you accidentally commit a secret:**
1. Rotate the credential IMMEDIATELY
2. Use `git filter-branch` or BFG to remove from history
3. Force push (coordinate with team)
4. Document the incident

---

## Evidence Collection Schedule

| Frequency | Evidence Type | Due By |
|-----------|--------------|--------|
| **Weekly** | Vulnerability scan summaries | Friday |
| **Monthly** | Access exports, CI/CD reports, monitoring snapshots | 5th of next month |
| **Quarterly** | Access review certifications, risk register updates | 15th of quarter end month |
| **Annually** | Full risk assessment, policy reviews, DR test results | End of Q4 |

---

## Adding New Evidence

1. Review `EVIDENCE_INTAKE_CHECKLIST.md`
2. Place file in correct folder (`soc2/YYYY/YYYY-MM/<control-area>/`)
3. Use correct naming convention
4. Verify no secrets are included
5. Commit with descriptive message: `SOC2 evidence: <control> - <description>`
6. Update the main repo's `STATUS.md` if needed

---

## Access Control

- **Repository visibility:** PRIVATE
- **Access:** Aegis personnel only (Security Lead, CEO)
- **MFA required:** Yes
- **Branch protection:** Required reviews for main branch

---

## Related Documentation

- Main repo: `aegis-platform/docs/compliance/soc2/`
- Control matrix: `aegis-platform/docs/compliance/soc2/control-matrix.csv`
- Evidence map: `aegis-platform/docs/compliance/soc2/evidence-map.md`

---

## Contact

For questions about evidence collection: Carlos Sanchez (CEO/Security Lead)
