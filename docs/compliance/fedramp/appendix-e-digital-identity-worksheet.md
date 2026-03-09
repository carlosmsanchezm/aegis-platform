# FedRAMP Appendix E -- Digital Identity Determination Worksheet

**System Name:** Aegis Platform
**Version:** 1.0.0-DRAFT
**Last Updated:** 2026-03-09
**Classification:** Customer Confidential
**Reference Standard:** NIST SP 800-63-3, Digital Identity Guidelines

## 1. Purpose

This worksheet documents the digital identity risk assessment and assurance level determination for the Aegis Platform, as required by FedRAMP for inclusion in the System Security Plan (SSP). The determination follows NIST SP 800-63-3 guidance for selecting Identity Assurance Level (IAL), Authenticator Assurance Level (AAL), and Federation Assurance Level (FAL).

## 2. System Context

Aegis Platform is a hub-and-spoke Kubernetes platform for multi-cluster workload scheduling. Users interact with the platform through:

1. **Web UI** -- Browser-based interface for workload management and monitoring
2. **Platform API** -- gRPC/HTTP API for programmatic access
3. **Workspace Access** -- TLS-tunneled connections to provisioned workspaces on spoke clusters
4. **Administrative Access** -- kubectl, Helm, and Keycloak admin console for platform management

All user authentication is centralized through **Keycloak**, which serves as the platform's Identity Provider (IdP) and supports OIDC, SAML 2.0, and X.509 certificate authentication.

## 3. Identity Assurance Level (IAL) Determination

### 3.1 IAL Definition (NIST SP 800-63A)

IAL defines the degree of confidence that the applicant's claimed identity is their real identity.

| Level | Description | Identity Proofing |
|-------|-------------|-------------------|
| IAL1 | Self-asserted identity | No identity proofing required |
| IAL2 | Remote or in-person identity proofing | Evidence collection and validation |
| IAL3 | In-person identity proofing with physical verification | Supervised, in-person proofing |

### 3.2 Aegis IAL Support

| IAL | Aegis Capability | Implementation |
|-----|-------------------|----------------|
| **IAL1** | Fully supported | Keycloak self-registration with email verification. User provides email and password. No identity proofing beyond email confirmation. Suitable for development environments and non-sensitive workloads. |
| **IAL2** | Supported via federation | Keycloak SAML/OIDC federation with enterprise IdP that performs identity proofing. Customer's enterprise IdP (e.g., Azure AD, Okta, PingFederate) handles identity proofing to IAL2 standards. Aegis inherits the proofing assurance of the upstream IdP. Also supported via Keycloak's X.509 certificate authentication for CAC/PIV-enabled environments where identity proofing was performed during card issuance. |
| **IAL3** | Supported via federation | Requires customer enterprise IdP with IAL3 capability. Aegis does not directly perform in-person proofing but can consume IAL3 assertions from federated identity providers. Applicable in DoD/IC contexts where CAC/PIV issuance includes supervised in-person proofing. |

### 3.3 IAL Determination Rationale

**Recommended Minimum: IAL1 (development), IAL2 (production with CUI)**

| Factor | Assessment |
|--------|------------|
| **Data sensitivity** | Aegis manages workload scheduling metadata, cluster configurations, and user access records. When processing CUI, IAL2 is required. |
| **Potential impact of identity fraud** | Unauthorized access could lead to workload tampering, resource abuse, or exposure of cluster configurations. Moderate impact. |
| **Regulatory requirements** | NIST 800-171 requires IAL2 minimum for CUI-handling systems. FedRAMP Moderate baseline requires IAL2. |
| **Aegis implementation** | For production deployments, customers should configure Keycloak federation with an enterprise IdP that performs IAL2 identity proofing. Self-registration (IAL1) should be disabled in production. |

### 3.4 IAL Configuration Guidance

| Deployment | Recommended IAL | Configuration |
|------------|----------------|---------------|
| Development/Test | IAL1 | Keycloak self-registration enabled, email verification required |
| Production (non-CUI) | IAL1 or IAL2 | Federation with enterprise IdP recommended |
| Production (CUI/FedRAMP) | IAL2 minimum | Federation with IAL2-capable IdP required; self-registration disabled; admin-approved registration only |
| Production (DoD/IC) | IAL2 or IAL3 | CAC/PIV authentication via Keycloak X.509; federation with DoD IdP |

