# SBIR/STTR Grant Guide for a 1-Person Small Business (2025-2026)

**Focus Area**: AI/ML Infrastructure, GPU Scheduling, Secure Computing, DevSecOps, Kubernetes for DoD

**Last Updated**: March 2026

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Eligibility Requirements](#2-eligibility-requirements)
3. [Registration Requirements](#3-registration-requirements)
4. [Target Agencies for AI/ML and DevSecOps](#4-target-agencies-for-aiml-and-devsecops)
5. [Phase I Specifics](#5-phase-i-specifics)
6. [Phase II Specifics](#6-phase-ii-specifics)
7. [Open Topics and BAAs of Interest](#7-open-topics-and-baas-of-interest)
8. [Common Pitfalls for First-Time Applicants](#8-common-pitfalls-for-first-time-applicants)
9. [Timeline: Application to Award](#9-timeline-application-to-award)
10. [Simplified Submission Options](#10-simplified-submission-options)
11. [Can a 1-Person Company Realistically Win?](#11-can-a-1-person-company-realistically-win)
12. [Action Plan: Step-by-Step](#12-action-plan-step-by-step)
13. [Resources and Links](#13-resources-and-links)

---

## 1. Executive Summary

SBIR (Small Business Innovation Research) and STTR (Small Business Technology Transfer) are federal programs that collectively award over **$4 billion annually** to small businesses for R&D. Eleven federal agencies participate. For a company building AI/ML infrastructure, GPU scheduling, and DevSecOps tools for Kubernetes -- the Department of Defense (DoD) agencies are the primary targets.

**Key takeaway for a 1-person company**: Yes, solo founders win SBIR Phase I awards regularly. The program was designed for small innovative companies. However, you must be structured correctly (entity type, registrations) and your proposals must be technically excellent and demonstrate feasibility.

**SBIR vs. STTR difference**: SBIR requires the small business to perform at least 2/3 of the Phase I work. STTR *requires* a formal partnership with a U.S. research institution (university, FFRDC, or nonprofit research org) and the small business must perform at least 40% of the work while the research institution performs at least 30%. For a solo founder without an existing university partnership, **SBIR is the easier path**.

---

## 2. Eligibility Requirements

### 2.1 Entity Type

| Entity Type | SBIR Eligible? | Notes |
|-------------|---------------|-------|
| C-Corporation | Yes | Preferred by many SBIR recipients; cleanest structure for equity, future investment |
| S-Corporation | Yes | Works fine; pass-through taxation |
| LLC | Yes | Must be organized for profit; single-member LLCs are eligible |
| Sole Proprietorship | Yes | Eligible but less common; harder to separate personal/business liability |
| Partnership | Yes | Less common |

**Recommendation**: An **LLC** (single-member, taxed as S-Corp if desired) is the simplest starting structure. A **C-Corp** is better if you plan to raise venture capital later or want the cleanest separation. Either works for SBIR.

### 2.2 Size and Employee Requirements

- **Fewer than 500 employees** (including affiliates) at the time of award
- A 1-person company absolutely qualifies
- Part-time employees, contractors, and consultants count differently:
  - **Employees** (W-2) count toward the 500
  - **Contractors/consultants** (1099) do NOT count toward the 500 but their work counts toward the percentage-of-work requirements
- No minimum employee count

### 2.3 Ownership Rules

- The business must be **more than 50% owned and controlled by one or more individuals who are U.S. citizens or permanent resident aliens**
- Alternatively, the business can be more than 50% owned by other small businesses that themselves meet the ownership criteria
- Venture capital, hedge fund, or private equity ownership: companies majority-owned by VCs/PEs are eligible for SBIR **only at agencies that have opted in** to the 2011 SBIR/STTR Reauthorization Act provisions (DoD, NIH, DOE, NSF, and NASA have opted in -- but extra rules apply and disclosure is required)
- For a solo founder who is a U.S. citizen: you meet all ownership requirements trivially

### 2.4 U.S. Citizenship / Residency

- **PI (Principal Investigator)**: Must be "primarily employed" by the small business. For SBIR, the PI must spend more than 50% of their time with the company (more than one-half of their working time). There is no explicit citizenship requirement for the PI *per se*, but the PI must be legally authorized to work in the U.S.
- **Company ownership**: More than 50% U.S. citizens or permanent residents
- **Place of performance**: Research must be performed in the United States
- **STTR co-PI**: Can be at the research institution

### 2.5 Additional Requirements

- Must be a for-profit business (not nonprofit)
- Must be independently operated (not a subsidiary or division of a larger entity in a way that violates affiliation rules)
- Must have a place of performance in the United States
- Cannot have the PI primarily employed elsewhere (a common disqualifier -- if you have a full-time W-2 job, you technically cannot be the SBIR PI unless you leave or go part-time at the other job such that the SBIR company is your primary employment)

**CRITICAL**: If you currently have a full-time job elsewhere, you must make the SBIR company your primary employer before the award is made. Many agencies verify this. Some agencies require certification at proposal time; others at award time. Plan accordingly.

---

## 3. Registration Requirements

Registration is the longest lead-time item. **Start registrations immediately** -- some take 2-8 weeks.

### 3.1 Required Registrations (in order)

| Step | System | URL | Timeline | Purpose |
|------|--------|-----|----------|---------|
| 1 | **Get a DUNS/UEI Number** | sam.gov | 1-2 days | Unique Entity Identifier; replaced DUNS in April 2022. Generated during SAM.gov registration |
| 2 | **Register in SAM.gov** | https://sam.gov | 2-8 weeks (can take longer) | System for Award Management; mandatory for ALL federal contracts and grants |
| 3 | **Register in SBIR.gov** | https://www.sbir.gov | 1 day (after SAM) | Company registry; used by all SBIR/STTR agencies |
| 4 | **Register in SBA Company Registry** | https://sbir.gov/registration | 1 day | Part of SBIR.gov; certifies your small business status |
| 5 | **Create Login.gov account** | https://login.gov | 15 minutes | Authentication for most federal systems |
| 6 | **DSIP (Defense SBIR/STTR Innovation Portal)** | https://www.dodsbirsttr.mil | 1 day | DoD-specific submission portal (formerly DoD SBIR/STTR portal) |
| 7 | **Grants.gov** (if applying to non-DoD) | https://grants.gov | 1-3 days | Some agencies (NSF, DOE) use Grants.gov for submission |
| 8 | **Agency-specific portals** | Varies | 1-2 days each | Some agencies have their own portals |

### 3.2 SAM.gov Registration Details

SAM.gov is the critical path item. Here is what you need:

**Required information**:
- Legal business name and DBA
- Business address (must be a U.S. address; PO Boxes are discouraged)
- EIN (Employer Identification Number) from the IRS -- get this first if you do not have it (Form SS-4, online at irs.gov, takes minutes)
- Bank account information (for EFT payments)
- NAICS codes (see below)
- CAGE code (assigned automatically during SAM registration)
- Goods and services you provide (use PSC codes)
- Business type and size information

**Relevant NAICS codes for your technology area**:
- **541511** - Custom Computer Programming Services
- **541512** - Computer Systems Design Services
- **541519** - Other Computer Related Services
- **541715** - Research and Development in the Physical, Engineering, and Life Sciences
- **518210** - Data Processing, Hosting, and Related Services
- **511210** - Software Publishers

**Tips**:
- SAM registration must be renewed annually
- Start this process BEFORE you identify a specific solicitation
- If you get stuck, call the Federal Service Desk: 866-606-8220
- The system can be finicky; use Chrome; save frequently

### 3.3 SBA Company Registry / SBIR.gov

After SAM.gov is active:
1. Go to https://www.sbir.gov
2. Create a company profile
3. Enter your UEI, SAM information
4. Certify that you meet small business size standards
5. This registration is checked by agencies before award

### 3.4 DSIP (DoD Portal)

The **Defense SBIR/STTR Innovation Portal (DSIP)** at https://www.dodsbirsttr.mil is where you:
- Browse DoD SBIR/STTR solicitations (all services)
- Submit proposals to DoD agencies (Air Force, Navy, Army, DARPA, MDA, CDAO, etc.)
- Check proposal status
- Manage awards

**Registration steps**:
1. Create an account at dodsbirsttr.mil
2. Link your company (UEI and SBIR.gov registration must be complete)
3. Verify PI information
4. Complete any agency-specific forms

---

## 4. Target Agencies for AI/ML and DevSecOps

For a company building Kubernetes-based AI/ML infrastructure, GPU scheduling, and DevSecOps tooling, the following DoD agencies are the highest-value targets.

### 4.1 Air Force Research Laboratory (AFRL) / Air Force SBIR

**Why target**: AFRL is the most SBIR-friendly DoD component. They fund heavily in AI/ML, cloud infrastructure, and DevSecOps. They pioneered the **SBIR Open Topic** (now AF SBIR Open Topic) and **AFWERX** programs, which are more startup-friendly.

| Item | Detail |
|------|--------|
| Portal | https://www.dodsbirsttr.mil (submit here) |
| Also see | https://www.afwerx.com (AFWERX programs) |
| Open Topic URL | https://www.afwerx.com/sbir-sttr |
| Typical Phase I | $50,000 - $250,000 |
| Phase I Duration | 3-12 months (Open Topic is often 3-6 months) |
| Submission | Rolling (Open Topic) or periodic (traditional) |
| Evaluation | Technical merit, PI qualifications, commercialization potential |

**AFWERX Open Topic**: This is probably your **best first target**. Key features:
- Rolling submissions (no fixed deadlines -- though they batch reviews)
- Shorter proposals (often 5-10 pages vs 20+ for traditional SBIR)
- Faster decisions (weeks to a few months, not 6+ months)
- Emphasis on dual-use commercial technology
- Direct connection to Air Force end users
- Phase I is typically $50K-$75K for a short feasibility study

### 4.2 DARPA (Defense Advanced Research Projects Agency)

**Why target**: DARPA funds cutting-edge technology. They have programs specifically around AI infrastructure, secure computing, and software engineering automation.

| Item | Detail |
|------|--------|
| Portal | https://www.dodsbirsttr.mil |
| Also see | https://www.darpa.mil/work-with-us/opportunities |
| DARPA SBIR page | https://www.darpa.mil/work-with-us/for-small-businesses |
| Typical Phase I | $100,000 - $250,000 |
| Phase I Duration | 6-14 months |
| Submission | Per solicitation (DARPA publishes specific topic-driven SBIRs) |
| Evaluation | Technical innovation, potential impact, team capability |

**Relevant DARPA programs (past and likely recurring themes)**:
- Software-defined infrastructure
- Autonomous systems computing
- Secure computing architectures
- AI/ML pipeline automation

**DARPA's approach**: More technically demanding proposals. They want paradigm-shifting innovation, not incremental improvements. If you can frame your Kubernetes AI/ML platform as a novel capability (e.g., autonomous secure workload orchestration), DARPA is interested.

### 4.3 CDAO (Chief Digital and Artificial Intelligence Office)

**Why target**: CDAO (successor to JAIC -- Joint AI Center) is the DoD's central AI/ML office. They specifically need AI infrastructure, MLOps, and data platforms.

| Item | Detail |
|------|--------|
| Portal | https://www.dodsbirsttr.mil |
| CDAO site | https://www.ai.mil |
| Typical Phase I | $100,000 - $250,000 |
| Solicitation | Through DoD SBIR solicitations; CDAO topics appear in DoD-wide SBIR releases |

**Relevant topics from CDAO**:
- AI/ML model deployment infrastructure
- MLOps pipelines
- Secure AI workload orchestration
- Data mesh / data fabric
- GPU resource management for AI training

### 4.4 Navy (NAVAIR, NAVSEA, ONR, NIWC)

**Why target**: The Navy has substantial SBIR budgets and specific needs for cloud-native infrastructure, cybersecurity, and edge computing.

| Item | Detail |
|------|--------|
| Portal | https://www.dodsbirsttr.mil |
| Navy SBIR site | https://www.navysbir.com |
| Typical Phase I | $140,000 - $250,000 |
| Phase I Duration | 6 months |
| Submission | DoD-wide solicitation cycles |

**Navy-specific interests**:
- Black Pearl (Navy's DevSecOps platform -- a Platform One equivalent)
- Overmatch (Naval networking initiative)
- Cloud infrastructure for tactical edge
- Cybersecurity tools

### 4.5 Army (AFC, DEVCOM, PEO)

**Why target**: The Army has a growing need for AI/ML infrastructure, especially through Army Futures Command.

| Item | Detail |
|------|--------|
| Portal | https://www.dodsbirsttr.mil |
| Army SBIR site | https://www.armysbir.army.mil |
| Typical Phase I | $50,000 - $250,000 |
| Phase I Duration | 6 months |
| Submission | DoD-wide solicitation cycles |

**Army-specific interests**:
- cArmy (Army cloud initiative)
- AI/ML for logistics and maintenance
- DevSecOps (Army Software Factory)
- Network modernization

### 4.6 Non-DoD Agencies Worth Considering

| Agency | Why | Phase I Amount | Portal |
|--------|-----|---------------|--------|
| **DHS** (S&T) | Cybersecurity, critical infrastructure | $100K-$200K | dodsbirsttr.mil or grants.gov |
| **NSF** | America's Seed Fund; very startup-friendly | $275K | https://seedfund.nsf.gov |
| **DOE** | HPC, GPU computing, scientific infrastructure | $200K-$250K | https://science.osti.gov/sbir |
| **NASA** | Cloud infrastructure, ML pipelines | $150K | https://sbir.nasa.gov |

**NSF SBIR (America's Seed Fund)** deserves special mention:
- Very startup-friendly
- Phase I is $275K for 12 months
- No specific topic -- it is an "open" program (your innovation must fit a broad technology area)
- Higher success rate than DoD SBIR for commercial technology
- Submission via https://seedfund.nsf.gov
- Three submission windows per year (roughly June, November, March -- check current dates)
- Does NOT require government end-use -- purely commercial innovations qualify

---

## 5. Phase I Specifics

### 5.1 Award Amounts by Agency

| Agency | Phase I Amount | Duration |
|--------|---------------|----------|
| DoD (traditional) | $50,000 - $250,000 | 6-12 months |
| DoD (AFWERX Open Topic) | $50,000 - $75,000 | 3-6 months |
| DoD (DARPA) | $100,000 - $250,000 | 6-14 months |
| NSF | $275,000 | 6-12 months |
| DOE | $200,000 - $250,000 | 6-12 months |
| NASA | $150,000 | 6-13 months |
| NIH | $293,498 (FY2025 guideline) | 6-12 months |
| DHS | $100,000 - $200,000 | 6 months |

### 5.2 What Phase I Requires

Phase I is a **feasibility study**. You are proving that your concept *can* work, not delivering a finished product.

**Deliverables typically include**:
1. **Technical report** (15-50 pages depending on agency): Describes research performed, technical approach, results, feasibility assessment
2. **Phase II proposal preparation**: Most agencies expect you to submit a Phase II proposal near the end of Phase I
3. **Prototype/demo** (sometimes): Some agencies want a working demo or prototype, others accept a paper study
4. **Commercialization plan**: How you will bring this to market (government and commercial)
5. **Financial reports**: Invoices, cost accounting

**What you are NOT expected to deliver**:
- A production-ready product
- Revenue or customers
- A large team

### 5.3 Evaluation Criteria

Most DoD agencies score on three dimensions (SBA-mandated):

| Criterion | Weight (typical) | What They Look For |
|-----------|------------------|-------------------|
| **Technical Merit** | ~40-50% | Innovation, technical approach soundness, understanding of the problem |
| **Qualifications of PI/Team** | ~20-30% | Relevant experience, publications, past performance, technical skills |
| **Commercialization Potential** | ~20-30% | Market size, dual-use potential, transition plan, business model |

**For a solo founder**: Emphasize your direct technical expertise. Include your GitHub contributions, prior work, publications, or deployed systems as evidence of qualification. A strong PI with deep domain expertise can outscore a larger team with weaker technical depth.

### 5.4 Proposal Structure (DoD Traditional)

Typical DoD SBIR Phase I proposal structure:
1. **Cover Page** (generated by DSIP portal)
2. **Technical Volume** (usually 20-25 page limit)
   - Identification and Significance of the Problem
   - Technical Approach / Phase I Work Plan
   - Related Work
   - Key Personnel and Facilities
   - Schedule and Milestones
3. **Cost Volume** (budget form + narrative)
4. **Company Commercialization Report (CCR)** -- auto-generated from your SBIR.gov profile
5. **Supporting Documents** (resumes, letters of support, relevant papers)

---

## 6. Phase II Specifics

### 6.1 Transition from Phase I to Phase II

- **Invitation-based**: You typically receive a Phase II invitation only if Phase I was successful
- **Competitive**: Not all Phase I winners get Phase II; it is a separate competitive evaluation
- **Timeline**: Phase II proposals are usually due near the end of Phase I (agencies give 30-90 days notice)
- **Evaluation**: More emphasis on prototype readiness, commercialization potential, and transition to the warfighter/customer

### 6.2 Phase II Award Amounts

| Agency | Phase II Amount | Duration |
|--------|----------------|----------|
| DoD (traditional) | $500,000 - $1,700,000 | 24 months |
| DoD (AFWERX) | $750,000 - $1,500,000 | 12-27 months |
| NSF | $1,000,000 | 24 months |
| DOE | $1,000,000 - $1,600,000 | 24 months |
| NASA | $850,000 | 24 months |

### 6.3 Phase II Expectations

- **Working prototype** or beta-level software
- **Demonstration** to end users / program office
- **Commercialization progress**: Letters of intent, pilot customers, revenue pipeline
- **Transition plan**: How the DoD will acquire/use the technology after SBIR funding ends

### 6.4 Beyond Phase II

| Program | Description |
|---------|-------------|
| **Phase III** | Non-SBIR federal funding to deploy the technology; sole-source contracting is allowed |
| **Phase II Enhancement (Phase II-E)** | Additional funding (some agencies match $1 of SBIR with $1 of outside investment, up to $500K) |
| **Sequential Phase II** | Second Phase II award for continued development |
| **Direct to Phase II (D2P2)** | Some agencies allow skipping Phase I if you can demonstrate feasibility with prior work or internal R&D |

---

## 7. Open Topics and BAAs of Interest

### 7.1 Where to Find Relevant Solicitations

| Source | URL | What It Has |
|--------|-----|-------------|
| DSIP Topic Search | https://www.dodsbirsttr.mil/topics-702/ | All open DoD SBIR/STTR topics |
| SBIR.gov | https://www.sbir.gov/solicitations/open | All agencies' open solicitations |
| AFWERX | https://www.afwerx.com/sbir-sttr | Air Force Open Topics |
| SAM.gov | https://sam.gov/content/opportunities | BAAs and contract opportunities |
| Beta.SAM.gov | Search for "BAA" + keywords | Broad Agency Announcements |
| Grants.gov | https://grants.gov | Non-DoD grant opportunities |

### 7.2 Keywords to Search

When searching for relevant topics, use combinations of:

**Primary keywords**:
- Kubernetes, container orchestration, cloud-native
- DevSecOps, CI/CD, software factory
- AI/ML infrastructure, MLOps
- GPU scheduling, GPU resource management
- Multi-cluster management, workload orchestration
- Platform One, Big Bang, Iron Bank
- Zero-trust architecture
- Software-defined infrastructure
- Secure computing environment
- Development environment, IDE, workspace

**Secondary keywords**:
- CNCF, service mesh, Istio, Envoy
- Container security, supply chain security
- Infrastructure as Code (IaC)
- Edge computing infrastructure
- ATO (Authority to Operate) automation
- STIG compliance automation
- cATO (continuous ATO)

### 7.3 Relevant Topic Areas (Recurring Themes)

These topic areas appear regularly in DoD SBIR solicitations:

1. **DevSecOps Pipeline Automation**: Tools that automate CI/CD with security built in. Direct alignment with Platform One / Big Bang ecosystem.

2. **AI/ML Workload Orchestration**: Scheduling and managing GPU-intensive AI/ML training and inference workloads across distributed infrastructure.

3. **Secure Multi-Tenant Computing**: Isolated computing environments for classified/sensitive workloads. Your multi-cluster Kubernetes platform maps directly here.

4. **Cloud-Native Application Security**: Zero-trust networking, container runtime security, supply chain integrity for containerized applications.

5. **Software Factory Tools**: The DoD is actively investing in software factories (Kessel Run, Army Software Factory, Navy DevSecOps). Tools that enhance developer productivity in these environments are in demand.

6. **Edge AI Infrastructure**: Deploying AI/ML models to tactical edge environments with resource constraints. Kubernetes-based orchestration at the edge.

7. **ATO/Compliance Automation**: Automating the Authority to Operate process, continuous compliance monitoring, STIG scanning.

### 7.4 Direct to Phase II (D2P2) Opportunity

Several DoD components offer **Direct to Phase II**, which lets you skip Phase I if you can demonstrate that you have already completed feasibility work equivalent to a Phase I. Given that you have a working platform (Aegis), this could be a strong option:

- **AFWERX D2P2**: Available through the Open Topic
- **Navy D2P2**: Periodically available
- **DARPA D2P2**: By invitation or specific topics

For D2P2, you must demonstrate:
- Prior R&D equivalent to Phase I (your existing codebase and deployments count)
- Technical feasibility has been established
- Clear plan for Phase II prototype development

---

## 8. Common Pitfalls for First-Time Applicants

### 8.1 Administrative Disqualifications (Instant Rejection)

These will get your proposal rejected without technical review:

1. **SAM.gov registration not active** at time of submission or award
2. **Missing or incomplete SBA Company Registry** registration
3. **Exceeding page limits** -- even by one page
4. **Wrong font/margin formatting** -- each solicitation specifies formatting rules (often 10-12pt, 1-inch margins)
5. **Late submission** -- even by one minute. The DSIP portal closes at the exact deadline time
6. **PI not primarily employed** by the company
7. **Missing cost volume** or cost volume errors
8. **Not following the topic's specific requirements** -- each topic may have unique submission instructions

### 8.2 Technical Proposal Weaknesses

9. **Too much about your existing product, not enough about the research**: SBIR funds *research*, not productization. Frame your work as answering a research question, not just building features.
10. **Not addressing the specific topic**: Generic "we have a great platform" proposals lose. Tailor every proposal to the specific topic's stated needs.
11. **No clear innovation**: What is novel? What does your approach do that existing solutions (Platform One, Rancher, OpenShift) cannot?
12. **Weak or no technical milestones**: Reviewers want a clear work plan with measurable milestones, not vague promises.
13. **Ignoring the evaluation criteria**: Read the solicitation's evaluation criteria and structure your proposal to address each one explicitly.

### 8.3 Business/Commercialization Weaknesses

14. **No commercialization plan**: Even for Phase I, you need a credible path to market.
15. **No letters of support**: Letters from potential DoD customers, industry partners, or end users significantly strengthen your proposal.
16. **Empty CCR (Company Commercialization Report)**: If you have no prior SBIR history, your CCR will be thin. Compensate by including strong commercialization narratives in your proposal.
17. **Unrealistic budget**: Do not under-bid or over-bid. Phase I budgets should be justified line by line.

### 8.4 Strategic Mistakes

18. **Only applying to one topic**: Apply to 3-5 relevant topics per cycle. Success rates are 15-25%; volume matters.
19. **Not networking with program officers**: Before submitting, try to talk to the Topic Author (the government person who wrote the topic). They can clarify what they really want. DSIP sometimes lists TPOC (Technical Point of Contact) information.
20. **Ignoring the AFWERX Open Topic**: It is the easiest entry point for first-time SBIR applicants.
21. **Waiting for the "perfect" solicitation**: Submit to the best available topics now. You can always apply again.
22. **Not budgeting for indirect costs**: You need an accounting system that tracks direct vs. indirect costs. Consider setting up a simple cost accounting system (even a spreadsheet) before your first award.

---

## 9. Timeline: Application to Award

### 9.1 DoD Traditional SBIR Timeline

| Phase | Timeline |
|-------|----------|
| Pre-Solicitation announced | ~30 days before open |
| Solicitation opens | Day 0 |
| Submission deadline | 30-60 days after open |
| Review period | 2-4 months after close |
| Selection notifications | 4-6 months after close |
| Contract negotiation | 1-3 months after selection |
| Contract start (award) | 6-9 months after submission |

**Total from submission to money in hand**: **6-12 months** (DoD traditional)

### 9.2 AFWERX Open Topic Timeline

| Phase | Timeline |
|-------|----------|
| Submission | Rolling (batched for review) |
| Review | 2-8 weeks after batch close |
| Selection | 4-12 weeks after submission |
| Award | 2-4 months after submission |

**Total from submission to money**: **2-6 months** (significantly faster)

### 9.3 NSF Timeline

| Phase | Timeline |
|-------|----------|
| Submission window | 3 per year (check seedfund.nsf.gov) |
| Review | 3-5 months |
| Award | 6-8 months after submission |

### 9.4 Typical DoD Solicitation Windows

DoD traditionally releases SBIR solicitations in **3 cycles per year**:

| Cycle | Pre-Release | Open Period | Close |
|-------|-------------|------------|-------|
| Cycle 1 | ~November | ~January | ~February |
| Cycle 2 | ~March | ~April | ~May |
| Cycle 3 | ~July | ~August | ~September |

Note: These are approximate and have shifted in recent years. The DoD has been moving toward more continuous/rolling submissions, especially through AFWERX. Always check https://www.dodsbirsttr.mil for current dates.

---

## 10. Simplified Submission Options

### 10.1 AFWERX Open Topic (Simplified)

The **AFWERX SBIR Open Topic** is the closest thing to a "simplified SBIR":
- Shorter proposals (often a pitch + 5-10 page white paper instead of 20+ pages)
- Rolling submissions
- Focus on commercial viability
- Often structured as a two-step process:
  1. Submit a brief pitch / abstract
  2. If selected, submit a full (but still shorter) proposal

### 10.2 Navy SBIR Direct to Phase II

The Navy sometimes offers streamlined D2P2 submissions for companies with demonstrated prior work.

### 10.3 NSF I-Corps

While not SBIR itself, NSF **I-Corps** ($50K, 7 weeks) can be a precursor to an NSF SBIR submission and helps you develop your commercialization story through customer discovery.

### 10.4 There Is No "SBIR-EZ" Program

Despite occasional rumors, there is **no official "SBIR-EZ" program** as of 2025-2026. However, several agencies have simplified their processes:
- AFWERX has the most streamlined process
- NSF's online submission system is relatively straightforward
- Some agencies accept video pitches as part of the submission

### 10.5 Pitch Days and Innovation Events

Several DoD organizations hold pitch events where you can pitch in person (or virtually) and receive on-the-spot awards:
- **AFWERX Pitch Days**: Rapid evaluation and award; sometimes contracts are signed the same day
- **Army xTech**: Competition-style events with SBIR tie-ins
- **NavalX**: Navy innovation events
- **SOFWERX**: Special Operations Command innovation events

These events are excellent for first-time applicants because they provide direct feedback and networking opportunities.

---

## 11. Can a 1-Person Company Realistically Win?

### 11.1 Short Answer

**Yes.** Solo founders and very small companies (1-3 people) win SBIR Phase I awards regularly. In fact, a significant percentage of first-time SBIR winners are very small companies.

### 11.2 Success Rates

| Context | Approximate Success Rate |
|---------|------------------------|
| DoD SBIR overall | 15-25% of proposals receive Phase I |
| AFWERX Open Topic | 20-30% (higher than traditional) |
| NSF SBIR | 20-25% |
| First-time applicants | ~15% (slightly lower than experienced firms) |
| Experienced SBIR firms | ~25-30% |

**Volume matters**: If you submit 5 well-written proposals per year across different agencies/topics, the probability of winning at least one Phase I in a given year is quite high (50-75%).

### 11.3 Advantages of a 1-Person Company

- **Low overhead**: Your cost structure is favorable; more money goes to direct R&D
- **Technical depth**: If you are the PI and the developer, you have unmatched knowledge of the technology
- **Agility**: You can pivot faster than larger companies
- **Authenticity**: Reviewers can see that you actually do the work (no "bait and switch" with a star PI who does not actually work on the project)

### 11.4 Disadvantages and How to Mitigate

| Disadvantage | Mitigation |
|-------------|------------|
| No past SBIR performance | Highlight commercial traction, open source contributions, or prior relevant work |
| Thin team | Use consultants/subcontractors to fill gaps (up to 33% of Phase I work can be subcontracted in SBIR) |
| No indirect cost rate | Use a simple cost accounting approach; some agencies accept a de minimis indirect rate |
| Perceived risk | Include strong technical milestones and risk mitigation in your proposal |
| No DCAA-compliant accounting | Not required for Phase I; set up basic job costing. Full DCAA compliance is a Phase II concern |

### 11.5 Cost Accounting Note

SBIR Phase I does not require a DCAA-audited accounting system. However, you do need:
- A separate bank account for the business
- Ability to track time spent on the project
- Ability to categorize expenses as direct vs. indirect
- Basic record-keeping (invoices, receipts, timesheets)

For Phase II and beyond, you will want a more robust accounting system. Consider tools like QuickBooks with job costing, or specialized government contract accounting software.

---

## 12. Action Plan: Step-by-Step

### Immediate Actions (Weeks 1-2)

- [ ] **Form your entity** (if not already done): LLC or C-Corp in your state
- [ ] **Get an EIN** from the IRS (online, instant)
- [ ] **Open a business bank account**
- [ ] **Start SAM.gov registration** (this is the longest lead time -- start NOW)

### Short-Term Actions (Weeks 2-4)

- [ ] **Complete SAM.gov registration** (follow up if delayed; call 866-606-8220)
- [ ] **Register on SBIR.gov** (Company Registry)
- [ ] **Create accounts on**:
  - Login.gov
  - DSIP (dodsbirsttr.mil)
  - Grants.gov (for NSF/DOE)
  - NSF Research.gov
- [ ] **Browse current solicitations** on DSIP and SBIR.gov
- [ ] **Identify 3-5 target topics** aligned with your technology

### Medium-Term Actions (Weeks 4-8)

- [ ] **Contact TPOCs** (Technical Points of Contact) for your target topics -- ask questions, understand their needs
- [ ] **Write your first proposal** (start with AFWERX Open Topic if available -- it is the shortest/simplest)
- [ ] **Develop a 1-page commercialization summary**:
  - Total addressable market
  - Government customers (DoD programs that need your technology)
  - Commercial customers
  - Competitive landscape
  - Your unique value proposition
- [ ] **Get letters of support** from potential users, partners, or DoD contacts if possible
- [ ] **Set up basic cost accounting** (spreadsheet or QuickBooks with project tracking)

### Ongoing Actions

- [ ] **Submit to every relevant topic** in each solicitation cycle (aim for 3-5 proposals per year)
- [ ] **Attend SBIR/STTR events**: AFWERX pitch days, SBIR Road Tours, agency industry days
- [ ] **Network**: Join SBIR-related LinkedIn groups; attend National SBIR/STTR Conference
- [ ] **Track deadlines** using a calendar with alerts
- [ ] **After your first win**: Set up proper timekeeping, prepare for Phase II planning

### Proposal Writing Resources

| Resource | URL |
|----------|-----|
| SBIR.gov proposal writing tips | https://www.sbir.gov/tutorials |
| DoD SBIR proposal instructions | Included in each solicitation document |
| SBA SBIR Road Tour (virtual) | https://www.sba.gov/events (search for SBIR) |
| National SBIR/STTR Conference | https://www.sbirsttrc.org (annual, typically in fall) |
| Dawnbreaker (free resources) | https://www.dawnbreaker.com/resources |

---

## 13. Resources and Links

### Primary Portals

| Resource | URL |
|----------|-----|
| SBIR.gov (all agencies) | https://www.sbir.gov |
| DoD DSIP Portal | https://www.dodsbirsttr.mil |
| AFWERX (Air Force) | https://www.afwerx.com |
| NSF SBIR (Seed Fund) | https://seedfund.nsf.gov |
| Navy SBIR | https://www.navysbir.com |
| Army SBIR | https://www.armysbir.army.mil |
| NASA SBIR | https://sbir.nasa.gov |
| DOE SBIR | https://science.osti.gov/sbir |

### Registration

| System | URL |
|--------|-----|
| SAM.gov | https://sam.gov |
| Login.gov | https://login.gov |
| Grants.gov | https://grants.gov |
| SBA Company Registry | https://sbir.gov/registration |

### DoD DevSecOps Context (Understand Your Customer)

| Resource | URL |
|----------|-----|
| Platform One | https://p1.dso.mil |
| Iron Bank (hardened containers) | https://ironbank.dso.mil |
| Big Bang (reference architecture) | https://repo1.dso.mil/big-bang/bigbang |
| DoD Enterprise DevSecOps Reference Design | Search for "DoD Enterprise DevSecOps Reference Design" on dodcio.defense.gov |
| DoD Software Modernization Strategy | https://www.defense.gov (search) |
| cATO information | https://software.af.mil/training/cato/ |

### Positioning Your Technology

When writing proposals, frame your Aegis platform in terms the DoD understands:

| Your Capability | DoD Frame |
|----------------|-----------|
| Multi-cluster Kubernetes management | Federated DevSecOps infrastructure |
| GPU workload scheduling | AI/ML training infrastructure for the warfighter |
| Hub-and-spoke architecture | Distributed computing for contested/degraded environments |
| Keycloak/OIDC auth | Zero-trust identity and access management |
| Workspace provisioning | Secure development environments (Software Factory) |
| Helm/IaC deployment | Infrastructure as Code for ATO automation |
| Network policies | Zero-trust microsegmentation |
| Multi-tenant isolation | Multi-classification-level computing |

---

## Appendix A: Budget Template for Phase I ($75,000 AFWERX Example)

| Category | Amount | Notes |
|----------|--------|-------|
| **Direct Labor** | $45,000 | PI at ~$75/hr x 600 hrs (~4 months full-time) |
| **Fringe Benefits** | $9,000 | ~20% of labor (self-employment tax, health, etc.) |
| **Materials/Supplies** | $3,000 | Cloud compute, test infrastructure |
| **Travel** | $3,000 | 1-2 trips for meetings with government stakeholders |
| **Subcontractors** | $10,000 | Specialist consultant for security/compliance review |
| **Indirect Costs** | $5,000 | ~10% overhead (office, internet, tools) |
| **Total** | **$75,000** | |

Note: Adjust rates to your actual costs. Agencies want to see realistic, justified budgets. Under-bidding looks suspicious; over-bidding wastes review capital.

---

## Appendix B: Proposal Checklist

Before submitting any SBIR proposal, verify:

- [ ] SAM.gov registration is ACTIVE (not expired)
- [ ] SBA Company Registry is current
- [ ] UEI number is correct in all forms
- [ ] PI employment certification is accurate
- [ ] Proposal meets page limits (count carefully -- cover page, TOC, and references may or may not count depending on agency)
- [ ] Font size and margins meet requirements
- [ ] All required volumes/attachments are included
- [ ] Budget totals match across all forms
- [ ] Budget narrative justifies every line item
- [ ] PI resume is included and current
- [ ] Technical approach directly addresses the topic requirements (not generic)
- [ ] Innovation is clearly stated (what is new?)
- [ ] Work plan has specific, measurable milestones
- [ ] Commercialization potential is addressed
- [ ] Proposal was reviewed by at least one other person before submission
- [ ] Proposal was submitted at least 24 hours before the deadline (portals crash near deadlines)
- [ ] You saved a PDF copy of the submitted proposal for your records

---

## Appendix C: Key Terminology

| Term | Meaning |
|------|---------|
| BAA | Broad Agency Announcement -- open-ended solicitation for research proposals |
| CCR | Company Commercialization Report -- auto-generated from your SBIR.gov profile |
| CAGE code | Commercial and Government Entity code -- assigned during SAM.gov registration |
| cATO | Continuous Authority to Operate -- DoD's modern approach to software accreditation |
| D2P2 | Direct to Phase II -- skip Phase I if feasibility is already demonstrated |
| DCAA | Defense Contract Audit Agency -- audits cost accounting on government contracts |
| DSIP | Defense SBIR/STTR Innovation Portal -- DoD submission and management portal |
| NAICS | North American Industry Classification System -- codes that describe your business type |
| PI | Principal Investigator -- the lead researcher on the project |
| STIG | Security Technical Implementation Guide -- DoD security configuration standards |
| TPOC | Technical Point of Contact -- the government person who wrote the topic |
| UEI | Unique Entity Identifier -- replaced DUNS number; obtained through SAM.gov |

---

*This guide represents information current as of early 2026. SBIR/STTR rules, amounts, and processes evolve. Always verify current solicitation requirements at the relevant agency portal before submitting.*
