# Aegis Platform -- System Security Plan (SSP): Sections 1-3

**Document ID:** AEGIS-SSP-001
**System Unique Identifier:** AEGIS-2026-001
**Version:** 1.0
**Date:** 2026-03-09
**Status:** DRAFT -- PRE-AUTHORIZATION
**Baseline:** FedRAMP Moderate (with LOW override eligibility for Model B self-hosted deployments)
**Classification:** Controlled Unclassified Information (CUI) -- Pre-Decisional

---

## Document Control

| Role | Name | Organization | Date |
|------|------|-------------|------|
| System Owner | Carlos Sanchez | Aegis Technologies | 2026-03-09 |
| Information System Security Officer (ISSO) | Carlos Sanchez | Aegis Technologies | 2026-03-09 |
| Authorizing Official (AO) | TBD | Sponsoring Agency | Pending |
| 3PAO Assessor | TBD | TBD | Pending |

### Revision History

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1 | 2026-01-17 | Carlos Sanchez | Initial scope and impact assessment |
| 0.2 | 2026-01-17 | Carlos Sanchez | SSP outline and appendix structure |
| 1.0 | 2026-03-09 | Carlos Sanchez | SSP Sections 1-3 consolidated draft |

---

# Section 1: Information System Name, Title, and Unique Identifier

## 1.1 System Name and Title

| Attribute | Value |
|-----------|-------|
| **System Name** | Aegis Platform |
| **System Abbreviation** | AEGIS |
| **System Title** | Aegis Platform -- Multi-Cluster GPU Workload Orchestration |
| **Unique Identifier** | AEGIS-2026-001 |
| **FedRAMP Package ID** | Pending assignment |
| **Cloud Service Offering (CSO)** | Aegis Platform |
| **Cloud Service Provider (CSP)** | Aegis Technologies |

## 1.2 System Categorization (FIPS 199)

The security categorization of the Aegis Platform information system is determined in accordance with Federal Information Processing Standard (FIPS) Publication 199, *Standards for Security Categorization of Federal Information and Information Systems*, and NIST Special Publication 800-60 Volume II, *Guide for Mapping Types of Information and Information Systems to Security Categories*.

### 1.2.1 Information Types Processed

The following information types are processed, stored, or transmitted by Aegis Platform, categorized per NIST SP 800-60:

| Information Type (SP 800-60) | Confidentiality | Integrity | Availability | Rationale |
|------------------------------|:-:|:-:|:-:|-----------|
| System Credentials (user authentication data) | M | M | L | Usernames, email addresses, JWT tokens for authentication. Password handling delegated to external IdP. |
| Workload Metadata | L | M | L | Job names, resource requests (CPU/GPU/memory), scheduling parameters, queue assignments. No customer application data. |
| Infrastructure Configuration | M | H | M | Cluster configurations, network settings, Helm values, Terraform state. Unauthorized modification could compromise system integrity. |
| Audit and Accountability Logs | L | M | M | System activity records, user actions, authentication events, authorization decisions. Required for continuous monitoring. |
| Application Code References | L | M | L | Container image references for deployed workloads. Actual application code and data remain in customer boundary. |

**Legend:** L = Low, M = Moderate, H = High

### 1.2.2 High-Water Mark Calculation

Per FIPS 199, the overall system categorization is determined by the highest impact level across all information types for each security objective:

| Security Objective | Maximum Impact Level | Driving Information Type |
|--------------------|:--------------------:|--------------------------|
| **Confidentiality** | **MODERATE** | System credentials, infrastructure configuration |
| **Integrity** | **HIGH** | Infrastructure configuration (unauthorized modification could impact all managed clusters) |
| **Availability** | **MODERATE** | Audit logs, infrastructure configuration (GPU workload continuity) |

### 1.2.3 Security Categorization Expression

Per FIPS 199, the security categorization is expressed as:

```
SC (Aegis Platform) = {(Confidentiality, MODERATE), (Integrity, HIGH), (Availability, MODERATE)}
```

### 1.2.4 Overall System Categorization: MODERATE

**Overall Categorization: MODERATE**

The initial high-water mark assessment yields HIGH due to infrastructure configuration integrity. However, the overall system categorization is determined as MODERATE based on the following analysis:

The HIGH integrity rating for infrastructure configuration reflects the potential impact of unauthorized modification to cluster configurations managed through the platform. In the CSP Deployment Model B (self-hosted), the customer retains operational control over their infrastructure, and Aegis Platform acts as an orchestration layer rather than the sole custodian of infrastructure configuration. The integrity impact is mitigated by:

- Infrastructure-as-Code (IaC) with version control providing configuration auditability and rollback capability
- GitOps-based deployment processes that enforce change management controls
- Kubernetes RBAC limiting configuration modification to authorized roles
- Separation between the Aegis control plane and customer workload execution boundaries

For **Model B (self-hosted) deployments**, a LOW categorization override may be applicable per the conditions documented in Section 1.2.5. The Moderate baseline is maintained as the default to ensure controls are sufficient for all deployment scenarios.

### 1.2.5 Conditions for LOW Impact Override (Model B Self-Hosted)

For deployments where the customer self-hosts Aegis Platform in their own authorization boundary (CSP Deployment Model B), a LOW impact categorization may apply when all of the following conditions are met:

| Condition | Rationale |
|-----------|-----------|
| Platform handles minimal PII (username, email for authentication only) | No sensitive PII processing by Aegis components |
| Customer controls all data within their authorization boundary | Aegis does not access or store customer workload data |
| Infrastructure configuration is customer-controlled | Customer operates and maintains their own environment |
| No Controlled Unclassified Information (CUI) processing by Aegis | CUI exists only in customer workloads, outside the Aegis boundary |

When all conditions are satisfied:

```
SC (Aegis Platform - Model B) = {(Confidentiality, LOW), (Integrity, LOW), (Availability, LOW)}
```

This qualifies the system for **LI-SaaS** or **FedRAMP 20x Low** authorization pathways.

### 1.2.6 Security Objectives

The following security objectives apply to the Aegis Platform:

| Security Objective | Description | Implementation Approach |
|--------------------|-------------|------------------------|
| **Confidentiality** | Prevent unauthorized disclosure of system credentials, infrastructure configuration, and workload metadata | TLS 1.2+ for all data in transit, encryption at rest via AWS KMS (AES-256), RBAC-enforced access controls, Keycloak OIDC with MFA |
| **Integrity** | Prevent unauthorized modification of infrastructure configuration, workload definitions, and audit records | GitOps-based change management, immutable container images, digitally signed JWT tokens, mTLS for hub-spoke communication, append-only audit logging |
| **Availability** | Ensure continued access to the platform for workload management and GPU scheduling | Kubernetes self-healing (pod restart, node replacement), RDS Multi-AZ (cloud), health-check-based routing, defined RTO/RPO targets |