## 4. Authenticator Assurance Level (AAL) Determination

### 4.1 AAL Definition (NIST SP 800-63B)

AAL defines the strength of the authentication process.

| Level | Description | Authentication Factors |
|-------|-------------|----------------------|
| AAL1 | Single-factor authentication | Password or other single factor |
| AAL2 | Multi-factor authentication | Two different authentication factors |
| AAL3 | Hardware-based multi-factor authentication | Hardware cryptographic authenticator + one other factor |

### 4.2 Aegis AAL Support

| AAL | Aegis Capability | Implementation |
|-----|-------------------|----------------|
| **AAL1** | Fully supported | Keycloak password authentication with configurable complexity policy (12-char minimum, complexity requirements, account lockout after 5 attempts, password history of 12). Session management with configurable idle timeout (default: 15 min). |
| **AAL2** | Fully supported | Keycloak TOTP/HOTP MFA with configurable enforcement. Platform API validates AMR (Authentication Methods References) claims in JWT tokens. MFA enforcement via `REQUIRE_PHISHING_RESISTANT_MFA=true` environment variable rejects tokens without MFA AMR claims. Supported authenticator types: TOTP apps (Google Authenticator, Authy), FIDO2/WebAuthn security keys, push notification via federated IdP. Also achievable via federation with enterprise IdP that enforces MFA. |
| **AAL3** | Supported via configuration | Requires hardware cryptographic authenticator (FIDO2 security key or CAC/PIV). Keycloak WebAuthn support enables FIDO2 registration and authentication. CAC/PIV via Keycloak X.509 authentication satisfies AAL3 when card contains cryptographic key. Customer configures Keycloak authentication flow to require hardware authenticator. |

### 4.3 AAL Determination Rationale

**Recommended Minimum: AAL1 (development), AAL2 (production)**

| Factor | Assessment |
|--------|------------|
| **Data sensitivity** | Platform manages compute workloads and cluster access. Compromise of authentication could lead to unauthorized workload execution or data access. |
| **Session characteristics** | Sessions involve remote access over the internet. All sessions are TLS-protected. Session tokens are short-lived (configurable, max 300 seconds for workspace tokens). |
| **Regulatory requirements** | NIST 800-171 requires AAL2 for CUI. FedRAMP Moderate requires AAL2 for all privileged access and remote access. |
| **Phishing resistance** | AAL2 with TOTP provides replay resistance but not phishing resistance. For phishing resistance, FIDO2/WebAuthn (AAL3) is recommended. |

### 4.4 AAL Configuration Guidance

| Deployment | Recommended AAL | Configuration |
|------------|----------------|---------------|
| Development/Test | AAL1 | Password authentication with Keycloak defaults |
| Production (non-CUI) | AAL2 | TOTP MFA enforced for all users in Keycloak; set `REQUIRE_PHISHING_RESISTANT_MFA=true` in platform-api |
| Production (CUI/FedRAMP) | AAL2 minimum | MFA enforced; AMR claim validation enabled; session idle timeout <= 15 min |
| Production (DoD/IC) | AAL3 | FIDO2/WebAuthn or CAC/PIV required; Keycloak flow configured to require hardware authenticator |
| Privileged access (all) | AAL2 minimum | Admin roles always require MFA regardless of deployment type |

### 4.5 AAL Technical Implementation Details

#### Keycloak MFA Configuration
```
Realm Settings > Authentication > Required Actions:
  - Configure OTP: ON (required for AAL2+)
  - WebAuthn Register: ON (required for AAL3)

Authentication Flow > Browser:
  - Username/Password (Required)
  - OTP Form (Required for AAL2)
    OR
  - WebAuthn Authenticator (Required for AAL3)
```

