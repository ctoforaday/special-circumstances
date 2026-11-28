# A lens's whole-report re-read marks what changed

Issue #1955. Written 2026-11-28 against origin/main 5f0911494, split that day from `plans/sitting-ends-at-write.md`. Paths are under `plugins/frank-exchange-of-views/tools/internal/`. UNVERIFIED marks what was not read or measured.

**Rulings implemented** (the owner, 2026-11-28, as relayed by the coordinating session): (1) the lens's pin and head numbers leave its work list and `kind` stays; (2) the whole-report re-read marks inline what changed since the lens's last delivery, so no number is needed to find it; (3) seat-facing text names no event id; (4) a removed whole block shows one line at the cut, no numbers; (5) the re-read learns what changed from **spans stored at the write** — asked to choose between a replay at the delivery, spans stored at the write and a match on the edits' stamped blocks, the owner wrote "I feel like we already do things like 2, don't we?"; (6) the chair's pin and head numbers stay, #1962.

**What ruling 5 supersedes.** The constitution-gaps plan (r8, approved 2026-11-14) says "delivery facts are derived in recordsql views and prepared statements, never stored or replayed in Go" and "no part reads report text at a past event". The first is superseded for one fact: which head move last wrote each block of the report is stored. The second stands untouched: nothing is replayed at the delivery, and the storing happens where the write already holds the text. `recordsql/reportcheckpoint.go:46-48` — "No reader reads the report as it stood at a past event: what an edit changed is on its row" — stays true; the checkpoint gains a second record of the same kind.

**Relation to `plans/sitting-ends-at-write.md`.** Independent. This plan's delivery cut reads the pin through `PinAt` (`record/deliverycut.go:74`), which exists today; that plan re-points it at stored rows and, by its ruling 3, lets a delivery in no sitting set it. Neither needs the other landed. Each takes an epoch; whichever merges second regenerates its epoch carriers. Both touch `lastReadOf` (`record/retirement.go:504-535`): after this plan it computes `kind` alone.

## I. Summary & Goals

