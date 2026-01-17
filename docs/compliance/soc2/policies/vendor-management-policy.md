# Vendor Management Policy

**Policy ID:** VMP-001
**Version:** 1.0
**Effective Date:** 2026-01-17
**Next Review:** 2027-01-17
**Owner:** Carlos Sanchez (Founder)
**Approved By:** Carlos Sanchez (Founder)

## 1. Purpose

This policy establishes requirements for assessing and managing security risks associated with third-party vendors and service providers at Aegis Technologies.

## 2. Scope

This policy applies to:
- All vendors with access to Company systems or data
- Cloud service providers
- Software-as-a-Service (SaaS) providers
- Contractors and consultants
- Business partners with data sharing

## 3. Vendor Classification

### 3.1 Risk Tiers

| Tier | Criteria | Assessment Level |
|------|----------|------------------|
| Critical | Access to customer data, critical systems | Full assessment |
| High | Access to internal data, important systems | Standard assessment |
| Medium | Limited access, non-sensitive data | Basic assessment |
| Low | No data access, replaceable | Minimal assessment |

### 3.2 Classification Factors

- Data sensitivity (customer, personal, confidential)
- System access level
- Business criticality
- Replaceability
- Regulatory implications

## 4. Policy Statements

### 4.1 Vendor Assessment

- All vendors assessed before engagement
- Assessment depth based on risk tier
- Reassessment annually for Critical/High tiers
- Reassessment upon significant changes

### 4.2 Security Requirements

Vendors must demonstrate:
- Information security program
- Access controls and authentication
- Encryption for data in transit/rest
- Incident response capability
- Business continuity planning

### 4.3 Contractual Requirements

Contracts must include:
- Data protection obligations
- Security control requirements
- Audit rights
- Incident notification (< 24 hours)
- Termination and data return provisions
- Insurance requirements (where applicable)

### 4.4 Ongoing Monitoring

- Review vendor security posture annually
- Monitor for security incidents/breaches
- Track compliance with contract terms
- Validate SOC 2/ISO certifications current

## 5. Assessment Requirements by Tier

### 5.1 Critical Tier

- Security questionnaire (comprehensive)
- SOC 2 Type II report review
- Penetration test results
- Business continuity plan review
- On-site assessment (where warranted)
- Executive approval required

### 5.2 High Tier

- Security questionnaire (standard)
- SOC 2 Type II or ISO 27001
- Insurance verification
- Security Lead approval

### 5.3 Medium Tier

- Security questionnaire (basic)
- SOC 2 or equivalent
- Contract review
- Department Lead approval

### 5.4 Low Tier

- Terms of service review
- Basic due diligence
- Manager approval

## 6. Assessment Process

### 6.1 Pre-Engagement

1. Identify business need
2. Classify vendor risk tier
3. Collect vendor documentation
4. Complete security questionnaire
5. Review SOC 2/certifications
6. Identify security gaps
7. Obtain required approvals
8. Execute contract with security terms

### 6.2 Annual Review

1. Request updated certifications
2. Send annual questionnaire
3. Review any incidents
4. Assess continued need
5. Update risk classification
6. Document review results

### 6.3 Termination

1. Disable vendor access
2. Request data deletion confirmation
3. Retrieve Company data
4. Document termination
5. Archive records

## 7. Security Questionnaire Topics

- Organization and governance
- Personnel security
- Physical security
- Access control
- Network security
- Application security
- Data protection
- Incident response
- Business continuity
- Compliance and audit

## 8. Roles and Responsibilities

> **Note:** Aegis is currently a solo-founder company. All roles are held by Carlos Sanchez.

| Role | Current Owner | Responsibilities |
|------|---------------|-----------------|
| Vendor Manager | Carlos Sanchez | Relationship management, contract oversight |
| Security Lead | Carlos Sanchez | Assessment, security requirements |
| Legal | External (TBD) | Contract review, terms |
| Finance | Carlos Sanchez | Payment terms, insurance |
| Business Owner | Carlos Sanchez | Business need justification |

## 9. Vendor Inventory

Maintain inventory including:
- Vendor name and contact
- Services provided
- Data accessed
- Risk tier
- Contract dates
- Assessment status
- Owner

## 10. Exception Process

Exceptions require:
- Business justification
- Risk assessment
- Compensating controls
- Executive approval (Critical tier)
- Time-limited (max 1 year)

## 11. Related Documents

- [Information Security Policy](./information-security-policy.md)
- [Risk Management Policy](./risk-management-policy.md)
- Vendor Security Questionnaire
- Standard Vendor Contract Terms

## 12. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial approved version |

---

## 13. Policy Acknowledgment

| Name | Role | Date | Signature |
|------|------|------|-----------|
| Carlos Sanchez | Founder | 2026-01-17 | /s/ Carlos Sanchez |

---

## 14. Current Vendor Inventory

| Vendor | Services | Risk Tier | Data Access | SOC 2 | Review Date |
|--------|----------|-----------|-------------|-------|-------------|
| GitHub | Source control, CI/CD | Critical | Source code | Yes | 2026-01-17 |
| AWS (GovCloud) | Cloud infrastructure | Critical | All data | Yes | 2026-01-17 |

> Vendor inventory tracked in: `../aegis-compliance-evidence/soc2/[YEAR]/vendor-management/`
