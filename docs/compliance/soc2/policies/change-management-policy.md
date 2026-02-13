# Change Management Policy

**Policy ID:** CMP-001
**Version:** 1.0
**Effective Date:** 2026-01-17
**Next Review:** 2027-01-17
**Owner:** Carlos Sanchez (Founder)
**Approved By:** Carlos Sanchez (Founder)

## 1. Purpose

This policy establishes requirements for managing changes to Aegis Technologies information systems to ensure changes are authorized, tested, and implemented in a controlled manner.

## 2. Scope

This policy applies to:
- Production systems and infrastructure
- Application code and configurations
- Database schemas and data
- Network configurations
- Security controls

## 3. Policy Statements

### 3.1 Change Categories

| Category | Description | Approval | Examples |
|----------|-------------|----------|----------|
| Standard | Pre-approved, low risk | Auto-approved | Dependency updates, minor configs |
| Normal | Requires review | Peer + Lead | Feature releases, infra changes |
| Emergency | Urgent fixes | Post-hoc | Security patches, outage fixes |

### 3.2 Change Process Requirements

All changes must include:
- Description of the change
- Business justification
- Risk assessment
- Rollback plan
- Testing evidence
- Approval documentation

### 3.3 Development Standards

- All code changes require pull request
- Minimum one peer review required
- Automated tests must pass
- Security scanning must pass
- No direct commits to main branch

### 3.4 Testing Requirements

| Environment | Purpose | Required Before |
|-------------|---------|-----------------|
| Development | Feature development | Code review |
| Staging | Integration testing | Production deploy |
| Production | Live service | N/A |

### 3.5 Deployment Requirements

- Deployments during maintenance windows (preferred)
- Deployment playbook/runbook required
- Rollback capability verified
- Monitoring active during deployment
- On-call engineer available

### 3.6 Emergency Changes

Emergency changes may bypass normal process when:
- Security vulnerability being actively exploited
- Production outage affecting customers
- Regulatory compliance deadline

Emergency changes require:
- Verbal approval from authorized approver
- Post-implementation documentation within 24 hours
- Retrospective review

### 3.7 Segregation of Duties

> **Solo Founder Note:** As a one-person company, full segregation of duties is not currently possible. Compensating controls:
> - All changes require PR (even self-merged) to maintain audit trail
> - GitHub branch protection enforced (no force push)
> - CloudTrail logging for all AWS actions
> - This limitation is documented and will be addressed as team grows

When team size permits:
- Developers cannot approve their own changes
- Developers do not have production access
- Deployments require separate approval from development

## 4. Procedures

### 4.1 Standard Change Procedure

1. Developer creates pull request
2. Automated tests run
3. Peer reviews and approves
4. Changes merged to main
5. Automated deployment to staging
6. Manual promotion to production

### 4.2 Normal Change Procedure

1. Change request created with details
2. Risk assessment completed
3. Testing plan documented
4. Change reviewed by peers
5. Engineering Lead approves
6. Scheduled deployment
7. Post-deployment verification

### 4.3 Emergency Change Procedure

1. Incident declared
2. Fix developed and tested minimally
3. Verbal approval obtained
4. Change deployed
5. Change documented within 24 hours
6. Retrospective conducted

## 5. Roles and Responsibilities

> **Note:** Aegis is currently a solo-founder company. Roles will be assigned as the team grows.

| Role | Current Owner | Responsibilities |
|------|---------------|-----------------|
| Developer | Carlos Sanchez | Develop changes, create PRs, document |
| Peer Reviewer | Carlos Sanchez (+ Dependabot/CI) | Code review, test verification |
| Engineering Lead | Carlos Sanchez | Approve normal changes, risk assessment |
| On-call Engineer | Carlos Sanchez | Emergency approvals, deployment support |
| Security | Carlos Sanchez | Security review for sensitive changes |

## 6. Tools and Systems

| Tool | Purpose |
|------|---------|
| GitHub | Code repository, pull requests |
| GitHub Actions | CI/CD, automated testing |
| Helm/ArgoCD | Kubernetes deployments |
| Jira/Linear | Change tracking |
| Slack | Change notifications |

## 7. Compliance Monitoring

- All changes logged in version control
- Deployment logs retained 90+ days
- Monthly change metrics reviewed
- Unauthorized changes investigated

## 8. Related Documents

- [Information Security Policy](./information-security-policy.md)
- [Incident Response Policy](./incident-response-policy.md)
- Deployment Runbook (link)
- Rollback Procedures (link)

## 9. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial approved version |

---

## 10. Policy Acknowledgment

| Name | Role | Date | Signature |
|------|------|------|-----------|
| Carlos Sanchez | Founder | 2026-01-17 | /s/ Carlos Sanchez |
