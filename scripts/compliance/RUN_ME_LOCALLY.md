# Run These Commands Locally

The sandbox environment cannot access external services. Run these commands on your local machine to set up the SOC 2 evidence vault.

## Prerequisites

```bash
# Verify you have gh and aws CLI
gh --version
aws --version

# Authenticate gh if needed
gh auth login

# Verify aws profile
aws sts get-caller-identity --profile aegis-new
```

---

## Option 1: Run the Setup Script

```bash
cd /path/to/aegis-platform
./scripts/compliance/setup_evidence_vault.sh
```

---

## Option 2: Manual Steps

### Step 1: Create the Private Repository

```bash
# Set your GitHub username
export GITHUB_OWNER="carlosmsanchezm"

# Create private repo
gh repo create aegis-compliance-evidence \
    --private \
    --description "SOC 2 Type II Evidence Vault - CONFIDENTIAL" \
    --disable-wiki

# Clone it
gh repo clone ${GITHUB_OWNER}/aegis-compliance-evidence
cd aegis-compliance-evidence
```

### Step 2: Configure Repository Settings

```bash
# Disable GitHub Pages and other features
gh api -X PUT "/repos/${GITHUB_OWNER}/aegis-compliance-evidence" \
    -f has_pages=false \
    -f has_wiki=false \
    -f allow_squash_merge=true \
    -f delete_branch_on_merge=true
```

### Step 3: Copy Setup Files

```bash
# From the aegis-compliance-evidence directory
cp /path/to/aegis-platform/docs/compliance/soc2/evidence-vault-setup/README.md ./README.md
cp /path/to/aegis-platform/docs/compliance/soc2/evidence-vault-setup/.gitignore ./.gitignore
cp /path/to/aegis-platform/docs/compliance/soc2/evidence-vault-setup/EVIDENCE_INTAKE_CHECKLIST.md ./
```

### Step 4: Create Folder Structure

```bash
YEAR=2025

# Governance
mkdir -p soc2/${YEAR}/governance/{org-chart,policies,risk-assessment}
mkdir -p soc2/${YEAR}/vendor-management
mkdir -p soc2/${YEAR}/training-policy-ack
mkdir -p soc2/${YEAR}/backups-dr

# Quarterly
for q in Q1 Q2 Q3 Q4; do
    mkdir -p soc2/${YEAR}/${q}/access-reviews
done

# Monthly
for month in 01 02 03 04 05 06 07 08 09 10 11 12; do
    mkdir -p soc2/${YEAR}/${YEAR}-${month}/{access-reviews,change-management,ci-cd-security,vuln-management,logging-monitoring,incident-response,communication}
done

# Templates
mkdir -p soc2/templates

# Add .gitkeep to empty directories
find soc2 -type d -empty -exec touch {}/.gitkeep \;
```

### Step 5: Initial Commit

```bash
git add -A
git commit -m "Initialize SOC 2 evidence vault structure

- Add folder structure for ${YEAR}
- Add README with naming conventions and policies
- Add .gitignore for secrets protection
- Add evidence intake checklist"

git push origin main
```

### Step 6: Set Up Branch Protection (Solo-Founder Safe)

```bash
gh api -X PUT "/repos/${GITHUB_OWNER}/aegis-compliance-evidence/branches/main/protection" \
    -f required_status_checks='null' \
    -F enforce_admins=false \
    -f required_pull_request_reviews='null' \
    -f restrictions='null' \
    -F allow_force_pushes=false \
    -F allow_deletions=false
```

---

## Step 7: Run First Evidence Collection

```bash
# From the aegis-platform repo
cd /path/to/aegis-platform

# Set output paths
VAULT_PATH="/path/to/aegis-compliance-evidence"
CURRENT_MONTH=$(date +%Y-%m)

# Export GitHub security baseline
GITHUB_OWNER=carlosmsanchezm GITHUB_REPOS=aegis-platform \
    ./scripts/compliance/export_github_security_baseline.sh \
    ${VAULT_PATH}/soc2/2025/${CURRENT_MONTH}/ci-cd-security

# Export AWS security baseline
AWS_PROFILE=aegis-new \
    ./scripts/compliance/export_aws_security_baseline.sh \
    ${VAULT_PATH}/soc2/2025/${CURRENT_MONTH}/access-reviews

# Export CI reports (SBOM, vulnerability scans)
./scripts/compliance/export_ci_reports.sh \
    ${VAULT_PATH}/soc2/2025/${CURRENT_MONTH}/vuln-management
```

### Step 8: Commit Evidence

```bash
cd ${VAULT_PATH}
git add -A
git status  # Review what's being committed - NO SECRETS!
git commit -m "SOC2 evidence: baseline exports ${CURRENT_MONTH}"
git push
```

---

## 🚨 IMPORTANT: Rotate Your GitHub PAT

You shared your GitHub Personal Access Token in the chat. **Rotate it now:**

1. Go to https://github.com/settings/tokens
2. Delete the exposed token: `REDACTED...`
3. Create a new token with the same permissions
4. Update your local `gh` auth: `gh auth login`

---

## Verification

After setup, verify:

```bash
# Check repo exists and is private
gh repo view carlosmsanchezm/aegis-compliance-evidence --json isPrivate

# Check branch protection
gh api /repos/carlosmsanchezm/aegis-compliance-evidence/branches/main/protection

# List evidence files
ls -la ${VAULT_PATH}/soc2/2025/$(date +%Y-%m)/
```
