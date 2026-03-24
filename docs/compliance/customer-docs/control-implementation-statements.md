# Aegis Control Implementation Statements

**Version:** 2.0.0-DRAFT
**Last Updated:** 2026-03-09
**Classification:** Customer Confidential

## Overview

This document provides Control Implementation Statements (CIS) for NIST 800-53 Rev 5 controls as implemented by Aegis Platform. Customers can use these statements directly in their System Security Plan (SSP) Section 13.

## How to Use This Document

For each control:
1. **Aegis Provides** - What Aegis implements automatically
2. **Customer Configures** - What the customer must set up
3. **Customer Implements** - What falls outside Aegis scope
4. **Evidence** - How to demonstrate implementation

## Control Family: Access Control (AC)

### AC-2: Account Management

**Control:** The organization manages information system accounts.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Keycloak IdP with user lifecycle management<br>- OIDC/SAML integration<br>- Role-based access control<br>- Audit logging of account events |
| **Customer Configures** | - Keycloak realm settings<br>- User provisioning workflows<br>- Account approval processes<br>- Inactivity timeout (default: 30 days) |
| **Customer Implements** | - Account request process<br>- Periodic account reviews<br>- Termination procedures |
| **Evidence** | - Keycloak user list export<br>- Account creation audit logs<br>- Access review records |

### AC-3: Access Enforcement

**Control:** The system enforces approved authorizations for logical access.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - RBAC enforcement at Platform API<br>- Namespace isolation in Kubernetes<br>- Project-level access boundaries<br>- Audit logging of access decisions |
| **Customer Configures** | - Role assignments in Keycloak<br>- Project membership<br>- Custom RBAC policies |
| **Customer Implements** | - Access request/approval workflow<br>- Separation of duties matrix |
| **Evidence** | - RBAC policy configurations<br>- Access denied audit logs<br>- Role assignment exports |

### AC-6: Least Privilege

**Control:** The organization employs the principle of least privilege.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Predefined roles (Admin, Owner, Developer, Viewer)<br>- Kubernetes RBAC for agents<br>- Service accounts with minimal permissions |
| **Customer Configures** | - Role assignments per user<br>- Custom role definitions (optional) |
| **Customer Implements** | - Privileged access management process<br>- Regular privilege audits |
| **Evidence** | - Role definition exports<br>- User-to-role mapping<br>- Kubernetes RBAC manifests |

### AC-11: Device Lock

**Control:** The system prevents further access after inactivity period.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Configurable session idle timeout<br>- SUSPENDED state (vs FAILED) for long operations<br>- Automatic session termination |
| **Customer Configures** | - `SESSION_IDLE_TIMEOUT` (default: 15 min)<br>- `AEGIS_PROXY_TOKEN_TTL_SECONDS` (max: 300s) |
| **Customer Implements** | - Workstation lock policies |
| **Evidence** | - Session timeout configuration<br>- Session termination audit logs |

### AC-17: Remote Access

**Control:** The organization authorizes, monitors, and controls remote access.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - TLS-encrypted remote access to workloads<br>- Short-lived connection tokens<br>- One-time use tokens (optional)<br>- Audit logging of all remote sessions |
| **Customer Configures** | - VPN requirements (outside Aegis)<br>- Token TTL settings<br>- One-time token enforcement |
| **Customer Implements** | - Remote access policy<br>- Authorized user lists |
| **Evidence** | - Connection session audit logs<br>- Token configuration exports |

## Control Family: Audit and Accountability (AU)

### AU-2: Event Logging

**Control:** The system generates audit records for defined events.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Authentication events<br>- Authorization decisions<br>- Workload lifecycle events<br>- Administrative actions<br>- Security-relevant configuration changes |
| **Customer Configures** | - Log export destinations<br>- Additional log categories |
| **Customer Implements** | - Log review procedures<br>- Alert definitions |
| **Evidence** | - Sample audit logs<br>- Event type documentation |

### AU-3: Content of Audit Records

**Control:** Audit records contain required information.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | All audit logs include:<br>- Timestamp (UTC, ISO 8601)<br>- Event type/ID<br>- Subject identity (user, service)<br>- Source IP address<br>- Action performed<br>- Resource affected<br>- Outcome (success/failure) |
| **Customer Configures** | - Additional context fields (optional) |
| **Customer Implements** | - Log format validation |
| **Evidence** | - Sample log entries demonstrating all fields |

### AU-6: Audit Record Review, Analysis, and Reporting

**Control:** The organization reviews and analyzes audit records.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Structured JSON log format<br>- SIEM-compatible output<br>- Prometheus metrics for alerting |
| **Customer Configures** | - SIEM integration (Splunk, ELK, etc.)<br>- Alert rules<br>- Dashboards |
| **Customer Implements** | - Log review schedule<br>- Incident escalation based on logs |
| **Evidence** | - SIEM integration configuration<br>- Alert rule definitions<br>- Review records |

