# SOC 2 to ISO 27001 Control Mapping

**Document ID:** ISMS-MAP-001
**Version:** 1.0
**Effective Date:** 2026-01-17
**Owner:** Carlos Sanchez

---

## 1. Purpose

This document maps SOC 2 Trust Services Criteria (Security) to ISO 27001:2022 Annex A controls, enabling evidence reuse and identifying gaps.

---

## 2. Mapping Summary

| SOC 2 Category | Controls | ISO 27001 Coverage | Gap Status |
|----------------|----------|-------------------|------------|
| CC1 - Control Environment | 5 | A.5.1-5.4 | ✅ Covered |
| CC2 - Communication | 3 | A.5.5-5.6, A.6.8 | ✅ Covered |
| CC3 - Risk Assessment | 4 | A.5.7-5.8, Clause 6.1 | ✅ Covered |
| CC4 - Monitoring | 2 | A.5.35-5.36 | ✅ Covered |
| CC5 - Control Activities | 3 | A.5.1, A.5.37 | ✅ Covered |
| CC6 - Logical Access | 8 | A.5.15-5.18, A.8.2-5 | ✅ Covered |
| CC7 - System Operations | 5 | A.5.24-5.28, A.8.8 | ✅ Covered |
| CC8 - Change Management | 1 | A.8.32 | ✅ Covered |
| CC9 - Risk Mitigation | 2 | A.5.19-5.23, A.5.29-30 | ✅ Covered |

**Overall Coverage: ~85% overlap between SOC 2 Security and ISO 27001**

---

## 3. Detailed Control Mapping

### CC1 - Control Environment → ISO 27001 Organizational Controls

| SOC 2 Control | Description | ISO 27001 Control(s) | Evidence Reuse |
|---------------|-------------|---------------------|----------------|
| CC1.1 | Integrity and ethical values | A.5.1, A.5.4 | ✅ Policies |
| CC1.2 | Board/management oversight | A.5.2, Clause 5.1 | ✅ Management review |
| CC1.3 | Organizational structure | A.5.2 | ✅ Org chart, RACI |
| CC1.4 | Commitment to competence | A.6.3, A.7.2 | ✅ Training records |
| CC1.5 | Accountability | A.5.4, A.6.4 | ✅ HR policies |

### CC2 - Communication → ISO 27001 Communication Controls

| SOC 2 Control | Description | ISO 27001 Control(s) | Evidence Reuse |
|---------------|-------------|---------------------|----------------|
| CC2.1 | Internal communication | Clause 7.4, A.6.3 | ✅ Awareness records |
| CC2.2 | External communication | Clause 7.4, A.5.5-5.6 | ✅ External comms |
| CC2.3 | Internal control deficiencies | A.5.36, Clause 10.1 | ✅ NC records |

### CC3 - Risk Assessment → ISO 27001 Risk Management

| SOC 2 Control | Description | ISO 27001 Control(s) | Evidence Reuse |
|---------------|-------------|---------------------|----------------|
| CC3.1 | Security objectives | Clause 6.2, A.5.1 | ✅ Objectives doc |
| CC3.2 | Risk identification/analysis | Clause 6.1.2, A.5.7 | ✅ Risk register |
| CC3.3 | Fraud risk | Clause 6.1.2 | ✅ Risk assessment |
| CC3.4 | Change assessment | A.5.8, A.8.32 | ✅ Change records |

### CC4 - Monitoring → ISO 27001 Performance Evaluation

| SOC 2 Control | Description | ISO 27001 Control(s) | Evidence Reuse |
|---------------|-------------|---------------------|----------------|
| CC4.1 | Monitoring activities | Clause 9.1, A.8.16 | ✅ Monitoring configs |
| CC4.2 | Deficiency communication | Clause 9.2, A.5.36 | ✅ Audit findings |

### CC5 - Control Activities → ISO 27001 Control Implementation

