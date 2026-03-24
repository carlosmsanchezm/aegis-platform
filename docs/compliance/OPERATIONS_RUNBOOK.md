# Compliance Operations Runbook

**Version:** 2.0
**Last Updated:** 2026-03-23
**Owner:** Carlos Sanchez
**Applies To:** SOC 2 Type II + ISO 27001 + CMMC + FedRAMP

---

## For AI Assistants (Claude)

**Instructions file:** [`CLAUDE_INSTRUCTIONS.md`](./CLAUDE_INSTRUCTIONS.md)

When asking Claude to run compliance tasks, use these prompts:

| Task | Prompt |
|------|--------|
| Weekly | "Read docs/compliance/CLAUDE_INSTRUCTIONS.md and run the weekly compliance check" |
| Monthly | "Read docs/compliance/CLAUDE_INSTRUCTIONS.md and run the monthly evidence collection" |
| Quarterly | "Read docs/compliance/CLAUDE_INSTRUCTIONS.md and run the quarterly access review" |

---

## Overview

This runbook provides single-command operations for maintaining SOC 2 and ISO 27001 compliance as a solo founder. All tasks are designed to be run locally and produce evidence that satisfies both frameworks.

### Quick Reference

| Task | Frequency | Command | Time |
|------|-----------|---------|------|
| Weekly Security Check | Every Monday | `./scripts/compliance/run_weekly_checks.sh` | 30 min |
| Monthly Evidence | 1st of month | `./scripts/compliance/run_monthly_evidence.sh` | 2-3 hrs |
| Quarterly Access Review | Q1/Q2/Q3/Q4 | Manual procedure | 4-6 hrs |
| Semi-Annual Internal Audit | Every 6 months | External auditor | 8-12 hrs |
| Management Review | After internal audit | Manual procedure | 2-4 hrs |

---

## Prerequisites

### Environment Setup

Add to your shell profile (`~/.zshrc` or `~/.bashrc`):

```bash
# Compliance evidence vault location
export EVIDENCE_VAULT="$HOME/code/aegis-compliance-evidence"

# Optional: AWS profile for compliance exports
export AWS_PROFILE="aegis-new"

# Optional: GitHub owner/repos (defaults provided in scripts)
export GITHUB_OWNER="carlosmsanchezm"
export GITHUB_REPOS="aegis-platform,aegis-ui,sovran"
```

Reload shell: `source ~/.zshrc`

### Required Tools

```bash
# Core tools
brew install gh jq awscli

# Authenticate
gh auth login
aws sso login --profile $AWS_PROFILE

# Optional (for SBOM/vulnerability scans)
brew install syft grype
```

### First-Time Setup

```bash
# 1. Create evidence vault (if not exists)
mkdir -p "$EVIDENCE_VAULT"
cd "$EVIDENCE_VAULT"
git init
echo "# Aegis Compliance Evidence Vault" > README.md
git add . && git commit -m "Initialize evidence vault"

# 2. Run first monthly evidence collection
cd ~/code/aegis-platform
./scripts/compliance/run_monthly_evidence.sh
```

---

## Weekly Tasks (1-2 hrs)

### When: Every Monday

### What: Security Review + Remediation

1. Run automated security check script
2. Review Dependabot alerts across all repositories
3. **Remediate all critical and high findings** — upgrade deps, apply patches across all 3 repos (aegis-platform, aegis-ui, sovran)
4. Write remediation evidence to vault
5. Update ALL framework status docs (SOC 2, ISO, CMMC, FedRAMP) — not just the master status
6. Document any findings that could not be auto-fixed

### Command

```bash
cd ~/code/aegis-platform
./scripts/compliance/run_weekly_checks.sh
```

### Options

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview actions without making changes |
| `--push` | Commit and push to remote evidence repo |

### Evidence Output

```
$EVIDENCE_VAULT/
└── soc2/
    └── 2026/
        └── 2026-01/
            └── ci-cd-security/
                └── weekly/
                    ├── 2026-01-20_security_alerts.json
                    ├── 2026-01-20_weekly_review.md
                    └── RUN_LOG_2026-01-20.txt
```

### Controls Covered

