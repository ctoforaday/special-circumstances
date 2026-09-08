# Derived seat identity — the last facts a seat asserts about itself

> STATUS 2026-09-08: **required pre-v2.0.0** (gblock). Blocks the three `--v2.0.0` tags and #785.
> Board: #792. §III.1 and §III.4 are SHIPPED (#831); §III.2 and §III.3 are the remaining work.
>
> Rewritten 2026-09-08 after two `/plan-audit` FAILs. **The ROUND is no longer in this plan** —
> gblock ruled it roundless and it moved to `plans/roundless.md`, which owns #753 entire. This
> plan owns the ATTESTED half of identity: what `agent_type` carries and what the cast refuses.
> The two audits are credited in place; both found real defects and the second killed a mechanism
> this plan had already committed to.

## I. Summary & Goals

`roster.go` states the defect against itself: **"register is the one call that takes a seat's word
for who it is."** Everything after register already derives — `BoundSeat` joins `agent_id` to the
register event and `ResolveSeat` refuses a `--seat-id` that disagrees. The self-assertion survives
at exactly one call, and every fact after it inherits from there.

**Goal: no fact about a seat is recovered from the shape of its name.** Not "the seat id is never
typed" — §III.3 keeps a typed id on the unattested path, deliberately and refusably.

| fact | source before | source after |
|---|---|---|
| role | `agent_type`, attested (#674) | unchanged |
| area | a substring of the typed id | `agent_type` — one configuration per seat (SHIPPED #831) |
| **round** | `RoundOf`'s `-r(\d+)` over the typed id | **removed from the record entirely** — `plans/roundless.md` |
| which dispatch | `agent_id`, injected | unchanged |

### Success criteria, stated as checks rather than adjectives

1. `register` refuses `red-lens-evidence` for a run whose cast does not include it, and refuses a
   `--seat-id` disagreeing with the record's binding for that `agent_id`. Both have a named test.
2. `agent_type` alone answers WHICH red seat is acting, for every red seat the engine dispatches
   — held by `TestEveryDispatchedAgentTypeIsAttestable` and `TestTheLensAreasMatchWhatTheEngine
   Declares` in all directions.
3. An archived run written before this change replays with identical projections.

(The round's criteria — `RoundOf` gone, no dispatched id matching `-r\d+` — belong to
`plans/roundless.md` and are stated there.)

### Why the round goes rather than gets derived — the audit's finding, and #753's

The previous draft said the round would come from `CurrentRoundOf(record)`. **It cannot.**
`CurrentRoundOf` is `max(round)` over NON-register events and skips `EVENT_TYPE_REGISTER`
deliberately (`internal/record/inquiry.go:273`, with a measured incident in its comment).
`debate.js:833` increments the round and dispatches the lens `parallel()` at `:891` with no round-N
event yet on the record — so every seat registering in round N would derive N-1, on every round,
and siblings inside one phase would derive different values depending on whose first work event
landed first. Off by one AND racy. (Found by `/plan-audit`, 2026-09-08.)

The fix is not a better clock. **#753 argues the clock should go**, and says exactly this: *"The
round stops being part of identity. `red-lens-r3-L2` → `red-lens-L2`. `RoundOf`, `roundlessShapes`,
`terminalSeats`, the `-r\d+` roster grammar, #676's cross-check … all exist to recover 'what round
is this' from a string. Roundless, the question does not arise."*

This plan takes #753's **identity half** and leaves its scheduling half alone. That is a complete
concept rather than a truncated one: after it no seat id carries a round, and whether the ENGINE
orchestrates in rounds is an independent question #753 owns.

### Why it is a release boundary, and the one honest hole in that argument

The seat id's shape is written into every event. A v2 record whose ids carry `-r<N>` and a v2.1
record whose ids do not are two vocabularies for one field, with nothing on a row saying which.

**The hole, stated because the audit found it and it is real**: the run record already carries a
purpose-built epoch for this reader — `eventSchema` (`record/schema_gen.go`, stamped into
`inputs/run-config.json` at setup, and deliberately NOT a binary version per #597). Bumping it
lets a reader tell the vocabularies apart with no tag boundary at all. So the plugin tag is not
the only carrier available, and §III.2 **bumps `eventSchema` as well** — that is what makes an
archived run readable rather than merely unbroken. The tag remains the boundary because gblock
ruled the change pre-v2.0.0, and because the binary that reads both vocabularies ships on it.

### What this is NOT

- **Not the deletion of `seat_id`.** It stays the key namespace; `deriveKey` scopes the per-seat
  ordinal to it and every SQL projection groups on it.
- **Not roundless orchestration.** #753's scheduling half is untouched.
- **Not the removal of `event.round`.** The field stays; its VALUE changes source (§III.2).
- **Not a change to gap ids.** `R<round>-N` still carries a round; #753 owns that.

## II. Technical Context

Re-verified against the `build/blue-lane-configs` tip on 2026-09-08 by opening each cited line.
The previous draft's §II was audited and four of its claims were stale; every citation below was
re-read rather than carried forward.

**The register path.** `internal/cli/seat/seat.go:247` `Of()` → `ResolveSeat(seatID,
BoundSeat(run), record.RoundIn(run))` at `:264`. `BoundSeat` (`:234`) returns nil unless the run is
valid AND `seatenv.AgentID() != ""`; otherwise `record.SeatOfAgent(run, AgentID())`. `ResolveSeat`
(`internal/seatenv/identity.go:120`) prefers the bound value and refuses a disagreeing flag with
`feov.Conflict` at `:138`.

**The round path.** `RoundIn(run)` → `RoundOf(seatID)` → `roundRe = -r(\d+)`
(`internal/record/round.go:8`). `RoundOf` answers 0 for `frontier`, `blue-synthesize` and
`blue-lane-\d+` (`round.go:23,27`); `RoundIn` derives terminal seats' rounds from the record
(`round.go:62`) — *"a derivation from FIELDS rather than from a name, which is the distinction this
whole exercise is about."* The pattern exists; it was applied only where a name could not answer.

**The sitting ordinal has a documented home and is NOT yet implemented.** `record.proto:451-453`,
directly above `optional int32 round = 5`: *"The ordinal is scoped to the seat now, so its keys are
monotonic across dispatches and the collision is unrepresentable. 'Which dispatch was this' is
still answerable — it is the count of that seat's register events at or before the row, which is
what capture.go already does."* **The last clause overclaims**: `capture.go:829-835` takes the
EARLIEST register per seat for span purposes and counts nothing. The events are on the record; the
counter is new work (§III.2).

**The attestation table** is at `internal/record/agentrole.go:30` and holds 11 keys — seven
`red-lens-<area>`, `red-chair`, two blue, one bench. Its sets did NOT collapse to scalars, and
`agentrole.go:24-25` says why (*"The sets remain sets because blue's are genuinely one-to-many"*).

**Lane count** is `lanes = 3` at `debate.js:56`, floored at 3 by `:82` — **and the floor is
overridable** via `laneFloorOverride`, so the domain is "any positive integer, with a stated reason
below 3". It reaches Go only as `Lanes *string` in `inputs/run-config.json`
(`internal/setup/run.go:315`), nil when unset, with the default living in JavaScript.

## III. Proposed Changes (the spec)

### Consumer census — commands and their output, not a recollection

```
$ grep -rnw 'RoundOf' --include='*.go' plugins/frank-exchange-of-views/tools | grep -v CurrentRoundOf
  16 sites. Production: seatenv/identity.go:117,159 · record/round.go:29,46,69 ·
  cli/merge/mint.go:45 · consistency/consistency.go:473
  Release gate: releasegate/fuzz/fuzz_test.go:998
  Tests: record/replay_test.go:121,123,126 · record/record_test.go:168 ·
  seatenv/identity_test.go:73 · comments at seatenv/identity.go:15,18 and record/record.go:121

$ grep -rn 'RoundIn(' --include='*.go' plugins/frank-exchange-of-views/tools | grep -v 'func RoundIn'
  84 sites. Production: cli/seat/seat.go:264,519. The other 82 are tests.
```

`cli/merge/mint.go:45` is the site the previous draft missed, and its comment is this plan's own
argument made two months earlier: *"The round comes from the seat's CONTEXT, not from re-reading
its id … a gap id is the run's primary public identifier … it must be minted from the FACT the
dispatcher supplies, not from a guess about a string's shape."*

### Tree

```
plugins/frank-exchange-of-views/
  skills/research-protocol/scripts/debate.js   [MODIFY] seat ids lose -r<N>
  tools/internal/record/round.go               [DELETE] RoundOf, synthesisSeats, terminalSeats, laneRe
  tools/internal/record/roster.go              [MODIFY] shapes lose \d+ except blue-lane
  tools/internal/record/sittingordinal.go      [NEW]    count of a seat's registers at or before a row
  tools/internal/record/castenum.go            [NEW]    §III.3's admissible-id enum
  tools/internal/record/schema_gen.go          [MODIFY] eventSchema 3 -> 4 (generated)
  tools/internal/cli/seat/seat.go              [MODIFY] ResolveSeat loses its inferRound argument
  tools/internal/cli/merge/mint.go             [MODIFY] round from context; intent unchanged
  tools/internal/consistency/consistency.go    [DELETE] :473 cross-check (closes #676)
  agents/red-lens-*.md, agents/red-chair.md    [MODIFY] III.4 residue
```

### III.1 One agent configuration per red seat — SHIPPED (#831)

Seven lens areas and the chair each have their own; the ~60 shared lines live in the
`adversarial-audit` skill; `repotree.ConstitutionText` is what the constitution gates read.
**11 agent files**, of which 7 are area-bound by `TestTheLensAreasMatchWhatTheEngineDeclares`. The
picker grows from 4 feov entries to 11 — checked rather than assumed: agent frontmatter here uses
five keys and nothing in-tree marks an agent internal.

Delivery is a declared skill rather than `@include` (unevidenced in this repo) or a generated file
(deterministic assembly order, unmeasured gain). §V.1 schedules the measurement that would
overturn it.

**Blue's lanes are NOT part of this and cannot be.** A lane index is an OPEN count where a lens
area is a CLOSED set — `lanes` is a run parameter with an overridable floor and no ceiling, and
`LANE_METHODS[i % len]` wraps, so a per-index configuration would have to exist for every count an
operator might ask for. They do not need one: §III.3's enum bounds them from the run's own
parameter. Red is ATTESTED; blue's lanes are DECLARED-and-refused.

### III.2 The round — MOVED to `plans/roundless.md`

This section specified deriving the round at register. It is deleted rather than revised, and the
route there is worth keeping because it cost two audits:

1. The first draft named `CurrentRoundOf`. `/plan-audit` found it is `max(round)` over NON-register
   events and skips `EVENT_TYPE_REGISTER` deliberately (`inquiry.go:273`), so every seat would have
   derived N-1 on every round, and siblings inside one `parallel()` would have derived different
   values. Off by one AND racy.
2. The second draft replaced it with a per-seat SITTING ORDINAL — the count of that seat's register
   events at or before the row. `/plan-audit` found `event.round` is compared ACROSS SEATS in
   shipped SQL (`schema.sql:828` chair-round against bench-round; `:1001-1004` a board-mass axis),
   so a per-seat ordinal cannot carry it.

Both attempts failed the same way: they tried to preserve a global clock while removing the string
it was recovered from. **The clock itself is the defect** (#753), so it goes — and with it the
`-r<N>` in every seat id, which is what this plan wanted.

`plans/roundless.md` §III.A is the pre-tag half (ids, `event.round`, gap ids, the views) and is
what unblocks this plan's goal. Its §III.B is the scheduling half and is post-tag.

### III.3 The declared path is REFUSABLE [NEW] — ruled by gblock 2026-09-07

When `agent_type` is absent — an operator at a shell, CI, the Go suite, the simulator —
`--seat-id` is accepted, and checked against two things it can fail:

1. **The run's CAST, as a closed list.** Not a shape: `red-lens-evidence` is refused for a run
   whose cast does not include it. The cast is computed at setup from the selected areas and the
   lane count and **written to the record**, rather than recomputed per call — which is what makes
   it a record fact rather than a Go copy of a JavaScript default. (The previous draft said "the
   record carries it" while the lane count lived only in a nullable JSON field with its default in
   `debate.js`. Writing the cast fixes that, and is what #752 asks for.)
2. **No conflict with the record's binding** for this `agent_id` — `ResolveSeat`'s existing
   `feov.Conflict` (`identity.go:138`), which today guards every act EXCEPT the one that creates
   the binding.

**Which seat kinds the enum covers, and which it cannot.** The previous draft claimed "the run's
admissible seat ids are enumerable" without qualification, and that was false:

| kind | enumerable at setup? |
|---|---|
| `red-lens-<area>`, `red-chair`, `blue-respond`, `judge`, `frontier`, `blue-synthesize`, `judge-terminal`, `assemble` | **yes** — fixed by the selected areas |
| `blue-lane-<N>` | **yes** — bounded by the run's lane count |
| `judge-petition-<petitioner>` | **derived, not listed** — admissible iff the tail is itself a cast member; `roster.go:122-131` already recurses |
| `operator` | **no, by construction** — not a party to the debate and has no agent configuration. Stays shape-checked only, stated rather than papered over |

**No `maxRounds` bound.** `roster.go:39-43` considered and REJECTED bounding the round, because a
resume legitimately lowers the ceiling and the bound would refuse seats from the run's own earlier
rounds. Roundless the question does not arise — there is no round in the id to bound. The rejection
is answered by deletion rather than overruled.

### III.4 `red-merge` → `red-chair` — SHIPPED (#831), with residue

Seat id `red-chair`, agent type `frank-exchange-of-views:red-chair`, **role stays `merge`**:
`chair` is already this package's word for a SIDE of the debate (`ChairOf` maps a role to
red/blue/bench; the operator command takes `--chair` over that vocabulary), so a role named
`chair` would sit in a chair.

**Residue the audit found in shipped work, fixed as part of this plan** — the carrier list omitted
the agent bodies:

- `agents/red-lens-*.md` still say *"A LENS FINDS; **THE MERGE** SPEAKS FOR THE ROUND"*, and
  `agents/red-chair.md` says *"AT THE MERGE SEAT"*. Eight shipped prompts instructing seats in the
  retired vocabulary.
- `promptCatalogue` at `integration/surface/promptverbs_test.go:601` and its duplicate at
  `releasegate/fuzz/promptverbs_test.go:601` still pin `"red-auditor.md": 0` — a deleted file — and
  list none of the eight new configurations, so by that catalogue's own argument each falls to a
  default instead of a decision.

## IV. Risk & Mitigation

| risk | mitigation |
|---|---|
| **Two sittings of one seat id are conflated by a reader that assumed uniqueness.** This is the work the old `-r<N>` was really doing. | The sitting ordinal is the discriminator and is on every row. The §III census is the reader list; each is checked. `events.key` already tolerates re-registration by construction — the collision it used to cause is recorded at `record.proto:446-449`. |
| ~~An act lands after its round advances and is stamped with the new round~~ | **Dissolved, not mitigated.** That mode belonged to a global clock read at register. A per-seat register count cannot be advanced by a sibling and cannot be read before the seat's own register exists. |
| **Archived runs stop projecting.** | Readers take both shapes; `eventSchema` 3 → 4 makes the vocabularies distinguishable on the record. §V replays an archived run and diffs projections. |
| **The cast written at setup drifts from what the engine dispatches** — a new record fact is a new thing to get wrong. | It is written by the same path that computes the dispatch, and §V binds it to `RED_AREAS` and the lane count the way the roster is already bound to the engine. |
| **The untyped population may be large**, making III.3's declared path the normal one rather than the exception. | Measure on the first run. A neighbouring surface (gray-area's `SubagentStop`) found `agent_type` absent on 146 of 165 rows and took four investigations to stop reading that absence as a defect (peer session, 2026-09-07). The base rate does not transfer — different surface — but the shape of the mistake does. |

## V. Verification Plan

Each check with what re-arms it.

```bash
# 1. The regex is GONE. Word-anchored so it does not match CurrentRoundOf (the previous draft's
#    gate matched it and could never pass), and scoped to tools/ so it covers releasegate/fuzz —
#    the one consumer a release most needs caught. Re-arms: any reintroduction.
! grep -rnw 'RoundOf' --include='*.go' plugins/frank-exchange-of-views/tools | grep -v CurrentRoundOf

# 2. No dispatched seat id carries a round. Re-arms: debate.js, roster.go.
(cd plugins/frank-exchange-of-views/tools && go test ./internal/record/ -run TestTheRosterMatchesWhatTheEngineActuallyDispatches)

# 3. THE TWO REFUSALS, one named test each — neither exists today, and neither is covered by the
#    tests that cover III.1. Re-arms: castenum.go, seat.go, roster.go.
#      TestASeatOutsideTheRunsCastIsRefusedAtRegister
#      TestADeclaredSeatIdThatContradictsTheBindingIsRefusedAtRegister
(cd plugins/frank-exchange-of-views/tools && go test ./internal/record/ ./internal/cli/seat/ -run Refused)

# 4. Archived runs replay identically. Re-arms: any reader change.
(cd plugins/frank-exchange-of-views/tools && go test ./internal/record/ -run 'Replay|Archived')

# 5. Full suites and gates. Re-arms: anything under tools/, the engine, agents/, skills/.
(cd plugins/frank-exchange-of-views/tools && go vet ./... && go test ./...)
node --test plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs \
             plugins/frank-exchange-of-views/tests/simulator/prompts.test.mjs
(cd scripts && go run ./check && go run ./archaeology && go run ./rulesweep)
```

**The driveable check, with its oracle named.** "Was any act stamped with a sitting it did not act
in" cannot be answered by reading `event.round`, because after §III.2 that field IS the derived
value — the previous draft proposed comparing it to itself, which answers nothing. The independent
witness is the **engine's dispatch label**, which carries the round the workflow believed
(`red-lens-<area>-r<n> · <slug>`) and is recorded in the transcript rather than by the seat. Join
`seat_turn.agentId` → the register event → the sitting ordinal and compare against the label's
`r<n>`. A disagreement is the failure mode this plan exists to remove; agreement across a whole run
is the evidence §IV's first row wants.

### V.1 The shared-prefix measurement (decides III.1's delivery mechanism)

A red seat's non-conversation input is ~9.3k tokens: `research-protocol` 3.7k +
`adversarial-audit` 3.3k + its own configuration 0.4k, all identical across seats, plus a ~1.9k
dispatch prompt that is per-seat and never shareable. **~7k of ~9.3k is in principle a shared
prefix**, across ~32 red sittings a run — order 224k tokens.

It cannot be mined from `run-archive/` (records and proofs only; the newest archived record has no
`seat_turn` table). `internal/seatturn` parses `cache_read_input_tokens` per turn keyed on
`agentId`, so a live run answers it with no new instrumentation. Read the first turn of each red
sitting and compare siblings:

- **~0 across siblings** — no cross-dispatch prefix caching; generation buys nothing on tokens.
- **~7k from the second sibling on** — caching already works through the skill; ordering is fine.
- **partial or erratic** — ORDER is the lever, which a generated file fixes and `skills:` cannot.
  Build the generator; per-seat bodies stay hand-written and it assembles rather than authors.

## VI. Deliberately not in this plan

- **Roundless ORCHESTRATION** (#753's other half): work-driven dispatch, per-gap exchange counts,
  the turn-report verb. This plan takes only the identity half, which is complete on its own.
- **Gap ids.** `R<round>-N` still carries a round. #753 owns it.
- **Deleting `seat_id`.** §I.
- **Blue lane METHOD as a second carrier.** `LANE_METHODS[i % len]` recovers a lane's method from
  its index — the fork #791 healed for lenses. Per-lane configurations would have fixed it for free
  and III.1 establishes they cannot exist, so #497 has to design it. It is the only place left
  where a seat's meaning is recovered from a number.
- **Refusing an unattested seat outright.** III.3 makes the declared path refusable; making absence
  itself fatal is a later call on evidence this plan produces.
