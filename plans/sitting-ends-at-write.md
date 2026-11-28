# Sitting ends stored at the write

Issue #1955. Written 2026-11-28 against main eb994eb27; revised after three audit rounds and the owner's rulings of that day, and re-read against origin/main 5f0911494, which holds EventSchema 84 (`next_opening` bound once, merged as 6065a1a9e). This plan's schema change takes **the next epoch** after whatever main then holds. Measurements, archaeology and census cited as "#1955" are in that issue's comments. Paths are under `plugins/frank-exchange-of-views/tools/internal/`; Line references are origin/main 5f0911494's: of the source files cited here epoch 84 moved `views.go` alone (eb994eb27's lines 642–650 by +7, later ones by +8), and of the goldens `plans.golden`, whose two cited lines are already main's. The "today" timings are eb994eb27's (§III.2 says what epoch 84 changes). UNVERIFIED marks what was not read or measured.

**Rulings implemented** (the owner, 2026-11-28, as relayed by the coordinating session): (1) sitting ends are stored at the write and each reader asks its own statement; (2) the pin is derived, never stored; (3) a report read in no sitting counts as sat through — the view's rule survives, the folds' goes; (4) of the three statements of a lens's whole-report debt in `dispatch next`, the plan-hold reason alone goes, and the class is #1963; (5) the `REPORT DELIVERED` lines leave the debate view; (6) the chair's pin and head numbers stay, #1962.

**Split.** The lens's pin and head numbers and the marking of its whole-report re-read are `plans/lens-reread-marking.md`. The two plans are independent: that plan's delivery cut reads the pin through `PinAt` (`record/deliverycut.go:74`), which exists today and which this plan re-points at the stored rows without changing its answer: `PinAt` asks `pinSQL` of `report_delivery` (`record/deliveredtext.go:68`), whose rule already counts a delivery in no sitting as sat through, so ruling 3 changes nothing there. Each takes an epoch; whichever merges second regenerates its epoch carriers. Both touch `lastReadOf` (`record/retirement.go:504-535`): there S4's statement serves the `kind` alone once that plan has removed the numbers.

## I. Summary & Goals

**Objective.** Every whole-record read computes `agent_sitting` and `report_delivery` row by row (`record/recordsql/views.go:662-669`), loads both into `WindowIndex` (`record/windowindex.go:48-58`) and folds them in Go. Rulings of 2026-11-28: store the ends at the write; each reader asks its own statement; the view's rule for no-sitting deliveries survives.

**Success criteria.**
- Each single-question statement of §III.3 ≤ 0.2 ms on eiffel 11-11 / mersenne 10-30 (modernc): measured 0.007–0.08 ms against stored tables (§III.2), where today's take 0.9–9.9 ms (at eb994eb27, before EventSchema 84: §III.2). The whole `area_change` view, every lens's bound at once, ≤ 0.3 ms: measured 0.16 / 0.17 ms at load 15 and 0.26 / 0.17 ms at load 76, where today's takes 3.9–7.1 ms.
- An append to a marked body table costs ≤ 0.3 ms more, as a warm insert AND at a cold prepare; any other append, nothing. Both numbers are measured in S1's first commit and the criterion is judged on both (§V 1).
- Race-nightly `record/migrate`: the views' 356 s of 1,742 s (#1955, before EventSchema 84) falls below 60 s.
- Stored state equals today's views on every archived run (the 39 tarballs `archiveNames` globs at origin/main 5f0911494: §V 2) before any view body changes.
- `readViews`, `mustViews`, `viewsRead`, `AgentSittingRow`, `ReportDeliveries`: zero hits; `WindowIndex` has no `sittings` or `deliveries` field.

**Value case.** Filtering the view still computes every row's ends, and the plan path gets slower (#1955). Storing makes each read an index seek and removes the panic-on-unloaded-index class (`windowindex.go:73-77`). The price is one trigger-kept table, paid for by a permanent oracle of `gap_state`'s kind.

## II. Technical Context

