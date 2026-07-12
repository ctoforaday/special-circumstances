---
name: validation-loop
description: Always-on verification discipline — a written, executable validation loop that outlives the plan; run it before claiming anything works.
---

# validation-loop

The checks outlive the spec. Work drifts beyond the plan; validation must not.

- BEFORE implementing, YOU MUST write down the validation loop — the exact commands that prove the change works (tests, build, lint, a driveable end-to-end check) — in the plan's §V, the project's `walkthrough.md`, or the session scratchpad. Written, not remembered: the written loop survives context compression when the intention doesn't.
- AFTER any nontrivial change, YOU MUST run the loop and observe the result — YOU MUST NOT claim success from a clean edit, a plausible diff, or memory of it working before.
- During work beyond the plan's original scope, YOU MUST keep running the same loop — drifting off-plan is when verification is most often dropped, and this rule exists precisely for that moment.
- AFTER a context compaction, YOU MUST re-read the written loop before resuming — trust the written checks, not the summarized memory of them.
- AFTER a failure, YOU MUST NOT declare the loop passing until the command output shows it (see [[anti-spinning]] for repair limits).
