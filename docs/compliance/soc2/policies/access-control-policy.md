# Access Control Policy

**Policy ID:** ACP-001
**Version:** 1.0
**Effective Date:** 2026-01-17
**Next Review:** 2027-01-17
**Owner:** Carlos Sanchez (Founder)
**Approved By:** Carlos Sanchez (Founder)

## 1. Purpose

This policy establishes requirements for controlling access to Aegis Technologies information systems and data to ensure only authorized individuals have access appropriate to their roles.

## 2. Scope

This policy applies to:
- All information systems and applications
- All data repositories and databases
- All network resources
- All employees, contractors, and third parties

## 3. Policy Statements

### 3.1 Access Principles

- **Least Privilege**: Users receive minimum access required for their role
- **Need-to-Know**: Access granted only for legitimate business purposes
- **Separation of Duties**: Critical functions require multiple approvers
- **Defense in Depth**: Multiple layers of access controls

### 3.2 Account Management

#### 3.2.1 Account Creation

- New accounts require manager approval
- Account requests must specify required access level
- Accounts are created with temporary passwords
- Users must change password on first login

#### 3.2.2 Account Modification

- Role changes require updated access review
- Access additions require manager approval
- Privilege escalation requires security approval

#### 3.2.3 Account Termination

- Access revoked within 24 hours of termination
- Manager notifies IT/Security of departures
- Shared credentials rotated upon departure
- Exit checklist completed for all departures

### 3.3 Authentication Requirements

| Access Type | Minimum Requirement |
|-------------|-------------------|
| Standard user | Strong password |
| Privileged access | MFA required |
| Production systems | MFA + approval |
| Customer data | MFA + audit logging |

#### 3.3.1 Password Requirements

- Minimum 12 characters
- Complexity: upper, lower, number, special
- No reuse of last 12 passwords
- Maximum age: 90 days (or passwordless)
- Account lockout after 5 failed attempts

#### 3.3.2 Multi-Factor Authentication

- Required for all privileged access
- Required for remote access
- Required for access to customer data
- Acceptable factors: TOTP, FIDO2, push notification

### 3.4 Access Reviews

| Review Type | Frequency | Reviewer |
|-------------|-----------|----------|
| User access | Quarterly | Managers |
| Privileged access | Monthly | Security |
| Service accounts | Quarterly | System owners |
| Third-party access | Quarterly | Vendor managers |

### 3.5 Privileged Access

- Privileged accounts are separate from standard accounts
- Privileged sessions are logged and monitored
- Just-in-time (JIT) access preferred where possible
- Standing privileged access reviewed monthly

### 3.6 Remote Access

- VPN or zero-trust access required
- MFA required for all remote access
- Session timeouts enforced (15 minutes idle)
- Split tunneling prohibited

### 3.7 Service Accounts

- Service accounts documented with owners
- Passwords rotated every 90 days
- Interactive login disabled where possible
- Permissions limited to specific functions

## 4. Procedures

### 4.1 Access Request Procedure

1. User submits access request with business justification
2. Manager approves request
3. Security reviews privileged access requests
4. IT/Admin provisions access
5. User acknowledges access granted

### 4.2 Access Review Procedure

1. Security generates access report
2. Managers review assigned users
3. Certify continued need or request removal
4. Security tracks completion
5. Deviations escalated to management

### 4.3 Termination Procedure

1. HR/Manager notifies IT of termination
2. IT disables accounts within 24 hours
3. Shared credentials rotated
4. Access logs preserved
5. Exit checklist completed

## 5. Roles and Responsibilities

> **Note:** Aegis is currently a solo-founder company. Roles will be assigned as the team grows.

| Role | Current Owner | Responsibilities |
|------|---------------|-----------------|
| Access Approver | Carlos Sanchez | Approve access, certify reviews, notify of changes |
| Security | Carlos Sanchez | Policy, privileged access approval, review coordination |
| IT/Admin | Carlos Sanchez | Account provisioning, deprovisioning, technical controls |
| Users | Future hires | Protect credentials, report suspicious activity |

## 6. Compliance Monitoring

- Access logs reviewed for anomalies
- Failed login attempts monitored
- Privileged actions audited
- Review completion tracked

## 7. Exceptions

Exceptions require:
- Written business justification
- Risk assessment
- Compensating controls
- Security Lead approval
- Time-limited duration (max 90 days)

## 8. Related Documents

- [Information Security Policy](./information-security-policy.md)
- [Incident Response Policy](./incident-response-policy.md)
- Access Request Form (link)
- Access Review Procedure (link)

## 9. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial approved version |

---

## 10. Policy Acknowledgment

| Name | Role | Date | Signature |
|------|------|------|-----------|
| Carlos Sanchez | Founder | 2026-01-17 | /s/ Carlos Sanchez |
