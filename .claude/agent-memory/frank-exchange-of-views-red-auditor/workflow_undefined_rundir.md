---
name: workflow-undefined-rundir
description: FEOV workflow script can pass a literal "undefined" run directory to red — abort and report, never fabricate an audit
metadata:
  type: project
---

The frank-exchange-of-views workflow script invoked red with run-directory paths of literally `undefined/blue/report.md` and `undefined/red/candidates/...` (uninitialized variable in the caller). No run directory, blue report, or debate.md existed in the repo (2026-07-12, branch port-plan-review).

**Why:** An audit with no audit surface must hard-fail — writing findings against a nonexistent report, or creating a literal `undefined/` directory, would mask the harness bug and launder a fake verdict into the debate.

**How to apply:** If any invocation path contains `undefined`/`null` or the named living report does not exist, return FAIL with friction naming the uninitialized variable; do not create the bogus path, do not audit substitute documents unprompted.

**Update 2026-07-12 (round 2):** Defect recurred — caller re-dispatched round 2 with the same literal `undefined` paths; no preflight guard was added despite round-1 adjudication in `undefined/debate.md`. Blue's round-1 report never landed on disk (environment blocked the write; content traveled in its envelope only), so even the partial `undefined/` tree contains no auditable report. Expect this to repeat until the workflow script binds the run-directory/topic variables; keep appending round positions to the existing `undefined/debate.md` transcript but create no new files under `undefined/`.
