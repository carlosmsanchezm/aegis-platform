# Quarterly Access Review Procedure

**SOC 2 Control:** CC6.4 (Restricting and Evaluating Access)
**Owner:** Carlos Sanchez (Founder)
**Last Updated:** 2026-01-17
**Review Frequency:** Quarterly

---

## Purpose

This procedure ensures that access to systems and data is reviewed quarterly to verify:
1. Access remains appropriate for each user's current role
2. Terminated users have been fully de-provisioned
3. Privileged access is limited to those who need it
4. Dormant accounts are identified and addressed

---

## Schedule

| Quarter | Review Period | Due Date | Evidence Location |
|---------|---------------|----------|-------------------|
| Q1 | Jan 1 - Mar 31 | Apr 15 | `soc2/[YEAR]/Q1/access-reviews/` |
| Q2 | Apr 1 - Jun 30 | Jul 15 | `soc2/[YEAR]/Q2/access-reviews/` |
| Q3 | Jul 1 - Sep 30 | Oct 15 | `soc2/[YEAR]/Q3/access-reviews/` |
| Q4 | Oct 1 - Dec 31 | Jan 15 | `soc2/[YEAR]/Q4/access-reviews/` |

---

## Systems In Scope

| System | Review Type | Reviewer | Evidence Export |
|--------|-------------|----------|-----------------|
| GitHub | Full user list + permissions | Founder | `export_github_security_baseline.sh` |
| AWS IAM | Users, roles, policies | Founder | `export_aws_security_baseline.sh` |
| Google Workspace | User accounts + groups | Founder | Admin Console export |
| Slack | Members + channels | Founder | Admin export |
| Third-party SaaS | Active users | Founder | Manual or API |

---

## Review Procedure

### Step 1: Export Current Access (Week 1 of Review Period)

Run the evidence collection scripts to capture current state:

```bash
# Set variables
VAULT_PATH="../aegis-compliance-evidence"
QUARTER="Q1"  # Adjust per quarter
YEAR=$(date +%Y)

# Export GitHub access
GITHUB_OWNER=carlosmsanchezm GITHUB_REPOS=aegis-platform \
    ./scripts/compliance/export_github_security_baseline.sh \
    ${VAULT_PATH}/soc2/${YEAR}/${QUARTER}/access-reviews/github/

# Export AWS access
AWS_PROFILE=aegis-new \
    ./scripts/compliance/export_aws_security_baseline.sh \
    ${VAULT_PATH}/soc2/${YEAR}/${QUARTER}/access-reviews/aws/
```

- [ ] GitHub user list exported
- [ ] GitHub repository permissions exported
- [ ] AWS IAM users exported
- [ ] AWS IAM roles exported
- [ ] AWS access key age report generated
- [ ] Google Workspace users exported (when implemented)

---

### Step 2: Generate User Access Matrix

Create a consolidated view of all user access:

| User | Role | GitHub | AWS | Google | Slack | Last Activity | Status |
|------|------|--------|-----|--------|-------|---------------|--------|
| carlos@aegis.dev | Founder | Admin | Root+IAM | Admin | Admin | Today | Active |
| [contractor] | Developer | Write | IAM User | - | Member | 30d ago | Review |

Template file: `../aegis-compliance-evidence/soc2/templates/access-review-template.md`

---

### Step 3: Review Each User (Week 2)

For each user, verify:

#### 3.1 Employment Status
- [ ] **User is still employed/contracted**
  - Cross-reference with HR records or contractor agreements
  - Action if terminated: Immediate offboarding (see offboarding checklist)

#### 3.2 Role Appropriateness
- [ ] **Access matches current role**
  - Compare access level to job responsibilities
  - Action if excessive: Submit access reduction request

#### 3.3 Privilege Review
- [ ] **Privileged access is justified**
  - Admin/root access documented and approved
  - Action if unjustified: Escalate for removal

#### 3.4 Activity Check
- [ ] **Account has been used in last 90 days**
  - Review last login dates
  - Action if dormant: Disable account, notify user

#### 3.5 MFA Verification
- [ ] **MFA is enabled on all accounts**
  - Check MFA status in each system
  - Action if missing: Require immediate MFA enrollment

---

### Step 4: Review Service Accounts & API Keys

#### 4.1 GitHub
- [ ] **Deploy keys reviewed**
  - List all deploy keys and their purpose
  - Remove any unused keys
- [ ] **Personal access tokens reviewed** (if visible)
  - Check for tokens older than 90 days
- [ ] **GitHub App permissions reviewed**

#### 4.2 AWS
- [ ] **IAM access keys reviewed**
  - Identify keys older than 90 days
  - Rotate or remove as needed
- [ ] **IAM roles reviewed**
  - Verify trust policies are appropriate
  - Remove unused roles
