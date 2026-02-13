# Internal Audit Checklist - ISO 27001:2022

**Audit Date:** February 2026
**Auditor:** [External Auditor TBD]
**Auditee:** Carlos Sanchez

---

## Instructions

For each requirement:
- ✅ **Conforming** - Requirement fully met with objective evidence
- ⚠️ **Observation** - Minor issue or opportunity for improvement
- ❌ **Nonconformity** - Requirement not met
- N/A - Not applicable to scope

---

## Clause 4: Context of the Organization

### 4.1 Understanding the organization and its context

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 4.1.1 | External issues determined | Documented external context | `iso27001/00-isms-scope-and-context.md` Section 2.1 | ☐ | |
| 4.1.2 | Internal issues determined | Documented internal context | `iso27001/00-isms-scope-and-context.md` Section 2.2 | ☐ | |

### 4.2 Understanding the needs and expectations of interested parties

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 4.2.1 | Interested parties identified | List of interested parties | `iso27001/00-isms-scope-and-context.md` Section 3 | ☐ | |
| 4.2.2 | Requirements of interested parties | Needs and expectations documented | `iso27001/00-isms-scope-and-context.md` Section 3 | ☐ | |

### 4.3 Determining the scope of the ISMS

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 4.3.1 | Scope boundaries defined | Scope statement | `iso27001/00-isms-scope-and-context.md` Section 4 | ☐ | |
| 4.3.2 | Scope considers issues (4.1) | Traceability | Document review | ☐ | |
| 4.3.3 | Scope considers requirements (4.2) | Traceability | Document review | ☐ | |
| 4.3.4 | Scope documented | Written scope | `iso27001/00-isms-scope-and-context.md` | ☐ | |

### 4.4 Information security management system

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 4.4.1 | ISMS established | ISMS documentation set | `docs/compliance/iso27001/` | ☐ | |
| 4.4.2 | ISMS implemented | Operating evidence | Evidence vault | ☐ | |
| 4.4.3 | ISMS maintained | Update history | Git history | ☐ | |
| 4.4.4 | ISMS continually improved | Improvement records | Corrective actions log | ☐ | |

---

## Clause 5: Leadership

### 5.1 Leadership and commitment

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 5.1.1 | Policy/objectives compatible with strategy | Policy alignment | `iso27001/01-leadership-policy-roles.md` | ☐ | |
| 5.1.2 | ISMS integrated into processes | Process integration | Development workflow, CI/CD | ☐ | |
| 5.1.3 | Resources available | Resource allocation | Budget records | ☐ | |
| 5.1.4 | Importance communicated | Communication records | Training records | ☐ | |
| 5.1.5 | Outcomes achieved | ISMS effectiveness | Metrics, incident records | ☐ | |
| 5.1.6 | Persons supported | Support evidence | Training, tooling | ☐ | |
| 5.1.7 | Continual improvement promoted | Improvement activities | CA log, reviews | ☐ | |
| 5.1.8 | Other roles supported | Role support | N/A (solo founder) | ☐ | |

### 5.2 Policy

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 5.2.1 | Policy appropriate | Policy content | `soc2/policies/information-security-policy.md` | ☐ | |
| 5.2.2 | Objectives included or framework | Objectives | `iso27001/01-leadership-policy-roles.md` Section 3.2 | ☐ | |
| 5.2.3 | Commitment to requirements | Commitment statement | Policy document | ☐ | |
| 5.2.4 | Commitment to improvement | Improvement commitment | Policy document | ☐ | |
| 5.2.5 | Policy documented | Written policy | Policy document | ☐ | |
| 5.2.6 | Policy communicated | Communication evidence | Training records | ☐ | |
| 5.2.7 | Policy available to stakeholders | Availability | Repository access | ☐ | |

### 5.3 Organizational roles, responsibilities and authorities

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 5.3.1 | Roles assigned for conformity | Role documentation | `iso27001/01-leadership-policy-roles.md` Section 4 | ☐ | |
| 5.3.2 | Roles assigned for performance reporting | Reporting structure | Same document | ☐ | |

---

## Clause 6: Planning

### 6.1 Actions to address risks and opportunities

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 6.1.1 | Risks/opportunities determined | Risk assessment | `iso27001/03-risk-register.csv` | ☐ | |
| 6.1.2 | Risk assessment process defined | Methodology | `iso27001/02-risk-methodology.md` | ☐ | |
| 6.1.2.a | Criteria established | Risk criteria | Methodology Section 5-6 | ☐ | |
| 6.1.2.b | Repeatable process | Process documentation | Methodology document | ☐ | |
| 6.1.2.c | Risks identified | Risk identification | Risk register | ☐ | Verify 17 risks |
| 6.1.2.d | Risks analyzed | Likelihood/impact | Risk register columns | ☐ | |
| 6.1.2.e | Risks evaluated | Risk levels assigned | Risk register | ☐ | |
| 6.1.3.a | Risk treatment options selected | Treatment decisions | `iso27001/04-risk-treatment-plan.csv` | ☐ | |
| 6.1.3.b | Controls determined | Control selection | Treatment plan | ☐ | |
| 6.1.3.c | Controls compared to Annex A | SoA created | `iso27001/05-statement-of-applicability.csv` | ☐ | Verify 93 controls |
| 6.1.3.d | Statement of Applicability | SoA complete | SoA document | ☐ | |
| 6.1.3.e | Risk treatment plan | Plan documented | Treatment plan | ☐ | |
| 6.1.3.f | Residual risk accepted | Risk acceptance | Risk register | ☐ | |

