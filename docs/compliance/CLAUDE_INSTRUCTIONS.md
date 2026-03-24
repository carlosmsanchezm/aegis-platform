# Compliance Task Instructions for Claude

**Purpose:** This file tells Claude exactly what to read and do for weekly, monthly, and quarterly compliance tasks.

**Critical principle:** Compliance is not just recording — it is **detect, remediate, document**. Every task must attempt to FIX what it finds, not just report it.

**Usage:** When asking Claude to run compliance tasks, reference this file:
```
Read docs/compliance/CLAUDE_INSTRUCTIONS.md and execute the [weekly/monthly/quarterly] task.
```

---

## Quick Reference

| Task | Frequency | One-Liner Prompt |
|------|-----------|------------------|
| Weekly | Every Monday | "Run the weekly compliance check per docs/compliance/CLAUDE_INSTRUCTIONS.md" |
| Monthly | 1st of month | "Run the monthly evidence collection per docs/compliance/CLAUDE_INSTRUCTIONS.md" |
| Quarterly | Q1/Q2/Q3/Q4 | "Run the quarterly access review per docs/compliance/CLAUDE_INSTRUCTIONS.md" |

---

## Weekly Security Check + Remediation (Every Monday)

### Files to Read (in order)
1. `docs/compliance/COMPLIANCE_PROGRAM_STATUS.md` - Current program status
2. `docs/compliance/OPERATIONS_RUNBOOK.md` - Detailed procedures
3. `scripts/compliance/run_weekly_checks.sh` - The script to run

### Phase 1: Detect
1. Read the files above to understand current state
2. Verify environment:
   ```bash
   echo $EVIDENCE_VAULT
   gh auth status
   ```
3. Run the weekly check:
   ```bash
   cd ~/code/aegis-platform
   ./scripts/compliance/run_weekly_checks.sh
   ```
4. Review output — note counts by severity:
   - Dependabot alerts (critical/high/medium/low) per repo
   - Secret scanning alerts
   - Code scanning alerts

### Phase 2: Remediate (REQUIRED — do not skip)

For any critical or high findings, actively fix them:

5. **Get alert details** — Use `gh api` to fetch specific alert info (package name, fixed version, manifest path) for each repo:
   ```bash
   gh api "/repos/carlosmsanchezm/REPO/dependabot/alerts?state=open&severity=critical,high" | jq '...'
   ```
6. **Fix vulnerabilities in aegis-platform** (this repo):
   - Go deps: Update version in go.mod, run `go mod tidy`, verify with `go build ./...`
   - If deps are already patched in go.mod but alerts persist, run `go mod tidy` to sync go.sum
7. **Fix vulnerabilities in aegis-ui** (`~/code/aegis-ui/`):
   - Check if direct or transitive dep
   - Direct: `yarn upgrade <package>@<version>`
   - Transitive: Add/update `resolutions` in package.json
   - Run `yarn install`, verify with `yarn tsc` and `yarn build:backend`
   - Commit on a `security/` branch
8. **Fix vulnerabilities in sovran** (`~/code/sovran/`):
   - NPM: Add `overrides` in package.json (both root and aegis-vscode-remote/extension/)
   - Python: Update constraints in pyproject.toml, run `uv lock`
   - Verify with `npm run build` and `uv sync`
   - Commit on a `security/` branch
9. **For findings that cannot be auto-fixed** (breaking API changes, vendor issues, human judgment needed):
   - Flag clearly in the report with the reason
   - If appropriate, open a corrective action in `docs/compliance/iso27001/09-corrective-actions-log.csv`

### Phase 3: Document (REQUIRED — across ALL frameworks)

10. **Write remediation evidence** to the vault:
    ```
    $EVIDENCE_VAULT/soc2/YYYY/YYYY-MM/vuln-management/YYYY-MM-DD_vulnerability_remediation_report.md
    ```
    Include: what was found, what was fixed, packages/versions, commits, build verification, and which controls are demonstrated.

