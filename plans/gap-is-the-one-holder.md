# A gap is the one PASS-holder — `verify` records, `mint` of 32/25 stands on a reading (#1727)

> **Status: DRAFT, revised 2026-11-06 on gblock's rulings; unaudited.** Written on `plan/mint-prerequisites`
> merged with `origin/main` `7d676a887` (EventSchema **64**, `requirements.json:3`, re-verified against
> `origin/main` `ec3a3e559` the same day; #1739 anchor-lifecycle is MERGED). Replaces
> `plans/mint-prerequisites.md` (renamed with `git mv`; its mechanism — a `MintPrerequisite` option, a
> `rests_on` link, a generic check — was dropped by gblock 2026-11-06). The scope is the ruling quoted in
> §I; option wording throughout is the lead agent's, not gblock's. The first draft's six forks are RULED
> (§VI); two new forks the rulings created are OPEN (F7, F8), neither decided here. Every census row in
> §III.9 was re-run by grep on this tree.

## I. Summary & Goals

**Objective.** A reading (`lens verify`) records what one source did for one sentence and holds nothing
by itself. A gap is the ONE thing that holds a PASS for a defect and the one thing blue answers. `mint`
of `claim_contradicts_source` (32) and `citation_figure_misattribution` (25) is refused unless the
minting seat holds its own STANDING reading of the mint's subject with the matching outcome — keyed on
the subject the mint already names, with no new field and no event id. A lens whose reading finds a
source against its sentence OWES the graded gap (or a later re-read): that unfinished promise is an
ordinary PASS blocker on the one list, cleared by the mint — after which the gap's materiality decides
the PASS. The in-act mint, the reading-to-gap link (`Verify.carried_by`, `carrier_minted`), the held
branch, the `gap_classes`/`found_source_gap_classes` options, the found-against ledger state and the
budget exemption for reading-minted gaps are deleted. One epoch bump; `migrate` brings the one archived
run that holds the old shape forward.

**The ruling (gblock, 2026-11-06, issue #1727, the two "Re-ruled"/"Amendment" comments), verbatim where
quoted:**

- A reading (`verify`) only records what one source did for one sentence. It holds nothing by itself.
- A gap is the ONE PASS-holder for a defect, and the one thing blue answers. This reverses #1647's F4
  (found-against as a third failed-check state); the ledger keeps `unread` and `unsound`.
- A citation nobody has read, or whose copy could not be read (unchecked, unread): blocks the PASS as
  today. Unchanged. A citation that was read and found wrong: the reading records; it holds nothing.
  "Found against" stops being a ledger state.
- The lens that read it owes a graded gap of class 32 or 25, or a later re-read of the same subject. The
  gap's materiality decides the PASS.
- `mint` of 32 and 25 is refused unless the minting seat holds its own standing reading of the mint's
  subject (the citation anchor, or the sentence for a found source) with the matching outcome. Keyed on
  the subject the mint already names: no new field, no event ids.
- Classes 2, 23, 27, 31 are judgements and stay plain mint; 31's hand check stays. 29's backing
  found-source reading records only.
- The in-act mint, `carried_by`, `carrier_minted`, and the `gap_classes`/`found_source_gap_classes`
  options go; "finds against" is a `finds_against` bool option on SourceOutcome.
- Part 0 stands: commit the Eiffel smoke archive to run-archive/.

**The debt's home, ruled 2026-11-06 (later; "Lens, chair sees").** The lens owes the gap or the re-read,
enforced by its own envelope in the same sitting. An ORPHANED found-wrong reading — the lens's sitting
ended without discharging it — BLOCKS the PASS as red's unfinished promise, shown on the chair's
pre-PASS blocker list with the lens as the seat whose act clears it. So the debt is an ordinary
`BlockerKind` built in `passBlockersOf`, one list (#1202), not a parallel fold. Further rulings the same
day: the owed gap is OUTSIDE the mint budget while open (F6 (b)); NO estoppel exemption for a
found-source 32 (F4 (a)); the capture's assembly screen screens the GAP, not the reading (§III.5); a
LOST QUOTE leaves the debt standing unless the edit took red's suggested text, then estoppel discharges
it (F1); a found-source `absent` creates the same debt (F2 — the class question is open, F7); two
predicates for not-the-work (F3 (a)); one open gap per class per anchor for every class (F5 (a)).

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
   for 25 with `absent` or a not-the-work copy; a found-source 32 by `--quote S` is admitted only on
   X's standing `refutes` reading of a found source for S (T3–T8).
3. No Verify names a gap and no gap names a reading, in the proto, the schema, the views, the help, the
   constitutions, the terms registry or the fuzz (§V check 7 zero-hit sweeps).
4. The ledger's failed-check states are `unread` and `unsound`; no `failed_check` blocker and no close
   refusal is produced by a `refutes`/`absent` reading (T9, T10). A standing found-against reading no
   gap of its class answers is a `reading_debt` blocker owned by its lens: it refuses the lens's
   envelope, readies the lens once per cause, is on the chair's list naming the lens, and holds
   `pass_permitted` false; the lens's mint of the class about the subject discharges it, after which the
   gap's materiality alone holds the PASS (T18–T24).
5. The reading-minted exemption (`carrier_minted`) is gone from the budget; the 32/25 gap a standing
   found-against reading of the minter's obliges is outside the budget while open, as the owed
   unread/unsound gap is (T11).
6. `feov-migrate` brings the Eiffel archive forward with zero refusals: each archived in-act reading
   becomes its Verify, Mint and Anchor as three events with no link fields; the held reading becomes a
   Verify alone (T13, T14).
7. The capture's assembly screen FAILS a run with an open MATERIAL 32/25 gap whose subject the
   assembled report still carries, or with a standing `reading_debt`; a refuted citation whose gap was
   graded low and closed passes it (T25).
8. The release fuzz drives the mint refusal, the admitted mint for both classes and both subject forms,
   the debt's PASS refusal and each discharge, with no zero tallies (CI only, §V check 12).

**Out of scope.** Any change to a class's meaning; prerequisites for 2, 23, 27, 31 (ruled plain mint;
31's hand check `record.go:1931` stays as it is); #1695 (findings fold into gaps).

## II. Technical Context

**Record shapes today (`origin/main` `7d676a887`, epoch 64; unchanged under `record/` at `ec3a3e559`).**

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
  carrier. The blocker clears by the ledger alone — the reader's backing re-read, blue's edit or retire —
  whatever gap stands. Three non-blocker readers read the same state: `CitationsFoundAgainst`
  (`anchorchecks.go:315`) → `counts.sources_refuted` (`evidenceview.go:564`); `FailingAnchors`
  (`anchorchecks.go:331`, through `EvidenceTables.Failed = "found_against"`, `evidencetables.go:56`) →
  the capture's assembly screen (`capture.go:484–495`, an audit FAIL when a refuted citation is still in
  the assembled report).
- The PASS gate's list is ONE fold, `passBlockersOf` (`passblockers.go:672–698`): every `BlockerKind`
  (`:28–44`, the `blockerKinds` table `:380+`, `TestEveryPassBlockerKindAgreesOnEverySurface`) holds
  `pass_permitted`; the lens work list (`sitting.go:430–472`), the chair's list (`:365`), `lensOwesOf`
  (`retirement.go:128–175`), dispatch's Source 4 (`dispatch.go:528–616`) and `verify`'s read-back
  (`verify.go:496–515`) each read it. `eventBlockersOf` (`:709`) is the stream-only arms; the gap and
  lens arms take `gaps []blockerGap` and `fresh` from SQL at each caller (`passBlockerReadsAt`).
- Estoppel: `appliedVerbatim` (`estoppel.go:55–77`) maps each gap whose `fix_new` a `BlueEdit` with
  `applied_verbatim` applied; `EstoppelConflict(f, quote)` (`:90`) is containment of the visible text
  either way; the guard is attached by `mint` on `--quote` (`mint.go:155–158`, `:576`) and not by
  `verify` (`reading.go:89–91`). A `BlueEdit` carries `old`, `new`, `answers`, `applied_verbatim`
  (`record.proto:1637–1646`).
- The mint budget exempts a reading-minted gap of a citation (`mintbudget.go:257` `w.act.pairs(m)`,
  `ownAnswersAt` `:325–345`) and the gaps answering the seat's own unread/unsound check
  (`ownFailedCheckAnswers`, `failedcheck.go:436`).
- `requireOneGapPerClassAndSubject` (`onegapperclass.go:50`) applies to a quote subject (sentence) and an
  inquiry subject; a gap `--about-kind anchor` is exempt (`:25–30`).
- `judgeAct`'s Mint arm (`actjudge.go:40–64`): about-anchor live, one-gap, superseded off docket,
  successor carries demand, budget, report voice. The replay never asks it (`actjudge.go:23–25`).
- `mint`'s subject: `--quote S` (location; placed by `seat.AppendPlaced`, `mint.go:164`) or
  `--about-kind anchor|section|inquiry|gap --about <id>` (`mint.go:220–222`;
  `ABOUT_KIND_ANCHOR` `means` at `record.proto:1179`).
- The report text is SQL (`RenderAt(q)`, `writeview.go:40`), never a fold over `evs`: a blocker arm that
  locates a sentence takes the rendered report as an input at the caller, as `gaps` and `fresh` are.

**Measured inputs (re-checked where cheap).**

| Measure | Value | How |
|---|---|---|
| Archived mints of 25/32 after `migrate`'s slug translation (`migrate/gapclass.go:43,47,56` → 25) | 25: 4 mints in 4 runs; 32: 6 in 1 run (Eiffel: 5 in-act, 1 plain G-b2ba7d66) | python over each `run-archive/*.tar.gz` `records/record.db` (2026-11-05) |
| Archives under `run-archive/` holding a Verify with `carried_by` | 0 of 28 SQLite archives | same scan |
| The Eiffel record | `~/.claude/scratch/contradiction-impl/elicit-adversary/2026-11-04_eiffel-tower-height.tar.gz`, 370 KB; layout `inputs/run-config.json`, `records/record.db`, `computations/<sha>/{acts,dir,rerun}`; 59 readings, 7 with `carried_by` (6 in-act: events 167, 182, 542, 791, 794, 797; 1 held: 798) | `tar tzf`; sqlite read (2026-11-05) |
| That record under main's `feov-migrate` at epoch 61 | 1213 events in → 1203 out, 0 refusals | 2026-11-05; re-run at 65 is T13 |
| Epoch | `origin/main` 64 (`ec3a3e559`, 2026-11-06); this plan takes **65** | `git show origin/main:plugins/frank-exchange-of-views/requirements.json` |
| Names coined here, grepped free on the Go tree, `debate.js`, `agents/`, `skills/` | `reading_debt`, `BlockerReadingDebt`, `readingDebtsOf`, `standingReadingOf`, `readingMintedClasses`, `AssemblyFailures`, `finds_against` | 0 hits outside `plans/` |

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
| D3 | `mint` of 32/25 is a live act judgement, `requireOwnStandingReading`, keyed on the mint's subject: a CITED source's defect names the citation (`--about-kind anchor --about C`), a FOUND source's defect names the sentence (`--quote S`); the seat's standing reading of that subject must carry the class's outcome; 25 ALSO admits the seat's standing not-the-work reading of C (F3 (a), RULED) | §III.3 |
| D4 | `stateFoundAgainst` leaves the ledger: a citation with a `refutes`/`absent` reading is holding or unchecked by the ordinary rules; `CheckFoundAgainst`, its blocker, its close refusal and its carriers go. The per-check `holds` stays (the demand's "check again that holds") | §III.5 |
| D4′ | The reading debt is a `BlockerKind`, `reading_debt`, built in `passBlockersOf` from the ledger's standing checks that do not hold (cited) and the seat's latest found-source readings (found): owned by the lens, owed while no gap of the class stands on the subject, discharged by the mint, the re-read, blue's retire or an estopped edit; it refuses the lens's envelope, readies the lens once per cause, is on the chair's list and holds the PASS. One list (#1202, RULED) | §III.4 |
| D5 | The one-gap rule gains an anchor arm: one open gap per class about one anchor, for EVERY class, except the gap a seat's own owed unread/unsound check obliges (F5 (a), RULED) | §III.3 |
| D6 | A gap carries nothing of its reading by field; the reading behind a gap is one Go helper, `record.ReadingBehind`, joining (seat, subject, class); `carrying_reading` and `demand_reach`'s reading arm are deleted | §III.6 |
| D7 | The `carrier_minted` budget exemption is deleted; the 32/25 gap the minter's standing found-against reading obliges is outside the budget while open, through `ownAnswersAt`'s join on `standingReadingOf` (F6 (b), RULED); a 25 gap about a citation whose unread check the minter owes is exempt by the existing own-failed-check rule | §III.3 |
| D8 | `requireNotARepeatReading`'s reads arm is deleted; the unread arm stays | §III.5 |
| D9 | `migrate`: the in-act reading act machinery is deleted; each archived member comes forward as its own event with the two fields dropped; stated residue | §III.7 |
| D10 | Estoppel: no exemption for a found-source 32 — the guard asks of every `--quote` mint as today, and `verify` attaches none (F4 (a), RULED; reverses the 2026-10-31 exemption for this case) | §III.3 |
| D11 | The capture's assembly screen screens the GAP: FAIL on an open MATERIAL 32/25 gap whose subject the assembled report still carries, or on a standing `reading_debt`; `counts.sources_refuted` counts the ledger's standing found-against readings (information) | §III.5 |

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

**Alternative considered for D4′'s source.** Keep `stateFoundAgainst`/`CheckFoundAgainst` and re-key the
`failed_check` blocker's discharge to the mint. Rejected as the LARGER change in meaning for a smaller
diff: the ledger's state is per ANCHOR ("what does the record say of this sentence") and two lenses
reading C against are one state; the debt is per READING ("what does this lens still owe") and they
are two debts with two owners. The `failed_check` kind's gloss, item, refusal, plan line, close
refusal (`requireFailedCheckAnswered`), repeat refusal, `FailingAnchors` row and `sources_refuted` all
key on the state and would each need a found-against exception — the shape #1647 F4 built and this
ruling reverses. The debt therefore reads the ledger's standing checks (`anchorStanding.standing()`,
`!c.holds`) as its source for the cited case and never its state; `stateFoundAgainst` is deleted.

### III.1 The record — `internal/record/recordpb/record.proto` [MODIFY] (contract: event schema, epoch 65)

- **`Verify` [DELETE fields]:** `carried_by` (21), `carrier_minted` (22) and their comment `:1431–1441`;
  `reserved 21, 22; reserved "carried_by", "carrier_minted";` beside the existing reserved list.
- **`SourceOutcome` options [DELETE, ADD]:** delete `gap_classes` (50014, `:152–158`) and
  `found_source_gap_classes` (50015, `:160–165`), reserve both numbers with a one-line comment as 50007
  and 50010 are. Add `optional bool finds_against = 50016;` on `EnumValueOptions` (number grepped free):
  "the source was read and says otherwise than the sentence, or nothing of it — the outcome the anchor
  ledger reads as a reading against the sentence, the outcome whose lens owes a graded gap or a
  re-read, and the outcome a mint of a class defined by such a reading stands on". Set
  `[(finds_against) = true]` on `REFUTES` and `ABSENT`. Each outcome's `means` (`:531–542`) loses its
  "mints … in the same act" clause and says what the reading records (A5 in §III.8).
- **`GapClass` [MODIFY `means` only]:** 25 (`:460`) and 32 (`:483`) each gain one clause: "minted on your
  own standing reading of the citation (`absent`, or a copy that is not the work)" / "minted on your own
  standing reading that refutes: of the citation you name, or of a source you found for the sentence you
  quote". No option, no value change. The class-to-outcome rule lives in Go (§III.3), as 31's does.
- **`AboutKind.ANCHOR` `means` (`:1179`) [MODIFY]:** "…a citation whose work went unread, or whose source
  your reading finds against — a cited source's defect names the citation, never the sentence".
- **Epoch:** `requirements.json` `eventSchema` 64 → **65**; `go -C scripts run ./schemagen`. A PR that
  lands before this one and bumps takes 65 and this plan takes the next.
- **Generated:** `go -C scripts run ./protogen` (`record.pb.go`, `decode_gen.go:1475` verify scan,
  `record.proto.sha256`); `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen`
  (`annotations_gen.go`: the `GapClasses`/`FoundSourceGapClasses` literals go, `FindsAgainst` arrives;
  `recordpb/annotations.go:39–42,92` and `devcmd/vocabgen/main.go:169–170` change by hand).
- **Schema:** `recordsql/testdata/schema.sql` (`verify` columns `:789–797`; the `carrying_reading` view
  `:1419–1423`; `mint`/comments `:280–282`, `:374–379`) and `schema.epoch.json`.

### III.2 `lens verify` records only — `internal/cli/lens/verify.go` [MODIFY] (contract: CLI flag set)

| Site | Change |
|---|---|
| package comment `:87–93` ("# A reading that finds a defect mints its gap") | rewritten: a reading records what the source did for the sentence and holds nothing; a defect it shows is a gap `mint` opens, standing on the reading; a reading that finds against is its lens's debt until that gap or a re-read |
| `gapFlagsRequiredUnless` `:99`, `readingRequiredGapFlags` `:106`, `readingGapFlags` `:109`, `var gf` `:112`, `gf.register(c, readingSubject)` `:162`, the `RequiredUnless` loop `:163–165` | deleted: every gap flag is an unknown flag on `verify` |
| `seat.Supplies` for `gap_id`, `class_material`, `carried_by`, `carrier_minted` `:157–160` | deleted |
| `--title` usage `:146` | "kept on your reading beside its url; a gap you mint on this reading names the source through it" |
| `writeVerify` `:367–399` (`classes`, `backing`, the gap-flag refusals, `gf.fill`, the class check) | the reading is written alone; `markedSentenceOf` `:373` stays (a backing found-source reading of a marked sentence is refused — `refuseMarked` `:591–616` loses "mints unsourced_claim" and says a found source backs only an unmarked sentence) |
| `recordReading` `:412–517` | the `gap != nil` branch deleted (Anchor, `PlacesToSentenceOf`, `uncitedSentence`, `verifiedFix`, `AppendReading`); `verifyResult.Gap` deleted; the result gains `event` (the reading's `events.id`); the read-back `:496–515` words every blocker on the anchor by its kind through `Blocker.WorkItem()` — the `failed_check` arm keeps unread rows, and a `reading_debt` this reading raised is worded as the seat's own (§III.4) |
| `givenGapFlags` `:519`, `mintsNothing` `:529`, `classWords` `:552`, `readingGap` + `words` `:644–672`, `failedCheckWords.words`' found-against arm `:694–709`, `sentenceOfRefusal` `:762–773` | deleted |
| `verifyResult.Human` | names the event: `recorded as event N`; a `refutes`/`absent` adds the debt's own item sentence (one source: the kind row's `item`) |

`mint.go`'s `readingSubject` (`:259–266`), `gapSubject.span` (`:249`) and `record.ReadingClassField`
(`enums.go:344–360`) are deleted; `gapFlags` (`:235–238`) is `mint`'s alone and its comment says so.
`seat.go:1085` (`verify --reason`) loses "Where this reading mints a gap…".

### III.3 `mint` of 32/25 stands on the seat's own standing reading — `internal/record/` [NEW `standingreading.go`; MODIFY `actjudge.go`, `onegapperclass.go`, `mintbudget.go`] and `cli/lens/mint.go`

**The table** (one Go table, `readingMintedClasses`, one row per class; `TestEveryReadingMintedClassIsDriven`
fails a row no test refuses and admits). It is kept as a Go table beside 31's hand check at
`record.go:1931` (§VII item 2): two hand checks of two kinds — 31's is a shape check the replay runs,
this one a live judgement — and the option mechanism that would have housed both was dropped by ruling.

| Class | Subject the mint names | The seat's standing reading that admits it |
|---|---|---|
| 32 `claim_contradicts_source` | `--about-kind anchor --about C` (a cited source) | of citation C, reads the work (`VerifyReads`), outcome `refutes` |
| 32 | `--quote S` (a found source) | `verify --url` by the seat whose `claim` locates once in the landing report in S's sentence, reads the work, outcome `refutes` |
| 25 `citation_figure_misattribution` | `--about-kind anchor --about C` | of citation C: reads the work with outcome `absent`, OR `source_completeness` `not_the_work` (F3 (a), RULED: two predicates — the ledger's `finds_against` is refutes/absent on a copy that READ the work; this row's second alternative is the mint check's alone, and a not-the-work copy stays `unread` in the ledger) |
| 25 | `--quote S` | refused: the class names a cited source; a found source's `absent` is F7's question (§III.4, §VI) |

**"Standing"** = the seat's LATEST reading of that subject since the subject's sentence last changed:
for a citation, the ledger's own rule (`anchorStanding.sinceEdit`, anchor not retired — read through
`anchorChecksOf`, which the check reuses rather than restating; the ledger's `stands` additionally
requires `reads`, so 25's not-the-work alternative reads the seat's latest check since the edit
directly, `reads` or not); for a found source, the seat's latest reading of that url, whose stored claim
locates once in the landing report inside the mint's sentence (`readingSentenceIn`, `reading.go:185`,
stays for this). A reading of a sentence an edit has since changed does not stand for the MINT: a mint
of 32 after blue's edit is refused until the seat reads again — the ruling's "standing" clause. (What the
seat still OWES on such a reading is §III.4's lost-quote rule.) `standingReadingOf(seat, subject)` is the
one statement of this, shared by the mint check, `ReadingBehind`, the budget join and the debt fold.

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
citation it contradicts (`--about-kind anchor --about C-…`) where the source is one blue cites — a
cited source's defect is keyed on the citation, the subject the mint names". For a `--quote` 25:
"`--class citation_figure_misattribution` is a cited source's defect: name the citation
(`--about-kind anchor --about C-…`) your reading of it found absent or not the work".

**The one-gap rule [MODIFY `onegapperclass.go`] (D5, F5 (a) RULED):** delete the held arm (`:91–93`),
`pairedSentenceOf` (`:101–108`), `errReadingHeld`'s use and the "A READING'S MINT IS NOT REFUSED, IT IS
HELD" paragraph (`:40–44`). ADD the anchor arm: where `about_kind` is `anchor`, an OPEN gap of the same
class about the same anchor — ANY class — refuses the mint with `oneGapRefusal` ("citation C-…" as the
subject) — the inquiry arm's SQL shape (`openGapsOfClassAt`, `:117–143`) with `about_kind = 'anchor'` —
UNLESS the mint answers the seat's own owed unread/unsound check (`answersOwnFailedCheckAt`,
`mintbudget.go:375`), the one exception. The comment `:25–30` is rewritten: a gap about an anchor IS a
subject; what the rule still does not apply to is a section or a gap. Consequence, stated: a second lens
Y whose reading also refutes C mints no second 32 about C; Y's debt is discharged by X's gap (§III.4) and
Y's point, if any, is a finding about it.

**The budget [MODIFY `mintbudget.go`] (D7, F6 (b) RULED):** delete the `w.act.pairs(m)` arm (`:252–259`)
and the `carried_by` query in `ownAnswersAt` (`:325–345`). `ownAnswersAt` keeps `ownFailedCheckAnswers`
and `ownsInquiryGap` and gains one join: each open gap of a `readingMintedClasses` class by seatID
whose subject seatID's standing reading finds against — about anchor C where `standingReadingOf(seat, C)`
is `refutes`/`absent`/not-the-work, or quoting the sentence `standingReadingOf(seat, url)` locates to —
so the owed gap is outside the budget exactly while the debt it discharges would otherwise stand, as the
owed unread/unsound gap is outside it while its check stands. `mintStanding.Answering` counts it. The
header comment `:32–44` and `mintAnswerOutsideBudget` (`:308`), `verbs.go:334`, `setup.go:89`,
`mint.md:15` say one sentence: "a gap about an anchor whose check of yours read nothing or judged a
computation unsound, or of the class your standing reading of its source obliges, takes no mint; so does
the logic lens's gap about an inquiry". Cost, stated: a lens at its cap can still open 32/25 gaps, one
per refuted subject (the one-gap rule bounds it). Nothing is exempt for having been minted by a reading.

**Estoppel [`cli/lens/mint.go:155–158`, `reading.go:89–91`] (D10, F4 (a) RULED).** The guard is keyed on
`--quote` (`requireNotEstopped(l, quote, …)`, `mint.go:576`); a cited source's 32/25 names no quote, so
estoppel never asks of it; a found-source 32 is a quote mint and meets the guard like any gap: red
mints `--supersedes` the prescribing gap, as the refusal says. The "verify ATTACHES NO SUCH GUARD"
paragraphs are deleted with the exemption they stated. One logged estoppel where red's found source
refutes text blue applied from red's own `--new` is the accepted cost.

**`cli/lens/mint.go` [MODIFY]:** `--about-kind` usage (`:220–221`): "anchor — a citation or computation;
a CITED source's defect (claim_contradicts_source, citation_figure_misattribution) names the citation
here, the subject your reading of it is keyed on; a source YOU found is a `--quote` of the sentence";
`--class` usage for 32/25 is the enum help's, rendered from the `means` (§III.1). `mint.md` (§III.8)
carries the rule in one paragraph.

### III.4 The reading debt — `BlockerReadingDebt` (D4′) [NEW in `record/standingreading.go`; MODIFY `passblockers.go`, `sitting.go`, `dispatch.go`, `verdict.go`]

**The fact.** `readingDebtsOf(evs, win, gaps, report) []Blocker`, one `Blocker{Kind: BlockerReadingDebt}`
per lens X and per reading of X's that finds against (`VerifyReads` and `FindsAgainstOf(outcome)`) and
STANDS, while NO gap of the class the outcome obliges stands on the subject:

| Reading | Stands while | The gap that discharges it | Also discharged by |
|---|---|---|---|
| of citation C (`refutes` → 32; `absent` → 25) | X's latest check of C that reads is this one, from a sitting opened after the latest edit naming C, C not retired — `anchorStanding.standing()` with `!c.holds`, the ledger's rule | a gap of the class ABOUT ANCHOR C by ANY seat that is open now, or was minted from this reading's sitting on (closed since) | X's later reading of C that reads and does not find against (the ledger replaces X's standing check); blue's edit naming C (the check is of a sentence the report no longer carries); blue's retire of C |
| of a found source U for sentence S | X's latest reading of U (`standingReadingOf(X, U)`), whatever the stored claim now locates to | a 32 by any seat whose quote is in the sentence the claim locates to (`gapSentenceIn` = `readingSentenceIn`), open now or minted from this reading's sitting on; for `absent`, F7 | X's later reading of U that does not find against (any sentence — X's latest reading of U is the one that stands); the estopped edit (below); F8 where the sentence was deleted |

`Blocker` fields: `Subject` C or U; `Owner` X; `since` the reading's `events.id`; `outcome` (new field,
the reading's); `Detail` the sentence as the reading read it (found) or "" (cited); `lost` (new,
`int64`) the `BlueEdit` events.id after which the found claim stopped locating, 0 while it locates.
`owed()` is true for the kind while it stands — the debt IS the owing. `Blocker.WorkItem()`, the kind's
`refusal` and `why` are one row of `blockerKinds`, so `TestEveryPassBlockerKindAgreesOnEverySurface`
gains a row and every surface words it from the row. Cited and found readings are two shapes of ONE
kind: the item names the subject by its form (`readingForm`).

