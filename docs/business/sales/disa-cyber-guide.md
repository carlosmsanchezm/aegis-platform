# DISA, Cybersecurity Programs, and DoD Market Entry Guide for Aegis Platform

**Last Updated:** 2026-03-08
**Document Owner:** Carlos Sanchez, Founder/CEO
**Classification:** Internal -- Business Strategy
**Status:** Research document -- not legal advice

---

## Executive Summary

This guide covers the U.S. Department of Defense (DoD) and federal cybersecurity ecosystem relevant to Aegis Platform -- a self-hosted Kubernetes-based GPU workload scheduling platform with audit logging, mTLS, policy enforcement, and FIPS-capable cryptography. It addresses DISA programs, CMMC certification, FedRAMP applicability for self-hosted software, the ATO process, CDAO/DoD CIO initiatives, and cyber-focused funding opportunities.

**Key finding for Aegis:** As a self-hosted product deployed inside a customer's authorization boundary, Aegis does NOT need its own FedRAMP authorization or ATO. Instead, Aegis needs to produce vendor security documentation packages that help customers include Aegis in their own ATO. The most impactful near-term actions are (1) achieving CMMC Level 2 if pursuing CUI-handling contracts directly, (2) building an OSCAL-formatted security documentation package, and (3) engaging DISA's DevSecOps and container security programs.

---

## Table of Contents

