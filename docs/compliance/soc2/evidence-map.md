# SOC 2 & ISO 27001 Evidence Map

**Document ID:** COMPLIANCE-EVIDENCE-MAP-001
**Version:** 2.0
**Updated:** 2026-01-17

---

## IMPORTANT: Directory Structure Clarification

### What Goes Where

| Location | Purpose | Contents |
|----------|---------|----------|
| `docs/compliance/soc2/` | **Documentation** (in repo) | Policies, procedures, control narratives, STATUS.md |
| `docs/compliance/iso27001/` | **Documentation** (in repo) | ISMS docs, risk register, SoA, audit templates |
| `~/code/aegis-compliance-evidence/` | **Evidence Vault** (separate) | Actual exports, logs, screenshots, audit reports |

### Why Separate?

| Documentation (in main repo) | Evidence (separate vault) |
|------------------------------|---------------------------|
| ✅ Can be shared publicly | ❌ Contains sensitive data |
| ✅ Describes what controls exist | ❌ Contains actual IAM users/ARNs |
| ✅ Templates and procedures | ❌ Contains vulnerability details |
| ✅ Version controlled in product repo | ❌ May contain PII |

### ⚠️ Evidence Should NOT Be in Main Repo

If you have evidence in `aegis-platform.../compliance-evidence/`, run the migration script:
```bash
./scripts/compliance/migrate-evidence-vault.sh
```

---

## Evidence Vault Structure

The evidence vault lives at: `~/code/aegis-compliance-evidence/`

```
~/code/aegis-compliance-evidence/           # PRIVATE - separate from main repo
│
├── README.md                               # Vault documentation
│
├── soc2/                                   # SOC 2 Type II evidence
│   ├── 2025/
│   │   ├── retroactive/                    # Sept 2025 - Jan 2026 historical
│   │   │   ├── git-commit-history.csv
│   │   │   ├── github-prs.json
│   │   │   ├── github-actions-runs.json
│   │   │   └── dependabot-alerts-*.json
│   │   ├── 2025-09/ through 2025-12/       # Monthly evidence
│   │   ├── Q4/                             # Q4 2025 access review
│   │   ├── governance/                     # Annual/static documents
│   │   ├── risk-management/                # Risk assessments
│   │   └── vendor-soc2-reports/            # AWS, GitHub SOC 2 reports
│   │
│   └── 2026/
│       ├── 2026-01/                        # January 2026
│       │   ├── access-reviews/             # IAM exports, MFA status
│       │   ├── ci-cd-security/             # Branch protection, workflows
│       │   ├── vuln-management/            # Dependabot alerts
│       │   ├── incident-response/          # Incident logs
│       │   ├── logging-monitoring/         # CloudTrail status
│       │   ├── change-management/          # PR exports
│       │   ├── vendor-management/          # Vendor status
│       │   ├── backups-dr/                 # Backup verification
│       │   └── training-policy-ack/        # Training records
│       ├── 2026-02/ ... 2026-12/
│       ├── Q1/, Q2/, Q3/, Q4/              # Quarterly access reviews
│       ├── governance/
│       ├── risk-management/
│       └── vendor-soc2-reports/
│
└── iso27001/                               # ISO 27001-specific evidence
    └── 2026/
        ├── internal-audits/                # Clause 9.2 - audit reports
        ├── management-reviews/             # Clause 9.3 - review minutes
        ├── risk-assessments/               # Clause 8.2 - risk assessment records
        └── corrective-actions/             # Clause 10.1 - CA evidence
```

---

## ISO 27001 Evidence Sharing

**~85% of SOC 2 evidence is directly reusable for ISO 27001.**

| SOC 2 Evidence Folder | ISO 27001 Controls | Reuse |
|-----------------------|-------------------|-------|
| access-reviews/ | A.5.15-18, A.8.2-5 | ✅ Direct |
| ci-cd-security/ | A.8.4, A.8.25, A.8.32 | ✅ Direct |
| vuln-management/ | A.8.8 | ✅ Direct |
| change-management/ | A.8.32 | ✅ Direct |
| incident-response/ | A.5.24-28 | ✅ Direct |
| training-policy-ack/ | A.6.3 | ✅ Direct |
| vendor-management/ | A.5.19-23 | ✅ Direct |

