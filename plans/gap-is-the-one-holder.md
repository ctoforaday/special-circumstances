# A gap is the one PASS-holder — `verify` records, `mint` of 32/25 stands on a reading (#1727)

> **Status: DRAFT, revised 2026-11-06 after audit round 3 (one gap, nine notes; gblock's rulings of the
> same day folded, three relayed by the lead agent).** Written on `plan/mint-prerequisites` merged with
> `origin/main` at `fce8d6422` (EventSchema **64**, `recordpb/schema_gen.go:18`; every line reference
> below is re-pinned on that merge). Replaces `plans/mint-prerequisites.md` (renamed with `git mv`; its
> mechanism — a `MintPrerequisite` option, a `rests_on` link, a generic check — was dropped by gblock
> 2026-11-06). The scope is the ruling quoted in §I; option wording throughout is the lead agent's, not
> gblock's. Every fork is RULED (§VI). Every census row in §III.9 was re-run by grep on this tree, and
> §III.7's shapes and T13's expected set are read off the Eiffel archive's rows (sqlite, 2026-11-06).

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
it (F1); a found-source `absent` creates the same debt (F2) where the sentence carries no citation, and
none where it does; an `unsourced_claim` gap by any seat on the sentence discharges it (F7 (a)); deleting
the sentence outright discharges a found-source debt (F8 (a)); two predicates for not-the-work (F3 (a));
one open gap per class per anchor for every class (F5 (a)).

**Ruled 2026-11-06, relayed by the lead agent (not in #1727's comments; attributed to gblock through the
lead):** `counts.sources_refuted` is DROPPED with the ledger state it counted — its only consumer would be
one help line (field-destination rule); the lost-quote edit lookup (K6) IS BUILT, since the estoppel
exception requires finding the edit; the as-of join for a lost reading IS BUILT, since without it the
migrated Eiffel record carries two phantom debts, which the migrate ruling forbids.

**Why (the value case against the naive version).** "Delete `carried_by`, let `mint` take the class as
before" leaves G-b2ba7d66's shape — a plain `claim_contradicts_source` mint with no reading, its evidence
lost to prose in `mint_reason` — fully admitted, the defect #1727 was filed on. The previous draft's link
field doubled the concept and needed budget, estoppel, multiplicity and successor exemptions to stay
consistent. This version keeps ONE holder (the gap) and ONE observation (the reading) and joins them on
fields both already carry — the seat, the subject, the class — a question of the record, not a field a
writer must keep in step.

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
   `pass_permitted` false; a gap of the class the reading obliges on its subject — 32 or 25 by the lens,
   or for a found source read `absent` an `unsourced_claim` by any seat — discharges it, after which the
   gap's materiality alone holds the PASS. A found source's subject is the SENTENCE its claim locates
   to: one lens reading one url against two sentences owes two debts and may mint two gaps. A found
   source read `absent` against a sentence that carries a citation creates no debt (T18–T24). The record's
   derived outcome reason and the engine's termination reason name a standing debt with ONE clause, read
   from one generated table (T26).
5. The reading-minted exemption (`carrier_minted`) is gone from the budget; every open 32/25 gap is
   outside the budget WHILE OPEN — each stands on its minter's reading by the mint check — as the owed
   unread/unsound gap is; so is the 29 a found-source `absent` reading of the minter's obliges (T11).
6. `feov-migrate` brings the Eiffel archive forward with zero refusals: each of the four archived in-act
   readings of a CITATION comes forward as its Verify, its Anchor and a Mint ABOUT that citation; the two
   in-act readings of a FOUND source come forward as the quote mints they already are; the held reading
   is a Verify alone; no `reading_debt` stands on the migrated record (T13, T14).
7. The capture's assembly screen FAILS a run with an open MATERIAL 32/25 gap whose subject the
   assembled report still carries, or with a standing `reading_debt`; a refuted citation whose gap was
   graded low and closed passes it (T25).
8. The release fuzz drives the mint refusal, the admitted mint for both classes and both subject forms,
   the debt's PASS refusal and each discharge, with no zero tallies (CI only, §V check 12).
