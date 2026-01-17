# Employee Offboarding Checklist

**SOC 2 Control:** CC6.3 (Removing Access to Protected Information)
**Owner:** Carlos Sanchez (Founder)
**Last Updated:** 2026-01-17
**Review Frequency:** Annually or upon process change

---

## Purpose

This checklist ensures all access is revoked promptly when an employee or contractor leaves the organization. It supports SOC 2 compliance by documenting the de-provisioning process and preventing unauthorized access after separation.

---

## Offboarding Types and Timing

| Termination Type | Access Revocation Timing | Special Considerations |
|------------------|--------------------------|------------------------|
| Voluntary resignation | Last day of employment | Standard process |
| Involuntary termination | Immediately upon notification | Escort if needed, revoke before conversation |
| Contractor end-of-term | Contract end date | Verify no extensions pending |
| Role change (internal) | Same day as role change | Modify, don't fully revoke |
| Security incident | Immediately | Document for incident report |

---

## Pre-Departure (Before Last Day)

### 1. Initiation
- [ ] **Separation type documented**
  - Type: ☐ Voluntary ☐ Involuntary ☐ Contractor EOT ☐ Other: ________
  - Last day: __________
  - Reason (if applicable): ____________________
- [ ] **HR notified**
  - Date notified: __________
- [ ] **IT/Security notified**
  - Date notified: __________
- [ ] **Manager confirmed access list**
  - All systems employee had access to: ____________________

### 2. Knowledge Transfer
- [ ] **Documentation updated** (runbooks, wikis, etc.)
- [ ] **Passwords for shared accounts changed** (if any)
- [ ] **Project handoff completed**
  - Successor: ____________________
  - Handoff date: __________
- [ ] **Open PRs/issues reassigned**

### 3. Asset Recovery
- [ ] **Company laptop returned**
  - Asset tag: __________
  - Date returned: __________
  - Condition: ☐ Good ☐ Damaged ☐ Missing
- [ ] **Security keys/hardware tokens returned**
  - Items: ____________________
  - Date returned: __________
- [ ] **Access badges/keycards returned**
- [ ] **Company credit cards cancelled**
- [ ] **Mobile devices wiped** (if BYOD with company data)

---

## Day of Departure: Access Revocation

### 4. Identity Provider (CRITICAL - DO FIRST)
- [ ] **Google Workspace account disabled/suspended**
  - Disabled by: __________
  - Time: __________
  - Account preserved for: ☐ 30 days ☐ 90 days ☐ Other: ____
- [ ] **SSO sessions terminated**
- [ ] **Password reset** (prevents cached credential use)
- [ ] **MFA devices removed**

### 5. Source Control (GitHub)
- [ ] **Removed from GitHub organization**
  - Removed by: __________
  - Time: __________
- [ ] **Removed from all teams**
- [ ] **Personal access tokens revoked** (if visible in admin)
- [ ] **Deploy keys reviewed** (remove any created by this user)
- [ ] **Webhook secrets rotated** (if user had access)

### 6. Cloud Provider (AWS)
- [ ] **IAM user disabled/deleted**
  - Action: ☐ Disabled ☐ Deleted
  - Performed by: __________
  - Time: __________
- [ ] **Access keys deactivated**
  - Access Key ID(s): ____________________
- [ ] **Console password deleted**
- [ ] **MFA device deactivated**
- [ ] **IAM roles reviewed** (remove any trust relationships to user)
- [ ] **Secrets Manager access reviewed**
  - Secrets rotated: ☐ Yes ☐ N/A

### 7. Communication Tools
- [ ] **Slack/Discord access revoked**
  - Time: __________
- [ ] **Email forwarding removed** (or set up approved forward)
- [ ] **Shared mailbox access revoked**
- [ ] **Calendar delegate access removed**

### 8. Development Tools
- [ ] **CI/CD secrets rotated** (if user had access)
  - GitHub Actions secrets: ☐ Reviewed ☐ Rotated
  - Environment variables: ☐ Reviewed ☐ Rotated
- [ ] **Container registry access revoked**
- [ ] **Artifact repository access revoked**
- [ ] **Database credentials rotated** (if direct access)

### 9. Third-Party Services
- [ ] **SaaS application access revoked**
  - Applications: ____________________
- [ ] **Vendor portal access removed**
- [ ] **Customer-facing tool access removed**

---

## Post-Departure Verification

### 10. Access Verification (Within 24 Hours)
- [ ] **Login attempt audit** - Verify no successful logins after termination
  - Logs reviewed by: __________
  - Date: __________
  - Findings: ☐ Clean ☐ Suspicious activity (escalate)
- [ ] **All access revocations verified**
  - Verified by: __________
  - Date: __________

### 11. Secrets Rotation (Within 7 Days)
- [ ] **Shared credentials rotated** (any the user knew)
  - List: ____________________
- [ ] **API keys regenerated** (any the user created)
- [ ] **Service account passwords rotated** (if user had access)
- [ ] **Encryption keys rotated** (if privileged user)

### 12. Post-Departure Monitoring (30 Days)
- [ ] **Monitor for attempted access**
  - Monitoring enabled: ☐ Yes
  - Alert configured: ☐ Yes
- [ ] **Review audit logs weekly**
  - Week 1 reviewed: __________
  - Week 2 reviewed: __________
  - Week 3 reviewed: __________
  - Week 4 reviewed: __________

---

## Special Procedures

### For Privileged Users (Admin Access)
Additional steps required:
- [ ] **Emergency access procedures updated**
- [ ] **Root account credentials rotated** (if shared)
- [ ] **Infrastructure access keys rotated**
- [ ] **DNS/domain registrar access reviewed**
- [ ] **Certificate private keys rotated** (if accessible)
- [ ] **Backup encryption keys rotated**

### For Security Incidents
If offboarding due to security concern:
- [ ] **Legal counsel notified**
- [ ] **Forensic image taken before wiping devices**
- [ ] **Access logs preserved**
- [ ] **Incident report filed**
  - Incident ID: __________

### For Contractors
Additional verification:
- [ ] **NDA remains in effect** (confirm end date)
- [ ] **Subcontractor access revoked** (if applicable)
- [ ] **Client-specific access revoked**

---

## Completion Sign-Off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Direct Manager | | | |
| IT/Security | Carlos Sanchez | | |
| HR (if applicable) | | | |

---

## Evidence Storage

All offboarding documentation should be stored in:
```
../aegis-compliance-evidence/soc2/[YEAR]/hr/offboarding/[EMPLOYEE_ID]/
├── separation-agreement.pdf
├── asset-return-form.pdf
├── access-revocation-log.csv
├── final-access-audit.pdf
└── offboarding-checklist-signed.pdf
```

---

## Access Revocation Log Template

| System | Account/Username | Revocation Time | Performed By | Verified |
|--------|------------------|-----------------|--------------|----------|
| Google Workspace | | | | ☐ |
| GitHub | | | | ☐ |
| AWS IAM | | | | ☐ |
| Slack | | | | ☐ |
| [Add others] | | | | ☐ |

---

## Emergency Contact

If you need to perform emergency offboarding outside business hours:

**Primary:** Carlos Sanchez - [phone] / [email]

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Claude (SOC2 Engineer) | Initial version |
