# Leadership, Policy, and Roles

**Document ID:** ISMS-LEAD-001
**Version:** 1.0
**Effective Date:** 2026-01-17
**Next Review:** 2027-01-17
**Owner:** Carlos Sanchez (Top Management)
**Classification:** Internal

---

## 1. Purpose

This document fulfills ISO/IEC 27001:2022 Clause 5 requirements for leadership commitment, information security policy, and organizational roles, responsibilities, and authorities.

---

## 2. Leadership Commitment (Clause 5.1)

### 2.1 Top Management Statement

As founder and sole operator of Aegis Technologies, I, Carlos Sanchez, commit to:

1. **Policy Alignment**: Ensuring the information security policy and objectives are compatible with the strategic direction of Aegis Technologies

2. **ISMS Integration**: Integrating ISMS requirements into business processes, particularly the software development lifecycle

3. **Resource Provision**: Ensuring resources needed for the ISMS are available within budget constraints

4. **Communication**: Communicating the importance of effective information security and conforming to ISMS requirements

5. **Achieving Outcomes**: Ensuring the ISMS achieves its intended outcomes - secure software and customer trust

6. **Directing and Supporting**: Directing and supporting persons to contribute to ISMS effectiveness

7. **Continual Improvement**: Promoting continual improvement of the ISMS

8. **Supporting Roles**: Supporting other relevant management roles (future hires) to demonstrate leadership

### 2.2 Evidence of Commitment

| Commitment | Evidence | Location |
|------------|----------|----------|
| Policy approval | Signed policies | soc2/policies/, iso27001/policies/ |
| Resource allocation | Budget for tools, audits | Financial records |
| Risk ownership | Risk register approval | 03-risk-register.csv |
| ISMS operation | Evidence collection | aegis-compliance-evidence/ |

---

## 3. Information Security Policy (Clause 5.2)

### 3.1 Policy Statement

Aegis Technologies is committed to protecting the confidentiality, integrity, and availability of information assets. We achieve this by:

- Implementing security controls appropriate to identified risks
- Complying with legal, regulatory, and contractual obligations
- Providing secure software products to our customers
- Continuously improving our information security practices

### 3.2 Policy Objectives

| Objective | Metric | Target | Measurement |
|-----------|--------|--------|-------------|
| Secure development | Critical vulnerabilities | 0 open > 7 days | Weekly Dependabot review |
| Access control | MFA coverage | 100% | Quarterly access review |
| Incident response | Mean time to respond | < 4 hours P1 | Incident log |
| Compliance | Audit findings | 0 critical | Annual audit |

### 3.3 Policy Documentation

| Policy | ID | ISO Control | Approval Status |
|--------|-----|-------------|-----------------|
| Information Security Policy | ISP-001 | A.5.1 | ✅ Approved 2026-01-17 |
| Access Control Policy | ACP-001 | A.5.15-18 | ✅ Approved 2026-01-17 |
| Change Management Policy | CMP-001 | A.8.32 | ✅ Approved 2026-01-17 |
| Incident Response Policy | IRP-001 | A.5.24-28 | ✅ Approved 2026-01-17 |
| Risk Management Policy | RMP-001 | Clause 6.1 | ✅ Approved 2026-01-17 |
| Vendor Management Policy | VMP-001 | A.5.19-23 | ✅ Approved 2026-01-17 |
| Acceptable Use Policy | ISMS-POL-AUP-001 | A.5.10 | ✅ Approved 2026-01-17 |

---

## 4. Organizational Roles, Responsibilities, and Authorities (Clause 5.3)

### 4.1 Current Organization

```
┌─────────────────────────────────────┐
│         Carlos Sanchez              │
│      Founder / CEO                  │
│                                     │
│  Roles:                             │
│  • Top Management                   │
│  • Information Security Manager     │
│  • Risk Owner                       │
│  • Engineering Lead                 │
│  • Incident Commander               │
└─────────────────────────────────────┘
```

### 4.2 Role Definitions

#### Top Management (Carlos Sanchez)

| Responsibility | Authority | Accountability |
|----------------|-----------|----------------|
| Establish ISMS policy | Approve policies | ISMS effectiveness |
| Allocate resources | Budget decisions | Resource adequacy |
| Assign roles | Delegate authority | Role performance |
| Conduct management review | Final decisions | Improvement actions |

#### Information Security Manager (Carlos Sanchez)

| Responsibility | Authority | Accountability |
|----------------|-----------|----------------|
| Maintain ISMS documentation | Update all docs | Documentation currency |
| Coordinate risk assessments | Lead assessments | Risk identification |
| Manage security incidents | Incident decisions | Incident resolution |
| Monitor compliance | Access all systems | Control effectiveness |
| Report to top management | Escalate issues | ISMS performance |

#### Risk Owner (Carlos Sanchez)

| Responsibility | Authority | Accountability |
|----------------|-----------|----------------|
| Evaluate risks | Accept/treat risks | Risk treatment |
| Approve treatment plans | Allocate budget | Treatment effectiveness |
| Monitor residual risk | Adjust controls | Risk levels |

