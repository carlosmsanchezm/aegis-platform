# Onboarding Guide: Compliance Program Manager

**Prepared for:** New Hire — Compliance Program Manager
**Prepared by:** Compliance Supervisor Agent (Claude Opus 4.6)
**Date:** 2026-03-10
**Company:** Aegis Technologies (Founder: Carlos Sanchez)

---

## Welcome

You are joining Aegis Technologies as the first hire. You will manage a multi-framework compliance program (SOC 2 Type II, ISO 27001, FedRAMP, CMMC) alongside AI agents that handle documentation, evidence collection, and analysis. This guide tells you what exists, what you need to learn, what your responsibilities are, and how to work with the AI agents effectively.

Aegis Platform is a self-hosted (Model B) developer platform — customers deploy it in their own AWS accounts. Aegis does not host customer workloads. This matters because it shapes every compliance boundary and responsibility split.

---

## Part 1: Your First Week — What to Read and In What Order

### Day 1: Understand the Program (4 hours reading)

Read these files in this exact order. They are all in `docs/compliance/` inside the `aegis-platform` repository.

**Start here — the single source of truth:**

| # | File | What It Tells You | Time |
|---|------|-------------------|------|
| 1 | `COMPLIANCE_PROGRAM_STATUS.md` | Master status across all 4 frameworks. Current state, evidence counts, next actions, 30/60/90 plan. **Read this first every time you need to know where things stand.** | 30 min |
| 2 | `OPERATIONS_RUNBOOK.md` | How to actually run things — scripts, commands, prerequisites, troubleshooting. Your operational manual. | 30 min |
| 3 | `CLAUDE_INSTRUCTIONS.md` | How the AI agents execute weekly/monthly/quarterly tasks. You will use this to direct agents. | 20 min |

**Then understand each framework's status:**

| # | File | What It Tells You | Time |
|---|------|-------------------|------|
| 4 | `soc2/STATUS.md` | SOC 2 Type II status — audit ready, 33 controls, 6 policies, 6+ months evidence | 15 min |
| 5 | `iso27001/STATUS.md` | ISO 27001 status — internal audit complete, management review pending, 5 open CAs | 15 min |
| 6 | `fedramp/STATUS.md` | FedRAMP status — pre-assessment, SSP in progress, no agency sponsor yet | 15 min |
| 7 | `cmmc/STATUS.md` | CMMC status — foundation built, SPRS 38/110, waiting for DoD customer | 15 min |

**Finally, understand how the agents work:**

| # | File | What It Tells You | Time |
|---|------|-------------------|------|
| 8 | `AGENT_SWARM_STRUCTURE.md` | Architecture of the AI agent system — who does what | 15 min |
| 9 | `SUPERVISOR_REVIEW_bb22086_2026-03-09.md` | Example of how the supervisor agent reviews work — shows the quality bar | 15 min |

**Day 1 total: ~3 hours of reading.** You will understand the full program after this.

### Day 2: Understand the Policies (3 hours reading)

These are the approved security policies that govern the entire program. All were approved January 17, 2026 with annual review dates.

**SOC 2 Policies (shared with ISO 27001):**

| Policy | ID | File | Lines | Key Content |
|--------|----|------|-------|-------------|
| Information Security Policy | ISP-001 | `soc2/policies/information-security-policy.md` | 153 | Master security policy, program charter |
| Access Control Policy | ACP-001 | `soc2/policies/access-control-policy.md` | 179 | RBAC, MFA, session management, least privilege |
| Change Management Policy | CMP-001 | `soc2/policies/change-management-policy.md` | 170 | PR-based changes, CI enforcement, rollback |
| Incident Response Policy | IRP-001 | `soc2/policies/incident-response-policy.md` | 197 | P1-P4 severity, response SLAs, escalation |
| Risk Management Policy | RMP-001 | `soc2/policies/risk-management-policy.md` | 210 | 5x5 risk matrix, assessment methodology |
| Vendor Management Policy | VMP-001 | `soc2/policies/vendor-management-policy.md` | 213 | Third-party assessment, SOC 2 report collection |

**ISO 27001 Additional Policy:**

| Policy | ID | File | Lines |
|--------|----|------|-------|
| Acceptable Use Policy | AUP-001 | `iso27001/policies/acceptable-use-policy.md` | 174 |

