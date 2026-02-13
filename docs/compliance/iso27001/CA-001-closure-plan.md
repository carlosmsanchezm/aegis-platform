# Corrective Action Closure Plan: CA-001

**CA ID:** CA-001
**Title:** Evidence Vault Not Operational
**Opened:** 2026-01-17
**Due Date:** 2026-02-15
**Owner:** Carlos Sanchez
**ISO Clause:** Clause 7.5 (Documented Information), A.5.33 (Record Protection)

---

## 1. Problem Statement

The evidence vault directory structure and collection automation are documented but not yet executed. Evidence collection scripts exist in the repository but:

1. The vault directory (`aegis-compliance-evidence/`) has not been created locally
2. No monthly evidence exports have been run
3. Retroactive evidence (Sept-Dec 2025) has not been extracted
4. Audit readiness depends on having populated evidence

**Risk Reference:** RISK-015 (Risk Score: Medium - 9)

---

## 2. Root Cause Analysis

### Why did this happen?

| # | Why | Finding |
|---|-----|---------|
| 1 | Why wasn't vault created? | Focus on documentation, not execution |
| 2 | Why focus on documentation? | Building ISMS from scratch, needed docs first |
| 3 | Why docs first? | Cannot operate what isn't defined |
| 4 | Why not parallel execution? | Solo founder time constraints |
| 5 | Why not prioritized? | Underestimated gap until audit prep |

**Root Cause:** New ISMS implementation prioritized documentation over operational execution; solo founder bandwidth limitation.

**Systemic Issue:** No checkpoint in implementation plan to verify operational readiness.

---

## 3. Minimum to Close CA-001

You can close CA-001 once these are true:

| # | Requirement | How to Verify |
|---|-------------|---------------|
| 1 | `~/code/aegis-compliance-evidence/` exists with expected folders | `ls -la ~/code/aegis-compliance-evidence/` |
| 2 | Ran the three export scripts at least once | Files in current month folder |
| 3 | Created an incident log entry for the month | File exists (even "no incidents") |
| 4 | Verification checklist items 1, 6, 7, 9, 10 are checked | This document |
| 5 | (If using repo) Evidence is committed/pushed | `git status` clean |

Everything else (retroactive exports, Q4 folders, etc.) is nice-to-have.

---

## 4. Detailed Execution Steps

### Step 0: Verify Auth (Run First - Reduces Failures)

```bash
# Check GitHub CLI auth
gh auth status

# Check AWS credentials
aws sts get-caller-identity --profile "${AWS_PROFILE:-default}"

# If either fails, fix before proceeding:
# - gh auth login
# - aws configure --profile default
```

### Step 1: Create Vault Directory Structure

```bash
# Run locally (not in sandbox)
VAULT_BASE="$HOME/code/aegis-compliance-evidence"

# Create base structure
mkdir -p "$VAULT_BASE"/{soc2,iso27001}

# SOC 2 evidence structure - use date-driven approach
YEAR=$(date +%Y)
MONTH=$(date +%Y-%m)

mkdir -p "$VAULT_BASE/soc2/2025"/{retroactive,2025-09,2025-10,2025-11,2025-12}
mkdir -p "$VAULT_BASE/soc2/2025/Q4/access-reviews"
mkdir -p "$VAULT_BASE/soc2/$YEAR/$MONTH"/{access-reviews,ci-cd-security,vuln-management,incident-response,logging-monitoring}
mkdir -p "$VAULT_BASE/soc2/$YEAR"/{Q1,Q2,Q3,Q4,governance,risk-management,vendor-management}

# ISO 27001 evidence structure
mkdir -p "$VAULT_BASE/iso27001/$YEAR"/{internal-audits,management-reviews,risk-assessments,corrective-actions}

# Verify
find "$VAULT_BASE" -type d | head -30
```

### Step 2: Extract Retroactive Evidence (Optional - Historical Reference)

> **Note:** Retroactive evidence is exported on today's date from system logs/history for reference. It does not substitute for ongoing operation evidence.

