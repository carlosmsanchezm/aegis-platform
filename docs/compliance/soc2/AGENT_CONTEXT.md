# SOC 2 Type II Compliance - AI Agent Context Document

**Document Purpose:** Provide full context for AI agents working on SOC 2 Type II compliance for Aegis Technologies. Read this document first before performing any compliance-related work.

**Last Updated:** 2026-01-17
**Company:** Aegis Technologies
**Founder:** Carlos Sanchez (@carlosmsanchezm)

---

## 1. Company Context

### 1.1 What is Aegis?

Aegis is a **multi-cluster GPU workload orchestration platform** that helps organizations manage and run AI/ML workloads across distributed Kubernetes clusters.

### 1.2 Company Size

**Solo-founder company.** Carlos Sanchez is the only person. This affects:
- All policies are owned and signed by Carlos
- Segregation of duties has compensating controls (audit trails, automation)
- Internal audits must use external auditors (can't audit yourself)
- Management reviews are self-reviews (documented formally)

### 1.3 Deployment Model

**Model B (Self-Hosted Software)** - Customers deploy and operate Aegis in their own environments.

This means:
- Aegis does NOT operate production environments for customers
- Aegis does NOT process customer data in its own infrastructure
- SOC 2 scope covers SDLC and corporate systems, NOT customer operations
- Availability TSC is OUT OF SCOPE (customers manage their own availability)

---

## 2. SOC 2 Scope

### 2.1 Trust Services Criteria

| Criteria | In Scope | Rationale |
|----------|----------|-----------|
| **Security** | ✅ Yes | Always required |
| Availability | ❌ No | Model B - customers operate their own environments |
| Processing Integrity | ❌ No | Not processing customer data |
| Confidentiality | ❌ No | Future consideration |
| Privacy | ❌ No | Minimal PII processing |

### 2.2 Systems in Scope

| System | Purpose | Owner |
|--------|---------|-------|
| GitHub (3 repos) | Source control, CI/CD | carlosmsanchezm |
| AWS GovCloud | Development infrastructure | Carlos Sanchez |
| Developer workstations | Development | Carlos Sanchez |
| Google Workspace (planned) | Corporate identity | Carlos Sanchez |

### 2.3 Repositories in Scope

```
┌────────────────────────────────────┬─────────────────────────────┬─────────────────┐
│ Repository                         │ Purpose                     │ Controls        │
├────────────────────────────────────┼─────────────────────────────┼─────────────────┤
│ carlosmsanchezm/aegis-platform     │ Platform backend, K8s       │ ✅ Branch prot  │
│                                    │ agents, proxy               │ ✅ Dependabot   │
│                                    │                             │ ✅ CODEOWNERS   │
├────────────────────────────────────┼─────────────────────────────┼─────────────────┤
│ carlosmsanchezm/aegis-ui           │ Frontend UI (Backstage)     │ ✅ Branch prot  │
│                                    │                             │ ✅ Dependabot   │
├────────────────────────────────────┼─────────────────────────────┼─────────────────┤
│ carlosmsanchezm/sovran             │ Infrastructure/IaC,         │ ✅ Branch prot  │
│                                    │ VS Code extension           │ ✅ Dependabot   │
└────────────────────────────────────┴─────────────────────────────┴─────────────────┘
```

### 2.4 Out of Scope

- Customer-deployed Aegis instances
- Customer data processed in customer environments
- Physical security (fully remote, cloud-hosted)
- Production operations (Model B only)

---

## 3. Evidence Locations

### 3.1 Documentation Repository

```
aegis-platform/docs/compliance/
├── soc2/
│   ├── STATUS.md                    # Current compliance status
│   ├── 00-scope-and-systems.md      # Audit scope definition
│   ├── 01-control-narratives.md     # Control descriptions
│   ├── control-matrix.csv           # Control mapping (30 controls)
│   ├── policy-register.md           # Policy tracking
│   ├── retroactive-evidence-summary.md
│   ├── policies/                    # 6 approved policies
│   │   ├── information-security-policy.md (ISP-001)
│   │   ├── access-control-policy.md (ACP-001)
│   │   ├── change-management-policy.md (CMP-001)
│   │   ├── incident-response-policy.md (IRP-001)
│   │   ├── risk-management-policy.md (RMP-001)
│   │   └── vendor-management-policy.md (VMP-001)
│   ├── procedures/
│   │   ├── employee-onboarding-checklist.md (CC6.2)
│   │   ├── employee-offboarding-checklist.md (CC6.3)
│   │   └── quarterly-access-review.md (CC6.4)
│   └── evidence-vault-setup/
│       ├── MONTHLY_CHECKLIST.md     # Monthly tasks
│       └── EVIDENCE_INTAKE_CHECKLIST.md
└── iso27001/                        # ISO 27001 framework (parallel effort)
```

### 3.2 Evidence Vault (Separate Private Repository)

```
aegis-compliance-evidence/           # PRIVATE repo
├── soc2/
│   └── 2025/
│       ├── retroactive/             # Sept 2025 - Jan 2026 historical evidence
│       │   ├── git-commit-history.csv
│       │   ├── github-prs.json
│       │   ├── github-actions-runs.json
│       │   └── dependabot-alerts.json
│       ├── 2025-01/                 # Monthly evidence
│       │   ├── access-reviews/      # AWS exports
│       │   └── ci-cd-security/      # GitHub exports
│       ├── Q1/access-reviews/       # Quarterly reviews
│       ├── Q2/access-reviews/
│       ├── Q3/access-reviews/
│       └── Q4/access-reviews/
└── iso27001/                        # ISO-specific evidence
```

### 3.3 Evidence Collection Scripts

```
aegis-platform/scripts/compliance/
├── export_github_security_baseline.sh  # GitHub security exports
├── export_aws_security_baseline.sh     # AWS security exports
├── export_ci_reports.sh                # SBOM and vuln scans
└── RUN_ME_LOCALLY.md                   # Instructions
```

---

## 4. Current Status (As of 2026-01-17)

### 4.1 Compliance Readiness

| Phase | Status |
|-------|--------|
| Scope Definition | ✅ Complete |
| Control Matrix | ✅ Complete (30 controls) |
| Policies | ✅ 6 policies v1.0 approved |
| Procedures | ✅ Onboarding, offboarding, access review |
| Evidence Vault | ✅ Operational |
| Retroactive Evidence | ✅ 4+ months (Sept 2025 - Jan 2026) |
| Vulnerability Remediation | ✅ All repos at 0 vulnerabilities |

### 4.2 Observation Period

| Metric | Value |
|--------|-------|
| Start Date | September 2025 |
| Current Duration | 4+ months |
| Minimum Required | 3-6 months |
| Status | ✅ Ready for auditor engagement |

### 4.3 Evidence Counts (Retroactive)

| Evidence Type | Count | Controls |
|---------------|-------|----------|
| Git commits | 725 (465+99+161) | CC8.1 |
| Pull requests | 58+ | CC8.1, CC6.1 |
| CI/CD runs | 200+ | CC7.2, CC8.1 |
| Dependabot alerts fixed | 55 (11+25+19) | CC7.1 |

---

## 5. Standing Instructions for AI Agents

### 5.1 When Working on SOC 2 Compliance

**ALWAYS:**
1. Check `docs/compliance/soc2/STATUS.md` first for current state
2. Reference the control matrix for control IDs (CC1.1, CC6.4, etc.)
3. Use evidence naming convention: `YYYY-MM-DD_<system>_<control>_<description>.<ext>`
4. Update STATUS.md after completing any significant work
5. Remember: solo founder = Carlos Sanchez owns/signs everything

**NEVER:**
1. Commit secrets or credentials to any repository
2. Modify production systems without explicit approval
3. Create new policies without updating the policy register
4. Skip documentation - auditors need paper trails

### 5.2 Monthly Evidence Collection (Run on 1st of Each Month)

```bash
# Set variables
VAULT_PATH="../aegis-compliance-evidence"
MONTH=$(date +%Y-%m)
YEAR=$(date +%Y)

# 1. GitHub security baseline (ALL 3 REPOS)
GITHUB_OWNER=carlosmsanchezm GITHUB_REPOS="aegis-platform aegis-ui sovran" \
    ./scripts/compliance/export_github_security_baseline.sh \
    ${VAULT_PATH}/soc2/${YEAR}/${MONTH}/ci-cd-security/

# 2. AWS security baseline
AWS_PROFILE=aegis-new \
    ./scripts/compliance/export_aws_security_baseline.sh \
    ${VAULT_PATH}/soc2/${YEAR}/${MONTH}/access-reviews/

# 3. Commit evidence
cd ${VAULT_PATH}
git add -A
git status  # Verify NO SECRETS
git commit -m "SOC2 evidence: ${MONTH} monthly collection"
git push
```

### 5.3 Quarterly Access Review (Due 15th of Apr/Jul/Oct/Jan)

1. Run evidence collection scripts
2. Generate access matrix for all systems
3. Verify all access is still appropriate
4. Document any changes made
5. Store evidence in `soc2/YYYY/QX/access-reviews/`

See: `docs/compliance/soc2/procedures/quarterly-access-review.md`

### 5.4 When Vulnerabilities Are Found

1. **Dependabot alerts** - Fix within SLA:
   - Critical/High: 7 days
   - Medium: 30 days
   - Low: 90 days

2. **Process:**
   ```bash
   # Check for open alerts
   gh api repos/carlosmsanchezm/aegis-platform/dependabot/alerts --jq '.[] | select(.state == "open")'
   gh api repos/carlosmsanchezm/aegis-ui/dependabot/alerts --jq '.[] | select(.state == "open")'
   gh api repos/carlosmsanchezm/sovran/dependabot/alerts --jq '.[] | select(.state == "open")'

   # Fix vulnerabilities (language-specific)
   # Go: go get <package>@latest && go mod tidy
   # Node/Yarn: Add to "resolutions" in package.json
   # Python: Update in pyproject.toml

   # Commit with security prefix
   git commit -m "security: fix <CVE> vulnerability in <package>"
   ```

3. **Document** the fix in STATUS.md vulnerability remediation section

### 5.5 When Adding New Repositories to Scope

1. Update `docs/compliance/soc2/00-scope-and-systems.md`
2. Enable GitHub security controls:
   ```bash
   # Branch protection
   gh api repos/carlosmsanchezm/<repo>/branches/main/protection -X PUT \
       -f required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true}' \
       -f enforce_admins=true \
       -f required_status_checks=null

   # Enable Dependabot
   gh api repos/carlosmsanchezm/<repo>/vulnerability-alerts -X PUT
   ```
3. Fix any existing Dependabot alerts
4. Update evidence collection scripts to include new repo
5. Update STATUS.md

---

## 6. Control Reference Quick Guide

### 6.1 Key Controls by Category

| Control | Name | Evidence Source |
|---------|------|-----------------|
| CC1.1-CC1.5 | Control Environment | Policies, org chart, training records |
| CC2.1-CC2.3 | Communication | Security announcements, notifications |
| CC3.1-CC3.4 | Risk Assessment | Risk register, risk assessment docs |
| CC4.1-CC4.2 | Monitoring | Dashboards, alert configurations |
| CC5.1-CC5.3 | Control Activities | This control matrix, policies |
| CC6.1-CC6.8 | Logical Access | GitHub/AWS access configs, MFA evidence |
| CC7.1-CC7.5 | System Operations | Vuln scans, incident records, logs |
| CC8.1 | Change Management | PRs, CI/CD runs, branch protection |
| CC9.1-CC9.2 | Risk Mitigation | Vendor assessments, BC/DR plans |

### 6.2 Evidence Generation Commands

| Control | Command |
|---------|---------|
| CC6.x (Access) | `./scripts/compliance/export_github_security_baseline.sh` |
| CC6.6 (MFA) | `./scripts/compliance/export_aws_security_baseline.sh` |
| CC7.1 (Vulns) | `gh api repos/.../dependabot/alerts` |
| CC7.2 (Monitoring) | `./scripts/compliance/export_aws_security_baseline.sh` |
| CC8.1 (Change) | `gh pr list --state all --json ...` |

---

## 7. Auditor Engagement Info

### 7.1 Key Messages for Auditor

1. **Observation period:** Sept 2025 - Present (4+ months)
2. **Solo founder:** All roles held by Carlos Sanchez, compensating controls documented
3. **Model B only:** Self-hosted software, customers operate their own environments
4. **Security TSC only:** Availability out of scope for Model B
5. **Evidence vault:** Private repository with structured evidence

### 7.2 Recommended Audit Firms

- Schellman (SOC 2 + ISO 27001)
- A-LIGN
- Prescient Assurance
- Johanson Group

### 7.3 Estimated Costs

| Item | Cost |
|------|------|
| SOC 2 Type II audit | $10,000-20,000 |
| Combined with ISO 27001 | +$5,000-10,000 |
| Annual surveillance | $5,000-10,000 |

---

## 8. Common Tasks Reference

### 8.1 Check Current Compliance Status

```bash
cat docs/compliance/soc2/STATUS.md
```

### 8.2 Verify All Repos Have Security Controls

```bash
for repo in aegis-platform aegis-ui sovran; do
    echo "=== $repo ==="
    gh api repos/carlosmsanchezm/$repo/branches/main/protection --jq '.required_pull_request_reviews'
    gh api repos/carlosmsanchezm/$repo/vulnerability-alerts -I 2>&1 | head -1
done
```

### 8.3 Check for Open Vulnerabilities

```bash
for repo in aegis-platform aegis-ui sovran; do
    echo "=== $repo ==="
    gh api repos/carlosmsanchezm/$repo/dependabot/alerts --jq '[.[] | select(.state == "open")] | length'
done
```

### 8.4 Generate Evidence for Specific Control

See `scripts/compliance/` for automated scripts.

---

## 9. File Quick Reference

| Need | File |
|------|------|
| Current status | `docs/compliance/soc2/STATUS.md` |
| Audit scope | `docs/compliance/soc2/00-scope-and-systems.md` |
| Control details | `docs/compliance/soc2/01-control-narratives.md` |
| Control mapping | `docs/compliance/soc2/control-matrix.csv` |
| Policy list | `docs/compliance/soc2/policy-register.md` |
| Onboarding | `docs/compliance/soc2/procedures/employee-onboarding-checklist.md` |
| Offboarding | `docs/compliance/soc2/procedures/employee-offboarding-checklist.md` |
| Access review | `docs/compliance/soc2/procedures/quarterly-access-review.md` |
| Monthly tasks | `docs/compliance/soc2/evidence-vault-setup/MONTHLY_CHECKLIST.md` |
| Retroactive evidence | `docs/compliance/soc2/retroactive-evidence-summary.md` |
| ISO 27001 status | `docs/compliance/iso27001/STATUS.md` |

---

## 10. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez + Claude | Initial agent context document |

---

## 11. For AI Agents: Pre-Flight Checklist

Before doing any SOC 2 compliance work, verify:

- [ ] Read this document completely
- [ ] Check `docs/compliance/soc2/STATUS.md` for current state
- [ ] Confirm which repos are in scope (currently 3)
- [ ] Understand the task context (monthly, quarterly, incident, etc.)
- [ ] Know where evidence should be stored
- [ ] Plan to update STATUS.md after completing work

**You are now ready to assist with SOC 2 Type II compliance work.**
