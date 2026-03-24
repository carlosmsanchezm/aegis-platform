# FedRAMP Appendix F -- Rules of Behavior

**System Name:** Aegis Platform
**Version:** 1.0.0-DRAFT
**Last Updated:** 2026-03-09
**Classification:** Customer Confidential
**Policy Reference:** ISP-001 (Information Security Policy)
**Owner:** Carlos Sanchez (Founder)
**Approved By:** Carlos Sanchez (Founder)

## 1. Purpose

This document establishes the Rules of Behavior (RoB) for all users of the Aegis Platform, as required by FedRAMP and NIST SP 800-53 control PL-4. These rules define acceptable and unacceptable activities when accessing or using the Aegis Platform and its associated resources, including multi-cluster Kubernetes environments, workload scheduling services, workspace access, and administrative interfaces.

All users must read, understand, and acknowledge these rules before being granted access to the Aegis Platform.

## 2. Scope

These Rules of Behavior apply to:

- All individuals granted access to the Aegis Platform, including employees, contractors, consultants, temporary staff, and third-party partners
- All methods of access: web UI, Platform API (gRPC/HTTP), workspace connections (SSH/TLS tunnel), kubectl, Helm, and administrative consoles (Keycloak, AWS)
- All platform components: hub cluster services (platform-api, proxy, Keycloak), spoke cluster agents (k8s-agent, spoke-proxy), workspace pods, and supporting infrastructure (PostgreSQL, step-ca PKI, cert-manager)
- All data processed, stored, or transmitted by the Aegis Platform, including workload metadata, cluster configurations, user credentials, audit logs, and container images
- Both cloud deployments (AWS EKS) and local development environments (Docker Desktop)

## 3. Acceptable Use

### 3.1 General Principles

Users of the Aegis Platform shall:

- Use platform resources only for authorized business purposes related to workload scheduling, development, and operational activities
- Protect the confidentiality, integrity, and availability of all information processed by the platform
- Comply with all applicable federal, state, and organizational security policies, including NIST SP 800-53, NIST SP 800-171, and FedRAMP requirements
- Report any suspected security incidents, policy violations, or system anomalies immediately through established channels
- Use only authorized devices and approved network connections to access the platform
- Keep authentication credentials (passwords, MFA tokens, API keys) confidential and never share them with others

### 3.2 Authentication and Access

Users shall:

- Use only their assigned personal accounts; never use shared accounts or another user's credentials
- Enable and use multi-factor authentication (MFA) when required by the deployment's security profile
- Select strong passwords meeting the platform's complexity requirements (minimum 12 characters, including uppercase, lowercase, numeric, and special characters)
- Change passwords immediately if compromise is suspected
- Lock or log out of sessions when leaving workstations unattended
- Not attempt to bypass authentication mechanisms, session timeouts, or access controls
- Accept the minimum level of access necessary for their role (least privilege)

### 3.3 Workload and Workspace Usage

Users shall:

- Submit workloads only to clusters and projects they are authorized to use
- Use workspace environments only for authorized development and testing activities
- Not install unauthorized software or modify system configurations within workspace pods beyond what is permitted by their role
- Not use workspace resources for cryptocurrency mining, unauthorized data processing, or personal projects
- Terminate workspace sessions when no longer needed to conserve platform resources
- Not attempt to escape container isolation, escalate privileges within Kubernetes, or access resources outside their namespace

### 3.4 API and Programmatic Access

Users shall:

- Use only documented and authorized API endpoints for programmatic access
- Protect API tokens and service account credentials with the same care as passwords
- Not share API tokens in source code, chat messages, documentation, or other insecure channels
- Implement proper token lifecycle management (short-lived tokens, rotation, revocation)
- Rate-limit automated API calls to avoid impacting platform availability
- Use TLS for all API communication; never disable certificate verification

### 3.5 Data Handling

Users shall:

- Handle data in accordance with its classification level
- Not store classified or sensitive data in unauthorized locations (e.g., personal devices, unapproved cloud storage)
- Not exfiltrate platform data, audit logs, or configuration information without authorization
- Follow data retention and disposal procedures
- Encrypt sensitive data when transmitted outside the platform boundary

### 3.6 Administrative Access

Personnel with administrative access (Keycloak admin, kubectl cluster-admin, AWS IAM admin) shall additionally:

