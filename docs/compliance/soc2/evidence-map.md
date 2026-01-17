# SOC 2 Evidence Map

**Document ID:** SOC2-EVIDENCE-MAP-001
**Version:** 1.0
**Created:** 2025-01-17

## Overview

This document maps each SOC 2 control to its evidence location in the evidence vault. Evidence is stored in a **private location** outside the main repository to protect sensitive information.

## Evidence Vault Structure

```
compliance-evidence/                    # PRIVATE - not in main repo
├── soc2/
│   ├── 2025/
│   │   ├── Q1/                        # Quarterly evidence
│   │   │   └── access-reviews/
│   │   ├── 2025-01/                   # Monthly evidence
│   │   │   ├── access-reviews/
│   │   │   ├── change-management/
│   │   │   ├── ci-cd-security/
│   │   │   ├── vuln-management/
│   │   │   ├── logging-monitoring/
│   │   │   ├── incident-response/
│   │   │   ├── vendor-management/
│   │   │   ├── backups-dr/
│   │   │   └── training-policy-ack/
│   │   ├── 2025-02/
│   │   │   └── ...
│   │   ├── governance/                # Annual/static documents
│   │   ├── risk-management/
│   │   └── vendor-soc2-reports/       # Third-party reports
│   └── templates/                     # Reusable templates
└── README.md                          # Vault documentation
```

## Evidence Vault Setup

### Option 1: Local Encrypted Folder (Recommended for Solo/Small Team)

```bash
# Create encrypted disk image (macOS)
hdiutil create -size 1g -encryption AES-256 -type SPARSE \
  -fs APFS -volname "compliance-evidence" ~/compliance-evidence.sparseimage

# Mount when needed
hdiutil attach ~/compliance-evidence.sparseimage

# Create structure
mkdir -p /Volumes/compliance-evidence/soc2/2025/{Q1,governance,risk-management}
mkdir -p /Volumes/compliance-evidence/soc2/2025/$(date +%Y-%m)/{access-reviews,change-management,ci-cd-security,vuln-management,logging-monitoring,incident-response,vendor-management,backups-dr,training-policy-ack}
```

### Option 2: Private GitHub Repository

```bash
# Create private repo
gh repo create aegis-compliance-evidence --private --description "SOC 2 Evidence Vault"

# Clone and setup
git clone git@github.com:carlosmsanchezm/aegis-compliance-evidence.git
cd aegis-compliance-evidence

# Create structure
mkdir -p soc2/2025/{Q1,governance,risk-management}
mkdir -p soc2/2025/$(date +%Y-%m)/{access-reviews,change-management,ci-cd-security,vuln-management,logging-monitoring,incident-response,vendor-management,backups-dr,training-policy-ack}

# Add .gitignore for sensitive files
echo "*.pem" >> .gitignore
echo "*.key" >> .gitignore
echo "*-secrets.json" >> .gitignore
```

---

## Control → Evidence Mapping

### CC1: Control Environment

| Control | Evidence Required | Vault Path | Collection Method | Frequency |
|---------|------------------|------------|-------------------|-----------|
| CC1.1 | Code of conduct acknowledgments | `governance/code-of-conduct-acks/` | Signed PDFs or digital acks | Annual |
| CC1.2 | Security meeting minutes | `YYYY-MM/governance/` | Meeting notes export | Quarterly |
| CC1.3 | Org chart with security roles | `governance/` | Org chart PDF | Annual |
| CC1.4 | Training completion records | `training-policy-ack/` | LMS export or tracker | Annual |
| CC1.5 | Security in performance reviews | `governance/` | Policy document | Annual |

### CC2: Communication

| Control | Evidence Required | Vault Path | Collection Method | Frequency |
|---------|------------------|------------|-------------------|-----------|
| CC2.1 | Security announcements | `YYYY-MM/communication/` | Email/Slack exports | As needed |
| CC2.2 | Customer security notifications | `YYYY-MM/communication/` | Email records | As needed |
| CC2.3 | Security finding communications | `YYYY-MM/communication/` | Jira/email exports | As needed |

### CC3: Risk Assessment

| Control | Evidence Required | Vault Path | Collection Method | Frequency |
|---------|------------------|------------|-------------------|-----------|
| CC3.1 | Security objectives doc | `risk-management/` | Document | Annual |
| CC3.2 | Risk register | `risk-management/` | Spreadsheet | Quarterly |
| CC3.3 | Fraud risk assessment | `risk-management/` | Document section | Annual |
| CC3.4 | Change impact assessments | `YYYY-MM/change-management/` | PR descriptions | Per change |

### CC4: Monitoring

| Control | Evidence Required | Vault Path | Collection Method | Frequency |
|---------|------------------|------------|-------------------|-----------|
| CC4.1 | Control testing records | `YYYY-MM/governance/` | Testing spreadsheet | Annual |
| CC4.2 | Finding remediation records | `YYYY-MM/vuln-management/` | Jira exports | Ongoing |

### CC5: Control Activities

