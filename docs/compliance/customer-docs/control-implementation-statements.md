# Aegis Control Implementation Statements

**Version:** 1.0.0-DRAFT
**Last Updated:** [DATE]
**Classification:** Customer Confidential

## Overview

This document provides Control Implementation Statements (CIS) for NIST 800-53 Rev 5 controls as implemented by Aegis Platform. Customers can use these statements directly in their System Security Plan (SSP) Section 13.

## How to Use This Document

For each control:
1. **Aegis Provides** - What Aegis implements automatically
2. **Customer Configures** - What the customer must set up
3. **Customer Implements** - What falls outside Aegis scope
4. **Evidence** - How to demonstrate implementation

## Control Family: Access Control (AC)

### AC-2: Account Management

**Control:** The organization manages information system accounts.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Keycloak IdP with user lifecycle management<br>- OIDC/SAML integration<br>- Role-based access control<br>- Audit logging of account events |
| **Customer Configures** | - Keycloak realm settings<br>- User provisioning workflows<br>- Account approval processes<br>- Inactivity timeout (default: 30 days) |
| **Customer Implements** | - Account request process<br>- Periodic account reviews<br>- Termination procedures |
| **Evidence** | - Keycloak user list export<br>- Account creation audit logs<br>- Access review records |

### AC-3: Access Enforcement

**Control:** The system enforces approved authorizations for logical access.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - RBAC enforcement at Platform API<br>- Namespace isolation in Kubernetes<br>- Project-level access boundaries<br>- Audit logging of access decisions |
| **Customer Configures** | - Role assignments in Keycloak<br>- Project membership<br>- Custom RBAC policies |
| **Customer Implements** | - Access request/approval workflow<br>- Separation of duties matrix |
| **Evidence** | - RBAC policy configurations<br>- Access denied audit logs<br>- Role assignment exports |

### AC-6: Least Privilege

**Control:** The organization employs the principle of least privilege.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Predefined roles (Admin, Owner, Developer, Viewer)<br>- Kubernetes RBAC for agents<br>- Service accounts with minimal permissions |
| **Customer Configures** | - Role assignments per user<br>- Custom role definitions (optional) |
| **Customer Implements** | - Privileged access management process<br>- Regular privilege audits |
| **Evidence** | - Role definition exports<br>- User-to-role mapping<br>- Kubernetes RBAC manifests |

### AC-11: Device Lock

**Control:** The system prevents further access after inactivity period.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Configurable session idle timeout<br>- SUSPENDED state (vs FAILED) for long operations<br>- Automatic session termination |
| **Customer Configures** | - `SESSION_IDLE_TIMEOUT` (default: 15 min)<br>- `AEGIS_PROXY_TOKEN_TTL_SECONDS` (max: 300s) |
| **Customer Implements** | - Workstation lock policies |
| **Evidence** | - Session timeout configuration<br>- Session termination audit logs |

### AC-17: Remote Access

**Control:** The organization authorizes, monitors, and controls remote access.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - TLS-encrypted remote access to workloads<br>- Short-lived connection tokens<br>- One-time use tokens (optional)<br>- Audit logging of all remote sessions |
| **Customer Configures** | - VPN requirements (outside Aegis)<br>- Token TTL settings<br>- One-time token enforcement |
| **Customer Implements** | - Remote access policy<br>- Authorized user lists |
| **Evidence** | - Connection session audit logs<br>- Token configuration exports |

## Control Family: Audit and Accountability (AU)

### AU-2: Event Logging

**Control:** The system generates audit records for defined events.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Authentication events<br>- Authorization decisions<br>- Workload lifecycle events<br>- Administrative actions<br>- Security-relevant configuration changes |
| **Customer Configures** | - Log export destinations<br>- Additional log categories |
| **Customer Implements** | - Log review procedures<br>- Alert definitions |
| **Evidence** | - Sample audit logs<br>- Event type documentation |

### AU-3: Content of Audit Records

**Control:** Audit records contain required information.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | All audit logs include:<br>- Timestamp (UTC, ISO 8601)<br>- Event type/ID<br>- Subject identity (user, service)<br>- Source IP address<br>- Action performed<br>- Resource affected<br>- Outcome (success/failure) |
| **Customer Configures** | - Additional context fields (optional) |
| **Customer Implements** | - Log format validation |
| **Evidence** | - Sample log entries demonstrating all fields |

### AU-6: Audit Record Review, Analysis, and Reporting

**Control:** The organization reviews and analyzes audit records.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Structured JSON log format<br>- SIEM-compatible output<br>- Prometheus metrics for alerting |
| **Customer Configures** | - SIEM integration (Splunk, ELK, etc.)<br>- Alert rules<br>- Dashboards |
| **Customer Implements** | - Log review schedule<br>- Incident escalation based on logs |
| **Evidence** | - SIEM integration configuration<br>- Alert rule definitions<br>- Review records |

### AU-8: Time Stamps

**Control:** The system uses internal clocks to generate timestamps.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - UTC timestamps in all logs<br>- ISO 8601 format<br>- NTP-synchronized containers |
| **Customer Configures** | - NTP servers for nodes |
| **Customer Implements** | - Time synchronization monitoring |
| **Evidence** | - Sample timestamps<br>- NTP configuration |

## Control Family: Identification and Authentication (IA)

