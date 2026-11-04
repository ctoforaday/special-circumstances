---
name: validation-loop
description: Always-on verification discipline — a written, executable validation loop that outlives the plan; run it before claiming anything works.
disable-slash-command: true
user-invocable: false
---

# validation-loop

The checks outlive the spec. Work drifts beyond the plan; validation must not.

- BEFORE implementing, YOU MUST write down the validation loop — the exact commands that prove the change works (tests, build, lint, a driveable end-to-end check) — in the plan's §V, the project's `walkthrough.md`, the session `CHECKPOINT.md` (see [[context-checkpointing]]), or the session scratchpad. Written, not remembered: the written loop survives context compression when the intention doesn't. Record with each check **what re-arms it** — the file or condition that makes it fire — because that is the first thing a summary drops, and a paraphrase of what a check wants is not the check.
- AFTER any nontrivial change, YOU MUST run the loop and observe the result — YOU MUST NOT claim success from a clean edit, a plausible diff, memory of it working before, or a failure you believe you have since fixed — only the command's output closes it (see [[anti-spinning]] for repair limits). The output closes it only when it proves it exercised the artifact the change produced — the binary, build, install or deployment, by path, build id or a phrase from the change — rather than whichever copy the environment resolved first: a pass against a stale artifact reads exactly like a pass. A pushed change's continuous-integration run is a check in the loop like any other, and a claim of green names the run and its result on the pushed head.
- During work beyond the plan's original scope, YOU MUST keep running the same loop — drifting off-plan is when verification is most often dropped, and this rule exists precisely for that moment.
- AFTER a context compaction, YOU MUST re-read the written loop before resuming — trust the written checks, not the summarized memory of them. Where a gate is involved, YOU MUST **reproduce it** rather than recall what it wanted.