### AU-8: Time Stamps

**Control:** The system uses internal clocks to generate timestamps.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - UTC timestamps in all logs<br>- ISO 8601 format<br>- NTP-synchronized containers |
| **Customer Configures** | - NTP servers for nodes |
| **Customer Implements** | - Time synchronization monitoring |
| **Evidence** | - Sample timestamps<br>- NTP configuration |

## Control Family: Configuration Management (CM)

### CM-2: Baseline Configuration

**Control:** The organization develops, documents, and maintains a current baseline configuration of the information system.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Documented Helm chart default values (`values.yaml`, `values/common.yaml`) as the canonical baseline<br>- Immutable container images with pinned versions and SHA-tagged deployments for cloud<br>- Infrastructure-as-code via Terraform for cloud infrastructure<br>- Git-tracked configuration with full change history<br>- Hardening profile support (`dev`, `standard`) with documented differences |
| **Customer Configures** | - Selection of hardening profile (`dev` vs `standard`)<br>- Custom Helm values overlays for environment-specific settings<br>- Terraform variables for cloud infrastructure parameters<br>- NetworkPolicy configurations beyond defaults |
| **Customer Implements** | - Baseline configuration approval and sign-off process<br>- Deviation documentation and justification<br>- Periodic review of baseline against operational configuration |
| **Evidence** | - Git history of `values.yaml` and Terraform files<br>- Helm release manifests (`helm get values`, `helm get manifest`)<br>- Container image SBOMs<br>- Terraform state files |

### CM-6: Configuration Settings

**Control:** The organization establishes and documents mandatory configuration settings for IT products.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Configuration Hardening Guide with recommended settings per DISA STIG alignment<br>- TLS 1.2+ enforcement with configurable cipher suites<br>- Session timeout defaults (15 min idle, 300s token TTL)<br>- Password policy defaults in Keycloak (12-char minimum, complexity, lockout after 5 attempts)<br>- Kubernetes SecurityContext defaults (non-root, read-only root filesystem, dropped capabilities) |
| **Customer Configures** | - FIPS mode enablement (`FIPS_MODE=true`)<br>- MFA enforcement (`REQUIRE_PHISHING_RESISTANT_MFA=true`)<br>- Session and token lifetime tuning<br>- Keycloak password policy customization<br>- Network policy strictness level |
| **Customer Implements** | - Configuration review and approval process<br>- Periodic compliance scanning against documented settings<br>- Deviation tracking and risk acceptance documentation |
| **Evidence** | - Helm values exports showing active configuration<br>- Keycloak realm configuration export<br>- Kubernetes pod security standards audit results<br>- Configuration Hardening Guide completion checklist |

### CM-7: Least Functionality

**Control:** The organization configures the system to provide only essential capabilities and prohibits or restricts the use of unnecessary functions, ports, protocols, and services.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Minimal base container images (distroless/UBI9-minimal) with no shell or package manager in production builds<br>- Kubernetes NetworkPolicy templates restricting pod-to-pod communication to only required paths<br>- Default-deny egress network policies for all namespaces<br>- Service accounts with minimal RBAC permissions (k8s-agent uses scoped ClusterRole)<br>- No unnecessary ports exposed; gRPC and HTTP endpoints on defined ports only |
| **Customer Configures** | - NetworkPolicy enforcement (enable/customize per namespace)<br>- Disable unused Aegis features (e.g., workspace provisioning if not needed)<br>- Firewall/security group rules for cluster nodes<br>- Ingress controller configuration to limit exposed endpoints |
| **Customer Implements** | - Port and protocol scanning to verify minimal exposure<br>- Periodic review of enabled services and features<br>- Documentation of business justification for each enabled capability |
| **Evidence** | - Container image layer analysis showing minimal packages<br>- NetworkPolicy manifests<br>- Port scan results<br>- `kubectl get svc` output showing only required services |

### CM-8: System Component Inventory

**Control:** The organization develops and documents an inventory of system components that accurately reflects the current system and is at a sufficient level of granularity for tracking and reporting.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Software Bill of Materials (SBOM) for each container image (generated during CI/CD build)<br>- Helm chart dependency listing (`Chart.lock` with pinned versions)<br>- Kubernetes resource inventory via controller-runtime (all AegisWorkload CRDs tracked)<br>- Platform API cluster registry with component versions and health status<br>- Container image registry with SHA digests for all deployed images |
| **Customer Configures** | - SBOM export integration with customer asset management tooling<br>- Cluster registration to ensure all spokes are tracked<br>- Image scanning integration (e.g., Trivy, Grype) for vulnerability correlation |
| **Customer Implements** | - Enterprise asset inventory integration<br>- Periodic reconciliation of deployed components against approved inventory<br>- Decommissioning procedures for removed components |
| **Evidence** | - SBOM files (SPDX or CycloneDX format)<br>- `helm list` output across namespaces<br>- Platform API `/clusters` endpoint listing registered components<br>- Container image registry listing with digests |

