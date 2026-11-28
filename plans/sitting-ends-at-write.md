# Sitting ends stored at the write

Issue #1955. Written 2026-11-28 against main eb994eb27; revised after audit rounds 1 and 2 and the owner's rulings of the same day. EventSchema 84 (`next_opening` bound once) has since merged as 6065a1a9e; this plan's schema change takes **the next epoch** after whatever main then holds. The measurements, the archaeology and the census cited as "#1955" are in that issue's comments. Paths are under `plugins/frank-exchange-of-views/tools/internal/`; `views.go` lines are as of eb994eb27 and moving. UNVERIFIED marks what was not read or measured.

## I. Summary & Goals

**Objective.** Every whole-record read computes `agent_sitting` and `report_delivery` row by row (`record/recordsql/views.go:654-661`), loads both into `WindowIndex` (`record/windowindex.go:48-58`) and folds them in Go. Rulings of 2026-11-28: store the ends at the write; each reader asks its own statement; lens pin/head numbers leave the work list; the view's rule for no-sitting deliveries survives.

**Success criteria.**
- Each single-question statement of §III.3 ≤ 0.2 ms on eiffel 11-11 / mersenne 10-30 (modernc): measured 0.007–0.08 ms against stored tables (§III.2), where today's take 0.9–9.9 ms. The whole `area_change` view, every lens's bound at once, ≤ 0.3 ms: measured 0.16 / 0.17 ms at load 15 and 0.26 / 0.17 ms at load 76, where today's takes 3.9–7.1 ms.
- An append to a marked body table costs ≤ 0.3 ms more; any other append, nothing. Measured in S1's first commit (§V 1).
- Race-nightly `record/migrate`: the views' 356 s of 1,742 s (#1955) falls below 60 s.
- Stored state equals today's views on every archived run (39 at origin/main 5f0911494) before any view body changes.
- `readViews`, `mustViews`, `viewsRead`, `AgentSittingRow`, `ReportDeliveries`: zero hits; `WindowIndex` has no `sittings` or `deliveries` field.
- No lens work-list surface carries a pin or head number (§III.4).
- A whole re-read marks every block that reads differently from the lens's last delivery: zero unmarked on every archived re-read (§III.4, measured).

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
| `delivery`, `DeliveryLine` `viewjson.go:1951, :1970` | `view/view.go:690`, `report/assemble.go:1307` | line deleted; `served` reads `report_delivery` | `debateView`'s querier |

`WindowIndex` keeps `of`, `latest`, `reach`. The report-read arm at `queries.go:163-167` and the loads at `replay.go:99, :110` go.

**Stated behaviour change (ruling 3).** A report read in no sitting counts as sat through and can set a pin for every reader; `servedOf` drops it today (`reportread.go:28-30`). The upheld-position predicate is the one reader it does not reach.

### 4. Agent-facing text

**Property.** No lens work-list surface carries a pin or head number: not the `last_read` object, not an item's subject id, not an item's sentence, not a delivered piece. Survivors outside the work list are rows below with their reason. `TestALensWorkListCarriesNoPinOrHeadNumber` (§V 5) holds it.

Census, re-run after audit round 1 over the plugin and the generators' sources, the generated copies, tests and goldens apart — 145 hits, every one in a row below (`views.go` hits renumbered to eb994eb27, eight lines above the working tree the grep ran in):

```
grep -rnE 'your pin|last_read|SINCE YOUR PIN|since the pin|its pin|pin [0-9%]|\(pin |head [0-9%$]|head %d|at head|report head|SubjectReport' \
  plugins/frank-exchange-of-views scripts/agentgen/src \
  --include='*.go' --include='*.md' --include='*.js' --include='*.mjs' --include='*.sh' --include='*.proto' --include='*.json' \
| grep -vE '_test\.go:|/testdata/|_gen\.go:|\.pb\.go:|/agents(-claude)?/|frank-exchange-of-views/skills/(adversarial-audit|research-protocol)/SKILL\.md:|debate\.test\.mjs:'
```