## 1.3 System Owner

| Attribute | Value |
|-----------|-------|
| **Organization** | Aegis Technologies |
| **System Owner** | Carlos Sanchez, Founder and CEO |
| **Email** | carlos@aegis-platform.tech |
| **Phone** | [REDACTED] |
| **Address** | [REDACTED] |

## 1.4 Authorizing Official

| Attribute | Value |
|-----------|-------|
| **Name** | TBD (Sponsoring Federal Agency) |
| **Title** | Authorizing Official (AO) |
| **Organization** | Pending agency sponsorship |
| **Email** | Pending |

## 1.5 Other Key Contacts

| Role | Name | Organization | Responsibility |
|------|------|-------------|----------------|
| Information System Security Officer (ISSO) | Carlos Sanchez | Aegis Technologies | Day-to-day security operations, continuous monitoring, incident response |
| Chief Information Security Officer (CISO) | Carlos Sanchez | Aegis Technologies | Security policy, risk acceptance, security program oversight |
| System Administrator | Carlos Sanchez | Aegis Technologies | System operations, configuration management, patching |
| 3PAO Lead Assessor | TBD | TBD | Independent security assessment |

> **Note:** Aegis Technologies is a solo-founder organization. Compensating controls for separation of duties are documented in Appendix L (Separation of Duties Matrix). As the organization scales, roles will be distributed to dedicated personnel.

## 1.6 Assignment of Security Responsibility

| Security Function | Responsible Party | Description |
|-------------------|-------------------|-------------|
| Security policy development | CISO (Carlos Sanchez) | Establish and maintain information security policies aligned with NIST 800-53 Rev 5 |
| Access control management | ISSO (Carlos Sanchez) | Manage user accounts, RBAC roles, MFA enforcement via Keycloak |
| Configuration management | System Administrator (Carlos Sanchez) | Maintain baseline configurations, manage changes via GitOps |
| Continuous monitoring | ISSO (Carlos Sanchez) | Execute monthly vulnerability scans, quarterly access reviews, annual assessments |
| Incident response | ISSO (Carlos Sanchez) | Detect, respond to, and report security incidents per FedRAMP requirements |
| Contingency planning | System Owner (Carlos Sanchez) | Maintain and test system contingency and disaster recovery plans |
| Risk management | CISO (Carlos Sanchez) | Conduct risk assessments, maintain risk register, manage POA&M |
| Audit and accountability | ISSO (Carlos Sanchez) | Configure and review audit logs, ensure log integrity and retention |

## 1.7 E-Authentication Determination

The E-Authentication risk assessment for Aegis Platform is documented in **Appendix E: Digital Identity Worksheet**. Preliminary determination:

| Assurance Level | Value | Justification |
|-----------------|-------|---------------|
| **Identity Assurance Level (IAL)** | 1 (Self-asserted) | Minimal PII collection; identity proofing not required for platform access. Users authenticated via federated IdP. |
| **Authenticator Assurance Level (AAL)** | 2 (MFA required) | Platform manages access to sensitive infrastructure; multi-factor authentication enforced via Keycloak. |
| **Federation Assurance Level (FAL)** | 2 (Encrypted assertions) | SAML/OIDC federation with customer identity providers; assertions are digitally signed and encrypted. |

The full E-Authentication analysis, including risk assessment per NIST SP 800-63-3, is provided in Appendix E.

---

# Section 2: Information System Description

## 2.1 System Function and Purpose

Aegis Platform is a compliance-first, multi-cluster Kubernetes control plane that orchestrates GPU and CPU workloads across distributed clusters while enforcing security compliance boundaries. The platform provides federal agencies and regulated organizations with:

- **Unified multi-cluster management**: A central hub that discovers, registers, and monitors spoke clusters across regions and cloud providers
- **Intelligent workload scheduling**: Automated placement of GPU/CPU workloads based on cluster capacity, flavor availability, heartbeat freshness, and time-to-first-GPU metrics
- **Budget enforcement**: Per-queue spending limits with configurable HARD (reject) and SOFT (warn) policy modes, including cost estimation from GPU-hour pricing
- **Interactive development workspaces**: Secure remote development environments with VS Code integration through authenticated WebSocket tunnels
- **Multi-tenant project isolation**: Project-based resource isolation with per-project credentials, policy domain enforcement, and data classification levels
- **Compliance automation**: Built-in audit logging, OSCAL evidence export, session management controls (AC-11/AC-12), and policy domain enforcement with region and information level constraints

### 2.1.1 Mission and Business Objectives

Aegis Platform supports the following mission objectives for federal customers:

| Objective | Description |
|-----------|-------------|
| Accelerate AI/ML workloads | Enable researchers and engineers to access GPU resources across clusters with p95 < 90s time-to-first-GPU |
| Enforce compliance boundaries | Automated enforcement of data residency, information level classification, and access control policies |
| Reduce operational overhead | Self-service workload submission and cluster management through a unified interface |
| Enable multi-cloud operations | Manage workloads across AWS, on-premises, and air-gapped environments from a single control plane |
| Support secure remote development | Provide browser-based and VS Code-based access to GPU-accelerated development environments |

### 2.1.2 System Status

**Operational Status:** Under Development / Pre-Authorization

| Milestone | Status | Target Date |
|-----------|--------|-------------|
| System development | In Progress | Ongoing |
| SOC 2 Type I certification | Complete | 2025 |
| ISO 27001 alignment | Complete | 2025 |
| FedRAMP SSP development | In Progress | Q2 2026 |
| 3PAO readiness assessment | Planned | Q3 2026 |
| Agency sponsorship | Pending | Q3 2026 |
| FedRAMP authorization | Planned | Q4 2026 |

## 2.2 Platform Architecture Overview

Aegis Platform follows a hub-and-spoke architectural pattern in which a centralized hub cluster provides control plane services and one or more spoke clusters execute workloads.

### 2.2.1 Hub Components

The hub cluster runs in the `aegis-system` namespace (with supporting services in dedicated namespaces) and contains the following components:

| Component | Technology | Purpose | Namespace |
|-----------|-----------|---------|-----------|
| **Platform API** | Go, gRPC + REST gateway | Central orchestration: workload placement, budget enforcement, cluster management, proxy ticket minting, session management | `aegis-system` |
| **Proxy** | Go, WebSocket | Authenticated reverse proxy for VS Code workspace connections using JWT + mTLS | `aegis-system` |
| **Keycloak** | Java, OpenID Connect | Identity provider for SSO, MFA enforcement, SAML/OIDC federation with customer IdPs | `keycloak` |
| **PostgreSQL** | PostgreSQL 15+ | Persistent storage for cluster state, workload records, project configuration, audit logs. RDS in cloud deployments. | `aegis-system` (local) / RDS (cloud) |
| **ingress-nginx** | NGINX Ingress Controller | TLS termination and HTTP/gRPC routing for external access | `aegis-system` |

### 2.2.2 Spoke Components

Each spoke cluster runs the following components, deployed via the `aegis-spoke` Helm chart:

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **K8s Agent** | Go, controller-runtime (Kubernetes Operator) | Per-cluster operator that registers with the hub, sends heartbeats, reconciles `AegisWorkload` CRDs into Kubernetes Jobs with Kueue integration |
| **Spoke Proxy** | Go, WebSocket/TLS | Secure tunnel endpoint for user connections to workspaces running in the spoke cluster |

Spoke agents communicate outbound-only to the hub via gRPC over mTLS. No inbound connections to spoke clusters are required for control plane operations.

### 2.2.3 Supporting Infrastructure

| Component | Technology | Purpose | Namespace / Location |
|-----------|-----------|---------|---------------------|
| **step-ca** | Smallstep Certificate Authority | Internal PKI for automated certificate issuance (mTLS, service identity) | `aegis-pki` |
| **cert-manager** | Jetstack cert-manager | Kubernetes-native certificate lifecycle management | `cert-manager` |
| **step-issuer** | Smallstep step-issuer | CertManager ClusterIssuer integration with step-ca | `cert-manager` |
| **AWS EKS** | Amazon Elastic Kubernetes Service | Managed Kubernetes control plane (cloud deployments) | AWS |
| **AWS VPC** | Amazon Virtual Private Cloud | Network isolation with public and private subnets | AWS |
| **AWS RDS** | Amazon Relational Database Service | Managed PostgreSQL with Multi-AZ, encryption at rest | AWS |
| **AWS ECR** | Amazon Elastic Container Registry | Container image registry for cloud deployments | AWS |
| **AWS Route53** | Amazon Route53 | DNS management for `aegist.dev` domain | AWS |
| **AWS CloudWatch** | Amazon CloudWatch | Log aggregation and monitoring | AWS |
| **AWS NLB** | Amazon Network Load Balancer | Layer-4 load balancing for gRPC and spoke-proxy traffic | AWS |

### 2.2.4 Architecture Diagram (Textual)

```
                        +-------------------------------------+
                        |         Aegis Hub Cluster            |
                        |           (aegis-system)             |
                        |                                     |
  +----------+          |  +--------------+  +-------------+  |
  | Aegis UI |--REST--> |  | Platform API |  | PostgreSQL  |  |
  |(Backstage)|         |  | (gRPC+REST)  |<-|  (RDS/Local)|  |
  +----------+          |  +------+-------+  +-------------+  |
                        |         |                           |
  +----------+          |  +------v-------+  +-------------+  |
  |  Sovran  |--gRPC--> |  |   Keycloak   |  | cert-manager|  |
  |(VS Code) |          |  |   (OIDC)     |  |  + step-ca  |  |
  +----------+          |  +--------------+  +-------------+  |
                        +-----------+---------+---------------+
                                    | gRPC (mTLS)
              +---------------------+---------------------+
              v                     v                     v
  +-------------------+  +-------------------+  +-------------------+
  |  Spoke Cluster A  |  |  Spoke Cluster B  |  |  Spoke Cluster N  |
  |                   |  |                   |  |                   |
  |  +-------------+  |  |  +-------------+  |  |  +-------------+  |
  |  |  K8s Agent   |  |  |  |  K8s Agent   |  |  |  |  K8s Agent   |  |
  |  |  (Operator)  |  |  |  |  (Operator)  |  |  |  |  (Operator)  |  |
  |  +------+------+  |  |  +------+------+  |  |  +------+------+  |
  |         v         |  |         v         |  |         v         |
  |  +-------------+  |  |  +-------------+  |  |  +-------------+  |
  |  |  Workload   |  |  |  |  Workload   |  |  |  |  Workload   |  |
  |  |    Pods     |  |  |  |    Pods     |  |  |  |    Pods     |  |
  |  |  (GPU/CPU)  |  |  |  |  (GPU/CPU)  |  |  |  |  (GPU/CPU)  |  |
  |  +-------------+  |  |  +-------------+  |  |  +-------------+  |
  +-------------------+  +-------------------+  +-------------------+
```

## 2.3 User Types and Roles

### 2.3.1 User Categories

| User Type | Description | Authentication Method | Access Level |
|-----------|-------------|----------------------|-------------|
| **Platform Administrator** | Manages platform configuration, cluster registration, user roles, and system health | Keycloak OIDC + MFA | Full system access |
| **Project Owner** | Manages project settings, budgets, queue configuration, and team membership | Keycloak OIDC + MFA | Project-scoped administrative access |
| **Developer** | Submits workloads, accesses workspaces, views workload status within assigned projects | Keycloak OIDC + MFA | Project-scoped operational access |
| **Viewer** | Read-only access to workload status and project dashboards | Keycloak OIDC + MFA | Project-scoped read-only access |
| **Service Account (K8s Agent)** | Automated spoke cluster agent communicating with hub API | mTLS client certificate + OIDC client credentials | Cluster registration, heartbeat, workload reconciliation |
| **Service Account (Spoke Proxy)** | Automated spoke proxy providing workspace tunnel endpoints | mTLS client certificate | WebSocket tunnel establishment |

### 2.3.2 Role-Based Access Control (RBAC)

Authorization is enforced at the Platform API level. Roles map to specific gRPC method permissions:

| Role | Workload Management | Cluster Management | Project Management | Budget Management | Audit Access |
|------|:---:|:---:|:---:|:---:|:---:|
| **Admin** | Full | Full | Full | Full | Full |
| **Project Owner** | Project-scoped | View only | Project-scoped | Project-scoped | Project-scoped |
| **Developer** | Create/View (project) | View only | View only | View only | Own actions |
| **Viewer** | View only (project) | View only | View only | View only | None |

All authorization decisions are recorded in the audit log.

## 2.4 Deployment Model

### 2.4.1 CSP Deployment Model

Aegis Platform supports the following deployment models:

| Model | Description | Authorization Boundary Ownership | FedRAMP Applicability |
|-------|-------------|--------------------------------|----------------------|
| **Model B (Self-Hosted)** | Customer deploys Aegis software in their own cloud environment (AWS GovCloud, commercial AWS, or on-premises) | Customer owns the authorization boundary; Aegis software is a component within the customer's ATO package | Primary model -- this SSP covers the Aegis software components |
| **Model C (Managed)** | Aegis deploys and operates the platform in a customer-designated cloud account (future) | Shared responsibility; Aegis operates within the customer's boundary under a service agreement | Future -- separate SSP addendum required |

**Primary deployment target for FedRAMP authorization:** Model B (Self-Hosted) on AWS GovCloud or AWS commercial regions, leveraging AWS's existing FedRAMP High authorization for IaaS controls.

### 2.4.2 Cloud Service Model

| Attribute | Value |
|-----------|-------|
| **Service Model** | SaaS / PaaS hybrid |
| **SaaS Characteristics** | Web-based UI (Backstage), managed authentication (Keycloak), centralized workload management |
| **PaaS Characteristics** | Container orchestration, developer workspace provisioning, Kubernetes-native extensibility |

## 2.5 System Interconnections

### 2.5.1 Interconnections Table

| External System | Organization | Connection Type | Information Exchanged | Direction | Authorization Status |
|-----------------|-------------|----------------|----------------------|-----------|---------------------|
| AWS GovCloud IaaS | Amazon Web Services | Network (VPC, API) | Compute, storage, network, database services | Bidirectional | FedRAMP High Authorized |
| AWS Elastic Kubernetes Service (EKS) | Amazon Web Services | API (HTTPS) | Kubernetes API calls, cluster management | Bidirectional | FedRAMP High Authorized (as part of AWS) |
| AWS RDS (PostgreSQL) | Amazon Web Services | TCP (5432) | Platform state, workload metadata, audit records | Bidirectional | FedRAMP High Authorized (as part of AWS) |
| AWS Elastic Container Registry (ECR) | Amazon Web Services | HTTPS | Container image pull/push | Bidirectional | FedRAMP High Authorized (as part of AWS) |
| AWS Route53 | Amazon Web Services | HTTPS (API) | DNS record management for `aegist.dev` | Outbound | FedRAMP High Authorized (as part of AWS) |
| AWS CloudWatch | Amazon Web Services | HTTPS (API) | Log aggregation, metrics, alarms | Outbound | FedRAMP High Authorized (as part of AWS) |
| AWS CloudTrail | Amazon Web Services | HTTPS (API) | AWS API audit logging | Outbound | FedRAMP High Authorized (as part of AWS) |
| Customer Identity Provider | Customer Organization | SAML 2.0 / OIDC | Authentication assertions, user attributes | Bidirectional | Customer responsibility |
| Cloudflare DNS | Cloudflare, Inc. | HTTPS (API) | DNS record management for `aegis-platform.tech` | Outbound | SOC 2 Type II |
| GitHub Actions | GitHub, Inc. | HTTPS | CI/CD pipeline execution, code repository | Outbound | SOC 2 Type II, FedRAMP Tailored LI-SaaS |
| Smallstep Certificate Authority | Self-hosted (in-boundary) | gRPC/HTTPS | Certificate signing requests, certificate issuance | Internal | Within authorization boundary |

### 2.5.2 Interconnection Security Agreements (ISA)

| Interconnection | ISA Status | MOU/MOA Required |
|-----------------|-----------|-----------------|
| AWS GovCloud | Covered by AWS FedRAMP authorization and customer agreement | Per customer's AWS agreement |
| Customer IdP | Required per customer deployment | Yes -- federated identity trust agreement |
| GitHub Actions | Covered by GitHub Terms of Service | Review for federal use |
| Cloudflare DNS | Covered by Cloudflare service agreement | Review for federal use |

## 2.6 Applicable Laws, Regulations, and Standards

| Law/Regulation/Standard | Applicability |
|--------------------------|--------------|
| **Federal Information Security Modernization Act (FISMA)** | Governing law for federal information security; requires security authorization for systems processing federal information |
| **FedRAMP (Federal Risk and Authorization Management Program)** | Standardized approach to security assessment, authorization, and continuous monitoring for cloud services |
| **NIST SP 800-53 Rev 5** | Security and privacy controls catalog; Moderate baseline applied |
| **NIST SP 800-37 Rev 2** | Risk Management Framework (RMF) for information systems |
| **NIST SP 800-60 Vol II** | Mapping information types to security categories |
| **FIPS 199** | Standards for security categorization of federal information systems |
| **FIPS 200** | Minimum security requirements for federal information systems |
| **FIPS 140-2/3** | Cryptographic module validation requirements |
| **NIST SP 800-63-3** | Digital identity guidelines (IAL, AAL, FAL) |
| **NIST SP 800-171 Rev 2/3** | Protecting CUI in nonfederal systems (applicable if customer processes CUI) |
| **NIST SP 800-161 Rev 1** | Cybersecurity supply chain risk management |
| **OMB Circular A-130** | Management of federal information resources |
| **Privacy Act of 1974** | Applicable if system maintains records retrievable by PII (minimal for Aegis) |
| **EO 14028** | Improving the Nation's Cybersecurity -- SBOM requirements, zero trust architecture |

---

# Section 3: System Environment and Boundary

## 3.1 Authorization Boundary

### 3.1.1 Boundary Description

The Aegis Platform authorization boundary encompasses all software components developed, maintained, and distributed by Aegis Technologies for the purpose of multi-cluster GPU workload orchestration. The boundary includes the hub control plane services, spoke agent software, supporting infrastructure automation (Terraform, Helm charts), and the configuration artifacts necessary for secure deployment.

**Within the authorization boundary:**

| Category | Components |
|----------|------------|
| **Hub services** | Platform API (Go/gRPC), Proxy (Go/WebSocket), Keycloak (OIDC IdP), PostgreSQL (data store), ingress-nginx (TLS termination and routing) |
| **Spoke services** | K8s Agent (Go/controller-runtime operator), Spoke Proxy (Go/WebSocket/TLS) |
| **PKI infrastructure** | step-ca (certificate authority), cert-manager (certificate lifecycle), step-issuer (ClusterIssuer integration) |
| **Infrastructure-as-Code** | Terraform modules (AWS EKS, VPC, RDS, NLB, Route53), Helm charts (aegis-services, aegis-spoke), deployment scripts |
| **Configuration artifacts** | Keycloak realm configuration, network policies, security group definitions, RBAC configurations |
| **CI/CD pipeline definitions** | GitHub Actions workflows for build, test, and deployment |

**Outside the authorization boundary (inherited or external):**