```bash
cd ~/code/aegis-platform-observability-integration
VAULT_BASE="$HOME/code/aegis-compliance-evidence"

# Git commit history (Sept 2025 - Jan 2026)
git log --since="2025-09-01" --until="2026-01-31" --oneline --format='%H,%ai,%an,"%s"' \
  > "$VAULT_BASE/soc2/2025/retroactive/git-commit-history.csv"

# Count commits
echo "Commits extracted: $(wc -l < $VAULT_BASE/soc2/2025/retroactive/git-commit-history.csv)"

# PR history
gh pr list --state merged --limit 100 --json number,title,mergedAt,author \
  > "$VAULT_BASE/soc2/2025/retroactive/github-prs.json"

# GitHub Actions runs
gh run list --limit 200 --json databaseId,status,conclusion,createdAt,workflowName \
  > "$VAULT_BASE/soc2/2025/retroactive/github-actions-runs.json"

# Dependabot alerts (historical)
gh api repos/carlosmsanchezm/aegis-platform/dependabot/alerts \
  > "$VAULT_BASE/soc2/2025/retroactive/dependabot-alerts-platform.json"
gh api repos/carlosmsanchezm/aegis-ui/dependabot/alerts \
  > "$VAULT_BASE/soc2/2025/retroactive/dependabot-alerts-ui.json"
gh api repos/carlosmsanchezm/sovran/dependabot/alerts \
  > "$VAULT_BASE/soc2/2025/retroactive/dependabot-alerts-sovran.json"

# Mark as historical
echo "Exported on $(date -u +%Y-%m-%dT%H:%M:%SZ) for historical reference only" \
  > "$VAULT_BASE/soc2/2025/retroactive/README.md"
```

### Step 3: Run Current Month Evidence Collection (REQUIRED)

```bash
cd ~/code/aegis-platform-observability-integration
VAULT_BASE="$HOME/code/aegis-compliance-evidence"

# Date-driven (re-runnable monthly without editing)
YEAR=$(date +%Y)
MONTH=$(date +%Y-%m)
MONTH_DIR="$VAULT_BASE/soc2/$YEAR/$MONTH"

# Ensure directory exists
mkdir -p "$MONTH_DIR"/{access-reviews,ci-cd-security,vuln-management,incident-response,logging-monitoring}

# GitHub security baseline
./scripts/compliance/export_github_security_baseline.sh "$MONTH_DIR/ci-cd-security"

# AWS security baseline
./scripts/compliance/export_aws_security_baseline.sh "$MONTH_DIR/access-reviews"

# CI/CD reports (may require syft/trivy - skip if not installed)
./scripts/compliance/export_ci_reports.sh "$MONTH_DIR/vuln-management" || echo "CI reports skipped - install syft/trivy for full export"
```

### Step 4: Create Current Month Incident Log (REQUIRED)

```bash
VAULT_BASE="$HOME/code/aegis-compliance-evidence"
YEAR=$(date +%Y)
MONTH=$(date +%Y-%m)
INCIDENT_LOG="$VAULT_BASE/soc2/$YEAR/$MONTH/incident-response/$MONTH-incident-log.md"

cat > "$INCIDENT_LOG" << EOF
# Incident Log - $(date +%B) $YEAR

## Summary
- **Security Incidents:** 0
- **Near Misses:** 0
- **Dependabot Alerts Resolved:** 55 (11 + 25 + 19 across 3 repos)
- **Access Changes:** 0 (solo founder)

## Incidents
None this month.

## Observations
- ISMS documentation completed
- All Dependabot vulnerabilities remediated across 3 repositories
- Evidence vault now operational (CA-001 closed)

## Prepared By
Carlos Sanchez - $(date +%Y-%m-%d)
EOF

echo "Created: $INCIDENT_LOG"
```

### Step 5: Verify Completeness (REQUIRED)

