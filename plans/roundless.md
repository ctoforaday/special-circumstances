# Roundless — the global clock leaves the record

> STATUS 2026-09-08: proposed, **required pre-v2.0.0, landing whole, gated by a real run**
> (gblock). Implements #753 entire. Split out of `plans/derived-seat-identity.md`, which owns the
> ATTESTED half of identity and hands the round here. Depends on #840 (`feov-record migrate`):
> archived runs are migrated, never dual-read — *no archaeology, no backwards compat outside
> replay* (gblock, 2026-09-08).
>
> Revision 8. Seven `/plan-audit` FAILs preceded it; round 5 reopened NO fork — four spec
> sentences forced by rulings already made, seven tree facts, and one self-inflicted wound
> (revision 7's edit deleted §III.B.2.1's body; restored here from a812e857). **Implementation
> starts on this revision** unless round 6 finds a new fork.
> Revision 7. Six FAILs preceded it; round 4 left ONE design gap — the run had no
> fixed point once `maxRounds` went — and gblock ruled it: **a MINT BUDGET per lens**, and a
> sub-material gap does not make blue ready. §III.B.2.2. Eight tree facts closed alongside.
> Revision 6. Five FAILs preceded it; round 3 classified its twelve gaps as five
> design and seven tree facts, and this revision closes the five: readiness for the seats that
> are not parties to a gap, an exchange unit defined over SITTINGS so a silent party cannot stall
> forever, the verb→role table after the mint move, the PASS refusal made material-aware, and the
> four-outcome vocabulary. Revision 5. Revision 4 ran the census; this one places the
> chair in the cycle (gblock, 2026-09-08: **lenses mint; the chair no longer does it for them** —
> #846 is the shipped half-state that ruling corrects), which is what revision 4's worst gap was.

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
  tools/internal/record/recordpb/record.proto        [MODIFY] Event.round, TelemetryLine.round deleted;
                                                              six (means) strings; RoundVerdict → Gate
                                                              (NOT `Verdict`: `enum Verdict` exists at :277)
  tools/internal/record/recordsql/{schema,views}.go  [MODIFY] the SOURCE; testdata/schema.sql is its
                                                              golden and is regenerated, never edited
  tools/internal/record/round.go                     [DELETE]
  tools/internal/record/roster.go                    [MODIFY] :58-64 shapes lose -r\d+; :111 lensAreaRe
  tools/internal/record/queries.go                   [MODIFY] :101 RoundsWithRevision → per epoch
  tools/internal/record/record.go                    [MODIFY] :1032 refusals beside requirePassClosesAllGaps
  tools/internal/hookgate/hookgate.go                [MODIFY] :114 id read
  tools/internal/record/sitting.go                   [MODIFY] SittingOf gains the ordinal (§III.A.0)
  tools/internal/record/replay.go                    [MODIFY] MintGapID(run) → G<n>
  tools/internal/record/impasse.go                   [NEW]    §III.B.2's two counters
  tools/internal/record/convergence.go               [MODIFY] corrected predicate; refuses
  tools/internal/record/verdict.go                   [MODIFY] :66,:75 CEILING from impasse, not max(round)
  tools/internal/record/migrate/{replay,source}.go   [MODIFY] a STATEFUL reference remap (§III.A.5); origin/main only
  tools/internal/capture/capture.go                  [MODIFY] :743 record-parity audit
  tools/internal/flags/shapes.go                     [MODIFY] gapIDShape ^G\d+$
  tools/internal/seatenv/identity.go                 [MODIFY] ResolveSeat loses inferRound
  tools/internal/cli/seat/seat.go                    [MODIFY] :186,:265 Identity loses Round
  tools/internal/cli/merge/mint.go                   [MODIFY] :51
  tools/internal/cli/merge/dispatch.go               [NEW]    the turn-report verb; role stays merge, package stays merge
  tools/internal/record/recordpb/record.proto        [MODIFY] + EVENT_TYPE_DISPATCH, EVENT_TYPE_CAST, Dispatch, Party, Cast
  tools/internal/record/recordsql/schema.go          [MODIFY] + dispatch, dispatch_party, dispatch_party_gap, cast_seat
  tools/internal/setup/run.go                        [MODIFY] :315,:318,:461 writes CAST + K/K_max/M; loses MaxRounds
  tools/internal/cli/setup.go                        [MODIFY] :72 --max-rounds → --mint-budget/--impasse/--impasse-max
  commands/research.md                               [MODIFY] :3,:6,:20 --max-rounds, --smoke's maxRounds: 2, "every round"
  README.md                                          [MODIFY] :9
  tools/internal/seatprobe/{build,attempted}.go      [MODIFY] :46-49 -r ids; :116 mints and :147 closes AS a lens, not the chair; attempted.go:122
  tools/internal/cli/lens/{mint,regrade,close,nearmatch,class}.go [MOVE from merge/] originator owns G
  tools/internal/record/refs.go                      [MODIFY] :315 material-aware; + every-lens-sat-against-head
  tools/internal/record/recordsql/views.go           [NEW]    events_w with the two windows (§III.A.0)
  tools/releasegate/fuzz/termination_test.go         [MODIFY] the round-loop termination model → the triple
  tools/releasegate/fuzz/{fuzz,blueestoppel,delivery}_test.go [MODIFY] RoundOf/MaxRounds sites
  tools/internal/debatejs/debatejs.go                [MODIFY] :78,101,247 MaxRounds
  tools/internal/seatprobe/boards.go                 [MODIFY] a lens board demands mint; :763-771 NoSituation
  tools/internal/record/roles.go                     [MODIFY] SampleSeatOf composes prefix+"r1"
  tests/simulator/{debate,prompts}.test.mjs          [MODIFY] 117 + 8 round references
  tools/internal/cli/seat/verbs.go                   [MODIFY] :222-223 and ~84 help lines naming rounds
  tools/internal/consistency/consistency.go          [MODIFY] :473 cross-check deleted (#676)
  tools/internal/seatclass/seatclass.go              [MODIFY] :32-38 prompt regexes lose the round
  tools/internal/report/{chart,md,docs,assemble,proofs,motions}.go   [MODIFY] per §III.C
  tools/internal/dashboard/{model,render,dashboard}.go                [MODIFY] per §III.C
  tools/internal/{cost,scorecard,view,graph,capture,verify}/…        [MODIFY] per §III.C
  tools/internal/seatprobe/boards.go                 [MODIFY] 58 round-bearing ids
  tests/simulator/testdata/*.golden                  [MODIFY] 8 files carrying -r<N>
  agents/*.md (11), skills/{research-protocol,adversarial-audit}/SKILL.md      [MODIFY] round vocabulary
  docs/{seat-surface-naming,why-a-seat-stops,gap-pattern-memory-delivery,record-flow,
        seat-command-triggers,seat-duty-channel,propagation-and-anchoring,finding-markers}.md [MODIFY]
```

### §III.A.0 Two windows — the sitting ordinal and the epoch — specified once, here

Both plans depend on it and neither specified it. **`events` is a STRICT table**
(`recordsql/schema.go:42`), not a view, so the windows are a [NEW] view `events_w` that every view
currently joining `e."round"` joins instead (`gap` :820-828, `gap_edit` :789,
`convergence_vs_verdict` :983+):

```sql
CREATE VIEW "events_w" AS
SELECT e.*,
  count(*) FILTER (WHERE e."type" = 'register')
    OVER (PARTITION BY e."seat_id" ORDER BY e."id")                        AS "sitting",
  count(*) FILTER (WHERE e."type" = 'register' AND e."seat_id" = 'red-chair')
    OVER (ORDER BY e."id")                                                 AS "epoch"
FROM "events" e;
```

(Both verified against modernc sqlite by the round-4 audit: `sitting` is 0 on `harness` rows,
`epoch` is 0 before the first chair register. Revision 6's form referenced an unbound `e2` and did
not execute.) **`sitting`** is the count of that seat's `register` events at or before the row — never stamped
by a seat. It applies to EVERY seat kind uniformly, petition sittings included, which resolves
the contradiction the audit found in A.1: nothing is composed back into the name.
`SITTING_OPEN`/`SITTING_CLOSE` rows are written under seat `harness` (`sittingwrite/write.go:38`)
and carry sitting 0 — the harness's observation, not a seat's act.

**`events.epoch`** is the second window and it is GLOBAL: the count of `red-chair` register events
at or before the row, whoever authored it. It is what the bucket readers used the round for —
"which dispatch cycle" — and it is defined for a bench close, a telemetry row or a lane's draft
because it asks about the chair's registers, not the row's seat's. Direction: at-or-before; the
first chair sitting opens epoch 1 and everything before it is epoch 0.

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
| bucket by round | **`events.epoch`** (§III.A.0). `chart.go`'s `minted[]` indexes by epoch; `spotcheck.go`'s `mergeSat[]` keys on it; `cost.go:135`'s key is `%02d|seat|tier` with the epoch; `assemble.go:1369` heads sections by it; `changes.go:58` becomes "since this seat's previous sitting" (`events.sitting`). The SQL-in-Go readers the audit found: `verdict.go:66 max("round")` → CEILING derives from impasse (§III.B.2); `queries.go:101 RoundsWithRevision` → epochs with a revision, `capture.go:743` compares against that; `CurrentRoundOf` (`inquiry.go:273,304,344`, `view.go:435`) → `CurrentEpoch`. |
| identity | the sitting ordinal. `model.go:552,560` look up `(seat, sitting)` **sourced from the record**, not `seatclass.ClassifySeat` of the transcript head: the dashboard joins the sitting's `agent_id` (`SITTING_OPEN`, #735) to the register event that names the seat and reads `events.sitting`. `model.go:207,210` label `seat #sitting`. `mint.go:51` → `MintGapID(run)`, now a LENS verb (§III.B.3). `seatclass.go:34-36`'s archived headings are not migrated and not read: `seatclass` classifies live transcripts; an archived run's classes come from its migrated record. |
| per-round projection | **per epoch.** `TelemetryLine.round` → `TelemetryLine.epoch`; `convergence_vs_verdict` keys on the epoch whose `verdict` it is; `round_verdict` → `verdict`; board JSON keys `acts_this_round`/`last_round`/`closed_round` (`viewjson.go:614,616,142`) → `acts_this_epoch`/`last_epoch`/`closed_seq`; `gap_edit` view's `round` (`schema.sql:789`) → `epoch`. |
| display | `"r%d"` → `"#%d"` (sitting) where the seat is named beside it; dropped where it is not. |

`chart.go:41`'s `if last < 2 { return "" }` — "a single round is not a series" — becomes "fewer
than two chair sittings", which is the same suppression with the same meaning.

### §III.A.3 Gap ids become `G<n>` [MODIFY]

`MintGapID(run)`: `SELECT count(*) FROM "mint"`, formatted `G%d`. Run-global, monotonic, no round.
`G` is the free prefix letter; the two schemes cannot collide because `^G\d+$` and `^R\d+-\d+$`
share no string. Shape readers updated: `flags/shapes.go:40`, `report/md.go:57`,
`idnamespace_test.go:52-53` (reads the shape from flags, so it follows). Archived `R<r>-<n>` ids
are translated at migration (§III.A.5) and never read by a live reader.

### §III.A.4 The chair's job changes; no seat is added [MODIFY]

Revision 4 added a `red-dispatch` seat. It does not exist: **`red-chair` IS the dispatch chair.**
Its job turns around — it stops transcribing lens findings into gaps and starts running the
debate: dispatch, verdict, closing arguments, spot-check. That is the division of labour #831's
rename asserted and did not implement (#846). Role stays `merge` on the record (#831's reason:
`chair` already means a SIDE); the agent body is rewritten and *"AT THE MERGE SEAT YOU
COALESCE"* goes because there is nothing to coalesce. No row is added to `agentrole.go`,
`roster.go`, `seatclass.KnownSeats`, `hookgate`, `seatprobe/build.go` or `promptCatalogue` — the
ten-file cost the audit priced for a new seat, and the reason this is cheaper than revision 4.

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

**#840's registry does half of this as it stands.** `migrate/registry.go:15-32` (on
`origin/main`; absent here) is per-word, but `Translate(old, dst)` receives `dst` *for id
minting* — so `G<n>` in mint order IS expressible there (round 3 corrected revision 5, which
overstated). What needs STATE is the reference remap: the new id applied to every later reference
— `close.gap_id`, `regrade.gap_id`, `closing`, `blue_edit.answers`, `spot_check.ids`, `motion`
docket/grade `gap_id`, `mint.supersedes`, `manifest_row`. And `replay.go:89` passes
`Identity{SeatID, Round}` verbatim — the two fields this plan deletes. So `migrate/{replay,source}.go`
gain a stateful reference remap; seat ids follow the §III.A.1 rule; the sitting is recomputed by
the window. `Append` does not shape-check `mint.gap_id` (`record.go:619`), so the remap is gated
by §V criterion 4.

**Migration and the new refusals.** Re-driving goes through `record.Append`, where §III.B.1's
fourth refusal (every cast lens sat against the head) reads CAST and dispatch pins — which an
archived run has neither of, so a migrated PASS would be refused and criterion 4 would fail on
"same final verdict". So: migration SYNTHESIZES a `Cast` from the archived run's registered seats,
written first; and the head-pin refusal is gated OFF on the migration path, because the property
it enforces held in every archived run by construction (the round loop sat every lens every
round). The material-aware PASS refusal stays ON — an archived PASS over an open material gap was
wrong then too, and migration should say so. Posted to #840, 2026-09-08.

**This reaches back into merged work**: #791's `L\d+` arms in `FindingLabelAlt`/`roleRe`, #831's
`red-merge` acceptance in `roleSeats`, `seatclass`, the dashboard and seatprobe reads. All are
dual-shape readers; all retire once migration covers them. The decisive argument, from the audit:
`store.go:339 ensureSchema` returns early on `hasEvents`, so an archived database keeps its OWN
views verbatim — a Go reader accepting both shapes could never have reached that layer anyway.

Dependency edge, corrected: **both legs are built** on `origin/main` (`replay-migration.md`,
status 2026-09-08 — `jsonlsource.go` + `eraentries.go`; all seven archives migrate). The remap
applies to both sources. Two `store.go` files, distinguished: `internal/record/store.go:82-92` is
the JSONL refusal; `internal/record/recordsql/store.go:339` is `ensureSchema`'s early return and
`:451` the insert.

### §III.B.1 The chair dispatches, through a verb [NEW]

**A seat can read the record; the workflow cannot** (`debate.js` imports no `fs`; every seat is
told `${binDir}/feov-record` at `:226-232`). So dispatch is a seat's act, and the seat is
`red-chair`: it sits, reads the board, returns who to engage. The workflow obeys it and does
nothing else. Each chair sitting opens an epoch (§III.A.0).

**It does not return a composed envelope.** The audit was right that a chair whose word the
workflow simply believes has moved the self-assertion channel, not closed it. So the chair acts
through a VERB — #753's mechanism 3, the piece that plan called *"the highest-value piece"* and an
earlier draft dropped: `feov-record dispatch next` computes readiness FROM THE BOARD (open gaps
below their limits, their parties, their pins) and records the decision as an event. The chair
relays the verb's JSON into its envelope, the same move `claim_count` made. The workflow dispatches
what the record says, and:

- **Readiness is not "open gaps" alone — that made a fresh run terminate into a PASS on an empty
  board with no lens ever sitting** (round 3's first gap; the rubber-stamp #277/#535 exist to
  prevent). Readiness has three sources, and the verb reports all three:
  1. **A lens is ready when the report HEAD is past its last pin.** Head = `events."id"` of the
     latest `blue_edit` or `base_ingest`; a lens's last pin = the `id` its last dispatch was pinned
     to (0 if it has never sat). On a fresh run every cast lens has pin 0 < head, so every lens is
     ready and the board cannot be passed empty. After blue edits, every lens is ready again.
  2. **A gap's LENS and BLUE parties are ready when the gap is open, below its limits (§III.B.2),
     AND MATERIAL** (`current_severity` mass ≥ 2.0). **The BENCH is ready when a gap is AT
     impasse, docketed, and unruled** — a separate clause, because a gap reaches the docket exactly
     when it reaches its limit, so a rule requiring "below its limits" would never ready the bench
     (round 5). The verb dispatches the bench, like every other debate party. **A sub-material gap
     makes no party ready** (gblock, 2026-09-08): it is on the board, blue may answer it when
     dispatched for something else, but a trifle alone cannot spin the cycle. This settles
     PASS-versus-dispatch precedence by construction: if blue is ready for G then G is material, so
     PASS is not permitted; the two cannot both hold.
  3. **The bookends stay workflow-driven, stated:** `frontier`, `blue-lane-N`, `blue-synthesize`
     run BEFORE the first chair sitting (epoch 0); `judge-terminal` and `assemble` run AFTER the
     verb returns a termination. The verb governs the debate between synthesis and terminal, not
     the phases with no board to read. This is why epoch 0 exists and why the `> 0` guards on the
     round-keyed views drop it harmlessly.
- **The dispatch is a RECORDED event, with a schema** — revision 6 named it and gave it none.
  `EVENT_TYPE_DISPATCH` [NEW], body `Dispatch { int64 pin; repeated Party parties }`,
  `Party { string seat_id; repeated string gap_ids }`; tables `dispatch(event_id, pin)`,
  `dispatch_party(event_id, ord, seat_id)`, `dispatch_party_gap(event_id, ord, gap_id)`. The
  per-(seat, gap) rows are what make a null turn computable: a blue sitting dispatched for
  {G1, G2, G3} that acts on G1 is a null turn for G2 and G3. `EVENT_TYPE_CAST` [NEW], body
  `Cast { repeated string seat_ids }`, table `cast_seat(event_id, ord, seat_id)`. Both under seat
  `harness` (the `Append` path `sittingwrite/write.go:86` already takes; `requireDispatchableSeat`
  fires only in `RegisterSeat`, `record.go:191`).
- **The run parameters — `K`, `K_max`, `M` (§III.B.2.2), the convergence fraction — live in
  `inputs/run-config.json` beside `lanes`**, which is where `verdict.go:88` already reads run
  configuration. Not on the event stream: they are the run's terms, set once, and that file is
  their home already.
- **The workflow's relay is audited, because the chair CAN lie about it.** Revision 6 said the
  chair cannot lie about readiness because it does not compute it; that is true of the verb and
  false of the envelope the workflow actually reads — the chair relays the verb's JSON and could
  drop or add a party. So a capture-time PARITY audit (in `capture.go:743`'s rewrite, already in
  scope): within the window [first dispatch, termination), every register between dispatch D and D+1 is
  either a party named in D or **the seat that authors D+1** — the dispatcher registering to run
  the verb is exempt BY RULE, not by allowlist (round 5: `red-chair`'s own register falls between
  every D and D+1); and every party named in D registered. Registers before the first dispatch
  (the bookends) are outside the window by position; `judge-terminal` and `assemble` register
  after termination and are likewise outside it. A mismatch is a FAIL check in §V, not a warning. The verb records the truth; the
  audit catches a relay that departed from it.
- **PASS is refused until every cast lens has SAT against the current head** — a fourth refusal
  beside §III.B.2.1's, in `refs.go`, the direct replacement for the round loop's "every lens sits
  every round". "Sat" is the lens's `register` FOLLOWING the dispatch pinned at the head (its
  sitting ordinal advanced), not the dispatch row: a lens dispatched that never registered has not
  sat (round 5), and rule 1 keeps readying it until it does.
- a seat outside the run's CAST is refused by the verb, not by the workflow. **The cast is on the
  record**: `EVENT_TYPE_CAST` [NEW], written ONCE by `setup` under seat `harness` before any seat
  registers, listing every admissible seat id — the areas selected, `red-chair`, `blue-lane-1..N`
  for the lane count, `blue-synthesize`, `blue-respond`, `frontier`, `judge`, `judge-terminal`,
  `assemble`. `register` and this verb read it. This is the table `derived-seat-identity.md`
  §III.3 named and did not specify; it is specified here and that plan points here. #752.
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

**The exchange unit is defined over SITTINGS, not acts** — round 3 showed an act-based unit is
unreachable: a blue seat dispatched for G that records nothing on G completes no exchange, so
`exchanges(G)` stays 0 and `K_max` bounds nothing. The dispatch event (§III.B.1) names the
parties and the pin, so the record knows which sittings were FOR G:

- **One exchange on G = a red-party sitting dispatched for G, followed by a blue-party sitting
  dispatched for G.** A sitting that closes having recorded no act on G is a **null turn** — it
  still completes the exchange. Silence is a turn taken, and `exchanges(G)` climbs.
- **Movement is a STATE change, not an act:** the gap's grade changed (`regrade` to a different
  value); its anchored text changed (`blue_edit` with `answers = G` and `old != new`); its lineage
  moved (a `mint` superseding G); or it closed. **Null acts** — a turn, not movement: `closing`,
  `motion`, `spot_check`, a `regrade` to the same grade, a `blue_edit` on G with `old == new`.

All four movement joins are direct on the gap id — `blue_edit.answers` names the gap, so no
anchor path is needed and a gap minted without `found_by` is covered. For the record, an act on G
by red is a `mint`, `regrade`, `close`, `closing` or `spot_check` naming G, or a `mint` whose
`supersedes` names G; by blue a `blue_edit` whose `answers` is G, a `closing` on G, or a `motion`
on G (`retire` carries no gap reference, `record.proto:1159-1172`, and reaches G only through the
`blue_edit` that answers it).

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

**Termination** (#753 open problem 1), three ways, each written to `outcome`:

- **VERIFIED** — the chair issues PASS: no material gap open (§III.B.2.1), every docketed gap
  ruled, the bench agrees. The verdict is a dispatch-cycle act, so PASS is reachable in the loop —
  revision 4's termination omitted it.
- **CEILING** — every open MATERIAL gap is at impasse and has had its bench sitting, and at least
  one was ruled `carried` rather than closed. Open sub-material gaps do not enter the quantifier —
  they are never a party, never reach impasse, never see the bench — and are listed in the report
  exactly as §III.B.2.1 lists them at PASS: "open, below material, not certified against". (Round
  5: quantifying over ALL open gaps left a run with one trifle and one carried material gap in no
  terminal state.) `RUN_OUTCOME_CEILING`'s `(means)` moves from "the round
  ceiling was reached" to "every open gap reached its limit"; the value is kept because it is what
  the outcome IS, and archived CEILING runs mean the same thing under the new gloss.
- **HALTED** — unchanged.
- **UNVERIFIED** — the run ended WITHOUT one of the three: chair error, workflow abort, operator
  stop. `RUN_OUTCOME_UNVERIFIED` (`record.proto:291`) keeps exactly that meaning;
  `debate.js:1266`'s arm and `bench outcome --ended` (`cli/bench/outcome.go:115`) keep writing
  it. **`ended` is a different field from the outcome** (`enums.go:240`: `{deadlock, ceiling}`),
and revision 6 conflated them. Post-change `ended` is `{ceiling}` at run level — the run-level
judged deadlock this plan deletes was the other value. An archived `ended = deadlock` is
translated to `ceiling` at migration (the bench could not resolve it; under the new gloss that IS
"every open gap reached its limit"), through `checkWord` (`enums.go:454`) with a registry entry.
Posted to #840.

**Deadlock is per gap; the run-level branch goes.** The bench's `carried` ruling on a gap at
impasse IS that gap's deadlock; `verdict.go:79-82`'s #289 branch ("deadlock not on record") is
deleted with the round-scoped definition it enforced.

`feov-record dispatch next` returns **empty iff no party is ready**, and alongside the party list
it reports two derived facts: `pass_permitted` (§III.B.2.1's predicates all hold and every cast
lens has sat against the head) and `ceiling` (every open MATERIAL gap is at impasse and has had
its bench sitting). The verb does not write the outcome: on `pass_permitted` the chair's sitting
writes PASS through the existing refusals; on `ceiling` the workflow writes CEILING; on empty with
neither, the run ends UNVERIFIED with the verb's report on the record as the reason. HALTED
arrives on its own channel as today.

### §III.B.2.2 THE MINT BUDGET — the run-level bound that is not a clock [NEW]

Round 4 found the fixed point missing: lens sits → mints material G → blue ready → blue edits →
every lens ready again → any lens mints G′ → forever. `K_max` bounds exchanges PER GAP and nothing
bounded the number of gaps; `maxRounds` had been withdrawn with no replacement. **gblock's ruling:
a MINT BUDGET per lens.**

**Each cast lens may mint at most `M` gaps per run.** `M` is a run parameter (default 5, recorded
in `inputs/run-config.json`, a starting point the gate run measures — #753's baseline ran ~20
findings a run across four lenses before the chair coalesced them). A lens whose budget is spent
still sits when the head moves — it verifies, records findings, regrades and closes ITS OWN gaps —
but `mint` is refused, **including a `mint` that supersedes**: lineage is a new gap and counts
against `M`, or the bound `gaps ≤ M × L` does not hold. A spent lens therefore cannot move
lineage, which is consistent with what it can still do. **Cross-lens supersede** — lens B's mint
supersedes lens A's gap, as the nearmatch screen expects — leaves A's gap closeable only by A
(originator closes, §III.B.3); `requireSupersededAreClosed` is satisfied by A's close, and A is
dispatched for it under rule 1 the next time the head moves. The refusal is at the write path, from the record's own count of that
seat's mints.

**Why it is the right bound and not merely a bound.** An epoch cap would also terminate, and it
would be `maxRounds` with the chair's sitting standing in for the round — a global clock in all
but name, which is the thing this plan removes. A budget is PER SEAT, and it puts the cost of a
trifle where the trifle is born: a lens with five mints and a report full of real defects does not
spend one on a nitpick. §III.B.2.1's refusal catches triviality at the verdict; the budget makes
it expensive at the source. That is the incentive gblock asked for, applied one step earlier than
revision 6 could.

**What bounds `--smoke`.** `commands/research.md` sets `maxRounds: 2` for a smoke run. Roundless,
a smoke is `M = 1, K_max = 2` — one mint per lens, two exchanges per gap — which bounds it tighter
than two rounds did and keeps the run shaped like a run rather than a truncation.

**Termination, now provable rather than hoped.** Gaps ≤ `M × |cast lenses|`. Each gap ≤ `K_max`
exchanges (§III.B.2). So total exchanges ≤ `M × L × K_max`. Once every budget is spent, a head
move re-readies lenses that can only regrade or close; no new gap appears; every open gap reaches
impasse or closes; and either no material gap is open and every lens has sat against the head
(VERIFIED) or every open gap is at impasse with the bench having sat (CEILING). The deleted
`exhausted` arm (`debate.js:1268`, `verdict.go:75`) is replaced by this — CEILING's definition is
unchanged; the budget is what guarantees the state is reached.

### §III.B.2.1 Triviality must not hold the report open — the refusal, corrected

`divergent` exists (`schema.sql:988-991`) and refuses nothing. The audit found its predicate wrong
on three counts for this purpose, so the refusal uses a **corrected** predicate rather than the
view as it stands:

| as written | defect | corrected |
|---|---|---|
| `fresh_mints = 0` | a run minting one fresh trifle per sitting is never divergent — the exact trajectory this exists to stop | `fresh MATERIAL mints = 0`: fresh gaps whose `current_severity` mass ≥ 2.0 |
| `max_severity_mass <= 2.0` | inclusive of `GRADE_MEDIUM` = material; would force PASS over a live material gap | `< 2.0`, strict |
| joins `gs.value = g.severity` | the MINT grade; a `regrade` cannot move it | joins `current_severity` (`:845`), the fold over `regrade` |
| `mass < 35.0` | a magic number with no derivation | `mass < 0.25 × peak board mass` for this run; the fraction a run parameter (default 0.25) recorded at setup. A starting point the gate run measures, stated as such |

**Two refusals, and one MODIFIED.** `requirePassClosesAllGaps` (`refs.go:315`, called from
`record.go:1032`) refuses PASS over ANY open gap — so "below material does not hold the gate" was
unreachable without changing it. It becomes `requirePassClosesAllMaterialGaps`: refuses while any
open gap has `current_severity` mass ≥ 2.0. **An open sub-material gap at PASS stays open on the
board and is listed in the report under "open, below material, not certified against"** — not
auto-disposed, not `carried`, not `defect_accepted`; red's finding stays visible and the report
says what it was not certified against. The two new refusals sit beside it: a FAIL is refused
while the corrected predicate holds (red must raise something material or PASS); and a gap below
material is recorded but does not hold the gate. Neither overrides a live material gap — that is
what `< 2.0` on `current_severity` buys.

**The JS twin retires.** `debate.js:1012-1020` computes the same predicate from the envelope
(`mass < 35 && maxSevMass <= MASS['medium'] && freshMints === 0`) and only logs — it is what acts
on `divergent` today, and it acts by printing. It goes; the refusal is the record's.
`convergence.go:34`'s reader and `cost.go:428`'s `convergence_vs_verdict_flags` follow the view.

**Inflation** is the game; `regrade` events are on the record and
`ACCEPTED_DELTA_DOCKET_THRESHOLD` (`debate.js:276`) already dockets cumulative deltas. §V measures
regrade frequency on the gate run rather than assuming it away.

### §III.B.3 Lenses mint; the chair dispatches; dispatch is per dispute [NEW/MODIFY]

**A lens mints its own gap** (#846). `mint` becomes a lens verb (`cli/lens/mint.go`); `finding`
stays as the graded observation a mint may cite. The screen the chair did by eye is already a tool
op — `nearmatch.go`: *"the tool SCREENS — it ranks candidates and never decides"* — so a lens mint
runs it and must either supersede the top match or state why it is distinct. What the chair did
here was measured (#747): coalesced once in twenty findings, dropped three silently. The lossy step
is removed, not delegated.

**The originator closes** (gblock, earlier in this workstream: *"originator closes"*). A gap
belongs to the lens that minted it for its whole life. The verb→role table after the move:

| verb | before | after |
|---|---|---|
| `nearmatch`, `class new`, `mint` | merge | **lens** (`cli/lens/`) |
| `regrade`, `close` on G | merge | **the lens that minted G** — `gap.minted_by` (`schema.sql:821`, `events.seat_id` joined on the mint's `event_id`); refused for any other |
| `finding`, `verify`, `reproduce` | lens | lens |
| `spot_check`, `verdict`, `closing` (as red) | merge | **chair** |
| `dispatch next` | — | **chair** [NEW] |

So the red party to every exchange on G is G's own lens, and the chair is party to NONE — which
is what makes per-dispute dispatch save chair sittings rather than add them. `cli/merge/` loses
`mint.go`, `regrade.go`, `close.go`, `nearmatch.go`, `class.go` to `cli/lens/`, with the
ownership check.

**The chair runs the debate.** Its sitting: read the board through the dispatch verb; issue the
verdict when the board permits one (§III.B.2.1); file closing arguments on docketed gaps;
spot-check. It mints nothing and closes nothing. `debate.js:899-901`'s *"board's only writer…
COALESCE"* prompt and `red-chair.md`'s merge duties are rewritten to this; the lens bodies stop
saying *"the gap ids are the chair's"*.

**Dispatch is per dispute.** The verb engages the seats party to an open gap below its limits —
the lens that minted it, the blue seat answering, the bench if docketed. Seats with nothing open
are not dispatched. That is where wall-clock is won on a two-slot budget.

### §III.C Consumer census — run 2026-09-08, results pasted

```
$ grep -rn '\.Round\b' --include='*.go' . | grep -v _test.go      → 80 lines, 29 files
  false positives (math.Round / time.Round / round2): 16 (capture.go:187,193 · cost.go:381 ·
             liveness.go:143,188,193 · dashboard.go:145-148 · render.go:31,514 · nearmatch.go:66)
  writers:   record.go:271,320,515
  identity:  seatenv/identity.go:107 · cli/seat/seat.go:186,265 · cli/merge/mint.go:51 ·
             dashboard/model.go:207-208,210,552,560
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

$ grep -n '"round"\|_round"' …/schema.sql                            → 27 lines (+ :269,:668 round_verdict)
  column+index: :5, :16 · table: round_verdict :269,:668 ·
  *_round columns: registered :684, minted :820, merge_closed :824, bench_closed :827,
  closed :828, ruled :1046,:1078,:1104, appealed :1049, filed :1069, proposed :1098,
  status :1101 · convergence_vs_verdict :983-1017

$ SQL-in-Go readers (missed by the two greps; found by the audit):
  verdict.go:66 max("round") · queries.go:101 RoundsWithRevision · capture.go:743-744 ·
  inquiry.go:273,304,344 CurrentRoundOf · view/view.go:435 · replay.go:416
$ schema SOURCE: recordsql/schema.go:45,56,127 · recordsql/views.go (40); testdata/schema.sql
  is the golden (golden_test.go:45,62)
$ gap_edit view: schema.sql:789 · register grammar: roster.go:58-64,111 · hookgate.go:114
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
set -e
T=plugins/frank-exchange-of-views/tools
# 1. The round is gone from identity and the record. Re-arms: any reintroduction.
test "$(grep -rnw 'RoundOf\|RoundIn' --include='*.go' $T | wc -l)" -eq 0
test "$(grep -n '"round"\|_round\|round_verdict' $T/internal/record/recordsql/schema.go $T/internal/record/recordsql/views.go | wc -l)" -eq 0
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
         TestTheSittingOrdinalIsRegisterInclusiveAndPerSeat \
         TestTheEpochCountsChairRegistersForEveryRow \
         TestOnlyTheMintingLensMayRegradeOrCloseItsGap \
         TestPassIsRefusedUntilEveryCastLensSatAgainstTheHead \
         TestASilentSittingIsANullTurnThatStillCountsAnExchange \
         TestASubMaterialGapDoesNotReadyBlue \
         TestAMintPastTheBudgetIsRefused \
         TestDispatchedPartiesAndRegistersAgree; do
  n=$(cd $T && go test ./internal/record/ ./internal/cli/... ./internal/capture/ -run "^$t\$" -v 2>&1 | grep -c "^--- PASS: $t")
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
- ~~Keeping `maxRounds` for one run~~ — **withdrawn.** With `event.round` gone, `max("round")`
  bounds nothing; `maxRounds` is retired IN this plan (46 sites, §III.C) and `K_max` per gap is the
  ceiling. A bound on a clock that does not exist is scenery.
- **The attested half of identity** — `plans/derived-seat-identity.md`.
- **Per-lane METHOD as a carrier** (#497).