| Framework | Control | Evidence |
|-----------|---------|----------|
| SOC 2 | CC7.1 | Vulnerability alert exports |
| SOC 2 | CC8.1 | Security review record |
| ISO 27001 | A.8.8 | Technical vulnerability management |
| ISO 27001 | A.8.32 | Change management evidence |

### Troubleshooting

| Issue | Solution |
|-------|----------|
| "EVIDENCE_VAULT not set" | Run `export EVIDENCE_VAULT=$HOME/code/aegis-compliance-evidence` |
| "gh CLI not authenticated" | Run `gh auth login` |
| "Permission denied" | Run `chmod +x scripts/compliance/run_weekly_checks.sh` |

---

## Monthly Tasks (2-3 hours)

### When: 1st of each month

### What: Full Evidence Collection

1. Export GitHub security baseline
2. Export AWS security baseline
3. Export CI/CD reports (SBOM + vulnerability scans)
4. Create/update incident log
5. Generate monthly summary

### Command

```bash
cd ~/code/aegis-platform
./scripts/compliance/run_monthly_evidence.sh
```

### Options

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview actions without making changes |
| `--push` | Commit and push to remote evidence repo |
| `--skip-aws` | Skip AWS export (if not authenticated) |
| `--skip-ci` | Skip CI reports (if tools not installed) |

### Evidence Output

```
$EVIDENCE_VAULT/
└── soc2/
    └── 2026/
        └── 2026-01/
            ├── RUN_LOG_2026-01-01.txt
            ├── 2026-01-01_monthly_summary.md
            ├── access-reviews/
            │   ├── 2026-01-01_iam_users.json
            │   ├── 2026-01-01_iam_mfa_status.json
            │   ├── 2026-01-01_iam_roles.json
            │   ├── 2026-01-01_access_key_age.json
            │   ├── 2026-01-01_cloudtrail_config.json
            │   └── 2026-01-01_aws_security_summary.json
            ├── ci-cd-security/
            │   ├── aegis-platform/
            │   │   ├── 2026-01-01_repo_settings.json
            │   │   ├── 2026-01-01_branch_protection_main.json
            │   │   ├── 2026-01-01_collaborators.json
            │   │   └── 2026-01-01_dependabot_summary.json
            │   ├── aegis-ui/
            │   │   └── ...
            │   ├── sovran/
            │   │   └── ...
            │   └── 2026-01-01_security_baseline_summary.json
            ├── vuln-management/
            │   ├── sbom/
            │   │   └── 2026-01-01_platform-api_sbom.json
            │   ├── vuln-scans/
            │   │   └── 2026-01-01_platform-api_vulns.json
            │   └── 2026-01-01_ci_security_summary.json
            ├── incident-response/
            │   └── 2026-01_incident_log.md
            └── logging-monitoring/
                └── (CloudTrail exports)
```

### Controls Covered

| Framework | Control | Evidence |
|-----------|---------|----------|
| SOC 2 | CC6.1 | IAM user/role exports |
| SOC 2 | CC6.6 | MFA status exports |
| SOC 2 | CC6.7 | Collaborator/permission exports |
| SOC 2 | CC7.1 | Vulnerability scan results |
| SOC 2 | CC7.2 | CloudTrail/monitoring exports |
| SOC 2 | CC7.3 | Incident log |
| SOC 2 | CC8.1 | Branch protection, workflow exports |
| ISO 27001 | A.5.15-18 | Access control evidence |
| ISO 27001 | A.8.8 | SBOM + vulnerability scans |
| ISO 27001 | A.8.32 | Change management evidence |
| ISO 27001 | Clause 8.2 | Risk assessment inputs |

### Troubleshooting

| Issue | Solution |
|-------|----------|
| "AWS not authenticated" | Run `aws sso login --profile $AWS_PROFILE` or use `--skip-aws` |
| "syft/grype not found" | Run `brew install syft grype` or use `--skip-ci` |
| Export hangs | Check network; GitHub API may rate-limit |
| "No go.mod found" | Update GO_MODULES in export_ci_reports.sh |

---

## Quarterly Tasks (4-6 hours)

### When: January 15, April 15, July 15, October 15

### What: Access Review

1. Review all user access across systems
2. Verify appropriate permissions
3. Remove stale accounts
4. Document review results