### IA-2: Identification and Authentication (Organizational Users)

**Control:** The system uniquely identifies and authenticates users.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - OIDC authentication via Keycloak<br>- SAML 2.0 support<br>- Unique user identifiers in all logs<br>- Session binding to authenticated identity |
| **Customer Configures** | - Identity provider federation<br>- User attribute mapping<br>- Authentication policies |
| **Customer Implements** | - User onboarding/verification process |
| **Evidence** | - Keycloak IdP configuration<br>- Authentication logs |

### IA-2(1): Multi-Factor Authentication

**Control:** The system implements MFA for privileged and network access.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - MFA enforcement option (`REQUIRE_PHISHING_RESISTANT_MFA=true`)<br>- AMR claim validation<br>- Rejection of non-MFA tokens when enforced |
| **Customer Configures** | - Keycloak MFA settings<br>- Authenticator requirements<br>- Environment variable to enforce |
| **Customer Implements** | - MFA device management<br>- Recovery procedures |
| **Evidence** | - MFA configuration screenshots<br>- Auth logs showing AMR claims |

### IA-2(12): PIV/CAC Authentication

**Control:** The system accepts PIV credentials.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Keycloak X.509 authentication support<br>- Certificate-to-user mapping |
| **Customer Configures** | - CAC/PIV Keycloak adapter<br>- Certificate trust chain<br>- User attribute extraction |
| **Customer Implements** | - Card reader deployment<br>- Certificate revocation checking |
| **Evidence** | - Keycloak X.509 configuration<br>- Successful CAC auth logs |

### IA-5: Authenticator Management

**Control:** The organization manages system authenticators.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Password policy enforcement via Keycloak<br>- Token expiration controls<br>- Credential storage encryption |
| **Customer Configures** | - Password complexity rules<br>- Password history<br>- Token lifetimes |
| **Customer Implements** | - Initial credential distribution<br>- Password reset process |
| **Evidence** | - Password policy configuration<br>- Token lifetime settings |

## Control Family: System and Communications Protection (SC)

### SC-8: Transmission Confidentiality and Integrity

**Control:** The system protects data in transit.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - TLS 1.2+ for all external connections<br>- mTLS for internal service communication<br>- FIPS cipher suites (when enabled) |
| **Customer Configures** | - TLS certificates<br>- Cipher suite selection<br>- FIPS mode enablement |
| **Customer Implements** | - Certificate lifecycle management |
| **Evidence** | - TLS configuration<br>- SSL Labs scan results<br>- Network capture showing encryption |

### SC-12: Cryptographic Key Management

**Control:** The organization establishes and manages cryptographic keys.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - step-ca integration (optional)<br>- cert-manager automation<br>- Key rotation support<br>- Kubernetes Secrets for key storage |
| **Customer Configures** | - CA hierarchy<br>- Certificate lifetimes<br>- Rotation schedules |
| **Customer Implements** | - Root CA protection<br>- Key recovery procedures |
| **Evidence** | - PKI configuration<br>- Certificate inventory<br>- Rotation logs |

### SC-13: Cryptographic Protection

**Control:** The system implements cryptographic mechanisms.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - FIPS 140-2/3 capable builds<br>- AES-256-GCM for encryption<br>- RSA-2048/ECDSA-256 for signatures<br>- BoringCrypto (Go), FIPS OpenSSL (Node.js) |
| **Customer Configures** | - FIPS mode enablement<br>- Crypto policy selection |
| **Customer Implements** | - FIPS validation evidence for auditors |
| **Evidence** | - FIPS module configuration<br>- Crypto library versions |

## Control Family: System and Information Integrity (SI)

### SI-4: System Monitoring

**Control:** The organization monitors the system for attacks and indicators of compromise.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Prometheus metrics export<br>- Authentication failure counters<br>- Rate limiting alerts<br>- Structured security logs |
| **Customer Configures** | - Alert thresholds<br>- SIEM integration<br>- Dashboards |
| **Customer Implements** | - SOC monitoring<br>- Incident response triggers |
| **Evidence** | - Monitoring configuration<br>- Sample alerts<br>- Dashboard screenshots |

## Control Summary Matrix

| Control | Aegis Implements | Customer Configures | Customer Implements |
|---------|-----------------|--------------------|--------------------|
| AC-2 | ✅ | ✅ | ✅ |
| AC-3 | ✅ | ✅ | ✅ |
| AC-6 | ✅ | ✅ | ✅ |
| AC-11 | ✅ | ✅ | ⬜ |
| AC-17 | ✅ | ✅ | ✅ |
| AU-2 | ✅ | ✅ | ✅ |
| AU-3 | ✅ | ⬜ | ⬜ |
| AU-6 | ✅ | ✅ | ✅ |
| AU-8 | ✅ | ✅ | ⬜ |
| IA-2 | ✅ | ✅ | ✅ |
| IA-2(1) | ✅ | ✅ | ✅ |
| IA-2(12) | ✅ | ✅ | ✅ |
| IA-5 | ✅ | ✅ | ✅ |
| SC-8 | ✅ | ✅ | ✅ |
| SC-12 | ✅ | ✅ | ✅ |
| SC-13 | ✅ | ✅ | ⬜ |
| SI-4 | ✅ | ✅ | ✅ |

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0-DRAFT | TBD | TBD | Initial draft |
