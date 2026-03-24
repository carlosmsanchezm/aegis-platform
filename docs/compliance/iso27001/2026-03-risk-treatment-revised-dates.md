# Proposed Revised Risk Treatment Due Dates

**Prepared:** 2026-03-23
**For Approval By:** Carlos Sanchez
**CA Reference:** CA-005 (overdue risk treatments, due 2026-03-31)
**Status:** PENDING APPROVAL

---

## Context

10 of 17 active risk treatments have passed their original due dates. The original dates (set 2026-01-17) were aggressive given the solo founder is simultaneously building the platform for pilot release, pursuing Iron Bank/FedRAMP readiness, and managing the compliance program. The revised dates below are realistic given:

- **Single-founder bandwidth** — ~10 hrs/month available for compliance operations
- **Current priority** — Government submissions (Iron Bank, FedRAMP prep) and v1.0 pilot release
- **Risk severity** — Treatments are re-prioritized by actual risk impact, not original scheduling order

---

## Proposed Revisions

| Risk ID | Description | Original Due | Days Overdue | Revised Due | Justification |
|---------|-------------|-------------|-------------|-------------|---------------|
| **RISK-007** | Credential exposure in code | 2026-02-15 | 36 | **2026-04-15** | Secret scanning is enabled (treatment partially complete). Remaining: pre-commit hooks + rotation procedure. Low urgency — 0 secret scanning alerts found. |
| **RISK-008** | Phishing attack on founder (FIDO2 keys) | 2026-02-15 | 36 | **2026-05-15** | MFA active on all accounts (partial treatment). Remaining: purchase YubiKeys ($50-100) and register. Deferred to after ISO audit scheduling but flagged as management review action item. |
| **RISK-011** | Unpatched developer workstation | 2026-02-15 | 36 | **2026-04-30** | macOS auto-updates are enabled (partial treatment). Remaining: document approved software list + monthly verification procedure. Low effort but low priority vs. platform work. |
| **RISK-014** | Container vulnerabilities (Trivy in CI) | 2026-02-28 | 23 | **2026-04-30** | No progress — Trivy not yet added. Elevated priority due to DISA STIG findings and Iron Bank readiness requirement. Aligned with Priority 5B (container scanning cadence expansion). |
| **RISK-002** | Solo founder SPOF | 2026-03-01 | 22 | **2026-06-30** | Runbooks in git, password manager configured (partial). Remaining: key person insurance evaluation, emergency contact procedure. Inherently residual risk that decreases with first hire. |
| **RISK-003** | No corporate identity provider | 2026-03-01 | 22 | **2026-07-31** | Low urgency — solo founder with MFA on personal accounts. Google Workspace setup deferred until either first hire or customer requirement triggers it. Budget approved but not yet needed. |
| **RISK-004** | GitHub personal account limitations | 2026-03-01 | 22 | **2026-07-31** | No functional impact currently. Org migration is beneficial for audit optics but not a security blocker. Aligned with RISK-003 timeline (do together with Google Workspace). |
| **RISK-010** | Laptop theft/loss exposure | 2026-03-01 | 22 | **2026-05-31** | Find My Mac enabled, FileVault on, biometric auth active (partial). Remaining: document recovery procedure + verify Time Machine encryption. Low effort, scheduling around pilot release. |
| **RISK-013** | Insufficient logging | 2026-03-01 | 22 | **2026-04-30** | CloudTrail enabled (partial). Remaining: multi-region enablement + log retention config + monthly spot-check procedure. Moderately important for FedRAMP AU family controls. |
| **RISK-016** | Excessive API permissions | 2026-03-01 | 22 | **2026-05-31** | No incidents from current permissions. Remaining: IAM policy audit + GitHub token scoping + credential rotation procedure. Important for access review evidence but not blocking. |

---

## Summary

| Category | Count | Revised Timeline |
|----------|-------|-----------------|
| Due by Apr 15 | 1 | RISK-007 (credential exposure) |
| Due by Apr 30 | 3 | RISK-011, RISK-014, RISK-013 |
| Due by May 31 | 3 | RISK-008, RISK-010, RISK-016 |
| Due by Jun 30 | 1 | RISK-002 |
| Due by Jul 31 | 2 | RISK-003, RISK-004 |

All 10 overdue treatments now have revised dates within the next 4 months. The most security-impactful items (RISK-014 container scanning, RISK-007 credential exposure, RISK-013 logging) are front-loaded.

---

## Approval

| | Decision |
|---|----------|
| [ ] | **Approved** — Update 04-risk-treatment-plan.csv with revised dates |
| [ ] | **Approved with modifications** — (note changes below) |
| [ ] | **Rejected** — (provide alternative direction) |

**Modifications (if any):**

**Signature:** _________________ **Date:** _________

Carlos Sanchez, CEO / Information Security Manager