| Carrier | Change, or reason to survive |
|---|---|
| `record/sitting.go:105, :194, :197`; `LastReadJSON` `retirement.go:490-494`; `sittingwrite/delivery.go:75` | T2: `last_read` drops `pin` and `head`, keeps `kind`; the two sentences lose their numbers |
| `record/sitting.go:178`; `record/deliveredtext.go:248, :250`; `record/itemkind.go:92, :93, :115` | T2: a report item's subject is `{kind: report}` with no id (`about` allows it, `itemkind.go:114-118`); the comment "its id is the report head" goes |
| `record/sitting.go:276, :278` | T2: "dispatched against report head %d" loses the number, for every seat it reaches |
| `debate.js:886` (`lastReadClause`, authored: "at head H") | T2: reworded to the kinds alone |
| outside the regex: `sittingwrite/delivery.go:348` (a note line prints `· report <id>`), `record/itemkind.go:109` (`"id"` with no `omitempty`) | T2: a subject with no id prints its kind alone and its JSON omits `id`; left alone, T2 would print `· report ` and `"id":""` |
| `record/deliverycut.go:17, :25, :27, :88, :256`; `record/deliveredtext.go:215, :283, :285`; `cli/seat/help/register.md:11`; `scripts/agentgen/src/skills/adversarial-audit/SKILL.md:26`, `skills/research-protocol/SKILL.md:46` | T2: each describes what a lens is delivered and gains the marked re-read, below |
| `recordpb/record.proto:313` (`EVENT_TYPE_REPORT_READ`'s `means`), `terms/terms.json:862` ("report read"), `record/reportread.go:7`, `recordsql/views.go:672` | S4 (the view's comment in S1): each defines the pin as "in a sitting it sat through", which ruling 3 widens to "or in no sitting". Regenerate: `protogen`, `vocabgen`, `vocabdoc` (`docs/vocabulary.md:511, :517, :523`) |
| `### REPORT DELIVERED` `record/viewjson.go:1978, :1984` | T1: delete, with `report/reportread_test.go` |
| the word, no number, on a seat-facing surface: `cli/seat/verbs.go:330, :331, :369`; `cli/seat/help/{dispatch.md:13, spot-check.md:7, engage.md:7}`; `scripts/agentgen/src/{fragments/lens-duties.md:5, red-lens-voice.md:15, red-chair.md:12, :14}`; `debate.js:899, :900, :902`; `record/queue.go:48, :49`; `record/itemkind.go:22, :65`; `view/changes.go:21, :55`; `terms/terms.json:852, :873`; `hooks/README.md:115`; `docs/seat-command-triggers.md:75`; `docs/citations-and-bibliography.md:162`; `skills/research-protocol/references/report_template.md:8` | survive: ruling 2 removes the lens's numbers, not the word |
| `cli/seat/verbs.go:362` (`show report --json` `head`) | survives: a selector read's own stamp, outside the work list; it goes with the chair's numbers to #1962 |
| the plan's `head`: `record/dispatch.go:71`; `debate.js:403, :968, :988, :1118` | survives: ruled; the engine reads it |
| the chair's numbers: `cli/chair/dispatch.go:77, :90`; `cli/chair/spot_check.go:51`; `record/passblockers.go:479, :488, :493`; `record/dispatch.go:21, :105, :1329`; `record/retirement.go:99, :101, :103, :105, :127, :129, :138`; `record/spotcheck.go:260`; `debate.js:394, :409, :920` (and `:428`, `tests/simulator/testdata/{m3-chair3,o3-chair5}-relays.json`, outside the grep) | kept in this work (ruled 2026-11-28); tracked as #1962 |
| another sense — a dispatch's pin, a cite's git pin, a hash pin, a false match on "that heading" — or a report subject with no id: `record/record.go:1478`; `cli/blue/engage.go:56`; `cli/chair/dispatch.go:46`; `recordpb/record.proto:306, :2527, :2531`; `record/successor.go:256`; `record/migrate/replay.go:235, :566`; `setup/setup.go:222`; `setup/run.go:202`; `skills/research/SKILL.md:29`; `native/build.sh:107, :167`; `record/passage.go:58`; `report/layout.go:27`; `record/passblockers.go:460` | not carriers |
| reaches no seat — code comments on the report head or the pin, the tool's own refusals and test helpers: `cli/chair/dispatch.go:15`; `cli/seat/selector.go:261`; `record/anchorchecks.go:77`; `record/cast.go:21`; `record/dispatch.go:17, :201, :543, :957`; `record/passblockers.go:83`; `record/reportjson.go:10, :52`; `record/retirement.go:92, :111, :134, :264, :273, :355, :356, :420`; `record/sitting.go:111`; `record/spotcheck.go:45, :84, :236`; `record/viewjson.go:821, :1944, :1967`; `recordsql/views.go:682, :734, :737, :738, :744`; `recordpb/record.proto:1168, :2600`; `record/migrate/planids.go:20-25, :39`; `record/migrate/replay.go:577`; `record/deliveredtext.go:22, :179`; `record/deliverytest/deliverytest.go:41`; `seatprobe/boards.go:959` | survive; comments that restate a changed rule are reworded with it |