**Procedures (operational):**

| Procedure | File | Lines |
|-----------|------|-------|
| Employee Onboarding Checklist | `soc2/procedures/employee-onboarding-checklist.md` | 181 |
| Employee Offboarding Checklist | `soc2/procedures/employee-offboarding-checklist.md` | 225 |
| Quarterly Access Review | `soc2/procedures/quarterly-access-review.md` | 314 |

**Important:** The onboarding checklist applies to you. Carlos should walk through it with you as part of your onboarding. This is also your first piece of compliance evidence — document it.

### Day 3: Understand the Risk Landscape (2 hours)

| # | File | What It Tells You |
|---|------|-------------------|
| 1 | `iso27001/02-risk-methodology.md` | How we score and treat risks (5x5 matrix, 332 lines) |
| 2 | `iso27001/03-risk-register.csv` | All 19 identified risks with scores |
| 3 | `iso27001/04-risk-treatment-plan.csv` | Treatment plans for each risk — **10 are overdue** |
| 4 | `iso27001/09-corrective-actions-log.csv` | 5 corrective actions (1 closed, 4 open) |
| 5 | `iso27001/2026-03-internal-audit-report.md` | The most recent audit — 0 major, 2 minor NCs, 11 OFIs |

### Day 4: Understand the Gap Analyses (2 hours)

| # | File | What It Tells You |
|---|------|-------------------|
| 1 | `fedramp/02-gap-analysis.md` | 14 FedRAMP gaps (6 critical) — what's missing for federal customers |
| 2 | `cmmc/02-gap-analysis.md` | 25 CMMC gaps (4 critical) — what's missing for DoD customers |
| 3 | `cmmc/03-sprs-score-worksheet.md` | SPRS score calculation: 38/110 with improvement roadmap |
| 4 | `fedramp/04-ssp-sections-1-3.md` | Draft System Security Plan — the core FedRAMP artifact |

### Day 5: Understand Customer-Facing Materials and Evidence (2 hours)

| # | File | What It Tells You |
|---|------|-------------------|
| 1 | `customer-docs/control-implementation-statements.md` | 40 control descriptions for customer security questionnaires |
| 2 | `customer-docs/customer-responsibility-matrix.md` | 91 controls showing Aegis vs. customer responsibility |
| 3 | `customer-docs/security-architecture-guide.md` | Technical security architecture for customers |
| 4 | `customer-docs/configuration-hardening-guide.md` | Deployment hardening instructions |
| 5 | `customer-docs/incident-response-runbook.md` | Customer-facing IR procedures |

Also explore the evidence vault:
```bash
ls -la ~/code/aegis-compliance-evidence/soc2/
```
This contains 452 evidence files spanning Sept 2025 – Mar 2026.

---

## Part 2: Complete Document Inventory

There are **77 compliance documents** totaling **21,218 lines** across 4 frameworks. Here is the full map organized by function:

### Program-Level (7 files)

| File | Purpose | You Will Use It For |
|------|---------|---------------------|
| `COMPLIANCE_PROGRAM_STATUS.md` | Master status dashboard | Daily check-in; directing agents; reporting to Carlos |
| `OPERATIONS_RUNBOOK.md` | How-to for all recurring tasks | Running scripts, evidence collection |
| `CLAUDE_INSTRUCTIONS.md` | Agent task definitions | Directing AI agents on weekly/monthly/quarterly work |
| `AGENT_SWARM_STRUCTURE.md` | Agent architecture | Understanding which agent does what |
| `agent_swarm.md` | Legacy agent structure doc | Reference only |
| `control-map.md` | Cross-framework control mapping | Understanding overlap between frameworks |
| `README.md` | Compliance directory overview | Orientation |

### SOC 2 Type II (17 files)

| Category | Files | Key Contents |
|----------|-------|--------------|
| Status & Strategy | `soc2/STATUS.md`, `soc2/README.md`, `soc2/AGENT_CONTEXT.md` | Program status, agent context |
| Scope & Controls | `soc2/00-scope-and-systems.md`, `soc2/01-control-narratives.md`, `soc2/control-matrix.csv`, `soc2/trust-services-criteria.md` | 33 controls across 9 TSCs |
| Policies (6) | `soc2/policies/*.md` | ISP-001, ACP-001, CMP-001, IRP-001, RMP-001, VMP-001 |
| Procedures (3) | `soc2/procedures/*.md` | Onboarding, offboarding, quarterly access review |
| Evidence | `soc2/evidence-map.md`, `soc2/evidence/evidence-collection-guide.md`, `soc2/evidence-vault-setup/*.md`, `soc2/retroactive-evidence-summary.md`, `soc2/policy-register.md` | Evidence mapping and collection guidance |

