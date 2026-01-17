# Internal Audit Program

**Document ID:** ISMS-IAP-001
**Version:** 1.0
**Effective Date:** 2026-01-17
**Next Review:** 2027-01-17
**Owner:** Carlos Sanchez (Information Security Manager)
**ISO 27001 Reference:** Clause 9.2

---

## 1. Purpose

This document establishes the internal audit program for Aegis Technologies' ISMS as required by ISO/IEC 27001:2022 Clause 9.2.

---

## 2. Audit Program Requirements

Per ISO 27001 Clause 9.2, the organization shall:
- Plan, establish, implement, and maintain an audit program
- Define audit criteria and scope for each audit
- Select auditors to ensure objectivity and impartiality
- Ensure audit results are reported to relevant management
- Retain documented information as evidence

---

## 3. Audit Program Overview

### 3.1 Audit Cycle

| Audit Type | Frequency | Scope | Auditor |
|------------|-----------|-------|---------|
| Full ISMS Audit | Annual | All clauses + Annex A controls | External |
| Focused/Targeted Audit | As needed | Specific areas | External or qualified internal |
| SOC 2 Audit | Annual | Trust Services Criteria | External (CPA firm) |

### 3.2 Three-Year Audit Plan

| Year | Audit Focus | Timing | Notes |
|------|-------------|--------|-------|
| **Year 1 (2026)** | Full ISMS (certification) | Q2 2026 | Initial certification |
| Year 1 | SOC 2 Type II | Q2-Q3 2026 | Align with ISO audit |
| **Year 2 (2027)** | Surveillance audit + high-risk areas | Q2 2027 | Focus on changes |
| **Year 3 (2028)** | Surveillance audit + remaining areas | Q2 2028 | Pre-recertification |
| **Year 3 (2028)** | Recertification audit | Q4 2028 | Full scope |

### 3.3 Solo Founder Consideration

> **Important:** As the founder, Carlos Sanchez cannot audit his own work. Internal audits MUST be performed by an external party to ensure independence and impartiality.

**Options for Internal Auditor:**
1. **Consultant** - Hire ISO 27001 qualified consultant (~$2,000-5,000)
2. **Certification body pre-assessment** - Some offer internal audit services
3. **Peer audit** - Another startup founder with ISO 27001 experience
4. **Compliance platform** - Drata, Vanta, Sprinto may offer audit services

---

## 4. 2026 Internal Audit Schedule

### 4.1 Audit Schedule

| Audit | Scope | Date | Auditor | Status |
|-------|-------|------|---------|--------|
| IA-2026-01 | Full ISMS (pre-certification) | Feb 2026 | External (TBD) | 📋 Scheduled |
| IA-2026-02 | Post-certification follow-up | Aug 2026 | External (TBD) | 📋 Planned |

### 4.2 Audit Scope Matrix (Full ISMS Audit)

| Area | Clauses/Controls | Sample Size | Method |
|------|-----------------|-------------|--------|
| Context & Scope | 4.1-4.4 | All | Document review |
| Leadership | 5.1-5.3 | All | Document review, interview |
| Planning | 6.1-6.3 | All | Document review |
| Support | 7.1-7.5 | Sample | Document review, evidence sampling |
| Operation | 8.1-8.3 | Sample | Evidence sampling |
| Performance Evaluation | 9.1-9.3 | All | Document review |
| Improvement | 10.1-10.2 | Sample | NC records review |
| Annex A - Organizational | A.5.1-5.37 | 30% sample | Control testing |
| Annex A - People | A.6.1-6.8 | 50% sample | Control testing |
| Annex A - Physical | A.7.x (applicable) | All applicable | Control testing |
| Annex A - Technological | A.8.1-8.34 | 30% sample | Control testing |

---

## 5. Audit Procedure

### 5.1 Audit Planning

1. **Define scope and objectives**
   - Which clauses and controls
   - Specific focus areas (new implementations, previous findings)

2. **Prepare audit plan**
   - Schedule interviews
   - Identify evidence required
   - Prepare audit checklist

3. **Notify auditee**
   - Share audit scope and schedule
   - Request evidence preparation

### 5.2 Audit Execution

1. **Opening meeting**
   - Confirm scope and schedule
   - Explain audit process

2. **Evidence collection**
   - Document review
   - System inspection
   - Interviews (if applicable)
   - Observation

3. **Finding classification**
   - Nonconformity (Major/Minor)
   - Observation (improvement opportunity)
   - Positive practice

