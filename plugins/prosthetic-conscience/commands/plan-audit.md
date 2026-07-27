---
description: Run an implementation plan through the plan-auditor gate — binary PASS/FAIL against the spec-driven-development standard, with actionable gaps.
argument-hint: [path-to-plan-file]
---

Run the plan-auditor gate on the implementation plan at `$ARGUMENTS` (if no path given, ask which plan file to audit — do not guess).

1. BEFORE spawning, confirm the plan's behavioral forks are resolved ([[spec-driven-development]]: decide forks before the gate) — if an unresolved fork with a behavioral, semantic, cost, or reversibility implication remains, STOP and put it to the human first; do not audit a design that may be replaced. Then spawn the `plan-auditor` agent on the plan file. It audits **Alignment / Completeness / Safety** against the spec-driven-development standard and returns `VERDICT: PASS|FAIL` with gaps.
2. Relay the verdict verbatim — YOU MUST NOT soften a FAIL or summarize gaps away.
3. On FAIL: iterate spec fixes and re-audit **without per-round human approval** — a scratch plan is reversible ([[semantic-consent]]). Human approval attaches to the **final PASSed plan** and to genuine design forks, not to each edit. The plan is approved only on PASS.
