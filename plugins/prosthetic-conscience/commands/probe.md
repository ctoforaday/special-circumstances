---
description: Harness-spike probe (Phase 1) — verify prosthetic-conscience's skill/agent/command/hook wiring and the communication model are live.
---

Run the Phase 1 harness probe. Report PASS/FAIL per primitive. Model terse-communication: no filler.

1. **Command** — confirm this loaded from `prosthetic-conscience` (if you are reading it, it did).
2. **Agent + skill preloading** — spawn the `probe` subagent. Require it to (a) quote the `critical-stance` verification-probe sentence verbatim (proves `skills:` preloading reached the subagent), and (b) write the line `phase-1 probe ok` to `.claude/spike/probe-agent-write.txt`.
3. **Hook** — read `.claude/prosthetic-conscience-hook.log`; check whether the subagent's write was recorded (this is the test of whether plugin hooks fire *inside subagents*).
4. **Voice** — confirm `terse-communication` and `design-by-contract` are discoverable, and that the probe agent's report read as terse (no filler/narration).
5. **Report** a table: skill / agent / command / hook / voice → PASS or FAIL, one terse line each. Be explicit about FAILs — this probe surfaces breakage, it does not reassure.