## Control Family: Contingency Planning (CP)

### CP-2: Contingency Plan

**Control:** The organization develops a contingency plan for the information system that identifies essential missions and business functions.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Hub-and-spoke architecture with inherent spoke isolation (spoke failures do not affect hub or other spokes)<br>- Helm-based deployment enabling rapid redeployment of hub components from version-controlled charts<br>- Terraform infrastructure-as-code for reproducible cloud environment provisioning<br>- Documented disaster recovery procedures in `AGENT_DEPLOYMENT_GUIDE.md` and `terraform/DESTROY_CHECKLIST.md`<br>- Stateless service design (platform-api, proxy, k8s-agent) enabling horizontal scaling and fast restart |
| **Customer Configures** | - RDS automated backup schedule and retention period (default: 7 days)<br>- Multi-AZ deployment for database high availability<br>- EKS cluster auto-scaling parameters<br>- Cross-region replication settings (if required) |
| **Customer Implements** | - Business Continuity Plan (BCP) incorporating Aegis components<br>- Recovery Time Objective (RTO) and Recovery Point Objective (RPO) definitions<br>- Contingency plan testing schedule (annual minimum)<br>- Alternate processing site identification and procedures |
| **Evidence** | - Documented contingency plan referencing Aegis architecture<br>- Terraform state showing infrastructure reproducibility<br>- Helm chart version history enabling rollback<br>- RDS backup configuration and retention policy |

### CP-9: System Backup

**Control:** The organization conducts backups of user-level and system-level information contained in the system.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - PostgreSQL database migrations versioned and tracked in `charts/aegis-services/files/platform-api/migrations/`<br>- Keycloak realm configuration exportable as JSON (`aegis-realm.json`)<br>- All platform configuration stored in Git (Helm values, Terraform, scripts)<br>- Stateless services (platform-api, proxy, k8s-agent) require no application-level backup<br>- Database schema supports point-in-time recovery when backed by RDS |
| **Customer Configures** | - RDS automated backup retention (recommended: 35 days for compliance)<br>- RDS snapshot schedule and cross-region copy<br>- etcd backup schedule for Kubernetes cluster state<br>- Keycloak realm export schedule |
| **Customer Implements** | - Backup verification and restoration testing (quarterly minimum)<br>- Offsite backup storage procedures<br>- Backup encryption key management<br>- Backup media handling and disposal procedures |
| **Evidence** | - RDS backup configuration and snapshot listing<br>- Git repository with complete configuration history<br>- Backup restoration test results and records<br>- Keycloak realm export files with timestamps |

### CP-10: System Recovery and Reconstitution

**Control:** The organization provides for the recovery and reconstitution of the system to a known state after a disruption, compromise, or failure.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Helm chart-based redeployment from Git (full hub recovery in < 30 minutes with existing infrastructure)<br>- Terraform plan/apply for infrastructure reconstitution from code<br>- Database migration framework ensuring schema consistency after recovery<br>- Spoke re-registration via k8s-agent heartbeat (automatic reconnection after hub recovery)<br>- `scripts/deploy.sh` and `scripts/destroy.sh` for scripted recovery workflows |
| **Customer Configures** | - Recovery priority ordering for services (recommended: PostgreSQL, Keycloak, platform-api, proxy)<br>- DNS failover configuration<br>- Load balancer health check thresholds |
| **Customer Implements** | - Recovery testing procedures and schedule<br>- Communication plan during recovery operations<br>- Post-recovery verification checklist<br>- Lessons learned documentation after recovery events |
| **Evidence** | - Recovery procedure documentation<br>- Recovery test results with measured RTO<br>- Helm deployment logs showing successful reconstitution<br>- Spoke reconnection audit logs after hub recovery |

## Control Family: Identification and Authentication (IA)

### IA-2: Identification and Authentication (Organizational Users)

**Control:** The system uniquely identifies and authenticates users.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - OIDC authentication via Keycloak<br>- SAML 2.0 support<br>- Unique user identifiers in all logs<br>- Session binding to authenticated identity |
| **Customer Configures** | - Identity provider federation<br>- User attribute mapping<br>- Authentication policies |
| **Customer Implements** | - User onboarding/verification process |
| **Evidence** | - Keycloak IdP configuration<br>- Authentication logs |

### IA-2(1): Multi-Factor Authentication