**ISO-specific evidence** (not covered by SOC 2):
- Internal audit reports → `iso27001/YYYY/internal-audits/`
- Management review minutes → `iso27001/YYYY/management-reviews/`
- Corrective action records → `iso27001/YYYY/corrective-actions/`

See: `docs/compliance/iso27001/soc2-crosswalk.csv` for complete mapping.

---

## Evidence Vault Setup

### Option 1: Local Directory (Simplest)

```bash
# Create the vault
mkdir -p ~/code/aegis-compliance-evidence

# Run the migration script to move existing evidence
cd ~/code/aegis-platform-observability-integration
./scripts/compliance/migrate-evidence-vault.sh

# Add vault to shell for easy access
echo 'export EVIDENCE_VAULT="$HOME/code/aegis-compliance-evidence"' >> ~/.zshrc
```

### Option 2: Private GitHub Repository

```bash
# Create private repo
gh repo create aegis-compliance-evidence --private --description "Compliance Evidence Vault"

# Clone
git clone git@github.com:carlosmsanchezm/aegis-compliance-evidence.git ~/code/aegis-compliance-evidence

# Run migration
cd ~/code/aegis-platform-observability-integration
./scripts/compliance/migrate-evidence-vault.sh
```

### Option 3: Encrypted Disk Image (macOS)

```bash
# Create encrypted sparse image
hdiutil create -size 2g -encryption AES-256 -type SPARSE \
  -fs APFS -volname "aegis-compliance-evidence" ~/aegis-compliance-evidence.sparseimage

# Mount when needed
hdiutil attach ~/aegis-compliance-evidence.sparseimage

# Vault will be at /Volumes/aegis-compliance-evidence/
```

---

## Control → Evidence Mapping

### CC1: Control Environment

| Control | Evidence Required | Vault Path | Frequency |
|---------|------------------|------------|-----------|
| CC1.1 | Code of conduct acknowledgments | `soc2/YYYY/governance/` | Annual |
| CC1.2 | Security meeting minutes | `soc2/YYYY/YYYY-MM/governance/` | Quarterly |
| CC1.3 | Org chart with security roles | `soc2/YYYY/governance/` | Annual |
| CC1.4 | Training completion records | `soc2/YYYY/training-policy-ack/` | Annual |
| CC1.5 | Security in performance reviews | `soc2/YYYY/governance/` | Annual |

### CC2: Communication

| Control | Evidence Required | Vault Path | Frequency |
|---------|------------------|------------|-----------|
| CC2.1 | Security announcements | `soc2/YYYY/YYYY-MM/communication/` | As needed |
| CC2.2 | Customer security notifications | `soc2/YYYY/YYYY-MM/communication/` | As needed |
| CC2.3 | Security finding communications | `soc2/YYYY/YYYY-MM/communication/` | As needed |

### CC3: Risk Assessment

| Control | Evidence Required | Vault Path | Frequency |
|---------|------------------|------------|-----------|
| CC3.1 | Security objectives doc | `soc2/YYYY/risk-management/` | Annual |
| CC3.2 | Risk register | `soc2/YYYY/risk-management/` | Quarterly |
| CC3.3 | Fraud risk assessment | `soc2/YYYY/risk-management/` | Annual |
| CC3.4 | Change impact assessments | `soc2/YYYY/YYYY-MM/change-management/` | Per change |

### CC4: Monitoring

| Control | Evidence Required | Vault Path | Frequency |
|---------|------------------|------------|-----------|
| CC4.1 | Control testing records | `soc2/YYYY/YYYY-MM/governance/` | Annual |
| CC4.2 | Finding remediation records | `soc2/YYYY/YYYY-MM/vuln-management/` | Ongoing |

### CC5: Control Activities

| Control | Evidence Required | Vault Path | Frequency |
|---------|------------------|------------|-----------|
| CC5.1 | Control matrix | **In repo:** `docs/compliance/soc2/control-matrix.csv` | Annual |
| CC5.2 | Technical control configs | `soc2/YYYY/YYYY-MM/ci-cd-security/` | Monthly |
| CC5.3 | Policy documents | **In repo:** `docs/compliance/soc2/policies/` | Annual |

### CC6: Logical Access

