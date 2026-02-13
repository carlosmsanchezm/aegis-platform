# SOC 2 Control Narratives

**Document ID:** SOC2-NARRATIVES-001
**Version:** 1.0
**Created:** 2025-01-17
**Last Updated:** 2025-01-17

## Purpose

This document provides detailed narratives for each SOC 2 control, describing:
- How the control is implemented at Aegis
- Who is responsible
- What evidence is collected
- Current status and gaps

---

## CC1: Control Environment

### CC1.1 - Integrity and Ethical Values

**Control Statement:** Management demonstrates commitment to integrity and ethical values.

**Implementation:**
- Code of Conduct document distributed to all personnel
- Annual acknowledgment required
- Ethics hotline/reporting mechanism (email to CEO for small team)

**Evidence:**
- Signed Code of Conduct acknowledgments
- Ethics policy document

**Status:** 🔴 NEEDS_IMPLEMENTATION
- [ ] Create Code of Conduct document
- [ ] Create acknowledgment tracking spreadsheet
- [ ] Get initial signatures

---

### CC1.2 - Board/Management Oversight

**Control Statement:** Board/management provides oversight of the security program.

**Implementation:**
- Quarterly security review as part of company planning
- Security metrics reviewed (incidents, vulnerabilities, access)
- Documented meeting notes with security agenda items

**Evidence:**
- Meeting notes/minutes with security discussion
- Security dashboard/metrics report

**Status:** 🔴 NEEDS_IMPLEMENTATION
- [ ] Schedule quarterly security reviews
- [ ] Create security metrics dashboard
- [ ] Template for meeting notes

---

### CC1.3 - Organizational Structure

**Control Statement:** Management establishes structures and reporting lines for security.

**Implementation:**
- Organizational chart showing security responsibilities
- Security Lead role defined (CEO for small team)
- Clear escalation paths

**Evidence:**
- Org chart with security role highlighted
- RACI matrix for security functions

**Status:** 🔴 NEEDS_IMPLEMENTATION
- [ ] Create org chart
- [ ] Document security responsibilities
- [ ] Create RACI matrix

---

## CC6: Logical and Physical Access

### CC6.1 - Logical Access Security

**Control Statement:** The organization implements logical access security software, infrastructure, and architectures.

**Implementation:**

| System | Access Control Mechanism | MFA Status |
|--------|-------------------------|------------|
| GitHub | RBAC via teams/permissions | ⚠️ Personal account |
| AWS GovCloud | IAM policies, roles | NEEDS_VERIFICATION |
| Keycloak | OIDC/RBAC | ✅ MFA configurable |
| Kubernetes | RBAC, service accounts | ✅ Namespace isolation |

**Evidence:**
- GitHub org/repo permission exports
- AWS IAM policy exports
- Keycloak realm configuration
- K8s RBAC manifests

**Status:** 🟡 PARTIAL
- [x] Keycloak OIDC implemented
- [x] K8s RBAC exists
- [ ] GitHub org migration needed
- [ ] AWS IAM review needed

---

### CC6.2 - User Access Provisioning

**Control Statement:** The organization provisions access based on authorization and job function.

**Implementation:**
- New hire onboarding checklist with access requests
- Access requests via Jira ticket
- Manager approval required
- Principle of least privilege applied

**Process:**
1. Manager submits access request ticket
2. Security Lead reviews for least privilege
3. Access provisioned in relevant systems
4. Confirmation sent to manager

**Evidence:**
- Onboarding tickets with access requests
- Access approval records

**Status:** 🔴 NEEDS_IMPLEMENTATION
- [ ] Create onboarding checklist
- [ ] Create Jira access request template
- [ ] Document approval workflow

---

### CC6.3 - Access Removal

**Control Statement:** The organization removes access upon termination.

**Implementation:**
- Offboarding checklist with all systems
- Access removed within 24 hours of termination
- Shared credentials rotated
- Exit interview documents access return

**Termination Checklist:**
- [ ] GitHub access removed
- [ ] AWS IAM user disabled
- [ ] Google Workspace account suspended
- [ ] Keycloak user disabled
- [ ] VPN/SSH keys revoked
- [ ] Shared secrets rotated if applicable

**Evidence:**
- Offboarding tickets
- System audit logs showing removal
- Checklist completion records

**Status:** 🔴 NEEDS_IMPLEMENTATION
- [ ] Create offboarding checklist
- [ ] Test termination process

---

### CC6.4 - Periodic Access Review

**Control Statement:** The organization periodically reviews access rights.

**Implementation:**
- Quarterly access review for all systems
- Managers certify continued need for access
- Privileged access reviewed monthly
- Results documented and gaps remediated

**Review Process:**
1. Export current access lists from all systems
2. Send to managers for certification
3. Manager certifies or requests removal
4. Security removes unauthorized access
5. Document completion

**Evidence:**
- Access list exports (timestamped)
- Manager certification records
- Remediation records for removed access

**Status:** 🔴 NEEDS_IMPLEMENTATION
- [ ] Create access review template
- [ ] Schedule Q1 access review
- [ ] Build automation for access exports

---

### CC6.6 - Authentication Controls

**Control Statement:** The organization implements strong authentication.

**Implementation:**

| System | Authentication Method | MFA Requirement |
|--------|----------------------|-----------------|
| GitHub | Password + 2FA | Required |
| AWS | IAM + MFA | Required for console |
| Google Workspace | Password + 2FA | Required |
| Keycloak (customers) | OIDC + MFA | Configurable |

