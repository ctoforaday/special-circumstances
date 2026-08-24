# capture-audit — mechanized post-hoc checks (capture-research-run.mjs)

Presence/consistency tier only: these checks catch a missing line and a self-inconsistent
self-report; a plausible-but-wrong value is vacuity, whose auditor is the next run/
retrospective over these same git-tracked artifacts.

- **liveness: PASS** — the run reached a terminal outcome on the record (CEILING) — it finished under its own power
- **telemetry: PASS** — 4 telemetry round(s) vs 4 red round(s) on the record
- **friction-parity: PASS** — 0 envelope entries joined to a seat, and every one of those seats is on the record (42 friction entries recorded in total)
    33 envelope entries COULD NOT BE JOINED to a seat and were not judged (no agent_id on the record — this run's PreToolUse hook did not fire, so the binding `register` writes was never supplied):
    - Registration-time: the engine's identity-binding PreToolUse hook did not inject the run di
    - cite's generic web fetch cannot reach a GitHub issue page the authenticated mcp__github__i
    - Unused-verb survey (whole-tree walk, none guessed at): line-of-inquiry propose/move unused
    - Register-time: the engine's identity-binding PreToolUse hook did not reach this seat (tool
    - No other capability gap: the class registry lacked a listing verb, but probing mint with a
    - Engine identity-binding PreToolUse hook did not reach this run (register: 'THE ENGINE'S HO
    - cite has no amend/re-title/re-point verb. c-8eb27e2f was fetched round 0 titled 'Auer/Cesa
    - Verb survey (read in the tree, not used, and why): retire — considered for R1-8's misattri
    - PreToolUse identity-binding hook did not reach this run (affects every seat, not just this
    - Explicit --none friction close for the round: nothing else blocked. Noted (not duplicated,
    - Engine PreToolUse identity-binding hook did not reach this seat (register announced 'THE E
    - Standing gap unaddressed from blue's surface: no cite-amend verb exists to correct the R1-
    - No other new capability gap blocked this sitting. Native Read/Grep/Bash sufficed; the W1.1
    - Closed on the record via friction --none: nothing on the bench's own surface blocked this 
    - The identity-binding PreToolUse hook again did not reach this seat at register ("THE ENGIN
    - Self-inflicted, not a tool gap: an early --reason string passed to `position` contained ba
    - Engine identity-binding PreToolUse hook did not register this run (announced at register; 
    - No cite-amend verb: R1-8's bibliography entry c-8eb27e2f names Auer/Cesa-Bianchi/Fischer 2
    - Q11's blue-blocking work item keys on red's standing round-3 'unsupported' vote, which blu
    - Motion M2 (blue's round-3 grade dispute on R3-2's impact) reached this sitting unruled -- 
    - Verbs read in the full tree this sitting and not used, and why: declare (no definitional d
    - Register-time: the engine's identity-binding PreToolUse hook is not reaching this run (aff
    - motion grade rule --as accepted has no persisted consequence on a CLOSED gap: regrade refu
    - Engine PreToolUse identity-binding hook did not reach this run this sitting (register anno
    - No cite-amend verb exists to correct the R1-8 bibliography title misattribution (c-8eb27e2
    - Register's own output instructed this seat to log once that the engine's PreToolUse hook i
    - motion grade exposes no bench-facing 'rule' verb to this seat (only file/appeal) -- alread
    - The opinion --as enum available to this seat (repaired / repaired_with_regression / amends
    - Verbs read this sitting and not used, and why: declare (nothing was a holding-about-how-th
    - Run-wide (recorded once, as register instructed): the engine's PreToolUse identity-binding
    - Sitting-specific, second occurrence this run (first at R3-2 by a prior bench sitting, now 
    - Register-time: the engine's identity-binding PreToolUse hook did not reach this run (regis
    - No new capability gap this sitting beyond two already-argued, recurring ones, both re-flag
- **context-use: PASS** — peak 256k = 26% of its 1000k window (agent a7127d1d859290c6e); 32 seats measured; 0 over the 50% tripwire
- **assembly-screen: PASS** — 11 citation(s), 0 found against by red (refutes|absent), none still cited in the assembled report; 0 cited source(s) nobody has checked
- **stray-records: PASS** — no event shards outside a run directory
- **discarded-events: PASS** — no seat's replay discarded recorded work
- **record-parity: PASS** — 4 red round(s) vs 5 blue sitting(s) and 5 recorded round record(s) (floor: redRounds-1 — a PASS exit has no final blue response)
- **backfill: WARN** — 32 seat(s) on the record; 28 with enough events (>=5) and a known start to measure
  blue-lane-2: 8 event(s) written in 59.222s at the end of a 10m29.201s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence
  blue-lane-1: 16 event(s) written in 1m19.504s at the end of a 9m44.627s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence
  red-lens-r1-L1: 19 event(s) written in 207ms at the end of a 15m11.448s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence
  red-lens-r1-L5: 13 event(s) written in 49.917s at the end of a 6m53.004s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence
  red-lens-r1-L6: 7 event(s) written in 52.593s at the end of a 6m9.147s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence
  red-lens-r2-L5: 9 event(s) written in 1m4.848s at the end of a 7m28.012s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence
  red-lens-r3-L5: 5 event(s) written in 33.24s at the end of a 8m43.606s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence
  red-lens-r3-L6: 7 event(s) written in 29.382s at the end of a 6m4.17s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence
  judge-r3: 6 event(s) written in 40.774s at the end of a 7m53.247s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence
  red-lens-r4-L5: 5 event(s) written in 30.627s at the end of a 8m6.647s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence
  judge-terminal: 6 event(s) written in 21.661s at the end of a 11m5.965s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence
- **attestation-integrity: PASS** — 4/12 anchored closure(s) sampled and reconciled against actual tool calls
- **model-tier: WARN** — bulk=fable, judgment=sonnet; WARN red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted; WARN blue-lane ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted; WARN blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted; WARN red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted; WARN red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted; WARN blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted; WARN red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted; WARN blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted; WARN blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted; WARN frontier ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted

- tarball: 32 transcript(s)
- record archive: run-archive/2026-08-23_sleeper-service-plan.tar.gz (202 KB) — the raw record, which research/ does not keep
- cost.md: written (telemetry join included)
- report.md: cost breakdown folded in (## Cost)
- scorecards: 17 row(s) across 3 chair(s) -> feov-memory/
- precedent harvest: 12 ruling(s) -> /home/user/special-circumstances/.claude/worktrees/sleeper-plan/law/proposed/2026-08-23_sleeper-service-plan.md (PERSUASIVE, awaiting review) [the envelopes claim 11 — the record is the source]
- class harvest: 10 class(es) -> /home/user/special-circumstances/.claude/worktrees/sleeper-plan/law/proposed (PROPOSED, not staged into any run until adopted)
- run-live marker: removed