#### Platform API AMR Validation
```
Environment Variables:
  REQUIRE_PHISHING_RESISTANT_MFA=true   # Reject tokens without MFA AMR
  SESSION_IDLE_TIMEOUT=900               # 15-minute idle timeout
  AEGIS_PROXY_TOKEN_TTL_SECONDS=300      # 5-minute max token TTL
```

#### AMR Claim Values Recognized
| AMR Value | Meaning | AAL |
|-----------|---------|-----|
| `pwd` | Password authentication | AAL1 |
| `otp` | One-time password (TOTP) | AAL2 (with pwd) |
| `hwk` | Hardware key proof | AAL3 (with pwd) |
| `swk` | Software key proof | AAL2 (with pwd) |
| `pin` | PIN authentication | AAL2 (with pwd) |
| `fpt` | Fingerprint biometric | AAL2 (with pwd) |
| `sc` | Smart card (CAC/PIV) | AAL3 |

## 5. Federation Assurance Level (FAL) Determination

### 5.1 FAL Definition (NIST SP 800-63C)

FAL defines the strength of the federation protocol and assertion protection.

| Level | Description | Assertion Type |
|-------|-------------|----------------|
| FAL1 | Bearer assertion | Signed assertion (e.g., OIDC ID Token with RS256) |
| FAL2 | Holder-of-key assertion | Assertion bound to a cryptographic key held by the subscriber |
| FAL3 | Holder-of-key assertion with approved cryptographic module | Hardware-bound assertion |

### 5.2 Aegis FAL Support

| FAL | Aegis Capability | Implementation |
|-----|-------------------|----------------|
| **FAL1** | Fully supported (default) | OIDC bearer tokens (JWT) signed by Keycloak using RS256 or ES256. Tokens validated by Platform API using Keycloak's JWKS endpoint. Token includes `iss`, `sub`, `aud`, `exp`, `iat`, `amr` claims. TLS protects tokens in transit. Used for all user-to-platform communication (UI, API). |
| **FAL2** | Supported for service communication | mTLS + OIDC for spoke-to-hub communication. K8s Agent and Spoke Proxy authenticate using client certificates (holder-of-key) issued by step-ca via cert-manager. The client certificate binds the assertion to a cryptographic key held by the spoke. gRPC channels use mTLS with certificate pinning to the platform's trust bundle (`aegis-trust-bundle`). |
| **FAL3** | Not natively supported | Would require hardware security module (HSM) integration for key storage. Can be achieved by running step-ca with HSM backend (e.g., AWS CloudHSM, YubiHSM). Not currently part of default deployment. |

### 5.3 FAL Determination Rationale

**Recommended Minimum: FAL1 (user sessions), FAL2 (service-to-service)**

| Factor | Assessment |
|--------|------------|
| **Assertion type** | User sessions use OIDC bearer tokens (FAL1). Service-to-service uses mTLS client certificates (FAL2). |
| **Token protection** | All tokens transmitted over TLS 1.2+. OIDC tokens are signed (RS256/ES256) and have short expiration. Workspace access tokens have 5-minute max TTL with optional one-time use. |
| **Federation topology** | Keycloak acts as the federation hub. Enterprise IdPs federate to Keycloak via SAML/OIDC. Keycloak issues platform-specific tokens consumed by all Aegis services. |
| **Spoke communication** | mTLS certificates are scoped to individual spoke clusters. Certificate rotation managed by cert-manager with configurable lifetimes. Trust boundary enforced by the `aegis-trust-bundle` secret. |

### 5.4 FAL Configuration Guidance

| Communication Path | FAL | Mechanism |
|--------------------|-----|-----------|
| User browser to UI | FAL1 | OIDC bearer token via Keycloak |
| UI to Platform API | FAL1 | JWT bearer token in gRPC metadata |
| Enterprise IdP to Keycloak | FAL1 | SAML assertion or OIDC token |
| K8s Agent to Platform API | FAL2 | mTLS client certificate + service identity |
| Spoke Proxy to Hub | FAL2 | mTLS with cert-manager-issued certificate |
| User to Workspace | FAL1 | Short-lived connection token (max 300s, optional one-time use) |

## 6. Summary Determination Table