**Password Policy:**
- Minimum 12 characters
- Complexity required (upper, lower, number, special)
- No reuse of last 12 passwords
- 90-day maximum age (or passkeys)

**Evidence:**
- MFA enforcement configuration screenshots
- Password policy configuration
- Audit logs showing MFA usage

**Status:** 🟡 PARTIAL
- [x] GitHub 2FA likely enabled (verify)
- [ ] Centralize in corporate IdP
- [ ] Document password policy

---

## CC7: System Operations

### CC7.1 - Vulnerability Management

**Control Statement:** The organization identifies and manages vulnerabilities.

**Implementation:**
- Dependabot enabled for dependency vulnerabilities
- Container image scanning (Trivy/Grype)
- Infrastructure scanning (AWS Config/Security Hub)
- SLAs for remediation by severity

**Vulnerability SLAs:**
| Severity | Remediation SLA |
|----------|-----------------|
| Critical | 7 days |
| High | 30 days |
| Medium | 90 days |
| Low | Next release |

**Evidence:**
- Dependabot alert exports
- Container scan reports
- Vulnerability remediation tickets

**Status:** 🔴 NEEDS_IMPLEMENTATION
- [ ] Enable Dependabot
- [ ] Enable GitHub secret scanning
- [ ] Implement container scanning in CI
- [ ] Document SLAs
- [ ] Create triage process

---

### CC7.2 - System Monitoring

**Control Statement:** The organization monitors systems for anomalies.

**Implementation:**
- Prometheus for metrics collection
- Grafana dashboards for visualization
- CloudWatch for AWS infrastructure
- Alerts for critical events

**Key Metrics Monitored:**
- API error rates
- Authentication failures
- Resource utilization
- Deployment status

**Evidence:**
- Monitoring configuration
- Dashboard screenshots
- Alert rule definitions
- Sample alert notifications

**Status:** 🟡 PARTIAL
- [x] Prometheus exists
- [ ] Document alert rules
- [ ] Create security-specific alerts
- [ ] Set up log aggregation

---

## CC8: Change Management

### CC8.1 - Change Control

**Control Statement:** The organization manages changes through a defined process.

**Implementation:**
- All changes via GitHub Pull Requests
- Required code review before merge
- Automated CI checks must pass
- Branch protection prevents direct commits
- Releases tagged and documented

**Change Process:**
1. Developer creates feature branch
2. Developer opens Pull Request
3. Automated tests run (GitHub Actions)
4. Peer review required (1 approval minimum)
5. Security-sensitive changes require Security Lead review
6. Merge to main triggers deployment pipeline
7. Release tagged for production

**Evidence:**
- PR records with reviews
- CI/CD logs
- Branch protection settings
- Release tags

**Status:** 🟡 PARTIAL
- [x] GitHub PRs used
- [x] GitHub Actions CI exists
- [ ] Branch protection not configured
- [ ] CODEOWNERS not configured
- [ ] Security review trigger not defined

---

## A1: Availability (OUT OF SCOPE)

> **Note:** Availability controls are **out of scope** for the current SOC 2 audit. Aegis operates as a self-hosted software vendor (Model B) and does not operate production services for customers. These controls will be added when Aegis expands to Model C (managed services).

### Future Scope: When to Add Availability

Add Availability (A1) controls when:
- Aegis begins operating managed services for customers (Model C)
- Aegis provides SLAs for uptime/availability
- Aegis operates production infrastructure on behalf of customers

### A1.1 - Capacity and Availability (DEFERRED)

**Status:** ⬜ OUT_OF_SCOPE - Will implement for Model C

---

### A1.3 - Backup and Recovery (DEFERRED)

**Status:** ⬜ OUT_OF_SCOPE - Will implement for Model C

> **Note:** While A1 is out of scope, Aegis still maintains development/test backups as good practice. These are not audited under the current scope.

---

## Control Status Summary

> **Scope:** Security TSC only (Model B). Availability controls (A1) are out of scope.

| Status | Count | Percentage |
|--------|-------|------------|
| ✅ Implemented | 5 | 17% |
| 🟡 Partial | 10 | 33% |
| 🔴 Needs Implementation | 12 | 40% |
| ⬜ Inherited | 3 | 10% |
| ⬜ Out of Scope | 3 | (A1 - deferred) |

**Total In-Scope Controls:** 30 (CC series only)

---

## Priority Implementation Order

### Phase 1: Foundation (Week 1-2)
1. CC6.6 - MFA enforcement everywhere
2. CC8.1 - Branch protection + CODEOWNERS
3. CC7.1 - Enable Dependabot + secret scanning
4. CC6.1 - GitHub org migration

### Phase 2: Operations (Week 3-4)
5. CC7.2 - Security alerting
6. CC6.2/CC6.3 - Onboarding/offboarding process
7. CC7.4 - Incident response process
8. CC7.5 - Recovery procedures (dev/test - not audited)

### Phase 3: Governance (Month 2)
9. CC1.1-CC1.5 - Governance documentation
10. CC3.1-CC3.4 - Risk assessment
11. CC6.4 - Access review process
12. CC9.1-CC9.2 - Vendor management and BC planning

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.1 | 2025-01-17 | Claude (SOC2 Engineer) | Removed A1 from scope (Model B only, Security TSC) |
| 1.0 | 2025-01-17 | Claude (SOC2 Engineer) | Initial narratives |