11. **Update ALL framework status docs** (not just the master):

    | Doc | What to update |
    |-----|---------------|
    | `docs/compliance/COMPLIANCE_PROGRAM_STATUS.md` | Last Updated, vault commit, alert totals, This Week checklist, evidence timeline, document history |
    | `docs/compliance/soc2/STATUS.md` | Vulnerability Remediation section (CC7.1), observation period, evidence vault commit |
    | `docs/compliance/iso27001/STATUS.md` | Vulnerability Management section (A.8.8, A.8.25, A.8.28, A.8.32) |
    | `docs/compliance/cmmc/STATUS.md` | Active Control Evidence section (SI.L2-3.14.1, RA.L2-3.11.2) |
    | `docs/compliance/fedramp/STATUS.md` | Active Control Evidence section (SI-2, RA-5, CM-3, SA-11) |
    | `docs/compliance/iso27001/04-risk-treatment-plan.csv` | RISK-001 (supply chain) completion evidence if vuln remediation performed |

12. **Check corrective actions** — Read `docs/compliance/iso27001/09-corrective-actions-log.csv` for CAs approaching their due dates. Flag any due within 14 days in the report.

13. **Commit** all changes with message: `compliance: weekly security check + remediation YYYY-MM-DD`

### Expected Output Location
```
$EVIDENCE_VAULT/soc2/YYYY/YYYY-MM/ci-cd-security/weekly/
├── YYYY-MM-DD_security_alerts.json
├── YYYY-MM-DD_weekly_review.md
└── RUN_LOG_YYYY-MM-DD.txt

$EVIDENCE_VAULT/soc2/YYYY/YYYY-MM/vuln-management/
└── YYYY-MM-DD_vulnerability_remediation_report.md   (if remediations performed)
```

### Sample Report Format
```
## Weekly Security Check + Remediation - YYYY-MM-DD

**Repositories Checked:** aegis-platform, aegis-ui, sovran

### Detection
| Metric | Count |
|--------|-------|
| Dependabot Alerts (open) | X |
| Critical | X |
| High | X |
| Secret Scanning | X |
| Code Scanning | X |

### Remediation
| Repo | Critical Fixed | High Fixed | Method | Commit |
|------|---------------|------------|--------|--------|
| aegis-platform | X | X | Go mod upgrade | abc1234 |
| aegis-ui | X | X | Yarn resolutions | def5678 |
| sovran | X | X | npm overrides + pip | ghi9012 |

### Controls Demonstrated
- SOC 2 CC7.1, ISO A.8.8, CMMC SI.L2-3.14.1, FedRAMP SI-2/RA-5

### Human Attention Required
- [List items that could not be auto-fixed, with reasons]
- [CAs approaching due date]

**Evidence:** $EVIDENCE_VAULT/soc2/YYYY/YYYY-MM/
```

---

## Weekly SBOM Generation (part of weekly cadence)

After running the security check and remediation, generate SBOMs:

```bash
cd ~/code/aegis-platform
chmod +x docs/compliance/sbom/generate-sbom.sh
./docs/compliance/sbom/generate-sbom.sh
```

Copy SBOM outputs to the evidence vault:
```bash
mkdir -p $EVIDENCE_VAULT/soc2/$(date +%Y)/$(date +%Y-%m)/vuln-management/sbom/
cp docs/compliance/sbom/*-sbom.json $EVIDENCE_VAULT/soc2/$(date +%Y)/$(date +%Y-%m)/vuln-management/sbom/
```

**Prerequisites:** `cyclonedx-gomod`, `cdxgen` (npm). If not installed, skip with a note — do not block the weekly run.

**Controls:** SI-2 (Flaw Remediation), CM-8 (System Component Inventory), A.8.8

---

## Weekly Container Image Scanning (part of weekly cadence)

Scan all Aegis container images for vulnerabilities using Grype or Trivy:

```bash
# Using grype (preferred)
for image in carlosmsanchez/aegis-platform-api:dev carlosmsanchez/aegis-proxy:dev carlosmsanchez/aegis-k8s-agent:dev carlosmsanchez/aegis-workspace-vscode:latest; do
  grype $image -o json > $EVIDENCE_VAULT/soc2/$(date +%Y)/$(date +%Y-%m)/vuln-management/$(date +%Y-%m-%d)_$(echo $image | tr '/:' '_')_scan.json 2>/dev/null || echo "SKIP: grype not installed or image not available"
done

# Alternative: using trivy
for image in ...; do
  trivy image $image --format json --output $EVIDENCE_VAULT/...
done
```

Report findings in the same table format as dependency alerts. If critical/high container vulnerabilities are found, remediate by updating base images or pinning fixed versions in Dockerfiles.

**Prerequisites:** `grype` or `trivy`. If neither installed, skip with a note — do not block the weekly run.

**Controls:** RA-5 (Vulnerability Monitoring and Scanning), A.8.7 (Protection against malware), Iron Bank readiness