### Procedure

```bash
# 1. Export current access state
./scripts/compliance/run_monthly_evidence.sh

# 2. Follow procedure document
open docs/compliance/soc2/procedures/quarterly-access-review.md

# 3. Save completed review to vault
cp completed_access_review.md "$EVIDENCE_VAULT/soc2/$(date +%Y)/$(date +%Y-%m)/access-reviews/"
```

### Evidence Output

```
$EVIDENCE_VAULT/
└── soc2/
    └── 2026/
        └── 2026-01/
            └── access-reviews/
                └── 2026-Q1_access_review.md
```

### Controls Covered

| Framework | Control | Evidence |
|-----------|---------|----------|
| SOC 2 | CC6.4 | Quarterly access review |
| ISO 27001 | A.5.18 | Access rights review |

---

## Semi-Annual Tasks (8-12 hours)

### When: February, August (or as scheduled)

### What: Internal Audit

**Important:** Solo founders must use an external auditor for independence per ISO 27001 Clause 9.2.

### Procedure

```bash
# 1. Prepare audit package
open docs/compliance/iso27001/11-internal-audit-plan.md
open docs/compliance/iso27001/12-internal-audit-checklist.md

# 2. Export evidence for auditor
./scripts/compliance/run_monthly_evidence.sh

# 3. Provide auditor with:
#    - Pre-filled audit report template
#    - Access to evidence vault
#    - Access to documentation

# 4. Save completed audit report
cp 2026-02-internal-audit-report.md "$EVIDENCE_VAULT/iso27001/$(date +%Y)/internal-audits/"
```

### Evidence Output

```
$EVIDENCE_VAULT/
└── iso27001/
    └── 2026/
        └── internal-audits/
            └── 2026-02-internal-audit-report.md
```

### Controls Covered

| Framework | Control | Evidence |
|-----------|---------|----------|
| ISO 27001 | Clause 9.2 | Internal audit report |
| SOC 2 | CC4.1 | Control monitoring |

---

## Management Review (After Internal Audit)

### When: After each internal audit

### What: Review ISMS Effectiveness

1. Review internal audit findings
2. Review risk register
3. Review corrective actions
4. Approve resources and improvements

### Procedure

```bash
# 1. Review template
open docs/compliance/iso27001/08-management-review-template.md

# 2. Complete review with decisions

# 3. Save to evidence vault
cp 2026-03-management-review.md "$EVIDENCE_VAULT/iso27001/$(date +%Y)/management-reviews/"
```

### Evidence Output

```
$EVIDENCE_VAULT/
└── iso27001/
    └── 2026/
        └── management-reviews/
            └── 2026-03-management-review.md
```

### Controls Covered

| Framework | Control | Evidence |
|-----------|---------|----------|
| ISO 27001 | Clause 9.3 | Management review minutes |
| SOC 2 | CC1.2 | Oversight activities |

---

## Corrective Actions

### When: As findings are identified

### What: Document and Track Remediation

### Procedure

```bash
# 1. Add entry to corrective actions log
open docs/compliance/iso27001/09-corrective-actions-log.csv

# 2. Create closure plan if needed
cp docs/compliance/iso27001/CA-001-closure-plan.md \
   docs/compliance/iso27001/CA-002-closure-plan.md

# 3. Track completion and save evidence
```

### Controls Covered

| Framework | Control | Evidence |
|-----------|---------|----------|
| ISO 27001 | Clause 10.1 | Corrective action log |
| SOC 2 | CC4.2 | Remediation activities |

---

## Quick Commands Reference

### First-Time Setup (CA-001 Closure)

```bash
# Set environment
export EVIDENCE_VAULT="$HOME/code/aegis-compliance-evidence"

# Verify auth
gh auth status
aws sts get-caller-identity --profile "${AWS_PROFILE:-default}"

# Create vault if needed
mkdir -p "$EVIDENCE_VAULT"

# Run first monthly evidence
cd ~/code/aegis-platform
./scripts/compliance/run_monthly_evidence.sh

# Verify
ls -la "$EVIDENCE_VAULT/soc2/$(date +%Y)/$(date +%Y-%m)/"
```

### Weekly Checks

