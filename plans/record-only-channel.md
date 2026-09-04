# The record as the only inter-agent channel (#62) — design + staging

> STATUS 2026-09-02: shipped — historical record, except the orchestrator-mechanics channel (Stage 2.5, marked NOT BUILT below). Stages 1–3 are production: hand-written debate.md is gone (`internal/setup/setup.go` renders it as a projection, `show --view debate` in `internal/view/view.go`), envelopes carry routing refs only (debate.js says "ROUTING REF ONLY — no prose"), and `verify` exists as a root read-only command.

**Decision (2026-07-23):** the clean one-way answer. Exactly ONE channel for inter-agent
CONTENT: the event record. `debate.md` becomes a rendered, read-only VIEW (never hand-written);
envelopes carry orchestration REFS (gap ids, "rule on these"), never a second copy of the content.

## The three channels today, and which is the violation
1. **Events** (mint, finding, position, closing, dispute, opinion, avenue, outcome) — the durable
   record. KEEP; make it the only content channel.
2. **`debate.md`** — hand-written `### RED/BLUE/LEAD` sections + `### RED/BLUE CLOSING` per gap.
   This is the violation: an agent→agent CONTENT channel parallel to the record. RETIRE.
3. **Envelopes** (grade_disputes, dispute_responses, closures, manifest) — the seat's function
   return to the sandboxed orchestrator (debate.js). LEGITIMATE as orchestration, but today it
   also carries dispute CONTENT (a second copy). Trim to REFS/orchestration only.

## Key enabling fact
`render.go` ALREADY builds the transcript projection from events (position/closing/opinion,
~L371-387) — same shape assemble.go's `debate()` renders. The tool can already show the debate
FROM the record; seats just don't emit the events. So this is emit-wiring + read-redirect +
orchestration-trim, NOT a from-scratch projection.

## Verbs that already exist (no schema work)
`closing --id --reason` · `dispute --id --dimension --proposed --reason` (blue) ·
`dispute-respond` (merge) · `position` · `opinion`. All event through the tool.

**CORRECTED 2026-09-02:** the dispute vocabulary later collapsed into motions — `motion` file/rule/appeal (`internal/cli/motion/verbs.go`); `dispute`/`dispute-respond` are no longer verb names. `closing`, `position`, `opinion` remain event types (`recordpb`).

## Staging — each stage reaches a CONSISTENT state and is run-verifiable

### Stage 1 — closings + positions onto the record (box 5 core)
- Prompts (red-merge, blue-respond): EMIT `closing` events (--id <gap> --reason <argument>) per
  closing argument, and the round narrative as `position` events — instead of appending
  `### CLOSING`/`### RED`/`### BLUE` to debate.md.
- Report/transcript fills in automatically (render.go + assemble debate() already read these).
- Confirm `show --view debate` (or equiv) renders the transcript from events for seat reads.
- Seats still READ debate.md this stage (transitional) — retired in Stage 3.
- round_record_appended: attest the EVENTS exist (not the debate.md section).
- VERIFY: a contested run — closings/positions on the record; report debate section populated
  per gap. (Blue must actually contest — pick a topic/lane setup that produces pushback.)

### Stage 2 — disputes onto the record + docket derives from refs
- Prompts: blue EMITS `dispute` events (content: gap_id/dimension/proposed/evidence); red EMITS
  `dispute-respond` events. Envelope grade_disputes shrinks to REFS (gap_id+dimension) so debate.js
  can still route the docket — content is the event, not the envelope.
- debate.js orchestration reads the refs (unchanged control flow; lighter payload).
- VERIFY: a run with grade disputes — dispute/dispute-respond on the record; docket still routes;
  report shows the dispute thread.

### Stage 3 — retire debate.md as a write AND read target
- Prompts: seats READ the transcript via the tool view (render.go projection), not `cat debate.md`.
- Remove every debate.md write instruction; remove the debate.md existence checks (replace with
  event-existence). Stop creating debate.md (setup skeleton).
- `debate.md` is gone; the transcript is `show --view debate`, rendered from the record.
- VERIFY: a run with NO debate.md on disk; seats coordinate purely through the record.

## The orchestrator on the record (added 2026-07-23 — the sharpest gap)
The lead is split: `debate.js` (mechanics — records NOTHING) + lead-judge agent (judgment —
recorded). In a record-only world the mechanics DECISIONS must also be on the record, or the
orchestration that shaped the run is unauditable. Record: docket composed (which gaps/disputes,
which traffic class, why), round opened/closed, deadlock determined, verdict computed (+
deadlocked/exhausted). Sandbox constraint: the script can't call the tool, so a lead-mechanics
command set is emitted through an AGENT PROXY (the lead-judge already runs each round and at
terminal — it records the mechanics events debate.js hands it in its prompt). New channel:
`lead` (mechanics), distinct from `bench` (judgment). This is what makes the record COMPLETE —
both the agents' content AND the routing. Likely its own stage (Stage 2.5) or folded into Stage 3.

**NOT BUILT (2026-09-02):** no `lead` mechanics channel exists — `internal/record/roles.go` roleSeats holds operator/lens/merge/blue/bench only, and the schema (`recordpb/record.pb.go`) has no docket-composed/round-opened/deadlock event types. Pieces landed by other routes: halts are on the record (#333) and the verdict is derived rather than computed off-record (#309).

## Diagnostic / cross-check surface — READ-ONLY (added 2026-07-23)
A first-class `verify` / `stats` command set over a run's record, replacing ad-hoc grep/python
forensics. STRICTLY READ-ONLY — inspection, never a second mutation path (one-way holds).
- `verify <run>`: assert invariants — board.open == report open == any derived count (WOULD HAVE
  CAUGHT #83); every gap has a terminal disposition; every found_by/gap_id ref resolves; PASS ⟹
  no open gaps (#67 post-hoc); no dual-write divergence (debate.md vs events) DURING the transition.
- `stats <run>`: event-type tallies, coverage (findings→gaps, minted vs un-minted, dialectic per
  gap), so a run's completeness is a number, not an eyeball.
  **CORRECTED 2026-09-02:** shipped folded into `verify`, not as a separate command — `internal/cli/verify.go` prints `verify.Run` checks plus `verify.Compute` stats in one report.
- Pairs with the retirement: as debate.md goes, `verify` is HOW we trust the record is complete.
- Sequence: build `verify` EARLY (even before Stage 1) — it becomes the run-verification harness
  for every subsequent stage, turning "inspect by hand" into "run verify."

## Cross-cutting
- One-way guard: after each stage, grep the prompts — no seat writes CONTENT anywhere but a verb.
- Each debate.js change = rule-sweep trailers (Rule-Class/Sibling-Sweep) + prompt-golden regen.
- The binDir-gated clauses are NOT golden-covered (#80) — Stage-N prompt emit-wiring verifies in
  the RUN, not goldens. State that in each PR.
- Ties: #66/#77 (Stage 1), #62 (the whole), #87 (friction verb — same "emit, don't route around"
  pattern; fold friction-emit into Stage 1 or its own step).

## Validation loop (every stage)
cd tools: go build ./... ; go test ./... -count=1 ; (UPDATE_GOLDENS for difftest if tool changed)
cd .. : node --test tests/simulator/*.test.mjs
REMOVED 2026-09-02: `node scripts/rule-sweep.mjs` no longer exists; superseded by the Go `scripts/rulesweep`.
then: a driven run (local binary + branch debate.js) exercising the stage's events; inspect the
assembled report + `show --view debate` on a FRESH record.