**"Open now, or minted from the reading's sitting on."** Ruling 1 says the MINT discharges the debt and
ruling 4 says a gap graded low and closed does not fail the run; ruling 8 says a second lens's debt is
discharged by any seat's OPEN gap. The three agree under one predicate: a gap of the class on the subject
that is open, or that was minted at or after the reading's sitting opened (whatever its state now). A
gap closed BEFORE X read is not it: X read after that adjudication and found against anew, and owes a
fresh gap (the one-gap rule admits one, the old gap being closed). Stated as a consequence, not a fork
(§VII item 4).

**The surfaces — all through the one list, no new arm but the fold's own:**

| Surface | Reads | The debt |
|---|---|---|
| `passBlockersOf` (`passblockers.go:672`) | every kind | `readingDebtsOf` is one more arm; the signature gains `report string`, read by `RenderAt(q)` in `passBlockerReadsAt` and at `sitting.go:365, 430` beside `blockerGapsOfStates`; `blockerGap` gains `class`, `aboutKind`, `aboutRef`, `location` (gap-view columns) so the debt's gap lookup reads the set every other arm reads |
| the PASS gate (`requireNoBlockers` `:825`), `pass_permitted`, `dispatch.md:25`'s "a blocker another seat must clear", `verdict.go:149` `anchorsStanding` | kinds | the kind holds the PASS like any other; the outcome reason's clause gains the debt's subjects ("…and each citation or found source a lens's reading finds against with no gap of its class standing") |
| the chair's list (`sitting.go:365` `hold`) | `AnotherSeatClears: !ChairOwned()` | item: "lens X's reading of citation C-… at event N finds the source against its sentence (`refutes`), and no gap of class claim_contradicts_source stands on it — PASS is refused while it stands; X clears it by minting that gap about the citation, graded as it judges, or by reading the citation again" |
| the lens's work list (`sitting.go:430–434`) | `b.Kind == BlockerFailedCheck && b.Owner == seatID` → `add` while `owed()` | the same arm, keyed on `b.owed() && b.Owner == seatID` for `failed_check` and `reading_debt` alike; `add` (Blocks true) — the envelope gate (`envelope.go:40–64` → `Sitting.Owed()` → `hookcmd.go:354`) refuses the `StructuredOutput` call with no new code |
| `lensOwesOf` (`retirement.go:128–175`) | `events` blockers with `owed()` | the debt is named `reading_debt <subject>`; the lens does not retire while it owes one (`retired()` `:72`) |
| dispatch Source 4 (`dispatch.go:528–616`) | `BlockerFailedCheck` with `owed()` → readied once per cause (`firstAfter(ends.through[owner], b.since)`) | the same `case` admits `reading_debt`; the plan's `Why`: "X: its reading of C-… at event N finds the source against the sentence and no gap of class … stands on it — ready, to mint the gap or read again"; `failedCheckReadiness`' found-against arm (`:1396–1405`) is deleted |
| `verify`'s read-back (`verify.go:496–515`) | `BlockerFailedCheck` on the anchor | every blocker on the anchor by `WorkItem()`; the one this reading raised is prefixed "this reading is your debt until …" |

