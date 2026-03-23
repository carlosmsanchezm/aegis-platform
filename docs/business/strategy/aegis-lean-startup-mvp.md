# Aegis MVP & Go-to-Market Plan (Lean Startup)

## Context

The code is ahead of the business. Aegis has working hub-and-spoke architecture, Keycloak auth, workspace launching, a VS Code extension, Backstage UI, budget enforcement, and compliance documentation. But zero customers, zero pilots, zero validated learning about whether a program office will actually buy this.

**The Lean Startup principle**: Stop building features. Start learning whether the business works. Every action in the next 90 days should reduce procurement and deployment risk, not add features.

**Riskiest assumption**: Not "can we build it" — you already have. It's "will a defense program office pay to replace their current painful workflow with Aegis, given the procurement friction involved?"

---

## What the MVP Is (and Isn't)

The MVP is NOT a stripped-down version of the platform. The code is already past MVP. The MVP is the **minimum package needed to get a funded pilot**:

| You need this | You do NOT need this |
|---|---|
| Working end-to-end demo (submit workspace → connect via Sovran) | Persistent storage (just implemented — nice to have, not blocking) |
| 5-minute demo video for async sharing | Pre-built ML images |
| NIST 800-53 control mapping document (for ISSM review) | FIPS crypto (important, but not blocking a pilot conversation) |
| One-page capability statement in government format | mTLS client certs |
| Pricing model that fits micro-purchase / SBIR / OTA | CUI data containment features (inactivity timeout, audit logging) |
| 2-3 discovery conversations per week | More documentation |

---

## Phase 1: Demo-Ready (Weeks 1-2)

**Goal**: A working demo you can show in 15 minutes that makes someone say "I want this for my team."

### What needs to work end-to-end:
1. Backstage UI accessible at `ui.aegis-platform.tech`
2. User logs into Keycloak
3. User clicks "Launch Workspace" → picks GPU size → workspace goes RUNNING
4. User opens VS Code → Sovran connects → terminal works → `nvidia-smi` (or just Python/bash on CPU)
5. User disconnects → session ends cleanly

### Actions:
- [ ] Verify workspace launch works on EKS (the stuck `ironbank-test-hub` workload needs cleanup)
- [ ] Verify Sovran connects to a workspace on EKS end-to-end
- [ ] Record a 5-minute screen recording of the full flow
- [ ] Have the demo script ready (`~/.claude/plans/aegis-demo-script.md` — already written)

### What NOT to do:
- Do not fix persistent storage, FIPS, mTLS, or any other feature
- Do not write more docs
- Do not build more UI features

---

## Phase 2: Pre-Sales Assets (Weeks 3-4)

**Goal**: Materials that let a government technical lead evaluate whether Aegis is worth their time.

### Assets to create:
1. **NIST 800-53 Control Mapping** (~40 hours) — For each relevant control, one paragraph on how Aegis addresses it. Not a full SSP — just enough for an ISSM to say "this is ATO-able." You already have `docs/compliance/customer-docs/control-implementation-statements.md` — refine it into a standalone deliverable.

2. **One-page Capability Statement** — Government-format capability statement (Past Performance, Core Competencies, Differentiators, NAICS codes, cage code, DUNS). This is what a contracting officer needs to evaluate you.

3. **Architecture One-Pager** — Single-page diagram showing hub-and-spoke, where Aegis sits in the customer's network, what security controls are built in. For technical leads and ISSMs.

4. **Pricing Sheet** — Simple: $X/year for Y engineers on Z clusters. Include SBIR/OTA/direct purchase options.

---

## Phase 3: Customer Discovery (Weeks 3-8, parallel with Phase 2)

**Goal**: 10-15 discovery conversations. Learn whether the pain is real for others and whether they can buy.

### Who to talk to:

| Priority | Target | How to reach them | What to ask |
|---|---|---|---|
| 1 | Former NTConcepts colleagues | Direct outreach | "What's your GPU dev environment like now? Still using Workspaces + Kubeflow?" |
| 2 | Platform One / Big Bang users | P1 Mattermost, your existing P1 access | "What do your ML teams use for GPU access? How long did it take to set up?" |
| 3 | DoD software factory engineers | Kessel Run, Army SWF, BESPIN LinkedIn contacts | "Walk me through getting a GPU for training. How many hops?" |
| 4 | Defense contractor ML leads | LinkedIn, AFCEA/AUSA conferences | "If I could give your data scientists VS Code on a GPU in 30 seconds inside your ATO boundary, would that save you time?" |

