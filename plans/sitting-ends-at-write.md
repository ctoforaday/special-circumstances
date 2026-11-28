# Sitting ends stored at the write

Issue #1955. Written 2026-11-28 against main eb994eb27; the `next_opening` rewrite (EventSchema 84) is treated as landed. The measurements, the archaeology and the census cited as "#1955" are in that issue's comments. Paths are under `plugins/frank-exchange-of-views/tools/internal/`; `views.go` lines are as of eb994eb27 and moving. UNVERIFIED marks what was not read or measured.

## I. Summary & Goals

**Objective.** Every whole-record read computes `agent_sitting` and `report_delivery` row by row (`record/recordsql/views.go:654-661`), loads both into `WindowIndex` (`record/windowindex.go:48-58`) and folds them in Go. Rulings of 2026-11-28: store the ends at the write; each reader asks its own statement; lens pin/head numbers leave the work list; the view's rule for no-sitting deliveries survives.

**Success criteria.**
- Each §III.3 statement ≤ 0.2 ms on eiffel 11-11 / mersenne 10-30 (modernc). Today (#1955): failed 1.28 / 1.15 ms, not-sat-through 0.94 / 0.86, one pin 1.36 / 0.94, all lens pins 10.0 / 9.4, full loads 5.2 / 4.5.
- An append to a marked body table costs ≤ 0.3 ms more; any other append, nothing.
- Race-nightly `record/migrate`: the views' 356 s of 1,742 s (#1955) falls below 60 s.
- Stored state equals today's views on all 41 archives before any view body changes.
- `readViews`, `mustViews`, `viewsRead`, `sittings`, `deliveries`: zero hits.

**Value case.** Filtering the view still computes every row's ends, and the plan path gets slower (#1955). Storing makes each read an index seek and removes the panic-on-unloaded-index class (`windowindex.go:73-77`). The price is one trigger-kept table, paid for by a permanent oracle of `gap_state`'s kind.

## II. Technical Context

- Go 1.27, modernc SQLite; the record is written only through `recordsql.InsertTx` (`record/recordsql/store.go:896`; sole non-test caller `record/record.go:824`).
- `InsertTx` writes the `events` row (`store.go:902`) before the body (`:919`): the agent and a bracket's seat are body columns, so a trigger on `events` cannot see them.
- An opening is a `register`, or a seat-naming `sitting_open`, whose `sitting_id` is its own id (`store.go:806-834`, `record/recordpb/sitting.go:61-80`). **No repair arm exists**: `repairs_sitting` is reserved in `Register` (`record.proto`); `recordsql/schema.go:105` is stale.
- A non-opening joins the seat's latest sitting or none (`store.go:836-843`), so a `report_read` can sit in no sitting.
- House pattern: `gap_state`, `gap_dirty`, `gap_flush`, `gap_refresh` (`recordsql/gapstate.go:33-82, :174-200`), oracle `GapStateRefreshDiff` (`:395`); `demand_lineage` writes directly (`demandlineage.go:42-53`).
- Constraints: no compatibility paths; each DDL change bumps the epoch; `feov-migrate` is the only upgrader; planguard plans statements as issued (`store.go:390-396`); release fuzz in CI only.

## III. Proposed Changes

### 1. Stored state `[NEW]` — `recordsql/sittingstate.go`

```
internal/record/recordsql/
  sittingstate.go        [NEW]    tables, markers, triggers, sitting_refresh, SittingStateRefreshDiff
  sittingstate_test.go   [NEW]    seeded trigger cases, the oracle after each append
  views.go, schema.go    [MODIFY] three thin views; two body indexes; the DDL assembly
  delivery.go            [DELETE] in S3/S4
internal/record/migrate/sqlitesource.go [MODIFY] skips the stored tables
```

Names grepped across the plugin (Go tree, `debate.js`, `agents/`, `skills/`) and `scripts/`: `sitting_state`, `sitting_dirty`, `sitting_refresh`, `seat_delivery`, `seat_pin`, `sitting_end(s)`, `sittingMarkers`, `SittingStateRefreshDiff` — zero hits each.

| Object | Shape |
|---|---|
| `sitting_state` | `sitting_id` PK (the opening's id), `seat_id`, `sitting`, `epoch`, `agent_id` (NULL: agentless), `next_opening`, `over`, `closed`, `admitted`, `sat_through`. One row per OPENING, agentless included. Indexes: `(agent_id)`; `(seat_id) WHERE next_opening IS NULL`; partial `WHERE NOT sat_through`; partial `WHERE closed AND NOT admitted`. |
| `sitting_dirty` | `sitting_id` PK; empty between appends. |
| `sitting_refresh` (view) | Today's `agent_sitting` rules (`views.go:642-663`) asked of `sitting_dirty` only, each read driven from the named sitting through an index, as `gap_refresh` is (`gapstate.go:27-31`): the closes by the two body indexes below, and "an opening of that agent between" by `sitting_open_agent` and `register_agent` (`schema.go:262, :265`), never by the view's COALESCE'd `opening` CTE, which no index reaches. Plus `sat_through = 1` for a NULL agent — the rule three readers COALESCE today (`views.go:676, :699, :813`). The rules' one home. |
| `seat_delivery` | `event_id` PK, `seat_id`, `sitting_id` (NULL: none), `head`, `whole` (NULL: not a lens). Append-only by refusal. Index `(seat_id, event_id)`. |
| body indexes | `sitting_close(agent_id, event_id)`, `envelope_admitted(agent_id, event_id)` in `QueryIndexDDL`. |

Triggers come from one Go table, as `gapMarkers` does. Each names sittings, runs `INSERT OR REPLACE INTO sitting_state SELECT * FROM sitting_refresh`, and clears `sitting_dirty`. One stage: an append writes one row to one of these tables and every such insert changes sitting state, so neither reason for `gap_flush` (`gapstate.go:20-25`) applies.

| AFTER INSERT ON | Names | Rule reproduced |
|---|---|---|
| `register`, `sitting_open`, WHEN the event's `sitting_id` is its own id | the new opening; the seat's row with `next_opening IS NULL` (seat from `sittings`) | `next_opening`, `over` (`views.go:646, :653`). Fires for an agentless opening, which ends the seat's previous sitting. A latest sitting keeps NULL. |
| `sitting_close` | every row with the close's `agent_id` | `closed`: no opening of that agent between (`:654-656`), so only the agent's highest row turns. `sat_through`: the agent's FIRST close precedes the seat's next opening (`:660-661`). Holds for one agent carrying two openings (`store.go:826-830`). |
| `envelope_admitted` | the same | `admitted` (`:657-659`). |
| `report_read` | none; inserts the `seat_delivery` row | `seat_id`, `sitting_id` off `events`; `whole` = `wholeDeliverySQL(NEW.event_id, seat)` over the stored tables. |

A register that joins its agent's bracket is no opening and changes nothing: the view reads the opening's own body (`views.go:649-650`).

**`whole` is stored at the insert.** It equals the view's as-the-rows-stand value: `requireReportRead` refuses a second report read in a sitting (`record/deliveredtext.go:165-168`), an earlier sitting's `sat_through` is final once the seat opens a later one, and a no-sitting row is always sat through. Its reader after the trims is `served[].whole` (`record/viewjson.go:2105`); `NextDeliveryWhole` (`deliveredtext.go:51`) asks the same fragment of the next id.

**Views `[MODIFY]`.** `agent_sitting`, `report_delivery`, `area_change` keep names and columns as thin views: `sitting_state WHERE agent_id IS NOT NULL`; `seat_delivery LEFT JOIN sitting_state` with `COALESCE(sat_through, 1)`; `area_change`'s bound renders the pin fragment. The old bodies go in the same PR: no second definition.

### 2. Do we need the pins?

`Dispatch.pin` (`record/dispatchnext.go:134`) is the dispatch's head — a homonym, outside this sweep.

| Use of a seat's pin | Site | Needs |
|---|---|---|
| delivery cut, "changes since" | `record/deliverycut.go:74-78`, `:246-253` | the number, as an id-range bound |
| `area_change` bound | `views.go:692-706` | the number, as a bound |
| spot-check stamp | `record/spotcheck.go:99-109`, read back `:49-61` | the number, stored on the event (`migrate/entries.go:54-61`) |
| lens readiness; stale area | `record/retirement.go:104`, `:362` | a comparison with the head |
| `last_read` kind | `retirement.go:517-531` | comparisons |
| whole-report owed | `record/passblockers.go:885-897` | not the pin: a count of sat-through deliveries |
| upheld position heard | `record/positionheard.go:227` | not the pin: a sat-through delivery after the ruling |
| stale-text refusal | `deliveredtext.go:137-151` | this sitting's delivery head |
| print only | `retirement.go:99, :105`; `passblockers.go:479-493, :836`; `spotcheck.go:260`; `record/sitting.go:194-198`; `viewjson.go:1978`; `StaleArea.Pin` → `debate.js:428` | nothing |

**Recommendation (a): no stored pin.** Three uses need the number as a bound, so the concept stays as ONE fragment — `MAX(head)` over the seat's `seat_delivery` rows whose sitting is sat through, an index range plus rowid seeks — rendered into `pinSQL` and `area_change`. A stored pin adds two writers (the close that turns `sat_through`; a report read into an already-sat sitting) to save a handful of seeks. Its cost is UNVERIFIED: §V step 1 measures it, and above 0.5 ms for all lenses the plan falls back to (b).

### 3. Readers `[MODIFY]` / `[DELETE]`

Census (run 2026-11-28; every hit is a row below or `windowindex.go`): `grep -rn 'satThroughOf\|servedOf\|pinOf\|deliveriesUpTo\|FailedSittingsOf\|AgentSittingsOf\|noTurnSittingsOf\|deliveredIn\|DeliveryLine\|mustViews\|readViews\|viewsRead' --include=*.go . | grep -v _test.go`

| Fold deleted | Callers | Statement | Runs |
|---|---|---|---|
| `FailedSittingsOf(win)` `failedsitting.go:71` | `failedsitting.go:60, :119` (`noTurnSittingsOf` ← `impasse.go:233`), `capture/capture.go:200`, `capture/failedsittings.go:28` | `agent_sitting WHERE closed AND NOT admitted`; `WorkArrivesAsItems` stays Go | loader's read transaction, a `failed` field beside `remands`; capture through `readSnapshot` (`narrowedview.go:111`) |
| `AgentSittingsOf` `:164` | `:149` ← `capture/sittingcloses.go:26` | `agent_sitting`, by run | `readSnapshot` |
| `satThroughOf` `dispatch.go:968` | `dispatch.go:315`, `retirement.go:176, :278`, `closingowed.go:135`, `composedsource.go:137`, `answersretired.go:39` | `SELECT sitting_id, over FROM sitting_state WHERE NOT sat_through`; absent = sat through | reads structs; `ClosingDuties` takes the run |
| `pinOf` | `retirement.go:289`, `spotcheck.go:102` | `pinSQL`; the stamp adds `head_move WHERE event_id > ?` | reads structs; the write's querier, dropping `eventsAt` (`spotcheck.go:94`) |
| `lastReadOf`'s pin; `deliveredIn` `reportread.go:38`; `servedOf` | `retirement.go:516-517`, `sitting.go:481`, `positionheard.go:227` | pin bounded by `event_id <` this sitting's opening, never by `sitting_id` (which drops NULL rows); `deliveryThisSittingSQL`; EXISTS a sat-through delivery after the ruling | loader derives the bound, asks in-transaction |
| `deliveriesUpTo` `windowindex.go:61` | `passblockers.go:892` | sat-through deliveries per lens, `LIMIT 2` | reads structs |
| `delivery`, `DeliveryLine` `viewjson.go:1951, :1970` | `view/view.go:690`, `report/assemble.go:1307` | line deleted; `served` reads `report_delivery` | `debateView`'s querier |

`WindowIndex` keeps `of`, `latest`, `reach`. The report-read arm at `queries.go:163-167` and the loads at `replay.go:99, :110` go.

**Stated behaviour change (ruling 3).** A report read in no sitting counts as sat through and can set a pin for every reader; `servedOf` drops it today (`reportread.go:28-30`).

### 4. Agent-facing text

Census: `grep -rcE 'your pin|last_read|SINCE YOUR PIN|since the pin|its pin|pin [0-9%]|\(pin '` over the plugin and `scripts/agentgen/src`.

| Carrier (hits) | Change |
|---|---|
| `record/sitting.go:194-198` (2); `LastReadJSON` `retirement.go:490-494` | drop `pin`, `head`; keep `kind` |
| `record/deliverycut.go:68-72, :88` (4); `deliveredtext.go:250, :283` (3) | the whole re-read marks inline what changed since the last delivery |
| `cli/seat/verbs.go:330, :331, :369` (3); `cli/seat/help/{register,dispatch,spot-check}.md` → `help_gen.go` (`cmdtreegen`) | reword; regenerate |
| `scripts/agentgen/src/{red-lens-voice.md, red-chair.md, fragments/lens-duties.md, skills/*/SKILL.md}` (5) → `agents/`, `agents-claude/` (12 files each; `scripts/agentgen/main.go:70`), `skills/` | reword; `agentgen` |
| `skills/research-protocol/scripts/debate.js:886, :900` authored; `:409, :428` generated, the chair's | reword `lastReadClause`; chair lines per Q3 |
| goldens: `difftest/testdata` (15 files), `tests/simulator/testdata` (4), `record/migrate/testdata` (1); `tests/simulator/debate.test.mjs` (8) | regenerate; read every line |
| `dispatch next`: plan `record/dispatch.go:651`, plan hold `passblockers.go:615`, standing item `:602` | print once: the plan's readiness line survives |
| `### REPORT DELIVERED` `viewjson.go:1984`, `report/reportread_test.go` | delete |
| `record/migrate/planids.go:36` | unchanged: its phrases describe archived bytes; the trimmed lines carry no id |
| `sittingwrite/delivery.go:75` | carries `LastReadJSON`: changes with it; the inline mark renders here |
| the word, no number: `record/queue.go:48-49`, `itemkind.go:22, :65`, `view/changes.go:21, :55`, `hooks/README.md:115`, `docs/seat-command-triggers.md:75`, `references/report_template.md:8`, `docs/vocabulary.md:517` (`vocabdoc`) | unchanged: ruling 2 removes the lens's numbers, not the word |
| the chair's: `cli/chair/dispatch.go:90`, `cli/chair/spot_check.go:51`, `record/passblockers.go:479-493`, `record/dispatch.go:21`, `retirement.go:99, :105` | Q3 |
| another sense (a cite's git pin, a hash pin): `setup/setup.go:222`, `setup/run.go:202`, `skills/research/SKILL.md:29`, `native/build.sh:107, :167` | not carriers |

### 5. Pull requests

| PR | Concept | Deletes | Epoch |
|---|---|---|---|
| T1 | `dispatch next` states the whole-report debt once; `REPORT DELIVERED` goes | two line arms, `DeliveryLine` | — |
| T2 | the lens work list names no pin or head; the re-read is marked inline | `LastReadJSON.Pin`, `.Head` | — |
| S1 | sitting ends and deliveries are stored; the views read them | the three computing view bodies | the next |
| S2 | failed and agent sittings are statements | `FailedSittingsOf(win)`, `AgentSittingsOf` | — |
| S3 | sat-through is a statement | `sittings`, `AgentSittingRow`, `recordsql.AgentSittings` | — |
| S4 | deliveries and the pin are statements; ruling 3 | `servedOf`, `pinOf`, `deliveriesUpTo`, `delivery`, `deliveries`, `viewsRead`, `mustViews`, `readViews`, `ReportDeliveries` | — |

T1 and T2 do not depend on the table and ship first. S1's epoch carriers: `requirements.json`, `recordpb/schema_gen.go`, `recordsql/testdata/schema.epoch.json`, `schema.sql`, `cmd/feov-migrate/testdata/help.golden`.

**Migration.** The replay writes every event through `InsertTx`, so the triggers fill the tables (`migrate/sqlitesource.go:267-272` says so of `gap_state`); that file gains an arm skipping the three new tables. Residue: a migrated spot-check's pins are restamped by the live rule (`migrate/entries.go:54-61`), so after S4 one can differ from its source where a no-sitting delivery exists.

**Safety net (a174f37df).** Land the seeded fixtures (790 lines, 15 records) with S2; ruling 3 then shows as a reviewed golden diff in S4. Keep the 6,140-line archive golden on the working branch through S4, then delete it: ~77 s under the race detector in a package over its nightly bound, and prose that churns on other streams' wording. On main the archive check is the oracle (§V 3), facts only.

## IV. Risk & Mitigation

| Risk | L × I × cost | Mitigation |
|---|---|---|
| A trigger fails to name a sitting an append changes | M × H × L | the refresh-diff oracle on archives, fuzz and seeded cases (§V 2-4) |
| An opening of a third type (the write path makes none) | L × H × L | CHECK on `events`: an own-id `sitting_id` only on `register` / `sitting_open` |
| Stored `whole` drifts from the rule | L × M × L | the pre-change equivalence includes `whole`; a seeded second-read-refused case |
| planguard does not plan a trigger's statements | M × M × L | a test issues `SELECT * FROM sitting_refresh` over seeded dirty rows |
| An inline mark misses a change, which reads as "unchanged" | M × H × M | the delivery states its change count and appends any change it could not place as a BEFORE/AFTER item (Q6) |
| Capture reads outside one snapshot (`replay.go:97-112`) | M × M × L | capture moves to `readSnapshot` in S2 |
| T2 removes the head a lens copies into `close --verified-against` (25 of 63 behind sittings, #1955) | H × L × L | stated loss: the close's own event id places it; no replacement number |
| Golden conflicts across ~8 streams | H × L × L | rebase on golden or schema change; regenerate at `-count=1` |

### Open questions for the owner

| # | Decision | Recommendation |
|---|---|---|
| Q1 | A stored per-seat pin? | (a) none: one fragment over the stored rows; (b) only if §V 1 measures over 0.5 ms |
| Q2 | Trigger form: refresh through `sitting_dirty` (the rule stays one SQL text, with an oracle) or point updates (cheaper writes, the rule restated imperatively) | refresh |
| Q3 | Do the chair's pin numbers go too (plan why-lines, `stale_areas[].pin`, the spot-check echo)? | not here: plan bytes are stored and keyed, and `spot_check.pins` is a record field; rule it separately |
| Q4 | `whole`: trigger-stored by the rule, stamped by the writer on `ReportRead`, or dropped | trigger-stored |
| Q5 | The archive golden | working branch only, then delete |
| Q6 | A change the re-read cannot mark inline | append it as an item and state the counts |
| Q7 | Which `dispatch next` line survives | the plan's readiness line |

### UNVERIFIED

- That no agent index exists on `sitting_close` / `envelope_admitted` (a grep of `recordsql/*.go` found none).
- Whether planguard plans trigger-body statements (`plans.golden:131, :875` hold the `gap_refresh` view and the flush insert).
- Stored-state timings and the per-append trigger cost: targets, not measurements.
- Which of golden lines :123 / :124 sit inside the stored plan block; whether any reader matches the trimmed phrases.
- What migrate does with an unknown non-body table, and how it registers an epoch step with no body translation.
- How many archived deliveries sit in no sitting.
- Whether inline marking can place every change in the current text.
- Whether "pin" is a registered term (`terms/terms.json` holds no `"pin"` key by grep; `docs/vocabulary.md:517` uses the word); test file line numbers in §V 4 are the #1955 design pass's, not re-read.

## V. Verification Plan

First: `unset $(env | grep -o '^FEOV_[A-Z_]*')`. `FEOV_RELEASE_GATE` is never set locally.

1. **Timing, before code.** Scratch copies of the eiffel and mersenne archives with the tables built by replay; time each §III.3 statement, the all-lens pin, and an append to each marked table with and without its trigger, arms interleaved, setup outside the timer. Decides §III.2 and Q2.
2. **Equivalence before the views change (real data).** S1's first commit carries `TestStoredSittingStateEqualsTheViewsOnEveryArchivedRun` over `archiveNames` (`record/backingarm_test.go:304`): `EXCEPT` both ways for `agent_sitting`, `report_delivery` (`whole` included) and `area_change`, with non-vacuity counts (130 still open, 423 deliveries, 154 whole — #1955). It is CI-green as the PR head before the commit that repoints the views; the PR body names that run.
3. **Permanent oracle.** `SittingStateRefreshDiff` names every sitting from the `sittings` view — never from `sitting_state`, where a missing row would hide itself — and compares. Asserted after EACH append in the seeded tests (an end-state check misses a row stale mid-run), on the archive sweep, in `record/plansweep_test.go:52`, and in the release fuzz beside the gap-state oracle.
4. **Seeded trigger tests `[NEW]`**: an agentless opening ends the previous sitting; a latest sitting keeps a NULL next opening; closed then admitted after a nudge; one agent with two openings; a close after the next opening is not sat through; a report read in no sitting. Mutation: delete each marker row, expect red. Existing (files confirmed, lines per the #1955 design pass): `satthrough_test.go`, `lensturn_test.go`, `deliveryviews_test.go`, `upheldwhole_test.go`, `behinddelivery_test.go`, `spotcheckborne_test.go`, `debateserved_test.go`.
5. `go -C scripts run ./check`; `TestEveryDDLChangeBumpsTheEpoch` with the restamped epoch golden; `shapes.golden` gains one line, `sitting_refresh`, under `PLANGUARD_ACCEPT_VIEW_COST=1` (reason for the commit: it walks `sitting_dirty`, which holds one append's sittings), and loses the three computing views' cost.
6. After text changes: `agentgen`, `enginegen`, `cmdtreegen`, each with `-check`; read every changed golden line and each affected `--help` before and after.
7. CI on the final head: the four `release-fuzz` parts and `release-fuzz-tally`; race-nightly's `record/migrate` time against 1,742 s.
8. Auditor gate: `/prosthetic-conscience:plan-audit`.