**How this differs from today's `stateFoundAgainst`.** Today the found-against `failed_check` blocker
holds the PASS until the reader's backing re-read, blue's edit or blue's retire — WHATEVER gap stands
(`failedcheck.go:39–47`), and its reader owes nothing. Under D4′ the lens's MINT discharges the blocker:
from the mint on, the gap's materiality decides the PASS, and a low gap closed "not a defect" leaves a
refuted citation in a PASSing report by the lens's grade. The debt is per reading (two lenses, two
debts, two owners); the state was per anchor. The close of the gap is refused by nothing but its demand
(`requireDemandMet`; `requireFailedCheckAnswered` names unread and unsound only).

**Dispatch.** The owner is readied ONCE PER CAUSE, as every owed blocker is. A lens that sat through a
sitting opened after the reading and left the debt is not readied again for it, is not retired (it
owes), and the debt still holds the PASS — the run then ends UNVERIFIED with the outcome naming it, as
an unminted unread gap ends it today. An abnormal end (`shortCallLimit`, `shortNeverClosed`) is not sat
through and readies the lens again (`satThroughOf`).

**The lost quote (F1, RULED "Debt stands"; refined: "unless it took red suggested text, then estoppel
applies").** When blue's edit changes the words the found-source reading's `claim` quoted, the claim no
longer `LocateOnce`s and the mint check refuses a `--quote` 32 until X reads again (the mint half of
F1 (a) stands); the DEBT stands. Located after the quote fails: the reading's `events.id` and seat
identify it; the edit that moved it is the first `BlueEdit` after the reading whose `old` contains the
claim's visible text (`anchortext.Visible` containment — the comparison `EstoppelConflict` already
makes) — `lost` names it, and the item says: "your reading of U at event N read it against '<claim>';
blue's edit at event M changed that sentence to '<the edit's new, as a sentence>' — read U against the
sentence as it now stands, or mint on a reading of it". It clears by X's later reading of U (its latest
reading of U is the one that stands; a reading that finds against the NEW sentence creates a fresh debt
on it, with its own mint route), or by a 32 on the new sentence (which, by the mint check, needs that
re-read first). **Estoppel discharge:** where the edit at M carries `applied_verbatim` (whatever gap it
`answers`, whichever red seat minted it — red is the party bound, `estoppel.go:82–87`), the debt is
discharged: red wrote the new sentence, so red holds no unfinished promise against it. X may still re-read
U against it; a finds-against re-read creates a fresh debt, whose mint route is the estopped one — a 32
`--supersedes` the prescribing gap (D10). Residue: where no edit's `old` holds the claim (the report was
re-ingested whole; the claim now occurs twice), `lost` is 0, the item names the reading alone, and the
debt clears by X's re-read of U or any seat's 32 on the sentence the claim locates to; where the edit's
`new` carries no sentence (deleted outright), F8.

