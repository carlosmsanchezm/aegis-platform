# Aegis Customer Responsibility Matrix

**Version:** 2.0.0-DRAFT
**Last Updated:** 2026-03-09
**Classification:** Customer Confidential

## Overview

This Customer Responsibility Matrix (CRM) defines the security responsibilities shared between Aegis and the customer for each NIST 800-53 control. Use this document to understand what Aegis provides versus what you must implement.

## Legend

| Symbol | Meaning |
|--------|---------|
| YES-P | Fully provided by Aegis |
| CFG | Aegis provides, customer configures |
| SUPP | Customer implements, Aegis supports |
| CUST | Customer responsibility only |
| N/A | Not applicable |
| INH | Inherited from cloud service provider (AWS) |

## Responsibility by Deployment Model

| Model | Aegis Responsibility | Customer Responsibility |
|-------|---------------------|------------------------|
| **Model B (Self-Hosted)** | Software capabilities, documentation | Deployment, operation, authorization |
| **Model C (Managed)** | Software + operation in customer account | Account access, authorization |

## Access Control (AC)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Account Management | AC-2 | CFG Keycloak IdP, RBAC engine | SUPP Account lifecycle, reviews | CUST Account lifecycle, reviews |
| Access Enforcement | AC-3 | YES-P RBAC at API level | CFG Role assignments | CFG Role assignments |
| Information Flow | AC-4 | CFG Network policy templates | SUPP Policy deployment | SUPP Policy approval |
| Separation of Duties | AC-5 | CFG Role definitions | SUPP SoD matrix, enforcement | SUPP SoD matrix approval |
| Least Privilege | AC-6 | YES-P Predefined minimal roles | CFG Role assignments | CFG Role assignments |
| Unsuccessful Logins | AC-7 | CFG Keycloak lockout config | CFG Threshold settings | CFG Threshold settings |
| System Use Notification | AC-8 | CFG Banner capability | SUPP Banner content | SUPP Banner content |
| Session Lock | AC-11 | YES-P Configurable timeout | CFG Timeout value | CFG Timeout value |
| Session Termination | AC-12 | YES-P Auto-termination | CFG Timeout value | CFG Timeout value |
| Remote Access | AC-17 | YES-P Secure tunnel, tokens | CFG VPN policy | CFG VPN policy |
| Wireless Access | AC-18 | N/A | CUST Network controls | CUST Network controls |

## Awareness and Training (AT)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Security Training | AT-2 | SUPP Product training materials | CUST Security awareness program | CUST Security awareness program |
| Role-Based Training | AT-3 | SUPP Admin training docs | CUST Role-specific training | CUST Role-specific training |

## Audit and Accountability (AU)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Event Logging | AU-2 | YES-P Security events logged | CFG Additional events | CFG Additional events |
| Content of Records | AU-3 | YES-P All required fields | N/A | N/A |
| Audit Storage | AU-4 | CFG Local storage, export | SUPP SIEM storage | SUPP SIEM storage |
| Audit Failure Response | AU-5 | CFG Alerting on failure | SUPP Alert response | SUPP Alert response |
| Audit Review | AU-6 | CFG SIEM integration | CUST Review procedures | CUST Review procedures |
| Audit Reduction | AU-7 | CFG Structured JSON logs | SUPP Query/dashboards | SUPP Query/dashboards |
| Time Stamps | AU-8 | YES-P UTC timestamps | CFG NTP configuration | CFG NTP configuration |
| Audit Protection | AU-9 | CFG Log integrity options | SUPP Tamper protection | SUPP Tamper protection |
| Audit Retention | AU-11 | CFG Retention config | SUPP Retention policy | SUPP Retention policy |

## Configuration Management (CM)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Baseline Config | CM-2 | CFG Default Helm values | SUPP Hardened values | SUPP Hardened values approval |
| Config Change Control | CM-3 | YES-P GitOps, PR reviews | SUPP Change process | CUST Change approval |
| Security Impact Analysis | CM-4 | SUPP Release notes | CUST Impact analysis | CUST Impact analysis |
| Access Restrictions | CM-5 | CFG RBAC for config | SUPP Access controls | SUPP Access controls |
| Config Settings | CM-6 | CFG Hardening guide | SUPP Implementation | SUPP Verification |
| Least Functionality | CM-7 | YES-P Minimal containers | CFG Disable unused features | CFG Disable unused features |
| SW Inventory | CM-8 | YES-P SBOM generation | SUPP Inventory tracking | SUPP Inventory tracking |

