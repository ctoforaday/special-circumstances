---
description: Run an implementation plan through the SDD auditor gate — binary PASS/FAIL with actionable gaps.
argument-hint: [path-to-plan-file]
---

Run the SDD auditor gate on the plan at `$ARGUMENTS` (if no path given, ask which plan file to audit — do not guess).

1. Spawn the `plan-auditor` agent on the plan file. It audits **Alignment / Completeness / Safety** per [[spec-driven-development]] and returns `VERDICT: PASS|FAIL` with gaps.
2. Relay the verdict verbatim — YOU MUST NOT soften a FAIL or summarize gaps away.
3. On FAIL: offer to fix the gaps in the plan (plan edits need the human's approval per [[semantic-consent]]); AFTER fixing, re-run the gate. The plan is approved only on PASS.