**A found source read `absent` (F2, RULED "Debt like refutes").** The same debt; the class that
discharges it is F7 (§VI) — no shipped class names "a found page silent on the sentence": 25's `means`
is a cited source's and its `--quote` form is refused; 32 is "states otherwise", not silence; 29 is the
REPORT's want of a citation, which the found page's silence neither shows nor changes. The plan ships
F7's ruled option; until ruled, T19's found-`absent` rows are written against the fork's options.

**Census: information surfaces that show a found-against reading, unchanged by the debt** — `show evidence`
source and `independent` rows (`evidenceview.go:225–266, 300–303`; `CarriedBy` `:407` deleted);
`counts.sources_refuted` (`:564`, §III.5); evidence.md "Citations" (`report/citations.go:21–60`);
docket.md's "read for the sentence" line (`assemble.go:1098–1125`, per gap via `ReadingBehind`, §III.6);
the chair's spot-check (`view.go:672–677`, samples closures, reads no reading).

### III.5 The found-against failed check is deleted; the screen reads the gap [MODIFY `anchorchecks.go`, `failedcheck.go`, `passblockers.go`, `dispatch.go`, `evidencetables.go`, `retirement.go`, `capture.go`]

- `CheckState` keeps `unread` and `unsound`; `CheckFoundAgainst` (`failedcheck.go:57`) is deleted, and
  with it: `failedCheck.outcome` (`:72–73`), `addFoundAgainst` and the `stateFoundAgainst` case
  (`:112–123`, `:161–166`), `carrierTree` (`:172–204`), `owed()`'s state test (`:89` → `answeredBy == ""`),
  the `CheckFoundAgainst` skips in `requireNotARepeatFailedCheck` (`:276`) and `ownFailedCheckAnswers`
  (`:436`), the found-against arm of `requireFailedCheckAnswered` (`:408–413`), the header paragraph
  `:39–47`; `answersOwnFailedCheck`'s comment `:241–243`.
- `anchorchecks.go`: `stateFoundAgainst` LEAVES the ledger (`:146–148`, `:159–160`; the `state()` loop
  `:187–195` becomes the computation-only `unsound` test; the header `:125–130`). A citation whose
  standing readings read the work is `stateHolds` whatever they found; one read by nobody since the
  edit is `stateUnchecked`; one read only by `unreachable`/not-the-work copies is `stateUnread`. The
  per-check `holds` (`anchorCheck.holds` `:56–59`, `:279` `reads && !FindsAgainst(b)`) STAYS: it is the
  demand's "check again that holds" (`latest_check`, `verifyHolds` `evidencetables.go:84` →
  `!FindsAgainstOf(outcome)`), the ledger's `holds()` (`:112–120`), and the debt's source (§III.4).
  `FindsAgainst` becomes `VerifyReads(v) && recordpb.FindsAgainstOf(v.GetOutcome())`; `VerifyReads`'
  comment (`:292–298`) loses "carried by the gap the reading minted".
- **`CitationsFoundAgainst` (`:306–323`) → `counts.sources_refuted`:** reads the standing checks (an
  anchor with a standing check that does not hold) instead of a state; the count's meaning is unchanged
  — citations a standing reading finds against NOW, gap or no gap, information (`verbs.go:306`).