1. [DISA Programs Relevant to DevSecOps/Container Platforms](#1-disa-programs)
2. [CMMC 2.0 -- Cybersecurity Maturity Model Certification](#2-cmmc)
3. [FedRAMP vs Self-Hosted Clarification](#3-fedramp-self-hosted)
4. [CDAO -- Chief Digital and AI Office](#4-cdao)
5. [DoD CIO Programs](#5-dod-cio)
6. [ATO Process and Vendor Documentation](#6-ato-process)
7. [Cyber-Specific Funding Opportunities](#7-cyber-funding)
8. [Recommended Action Plan for Aegis](#8-action-plan)

---

## 1. DISA Programs Relevant to DevSecOps/Container Platforms

### 1.1 What DISA Does

The Defense Information Systems Agency (DISA) is a DoD combat support agency that provides IT infrastructure and cybersecurity services. For software vendors, DISA matters because it:

- Publishes **Security Technical Implementation Guides (STIGs)** that define hardening requirements
- Manages the **DoD Information Network (DoDIN)** approval process
- Operates the **DoD Cloud Computing Security Requirements Guide (CC SRG)**
- Evaluates technologies through the **Security Readiness Review (SRR)** process
- Maintains security baselines that DoD components reference in procurement

### 1.2 DISA STIGs for Containers and Kubernetes

DISA publishes STIGs directly relevant to Aegis's technology stack. These are the hardening standards that DoD customers will be required to comply with.

| STIG | ID | Relevance to Aegis |
|------|----|--------------------|
| **Kubernetes STIG** | V-242381 through V-242460+ | Covers API server hardening, RBAC, network policies, etcd encryption, audit logging. Directly applicable to Aegis's hub and spoke clusters. |
| **Container Platform SRG** | SRG-APP-000001 through SRG-APP-000999+ | Security Requirements Guide for container orchestration platforms. Aegis would be evaluated against this. |
| **Docker Enterprise STIG** | V-235789+ | Container runtime hardening. Applicable to container image and runtime configuration. |
| **Application Security and Development STIG** | V-222387+ | General application security requirements. Applicable to platform-api, proxy, k8s-agent. |
| **Web Server STIG** (various) | Varies | Applicable to proxy service and any HTTP endpoints. |
| **PostgreSQL STIG** | V-214050+ | Applicable to Keycloak's backing database. |

**Where to find STIGs:** https://public.cyber.mil/stigs/downloads/

**Action for Aegis:** Create a STIG compliance matrix showing how Aegis addresses each relevant STIG requirement. This is a high-value sales document for DoD customers. Many requirements Aegis already meets (mTLS, RBAC, audit logging, NetworkPolicy templates). Document gaps and provide hardening guidance.

### 1.3 DISA's Approved Products List (APL)

The DISA APL (Approved Products List) historically covered hardware and unified communications products. It is **not directly relevant** to software platforms like Aegis.

However, DISA does maintain several relevant lists:

| List | Relevance | Action Required |
|------|-----------|-----------------|
| **APL** (Unified Capabilities) | Not relevant -- covers UC/telecom products | None |
| **DoD UC APL** | Not relevant | None |
| **Iron Bank** (Platform One) | Highly relevant -- DoD hardened container image repository | Submit Aegis container images for Iron Bank inclusion |
| **Software Forge** (Platform One) | Relevant -- DoD software catalog | Register Aegis as an available platform |
| **cATO Recipient List** | Relevant -- tracks software with continuous ATO | Eventual goal |

### 1.4 DISA Iron Bank and Platform One

**Iron Bank** is the DoD's repository of hardened, signed container images maintained under Platform One. This is the most relevant DISA program for Aegis.

**What Iron Bank Provides:**
- DoD-hardened base images (UBI, Alpine alternatives)
- Image scanning and signing pipeline
- Vulnerability tracking and remediation timelines
- Trusted image provenance for DoD deployments

**What Aegis Should Do:**
1. Submit Aegis container images (platform-api, proxy, k8s-agent, workspace-vscode) to Iron Bank
2. Base images on Iron Bank-approved base images (e.g., `registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal`)
3. Provide Dockerfiles, SBOMs, and hardening documentation per Iron Bank submission requirements
4. Maintain images with timely vulnerability patches

**Iron Bank Submission Process:**
1. Create an account at https://login.dso.mil
2. Submit container hardening request via https://repo1.dso.mil
3. Provide Dockerfile, justification for each package, and security scan results
4. Iron Bank team reviews and hardens the image
5. Image is published to `registry1.dso.mil`

**Platform One** is the DoD enterprise DevSecOps platform built on Big Bang (a Helm chart collection for Kubernetes security tooling). Aegis could potentially integrate with or deploy alongside Big Bang components.

### 1.5 How DISA Evaluates New Technology

DISA evaluates new technology through several mechanisms:

| Mechanism | Process | Timeline |
|-----------|---------|----------|
| **Security Readiness Review (SRR)** | DISA reviews architecture and security posture | 3-6 months |
| **Provisional Authorization (PA)** | For cloud services; DISA grants PA for DoD-wide use | 6-12 months |
| **Connection Approval** | Required to connect to DoDIN | Varies |
| **STIG Validation** | Vendors can participate in STIG creation for their products | 6-12 months |

For Aegis as self-hosted software, the most relevant path is supporting customers through their own authorization process rather than seeking a DISA-level authorization. DISA PA (Provisional Authorization) is primarily for cloud service providers, not self-hosted software.

---

## 2. CMMC 2.0 -- Cybersecurity Maturity Model Certification

### 2.1 What Is CMMC 2.0

The Cybersecurity Maturity Model Certification (CMMC) is a DoD program that verifies defense contractors have adequate cybersecurity practices to protect Federal Contract Information (FCI) and Controlled Unclassified Information (CUI).

CMMC 2.0 simplified the original 5-level model down to 3 levels:

| Level | Name | Controls | Assessment | Who Needs It |
|-------|------|----------|------------|--------------|
| **Level 1** | Foundational | 17 practices (FAR 52.204-21) | Annual self-assessment | Any company handling FCI |
| **Level 2** | Advanced | 110 practices (NIST SP 800-171 Rev 2) | Third-party assessment (C3PAO) or self-assessment depending on criticality | Companies handling CUI |
| **Level 3** | Expert | 110+ practices (NIST SP 800-172) | Government-led assessment (DIBCAC) | Highest-priority CUI programs |

### 2.2 What Level Does a Software Vendor Need?

This depends on what data flows through your systems during the contract:

| Scenario | CMMC Level Needed | Rationale |
|----------|-------------------|-----------|
| Selling COTS software, no CUI access | Level 1 (possibly none) | Only handling FCI (contract info) |
| Software development with CUI requirements/specs | Level 2 | Handling CUI (design docs, specs) |
| Software processing CUI at runtime | Level 2 | CUI flows through your product |
| Critical weapons system software | Level 3 | Highest sensitivity programs |

**For Aegis specifically:**
- If selling Aegis as COTS (commercial off-the-shelf) software where the customer deploys and operates it, and Aegis (the company) never touches CUI: **Level 1 may suffice**, or CMMC may not even apply if Aegis is sold through a prime contractor who handles all CUI.
- If Aegis (the company) receives CUI during development, integration, or support (e.g., customer shares classified architecture docs): **Level 2 required**.
- **Recommendation:** Plan for Level 2 to maximize addressable market. Many DoD RFPs increasingly require CMMC Level 2 for all participants in the supply chain.

### 2.3 How to Get Certified

**CMMC Level 1 (Self-Assessment):**
1. Implement 17 FAR 52.204-21 practices
2. Complete annual self-assessment
3. Submit score to SPRS (Supplier Performance Risk System)
4. Affirm compliance annually via senior official
5. **Cost:** Minimal -- primarily time and basic cybersecurity hygiene
6. **Timeline:** 1-3 months for a small company with good practices

**CMMC Level 2 (Third-Party Assessment):**
1. Implement all 110 NIST SP 800-171 Rev 2 practices
2. Develop a System Security Plan (SSP) for your CUI environment
3. Create a Plan of Action and Milestones (POA&M) for any gaps
4. Engage a CMMC Third-Party Assessment Organization (C3PAO)
5. Undergo assessment (typically 3-5 days onsite/virtual)
6. Receive certification (valid for 3 years)
7. Submit score to SPRS

**List of accredited C3PAOs:** https://cyberab.org/catalog

### 2.4 Timeline and Cost for a 1-Person Company

| Phase | Duration | Cost Estimate | Notes |
|-------|----------|---------------|-------|
| **Gap assessment** | 2-4 weeks | $0 (self) or $5K-$15K (consultant) | Compare current state to NIST 800-171 |
| **Remediation** | 2-6 months | $5K-$30K | Tooling, encryption, MFA, policies |
| **Documentation** | 1-2 months | $0 (self) or $5K-$10K (consultant) | SSP, POA&M, policies |
| **C3PAO assessment** | 3-5 days | $25K-$50K | Third-party assessment fee |
| **Total (Level 2)** | 4-10 months | $30K-$100K | |
| **Total (Level 1)** | 1-3 months | $0-$5K | Self-assessment only |

**For a 1-person company, realistic recommendations:**
- Start with Level 1 self-assessment immediately (days, not weeks)
- Begin Level 2 preparation using existing SOC 2 + ISO 27001 controls (significant overlap with NIST 800-171)
- Your existing compliance program already covers roughly 60-70% of NIST 800-171 requirements
- Budget $30K-$50K for the C3PAO assessment itself
- Consider using a CMMC-focused consultant ($5K-$15K) to avoid costly assessment failures

**Cost-saving approach:** Much of NIST 800-171 overlaps with what Aegis already has:
- Access control (SOC 2 CC6.x) maps to NIST 800-171 3.1.x
- Audit (SOC 2 CC7.x) maps to 3.3.x
- Configuration management (SOC 2 CC8.x) maps to 3.4.x
- Risk assessment (ISO 27001 Clause 6.1) maps to 3.11.x

### 2.5 Is CMMC Required Before Selling to DoD?

**As of March 2026:**

- The CMMC final rule (32 CFR Part 170) was published December 16, 2024, effective December 16, 2024
- The DFARS acquisition rule (48 CFR) implementing CMMC in contracts is being phased in
- **Phase 1 (starting mid-2025):** CMMC Level 1 self-assessments required in new contracts; Level 2 self-assessments where applicable
- **Phase 2 (starting mid-2026):** CMMC Level 2 C3PAO assessments required in applicable contracts
- **Phase 3 (starting mid-2027):** Full implementation including Level 3
- **Phase 4 (starting mid-2028):** CMMC required in all applicable DoD contracts, including option exercises

**Practical reality:**
- You can sell to DoD today without CMMC certification if the specific contract does not yet include the CMMC DFARS clause
- However, having CMMC certification (or being in process) is increasingly a competitive differentiator
- Many primes are already requiring subcontractors to demonstrate CMMC readiness
- **Recommendation:** Begin Level 1 self-assessment now. Plan Level 2 for when a specific contract requires it or to gain competitive advantage.

---

## 3. FedRAMP vs Self-Hosted Clarification

### 3.1 Self-Hosted Software Does NOT Need FedRAMP

**Confirmed:** FedRAMP (Federal Risk and Authorization Management Program) authorizes **cloud service offerings (CSOs)** -- meaning services operated by the vendor or a third party on behalf of the customer. Self-hosted software deployed and operated within a customer's own authorization boundary does **not** require its own FedRAMP authorization.

| Deployment Model | FedRAMP Required? | Why |
|------------------|-------------------|-----|
| SaaS (vendor-operated in vendor's cloud) | Yes | Vendor controls the environment |
| PaaS/IaaS (vendor-operated) | Yes | Vendor controls the environment |
| Self-hosted in customer's data center | No | Customer controls the environment |
| Self-hosted in customer's cloud (incl. GovCloud) | No | Customer controls the environment |
| Self-hosted in customer's FedRAMP-authorized IaaS | No, but the IaaS does | Customer controls the application |

**This is critical for Aegis (Model B -- self-hosted):** Aegis does not need FedRAMP authorization. The customer who deploys Aegis in their environment is responsible for obtaining their ATO, and Aegis falls within their authorization boundary.

### 3.2 What Documentation IS Needed for Self-Hosted DoD Software

Even though FedRAMP is not required, DoD customers deploying self-hosted software need substantial documentation from the vendor to support their ATO. What Aegis should provide:

| Document | Purpose | Aegis Status |
|----------|---------|--------------|
| **Vendor Security Architecture Guide** | Describes system components, data flows, trust boundaries | Exists (customer-docs/security-architecture-guide.md) |
| **Customer Responsibility Matrix (CRM)** | Delineates vendor vs customer security responsibilities | Exists (customer-docs/customer-responsibility-matrix.md) |
| **Control Implementation Statements** | How the product implements specific NIST 800-53 controls | Exists (customer-docs/control-implementation-statements.md) |
| **Hardening Guide / STIG Compliance** | Step-by-step deployment hardening instructions | Partially exists (configuration-hardening-guide.md) |
| **SBOM (Software Bill of Materials)** | Lists all software components for vulnerability tracking | Structure exists (compliance/sbom/) |
| **Vulnerability Disclosure Policy** | How vendor handles reported vulnerabilities | Needed |
| **Patch/Update Cadence** | How often security patches are released | Needed |
| **Incident Response Coordination Plan** | How vendor supports customer during security incidents | Exists (customer-docs/incident-response-runbook.md) |
| **FIPS 140-2/3 Cryptographic Inventory** | Documents all crypto modules and their FIPS status | Needed |
| **Supply Chain Risk Management (SCRM)** | Software supply chain security practices | Needed |
| **OSCAL-formatted SSP components** | Machine-readable control documentation | Partial (compliance/oscal/) |

### 3.3 RMF (Risk Management Framework) Requirements

The DoD Risk Management Framework (RMF), defined in DoDI 8510.01, is the process by which DoD systems receive an ATO. Aegis does not go through RMF itself -- the **customer's system** that includes Aegis goes through RMF.

However, Aegis needs to support the customer's RMF process by providing:

| RMF Step | What Customer Needs from Aegis |
|----------|-------------------------------|
| Step 1: Categorize | Data types processed, impact levels |
| Step 2: Select Controls | Which NIST 800-53 controls Aegis satisfies |
| Step 3: Implement | How controls are implemented (technical details) |
| Step 4: Assess | Evidence of control effectiveness (test results, scan reports) |
| Step 5: Authorize | Vendor risk acceptance statements |
| Step 6: Monitor | Ongoing vulnerability disclosures, patch availability |

### 3.4 When Would Aegis Need FedRAMP?

FedRAMP would become relevant only if Aegis pivots to Model C (managed service) where Aegis operates the platform on behalf of the customer. Current Model B does not require it.

If pursuing FedRAMP in the future, see the existing analysis at `docs/compliance/fedramp/STATUS.md`.

---

## 4. CDAO -- Chief Digital and AI Office

### 4.1 What Is CDAO

The Chief Digital and AI Office (CDAO) was established in June 2022 to accelerate the DoD's adoption of data, analytics, and artificial intelligence. It reports directly to the Deputy Secretary of Defense and consolidated several predecessor organizations:

- Joint AI Center (JAIC)
- Defense Digital Service (DDS)
- Advana (Advanced Analytics)
- Chief Data Officer functions

**Mission:** Accelerate DoD's adoption of data, analytics, and AI to generate decision advantage from the boardroom to the battlefield.

### 4.2 Current CDAO Initiatives Related to AI Infrastructure (2025-2026)

| Initiative | Description | Relevance to Aegis |
|------------|-------------|---------------------|
| **Task Force Lima** | DoD generative AI task force evaluating LLM applications | High -- needs GPU infrastructure for inference/training |
| **CDAO AI/ML Infrastructure** | Building shared compute infrastructure for DoD AI workloads | Direct -- Aegis is a GPU workload scheduling platform |
| **Responsible AI (RAI)** | Implementing DoD AI ethical principles | Moderate -- Aegis can support RAI through audit trails |
| **Data Mesh / Data Fabric** | Enterprise data architecture modernization | Low -- tangential |
| **JADC2 (Joint All-Domain C2)** | AI-enabled command and control | Moderate -- if GPU workloads support JADC2 models |
| **Tradewinds Solutions Marketplace** | Digital marketplace for approved AI/data tools | High -- pathway to list Aegis |

### 4.3 How to Engage with CDAO Programs

**Primary engagement channels:**

1. **Tradewinds Solutions Marketplace** (https://tradewindai.com)
   - CDAO's procurement marketplace for AI/data solutions
   - Open to commercial companies
   - Submit solutions for evaluation
   - Can lead to OTA (Other Transaction Authority) awards
   - **Action:** Register and submit Aegis as a solution

2. **CDAO Broad Agency Announcements (BAAs)**
   - Published on SAM.gov (search for CDAO or "Chief Digital and AI Office")
   - Topics typically cover AI infrastructure, data analytics, responsible AI
   - White paper submissions, not full proposals initially
   - CDAO maintains a standing BAA (similar to DARPA's) for innovative solutions

3. **Defense Innovation Unit (DIU)**
   - Works closely with CDAO
   - Commercial Solutions Opening (CSO) process
   - Focuses on commercial technology with military applications
   - https://www.diu.mil

4. **NSIN (National Security Innovation Network)**
   - Connects startups with DoD problems
   - X-Force fellowship, Hacking for Defense
   - Good entry point for first-time DoD vendors

### 4.4 CDAO Procurement Vehicles

| Vehicle | Type | How to Access | Best For |
|---------|------|---------------|----------|
| **Tradewinds OTA** | Other Transaction | Submit via tradewindai.com | Prototype agreements, less than $5M |
| **CDAO BAA** | BAA | White paper to SAM.gov posting | R&D partnerships |
| **DIU CSO** | Commercial Solutions Opening | Apply via diu.mil | Proven commercial tech |
| **SEWP V** | GSA contract vehicle | Partner with SEWP holder | Product sales |
| **GSA MAS** | Multiple Award Schedule | Obtain GSA Schedule | Recurring sales |

---

## 5. DoD CIO Programs

### 5.1 Cloud Computing Security Requirements Guide (CC SRG)

The DoD Cloud Computing SRG (published by DISA) defines security requirements for cloud service providers hosting DoD workloads. It establishes Impact Levels:

| Impact Level | Data Types | FedRAMP Baseline | Relevance to Aegis |
|--------------|-----------|------------------|---------------------|
| **IL2** | Public, non-CUI | FedRAMP Moderate | Not directly relevant (self-hosted) |
| **IL4** | CUI, FOUO | FedRAMP Moderate+ | Relevant if customer's cloud is IL4 |
| **IL5** | CUI, mission data, national security | FedRAMP High+ | Relevant for DoD mission workloads |
| **IL6** | Classified (SECRET) | FedRAMP High++ | Relevant for classified GPU workloads |

**For self-hosted Aegis:** The CC SRG applies to the underlying cloud infrastructure (e.g., AWS GovCloud at IL4/5), not directly to Aegis. However, Aegis should be deployable on IL4/5 infrastructure, which means:
- Supporting FIPS 140-2/3 cryptography
- Operating on approved container runtimes
- Meeting STIG requirements
- Supporting CAC/PIV authentication

### 5.2 DoD Enterprise DevSecOps Reference Design

The DoD Enterprise DevSecOps Reference Design (published by DoD CIO) is directly relevant to Aegis. It defines the architecture for DoD software factories.

**Key components of the reference design:**

| Component | Description | Aegis Alignment |
|-----------|-------------|-----------------|
| **Software Factory** | CI/CD pipeline with security gates | Aegis can deploy within a software factory |
| **Container Hardening** | DISA STIG-compliant containers | Aegis images should meet this |
| **Kubernetes Platform** | Hardened K8s with policy enforcement | Aegis IS a Kubernetes platform |
| **Service Mesh** | mTLS, observability, traffic management | Aegis provides mTLS natively |
| **Policy Engine** | OPA/Gatekeeper for admission control | Aegis can integrate with these |
| **Image Signing** | Cosign/Notary for image provenance | Aegis should implement this |
| **SBOM** | CycloneDX/SPDX for supply chain | Aegis has SBOM structure |

**DoD Software Factory ecosystem (relevant programs):**

| Program | Organization | Description |
|---------|-------------|-------------|
| **Platform One** | USAF | DoD enterprise DevSecOps platform (Big Bang) |
| **Party Bus** | USMC | Marine Corps software factory |
| **Black Pearl** | USN | Navy DevSecOps platform |
| **Army Software Factory** | USA | Army Futures Command |
| **Kobayashi Maru** | USSF | Space Force DevSecOps |
| **Kessel Run** | USAF | Air Force software factory |

**Action for Aegis:** Position Aegis as infrastructure that complements these software factories, specifically for GPU/AI workload scheduling -- a capability most software factories lack.

### 5.3 Relevance to Aegis Platform

Aegis's technical capabilities map well to DoD requirements:

| Aegis Capability | DoD Requirement | Standard/Reference |
|------------------|-----------------|-------------------|
| mTLS between services | Zero Trust Architecture | DoD Zero Trust Strategy (2022) |
| Audit logging (structured JSON) | Continuous monitoring | NIST 800-137, CC SRG |
| RBAC/ABAC | Least privilege access | NIST 800-53 AC-6 |
| NetworkPolicy enforcement | Microsegmentation | DoD Zero Trust Pillar: Network |
| FIPS-capable crypto | Cryptographic protection | FIPS 140-2/3, SC-13 |
| Container image scanning | Supply chain security | EO 14028, SSDF |
| Multi-cluster management | Edge/tactical deployments | JADC2, tactical cloud |
| GPU workload scheduling | AI/ML infrastructure | CDAO AI strategy |

---

## 6. ATO Process and Vendor Documentation

### 6.1 What a Software Vendor Needs to Provide

When a DoD customer deploys Aegis and seeks an ATO (Authority to Operate), the Authorizing Official (AO) evaluates the entire system -- including vendor-provided software. The vendor's responsibility is to provide documentation that accelerates the customer's ATO.

**Vendor documentation package (what Aegis should provide):**

| Document | Priority | Description | Aegis Status |
|----------|----------|-------------|--------------|
| **Security Architecture Guide** | Critical | System diagram, data flows, ports/protocols, trust boundaries | Exists |
| **Customer Responsibility Matrix** | Critical | Which controls vendor vs customer implements | Exists |
| **Control Implementation Statements** | Critical | How the product addresses each NIST 800-53 control | Exists |
| **STIG Compliance Report** | High | Checklist showing STIG compliance status | Not started |
| **Configuration Hardening Guide** | High | Step-by-step hardening instructions | Partially exists |
| **SBOM** | High | Software Bill of Materials (CycloneDX or SPDX) | Structure exists |
| **Vulnerability Scan Results** | High | Current scan results (container, dependency, SAST) | Not packaged |
| **FIPS Cryptographic Inventory** | High | All crypto modules, algorithms, FIPS validation status | Not started |
| **OSCAL SSP Component** | Medium | Machine-readable control documentation | Partial |
| **Penetration Test Report** | Medium | Third-party pen test results | Not started |
| **Supply Chain Risk Management Plan** | Medium | How vendor secures supply chain | Not started |
| **Incident Response Plan** | Medium | How vendor handles security incidents | Exists |
| **Data Flow Diagrams** | Medium | Where data moves, encryption at rest/transit | Partially exists |
| **Privacy Impact Assessment** | Low | PII handling (minimal for Aegis) | Not started |

### 6.2 OSCAL -- Open Security Controls Assessment Language

**What is OSCAL?**

OSCAL is a NIST-developed set of standardized, machine-readable formats (JSON, XML, YAML) for security assessment documentation. It was created to automate the production, exchange, and processing of security information.

**OSCAL document types:**

| Model | Purpose | Aegis Use Case |
|-------|---------|----------------|
| **Catalog** | Defines security controls (e.g., NIST 800-53) | Reference -- already have catalog files |
| **Profile** | Selects and tailors controls from a catalog | Define Aegis's applicable control baseline |
| **Component Definition** | Documents how a component implements controls | **Primary deliverable** -- describes how Aegis satisfies controls |
| **System Security Plan (SSP)** | Full system security documentation | Customer creates this; Aegis provides components |
| **Assessment Plan** | How to assess control implementation | For assessors |
| **Assessment Results** | Findings from assessment | For assessors |
| **POA&M** | Tracks remediation of findings | For system owners |

**Should Aegis support OSCAL?**

**Yes -- high priority.** Here is why:

1. **FedRAMP 20x** (the streamlined authorization program launching 2025-2026) requires OSCAL-formatted submissions
2. **DoD is moving toward OSCAL** for RMF documentation automation
3. **Competitive advantage** -- few vendors provide OSCAL-formatted documentation today
4. **Automation** -- OSCAL enables automated compliance checking, reducing ATO timelines

**What Aegis should produce in OSCAL format:**

1. **Component Definition** -- An OSCAL component-definition document describing:
   - Each Aegis component (platform-api, proxy, k8s-agent)
   - Which NIST 800-53 controls each component satisfies
   - How each control is implemented
   - Configuration parameters that affect control implementation

2. This component definition can then be imported into a customer's OSCAL SSP, automatically populating inherited control descriptions.

**Tools for OSCAL generation:**
- NIST OSCAL tools: https://github.com/usnistgov/OSCAL
- compliance-trestle (IBM/Red Hat): https://github.com/oscal-compass/compliance-trestle -- Python CLI for OSCAL authoring
- Lula (Defense Unicorns): https://github.com/defenseunicorns/lula -- Kubernetes-native compliance engine that generates OSCAL from live clusters

**Aegis already has:**
- `docs/compliance/oscal/fedramp_rev5_moderate_baseline.json`
- `docs/compliance/oscal/nist_sp_800_53_rev5_catalog.json`
- NIST 800-53 control mappings in `docs/compliance/control-map.md`

**Next step:** Use compliance-trestle or manually create an OSCAL component-definition JSON for Aegis.

### 6.3 The ATO Process (Customer Side)

For context, here is what the customer goes through. Understanding this helps Aegis provide the right documentation.

```
DoD RMF Process (DoDI 8510.01):

Step 1: CATEGORIZE the system (FIPS 199 impact level)
        --> Aegis provides: data types processed, impact analysis

Step 2: SELECT security controls (NIST 800-53 baseline)
        --> Aegis provides: control applicability guidance

Step 3: IMPLEMENT controls
        --> Aegis provides: control implementation statements, hardening guide

Step 4: ASSESS controls (by assessor/SCA)
        --> Aegis provides: test procedures, scan results, OSCAL artifacts

Step 5: AUTHORIZE (AO decision)
        --> Aegis provides: risk acceptance rationale, residual risk statement

Step 6: MONITOR continuously
        --> Aegis provides: vulnerability disclosures, patches, SBOM updates
```

**Types of ATO in DoD:**

| Type | Duration | Scope | Notes |
|------|----------|-------|-------|
| **ATO** | 3 years | System-specific | Standard authorization |
| **cATO** (Continuous) | Ongoing | System-specific | Requires continuous monitoring automation |
| **IATT** (Interim) | Up to 6 months | Limited scope | Temporary authorization for testing |
| **ATO with Conditions** | 3 years | System-specific | ATO with documented risk acceptance |
| **Type Authorization** | 3 years | Multiple instances | For software deployed in multiple locations |

**For Aegis:** A **Type Authorization** is ideal because it would allow a single authorization to cover Aegis deployed across multiple DoD sites.

---

## 7. Cyber-Specific Funding Opportunities

### 7.1 DHS SBIR/STTR Topics

The Department of Homeland Security (DHS) Science and Technology Directorate publishes SBIR topics annually. Recent and recurring topics relevant to Aegis:

| Topic Area | Typical Focus | Aegis Relevance |
|------------|---------------|-----------------|
| **Cybersecurity for Critical Infrastructure** | Protecting industrial control systems, networks | Moderate -- if Aegis secures critical infrastructure AI |
| **AI/ML Security** | Securing AI pipelines, adversarial ML defense | High -- Aegis schedules ML/AI workloads |
| **Container Security** | Hardening container environments, runtime protection | High -- direct match |
| **Zero Trust Architecture** | Implementing ZTA components | High -- Aegis implements mTLS, policy enforcement |
| **Supply Chain Security** | Software supply chain integrity | Moderate -- SBOM, image signing |

**How to find current topics:**
- DHS SBIR portal: https://www.dhs.gov/science-and-technology/sbir
- SBIR.gov: https://www.sbir.gov (search all agencies)
- Typical DHS SBIR Phase I: $150K-$200K, 6 months
- Typical DHS SBIR Phase II: $750K-$1.5M, 24 months

### 7.2 DoD SBIR/STTR Topics

DoD components publish SBIR topics through their respective channels:

| Component | Portal | Typical Cyber Topics |
|-----------|--------|---------------------|
| **DARPA** | sbir.darpa.mil | Advanced cybersecurity R&D |
| **Army** | SBIR portal via SAM.gov | Tactical network security, edge computing |
| **Navy** | NAVAIR/NAVSEA/SPAWAR SBIR | Maritime cybersecurity, network defense |
| **Air Force** | AFWERX/AF SBIR | Cloud security, DevSecOps tooling |
| **Space Force** | SpaceWERX | Satellite ground system security |
| **DISA** | DISA SBIR (less common) | Enterprise cyber defense |
| **NSA** | NSA SBIR (rare) | Cryptography, network defense |

**Key DoD SBIR programs for startups:**
- **AFWERX Open Topic** (Air Force): Accepts proposals year-round for any innovative technology. Phase I up to $75K.
- **Army xTech**: Pitch competitions for technology solutions. Prizes from $10K to $250K+.
- **NavalX**: Navy innovation accelerator with SBIR integration.

### 7.3 NSA Cybersecurity Programs

The National Security Agency (NSA) has several programs relevant to cybersecurity vendors:

| Program | Description | Engagement Path |
|---------|-------------|-----------------|
| **NSA Cybersecurity Directorate** | Publishes guidance on securing systems | Follow guidance, align product |
| **CNSA 2.0** (Commercial National Security Algorithm Suite) | Algorithm requirements for national security systems | Ensure Aegis supports required algorithms |
| **CSfC** (Commercial Solutions for Classified) | Framework for using commercial products for classified data | Long-term goal if serving classified environments |
| **NSA SBIR** | Occasional SBIR topics on cybersecurity | Monitor SBIR.gov |
| **Cybersecurity Collaboration Center (CCC)** | Public-private partnership for threat intelligence | Register for threat feeds |

**NSA/CISA Kubernetes Hardening Guide:**
NSA published "Kubernetes Hardening Guide" (jointly with CISA) which is widely referenced by DoD customers. Aegis should document compliance with this guide.

- Latest version: "Kubernetes Hardening Guide v1.2" (August 2022, with updates through 2024)
- Covers: pod security, network separation, authentication, audit logging, upgrade practices
- **Action:** Create a compliance matrix showing how Aegis addresses each recommendation

### 7.4 Cyber-Focused OTAs and BAAs

**Active or Recurring Vehicles:**

| Vehicle | Organization | Focus | How to Access |
|---------|-------------|-------|---------------|
| **Tradewinds OTA** | CDAO | AI/data solutions | tradewindai.com |
| **DARPA BAA** | DARPA | Cyber R&D | SAM.gov |
| **CISA BAA** | DHS/CISA | Cybersecurity tools | SAM.gov |
| **DHS S&T SVIP** | DHS | Silicon Valley Innovation Program | SAM.gov |
| **SOCOM Broad Agency Announcement** | SOCOM | SOF-specific cyber | SAM.gov |
| **Army Application Lab** | Army | Various including cyber | armyfuturescommand.com |
| **AFWERX STRATFI/TACFI** | Air Force | Strategic/Tactical Financing | AFWERX |
| **DIU Commercial Solutions Opening** | DIU | Commercial tech for DoD | diu.mil |
| **In-Q-Tel** | Intelligence Community | Strategic investment | in-q-tel.org |

**Most promising for Aegis (ranked):**

1. **DIU Commercial Solutions Opening** -- Aegis is commercial technology solving a DoD problem (GPU workload orchestration). DIU specifically seeks commercial products, not R&D projects.
2. **Tradewinds OTA** -- CDAO is actively seeking AI infrastructure solutions.
3. **AFWERX Open Topic SBIR** -- Low barrier to entry, accepts year-round.
4. **Army xTech** -- Pitch competition format, good visibility.
5. **DHS S&T SVIP** -- If positioning around critical infrastructure cybersecurity.

### 7.5 Other Relevant Funding Sources

| Source | Type | Amount | Notes |
|--------|------|--------|-------|
| **NSF SBIR** | Grant | $275K (Phase I), $1M (Phase II) | Cybersecurity topics in Computer and Information Science |
| **DOE SBIR** | Grant | $200K (Phase I), $1.1M (Phase II) | HPC/GPU computing for national labs |
| **NIST SBIR** | Grant | $100K (Phase I) | Cybersecurity measurement, standards |
| **Cyber Grants (state-level)** | Grant | Varies | State and local cybersecurity programs |
| **SBA HUBZone** | Preference | N/A | Contract preference if in HUBZone |
| **8(a) Program** | Sole-source capability | N/A | If eligible, enables sole-source DoD contracts up to $4.5M |

---

## 8. Recommended Action Plan for Aegis

### Immediate (Next 30 Days)

| Priority | Action | Estimated Effort | Cost |
|----------|--------|------------------|------|
| 1 | Complete CMMC Level 1 self-assessment and submit SPRS score | 2-3 days | $0 |
| 2 | Create NSA/CISA Kubernetes Hardening Guide compliance matrix | 3-5 days | $0 |
| 3 | Register on SAM.gov if not already registered (required for any federal contracting) | 1 day | $0 |
| 4 | Register on Tradewinds (tradewindai.com) | 1 day | $0 |
| 5 | Build FIPS 140 cryptographic inventory document | 2 days | $0 |

### Short-Term (60-90 Days)

| Priority | Action | Estimated Effort | Cost |
|----------|--------|------------------|------|
| 1 | Create DISA Kubernetes STIG compliance matrix | 1-2 weeks | $0 |
| 2 | Generate OSCAL component-definition for Aegis | 1 week | $0 |
| 3 | Submit Aegis images to Iron Bank | 2-3 weeks (iterative) | $0 |
| 4 | Create vendor security documentation package (consolidated) | 2 weeks | $0 |
| 5 | Submit AFWERX Open Topic SBIR Phase I proposal | 1-2 weeks | $0 |
| 6 | Apply to DIU Commercial Solutions Opening (if relevant topic open) | 2-3 weeks | $0 |

### Medium-Term (3-6 Months)

| Priority | Action | Estimated Effort | Cost |
|----------|--------|------------------|------|
| 1 | Implement FIPS 140 cryptography (Go BoringCrypto) | 2-4 weeks dev | $0 |
| 2 | Begin CMMC Level 2 preparation (gap analysis against NIST 800-171) | 2-3 weeks | $5K-$15K (optional consultant) |
| 3 | Pursue Tradewinds OTA submission | 2-3 weeks | $0 |
| 4 | Engage with Platform One / Iron Bank team | Ongoing | $0 |
| 5 | Publish vulnerability disclosure policy | 1-2 days | $0 |

### Long-Term (6-12 Months)

| Priority | Action | Estimated Effort | Cost |
|----------|--------|------------------|------|
| 1 | Achieve CMMC Level 2 certification | 3-5 months | $30K-$50K |
| 2 | Obtain GSA Schedule (MAS) for easier procurement | 2-4 months | $5K-$15K |
| 3 | Pursue Type Authorization for Aegis (with agency sponsor) | 6-12 months | $20K-$40K |
| 4 | Consider cATO capability (continuous monitoring automation) | Ongoing | $0 (dev time) |

### Key Registrations Needed

| Registration | URL | Purpose | Required? |
|-------------|-----|---------|-----------|
| **SAM.gov** | sam.gov | Federal contractor registration, required for any contract | Yes |
| **SPRS** | sprs.csd.disa.mil | Submit CMMC self-assessment scores | Yes (for CMMC) |
| **Tradewinds** | tradewindai.com | CDAO marketplace | Recommended |
| **SBIR.gov** | sbir.gov | SBIR/STTR proposal submission | For SBIR |
| **Platform One** | login.dso.mil | Iron Bank, Big Bang access | Recommended |
| **DIU** | diu.mil | Commercial Solutions Opening | Recommended |

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| **3PAO** | Third-Party Assessment Organization (FedRAMP assessors) |
| **AO** | Authorizing Official (person who grants ATO) |
| **APL** | Approved Products List |
| **ATO** | Authority to Operate |
| **BAA** | Broad Agency Announcement |
| **C3PAO** | CMMC Third-Party Assessment Organization |
| **cATO** | Continuous Authority to Operate |
| **CC SRG** | Cloud Computing Security Requirements Guide |
| **CDAO** | Chief Digital and AI Office |
| **CMMC** | Cybersecurity Maturity Model Certification |
| **CNSA** | Commercial National Security Algorithm Suite |
| **CRM** | Customer Responsibility Matrix |
| **CSfC** | Commercial Solutions for Classified |
| **CSO** | Commercial Solutions Opening (DIU) or Cloud Service Offering (FedRAMP) |
| **CUI** | Controlled Unclassified Information |
| **DFARS** | Defense Federal Acquisition Regulation Supplement |
| **DIBCAC** | Defense Industrial Base Cybersecurity Assessment Center |
| **DISA** | Defense Information Systems Agency |
| **DIU** | Defense Innovation Unit |
| **DoDIN** | Department of Defense Information Network |
| **FAR** | Federal Acquisition Regulation |
| **FCI** | Federal Contract Information |
| **FIPS** | Federal Information Processing Standard |
| **IATT** | Interim Authority to Test |
| **IL** | Impact Level (CC SRG) |
| **NIST** | National Institute of Standards and Technology |
| **NSIN** | National Security Innovation Network |
| **OSCAL** | Open Security Controls Assessment Language |
| **OTA** | Other Transaction Authority |
| **POA&M** | Plan of Action and Milestones |
| **RMF** | Risk Management Framework |
| **SBIR** | Small Business Innovation Research |
| **SCA** | Security Control Assessor |
| **SCRM** | Supply Chain Risk Management |
| **SPRS** | Supplier Performance Risk System |
| **SRG** | Security Requirements Guide |
| **SSP** | System Security Plan |
| **STIG** | Security Technical Implementation Guide |
| **STTR** | Small Business Technology Transfer |

## Appendix B: Key URLs

| Resource | URL |
|----------|-----|
| DISA STIGs | https://public.cyber.mil/stigs/downloads/ |
| DISA Iron Bank | https://ironbank.dso.mil |
| Platform One | https://p1.dso.mil |
| SAM.gov (contract registration) | https://sam.gov |
| SBIR.gov (grant proposals) | https://www.sbir.gov |
| SPRS (CMMC scores) | https://sprs.csd.disa.mil |
| Tradewinds (CDAO) | https://tradewindai.com |
| DIU | https://www.diu.mil |
| AFWERX | https://afwerx.com |
| NIST OSCAL | https://pages.nist.gov/OSCAL/ |
| compliance-trestle (OSCAL tool) | https://github.com/oscal-compass/compliance-trestle |
| Lula (K8s compliance) | https://github.com/defenseunicorns/lula |
| NSA K8s Hardening Guide | https://media.defense.gov/2022/Aug/29/2003066362/-1/-1/0/CTR_KUBERNETES_HARDENING_GUIDANCE_1.2_20220829.PDF |
| Cyber AB (C3PAO list) | https://cyberab.org |
| NIST SP 800-171 | https://csrc.nist.gov/publications/detail/sp/800-171/rev-2/final |
| FedRAMP Marketplace | https://marketplace.fedramp.gov |

## Appendix C: Aegis Compliance Asset Inventory

These existing Aegis documents support DoD market entry:

| Document | Path | DoD Use |
|----------|------|---------|
| Customer Responsibility Matrix | `docs/compliance/customer-docs/customer-responsibility-matrix.md` | ATO support |
| Security Architecture Guide | `docs/compliance/customer-docs/security-architecture-guide.md` | ATO support |
| Control Implementation Statements | `docs/compliance/customer-docs/control-implementation-statements.md` | ATO support |
| Configuration Hardening Guide | `docs/compliance/customer-docs/configuration-hardening-guide.md` | STIG compliance |
| Incident Response Runbook | `docs/compliance/customer-docs/incident-response-runbook.md` | ATO support |
| NIST 800-53 Control Map | `docs/compliance/control-map.md` | Control mapping |
| OSCAL Catalogs | `docs/compliance/oscal/` | OSCAL foundation |
| FedRAMP Gap Analysis | `docs/compliance/fedramp/02-gap-analysis.md` | Planning |
| FedRAMP Status | `docs/compliance/fedramp/STATUS.md` | Planning |
| SOC 2 Controls | `docs/compliance/soc2/` | Reusable evidence |
| ISO 27001 ISMS | `docs/compliance/iso27001/` | Reusable evidence |
| SBOM Structure | `docs/compliance/sbom/` | Supply chain |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-08 | Carlos Sanchez | Initial research document |