- Use administrative privileges only when performing authorized administrative tasks
- Use separate administrative accounts from standard user accounts when technically feasible
- Document all significant administrative actions (configuration changes, user management, security policy modifications)
- Not modify audit logs or disable audit logging mechanisms
- Follow the change management process for all configuration changes (PR-based review, documented rollback plan)
- Review and validate system configurations against the Configuration Hardening Guide

## 4. Security Responsibilities

### 4.1 All Users

All users are responsible for:

- **Credential Protection**: Safeguarding passwords, MFA devices, API tokens, SSH keys, and TLS certificates. Never store credentials in plaintext, share them in messages, or commit them to source code repositories.
- **Incident Reporting**: Immediately reporting suspected security incidents, unauthorized access, phishing attempts, or unusual system behavior to the security contact (as defined in the Incident Response Policy, IRP-001).
- **Security Awareness**: Completing required security training annually and staying informed about current threats and security best practices.
- **Physical Security**: Securing workstations and devices used to access the platform. Enabling device encryption, screen locks, and ensuring physical access to hardware authenticators is controlled.
- **Software Currency**: Keeping client-side software (browsers, kubectl, gRPC tools) updated to supported versions with current security patches.

### 4.2 System Administrators

In addition to all user responsibilities, system administrators are responsible for:

- Maintaining platform security configurations in accordance with the Configuration Hardening Guide
- Monitoring security alerts and responding within defined SLA timeframes (P1: Immediate, P2: < 1 hour, P3: < 4 hours, P4: < 24 hours)
- Performing regular access reviews (quarterly for standard access, monthly for privileged access)
- Ensuring audit log integrity and retention compliance
- Managing PKI components (step-ca, cert-manager) and certificate lifecycle
- Coordinating with the Security Administrator on security-relevant changes

### 4.3 Platform Developers

In addition to all user responsibilities, platform developers are responsible for:

- Following secure coding practices and the change management process (CMP-001)
- Submitting all code changes via pull request with required review
- Not introducing hardcoded credentials, secrets, or sensitive data into source code
- Running security scanning tools before submitting changes
- Documenting security-relevant changes in release notes

## 5. Prohibited Activities

The following activities are strictly prohibited when using the Aegis Platform:

### 5.1 Unauthorized Access

- Attempting to access systems, data, accounts, or resources for which authorization has not been granted
- Using another user's credentials or impersonating another user
- Attempting to escalate privileges beyond assigned role permissions
- Bypassing or disabling security controls, authentication mechanisms, or access enforcement
- Probing, scanning, or testing the security of platform components without explicit written authorization
- Accessing, reading, or modifying audit logs to conceal unauthorized activity

### 5.2 Unauthorized Disclosure

- Sharing, copying, or transmitting platform data, configurations, or audit logs to unauthorized individuals or systems
- Posting platform credentials, API keys, or internal URLs in public forums, repositories, or communication channels
- Taking screenshots or recordings of sensitive platform interfaces without authorization
- Discussing classified or sensitive platform information in unsecured communication channels

### 5.3 System Misuse

- Using platform resources for personal use, commercial activities not authorized by the organization, or illegal activities
- Installing or executing malicious software, scripts, or tools designed to compromise system integrity
- Deliberately degrading platform performance through resource exhaustion, denial-of-service activity, or excessive API calls
- Attempting to escape container isolation, break out of namespace boundaries, or access the underlying host
- Modifying or tampering with platform infrastructure components without following the change management process
- Disabling or circumventing monitoring, logging, or alerting mechanisms
- Using the platform for cryptocurrency mining, unauthorized network scanning, or hosting unauthorized services

### 5.4 Data Integrity Violations

- Modifying, deleting, or corrupting platform data, workload records, or cluster configurations without authorization
- Tampering with audit logs, metrics, or compliance evidence
- Introducing unauthorized data into the platform or falsifying workload metadata
- Bypassing data validation or input sanitization controls

## 6. Consequences of Non-Compliance

### 6.1 Enforcement Actions

Violations of these Rules of Behavior may result in one or more of the following actions, depending on the severity and nature of the violation:

| Severity | Examples | Potential Consequences |
|----------|----------|----------------------|
| **Minor** | Failure to lock workstation, weak password selection, minor policy deviation | Verbal warning, mandatory re-training, increased monitoring |
| **Moderate** | Sharing credentials, unauthorized software installation, repeated minor violations | Written warning, temporary access suspension, mandatory security training, access privilege reduction |
| **Serious** | Unauthorized data access, bypassing security controls, unauthorized system modification | Immediate access revocation, formal investigation, termination of employment/contract, legal referral |
| **Critical** | Data exfiltration, malicious activity, tampering with audit logs, aiding unauthorized access | Immediate access revocation, termination, legal action, referral to law enforcement, regulatory notification |

