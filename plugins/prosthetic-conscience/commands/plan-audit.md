---
description: Run an implementation plan through the auditor gate against the spec-driven-development standard — binary PASS/FAIL with actionable gaps.
argument-hint: [path-to-plan-file]
---

Run the auditor gate on the implementation plan at `$ARGUMENTS` (if no path given, ask which plan file to audit — do not guess).

1. Spawn the `auditor` agent, feeding it: the plan path, and the governing standard — the `spec-driven-development` skill (plan structure I–V; §V verification commands must be executable). It audits **Alignment / Completeness / Safety** and returns `VERDICT: PASS|FAIL` with gaps.
2. Relay the verdict verbatim — YOU MUST NOT soften a FAIL or summarize gaps away.
3. On FAIL: offer to fix the gaps in the plan (plan edits need the human's approval per [[semantic-consent]]); AFTER fixing, re-run the gate. The plan is approved only on PASS.