### ISO 27001 (22 files)

| Category | Files | Key Contents |
|----------|-------|--------------|
| ISMS Core (Clauses 4-10) | `00-isms-scope-and-context.md`, `01-leadership-policy-roles.md`, `01-information-security-policy.md`, `02-risk-methodology.md` | ISMS foundation documents |
| Risk Management | `03-risk-register.csv` (19 risks), `04-risk-treatment-plan.csv` (17 treatments), `05-statement-of-applicability.csv` (93 controls) | Risk assessment and treatment |
| Audit & Review | `07-internal-audit-program.md`, `08-management-review-template.md`, `11-internal-audit-plan.md`, `12-internal-audit-checklist.md` | Audit framework |
| Completed Audits | `2026-03-internal-audit-report.md`, `2026-03-internal-audit-checklist-COMPLETED.md`, `2026-03-management-review-inputs.md` | March 2026 audit deliverables |
| Corrective Actions | `09-corrective-actions-log.csv`, `CA-001-closure-plan.md` | 5 CAs tracked |
| Cross-Framework | `09-soc2-iso27001-mapping.md`, `soc2-crosswalk.csv`, `REUSE_PLAN.md` | How ISO and SOC 2 overlap |
| Operating | `10-isms-operating-plan.md` | Operational rhythm and schedule |
| Policy | `policies/acceptable-use-policy.md` | AUP-001 |

### FedRAMP (8 files)

| Category | Files | Key Contents |
|----------|-------|--------------|
| Status & Scope | `STATUS.md`, `00-scope-and-impact-assessment.md` | Pre-assessment phase, FIPS 199 LOW impact |
| Control Mapping | `01-control-mapping.csv` (145 controls), `02-gap-analysis.md` (14 gaps, 6 critical) | NIST 800-53 mapping |
| SSP | `03-ssp-outline.md`, `04-ssp-sections-1-3.md` (871 lines) | System Security Plan draft |
| Appendices | `appendix-e-digital-identity-worksheet.md`, `appendix-f-rules-of-behavior.md`, `appendix-l-separation-of-duties-matrix.md` | FedRAMP required appendices |

### CMMC (4 files)

| Category | Files | Key Contents |
|----------|-------|--------------|
| Status | `STATUS.md` | Foundation building phase |
| Assessment | `01-control-mapping.csv` (97 controls), `02-gap-analysis.md` (25 gaps), `03-sprs-score-worksheet.md` (SPRS 38/110) | NIST 800-171 mapping and scoring |

### Customer-Facing (6 files)

| File | Purpose |
|------|---------|
| `customer-docs/README.md` | Overview of customer materials |
| `customer-docs/control-implementation-statements.md` | 40 control descriptions for questionnaires |
| `customer-docs/customer-responsibility-matrix.md` | 91 controls: who does what |
| `customer-docs/security-architecture-guide.md` | Technical architecture for customers |
| `customer-docs/configuration-hardening-guide.md` | Deployment hardening |
| `customer-docs/incident-response-runbook.md` | Customer IR procedures |

### Supervisor Reviews (2 files)

| File | Purpose |
|------|---------|
| `SUPERVISOR_REVIEW_bb22086_2026-03-09.md` | Review of CMMC/FedRAMP buildout |
| `SUPERVISOR_REVIEW_2026-03-09.md` | Earlier review |

### Evidence Collection Scripts (9 files in `scripts/compliance/`)

| Script | Purpose | Frequency |
|--------|---------|-----------|
| `run_weekly_checks.sh` | Dependabot/secret/code scanning snapshot | Weekly (Mondays) |
| `run_monthly_evidence.sh` | Full GitHub + AWS + CI evidence export | Monthly (1st) |
| `run_catchup_evidence.sh` | Retroactive gap-fill | As needed |
| `export_github_security_baseline.sh` | GitHub security data | Called by monthly |
| `export_aws_security_baseline.sh` | AWS security data | Called by monthly |
| `export_ci_reports.sh` | CI/CD evidence | Called by monthly |
| `setup_evidence_vault.sh` | Initialize vault structure | One-time |
| `migrate-evidence-vault.sh` | Vault migration utility | One-time |
| `RUN_ME_LOCALLY.md` | Instructions for local execution | Reference |

