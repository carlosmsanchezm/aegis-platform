# Defense Sales Guide: Selling DevSecOps Infrastructure to Defense Primes and Defense Tech

**Audience**: 1-person company selling a Kubernetes-based DevSecOps platform (Aegis) to defense contractors and defense-adjacent AI/ML companies.

**Channels covered**: Channel 4 (Defense Primes) and Channel 5 (Direct Enterprise Sales)

**Last updated**: March 2026

---

## Table of Contents

1. [Target Companies and How They Buy](#1-target-companies-and-how-they-buy)
2. [OEM/Embed Licensing and Subcontracting](#2-oemembed-licensing-and-subcontracting)
3. [Conferences and Events](#3-conferences-and-events)
4. [Accelerators and Incubators](#4-accelerators-and-incubators)
5. [Marketing and Positioning](#5-marketing-and-positioning)
6. [Pricing and Deal Structure](#6-pricing-and-deal-structure)
7. [Legal Considerations](#7-legal-considerations)
8. [Cold Outreach Playbook](#8-cold-outreach-playbook)
9. [90-Day Action Plan](#9-90-day-action-plan)

---

## 1. Target Companies and How They Buy

### Tier 1: Large Primes (GPU Clusters, Classified AI Work)

These companies operate their own GPU clusters for classified AI/ML work and need DevSecOps infrastructure for their programs. The IT services primes (Booz Allen, SAIC, Leidos, GDIT) are the easiest near-term targets because they buy tools to embed in government deliverables.

| Company | Revenue | AI/ML Focus Areas | Typical DevSecOps Buying Center | Innovation Entry Point |
|---------|---------|-------------------|-------------------------------|----------------------|
| **Lockheed Martin** | ~$67B | AI Factory, Project Overmatch, classified ML, JADC2 | IT/DevSecOps org under CTO; individual program offices | LM Ventures; Open Innovation portal (lockheedmartin.com/openinnovation) |
| **Raytheon/RTX** | ~$69B | Sensor fusion ML, electronic warfare AI, LTAMDS radar AI | RTX Technology Research Center; Raytheon BBN | RTX Ventures; RTX Innovation Challenge |
| **Northrop Grumman** | ~$39B | Autonomous systems (X-47B lineage), space AI, IBCS | NG IT, Chief Technology Office | NG FAST Labs; NG Ventures |
| **General Dynamics** | ~$42B | GDIT is the IT buyer; classified clouds, zero-trust, AI/ML services | GDIT Cloud and Edge Solutions division | GDIT partner ecosystem program |
| **BAE Systems** | ~$26B | EW, cyber, SIGINT ML pipelines | FAST Labs (research); Platform Solutions | BAE Systems AI Lab; UK-US dual presence |
| **L3Harris** | ~$21B | ISR, satellite imagery AI, electronic warfare | Technology and Operations orgs | L3Harris Innovation Center |
| **Booz Allen Hamilton** | ~$9B | Largest consulting AI practice for DoD; builds AI tools for clients | DarkLabs (tech R&D); Digital Solutions group | Booz Allen partner program; they actively seek tools to bring to government clients |
| **SAIC** | ~$7.4B | AI/ML for DoD clients, Tenjin platform, digital engineering | Technology and Innovation org; Cloud Solutions | Strong channel partner candidate -- they resell tools wrapped in services |
| **Leidos** | ~$15B | Managed Detection and Response, classified IT, digital modernization | Digital Modernization sector | Leidos Innovation Center |
| **ManTech** | ~$2.6B | Cyber, intel community IT operations | Cyber and Intelligence division | Acquired by Carlyle 2022; more aggressive on tech adoption |

**Key insight**: Booz Allen, SAIC, Leidos, and GDIT are actually easier targets than weapon-system primes because they are IT services companies. They buy tools to embed in their government deliverables. They are constantly looking for DevSecOps tooling they can wrap with professional services and sell to DoD clients. A single relationship at SAIC or Booz Allen can put your product in front of dozens of government programs.

### Tier 2: Defense Tech Companies

These companies buy faster with less procurement overhead. Their engineering teams evaluate tools the way commercial tech companies do, but they need government-grade compliance.

| Company | What They Do | Why They Need DevSecOps | How to Reach Them |
|---------|-------------|------------------------|-------------------|
| **Palantir** | Foundry/Gotham/AIP for DoD data | They build their own but subcontract components; need GPU workload management | Government team at conferences; LinkedIn to their Federal CTO |
| **Anduril** | Lattice platform, autonomous systems | Massive Kubernetes infrastructure; rapid deployment cycles | Engineering blog readers; direct to platform eng team via LinkedIn |
| **Shield AI** | Autonomous drones (V-BAT, Hivemind) | ML training infrastructure, edge deployment | Smaller team, reachable via LinkedIn; attend their recruiting events |
| **Scale AI** | Data labeling, Donovan (DoD AI platform) | Need secure infra for classified labeling | Government team; their Federal division |
| **Rebellion Defense** | ML DevOps for DoD | Direct competitor and potential partner | Conference meetings |
| **Primer AI** | NLP for intel community | Classified NLP infrastructure | Small enough for direct CEO/CTO outreach |
| **Vannevar Labs** | NLP/computer vision for IC/DoD | Classified deployment infrastructure | Direct outreach; IC community events |
| **Hawkeye 360** | RF geolocation satellite AI | Satellite data processing infra | Direct outreach; space/intel conferences |
| **Shift5** | OT/IoT security for military | DevSecOps platform needs | Direct outreach; Tampa/SOCOM events |
| **Two Six Technologies** | AI/ML, cyber for IC | Classified GPU workloads | Direct outreach; IC conferences |

### Tier 3: Cleared Facilities and Research Labs

| Type | Examples | Why They Buy |
|------|---------|-------------|
| FFRDCs | MITRE, Aerospace Corp, RAND | Research computing infrastructure, often set technology standards |
| UARCs | APL (Johns Hopkins), ARL, Lincoln Lab (MIT) | GPU clusters for AI research, influence DoD tool selection |
| National Labs | LANL, Sandia, ORNL, LLNL | HPC/GPU scheduling with compliance, massive compute budgets |
| Mid-size cleared contractors | Hundreds of firms | Need compliant dev environments but cannot build their own |

### How These Companies Discover and Buy DevSecOps Tools

The buying process differs dramatically from commercial tech:

1. **Program-driven purchases**: Most purchases are tied to a specific government contract or program (e.g., "We won the JADC2 software contract and need a DevSecOps platform for it"). The tool must fit the program's ATO requirements. Sales cycle: 6-18 months.

2. **Word of mouth and referral from government**: DoD program offices increasingly specify tools or recommend them. If Platform One (Air Force) or Army Software Factory uses your tool, primes will learn about it. Getting on Iron Bank (DoD's hardened container registry) is a powerful signal.

3. **Engineering team discovery**: Engineers at primes find tools the same way commercial engineers do -- GitHub, blog posts, conference talks. But they then must navigate internal procurement, which can take 3-12 months. An internal champion who already tested your product informally is worth more than any sales pitch.

4. **OTA and SBIR sub-awards**: Primes doing OTA (Other Transaction Authority) work often need to demonstrate they are using innovative small business tech. Your product can fill that requirement while genuinely solving a problem.

5. **Corporate innovation and venture arms**: Most large primes have venture or innovation groups that scout technology. These are often the fastest way in because they exist to bypass normal procurement.

6. **IRAD (Internal R&D) budgets**: Primes can use their own R&D budgets ($25K-$200K) to evaluate technology without government approval. This is a good foot-in-the-door mechanism.

7. **GSA Schedule and contract vehicles**: Having your product on a GSA Schedule or available through an existing government contract vehicle removes procurement friction. Without it, you rely on the prime buying you as a COTS (Commercial Off The Shelf) product or including you as a subcontractor.

### Vendor Portals and Innovation Programs

| Company | Portal or Program | URL or Notes |
|---------|-------------------|-------------|
| Lockheed Martin | Supplier Wire / Open Innovation | supplierwire.lockheedmartin.com; lockheedmartin.com/openinnovation |
| Lockheed Martin | LM Ventures | Invests in startups; can lead to tech integration |
| Raytheon/RTX | Supplier Connection | suppliers.rtx.com |
| RTX | RTX Innovation Accelerator | Emerging tech evaluation |
| Northrop Grumman | NDSS Supplier Portal | northropgrumman.com/suppliers |
| General Dynamics | GDIT Partner Program / Supplier Portal | gdit.com/partners; gd.com/suppliers |
| BAE Systems | Supplier Registration | baesystems.com/suppliers |
| L3Harris | Supplier Diversity | l3harris.com/suppliers |
| Booz Allen Hamilton | iHub / Partner Ecosystem | boozallen.com/partnerships |
| SAIC | Partner Portal | saic.com/partners |
| Leidos | Supplier Portal / Innovation Factory | leidos.com/suppliers |

**Practical note**: Registering on supplier portals is necessary but not sufficient. These portals are primarily for procurement to manage existing relationships. Getting in the door still requires personal relationships and a compelling capability. Treat registration as a prerequisite, not a strategy.

---

## 2. OEM/Embed Licensing and Subcontracting

### Three Pathways into a Prime

**Path A: Direct Subcontract (Most Common)**

1. Prime wins a government contract that needs DevSecOps tooling
2. Prime's capture team identifies your product during proposal phase
3. You are named as a subcontractor in the proposal
4. If prime wins, you execute a subcontract agreement
5. Typical timeline: 6-18 months from first contact to revenue

This is the most common path because it ties your revenue to a specific government program. The downside is long timelines and dependency on the prime winning the contract. The upside is that once you are on a program, renewals and expansions are relatively easy.

**Path B: OEM/Embed License (Best for 1-Person Company)**

1. Prime's engineering team evaluates your tool independently (often funded by IRAD)
2. Internal champion sponsors an OEM license agreement
3. Your product is embedded in the prime's deliverable to the government
4. Government sees it as part of the prime's solution, not as a separate vendor
5. You provide support and maintenance to the prime, not the end customer
6. The prime handles compliance, ATO, cleared environment operations

This is the best path for a 1-person company because the prime takes on all the overhead of government compliance, security clearance requirements, and direct customer support. You focus on building and maintaining the product. The prime handles everything else.

**Path C: GSA Schedule / Contract Vehicle (Independent)**

1. You get your product listed on a GSA Schedule (MAS IT Category)
2. Government agencies can buy directly, or primes can order through it
3. Timeline: 3-6 months to get on schedule; ongoing compliance burden
4. This gives you independence but requires more resources to maintain
5. Consider using a GSA Schedule consultant ($5,000-$15,000) rather than doing it yourself

### Typical Contract Structures

**Subcontract Terms (what to expect)**:

- The prime dictates most terms through flow-down clauses from the prime contract
- Payment terms: Net 30-60 from the prime, but the prime may not get paid by the government for 60-90 days. Expect Net 90-120 in practice. Budget for this cash flow gap.
- Indemnification: You will indemnify the prime; the prime will not indemnify you. This is almost always non-negotiable.
- IP: Negotiate this carefully. See Section 7 for detailed guidance. This is the most important term to get right.
- Non-compete: Some primes will restrict you from selling to competing primes on the same program. Push back hard on broad non-competes that go beyond the specific program.
- DFARS flow-downs: Mandatory when the prime's contract is with DoD. You cannot opt out. See Section 7.

**OEM License Terms (what to negotiate)**:

- Annual subscription or perpetual license plus maintenance (18-22% of license fee annually)
- Pricing basis: per-cluster, per-node, per-user, or enterprise-wide
- Support SLA: response times, whether you support in classified environments (you probably cannot without clearance -- negotiate accordingly)
- Source code escrow: many primes require this for COTS products in classified environments. Use a standard escrow service like Iron Mountain or EscrowTech ($1,500-$3,000/year).
- Right to modify or extend: limited, with clear restrictions on redistribution
- Renewal terms: auto-renewal with price escalation cap (typically 3-5% per year)
- Exclusivity: resist granting exclusive rights; offer "preferred partner" status instead

### What Primes Look for in Small Technology Partners

1. **Proven technology that works today**: They do not want to fund your R&D unless it is through a structured SBIR or OTA contract. Demonstrate with a working product, not a roadmap.
2. **Company stability and continuity risk mitigation**: A 1-person company is a risk. Mitigate with: source code escrow, thorough documentation, willingness to train prime's engineers to be self-sufficient, and a written business continuity plan.
3. **Security clearance (or clearability)**: If the product will be used in classified environments, can you get a clearance? As a 1-person company, you can get a Personal Clearance (PCL) which is sufficient for many scenarios. For a Facility Clearance (FCL), you need a classified contract as a sponsor -- this is a chicken-and-egg problem that a mentor-protege relationship can solve.
4. **Compliance posture**: Evidence of FedRAMP readiness, NIST 800-53 control mapping, CMMC Level 2 preparation, FIPS 140-2 validated crypto. You do not need all of these on day one, but having a documented compliance roadmap demonstrates seriousness.
5. **Responsiveness**: Small companies that respond in hours, not weeks, are valued. This is your competitive advantage over larger vendors.
6. **Small business certifications**: These have concrete dollar value to primes. See below.

### Small Business Set-Aside Certifications

These are extremely valuable for defense sales because primes have legal obligations to subcontract to small businesses:

| Certification | What It Is | Why It Matters for You |
|--------------|-----------|----------------------|
| **Small Business (SBA)** | Default if under $47.5M avg annual revenue (NAICS 511210) | Primes must meet small business subcontracting plan goals; your size is an asset |
| **SDVOSB** | Service-Disabled Veteran-Owned Small Business | 3% government-wide contracting goal; sole-source authority up to $7M |
| **WOSB** | Women-Owned Small Business | Set-asides available in certain NAICS codes |
| **HUBZone** | Historically Underutilized Business Zone | 3% goal; price evaluation preference of 10% |
| **8(a)** | SBA 8(a) Business Development Program | Sole-source contracts up to $4.5M; mentor-protege program; 9-year development period |
| **SBIR/STTR awardee** | Small Business Innovation Research recipient | DoD-validated status; Phase III sole-source authority |

**If you qualify for any of these, get certified immediately.** Every large defense contract ($750K+) requires a small business subcontracting plan (FAR 19.704), and primes are graded on how well they execute these plans. Your small business status makes you inherently valuable to them.

### SBA and DoD Mentor-Protege Programs

**SBA All Small Mentor-Protege Program:**

- **What the prime provides**: Technical and management assistance, access to resources, subcontracting opportunities, sometimes office space or classified facility access, introductions to program managers
- **What you provide**: Small business status (helps the prime meet subcontracting goals), innovative technology the prime does not have internally, agility and speed
- **Duration**: Up to 6 years (two 3-year terms)
- **Application**: Through certify.sba.gov; requires a written mentor-protege agreement
- **Joint ventures**: Mentor and protege can form a joint venture that bids on set-aside contracts using the protege's small business status. This is extremely valuable to the prime.

**DoD Mentor-Protege Program:**

- Administered by OUSD(A&S) Office of Small Business Programs
- Similar structure to SBA program but specifically for DoD contracts
- Primes can receive credit toward their subcontracting goals
- Apply at acq.osd.mil/osbp/sb/mp.html

**Primes with active mentor-protege programs**: All of the Tier 1 primes maintain these programs. Lockheed Martin, Raytheon, Northrop Grumman, BAE Systems, Booz Allen Hamilton, SAIC, Leidos, General Dynamics.

**How to initiate**: Contact the prime's Small Business Liaison Officer (SBLO). This person is legally required to exist at every large defense contractor with government contracts over $750K. They are findable on the company's supplier diversity page and are generally responsive because their job performance is directly tied to small business engagement metrics. This is one of the most underutilized entry points for small defense tech companies.

---

## 3. Conferences and Events

### Tier 1: Must-Attend (Highest ROI for Defense DevSecOps Sales)

| Event | Typical Timing | Location | Cost to Attend | Cost to Exhibit | Why It Matters |
|-------|---------------|----------|----------------|-----------------|----------------|
| **Platform One DevSecOps Days** | 1-2x per year, varies | Various (often virtual or hybrid) | Usually free | N/A (apply to present) | **THE event for your exact product category**. Air Force Kessel Run, Platform One practitioners, Army Software Factory engineers. If you present here, you are in front of your direct buyers. Highest signal-to-noise ratio of any event. |
| **DoD DevSecOps Symposium** | Varies | Various | Free to $200 | Sponsorship-based | Direct buyers for DevSecOps tools. Smaller than AFCEA but almost everyone in the room is relevant to you. |
| **AFCEA TechNet Cyber** | May-June | Baltimore, MD | $300-$900 | $5,000-$15,000 | Cyber/IT focused. DISA, NSA, Cyber Command attendees. Strong networking at evening receptions. |
| **AUSA Annual Meeting** | October | Washington, DC | Free military; $200-$500 industry | $5,000-$50,000+ | 30,000+ attendees. Army program managers. Largest Army event. Overwhelming but essential for visibility. |
| **AFCEA TechNet Augusta** | August | Augusta, GA | $300-$700 | $3,000-$10,000 | Army Cyber Center of Excellence. Smaller, higher-quality conversations than AUSA. |

### Tier 2: High Value, Larger Investment

| Event | Typical Timing | Location | Cost to Attend | Cost to Exhibit | Why It Matters |
|-------|---------------|----------|----------------|-----------------|----------------|
| **NVIDIA GTC** | March | San Jose, CA (+ virtual) | $200-$2,500 | $10,000+ | GPU/AI focus with defense AI track. Meet GPU cluster operators from defense companies. Your product story (Kubernetes for GPU workloads) fits perfectly. |
| **KubeCon North America** | October-November | Varies (US cities) | $500-$1,500 | $5,000-$30,000 | Kubernetes community. DoD Platform One team attends and presents. Government track growing each year. |
| **DoDIIS (Defense Intelligence)** | August | Varies | $300-$800 | $5,000-$15,000 | Intelligence community IT. If targeting IC customers, this is essential. |
| **Sea-Air-Space** | April | National Harbor, MD | $300-$600 | $5,000+ | Navy/Marine Corps technology. Strong for naval programs. |
| **RSA Conference** | April | San Francisco | $500-$2,500 | $15,000+ | Security angle. Government cybersecurity buyers. Large but diffuse -- best if you have specific meetings planned. |
| **DEF CON / Black Hat** | August | Las Vegas | $460 (DEF CON, cash) / $2,800 (Black Hat) | N/A / $15,000+ | Security/hacker community. IC/DoD cyber operators attend. Aerospace Village at DEF CON has defense content. DIU scouts here. |

### Tier 3: Innovation Events (Low Cost, High Signal)

| Event | Typical Timing | Location | Cost to Attend | Why It Matters |
|-------|---------------|----------|----------------|----------------|
| **DIU Demo Days** | Quarterly (varies) | Various | Usually invite-only | Defense Innovation Unit showcases. Get invited by submitting to DIU's Commercial Solutions Opening. |
| **SOFWERX Events** | Monthly | Tampa, FL | Usually free | SOCOM innovation. Technical deep-dives with small audiences of special operations technology users. Very high value per interaction. |
| **AFWERX Challenge / Pitch Events** | Rolling | Various / virtual | Free | Direct Air Force engagement. Present your product to AF program managers. |
| **Army xTech** | Quarterly | Various | Free | Army innovation competition. Cash prizes plus access to Army programs. |
| **NavalX Tech Bridge Events** | Various | Regional hubs | Free | Navy innovation events at regional technology hubs. |
| **Platform One Community Events** | Varies | Virtual / in-person | Free | Meet P1 engineers who set DoD DevSecOps standards. |
| **NSIN Hacks** | Monthly | Various | Free | National Security Innovation Network problem-solving events. Build DoD relationships. |

### Cost-Effective Strategy for a 1-Person Company

**Do not exhibit at AUSA or AFCEA initially.** The booth cost ($5,000-$50,000) is not justified until you have a sales pipeline and need to demonstrate market presence. Instead:

1. **Apply to present** at Platform One DevSecOps Days (free; your product is directly relevant; highest-ROI event).
2. **Attend as a visitor** at AFCEA TechNet Cyber ($300-$900). Walk the floor, attend panels, network at evening receptions and in hallway conversations.
3. **Attend SOFWERX events** (free plus travel to Tampa). SOCOM operators need exactly what you build -- fast, secure deployment of AI/ML workloads in austere environments.
4. **Submit to DIU** for their commercial solutions opening (CSO) process. If selected, you present at Demo Days to actual acquisition professionals.
5. **Attend GTC** ($200-$2,500) for the GPU/AI cluster operator audience. Your product story maps directly to their problems.
6. **Apply to Army xTech and AFWERX challenges** (free). These are structured opportunities to present to military decision-makers.

**Annual conference budget for a 1-person company**: $3,000-$8,000 in registration and travel for 3-5 events. This is one of the highest-ROI investments you can make because defense sales are relationship-driven. One good conversation at AFCEA can lead to a $200K subcontract 6 months later.

### Government Industry Days

These are free events where a government program office briefs industry on upcoming procurements before a formal solicitation.

**What happens at an Industry Day:**
- Government presents requirements, timeline, evaluation criteria, and sometimes budget range
- Industry asks questions (which are usually published afterward on SAM.gov)
- Teaming opportunities emerge in the hallways and post-event mixers
- You learn about contracts 6-12 months before the RFP drops

**How to find them:**
1. **SAM.gov**: Filter by "Sources Sought" or "Special Notice." Many industry days are posted here with registration links.
2. **Agency procurement forecast pages**: Each DoD agency publishes annual procurement forecasts listing planned acquisitions.
3. **GovWin (Deltek)**: Paid service ($2,000-$5,000/year) that tracks industry days and pre-solicitation events. Expensive but valuable for pipeline building.
4. **PEO websites**: Program Executive Offices (Army PEO-EIS, Navy PEO-Digital, Air Force PEO-Digital) post events directly.
5. **LinkedIn**: Follow government program offices and acquisition commands. They post event announcements.
6. **OTA consortia** (NSTXL, SOSSEC, etc.): These manage OTA contract vehicles and host their own industry events.

**How to use Industry Days effectively:**
1. Attend every relevant one (free; you only pay travel)
2. Bring business cards and 50 copies of your 1-page capability statement (see Section 8)
3. Arrive early, stay late. The hallway conversations are more valuable than the briefing.
4. Introduce yourself to the government program team after presentations
5. Find the primes who are planning to bid -- they are all there scoping competitors and looking for teammates
6. Follow up within 48 hours: email every prime BD person you met with "Great meeting you at the [Program] Industry Day. We have [specific capability] that could strengthen your team's proposal."

---

## 4. Accelerators and Incubators

### Government-Affiliated Programs

| Program | Sponsoring Agency | What They Provide | Takes Equity? | How to Apply |
|---------|-------------------|-------------------|---------------|-------------|
| **AFWERX** (part of AFRL) | Air Force | SBIR/STTR funding ($50K-$2M+); direct access to Air Force end users; pitch events; pathway to Phase III production contracts | No -- non-dilutive | afwerx.com; SBIR.gov; Open Topic submissions (quarterly windows) |
| **NavalX** | Navy | Tech Bridges (regional innovation hubs); connects startups to Navy problems; facilitates OTA contracts but no direct funding | No | navysbir.com; nre.navy.mil; apply through regional Tech Bridges |
| **Army Applications Lab (AAL) / Army Futures Command** | Army | Rapid prototyping contracts via OTA ($50K-$5M); xTech competitions with cash prizes; connect to Army problem sets | No | armyfuturescommand.com; respond to Problem Statements and xTech calls |
| **NSIN (National Security Innovation Network)** | DoD-wide (OSD) | Hacks (hackathon-style events), X-Force fellowships, Propel accelerator program, customer discovery with DoD users | No | nsin.mil; apply to specific programs |
| **DIU (Defense Innovation Unit)** | DoD (OSD) | Prototype contracts via OTA (typically $1M-$10M); clear pathway to production contracts; strong credibility signal | No | diu.mil; respond to Area of Interest (AOI) postings |
| **In-Q-Tel (IQT)** | CIA and broader IC | Strategic investment ($500K-$3M typical); access to classified requirements; introductions across the intelligence community | **Yes** -- equity investment | iqt.org; submit through portal or get warm intro. Highly selective. |

### Private Defense-Focused Accelerators and Ecosystems

| Program | Location | What They Provide | Takes Equity? | Notes |
|---------|----------|-------------------|---------------|-------|
| **Capital Factory** | Austin, TX | Office space, mentor network, DoD connections; Center for Defense Innovation; active defense tech community | Varies by program | Strong Texas defense connections (Ft. Cavazos, Ft. Liberty, Army Futures Command). capitalfactory.com |
| **Techstars Allied Space Accelerator** | Various | $120K investment; 3-month program; space and defense focus | **Yes** -- ~6-10% equity | Relevant if product has space applications |
| **MassChallenge** | Boston | No equity; connections to defense community through government track | No | masschallenge.org |
| **Hacking 4 Defense (H4D)** | Stanford/universities | DoD problem validation with student teams; builds DoD relationships | No | h4d.stanford.edu; good for validating product-market fit |
| **NSTXL (National Security Technology Accelerator)** | Various | Not an accelerator per se; manages OTA consortium contract vehicles. Joining gives access to DoD OTA requests. | No | nstxl.org; join the consortium to receive and respond to OTA opportunities |

### SBIR/STTR: The Non-Dilutive Funding Engine

This is the single most important funding mechanism for a 1-person defense tech company. It deserves deep attention.

**Phase I**: $50,000-$275,000 (depending on the sponsoring agency) for 6-12 months of feasibility study. You prove your concept works for a specific military application.

**Phase II**: $500,000-$2,000,000 for 24 months of prototype development. You build a working prototype and demonstrate it with a government user.

**Phase III**: No dollar limit. Production and commercialization using non-SBIR funding sources (regular procurement dollars). This is where the real revenue is.

**Key facts that make SBIR uniquely valuable:**
- You keep your IP. The government gets a royalty-free license for government purposes, but you retain commercial rights. SBIR data rights are protected for 20 years.
- No equity dilution. This is non-dilutive government R&D funding.
- Phase III contracts can be awarded sole-source (no competition required). This means a program office that likes your Phase II product can give you a production contract without a competitive bid. This is enormously valuable.
- All agencies with more than $100M in extramural R&D budgets must participate. DoD is the largest SBIR funder at roughly $2B per year across the services.
- SBIR awardees receive favorable treatment in other DoD procurement evaluations.

**How to apply:**
1. Register at SAM.gov (get your UEI number -- this replaced DUNS)
2. Register at SBIR.gov
3. Browse open topics at SBIR.gov or agency-specific sites (AFWERX for Air Force, NavalSBIR for Navy)
4. Write a proposal (Phase I proposals are typically 10-25 pages depending on the agency)
5. Submit during the open solicitation window (AFWERX has quarterly open topics)
6. Award decisions: 60-120 days after submission close

**AFWERX Open Topic SBIR -- your best entry point:**

Unlike traditional SBIR topics (which describe a very specific problem the government wants solved), AFWERX Open Topic lets you propose your own solution to a broadly defined problem area. DevSecOps for classified AI infrastructure fits squarely within their interest areas (Digital Infrastructure, Cyber Resilience, AI/ML Operations).

**How to frame your SBIR proposal:**

Do NOT write: "Aegis is a Kubernetes platform that does multi-cluster workload scheduling."

DO write: "Classified AI workload deployment currently requires manual, error-prone processes that take weeks and create security gaps. We propose to demonstrate that automated, continuously-compliant deployment can reduce this from weeks to minutes while generating ATO evidence automatically. Our approach uses a hub-and-spoke Kubernetes architecture with NIST 800-53 controls enforced at deploy time, specifically designed for GPU workloads across IL4/IL5 environments."

Reference specific government pain points: cATO requirements, DISA STIG compliance, Platform One compatibility, time-to-ATO reduction, DoD Enterprise DevSecOps Reference Design implementation.

### Recommended Path for a 1-Person Company (Priority Order)

1. **Immediate**: Apply to the next AFWERX Open Topic SBIR window. Non-dilutive, $50K-$275K Phase I, validates your product with Air Force users, and opens the door to Phase II ($500K-$2M) and Phase III (no limit, sole-source).
2. **Within 60 days**: Submit a capability statement to DIU for any relevant AOIs in DevSecOps, AI/ML, or cloud infrastructure.
3. **Within 90 days**: Join the NSTXL consortium for OTA opportunity access. Respond to relevant OTA requests as they appear.
4. **Within 6 months**: Apply to Army xTech competition. Attend Capital Factory events if in Austin area. Engage with NSIN Propel if accepting applications.
5. **Evaluate carefully before pursuing**: In-Q-Tel takes equity and expects VC-style growth trajectory. Only pursue if that aligns with your company vision and you are comfortable with IC community requirements.

---

## 5. Marketing and Positioning

### Building Credibility as a 1-Person Company

The defense community is relationship-driven and risk-averse. Here is how to overcome the "too small to trust" objection.

**Reframe the narrative:**

Do not position yourself as "a 1-person startup trying to sell to the Pentagon." Position yourself as "a founder-led deep tech company with a working product built by an engineer who understands classified infrastructure." The defense community actually respects solo technical founders who build real products. Platform One started as a small team of frustrated airmen. Kessel Run was a startup-within-government. The defense tech community values builders over slide decks.

**Credibility accelerators, ranked by impact:**

1. **Security clearance**: If you have or can obtain a Secret or TS/SCI clearance, this immediately differentiates you. Many defense buyers will not even take a meeting without clearance because they cannot discuss their actual requirements.

2. **SBIR award**: Even a Phase I award gives you "DoD-validated" status. When a prime's procurement team asks "has the government evaluated this?", you can say yes.

3. **Platform One / Iron Bank listing**: If your container images are accepted into Iron Bank (DoD's hardened container registry at ironbank.dso.mil), you are pre-approved for DoD use. This dramatically shortens the ATO process for anyone deploying your product.

4. **cATO capability**: If your platform demonstrably enables continuous Authority to Operate, this is a massive selling point. cATO is one of the DoD CIO's top priorities and every program office is struggling with it.

5. **FedRAMP documentation**: Even "FedRAMP Ready" or a thorough System Security Plan (SSP) shows you understand government compliance. Full FedRAMP authorization is expensive ($200K-$500K+) and probably premature for a 1-person company, but having an SSP that maps your controls to NIST 800-53 costs only your time.

6. **Open source components**: Government increasingly favors open source. If parts of your platform are open source, highlight this. It reduces vendor lock-in concerns and aligns with DoD open source policy.

7. **Compliance certifications**: SOC 2 Type II, ISO 27001, CMMC Level 2 assessment. SOC 2 Type II is the most accessible starting point.

### White Papers and Technical Content

**White papers that defense buyers actually read (not marketing brochures):**

1. **"Enabling cATO for Classified AI Workloads on Kubernetes"** -- Map your product's capabilities to specific NIST 800-53 controls. Show how automated policy enforcement reduces ATO timeline from 12-18 months to weeks. Include a control-by-control mapping table.

2. **"Multi-Cluster DevSecOps for Air-Gapped Environments"** -- Address the specific pain of deploying AI/ML workloads across classification levels (IL4, IL5, IL6). Show how hub-and-spoke architecture handles the air-gap problem.

3. **"Secure Multi-Cluster GPU Scheduling for Classified AI Workloads"** -- Your core value proposition. GPU workload placement across clusters with compliance enforcement.

4. **"Compliance-First Kubernetes: STIG, FIPS, and Continuous ATO"** -- The compliance angle. Map to DISA STIGs and FIPS 140-2 requirements.

5. **"VS Code Remote Development in Classified Environments"** -- The Sovran/workspace differentiator. Secure remote development for developers with clearances.

**Format requirements:**
- 3-10 pages, PDF, professional but not flashy design
- Include architecture diagrams with security boundaries clearly marked
- Include compliance mapping tables (Control ID | Control Name | How Aegis Implements It)
- Include a specific technical walkthrough, not just abstract claims
- No marketing language. Defense engineers will dismiss anything that reads like a sales brochure.

**Distribution channels:**
- Post on your website with ungated access (defense buyers hate lead-capture forms)
- Share on LinkedIn with a substantive technical summary, not just a link
- Submit to AFCEA Signal magazine as a contributed article
- Submit to Defense Systems, C4ISRNET, or Federal News Network as thought leadership
- Present the content at Platform One DevSecOps Days or AFCEA events
- Email directly to contacts you make at conferences and industry days
- Use as leave-behinds when meeting people in person

**On case studies:** You likely do not have defense case studies yet. Create "reference architectures" instead. These show how your product would be deployed in a defense scenario using notional architecture (e.g., "A classified AI training cluster running on EKS GovCloud with IL5 controls, hub in us-gov-west-1, spokes in two operational environments"). This demonstrates domain knowledge without claiming customers you do not have.

### LinkedIn Strategy for Defense Tech

LinkedIn is the primary professional network for defense tech. It is more important than any other social platform for this market.

**Profile optimization:**
- Headline: "Founder, [Company] | DevSecOps for Classified AI Infrastructure | Kubernetes | NIST 800-53"
- Include clearance status if applicable (state the level; never reveal specific programs or caveats)
- About section: Technical founder story. What problem you solve. Why it matters for national security.
- Featured section: Link to white papers, product demo video, GitHub repository, and any SBIR awards.

**Content strategy (1-2 posts per week minimum):**
- Technical depth posts: How your product solves a specific defense DevSecOps problem. "Here's how we enforce STIG compliance at deploy time across 50 clusters simultaneously."
- Industry commentary: React to DoD CIO memos, Platform One announcements, new DISA STIGs, CMMC rule updates. Show you are paying attention to the policy landscape.
- Conference recaps: Share takeaways from events you attend. Tag people you met.

**Connections to build (target 500+ defense-relevant connections):**
- Platform One engineers and leadership (search "Platform One" on LinkedIn)
- Military software factory leads (Army Software Factory, Navy DevSecOps, Kessel Run)
- DoD CIO office staff
- DIU staff and portfolio companies
- Program managers at primes (search "[Company Name] DevSecOps" or "[Company Name] Platform Engineer")
- Defense tech founders (they refer each other frequently)
- Small Business Liaison Officers at primes
- Contracting officers at DoD agencies

**Groups to join:**
- AFCEA (national and local chapters)
- NDIA (National Defense Industrial Association)
- Platform One / DoD DevSecOps Community
- GovCon professional groups
- Kubernetes government user groups

**What NOT to do:**
- Do not spam connection requests with immediate sales pitches
- Do not post or reference classified information, CUI, or FOUO documents
- Do not claim government customers you do not have
- Do not criticize specific programs, agencies, or individuals publicly
- Do not share proprietary information about prime contractors

### Website Requirements for Government Buyers

Government procurement professionals and prime contractor engineers will visit your website before taking a meeting. They need to find specific information quickly:

1. **Clear product description**: What it does, what problem it solves, for whom. Written for a technical audience.
2. **Compliance and security page**: NIST 800-53 control mapping, FedRAMP status or roadmap, SOC 2 status, FIPS 140-2 status for cryptographic modules, CMMC readiness statement. This page should be detailed and specific.
3. **Architecture documentation**: Publicly available architecture overview showing security boundaries, data flows, and trust relationships.
4. **Contact information**: Real phone number, real address (P.O. box is acceptable), named contact person. Government buyers are suspicious of companies without verifiable contact info.
5. **CAGE code and UEI number**: Display prominently. This signals you are registered to do government business.
6. **Small business certifications**: Display SBA certifications, NAICS codes, and set-aside eligibility.
7. **Downloads page**: White papers, reference architectures, data sheets available without requiring email registration. Defense buyers work on restricted networks and will not fill out marketing forms.

**Design guidance**: Clean, professional, fast-loading. Use product screenshots and architecture diagrams. Do not use stock photos of jets, soldiers, or American flags unless you actually work with the military. Government buyers find this approach cringeworthy from companies without proven defense experience.

### Capability Statement

A **1-page PDF** that every government buyer expects. See Section 8 for the full template.

---

## 6. Pricing and Deal Structure

### Government vs. Commercial Pricing

**The fundamental rule**: Government buyers expect pricing transparency and generally expect a discount from commercial pricing (typically 10-20%). If you sell commercially at a given price, the government will discover this through market research, GSA negotiations, or prime contractor due diligence.

**GSA Schedule pricing**: If you put your product on a GSA Schedule, you must disclose your commercial sales practices and offer the government "Most Favored Customer" pricing or demonstrate that your government price is fair and reasonable.

**For a 1-person company, price to value, not to cost:**
- Your product replaces manual DevSecOps work that costs $200K-$400K per year in labor at a prime
- A team of 3-5 DevSecOps engineers at a prime costs $600K-$2M per year fully loaded (salary + benefits + clearance premium + overhead + fee)
- Price your product at 10-30% of the labor cost it displaces. This gives the buyer a clear ROI story.

### Pricing Models

| Model | Description | Government Preference | Best For |
|-------|-------------|----------------------|----------|
| **Per-cluster subscription** | Fixed annual fee per managed cluster | Moderate -- predictable | Multi-cluster deployments with known cluster count |
| **Per-node subscription** | Annual fee per Kubernetes node | Less preferred -- harder to budget | Large, variable-size deployments |
| **Enterprise license** | Flat annual fee for unlimited use within an org or program | **Strong preference** -- simplest to procure | Large primes, entire program offices |
| **Perpetual + maintenance** | One-time license fee plus 18-22% annual maintenance | **Historically preferred** (software as capital asset) | Traditional government procurement |
| **Per-deployment** | Fixed fee per deployment or environment | Moderate -- simple to understand | Smaller programs with clear scope |
| **Usage-based** | Pay per workload, per API call, etc. | Disliked -- unpredictable | Avoid for government sales |

### Suggested Price Points

| Deal Type | Price Range | Target Customer |
|-----------|------------|-----------------|
| **Single cluster license** | $75K-$150K/year | Small defense contractor, single program |
| **Multi-cluster (up to 5 spokes)** | $250K-$500K/year | Prime program office, large contractor |
| **Enterprise unlimited** | $500K-$1M/year | Enterprise-wide deployment at a prime |
| **IRAD evaluation license** | $25K-$50K for 6 months | Foot-in-the-door with primes using their R&D budget |
| **OEM/embed license** | $150K-$500K/year (minimum commitment) | Prime bundling in their deliverable to government |
| **Professional services** | $200-$350/hour or $2,500-$4,000/day | Integration, training, customization |

**Start higher than you think.** You can always negotiate down. You cannot negotiate up. Defense procurement professionals expect negotiation; if you do not leave room for it, you signal inexperience.

### Government Purchase Card (GPC) Sweet Spot

Micro-purchases under **$10,000** can be bought with a government purchase card -- no procurement process, no contracting officer, no competition required. Consider:
- $9,500 pilot license (90 days, single cluster)
- Gets your software into their environment with zero procurement friction
- Conversion to full license after successful pilot
- This is the fastest path to a government deployment

### Perpetual vs. Subscription in Government

- Government procurement traditionally favored perpetual licenses because software was categorized as a capital asset.
- Since the Cloud Smart policy (2019) and the DoD CIO's push for cloud-native solutions, subscription and SaaS pricing has become increasingly accepted.
- Many programs now have "IT as a Service" budget lines that explicitly expect subscription pricing.
- **Recommendation**: Offer both options. Let the customer choose based on their funding type (O&M funds favor subscription; procurement funds favor perpetual). For your cash flow, subscription is better.

### Support and Maintenance Contracts

Defense buyers expect formal, documented support with defined SLAs:

| Level | Response Time | Availability | Typical Price |
|-------|-------------|-------------|---------------|
| Basic | Next business day | Business hours (M-F, 8am-6pm ET) | Included in subscription |
| Standard | 4-hour response | 12x5 (extended business hours) | +20% of license fee |
| Premium | 1-hour response, 4-hour resolution target | 24x7 | +40% of license fee |

**As a 1-person company**: Only offer Basic and Standard initially. You cannot credibly deliver 24x7 support alone. When a prime asks for Premium, negotiate a tiered model where the prime's internal team handles L1/L2 and you handle L3 (engineering escalation). This is standard practice.

### Training and Professional Services

| Service | Duration | Price Range | Notes |
|---------|----------|-------------|-------|
| Administrator training | 2 days | $3,000-$5,000 per day | On-site at customer facility |
| DevSecOps workshop | 3-5 days | $3,000-$5,000 per day | Hands-on with customer's environment |
| Deployment and integration | 1-4 weeks | $2,500-$4,000 per day | Install, configure, integrate with CI/CD |
| Compliance mapping engagement | 2-4 weeks | $2,500-$4,000 per day | Map product controls to customer's ATO package |
| Custom feature development | Varies | $250-$400 per hour | Customer-specific feature requests |

**Strategy**: Use training and professional services as a wedge. A $10,000 training engagement (GPC-purchasable) is easier to approve than a $200,000 license. Once you are inside the building, delivering value and building relationships with the engineering team, the license sale follows naturally.

---

## 7. Legal Considerations

### ITAR/EAR: Does AI Infrastructure Software Have Export Control Issues?

**Short answer**: Pure DevSecOps infrastructure software (Kubernetes orchestration, CI/CD, deployment automation) is generally NOT ITAR-controlled. However, there are important nuances.

**ITAR (International Traffic in Arms Regulations):**
- Controlled by the State Department, Directorate of Defense Trade Controls (DDTC)
- Applies to defense articles and defense services listed on the U.S. Munitions List (USML)
- Pure infrastructure software is NOT on the USML
- **However**: If your software is specifically designed for, modified for, or configured for a defense article (e.g., custom-built to deploy missile guidance software), it could be classified as a defense service under USML Category XI(d)
- **Risk assessment for Aegis**: Low. Generic Kubernetes orchestration is not ITAR-controlled. But if a prime asks you to customize the platform specifically for a weapons program, get export control legal advice before proceeding.

**EAR (Export Administration Regulations):**
- Controlled by Commerce Department, Bureau of Industry and Security (BIS)
- Covers "dual-use" items (civilian items with potential military applications)
- Encryption software above certain key lengths is EAR-controlled
- Your platform uses standard TLS/encryption libraries (OpenSSL, Go crypto) which generally qualify for License Exception ENC
- **Likely classification**: ECCN 5D002 (encryption software) with License Exception ENC available, or EAR99 (not controlled at all)

**Practical steps:**
1. Do not export outside the US, Canada, or Five Eyes countries without an export control review
2. Be aware of "deemed export" rules: sharing controlled technology with a foreign national in the US can constitute an export
3. Include standard ITAR/EAR compliance language in contracts and terms of service
4. Before your first defense contract, get an export control attorney review ($2,000-$5,000 one-time)
5. Self-classify your product under the EAR Commerce Control List. Most likely EAR99 or ECCN 5D002 with License Exception ENC.

### DFARS Clauses You Must Know

These flow down from the prime's contract with the government to your subcontract. You cannot opt out.

| Clause | What It Requires | Impact on You |
|--------|-----------------|---------------|
| **DFARS 252.204-7012** (Safeguarding CDI) | Implement NIST SP 800-171 for Covered Defense Information; report cyber incidents within 72 hours via dibnet.dod.mil | **Critical**: Must have NIST 800-171 controls in place. Submit self-assessment score to SPRS. |
| **DFARS 252.204-7020** (NIST 800-171 Assessment) | Allow DoD to assess your implementation; submit score to SPRS | Primes will check your SPRS score before teaming. Do this proactively. |
| **DFARS 252.204-7021** (CMMC Requirements) | Cybersecurity Maturity Model Certification at required level | CMMC Level 1 (15 practices, self-assessment) for FCI. Level 2 (110 practices, third-party assessment, $30K-$100K+) for CUI. |
| **DFARS 252.227-7014** (Rights in Noncommercial Technical Data) | Government gets unlimited rights to technical data produced under contract | **Negotiate**: Ensure pre-existing IP is identified as "Limited Rights" data. See IP section below. |
| **DFARS 252.227-7015** (Technical Data -- Commercial Items) | Government gets a commercial license | More favorable than 7014. Push for this clause. Applies when your product is a commercial item. |
| **DFARS 252.204-7018** (Prohibition on Certain Telecommunications) | Cannot use equipment from Huawei, ZTE, Hytera, Hikvision, Dahua, Kaspersky | Audit your supply chain. Ensure no prohibited components. |

### Intellectual Property Rights in Government Contracts

This is the most important legal issue for a software company in defense.

**Categories of IP rights (from most to least restrictive for the government):**

1. **Unlimited Rights**: Government can use, modify, reproduce, release, and disclose without restriction. **Avoid this for your core product.**
2. **Government Purpose Rights (GPR)**: Government use within government and to support contractors. Converts to unlimited after 5 years unless negotiated otherwise.
3. **Limited Rights (technical data) / Restricted Rights (computer software)**: Government can use internally but cannot release outside government. **This is what you want for your pre-existing commercial product.**
4. **Specifically Negotiated License Rights (SNLR)**: Custom terms. **Best option if you can negotiate it.**
5. **SBIR data rights**: Special protection for 20 years. Government gets government-purpose rights only. **Best available if going the SBIR route.**

**Key principle**: Always assert that Aegis is a **commercial item** under FAR 12.101. This gives you the strongest IP protection. Commercial item determination means the government buys a license to your existing product rather than funding development. The government gets a commercial license, and you keep everything.

**Critical protection steps:**
1. **Document all pre-existing IP** before signing any contract. Create a formal IP assertion letter listing all pre-existing software, with dates and version numbers. Use git history as evidence of pre-existence.
2. **Mark everything**. Every deliverable must carry restrictive markings (e.g., "RESTRICTED RIGHTS" under DFARS 252.227-7014).
3. **Separate pre-existing from contract-funded work**. Code developed at private expense qualifies for limited/restricted rights. Mixed funding is complex -- document carefully.
4. **Negotiate at proposal time**. IP rights are easiest to negotiate before contract award. After award, you have very little leverage.

### Indemnification and Insurance

- Primes will require you to indemnify them for IP infringement claims, software defects, and security breaches
- Government contracts include FAR 52.233-1 (Disputes clause) for dispute resolution
- **Insurance you need:**
  - **E&O / Professional Liability**: $1,500-$5,000/year for $1M-$2M coverage. Some primes require $5M coverage.
  - **Cyber Liability**: $1,000-$3,000/year. Covers data breach costs. Increasingly required.
  - **General Commercial Liability**: $500-$1,500/year. Standard business insurance.
- **Total insurance budget**: $3,000-$10,000/year. Non-negotiable cost of doing defense business.

### Must-Have Legal Documents

Engage a government contracts attorney ($5,000-$15,000 for initial preparation):

1. **COTS License Agreement**: Software license terms acceptable to government buyers with DFARS-compatible IP provisions
2. **DFARS Flow-Down Compliance Matrix**: Shows how you comply with common flow-down clauses
3. **NIST 800-171 System Security Plan (SSP)**: Documents your cybersecurity controls; required for SPRS submission
4. **IP Assertion Letter Template**: Pre-identifies your pre-existing commercial IP with dates and funding sources
5. **Mutual NDA Template (with CUI protections)**: For discussions involving controlled unclassified information
6. **Capability Statement**: 1-page PDF expected by every defense buyer

**Finding a government contracts attorney:**
- Pillsbury Winthrop Shaw Pittman, Wiley Rein (DC-based firms with deep GovCon practices)
- Small firm specialists (often former JAG or government procurement attorneys)
- Ask for referrals at AFCEA events or from other defense tech founders
- Budget: $5,000-$15,000 for initial document preparation; $300-$600/hour for ongoing

---

## 8. Cold Outreach Playbook

### Identifying the Right Person

**At a large prime (Lockheed, Raytheon, Booz Allen, SAIC, etc.):**

| Role | Typical Title Patterns | Why They Matter | Reachability |
|------|----------------------|-----------------|-------------|
| Technical decision maker | "DevSecOps Lead," "Platform Engineering Manager," "Cloud Architect," "Chief Engineer -- Software" | Will evaluate your product technically and champion it internally | Medium -- findable on LinkedIn |
| Program manager | "Program Manager -- [program]," "Director of [domain]" | Controls budget; decides what tools to buy | Low cold, high at Industry Days |
| Innovation/ventures | "Director of Innovation," "VP Technology Ventures" | Chartered to find new tech; most receptive to cold outreach | High -- it is their job |
| Small Business Liaison (SBLO) | "Small Business Program Manager," "Supplier Diversity" | Legally required to engage small businesses; measured on engagement metrics | **Very High** -- best entry point |
| Capture/BD manager | "Capture Manager," "BD Manager -- [domain]," "Solutions Architect" | Writing proposals; looking for teammates and tools to include | Medium-High |

**At a defense tech company (Anduril, Shield AI, Scale AI, etc.):**

| Role | Typical Title Patterns | Why They Matter | Reachability |
|------|----------------------|-----------------|-------------|
| Platform/infra engineering | "Head of Platform," "Staff Platform Engineer," "DevOps Lead," "SRE Manager" | Direct user of your product | High -- engineers respond to technical substance |
| Engineering leadership | "VP Engineering," "CTO," "Head of Infrastructure" | Budget authority | Medium -- defense tech companies are smaller than primes |
| Government programs | "Director of Government Programs," "Federal CTO" | Understands compliance context | Medium |

### How to Find Specific People

1. **LinkedIn (free search)**: Search "[Company Name] DevSecOps" or "[Company Name] Platform Engineer Kubernetes." Filter by current company.
2. **LinkedIn Sales Navigator** ($99/month): Advanced filters by company, title, seniority, function, and keywords. Worth the cost for sustained outreach.
3. **Conference speaker lists**: Check AFCEA, AUSA, KubeCon, Platform One DevSecOps Days, GTC speaker lists. Speakers are public-facing and receptive to follow-up.
4. **GitHub**: Search for employees who contribute to Kubernetes, Helm, or cloud-native projects. Commit history reveals what they actually work on.
5. **FPDS (fpds.gov)**: Search for contracts by prime and NAICS code. Contract descriptions reveal active programs and help you identify relevant program managers.
6. **USASpending.gov**: Search by company and keyword to find active funded contracts.
7. **GovWin (Deltek)**: Paid ($2,000-$5,000/year). Tracks opportunities, identifies key personnel, shows competitive landscape.
8. **SAM.gov**: Free. Active solicitations often list government POCs. Awards show program structure.

### LinkedIn Outreach Templates

**Connection request (build relationship first, no pitch):**
```
Hi [Name], I'm building DevSecOps tooling for classified Kubernetes
environments and your work on [specific project/post/talk] caught
my attention. Would be great to connect.
```

**Follow-up message (after connection accepted, wait 2-3 days):**
```
Thanks for connecting, [Name]. I noticed [Company] is working on
[specific program or capability -- from public sources: press release,
conference talk, job posting]. We've built a platform that automates
multi-cluster Kubernetes deployment with NIST 800-53 controls baked
in -- specifically designed for teams running GPU workloads in
restricted environments.

Would a 15-minute technical overview be useful? Happy to share our
architecture doc either way -- no strings attached.
```

**Outreach to SBLO:**
```
Hi [Name], I'm the founder of [Company], a small business building
DevSecOps infrastructure for classified AI workloads on Kubernetes.
We're interested in exploring subcontracting and potential mentor-
protege opportunities with [Prime]. Could we schedule a brief call
to discuss how our capabilities might align with [Prime]'s programs?

Our CAGE code is [XXXXX] and we're registered in SAM.gov.
```

**Key principles:**
- Reference something specific about their work (from public sources only)
- Lead with the technical problem, not your company
- Offer value first (architecture doc, white paper)
- Keep messages under 100 words
- Never mention pricing in cold outreach
- Never reference classified programs
- Be patient -- defense sales cycles are 6-18 months

### Email Outreach

**Finding email addresses:**
- Most defense companies: firstname.lastname@company.com
- Verify with Hunter.io, Apollo.io, or RocketReach
- Government .mil addresses are often searchable on agency websites

**Email template (to technical decision maker at a prime):**
```
Subject: Multi-cluster DevSecOps for [specific domain]

[Name],

I saw your [talk at AFCEA / LinkedIn post about / team's work on]
[specific topic from public source]. We've been solving a related
problem from the tooling side.

We built a Kubernetes-native platform that handles multi-cluster
workload scheduling with continuous compliance -- NIST 800-53
controls enforced automatically at deploy time. It's designed for
teams running GPU workloads across classification levels where
manual deployment processes create bottlenecks.

Two things that might be relevant to [Company]:
- Automated ATO evidence generation (cuts months off the process)
- Hub-and-spoke architecture that works across air-gapped networks

Technical overview here: [link to white paper]

Worth a 15-minute call to see if there's a fit?

[Your name]
[Phone]
[Company]
[Website]
```

**Follow-up cadence:**
- Day 0: Initial email
- Day 5: Follow up with different value (relevant article, new white paper, shared conference)
- Day 14: Final brief follow-up
- After 3 touches with no response: Move on. Approach in person at conferences if possible.

### The 1-Page Capability Statement

Every defense company expects this document. It is the defense equivalent of a pitch deck. Strictly one page.

```
[COMPANY LOGO]                              [COMPANY NAME]
                                            CAGE Code: XXXXX
                                            UEI: XXXXXXXXXXXX
                                            NAICS: 511210, 541512, 541519
                                            Small Business: Yes
                                            Certifications: [SB, SDVOSB, etc.]

CORE CAPABILITY
Multi-cluster Kubernetes DevSecOps platform for classified AI/ML
workloads. Automated compliance (NIST 800-53), continuous ATO
evidence generation, GPU workload scheduling across classification
levels.

DIFFERENTIATORS
- Hub-and-spoke architecture for air-gapped multi-cluster deployment
- Continuous compliance enforcement at deploy time
- Built on open-source Kubernetes -- no vendor lock-in
- Designed for IL4/IL5/IL6 environments with FIPS-validated crypto
- Reduces time-to-ATO from months to weeks
- Secure VS Code remote workspaces in classified environments

PAST PERFORMANCE
[List relevant projects, SBIR awards, commercial deployments, or
relevant prior experience]

RELEVANT EXPERIENCE
[Founder bio: clearance level, years of DoD/IC experience, technical
background, Kubernetes/cloud-native expertise]

CONTACT
[Name] | [Email] | [Phone] | [Website]
```

Print 50 copies on quality paper for every conference and industry day. Hand them to every relevant person you meet.

---

## 9. 90-Day Action Plan

### Days 1-30: Foundation

| # | Action | Cost | Time | Priority |
|---|--------|------|------|----------|
| 1 | Register at SAM.gov (get UEI number and CAGE code) | Free | 2-4 weeks processing | **P0** |
| 2 | Assign NAICS codes: 511210, 541512, 541519 | Free (part of SAM.gov) | Included above | **P0** |
| 3 | Verify SBA small business certification | Free | 1-2 weeks | **P0** |
| 4 | Create 1-page capability statement | Free | 1 day | **P0** |
| 5 | Complete NIST 800-171 self-assessment, submit score to SPRS | Free | 1-2 weeks | **P1** |
| 6 | Optimize LinkedIn profile for defense tech audience | Free | 2 hours | **P1** |
| 7 | Write first white paper on cATO or classified AI DevSecOps | Free | 1 week | **P1** |
| 8 | Add compliance documentation page to website | Free | 2 days | **P1** |
| 9 | Initial consult with government contracts attorney (1-2 hours) | $500-$1,000 | Half day | **P1** |
| 10 | Identify next 3 relevant Industry Days on SAM.gov | Free | 2 hours | **P2** |
| 11 | Research next AFWERX Open Topic SBIR window | Free | 1 hour | **P2** |
| 12 | Register on 3-5 prime vendor portals | Free | 1 day | **P2** |

### Days 31-60: Outreach

| # | Action | Cost | Time | Priority |
|---|--------|------|------|----------|
| 13 | Begin LinkedIn outreach: 10 targeted connections/week | Free | 30 min/day | **P0** |
| 14 | Contact SBLOs at 3 target primes (start with Booz Allen, SAIC, GDIT) | Free | 1 day each | **P0** |
| 15 | Submit AFWERX SBIR Phase I proposal (or prepare draft for next window) | Free | 2-3 weeks writing | **P0** |
| 16 | Submit capability statement to DIU for relevant AOIs | Free | 1 day | **P1** |
| 17 | Attend 1 Industry Day | Travel only | 1-2 days | **P1** |
| 18 | Apply to present at Platform One DevSecOps Days | Free | 2 hours | **P1** |
| 19 | Activate LinkedIn Sales Navigator | $99/month | Ongoing | **P2** |
| 20 | Join NSTXL consortium | Free to apply | 1 day | **P2** |
| 21 | Write second white paper or reference architecture | Free | 1 week | **P2** |
| 22 | Connect with 20+ defense tech professionals on LinkedIn | Free | Ongoing | **P2** |

### Days 61-90: Engage

| # | Action | Cost | Time | Priority |
|---|--------|------|------|----------|
| 23 | Attend AFCEA TechNet or AUSA (whichever is next) | $300-$1,500 + travel | 2-3 days | **P0** |
| 24 | Follow up on SBIR submission | Free | Check periodically | **P0** |
| 25 | Deliver technical demos to 2-3 prime engineering teams | Free | 1-2 hours each | **P0** |
| 26 | Begin GSA Schedule application (or engage consultant) | Free self-apply; $5K-$15K consultant | Ongoing (3-6 months total) | **P1** |
| 27 | Attend a SOFWERX event in Tampa | Free + travel | 1-2 days | **P1** |
| 28 | Identify 1-2 upcoming contracts where you could sub | Free (SAM.gov) | Ongoing | **P1** |
| 29 | Submit Iron Bank container hardening request | Free | 2-4 weeks process | **P2** |
| 30 | Publish 2-3 blog posts on defense DevSecOps topics | Free | 2-3 hrs/post | **P2** |
| 31 | Get E&O and cyber liability insurance quotes | $3K-$10K/year | 2-3 hours | **P2** |
| 32 | Apply to AFWERX Challenge or Army xTech if open | Free | 1 day | **P2** |

### Key Milestones by Day 90

- [ ] SAM.gov registration complete with CAGE code and UEI
- [ ] 1-page capability statement created, printed, and distributed
- [ ] SBIR proposal submitted (or draft ready for next window)
- [ ] Connected with SBLOs at 3+ large primes
- [ ] 50+ LinkedIn connections in defense DevSecOps community
- [ ] Attended 1+ conference or industry day
- [ ] 1+ white paper published and distributed
- [ ] 1+ technical demo delivered to a prime or defense tech company
- [ ] Government contracts attorney retained
- [ ] Compliance documentation visible on website
- [ ] NIST 800-171 self-assessment score in SPRS
- [ ] Registered on 3+ prime vendor portals

### Month 4-6 Targets

- [ ] Convert 1-2 demos to IRAD evaluations or pilot licenses
- [ ] Submit to 1-2 prime innovation programs
- [ ] Attend second conference
- [ ] Pursue mentor-protege discussions with interested prime
- [ ] Apply for 8(a) or other SBA certifications if eligible
- [ ] Start LinkedIn posting cadence (1-2x/week)

### Month 7-12 Targets

- [ ] Close first deal (even if small -- IRAD eval, pilot, training engagement)
- [ ] Convert SBIR Phase I to Phase II proposal (if awarded)
- [ ] Build 2-3 reference relationships at primes
- [ ] Get on Iron Bank (if container assessment process is complete)
- [ ] Establish 1 mentor-protege relationship
- [ ] Begin planning for first exhibit booth at a conference (AUSA or AFCEA)

---

## Appendix A: Key Websites and Resources

| Resource | URL | Purpose |
|----------|-----|---------|
| SAM.gov | sam.gov | Entity registration, contract opportunities, industry day notices |
| SBIR.gov | sbir.gov | SBIR/STTR opportunity listings and submission portal |
| AFWERX | afwerx.com | Air Force SBIR and innovation programs |
| DIU | diu.mil | Defense Innovation Unit commercial solutions |
| FPDS | fpds.gov | Federal contract award data (who won what) |
| USASpending | usaspending.gov | Government spending transparency data |
| GovWin (Deltek) | govwin.com | Opportunity tracking and intelligence (paid) |
| SPRS | sprs.csd.disa.mil | Submit NIST 800-171 self-assessment scores |
| Iron Bank | ironbank.dso.mil | DoD hardened container registry |
| Platform One | p1.dso.mil | DoD DevSecOps platform and community |
| NSTXL | nstxl.org | OTA consortium for defense innovation |
| DIBNET | dibnet.dod.mil | Defense Industrial Base cyber incident reporting |
| SBA Certify | certify.sba.gov | Small business certification portal |
| DSBS | dsbs.sba.gov | Dynamic Small Business Search (verify your listing) |
| NSIN | nsin.mil | National Security Innovation Network programs |

## Appendix B: NAICS Codes for DevSecOps Software

| NAICS Code | Description | When to Use |
|------------|-------------|-------------|
| 511210 | Software Publishers | Primary code for product sales |
| 541512 | Computer Systems Design Services | Consulting, integration, deployment |
| 541519 | Other Computer Related Services | Managed services, ongoing support |
| 541511 | Custom Computer Programming Services | Custom development work |
| 541513 | Computer Facilities Management Services | Cloud/infrastructure management |
| 518210 | Computing Infrastructure Providers | Hosted/SaaS version |

## Appendix C: Glossary of Defense Procurement Terms

| Term | Definition |
|------|-----------|
| **ATO** | Authority to Operate -- formal permission to run a system on a DoD network |
| **cATO** | Continuous ATO -- automated, ongoing authorization using real-time monitoring |
| **CAGE Code** | Commercial and Government Entity code -- 5-character company identifier in DoD systems |
| **CDI** | Covered Defense Information -- unclassified info requiring safeguarding under DFARS |
| **CMMC** | Cybersecurity Maturity Model Certification -- tiered standard for defense contractors |
| **COTS** | Commercial Off-The-Shelf -- commercial product used without significant modification |
| **CSO** | Commercial Solutions Opening -- DIU's solicitation mechanism |
| **CUI** | Controlled Unclassified Information -- sensitive but unclassified info with specific handling rules |
| **DFARS** | Defense Federal Acquisition Regulation Supplement -- defense procurement regulations |
| **FAR** | Federal Acquisition Regulation -- primary federal procurement regulation |
| **FCL** | Facility Clearance -- allows a company facility to handle classified information |
| **FedRAMP** | Federal Risk and Authorization Management Program -- cloud security authorization |
| **GPC** | Government Purchase Card -- credit card for micro-purchases under $10K |
| **GPR** | Government Purpose Rights -- IP category allowing government use and support contractor disclosure |
| **GSA Schedule** | General Services Administration contract vehicle simplifying government purchasing |
| **IL4/IL5/IL6** | Impact Level 4/5/6 -- DoD cloud classification (IL4=CUI, IL5=higher CUI, IL6=SECRET) |
| **IRAD** | Internal Research and Development -- prime's own R&D budget (not government-funded) |
| **OTA** | Other Transaction Authority -- flexible contracting not subject to FAR |
| **PCL** | Personnel Clearance -- individual security clearance (Secret, TS, TS/SCI) |
| **RFP** | Request for Proposal -- formal solicitation inviting proposals |
| **SBIR** | Small Business Innovation Research -- non-dilutive R&D funding program |
| **SBLO** | Small Business Liaison Officer -- person at a prime responsible for small business engagement |
| **SNLR** | Specifically Negotiated License Rights -- custom IP rights terms |
| **SPRS** | Supplier Performance Risk System -- DoD database for contractor cybersecurity scores |
| **SSP** | System Security Plan -- document describing security controls for a system |
| **STIG** | Security Technical Implementation Guide -- DoD configuration standards |
| **STTR** | Small Business Technology Transfer -- SBIR variant requiring research institution partner |
| **UEI** | Unique Entity Identifier -- replaced DUNS number for government registration |

---

*This guide is based on publicly available information about defense industry procurement and sales practices as of early 2026. Specific pricing, program details, and procurement thresholds may change. Consult a government contracts attorney before entering into any defense contract.*
