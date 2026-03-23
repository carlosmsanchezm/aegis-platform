# DIU and OTA Engagement Guide for Small Companies (2025-2026)

> **Audience**: 1-person small business with a commercial technology product (e.g., DevSecOps platform, Kubernetes infrastructure, AI/ML tooling) seeking to sell to the Department of Defense through non-traditional contracting pathways.
>
> **Last updated**: March 2026. Verify all URLs, dates, and dollar figures against current DIU postings before acting.

---

## Table of Contents

1. [What Is DIU](#1-what-is-diu)
2. [Current DIU Focus Areas (2025-2026)](#2-current-diu-focus-areas-2025-2026)
3. [How to Submit a Solution to DIU](#3-how-to-submit-a-solution-to-diu)
4. [Commercial Solutions Opening (CSO)](#4-commercial-solutions-opening-cso)
5. [OTA Agreements Explained](#5-ota-agreements-explained)
6. [Eligibility for a 1-Person Company](#6-eligibility-for-a-1-person-company)
7. [Prototype to Production Pathway](#7-prototype-to-production-pathway)
8. [Typical Award Sizes and Timelines](#8-typical-award-sizes-and-timelines)
9. [Success Stories from Small Companies](#9-success-stories-from-small-companies)
10. [NSIN and Related Programs](#10-nsin-and-related-programs)
11. [Getting on DIU's Radar: Contacts, Events, and Strategies](#11-getting-on-dius-radar-contacts-events-and-strategies)
12. [Registration Checklist](#12-registration-checklist)
13. [Positioning a Kubernetes/DevSecOps Platform for DIU](#13-positioning-a-kuberneticsdevsecops-platform-for-diu)
14. [Key Terminology](#14-key-terminology)

---

## 1. What Is DIU

### Mission

The **Defense Innovation Unit (DIU)** is a DoD organization that accelerates the adoption of commercial technology into the U.S. military. It was established in 2015 (originally as "DIUx") and reports directly to the Under Secretary of Defense for Research and Engineering (USD(R&E)).

In 2024, DIU was elevated in status -- it now reports directly to the Deputy Secretary of Defense and has expanded authorities. This elevation signals the Pentagon's increased seriousness about commercial technology adoption.

### How DIU Works

DIU operates fundamentally differently from traditional defense acquisition:

- **Problem-first**: DoD components (Army, Navy, Air Force, Space Force, combatant commands) bring operational problems to DIU. DIU then finds commercial solutions.
- **Speed**: DIU aims to award prototype agreements within 60-90 days of a solicitation, compared to 12-18 months for traditional contracts.
- **Commercial solutions**: DIU specifically seeks companies that have existing commercial products or near-commercial prototypes. They are NOT looking for companies that only do government work.
- **OTA authority**: DIU uses Other Transaction (OT) agreements rather than FAR-based contracts, which dramatically reduces compliance burden.
- **Dual-use focus**: The technology must have both commercial and military applications.

### What DIU Is NOT

- Not a grant program (you must deliver a working prototype)
- Not a research lab (they want near-ready solutions, not basic research)
- Not DARPA (DARPA funds R&D from scratch; DIU adapts existing commercial tech)
- Not a traditional contracting office (no FAR, no DFARS for prototypes)

### Offices

| Location | Role |
|----------|------|
| Mountain View, CA (HQ) | Primary hub, Silicon Valley outreach |
| Boston, MA | East Coast tech ecosystem |
| Austin, TX | Emerging tech hub |
| Washington, DC (Pentagon) | DoD coordination |
| Chicago, IL | Midwest tech ecosystem |

---

## 2. Current DIU Focus Areas (2025-2026)

DIU organizes work into portfolio areas. The ones most relevant to a Kubernetes/DevSecOps/AI platform:

### Autonomy and AI

- Autonomous systems, computer vision, predictive maintenance
- AI/ML for decision support, data analytics, natural language processing
- Responsible AI frameworks and testing
- **Relevance**: If your platform enables AI workload orchestration or MLOps across classified/unclassified environments

### Cyber

- Zero-trust architecture implementation
- Software supply chain security
- Continuous monitoring and threat detection
- Cloud security posture management
- **Relevance**: Multi-cluster Kubernetes security, workload isolation, compliance enforcement

### Enterprise and Digital Infrastructure

- Cloud infrastructure modernization (this is directly relevant)
- DevSecOps pipelines and platform engineering
- Container orchestration and microservices
- Software factories and continuous integration/deployment
- Edge computing platforms
- **Relevance**: This is the sweet spot for a hub-and-spoke Kubernetes platform

### Space

- Satellite data processing, ground station software
- Space domain awareness
- **Relevance**: If your platform can orchestrate workloads across ground stations or edge nodes

### Human Systems

- Remote collaboration, distributed workforce tools
- Developer experience and productivity
- **Relevance**: Remote development environments, VS Code workspaces

### Key Themes DIU Cares About in 2025-2026

1. **Platform One / Big Bang alternatives**: DoD is looking for alternatives and complements to the Air Force's Platform One (now Party Bus). If you can show how your platform serves a similar function (DevSecOps, container orchestration, compliance) but is more flexible or commercially maintained, this is high-value.
2. **CJADC2 (Combined Joint All-Domain Command and Control)**: Multi-domain data sharing requires infrastructure that can span multiple classification levels and networks.
3. **Zero Trust**: Executive Order 14028 mandates zero-trust architecture across federal agencies. Kubernetes-native zero-trust is in demand.
4. **Software factories**: Each military service is building software factories. They need underlying platform infrastructure.
5. **Edge/hybrid cloud**: Running workloads across cloud, on-premise, and tactical edge -- exactly what a hub-and-spoke architecture addresses.
6. **AI infrastructure**: Not just AI models, but the platforms that deploy, manage, and monitor AI workloads at scale.

---

## 3. How to Submit a Solution to DIU

### The Portal

**URL**: https://www.diu.mil/work-with-us

DIU uses a submission portal where companies can:

1. **Respond to an open CSO area of interest** (preferred -- higher chance of engagement)
2. **Submit an unsolicited solution** through the general portal

### Step-by-Step Process

**Step 1: Create an account on the DIU portal**
- Go to https://submissions.diu.mil (or follow links from diu.mil)
- Register with your company information
- No special clearances needed to submit

**Step 2: Review open Areas of Interest (AOIs)**
- Each CSO lists specific AOIs describing problems DoD needs solved
- Match your capability to the closest AOI
- If no AOI fits, you can submit a general/unsolicited solution

**Step 3: Prepare your submission**

The submission is intentionally lightweight compared to traditional government proposals. Typical requirements:

| Element | Details |
|---------|---------|
| Company overview | 1-2 paragraphs: who you are, what you do commercially |
| Solution description | How your technology addresses the AOI (2-5 pages typical) |
| Technical approach | Architecture, key differentiators, current TRL (Technology Readiness Level) |
| Commercial traction | Existing customers, revenue, deployments -- this is critical |
| Team | Key personnel (even if it is just you) |
| Prototype plan | What you would demonstrate in a 12-24 month prototype phase |
| Cost estimate | Rough order of magnitude for the prototype |
| Relevant past performance | Commercial customers, not necessarily government |

**What DIU does NOT require in submissions**:
- No SAM.gov registration (needed later for award, not for submission)
- No security clearance
- No CMMC certification (for prototype phase)
- No extensive cost accounting systems
- No certified cost or pricing data

**Step 4: DIU reviews and responds**
- DIU typically acknowledges receipt quickly
- Technical review takes 2-6 weeks
- If interested, DIU schedules a pitch meeting (30-60 minutes, usually virtual)
- Multiple companies may pitch for the same AOI

**Step 5: Down-select and negotiation**
- DIU selects one or more companies for prototype OT agreements
- Negotiation of terms, milestones, and pricing
- Award of prototype OT agreement

### Tips for a Strong Submission

1. **Lead with commercial traction**: DIU's entire model is built on the premise that commercial tech is ahead of government-developed tech. Show that real companies pay for your product.
2. **Speak their language but don't over-militarize**: Use terms like "multi-tenant," "zero-trust," "edge deployment," "air-gapped operation" but describe your commercial product, not a hypothetical military version.
3. **Be specific about what you would prototype**: "We would deploy our platform on an IL-5 EKS cluster and demonstrate multi-cluster workload scheduling across two ATO'd environments" is better than "We would adapt our platform for DoD use."
4. **Show TRL 6+ readiness**: Your technology should be at least at demonstration level in a relevant environment. A commercial product in production is ideal (TRL 8-9).
5. **Address security early**: Mention your approach to FedRAMP, STIG compliance, container hardening, SBOM generation -- even if not yet certified.

---

## 4. Commercial Solutions Opening (CSO)

### What Is a CSO?

A **Commercial Solutions Opening (CSO)** is DIU's primary solicitation mechanism. It is NOT a traditional RFP (Request for Proposal). Key characteristics:

- **Always open**: The CSO is a standing, open solicitation. Companies can submit at any time.
- **Areas of Interest**: Within the CSO, DIU publishes specific AOIs that describe problems the DoD needs solved. AOIs are added, updated, and closed on a rolling basis.
- **Merit-based**: Submissions are evaluated on technical merit, commercial viability, and relevance to the AOI -- not lowest price.
- **Competitive but collaborative**: DIU may award to multiple vendors for the same problem area.

### How CSO Differs from Traditional RFPs

| Aspect | Traditional RFP (FAR-based) | DIU CSO |
|--------|---------------------------|---------|
| Page count | 100-500+ pages | 5-15 pages typical |
| Evaluation | Lowest Price Technically Acceptable (often) | Best technical solution with commercial viability |
| Timeline to award | 12-24 months | 60-90 days (goal) |
| Contract type | FAR/DFARS contract | Other Transaction agreement |
| Compliance burden | Extensive (DCAA, CAS, etc.) | Minimal for prototype phase |
| Who can compete | Typically only established defense contractors | Designed for commercial companies |

### Finding Open CSOs and AOIs

1. **Primary source**: https://www.diu.mil/work-with-us -- lists all current AOIs
2. **SAM.gov**: CSO also posted on SAM.gov (search for "DIU" or "Defense Innovation Unit")
3. **Subscribe to DIU newsletter**: Sign up at diu.mil for updates on new AOIs
4. **Follow DIU on LinkedIn**: They announce new AOIs and events

### AOIs Relevant to DevSecOps/AI/Infrastructure (Check for Current Postings)

DIU regularly posts AOIs in these categories that could match a Kubernetes platform:

- **Cloud infrastructure modernization**: Multi-cloud/hybrid-cloud orchestration
- **DevSecOps platforms**: CI/CD, container security, software factory infrastructure
- **AI/ML infrastructure**: MLOps platforms, model deployment and monitoring
- **Zero-trust implementation**: Network segmentation, identity-aware proxies, microsegmentation
- **Edge computing**: Deploying cloud-native workloads to resource-constrained environments
- **Autonomous systems infrastructure**: Backend platforms for managing autonomous system fleets
- **Cybersecurity**: Continuous monitoring, vulnerability management, compliance automation

### Responding to a CSO AOI

1. Read the AOI carefully -- note the specific problem statement, desired outcomes, and evaluation criteria
2. Prepare your submission per the portal requirements (see Section 3)
3. Submit through the DIU portal before any stated deadline (some AOIs are open-ended, others have deadlines)
4. Be prepared for a pitch meeting if selected for further evaluation

---

## 5. OTA Agreements Explained

### What Is an Other Transaction (OT)?

An **Other Transaction (OT)** authority is a contracting mechanism authorized by 10 U.S.C. 4022 that allows DoD to enter into agreements that are NOT traditional procurement contracts, grants, or cooperative agreements. This means they are NOT subject to the Federal Acquisition Regulation (FAR) or Defense FAR Supplement (DFARS).

### Types of OTs

| Type | Statute | Purpose | Typical Value |
|------|---------|---------|---------------|
| **Prototype OT** | 10 U.S.C. 4022 | Develop and demonstrate a prototype | $500K - $50M+ |
| **Production OT** | 10 U.S.C. 4022(f) | Produce and deploy a successful prototype | Can be very large ($100M+) |
| **Research OT** | 10 U.S.C. 4021 | Basic/applied research | Varies |

DIU primarily uses **Prototype OTs** with a follow-on **Production OT** pathway.

### Why OTs Are Good for Small Companies

**Reduced compliance burden:**
- No FAR/DFARS clauses (hundreds of pages of regulations you do not have to comply with)
- No DCAA (Defense Contract Audit Agency) accounting system requirements for prototype phase
- No certified cost or pricing data requirement
- No Cost Accounting Standards (CAS) compliance
- Simplified intellectual property terms (you typically retain your commercial IP)

**Speed:**
- Awards in weeks to months, not years
- Simplified negotiation process
- No lengthy source selection boards

**IP protection:**
- Under OTs, the government typically gets government-purpose rights to modifications made during the prototype, but the contractor retains full rights to pre-existing IP and commercial products
- This is dramatically better than traditional contracts where the government may claim unlimited rights
- You can negotiate IP terms -- DIU is generally reasonable about protecting commercial IP

**Flexible terms:**
- Payment milestones (not cost-reimbursement)
- Fixed-price elements are common
- No requirement for government-approved accounting systems during prototype

**Nontraditional defense contractor status:**
- OT prototype authority requires participation by a "nontraditional defense contractor" (see Section 6)
- A 1-person commercial tech company qualifies by default
- This is actually an advantage: your nontraditional status enables the use of OT authority

### Key OT Terms to Know

| Term | Meaning |
|------|---------|
| **Nontraditional defense contractor** | A company that has not done significant FAR-based contract work in the past year -- this is YOU, and it is an advantage |
| **Prototype** | A working demonstration of your technology in a relevant DoD environment |
| **Follow-on production** | A sole-source production contract/OT that can be awarded after a successful prototype without re-competing |
| **Milestone payments** | You get paid when you hit defined technical milestones (not cost-reimbursement) |
| **Government-purpose rights** | Government can use for any government purpose but cannot share commercially |
| **IRAD** | Independent Research and Development -- your own R&D costs that you do not charge to the government |

### OT vs. SBIR/STTR

| Aspect | OT (via DIU) | SBIR/STTR |
|--------|-------------|-----------|
| What is funded | Adapting existing commercial tech | R&D of new technology |
| TRL at entry | 6+ (demonstrated in relevant environment) | 1-4 (concept to lab demo) |
| Timeline | Months to award | Months to years |
| Follow-on path | Direct to production OT (sole source) | Phase III (competitive or sole source) |
| IP terms | Negotiable, generally favorable | Statutory protections, favorable |
| Best for | Companies with existing products | Companies with novel R&D |
| Cost sharing | Often required (but negotiable) | Generally not required |

---

## 6. Eligibility for a 1-Person Company

### The Short Answer

**Yes, a 1-person company can engage with DIU and receive OT agreements.** There is no minimum company size, employee count, or revenue threshold. In fact, being a small, nontraditional company is an advantage because:

1. You automatically qualify as a **nontraditional defense contractor** (which enables OT authority)
2. DIU's entire mission is to bring commercial companies into the defense ecosystem
3. Small companies are more agile and can prototype faster

### Entity Requirements

You must be a legal business entity. The following are acceptable:

| Entity Type | Acceptable? | Notes |
|-------------|-------------|-------|
| LLC (single-member) | Yes | Most common for solo founders |
| S-Corp | Yes | |
| C-Corp | Yes | Preferred if seeking VC funding later |
| Sole proprietorship | Technically yes, but not recommended | LLC provides better liability protection |
| Partnership | Yes | |

### Registration Requirements

**For submission/pitch** (no registration needed):
- You can submit to DIU's CSO portal and pitch without any government registrations
- This is a major advantage of DIU over traditional contracting

**Before award** (required for receiving money):

| Registration | What | Where | Timeline | Cost |
|-------------|------|-------|----------|------|
| **SAM.gov** | System for Award Management | sam.gov | 2-4 weeks to process | Free |
| **UEI** | Unique Entity ID (replaced DUNS) | Obtained through SAM.gov | Part of SAM registration | Free |
| **CAGE Code** | Commercial and Government Entity code | Assigned during SAM registration | Part of SAM registration | Free |
| **EIN** | Employer Identification Number | IRS (irs.gov) | Immediate to 4 weeks | Free |
| **Bank account** | Business bank account for payments | Any bank | Varies | Varies |

**NOT required (for prototype OT phase):**
- Security clearance (facility or personnel)
- CMMC certification
- DCAA-approved accounting system
- GSA Schedule
- SBA certification (8(a), HUBZone, etc.) -- though having these can help with other contracts

### Practical Considerations for a Solo Founder

1. **Teaming**: DIU may suggest teaming with a larger company (system integrator) for production phase. This is normal and can be beneficial. For prototype, solo is fine.
2. **Bandwidth**: A prototype OT will require dedicated effort. DIU understands small companies have constraints, but you need to deliver on milestones.
3. **Cost sharing**: DIU prototype OTs sometimes require cost sharing (you contribute some of your own resources). For a 1-person company, your time/existing product development can count as cost share.
4. **Subcontracting**: You can subcontract parts of the work. No special approvals needed under OT (unlike FAR contracts with subcontracting plan requirements).
5. **Insurance**: Consider general liability and professional liability (E&O) insurance. Not legally required for OT but prudent.

### SBA Small Business Designations (Optional but Helpful)

While not required for DIU, these designations open additional opportunities:

| Designation | Benefit | How to Get |
|-------------|---------|------------|
| Small Business | Default if under size standard | Automatic based on NAICS code |
| SDVOSB | Service-Disabled Veteran-Owned | SBA certification |
| WOSB | Women-Owned Small Business | SBA certification |
| 8(a) | Disadvantaged business | SBA application (lengthy) |
| HUBZone | Located in underutilized area | SBA certification |

NAICS codes relevant to your work:
- **541512**: Computer Systems Design Services
- **541511**: Custom Computer Programming Services
- **518210**: Computing Infrastructure Providers, Data Processing
- **541519**: Other Computer Related Services

---

## 7. Prototype to Production Pathway

This is the most important section for long-term revenue potential. The prototype-to-production pathway is what makes DIU uniquely valuable.

### The Pathway

```
CSO Submission --> Pitch --> Prototype OT Award --> Prototype Execution -->
Prototype Assessment --> Production OT (sole source) --> Full Deployment
```

### Phase Details

**Phase 1: Prototype OT (12-24 months typical)**

- Demonstrate your technology in a relevant DoD environment
- Meet defined milestones (technical demonstrations, user feedback sessions, security assessments)
- Typical prototype value: $500K - $5M for small companies
- Government evaluates: Does this solve the problem? Can it scale? Is it secure?

**Phase 2: Prototype Completion Assessment**

- DIU and the DoD customer assess prototype results
- Key questions: Did it work? Do users want it? Can it meet security requirements? Is there a viable path to production?
- If successful, the DoD component can proceed to production WITHOUT full and open competition

**Phase 3: Production OT or FAR Contract**

Under 10 U.S.C. 4022(f), a successful prototype can transition to production:

- **Sole-source production OT**: If you competitively won the prototype, the production follow-on can be awarded sole-source (no re-competition). This is huge.
- **Production FAR contract**: Alternatively, the government can award a traditional contract. This requires more compliance but may be preferred by some DoD components.
- **Indefinite Delivery/Indefinite Quantity (IDIQ)**: Sometimes structured as an IDIQ with task orders

**Production values can be 10-100x the prototype value.** A $1M prototype can lead to a $50M+ production contract.

### What Makes the Transition Happen

1. **User champion**: A military end-user who loves your product and advocates internally
2. **Funded requirement**: The DoD component must have budget allocated for the production capability
3. **ATO (Authority to Operate)**: Your system must receive security authorization. For cloud systems, this often means FedRAMP or DoD IL (Impact Level) authorization.
4. **Acquisition strategy**: The DoD component's acquisition office must write a strategy memo justifying the sole-source follow-on

### Common Failure Points (and How to Avoid Them)

| Failure Point | How to Avoid |
|--------------|--------------|
| No user champion | Spend time with actual military users during prototype, not just DIU staff |
| No funding for production | Ask early: "Who owns the production budget?" Get the PEO/PM involved from day 1 |
| ATO takes too long | Start security documentation during prototype, not after. Build to STIG/SRG standards from the start |
| Technology works but does not integrate | Understand the existing environment (e.g., IL-5 AWS GovCloud, NIPR/SIPR networks) early |
| Change of personnel | Document everything. Military people rotate every 2-3 years |

---

## 8. Typical Award Sizes and Timelines

### Prototype OT Awards

| Metric | Range | Typical |
|--------|-------|---------|
| Award value | $250K - $50M | $1M - $5M for small tech companies |
| Duration | 6 - 24 months | 12-18 months |
| Time from submission to award | 30 - 180 days | 60-120 days |
| Number of milestones | 3 - 8 | 4-6 |
| Cost sharing requirement | 0% - 50% | Often 33% for nontraditional |

### Production OT/Contract Awards

| Metric | Range | Notes |
|--------|-------|-------|
| Award value | $5M - $500M+ | Depends on scope of deployment |
| Duration | 1 - 5 years | Often with option years |
| Vehicles | Production OT, FAR contract, IDIQ | Varies by DoD component preference |

### Timeline Summary

```
Month 0:      Submit to CSO
Month 1-2:    DIU review, pitch meeting
Month 2-3:    Down-select, negotiation
Month 3-4:    Prototype OT awarded
Month 4-18:   Prototype execution (milestones)
Month 18-20:  Assessment, production planning
Month 20-24:  Production OT awarded
Month 24+:    Production deployment, revenue growth
```

### Payment Structure

- **Milestone-based**: Most common. You submit a milestone report, DIU reviews, payment is released.
- **Fixed-price elements**: Each milestone has a fixed price. You do not report costs.
- **Cost-type elements**: Rare for small companies in prototype phase.
- **Payment timing**: Government typically pays within 30 days of milestone acceptance. Net-30 is standard.

---

## 9. Success Stories from Small Companies

### Relevant Examples in DevSecOps / Infrastructure / AI Space

**Chainguard (Container Security)**
- Small company focused on secure container base images and software supply chain security
- Engaged with DoD through multiple pathways including DIU-adjacent programs
- Built product around distroless container images, SBOM generation, and vulnerability scanning
- Demonstrates that niche container/Kubernetes security products have DoD demand

**Anchore (Container Compliance)**
- Started as a small company with an open-source container scanning tool
- Won DoD contracts for container compliance and SBOM enforcement
- Their product enforces policies on container images before deployment -- directly relevant to DevSecOps platforms
- Now embedded in Platform One / Iron Bank pipeline

**Second Front Systems (Game Warden)**
- Small company that built a DevSecOps platform (Game Warden) for DoD
- Won DIU prototype and transitioned to production
- Platform provides Kubernetes-based hosting for DoD applications with continuous ATO
- Directly comparable to a hub-and-spoke Kubernetes platform
- Demonstrates that "platform-as-a-service for DoD" is a funded category

**Rancher Government Solutions (Kubernetes)**
- Spun out to focus specifically on DoD Kubernetes deployments
- SUSE Rancher (RKE2/K3s) became the basis for multiple DoD Kubernetes environments
- Shows that Kubernetes distribution/management is a legitimate DoD product category

**Istio/Service Mesh Providers**
- Multiple small companies have won DoD work around service mesh, zero-trust networking, and multi-cluster Kubernetes management
- Tetrate (Istio-based service mesh) engaged with DoD on zero-trust networking

**Rise8 (DevSecOps Consulting + Products)**
- Small company focused on DoD DevSecOps transformation
- Works directly with software factories across military services
- Shows there is demand for both products and expertise in this space

### Key Takeaway

Companies with 5-50 employees have successfully won DIU prototypes and production contracts in the DevSecOps/infrastructure space. Being small is not a disqualifier -- it is often an advantage because:
- You move faster
- Your technology is more focused
- You qualify as nontraditional (enabling OT authority)
- DIU specifically looks for commercial companies that large primes cannot replicate

---

## 10. NSIN and Related Programs

### National Security Innovation Network (NSIN)

**Website**: https://www.nsin.mil

NSIN is a DoD program office under USD(R&E) that connects innovators (startups, universities, hackers) with DoD problems. It complements DIU.

| Aspect | Details |
|--------|---------|
| Mission | Connect national security problems with solutions from non-traditional sources |
| Programs | Hacks, X-Force fellowships, innovation workshops, Propel (accelerator) |
| Relevance to you | Lower barrier of entry than DIU; good for getting on DoD radar |
| Award sizes | Varies; some programs are non-monetary (networking/access) |

**NSIN Programs Worth Exploring:**

1. **NSIN Propel**: An accelerator program for startups with dual-use technology. Provides mentoring, DoD customer introductions, and sometimes funding.
2. **Hacks**: Hackathon-style events where companies/individuals build solutions to DoD problems over a weekend. Good for visibility.
3. **X-Force Fellowship**: Places STEM talent into DoD organizations. Less relevant for a solo founder but good to know about.

### Other Relevant Programs

**AFWERX (Air Force)**
- Website: https://afwerx.com
- Air Force's innovation arm
- Runs SBIR/STTR Open Topics, challenge competitions, and Spark Tank
- AFWERX SBIR Open Topics are particularly accessible for small companies
- Relevant if Air Force is a target customer for your platform

**Army Applications Laboratory (AAL)**
- Army's rapid prototyping organization
- Runs challenges and prototype opportunities
- Focus on AI/ML, autonomy, and software

**NavalX**
- Navy's innovation organization
- Connects commercial tech with Navy/Marine Corps problems
- Runs Tech Bridges (regional innovation hubs)

**CDAO (Chief Digital and Artificial Intelligence Office)**
- Oversees DoD AI strategy
- Manages the Joint AI Center (JAIC) successor programs
- Relevant if your platform enables AI/ML workload management

**Kessel Run (Air Force)**
- Air Force software factory
- Builds and deploys cloud-native applications
- Uses DevSecOps practices heavily
- Could be a customer for underlying platform infrastructure

**Platform One / Party Bus (Air Force)**
- Air Force's enterprise DevSecOps platform
- Provides CI/CD, container registry (Iron Bank), Kubernetes hosting
- Your platform could complement or provide an alternative to Platform One for other services

**SOCOM (Special Operations Command)**
- Has its own innovation office and OT authority
- Often moves faster than conventional forces
- Interested in AI, edge computing, and rapid deployment capabilities

### Strategy: Use NSIN/AFWERX as On-Ramps to DIU

```
NSIN event/hack --> Relationships with DoD problem owners -->
DoD sponsor identifies need --> DIU CSO submission with DoD champion
```

Having a DoD user who says "I need this technology" dramatically increases your chances with DIU.

---

## 11. Getting on DIU's Radar: Contacts, Events, and Strategies

### DIU Events

| Event | What | When | How to Attend |
|-------|------|------|---------------|
| **DIU Demo Days** | Companies demo tech to DoD decision-makers | Periodic (check diu.mil) | Application/invitation |
| **National Defense Industrial Association (NDIA) events** | Industry conferences with DIU participation | Multiple per year | Registration (paid) |
| **TechNet (AFCEA)** | Defense IT conference, DIU often presents | February (West), August (Augusta) | Registration (paid) |
| **Def Con / Black Hat** | Security conferences, DIU scouts here | Summer (Las Vegas) | Registration |
| **AUSA (Association of the US Army)** | Army-focused, DIU often has a booth | October | Registration |
| **SOFIC** | Special operations conference | May (Tampa) | Registration |
| **AWS re:Invent / KubeCon** | Cloud/K8s conferences where DIU scouts | November-December / varies | Registration |
| **RSA Conference** | Cybersecurity, DoD innovation panels | April (San Francisco) | Registration |

### Online Presence

1. **DIU website**: Subscribe to newsletter at diu.mil
2. **LinkedIn**: Follow DIU, NSIN, AFWERX, and individual DIU portfolio directors
3. **SAM.gov**: Set up saved searches for DIU solicitations
4. **GovWin (Deltek)**: Market intelligence on upcoming DoD opportunities (paid, but has free trial)
5. **FPDS.gov**: Search past DIU awards to understand what they have funded

### Building Relationships

**Direct outreach to DIU:**
- DIU welcomes unsolicited inquiries from commercial companies
- Email: info@diu.mil (general inquiries)
- Better: identify the specific portfolio director for your area (Cyber, AI, etc.) via LinkedIn and reach out directly
- DIU staff are generally responsive and used to talking to small companies

**Through intermediaries:**
- **Venture capital firms** that focus on defense tech (e.g., Shield Capital, Lux Capital, a16z defense) often have DIU relationships and can make introductions
- **Defense tech accelerators**: Hacking 4 Defense (H4D), NSIN Propel, Capital Factory (Austin), MassChallenge (Boston)
- **Congressional contacts**: Your congressional representatives can make introductions to DIU. This is legitimate and common.

**Through DoD end users:**
- The strongest path to a DIU award is having a DoD user who wants your technology
- Attend events where military technologists gather
- Participate in NSIN hacks to meet military personnel
- The DoD user can then sponsor your solution through DIU

### Recommended Networking Strategy for a Solo Founder

1. **Month 1**: Register on DIU portal, subscribe to newsletter, follow on LinkedIn. Review all current AOIs.
2. **Month 1-2**: Attend 1-2 NSIN events or virtual hacks. Meet military problem owners.
3. **Month 2-3**: Identify the most relevant current DIU AOI. Prepare submission.
4. **Month 3**: Submit to DIU CSO. Simultaneously reach out to DIU portfolio director on LinkedIn.
5. **Ongoing**: Attend one major defense tech event per quarter (NDIA, TechNet, AUSA, etc.)
6. **Ongoing**: Build relationships with 2-3 defense tech VC firms (even if not raising money, they are connectors)

---

## 12. Registration Checklist

Complete these in order. Items 1-3 should be done immediately. Items 4-7 should be done before award.

| # | Item | Where | Status | Notes |
|---|------|-------|--------|-------|
| 1 | Business entity (LLC/Corp) | State Secretary of State | Required | If not already established |
| 2 | EIN | IRS (irs.gov) | Required | Immediate online, or 4 weeks by mail |
| 3 | DIU portal account | submissions.diu.mil | Required | Free, quick |
| 4 | SAM.gov registration | sam.gov | Before award | 2-4 weeks processing; free |
| 5 | UEI (via SAM.gov) | sam.gov | Before award | Issued as part of SAM registration |
| 6 | CAGE code | Assigned via SAM | Before award | Automatic with SAM registration |
| 7 | Business bank account | Any bank | Before award | For receiving government payments |
| 8 | SBA size certification | sba.gov (optional) | Optional | Confirms small business status |
| 9 | LinkedIn company page | linkedin.com (optional) | Recommended | DIU staff will look you up |
| 10 | Basic website | Your domain (optional) | Recommended | Credibility for reviewers |

---

## 13. Positioning a Kubernetes/DevSecOps Platform for DIU

Based on the Aegis platform's capabilities (hub-and-spoke Kubernetes, multi-cluster management, workload scheduling, DevSecOps, compliance enforcement), here is how to position for DIU.

### Value Propositions That Resonate with DoD

| Capability | DoD Value Proposition |
|------------|----------------------|
| Hub-and-spoke multi-cluster | "Manage workloads across classification boundaries and geographic locations from a single control plane" |
| Kubernetes workload scheduling | "Automated, policy-driven workload placement across heterogeneous clusters" |
| OIDC/Keycloak auth | "Zero-trust identity federation across distributed clusters" |
| Compliance enforcement | "Continuous compliance monitoring mapped to NIST 800-53 / FedRAMP controls" |
| Remote dev environments | "Secure, pre-configured development environments accessible from any authorized device" |
| gRPC agent architecture | "Lightweight spoke agents that work across air-gapped and bandwidth-constrained networks" |
| Helm-based deployment | "GitOps-ready, declarative platform deployment aligned with DoD DevSecOps Reference Design" |

### Mapping to DoD Standards and Frameworks

| Your Feature | DoD Standard/Framework |
|-------------|----------------------|
| Container orchestration | DoD DevSecOps Reference Design |
| Auth/OIDC | DoD Zero Trust Reference Architecture |
| Compliance automation | NIST 800-53 / FedRAMP / DISA STIGs |
| Network policies | DoD Cloud Computing SRG (IL2-IL6) |
| SBOM/supply chain | Executive Order 14028, NIST SSDF |
| Multi-cluster | CJADC2 infrastructure requirements |

### Recommended AOI Targets

When reviewing current DIU AOIs, prioritize those mentioning:

1. DevSecOps platform / software factory infrastructure
2. Multi-cloud / hybrid cloud orchestration
3. Kubernetes management / container orchestration
4. Zero-trust architecture implementation
5. Edge computing platforms
6. AI/ML infrastructure and MLOps
7. Continuous ATO / compliance automation

### Competitive Differentiation

Position against:
- **Platform One**: "We provide a commercially maintained alternative that DoD components can deploy without depending on Air Force infrastructure"
- **Large SI custom builds**: "Our platform is a product, not a custom development project -- faster to deploy, lower total cost of ownership"
- **Hyperscaler-native tools (EKS, GKE, AKS)**: "We provide a cloud-agnostic control plane that works across any Kubernetes distribution, including air-gapped environments"
- **Rancher/OpenShift**: "We focus specifically on workload scheduling and compliance across distributed clusters, not just cluster management"

---

## 14. Key Terminology

| Term | Definition |
|------|-----------|
| **AOI** | Area of Interest -- a specific problem statement within a CSO |
| **ATO** | Authority to Operate -- security authorization required before deploying in DoD environments |
| **cATO** | Continuous ATO -- automated, ongoing security authorization (the goal) |
| **CAGE Code** | Commercial and Government Entity code -- identifies your company in DoD systems |
| **CDAO** | Chief Digital and Artificial Intelligence Office -- DoD AI leadership |
| **CMMC** | Cybersecurity Maturity Model Certification -- required for handling CUI (Controlled Unclassified Information) |
| **CSO** | Commercial Solutions Opening -- DIU's standing solicitation mechanism |
| **CUI** | Controlled Unclassified Information -- sensitive but not classified data |
| **DCAA** | Defense Contract Audit Agency -- audits cost-type contracts (NOT required for OT prototypes) |
| **DFARS** | Defense FAR Supplement -- additional regulations for DoD contracts (NOT applicable to OTs) |
| **FAR** | Federal Acquisition Regulation -- the main body of federal contracting rules (NOT applicable to OTs) |
| **FedRAMP** | Federal Risk and Authorization Management Program -- cloud security authorization |
| **IL** | Impact Level -- DoD data classification levels for cloud (IL2, IL4, IL5, IL6) |
| **NAICS** | North American Industry Classification System -- codes that categorize your business |
| **OT/OTA** | Other Transaction / Other Transaction Authority |
| **PEO** | Program Executive Officer -- senior acquisition leader who owns production budgets |
| **PM** | Program Manager -- military/civilian who manages a specific program |
| **SAM** | System for Award Management -- federal contractor registration database |
| **SBOM** | Software Bill of Materials -- inventory of software components |
| **SBIR** | Small Business Innovation Research -- federal R&D funding for small businesses |
| **SRG** | Security Requirements Guide -- DISA security standards |
| **STIG** | Security Technical Implementation Guide -- specific technical security configurations |
| **TRL** | Technology Readiness Level -- 1 (basic research) to 9 (proven in operations) |
| **UEI** | Unique Entity ID -- identifies your business in SAM.gov (replaced DUNS number) |

---

## Appendix: Quick Action Plan

### Week 1
- [ ] Register on DIU submission portal
- [ ] Subscribe to DIU newsletter
- [ ] Follow DIU, NSIN, AFWERX on LinkedIn
- [ ] Review all current AOIs on diu.mil
- [ ] Begin SAM.gov registration (takes 2-4 weeks)

### Week 2-3
- [ ] Identify the 2-3 most relevant current AOIs
- [ ] Draft a 5-page solution brief for your strongest AOI match
- [ ] Research the specific DoD component that would use your technology
- [ ] Identify DIU portfolio director for your area on LinkedIn

### Week 4-6
- [ ] Submit to DIU CSO portal
- [ ] Send a concise LinkedIn message to the relevant DIU portfolio director
- [ ] Register for the next relevant NSIN event or defense tech conference
- [ ] Explore AFWERX SBIR Open Topics as a parallel path

### Month 2-3
- [ ] Follow up on DIU submission if no response
- [ ] Attend first defense tech event
- [ ] Begin building relationships with 2-3 DoD end users in your problem space
- [ ] Consider applying to NSIN Propel or similar accelerator

### Month 3-6
- [ ] If DIU pitch meeting: prepare 15-minute demo, 15-minute Q&A
- [ ] If no DIU traction: submit to AFWERX or direct to a service software factory
- [ ] Continue attending events and building network
- [ ] Start FedRAMP/STIG documentation preparation (long lead time)

---

## Appendix: Useful Links

| Resource | URL |
|----------|-----|
| DIU Main Site | https://www.diu.mil |
| DIU Work With Us | https://www.diu.mil/work-with-us |
| DIU Portfolio Areas | https://www.diu.mil/portfolios |
| NSIN | https://www.nsin.mil |
| AFWERX | https://afwerx.com |
| SAM.gov | https://sam.gov |
| SBA.gov | https://sba.gov |
| FedRAMP | https://www.fedramp.gov |
| DoD DevSecOps Reference Design | Search for "DoD Enterprise DevSecOps Reference Design" on dodcio.defense.gov |
| DoD Zero Trust Reference Architecture | Search on dodcio.defense.gov |
| Iron Bank (approved container images) | https://ironbank.dso.mil |
| DISA STIGs | https://public.cyber.mil/stigs |

---

*This guide is based on publicly available information about DIU and OTA processes as of early 2026. Specific AOIs, contacts, and program details change frequently. Always verify against current diu.mil postings before submitting.*
