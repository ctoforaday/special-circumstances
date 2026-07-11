---
description: Harness-spike probe (Phase 1) — verify prosthetic-conscience's skill/agent/command/hook wiring is live.
---

Run the Special Circumstances Phase 1 harness probe and report a pass/fail for each of the four plugin primitives.

1. **Command** — confirm this instruction loaded from the `prosthetic-conscience` plugin (if you are reading this, it did).
2. **Agent + skill preloading** — spawn the `probe` subagent. Ask it to (a) quote the `critical-stance` verification-probe sentence verbatim to prove `skills:` preloading reached it, and (b) write the line `phase-1 probe ok` to `.claude/spike/probe-agent-write.txt`.
3. **Hook** — read `.claude/prosthetic-conscience-hook.log` and check whether the subagent's write to `probe-agent-write.txt` was recorded there by the `PostToolUse` hook.
4. **Summarize** a table: primitive (skill / agent / command / hook) → PASS or FAIL, with a one-line note each. Be explicit and honest about any FAIL — this probe exists to surface breakage, not to reassure.
