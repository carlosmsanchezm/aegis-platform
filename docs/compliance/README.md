# Aegis Compliance Documentation

This directory contains all compliance-related documentation for Aegis Platform, organized to support SOC 2 Type II certification and customer ATO (Authority to Operate) requirements.

## Directory Structure

```
docs/compliance/
├── README.md                          # This file
├── control-map.md                     # NIST 800-53 control mappings (existing)
│
├── soc2/                              # SOC 2 Type II certification materials
│   ├── README.md                      # SOC 2 overview and roadmap
│   ├── trust-services-criteria.md     # TSC mapping and implementation
│   ├── policies/                      # Company security policies
│   │   ├── information-security-policy.md
│   │   ├── access-control-policy.md
│   │   ├── change-management-policy.md
│   │   ├── incident-response-policy.md
│   │   ├── risk-management-policy.md
│   │   └── vendor-management-policy.md
│   ├── procedures/                    # Operational procedures
│   │   ├── access-review-procedure.md
│   │   ├── change-control-procedure.md
│   │   └── incident-response-procedure.md
│   └── evidence/                      # Evidence collection templates
│       └── evidence-collection-guide.md
│
├── customer-docs/                     # Documentation for customer ATOs
│   ├── README.md                      # Overview for customers
│   ├── security-architecture-guide.md # System architecture for SSP
│   ├── control-implementation-statements.md  # CIS for each control
│   ├── configuration-hardening-guide.md      # Secure deployment guide
│   ├── incident-response-runbook.md          # IR procedures
│   └── customer-responsibility-matrix.md     # Shared responsibility model
│
├── sbom/                              # Software Bill of Materials
│   ├── README.md                      # SBOM generation guide
│   ├── generate-sbom.sh               # Automation script
│   └── sbom-template.json             # CycloneDX template
│
└── oscal/                             # OSCAL catalogs (existing)
    ├── README.md
    ├── fedramp_rev5_moderate_baseline.json
    ├── nist_sp_800_53_rev5_catalog.json
    └── nist_sp_800_171_rev3_catalog.json
```

## Compliance Roadmap (Per SOC 2 Report)

### Phase 1: Software Compliance (Q1)
- [ ] FIPS crypto configuration
- [ ] Session SUSPENDED state implementation
- [ ] Workload garbage collection
- [ ] Structured audit logging
- [ ] Network policies / tenant isolation

### Phase 2: PKI & Documentation (Q2)
- [ ] step-ca + cert-manager integration
- [ ] MFA enforcement option
- [ ] Control Implementation Statements
- [ ] Security Architecture Guide

### Phase 3: SOC 2 Certification (Q3-Q4)
- [ ] SOC 2 readiness assessment
- [ ] Policy implementation
- [ ] Evidence collection (6 months minimum)
- [ ] SOC 2 Type II audit

## Quick Links

- [SOC 2 Trust Services Criteria](./soc2/trust-services-criteria.md)
- [Security Architecture Guide](./customer-docs/security-architecture-guide.md)
- [Customer Responsibility Matrix](./customer-docs/customer-responsibility-matrix.md)
- [SBOM Generation](./sbom/README.md)

## Model B vs Model C

| Deployment Model | Description | Customer Needs from You |
|-----------------|-------------|------------------------|
| **Model B** (Self-Hosted) | Customer deploys Aegis in their environment | Compliant software, documentation for their ATO |
| **Model C** (Managed) | You deploy/manage Aegis in customer's account | Same as B + SOC 2 Type II required, possibly clearances |

## Contact

For compliance questions, contact: [compliance@aegis.io]