### 6.2 Information security objectives

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 6.2.1 | Objectives established | Documented objectives | `iso27001/01-leadership-policy-roles.md` Section 3.2 | ☐ | |
| 6.2.2 | Objectives measurable | Metrics defined | Objectives table | ☐ | |
| 6.2.3 | Objectives communicated | Communication evidence | Training records | ☐ | |
| 6.2.4 | Objectives monitored | Monitoring records | Monthly metrics | ☐ | |

### 6.3 Planning of changes

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 6.3.1 | Changes planned | Change management | `soc2/policies/change-management-policy.md` | ☐ | |

---

## Clause 7: Support

### 7.1 Resources

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 7.1.1 | Resources determined and provided | Resource documentation | `iso27001/01-leadership-policy-roles.md` Section 6 | ☐ | |

### 7.2 Competence

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 7.2.1 | Competence determined | Competence requirements | `iso27001/01-leadership-policy-roles.md` Section 7 | ☐ | |
| 7.2.2 | Competence ensured | Training/experience | Training records | ☐ | |
| 7.2.3 | Actions for competence | Training plan | Training plan | ☐ | |
| 7.2.4 | Competence evidence retained | Training records | Evidence vault | ☐ | |

### 7.3 Awareness

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 7.3.1 | Awareness of policy | Communication evidence | Training records | ☐ | |
| 7.3.2 | Awareness of contribution | Communication | Training content | ☐ | |
| 7.3.3 | Awareness of consequences | Communication | Policy acknowledgment | ☐ | |

### 7.4 Communication

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 7.4.1 | Communication determined | Communication plan | `iso27001/01-leadership-policy-roles.md` Section 5 | ☐ | |

### 7.5 Documented information

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 7.5.1 | Required documentation | ISMS document set | `docs/compliance/iso27001/` | ☐ | |
| 7.5.2 | Document creation/update | Version control | Git history | ☐ | |
| 7.5.3 | Document control | Access control | GitHub permissions | ☐ | |

---

## Clause 8: Operation

### 8.1 Operational planning and control

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 8.1.1 | Processes planned | Process documentation | Procedures | ☐ | |
| 8.1.2 | Processes implemented | Operating evidence | Evidence vault | ☐ | |
| 8.1.3 | Changes controlled | Change records | GitHub PRs | ☐ | Sample 10 PRs |
| 8.1.4 | Outsourced processes controlled | Vendor management | Vendor assessments | ☐ | |

### 8.2 Information security risk assessment

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 8.2.1 | Risk assessments performed | Risk register | `iso27001/03-risk-register.csv` | ☐ | |
| 8.2.2 | Results retained | Risk assessment records | Evidence vault | ☐ | |

### 8.3 Information security risk treatment

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 8.3.1 | Risk treatment plan implemented | Treatment evidence | Treatment plan + evidence | ☐ | |
| 8.3.2 | Results retained | Treatment records | Evidence vault | ☐ | |

---

## Clause 9: Performance Evaluation

### 9.1 Monitoring, measurement, analysis and evaluation

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 9.1.1 | What to monitor determined | Metrics defined | `iso27001/01-leadership-policy-roles.md` Section 3.2 | ☐ | |
| 9.1.2 | Methods determined | Monitoring methods | Operating plan | ☐ | |
| 9.1.3 | Monitoring performed | Monitoring records | Monthly metrics | ☐ | |
| 9.1.4 | Results analyzed | Analysis records | Monthly/quarterly reviews | ☐ | |

### 9.2 Internal audit

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 9.2.1 | Audit programme | Audit programme | `iso27001/07-internal-audit-program.md` | ☐ | |
| 9.2.2.a | Audit planned | Audit plan | `iso27001/11-internal-audit-plan.md` | ☐ | |
| 9.2.2.b | Criteria/scope defined | Audit criteria | Audit plan | ☐ | |
| 9.2.2.c | Auditors objective | External auditor | Auditor independence | ☐ | |
| 9.2.2.d | Results reported | Audit report | **THIS AUDIT** | ☐ | |
| 9.2.2.e | Records retained | Audit records | Evidence vault | ☐ | |

### 9.3 Management review

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 9.3.1 | Management review planned | Review schedule | Operating plan | ☐ | |
| 9.3.2 | Required inputs considered | Input checklist | Review template | ☐ | |
| 9.3.3 | Required outputs documented | Output records | **PENDING FIRST REVIEW** | ☐ | |

---

## Clause 10: Improvement