## Contingency Planning (CP)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Contingency Plan | CP-2 | SUPP DR guidance | CUST BCP/DR plan | CUST BCP/DR plan |
| Contingency Training | CP-3 | SUPP Training materials | CUST Training program | CUST Training program |
| Contingency Testing | CP-4 | SUPP Test procedures | CUST Testing | CUST Testing |
| Backup | CP-9 | CFG Backup guidance | SUPP Backup implementation | SUPP Backup implementation |
| Recovery | CP-10 | CFG Recovery procedures | SUPP Recovery testing | SUPP Recovery testing |

## Identification and Authentication (IA)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| User I&A | IA-2 | YES-P OIDC/SAML authentication | CFG IdP configuration | CFG IdP configuration |
| MFA | IA-2(1) | YES-P MFA support | CFG MFA enforcement | CFG MFA enforcement |
| Phishing-Resistant MFA | IA-2(6) | YES-P FIDO2/WebAuthn support | CFG Authenticator config | CFG Authenticator config |
| PIV/CAC | IA-2(12) | CFG X.509 support | SUPP CAC integration | SUPP CAC integration |
| Device I&A | IA-3 | CFG mTLS for services | SUPP Device certificates | SUPP Device certificates |
| Identifier Management | IA-4 | CFG Keycloak ID management | SUPP ID lifecycle | SUPP ID lifecycle |
| Authenticator Mgmt | IA-5 | CFG Password policies | SUPP Policy enforcement | SUPP Policy enforcement |
| Cryptographic Auth | IA-7 | YES-P TLS, JWT signing | CFG Algorithm selection | CFG Algorithm selection |

## Incident Response (IR)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| IR Policy | IR-1 | SUPP IR runbook | CUST IR policy | CUST IR policy |
| IR Training | IR-2 | SUPP Training materials | CUST Training program | CUST Training program |
| IR Testing | IR-3 | SUPP Test scenarios | CUST Testing | CUST Testing |
| IR Handling | IR-4 | SUPP Runbook procedures | CUST IR execution | CUST IR execution |
| IR Monitoring | IR-5 | CFG Security alerts | SUPP Alert triage | SUPP Alert triage |
| IR Reporting | IR-6 | SUPP Report templates | CUST Reporting | CUST Reporting |

## Maintenance (MA)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Controlled Maintenance | MA-2 | SUPP Maintenance windows | SUPP Scheduling | SUPP Scheduling |
| Maintenance Tools | MA-3 | YES-P kubectl, Helm | CFG Tool access | CFG Tool access |
| Remote Maintenance | MA-4 | CFG Secure admin access | SUPP Access controls | SUPP Access controls |
| Maintenance Personnel | MA-5 | CFG RBAC for maintenance | SUPP Personnel authorization | SUPP Personnel authorization |
| Timely Maintenance | MA-6 | YES-P Security patches | SUPP Patch deployment | SUPP Patch approval |

## Physical and Environmental Protection (PE)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| PE Policy | PE-1 | INH AWS data centers | CUST Policy (on-prem) | CUST Policy (on-prem) |
| Physical Access Auth | PE-2 | INH AWS data centers | CUST On-prem access lists | CUST On-prem access lists |
| Physical Access Control | PE-3 | INH AWS data centers | CUST On-prem access control | CUST On-prem access control |
| Monitoring Physical Access | PE-6 | INH AWS data centers | CUST On-prem monitoring | CUST On-prem monitoring |
| Visitor Access | PE-7 | INH AWS data centers | CUST Visitor management | CUST Visitor management |
| Emergency Shutoff | PE-10 | INH AWS data centers | CUST On-prem power | CUST On-prem power |
| Environmental Protection | PE-13 | INH AWS data centers | CUST Fire protection | CUST Fire protection |
| Temperature/Humidity | PE-14 | INH AWS data centers | CUST Environmental controls | CUST Environmental controls |

