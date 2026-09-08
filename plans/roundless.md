# Roundless — the global clock leaves the record

> STATUS 2026-09-08: proposed. Implements #753's argument. Split out of
> `plans/derived-seat-identity.md`, which now owns only the ATTESTED half of identity and hands
> the round here (gblock, 2026-09-08).
>
> **ALL OF IT is required pre-v2.0.0, and a real run must exercise it before the tag** (gblock,
> 2026-09-08). §III.A and §III.B are still stated separately because they fail differently and
> land in that order — but neither is deferred, and the tag waits on both plus the run.

## I. Summary & Goals

**The round is a global clock over work that has per-item lifecycles** (#753). It does three
separate jobs, and they are worth naming separately because they end differently:

| job | where | ends how |
|---|---|---|
| part of a seat's IDENTITY | `red-lens-r3-evidence`, `RoundOf`'s `-r(\d+)` | deleted (§III.A.1) |
| the record's TIME AXIS | `event.round`, 49 lines of `schema.sql`, `R<round>-N` gap ids | replaced by the event axis (§III.A.2-3) |
| the ORCHESTRATION unit | `debate.js`'s round loop, `maxRounds` | replaced by work-driven dispatch (§III.B) |

**Goal: no fact in the record is scoped to a round, and no seat is dispatched because a clock
ticked.**

### Why the axis job is the interesting one

Rounds look like scheduling but the record uses them as a **time axis**, and only for that:

- `schema.sql:828` — `COALESCE(MIN(ce."round", be."round"), …) AS "closed_round"`, the chair's
  round against the bench's. A CROSS-SEAT comparison.
- `schema.sql:1001-1004` — `FROM (SELECT DISTINCT "round" FROM "events" WHERE "round" > 0) r
  JOIN "gap" g ON g."minted_round" <= r."round" AND (g."open" OR g."closed_round" > r."round")`.
  Board mass at each point in time.

Neither needs a ROUND. Both need an ORDER, and `events` already carries a finer and truer one in
its timestamp and per-run sequence. *Open at round N* becomes *open at event T*, which is strictly
more precise: it can answer "what did the board look like when the chair closed R2-3", which the
round-bucketed view cannot.

**This is why a per-seat sitting ordinal cannot substitute for the round** — a fact established
the expensive way, by proposing it and having `/plan-audit` find these two queries. A sitting
ordinal is per-seat; these comparisons are across seats. The axis has to be global, and the event
sequence already is.

### A and B land together, and the run is the gate

An earlier draft made §III.B post-tag on the argument that it changes only BEHAVIOUR while §III.A
changes what a record SAYS. gblock overruled it: **the whole change lands and is tested before the
tag.**

The reason that is the better call, stated so it is not re-argued: §III.A alone would ship a
record vocabulary with no round in it, still produced by a round loop — every seat id roundless
while `debate.js` counts rounds to decide who sits. That is a half-state that reads as done, and
it would be the *shipped* state for however long §III.B took. The two halves are one concept and
the tag is the one boundary that can hold them together.

They still land in ORDER — A then B — because B's dispatch chair reads a board whose axis A
defines, and because A's carriers can be swept while the loop still works.

**The run is not a smoke test, it is the gate.** §V's driveable check is a full debate on real
data, and §III.B has no unit-testable claim worth the name: "does work-driven dispatch reach the
same verdict for less work" is answerable only by running it. That run also carries #792's B4 and
`plans/derived-seat-identity.md` §V.1's shared-prefix measurement — one run, three answers.

## II. Technical Context

Verified against `build/blue-lane-configs` tip, 2026-09-08.

**Where the round enters the record.** `envelope()` (`internal/record/record.go:399`) stamps
`ev.Round` on every event from a value `ResolveSeat` inferred — today via `record.RoundIn(run)` →
`RoundOf(seatID)` → `roundRe = -r(\d+)` (`internal/record/round.go:8`).

**Where it is read.** `grep -rn 'GetRound()' --include='*.go' | grep -v _test` — 30+ production
readers: `record/viewjson.go` (×7), `record/motion.go` (×3), `record/inquiry.go` (×5),
`report/assemble.go` (×3), `record/spotcheck.go`, `record/evidenceview.go`, `record/citationid.go`,
`report/docs.go`, `view/changes.go`, `view/view.go`, `graph/graph.go`, `capture/capture.go`.
`schema.sql` mentions round on **49 lines**.

**Gap ids are round-scoped counters.** `MintGapID` (`record/replay.go:413`) is
`SELECT count(*) FROM "mint" … WHERE e."round" = ?`, formatted `R%d-%d`. The gap id is the run's
primary public identifier — printed in the report, referenced by `--supersedes`, `--id` and
`found_by` — so its scheme is record vocabulary, not an internal detail.

**The workflow cannot schedule from the record.** A Workflow script has no filesystem access and
learns nothing except through `agent()` return values (#753). Today the round loop is the
scheduler and the exit conditions are read out of envelopes the seat composed — `sitting.halt`
(`debate.js:713`), `deadlock`, red's verdict. **That is seat self-assertion in the one channel
the identity work never touched.**

**Concurrency is ~2, measured.** #753: the lens `parallel()` lands at sum ÷ ~1.6 every round —
`min(16, CPUs − 2)` on a 4-core box. Any design that keeps seats resident deadlocks by arithmetic.

## III. Proposed Changes (the spec)

### §III.A — The vocabulary (lands first)

#### A.1 Seat ids lose `-r<N>` [MODIFY]

`red-lens-r3-evidence` → `red-lens-evidence`; `red-chair-r2` → `red-chair`; `blue-respond-r1` →
`blue-respond`; `judge-r2` → `judge`. `blue-lane-<N>` keeps its index: that is a lane, not a clock.

One seat id now covers many sittings, which the record already supports — `events.key` scopes the
idempotency ordinal per seat precisely so a re-dispatched seat does not collide with its own
earlier acts (`record.proto:444-450`, with the measured `UNIQUE constraint failed` that forced it).

Deleted with it: `RoundOf`, `synthesisSeats`, `terminalSeats`, `laneRe` (`round.go`), and
`consistency.go:473`'s cross-check — closing #676, because there is no longer a name to disagree
with a field.

**Petition sittings need a new uniqueness argument.** `debate.js:695-707` states the petition id is
unique *by construction* only because the petitioner's id carries the round: *"each of
`blue-synthesize`, `red-chair-rN` and `blue-respond-rN` petitions at most once"*. Roundless,
`judge-petition-red-chair` repeats. Replaced by the petition's own ordinal on the record — the
count of that petitioner's petition sittings — which is the same shape as every other id here.

#### A.2 `event.round` leaves the schema; the axis becomes the event sequence [DELETE/MODIFY]

`event.round` is removed. The two comparisons that need an axis are rewritten against
`events."id"`/`events."ts"`, which every event already carries:

- `closed_round` → `closed_at`: `MIN` over the closing events' sequence rather than their rounds.
- The board-mass view stops iterating `SELECT DISTINCT round` and iterates the events that CHANGE
  the board (mints and closures), giving a point per state change rather than per round.

Every one of the 30+ `GetRound()` readers is enumerated in the census (§III.C) and classified:
axis-users are rewritten, display-users take the sitting ordinal or drop the column.

#### A.3 Gap ids lose the round [MODIFY]

`R<round>-N` → a run-global monotonic id. `MintGapID`'s round-scoped `count(*)` becomes a
run-scoped one. Archived ids remain readable and are never rewritten; the two schemes cannot
collide because the new one has no `-`.

**This is the most visible change in the plan** and the one a consumer notices first: it is in
every report, every `--supersedes`, every `found_by`.

#### A.4 The dispatch chair joins the cast [NEW]

Named here rather than in §III.B because a new seat is a new agent configuration, a new seat id
and a new attestation row — all vocabulary. Its BEHAVIOUR is §III.B.

#### A.5 Archived runs are MIGRATED, and the dual-shape reader is retired [MODIFY/DELETE]

**Ruled 2026-09-08: no archaeology, and no backwards compatibility outside replay.** Readers speak
ONE vocabulary. An archived run is re-driven through the current write path by `feov-record
migrate` (#840), and the old→new mapping lives THERE, once — not in every reader forever:

| old | new | where the fact goes |
|---|---|---|
| `red-lens-r3-evidence` | `red-lens-evidence` | the sitting ordinal (third register of that seat) |
| `red-chair-r2`, `red-merge-r2` | `red-chair` | the sitting ordinal |
| `blue-respond-r1`, `judge-r2` | `blue-respond`, `judge` | the sitting ordinal |
| `judge-petition-red-chair-r1` | `judge-petition-red-chair` | the petition's own ordinal |
| `R3-2` | its run-global id | position in the run's mint sequence |
| `event.round = 3` | *(dropped)* | the event's own sequence and timestamp |
| `L2-F1` | `<area>-F1` | the area of the seat that filed it |
| `role: merge` on a `red-merge-*` seat | `role: merge` | unchanged — the role never moved |

**This reaches back into merged work.** #791 taught `FindingLabelAlt` and `roleRe` to read `L\d+`
beside area names; #831 taught `roleSeats`, `seatclass`'s rounded table and tier map, and the
dashboard and seatprobe prefix reads to accept `red-merge` beside `red-chair`. Every one of those
is a dual-shape reader, and every one is retired once migration covers its case — the archived
form is migrated at the boundary and the live reader stops knowing it ever existed. They are
enumerated in the census (§III.C) under **retire-after-migration** rather than left as scenery.

**What this depends on, stated so it is not discovered:** #840 leg 1 (record.db archives —
`run-archive/2026-09-02_quadratic-formula.tar.gz` is the only one) and leg 2 (the six older JSONL
archives `store.go:82-92` refuses today). If leg 2 does not land before the tag, those six are
not readable by v2 at all, and that is a decision to record on #792 rather than a surprise.

### §III.B — The scheduling (lands second, same tag)

#### B.1 The dispatch chair, and why it is a seat rather than a verb

**A seat can read the record; the workflow cannot.** That single asymmetry decides the design.
The dispatch chair is dispatched, reads the board, and RETURNS who to engage next; the workflow
does nothing but obey it. It replaces the round loop.

It is the answer to the objection that sank the alternatives: a standing cast deadlocks against a
concurrency cap of ~2 (#753), and a `--wait` mode holds an `agent()` slot while blocked. The
dispatch chair holds a slot only while it is deciding, which is short.

It also closes the self-assertion channel the identity work never reached. Today the workflow
believes what a seat's envelope says is ready. The dispatch chair computes readiness from the
board with the tool, so what gets dispatched stops being the acting seat's own word — the same
move `claim_count` already made when two honest merges differed 2×.

#### B.2 IMPASSE replaces `maxRounds` — the limit before the bench, and before giving up

The global clock is irrelevant once dispatch is work-driven, but the thing it was really bounding
is not: **how long may a dispute go on before someone else decides it, and before we stop.**
Two limits, and only one of them is new.

**IMPASSE — the limit before the bench.** A gap is at impasse when it has been argued N times with
**no movement**. Movement is a fact on the record, not a judgement:

| movement | event |
|---|---|
| the grade moved | `regrade` |
| the text moved | `blue_edit` touching the gap's anchor |
| the lineage moved | a successor minted with `supersedes` |
| it is over | `close` |

An exchange with none of those is unproductive, and N consecutive unproductive exchanges send the
gap to the bench's docket. **Countable from the record; never stored** — which is what makes it
refusable rather than a seat's claim about itself.

**This is a generalisation, not an invention, and the narrow version is already in the tree.**
`debate.js:1077` reads:

> *Carried persistence: a standing carried ruling absorbs re-raises until red's grade moves or a
> successor descends — otherwise the judge re-rules the same question at ~$10-13 a sitting and
> each sitting is a fresh drift chance.*

That is the impasse rule, already cost-justified, applied to exactly one case (a carried ruling),
tested on exactly one kind of movement (grade equality), computed in JavaScript from an envelope
the seat composed, and paced by rounds. B.2 keeps the rule and removes all four restrictions.

It also replaces the docket trigger. Today a gap is docketed for being RE-RAISED (`debate.js:1071`)
— a raw count of two challenges with no test of whether anything moved, which is why the carried
exception had to be bolted on beside it. At impasse, one rule does both jobs.

**DEADLOCK — the limit before giving up — already exists and does not move.** It is the judge's
boolean on the bench envelope (`debate.js:511-513`), and the run's stated end condition is
*"revision rounds until red-PASS or judged deadlock"* (`debate.js:8`). The vocabulary divides
cleanly and neither term is being redefined:

| | scope | who decides | means |
|---|---|---|---|
| **impasse** | one gap | the record, counted | the parties have stopped moving; the bench takes it |
| **deadlock** | the run | the judge | the bench cannot resolve it either; stop |

`impasse` is a free name — checked across the Go tree, the engine, the agents and the skills
before choosing it, because `chair` was not (`ChairOf` already means a SIDE of the debate) and
that collision was caught by reading rather than by any test.

**What terminates the run**, with the clock gone: no gap is open below its impasse limit, and
every gap at impasse has had its bench sitting. `maxRounds` is retired only after this is measured
to terminate (§VI) — two bounds briefly coexist rather than trading a proven one for a new one.

#### B.2.1 TRIVIALITY MUST NOT HOLD THE REPORT OPEN — and the test already exists

Impasse bounds how long ONE gap is argued. It does nothing about the other exhaustion mode
(gblock, 2026-09-08): **red trailing off into nitpicks forever.** Every new trifle is a fresh gap,
so no impasse counter ever fires, and the report never closes.

**The test for it is already in the record, and nothing refuses on it.** `schema.sql:988-991`:

```sql
(rv."verdict" = 'fail'
   AND COALESCE(b."mass", 0.0) < 35.0
   AND COALESCE(b."max_severity_mass", 0.0) <= 2.0
   AND COALESCE(f."fresh_mints", 0) = 0)         AS "divergent"
```

Red said FAIL, while the board's total mass was low, nothing on it was above **material**, and no
fresh gap was minted. That is the failure mode, stated in SQL, computed every round, surfaced
through `convergence_vs_verdict` and `record.ConvergenceVsVerdict` — **and read only by a human in
a report.** It refuses nothing. `policy-without-mechanism`, in the one place the run decides
whether it is finished.

Note `max_severity_mass <= 2.0`: `2.0` is `GRADE_MEDIUM`'s mass, and the enum's own word for
`GRADE_MEDIUM` is **"material"**, against `GRADE_TRIVIAL` = *"cosmetic; nothing downstream changes
if it is wrong"* (`record.proto:264-268`). The line gblock is asking for is already drawn, in the
schema, in the enum's own vocabulary.

**So the change is to make the existing measurement REFUSE**, in two parts:

1. **A gap below material is recorded but does not hold the gate.** Red's duty is unchanged —
   *everything real gets raised and stays raised* — and a trivial finding is still a finding on
   the record. It simply cannot be the reason a report is not certified. This is the enum being
   read for the decision it already describes.
2. **A FAIL on a convergent board is refused.** When `divergent` holds, red must raise something
   material or PASS. This is the same shape as the `verdict --as PASS` refusal that already exists
   while a motion is pending: a verdict the board contradicts is not accepted.

**Two things this must not become**, both already argued in red's own constitution:

- **Not a soft-pass.** *"An unearned FAIL is red losing, exactly as an unearned PASS is."* The
  refusal fires only when the board itself says nothing material is open — it never overrides a
  live material gap.
- **Not an incentive to inflate.** The obvious game is grading nitpicks material. Grade movement
  is already on the record as `regrade` events, red's constitution already makes grade stability a
  measured property, and `ACCEPTED_DELTA_DOCKET_THRESHOLD` (`debate.js:276`) already batch-dockets
  cumulative grade deltas for bench review. Inflation is visible; §V measures it across the gate
  run rather than assuming it will not happen.

**The `35.0` is a bare constant and must not survive the move.** `2.0` is derivable from the enum;
`35.0` is a magic number in a view with nothing saying where it came from. Roundless it is
restated relative to the board (a fraction of peak mass) or given a written derivation, because a
threshold that decides whether a run may end has to be one a reader can check.

**Roundless, `divergent` becomes per-sitting** rather than per-round, on the event axis §III.A.2
defines — the same rewrite the other round-keyed views take.

#### B.3 The chair engages the seats in an open dispute

Dispatch is per dispute, not per phase: the chair (or the judge, for a docketed dispute) engages
the lanes and lenses party to it. Seats that have nothing open are not dispatched at all, which is
where the wall-clock is won on a two-slot budget.

### §III.C Consumer census

To be pasted from these commands, one line per consumer, classified axis / display / identity:

```bash
grep -rn 'GetRound()' --include='*.go' plugins/frank-exchange-of-views/tools | grep -v _test
grep -n 'round' plugins/frank-exchange-of-views/tools/internal/record/recordsql/testdata/schema.sql
grep -rnw 'RoundOf\|RoundIn' --include='*.go' plugins/frank-exchange-of-views/tools
grep -rn 'R%d-%d\|MintGapID' --include='*.go' plugins/frank-exchange-of-views/tools
grep -rn 'r1\|-r[0-9]' plugins/frank-exchange-of-views/tools/internal/seatprobe/boards.go
ls plugins/frank-exchange-of-views/tests/simulator/testdata/*r1*.golden
```

Known already, and not yet exhaustive: `seatprobe/boards.go` hardcodes round-bearing ids **48
times**; eight `prompt-*-r1*.golden` plus `seat-roster.golden`; `docs/seat-surface-naming.md`,
`docs/why-a-seat-stops.md`, `docs/gap-pattern-memory-delivery.md`;
`releasegate/fuzz/fuzz_test.go:998`, `record/replay_test.go:84`, `record/record_test.go:168` and
`seatenv/identity_test.go:73` all call `RoundOf` and will not compile after A.1; and
`record/findinglabel.go:18`'s `roleRe` begins returning `"chair"`/`"respond"` for the newly
roundless ids where it returns `""` today.

## IV. Risk & Mitigation

| risk | mitigation |
|---|---|
| **Archived runs speak the old vocabulary.** | **MIGRATED, never dual-read** (gblock, 2026-09-08: *no archaeology and no backwards compat outside replay*). #840's `feov-record migrate` re-drives an old record through the CURRENT write path, and that is the one place the old→new mapping lives: `red-lens-r3-evidence` → `red-lens-evidence` at sitting 3, `R3-2` → its run-global id, `event.round` → the event axis. Readers speak ONE vocabulary. This plan DEPENDS on #840 leg 1 (record.db archives) and its mapping is specified in §III.A.5. |
| **The dispatch chair becomes a single point of failure** — nothing is dispatched if it errs. | It is the round loop's replacement and inherits its failure mode; the loop could also wedge. B.1 states the termination condition (no open disputes under their limits ⇒ the run ends) so a wedged chair is distinguishable from a finished run. |
| **Work-driven dispatch starves a seat** whose disputes are never selected. | The exchange count is per gap and bounded, so a gap cannot be argued forever; a seat with no open dispute is idle by design, not starved. §V measures per-seat dispatch counts against the old round-based distribution. |
| **Gap-id change breaks a consumer's saved references.** | Unavoidable, and at the major version for exactly that reason: it happens once, announced, rather than silently later. |
| **A and B land in one tag, so a defect in B cannot be shipped separately from A.** | That is the point — A alone is the half-state. The order (A then B) means A is green and swept before B starts, and §V's run gates both. |

## V. Verification Plan

Each check with what re-arms it.

```bash
# The round is gone from identity AND from the record. Re-arms: any reintroduction.
! grep -rnw 'RoundOf' --include='*.go' plugins/frank-exchange-of-views/tools | grep -v CurrentRoundOf
! grep -rn 'GetRound()' --include='*.go' plugins/frank-exchange-of-views/tools

# The named refusals and the named behaviours EXIST — anchored, because `go test` exits 0 when a
# -run pattern matches nothing, which is how the previous plan's gates passed while testing
# nothing. Re-arms: the tests themselves.
(cd plugins/frank-exchange-of-views/tools && go test ./internal/record/ \
  -run '^(TestAGapsExchangeCountIsCountedFromTheRecord|TestAPetitionSittingIsUniqueWithoutARound)$' -v \
  | grep -E '^=== RUN' | wc -l)   # must print 2, not 0

# Archived projections are identical across the change. Re-arms: any reader change.
(cd plugins/frank-exchange-of-views/tools && go test ./internal/record/ -run '^TestArchivedRunProjectsIdentically$')

# Full suites and gates.
(cd plugins/frank-exchange-of-views/tools && go vet ./... && go test ./...)
node --test plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs \
             plugins/frank-exchange-of-views/tests/simulator/prompts.test.mjs
(cd scripts && go run ./check && go run ./archaeology && go run ./rulesweep)
```

**The driveable check.** §III.B's claim is that work-driven dispatch does less work for the same
verdict. The measurement is a run's total sittings and wall-clock against the round-based baseline
in #753's table (lens phase = sum ÷ ~1.6, blue edits 11 → 9 → 7 → 3). A roundless run that
dispatches the same number of seats has not paid for itself.

**SIZE THE RUN BEFORE IT STARTS, AND SAY WHICH TARGETS ARE RATES.** This run carries three
verdicts — #792's B4, this plan's cost comparison, and `derived-seat-identity.md` §V.1's cache
measurement — and that is exactly why its size is a correctness question rather than a convenience
one. **A small run passes a ceiling by being small and produces a favourable cost comparison from
the same smallness**, and the two would read as independent corroboration while being one
artifact. (Raised by a peer session, 2026-09-08, against their own earlier caveat that B4's
`≤75,000` log-char target and its `161/24/13/9/2` report census were both taken on a FULL run.)

So, before the run rather than after:

- fix the topic and the corpus size against the #753 baseline run, so the cost comparison has a
  denominator it shares;
- restate every ceiling target as a RATE (per 1,000 words of report prose, per sitting) or record
  it explicitly as **not measured** — a ceiling passed by a short run is not evidence;
- state which of the three verdicts a given number serves, because one artifact answering three
  questions must not be read as three artifacts agreeing.

## VI. Deliberately not in this plan

- **The ATTESTED half of identity** — `agent_type` per seat, the cast enum, the register-time
  conflict refusal. `plans/derived-seat-identity.md` owns those; this plan owns only the round.
- **`--wait` on a seat's work list** (#753 mechanism 1). Rejected there by arithmetic: a blocked
  agent holds an `agent()` slot and the cap is ~2.
- **Retiring `maxRounds` as a run parameter** before B.2's per-item limits are measured. Two
  bounds briefly coexist; the global one is removed after the per-item one is shown to terminate.
