# SOC 2 Policy Register

**Last Updated:** 2026-01-17
**Maintained By:** Carlos Sanchez (Founder)

---

## Policy Summary

| Policy ID | Policy Name | Version | Effective | Next Review | Owner | Status |
|-----------|-------------|---------|-----------|-------------|-------|--------|
| ISP-001 | Information Security Policy | 1.0 | 2026-01-17 | 2027-01-17 | Carlos Sanchez | ✅ Approved |
| ACP-001 | Access Control Policy | 1.0 | 2026-01-17 | 2027-01-17 | Carlos Sanchez | ✅ Approved |
| CMP-001 | Change Management Policy | 1.0 | 2026-01-17 | 2027-01-17 | Carlos Sanchez | ✅ Approved |
| IRP-001 | Incident Response Policy | 1.0 | 2026-01-17 | 2027-01-17 | Carlos Sanchez | ✅ Approved |
| RMP-001 | Risk Management Policy | 1.0 | 2026-01-17 | 2027-01-17 | Carlos Sanchez | ✅ Approved |
| VMP-001 | Vendor Management Policy | 1.0 | 2026-01-17 | 2027-01-17 | Carlos Sanchez | ✅ Approved |

---

## Personnel Acknowledgments

### Current Personnel

| Name | Role | Start Date | All Policies Acknowledged | Date |
|------|------|------------|---------------------------|------|
| Carlos Sanchez | Founder | 2025 | ✅ Yes | 2026-01-17 |

### Acknowledgment Evidence

Policy acknowledgments stored in: `../aegis-compliance-evidence/soc2/[YEAR]/training-policy-ack/`

File naming: `YYYY-MM-DD_[name]_policy-acknowledgment.pdf`

---

## Annual Review Calendar

| Month | Activity |
|-------|----------|
| January | Annual policy review cycle begins |
| January | Q4 access review due (previous year) |
| April | Q1 access review due |
| July | Q2 access review due |
| October | Q3 access review due |
| December | Prepare for annual policy review |

### 2027 Policy Review Schedule

| Policy | Review Due | Reviewer | Status |
|--------|------------|----------|--------|
| ISP-001 | 2027-01-17 | Carlos Sanchez | Pending |
| ACP-001 | 2027-01-17 | Carlos Sanchez | Pending |
| CMP-001 | 2027-01-17 | Carlos Sanchez | Pending |
| IRP-001 | 2027-01-17 | Carlos Sanchez | Pending |
| RMP-001 | 2027-01-17 | Carlos Sanchez | Pending |
| VMP-001 | 2027-01-17 | Carlos Sanchez | Pending |

---

## SOC 2 Control Coverage

| Policy | Primary Controls Covered |
|--------|-------------------------|
| ISP-001 | CC1.1, CC1.2, CC1.3, CC1.4, CC2.1, CC2.2 |
| ACP-001 | CC6.1, CC6.2, CC6.3, CC6.4, CC6.5, CC6.6, CC6.7 |
| CMP-001 | CC8.1 |
| IRP-001 | CC7.3, CC7.4, CC7.5 |
| RMP-001 | CC3.1, CC3.2, CC3.3, CC3.4, CC4.1, CC4.2 |
| VMP-001 | CC9.1, CC9.2 |

---

## Procedures (Supporting Documents)

| Procedure | Related Policy | Location |
|-----------|---------------|----------|
| Employee Onboarding Checklist | ACP-001 | `procedures/employee-onboarding-checklist.md` |
| Employee Offboarding Checklist | ACP-001 | `procedures/employee-offboarding-checklist.md` |
| Quarterly Access Review | ACP-001 | `procedures/quarterly-access-review.md` |

---

## Solo Founder Notes

As a one-person company, the following adjustments apply:

1. **Segregation of Duties**: Not fully achievable. Compensating controls:
   - All changes via PR (audit trail)
   - Branch protection enforced
   - CloudTrail logging enabled

2. **Access Reviews**: Self-review quarterly, documented in evidence vault

3. **Incident Response**: Single point of contact, external resources on retainer as company grows

4. **Policy Reviews**: Annual self-review with documented evidence

These limitations are acceptable for SOC 2 Type II as long as they are:
- Documented (done)
- Compensating controls in place (done)
- Communicated to auditor upfront (do this)

---

## Auditor Notes

When engaging a SOC 2 auditor, provide:

1. This policy register
2. Evidence vault access (read-only)
3. Solo founder context upfront
4. Compensating controls documentation

Expected management representation letter points:
- Company is sole proprietorship / single founder
- Segregation of duties limitations acknowledged
- Compensating controls documented

---

## Revision History

| Date | Author | Changes |
|------|--------|---------|
| 2026-01-17 | Carlos Sanchez | Initial policy register created, all policies v1.0 approved |