---

## Monthly Evidence Collection (1st of Month)

### Files to Read (in order)
1. `docs/compliance/COMPLIANCE_PROGRAM_STATUS.md` - Current program status
2. `docs/compliance/OPERATIONS_RUNBOOK.md` - Detailed procedures
3. `scripts/compliance/run_monthly_evidence.sh` - The script to run

### Steps
1. Read the files above to understand current state
2. Verify environment:
   ```bash
   echo $EVIDENCE_VAULT
   gh auth status
   aws sts get-caller-identity --profile "${AWS_PROFILE:-default}"
   ```
3. Run the monthly evidence collection:
   ```bash
   cd ~/code/aegis-platform
   ./scripts/compliance/run_monthly_evidence.sh
   ```
4. If evidence vault is a git repo, commit and push:
   ```bash
   cd $EVIDENCE_VAULT
   git add .
   git commit -m "Monthly evidence: YYYY-MM (XX files)"
   git push origin main
   ```
5. **Run a weekly security check + remediation** (same as weekly Phase 1-2 above) to ensure the monthly snapshot includes fresh vulnerability data and any new alerts are fixed
6. Update ALL status docs (same as weekly Phase 3, step 11):
   - `COMPLIANCE_PROGRAM_STATUS.md` — Last Updated, vault commit, total files
   - `soc2/STATUS.md` — Evidence vault commit, observation period
   - `iso27001/STATUS.md` — Evidence section if remediation performed
   - `cmmc/STATUS.md` — Active evidence if controls advanced
   - `fedramp/STATUS.md` — Active evidence if controls advanced

### Expected Output Location
```
$EVIDENCE_VAULT/soc2/YYYY/YYYY-MM/
├── RUN_LOG_YYYY-MM-DD.txt
├── YYYY-MM-DD_monthly_summary.md
├── access-reviews/          # AWS IAM, MFA, roles
├── ci-cd-security/          # GitHub security exports
├── vuln-management/         # SBOM, vulnerability scans
├── incident-response/       # Incident log
└── logging-monitoring/      # CloudTrail exports
```

### Sample Report Format
```
## Monthly Evidence Collection - YYYY-MM

**Collection Date:** YYYY-MM-DD
**Evidence Vault:** $EVIDENCE_VAULT

| Category | Files | Status |
|----------|-------|--------|
| GitHub Security | X | ✅ |
| AWS Security | X | ✅ |
| CI/CD Reports | X | ✅ |
| Incident Log | 1 | ✅ |
| **Total** | **XX** | |

**Evidence Vault Commit:** `abc1234`
**COMPLIANCE_PROGRAM_STATUS.md Updated:** Yes/No

**Next Monthly Run:** YYYY-MM-01
```

---

## Quarterly Access Review (Q1: Jan 15, Q2: Apr 15, Q3: Jul 15, Q4: Oct 15)

### Files to Read (in order)
1. `docs/compliance/COMPLIANCE_PROGRAM_STATUS.md` - Current program status
2. `docs/compliance/OPERATIONS_RUNBOOK.md` - Detailed procedures
3. `docs/compliance/soc2/procedures/quarterly-access-review.md` - Access review procedure

### Steps
1. Read the files above to understand the procedure
2. Run monthly evidence first to get fresh access data:
   ```bash
   ./scripts/compliance/run_monthly_evidence.sh
   ```
3. Review access data:
   - `$EVIDENCE_VAULT/soc2/YYYY/YYYY-MM/access-reviews/` (AWS IAM)
   - `$EVIDENCE_VAULT/soc2/YYYY/YYYY-MM/ci-cd-security/*/collaborators.json` (GitHub)
4. For each system, verify:
   - [ ] All accounts are for current employees/contractors
   - [ ] Access levels are appropriate for roles
   - [ ] No stale accounts (unused > 90 days)
   - [ ] MFA enabled for all accounts
5. Document findings in access review report
6. Save report to evidence vault:
   ```bash
   cp YYYY-QX_access_review.md $EVIDENCE_VAULT/soc2/YYYY/YYYY-MM/access-reviews/
   ```

