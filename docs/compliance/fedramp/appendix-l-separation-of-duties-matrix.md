# FedRAMP Appendix L -- Separation of Duties Matrix

**System Name:** Aegis Platform
**Version:** 1.0.0-DRAFT
**Last Updated:** 2026-03-09
**Classification:** Customer Confidential
**Control Reference:** AC-5 (Separation of Duties), NIST SP 800-53 Rev 5
**Owner:** Carlos Sanchez (Founder)
**Approved By:** Carlos Sanchez (Founder)

## 1. Purpose

This document defines the separation of duties (SoD) matrix for the Aegis Platform, identifying incompatible role combinations and the controls enforcing separation. As required by FedRAMP and NIST SP 800-53 control AC-5, this matrix ensures that no single individual can control all critical aspects of a process or function without appropriate oversight.

This document also addresses the solo founder context of Aegis Technologies, documenting compensating controls that mitigate risks when full separation of duties is not achievable due to organizational size.

## 2. Platform Roles

### 2.1 Role Definitions

| Role | Description | Keycloak Realm Role | Kubernetes RBAC | AWS IAM |
|------|-------------|--------------------|-----------------|---------|
| **System Administrator** | Manages platform infrastructure, Kubernetes clusters, and operational health. Deploys and upgrades Aegis components. | `admin` | `cluster-admin` (hub cluster) | `AdministratorAccess` (scoped to Aegis resources) |
| **Platform Developer** | Develops and maintains Aegis platform code (platform-api, proxy, k8s-agent). Submits code changes via PR. | `developer` | `edit` (aegis-system namespace) | `ReadOnlyAccess` + ECR push |
| **Security Administrator** | Manages security policies, access controls, PKI configuration, and compliance. Conducts access reviews and security assessments. | `admin` + `security` group | `cluster-admin` (for security operations) | `SecurityAudit` + IAM management |
| **Auditor / Compliance Officer** | Reviews audit logs, compliance evidence, and security controls. Read-only access to all systems for audit purposes. | `viewer` + `auditor` group | `view` (all namespaces) | `SecurityAudit` (read-only) |
| **Tenant Administrator** | Manages users and resources within a specific project/tenant scope. Cannot affect other tenants or platform-level configuration. | `owner` (project-scoped) | `admin` (project namespace only) | N/A (no AWS access) |

### 2.2 Role Hierarchy

```
System Administrator
    |
    +-- Security Administrator (subset: security-specific functions)
    |
    +-- Platform Developer (subset: development functions)
    |
    +-- Tenant Administrator (subset: project-scoped functions)
    |
    +-- Auditor / Compliance Officer (cross-cutting: read-only audit)
```

## 3. Separation of Duties Matrix

### 3.1 Primary Action Matrix

This matrix identifies which roles are authorized to perform each action category. An "X" indicates the role is authorized; a dash indicates the role must NOT perform the action.

| Action Category | System Admin | Platform Developer | Security Admin | Auditor | Tenant Admin |
|----------------|:------------:|:-----------------:|:--------------:|:-------:|:------------:|
| **User Account Creation** | X | -- | X | -- | X (own tenant) |
| **User Account Deletion** | X | -- | X | -- | -- |
| **Role/Permission Assignment** | X | -- | X | -- | X (own tenant) |
| **Privileged Role Assignment** | -- | -- | X | -- | -- |
| **System Configuration Changes** | X | -- | -- | -- | -- |
| **Security Policy Changes** | -- | -- | X | -- | -- |
| **Keycloak Realm Configuration** | -- | -- | X | -- | -- |
| **PKI/Certificate Management** | X | -- | X | -- | -- |
| **Application Code Changes** | -- | X | -- | -- | -- |
| **Code Review/Approval** | X | X (not own code) | X | -- | -- |
| **Production Deployment** | X | -- | -- | -- | -- |
| **Deployment Approval** | X | -- | X | -- | -- |
| **Helm Values Modification** | X | -- | -- | -- | -- |
| **Terraform Infrastructure Changes** | X | -- | -- | -- | -- |
| **Audit Log Access (Read)** | X | -- | X | X | -- |
| **Audit Log Modification** | -- | -- | -- | -- | -- |
| **Audit Log Deletion** | -- | -- | -- | -- | -- |
| **Compliance Evidence Collection** | -- | -- | X | X | -- |
| **Incident Response Execution** | X | X (containment) | X | -- | -- |
| **Incident Investigation** | -- | -- | X | X | -- |
| **Database Direct Access** | X | -- | -- | -- | -- |
| **Database Schema Changes** | X | X (via migration) | -- | -- | -- |
| **Backup Execution** | X | -- | -- | -- | -- |
| **Backup Restoration** | X | -- | X (approval) | -- | -- |
| **Network Policy Changes** | X | -- | X | -- | -- |
| **Workload Submission** | -- | X | -- | -- | X (own tenant) |
| **Workload Termination (any)** | X | -- | X | -- | -- |
| **Workload Termination (own)** | -- | X | -- | -- | X |
| **Cluster Registration** | X | -- | -- | -- | -- |
| **Cluster Deregistration** | X | -- | X (approval) | -- | -- |