### The questions (never ask "would you use this?"):
- "Walk me through what happens from the moment you need a GPU to the moment you have a working environment."
- "How long does that take? Who's involved?"
- "What breaks? What's the worst part?"
- "If I gave you a Helm chart for a self-hosted GPU platform, what would stop you from deploying it this month?"
- "Is developer productivity on GPUs a line item in your program, or would this need new funding?"

### What to track:
After every conversation, record:
- Did they describe the pain independently (not prompted)?
- Did they commit to a next step (demo, intro to someone, review of architecture)?
- What blockers did they mention (ATO, budget, authority, existing contracts)?

---

## Phase 4: Design Partner Acquisition (Weeks 7-12)

**Goal**: 1-2 signed design partners who will let you deploy Aegis in their environment.

### The offer:
> "We'll deploy Aegis on your cluster for free. You give us 2-4 weeks of access and a technical lead who can give us 2 hours/week of feedback. We do all the deployment work. At the end, you have a working GPU workspace platform. If it works, we talk about a paid engagement."

### What this gets you:
- Real deployment in a real environment (not your own EKS account)
- Deployment playbook for the next customer
- ISSM/CISO feedback on ATO feasibility
- A reference customer for sales conversations
- Validated learning: does the product actually work when YOU are not operating it?

### Contracting vehicles for design partners:
- **CRADA** (Cooperative Research and Development Agreement) — no money changes hands, just collaboration
- **Engineering services agreement** — you bill hours for deployment support (~$150-200/hr)
- **SBIR Phase I** — $50K-$250K for a 6-month feasibility study with a government partner
- **OTA prototype** — through NSTXL, SOSSEC, or National Armaments Consortium

---

## Phase 5: Funded Pilot (Weeks 10-16)

**Goal**: Convert one design partner to a paid pilot, or secure SBIR/OTA funding.

### Apply to these vehicles:
- [ ] Air Force SBIR Open Topic (AF SBIR/STTR)
- [ ] SOCOM OTA through NSTXL
- [ ] DIU (Defense Innovation Unit) — if they have an active AI/ML infrastructure topic
- [ ] Direct-to-Phase-II SBIR — you have prior work (working code) to justify skipping Phase I

---

## 90-Day Checkpoint (Week 12)

### Evaluate against these metrics:

| Metric | Target | Pivot signal |
|---|---|---|
| Discovery conversations completed | 10+ | If you can't get meetings, the channel is wrong |
| Design partners with scheduled deployment | 1+ | If nobody commits time, the pain isn't sharp enough |
| Warm leads with defined next steps | 2+ | If meetings don't convert to next steps, positioning is off |
| Funded opportunities in pipeline (SBIR/OTA) | 1+ | If no vehicle fits, reconsider pricing/packaging |

### Decision framework:
- **0 of 4 targets hit**: Pivot. Consider: selling the Sovran extension alone as the wedge, or pivoting to a different customer segment, or joining an existing defense tech company (like Palantir) with this domain expertise.
- **1 of 4**: Investigate why the others failed. One more 90-day cycle.
- **2+ of 4**: Persevere. Start hiring.

---

## What to STOP Doing Right Now

| Stop this | Why |
|---|---|
| Building new features (storage, mTLS, FIPS, audit logging) | Code is ahead of the business. No customer has asked for these. |
| Writing more docs | You have 30+ docs. A customer needs 3: control mapping, capability statement, architecture one-pager. |
| Optimizing the architecture | The architecture is sound. Optimize for deployment speed, not technical elegance. |
| Running the EKS cluster without a purpose | It costs money. Use it for demos only. Tear it down between demos. |

## What to START Doing Right Now

| Start this | Why |
|---|---|
| Talking to potential customers (5/week) | Validated learning only comes from customer behavior |
| Recording a demo video | Async sharing is 10x more scalable than live demos |
| Building the control mapping doc | Single highest-ROI pre-sales asset for ISSM conversations |
| Applying to SBIR/OTA | Funded discovery while you search for design partners |

---

## The Wedge Strategy

The research identified an important Lean Startup concept: **lead with a wedge, expand to the platform.**

The full Aegis platform (control plane, auth, workspace orchestration, multi-cluster, budgets) is hard to sell because the value is diffuse. But **one piece** solves a sharp, specific pain:

> **"VS Code Remote on GPU hardware, inside your ATO boundary, in 30 seconds."**

That's the wedge. Every ML engineer in a SCIF or CUI environment who's suffering through Workspaces + Jupyter wants this. Lead with the demo of Sovran connecting to a GPU workspace. Once deployed, expand to the full platform (multi-cluster, budgets, multi-tenancy).

This is how Slack, Dropbox, and every successful platform company started — with a wedge that solved one undeniable pain point.