```bash
cd ~/code/aegis-platform
./scripts/compliance/run_weekly_checks.sh

# With auto-push
./scripts/compliance/run_weekly_checks.sh --push

# Dry run first
./scripts/compliance/run_weekly_checks.sh --dry-run
```

### Monthly Evidence

```bash
cd ~/code/aegis-platform
./scripts/compliance/run_monthly_evidence.sh

# With auto-push
./scripts/compliance/run_monthly_evidence.sh --push

# Skip AWS if not needed
./scripts/compliance/run_monthly_evidence.sh --skip-aws

# Dry run first
./scripts/compliance/run_monthly_evidence.sh --dry-run
```

---

## Annual Calendar

| Month | Tasks |
|-------|-------|
| January | Weekly x4, Monthly, Q4 Access Review |
| February | Weekly x4, Monthly, Internal Audit |
| March | Weekly x4, Monthly, Management Review |
| April | Weekly x4, Monthly, Q1 Access Review |
| May | Weekly x4, Monthly |
| June | Weekly x4, Monthly |
| July | Weekly x4, Monthly, Q2 Access Review |
| August | Weekly x4, Monthly, Internal Audit |
| September | Weekly x4, Monthly, Management Review |
| October | Weekly x4, Monthly, Q3 Access Review |
| November | Weekly x4, Monthly |
| December | Weekly x4, Monthly, Annual Policy Review |

---

## Troubleshooting Guide

### Common Issues

#### EVIDENCE_VAULT Not Set

```bash
# Error
[ERROR] EVIDENCE_VAULT environment variable is not set.

# Fix
export EVIDENCE_VAULT="$HOME/code/aegis-compliance-evidence"

# Permanent fix - add to ~/.zshrc
echo 'export EVIDENCE_VAULT="$HOME/code/aegis-compliance-evidence"' >> ~/.zshrc
source ~/.zshrc
```

#### GitHub CLI Not Authenticated

```bash
# Error
[ERROR] GitHub CLI not authenticated

# Fix
gh auth login

# Verify
gh auth status
```

#### AWS CLI Not Authenticated

```bash
# Error
AWS: not authenticated (skipping AWS export)

# Fix for SSO
aws sso login --profile $AWS_PROFILE

# Fix for IAM credentials
aws configure --profile $AWS_PROFILE

# Verify
aws sts get-caller-identity --profile $AWS_PROFILE
```

#### Permission Denied on Scripts

```bash
# Error
permission denied: ./scripts/compliance/run_weekly_checks.sh

# Fix
chmod +x scripts/compliance/*.sh
```

#### GitHub API Rate Limiting

```bash
# Error
gh: API rate limit exceeded

# Fix - wait and retry, or use authenticated token
gh auth refresh
```

#### No go.mod Found (CI Reports)

```bash
# Warning
No go.mod at /path/to/module, skipping

# Fix - update GO_MODULES in export_ci_reports.sh
# Or skip CI reports
./scripts/compliance/run_monthly_evidence.sh --skip-ci
```

---

## Document References

| Document | Location |
|----------|----------|
| SOC 2 Status | `docs/compliance/soc2/STATUS.md` |
| ISO 27001 Status | `docs/compliance/iso27001/STATUS.md` |
| Evidence Map | `docs/compliance/soc2/evidence-map.md` |
| Risk Register | `docs/compliance/iso27001/03-risk-register.csv` |
| Statement of Applicability | `docs/compliance/iso27001/05-statement-of-applicability.csv` |
| Corrective Actions Log | `docs/compliance/iso27001/09-corrective-actions-log.csv` |
| ISMS Operating Plan | `docs/compliance/iso27001/10-isms-operating-plan.md` |
| Internal Audit Plan | `docs/compliance/iso27001/11-internal-audit-plan.md` |
| Internal Audit Checklist | `docs/compliance/iso27001/12-internal-audit-checklist.md` |
| Quarterly Access Review | `docs/compliance/soc2/procedures/quarterly-access-review.md` |

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 2.0 | 2026-03-23 | Claude Opus 4.6 | Added remediation to weekly tasks, fixed repo paths, expanded scope to all 4 frameworks, added cross-framework doc update requirements |
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial runbook |