- **`FailingAnchors` (`:325–340`) → `AssemblyFailures(evs, win, gaps, report) []assemblyFailure`** (D11),
  the capture's assembly screen's one read (`capture.go:470–495`); `EvidenceTables.Failed`
  (`evidencetables.go:47–49, 56, 60`) is deleted with its one reader. A row is one of: (i) a computation
  whose ledger state is `stateUnsound` (unchanged); (ii) an OPEN MATERIAL gap of class 32 or 25 whose
  subject the assembled report still carries — about anchor C, C not retired; or quoting a sentence
  that still locates (`gapSentenceIn`); (iii) a standing `reading_debt` (`readingDebtsOf` — a found-wrong
  reading with no gap at all, red's unfinished promise). The FAIL detail names each row by its why:
  "citation C-… (url): open material claim_contradicts_source gap G-… by X" / "sentence '…': open
  material … gap G-…" / "citation C-…: X's reading at event N finds it against and no gap stands on it".
  A refuted citation whose gap was graded low, or closed, is in no row. The PASS detail's "ever found
  against" count (`:511–524`, `EvidenceVerificationJSON.FindsAgainst`) is unchanged through the predicate.
- `passblockers.go`: the `CheckFoundAgainst` arms in `failedFact` (`:229–231`), `failedClears`
  (`:240–242`), `failedRemedy` (`:255–256`), `failedOwed` (`:261–272` → two arms), `blueFailedItem`
  (`:276–278`), `failedCheckRefusal` (`:302–308`, `:317–321`), `blockerKinds`' `failed_check` row
  (`:403–420`), `Blocker.owed` (`:120`); `dispatch.go:1392–1405` (`failedCheckReadiness`); the
  `reading_debt` row arrives (§III.4).
- `retirement.go:128–175` (`lensOwesOf`): the comment "A found-against check never holds its reader" goes;
  the fold is unchanged in code (it names every `owed()` blocker by owner). `sitting.go:426–434`'s comment
  ("A FOUND-AGAINST CHECK IS NEVER ITS DEBT") is rewritten to the two owed kinds.
- **`requireNotARepeatReading` [DELETE the reads arm] (D8):** `failedcheck.go:313–328` keeps only the
  unread delegation (`v.GetAnchor() != "" && !VerifyReads(v)` → `requireNotARepeatFailedCheck`);
  `repeatReadingRefusal` (`:330–381`) is deleted. The rule existed so a second identical reading would
  not owe a second gap (`:256–266`); a repeated finds-against reading replaces the seat's standing check
  and the SAME debt stands (one per standing reading per seat). A same-sitting repeat is already refused
  — a citation reading by its keyed retry (`seat.KeysOn(c, "anchor")`, `verify.go:155`), a found-source
  reading by `requireNotARetriedReading` (`reading.go:217`, whose `carried_by` column read `:225` and
  "gap X carries it" words `:244–247` go). Bound: the sitting cap (`sittingcap.go:63`).
- **Lens credit:** no change. Retirement (`retirement.go:277–302`) counts fresh material mints; a lens
  that reads `refutes` and mints is productive as today; one that reads and does not mint is barren AND
  owes, so it is not retired and its envelope is refused. The scorecard counts no readings.

### III.6 The reading behind a gap — one Go helper, no SQL link (D6)

- **`record.ReadingBehind(evs, win, gap, report) (*Verify, bool)`** [NEW, in `standingreading.go`]: for a
  gap of a class in `readingMintedClasses` by seat X — about anchor C: `standingReadingOf(X, C)`; with a
  quote: X's latest found-source reading whose stored claim locates once at the gap's sentence. False
  for any other gap or where none stands. The ONE statement of "the reading behind this gap"; every Go
  reader below calls it. No SQL twin: the cited case needs none (the gap's `about_ref` and `minted_by`
  are columns; a SQL reader joins `latest_check` on them), and the found-source case needs the report.
- **`carrying_reading` view [DELETE]** (`views.go:412–420`), the index `verify_carried_by`
  (`schema.go:259–261`), and `demand_reach`'s `'reading'` arm (`views.go:493–495`); `demand_reach`'s
  header comment (`:435–459`) and the gap view's `evidence_met` comment (`:597–611`) lose the reading
  clause. `Reach.About` (`demand.go:30–35`) is always true and is deleted with the two `'reading'`
  wordings in `DemandRefusal` (`:134–139`); `reachedGapsSQL` (`windowindex.go:104`) drops
  `"via" = 'reading'`; `windowindex.go:81, 95–99` comments. **What #1739's `carrying_reading` becomes:
  nothing.** Its four readers (issue comment 4 — `evidence_answered` arms 2/3 through `demand_reach`,
  the reaching retire `reachingRetiresOf`, `answers_retired_event`) read the `'about'` arm, which a
  cited source's 32/25 gap populates by its `about_ref`; the reaching retire of C meets the gap's demand
  through that arm as an unread gap's today; the reader's own backing re-read of C meets it through
  `latest_check` (arm 3, `lc.seat = dr.seat`, the seat being the MINTER — the reader, by the mint
  check). A found-source 32 (`--quote`) reaches no anchor, as today, and its demand is met by blue's cite
  `--answers G` (arm 1) alone.
- **Other readers:** `report/assemble.go:1098–1125` (`readingsCarried`, docket.md's "read for the
  sentence" line, #1694 F12) → `ReadingBehind`, one line per gap that has one; `evidenceview.go:262–266,
  407` (`CarriedBy`) deleted; `verbs.go:302` (`show evidence` help) → "docket.md's `read for the
  sentence` line under the gap names the source; under `independent` the row whose seat and claim are
  the gap's minter and sentence is that reading"; `seatprobe/seatprobe.go:206–211` (`carrier_minted`
  count) → counts Mints of `readingMintedClasses` classes by the seat; `capture.go:483` → §III.5.

### III.7 `migrate` — one translation (contract: record migration)

- **Archived shapes** (epochs 55–64 only; one archive, Eiffel): (1) an in-act reading — Mint m, Anchor a,
  Verify v with `carried_by` = m's gap, `carrier_minted` 1 (6 in Eiffel); (2) a held reading — v with
  `carried_by` naming another act's gap, `carrier_minted` 0 (1: event 798).
- **Translation:** delete `readingActs`, `readingActMembers`, `heldMember`, `writeReadingAct`, `failAct`
  and the `held`/`actMember` plumbing (`migrate/replay.go:127–130, 237–260, 547–561, 902–990`),
  `Replay.WriteBodies`/`readingActOf` (`record/replaywriter.go:130–207`), `writeView.act`
  (`writeview.go:30–34`), `appendBodies`' `act` and held arm (`record.go:674–720`), `appendAct`'s `act`.
  Every member is written singly in stream order under the replay's one writer: the Mint (a quote mint
  of class 32, shape-valid), the Anchor, the Verify with no link fields — the two fields are absent from
  the proto, so the source's `carried_by`/`carrier_minted` columns are dropped at `OldEvent` → body
  translation (one line in the Verify arm, beside `eraentries.go:344–348`'s renames); `remap.idFields`
  loses `Verify.carried_by` (`remap.go:116`). `raisedReading` (`:859–873`) stays (a raise names a reading
  by event id, a different relation).
- **Residue, stated:** every archived 25/32 mint comes forward UNJUDGED by `requireOwnStandingReading`
  — 4 archived 25 mints in 4 runs, Eiffel's 6 class-32 mints — because the check is a live act judgement
  and the replay asks none. Archived found-against checks hold nothing after migration: there is no such
  state; where an archived finds-against reading has a 32/25 gap minted from its sitting on about its
  subject, no debt stands either; where none has (an archived plain `refutes` with no gap), a
  `reading_debt` stands on the migrated record, owned by a retired lens — the record is read, not
  dispatched, so it is a true statement about the archived run and holds nothing live. The held reading
  798 comes forward as a reading of G-1d97d860's sentence that says `refutes` and names nothing;
  `ReadingBehind(G-1d97d860)` finds it by (seat, anchor) if its seat is the gap's minter, else the gap has
  no reading behind it on the docket. No fill (the ruled pattern).
- **Part 0 — the archive [NEW `run-archive/2026-11-04_eiffel-tower-height.tar.gz`]:** the one record that
  holds the shape; layout matches `2026-10-30_mersenne-31-prime-sum.tar.gz` (inputs, records,
  computations — no fetched-source cache). Its own PR, asked before it lands (semantic consent: an
  outward change to the archive set). It re-renders every archive-iterating golden (§III.9 G-arch, 39
  tests) and MUST pass `TestEveryArchivedRunMigrates` at epoch 64 first.

### III.8 Agent-facing text [MODIFY / REGENERATE] (contract: prompts, constitutions, help)

Every sentence below is rewritten to the new model, present tense; generated copies are regenerated.
The census (quotes trimmed) is in §III.9 A1–A13.

| Source | Says today | Becomes |
|---|---|---|
| `help/verify.md:15,19,21,25,27,29` | in-act mint; one-gap holds the reading; found against holds the PASS; estoppel exemption; `--url` mints by outcome; result names `gap` | a reading records; a `refutes`/`absent` on a citation is the ground of a gap you mint about the citation (an act; the flags are `mint`'s page's); a refuting found source is the ground of a gap quoting the sentence; a reading that finds against is YOUR DEBT — it holds the PASS as your unfinished promise, on the chair's list with your name, until a gap of its class stands on the subject or you read the subject again, and your envelope is refused while it stands; from the mint on, the gap's grade is what holds; the result names the event |
| `help/mint.md:9,15,17` | one-gap on sentence/inquiry; budget exemption; "a reading … mints its gap in the same act" | one-gap on a sentence, an inquiry OR a citation (every class; the exception is the gap your own unread/unsound check obliges); the standing-reading rule for 32/25 in one paragraph — a CITED source's defect names the citation (`--about-kind anchor`), the subject your reading is keyed on; a FOUND source's names the sentence (`--quote`); what standing means (your latest reading of that subject since its sentence last changed); the refusal; the budget sentence (§III.3); estoppel asks of every `--quote` mint |
| `help/verdict.md` blocker glosses (`PassBlockerGlosses`), `help/dispatch.md:25`, `help/outcome.md` | the kind list | the `reading_debt` gloss arrives: "a lens's reading that finds a source against its sentence, with no gap of its class standing on it" |
| `help/close.md:13` | "your gap another lens's reading named … found-against check clears…" | unread and unsound only |
| `help/edit.md:13` | "a lens's reading … still mints its gap" | an edit that changes a sentence a found source was read against leaves its lens's debt standing; one that applies red's text verbatim discharges it |
| `help/reproduce.md:23,27` | unsound; "on a failed check" | unchanged — read before/after all the same |
| `seat/verbs.go:302,306,334`; `seat/seat.go:1085`; `cli/setup.go:89` | `carried_by`; `sources_refuted`; budget exemption; reason-is-the-gap's-argument | §III.6; `sources_refuted` = citations a standing reading finds against now, gap or no gap; the exemption sentence (§III.3); reason is the reading's |
| `record/passblockers.go` found-against arms, `dispatch.go:1392–1405`, `reading.go:120,128,132` refusals, `failedcheck.go` refusals, `mintbudget.go:308`, `verdict.go:140–160` | | deleted / unread-unsound only; the debt kind's row (§III.4, §III.5) |
| `record.proto` `SourceOutcome` `means` `:531–542`, `GapClass` 25/32, `AboutKind.ANCHOR` | "mints … in the same act" | §III.1 |
| `scripts/agentgen/src/fragments/lens-audit-and-mint.md:3,6,7,9,10`; `fragments/lens-duties.md:5,6`; `fragments/blue-search-and-cite.md:30–31,39,43`; `agents/red-lens-evidence.md:17,19`; `agents/lead-judge.md:81`; `agents/red-chair.md` (the blocker list); `skills/research-protocol/SKILL.md:14–21` | "records its gap in the same act"; "whose reading it carries"; "found against" as a failed check | the reading records; the lens mints the gap about the citation on its reading, or quoting the sentence for a found source, and owes one or a re-read while its reading finds against — the chair sees the debt on its list and names the lens; the demand is met through the citation the gap is about |
| `skills/research-protocol/scripts/debate.js:883` (authored `ledgerClause`) | "your reading records the gap, and blue answers it with a citation" | "where it refutes the sentence, you mint the gap on that reading — a reading that finds against is yours to answer; where it backs an uncited sentence, you mint the gap that asks for the citation" (acts, no verb path; `TestTheSeatPromptsNameNoVerb`); `:1090–1097` unchanged |
| `tools/internal/terms/terms.json` rows `reading` (`:762–763`), `failed check` (`:1091–1092`); NEW row `reading debt` | in-act mint; third state | reading: "…it holds nothing by itself; a gap a lens mints on it is the finding"; failed check: two states; reading debt: "a lens's standing reading that finds a source against its sentence with no gap of its class on the subject — the lens's unfinished promise, a PASS blocker it clears by the mint or a re-read"; then `go -C scripts run ./vocabdoc` (`docs/vocabulary.md:464,632`) |
| `docs/citations-and-bibliography.md:162–164,174–179,181–188,195–197,209–217`; `docs/seat-command-triggers.md:55,67` | three states; in-act; `carried_by` | two states and the debt; the standing-reading mint; the docket line |
| `report/layout.go:167,184` | "as a lens's `verify` does where its reading mints a gap" | deleted clause |
| `plans/compute-and-show.md` (40 lines, §III.9 row P) | `Verify.carried_by`, carrying reading, `found_against` | one dated note at `:515–523` and `:1572` that #1727 removed the link and the state; the rest is history and stays |
| GENERATED: 12 `agents/*.md` (3 `carried_by` hits each from the inlined manual), `skills/adversarial-audit/SKILL.md`, `skills/research-protocol/SKILL.md`, `help_gen.go`, `annotations_gen.go`, `vocabulary.md`, `debate.js` marked regions (`:105–123`, `:373–451` — no reading word; expected unchanged) | | `agentgen`, `cmdtreegen`, `vocabgen`, `vocabdoc`, `enginegen` |

AFTER regenerating, `verify --help`, `mint --help`, `close --help`, `edit --help`, `verdict --help`,
`dispatch --help` and `show evidence --help` are each read before and after, one command at a time, for
semantic loss.

### III.9 Carrier/consumer census (re-run 2026-11-06 on this tree by three grep sweeps; totals at the end)

Commands (worktree root; `T=plugins/frank-exchange-of-views/tools`):
`grep -rnE 'carried_by|CarriedBy|carrier_minted|CarrierMinted|carrying_reading' -I $T --include=*.go --include=*.sql --include=*.golden --include=*.md | grep -v record.pb.go`;
`grep -rnE 'GapClassesOf|FoundSourceGapClassesOf|ReadingGapClasses|FindsAgainst|ReadingClassField|gap_classes|found_source_gap_classes' --include=*.go --include=*.proto $T`;
`grep -rnE 'AppendReading|errReadingHeld|readingAct|pairedSentenceOf|stateFoundAgainst|CheckFoundAgainst|found_against|carrierTree|repeatReadingRefusal|readingSubject|readingGap|FailingAnchors|CitationsFoundAgainst' --include=*.go $T`;
the T-sweep `grep -rlE 'AppendReading|CarriedBy|CarrierMinted|carried_by|carrier_minted|ReadingGapClasses|GapClassesOf|ReadingClassField|citeReading\(|mustRead\(|readingGap|\.Gap\.ID|stateFoundAgainst|FoundAgainst|found_against|carrying_reading|FailingAnchors' --include=*_test.go $T`;
the phrase sweep `grep -rnE 'same act|carried_by|finds against|found against|found-against' plugins/frank-exchange-of-views/{tools/internal/cli/seat/help,docs,skills/research-protocol/scripts/debate.js} scripts/agentgen/src $T/internal/terms/terms.json`.

**Code (non-test), 30 sites:**

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
| C12 | `record/passblockers.go:28–44, 86–92, 120, 223–325, 403–420, 623–698, 709–730, 764–800`; `record/dispatch.go:528–616, 1387–1405`; `record/verdict.go:140–160`; `record/sitting.go:365, 426–434` | 4, 5 |
| C13 | `record/retirement.go:128–148` (comment) | 5 |
| C14 | `record/evidenceview.go:262–266, 269–283, 407, 564` | 5, 6 |
| C15 | `record/demand.go:16–35, 83, 123–139`; `record/windowindex.go:81, 95–104` | 6 |
| C16 | `recordsql/views.go:412–420, 435–459, 493–495, 597–621`; `recordsql/schema.go:259–261`; `recordsql/evidencetables.go:47–49, 56, 60, 75–84` | 1, 5, 6 |
| C17 | `record/replaywriter.go:130–207` | 7 |
| C18 | `record/migrate/replay.go:127–130, 237–260, 547–561, 859–873 (stays), 902–990`; `migrate/remap.go:116`; `migrate/eraentries.go:331–348` (Verify arm, the drop) | 7 |
| C19 | `cli/lens/verify.go` (§III.2 table) | 2 |
| C20 | `cli/lens/mint.go:130–158, 217–222, 235–266, 576–584` | 3 |
| C21 | `cli/seat/verbs.go:302, 306, 334`; `cli/seat/seat.go:1085`; `cli/setup.go:89` | 8 |
| C22 | `report/assemble.go:1098–1125`; `report/layout.go:167, 184`; `report/citations.go` (unchanged) | 6, 8 |
| C23 | `seatprobe/seatprobe.go:206–211` | 6 |
| C24 | `capture/capture.go:470–524` (`FailingAnchors` → `AssemblyFailures`; the FAIL/PASS details) | 5 |
| C25 | `recordsql/testdata/schema.sql`, `schema.epoch.json`; `requirements.json:3` | 1 |
| C26 | NEW `record/standingreading.go` (`readingMintedClasses`, `standingReadingOf`, `requireOwnStandingReading`, `mintSentenceAt`, `ReadingBehind`, `readingDebtsOf`, `AssemblyFailures`); `recordpb` accessor `FindsAgainstOf`; `Blocker.outcome`, `Blocker.lost`; `blockerGap` widened; the `reading_debt` row in `blockerKinds` | 3, 4, 5, 6 |
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
`onegapplacement_test.go`, `passblockers_test.go` (the kind table's rows: "citation found against" →
`reading_debt`), `reachingretire_test.go`, `recordsql/evidencedemand_test.go`, `reopened_test.go` (7),
`replaywriter_test.go` (3), `unreadcitation_test.go`; capture (1): `capture_test.go` (4); fuzz (11):
§III.10. **Plus 10 cli files outside the sweep that run `verify` with gap flags through `gapFlagArgs`
(`verifyclaim_test.go:211`):** `verifyclaim_test.go:62,64`, `mintbare_test.go:157`,
`verifyconfidence_test.go:28`, `bridgekind_test.go:63`, `deadpath_test.go:625`, `gapanchor_test.go:177`,
`requiredmarker_test.go:120`, `verbs_test.go:117`, `answers_test.go:297`, `verifycopy_test.go:203`. The
31 test `record/inquiry_test.go:427` is untouched.

**Goldens (G), 8 files + the archive set:**

| ID | File | Why |
|---|---|---|
| G1 | `difftest/testdata/manual_runs_every_help_page_live.golden` (28 lines + the kind glosses) | verify, mint, close, edit, verdict, dispatch, show evidence pages |
| G2 | `difftest/testdata/gap_class_enum.golden:29,36` | 25/32 `means` |
| G3 | `difftest/testdata/error_catalogue.golden` (8) | reading refusals go; the standing-reading refusals and the debt's gate paragraph arrive |
| G4 | `record/testdata/plans.golden` (42) | verify INSERTs, `carrying_reading`, budget SELECT, the gap-view columns `blockerGap` reads |
| G5 | `record/testdata/shapes.golden:458` | `carrying_reading` |
| G6 | `recordsql/testdata/schema.sql` (18) + `schema.epoch.json` | columns, view, index, epoch |
| G7 | `migrate/testdata/prose-rewrites.golden` (6) | rewritten refusals |
| G8 | `tests/simulator/testdata/prompt-blue-synthesize.golden:7`; `prompt-red-lens-evidence*.golden`; `prompt-red-chair*.golden` | `debate.js:883`; the chair's blocker list |
| G-arch | the archive-iterating set: 39 tests across `migrate/*_test.go` (`everyarchive`, `everyrender`, `proserewrites`, `readingsforward:405`, `archiveonce`, `argumentresidue`, `remaplists` ×2, `gapclass_internal`, `archivedgaps`, `evidenceanchor`, `gapplace`, `benchfiling`, `benchoccasion` ×2, `latestsitting`, `regraderesult`, `storedwindows`, `eventcache`, `reportcheckpoint`, `unmigratedarchive`, `computationarchive`, `computationresidue` ×9, `correctionarchive`, `dispatchpin`, `cutfields`, `remanddirection`, `reproduceanchor`, `findingarea`, `inquirymethod`, `proserefs`), `record/backingarm_test.go:254`, `devcmd/archiverun/outcomes_test.go:103` | Part 0 adds an archive |

**Agent-facing (A), 13 sources:** A1 `help/verify.md` (7 lines); A2 `help/mint.md` (3); A3 `help/close.md:13`;
A4 `help/edit.md:13`; A5 `record.proto` outcome/class/about-kind `means` (9); A6 `seat/verbs.go:302,306,334`,
`seat/seat.go:1085`, `cli/setup.go:89`; A7 `agentgen/src/fragments/lens-audit-and-mint.md:3,6,7,9,10`,
`lens-duties.md:5,6`, `blue-search-and-cite.md:30–31,39,43`; A8 `agentgen/src/agents/red-lens-evidence.md:17,19`,
`lead-judge.md:81`, `red-chair.md`; A9 `agentgen/src/skills/research-protocol/SKILL.md:14–21`; A10
`debate.js:883` (authored; generated regions hold no reading word — grep `carried_by|same
act|claim_contradicts_source` on the file: 0 hits in `:105–123`, `:373–451`); A11 `terms.json:762–763,
1091–1092` + the new row; A12 `docs/citations-and-bibliography.md` (6 passages),
`docs/seat-command-triggers.md:55,67`; A13 the generated set (12 agents, 2 preloaded skills,
`help_gen.go`, `annotations_gen.go`, `vocabulary.md:464,632`), plus `help/verdict.md`, `help/dispatch.md:25`,
`help/outcome.md` through `PassBlockerGlosses`. Refusal strings (Go, agent-read): `reading.go:120,128,132`,
`failedcheck.go:289–299,336–344,410–423`, `passblockers.go:223–325`, `dispatch.go:1392–1405`,
`verify.go:394,529–546,694–709`, `verdict.go:140–160`.

**Views (V), 3:** V1 `carrying_reading` (deleted); V2 `demand_reach` (reading arm deleted); V3 the gap
view's `evidence_met` comment. **1a′ readers (L), 4:** all through V2's 'about' arm after the change (§III.6).
**Plans (P):** `plans/compute-and-show.md` 40 lines (§III.8).

**Census totals:** 30 code sites (27 modified, 1 new, 2 generated families); 45 T-sweep test files + 10
gap-flag files; 11 fuzz files with 18 tally fields and 2 lifecycle tallies (§III.10); 8 golden files + 39
archive-iterating tests; 13 agent-facing sources + the generated set; 3 views; 4 1a′ readers; 1 plan.

### III.10 Release fuzz — `releasegate/fuzz/` [MODIFY] (CI only, never run locally)

| Site | Today | Becomes |
|---|---|---|
| `reading_test.go` (51 lines: `readingClasses:32`, `readingGapArgs:61`, `checkReading:144–173`, `readCitation:275`, `foundAgainstCleared:399`, `checkCarriedBy:458`, `notTheWorkReading:474`, `readFoundSource:577`, `readAnotherSource:715`, `fixFoundAgainst:774`, `readingTally:831`, `readRepeatedSentence:1101`, `readOverPrescription:1270`, `readNotTheWorkFound:1335`) | in-act mint, held readings, `carried_by` checks, found-against remedy | `verify` alone; then, on a coin, `mint --class … --about-kind anchor --about C` (cited) or `--quote S` (found) and the stored gap's `ReadingBehind` is the reading; the second lens's mint about C refused by the anchor arm; `readOverPrescription` expects the estoppel refusal (D10) and then the `--supersedes` mint |
| `foundagainstseed_test.go` (47: `forcedFoundAgainstChain:111`, `foundAgainstReading:284`, `foundAgainstReread:297`, 6 tests `:411–811`) | forced found-against chain through the failed check | forced chain: reading → the `reading_debt` holds the PASS and the lens's envelope → mint on it → the gap's materiality holds → blue's edit/retire → the demand met through the 'about' arm; the "check holds the PASS whatever gap stands" assertions go |
| `failedcheckstate_test.go:18–100, 143–178, 190–210` | `CheckFoundAgainst` drives | two states only; the found-against test deleted |
| `anothersource_test.go:18–82` | second source held by the first's gap | second found-source mint refused by the one-gap rule on the sentence; the second lens's debt discharged by the first's gap |
| `lifecycleseed_test.go:94–95, 153–154, 314–334` (`ReadingDemandRetired`/`Reread`, `lifecycleRead`); `sourcedemandseed_test.go:56–117` | demand reached through a carried reading | demand reached through the gap's `about_ref`; the retire and the minter's backing re-read each met |
| `fuzz_test.go:1143–1159` (`sourceQuoteRefused`, `FoundSourceGapClassesOf`), `:1351–1374`, `:3333–3383`, `:3707–3775`, `:4781–4799` (`CheckFoundAgainst` arms), `:5861, 5972, 5977`, `:6776`, tally copy-out `:7600–7632`, outcome fields `:6199–6224` | | rewritten per row; the `CheckFoundAgainst` arms deleted |
| `sweeptally_test.go:201–220` fields `FoundAgainstClosesRefused`, `FoundAgainstRepeats`, `FoundAgainstEdits`, `FoundAgainstRetires`, `FoundAgainstClearedByOwnReading`, `FoundAgainstStage`, `ReadingDemandRefused`, `NotTheWorkFlagsRefused`, `FoundSourceBudgetRefused`, `ReadingsMinted`, `ReadingsHeld`, `FoundSourceGaps` and their gates `:1159–1175` | 18 reading-shaped fields | deleted: the eight found-against/held/minted/flag fields. Replaced (new meaning, new name): `reading_mints_refused` (map by class), `reading_mints_landed` (by class and subject form), `second_reading_mints_refused` (the anchor arm), `found_source_mints_estopped`, `reading_demand_met_by_retire`, `reading_demand_met_by_reread` — each with a zero-is-false-green gate; `FoundSourceBudgetRefused` deleted (F6 (b): the owed gap is never budget-refused) |
| new drive | — | per sweep, for each row of `readingMintedClasses`: one mint WITHOUT a standing reading (expects the refusal naming `verify`); one with the wrong outcome; one after blue's edit of the sentence (not standing); one with a standing reading (lands, at the cap too — `reading_mints_landed_at_cap`) |
| new drive (the debt) | — | a lens reads `refutes` and returns its envelope → refused, tallied `envelopes_refused_on_reading_debt`; the chair's `gate --as PASS` → refused naming the lens, `pass_refused_on_reading_debt`; discharged on a coin by its mint, its re-read, another lens's gap of the class, blue's retire, an estopped edit (`edit --accept` of a red fix on the found sentence), or left standing to the run's end (`UNVERIFIED` naming it) → `reading_debts_discharged` (map by discharge, each non-zero); a forced short sitting → `reading_debt_readied_again`; a lost quote by a non-estopped edit → the debt stands and is re-read, `reading_debt_lost_quote_reread` |

### III.11 Required tests (each names how it fails if the change is wrong; CLI tests drive the shipped path — `run(t, "lens", "verify", "--run", …)` / `"mint"` — as a seat does)

- **T1** `verify --anchor C --as refutes --class claim_contradicts_source …` → unknown flag, exit non-zero,
  event count unchanged; one row per deleted flag. *Fails if* a gap flag survives on `verify`.
- **T2** a `refutes` reading of C writes one Verify and nothing else (count +1; no Mint, no Anchor); the
  JSON carries `event`; `PassBlockers` holds NO `failed_check` on C. *Fails if* the reading still mints,
  or a failed check is built from it.
- **T3** `mint --class claim_contradicts_source --about-kind anchor --about C` by X with no reading of C →
  refused naming `verify` and C; nothing written. Then X reads C `supports` → still refused (outcome named);
  X reads C `refutes` → lands, `about_kind anchor`, `about_ref C`, `ReadingBehind` = that reading.
  *Fails if* G-b2ba7d66's shape lands, or the outcome is not checked.
- **T4** standing: X reads C `refutes`; blue edits C's sentence; X's mint is refused ("of a sentence blue
  has since edited"); X reads again `refutes` from a sitting opened after the edit → lands. Also: X's
  `refutes` then X's later `supports` → refused (latest replaces). *Fails if* "standing" is "ever read".
- **T5** another seat Y's `refutes` of C does not admit X's mint. *Fails if* the key drops the seat.
- **T6** 25: X reads C `absent` → `mint --class citation_figure_misattribution --about-kind anchor --about C`
  lands; X reads C on a not-the-work copy → lands AND the unread check's `AnsweredBy` is the gap, the
  mint took no budget, and the ledger's state for C stays `unread` (two predicates, F3 (a)); X's
  `refutes` → refused naming `absent`. *Fails if* either alternative of the OR is missing, the
  not-the-work gap does not answer the unread check, or `finds_against` admits not-the-work.
- **T7** found source: `verify --url U --quote S --as refutes` by X, then `mint --class
  claim_contradicts_source --quote S` lands with X's reading behind it; `--quote S'` (another sentence)
  refused; the repeated-sentence fixture (`readingplace_test`'s) places by the quote as any mint does.
  `mint --class citation_figure_misattribution --quote S` refused naming the citation form. *Fails if* a
  found-source 32 is refused, or a found-source 25 admitted.
- **T8** one-gap anchor arm: X's 32 about C open; Y (with its own `refutes` of C) mints 32 about C →
  `oneGapRefusal` naming X's gap; Y's gap of another class about C, with one of that class already open
  about C → refused (every class); Y's owed unread gap about C (Y's `unreachable` first since an edit)
  lands over X's open gap of the same class. *Fails if* duplicates land, the arm is class-limited, or the
  obliged gap is turned away.
- **T9** with X's `refutes` of C standing and no gap: `gate --as PASS` by the chair IS refused, by a
  `reading_debt` blocker naming X and C, and by no `failed_check`; after X's 32 about C graded low and
  closed `not a defect`, `gate --as PASS` is admitted (no other blocker); `show evidence`
  `counts.sources_refuted` = 1 throughout. *Fails if* a found-against failed check survives, the debt
  does not hold the PASS, the mint does not discharge it, or the ledger read is lost.
- **T10** `close --disposition repaired` of X's 32 about C while X's `refutes` stands: refused by the
  DEMAND (`requireDemandMet`: "a reading of C by X … that backs the sentence"), not by a failed check;
  after X's backing re-read, or blue's `retire --anchor C`, admitted. *Fails if* `requireFailedCheckAnswered`
  still names found against, or the demand no longer reaches C through the 'about' arm.
- **T11** budget (F6 (b)): at X's cap, a 32 about C on X's standing `refutes` is admitted and
  `mintStanding.Answering` counts it; a 32 about D where X's standing reading of D is `supports` is
  refused by the budget; after blue edits C's sentence the 32 about C counts against X. *Fails if* the owed
  gap is budget-refused, an unowed one is exempt, or the exemption outlives the reading's standing.
- **T12** docket.md golden: the gap's "read for the sentence" line names the source via `ReadingBehind`
  (cited: C's url; found: U); `show evidence` row has no `carried_by`. `seatprobe` counts the gap.
- **T13** `TestTheEiffelReadingActsComeForwardAsThreeEvents` on `run-archive/2026-11-04_eiffel-tower-height.tar.gz`
  at epoch 65: zero refusals; for each of events 167, 182, 542, 791, 794, 797 the migrated record holds
  the Mint, the Anchor and the Verify in source order, the Verify's columns holding no `carried_by`;
  798 is a Verify alone; G-b2ba7d66 and the 5 in-act 32 gaps are on the gap view with their closures;
  `PassBlockers` holds no `failed_check` on any citation a reading refuted, and the `reading_debt` set is
  exactly the archived finds-against readings with no 32/25 gap from their sitting on (asserted by
  count and subject). *Fails if* a member is dropped, a refusal appears, or a blocker is rebuilt wrong.
- **T14** a synthetic `[]OldEvent` stream (built in-test as `gapclass_test.go` and `cutfields_test.go`
  build theirs) at epoch 64 holding one in-act act, one held reading and one plain 32 mint → migrates
  with zero refusals to the same shapes. *Fails if* the translation depends on Eiffel's particulars.
- **T15** `TestEveryArchivedRunBehindThisBinaryIsRefusedUnmigrated` re-armed by the bump; `schema.sql`
  golden shows no `carried_by`, no `carrying_reading`, no `verify_carried_by`.
- **T16** `TestEveryReadingMintedClassIsDriven`: for each row of `readingMintedClasses` a refused and an
  admitted mint exist in the suite; deleting a row turns the class's T3/T6 red. *Fails if* a class is
  admitted with no check.
- **T17** G1–G3, G8: the help, enum help and catalogue goldens; `TestTheSeatPromptsNameNoVerb` and
  `TestNoPromptGrowsItsCommandCatalogue` green on the new `debate.js:883` and constitutions.
- **T18** the debt on every surface: X reads C `refutes`; `PassBlockers` holds one `reading_debt`
  {Owner X, Subject C, since N, outcome refutes}; X's work list holds it blocking;
  `EnvelopeRefusal(run, X)` names it; the chair's list holds it with `AnotherSeatClears`; the plan
  readies X ("to mint the gap or read again"); `lensOwesOf[X]` names it; `gate --as PASS` refused on it.
  *Fails if* any surface misses it, or a surface words it from its own fold
  (`TestEveryPassBlockerKindAgreesOnEverySurface`'s new row).
- **T19** discharge, one row each, the blocker gone and X's envelope admitted: X's 32 about C; Y's 32
  about C already open before X read (X's mint then refused by the anchor arm, and nothing owed); a 32
  about C closed BEFORE X read → the debt STANDS and X's mint lands; X's later `supports` of C; blue's
  edit naming C (C unchecked); blue's `retire --anchor C`; found source: X's 32 `--quote S`; X's later
  reading of U for S' `supports`; a found-source `absent` → per F7's ruled option. *Fails if* any
  discharge is missing or a non-discharge clears it.
- **T20** once per cause: X's sitting opened after the reading ends sat through with the debt standing →
  the next plan does not ready X for it, `lensOwesOf` still names it, X is not retired at 3 barren
  sittings, `pass_permitted` stays false and `outcome` names C; X's sitting stopped at the call limit →
  the next plan readies X again. *Fails if* the debt readies forever or an unsat sitting ends it.
- **T21** two lenses, one citation: X and Y each read C `refutes`; two debts; X's 32 discharges both;
  Y's point is a finding. *Fails if* Y's debt survives X's gap or Y's duplicate lands.
- **T22** lost quote (F1): X reads U `refutes` for S; blue's ordinary edit changes S's quoted words →
  the debt stands with `lost` = the edit's event and the item naming the new sentence; X's `mint --quote
  <new S>` refused (no standing reading of it); X reads U for the new sentence `supports` → gone; or
  `refutes` → a fresh debt since the new reading, discharged by X's 32 quoting the new sentence. *Fails
  if* the lost quote discharges the debt, or the item cannot name the edit.
- **T23** estoppel discharge: as T22, but the edit is `edit --accept` of a red gap's fix on S (any red
  seat's, not X's) → the debt is gone with no act of X's; X re-reads the new sentence `refutes` → a
  fresh debt; X's `mint --quote <new S>` → the estoppel refusal naming the prescribing gap; with
  `--supersedes` that gap → lands and discharges. *Fails if* a verbatim red edit leaves the debt, or an
  ordinary edit discharges it, or the fresh debt has no route.
- **T24** found-`absent` debt (F2): X reads U `absent` for S → a `reading_debt`; discharge per F7's
  ruled option; X's later `supports` of U → gone. *Fails if* `absent` on a found source creates no debt.
- **T25** the assembly screen (D11): a captured run with X's 32 about C open and material → FAIL naming C
  and G; the same gap graded low → PASS; closed `not a defect` → PASS; X's `refutes` with no gap →
  FAIL naming X's unfinished promise; a found-source material 32 quoting a sentence the report carries →
  FAIL naming the sentence; a computation judged unsound → FAIL (unchanged). *Fails if* the screen reads
  the reading instead of the gap, or misses the debt, or the computation arm is lost.

## IV. Risk & Mitigation

Grades: likelihood × impact × complexity, each L/M/H.

| # | Risk | L×I×C | Mitigation |
|---|---|---|---|
| K1 | A lens reads `refutes` and never mints: the defect reaches no board (the m32 shape) | M×H×M | The debt holds the PASS and the lens's envelope (T18); readied once per cause (T20); left standing, the run ends UNVERIFIED naming it (T20), as an unminted unread gap does today. Residual: none the ruling did not choose — a lens that grades its gap low and closes it has made the judgement the ruling assigns it (T9, T25) |
| K2 | A lens mints 32 by `--quote S` for a CITED source and is refused as a found-source mint | M×L×L | The `--quote` refusal names the citation form and says why the subject is the citation (§III.3); `mint --help` and `--about-kind` usage say so; fuzz drives both forms (§III.10) |
| K3 | The anchor arm of the one-gap rule turns away an obliged unread/unsound gap | L×H×M | the exemption keyed on `answersOwnFailedCheckAt` (D5); T8's third row |
| K4 | The demand of a found-source 32 is met only by blue's cite: a gap blue cannot answer by retire | L×M×L | Unchanged from today; stated in `mint.md`; the bench's docket route stands |
| K5 | `AssemblyFailures`/`sources_refuted` lost with the state deletion, silencing the assembly screen (a plausible zero) | M×H×L | T25's six rows, each a FAIL or PASS the other rows would not produce; §V check 9 deletes each arm and expects red |
| K6 | The lost-quote item names the wrong edit, or none (`old` containment misses: the claim spans two edits, or the report was re-ingested) | M×L×M | `lost` = 0 and the item names the reading alone; the debt stands and clears by the re-read either way (§III.4 residue; T22); the item is information, the predicate is the reading's standing |
| K7 | Estoppel discharge by ANY red seat's verbatim fix lets a lens's debt vanish on a sibling's prescription it never saw | L×M×L | the ruling's choice (red is one party, `estoppel.go:82–87`); the lens may still re-read and a finds-against re-read creates a fresh debt (T23) |
| K8 | A PR bumping to 65 lands first | M×L×L | take the next number; `schemagen` + `TestEveryArchivedRunBehindThisBinaryIsRefusedUnmigrated` |
| K9 | A half-state: `verify` still minting while `mint` already checks (or the reverse) across the phases | M×M×M | one PR (§V phases 1a–1e), each phase green; the PR opens after 1e |
| K10 | The fuzz (CI only) is the first to drive the two subject forms and the debt at scale; a seat-level defect shows after merge | M×M×M | a `--smoke` universe run reads the record (§V check 11); release-fuzz cadence: a full pass before the next logic PR |
| K11 | Archived 25/32 mints and G-b2ba7d66 come forward unjudged; archived finds-against readings with no gap read as standing debts | L×L×L | stated residue (§III.7); T13 asserts the debt set by count and subject; an archived record is read, not dispatched |
| K12 | `passBlockersOf` gains a `report` input: a caller that passes "" silently builds no found-source debt (a plausible zero) | M×M×L | `readingDebtsOf` refuses an empty report where any found-source finds-against reading exists (a `feov.Conflict`, surfaced by `PassBlockers`); §V check 9 passes "" at one caller and expects red |
| K13 | F7 unruled at Part 1: found-`absent` readings create a debt with no mint route | M×M×L | Part 1 does not open until F7 is ruled; T19/T24's rows are written to the ruled option |

## V. Phases and Verification Plan

**Parts.**

| Part | Content | Ships alone? | Epoch |
|---|---|---|---|
| **0** | `run-archive/2026-11-04_eiffel-tower-height.tar.gz`; regenerate G-arch; nothing else | yes — data only; asked first | none |
| **1** | D1–D11: the record change, `verify` records only, the standing-reading mint check, the one-gap anchor arm, the found-against deletion, the `reading_debt` kind, the budget change, the screen, the helper and view deletions, migrate, every agent-facing text, the fuzz, T1–T25 | **no** — the in-act mint and the mint check are two ways of minting one class; help, constitutions and the terms registry can state only one, and the link's writer (`verify`) and its readers (views, budget, failed check) must change at one epoch | 65 |

Within Part 1, one agent per phase, each leaving the tree green: (1a) proto, generated files, epoch,
`finds_against` predicate, `readingMintedClasses`, `requireOwnStandingReading` + `mintSentenceAt` +
`ReadingBehind` + `standingReadingOf`, the anchor arm — `verify` still minting; (1b) `verify` records
only, the `reading_debt` kind and `AssemblyFailures`, deletions in
`reading.go`/`record.go`/`onegapperclass.go`/`mintbudget.go`/`failedcheck.go`/`passblockers.go`/`dispatch.go`/
views/`demand.go`/`windowindex.go`, readers switched; (1c) migrate + T13/T14; (1d) agent-facing text +
goldens; (1e) fuzz. The PR opens after 1e and after F7 is ruled.

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
   `grep -rnE 'GapClassesOf|FoundSourceGapClassesOf|ReadingGapClasses|ReadingClassField|AppendReading|errReadingHeld|readingAct|pairedSentenceOf|CheckFoundAgainst|stateFoundAgainst|carrierTree|repeatReadingRefusal|readingSubject|FailingAnchors' --include=*.go plugins/frank-exchange-of-views/tools`;
   `grep -rnE 'same act|found-against|found against' plugins/frank-exchange-of-views/tools/internal/cli/seat/help plugins/frank-exchange-of-views/docs plugins/frank-exchange-of-views/skills plugins/frank-exchange-of-views/agents scripts/agentgen/src plugins/frank-exchange-of-views/tools/internal/terms/terms.json | grep -v 'supersed\|channel'` (the two unrelated "same act" senses; "finds against" stays as the option's name).
   *Re-armed by:* any text or code edit.
8. Read each help page before and after, one command at a time: `go -C plugins/frank-exchange-of-views/tools run ./cmd/feov-record lens verify --help` (then `lens mint`, `lens close`, `blue edit`, `verdict`, `dispatch`, `show evidence`). *Re-armed by:* any help source.
9. Delete-the-row checks by hand, each reverted: remove 32's row from `readingMintedClasses` → T3 red; remove 25's not-the-work alternative → T6 red; remove the anchor arm's own-failed-check exemption → T8 row 3 red; remove `(finds_against) = true` from `ABSENT` → T6, T24 and the ledger's `absent` tests red; delete the `reading_debt` row from `blockerKinds` → T18 red (`TestEveryPassBlockerKindAgreesOnEverySurface`); delete the "or minted from the reading's sitting on" clause → T9's closed-low row red; delete the "by ANY seat" → T19 row 2 and T21 red; delete the `applied_verbatim` discharge → T23 red; delete the `lost` edit lookup → T22's item row red; delete `AssemblyFailures`' material-gap arm → T25 rows 1–3 red; delete its debt arm → T25 row 4 red; delete its computation arm → T25 row 6 red; pass `""` as `report` at one `passBlockersOf` caller → K12's refusal test red; delete the budget join on `standingReadingOf` → T11 red; delete the Verify-column drop in migrate → T13 red. A row that stays green is a finding. *Re-armed by:* any change to the table, the arm, the option, the translation.
10. `go -C scripts run ./check 2>&1 | tee ~/.claude/scratch/mint-prereq/plan2/check.log | tail -40` → read the fail count in the log. *Re-armed by:* every push.
11. `scripts/universe.sh build`, then `grep -rn 'standing reading' <the universe's installed plugin help>` proves the universe carries the change; one `--smoke` run (one lane, haiku both tiers); read its record: the `verify` table has no gap column; every 32/25 mint has `ReadingBehind`; no plain 32 mint by a seat with no reading; every finds-against reading has a 32/25 gap from its sitting on or a later non-finds-against reading, or the outcome names it. *Re-armed by:* Part 1's merge candidate.
12. **CI only — release fuzz** (`FEOV_RELEASE_GATE=1`, the sharded job in `.github/workflows/hooks.yml`, scheduled and on tag; never run locally by any door): `reading_mints_refused` non-zero for both classes, `reading_mints_landed` non-zero for both subject forms and at the cap, `second_reading_mints_refused`, `found_source_mints_estopped`, `reading_demand_met_by_retire`, `reading_demand_met_by_reread`, `envelopes_refused_on_reading_debt`, `pass_refused_on_reading_debt`, every key of `reading_debts_discharged`, `reading_debt_readied_again`, `reading_debt_lost_quote_reread` non-zero. A logic PR waits for one full pass (cadence ruling 2026-10-29). *Re-armed by:* Part 1's merge candidate.

## VI. Forks

**RULED (gblock, 2026-11-06; option wording the lead agent's):**

| Fork | Question | Ruling | Home |
|---|---|---|---|
| F1 | a found-source reading whose quote stops locating after blue's edit | **Debt stands** (not the recommended "owes nothing"); refined: unless the edit took red's suggested text verbatim, then estoppel discharges it. The mint half of (a) stands: no `--quote` mint until the lens reads the new sentence | §III.4 lost quote; T22, T23 |
| F2 | `absent` on a FOUND source | **Debt like refutes** (not the recommended "no debt"); the discharging class is F7 | §III.4; T24 |
| F3 | not-the-work as `finds_against` for 25 | **(a) two predicates**: the ledger's `finds_against` is refutes/absent on a copy that read the work; 25's mint check also admits the seat's standing not-the-work reading | §III.3 table; T6 |
| F4 | estoppel on a found-source 32 | **(a) no exemption**; reverses the 2026-10-31 exemption for this case | D10; §III.3 |
| F5 | the one-gap anchor arm's scope | **(a) every class** about an anchor, except the gap a seat owes for its own unread/unsound check; a second lens's debt is discharged by any seat's gap of the class about the citation | D5; T8, T21 |
| F6 | the owed 32/25 gap and the mint budget | **(b) outside the budget while open**, as the owed unread/unsound gap is | D7; T11 |
| §VII 1′ (first draft) | the debt's home: a `BlockerKind` or a parallel fold | **a `BlockerKind`** on the one list; the chair sees it, the lens clears it | D4′; §III.4 |
| §VII 1 (first draft) | the assembly screen on a refuted citation | **screens the gap**: open material 32/25, or a found-wrong reading with no gap | D11; T25 |

**OPEN — none decided here (the rulings created them):**

**F7 — which gap discharges the debt a found source read `absent` creates (from F2).** No shipped class
names "a page the lens found is silent on the sentence": 25 is a CITED source's defect and its `--quote`
form is refused; 32 is "states otherwise", not silence; 29 `unsourced_claim` is the REPORT's want of a
citation, true or false of the sentence whatever the found page says.
- (a) **29 discharges it where the sentence is uncited; where the sentence carries a citation the debt
  is not created.** A found page silent on an UNCITED sentence is the lens's failed search for backing,
  which is what 29's gap carries to blue ("the gap stands until blue cites one"); a found page silent
  on a CITED sentence says nothing against the citation C. Cost: the debt's existence depends on another
  anchor's presence in the sentence (one more input to the fold: the sentence's bare anchor ids,
  `claimcount.BareAnchorIDs`, already read by `gapSentenceIn`); 29 is a judgement class forced by a
  debt; a lens that found three silent pages for one uncited sentence owes one 29 (the one-gap rule).
- (b) **The debt is created and clears only by the re-read** (the lens's later reading of U, or of any
  source for the sentence). Cost: a debt with no gap route — at an honest `absent` the lens must read
  something else to clear it, which is an act for the record's sake; the fork's wording "like refutes"
  is not met (refutes has a mint route).
- (c) **No debt** (F2's recommended (a), re-proposed): the reading records; a found page's silence is
  information. Cost: reverses the 2026-11-06 ruling on F2; the lens's `absent` reading reaches no board
  unless the lens mints 29 of its own accord.
- **Recommendation: (a).** It implements the ruling with a class that already carries the found page's
  meaning to blue, and states the one case (a cited sentence) where silence is no defect. Not decided
  here; Part 1 does not open on an unruled F7 (K13).

**F8 — the found-source debt when the sentence is deleted outright (from F1's residue).** Blue's edit
removes the sentence the claim quoted and puts no sentence in its place (`new` empty, or `new` with no
sentence the claim could be read against).
- (a) **The deletion discharges the debt.** The sentence the source was read against is gone; nothing is
  left to mint about or to read against; the reading stays as information (`show evidence`
  `independent`). Cost: a deletion is a route around a debt — blue deletes a refuted sentence and the
  lens's promise is void; but the defect left the report with the sentence, which is the remedy the
  gap would have demanded.
- (b) **The debt stands until the lens records a reading of U for another sentence**, or the run ends
  naming it. Cost: a hollow act (a reading against a sentence the source was never found for) or an
  UNVERIFIED end over a sentence the report no longer carries.
- **Recommendation: (a).** A deletion is blue's edit naming the sentence, and the cited-source row's
  rule already treats blue's edit of the sentence as a discharge; (a) makes the found-source row agree
  where the edit leaves no successor sentence. Not decided here.

## VII. Inconsistencies between the ruling and the code, reported, not resolved

1. "Keyed on the subject the mint already names" — for a cited source that subject is the citation
   (`--about-kind anchor`), not the sentence, or the gap reaches no anchor after `carrying_reading` goes
   (`demand_reach` `views.go:493–495` is the only reach a quote gap had) and its `source` demand can be
   met only by a new cite. Kept as designed (D3); `ABOUT_KIND_ANCHOR`'s `means` (`record.proto:1179`),
   `mint --about-kind`'s usage (`mint.go:220–221`) and `mint.md` say so (§III.1, §III.3, §III.8).
2. The class-to-outcome rule has no declared home (the option mechanism was dropped), so it is a Go
   table in `record` (`readingMintedClasses`, §III.3) beside 31's hand check at `record.go:1931` — two
   hand checks of two kinds in two files. Kept: 31's is a shape check the replay runs; this one is a
   live judgement.
3. The ruling's "enforced by its own envelope in the same sitting" and "an orphaned reading blocks the
   PASS" are ONE mechanism in the code, not two: the lens work list `add`s every owed blocker the lens
   owns (`sitting.go:430–434`) and the envelope gate reads `Blocks`, so a `reading_debt` refuses the
   envelope from the moment the reading lands — there is no "same sitting" window in which the debt
   holds the envelope but not the PASS. The chair's `gate --as PASS` in a sitting concurrent with the
   lens's is refused on it too. This is what every owed kind does today and is taken as the ruling's
   intent; stated so nobody reads "orphaned" as a state the code tracks.
4. Ruling 8 says a second lens's debt is discharged by any seat's OPEN gap; ruling 1 says the MINT
   discharges and ruling 4 says a closed low gap does not fail the run. Reconciled in §III.4 as "open
   now, or minted from the reading's sitting on" — a predicate, not a stored fact; the one case the
   rulings do not reach (a gap of the class closed BEFORE the lens read) is decided there as not
   discharging, so a fresh reading after an adjudication owes a fresh gap.
5. `passBlockersOf` is stream + SQL inputs (`gaps`, `fresh`); the found-source debt needs the rendered
   report (`RenderAt`), so the fold gains a `report` input at every caller (`passBlockerReadsAt`,
   `sitting.go:365, 430`) — the first blocker arm to read report text. K12 states the plausible zero
   and its guard.
6. `migrate`: an archived plain finds-against reading with no gap from its sitting on reads as a
   standing `reading_debt` owned by a retired lens on a migrated record (§III.7). The record is not
   dispatched, so nothing acts on it; it is a true statement about the archived run. Stated, not filled.
