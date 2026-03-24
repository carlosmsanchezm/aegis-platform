# Monthly SOC 2 Evidence Collection Checklist

**Run on:** 1st of each month (or closest business day)
**Time required:** ~5 minutes
**Owner:** Carlos Sanchez

---

## Quick Commands (Copy/Paste)

```bash
# Set the month (adjust as needed)
MONTH=$(date +%Y-%m)
YEAR=$(date +%Y)

# Navigate to platform repo
cd ~/code/aegis-platform  # Adjust path as needed

# 1. Export GitHub security baseline (ALL 3 REPOS)
GITHUB_OWNER=carlosmsanchezm GITHUB_REPOS="aegis-platform aegis-ui sovran" \
    ./scripts/compliance/export_github_security_baseline.sh \
    ../aegis-compliance-evidence/soc2/${YEAR}/${MONTH}/ci-cd-security/

# 2. Export AWS security baseline
AWS_PROFILE=aegis-new \
    ./scripts/compliance/export_aws_security_baseline.sh \
    ../aegis-compliance-evidence/soc2/${YEAR}/${MONTH}/access-reviews/

# 3. Commit to evidence vault
cd ../aegis-compliance-evidence
git add -A
git status  # Review - NO SECRETS!
git commit -m "SOC2 evidence: ${MONTH} monthly collection"
git push
```

---

## Monthly Checklist

### Evidence Collection
- [ ] GitHub security baseline exported
- [ ] AWS security baseline exported
- [ ] Evidence committed to vault
- [ ] No secrets in committed files (verified via git status)

### Security Review (All 3 Repos)
- [ ] Dependabot alerts reviewed for aegis-platform
- [ ] Dependabot alerts reviewed for aegis-ui
- [ ] Dependabot alerts reviewed for sovran
  - SLA: High/Critical: 7 days | Medium: 30 days | Low: 90 days
- [ ] Any new security advisories reviewed
- [ ] Failed CI runs reviewed (any security-related failures?)

**Quick vuln check:**
```bash
for repo in aegis-platform aegis-ui sovran; do
    echo "=== $repo ==="
    gh api repos/carlosmsanchezm/$repo/dependabot/alerts \
        --jq '[.[] | select(.state == "open")] | length' 2>/dev/null || echo "0"
done
```

### Quick Health Check
- [ ] MFA still enabled on all accounts
- [ ] No unexpected users/collaborators added
- [ ] CloudTrail logging active
- [ ] Branch protection still enforced

---

## Quarterly Tasks (In Addition to Monthly)

**Q1 (January-March) → Due April 15**
**Q2 (April-June) → Due July 15**
**Q3 (July-September) → Due October 15**
**Q4 (October-December) → Due January 15**

- [ ] Run quarterly access review (`procedures/quarterly-access-review.md`)
- [ ] Document access review in evidence vault
- [ ] Review and update vendor inventory
- [ ] Check for policy updates needed

---

## Annual Tasks (January)

- [ ] Annual policy review (all 6 policies)
- [ ] Update policy versions if changes made
- [ ] Annual risk assessment
- [ ] Review and update control matrix
- [ ] Archive previous year's evidence

---

## Evidence Naming Convention

```
YYYY-MM-DD_<system>_<control>_<description>.<ext>

Examples:
2026-02-01_github_CC8.1_branch-protection.json
2026-02-01_aws_CC6.6_iam-mfa-status.json
```

---

## Troubleshooting

### Script fails with authentication error
```bash
# Re-authenticate GitHub
gh auth login

# Re-authenticate AWS
aws sso login --profile aegis-new
```

### Evidence vault push fails
```bash
cd ../aegis-compliance-evidence
git pull --rebase
git push
```

### Forgot to collect last month
- Run the collection commands with explicit date:
```bash
MONTH=2026-01  # Set to missed month
# Then run normal collection commands
```
- Add note in commit message: "Late collection for ${MONTH}"

---

## Calendar Reminders

Set these recurring reminders:

| Reminder | Frequency | Day |
|----------|-----------|-----|
| Monthly evidence collection | Monthly | 1st |
| Q1 access review | Annual | April 1 |
| Q2 access review | Annual | July 1 |
| Q3 access review | Annual | October 1 |
| Q4 access review | Annual | January 1 |
| Annual policy review | Annual | January 15 |
