# Supervisor Review: Commit bb22086 — CMMC + FedRAMP Buildout

**Review Date:** 2026-03-09
**Reviewer:** Compliance Program Supervisor (Claude Opus 4.6)
**Commit:** `bb22086` on `main`
**Scope:** 13 files changed, 3,964 insertions
**Prior Report Reviewed:** `FEDRAMP_CMMC_BUILDOUT_REPORT_2026-03-09.md`

---

## Overall Assessment: APPROVED WITH FINDINGS

The buildout represents substantial, high-quality work that establishes a credible CMMC foundation and advances FedRAMP preparation. The agent correctly leveraged existing SOC 2 and ISO 27001 assets, produced cross-framework mappings, and maintained honest assessments of gaps. The work is directionally correct and suitable for internal use and pre-assessment discussions.

**However, I identified 5 findings (2 moderate, 3 minor) that should be corrected before any external review.**

---

## Findings

### FINDING-001: SPRS Score Inconsistency Across Documents (MODERATE)

Three documents report different SPRS scores for the same assessment:

| Document | Reported SPRS Score |
|----------|---------------------|
| `COMPLIANCE_PROGRAM_STATUS.md` (line 40) | "Estimated ~42/110" |
| `cmmc/STATUS.md` (lines 19, 105-116) | "Not yet calculated" |
| `cmmc/03-sprs-score-worksheet.md` (line 19) | **38/110** |
| `FEDRAMP_CMMC_BUILDOUT_REPORT` (line 82) | **38/110** |

**Impact:** An auditor or C3PAO seeing contradictory scores across documents would flag this as a documentation control weakness. The master status document says 42, the worksheet says 38, and the CMMC STATUS says "not yet calculated" — all in the same commit.

**Root Cause:** The STATUS.md was likely written before the SPRS worksheet was completed, and the master status doc used a rough estimate. Neither was updated after the worksheet finalized the score.

**Remediation:**
1. Update `COMPLIANCE_PROGRAM_STATUS.md` line 40: change `~42/110` to `38/110 (preliminary)`
2. Update `cmmc/STATUS.md` lines 19 and 105-116: replace "Not yet calculated" with `38/110 (preliminary — see 03-sprs-score-worksheet.md)`
3. All documents should reference the worksheet as the authoritative source

---

### FINDING-002: Buildout Report Control Count Discrepancy (MODERATE)

The buildout report's Section 3 "Control Status Breakdown" (line 59-66) reports:

| Status | Report Says | Gap Analysis Says | SPRS Worksheet Says |
|--------|-------------|-------------------|---------------------|
| Implemented | **52 (54%)** | **46 (47%)** | **46** |
| Not Yet Mapped | **3 (3%)** | — | — |

The "52 implemented" figure appears to incorrectly combine the 46 truly implemented controls with some inherited/N/A controls or round up from partial implementations. The gap analysis and SPRS worksheet both agree on 46 implemented + 11 inherited = 57 fully-scoring controls.

**Impact:** Internal inconsistency undermines credibility if the report is shared externally. A C3PAO would use the SPRS worksheet as authoritative, not the report summary.

**Remediation:** Update the buildout report Section 3 table to match the gap analysis: 46 implemented, 25 partial, 11 inherited/N/A, 6 not implemented, 9 not assessed.

---

### FINDING-003: SSP Baseline Declaration vs. Program Strategy (MINOR)

The SSP document header (line 8) declares:

> **Baseline:** FedRAMP Moderate (with LOW override eligibility for Model B self-hosted deployments)

But `COMPLIANCE_PROGRAM_STATUS.md` and the FedRAMP STATUS target "FedRAMP 20x Low or LI-SaaS." The SSP Section 1.2 does explain the LOW override rationale well, but the top-level baseline declaration could confuse a 3PAO reviewer who sees "Moderate" prominently and "Low" buried in the details.