| Category | Component | Responsible Party |
|----------|-----------|-------------------|
| **IaaS** | AWS GovCloud / AWS Commercial (EKS, EC2, VPC, RDS, S3, EBS, CloudWatch, CloudTrail, KMS) | AWS (FedRAMP High Authorized) |
| **Customer IdP** | Customer's SAML/OIDC identity provider | Customer organization |
| **Customer workloads** | Application code, data, and containers executed on spoke clusters | Customer organization |
| **Container registry** | Customer's ECR or other container registry for workload images | Customer organization |
| **End-user devices** | Browsers, VS Code installations used to access the platform | Customer organization |

### 3.1.2 Authorization Boundary Diagram (Textual)

```
+------------------------------------------------------------------------+
|                        AUTHORIZATION BOUNDARY                          |
|                    (FedRAMP CSO: Aegis Platform)                       |
|                                                                        |
|  +------------------------------------------------------------------+  |
|  |                    AEGIS PLATFORM COMPONENTS                      |  |
|  |                                                                    |  |
|  |   +--------------+   +--------------+   +--------------+          |  |
|  |   |  Platform    |   |  Aegis UI    |   |  Workspace   |          |  |
|  |   |    API       |   |  (Backstage) |   |    Proxy     |          |  |
|  |   +------+-------+   +------+-------+   +------+-------+          |  |
|  |          |                  |                   |                  |  |
|  |   +------+------------------+-------------------+-------+         |  |
|  |   |              Kubernetes Control Plane               |         |  |
|  |   +-------------------------+---------------------------+         |  |
|  |                             |                                     |  |
|  |   +-------------------------+---------------------------+         |  |
|  |   |           Identity Provider (Keycloak)              |         |  |
|  |   +--------------------------+--------------------------+         |  |
|  |                              |                                    |  |
|  |   +--------------------------+--------------------------+         |  |
|  |   |    PKI (step-ca, cert-manager, step-issuer)         |         |  |
|  |   +-----------------------------------------------------+         |  |
|  |                                                                    |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  +------------------------------------------------------------------+  |
|  |                SUPPORTING INFRASTRUCTURE (Inherited)              |  |
|  |   (From FedRAMP-Authorized IaaS -- e.g., AWS GovCloud)           |  |
|  |                                                                    |  |
|  |   * Compute: EKS, EC2       * Network: VPC, NLB, Security Groups |  |
|  |   * Storage: S3, EBS        * Database: RDS (PostgreSQL)         |  |
|  |   * Logging: CloudWatch     * Monitoring: CloudTrail             |  |
|  |   * Key Management: KMS     * DNS: Route53                       |  |
|  |   * Container Registry: ECR                                       |  |
|  |                                                                    |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
+------------------------------------------------------------------------+

                         EXTERNAL CONNECTIONS

    +--------------+         +--------------+         +--------------+
    |   GitHub     |         |    Customer  |         |  Cloudflare  |
    |  (CI/CD)     |<------->|      IdP     |<------->|    DNS       |
    |              |  HTTPS  |  (SAML/OIDC) |  HTTPS  |              |
    +--------------+         +--------------+         +--------------+
```

### 3.1.3 Boundary Component Inventory

| Component | In Boundary | Responsibility Model | Control Implementation |
|-----------|:-----------:|---------------------|----------------------|
| Platform API | Yes | CSP (Aegis) | Aegis implements and maintains |
| Aegis UI (Backstage) | Yes | CSP (Aegis) | Aegis implements and maintains |
| Workspace Proxy | Yes | CSP (Aegis) | Aegis implements and maintains |
| Spoke Proxy | Yes | CSP (Aegis) | Aegis implements and maintains |
| K8s Agent | Yes | CSP (Aegis) | Aegis implements and maintains |
| Keycloak | Yes | CSP (Aegis) | Aegis configures and maintains |
| step-ca PKI | Yes | CSP (Aegis) | Aegis deploys and configures |
| cert-manager | Yes | CSP (Aegis) | Aegis deploys and configures |
| Helm Charts / Terraform | Yes | CSP (Aegis) | Aegis develops and maintains |
| PostgreSQL (local) | Yes | CSP (Aegis) | Aegis configures |
| AWS EKS | Inherited | AWS (FedRAMP Authorized) | Customer configures, AWS operates |
| AWS RDS | Inherited | AWS (FedRAMP Authorized) | Customer configures, AWS operates |
| AWS VPC/Networking | Inherited | AWS (FedRAMP Authorized) | Customer configures, AWS operates |
| GitHub Actions | External | GitHub | GitHub operates; Aegis configures pipelines |
| Customer IdP | External | Customer | Customer operates |
| Customer Workloads | External | Customer | Customer responsibility |

## 3.2 Network Architecture and Data Flow

### 3.2.1 Hub Cluster Network Architecture

The hub cluster is deployed within an AWS VPC with network segmentation enforced through subnets and security groups:

**VPC Configuration:**

| Network Element | Configuration | Purpose |
|-----------------|-------------|---------|
| VPC CIDR | Customer-configurable (default: 10.0.0.0/16) | Network address space |
| Public Subnets | Multi-AZ (2-3 availability zones) | Load balancer endpoints (NLB, ALB) |
| Private Subnets | Multi-AZ (2-3 availability zones) | EKS worker nodes, RDS instances |
| Internet Gateway | Attached to VPC | Outbound internet access for public subnets |
| NAT Gateway | One per AZ | Outbound internet access for private subnets |

**Kubernetes Namespace Isolation:**

| Namespace | Components | Network Policy |
|-----------|------------|---------------|
| `aegis-system` | Platform API, Proxy, ingress-nginx | Default deny ingress; allow from ingress controller and inter-service traffic |
| `keycloak` | Keycloak, Keycloak PostgreSQL | Default deny ingress; allow from aegis-system and ingress controller |
| `aegis-pki` | step-ca, step-issuer | Default deny ingress; allow from cert-manager namespace |
| `cert-manager` | cert-manager controller | Default deny ingress; allow from Kubernetes API server |

Network policies enforce a default-deny ingress posture with explicit allowlists for required inter-namespace communication.

### 3.2.2 Spoke Cluster Connectivity

Spoke clusters communicate with the hub through outbound-only gRPC connections over mTLS:

**Connection Model:**

| Flow | Protocol | Direction | Authentication | Encryption |
|------|----------|-----------|----------------|------------|
| K8s Agent to Platform API | gRPC | Outbound from spoke | mTLS client certificate + OIDC client credentials | TLS 1.2+ |
| Spoke Proxy to workspace pods | TCP | Internal (within spoke cluster) | Service-level routing | In-cluster |
| User to Spoke Proxy | WebSocket | Inbound to spoke (via NLB) | JWT token (short-lived, 5-min max) | TLS 1.2+ |

