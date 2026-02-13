# Aegis Customer Responsibility Matrix

**Version:** 1.0.0-DRAFT
**Last Updated:** [DATE]
**Classification:** Customer Confidential

## Overview

This Customer Responsibility Matrix (CRM) defines the security responsibilities shared between Aegis and the customer for each NIST 800-53 control. Use this document to understand what Aegis provides versus what you must implement.

## Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Fully provided by Aegis |
| ⚙️ | Aegis provides, customer configures |
| 🔧 | Customer implements, Aegis supports |
| 👤 | Customer responsibility only |
| ➖ | Not applicable |

## Responsibility by Deployment Model

| Model | Aegis Responsibility | Customer Responsibility |
|-------|---------------------|------------------------|
| **Model B (Self-Hosted)** | Software capabilities, documentation | Deployment, operation, authorization |
| **Model C (Managed)** | Software + operation in customer account | Account access, authorization |

## Access Control (AC)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Account Management | AC-2 | ⚙️ Keycloak IdP, RBAC engine | 🔧 Account lifecycle, reviews | 👤 Account lifecycle, reviews |
| Access Enforcement | AC-3 | ✅ RBAC at API level | ⚙️ Role assignments | ⚙️ Role assignments |
| Information Flow | AC-4 | ⚙️ Network policy templates | 🔧 Policy deployment | 🔧 Policy approval |
| Separation of Duties | AC-5 | ⚙️ Role definitions | 🔧 SoD matrix, enforcement | 🔧 SoD matrix approval |
| Least Privilege | AC-6 | ✅ Predefined minimal roles | ⚙️ Role assignments | ⚙️ Role assignments |
| Unsuccessful Logins | AC-7 | ⚙️ Keycloak lockout config | ⚙️ Threshold settings | ⚙️ Threshold settings |
| System Use Notification | AC-8 | ⚙️ Banner capability | 🔧 Banner content | 🔧 Banner content |
| Session Lock | AC-11 | ✅ Configurable timeout | ⚙️ Timeout value | ⚙️ Timeout value |
| Session Termination | AC-12 | ✅ Auto-termination | ⚙️ Timeout value | ⚙️ Timeout value |
| Remote Access | AC-17 | ✅ Secure tunnel, tokens | ⚙️ VPN policy | ⚙️ VPN policy |
| Wireless Access | AC-18 | ➖ | 👤 Network controls | 👤 Network controls |

## Awareness and Training (AT)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Security Training | AT-2 | 🔧 Product training materials | 👤 Security awareness program | 👤 Security awareness program |
| Role-Based Training | AT-3 | 🔧 Admin training docs | 👤 Role-specific training | 👤 Role-specific training |

## Audit and Accountability (AU)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Event Logging | AU-2 | ✅ Security events logged | ⚙️ Additional events | ⚙️ Additional events |
| Content of Records | AU-3 | ✅ All required fields | ➖ | ➖ |
| Audit Storage | AU-4 | ⚙️ Local storage, export | 🔧 SIEM storage | 🔧 SIEM storage |
| Audit Failure Response | AU-5 | ⚙️ Alerting on failure | 🔧 Alert response | 🔧 Alert response |
| Audit Review | AU-6 | ⚙️ SIEM integration | 👤 Review procedures | 👤 Review procedures |
| Audit Reduction | AU-7 | ⚙️ Structured JSON logs | 🔧 Query/dashboards | 🔧 Query/dashboards |
| Time Stamps | AU-8 | ✅ UTC timestamps | ⚙️ NTP configuration | ⚙️ NTP configuration |
| Audit Protection | AU-9 | ⚙️ Log integrity options | 🔧 Tamper protection | 🔧 Tamper protection |
| Audit Retention | AU-11 | ⚙️ Retention config | 🔧 Retention policy | 🔧 Retention policy |

## Configuration Management (CM)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Baseline Config | CM-2 | ⚙️ Default Helm values | 🔧 Hardened values | 🔧 Hardened values approval |
| Config Change Control | CM-3 | ✅ GitOps, PR reviews | 🔧 Change process | 👤 Change approval |
| Security Impact Analysis | CM-4 | 🔧 Release notes | 👤 Impact analysis | 👤 Impact analysis |
| Access Restrictions | CM-5 | ⚙️ RBAC for config | 🔧 Access controls | 🔧 Access controls |
| Config Settings | CM-6 | ⚙️ Hardening guide | 🔧 Implementation | 🔧 Verification |
| Least Functionality | CM-7 | ✅ Minimal containers | ⚙️ Disable unused features | ⚙️ Disable unused features |
| SW Inventory | CM-8 | ✅ SBOM generation | 🔧 Inventory tracking | 🔧 Inventory tracking |

## Contingency Planning (CP)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Contingency Plan | CP-2 | 🔧 DR guidance | 👤 BCP/DR plan | 👤 BCP/DR plan |
| Contingency Training | CP-3 | 🔧 Training materials | 👤 Training program | 👤 Training program |
| Contingency Testing | CP-4 | 🔧 Test procedures | 👤 Testing | 👤 Testing |
| Backup | CP-9 | ⚙️ Backup guidance | 🔧 Backup implementation | 🔧 Backup implementation |
| Recovery | CP-10 | ⚙️ Recovery procedures | 🔧 Recovery testing | 🔧 Recovery testing |

