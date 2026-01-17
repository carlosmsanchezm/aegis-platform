# Employee Onboarding Checklist

**SOC 2 Control:** CC6.2 (Prior to Issuing System Credentials and Granting System Access)
**Owner:** Carlos Sanchez (Founder)
**Last Updated:** 2026-01-17
**Review Frequency:** Annually or upon process change

---

## Purpose

This checklist ensures all new personnel are properly vetted, trained, and granted appropriate access before being given system credentials. It supports SOC 2 compliance by documenting the authorization and provisioning process.

---

## Pre-Employment (Before Day 1)

### 1. Background Verification
- [ ] **Background check completed** (if applicable based on role/jurisdiction)
  - Vendor: _________________
  - Date completed: __________
  - Result: ☐ Clear ☐ Review required
- [ ] **Reference checks completed** (minimum 2 professional references)
  - Reference 1: _____________ Date: _______
  - Reference 2: _____________ Date: _______
- [ ] **Right to work verification** (I-9 or equivalent)

### 2. Employment Documentation
- [ ] **Offer letter signed**
  - Date signed: __________
- [ ] **Employment agreement signed**
  - Date signed: __________
- [ ] **Confidentiality/NDA signed**
  - Date signed: __________
  - Document location: `../aegis-compliance-evidence/soc2/[YEAR]/hr/nda/`
- [ ] **Acceptable Use Policy acknowledged**
  - Date acknowledged: __________
- [ ] **Information Security Policy acknowledged**
  - Date acknowledged: __________

### 3. Access Request Preparation
- [ ] **Role and access requirements documented**
  - Job title: ____________________
  - Department: __________________
  - Manager: ____________________
  - Access level: ☐ Standard ☐ Privileged ☐ Admin
- [ ] **Manager approval for access obtained**
  - Approver: ____________________
  - Date approved: __________

---

## Day 1: Account Provisioning

### 4. Identity Provider Setup
- [ ] **Google Workspace account created** (when implemented)
  - Email: ____________________@aegis.dev
  - Created by: __________
  - Date: __________
- [ ] **MFA enrolled**
  - Method: ☐ Authenticator App ☐ Security Key ☐ Phone
  - Verified by: __________
  - Date: __________

### 5. GitHub Access
- [ ] **GitHub account linked/invited**
  - GitHub username: __________________
  - Organization: carlosmsanchezm (or future org)
  - Team(s): ____________________
  - Role: ☐ Read ☐ Write ☐ Maintain ☐ Admin
- [ ] **Repository access granted** (principle of least privilege)
  - Repositories: ____________________
  - Approved by: __________
  - Date: __________

### 6. AWS Access (If Required)
- [ ] **IAM user created** (only if programmatic access needed)
  - Username: ____________________
  - Console access: ☐ Yes ☐ No
  - Programmatic access: ☐ Yes ☐ No
- [ ] **MFA enabled on AWS account**
  - Method: ☐ Virtual MFA ☐ Hardware token
  - Date enabled: __________
- [ ] **IAM policies attached** (least privilege)
  - Policies: ____________________
  - Approved by: __________
- [ ] **Access keys rotated within 90 days reminder set**

### 7. Communication Tools
- [ ] **Slack/Discord access** (if applicable)
  - Channels: ____________________
- [ ] **Calendar access granted**
- [ ] **Internal documentation access** (Notion/Confluence/wiki)

---

## Week 1: Security Training

### 8. Required Training Completion
- [ ] **Security Awareness Training completed**
  - Training provider/course: ____________________
  - Completion date: __________
  - Certificate location: `../aegis-compliance-evidence/soc2/[YEAR]/training-policy-ack/`
- [ ] **Secure Development Training** (for engineers)
  - Course: ____________________
  - Completion date: __________
- [ ] **Incident Response Procedures reviewed**
  - Acknowledged understanding: ☐ Yes
  - Date: __________
- [ ] **Data Classification Training** (if handling customer data)
  - Completion date: __________

### 9. Policy Acknowledgments
- [ ] **Employee Handbook acknowledged**
- [ ] **Code of Conduct signed**
- [ ] **Data Protection Policy acknowledged**
- [ ] **Remote Work Policy acknowledged** (if applicable)

---

## Access Verification (Within 30 Days)

### 10. Manager Verification
- [ ] **Manager confirms access is appropriate**
  - Verifier: ____________________
  - Date verified: __________
  - Access changes needed: ☐ No ☐ Yes (document below)
    - Changes: ____________________

### 11. Security Team Verification
- [ ] **Access logs reviewed for anomalies**
  - Reviewed by: __________
  - Date: __________
  - Findings: ☐ None ☐ Documented

---

## Completion Sign-Off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| New Employee | | | |
| Direct Manager | | | |
| Security/Compliance | Carlos Sanchez | | |

---

## Evidence Storage

All onboarding documentation should be stored in:
```
../aegis-compliance-evidence/soc2/[YEAR]/hr/onboarding/[EMPLOYEE_ID]/
├── background-check-summary.pdf
├── signed-nda.pdf
├── signed-offer-letter.pdf
├── access-approval-form.pdf
├── training-certificates/
│   ├── security-awareness.pdf
│   └── secure-development.pdf
└── onboarding-checklist-signed.pdf
```

---

## Quick Reference: Access by Role

| Role | GitHub | AWS Console | AWS CLI | Admin Access |
|------|--------|-------------|---------|--------------|
| Founder/CEO | Admin | Root + IAM | Yes | Full |
| Senior Engineer | Maintain | IAM User | Yes | Limited |
| Engineer | Write | IAM User | Yes | None |
| Contractor | Read/Write | None | Limited | None |
| Support | Read | None | None | None |

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Claude (SOC2 Engineer) | Initial version |