**AWS NLB Relay (Hybrid/Cloud):**

For cloud spoke deployments, an AWS Network Load Balancer (NLB) provides Layer-4 connectivity:

| NLB Target | Port | Protocol | Purpose |
|------------|------|----------|---------|
| Spoke Proxy | 443 | TLS passthrough | WebSocket tunnel for workspace access |
| Platform API (hub) | 8081 | TLS passthrough | gRPC control plane (hub-side NLB) |

### 3.2.3 External Interfaces

| Interface | Endpoint | Protocol | Purpose | Authentication |
|-----------|----------|----------|---------|----------------|
| User Web Access | `https://<domain>/` | HTTPS (443) | Backstage UI for workload management | Keycloak OIDC + MFA |
| VS Code Extension | `grpcs://<domain>:8081` | gRPC over TLS | Control plane API for workload submission | Keycloak OIDC JWT |
| Workspace Proxy | `wss://<domain>/` | WebSocket over TLS | Remote workspace connectivity | Short-lived JWT ticket |
| Keycloak Admin | `https://<domain>/auth/` | HTTPS (8443) | Identity provider administration | Keycloak admin credentials + MFA |
| Keycloak OIDC | `https://<domain>/auth/realms/aegis` | HTTPS (8443) | OIDC discovery, token exchange, SAML federation | Protocol-level (OIDC/SAML) |
| Customer IdP Federation | Customer-specified | SAML 2.0 / OIDC | Federated authentication with customer identity provider | Encrypted assertions |

## 3.3 Data Flow Descriptions

### 3.3.1 Data Flow 1: User Authentication

```
User --> Aegis UI --> Keycloak --> Customer IdP (SAML/OIDC) --> JWT --> Platform API
```

**Step-by-step:**

1. User accesses Aegis UI (Backstage) via HTTPS
2. UI redirects to Keycloak for OIDC authentication
3. Keycloak redirects to customer IdP for federated authentication (if configured)
4. Customer IdP authenticates user (with MFA if required by customer policy)
5. SAML assertion or OIDC token returned to Keycloak
6. Keycloak issues signed JWT (ID token + access token) with PKCE
7. UI stores session token and authenticates API requests via JWT Bearer

| Data Element | Classification | Protection Mechanism |
|--------------|:-:|------------|
| Username | Low | TLS in transit, encrypted at rest in Keycloak DB |
| Email address | Low | TLS in transit, encrypted at rest in Keycloak DB |
| Password | N/A | Never handled by Aegis -- delegated to external IdP |
| JWT Token | Moderate | TLS in transit, short-lived (configurable, default 5 min), digitally signed (RS256) |
| Session cookie | Moderate | Secure flag, HttpOnly, SameSite=Strict |

### 3.3.2 Data Flow 2: Workload Submission

```
User --> Aegis UI / VS Code --> Platform API --> K8s Agent --> Spoke Cluster --> Workload Pod
```

**Step-by-step:**

1. Authenticated user submits workload via UI or VS Code extension (gRPC)
2. Platform API validates JWT, checks RBAC authorization, verifies project membership
3. Platform API checks budget allocation and policy domain constraints
4. Platform API selects optimal spoke cluster based on flavor availability and health
5. Workload definition stored in PostgreSQL and queued for scheduling
6. K8s Agent on the target spoke cluster polls for new workloads (gRPC heartbeat)
7. K8s Agent creates AegisWorkload CRD, which reconciles into a Kubernetes Job
8. Kueue manages admission and resource fairness for the Job
9. K8s Agent reports workload status back to Platform API

| Data Element | Classification | Protection Mechanism |
|--------------|:-:|------------|
| Workload specification | Low | TLS (gRPC), RBAC authorization |
| Resource requests (CPU/GPU/memory) | Low | TLS (gRPC) |
| Container image reference | Low | TLS (gRPC) |
| Workload execution logs | Customer-controlled | Within customer boundary; not stored by Aegis |
| Scheduling metadata | Low | TLS, stored in PostgreSQL with encryption at rest |

### 3.3.3 Data Flow 3: Workspace Access

```
User --> VS Code --> Spoke Proxy (WebSocket/TLS) --> Workspace Pod (SSH/REH port 2222)
```

**Step-by-step:**

1. User requests workspace connection via VS Code extension
2. Platform API mints a short-lived connection ticket (JWT, 5-min expiry)
3. VS Code connects to Spoke Proxy via WebSocket over TLS
4. Spoke Proxy validates connection ticket
5. Spoke Proxy establishes TCP tunnel to workspace pod's SSH/REH port
6. Binary data relayed bidirectionally through the WebSocket tunnel

| Data Element | Classification | Protection Mechanism |
|--------------|:-:|------------|
| Connection ticket | Moderate | Short-lived (5 min), single-use optional, digitally signed |
| WebSocket traffic | Customer-controlled | TLS 1.2+ encryption, content is customer workspace I/O |
| Workspace data | Customer-controlled | Within customer boundary |

### 3.3.4 Data Flow 4: Administrative Operations

```
Admin --> Platform API --> Configuration Change --> Audit Log
```

| Data Element | Classification | Protection Mechanism |
|--------------|:-:|------------|
| Configuration changes | Moderate | TLS, RBAC, audit logging, GitOps version control |
| Admin credentials | Moderate | MFA required, short session timeout |
| Audit logs | Moderate | Append-only, structured JSON, integrity-protected storage |

### 3.3.5 Data Flow 5: Hub-Spoke Communication

```
K8s Agent --> gRPC (mTLS) --> Platform API --> PostgreSQL
```

| Data Element | Classification | Protection Mechanism |
|--------------|:-:|------------|
| Cluster registration | Moderate | mTLS + OIDC client credentials |
| Heartbeat (cluster health, capacity) | Low | mTLS, periodic (configurable interval) |
| Workload status updates | Low | mTLS |
| Certificate material (mTLS certs) | High | step-ca issued, short-lived (90 days), auto-renewed |

## 3.4 Ports, Protocols, and Services (PPS)

### 3.4.1 Hub Cluster PPS

