---
description: Run an implementation plan through the plan-auditor gate — binary PASS/FAIL against the spec-driven-development standard, with actionable gaps.
argument-hint: [path-to-plan-file]
---

Run the plan-auditor gate on the implementation plan at `$ARGUMENTS` (if no path given, ask which plan file to audit — do not guess).

1. Spawn the `plan-auditor` agent on the plan file. It audits **Alignment / Completeness / Safety** against the spec-driven-development standard and returns `VERDICT: PASS|FAIL` with gaps.
2. Relay the verdict verbatim — YOU MUST NOT soften a FAIL or summarize gaps away.
3. On FAIL: offer to fix the gaps (plan edits need the human's approval per [[semantic-consent]]); AFTER fixing, re-run the gate. The plan is approved only on PASS.