Generated copies follow their sources: `agents/` and `agents-claude/` (12 files each; `scripts/agentgen/main.go:70`) and the two preloaded skills by `agentgen`; `cli/seat/help_gen.go` by `cmdtreegen`. Goldens — `difftest/testdata` (15 files), `tests/simulator/testdata` (4), `record/migrate/testdata` (1) — and `tests/simulator/debate.test.mjs` (8 hits) are regenerated or edited, every changed line read.

**Ruling 3's sweep.** `git grep -nE 'sat through|sit through'` over the same scope: 82 hits, 32 on a line that names a pin or a delivery; the other 50 speak of sittings sat through and were classed by that filter, not read one by one.

| Hits | Change in S4 |
|---|---|
| the pin defined, third person: `terms/terms.json:862, :873`; `recordpb/record.proto:313`; `cli/seat/help/dispatch.md:13`; `record/dispatch.go:21` (→ `debate.js:394`); `record/passblockers.go:479`; `scripts/agentgen/src/red-chair.md:14`; `docs/seat-command-triggers.md:75` | each gains the clause "a delivery outside any sitting counts as sat through"; regenerate (`vocabgen`, `vocabdoc` for `docs/vocabulary.md:517, :523`; `enginegen`; `cmdtreegen`; `agentgen`); `tests/simulator/debate.test.mjs:574, :1478` follow |
| second person, about the reader's own sitting: `scripts/agentgen/src/fragments/lens-duties.md:5`; `record/reportread.go:13` (`ServedPinNote`); `record/sitting.go:194` (rewritten by T2) | survive: a seat reading them is in a sitting, where the sentence is exact |
| comments that state the rule: `record/deliveredtext.go:54, :67, :70`; `record/reportread.go:5, :42`; `record/retirement.go:68, :501`; `record/viewjson.go:1950`; `record/dispatch.go:310`; `record/passblockers.go:883`; `recordpb/record.proto:1178, :2593` | reworded with the code they describe, or deleted with their fold |
| another subject: `capture/sittingcloses.go:15, :47` (closes), `record/positionheard.go:251` (an instruction), `record/retirement.go:136` (a zero pin, which still means no sat-through delivery of either kind) | survive |

