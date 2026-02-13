# SOC 2 Trust Services Criteria Mapping

## Overview

This document maps SOC 2 Trust Services Criteria (TSC) to Aegis implementations and identifies gaps requiring remediation.

## Security (CC Series) - REQUIRED

### CC1: Control Environment

| Criteria | Description | Aegis Implementation | Status |
|----------|-------------|---------------------|--------|
| CC1.1 | COSO Principle 1: Integrity and ethical values | Code of conduct, ethics policy | 🔴 TODO |
| CC1.2 | COSO Principle 2: Board oversight | Board/management oversight of security | 🔴 TODO |
| CC1.3 | COSO Principle 3: Management structure | Defined security responsibilities | 🔴 TODO |
| CC1.4 | COSO Principle 4: Commitment to competence | Security training program | 🔴 TODO |
| CC1.5 | COSO Principle 5: Accountability | Performance reviews include security | 🔴 TODO |

### CC2: Communication and Information

| Criteria | Description | Aegis Implementation | Status |
|----------|-------------|---------------------|--------|
| CC2.1 | Internal security communication | Security awareness program | 🔴 TODO |
| CC2.2 | Internal control communication | Security policies distributed | 🔴 TODO |
| CC2.3 | External communication | Security page, customer notifications | 🔴 TODO |

### CC3: Risk Assessment

| Criteria | Description | Aegis Implementation | Status |
|----------|-------------|---------------------|--------|
| CC3.1 | Risk objectives defined | Security objectives documented | 🔴 TODO |
| CC3.2 | Risk identification | Risk assessment process | 🔴 TODO |
| CC3.3 | Fraud risk assessment | Fraud controls documented | 🔴 TODO |
| CC3.4 | Change risk assessment | Change impact analysis | 🔴 TODO |

### CC4: Monitoring Activities

| Criteria | Description | Aegis Implementation | Status |
|----------|-------------|---------------------|--------|
| CC4.1 | Ongoing/separate evaluations | Security monitoring, audits | 🟡 Partial |
| CC4.2 | Deficiency communication | Issue tracking, remediation | 🟡 Partial |

### CC5: Control Activities

| Criteria | Description | Aegis Implementation | Status |
|----------|-------------|---------------------|--------|
| CC5.1 | Control activities selection | Controls mapped to risks | 🔴 TODO |
| CC5.2 | Technology controls | Technical controls documented | 🟡 Partial |
| CC5.3 | Policy deployment | Policies enforced via technology | 🟡 Partial |

### CC6: Logical and Physical Access Controls

| Criteria | Description | Aegis Implementation | Status |
|----------|-------------|---------------------|--------|
| CC6.1 | Logical access security | Keycloak OIDC, RBAC | 🟢 Implemented |
| CC6.2 | Access provisioning | User provisioning process | 🟡 Partial |
| CC6.3 | Access removal | Deprovisioning process | 🔴 TODO |
| CC6.4 | Access review | Quarterly access reviews | 🔴 TODO |
| CC6.5 | Physical access | N/A (cloud-hosted) | ✅ N/A |
| CC6.6 | Logical access authentication | MFA, strong passwords | 🟡 Partial |
| CC6.7 | Access restriction | Least privilege, RBAC | 🟢 Implemented |
| CC6.8 | Malware prevention | Container scanning, runtime protection | 🟡 Partial |

### CC7: System Operations

| Criteria | Description | Aegis Implementation | Status |
|----------|-------------|---------------------|--------|
| CC7.1 | Vulnerability detection | Vulnerability scanning | 🟡 Partial |
| CC7.2 | System monitoring | Prometheus, logging | 🟢 Implemented |
| CC7.3 | Change evaluation | Change management process | 🟡 Partial |
| CC7.4 | Incident response | IR procedures | 🔴 TODO |
| CC7.5 | Incident recovery | Recovery procedures | 🔴 TODO |

### CC8: Change Management

| Criteria | Description | Aegis Implementation | Status |
|----------|-------------|---------------------|--------|
| CC8.1 | Change management process | Git-based, PR reviews | 🟢 Implemented |

### CC9: Risk Mitigation

| Criteria | Description | Aegis Implementation | Status |
|----------|-------------|---------------------|--------|
| CC9.1 | Business continuity | BCP/DR plans | 🔴 TODO |
| CC9.2 | Vendor risk management | Vendor assessments | 🔴 TODO |

## Availability (A Series) - FOR MODEL C

| Criteria | Description | Aegis Implementation | Status |
|----------|-------------|---------------------|--------|
| A1.1 | Capacity planning | Resource monitoring, scaling | 🟢 Implemented |
| A1.2 | Environmental protections | Cloud provider controls | ✅ N/A |
| A1.3 | Recovery procedures | Backup, restore procedures | 🔴 TODO |

## Confidentiality (C Series) - IF HANDLING CUSTOMER DATA

| Criteria | Description | Aegis Implementation | Status |
|----------|-------------|---------------------|--------|
| C1.1 | Confidential info identification | Data classification | 🔴 TODO |
| C1.2 | Confidential info disposal | Data retention/deletion | 🔴 TODO |

## Gap Summary

### Critical Gaps (Block SOC 2)

1. **Incident Response Policy** - No formal IR policy or procedures
2. **Access Review Process** - No quarterly access reviews
3. **Risk Assessment** - No formal risk assessment process
4. **Security Training** - No security awareness program

### Medium Priority Gaps

5. Access deprovisioning process
6. Vendor management program
7. Business continuity planning
8. Data classification scheme

### Already Implemented

- Logical access controls (Keycloak, RBAC)
- System monitoring (Prometheus, structured logging)
- Change management (Git, PR reviews)
- TLS encryption
- Container security basics

## Remediation Roadmap

| Gap | Owner | Target Date | Effort |
|-----|-------|-------------|--------|
| Incident Response Policy | TBD | TBD | 1 week |
| Access Review Procedure | TBD | TBD | 1 week |
| Risk Assessment Process | TBD | TBD | 2 weeks |
| Security Training Program | TBD | TBD | 2 weeks |
| Access Deprovisioning | TBD | TBD | 1 week |
| Vendor Management | TBD | TBD | 2 weeks |
| BCP/DR Plans | TBD | TBD | 2 weeks |