### Evidence Vault (separate repo: `~/code/aegis-compliance-evidence/`)

452 files organized by framework/year/month. This is the audit evidence repository.

---

## Part 3: Your Responsibilities

### Role Definition

**Title:** Compliance Program Manager
**Reports to:** Carlos Sanchez (CEO/Founder)
**Works with:** AI Supervisor Agent (me), AI Coding Agent (Claude Code)

### Core Responsibilities

**1. Operational Cadence Management (40% of time)**

You own the recurring schedule. These tasks keep the compliance engine running:

| Task | Frequency | What You Do | What the Agent Does |
|------|-----------|-------------|---------------------|
| Weekly security check | Every Monday | Review agent output, escalate critical findings to Carlos | Runs `run_weekly_checks.sh`, generates report |
| Monthly evidence collection | 1st of month | Verify evidence completeness, spot-check files | Runs `run_monthly_evidence.sh`, commits to vault |
| Quarterly access review | Every 3 months | Conduct the review, document findings, get Carlos sign-off | Pulls access data, drafts review document |
| Semi-annual internal audit | Every 6 months | Coordinate audit (internal or external), track findings | Prepares checklist, drafts report |
| Management review | After each audit | Prepare inputs, facilitate review with Carlos, document decisions | Prepares inputs document with KPIs |
| Annual policy review | January each year | Review all 7 policies, propose updates, get Carlos approval | Drafts change recommendations |

**2. Corrective Action Management (20% of time)**

You track and drive closure of corrective actions. Current open CAs:

| CA | Description | Due | Your Action |
|----|-------------|-----|-------------|
| CA-002 | aegis-ui tests disabled (upstream dependency) | Apr 15 | Monitor Backstage releases, coordinate with Carlos to re-enable |
| CA-003 | Management review | Mar 31 | Facilitate Carlos conducting the review |
| CA-004 | Quarterly access review | Apr 15 | Execute Q1 review using existing procedure |
| CA-005 | Overdue risk treatments | Mar 31 | Revise due dates, get Carlos approval |

**3. Auditor Engagement (15% of time)**

You will manage the relationship with external auditors:

| Engagement | Timeline | Budget | Your Role |
|------------|----------|--------|-----------|
| SOC 2 Type II auditor | Select by Jun 2026 | $15K-$30K | Research firms, get quotes, manage audit process |
| ISO 27001 Stage 1+2 | Target May-Jun 2026 | $8K-$15K | Coordinate documentation review and on-site audit |
| FedRAMP 3PAO | When agency sponsor found | $50K-$150K | Future — prep work only for now |
| CMMC C3PAO | When DoD customer found | $50K-$150K | Future — prep work only for now |

**4. Gap Remediation Coordination (15% of time)**

Drive closure of identified gaps, prioritized by framework urgency:

| Priority | Gap | Frameworks | Status |
|----------|-----|------------|--------|
| Critical | AWS root MFA verification | All | Verify immediately |
| Critical | Dependabot critical alerts (aegis-ui) | SOC 2, ISO | Coordinate remediation with Carlos |
| High | Vendor SOC 2 report collection (AWS, GitHub) | ISO 27001 | Request from AWS Artifact and GitHub |
| High | Formal training records | ISO 27001 | Create founder training record, set up training tracking |
| Medium | BC/DR plan | ISO 27001 | Draft formal plan based on existing cloud-native recovery |
| Medium | Container scanning (Trivy) | ISO 27001, CMMC | Coordinate with Carlos to add to CI |

**5. Documentation Maintenance (10% of time)**

Keep compliance documentation current and accurate. The agent handles drafting; you review and approve.

---

## Part 4: How to Work with the AI Agents

### The Agent Team