### Sample Report Format
```
## Quarterly Access Review - YYYY-QX

**Review Period:** YYYY-MM-DD to YYYY-MM-DD
**Reviewer:** Carlos Sanchez
**Date Completed:** YYYY-MM-DD

### Systems Reviewed

| System | Accounts | Active | Stale | MFA | Status |
|--------|----------|--------|-------|-----|--------|
| GitHub | X | X | 0 | N/A | ✅ |
| AWS IAM | X | X | 0 | X | ✅ |

### Findings

| # | System | Finding | Action | Due |
|---|--------|---------|--------|-----|
| 1 | - | No findings | - | - |

### Attestation

I certify that I have reviewed all user access and confirm it is appropriate.

**Signature:** Carlos Sanchez
**Date:** YYYY-MM-DD
```

---

## Semi-Annual Internal Audit (Feb, Aug)

### Files to Read (in order)
1. `docs/compliance/COMPLIANCE_PROGRAM_STATUS.md`
2. `docs/compliance/iso27001/11-internal-audit-plan.md`
3. `docs/compliance/iso27001/12-internal-audit-checklist.md`
4. `docs/compliance/iso27001/2026-02-internal-audit-report-PREFILLED.md`

### Steps
1. **Schedule external auditor** (required for solo founder per ISO 27001 Clause 9.2)
2. Prepare audit package:
   - Run monthly evidence collection
   - Provide auditor access to evidence vault
   - Share audit checklist
3. Auditor conducts audit using checklist
4. Save completed audit report to:
   ```
   $EVIDENCE_VAULT/iso27001/YYYY/internal-audits/
   ```
5. Open corrective actions for any findings
6. Schedule management review

---

## Troubleshooting

### EVIDENCE_VAULT Not Set
```bash
export EVIDENCE_VAULT="$HOME/code/aegis-compliance-evidence"
# Add to ~/.zshrc for permanence
```

### GitHub CLI Not Authenticated
```bash
gh auth login
gh auth status
```

### AWS CLI Not Authenticated
```bash
aws sso login --profile $AWS_PROFILE
# Or use --skip-aws flag
./scripts/compliance/run_monthly_evidence.sh --skip-aws
```

### Script Permission Denied
```bash
chmod +x scripts/compliance/*.sh
```

---

## FedRAMP Readiness Tasks

### FedRAMP Gap Review (As Needed)

When asked to review FedRAMP readiness or work on FedRAMP tasks:

### Files to Read (in order)
1. `docs/compliance/fedramp/STATUS.md` - FedRAMP readiness status
2. `docs/compliance/fedramp/02-gap-analysis.md` - Detailed gap analysis
3. `docs/compliance/fedramp/00-scope-and-impact-assessment.md` - FIPS 199 categorization
4. `docs/compliance/fedramp/01-control-mapping.csv` - Control reuse from SOC 2/ISO

### FedRAMP Status Quick Summary
- **Target Baseline:** FedRAMP 20x Low or LI-SaaS
- **Current Phase:** Pre-Assessment (Research & Planning)
- **Reuse from SOC 2/ISO 27001:** ~25-30% of controls

### Critical Gaps (Must Address Before 3PAO)
1. FIPS 140-2/3 cryptographic modules (IA-7, SC-13)
2. System Security Plan (SSP) with all appendices
3. Agency sponsor or Program Authorization path
4. 3PAO selection and engagement
5. Continuous monitoring plan (ConMon)
6. Supply Chain Risk Management Plan (SCRMP)

### FedRAMP Document Locations

| Document | Path |
|----------|------|
| FedRAMP Status | `docs/compliance/fedramp/STATUS.md` |
| Scope & Impact Assessment | `docs/compliance/fedramp/00-scope-and-impact-assessment.md` |
| Control Mapping | `docs/compliance/fedramp/01-control-mapping.csv` |
| Gap Analysis | `docs/compliance/fedramp/02-gap-analysis.md` |
| SSP Outline | `docs/compliance/fedramp/03-ssp-outline.md` |
| SSP Sections 1-3 | `docs/compliance/fedramp/04-ssp-sections-1-3.md` |
| Appendix E: Digital Identity | `docs/compliance/fedramp/appendix-e-digital-identity-worksheet.md` |
| Appendix F: Rules of Behavior | `docs/compliance/fedramp/appendix-f-rules-of-behavior.md` |
| Appendix L: Separation of Duties | `docs/compliance/fedramp/appendix-l-separation-of-duties-matrix.md` |

---

## CMMC Readiness Tasks

### CMMC Gap Review (As Needed)

When asked to review CMMC readiness or work on CMMC tasks:

