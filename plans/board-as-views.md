# The board becomes views: retiring the BoardState fold, reader by reader

## I. Summary & goals

**Goal: simplification, not performance.** A question about the record is authored once, as
SQL a reader can see and a SQL test can hold — never recomputed by Go pulling tables and
walking events. `BoardState` is the last large Go fold: one function replays every event
into a `*record.Board`, and 34 non-test call sites then walk that struct, each taking its
own slice of it. Step 4 (#700/#701/#703) converted the point questions; #708 moved the
question families (`gap`, `motion_answers`, `line_of_inquiry`) into `ViewsDDL`. What
remains is the board's own consumers: the seat-facing JSON projections, the markdown
renders, report assembly, capture's audits, the dashboard, the scorecard, the graph.

**The end state:** each projection reads views (or plain queries over event tables) into
its OWN output type. `Board` is not rebuilt from SQL — it is retired per reader, because a
God struct fed by views is the same God struct. What outlives the fold is the small set of
readers that genuinely need the ordered event stream with full bodies (§IV), and the
consistency oracle, which keeps its raw walk precisely so the converted projections have an
independent ground truth to be held against.

**Non-goals:** no behaviour change visible to any seat, renderer, or auditor — every
converted projection must produce byte-identical output on the same record; no schema
TABLE changes (views and their tests only); no new performance claims (perf already fell
out of #700 and is not the bar here).

## II. Decisions

1. **Per-reader retirement, not a Board rebuilt on views.** Rebuilding `Board` from SQL
   would keep every consumer coupled to one struct whose shape nobody chose on purpose.
   Each projection instead asks its own questions of the views. The fold shrinks until its
   remaining callers are the deliberate ones in §IV, then the dead fields come off `Gap`
   and `Board` — deletion is the measure of progress.
2. **Views grow by FAMILY, not per caller.** When a wave needs a fact no view carries
   (closure prose, anchor triples, found_by lists, observation fates), the fact joins the
   family view that owns the entity, with a SQL test in `recordsql` — the #708 pattern.
   A wave that needs a fact only ONE reader will ever ask MAY use a plain query, stated
   in the PR; a second asker is the signal it belongs in the view.
3. **Byte-identity is the conversion contract.** Every projection here feeds a golden, a
   difftest transcript, the fuzz differential gate, or the report — surfaces that diff
   bytes. A conversion that changes bytes is wrong until the change is separately argued.
4. **The consistency oracle is the arbiter and is never converted.** Its package doc
   already states it derives ground truth "deliberately not through record.BoardState";
   after each wave it holds the converted projection to the raw walk. Where a wave
   converts a projection the oracle does not yet check, the wave ADDS that check first —
   convert nothing you cannot cross-examine.
5. **Ordering facts come from the record, not re-derived.** Board order is
   `gap.minted_event`; event order is `events.id`. No reader re-sorts by timestamp or
   re-joins what a view already carries — that is the two-copies defect this whole line
   removes.

## III. Order of work

Each wave is one PR, sized to review. Every wave: (a) extend the family views + SQL tests,
(b) convert the callers, (c) add or confirm the oracle's cross-check, (d) byte-diff the
wave's rendered surfaces on a driven run, (e) delete whatever Board fields lost their last
reader. The 34 call sites, assigned:

**Wave 1 — the seat-facing JSON projections** (`internal/record/viewjson.go:418,681,773,
834,959`; consumed by `cli/seat/verbs.go:478,587`, `cli/verify.go:42`). Board JSON, the
worklist (`show work`: blocks, closed_index/estoppel register, affordances), findings
JSON. The gap view gains the closure family (closure prose, anchor triple, carried_from,
successor, closed_by) and found_by; findings become a `finding_state` family view.
This wave is load-bearing for every seat and goes first while attention is highest.

**Wave 2 — motions, evidence, inquiries, verdict** (`record/motionview.go:109`,
`record/evidenceview.go:367`, `record/verdict.go:51`, `record/refs.go:316` — the PASS
gate's motion arm, whose "unruled" semantics (`motion_answers.ruled_by IS NULL`, plus
direction motions created by their rulings) the family views can now state honestly;
converting it was deliberately refused in #703 while they could not.

**Wave 3 — the markdown renders** (`internal/view/view.go:143,155,228`: ledger, archive,
debate, lines-of-inquiry, changes). Same families, prose columns included; goldens and the
difftest transcripts are the byte gate.

**Wave 4 — report assembly and docs** (`report/assemble.go:111`, `report/docs.go:70`).

**Wave 5 — the operators' consumers** (`capture/capture.go:562,801,906,1122,1717`,
`capture/proofbacking.go:93`, `capture/lanecoverage.go:113`, `capture/proofrerun.go:114`,
`cli/scorecard.go:56`, `cli/tiers.go:54`, `cli/graph.go:40`, `dashboard/model.go:282`,
`dashboard/render.go:166`). Mostly consumers of wave-1/2 projections; whatever still
reaches for the Board directly gets its question named and viewed.

**Wave 6 — merge-side text logic** (`cli/merge/mint.go:176` EstoppelConflict,
`cli/merge/nearmatch.go:35`). Row queries feed the existing Go text-matching; the
matching itself stays Go (§IV).

**Wave 7 — deletion.** `Board`/`Gap` fields with no remaining reader come off; what
remains of the fold is exactly §IV's list, each with its reason at the site.

**The TEST and FUZZER carriers — tracked, not truncated.** ~60 `BoardState(` calls live in
22 test files, and they are three different things with three fates:

- **Tests OF a converted projection** convert IN the wave that converts their subject —
  a projection reading views while its test still walks the fold is the half-state that
  reads as done.
- **The parity tests** (#700/#701/#703's `*_parity_test.go`) exist to hold query against
  fold; each retires in the wave that deletes the LAST field its fold half reads, and
  wave 7's PR lists every retirement by name. Until then they are the fold's remaining
  legitimate test readers.
- **The releasegate fuzz harness** (`releasegate/fuzz/fuzz_test.go`, `coverage_test.go`)
  reads `Board`/`Gap` as its ORACLE — including the ledger differential gate §V.3 leans
  on. Before wave 7 deletes any field the harness reads, the harness is re-pointed at the
  same views the production readers use (a wave-6.5 slice of its own), and the oracle
  duty passes to the consistency check, which never converts. The fold's OWN tests
  (replay_test.go and kin) live and die with the fold in wave 7.

A wave MAY land in slices, but a slice is a complete concept: a projection either still
reads the fold or reads views — never half of each behind one name.

## IV. What stays Go, and why (the standing exceptions)

- `consistency.Check` (`consistency/consistency.go:197`) — the arbiter (§II.4).
- `seatprobe.Read` (`seatprobe/seatprobe.go:213`) — wants full event bodies; that is
  `recordsql.Events`' job, not a view's.
- Text matching (EstoppelConflict's quote matching, nearmatch scoring) — Go logic over
  rows the views hand it; SQL string-distance would be a second implementation of a rule
  Go already owns.
- `lastActivity` — SQL date parsing is a different timestamp parser than RFC3339Nano
  (documented at the site, #701).
- Surface spelling (hyphen joins, `undefined` vs em-dash rendering of ungraded axes) —
  presentation, decided per consumer, never a record fact.

## Risks, graded (likelihood × impact × cost-to-mitigate)

| Risk | Grade | Answer |
|---|---|---|
| A converted projection silently disagrees with the fold on an edge no test seeds (regrade/closure interleavings, legacy multi-rule rows) | med × high × low | §II.3 byte-identity + §II.4 oracle-first: no projection converts before the oracle checks it, and #708's raw-`Insert` pattern seeds guard-forbidden states in SQL tests |
| Byte-identity masks nondeterminism the fold currently hides (map iteration, JSON key order) — the diff flaps instead of failing honestly | med × med × low | Drive the byte gate (§V.3) twice per side before trusting one diff; a flapping surface is a finding about the projection, fixed (stable ordering from `minted_event`/`events.id`) before conversion proceeds |
| Two wave-5 consumers (scorecard, graph) have NO golden or transcript byte gate today — the dashboard DOES (whole-page goldens, `dashboard/testdata/render-*.golden`, which the conversion treats as immutable gates, never surfaces to re-pin) | high × med × med | Each ungated wave-5 slice first pins its surface (a golden or an oracle check) and only then converts — a conversion with no gate is refused by this plan, not waved through |
| Wave 7 deletion breaks an out-of-plan reader (new code landed mid-arc still walking the fold) | med × low × low | Deletion compiles or it doesn't — the census re-runs in wave 7's PR rather than trusting §III's snapshot; anything new converts or is added to §IV with a reason |
| The fuzz harness re-point (wave 6.5) changes what the sweep can catch — its oracle was independent of the views, and becomes them | low × high × med | The consistency oracle stays raw and becomes the sweep's independent arbiter; the harness re-point PR must show the ledger differential gate still fails on a seeded divergence (kill one mutant by hand) |
| Risk accepted: the arc spans many PRs and main moves under it | — | Each wave is self-contained and green; a stalled arc leaves no half-state, only unconverted readers — the same shape the codebase has today |

## V. Verification plan

Per wave, in order, each written into the wave PR:

1. `go test ./internal/record/recordsql -count=1` — the new family-view SQL tests,
   including guard-forbidden states seeded through raw `Insert` (the #708 pattern), and
   the regenerated schema golden READ as a diff.
2. `go test ./internal/record ./internal/consistency -count=1` — the oracle's cross-check
   for the wave's projection exists and passes; parity tests hold view reads to the fold
   they replace while both still exist.
3. Byte gate: drive a seeded run through the real binary and byte-compare every rendered
   surface the wave touches (`show board/work/findings/motions/…`, the markdown renders,
   the assembled report) before and after the conversion — the same differential
   discipline the fuzz's ledger gate applies. Goldens and difftest transcripts must pass
   UNCHANGED; a regenerated golden in a conversion wave is a defect, not an update.
4. `go test -count=1 ./...` on the module, then `go run ./check` — nothing merges red
   (#694's lesson is written in this repo's history).
5. After wave 7: the releasegate fuzz sweep once, as the end-to-end drive over the fully
   converted surface.

Re-arms: any edit under `internal/record`, `internal/view`, `internal/report`,
`internal/capture`, `internal/dashboard`, or `ViewsDDL` re-runs 1–4 for the touched wave.