**Remediation:** Consider rewording the SSP header to: "**Target Baseline:** FedRAMP Low / LI-SaaS (Model B self-hosted); FedRAMP Moderate controls documented for maximum reuse." This better aligns with the actual program strategy.

---

### FINDING-004: CSV Parsing Vulnerability in Control Mapping (MINOR)

The control mapping CSV (`01-control-mapping.csv`) contains fields with embedded commas inside quoted values (e.g., `"AC-6(9),AC-6(10)"`). While valid CSV per RFC 4180, naive parsing tools (Excel auto-import, simple awk scripts) may misparse approximately 5-7 rows, producing incorrect column alignment.

**Impact:** Low. Any proper CSV parser handles this correctly. But if someone opens it in a basic text editor or uses shell scripts for quick analysis, they'll get wrong counts.

**Remediation:** Either escape internal commas differently (semicolons are already used for SOC2/ISO mappings — good) or add a note in the file header about quoted fields. Consider using semicolons consistently for multi-value FedRAMP mappings too.

---

### FINDING-005: Control Implementation Statements Reference "17 controls" in Handoff Context (MINOR)

The CMMC STATUS.md (line 83) references "Control Implementation Statements (17 controls)" as an existing asset, but the same commit expanded these to 40 controls. Within the same commit, the reference is stale.

**Remediation:** Update the CMMC STATUS.md documentation assets table to reflect "40 controls" instead of "17 controls."

---

## Quality Assessment by Deliverable

### CMMC STATUS.md — Grade: A-

Thorough strategic document. Correctly identifies the "Level 1 now, Level 2 when triggered" strategy appropriate for a startup without DoD customers. Cost estimates are realistic ($65K-$212K for full Level 2). The phase timeline is well-structured. Cross-framework synergies are correctly mapped. Deducted for the SPRS score inconsistency (Finding-001).

### CMMC Control Mapping (CSV) — Grade: A

97 controls properly extracted from the OSCAL catalog. Cross-framework mappings to SOC 2, ISO 27001, and FedRAMP 800-53 are correct and valuable. Status assessments are honest — the agent didn't inflate "Partial" to "Implemented" or handwave away gaps. The exclusion of 33 withdrawn Rev 3 controls is correctly noted.

### CMMC Gap Analysis — Grade: A

The strongest deliverable in the set. 25 gaps properly prioritized across 4 tiers. Each critical gap has detailed remediation steps with realistic effort and cost estimates. The FedRAMP overlap analysis (Section showing 8 shared gaps with estimated 40-50% effort savings) is particularly valuable for planning. The remediation roadmap is phased logically.

### CMMC SPRS Worksheet — Grade: A-

Detailed per-control scoring with weights, points lost, and POA&M designations. The scoring methodology explanation is clear. The score improvement projections (38 → 48 → 76 → 110 across phases) provide actionable targets. Deducted slightly because the scoring methodology note on line 335-347 shows two different calculations (56 points lost vs. 72 points lost) which, while explained, could confuse a reader.

### FedRAMP SSP Sections 1-3 — Grade: A-

871 lines of substantive SSP content. Section 1 (FIPS 199 categorization) is well-reasoned with the Moderate-to-Low override logic for Model B deployments. Section 2 (system description) accurately reflects the hub-spoke architecture with correct component descriptions. Section 3 (boundary and environment) includes proper ports/protocols tables and interconnection descriptions. Deducted for the baseline declaration issue (Finding-003).

### Appendix E: Digital Identity Worksheet — Grade: A

Excellent technical depth. IAL/AAL/FAL determinations are correctly aligned with NIST 800-63-3. The mapping of Keycloak capabilities to each assurance level is accurate. The AMR claim validation documentation is a useful implementation reference. Configuration guidance tables by deployment type are practical.

### Appendix F: Rules of Behavior — Grade: A

