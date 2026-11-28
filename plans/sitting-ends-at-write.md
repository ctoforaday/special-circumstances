# Sitting ends stored at the write

Issue #1955. Written 2026-11-28 against main eb994eb27; revised after two audit rounds and the owner's rulings of that day. EventSchema 84 (`next_opening` bound once) has merged as 6065a1a9e; this plan's schema change takes **the next epoch** after whatever main then holds. Measurements, archaeology and census cited as "#1955" are in that issue's comments. Paths are under `plugins/frank-exchange-of-views/tools/internal/`; `views.go` lines are as of eb994eb27 and moving. UNVERIFIED marks what was not read or measured.

**Rulings implemented** (the owner, 2026-11-28, as relayed by the coordinating session): (1) sitting ends are stored at the write and each reader asks its own statement; (2) the pin is derived, never stored; (3) a report read in no sitting counts as sat through — the view's rule survives, the folds' goes; (4) of the three statements of a lens's whole-report debt in `dispatch next`, the plan-hold reason alone goes, and the class is #1963; (5) the `REPORT DELIVERED` lines leave the debate view; (6) the chair's pin and head numbers stay, #1962.

**Split.** The lens's pin and head numbers and the marking of its whole-report re-read are `plans/lens-reread-marking.md`. The two plans are independent: that plan's delivery cut reads the pin through `PinAt` (`record/deliverycut.go:74`), which exists today and which this plan re-points at the stored rows without changing its answer, except for ruling 3. Each takes an epoch; whichever merges second regenerates its epoch carriers. Both touch `lastReadOf` (`record/retirement.go:504-535`): there S4's statement serves the `kind` alone once that plan has removed the numbers.

## I. Summary & Goals

**Objective.** Every whole-record read computes `agent_sitting` and `report_delivery` row by row (`record/recordsql/views.go:654-661`), loads both into `WindowIndex` (`record/windowindex.go:48-58`) and folds them in Go. Rulings of 2026-11-28: store the ends at the write; each reader asks its own statement; the view's rule for no-sitting deliveries survives.

**Success criteria.**
- Each single-question statement of §III.3 ≤ 0.2 ms on eiffel 11-11 / mersenne 10-30 (modernc): measured 0.007–0.08 ms against stored tables (§III.2), where today's take 0.9–9.9 ms. The whole `area_change` view, every lens's bound at once, ≤ 0.3 ms: measured 0.16 / 0.17 ms at load 15 and 0.26 / 0.17 ms at load 76, where today's takes 3.9–7.1 ms.
- An append to a marked body table costs ≤ 0.3 ms more; any other append, nothing. Measured in S1's first commit (§V 1).
- Race-nightly `record/migrate`: the views' 356 s of 1,742 s (#1955) falls below 60 s.
- Stored state equals today's views on every archived run (39 at origin/main 5f0911494) before any view body changes.
- `readViews`, `mustViews`, `viewsRead`, `AgentSittingRow`, `ReportDeliveries`: zero hits; `WindowIndex` has no `sittings` or `deliveries` field.

**Value case.** Filtering the view still computes every row's ends, and the plan path gets slower (#1955). Storing makes each read an index seek and removes the panic-on-unloaded-index class (`windowindex.go:73-77`). The price is one trigger-kept table, paid for by a permanent oracle of `gap_state`'s kind.

## II. Technical Context

