# Roundless — the global clock leaves the record

> STATUS 2026-09-08: proposed, **required pre-v2.0.0, landing whole, gated by a real run**
> (gblock). Implements #753 entire. Split out of `plans/derived-seat-identity.md`, which owns the
> ATTESTED half of identity and hands the round here. Depends on #840 (`feov-record migrate`):
> archived runs are migrated, never dual-read — *no archaeology, no backwards compat outside
> replay* (gblock, 2026-09-08).
>
> Revision 4. Three `/plan-audit` FAILs preceded it, each on the same root — a global property of
> `event.round` asserted without enumerating its readers. §III.C is the census, run and pasted.

## I. Summary & Goals

**The round is a global clock over work that has per-item lifecycles** (#753). It does three jobs,
named separately because they end differently:

| job | where | ends how |
|---|---|---|
| part of a seat's IDENTITY | `red-lens-r3-evidence`, `RoundOf`'s `-r(\d+)` | deleted (§III.A.1) |
| the record's TIME AXIS and BUCKET | `event.round`, `TelemetryLine.round`, ~12 round-named columns, `R<round>-N` | replaced (§III.A.2–3) |
| the ORCHESTRATION unit | `debate.js`'s round loop, `maxRounds` | replaced by a dispatch chair and per-gap limits (§III.B) |

### Success criteria

1. `grep -rnw 'RoundOf\|RoundIn' --include='*.go' plugins/frank-exchange-of-views/tools` → **0**.
2. `grep -n '"round"' …/recordsql/testdata/schema.sql` → **0**; `record.proto` has no field named
   `round` on any message.
3. No id the engine dispatches matches `-r\d+`; no gap id matches `^R\d+-\d+$`.
4. `run-archive/2026-09-02_quadratic-formula.tar.gz` migrated through #840 projects with the same
   gap count, the same closure count and the same final verdict as its pre-migration replay.
5. The gate run (§V) reaches a verdict with fewer total sittings than #753's baseline run, or the
   plan records that it did not and why.

### What the round actually does in the record — from the census, not from reading a slice

Three prior drafts asserted a global property of `event.round` ("only an axis") and were wrong
each time. §III.C enumerates every reader. What they show:

- **Axis** (ordering/comparison across seats): `closed_round`, `minted_round <= r.round`,
  `viewjson.go:1305 e.Round >= since`, `inquiry.go:311`, `motionview.go` sort, `docs.go` max.
- **Bucket** (grouping/indexing): `chart.go:47 minted[g.Round]++` with `if last < 2 { return "" }`;
  `viewjson.go:1162 byRound[]`; `spotcheck.go:105 mergeSat[round]`; `cost.go:135` key
  `%02d|seat|tier`; `changes.go:58` filter; `assemble.go:1369 "### Round %d"` headings.
- **Identity** (pair lookup / recomposition): `dashboard/model.go:552,560 (seat, round)`;
  `model.go:207 seat + "-r" + round`; `mint.go:51 MintGapID(run, s.Round)`.
- **Per-round projection**: `TelemetryLine.round` (`record.proto:532`), written by
  `view.Telemetry`, read by `scorecard.go`, `cost.go:419`, `dashboard/render.go:104`,
  `model.go:615`, `convergence.go`.
- **Display**: `"r%d"` in `proofs.go`, `spotcheck.go`, `assemble.go:702`, `motions.go:48`,
  `inquiry.go:176`, `available.go:70`.

Each class gets a named replacement in §III.A.2. That is the whole difference from the drafts
that failed.

### Why it lands whole, before the tag

§III.A alone ships a record vocabulary with no round in it, still produced by a loop that counts
rounds — a half-state that reads as done. gblock ruled the whole change lands and a real run gates
it. A then B in order, because B's chair reads a board A defines.

## II. Technical Context

Verified against `build/blue-lane-configs` tip, 2026-09-08, each line opened.

- `envelope()` stamps `ev.Round` on every event (`record/record.go:320`); `insertNumbered` takes
  it (`:271, :515`); `store.go:451` writes it.
- `RoundIn` → `RoundOf` → `roundRe` (`record/round.go:8`); special-cases at `round.go:23,27`;
  `RoundIn`'s record-derived branch at `:62`.
- `MintGapID(run, round)` (`record/replay.go:413`) is `count(*) FROM mint … WHERE e.round = ?`,
  formatted `R%d-%d` at `:420`. Shape readers: `flags/shapes.go:40 gapIDShape = ^R\d+-\d+$`
  (refuses `--id`/`--supersedes`), `report/md.go:57 idToken`, `idnamespace_test.go:52` (reads the
  namespace FROM `flags.GapID().Shape()`).
- `convergence_vs_verdict` (`schema.sql:983-1017`): `divergent` = `verdict='fail' AND mass < 35.0
  AND max_severity_mass <= 2.0 AND fresh_mints = 0`. Its mass subquery (`:998-1008`) joins
  `gs.value = g.severity` — the **mint** grade — while the `gap` view exposes `current_severity`
  folded over `regrade` (`:845`). `fresh_mints` (`:1010-1016`) counts gaps with
  `supersedes_count = 0`, at ANY grade.
- `mint.location` (`:392`) is the gap's own anchor. `blue_edit.answers` (`:599`) names the gap an
  edit answers. `regrade.gap_id` (`:464`), `close.gap_id` (`:439`). All direct joins.
- `debate.js:740` `sitting.halt`; `:1074-1076` the carried-persistence comment, `:1077` its `if`;
  `:1071` the re-raise docket trigger; `:511-513` `deadlock` on the bench envelope; `:8` the run's
  stated end condition.
- Prefix letters in use: `R` gap, `M` motion, `Q` inquiry, `P` proof, `L` archived finding.
  **`G` is free.**
- `record.proto` `(means)` strings naming rounds: `:239 POSITION`, `:246 VERDICT`, `:249
  BASE_INGEST`, `:289 RUN_OUTCOME_CEILING`, `:322 DISPOSITION_CARRIED`, `:1494 RULING_BINDS_BLUE`.
- `seatclass.go:32-38` recovers the round from dispatch-prompt TEXT (`Red audit, round (\d+)`).
- #753 open problem 2: *"#709 makes this tractable — the report is the record, every change an
  appended event, so 'the report as of event N' is reconstructible and a lens can audit a pinned
  version."* Open problem 1: the deadlock judgement is *"explicitly round-scoped … becomes per-gap
  stalling, which is a redefinition rather than a port."*

## III. Proposed Changes (the spec)

### Tree

```
plugins/frank-exchange-of-views/
  skills/research-protocol/scripts/debate.js         [MODIFY] round loop → dispatch chair; ids lose -r<N>
  agents/red-dispatch.md                             [NEW]    the dispatch chair's configuration
  tools/internal/record/recordpb/record.proto        [MODIFY] Event.round, TelemetryLine.round deleted;
                                                              six (means) strings; RoundVerdict → Verdict
  tools/internal/record/recordsql/testdata/schema.sql[MODIFY] round column/index gone; sitting window;
                                                              ~12 *_round columns → *_seq; views re-axed
  tools/internal/record/round.go                     [DELETE]
  tools/internal/record/sitting.go                   [MODIFY] SittingOf gains the ordinal (§III.A.0)
  tools/internal/record/replay.go                    [MODIFY] MintGapID(run) → G<n>
  tools/internal/record/impasse.go                   [NEW]    §III.B.2's two counters
  tools/internal/record/convergence.go               [MODIFY] corrected predicate; refuses
  tools/internal/record/verdict.go                   [MODIFY] the refusal arms
  tools/internal/record/migrate/registry.go          [MODIFY] the id translations §III.A.5 lists
  tools/internal/flags/shapes.go                     [MODIFY] gapIDShape ^G\d+$
  tools/internal/seatenv/identity.go                 [MODIFY] ResolveSeat loses inferRound
  tools/internal/cli/seat/seat.go                    [MODIFY] :186,:265 Identity loses Round
  tools/internal/cli/merge/mint.go                   [MODIFY] :51
  tools/internal/cli/dispatch/                       [NEW]    the turn-report verb (§III.B.1)
  tools/internal/consistency/consistency.go          [MODIFY] :473 cross-check deleted (#676)
  tools/internal/seatclass/seatclass.go              [MODIFY] :32-38 prompt regexes lose the round
  tools/internal/report/{chart,md,docs,assemble,proofs,motions}.go   [MODIFY] per §III.C
  tools/internal/dashboard/{model,render,dashboard}.go                [MODIFY] per §III.C
  tools/internal/{cost,scorecard,view,graph,capture,verify}/…        [MODIFY] per §III.C
  tools/internal/seatprobe/boards.go                 [MODIFY] 58 round-bearing ids
  tests/simulator/testdata/*.golden                  [MODIFY] 8 files carrying -r<N>
  docs/{seat-surface-naming,why-a-seat-stops,gap-pattern-memory-delivery}.md  [MODIFY]
```

### §III.A.0 The sitting ordinal — specified once, here

Both plans depend on it and neither specified it. **`events.sitting`** is the count of that seat's
`register` events at or before the row, computed as a window over `events` and exposed as a
column of the `events` view — never stamped by a seat. Register-inclusive by construction, per-seat
so no sibling can move it, 1 on a seat's first register. It applies to EVERY seat kind uniformly,
petition sittings included — which resolves the contradiction the audit found in A.1: a
`judge-petition-red-chair` sitting is distinguished from the previous one exactly as `red-chair`'s
third sitting is from its second, by the ordinal, and nothing is composed back into the name.

### §III.A.1 Seat ids lose `-r<N>` [MODIFY/DELETE]

`red-lens-r3-evidence` → `red-lens-evidence`; `red-chair-r2` → `red-chair`; `blue-respond-r1` →
`blue-respond`; `judge-r2` → `judge`; `judge-petition-red-chair-r1` → `judge-petition-red-chair`.
`blue-lane-<N>` keeps its index: a lane, not a clock. `events.key` already scopes the idempotency
ordinal per seat (`record.proto:444-450`), so one id over many sittings collides with nothing.

Deleted: `RoundOf`, `RoundIn`, `synthesisSeats`, `terminalSeats`, `laneRe`, `roundRe`,
`consistency.go:473` (closes #676), `seatenv.Seat.HasRound`, `Identity.Round`. `ResolveSeat` loses
its `inferRound` parameter. `findinglabel.go:18 roleRe` is re-anchored so it does not start
returning `"chair"`/`"respond"` for the newly roundless ids.

### §III.A.2 `event.round` and `TelemetryLine.round` leave the schema [DELETE/MODIFY]

Replacement **per reader class** from §III.C:

| class | replacement |
|---|---|
| axis | `events."id"` (the run's sequence). `closed_round` → `closed_seq`; `minted_round <= r.round` → `minted_seq <= r.seq`; every `*_round` column → `*_seq`. |
| bucket by round | **bucket by the chair's sitting.** A red-chair sitting is the natural unit the round was standing in for: `chart.go`'s `minted[]` indexes by chair sitting; `spotcheck.go`'s `mergeSat[]` keys on it; `cost.go`'s key is `%02d|seat|tier` with the chair sitting; `assemble.go:1369` heads sections by it; `changes.go:58` becomes "since this seat's previous sitting" (per-seat, from `events.sitting`). |
| identity | the sitting ordinal. `model.go:552,560` look up `(seat, sitting)`; `model.go:207` labels `seat` + ` #` + sitting; `mint.go:51` → `MintGapID(run)`. |
| per-round projection | **per chair sitting.** `TelemetryLine.round` → `TelemetryLine.sitting`; `convergence_vs_verdict` keys on the chair sitting whose `verdict` it is; `round_verdict` table → `verdict`. |
| display | `"r%d"` → `"#%d"` (sitting) where the seat is named beside it; dropped where it is not. |

`chart.go:41`'s `if last < 2 { return "" }` — "a single round is not a series" — becomes "fewer
than two chair sittings", which is the same suppression with the same meaning.

### §III.A.3 Gap ids become `G<n>` [MODIFY]

`MintGapID(run)`: `SELECT count(*) FROM "mint"`, formatted `G%d`. Run-global, monotonic, no round.
`G` is the free prefix letter; the two schemes cannot collide because `^G\d+$` and `^R\d+-\d+$`
share no string. Shape readers updated: `flags/shapes.go:40`, `report/md.go:57`,
`idnamespace_test.go:52-53` (reads the shape from flags, so it follows). Archived `R<r>-<n>` ids
are translated at migration (§III.A.5) and never read by a live reader.

### §III.A.4 The dispatch chair joins the cast [NEW]

`agents/red-dispatch.md`, agent type `frank-exchange-of-views:red-dispatch`, seat id
`red-dispatch`, role `dispatch` (a new role; `chairOfRole` maps it to `red`). Named here because a
new seat is vocabulary. Its behaviour is §III.B.1.

### §III.A.5 Archived runs are MIGRATED [MODIFY, in #840's registry]

Readers speak one vocabulary. `feov-record migrate` re-drives an old record through the current
write path, and the translation lives in its registry, once:

| old | new |
|---|---|
| `red-lens-r3-evidence` | `red-lens-evidence`, sitting 3 |
| `red-chair-r2`, `red-merge-r2` | `red-chair`, sitting 2 |
| `blue-respond-r1`, `judge-r2`, `judge-petition-<x>-rN` | roundless id, sitting N |
| `R3-2` | `G<position in the run's mint order>` |
| `event.round` | dropped; `events."id"` is the axis |
| `L2-F1` | `<area>-F1` |
| `role: merge` | unchanged |

**This reaches back into merged work**: #791's `L\d+` arms in `FindingLabelAlt`/`roleRe`, #831's
`red-merge` acceptance in `roleSeats`, `seatclass`, the dashboard and seatprobe reads. All are
dual-shape readers; all retire once migration covers them. The decisive argument, from the audit:
`store.go:339 ensureSchema` returns early on `hasEvents`, so an archived database keeps its OWN
views verbatim — a Go reader accepting both shapes could never have reached that layer anyway.

Dependency edge: #840 leg 1 covers the one `record.db` archive; leg 2 the six JSONL archives
`store.go:82-92` refuses. If leg 2 misses the tag, those six are unreadable by v2 — a decision for
#792, not a surprise.

### §III.B.1 The dispatch chair, and the turn-report verb [NEW]

**A seat can read the record; the workflow cannot** (`debate.js` imports no `fs`; every seat is
told `${binDir}/feov-record` at `:226-232`). So dispatch is a seat: `red-dispatch` is dispatched,
reads the board, and returns who to engage. The workflow obeys it and does nothing else.

**It does not return a composed envelope.** The audit was right that a chair whose word the
workflow simply believes has moved the self-assertion channel, not closed it. So the chair acts
through a VERB — #753's mechanism 3, the piece that plan called *"the highest-value piece"* and an
earlier draft dropped: `feov-record dispatch next` computes readiness FROM THE BOARD (open gaps
below their limits, their parties, their pins) and records the decision as an event. The chair
relays the verb's JSON into its envelope, the same move `claim_count` made. The workflow dispatches
what the record says, and:

- a seat outside the run's cast is refused by the verb, not by the workflow;
- an empty list is the termination signal (§III.B.2), not an error;
- an error from the chair halts the run with the error on the record, exactly as `sitting.halt`
  does today (`debate.js:740`).

**Cost against the measured cap of ~2**: one serial sitting between fan-outs. It never holds a
slot while others work.

**The moving target** (#753 open problem 2): the verb PINS each dispatch to `events."id"` at
dispatch time. A lens audits the report as of that pin (#709: the report is the record, every
change an appended event, so the pinned text is reconstructible) and anchors its findings to it.
A quote not found in the pinned text is REJECTED as today; a quote found there but since edited is
a finding against text blue has moved, which the chair reconciles rather than the lens.

### §III.B.2 IMPASSE — the limit before the bench [NEW]

**The exchange unit, defined from the event vocabulary.** An act on gap G by red is a `mint`,
`regrade`, `close`, `closing` or `spot_check` naming G, or a `mint` whose `supersedes` names G.
An act on G by blue is a `blue_edit` whose `answers` is G, a `closing` on G, a `motion` on G, or a
`retire` of G's claim. **One exchange = one red act followed by one blue act on G**, in `events`
order.

**Movement in an exchange** = any of: `regrade` on G; `blue_edit` with `answers = G`; a `mint`
superseding G; `close` on G. All four are direct joins on the gap id — `blue_edit.answers` names
the gap, so no anchor path is needed and a gap minted without `found_by` is covered.

**Two counters, both from the record, in `impasse.go`:**

- `exchanges(G)` — total exchanges on G. **Monotone. Never resets.**
- `stalled(G)` — consecutive exchanges on G with no movement. Resets on movement.

**Impasse holds when `stalled(G) ≥ K` OR `exchanges(G) ≥ K_max`.** The second bound is what
oscillation cannot reset: a regrade per exchange keeps `stalled` at 0 forever, and `exchanges`
still climbs. Defaults `K = 2, K_max = 6`, as run parameters beside `lanes`; both recorded at setup
so the limits are facts on the record, not constants in JavaScript.

At impasse G goes to the bench's docket. This subsumes the current trigger (`debate.js:1071`
re-raise count, no movement test) and the carried exception (`:1074-1077`, movement tested on grade
only) — one rule, four restrictions removed.

**Termination** (#753 open problem 1): the run ends when no gap is open below impasse and every
gap at impasse has had its bench sitting. `feov-record dispatch next` returns empty exactly then.
**Deadlock is redefined, not ported**: today it is round-scoped (*"no gap carried AND no new gaps
this round"*); it becomes the bench's judgement that a docketed gap cannot be resolved, per gap,
and the run's end is the termination condition above rather than a deadlock verdict.

### §III.B.2.1 Triviality must not hold the report open — the refusal, corrected

`divergent` exists (`schema.sql:988-991`) and refuses nothing. The audit found its predicate wrong
on three counts for this purpose, so the refusal uses a **corrected** predicate rather than the
view as it stands:

| as written | defect | corrected |
|---|---|---|
| `fresh_mints = 0` | a run minting one fresh trifle per sitting is never divergent — the exact trajectory this exists to stop | `fresh MATERIAL mints = 0`: fresh gaps whose `current_severity` mass ≥ 2.0 |
| `max_severity_mass <= 2.0` | inclusive of `GRADE_MEDIUM` = material; would force PASS over a live material gap | `< 2.0`, strict |
| joins `gs.value = g.severity` | the MINT grade; a `regrade` cannot move it | joins `current_severity` (`:845`), the fold over `regrade` |
| `mass < 35.0` | a magic number with no derivation | restated as a fraction of the run's peak board mass, recorded at setup |

**Two refusals, in `verdict.go`:** a FAIL is refused while the corrected predicate holds (red must
raise something material or PASS); and a gap below material is recorded but does not hold the
gate. Neither overrides a live material gap — that is what `< 2.0` on `current_severity` buys.

**Inflation** is the game; `regrade` events are on the record and
`ACCEPTED_DELTA_DOCKET_THRESHOLD` (`debate.js:276`) already dockets cumulative deltas. §V measures
regrade frequency on the gate run rather than assuming it away.

### §III.B.3 Dispatch is per dispute

The verb engages the seats party to an open dispute below its limits — the lens that minted, the
blue seat answering, the bench if docketed. Seats with nothing open are not dispatched. That is
where wall-clock is won on a two-slot budget.

### §III.C Consumer census — run 2026-09-08, results pasted

```
$ grep -rn '\.Round\b' --include='*.go' . | grep -v _test.go      → 80 lines, 29 files
  false positives (math.Round / time.Round / round2): 14
  writers:   record.go:271,320,515
  identity:  seatenv/identity.go:107 · cli/seat/seat.go:186,265 · cli/merge/mint.go:51 ·
             dashboard/model.go:207-208,552,560
  axis:      view/view.go:441,464,470 · report/docs.go:307-308 · record/viewjson.go:1305 ·
             record/inquiry.go:311 · record/motionview.go:94-95
  bucket:    report/chart.go:31-32,47-48 · cost/cost.go:135,242
  telemetry: view/view.go:876 · cost/cost.go:121,202-222,419 · record/convergence.go:43 ·
             dashboard/render.go:104-105 · dashboard/model.go:615
  display:   report/proofs.go:142-148 · report/motions.go:48 · report/assemble.go:964 ·
             record/spotcheck.go:152-157 · record/available.go:70 · record/gapedit.go:57 ·
             record/motion.go:350,423 · record/motionview.go:79 · scorecard.go:232 ·
             verify.go:111 · dashboard/render.go:282-283,369 · capture.go:214
  generated: recordpb/record.pb.go:1728-1729,2315-2316

$ grep -rn 'GetRound()' --include='*.go' . | grep -v _test.go      → 45 lines, 15 files
  writer:    recordsql/store.go:451
  bucket:    viewjson.go:1162,1187,1252 · spotcheck.go:105,114 · changes.go:58 ·
             assemble.go:1369 · inquiry.go:350 · capture.go:215
  axis:      viewjson.go:508,527,879,902-903 · docs.go:315-316 · inquiry.go:295 ·
             consistency.go:123,152,195 · view/view.go:600 · assemble.go:1129
  identity:  consistency.go:473,475 (deleted)
  carry:     viewjson.go:994,1070 · evidenceview.go:254,283,320,335 · citationid.go:312 ·
             assemble.go:904 · graph.go:121 · inquiry.go:166,218 · motion.go:325,350,453
  display:   changes.go:185 · assemble.go:702 · inquiry.go:176
  telemetry: view/view.go:182
  generated: record.pb.go:1727,2314

$ grep -n '"round"\|_round"' …/schema.sql                            → 30 lines
  column+index: :5, :16 · table: round_verdict :269,:668 ·
  *_round columns: registered :684, minted :820, merge_closed :824, bench_closed :827,
  closed :828, ruled :1046,:1078,:1104, appealed :1049, filed :1069, proposed :1098,
  status :1101 · convergence_vs_verdict :983-1017

$ record.proto: Event.round (:454) · TelemetryLine.round (:532) · RoundVerdict (:1638) ·
  (means) strings :239,:246,:249,:289,:322,:1494

$ gap-id shape readers: flags/shapes.go:39-40,126 · report/md.go:57 · replay.go:411-420 ·
  cli/merge/mint.go:51,263 · capture.go:904 (comment) · findinglabel.go:34 (comment)
$ prompt-text round: seatclass.go:32-38 (5 regexes) · capture.go:737 (comment)
$ seatprobe/boards.go: 58 round-bearing ids
$ goldens carrying -r<N>: prompt-{blue-respond,red-chair,red-lens-dark-side,red-lens-evidence,
  red-lens-logic}-r1, prompt-red-lens-{evidence-r2-consolidated,logic-r2}, seat-roster (8)
$ maxRounds: 46 sites in 12 files — dashboard/model.go 10, debate.js 9, verdict.go 6,
  setup/run.go 4, debatejs 3, render.go 3, cli/setup.go 3, cli/dashboard.go 3, flags 2,
  seatprobe/production.go 1, roster.go 1, capture/lanecoverage.go 1
$ RoundOf/RoundIn: 16 + 86 sites (2 production for RoundIn: seat.go:264,519)
$ docs naming rounds: seat-surface-naming, why-a-seat-stops, gap-pattern-memory-delivery
$ plans/derived-seat-identity.md:190 still says "§III.B is post-tag" — corrected with this revision
```

## IV. Risk & Mitigation

| risk | mitigation |
|---|---|
| A reader still assumes `(seat, round)` is unique. | `events.sitting` is the discriminator on every row; §III.C is the reader list; each is rewritten per its class, not swept by regex. |
| Archived runs. | Migrated (#840), never dual-read. §V criterion 4 diffs projections across the migration. |
| The dispatch verb returns work for a seat that no longer exists, or nothing. | Cast membership is refused in the verb; empty is termination, recorded as such; an error halts on the record. The chair cannot lie about readiness because it does not compute it. |
| Oscillation defeats impasse. | `exchanges(G)` is monotone and bounds it regardless of `stalled(G)`. |
| The convergence refusal soft-passes. | Strict `< 2.0` on `current_severity`; material fresh mints block. §V drives both arms. |
| Grade inflation to escape the refusal. | Visible as `regrade` events; docketed by `ACCEPTED_DELTA_DOCKET_THRESHOLD`; measured on the gate run. |
| The gate run is sized to pass. | §V: topic and corpus fixed against #753's baseline; ceilings restated as rates or recorded not-measured; each number attributed to the verdict it serves. |

## V. Verification Plan

```bash
T=plugins/frank-exchange-of-views/tools
# 1. The round is gone from identity and the record. Re-arms: any reintroduction.
test "$(grep -rnw 'RoundOf\|RoundIn' --include='*.go' $T | wc -l)" -eq 0
test "$(grep -n '"round"' $T/internal/record/recordsql/testdata/schema.sql | wc -l)" -eq 0
test "$(grep -n 'int32 round' $T/internal/record/recordpb/record.proto | wc -l)" -eq 0

# 2. The named tests EXIST and pass — each anchored, and counted, because `go test -run` exits 0
#    on no match. Re-arms: the tests themselves.
for t in TestAGapsExchangesAndStallsAreCountedFromTheRecord \
         TestImpasseFiresOnStallAndOnTheMonotoneBound \
         TestAFailOnAConvergentBoardIsRefused \
         TestABelowMaterialGapDoesNotHoldTheGate \
         TestAMaterialFreshMintBlocksTheConvergenceRefusal \
         TestTheDispatchVerbRefusesASeatOutsideTheCast \
         TestAnEmptyDispatchIsTermination \
         TestGapIdsAreRunGlobalAndNeverRoundShaped \
         TestTheSittingOrdinalIsRegisterInclusiveAndPerSeat; do
  n=$(cd $T && go test ./internal/record/ ./internal/cli/... -run "^$t\$" -v 2>&1 | grep -c "^--- PASS: $t")
  test "$n" -eq 1 || { echo "MISSING OR FAILING: $t"; exit 1; }
done

# 3. Migration preserves projections. Re-arms: #840's registry, any reader.
(cd $T && go test ./internal/record/migrate/ -run '^TestQuadraticFormulaProjectsIdenticallyAfterMigration$' -v 2>&1 | grep -c '^--- PASS') | xargs test 1 -eq

# 4. Full suites and gates.
(cd $T && go vet ./... && go test ./...)
node --test plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs \
             plugins/frank-exchange-of-views/tests/simulator/prompts.test.mjs
(cd scripts && go run ./check && go run ./archaeology && go run ./rulesweep)
```

**The gate run.** One run, three verdicts — #792's B4, this plan's cost comparison, and
`derived-seat-identity.md` §V.1's cache measurement. Sized BEFORE it starts: topic and corpus fixed
against #753's baseline (`2026-08-23_research-loop-counterparts`) so the comparison shares a
denominator; every ceiling restated as a rate per 1,000 words of report prose or per sitting, or
recorded **not measured**; each number attributed to the verdict it serves. Measured on it:
total sittings and wall-clock against #753's table; `regrade` frequency (inflation); how many
gaps reached impasse by `stalled` versus by `exchanges` (whether the monotone bound ever fires);
and whether any lens finding was rejected for a quote the pin should have preserved.

## VI. Deliberately not in this plan

- **`--wait`** (#753 mechanism 1): a blocked agent holds a slot; the cap is ~2.
- **Retiring `maxRounds`** before impasse is measured to terminate on the gate run. Two bounds
  coexist for one run; the global one goes after.
- **The attested half of identity** — `plans/derived-seat-identity.md`.
- **Per-lane METHOD as a carrier** (#497).