| Source | Destination | Port | Protocol | Service | Direction | Boundary Crossing |
|--------|-------------|:----:|----------|---------|-----------|-------------------|
| Internet (Users) | ingress-nginx | 443 | HTTPS/TLS 1.2+ | UI access, API gateway | Inbound | External to boundary |
| ingress-nginx | Platform API | 8080 | HTTP/REST | Internal routing (REST gateway) | Internal | None |
| ingress-nginx | Platform API | 8081 | gRPC/HTTP2 | Internal routing (gRPC) | Internal | None |
| ingress-nginx | Keycloak | 8443 | HTTPS | Internal routing (OIDC) | Internal | None |
| ingress-nginx | Hub Proxy | 8085 | WebSocket | Internal routing (workspace proxy) | Internal | None |
| Platform API | PostgreSQL | 5432 | TCP/TLS | Database queries | Internal | None (local) / VPC-internal (RDS) |
| Platform API | Keycloak | 8443 | HTTPS | OIDC token validation | Internal | None |
| Spoke K8s Agent | Platform API | 8081 | gRPC/mTLS | Cluster registration, heartbeat | Inbound | Spoke to hub |
| cert-manager | step-ca | 443 | HTTPS | Certificate signing | Internal | Cross-namespace |
| All pods | AWS endpoints | 443 | HTTPS | ECR pull, CloudWatch, KMS | Outbound | Boundary to AWS |

### 3.4.2 Spoke Cluster PPS

| Source | Destination | Port | Protocol | Service | Direction | Boundary Crossing |
|--------|-------------|:----:|----------|---------|-----------|-------------------|
| K8s Agent | Hub Platform API | 8081 | gRPC/mTLS | Registration, heartbeat, workload sync | Outbound | Spoke to hub |
| K8s Agent | Hub Keycloak | 8443 | HTTPS | OIDC client-credentials token exchange | Outbound | Spoke to hub |
| Internet (Users) | Spoke Proxy (via NLB) | 443 | WebSocket/TLS | Workspace tunnel access | Inbound | External to spoke |
| Spoke Proxy | Workspace Pod | 2222 | TCP | SSH/Remote Extension Host | Internal | None |
| All pods | AWS endpoints | 443 | HTTPS | ECR pull, CloudWatch | Outbound | Boundary to AWS |

### 3.4.3 Denied by Default

All ports, protocols, and services not explicitly listed above are denied by default through:

- AWS Security Groups (stateful firewall at the instance/ENI level)
- Kubernetes NetworkPolicy (default deny ingress applied to all namespaces)
- NLB/ALB listener rules (only configured ports accepted)

## 3.5 System Interconnections Detail

### 3.5.1 AWS Services (Inherited Controls)

Aegis Platform leverages the following AWS services, inheriting their FedRAMP-authorized security controls:

| AWS Service | FedRAMP Status | Controls Inherited | Aegis Responsibility |
|-------------|---------------|-------------------|---------------------|
| **EKS** | FedRAMP High (as part of AWS) | PE-*, PS-*, MP-*, some CM-*, SC-* | Kubernetes cluster configuration, RBAC, pod security |
| **RDS (PostgreSQL)** | FedRAMP High (as part of AWS) | PE-*, SC-28 (encryption at rest), CP-9 (backups) | Database configuration, access credentials, schema management |
| **VPC** | FedRAMP High (as part of AWS) | SC-7 (boundary protection) | Subnet design, security group rules, NACLs |
| **ECR** | FedRAMP High (as part of AWS) | CM-2 (image baseline) | Image scanning, access policies |
| **Route53** | FedRAMP High (as part of AWS) | SC-20/SC-21 (secure DNS) | DNS record management |
| **CloudWatch** | FedRAMP High (as part of AWS) | AU-6 (log review), SI-4 (monitoring) | Log group configuration, alarm definitions |
| **CloudTrail** | FedRAMP High (as part of AWS) | AU-2/AU-3 (audit events) | Trail configuration, log integrity validation |
| **KMS** | FedRAMP High (as part of AWS) | SC-12/SC-13 (cryptographic key management) | Key policy configuration, key rotation schedule |
| **NLB** | FedRAMP High (as part of AWS) | SC-7 (boundary protection) | Target group configuration, health checks |

### 3.5.2 External Identity Provider Federation

| Attribute | Value |
|-----------|-------|
| **Federation Protocol** | SAML 2.0 or OpenID Connect |
| **Federation Endpoint** | Keycloak identity brokering |
| **Assertion Encryption** | Required (signed + encrypted SAML assertions; signed JWTs for OIDC) |
| **Attribute Mapping** | Username, email, group membership |
| **Session Management** | Keycloak enforces session timeout, concurrent session limits |
| **MFA Requirement** | Configurable per Keycloak authentication flow; recommended AAL-2 |
| **Responsibility** | Customer IdP configuration and user lifecycle management is customer responsibility |

### 3.5.3 DNS Providers

| Provider | Domain | Purpose | Protocol |
|----------|--------|---------|----------|
| AWS Route53 | `aegist.dev` | Primary DNS for hub and spoke endpoints | HTTPS (API), DNS (UDP/TCP 53) |
| Cloudflare | `aegis-platform.tech` | Documentation site, secondary DNS | HTTPS (API), DNS (UDP/TCP 53) |

## 3.6 Leveraged Authorizations

Aegis Platform leverages the following existing FedRAMP authorizations:

| CSP | Service | FedRAMP ID | Impact Level | Controls Inherited |
|-----|---------|-----------|:------------:|-------------------|
| Amazon Web Services | AWS GovCloud (US) | F1603047866 | High | Physical security, environmental controls, media protection, personnel security, infrastructure maintenance |
| Amazon Web Services | AWS US East/West | F1603047866 | High | Same as above (for commercial region deployments) |

**Customer Responsibility:** When deploying Aegis Platform in Model B (self-hosted), the customer is responsible for:

1. Establishing their own authorization boundary that includes Aegis components
2. Configuring AWS services (EKS, RDS, VPC) per their organization's security requirements
3. Managing their AWS FedRAMP-authorized service agreement
4. Implementing any additional controls required by their Authorizing Official

## 3.7 Cryptographic Controls Summary

### 3.7.1 Data in Transit