| Control | Evidence Required | Vault Path | Frequency |
|---------|------------------|------------|-----------|
| CC6.1 | Access control configurations | `soc2/YYYY/YYYY-MM/access-reviews/` | Monthly |
| CC6.2 | Provisioning tickets | `soc2/YYYY/YYYY-MM/access-reviews/` | Per event |
| CC6.3 | Termination records | `soc2/YYYY/YYYY-MM/access-reviews/` | Per event |
| CC6.4 | Access review certifications | `soc2/YYYY/QX/access-reviews/` | Quarterly |
| CC6.5 | Physical access (inherited) | `soc2/YYYY/vendor-soc2-reports/` | Annual |
| CC6.6 | MFA configuration | `soc2/YYYY/YYYY-MM/access-reviews/` | Monthly |
| CC6.7 | RBAC policies | `soc2/YYYY/YYYY-MM/access-reviews/` | Monthly |
| CC6.8 | Malware protection | `soc2/YYYY/YYYY-MM/vuln-management/` | Weekly |

### CC7: System Operations

| Control | Evidence Required | Vault Path | Frequency |
|---------|------------------|------------|-----------|
| CC7.1 | Vulnerability scan reports | `soc2/YYYY/YYYY-MM/vuln-management/` | Weekly |
| CC7.2 | Monitoring dashboards | `soc2/YYYY/YYYY-MM/logging-monitoring/` | Monthly |
| CC7.3 | Alert investigation records | `soc2/YYYY/YYYY-MM/incident-response/` | Per event |
| CC7.4 | Incident response records | `soc2/YYYY/YYYY-MM/incident-response/` | Per event |
| CC7.5 | Recovery test results | `soc2/YYYY/backups-dr/` | Annual |

### CC8: Change Management

| Control | Evidence Required | Vault Path | Frequency |
|---------|------------------|------------|-----------|
| CC8.1 | Change records (PRs) | `soc2/YYYY/YYYY-MM/change-management/` | Monthly |

### CC9: Risk Mitigation

| Control | Evidence Required | Vault Path | Frequency |
|---------|------------------|------------|-----------|
| CC9.1 | Vendor assessments | `soc2/YYYY/vendor-management/` | Annual |
| CC9.2 | BC/DR plans | `soc2/YYYY/backups-dr/` | Annual |

---

## Automation Scripts

| Script | Purpose | Output Location |
|--------|---------|-----------------|
| `scripts/compliance/export_github_security_baseline.sh` | GitHub security settings | `$EVIDENCE_VAULT/soc2/YYYY/YYYY-MM/` |
| `scripts/compliance/export_aws_security_baseline.sh` | AWS IAM, CloudTrail, etc. | `$EVIDENCE_VAULT/soc2/YYYY/YYYY-MM/access-reviews/` |
| `scripts/compliance/export_ci_reports.sh` | SBOM, vuln scans | `$EVIDENCE_VAULT/soc2/YYYY/YYYY-MM/vuln-management/` |
| `scripts/compliance/migrate-evidence-vault.sh` | Move evidence to correct location | One-time migration |

All scripts respect the `$EVIDENCE_VAULT` environment variable.

---

## Collection Schedule

See: `docs/compliance/iso27001/10-isms-operating-plan.md` for detailed schedule.

### Quick Reference

| Frequency | Tasks |
|-----------|-------|
| **Weekly** | Dependabot review, security alerts |
| **Monthly** | Run all export scripts, incident log |
| **Quarterly** | Access review, risk register review |
| **Semi-annual** | Internal audit, management review |
| **Annual** | Policy review, vendor assessments, DR test |

---

## Sensitive Evidence Handling

### ✅ DO Include in Vault
- Configuration exports (sanitized)
- Access lists and role assignments
- Meeting notes and approvals
- Scan summaries
- Policy acknowledgments

### ❌ DO NOT Include (or Encrypt)
- Actual credentials or API keys
- Penetration test raw findings
- Detailed vulnerability exploit information
- Customer-specific data
- Personally identifiable information (PII)

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 2.0 | 2026-01-17 | Carlos Sanchez | Clarified structure; added migration script reference |
| 1.0 | 2025-01-17 | Claude (SOC2 Engineer) | Initial evidence map |
