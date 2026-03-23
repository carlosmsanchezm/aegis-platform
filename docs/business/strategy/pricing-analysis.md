# Aegis Platform: Value-Based Pricing Analysis

## The Compliance Moat

Aegis's competitive moat is not code complexity — it's compliance engineering. A talented DevSecOps engineer with AI tools could replicate the platform functionality in 2-3 months. But building the platform is not the hard part. The hard part is everything that comes after.

### Why "Just Build It Yourself" Fails in Regulated Environments

**ATO Timeline (6-18 months, regardless of engineering speed):**
Every NIST 800-53 control needs a control implementation statement. Every control needs evidence. The ISSM reviews the entire architecture. The AO signs off. This process takes 6-18 months regardless of how fast you build the platform, because it's limited by human review cycles, not engineering speed. Aegis comes with control implementations already documented, OSCAL evidence export built in, and a security architecture that an ISSM can evaluate in weeks instead of months.

**CMMC Level 2 Continuous Compliance:**
CMMC isn't a one-time checkbox — it requires continuous monitoring, evidence collection, and periodic assessment by a C3PAO. Aegis automates evidence collection. A DIY solution means someone has to manually gather evidence, document controls, and prepare for assessments. That's ongoing labor, not a one-time build cost.

**ITAR and CUI Data Handling Controls:**
The moment CUI or ITAR data touches the platform, you need to prove these controls work with evidence:
- AC-11 (Session Lock) — Aegis: auto-disconnect on inactivity
- AC-12 (Session Termination) — Aegis: session revocation with JTI tracking
- AU-3 (Audit Content) — Aegis: structured JSON logs with subject, action, resource, outcome, source IP
- MP-7 (Media Protection) — Aegis: RAM-disk sandbox, clipboard wipe, destroyed on exit
- SC-8 (Transmission Confidentiality) — Aegis: TLS 1.2+ everywhere, mTLS between services
- SC-13 (Cryptographic Protection) — Aegis: FIPS-validated BoringCrypto (CMVP #4407)

A DIY team would need to implement AND document every single one of these controls from scratch, then defend them to an ISSM.

**AI Doesn't Change the Compliance Timeline:**
AI can write code faster. It cannot:
- Make an AO sign faster
- Accelerate DISA STIG review timelines
- Compress the C3PAO assessment schedule for CMMC
- Eliminate the human judgment required for risk acceptance decisions
- Generate legally defensible control implementation evidence

The regulatory process is the bottleneck. Aegis shortcuts that process by arriving pre-hardened with documented controls.

### The Risk Calculation for a Prime

A prime like Booz Allen, SAIC, or Leidos asks: "Do we spend $600K-$2M/year on 3-5 engineers to build and maintain a GPU platform, plus take on the ATO risk if they miss a control, plus accept the schedule risk of 6-12 months before any data scientist can use it? Or do we license Aegis for $150K-$300K/year, deploy in a week, and get pre-documented controls our ISSM can verify?"

The license is cheaper AND lower risk. That's why they'll buy it even if they technically could build it.

---

## TCO Comparison: Aegis vs DIY (3-Year, 50 GPU Users, 3 Clusters)

Source: competitive-analysis-kubeflow.md — validated against real NTConcepts/DoD infrastructure costs.

| Cost Category | With Aegis | DIY (Kubeflow + Workspaces) |
|---|---|---|
| Initial setup labor | **$50K** (1 engineer, 2 weeks) | **$600K** (3 engineers, 6 months) |
| Annual platform ops (3yr) | **$150K** ($50K/yr, 1 part-time eng) | **$2.1M** ($700K/yr, 3-5 engineers) |
| GPU/cloud infra (3yr) | **$350K** (direct GPU cost) | **$1.2M** (GPU + TGW + NFW + Workspaces + AD) |
| Compliance prep | **$50K** (templates + built controls) | **$250K** (from scratch) |
| Third-party tools (3yr) | **$0** (FinOps built-in) | **$450K** ($150K/yr Kubecost etc.) |
| **3-Year customer engineering cost** | **~$600K** | **~$4.6M** |
| **Customer engineering savings** | | **~$4.0M (87%)** |

Even with an Aegis license at $300K/year, the customer's total cost (license + engineering) is **$1.5M** vs **$4.6M** for DIY — still **67% savings**.

---

## Value Displacement Math

What Aegis replaces in the customer's environment:

| What Customer Avoids | Annual Cost |
|---|---|
| 3-5 DevSecOps platform engineers (fully loaded: salary + benefits + clearance premium + overhead) | $600K-$2M/year |
| 3-6 months of setup time (opportunity cost of delayed mission) | Unquantifiable but real |
| Transit Gateway + Network Firewall + Workspaces + AD + Kubeflow assembly | $400K+ first year |
| Ongoing maintenance, patching, compliance documentation | $200K-$400K/year |
| ATO risk if DIY team misses a control | Program schedule risk |

**Pricing rule from defense-sales-guide.md:** Price at 10-30% of the labor cost displaced.

At 10-30% of $600K-$2M/year labor displacement: **$60K-$600K/year** depending on scale.

---

## Revised Pricing Tiers (Value-Based)

| Tier | Annual Price | Engineers | Clusters | Value Justification | Procurement Vehicle |
|---|---|---|---|---|---|
| **Evaluation** | $9,500 (90 days) | Up to 5 | 1 | GPC-purchasable door opener. Zero procurement friction. | Government Purchase Card |
| **Program** | $75K-$150K/year | Up to 25 | 1 | Replaces 1-2 platform engineers ($200K-$400K/yr savings) | Simplified Acquisition |
| **Enterprise** | $250K-$500K/year | Up to 100 | Up to 5 | Replaces 3-5 platform engineers ($600K-$2M/yr savings) | Contract / OTA |
| **Enterprise Unlimited** | $500K-$1M/year | Unlimited | Unlimited | Full org deployment, replaces entire DIY platform team | Enterprise Agreement |

### Previous Pricing (Too Low)
| Tier | Old Price | Issue |
|---|---|---|
| Starter | $48K/year | Below the defense-sales-guide.md floor of $75K for single cluster |
| Team | $120K/year | Below the $250K floor for multi-cluster |

### Why the Evaluation Tier at $9,500 is Strategic
- Under $10,000 = Government Purchase Card eligible
- No competition required, no contracting officer involvement
- A contracting officer can buy this TODAY with their card
- Gets Aegis into the customer's environment with zero procurement friction
- Conversion to Program tier after successful 90-day evaluation

---

## Pricing by Customer Segment

| Segment | Recommended Tier | Price Range | Sales Motion |
|---|---|---|---|
| Defense tech company (Anduril, Shield AI, Scale AI) | Program or Enterprise | $75K-$500K/year | Direct sale, PO, 30-60 day cycle |
| Prime subcontractor (Booz Allen, SAIC, GDIT) | Enterprise or Unlimited | $250K-$1M/year | OEM/embed license, 6-12 month cycle |
| DoD program office (direct) | Evaluation → Program | $9.5K → $75K-$150K/year | GPC pilot → simplified acquisition |
| SBIR Phase I | N/A (funded deployment) | $50K-$75K grant | Deploy product at partner site |
| SBIR Phase II | Program + services | $500K-$1.7M over 24 months | Full deployment + integration |
| OTA Prototype (DIU, NSTXL) | Enterprise + services | $500K-$5M | Prototype agreement with production follow-on |

---

## Key Talking Points for Sales Conversations

**For technical leads:**
"Your team is spending 3-6 months assembling Transit Gateway, Network Firewall, Workspaces, and Kubeflow — just to give a data scientist a GPU. Aegis gives them VS Code on GPU hardware in 90 seconds, with NIST 800-53 and 800-171 controls built in. Deploy in a week, not six months."

**For program managers:**
"Your platform engineering team costs $600K-$2M a year to build and maintain GPU infrastructure. Aegis replaces that at a fraction of the cost, and your engineers go back to mission work instead of infrastructure maintenance."

**For ISSMs/CISOs:**
"Aegis deploys inside your existing ATO boundary with pre-documented NIST 800-53 control implementations, OSCAL evidence export, FIPS-validated cryptography, and structured audit logging. Your ISSM can verify our controls in weeks, not months."

**For contracting officers:**
"Evaluation license is under $10K — GPC-eligible, no competition required. Get it deployed this month, evaluate for 90 days, then we talk about a program license."

---

---

## The Innovation Narrative — Why Aegis Exists

### The Problem Nobody Else Has Solved

Today, if a DoD data scientist needs a GPU to train a model inside an ATO boundary, there are exactly three options:

**Option 1: DIY stack.** The program office's platform team spends 3-6 months assembling Transit Gateway, Network Firewall, AWS Workspaces, Active Directory, and Kubeflow across 4+ AWS accounts. The data scientist gets browser-based Jupyter on a remote desktop. It costs $600K-$2M/year in platform engineering labor. Every program that needs GPUs builds this from scratch because no productized solution exists for regulated environments.

**Option 2: Commercial MLOps (Databricks, ClearML, Run:ai).** These are SaaS-first products designed for commercial companies. They don't deploy inside an ATO boundary. They don't have NIST 800-53 control implementations. They don't support air-gapped environments. They don't do IL-4/5. A DoD program can't use them without a separate FedRAMP authorization process that takes 12-18 months — and most of these vendors haven't pursued FedRAMP because the DoD market isn't big enough for them to justify it.

**Option 3: Don't use GPUs.** Many programs just don't bother because the infrastructure overhead is too high. Missions that need ML/AI capabilities get delayed or deprioritized because the compute infrastructure doesn't exist in their security boundary.

### Aegis Creates Option 4

A self-hosted, compliance-first GPU control plane that deploys inside the customer's ATO boundary in one week via Helm chart, gives data scientists native VS Code (not browser Jupyter) on GPU hardware in 90 seconds, and comes with NIST 800-53/800-171/CMMC controls pre-implemented with OSCAL evidence export. No FedRAMP needed because it's self-hosted. No 3-6 month build because it's a product. No degraded developer experience because Sovran gives you your actual IDE, not a remote desktop.

### Why Sovran is the Differentiator

Without Sovran, Aegis is "just another Kubernetes platform with compliance controls" — useful but not differentiated enough. With Sovran, Aegis delivers something nobody else in defense or commercial has: a native VS Code IDE experience on remote GPU hardware with single-use JWT session tokens, RAM-disk sandboxing for CUI protection, clipboard wipe on disconnect, and automatic session revocation on inactivity. The data scientist doesn't know or care about the infrastructure — they click Connect in VS Code and they're on a GPU. That developer experience, inside a security boundary this strict, does not exist anywhere else.

### The Innovation for SBIR/CSO Evaluators

The innovation is not any single technology component. The innovation is a systems integration that eliminates a 3-6 month infrastructure deployment timeline that currently blocks DoD AI readiness, while simultaneously solving the developer experience problem (browser Jupyter vs native IDE) and the compliance evidence problem (manual spreadsheets vs automated OSCAL export) — all in a self-hosted package that requires zero FedRAMP authorization. Each of these problems has been solved individually in commercial environments. Nobody has solved all three simultaneously inside a DoD ATO boundary as a deployable product.

---

## Government Procurement Vehicles — What They Are and What They Get Us

### Types of Opportunities (Not All Are "Contracts")

| Vehicle | What It Is | How It Works | What You Get |
|---|---|---|---|
| **CSO (Commercial Solutions Opening)** | An OTA mechanism for commercial/nontraditional companies | Submit a solution brief (5-10 pages). If selected, negotiate a prototype agreement. | Funded prototype ($500K-$5M), sole-source production follow-on |
| **BAA (Broad Agency Announcement)** | Open call for white papers on research areas | Submit a white paper describing your capability. Agency funds interesting ones. | Research funding, government relationship |
| **SBIR (Small Business Innovation Research)** | Non-dilutive grant for small businesses | Phase I ($50-275K, 6 months) → Phase II ($500K-$1.7M, 24 months) → Phase III (unlimited, sole-source) | Funding to deploy your product + sole-source authority |
| **OTA (Other Transaction Authority)** | Prototype agreement outside traditional FAR procurement | Faster, less paperwork than FAR contracts. Designed for commercial tech. | Revenue + path to production contract |

**Aegis qualifies for all of these because:**
- We're a nontraditional defense contractor (enables OTA authority)
- We're a small business (enables SBIR)
- We have a working product at TRL 6+ (what DIU and CSOs want)
- We're a product company, not a services company (exactly what these mechanisms are designed to pull into DoD)

### These Are NOT Customer Acquisitions — They're Funded Pilots

Submitting to CSOs/BAAs/SBIRs does not get customers. It gets funded pilots with government partners. The distinction matters:

- A **customer** pays you for a product license and uses it in production
- A **funded pilot** is the government paying you to deploy your product on their infrastructure and prove it works

The funded pilot is what converts into a customer. Here's the chain:

### Best-Case Scenario: Kessel Run CSO

1. Submit solution brief → Kessel Run reviews (30-90 days)
2. Invited to present/demo → You show VS Code on GPU in 90 seconds
3. OTA prototype agreement signed ($500K-$5M)
4. Deploy Aegis on Kessel Run's cluster (2-4 weeks)
5. Their data scientists use it for 3-6 months
6. Prototype succeeds → Sole-source production follow-on contract
7. **Result:** Revenue, deployed product in real DoD environment, reference customer

That reference customer unlocks everything else — primes like Booz Allen and SAIC take your call when you can say "Kessel Run is using us."

### Best-Case Scenario: SBIR Phase I → II → III

1. AFWERX awards Phase I ($50-75K, 6 months) → Deploy on AF partner cluster
2. Phase I succeeds → Submit Phase II proposal ($500K-$1.7M, 24 months)
3. Phase II: Deploy to 3-5 AF programs, get Iron Bank images approved, production hardening
4. Phase III: **Unlimited value, sole-source.** Any DoD organization can buy from you without competition.

Phase III is how companies like SpaceX, Palantir, and Joby Aviation scaled early DoD revenue.

### Best-Case Scenario: DIU CSO

1. Submit 5-page solution brief at diu.mil → DIU reviews (60-90 days)
2. Prototype agreement ($500K-$5M) → Deploy at DoD partner site
3. Prototype succeeds → DIU can sole-source production contract ($5M-$50M+)

### Why Pursue Both CSOs AND Direct Sales

| Path | Timeline to Revenue | Deal Size | Requires SAM.gov? |
|---|---|---|---|
| Defense tech company (Anduril, Shield AI) | 3-6 months | $75K-$250K/yr | No |
| CSO/OTA (Kessel Run, DIU) | 9-18 months | $500K-$5M | For award, not submission |
| SBIR Phase I | 6-9 months | $50-75K | Yes |

Direct sales to defense tech companies give you fast revenue and proof points. CSOs/SBIRs give you government reference customers and sole-source authority. They feed each other — a pilot at Shield AI helps your Kessel Run pitch, and a Kessel Run prototype helps your pitch to Booz Allen.

### Key Insight: Some Submissions Don't Need SAM.gov

Kessel Run CSO and DIU CSO accept submissions without SAM.gov registration. SAM.gov is only required before award. This means Aegis can submit solution briefs to its two highest-fit opportunities TODAY while SAM.gov processes in the background.

---

*Last updated: March 2026. Pricing subject to adjustment based on customer discovery feedback.*
