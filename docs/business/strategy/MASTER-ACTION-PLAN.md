# Aegis Platform: Master Government Market Action Plan

> **One document to rule them all.** This synthesizes the detailed guides in `docs/business/` into a single ordered action plan for a 1-person company entering the US government defense market.

**Detailed guides** (read these for specifics):
- [Gov Contracting Setup](gov-contracting-setup-guide.md) - Entity, SAM.gov, certifications, insurance
- [SBIR/STTR Guide](sbir-sttr-guide.md) - Grant applications, agencies, proposal writing
- [DIU & OTA Guide](diu-ota-guide.md) - Defense Innovation Unit, prototype agreements
- [Platform One & Iron Bank](platform-one-iron-bank-guide.md) - DoD container registry, Big Bang integration
- [Defense Sales Guide](defense-sales-guide.md) - Selling to primes and defense tech companies
- [DISA & Cyber Guide](disa-cyber-guide.md) - CMMC, STIG, compliance documentation, CDAO

---

## The Strategy in One Sentence

**Register as a government contractor (SAM.gov), submit SBIR proposals for non-dilutive funding (AFWERX first), engage DIU for prototype contracts, harden images for Iron Bank/Platform One, and sell directly to defense tech companies - all in parallel.**

---

## Month 1: Foundation (Cost: ~$3,000-$5,000)

### Week 1-2: Business Setup
- [ ] **Confirm LLC is properly formed** (or form one)
- [ ] **Get EIN from IRS** (free, instant at irs.gov)
- [ ] **Open business bank account** (need EIN + formation docs)
- [ ] **Start SAM.gov registration** - This is the #1 bottleneck (takes 2-6 weeks)
  - Need: Legal name, EIN, physical address, bank info
  - Select NAICS codes: 511210 (primary), 541511, 541512, 518210, 541715
  - Self-certify as small business

### Week 2-3: Insurance
- [ ] **Get General Liability** ($1M/$2M) - ~$500-$1,200/year
- [ ] **Get Professional Liability (E&O)** ($1M) - ~$800-$2,000/year
- [ ] **Get Cyber Liability** ($1-$2M) - ~$1,000-$3,000/year
- Providers: Hiscox, Hartford, NEXT Insurance

### Week 3-4: Registration Blitz
- [ ] **Register on SBIR.gov** (company registry)
- [ ] **Register on DSIP** (dodsbirsttr.mil - DoD SBIR portal)
- [ ] **Register on DIU portal** (submissions.diu.mil)
- [ ] **Register on Tradewinds** (tradewindai.com - CDAO marketplace)
- [ ] **Register on Grants.gov** (for NSF/DOE SBIRs)
- [ ] **Create Login.gov account** (authentication for federal systems)
- [ ] **Create Repo One account** (repo1.dso.mil - Platform One/Iron Bank)

### Week 4: CMMC Level 1
- [ ] **Complete CMMC Level 1 self-assessment** (17 practices from FAR 52.204-21)
- [ ] **Upload SPRS score** at sprs.csd.disa.mil
- [ ] Takes 1-2 days of actual work

---

## Month 2: First Submissions (Cost: $0 - your time)

### SBIR Track
- [ ] **Browse current solicitations** on DSIP and SBIR.gov
- [ ] **Identify 3-5 target topics** related to: AI/ML infrastructure, DevSecOps, GPU scheduling, secure computing, Kubernetes for DoD
- [ ] **Contact TPOCs** (Technical Points of Contact) for top 2-3 topics
- [ ] **Start writing AFWERX Open Topic proposal** (5-10 pages, rolling deadline)
  - This is the easiest first SBIR: shorter proposal, faster decisions, $50K-$75K Phase I
  - Frame Aegis as: "Secure multi-cluster GPU workload scheduling for DoD AI programs"

### DIU Track
- [ ] **Review all current DIU Areas of Interest** at diu.mil/work-with-us
- [ ] **Draft 5-page solution brief** for the most relevant AOI
- [ ] **Submit to DIU CSO portal**

### Tradewinds Track
- [ ] **Submit Aegis solution to Tradewinds marketplace**

### Direct Sales Track
- [ ] **Create capability statement** (1-page PDF: company, CAGE code, NAICS, capabilities, differentiators, contact)
- [ ] **Create website security/compliance page** (FIPS, STIG, NIST 800-53 capabilities)
- [ ] **Connect with 20+ defense tech professionals on LinkedIn**
- [ ] **Register on 3-5 prime vendor portals** (Lockheed, RTX, GDIT, Booz Allen, SAIC)