9. `counts.sources_refuted` is gone from `show evidence` (JSON, help, goldens, tests); nothing counts a
   found-against reading (T9's JSON row).

**Out of scope.** Any change to a class's meaning; prerequisites for 2, 23, 27, 31 (ruled plain mint;
31's hand check `record.go:1931` stays as it is); #1695 (findings fold into gaps).

## II. Technical Context

**Record shapes today (`origin/main` merged at `fce8d6422`, epoch 64).** Only what the design turns on;
the rest is the census (§III.9).

- The link: `Verify.carried_by` (21) / `carrier_minted` (22), `record.proto:1431–1444`, written by
  `AppendReading` (`reading.go:92–103`) and the held arm (`record.go:705–716`), read by the
  `carrying_reading` view (`views.go:412–420`) → `demand_reach`'s `'reading'` arm (`:493–495`), and by
  `mintbudget.go:330`, `reading.go:225`, `evidenceview.go:436`, `assemble.go:1103–1120`,
  `seatprobe.go:206–210`.
- The outcome options `gap_classes` (50014) / `found_source_gap_classes` (50015), `record.proto:152–165`,
  set at `:531–541`; readers `GapClassesOf`/`FoundSourceGapClassesOf`, `record.FindsAgainst`
  (`reading.go:34`), `verifyHolds` (`evidencetables.go:75–84`).
- The ledger's third failed state `stateFoundAgainst` (`anchorchecks.go:146–148`) → `CheckFoundAgainst`
  (`failedcheck.go:57`): blocker from `v.GetCarriedBy()` (`:112–123, 177`), `owed()` false (`:89`), close
  refusal `:395–427`. Readers of the state beyond the blocker: `CitationsFoundAgainst` (`anchorchecks.go:315`
  → `counts.sources_refuted`, `evidenceview.go:324–329, 593`), `FailingAnchors` (`:325–340`, through
  `EvidenceTable.Failed` = `"found_against"`, `evidencetables.go:48–55`) → the assembly screen
  (`capture.go:473–523`). Every other `anchorChecksOf().state()` reader asks `unchecked`/`unread`/`unsound`
  only: `evidenceview.go:589` (`sources_unverified`), `unchecked.go:47`, `computationverdict.go:224`,
  `citationid.go:324` (a computation's verdict), `failedcheck.go:147`. No `show evidence` row carries a
  ledger state word (`EvidenceSourceJSON`, `evidenceview.go:65–124`, checked).
- The one list: `passBlockersOf` (`passblockers.go:692–714`) = the gap arms + `owedBlockersOf`
  (`closingowed.go:209`: `eventBlockersOf` `passblockers.go:732` + closings + inquiries) + grade movement +
  the lens gate, which reads `lensOwesOf(…, events)` (`retirement.go:148`) over the SAME `events` slice.
  `owedBlockersOf` is also called directly, with no `found`, at `dispatch.go:311` (the plan's lens folds)
  and `retirement.go:445` (`passLensGateOfRun`), and in tests (`failedcheck_test.go:283`,
  `anchorowner_test.go:70,125`, `inquiry_test.go:351`). `passBlockersOf`'s direct callers:
  `sitting.go:354, 371, 435, 500` (through `sittingOf`'s `sittingReads`, `:124`; `SittingOf` `:118` builds a
  zero `sittingReads{}` and is called from tests only, `available_test.go:33`, `citejoin_test.go:90`),
  `dispatch.go:542`, `passBlockersOfReads` (`:833`). Its SQL inputs are read at three sites inside one
  snapshot: `passBlockerReadsAt` (`:816`, for `PassBlockers` `:784`, `passBlockersAt` `:803`,
  `passLensGateOfRun` `retirement.go:431`), `planReadsAt` (`dispatch.go:264`), `workView`
  (`viewjson.go:1459` → `sittingReads` `:1549` → `sittingOf` `:1558`). `gaps []blockerGap` is the OPEN set
  only (`blockerGapsOfStates` drops `!g.Open`, `:664–674`; `blockerGapsOfOpen` `:679`).
- The kinds: `BlockerKind` (`passblockers.go:29–44`), the `blockerKinds` table (`:348+`; `failed_check`
  row `:413–429`), `BlockerKinds()` (`:772`) → the plan schema's `blockers.kind` enum through
  `planfields.go:58` → `enginegen` → `debate.js:415`. `verdict.go:149` `anchorsStanding` names
  `unchecked`, `failed_check`, `open_inquiry`, `unchecked_inquiry` in the outcome reason; `debate.js:1094–1101`
  (authored) names `unchecked` and `failed_check` ONLY, with its own clause strings — the two carriers
  already disagree on the inquiry kinds; `debate.test.mjs:264–275` pins the JS clause against its fixture.
- The report text is SQL: `RenderAt(q)` (`record/queries.go:282`) → `reportSourceAt(q, at)` (`:259`) →
  `recordsql.ReportSourceAt` (`recordsql/reportcheckpoint.go:61`), the newest checkpoint at or below `at`
  and the ops after it, a checkpoint every `reportCheckpointEvery` = 8 ops (`reportcheckpoint.go:20`).
  The work path already renders once per snapshot through `reportRender` (`viewjson.go:1113–1129`).
- Estoppel: `appliedVerbatim` (`estoppel.go:55–77`), `EstoppelConflict` (`:90`, visible-text containment
  either way, `:86–100`); the guard is `mint`'s on `--quote` (`mint.go:155–158`, `:576`), not `verify`'s
  (`reading.go:89–91`). `BlueEdit` carries `old`, `new`, `answers`, `applied_verbatim`
  (`record.proto:1639–1649`).
- `satThroughOf` (`dispatch.go:910–951`) stores each sat-through sitting's OPENING (`:944`), so
  `firstAfter(ends.through[owner], b.since)` (`:615`) is "a sat-through sitting that OPENED after the
  blocker arose". `AppendPlaced` (`record.go:658–664`) writes an Anchor only where the act names an id AND
  a quote; an Anchor is a report op (`reportcheckpoint_test.go:65`). `failedChecksOf` reads every Mint
  about an anchor off `evs` into a `mints` map (`failedcheck.go:95–106`) — the shape the debt's
  closed-since lookup reuses.

**Measured inputs (re-checked where cheap).**

| Measure | Value | How |
|---|---|---|
| Archived mints of 25/32 after `migrate`'s slug translation (`migrate/gapclass.go:43,47,56` → 25) | 25: 4 mints in 4 runs; 32: 6 in 1 run (Eiffel: 5 in-act, 1 plain G-b2ba7d66) | python over each `run-archive/*.tar.gz` `records/record.db` (2026-11-05) |
| Archives under `run-archive/` holding a Verify with `carried_by` | 0 of 28 SQLite archives | same scan |
| The Eiffel record | `~/.claude/scratch/contradiction-impl/elicit-adversary/2026-11-04_eiffel-tower-height.tar.gz`, 370 KB, epoch 56; layout `inputs/run-config.json`, `records/record.db`, `computations/<sha>/{acts,dir,rerun}`; 59 readings, 7 with `carried_by` — 6 in-act, of which FOUR are of a citation (`verify.anchor` set: 167 `absent` C-2dc96eab → G-8dbfce13 (25); 791, 794, 797 `refutes` C-a51c49f1, C-a1560eb7, C-486ed6b5 → G-f66af91d, G-dd7b403e, G-de476176 (32)) and TWO of a found source (`verify.url` set, no anchor: 182 → G-3b2b7d93, 542 → G-1d97d860, both 32); 1 held (798, `refutes` C-3094c824, `carried_by` G-1d97d860, `carrier_minted` 0). Every in-act Mint has `location` = the Verify's `claim` and NULL `about_kind`; the four cited ones' Anchor rows are `sentence_of` = C, the two found ones' are `location` = claim | `tar tzf`; sqlite read (2026-11-06, `~/.claude/scratch/mint-prereq/plan2-rev2/eiffel/`) |
| The seven readings' later history (for T13) | C-2dc96eab retired at 223; the four cited `refutes` each followed by the same lens's `supports` of the same anchor in sitting 6 (952, 953, 954, 955) and by a blue edit naming the anchor (835, 858, 848, 846); 182's claim lost by edit 238 (`old` ⊇ the claim's visible text, `answers` G-3b2b7d93, not verbatim); 542's claim lost by edit 692 (`old` ⊂ the claim, `answers` G-1d97d860, not verbatim); each in-act gap closed `repaired` (373, 374, 956–959); the lens's sittings opened at 97, 315, 471, 601, 747, 914, 1015, 1144 | same read |
| That record under main's `feov-migrate` at epoch 61 | 1213 events in → 1203 out, 0 refusals | 2026-11-05; re-run at 65 is T13 |
| Epoch | `origin/main` 64 at `fce8d6422` (merged 2026-11-06, after `bdcc411c4`); this plan takes **65** | `recordpb/schema_gen.go:18`, `requirements.json` |
| Names coined here, grepped free on the Go tree, `debate.js`, `agents/`, `skills/` | `reading_debt`, `BlockerReadingDebt`, `readingDebtsOf`, `standingReadingOf`, `readingMintedClasses`, `AssemblyFailures`, `finds_against`, `locatedReading`, `locatedReadingsAt`, `anchor.BackingMark`, `StandingClauses`, `STANDING_CLAUSES` | 0 hits outside `plans/` (`foundReading` is taken — a Verify builder in `foundsource_test.go:33`, rewritten in Part 1 — so the fold's input is `locatedReading`) |

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
| D3 | `mint` of 32/25 is a live act judgement, `requireOwnStandingReading`, keyed on the mint's subject: a CITED source's defect names the citation (`--about-kind anchor --about C`), a FOUND source's defect names the sentence (`--quote S`); the seat's standing reading of that subject — for a found source, the seat's latest found-source reading whose claim locates to S's sentence, whatever its url — must carry the class's outcome; 25 ALSO admits the seat's standing not-the-work reading of C (F3 (a), RULED) | §III.3 |
| D4 | `stateFoundAgainst` leaves the ledger: a citation with a `refutes`/`absent` reading is holding or unchecked by the ordinary rules; `CheckFoundAgainst`, its blocker, its close refusal and its carriers go. The per-check `holds` stays (the demand's "check again that holds") | §III.5 |
| D4′ | The reading debt is a `BlockerKind`, `reading_debt`, built in `eventBlockersOf` — the stream arms every owed-blocker reader folds (`owedBlockersOf`), so retirement and the lens gate see it — from the ledger's standing checks that do not hold (cited) and the seat's latest found-source reading PER SENTENCE (found): one per (lens, subject), owned by the lens, owed while no gap of the class stands on the subject (for a found source read `absent`: an uncited sentence only, and the class is 29 by any seat — F7 (a)), discharged by the mint, the re-read of the subject, blue's retire, blue's cite or deletion of the sentence (F8 (a)) or an estopped edit; it refuses the lens's envelope, readies the lens once per cause, is on the chair's list and holds the PASS. One list (#1202, RULED) | §III.4 |
| D5 | The one-gap rule gains an anchor arm: one open gap per class about one anchor, for EVERY class, except the gap a seat's own owed unread/unsound check obliges (F5 (a), RULED) | §III.3 |
| D6 | A gap carries nothing of its reading by field; the reading behind a gap is one Go helper, `record.ReadingBehind`, joining (seat, subject, class); `carrying_reading` and `demand_reach`'s reading arm are deleted | §III.6 |
| D7 | The `carrier_minted` budget exemption is deleted; every open gap of a `readingMintedClasses` class is outside the budget WHILE OPEN (F6 (b), RULED) — by class, no join: the mint check made each stand on its minter's reading, and blue's later edit does not bring it back inside; the owed 29 (an open 29 by the seat quoting a sentence the seat's earlier found-source reading read `absent`) likewise; a 25 gap about a citation whose unread check the minter owes is exempt by the existing own-failed-check rule | §III.3 |
| D8 | `requireNotARepeatReading`'s reads arm is deleted; the unread arm stays | §III.5 |
| D9 | `migrate`: the in-act reading act machinery is deleted; each archived member comes forward as its own event with the two link fields dropped, and ONE translation: an in-act Mint whose carrying Verify names a citation becomes a Mint ABOUT that citation (`about_kind` anchor, `about_ref` = the Verify's `anchor`, `location` dropped) — so no reading debt is read off the migrated record; stated residue | §III.7 |
| D10 | Estoppel: no exemption for a found-source 32 — the guard asks of every `--quote` mint as today, and `verify` attaches none (F4 (a), RULED; reverses the 2026-10-31 exemption for this case) | §III.3 |
| D11 | The capture's assembly screen screens the GAP: FAIL on an open MATERIAL 32/25 gap whose subject the assembled report still carries, or on a standing `reading_debt`; `counts.sources_refuted` and `CitationsFoundAgainst` are DELETED with the state (RULED: no destination but one help line) | §III.5 |
| D12 | The outcome reason's clause naming the kinds that hold a PASS by their own state is ONE Go table, `StandingClauses`, that `verdict.go`'s `anchorsStanding` folds and `enginegen` renders into `debate.js`'s generated region as `STANDING_CLAUSES`; the authored `anchorsStanding` folds the constant. `reading_debt` is a row | §III.8 |

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

- **`Verify` [DELETE fields]:** `carried_by` (21), `carrier_minted` (22) and their comment `:1431–1444`;
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
  quote". 29 (`:472`) says "a source a lens read that backs the sentence, or says nothing of it, is not
  the report's citation". No option, no value change. The class-to-outcome rule lives in Go (§III.3), as
  31's does.
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
| `writeVerify` `:367–399` (`classes`, `backing`, the gap-flag refusals, `gf.fill`, the class check) | the reading is written alone; `markedSentenceOf` `:373` stays (a backing found-source reading of a marked sentence is refused — `refuseMarked` `:591–616` loses "mints unsourced_claim" and says a found source backs only an unmarked sentence); the guard keeps its `backing` key (`:369`, `!FindsAgainst`): an `absent` found-source reading of a marked sentence is recorded, and under F7 (a) owes nothing |
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
| 32 | `--quote S` (a found source) | the seat's LATEST `verify --url` (any url) whose `claim` locates once in the landing report inside S's sentence, reads the work, outcome `refutes` — the subject is the sentence, as the ruling says, so one url read against two sentences admits two mints, and the seat's later reading of another source for the same sentence replaces it |
| 25 `citation_figure_misattribution` | `--about-kind anchor --about C` | of citation C: reads the work with outcome `absent`, OR `source_completeness` `not_the_work` (F3 (a), RULED: two predicates — the ledger's `finds_against` is refutes/absent on a copy that READ the work; this row's second alternative is the mint check's alone, and a not-the-work copy stays `unread` in the ledger) |
| 25 | `--quote S` | refused: the class names a cited source. A found source read `absent` for an uncited sentence obliges 29 `unsourced_claim` (F7 (a), RULED) — a judgement class, plain mint, NOT a row of this table: any seat mints it with no reading behind it, as today |

**"Standing"** = the seat's LATEST reading of that subject since the subject's sentence last changed:
for a citation, the ledger's own rule (`anchorStanding.sinceEdit`, anchor not retired — read through
`anchorChecksOf`, which the check reuses rather than restating; the ledger's `stands` additionally
requires `reads`, so 25's not-the-work alternative reads the seat's latest check since the edit
directly, `reads` or not); for a found source, the seat's latest found-source reading — ANY url — whose
stored claim locates once in the landing report inside the subject sentence (`readingSentenceIn`,
`reading.go:185`, stays for this); the url is no key (the ruling: "the sentence for a found source"). A
reading whose claim no longer locates stands for no MINT: a mint of 32 after blue's edit is refused until
the seat reads the sentence as it now stands — the ruling's "standing" clause. (What the seat still OWES
on such a reading is §III.4's lost-quote rule.) `standingReadingOf(seat, subject)` — subject an anchor id
or a sentence span of the landing report — is the one statement of this, shared by the mint check,
`ReadingBehind`, the budget's 29 join and the debt fold.

**Where it runs.** `requireOwnStandingReading(w writeView, seatID string, m *recordpb.Mint) error`, called
in `judgeAct`'s Mint arm (`actjudge.go:40`) AFTER `requireAboutAnchorLive` (a 32 about a retired C is
refused as about nothing, not told to verify it) and BEFORE the one-gap rule and the budget: a mint
refused for its ground is not first told about a held subject or a full budget. A live act judgement, never the
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

**The budget [MODIFY `mintbudget.go`] (D7, F6 (b) RULED — "outside the budget while open"):** delete
the `w.act.pairs(m)` arm (`:252–259`) and the `carried_by` query in `ownAnswersAt` (`:325–345`).
`ownAnswersAt` keeps `ownFailedCheckAnswers` and `ownsInquiryGap` and gains two arms: (i) every OPEN gap
of a `readingMintedClasses` class by seatID — by class alone, no reading join: the mint check admitted it
only on seatID's standing reading, so "owed" is what the class means, and the exemption lasts exactly
while the gap is open, as the ruling says — blue's later edit of the sentence leaves the gap open and
outside; (ii) every open 29 by seatID quoting a sentence that a found-source reading of seatID's,
recorded before the mint, read `absent` for (the gap F7 (a) obliges; `ReadingBehind`'s join, not
standing — the obligation was the mint's, and the exemption is the gap's while open). `mintStanding.Answering`
counts both. The header comment `:32–44` and `mintAnswerOutsideBudget` (`:308`), `verbs.go:334`,
`setup.go:89`, `mint.md:15` say one sentence: "a gap about an anchor whose check of yours read nothing or
judged a computation unsound, a gap of a class your reading defines (claim_contradicts_source,
citation_figure_misattribution), or the unsourced_claim gap a source you found silent on an uncited
sentence obliges, takes no mint while open; so does the logic lens's gap about an inquiry". Cost, stated:
a lens at its cap can still open 32/25 gaps, one per refuted subject (the one-gap rule bounds it), and
each stays outside the budget until it closes. Nothing is exempt for having been minted by a reading.

**Estoppel [`cli/lens/mint.go:155–158`, `reading.go:89–91`] (D10, F4 (a) RULED).** The guard is keyed on
`--quote` (`requireNotEstopped(l, quote, …)`, `mint.go:576`); a cited source's 32/25 names no quote, so
estoppel never asks of it; a found-source 32 is a quote mint and meets the guard like any gap: red
mints `--supersedes` the prescribing gap, as the refusal says. The "verify ATTACHES NO SUCH GUARD"
paragraphs (`mint.go:155–158`, `verify.go:408–411`) and the "A READING'S GAP IS EXEMPT … (AppendReading)"
paragraph of `estoppel.go:31–35` are deleted with the exemption they stated; `EstoppelRejectionsOf`'s
"counts lens mints alone" clause (`:35`) stays true and loses its reason. One logged estoppel where red's
found source refutes text blue applied from red's own `--new` is the accepted cost.

**`cli/lens/mint.go` [MODIFY]:** `--about-kind` usage (`:220–221`): "anchor — a citation or computation;
a CITED source's defect (claim_contradicts_source, citation_figure_misattribution) names the citation
here, the subject your reading of it is keyed on; a source YOU found is a `--quote` of the sentence";
`--class` usage for 32/25 is the enum help's, rendered from the `means` (§III.1). `mint.md` (§III.8)
carries the rule in one paragraph.

### III.4 The reading debt — `BlockerReadingDebt` (D4′) [NEW in `record/standingreading.go`; MODIFY `passblockers.go`, `closingowed.go`, `sitting.go`, `dispatch.go`, `retirement.go`, `verdict.go`, `enginegen/vocabulary.go`]

**The fact.** `readingDebtsOf(evs, win, gaps, found) []Blocker`, one `Blocker{Kind: BlockerReadingDebt}`
per lens X and per SUBJECT a reading of X's finds against (`VerifyReads` and `FindsAgainstOf(outcome)`)
and STANDS, while NO gap of the class the outcome obliges stands on the subject. A found source's
subject is the SENTENCE its claim locates to — the ruling's key — and `found []locatedReading` is every
found-source finds-against reading with the sentence it locates to in the report as it stands, computed
at the read (below), as `gaps` is read there:

| Reading | Stands while | The gap that discharges it | Also discharged by |
|---|---|---|---|
| of citation C (`refutes` → 32; `absent` → 25) | X's latest check of C that reads is this one, from a sitting opened after the latest edit naming C, C not retired — `anchorStanding.standing()` with `!c.holds`, the ledger's rule | a gap of the class ABOUT ANCHOR C by ANY seat that is open now, or was minted from this reading's sitting on (closed since) | X's later reading of C that reads and does not find against (the ledger replaces X's standing check); blue's edit naming C (the check is of a sentence the report no longer carries); blue's retire of C |
| of a found source for sentence S, `refutes` | this is X's LATEST found-source reading (any url) whose claim locates to S — `standingReadingOf(X, S)`; a reading whose claim no longer locates is LOST and stands alone (the lost-quote rule, below) | a 32 by any seat whose quote is in S (`gapSentenceIn` = `readingSentenceIn`), open now or minted from this reading's sitting on; for a lost reading, the gap joined as of the report before the loss (below) | X's later found-source reading of S, any url, that does not find against (the re-read of the subject the ruling names — a backing reading of another source for the same sentence discharges; a reading of the same url against ANOTHER sentence does not: that is another subject); the estopped edit; blue's edit that deletes the sentence outright (F8 (a)) |
| of a found source for sentence S, `absent` (F7 (a)) | as `refutes`, AND S carries no evidence mark — `anchor.BackingMark(sentence)` is "" (the predicate `refuseMarked` applies, `verify.go:591–616`: `anchor.Each` over the sentence span, the first id `anchor.Backs`; factored into `internal/anchor` as ONE helper the reading guard and this fold both call, replacing the guard's inline loop). `claimcount.BareAnchorIDs` is NOT it — it lists anchors with no prose before them, so a citation at the end of a sentence's prose reads as unmarked (T24's marked rows go red under it) | a 29 `unsourced_claim` by ANY seat whose quote is in S, open now or minted from this reading's sitting on — plain mint, no reading behind it | as `refutes`; also blue's cite that marks S (`BackingMark` non-empty) |

`Blocker` fields: `Subject` C, or the sentence (visible, whitespace-collapsed) for a found source;
`Owner` X; `since` the reading's `events.id`; `outcome` (new field, the reading's); `Detail` the url read
(found) or "" (cited); `lost` (new, `int64`) the `BlueEdit` events.id that took the claim out, 0 while it
locates or no edit is found. `owed()` is true for the kind while it stands — the debt IS the owing.
`Blocker.WorkItem()`, the kind's `refusal` and `why` are one row of `blockerKinds`, so
`TestEveryPassBlockerKindAgreesOnEverySurface` gains a row and every surface words it from the row. Cited
and found readings are two shapes of ONE kind: the item names the subject by its form (`readingForm`).

**"Open now, or minted from the reading's sitting on."** Ruling 1 says the MINT discharges the debt and
ruling 4 says a gap graded low and closed does not fail the run; ruling 8 says a second lens's debt is
discharged by any seat's OPEN gap. The three agree under one predicate: a gap of the class on the subject
that is open, or that was minted at or after the BOUND (whatever its state now). The bound is the opening
of the reading's sitting (`win.Of(e).Sitting` through the sitting's opening place) where the reading has
one; a reading in no sitting (`Sitting` 0 — a fixture, a seat acting unregistered) takes its own
`events.id`, since a zero bound admits every gap ever. The sitting's opening, not the reading's id,
because a gap minted BEFORE the reading in the same sitting is the reading's own (every archived in-act
member: Mint 180, Verify 182). A gap closed BEFORE X's sitting opened is not it: X read after that
adjudication and found against anew, and owes a fresh gap (the one-gap rule admits one, the old gap being
closed). Stated as a consequence, not a fork (§VII item 4). **Two sources for the two halves:** `gaps
[]blockerGap` is the OPEN set only (`blockerGapsOfStates` drops `!g.Open`, `passblockers.go:664–674`), so
"open now" is read off `gaps` and "minted from the bound on" is read off the Mint events in `evs` — the
`mints` map `failedChecksOf` already builds (`failedcheck.go:95–106`: gap id, seat, `about_ref`,
`events.id`), widened with the Mint's class and `location`. Pinned by T9's closed-low row, T19 row 4 and
T13's 182/542 rows: a fold that reads `gaps` alone turns each red.

**The surfaces — all through the one list, no new arm but the fold's own:**

| Surface | Reads | The debt |
|---|---|---|
| THE FOLD'S HOME: `eventBlockersOf` (`passblockers.go:732`), through `owedBlockersOf` (`closingowed.go:209`) | the stream arms every owed-blocker reader folds | `readingDebtsOf` is one more stream arm, placed after `failedChecksOf`'s. It goes HERE and not in `passBlockersOf` because `lensOwesOf` (`retirement.go:148`) reads the `events` slice `owedBlockersOf` returns — at `passblockers.go:704–706` computed BEFORE `passBlockersOf`'s own arms — and `dispatch.go:311` and `retirement.go:445` call `owedBlockersOf` directly: an arm appended in `passBlockersOf` would hold the PASS while retirement and `passLensGateOf` never saw it, and the lens would retire while it owes. `eventBlockersOf` and `owedBlockersOf` each gain `found []locatedReading`; callers: `passblockers.go:704`, `dispatch.go:311`, `retirement.go:445`, and the tests `failedcheck_test.go:283`, `anchorowner_test.go:70,125`, `inquiry_test.go:351` (each passes `nil`: fixtures with no found-source reading, stated at the call). Gate: the compiler; T18's `lensOwesOf[X]` row; T20's not-retired-at-3-barren-sittings row |
| `passBlockersOf` (`:692`) and its inputs | every kind | the signature gains `found`, passed through to `owedBlockersOf`; `found` is computed by ONE function `locatedReadingsAt(q, evs, win)` at each of the three read sites inside their snapshot — `passBlockerReadsAt` (`:816`, into `planReads`, which gains `found`), `planReadsAt` (`dispatch.go:264`) and `workView` (`viewjson.go:1459`, into `workReads` beside its `reportRender`, handed to `sittingOf` through `sittingReads` `:1549`) — and passed by the direct callers (`sitting.go:354, 371, 435, 500` from `reads.found`; `dispatch.go:542` and `:311`, `retirement.go:445` from `r.found`; `passBlockersOfReads` `:833`) as `gaps` and `fresh` are. `SittingOf` (`sitting.go:118`), which builds a zero `sittingReads{}`, gains `found` as a parameter: its callers are tests only (`available_test.go:33`, `citejoin_test.go:90`) and each passes what its stage holds. `locatedReadingsAt` renders the report ONCE through the checkpointed path (`RenderAt(q)`: one checkpoint read + ≤7 ops per fold; in `workView` the render `reportRender` already holds) and locates every found-source finds-against reading's claim in it (`readingSentenceIn`); for a LOST claim it renders once more as of the report before the loss (below; RULED built). A render failure is the READ's error, returned to its caller (`PassBlockers`, `WorkOfSeat`, the plan) as every other read error is — never a panic, never a fold over an empty report (K12); `blockerGap` gains `class`, `aboutKind`, `aboutRef`, `location` (gap-view columns) so the debt's open-gap lookup reads the set every other arm reads |
| the PASS gate (`requireNoBlockers` `:845`), `pass_permitted`, `dispatch.md:25`'s "a blocker another seat must clear" | kinds | the kind holds the PASS like any other |
| the outcome reason: `verdict.go:149` `anchorsStanding` AND the engine's `terminationWhyOf` (`debate.js:1094–1101`, authored `anchorsStanding`) | the kinds that hold a PASS by their own state | ONE table (D12, §III.8): `record.StandingClauses() []StandingClause{Kind BlockerKind; Clause string}` — `unchecked`/"unchecked", `failed_check`/"on a failed check", `reading_debt`/"on a reading debt", `open_inquiry`/"inquiries open", `unchecked_inquiry`/"inquiries unchecked" — folded by `anchorsStanding` (subjects deduplicated per kind, the clause `"; <clause>: a, b"`), rendered by `enginegen` into the generated vocabulary region as `const STANDING_CLAUSES = [{ kind: 'reading_debt', clause: 'on a reading debt' }, …]`, and folded by the authored JS `anchorsStanding` over that constant. The debt's subject in the clause is its `Subject` (C-…, or the sentence) |
| the chair's list (`sitting.go:374` `hold`) | `AnotherSeatClears: !ChairOwned()` | item: "lens X's reading of citation C-… at event N finds the source against its sentence (`refutes`), and no gap of class claim_contradicts_source stands on it — PASS is refused while it stands; X clears it by minting that gap about the citation, graded as it judges, or by reading the citation again at its next sitting" (a citation reading is keyed `anchor` — `seat.KeysOn(c, "anchor")` `verify.go:155`, `record/supersede.go:99` — so the same sitting takes no second one; in the sitting of the reading the lens's own discharge is the mint) |
| the lens's work list (`sitting.go:438–442`) | `b.Kind == BlockerFailedCheck && b.Owner == seatID` → `add` while `owed()` | the same arm, keyed on `b.owed() && b.Owner == seatID` for `failed_check` and `reading_debt` alike; `add` (Blocks true) — the envelope gate (`envelope.go:40–64` → `Sitting.Owed()` → `hookcmd.go:354`) refuses the `StructuredOutput` call with no new code |
| `lensOwesOf` (`retirement.go:148–175`) | `events` blockers with `owed()` | unchanged in code: the arm names every `owed()` blocker by owner, and the debt is one; named `reading_debt <subject>`; the lens does not retire while it owes one (`retired()` `:72`) |
| dispatch Source 4 (`dispatch.go:526–625`) | `BlockerFailedCheck` with `owed()` → readied once per cause (`firstAfter(ends.through[owner], b.since)` `:615`) | the same `case` admits `reading_debt`; the plan's `Why`: "X: its reading of C-… at event N finds the source against the sentence and no gap of class … stands on it — ready, to mint the gap or read again"; `failedCheckReadiness`' found-against arm (`:1391–1405`) is deleted |
| `verify`'s read-back (`verify.go:496–515`) | `BlockerFailedCheck` on the anchor | every blocker on the anchor by `WorkItem()`; the one this reading raised is prefixed "this reading is your debt until …" |

**How this differs from today's `stateFoundAgainst`.** Today's blocker holds the PASS whatever gap
stands (`failedcheck.go:39–47`) and its reader owes nothing; under D4′ the MINT discharges it, the gap's
grade then holds, the debt is per (lens, subject) rather than per anchor, and the gap's close is refused
by nothing but its demand (`requireFailedCheckAnswered` names unread and unsound only).

**Dispatch.** The owner is readied ONCE PER CAUSE, by the predicate every owed blocker uses:
`firstAfter(ends.through[owner], b.since)` (`dispatch.go:615`), where `ends.through` holds the OPENING
places of the sittings the seat sat through (`satThroughOf`, `:944`). So the reading's own sitting never
counts (it opened before the reading): a lens that reads `refutes`, mints nothing and ends that sitting —
its envelope refused, its agent stopped — IS readied at the next plan, once. A lens that then sits through
a sitting opened after the reading and still leaves the debt is not readied again for it, is not retired
(it owes), and the debt holds the PASS on the chair's list with the lens named — the run then ends
UNVERIFIED with the outcome naming it, as an unminted unread gap ends it today. An abnormal end
(`shortCallLimit`, `shortNeverClosed`) is not sat through and readies the lens again. T20 pins both.

**The lost quote (F1, RULED "Debt stands"; refined: "unless it took red suggested text, then estoppel
applies").** When blue's edit changes the words the found-source reading's `claim` quoted, the claim no
longer `LocateOnce`s: the reading has no subject sentence in the report as it stands, the mint check
refuses a `--quote` 32 until X reads the new sentence (the mint half of F1 (a) stands), and the DEBT
stands. `locatedReadingsAt` finds the edit (RULED 2026-11-06: BUILT, not deferred — the estoppel exception below requires it): the first `BlueEdit` after the reading whose `old` and the
claim CONTAIN each other's visible text either way (`anchortext.Visible`; the comparison
`EstoppelConflict` makes, `estoppel.go:86–100`) — either way because edits are sub-sentence spans as often
as whole ones: in the archive, edit 238's `old` holds 182's claim, and edit 692's `old` is a span INSIDE
542's claim. `lost` names it. **The gap that discharges a lost reading** is found where the claim last
located: the report as of `lost − 1` (`reportSourceAt(q, lost−1)`, one more checkpoint read + ≤7 ops per
lost reading; where `lost` is 0, as of the reading's own event), and the 32 by any seat minted from the
reading's sitting on whose quote locates in the claim's sentence THERE discharges it — a gap minted after
the loss cannot, and by the mint check none is (its minter had to read the new sentence first). The item
says: "your reading of U at event N read it against '<claim>'; blue's edit at event M changed that
sentence to '<the edit's new, as a sentence>' — read a source against the sentence as it now stands, or
mint on a reading of it". It also clears by X's later found-source reading whose claim locates to the
edit's new sentence and does not find against (one that finds against creates a fresh debt on the new
sentence, with its own mint route). **Estoppel discharge:** where the edit at M carries
`applied_verbatim` (whatever gap it `answers`, whichever red seat minted it — red is the party bound,
`estoppel.go:82–87`), the debt is discharged: red wrote the new sentence, so red holds no unfinished
promise against it. X may still re-read; a finds-against re-read creates a fresh debt, whose mint route is
the estopped one — a 32 `--supersedes` the prescribing gap (D10). Residue: where no edit's `old` meets the
claim either way (the report re-ingested whole; the claim now occurs twice), `lost` is 0, the item names
the reading alone, the gap join is made as of the reading's own event, and the debt otherwise clears by
X's later found-source reading of the same url — the one identity left to it — that does not find against.

**The deleted sentence (F8 (a), RULED "deletion discharges").** Where the edit at M that meets the claim
in its `old` carries no sentence in its `new` (deleted outright), the debt is DISCHARGED — as blue's edit
of the cited row's sentence discharges that row; the reading stays as information (`show evidence`
`independent`). The lens may read U again against any sentence the claim reappears in; a finds-against
reading there is a fresh debt on that sentence. The predicate is the edit's, not the claim's absence: a
claim that stops locating because its words CHANGED is the lost quote above (debt stands); one that stops
locating because its sentence was REMOVED is this (debt gone). Accepted cost, stated in the ruling's
option: blue deletes a refuted sentence and the lens's promise is void — the defect left the report with
the sentence.

**A found source read `absent` (F2, RULED "Debt like refutes"; F7 (a), RULED).** The same debt, with one
input more: the sentence's marks. A found page silent on an UNCITED sentence is the lens's failed search
for backing, which is what 29 `unsourced_claim` carries to blue ("the gap stands until blue cites one"),
so a 29 by any seat on the sentence discharges it; a found page silent on a sentence that carries a
citation (or a computation) says nothing against that evidence, and the reading creates no debt. 29 stays
a plain mint: nothing checks a reading behind it, and a lens that found three silent pages for one
uncited sentence owes one 29 (the one-gap rule). Consistent with `verify` today: `refuseMarked`
(`verify.go:584–616`) refuses a BACKING found-source reading of a marked sentence and lets an `absent`
one record (`verify.go:369`, `backing` excludes `FindsAgainst`); `mint --class unsourced_claim --quote S`
asks nothing of S's marks (no reader of `anchor.Backs` in `mint.go` or `record/` judges a 29) — the
predicate "unmarked" is the debt's and the reading guard's, never the 29 mint's (§VII item 7), and it is
ONE helper, `anchor.BackingMark`, for both (above).

**Census: information surfaces that show a found-against reading, unchanged by the debt** — `show evidence`
source and `independent` rows (`evidenceview.go:223–266, 303`; `CarriedBy` `:266, 436` deleted);
evidence.md "Citations" (`report/citations.go:21–60`); docket.md's "read for the sentence" line
(`assemble.go:990, 1096–1120`, per gap via `ReadingBehind`, §III.6); `counts.sources_refuted` is NOT among
them — deleted (§III.5);
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
- **`CitationsFoundAgainst` (`:306–323`) and `counts.sources_refuted` [DELETE] (RULED, relayed):** the
  count's only destination would be one help line (`verbs.go:307`), which the field-destination rule
  refuses. Deleted: the function; `EvidenceCounts.SourcesRefuted` and its comment (`evidenceview.go:324–329`),
  its fill (`:593`), the `sources_refuted` clause of `verbs.go:307` and the `counts.sources_refuted`
  key in the OUTPUT line (`:303`); tests `evidenceview_test.go:265–267, 294` (the assertions, not the
  tests), `foundagainstm30_test.go` whole (`TestM30SourcesRefutedCountsOnlyTheStanding` and its helper
  read the state); the manual golden's three OUTPUT lines (G1) re-render. The remaining ledger readers
  (§II) are unaffected: `sources_unverified` counts `unchecked|unread` and counted no found-against
  citation before. The capture's PASS detail "ever found against" count (`capture.go:499–523`) reads
  `EvidenceVerificationJSON.FindsAgainst` per reading, not this count, and stays.
- **`FailingAnchors` (`:325–340`) → `AssemblyFailures(run Run) ([]AssemblyFailure, error)`** (D11), an
  exported Run-level entry (capture is another package): it reads `passBlockerReadsAt` on one snapshot —
  the fold's inputs, `found` included — and folds; `AssemblyFailure{Kind, Subject, Why string}` is the
  exported row. The capture's assembly screen's one read (`capture.go:473–495`); `EvidenceTable.Failed`
  (`evidencetables.go:47–49`, the citation row's `"found_against"` `:55`, the computation row's `"unsound"` `:60`) is deleted with its one reader (`anchorchecks.go:335`) — the computation arm below reads `stateUnsound` directly. A row is one of: (i) a computation
  whose ledger state is `stateUnsound` (unchanged); (ii) an OPEN MATERIAL gap of class 32 or 25 whose
  subject the assembled report still carries — about anchor C, C not retired; or quoting a sentence
  that still locates (`gapSentenceIn`); (iii) a standing `reading_debt` (`readingDebtsOf` — a found-wrong
  reading with no gap at all, red's unfinished promise). The FAIL detail names each row by its why:
  "citation C-… (url): open material claim_contradicts_source gap G-… by X" / "sentence '…': open
  material … gap G-…" / "citation C-…: X's reading at event N finds it against and no gap stands on it".
  A refuted citation whose gap was graded low, or closed, is in no row. The PASS detail's "ever found
  against" count (`:499–523`, `EvidenceVerificationJSON.FindsAgainst`) is unchanged through the predicate.
- `passblockers.go`: the `CheckFoundAgainst` arms in `failedFact` (`:227`), `failedClears` (`:238`),
  `failedRemedy` (`:253`), `failedOwed` (`:266` → two arms), `blueFailedItem` (`:277`),
  `failedCheckRefusal` (`:299–325`), `blockerKinds`' `failed_check` row (`:413–429`: gloss, `why`'s
  found-against case), `Blocker.owed` (`:119–121`), the `answeredBy`/`carriers` comments (`:86–93`);
  `dispatch.go:1391–1405` (`failedCheckReadiness`); the `reading_debt` row arrives (§III.4).
- `retirement.go:140–146, 164–168` (`lensOwesOf`): the comments "A found-against check never holds its
  reader" / "A FOUND-AGAINST CHECK IS NEVER OWED" go; the fold is unchanged in code (it names every
  `owed()` blocker by owner). `sitting.go:438–440`'s comment
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

- **`record.ReadingBehind(evs, win, gap, found) (*Verify, bool)`** [NEW, in `standingreading.go`]: for a
  gap of a class in `readingMintedClasses` by seat X — about anchor C: `standingReadingOf(X, C)`; with a
  quote: `standingReadingOf(X, the gap's sentence)` — X's latest found-source reading, any url, whose
  claim locates to the gap's sentence, read off `found`; where the gap's quote no longer locates, the
  lost reading whose as-of join (§III.4) named this gap. False for any other gap or where none stands. The
  ONE statement of "the reading behind this gap"; every Go reader below calls it. No SQL twin: the cited
  case needs none (the gap's `about_ref` and `minted_by` are columns; a SQL reader joins `latest_check` on
  them), and the found-source case needs the report.
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

- **Archived shapes, read off the archive's rows** (epochs 55–64; one archive, Eiffel, epoch 56; §II
  table): (1) an in-act reading OF A CITATION — Mint m (`location` = the claim, NULL `about_kind`),
  Anchor a (`id` = m's gap, `sentence_of` = C), Verify v (`anchor` C, `carried_by` m's gap,
  `carrier_minted` 1): FOUR (165/166/167 → G-8dbfce13, class 25, C-2dc96eab; 789/790/791 → G-f66af91d;
  792/793/794 → G-dd7b403e; 795/796/797 → G-de476176, class 32, C-a51c49f1 / C-a1560eb7 / C-486ed6b5);
  (2) an in-act reading OF A FOUND SOURCE — Mint m (`location` = the claim), Anchor a (`location` = the
  claim), Verify v (`url`, no anchor): TWO (180/181/182 → G-3b2b7d93; 540/541/542 → G-1d97d860); (3) a
  held reading — v with `carried_by` naming another act's gap, `carrier_minted` 0: ONE (798, of
  C-3094c824, naming G-1d97d860). The previous draft called all six quote gaps and all six cited; the
  rows say four cited and two found.
- **The machinery deleted:** `readingActs`, `readingActMembers`, `heldMember`, `writeReadingAct`,
  `failAct` and the `held`/`actMember` plumbing (`migrate/replay.go:127–130, 237–260, 547–561,
  902–990`), `Replay.WriteBodies`/`readingActOf` (`record/replaywriter.go:130–207`), `writeView.act`
  (`writeview.go:30–34`), `appendBodies`' `act` and held arm (`record.go:674–720`), `appendAct`'s `act`.
  Every member is written singly in stream order under the replay's one writer; the two link fields are
  absent from the proto, so the source's `carried_by`/`carrier_minted` columns are dropped at `OldEvent`
  → body translation (the Verify arm, beside `eraentries.go:344–348`'s renames); `remap.idFields` loses
  `Verify.carried_by` (`remap.go:116`). `raisedReading` (`:859–873`) stays (a raise names a reading by
  event id, a different relation).
- **The ONE translation (RULED 2026-11-06):** an in-act Mint whose carrying Verify names a citation —
  shape (1): the Mint's `carried_by` partner, read in the same pass that today builds `readingActs` —
  comes forward ABOUT that citation: `about_kind` anchor, `about_ref` = v's `anchor`, `location` dropped.
  Both values are in the source rows; no fill. Shape (2)'s Mint is already the new design's found-source
  mint (`--quote` the sentence) and is written as it is; shape (3)'s Verify is written alone. The Anchor
  rows of shape (1) STAY: an Anchor is a report op (`reportcheckpoint_test.go:65`), so dropping one
  would change the rendered report at every later event and break the `old` of every later edit of that
  sentence; `validate`'s Anchor arm admits `sentence_of` (`record.go:970`). Residue, stated: the four
  migrated gaps are about their citation AND carry the token their in-act mint placed at the citation's
  sentence — a pair a live `--about-kind anchor` mint never writes (`AppendPlaced`, `record.go:660`
  places only with a quote); readers of gap tokens (the board, `gapSentenceIn`) see the token as today.
- **What the migrated record then says (T13's expected set, computed from the rows above):** no
  `reading_debt` stands at head — 167's anchor C-2dc96eab was retired (223); 791/794/797 each have the
  lens's later `supports` of the same anchor (952/953/955) and a blue edit naming it (835/858/848), so
  none stands; 798 likewise (954; edit 846); 182 and 542 are LOST (edits 238, 692) and each joins its gap
  as of the report before the loss (G-3b2b7d93 at 180, G-1d97d860 at 540 — both minted in the reading's
  sitting, BEFORE the reading: the sitting-open bound). Through event 845 the set is exactly ONE:
  {owner red-lens-evidence, subject C-3094c824, since 798, `refutes`} — the held reading's citation had
  no gap of its class and no later check until edit 846 named it. `ReadingBehind` names 167 for
  G-8dbfce13, 182 for G-3b2b7d93, 542 for G-1d97d860 (the as-of join; NOT 798, a citation reading), and
  952/953/955 (the standing `supports`) for the three repaired 32s.
- **Residue, stated:** every archived 25/32 mint comes forward UNJUDGED by `requireOwnStandingReading` —
  4 archived 25 mints in 4 runs, Eiffel's 6 class-32 mints (G-b2ba7d66's plain mint among them) — because
  the check is a live act judgement and the replay asks none. Archived found-against checks hold nothing
  after migration: there is no such state. The four translated gaps are closed, so the one-gap anchor arm
  asks nothing of them.
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
| `help/verify.md:15,19,21,25,27,29` | in-act mint; one-gap holds the reading; found against holds the PASS; estoppel exemption; `--url` mints by outcome; result names `gap` | a reading records; a `refutes`/`absent` on a citation is the ground of a gap you mint about the citation (an act; the flags are `mint`'s page's); a refuting found source is the ground of a gap quoting the sentence; a found source silent on an uncited sentence is the ground of the gap that asks blue for a citation, and on a cited one is information; a reading that finds against is YOUR DEBT — it holds the PASS as your unfinished promise, on the chair's list with your name, until a gap of its class stands on the subject or you read the subject again, and your envelope is refused while it stands; from the mint on, the gap's grade is what holds; the result names the event |
| `help/mint.md:9,15,17` | one-gap on sentence/inquiry; budget exemption; "a reading … mints its gap in the same act" | one-gap on a sentence, an inquiry OR a citation (every class; the exception is the gap your own unread/unsound check obliges); the standing-reading rule for 32/25 in one paragraph — a CITED source's defect names the citation (`--about-kind anchor`), the subject your reading is keyed on; a FOUND source's names the sentence (`--quote`); what standing means (your latest reading of that subject since its sentence last changed); the refusal; the budget sentence (§III.3); estoppel asks of every `--quote` mint |
| `help/verdict.md` blocker glosses (`PassBlockerGlosses`), `help/dispatch.md:25`, `help/outcome.md` | the kind list | the `reading_debt` gloss arrives: "a lens's reading that finds a source against its sentence, with no gap of its class standing on it" |
| `help/close.md:13` | "your gap another lens's reading named … found-against check clears…" | unread and unsound only |
| `help/edit.md:13` | "a lens's reading … still mints its gap" | an edit that changes a sentence a found source was read against leaves its lens's debt standing; one that applies red's text verbatim discharges it |
| `help/reproduce.md:23,27` | unsound; "on a failed check" | unchanged — read before/after all the same |
| `seat/verbs.go:303,307,339`; `seat/seat.go:1085`; `cli/setup.go:89` | `carried_by` (`:303`); `counts.sources_refuted` (`:303` key, `:307` sentence); budget exemption (`:339`); reason-is-the-gap's-argument | §III.6; the `sources_refuted` key and sentence DELETED (§III.5); the exemption sentence (§III.3); reason is the reading's |
| `record/passblockers.go` found-against arms, `dispatch.go:1391–1405`, `reading.go:120,128,132` refusals, `failedcheck.go` refusals, `mintbudget.go:308`, `verdict.go:140–160` | | deleted / unread-unsound only; the debt kind's row (§III.4, §III.5) |
| `record/dispatch.go:21` — `Party.GapIDs`'s `desc:` tag, a Go struct tag that RENDERS into the agent-facing plan schema: `debate.js:394` (`LENS_OR_BLUE_PARTY.gap_ids.description`, the generated envelope vocabulary) | "empty for blue readied to remedy a failed check — … edit the sentence a lens's reading finds against or retire or replace its citation" | the clause goes (a found-against reading readies blue no longer; its GAP does, as a material gap) and the lens's cause arrives: "empty for a lens readied to audit the report — …, or by a check of its that failed, or by a reading debt of its own (a source its reading found against its sentence, with no gap of that class standing — it mints the gap or reads again), or a position the bench upheld naming its area; empty for blue readied to remedy a failed check — fix or retire a computation judged unsound, retire a citation no check could read — or to record a fate on an open inquiry; …". Property, gated by §V check 7: no `desc:` tag under `internal/record` says "finds against"/"found against"; every one naming a readiness cause names unread/unsound and the reading debt. The `blockers.kind` enum in the same region (`debate.js:415`) gains `reading_debt` from `BlockerKind`, and the region gains one constant, `STANDING_CLAUSES` (D12, next row). All by `enginegen`, never by hand |
| **D12 — the outcome reason's clause, one table, two carriers.** `verdict.go:149–175` `anchorsStanding` (Go, the record's derived reason) and `debate.js:1094–1101` `anchorsStanding` (authored JS, the engine's `terminationWhyOf`) each enumerate `BlockerKind` words and compose the clause: Go names `unchecked`, `failed_check`, `open_inquiry`, `unchecked_inquiry`; the JS names `unchecked` and `failed_check` ONLY ("Those two kinds only") — they disagree today on the inquiry kinds, and the debt would widen the gap: the record's reason would name a `reading_debt` the engine's would not | two hand-written enumerations of one enum | ONE declaration: `record.StandingClauses()` (`verdict.go`, beside `anchorsStanding`), a `[]StandingClause{Kind BlockerKind; Clause string}` in kind-table order — `unchecked`/"unchecked", `failed_check`/"on a failed check", `reading_debt`/"on a reading debt", `open_inquiry`/"inquiries open", `unchecked_inquiry`/"inquiries unchecked". Go: `anchorsStanding` folds it (one loop; subjects deduplicated per kind — today only `failed_check` deduplicates, and the other kinds' subjects are unique by construction, so no output changes). JS: `enginegen`'s `renderVocabulary` (`vocabulary.go:89`) emits `const STANDING_CLAUSES = [{ kind: 'unchecked', clause: 'unchecked' }, …]` inside the generated region (`vocabularyFromTheRecord` `:51` reads the table; the render REFUSES a clause whose kind `BlockerKinds()` does not hold, as `TestAVocabularyAHarnessCouldNotValidateAgainstIsRefused` `vocabulary_test.go:118` refuses the other unvalidatable shapes), and the authored `anchorsStanding` (`:1094–1101`) becomes a fold over `STANDING_CLAUSES` — its comment loses "Those two kinds only" and says the table is the record's. **Why generation and not a parity test:** the kinds list beside it (`:415`) is already generated from the same enum into the same region, so the clause table is one more declaration rendered there, and `enginegen -check` (CI) already fails a stale copy; a Go test holding the authored JS to the Go table would be a regex over script text, the shape #1466 refuses. **Gates:** `enginegen -check`; `TestTheVocabularyRegionStatesEachSetAndThePlanAsDeclared` (`vocabulary_test.go:46`) gains the constant in its `want`; `TestACeilingReasonNamesOnlyAStandingReopenedAnchor` (`reopened_test.go:544`) gains a `reading_debt` row ("; on a reading debt: C-1"); `debate.test.mjs:264–275` gains fixture blockers of kind `reading_debt`, `open_inquiry` and `unchecked_inquiry` and expects `clause` = "; unchecked: C-706fda63, K-cd9f4c5a; on a failed check: C-0b0e0d0a; on a reading debt: C-0b0e0d0a; inquiries open: Q-…; inquiries unchecked: Q-…" (the inquiry rows are the disagreement found by this sweep, fixed by the same table — stated in §VII) |
| `record.proto` `SourceOutcome` `means` `:531–542`, `GapClass` 25/32, `AboutKind.ANCHOR` | "mints … in the same act" | §III.1 |
| `scripts/agentgen/src/fragments/lens-audit-and-mint.md:3,6,7,9,10`; `fragments/lens-duties.md:5,6`; `fragments/blue-search-and-cite.md:30–31,35,43`; `agents/red-lens-evidence.md:17,19`; `agents/lead-judge.md:81`; `agents/red-chair.md` (the blocker list); `skills/research-protocol/SKILL.md:14–21` | "records its gap in the same act"; "whose reading it carries"; "found against" as a failed check | the reading records; the lens mints the gap about the citation on its reading, or quoting the sentence for a found source, and owes one or a re-read while its reading finds against — the chair sees the debt on its list and names the lens; the demand is met through the citation the gap is about |
| `skills/research-protocol/scripts/debate.js:883` (authored `ledgerClause`) | "your reading records the gap, and blue answers it with a citation" | "where it refutes the sentence, you mint the gap on that reading — a reading that finds against is yours to answer; where it backs an uncited sentence, or says nothing of it, you mint the gap that asks for the citation" (acts, no verb path; `TestTheSeatPromptsNameNoVerb`). The authored `anchorsStanding` `:1094–1101` is D12's fold (row above). Sweep (§V check 7's last line, the quoted kind words `unchecked`, `failed_check`, `reading_debt`, `open_inquiry`, `unchecked_inquiry` over the file) → today `:415` (generated) and `:1099–1100` (authored); after, only the generated region names a kind word, and `:1099–1100` name `STANDING_CLAUSES` |
| `tools/internal/terms/terms.json` rows `reading` (`:772–773`), `failed check` (`:1101–1102`); NEW row `reading debt` | in-act mint; third state | reading: "…it holds nothing by itself; a gap a lens mints on it is the finding"; failed check: two states; reading debt: "a lens's standing reading that finds a source against its sentence with no gap of its class on the subject — the lens's unfinished promise, a PASS blocker it clears by the mint or a re-read"; then `go -C scripts run ./vocabdoc` (`docs/vocabulary.md:464,632`) |
| `docs/citations-and-bibliography.md:162–164,174–179,181–188,195–197,209–217`; `docs/seat-command-triggers.md:55,67` | three states; in-act; `carried_by` | two states and the debt; the standing-reading mint; the docket line |
| `report/layout.go:167,184` | "as a lens's `verify` does where its reading mints a gap" | deleted clause |
| `plans/compute-and-show.md` (40 lines, §III.9 row P) | `Verify.carried_by`, carrying reading, `found_against` | one dated note at `:515–523` and `:1572` that #1727 removed the link and the state; the rest is history and stays |
| GENERATED: 12 `agents/*.md` (3 `carried_by` hits each from the inlined manual), `skills/adversarial-audit/SKILL.md`, `skills/research-protocol/SKILL.md`, `help_gen.go`, `annotations_gen.go`, `vocabulary.md`, `debate.js` marked regions — `:105–123` unchanged; `:373–451` CHANGES at `:394` (the `Party.GapIDs` desc), `:415` (the `blockers.kind` enum gains `reading_debt`) and ONE new line (`STANDING_CLAUSES`) | | `agentgen`, `cmdtreegen`, `vocabgen`, `vocabdoc`, `enginegen`; the simulator's prompt goldens that carry the plan schema re-render (G8) |

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
| C1 | `recordpb/record.proto:152–165` (50014/15), `:529–545` (outcomes), `:460`, `:483` (25, 32), `:1179` (AboutKind), `:1431–1444` (Verify) | 1 |
| C2 | `recordpb/gapclass.go:63–77`; `recordpb/annotations.go:39–42,92`; `devcmd/vocabgen/main.go:169–170`; `annotations_gen.go` (5, generated); `record.pb.go` (31, generated); `recordsql/decode_gen.go:1475` (generated) | 1 |
| C3 | `record/enums.go:344–360` (`ReadingClassField`) | 2 |
| C4 | `record/reading.go:19–183, 225, 244–247, 259–269` (keep `readingSentenceIn` `:185`, `readingForm`, `requireNotARetriedReading` minus `:225,244–247`) | 2, 3, 5 |
| C5 | `record/record.go:674–720` (`appendBodies` act/held), `:747` (`appendAct`), `:1735–1741` (`validate` Verify arm: `requireReadingCarrier` goes), `:1931` (31's check STAYS) | 3, 7 |
| C6 | `record/actjudge.go:20–23` (list), `:40–64` (Mint arm: `requireAboutAnchorLive`, then `requireOwnStandingReading`, then the one-gap rule), `:131–135` (`requireReadingNamesCarrier`); `record/estoppel.go:31–35` (the "A READING'S GAP IS EXEMPT … (AppendReading)" paragraph) | 3 |
| C7 | `record/writeview.go:30–34` | 7 |
| C8 | `record/onegapperclass.go:16–50, 75–77, 89–96, 101–108` | 3 |
| C9 | `record/mintbudget.go:32–44, 252–259, 308, 320–345, 375` | 3 |
| C10 | `record/failedcheck.go:12–47, 57, 72–89, 110–123, 161–166, 172–204, 241–243, 256–381, 383–427, 429–441` | 5 |
| C11 | `record/anchorchecks.go:125–130, 146–163, 187–195, 279, 292–298, 306–340` (`CitationsFoundAgainst` and `FailingAnchors` deleted) | 5 |
| C12 | `record/passblockers.go:29–44, 86–93, 119–121, 227–325, 413–429, 664–714, 732–750, 784–835`; `record/closingowed.go:206–212` (`owedBlockersOf`, gains `found`); `record/dispatch.go:21` (the `desc:` tag, A10), `:255–283` (`planReads`/`planReadsAt`), `:311`, `:526–625`, `:542`, `:1391–1405`; `record/verdict.go:140–175` (D12's table beside `anchorsStanding`); `record/sitting.go:118–130` (`SittingOf`, `sittingReads`), `:354, 371, 435, 500` (the direct `passBlockersOf` callers), `:438–442`; `record/viewjson.go:1101–1129` (`workReads`, `reportRender`), `:1459` (`workView`), `:1549` (the `sittingReads` build); `record/retirement.go:140–175, 431–447` | 4, 5 |
| C13 | `record/retirement.go:140–146, 164–168` (comments) | 5 |
| C14 | `record/evidenceview.go:262–266, 269–283, 324–329, 436, 593` | 5, 6 |
| C15 | `record/demand.go:16–35, 83, 123–139`; `record/windowindex.go:81, 95–104` | 6 |
| C16 | `recordsql/views.go:412–420, 435–459, 493–495, 597–621`; `recordsql/schema.go:259–261`; `recordsql/evidencetables.go:47–49, 55, 60, 75–84` | 1, 5, 6 |
| C17 | `record/replaywriter.go:130–207` | 7 |
| C18 | `record/migrate/replay.go:127–130, 237–260, 547–561, 859–873 (stays), 902–990`; `migrate/remap.go:116`; `migrate/eraentries.go:331–348` (Verify arm, the drop) | 7 |
| C19 | `cli/lens/verify.go` (§III.2 table) | 2 |
| C20 | `cli/lens/mint.go:130–158, 217–222, 235–266, 576–584` | 3 |
| C21 | `cli/seat/verbs.go:303, 307, 339`; `cli/seat/seat.go:1085`; `cli/setup.go:89` | 8 |
| C22 | `report/assemble.go:1098–1125`; `report/layout.go:167, 184`; `report/citations.go` (unchanged) | 6, 8 |
| C23 | `seatprobe/seatprobe.go:206–211` | 6 |
| C24 | `capture/capture.go:371, 473–523` (`FailingAnchors` → `AssemblyFailures`; the FAIL/PASS details) | 5 |
| C25 | `recordsql/testdata/schema.sql`, `schema.epoch.json`; `requirements.json:3` | 1 |
| C26 | NEW `record/standingreading.go` (`readingMintedClasses`, `standingReadingOf`, `requireOwnStandingReading`, `mintSentenceAt`, `ReadingBehind`, `readingDebtsOf`, `locatedReading`, `locatedReadingsAt`, `AssemblyFailures`); `anchor.BackingMark` (`internal/anchor`, factored from `verify.go:591–616`); `recordpb` accessor `FindsAgainstOf`; `Blocker.outcome`, `Blocker.lost`; `blockerGap` widened; the `reading_debt` row in `blockerKinds`; `record.StandingClauses` (`verdict.go`); `devcmd/enginegen/vocabulary.go:51–134` (renders `STANDING_CLAUSES`) | 3, 4, 5, 6, 8 |
| C27 | `releasegate/fuzz` — §III.10, 14 files | 10 |

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
`replaywriter_test.go` (3), `unreadcitation_test.go`; capture (1): `capture_test.go` (4); fuzz (14):
§III.10. **Plus the signature callers outside the sweep:** `owedBlockersOf` — `failedcheck_test.go:283`,
`anchorowner_test.go:70,125`, `inquiry_test.go:351` (pass `nil` `found`); `SittingOf` — `available_test.go:33`,
`citejoin_test.go:90`; `evidenceview_test.go:265–267, 294` (the `SourcesRefuted` assertions go); the clause
carriers `reopened_test.go:544`, `devcmd/enginegen/vocabulary_test.go:46`, `plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs:264–275`. **Plus 10 cli files outside the sweep that run `verify` with gap flags through `gapFlagArgs`
(`verifyclaim_test.go:211`):** `verifyclaim_test.go:62,64`, `mintbare_test.go:157`,
`verifyconfidence_test.go:28`, `bridgekind_test.go:63`, `deadpath_test.go:625`, `gapanchor_test.go:177`,
`requiredmarker_test.go:120`, `verbs_test.go:117`, `answers_test.go:297`, `verifycopy_test.go:203`. The
31 test `record/inquiry_test.go:427` is untouched. (Note: a broad grep for `\.Gap\.ID` also matches `record/queue_test.go:57`, which tests gap queue items and is untouched by reading carriers, yielding 46 files on disk; the 45 T-sweep files are the reading- and carrier-touching files listed above).

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
| G8 | `tests/simulator/testdata/prompt-blue-synthesize.golden:7`; `prompt-red-lens-evidence*.golden`; `prompt-red-chair.golden` | `debate.js:883`; the chair's blocker list |
| G-arch | the archive-iterating set: 39 tests across `migrate/*_test.go` (`everyarchive`, `everyrender`, `proserewrites`, `readingsforward:405`, `archiveonce`, `argumentresidue`, `remaplists` ×2, `gapclass_internal`, `archivedgaps`, `evidenceanchor`, `gapplace`, `benchfiling`, `benchoccasion` ×2, `latestsitting`, `regraderesult`, `storedwindows`, `eventcache`, `reportcheckpoint`, `unmigratedarchive`, `computationarchive`, `computationresidue` ×9, `correctionarchive`, `dispatchpin`, `cutfields`, `remanddirection`, `reproduceanchor`, `findingarea`, `inquirymethod`, `proserefs`), `record/backingarm_test.go:254`, `devcmd/archiverun/outcomes_test.go:103` | Part 0 adds an archive |

**Agent-facing (A), 13 sources:** A1 `help/verify.md` (7 lines); A2 `help/mint.md` (3); A3 `help/close.md:13`;
A4 `help/edit.md:13`; A5 `record.proto` outcome/class/about-kind `means` (9); A6 `seat/verbs.go:303,307,339`,
`seat/seat.go:1085`, `cli/setup.go:89`; A7 `agentgen/src/fragments/lens-audit-and-mint.md:3,6,7,9,10`,
`lens-duties.md:5,6`, `blue-search-and-cite.md:30–31,35,43`; A8 `agentgen/src/agents/red-lens-evidence.md:17,19`,
`lead-judge.md:81`, `red-chair.md`; A9 `agentgen/src/skills/research-protocol/SKILL.md:14–21`; A10
`debate.js:883` and `:1094–1101` (authored) AND the generated region `:373–451` at `:394` (the `desc:` of
`dispatch.go:21`, grep `finds against` on the file: 1 hit), `:415` (the kind enum) and the new
`STANDING_CLAUSES` line — through `enginegen`; A11 `terms.json:772–773, 1101–1102` + the new row; A12 `docs/citations-and-bibliography.md` (6 passages),
`docs/seat-command-triggers.md:55,67`; A13 the generated set (12 agents, 2 preloaded skills,
`help_gen.go`, `annotations_gen.go`, `vocabulary.md:464,632`), plus `help/verdict.md`, `help/dispatch.md:25`,
`help/outcome.md` through `PassBlockerGlosses`. Refusal strings (Go, agent-read): `reading.go:120,128,132`,
`failedcheck.go:289–299,336–344,410–423`, `passblockers.go:227–325`, `dispatch.go:1391–1405`,
`verify.go:394,529–546,694–709`, `verdict.go:140–175`.

**Views (V), 3:** V1 `carrying_reading` (deleted); V2 `demand_reach` (reading arm deleted); V3 the gap
view's `evidence_met` comment. **1a′ readers (L), 4:** all through V2's 'about' arm after the change (§III.6).
**Plans (P):** `plans/compute-and-show.md` 40 lines (§III.8).

**Census totals:** 31 code sites (28 modified, 1 new, 2 generated families); 45 T-sweep test files + 10
gap-flag files + 11 signature/clause-carrier test files; 14 fuzz files with 18 tally fields and 2 lifecycle tallies (§III.10); 8 golden files + 39
archive-iterating tests; 13 agent-facing sources + the generated set (one `desc:` tag among them); 3
views; 4 1a′ readers; 1 plan.

### III.10 Release fuzz — `releasegate/fuzz/` [MODIFY] (CI only, never run locally)

| Site | Today | Becomes |
|---|---|---|
| `reading_test.go` (51 lines: `readingClasses:32`, `readingGapArgs:61`, `checkReading:144–173`, `readCitation:275`, `foundAgainstCleared:399`, `checkCarriedBy:458`, `notTheWorkReading:474`, `readFoundSource:577`, `readAnotherSource:715`, `fixFoundAgainst:774`, `readingTally:831` — callers: `fuzz_test.go:7882`, `onegaptally_test.go:53`, `readRepeatedSentence:1101`, `readOverPrescription:1270`, `readNotTheWorkFound:1335`) | in-act mint, held readings, `carried_by` checks, found-against remedy | `verify` alone; then, on a coin, `mint --class … --about-kind anchor --about C` (cited) or `--quote S` (found) and the stored gap's `ReadingBehind` is the reading; the second lens's mint about C refused by the anchor arm; `readOverPrescription` expects the estoppel refusal (D10) and then the `--supersedes` mint |
| `foundagainstseed_test.go` (47: `forcedFoundAgainstChain:111`, `foundAgainstReading:284`, `foundAgainstReread:297`, 6 tests `:411–811`) | forced found-against chain through the failed check | forced chain: reading → the `reading_debt` holds the PASS and the lens's envelope → mint on it → the gap's materiality holds → blue's edit/retire → the demand met through the 'about' arm; the "check holds the PASS whatever gap stands" assertions go |
| `failedcheckstate_test.go:18–100, 143–178, 190–210` | `CheckFoundAgainst` drives | two states only; the found-against test deleted |
| `anothersource_test.go:18–82` | second source held by the first's gap | second found-source mint refused by the one-gap rule on the sentence; the second lens's debt discharged by the first's gap |
| `lifecycleseed_test.go:94–95, 110, 126–127, 153–154, 312–334, 533–574` (`ReadingDemandRetired`/`Reread`, `lifecycleRead`, `lifecycleReadingGap`); `sourcedemandseed_test.go:56–117` | demand reached through a carried reading | demand reached through the gap's `about_ref`; the retire and the minter's backing re-read each met |
| `fuzz_test.go:549–559, 1220–1246` (`sourceQuoteRefused`, `FoundSourceGapClassesOf`), `:797–805, 2697–2708, 6909` (`forceFoundAgainst`), `:1462, 3479, 3871, 4895` (`CheckFoundAgainst` arms), `:5214, 5988, 6073` (the chain and quote drives), tally fields `:6345–6350` and their copy-out | | rewritten per row; the `CheckFoundAgainst` arms deleted |
| `lensturn_test.go:225–245` (`carryingOf`: `SELECT "anchor", "gap_id" FROM "carrying_reading"`; callers `:280`, `remandroute_test.go:514`) | the reading-to-gap link off the view | `ReadingBehind` over the run's record for each 32/25 gap — the SQL and the view are gone (§V check 7 greps `carrying_reading` over `*.go`, the fuzz included) |
| `parts_test.go:140–153` (`withholdsPart`: `r.forceFoundAgainst`, 19 uses across the package); `wordsbefore_test.go:11–15` (comment naming `fixFoundAgainst`'s oracle) | the found-against seed's lane rule and its oracle's name | renamed with the seed (`forceReadingDebt`, the debt chain); the comment follows the oracle's new name |
| `sweeptally_test.go:205–224` fields `FoundAgainstClosesRefused`, `FoundAgainstRepeats`, `FoundAgainstEdits`, `FoundAgainstRetires`, `FoundAgainstClearedByOwnReading`, `FoundAgainstStage`, `ReadingDemandRefused`, `NotTheWorkFlagsRefused`, `FoundSourceBudgetRefused`, `ReadingsMinted`, `ReadingsHeld`, `FoundSourceGaps`, their copy-out `:267, 636, 777–800`, the log `:1162–1166` and their zero gates `:1181–1196` | 18 reading-shaped fields | deleted: the eight found-against/held/minted/flag fields. Replaced (new meaning, new name): `reading_mints_refused` (map by class), `reading_mints_landed` (by class and subject form), `second_reading_mints_refused` (the anchor arm), `found_source_mints_estopped`, `reading_demand_met_by_retire`, `reading_demand_met_by_reread` — each with a zero-is-false-green gate; `FoundSourceBudgetRefused` deleted (F6 (b): the owed gap is never budget-refused) |
| new drive | — | per sweep, for each row of `readingMintedClasses`: one mint WITHOUT a standing reading (expects the refusal naming `verify`); one with the wrong outcome; one after blue's edit of the sentence (not standing); one with a standing reading (lands, at the cap too — `reading_mints_landed_at_cap`); and ONE URL READ AGAINST TWO SENTENCES, both `refutes`, by one lens in one sitting (admitted: `requireNotARetriedReading` refuses only the same sentence) → two `reading_debt`s, then `mint --quote S1` and `mint --quote S2` both land, the second debt standing after the first mint and the second mint admitted after it — `one_url_two_sentences` non-zero, and a gate on the url-keyed failure (one debt, or the second mint refused) |
| new drive (the debt) | — | a lens reads `refutes` and returns its envelope → refused, tallied `envelopes_refused_on_reading_debt`; the chair's `gate --as PASS` → refused naming the lens, `pass_refused_on_reading_debt`; discharged on a coin by its mint, its re-read, another lens's gap of the class, blue's retire, an estopped edit (`edit --accept` of a red fix on the found sentence), blue's deletion of the found sentence (F8 (a)), any seat's 29 on an uncited sentence a found source was read `absent` for, blue's cite of that sentence (F7 (a)), or left standing to the run's end (`UNVERIFIED` naming it) → `reading_debts_discharged` (map by discharge, each non-zero); a found `absent` on a CITED sentence → no debt, `found_absent_on_marked_no_debt` non-zero; a forced short sitting → `reading_debt_readied_again`; a lost quote by a non-estopped edit → the debt stands and is re-read, `reading_debt_lost_quote_reread` |

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
  `mint --class citation_figure_misattribution --quote S` refused naming the citation form. THE KEY IS
  THE SENTENCE: X reads ONE url U against S1 and S2, both `refutes`, in one sitting (admitted) →
  `PassBlockers` holds TWO `reading_debt`s for X; `mint --quote S1` lands; the S2 debt still stands;
  `mint --quote S2` lands; no debt. Then X reads another url U2 against S1 `supports` → the S1 reading
  that stands is U2's: a 32 `--quote S1` is refused and no S1 debt is owed; X's reading of U against S3
  leaves S1's and S2's standing readings as they were. *Fails if* a found-source 32 is refused, a
  found-source 25 admitted, or the key is the url (one debt; the second mint refused — the url-keyed
  design's shape).
- **T8** one-gap anchor arm: X's 32 about C open; Y (with its own `refutes` of C) mints 32 about C →
  `oneGapRefusal` naming X's gap; Y's gap of another class about C, with one of that class already open
  about C → refused (every class); Y's owed unread gap about C (Y's `unreachable` first since an edit)
  lands over X's open gap of the same class. *Fails if* duplicates land, the arm is class-limited, or the
  obliged gap is turned away.
- **T9** with X's `refutes` of C standing and no gap: `gate --as PASS` by the chair IS refused, by a
  `reading_debt` blocker naming X and C, and by no `failed_check`; after X's 32 about C graded low and
  closed `not a defect`, `gate --as PASS` is admitted (no other blocker); `show evidence --json`'s `counts`
  object has no `sources_refuted` key throughout. *Fails if* a found-against failed check survives, the debt
  does not hold the PASS, the mint does not discharge it, or the count survives.
- **T10** `close --disposition repaired` of X's 32 about C while X's `refutes` stands: refused by the
  DEMAND (`requireDemandMet`: "a reading of C by X … that backs the sentence"), not by a failed check;
  after X's backing re-read, or blue's `retire --anchor C`, admitted. *Fails if* `requireFailedCheckAnswered`
  still names found against, or the demand no longer reaches C through the 'about' arm.
- **T11** budget (F6 (b), "outside the budget while open"): at X's cap, a 32 about C on X's standing
  `refutes` is admitted and `mintStanding.Answering` counts it; after blue edits C's sentence the gap is
  STILL outside (open), and X's next ordinary mint is still refused by the cap; once the 32 closes, X's
  `Answering` no longer counts it. At the cap, X's 29 quoting an uncited sentence a found-source `absent`
  reading of X's locates to is admitted and counted; a 29 quoting a sentence X read `supports` is refused
  by the budget; a 29 by Y (no reading) on the `absent` sentence counts against Y. *Fails if* an owed gap
  is budget-refused, an unowed 29 is exempt, or the exemption ends before the gap closes.
- **T12** docket.md golden: the gap's "read for the sentence" line names the source via `ReadingBehind`
  (cited: C's url; found: U); `show evidence` row has no `carried_by`. `seatprobe` counts the gap.
- **T13** `TestTheEiffelReadingActsComeForward` on `run-archive/2026-11-04_eiffel-tower-height.tar.gz`
  at epoch 65, every expectation a row of §III.7's table: zero refusals; the four cited in-act Mints
  (165, 789, 792, 795) hold `about_kind` anchor, `about_ref` C-2dc96eab / C-a51c49f1 / C-a1560eb7 /
  C-486ed6b5 and NULL `location`, their Anchor rows (166, 790, 793, 796) still `sentence_of` the
  citation; the two found in-act Mints (180, 540) keep their `location`; every Verify row holds no
  `carried_by`; 798 is a Verify alone; the 6 class-32 gaps and the 25 are on the gap view with their
  closures. `ReadingBehind` names 167 for G-8dbfce13, 182 for G-3b2b7d93 and 542 for G-1d97d860 (never
  798), 952/953/955 for G-f66af91d/G-dd7b403e/G-de476176. `PassBlockers` at head holds NO
  `failed_check` on a refuted citation and NO `reading_debt`; the fold over the stream through event 845
  (`passBlockersOf` on `evs` cut there, `found` located as of 845) holds exactly one `reading_debt`
  {red-lens-evidence, C-3094c824, since 798, `refutes`}, and through 237 none for reading 182 (its gap
  180 is the sitting's). *Fails if* a member is dropped, a refusal appears, the translation misses a
  field, the as-of join or the sitting-open bound is lost (182 or 542 reads as a debt), the `supports`
  re-reads do not replace (791/794/797 read as debts), or 798's debt is missing through 845.
- **T14** a synthetic `[]OldEvent` stream (built in-test as `gapclass_test.go` and `cutfields_test.go`
  build theirs) at epoch 64 holding one cited in-act act, one found in-act act, one held reading and one
  plain 32 mint → migrates with zero refusals to the same shapes (the cited Mint about its anchor, the
  found one quoting). *Fails if* the translation depends on Eiffel's particulars.
- **T15** `TestEveryArchivedRunBehindThisBinaryIsRefusedUnmigrated` re-armed by the bump; `schema.sql`
  golden shows no `carried_by`, no `carrying_reading`, no `verify_carried_by`.
- **T16** `TestEveryReadingMintedClassIsDriven`, the `blockerKinds` coverage pattern
  (`passblockers_test.go:124–134`): a fixture table with one row per class — the refusing act, the
  admitting reading, the subject form(s) — iterated against `readingMintedClasses`; a class with no row
  fails naming it, and each row drives the refusal and the admitted mint through the CLI. Deleting a row
  of the Go table turns the class's T3/T6 red. *Fails if* a class is admitted with no check.
- **T17** G1–G3, G8: the help, enum help and catalogue goldens; `TestTheSeatPromptsNameNoVerb` and
  `TestNoPromptGrowsItsCommandCatalogue` green on the new `debate.js:883` and constitutions.
- **T26** the outcome clause (D12): Go — `TestACeilingReasonNamesOnlyAStandingReopenedAnchor` gains a row
  where X's `refutes` of C-1 stands with no gap → the CEILING reason is the term + "; on a reading debt:
  C-1" and names no other blocker; JS — `plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs:264` with the widened fixture expects the clause
  naming each kind of `STANDING_CLAUSES` once, in table order; `enginegen -check` clean after regenerate;
  `TestTheVocabularyRegionStatesEachSetAndThePlanAsDeclared` carries the constant. *Fails if* the Go
  clause or the JS clause names a kind the other does not, or the JS names a kind word of its own.
- **T18** the debt on every surface: X reads C `refutes`; `PassBlockers` holds one `reading_debt`
  {Owner X, Subject C, since N, outcome refutes}; X's work list holds it blocking;
  `EnvelopeRefusal(run, X)` names it; the chair's list holds it with `AnotherSeatClears`; the plan
  readies X ("to mint the gap or read again"); `lensOwesOf[X]` names it — read through `PlanDispatch`
  (`dispatch.go:311`) AND `passLensGateOfRun` (`retirement.go:445`), the two callers that fold
  `owedBlockersOf` without `passBlockersOf`; `gate --as PASS` refused on it. *Fails if* any surface misses it, or a surface words it from its own fold
  (`TestEveryPassBlockerKindAgreesOnEverySurface`'s new row).
- **T19** discharge, one row each, the blocker gone and X's envelope admitted: X's 32 about C; Y's 32
  about C already open before X read (X's mint then refused by the anchor arm, and nothing owed); a 32
  about C closed BEFORE X's sitting opened → the debt STANDS and X's mint lands; a 32 about C minted in
  X's sitting BEFORE X's reading → discharged (the sitting-open bound); X's later `supports` of C; blue's
  edit naming C (C unchecked); blue's `retire --anchor C`; found source: X's 32 `--quote S`; X's later
  reading of ANOTHER url for S `supports` → gone; X's later reading of U for S' (another sentence) → S's
  debt STANDS; blue's edit whose `new` holds no sentence for S (deleted outright, F8
  (a)) → gone, while an edit that only rewords S leaves it (T22); found `absent` (F7 (a)): X reads U
  `absent` for uncited S → a debt; Y's `mint --class unsourced_claim --quote S` (no reading of Y's) →
  gone; or blue's `cite` marking S → gone. *Fails if* any discharge is missing, a non-discharge clears it,
  or the deletion row is confused with the lost quote.
- **T20** once per cause, by `firstAfter(ends.through[owner], b.since)` over sat-through OPENINGS: X
  reads `refutes` and its own sitting closes with the debt standing → the next plan READIES X ("to mint
  the gap or read again"); X sits through that sitting and leaves the debt → the plan after does not
  ready X for it, `lensOwesOf` still names it, X is not retired at 3 barren sittings, `pass_permitted`
  stays false, the chair's list names X and `outcome` names C; X's sitting stopped at the call limit →
  the next plan readies X again. *Fails if* the reading's own sitting counts as sat through, the debt
  readies forever, or an unsat sitting ends it.
- **T21** two lenses, one citation: X and Y each read C `refutes`; two debts; X's 32 discharges both;
  Y's point is a finding. *Fails if* Y's debt survives X's gap or Y's duplicate lands.
- **T22** lost quote (F1): X reads U `refutes` for S; blue's ordinary edit changes S's quoted words →
  the debt stands with `lost` = the edit's event and the item naming the new sentence — one row where
  the edit's `old` is the whole sentence (⊇ the claim) and one where it is a phrase inside it (⊂ the
  claim), the archive's two shapes (238, 692); X's `mint --quote <new S>` refused (no standing reading of
  it); X reads U for the new sentence `supports` → gone; or `refutes` → a fresh debt since the new
  reading, discharged by X's 32 quoting the new sentence. The as-of join: X reads U `refutes` for S, mints
  32 `--quote S`, THEN blue rewords S → no debt (the gap joined as of `lost − 1`). *Fails if* the lost
  quote discharges the debt, the item cannot name either edit shape, or a lost reading whose gap was
  minted before the loss reads as a debt.
- **T23** estoppel discharge: as T22, but the edit is `edit --accept` of a red gap's fix on S (any red
  seat's, not X's) → the debt is gone with no act of X's; X re-reads the new sentence `refutes` → a
  fresh debt; X's `mint --quote <new S>` → the estoppel refusal naming the prescribing gap; with
  `--supersedes` that gap → lands and discharges. *Fails if* a verbatim red edit leaves the debt, or an
  ordinary edit discharges it, or the fresh debt has no route.
- **T24** found-`absent` debt (F2, F7 (a)): X reads U `absent` for UNCITED S → a `reading_debt` {Owner X,
  Subject U, outcome absent} holding the PASS and X's envelope; X's own `mint --class unsourced_claim
  --quote S` → gone, the mint admitted with no reading check and at X's cap (T11); a second `absent`
  reading by X of another found source for S → the same one 29 discharges both. X reads U `absent` for a
  sentence marked by citation C — the citation token at the END of the sentence's prose, the shape
  `claimcount.BareAnchorIDs` reads as unmarked — → NO debt, `PassBlockers` holds nothing for S, X's
  envelope admitted, the reading on `show evidence` `independent`; the same for a sentence marked by a
  computation token. X's later `supports` of another source for S → gone. *Fails if* `absent` on a found
  source creates no debt on an uncited sentence, creates one on a marked sentence (the wrong helper), 29
  does not discharge it, or 29 is refused for want of a reading.
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
| K5 | `AssemblyFailures` lost with the state deletion, silencing the assembly screen (a plausible zero) | M×H×L | T25's six rows, each a FAIL or PASS the other rows would not produce; §V check 9 deletes each arm and expects red |
| K6 | The lost-quote item names the wrong edit, or none (`old` containment misses: the claim spans two edits, or the report was re-ingested) | M×L×M | `lost` = 0 and the item names the reading alone; the debt stands and clears by the re-read either way (§III.4 residue; T22); the item is information, the predicate is the reading's standing |
| K7 | Estoppel discharge by ANY red seat's verbatim fix lets a lens's debt vanish on a sibling's prescription it never saw | L×M×L | the ruling's choice (red is one party, `estoppel.go:82–87`); the lens may still re-read and a finds-against re-read creates a fresh debt (T23) |
| K8 | A PR bumping to 65 lands first | M×L×L | take the next number; `schemagen` + `TestEveryArchivedRunBehindThisBinaryIsRefusedUnmigrated` |
| K9 | A half-state: `verify` still minting while `mint` already checks (or the reverse) across the phases | M×M×M | one PR (§V phases 1a–1e), each phase green; the PR opens after 1e |
| K10 | The fuzz (CI only) is the first to drive the two subject forms and the debt at scale; a seat-level defect shows after merge | M×M×M | a `--smoke` universe run reads the record (§V check 11); release-fuzz cadence: a full pass before the next logic PR |
| K11 | Archived 25/32 mints and G-b2ba7d66 come forward unjudged; archived finds-against readings with no gap read as standing debts | L×L×L | stated residue (§III.7); T13 asserts the debt set by count and subject; an archived record is read, not dispatched |
| K12 | `eventBlockersOf` gains a `found` input located from the rendered report: a read site that hands it `nil`, or a render that fails, silently builds no found-source debt (a plausible zero); and an arm placed in `passBlockersOf` instead of `eventBlockersOf` is invisible to the two direct `owedBlockersOf` callers, so the lens retires while owing | M×M×L | ONE function, `locatedReadingsAt(q, evs, win)`, at the three read sites, carried on `planReads`/`workReads`; a render failure is the read's ERROR, returned where every other read error goes (never a panic, never an empty fold); §V check 9 replaces `found` with `nil` at each of the FIVE hand-off sites (`passBlockerReadsAt`, `planReadsAt`, `workView`, `dispatch.go:311`, `retirement.go:445`) and expects the surface each serves red; T20's not-retired row pins the home |
| K13 | The found-`absent` debt's "unmarked" input (F7 (a)) is read from the report as it stands: a sentence blue cites after the reading drops the debt silently, a 29 is minted on it that nothing now obliges, or the mark predicate misses a mark (a plausible zero either way) | L×M×L | Blue's cite IS the discharge (§III.4; T19's cite row); the orphan 29 is an ordinary gap on its own grade; the mark predicate is `anchor.BackingMark`, the one `refuseMarked` uses, and T24's marked rows (a citation at the end of the prose, a computation) each expect no debt — `claimcount.BareAnchorIDs` in its place turns them red; §V check 9 deletes the predicate and expects T24 red |

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
goldens; (1e) fuzz. The PR opens after 1e.

**Validation loop.** Run in the worktree after `unset $(env | grep -o '^FEOV_[A-Z_]*')`; absolute paths;
each check names what re-arms it; read the FAIL COUNT, never an exit code.

1. `go -C scripts run ./protogen && git diff --exit-code -- '*record.pb.go' '*record.proto.sha256' '*decode_gen.go'` → no diff once regenerated output is committed. *Re-armed by:* any `record.proto` edit.
2. `go -C scripts run ./schemagen && grep -n eventSchema plugins/frank-exchange-of-views/requirements.json` → 65 both sides. *Re-armed by:* `requirements.json`, any event-shape change.
3. `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/cmdtreegen && go -C scripts run ./agentgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen && go -C scripts run ./vocabdoc && go -C scripts run ./agentgen -check && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check && git status --short` → exactly these change: `agents/*.md` (12), the two preloaded skills, `help_gen.go`, `annotations_gen.go`, `docs/vocabulary.md`, and `debate.js` in its `:373–451` region (`git diff -U0 -- '*debate.js'` shows three generated hunks — `:394` the party desc, `:415` the kind enum, the new `STANDING_CLAUSES` line — plus the authored `:883` and `:1094–1101`, and nothing else); both `-check` clean. *Re-armed by:* help .md, `terms.json`, agentgen sources, proto options, any `desc:` tag on `record.Plan`/`Party`/`PlanBlocker`, `BlockerKind`, `StandingClauses`.
4. `UPDATE_GOLDENS=1 go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/difftest/ ./internal/record/... ./internal/report/... ./internal/capture/... && git diff --stat` then READ every changed golden line (G1–G8, G-arch). *Re-armed by:* any agent-facing text, schema, archive.
5. `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/... 2>&1 | tee ~/.claude/scratch/mint-prereq/plan2/test.log | grep -c '^--- FAIL'` → 0; read the log's tail. *Re-armed by:* any Go change.
6. `go -C plugins/frank-exchange-of-views/tools test -count=1 -run 'TestTheEiffel|TestASyntheticInActRecord|TestEveryArchivedRun' ./internal/record/migrate/` → PASS (T13, T14, G-arch). *Re-armed by:* `migrate/`, the archive.
7. Zero-hit sweeps, each a gate — prints nothing, and is run on this tree BEFORE Part 1 to see it
   print the census's hits (a sweep that is empty before the change measures nothing):
   `grep -rnE 'carried_by|CarriedBy|carrier_minted|CarrierMinted|carrying_reading|verify_carried_by' --include=*.go --include=*.md --include=*.js --include=*.json --include=*.sql --include=*.golden plugins/frank-exchange-of-views scripts/agentgen | grep -v 'reserved'` (the `*.go` arm reaches `tools/releasegate/fuzz/` — `lensturn_test.go:231` is a hit today);
   `grep -rnE 'GapClassesOf|FoundSourceGapClassesOf|ReadingGapClasses|ReadingClassField|AppendReading|errReadingHeld|readingAct|pairedSentenceOf|CheckFoundAgainst|stateFoundAgainst|carrierTree|repeatReadingRefusal|readingSubject|FailingAnchors|CitationsFoundAgainst|SourcesRefuted|sources_refuted|forceFoundAgainst|fixFoundAgainst' --include=*.go --include=*.golden plugins/frank-exchange-of-views/tools`;
   `grep -n "'unchecked'\|'failed_check'\|'reading_debt'\|'open_inquiry'\|'unchecked_inquiry'" plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js | awk -F: '$1<373 || $1>451'` → nothing: no authored line of the engine names a `BlockerKind` word (today `:1099–1100`);
   `grep -rn 'desc:"' plugins/frank-exchange-of-views/tools/internal/record --include=*.go | grep -E 'finds against|found against|failed check'` → only tags that also say `reading debt` (read each; `dispatch.go:21` is the one hit today and must name both states and the debt after);
   `grep -rnE '(gap|reading|mint)[^.]{0,80}same act|same act[^.]{0,80}(gap|reading|mint)|found-against|found against|finds against' plugins/frank-exchange-of-views/tools/internal/cli/seat/help plugins/frank-exchange-of-views/docs plugins/frank-exchange-of-views/skills plugins/frank-exchange-of-views/agents scripts/agentgen/src plugins/frank-exchange-of-views/tools/internal/terms/terms.json | grep -v 'those changes twice\|supersed'` — the gap-or-reading-near-"same act" shape is the in-act mint's (26 hits today, `mint.md:17`, `verify.md`, `citations-and-bibliography.md:175,197`, the research-protocol SKILL, the outcome `means` as inlined), and the two exclusions are the two other senses that survive the change (`StaleRule`'s "the same act is not refused for those changes twice"; the supersede row's "closes it in the same act"); `edit.md:12`, `reproduce.md:19`, `seat-command-triggers.md:6,19,101,102` and `debate.js:824` fall outside the shape by construction (no gap/reading/mint inside the clause). Measured on this tree 2026-11-06: 139 hits, 100 of them the same-act shape (77 in the seven `red-lens-*.md` from the inlined manual and outcome `means`). "finds against" is admitted only as the option's name — `finds_against` with the underscore — so the sweep's bare phrase is zero.
   *Re-armed by:* any text or code edit.
8. Read each help page before and after, one command at a time: `go -C plugins/frank-exchange-of-views/tools run ./cmd/feov-record lens verify --help` (then `lens mint`, `lens close`, `blue edit`, `verdict`, `dispatch`, `show evidence`). *Re-armed by:* any help source.
9. Delete-the-row checks by hand, each reverted: remove 32's row from `readingMintedClasses` → T3 red; remove 25's not-the-work alternative → T6 red; remove the anchor arm's own-failed-check exemption → T8 row 3 red; remove `(finds_against) = true` from `ABSENT` → T6, T24 and the ledger's `absent` tests red; delete the `reading_debt` row from `blockerKinds` → T18 red (`TestEveryPassBlockerKindAgreesOnEverySurface`); delete the "or minted from the reading's sitting on" clause → T9's closed-low row red; delete the "by ANY seat" → T19 row 2 and T21 red; delete the `applied_verbatim` discharge → T23 red; delete the `lost` edit lookup → T22's item row red; delete the deleted-sentence discharge → T19's deletion row red; delete the unmarked predicate from the found-`absent` arm → T24's marked rows red; delete the 29 join in the budget → T11's 29 row red; delete the class arm of the budget → T11's first row red; delete `AssemblyFailures`' material-gap arm → T25 rows 1–3 red; delete its debt arm → T25 row 4 red; delete its computation arm → T25 row 6 red; replace `found` with `nil` at EACH hand-off site in turn — `passBlockerReadsAt` (serves `PassBlockers` `:784`, `passBlockersAt` `:803`) → T18's gate row red; `retirement.go:445` (`passLensGateOfRun`'s own `owedBlockersOf` call) → T18's `lensOwesOf` row red; `planReadsAt` (`dispatch.go:264` → `:542`) → T18's plan row red; `dispatch.go:311` (the plan's lens folds' own `owedBlockersOf` call) → T20's not-retired row red; `workView` (`viewjson.go:1459` → `sittingOf` `:435` lens, `:371` chair) → T18's work-list and chair's-list rows red (the blue `:354` and bench `:500` callers filter by kind and are unaffected — stated, not a finding); move `readingDebtsOf`'s arm from `eventBlockersOf` into `passBlockersOf` → T20's not-retired row and T18's `lensOwesOf` row red; delete the `reading_debt` row of `StandingClauses` and regenerate → T26's Go and JS rows red; key the found-source debt on the url instead of the sentence → T7's two-sentence rows red; replace the sitting-open bound with the reading's id → T13's "through 237" and T19's before-the-reading rows red; replace `BackingMark` with `BareAnchorIDs` → T24's end-of-prose citation row red; replace the either-way containment with one direction → T22's ⊂ row red; delete the as-of join → T13's 182/542 rows and T22's last row red; delete the Verify-column drop in migrate → T13 red; delete the about-anchor translation → T13's four cited Mints red. A row that stays green is a finding. *Re-armed by:* any change to the table, the arm, the option, the translation.
10. `go -C scripts run ./check 2>&1 | tee ~/.claude/scratch/mint-prereq/plan2/check.log | tail -40` → read the fail count in the log. *Re-armed by:* every push.
11. `scripts/universe.sh build`, then `grep -rn 'standing reading' <the universe's installed plugin help>` proves the universe carries the change; one `--smoke` run (one lane, haiku both tiers); read its record: the `verify` table has no gap column; every 32/25 mint has `ReadingBehind`; no plain 32 mint by a seat with no reading; every finds-against reading has a 32/25 gap from its sitting on or a later non-finds-against reading, or the outcome names it. *Re-armed by:* Part 1's merge candidate.
12. **CI only — release fuzz** (`FEOV_RELEASE_GATE=1`, the sharded job in `.github/workflows/hooks.yml`, scheduled and on tag; never run locally by any door): `reading_mints_refused` non-zero for both classes, `reading_mints_landed` non-zero for both subject forms and at the cap, `second_reading_mints_refused`, `found_source_mints_estopped`, `reading_demand_met_by_retire`, `reading_demand_met_by_reread`, `envelopes_refused_on_reading_debt`, `pass_refused_on_reading_debt`, every key of `reading_debts_discharged`, `reading_debt_readied_again`, `reading_debt_lost_quote_reread`, `found_absent_on_marked_no_debt` non-zero. A logic PR waits for one full pass (cadence ruling 2026-10-29). *Re-armed by:* Part 1's merge candidate.

## VI. Forks

**RULED (gblock, 2026-11-06; option wording the lead agent's):**

| Fork | Question | Ruling | Home |
|---|---|---|---|
| F1 | a found-source reading whose quote stops locating after blue's edit | **Debt stands** (not the recommended "owes nothing"); refined: unless the edit took red's suggested text verbatim, then estoppel discharges it. The mint half of (a) stands: no `--quote` mint until the lens reads the new sentence | §III.4 lost quote; T22, T23 |
| F2 | `absent` on a FOUND source | **Debt like refutes** (not the recommended "no debt"); the discharging class and the cited case are F7 | §III.4; T24 |
| F3 | not-the-work as `finds_against` for 25 | **(a) two predicates**: the ledger's `finds_against` is refutes/absent on a copy that read the work; 25's mint check also admits the seat's standing not-the-work reading | §III.3 table; T6 |
| F4 | estoppel on a found-source 32 | **(a) no exemption**; reverses the 2026-10-31 exemption for this case | D10; §III.3 |
| F5 | the one-gap anchor arm's scope | **(a) every class** about an anchor, except the gap a seat owes for its own unread/unsound check; a second lens's debt is discharged by any seat's gap of the class about the citation | D5; T8, T21 |
| F6 | the owed 32/25 gap and the mint budget | **(b) outside the budget while open**, as the owed unread/unsound gap is | D7; T11 |
| §VII 1′ (first draft) | the debt's home: a `BlockerKind` or a parallel fold | **a `BlockerKind`** on the one list; the chair sees it, the lens clears it | D4′; §III.4 |
| §VII 1 (first draft) | the assembly screen on a refuted citation | **screens the gap**: open material 32/25, or a found-wrong reading with no gap | D11; T25 |
| F7 (from F2) | which gap discharges the debt a found source read `absent` creates | **(a)**: where the sentence carries NO citation, a 29 `unsourced_claim` by any seat on the sentence discharges it (the lens looked and found no backing; 29 is what carries that to blue); where the sentence ALREADY carries a citation, a found page silent on it creates NO debt. Rejected: (b) clears only by the re-read (a debt with no gap route); (c) no debt (reverses F2) | §III.4 found-`absent`; T24; K13 |
| F8 (from F1's residue) | the found-source debt when blue deletes the sentence outright | **(a)**: the deletion discharges it, as blue's edit does on the cited row; the lens may read again if the claim reappears elsewhere. Rejected: (b) stands until a reading of U for another sentence (a hollow act or an UNVERIFIED end over a sentence the report no longer carries) | §III.4 deleted sentence; T19 |

No fork is open.

## VII. Inconsistencies between the ruling and the code, reported, not resolved

1. "Keyed on the subject the mint already names": for a cited source that subject is the CITATION
   (`--about-kind anchor`), not the sentence — a quote gap reaches no anchor once `carrying_reading` goes
   (`views.go:493–495`), so its `source` demand could be met only by a new cite. Kept as designed (D3);
   `ABOUT_KIND_ANCHOR`'s `means` (`record.proto:1179`), `mint --about-kind`'s usage and `mint.md` say so.
2. The class-to-outcome rule has no declared home (the option mechanism was dropped): a Go table
   (`readingMintedClasses`, §III.3) beside 31's hand check (`record.go:1931`) — two hand checks of two kinds.
3. "Enforced by its own envelope" and "an orphaned reading blocks the PASS" are ONE mechanism: every owed
   blocker the lens owns is a blocking work item (`sitting.go:438–442`), and the envelope gate reads
   `Blocks`; "orphaned" is no state the code tracks.
4. Rulings 1, 4 and 8 (the MINT discharges; a closed low gap does not fail the run; any seat's OPEN gap
   discharges a second lens) are reconciled as "open now, or minted from the reading's sitting on" — a
   predicate, not a stored fact; the one unreached case (a gap closed BEFORE the lens's sitting opened) is
   decided as NOT discharging. The bound is the sitting's OPENING, else the reading's id: `max(sitting
   open, events.id)` collapses to the reading's id and would read every archived in-act gap (Mint before
   Verify) as an undischarged debt (T13).
5. "No threading of a report through callers": `passBlockersOf` is a fold with no querier, so the report
   is rendered ONCE at each of the three read sites inside their snapshot and the fold receives the
   located sentences as it receives `gaps` and `fresh` — the report string crosses no call. Cost per fold:
   one checkpoint read + ≤7 ops, plus one per LOST reading (the as-of join, RULED built). A fold over a
   report rendered outside the snapshot would read a different record than its `evs` (`readSnapshot`,
   `narrowedview.go:111`).
6. The ruling's translation names "6 quote gaps minted by a cited reading"; the rows hold FOUR of a cited
   reading and TWO of a found source (§III.7). The translation applies where the Verify names an anchor;
   the two found-source Mints are already the design's quote mints. The four translated gaps keep the
   token their in-act mint placed — a pair the live tool never writes; residue.
7. The audit read `firstAfter(ends.through[owner], b.since)` as a sitting END; `satThroughOf` stores
   OPENINGS (`dispatch.go:944`), so the reading's own sitting never qualifies and the orphan lens IS
   readied once (T20). No fork.
8. Under the sentence key, a later found-source reading of ANOTHER source that backs the same sentence
   discharges the debt, and a later reading of the SAME url against another sentence does not (T7, T19).
9. F7 (a) says "citation"; the code's one predicate for a sentence's evidence is `anchor.Backs`
   (`anchor/anchor.go:316`) — a citation OR a computation — and 29's `means` (`record.proto:472`) is "no
   citation or computation of its own". Read as "mark" (either kind); T24's computation row pins it. Also:
   nothing judges a 29 MINT by the sentence's marks (`refuseMarked`, `verify.go:584–616`, guards the
   BACKING READING), so the discharge asks nothing of the 29 beyond its quote's sentence.
10. The audit's "`show evidence`'s per-citation state word flips to holding" names a word the JSON does not
    carry: `EvidenceSourceJSON` (`evidenceview.go:65–124`) has no ledger-state field, and the ledger's
    readers are listed in §II — each unaffected by the state's deletion or deleted with it. No golden line
    to read; stated so the claim is seen checked.
11. Found by the gap-class sweep, outside the ruling: `verdict.go`'s `anchorsStanding` names the two
    inquiry kinds and `debate.js`'s does not — the record's outcome reason and the engine's termination
    reason already disagree on a run that ends over an open inquiry. D12's table fixes the class; the
    `plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs` fixture gains the inquiry rows with the debt's (T26). Not a fork: the Go clause is the
    record's own derived reason and the JS mirrors it by comment ("as the record's own derived reason does").
12. The three rulings of 2026-11-06 folded in this revision (`sources_refuted` dropped; the lost-quote lookup
    and the as-of join built) are relayed by the lead agent and are not in #1727's comments; attributed to
    gblock through the lead, per the option-text rule.
