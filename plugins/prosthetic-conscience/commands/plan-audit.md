---
description: Run an implementation plan through the plan-auditor gate — binary PASS/FAIL against the spec-driven-development standard, with actionable gaps.
argument-hint: [path-to-plan-file]
---

Run the plan-auditor gate on the implementation plan at `$ARGUMENTS` (if no path given, ask which plan file to audit — do not guess).

1. BEFORE spawning, confirm the plan's behavioral forks are resolved ([[spec-driven-development]]: decide forks before the gate) — if an unresolved fork with a behavioral, semantic, cost, or reversibility implication remains, STOP and put it to the human first; do not audit a design that may be replaced. Then spawn the `plan-auditor` agent on the plan file. It audits **Alignment / Completeness / Safety** against the spec-driven-development standard and returns `VERDICT: PASS|FAIL` with gaps.
2. Relay the verdict verbatim — YOU MUST NOT soften a FAIL or summarize gaps away.
3. On FAIL: iterate spec fixes and re-audit **without per-round human approval** — a scratch plan is reversible ([[semantic-consent]]). Human approval attaches to the **final PASSed plan** and to genuine design forks, not to each edit. The plan is approved on the **second consecutive PASS** — the confirming round re-audits the unrevised plan, so it costs one audit, not a rewrite. There is NO round cap: silent defects have surfaced as late as round 21. YOU MUST take a fork to the human instead of into another round when (a) a gap's fix adds machinery — build it, or accept the residue as stated — or (b) two rounds' safety gaps all stem from one absolute guarantee — keep it, or go best-effort. Measured: whole runs of rounds went to machinery that was later dropped, and one plan's guarantees drove eight.
4. AFTER approval, the NOTES carry into implementation: the PR body discharges or dismisses each one, and an independent review of the built change is REQUIRED before merge. NOTES are safe to leave to implementation only because that review follows — without it, the audit was the only reviewer.