#### Engineering Lead (Carlos Sanchez)

| Responsibility | Authority | Accountability |
|----------------|-----------|----------------|
| Implement technical controls | Configure systems | Control operation |
| Manage development security | Approve changes | Secure code |
| Maintain infrastructure | Access to AWS/GitHub | System availability |

### 4.3 Solo Founder Adaptations

Since all roles are held by one person, the following compensating controls apply:

| Challenge | Compensating Control | Evidence |
|-----------|---------------------|----------|
| No segregation of duties | All changes via PR with audit trail | GitHub logs |
| Self-review of work | External audit for independence | Audit reports |
| Single point of failure | Documentation, runbooks | Procedures |
| Self-certification | Formal documentation required | Signed reviews |

### 4.4 Future Organization (Template)

When the team grows, roles will be assigned as follows:

| Role | Minimum Qualifications | Reports To |
|------|----------------------|------------|
| Information Security Manager | 5+ years security experience | CEO |
| Security Engineer | 3+ years, relevant certs | ISM |
| DevOps Engineer | CI/CD, cloud security | Engineering Lead |
| Developer | Secure coding training | Engineering Lead |

---

## 5. Communication (Clause 7.4)

### 5.1 Internal Communication

| What | When | Who | How |
|------|------|-----|-----|
| Security policy changes | Within 1 week of change | All personnel | Email, Git commit |
| Security incidents | Immediately | Affected parties | Direct contact |
| Risk assessment results | After completion | Risk owners | Review meeting |
| Audit findings | Within 2 weeks | Relevant roles | Email, tracking |

### 5.2 External Communication

| What | When | To Whom | How | Authority |
|------|------|---------|-----|-----------|
| Security incidents | Per IR policy | Affected customers | Email | ISM |
| Compliance status | On request | Customers | Portal/email | ISM |
| Regulatory notifications | Per legal requirements | Regulators | Official channels | Top Management |
| Vulnerability disclosures | After mitigation | Public | Security advisory | ISM |

---

## 6. Resources (Clause 7.1)

### 6.1 Current Resources

| Resource | Description | Adequacy |
|----------|-------------|----------|
| Personnel | 1 (founder) | Adequate for current scope |
| Budget | $20-30K/year compliance | Adequate for certification |
| Tools | GitHub, AWS, automation | Adequate |
| Time | ~10 hrs/month ISMS maintenance | Adequate |

### 6.2 Resource Planning

| Need | Timeline | Budget | Priority |
|------|----------|--------|----------|
| External internal auditor | Q1 2026 | $2-5K | High |
| ISO 27001 certification audit | Q2 2026 | $8-15K | High |
| Google Workspace (IdP) | Q1 2026 | $6/user/month | Medium |
| First security hire | TBD | Market rate | Future |

---

## 7. Competence (Clause 7.2)

### 7.1 Current Competence

| Person | Role | Competence Evidence |
|--------|------|---------------------|
| Carlos Sanchez | All roles | 10+ years engineering, cloud security experience, compliance projects |

### 7.2 Competence Requirements

| Role | Required Competence | How to Demonstrate |
|------|--------------------|--------------------|
| Top Management | Business strategy, risk management | Experience, training |
| ISM | ISO 27001, SOC 2, technical security | Certifications, experience |
| Engineering | Secure development, cloud security | Training, code review |

### 7.3 Training Plan

| Training | Audience | Frequency | Status |
|----------|----------|-----------|--------|
| Security awareness | All personnel | Annual | ✅ Complete (founder) |
| Secure coding | Developers | Annual | ✅ Complete |
| Incident response | IR team | Annual | ✅ Complete |
| ISO 27001 awareness | All | Once + refresher | ✅ Complete |

---

## 8. Documented Information (Clause 7.5)

### 8.1 Document Control

| Aspect | Control |
|--------|---------|
| Creation | Markdown in Git repository |
| Approval | Commit with approval comment, signed by owner |
| Version control | Git versioning |
| Distribution | Repository access control |
| Retention | Git history (indefinite) |
| Disposal | N/A for documentation |

### 8.2 Document Hierarchy

```
┌─────────────────────────────────────┐
│           Policies                  │  ← What we do
│  (ISP-001, ACP-001, etc.)          │
└─────────────────┬───────────────────┘
                  │
┌─────────────────▼───────────────────┐
│          Procedures                 │  ← How we do it
│  (Checklists, runbooks)            │
└─────────────────┬───────────────────┘
                  │
┌─────────────────▼───────────────────┐
│           Evidence                  │  ← Proof we did it
│  (Logs, exports, screenshots)      │
└─────────────────────────────────────┘
```

---

## 9. Approval

This document has been reviewed and approved:

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Top Management | Carlos Sanchez | /s/ Carlos Sanchez | 2026-01-17 |

---

## 10. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial leadership and roles document |