## Personnel Security (PS)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Position Risk | PS-2 | N/A | CUST Risk designations | CUST Risk designations |
| Personnel Screening | PS-3 | N/A | CUST Background checks | CUST Background checks |
| Personnel Termination | PS-4 | CFG Account disable | CUST Termination process | CUST Termination process |
| Personnel Transfer | PS-5 | CFG Role changes | CUST Transfer process | CUST Transfer process |
| Access Agreements | PS-6 | N/A | CUST User agreements | CUST User agreements |
| Personnel Sanctions | PS-8 | N/A | CUST Sanctions process | CUST Sanctions process |

## Planning (PL)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Planning Policy | PL-1 | SUPP Compliance docs | CUST Security planning policy | CUST Security planning policy |
| System Security Plan | PL-2 | SUPP SSP inputs (CIS, CRM, SAG) | CUST SSP development | CUST SSP development |
| Rules of Behavior | PL-4 | SUPP Template provided | CUST RoB development | CUST RoB development |

## Risk Assessment (RA)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Risk Assessment | RA-3 | SUPP Security assessment | CUST System risk assessment | CUST System risk assessment |
| Vulnerability Scanning | RA-5 | YES-P Container scanning | SUPP Infrastructure scanning | SUPP Infrastructure scanning |

## System and Communications Protection (SC)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| App Partitioning | SC-2 | YES-P Microservices separation | N/A | N/A |
| Info in Shared Resources | SC-4 | YES-P Tenant isolation | CFG Namespace config | CFG Namespace config |
| Boundary Protection | SC-7 | CFG NetworkPolicy templates | SUPP Firewall rules | SUPP Firewall rules |
| Transmission Confidentiality | SC-8 | YES-P TLS 1.2+ | CFG Certificate management | CFG Certificate management |
| Crypto Key Mgmt | SC-12 | CFG PKI integration | SUPP CA management | SUPP CA management |
| Crypto Protection | SC-13 | YES-P FIPS-capable | CFG FIPS enablement | CFG FIPS enablement |
| Public Access | SC-14 | CFG Access controls | SUPP Public exposure rules | SUPP Public exposure rules |

## Supply Chain Risk Management (SR)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| SR Policy | SR-1 | SUPP Vendor mgmt policy | CUST SR policy | CUST SR policy |
| SCRM Plan | SR-2 | SUPP SCRM documentation | CUST SCRM plan development | CUST SCRM plan development |
| Supply Chain Controls | SR-3 | YES-P CI/CD scanning, SBOM, pinned deps | SUPP Integration testing | SUPP Integration testing |
| Provenance | SR-4 | CFG Image signing, SHA pinning | SUPP Verification | SUPP Verification |
| Acquisition Strategies | SR-5 | SUPP License compliance, vendor assessment | CUST Procurement controls | CUST Procurement controls |
| Supplier Assessments | SR-6 | SUPP SOC 2 reports, vendor inventory | CUST Supplier reviews | CUST Supplier reviews |
| Supply Chain Operations Security | SR-7 | CFG Secure build pipeline | SUPP Operations security | SUPP Operations security |

## System and Information Integrity (SI)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Flaw Remediation | SI-2 | YES-P Security updates | SUPP Patching | SUPP Patch approval |
| Malicious Code | SI-3 | CFG Container scanning | SUPP Runtime protection | SUPP Runtime protection |
| System Monitoring | SI-4 | YES-P Prometheus metrics | SUPP Monitoring setup | SUPP Monitoring setup |
| Security Alerts | SI-5 | YES-P Alert generation | SUPP Alert distribution | SUPP Alert distribution |
| SW/Firmware Integrity | SI-7 | YES-P Image signing | SUPP Verification | SUPP Verification |

## Summary Statistics

### Control Family Coverage

| Family | Controls Covered | Status |
|--------|-----------------|--------|
| Access Control (AC) | 11 | Complete |
| Awareness and Training (AT) | 2 | Complete |
| Audit and Accountability (AU) | 9 | Complete |
| Configuration Management (CM) | 7 | Complete |
| Contingency Planning (CP) | 5 | Complete |
| Identification and Authentication (IA) | 8 | Complete |
| Incident Response (IR) | 6 | Complete |
| Maintenance (MA) | 5 | Complete |
| Physical and Environmental Protection (PE) | 8 | Complete (AWS inherited) |
| Personnel Security (PS) | 6 | Complete |
| Planning (PL) | 3 | Complete |
| Risk Assessment (RA) | 2 | Complete |
| System and Communications Protection (SC) | 7 | Complete |
| Supply Chain Risk Management (SR) | 7 | **NEW** |
| System and Information Integrity (SI) | 5 | Complete |
| **Total** | **91** | **15 families** |

