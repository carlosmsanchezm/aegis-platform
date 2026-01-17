# SOC 2 Evidence Collection Guide

**Version:** 1.0.0-DRAFT
**Last Updated:** [DATE]

## Overview

This guide describes the evidence that must be collected throughout the SOC 2 audit period (minimum 6 months) to demonstrate control effectiveness.

## Evidence Collection Schedule

### Continuous (Automated)

| Evidence | Source | Retention |
|----------|--------|-----------|
| Authentication logs | Keycloak, Platform API | 90+ days |
| Authorization decisions | Platform API | 90+ days |
| Change logs | GitHub, ArgoCD | Indefinite |
| Deployment records | GitHub Actions | 90+ days |
| Vulnerability scans | Trivy, Grype | 90+ days |

### Weekly

| Evidence | Collector | Storage |
|----------|-----------|---------|
| Infrastructure scan results | Security | S3/GDrive |
| Failed login summary | Automated | Dashboard |

### Monthly

| Evidence | Collector | Storage |
|----------|-----------|---------|
| Access review completion | Security | Ticketing system |
| Patch status report | Engineering | Dashboard |
| Incident summary | Security | Incident tracker |
| Vendor status | Security | Vendor inventory |

### Quarterly

| Evidence | Collector | Storage |
|----------|-----------|---------|
| User access certifications | Managers | GDrive |
| Privileged access review | Security | GDrive |
| Risk register update | Security | GDrive |
| Policy reviews | Security | GDrive |

### Annually

| Evidence | Collector | Storage |
|----------|-----------|---------|
| Risk assessment | Security | GDrive |
| Penetration test results | Third party | GDrive |
| DR test results | Engineering | GDrive |
| Security training records | HR | HRIS |
| Policy approvals | Executive | GDrive |

## Evidence by Trust Services Criteria

### CC1: Control Environment

| Evidence | Description |
|----------|-------------|
| Org chart | Current organizational structure |
| Job descriptions | Roles with security responsibilities |
| Code of conduct | Signed acknowledgments |
| Security policy | Approved policy document |
| Training records | Security awareness completion |

### CC2: Communication and Information

| Evidence | Description |
|----------|-------------|
| Security policy | Distribution evidence |
| Onboarding checklist | New hire security training |
| Customer notifications | Security incident notifications |
| Security page | Public security information |

### CC3: Risk Assessment

| Evidence | Description |
|----------|-------------|
| Risk assessment | Annual risk assessment document |
| Risk register | Current risk inventory |
| Treatment plans | Documented remediation |
| Risk reports | Quarterly management reports |

### CC4: Monitoring Activities

| Evidence | Description |
|----------|-------------|
| Audit findings | Internal/external audit results |
| Remediation tracking | Finding closure evidence |
| Control testing | Periodic control validation |

### CC5: Control Activities

| Evidence | Description |
|----------|-------------|
| Control matrix | Controls mapped to risks |
| Operating procedures | Documented procedures |
| Control testing | Effectiveness testing |

### CC6: Logical and Physical Access

| Evidence | Description |
|----------|-------------|
| User access list | Current user inventory |
| Access reviews | Quarterly certification |
| Termination tickets | Access removal evidence |
| MFA configuration | MFA enforcement settings |
| Password policy | Policy configuration |
| Login failures | Failed authentication reports |

### CC7: System Operations

| Evidence | Description |
|----------|-------------|
| Vulnerability scans | Scan reports |
| Patch records | Patching evidence |
| Incident tickets | Incident handling records |
| Monitoring alerts | Alert configurations |
| Change tickets | Change management records |

### CC8: Change Management

| Evidence | Description |
|----------|-------------|
| Change policy | Approved policy |
| Pull requests | Code review evidence |
| Deployment records | Release history |
| Approval records | CAB or lead approvals |
| Rollback records | Failed deployment handling |

### CC9: Risk Mitigation

| Evidence | Description |
|----------|-------------|
| BC/DR plan | Current plan document |
| Backup records | Backup completion logs |
| Recovery tests | DR test results |
| Vendor assessments | Third-party reviews |
| Insurance certificates | Cyber insurance |

## Evidence Collection Automation

### Log Aggregation

```yaml
# Fluent Bit configuration for evidence collection
fluent-bit:
  outputs:
    # Long-term evidence storage
    - name: s3
      match: audit.*
      bucket: aegis-soc2-evidence
      region: us-west-2
      s3_key_format: /audit/%Y/%m/%d/$TAG
      total_file_size: 100M

    # SIEM for real-time monitoring
    - name: splunk
      match: "*"
      host: splunk.internal
```

### Automated Reports

```bash
#!/bin/bash
# monthly-evidence-report.sh

# Generate access list
kubectl get secrets -n aegis-system \
  -o json | jq '.items[].metadata.name' \
  > /evidence/$(date +%Y-%m)/service-accounts.txt

# Generate deployment history
gh run list --repo aegis/platform \
  --json conclusion,createdAt,headSha \
  --limit 100 \
  > /evidence/$(date +%Y-%m)/deployments.json

# Generate vulnerability summary
grype sbom:platform-api-sbom.json \
  -o json > /evidence/$(date +%Y-%m)/vulnerabilities.json
```

## Evidence Storage

### Structure

```
evidence/
├── 2025/
│   ├── Q1/
│   │   ├── access-reviews/
│   │   ├── risk-register/
│   │   └── policies/
│   ├── Q2/
│   └── ...
├── continuous/
│   ├── audit-logs/
│   ├── change-logs/
│   └── scan-results/
└── annual/
    ├── penetration-test/
    ├── risk-assessment/
    └── dr-test/
```

### Naming Convention

```
[YYYY-MM-DD]_[EVIDENCE-TYPE]_[DESCRIPTION].[EXT]

Examples:
2025-01-15_access-review_q1-user-access.pdf
2025-01-15_vuln-scan_platform-api.json
2025-01-15_change-log_release-v1.2.0.json
```

### Retention Requirements

| Evidence Type | Minimum Retention |
|--------------|-------------------|
| Audit logs | 1 year |
| Access reviews | 3 years |
| Policy documents | Current + 1 version |
| Incident records | 3 years |
| Risk assessments | 3 years |
| Contracts | Contract term + 7 years |

## Evidence Quality Checklist

For each piece of evidence:

- [ ] Dated (timestamp or date range clear)
- [ ] Complete (full period covered)
- [ ] Authentic (from authoritative source)
- [ ] Accurate (reflects actual state)
- [ ] Attributable (who collected/approved)
- [ ] Relevant (addresses specific control)

## Common Evidence Issues

### Gaps

- Missing periods (vacation, system outage)
- Inconsistent collection frequency
- Lost evidence due to retention policy

**Mitigation**: Automate collection, use redundant storage

### Quality

- Screenshots without dates
- Partial reports
- Evidence from wrong period

**Mitigation**: Standardize collection procedures, use templates

### Availability

- Can't locate evidence during audit
- Evidence in inaccessible format
- Permissions issues

**Mitigation**: Centralized repository, regular access testing

## Audit Preparation

### 30 Days Before

- [ ] Verify all evidence collected for period
- [ ] Organize evidence by TSC
- [ ] Create evidence index
- [ ] Pre-review for gaps

### 1 Week Before

- [ ] Provide auditor access to repository
- [ ] Brief key personnel
- [ ] Prepare interview schedules
- [ ] Test auditor access

### During Audit

- [ ] Designate evidence coordinator
- [ ] Track evidence requests
- [ ] Provide requested items within 24 hours
- [ ] Document any clarifications

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0-DRAFT | TBD | TBD | Initial draft |