---

## Month 3: Technical Hardening (Cost: $0 - dev time)

### Iron Bank Preparation
- [ ] **Rebuild all images on Iron Bank-approved base** (UBI9 minimal from registry1.dso.mil)
- [ ] **Ensure all containers run as non-root** with read-only filesystem
- [ ] **Generate SBOMs** for all images (Syft/CycloneDX)
- [ ] **Scan for CVEs** and remediate all critical/high
- [ ] **Test FIPS mode end-to-end** (`GODEBUG=fips140=only`)

### Compliance Documentation
- [ ] **Create DISA Kubernetes STIG compliance matrix**
- [ ] **Create NSA/CISA Kubernetes Hardening Guide compliance matrix**
- [ ] **Build FIPS 140 cryptographic inventory** (BoringCrypto cert #, OpenSSL FIPS version)
- [ ] **Generate OSCAL component definition** for Aegis
- [ ] **Package complete ATO support documentation**:
  - Customer Responsibility Matrix (exists)
  - Control Implementation Statements (exists)
  - Hardening Guide (exists, complete it)
  - STIG Compliance Checklist (new)
  - SBOM (generate)
  - Vulnerability scan results (generate)

---

## Month 4-5: Submit and Engage (Cost: $1,000-$3,000 for conferences)

### SBIR Submissions
- [ ] **Submit AFWERX Open Topic proposal** (if not already done)
- [ ] **Submit to 2-3 additional SBIR topics** from next DoD solicitation cycle
- [ ] **Consider NSF SBIR** ($275K Phase I, very startup-friendly)
- [ ] **Consider Direct-to-Phase-II (D2P2)** if agency offers it (you already have a working platform)

### Iron Bank Submission
- [ ] **Submit Dockerfiles and hardening docs to Iron Bank** via repo1.dso.mil
- [ ] **Iterate on feedback** (expect 1-4 weeks of remediation)

### Networking
- [ ] **Attend first defense tech conference** (pick ONE: AUSA, TechNet, or a free NSIN/AFWERX event)
- [ ] **Send 10-15 targeted LinkedIn outreach messages** to defense tech companies
- [ ] **Schedule 3-5 intro calls/demos** with interested companies
- [ ] **Identify NSIN Propel or AFWERX accelerator** applications

### Content Marketing
- [ ] **Write first white paper**: "Secure Multi-Cluster GPU Scheduling for Classified AI Workloads"
- [ ] **Start LinkedIn posting** (1-2x/week): DevSecOps insights, Kubernetes security, DoD modernization

---

## Month 6: Revenue Pursuit (Cost: $0-$2,000)

### Fastest Path to First Dollar
The most likely paths to first revenue, ranked by speed:

| Path | Timeline | Revenue | Probability |
|------|----------|---------|-------------|
| **Defense tech company pilot** (Tier 2: Anduril, Shield AI, etc.) | 3-6 months | $25K-$150K | Medium |
| **AFWERX SBIR Phase I** | 2-6 months from submission | $50K-$75K | 20-30% per submission |
| **Government purchase card pilot** ($9,500 license) | 1-3 months | $9,500 | Medium (if you have a contact) |
| **DIU prototype OT** | 3-6 months from submission | $500K-$5M | 10-20% |
| **NSF SBIR Phase I** | 6-8 months from submission | $275K | 20-25% |
| **Prime subcontractor** | 6-12 months | $50K-$200K | Low-Medium |

### Parallel Actions
- [ ] **Follow up on all SBIR submissions**
- [ ] **Follow up on DIU submission** (if no response)
- [ ] **Convert 1-2 demos to pilot evaluations**
- [ ] **Apply to accelerator** (NSIN Propel, Capital Factory defense, AFWERX)
- [ ] **Begin SBA certification applications** (8(a) if eligible - this is the highest-ROI certification)

---

## Month 7-12: Scale What Works

### If SBIR wins:
- Execute Phase I work plan
- Plan Phase II proposal (due near end of Phase I)
- Phase II = $500K-$1.7M over 24 months

### If DIU prototype wins:
- Execute prototype milestones
- Build DoD user champion
- Start ATO documentation prep for production follow-on
- Production OT can be $5M-$50M+ (sole-source)

### If direct sales work:
- Close first enterprise deal ($75K-$250K/year)
- Build reference account
- Use for past performance in future proposals
- Approach primes with proven deployment

### Technical Evolution
- [ ] **Test Aegis on RKE2** (Platform One's preferred K8s) if not done
- [ ] **Big Bang compatibility testing** (Istio, Prometheus, Fluentd)
- [ ] **Get images approved in Iron Bank**
- [ ] **Begin CMMC Level 2 preparation** if contracts require it

---

## Year 2: Compound Growth

### With 1-2 reference customers:
- Apply for GSA Schedule (MAS) - opens government-wide purchasing
- Pursue 8(a) certification if eligible - sole-source contracts up to $4.5M
- Engage defense primes as subcontractor with proven deployments
- Submit for Platform One addon evaluation
- Attend 3-4 conferences/year

### Revenue Targets
| Year | Conservative | Moderate | Aggressive |
|------|-------------|----------|-----------|
| Year 1 | $50K-$150K (1 SBIR + 1 pilot) | $200K-$400K | $500K+ |
| Year 2 | $200K-$500K | $500K-$1M | $1M-$2M |
| Year 3 | $500K-$1M | $1M-$3M | $3M-$5M |

---

## Budget Summary

### Startup (First 6 Months)

| Category | Amount |
|----------|--------|
| LLC formation + legal | $100-$3,000 |
| Insurance (6 months) | $1,500-$3,500 |
| SAM.gov + registrations | Free |
| CMMC Level 1 | Free (your time) |
| SBIR proposals | Free (your time) |
| DIU submissions | Free (your time) |
| 1 conference attendance | $1,000-$3,000 |
| Website/marketing | $500-$2,000 |
| **Total** | **$3,000-$12,000** |

### Annual Recurring

| Category | Amount |
|----------|--------|
| Insurance | $3,000-$7,000 |
| State LLC fees | $0-$800 |
| Accounting/bookkeeping | $2,000-$5,000 |
| Conferences (2-3/year) | $3,000-$8,000 |
| **Total Annual** | **$8,000-$21,000** |

### When You Win First Contract

| Category | Amount |
|----------|--------|
| Government contract accountant | $2,000-$5,000 |
| CMMC Level 2 (if required) | $30,000-$75,000 |
| Legal review | $2,000-$5,000 |

---

## Critical Rules

1. **Submit to everything relevant.** Success rates are 15-25%. Volume is how you win.
2. **AFWERX Open Topic is your first SBIR.** Shortest proposal, fastest decisions, rolling deadline.
3. **SAM.gov is the foundation.** Nothing happens without it. Start registration day 1.
4. **Sell to defense TECH companies first** (Tier 2). They buy faster than primes or government.
5. **Relationships before transactions.** Defense sales take 6-18 months. Start networking now.
6. **FIPS crypto is non-negotiable.** Without FIPS, nothing in DoD is credible. Do this in Month 3.
7. **Iron Bank images = distribution.** Get your images in Iron Bank and DoD programs can consume them.
8. **Don't build for FedRAMP.** You're self-hosted. Redirect compliance effort to STIG/FIPS.
9. **8(a) certification is the highest-ROI action** if you qualify. Sole-source contracts up to $4.5M.
10. **Phase III SBIR contracts are unlimited and sole-source.** Win Phase I → Phase II → Phase III is a path to significant recurring revenue.

---

## Key Portals Quick Reference

| Portal | URL | What For |
|--------|-----|---------|
| SAM.gov | sam.gov | Contractor registration (MANDATORY) |
| SBIR.gov | sbir.gov | SBIR/STTR proposals |
| DSIP | dodsbirsttr.mil | DoD SBIR submission |
| AFWERX | afwerx.com/sbir-sttr | Air Force Open Topic SBIR |
| DIU | diu.mil/work-with-us | Commercial Solutions Opening |
| Tradewinds | tradewindai.com | CDAO AI marketplace |
| Iron Bank | ironbank.dso.mil | DoD container registry |
| Repo One | repo1.dso.mil | Platform One source code |
| SPRS | sprs.csd.disa.mil | CMMC score upload |
| NSF Seed Fund | seedfund.nsf.gov | NSF SBIR ($275K Phase I) |
| Login.gov | login.gov | Federal authentication |
| SBA Certify | certify.sba.gov | 8(a), HUBZone, WOSB |
| STIGs | public.cyber.mil/stigs | DISA security standards |

---

*Last updated: March 2026. This plan assumes a 1-person company that can grow as revenue comes in. Adjust timelines based on your bandwidth and which opportunities materialize first.*
