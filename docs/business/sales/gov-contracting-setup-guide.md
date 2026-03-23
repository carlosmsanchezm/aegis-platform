# US Government Contracting Setup Guide -- 1-Person Software Company (2025-2026)

> **Purpose**: Step-by-step checklist to become eligible for US federal government contracting as a solo-founder software company selling AI/ML, Kubernetes platform, DevSecOps, and cybersecurity products.
>
> **Last updated**: March 2026. Verify all URLs, fees, and thresholds before acting -- government rules change frequently.

---

## Table of Contents

1. [Recommended Order of Operations](#1-recommended-order-of-operations)
2. [Business Entity Requirements](#2-business-entity-requirements)
3. [SAM.gov Registration](#3-samgov-registration)
4. [CAGE Code](#4-cage-code)
5. [Small Business Certifications](#5-small-business-certifications)
6. [Security Clearances](#6-security-clearances)
7. [Insurance and Bonds](#7-insurance-and-bonds)
8. [GSA Schedule / Multiple Award Schedule](#8-gsa-schedule--multiple-award-schedule-mas)
9. [Cost Accounting Standards](#9-cost-accounting-standards)
10. [Estimated Costs](#10-estimated-costs)
11. [Key NAICS Codes](#11-key-naics-codes)
12. [Ongoing Maintenance](#12-ongoing-maintenance)
13. [Where to Find Opportunities](#13-where-to-find-opportunities)

---

## 1. Recommended Order of Operations

Complete these in sequence. Some steps run in parallel once started.

| Phase | Step | Timeline | Notes |
|-------|------|----------|-------|
| **Phase 1: Foundation** | | **Weeks 1-4** | |
| | 1. Form business entity (LLC or S-Corp) | 1-2 weeks | See Section 2 |
| | 2. Get EIN from IRS | Same day (online) | Free, takes 5 minutes |
| | 3. Open business bank account | 1-3 days | Need EIN + formation docs |
| | 4. Get business address (if using home, that works) | Immediate | PO Box not accepted for SAM |
| **Phase 2: Federal Registration** | | **Weeks 2-8** | |
| | 5. Register on SAM.gov (includes UEI assignment) | 2-6 weeks for approval | Free. Start early -- this is the bottleneck |
| | 6. CAGE code auto-assigned during SAM registration | Included in SAM | No separate action needed |
| | 7. Identify your NAICS codes | During SAM registration | See Section 11 |
| **Phase 3: Insurance** | | **Weeks 3-6** | |
| | 8. Get General Liability insurance | 1-2 days | ~$500-1,200/year |
| | 9. Get Professional Liability (E&O) insurance | 1-2 days | ~$800-2,000/year |
| | 10. Get Cyber Liability insurance | 1-2 days | ~$1,000-3,000/year |
| **Phase 4: Certifications** | | **Weeks 4-24** | |
| | 11. Self-certify as Small Business in SAM.gov | During SAM registration | Free, immediate |
| | 12. Apply for any applicable set-aside certifications | 4-12 weeks each | 8(a), SDVOSB, WOSB, HUBZone |
| **Phase 5: Start Bidding** | | **Week 8+** | |
| | 13. Begin monitoring contract opportunities | Ongoing | SAM.gov, GovWin, etc. |
| | 14. Respond to RFIs and small RFQs | Ongoing | Build past performance |
| **Phase 6: Optional Growth** | | **Months 6-18** | |
| | 15. Apply for GSA MAS Schedule (optional) | 6-18 months | Consider after first contract wins |
| | 16. Pursue facility/personnel clearance (if needed) | 6-18 months | Only when a contract requires it |

---

## 2. Business Entity Requirements

### Which Entity Type?

| Entity | Pros | Cons | Verdict |
|--------|------|------|---------|
| **LLC (taxed as S-Corp)** | Simple formation, liability protection, pass-through taxation, flexible. Can elect S-Corp tax treatment via IRS Form 2553. | Some agencies prefer "Inc." for optics (but legally irrelevant). | **Best choice for a 1-person company.** |
| **S-Corp** | Pass-through taxation, payroll tax savings on distributions above reasonable salary. | More administrative burden (payroll, reasonable salary requirement). | Good, but LLC with S-Corp election gives same tax benefit. |
| **C-Corp** | Preferred by VCs if you plan to raise. Unlimited shareholders. | Double taxation. More complex. Overkill for 1-person gov contracting. | Only if you plan to raise venture capital. |

**Recommendation**: Form an **LLC** and elect S-Corp tax treatment (IRS Form 2553) once revenue exceeds ~$40-50K/year. This gives you liability protection, pass-through taxation, and payroll tax savings.

### State of Incorporation

| Option | Pros | Cons |
|--------|------|------|
| **Your home state** | Simplest. No foreign qualification needed. Lower total costs. | State-specific tax rules apply. |
| **Delaware** | Business-friendly courts, well-established case law, privacy. | Must also register as foreign entity in your home state (double fees). |
| **Wyoming** | No state income tax, low fees, strong privacy. | Same foreign qualification issue if you live elsewhere. |

**Recommendation**: **Incorporate in your home state** unless you have a specific legal reason for Delaware. For a 1-person gov contracting company, the simplicity outweighs any Delaware advantages. If you are in a state with no income tax (TX, FL, WA, WY, NV, etc.), you already have the tax benefit.

### EIN (Employer Identification Number)

- **What**: Federal tax ID for your business. Required for SAM.gov, bank accounts, and everything else.
- **How**: Apply online at [IRS EIN Application](https://www.irs.gov/businesses/small-businesses-self-employed/apply-for-an-employer-identification-number-ein-online). Free. Takes 5-10 minutes.
- **Timeline**: Immediate -- you get the EIN number at the end of the online session.
- **Cost**: Free.

---

## 3. SAM.gov Registration (System for Award Management)

SAM.gov is the **mandatory gateway** to all federal contracting. No SAM registration = no federal contracts. Period.

### What You Get From SAM Registration

- **UEI (Unique Entity Identifier)**: Replaced DUNS number in April 2022. Assigned automatically during SAM registration. You no longer need to go to Dun & Bradstreet separately.
- **CAGE Code**: Assigned automatically during SAM registration (or within a few days after).
- **Active entity status**: Makes you searchable by contracting officers.

### Step-by-Step Registration Process

1. **Create a Login.gov account** at [login.gov](https://login.gov)
   - Use your business email
   - Set up MFA (required)

2. **Go to [SAM.gov](https://sam.gov)** and sign in with Login.gov

3. **Start entity registration** -- Click "Get Started" under Entity Registration

4. **Provide required information**:

   | Field | What to Enter |
   |-------|---------------|
   | Legal Business Name | Your LLC name exactly as filed with the state |
   | Physical Address | Must be a physical street address (no PO Box) |
   | Mailing Address | Can be different from physical |
   | Start Date | Date of LLC formation |
   | State of Incorporation | Where you filed |
   | Fiscal Year End | Typically 12/31 for calendar year |
   | EIN | Your IRS EIN |
   | Entity Type | Select "Business or Organization" then appropriate subtype |
   | Congressional District | Based on your physical address |

5. **Select NAICS codes** (see Section 11 for recommendations)

6. **Select SBA size standard** -- self-certify as small business

7. **Provide banking information** for electronic funds transfer (EFT)
   - Bank routing number
   - Account number
   - Account type (checking)

8. **Enter Points of Contact**:
   - Government Business POC (you)
   - Electronic Business POC (you)
   - Past Performance POC (you, or N/A if new)

9. **Complete representations and certifications**
   - Annual representations and certifications (FAR 52.204-8)
   - Answer questions about ownership, size, certifications
   - Most are straightforward yes/no for a new small business

10. **Submit and wait for validation**
    - IRS TIN validation: 2-5 business days
    - CAGE code assignment: 2-5 business days
    - Total approval: typically **2-6 weeks** (can take up to 10-12 weeks in backlogs)

### Timeline

| Step | Duration |
|------|----------|
| Gather documents | 1-2 days |
| Complete online forms | 2-4 hours |
| IRS TIN validation | 2-5 business days |
| CAGE code assignment | 2-5 business days |
| Full registration approval | 2-6 weeks (occasionally longer) |

### Annual Renewal

- SAM.gov registration must be **renewed annually** (365 days from activation)
- You will receive email reminders starting 60 days before expiration
- Renewal is free but takes 1-2 weeks for re-validation
- **If your registration lapses, you cannot receive contract awards or payments**

### Cost

- **Free.** There is no fee for SAM.gov registration.
- **WARNING**: Scam companies will contact you offering to "register" you on SAM.gov for $400-800. These are scams. SAM.gov registration is always free and self-service.

---

## 4. CAGE Code (Commercial and Government Entity Code)

### What Is It?

A **5-character alphanumeric identifier** assigned to entities doing business with the US federal government. Used to identify your company in the defense and federal procurement systems.

### How to Get One

- **You do NOT need to apply separately.** A CAGE code is automatically assigned during SAM.gov registration.
- If you need one before SAM registration (rare), you can request one from DLA (Defense Logistics Agency) via the [CAGE program](https://cage.dla.mil).

### Timeline

- Auto-assigned during SAM.gov processing: **2-5 business days** after submission.
- Standalone request: **3-5 business days**.

### Cost

- **Free.**

---

## 5. Small Business Certifications

### SBA Size Standards for Software Companies

The SBA determines "small business" status by **NAICS code**. For software and IT services:

| NAICS Code | Description | Size Standard (2025-2026) |
|------------|-------------|---------------------------|
| 511210 | Software Publishers | $47 million annual revenue |
| 518210 | Computing Infrastructure, Data Processing | $40 million annual revenue |
| 541511 | Custom Computer Programming | $34 million annual revenue |
| 541512 | Computer Systems Design | $34 million annual revenue |
| 541519 | Other Computer Related Services | $30 million annual revenue |
| 541715 | R&D (Physical, Engineering, Life Sciences) | 1,000 employees |
| 561621 | Security Systems Services | $25 million annual revenue |

**As a 1-person company, you automatically qualify as small business under all of these.** Self-certify during SAM.gov registration.

> **Note**: Size standards are updated periodically. Verify current thresholds at [SBA Size Standards Table](https://www.sba.gov/federal-contracting/contracting-guide/size-standards).

### Certification Overview

| Certification | What It Is | Eligibility | Worth It? | Timeline |
|---------------|-----------|-------------|-----------|----------|
| **Small Business** | Base certification | Revenue under size standard threshold | **Yes -- mandatory** | Immediate (self-certify in SAM) |
| **SDB (Small Disadvantaged Business)** | For businesses owned by socially/economically disadvantaged individuals | 51%+ owned by qualifying individuals (includes certain racial/ethnic minorities, per SBA presumptions) | **Yes if you qualify** -- 10% price evaluation preference | Self-certify in SAM.gov (as of 2023 rule change) |
| **8(a) Business Development** | 9-year business development program | Must qualify as SDB, owner must demonstrate social + economic disadvantage, business operational 2+ years, below $400K personal net worth (excl. home + business) | **Highly valuable if eligible** -- sole-source contracts up to $4.5M for services | 3-6 months for approval |
| **WOSB / EDWOSB** | Woman-Owned Small Business | 51%+ owned/controlled by women | **Yes if eligible** -- set-asides in underrepresented industries | Self-certify or use SBA-approved certifier |
| **SDVOSB / VOSB** | Service-Disabled Veteran-Owned Small Business | Owner is a service-disabled veteran (any % disability rating) | **Very valuable** -- significant set-asides especially at VA/DoD | Apply via SBA VetCert portal, 2-4 months |
| **HUBZone** | Historically Underutilized Business Zone | Principal office in a HUBZone AND 35%+ of employees reside in HUBZone | **Possible for 1-person** if you live and work in a HUBZone | 2-3 months for approval |

### Deep Dive: 8(a) Business Development Program

The 8(a) program is the **single most powerful certification** for winning government contracts as a small business.

**Benefits**:
- Sole-source contracts up to **$4.5 million** (services) or **$7 million** (manufacturing) -- agencies can award directly to you without competition
- Competitive 8(a) set-aside contracts (only compete against other 8(a) firms)
- Mentoring from large businesses (mentor-protege program)
- Management and technical assistance
- 9-year program term (4-year developmental stage + 5-year transitional stage)

**Eligibility Requirements**:
- Must be a small business per SBA size standards
- At least 51% unconditionally owned/controlled by one or more socially and economically disadvantaged individuals
- Owner must be a US citizen
- Owner must demonstrate **social disadvantage** (SBA presumes this for: Black Americans, Hispanic Americans, Native Americans, Asian Pacific Americans, Subcontinent Asian Americans; others can prove it with a preponderance of evidence via personal narrative)
- Owner must demonstrate **economic disadvantage**: adjusted net worth under **$850,000** (excluding primary residence and business value), with total assets under $6.5M and 3-year average AGI under $400K (thresholds may change; verify with SBA)
- Business must have been in operation for at least **2 years** (can be waived in some cases)
- Good character (no recent criminal convictions or debarments)

**Is It Worth Pursuing?**
- **If you qualify, absolutely yes.** The sole-source authority alone makes it the highest-ROI certification in government contracting.
- Many small gov-con companies built their entire early revenue base on 8(a) sole-source contracts.
- Apply once your business has been operating for 2 years and you have some initial capability to demonstrate.

### How to Check HUBZone Eligibility

1. Go to the [SBA HUBZone Map](https://maps.certify.sba.gov/hubzone/map)
2. Enter your address
3. If your home/office is in a qualified HUBZone AND you live there, you may qualify as a 1-person company (since 100% of your "employees" reside in a HUBZone)

---

## 6. Security Clearances

### Can You Do Government Work Without a Clearance?

**Yes, absolutely.** The majority of government IT and software contracts do **not** require security clearances. Unclassified work (including CUI -- Controlled Unclassified Information) does not require a clearance.

You can pursue:
- All civilian agency contracts (GSA, HHS, DOE, etc.)
- Unclassified DoD contracts
- Most cybersecurity work (NIST 800-171 compliance, etc.)
- Commercial cloud services (FedRAMP)
- Open-source and unclassified DevSecOps work

**Only pursue clearances when a specific contract requires it.**

### Facility Clearance (FCL)

| Aspect | Details |
|--------|---------|
| **What** | Authorization for a company to access classified information at a specific location |
| **Who needs it** | Companies working on classified contracts (SECRET, TOP SECRET) |
| **How to get it** | A government agency must **sponsor** you -- you cannot self-apply. A contracting officer submits the sponsorship after you win (or are in process for) a classified contract. |
| **Administering body** | DCSA (Defense Counterintelligence and Security Agency) |
| **Requirements for 1-person company** | You need a Key Management Personnel (KMP) list (just you). You need an approved Facility Security Officer (FSO) -- can be yourself but requires FSO training. |
| **Timeline** | 6-18 months from sponsorship to clearance |
| **Cost** | Government pays for the investigation. Your costs: FSO training (~$1,500-3,000), NISS access, security infrastructure. |

### Personnel Clearance

| Level | Investigation | Timeline | Notes |
|-------|--------------|----------|-------|
| **Confidential** | ANACI or T3 | 2-6 months | Rare; most contracts need SECRET minimum |
| **SECRET** | T3 (Tier 3) | 4-8 months | Most common; covers ~80% of classified work |
| **TOP SECRET** | T5 (Tier 5) | 8-18 months | Includes polygraph for some agencies (CI or FS poly) |
| **TS/SCI** | T5 + SCI eligibility | 12-24 months | For intelligence community work |

### Key Points for a 1-Person Company

1. **You cannot initiate a clearance yourself.** A government agency or cleared contractor must sponsor you.
2. **Strategy**: Win an unclassified contract first. Build relationships. When a classified opportunity arises, the agency can sponsor your clearance.
3. **Interim clearances**: Available for SECRET level. Allows you to start work while the full investigation proceeds. Granted based on preliminary review of your SF-86.
4. **Foreign influence**: Extensive foreign travel, foreign contacts, or foreign financial interests can delay or complicate clearance.

---

## 7. Insurance and Bonds

### Required Insurance

Most government RFPs require proof of insurance. Get these **before** you start bidding.

| Insurance Type | What It Covers | Typical Cost (1-person, software) | Required? |
|----------------|---------------|----------------------------------|-----------|
| **General Liability** | Bodily injury, property damage, personal injury at client sites | $500-1,200/year | Yes -- almost always required. $1M per occurrence / $2M aggregate is standard. |
| **Professional Liability (E&O)** | Errors, omissions, negligence in professional services delivery | $800-2,000/year | Yes -- required for services contracts. $1M minimum. |
| **Cyber Liability** | Data breaches, cyberattacks, privacy violations, incident response | $1,000-3,000/year | Increasingly required, especially for IT/cybersecurity contracts. $1-2M minimum. |
| **Workers' Compensation** | Work-related injuries/illness | $300-800/year | Required in most states even for 1-person (owner-only may be exempt in some states; check your state). |
| **Commercial Auto** | Vehicle accidents during business use | $1,200-2,000/year | Only if you drive for business purposes. |

### Where to Get It

- **Hiscox**: Good for small/solo tech companies. Online quotes.
- **Hartford**: Established carrier, competitive for small business.
- **biBERK** (Berkshire Hathaway): Online-first, competitive pricing for sole proprietors.
- **NEXT Insurance**: Quick online process for small businesses.
- **Coalition**: Specializes in cyber liability.

### Bonding

| Bond Type | When Needed | Notes |
|-----------|-------------|-------|
| **Bid Bond** | When submitting bids for construction or large contracts | Rarely needed for software/IT services |
| **Performance Bond** | When required by contract terms | Uncommon for IT services under $150K |
| **Payment Bond** | For construction contracts over $35K (Miller Act) | Not applicable to software |
| **Fidelity Bond** | When handling government funds or sensitive assets | Sometimes required; ~$200-500/year |

**For a software/IT company**: You likely will **not** need bonds initially. Bonds are primarily required for construction, manufacturing, and large services contracts. If a specific solicitation requires a bond, obtain one at that time through a surety company.

---

## 8. GSA Schedule / Multiple Award Schedule (MAS)

### What Is It?

The GSA MAS is a **long-term government-wide contract** (up to 20 years with option periods). It pre-negotiates prices and terms so government buyers can purchase your products/services more quickly without a full competitive procurement process.

Think of it as getting onto the government's "approved vendor catalog."

### Relevant GSA Schedule Categories

| SIN (Special Item Number) | Category | Relevance |
|---------------------------|----------|-----------|
| 54151S | IT Professional Services | Custom software dev, DevSecOps, consulting |
| 54151HACS | Highly Adaptive Cybersecurity Services | Cybersecurity assessments, incident response |
| 511210 | Software Licenses | SaaS, platform licensing |
| 518210 | Cloud and Cloud-Related IT Services | IaaS, PaaS, cloud migration |
| OLM | Order-Level Materials | Ancillary supplies needed for service delivery |

### Is It Worth Pursuing for a 1-Person Company?

**Not immediately. Defer to Phase 6 (month 6-18+).**

**Pros**:
- Access to a massive buyer pool (all federal agencies)
- Streamlined procurement (buyers prefer GSA vendors)
- ~33% of federal IT spending goes through GSA
- Once you're on, government buyers can place orders directly

**Cons**:
- Application is complex and time-consuming (60-120+ hours of work)
- Requires 2 years of corporate experience (or relevant individual experience)
- Requires at least 1-2 past performance references (can be commercial)
- Requires a Price Proposal with commercial sales documentation
- Legal/consultant costs: $5,000-$15,000 if using a consultant
- Annual reporting requirements (Industrial Funding Fee of 0.75% of sales)
- You must offer the government your "Most Favored Customer" pricing

### How to Get on the GSA Schedule

1. **Prerequisites** (must have before applying):
   - Active SAM.gov registration
   - At least 2 years in business
   - Financial statements (2 years)
   - Past performance references (minimum 2, can be commercial clients)
   - Adequate accounting system
   - Commercial price list or market rate sheet

2. **Prepare your offer** (the heavy lift):
   - Identify the right SIN(s) from the GSA Schedule solicitation
   - Prepare technical proposal demonstrating capability
   - Prepare pricing: commercial sales practices format, price lists with proposed government pricing
   - Compile past performance documentation
   - Complete all required representations and certifications
   - Prepare digital certificates for MAS submissions

3. **Submit via GSA eBuy/eOffer system**

4. **Negotiate with GSA contracting officer**
   - They will review your pricing and may ask for discounts
   - Expect 2-3 rounds of negotiation

5. **Award**

### Timeline

| Step | Duration |
|------|----------|
| Preparation | 1-3 months |
| GSA review and negotiation | 3-12 months |
| **Total** | **6-18 months** |

### Recommendation

**Wait until you have**:
- At least 2 years of operating history
- 2+ past performance references (government or commercial)
- Revenue sufficient to justify the investment
- Bandwidth to manage the application process

In the meantime, pursue contracts through **open-market competition** (SAM.gov opportunities), **micro-purchases** (under $10K simplified procedures), **simplified acquisition** (under $250K), and **subcontracting** to larger primes who already have GSA schedules.

---

## 9. Cost Accounting Standards (CAS)

### What's Required for Small Companies?

| Company Size | CAS Requirement |
|--------------|----------------|
| Small business (under $7.5M in CAS-covered contracts) | **Exempt from CAS** |
| Contracts under $2M individually | **Exempt from CAS** |
| Between thresholds | **Modified CAS coverage** (4 standards) |
| Above full threshold ($50M+) | **Full CAS coverage** (19 standards) |

**As a 1-person company, you are almost certainly exempt from CAS.** However, you still need:

### Adequate Accounting System

Even without CAS, government contracting requires an accounting system that can:

1. **Segregate costs by contract** -- track direct costs to specific contracts
2. **Distinguish between direct and indirect costs** -- know your labor rates, overhead, G&A
3. **Accumulate costs by cost element** -- labor, materials, travel, etc.
4. **Comply with FAR Part 31** -- cost allowability principles (no alcohol, entertainment, lobbying, etc.)
5. **Support incurred cost submissions** -- if you have cost-reimbursement contracts

### Practical Setup for a 1-Person Company

- **QuickBooks Online** or **Xero** with proper chart of accounts is sufficient
- Set up separate **classes or categories** for each contract
- Track **direct labor hours** by contract (use a timesheet tool: Toggl, Harvest, or even a spreadsheet)
- Maintain separate pools for:
  - Direct labor
  - Direct materials/subcontractors
  - Overhead (indirect costs related to production)
  - General & Administrative (G&A -- all business costs not direct or overhead)
  - Fringe benefits

### When CAS Matters

You will need a DCAA-compliant accounting system if you pursue **cost-reimbursement** or **time-and-materials** contracts. For **firm-fixed-price** contracts (most common for small software companies), the accounting requirements are lighter.

**Recommendation**: Start with firm-fixed-price contracts. Set up a clean QuickBooks chart of accounts from day one. Hire a government contracting accountant ($2,000-5,000/year) to set up your indirect rate structure.

---

## 10. Estimated Costs

### Startup Costs (One-Time)

| Item | Estimated Cost | Notes |
|------|---------------|-------|
| LLC formation (state filing) | $50-500 | Varies by state (WY: $100, DE: $90, CA: $70, NY: $200) |
| Registered agent (if needed) | $50-300/year | Required in most states; can be yourself if resident |
| EIN application | Free | IRS online |
| SAM.gov registration | Free | Do NOT pay third-party "registration services" |
| CAGE code | Free | Auto-assigned in SAM |
| Business bank account | Free-$25/month | Many banks offer free business checking |
| **Legal setup (optional but recommended)** | $1,000-3,000 | Attorney review of operating agreement, structure advice |
| **Accounting system setup** | $500-2,000 | QuickBooks + gov-con accountant initial setup |
| **Total One-Time** | **~$1,600-6,000** | |

### Annual Recurring Costs

| Item | Estimated Annual Cost | Notes |
|------|--------------------|-------|
| State LLC annual report/fee | $0-800 | Varies by state (CA: $800 minimum tax, WY: $60, DE: $300) |
| General Liability insurance | $500-1,200 | |
| Professional Liability (E&O) | $800-2,000 | |
| Cyber Liability insurance | $1,000-3,000 | |
| Workers' Comp (if required) | $300-800 | May be exempt as sole owner in some states |
| Accounting software (QuickBooks) | $300-600 | |
| Gov-con accountant | $2,000-5,000 | Annual indirect rate setup + incurred cost prep |
| SAM.gov renewal | Free | Annual, takes 1-2 weeks |
| **Total Annual** | **~$5,000-13,500** | |

### Optional / Future Costs

| Item | Estimated Cost | When |
|------|---------------|------|
| 8(a) certification application | Free (SBA charges nothing) | Year 2+ |
| GSA Schedule consultant | $5,000-15,000 | Year 2+ |
| GSA Schedule Industrial Funding Fee | 0.75% of GSA sales | After GSA award |
| FSO training (for clearances) | $1,500-3,000 | When classified contract requires it |
| CMMC assessment (Level 2) | $25,000-75,000 | When DoD contract requires it (2025+ rollout) |
| GovWin subscription | $1,500-10,000/year | For opportunity intelligence (optional) |

### Total Budget to Get Started

| Scenario | Budget |
|----------|--------|
| **Minimum viable** (entity + SAM + basic insurance) | ~$3,000-5,000 |
| **Recommended** (add accountant + legal review) | ~$6,000-12,000 |
| **Full setup** (add GSA Schedule pursuit) | ~$15,000-25,000 |

---

## 11. Key NAICS Codes

Select all that apply to your products and services during SAM.gov registration. You can list multiple codes. Your **primary NAICS** should be the one that best represents your main revenue-generating activity.

### Recommended NAICS Codes for AI/ML + Kubernetes + DevSecOps + Cybersecurity

| NAICS Code | Title | Use For |
|------------|-------|---------|
| **511210** | Software Publishers | Aegis platform as a software product/SaaS. **Recommended as primary NAICS.** |
| **518210** | Computing Infrastructure Providers, Data Processing, Web Hosting, and Related Services | Cloud infrastructure, Kubernetes platform hosting |
| **541511** | Custom Computer Programming Services | Custom software development, integrations |
| **541512** | Computer Systems Design Services | Systems architecture, DevSecOps implementation |
| **541519** | Other Computer Related Services | AI/ML services, consulting, training |
| **541715** | Research and Development in the Physical, Engineering, and Life Sciences | AI/ML R&D (if applicable) |
| **541330** | Engineering Services | Platform engineering, infrastructure engineering |
| **561621** | Security Systems Services (except Locksmiths) | Cybersecurity monitoring and management |
| **541690** | Other Scientific and Technical Consulting Services | Technical consulting, advisory |

### PSC (Product Service Codes) to Know

You will also need to identify PSC codes when bidding. Common ones for your work:

| PSC Code | Description |
|----------|-------------|
| D302 | IT and Telecom -- Systems Development |
| D306 | IT and Telecom -- Systems/Programming/Analysis |
| D307 | IT and Telecom -- IT Strategy and Architecture |
| D310 | IT and Telecom -- Cyber Security and Data Backup |
| D399 | IT and Telecom -- Other Services |
| 7030 | IT Software (ADP) |

---

## 12. Ongoing Maintenance

### Annual Tasks

| Task | When | Penalty for Missing |
|------|------|-------------------|
| Renew SAM.gov registration | 30 days before expiration (annually) | Cannot receive awards or payments |
| Update representations and certifications in SAM | With renewal | Non-compliance |
| Renew all insurance policies | At policy expiration | Cannot bid on new contracts; may violate existing contracts |
| File state annual report / pay franchise tax | Per state schedule | Loss of good standing; eventual dissolution |
| Update size standard self-certification | If revenue changes materially | False certification is a federal crime |
| Review/update NAICS codes in SAM | Annually | May miss relevant opportunities |
| File small business subcontracting reports (if applicable) | Per contract terms | Contract compliance issues |

### Tax Obligations

- **Quarterly estimated taxes**: Federal + state (if applicable)
- **Self-employment tax**: 15.3% on net self-employment income (or reasonable salary if S-Corp elected)
- **Annual tax return**: Form 1120-S (S-Corp) or Schedule C (sole proprietor/LLC)
- **State franchise/privilege taxes**: Varies by state

---

## 13. Where to Find Opportunities

### Free Resources

| Resource | URL | What You Find |
|----------|-----|---------------|
| **SAM.gov Contract Opportunities** | [sam.gov/content/opportunities](https://sam.gov/content/opportunities) | All federal opportunities over $25K |
| **FPDS** (Federal Procurement Data System) | [fpds.gov](https://www.fpds.gov) | Historical contract data -- research who won what |
| **USASpending** | [usaspending.gov](https://www.usaspending.gov) | Federal spending data -- find agencies spending in your space |
| **SBIR/STTR** | [sbir.gov](https://www.sbir.gov) | Small Business Innovation Research grants -- excellent for AI/ML R&D |
| **GSA eBuy** | [ebuy.gsa.gov](https://ebuy.gsa.gov) | RFQs through GSA Schedules |
| **Agency forecast sites** | Various | Many agencies publish annual procurement forecasts |

### Paid Tools (Consider Later)

| Tool | Cost | Value |
|------|------|-------|
| **GovWin** (Deltek) | $2,500-10,000+/year | Opportunity intelligence, competitor analysis, early alerts |
| **Bloomberg Government** | ~$6,000/year | Policy + procurement intelligence |
| **GovTribe** | $600-1,200/year | Contract data, opportunity tracking |

### Best Entry Strategies for a 1-Person Company

1. **Micro-purchases** (under $10,000): Agencies can buy directly with a government purchase card. No formal competition. Network with contracting officers.

2. **Simplified Acquisitions** ($10K-$250K): Streamlined process, less paperwork. Small business set-asides are common in this range.

3. **Subcontracting to primes**: Partner with large contractors (Booz Allen, Deloitte, SAIC, Leidos, etc.) who need specialized Kubernetes/DevSecOps skills. This builds past performance without winning your own prime contracts.

4. **SBIR/STTR grants**: Phase I grants of $50K-275K for R&D. Excellent for AI/ML platform development. Non-dilutive funding. No past performance required.

5. **Task orders on existing vehicles**: Once you have GSA or a BPA (Blanket Purchase Agreement), agencies can place orders quickly.

---

## Checklist Summary

Print this and check off as you complete each item.

### Phase 1: Business Foundation (Weeks 1-4)
- [ ] Form LLC in home state
- [ ] Obtain EIN from IRS (same day, online)
- [ ] Open business bank account
- [ ] Set up business accounting (QuickBooks or similar)
- [ ] Establish a physical business address

### Phase 2: Federal Registration (Weeks 2-8)
- [ ] Create Login.gov account
- [ ] Register on SAM.gov
- [ ] Verify UEI is assigned
- [ ] Verify CAGE code is assigned
- [ ] Select NAICS codes (see Section 11)
- [ ] Self-certify as small business
- [ ] Complete all representations and certifications
- [ ] Receive confirmation of active registration

### Phase 3: Insurance (Weeks 3-6)
- [ ] Obtain General Liability insurance ($1M/$2M)
- [ ] Obtain Professional Liability / E&O insurance ($1M)
- [ ] Obtain Cyber Liability insurance ($1-2M)
- [ ] Obtain Workers' Compensation (if required in your state)
- [ ] Save all certificates of insurance (COIs) for future proposals

### Phase 4: Certifications (Weeks 4-24)
- [ ] Determine eligibility for set-aside programs:
  - [ ] 8(a) Business Development (if socially/economically disadvantaged)
  - [ ] SDVOSB (if service-disabled veteran)
  - [ ] WOSB/EDWOSB (if woman-owned)
  - [ ] HUBZone (if located in qualifying zone)
- [ ] Apply for all eligible certifications
- [ ] Track application status

### Phase 5: Start Competing (Week 8+)
- [ ] Set up saved searches on SAM.gov for your NAICS codes
- [ ] Research agencies spending in your technology areas (USASpending)
- [ ] Identify 3-5 prime contractors to approach for subcontracting
- [ ] Develop a capability statement (1-2 page PDF)
- [ ] Attend SBA matchmaking events / industry days
- [ ] Respond to first RFI or sources sought notice
- [ ] Submit first proposal

### Phase 6: Growth (Months 6-18)
- [ ] Evaluate GSA MAS application readiness
- [ ] Apply for GSA Schedule (if 2+ years of past performance available)
- [ ] Consider SBIR/STTR applications for AI/ML R&D
- [ ] Build past performance portfolio
- [ ] Pursue facility clearance (only when contract requires it)

---

## Key Contacts and Resources

| Resource | Contact |
|----------|---------|
| SBA District Office Locator | [sba.gov/local-assistance](https://www.sba.gov/local-assistance) |
| SAM.gov Help Desk | 866-606-8220 (Federal Service Desk) |
| SBA 8(a) Program | [sba.gov/8a](https://www.sba.gov/federal-contracting/contracting-assistance-programs/8a-business-development-program) |
| DCSA (Security Clearances) | [dcsa.mil](https://www.dcsa.mil) |
| SBA SCORE Mentors (free) | [score.org](https://www.score.org) |
| Procurement Technical Assistance Centers (PTAC) | [aptac-us.org](https://www.aptac-us.org) -- free government contracting counseling |

---

## Important Warnings

1. **Never pay for SAM.gov registration.** It is free. Companies that charge for this are scams or unnecessary middlemen.

2. **Never misrepresent your size status or certifications.** False Claims Act violations carry penalties of up to $11,000+ per false claim, plus treble damages. Size standard fraud is a federal crime.

3. **Organizational Conflict of Interest (OCI)**: If you consult for an agency on requirements, you may be barred from competing on the resulting contract. Understand OCI rules before engaging with agencies.

4. **CMMC (Cybersecurity Maturity Model Certification)**: As of 2025-2026, DoD is rolling out CMMC 2.0. If you plan to handle CUI (Controlled Unclassified Information) on DoD contracts, you will need at least CMMC Level 2 certification. This requires implementing all 110 NIST SP 800-171 controls and passing a third-party assessment. Plan and budget for this ($25K-75K assessment cost) if DoD is a target market.

5. **Buy American Act / Trade Agreements Act (TAA)**: If selling products (including software), ensure your products are manufactured or substantially transformed in the US or a TAA-designated country. SaaS hosted on US-based infrastructure generally complies.

6. **Section 889**: Federal agencies cannot procure telecommunications or video surveillance equipment from Huawei, ZTE, Hytera, Hikvision, or Dahua (or their subsidiaries). Ensure your technology stack does not include prohibited components.
