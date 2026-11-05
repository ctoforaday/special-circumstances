# Mint requires its class's prerequisite — `verify` records only (#1727)

> **Status: DRAFT, unaudited.** Written 2026-11-05 on `plan/mint-prerequisites` at `origin/main`
> `68a735d51` (EventSchema **61**, `requirements.json:3`, taken by `6ed87cbb7`). Scope ruled by gblock
> 2026-11-04 (issue #1727, comments 2–4; the picked options' wording is the lead agent's, not his). Eight
> forks are open (§Forks); none is decided here. Every census row was re-run by grep on this tree; the
> commands are in §III.9.

## I. Summary & Goals

**Objective.** `mint` stays the one act that creates a gap, and each `GapClass` value DECLARES what
must be on the record before a gap of that class can be minted. One generic check in the mint's write
reads the declaration and refuses an unmet prerequisite, naming the act that meets it. `lens verify`
records a reading and nothing else: it takes no gap flags, writes no Mint and no Anchor, and names no
gap. The link between a gap and the reading it rests on moves to the gap: the Mint names the event as a
field, which reverses #1694's `Verify.carried_by` / `carrier_minted`.

**Why.** gblock, 2026-11-04: "I would think the answer is to revise to mint to reject if the source isn't
verified? Not to magically mint the gap on verify?" In #1694's smoke (Eiffel Tower, `0e8768a`) the
adversary minted `claim_contradicts_source` with a plain mint and no reading (G-b2ba7d66, event 1054):
the reading row (url, copy, outcome, confidence) was lost to prose in `mint_reason`. A class whose
definition is a reading should be refused without one, and the refusal should name `verify`.

**Ruled scope (not re-opened here).**

| # | Ruling |
|---|---|
| R1 | `mint` stays the one act that creates a gap |
| R2 | `verify` records only and takes no gap flags |
| R3 | each GapClass declares its mint prerequisite as an enum-value option, as SourceOutcome values declare `gap_classes` |
| R4 | one generic check in mint enforces it; the refusal names the act that satisfies the prerequisite |
| R5 | no class gets a hand-written check; #1721's out_of_scope check (`record.go:1901`, not ~1970 as the issue says) folds into the declared option |
| R6 | the gap names its reading as a field (reverses #1694's `carried_by`/`carrier_minted`) |
| R7 | the classes: 32 `claim_contradicts_source`, 25 `citation_figure_misattribution`, 31 `out_of_scope`, 27 `live_source_drift`, 2 `claim_contradicts_own_record`, 23 `figure_recount_fails` (table in §III.1) |
| R8 | a second reading landing in an open gap becomes the ordinary one-gap-per-class refusal |
| R9 | every other class is a judgement about the text and has no prerequisite (13 and 26 named in the issue) |

**Success criteria.**

1. `verify --class …` (and every other gap flag) is an unknown-flag refusal; a `verify` writes exactly
   one event, a Verify, whatever its outcome (tests T5, T6).
2. For each class in R7, a mint without its prerequisite is refused, nothing is written, and the
   refusal names the satisfying act; a mint with it lands (T2, T7–T13, T23, T24).
3. No code path compares a `GapClass` value to decide a mint's admission; deleting a class's declared
   prerequisite from `record.proto` turns its refusal test red (T1, T3).
4. Every reader of the reading-to-gap link reads it off the Mint's field; no `Verify.carried_by` or
   `carrier_minted` exists in the proto, the schema, the help, the constitutions or the fuzz (§III.9
   census, gated by the zero-hit greps in §V).
5. `feov-migrate` brings an archived #1694-era run forward with zero refusals, each archived reading act
   becoming a reading followed by a Mint that names it (T19 on the Eiffel archive).
6. The release fuzz drives every prerequisite class through its declared act and through its refusal,
   with no zero tallies (CI only, §V check 12).

**Out of scope.** Changing any class's meaning; giving 13, 26 or any R9 class a prerequisite; the
finding-into-gap fold (#1695); 1a′'s lifecycle itself (#1739), whose seam is §III.8.

## II. Technical Context

**Record shapes today (main `68a735d51`).**

- `GapClass` (`record.proto:384–484`), 32 values. Options on each: `means`, `material_default`
  (50011), `neighbor` (50012), `distinguisher` (50013). None restricts how a gap of it is minted, except
  #1721's hand check (`record.go:1900–1902`) inside `requireMintSubject` (`record.go:1892`), which runs in
  `validate` (`record.go:977`) — the SHAPE check every writer runs, the replay included.
- `SourceOutcome` (`record.proto:528–541`) declares `gap_classes` (50014: refutes → 32, 26; absent →
  25, 26) and `found_source_gap_classes` (50015: supports, supports_with_bridge, weak → 29; refutes →
  32). Read through `recordpb.GapClassesOf` / `FoundSourceGapClassesOf` (`recordpb/gapclass.go:63–75`),
  `record.ReadingGapClasses` (`reading.go:46`), `record.FindsAgainst` (`reading.go:34`, "it reads the
  work and its outcome admits a gap").
- `Verify.carried_by` (21, references `mint.gap_id`) and `carrier_minted` (22) (`record.proto:1425–1434`),
  set by the tool only: by `AppendReading` (`reading.go:92`) on an in-act mint, or by `appendBodies`
  (`record.go:704–711`) when the one-gap rule HOLDS the reading's Mint (`onegapperclass.go:91–93`,
  `errReadingHeld`, `reading.go:75`).
- `Mint` (`record.proto:882`) has no field naming an event.

**The reading act today.** `lens verify` (`cli/lens/verify.go`) registers mint's gap flags
(`gf.register(c, readingSubject)`, `:160`; `readingRequiredGapFlags`, `:108`); where
`ReadingGapClasses` of the stamped body is non-empty it requires them (`:415–421`), and writes Mint +
Anchor + Verify in one transaction (`recordReading`, `:437`; `AppendReading`). The Anchor places the gap
by identity at the end of the citation's sentence (`sentence_of`) for `--anchor`, or at `--quote` for
`--url`. The gap of a citation reading is outside the mint budget (`mintbudget.go:255`,
`ownAnswersAt` `:316`); a found-source reading's gap counts (F6 as narrowed 2026-11-03). A reading's gap
is exempt from estoppel (ruled 2026-10-31; `mint.go:144`).

**The failed check.** `failedChecksOf` (`failedcheck.go:95`) builds a found-against check's carriers
from `v.GetCarriedBy()` and its successor tree (`carrierTree`, `:177`); `owed()` is false for
found-against (`:80`); `requireFailedCheckAnswered` (`:390`) refuses the close of any carrier.
`repeatReadingRefusal` (`:333`) refuses a seat's repeat reading while the gap carrying its first is open.

**Measured inputs to this plan (2026-11-05, this tree).**

| Measure | Value | How |
|---|---|---|
| Archived mints of R7 classes, after `migrate`'s class translation (`migrate/gapclass.go`) | 46 in 27 SQLite archives: `claim_contradicts_own_record` 31 (19 runs), `figure_recount_fails` 10 (7), `citation_figure_misattribution` 4 (4), `live_source_drift` 1 (1); 32 and 31: 0 | python over each `run-archive/*.tar.gz`'s `records/record.db`, `SELECT class, count(*) FROM mint`, mapped through `archivedGapClasses`; the 6 era (JSONL) archives not counted |
| Archives holding a Verify with `carried_by` | 0 of 33: the newest is `2026-10-29_is-91-prime-m36` (epoch 46); #1694 took epoch 59 | same scan |
| The #1694 smoke record | `~/.claude/scratch/contradiction-impl/elicit-adversary/2026-11-04_eiffel-tower-height.tar.gz`, epoch 56 (#1694's branch numbering); 59 readings, 7 with `carried_by`: 6 in-act (events 167 absent→G-8dbfce13, 182 refutes→G-3b2b7d93, 542 refutes→G-1d97d860, 791→G-f66af91d, 794→G-dd7b403e, 797→G-de476176) and 1 held (798 refutes→G-1d97d860, `carrier_minted` 0); class-32 mints 6 (5 in-act, 1 plain: G-b2ba7d66) | sqlite read of the extracted archive |
| That record under main's `feov-migrate` | 1213 events in → 1203 out, **0 refusals**, epoch 61; all 7 links carried | `go build ./cmd/feov-migrate`; `feov-migrate --from … --to …` in `~/.claude/scratch/mint-prereq/` |

**Constraints.** No backwards compatibility: `carried_by` and `carrier_minted` are deleted and reserved,
and `migrate` is the only path that upgrades a record (CLAUDE.md, Change discipline). Seat prompts name
acts, never flags (`TestTheSeatPromptsNameNoVerb`). Generated carriers are regenerated, never edited:
agent definitions and the two preloaded skills (`agentgen`), `debate.js`'s two marked regions
(`enginegen`), `help_gen.go` (`cmdtreegen`), `annotations_gen.go` (`vocabgen`), `docs/vocabulary.md`
(`vocabdoc`), `record.pb.go` (`protogen`).

## III. Proposed Changes

### III.0 The design, in one table

| # | Decision | Where |
|---|---|---|
| D1 | A new enum `MintPrerequisite`; each value carries `means` and a new option `satisfying_act` (the act its refusal names). Each `GapClass` value declares `repeated MintPrerequisite mint_prerequisite`; any one declared value satisfies it (25 declares two) | §III.1 |
| D2 | One generic check, `requireMintPrerequisite`, keyed only on the declaration: for each declared value it runs that VALUE's predicate (one arm per `MintPrerequisite` value, never per class) and refuses when none holds | §III.2 |
| D3 | The Mint names the event its prerequisite rests on, as a field — `rests_on` (int64, references `events.id`) under F3's recommendation | §III.1, F3 |
| D4 | `verify` loses every gap flag and the in-act write; it prints the reading's event id so a mint can name it | §III.3 |
| D5 | `mint` takes the event (`--rests-on <event>` under F7's recommendation); for a reading, the gap's subject is derived from the reading by identity | §III.4, F7 |
| D6 | Every reader of the link reads `mint.rests_on`; in SQL through one view, in Go through one helper | §III.5 |
| D7 | `migrate` rewrites each archived reading act as reading-then-Mint, translating `rests_on` through `newID` | §III.6 |
| D8 | A second reading in an open gap is the ordinary one-gap refusal of the mint that names it; the held-reading arm is deleted | §III.2 |

**Alternatives considered for D1/D2** (the ruling fixes "an enum-value option"; these are the shapes of it):

- *A structured option message* (`MintPrerequisite { EventType act; repeated SourceOutcome outcomes; bool own_seat; AboutKind subject; … }`) evaluated field by field. Rejected: 31 (a subject shape), 2 (any event), 23 (a reproduce with a result) and 27 (a copy comparison) need different field sets, so the "generic" evaluator grows a branch per field combination, which is a per-prerequisite arm in disguise with less to read.
- *Reversing `gap_classes` only* (keep the SourceOutcome lists; mint checks a reading's outcome admits the class). Rejected: covers 32 and 25 and nothing else in R7, and the declaration would sit on the outcome, not the class (R3 says the class declares).
- *Per-class predicates.* Ruled out (R5).

### III.1 The record — `internal/record/recordpb/record.proto` [MODIFY] (contract: event schema)

**New enum** (grep before coining, §III.9 row N1: `MintPrerequisite`, `mint_prerequisite`,
`satisfying_act`, `rests_on`, `rests-on` — 0 hits each outside `plans/`; "prerequisite" appears once, in
an OCR comment, `cli/ocr.go:67`, another sense):

```proto
// MintPrerequisite is WHAT THE RECORD MUST HOLD BEFORE A GAP OF A CLASS IS MINTED, declared on the
// class (GapClass's mint_prerequisite) and checked by one rule (record.requireMintPrerequisite).
enum MintPrerequisite {
  MINT_PREREQUISITE_UNSPECIFIED = 0;
  MINT_PREREQUISITE_OWN_READING_REFUTES = 1 [(means) = "…", (satisfying_act) = "verify"];
  MINT_PREREQUISITE_OWN_READING_ABSENT = 2 [(means) = "…", (satisfying_act) = "verify"];
  MINT_PREREQUISITE_OWN_READING_NOT_THE_WORK = 3 [(means) = "…", (satisfying_act) = "verify"];
  MINT_PREREQUISITE_ABOUT_AN_INQUIRY = 4 [(means) = "…", (satisfying_act) = "mint --about-kind inquiry"];
  MINT_PREREQUISITE_NAMES_A_RECORD_EVENT = 5 [(means) = "…", (satisfying_act) = "<the act recorded the event>"];
  MINT_PREREQUISITE_REPRODUCE_DIFFERS = 6 [(means) = "…", (satisfying_act) = "reproduce"];   // F1
  MINT_PREREQUISITE_READING_OF_A_DIFFERING_COPY = 7 [(means) = "…", (satisfying_act) = "verify"]; // F2
}
extend google.protobuf.EnumValueOptions {
  repeated MintPrerequisite mint_prerequisite = 50016;   // on GapClass values
  optional string satisfying_act = 50017;                // on MintPrerequisite values
}
```

`satisfying_act` names an act, not a command path: the refusal wording composes it the way every refusal
names a verb today (`act := "read that citation with …"`, `verify.go:634`); a seat prompt never renders it
(`TestTheSeatPromptsNameNoVerb` scans prompts, and the option reaches only refusals and `mint --help`).

**The class table** (issue comment 3), with the value each class declares:

| Class | Declares | What the predicate checks (D2), on the event `rests_on` names unless stated |
|---|---|---|
| 32 `claim_contradicts_source` | `OWN_READING_REFUTES` | a Verify by the minting seat; outcome `refutes`; reads the work (`VerifyReads`); it STANDS — the seat's latest reading of that anchor that reads (the ledger's replacement rule, `anchorchecks.go:96`), or for a found source the seat's latest reading of that url for that sentence; its sentence is in the report as it stands |
| 25 `citation_figure_misattribution` | `OWN_READING_ABSENT`, `OWN_READING_NOT_THE_WORK` | a standing Verify by the minting seat of a CITATION (anchor non-empty), and either outcome `absent` reading the work, or `source_completeness` `not_the_work` |
| 31 `out_of_scope` | `ABOUT_AN_INQUIRY` | no event: the Mint's `about_kind` is `inquiry` (the #1721 rule, moved — F4) |
| 27 `live_source_drift` | `READING_OF_A_DIFFERING_COPY` | **F2** — not evaluable from today's fields |
| 2 `claim_contradicts_own_record` | `NAMES_A_RECORD_EVENT` | an event exists with that id, before the mint; any seat, any type |
| 23 `figure_recount_fails` | `REPRODUCE_DIFFERS` | **F1** — under its recommendation: a Reproduce with `rerun` `differs` of the computation the gap's subject names, conditional on the subject being a computation's |

The table follows the issue literally on "own": 32 and 25 say "the seat's own", 27 and 23 do not, so
their predicates accept any seat's event. A reviewer who wants "own" there says so; it is a one-line arm
change.

**`Mint` [MODIFY]:** `optional int64 rests_on = 27 [(sql) = { references: "events.id" }];` — 26 is
reserved (`record.proto:989`), 25 is `class_material`. The reference follows
`Stale.through` and `ReportRead.head` (`record.proto:2497`, `:2514`). Not unique: one record event may
ground several gaps of class 2 (one inquiry fate contradicted in two sentences), and a successor names
its own event (or none; its carrier tree carries the check, `carrierTree`).

**`Verify` [MODIFY]:** delete `carried_by` (21) and `carrier_minted` (22); `reserved 21, 22;
reserved "carried_by", "carrier_minted";` beside the existing reserved list (`:1439–1440`). Delete the
comment block `:1425–1432`.

**`SourceOutcome` [MODIFY] — F5.** `gap_classes` (50014) and `found_source_gap_classes` (50015) say
which classes a reading's act mints in; no reading mints after this change, so both statements become
false and both options are deleted (numbers reserved with a comment, as 50007 and 50010 are). What
`FindsAgainst` needs instead is F5's choice. Each outcome's `means` (`:529–541`) loses its "mints … in
the same act" clause and says what the reading records and which class's mint it can ground.

**`GapClass` [MODIFY]:** add `(mint_prerequisite) = …` on 31 (Part 1), 32, 25, 2 (Part 2), 23, 27
(Part 3, per F1/F2). Each class's `means` is unchanged; the prerequisite's own `means` renders beside it
in `mint --help` (§III.7). The `means` and `satisfying_act` strings elided as "…" above are written in the
Part that adds each value and read in G2 before merge.

**Epoch.** One bump, in Part 2: `requirements.json` `eventSchema` → **E**, the next free number when
Part 2 merges (62 if #1739 has not merged; #1739's branch also writes 61, diff line 5713, so whichever
lands second takes the next). `go -C scripts run ./schemagen` regenerates `record.EventSchema`. Part 1
adds options only, which change no table (`gap_classes` is "an option, not a vocabulary facet",
`record.proto:155–157`), so Part 1 bumps nothing — **checked by `go -C scripts run ./schemagen` reporting
no change and `recordsql/testdata/schema.sql` unchanged.** Part 3 under F2(a) adds a Verify column and
bumps again (E+1).

### III.2 The generic check — `internal/record/` [NEW `mintprerequisite.go`, MODIFY `record.go`, `actjudge.go`, `onegapperclass.go`, `reading.go`, `writeview.go`, `replaywriter.go`]

**`requireMintPrerequisite(w writeView, seatID string, m *recordpb.Mint) error`** — an act judgement,
called from `judgeAct`'s Mint arm (`actjudge.go:36–48`) BEFORE the one-gap rule and the budget (a mint
that will be refused for its prerequisite should not be told about a held subject or a full budget first).
It reads `recordpb.MintPrerequisitesOf(m.GetClass())` (new, beside `MaterialDefault`, generated
annotations) and returns nil for an empty declaration. For a non-empty one it evaluates each declared
value through `prerequisiteArms[value]` and passes when any holds. A refusal:

> `--class claim_contradicts_source rests on your own standing reading that refutes the sentence, and
> this mint names none (or: event N is <what it is>). Record the reading with verify, then mint naming it
> (--rests-on <its event>). Nothing is recorded.`

composed from the value's `means` and `satisfying_act`, never a per-class string. `prerequisiteArms` is
a `map[recordpb.MintPrerequisite]func(writeView, string, *recordpb.Mint) (ok bool, why string, err error)`
with one entry per value; `TestEveryMintPrerequisiteHasAnArm` (T1) fails a value with no entry. An
event named by `rests_on` on a class that declares no event-bearing prerequisite is F6's question.

**Where it runs — F4.** The recommendation is `judgeAct` only. Consequence, stated: the replay does not
ask it, so the 46 archived R7-class mints and G-b2ba7d66 come forward naming no event (§III.6 residue),
and #1721's out_of_scope rule moves from the shape check (every writer) to the live judgement — the
replay stops asking it, on a vocabulary value no archive holds (0 of 27, §II).

**#1721's hand check [DELETE]:** the `out_of_scope` case at `record.go:1900–1902` and its comment
`:1898–1899`. `requireMintSubject` keeps the two subject cases.

**The one-gap rule [MODIFY] `onegapperclass.go`:** delete the held arm (`:91–93`), `pairedSentenceOf`
(`:102–108`), and the "A READING'S MINT IS NOT REFUSED, IT IS HELD" paragraph (`:40–45`). A mint whose
`rests_on` names a reading of a citation that reads the work is placed at that citation's sentence
(§III.4), so the rule finds the sentence by the Anchor's `sentence_of`: `sentenceOfFor(w, m)` reads the
named Verify's `anchor` off the record (`SELECT anchor FROM verify WHERE event_id = ?`) — a field, not
the act pairing. A second lens's mint naming its own refuting reading of a sentence an open gap of the
class already holds gets `oneGapRefusal` (`:150`), in its existing words (R8).

**The reading act [DELETE]:** `AppendReading`, `readingAct`, `pairs`, `errReadingHeld`,
`requireReadingCarrier`, `readingClassRefusal`, `requireStandingCarrier`, `requireReadingNamesCarrier`
(`reading.go:59–183`, `:259–269`), and their callers: `appendBodies`' `act` parameter and held branch
(`record.go:674–720`), `appendAct`'s `act` (`:744`, `:774`, `:782`), `writeView.act`
(`writeview.go:30–34`), `Replay.WriteBodies` and `readingActOf` (`replaywriter.go:130–207`, now no caller:
§III.6), the `act judgement` list's "a reading that names no gap" (`actjudge.go:20`). `readingSentenceIn`
(`reading.go:185`) stays: the found-source arm of `OWN_READING_REFUTES` uses it to ask whether the
reading's sentence is in the report. `ReadingGapClasses` is deleted (F5); `FindsAgainst` stays under F5.

**The mint budget [MODIFY] `mintbudget.go` (F6 of #1647 as narrowed 2026-11-03, unchanged in meaning):**
a mint whose `rests_on` names a Verify by the minting seat WITH an anchor is admitted at any standing
(replaces `w.act.pairs(m) && …`, `:255`); `ownAnswersAt` (`:316–340`) selects
`m.gap_id FROM mint m JOIN verify v ON v.event_id = m.rests_on JOIN events e ON e.id = m.event_id WHERE
e.seat_id = ? AND COALESCE(v.anchor,'') <> ''` (and the Verify's own seat equals the mint's — the
predicate already required it). The comments `:36–44`, `mintAnswerOutsideBudget` (`:303`) and the help
sentence say "a gap resting on your reading of a citation".

**Estoppel [MODIFY] `cli/lens/mint.go:130–148`:** the 2026-10-31 exemption ("Exempt, stated") moves
with the gap: the estoppel guard is not attached where `--rests-on` names a Verify. Keyed on the field,
never the class.

### III.3 `lens verify` records only — `internal/cli/lens/verify.go` [MODIFY] (contract: CLI flag set)

| Site | Change |
|---|---|
| package comment `:89–95` ("# A reading that finds a defect mints its gap") | rewritten: a reading records what the source did; a class that rests on a reading is minted by `mint` naming it |
| `gapFlagsRequiredUnless` `:104`, `readingRequiredGapFlags` `:108`, `readingGapFlags` `:111`, `var gf` `:114`, `gf.register(c, readingSubject)` `:160`, the `RequiredUnless` loop after it | deleted: the gap flags are unknown flags on `verify` |
| `seat.Supplies` for `gap_id`, `class_material`, `carried_by`, `carrier_minted` `:155–158` | deleted |
| `writeVerify` `:393–431` (`classes`, `backing`, the gap-flag refusals, `gf.fill`, the class check) | the reading is written alone; `markedSentenceOf` stays or goes per F6 |
| `recordReading` `:437–540` | the `gap != nil` branch deleted (Anchor, guards, `AppendReading`); the result's `Gap` deleted; the result gains `event` (the reading's events.id) — what `--rests-on` names |
| `givenGapFlags` `:545`, `mintsNothing` `:556`, `classWords` `:579`, `uncitedSentence` `:600`, `readingGap` `:684–705`, `sentenceOfRefusal` `:798` | deleted (`refuseMarked` `:617` per F6) |
| `failedCheckWords.words` `:720–745` | "gap X carries it to blue" reads the link helper (§III.5); a found-against check of this reading's own with no gap says the gap is the lens's to mint, naming the reading's event and the class(es) whose prerequisite it meets |
| `verifyResult.Human` | names the event: `recorded as event N` |

`mint.go`'s `readingSubject` (`:251–256`) and `record.ReadingClassField` (`enums.go:344–360`) are
deleted; `gapFlags` (`mint.go:228`) is `mint`'s alone, and its comment ("shared by `mint` and `verify`")
says so.

### III.4 `lens mint` names the event — `internal/cli/lens/mint.go` [MODIFY] (contract: CLI flag set) — F7

Under F7's recommendation:

- New flag `--rests-on <event>` (int), stored in `Mint.rests_on`. `seat.KeysOn` unchanged; the keyed
  retry compares it (`mintAsserted`).
- Where the named event is a Verify (any class), the gap's subject is derived from it, by identity:
  - a reading of a citation that reads the work → an Anchor with `sentence_of` the citation, `Mint.location`
    the citation's sentence as `citedSentence` reads it (as `recordReading` places today, `verify.go:449–462`,
    with `PlacesToSentenceOf` and `sentenceOfRefusal` moving here);
  - a reading of a citation that reads nothing (`not_the_work`) → `about_kind anchor`, `about_ref` the
    citation, no Anchor — the shape that answers the unread check (`failedcheck.go:30–35`);
  - a reading of a source the lens found → the reading's `claim` as the quote (`seat.Places`).
  `--quote` and `--about-kind/--about` beside such a `--rests-on` are refused ("one subject").
- Where the named event is not a Verify (class 2, 23), the subject is given as today.
- The mint's result echoes `rests_on`.

### III.5 The link readers — one statement each side [MODIFY]

**SQL.** If #1739 has merged: its view `carrying_reading` (`views.go`, diff `14338–14342`) is redefined —
its own comment anticipates this ("the day the link is stated from the gap's side instead of the
reading's this view is the one statement that changes"):

```sql
CREATE VIEW "carrying_reading" AS
SELECT m."gap_id", v."anchor", e."seat_id", v."event_id"
FROM "mint" m JOIN "verify" v ON v."event_id" = m."rests_on" JOIN "events" e ON e."id" = v."event_id";
```

If #1739 has not merged, Part 2 adds this view, and #1739 rebases onto it (§III.8).

**Go.** One helper, `record.ReadingOfGap(evs) map[gapID]*Verify` / `GapsOnReading` — the stream fold every
Go reader below calls instead of `GetCarriedBy()`:

| Reader | Today | Becomes |
|---|---|---|
| `failedcheck.go:111–121` (`addFoundAgainst`) | `v.GetCarriedBy()` | the gap whose Mint rests on `c.ev`'s event, then `carrierTree` as now |
| `failedcheck.go:365–378` (`repeatReadingRefusal`, found-source arm) | `prev.GetCarriedBy()` | the gap resting on `prev` |
| `failedcheck.go:40–48`, `:80`, `:426–428` comments | "minted in the reading's own act and named by the reading's carried_by" | "named by a mint's rests_on" |
| `mintbudget.go:255`, `:313–340` | `carrier_minted` | §III.2 |
| `reading.go:217–258` (`requireNotARetriedReading`'s "gap X carries it") | `v.carried_by` column | the gap resting on the stored reading |
| `evidenceview.go:262–283`, `:375` (`EvidenceVerificationJSON.CarriedBy`, JSON `carried_by`) | stored field | the helper; the JSON key keeps its name — the relation "a gap carries a reading" is unchanged, only where it is stored moves — and the row gains `event` |
| `report/assemble.go:970`, `:1078–1100` (`readingsCarried`, docket.md's source line, #1694 F12) | one line per Verify naming the gap | the one reading the gap's Mint names |
| `seatprobe/seatprobe.go:206–211` | `v.GetCarrierMinted()` | Mints by the seat whose `rests_on` is a Verify |
| `cli/lens/verify.go:503–506` | stored `carried_by` | deleted (§III.3) |
| `capture/capture.go:483` (`EvidenceVerificationJSON.FindsAgainst`) | `FindsAgainst` via `gap_classes` | F5's predicate |
| `anchorchecks.go:144`, `:259` (`holds: reads && !FindsAgainst(b)`) | same | F5's predicate |

**`requireFailedCheckAnswered`** (`failedcheck.go:390–425`) is unchanged in logic; its carriers come
from the helper. Its refusal text and `close.md:13` lose "your gap another lens's reading named" (no
reading names a gap; a second lens's reading is a check with no gap).

### III.6 `migrate` — one translation (contract: record migration)

**What an archived record can hold.** Only records written at epochs 55 (#1694's branch numbering)
through E−1 hold a Verify with `carried_by`. Two shapes:

1. **An in-act reading** (`carrier_minted` 1): Mint m, Anchor a, Verify v, adjacent, one seat
   (`readingActs`, `migrate/replay.go:866–897`, reads these SOURCE fields and stays).
2. **A held reading** (`carrier_minted` 0): v names an open gap another act minted.

**Translation (one, in `migrate/replay.go`; `writeReadingAct` `:914` rewritten, `failAct` `:244–260`
kept for the act's all-or-nothing refusal):**

- Shape 1: write v FIRST, stamped with m's archived stamp (the three were one transaction; this keeps
  `events.id` order and stamp order the same — the replay writes "under the ORIGINAL clock"); then m
  with `rests_on = newID[v]`; then a. Translated as `Stale.through` is (`replay.go:391–399`): a
  `rests_on` whose event wrote nothing is a refusal naming both events.
- Shape 2: v comes forward with no link (its fields are gone). **Residue, stated:** its found-against
  check has no gap carrying it; it holds the PASS until blue edits or retires or its reader's own
  backing re-read, as a migrated pre-#1694 reading does (`failedcheck.go:47–48`). Eiffel: 1 (event 798).
- `remap.idFields` loses `Verify.carried_by` (`remap.go:116`); `rests_on` is an int64 and is not in any
  of the three string lists (`TestEveryStringFieldIsInOneList`).

**Residue, stated (no fill, ruled pattern of #1647 §III.9):** every archived mint of an R7 class that
names no event comes forward naming none — 46 in 27 archives (§II) plus, in the Eiffel archive, G-b2ba7d66
(32, plain) and the 4 class-2 mints. The replay does not judge (F4), so none is refused.

**The archived run [Part 0, NEW `run-archive/2026-11-04_eiffel-tower-height.tar.gz`].** No archive under
`run-archive/` holds the shape the translation reads (§II). The #1694 smoke record does, migrates with
zero refusals at epoch 61, and is the test's input. Committing it is an outward change to the archive set
and is asked before Part 0 lands (semantic consent). Adding an archive re-renders every
archive-iterating golden (§III.9 row G-arch).

### III.7 Agent-facing text [MODIFY / REGENERATE] (contract: prompts and constitutions)

Every sentence below is rewritten to the new model; generated copies are regenerated, never edited.

| Source | Lines | Says today | Becomes |
|---|---|---|---|
| `cli/seat/help/verify.md` | `:15`, `:17`, `:19`, `:23`, `:25`, `:27` | the reading mints its gap in the same act; where it stands; one-gap holds the reading; estoppel exemption; `--url` mints by outcome, budget refuses the reading whole; result names `gap` | a reading records what the source did and prints its event; the gap is `mint`'s, naming that event; the found-against check holds the PASS with or without a gap; at the cap the reading is recorded, the mint is what is refused |
| `cli/seat/help/mint.md` | `:9` (one-gap), `:13` (budget: "a gap your reading of a citation mints in its own act"), `:15` ("A reading that finds the source against its sentence mints its gap in the same act") | | the class's prerequisite (rendered per class from the declaration, below), `--rests-on`, the subject derived from a reading, the exemptions keyed on "a gap resting on your reading of a citation", estoppel exemption |
| `cli/seat/help/close.md` | `:13` | "your gap another lens's reading named" | the gap resting on the reading, and its successors |
| `cli/seat/verbs.go` | `:302` (`show evidence` help) | `carried_by` from the reading | the gap whose mint rests on the reading; `event` |
| enum help (`cli/enumhelp`) | `mint --help`'s class list | means / neighbor / distinguisher / material default | + "minted only on: <prerequisite means>" per declaring class, from the option |
| `record.proto` `SourceOutcome` `means` | `:529–541` | "…mints a gap … in the same act" | what the reading records; never "mints" |
| `scripts/agentgen/src/fragments/lens-audit-and-mint.md` | `:6`, `:8` | "records its gap in the same act" | the reading records; the lens mints the gap that rests on it, outside its budget for a citation |
| `scripts/agentgen/src/skills/research-protocol/SKILL.md` | `:16–22` | "the reading records a gap … in the same act" | the reading records; a gap resting on it is minted |
| `skills/research-protocol/scripts/debate.js` (AUTHORED region) | `:883` (`ledgerClause`: "your reading records the gap") | | "you mint the gap resting on your reading" (an act, no verb path) |
| `tools/internal/terms/terms.json` | `reading` `:762–763`; `failed check` `:1091–1092` | "mints the gap that carries it in the same act or names the open gap…" | the reading records; a gap rests on it by its mint; then `go -C scripts run ./vocabdoc` (`docs/vocabulary.md:463`, `:631`) |
| `docs/citations-and-bibliography.md` | `:171`, `:187–188`, `:203` | in-act mint, `carried_by` | the new model |
| `docs/seat-command-triggers.md` | `:55` | `lens verify` row | records only |
| GENERATED: `agents/*.md` (12, 3 hits each: the inlined manual), `skills/adversarial-audit/SKILL.md`, `skills/research-protocol/SKILL.md`, `help_gen.go`, `annotations_gen.go` (`:186` out_of_scope, `:204–208` outcomes) | | | `go -C scripts run ./agentgen`; `go -C plugins/frank-exchange-of-views/tools run ./devcmd/cmdtreegen`; `… ./devcmd/vocabgen` |
| `debate.js` generated regions | `:105–123`, `:373–451` | carry no GapClass, SourceOutcome or reading word (grep, §III.9 row A9) | regenerated by `enginegen`; expected unchanged |

AFTER regenerating, `verify --help`, `mint --help` and `close --help` are each read before and after,
one at a time, for semantic loss (CLAUDE.md, agent-facing text).

### III.8 Seam with #1739 (1a′, draft, `feov/anchor-lifecycle`)

- **The view.** #1739's `carrying_reading` is the one SQL statement of the link; its four 1a′ readers
  (issue comment 4) read it: `evidence_answered` arms 3 and 4 (`own_check`, `evidence_met`, diff
  `14420–14500`); the reaching retire in `gapActsOf` and the lens work-list item (`blueAnswers` first
  arm, `gapNewsOf` row); `answers_retired_event` and S12/S15's wording (`compute-and-show.md:122`,
  `:125`, `:581–582`, `:625`, `:762`, `:830`, `:950–959`). They change with the view's one statement.
- **The claim "one helper" does not hold for Go readers.** #1739 leaves every Go reader on main reading
  `GetCarriedBy()` directly (§III.5's table) and ADDS three in its fuzz: `lifecycleseed_test.go`
  (diff line 18135) and `sourcedemandseed_test.go` (18931–18935), both driving `verify` with gap flags.
  Those seeds are rewritten by whichever PR lands second.
- **Order.** Either order works; the second to land carries the other's sites. If #1739 lands first,
  Part 2 redefines its view and rewrites its two seeds. If Part 2 lands first, #1739 rebases onto
  Part 2's view and helper and drops its gap-flag `verify` drives.
- **compute-and-show text.** `compute-and-show.md:1252` ("the evidence lens's `verify --as refutes` of the
  crop's citation mints its gap at the sentence, carried by `Verify.carried_by`") and `:1110` ("no
  archived Verify carries `carried_by`" — false once Part 0 commits the Eiffel archive) are rewritten in
  that plan by Part 2's PR.
- **Epoch collision.** Both write 61 today (§III.1).

### III.9 Consumer census (re-run on this tree, 2026-11-05)

Commands (all from the worktree root; `F=plugins/frank-exchange-of-views`):
`grep -rnE 'carried_by|CarriedBy|carrier_minted|CarrierMinted' -I . | grep -v '^./run-archive\|record.pb.go'`
(356 lines; 137 in `plans/contradiction-is-a-gap.md`, 21 in `plans/compute-and-show.md`);
`grep -rnE 'GapClassesOf|FoundSourceGapClassesOf|ReadingGapClasses|FindsAgainst|ReadingClassField' --include=*.go $F/tools`;
`grep -rnE "out_of_scope is an inquiry|OUT_OF_SCOPE" --include=*.go $F/tools`;
the phrase sweep `grep -rlE "your reading records the gap|records its gap in the same act|mints its gap in the same act|records the gap, and blue" $F scripts`;
the test sweep in row T-sweep. Plans and `run-archive/` are not carriers.

**Code (non-test), 27 sites:**

| ID | Site | Part |
|---|---|---|
| C1 | `record.proto:152–165` (options 50014/15), `:529–541` (outcomes), `:1425–1434` (Verify fields), Mint `:882`, GapClass `:384–484` | 1, 2 |
| C2 | `recordpb/gapclass.go:63–75` (`GapClassesOf`, `FoundSourceGapClassesOf`) | 2 |
| C3 | `recordpb/annotations.go:37–40`, `:90` | 1, 2 |
| C4 | `recordpb/annotations_gen.go:186`, `:204–208` (generated) | 1, 2 |
| C5 | `devcmd/vocabgen/main.go:169–170` | 1, 2 |
| C6 | `record/enums.go:344–360` (`ReadingClassField`) | 2 |
| C7 | `record/reading.go:19–183`, `:217–269` | 2 |
| C8 | `record/record.go:674–720`, `:744`, `:774`, `:782`, `:1898–1902` | 1 (`:1898–1902`), 2 |
| C9 | `record/actjudge.go:20`, `:36–48` | 1, 2 |
| C10 | `record/writeview.go:30–34` | 2 |
| C11 | `record/onegapperclass.go:40–45`, `:91–108` | 2 |
| C12 | `record/mintbudget.go:36–44`, `:255`, `:303`, `:313–340` | 2 |
| C13 | `record/failedcheck.go:40–48`, `:80–82`, `:111–121`, `:307–380`, `:426–428` | 2 |
| C14 | `record/anchorchecks.go:144`, `:259` | 2 (F5) |
| C15 | `record/evidenceview.go:262–283`, `:354`, `:375` | 2 |
| C16 | `record/replaywriter.go:130–207` | 2 |
| C17 | `record/migrate/replay.go:127–130`, `:236–260`, `:844–960` | 2 |
| C18 | `record/migrate/remap.go:116` | 2 |
| C19 | `cli/lens/verify.go` (§III.3 table) | 2 |
| C20 | `cli/lens/mint.go:130–149` (estoppel guard), `:225–260` (`gapFlags`, `readingSubject`), flag registration `:205–220` | 2 |
| C21 | `cli/seat/verbs.go:302` | 2 |
| C22 | `report/assemble.go:970`, `:1078–1100` | 2 |
| C23 | `seatprobe/seatprobe.go:206–211` | 2 |
| C24 | `capture/capture.go:483` | 2 (F5) |
| C25 | `record/recordsql/decode_gen.go:1471`, `:1539–1542` (generated by `protogen`) | 2 |
| C26 | new: `record/mintprerequisite.go`; `recordpb` accessor `MintPrerequisitesOf`, `SatisfyingAct` | 1 |
| C27 | `releasegate/fuzz` — §III.10 | 1, 2, 3 |

**Tests that drive the in-act reading or the hand check (T-sweep: files matching
`AppendReading|CarriedBy|CarrierMinted|ReadingGapClasses|citeReading\(|mustRead\(|readingGap|\.Gap\.ID|GapClassesOf|ReadingClassField`
under `internal`, `releasegate`), 26 files, each rewritten or deleted row by row in Part 2:**
`internal/capture/capture_test.go` (4), `cli/enum_help_test.go` (1), `cli/keyedretry_test.go` (4),
`cli/m23corroboration_test.go` (2), `cli/mintreason_test.go` (11), `cli/readingact_test.go` (55),
`cli/readingattribution_test.go` (4), `cli/readingcap_test.go` (3), `cli/readingcopy_test.go` (16),
`cli/readinghelp_test.go` (1), `cli/readinghelpers_test.go` (2), `cli/readingmigrate_test.go` (6),
`cli/readingplace_test.go` (16), `cli/readingsource_test.go` (5), `cli/selector_test.go` (1),
`record/foundagainst_test.go` (4), `record/foundsource_test.go` (12),
`record/migrate/readingsforward_test.go` (2), `record/onegapplacement_test.go` (1),
`record/reopened_test.go` (4), `record/replaywriter_test.go` (3), and the 5 fuzz files of §III.10.
Plus the out_of_scope tests: `record/inquiry_test.go:429–437`, `cli/mintsubject_test.go`.
`readingmigrate_test.go`'s two tests (`:54`, `:186`) write a record AT THIS EPOCH with in-act readings;
after Part 2 that shape cannot be written live, so they are deleted and T19 replaces them on the archive.

**Goldens (G), 9 files + the archive set:**

| ID | File | Why |
|---|---|---|
| G1 | `internal/difftest/testdata/manual_runs_every_help_page_live.golden` | verify, mint, close, show evidence pages (18 hits) |
| G2 | `internal/difftest/testdata/gap_class_enum.golden` | the prerequisite line per class |
| G3 | `internal/difftest/testdata/error_catalogue.golden` | out_of_scope refusal, reading refusals, new prerequisite refusals |
| G4 | `internal/record/testdata/plans.golden` | `:1124`, `:1132`, `:1172`, `:1180`, `:3342`, `:3755`, `:3899` (verify INSERTs, budget and retry SELECTs) |
| G5 | `internal/record/recordsql/testdata/schema.sql` (+ `schema.epoch.json`) | `:784–792` verify columns; mint `rests_on` |
| G6 | `internal/record/migrate/testdata/prose-rewrites.golden` | 4 hits |
| G7 | `tests/simulator/testdata/prompt-red-lens-evidence.golden`, `prompt-red-lens-evidence-engaged.golden` | `debate.js:883` |
| G8 | `internal/terms` generated definitions; `docs/vocabulary.md` | terms rows |
| G-arch | every archive-iterating golden under `internal/record/migrate/testdata/` (`archived_gaps.golden`, `archived_renders.golden`, `archived_gap_rewrites.golden`, `archived_inquiry_methods.golden`, `prose-rewrites.golden`) | Part 0 adds an archive |

**Agent-facing (A), 13 sources:** A1 `help/verify.md`; A2 `help/mint.md`; A3 `help/close.md`; A4
`seat/verbs.go:302`; A5 `record.proto` outcome `means`; A6 enum help rendering (`cli/enumhelp`); A7
`agentgen/src/fragments/lens-audit-and-mint.md:6,8`; A8 `agentgen/src/skills/research-protocol/SKILL.md:16–22`;
A9 `debate.js:883` (authored; generated regions `:105–123`, `:373–451` hold no class/outcome/reading word —
grep `claim_contradicts_source|out_of_scope|same act|carried_by` on the file: 0 hits); A10
`terms.json:762`, `:1091`; A11 `docs/citations-and-bibliography.md:171,187–188,203`; A12
`docs/seat-command-triggers.md:55`; A13 the generated set (12 agents, 2 preloaded skills, `help_gen.go`,
`annotations_gen.go`, `vocabulary.md`).

**Views (V), 2:** V1 `carrying_reading` (#1739's, or new in Part 2); V2 `schema.sql`'s `verify` table and
`mint` table (G5). On main no view reads `carried_by` (`grep carried_by recordsql/views.go`: 0).

**1a′ readers (L), 4 (issue comment 4), all through V1:** L1 `evidence_answered` arm 3; L2 arm 4 /
`own_check`; L3 the reaching retire in `gapActsOf` + `blueAnswers`' first arm + the `gapNewsOf` row; L4
`answers_retired_event` and S12/S15's wording. Plus the two #1739 fuzz seeds that read the field directly
(§III.8).

**Census totals:** 27 code sites; 26 test files driving the reading act + 2 out_of_scope test files; 5
fuzz files + 4 tally fields + the class drive (§III.10); 9 golden files + the archive-golden set; 13
agent-facing sources (+ the generated set); 2 views; 4 1a′ readers + 2 #1739 seeds.

### III.10 Release fuzz — `releasegate/fuzz/` [MODIFY] (CI only)

| Site | Today | Becomes |
|---|---|---|
| `fuzz_test.go:139–225` (`gapClassesByMateriality`, `classDrive`, `unlandedClassDrive`) | each `always` class plain-minted once per sweep; 2, 23, 25, 27, 32 are `always` | for a class whose declaration is non-empty, the drive first performs the declared prerequisite's act (read off `MintPrerequisitesOf`, never a class list), then mints naming it; the gate stays "every driven class landed" |
| `fuzz_test.go:3770–3800` (`outOfScopeClass` arm) | inquiry subject chosen by hand for 31 | the same generic prerequisite step (`ABOUT_AN_INQUIRY` → pick an inquiry) |
| `fuzz_test.go:1120–1140` (quoted found-source reading with `FoundSourceGapClassesOf` gap flags) | gap flags on `verify` | `verify` alone, then (on a coin) `mint --rests-on` |
| `reading_test.go` (37 hits: `readingClasses` `:32`, `readingGapArgs` `:61`, `checkReading` `:144`, `readCitation` `:278`, `repeatReadingRefused` `:379`, `foundAgainstCleared` `:402`, `checkCarriedBy` `:460`, `notTheWorkReading` `:476`, `readFoundSource` `:579`, `readAnotherSource` `:717`, `readingTally` `:830`, `readRepeatedSentence` `:1100`, `readOverPrescription` `:1269`, `readNotTheWorkFound` `:1334`) | in-act mint, held readings, `carried_by` checks | reading then mint; held → the second lens's mint refused by the one-gap rule; checks read `rests_on` |
| `foundagainstseed_test.go` (`:146`, `:285–297`, `:607–636`, `:799`) | forced chain via in-act mint | forced chain: reading, then mint naming it |
| `failedcheckstate_test.go:44–62`, `:161–163`, `:208–210` | `verify … --class claim_contradicts_source` | reading, then mint |
| `anothersource_test.go:23`, `:47`, `:80–82` | second source held by the first's gap | second source's mint refused by one-gap |
| tally fields `sweeptally_test.go:203–215`, `:625–634`, `:759–775`, `:1120–1135` | `readings_minted`, `readings_held`, `found_source_gaps`, `found_source_budget_refused` | deleted: `readings_minted`, `readings_held`; renamed in meaning, so replaced (no old word kept): `mints_resting_on_readings`, `second_reading_mints_refused`, `found_source_mints_budget_refused`, and `prerequisite_refused` (map by MintPrerequisite value) — each with a zero-is-false-green gate |
| new drive | — | for each declaring class, one mint WITHOUT its prerequisite per sweep, expecting the refusal naming the declared act; tallied in `prerequisite_refused` |

### III.11 Required tests (each names how it fails if the change is wrong; driven through the CLI as a seat runs it — `run(t, "mint", "--run", …)` — except the descriptor tests)

**Part 1 (mechanism + 31):**

- **T1** `TestEveryMintPrerequisiteHasAnArm` (record): every non-zero `MintPrerequisite` value has an entry
  in `prerequisiteArms`, a non-empty `means` and `satisfying_act`. *Fails if* a value is added with no
  arm — the generic check would otherwise skip it and admit every mint of a declaring class.
- **T2** `TestAnOutOfScopeGapNotAboutAnInquiryIsRefusedNamingTheAct` (cli): `mint --class out_of_scope
  --quote <sentence>` is refused, the refusal names `--about-kind inquiry`, and the run's event count is
  unchanged; `mint --class out_of_scope --about-kind inquiry --about Q-…` lands. *Fails if* the fold
  drops the rule (first arm lands) or the rule over-refuses (second arm refused).
- **T3** `TestTheRefusalIsTheDeclarations` (record): for each class, the refusal T2's shape produces names
  exactly the `satisfying_act` of the class's declared values, read from the descriptor; and a mint of a
  class with an empty declaration is never refused by `requireMintPrerequisite`. *Fails if* a class-keyed
  string or branch exists (its wording would not be the declaration's). §V step 9 also deletes 31's
  option by hand and expects T2 red.
- **T4** goldens G2, G3: `mint --help` prints 31's prerequisite; the error catalogue holds the new refusal.
  *Fails if* the rendering is absent.

**Part 2 (readings, record event, epoch E, migrate):**

- **T5** `verify … --class claim_contradicts_source` → unknown flag, exit non-zero, no event; same for each
  of the ten gap flags. *Fails if* any gap flag survives on `verify`.
- **T6** a `refutes` reading of citation C writes one Verify and nothing else (event count +1; no Mint, no
  Anchor); its JSON carries `event`; `PassBlockers` holds a found-against failed check on C with
  `AnsweredBy` "". *Fails if* the reading still mints, or the check needs a gap to exist.
- **T7** `mint --class claim_contradicts_source --quote <C's sentence>` with no `--rests-on` → refused
  naming `verify`; nothing written. *Fails if* a plain 32 mint lands (the Eiffel G-b2ba7d66 shape).
- **T8** each predicate arm refuses its own wrong event: another seat's refuting reading; own `supports`;
  own `refutes` since replaced by own later backing reading (not standing); a Cite event; an id beyond
  the record; own `refutes` whose sentence an edit has since changed. *Fails if* any arm is lax — one
  row per condition, so deleting a condition turns exactly one row red.
- **T9** own refuting reading of C, `mint --class claim_contradicts_source --rests-on N` lands: stored
  `rests_on` = N, Anchor `sentence_of` C, the gap at the occurrence C's token marks where the sentence's
  text appears twice in the report (the `readingplace_test` fixture); the found-against check's
  `AnsweredBy` is the gap; at the lens's budget cap it is still admitted and does not count. *Fails if*
  placement goes by text (the repeated sentence is refused as ambiguous), the link reader still reads
  the Verify (no carrier found), or the exemption is lost (refused at cap).
- **T10** found source: `verify --url … --as refutes` records at the cap; `mint --rests-on` it lands at the
  quote below the cap and is refused at the cap naming the budget. *Fails if* F6-as-narrowed is inverted.
- **T11** lens B's refuting reading of C records; B's `mint --rests-on <B's reading>` is refused with
  `oneGapRefusal`'s words naming A's gap; nothing written. *Fails if* a held arm survives or a second gap
  lands.
- **T12** class 25: `--rests-on <own absent reading of C>` lands at C's sentence; `<own not_the_work
  reading of C>` lands `about_kind anchor`/`about_ref C` and makes C's unread check `owed()` false;
  `<own refutes reading>` is refused naming outcome `absent` or a copy not the work. *Fails if* either
  alternative of the OR is missing, or the not_the_work gap does not answer the unread check.
- **T13** class 2: without `--rests-on` refused naming the event the report contradicts; with an existing
  event (an inquiry fate) lands with `--quote`; with an id beyond the record refused. *Fails if* the field
  is optional or unchecked.
- **T14** with T9's gap open and C's check standing, `close --disposition repaired` is refused naming C;
  after the reader's own backing re-read it is admitted. *Fails if* `requireFailedCheckAnswered` lost the
  carrier (close admitted over a standing check).
- **T15** estoppel: blue applies a red `--new` verbatim; a mint resting on a refuting reading of that text
  lands with no estoppel log; a mint of a non-declaring class quoting it is refused by estoppel. *Fails if*
  the exemption is dropped or widened.
- **T16** docket.md golden: the gap's source line names the reading its mint rests on. `show evidence`
  JSON: the reading's row has `carried_by` = the gap and `event` = N.
- **T17** the mint result echoes `rests_on`; `seatprobe` counts the gap as the seat's (fixture).
- **T18** schema: `verify` has no `carried_by`/`carrier_minted`; `mint.rests_on` references `events(id)`
  (G5); opening an epoch E−1 record is refused naming `migrate` (the existing epoch tests, re-armed by
  the bump).
- **T19** `TestTheEiffelReadingActsComeForwardAsReadingThenMint` on
  `run-archive/2026-11-04_eiffel-tower-height.tar.gz`: zero refusals; for each archived carrying reading
  (167, 182, 542, 791, 794, 797) the Mint of its gap has `rests_on` = that reading's migrated id, and the
  named event is a Verify by the minting seat with the archived outcome and anchor; the migrated Verify
  precedes its Mint and their stamps are non-decreasing in id order; reading 798 is named by no Mint;
  G-b2ba7d66 and the 4 class-2 mints rest on nothing; `PassBlockers` on the migrated record holds the
  same found-against checks as the source's, with the same `AnsweredBy`, except 798's (none). *Fails if*
  `rests_on` is copied untranslated (the named event is not that reading), the order is kept (Mint before
  its own prerequisite), or a held reading is linked by position.
- **T20** the residue count: migrating every `run-archive/` SQLite archive leaves 46 R7-class mints with no
  `rests_on` and zero refusals attributable to the prerequisite. *Fails if* the replay judges the
  prerequisite (refusals appear) — the F4 consequence made measurable.

**Part 3 (23, 27 — per F1, F2):**

- **T21** class 23 under F1(a): `mint --class figure_recount_fails --about-kind anchor --about K-…` without
  `--rests-on` refused naming `reproduce`; with a reproduce of K-… whose `rerun` is `reproduced` refused;
  with `differs` lands; a mint quoting a sentence no computation marks lands with no `--rests-on`, and
  `mint --help` says so. *Fails if* the condition is missing (the source-figure case is refused with no act
  to satisfy it) or ignored (a computation's recount lands with no reproduce).
- **T22** class 27 under F2(a): a reading of C whose stamped copy sha differs from C's `sha256` satisfies
  it; an equal sha refuses naming `verify`. *Fails if* the comparison reads `copy_substitution` only (a
  re-fetched url with new bytes reads as no drift).

## IV. Phases

Parts are independently shippable where marked; each PR runs §V before merge.

| Part | Content | Ships alone? | Epoch |
|---|---|---|---|
| **0** | commit the Eiffel archive to `run-archive/`; regenerate G-arch; nothing else | yes — data only; asked first (§III.6) | none |
| **1** | `MintPrerequisite`, the two options, `requireMintPrerequisite` with the `ABOUT_AN_INQUIRY` arm, 31 declared, #1721's hand check deleted; T1–T4; fuzz class drive reads the declaration | yes — no schema change; changes only where 31's rule runs (F4) | none |
| **2** | `rests_on`; arms for 32, 25, 2; `verify` records only; mint `--rests-on`; link helper and view; delete the reading act, `carried_by`, `carrier_minted`, the outcome options (F5); migrate translation; all agent-facing text; fuzz rewrite; T5–T20 | no — the reversal, its readers and its migration are one change at one epoch | E |
| **3** | 23 (F1) and 27 (F2) arms and declarations; T21–T22 | yes, after 2 | E+1 if F2(a) |

Within Part 2, one agent per phase, each leaving the tree green: (2a) proto + generated + the generic
arms + the helper/view, readers switched, `verify` still minting (both paths compile); (2b) `verify`
records only + mint `--rests-on` + deletions; (2c) migrate + T19/T20; (2d) agent-facing text + goldens;
(2e) fuzz. The PR is not opened until 2e.

## V. Verification Plan

Run in the worktree, after `unset $(env | grep -o '^FEOV_[A-Z_]*')`. Each check names what re-arms it.

1. `go -C scripts run ./protogen && git diff --exit-code -- '*record.pb.go' '*record.proto.sha256' '*decode_gen.go'` → no diff after regeneration is committed. *Re-armed by:* any `record.proto` edit.
2. `go -C scripts run ./schemagen` → `record.EventSchema` = `requirements.json`'s; Part 1: no change. *Re-armed by:* `requirements.json`, any event-shape change.
3. `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/cmdtreegen && go -C scripts run ./agentgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen && go -C scripts run ./vocabdoc && git status --short` → only intended files change; `agentgen -check` / `enginegen -check` clean. *Re-armed by:* help .md, terms.json, agentgen sources, proto options.
4. `UPDATE_GOLDENS=1 go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/difftest/ ./internal/record/... ./internal/report/...` then `git diff --stat` and READ every changed golden line (G1–G8, G-arch). *Re-armed by:* any agent-facing text, schema, archive.
5. `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/... ./internal/cli/... ./internal/report/... ./internal/capture/... ./internal/seatprobe/...` → PASS, read the FAIL count. *Re-armed by:* any Go change in Parts 1–3.
6. `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestTheEiffel|TestTheResidue' ./internal/record/migrate/` → PASS (T19, T20). *Re-armed by:* `migrate/replay.go`, `remap.go`, the archive.
7. Zero-hit sweeps (Part 2), each must print nothing:
   `grep -rnE 'carried_by|CarriedBy|carrier_minted|CarrierMinted' --include=*.go --include=*.md --include=*.js --include=*.json plugins/frank-exchange-of-views scripts/agentgen | grep -v 'reserved\|carrying_reading\|"carried_by,omitempty"'`;
   `grep -rnE 'GapClassesOf|FoundSourceGapClassesOf|ReadingGapClasses|ReadingClassField|AppendReading|errReadingHeld' --include=*.go plugins/frank-exchange-of-views/tools`;
   `grep -rlE 'your reading records the gap|records its gap in the same act|mints its gap in the same act' plugins/frank-exchange-of-views scripts`;
   `grep -rn 'GapClass_GAP_CLASS_OUT_OF_SCOPE' --include=*.go plugins/frank-exchange-of-views/tools/internal | grep -v _test` (Part 1).
   *Re-armed by:* any text or code edit.
8. Read `verify --help`, `mint --help`, `close --help`, `show evidence --help` before and after, one at a time (`go -C plugins/frank-exchange-of-views/tools run ./cmd/feov-record lens verify --help`, etc.). *Re-armed by:* any help source.
9. Delete-the-row checks, by hand, each reverted after: remove `(mint_prerequisite)` from 31 → T2 red; from 32 → T7 red; remove the `OWN_READING_NOT_THE_WORK` arm entry → T1 and T12 red; delete the `rests_on` translation in migrate → T19 red; delete the budget exemption arm → T9 red. A row that stays green is a finding. *Re-armed by:* any change to the arms, declarations or translation.
10. `go -C scripts run ./check` → every local gate green, tests at `-count=1`; read the fail count, not the exit code. *Re-armed by:* every push.
11. `scripts/universe.sh build` then grep the installed copy for a phrase of the new `verify --help`; one cheap smoke (`--smoke`: one lane, haiku both tiers) and read its record: no Verify has a gap field, each 32/25 mint has `rests_on`, and no plain 32 mint landed. *Re-armed by:* Part 2's merge candidate.
12. **CI only — release fuzz** (`FEOV_RELEASE_GATE` sharded job in `.github/workflows/hooks.yml`, scheduled and on tag): `prerequisite_refused` non-zero for every declared value, `mints_resting_on_readings`, `second_reading_mints_refused`, `found_source_mints_budget_refused` non-zero, every driven class landed. Never run locally. A logic PR waits for one full pass (release-fuzz cadence ruling 2026-10-29). *Re-armed by:* Part 1, 2 or 3 merge candidates.

## Risks

| # | Risk | Mitigation |
|---|---|---|
| K1 | A lens records a refuting reading and never mints: the check holds the PASS with no fix spec for blue | F8; blue is readied by the check; the reading's reason is what blue reads; `verify`'s output tells the lens the gap is its to mint and which class its reading grounds |
| K2 | Seats type a wrong event id | every arm refuses a wrong event loudly, naming what the event is (T8) |
| K3 | The Eiffel record's epoch label 56 is #1694's branch numbering, not main's 56 | measured: main's migrate admits it with 0 refusals (§II); T19 re-proves it at each epoch |
| K4 | Two links for one gap during Part 2's phases | phase 2a keeps `verify` writing both until 2b; the helper reads only `rests_on`; T-sweep files are switched in 2b |
| K5 | #1739 and Part 2 race for the epoch and the view | §III.8; the second rebases; CLAUDE.md's rebase-and-regenerate rule |
| K6 | A class's prerequisite set changes later and archived gaps "violate" it | the replay does not judge (F4); a live mint is judged by the declaration of its day |

## Forks — RULED by gblock 2026-11-05 (options picked; the option wording is the lead agent's)

| Fork | Ruling |
|---|---|
| F1 class 23 | (a) conditional: `REPRODUCE_DIFFERS` where the subject is a computation's; every other 23 mint stays plain and `mint --help` says so; a recount act is a follow-up issue |
| F2 class 27 | (a) stamp the read copy's hash on `Verify`, Part 3 (epoch E+1); 27 is conditional on a citation subject; computation drift stays plain |
| F3 / F7 | (a) one field `Mint.rests_on` (events.id), `--rests-on <event>`; a named reading supplies the subject; `--quote`/`--about` beside it refused |
| F4 | (a) ALL prerequisites, 31 included, in the live act judgement only; replay judges none. gblock asked that both kinds sit in ONE place; the shape check for all refuses the 46 archived mints |
| F5 | (b) delete `gap_classes` and `found_source_gap_classes`; add a `finds_against` bool option on SourceOutcome |
| F6 | (a) the backing reading records; 29 is a plain mint that MAY name its own reading; `refuseMarked` stays |
| F8 | (a) never owed |
| Part 0 | commit the Eiffel smoke archive to `run-archive/` (records and proofs only), its own PR |

The options as drafted follow, kept for the reasons.


**F1 — class 23, a recount of a figure printed in a source (ruled open in issue comment 3).** No act
records a recount of a figure a source prints; `compute` is blue's (`cli/blue/compute.go`) and a lens's
surface has `reproduce` (of blue's computations) and `verify` (what a source did for a sentence). 23's
`means` also covers a recount of the report's own arithmetic, which no act records either.
- (a) **Conditional prerequisite.** 23 declares `REPRODUCE_DIFFERS`, whose arm applies where the gap's
  subject is a computation's (a sentence a `K-` token marks, or `about_kind anchor` naming a `K-`); every
  other 23 mint stays plain, and `mint --help` says so. Cost: one value, one arm with one condition, T21.
  No new act, no epoch.
- (b) **A red recount act.** A lens records a recount as a program run (a red-side computation that places
  no anchor, or a `recount` verb), and 23 declares "a differing reproduce or your own recount". Cost: a
  new verb and event (epoch), a store path, help, constitutions, a fuzz drive; it overlaps
  compute-and-show Part 1b (`compute --input C-…` reads a citation's raw bytes), which a recount of a
  source's table would need.
- (c) **A reading outcome for a recount.** A new `SourceOutcome` value ("the source prints the figure and
  its own arithmetic does not hold"), and 23 accepts that reading. Cost: one value and its facets, but
  `SourceOutcome` means what the source did FOR THE CLAIM, and here the source supports the claim — the
  value would mix two questions on one axis; the recount itself is kept nowhere.
- **Recommendation: (a)** in Part 3, with (b) filed as a follow-up to be weighed after Part 1b ships and a
  smoke run shows source-figure recounts happen.

**F2 — class 27, "a reading of a copy that differs from the cited one" is not evaluable today.** `Verify`
records `copy_substitution` and `copy_url` but no hash of the bytes it read (`sha256`, field 15, is
reserved, `record.proto:1439–1440`); `Cite` records `sha256`. A url re-fetched with new bytes stamps the
same `copy_substitution` and `copy_url` as the cite. And 27's `means` covers "a computation rests on a …
tree's current state", which no reading shows.
- (a) **Stamp the copy's hash on the reading** (new tool-stamped Verify field, name grepped first; the
  cache entry's `Sha`, `fetchcache.go:149`), and declare `READING_OF_A_DIFFERING_COPY` conditional on the
  gap's subject being a citation's: the reading's hash differs from the citation's `sha256`. Computation
  drift stays plain. Cost: one field (epoch E+1), one arm, T22; migrated readings carry no hash and cannot
  ground a 27 mint (stated).
- (b) **No prerequisite for 27 now**; plain mint as today; the fork reopens when (a) is wanted. Cost: none;
  R7's sixth row is not delivered.
- (c) Compare `copy_substitution`/`copy_url` only. Cost: small, but the drift that matters (same url, new
  bytes) reads as "no drift" — a silent zero (facts-are-fields). Not recommended.
- **Recommendation: (a)** as Part 3; (b) if Part 3 is not wanted now.

**F3 — one event field or one per kind.**
- (a) **One field, `Mint.rests_on`** (int64, `events.id`), for every prerequisite that names an event:
  readings (32, 25, 27), the contradicted record event (2), the reproduce (23). One flag, one translation,
  one helper; the declaration says what the event must be. Cost: the field's meaning is "the event this
  class rests on", read through the class.
- (b) **Per kind**: `Mint.reading` (32, 25, 27), `Mint.contradicts` (2), `Mint.reproduce` (23). Cost: three
  fields, three flags, three translations; each arm reads its own field; the generic check needs a
  value→field map.
- **Recommendation: (a).** The ruling's words "names its reading as a field" and "names the record event it
  contradicts, as a field" are both met by one field whose referent the class declares.

**F4 — where the one rule runs.**
- (a) **Live judgement only** (`judgeAct`). Archived mints come forward unjudged (46 + Eiffel's); #1721's
  out_of_scope rule leaves the shape check, so the replay stops asking it (0 archived class-31 mints).
- (b) **One function, two callers**: body-only prerequisites (31) asked in `validate` by every writer,
  record-state ones in `judgeAct`. Cost: a mode flag on the arms; keeps 31's replay coverage.
- (c) Shape check for all. Refuses 46 archived mints on replay — data loss; contradicts migrate's job. Not
  viable.
- **Recommendation: (a).** One home, and the replay already declines to judge every other act judgement.

**F5 — the SourceOutcome options `gap_classes` / `found_source_gap_classes`.** Both state what a
reading's act mints; after R2 nothing does. `FindsAgainst` (`reading.go:34`) and the anchor ledger
(`anchorchecks.go:259`), `capture.go:483` and the failed check read "finds against" through them.
- (a) Delete both; derive "finds against" from the class declarations (an outcome finds against iff some
  class's prerequisite requires a reading with that outcome). Cost: none in schema; couples the ledger's
  question to the mint rule's table.
- (b) **Delete both; add a `finds_against` bool option on SourceOutcome** (`refutes`, `absent`). Cost: one
  option; the ledger reads its own declaration.
- (c) Keep `gap_classes` and reword it as "the classes a reading with this outcome can ground". Cost: a
  second declaration of what the GapClass prerequisites already say — two carriers of one relation.
- **Recommendation: (b).**

**F6 — `unsourced_claim` (29) and the found-source backing reading.** Today a backing `verify --url`
mints 29 in the act (`record.proto:529–534`), and `refuseMarked` (`verify.go:617`) refuses a backing
reading on a marked sentence because that mint would be false. R7/R9 give 29 no prerequisite.
- (a) **The backing reading records; 29 is a plain mint that MAY name a reading** (`--rests-on` admitted on
  any class, validated as an event of the minting seat), so the docket keeps the found source for blue;
  `refuseMarked` stays, refusing a reading of the wrong form. Cost: `rests_on` optional beyond the
  declaring classes; one rule ("an undeclared class may name its own event").
- (b) `--rests-on` only on declaring classes; 29 names nothing; the docket loses the found source;
  `refuseMarked` deleted (nothing false left to prevent). Cost: blue loses the url it would cite.
- (c) Give 29 a prerequisite (own backing reading of a found source). Cost: reverses R9 — a lens could no
  longer mint an uncited claim it found no source for.
- **Recommendation: (a).**

**F7 — how a mint names its event, and the subject.**
- (a) **`--rests-on <event>`**, the reading's event printed by `verify`; a named reading supplies the subject
  by identity (sentence_of / about the anchor / its quote), and `--quote`/`--about` beside it are refused.
  Cost: seats carry an integer from one call to the next; every wrong id is refused (T8).
- (b) **The mint names the subject** (`--about-kind anchor --about C-…`, or `--quote`) and the tool resolves
  the seat's standing qualifying event for it, storing its id. Cost: no id typing; but class 2 still needs
  an event id, a sentence holding two citations or repeated text is ambiguous and refused (gblock's
  placement ruling, 2026-10-30), and the resolution rule is a second statement of "standing".
- **Recommendation: (a).**

**F8 — a found-against check no gap carries.** After R2 a refuting or absent reading of a citation holds
the PASS before any mint (issue: "a refuting reading of a citation still holds the PASS as a failed
check").
- (a) **Never owed** (as today, `failedcheck.go:80`): blue is readied by the check alone; the lens's mint
  is its judgement. Cost: blue may act on the reading's reason without a fix spec.
- (b) **Owed by its reader**, like unread and unsound, outside the budget. Cost: deadlock — a second lens's
  owed gap is refused by the one-gap rule (R8) while the first lens's gap holds the sentence, so it can
  never discharge the debt without an exemption keyed on position.
- **Recommendation: (a).**