| Assurance Level | Recommended Level | Aegis Default | FedRAMP Moderate Req. | Justification |
|-----------------|-------------------|---------------|----------------------|---------------|
| **IAL** | IAL2 (production) | IAL1 (self-registration) | IAL2 | Keycloak federation with enterprise IdP provides IAL2. Self-registration disabled in production. |
| **AAL** | AAL2 (production) | AAL1 (password only) | AAL2 | Keycloak TOTP MFA with AMR validation provides AAL2. Enable via `REQUIRE_PHISHING_RESISTANT_MFA=true`. |
| **FAL** | FAL1 (user), FAL2 (service) | FAL1 (user), FAL2 (service) | FAL1 | OIDC bearer tokens for users; mTLS for service-to-service. Exceeds FAL1 for service communication. |

## 7. Risk Assessment for Digital Identity

### 7.1 Identity Risk Summary

| Risk Category | Impact Level | Likelihood | Mitigation |
|---------------|-------------|------------|------------|
| Identity fraud (unauthorized user) | Moderate | Low (with IAL2) | Enterprise IdP federation with identity proofing |
| Authentication bypass | High | Low (with AAL2) | MFA enforcement, account lockout, session management |
| Token theft/replay | Moderate | Moderate | Short token lifetimes, TLS everywhere, one-time tokens for workspace access |
| Federation assertion manipulation | High | Low | JWT signature validation, JWKS endpoint verification, TLS for token transit |
| Spoke impersonation | High | Very Low (with FAL2) | mTLS with cert-manager certificates, trust bundle pinning |

### 7.2 Compensating Controls

Where the default configuration does not meet the recommended assurance level, these compensating controls reduce risk:

| Gap | Compensating Control |
|-----|---------------------|
| IAL1 default (no identity proofing) | Admin-only account creation; manual identity verification before provisioning |
| AAL1 default (no MFA) | Short session timeouts (15 min); IP-based access restrictions; audit logging of all actions |
| FAL1 for user sessions | Token lifetime < 30 min; workspace tokens < 5 min; session binding to source IP (optional) |

## 8. Customer Actions Required

### 8.1 For FedRAMP Moderate Authorization

| Action | Priority | Description |
|--------|----------|-------------|
| Disable self-registration | P1 | Configure Keycloak to require admin approval for account creation |
| Enable MFA enforcement | P1 | Set `REQUIRE_PHISHING_RESISTANT_MFA=true`; configure Keycloak TOTP required action |
| Configure enterprise IdP federation | P1 | Federate Keycloak with customer IdP that performs IAL2 identity proofing |
| Review session timeouts | P2 | Confirm `SESSION_IDLE_TIMEOUT` <= 900 seconds (15 min) |
| Configure workspace token policy | P2 | Set `AEGIS_PROXY_TOKEN_TTL_SECONDS` <= 300; enable one-time use tokens |
| Document identity proofing process | P2 | Document how users are identity-proofed before receiving platform access |

### 8.2 For DoD IL4/IL5

All FedRAMP Moderate actions plus:

| Action | Priority | Description |
|--------|----------|-------------|
| Enable CAC/PIV authentication | P1 | Configure Keycloak X.509 authentication flow for smart card login |
| Require hardware authenticator | P1 | Configure Keycloak WebAuthn flow requiring FIDO2 security key |
| Evaluate HSM for PKI | P2 | Consider HSM-backed step-ca for FAL3 service assertions |

## 9. References

| Document | Relevance |
|----------|-----------|
| NIST SP 800-63-3 | Digital Identity Guidelines (parent document) |
| NIST SP 800-63A | Enrollment and Identity Proofing (IAL) |
| NIST SP 800-63B | Authentication and Lifecycle Management (AAL) |
| NIST SP 800-63C | Federation and Assertions (FAL) |
| FedRAMP Digital Identity Requirements | FedRAMP-specific identity guidance |
| Aegis Security Architecture Guide | Platform security architecture details |
| Aegis Configuration Hardening Guide | Security configuration settings |

## 10. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0-DRAFT | 2026-03-09 | Carlos Sanchez | Initial draft |