- Go 1.27, modernc SQLite; the record is written only through `recordsql.InsertTx` (`record/recordsql/store.go:896`; sole non-test caller `record/record.go:824`, which the replay writer reaches too).
- `InsertTx` writes the `events` row (`store.go:902`) before the body (`:919`): the agent and a bracket's seat are body columns, so a trigger on `events` cannot see them.
- **Ids are minted `max + 1` inside the writing transaction (`store.go:801-804`).** Every new row's id exceeds every existing one. This invariant is what makes the trigger table in §III.1 complete: a new opening exceeds every existing close and admission, so it never lies "between" an earlier opening and its close (`views.go:662-667`), and an earlier sitting's `closed` / `admitted` cannot change at an opening. The oracle (§V 3) is its gate.
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
  views.go, schema.go    [MODIFY] three thin views; two body indexes; the DDL assembly; the `events` CHECK
  delivery.go            [DELETE] in S3/S4
internal/record/migrate/sqlitesource.go [MODIFY] skips the stored tables
```

Names grepped across the plugin (Go tree, `debate.js`, `agents/`, `skills/`) and `scripts/`: `sitting_state`, `sitting_dirty`, `sitting_refresh`, `seat_delivery`, `seat_pin`, `sitting_end(s)`, `sittingMarkers`, `SittingStateRefreshDiff` — zero hits each.

| Object | Shape |
|---|---|
| `sitting_state` | `sitting_id` PK (the opening's id), `seat_id`, `sitting`, `epoch`, `agent_id` (NULL: agentless), `next_opening`, `over`, `closed`, `admitted`, `sat_through`. One row per OPENING, agentless included. Indexes: `(agent_id)`; `(seat_id) WHERE next_opening IS NULL`; `(sitting_id) WHERE NOT sat_through`; `(sitting_id) WHERE closed AND NOT admitted`. |
| `sitting_dirty` | `sitting_id` PK; empty between appends. |
| `sitting_refresh` (view) | Today's `agent_sitting` rules (`views.go:649-671`) asked of `sitting_dirty` only, each read driven from the named sitting through an index, as `gap_refresh` is (`gapstate.go:27-31`): the closes by the two body indexes below, and "an opening of that agent between" by `sitting_open_agent` and `register_agent` (`schema.go:262, :265`), never by the view's COALESCE'd `opening` CTE, which no index reaches. Plus `sat_through = 1` for a NULL agent — the rule three readers COALESCE today (`views.go:684, :707, :821`). The rules' one home. |
| `seat_delivery` | `event_id` PK, `seat_id`, `sitting_id` (NULL: none), `head`, `whole` (NULL: not a lens). Append-only by refusal. Index `(seat_id, event_id)`. |
| body indexes | `sitting_close(agent_id, event_id)`, `envelope_admitted(agent_id, event_id)` in `QueryIndexDDL`. |
| CHECK on `events` | A row whose `sitting_id` is its own id is a `register` or a `sitting_open`. A table CHECK in `EnvelopeDDL`, beside the two `recordsql/schema.go:102, :109` hold; carried by `schema.sql`. |

**Decided: the triggers refresh through `sitting_dirty`**, not by point updates. The rule stays one SQL text and the stored state has an oracle; point updates would restate the rule imperatively with nothing to hold them to it. Triggers come from one Go table, as `gapMarkers` does. Each names sittings, runs `INSERT OR REPLACE INTO sitting_state SELECT * FROM sitting_refresh`, and clears `sitting_dirty`. One stage: an append writes one row to one of these tables and every such insert changes sitting state, so neither reason for `gap_flush` (`gapstate.go:20-25`) applies.

**Fork, decided and stated: over criterion 2, the fallback is point updates.** If §V 1 measures either write-side number over 0.3 ms, each trigger instead UPDATEs, on the rows it names, only the columns its table can change, and `sitting_refresh` stays as the rule the oracle recomputes (§V 3) — which is then what holds the point updates to the rule. The price is the rule stated twice, which is why it is the fallback and not the design. The other branch, accepting a cost over the criterion, is not taken. The switch is made in S1 before the views are repointed and is stated under that pull request's "Plan deviations" (§III.5). It is not the pin's stored fallback, which is deleted (§III.2).

**Decided: the CHECK on `events` is S1's DDL**, not a mitigation left in §IV. The write path cannot violate it: `sittingOf` gives an event its own id as `sitting_id` only where `SeatOpeningSitting` says it opens a sitting (`store.go:806-834`), which is a register or a seat-naming `sitting_open` (`record/recordpb/sitting.go:61-74`); every other event joins the seat's latest sitting or none (`store.go:836-843`), and the replay writes through the same path. So the CHECK refuses only a row the write path did not place, the job `schema.go:103-109`'s CHECK does for a register with no sitting, and the trigger table rests on it: an opening on a third table would fire no trigger.

| AFTER INSERT ON | Names | Rule reproduced |
|---|---|---|
| `register`, `sitting_open`, WHEN the event's `sitting_id` is its own id | the new opening; the seat's row with `next_opening IS NULL` (seat from `sittings`) | `next_opening` (`views.go:653`, found at `:658`), `over` (`:661`). Fires for an agentless opening, which ends the seat's previous sitting. A latest sitting keeps NULL. By the id invariant (§II) no other row can change. |
| `sitting_close` | every row with the close's `agent_id`, on whichever seat | `closed`: no opening of that agent between (`:662-664`), so only the agent's highest row turns. `sat_through`: the agent's FIRST close precedes the seat's next opening (`:668-669`), per row and per seat. Holds for one agent carrying two openings, on one seat or two (`store.go:826-830`). |
| `envelope_admitted` | the same | `admitted` (`:665-667`). |
| `report_read` | none; inserts the `seat_delivery` row | `seat_id`, `sitting_id` off `events`; `whole` = `wholeDeliverySQL(NEW.event_id, seat)` over the stored tables. |

A register that joins its agent's bracket is no opening and changes nothing: the view reads the opening's own body (`views.go:656-657`).

**Decided: `whole` is stored by the trigger at the insert.** It equals the view's as-the-rows-stand value: `requireReportRead` refuses a second report read in a sitting (`record/deliveredtext.go:165-168`, called at `record.go:1508` for every writer, the replay included), an earlier sitting's `sat_through` is final once the seat opens a later one, and a no-sitting row is always sat through. Its reader after the trims is `served[].whole` (`record/viewjson.go:2105`); `NextDeliveryWhole` (`deliveredtext.go:51`) asks the same fragment of the next id.

**Views `[MODIFY]`.** `agent_sitting`, `report_delivery`, `area_change` keep names and columns as thin views: `sitting_state WHERE agent_id IS NOT NULL`; `seat_delivery LEFT JOIN sitting_state` with `COALESCE(sat_through, 1)`; `area_change`'s bound renders the pin fragment. The old bodies go in the same PR: no second definition.

### 2. The pin: measured, and not stored

`Dispatch.pin` (`record/dispatchnext.go:134`) is the dispatch's head — a homonym, outside this sweep.

| Use of a seat's pin | Site | Needs |
|---|---|---|
| delivery cut, "changes since" | `record/deliverycut.go:74-78`, `:246-253` | the number, as an id-range bound |
| `area_change` bound | `views.go:700-714` | the number, as a bound |
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

**The "Today's view" column predates EventSchema 84.** On main the full `agent_sitting` load is 1.80 / 1.66 ms (#1959's measurement, from 3.19 / 2.90), not the 3.04 / 2.75 this table's harness read at eb994eb27. The column's other rows read through `agent_sitting` too and were not re-measured on epoch 84 (UNVERIFIED), nor was criterion 1's 0.9–9.9 ms range. The stored-tables column and the criteria are absolute and do not depend on it; S1's first commit repeats the whole table on its own base (§V 1).

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

**Fork, decided and stated: ruling 1 is partly met for `satThroughOf`.** Its six callers do not each ask a statement of their own. ONE statement reads the not-sat-through set once (the 0.007 ms row of §III.2) and each caller compares its debt's place against that set in Go, as it compares against the fold today. The other branch, a statement per debt, is not built: each caller asks once per debt item, and that is a query per item where one read answers them all. This is the one fold whose callers share a read, and S3's pull request body says so under "Plan deviations" (§III.5).

`WindowIndex` keeps `of`, `latest`, `reach`. The report-read arm at `queries.go:163-167` and the loads at `replay.go:99, :110` go.

**Stated behaviour change (ruling 3).** A report read in no sitting counts as sat through and can set a pin for the readers that fold `servedOf`, which drops it today (`reportread.go:28-30`): `pinOf` in the lens fold and the spot-check stamp, and `lastReadOf`. `PinAt` and `area_change` already follow the view's rule (`pinSQL`, `record/deliveredtext.go:68`; `views.go:707`): ruling 3 changes neither's answer. The upheld-position predicate is the one reader it does not reach.

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

**The working branch's test follows in the same step.** a174f37df's renderer names no `DeliveryLine`: both copies (`sittingdeliveryrender_test.go:283-297`, in `record` and `record/migrate`) read the `### REPORT DELIVERED` lines out of `debate.md` and count them as `deliveries` and `deliveries of the changes since a pin above 0`, and the archive test requires both non-zero. T1 therefore breaks it at run time, not at compile. In T1's step the working branch is rebased onto T1, the line arm goes, each of the two totals is re-keyed on the debate view's `served` row where that row carries the fact and dropped where only the line's sentence did, both goldens are regenerated, and the diff — delivery lines and those totals, nothing else — is read. That golden is the baseline of §V 8.