| SOC 2 Control | Description | ISO 27001 Control(s) | Evidence Reuse |
|---------------|-------------|---------------------|----------------|
| CC5.1 | Control selection | Clause 6.1.3, SoA | ✅ SoA |
| CC5.2 | Technology controls | A.8 (all) | ✅ Technical configs |
| CC5.3 | Policy deployment | A.5.1, Clause 7.5 | ✅ Policies |

### CC6 - Logical Access → ISO 27001 Access Control

| SOC 2 Control | Description | ISO 27001 Control(s) | Evidence Reuse |
|---------------|-------------|---------------------|----------------|
| CC6.1 | Access security | A.5.15, A.8.2-3 | ✅ Access configs |
| CC6.2 | Access provisioning | A.5.16, A.5.18 | ✅ Onboarding records |
| CC6.3 | Access removal | A.5.11, A.5.18, A.6.5 | ✅ Offboarding records |
| CC6.4 | Access review | A.5.18 | ✅ Access reviews |
| CC6.5 | Physical access | A.7.1-7.6 | 🔗 Inherited (AWS) |
| CC6.6 | Authentication | A.5.17, A.8.5 | ✅ MFA evidence |
| CC6.7 | RBAC | A.5.18, A.8.2-3 | ✅ RBAC configs |
| CC6.8 | Malware protection | A.8.7 | ✅ Scanning reports |

### CC7 - System Operations → ISO 27001 Operations Security

| SOC 2 Control | Description | ISO 27001 Control(s) | Evidence Reuse |
|---------------|-------------|---------------------|----------------|
| CC7.1 | Vulnerability detection | A.5.7, A.8.8 | ✅ Vuln reports |
| CC7.2 | System monitoring | A.8.15-16 | ✅ Log configs |
| CC7.3 | Event evaluation | A.5.25 | ✅ Alert triage |
| CC7.4 | Incident response | A.5.24-28 | ✅ IR records |
| CC7.5 | Recovery | A.5.27, A.5.29-30 | ✅ Recovery tests |

### CC8 - Change Management → ISO 27001 Change Control

| SOC 2 Control | Description | ISO 27001 Control(s) | Evidence Reuse |
|---------------|-------------|---------------------|----------------|
| CC8.1 | Change management | A.8.32 | ✅ PR records |

### CC9 - Risk Mitigation → ISO 27001 Supplier & Continuity

| SOC 2 Control | Description | ISO 27001 Control(s) | Evidence Reuse |
|---------------|-------------|---------------------|----------------|
| CC9.1 | Vendor risk management | A.5.19-23 | ✅ Vendor assessments |
| CC9.2 | Business continuity | A.5.29-30 | ✅ BC/DR plans |

---

## 4. ISO 27001 Controls NOT Covered by SOC 2

These controls require additional implementation:

| ISO Control | Title | Gap | Action Required |
|-------------|-------|-----|-----------------|
| A.5.5 | Contact with authorities | Partial | Document emergency contacts |
| A.5.6 | Contact with special interest groups | Partial | Document security community engagement |
| A.5.10 | Acceptable use | No policy | Create acceptable use policy |
| A.5.13 | Information labelling | Partial | Implement document classification |
| A.5.32 | Intellectual property | Partial | License compliance process |
| A.5.34 | Privacy/PII | Partial | Privacy impact assessment |
| A.6.1 | Screening | Not implemented | Background check process |
| A.6.2 | Employment terms | Partial | Update employment agreements |
| A.6.4 | Disciplinary process | Not documented | Add to employee handbook |
| A.8.10 | Information deletion | Partial | Data retention/deletion procedures |
| A.8.30 | Outsourced development | Not applicable yet | For future contractors |

---

## 5. Evidence Reuse Matrix

### Directly Reusable Evidence

