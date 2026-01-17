# Evidence Intake Checklist

Use this checklist before adding any evidence to the vault.

---

## Pre-Commit Checklist

### 1. File Placement
- [ ] File is in correct year folder: `soc2/YYYY/`
- [ ] File is in correct month folder: `soc2/YYYY/YYYY-MM/`
- [ ] File is in correct control area subfolder (e.g., `access-reviews/`, `vuln-management/`)

### 2. File Naming
- [ ] File name follows convention: `YYYY-MM-DD_<system>_<control>_<description>.<ext>`
- [ ] Date is accurate (collection date, not upload date)
- [ ] System identifier is correct (github, aws, keycloak, manual, etc.)
- [ ] Control ID is included (CC6.1, CC8.1, etc.)
- [ ] Description is clear and concise

### 3. Content Review
- [ ] File contains actual evidence (not just placeholders)
- [ ] Evidence is complete (not truncated or partial)
- [ ] Timestamps are visible/included in the evidence
- [ ] Evidence clearly demonstrates the control

### 4. 🚨 SECRET SCAN 🚨
- [ ] **NO passwords** in the file
- [ ] **NO API keys** in the file
- [ ] **NO tokens** in the file
- [ ] **NO private keys** in the file
- [ ] **NO connection strings** in the file
- [ ] Searched file for: `password`, `secret`, `key`, `token`, `credential`
- [ ] JSON files reviewed for sensitive fields

### 5. Privacy Check
- [ ] No customer PII (names, emails, etc.) unless necessary
- [ ] No customer-specific data that could identify them
- [ ] Aggregated/anonymized where possible

### 6. Commit Message
- [ ] Message follows format: `SOC2 evidence: <control> - <description>`
- [ ] Message describes what evidence was added

---

## Quick Commands

```bash
# Check what you're about to commit
git diff --staged

# Search for potential secrets in staged files
git diff --staged | grep -iE "(password|secret|key|token|credential)"

# Commit with proper message
git commit -m "SOC2 evidence: CC8.1 - GitHub branch protection settings Jan 2025"

# If you find a secret - unstage the file
git reset HEAD <file>
```

---

## Monthly Evidence Collection

Run these scripts on the first week of each month:

```bash
# From the main aegis-platform repo
cd /path/to/aegis-platform

# Export GitHub security baseline
./scripts/compliance/export_github_security_baseline.sh /path/to/aegis-compliance-evidence/soc2/2025/2025-01/ci-cd-security

# Export AWS security baseline
AWS_PROFILE=myclaude ./scripts/compliance/export_aws_security_baseline.sh /path/to/aegis-compliance-evidence/soc2/2025/2025-01/access-reviews

# Export CI reports (SBOM, vuln scans)
./scripts/compliance/export_ci_reports.sh /path/to/aegis-compliance-evidence/soc2/2025/2025-01/vuln-management
```

---

## Evidence by Control Area

| Control Area | What to Collect | Frequency |
|--------------|----------------|-----------|
| `access-reviews/` | User lists, MFA status, IAM exports, quarterly certs | Monthly + Quarterly |
| `change-management/` | PR exports, release tags, branch protection | Monthly |
| `ci-cd-security/` | GitHub security settings, workflow configs | Monthly |
| `vuln-management/` | Dependabot summaries, scan reports, remediation tickets | Weekly/Monthly |
| `logging-monitoring/` | Alert configs, dashboard screenshots | Monthly |
| `incident-response/` | IR tickets, alert triage records | As needed |
| `communication/` | Security announcements, customer notifications | As needed |
| `governance/` | Policies, org charts, meeting notes | Quarterly/Annual |
| `vendor-management/` | Vendor SOC 2 reports, assessments | Annual |
| `training-policy-ack/` | Training completions, policy sign-offs | Annual |