Deleting a why arm changes the bytes, and so the key, of plans stored from then on. The class — other debts print the same three lines, the question-split debt among them (golden :122, :125, :132) — is #1963, decided there.

**`REPORT DELIVERED` leaves the debate view.**

| Search | Hit | Changes? |
|---|---|---|
| `grep -rn 'DeliveryLine\|REPORT DELIVERED'` over the plugin and `scripts/` | `record/viewjson.go:1965-1985` (`DeliveryLine`) | deleted; `delivery` (`:1951-1963`) stays for `served` until S4 |
| | `view/view.go:690`, `report/assemble.go:1307` | the call and its loop go |
| | `report/reportread_test.go:27, :62`; `record/debateserved_test.go:41-60` | the line's tests go; `debateserved_test.go:16`'s `served` test stays |
| | no golden, agent, skill or help page holds the phrase | — |

The debate JSON's `served` (`viewjson.go:1933-1936, :2105`) stays and reads `report_delivery`.

**Ruling 3's sweep.** `git grep -nE 'sat through|sit through'` over the plugin and `scripts/agentgen/src`, apart from `agents/`, `agents-claude/`, `record.pb.go`, `*_gen.go`, `_test.go` files and `testdata/`; re-run on origin/main 5f0911494: 82 hits, 32 on a line that names a pin or a delivery, 50 on a line that names neither. All 82 are read. The same-line filter is no classifier — three of the 50 are the pin's definition continued from the line above — so the 50 are classed by what each says: **5 carry the pin's or the delivery's rule and change** (marked †), **43 speak of a SITTING sat through and stay true** (ruling 3 moves no sitting's `sat_through`), **2 use the words in another sense**. Property: every statement of the pin's rule is a row below, with its pull request. The property is bounded by the spellings swept. The hyphenated `sat-through|sit-through` adds two hits, both rows below (‡). `sits through` adds six (`record/composedsource.go:25`; `record/dispatch.go:546, :661`; `record/retirement.go:157, :219`; `record/sitting.go:458`), each of a sitting, which stay true. A definition that says only "its latest complete delivery" (`record/retirement.go:265`, `record/sitting.go:101`) names no sitting and stays true under ruling 3.

| Hits | Change | PR |
|---|---|---|
| the pin defined, third person: `terms/terms.json:862, :873`; `recordpb/record.proto:313`; `cli/seat/help/dispatch.md:13`; `record/dispatch.go:21` (→ `debate.js:394`); `record/passblockers.go:479`; `scripts/agentgen/src/red-chair.md:14`; `docs/seat-command-triggers.md:75` | each gains the clause "a delivery outside any sitting counts as sat through"; regenerate (`protogen`; `vocabgen`; `vocabdoc` for `docs/vocabulary.md:517, :523`; `enginegen`; `cmdtreegen`; `agentgen`); `tests/simulator/debate.test.mjs:574, :1478` follow | S4 |
| second person, about the reader's own sitting: `scripts/agentgen/src/fragments/lens-duties.md:5`; `record/reportread.go:13` (`ServedPinNote`); `record/sitting.go:194` | survive: a seat reading them is in a sitting, where the sentence is exact | — |
| comments that state the rule: `record/deliveredtext.go:54, :67, :70`; `record/reportread.go:5, :42`; `record/retirement.go:68, :501`; `record/viewjson.go:1950`; `record/dispatch.go:310`; `record/passblockers.go:883`; `recordpb/record.proto:1178, :2593`; `recordsql/views.go:680` (names the pin; no hit of the sweep) | reworded with the code they describe, or deleted with their fold | S4; `views.go:680` with its view in S1 |
| † the pin defined on the line after the one that names it: `record/retirement.go:45` (the lens fold's account, from `:44`), `:133` (`servedClause`, from `:132`); `record/spotcheck.go:84` (the stamp's, from `:83`) | each gains the clause of the first row; `spotcheck.go:83-84` also stops naming the lens fold as the pin's source, which becomes `pinSQL` (§III.3) | S4 |
| † `recordsql/views.go:675`: `report_delivery`'s comment gives `sat_through` as "the sitting's" and names the agentless sitting alone | rewritten with the view body it heads: the row's `sat_through` is its sitting's, and 1 for a delivery in no sitting — the rule the view applies today (`:684`) | S1 |
| † `record/reportread.go:16`: `served`, "the sitting it was in" | deleted with `servedOf` | S4 |
| ‡ `recordsql/views.go:678`: `whole` counts "rows of the lens in sat-through sittings"; `:696`: `area_change`'s bound, "the head of its latest sat-through delivery" | rewritten with the view bodies they head. `:678` gains the clause: a row in no sitting counts, as the view counts it today (`:821`). `:696` names the delivery, not a sitting, and stays exact | S1 |
| another subject: `capture/sittingcloses.go:15, :47` (closes), `record/positionheard.go:251` (an instruction), `record/retirement.go:136` (a zero pin, which still means no sat-through delivery of either kind) | survive | — |
| a sitting sat through, 43: `hooks/README.md:194`; `capture/capture.go:632, :686`; `capture/sittingcloses.go:52`; `cli/seat/help/dispatch.md:21`; `record/answersretired.go:20`; `record/closingowed.go:38, :39, :53, :57`; `record/composedsource.go:68, :103, :116, :130`; `record/dispatch.go:537, :565, :590, :625, :654, :952, :956, :961, :962, :965, :967, :991, :997`; `record/inquirycheck.go:108`; `record/passblockers.go:116`; `record/queue.go:147`; `record/reportread.go:37`; `record/retirement.go:67, :93, :101, :113, :141, :181, :292`; `record/sitting.go:475, :477`; `recordsql/views.go:631`; `terms/terms.json:841` and its generated `docs/vocabulary.md:505` | stay true. `dispatch.go:654` and `sitting.go:475, :477` state the upheld-position predicate, which ruling 3 does not reach. Ten sit on a declaration this plan rewrites, the rule unchanged: `views.go:631` (the agentless opening is no row of `agent_sitting` and is sat through; the rule's home becomes `sitting_refresh`), `dispatch.go:952-997` (eight, `satThroughOf`), `reportread.go:37` (`deliveredIn`) | S1; S3; S4 |
| another sense, 2: `record/migrate/planids.go:20, :27` | phrases of archived plan lines, quoted for id translation; unchanged | — |

**Not this plan's.** The chair's pin and head numbers — plan why-lines, `stale_areas[].pin`, the spot-check echo — stay (ruled 2026-11-28; #1962). The lens's numbers and their census are `plans/lens-reread-marking.md`. `headingOf` dropping a change item's "In section:" line is #1964.

### 5. Pull requests

| PR | Concept | Deletes | Epoch |
|---|---|---|---|
| T1 | `dispatch next` drops the plan-hold reason for a whole-report debt; `REPORT DELIVERED` leaves the debate view | the whole-report kind's `why`; `DeliveryLine` | — |
| S1 | sitting ends and deliveries are stored; the views read them | the three computing view bodies | the next |
| S2 | failed and agent sittings are statements | `FailedSittingsOf(win)`, `AgentSittingsOf` | — |
| S3 | sat-through is a statement | the `sittings` field, `AgentSittingRow`, `recordsql.AgentSittings` | — |
| S4 | deliveries and the pin are statements; ruling 3 | `servedOf`, `pinOf`, `deliveriesUpTo`, `delivery`, the `deliveries` field, `viewsRead`, `mustViews`, `readViews`, `ReportDeliveries` | — |

T1 does not depend on the table and ships first. S1's carriers: `requirements.json`, `recordpb/schema_gen.go`, `recordsql/testdata/schema.epoch.json`, `schema.sql`, `cmd/feov-migrate/testdata/help.golden`, `record/testdata/plans.golden`, `record/testdata/shapes.golden`; and the statements of the stored-projection roster below, which no gate reads. No pull request leaves a dual rule: S1 deletes the view bodies with the tables' arrival, and each of S2–S4 deletes the fold it replaces.

**The stored-projection roster's statements.** `sitting_state` and `seat_delivery` join `demand_lineage` and `gap_state`. Swept on origin/main 5f0911494, tests and `testdata/` apart: `git grep -nIiE 'stored projection|gap_state.{0,80}demand_lineage|demand_lineage.{0,80}gap_state'` and `git grep -nI 'gap_dirty\|gap_flush\|GapStateDDL\|DemandLineageDDL'`. Property: every statement of the roster is a row here, with its pull request.

| Statement | Change | PR |
|---|---|---|
| `plans/record-sqlite.md:207-219` — the pattern's home (the section opens at `:192`), cited by `recordsql/gapstate.go:14` and `demandlineage.go:20` | three edits. `:211` calls `gap_state` "the one that is UPDATEd": `sitting_state` is replaced in place too (`INSERT OR REPLACE`), so the sentence names both. `:212-216`, the two stages, gains the contrast: `sitting_state` refreshes in ONE stage, inside the marked table's own trigger through `sitting_dirty`, with no flush row, because every insert on its marked tables changes sitting state; and `seat_delivery` is written once at its `report_read`, as `demand_lineage` is at its mint. `:218-219`, the append-only guards "on `events`, `report_checkpoint` and `demand_lineage`", gains `seat_delivery` | S1 |
| `recordsql/stmtcache.go:33-37` — which trigger programs an insert's prepare compiles | gains the sitting refresh on `register`, `sitting_open`, `sitting_close` and `envelope_admitted`, and the `seat_delivery` insert on `report_read` | S1 |
| `recordsql/gapstate.go:14-31`, `demandlineage.go:14-21` | each states its own table and cites the home; unchanged. `sittingstate.go` carries the same citation and says why one stage | S1, the new file |
| `recordsql/schema.go:323-324` (the DDL assembly); `migrate/sqlitesource.go:262-272` (the tables a migration does not carry) | gain the new DDL (§III.1) and the skip arm (Migration, below) | S1 |
| `recordsql/views.go:347, :359-360` | name the table each view reads, not the roster; unchanged | — |

**Pull request bodies.** Each carries a "Plan deviations" section, "none" where there is none. S1's names the point-update fallback if §V 1 forced it (§III.1). S3's states that ruling 1 is partly met for `satThroughOf`: the not-sat-through set is read once and compared in Go (§III.3). S2, S3 and S4 each carry §V 8's command and its result.

**Migration.** The replay writes every event through `InsertTx` (`record/replaywriter.go:84` → `record.go:824`), so the triggers fill the tables; `migrate/sqlitesource.go:262-272` gains an arm skipping the three stored tables, without which a source holding their rows is refused (`:284-293`). Residue: a migrated spot-check's pins are restamped by the live rule (`migrate/entries.go:54-61`), so after S4 one can differ from its source where a no-sitting delivery exists.

**Safety net (a174f37df).** The seeded fixtures (790 lines, 15 records) land with S1, so they stand under the rule while S1 replaces its definition; ruling 3 then shows as a reviewed golden diff in S4. The 6,140-line archive golden stays on the working branch through S4 and is then deleted (Q5): ~77 s under the race detector in a package over its nightly bound, and prose that churns on other streams' wording. On main the archive check is the oracle (§V 3), facts only. T1 deletes the lines the working-branch renderer reads, so that test and both goldens are updated in T1's step (§III.4, T1); from S1 the working branch holds the migrate half alone, and §V 8 gates S2, S3 and S4 on it.

## IV. Risk & Mitigation

| Risk | L × I × cost | Mitigation |
|---|---|---|
| A trigger fails to name a sitting an append changes | M × H × L | the oracle after every append, on archives and in the fuzz; the narrowing mutations (§V 3-4) |
| An opening of a third type (the write path makes none) | L × H × L | the CHECK on `events`, S1's DDL (§III.1): an own-id `sitting_id` only on `register` / `sitting_open` |
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
- `satThroughOf` is answered by the not-sat-through set, read once and compared in Go: ruling 1 partly met, stated in §III.3 and in S3's pull request body — this plan's; a fork the round-3 audit lists, decided, not open.
- Over criterion 2 the fallback is point updates, with `sitting_refresh` kept as the oracle's rule (§III.1) — this plan's; the same.
- The CHECK on `events` is S1's DDL (§III.1) — this plan's.

### Open question for the owner

| # | Decision | What happens unless he says otherwise |
|---|---|---|
| Q5 | The 6,140-line archive golden of a174f37df | it stays on the working branch through S4 and is then deleted; it never lands |

### UNVERIFIED

- That no agent index exists on `sitting_close` / `envelope_admitted` (a grep of `recordsql/*.go` found none).
- Whether planguard plans trigger-body statements, and whether it accepts a `SCAN … USING INDEX` (`plans.golden:131, :875` hold the `gap_refresh` view and the flush insert).
- The trigger's per-append cost: a target until S1's first commit measures it.
- The stored-state timings on a record with agentless openings, no-sitting deliveries, unsat sittings or an upheld position (neither measured record has one).
- The "today" timings on EventSchema 84 other than `agent_sitting`'s full load (§III.2); the race-nightly baseline after 6065a1a9e.
- That no test fixture writing `events` outside `InsertTx` (eleven test files hold such an insert) carries an own-id `sitting_id` on a third type; the CHECK's `type` spellings as `events` stores them.
- That `Motion.ruled` (`record/motion.go:243-245`) is the ruling's event id in every loader; whether a `means` annotation edit moves the epoch.
- How migrate registers an epoch step with no body translation.
- Test file line numbers in §V 4's existing list are the #1955 design pass's, not re-read.

## V. Verification Plan

First: `unset $(env | grep -o '^FEOV_[A-Z_]*')`. `FEOV_RELEASE_GATE` is never set locally.

1. **Timing.** The read side is measured (§III.2). S1's first commit measures the write side: an append to each marked table with and without its trigger, arms interleaved, setup outside the timer, against the 0.3 ms criterion; and repeats the read table against the real tables. The write side is TWO numbers per marked table, both required, and criterion 2 is judged on both: a WARM insert (the statement already prepared: the trigger's run) and a COLD prepare (the process's first insert on that table). A trigger's program compiles into its insert once per process (`recordsql/stmtcache.go:33-37`; the measured reason `gap_state` took two stages, `gapstate.go:20-25`), so a replay pays the compile once and a process that writes one act pays it on that act. A register that JOINS its agent's bracket is timed in both arms too: its trigger's WHEN is false and its insert still compiles the program.
2. **Equivalence before the views change (real data).** S1's first commit carries `TestStoredSittingStateEqualsTheViewsOnEveryArchivedRun` over `archiveNames` (`record/backingarm_test.go:304`): `EXCEPT` both ways for `agent_sitting`, `report_delivery` (`whole` included) and `area_change`, with non-vacuity counts (130 still open, 423 deliveries, 154 whole — #1955). The count is 39: `archiveNames` globs `run-archive/*.tar.gz`, and origin/main 5f0911494 holds 39. `run-archive/bundle-store/` holds two more, which `migrate.ArchivedRuns` walks as well: the 41 of #1955's design pass and of a174f37df's sweep (§V 8). The directory's 40 top-level `.gz` files are those 39 and `2026-08-28_checkpoint-seals-baseline.jsonl.gz`, which is no run. It is CI-green as the PR head before the commit that repoints the views; the PR body names that run. The repointing commit deletes the test — once the views read the tables it compares them with themselves — and the oracle (3) takes over.
3. **Permanent oracle.** `SittingStateRefreshDiff` names every sitting from the `sittings` view — never from `sitting_state`, where a missing row would hide itself — refreshes, and compares; and it recomputes every `seat_delivery` row from `report_read` and `events`, `whole` by the fragment asked of the row's own id. Asserted after EACH append in the seeded tests (an end-state check misses a row stale mid-run), on the archive sweep, in `record/plansweep_test.go:52`, and in the release fuzz beside the gap-state oracle.
4. **Seeded trigger tests `[NEW]`** (`recordsql/sittingstate_test.go`): an agentless opening ends the previous sitting; a latest sitting keeps a NULL next opening; closed then admitted after a nudge; one agent with two openings on one seat; **one agent opening on TWO seats** (the case that separates "every row of the agent" from "the agent's highest row"); a close after the next opening is not sat through; a report read in no sitting; a second report read in a sitting is refused. Mutations, each expected red: delete each marker row; **narrow each naming set** (the close's to the agent's highest row; the opening's to the new row alone). In package `record`, with S4: **an upheld position whose lens has a delivery complete after the ruling in a sitting opened before it — the blocker stands — then a sat-through delivery in a sitting opened after it — the blocker clears**; mutation: bound the statement by the delivery's id instead of its sitting's, expect red. Existing (files confirmed): `satthrough_test.go`, `lensturn_test.go`, `deliveryviews_test.go`, `upheldwhole_test.go`, `behinddelivery_test.go`, `spotcheckborne_test.go`, `debateserved_test.go`.
5. `go -C scripts run ./check`; `TestEveryDDLChangeBumpsTheEpoch` with the restamped epoch golden; `shapes.golden` gains one line, `sitting_refresh`, under `PLANGUARD_ACCEPT_VIEW_COST=1` (reason for the commit: it walks `sitting_dirty`, which holds one append's sittings), and loses the three computing views' cost.
6. After text changes (T1, S4): `agentgen`, `enginegen`, `cmdtreegen`, `protogen`, `vocabgen`, `vocabdoc`, each gate's `-check`; read every changed golden line and each affected `--help` before and after.
7. CI on the final head: the four `release-fuzz` parts and `release-fuzz-tally`; race-nightly's `record/migrate` time against 1,742 s.
8. **The archive golden gates S2, S3 and S4 (real data).** a174f37df's archive golden — `record/migrate/sittingdeliveryanswers_test.go` with `testdata/archived_sitting_delivery_answers.golden` and its renderer, over the 41 archives `migrate.ArchivedRuns` walks — lives on the local working branch `feov/sitting-delivery-statements` and never on main (Q5's default); the seeded fixtures of the same commit land with S1. Before S2, S3 or S4 merges, the working branch is rebased onto that pull request's head and, with every `FEOV_*` unset and no `UPDATE_GOLDENS`, this runs:
   `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestEveryArchivedRunAnswersItsSittingAndDeliveryQuestionsAsItsGoldenSays|TestTheSittingAndDeliveryRendererIsOneText' ./internal/record/migrate/`
   It passes with NO DIFF against the golden as T1's step left it (§III.4), or the pull request does not merge. S4 alone may differ, and only as ruling 3 predicts: no archive holds a delivery in no sitting (0 of 423, a174f37df's totals), so no ANSWER changes on the archives and the diff is at most the sentences §III.4's sweep rewords, where the golden prints one; it is regenerated with `UPDATE_GOLDENS=1` and read line by line, and a changed answer is a defect. Ruling 3's changed answer shows on the seeded golden's no-sitting record, on main since S1. The pull request body carries the exact command, the head it ran on, the archive count the test logs and the result, and for S4 each kind of changed line. When S4 merges, the working branch's golden and test are deleted.
9. Auditor gate: `/prosthetic-conscience:plan-audit`.