## Identification and Authentication (IA)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| User I&A | IA-2 | ✅ OIDC/SAML authentication | ⚙️ IdP configuration | ⚙️ IdP configuration |
| MFA | IA-2(1) | ✅ MFA support | ⚙️ MFA enforcement | ⚙️ MFA enforcement |
| Phishing-Resistant MFA | IA-2(6) | ✅ FIDO2/WebAuthn support | ⚙️ Authenticator config | ⚙️ Authenticator config |
| PIV/CAC | IA-2(12) | ⚙️ X.509 support | 🔧 CAC integration | 🔧 CAC integration |
| Device I&A | IA-3 | ⚙️ mTLS for services | 🔧 Device certificates | 🔧 Device certificates |
| Identifier Management | IA-4 | ⚙️ Keycloak ID management | 🔧 ID lifecycle | 🔧 ID lifecycle |
| Authenticator Mgmt | IA-5 | ⚙️ Password policies | 🔧 Policy enforcement | 🔧 Policy enforcement |
| Cryptographic Auth | IA-7 | ✅ TLS, JWT signing | ⚙️ Algorithm selection | ⚙️ Algorithm selection |

## Incident Response (IR)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| IR Policy | IR-1 | 🔧 IR runbook | 👤 IR policy | 👤 IR policy |
| IR Training | IR-2 | 🔧 Training materials | 👤 Training program | 👤 Training program |
| IR Testing | IR-3 | 🔧 Test scenarios | 👤 Testing | 👤 Testing |
| IR Handling | IR-4 | 🔧 Runbook procedures | 👤 IR execution | 👤 IR execution |
| IR Monitoring | IR-5 | ⚙️ Security alerts | 🔧 Alert triage | 🔧 Alert triage |
| IR Reporting | IR-6 | 🔧 Report templates | 👤 Reporting | 👤 Reporting |

## Maintenance (MA)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Controlled Maintenance | MA-2 | 🔧 Maintenance windows | 🔧 Scheduling | 🔧 Scheduling |
| Maintenance Tools | MA-3 | ✅ kubectl, Helm | ⚙️ Tool access | ⚙️ Tool access |
| Remote Maintenance | MA-4 | ⚙️ Secure admin access | 🔧 Access controls | 🔧 Access controls |
| Timely Maintenance | MA-6 | ✅ Security patches | 🔧 Patch deployment | 🔧 Patch approval |

## Personnel Security (PS)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Position Risk | PS-2 | ➖ | 👤 Risk designations | 👤 Risk designations |
| Personnel Screening | PS-3 | ➖ | 👤 Background checks | 👤 Background checks |
| Personnel Termination | PS-4 | ⚙️ Account disable | 👤 Termination process | 👤 Termination process |
| Personnel Transfer | PS-5 | ⚙️ Role changes | 👤 Transfer process | 👤 Transfer process |
| Access Agreements | PS-6 | ➖ | 👤 User agreements | 👤 User agreements |

## Risk Assessment (RA)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Risk Assessment | RA-3 | 🔧 Security assessment | 👤 System risk assessment | 👤 System risk assessment |
| Vulnerability Scanning | RA-5 | ✅ Container scanning | 🔧 Infrastructure scanning | 🔧 Infrastructure scanning |

## System and Communications Protection (SC)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| App Partitioning | SC-2 | ✅ Microservices separation | ➖ | ➖ |
| Info in Shared Resources | SC-4 | ✅ Tenant isolation | ⚙️ Namespace config | ⚙️ Namespace config |
| Boundary Protection | SC-7 | ⚙️ NetworkPolicy templates | 🔧 Firewall rules | 🔧 Firewall rules |
| Transmission Confidentiality | SC-8 | ✅ TLS 1.2+ | ⚙️ Certificate management | ⚙️ Certificate management |
| Crypto Key Mgmt | SC-12 | ⚙️ PKI integration | 🔧 CA management | 🔧 CA management |
| Crypto Protection | SC-13 | ✅ FIPS-capable | ⚙️ FIPS enablement | ⚙️ FIPS enablement |
| Public Access | SC-14 | ⚙️ Access controls | 🔧 Public exposure rules | 🔧 Public exposure rules |

## System and Information Integrity (SI)

| Control | ID | Aegis | Customer (Model B) | Customer (Model C) |
|---------|----|---------|--------------------|-------------------|
| Flaw Remediation | SI-2 | ✅ Security updates | 🔧 Patching | 🔧 Patch approval |
| Malicious Code | SI-3 | ⚙️ Container scanning | 🔧 Runtime protection | 🔧 Runtime protection |
| System Monitoring | SI-4 | ✅ Prometheus metrics | 🔧 Monitoring setup | 🔧 Monitoring setup |
| Security Alerts | SI-5 | ✅ Alert generation | 🔧 Alert distribution | 🔧 Alert distribution |
| SW/Firmware Integrity | SI-7 | ✅ Image signing | 🔧 Verification | 🔧 Verification |

## Summary Statistics

### Model B (Self-Hosted)

| Responsibility | Count |
|----------------|-------|
| Aegis Provides (✅) | 35 |
| Aegis Provides, Customer Configures (⚙️) | 42 |
| Customer Implements, Aegis Supports (🔧) | 38 |
| Customer Only (👤) | 25 |
| Not Applicable (➖) | 10 |

### Model C (Managed)

| Responsibility | Count |
|----------------|-------|
| Aegis Provides (✅) | 35 |
| Aegis Provides, Customer Configures (⚙️) | 45 |
| Customer Implements, Aegis Supports (🔧) | 30 |
| Customer Only (👤) | 30 |
| Not Applicable (➖) | 10 |

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

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0-DRAFT | TBD | TBD | Initial draft |
