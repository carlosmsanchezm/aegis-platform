# Aegis Compliance Documentation

**Last Updated:** 2026-01-17
**Compliance Status:** ✅ SOC 2 Audit Ready | 🟡 ISO 27001 Documentation Complete (ISMS Not Yet Operating)

This directory contains all compliance-related documentation for Aegis Platform, supporting SOC 2 Type II and ISO 27001 certifications.

---

## Compliance Status Dashboard

| Certification | Status | Evidence Period | Target Date |
|--------------|--------|-----------------|-------------|
| **SOC 2 Type II** | ✅ Audit Ready | Sept 2025 - Present (4+ months) | Q2 2026 |
| **ISO 27001:2022** | 🟡 Docs Complete | Using SOC 2 evidence (Internal audit pending) | Q2 2026 |

### Key Metrics

| Metric | Value |
|--------|-------|
| Policies Approved | 7 |
| SOC 2 Controls Implemented | 30 of 30 (Security TSC) |
| ISO 27001 Controls Implemented | 68 of 81 applicable (84%) |
| Open Vulnerabilities | 0 |
| Retroactive Evidence | 4+ months |

---

## Directory Structure

```
docs/compliance/
├── README.md                           # This file
├── control-map.md                      # NIST 800-53 control mappings
│
├── soc2/                               # SOC 2 Type II
│   ├── STATUS.md                       # ✅ Current status
│   ├── AGENT_CONTEXT.md                # AI agent context document
│   ├── 00-scope-and-systems.md         # Audit scope
│   ├── 01-control-narratives.md        # Control descriptions
│   ├── control-matrix.csv              # Control mapping
│   ├── policy-register.md              # Policy tracking
│   ├── evidence-map.md                 # Evidence location mapping
│   ├── retroactive-evidence-summary.md # Backdated evidence
│   ├── policies/                       # 6 approved policies
│   ├── procedures/                     # Onboarding, offboarding, access review
│   └── evidence-vault-setup/           # Vault configuration
│
├── iso27001/                           # ISO 27001:2022
│   ├── STATUS.md                       # 🟡 Current status (ISMS NOT YET OPERATING)
│   ├── REUSE_PLAN.md                   # SOC 2 to ISO reuse mapping
│   ├── 00-isms-scope-and-context.md    # ISMS scope (Clause 4)
│   ├── 01-leadership-policy-roles.md   # Leadership & roles (Clause 5)
│   ├── 01-information-security-policy.md # Top-level policy (Clause 5.2)
│   ├── 02-risk-methodology.md          # Risk process (Clause 6.1.2)
│   ├── 03-risk-register.csv            # Risk inventory (Clause 8.2)
│   ├── 04-risk-treatment-plan.csv      # Treatment plan (Clause 8.3)
│   ├── 05-statement-of-applicability.csv # SoA - 93 controls (Clause 6.1.3d)
│   ├── 07-internal-audit-program.md    # Audit program (Clause 9.2)
│   ├── internal-audit-report-template.md # Audit report template
│   ├── 08-management-review-template.md # Review template (Clause 9.3)
│   ├── 09-corrective-actions-log.csv   # Corrective actions (Clause 10.1)
│   ├── 09-soc2-iso27001-mapping.md     # Control mapping
│   ├── soc2-crosswalk.csv              # SOC 2 ↔ ISO control crosswalk
│   └── policies/                       # ISO-specific policies
│
├── customer-docs/                      # Customer ATO documentation
│   ├── security-architecture-guide.md
│   ├── control-implementation-statements.md
│   ├── configuration-hardening-guide.md
│   ├── incident-response-runbook.md
│   └── customer-responsibility-matrix.md
│
├── sbom/                               # Software Bill of Materials
│   └── README.md
│
└── oscal/                              # OSCAL catalogs
    ├── fedramp_rev5_moderate_baseline.json
    └── nist_sp_800_53_rev5_catalog.json
```

---

## Evidence Vault (Separate Repository)

**Location:** `../aegis-compliance-evidence` (private repository)

```
aegis-compliance-evidence/
├── soc2/
│   └── 2025/
│       ├── retroactive/               # Sept-Jan evidence (both standards)
│       │   ├── git-commit-history.csv
│       │   ├── github-prs.json
│       │   ├── github-actions-runs.json
│       │   └── dependabot-alerts.json
│       └── 2025-01/                   # Monthly evidence
│           ├── access-reviews/
│           └── ci-cd-security/
└── iso27001/
    └── 2026/
        ├── internal-audits/
        ├── management-reviews/
        └── risk-assessments/
```