### Files to Read (in order)
1. `docs/compliance/cmmc/STATUS.md` - CMMC program status
2. `docs/compliance/cmmc/01-control-mapping.csv` - 97 controls mapped to NIST 800-171
3. `docs/compliance/cmmc/02-gap-analysis.md` - Detailed gap analysis with remediation roadmap
4. `docs/compliance/cmmc/03-sprs-score-worksheet.md` - SPRS score calculation

### CMMC Status Quick Summary
- **Target Level:** CMMC Level 2 (NIST 800-171 Rev 3)
- **Current Phase:** Foundation Building
- **SPRS Score:** 38/110 (preliminary)
- **Reuse from SOC 2/ISO 27001:** ~40% of controls

### SPRS Score Recalculation

When controls are implemented or updated:
1. Read `docs/compliance/cmmc/01-control-mapping.csv`
2. Update the `Aegis_Status` column for changed controls
3. Read `docs/compliance/cmmc/03-sprs-score-worksheet.md`
4. Update individual control scores (5=full, 3=partial, 1=planned, 0=not implemented)
5. Recalculate the total SPRS score
6. Update `docs/compliance/cmmc/STATUS.md` with new score

### Control Implementation Statement Updates

When new controls are implemented:
1. Read `docs/compliance/customer-docs/control-implementation-statements.md`
2. Add or update the control entry following the existing format:
   - **Aegis Provides**: What the platform handles
   - **Customer Configures**: Settings customers configure
   - **Customer Implements**: Customer responsibilities
   - **Evidence**: How to demonstrate compliance
3. Update `docs/compliance/customer-docs/customer-responsibility-matrix.md` if new family
4. Update the CMMC control mapping CSV status

### CMMC Document Locations

| Document | Path |
|----------|------|
| CMMC Status | `docs/compliance/cmmc/STATUS.md` |
| Control Mapping (97 controls) | `docs/compliance/cmmc/01-control-mapping.csv` |
| Gap Analysis | `docs/compliance/cmmc/02-gap-analysis.md` |
| SPRS Score Worksheet | `docs/compliance/cmmc/03-sprs-score-worksheet.md` |
| Control Implementation Statements | `docs/compliance/customer-docs/control-implementation-statements.md` |
| Customer Responsibility Matrix | `docs/compliance/customer-docs/customer-responsibility-matrix.md` |

---

## Key File Locations

| Purpose | Path |
|---------|------|
| Program Status (single source of truth) | `docs/compliance/COMPLIANCE_PROGRAM_STATUS.md` |
| Operations Runbook | `docs/compliance/OPERATIONS_RUNBOOK.md` |
| Weekly Script | `scripts/compliance/run_weekly_checks.sh` |
| Monthly Script | `scripts/compliance/run_monthly_evidence.sh` |
| SOC 2 Status | `docs/compliance/soc2/STATUS.md` |
| ISO 27001 Status | `docs/compliance/iso27001/STATUS.md` |
| **FedRAMP Status** | `docs/compliance/fedramp/STATUS.md` |
| FedRAMP SSP Sections 1-3 | `docs/compliance/fedramp/04-ssp-sections-1-3.md` |
| **CMMC Status** | `docs/compliance/cmmc/STATUS.md` |
| CMMC Control Mapping | `docs/compliance/cmmc/01-control-mapping.csv` |
| CMMC Gap Analysis | `docs/compliance/cmmc/02-gap-analysis.md` |
| CMMC SPRS Worksheet | `docs/compliance/cmmc/03-sprs-score-worksheet.md` |
| Access Review Procedure | `docs/compliance/soc2/procedures/quarterly-access-review.md` |
| Internal Audit Plan | `docs/compliance/iso27001/11-internal-audit-plan.md` |
| Internal Audit Checklist | `docs/compliance/iso27001/12-internal-audit-checklist.md` |
| Corrective Actions Log | `docs/compliance/iso27001/09-corrective-actions-log.csv` |

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 2.0 | 2026-03-23 | Major rewrite: Added Phase 2 (Remediate) and Phase 3 (Document across ALL frameworks) to weekly task. Fixed wrong repo paths. Added cross-framework status doc update requirements. Added remediation evidence template. Encoded detect→remediate→document principle. |
| 1.2 | 2026-03-09 | Added CMMC section (key files, SPRS recalculation, control statement updates); updated FedRAMP docs with SSP + appendices |
| 1.1 | 2026-01-17 | Added FedRAMP readiness tasks |
| 1.0 | 2026-01-17 | Initial instructions |
