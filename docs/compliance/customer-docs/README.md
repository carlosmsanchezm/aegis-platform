# Customer Compliance Documentation

## Overview

This directory contains documentation that customers need to include Aegis in their Authority to Operate (ATO) package. Whether your customer is pursuing FedRAMP, DoD IL-4/5, or CMMC certification, these documents provide the evidence they need.

## Document Inventory

| Document | Purpose | Customer Uses In |
|----------|---------|-----------------|
| [Security Architecture Guide](./security-architecture-guide.md) | System architecture, data flows, boundaries | SSP Section 9 |
| [Control Implementation Statements](./control-implementation-statements.md) | How Aegis implements each control | SSP Section 13 |
| [Configuration Hardening Guide](./configuration-hardening-guide.md) | Secure deployment instructions | Implementation, CM baseline |
| [Incident Response Runbook](./incident-response-runbook.md) | Security event handling | IR Plan, POA&M |
| [Customer Responsibility Matrix](./customer-responsibility-matrix.md) | Who does what | SSP, shared responsibility |

## How Customers Use These Documents

### FedRAMP Authorization

```
Customer's SSP (System Security Plan)
├── Section 9: System Description
│   └── Uses: Security Architecture Guide
├── Section 13: Control Implementation
│   └── Uses: Control Implementation Statements
├── Appendix: CM Baseline
│   └── Uses: Configuration Hardening Guide
└── Attachments
    └── Customer Responsibility Matrix
```

### DoD IL-4/IL-5 Authorization

Same as FedRAMP, plus:
- STIG compliance documentation
- CAC/PIV integration evidence
- FIPS 140-2/3 validation evidence

### CMMC Level 2

```
CMMC Assessment Package
├── System Security Plan
│   └── Uses: Control Implementation Statements (mapped to NIST 800-171)
├── POA&M
│   └── Uses: Gap analysis from CIS
└── Evidence
    └── Uses: Configuration screenshots, audit logs
```

## Model B vs Model C Differences

### Model B (Self-Hosted)

Customer deploys Aegis themselves. They need:
- Technical documentation for their engineers
- Evidence that software CAN be compliant
- Configuration guidance

### Model C (Managed in Customer Account)

You deploy/manage Aegis in their account. They need:
- Everything from Model B
- Your SOC 2 Type II report
- Personnel security documentation
- Operational procedures you follow
- MSP agreement terms

## Document Maintenance

| Document | Review Frequency | Owner |
|----------|-----------------|-------|
| Security Architecture Guide | Each major release | Engineering |
| Control Implementation Statements | Quarterly | Security |
| Configuration Hardening Guide | Each release | Engineering |
| Incident Response Runbook | Annually | Security |
| Customer Responsibility Matrix | Annually | Security |

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | TBD | Initial release |