| Evidence Type | SOC 2 Use | ISO 27001 Use | Location |
|---------------|-----------|---------------|----------|
| Policies | CC controls | Clause 5.2, A.5.1 | `docs/compliance/soc2/policies/` |
| Access reviews | CC6.4 | A.5.18 | `evidence-vault/soc2/YYYY/QX/access-reviews/` |
| Vulnerability reports | CC7.1 | A.8.8 | `evidence-vault/soc2/YYYY/MM/vuln-management/` |
| Training records | CC1.4 | A.6.3 | `evidence-vault/soc2/YYYY/training-policy-ack/` |
| Incident records | CC7.4 | A.5.24-28 | `evidence-vault/soc2/YYYY/MM/incident-response/` |
| Change records (PRs) | CC8.1 | A.8.32 | GitHub + `evidence-vault/soc2/YYYY/MM/ci-cd-security/` |
| Vendor assessments | CC9.1 | A.5.19-23 | `evidence-vault/soc2/YYYY/vendor-management/` |
| Risk register | CC3.2 | Clause 8.2 | `docs/compliance/*/risk-register.md` |
| Audit logs | CC7.2 | A.8.15 | CloudTrail, GitHub audit log |
| MFA evidence | CC6.6 | A.8.5 | AWS IAM, GitHub security exports |

### Evidence Requiring ISO-Specific Format

| Evidence Type | SOC 2 Format | ISO 27001 Requirement | Action |
|---------------|--------------|----------------------|--------|
| Management review | Not required | Clause 9.3 minutes | Create template |
| Internal audit | Not formal | Clause 9.2 report | Conduct audit |
| Statement of Applicability | N/A | Clause 6.1.3d | Created |
| ISMS scope | Informal | Clause 4.3 | Created |
| Risk methodology | Informal | Clause 6.1.2 | Created |

---

## 6. Retroactive Evidence Applicability

Your existing 4+ months of SOC 2 evidence (Sept 2025 - Jan 2026) applies to ISO 27001:

| Evidence | SOC 2 Period | ISO 27001 Applicability |
|----------|--------------|------------------------|
| Git commits (461) | Sept 2025 - Jan 2026 | ✅ A.8.32 Change management |
| Pull requests (58) | Sept 2025 - Jan 2026 | ✅ A.8.4 Access to source code |
| CI/CD runs (200) | Oct 2025 - Jan 2026 | ✅ A.8.25 Secure development |
| Dependabot alerts | Jan 2026 | ✅ A.8.8 Vulnerability management |
| AWS exports | Jan 2026 | ✅ A.8.15 Logging, A.8.5 Authentication |
| GitHub security exports | Jan 2026 | ✅ A.5.15-18 Access controls |

**Conclusion:** Your SOC 2 retroactive evidence provides ISO 27001 ISMS operating evidence.

---

## 7. Certification Timeline Impact

| Scenario | SOC 2 Only | ISO 27001 from Scratch | ISO 27001 with SOC 2 |
|----------|------------|------------------------|---------------------|
| Documentation | Done | 2-3 months | 2-4 weeks |
| Implementation | Done | 3-4 months | 2-4 weeks |
| Operating period | 4+ months | 3-6 months | Use existing |
| Audit prep | Done | 1 month | 1-2 weeks |
| **Total** | **Done** | **9-14 months** | **1-2 months** |

---

## 8. Gap Remediation Plan

| Gap | Priority | Action | Due Date | Status |
|-----|----------|--------|----------|--------|
| Acceptable use policy | High | Create A.5.10 policy | 2026-02-15 | 📋 |
| Contact with authorities | High | Document A.5.5 contacts | 2026-02-01 | 📋 |
| Information labelling | Medium | Implement A.5.13 | 2026-03-01 | 📋 |
| Background check process | Medium | Create A.6.1 procedure | 2026-03-01 | 📋 |
| Internal audit | High | Schedule/conduct Clause 9.2 | 2026-02-28 | 📋 |
| Management review | High | Conduct Clause 9.3 | 2026-02-28 | 📋 |

---

## 9. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-17 | Carlos Sanchez | Initial SOC 2 to ISO 27001 mapping |