**Control:** The system implements MFA for privileged and network access.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - MFA enforcement option (`REQUIRE_PHISHING_RESISTANT_MFA=true`)<br>- AMR claim validation<br>- Rejection of non-MFA tokens when enforced |
| **Customer Configures** | - Keycloak MFA settings<br>- Authenticator requirements<br>- Environment variable to enforce |
| **Customer Implements** | - MFA device management<br>- Recovery procedures |
| **Evidence** | - MFA configuration screenshots<br>- Auth logs showing AMR claims |

### IA-2(12): PIV/CAC Authentication

**Control:** The system accepts PIV credentials.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Keycloak X.509 authentication support<br>- Certificate-to-user mapping |
| **Customer Configures** | - CAC/PIV Keycloak adapter<br>- Certificate trust chain<br>- User attribute extraction |
| **Customer Implements** | - Card reader deployment<br>- Certificate revocation checking |
| **Evidence** | - Keycloak X.509 configuration<br>- Successful CAC auth logs |

### IA-5: Authenticator Management

**Control:** The organization manages system authenticators.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Password policy enforcement via Keycloak<br>- Token expiration controls<br>- Credential storage encryption |
| **Customer Configures** | - Password complexity rules<br>- Password history<br>- Token lifetimes |
| **Customer Implements** | - Initial credential distribution<br>- Password reset process |
| **Evidence** | - Password policy configuration<br>- Token lifetime settings |

## Control Family: Incident Response (IR)

### IR-2: Incident Response Training

**Control:** The organization provides incident response training to system users consistent with assigned roles and responsibilities.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Incident Response Runbook with Aegis-specific procedures (`docs/compliance/customer-docs/incident-response-runbook.md`)<br>- Documented incident classification severity levels (P1-P4) with response time targets<br>- Platform-specific troubleshooting guides (`docs/TROUBLESHOOTING-CLUSTER-DROPDOWN.md`, `docs/aws-tunnel-dev-setup.md`)<br>- Audit log interpretation guidance for security investigations |
| **Customer Configures** | - Incident response team contact list and escalation paths<br>- Communication channels for incident coordination<br>- Integration of Aegis runbook into organizational IR training program |
| **Customer Implements** | - Annual incident response training program for all personnel<br>- Role-specific training for platform administrators<br>- Tabletop exercises incorporating Aegis-specific scenarios<br>- Training records and completion tracking |
| **Evidence** | - Training completion records<br>- Tabletop exercise after-action reports<br>- Incident response team roster with training dates<br>- Aegis runbook acknowledgment records |

### IR-4: Incident Handling

**Control:** The organization implements an incident handling capability for security incidents that includes preparation, detection and analysis, containment, eradication, and recovery.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Structured audit logging for detection (authentication failures, authorization denials, anomalous API calls)<br>- Prometheus metrics and alerting for anomaly detection (rate limiting triggers, error rate spikes)<br>- Keycloak account lockout and session invalidation for containment<br>- Helm rollback capability for eradication (`helm rollback` to known-good release)<br>- Spoke isolation capability (hub can disconnect compromised spoke clusters)<br>- Workload termination via Platform API for compromised workloads |
| **Customer Configures** | - Alert thresholds for detection triggers<br>- SIEM correlation rules for Aegis-specific events<br>- Automated containment actions (e.g., Keycloak lockout threshold)<br>- Incident ticket system integration |
| **Customer Implements** | - Incident handling procedures incorporating Aegis capabilities<br>- Forensic evidence collection procedures<br>- Root cause analysis process<br>- Post-incident review and lessons learned documentation<br>- External notification procedures (customers, regulators) |
| **Evidence** | - Incident ticket records with timeline<br>- Audit logs from incident period<br>- Containment action logs (account lockouts, session terminations)<br>- Post-incident review reports |

### IR-5: Incident Monitoring

**Control:** The organization tracks and documents information system security incidents.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Real-time security event logging across all platform components<br>- Prometheus metrics for security-relevant indicators (auth failures/sec, API error rates, unauthorized access attempts)<br>- Structured JSON log format enabling automated monitoring and correlation<br>- Cluster health monitoring with stale-since tracking for spoke connectivity anomalies<br>- Workload lifecycle event tracking (unexpected terminations, privilege escalation attempts) |
| **Customer Configures** | - SIEM integration for centralized incident monitoring<br>- Alert rules for security-relevant thresholds<br>- Dashboard creation for incident trend analysis<br>- Log retention policies meeting compliance requirements |
| **Customer Implements** | - Incident tracking system (ticketing)<br>- Incident trend analysis and reporting (monthly/quarterly)<br>- Metrics for incident response effectiveness (MTTD, MTTR)<br>- Management reporting on incident trends |
| **Evidence** | - SIEM dashboards showing incident metrics<br>- Incident tracking system reports<br>- Trend analysis documents<br>- Management briefing records |

### IR-6: Incident Reporting