**T1, the whole-report debt once (ruled 2026-11-28).** Golden lines :123 and :124 of `difftest/testdata/docket_ruling_requires_each_unconditional_field.golden` are both why-lines of the STORED plan — the readiness line (`dispatch.go:651`) and the plan-hold reason (`passblockers.go:615`) — and :131 is the standing's gate item (`passblockers.go:602`). The readiness line survives; the plan-hold reason and the standing item go. Deleting a why arm changes the bytes, and so the key, of plans stored from then on. It is safe: archived plans keep both lines, `migrate/planids.go:36` matches id phrases and these lines carry no id, and no non-test reader matches either phrase. Consumer census (`grep -rn 'kindOf(\|WorkItem()\|\.why\b' --include=*.go . | grep -v _test`): a kind's `why` is rendered in one place, `record/dispatch.go:711`, the plan's why-lines; its `item`, through `WorkItem()` (`passblockers.go:742`), at `record/sitting.go:375, :392, :470, :495, :503, :520, :524` and `cli/lens/verify.go:413, :415`. For the whole-report kind only `sitting.go:392` applies: the chair's list takes every gate blocker (`:380-394`), while the blue, lens and `verify` arms take other kinds or an anchor's subject. So the item that goes is the chair's work-list line wherever that list prints — under the plan, after each act, in `show work`. The chair's list then no longer holds everything the gate refuses on, as `sitting.go:381-384` says it does: the PASS is still refused by the kind's `refusal`, and the plan's readiness line and `pass_permitted` are how the chair learns of the debt. The class — every blocker that readies its owner prints these three lines, the question-split debt among them (golden :122, :125, :132) — is #1963, outside this plan.

**T2, the marked re-read.** The owner asked (2026-11-28) why a change could ever fail to be placed, when and how often, and whether there are other answers.

*It can always be placed.* The round-1 design SEARCHED the final text for an edit's stamped new side, and a search fails when that text has moved on. The replay never has to search for what it wrote: `reportproj.Replay` holds the running text before and after every op (`reportproj/render.go:324-346`, "each mutation reads only the running text"); an edit lands where `bluedoc.LocateUniqueReplacing` finds its old side in that running text (`:106-122`), unique by the write's own validation; and `ApplySplice` replaces `report[start:end]` (`:48-56`). It is the carrying that keeps anchor tokens in place through edits (`keepUncarried`, `:59-66`). So the changed text can be carried through the fold, as the owner said.

*When the search fails, and how often.* Measured on the 39 archived runs (`archiveNames`; origin/main 5f0911494; load 2), by a scratch fold written for this plan — `ReplayTraced` does not exist yet, and every figure for the carried design below is that fold's. Nine runs hold lens deliveries: 63 whole re-reads over 10 distinct windows — a window is one record's head moves between a lens's pin and its re-read's head, and the lenses of a run usually share it — holding 220 distinct head moves and, in the 10 re-read texts, 797 blocks.

By change — can a search find the edit's stamped new side in the text of the re-read?

| Of 220 head moves | Raw text search | Search by blocks of visible text (round 1) |
|---|---|---|
| found once | 162 (73.6%) | 187 (85.0%) |
| not found | 57 (25.9%) | 32 (14.5%) |
| found more than once | 0 | 0 |
| a pure cut, nothing to search for | 1 (0.5%) | 1 (0.5%) |
| re-reads holding a change it cannot find | 63 of 63 | 56 of 63 |

Carried through the replay, 209 of the 220 (95.0%) still have text of their own standing at the re-read; the other 11 — 10 whose words a later edit overwrote or cut, and the pure cut — have none, and no design has anything of theirs to place: they show as the block they touched, or as their successor's text. So of the block search's 32 misses at least 22 are changes whose text still stands in part, in a paragraph edited again.

By block — the measure that compares the designs: is every block that reads differently from the last-delivered text marked?

| Design | Blocks marked, of 797 | Changed blocks left unmarked | Marked blocks that read as before |
|---|---|---|---|
| carried through the replay | 234 (29%) | 0 | 5 |
| stamped-block match: mark a block whose visible text is a block of any later edit's stamped new side — no search of a span, no replay | 237 (30%) | 1, in 7 of the 63 re-reads | 9 |

Over the 47 distinct change-delivery windows (3,644 blocks) the same two leave 0 and 3 unmarked. The stamped-block match's misses are a table row and a list item — a stamped side read out of its table or list parses into different blocks — and a paragraph edited twice. Widening passages to blank lines (below) will mark more than 29%; by how much is not measured.

The search fails whenever blue edits a paragraph twice between two deliveries to a lens, and a raw search fails as well whenever a cite, a finding or a gap later puts an anchor token into an edited paragraph (25 of the 57). It never failed because text stood twice, and once because of a cut. `deliverycut.go:271`'s anchor-only filter fired on none of the 1,227 (head move, re-read) pairs. Seven older archives (2026-08-22 to 2026-09-02) hold ops that do not replay from the first ("no part of the report is recorded yet") and were not measured.

*Other answers considered.*

| Answer | Verdict |
|---|---|
| Text search, the misses appended as BEFORE/AFTER items (round 1) | rejected: it cannot find 14.5% of changes, in 89% of re-reads, and a superseded edit's item shows text the report no longer holds |
| A rendered diff of two texts, the last-delivered and the present | rejected: it needs the last-delivered text kept or rebuilt, and a diff guesses its alignment where paragraphs repeat or move; the replay knows |
| Stamped-block match | not chosen, and the cheapest: it is the mechanism `recordsql/reportcheckpoint.go:46-48` names ("what an edit changed is on its row"), with no replay and nothing stored. But it left a changed block unmarked in 7 of 63 re-reads, it cannot know that it did, and it cannot show a removal. It is Q8's third option |
| Carried through the replay, run at the cut | **chosen**: exact by construction; no schema, no cost on any write; needed a few times per lens per run |
| Carried through the replay at every write, each run's writer stored beside the checkpoint | Q8's second option: no replay at the cut, at the price of an epoch for T2 and work on every op |

*Design.*

- `reportproj.ReplayTraced` `[NEW]` (name grepped: zero hits; "provenance" is taken, for sources) folds the ops as `Replay` does and returns the text with its RUNS — `(start, end, writer)`, the writer being the event id of the head move that last wrote those characters — and its cut points. After each op it takes the smallest splice that turns the text before into the text after (their common prefix and suffix), so seam tidying, husk removal, newline collapse and a part's joining (`:53-54, :215-292, :372-377`) are carried without any mutation reporting an offset.
- *Carrying.* A run before the splice stays; one after it shifts by the change in length; one inside the replaced stretch is absorbed; one across an edge is cut back to the edge. A head-move op — an `edit`, an `append`, and the `remove` rows of an edit's own event — then writes its new stretch as a run under its id. Nothing merges in the fold: runs that abut or overlap in the final text fall into one passage.
- *A cut.* Where a head move removes text inside a block, what is left of that block is touched on the cut's side and so marked: the paragraph reads differently. Where it removes whole blocks, a point stays between their neighbours, carried as a run is, and shows as one line. Recommended over showing nothing (Q9): a removal that leaves no trace is a change that reads as "unchanged".
- *Anchor-only ops* — an `insert`, a `sentence_of`, a retire's `remove` — move and split runs as any splice does and write none. A token put into marked text lies inside the mark; one put into unmarked text marks nothing, which is why `HeadMoves` leaves them out (`recordsql/views.go:734-742`).
- *The cut.* In `CutDelivery`'s whole branch (`deliverycut.go:66-72`), for a lens whose pin is above zero — its full re-read, or a whole delivery an upheld position owes — the cut reads every op (`recordsql.ReportOpsBetween(q, 0, Newest)`) and traces it. A PASSAGE is the smallest stretch of whole blocks (`anchor.Blocks`), widened outward to blank lines so that no mark lands inside a table or a tight list, that holds a run whose writer is after the pin, or a cut point inside a block. The header's "never located in the report" (`:19-21`) is reworded: nothing is searched for.
- *What the lens sees.* Mark lines, each a paragraph of its own, with no number and no event id:

  ```
  [CHANGED SINCE YOUR LAST DELIVERY — the text from here to the closing mark is new or edited]
  …the passage as it stands…
  [END OF CHANGED TEXT]

  [TEXT WAS REMOVED HERE SINCE YOUR LAST DELIVERY]
  ```

  A first whole delivery is labelled `THE WHOLE REPORT`, as today (`:115`), and carries no mark. A marked one is labelled `THE WHOLE REPORT — N passage(s) changed since your last delivery are marked, and R place(s) where text was removed`; `reportItemPreamble` prints the label on every item (`reporttext.go:57-65`). No BEFORE/AFTER item follows: nothing is left unplaced.
- *Pieces.* A mark line is a block, so `splitFit` may cut on either side of it (`:158, :207-215`), and a passage may cross a `\n## ` section cut (`:117-125`). On the pieces AS HANDED: (1) every opening mark is closed by an end mark in the same or a later piece, with no opening between; (2) a piece whose first report text lies inside a passage begins with `[CHANGED TEXT CONTINUES]`, and no other piece does; (3) the report blocks between marks are exactly the passages; (4) the pieces with mark lines removed, concatenated, are the report (today's invariant at `:112-113`, restated). Marks, the continuation line and the label are inside what `fits` is asked of (`packWhole` asks `fits(label, cur+u)`, `:134-147`; `itemFits` renders the label into the preamble, `reporttext.go:74-76`), so nothing is added after the budget check `handReportText` repeats (`:98-100`).

*What can fail, and what holds it.* Round 1's count check was true by construction and is gone. Three checks replace it, each setting two differently-made things against each other:

1. In the cut: the traced text equals `RenderAt`'s, the checkpoint and its tail, and the runs tile it — every character in exactly one run, from the first part's on. This holds the text and the runs' coverage, not their writers; check 3 holds those. A fold that drifts refuses the delivery.
2. In the cut: the pieces, read back, satisfy (1)–(4) against the passages. A packer that drops, delays or misplaces a mark line refuses the delivery (a stuck delivery and a tool-defect log, `reporttext.go:83-85, :114-121`).
3. In tests, on seeded records and every archive: every block of the cut text whose visible text is no block of the last-delivered text lies in a passage. This ran in the measurement above, on the scratch fold: 0 unmarked of 4,004 blocks over the 63 re-reads (797 distinct), and 0 of 14,342 over the 194 change deliveries — after its first run found 13, a cut at the head of a paragraph that the scratch fold had left as a bare point. The cut rule above is the fix, and the 13 are the evidence the check can fail. The other direction is loose by design: 5 of the 234 distinct marked blocks (2%) read the same as a last-delivered block, where an edit left a block's visible text as it was.

Check 3 needs the last-delivered text, which only a replay to a past event gives. This plan reads the standing ruling — "this plan asks nothing as of a past event" — as binding the rules a plan ships, not its tests; and it reads check 1's from-the-first replay as within it, because the product is the present text with each run's writer and the mark is an id comparison, writer after pin. Both readings are the owner's to confirm (Q8).

*Cost.* A from-the-first replay took 19–630 ms on the nine records (169 ops over 16 kB: 79 ms; 254 ops over 80 kB: 630 ms), by a harness that replays op by op and keeps every text — an upper bound. It runs once per marked re-read, in the delivery writer's hook.

Lost with the numbers: a lens can no longer copy a head into `close --verified-against` (25 of 63 behind sittings did, unprompted — #1955).

Found on the way, outside this plan and to be filed: `headingOf`'s raw lookup (`deliverycut.go:281-283`) misses for 143 of 399 distinct changes (36%) across the 47 distinct change-delivery windows, and each miss silently drops that item's "In section:" line.

### 5. Pull requests

| PR | Concept | Deletes | Epoch |
|---|---|---|---|
| T1 | `dispatch next` states the whole-report debt once; `REPORT DELIVERED` goes | the plan-hold why arm and the standing item for this kind; `DeliveryLine` | — |
| T2 | the lens work list names no pin or head; the re-read marks what the replay carried | `LastReadJSON.Pin`, `.Head`; report-item subject ids | — |
| S1 | sitting ends and deliveries are stored; the views read them | the three computing view bodies | the next |
| S2 | failed and agent sittings are statements | `FailedSittingsOf(win)`, `AgentSittingsOf` | — |
| S3 | sat-through is a statement | the `sittings` field, `AgentSittingRow`, `recordsql.AgentSittings` | — |
| S4 | deliveries and the pin are statements; ruling 3 | `servedOf`, `pinOf`, `deliveriesUpTo`, `delivery`, the `deliveries` field, `viewsRead`, `mustViews`, `readViews`, `ReportDeliveries` | — |

T1 and T2 do not depend on the table and ship first. S1's carriers: `requirements.json`, `recordpb/schema_gen.go`, `recordsql/testdata/schema.epoch.json`, `schema.sql`, `cmd/feov-migrate/testdata/help.golden`, `record/testdata/plans.golden`, `record/testdata/shapes.golden`. No pull request leaves a dual rule: S1 deletes the view bodies with the tables' arrival, and each of S2–S4 deletes the fold it replaces.

**Migration.** The replay writes every event through `InsertTx` (`record/replaywriter.go:84` → `record.go:824`), so the triggers fill the tables; `migrate/sqlitesource.go:262-272` gains an arm skipping the three stored tables, without which a source holding their rows is refused (`:284-293`). Residue: a migrated spot-check's pins are restamped by the live rule (`migrate/entries.go:54-61`), so after S4 one can differ from its source where a no-sitting delivery exists.

**Safety net (a174f37df).** The seeded fixtures (790 lines, 15 records) land with S1, so they stand under the rule while S1 replaces its definition; ruling 3 then shows as a reviewed golden diff in S4. The 6,140-line archive golden stays on the working branch through S4 and is then deleted (Q5): ~77 s under the race detector in a package over its nightly bound, and prose that churns on other streams' wording. On main the archive check is the oracle (§V 3), facts only.

## IV. Risk & Mitigation

| Risk | L × I × cost | Mitigation |
|---|---|---|
| A trigger fails to name a sitting an append changes | M × H × L | the oracle after every append, on archives and in the fuzz; the narrowing mutations (§V 3-4) |
| An opening of a third type (the write path makes none) | L × H × L | CHECK on `events`: an own-id `sitting_id` only on `register` / `sitting_open` |
| Stored `whole` drifts from the rule | L × M × L | the oracle recomputes it (§V 3); the pre-change equivalence includes it |
| planguard does not plan a trigger's statements | M × M × L | a test issues `SELECT * FROM sitting_refresh` over seeded dirty rows |
| The trace misses a change, which reads as "unchanged" | L × H × L | the block comparison on seeded records and every archive (§V 5-6); zero unmarked measured |
| The packer drops, delays or misplaces a mark line | M × H × L | the cut reads its pieces back against rules (1)–(4) and refuses; three mutants (§V 5) |
| The from-the-first replay is slow on a large report, inside a hook | M × L × M | once per marked re-read; 630 ms worst measured; Q8's alternative removes it |
| Capture reads outside one snapshot (`replay.go:97-112`) | M × M × L | capture moves to `readSnapshot` in S2 |
| Golden conflicts across ~8 streams | H × L × L | rebase on golden or schema change; regenerate at `-count=1` |

### Decided

- The pin is derived, never stored — ruled by the owner 2026-11-28; measured in §III.2.
- The chair's pin and head numbers stay in this work — ruled 2026-11-28; tracked as #1962.
- The whole-report debt prints once, the plan's readiness line; the class is #1963 — ruled 2026-11-28.
- The triggers refresh through `sitting_dirty` (§III.1) — this plan's.
- `whole` is stored by the trigger at the insert (§III.1) — this plan's.
- The re-read marks what the replay carried; nothing is searched for and nothing is appended (§III.4) — this plan's answer to the owner's question, with Q8 and Q9 left to him.

### Open questions for the owner

| # | Decision | Recommendation |
|---|---|---|
| Q5 | The archive golden | working branch only, deleted after S4 |
| Q8 | How the re-read learns what changed. Two standing texts bear on it: constitution-gaps r8, "delivery facts are derived in recordsql views and prepared statements, never stored or replayed in Go" and "no part reads report text at a past event"; and `reportcheckpoint.go:46-48`, "No reader reads the report as it stood at a past event: what an edit changed is on its row". (a) Carried through a replay from the first op, at the cut: exact, but a replay in Go that passes through every past text. (b) Carried at every write and stored beside the checkpoint: exact, but stored, with an epoch for T2. (c) The stamped-block match: neither replayed nor stored, and one changed block in 797 unmarked on the archives, silently. And: may a TEST replay a record to a past event, as §V 5-6 do? | (a): it hands and decides on no past text — its product is the present text with each run's writer — and it costs nothing on a write. A test may |
| Q9 | A head move that removes whole blocks: one `[TEXT WAS REMOVED HERE …]` line, or nothing | the line |

### UNVERIFIED

- That no agent index exists on `sitting_close` / `envelope_admitted` (a grep of `recordsql/*.go` found none).
- Whether planguard plans trigger-body statements, and whether it accepts a `SCAN … USING INDEX` (`plans.golden:131, :875` hold the `gap_refresh` view and the flush insert).
- The trigger's per-append cost: a target until S1's first commit measures it.
- The stored-state timings on a record with agentless openings, no-sitting deliveries, unsat sittings or an upheld position (neither measured record has one).
- Whether dropping a report item's subject id changes `changed_since_delivered`: `itemID` marshals the whole item (`sittingwrite/delivery.go:393-396`) and is read at `:379-383`; where `changed` is computed was not found.
- That `anchortext.Visible` is the right reduction for the block comparison of check 3 (`anchortext/anchortext.go:82-86` says it is not right for every reader).
- Everything measured for the carried design is a scratch fold's (scratch `issue-1955/plan-rev2/`), not `ReplayTraced`'s; its cost figures are a per-op harness's.
- How much of a re-read lies in a passage once passages are widened to blank lines.
- The archive count: 39 at origin/main; #1955's design pass says 41.
- The `[TEXT WAS REMOVED HERE …]` line on real data: no archived re-read window removes whole blocks, so seeded tests alone exercise it.
- That the seven archives whose ops do not replay from the first fail `RenderAt` the same way, and hold no report read.
- The 50 `sat through` hits that name no pin or delivery: classed by a filter, not read.
- Whether any surface besides the standing renders `passblockers.go:602`'s item; how `difftest` declares a scenario that renders delivery items.
- That `Motion.ruled` (`record/motion.go:243-245`) is the ruling's event id in every loader; whether a `means` annotation edit moves the epoch.
- How migrate registers an epoch step with no body translation.
- Test file line numbers in §V 4's existing list are the #1955 design pass's, not re-read.

## V. Verification Plan

First: `unset $(env | grep -o '^FEOV_[A-Z_]*')`. `FEOV_RELEASE_GATE` is never set locally.

1. **Timing.** The read side is measured (§III.2). S1's first commit measures the write side: an append to each marked table with and without its trigger, arms interleaved, setup outside the timer, against the 0.3 ms criterion; and repeats the read table against the real tables.
2. **Equivalence before the views change (real data).** S1's first commit carries `TestStoredSittingStateEqualsTheViewsOnEveryArchivedRun` over `archiveNames` (`record/backingarm_test.go:304`): `EXCEPT` both ways for `agent_sitting`, `report_delivery` (`whole` included) and `area_change`, with non-vacuity counts (130 still open, 423 deliveries, 154 whole — #1955). It is CI-green as the PR head before the commit that repoints the views; the PR body names that run. The repointing commit deletes the test — once the views read the tables it compares them with themselves — and the oracle (3) takes over.
3. **Permanent oracle.** `SittingStateRefreshDiff` names every sitting from the `sittings` view — never from `sitting_state`, where a missing row would hide itself — refreshes, and compares; and it recomputes every `seat_delivery` row from `report_read` and `events`, `whole` by the fragment asked of the row's own id. Asserted after EACH append in the seeded tests (an end-state check misses a row stale mid-run), on the archive sweep, in `record/plansweep_test.go:52`, and in the release fuzz beside the gap-state oracle.
4. **Seeded trigger tests `[NEW]`** (`recordsql/sittingstate_test.go`): an agentless opening ends the previous sitting; a latest sitting keeps a NULL next opening; closed then admitted after a nudge; one agent with two openings on one seat; **one agent opening on TWO seats** (the case that separates "every row of the agent" from "the agent's highest row"); a close after the next opening is not sat through; a report read in no sitting; a second report read in a sitting is refused. Mutations, each expected red: delete each marker row; **narrow each naming set** (the close's to the agent's highest row; the opening's to the new row alone). In package `record`, with S4: **an upheld position whose lens has a delivery complete after the ruling in a sitting opened before it — the blocker stands — then a sat-through delivery in a sitting opened after it — the blocker clears**; mutation: bound the statement by the delivery's id instead of its sitting's, expect red. Existing (files confirmed): `satthrough_test.go`, `lensturn_test.go`, `deliveryviews_test.go`, `upheldwhole_test.go`, `behinddelivery_test.go`, `spotcheckborne_test.go`, `debateserved_test.go`.
5. **T2's tests `[NEW]`.**
   - `TestALensWorkListCarriesNoPinOrHeadNumber` (`sittingwrite`): renders a lens's work list — the JSON, each rendered item and the notes lines (`delivery.go:338-352`) — for every `last_read` kind, an owed register and a delivery of each sort. It asserts that `last_read` has no `pin` or `head` key, that no report item has a subject id (no `"id":""`, no `· report ` with a trailing space), and that no head-move event id of the seeded record appears as a token. The fixture pads the record so every head move's id is above 1,000 and asserts first that no count the list prints reaches that.
   - `TestTheReReadMarksWhatChangedSinceTheLastDelivery` (`record`): a lens with one sat-through whole delivery, then two edits to abutting sentences (one passage), an edit re-edited later (its successor's text marked, no item), a cut inside a paragraph (the paragraph marked), a cut of a whole paragraph (the removal line), a cite into an edited paragraph (still marked), a retire in untouched text (no mark), a passage crossing a `\n## ` heading, and a budget that splits a marked passage (the continuation line). It asserts rules (1)–(4) on the pieces and check 3 against a replay to the pin. Mutants, each red: omit a mark line; omit the continuation line; delay an end mark by a block; drop the cut-inside-a-block rule; let an anchor insert write a run.
   - A `difftest` golden `[NEW]` renders the first item of a first whole delivery and every item of a marked re-read, so both labels and all four mark lines are read in review.
6. **Real data for T2.** `TestTheTraceMarksEveryChangedBlockOnEveryArchivedRun` (`record/migrate`, over `archiveNames`): at every lens re-read, check 1 (traced text = render) and check 3 (zero unmarked blocks), with non-vacuity counts — 63 re-reads, 220 distinct head moves, and how many `deliverycut.go:271` filtered (none today). It replays to each pin, a past event, in a test (Q8 b). This plan's measurement is its first run.
7. `go -C scripts run ./check`; `TestEveryDDLChangeBumpsTheEpoch` with the restamped epoch golden; `shapes.golden` gains one line, `sitting_refresh`, under `PLANGUARD_ACCEPT_VIEW_COST=1` (reason for the commit: it walks `sitting_dirty`, which holds one append's sittings), and loses the three computing views' cost.
8. After text changes: `agentgen`, `enginegen`, `cmdtreegen`, `protogen`, `vocabgen`, `vocabdoc`, each gate's `-check`; read every changed golden line and each affected `--help` before and after.
9. CI on the final head: the four `release-fuzz` parts and `release-fuzz-tally`; race-nightly's `record/migrate` time against 1,742 s.
10. Auditor gate: `/prosthetic-conscience:plan-audit`.