4. **Closing meeting**
   - Present preliminary findings
   - Clarify any issues
   - Agree on timelines

### 5.3 Audit Reporting

| Element | Content |
|---------|---------|
| Audit ID | IA-YYYY-XX |
| Audit date(s) | Date range |
| Auditor(s) | Name, qualifications |
| Scope | Clauses and controls audited |
| Methodology | How audit was conducted |
| Findings | NC, observations, positives |
| Conclusion | Overall ISMS conformity assessment |
| Recommendations | Improvement suggestions |

### 5.4 Finding Definitions

| Type | Definition | Action Required |
|------|------------|-----------------|
| **Major NC** | Absence or complete failure of control; significant risk | Immediate corrective action before certification |
| **Minor NC** | Partial implementation; isolated failure | Corrective action within defined timeframe |
| **Observation** | Improvement opportunity; potential weakness | Consider for improvement |
| **Positive** | Effective control beyond requirements | Document and share |

---

## 6. Corrective Action Process

### 6.1 Corrective Action Requirements

For each nonconformity:
1. **Root cause analysis** - Identify why the NC occurred
2. **Correction** - Fix the immediate issue
3. **Corrective action** - Prevent recurrence
4. **Verification** - Confirm effectiveness

### 6.2 Corrective Action Timeline

| NC Type | Root Cause Analysis | Corrective Action | Verification |
|---------|--------------------|--------------------|--------------|
| Major | 1 week | 30 days | 90 days |
| Minor | 2 weeks | 60 days | 90 days |

### 6.3 Corrective Action Record

| Field | Description |
|-------|-------------|
| NC ID | From audit report |
| NC Description | What was found |
| Root Cause | Why it occurred |
| Correction | Immediate fix |
| Corrective Action | Systemic fix |
| Responsible | Owner |
| Due Date | Completion target |
| Status | Open/In Progress/Closed |
| Verification Date | When effectiveness verified |
| Verified By | Auditor name |

---

## 7. Audit Evidence Requirements

### 7.1 Evidence by Clause

| Clause | Required Evidence |
|--------|-------------------|
| 4.1 | Context analysis documentation |
| 4.2 | Interested parties register |
| 4.3 | ISMS scope document |
| 5.2 | Information security policy (signed) |
| 6.1.2 | Risk assessment methodology, risk register |
| 6.1.3 | Statement of Applicability, risk treatment plan |
| 7.2 | Training records, competence evidence |
| 7.5 | Document control records |
| 8.2 | Risk assessment results |
| 9.2 | This internal audit program and reports |
| 9.3 | Management review minutes |
| 10.1 | Corrective action records |

### 7.2 Evidence Location

All evidence stored in:
- `docs/compliance/iso27001/` - ISMS documentation
- `docs/compliance/soc2/` - Shared policies and procedures
- `../aegis-compliance-evidence/` - Evidence vault
- GitHub, AWS - System evidence

---

## 8. Auditor Qualifications

### 8.1 Required Qualifications

Internal auditors shall have:
- Knowledge of ISO 27001:2022 requirements
- Understanding of audit principles (ISO 19011)
- Independence from area being audited
- Objectivity and impartiality

### 8.2 Preferred Qualifications

- ISO 27001 Lead Auditor certification
- Previous ISMS audit experience
- Understanding of software development industry
- SOC 2 audit experience (for integrated audits)

---

## 9. Records Retention

| Record | Retention Period |
|--------|------------------|
| Audit program | 3 years after cycle completion |
| Audit reports | 3 years |
| Corrective action records | 3 years after closure |
| Evidence samples | 1 year after audit |

---

## 10. Integration with SOC 2 Audits

### 10.1 Combined Audit Benefits

- Shared evidence reduces duplication
- Single audit period for both standards
- Cost savings from combined engagement
- Consistent findings across standards

### 10.2 Integrated Audit Schedule

| Month | Activity |
|-------|----------|
| Jan | ISMS internal audit prep |
| Feb | ISMS internal audit |
| Mar | Corrective actions |
| Apr | ISO 27001 Stage 1 + SOC 2 fieldwork |
| May | Gap remediation |
| Jun | ISO 27001 Stage 2 + SOC 2 completion |
| Jul | Reports issued |

---

## 11. Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Information Security Manager | Carlos Sanchez | /s/ Carlos Sanchez | 2026-01-17 |
| Top Management | Carlos Sanchez | /s/ Carlos Sanchez | 2026-01-17 |

---

## 12. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial internal audit program |
