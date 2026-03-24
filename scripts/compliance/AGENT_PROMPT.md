# Compliance Automation Agent Prompt

**Usage:** Give this prompt to any Claude Code agent to run the compliance cadence.

```
Read docs/compliance/ONBOARDING_COMPLIANCE_MANAGER.md for full program context,
then docs/compliance/COMPLIANCE_PROGRAM_STATUS.md for current state and next actions,
then docs/compliance/CLAUDE_INSTRUCTIONS.md for task procedures.

Determine what tasks are due based on today's date. Execute them following the
steps in CLAUDE_INSTRUCTIONS.md — run scripts, collect evidence, REMEDIATE any
findings (fix vulnerabilities, update configs, upgrade dependencies), then
document the remediation as evidence across ALL framework status docs.
Update COMPLIANCE_PROGRAM_STATUS.md with new dates and metrics, check off
completed items, and commit. Report what was done and flag anything needing
human attention.
```

## Reading Order

| # | File | Purpose |
|---|------|---------|
| 1 | `docs/compliance/ONBOARDING_COMPLIANCE_MANAGER.md` | Full program context — frameworks, documents, responsibilities |
| 2 | `docs/compliance/COMPLIANCE_PROGRAM_STATUS.md` | Current state — what's done, what's due, 30/60/90 plan |
| 3 | `docs/compliance/CLAUDE_INSTRUCTIONS.md` | Task procedures — exact steps for weekly/monthly/quarterly tasks |

## Task Cadence

| Frequency | Task | Script | Trigger |
|-----------|------|--------|---------|
| Weekly (Mondays) | Security check + remediation | `scripts/compliance/run_weekly_checks.sh` | Every Monday |
| Monthly (1st) | Evidence collection | `scripts/compliance/run_monthly_evidence.sh` | 1st of each month |
| Quarterly | Access review | Manual (procedure in CLAUDE_INSTRUCTIONS.md) | Jan 15, Apr 15, Jul 15, Oct 15 |
| Semi-annual | Internal audit | Manual (see iso27001/11-internal-audit-plan.md) | Feb, Aug |
| Annual | Policy review | Manual (all 7 policies) | January |

## Critical Rule: Detect → Remediate → Document

The compliance cadence is NOT just recording findings. Every run must:

1. **Detect** — Run scripts, collect evidence, identify findings
2. **Remediate** — Fix what was found (upgrade deps, patch vulns, close gaps)
3. **Document** — Record what was fixed as evidence across ALL frameworks

If a finding cannot be fixed by the agent (requires human judgment, breaking API change, vendor decision), flag it for human attention with a clear explanation of why.

## After Every Run

1. **Remediate findings** — Fix vulnerabilities, upgrade dependencies, resolve alerts across ALL repos (aegis-platform, aegis-ui, sovran)
2. **Write remediation evidence** — Create a report in `$EVIDENCE_VAULT` documenting what was found, what was fixed, and which controls it demonstrates
3. **Update ALL framework status docs** — Not just `COMPLIANCE_PROGRAM_STATUS.md`, but also:
   - `docs/compliance/soc2/STATUS.md` (CC7.1 vulnerability management)
   - `docs/compliance/iso27001/STATUS.md` (A.8.8 technical vulnerability management)
   - `docs/compliance/cmmc/STATUS.md` (SI family flaw remediation)
   - `docs/compliance/fedramp/STATUS.md` (SI-2, RA-5 flaw remediation)
4. **Update risk treatment plan** — `docs/compliance/iso27001/04-risk-treatment-plan.csv` if remediation advances any risk treatments (especially RISK-001 supply chain)
5. **Check corrective actions** — `docs/compliance/iso27001/09-corrective-actions-log.csv` for any CAs approaching due dates; flag ones needing human attention
6. **Commit** with message: `compliance: [task type] YYYY-MM-DD`
7. **Report** findings, remediations performed, and anything needing human attention