**Control:** The organization requires personnel to report suspected security incidents to the organizational incident response capability and reports security incident information to designated authorities.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Audit trail enabling incident reconstruction with timestamps, user identities, and affected resources<br>- Exportable log formats compatible with reporting requirements (JSON, SIEM-compatible)<br>- Incident report templates in the Incident Response Runbook<br>- API endpoints for programmatic extraction of security events for reporting |
| **Customer Configures** | - Incident reporting channels and contact information<br>- Automated alerts for reportable event categories<br>- Log export schedules for reporting compliance |
| **Customer Implements** | - Incident reporting policy and procedures<br>- Regulatory reporting obligations (US-CERT, CISA, sector-specific ISACs)<br>- Customer notification procedures per contractual SLAs<br>- Internal reporting to management and governance bodies |
| **Evidence** | - Submitted incident reports<br>- US-CERT/CISA reporting confirmations<br>- Customer notification records<br>- Internal incident briefings |

## Control Family: Maintenance (MA)

### MA-2: Controlled Maintenance

**Control:** The organization schedules, performs, documents, and reviews records of maintenance and repairs on system components.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Helm chart versioning with documented release notes for all maintenance activities<br>- Rolling update deployment strategy (zero-downtime for platform-api, proxy)<br>- Database migration framework with up/down migrations for schema changes<br>- `scripts/deploy.sh` for consistent, repeatable maintenance procedures<br>- Git-tracked change history for all configuration and code changes |
| **Customer Configures** | - Maintenance window scheduling (recommended: low-traffic periods)<br>- Pre-maintenance backup verification<br>- Notification procedures for planned maintenance<br>- Rollback trigger criteria |
| **Customer Implements** | - Maintenance request and approval process<br>- Maintenance activity logging and documentation<br>- Post-maintenance verification procedures<br>- Maintenance records retention (minimum 1 year) |
| **Evidence** | - Helm release history (`helm history`)<br>- Git commit and PR history for maintenance changes<br>- Maintenance window schedules and notifications<br>- Post-maintenance verification results |

### MA-5: Maintenance Personnel

**Control:** The organization establishes a process for maintenance personnel authorization and maintains a list of authorized maintenance organizations or personnel.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Keycloak RBAC ensuring only authorized roles (Admin) can perform maintenance operations<br>- Audit logging of all administrative and maintenance actions with user identity<br>- Service account restrictions preventing unauthorized automated maintenance<br>- kubectl and Helm access gated by Kubernetes RBAC (cluster-admin role required for maintenance) |
| **Customer Configures** | - Keycloak role assignments limiting maintenance permissions to authorized personnel<br>- Kubernetes RBAC bindings for cluster maintenance access<br>- AWS IAM policies for infrastructure maintenance (EKS, RDS, VPC) |
| **Customer Implements** | - Authorized maintenance personnel list with current clearance/authorization status<br>- Escort procedures for non-authorized personnel (if applicable)<br>- Background check requirements for maintenance personnel<br>- Annual review and re-authorization of maintenance personnel |
| **Evidence** | - Keycloak role assignment exports for Admin roles<br>- Kubernetes RBAC binding manifests<br>- AWS IAM policy attachments<br>- Authorized personnel list with review dates |

## Control Family: Physical and Environmental Protection (PE)

### PE-2: Physical Access Authorizations

**Control:** The organization develops, approves, and maintains a list of individuals with authorized access to the facility where the information system resides.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - **Inherited from AWS**: All physical access controls for EKS compute infrastructure, RDS database servers, and underlying AWS data center facilities are provided by AWS and documented in AWS SOC 2 Type II and FedRAMP authorization packages<br>- For local Docker Desktop deployments: not applicable (developer workstation) |
| **Customer Configures** | - Not applicable for cloud-hosted components (inherited from AWS) |
| **Customer Implements** | - Physical access authorization for any customer-managed infrastructure (on-premises spoke clusters, if applicable)<br>- Review of AWS physical security controls via AWS compliance reports<br>- Physical access to developer workstations running local deployments |
| **Evidence** | - AWS SOC 2 Type II report (physical security sections)<br>- AWS FedRAMP authorization package<br>- Customer facility access logs (for on-premises components, if any) |

### PE-3: Physical Access Control

**Control:** The organization enforces physical access authorizations at entry/exit points to the facility where the information system resides.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - **Inherited from AWS**: Physical access control mechanisms (biometric readers, mantraps, CCTV, 24/7 security guards) for all AWS data centers hosting EKS and RDS infrastructure are managed by AWS<br>- Aegis platform operates entirely within Kubernetes on cloud infrastructure with no Aegis-owned physical facilities |
| **Customer Configures** | - Not applicable for cloud-hosted components (inherited from AWS) |
| **Customer Implements** | - Physical access controls for customer data centers (if hosting on-premises spoke clusters)<br>- Visitor management for facilities housing any non-cloud Aegis components<br>- Annual review of AWS physical security compliance reports |
| **Evidence** | - AWS SOC 2 Type II report (physical access control sections)<br>- AWS data center compliance certifications (ISO 27001, SOC 2)<br>- Customer facility physical access logs (if applicable) |