Comprehensive RoB covering all required FedRAMP PL-4 elements: acceptable use, prohibited activities, consequences, acknowledgment template. Correctly references existing policies (ISP-001, ACP-001, CMP-001, IRP-001). The platform-specific sections (workspace usage, API access, admin access) add real value beyond a generic template.

### Appendix L: Separation of Duties — Grade: A

The best handling of the solo-founder SoD challenge I've seen. The 5-role matrix with conflict-of-interest ratings is thorough. The compensating controls matrix (Section 4.2) addresses every identified conflict with specific technical controls and evidence pointers. The organizational growth plan (Section 4.4) showing how SoD improves at 2/3/5/10 employees is exactly what an auditor wants to see.

### Customer-Facing Docs Expansion — Grade: B+

The expansion from 17 to 40 controls and addition of 8 new families is solid. The responsibility model (Aegis Provides / Customer Configures / Customer Implements / Evidence) is consistent and useful. Deducted because the responsibility matrix still doesn't cover all NIST 800-53 families — AT, MP, SA, and PM families are thin or missing. This is noted as remaining work in the buildout report.

### COMPLIANCE_PROGRAM_STATUS Updates — Grade: B+

CMMC quick-status section added correctly. Framework comparison table updated. Deducted for the SPRS score discrepancy (Finding-001) and for not updating the "This Week" checklist to reflect that critical Dependabot alerts were already remediated (the handoff says commit `5ae9973` fixed them, but the checklist still shows them as pending).

### CLAUDE_INSTRUCTIONS Updates — Grade: A

Clean addition of CMMC section with file locations, SPRS recalculation procedure, and control statement update workflow. The instructions are actionable and well-structured for agent use.

---

## Verified Claims

| Claim | Verification | Result |
|-------|-------------|--------|
| 97 controls mapped in CSV | `wc -l` = 98 (97 data + 1 header) | **Confirmed** |
| 25 gaps identified | Counted in gap analysis | **Confirmed** |
| 4 critical gaps | FIPS, SSP, SCRMP, Personnel Screening | **Confirmed** |
| SPRS score = 38 | Math verified: 110 - 72 = 38 | **Confirmed** |
| 40 control implementation statements | Counted AC through SR families | **Confirmed** |
| SSP sections 1-3 drafted | 871 lines of substantive content | **Confirmed** |
| 3 appendices created (E, F, L) | All present and substantive | **Confirmed** |
| Cross-framework references accurate | Spot-checked SOC 2/ISO/FedRAMP mappings | **Confirmed** |

---

## Priority Actions Coming Out of This Review

### Immediate (This Week)

1. **Fix Finding-001** — Align SPRS score to 38/110 across all three documents
2. **Fix Finding-002** — Correct buildout report control counts
3. **Fix Finding-005** — Update "17 controls" reference to "40 controls" in CMMC STATUS

### Soon (This Month)

4. **Fix Finding-003** — Clarify SSP baseline declaration
5. **Schedule ISO 27001 internal audit** — This remains the #1 blocker (overdue since Feb 28)
6. **Run weekly check Mar 16** — Keep evidence cadence on track
7. **Triage 86 non-critical Dependabot alerts** — Create remediation plan with severity-based SLAs

### This Quarter

8. **Q1 Access Review** — Due Apr 15 (SOC 2 CC6.4)
9. **Monthly evidence collection** — Apr 1
10. **Begin SOC 2 auditor selection** — Evidence is sufficient (6+ months by June)

---

## Summary

The agent produced 3,964 lines of compliance documentation across 13 files in a single commit. The work is thorough, technically accurate, and strategically sound. The CMMC program structure follows industry best practices for a startup at this stage, and the FedRAMP pre-work positions Aegis well for when an agency sponsor materializes.

The five findings are all documentation consistency issues — no architectural, strategic, or technical errors were identified. After correcting the findings, this body of work is ready for use in pre-assessment conversations with C3PAOs and 3PAOs.

---

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-09 | Compliance Supervisor (Claude Opus 4.6) | Initial review of commit bb22086 |