### 3.2 Conflict of Interest Matrix

This matrix identifies incompatible role combinations -- roles that should NOT be held by the same individual simultaneously due to conflict of interest.

| Role Combination | Conflict Level | Reason | Compensating Control (if combined) |
|-----------------|:--------------:|--------|-----------------------------------|
| System Admin + Auditor | HIGH | Admin could conceal unauthorized changes from audit | Tamper-evident audit logs stored externally; third-party audit review |
| System Admin + Security Admin | MODERATE | Single point of control for both operations and security policy | Automated policy enforcement via Keycloak; IaC review requirements; scheduled external reviews |
| Platform Developer + Production Deployer | MODERATE | Developer could introduce and deploy malicious code | Mandatory PR review; CI/CD pipeline with automated testing; signed container images |
| Security Admin + Platform Developer | MODERATE | Security admin could weaken controls via code changes | Branch protection; mandatory code review; security scanning in CI/CD |
| System Admin + Platform Developer | LOW-MODERATE | Admin with code access could modify and deploy changes | PR requirements; deployment audit logs; GitOps with change tracking |
| Auditor + any operational role | HIGH | Auditor independence compromised | External auditor engagement; read-only access enforcement |
| Tenant Admin + System Admin | MODERATE | Tenant admin could escalate to platform-level access | Namespace isolation; Keycloak project-scoped roles; API-level tenant enforcement |

## 4. Solo Founder Context and Compensating Controls

### 4.1 Current Organizational Context

Aegis Technologies is currently a solo-founder company. Carlos Sanchez (Founder) necessarily holds all roles defined in Section 2. This creates inherent separation of duties conflicts that cannot be fully resolved through organizational controls alone.

This section documents the compensating controls that mitigate the risks of consolidated role ownership, providing assurance equivalent to separation of duties to the extent possible for the current organizational size.

### 4.2 Compensating Controls Matrix

| SoD Conflict | Risk | Compensating Control | Implementation | Evidence |
|-------------|------|---------------------|----------------|----------|
| **Admin modifies system and reviews own changes** | Unauthorized changes go undetected | **Automated audit logging (tamper-evident)** | All administrative actions logged with immutable timestamps in structured JSON. Logs exported to external SIEM where available. Kubernetes audit policy captures all API server operations. AWS CloudTrail logs all infrastructure changes independently. | CloudTrail logs, Kubernetes audit logs, platform-api structured logs |
| **Developer deploys own code** | Malicious or untested code reaches production | **Infrastructure-as-code review requirements** | All code changes require PR (GitHub branch protection). CI/CD pipeline runs automated tests, security scanning, and linting. No direct commits to main branch. Even self-merged PRs create an auditable record with diff, test results, and timestamp. Container images are SHA-tagged and traceable to specific commits. | GitHub PR history, CI/CD pipeline runs, container image digests |
| **Security admin sets and audits own policies** | Security gaps concealed from oversight | **Scheduled third-party reviews** | External security assessments conducted annually. SOC 2 Type II audit by independent auditor. Penetration testing by external firm. Compliance evidence collected automatically and stored in separate repository (`aegis-compliance-evidence`). | Third-party audit reports, penetration test results, compliance evidence vault |
| **Admin accesses and could modify audit logs** | Audit trail tampering | **Role-based access enforced by Keycloak even for admin** | Keycloak RBAC is enforced at the Platform API level regardless of caller identity. Admin actions are logged before execution (write-ahead audit). Audit log modification/deletion endpoints do not exist in the Platform API. Database-level audit logs have separate retention from application logs. | Keycloak RBAC configuration, Platform API source code (no log modification endpoints), database audit table schema |
| **Single person handles incident response** | Incomplete investigation, conflicts of interest | **Documented IR procedures with external escalation** | Incident Response Policy (IRP-001) defines escalation to external counsel and forensics providers. Automated alerting ensures incidents are detected even without active monitoring. Post-incident reviews documented and stored in compliance evidence. | IRP-001, external resource contact list, incident report templates |
| **Single person manages backups and restoration** | Backup integrity not independently verified | **Automated backup with cross-account verification** | RDS automated backups with AWS-managed encryption. Backup restoration tested quarterly with documented results. Terraform state stored in S3 with versioning and cross-region replication. | RDS backup configuration, restoration test records, S3 versioning policy |

### 4.3 Technical Controls Enforcing Separation

Even with a single operator, the following technical controls enforce separation at the system level:

| Control | Mechanism | What It Prevents |
|---------|-----------|-----------------|
| **Keycloak RBAC enforcement** | Platform API validates JWT claims and role memberships for every request. Role checks are performed in application code, not bypassable by infrastructure access alone. | Admin bypassing access controls at the application layer. |
| **GitHub branch protection** | `main` branch requires PR. Force push disabled. Branch deletion disabled. These rules are enforced by GitHub, not by the repository owner. | Direct, unreviewed code deployment to production. |
| **Kubernetes admission control** | Pod Security Standards enforced via admission controller. Namespace-level resource quotas and network policies applied. | Container escape, privilege escalation, cross-namespace access. |
| **AWS CloudTrail** | Independent, AWS-managed logging of all API calls. Logs stored in S3 with object lock. Trail cannot be disabled without generating its own audit event. | Infrastructure changes occurring without audit trail. |
| **Immutable container images** | Container images built in CI/CD, pushed to ECR with SHA digest. Running containers use `imagePullPolicy: Always` in cloud. No runtime image modification possible. | Post-build tampering with deployed software. |
| **Database migration framework** | Schema changes only applied through versioned, numbered migration files. Migrations are idempotent and reversible (up/down). | Untracked database schema modifications. |
| **Git-based configuration management** | All Helm values, Terraform configurations, and scripts stored in Git. Every change has author, timestamp, and diff. | Configuration drift without audit trail. |

### 4.4 Organizational Growth Plan

As Aegis Technologies grows, the following role separations will be implemented in priority order:

| Team Size | Role Separation | SoD Conflicts Resolved |
|-----------|----------------|----------------------|
| **2 people** | Separate Platform Developer from System Administrator. Developer cannot deploy; Admin cannot commit code without review. | Developer + Deployer conflict |
| **3 people** | Add dedicated Security Administrator. Security policy changes require separate approval from operations. | Admin + Security Admin conflict |
| **5+ people** | Add dedicated Auditor/Compliance Officer with no operational access. All audit and compliance functions independent. | Admin + Auditor conflict; full SoD achieved |
| **10+ people** | Implement formal change advisory board (CAB). Rotate security review responsibilities. | All primary conflicts resolved with organizational controls |

### 4.5 Auditor Guidance

When evaluating the solo founder SoD posture, auditors should consider:

1. **Technical controls provide automated enforcement** that does not depend on human separation. Keycloak RBAC, GitHub branch protection, and AWS CloudTrail operate independently of the operator.

2. **Compensating controls are documented and tested.** Each control in Section 4.2 has associated evidence that can be independently verified.

3. **The risk profile is proportional to the current scale.** A single-operator system with a small number of users and workloads has a smaller attack surface than a large enterprise deployment.

4. **Growth milestones trigger SoD improvements.** The organizational growth plan (Section 4.4) commits to implementing traditional SoD as team size permits.

5. **External review provides independent oversight.** Annual third-party security assessments, SOC 2 audits, and penetration testing provide independent verification that compensating controls are effective.

## 5. Enforcement Mechanisms

### 5.1 Keycloak Role Enforcement

| Keycloak Realm Role | Permitted Platform API Operations | Enforcement Point |
|---------------------|----------------------------------|-------------------|
| `admin` | All operations | Platform API JWT validation |
| `owner` | Project-scoped: user management, workload management, cluster viewing | Platform API project-level RBAC |
| `developer` | Workload submission, workspace access, status viewing | Platform API role check |
| `viewer` | Read-only access to authorized projects | Platform API role check |
| `auditor` (group) | Read-only access to audit logs and compliance endpoints | Platform API group membership check |
| `security` (group) | Security configuration, access reviews, policy management | Platform API group membership check |

### 5.2 Kubernetes RBAC Enforcement

| ClusterRole | Bound To | Permissions | Namespace Scope |
|-------------|----------|-------------|-----------------|
| `cluster-admin` | System Administrator SA | All resources, all verbs | Cluster-wide |
| `aegis-agent` | k8s-agent service account | AegisWorkload CRD, Pods, Services | aegis-system + workload namespaces |
| `aegis-viewer` | Auditor service account | Get, List, Watch on all resources | All namespaces (read-only) |
| `namespace-admin` | Tenant Administrator SA | All resources within namespace | Project namespace only |

### 5.3 AWS IAM Enforcement

| IAM Policy | Attached To | Permissions | Boundary |
|------------|-------------|-------------|----------|
| `AegisAdminPolicy` | System Administrator | EKS, RDS, ECR, VPC, Route53 management | Aegis-tagged resources only |
| `AegisDeveloperPolicy` | Platform Developer | ECR push, EKS read-only, CloudWatch read | Aegis ECR repositories only |
| `AegisAuditorPolicy` | Auditor | CloudTrail read, CloudWatch read, S3 read (logs) | Read-only, all Aegis accounts |
| `AegisSecurityPolicy` | Security Administrator | IAM read, SecurityHub, GuardDuty, Inspector | Security services only |