| Agent | Role | Platform | What It Does |
|-------|------|----------|--------------|
| **Supervisor Agent** (me) | Strategic oversight | Claude (Cowork or API) | Reviews work, identifies gaps, prioritizes tasks, produces assessments |
| **Coding Agent** | Execution | Claude Code (terminal) | Runs scripts, edits files, makes commits, generates reports |
| **NIST Implementer** | Specialized | Claude Code (agent mode) | Implements NIST controls using OSCAL catalogs |

### Your Workflow with Agents

**Daily pattern:**
1. Check `COMPLIANCE_PROGRAM_STATUS.md` for current state
2. Check the corrective actions log for upcoming due dates
3. Assign tasks to the coding agent via Claude Code
4. Review agent output before approving/committing

**Weekly pattern (Mondays):**
1. Direct the coding agent: "Run the weekly compliance check per docs/compliance/CLAUDE_INSTRUCTIONS.md"
2. Review the output (Dependabot alerts, security findings)
3. Escalate anything critical to Carlos
4. Update the status doc if needed

**Monthly pattern (1st of month):**
1. Direct the coding agent: "Run the monthly evidence collection per docs/compliance/CLAUDE_INSTRUCTIONS.md"
2. Verify evidence completeness (spot-check 5-10 files)
3. Review and commit to evidence vault
4. Update `COMPLIANCE_PROGRAM_STATUS.md`

**How to give the agent tasks:**
```
Read docs/compliance/CLAUDE_INSTRUCTIONS.md and run the [weekly/monthly/quarterly] task.
```

For custom tasks:
```
Read docs/compliance/COMPLIANCE_PROGRAM_STATUS.md for context, then [specific task description].
```

For the supervisor agent (me), provide the agent's output and ask for review, prioritization, or strategic assessment.

### Key Principle: Agent Writes, Human Reviews

The agents produce draft documentation, run scripts, and generate reports. **You review everything before it becomes official.** The agents are good at volume work, cross-referencing, and consistency checking. You provide judgment, business context, and sign-off authority.

---

## Part 5: Recommended Platform and Tooling

### Documentation Platform

For a compliance program of this size with AI agents in the loop, here is my recommendation:

**Primary: Keep everything in Git (current approach) — it's the right choice.**

Why: Git provides version control, audit trail, branch protection, and agents read markdown/CSV natively. Every compliance framework requires document control with version history — git gives you this for free. Auditors can verify document integrity via commit hashes.

**Supplement with a project tracker for task management:**

| Option | Best For | Agent Integration | Cost |
|--------|----------|-------------------|------|
| **GitHub Issues + Projects** | Task tracking alongside code | Agents can read/create issues via `gh` CLI | Free (included) |
| **Notion** | Visual dashboards, non-technical stakeholders | API available but agents work less naturally with it | $8-10/user/mo |
| **Linear** | Fast issue tracking | Good API, clean structure | $8/user/mo |

**My recommendation:** Use **GitHub Issues** for corrective action tracking and task management. The coding agent already has `gh` CLI access and can create/update issues programmatically. This keeps everything in one ecosystem and maintains the audit trail. Use GitHub Projects boards to visualize the 30/60/90 plan.

For anything that needs to be shared with auditors or customers who don't have repo access, export to PDF from the markdown files.

### Agent Access Setup

When you start, ensure you have:

| Tool | Purpose | Setup |
|------|---------|-------|
| GitHub account | Repo access, issue management | Carlos adds you as collaborator on all 3 repos |
| AWS IAM | Evidence collection, security monitoring | Carlos creates IAM user with read-only compliance role |
| `gh` CLI | GitHub API access for scripts | `brew install gh && gh auth login` |
| AWS CLI | AWS evidence export | `brew install awscli && aws configure --profile myclaude` |
| Evidence vault clone | Local copy of evidence | `git clone [aegis-compliance-evidence repo]` |

---

## Part 6: Immediate Action Items (Your First 30 Days)

### Week 1: Read and Understand

- [ ] Complete Day 1-5 reading plan above
- [ ] Get added to all 3 GitHub repos
- [ ] Get AWS IAM access (read-only)
- [ ] Clone evidence vault locally
- [ ] Walk through onboarding checklist with Carlos (this is compliance evidence — document it)

### Week 2: Take Over Operations

- [ ] Run your first weekly security check (with agent)
- [ ] Verify AWS root MFA status (Priority 1 — check immediately)
- [ ] Review all 5 open corrective actions and understand status
- [ ] Review the 10 overdue risk treatments