## Control Family: Planning (PL)

### PL-1: Planning Policy and Procedures

**Control:** The organization develops, documents, and disseminates a security planning policy and procedures.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Comprehensive compliance documentation suite (`docs/compliance/`) including control mapping, gap analysis, and SSP outline<br>- Customer-facing security documentation (Security Architecture Guide, Control Implementation Statements, Customer Responsibility Matrix)<br>- FedRAMP readiness artifacts (scope assessment, control mapping, appendices)<br>- Annual policy review cadence documented in all policy documents |
| **Customer Configures** | - Integration of Aegis security documentation into organizational security planning framework<br>- Mapping of Aegis-provided controls to customer's SSP structure |
| **Customer Implements** | - System Security Plan (SSP) development and maintenance<br>- Plan of Action and Milestones (POA&M) tracking for identified gaps<br>- Annual planning policy review and update cycle<br>- Dissemination of security planning documentation to relevant personnel |
| **Evidence** | - Completed SSP incorporating Aegis Control Implementation Statements<br>- POA&M tracking document<br>- Policy review records<br>- Distribution and acknowledgment records |

### PL-2: System Security Plan

**Control:** The organization develops a system security plan that describes the system boundary, operational environment, security requirements, and implementation of security controls.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Security Architecture Guide (`docs/compliance/customer-docs/security-architecture-guide.md`) with system boundary diagrams, data flow diagrams, and component inventory<br>- Control Implementation Statements (this document) for SSP Section 13<br>- Customer Responsibility Matrix defining shared responsibility boundaries<br>- Network architecture documentation (`docs/networking-architecture.md`)<br>- Infrastructure reference documentation (`docs/infrastructure-reference.md`) |
| **Customer Configures** | - System boundary definition incorporating Aegis components<br>- Interconnection details for federated identity providers and external systems<br>- Environment-specific configuration documentation |
| **Customer Implements** | - SSP authoring per NIST SP 800-18 or FedRAMP template<br>- Authorization boundary determination<br>- Interconnection Security Agreements (ISAs) for connected systems<br>- SSP review and update (at least annually or upon significant change)<br>- ATO package preparation and submission |
| **Evidence** | - Completed SSP document<br>- Authorization boundary diagram<br>- Interconnection agreements<br>- ATO decision letter |

## Control Family: Personnel Security (PS)

### PS-3: Personnel Screening

**Control:** The organization screens individuals prior to authorizing access to the information system.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Keycloak account provisioning workflow supporting multi-step approval before access is granted<br>- SAML/OIDC federation enabling integration with enterprise HR and identity proofing systems<br>- Audit logging of initial account creation and first-access events |
| **Customer Configures** | - Keycloak registration flow to require administrative approval before account activation<br>- Identity provider federation to tie access grants to HR-verified identity |
| **Customer Implements** | - Background investigation procedures appropriate to system risk level<br>- Screening criteria aligned with position sensitivity designations<br>- Re-investigation schedule for personnel with ongoing access<br>- Documentation of screening completion before access provisioning |
| **Evidence** | - Personnel screening completion records<br>- Keycloak account creation timestamps correlated with screening dates<br>- HR-to-IT access provisioning workflow documentation |

### PS-4: Personnel Termination

**Control:** Upon termination of individual employment, the organization disables information system access within defined time periods, terminates authenticators, and retrieves all organizational information system-related property.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Keycloak account disable/delete functionality with immediate session invalidation<br>- OIDC token revocation propagating across all Aegis services within token expiration window<br>- Audit logging of account disable events and subsequent access attempts<br>- Kubernetes RBAC binding removal capability for direct cluster access |
| **Customer Configures** | - Keycloak session timeout and token lifetime settings (shorter values = faster access revocation upon termination)<br>- SCIM provisioning integration for automated account lifecycle (if available)<br>- Alert rules for access attempts by disabled accounts |
| **Customer Implements** | - Termination notification procedures (HR to IT, < 24 hours)<br>- Credential revocation checklist (Keycloak, kubectl, AWS IAM, VPN)<br>- Property retrieval procedures (hardware, tokens, badges)<br>- Exit interview and access confirmation procedures |
| **Evidence** | - Keycloak account disable audit logs with timestamps<br>- Access attempt logs after disable timestamp (demonstrating enforcement)<br>- Termination checklist completion records<br>- Credential rotation records for shared accounts |

### PS-5: Personnel Transfer