### Model B (Self-Hosted)

| Responsibility | Count |
|----------------|-------|
| Aegis Provides (YES-P) | 22 |
| Aegis Provides, Customer Configures (CFG) | 27 |
| Customer Implements, Aegis Supports (SUPP) | 23 |
| Customer Only (CUST) | 11 |
| Inherited from AWS (INH) | 8 |

### Model C (Managed)

| Responsibility | Count |
|----------------|-------|
| Aegis Provides (YES-P) | 22 |
| Aegis Provides, Customer Configures (CFG) | 29 |
| Customer Implements, Aegis Supports (SUPP) | 19 |
| Customer Only (CUST) | 13 |
| Inherited from AWS (INH) | 8 |

## NIST 800-171 Mapping

The controls above map to NIST SP 800-171 Rev 2 requirements as follows:

| 800-171 Family | 800-53 Controls in This Matrix | Coverage |
|----------------|-------------------------------|----------|
| 3.1 Access Control | AC-2, AC-3, AC-4, AC-5, AC-6, AC-7, AC-8, AC-11, AC-12, AC-17, AC-18 | Complete |
| 3.2 Awareness and Training | AT-2, AT-3 | Complete |
| 3.3 Audit and Accountability | AU-2, AU-3, AU-4, AU-5, AU-6, AU-7, AU-8, AU-9, AU-11 | Complete |
| 3.4 Configuration Management | CM-2, CM-3, CM-4, CM-5, CM-6, CM-7, CM-8 | Complete |
| 3.5 Identification and Authentication | IA-2, IA-2(1), IA-2(6), IA-2(12), IA-3, IA-4, IA-5, IA-7 | Complete |
| 3.6 Incident Response | IR-1, IR-2, IR-3, IR-4, IR-5, IR-6 | Complete |
| 3.7 Maintenance | MA-2, MA-3, MA-4, MA-5, MA-6 | Complete |
| 3.8 Media Protection | (Not applicable -- no removable media) | N/A |
| 3.9 Personnel Security | PS-2, PS-3, PS-4, PS-5, PS-6, PS-8 | Complete |
| 3.10 Physical Protection | PE-1 through PE-14 | Complete (AWS inherited) |
| 3.11 Risk Assessment | RA-3, RA-5 | Complete |
| 3.12 Security Assessment | (Covered by PL-1, PL-2) | Complete |
| 3.13 System and Communications Protection | SC-2, SC-4, SC-7, SC-8, SC-12, SC-13, SC-14 | Complete |
| 3.14 System and Information Integrity | SI-2, SI-3, SI-4, SI-5, SI-7 | Complete |

## Using This Matrix

### For Your SSP

Copy the relevant rows into your System Security Plan. For each control:

1. Use the "Aegis Provides" column to describe inherited controls
2. Document your implementation for "Customer Configures" and "Customer Implements"
3. Attach this CRM as an appendix to your SSP

### For Auditors

This matrix demonstrates:
- Clear delineation of responsibilities
- No security gaps between Aegis and customer
- Documented handoff points
- Complete coverage of 15 NIST 800-53 control families
- Full NIST 800-171 mapping for CUI-handling systems

### For Supply Chain

The SR (Supply Chain Risk Management) family coverage demonstrates:
- Aegis maintains vendor inventory with risk-tiered assessments
- Software dependencies are scanned, pinned, and tracked via SBOM
- Container images are signed and SHA-pinned for cloud deployments
- Customers receive transparency into Aegis supply chain practices

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0-DRAFT | TBD | TBD | Initial draft with 12 families |
| 2.0.0-DRAFT | 2026-03-09 | Carlos Sanchez | Added SR family (7 controls), PE family (8 controls), PL family (3 controls); updated PS family; added NIST 800-171 mapping; updated statistics to 91 controls across 15 families |