| Control | Evidence Required | Vault Path | Collection Method | Frequency |
|---------|------------------|------------|-------------------|-----------|
| CC5.1 | Control matrix | `governance/` | This repo: `control-matrix.csv` | Annual |
| CC5.2 | Technical control configs | `YYYY-MM/ci-cd-security/` | `export_github_security_baseline.sh` | Monthly |
| CC5.3 | Policy documents | `governance/` | This repo: `soc2/policies/` | Annual |

### CC6: Logical Access

| Control | Evidence Required | Vault Path | Collection Method | Frequency |
|---------|------------------|------------|-------------------|-----------|
| CC6.1 | Access control configurations | `YYYY-MM/access-reviews/` | `export_github_security_baseline.sh`, `export_aws_security_baseline.sh` | Monthly |
| CC6.2 | Provisioning tickets | `YYYY-MM/access-reviews/` | Jira/ticket exports | Per event |
| CC6.3 | Termination records | `YYYY-MM/access-reviews/` | Offboarding tickets | Per event |
| CC6.4 | Access review certifications | `QX/access-reviews/` | Manager sign-offs | Quarterly |
| CC6.5 | Physical access (inherited) | `vendor-soc2-reports/` | AWS SOC 2 report | Annual |
| CC6.6 | MFA configuration | `YYYY-MM/access-reviews/` | `export_aws_security_baseline.sh` | Monthly |
| CC6.7 | RBAC policies | `YYYY-MM/access-reviews/` | IAM/K8s RBAC exports | Monthly |
| CC6.8 | Malware protection | `YYYY-MM/vuln-management/` | Container scan reports | Weekly |

### CC7: System Operations

| Control | Evidence Required | Vault Path | Collection Method | Frequency |
|---------|------------------|------------|-------------------|-----------|
| CC7.1 | Vulnerability scan reports | `YYYY-MM/vuln-management/` | `export_ci_reports.sh` | Weekly |
| CC7.2 | Monitoring dashboards | `YYYY-MM/logging-monitoring/` | Dashboard screenshots | Monthly |
| CC7.3 | Alert investigation records | `YYYY-MM/incident-response/` | Alert triage tickets | Per event |
| CC7.4 | Incident response records | `YYYY-MM/incident-response/` | IR tickets | Per event |
| CC7.5 | Recovery test results | `backups-dr/` | DR test reports | Annual |

### CC8: Change Management

| Control | Evidence Required | Vault Path | Collection Method | Frequency |
|---------|------------------|------------|-------------------|-----------|
| CC8.1 | Change records (PRs) | `YYYY-MM/change-management/` | `export_github_security_baseline.sh` | Monthly |

### CC9: Risk Mitigation

| Control | Evidence Required | Vault Path | Collection Method | Frequency |
|---------|------------------|------------|-------------------|-----------|
| CC9.1 | Vendor assessments | `vendor-management/` | Vendor questionnaires | Annual |
| CC9.2 | BC/DR plans | `backups-dr/` | Plan documents | Annual |

### A1: Availability

| Control | Evidence Required | Vault Path | Collection Method | Frequency |
|---------|------------------|------------|-------------------|-----------|
| A1.1 | Capacity metrics | `YYYY-MM/logging-monitoring/` | Monitoring exports | Monthly |
| A1.2 | Environmental (inherited) | `vendor-soc2-reports/` | AWS SOC 2 report | Annual |
| A1.3 | Backup/recovery tests | `backups-dr/` | Test results | Annual |

---

## Automation Scripts

| Script | Purpose | Output Location | Frequency |
|--------|---------|-----------------|-----------|
| `scripts/compliance/export_github_security_baseline.sh` | GitHub security settings | `ci-cd-security/`, `access-reviews/` | Monthly |
| `scripts/compliance/export_aws_security_baseline.sh` | AWS IAM, CloudTrail, etc. | `access-reviews/` | Monthly |
| `scripts/compliance/export_ci_reports.sh` | SBOM, vuln scans | `vuln-management/` | Weekly |
| `docs/compliance/sbom/generate-sbom.sh` | Software Bill of Materials | `ci-cd-security/` | Per release |

---

## Collection Schedule

### Weekly
- [ ] Run `export_ci_reports.sh` for vulnerability scans
- [ ] Review critical/high vulnerabilities

### Monthly
- [ ] Run `export_github_security_baseline.sh`
- [ ] Run `export_aws_security_baseline.sh`
- [ ] Archive monitoring dashboards
- [ ] Review and close stale tickets

### Quarterly
- [ ] Conduct access review (all systems)
- [ ] Update risk register
- [ ] Review vendor status

### Annually
- [ ] Full risk assessment
- [ ] Policy reviews and approvals
- [ ] DR/BC test
- [ ] Security training refresh
- [ ] Collect vendor SOC 2 reports

---

## Sensitive Evidence Handling

### DO Include in Vault
- Configuration exports (sanitized)
- Access lists and role assignments
- Meeting notes and approvals
- Scan summaries (not raw findings with exploit details)
- Policy acknowledgments

### DO NOT Include in Vault (or Encrypt)
- Actual credentials or API keys
- Penetration test raw findings
- Detailed vulnerability exploit information
- Customer-specific data
- Personally identifiable information (PII)

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-01-17 | Claude (SOC2 Engineer) | Initial evidence map |