---

## SOC 2 + ISO 27001 Integration

These certifications share ~85% of controls. Our approach:

| Aspect | Strategy |
|--------|----------|
| **Policies** | SOC 2 policies apply to both (with ISO clause references) |
| **Evidence** | Single evidence vault serves both standards |
| **Audit** | Combined audit engagement recommended |
| **Maintenance** | Monthly evidence collection supports both |

### Control Overlap

| SOC 2 TSC | ISO 27001 Annex A | Overlap |
|-----------|-------------------|---------|
| CC1-CC5 (Environment) | A.5.1-5.8, Clauses 4-6 | 95% |
| CC6 (Access) | A.5.15-5.18, A.8.2-5 | 95% |
| CC7 (Operations) | A.5.24-28, A.8.8, A.8.15-16 | 90% |
| CC8 (Change) | A.8.32 | 100% |
| CC9 (Risk Mitigation) | A.5.19-23, A.5.29-30 | 90% |

---

## Quick Start Guides

### Monthly Evidence Collection (5 min)

```bash
# Run on 1st of each month
cd aegis-platform
MONTH=$(date +%Y-%m)

# GitHub export
GITHUB_OWNER=carlosmsanchezm GITHUB_REPOS=aegis-platform \
    ./scripts/compliance/export_github_security_baseline.sh \
    ../aegis-compliance-evidence/soc2/$(date +%Y)/${MONTH}/ci-cd-security/

# AWS export
AWS_PROFILE=myclaude \
    ./scripts/compliance/export_aws_security_baseline.sh \
    ../aegis-compliance-evidence/soc2/$(date +%Y)/${MONTH}/access-reviews/

# Commit
cd ../aegis-compliance-evidence
git add -A && git commit -m "Evidence: ${MONTH} collection" && git push
```

### Quarterly Access Review

See: `soc2/procedures/quarterly-access-review.md`

### Annual Policy Review

See: `soc2/policy-register.md`

---

## Certification Timeline

```
Jan 2026    Feb         Mar         Apr         May
    │         │           │           │           │
    ├─────────┼───────────┼───────────┼───────────┤
    │         │           │           │           │
    ▼         ▼           ▼           ▼           ▼
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
│ ISO     │ │ Internal│ │ ISO     │ │ ISO     │ │ Certs   │
│ Docs    │ │ Audit   │ │ Stage 1 │ │ Stage 2 │ │ Issued  │
│ Created │ │ + Mgmt  │ │ Audit   │ │ + SOC 2 │ │         │
│         │ │ Review  │ │         │ │ Audit   │ │         │
└─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘
    ✅         📋          📋          📋          🎯
```

---

## Service Model Context

| Deployment Model | Description | Compliance Needs |
|-----------------|-------------|------------------|
| **Model B** (Self-Hosted) | Customer deploys Aegis in their environment | ✅ Current - SOC 2 + ISO 27001 for SDLC |
| **Model C** (Managed) | Aegis operates for customers | Future - Add Availability TSC |

**Current Scope:** Model B only (Security TSC)

---

## Key Documents

### For Auditors
- [SOC 2 Status](./soc2/STATUS.md)
- [ISO 27001 Status](./iso27001/STATUS.md)
- [Control Matrix](./soc2/control-matrix.csv)
- [Statement of Applicability](./iso27001/02-statement-of-applicability.md)
- [Policy Register](./soc2/policy-register.md)

### For Customers
- [Security Architecture Guide](./customer-docs/security-architecture-guide.md)
- [Customer Responsibility Matrix](./customer-docs/customer-responsibility-matrix.md)
- [Control Implementation Statements](./customer-docs/control-implementation-statements.md)

### For Team
- [Monthly Checklist](./soc2/evidence-vault-setup/MONTHLY_CHECKLIST.md)
- [Quarterly Access Review](./soc2/procedures/quarterly-access-review.md)
- [Risk Register](./iso27001/04-risk-register.md)

---

## Contact

**Compliance Questions:** Carlos Sanchez (Founder)
**Email:** carlos@aegis.dev
**GitHub:** @carlosmsanchezm

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 2.0 | 2026-01-17 | Added ISO 27001 framework, updated status |
| 1.0 | 2025-01-17 | Initial SOC 2 structure |