**Control:** The organization reviews and confirms ongoing operational need for current logical and physical access authorizations when individuals are reassigned or transferred.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Keycloak role and group management enabling granular access changes without account recreation<br>- Project-level access boundaries in Platform API enabling reassignment between projects<br>- Audit logging of all role and group membership changes<br>- RBAC enforcement ensuring access changes take effect immediately upon role modification |
| **Customer Configures** | - Keycloak group and role structures aligned with organizational functions<br>- Project membership assignments reflecting current responsibilities<br>- Automated role change workflows via SCIM or Keycloak admin API |
| **Customer Implements** | - Transfer access review procedures (review within 5 business days of transfer)<br>- Access re-certification for transferred personnel<br>- Communication between sending and receiving managers regarding access needs<br>- Documentation of access changes and business justification |
| **Evidence** | - Keycloak role change audit logs<br>- Access review records associated with transfer events<br>- Manager authorization for access modifications<br>- Before/after role comparison documentation |

## Control Family: System and Communications Protection (SC)

### SC-8: Transmission Confidentiality and Integrity

**Control:** The system protects data in transit.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - TLS 1.2+ for all external connections<br>- mTLS for internal service communication<br>- FIPS cipher suites (when enabled) |
| **Customer Configures** | - TLS certificates<br>- Cipher suite selection<br>- FIPS mode enablement |
| **Customer Implements** | - Certificate lifecycle management |
| **Evidence** | - TLS configuration<br>- SSL Labs scan results<br>- Network capture showing encryption |

### SC-12: Cryptographic Key Management

**Control:** The organization establishes and manages cryptographic keys.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - step-ca integration (optional)<br>- cert-manager automation<br>- Key rotation support<br>- Kubernetes Secrets for key storage |
| **Customer Configures** | - CA hierarchy<br>- Certificate lifetimes<br>- Rotation schedules |
| **Customer Implements** | - Root CA protection<br>- Key recovery procedures |
| **Evidence** | - PKI configuration<br>- Certificate inventory<br>- Rotation logs |

### SC-13: Cryptographic Protection

**Control:** The system implements cryptographic mechanisms.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - FIPS 140-2/3 capable builds<br>- AES-256-GCM for encryption<br>- RSA-2048/ECDSA-256 for signatures<br>- BoringCrypto (Go), FIPS OpenSSL (Node.js) |
| **Customer Configures** | - FIPS mode enablement<br>- Crypto policy selection |
| **Customer Implements** | - FIPS validation evidence for auditors |
| **Evidence** | - FIPS module configuration<br>- Crypto library versions |

## Control Family: Supply Chain Risk Management (SR)

### SR-2: Supply Chain Risk Management Plan

**Control:** The organization develops a supply chain risk management plan that addresses risks associated with the development, acquisition, maintenance, and disposal of systems, components, and services.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Vendor Management Policy (VMP-001) governing all third-party dependencies<br>- Tiered vendor risk classification (Critical, High, Medium, Low) with assessment requirements per tier<br>- Current vendor inventory maintained with risk tiers (GitHub: Critical, AWS: Critical)<br>- Open-source dependency management via Go modules with checksum verification (`go.sum`)<br>- Container base image provenance tracking (UBI9, Alpine with documented sources) |
| **Customer Configures** | - Integration of Aegis supply chain controls into organizational SCRM plan<br>- Vendor assessment requirements for Aegis as a supplier<br>- Acceptable risk thresholds for third-party components |
| **Customer Implements** | - Organizational supply chain risk management plan<br>- Supply chain risk assessment for Aegis and its dependencies<br>- Ongoing monitoring of supply chain threats and vulnerabilities<br>- Supply chain incident response procedures |
| **Evidence** | - Aegis Vendor Management Policy and vendor inventory<br>- Go module dependency listing with checksums<br>- Container base image provenance documentation<br>- SBOM files for all Aegis container images |

### SR-3: Supply Chain Controls and Processes

**Control:** The organization establishes controls and processes to address supply chain risks.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Automated dependency vulnerability scanning in CI/CD pipeline (Dependabot, `go vet`)<br>- Container image scanning for known vulnerabilities before deployment<br>- Pinned dependency versions with checksum verification in `go.sum` and `Chart.lock`<br>- GitHub branch protection preventing unauthorized code introduction<br>- Signed container images with SHA digest pinning for cloud deployments<br>- SBOM generation for supply chain transparency |
| **Customer Configures** | - Image scanning policies and severity thresholds<br>- Approved container registry sources<br>- Dependency update review cadence |
| **Customer Implements** | - Periodic review of Aegis SBOM for newly disclosed vulnerabilities<br>- Integration testing of Aegis updates before production deployment<br>- Supply chain compromise detection monitoring<br>- Procedures for responding to compromised dependencies (e.g., CVE in base image) |
| **Evidence** | - CI/CD pipeline scan results<br>- Dependabot alerts and resolution history<br>- Container image scan reports<br>- SBOM diff between versions showing dependency changes |