- Go 1.27, modernc SQLite; the record is written only through `recordsql.InsertTx` (`record/recordsql/store.go:896`; sole non-test caller `record/record.go:824`, which the replay writer reaches too).
- `InsertTx` writes the `events` row (`store.go:902`) before the body (`:919`): the agent and a bracket's seat are body columns, so a trigger on `events` cannot see them.
- **Ids are minted `max + 1` inside the writing transaction (`store.go:801-804`).** Every new row's id exceeds every existing one. This invariant is what makes the trigger table in §III.1 complete: a new opening exceeds every existing close and admission, so it never lies "between" an earlier opening and its close (`views.go:654-659`), and an earlier sitting's `closed` / `admitted` cannot change at an opening. The oracle (§V 3) is its gate.
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
| `sitting_state` | `sitting_id` PK (the opening's id), `seat_id`, `sitting`, `epoch`, `agent_id` (NULL: agentless), `next_opening`, `over`, `closed`, `admitted`, `sat_through`. One row per OPENING, agentless included. Indexes: `(agent_id)`; `(seat_id) WHERE next_opening IS NULL`; `(sitting_id) WHERE NOT sat_through`; `(sitting_id) WHERE closed AND NOT admitted`. |
| `sitting_dirty` | `sitting_id` PK; empty between appends. |
| `sitting_refresh` (view) | Today's `agent_sitting` rules (`views.go:642-663`) asked of `sitting_dirty` only, each read driven from the named sitting through an index, as `gap_refresh` is (`gapstate.go:27-31`): the closes by the two body indexes below, and "an opening of that agent between" by `sitting_open_agent` and `register_agent` (`schema.go:262, :265`), never by the view's COALESCE'd `opening` CTE, which no index reaches. Plus `sat_through = 1` for a NULL agent — the rule three readers COALESCE today (`views.go:676, :699, :813`). The rules' one home. |
| `seat_delivery` | `event_id` PK, `seat_id`, `sitting_id` (NULL: none), `head`, `whole` (NULL: not a lens). Append-only by refusal. Index `(seat_id, event_id)`. |
| body indexes | `sitting_close(agent_id, event_id)`, `envelope_admitted(agent_id, event_id)` in `QueryIndexDDL`. |

**Decided: the triggers refresh through `sitting_dirty`**, not by point updates. The rule stays one SQL text and the stored state has an oracle; point updates would restate the rule imperatively with nothing to hold them to it. Triggers come from one Go table, as `gapMarkers` does. Each names sittings, runs `INSERT OR REPLACE INTO sitting_state SELECT * FROM sitting_refresh`, and clears `sitting_dirty`. One stage: an append writes one row to one of these tables and every such insert changes sitting state, so neither reason for `gap_flush` (`gapstate.go:20-25`) applies.

| AFTER INSERT ON | Names | Rule reproduced |
|---|---|---|
| `register`, `sitting_open`, WHEN the event's `sitting_id` is its own id | the new opening; the seat's row with `next_opening IS NULL` (seat from `sittings`) | `next_opening`, `over` (`views.go:646, :653`). Fires for an agentless opening, which ends the seat's previous sitting. A latest sitting keeps NULL. By the id invariant (§II) no other row can change. |
| `sitting_close` | every row with the close's `agent_id`, on whichever seat | `closed`: no opening of that agent between (`:654-656`), so only the agent's highest row turns. `sat_through`: the agent's FIRST close precedes the seat's next opening (`:660-661`), per row and per seat. Holds for one agent carrying two openings, on one seat or two (`store.go:826-830`). |
| `envelope_admitted` | the same | `admitted` (`:657-659`). |
| `report_read` | none; inserts the `seat_delivery` row | `seat_id`, `sitting_id` off `events`; `whole` = `wholeDeliverySQL(NEW.event_id, seat)` over the stored tables. |

A register that joins its agent's bracket is no opening and changes nothing: the view reads the opening's own body (`views.go:649-650`).

**Decided: `whole` is stored by the trigger at the insert.** It equals the view's as-the-rows-stand value: `requireReportRead` refuses a second report read in a sitting (`record/deliveredtext.go:165-168`, called at `record.go:1508` for every writer, the replay included), an earlier sitting's `sat_through` is final once the seat opens a later one, and a no-sitting row is always sat through. Its reader after the trims is `served[].whole` (`record/viewjson.go:2105`); `NextDeliveryWhole` (`deliveredtext.go:51`) asks the same fragment of the next id.

**Views `[MODIFY]`.** `agent_sitting`, `report_delivery`, `area_change` keep names and columns as thin views: `sitting_state WHERE agent_id IS NOT NULL`; `seat_delivery LEFT JOIN sitting_state` with `COALESCE(sat_through, 1)`; `area_change`'s bound renders the pin fragment. The old bodies go in the same PR: no second definition.

### 2. The pin: measured, and not stored

`Dispatch.pin` (`record/dispatchnext.go:134`) is the dispatch's head — a homonym, outside this sweep.

| Use of a seat's pin | Site | Needs |
|---|---|---|
| delivery cut, "changes since" | `record/deliverycut.go:74-78`, `:246-253` | the number, as an id-range bound |
| `area_change` bound | `views.go:692-706` | the number, as a bound |
| spot-check stamp | `record/spotcheck.go:99-109`, read back `:49-61` | the number, stored on the event (`migrate/entries.go:54-61`) |
| lens readiness; stale area | `record/retirement.go:104`, `:362` | a comparison with the head |
| `last_read` kind | `retirement.go:517-531` | comparisons |
| whole-report owed | `record/passblockers.go:885-897` | not the pin: a count of sat-through deliveries |
| upheld position heard | `record/positionheard.go:227` | not the pin: a sat-through delivery in a sitting opened after the ruling |
| stale-text refusal | `deliveredtext.go:137-151` | this sitting's delivery head |
| print only | `retirement.go:99, :105`; `passblockers.go:479-493, :836`; `spotcheck.go:260`; `record/sitting.go:194-198`; `viewjson.go:1978`; `StaleArea.Pin` → `debate.js:428` | nothing |

**Measured 2026-11-28** (modernc, eb994eb27, best of 25, the two arms adjacent in one process). `sitting_state` and `seat_delivery` were built as TEMP tables from today's views plus agentless openings, with this plan's indexes; `EXCEPT` both ways against `agent_sitting`, `report_delivery` and `area_change` returned zero rows on both records. One-minute load was 15–16 on 16 cores, and the "today" arm reproduces #1955's figures within 10%. Milliseconds, eiffel 11-11 / mersenne 10-30:

| Question | Today's view | Stored tables |
|---|---|---|
| failed sittings | 1.17 / 1.05 | 0.021 / 0.012 |
| sittings not sat through | 0.92 / 0.85 | 0.007 / 0.007 |
| every agent sitting | 3.04 / 2.75 | 0.080 / 0.075 |
| one lens's pin | 1.35 / 0.91 | 0.024 / 0.022 |
| all seats' pins, grouped | 9.87 / 9.47 | 0.051 / 0.046 |
| a pin before this sitting (by event id) | — | 0.034 / 0.033 |
| this sitting's delivery | 0.081 / 0.091 | 0.019 / 0.019 |
| a lens's sat-through deliveries, capped at two | — | 0.017 / 0.016 |
| upheld position heard | — | 0.019 / 0.019 |
| every `report_delivery` row | 1.95 / 1.61 | 0.052 / 0.044 |
| every `area_change` row | 3.95 / 3.92 | 0.163 / 0.171 |

Plans: the one-lens pin is `SEARCH seat_delivery USING INDEX (seat_id=?)` plus a rowid seek. The failed statement reads the partial index `WHERE closed AND NOT admitted` once it carries no `agent_id IS NOT NULL` term — with the term the planner searched `(agent_id)` over every agent sitting (0.034 / 0.033 ms); the term is redundant, since an agentless row is never closed (zero such rows on both records). That one figure is from a second run at load 76. The grouped pin is a `SCAN … USING INDEX`, so the plan path asks one pin per cast lens (0.02 ms each) instead. Limits of this data: both records hold no agentless opening, no delivery in no sitting, no sitting not sat through and no upheld position, so those arms were timed on empty results. Raw output and the SQL are in scratch `issue-1955/plan-rev/` and go on #1955.

**Decided — ruled by the owner 2026-11-28: the pin is derived, never stored.** Three uses need the number as a bound, so the concept stays as ONE fragment — `MAX(head)` over the seat's `seat_delivery` rows whose sitting is sat through — rendered into `pinSQL` and `area_change`. At 0.02 ms a stored copy would save nothing and add two writers. The fallback is deleted.

### 3. Readers `[MODIFY]` / `[DELETE]`

Census (run 2026-11-28; the audit's re-run found no omission): `grep -rn 'satThroughOf\|servedOf\|pinOf\|deliveriesUpTo\|FailedSittingsOf\|AgentSittingsOf\|noTurnSittingsOf\|deliveredIn\|DeliveryLine\|mustViews\|readViews\|viewsRead' --include=*.go . | grep -v _test.go`

| Fold deleted | Callers | Statement | Runs |
|---|---|---|---|
| `FailedSittingsOf(win)` `failedsitting.go:71` | `failedsitting.go:60, :119` (`noTurnSittingsOf` ← `impasse.go:233`), `capture/capture.go:200`, `capture/failedsittings.go:28` | `sitting_state WHERE closed AND NOT admitted` (no agent term: an agentless row is never closed); `WorkArrivesAsItems` stays Go | loader's read transaction, a `failed` field beside `remands`; capture through `readSnapshot` (`narrowedview.go:111`) |
| `AgentSittingsOf` `:164` | `:149` ← `capture/sittingcloses.go:26` | `agent_sitting`, by run | `readSnapshot` |
| `satThroughOf` `dispatch.go:968` | `dispatch.go:315`, `retirement.go:176, :278`, `closingowed.go:135`, `composedsource.go:137`, `answersretired.go:39` | `SELECT sitting_id, over FROM sitting_state WHERE NOT sat_through`; absent = sat through | reads structs; `ClosingDuties` takes the run |
| `pinOf` | `retirement.go:289`, `spotcheck.go:102` | `pinSQL`, once per cast lens; the stamp adds `head_move WHERE event_id > ?` | reads structs; the write's querier, dropping `eventsAt` (`spotcheck.go:94`) |
| `lastReadOf`'s pin; `deliveredIn` `reportread.go:38` | `retirement.go:516-517`, `sitting.go:481` | pin bounded by `event_id <` this sitting's opening, never by `sitting_id` (which drops NULL rows); `deliveryThisSittingSQL` | loader derives the bound, asks in-transaction |
| `servedOf` in `upheldPositionBlockers` | `positionheard.go:227` | the fold's predicate, unchanged: `EXISTS (seat_delivery d JOIN sitting_state s ON s.sitting_id = d.sitting_id WHERE d.seat_id = ?lens AND d.sitting_id > ?ruled AND s.sat_through)` — a sat-through delivery in a sitting OPENED after the ruling. A delivery completing after the ruling in a sitting opened before it does not clear the blocker; a no-sitting row never does | loader passes each upheld ruling's place |
| `deliveriesUpTo` `windowindex.go:61` | `passblockers.go:892` | sat-through deliveries per lens, `LIMIT 2` | reads structs |
| `delivery` `viewjson.go:1951` (`DeliveryLine` `:1970` went in T1) | `viewjson.go:2104-2105` | `served` reads `report_delivery`: seat, head, `whole` | `debateView`'s querier (`viewjson.go:2159`) |

`WindowIndex` keeps `of`, `latest`, `reach`. The report-read arm at `queries.go:163-167` and the loads at `replay.go:99, :110` go.

**Stated behaviour change (ruling 3).** A report read in no sitting counts as sat through and can set a pin for every reader; `servedOf` drops it today (`reportread.go:28-30`). The upheld-position predicate is the one reader it does not reach.

### 4. Agent-facing text

**T1 — the whole-report debt (ruled 2026-11-28).** `dispatch next` states a lens's whole-report debt three times: golden lines :123 and :124 of `difftest/testdata/docket_ruling_requires_each_unconditional_field.golden` are why-lines of the STORED plan — the readiness line (`record/dispatch.go:651`) and the plan-hold reason (`record/passblockers.go:615`) — and :131 is the chair's standing item (`passblockers.go:602`). Only the plan-hold reason goes. The readiness line and the standing item stay, so the chair's standing still lists everything the PASS gate refuses on (`record/sitting.go:381-394`).

Consumer census, run 2026-11-28:

| Search | Hit | Changes? |
|---|---|---|
| `grep -rn 'kindOf(\|WorkItem()\|\.why\b' --include=*.go . \| grep -v _test` | a kind's `why` is rendered at `record/dispatch.go:711` and nowhere else | the whole-report kind's `why` (`passblockers.go:613-616`) is deleted; `:711` skips a kind with none, and so do the two tests that walk every kind's `why` (`record/export_test.go:50`, `record/passblockers_test.go:272`) |
| | a kind's `item`, through `WorkItem()` (`passblockers.go:742`): `record/sitting.go:375, :392, :470, :495, :503, :520, :524`; `cli/lens/verify.go:413, :415` | unchanged |
| `grep -rn 'wholeOwedWhat\|BlockerWholeReportOwed' --include=*.go . \| grep -v _test` | `record/retirement.go:71, :282`; `passblockers.go:47, :598-616, :906, :911-912`; `record/verdict.go:176`; `record/dispatch.go:649-651` | only `:613-616` |
| `grep -rln 'ready at any head'` over the plugin | `difftest/testdata/docket_ruling_requires_each_unconditional_field.golden`, `difftest/testdata/projections_debate_changelog_citations.golden` | regenerated, each changed line read; no `.mjs`, `.json` or `.js` reader matches the phrase |
| `record/migrate/planids.go:36` | its phrases are id phrases | unchanged: the deleted line carries no id; archived plans keep it |

Deleting a why arm changes the bytes, and so the key, of plans stored from then on. The class — other debts print the same three lines, the question-split debt among them (golden :122, :125, :132) — is #1963, decided there.

**`REPORT DELIVERED` leaves the debate view.**

| Search | Hit | Changes? |
|---|---|---|
| `grep -rn 'DeliveryLine\|REPORT DELIVERED'` over the plugin and `scripts/` | `record/viewjson.go:1965-1985` (`DeliveryLine`) | deleted; `delivery` (`:1951-1963`) stays for `served` until S4 |
| | `view/view.go:690`, `report/assemble.go:1307` | the call and its loop go |
| | `report/reportread_test.go:27, :62`; `record/debateserved_test.go:41-60` | the line's tests go; `debateserved_test.go:16`'s `served` test stays |
| | no golden, agent, skill or help page holds the phrase | — |

The debate JSON's `served` (`viewjson.go:1933-1936, :2105`) stays and reads `report_delivery`.

**Ruling 3's sweep.** `git grep -nE 'sat through|sit through'` over the plugin and `scripts/agentgen/src`, generated copies, tests and goldens apart: 82 hits, 32 on a line that names a pin or a delivery; the other 50 speak of sittings sat through and were classed by that filter, not read one by one.

| Hits | Change in S4 |
|---|---|
| the pin defined, third person: `terms/terms.json:862, :873`; `recordpb/record.proto:313`; `cli/seat/help/dispatch.md:13`; `record/dispatch.go:21` (→ `debate.js:394`); `record/passblockers.go:479`; `scripts/agentgen/src/red-chair.md:14`; `docs/seat-command-triggers.md:75` | each gains the clause "a delivery outside any sitting counts as sat through"; regenerate (`protogen`; `vocabgen`; `vocabdoc` for `docs/vocabulary.md:517, :523`; `enginegen`; `cmdtreegen`; `agentgen`); `tests/simulator/debate.test.mjs:574, :1478` follow |
| second person, about the reader's own sitting: `scripts/agentgen/src/fragments/lens-duties.md:5`; `record/reportread.go:13` (`ServedPinNote`); `record/sitting.go:194` | survive: a seat reading them is in a sitting, where the sentence is exact |
| comments that state the rule: `record/deliveredtext.go:54, :67, :70`; `record/reportread.go:5, :42`; `record/retirement.go:68, :501`; `record/viewjson.go:1950`; `record/dispatch.go:310`; `record/passblockers.go:883`; `recordpb/record.proto:1178, :2593`; `recordsql/views.go:672` | reworded with the code they describe, or deleted with their fold |
| another subject: `capture/sittingcloses.go:15, :47` (closes), `record/positionheard.go:251` (an instruction), `record/retirement.go:136` (a zero pin, which still means no sat-through delivery of either kind) | survive |

**Not this plan's.** The chair's pin and head numbers — plan why-lines, `stale_areas[].pin`, the spot-check echo — stay (ruled 2026-11-28; #1962). The lens's numbers and their census are `plans/lens-reread-marking.md`. `headingOf` dropping a change item's "In section:" line is #1964.

### 5. Pull requests

| PR | Concept | Deletes | Epoch |
|---|---|---|---|
| T1 | `dispatch next` drops the plan-hold reason for a whole-report debt; `REPORT DELIVERED` leaves the debate view | the whole-report kind's `why`; `DeliveryLine` | — |
| S1 | sitting ends and deliveries are stored; the views read them | the three computing view bodies | the next |
| S2 | failed and agent sittings are statements | `FailedSittingsOf(win)`, `AgentSittingsOf` | — |
| S3 | sat-through is a statement | the `sittings` field, `AgentSittingRow`, `recordsql.AgentSittings` | — |
| S4 | deliveries and the pin are statements; ruling 3 | `servedOf`, `pinOf`, `deliveriesUpTo`, `delivery`, the `deliveries` field, `viewsRead`, `mustViews`, `readViews`, `ReportDeliveries` | — |

T1 does not depend on the table and ships first. S1's carriers: `requirements.json`, `recordpb/schema_gen.go`, `recordsql/testdata/schema.epoch.json`, `schema.sql`, `cmd/feov-migrate/testdata/help.golden`, `record/testdata/plans.golden`, `record/testdata/shapes.golden`. No pull request leaves a dual rule: S1 deletes the view bodies with the tables' arrival, and each of S2–S4 deletes the fold it replaces.

**Migration.** The replay writes every event through `InsertTx` (`record/replaywriter.go:84` → `record.go:824`), so the triggers fill the tables; `migrate/sqlitesource.go:262-272` gains an arm skipping the three stored tables, without which a source holding their rows is refused (`:284-293`). Residue: a migrated spot-check's pins are restamped by the live rule (`migrate/entries.go:54-61`), so after S4 one can differ from its source where a no-sitting delivery exists.

**Safety net (a174f37df).** The seeded fixtures (790 lines, 15 records) land with S1, so they stand under the rule while S1 replaces its definition; ruling 3 then shows as a reviewed golden diff in S4. The 6,140-line archive golden stays on the working branch through S4 and is then deleted (Q5): ~77 s under the race detector in a package over its nightly bound, and prose that churns on other streams' wording. On main the archive check is the oracle (§V 3), facts only.

## IV. Risk & Mitigation

| Risk | L × I × cost | Mitigation |
|---|---|---|
| A trigger fails to name a sitting an append changes | M × H × L | the oracle after every append, on archives and in the fuzz; the narrowing mutations (§V 3-4) |
| An opening of a third type (the write path makes none) | L × H × L | CHECK on `events`: an own-id `sitting_id` only on `register` / `sitting_open` |
| Stored `whole` drifts from the rule | L × M × L | the oracle recomputes it (§V 3); the pre-change equivalence includes it |
| planguard does not plan a trigger's statements | M × M × L | a test issues `SELECT * FROM sitting_refresh` over seeded dirty rows |
| Capture reads outside one snapshot (`replay.go:97-112`) | M × M × L | capture moves to `readSnapshot` in S2 |
| Golden conflicts across ~8 streams | H × L × L | rebase on golden or schema change; regenerate at `-count=1` |

### Decided

- The pin is derived, never stored — ruled by the owner 2026-11-28; measured in §III.2.
- A report read in no sitting counts as sat through — ruled 2026-11-28; the stated behaviour change of §III.3.
- Of the whole-report debt's three statements only the plan-hold reason goes; the class is #1963 — ruled 2026-11-28.
- The chair's pin and head numbers stay; #1962 — ruled 2026-11-28.
- The triggers refresh through `sitting_dirty` (§III.1) — this plan's.
- `whole` is stored by the trigger at the insert (§III.1) — this plan's.

### Open question for the owner

| # | Decision | What happens unless he says otherwise |
|---|---|---|
| Q5 | The 6,140-line archive golden of a174f37df | it stays on the working branch through S4 and is then deleted; it never lands |

### UNVERIFIED

- That no agent index exists on `sitting_close` / `envelope_admitted` (a grep of `recordsql/*.go` found none).
- Whether planguard plans trigger-body statements, and whether it accepts a `SCAN … USING INDEX` (`plans.golden:131, :875` hold the `gap_refresh` view and the flush insert).
- The trigger's per-append cost: a target until S1's first commit measures it.
- The stored-state timings on a record with agentless openings, no-sitting deliveries, unsat sittings or an upheld position (neither measured record has one).
- The archive count: 39 at origin/main; #1955's design pass says 41.
- The 50 `sat through` hits that name no pin or delivery: classed by a filter, not read.
- That `Motion.ruled` (`record/motion.go:243-245`) is the ruling's event id in every loader; whether a `means` annotation edit moves the epoch.
- How migrate registers an epoch step with no body translation.
- Test file line numbers in §V 4's existing list are the #1955 design pass's, not re-read.

## V. Verification Plan

First: `unset $(env | grep -o '^FEOV_[A-Z_]*')`. `FEOV_RELEASE_GATE` is never set locally.

1. **Timing.** The read side is measured (§III.2). S1's first commit measures the write side: an append to each marked table with and without its trigger, arms interleaved, setup outside the timer, against the 0.3 ms criterion; and repeats the read table against the real tables.
2. **Equivalence before the views change (real data).** S1's first commit carries `TestStoredSittingStateEqualsTheViewsOnEveryArchivedRun` over `archiveNames` (`record/backingarm_test.go:304`): `EXCEPT` both ways for `agent_sitting`, `report_delivery` (`whole` included) and `area_change`, with non-vacuity counts (130 still open, 423 deliveries, 154 whole — #1955). It is CI-green as the PR head before the commit that repoints the views; the PR body names that run. The repointing commit deletes the test — once the views read the tables it compares them with themselves — and the oracle (3) takes over.
3. **Permanent oracle.** `SittingStateRefreshDiff` names every sitting from the `sittings` view — never from `sitting_state`, where a missing row would hide itself — refreshes, and compares; and it recomputes every `seat_delivery` row from `report_read` and `events`, `whole` by the fragment asked of the row's own id. Asserted after EACH append in the seeded tests (an end-state check misses a row stale mid-run), on the archive sweep, in `record/plansweep_test.go:52`, and in the release fuzz beside the gap-state oracle.
4. **Seeded trigger tests `[NEW]`** (`recordsql/sittingstate_test.go`): an agentless opening ends the previous sitting; a latest sitting keeps a NULL next opening; closed then admitted after a nudge; one agent with two openings on one seat; **one agent opening on TWO seats** (the case that separates "every row of the agent" from "the agent's highest row"); a close after the next opening is not sat through; a report read in no sitting; a second report read in a sitting is refused. Mutations, each expected red: delete each marker row; **narrow each naming set** (the close's to the agent's highest row; the opening's to the new row alone). In package `record`, with S4: **an upheld position whose lens has a delivery complete after the ruling in a sitting opened before it — the blocker stands — then a sat-through delivery in a sitting opened after it — the blocker clears**; mutation: bound the statement by the delivery's id instead of its sitting's, expect red. Existing (files confirmed): `satthrough_test.go`, `lensturn_test.go`, `deliveryviews_test.go`, `upheldwhole_test.go`, `behinddelivery_test.go`, `spotcheckborne_test.go`, `debateserved_test.go`.
5. `go -C scripts run ./check`; `TestEveryDDLChangeBumpsTheEpoch` with the restamped epoch golden; `shapes.golden` gains one line, `sitting_refresh`, under `PLANGUARD_ACCEPT_VIEW_COST=1` (reason for the commit: it walks `sitting_dirty`, which holds one append's sittings), and loses the three computing views' cost.
6. After text changes (T1, S4): `agentgen`, `enginegen`, `cmdtreegen`, `protogen`, `vocabgen`, `vocabdoc`, each gate's `-check`; read every changed golden line and each affected `--help` before and after.
7. CI on the final head: the four `release-fuzz` parts and `release-fuzz-tally`; race-nightly's `record/migrate` time against 1,742 s.
8. Auditor gate: `/prosthetic-conscience:plan-audit`.