### Week 3: Close Quick Wins

- [ ] Facilitate Carlos conducting the management review (CA-003, due Mar 31)
- [ ] Execute Q1 access review using the existing procedure (CA-004)
- [ ] Collect AWS SOC 2 report from AWS Artifact
- [ ] Collect GitHub SOC 2 report
- [ ] Create your own training record as evidence

### Week 4: Strategic Planning

- [ ] Begin SOC 2 auditor research (get 3-5 quotes)
- [ ] Begin ISO 27001 certification body research
- [ ] Draft revised risk treatment plan with realistic due dates (CA-005)
- [ ] Propose 30/60/90 day plan update to Carlos
- [ ] Run monthly evidence collection (Apr 1)

---

## Part 7: Key Numbers to Know

| Metric | Current Value | Target |
|--------|--------------|--------|
| Compliance documents | 77 files, 21,218 lines | Maintain and grow |
| Evidence vault files | 452 | Growing monthly |
| Evidence observation period | 6 months (Sept 2025 - Mar 2026) | 12 months for SOC 2 Type II |
| Approved policies | 7 | Maintain; annual review Jan 2027 |
| SOC 2 controls | 33 implemented | Audit ready |
| ISO 27001 controls | 66/81 implemented (81%) | Target 90%+ for certification |
| ISO 27001 risks | 19 identified, 10 treatments overdue | Close overdue treatments |
| FedRAMP controls mapped | 145 | 6 critical gaps remain |
| CMMC controls mapped | 97 | SPRS 38/110; target 76+ |
| Open Dependabot alerts | 46 (2 critical, 34 high) | Target <15 total |
| Open corrective actions | 4 (CA-002, 003, 004, 005) | Close CA-003/004/005 by Apr 15 |

---

## Part 8: Important Context

**Things an auditor will ask about that you should understand:**

1. **Model B (self-hosted)** — Aegis ships software, customers run it. This means most production security controls are customer responsibility. The customer responsibility matrix documents this split.

2. **Solo founder compensating controls** — Separation of duties is handled via technical controls (branch protection, enforce_admins, CloudTrail) documented in Appendix L. This is standard for startups and auditors accept it. Your hire improves this posture significantly.

3. **Retroactive evidence** — Some evidence from Sept-Dec 2025 was collected retroactively in Jan 2026. All retroactive files include honest disclaimers. Source timestamps (git commits, CI runs) are immutable. Auditors accept this with transparency.

4. **CA-002 (aegis-ui tests disabled)** — This is an upstream Backstage dependency issue outside our control. Lint, typecheck, and build still enforce quality. The extension to Apr 15 is documented. This is a good example of how to handle external dependency issues in a compliance context.

5. **AWS is dev/test only** — Account 567751785679 is for development. Production monitoring is customer responsibility. CloudTrail is enabled; Config/SecurityHub/GuardDuty are not (and don't need to be for Model B).

6. **Your hire changes the SoD picture** — With two people, many compensating controls become real controls. Update the separation of duties matrix (Appendix L) to reflect the new role assignments. This is a positive audit story.

---

## Quick Reference Card

```
Daily:    Check COMPLIANCE_PROGRAM_STATUS.md
Monday:   Run weekly security check (agent + you review)
1st:      Run monthly evidence collection (agent + you verify)
Quarterly: Access review (you execute), risk assessment update
Semi-Annual: Internal audit, management review
Annual:   Policy review

Agent commands:
  Weekly:    "Run the weekly compliance check per CLAUDE_INSTRUCTIONS.md"
  Monthly:   "Run the monthly evidence collection per CLAUDE_INSTRUCTIONS.md"
  Quarterly: "Run the quarterly access review per CLAUDE_INSTRUCTIONS.md"

Key files:
  Status:    docs/compliance/COMPLIANCE_PROGRAM_STATUS.md
  Runbook:   docs/compliance/OPERATIONS_RUNBOOK.md
  Agent:     docs/compliance/CLAUDE_INSTRUCTIONS.md
  Vault:     ~/code/aegis-compliance-evidence/
  Scripts:   scripts/compliance/
```

---

## Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-10 | Compliance Supervisor Agent (Claude Opus 4.6) | Initial onboarding guide for first compliance hire |