- [ ] **Service-linked roles reviewed**

#### 4.3 Third-Party Integrations
- [ ] **OAuth tokens reviewed**
  - List all connected applications
  - Revoke unused authorizations
- [ ] **API keys in secrets manager reviewed**
  - Verify all keys are still needed
  - Rotate any keys older than 1 year

---

### Step 5: Document Findings (Week 3)

Complete the Access Review Summary:

```markdown
# Q[X] [YEAR] Access Review Summary

**Review Period:** [Start Date] - [End Date]
**Reviewer:** Carlos Sanchez
**Review Date:** [Date]

## Summary Statistics

| Metric | Count |
|--------|-------|
| Total users reviewed | X |
| Access appropriate | X |
| Access modified | X |
| Accounts disabled | X |
| Dormant accounts found | X |
| MFA gaps found | X |

## Findings

### Users with Excessive Access
| User | System | Current | Should Be | Action Taken |
|------|--------|---------|-----------|--------------|
| - | - | - | - | - |

### Dormant Accounts (>90 days inactive)
| User | System | Last Activity | Action Taken |
|------|--------|---------------|--------------|
| - | - | - | - |

### MFA Gaps
| User | System | Action Taken |
|------|--------|--------------|
| - | - | - |

### Service Account Issues
| Account | System | Issue | Action Taken |
|---------|--------|-------|--------------|
| - | - | - | - |

## Actions Taken

1. [Action 1]
2. [Action 2]

## Follow-Up Items

| Item | Owner | Due Date | Status |
|------|-------|----------|--------|
| - | - | - | - |

## Sign-Off

| Role | Name | Date |
|------|------|------|
| Reviewer | Carlos Sanchez | |
| Approver | Carlos Sanchez | |
```

---

### Step 6: Commit Evidence (Week 4)

```bash
cd ../aegis-compliance-evidence

# Verify no secrets in files
grep -r "AKIA" soc2/${YEAR}/${QUARTER}/access-reviews/ && echo "WARNING: Possible AWS key!"
grep -r "ghp_" soc2/${YEAR}/${QUARTER}/access-reviews/ && echo "WARNING: Possible GitHub token!"

# Commit evidence
git add soc2/${YEAR}/${QUARTER}/access-reviews/
git commit -m "SOC2 evidence: ${QUARTER} ${YEAR} access review completed"
git push
```

- [ ] All exports saved to evidence vault
- [ ] Access review summary completed
- [ ] No secrets in committed files
- [ ] Evidence committed and pushed

---

## Review Checklist Summary

### Quick Checklist (Copy for Each Review)

```
Q[X] [YEAR] Access Review Checklist

Date: __________
Reviewer: __________

[ ] GitHub user access exported
[ ] GitHub permissions verified appropriate
[ ] AWS IAM users exported
[ ] AWS access keys <90 days old (or documented exception)
[ ] AWS MFA enabled for all users
[ ] Google Workspace reviewed (when implemented)
[ ] Dormant accounts (>90 days) addressed
[ ] Service accounts reviewed
[ ] Excessive privileges remediated
[ ] Findings documented
[ ] Evidence committed to vault
[ ] Review summary signed off
```

---

## Escalation Procedures

| Finding | Escalation |
|---------|------------|
| Terminated user still has access | Immediate revocation, incident report |
| Unauthorized admin access | Immediate removal, investigation |
| Shared credentials discovered | Immediate rotation, policy reminder |
| MFA not enabled | 24-hour deadline to enable |
| Access key >180 days old | Rotation within 7 days |

---

## Evidence Storage Structure

```
../aegis-compliance-evidence/soc2/[YEAR]/
├── Q1/
│   └── access-reviews/
│       ├── github/
│       │   ├── YYYY-MM-DD_github_collaborators.json
│       │   ├── YYYY-MM-DD_github_repo_settings.json
│       │   └── YYYY-MM-DD_github_branch_protection.json
│       ├── aws/
│       │   ├── YYYY-MM-DD_aws_iam_users.json
│       │   ├── YYYY-MM-DD_aws_iam_roles.json
│       │   └── YYYY-MM-DD_aws_access_key_report.json
│       ├── Q1-access-review-summary.md
│       └── Q1-access-review-checklist-signed.pdf
├── Q2/
├── Q3/
└── Q4/
```

---

## Automation Opportunities

### Future Improvements
1. **Automated dormant account detection** - CloudWatch alarm on last login
2. **Automated key age alerts** - Lambda to check key ages weekly
3. **Slack/email reminders** - Scheduled reminders before due dates
4. **Dashboard** - Grafana dashboard showing access metrics

### Scripts Available
- `scripts/compliance/export_github_security_baseline.sh`
- `scripts/compliance/export_aws_security_baseline.sh`

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Claude (SOC2 Engineer) | Initial version |