| Communication Path | Protocol | Minimum Version | Cipher Suites | Certificate Source |
|--------------------|----------|:-:|---------------|-------------------|
| User to ingress-nginx | TLS | 1.2 | AES-128/256-GCM, CHACHA20-POLY1305 | cert-manager (step-ca issued or Let's Encrypt) |
| Platform API to PostgreSQL | TLS | 1.2 | AES-256-GCM | RDS-managed (AWS) |
| Hub to Spoke (gRPC) | mTLS | 1.2 | AES-256-GCM | step-ca issued (90-day lifetime, auto-renewed) |
| Keycloak OIDC | TLS | 1.2 | AES-256-GCM | cert-manager issued |
| Spoke Proxy to User | TLS | 1.2 | AES-256-GCM | Spoke-proxy self-signed CA (auto-generated by Helm hook) |

### 3.7.2 Data at Rest

| Data Store | Encryption Method | Key Management | FIPS Status |
|------------|------------------|----------------|:----------:|
| RDS (PostgreSQL) | AES-256 | AWS KMS (customer-managed key) | FIPS 140-2 validated (AWS KMS) |
| EBS Volumes | AES-256 | AWS KMS | FIPS 140-2 validated (AWS KMS) |
| S3 (if used for backups) | AES-256 (SSE-KMS) | AWS KMS | FIPS 140-2 validated (AWS KMS) |
| Keycloak database | AES-256 | Application-level + storage encryption | Planned (BoringCrypto) |
| etcd (Kubernetes secrets) | AES-CBC or AES-GCM | EKS-managed envelope encryption | FIPS 140-2 validated (AWS KMS) |

### 3.7.3 FIPS 140 Compliance Roadmap

| Component | Current Status | Target | Module |
|-----------|---------------|--------|--------|
| Go services (Platform API, Proxy, K8s Agent) | Standard Go crypto | FIPS-compliant | Go BoringCrypto |
| Keycloak | Standard Java crypto | FIPS-compliant | BouncyCastle FIPS |
| AWS infrastructure encryption | FIPS 140-2 validated | Current | AWS KMS |

Full cryptographic inventory is documented in **Appendix K (FIPS 140 Cryptographic Modules)** and **Appendix Q (Cryptographic Inventory)**.

## 3.8 Physical Environment

Aegis Platform is a cloud-native application with no dedicated physical infrastructure. All physical security controls are inherited from the underlying IaaS provider (AWS):

| Physical Control Family | Responsibility | Implementation |
|------------------------|---------------|----------------|
| PE-1 through PE-20 | Inherited (AWS) | AWS data center physical security per FedRAMP High authorization |
| Environmental controls | Inherited (AWS) | AWS HVAC, fire suppression, power management |
| Media protection | Inherited (AWS) + CSP | AWS physical media; Aegis ensures no sensitive data written to removable media |

## 3.9 Logical Environment

### 3.9.1 Operating System and Runtime

| Layer | Technology | Version | Hardening |
|-------|-----------|---------|-----------|
| Container OS | Distroless / Alpine Linux | Current | Minimal attack surface, no shell in production images |
| Container Runtime | containerd | EKS-managed | CIS Benchmark for Kubernetes |
| Kubernetes | EKS-managed | 1.28+ | CIS Kubernetes Benchmark, Pod Security Standards |
| Go Runtime | Go 1.24+ | Current | Compiled binaries, no runtime interpreter |
| Java Runtime (Keycloak) | OpenJDK 17 | Current | UBI9-based image with security updates |

### 3.9.2 Development and Staging Environments

| Environment | Purpose | Network Isolation | Data Classification |
|-------------|---------|------------------|-------------------|
| **Local Dev** | Developer testing on Docker Desktop | Isolated (localhost only) | No federal data |
| **Preview** | Ephemeral CI/CD environments per pull request | Isolated VPC per deployment | No federal data |
| **Staging** | Pre-production validation | Separate VPC, no production data access | Synthetic test data only |
| **Production** | Federal customer deployments | Customer-controlled VPC | Per customer classification |

## 3.10 Cross-References to SSP Appendices

The following SSP appendices provide additional detail referenced throughout Sections 1-3:

| Appendix | Title | Section Reference |
|----------|-------|-------------------|
| **Appendix A** | Control Implementation Statements | Sections 1.2, 2.3, 3.1 |
| **Appendix E** | Digital Identity Worksheet (IAL/AAL/FAL) | Section 1.7 |
| **Appendix F** | Rules of Behavior | Section 2.3 |
| **Appendix G** | Information System Contingency Plan (ISCP) | Section 3.1 |
| **Appendix H** | Configuration Management Plan | Section 3.2 |
| **Appendix I** | Incident Response Plan | Section 3.1 |
| **Appendix J** | Control Implementation Summary (CIS) Workbook | Section 1.2 |
| **Appendix K** | FIPS 140 Cryptographic Modules | Section 3.7 |
| **Appendix L** | Separation of Duties Matrix | Sections 1.5, 1.6 |
| **Appendix M** | Integrated Inventory Workbook | Section 3.1 |
| **Appendix N** | Continuous Monitoring Plan | Sections 1.2, 3.1 |
| **Appendix O** | Plan of Action and Milestones (POA&M) | Sections 1.2, 3.7 |
| **Appendix P** | Supply Chain Risk Management Plan (SCRMP) | Section 3.1 |
| **Appendix Q** | Cryptographic Inventory | Section 3.7 |

---

## Document Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| System Owner / CEO | Carlos Sanchez | _____________ | ______ |
| ISSO / CISO | Carlos Sanchez | _____________ | ______ |
| Authorizing Official | TBD (Agency) | _____________ | ______ |

---

## Glossary

| Term | Definition |
|------|------------|
| AO | Authorizing Official -- federal official who accepts risk and grants authorization to operate |
| ATO | Authority to Operate -- formal authorization to operate a federal information system |
| CSO | Cloud Service Offering -- the cloud service subject to FedRAMP authorization |
| CSP | Cloud Service Provider -- organization providing the cloud service |
| CUI | Controlled Unclassified Information -- information requiring safeguarding per 32 CFR 2002 |
| FedRAMP | Federal Risk and Authorization Management Program |
| FIPS | Federal Information Processing Standards |
| Hub | Central Aegis control plane cluster running Platform API, Keycloak, and supporting services |
| IaaS | Infrastructure as a Service |
| IAL | Identity Assurance Level (per NIST SP 800-63) |
| AAL | Authenticator Assurance Level (per NIST SP 800-63) |
| FAL | Federation Assurance Level (per NIST SP 800-63) |
| IdP | Identity Provider -- system that creates and manages user identities |
| ISSO | Information System Security Officer |
| JWT | JSON Web Token -- compact token format for transmitting claims between parties |
| mTLS | Mutual Transport Layer Security -- TLS with client and server certificate authentication |
| NIST | National Institute of Standards and Technology |
| OIDC | OpenID Connect -- identity layer on top of OAuth 2.0 |
| PII | Personally Identifiable Information |
| PKI | Public Key Infrastructure |
| POA&M | Plan of Action and Milestones |
| RBAC | Role-Based Access Control |
| RDS | Amazon Relational Database Service |
| RMF | Risk Management Framework (NIST SP 800-37) |
| SAML | Security Assertion Markup Language |
| SBOM | Software Bill of Materials |
| Spoke | Remote Aegis cluster running K8s Agent and Spoke Proxy for workload execution |
| SSP | System Security Plan |
| 3PAO | Third Party Assessment Organization |
| VPC | Virtual Private Cloud |
