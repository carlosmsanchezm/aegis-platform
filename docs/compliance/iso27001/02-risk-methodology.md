# Risk Assessment Methodology

**Document ID:** ISMS-RAM-001
**Version:** 1.0
**Effective Date:** 2026-01-17
**Next Review:** 2027-01-17
**Owner:** Carlos Sanchez (Information Security Manager)
**ISO 27001 Reference:** Clause 6.1.2

---

## 1. Purpose

This document defines the methodology for identifying, analyzing, and evaluating information security risks at Aegis Technologies, as required by ISO/IEC 27001:2022 Clause 6.1.2.

---

## 2. Scope

This methodology applies to all information assets, processes, and systems within the ISMS scope as defined in ISMS-SCOPE-001.

---

## 3. Risk Assessment Process Overview

```
┌─────────────────┐
│ 1. Context      │  Define scope, criteria, stakeholders
└────────┬────────┘
         ▼
┌─────────────────┐
│ 2. Risk         │  Identify threats, vulnerabilities, assets
│    Identification│
└────────┬────────┘
         ▼
┌─────────────────┐
│ 3. Risk         │  Assess likelihood and impact
│    Analysis     │
└────────┬────────┘
         ▼
┌─────────────────┐
│ 4. Risk         │  Compare against acceptance criteria
│    Evaluation   │
└────────┬────────┘
         ▼
┌─────────────────┐
│ 5. Risk         │  Select controls, document plan
│    Treatment    │
└────────┬────────┘
         ▼
┌─────────────────┐
│ 6. Monitoring   │  Review, update, report
│    & Review     │
└─────────────────┘
```

---

## 4. Risk Identification

### 4.1 Asset Identification

Information assets shall be identified and categorized:

| Asset Category | Examples | Owner |
|----------------|----------|-------|
| **Information** | Source code, configs, customer data, documentation | Carlos Sanchez |
| **Software** | Aegis Platform, dependencies, tools | Carlos Sanchez |
| **Hardware** | Laptops, cloud infrastructure (virtual) | Carlos Sanchez |
| **Services** | GitHub, AWS, Google (planned) | Carlos Sanchez |
| **People** | Founder, future employees, contractors | Carlos Sanchez |
| **Intangibles** | Reputation, intellectual property | Carlos Sanchez |

### 4.2 Threat Identification

Common threat sources for Aegis:

| Threat Source | Examples |
|---------------|----------|
| **External Attackers** | Nation-state actors, criminals, hacktivists |
| **Insider Threats** | Malicious insiders, negligent employees |
| **Supply Chain** | Compromised dependencies, vendor breaches |
| **Environmental** | Cloud provider outages, natural disasters |
| **Technical** | Software vulnerabilities, misconfigurations |

### 4.3 Vulnerability Identification

Sources for identifying vulnerabilities:
- Dependabot alerts (automated)
- Container image scanning
- Code review findings
- Penetration testing results
- Security advisories (GitHub, CVE)
- Configuration audits

---

## 5. Risk Analysis

### 5.1 Likelihood Scale

| Level | Rating | Description | Frequency |
|-------|--------|-------------|-----------|
| 5 | Almost Certain | Expected to occur | >90% chance in 12 months |
| 4 | Likely | Will probably occur | 50-90% chance in 12 months |
| 3 | Possible | May occur | 10-50% chance in 12 months |
| 2 | Unlikely | Not expected | 1-10% chance in 12 months |
| 1 | Rare | Exceptional circumstances | <1% chance in 12 months |

### 5.2 Impact Scale

| Level | Rating | Description | Examples |
|-------|--------|-------------|----------|
| 5 | Critical | Catastrophic impact | Complete data breach, business failure, legal action |
| 4 | High | Major impact | Significant data loss, major customer impact, >$100K |
| 3 | Medium | Moderate impact | Limited data exposure, some customers affected, $10-100K |
| 2 | Low | Minor impact | Minimal data impact, few users affected, <$10K |
| 1 | Negligible | Insignificant | No data loss, no customer impact, minimal cost |

### 5.3 Impact Categories

Impact should be assessed across multiple dimensions:

| Category | Description | Example (Level 5) |
|----------|-------------|-------------------|
| **Confidentiality** | Unauthorized disclosure | Source code leaked publicly |
| **Integrity** | Unauthorized modification | Malicious code in release |
| **Availability** | Loss of access | Development systems down >24h |
| **Financial** | Monetary loss | Legal liability >$1M |
| **Reputational** | Damage to reputation | Public security incident |
| **Compliance** | Regulatory violation | Loss of certification |

**Highest impact across categories is used for risk calculation.**

### 5.4 Risk Rating Matrix

| | Impact 1 | Impact 2 | Impact 3 | Impact 4 | Impact 5 |
|--|----------|----------|----------|----------|----------|
| **Likelihood 5** | Medium (5) | Medium (10) | High (15) | Critical (20) | Critical (25) |
| **Likelihood 4** | Low (4) | Medium (8) | High (12) | High (16) | Critical (20) |
| **Likelihood 3** | Low (3) | Medium (6) | Medium (9) | High (12) | High (15) |
| **Likelihood 2** | Low (2) | Low (4) | Medium (6) | Medium (8) | High (10) |
| **Likelihood 1** | Low (1) | Low (2) | Low (3) | Medium (4) | Medium (5) |

**Risk Score = Likelihood × Impact**

---

## 6. Risk Evaluation

### 6.1 Risk Acceptance Criteria