```bash
VAULT_BASE="$HOME/code/aegis-compliance-evidence"
YEAR=$(date +%Y)
MONTH=$(date +%Y-%m)

echo "=== Vault File Count ==="
find "$VAULT_BASE" -type f ! -name ".DS_Store" | wc -l

echo ""
echo "=== Current Month Evidence ($MONTH) ==="
find "$VAULT_BASE/soc2/$YEAR/$MONTH" -type f ! -name ".DS_Store" | head -20

echo ""
echo "=== Key Files Check ==="
echo "Access reviews: $(ls $VAULT_BASE/soc2/$YEAR/$MONTH/access-reviews/*.json 2>/dev/null | wc -l) files"
echo "CI/CD security: $(ls $VAULT_BASE/soc2/$YEAR/$MONTH/ci-cd-security/*.json 2>/dev/null | wc -l) files"
echo "Incident log:   $(ls $VAULT_BASE/soc2/$YEAR/$MONTH/incident-response/*.md 2>/dev/null | wc -l) files"

echo ""
echo "=== ISO 27001 Structure ==="
ls -la "$VAULT_BASE/iso27001/$YEAR/"
```

### Step 6: (Optional) Commit to Private Repo

If your evidence vault is a private git repo:

```bash
cd "$VAULT_BASE"
git status
git add -A
git commit -m "Evidence: baseline exports $(date +%Y-%m)"
git push
```

---

## 5. Verification Checklist

**Minimum to close (items 1, 6, 7, 9, 10):**

| # | Checkpoint | Required? | Verified? | Date |
|---|------------|-----------|-----------|------|
| 1 | Vault directory exists at `~/code/aegis-compliance-evidence/` | ✅ Yes | ☐ | |
| 2 | Retroactive folder has git history CSV | Optional | ☐ | |
| 3 | Retroactive folder has PR JSON | Optional | ☐ | |
| 4 | Retroactive folder has CI runs JSON | Optional | ☐ | |
| 5 | Retroactive folder has Dependabot alerts (3 repos) | Optional | ☐ | |
| 6 | Current month has access review exports | ✅ Yes | ☐ | |
| 7 | Current month has CI/CD security exports | ✅ Yes | ☐ | |
| 8 | Current month has vuln management exports | Optional | ☐ | |
| 9 | Current month has incident log (even if "no incidents") | ✅ Yes | ☐ | |
| 10 | ISO 27001 directory structure exists | ✅ Yes | ☐ | |
| 11 | Monthly evidence export added to calendar | Recommended | ☐ | |
| 12 | Operating plan reviewed and understood | Recommended | ☐ | |

---

## 6. Troubleshooting Common Failures

### GitHub CLI Issues

```bash
# Error: "gh: command not found"
brew install gh

# Error: "not logged in" or scope issues
gh auth login
gh auth refresh -s read:org,repo,security_events

# Error: "Resource not accessible by integration" for Dependabot
# Need security_events scope - re-auth with:
gh auth refresh -s security_events
```

### AWS CLI Issues

```bash
# Error: "Unable to locate credentials"
aws configure --profile default

# Error: "ExpiredToken"
aws sso login --profile default
# or for access keys:
aws configure --profile default

# Wrong profile
export AWS_PROFILE=your-profile-name
```

### Script Permission Issues

```bash
# Error: "Permission denied"
chmod +x scripts/compliance/*.sh
```

---

## 7. Preventive Actions

To prevent recurrence:

| # | Action | Implementation |
|---|--------|----------------|
| 1 | Add "verify evidence collection" step to implementation checklist | Update STATUS.md |
| 2 | Set recurring monthly calendar reminder | First Monday each month |
| 3 | Add evidence verification to internal audit checklist | Done in checklist |
| 4 | Set `$EVIDENCE_VAULT` in shell profile | `echo 'export EVIDENCE_VAULT="$HOME/code/aegis-compliance-evidence"' >> ~/.zshrc` |

---

## 8. Closure

### Closure Criteria Met?

| Criteria | Met? | Evidence |
|----------|------|----------|
| Vault exists with expected structure | ☐ | Directory listing |
| At least one month of evidence collected | ☐ | Files in current month folder |
| Incident log created | ☐ | Incident log file |
| Root cause addressed | ☐ | Operating plan in place |
| Preventive actions implemented | ☐ | Calendar reminder set |

### Closure Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Owner | Carlos Sanchez | _____________ | [Date] |
| Verifier | [Internal Auditor] | _____________ | [Date] |

---

## 9. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.1 | 2026-01-17 | Carlos Sanchez | Added date-driven scripts, auth checks, troubleshooting, clarified minimum requirements |
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial closure plan |
