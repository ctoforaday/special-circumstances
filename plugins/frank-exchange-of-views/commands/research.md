---
description: Run the research debate engine — additive blue team vs gate-keeping red team, judged termination, full adversarial record preserved.
argument-hint: <topic> [--lanes N] [--max-rounds N]
---

Run a frank exchange of views on the topic in `$ARGUMENTS` (if no topic given, ask — do not guess). Parse optional flags: `--lanes` (blue candidate drafts, default 3) and `--max-rounds` (safety ceiling only — the real terminator is red-PASS or judged deadlock; default 12). Red's citation-verification passes scale automatically with the report's claim count.

1. Create the run directory: `research/<yyyy-mm-dd>_<short-slug>/` (with `blue/candidates/` and `red/candidates/` subdirectories) in the current project. Seed empty `debate.md`.
2. Invoke the **Workflow** tool with `scriptPath` = `${CLAUDE_PLUGIN_ROOT}/skills/research-protocol/scripts/workflow.js` and `args` = `{ "topic": "<topic>", "runDir": "<run directory path>", "lanes": N, "maxRounds": N }`.
3. AFTER the workflow returns, relay its envelope verbatim (verdict, rounds, outstanding gaps) plus the run-directory path — YOU MUST NOT re-summarize the report's content; the report is the deliverable, and it is for the human.
4. If the verdict is UNVERIFIED, say so plainly with the outstanding gap count — the gate never soft-passes, and neither does the relay.