### 6.2 Investigation Process

- All reported violations will be documented and investigated in accordance with the Incident Response Policy (IRP-001)
- Investigations will preserve evidence and maintain chain of custody
- Users under investigation may have their access suspended pending investigation outcome
- Investigation findings will be documented and retained for a minimum of 3 years

### 6.3 Appeal Process

- Users may appeal enforcement actions through the organizational grievance process
- Appeals must be submitted in writing within 10 business days of notification
- Appeals will be reviewed by management independent of the original investigation

## 7. Remote Access Provisions

Users accessing the Aegis Platform remotely shall additionally:

- Use only approved and secured network connections (VPN, zero-trust access)
- Ensure the remote device meets organizational security standards (disk encryption, current patches, endpoint protection)
- Not access the platform from public or untrusted networks without VPN protection
- Enable multi-factor authentication for all remote sessions
- Be aware that all remote sessions are logged and subject to monitoring
- Immediately report lost or stolen devices that have been used to access the platform

## 8. Use of Government Resources

When the Aegis Platform is used within a government context:

- Users acknowledge that the platform and all data therein are government resources subject to monitoring
- There is no expectation of privacy when using government-provisioned Aegis instances
- All activity may be monitored, recorded, and audited
- Evidence of unauthorized activity may be used for administrative, criminal, or adverse action proceedings
- Government system use banners (AC-8) will be displayed at login and must be acknowledged before access is granted

## 9. Policy Updates and Review

- These Rules of Behavior will be reviewed and updated at least annually, or upon significant changes to the platform or threat landscape
- Users will be notified of material changes and required to re-acknowledge updated rules
- Suggestions for improvements may be submitted to the Security Administrator
- The current version is always available in the platform's compliance documentation repository

## 10. Acknowledgment and Digital Signature

### 10.1 Acknowledgment Statement

By signing below, I acknowledge that:

1. I have read, understand, and agree to comply with the Aegis Platform Rules of Behavior
2. I understand the consequences of non-compliance as described in Section 6
3. I will complete required security awareness training before accessing the platform
4. I will report any suspected security incidents or policy violations immediately
5. I understand that my activities on the Aegis Platform may be monitored and recorded
6. I understand that these rules may be updated and I will review and re-acknowledge updates when notified

### 10.2 Signature Template

```
AEGIS PLATFORM RULES OF BEHAVIOR ACKNOWLEDGMENT

Full Name:       ________________________________________
Title/Role:      ________________________________________
Organization:    ________________________________________
Email:           ________________________________________
Date:            ________________________________________

Digital Signature: ________________________________________
                   (Electronic signature constitutes agreement)

Witness/Approving Official (if required):

Full Name:       ________________________________________
Title/Role:      ________________________________________
Date:            ________________________________________
Digital Signature: ________________________________________
```

### 10.3 Re-Acknowledgment Schedule

| Event | Re-Acknowledgment Required |
|-------|---------------------------|
| Annual review | Within 30 days of annual review publication |
| Material policy change | Within 14 days of change notification |
| Security incident involving user | Before access is restored |
| Role change (standard to privileged) | Before elevated access is granted |

### 10.4 Current Acknowledgments

| Name | Role | Date | Signature |
|------|------|------|-----------|
| Carlos Sanchez | Founder / System Administrator | 2026-03-09 | /s/ Carlos Sanchez |

Acknowledgment records tracked in: `aegis-compliance-evidence/soc2/[YEAR]/training-policy-ack/`

## 11. Related Documents

| Document | Relationship |
|----------|-------------|
| Information Security Policy (ISP-001) | Parent policy establishing the security program |
| Access Control Policy (ACP-001) | Detailed access control requirements |
| Change Management Policy (CMP-001) | Change process requirements referenced in Section 3.6 |
| Incident Response Policy (IRP-001) | Incident reporting and handling procedures |
| Configuration Hardening Guide | Security configuration standards |
| Vendor Management Policy (VMP-001) | Third-party access requirements |
| FedRAMP Appendix L -- Separation of Duties Matrix | Role conflict and compensating controls |

## 12. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0-DRAFT | 2026-03-09 | Carlos Sanchez | Initial draft |