### SR-5: Acquisition Strategies, Tools, and Methods

**Control:** The organization employs acquisition strategies, contract tools, and procurement methods to protect against, identify, and mitigate supply chain risks.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Open-source licensing compliance (all Aegis dependencies use OSI-approved licenses)<br>- Vendor contractual security requirements documented in Vendor Management Policy<br>- SOC 2 Type II report availability for critical vendors (GitHub, AWS)<br>- Multiple sourcing options for critical dependencies (e.g., container base images available from multiple registries)<br>- Evaluation criteria for new third-party components (security, licensing, maintenance activity) |
| **Customer Configures** | - Procurement approval workflow for Aegis platform and updates<br>- License compliance scanning integration<br>- Approved vendor/supplier list including Aegis Technologies |
| **Customer Implements** | - Acquisition security requirements in contracts with Aegis Technologies<br>- Independent security assessment of Aegis before deployment<br>- Ongoing supplier performance monitoring<br>- Contractual provisions for security incident notification, audit rights, and data protection |
| **Evidence** | - Aegis license compliance report<br>- Vendor security assessment documentation<br>- Contract terms addressing supply chain security<br>- Supplier performance review records |

## Control Family: System and Information Integrity (SI)

### SI-4: System Monitoring

**Control:** The organization monitors the system for attacks and indicators of compromise.

| Responsibility | Implementation |
|----------------|----------------|
| **Aegis Provides** | - Prometheus metrics export<br>- Authentication failure counters<br>- Rate limiting alerts<br>- Structured security logs |
| **Customer Configures** | - Alert thresholds<br>- SIEM integration<br>- Dashboards |
| **Customer Implements** | - SOC monitoring<br>- Incident response triggers |
| **Evidence** | - Monitoring configuration<br>- Sample alerts<br>- Dashboard screenshots |

## Control Summary Matrix

| Control | Aegis Implements | Customer Configures | Customer Implements |
|---------|-----------------|--------------------|--------------------|
| AC-2 | YES | YES | YES |
| AC-3 | YES | YES | YES |
| AC-6 | YES | YES | YES |
| AC-11 | YES | YES | -- |
| AC-17 | YES | YES | YES |
| AU-2 | YES | YES | YES |
| AU-3 | YES | -- | -- |
| AU-6 | YES | YES | YES |
| AU-8 | YES | YES | -- |
| CM-2 | YES | YES | YES |
| CM-6 | YES | YES | YES |
| CM-7 | YES | YES | YES |
| CM-8 | YES | YES | YES |
| CP-2 | YES | YES | YES |
| CP-9 | YES | YES | YES |
| CP-10 | YES | YES | YES |
| IA-2 | YES | YES | YES |
| IA-2(1) | YES | YES | YES |
| IA-2(12) | YES | YES | YES |
| IA-5 | YES | YES | YES |
| IR-2 | YES | YES | YES |
| IR-4 | YES | YES | YES |
| IR-5 | YES | YES | YES |
| IR-6 | YES | YES | YES |
| MA-2 | YES | YES | YES |
| MA-5 | YES | YES | YES |
| PE-2 | Inherited (AWS) | -- | YES |
| PE-3 | Inherited (AWS) | -- | YES |
| PL-1 | YES | YES | YES |
| PL-2 | YES | YES | YES |
| PS-3 | YES | YES | YES |
| PS-4 | YES | YES | YES |
| PS-5 | YES | YES | YES |
| SC-8 | YES | YES | YES |
| SC-12 | YES | YES | YES |
| SC-13 | YES | YES | -- |
| SI-4 | YES | YES | YES |
| SR-2 | YES | YES | YES |
| SR-3 | YES | YES | YES |
| SR-5 | YES | YES | YES |

**Total Controls: 40** (expanded from 17 in v1.0)

**Coverage by Family:**
- Access Control (AC): 5 controls
- Audit and Accountability (AU): 4 controls
- Configuration Management (CM): 4 controls
- Contingency Planning (CP): 3 controls
- Identification and Authentication (IA): 4 controls
- Incident Response (IR): 4 controls
- Maintenance (MA): 2 controls
- Physical and Environmental Protection (PE): 2 controls (inherited from AWS)
- Planning (PL): 2 controls
- Personnel Security (PS): 3 controls
- System and Communications Protection (SC): 3 controls
- Supply Chain Risk Management (SR): 3 controls
- System and Information Integrity (SI): 1 control

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0-DRAFT | TBD | TBD | Initial draft with 17 controls |
| 2.0.0-DRAFT | 2026-03-09 | Carlos Sanchez | Expanded to 40 controls; added CM, CP, IR, MA, PE, PL, PS, SR families |