| Risk Level | Score | Treatment Requirement | Approval Authority |
|------------|-------|----------------------|-------------------|
| **Critical** | 16-25 | Immediate treatment required | CEO/Founder |
| **High** | 10-15 | Treatment within 30 days | Information Security Manager |
| **Medium** | 5-9 | Treatment within 90 days | Risk Owner |
| **Low** | 1-4 | Accept or treat opportunistically | Risk Owner |

### 6.2 Risk Appetite Statement

> Aegis Technologies has a **low risk appetite** for information security risks that could:
> - Compromise customer trust
> - Expose source code or intellectual property
> - Result in compliance failures
> - Cause reputational damage
>
> The organization accepts that some residual risk is unavoidable but will prioritize treating risks that could impact customers or compliance.

---

## 7. Risk Treatment

### 7.1 Treatment Options

| Option | Description | When to Use |
|--------|-------------|-------------|
| **Mitigate** | Implement controls to reduce risk | Most common - reduce likelihood or impact |
| **Transfer** | Share risk with third party | Insurance, outsourcing to certified vendors |
| **Accept** | Acknowledge and monitor | Low risks, cost exceeds benefit |
| **Avoid** | Eliminate the risk source | High risks that can't be mitigated |

### 7.2 Control Selection

Controls shall be selected from:
1. ISO 27001:2022 Annex A (primary source)
2. SOC 2 Trust Services Criteria (already implemented)
3. Industry best practices (NIST, CIS)
4. Regulatory requirements (FIPS, FedRAMP)

### 7.3 Risk Treatment Plan

For each risk requiring treatment, document:
- Risk ID and description
- Current risk level
- Selected treatment option
- Proposed controls
- Expected residual risk
- Implementation timeline
- Responsible owner
- Resources required

---

## 8. Risk Register

### 8.1 Risk Register Fields

| Field | Description |
|-------|-------------|
| Risk ID | Unique identifier (RISK-001) |
| Risk Description | Clear statement of the risk |
| Asset(s) Affected | Information assets at risk |
| Threat | Threat source/event |
| Vulnerability | Weakness exploited |
| Existing Controls | Current mitigations |
| Likelihood (Inherent) | Without additional controls |
| Impact | Potential consequence |
| Risk Score (Inherent) | L × I |
| Treatment Option | Mitigate/Transfer/Accept/Avoid |
| Treatment Plan | Actions to implement |
| Residual Likelihood | After treatment |
| Residual Impact | After treatment |
| Residual Risk Score | Target risk level |
| Owner | Accountable person |
| Due Date | Treatment completion date |
| Status | Open/In Progress/Closed |

### 8.2 Risk Register Location

Risk Register maintained at: `docs/compliance/iso27001/04-risk-register.md`

---

## 9. Monitoring and Review

### 9.1 Review Frequency

| Activity | Frequency | Responsible |
|----------|-----------|-------------|
| Risk register review | Monthly | Information Security Manager |
| Full risk assessment | Annually | Information Security Manager |
| Post-incident risk review | After P1/P2 incidents | Information Security Manager |
| Change-triggered review | After significant changes | Risk Owner |

### 9.2 Triggers for Risk Re-assessment

- Significant changes to ISMS scope
- New products or services
- Major security incidents
- Changes in threat landscape
- Regulatory or compliance changes
- Technology changes
- Audit findings

### 9.3 Key Risk Indicators (KRIs)

| KRI | Threshold | Frequency |
|-----|-----------|-----------|
| Open critical/high vulnerabilities | <5 older than 30 days | Weekly |
| Open high/critical risks | <3 | Monthly |
| Overdue risk treatments | 0 | Monthly |
| Security incidents | <2 per quarter | Quarterly |
| Failed access attempts | Baseline + 50% | Daily |

---

## 10. Documentation and Records

### 10.1 Required Records

- Risk assessment reports
- Risk register (current and historical)
- Risk treatment plans
- Management review inputs/outputs
- Risk acceptance records

### 10.2 Retention

Risk assessment records shall be retained for minimum 3 years (or as required by compliance obligations).

---

## 11. Roles and Responsibilities

| Role | Responsibilities |
|------|-----------------|
| **Top Management** | Approve risk acceptance criteria, approve high/critical risk acceptance |
| **Information Security Manager** | Maintain methodology, coordinate assessments, maintain register |
| **Risk Owners** | Identify risks, implement treatments, report status |
| **All Personnel** | Report potential risks and incidents |

> **Note:** As a solo-founder company, Carlos Sanchez currently holds all roles.

---

## 12. Related Documents

- ISMS Scope (ISMS-SCOPE-001)
- Statement of Applicability (ISMS-SOA-001)
- Risk Register (ISMS-RR-001)
- Risk Treatment Plan (ISMS-RTP-001)
- SOC 2 Risk Management Policy (RMP-001)

---

## 13. Alignment with SOC 2

This methodology aligns with SOC 2 Trust Services Criteria:
- CC3.1: Security objectives specified
- CC3.2: Risks identified and analyzed
- CC3.3: Fraud risk considered
- CC3.4: Changes assessed for risk

**SOC 2 risk assessment evidence can be reused for ISO 27001.**

---

## 14. Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Information Security Manager | Carlos Sanchez | /s/ Carlos Sanchez | 2026-01-17 |
| Top Management | Carlos Sanchez | /s/ Carlos Sanchez | 2026-01-17 |

---

## 15. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial risk assessment methodology |
