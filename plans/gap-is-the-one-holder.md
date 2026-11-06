# A gap is the one PASS-holder — `verify` records, `mint` of 32/25 stands on a reading (#1727)

> **Status: DRAFT, unaudited.** Written 2026-11-06 on `plan/mint-prerequisites` merged with `origin/main`
> `7d676a887` (EventSchema **64**, `requirements.json:3`; #1739 anchor-lifecycle is MERGED). Replaces
> `plans/mint-prerequisites.md` (renamed with `git mv`; its mechanism — a `MintPrerequisite` option, a
> `rests_on` link, a generic check — was dropped by gblock 2026-11-06). The scope is the ruling quoted in
> §I; its option wording is the lead agent's, not gblock's. Six forks are open (§VI); none is decided
> here. Every census row in §III.9 was re-run by grep on this tree.

## I. Summary & Goals

**Objective.** A reading (`lens verify`) records what one source did for one sentence and holds nothing
by itself. A gap is the ONE thing that holds a PASS for a defect and the one thing blue answers. `mint`
of `claim_contradicts_source` (32) and `citation_figure_misattribution` (25) is refused unless the
minting seat holds its own STANDING reading of the mint's subject with the matching outcome — keyed on
the subject the mint already names, with no new field and no event id. The in-act mint, the
reading-to-gap link (`Verify.carried_by`, `carrier_minted`), the held branch, the `gap_classes` /
`found_source_gap_classes` options, the found-against failed-check state and the budget exemption for
reading-minted gaps are deleted. One epoch bump; `migrate` brings the one archived run that holds the
old shape forward.

**The ruling (gblock, 2026-11-06, issue #1727, last comment), verbatim:**

- A reading (`verify`) only records what one source did for one sentence. It holds nothing by itself.
- A gap is the ONE PASS-holder for a defect, and the one thing blue answers. This reverses #1647's F4
  (found-against as a third failed-check state); the ledger keeps `unread` and `unsound`.
- At the gate, a `refutes`/`absent` reading of a citation that no gap of that class on the sentence
  stands for, and that the same lens has not re-read since, refuses the PASS and routes back to that
  lens as a blocking work-list item: mint, or re-read. A debt, not a state.
- `mint` of `claim_contradicts_source` (32) and `citation_figure_misattribution` (25) is refused unless
  the minting seat holds its own standing reading of the mint's subject (the citation anchor, or the
  sentence for a found source) with the matching outcome. Keyed on the subject the mint already names:
  no new field, no event ids.
- Classes 2, 23, 27, 31 are judgements and stay plain mint; 31's hand check stays. 29's backing
  found-source reading records only.
- The in-act mint, `carried_by`, `carrier_minted`, and the `gap_classes`/`found_source_gap_classes`
  options go; "finds against" is a `finds_against` bool option on SourceOutcome.
- Part 0 stands: commit the Eiffel smoke archive to run-archive/.

**The third bullet, as re-stated by gblock the same day (2026-11-06, relayed; the wording is the
relay's):** an UNCHECKED or UNREAD citation is red's broken promise about quality control and holds the
PASS because its importance is unknown (unchanged). A READ-AND-REFUTED citation is a known fact about
blue's work whose importance the lens grades, so it must NOT hold the PASS by itself; the lens's graded
gap does, by its materiality. The found-against state leaves the ledger; the lens whose reading finds
against OWES a gap of class 32/25 about that subject, or a later re-read with a non-finds-against
outcome — the LENS's duty, modelled on the owed unread/unsound gap, a blocking item on its own work list
that refuses its envelope, and not a PASS blocker. §III.4 designs that debt.

**Why (the value case against the naive version).** The naive version is "delete `carried_by`, let
`mint` take the class as before": it leaves G-b2ba7d66's shape — the Eiffel adversary's plain
`claim_contradicts_source` mint with no reading, its evidence (url, copy, outcome, confidence) lost to
prose in `mint_reason` — fully admitted, which is the defect #1727 was filed on. The previous draft's
version (a link field from gap to reading, a per-class option, a generic check) doubled the concept: the
reading and the gap each held half of one fact, and the audit found the halves needing budget and
estoppel exemptions, multiplicity rules and successor rules to stay consistent. This version keeps ONE
holder (the gap) and ONE observation (the reading) and joins them where a reader needs both on fields
both already carry: the seat, the subject and the class. The join is a question of the record, not a
field a writer must keep in step.

**Success criteria.**

1. `verify --class …` and every other gap flag is an unknown flag; a `verify` writes exactly one
   event, a Verify, whatever its outcome (T1, T2).
2. `mint --class claim_contradicts_source --about-kind anchor --about C` by seat X is refused, nothing
   written, the refusal naming `verify` and C, unless X's standing reading of C is `refutes`; the same
   for 25 with `absent` (or a not-the-work copy, F3); a found-source 32 by `--quote S` is admitted only
   on X's standing `refutes` reading of a found source for S (T3–T8).
3. No Verify names a gap and no gap names a reading, in the proto, the schema, the views, the help, the
   constitutions, the terms registry or the fuzz (§V check 7 zero-hit sweeps).
4. The ledger's failed-check states are `unread` and `unsound`; no PASS blocker or close refusal is
   produced by a `refutes`/`absent` reading (T9, T10). The reading's lens owes a gap or a re-read: a
   blocking item on ITS work list that refuses its envelope, discharged by any open gap of the class on
   the subject, its own non-finds-against re-read, blue's edit or retire (T18–T21).
5. The reading-minted exemption (`carrier_minted`) is gone from the budget; whether the gap a reading
   debt obliges counts is F6 (T11 asserts whichever option is ruled); a 25 gap about a citation whose
   unread check its reader owes is exempt by the existing own-failed-check rule in either case.
6. `feov-migrate` brings the Eiffel archive forward with zero refusals: each archived in-act reading
   becomes its Verify, Mint and Anchor as three events with no link fields; the held reading becomes a
   Verify alone (T13, T14).
7. The release fuzz drives the mint refusal and the admitted mint for both classes and both subject
   forms with no zero tallies (CI only, §V check 12).

**Out of scope.** Any change to a class's meaning; prerequisites for 2, 23, 27, 31 (ruled plain mint;
31's hand check `record.go:1931` stays as it is); #1695 (findings fold into gaps).

## II. Technical Context

**Record shapes today (`origin/main` `7d676a887`, epoch 64).**

- `Verify.carried_by` (21, references `mint.gap_id`) and `carrier_minted` (22), `record.proto:1431–1441`,
  set only by the tool: `AppendReading` (`record/reading.go:92–103`) on an in-act mint, `appendBodies`'
  held arm (`record/record.go:705–716`) when the one-gap rule HOLDS the reading's Mint
  (`onegapperclass.go:91–93`, `errReadingHeld` `reading.go:75`). Read through the `carrying_reading` view
  (`recordsql/views.go:412–420`, index `verify_carried_by` `schema.go:261`), which only `demand_reach`'s
  `'reading'` arm reads (`views.go:493–495`); `demand_reach` feeds the gap view's `evidence_met` arms 2
  and 3 (`views.go:597–621` → `evidence_answered` `:702`), `windowindex.go:104` (`reachedGapsSQL`) and
  `demand.go:83` (`EvidenceDemandOf`). Two Go queries read the column directly: `mintbudget.go:330`
  (`ownAnswersAt`) and `reading.go:225` (`requireNotARetriedReading`).
- `SourceOutcome` (`record.proto:529–545`) declares `gap_classes` (50014, `:158`: refutes → 32, 26;
  absent → 25, 26) and `found_source_gap_classes` (50015, `:165`: supports/bridge/weak → 29; refutes →
  32). Readers: `recordpb.GapClassesOf`/`FoundSourceGapClassesOf` (`recordpb/gapclass.go:63–74`),
  `record.FindsAgainst` (`reading.go:34`: reads the work AND `gap_classes` non-empty),
  `record.ReadingGapClasses` (`reading.go:46`), `record.ReadingClassField` (`enums.go:351`),
  `recordsql.verifyHolds` (`evidencetables.go:84`), `evidenceview.go:269–283`, `capture.go:513`.
- The anchor ledger (`anchorchecks.go`) has six states; `stateFoundAgainst` (`:148`) maps to
  `CheckFoundAgainst` (`failedcheck.go:57`), the third failed-check state, whose blocker is built by
  `failedChecksOf`'s `addFoundAgainst` (`:112–123`) from `v.GetCarriedBy()` and `carrierTree` (`:177`);
  `owed()` is false for it (`:89`); `requireFailedCheckAnswered` (`:395–427`) refuses the close of any
  carrier. Three non-blocker readers read the SAME ledger state: `CitationsFoundAgainst`
  (`anchorchecks.go:315`) → `counts.sources_refuted` (`evidenceview.go:564`); `FailingAnchors`
  (`anchorchecks.go:331`, through `EvidenceTables.Failed = "found_against"`, `evidencetables.go:56`) →
  the capture's assembly screen (`capture.go:484–495`, an audit FAIL when a refuted citation is still in
  the assembled report).
- The mint budget exempts a reading-minted gap of a citation (`mintbudget.go:257` `w.act.pairs(m)`,
  `ownAnswersAt` `:325–345`); the estoppel guard is not attached by `verify` (`mint.go:155–158`,
  `reading.go:89–91`).
- `requireOneGapPerClassAndSubject` (`onegapperclass.go:50`) applies to a quote subject (sentence) and an
  inquiry subject; a gap `--about-kind anchor` is exempt (`:25–30`), so the gap a failed check obliges is
  never turned away.
- `judgeAct`'s Mint arm (`actjudge.go:40–64`): about-anchor live, one-gap, superseded off docket,
  successor carries demand, budget, report voice. The replay never asks it (`actjudge.go:23–25`).
- `mint`'s subject: `--quote S` (location; placed by `seat.AppendPlaced`, `mint.go:164`) or
  `--about-kind anchor|section|inquiry|gap --about <id>` (`mint.go:220–222`;
  `ABOUT_KIND_ANCHOR` `means` at `record.proto:1179`: "a citation or computation, by anchor, whose check
  failed: a computation judged unsound, a citation whose work went unread").

**Measured inputs (the previous draft's §II, re-checked where cheap).**

| Measure | Value | How |
|---|---|---|
| Archived mints of 25/32 after `migrate`'s slug translation (`migrate/gapclass.go:43,47,56` → 25) | 25: 4 mints in 4 runs; 32: 6 in 1 run (Eiffel: 5 in-act, 1 plain G-b2ba7d66); 46 of the previous draft's six classes | python over each `run-archive/*.tar.gz` `records/record.db` (2026-11-05; cost sheet §6, 2026-11-06) |
| Archives under `run-archive/` holding a Verify with `carried_by` | 0 of 28 SQLite archives (newest `2026-10-30_mersenne-31-prime-sum`, pre-#1694) | same scan |
| The Eiffel record | `~/.claude/scratch/contradiction-impl/elicit-adversary/2026-11-04_eiffel-tower-height.tar.gz`, 370 KB; layout `inputs/run-config.json`, `records/record.db`, `computations/<sha>/{acts,dir,rerun}` — the layout of `2026-10-30_mersenne-31-prime-sum.tar.gz`; 59 readings, 7 with `carried_by` (6 in-act: events 167, 182, 542, 791, 794, 797; 1 held: 798) | `tar tzf`; sqlite read (2026-11-05) |
| That record under main's `feov-migrate` at epoch 61 | 1213 events in → 1203 out, 0 refusals | 2026-11-05; re-run at 65 is T13 |
| Epoch | main 64; #1739 took 64; this plan takes **65** | `requirements.json:3` on `origin/main` |

**Constraints.** No backwards compatibility: deleted fields and options are reserved, `migrate` is the
only path that upgrades a record, and the live code accepts and prints only the new form (CLAUDE.md).
Seat prompts name acts, never a flag (`TestTheSeatPromptsNameNoVerb`); refusals and help pages may name
flags. Generated carriers are regenerated, never edited: `agents/*.md` and the two preloaded skills
(`agentgen`), `debate.js`'s two marked regions (`enginegen`), `help_gen.go` (`cmdtreegen`),
`annotations_gen.go` (`vocabgen`), `docs/vocabulary.md` (`vocabdoc`), `record.pb.go`/`decode_gen.go`
(`protogen`), `record.EventSchema` (`schemagen`).

## III. Proposed Changes

### III.0 The design in one table

| # | Decision | Where |
|---|---|---|
| D1 | `Verify` loses `carried_by` and `carrier_minted`; `SourceOutcome` loses `gap_classes` and `found_source_gap_classes` and gains `finds_against` (bool option, true on `refutes` and `absent`); epoch 65 | §III.1 |
| D2 | `verify` records only: no gap flag, no Mint, no Anchor; one event per call | §III.2 |
| D3 | `mint` of 32/25 is a live act judgement, `requireOwnStandingReading`, keyed on the mint's subject: a CITED source's defect names the citation (`--about-kind anchor --about C`), a FOUND source's defect names the sentence (`--quote S`); the seat's standing reading of that subject must carry the class's outcome | §III.3 |
| D4 | `stateFoundAgainst` leaves the ledger: a citation with a `refutes`/`absent` reading is holding or unchecked by the ordinary rules; `CheckFoundAgainst`, its blocker, its close refusal and its carriers go. The per-check `holds` stays (the demand's "check again that holds"); `counts.sources_refuted` and the assembly screen read the standing checks that do not hold, as information | §III.5 |
| D4′ | The reading debt: a lens whose standing reading finds against owes a gap of class 32/25 about the subject or a non-finds-against re-read — a blocking item on its OWN work list (envelope refused while it stands), readying it once per cause; never a PASS blocker; discharged by any seat's open gap of the class on the subject | §III.4 |
| D5 | The one-gap rule gains an anchor arm: one open gap per class about one anchor, except the gap a seat's own owed failed check obliges | §III.3 |
| D6 | A gap carries nothing of its reading by field; the reading behind a gap is one Go helper, `record.ReadingBehind`, joining (seat, subject, class); `carrying_reading` and `demand_reach`'s reading arm are deleted | §III.6 |
| D7 | The budget exemption for reading-minted gaps is deleted: a 32 gap pays like any gap; a 25 gap about a citation whose unread check the minter owes is exempt by the existing own-failed-check rule | §III.3 |
| D8 | `requireNotARepeatReading`'s reads arm is deleted; the unread arm stays | §III.5 |
| D9 | `migrate`: the in-act reading act machinery is deleted; each archived member comes forward as its own event with the two fields dropped; stated residue | §III.7 |

**Alternatives considered for D3's subject form.** (i) Admit `--quote S` for a cited source's 32 too,
the check finding the seat's refuting reading of any citation in S. Rejected: the gap would then reach no
anchor in `demand_reach` (the 'about' arm is the only one left after D6), so its `source` demand could
be met only by blue's new cite, never by retiring the refuted citation or by the reader's backing
re-read — the two remedies the citation ruling names; and the one-gap rule would need the sentence of
an about-less gap. The ruling's own words give the subject: "the citation anchor, or the sentence for a
found source". (ii) A Go-side fold that reconstructs the gap-to-reading link from sentence position for
every reader. Rejected: a relation recovered from position is the shape #1694 replaced with a field
(`reading.go:115`); D6's join is on identities (seat, anchor id, class) for the cited case and on the
located sentence only where the record never had an identity for it (a found source).

### III.1 The record — `internal/record/recordpb/record.proto` [MODIFY] (contract: event schema, epoch 65)

- **`Verify` [DELETE fields]:** `carried_by` (21), `carrier_minted` (22) and their comment `:1431–1441`;
  `reserved 21, 22; reserved "carried_by", "carrier_minted";` beside the existing reserved list.
- **`SourceOutcome` options [DELETE, ADD]:** delete `gap_classes` (50014, `:152–158`) and
  `found_source_gap_classes` (50015, `:160–165`), reserve both numbers with a one-line comment as 50007
  and 50010 are. Add `optional bool finds_against = 50016;` on `EnumValueOptions` (number grepped free;
  the names `finds_against`, `FindsAgainstOf` have 0 hits outside `plans/`): "the source was read and
  says otherwise than the sentence, or nothing of it — the outcome the anchor ledger reads as a reading
  against the sentence, and the outcome a mint of a class defined by such a reading stands on". Set
  `[(finds_against) = true]` on `REFUTES` and `ABSENT`. Each outcome's `means` (`:531–542`) loses its
  "mints … in the same act" clause and says what the reading records (A5 in §III.8).
- **`GapClass` [MODIFY `means` only]:** 25 (`:460`) and 32 (`:483`) each gain one clause: "minted on your
  own standing reading of the citation (`absent`, or a copy that is not the work)" / "minted on your own
  standing reading that refutes: of the citation you name, or of a source you found for the sentence you
  quote". No option, no value change. The class-to-outcome rule lives in Go (§III.3), as 31's does.
- **`AboutKind.ANCHOR` `means` (`:1179`) [MODIFY]:** "…a citation whose work went unread, or whose source
  your reading finds against".
- **Epoch:** `requirements.json` `eventSchema` 64 → **65**; `go -C scripts run ./schemagen`. #1739 took 64
  and is merged, so there is no collision; a PR that lands before this one and bumps takes 65 and this
  plan takes the next.
- **Generated:** `go -C scripts run ./protogen` (`record.pb.go`, `decode_gen.go:1475` verify scan,
  `record.proto.sha256`); `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen`
  (`annotations_gen.go`: the `GapClasses`/`FoundSourceGapClasses` literals go, `FindsAgainst` arrives;
  `recordpb/annotations.go:39–42,92` and `devcmd/vocabgen/main.go:169–170` change by hand).
- **Schema:** `recordsql/testdata/schema.sql` (`verify` columns `:789–797`; the `carrying_reading` view
  `:1419–1423`; `mint`/comments `:280–282`, `:374–379`) and `schema.epoch.json`.

### III.2 `lens verify` records only — `internal/cli/lens/verify.go` [MODIFY] (contract: CLI flag set)

| Site | Change |
|---|---|
| package comment `:87–93` ("# A reading that finds a defect mints its gap") | rewritten: a reading records what the source did for the sentence and holds nothing; a defect it shows is a gap `mint` opens, standing on the reading |
| `gapFlagsRequiredUnless` `:99`, `readingRequiredGapFlags` `:106`, `readingGapFlags` `:109`, `var gf` `:112`, `gf.register(c, readingSubject)` `:162`, the `RequiredUnless` loop `:163–165` | deleted: every gap flag is an unknown flag on `verify` |
| `seat.Supplies` for `gap_id`, `class_material`, `carried_by`, `carrier_minted` `:157–160` | deleted |
| `--title` usage `:146` ("names the source to whoever answers the gap your reading mints") | "kept on your reading beside its url; a gap you mint on this reading names the source through it" |
| `writeVerify` `:367–399` (`classes`, `backing`, the gap-flag refusals, `gf.fill`, the class check) | the reading is written alone; `markedSentenceOf` `:373` stays (the previous draft's F6 is unchanged in meaning: a backing found-source reading of a marked sentence is refused — `refuseMarked` `:591–616` loses "mints unsourced_claim" and says a found source backs only an unmarked sentence) |
| `recordReading` `:412–517` | the `gap != nil` branch deleted (Anchor, `PlacesToSentenceOf`, `uncitedSentence`, `verifiedFix`, `AppendReading`); `verifyResult.Gap` deleted; the result gains `event` (the reading's `events.id`) so the seat can cite it in `--reason`; `FailedChecks` keeps unread rows only (`:496–515`: the `CheckFoundAgainst` arm goes) |
| `givenGapFlags` `:519`, `mintsNothing` `:529`, `classWords` `:552`, `readingGap` + `words` `:644–672`, `failedCheckWords.words`' found-against arm `:694–709`, `sentenceOfRefusal` `:762–773` | deleted |
| `verifyResult.Human` | names the event: `recorded as event N`; a `refutes`/`absent` on a citation adds one sentence: "a gap of class claim_contradicts_source (or citation_figure_misattribution) about this citation is yours to mint, standing on this reading" — an act, no path |

`mint.go`'s `readingSubject` (`:259–266`), `gapSubject.span` (`:249`) and `record.ReadingClassField`
(`enums.go:344–360`) are deleted; `gapFlags` (`:235–238`) is `mint`'s alone and its comment says so.
`seat.go:1085` (`verify --reason`) loses "Where this reading mints a gap…".

### III.3 `mint` of 32/25 stands on the seat's own standing reading — `internal/record/` [NEW `standingreading.go`; MODIFY `actjudge.go`, `onegapperclass.go`, `mintbudget.go`] and `cli/lens/mint.go`

**The table** (one Go table, `readingMintedClasses`, one row per class; `TestEveryReadingMintedClassIsDriven`
fails a row no test refuses and admits):

| Class | Subject the mint names | The seat's standing reading that admits it |
|---|---|---|
| 32 `claim_contradicts_source` | `--about-kind anchor --about C` (a cited source) | of citation C, reads the work (`VerifyReads`), outcome `refutes` |
| 32 | `--quote S` (a found source) | `verify --url` by the seat whose `claim` locates once in the landing report in S's sentence, reads the work, outcome `refutes` |
| 25 `citation_figure_misattribution` | `--about-kind anchor --about C` | of citation C: reads the work with outcome `absent`, OR `source_completeness` `not_the_work` (F3) |
| 25 | `--quote S` | refused: a found source's `absent` is no defect of the report (#1647 F5); the refusal says the class names a citation |

**"Standing"** = the seat's LATEST reading of that subject since the subject's sentence last changed:
for a citation, the ledger's own rule (`anchorStanding.sinceEdit`, anchor not retired — read through
`anchorChecksOf`, which the check reuses rather than restating; the ledger's `stands` additionally
requires `reads`, so 25's not-the-work alternative reads the seat's latest check since the edit
directly, `reads` or not); for a found source, the seat's latest reading of that url whose stored claim
still locates once in the landing report, inside the mint's sentence (`readingSentenceIn`,
`reading.go:185`, stays for this). A reading of a sentence an edit has since changed does not stand: a
mint of 32 after blue's edit is refused until the seat reads again — the ruling's "standing" clause.

**Where it runs.** `requireOwnStandingReading(w writeView, seatID string, m *recordpb.Mint) error`, called
FIRST in `judgeAct`'s Mint arm (`actjudge.go:40`), before the one-gap rule and the budget: a mint refused
for its ground is not first told about a held subject or a full budget. A live act judgement, never the
shape check — the replay writes every archived 32/25 mint as it was made (§III.7 residue). The mint's
sentence for the `--quote` form is found as the one-gap rule finds it (`anchortext.Attach` of the quote
into the landing report, then `gapSentenceIn`, `onegapperclass.go:79–85`); that code is factored into one
helper, `mintSentenceAt(w, m)`, both rules call, so the two rules cannot disagree about which sentence a
mint is on. Any other class returns nil at the table lookup.

**The refusal** (composed from the row; never from a class-keyed string beyond the row's own):

> `--class claim_contradicts_source` states that the source says otherwise than the sentence. That is
> a reading's finding, and the record holds no standing reading of yours of citation C-… that reads the
> work and refutes it (your latest, event N, is `supports` / is of a sentence blue has since edited /
> none). Read it first — `verify --anchor C-… --as refutes`, with what the source says in --reason —
> and mint on that reading. Nothing is recorded.

For a `--quote` 32: "…no standing reading of yours of a source you found for this sentence that
refutes it; read the source with `verify --url … --quote <this sentence> --as refutes`, or name the
citation it contradicts (`--about-kind anchor --about C-…`) where the source is one blue cites". For a
`--quote` 25: "`--class citation_figure_misattribution` is a cited source's defect: name the citation
(`--about-kind anchor --about C-…`) your reading of it found absent or not the work".

**`--rests-on`-style evidence: none.** The subject and the seat are on the Mint already; the reading is
found by them. A seat types no id.

**The one-gap rule [MODIFY `onegapperclass.go`]:** delete the held arm (`:91–93`), `pairedSentenceOf`
(`:101–108`), `errReadingHeld`'s use and the "A READING'S MINT IS NOT REFUSED, IT IS HELD" paragraph
(`:40–44`). ADD the anchor arm (D5): where `about_kind` is `anchor`, an OPEN gap of the same class about
the same anchor refuses the mint with `oneGapRefusal` ("citation C-…" as the subject) — the inquiry
arm's SQL shape (`openGapsOfClassAt`, `:117–143`) with `about_kind = 'anchor'` — UNLESS the mint answers
the seat's own owed failed check (`answersOwnFailedCheckAt`, `mintbudget.go:375`), which is the case the
rule's comment `:28–30` exempts today and keeps exempting. The comment is rewritten: a gap about an
anchor IS a subject; what the rule still does not apply to is a section or a gap. Consequence, stated: a
second lens whose reading also refutes C mints no second 32 about C — its reading stands as information
(§III.4) and its point is a finding about the standing gap, the rule's existing route; there is no
deadlock because nothing is owed.

**The budget [MODIFY `mintbudget.go`] (D7):** delete the `w.act.pairs(m)` arm (`:252–259`) and the
`carried_by` query in `ownAnswersAt` (`:325–345`); `ownAnswersAt` keeps its two other sets
(`ownFailedCheckAnswers`, `ownsInquiryGap`). The header comment `:32–44` and `mintAnswerOutsideBudget`
(`:308`) say: a gap about an anchor whose check of yours read nothing or judged a computation unsound
takes no mint; so does the logic lens's gap about an inquiry. Whether the 32/25 gap a reading debt
obliges takes one is **F6** (§VI): under (b) `ownAnswersAt` gains the join on `standingReadingOf` and
the sentence reads "the gap your check or reading obliges"; under (a) every other gap takes one and the
debt item is non-blocking at the cap. Either way the `carrier_minted` exemption of 2026-10-30 (F6 of
#1647, narrowed 11-03) is gone as a mechanism: nothing is exempt for having been minted by a reading.

**Estoppel [`cli/lens/mint.go:155–158`, `reading.go:89–91`] — F4.** The guard is keyed on `--quote`
(`requireNotEstopped(l, quote, …)`, `mint.go:576`); a cited source's 32/25 names no quote, so estoppel
never asks of it; a found source's 32 is a quote mint and meets the guard like any gap. The "verify
ATTACHES NO SUCH GUARD" paragraphs are deleted. Whether a found-source 32 should be exempt is F4.

**`cli/lens/mint.go` [MODIFY]:** `--about-kind` usage (`:220–221`) names "a citation whose source your
reading finds against"; `--class` usage for 32/25 is the enum help's, rendered from the `means` (§III.1).
`mint.md` (§III.8) carries the rule in one paragraph.

### III.4 The reading debt — the lens owes a gap or a re-read (D4′) [NEW in `record/standingreading.go`; MODIFY `sitting.go`, `retirement.go`, `dispatch.go`, `lens/verify.go`]

**The fact.** `readingDebtsOf(evs, win, open, report) []readingDebt`: for each lens X and each standing
reading of X's that finds against (`VerifyReads` and `FindsAgainstOf(outcome)`):

| Reading | Standing when | The debt stands while | Discharged by |
|---|---|---|---|
| of citation C | the ledger's rule: X's latest reading of C that reads, from a sitting opened after the latest edit naming C, C not retired (`anchorStanding.stands`) | no OPEN gap of the class the outcome grounds (`refutes` → 32, `absent` → 25) is about anchor C — by ANY seat | any seat's mint of that class about C (the one-gap anchor arm, D5, then refuses a second, so a second lens's debt is discharged by the first's gap and no exemption is needed); X's later reading of C that reads and does not find against (the ledger replaces X's standing reading); blue's edit naming C (the reading is of a sentence the report no longer carries); blue's retire of C |
| of a found source U for sentence S | X's latest reading of U whose stored claim locates once in the report (`readingSentenceIn`) | no OPEN gap of class 32 whose quote is in that sentence (`gapSentenceIn`) — by any seat | any seat's 32 quoting the sentence (the one-gap sentence arm refuses a second); X's later non-finds-against reading of U for the sentence; an edit after which the claim no longer locates once (F1: the reading then stands for no sentence, and nothing is owed on it) |

`readingDebt{seat, subject (anchor or url), sentence, outcome, since (the reading's events.id), class}`.
One statement, in Go; it reads the ledger (`anchorChecksOf`) for the cited case and the report for the
found case, as `requireOwnStandingReading` does — the two share `standingReadingOf(seat, subject)`, so
the debt and the mint check cannot disagree about which reading stands.

**Where it lives — modelled on the owed unread/unsound gap, minus the PASS:**

| Surface | Today (unread/unsound) | The debt |
|---|---|---|
| the lens's work list | `sitting.go:430–434`: a `BlockerFailedCheck` owned by the lens is `add` (Blocks true) while `owed()` | a new arm beside it: each `readingDebt` of this seat is `add`; the item: "your reading of citation C-… at event N finds the source against its sentence (`refutes`), and no gap of class claim_contradicts_source stands on it — mint that gap about the citation, graded as you judge it, or read the citation again where the source backs the sentence after all. Your envelope is refused while this stands" (acts named; the flags are on `mint --help`) |
| the envelope gate | `envelope.go:40–64` `EnvelopeRefusal` → `Sitting.Owed()` (items with `Blocks` this seat clears) → `hookcmd.go:354` denies the `StructuredOutput` call | unchanged: the new item is blocking, so the gate refuses on it with no new code — this is why the debt is a work-list item and not a Blocker |
| the PASS gate | `passBlockersOf` (`passblockers.go:672`) builds every `BlockerKind`; all of them hold `pass_permitted` | NOT a `BlockerKind`; `readingDebtsOf` is never called from `passBlockersOf`, `eventBlockersOf` or `requireNoBlockers`. The gate refuses nothing on a debt |
| retirement / owing | `lensOwesOf` (`retirement.go:128–175`) names each owed blocker; a lens that owes does not retire (`retired()` `:72`) | a third arm: each debt of the lens, named `reading <since> of <subject>`; the lens does not retire while it owes one |
| dispatch readiness | Source 4 (`dispatch.go:528–616`): the owed failed check's lens is readied ONCE PER CAUSE — `firstAfter(ends.through[owner], b.since)`; a sitting stopped at the call limit or never closed is not sat through (`satThroughOf`, `:910–958`) and readies it again | the same arm over `readingDebtsOf`: readied until it sits through a sitting opened after the reading; the plan's reason: "X: its reading of C-… at event N finds the source against the sentence and no gap of class … stands on it — ready, to mint the gap or read again". `failedCheckReadiness`' found-against arm (`:1396–1405`) is deleted |
| `verify`'s own result | `verify.go:496–515` words the standing failed checks on the anchor | a `refutes`/`absent` reading of a citation says: "this reading is your debt until a gap of class … stands on the citation or you read it again — PASS is not held by it" |

**Does dispatch's PASS wait?** Only through `lens_ready` (`passblockers.go:689`): while the lens is
READY — once per cause, until it sits through a sitting opened after the reading — the PASS waits, as it
does for any ready lens. A lens that sat through since the reading and left the debt standing is not
readied again for it (the once-per-cause rule), is not retired (it owes), and is not ready at the head;
the PASS does not wait on the debt itself. That is the ruling's shape: the reading holds nothing; the
lens's graded gap, once minted, holds the PASS by its materiality.

**An abnormal end of the lens's sitting** (the per-sitting call limit, `shortCallLimit`; a stop not on
the record, `shortNeverClosed`): the debt is a fold over the record and stands; the sitting was not sat
through, so the lens is readied again at the next plan, the reason naming the short sitting
(`unsatSitting`). Its envelope was refused in the short sitting and is refused in the next while the
debt stands.

**The one-gap collision.** A second lens Y whose reading also refutes C owes a debt the moment it
reads; X's open 32 about C discharges it (the table's first row: "by ANY seat"). Y's own mint about C is
refused by the anchor arm (D5) and Y is not asked to make it: the discharge rule lives in
`readingDebtsOf`'s gap lookup (open gaps of the class about the anchor, `openGapsOfClassAt`'s anchor
form, the same query D5 uses), so the two rules read one set. Y's point, if any, is a finding about X's
gap.

**The budget: FORK F6.** Whether the gap the debt obliges counts against the lens's mint budget — see
§VI. At the cap, under "counts", a lens with a debt can neither mint (budget) nor honestly re-read, and
its envelope is refused until a gap of its closes: the deciding fact for F6.

**Census: where the board ALREADY shows a `refutes`/`absent` reading that no gap of its class answers
(information surfaces, unchanged by the debt).**

| Surface | File | Shows the reading? | Says no gap stands on it? |
|---|---|---|---|
| `show evidence` source rows | `record/evidenceview.go:225–266` (`EvidenceVerificationJSON`: outcome, confidence, text, seat, epoch per reading), `:407` fills `CarriedBy` (deleted by D1) | yes, every reading of every citation | today only by `carried_by` absent; after D1 by nothing on the row |
| `show evidence` `independent` rows | `evidenceview.go:300–303` (found-source readings) | yes | same |
| `counts.sources_refuted` | `evidenceview.go:321–325`, `:564` (`CitationsFoundAgainst`, the ledger's standing set) | the COUNT of citations a standing reading finds against NOW, gap or no gap | no — one number, no gap join |
| evidence.md "Citations" | `report/citations.go:21–60` (every reading per source, outcome and text; "Sources the report does not cite") | yes, the whole document | no |
| docket.md "read for the sentence" | `report/assemble.go:1098–1125` (`readingsCarried`, per GAP by `carried_by`) | only readings a gap carries — a reading no gap answers is NOT here; after D6 the line is `ReadingBehind(gap)` and still per gap | n/a |
| the chair's spot-check | `view/view.go:672–677` (`SpotCheckOf`), `seatprobe/boards.go:416–427` | samples the CLOSURE archive; reads no reading | no |
| the chair's work list / PASS gate | `record/sitting.go:365` (chair items from `passBlockersOf`), `passblockers.go:672–721` | today: the found-against `failed_check` blocker (deleted by D4); after: nothing | after D4: nothing |
| the capture's assembly screen | `capture/capture.go:470–495` (`FailingAnchors` → audit FAIL "evidence anchor(s) a check finds AGAINST … still in the assembled report") | post-run audit, keyed on the ledger's standing set; stays (D4 keeps the read) | no |

The chair is not told of a debt by a blocker (it is not one); `show evidence`'s rows and
`counts.sources_refuted` are its view of a refuted citation, and `telepathy`/the plan's `Why` line names
the lens readied for it.

### III.5 The found-against failed check is deleted as a PASS-holder [MODIFY `anchorchecks.go`, `failedcheck.go`, `passblockers.go`, `dispatch.go`, `evidencetables.go`, `retirement.go`]

- `CheckState` keeps `unread` and `unsound`; `CheckFoundAgainst` (`failedcheck.go:57`) is deleted, and
  with it: `failedCheck.outcome` (`:72–73`), `addFoundAgainst` and the `stateFoundAgainst` case
  (`:112–123`, `:161–166`), `carrierTree` (`:172–204`), `owed()`'s state test (`:89` → `answeredBy == ""`),
  the `CheckFoundAgainst` skips in `requireNotARepeatFailedCheck` (`:276`) and `ownFailedCheckAnswers`
  (`:436`), the found-against arm of `requireFailedCheckAnswered` (`:408–413`), the header paragraph
  `:39–47`; `answersOwnFailedCheck`'s comment `:241–243`.
- `anchorchecks.go`: `stateFoundAgainst` LEAVES the ledger (`:146–148`, `:159–160`, the `state()` loop
  `:187–195` becomes the computation-only `unsound` test; the header `:125–130`). A citation whose
  standing readings read the work is `stateHolds` whatever they found; one read by nobody since the
  edit is `stateUnchecked`; one read only by `unreachable`/not-the-work copies is `stateUnread` — the
  ordinary rules. The per-check `holds` (`anchorCheck.holds` `:56–59`, `:279` `reads && !FindsAgainst(b)`)
  STAYS: it is the demand's "check again that holds" (`latest_check`, `verifyHolds` `evidencetables.go:84`
  → `!FindsAgainstOf(outcome)`) and the ledger's `holds()` (`:112–120`, the gap view's backing arm). With
  `FindsAgainst` now `VerifyReads(v) && recordpb.FindsAgainstOf(v.GetOutcome())`. `VerifyReads`' comment
  (`:292–298`) loses "carried by the gap the reading minted". `CitationsFoundAgainst` (`:306–323`) and
  `FailingAnchors` (`:325–340`) stop reading a state and read the standing checks: an anchor with a
  standing check that does not hold — information for `counts.sources_refuted` (`evidenceview.go:564`)
  and the assembly screen (`capture.go:484`, §VII.1); `EvidenceTables.Failed` (`evidencetables.go:47–49,
  56, 60`) is deleted — its one reader was `FailingAnchors`' comparison.
- `passblockers.go`: the `CheckFoundAgainst` arms in `failedFact` (`:229–231`), `failedClears`
  (`:240–242`), `failedRemedy` (`:255–256`), `failedOwed` (`:261–272` → two arms), `blueFailedItem`
  (`:276–278`), `failedCheckRefusal` (`:302–308`, `:317–321`), `blockerKinds`' `failed_check` row
  (`:403–420`), `Blocker.owed` (`:120`); `dispatch.go:1392–1405` (`failedCheckReadiness`).
- `retirement.go:128–175` (`lensOwesOf`): the comment "A found-against check never holds its reader" goes
  with the state; the fold gains the debt arm (§III.4). `sitting.go:426–434`'s comment ("A FOUND-AGAINST
  CHECK IS NEVER ITS DEBT") goes with the new arm beside it.
- **`requireNotARepeatReading` [DELETE the reads arm] (D8):** `failedcheck.go:313–328` keeps only the
  unread delegation (`v.GetAnchor() != "" && !VerifyReads(v)` → `requireNotARepeatFailedCheck`);
  `repeatReadingRefusal` (`:330–381`) is deleted. Why delete rather than re-key: the rule existed so a
  second identical reading would not owe a second gap (`:256–266`); a reading now owes nothing. A
  same-sitting repeat is already refused — a citation reading by its keyed retry (`seat.KeysOn(c, "anchor")`,
  `verify.go:155`), a found-source reading by `requireNotARetriedReading` (`reading.go:217`, whose
  `carried_by` column read `:225` and "gap X carries it" words `:244–247` go). A cross-sitting identical
  reading replaces itself in the ledger and changes nothing. Bound: the sitting cap (`sittingcap.go:63`).
- **Lens credit:** no change to productivity. Retirement (`retirement.go:277–302`) counts fresh material
  mints; a lens that reads `refutes` and mints 32 is productive as today; a lens that reads and does not
  mint is barren for that sitting AND owes the debt, so it is not retired (`retired()` `:72`) and its
  envelope is refused — the debt, not a credit change, is what answers "a lens whose sitting produced
  only readings". The scorecard (`scorecard/scorecard.go`) counts no readings today and none after.

### III.6 The reading behind a gap — one Go helper, no SQL link (D6)

- **`record.ReadingBehind(evs, win, gap) (*Verify, bool)`** [NEW, in `standingreading.go`]: for a gap of a
  class in `readingMintedClasses` by seat X — about anchor C: X's standing reading of C (the ledger's
  replacement rule); with a quote: X's latest found-source reading whose stored claim locates once in the
  report at the gap's sentence. False for any other gap or where none stands. It is the ONE statement
  of "the reading behind this gap", and every Go reader below calls it. There is no SQL twin: the
  cited case needs none (the gap's `about_ref` and `minted_by` are columns; a SQL reader joins
  `latest_check` on them directly), and the found-source case needs the report.
- **`carrying_reading` view [DELETE]** (`views.go:412–420`), the index `verify_carried_by`
  (`schema.go:259–261`), and `demand_reach`'s `'reading'` arm (`views.go:493–495`); `demand_reach`'s
  header comment (`:435–459`) and the gap view's `evidence_met` comment (`:597–611`: arm 2 "or whose
  reading it carries") lose the reading clause. `Reach.About` (`demand.go:30–35`) is always true and is
  deleted with the two `'reading'` wordings in `DemandRefusal` (`:134–139`); `reachedGapsSQL`
  (`windowindex.go:104`) drops `"via" = 'reading'`; `windowindex.go:81, 95–99` comments.
  **What #1739's `carrying_reading` becomes: nothing.** Its four readers (issue comment 4 — `evidence_answered`
  arms 2/3 through `demand_reach`, the reaching retire `reachingRetiresOf`, `answers_retired_event`)
  read the `'about'` arm, which a cited source's 32/25 gap populates by its `about_ref`; the reaching
  retire of C meets the gap's demand through that arm as it does an unread gap's today; the reader's own
  backing re-read of C meets it through `latest_check` (arm 3, `lc.seat = dr.seat` where the seat is the
  MINTER — which is the reader, by the mint check). A found-source 32 (`--quote`) reaches no anchor, as
  a found-source reading's gap reaches none today (`cr.anchor` NULL), and its demand is met by blue's
  cite `--answers G` (arm 1) alone.
- **Other readers:** `report/assemble.go:1098–1125` (`readingsCarried`, docket.md's "read for the sentence"
  line, #1694 F12) → `ReadingBehind`, one line per gap that has one; `evidenceview.go:262–266, 407`
  (`CarriedBy`) deleted — the row's own fields (anchor or url+claim, seat, outcome) are the join key;
  `verbs.go:302` (`show evidence` help: "`carried_by` … is how blue finds the source a lens found for a
  gap it answers") → "docket.md's `read for the sentence` line under the gap names the source; under
  `independent` the row whose seat and claim are the gap's minter and sentence is that reading";
  `seatprobe/seatprobe.go:206–211` (`carrier_minted` count) → counts Mints of `readingMintedClasses`
  classes by the seat; `capture.go:483` (`FindsAgainst`) unchanged through the predicate.

### III.7 `migrate` — one translation (contract: record migration)

- **Archived shapes** (epochs 55–64 only; one archive, Eiffel): (1) an in-act reading — Mint m, Anchor a,
  Verify v with `carried_by` = m's gap, `carrier_minted` 1 (6 in Eiffel); (2) a held reading — v with
  `carried_by` naming another act's gap, `carrier_minted` 0 (1: event 798).
- **Translation:** delete `readingActs`, `readingActMembers`, `heldMember`, `writeReadingAct`, `failAct`
  and the `held`/`actMember` plumbing (`migrate/replay.go:127–130, 237–260, 547–561, 902–990`),
  `Replay.WriteBodies`/`readingActOf` (`record/replaywriter.go:130–207`), `writeView.act`
  (`writeview.go:30–34`), `appendBodies`' `act` and held arm (`record.go:674–720`), `appendAct`'s `act`.
  Every member is written singly in stream order under the replay's one writer, as every other archived
  event is: the Mint (a quote mint of class 32, shape-valid), the Anchor, the Verify with no link fields —
  the two fields are absent from the proto, so the source's `carried_by`/`carrier_minted` columns are
  dropped at `OldEvent` → body translation (one line in the Verify arm, beside `eraentries.go:344–348`'s
  renames); `remap.idFields` loses `Verify.carried_by` (`remap.go:116`). `raisedReading` (`:859–873`)
  stays (a raise names a reading by event id, a different relation).
- **Residue, stated:** every archived 25/32 mint comes forward UNJUDGED by `requireOwnStandingReading`
  — 4 archived 25 mints in 4 runs, Eiffel's 6 class-32 mints (5 in-act, 1 plain G-b2ba7d66) — because
  the check is a live act judgement (§III.3) and the replay asks none. The archived found-against checks
  (#1647 §III.9's "carrier-less checks") hold nothing after migration: there is no such state. The
  held reading 798 comes forward as a reading of G-1d97d860's sentence that says `refutes` and names
  nothing; `ReadingBehind(G-1d97d860)` finds it by (seat, anchor) if its seat is the gap's minter, else
  the gap has no reading behind it on the docket. No fill (the ruled pattern).
- **Part 0 — the archive [NEW `run-archive/2026-11-04_eiffel-tower-height.tar.gz`]:** the one record that
  holds the shape; layout matches `2026-10-30_mersenne-31-prime-sum.tar.gz` (inputs, records,
  computations — records and computation bundles, no fetched-source cache). Its own PR, asked before it
  lands (semantic consent: an outward change to the archive set). It re-renders every archive-iterating
  golden (§III.9 G-arch, 39 tests) and MUST pass `TestEveryArchivedRunMigrates` at epoch 64 first — on
  main today the in-act translation reads it; after Part 1 the deletion does.

### III.8 Agent-facing text [MODIFY / REGENERATE] (contract: prompts, constitutions, help)

Every sentence below is rewritten to the new model, present tense; generated copies are regenerated.
The census (quotes trimmed) is in §III.9 A1–A13.

| Source | Says today | Becomes |
|---|---|---|
| `help/verify.md:15,19,21,25,27,29` | in-act mint; one-gap holds the reading; found against holds the PASS; estoppel exemption; `--url` mints by outcome; result names `gap` | a reading records; a `refutes`/`absent` on a citation is the ground of a gap you mint about the citation (an act; the flags are `mint`'s page's); a refuting found source is the ground of a gap quoting the sentence; a reading holds no PASS — it is YOUR DEBT until a gap of its class stands on the subject or you read again, and your envelope is refused while it stands; the result names the event |
| the lens constitutions' duty fragments (A7–A9) and `terms.json` `reading` | — | one sentence each on the debt: a reading that finds against obliges its lens's graded gap or its re-read; the PASS is held by the gap's materiality, never by the reading |
| `help/mint.md:9,15,17` | one-gap on sentence/inquiry; budget exemption; "a reading … mints its gap in the same act" | one-gap on a sentence, an inquiry OR a citation; the standing-reading rule for 32/25 (one paragraph: the subject form per source kind, what standing means, the refusal); the budget sentence without the reading exemption |
| `help/close.md:13` | "your gap another lens's reading named … found-against check clears…" | unread and unsound only |
| `help/edit.md:13` | "a lens's reading … still mints its gap" | deleted clause |
| `help/reproduce.md:23,27`, `dispatch.md:25` | unsound; "on a failed check" | unchanged (unread/unsound stay) — read before/after all the same |
| `seat/verbs.go:302,306,334`; `seat/seat.go:1085`; `cli/setup.go:89` | `carried_by`; `sources_refuted`; budget exemption; reason-is-the-gap's-argument | §III.6; `sources_refuted` unchanged; exemption sentence without the reading clause; reason is the reading's |
| `record/passblockers.go` found-against arms, `dispatch.go:1392–1405`, `reading.go:120,128,132` refusals, `failedcheck.go` refusals, `mintbudget.go:308` | | deleted / unread-unsound only (§III.5, §III.3) |
| `record.proto` `SourceOutcome` `means` `:531–542`, `GapClass` 25/32, `AboutKind.ANCHOR` | "mints … in the same act" | §III.1 |
| `scripts/agentgen/src/fragments/lens-audit-and-mint.md:3,6,7,9,10`; `fragments/lens-duties.md:5,6`; `fragments/blue-search-and-cite.md:30–31,39,43`; `agents/red-lens-evidence.md:17,19`; `agents/lead-judge.md:81`; `skills/research-protocol/SKILL.md:14–21` | "records its gap in the same act"; "whose reading it carries"; "found against" as a failed check | the reading records; the lens mints the gap about the citation on its reading, or quoting the sentence for a found source; a found-against check is no longer a thing; the demand is met through the citation the gap is about |
| `skills/research-protocol/scripts/debate.js:883` (authored `ledgerClause`) | "your reading records the gap, and blue answers it with a citation" | "where it refutes the sentence, you mint the gap on that reading; where it backs an uncited sentence, you mint the gap that asks for the citation" (acts, no verb path; `TestTheSeatPromptsNameNoVerb`); `:1090–1097` unchanged |
| `tools/internal/terms/terms.json` rows `reading` (`:762–763`), `failed check` (`:1091–1092`) | in-act mint; third state | reading: "…it holds nothing by itself; a gap a lens mints on it is the finding"; failed check: two states; then `go -C scripts run ./vocabdoc` (`docs/vocabulary.md:464,632`) |
| `docs/citations-and-bibliography.md:162–164,174–179,181–188,195–197,209–217`; `docs/seat-command-triggers.md:55,67` | three states; in-act; `carried_by` | two states; the standing-reading mint; the docket line |
| `report/layout.go:167,184` | "as a lens's `verify` does where its reading mints a gap" | deleted clause |
| `plans/compute-and-show.md` (40 lines, §III.9 row P) | `Verify.carried_by`, carrying reading, `found_against` | one dated note at `:515–523` and `:1572` that #1727 removed the link and the state; the rest is history and stays |
| GENERATED: 12 `agents/*.md` (3 `carried_by` hits each from the inlined manual), `skills/adversarial-audit/SKILL.md`, `skills/research-protocol/SKILL.md`, `help_gen.go`, `annotations_gen.go`, `vocabulary.md`, `debate.js` marked regions (`:105–123`, `:373–451` — no reading word; expected unchanged) | | `agentgen`, `cmdtreegen`, `vocabgen`, `vocabdoc`, `enginegen` |

AFTER regenerating, `verify --help`, `mint --help`, `close --help`, `edit --help` and `show evidence --help`
are each read before and after, one command at a time, for semantic loss.

### III.9 Carrier/consumer census (re-run 2026-11-06 on this tree by three grep sweeps; totals at the end)

Commands (worktree root; `T=plugins/frank-exchange-of-views/tools`):
`grep -rnE 'carried_by|CarriedBy|carrier_minted|CarrierMinted|carrying_reading' -I $T --include=*.go --include=*.sql --include=*.golden --include=*.md | grep -v record.pb.go`;
`grep -rnE 'GapClassesOf|FoundSourceGapClassesOf|ReadingGapClasses|FindsAgainst|ReadingClassField|gap_classes|found_source_gap_classes' --include=*.go --include=*.proto $T`;
`grep -rnE 'AppendReading|errReadingHeld|readingAct|pairedSentenceOf|stateFoundAgainst|CheckFoundAgainst|found_against|carrierTree|repeatReadingRefusal|readingSubject|readingGap' --include=*.go $T`;
the T-sweep `grep -rlE 'AppendReading|CarriedBy|CarrierMinted|carried_by|carrier_minted|ReadingGapClasses|GapClassesOf|ReadingClassField|citeReading\(|mustRead\(|readingGap|\.Gap\.ID|stateFoundAgainst|FoundAgainst|found_against|carrying_reading' --include=*_test.go $T`;
the phrase sweep `grep -rnE 'same act|carried_by|finds against|found against|found-against' plugins/frank-exchange-of-views/{tools/internal/cli/seat/help,docs,skills/research-protocol/scripts/debate.js} scripts/agentgen/src $T/internal/terms/terms.json`.

**Code (non-test), 30 sites (67 `carried_by` lines + 55 option lines + ~130 act lines, de-duplicated by file):**

| ID | Site | §III |
|---|---|---|
| C1 | `recordpb/record.proto:152–165` (50014/15), `:529–545` (outcomes), `:460`, `:483` (25, 32), `:1179` (AboutKind), `:1431–1441` (Verify) | 1 |
| C2 | `recordpb/gapclass.go:63–77`; `recordpb/annotations.go:39–42,92`; `devcmd/vocabgen/main.go:169–170`; `annotations_gen.go` (5, generated); `record.pb.go` (31, generated); `recordsql/decode_gen.go:1475` (generated) | 1 |
| C3 | `record/enums.go:344–360` (`ReadingClassField`) | 2 |
| C4 | `record/reading.go:19–183, 225, 244–247, 259–269` (keep `readingSentenceIn` `:185`, `readingForm`, `requireNotARetriedReading` minus `:225,244–247`) | 2, 3, 5 |
| C5 | `record/record.go:674–720` (`appendBodies` act/held), `:747` (`appendAct`), `:1735–1741` (`validate` Verify arm: `requireReadingCarrier` goes), `:1931` (31's check STAYS) | 3, 7 |
| C6 | `record/actjudge.go:20–23` (list), `:40–64` (Mint arm order), `:131–135` (`requireReadingNamesCarrier`) | 3 |
| C7 | `record/writeview.go:30–34` | 7 |
| C8 | `record/onegapperclass.go:16–50, 75–77, 89–96, 101–108` | 3 |
| C9 | `record/mintbudget.go:32–44, 252–259, 308, 320–345, 375` | 3 |
| C10 | `record/failedcheck.go:12–47, 57, 72–89, 110–123, 161–166, 172–204, 241–243, 256–381, 383–427, 429–441` | 5 |
| C11 | `record/anchorchecks.go:125–130, 146–163, 187–195, 279, 292–298, 306–340` | 5 |
| C12 | `record/passblockers.go:86–92, 120, 223–325, 403–420`; `record/dispatch.go:1387–1405` | 5 |
| C13 | `record/retirement.go:128–148` (comment) | 5 |
| C14 | `record/evidenceview.go:262–266, 269–283, 407, 564` | 6 |
| C15 | `record/demand.go:16–35, 83, 123–139`; `record/windowindex.go:81, 95–104` | 6 |
| C16 | `recordsql/views.go:412–420, 435–459, 493–495, 597–621`; `recordsql/schema.go:259–261`; `recordsql/evidencetables.go:47–49, 56, 60, 75–84` | 1, 5, 6 |
| C17 | `record/replaywriter.go:130–207` | 7 |
| C18 | `record/migrate/replay.go:127–130, 237–260, 547–561, 859–873 (stays), 902–990`; `migrate/remap.go:116`; `migrate/eraentries.go:331–348` (Verify arm, the drop) | 7 |
| C19 | `cli/lens/verify.go` (§III.2 table) | 2 |
| C20 | `cli/lens/mint.go:155–158, 217–222, 235–266, 576–584` | 3 |
| C21 | `cli/seat/verbs.go:302, 306, 334`; `cli/seat/seat.go:1085`; `cli/setup.go:89` | 8 |
| C22 | `report/assemble.go:1098–1125`; `report/layout.go:167, 184`; `report/citations.go` (unchanged; §III.4 census) | 6, 8 |
| C23 | `seatprobe/seatprobe.go:206–211` | 6 |
| C24 | `capture/capture.go:470–495, 511–524` (`FailingAnchors`, `FindsAgainst`: through the predicates, text unchanged) | 5 |
| C25 | `recordsql/testdata/schema.sql`, `schema.epoch.json`; `requirements.json:3` | 1 |
| C26 | NEW `record/standingreading.go` (`readingMintedClasses`, `standingReadingOf`, `requireOwnStandingReading`, `mintSentenceAt`, `ReadingBehind`, `readingDebtsOf`); `recordpb` accessor `FindsAgainstOf`; the debt arms in `sitting.go:414–472`, `retirement.go:128–175`, `dispatch.go:528–616`, `envelope.go` (no change: reads `Blocks`) | 3, 4, 6 |
| C27 | `releasegate/fuzz` — §III.10 | 10 |

**Tests (T-sweep), 45 files / 364 lines, each rewritten or deleted row by row in Part 1:**
cli (16): `closuredemand_test.go` (10), `democarriers_test.go` (11), `enum_help_test.go` (1),
`keyedretry_test.go` (4), `m23corroboration_test.go` (3), `mintreason_test.go` (11),
`readingact_test.go` (59), `readingattribution_test.go` (4), `readingcap_test.go` (3),
`readingcopy_test.go` (16), `readinghelp_test.go` (1), `readinghelpers_test.go` (2),
`readingmigrate_test.go` (6 — both tests write an in-act record at this epoch, impossible after Part 1:
deleted, T13/T14 replace them), `readingplace_test.go` (18), `readingsource_test.go` (9),
`selector_test.go` (1); record (17): `backingarm_test.go`, `blockerowner_test.go:189`,
`envelope_test.go:338`, `export_test.go`, `foundagainst_test.go` (12), `foundagainstm30_test.go`,
`foundagainstm32_test.go`, `foundsource_test.go` (12), `gapacts_test.go`, `migrate/readingsforward_test.go:164`,
`onegapplacement_test.go`, `passblockers_test.go`, `reachingretire_test.go`,
`recordsql/evidencedemand_test.go`, `reopened_test.go` (7), `replaywriter_test.go` (3),
`unreadcitation_test.go`; capture (1): `capture_test.go` (4); fuzz (11): §III.10. **Plus 10 cli files
outside the sweep that run `verify` with gap flags through `gapFlagArgs` (`verifyclaim_test.go:211`):**
`verifyclaim_test.go:62,64`, `mintbare_test.go:157`, `verifyconfidence_test.go:28`, `bridgekind_test.go:63`,
`deadpath_test.go:625`, `gapanchor_test.go:177`, `requiredmarker_test.go:120`, `verbs_test.go:117`,
`answers_test.go:297`, `verifycopy_test.go:203`. The 31 test `record/inquiry_test.go:427` is untouched.

**Goldens (G), 8 files + the archive set:**

| ID | File | Why |
|---|---|---|
| G1 | `difftest/testdata/manual_runs_every_help_page_live.golden` (28 lines) | verify, mint, close, edit, show evidence pages |
| G2 | `difftest/testdata/gap_class_enum.golden:29,36` | 25/32 `means` |
| G3 | `difftest/testdata/error_catalogue.golden` (8) | reading refusals go; the standing-reading refusals arrive |
| G4 | `record/testdata/plans.golden` (42) | verify INSERTs, `carrying_reading`, budget SELECT |
| G5 | `record/testdata/shapes.golden:458` | `carrying_reading` |
| G6 | `recordsql/testdata/schema.sql` (18) + `schema.epoch.json` | columns, view, index, epoch |
| G7 | `migrate/testdata/prose-rewrites.golden` (6) | rewritten refusals |
| G8 | `tests/simulator/testdata/prompt-blue-synthesize.golden:7`; `prompt-red-lens-evidence*.golden` | `debate.js:883` |
| G-arch | the archive-iterating set: 39 tests across `migrate/*_test.go` (`everyarchive`, `everyrender`, `proserewrites`, `readingsforward:405`, `archiveonce`, `argumentresidue`, `remaplists` ×2, `gapclass_internal`, `archivedgaps`, `evidenceanchor`, `gapplace`, `benchfiling`, `benchoccasion` ×2, `latestsitting`, `regraderesult`, `storedwindows`, `eventcache`, `reportcheckpoint`, `unmigratedarchive`, `computationarchive`, `computationresidue` ×9, `correctionarchive`, `dispatchpin`, `cutfields`, `remanddirection`, `reproduceanchor`, `findingarea`, `inquirymethod`, `proserefs`), `record/backingarm_test.go:254`, `devcmd/archiverun/outcomes_test.go:103` | Part 0 adds an archive |

**Agent-facing (A), 13 sources:** A1 `help/verify.md` (7 lines); A2 `help/mint.md` (3); A3 `help/close.md:13`;
A4 `help/edit.md:13`; A5 `record.proto` outcome/class/about-kind `means` (9); A6 `seat/verbs.go:302,306,334`,
`seat/seat.go:1085`, `cli/setup.go:89`; A7 `agentgen/src/fragments/lens-audit-and-mint.md:3,6,7,9,10`,
`lens-duties.md:5,6`, `blue-search-and-cite.md:30–31,39,43`; A8 `agentgen/src/agents/red-lens-evidence.md:17,19`,
`lead-judge.md:81`; A9 `agentgen/src/skills/research-protocol/SKILL.md:14–21`; A10 `debate.js:883` (authored;
generated regions hold no reading word — grep `carried_by|same act|claim_contradicts_source` on the file:
0 hits in `:105–123`, `:373–451`); A11 `terms.json:762–763, 1091–1092`; A12
`docs/citations-and-bibliography.md` (6 passages), `docs/seat-command-triggers.md:55,67`; A13 the generated
set (12 agents, 2 preloaded skills, `help_gen.go`, `annotations_gen.go`, `vocabulary.md:464,632`).
Refusal strings (Go, agent-read): `reading.go:120,128,132`, `failedcheck.go:289–299,336–344,410–423`,
`passblockers.go:223–325`, `dispatch.go:1392–1405`, `verify.go:394,529–546,694–709`.

**Views (V), 3:** V1 `carrying_reading` (deleted); V2 `demand_reach` (reading arm deleted); V3 the gap
view's `evidence_met` comment. **1a′ readers (L), 4:** all through V2's 'about' arm after the change (§III.6).
**Plans (P):** `plans/compute-and-show.md` 40 lines (§III.8).

**Census totals:** 30 code sites (27 modified, 1 new, 2 generated families); 45 T-sweep test files + 10
gap-flag files; 11 fuzz files with 18 tally fields and 2 lifecycle tallies (§III.10); 8 golden files + 39
archive-iterating tests; 13 agent-facing sources + the generated set; 3 views; 4 1a′ readers; 1 plan.

### III.10 Release fuzz — `releasegate/fuzz/` [MODIFY] (CI only, never run locally)

| Site | Today | Becomes |
|---|---|---|
| `reading_test.go` (51 lines: `readingClasses:32`, `readingGapArgs:61`, `checkReading:144–173`, `readCitation:275`, `foundAgainstCleared:399`, `checkCarriedBy:458`, `notTheWorkReading:474`, `readFoundSource:577`, `readAnotherSource:715`, `fixFoundAgainst:774`, `readingTally:831`, `readRepeatedSentence:1101`, `readOverPrescription:1270`, `readNotTheWorkFound:1335`) | in-act mint, held readings, `carried_by` checks, found-against remedy | `verify` alone; then, on a coin, `mint --class … --about-kind anchor --about C` (cited) or `--quote S` (found) and the stored gap's `ReadingBehind` is the reading; the second lens's mint about C refused by the anchor arm |
| `foundagainstseed_test.go` (47: `forcedFoundAgainstChain:111`, `foundAgainstReading:284`, `foundAgainstReread:297`, 6 tests `:411–811`) | forced found-against chain through the failed check | forced chain: reading → mint on it → blue's edit/retire → the gap's demand met through the 'about' arm; the "check holds the PASS" assertions go |
| `failedcheckstate_test.go:18–100, 143–178, 190–210` | `CheckFoundAgainst` drives | two states only; the found-against test deleted |
| `anothersource_test.go:18–82` | second source held by the first's gap | second found-source mint refused by the one-gap rule on the sentence |
| `lifecycleseed_test.go:94–95, 153–154, 314–334` (`ReadingDemandRetired`/`Reread`, `lifecycleRead`); `sourcedemandseed_test.go:56–117` | demand reached through a carried reading | demand reached through the gap's `about_ref`; the retire and the minter's backing re-read each met |
| `fuzz_test.go:1143–1159` (`sourceQuoteRefused`, `FoundSourceGapClassesOf`), `:1351–1374`, `:3333–3383`, `:3707–3775`, `:4781–4799` (`CheckFoundAgainst` arms), `:5861, 5972, 5977`, `:6776`, tally copy-out `:7600–7632`, outcome fields `:6199–6224` | | rewritten per row; the `CheckFoundAgainst` arms deleted |
| `sweeptally_test.go:201–220` fields `FoundAgainstClosesRefused`, `FoundAgainstRepeats`, `FoundAgainstEdits`, `FoundAgainstRetires`, `FoundAgainstClearedByOwnReading`, `FoundAgainstStage`, `ReadingDemandRefused`, `NotTheWorkFlagsRefused`, `FoundSourceBudgetRefused`, `ReadingsMinted`, `ReadingsHeld`, `FoundSourceGaps` and their gates `:1159–1175` | 18 reading-shaped fields | deleted: the eight found-against/held/minted/flag fields. Replaced (new meaning, new name): `reading_mints_refused` (map by class: the mint refused for no standing reading), `reading_mints_landed` (by class and subject form: anchor, quote), `second_reading_mints_refused` (the anchor arm), `found_source_mints_budget_refused`, `reading_demand_met_by_retire`, `reading_demand_met_by_reread` — each with a zero-is-false-green gate |
| new drive | — | per sweep, for each row of `readingMintedClasses`: one mint WITHOUT a standing reading (expects the refusal naming `verify`); one with the wrong outcome; one after blue's edit of the sentence (not standing); one with a standing reading (lands) |
| new drive (the debt) | — | a lens reads `refutes` and returns its envelope → refused, tallied `envelopes_refused_on_reading_debt`; discharged on a coin by its mint, its re-read, another lens's gap of the class, blue's edit or retire → `reading_debts_discharged` (map by discharge); a forced short sitting → `reading_debt_readied_again`; each gated non-zero |

### III.11 Required tests (each names how it fails if the change is wrong; CLI tests drive the shipped path — `run(t, "lens", "verify", "--run", …)` / `"mint"` — as a seat does)

- **T1** `verify --anchor C --as refutes --class claim_contradicts_source …` → unknown flag, exit non-zero,
  event count unchanged; one row per deleted flag. *Fails if* a gap flag survives on `verify`.
- **T2** a `refutes` reading of C writes one Verify and nothing else (count +1; no Mint, no Anchor); the
  JSON carries `event`; `PassBlockers` holds NO `failed_check` on C. *Fails if* the reading still mints,
  or a blocker is built from it.
- **T3** `mint --class claim_contradicts_source --about-kind anchor --about C` by X with no reading of C →
  refused naming `verify` and C; nothing written. Then X reads C `supports` → still refused (outcome named);
  X reads C `refutes` → lands, `about_kind anchor`, `about_ref C`, `ReadingBehind` = that reading.
  *Fails if* G-b2ba7d66's shape lands, or the outcome is not checked.
- **T4** standing: X reads C `refutes`; blue edits C's sentence; X's mint is refused ("of a sentence blue
  has since edited"); X reads again `refutes` from a sitting opened after the edit → lands. Also: X's
  `refutes` then X's later `supports` → refused (latest replaces). *Fails if* "standing" is "ever read".
- **T5** another seat Y's `refutes` of C does not admit X's mint. *Fails if* the key drops the seat.
- **T6** 25: X reads C `absent` → `mint --class citation_figure_misattribution --about-kind anchor --about C`
  lands; X reads C on a not-the-work copy → lands (F3 as recommended) AND the unread check's `AnsweredBy`
  is the gap and the mint took no budget; X's `refutes` → refused naming `absent`. *Fails if* either
  alternative of the OR is missing, or the not-the-work gap does not answer the unread check.
- **T7** found source: `verify --url U --quote S --as refutes` by X, then `mint --class
  claim_contradicts_source --quote S` lands with X's reading behind it; `--quote S'` (another sentence)
  refused; the repeated-sentence fixture (`readingplace_test`'s) places by the quote as any mint does.
  `mint --class citation_figure_misattribution --quote S` refused naming the citation form. *Fails if* a
  found-source 32 is refused, or a found-source 25 admitted.
- **T8** one-gap anchor arm: X's 32 about C open; Y (with its own `refutes` of C) mints 32 about C →
  `oneGapRefusal` naming X's gap; Y's gap of another class about C lands; Y's owed unread gap about C
  (Y's `unreachable` first since an edit) lands over X's open gap of the same class. *Fails if* duplicates
  land, or the obliged gap is turned away.
- **T9** with X's `refutes` of C standing and no gap: `gate --as PASS` by the chair is NOT refused on C
  (no other blocker); `show evidence` `counts.sources_refuted` = 1. *Fails if* a found-against blocker
  survives, or the ledger read is lost with it.
- **T10** `close --disposition repaired` of X's 32 about C while X's `refutes` stands: refused by the
  DEMAND (`requireDemandMet`: "a reading of C by X … that backs the sentence"), not by a failed check;
  after X's backing re-read, or blue's `retire --anchor C`, admitted. *Fails if* `requireFailedCheckAnswered`
  still names found against, or the demand no longer reaches C through the 'about' arm.
- **T11** budget: at X's cap, a 32 about C is refused by the budget; a 25 about C where X's own unread
  check is owed is admitted. *Fails if* the exemption survives for 32 or is lost for the owed gap.
- **T12** docket.md golden: the gap's "read for the sentence" line names the source via `ReadingBehind`
  (cited: C's url; found: U); `show evidence` row has no `carried_by`. `seatprobe` counts the gap.
- **T13** `TestTheEiffelReadingActsComeForwardAsThreeEvents` on `run-archive/2026-11-04_eiffel-tower-height.tar.gz`
  at epoch 65: zero refusals; for each of events 167, 182, 542, 791, 794, 797 the migrated record holds
  the Mint, the Anchor and the Verify in source order, the Verify's columns holding no `carried_by`;
  798 is a Verify alone; G-b2ba7d66 and the 5 in-act 32 gaps are on the gap view with their closures;
  `PassBlockers` holds no `failed_check` on any citation a reading refuted. *Fails if* a member is
  dropped, a refusal appears (the replay judging), or a blocker is rebuilt.
- **T14** a synthetic `[]OldEvent` stream (built in-test as `gapclass_test.go` and `cutfields_test.go`
  build theirs) at epoch 64 holding one in-act act, one held reading and one plain 32 mint → migrates
  with zero refusals to the same shapes. *Fails if* the translation depends on Eiffel's particulars.
- **T15** `TestEveryArchivedRunBehindThisBinaryIsRefusedUnmigrated` re-armed by the bump; `schema.sql`
  golden shows no `carried_by`, no `carrying_reading`, no `verify_carried_by`.
- **T16** `TestEveryReadingMintedClassIsDriven`: for each row of `readingMintedClasses` a refused and an
  admitted mint exist in the suite (table-driven over the rows); deleting a row turns the class's T3/T6
  red. *Fails if* a class is admitted with no check.
- **T17** G1–G3, G8: the help, enum help and catalogue goldens; `TestTheSeatPromptsNameNoVerb` and
  `TestNoPromptGrowsItsCommandCatalogue` green on the new `debate.js:883` and constitutions.
- **T18** the debt: X reads C `refutes`; X's work list holds a blocking item naming C and event N;
  `EnvelopeRefusal(run, X)` is non-empty and names it; `PassBlockers` holds nothing on C; the plan readies
  X ("to mint the gap or read again"). *Fails if* the debt is a PASS blocker, or is not blocking for X.
- **T19** discharge, one row each: X's 32 about C → item gone, envelope admitted; Y's 32 about C (X's
  not minted) → X's item gone; X's later `supports` of C → gone; blue's edit naming C → gone (and C
  unchecked); blue's `retire --anchor C` → gone; found source: X's 32 `--quote S` → gone; an edit after
  which X's claim no longer locates → gone (F1). *Fails if* any discharge is missing (one row red) or
  the debt outlives the ledger's standing.
- **T20** once per cause: X's sitting opened after the reading ends sat through with the debt standing →
  the next plan does not ready X for it, `lensOwesOf` still names it, X is not retired at 3 barren
  sittings; X's sitting stopped at the call limit → the next plan readies X again. *Fails if* the debt
  readies forever (PASS held by `lens_ready` indefinitely) or an unsat sitting ends it.
- **T21** `TestEveryWorkListItemKindAgreesOnEverySurface`'s sibling for the debt: the item, the plan's
  reason, `lensOwesOf`'s name and `verify`'s result sentence are produced from one `readingDebt` (a
  table test over the fold's rows). *Fails if* a surface words the debt from its own fold.

## IV. Risk & Mitigation

Grades: likelihood × impact × complexity, each L/M/H.

| # | Risk | L×I×C | Mitigation |
|---|---|---|---|
| K1 | A lens reads `refutes` and never mints: the defect reaches no board and holds no PASS (the m32 shape) | M×H×M | The debt (§III.4): the lens's envelope is refused and it is readied once per cause; T18–T20. Residual: a lens that sat through once and left the debt is not readied again, so a refuted citation can reach a PASS with no gap — the ruling's chosen cost ("the lens's graded gap holds, by its materiality"), stated in `verify --help` |
| K1′ | Deadlock at the cap: a lens with a debt cannot mint (budget) and cannot honestly re-read; its envelope is refused until a gap of its closes | M×H×L | F6 (§VI); under (b) the obliged gap takes no mint; under (a) the item must become non-blocking at the cap — one of the two is chosen before Part 1 |
| K2 | A lens mints 32 by `--quote S` for a CITED source (the help not read) and is refused as a found-source mint | M×L×L | The `--quote` refusal names the citation form with the citation id found in S (§III.3); fuzz drives both forms (§III.10) |
| K3 | The anchor arm of the one-gap rule turns away an obliged unread/unsound gap | L×H×M | the exemption keyed on `answersOwnFailedCheckAt` (D5); T8's third row |
| K4 | The demand of a found-source 32 is met only by blue's cite: a gap blue cannot answer by retire | L×M×L | Unchanged from today (a found-source reading's gap reaches no anchor); stated in `mint.md`; the bench's docket route stands |
| K5 | `FailingAnchors`/`sources_refuted` lost with the state deletion, silencing the assembly screen (a plausible zero) | M×H×L | D4 keeps the ledger state as a read and rewrites `FailingAnchors` off `standing()`/`holds()`; `capture_test.go`'s assembly-screen tests re-armed; §V check 9 deletes the arm and expects red |
| K6 | Eiffel's epoch label (56, #1694 branch numbering) vs main's 56 | L×M×L | measured: main's migrate at 61 admitted it with 0 refusals; T13 re-proves at 65 |
| K7 | A PR bumping to 65 lands first | M×L×L | take the next number; `schemagen` + `TestEveryArchivedRunBehindThisBinaryIsRefusedUnmigrated` |
| K8 | A half-state: `verify` still minting while `mint` already checks (or the reverse) across the phases | M×M×M | one PR (§V phases 1a–1e), each phase green; the PR opens after 1e |
| K9 | The fuzz (CI only) is the first to drive the two subject forms at scale; a seat-level defect shows after merge | M×M×M | a `--smoke` universe run reads the record for no Verify with a gap field, each 32/25 mint with a reading behind it (§V check 11); release-fuzz cadence: a full pass before the next logic PR |
| K10 | 46 archived 25/32 mints and G-b2ba7d66 come forward unjudged and read as "the rule has exceptions" | L×L×L | stated residue (§III.7); the replay judges nothing, by design |

## V. Phases and Verification Plan

**Parts.**

| Part | Content | Ships alone? | Epoch |
|---|---|---|---|
| **0** | `run-archive/2026-11-04_eiffel-tower-height.tar.gz`; regenerate G-arch; nothing else | yes — data only; asked first | none |
| **1** | D1–D9: the record change, `verify` records only, the standing-reading mint check, the one-gap anchor arm, the found-against deletion, the budget change, the helper and view deletions, migrate, every agent-facing text, the fuzz, T1–T17 | **no** — the in-act mint and the mint check are two ways of minting one class; help, constitutions and the terms registry can state only one, and the link's writer (`verify`) and its readers (views, budget, failed check) must change at one epoch | 65 |

Within Part 1, one agent per phase, each leaving the tree green: (1a) proto, generated files, epoch,
`finds_against` predicate, `readingMintedClasses`, `requireOwnStandingReading` + `mintSentenceAt` +
`ReadingBehind`, the anchor arm — `verify` still minting; (1b) `verify` records only, deletions in
`reading.go`/`record.go`/`onegapperclass.go`/`mintbudget.go`/`failedcheck.go`/`passblockers.go`/`dispatch.go`/
views/`demand.go`/`windowindex.go`, readers switched; (1c) migrate + T13/T14; (1d) agent-facing text +
goldens; (1e) fuzz. The PR opens after 1e.

**Validation loop.** Run in the worktree after `unset $(env | grep -o '^FEOV_[A-Z_]*')`; absolute paths;
each check names what re-arms it; read the FAIL COUNT, never an exit code.

1. `go -C scripts run ./protogen && git diff --exit-code -- '*record.pb.go' '*record.proto.sha256' '*decode_gen.go'` → no diff once regenerated output is committed. *Re-armed by:* any `record.proto` edit.
2. `go -C scripts run ./schemagen && grep -n eventSchema plugins/frank-exchange-of-views/requirements.json` → 65 both sides. *Re-armed by:* `requirements.json`, any event-shape change.
3. `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/cmdtreegen && go -C scripts run ./agentgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen && go -C scripts run ./vocabdoc && go -C scripts run ./agentgen -check && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check && git status --short` → only intended files change; both `-check` clean. *Re-armed by:* help .md, `terms.json`, agentgen sources, proto options.
4. `UPDATE_GOLDENS=1 go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/difftest/ ./internal/record/... ./internal/report/... ./internal/capture/... && git diff --stat` then READ every changed golden line (G1–G8, G-arch). *Re-armed by:* any agent-facing text, schema, archive.
5. `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/... 2>&1 | tee ~/.claude/scratch/mint-prereq/plan2/test.log | grep -c '^--- FAIL'` → 0; read the log's tail. *Re-armed by:* any Go change.
6. `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestTheEiffel|TestASyntheticInActRecord|TestEveryArchivedRun' ./internal/record/migrate/` → PASS (T13, T14, G-arch). *Re-armed by:* `migrate/`, the archive.
7. Zero-hit sweeps, each prints nothing:
   `grep -rnE 'carried_by|CarriedBy|carrier_minted|CarrierMinted|carrying_reading|verify_carried_by' --include=*.go --include=*.md --include=*.js --include=*.json --include=*.sql --include=*.golden plugins/frank-exchange-of-views scripts/agentgen | grep -v 'reserved'`;
   `grep -rnE 'GapClassesOf|FoundSourceGapClassesOf|ReadingGapClasses|ReadingClassField|AppendReading|errReadingHeld|readingAct|pairedSentenceOf|CheckFoundAgainst|carrierTree|repeatReadingRefusal|readingSubject' --include=*.go plugins/frank-exchange-of-views/tools`;
   `grep -rnE 'same act|found-against|found against|finds against' plugins/frank-exchange-of-views/tools/internal/cli/seat/help plugins/frank-exchange-of-views/docs plugins/frank-exchange-of-views/skills plugins/frank-exchange-of-views/agents scripts/agentgen/src plugins/frank-exchange-of-views/tools/internal/terms/terms.json | grep -v 'supersed\|channel'` (the two unrelated "same act" senses).
   *Re-armed by:* any text or code edit.
8. Read each help page before and after, one command at a time: `go -C plugins/frank-exchange-of-views/tools run ./cmd/feov-record lens verify --help` (then `lens mint`, `lens close`, `blue edit`, `show evidence`). *Re-armed by:* any help source.
9. Delete-the-row checks by hand, each reverted: remove 32's row from `readingMintedClasses` → T3 red; remove 25's not-the-work alternative → T6 red; remove the anchor arm's own-failed-check exemption → T8 row 3 red; remove `(finds_against) = true` from `ABSENT` → T6 and the ledger's `absent` tests red; delete `FailingAnchors`' citation case → `capture_test.go` assembly-screen tests red; delete the Verify-column drop in migrate → T13 red (the replay refuses an unknown column); delete the debt's dispatch arm → T20 red; delete the "by ANY seat" from the debt's gap lookup → T19 row 2 red; delete the debt's `lensOwesOf` arm → T20's retirement row red. A row that stays green is a finding. *Re-armed by:* any change to the table, the arm, the option, the translation.
10. `go -C scripts run ./check 2>&1 | tee ~/.claude/scratch/mint-prereq/plan2/check.log | tail -40` → read the fail count in the log. *Re-armed by:* every push.
11. `scripts/universe.sh build`, then `grep -rn 'standing reading' <the universe's installed plugin help>` (a phrase from the new `mint --help`) proves the universe carries the change; one `--smoke` run (one lane, haiku both tiers); read its record: the `verify` table has no gap column; every 32/25 mint has `ReadingBehind`; no plain 32 mint by a seat with no reading. *Re-armed by:* Part 1's merge candidate.
12. **CI only — release fuzz** (`FEOV_RELEASE_GATE=1`, the sharded job in `.github/workflows/hooks.yml`, scheduled and on tag; never run locally by any door): `reading_mints_refused` non-zero for both classes, `reading_mints_landed` non-zero for both subject forms, `second_reading_mints_refused`, `found_source_mints_budget_refused`, `reading_demand_met_by_retire`, `reading_demand_met_by_reread` non-zero; every driven class landed. A logic PR waits for one full pass (cadence ruling 2026-10-29). *Re-armed by:* Part 1's merge candidate.

## VI. Forks — OPEN, none decided here (2–3 options each; cost; recommendation)

**F1 — the location of a found-source reading when its quote stops matching.** A found-source 32's
standing reading is located to the mint's sentence by `readingSentenceIn` (`LocateOnce` of the stored
claim). Blue's edit of a neighbouring clause leaves the claim locating; an edit of the quoted words
makes it locate nowhere, and the reading no longer stands for any sentence.
- (a) **No reading stands; the mint is refused until the lens reads again** (`verify --url … --quote
  <the edited sentence>`). Cost: none; the honest answer — the sentence the source was read against is
  gone. The refusal names the latest reading and says its sentence is not in the report.
- (b) Locate by the citation-free sentence's position (the paragraph, the nearest anchor). Cost: a
  relation recovered from position (facts-are-fields); a wrong sentence reads as a standing reading.
- **Recommendation: (a).** A reading that stands for no sentence owes nothing (§III.4, found-source row), so there is
  nothing to locate for; the only consumer is the mint check, which should refuse.

**F2 — `absent` on a FOUND source.** Today it mints nothing (#1647 F5).
- (a) **Admits no mint and creates no debt**: a source silent on a claim blue never cited it for is no
  defect of the report's; the reading is recorded (information). Cost: none.
- (b) Admit a found-source 25 on it. Cost: 25's `means` is a cited source's; a new meaning for a class.
- **Recommendation: (a).**

**F3 — does a not-the-work copy count as `finds_against` for 25?** The ruling says 25's ground is "the
seat's own reading of that citation with outcome `absent`, or a copy that is not the work".
- (a) **Two different predicates**: `finds_against` (the SourceOutcome option, the ledger's read) is
  true only of `refutes`/`absent` that READ the work — a not-the-work copy stays `unread` in the ledger
  as today; 25's mint check separately admits the seat's standing not-the-work reading of the citation.
  Cost: 25's row in `readingMintedClasses` has two alternatives (one `||`); the 25 gap then also
  answers the owed unread check (`answersOwnFailedCheck`, any class) and is exempt from the budget by
  that rule. Consistent with "the ledger keeps `unread`".
- (b) Make not-the-work a `finds_against` outcome in the ledger too. Cost: `unread` and "found against"
  merge for that copy; `sources_refuted` and the assembly screen would count a sign-in page as a
  refutation; reverses the unread state's definition (`anchorchecks.go:141–143`).
- (c) 25 on `absent` only; a not-the-work reader mints its owed unread gap in any class. Cost: drops
  the ruling's "or a copy that is not the work"; a lens whose cached copy is a different document has
  no class that names that.
- **Recommendation: (a).**

**F4 — estoppel on a found-source 32 (`--quote S`).** The 2026-10-31 ruling exempted a reading's gap
("source refutation is new evidence"). A cited source's 32/25 names no quote and estoppel never asks
of it; a found-source 32 is a quote mint and meets the guard like any gap.
- (a) **No exemption**: the gap pays like any gap; red mints `--supersedes` the prescribing gap, as the
  refusal says. Cost: one logged estoppel where red's found source refutes text blue applied from red's
  own `--new`; reverses the 10-31 exemption for this case.
- (b) Exempt a mint whose class is in `readingMintedClasses` and whose standing reading exists (the same
  check). Cost: a class-keyed arm on the estoppel guard; the exemption the previous audit found hard to
  keep consistent, now narrower.
- **Recommendation: (a)**, as the simpler statement of "the gap pays like any gap"; gblock's 10-31
  ruling is the reason to ask rather than decide.

**F6 — does the gap the reading debt obliges count against the mint budget?** An owed unread/unsound gap
is outside the budget (`ownFailedCheckAnswers`, `mintbudget.go:375`); the ruling calls the 32/25 gap a
graded finding. D7 as drafted says it pays.
- (a) **Counts like any gap.** The lens's budget is the lens's judgement of what matters; a refuted
  citation competes with its other findings. Cost: at the cap the debt cannot be discharged by a mint
  (`requireMintWithinBudget`) nor honestly by a re-read, and the envelope is refused until a gap of the
  lens's closes (K1′) — so under (a) the debt item must be non-blocking (`may`) while the lens's standing
  is at its cap, and the debt readies nobody then; a second rule.
- (b) **Outside the budget while open, as the unread/unsound owed gap is**: `ownAnswersAt` gains the
  gaps of `readingMintedClasses` by the seat about an anchor the seat's standing reading finds against
  (and the found-source 32 quoting its reading's sentence) — one join, the same `standingReadingOf`.
  Cost: a lens at its cap can still open 32/25 gaps without limit, one per refuted citation; the
  exemption's wording on four surfaces (`mintAnswerOutsideBudget`, `verbs.go:334`, `setup.go:89`,
  `mint.md:15`) says "the gap your check or reading obliges".
- (c) Outside the budget only while the debt stands (the first such gap per subject), counted once a
  second lens's gap or a re-read would have discharged it. Cost: a count that moves as other seats act.
- **Recommendation: (b).** One rule for the three obliged gaps (unread, unsound, found against), no
  deadlock, no cap-conditional item; the ruling's "graded finding" is honoured by the grade (the gap's
  materiality is what holds the PASS), not by the count.

**F5 — the one-gap anchor arm's scope (D5).**
- (a) **Every class about an anchor, except the gap that answers the seat's own owed failed check.**
  Cost: a second lens's 32 about C is refused and its point is a finding; one new SQL arm.
- (b) No anchor arm: duplicates of 32 about one citation are admitted (today's behaviour for about-anchor
  gaps). Cost: the duplicate the one-gap rule exists to refuse, on the shape the ruling moves 32/25 to.
- **Recommendation: (a).**

## VII. Inconsistencies between the ruling and the code, reported, not resolved

1. "`stateFoundAgainst` leaves the ledger" — the state has two readers that are not PASS-holders:
   `CitationsFoundAgainst` (`anchorchecks.go:315–323`) → `counts.sources_refuted` (`evidenceview.go:564`,
   documented at `verbs.go:306`), and `FailingAnchors` (`anchorchecks.go:331–340`) → the capture's
   assembly screen (`capture.go:470–495`), which FAILS a captured run's audit when a citation a standing
   reading finds against is still in the assembled report. The plan re-derives both from the standing
   checks (D4) so the count and the screen keep their meaning; but the screen then still calls a
   read-and-refuted citation a failure of the run, which is the judgement the re-stated ruling moves to
   the lens's grade. Whether the assembly screen's citation arm stays, becomes informational, or reads
   the open material 32/25 gaps instead is gblock's; not decided here.
1′. The debt is "modelled on the owed unread/unsound gap", but that gap's home is `BlockerFailedCheck`
   — a PASS blocker by kind (`passBlockersOf`, `passblockers.go:672–721`; every `BlockerKind` holds
   `pass_permitted`). A debt that must not hold the PASS cannot be a `Blocker`, so it is a parallel fold
   (`readingDebtsOf`) with its own arm on each of the four surfaces (work list, envelope, retirement,
   dispatch) rather than a row in the blocker table. The envelope gate needs no new code (it reads
   `Blocks`); the other three gain an arm each. The one-list principle the gate was built on (#1202)
   now has two lists for "what a lens owes": `passBlockersOf` and `readingDebtsOf`, joined only in
   `lensOwesOf`.
2. "Keyed on the subject the mint already names" — for a cited source that subject must be the
   citation (`--about-kind anchor`), not the sentence, or the gap reaches no anchor after
   `carrying_reading` goes (`demand_reach` `views.go:493–495` is the only reach a quote gap had) and its
   `source` demand can be met only by a new cite. `ABOUT_KIND_ANCHOR`'s `means` (`record.proto:1179`) and
   `mint --about-kind`'s usage (`mint.go:220–221`) today name only a failed check as the reason to mint
   about a citation; both are rewritten (§III.1, §III.3). The one-gap rule today exempts every
   about-anchor gap (`onegapperclass.go:25–30`), so without D5 the ruling's "one gap of that class on
   the sentence" is not enforced for the shape 32/25 move to.
3. "`mint` of 32/25 … refused unless the minting seat holds its own standing reading" with "classes
   … stay plain mint; 31's hand check stays": the class-to-outcome rule has no declared home (the option
   mechanism was dropped), so it is a Go table in `record` (`readingMintedClasses`, §III.3) beside 31's
   hand check at `record.go:1931` — two hand checks of two shapes in two files. The plan puts the table
   in one new file with its delete-the-row test; whether gblock wants 31's check moved beside it is not
   asked here (31's is a shape check the replay runs; this one is a live judgement; they differ in kind).
4. The budget exemption's deletion (D7) reverses F6 (2026-10-30, narrowed 2026-11-03) and the
   estoppel question (F4) reverses the 2026-10-31 exemption for the found-source case; both are
   consequences of "the gap pays like any gap" rather than things the ruling says.