### 10.1 Continual improvement

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 10.1.1 | ISMS continually improved | Improvement evidence | CA log, updates | ☐ | |

### 10.2 Nonconformity and corrective action

| # | Requirement | Evidence to Review | Evidence Location | Status | Notes |
|---|-------------|-------------------|-------------------|--------|-------|
| 10.2.1 | NC addressed | NC records | `iso27001/09-corrective-actions-log.csv` | ☐ | |
| 10.2.2 | Root cause analysis | RCA records | CA log | ☐ | |
| 10.2.3 | Corrective actions implemented | CA evidence | CA log + evidence | ☐ | |
| 10.2.4 | Records retained | NC/CA records | Evidence vault | ☐ | |

---

## Annex A Controls (Sample)

### A.5 Organizational Controls (Sample: 10 of 37)

| Control | Title | Evidence to Review | Evidence Location | Status | Notes |
|---------|-------|-------------------|-------------------|--------|-------|
| A.5.1 | Policies | Policy set | `soc2/policies/` | ☐ | 7 policies |
| A.5.7 | Threat intelligence | Dependabot config | GitHub settings | ☐ | |
| A.5.15 | Access control | Access policy | `soc2/policies/access-control-policy.md` | ☐ | |
| A.5.17 | Authentication | MFA evidence | AWS/GitHub exports | ☐ | |
| A.5.18 | Access rights | Access reviews | Quarterly reviews | ☐ | |
| A.5.21 | Supply chain | SBOM, Dependabot | CI/CD outputs | ☐ | |
| A.5.24 | Incident planning | IR policy | `soc2/policies/incident-response-policy.md` | ☐ | |
| A.5.29 | Continuity | BC/DR plan | DR documentation | ☐ | |
| A.5.33 | Record protection | Evidence vault | Vault structure | ☐ | |
| A.5.36 | Compliance | Audit programme | This audit | ☐ | |

### A.6 People Controls (Sample: 3 of 8)

| Control | Title | Evidence to Review | Evidence Location | Status | Notes |
|---------|-------|-------------------|-------------------|--------|-------|
| A.6.2 | Employment terms | NDA template | Onboarding docs | ☐ | |
| A.6.3 | Awareness/training | Training records | Evidence vault | ☐ | |
| A.6.5 | Termination | Offboarding checklist | `soc2/procedures/` | ☐ | |

### A.8 Technological Controls (Sample: 10 of 34)

| Control | Title | Evidence to Review | Evidence Location | Status | Notes |
|---------|-------|-------------------|-------------------|--------|-------|
| A.8.2 | Privileged access | Admin access list | AWS/GitHub exports | ☐ | |
| A.8.4 | Source code access | Branch protection | GitHub settings | ☐ | |
| A.8.5 | Secure authentication | MFA status | MFA exports | ☐ | |
| A.8.7 | Malware protection | Container scanning | CI/CD logs | ☐ | |
| A.8.8 | Vulnerability management | Dependabot | Alert exports | ☐ | Sample 20 alerts |
| A.8.15 | Logging | CloudTrail config | AWS exports | ☐ | |
| A.8.24 | Cryptography | TLS config | Architecture docs | ☐ | |
| A.8.25 | Secure development | SDLC docs | Development process | ☐ | |
| A.8.28 | Secure coding | Code review | PR review evidence | ☐ | Sample 10 PRs |
| A.8.32 | Change management | Change process | PR workflow | ☐ | |

---

## Evidence Sampling Results

### Pull Request Sample Review

| # | Repo | PR # | Title | Branch Protection | Review Required | Merged By | Result |
|---|------|------|-------|-------------------|-----------------|-----------|--------|
| 1 | | | | ☐ Yes ☐ No | ☐ Yes ☐ No | | ☐ ✅ ☐ ❌ |
| 2 | | | | ☐ Yes ☐ No | ☐ Yes ☐ No | | ☐ ✅ ☐ ❌ |
| 3 | | | | ☐ Yes ☐ No | ☐ Yes ☐ No | | ☐ ✅ ☐ ❌ |
| ... | | | | | | | |

### Dependabot Alert Sample Review

| # | Repo | Alert | Severity | Detection Date | Resolution Date | Days Open | Result |
|---|------|-------|----------|----------------|-----------------|-----------|--------|
| 1 | | | | | | | ☐ ✅ ☐ ❌ |
| 2 | | | | | | | ☐ ✅ ☐ ❌ |
| ... | | | | | | | |

---

## Findings Summary

### Nonconformities (Major)

| # | Clause/Control | Finding | Evidence |
|---|----------------|---------|----------|
| | | | |

### Nonconformities (Minor)

| # | Clause/Control | Finding | Evidence |
|---|----------------|---------|----------|
| | | | |

### Observations / OFI

| # | Clause/Control | Observation | Recommendation |
|---|----------------|-------------|----------------|
| | | | |

---

## Audit Sign-Off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Lead Auditor | [TBD] | _____________ | [TBD] |
| Auditee | Carlos Sanchez | _____________ | [TBD] |