## 6. Action-Level Approval Requirements

### 6.1 Actions Requiring Dual Authorization

The following actions require approval from a role different from the requestor. In the solo founder context, compensating controls (automated checks, external review) substitute for human dual authorization.

| Action | Requestor Role | Approver Role | Solo Founder Compensating Control |
|--------|---------------|---------------|----------------------------------|
| Privileged role assignment | Any | Security Administrator | Keycloak role assignment audit log + quarterly access review |
| Production deployment | Platform Developer | System Administrator | CI/CD pipeline gates (tests, scanning); PR merge requirement |
| Security policy change | Security Administrator | System Administrator | Git-tracked policy with PR; annual third-party review |
| Emergency access grant | System Administrator | Security Administrator | Time-limited access (max 24 hours); post-hoc documentation within 24 hours |
| Backup restoration | System Administrator | Security Administrator | Documented restoration procedure; pre/post integrity verification |
| Cluster deregistration | System Administrator | Security Administrator | Platform API audit log; confirmation workflow in UI |
| Database direct access | System Administrator | Security Administrator | Database audit logging; connection via bastion with session recording |
| Audit log export/deletion | Not permitted | Not permitted | Technical control: no delete endpoint exists; export is append-only |

### 6.2 Actions Requiring Documentation Only

| Action | Performing Role | Documentation Requirement |
|--------|----------------|--------------------------|
| Standard user account creation | Tenant Administrator | Account request ticket with business justification |
| Workload submission | Developer / Tenant Admin | Workload manifest tracked in Platform API audit log |
| Configuration change (non-security) | System Administrator | PR with description, test results, rollback plan |
| Security scanning execution | Security Administrator | Scan report stored in compliance evidence repository |

## 7. Monitoring and Compliance

### 7.1 SoD Violation Detection

| Detection Method | Frequency | What It Detects | Response |
|-----------------|-----------|-----------------|----------|
| Keycloak role assignment audit | Real-time (audit log) | Unauthorized role combinations | Alert + investigation |
| Quarterly access review | Quarterly | Role drift, unnecessary privileges | Access adjustment + documentation |
| CI/CD pipeline enforcement | Per commit | Code changes without PR, failed tests | Deployment blocked |
| AWS CloudTrail monitoring | Real-time | Infrastructure changes outside approved process | Alert + investigation |
| Git commit analysis | Per PR | Self-approved changes (when team > 1) | Change reverted + review |

### 7.2 Reporting

| Report | Frequency | Audience | Content |
|--------|-----------|----------|---------|
| Access review completion | Quarterly | Management, Auditor | Role assignments, changes, exceptions |
| SoD exception report | Quarterly | Security Administrator, Auditor | Active SoD exceptions with compensating controls |
| Compensating control effectiveness | Annually | Management, External auditor | Evidence that compensating controls are operating effectively |
| Organizational growth SoD assessment | Semi-annually | Management | Current team size vs. SoD milestone targets |

## 8. Exception Process

### 8.1 SoD Exceptions

When separation of duties cannot be maintained (beyond the solo founder context), exceptions require:

1. **Business justification** documenting why separation is not feasible
2. **Risk assessment** identifying the specific risks of the combined roles
3. **Compensating controls** defining additional measures to mitigate risk
4. **Management approval** with documented sign-off
5. **Time limitation** -- exceptions are valid for a maximum of 1 year and must be re-evaluated
6. **Enhanced monitoring** -- exception holders subject to increased audit scrutiny

### 8.2 Current Exceptions

| Exception | Holder | Justification | Compensating Controls | Expiration | Review Date |
|-----------|--------|---------------|-----------------------|------------|-------------|
| All roles combined | Carlos Sanchez (Founder) | Solo founder; team size does not support separation | See Section 4.2 (full compensating control matrix) | Until second hire or 2027-03-09, whichever comes first | 2026-09-09 (semi-annual) |

## 9. References

| Document | Relevance |
|----------|-----------|
| NIST SP 800-53 Rev 5, AC-5 | Separation of Duties control definition |
| FedRAMP Moderate Baseline | SoD requirements for cloud service providers |
| ISP-001 Information Security Policy | Organizational security governance |
| ACP-001 Access Control Policy | Access control requirements and role definitions |
| CMP-001 Change Management Policy | Change process and segregation requirements |
| IRP-001 Incident Response Policy | Incident response roles and escalation |
| VMP-001 Vendor Management Policy | Third-party access and review requirements |

## 10. Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0-DRAFT | 2026-03-09 | Carlos Sanchez | Initial draft |