**Objective.** A lens owed its full re-read is delivered the whole report with nothing saying what changed (`record/deliverycut.go:66-72`), and its work list prints a pin and a head — event ids no lens verb takes (#1955: 63 behind sittings, none used the pin in a command). The numbers go, and the re-read marks the changed text.

**Success criteria.**
- No lens work-list surface carries a pin or head number (§III.3's property; `TestALensWorkListCarriesNoPinOrHeadNumber`).
- At every re-read on every archived run, no block that reads differently from the lens's last delivery is left unmarked. Measured for the stored rule by a scratch fold: 0 of 797 blocks over 63 re-reads (§III.4).
- The stored block writers equal a from-the-first traced replay after every append in the seeded tests, on every archived run and in every fuzz debate.
- A marked re-read's delivery folds no op that an unmarked read of the report does not already fold: at most `reportCheckpointEvery - 1` (7).
- No read of the report but the marked cut's reads a stored row. The stamp costs at most 2.5 times today's per op it folds. A scratch measurement over the 39 archives puts one block parse at about one replayed op (0.9–1.4 times), so the rule's one parse per op makes the stamp about 2 to 2.4 times today's 0.3–2.4 ms per op; R1's first commit measures the shipped fold (§V 1). Over the bound, R1 does not merge as specified: §IV names what replaces the advance.

**Value case.** A search for each edit's new text cannot find 14.5% of changes, in 89% of re-reads; a replay at the delivery finds all of them and costs 19–630 ms and a pass through every past text. Storing rides a fold the write already makes every eighth op and leaves the delivery a comparison of ids. The price is one table and a stricter fold, held by two oracles.

## II. Technical Context

- Go 1.27, modernc SQLite. The report is a left fold of `report_op` rows — a part's `append`, an edit's splice, an anchor's `insert` / `sentence_of`, a `remove` — each applied to the running text (`reportproj/render.go:12-37, :324-346`).
- **Where the write holds the text.** `InsertTx` ends with `stampReportCheckpoint` (`recordsql/store.go:925`). At the record's first op, and at each op that brings the ops since the checkpoint to `reportCheckpointEvery` (8, `recordsql/reportcheckpoint.go:12-20`), it reads the checkpoint's text and the ops since, folds them (`:165`) and replaces the one row (`:170-171`; DDL `recordsql/schema.go:196-221`). Every writer passes it: a seat's act, a seeded fixture, a migration's replay.
- A reader starts from that row and folds the fewer than eight ops after it (`ReportSourceAt`, `reportcheckpoint.go:62-91`; `renderSource`, `record/queries.go:255-293`). A missing checkpoint is an error, never a longer replay (`:50-54`).
- The edit's own write already records what it touched: `cli/blue/edit.go:188-196` cuts `old_paragraphs` / `new_paragraphs` (`bluedoc.TouchedParagraphs`) from the report the edit lands on. Their unit is the block of `anchor.Blocks` (`recordpb/record.proto:1787-1792`).
- A head move is a part or an edit (`HeadMoves`, `recordsql/views.go:747`); a retire and every anchor placement change tokens, not text a lens reads.
- A lens's delivery is cut once, at its first item (`record/deliverycut.go:14-21`): the whole report where `NextDeliveryWhole` says so (`:66-72`), packed by `packWhole` (`:112-150`) and handed item by item (`sittingwrite/reporttext.go:57-110`). The pin is `PinAt` (`deliverycut.go:74`).
- Constraints: no compatibility paths; a DDL change bumps the epoch; `feov-migrate` is the only upgrader; no shipped rule reads the record as of a past event; release fuzz in CI only.

## III. Proposed Changes

```
internal/reportproj/
  traced.go              [NEW]    ReplayTraced: the fold, carrying each block's writer
  traced_test.go         [NEW]
internal/record/recordsql/
  schema.go              [MODIFY] report_checkpoint_block; report_checkpoint.removed_at_end
  reportcheckpoint.go    [MODIFY] the stamp folds traced and writes the rows; ReportBlockWritersAt
  reportblocks_test.go   [NEW]    oracle A after each append
internal/record/
  deliverycut.go         [MODIFY] the marked whole delivery
  sitting.go, retirement.go, deliveredtext.go, itemkind.go  [MODIFY] no pin or head number
internal/record/migrate/sqlitesource.go [MODIFY] skips the stored rows
internal/sittingwrite/delivery.go       [MODIFY] the notes line, last_read
```

### 1. Block writers, stored at the write `[NEW]`

Names grepped across the plugin and `scripts/`: `report_checkpoint_block`, `report_block`, `block_writer`, `ReplayTraced` — zero hits each. "Provenance" is taken, for sources.

| Object | Shape |
|---|---|
| `report_checkpoint_block` `[NEW]` | `ord` INTEGER PK — the block's position among `anchor.Blocks` of the checkpoint's text; `writer` INTEGER NOT NULL REFERENCES `events(id)` — the latest head move that wrote or cut within the block; `removed_before` INTEGER REFERENCES `events(id)` — the latest head move that removed whole blocks just before it, NULL for none. Exactly one row per block. Replaced whole with the checkpoint row, in its transaction. |
| `report_checkpoint.removed_at_end` `[NEW]` column | INTEGER REFERENCES `events(id)`, NULL for none: the latest head move that removed whole blocks at the end of the text, where no block follows to carry it. |
| `reportproj.ReplayTraced(text, blocks, parted, ops)` `[NEW]` | `Replay` (`render.go:324-346`) that also returns the blocks' writers. `Replay` stays for every reader that wants text alone. |
| `recordsql.ReportBlockWritersAt(q)` `[NEW]` | the checkpoint's rows: `SELECT "ord", "writer", "removed_before" FROM "report_checkpoint_block" ORDER BY "ord"`, with `removed_at_end`. Two callers: the stamp, in its write transaction, and the marked cut, in the snapshot it already reads the report in (`readSnapshot`, `deliverycut.go:55`). `ReportSourceAt` and every board, work-list and report read are unchanged. |

**Per block, not per character run.** The re-read asks one thing: which blocks hold text written after this lens's last sat-through delivery. A row per block answers it by an id comparison; its count is bounded by the report, not by the number of edits; it needs no second encoding of text offsets; and it is refusable — rows that do not number exactly the blocks of the checkpoint's text are a broken checkpoint, and the cut and the stamp, the two readers, say so. Runs would be finer than any mark this plan draws.

**How it advances.** The fold takes each op's smallest splice — the common prefix and suffix of the text before and after it — so seam tidying, husk removal, newline collapse and a part's joining (`render.go:53-54, :215-292, :372-377`) need report no offset. Blocks are parsed once per op: an op's blocks after are the next op's blocks before. Then:

- a block of the text after the op **keeps its writer exactly when the same characters were one whole block before it, outside the splice**;
- every other block takes the op's event id where the op belongs to a head move. The fold tells that from the op stream alone, since `reportproj` imports nothing of the record's (`render.go:302-304`): an `append`, an `edit`, or a `remove` that shares its event id with an `edit` — a stamp's ops and a reader's tail both begin and end on an event (`ReportOpsBetween`, `reportcheckpoint.go:98-118`);
- under an anchor-only op (`insert`, `sentence_of`, a retire's `remove`) such a block inherits the latest writer among the blocks it came from: a token changes no text a lens reads;
- a head move that removes whole blocks and writes nothing sets `removed_before` on the block that follows, or `removed_at_end` where none does.

So a cut inside a paragraph, at its head or tail, changes that paragraph's row; a paragraph split or joined changes the rows it becomes; a paragraph a cite lands in keeps its writer. The unit is `bluedoc.TouchedParagraphs`' own; the rule differs from it where a cut leaves nothing new, and the block cut from must still show.

**Between checkpoints nothing is stored.** The stamp advances the rows through the ops it folds, as it advances the text (`reportcheckpoint.go:160-171`); the marked cut, the one reader that wants writers, folds the tail with `ReplayTraced`, the same fewer-than-eight ops it folds for the text. No fold is added anywhere: the two that exist carry one more thing.

**A failed checkpoint** (`schema.go:206-210`) holds no rows; the report is the failure, as today.

### 2. The marked re-read `[MODIFY]`

In `CutDelivery`'s whole branch (`deliverycut.go:66-72`), for a lens whose pin is above zero — its full re-read, or a whole delivery an upheld position owes — the cut reads the checkpoint, its rows and its tail in one snapshot, refuses rows that do not fit the checkpoint's text, and folds the tail traced: the present text and its blocks' writers.

- A PASSAGE is the smallest stretch of whole blocks, widened outward to blank lines so that no mark lands inside a table or a tight list, that holds a block whose `writer` is after the pin.
- What the lens sees — mark lines, each a paragraph of its own, with no number and no event id:

  ```
  [CHANGED SINCE YOUR LAST DELIVERY — the text from here to the closing mark is new or edited]
  …the passage as it stands…
  [END OF CHANGED TEXT]

  [TEXT WAS REMOVED HERE SINCE YOUR LAST DELIVERY]
  ```

  The last stands alone, before a block whose `removed_before` is after the pin, or at the end of the report where `removed_at_end` is (ruling 4).
- Labels. A first whole delivery is `THE WHOLE REPORT`, as today (`:115`), with no mark. A marked one is `THE WHOLE REPORT — N passage(s) changed since your last delivery are marked, and R place(s) where text was removed`; `reportItemPreamble` prints the label on every item (`reporttext.go:57-65`). No BEFORE/AFTER item follows.
- Pieces. A mark line is a block, so `splitFit` may cut on either side of it (`:158, :207-215`), and a passage may cross a `\n## ` section cut (`:117-125`). On the pieces AS HANDED: (1) every opening mark is closed by an end mark in the same or a later piece, with no opening between; (2) a piece whose first report text lies inside a passage begins with `[CHANGED TEXT CONTINUES]`, and no other piece does; (3) the report blocks between marks are exactly the passages; (4) the pieces with mark lines removed, concatenated, are the report (today's invariant at `:112-113`, restated). Marks, the continuation line and the label are inside what `fits` is asked of (`packWhole` asks `fits(label, cur+u)`, `:134-147`; `itemFits` renders the label into the preamble, `reporttext.go:74-76`), so nothing is added after the budget check `handReportText` repeats (`:98-100`).
- The cut reads its pieces back against (1)–(4) and refuses on a miss: a stuck delivery and a tool-defect log (`reporttext.go:83-85, :114-121`).
- The header's "never located in the report" (`:19-21`) stays true: nothing is searched for.

Consumer census — `grep -rn 'packWhole\|CutDelivery(\|reportproj\.Replay\b\|ReportSourceAt(\|reportSourceAt(\|\.Fold(\|renderSource(\|report_checkpoint' --include=*.go .` (run 2026-11-28):

| Consumer | Changes? |
|---|---|
| `recordsql/reportcheckpoint.go:133-175` (`stampReportCheckpoint`: reads `:143, :155`, folds `:165`, writes `:170`) | yes: the traced fold; writes the rows |
| `recordsql/reportcheckpoint.go:62-91` (`ReportSourceAt`, read `:74`) | no: text and tail alone, as today |
| `recordsql/schema.go:214` | yes: `removed_at_end`, and the new table beside it |
| `record/queries.go:255, :260-264, :269-273, :293` (`reportRenderer`, `renderSource`, `reportSourceAt`, `RenderAt`) | no: text alone |
| `record/viewjson.go:1149, :1179` | no |
| `record/deliverycut.go:71` (`packWhole`), `:52-110` (`CutDelivery`) | yes |
| `sittingwrite/reporttext.go:83` (`CutDelivery`'s one shipped caller) | no: it hands what it is cut |
| `record/migrate/order.go:70` (`reportproj.Replay` over ops as written) | no |
| `record/migrate/sqlitesource.go:260` | yes: the arm names the new table too |
| `record/recordtest/recordtest.go:157-161` | no |
| tests: `recordsql/reportcheckpoint_test.go`, `record/migrate/reportcheckpoint_test.go`, `record/behinddelivery_test.go:27`, `record/deliver_internal_test.go:17`, `record/deliverytest/deliverytest.go:33`, `cli/raiserturn_test.go:161`, `record/retiredelivery_test.go:63`, `record/plansweep_test.go:39` | the checkpoint tests gain the rows; the delivery tests follow the labels |

### 3. The lens's numbers leave its work list `[MODIFY]` / `[DELETE]`

**Property.** No lens work-list surface carries a pin or head number: not the `last_read` object, not an item's subject id, not an item's sentence, not a notes line, not a delivered piece. Survivors outside the work list are rows below with their reason.

`last_read` consumers — `grep -rn 'LastRead\b\|LastReadJSON\|lastReadOf\|\.LastRead\.' --include=*.go . | grep -v _test`: `record/retirement.go:490-494, :504-535` (the struct loses `Pin` and `Head`; `lastReadOf` keeps computing the kind from them as locals); `record/sitting.go:105, :155-156, :177-198`; `record/queue.go:103` (reads `Kind`: unchanged); `sittingwrite/delivery.go:75, :373`. Three test files name `last_read`.

Census over the plugin and the generators' sources, the generated copies, tests and goldens apart — 145 hits, every one in a row below:

```
grep -rnE 'your pin|last_read|SINCE YOUR PIN|since the pin|its pin|pin [0-9%]|\(pin |head [0-9%$]|head %d|at head|report head|SubjectReport' \
  plugins/frank-exchange-of-views scripts/agentgen/src \
  --include='*.go' --include='*.md' --include='*.js' --include='*.mjs' --include='*.sh' --include='*.proto' --include='*.json' \
| grep -vE '_test\.go:|/testdata/|_gen\.go:|\.pb\.go:|/agents(-claude)?/|frank-exchange-of-views/skills/(adversarial-audit|research-protocol)/SKILL\.md:|debate\.test\.mjs:'
```

| Carrier | Change, or reason to survive |
|---|---|
| `record/sitting.go:105, :194, :197`; `LastReadJSON` `retirement.go:490-494`; `sittingwrite/delivery.go:75` | R2: `last_read` drops `pin` and `head`, keeps `kind`; the two sentences lose their numbers |
| `record/sitting.go:178`; `record/deliveredtext.go:248, :250`; `record/itemkind.go:92, :93, :115` | R2: a report item's subject is `{kind: report}` with no id (`about` allows it, `itemkind.go:114-118`); the comment "its id is the report head" goes |
| `record/sitting.go:276, :278` | R2: "dispatched against report head %d" loses the number, for every seat it reaches |
| `debate.js:886` (`lastReadClause`, authored: "at head H") | R2: reworded to the kinds alone |
| outside the regex: `sittingwrite/delivery.go:348` (a note line prints `· report <id>`), `record/itemkind.go:109` (`"id"` with no `omitempty`) | R2: a subject with no id prints its kind alone and its JSON omits `id`; left alone, R2 would print `· report ` and `"id":""` |
| `record/deliverycut.go:17, :25, :27, :88, :256`; `record/deliveredtext.go:215, :283, :285`; `cli/seat/help/register.md:11`; `scripts/agentgen/src/skills/adversarial-audit/SKILL.md:26`, `skills/research-protocol/SKILL.md:46` | R1: each describes what a lens is delivered and gains the marked re-read |
| `recordpb/record.proto:313` (`EVENT_TYPE_REPORT_READ`'s `means`), `terms/terms.json:862` ("report read"), `record/reportread.go:7`, `recordsql/views.go:680` | not this plan's — `plans/sitting-ends-at-write.md`, S4: each defines the pin as "in a sitting it sat through", which ruling 3 widens to "or in no sitting". Regenerate: `protogen`, `vocabgen`, `vocabdoc` (`docs/vocabulary.md:511, :517, :523`) |
| `### REPORT DELIVERED` `record/viewjson.go:1978, :1984` | not this plan's — `plans/sitting-ends-at-write.md`, T1 deletes it |
| the word, no number, on a seat-facing surface: `cli/seat/verbs.go:330, :331, :369`; `cli/seat/help/{dispatch.md:13, spot-check.md:7, engage.md:7}`; `scripts/agentgen/src/{fragments/lens-duties.md:5, red-lens-voice.md:15, red-chair.md:12, :14}`; `debate.js:899, :900, :902`; `record/queue.go:48, :49`; `record/itemkind.go:22, :65`; `view/changes.go:21, :55`; `terms/terms.json:852, :873`; `hooks/README.md:115`; `docs/seat-command-triggers.md:75`; `docs/citations-and-bibliography.md:162`; `skills/research-protocol/references/report_template.md:8` | survive: ruling 2 removes the lens's numbers, not the word |
| `cli/seat/verbs.go:362` (`show report --json` `head`) | survives: a selector read's own stamp, outside the work list; #1962 holds the wider question |
| the plan's `head`: `record/dispatch.go:71`; `debate.js:403, :968, :988, :1118` | survives: ruled; the engine reads it |
| the chair's numbers: `cli/chair/dispatch.go:77, :90`; `cli/chair/spot_check.go:51`; `record/passblockers.go:479, :488, :493`; `record/dispatch.go:21, :105, :1329`; `record/retirement.go:99, :101, :103, :105, :127, :129, :138`; `record/spotcheck.go:260`; `debate.js:394, :409, :920` (and `:428`, `tests/simulator/testdata/{m3-chair3,o3-chair5}-relays.json`, outside the grep) | stay (ruled 2026-11-28); #1962 |
| another sense — a dispatch's pin, a cite's git pin, a hash pin, a false match on "that heading" — or a report subject with no id: `record/record.go:1478`; `cli/blue/engage.go:56`; `cli/chair/dispatch.go:46`; `recordpb/record.proto:306, :2527, :2531`; `record/successor.go:256`; `record/migrate/replay.go:235, :566`; `setup/setup.go:222`; `setup/run.go:202`; `skills/research/SKILL.md:29`; `native/build.sh:107, :167`; `record/passage.go:58`; `report/layout.go:27`; `record/passblockers.go:460` | not carriers |
| reaches no seat — code comments on the report head or the pin, the tool's own refusals and test helpers: `cli/chair/dispatch.go:15`; `cli/seat/selector.go:261`; `record/anchorchecks.go:77`; `record/cast.go:21`; `record/dispatch.go:17, :201, :543, :957`; `record/passblockers.go:83`; `record/reportjson.go:10, :52`; `record/retirement.go:92, :111, :134, :264, :273, :355, :356, :420`; `record/sitting.go:111`; `record/spotcheck.go:45, :84, :236`; `record/viewjson.go:821, :1944, :1967`; `recordsql/views.go:690, :742, :745, :746, :752`; `recordpb/record.proto:1168, :2600`; `record/migrate/planids.go:20-25, :39`; `record/migrate/replay.go:577`; `record/deliveredtext.go:22, :179`; `record/deliverytest/deliverytest.go:41`; `seatprobe/boards.go:959` | survive; comments that restate a changed rule are reworded with it |

Generated copies follow their sources: `agents/` and `agents-claude/` (12 files each; `scripts/agentgen/main.go:70`) and the two preloaded skills by `agentgen`; `cli/seat/help_gen.go` by `cmdtreegen`; `debate.js`'s marked regions by `enginegen`. Goldens — `difftest/testdata` (15 files), `tests/simulator/testdata` (4), `record/migrate/testdata` (1) — and `tests/simulator/debate.test.mjs` (8 hits) are regenerated or edited, every changed line read.

Lost with the numbers: a lens can no longer copy a head into `close --verified-against` (25 of 63 behind sittings did, unprompted — #1955).

### 4. Why not a search — the owner's question, measured

The owner asked (2026-11-28) why a change could ever fail to be placed, when and how often, and what other answers exist.

*It can always be placed.* An earlier draft SEARCHED the final text for an edit's stamped new side, and a search fails when that text has moved on. The fold never has to search for what it wrote: it holds the running text before and after every op (`render.go:324-346`); an edit lands where `bluedoc.LocateUniqueReplacing` finds its old side in that running text (`:106-122`), unique by the write's own validation; and `ApplySplice` replaces `report[start:end]` (`:48-56`). It is the carrying that keeps anchor tokens in place through edits (`keepUncarried`, `:59-66`).

*How often the search fails.* Measured on the 39 archived runs (`archiveNames`; origin/main 5f0911494; load 2), by scratch folds written for these plans — neither `ReplayTraced` nor the table exists yet. Nine runs hold lens deliveries: 63 whole re-reads over 10 distinct windows — a window is one record's head moves between a lens's pin and its re-read's head, and the lenses of a run usually share it — holding 220 distinct head moves and, in the 10 re-read texts, 797 blocks.

| Of 220 head moves: can a search find the edit's stamped new side? | Raw text search | Search by blocks of visible text |
|---|---|---|
| found once | 162 (73.6%) | 187 (85.0%) |
| not found | 57 (25.9%) | 32 (14.5%) |
| found more than once | 0 | 0 |
| a pure cut, nothing to search for | 1 (0.5%) | 1 (0.5%) |
| re-reads holding a change it cannot find | 63 of 63 | 56 of 63 |

Of the 220, 209 (95.0%) still have text of their own standing at the re-read; the other 11 — 10 whose words a later edit overwrote or cut, and the pure cut — have none, and no design has anything of theirs to place. So at least 22 of the block search's 32 misses are changes whose text still stands in part, in a paragraph edited again. A raw search fails as well whenever a cite, a finding or a gap later puts an anchor token into an edited paragraph (25 of the 57). No search failed because text stood twice. `deliverycut.go:271`'s anchor-only filter fired on none of the 1,227 (head move, re-read) pairs.

| By block: is every block that reads differently from the last-delivered text marked? | Marked, of 797 | Changed blocks left unmarked | Marked blocks that read as before |
|---|---|---|---|
| this plan's rule (§III.1), run from each pin | 234 (29%) | 0 | 5 |
| characters carried through the fold | 234 | 0 | 5 |
| stamped-block match: mark a block whose visible text is a block of any later edit's stamped new side | 237 | 1, in 7 of the 63 re-reads | 9 |

The plan's rule and the character fold mark the same blocks on all 63 re-reads. Over the 47 distinct change-delivery windows (3,644 blocks) the rule leaves 0 unmarked and marks 3 blocks that read as before; the stamped-block match leaves 3 unmarked — a table row and a list item, which parse differently out of their table or list, and a paragraph edited twice. No archived window removes a whole block. Widening passages to blank lines will mark more than 29%; by how much is not measured. Seven older archives (2026-08-22 to 2026-09-02) hold ops that do not replay from the first and were not measured.

*Cost of the alternative the ruling set aside.* A from-the-first replay at the delivery took 19–630 ms on the nine records (169 ops over 16 kB: 79 ms; 254 ops over 80 kB: 630 ms), by a harness that replays op by op and keeps every text — an upper bound. The stored design makes none.

| Answer | Verdict |
|---|---|
| Text search, the misses appended as BEFORE/AFTER items | rejected: it cannot find 14.5% of changes, in 89% of re-reads, and a superseded edit's item shows text the report no longer holds |
| A rendered diff of two texts, the last-delivered and the present | rejected: it needs the last-delivered text kept or rebuilt, and a diff guesses its alignment where paragraphs repeat or move |
| Stamped-block match | rejected: no replay and nothing new stored, but it left a changed block unmarked in 7 of 63 re-reads, cannot know that it did, and cannot show a removal |
| Carried through a replay from the first op, at the delivery | set aside by ruling 5: exact, but a replay in Go through every past text, in a hook |
| **Block writers stored at the write** | **ruled**: exact, nothing replayed at the delivery, stored where the write already folds the text |

### 5. Pull requests

| PR | Concept | Deletes | Epoch |
|---|---|---|---|
| R1 | the report's checkpoint records who last wrote each block, and a lens's whole re-read marks what changed since its last delivery | the unmarked re-read | the next |
| R2 | the lens work list names no pin or head | `LastReadJSON.Pin`, `.Head`; report-item subject ids; the numbers in `sitting.go:194, :197, :276, :278` | — |

R1 lands first: the pin and head are today the only route to what changed in a full re-read (#1955), and R2 removes them. R1 is one concept and not two: a table no reader reads has no destination. Its first commit carries the table, the fold, the stamp and oracle A green on CI before the commit that makes the cut read them; the PR body names that run. No pull request leaves a dual path: there is no second way to mark and none to remove.

**Migration.** The replay writes every event through `InsertTx` (`record/replaywriter.go:84` → `record.go:824`), so `stampReportCheckpoint` fills the rows on the migrated record, in its own event ids; nothing is translated. `migrate/sqlitesource.go:257-261` skips `report_checkpoint` as a fold of the events and names the new table beside it; without that a source holding the rows is refused (`:284-293`). Carriers of the epoch: `requirements.json`, `recordpb/schema_gen.go`, `recordsql/testdata/schema.epoch.json`, `schema.sql`, `cmd/feov-migrate/testdata/help.golden`, `record/testdata/plans.golden`. No residue: the rows are derived, and the source's are not carried.

## IV. Risk & Mitigation

| Risk | L × I × cost | Mitigation |
|---|---|---|
| The advance rule misses a change, which reads as "unchanged" | L × H × L | oracle B on seeded records and every archive (§V 3); zero unmarked measured; the mutants of §V 4 |
| Stored rows drift from the fold (a stamp that skips an op, a row order, a failed transaction) | L × H × L | oracle A after every append, on archives and in the fuzz (§V 2) |
| The packer drops, delays or misplaces a mark line | M × H × L | the cut reads its pieces back against rules (1)–(4) and refuses; three mutants (§V 4) |
| `anchor.Blocks` changes between binaries at one epoch, so stored rows no longer fit the text | L × H × M | the count check refuses a misfit; a change to the block reader bumps the epoch, as a renderer change must (`schema.go:157-159` names the same hazard for the text) |
| The stamp grows: a block parse per op, a row per block every eighth op | H × M × M | scratch puts it at 2–2.4 times today's; R1's first commit measures the shipped fold against the 2.5 bound (§V 1). Over it, the advance moves to once per stamp — each op's splice carried as a touched range through the stamp's ops, blocks parsed at the stamp's two ends, a quarter of a parse per op — which is a different rule and is held to oracle B before it replaces this one |
| Passages widened to blank lines mark most of a heavily edited report | M × L × L | the label states N; §V 5 reports the share per archived re-read |
| R2 before R1 would leave a re-read with neither number nor mark | L × M × L | the sequence; R2's description names R1's merge |
| Golden conflicts across ~8 streams, and with the other plan's epoch | H × L × L | rebase on golden or schema change; regenerate at `-count=1` |

### Decided

- The lens's numbers go, `kind` stays; the re-read is marked — ruled 2026-11-28.
- A removed whole block shows one line — ruled 2026-11-28.
- What changed is learned from spans stored at the write — ruled 2026-11-28. This plan's reading of it: block writers beside the report checkpoint (§III.1), per block and not per run, for the reasons there.
- The chair's pin and head numbers stay; #1962 — ruled 2026-11-28.
- A test may replay a record to a past event; no shipped path does — as the ruling was relayed.

### Open questions for the owner

| # | Decision | What happens unless he says otherwise |
|---|---|---|
| Q1 | Does a marked passage also show what it said before? An ordinary sitting's change items do (BEFORE / AFTER); the re-read is of the text as it stands. | marks only |
| Q2 | About 2% of marked blocks read exactly as they did (an edit that left a block's visible text alone), and widening to blank lines marks whole tables and tight lists. Mark generously, or compare a block's visible text before marking? | generously: a comparison would need the earlier text, which nothing keeps |

### UNVERIFIED

- Every figure for the stored rule is a scratch fold's, run from each pin over texts replayed in a test, with head moves read from the `head_move` view rather than the op stream (scratch `issue-1955/plan-rev2/`, `plan-rev3/`); the shipped fold, the stamp and the table are unbuilt.
- The shipped stamp's added cost, and a traced tail fold's. The scratch figure: over the 39 archives a replay from the first took 9–620 ms, one block parse per op 4–860 ms, and the scratch carry, which parses twice per op, 10–1,750 ms.
- How much of a re-read lies in a passage once passages are widened to blank lines.
- The removal line on real data: no archived re-read window removes a whole block, so seeded tests alone exercise it.
- That `anchortext.Visible` is the right reduction for oracle B's comparison (`anchortext/anchortext.go:82-86` says it is not right for every reader).
- That the seven archives whose ops do not replay from the first hold a failed checkpoint and no report read.
- Whether dropping a report item's subject id changes `changed_since_delivered`: `itemID` marshals the whole item (`sittingwrite/delivery.go:393-396`) and is read at `:379-383`; where `changed` is computed was not found.
- How `difftest` declares a scenario that renders delivery items.
- Whether a row per block rewritten at every stamp passes planguard as issued, and what `shapes.golden` records for it.

## V. Verification Plan

First: `unset $(env | grep -o '^FEOV_[A-Z_]*')`. `FEOV_RELEASE_GATE` is never set locally.

1. **Timing, R1's first commit.** `stampReportCheckpoint` on the eiffel 11-11 and the 80 kB auditing 11-10 records, with `Replay` and with `ReplayTraced` and the row write, arms interleaved, setup outside the timer; a tail fold with and without writers. Against §I's 2.5 bound; over it, §IV's row decides what merges.
2. **Oracle A — the stored rows are the fold's.** `ReportBlockWritersDiff` `[NEW]`: the checkpoint's rows advanced through its tail, against `ReplayTraced` run from the first op in one pass. It replays in a test; no shipped path does. Asserted after EACH append in `recordsql/reportblocks_test.go`, on every archived run (`archiveNames`, `record/backingarm_test.go:304`), in `record/plansweep_test.go`, and in the release fuzz beside the gap-state oracle. It holds chunking and storage; it cannot see a wrong rule, which is oracle B's.
3. **Oracle B — the rule marks what changed (real data).** `TestTheBlockWritersMarkEveryChangedBlockOnEveryArchivedRun` (`record/migrate`): at every lens re-read, replay to the pin and to the re-read in the test, and assert that every block of the re-read's text whose visible text is no block of the pin's has a writer after the pin. Non-vacuity: 63 re-reads, 220 distinct head moves, 797 blocks, and how many `deliverycut.go:271` filtered (none today). The same assertion runs on the seeded record of step 4. This check's first run, on a scratch fold, found 13 unmarked blocks — a cut at the head of a paragraph — which is why the rule is stated by blocks before and after.
4. **Seeded tests `[NEW]`.** `TestTheReReadMarksWhatChangedSinceTheLastDelivery` (`record`): a lens with one sat-through whole delivery, then two edits to abutting sentences (one passage), an edit re-edited later (its successor's text marked, no item), a cut at the head of a paragraph and one at its tail (the paragraph marked), a cut of a whole paragraph (the removal line), a paragraph split in two, a cite into an edited paragraph (still marked), a retire in untouched text (no mark), a passage crossing a `\n## ` heading, a table row edited (the table marked whole), and a budget that splits a marked passage (the continuation line). It asserts rules (1)–(4) on the pieces and oracles A and B. Mutants, each red: omit a mark line; omit the continuation line; delay an end mark by a block; let a suffix block keep its writer without the whole-block test; let an anchor insert write; skip an op in the stamp.
5. **Share marked.** The archive test of step 3 reports, per re-read, the blocks in a passage before and after widening.
6. **R2's test `[NEW]`.** `TestALensWorkListCarriesNoPinOrHeadNumber` (`sittingwrite`): renders a lens's work list — the JSON, each rendered item and the notes lines (`delivery.go:338-352`) — for every `last_read` kind, an owed register and a delivery of each sort. It asserts that `last_read` has no `pin` or `head` key, that no report item has a subject id (no `"id":""`, no `· report ` with a trailing space), and that no head-move event id of the seeded record appears as a token. The fixture pads the record so every head move's id is above 1,000 and asserts first that no count the list prints reaches that.
7. **Golden `[NEW]`.** A `difftest` golden renders the first item of a first whole delivery and every item of a marked re-read, so both labels and all four mark lines are read in review.
8. `go -C scripts run ./check`; `TestEveryDDLChangeBumpsTheEpoch` with the restamped epoch golden; planguard's goldens read line by line.
9. After text changes: `agentgen`, `enginegen`, `cmdtreegen`, each gate's `-check`; every changed golden line and each affected `--help` read before and after.
10. CI on the final head: the four `release-fuzz` parts and `release-fuzz-tally`.
11. Auditor gate: `/prosthetic-conscience:plan-audit`.
