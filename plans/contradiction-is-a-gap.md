# A source reading against the report is a gap — the contradiction is deleted

> **F13 revision (2026-11-03), written on the implementation branch at `6102f2b09` (phase 1 built).**
> gblock ruled `lens corroborate` dropped (2026-11-03, §III.2d); the two sessions agreed the design, and
> gblock asked for a TARGETED one-round audit of the changed rows before phase 2 resumes. Separately,
> the peer session bridge-cse-01s7d2sa1pneo9oqzevmlwdg-4f relayed (2026-11-03) gblock's ruling that
> findings fold into gaps under a separate plan, issue #1695: this plan keeps only the finding rows that
> DELETE something (the contradiction about-kind, the raise, its credit) and points at #1695 for the rest
> (§III.16). Every changed row or section carries `<!-- CHANGED F13 -->`.
>
> **Changed sections:** status block; §I (objective, criteria 1, 4); §II (record shapes, mechanisms,
> materiality, constraints); §III.0 (rows 1a, 1b, 1c, 1e, corroboration rows, new rows 11–12); §III.1
> (facet, `ReadingGapClasses`, Verify fields, `keyFields`, migrate lists); §III.2 (flags, class field,
> required marker, `ReasonIs`, footnote bullet, guard table, field table, estoppel, result words, repeat,
> crash-retry); §III.2a (subject, mechanism, sweep table, `verifiedFix`); §III.2b (property, sweep table
> rows); §III.2c; **§III.2d (NEW: F13's design, deletions, phase-1 code to change)**; §III.3 (last
> bullet); §III.6 (finding's about-kind); §III.7 (help pages, terms, agentgen sources); §III.8 (`seatprobe`); §III.9 (survey, items 1,
> 3, NEW item 7, residue, tests); §III.10 (deleted and rewritten tests); §III.11 (reading drive, gates);
> §III.12 (NEW F13 census; tests; lane table); §III.13; §III.14 (phases 2–5); §III.15 (tests 1, 9–13,
> 15–35, NEW 9a–9d, 13a, 35a–35d); §III.16 (seam: the corroboration arm, #1695); §IV (R5, NEW R21–R23); §V (4, 5,
> 6, 8); Forks (F5, GAP 2, F12, NEW F13, NEW #1695 row).
>
> **FINDING-DEPENDENT rows** (each tagged `<!-- FINDING-DEPENDENT -->`). KEPT, because each deletes
> something or states that the reading writes none: §III.2 guard-table `--supersedes`/`--found-by` row
> and field-table `found_by` row; §III.2c's F12 resolution; §III.6 finding's about-kind; §III.7
> `finding.md`/`finding.go`; §III.9 items 1 and 3 (raises dropped, `found_by` entries removed), the raise
> residue line, test 9 and partner f; §III.11 `open_by_found_by` (F7, reads only); Forks F7, F12, #1695.
> MOVED to #1695 (a one-line pointer in §III.16): the reading result's "add what it lacks as a finding
> about G", the repeat refusal's "a point the gap lacks is a finding about G" (phase-1 code, §III.2d),
> and §III.9's former translation of each raise into a finding about its citation.
>
> **F13b revision (2026-11-03): the targeted round's verdict discharged; no further audit (gblock ruled
> ONE round).** <!-- CHANGED F13b --> The verdict (`~/.claude/scratch/plan1647/roundF13-verdict.md`) failed two
> rows and left notes and two forks. gblock ruled the forks and the key, 2026-11-03 (the options he picked;
> their wording is the lead session's, not his): (1) a found-source reading keys on its SOURCE and its
> SENTENCE — readings by one seat in one sitting that differ in url or in sentence are each recorded, an
> identical retry is refused naming the stored reading; (2) a gap a `verify --url` reading mints COUNTS
> against the lens's mint budget, and a `verify --anchor` reading's gap stays outside it (F6 amended);
> (3) F13 stays in this PR. Relayed by the #1657 session the same day: gblock picked "drop them, state it"
> for the 15 lens-placed `C-` anchors (§III.9 item 7's basis), and two 1a′ seam picks (§III.16). Every
> row this revision touches carries `<!-- CHANGED F13b -->`.
>
> **Changed sections (F13b):** this block; §II (materiality); §III.0 (rows 1e, F6's budget row, 12);
> §III.1 (`unsourced_claim` means and distinguisher, the backing outcomes' means, `keyFields`); §III.2
> (crash-retry and the key, the repeat rule, the `--key` guard row, `Supplies("claim")`, the requirement
> order); §III.2a (`unplacedLocationAt`, `consistency.go` rows); §III.2d (the key, materiality, the
> guard's order, the scope, the phase-1 rework list, the deletions table, tests); §III.5 (F6 amended);
> §III.9 (item 7: the ruling, the label map's source, list removal, every report-text field, the
> migrate budget measure; residue wording); §III.12 (identifier count; the label readers);
> §III.10 (`TestAReadingsGapTakesNoMint`, `TestARetriedNegativeReadingIsRefusedNamingItsGap`); §III.11
> (the `--url` drive's budget and key oracle, two tally fields); §III.14 (phase 2's row);
> §III.15 (9b, 9c, NEW 9e, 13a, NEW 13b, 20, 33, 35, 35b, 35c, NEW 35e, NEW 35f); §III.16 (O1's consequence
> for 1a′, E1′, the two 1a′ picks); §IV (R2, R21, R23, NEW R24); §V (5.v, 5.oo, 5.pp, NEW 5.vv–5.zz, 9);
> Forks (F6, F13, O1 closed, NEW key and scope rows). Also §I (the budget clause), §III.2's
> budget-exemption bullet, and §III.7 (`verify.md`: the budget and the key).

> **Status: APPROVED by gblock, 2026-11-01. The audit loop ended at round 8 (gblock, 2026-11-01: "Revise
> once, then implement"); that revision was the last of the loop. The F13 revision above (2026-11-03) is
> audited by ONE targeted round over its CHANGED rows before phase 2 (gblock) <!-- CHANGED F13 -->. <!-- CHANGED F13b --> That round ran and failed (two gaps); the F13b revision discharges it, and no further round runs (gblock: ONE round); phase 2 resumes on F13b. Audit round 8's three gaps,
> its F12 ruling and its notes are discharged below and carried as REQUIRED TESTS (§III.15). Independent
> review of the implementation is required before merge.**

Written against `origin/main` **`9ee7ad95e`** (EventSchema 54: #1677, a hearing ends its turn and the
engine reads the record off the hook, took 53; #1678, the #1677 smoke's fixes, took 54). `origin/main`
has since moved to `a7485a7ee` by docs only (the repo guide moved to `AGENTS.md`; plan docs; one
archaeology word): no FEOV code moved and the epoch is still 54, so no reference below moves with it.
What #1677 changed that this plan reads: petitions and positions are no longer dispatch occasions
(`Party.Occasions`, `Dispatch.occasions` deleted, `Dispatch` field 4 reserved), the bench hears a filing
at any barrier where no seat runs, `debate.js` learns state only from `record_read` (written by a
PreToolUse hook on every seat's envelope) while the chair's plan is still the chair's envelope
(`debate.js:953-957`), and the client floor is Claude Code 2.1.305. #1678 added `Verify.sha256 = 20`
(the copy a labelled corroboration read). No disposition below rested on the old dispatch, the old
bench hearing or an envelope read by `debate.js` (§III.12, "What #1677 and #1678 moved"). The references
of the round-7 text (read at `0c57758e8`) were carried by a `git diff -U0` line map
(`~/.claude/scratch/plan1647-r10/mapref2.py`, its attribution of a bare `:N` fixed to the nearest
preceding file); every reference the map could not carry unchanged — a line those PRs rewrote — was
re-read by hand, and the word census, the phrase sweep and every sweep of `sweeps.sh` were re-run at
`9ee7ad95e` (raw output beside the script, `new-*.txt`, `census-new.txt`, `phrase-new.txt`; the field
sweep `new-mintfields.txt` and the owed-ness sweep `new-fc3.txt` were carried by the map, their files
unchanged but for line numbers). Two open PRs touch files this plan cites with no overlap in the concept
— #1672 (the record writes the plan into the chair's envelope) and #1676 (an exchange ends on the lens's
re-audit); nothing here is derived against an unmerged PR. #1644 renamed proof to computation: a `P-`
id is a `K-` id, `proof_id` is `computation_id`, and every sentence below that speaks of a proof's
failed check speaks of a computation's. Implements gblock's ruling of 2026-10-30; every fork is RULED
(gblock, 2026-10-30, 2026-10-31 and 2026-11-01, §Forks). This plan takes **EventSchema 55**. Field
numbers, the epoch and every line reference are re-derived at the implementation rebase; none is
carried from this document.

> **Owner ruling (gblock, 2026-10-30, his words):** "why isn't the contradiction recorded as a gap in
> the first place like red would for any other problem it finds? We're letting it mint a weird
> placeholder." And: "Our standard behaviour is reject a duplicate, not dedup — why are you treating
> contradictions as unique?" He approved planning the redesign: a `refutes` or `absent` source reading
> is recorded as a gap of a real class, like any other problem red finds, and the contradiction
> concept is deleted.

## I. Summary & Goals

**Objective.** <!-- CHANGED F13 --> This plan also deletes `lens corroborate` (F13, §III.2d): `verify`
is the one reading verb — `--anchor` for a citation blue made, `--url` with `--quote` for a source the
lens found — and no reading places a citation. A reading of a source the lens found mints its gap in
the same act: `unsourced_claim` where the source backs an uncited sentence, `claim_contradicts_source`
where it refutes the sentence. Below, "the reading verbs" is `verify` in its two forms.

Today a lens's `verify`/`corroborate` reading whose outcome is `refutes` or `absent`
opens a **contradiction** (`record/citationid.go:522` `contradictionsOf`): a record object with its own
id (the reading's `events.id`) that a finding must *raise* (`--about-kind contradiction --about <id>`)
and that only the gap minted from that finding (`found_by`) *answers*. Until then it is a PASS blocker
of its own kind (`passblockers.go:35`), a work-list debt of the reader (`sitting.go:374-439`), a
readiness reason (`dispatch.go:535-541`), and it is exempt from the one-gap-per-class rule
(`onegapperclass.go:68`) and from the mint budget (`mintbudget.go:301-305`). A reading after the raise
opens a *new* contradiction (`citationid.go:525-529`), so N readings of one sentence yield N
contradictions, each owed its own exempt gap: in release-fuzz run 40872228665 (PR #1642's finding),
22 of 45 mints per run answered a contradiction and all 37 drawn runs ended CEILING. m32 shows the
same shape on real seats: four lenses refuted one claim on `C-a36dce3f`, four findings raised four
contradictions, one gap answered one of them.

Throughout this plan, a **negative reading** is a reading that finds against: it reads the work
(`VerifyReads` — not `unreachable`, not any outcome of a copy the tool stamped `not_the_work`) and its
outcome is `refutes` or `absent` (`record.FindsAgainst`, §III.1). A reading that reads nothing is never
a negative reading, whatever its `--as`, and mints nothing (§III.2b).

After this change a negative reading **is** the gap: the reading act mints a gap of a real class on the
sentence. Its Mint meets every write-path rule a mint meets (class, subject, grades, check kind, one-gap
rule, report voice; the budget — outside it for a citation reading's gap through F6's act-paired
exemption, counted for a found-source reading's gap, F6 amended <!-- CHANGED F13b -->) and, of `lens mint`'s verb-level
checks, those §III.2's guard table marks MET — the verified fix among them; it is EXEMPT from estoppel
(`requireNotEstopped`; ruled F10: a source's refutation is new evidence, not red relitigating its own
prescription). It is filled as a plain mint is (F11, ruled): the reading's `--reason` is the gap's
argument (`mint_reason`), `fix_basis` defaults to `proposed`, and every field `lens mint` sets on its
body has a stated disposition (§III.2's field table). A second negative
reading of a sentence that already holds an open gap of that class is the ordinary duplicate: its mint
is refused by the one-gap rule, the reading is kept as evidence and names the standing gap. Blue answers
it like any gap. The contradiction — its about-kind, raise, fold, blocker kind, `unanswered_contradictions`,
work items, readiness reasons, exemptions, term, help, constitution and prompt text, goldens and fuzz
drives — is deleted, with no alias and no "formerly" prose.

**Success criteria (each a §V check).**

1. <!-- CHANGED F13 --> A `verify --anchor` with `--as refutes` or `--as absent`, or a `verify --url` with
   `--as refutes` (or a backing outcome on a sentence no evidence anchor marks, §III.2d), of a copy that
   is the work writes the reading AND an open gap in one transaction; no reading writes a label, a
   footnote or a citation anchor; the reading's JSON result names the gap
   (`gap.id`, `gap.minted: true`). A `verify`'s gap stands at the end of the sentence its `--anchor`
   token sits in, placed by that token's identity, also where the report holds the sentence's words
   twice, and blue's `--accept` of its verified fix edits that occurrence. A reading that reads
   nothing — `unreachable`, or ANY outcome of a copy the tool stamped `not_the_work`, a backing one
   included — writes no gap, no footnote and no `carried_by`, and leaves a `verify`'s citation on the
   `unread` failed check; its gap flags are refused. Every reading, `verify`'s included, carries the
   `source_completeness` of the copy it read. A reading's gap carries its argument and its `fix_basis`
   wherever a plain mint's appear: `show board --json`, the blue work item, the gap view, and docket.md's
   open-gap block ("minted because" and `fix_basis`) and closure index ("minted because") — never
   report.md, which prints neither for any gap. The gap is linked to its source (F12): each reading's
   row in `show evidence` names the gap that carries it (`carried_by`), and docket.md's block for the gap
   names the source each carrying reading read.
2. A second negative reading of the same sentence with the same class, while that gap is open, writes the
   reading, writes no gap, and names the standing gap (`gap.minted: false`). Gaps per (class, sentence)
   open at once: **≤ 1**, asserted over every fuzz run by the sweep tally.
3. The identifiers grep of §V.4 returns **0** lines outside migrate's archived-word tables; the word grep
   returns only lines §III.12 lists as KEEP (ordinary English, the two self-contradiction classes,
   historical plans).
4. <!-- CHANGED F13 --> `migrate` brings every archive forward with **0** refusals (`TestEveryArchivedRunMigrates`), and the
   migrated m23, m30, m32, sqlite-schema-source and the six archives holding a red-placed citation
   (quadratic-formula, b7, b9, m27, m29, m35) hold the dispositions §III.9 states; a record written at
   this epoch — a capped lens's refuting reading in it — migrates forward with **0** refusals.
5. `go -C scripts run ./check`: 0 failed. PR CI's release fuzz (sharded, CI only): 0 failed, with the
   new gates of §III.11 non-zero, and no gap minted by a reading act per run exceeding one per (class,
   sentence). **This plan removes the contradiction mint loop; it does NOT by itself end the drawn
   runs' CEILING, and a drawn run ending CEILING is not a failure of this plan.** #1642's own run
   (40879063345, tally artifacts at `~/.claude/scratch/ceiling/a79063345/`, read for this revision): 37 of
   40 drawn runs ended CEILING; at their end 32 of the 37 hold a `reopened` blocker or an open gap
   carrying an unsound judgement (`unsound_proof` 118 of 236 open gaps; `reopened` in 21 runs), and 35 had
   readied a lens on a reopened or unsound anchor; only 3 of the 21 runs holding an open
   contradiction-answering gap hold neither. (The brief's "36 of 37" is not reproduced by any of these
   readings; the measured figures stand.) What this plan owes the fuzz is the DIAGNOSTIC: #1642's four
   tally fields carried and re-keyed (F7, §III.11), so the next CEILING investigation reads what holds
   each run off the tally — the run that lands this plan reports, from its PR CI tally, the per-run
   `final_blockers` and `open_by_found_by` counts beside #1642's above.
6. One haiku smoke (`--smoke` configuration) in which a lens records a negative reading ends with that
   reading's gap on the board — or the run is reported NOT MEASURED on that path and the real-data drive
   of §V.9 stands in for it.

**Value case.** The naive alternative is to keep the contradiction and patch its loop (PR #1642: a fuzz
stock that reads `supports` once a lens is past it). That patches the harness, not the engine: real seats
have no stock, and m32 already shows the multiplication. It also keeps a second lifecycle (open → raised →
answered) parallel to the gap's, with its own id space, its own exemptions from the two rules that bound
gaps, and ~40 files of prose explaining it. Deleting it beats the patch on three counts: (a) the loop is
removed by the rule that already governs every other gap (one gap per class per sentence), not by a
special case; (b) the PASS gate loses a blocker kind and a lens loses a debt kind, so the plan, work list
and help pages shrink; (c) a seat records a refuting source in one act instead of three (read → raise →
mint). It pays for its audit because it is a deletion of a concept with one replacement mechanism that
is already proven elsewhere (the failed check, §III.3).

## II. Technical Context

- Go 1.27 (`GOTOOLCHAIN=local` in CI), FEOV tools module `plugins/frank-exchange-of-views/tools`; dev
  tools in `scripts/` (one module). SQLite via modernc; the record is protobuf-declared
  (`internal/record/recordpb/record.proto`) with SQL generated by `scripts/protogen` and the epoch by
  `scripts/schemagen` from `plugins/frank-exchange-of-views/requirements.json` (`"eventSchema": 54`).
- **Epoch.** `origin/main` is at EventSchema 54 (#1677 took 53, #1678 took 54). This plan takes
  **EventSchema 55**, and re-derives the number from `origin/main` at the implementation rebase; the
  number is never carried from this document.
- **The `lane` role (#1663).** Lanes perform no reading, mint no gap and hold no failed check (§III.12,
  "The `lane` role"); their agent definition regenerates because it inlines the tool's `manual`.
- **#1652's `--grounds` and this plan's gaps.** A docket motion on a gap the bench ruled closed names
  new grounds (`record/grounds.go`, `requireNewGrounds`): an act recorded after the ruling and about THAT
  gap — a finding about it, a gap minted about it or superseding it, a computation answering it, a
  citation placed by an edit answering it — or a quote of text such an edit wrote. A reading's gap is an
  ordinary Mint with an ordinary `G-` id, minted on its sentence and naming no gap it supersedes (the
  reading verbs take no `--supersedes`), so it is grounds for another gap's reconsideration only through
  a later act that names it under that rule, as any plain mint of a sentence is. migrate creates no gap
  for an archived refutation: an archived answering gap (m30 `G-dc877265`, m32 `G-f417112d`) is carried
  with its own `G-` id, and an archived reading no gap answered gets none — the tool does not invent a
  gap the seat never minted (its grades would be fabricated).
- **Record shapes touched** (contract): `AboutKind` loses `ABOUT_KIND_CONTRADICTION = 4` (number and name
  `reserved`) and the `gap_may_name` facet (every remaining kind admits a gap, so the facet is dead);
  `GapClass` gains one value (F1); `SourceOutcome` gains a facet naming the classes each negative
  outcome admits (F1) and a second, `found_source_gap_classes`, naming the classes a reading of a source
  the lens found admits (F13); `Verify` gains `carried_by` (references `mint.gap_id`) and
  `carrier_minted`, and loses `label`, `sha256` and `work_status`, the footnote's fields (F13, numbers
  reserved) <!-- CHANGED F13 -->; `Anchor` gains
  `sentence_of` (a marker placed at the end of the sentence another anchor's token sits in, found by
  that token's identity, §III.2a); `Mint` is unchanged. `show evidence`'s verification row gains
  `carried_by` (F12, §III.2c), and its `Refuted()` becomes `FindsAgainst()` (GAP 2).
- **Existing mechanisms reused, not invented:** the one-gap rule (`onegapperclass.go`, eb0c5a99f), the
  failed-check state and its close gate (`failedcheck.go`, `requireFailedCheckAnswered`), the class's
  `material_default` stamped at write (`Mint.class_material`, `recordpb/gapclass.go:18`), the
  placed-append that already writes two events in one act (`seat.AppendPlaced`, used today by
  corroborate's footnote, which F13 deletes) <!-- CHANGED F13 -->, the one-gap rule's `heldGap` (`onegapperclass.go:100`, the open gap it names), the
  exact-span edit that replay locates byte for byte (`BlueEdit.exact_span`, `reportproj/render.go:89-97`),
  and replay-time reference translation in migrate (`migrate/replay.go:391-402`).
- **Materiality** is the class's: `material_default` on the `GapClass` value, stamped by the write path
  onto `Mint.class_material`, never set by a seat. Every class this plan admits for a reading of a
  CITATION (`claim_contradicts_source`, `citation_figure_misattribution`,
  `within_source_condition_misattribution`) is `CLASS_MATERIAL_ALWAYS`, so such a gap holds the PASS
  whatever it is graded; grades still order it (mass). Of the two classes a reading of a source the lens
  found admits (F13), `claim_contradicts_source` is `always` and `unsourced_claim` is
  `CLASS_MATERIAL_BY_GRADE` (`record.proto:461-463`): a backing found-source reading's gap holds the PASS
  where its grades make it material, exactly as every `unsourced_claim` gap does. This plan does not
  change that class's default: materiality is the class's, and the ruled option names `unsourced_claim`
  (O1 closed, §III.2d). No seat picks materiality. <!-- CHANGED F13 --> <!-- CHANGED F13b -->
- **Constraints.** No backwards compatibility (CLAUDE.md change discipline): no alias for the about-kind,
  no tolerance path for an archived `contradiction` word in live code, no "formerly" prose. Archived
  records are brought forward by one `migrate` translation (F13 adds a second: §III.9 item 7). No alias
  for `corroborate` and no "formerly" prose about it. <!-- CHANGED F13 --> The release gate
  (`FEOV_RELEASE_GATE=1`) runs in PR CI only — never locally (release fuzz cadence ruling, 2026-10-29).
- **Brief claims checked at `8236e3a17` and found wrong or partial.**
  - "GapClass has no class for the cited source refutes / does not contain the claim": `citation_figure_misattribution`'s `(means)` already reads "the figure, statement or support is absent from or contradicted by its text". F1 (ruled) splits it: `refutes` gets the new class, and `citation_figure_misattribution` is narrowed to `absent`.
  - "#1522 made a refutes|absent reading a contradiction": the contradiction dates from gblock's 2026-10-06 rule (`citationid.go:493`); #1522 (a9ab33f49) made the finding insufficient, so only the gap minted from it answers.
  - "m19 contains contradictions": m19 (epoch 24) holds one negative reading (ev 114, `C-706fda63`), but the `contradiction` about-kind did not exist before epoch 30/31, so nothing in it was ever raised. Only m30 and m32 hold raises.
  - "N sources refuting one sentence yield N gaps": that holds in the release fuzz. No archive shows it. m32 shows N raises with one gap: four readings, four raising findings, one answering gap, three raised and never answered.
- **Out of scope.** `unreachable` and not-the-work readings keep the failed check's *owed-gap* shape
  (the lens owes a gap about the anchor): nothing was read, so there is no claim to mint against at
  the act. Computations judged `unsound` are unchanged. The `cross_section_contradiction` and
  `claim_contradicts_own_record` classes are different concepts and stay.

## III. Proposed Changes

### III.0 The design, decided

| Question (task item) | Decision | Fork |
|---|---|---|
| 1. One class or two | A new `claim_contradicts_source` for `refutes`; the existing `citation_figure_misattribution` narrowed to `absent`; the lens picks among the classes its outcome allows | **F1** RULED |
| 1a. Which readings mint | A reading that FINDS AGAINST: it reads the work (`VerifyReads`: not `unreachable`, not a `not_the_work` copy) AND its outcome has `gap_classes` (a citation reading), or — for a reading of a source the lens found (`verify --url`) — it reads the work and its outcome has `found_source_gap_classes` (F13: backing outcomes → `unsourced_claim`, `refutes` → `claim_contradicts_source`, `absent` none, F5). One predicate, `record.ReadingGapClasses`, decided after the tool stamps the copy's completeness; every site in §III.2b's sweep calls it, none decides by outcome alone <!-- CHANGED F13 --> | F13 |
| 1a′. The completeness stamp | Every reading stamps `source_completeness` from the run's cache entry for the url it read — the cited url for `verify --anchor`, `--url` for `verify --url` (today only the `--url` form, then `corroborate`, stamps it: `verify.go:495`, guarded by `body.GetUrl()`), so a `verify` of a `not_the_work` copy counts as a reading. EVERY reading of a `not_the_work` copy is unread — mints no gap, holds the PASS through the citation's `unread` check where it read one and owes that check's gap — a BACKING one included <!-- CHANGED F13 --> | **GAP 1** RULED |
| 1b. Who names class, quote, problem | Quote = the reading's claim (the tool reads it by the anchor for `verify --anchor`; `--quote` for `verify --url`) <!-- CHANGED F13 -->. Class = the lens's `--class`, refused outside the outcome's admitted set. Problem, check, check-kind, severity, likelihood, impact = the lens's flags on the reading act, required on a reading that finds against (1a) and refused on any other; fix, new text and complexity optional as on `mint`, and refused on any other | **F2** RULED |
| 1b′. The gap's argument and its other fields | The reading act takes `--reason` as plain mint does: its text is the gap's `mint_reason` as well as the reading's `Verify.text`, and required when the reading finds against. `fix_basis` defaults to `proposed`, as for mint. Every field `lens mint` sets has a disposition (§III.2's field table); no reader path is added | **F11** RULED |
| 1c. Where the gap stands | `verify --anchor`: at the end of the sentence its `--anchor` token sits in — where every gap's token stands, at its quote's end — found by that token's identity (`Anchor.sentence_of`), never by searching the sentence's text. `verify --url` names no anchor: its gap stands at its `--quote`, and a quote the report holds more than once is refused, as every text placement is <!-- CHANGED F13 --> | **F9** RULED |
| 1c′. `--new` on a reading | Blue's `--accept` and `record.ProposalOver` locate a gap's prescribed span by the gap's anchor identity — the occurrence of its location whose anchor run holds its token — and the accept's edit is recorded with `exact_span`, located byte for byte on replay. So `--new` on a `verify` over a sentence held twice is verified and accepted at the cited occurrence; nothing refuses it for the repeat | **Fork** RULED (locate by gap anchor) |
| 1d. Materiality | The class's `material_default` (both `always`), stamped at the write as for every gap | — |
| 1e. Estoppel | A reading's gap is EXEMPT: the reading verbs attach no `requireNotEstopped`, also where the sentence is text blue applied verbatim from a red `--new`. The gap opens; a `verify --anchor`'s failed check, and a `verify --url` gap where its class or its grades make it material, hold the PASS <!-- CHANGED F13 --> <!-- CHANGED F13b -->. The exemption and its reason are stated wherever estoppel is taught or enforced (§III.2) | **F10** RULED |
| 2. Duplicate | The reading is kept and the one-gap rule refuses the *mint*; `Verify.carried_by` names the standing gap | **F3** RULED |
| 3. Backing re-check | Found-against is a third **failed-check** state beside unread/unsound. It holds the PASS until its reader's own later reading backs the sentence, blue edits the sentence, or blue retires the citation; another lens's backing reading clears nothing. The lens's close of its gap is refused while it stands (`requireFailedCheckAnswered`); a closing docket ruling closes the gap and the check still holds the PASS. Once it clears, the minting lens closes the gap with its ordinary close (on main; #1657 Part 1a′, landing after this plan, adds its `source` demand on top, §III.16). No new answer kind | **F4**, **F8** RULED |
| A found-source reading's `absent` | Records the reading; mints no gap (F5, carried to `verify --url`) <!-- CHANGED F13 --> | **F5** RULED |
| Archived corroboration readings | An archived negative corroboration has no anchor, so after the change it holds no failed check and no PASS: the loss is accepted and stated per archive (§III.9). An archived BACKING corroboration's label, footnote and red-placed citation anchor are translated away: the reading comes forward as a reading of the source it found (§III.9 item 7) <!-- CHANGED F13 --> | **GAP 2** RULED; F13 |
| Mint budget | A gap a reading of a CITATION mints (`verify --anchor`) is outside the mint budget: it takes no mint and does not count while open; a gap a reading of a source the lens FOUND mints (`verify --url`, either class) takes a mint and counts while open, as a plain mint does; the "answers its own raise" exemption is deleted <!-- CHANGED F13b --> | **F6** RULED; amended F13b (gblock 2026-11-03) |
| 4. Delete | Everything in the census (§III.12) | — |
| 5. migrate | One translation, §III.9 | — |
| 6. Epoch | 55, re-derived at rebase | — |
| 7. PR #1642 | Closed as superseded; its four diagnostic tally fields are carried into this PR (§III.11) | **F7** RULED |
| 8. Delivery | ONE PR, implemented one phase per fresh agent (§III.14) | RULED |
| 9. The gap's link to its source | `carried_by` printed on `show evidence`'s rows and docket.md's source line (§III.2c) | **F12** RULED |
| 10. #1657 | This plan governs a refuting reading and lands first; 1a′ builds its `source` demand on `carried_by` (§III.16); 1a′'s corroboration arm is deleted, and a found-source reading's gap is answered by blue's `cite --answers` once 1a′ adds it <!-- CHANGED F13 --> | RULED |
| 11. The reading verb | ONE verb, `verify`, taking exactly one of `--anchor` and `--url` (with `--title` and `--quote`), enforced by cobra's flag groups at parse; `corroborate` deleted, no alias (§III.2d) <!-- CHANGED F13 --> | **F13** RULED (verb shape: this revision's decision) |
| 12. A source a lens found | Its reading mints its gap in the same act (`unsourced_claim` for a backing outcome on a sentence no evidence anchor marks; `claim_contradicts_source` for `refutes`), against the lens's mint budget; it places no citation, label or footnote; blue answers with a cite and a lens verifies the cite (§III.2d). One seat's readings in one sitting are keyed on source AND sentence: each that differs in url or in sentence is recorded, an identical retry is refused naming the stored reading <!-- CHANGED F13 --> <!-- CHANGED F13b --> | **F13** RULED; key and budget F13b |

### III.1 Record enums and messages — `internal/record/recordpb/record.proto` [MODIFY] (contract: event schema)

- `AboutKind`: `ABOUT_KIND_CONTRADICTION = 4` [DELETE]; `reserved 4; reserved "ABOUT_KIND_CONTRADICTION";`.
  The `gap_may_name` facet [DELETE] with its extension number reserved — after the deletion every kind
  admits a gap, and a facet true on every value is a field nothing can refuse.
- `GapClass`: [NEW] `GAP_CLASS_CLAIM_CONTRADICTS_SOURCE = 31` — means: "a source read against the
  sentence — cited by blue or found by red — states otherwise than the report"; `material_default =
  CLASS_MATERIAL_ALWAYS`; neighbor `citation_figure_misattribution`; distinguisher: "does the source say
  otherwise than the sentence, or not say it at all?". [MODIFY] `citation_figure_misattribution`'s means
  drops "or contradicted by" (it becomes the citation that does not back its sentence: absent from its
  text, a different document, a sentence the source is not about); [MODIFY] `unsourced_claim`'s
  distinguisher stops saying "says something else". <!-- CHANGED F13b --> [MODIFY] (F13b) `unsourced_claim`
  (`record.proto:461-463`, at `6102f2b09` still "a claim of fact or a quotation the report states with no
  source or computation behind it on the record", distinguisher "is a source recorded against the claim
  that does not back it, or is there no source behind the claim at all?") — both are false of the gap a
  backing `verify --url` mints, whose own carrying Verify puts a backing source on the record and whose
  source `show evidence` names (§III.2c). Rewritten: means "a claim of fact or a quotation the REPORT
  states with no citation or computation of its own behind it — no anchor, an anchor that resolves to
  nothing in the evidence, or an appeal to a result it never names. A source a lens read that backs the
  sentence is not the report's citation: the gap stands until blue cites one"; distinguisher "does the
  report cite a source for the sentence that does not back it, or does the report cite nothing for it,
  whatever a lens found?". Its neighbor stays `citation_figure_misattribution`. The property: the
  definition a lens, blue and the bench read is true of an `unsourced_claim` gap whose carrying reading
  names a source that backs the sentence and that the report does not cite. Carriers it reaches:
  `difftest/testdata/gap_class_enum.golden` and the manual golden's `--class` renderings (`mint`,
  `verify`) REGENERATE, each changed line read. Name grep (AGENTS.md naming rule) run at `9ee7ad95e`:
  `contradicts_source` has 0 hits; value 31 is free (`GAP_CLASS_UNCLASSIFIED = 30` is the last); the sibling `claim_contradicts_own_record` is the parallel the name
  follows (report vs record → report vs source).
- `SourceOutcome`: [NEW] facet `(gap_classes)` on `REFUTES` (`claim_contradicts_source`,
  `within_source_condition_misattribution`) and `ABSENT` (`citation_figure_misattribution`,
  `within_source_condition_misattribution`); absent on every other value. The facet says which classes
  an OUTCOME admits; it does not say a reading mints, because an outcome of a copy that is not the work
  reads nothing (`VerifyReads`, `anchorchecks.go:284-289`, on the `source_completeness` the tool stamps
  from the run's cache — on every reading once §III.2 makes `verify` stamp it; today only a
  corroboration carries it, `verify.go:495`).
- `SourceOutcome`: [NEW] facet `(found_source_gap_classes)` (F13), `repeated GapClass`, the extension
  number after `gap_classes`' 50014 (re-derived at the build): the classes a reading of a source the lens
  FOUND (`verify --url`, no anchor) admits — `SUPPORTS`, `SUPPORTS_WITH_BRIDGE`, `WEAK` →
  `unsourced_claim`; `REFUTES` → `claim_contradicts_source`; absent on `ABSENT` (F5) and `UNREACHABLE`.
  Read by [NEW] `recordpb.FoundSourceGapClassesOf`, beside `GapClassesOf` (`recordpb/gapclass.go:56`).
  `gap_classes` keeps its meaning, now stated as the classes a reading of a CITATION admits. Name grep at
  `6102f2b09`: `found_source_gap_classes`, `FoundSourceGapClassesOf` 0 hits. <!-- CHANGED F13 -->
  [MODIFY] (F13b) the `(means)` of `SUPPORTS`, `SUPPORTS_WITH_BRIDGE` and `WEAK` (`record.proto:514-516`
  at `6102f2b09`) say nothing of minting, which a backing `verify --url` reading now does; each gains
  "On a source you found (no citation), a reading of the work with this outcome on a sentence the report
  cites nothing for mints an `unsourced_claim` gap in the same act". `REFUTES`' means (`:517`, "A reading
  of the work with this outcome mints a gap on the sentence") is true of both forms and stays; `ABSENT`'s
  ("A reading of the work with this outcome on a citation mints a gap …") limits itself to a citation and
  gains "on a source you found it records the reading and mints nothing" (F5). The `--as` renderings in the manual golden
  REGENERATE, each changed line read. <!-- CHANGED F13b -->
  So the record declares two predicates over the stored `Verify`, and no site decides from `outcome`
  alone:
  - [NEW] `record.FindsAgainst(v) bool` = `VerifyReads(v)` and the outcome has `gap_classes`. It
    replaces `contradicts()` (`citationid.go:704`) and is the ledger's found-against test
    (`anchorchecks.go:260-261`: `holds = reads && !FindsAgainst`, found against = `FindsAgainst`), so a
    value added to the enum cannot fall between the reading rule and the anchor ledger. Name grep at
    `9ee7ad95e`: `FindsAgainst` 0 hits; it extends the ledger's own term (`stateFoundAgainst`,
    `CitationsFoundAgainst`) in the same sense.
  - [NEW] `record.ReadingGapClasses(v) []recordpb.GapClass` = the classes this reading's gap may take
    (F13): nil unless `VerifyReads(v)`; for a reading of a citation (`anchor` set), the outcome's
    `gap_classes` where `FindsAgainst(v)`, else nil; for a reading of a source the lens found (no
    `anchor`), the outcome's `found_source_gap_classes` — so a backing found-source reading mints and does
    NOT find against (its source backs the sentence; the defect is the sentence's missing citation), and
    its `absent` mints nothing (F5). "This reading mints" is `len(ReadingGapClasses(v)) > 0`, everywhere;
    "finds against" is `FindsAgainst`, the ledger's state, and is no longer the same set. <!-- CHANGED F13 --> Name
    grep at `9ee7ad95e`: `ReadingGapClasses` and `gap_classes` 0 hits (`archivedGapClasses` in
    `migrate/gapclass.go` is the archived-slug table, another sense). The `REFUTES`/
  `ABSENT` `(means)` lose "the strongest finding this verb can carry" in favour of "mints a gap".
  `within_source_condition_misattribution` is admitted because both archived answering gaps (m30
  `G-dc877265`, m32 `G-f417112d`) were minted in it by the lens that read the source.
- `Verify`: [NEW] `optional string carried_by = 21 [(sql) = { references: "mint.gap_id" }]` and
  `optional bool carrier_minted = 22` (built in phase 1). [DELETE] (F13) `label = 11`, `work_status = 15`
  and #1678's `sha256 = 20` — the footnote a backing corroboration minted, its retraction stamp and the
  copy it read — each `reserved` by number and name; no reading mints a footnote, so no live writer sets
  them, and the reading's own `url`, `title`, `copy_substitution`, `copy_url` and `source_completeness`
  name what it read. Their comments (`record.proto:1349-1356` "the one verb that reads a source red found
  itself", `:1375-1391`, `:1402-1406`, `:1418-1431`) [REWRITE] or go with the fields. <!-- CHANGED F13 -->
  `carried_by` is the gap that carries this reading against its sentence: the one the reading minted, or
  the open gap of the reading's class on that sentence that refused the mint. `carrier_minted` is true
  when this reading's own act minted it, false when it names a standing gap. Both are set by the tool,
  on a reading whose `ReadingGapClasses` is non-empty only (`seat.Supplies`, as `mint.go:288` declares the Mint's `gap_id`), so no
  flag sets them.
  - **Why not `gap_id`.** The idempotency key is matched by field NAME (`record.go:483` `keyFields`,
    `"gap_id"` first, ahead of `anchor` and `url`): a Verify carrying a field named `gap_id` would key
    on the gap instead of its citation. <!-- CHANGED F13b --> **The key (F13b, RULED gblock 2026-11-03: a
    found-source reading keys on its source and its sentence).** `keyFields` becomes `gap_id, label, id,
    anchor` — `url` LEAVES it, and `carried_by` is NOT added (the F13 text added it). The sweep the round
    asked for, every reading shape against one seat and one sitting (the key is `seat:verify:#sitting:<first
    keyFields value>` under a UNIQUE index, derived from the body as written, under replay too,
    `record.go:522-560`, `:844-890`):

    | The second reading's shape (the first: same seat, same sitting) | same url, same sentence | same url, other sentence | other url, same sentence | other url, other sentence |
    |---|---|---|---|---|
    | citation reading (`--anchor`) | one anchor: refused (key `anchor`), as today | another anchor: recorded | — (the url is the citation's) | recorded |
    | found-source, minting its own G (F13 key: `carried_by` = G) | F13: refused; F13b: refused (retry rule) | F13: recorded (another G); F13b: recorded | only with another class (a standing gap of its class would make it a held reading): F13 recorded (its own G′); F13b recorded | recorded both |
    | found-source, held (F13 key: `carried_by` = the standing G) | F13: refused; F13b: refused | F13: recorded; F13b: recorded | **F13: refused — the archived b7 shape; F13b: recorded, `carried_by` = G** | recorded both |
    | found-source, minting nothing (F13 key: `url`) | refused both | **F13: refused (the stated residue); F13b: recorded** | recorded both | recorded both |

    So under F13's key a second url read against the sentence one first url already carried was refused
    with "has already recorded a verify on "G-…" this sitting" — b7's archived shape (events 152-155,
    `red-lens-evidence`, keys `…:verify:#2:…`, four urls, one identical sentence), and the act F13's own
    supersede row sent the seat to. Under F13b every found-source reading by one seat in one sitting that
    differs in url or in sentence is recorded, the `absent` residue included.
    **The mechanism.** A found-source Verify carries no `keyFields` value (no `anchor`; `label` is
    reserved), so `deriveKey` gives it the ordinal key `seat:verify:#N` (`record.go:551-559`), which never
    collides; the identity a retry repeats is checked by [NEW] `requireNotARetriedReading` (`record.go`'s
    Verify arm, inside the writing transaction, run FIRST in the arm, before `requireNotARepeatReading`):
    a Verify with no `anchor` whose seat already holds a Verify with no `anchor`, the same `url` and the
    same `claim` (as stored, `strings.TrimSpace` as the repeat rule compares it), attributed to the seat's
    current sitting (`events.sitting_id IS (SELECT max(id) FROM sittings WHERE seat_id = ?)`, the equality
    `oncePerSittingSQL` asks, `oncepersitting.go:41-44`) is refused, nothing recorded, NAMING THE STORED
    READING: "you recorded a reading of <url> for this sentence at event N this sitting (<outcome>; gap G
    carries it | it mints no gap) — the record keeps your first one rather than quietly replacing it" + the
    superseders tail. It is ON under replay, as the key collision it replaces was: `run-archive/` holds no
    pair it refuses (re-measured, `~/.claude/scratch/plan1647/f13b/fs-all.txt`: the 22 archived
    found-source readings, 19 in the sqlite archives and 3 in the shard-era is-7-prime-run2 and
    sqlite-schema-source, include no two by one seat with one url and one claim). Name grep at `6102f2b09`:
    `requireNotARetriedReading` 0 hits.
    Rejected: a composite key string (`url` and `claim` joined into `events.key`) — two fields encoded into
    one string and a refusal that can name only the key, not the stored reading (facts-are-fields), and a
    `keyFields` entry that is no longer one field; keying on `carried_by` (F13) — the b7 refusal above.
    `Cite` keys on `label` (every cite carries one), so `url` leaving the list moves no other type:
    `Verify` is the only body `url` keyed (`supersede.go:102-105`). The premise F13 cited — "quadratic-formula
    ev 316, 318, 320, one seat, one url, three sentences" — is false: no such events; quadratic's
    labelled readings by one seat in one sitting are 295-297 (`red-lens-r2-L1`), three different urls and
    three sentences, and 514-516 (`red-lens-r3-L5`), likewise. The shape the key must admit is b7's: one
    sentence, four urls. <!-- CHANGED F13b -->
    A field named `gap_id` would also print as `--id` (`flags/names.go:402`), a flag the reading
    verbs do not have. Name grep run at `9ee7ad95e`: `carried_by` and `carrier_minted`
    have 0 hits. The names' other readers are swept in §III.12 ("Readers of the new Verify fields").
  - The write-path rules, in `record.go`'s Verify case, not table CHECKs (archived negative readings
    carry neither field, §III.9.2, and replay (`Migrating`) admits them as every other new write rule
    does — and a record written at this epoch replays its reading act whole, §III.9 item 6), each keyed on `ReadingGapClasses` of the body as stamped — never on `outcome` alone: a
    reading whose `ReadingGapClasses` is non-empty names its gap; a reading whose `ReadingGapClasses` is
    empty — a backing reading of a citation, `unreachable`, ANY outcome of a `not_the_work` copy, a
    found-source reading's `absent` — carrying either field is refused <!-- CHANGED F13 --> ("this reading reads nothing of the work, or backs the
    sentence, so no gap carries it"); `carrier_minted` true is admitted only inside the reading's
    placed append whose Mint has that `gap_id`; `carried_by` with `carrier_minted` false must name an
    open gap of a class in `ReadingGapClasses` that stands in the reading's sentence (for `verify`, the
    sentence its anchor sits in, §III.2a). A field the writer can refuse, not a relation recovered from sentence
    position or event adjacency (facts-are-fields).
- `Anchor`: [NEW] `optional string sentence_of = 7` (1-2 used, 3-6 reserved) — the id of an anchor
  already in the report; this marker is placed at the END of the sentence that anchor's token sits in
  (`anchor.Sentences`), where `anchortext.Attach` places a marker on a quote of that whole sentence —
  past its last content character and any closing quotation mark touching it (`placement`, #1654),
  before its terminator — at the occurrence holding that token, found by the token's identity (F9,
  §III.2a). Every gap's token so stands at its quote's end: a citation's token mid-sentence (a clause
  citation) does not pull the gap's to it. An Anchor carries exactly one of `location` and
  `sentence_of`; `record.go:949-951` ("anchor requires location") [REWRITE] to "requires location or
  sentence_of, not both", and the arm's comment (`:942-943`, "places one marker … at its quote's end")
  [REWRITE] with it. The message's comment (`record.proto:1179-1181`, "at the end of its quote
  `location`") [REWRITE]: or at the end of the sentence `sentence_of`'s token sits in. No SQL
  reference: the placement guard (§III.2a) refuses a `sentence_of` whose token the landing report does
  not hold. Name grep at `9ee7ad95e`: `sentence_of`/`SentenceOf` 0 hits.
- migrate's field lists (`migrate/remap.go:108-144`, gated by `TestEveryStringFieldIsInOneList`):
  `Verify.carried_by` and `Anchor.sentence_of` are ids → `idFields` (an archived id equal to the
  value is respelled, as `Verify.anchor` and `Anchor.id` are); `carrier_minted` is a bool and is in no
  list (the lists class string fields only). (F13) `Verify.label` leaves `idFields` and `mints`, and
  `Verify.sha256` leaves `dataFields`: the archived columns are read by §III.9 item 7's translation,
  never by the live lists. <!-- CHANGED F13 -->
- `Mint`: the `mint_reason` comment (`record.proto:968-978`, "`lens mint` stores --reason here and
  nowhere else") [REWRITE]: `lens mint` and a reading that mints store --reason here. Comments at `record.proto:904-907` [MODIFY] (no kind is refused). No field change.
- Regenerate: `go -C scripts run ./protogen`, `go -C scripts run ./schemagen`; `requirements.json`
  `eventSchema` 54 → 55.

### III.2 The reading act — `internal/cli/lens/verify.go` [MODIFY] (contract: CLI flag set; prompt a seat is told)

- **Every reading stamps the completeness of the copy it read (GAP 1, RULED: "verify stamps it").**
  Today the stamp (`verify.go:492-501`) runs only where `body.GetUrl() != ""`, and `newVerify`
  (`verify.go:97`) builds `Verify{Anchor}` with no url — the source is `citedURL`'s, passed beside the
  body as `source` — so a `verify` never carries `source_completeness`, `VerifyReads` reads its
  `not_the_work` copy as a reading, and `verify --as refutes` of a sign-in page would mint. [MODIFY]: the
  stamp keys on `source` (the cited url's cache entry for `verify --anchor`, `--url`'s for `verify
  --url`, the one `held` entry `fetchcache.Lookup` already read at `:366`), never on the body's url, and
  runs before every decision below, so each reads the stamped body. (F13) The footnote decision
  (`verify.go:426`, `mayCite == cites && backsTheClaim(...)`), `citing`/`cites`/`adjudates` and
  `backsTheClaim` (`:336-350`) are DELETED with the footnote (§III.2d): no reading mints a label, a
  footnote or an Anchor of a citation. A backing `verify --anchor` of a `not_the_work` copy leaves its
  citation on the `unread` failed check, owned by its reader, who owes that check's gap about the anchor
  (the ledger already reads it so, `anchorchecks.go:259-266`, once the stamp is there); a `verify --url`
  of such a copy reads nothing and mints nothing. Tested by
  `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` rows (g)-(i) (§III.10); partner §V.5.w. <!-- CHANGED F13 -->
- `verify` gains the mint's judgment flags <!-- CHANGED F13 -->, the same `flags.*` names and descriptions
  `lens mint` registers (`mint.go:265-280`): `--class`, `--problem`, `--fix`, `--new`, `--check`,
  `--check-kind`, `--severity`, `--likelihood`, `--impact`, `--complexity`. Where `ReadingGapClasses` of
  the body is non-empty, the seven `mint` requires — `--class`, `--problem`, `--check`, `--check-kind`,
  `--severity`, `--likelihood`, `--impact` — are required, each refusal naming its flag, and `--fix`,
  `--new` and `--complexity` are optional, as on `mint`. Where it is empty, all ten are refused (§III.1) — decided in
  `writeVerify` AFTER the copy's completeness is stamped (`verify.go:492-501`, keyed on `source` above), never from `--as` at
  flag parse. The refusal names why nothing is minted: "a reading that backs the sentence mints
  nothing"; "this copy is not the work (`not_the_work`), so the reading reads nothing of it and mints
  no gap — the citation stands unread, and the gap owed is about its anchor"; "an unreachable source
  was not read and mints no gap"; "a reading of a source you found that is `absent` records the reading
  and mints no gap". <!-- CHANGED F13 -->
  `--class` is refused outside `ReadingGapClasses`, and the refusal lists them. One flag registration
  helper shared by `mint` and the reading verbs, so the two cannot drift — and it registers each flag
  with the description true ON THAT VERB'S PAGE. [NEW] `gapFlags` in `cli/lens/mint.go` holds the four
  `flags.GradeValue`s, `register(c, subject)` registers the ten flags, and `fill(cmd, text, key)` builds the
  body (the field table below), so one function fills a plain mint's Mint and a reading's. Name grep at
  `9ee7ad95e`: `gapFlags` 0 hits. Swept: the descriptions `mint.go:265-281` gives these
  flags, each read for a word the reading verbs do not have or a duty they do not carry:
  - `--new` ("the exact text --quote should become", `:274`) speaks of `--quote` alone, which a `verify
    --anchor` call does not take (F13: `verify` registers `--quote` for its `--url` form, so
    `TestEveryFlagAHelpPageNamesIsOnThatPage`, `cli/deadpath_test.go:356`, no longer catches it; the
    wording is what is wrong). The
    helper takes the verb's SUBJECT phrase: `mint` "--quote"; `verify` "the sentence your --anchor's
    token sits in, or your --quote" (F13: one page serves both forms; the refusal at the act names the
    one the call used, §III.2a) <!-- CHANGED F13 -->;
  - `--class` ("unclassified only when none fits, and then --reason names the nearest value", `:265-266`):
    no reading admits `unclassified`. On the reading verbs: "the KIND of defect this reading finds, from
    the values below". `enumhelp.Flag` (`enumhelp.go:55`) renders and parses the `record.EnumField` it is
    handed, and `record.MustEnum` reads `EnumFields`, keyed by event type and payload key — `Verify` has
    no class field, so `MustEnum("verify", "class")` panics, and a row there would declare a field the
    body does not hold. [NEW] `record.ReadingClassField() EnumField` (F13: one verb, no parameter):
    `MustEnum("mint", "class")` (its `Key`, `Flag`, `Why`) with `Values` narrowed to the union of every
    outcome's `gap_classes` and `found_source_gap_classes`, read off the `SourceOutcome` descriptor — four
    classes: `claim_contradicts_source`, `within_source_condition_misattribution`,
    `citation_figure_misattribution`, `unsourced_claim`. `enumhelp` is unchanged, so the parse refuses a
    class outside the union and the page lists it; `ReadingGapClasses` then narrows to the form's and the
    outcome's. <!-- CHANGED F13 -->
    `TestEverySetShapedFlagIsEitherDeclaredOrExempt` (`cli/enum_help_test.go:200`) keys a site by the
    RECORD TYPE the verb writes (`setKey`, `:156-158`), and walks three sites no `EnumFields` row declares, each entering `enforcedElsewhere` (`:72`) with its
    reason: `verify --class` — "the reading's Mint's class, from `record.ReadingClassField`; a Verify has
    no class field"; `verify --check-kind` — "the reading's Mint's check kind, parsed by `enumhelp.Flag`
    over `record.MustEnum("mint", "check_kind")`; a Verify has no check_kind field"; `verify --severity`
    — `gradeParseTime`, as `mint --severity` is (`:114`). Name grep at `9ee7ad95e`: `ReadingClassField`
    0 hits;
  - `--problem` ("Your argument for the gap goes in --reason, which the report never prints", `:272`): the
    whole sentence is true on the reading verbs and they keep it unchanged: their `--reason` is that
    argument (below), and `the report` (the registry's term: report.md) prints no gap's argument — docket.md
    prints it as "minted because" (`assemble.go:1055`, `:1124`; closure index `:1142`), and `show report`
    bare prints neither `mint_reason` nor `fix_basis`. `mint`'s text, `mint.md:23` and `mint`'s `ReasonIs`
    row are byte-unchanged;
  - `--fix`, `--check`, `--check-kind`, `--severity`, `--likelihood`, `--impact`, `--complexity`: true on
    every verb, one text.
  **The conditional requirement is marked.** `--class` and the six others above carry
  `seat.RequiredUnless(c, name, cond)` on the reading verbs (`seat.go:763`; `mint`'s are marked from the
  schema by `markRequired` and do not move), with `cond` (F13: one verb) = "unless this reading reads
  nothing of the work, backs a citation's sentence, or is `absent` on a source you found", so the page
  opens "REQUIRED unless … — " and no seat loses a call to learn it.
  `TestHelpSaysRequiredExactlyWhereOmissionIsRefused` reads a marked flag the clean set never passed as
  "says REQUIRED and the verb ran clean without it", so its `requiredProbes` entry for `lens verify`
  (`cli/requiredmarker_test.go:98-115`; the `lens corroborate` entry DELETED with the verb) [REWRITE] to a
  clean set that finds against: `--as refutes`, `--confidence`, `--reason`, the seven flags, on the
  fixture's citation. <!-- CHANGED F13 --> Each omission is then refused naming its flag; `--fix`, `--new`, `--complexity` stay out
  of the set and unmarked. The backing path is driven by `TestABackingReadingWithGapFlagsIsRefused`.
  <!-- CHANGED F13b --> **The subject's one-of.** `--anchor` and `--url` stand in for each other, which no
  single-flag marker states, so `cli/requiredmarker_test.go` gains the pair in `oneOfAnother`
  (`"anchor": "url"`, `"url": "anchor"`, `:64-72`) and `"verify"` in `verbsWithAOneOf` (`:78`); without
  them the probe reads the flag group's refusal as "REFUSED without --anchor and its help does not say
  so". The page carries the alternative in prose, as `mint`'s `--reason`/`--problem` does
  (`helpcontract_test.go`'s `satisfiedByAnother`). Gate: `TestHelpSaysRequiredExactlyWhereOmissionIsRefused`.
  **Which refusal comes first.** A backing `--url` reading of a sentence an evidence anchor marks, sent
  without the gap flags, meets two refusals: the marked-sentence guard (§III.2d) and "REQUIRED --class".
  The guard's sentence is the actionable one (read that citation with `--anchor`; the gap flags would
  only meet the guard next), so `writeVerify` asks `uncitedSentence` of the report it renders, after the
  completeness stamp and `ReadingGapClasses`, BEFORE the gap-flag requirement; the guard runs again on
  the landing report inside the append, where it is authoritative under the lock. Tested by the
  no-gap-flags row of `TestABackingFoundSourceReadingOfAMarkedSentenceIsRefused` (§III.15 35c).
  `mint`'s rendered flag text is byte-unchanged. Tested by
  `TestTheReadingVerbsDescribeTheirGapFlagsByTheirSubject` (§III.10) and the gate above.
- **`--reason` is the reading's account and the gap's argument (F11, RULED: "Reading takes --reason").**
  `writeVerify` reads the one `seat.Reason(cmd)` (`verify.go:405-409`) and writes it to the two fields
  whose readers ask for it, at the act: `Verify.text`, as today, and, on the Mint the reading builds,
  `p.MintReason` as `mint.go:69-71` does. What the source says against the sentence IS why the gap is
  worth one. It is one text from one flag. No second flag is added, no reader joins `Verify.text` onto a
  gap, and no field is copied from a stored field or from `problem`. `Verify.text` is `(sql).required`
  with flag `reason`, so it cannot be the field left blank, and `--reason` is already REQUIRED on both
  reading verbs: a reading that mints is refused without it, nothing recorded. `problem` stays
  the only statement of the defect. [NEW] `ReasonIs` row (`cli/seat/seat.go:1082`) for `verify` (F13:
  the `corroborate` row is deleted with the verb) <!-- CHANGED F13 -->: "what the source says, in your words, or, where it could not be read, what you tried.
  Where this reading mints a gap, the same text is that gap's argument, the case blue answers and a bench
  weighs; where an open gap of its class already carries the sentence, it stays on your reading alone and
  that gap keeps its own. " + `flags.ReasonNotProcess`. Where the one-gap rule names a standing gap, no
  Mint is written: the reason stays on the Verify, and the standing gap keeps its own `mint_reason`.
- The write: one transaction holding the Mint, the Anchor that places the gap's marker (§III.2a), and
  then the Verify carrying `carried_by` — Mint first, because `carried_by` references `mint.gap_id` and
  the store runs with `foreign_keys(1)`. The Mint's `location` is the reading's claim, anchors stripped
  (`citedSentence` for `verify --anchor`, `seat.Location` for `verify --url`) <!-- CHANGED F13 -->, so every reader of a gap's quote
  (the board, the gap view, `report/assemble.go`) reads the sentence as today. `record.go` gains the
  reading's placed append beside `AppendPlaced` (one `appendOwn` over the three bodies). `appendOwn`
  returns the FIRST body's event (`record.go:679`), which is now the Mint, and `writeVerify` reads the
  stored Verify with `recordpb.BodyAs[*recordpb.Verify](ev)` (`verify.go:525`), whose miss is the verb's
  "written as %T" error: so the reading's append returns every event it wrote (`[]*Event`, Mint and
  Anchor first when minted, the Verify last), and `writeVerify` reads the Verify off the last and the
  gap off the stored `carried_by`, never off the first event; the Mint passes
  every mint validation (`record.go:955-980`: unclassified reason, subject, one-gap, budget) in the same
  `writeView`, so the one-gap read and the insert share the write lock (`_txlock=immediate`, as
  `onegapperclass.go:32-35` already relies on).
- **The budget exemption comes from the act, not the record (F6).** The Mint's budget check runs before
  the Verify exists in the transaction, so it cannot read a Verify naming the gap. The reading's placed append
  hands the Mint's validation the act's paired bodies (its Anchor and its Verify); `requireMintWithinBudget`
  admits a Mint so paired without counting it where the paired Verify has an `anchor`, and counts it
  where it has none (F6 amended, §III.5) <!-- CHANGED F13b -->, and the one-gap rule places it by the paired
  Anchor (§III.2a). A plain `mint` has no paired body and is counted and placed as today. Replay
  writes events one at a time (`migrate/replay.go:448`), so it would hand the Mint no pair: migrate
  re-drives the reading's act whole through the same placed append (§III.9 item 6), and the Mint's
  validation meets the same paired bodies on replay as live.
- **Which of `lens mint`'s checks the reading's Mint meets (F10, RULED).** Every rule of `record.go`'s
  Mint arm (`:952-…`) runs on the reading's Mint as on any: `requireUnclassifiedReason`,
  `requireMintSubject`, the one-gap rule (its held gap returned, below), `requireMintWithinBudget` (the
  act-paired exemption, §III.5), `requireAboutAnchor`, the required check kind, likelihood and impact,
  the report-voice refusal, and the rest of the arm. The handler `mint.go` builds attaches more, at the
  verb; swept over the handler (`mint.go:36-291`), every check it makes or guard it attaches:

  | `lens mint`'s check (`mint.go`) | The reading's Mint |
  |---|---|
  | `seat.Reason` (`:44`) → `Mint.mint_reason` | MET (F11): `--reason`, required when the reading finds against; the same text is `Verify.text` |
  | `--class` through `record.GapClassOf` (`:75-81`) | MET, and narrowed to `ReadingGapClasses` |
  | `record.ResolveAbout` and the one-subject check (`:99-108`) | NOT APPLICABLE: the reading verbs take no `--about-kind`/`--about`; the subject is the reading's sentence, which `requireMintSubject` reads off `location` |
  | placement: `seat.AppendPlaced` with `mintPlacementRefusal` (`:212`, the `seat.Places` guard) | MET by `verify --url`, through `claimRefusal` (the same `anchortext.Refusal` arms); `verify --anchor` meets `seat.PlacesToSentenceOf` instead (§III.2a) <!-- CHANGED F13 --> |
  | `verifiedFix` on `--new` (`:136-139`, `record.CheckOnReport`) | MET, placed by the verb's subject (§III.2a) |
  | `--check-kind` through `record.CheckKindOf` (`:147-153`); the four grades (`:154-157`) | MET (the shared flag helper) |
  | `--supersedes`, `--found-by` (`:158-159`) | NOT TAKEN: the reading verbs register neither; the gap supersedes none, and its reading is named by `carried_by`. What `found_by` becomes is #1695's (§III.16) <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT --> |
  | `--key`: `record.ExistingMintByKey`, `seat.RepeatsKeyed`, `seat.PlaceOwed` (`:164-180`) | NOT TAKEN: the reading verbs take no `--key`; a retry meets the duplicate-key refusal (`--anchor`) or `requireNotARetriedReading` (`--url`), each naming the stored reading (crash-retry, below) <!-- CHANGED F13b --> |
  | `record.NewID("gap")` (`:182`) | MET (one id source) |
  | `requireNotEstopped` (`:208`, `record.EstoppelConflict`), and on its refusal the `LOG_TYPE_ESTOPPEL` entry and `record.ToolLogged` (`:211-247`) | **EXEMPT** (F10). Nothing is checked and nothing is logged, so `EstoppelRejectionsOf` and the change log's "Estoppel rejections" (`view/changes.go:93-103`) count lens mints only, and stay true |
  | `reportvoice.Advised` tells on the result (`:250-260`) | MET: the reading's `gap` result carries `voice_tells` over problem, fix and `--new`, and its human text prints them under `mintResult.Human`'s note — "the risk matrix, the first sentence of the problem and of the fix, while the gap is open; the body, a --new replacement blue accepts" (`mint.go:320`) — `verifyResult`'s "the source's note and Bibliography entry" note is DELETED with the footnote (F13) <!-- CHANGED F13 --> |
  | `mintResult.Check` and its echo (`:300-308`, `:324-325`: "acceptance check RECORDED as: … read it back — a shell may have rewritten it before this tool saw it, and no verb amends a check after mint") | MET (GAP 3): the measured shell-corruption defect the echo exists for (quadratic-formula red-merge-r3) reaches a reading's `--check` alike. `readingGap` gains `acceptance_check`, read back from the stored Mint, and the reading's human text prints the same two lines when it mints. A reading that names a standing gap prints no check (none was written) |
  | `seat.Supplies("gap_id")`, `seat.Supplies("class_material")` (`:288-289`) | MET: the reading verbs declare both, beside `carried_by` and `carrier_minted`. <!-- CHANGED F13b --> `verify` also KEEPS `seat.Supplies(c, "claim", …)` (`verify.go:117`): `Verify.claim` is `(sql) required, flag: "quote"`, and without the declaration `markRequired` makes `--quote` REQUIRED on the `--anchor` form, which takes no `--quote`. Its reason is reworded for both forms: "the sentence holding the --anchor token in the report as it stands (citedSentence); on a source you found, your --quote", and `--quote`'s requirement on the `--url` form is the flag group `MarkFlagsRequiredTogether(url, title, quote)` (§III.2d). Gate: every `verify --anchor` CLI test (a `--quote` demanded there fails them) |

  **Every field `lens mint` sets on its body, and every reader of a field the reading leaves unset (F11).**
  Sweep at `0c57758e8`, carried to `9ee7ad95e` by the line map (`mint.go`'s changes there are #1663 smoke D7's refusal clause and `flags.IDList` for two flags, neither a field set): every `p.<Field> =` in `newMint` and `verifiedFix` (`mint.go:68-183, :380-381`), the
  write path's stamp, and `grep -rnE '<field>|<Getter>'` over the tools tree for each, no tests, every hit
  read (`new-mintfields.txt`). Both verbs fill through `gapFlags.fill`, so these rows are one code path:

  | Field (`mint.go`) | The reading's Mint | Readers, where it is unset or defaulted |
  |---|---|---|
  | `mint_reason` (`:69-71`) | `--reason` | Carried, so `GapJSON.MintReason` (`viewjson.go:129, :448`), the gap view (`views.go:529`) and docket.md's "minted because" (open-gap block `assemble.go:1055` → `:1124`; closure index `:1142`) read it as a plain mint's. `requireUnclassifiedReason` (`record.go:1947`): no admitted class is `unclassified`. `redvoice.go:26`: never voice-checked, as on a mint |
  | `mint_key` (`:72`) | `proto.String("")` — PRESENT and empty, exactly as a plain mint given no `--key` (`mint.go:72` sets `proto.String(seat.Str(cmd, flags.Key))`). `gapFlags.fill(cmd, text, key)` takes the key as a parameter: `mint` passes `--key`, the reading verbs (which register no `--key`) pass `""`, so fill never leaves it nil | `ExistingMintByKey` (`replay.go:360`), whose `keyedAct` returns nil on an empty key (`id.go:47`), so no retry finds it |
  | `class` (`:75-81`) | `--class`, narrowed (§III.1) | set |
  | `location` (`:113`) | the claim, anchors stripped | set |
  | `about_kind`, `about_ref` (`:114-115`) | UNSET, as a plain `--quote` mint: the subject is the sentence | `requireMintSubject` (`record.go:1937`) reads `location`; `requireAboutAnchor` (`record.go:973`) has no kind to check; the one-gap rule takes its sentence path (`onegapperclass.go:54`); `mintbudget.go:386, :403` and the about-anchor scan (`failedcheck.go:63`) pass it by, its carrier being `carried_by` (§III.3); `gapacts.go:41, :52`, `blueanswered.go:103`, `grounds.go:117`, `viewjson.go:415, :1065`, `views.go:526-527` read a sentence gap |
  | `problem` (`:116`) | `--problem`, required | set |
  | `required_fix` (`:117`) | `--fix`, optional | `proto.String` of the flag, `""` when absent, as a plain mint's: the report prints it blank alike |
  | `fix_basis` (`:135`) | `"proposed"` by default, `"verified"` through `verifiedFix` (`:376`) | Set. Without it `viewjson.go:811`'s "never omitted" is false and `fixProposal` (`assemble.go:1274-1284`) prints nothing; `estoppel.go:131`, `changesjson.go:113`, `view/changes.go:134` read it as a plain mint's |
  | `fix_new` (`:375`) | set by `verifiedFix` on `--new`, else UNSET | As a plain mint with no `--new`: `locateGap` and the board give no pair, `assemble.go:1277` prints none, estoppel skips it (`estoppel.go:131`) |
  | `acceptance_check`, `check_kind` (`:145-153`) | `--check`, `--check-kind`, required | set |
  | `severity`, `likelihood`, `impact` (`:154-156`) | flags, required | set |
  | `complexity_cost` (`:157`) | `--complexity`, optional | `nullWord` on the board, `grade` in the report, as a plain mint's |
  | `supersedes` (`:158`) | UNSET: the reading verbs take no `--supersedes` | A gap with no ancestor, as a plain mint's: `onegapperclass.go:64` (nothing exempt), `benchgrade.go:89, :198, :206`, `impasse.go:188`, `lensturn.go:52`, `gapacts.go:44`, `successor.go:41`, `record.go:1027`, `viewjson.go:166, :442`, `verify/verify.go:235`, `view/view.go:304, :457, :467`, `graph/graph.go:232, :280`, `dashboard/model.go:500`, `consistency.go:162`. A later gap naming it is §III.3's successor |
  | `found_by` (`:159`) | UNSET: the reading verbs take no `--found-by`; the reading is named by `carried_by` | A gap no finding surfaced, as a plain mint's: `provenance` prints no "surfaced by" (`assemble.go:982`), and the argument and the source's reading stand on `mint_reason` and `show evidence`; `requireFindings` (`record.go:1024`), `verify/verify.go:165, :442`, `view/view.go:273, :308`, `assemble.go:1345`, `viewjson.go:165, :442, :859, :1124, :1250, :1554`, `views.go:570` (no reader). The reading writes no `found_by`; what the field becomes is #1695's, after this plan and 1a′ (§III.16) <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT --> |
  | `gap_id` (`:183`) | `record.NewID("gap")` | set |
  | `class_material` | stamped by the write path | set |

  The property: a reading's gap carries its argument and its `fix_basis` wherever a plain mint's appear,
  through the readers that exist. Tested by `TestAReadingsGapIsFilledAsAPlainMints` and
  `TestAReadingsGapCarriesItsArgumentWhereverAMintsDoes` (§III.10); partners §V.5.ee, ff.

  **Why estoppel does not reach a reading (gblock, 2026-10-31: "Exempt, stated").** Estoppel binds red
  to its own prescription: a gap whose quote is text blue applied verbatim from a red `--new` is red
  relitigating its words (`record/estoppel.go:14-29`). A reading is a source read against the sentence:
  its refutation is new evidence, which the prescription never weighed. Reachable: `verify --url
  --as refutes --quote` of text blue applied verbatim, and `verify --anchor` of a citation whose token a
  red `--new` carried onto its rewording. The gap opens; a `verify --anchor`'s found-against check holds
  the PASS until it clears by the ledger's rule (§III.3), and a `verify --url` refuting gap, material by
  its class, holds it while open. <!-- CHANGED F13 --> Tested by `TestAReadingOverAnAppliedPrescriptionMintsItsGap` (§III.10), partner §V.5.cc.

  **Sweep — where estoppel is taught or enforced**, run at `9ee7ad95e`:
  `grep -rIn -E '[Ee]stopp|ESTOPP' plugins scripts docs | grep -v -E '\.golden:|testdata/|record\.pb\.go|_test\.go|\.test\.mjs|/agents/'`
  (94 lines at `9ee7ad95e`: #1678 removed `EstoppedJSON`'s constant closer, another sense; `~/.claude/scratch/plan1647-r10/new-est.txt`), every hit read. Two senses: the
  PRESCRIPTION estoppel (the mint guard), and the work list's `estopped` register of bench-ruled
  closures. Dispositions:
  - enforced — `cli/lens/mint.go:185-208` (the guard's comment) and `:417-435` (`requireNotEstopped`'s
    doc) [REWRITE]: add that the reading verbs attach no such guard, and why; `cli/lens/verify.go`'s
    reading append [NEW comment] says the same at the write;
  - taught to the record's readers — `record/estoppel.go:14-29` (the rule's statement) [REWRITE]: a
    reading's gap is exempt, a source's refutation being new evidence; `record/recordpb/record.proto:1618`
    (`applied_verbatim`: "red may not open a clean slate against its own words") and
    `record/changesjson.go:74` ("Estoppel is enforced on the first") [REWRITE]: on a lens's mint, never on
    a reading's gap;
  - taught to seats — `help/mint.md:19` [REWRITE] adds: "A gap your reading of a source mints in its own
    act is not refused: a source that finds against the sentence is new evidence, not your prescription
    relitigated"; `help/verify.md` (the new reading paragraph, §III.7) [NEW
    sentence] states it <!-- CHANGED F13 --> in the same words; `help/edit.md:13` (blue's: the estoppel "that stops red
    re-opening its own words") [REWRITE] adds "— a lens's reading of a source that finds against that
    text still mints its gap"; the generated `agents/*.md` (each inlines the help pages through the
    tool's `manual`) and the help golden REGENERATE;
  - counted, true as is — `record/estoppel.go:71-103, :238-257` (`EstoppelConflict`,
    `EstoppelRejectionsOf`), `view/changes.go:17, :93-103`, `record.proto:109, :1766` (`LOG_TYPE_ESTOPPEL`
    means "the TOOL refused a mint"), `record/record.go:1257`, `cli/seat/verbs.go:172`,
    `record/toollogged.go:5`, `recordpb/descriptions.go:202`, `cli/blue/edit.go:68, :188-201`,
    `migrate/entries.go` (archived log entries), `seatenv/identity.go:116`, `cli/seat/seat.go:819`,
    `cli/motion/verbs.go:95`, `report/assemble.go:1265`, `scripts/agentgen/src/lead-judge.md:101` — KEEP;
  - the rest of the guard's own lines, true as is — `record/estoppel.go:31-48` (`minEstoppelOverlap`, what the
    guard compares), `cli/lens/mint.go:341-345, :385` (the `estopped` error and `contains`),
    `record.proto:1739, :1787, :1790` (a `Log` comment and the reserved `estopped_by`) — KEEP;
  - another sense — `record/outcomeclosed.go:24` (`cli/enginestopped.go`, the engine's stop) — KEEP;
  - the `estopped` register (another sense: a bench ruling's bar, read as advice by every seat, and
    refused at no write) — `record/viewjson.go:650-1276`, `cli/seat/verbs.go:327, :343`,
    `record/sitting.go:589`, `record/declarations.go:28`, `record/enums.go:166`,
    `cli/motion/verbs.go:26`, `recordsql/views.go:487` — KEEP; a reading's gap meets it as a plain mint
    does;
  - the fuzz — the estoppel drive (`fuzz_test.go:576-674, :1407-1411`) mints with `lens mint` and KEEPs;
    its gates (`sweeptally_test.go:1052-1077`) count tool estoppel entries, which no reading writes, and
    KEEP; the reading drive gains the prescribed-text case (§III.11).
  The exemption's own readers: the verb (no guard), the write (no log entry), the help pages above, and
  the fuzz's new oracle; §V.4 re-runs the sweep.
- One-gap refusal **inside the reading act** (F3): `requireOneGapPerClassAndSubject` returns the held
  gap as a value rather than only an error; the reading path writes the Verify with `carried_by` = the held
  gap and no Mint and no Anchor. The plain `mint` verb keeps its refusal unchanged.
  **The decision is made inside the write's transaction, never before it.** A held gap read before the
  lock is stale by the time the pair lands: a second lens's reading of the same sentence and class
  committing in between would make this reading's Mint a duplicate, and the one-gap rule inside the
  transaction would then refuse the whole act. So the reading's append is one `appendOwn` whose Mint
  validation, on finding a held gap, does not refuse: it hands the held gap back to the append
  (the one-gap rule's existing `heldGap`, `onegapperclass.go:99-104`, returned on the `writeView` as a value, not an error), which drops the Mint and the Anchor from the
  bodies still to write and sets the Verify's `carried_by` to it and `carrier_minted` false — all under
  the same lock. Tested by `TestTheReadingsHeldGapIsDecidedInsideItsWrite` (§III.10).
- Repeat: `requireNotARepeatFailedCheck` (`failedcheck.go`) extends to a reading for which
  `FindsAgainst` holds — or, for a reading of a source the lens found, whose `ReadingGapClasses` is
  non-empty (F13: a backing found-source reading mints and does not find against) — by the same lens,
  of the same sentence from the same source (the anchor for `verify --anchor`; the url and claim for
  `verify --url`), with the same outcome as an earlier reading of its own of that kind, sentence
  unchanged, <!-- CHANGED F13 --> while a gap of that earlier reading's carrier
  tree (`carried_by` and its successors, §III.3) is open — refused, nothing recorded, naming the earlier
  reading's event and the carrier (`AnsweredBy()`): "your reading of citation C at event N finds the
  source against this sentence, the sentence has not changed, and gap G carries that reading and is
  open — nothing is recorded; the reading and its gap stand" (the `--url` form: "your reading of <url>
  for this sentence at event N … gap G carries that reading"). The phase-1 refusal's closing clause, "a
  point the gap lacks is a finding about G", is DELETED: the advice to make a finding moves to #1695
  (§III.16) <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->. <!-- CHANGED F13b --> Both of the rule's gates
  move to `ReadingGapClasses`, not only the loop's: `requireNotARepeatReading`'s early return
  `if Migrating || !FindsAgainst(v) { return nil }` (`failedcheck.go:319`) cuts a backing found-source
  reading off before the unanchored arm, so it becomes `if Migrating || len(ReadingGapClasses(v)) == 0`,
  and the loop's `!FindsAgainst(prev)` (`:363`) becomes `len(ReadingGapClasses(prev)) == 0`; for a
  citation reading the two predicates are one set (§III.1), so the anchored arm is unchanged.
  `repeatReadingRefusal`'s text (`:336`, "finds the source against this sentence (%s)") is false of a
  backing repeat, so it words by form: the anchored form as above; the found-source form "your reading of
  <url> for this sentence at event N (<outcome>) is carried by gap G, which is open, and the sentence has
  not changed — nothing is recorded; the reading and its gap stand. A reading of yours with another
  outcome is recorded". Within one sitting the retry rule (§III.1) answers first. A reading with a
  *different* outcome is not a repeat. A reading for which `FindsAgainst` does not hold (`unreachable`, any outcome of a
  `not_the_work` copy) never meets this arm: over a citation it meets the `unread` arm as today, keyed
  on that state (§III.3).
- Text that states the old model without the word [REWRITE]: the `--as` flag help (`verify.go:304`,
  "refutes and absent are findings") and `human()` (`verify.go:702-705`, "that is a finding, and this event
  is the evidence for it"); see §III.7's phrase sweep.
- **`human()` words a reading by whether it reads the work, then by outcome (GAP 2)** [REWRITE]
  `verify.go:694-712`. Today it switches on the outcome string alone, so a `refutes` of a copy the tool
  stamped `not_the_work` prints "REFUTES the claim … recorded" beside the failed-check words "reads
  nothing of the source". `verifyResult` gains `SourceCompleteness` (read back from the stored Verify)
  and `human()` opens on `record.VerifyReads` of the stored body: a reading that reads nothing prints
  "recorded as <outcome> on a copy that is not the work (not_the_work): this reading reads nothing of the
  source and mints nothing" (or the `unreachable` sentence, unchanged); only then does the outcome pick
  the sentence, and the `refutes`/`absent` sentences end on the gap ("This reading mints gap G" / "Gap G
  … carries this reading"), never "that is a finding".
- Result: `verifyResult.Contradiction` / `readingContradiction` / `contradictionOf` / `words()`
  (`verify.go:616-692`) [DELETE]; [NEW] `Gap *readingGap {id, class, minted, acceptance_check,
  voice_tells}` — every field of `mintResult` a reading's gap has (`GapID` → `id`; `Check` →
  `acceptance_check`, the stored Mint's; `VoiceTells`; `Idempotent` has no reading counterpart, a reading
  taking no `--key`) — and its sentence:
  "This reading mints gap G (class C)" / "Gap G, open on this sentence with class C, carries this
  reading; no second gap is minted" (F13: the clause "add what it lacks as a finding about G" moves to
  #1695, §III.16) <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->. `gap.minted` is read back
  from the stored `carrier_minted`, never computed by the verb, and `Gap` is set exactly when the
  stored Verify carries `carried_by`: a reading whose `ReadingGapClasses` is empty prints no `Gap`
  sentence, whatever its `--as` (its citation's words are its failed check's, by state, below). Where
  `minted` is true the sentence is followed by `mintResult.Human`'s check echo and voice note, worded as
  there (guard table above). Tested by the GAP 3 row of `TestAReadingsGapIsFilledAsAPlainMints`.
- **The failed-check words key on the state** (`verify.go:530-542`, `:632-657`) [REWRITE]. Today the
  verb picks any `BlockerFailedCheck` whose Subject is the anchor, guarded only by `!VerifyReads(stored)`
  (the NEW reading's outcome), and words it as unread. A found-against citation stays found against
  after a later `unreachable` or not-the-work reading (`anchorStanding.stands`, `anchorchecks.go:84-94`:
  only the same seat's later reading that READS replaces its check), so that reading would print "stays
  unread … retires the citation or cites a source a lens can read" and, with no carrier, the owed-mint
  sentence. After the change:
  - the guard is the anchor alone: after every verification of a citation the verb reads the anchor's
    standing failed checks (`PassBlockers`, kind `failed_check`, Subject = anchor), every one, not the
    last match;
  - `FailedCheck *failedCheckWords` becomes `FailedChecks []failedCheckWords` (JSON `failed_checks`),
    each carrying `state` (`Blocker.CheckState()`), `owner`, `since` and `answered_by`. A found-against
    check whose `since` is this reading's own event is left out: its words are the `Gap` sentence;
  - `words()` switches on `state`. `unread`: today's three sentences, unchanged ("stays unread", the
    carrier's remedy, the owner's owed mint, another lens's mint). `found_against`: "citation C stays
    found against: L's reading at event N finds the source against this sentence, and PASS is refused
    while the report carries it; gap G carries it to blue, which edits the sentence or retires or
    replaces the citation" (no carrier: "no gap names it; blue edits the sentence or retires or replaces
    the citation"). When this reading backs the sentence and is not L's: "your backing reading beside
    L's standing one is a disagreement and clears nothing". When this reading reads nothing: the clause
    opens "this reading reads nothing of the source, and". No `found_against` sentence names an owed
    mint, the budget, or "read nothing first".
  - Tested by `TestAReadingOfAFoundAgainstCitationIsWordedByItsState` (§III.10).
- **Crash-retry and the repeat refusal.** <!-- CHANGED F13b --> `verify` takes no `--key`. A citation reading
  keys on its `anchor` per seat and sitting (`record.go:540-549`); a found-source reading takes the
  ordinal key and is held to one reading per (url, sentence) per seat and sitting by
  `requireNotARetriedReading` (§III.1, F13b). `seat.KeysOn(c, "anchor")`; the found-source form's answer
  is the unkeyed `Verify` row of `superseders` (§III.2d's deletions table), which `AnswersFor`
  (`cli/seat/answer.go:76-96`) renders on every page because it carries no key. A crash before the commit
  leaves nothing (one transaction, R8), and the retry writes the act afresh. A retry after the commit is
  the same seat, sentence and source in the same sitting: a citation reading meets the duplicate-key
  refusal; a found-source reading meets `requireNotARetriedReading`, the arm's first rule, which names
  the stored reading's event, its outcome and its carrying gap. That refusal is how a crash-retried
  reading learns its first write landed; it is not an idempotent return. A reading of a later sitting
  with the same outcome, while that gap is open, meets the repeat refusal above. (F13)
  `ExistingCorroboration` (`citationid.go:439-473`) and the verb's retry branch (`verify.go:460-484`:
  `sameReading`, `statedOtherwise` on a corroboration, `seat.PlaceOwed` of a label) are DELETED: no
  reading mints a label to key on. A found-source reading that mints nothing (`absent`, F5;
  `unreachable`; any outcome of a `not_the_work` copy) is held by the same rule: a second such reading of
  that url and that sentence in the sitting is refused naming the first; of that url and another
  sentence it is recorded (F13's url-keyed residue is gone). Tested by row (h) of
  `TestAReadingOfACopyThatIsNotTheWorkMintsNothing`, by `TestOneSourceBacksManySentences` and
  `TestOneSentenceReadAgainstManySources` (§III.2d), and by `TestARetriedNegativeReadingIsRefusedNamingItsGap`
  (`cli/keyedretry_test.go`, §III.10), which gains a `verify --url` row whose refusal names the stored
  reading's event. <!-- CHANGED F13b -->

### III.2a Where a reading's gap stands — by the anchor's identity (F9) [NEW] (contract: event schema)

gblock's ruling (2026-10-30, his words): "I think if only text is provided out [it] should be refused
but it's fair to point at an anchor using the right flag".

**How each reading verb names its subject today** (read at `9ee7ad95e`):
- `verify` takes no text. `--anchor` is required (`verify.go:113-114`), checked to name a citation
  (`record.CitationExists`), and the claim is read by the anchor's identity:
  `bluedoc.SentenceAround(report, anchor.Token(id))` (`verify.go:129-150`). It places nothing today
  (`quote = ""` for a non-citing verb, `verify.go:511-514`).
- `corroborate` (deleted by F13; its subject is `verify --url`'s) takes text and no anchor: `--quote` is the report's sentence (`verify.go:291`), placed
  by text (`seat.AppendPlaced` → `Places` → `anchortext.Attach` → `LocateOnce`, `verbs.go:1384-1417`,
  `anchortext.go:315-336`), which refuses a quote the report holds more than once (`AmbiguousError`).
  It has no anchor flag: the source it read is one blue never cited.

**So the path that places is `verify --anchor`'s, and the path that refuses is `verify --url`'s**
(F13: one verb, two subjects) <!-- CHANGED F13 -->. A `verify --anchor`
names its citation by `--anchor`, the right flag; its gap is placed at the end of the sentence that
anchor sits in, by the anchor's identity. A `verify --url` names text only; its gap is placed at its
`--quote` by text, as a plain mint's is, and a quote the report holds more than once
is refused, nothing recorded — `claimRefusal` (`verify.go:717-728`) words it through `anchortext.Refusal`
(`anchortext.go:508-528`), which names each occurrence's section (`WhereAmbiguous`) and tells the seat
to quote the sentence with the text before it in its paragraph so the quote occurs once. Since #1654
`Refusal` also words `ErrInQuotation` ("an anchor sits after the quotation it ends on: quote through
the closing mark", `:524-525`), raised where `Attach` (doc `:547`, func `:554-568`) finds the quote stopping short of a
closing double quotation mark; `claimRefusal` inherits that arm for a `--url` reading's gap, and it is
the same sentence the seat gets today. The `--url` form takes no `--anchor` (cobra's mutually exclusive
group, §III.2d). <!-- CHANGED F13 -->

**The mechanism — `Anchor.sentence_of`** (§III.1). Today an Anchor is placed by text: `location` is
the quote, and replay's `InsertAnchor` (`anchortext.go:534-545`, from `reportproj/render.go:123`)
places the marker at the location's FIRST occurrence, past a closing quotation mark the quote carries
(`placement`, `:363-380`, #1654); the write-time guard (`Attach`) is what refuses a second occurrence. A
reading's gap placed that way — the Mint's quote being the anchor-stripped cited sentence — is refused
whole whenever the report holds that sentence's words twice, and `requireOneGapPerClassAndSubject`
(`onegapperclass.go:82-85`) returns nil on the same failure, so the refusal reached the seat as a
placement error on a reading it had every right to record. The anchor token itself is unique by
construction (`Anchor.id` is `unique`), and the matcher skips tokens (`annotationLen`,
`anchortext.go:50`, its doc `:43-49`), so no text location can name it. The placement therefore names the anchor, and
puts the gap's token where every gap's token stands — at the end of its sentence, never beside a
citation's token in mid-sentence (a clause citation):
- [NEW] `anchortext.AttachToSentenceOf(doc, id, of string) (string, error)`: doc with `Token(id)` at the
  end of the sentence (`anchor.Sentences`) holding `Token(of)` — at `placement` of that sentence's
  content end, the offset `Attach` gives a quote of that whole sentence at that occurrence, so a token
  of `of`'s at the sentence's end stays after the gap's as it would beside a plain mint's. Refuses, each
  a sentinel: `of`'s token absent from doc (`ErrAnchorGone`); `of` standing bare
  (`claimcount.BareAnchorIDs`, as `citedSentence` refuses it, `ErrAnchorBare`); the sentence end inside
  a fence or on a heading (`ErrInFence`, `ErrOnHeading`, as `Attach`). It never stops inside a
  quotation (`placement` carries it past the closing mark), so `ErrInQuotation` cannot arise. Never
  searches text. Name grep at `9ee7ad95e`: `AttachToSentenceOf`, `ErrAnchorGone`, `ErrAnchorBare` 0 hits.
- [NEW] `seat.PlacesToSentenceOf(id, of, refuse)` beside `seat.Places` (`verbs.go:1389`), the guard
  asked of the report the pair lands on (`record.CheckOnReport`), so a retire of the citation committing
  in between refuses the act, nothing recorded.
- [MODIFY] `record.AppendPlaced` (`record.go:664-668`) and the reading's placed append write
  `Anchor{id, sentence_of}`; `record.go:941-951` admits exactly one of `location` and `sentence_of`.
- [MODIFY] the `report_op` view (`recordsql/views.go:809-821`): a fifth row kind, `'sentence_of'`,
  `a = sentence_of`, `b = id`, from `anchor` rows with `sentence_of` set. The `'insert'` arm's filter
  (`location != ''`) would otherwise drop such an Anchor from the render with no error — the silent
  zero — so the new arm is required, and its comment (`:793-808`) [REWRITE] names five kinds. The
  `ReportOp` doc (`record/queries.go:251-256`, `"append"`, `"edit"`, `"insert"` or `"remove"`) [REWRITE] with it.
  The `report_op` readers that key on a kind at this base hold: `PartStanding` (`queries.go:381`, `"kind"
  != 'append'`) reads a sentence_of row as a change to the report, as it reads an insert, so it closes the door
  on a lane still writing; `migrate/order.go:118` reads `"append"` alone; `reportVersionQuery`
  (`landingreport.go:28`) counts every row.
- [MODIFY] `reportproj.mutationOf` (`render.go:295-308`): `'sentence_of'` → `sentenceOfMut{of, marker}`,
  which leaves text unchanged where the marker already stands (as `insertMut` does, `render.go:113-131`)
  and otherwise applies `AttachToSentenceOf`; a missing or bare `of` token is a replay error naming the
  mutation — loud, never a skip. Its `describe()` names both ids. The default arm's error
  (`render.go:306`, "the report_op view emits only append, edit, insert and remove") [REWRITE] names the
  five kinds. The fold stays a left fold (`record.reportRenderer`, `TestTheReportReplayIsAFold`):
  `sentenceOfMut.apply` reads only the text it is handed.
- [MODIFY] the one-gap rule's placement (`onegapperclass.go:82-85`): for a Mint paired with a
  sentence_of Anchor, `placed` is `AttachToSentenceOf(report, gapID, of)`; its failure is a refusal,
  never the `return nil` the text path keeps (there the act's own `Places` guard has already refused,
  since guards run before validation, `record.go` `Append` doc). `gapSentenceIn` (`viewjson.go:2019`)
  then reads the sentence the marker sits in — the cited sentence, whatever text repeats elsewhere.
- [MODIFY] migrate's placer (`migrate/gapplace.go:78-79`, the Anchor arm, and `places`, `:104-115`): an
  Anchor with `sentence_of` is checked by `AttachToSentenceOf` on the migrated render, never by
  `places(id, location)`, whose empty location would pass it unchecked. Its Mint arm (`:55-70`) skips a
  gap the stream anchors itself (`p.anchored`), which every reading's gap is. No archive holds one; the
  next migration of a record written at this epoch does. Tested by
  `TestMigrateChecksAReadingsGapBySentenceOf` (§III.9).
- `verify --anchor`'s Anchor for its gap is `{id: G, sentence_of: <--anchor>}`. `verify --url`'s is
  `{id: G, location: <--quote>}`, exactly as `lens mint` places one. <!-- CHANGED F13 -->

**`--new` on a reading — blue's accept locates the span by the gap's anchor (RULED, gblock
2026-10-30).** A verified fix is one blue's accept applies over the report with the gap placed
(`mint.go:350-383`, `verifiedFix` → `blue.PlanAccept`, `cli/blue/edit.go:335-341` →
`record.ProposalOver`, `estoppel.go:175-211`). Today `ProposalOver` locates the gap's location by its
text (`LocateOnce`, `:191`): over a sentence the report holds twice it returns the recorded pair, and the
accept's plan (`planTyped` → `validateEdit` → `reportproj.PlanSplice`, `plan.go:79`, the ordinary locate
that skips the anchor layer) refuses it as ambiguous. After the change:
- [MODIFY] `record.ProposalOver(report, gapID, loc, fix)`: it finds `Token(gapID)` in report (one place,
  by construction) and, among the occurrences of `loc` (the walk `LocateOnce` already makes for
  `AmbiguousError`), takes the one whose anchor run after its trailing punctuation holds that token;
  `old` is that occurrence's literal bytes through the run, `new` as today. No such occurrence (the
  token gone, or a run that does not hold it): the recorded pair, as today. This holds for every gap, a
  plain mint's included: a quote that became repeated by a later edit is located by its gap's token
  too.
- [MODIFY] the accept's plan (`blue.PlanAccept` → `planTyped`, `edit.go:335-349`): the pair carries the
  gap's token, which the report holds once, so where the ordinary locate of `old` is ambiguous the
  accept locates it byte for byte (`bluedoc.LocateLiteral`, the exact-span locate) and plans the edit
  with `exact` true; `edit --accept` records it with `exact_span`, which replay reads and never
  re-decides (`reportproj/render.go:89-97`; migrate's `locate`, `gapplace.go:159-166`). A typed
  `blue edit` keeps the ordinary locate: its quote is the seat's text, not a gap's identity.
- The board's pair (`viewjson.go:1995`, `locateGap`) and estoppel's comparison (`proposalAt`,
  `estoppel.go:159-164`, read by `ProposalAppliedVerbatim`, `:220-232`, `landing.go:109`, `edit.go:206`)
  read the same `ProposalOver`, so the pair the board serves, the pair the accept applies and the pair
  estoppel compares are one.
- [MODIFY] `verifiedFix` (`mint.go:350-383`) builds `placed` by the gap's placement — `Attach` for a
  `--quote` (`mint`, `verify --url`), `AttachToSentenceOf` for a `verify --anchor` — and its refusal
  (`:370-381`, "with this gap's anchor at the end of --quote", "keep the sentence …") names the subject
  the call took: `--quote` for `mint` and `verify --url`, "the sentence your --anchor's token sits in"
  for `verify --anchor`, which takes no `--quote`. <!-- CHANGED F13 --> The dropped-anchor clause #1663 smoke D7 added (`mint.go:364-372`, `carry`:
  "write <token> into --new after the text it marks — --quote stays as it is") names `--quote` too, and
  takes the verb's subject the same way (row (g) of `TestAVerifyOfARepeatedSentencePlacesItsGapByTheAnchor`
  asserts no "--quote" in a `verify`'s refusal). `help/verify.md` says it in the same words. "`--new` on a repeated sentence is
  refused" is not part of this design: the accept edits the cited occurrence.

**Sweep — every reading-act site that places a gap or reads where one stands, and every text-locate of a
proposal span**, run at `9ee7ad95e` (45 hits, `~/.claude/scratch/plan1647-r10/new-place1.txt, new-place2.txt`):
`grep -rn -E 'anchortext\.(Attach|LocateOnce|InsertAnchor)|seat\.(AppendPlaced|Places|PlaceOwed)|record\.AppendPlaced|reportproj\.Place|gapSentenceIn|ProposalOver|SentenceAround' plugins/frank-exchange-of-views/tools/internal plugins/frank-exchange-of-views/tools/releasegate | grep -v _test.go`,
plus `grep -rn -E 'PlanAccept\(|proposalAt\(|ProposalAppliedVerbatim\(|PlanSplice\(|LocateUnique(Replacing)?\(' plugins/frank-exchange-of-views/tools/internal | grep -v _test.go`,
every hit read; the reading-act ones, the proposal-span ones, and those the new placement reaches:

| Site | Today | After | Disposition |
|---|---|---|---|
| `cli/lens/verify.go:95-120` `newVerify` / `:129-151` `citedSentence` | anchor required; claim read by the anchor's identity | unchanged; the anchor is also what places the gap | KEEP |
| `cli/lens/verify.go:502-515` `writeVerify` placement | `quote=""` for verify; corroborate's quote placed by text | `--anchor`: the reading's append writes `Anchor{G, sentence_of}`; `--url`: `Anchor{G, location}` by text, ambiguity refused; no label or citation Anchor on either (F13) <!-- CHANGED F13 --> | REWRITE |
| `cli/lens/verify.go:717-728` `claimRefusal` | corroborate's placement refusal through `anchortext.Refusal` (ambiguity, crossing, word split, heading, `ErrInQuotation`) | the same refusal for a `verify --url` reading's gap <!-- CHANGED F13 --> | KEEP |
| `cli/seat/verbs.go:1384-1417` `Places`, `AppendPlaced` | text placement guard | gains `PlacesToSentenceOf`; text path unchanged | MODIFY |
| `cli/seat/verbs.go:1452-1474` `PlaceOwed`; its callers `verify.go:481`, `cite.go:88`, `finding.go:93`, `mint.go:177`, `compute.go:114` | re-places a migrated act's stored location | not reached by a reading's gap (its Anchor is in its own write) | KEEP |
| `record/record.go:664-668` `AppendPlaced`, `:941-951` Anchor validation and its comment `:942-943` | location required; "at its quote's end" | location XOR sentence_of, comment names both | MODIFY |
| `record/recordpb/record.proto:1179-1187` `Anchor` and its comment | "at the end of its quote `location`" | `sentence_of = 7`, comment names both placements | MODIFY |
| `record/onegapperclass.go:82-92` | `Attach` by text; nil on failure | sentence_of placement for a paired Anchor, failure refused | MODIFY |
| `record/recordsql/views.go:793-821` `report_op`; `record/queries.go:251-256` `ReportOp` doc | four kinds; insert rows need a location | `'sentence_of'` arm; doc names five kinds | MODIFY |
| `record/queries.go:381` `PartStanding`; `record/migrate/order.go:118`; `record/landingreport.go:28` | read `"kind"` against `'append'`, or count rows | a sentence_of row is a change, as an insert is | KEEP |
| `reportproj/render.go:113-131, :295-308` | insert by location | `sentenceOfMut` | MODIFY |
| `record/migrate/gapplace.go:55-70, :78-79, :104-115` | places archived anchors by location | checks sentence_of placements by identity | MODIFY |
| `record/migrate/replay.go:222-501` (the replay loop) | writes each source event alone | re-drives a reading act's Mint, `{G, sentence_of}` Anchor and Verify as one act, each member keeping its own per-event duties (§III.9 item 6) | MODIFY |
| `record/batch.go:82-114` `Batch.Append` | one body per savepoint | [NEW] `Batch.AppendBodies` beside it, the act's bodies in one savepoint (§III.9 item 6) | MODIFY |
| `record/id.go:65-91` `unplacedLocationAt` | the location of an act no Anchor places | a reading's gap always has its Anchor; it also reads `"label","claim" FROM "verify"` (`:88`), a column F13 reserves | REWRITE (F13b): the `verify` arm of the UNION is DELETED — no reading places an anchor; the rest KEEP <!-- CHANGED F13b --> |
| `record/refs.go:130` `requireAboutAnchor`; `record/refs.go:205` `HoldsAnchor` | an about-anchor mint's subject; whether the record holds an anchor id (migrate's `proseRefs` asks it, `migrate/proserefs.go:74`) | each UNIONs `"label" FROM "verify"` | REWRITE (F13b): the `verify` arm DELETED in both; a label is no anchor any act holds. Gates: `TestWriteLookupsAgreeWithTheFoldsTheyReplaced`, every about-anchor mint test, `TestEveryArchivedRunMigrates` (no such column) <!-- CHANGED F13b --> |
| `consistency/consistency.go:254-256, :521` | reads `Anchor.id` | id only | KEEP |
| `consistency/consistency.go:272` | `*recordpb.Verify`'s `GetLabel()` as a placed id | the field is reserved | DELETE the arm's `placedIDs` read; its `verifyEvents` count KEEPs (F13b; gate: compile) <!-- CHANGED F13b --> |
| `record/estoppel.go:175-211` `ProposalOver` | text-located span | located by the gap's token | MODIFY |
| `record/estoppel.go:159-164, :220-232` `proposalAt`, `ProposalAppliedVerbatim`; `record/landing.go:109`; `cli/blue/edit.go:206` | read `ProposalOver` | unchanged code, the identity-located pair | KEEP |
| `record/viewjson.go:1995-2031` `locateGap`, `gapSentenceIn` | `ProposalOver`; sentence by token | unchanged code, the identity-located pair | KEEP |
| `cli/blue/edit.go:145` (`--accept`), `:335-349` `PlanAccept`/`planTyped`, `:401-445` `validateEdit` | ordinary locate of `old` | exact-span locate of the token-carrying `old` where the ordinary one is ambiguous; `validateEdit`'s stay-put check (`:418`) is skipped on an exact span, as today | MODIFY (`PlanAccept`/`planTyped`) |
| `reportproj/plan.go:55-115` `PlanSplice`; `bluedoc.go:45-74` `LocateUnique`, `LocateUniqueReplacing`; `cli/blue/edit.go:418, :465` | a typed edit's locate | unchanged: a typed edit's quote is the seat's | KEEP |
| `reportproj/render.go:89-105` (an edit's replay); `migrate/gapplace.go:159-166` `locate` | exact span byte for byte | unchanged; the accept's exact span replays through it | KEEP |
| `cli/lens/mint.go:350-383` `verifiedFix` | `Attach`; refusal speaks of `--quote` | placement by the verb's subject; refusal names it | MODIFY |
| `bluedoc.go:204, :343-370` `SentenceAround`, `:400`; `cli/blue/retire.go:180-197` | sentence by token; an edit's reopen check | identity already | KEEP |
| `cli/motion/verbs.go:170` (#1652, `seat.Places("", q, groundsQuoteRefusal)`) | a docket motion's grounds quote, checked against the report | another sense: a quote of report text, no gap placed | KEEP |
| `cli/blue/compute.go:114, :200`; `cli/blue/cite.go:88, :165`; `cli/lens/finding.go:93, :121`; `cli/lens/mint.go:111, :177, :212` | their own acts' text placement | not reached | KEEP |
| `record/recordsql/store.go:536` | a comment naming `seat.Places` | — | KEEP |
| `releasegate/fuzz/fuzz_test.go` reading drive | — | drives a cited sentence held twice, and blue's accept of its gap (§III.11) | REWRITE |

The new mechanism's own readers — `sentence_of`, `AttachToSentenceOf`, `PlacesToSentenceOf`,
`sentenceOfMut`, the `'sentence_of'` op kind, the kind readers, the identity locate in `ProposalOver` and the accept's
exact span — are the rows above and no others; §V.4 re-runs
`grep -rn 'sentence_of\|SentenceOf\|sentenceOfMut\|ErrAnchorGone\|ErrAnchorBare'` over the tools tree,
and every hit must be one of them or a test.

Tested by `TestAVerifyOfARepeatedSentencePlacesItsGapByTheAnchor` (§III.10), driven through the verbs
as a seat and blue run them.

### III.2b Which readings mint — the "negative reading" class sweep (GAP 1)

A reading's outcome alone does not say it read the work: an outcome of a `not_the_work` copy (a
sign-in page, a sales page) is no reading (`VerifyReads`, `anchorchecks.go:284-289`; `help/verify.md:11`;
the ledger puts its anchor in `unread`). That rests on `source_completeness`, and today only a
corroboration carries it: the stamp (`verify.go:492-501`) runs where `body.GetUrl() != ""`, and a
`verify`'s body has no url (`newVerify`, `verify.go:97`; its source is `citedURL`'s, `:153-172`), so a
`verify` of a `not_the_work` copy reads as a reading and its citation leaves `unread`. `help/verify.md:11`
already states the model the code does not implement; the only test is a synthetic `not_the_work`
Verify at the record layer (`record/reopened_test.go:260`), and no CLI test drives it. A site that
decides "negative reading" from `refutes`/`absent` alone, or reads an unstamped `verify`, would demand
gap flags on such a reading and mint a `claim_contradicts_source` gap no failed check carries —
closable freely — while the `unread` check separately owes its about-anchor gap. `verify --as absent`
of such a copy is already refused at `verify.go:379-389` (`WithoutBody`, read from the cache entry the
verb looks up by `source`, so it works for `verify` today), so the live hole is `refutes` (both forms)
and every backing outcome; `absent` of an uncached copy is a reading (`not_asked` reads).

**The ruling (gblock, 2026-10-30): `verify` stamps it.** §III.2's first bullet: every reading stamps the
`source_completeness` of the cache entry for the source it names — the cited url for `verify` — and
every reading of a `not_the_work` copy is unread, a backing one included.

**The property:** a reading for which `VerifyReads` is false — `unreachable`, or ANY outcome of a copy
stamped `not_the_work`, `supports`/`supports_with_bridge`/`weak` included, by `verify --anchor` or
`verify --url` — mints no gap, carries no `carried_by` or `carrier_minted`, has its gap flags refused,
and holds the PASS the way an unread reading does: for `--anchor`, the citation's `unread` failed check,
owned by its reader, and the about-anchor gap that check owes; for `--url`, the reading alone (it reads
no citation, so nothing is held). <!-- CHANGED F13 -->

**Sweep**, run over this plan and the code it specifies — every site that decides, counts or words a
negative reading — with `grep -n -i -E 'negative (reading|outcome|half|corroboration)|refutes|absent|gap_classes|found.against|finds.against'`
over this plan and `grep -rn -E 'SOURCE_OUTCOME_(REFUTES|ABSENT)|contradicts\(|VerifyReads'` over
`tools/internal`, `tools/releasegate` (no test files) at `9ee7ad95e` (22 hits, `~/.claude/scratch/plan1647-r10/new-cls.txt`), and — for the stamp the ruling adds — every
Verify writer (`grep -rn 'recordpb.Verify{' tools --include=*.go | grep -v _test`) and every reader of
`source_completeness` (`grep -rn -i 'source_completeness\|SourceCompleteness' plugins/frank-exchange-of-views scripts docs`, less
`fetchcache/` and the cite's own readers), every hit read:

| Site | Decided by | Disposition |
|---|---|---|
| Gap-flag requirement (§III.2) | `--as` | REWRITE: `ReadingGapClasses` after the completeness stamp |
| Write-path rule "a negative reading names its gap" (§III.1, `record.go` Verify case) | outcome | REWRITE: `ReadingGapClasses` of the stored body; either field on a non-minting reading refused |
| Repeat rule's negative arm (§III.2) | outcome | REWRITE: `FindsAgainst` of both readings for a citation; `ReadingGapClasses` non-empty for a found-source reading (F13) <!-- CHANGED F13 --> |
| Reading result's `Gap` words (§III.2) | outcome | REWRITE: set from the stored `carried_by` only |
| Held-gap / one-gap path in the reading append (§III.2) | reached only from a minting reading | KEEP (downstream of the requirement) |
| Budget exemption (§III.5) | the act's paired bodies; `carrier_minted` | KEEP: no Mint is paired with a non-minting reading, and the write rule refuses `carrier_minted` on one |
| Anchor ledger `anchorchecks.go:260-261`, `CitationsFoundAgainst`, `stateFoundAgainst` | `reads && contradicts` | REWRITE: `FindsAgainst` (same truth table; `contradicts` deleted) |
| `failedChecksOf` found-against arm (§III.3) | ledger state | KEEP (reads-gated by the ledger) |
| The footnote decision (`verify.go:426`, `mayCite == cites && backsTheClaim(...)`) and `backsTheClaim` (`:336-350`) | outcome, positive half | DELETE (F13): no reading mints a footnote, so neither decision nor helper has a reader (§III.2d) <!-- CHANGED F13 --> |
| Live Verify writers: `verify.go:97` (`newVerify`) and `:280` (`newCorroborate`, DELETED by F13: `newVerify` takes the `--url` form), through `writeVerify` | — | MODIFY: the stamp in `writeVerify` keys on `source`, so both forms stamp; migrate's archived Verify (`migrate/entries.go`) carries what the archive held — an archived `verify` has none and keeps reading (residue, §III.9) <!-- CHANGED F13 --> |
| The stamp `verify.go:492-501` | `body.GetUrl() != ""` | REWRITE: `source != ""`, before every decision that reads it <!-- CHANGED F13 --> |
| `anchorchecks.go:259-266` ledger Verify arm; `VerifyReads` `:284-289` | the stored stamp | KEEP (code): reads the stamp `verify` now writes; the `VerifyReads` comment `:278-283` [REWRITE] (§III.3) |
| `evidenceview.go:270-273, :422` (a verification row's `source_completeness`) | the stored stamp; empty on every `verify` today | KEEP (code): every verification row now carries it; `show evidence` goldens that print a `verify` row REGENERATE, each changed line read |
| `evidenceview.go:499`, `citationid.go:46-49, :89, :106`, `report/assemble.go:155, :202` (a source row's, a footnote's, the assembly's abstract note) | a cite's or a labelled corroboration's stamp | KEEP for a cite; the labelled-corroboration arms are DELETED (F13, §III.2d) <!-- CHANGED F13 --> |
| `record.proto:1392-1394` (`Verify.source_completeness` comment) | "which part of the work red's copy is" | KEEP (true of every reading now) |
| `docs/citations-and-bibliography.md:178` ("Every verdict naming a url is stamped") | a corroboration only, in fact | REWRITE: every reading is stamped, `verify`'s from its citation's url; any reading of a copy not the work reads nothing and leaves the citation unread ("reading", not "verdict": the registry's verdict is the chair's, GAP 1 of round 8) |
| `help/verify.md:11` | states the model | KEEP its unread paragraph's model (now true); its "does not close" clause REWRITEs (F8, §III.7) |
| `record/reopened_test.go:260` (synthetic `not_the_work` Verify) | record layer only | KEEP; the CLI rows of `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` drive the verb |
| `cli/corroborate_test.go:301-321` (a corroboration's stamp, `abstract`, `not_asked`) | corroborate only | MOVE (F13): the file is deleted with the verb, and these three stamp rows move to `verify --url` rows of `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` <!-- CHANGED F13 --> |
| `verify.go:379-389` `absent` on a copy without its body | outcome + completeness | KEEP |
| `verify.go:533` failed-check words guard | `!VerifyReads(stored)` | REWRITE by state (§III.2), already reads-aware |
| Fuzz reading drive (§III.11) | outcome | REWRITE: draws gap flags from the record's stamp of the copy the reading resolves to, and drives a `not_the_work` copy |
| Fuzz gates "negative readings that minted / refused a mint" (§III.11) | outcome | REWRITE: counted from stored `carried_by`, never from `--as` |
| §V.8 smoke query | outcome | REWRITE: `FindsAgainst` rows only, plus the converse query |
| Success criterion 1 (§I) | outcome | REWRITE: "of a copy that is the work" |
| Help `verify.md`, `mint.md` (`corroborate.md` DELETED, F13); agentgen fragments; `terms.json` `failed check` (§III.7) | prose "a negative reading mints" | REWRITE: "a reading of the work that finds the source against the sentence mints", and a reading of a source the lens found mints `unsourced_claim` or `claim_contradicts_source`; `verify.md` states that any outcome of a `not_the_work` copy, like `unreachable`, mints nothing and leaves the citation unread <!-- CHANGED F13 --> |
| Migrate (§III.9) archived negative readings | ledger | KEEP: archived readings carry no carrier; their state is the ledger's, reads-gated |
| `TestEveryNegativeOutcomeNamesItsClasses` | the facet itself | KEEP: it tests the declaration, which is per outcome by design |
| `record/evidenceview.go:287-289` `EvidenceVerificationJSON.Refuted()` (`Outcome == "refutes" \|\| "absent"`) and its comment `:393` | outcome STRING | REWRITE (GAP 2): renamed `FindsAgainst()`, defined by the one predicate — `record.FindsAgainst` is written over (outcome, completeness), and the row's method calls it with the row's `Outcome` and `SourceCompleteness`, so a `not_the_work` row is not found against |
| `capture/capture.go:483` (the assembly screen's "%d ever found against by red", through `Refuted`) | outcome string, through the row | REWRITE: calls `FindsAgainst()`; capture moves from KEEP to REWRITE (§III.12) |
| `cli/lens/verify.go:694-712` `human()` | outcome string | REWRITE (§III.2: reads before outcome) |
| `record/evidenceview_test.go:269`, `record/redcitation_test.go:337` (`Refuted()` calls) | outcome string | REWRITE: call `FindsAgainst()`; `evidenceview_test.go:269` gains a `not_the_work` `refutes` row asserting it false |

The outcome-STRING sweep (GAP 2; the enum-constant grep above misses a compare of the word), run at
`9ee7ad95e`, no tests: `grep -rn '"refutes"\|"absent"\|Refuted()' tools/internal tools/cmd tools/releasegate`
→ 7 hits: `evidenceview.go:287, :288, :393`, `verify.go:702, :704` (rows above), and
`hookprovenance.go:43, :91` (`"absent"`, a hook binary's install state: another sense, KEEP). Property:
every site that counts or words a found-against reading agrees with `record.FindsAgainst`. Tested by the
`not_the_work` `refutes` row of `evidenceview_test.go` and row (b) of
`TestAReadingOfACopyThatIsNotTheWorkMintsNothing` (`Refuted`'s successor false; no "REFUTES" in the human
text). Partner §V.5.jj.

Tested by `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` (§III.10).

### III.2c The link from a gap to its source (F12) [MODIFY] (contract: `show evidence` JSON; docket.md)

gblock's ruling (2026-11-01, round 8's fork): "Print the link." The round-8 verdict states the fork as
"found_by goes; a corroboration's gap has no link to its reading" (`~/.claude/scratch/plan1647/round8-verdict.md:12`):
`found_by` is not set on a reading's gap (§III.2's field table, UNSET), and the F12 Forks bullet that read
"`found_by` goes on a reading's gap" misread "goes" — this revision corrects the bullet, not the table
(the phase-1 builder's open item) <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->. So nothing a seat or reader sees ties the gap to the source
read for its sentence but `carried_by`, and two readers print it:
- [MODIFY] `record/evidenceview.go` `EvidenceVerificationJSON` (`:230-281`) gains `CarriedBy string
  json:"carried_by,omitempty"`, set from the stored Verify's `carried_by` where the row is built
  (`:403-420`). Every verification row carries it — under its source for a `verify --anchor`, under
  `independent` for a `verify --url` (`:446-450`, the split on the anchor; the key keeps its name: the
  readings of sources a lens found, which blue never cited). A reading no gap carries omits the key.
  `show evidence`'s help text names the field beside `source_completeness`, and says it is how blue finds
  the source a lens found for a gap it is answering (F13) <!-- CHANGED F13 -->.
- [MODIFY] `report/assemble.go`: docket.md's open-gap block (the block at `:1047-1057`, beside
  `mintReason` and `foundBy`) and its closure index (`:1142`) gain one line per Verify whose
  `carried_by` is the gap, in stream order: "read for the sentence: <url> — <seat>'s `<outcome>`
  reading, event N" (the citation's cited url for `verify --anchor`, `--url` for `verify --url`; "read
  for", since a backing found-source reading is not read against the sentence, F13) <!-- CHANGED F13 -->, read from the
  record's Verify rows by `carried_by`, never from the gap's text. report.md prints no such line, as it
  prints no gap's argument.
- Readers swept: `evidenceview.go`'s builders and its goldens (REGENERATE; each changed line read),
  `report/assemble.go` and its tests (`assemble_test` KEEP: no fixture holds a reading's gap), the
  `show evidence` help page. Name grep at `9ee7ad95e`: `CarriedBy` 0 hits.
- Tested by `TestAReadingsGapNamesItsSource` (§III.10); partner §V.5.ll.

### III.2d F13 — One reading verb; a source a lens found mints its gap; no reading places a citation [NEW] (contract: CLI verb set; event schema; prompt a seat is told) <!-- CHANGED F13 -->

**The ruling** (gblock, 2026-11-03, his words, in the #1657 session): "Feels like we created a weird
edge case here. Could red have simply created a gap starting that the statement was unsupported and had
blue go looking for a citation? Are we better off dropping this as a snowflake and replacing it with a
gap of a new class for needs-supporting-evidence? Or something? Do some archaeology". He picked option B,
whose text is that session's wording, not his: red mints `unsourced_claim` when its source supports an
uncited sentence, or `claim_contradicts_source` when its source contradicts the sentence, with its
source carried on the gap; blue answers with `cite --answers`; a lens verifies blue's cite; red-placed
`C-` anchors disappear. The two sessions agreed to fold it into this plan (one concept with deleting the
contradiction: the labelled corroboration was red's evidence workaround from before a lens could mint —
"A LENS CANNOT MINT", `plans/historical/red-citations.md:76`, 2026-08-22; lenses mint since 09-08) and
agreed the design below; gblock asked for a targeted one-round audit of the changed rows before phase 2.

**Archaeology** (the #1657 session's, `~/.claude/scratch/anchor-lifecycle/corroborate/`, 34 archives;
re-measured here for the labelled ones, `~/.claude/scratch/plan1647-f13/`): 22 corroborations in 10
archives, none after 10-27; outcomes 12 `supports`, 5 `supports_with_bridge`, 1 `weak`, 4 `refutes`, 0
`absent`; 15 placed a red `C-`; 13 of those were never read by anyone but the placer; 9 reached the
archived report's Bibliography, 8 of them a url blue never cited; of the 4 refutes, 3 went
finding → gap → edit and 1 (m23) stranded at UNVERIFIED (GAP 2's residue, §III.9).

**Verb decision: ONE verb, `verify`, taking exactly one of `--anchor` and `--url`; `corroborate` is
deleted with no alias (no-compat).** This revision decides it; the reasons:
- The two verbs exist because "cobra cannot express 'required only when'" (`verify.go:84-91`, the rule
  `cli/motion/command.go:15-22` states for a value-discerned subject). That objection is to a contract
  keyed on a flag's VALUE. Here the discriminant is WHICH flag is given, and cobra's flag groups state it
  at parse, before the handler: `MarkFlagsOneRequired(anchor, url)`, `MarkFlagsMutuallyExclusive(anchor,
  url)`, `(anchor, quote)`, `(anchor, title)`, `(url, page)`, and `MarkFlagsRequiredTogether(url, title,
  quote)`; `seat.RefuseFlagGroup` (`cli/seat/menu.go:257-275`) already renders a violated group with each
  flag's usage line. Nothing is policed in RunE.
- With the footnote gone, the two contracts differ in nothing but the subject: both write the same
  `Verify` with the same axes, the same gap flags, the same `--reason` duty and the same act
  (`AppendReading`). The record already splits on the field — an empty `Verify.anchor` is a reading of a
  source the lens found (`record.proto:1355-1356`) — so one verb makes the verb split and the record
  split one split: one help page, one `ReasonIs` row, one class field, one seatprobe attribution, one
  fuzz drive.
- Rejected: (b) keep a second verb, renamed — two verbs writing one body, every reader of "which verb"
  re-deriving what `anchor` already says; (c) `verify --url` with `--anchor` optional and policed in
  RunE — the shape the motion rule forbids; (d) `mint --source <url>` — a mint carries no reading, so the
  source would reach the record as prose, the defect facts-are-fields names.
- Flag text, `--url` form: `--url`, `--title` (`flags.DescURL`, `flags.DescTitle`), and `--quote` ("the
  REPORT's sentence your source backs or refutes, under every --as value, never the source's",
  `verify.go:291`'s text without its footnote clause). `seat.KeysOn(c, "anchor", "carried_by", "url")`.
  `--page` stays with `--anchor`. `seat.Records(…, "verify")` and every `corroborate` registration go.

**What a reading of a source the lens found mints** (declared on the enum, §III.1
`found_source_gap_classes`, read through `ReadingGapClasses`):

| `--as` on `verify --url`, of a copy that is the work | The act |
|---|---|
| `supports`, `supports_with_bridge`, `weak` | mints `unsourced_claim` on `--quote` — refused where an evidence anchor (a citation or computation token) already marks that sentence: "this sentence is marked by C-…; read that citation with `--anchor`" — no `C-`, no label, no footnote |
| `refutes` | mints `claim_contradicts_source` on `--quote`, cited sentence or not |
| `absent` | records the reading; mints nothing (F5) |
| `unreachable`, or any outcome of a `not_the_work` copy | records the reading; reads nothing; mints nothing |

- **The marked-sentence guard** [NEW]: `uncitedSentence(of)`, a `seat` guard asked of the report the act
  lands on (as `Places` is), refusing a backing `--url` reading whose quote's sentence holds a citation or
  computation token (`claimcount`'s anchor reading, not a text search). Why refused, not recorded: the
  ruling mints `unsourced_claim` for an UNCITED sentence; on a marked one that class would be false, and
  a reading that mints nothing and places nothing tells no reader anything the cited source's own reading
  does not. Measured: `census.json` marks 2 of the 15 labelled archived readings (quadratic `c-95969800`,
  b9 `c-d49a32c4`) as on a sentence blue cited (its overlap test; not re-verified sentence by sentence).
  Name grep at `6102f2b09`: `uncitedSentence` 0 hits.
  <!-- CHANGED F13b --> **Decided, not a fork** (the targeted round): `ReadingGapClasses` is decided off the
  stored body, never the report, so a backing `--url` reading requires the gap flags; on a marked
  sentence the tool could record it only by dropping the mint the seat stated, and "the tool never makes
  it up" (refuse, never ignore a seat's stated act) leaves refusal, while the ruled option mints
  `unsourced_claim` for an UNCITED sentence only. Residue, stated: the 2 archived uses on a cited
  sentence (quadratic `c-95969800`, b9 `c-d49a32c4`) have no live act beyond a plain mint or a reading of
  blue's citation with `--anchor`. Which refusal a call without gap flags meets first is §III.2's
  ordering (the guard). <!-- CHANGED F13b -->
- **Materiality — O1 CLOSED (F13b; the targeted round: not a fork under the stated rulings).** <!-- CHANGED F13b -->
  Materiality is the class's (`material_default`; no seat picks it), and the option gblock picked names
  `unsourced_claim`, which is `CLASS_MATERIAL_BY_GRADE`. A backing found-source gap is therefore
  material by its grades, exactly as every `unsourced_claim` gap is. "Holds PASS until it closes" was the
  two sessions' gloss, not gblock's text; changing the class's default would touch every
  `unsourced_claim` gap, and a new always-material class would contradict the option he picked. The
  cost line is: the gap **holds the PASS where its grades make it material**; a `claim_contradicts_source`
  gap holds it always. Its consequence for #1657 1a′ is §III.16's.
- **Placement**: `Anchor{G, location: --quote}` through `claimRefusal` (§III.2a); a quote the report holds
  twice is refused (F9).
- **What blue does, and the four acts.** Blue's work list carries G; `show evidence`'s row whose
  `carried_by` is G names the source the lens found (§III.2c). Blue cites a source on the sentence (`blue
  cite`; with `--answers G` once #1657 1a′ adds it, §III.16); the citation is `unchecked`, so the PASS is
  held until a lens reads it with `verify --anchor`; the minting lens closes G. One act becomes four —
  mint (inside the reading), cite, verify, close — and the source enters the report only by blue's cite.
  Between this plan and 1a′ nothing ties blue's cite to G: no failed check carries a found-source gap (it
  reads no anchor), so the lens's close of G is its own judgement, as §III.16 states.
- **The key** <!-- CHANGED F13b --> (F13b, RULED; §III.1): a found-source reading is identified by its
  source and its sentence. One seat's found-source readings in one sitting that differ in url or in
  sentence are each recorded — one source backing several uncited sentences, several sources read
  against one sentence (b7's archived shape: events 152-155, four urls, one sentence; the first mints G,
  each later one names G with `carrier_minted` false); an identical retry (same url, same sentence, same
  sitting) is refused by `requireNotARetriedReading`, naming the stored reading. It takes the ordinal
  key; `keyFields` loses `url` and gains nothing.
- **The budget** <!-- CHANGED F13b --> (F13b, RULED; §III.5): any gap a `verify --url` reading mints —
  `unsourced_claim` or `claim_contradicts_source` — takes a mint of the lens's budget at the act and
  counts while open, as a plain mint does; at the cap the WHOLE reading act is refused for budget,
  nothing recorded (the reading and its Mint are one act). A `verify --anchor` reading's gap stays
  outside the budget (F6). A found-source reading that names a STANDING gap (`carrier_minted` false)
  mints nothing and takes nothing.
- **Scope** <!-- CHANGED F13b --> (F13b, RULED: F13 stays in #1694, not split). Phase 1, already built
  at `6102f2b09`, carries the corroboration arms F13 deletes (`ReadingGapClasses`' corroboration-`absent`
  arm, the repeat rule's url-and-claim arm, `redcitation_test.go`'s seven tests, the labelled paths the
  footnote reads); shipping #1647 first with `corroborate` would build those arms out through phases 2-5
  — its help, its fuzz drive, its seatprobe row, its migrate arm — only for F13 to delete them, which
  doubles the `corroborate` census (§III.12) for no shipped state anyone runs between the two.

**Deleted, and what replaces each** (no element survives for an archived record: §III.9 item 7 is
migrate's):

| Element (site at `6102f2b09`) | Replaced by |
|---|---|
| `lens corroborate`: `newCorroborate` (`verify.go:271-301`), `lens/command.go:29`, `help/corroborate.md`, its `ReasonIs` row, `requiredProbes` row, `seatprobe` boards row (`internal/seatprobe/boards.go:535`), `diagnostics` leaf list | `verify --url --title --quote`. The boards row is MERGED into `:534`'s `verify` expectation for `red-lens-evidence`, not rewritten beside it: one verb, one expectation, its `Because` naming both the cited claim read against its source and the standard blue never cited, read with `--url`, whose reading mints the gap blue answers with a citation <!-- CHANGED F13b --> |
| The two-verb rationale (`verify.go:84-91`) | a comment stating the flag groups |
| `citing`, `cites`, `adjudates`, `backsTheClaim`, the footnote decision (`verify.go:313-350, :426-485`) | `ReadingGapClasses`: a backing `--url` reading mints `unsourced_claim` |
| `Verify.label` (11), `Verify.work_status` (15), `Verify.sha256` (20) | nothing: reserved; the reading's url, title and copy fields name what it read, and a citation's stamp is blue's `Cite`'s |
| The red citation anchor: `record.NewID("citation")` in the reading, the `C-` Anchor `AppendPlaced` writes with it | the gap's Anchor `{G, location}`; only `blue cite` places a citation anchor |
| `ExistingCorroboration` (`citationid.go:439-473`) and the verb's retry branch (`verify.go:460-484`, `PlaceOwed` of a label) | `requireNotARetriedReading` within a sitting and the repeat rule across sittings (§III.1, §III.2) <!-- CHANGED F13b --> |
| The url-keyed duplicate of a backing `not_the_work` corroboration (§III.2's former crash-retry paragraph) | `requireNotARetriedReading`: same url AND same sentence in the sitting, whatever the reading mints <!-- CHANGED F13b --> |
| `ReadingClassField(corroboration bool)` | `ReadingClassField()` over both facets (§III.2) |
| `refuseCorroborationTitleVoice` (`redvoice.go`, its call `record.go:1785`) and the title's voice tells (`verify.go:440-453`) | nothing: a reading's title reaches no report text |
| `citedSource`'s Verify arm, `Corroborated`, labelled entries in `CitedSources` (`citationid.go:31-210`, `:432`) | blue's cites only |
| Red sources' notes and Bibliography lines (`report/citations.go:28`, `report/assemble.go` labelled arms, `layout.go:182-183` text) | blue's cites only |
| `show evidence`'s `CorroboratedBy`, the verification row's `Label`, the labelled source row (`evidenceview.go:123-127, :234, :381, :409-411, :450-473`) | the row's `carried_by` (§III.2c) |
| The ledger's hold of a label and a corroboration's check of its own footnote (`anchorchecks.go:229-232, :265-267`) | nothing: no reading holds an anchor of its own |
| migrate's owed Anchor for a labelled Verify (`gapplace.go:81`), remap's `Verify.label`/`Verify.sha256` rows and `mints` entry | §III.9 item 7 |
| `supersede.go:102-105` (the url- and label-keyed Verify rows) | ONE unkeyed `Verify` row for the found-source form (`url` and `label` key nothing now): "a new reading of that source for this sentence, at your next sitting — a reading of another source, or of this source for another sentence, is recorded now". The F13 row keyed `carried_by` sent the seat to "a reading of another source for this sentence", the act F13's key refused; it is not built <!-- CHANGED F13b --> |
| `seat/source.go:68` (`show source --anchor`'s "names a corroboration of" arm, reached by a label) and `:20`'s "a source of its own to corroborate one" | every anchor `show source --anchor` takes is a cite, which carries its `sha256`; a source a lens found is searched by `--sha`, from `fetch`'s summary |
| 1a′'s corroboration arm (V1a′.10(6), `compute-and-show.md:693, :758-759`) | §III.16 |

**Phase-1 code that must change** (built at `6102f2b09`; assigned to **phase 2, before migrate**, §III.14):
- `record/recordpb/record.proto` + `gapclass.go`: the `found_source_gap_classes` facet and
  `FoundSourceGapClassesOf`; `Verify.label`, `work_status`, `sha256` reserved; regenerate (protogen,
  schemagen; epoch stays 55, which no released record holds).
- `record/reading.go`: `ReadingGapClasses` takes the found-source arm (the corroboration-`absent` arm
  becomes the facet's absence); the comments naming `corroborate` (`:19`, `:36-38`, `:71`, `:98`, `:177`).
- `record/failedcheck.go` `requireNotARepeatReading`/`repeatReadingRefusal` (`:315-375`): the unanchored
  arm keys on `ReadingGapClasses` non-empty, not `FindsAgainst`; its refusal drops "a point the gap lacks
  is a finding about %s" <!-- FINDING-DEPENDENT -->; the comment `:310`. <!-- CHANGED F13b --> Also the early
  return `if Migrating || !FindsAgainst(v)` (`:319`), which otherwise leaves the unanchored arm dead for
  every backing reading, and the refusal's text worded by form (§III.2's repeat bullet).
- <!-- CHANGED F13b --> `record/record.go`: `keyFields` LOSES `url` (and gains nothing; F13's `carried_by`
  is not built), and the Verify arm gains `requireNotARetriedReading` first (§III.1).
- <!-- CHANGED F13b --> `record/reading.go`'s two refusal strings, false of a backing found-source
  reading: `AppendReading`'s "this reading does not find against its sentence (outcome %s, copy %s), so
  it mints no gap: it is written alone" (`:84`) becomes "this reading mints no gap (outcome %s, copy %s,
  <form>): it is written alone", and `requireReadingCarrier`'s "a reading that finds its source against
  the sentence (%s) names the gap that carries it" (`:118`) becomes "a reading that mints a gap (%s) names
  the gap that carries it".
- <!-- CHANGED F13b --> `record/mintbudget.go`: the act-paired admission (`:308`, `w.act.pairs(m)`) and the
  open-count exclusion (`:360-361`, the `carrier_minted` query) each narrow to a Verify with an `anchor`
  (§III.5).
- <!-- CHANGED F13b --> The SQL readers of the reserved `verify.label` column: `record/id.go:88`
  (`unplacedLocationAt`), `record/refs.go:130` (`requireAboutAnchor`), `record/refs.go:205`
  (`HoldsAnchor`) — each's `verify` UNION arm DELETED (§III.2a's table); `consistency/consistency.go:272`
  and `record/migrate/proserefs_test.go:112-113` (each a `Verify.GetLabel()`), DELETED.
- `record/foundagainst_test.go` `TestEveryNegativeOutcomeNamesItsClasses` gains the found-source facet;
  `record/redcitation_test.go` — phase 1's three rewritten tests (`TestACorroborationThatFindsAgainstHoldsThePassThroughItsGap`,
  `TestASecondReadingOfAHeldSentenceNamesItsGap`, `TestARefutingCorroborationCarriesNoFootnote`) and the
  four it kept (`TestOneSourceCorroboratesManyClaims`, `TestASupportingCorroborationJoinsTheBibliography`,
  `TestRedsCitationAnchorsAreProtectedLikeBlues`, `TestTheCorroborationGuardsHoldEachHalfSeparately`) —
  the file is replaced by `record/foundsource_test.go` (§III.15 tests 9a–9d).
- The record half of the deletions table above (`citationid.go`, `evidenceview.go`, `anchorchecks.go`,
  `redvoice.go`, `supersede.go`, `refs.go`'s refusal text) — record-package code phase 2's loop builds.

**Tests** (NEW, §III.15): `TestAFoundSourceReadingMintsItsGap` (each row of the table above through the
verb: class, Anchor `{G, location}`, no label, no citation Anchor, `carried_by`); `TestOneSourceBacksManySentences`
(one seat, one url, three uncited sentences in one sitting → three Verifys, three gaps; <!-- CHANGED F13b -->
an `absent` reading of that url on two sentences in the sitting → both recorded; the same `absent`
reading of one sentence twice → the second refused, naming the first's event);
<!-- CHANGED F13b --> `TestOneSentenceReadAgainstManySources` (b7's archived shape, events 152-155, driven
through `verify --url`: one seat, one sitting, four urls, one uncited sentence, `--as supports` → four
Verifys; the first mints G (`unsourced_claim`, `carrier_minted` true), the other three name G
(`carrier_minted` false), no second gap; a fifth call repeating the second url and sentence → refused by
`requireNotARetriedReading`, the refusal naming the second reading's `events.id`, nothing recorded);
`TestABackingFoundSourceReadingOfAMarkedSentenceIsRefused` (<!-- CHANGED F13b --> with a row sent WITHOUT
the gap flags, refused with the guard's sentence, not "REQUIRED --class"); `TestNoReadingPlacesACitation`
(record layer: `CitedSources`, the Bibliography and `show evidence`'s sources hold blue's cites only
after every form and outcome); `TestVerifyTakesExactlyOneSubject` (each flag group violated → cobra's
refusal through `RefuseFlagGroup`, nothing recorded; `lens corroborate` → unknown command);
<!-- CHANGED F13b --> `TestAFoundSourceGapTakesAMint` (both arms of the budget ruling, through the verbs: a
lens AT its cap — (a) a `verify --url --as supports` of an uncited sentence and (b) a `verify --url
--as refutes` are each refused for budget, the refusal `requireMintWithinBudget`'s, and the record holds
no Verify, Mint or Anchor from either; (c) a `verify --anchor --as refutes` of a citation at the same cap
mints its gap; (d) one cap below, a `verify --url` mint is admitted and the lens's open count rises by
one, while (c)'s gap leaves it unchanged; (e) at the cap, a `verify --url` reading of a sentence a
standing gap of its class already carries is recorded naming that gap, no mint taken). Name grep at
`6102f2b09`: all seven 0 hits (`TestOneSentenceReadAgainstManySources`, `TestAFoundSourceGapTakesAMint`
included). The record half's tests are §III.15 9a–9e, the migrate tests 13a–13b, the CLI tests 35a–35f.

### III.3 The anchor ledger and the failed check — `internal/record/anchorchecks.go`, `failedcheck.go` [MODIFY] (F4)

- `stateFoundAgainst` stays as a ledger state (it is the citation's standing, and `CitationsFoundAgainst`,
  `counts.sources_refuted` and capture's assembled-report check read it); its comment stops saying "the
  contradiction's".
- `failedChecksOf` adds the found-against state: each standing negative reading of the sentence the report
  carries is a failed check whose owner is its reader and whose carrier is read from the reading's
  `carried_by` field, never from an about-anchor mint (the about-anchor scan at `failedcheck.go:60-64,
  :75-83` stays for `unread`/`unsound` only). One failed check per standing negative reading, so one
  anchor read against by two lenses holds two.
- **The carrier follows its successor.** A lens may close the carrying gap with `--superseded-by`
  (`requireFailedCheckAnswered` admits a close that names a successor, `failedcheck.go:180`). Read from
  `carried_by` alone, the check would then name the closed predecessor, and the successor — tied to no
  failed check — could close while the reading stands. So `answers` is `carried_by` and every gap on its
  successor TREE: `successorsOf` (`successor.go:29-47`) lists, per gap, both its close's `successor`
  and every gap whose mint names it in `supersedes`, so one gap can have several successors and the
  lineage branches. `answers` is the transitive closure from `carried_by` over `successorsOf`
  (cycle-safe: a gap is visited once). `answeredBy` is ONE gap, by a stated tie-break: the OPEN gap of
  `answers` minted last (greatest `minted_seq`); where none is open, the gap of `answers` minted last.
  OPEN IS THE GAP VIEW'S, never a Close scan: a closing docket ruling closes a gap through
  `closure_act`'s bench arm and writes no Close (below), so `failedChecksOf` (events only today,
  `failedcheck.go:56`) takes the open set its caller reads in the same snapshot from the `gap` view
  (`SELECT "gap_id" FROM "gap" WHERE "open"`, the column `gapState`, `refs.go:298-305`, already reads) —
  `failedChecksOf(evs, win, open)`, at each of its five callers (`failedcheck.go:113, :142, :186, :210`,
  `passblockers.go:713`) and its five test calls (`record/reopened_test.go:205, :210, :251, :262`, and
  `:263` inside the `Errorf`,
  which read the open set through the same `gap`-view query; the file moves from KEEP to REWRITE,
  §III.12). `grep -rn 'failedChecksOf(' tools/internal` finds no other caller.
  The close gate refuses the close of EVERY gap in `answers` while the check stands, a close naming
  its successor (`--superseded-by`) admitted as today (so no branch closes out from under it); the repeat rule refuses while ANY gap in `answers` is open and names
  `answeredBy`; every printed sentence names `answeredBy`. No field is added: the tree is the record's
  own `Close.successor` and `Mint.supersedes`. Tested by
  `TestASupersededReadingGapHandsTheCheckToItsSuccessor` (§III.10), which covers a linear chain and a
  branch.
- **What clears a found-against check is the ledger's rule** (`anchorStanding.stands`,
  `anchorchecks.go:84-94`): only the SAME seat's later reading that reads replaces its check. So a
  found-against check clears when its reader's own later reading of the sentence as it stands backs it,
  when blue edits the sentence (which leaves it unchecked), or when blue retires the citation. Another
  lens's backing reading beside it is a disagreement and clears nothing — as `failedClears` already says
  of a computation's `sound` beside a standing `unsound`. Every found-against sentence below that states a
  clear condition states this one; none says "a lens's reading" or "a lens reads its source" (the
  `unread` state's condition, which any reading of the work meets).
- **A closing docket ruling closes the carrying gap; the check alone keeps holding the PASS** (gblock,
  2026-10-30: "Gap closes, PASS still held"; the same as today's unread and unsound checks,
  `TestABenchRulingDoesNotClearAnUnreadCitation`). The bench's ruling closes a gap through
  `closure_act`'s bench arm (`recordsql/views.go:459-466`, read by the gap view at `:547-557`) and writes
  no Close, so `requireFailedCheckAnswered` (`record.go:1108`, Close only) does not run, and nothing is
  added to make it run. The failed check is not the gap: it stands until one of the three acts above,
  and no bench ruling clears it. No sentence says the gap does not close under a ruling.
- **OWED-NESS AND WORDING KEY ON THE STATE, NEVER ON `answeredBy == ""`.** `failedCheck` and `Blocker`
  gain the check's state (`unread | unsound | found_against`; `Blocker.CheckState()` exported for the CLI
  and the fuzz), and `Detail != ""` stops standing for "computation, else unread citation" (`failedCheck.computation`
  and the blocker's `Detail`, the computation's `bundle_sha`, since #1644). The rule every
  reader below holds: a lens **owes** a gap for a failed check of its own only when the state is `unread`
  or `unsound` and no gap carries it; a `found_against` check is never owed by its reader, carried or not
  — its gap is minted at the act, and where none is named (a migrated archived reading, §III.9) no mint
  answers it either. `answeredBy` says only *which gap carries it*, for printing and for the close gate.
  The readers, each changed to that rule (census: §III.12, "Readers of a failed check's owed-ness"):
  - dispatch's lens arm (`dispatch.go:546-553`): readies the owner for `unread`/`unsound` with no carrier;
    never for `found_against`, which readies blue alone (`:555-561`, cause `max(since, answeredAt)`);
  - the lens's work item (`sitting.go:401-408`): `add` (owed) for `unread`/`unsound` with no carrier;
    `may` for every `found_against` check — so `EnvelopeRefusal` (`envelope.go:44`, which reads this
    list) never refuses a reader's envelope over one;
  - retirement (`retirement.go:164-172`, `lensOwesOf`): a failed check is owed only for
    `unread`/`unsound` with no carrier; `found_against` is skipped, so it never holds its reader from
    retiring. Its doc comment (`:142-144`) says "a check of its that read nothing or judged a computation
    unsound and that no gap of its carries yet";
  - `answersOwnFailedCheck` (`failedcheck.go:109-119`) and `ownFailedCheckAnswers` (`:208-215`): the
    budget exemption of an about-anchor mint is for `unread`/`unsound` only. A reading's gap is outside
    the budget by F6's rule (§III.5), never through the failed-check carrier — so a held gap that another
    lens minted cannot enter this lens's exemption set, and no mint is exempted twice;
  - `requireNotARepeatFailedCheck` (`failedcheck.go:134-169`): `found_against` reads its own sentence
    (§III.2's repeat rule); the citation branch's "read nothing" wording is the `unread` state's only;
  - `requireFailedCheckAnswered` (`failedcheck.go:178-204`): a third message, for `found_against`:
    "gap G carries citation C, which L's reading at event N finds the source against, and the report still
    carries that sentence — the defect has not left the report, so this close is refused. Nothing is
    recorded. Your close is admitted once L's own later reading of the sentence as it stands backs it,
    blue edits the sentence, or blue retires the citation; another lens's backing reading is a
    disagreement and clears nothing; a close that hands the defect on names its successor with
    `--superseded-by`".
- **Every printed sentence of a failed check keys on the state** (`passblockers.go`):
  - `failedFact` (`:226-231`): `found_against` → "L's reading of citation C at event N finds the source
    against the sentence (`refutes` or `absent`), and the report still carries that sentence";
  - `failedClears` (`:234-240`): `found_against` → "it clears when L's own later reading of the sentence
    as it stands backs it, or when blue edits the sentence, which leaves it unchecked, or retires the
    citation; another lens's backing reading beside it is a disagreement and clears nothing. The lens's
    close of the gap that carries it is refused while it stands; a closing ruling of the bench closes that
    gap, and the check still holds the PASS — no bench ruling clears it";
  - `failedRemedy` (`:245-250`): `found_against` → "edits the sentence, or retires or replaces the citation";
  - `failedOwed` (`:254-259`): the "owes its gap about the anchor (`mint --about-kind anchor …`), which
    takes no mint of its budget" branch is `unread`/`unsound` with no carrier only; `found_against` with a
    carrier → "gap G carries it to blue, which …"; without one → "no gap names it — blue …" (no seat is
    told it owes a mint);
  - `blueFailedItem` (`:262-274`): "the lens whose check failed has minted no gap about it yet" is the
    `unread`/`unsound` no-carrier sentence only; `found_against` without a carrier → "no gap names it";
    the citation remedy sentence takes `failedRemedy`'s found-against wording, and "Retire the citation" is
    the `unread` sentence's;
  - the kind's `gloss` (`:404`), `refusal` (`:408-426`: its hard-coded "The lens whose check failed mints
    its gap about the anchor …" names the `unread`/`unsound` duty only, and `how` gains "found against"
    (participle, as "judged unsound")) and `why` (`:427-433`, `until`);
  - `failedCheckReadiness` (`dispatch.go:1293-1308`): `what, failed` and blue's remedy by state; the lens
    arm's "minted no gap about it" stays, as only `unread`/`unsound` reach it.
  - the verb's failed-check sentence (`verify.go:530-542`, `:632-657`): by state, after every
    verification of a citation, §III.2 — this site was judged state-safe in round 2 and is not: its
    guard is the new reading's outcome, not the check's state.
- **The lens's turn (#1659) reads a reading's gap as it reads any mint.** `lensTurnOn` and capture's
  `LensTurnAudit` take the Mint's place from `mintPlaces`, so blue is readied first on a reading's open
  material gap, and its minting lens once blue has acted on it. A Verify is no act on a gap
  (`gapActsOf`, `gapacts.go`, has no Verify arm), so a second lens's reading naming G, or the reader's
  own re-read, moves no turn. Both stay unchanged, and the check's readiness above is by state. The
  reader's backing re-read happens in any sitting a plan readies it for.
- Read and judged state-safe: `verdict.go:141-162` (`anchorsStanding`) and `debate.js:1087-1094` name
  the anchor only, whatever its state.
- The close gate: `requireFailedCheckAnswered` refuses the minting lens's (or any seat's) `close` of a
  gap in a found-against check's `answers` — successor (`--superseded-by`) admitted — until the check
  clears by the ledger's rule above. A closing docket ruling closes the gap and the check still holds
  the PASS (above). The gap's close stays the minting lens's ordinary `close`, so no new answer kind
  exists.
- **The mechanism's own sweep** — every site that SELECTS a failed check (by Subject, Owner, AnsweredBy
  or Detail) and decides or prints, re-run at `9ee7ad95e` (`~/.claude/scratch/plan1647-r10/new-fc.txt`) with
  `grep -rn -E 'BlockerFailedCheck|AnsweredBy\(\)|failedChecksOf\(|\.answeredBy|\.answers\b|\.Detail\b'`
  over `tools/internal`, `tools/cmd`, `tools/releasegate` and `debate.js`, every hit read (dispositions in
  §III.12's table "Readers of a failed check's owed-ness"): the record readers above; `verify.go:539`
  (§III.2); `sitting.go:321` (blue's arm, prints `blueFailedItem`, by state); `passblockers.go:714` (the
  producer, which now sets the state); the fuzz sites of §III.11 (`fuzz_test.go:1229, :1238, :3026, :3036,
  :3058, :3061, :3115-3124, :3434, :3735, :4208, :4253, :4285`, and the gates at `sweeptally_test.go:843-854`);
  KEEP — `verdict.go:148`, `debate.js:1093`, `fuzz_test.go:2972` (asserts the owner or blue is readied,
  true of every state), `fuzz_test.go:6990` (no VERIFIED over any failed check), `migrate/plankinds.go:19`; another kind's `.Detail`
  (#1660's `upheld_position`: `dispatch.go`'s lens arm, `passblockers.go`'s kind row, `positionheard.go`).
  The contradiction fold's hits (`citationid.go:554-685`, `onegapperclass.go:145`, `mintbudget.go:362-380`)
  are deleted. `PlanBlocker` (`dispatch.go:29-33`) carries kind, subject and owner only, and stays so:
  the state reaches a seat in the plan's `why` sentence, not as a field. The new mechanism's own readers
  — `CheckState()`, `failed_checks[].state`, the successor-following `answers` — are those same sites and
  no others: `failed_checks` has no reader in Go outside the verb (no golden or test reads
  `failed_check` today; its consumer is the seat), which §V.4's sweep re-runs.
- A carrier-less found-against check — which only a migrated archived reading produces (§III.9) — readies
  blue alone, sits on its reader's list as `may`, holds no reader from retiring, admits no budget-exempt
  mint, and clears by the ledger's rule: its reader's own later backing reading of the sentence, an edit
  of it, or a retire of the citation.
  Tests: `TestAMigratedFoundAgainstCheckReadiesBlueAlone` (§III.9) and
  `TestAFoundAgainstCheckIsWordedByItsState` (§III.10).
- `holds()` comment (`anchorchecks.go:104-107`) and `VerifyReads` comment (`:278-281`) [MODIFY].
- A reading of a source the lens found places no citation anchor (F13), so it reaches the board through its gap only, and no failed check carries that gap. <!-- CHANGED F13 -->

### III.4 The one-gap rule — `onegapperclass.go` [MODIFY]

- `answersRaisedContradictionAt` and its call (`:68-72`) [DELETE]; the exemption bullet (`:40-42`)
  [DELETE]. The rule now has no exemption but `--supersedes` and replay.
- [NEW] the held-gap return used by the reading act (III.2).

### III.5 Mint budget — `mintbudget.go` [MODIFY] (F6)

- `answersOwnRaiseAt`, its call (`:301-305`), `ownAnswersAt`'s contradiction half (`:369-397`), the
  `Answering` wording at `:506`, `:562` [DELETE/REWRITE]; the comment block `:33-41` [REWRITE].
- <!-- CHANGED F13b --> **F6 amended (gblock, 2026-11-03): only a reading of a CITATION's gap is outside the
  budget.** Any gap a `verify --url` reading mints (`unsourced_claim` or `claim_contradicts_source`)
  takes a mint at the act and counts while open, exactly as a plain mint; gaps from `verify --anchor`
  readings, which check blue's citations, stay outside it as F6 states. Unruled, F6 would have exempted a
  lens's most common class whenever it attached a backing source. Each of the two readers below
  therefore keys on the paired Verify's `anchor`, a field the record already holds: the act-paired
  admission (`requireMintWithinBudget`, `mintbudget.go:308`, `w.act.pairs(m)`) admits a paired Mint
  without counting only when the paired Verify has an `anchor`, and counts it otherwise; the open-count
  exclusion (`:360-361`) adds `AND COALESCE(v."anchor", '') <> ''`. At the cap a `verify --url` reading
  whose `ReadingGapClasses` is non-empty and that names no standing gap is refused whole for budget —
  nothing recorded; a reading that NAMES a standing gap (the held-gap path) writes no Mint and meets no
  budget. `mintAnswerOutsideBudget` (`:350`) and `lensMintStanding`'s "your readings minted them" (`:554`)
  say "your readings of citations". Tests `TestAFoundSourceGapTakesAMint` (§III.2d; both arms) and
  `TestAReadingsGapTakesNoMint` (a `verify --anchor` row only). Partner §V.5.ww.
  **On replay.** No archived reading comes forward minting a gap: §III.9 items 2 and 7 write no
  `carried_by` on an archived reading, so the found-source arm meets no archived act. A record written at
  this epoch replays its found-source act whole (item 6), under the same budget, admitted live within it;
  `TestARecordWrittenAtThisEpochMigratesWithNoRefusal` gains a `verify --url` minting row by a lens below
  its cap. Measured against `run-archive/`, §III.9 item 7 ("The budget on replay").
- [NEW] a gap the reading act minted from a CITATION reading takes no mint and does not count while
  open. <!-- CHANGED F13b --> Two readers, two sources, neither inferred from "some Verify names this gap":
  - **at the write** (`requireMintWithinBudget`, `mintbudget.go`): the Mint's validation runs before
    the Verify exists in the transaction, so the exemption is read from the ACT — the reading's placed append
    passes its paired Verify body to the Mint's validation (§III.2), and a Mint so paired is admitted
    without counting it;
  - **while open** (`lensMintStanding`): a gap is excluded from the seat's count when a Verify by that
    seat names it with `carrier_minted` true — the Mint written in the reading act. A gap the lens
    minted with plain `mint` and that a later reading names as the open gap refusing the mint
    (`carrier_minted` false) keeps its mint. So does a SUCCESSOR of a reading's gap: a gap the lens
    mints with plain `mint --supersedes G` (or names in `close --superseded-by`) carries the check on
    (§III.3) but was minted by no reading act, so it counts against the budget like every plain mint —
    only a gap named by its own reading with `carrier_minted` true is outside it. Tested by
    `TestAReadingsGapTakesNoMint` and `TestAPlainMintALaterReadingNamesKeepsItsMint` (§III.10), whose
    successor row mints `--supersedes` the reading's gap at the cap and is refused for budget.
  - **on replay** (`Migrating`): the budget stays on under replay (`record.go` Mint arm, "the roster and
    the mint budget stay on"), and replay appends one event at a time, so the act's pairing the write
    reads would be gone and a capped lens's reading Mint refused. migrate re-drives the reading's act
    whole (§III.9 item 6), so the Mint's validation receives the same paired bodies it received live.
- `mintAnswerOutsideBudget` (`:350`) [REWRITE]: "A gap your reading of a citation mints, or about <!-- CHANGED F13b --> an anchor whose check
  of yours read nothing or judged a computation unsound, takes no mint; …" (today's text names "a
  finding of yours that raised a contradiction" and "a computation whose sentence you judged unsound").

### III.6 PASS blockers, work list, readiness, retirement [MODIFY/DELETE]

- `passblockers.go`: `BlockerContradiction` (`:35`), `contradictionReader`, `contradictionCitationAnswer`,
  `contradictionAnswerBudget`, `contradictionGapSubject`, `contradictionFact` (`:139-156`), the kind's
  table row (`:340-367`), the fold call (`:707-708`), `raisedBy`/`raiser` fields (`:81`) [DELETE].
  `failed_check`'s row (`:403`) gains the found-against sentence.
- `citationid.go:478-708`: `contradictionsOf`, `ReadingContradiction`, `ContradictionOfReading`,
  `requireOpenContradiction`, `unansweredContradictions`, `contradiction`, `contradictionActs`, `idWord`
  [DELETE]; `contradicts` → replaced by the `gap_classes` facet reader (III.1). Comment at `:74` [MODIFY].
- `about.go`: the gap's contradiction refusal (`:57-59`), the contradiction case (`:99-106`) and the doc
  sentence (`:23-25`) [DELETE]; `aboutShapes`' comment (`:15-16`, #1674: "a contradiction's id is a number,
  which requireOpenContradiction reads") [REWRITE] to its first sentence. `aboutShapes` itself holds no
  contradiction row and is unchanged.
- `record.go`: the finding case (`:1062-1066`) and the mint's about-kind refusal (`:966-971`) [DELETE] —
  with every kind admitted the check has nothing to refuse; `enums.go:294-300` (finding's set "less
  contradiction") [REWRITE]: finding and mint carry one set. Kept under #1695's relay: it deletes the about-kind, and makes no finding. <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->
- `sitting.go:374-439` (the lens owes a contradiction it read; the at-cap special case) [DELETE]; the
  failed-check arm (`:401-408`) keys on the state (§III.3).
- `dispatch.go`: the readiness case (`:535-541`), `contradictionReadiness` (`:1281-1288`), Plan JSON
  `desc` text (`:19-31`) [DELETE/REWRITE].
- `retirement.go:20, :142` (a lens is not retired while it owes a contradiction) [REWRITE]; the
  failed-check skip (`:164-172`) keys on the state, so a found-against check never holds its reader
  from retiring (§III.3).
- `evidenceview.go:306-373`: `UnansweredContradictions`, `EvidenceContradictionJSON` [DELETE];
  `:578`, `:598` comments [REWRITE]. `narrowedview.go:20` [REWRITE].
- `convergence_refusal.go:77` [REWRITE].

### III.7 Agent-facing text [MODIFY / REGENERATE] (contract: prompts and constitutions)

- Help pages (`internal/cli/seat/help/`): `corroborate.md` is DELETED with the verb (F13), and `verify.md`
  teaches both forms — `--anchor` for a citation blue made, `--url` with `--title` and `--quote` for a
  source you found, exactly one of the two, and §III.2d's table of what a `--url` reading mints (an
  uncited sentence your source backs → `unsourced_claim`; a sentence it refutes →
  `claim_contradicts_source`; a sentence a citation already marks → read that citation with `--anchor`;
  your source places no citation: blue cites, and a lens reads blue's citation) <!-- CHANGED F13 -->; <!-- CHANGED F13b -->
  that a `--url` reading's gap takes a mint of your budget, as a plain mint does, while a reading of a
  citation's gap takes none; and that in one sitting you may read several sources against one sentence
  and one source against several sentences, each recorded, while repeating a reading of the same source
  for the same sentence is refused, naming the one you recorded. `verify.md`'s
  two contradiction paragraphs are replaced by one paragraph: a reading of the work that finds the source against the sentence
  mints its gap in the same act, the flags it needs, and that a reading that reads nothing —
  `unreachable`, or any outcome on a copy the tool stamped not_the_work, a backing one included — mints
  nothing, its gap flags are refused, and the citation stands
  unread (`verify.md:11`'s unread paragraph keeps that duty, now true of `verify`, which stamps the
  copy's completeness); where the gap stands — `verify`: "the gap stands at the end of the sentence your
  --anchor's token sits in, found by that token, so a sentence whose words the report holds twice takes
  its gap where you cited it, and a --new you give is checked, and accepted by blue, at that
  occurrence";
  `--url`: "the gap stands at your --quote, which the report must hold once; a quote it holds
  more than once is refused"; what the one-gap rule does to a second reading, and what clears the citation's failed check (your own
  later reading of the sentence as it stands that backs it, blue's edit, or blue's retire; another lens's
  backing reading is a disagreement and clears nothing); the gap's flags and its argument: "the --reason
  you give says what the source says; where your reading mints its gap it is also that gap's argument,
  which the board shows to blue and the bench, and where an open gap of its class already carries the
  sentence it stays on your reading alone"; and estoppel's exemption (F10), in the words
  `help/mint.md:19` gains (§III.2): "A gap your reading of a source mints in its own act is not
  refused: a source that finds against the sentence is new evidence, not your prescription
  relitigated";
  `corroborate.md:17`'s "How you found the source goes in your reason, which the report never prints"
  goes with the page; `verify.md` says how you found a source goes in your reason, which is the gap's
  argument when the reading mints <!-- CHANGED F13 -->. `mint.md:23` keeps its text: `mint`'s `ReasonIs` row is byte-unchanged.
  `finding.md`, `mint.md`, `dispatch.md` [REWRITE]: their contradiction sentences are deleted, and no
  sentence gains a finding duty (the finding verb is #1695's) <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->; `mint.md:19` and
  `edit.md:13` [REWRITE] for estoppel's exemption (§III.2); `finding.go:145-146` flag
  descriptions [REWRITE]: the contradiction kind's text deleted <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->; `verbs.go`, `flags/names.go` per census. The failed-check paragraphs name the
  third state: `mint.md:15` ("A CHECK OF YOURS THAT FAILED … you owe its gap") says the owed gap is an
  unread or unsound check's, and that a reading that finds the source against its sentence mints its gap in the same act;
  `close.md:11` [REWRITE] opens "YOUR CLOSE OF A GAP THAT CARRIES A FAILED CHECK IS REFUSED WHILE THE
  CHECK STANDS", adds the found-against case, states what clears each state by the ledger's rule (a
  found-against check: your own later reading of the sentence as it stands that backs it, blue's edit of
  the sentence, or blue's retire of the citation — another lens's backing reading clears nothing), and
  says a closing ruling of the bench closes the gap while the check still holds the PASS.
- **No page says a gap carrying a failed check does not close** (gblock's ruling, F8). Sweep, run at
  `9ee7ad95e` (123 hits; list at `~/.claude/scratch/plan1647-r10/new-dnc.txt`) over `scripts/agentgen/src`, `tools/internal`, `tools/releasegate`, `skills`, `docs`:
  `grep -rIn -i -E "(does not|cannot|can't|never) close"` plus the multi-line form
  `perl -0777 -ne 'while(/(gap|it)[^.]{0,60}?(does not|cannot|can.t|never) close[^.]{0,80}?(while|until|stand)/gis){…}'`,
  every hit read. Hits in this sense, each [REWRITE] to "the close is refused while the check stands; a
  closing ruling of the bench closes the gap, and the check still holds the PASS" in the page's own
  voice: `help/close.md:11`, `help/mint.md:15` ("it does not close while the check stands"),
  `help/verify.md:11` ("That gap does not close while the citation stands unread" — the rest of that
  unread paragraph is unchanged), `help/reproduce.md:21` ("That gap does not close while the judgement
  stands"), `terms.json:1080`, `fragments/lens-audit-and-mint.md:6` ("The gap does not close while the
  check it carries still stands failed"), `passblockers.go:238` (`failedClears`, unread) and `:423` (the
  refusal: "the gap does not close while the check stands, and no bench ruling clears it" → "a seat's
  close of the gap is refused while the check stands; a closing ruling of the bench closes that gap, and
  no bench ruling clears it" — the last clause is the substring `TestABenchRulingDoesNotClearAnUnreadCitation`
  asserts), `failedcheck.go:193, :198` (the close refusals' "so the gap does not close" → "so this close
  is refused"), `docs/citations-and-bibliography.md:162-166`; code and test comments in the same sense:
  `failedcheck.go:30`, `unreadcitation_test.go:119`, `unreadb6_test.go:21-22, :66`, `fuzz_test.go:3118,
  :4205`, `sweeptally_test.go:849`; REGENERATE `docs/vocabulary.md:629` and the help golden. Every other
  hit is another sense (a sitting the record cannot close, a quote or fence that does not close, the
  `remanded` disposition, `record/viewjson.go:1005`, `positionowed.go`, `remand_test.go`,
  `positionheard_test.go`) and KEEPs.
- `terms.json`: the `contradiction` entry [DELETE] (with its ordinary-English collision); `failed check`
  [REWRITE] to ONE sentence — `checkDefinition` (`terms/terms.go:150-163`) refuses a definition holding
  ". " ("the definition is ONE sentence"), gated by `TestTheEmbeddedRegistryLoads` — "A failed check is
  a standing check of a citation or computation that failed — a lens judged the computation's sentence
  unsound, every reading of the citation since its sentence last changed read nothing of the cited work,
  or a lens's standing reading found the cited source against the sentence and, in the same act, minted
  the gap that carries it or named the open gap that already does — which holds the PASS and is cleared
  by no bench ruling; an unread or unsound check obliges the seat whose check failed to mint a gap about
  the anchor, outside its mint budget; a seat's close of the gap that carries a failed check is refused
  while the check stands, and a closing ruling of the bench closes that gap while the check still holds
  the PASS; and the check clears when an edit changes its sentence, when blue retires the anchor, or when
  the anchor is checked and holds — for a computation or a found-against citation by the checking lens's
  own later check, since another lens's check beside it is a disagreement, and for an unread citation by
  any lens's reading of the work."
- `terms.json`: [NEW] entry `reading` (GAP 1 of round 8; the plan's prose leans on the word in a sense
  the registry did not hold), `seats` the same six as `failed check` (`lens`, `chair`, `blue`, `lane`,
  `bench`, `operator`), `bans: []` — no GATED ban on `reading`, and none on `contradiction` (no old-word
  lists, gblock 2026-10-13). Definition, ONE sentence: "A reading is a lens's recorded account of what
  one source did for one sentence of the report — `verify` of a citation, or of a source the lens
  found — with its outcome, its confidence and its reason; it reads the work unless the source
  was unreachable or the copy the tool stamped is not the work, and one that reads the work and finds
  the source against the sentence mints the gap that carries it in the same act or names the open gap of
  that class the sentence already holds, and one of a source the lens found that backs an uncited
  sentence mints the gap that asks blue to cite it, except an `absent` on a source the lens found, which
  mints nothing." <!-- CHANGED F13 -->
  `collisions`, each `{word, resolution}`:
  - "reading (OCR)": "the OCR reading is the machine's reading of a scanned document's page images
    (`reading_render_sha`, `render-page`'s `reading:` line, OCR-derived text); a lens's reading is a
    recorded check of a source against a sentence, and a page's OCR reading may be what a lens's
    reading finds wrong";
  - "verdict": "the verdict is the chair's PASS or FAIL; a reading has an outcome, and a reading that
    reads nothing is never called a verdict";
  - "report read": "a report read is the tool serving a lens the whole report; a reading is a lens's
    check of one source against one sentence".
  Name grep at `9ee7ad95e`: no `terms.json` entry is named `reading` (its hits are the `failed check`
  and OCR-derived-text definitions and the deleted `contradiction` entry). What applies: `checkDefinition`
  (one sentence, a period), `TestNoDefinitionUsesABannedVariant`, the manual golden (`show terms`),
  `agentgen -check` (each seat's inlined manual), `vocabdoc -check` (`docs/vocabulary.md`).
  **Sweep — every seat-facing use of "reading"**, run at `9ee7ad95e`:
  `grep -n -i -w -E 'readings?' cli/seat/help/*.md scripts/agentgen/src/{*.md,fragments/*.md,skills/*/SKILL.md} report/layout.go terms/terms.json`
  (41 lines in 18 files), every hit read:
  - the term's sense, consistent with the definition once §III.7's rewrites land — `help/verify.md:11`,
    `:17`, `help/corroborate.md:13`, `:15`, `:19` (DELETED with the page, F13), `red-lens-evidence.md:15`,
    `fragments/lens-audit-and-mint.md:6`, `:8`, `fragments/lens-duties.md:5`,
    `fragments/blue-search-and-cite.md:23`, `skills/research-protocol/SKILL.md:17`, `:79`,
    `report/layout.go:182-183`, `terms.json:1080` (REWRITE where §III.7 already lists them; else KEEP),
    and `terms.json:763`, `:772` (DELETE with the `contradiction` entry);
  - the OCR sense, resolved by the first collision — `help/verify.md:21` ("the reading may be what is
    wrong"), `help/render-page.md:7`, `:9`, `:11`, `red-lens-evidence.md:13`, `help/cite.md:13` (×2),
    `fragments/leaf-node-read.md:6`, `terms.json:580` — KEEP;
  - ordinary English (the gerund) — `help/cite.md:11`, `:19`, `help/compute.md:17`, `help/mint.md:11`,
    `help/reproduce.md:17`, `fragments/surface-discovery.md:1`, `fragments/lens-audit-and-mint.md:15`,
    `skills/adversarial-audit/SKILL.md:25`, `:27`, `skills/research-protocol/SKILL.md:36`,
    `lead-judge.md:41`, `:56`, `report/layout.go:289, :335, :358, :369` — KEEP.
  New prose this plan specifies says "a reading that reads nothing", never "a verdict that reads
  nothing" (§III.7). `terms.json` `lens retirement` [REWRITE] ("no contradiction it read" goes). Then `go -C scripts run ./vocabdoc` regenerates `docs/vocabulary.md`.
- agentgen sources [REWRITE]: `scripts/agentgen/src/fragments/lens-duties.md`, `fragments/lens-audit-and-mint.md`,
  `src/red-chair.md` (census), `src/skills/research-protocol/SKILL.md:17-19` ("it goes to the board as a
  finding — a PASS is refused until one is raised for it", found by the phrase sweep below), and
  `src/lead-judge.md` gains one sentence for the bench (F8, R4): "A closing ruling on a gap that carries a
  failed check — a citation no reading reads, a computation judged unsound, a citation whose source a lens's
  reading finds against it — closes the gap and leaves the check standing: the PASS stays held until a
  check that clears it holds, blue edits the sentence, or blue retires the anchor. Your ruling decides
  the gap, never the record's state of the anchor." `fragments/blue-search-and-cite.md:22-29` (blue's failed-check duty) [REWRITE]: a citation found against
  is fixed by an edit of its sentence, or by retiring or replacing the citation, and "the lens's gap about
  the anchor, once minted" becomes "the gap that carries it". `fragments/lens-audit-and-mint.md:6` (the
  lens's failed-check duty) [REWRITE] with it: the owed gap is the unread or unsound check's; a reading
  that finds the source against the sentence records its gap in the same act, with the mint's judgment flags,
  and its reason is the gap's argument; a found-against citation
  clears once your own later reading of the sentence as it stands backs it — another lens's never does —
  or blue edits or retires it; your close of the gap is refused while the check stands, and a closing
  ruling of the bench closes it while the check still holds the PASS.
  `fragments/blue-research-protocol.md` KEEPs.
  Then `go -C scripts run ./agentgen` regenerates `agents/*.md` and both preloaded skills. Never
  hand-edit the generated text.
- **Phrase sweep (the old model without the word).** The census grep finds the word; it cannot find a
  sentence that teaches "a negative reading is a finding to raise". It matches across line breaks (a
  wrapped sentence is one match; `[^.]` keeps it inside one sentence) and prints the line the match starts
  on. It runs over the whole FEOV tools tree (`tools/internal`, not only `cli/`, and `tools/releasegate`),
  less goldens and testdata, which regenerate. Run, and read every hit for that model:
  `grep -rIl . scripts/agentgen/src plugins/frank-exchange-of-views/tools/internal plugins/frank-exchange-of-views/tools/releasegate plugins/frank-exchange-of-views/skills plugins/frank-exchange-of-views/docs docs | grep -v -E '\.golden$|testdata/' | xargs perl -0777 -ne 'while(/(refut|absent|negative reading|found against|contradicting)[^.]{0,160}?(finding|rais)|(finding|rais)[^.]{0,160}?(refut|absent)/gis){ my $l = 1 + (substr($_,0,$-[0]) =~ tr/\n//); print "$ARGV:$l\n" }' | sort -u`
  At `9ee7ad95e` it prints 56 lines (list kept at `~/.claude/scratch/plan1647-r10/phrase-new.txt`; the same 56
  as at `a31a7b06d`, moved by the line map, and no file #1663 added prints one: `blue-lane.md`,
  `blue-researcher-frame.md` and `blue-search.md` hold none). The 29 beyond `cli/` and the
  agent-text paths, each read:
  REWRITE — `record/anchorchecks.go:279` (`VerifyReads` comment, §III.3), `record/evidenceview.go:245`
  (comment: "at high confidence it is a finding to act on" → the gap its reading minted),
  `record/recordpb/record.proto:509` (`REFUTES` means, §III.1), `record/evidenceview_test.go:250`
  (comment "THE FINDING HALF" → the negative half), `fuzz_test.go:1011` (the quote comment: a negative
  `--url` reading's quote is the sentence its gap is minted on, F13), and the census rows already listed —
  `blockerowner_test.go:200`, `narrowedview_test.go:34`, `redcitation_test.go:347`,
  `migrate/reproduceanchor_test.go:158`, `fuzz_test.go:5493`; DELETE — `record.proto:1143`,
  `dispatch.go:1287`, `contradictionquadratic_test.go:17`, `terms.json:763`; REGENERATE —
  `record.pb.go:737, :8645, :8707`; KEEP — `record.proto:501` (why the outcome field exists, as
  `verify.go:56`), `evidenceview.go:238` (the same history),
  `unchecked.go:29` (the reopened-evidence history), `narrowedview_test.go:49` (fixture text),
  `capture/dispatchparity.go:201`, `debatejs/debatejs.go:409`, `changesjson.go:32`, `viewjson.go:886`,
  `report/assemble.go:1183`, `report/assemble_test.go:997, :1000`, `seatprobe/naming.go:191` (ordinary
  English).
  Known hits at `9ee7ad95e` (27; the 9 under `cli/*_test.go` are the census's test rows or other-sense
  KEEPs — `missingflag_test.go:119`, `stategraph_test.go:91`, `verbs_test.go:114`):
  REWRITE — `scripts/agentgen/src/skills/research-protocol/SKILL.md:17` (the sentence spanning l.17-19,
  which a line grep cannot match), `fragments/lens-audit-and-mint.md:8`, `red-chair.md:12`,
  `verify.go:304`, `verify.go:703`, `help/corroborate.md:13` (DELETED with the page, F13), `help/finding.md:11`, `help/dispatch.md:15`
  and `:19`, `debate.js:915`, `docs/citations-and-bibliography.md:173` ("`refutes` and `absent` are
  findings, not failures to grade"); REGENERATE — `skills/research-protocol/SKILL.md:19`,
  `docs/vocabulary.md:463`; KEEP — `verify.go:56` (a code comment on why the outcome field exists),
  `fragments/lens-audit-and-mint.md:9` (materiality), `debate.js:227` (inquiry `refuted`),
  `docs/seat-surface-naming.md:749`, `docs/why-a-seat-stops.md:103` (ordinary English).
  The property each rewritten sentence holds: a reading of the work that finds the source against the
  sentence mints its gap in the same act, and a reading that reads nothing mints none. New prose says
  "a reading that reads nothing", never "a verdict" (the registry's `verdict` is the chair's); the
  opening of `verify.md:11`, "A VERDICT THAT READS NOTHING LEAVES THE CITATION UNREAD", [REWRITE]s to "A
  READING THAT READS NOTHING …" with the paragraph's F8 clause.
- (F13) Agent text for the one reading verb and the red footnote's deletion, each [REWRITE]:
  `scripts/agentgen/src/red-lens-evidence.md` (a source you find is read with the reading act, and where
  it backs an uncited sentence or refutes one it records the gap; nothing you read becomes a citation),
  `fragments/lens-duties.md`, `fragments/leaf-node-read.md`, `fragments/blue-search.md` (blue: a gap a
  lens's found source carries is answered by citing on the sentence — the source the lens found is in
  `show evidence`), `fragments/lens-audit-and-mint.md`, `src/skills/research-protocol/SKILL.md:14-19` ("The
  bibliography is BOTH sides'" → the bibliography is blue's citations), `references/report_template.md:154`,
  `debate.js:879` (§III.12's F13 census). Prompts name the act, never `--url`. <!-- CHANGED F13 -->
- `debate.js`: enginegen regions regenerate (`go -C plugins/frank-exchange-of-views/tools run
  ./devcmd/enginegen`); authored hits per census [REWRITE]; simulator goldens regenerate.
- Seat prompts name acts, never verbs or flags (`TestNoPromptGrowsItsCommandCatalogue`,
  `TestTheSeatPromptsNameNoVerb`): the new prompt text says "a reading that finds the source against the
  sentence records its gap in the same act", never `--class`.
- AFTER regeneration, every changed golden line and each affected `--help` page (verify, `corroborate`
  before its deletion, finding, mint, dispatch, show, cite) is read before and after <!-- CHANGED F13 -->, one command at a time, for semantic loss.

### III.8 Other Go consumers [MODIFY]

- The census (§III.12) found no FEOV-sense hit in `verify/`, `report/`, `capture/`, `bluedoc/`, `flags/` or
  `scripts/decisions/`: they read the source-outcome sense (`refutes`, `sources_refuted`) or the
  self-contradiction classes, and KEEP. `capture.go:470-476` (found-against citations in the assembled
  report) keeps its check. `cli/chair/dispatch.go`, `cli/setup.go`, `cli/seat/verbs.go`,
  `skills/research/SKILL.md`, `references/run-directory.md` and `docs/citations-and-bibliography.md` are
  REWRITTEN.
- **`seatprobe` attributes a reading's Mint and Anchor to the reading's verb (round-7 GAP 2)** [MODIFY]
  `seatprobe/seatprobe.go`. `Read` maps each event of a seat back to the verb it typed (`verbOfEvent`,
  `:153`; `:203-224`). The reading now writes a Mint, an Anchor and a Verify, so it would count `mint` as
  used by a seat that never typed it, and an unmet `mint` expectation would read as met. Sweep at
  `9ee7ad95e`: `grep -rn 'verbOfEvent\|EVENT_TYPE_MINT\b' tools/internal tools/devcmd`, no tests and no
  goldens, every hit read:
  - `seatprobe/seatprobe.go:153, :212` — REWRITE. A pass over the seat's events first: the `gap_id`s a
    Verify of this seat names with `carried_by` and `carrier_minted` true. `Read` skips a Mint of those
    ids, read from the fields, never from adjacency. A plain mint a later reading names (`carrier_minted`
    false) still counts as `mint`. The Anchor's comment (`:207-208`, "it places the marker of the act
    before it, whose own event names the verb") [REWRITE]: an Anchor places the marker of the act it
    belongs to, whose own event names the verb; a reading's Anchor stands between its Mint and its Verify.
    A Verify counts as `verify`, in either form (F13: one verb; the planned split on `anchor` into
    `verify`/`corroborate` is not built). <!-- CHANGED F13 -->
  - `seatprobe/boards.go:532` (the lens-audit board's `mint` expectation) [REWRITE] its `Because`: it
    argues for a defect no source reading carries, since a defect a reading finds against a source is
    minted by the reading's act and counts under the reading's verb. `:535` (the `corroborate`
    expectation for a standard blue never cited) <!-- CHANGED F13b --> is MERGED into `:534`'s `verify`
    expectation for the same seat, not rewritten beside it — two `verify` rows for `red-lens-evidence`
    would state one verb twice; the merged `Because` names both forms (a cited claim read against its
    source, and a standard blue never cited read with `--url`, whose reading mints the gap blue answers
    with a citation); `:563` KEEP. <!-- CHANGED F13 -->
  - `record/viewjson.go:1605, :1917`, `record/evidenceview.go:586` (a narrowed view's event families),
    `record/gapstates.go:207` (the gap family's acts), `record/record.go:458` (a key's shape),
    `recordpb/record.proto:259` — read a Mint as a gap, whichever verb wrote it: KEEP.
    `record/citationid.go:695` is deleted with the fold.
  `devcmd/seatprobe` reads `seatprobe.Read` and takes the change. Tested by
  `TestAReadingIsAttributedToTheVerbTheSeatTyped` (§III.10); partner §V.5.gg.

### III.9 migrate — one translation [NEW] (contract: record migration)

Archive survey (all 33 run archives in `run-archive/`: the 32 read at `a31a7b06d`, each migrated by that
binary and read with `record.PassBlockers` and `record.CitationsFoundAgainst` — the probe and its output
are at `~/.claude/scratch/plan1647-r6/probe2.out` — and m36, added by #1663 and read at `0c57758e8`, which
holds no negative reading, no raise and no stored `contradiction` blocker): the `contradiction` about-kind exists from epoch
30/31; only **m30** and **m32** hold raising findings, all about **citation** readings; no archive holds
a raised corroboration reading, and only m32 holds stored plans with `contradiction` blockers (events
223 and 310). Eight archives hold a negative reading. Positions are the migrated record's; "ev" is the
archived `events.id` where the tests name it.

| Archived run | Citation readings (carried: found-against while they stand) | Corroboration readings (no anchor) | Contradiction blockers the migrated record holds at its end TODAY | After this change, at its end |
|---|---|---|---|---|
| sqlite-schema-source | `absent` ×2 of `C-65ca0a9e` (pos 80 evidence, 169 dark-side) | — | 2 (subjects 80, 169) | `C-65ca0a9e` found against (`CitationsFoundAgainst` holds it today): carrier-less found-against check(s) by its readers, holding the PASS |
| quadratic-formula | `refutes` of `C-4e4c92af` (pos 553, dark-side) | `refutes` ×2 (pos 187, 531, evidence) | 3 (187, 531, 553) | 187, 531 LOST; `C-4e4c92af`'s sentence was since edited, so no found-against check — it stands `unchecked`, a blocker the record holds today as well |
| wave-a | `absent` ×6 of `C-3743f72c` (pos 131-207, six lenses) | — | 0 | 0 |
| m19 | `refutes` of `C-706fda63` (pos 114, ev 114) | — | 1 (114) | no found-against check (edited, edit 435); `C-706fda63` stands `unchecked`, as today |
| m23 | `refutes` of `C-d2e75055` (pos 401) | `refutes` (pos 400, evidence; the source's sentence) | 1 (400) | 400 LOST |
| 07d | — | `refutes` (pos 194, dark-side) | 1 (194) | 194 LOST |
| m30 (epoch 33) | `refutes` of `C-4f4b8cfa` (pos 427, ev 401, adversary); raise F-8410bec7; answering gap G-dc877265 (`within_source_condition_misattribution`) | — | 0 (answered) | 0: edited at 462 and read `supports` at 479, 482 (`foundagainstm30_test.go`) |
| m32 (epoch 41) | `refutes` ×4 of `C-a36dce3f` (pos 81, 101, 151, 174; ev 80/100/147/169); 4 raises; 1 answering gap G-f417112d, 3 raised and never answered | — | 0 (the edit at pos 294 reopened `C-a36dce3f`; `supports` at 338, 341, 363) | 0 at the end; between pos 174 and 293, four found-against checks (the cut the tests read) |

Every other archive holds no negative reading.

Translation (in `migrate/replay.go`, beside `contradictionRef`, which it replaces; `contradictionRef`
[DELETE]):

1. **A finding about a contradiction — a raise — is written as no event** (the #1695 relay, 2026-11-03:
   findings fold into gaps under #1695, so this plan makes or keeps no finding; it deletes the raise). <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->
   The precedent is `refusedArgument` (`replay.go:321-333`, epoch 38): an archived act the live model has
   no place for is left out, and the residue is stated. Two steps, because the archived word cannot reach
   replay as itself:
   - **Decode.** An archived event's enum words are resolved at decode (`registry.go:172` →
     `valuerenames.go` `enumValue` → `enumNumberOf`), before any replay translation runs, and the live
     `AboutKind` has no `contradiction` value once §III.1 deletes it — so without a row the archived
     finding is refused at decode and replay never sees it. [MODIFY] `migrate/valuerenames.go`: the
     `feov.record.v1.AboutKind` row gains `"contradiction": "anchor"`, beside `"proof": "anchor"`, with
     a comment in the same voice ("a raise names a reading by its event, never an anchor, and replay
     writes it as no event"). The archived finding then decodes as `about_kind anchor` with its archived
     `about_ref`, a reading's `events.id` (e.g. `"401"`).
   - **Replay.** `replay.go:391-402` can no longer key on `ABOUT_KIND_CONTRADICTION` (deleted). It
     recognises the raise by its reference: a Finding whose `about_kind` is `anchor` and whose
     `about_ref` names no anchor — `anchor.Kind(ref) == ""` (`internal/anchor`, the same test
     `record.go:944` puts to an Anchor's id) — is a raise. Never a regex over the ref: a shipped-tool
     regex is ruled out (gblock, 2026-10-18), and the anchor package is the one reader of what an id
     is. The ref is parsed as an `events.id` (`strconv.ParseInt`, as `contradictionRef` does) and carried
     through `newID` to the Verify the migrated record holds there (`batch.Querier()`); a ref that is no
     `events.id`, or names no Verify, is refused with the finding's id named — the loud miss, so nothing
     but a raise is dropped. A recognised raise is written as no event: it enters no `newID` entry, and
     its id is added to the batch's dropped-id set item 3 reads. A Finding about an anchor whose ref IS an
     anchor id (`C-…`, `K-…`, including every archived `proof` respelled to `anchor`, whose `P-` ref
     `remap.apply` has already respelled `K-` with the same hex, `replay.go:303`, before this step) is not
     a raise and is untouched.
   Measured (`~/.claude/scratch/plan1647-f13/arch`, every table and column of m30 and m32 searched for
   each id): 5 raises — m30 `F-8410bec7` (ev 402); m32 `F-d8329c9a` (92), `F-1c990edb` (109),
   `F-b59925d3` (162), `F-5e7dfd3d` (180). Their only structural readers are two `Mint.found_by` entries
   (m30 ev 406, G-dc877265; m32 ev 98, G-f417112d — item 3); every other mention is prose (m30 ev 406
   `mint_reason`; m32 ev 98 `mint_reason`, 194 and 258 `log.text`, 250 `finding.text`, 263
   `regrade.basis`), kept as written. Both gates: `TestEveryArchivedRunMigrates`, and
   `TestTheMigratedM30RaiseIsDropped` below; a delete of the `valueRenames` row turns both red (a decode
   refusal on m30 and m32). `TestValueRenames` (`valuerenames_internal_test.go:15`) holds the row to its
   three conditions. The row's readers are every decode of an `AboutKind` field: `Finding.about_kind`
   (the raises) and `Mint.about_kind`, which no archive holds as `contradiction` (the mint refused that
   kind when it was live, `record.go:966-971`), so the respelling reaches the raises only.
2. **An archived negative reading** carries no `carried_by`: its gap was minted later (forward reference) or
   never. migrate does NOT fill `carried_by` from the `found_by` chain (ruled, gblock 2026-10-30). Its
   found-against state is still a failed check on the migrated record (read from the ledger), with no
   named carrier.
3. **A mint whose `found_by` names a raise** loses that entry: the raise wrote no event (item 1), and
   `requireFindings` (`record.go:1024`) refuses a `found_by` naming no finding, so migrate removes the
   dropped id from the list — a reference migrate drops is translated inside migrate. Two mints:
   m30 G-dc877265 (ev 406) keeps its other entry, `F-9b6251e9`; m32 G-f417112d (ev 98) is left with none. Nothing else about
   `found_by` changes here (#1695's). <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->
4. **A stored plan's `contradiction` blocker** becomes a `failed_check` blocker whose subject is the
   reading's anchor and whose owner is unchanged (the reader, as a live found-against blocker's is).
   `TranslatePlanKinds` cannot do it: it is a pure kind rename (`planKindRenames`) with no record
   access, and the archived subject is the reading's `events.id`, not an anchor. So [NEW]
   `TranslateReadingBlockers(pl, newID, q)` in `migrate/planids.go`, called in `replay.go:373-390`
   after `TranslatePlanKinds` and before `TranslatePlanIDs`: for each blocker whose kind is the archived
   word `contradiction` (a literal in migrate's archived-word table; the live `BlockerContradiction` is
   deleted), it carries the subject through `newID`, reads the Verify the migrated record holds at that
   `events.id` through the batch's querier (`batch.Querier()`, as `CarryPlanVerdict` reads it), and sets
   kind `failed_check`, subject = that Verify's `anchor`; then re-keys and re-prints the plan. A subject
   that wrote no event, or names a Verify with no anchor (a corroboration), is a refusal naming the plan's
   event and the reading — the loud miss. `TranslatePlanIDs`'s contradiction branch (`planids.go:91-103`)
   [DELETE]; the `contradiction` id phrase in `planIDPhrases` stays — it translates ids inside the
   archived `why` prose, which carries the words it was printed with. The table's doc comment
   (`planids.go:14-30`) [REWRITE]: its `event` bullet gains the found-against format, `L's reading of
   citation C at event %d` (`passblockers.go`, `failedFact`), its `contradiction` bullet names archived
   bytes with no live function, and `TranslatePlanIDs`' doc (`:57`) drops "a contradiction blocker's
   subject". `TestEveryIDTheVerbPrintsIsTranslated` (`record/planidphrases_test.go:51`) gains that format
   through `record.BlockerWhy` (`export_test.go:37`): `BlockerWhy` is keyed by kind and renders one state,
   so it [REWRITE]s to return lines, `failed_check`'s once per `CheckState()`, with a distinct sentinel
   `since`; a found-against sentence printing an id after a word `planIDPhrases` lacks turns it red. m32 holds two such plans (events
   223 and 310, each with three `contradiction` blockers, subjects 100, 147, 169, owners
   red-lens-architecture, red-lens-computation, red-lens-adversary).
5. Migrate's comments in `entries.go:65-67` (the credit entry) and `outcomes.go:19` [REWRITE] to state
   archived bytes without the live concept word where they describe the live model.
6. **A record written at this epoch replays its reading acts whole, and every source event of the act keeps its own duties (GAP 3; round-6 GAP 1).** Replay appends one event
   at a time (`replay.go:448`, `batch.Append`), and three write rules of this plan read the reading's
   ACT rather than the record: the budget exemption (§III.5: the Mint's validation is handed the act's
   paired Verify), the one-gap placement (§III.2a: the Mint is placed by the paired `sentence_of`
   Anchor), and the Verify rule that admits `carrier_minted` true only inside the placed append whose
   Mint has that `gap_id` (§III.1). The budget stays on under `Migrating` (`record.go` Mint arm), so a
   capped lens's reading Mint would be refused, and the Verify refused outside its act. Sweep — every
   write rule of this plan by what it reads, under replay: the budget exemption (act; replay-on,
   FIXED here), the one-gap placement (act; off under replay, `onegapperclass.go:49`, so it needs no pair there), the
   `carrier_minted` rule (act; FIXED here), the held-gap decision (live only: replay never meets it, the
   archived Verify already says which), `carried_by` names an open gap in the reading's sentence
   (record; replays as written), the repeat rule and the close gate (off under `Migrating`,
   `failedcheck.go:135, :179`), the completeness-keyed refusal of a carrier on a non-minting reading
   (the body; replays as written), `PlacesToSentenceOf` (a verb guard; replay runs `gapplace`'s check
   instead, §III.2a). [NEW] in `replay.go`, beside the plan and finding translations: before replay, one
   pass over the source stream (`readingActs`) pairs each Verify whose `carrier_minted` is true with
   the Mint whose `gap_id` is its `carried_by` and the Anchor whose `id` is that gap, keyed by each
   member's source `events.id`. The pairing is read from the record's fields, never from adjacency.

   **What the act changes, and what it does not.** Only the WRITE is joined. The replay loop
   (`replay.go:222-501`) owes every source event a set of duties, and a reading act's three members
   keep each one, in their own iterations, as any event does. Swept — every per-source-event step of
   the loop, each read at `9ee7ad95e`:

   | Step (`replay.go`) | On each member of a reading act |
   |---|---|
   | `res.In[old.Word]++` (`:223`) | each member, in its own iteration — unchanged |
   | the correction-chain lookup (`:231-243`), `superseded` (`:228`) | a member is never a chain link or an outcome; a pending act meeting either is refused (below) |
   | `rm.seat(old.SeatID)` (`:245`) | each member; the three must resolve to one seat, else the act is refused |
   | `reg.Translate(src)`, `reproduceAnchors` (`:250-272`) | each member |
   | `opt.AcceptLoss`'s `continue` (`:253-256`) | not reached: a `LossError` is a registry row's documented untranslatable (`registry.go:38`), and neither a Mint, an Anchor nor a Verify at this epoch or any archived one has such a row. It is neither a refusal nor a write, so were a later row to reach a member, the Verify's iteration finds its act incomplete and refuses it (below), and no member is written alone |
   | `record.ParseStamp(old.TS)` and `record.Now` (`:273-279`) | each member's OWN timestamp is kept with its held body, and `Batch.AppendBodies` sets `record.Now` to it before that body's insert — live, the three events carry three stamps, and the migrated record carries the same three |
   | `rm.apply(body, …)` (`:303`) and `refusedArgument` (`:329`) | each member's body, before it is held |
   | the pin, stale, report-read and plan translations (`:348-396`) | not reached: no member is a Dispatch, Stale, ReportRead or Plan |
   | `proseRefs` (`:423`) | each member's body (a no-op on these three types today; called all the same, so a later prose field on any of them is not skipped) |
   | `place.step(body)` (`:432`) | each member, BEFORE the act is written: the Mint's arm skips it (`p.anchored`), the Anchor's runs `gapplace`'s sentence_of (or location) check on the render the act lands on, the Verify's `owed` places nothing (no label). An `anchorAfter` returned for a member is a refusal of the act — an act that places its own gap owes no Anchor |
   | `batch.Append(id, body)` (`:452`) | REPLACED for the act: the Mint's and the Anchor's iterations hold their translated bodies (with source event, seat id and stamp) on the pending act; the Verify's iteration writes all three through [NEW] `record.Batch.AppendBodies(id, []record.StampedBody)` — one savepoint, the three bodies through the batch's transaction in order, the reading's placed append's validation (so each rule reads the act), `record.Now` per body, and each written event and its `events.id` returned in order (`[]record.WrittenBody{Event, ID}`), the id read inside the savepoint after that body's insert |
   | `res.Out[wordOf(ev)]++`, `classMints`, `StatedFills` (`:457-463`) | per member, once the act lands, from THAT member's returned `Event`: three words counted, the Mint's fill and class translation recorded once |
   | `newID[old.ID] = wrote` from `SELECT MAX("id")` (`:484-492`) | per member, from the `ID` `AppendBodies` returned for THAT body: Mint→Mint, Anchor→Anchor, Verify→Verify. `MAX(id)` after the act is never read for a member: it is the Verify's id, so the Mint's old id would map to it and the Anchor's and Verify's get none |
   | the `anchorAfter` append (`:494-500`) | not reached (above) |
   | per-old-id refusals (`res.Refusals`, `OldID: old.ID`) | per member. The act's write refused: one refusal per member, each its own `OldID` and `Word`, the error naming the act ("the reading act of G — source events a, b, c — is one act, and the write refuses it: …"); no member enters `newID`. A member's own step refused before the write (translate, stamp, seat, `place.step`): that member's refusal as today, the pending act dropped, and each other member, held or still to come, refused naming it ("… is one act with event N, which was refused"). No member is ever written alone |

   Structural refusals, each naming the gap and the three source events: the members are not adjacent in
   the stream, not in the order Mint, Anchor, Verify, not of one seat, one is a correction-chain link, or
   the Verify's iteration holds fewer than two members
   — the act wrote them in one transaction, so none of these arises from a record the write made.
   [NEW] `Batch.AppendBodies` (`record/batch.go`, beside `Append`, `:82-114`): one `SAVEPOINT act`, the
   bodies through `appendAct` over the batch's transaction, a refusal of any rolling the savepoint back
   whole (as `Append`'s does), and a `[]record.WrittenBody` of the events written. Name grep at `9ee7ad95e`:
   `AppendBodies`, `StampedBody`, `WrittenBody`, `readingActs` 0 hits; `appendAct` (one body's write) is the sense it
   composes, not a rival.

   Property: every source event of a reading act maps through `newID` to the event written from it
   (Mint→Mint, Anchor→Anchor, Verify→Verify), each is written and counted once at its own source
   stamp, an act is written whole or each member refused, and a record written at this epoch migrates
   with 0 refusals. `TestARecordWrittenAtThisEpochMigratesWithNoRefusal` (tests below) drives it; its
   partners (§V.5.y) turn it red.

7. **An archived labelled corroboration loses its label, footnote and placed citation; the reading comes
   forward as a reading of the source its lens found (F13).** <!-- CHANGED F13 --> <!-- CHANGED F13b --> **Basis: RULED** —
   gblock picked "drop them, state it" for the 15 lens-placed `C-` anchors (relayed by the #1657 session,
   2026-11-03; the option wording is that session's). The 15 cross-check against that session's own
   census (`~/.claude/scratch/anchor-lifecycle/audit-n2/lensplaced.txt`: quadratic 6, b7 4, b9 1, m27 2,
   m29 1, m35 1, total 15), the table below. Decided: TRANSLATE, not keep. The live
   record has no field to hold the label (`Verify.label`, `work_status`, `sha256` are reserved, §III.1),
   and no rule lets a citation anchor stand with no `Cite`; keeping the anchors would need a live field to
   survive migration, which the change discipline forbids. Re-measured against `run-archive/`
   (`~/.claude/scratch/plan1647-f13/refs.txt`, `lab.txt`; every table and column of each archive searched
   for each label, case-insensitively):

   | Archive | Labels (outcome) | Archived Anchor events of a label | `Verify.anchor` naming a label | Report-text fields carrying a label's token | Prose mentions (kept as written) | Bibliography lines the archived report printed |
   |---|---|---|---|---|---|---|
   | quadratic-formula | 6: `c-16fa2ede`, `c-95969800`, `c-2d485fad` (supports), `c-bb0b387b` (weak), `c-a3573578` (bridge), `c-e2616972` (supports) | 0 (migrate places them today: `gapplace.go:81`, owed) | 3: ev 471, 757 (`c-16fa2ede`), 758 (`c-95969800`), all `supports` by other lenses | `Verify.claim` 471, 757, 758; `Finding.location` 506; `Anchor.location` 507 (`f-0f73ba00`); `BlueEdit.new` 715 | 25 events (friction, mint problem/check/fix, positions, opinions, edit text, manifest) | 0 |
   | is-91-prime-b7 | 4, all supports | 0 (owed) | 0 | 0 | 4 (close 170, log 173, inquiry review 175, spot check 176) | 4 |
   | is-91-prime-b9 | 1, `c-d49a32c4` (bridge) | 0 (owed) | 0; `BlueEdit.reopened` names it at 164, 167 | `BlueEdit.old`/`new` 164, 167; `Retire.claim` 166 | 24 events | 1 (blue cited that url too) |
   | is-91-prime-m27 | 2: `C-280e7b44` (supports), `C-cc6a15a8` (bridge) | 2 (ev 333, 335) | 0 | 0 | 0 | 2 |
   | is-91-prime-m29 | 1, `C-6003b65a` (bridge) | 1 (ev 387) | 0 | 0 | 0 | 1 |
   | is-91-prime-m35 | 1, `C-f7ad8bf8` (bridge) | 1 (ev 508) | 0 | 0 | 0 | 1 |
   | **15** | 9 supports, 5 bridge, 1 weak | 4 written, 11 owed | 3 (+2 `BlueEdit.reopened`) | 11 field values | 53 events | 9 (8 urls blue never cited) |

   No `Retire.anchors`, `Close.anchors`, `Finding.about_ref`, `Mint.about_ref` or `SpotCheck.ids` names a
   label; `BlueEdit.reopened` (an id list) names `c-d49a32c4` at b9 164 and 167, and the translation
   removes that element. The translation, in `migrate` alone:
   - a registry entry for the archived columns `verify.label`, `verify.work_status`, `verify.sha256`
     (`registry.go:120` refuses an unmapped column): each dropped, with the reason "the footnote a
     backing corroboration minted; no reading mints one";
   - `gapplace`'s Verify arm (`gapplace.go:81`, `p.owed("verify", label, claim)`) is DELETED: the 11
     owed anchors are not placed; an archived `Anchor` whose id is a label of a labelled Verify earlier in
     the stream (4 events) is written as no event, as item 1's raise is, and enters no `newID` entry;
   - `remap` gains the batch's dropped-label set: in EVERY field the exact token of a dropped label is
     REMOVED (the respelling every field already gets, `remap.go:84-90`, with an empty replacement), so the
     11 report-text values match the render the dropped anchors no longer mark; in `argumentFields` (prose)
     a bare label id is left as written — archived prose naming an id the record no longer holds, the
     residue `proseRefs` already states for other ids; <!-- CHANGED F13b --> **every report-text field the
     strip reaches**, by `remap`'s own classing (`dataFields`, `remap.go:126-140`): `ReportPart.text`,
     `BlueEdit.old`, `BlueEdit.new`, `Anchor.location`, `Cite.location`, `Computation.location`,
     `Finding.location`, `Mint.location`, `Mint.fix_new`, `Retire.claim`, `Verify.claim` and
     `DocketMotion.grounds_quote`; `TestEveryStringFieldIsInOneList` holds the list complete, and the token
     removal runs in every field whatever its list, so a report-text field added later is stripped too.
     WHY IT MUST BE EVERY FIELD: `migrate/proserefs.go:74` (`held`) skips, silently, any id
     `record.HoldsAnchor` rejects, and a dropped label is one — a token left in archived report text
     would vanish from every id list `proseRefs` fills without a word. Measured, it meets none: no
     `Close.anchor_tool` or `Declare.holding` of the six archives names a label (`refs.txt`; b7's four
     labels sit in `close.prose` at 170, which `proseRefs` does not read), and the no-token test (§III.15
     13b) makes any miss loud;
   - <!-- CHANGED F13b --> an id-LIST element (`idFields` lists: `BlueEdit.reopened`, `Retire.anchors`,
     `Close.anchors`, `SpotCheck.ids` and the rest) equal to a dropped label is REMOVED from the list, not
     respelled to `""`: the respelling with an empty replacement turns a list element into an empty
     string, not a removal, and `["F-84381c8e", ""]` would pass a test that only asks for the label's
     absence. b9's `reopened` at 164 and 167 is archived as `[f-84381c8e, c-d49a32c4]` (each, re-read in
     `blue_edit_reopened`); migrated, each is exactly the one respelled `f-84381c8e`;
   - a Verify whose `anchor` is a dropped label (3) becomes a reading of that label's source: `anchor`
     cleared, <!-- CHANGED F13b --> `url` and `title` from a label → (url, title) map built in the pass over
     the SOURCE stream that collects the dropped-label set — never through the batch's querier, because the
     migrated labelled Verify has no `label` column to be found by (§III.1 reserves it). A Verify naming a
     label the map lacks is a refusal naming it (the loud miss);
   - the labelled Verify itself comes forward unanchored — `url`, `title`, `claim` (token removed),
     outcome, confidence, bridge, text, copy fields — carrying no `carried_by`: migrate mints no gap for
     it (its grades would be fabricated, as GAP 2 rules for a refutation), so the 15 readings and the 3
     re-readings hold nothing.
   Tested by `TestTheMigratedRedCitationsAreReadingsOfTheirSources` (§III.15 test 13a). An exact-span
   edit whose bytes span a removed token is the one shape this cannot carry by text alone; the b9 edits
   164 and 167 are the archive's only edits over a label, and the test drives them.
   <!-- CHANGED F13b --> **The dropped Anchor events enter no `newID` entry, and nothing refers to them.**
   Re-measured (`~/.claude/scratch/plan1647/f13b/`): no column of m27, m29 or m35 holds 333, 335, 387 or
   508 as a reference — the one equal value is m35's `seat_turn.cache_creation` = 508, a token count; the
   targeted round's read of `Dispatch.pin`, `Stale.through`, `ReportRead.head`, plan bytes and every text
   column agrees, and found the same of the raise events (m30 402; m32 92, 109, 162, 180). So "enters no
   `newID` entry" refuses nothing on the archives, and the existing pin, stale and report-read refusals
   would make a miss loud (gate: `TestEveryArchivedRunMigrates`).
   <!-- CHANGED F13b --> **The budget on replay.** The budget stays on under `Migrating`, and two things
   could now move it on the archives; both measured, neither refuses:
   (1) *A found-source reading's mint* (F6 amended, §III.5): no archived reading comes forward with a
   `carried_by` (items 2 and 7), so replay meets no found-source mint. 0 refusals by construction.
   (2) *The claims measure shrinks*: `measureClaims` counts the citation tokens of the CURRENT report
   (`mintbudget.go:67`, `claimcount.Count`, one per token on a sentence), and removing the red tokens
   lowers it for every archived mint measured after them. Only the `logic` area reads it
   (`mintScales`, `:96-104`: evidence counts `cite` rows, computation `computation` rows, the other four
   prose paragraphs, which a token removed from a prose sentence does not change). The floor is the
   archive's `mintBudget` — 3 in m27, m29 and m35 (`inputs/run-config.json`), the default 15 in
   quadratic, b7 and b9, whose archives carry no run config. By archive: quadratic has no lens mint; b7
   and b9 hold at most 2 open per lens, under 15; m27 and m35 mint a logic gap with at most 2 open
   before it, under the floor 3; m29's logic lens mints at 3 open (twice, after the red token at 386), so
   its budget is read — the claims measure at each of its seven mints, rendered off the migrated
   record's report ops before each, is 8 with 1 red token counted after event 386 (7 before it): 8 →
   7 removed, against 4 open at most. 0 refusals. (Scratch tool `devcmd/zzbudget` in a copy of
   `origin/main`'s tools, run on m29 migrated by `origin/main`'s `feov-migrate`; not committed.)

**Residue (stated, accepted in this plan; ruled, gblock 2026-10-30):**
- **Archived negative CORROBORATION readings lose their PASS hold (GAP 2, RULED: "accept the loss,
  stated").** Today `contradictionsOf` opens a contradiction for a negative corroboration too, and a
  migrated record holds it as a blocker (`cli/m23contradiction_test.go:73-88`). After the change a
  corroboration's PASS hold is its gap, minted at the act; an archived one carries no `carried_by`,
  migrate mints no gap (the grades would be fabricated) and a corroboration has no anchor for the ledger
  to hold, so it holds nothing. Lost, by archive: quadratic-formula pos 187 and 531, m23 pos 400, 07d pos
  194 — four readings. None of the four moves a recorded outcome (quadratic ended CEILING, m23
  UNVERIFIED, 07d VERIFIED — its PASS was recorded live, before the contradiction existed, and is
  carried as archived), and at its end each of the three migrated records is still refused a PASS by
  blockers it holds today that do not rest on the lost reading (quadratic: `unchecked` citations and
  computations; m23: `stale_area` ×4, the chair's; 07d: `lens_ready` for the six lenses other than the
  reader, which the contradiction never readied). A `lens_ready` of the reader itself may rest on the
  contradiction's readiness reason and goes with it.
- **Archived negative CITATION readings are carried** as the ledger's found-against standing: where one
  still stands (sqlite-schema-source's `C-65ca0a9e` at its end; m32's `C-a36dce3f` between pos 174 and
  293), the migrated record holds a carrier-less found-against failed check per standing reading; where
  an edit has since changed the sentence (quadratic `C-4e4c92af`, m19 `C-706fda63`, m30, m32 at its end)
  the citation is `unchecked` or checked again, as the ledger already says today, and the contradiction
  that outlived the edit is gone. No archive's verdict moves: quadratic and m19 stay held by their
  `unchecked` blockers.
- Archived readings name no gap, so the close gate cannot tie an archived answering gap (m30
  `G-dc877265`, m32 `G-f417112d`) to its reading: in a run resumed after migration that gap may close
  while its reading still stands. Accepted. A carrier-less found-against check readies blue alone, sits
  on its reader's work list as not owed, holds no reader from retiring and admits no budget-exempt mint
  (§III.3); it clears by the ledger's rule — its reader's own later reading of the sentence as it stands
  that backs it, an edit, or a retire.
- An archived `verify` carries no `source_completeness` (the stamp is new on `verify`, and the archive
  holds no fetched-source cache to stamp it from), so an archived `verify` of a `not_the_work` copy keeps
  reading as a reading. No archive is known to hold one; the residue is stated, not measured.
- **The five archived raises are written as no event** (item 1): m30 `F-8410bec7`; m32 `F-d8329c9a`,
  `F-1c990edb`, `F-b59925d3`, `F-5e7dfd3d`. Two were named by a gap's `found_by` (m30 G-dc877265, which
  keeps `F-9b6251e9`; m32 G-f417112d, left with none); the raises' text survives in `run-archive/` and in
  the prose of later events that name their ids (six events, item 1), and nowhere in the migrated
  record. m30 ended with its bench's outcome and m32 UNVERIFIED; no verdict moves. <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->
- **The 15 red-placed citations leave the migrated records** (item 7): their tokens leave every report
  text, and the migrated reports print 8 fewer Bibliography lines (b7 4, m27 2, m29 1, m35 1); the
  archived reports printed 9 such lines, and b9's survives because blue cited that url too — 8 urls the
  reports then name nowhere. <!-- CHANGED F13b --> The readings stand as readings of the
  sources their lenses found, carrying no gap; the 3 later re-readings of `c-16fa2ede`/`c-95969800`
  (quadratic) become readings of those sources. No verdict moves: quadratic ended CEILING, b7, b9, m27,
  m29 and m35 carry their recorded outcomes (archived outcomes are carried, not recomputed). NOT
  MEASURED: whether any end-of-run PASS blocker rests on a red citation — b9's blue edits 164 and 167
  reopened `c-d49a32c4`, so b9's migrated end state may lose an `unchecked` blocker; the item-7 test
  prints each of the six records' end blockers before and after, and the implementing phase states the
  difference here. Prose that
  names a label is archived text and stays as written. <!-- CHANGED F13 -->

Tests [NEW / REWRITE], each driven through `migrate.Migrate` on the real archive, as `feov-migrate`
runs it:
- `TestTheMigratedM30RaiseIsDropped` (`migrate/`; renamed from
  `TestTheMigratedM30RaiseIsAFindingAboutItsCitation`): no event of the migrated m30 is F-8410bec7;
  G-dc877265's `found_by` is exactly `F-9b6251e9`; G-dc877265 open/closed as archived; `res.In` counts
  the raise and `res.Out` does not; 0 refusals. A row on m32: none of its four raises is written,
  G-f417112d has no `found_by`. <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->
- `TestTheMigratedRedCitationsAreReadingsOfTheirSources` (`migrate/`, F13, on the six archives of item
  7): 0 refusals; no migrated Anchor, Cite or report text holds any of the 15 labels; every migrated
  Verify that held a label, and the 3 that named one, has an empty `anchor` and the label's `url` and
  `title`; `record.CitedSources` holds blue's cites only; b9's edits 164 and 167 land, their `reopened`
  without `c-d49a32c4` — <!-- CHANGED F13b --> asserted as the EXACT list, the archived list minus the
  label (`["F-84381c8e"]` in its respelled form), so an element respelled to `""` fails; the 3
  re-readings' `url` and `title` equal the labelled readings' (the source-stream map); each migrated
  report renders (`TestEveryArchivedRunMigrates`' render) and its Bibliography holds no url only a label
  carried. <!-- CHANGED F13 -->
- <!-- CHANGED F13b --> `TestNoMigratedArchiveHoldsALensPlacedToken` (`migrate/`, every archive
  `TestEveryArchivedRunMigrates` migrates): for each of the 15 labels, in BOTH spellings — the archived
  lower-case `c-…` and the respelled `C-…` — and in token form (`<!--cite:…-->`, `<!--C-…-->`) and bare id
  form, no string field of any migrated event's body holds it outside `argumentFields` (prose, kept as
  written), and no migrated id field or id-list element equals it; the render of every migrated report
  holds none. It is the gate the silent skip in `proserefs.go:74` lacks. Partner §V.5.yy.
- `TestTheMigratedM32HoldsAFailedCheckPerStandingReading` (`record/`, beside `foundagainstm30_test.go`,
  through its `archivedEvents`/`archiveAt` helpers, `reopenedm19_test.go:58-94`): on the migrated m32 cut
  after pos 293 — the four readings standing, before the edit at 294 — `failedOf` holds four checks of
  `C-a36dce3f`, state `found_against`, owners red-lens-evidence, red-lens-architecture,
  red-lens-computation, red-lens-adversary, `AnsweredBy() == ""`; on the whole migrated m32, none (the
  edit and the `supports` re-reads); and the two migrated stored plans (the events m32 wrote as 223 and
  310, read through `newID`) each hold `failed_check` blockers with subject `C-a36dce3f` and owners
  red-lens-architecture, red-lens-computation and red-lens-adversary, no blocker of kind `contradiction`,
  and a key that verifies (`requirePrintedPlan` admitted it at the write). A delete of
  `TranslateReadingBlockers`'s call turns it red (a refusal, or a numeric subject).
- `TestAMigratedFoundAgainstCheckReadiesBlueAlone` (`record/`), on the same m32 cut (after pos 293),
  asserting for each found-against failed check of `C-a36dce3f`:
  1. dispatch: the plan's `why` readies `blue-respond` for it and names no reader lens for it;
  2. work list: each reader's item for it is in `may`, never `Open`, and `EnvelopeRefusal` for that reader
     does not name it;
  3. retirement: `lensOwesOf` names no `failed_check C-a36dce3f` for any reader, so a reader with nothing
     else owed retires (the lens fold reports it retired);
  4. budget: an about-anchor mint by a reader is counted (`answersOwnFailedCheck` returns "");
  5. the printed sentence: the blocker's `WorkItem()`, `why`, the PASS refusal and `blueFailedItem` say
     the reading found the source against the sentence and name blue's remedy ("edits the sentence, or
     retires or replaces the citation"), and contain none of "owes its gap about the anchor", "takes no
     mint of its budget", "has minted no gap about it yet", "read nothing of its source";
  6. a backing reading of the sentence by the reader (a sitting later), an edit naming the anchor, and a
     retire of the citation each remove it — three subtests on copies; a fourth subtest, a backing
     reading by ANOTHER lens, leaves it standing (a disagreement clears nothing).
  A second table row runs 1-5 on the whole migrated sqlite-schema-source, whose `C-65ca0a9e` stands
  found against at its end — the real end state, read-only (that run is closed to seat acts).
- `TestM23sArchivedCorroborationHoldsNothing` — `cli/m23contradiction_test.go` [REWRITE], renamed from
  `TestM23sRefutingCorroborationIsRefusedLiveAndItsArchiveKeepsIt` and asserting the migrated state the
  GAP 2 ruling accepts: migrate m23, 0 refusals, 520 events; event 400 is the evidence lens's
  `refutes` corroboration quoting the source's sentence, carrying no `carried_by` and no
  `carrier_minted`; `record.PassBlockers` holds NO blocker whose subject is `400` or whose detail is the
  source's sentence, no `failed_check` and no `found_against` state for it, and still holds the run's
  four `stale_area` blockers (the PASS stays refused at the end); the outcome is the
  UNVERIFIED its bench recorded. The live half stays and drives the act as a seat now runs it: on the
  record before event 400, the same reading as `verify --url` (F13) <!-- CHANGED F13 --> with the gap flags a negative reading needs
  (`--class claim_contradicts_source`, problem, check, check kind, grades) is refused with
  `claimRefusal`'s "--quote was not found in report.md" sentence, and the record holds no Verify, Mint or
  Anchor from it. Its doc comment [REWRITE] states the accepted loss.
- `TestAFindingAboutAnAnchorIsNoRaise` (`migrate/`, a synthetic archived stream at epoch 33: a finding
  `about_kind anchor`, `about_ref C-…`, and one `about_kind proof`, `about_ref P-…`, each after the anchor
  it names). Both come forward with `about_kind anchor`; the second's ref is respelled `K-…` with the
  same hex by `remap.apply` (`replay.go:303`, `hexID`, `remap.go:80-83`) before the raise test runs, and
  the first's is unchanged; 0 refusals — the raise test `anchor.Kind(ref) == ""` passes both by.
- `TestM23CorroborationReadingComesForward` (`migrate/`; the unraised corroboration and citation
  readings, 0 refusals).
- `TestARecordWrittenAtThisEpochMigratesWithNoRefusal` (`cli/`, GAP 3 and round-6 GAP 1): through the
  verbs, lens L at its mint cap records a `verify --as refutes` of C (Mint G, Anchor `{G, sentence_of:
  C}`, Verify `carrier_minted` true), lens M's identical reading names G (`carrier_minted` false), and
  blue accepts G's verified fix on a sentence the report holds twice (an exact-span edit). AFTER L's
  reading, the stream also holds: (1) a chair plan, written by the chair's `dispatch` as it runs, whose
  `why` names L's reading by event ("L's reading of citation C at event N", `failedFact`'s found-against
  sentence, N = L's Verify's `events.id`); (2) a Stale naming G's Mint — `through` = the Mint's
  `events.id` — and one naming each of the act's Anchor and Verify, written at the record layer
  (`record.Append(&recordpb.Stale{…})`, which the write admits with any `through`): no verb writes a
  Stale whose `through` falls inside a reading act (a stale refusal's `through` is a later change on the
  gap, `stale.go:43`), and the property is per source event, so each member is named once. Then
  `migrate.Migrate` from that run directory to a fresh one. Asserts: 0 refusals; `res.In` and
  `res.Out` count `mint`, `anchor` and `verify` alike (each member of the act once on each side); every migrated event's body and `ts` equal its source event's; L's
  open-gap count unchanged; the migrated render byte-identical; the migrated plan's `why` prints the
  migrated `events.id` of L's Verify; and the three Stales' `through` are the migrated ids of the Mint,
  the Anchor and the Verify respectively, each read back off the migrated record by its body (the Mint
  whose `gap_id` is G, the Anchor whose `id` is G, L's Verify). <!-- CHANGED F13b --> The stream also holds a
  `verify --url` reading by lens N, below its cap, that mints `unsourced_claim` on an uncited sentence
  (counted against N's budget live, F6 amended): it migrates with 0 refusals and N's open count is
  unchanged. As specified before round 6 — L, M and
  blue's accept only — it passed with the defect; the plan and the Stales are what fail on it.
  Partners §V.5.y.
- `TestMigrateChecksAReadingsGapBySentenceOf` (`migrate/`): the same stream with the cite of C dropped
  before migration — the Anchor `{G, sentence_of: C}` then names a token the migrated render lacks.
  `validate`'s Verify arm leaves a citation's existence to the verb (`record.go:1736-1739`) and
  `Verify.anchor` has no foreign key, so nothing in the write refuses this stream: `gapplace`'s check is
  what does. `place.step` runs on the Anchor member in its own iteration, before the act's write (item
  6), so the test asserts: the Anchor's source event is refused with `gapplace`'s sentence_of error —
  `errors.Is(…, anchortext.ErrAnchorGone)`, its text naming G and C; the Mint's and the Verify's source
  events are each refused naming that event; and the migrated record holds no Mint of G, no Anchor of G
  and no Verify carrying G. With C kept, it is admitted.
`reopenedm19_test.go`, `foundagainstm30_test.go` and `TestEveryArchivedRunMigrates` keep passing.

### III.10 Tests [MODIFY / DELETE / NEW]

- DELETE the four files that exist for the concept (§III.12): `integration/surface/contradictionanswer_test.go`,
  `record/contradictionquadratic_test.go`, `record/contradictionref_test.go`, `record/answerbudget_test.go`.
  (F13) DELETE the files that exist for `corroborate` and the red footnote: `cli/corroborate_test.go` (its
  stamp rows move to `TestAReadingOfACopyThatIsNotTheWorkMintsNothing`), `cli/corroborationsha_test.go`,
  `record/corroborationsource_test.go`; `record/redcitation_test.go` is replaced by
  `record/foundsource_test.go` (§III.2d's tests). <!-- CHANGED F13 -->
  `cli/m23contradiction_test.go` is rewritten to assert the migrated state the GAP 2 ruling accepts
  and renamed — the file `cli/m23corroboration_test.go`, the test `TestM23sArchivedCorroborationHoldsNothing`
  (§III.9), neither carrying the census word; its live quote-refusal half stays, driven with the gap
  flags a negative reading now needs.
- REWRITE the rest per census; each rewritten assertion tests the new behaviour, not the absence of the old.
- NEW (each with its delete-the-row partner in §V.5):
  - `TestARefutingReadingMintsItsGapInTheSameAct` (`verify --anchor` and `verify --url`; class from the admitted set; <!-- CHANGED F13 -->
    `class_material` = always; `Verify.carried_by` = the gap).
  - `TestASecondRefutingReadingNamesTheStandingGap` (another lens, same sentence, same class: reading
    written, no mint, `gap.minted=false`; a different admitted class mints).
  - `TestANegativeReadingWithoutGapFlagsIsRefused` / `TestABackingReadingWithGapFlagsIsRefused`.
  - `TestTheLensCloseOfAReadingsGapIsRefusedWhileTheCitationStandsFoundAgainst` (record layer): the
    minting lens's close is refused for every closing disposition, naming G and C; a backing reading by
    ANOTHER lens leaves the check standing and the close still refused (the ledger's rule: a
    disagreement clears nothing); then, on copies, the reader's own backing re-read a sitting later, blue's
    edit naming the anchor, and blue's retire of the citation each admit the close.
  - `TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld` (`record/`, mirrors
    `TestABenchRulingDoesNotClearAnUnreadCitation` at `unreadcitation_test.go:85`, on a board where the
    evidence lens's `refutes` reading of C-1 minted G-1): the chair dockets G-1, the bench rules
    `defect_accepted`; asserts G-1 is closed in the gap family (`workStatesOfFamilyT`); `failedOf` still
    holds one failed check of C-1 with `CheckState()` = found_against and `AnsweredBy()` = G-1; the
    chair's PASS is refused naming "C-1 (found against by red-lens-evidence" and "no bench
    ruling clears it"; `PlanDispatch` has `pass_permitted` false and `anchorsStanding` = "; on a failed
    check: C-1". Delete-the-row partner §V.5.n.
  - `TestASupersededReadingGapHandsTheCheckToItsSuccessor` (`record/`): L's `refutes` mints G1; L closes
    G1 `--superseded-by` G2 (admitted); the blocker's `AnsweredBy()` = G2 and its sentences name G2;
    L's close of G2 is refused naming C; L's repeat negative reading is refused naming G2. Branch row:
    G3 is then minted `--supersedes G1` (G1 now has two successors, G2 by its close and G3 by
    supersedes); `AnsweredBy()` = G3 (open, minted last); L's closes of G2 and of G3 are each refused
    naming C; the repeat refusal names G3; after G3 closes `--superseded-by` G4, `AnsweredBy()` = G4
    while G2 is still open. Bench-ruled branch (round-5 note): on a copy after G3 is minted, the chair
    dockets G3 and the bench rules it `defect_accepted` — a closing ruling, which writes no Close
    (`closure_act`'s bench arm) — and `AnsweredBy()` is then G2, the open gap minted last, read from the
    gap view; the check still stands (`CheckState()` found_against) and L's close of G2 is refused naming
    C. Partners §V.5.o and §V.5.aa.
  - `TestAReadingsGapTakesNoMint` (`cli/`, driven through the `verify` verb as a seat runs it, so the
    Mint's budget check meets the act's paired Verify before the Verify is written): a lens at its cap
    records a refuting reading; the reading and its gap land in one transaction; the lens's count is
    unchanged after; a plain `mint` at that cap is still refused. <!-- CHANGED F13b --> The refuting reading is
    a `verify --anchor`; a `verify --url` at the cap is refused for budget (F6 amended), which
    `TestAFoundSourceGapTakesAMint` (§III.2d) drives. Partner §V.5.e.
  - `TestAPlainMintALaterReadingNamesKeepsItsMint` (`cli/`): L mints G with plain `mint`, class
    `claim_contradicts_source`, quoting sentence S; a sitting later L records a `refutes` reading of S
    with that class — the one-gap rule refuses the reading's mint, `carried_by` = G, `carrier_minted`
    false; L's open-gap count still counts G (at cap, a further plain mint is refused). Successor row: a
    reading's gap G1 (`carrier_minted` true, outside the budget) and L at its cap; L's plain
    `mint --supersedes G1` is refused for budget — a successor is a plain mint and counts — while the
    check it would carry on stands. Partners §V.5.p and §V.5.bb.
  - `TestAReadingOfAFoundAgainstCitationIsWordedByItsState` (`cli/`, driven through `verify --json` and
    its human output): lens A's `refutes` of C mints G. Rows: (a) A's own later `unreachable` reading;
    (b) lens B's `unreachable` reading; (c) B's `supports` reading; (d) carrier-less — the negative
    reading written under `Migrating` with no `carried_by`, then (a) and (b) again. Each asserts
    `failed_checks[]` holds C's check with `state` = found_against, `owner` A, `answered_by` G (or "" in
    (d)); the human text says C stays found against and names blue's remedy; (c) also says the backing
    reading "clears nothing"; and none of "stays unread", "you owe a gap", "takes no mint of your
    budget", "read nothing first", "cites a source a lens can read" appears. An `unread` row keeps
    today's three sentences. Partner §V.5.l.
  - `TestARetriedNegativeReadingIsRefusedNamingItsGap` (`cli/keyedretry_test.go`): the same
    `verify --as refutes …` run twice in one sitting — the second refused (conflict), naming the first
    reading's event and its gap from the first's `--json`; the record holds one Verify and one Mint.
    The same for `verify --url --as refutes`; <!-- CHANGED F13b --> `verify --url --as absent` of one url and one
    sentence twice meets `requireNotARetriedReading`, its refusal naming the first reading's `events.id`,
    and the record holds one Verify and no Mint; the same `absent` of that url for a second sentence is
    recorded. <!-- CHANGED F13 --> Partner §V.5.m, vv.
  - `TestAFoundSourceAbsentMintsNothing` (F5; renamed from `TestACorroborationAbsentMintsNothing`). <!-- CHANGED F13 -->
  - `TestAReadingsGapIsFilledAsAPlainMints` (`cli/mintreason_test.go`, driven through `mint` and both
    forms of `verify`): one lens gives the same judgment flags and `--reason` to a plain `mint --quote S1`, a
    `verify --anchor C --as refutes` of S2 and a `verify --url U --title T --as refutes --quote S3` <!-- CHANGED F13 -->, and the test reads
    the three stored Mints. It walks `Mint`'s descriptor (`ProtoReflect().Descriptor().Fields()`): for
    every field, presence and value equal the plain mint's, or the field is in the test's two named
    sets — differs by subject (`gap_id`, `location`) or by presence only (`fix_new`) — so a field added
    to `Mint` fails until it has a disposition here and in §III.2's table. Rows: no `--new`, no
    `--complexity`, no `--fix` (each absent alike, `fix_basis` `proposed`); all three given (`fix_basis`
    `verified`, `fix_new` set). GAP 3 row (what the verb TELLS the seat back): with a `--check` holding
    shell-looking text (backticks and `$(…)`, passed as argv so nothing expands it), `mint`'s human text
    and each reading's human text holds "acceptance check RECORDED as: " followed by the
    stored check and "(read it back —", and each reading's `--json` `gap.acceptance_check` equals the
    stored Mint's; with a `--problem` that trips a voice tell, each reading's human text holds `mint`'s
    note ("the risk matrix, the first sentence of the problem") under the gap and not "the source's note
    and Bibliography entry" for it; a second lens's reading that names the standing gap prints no
    "acceptance check RECORDED". Partners §V.5.ee, ff, kk.
  - `TestAReadingsGapCarriesItsArgumentWhereverAMintsDoes` (`cli/mintreason_test.go`, beside
    `TestAGapCarriesRedsArgumentAndNotOnlyItsProblem`): lens L's `verify --anchor C --as refutes --reason R`
    with no `--new` mints G. Asserts: `show board --json`'s row for G has `mint_reason` R and `fix_basis`
    `proposed`; blue's `show work --json` item for G has `fix_basis` `proposed` (never omitted); the gap
    view's row agrees; `show report` holds `Token(G)` at the sentence's end and not R, the live text
    carrying no argument; `report.Assemble` over that run — what the bench's `assemble` verb runs
    (`cli/bench/assemble.go:42`); it returns `{Report, Documents, Verdict}` (`assemble.go:429-458`), so
    the test reads the files it wrote — `<run>/docket.md`'s "The board" open-gap block (`layout.go:167`)
    holds "minted because: R" and "fix_basis: **proposed**" in G's block, and `<run>/report.md` does NOT
    hold R. After L's own backing re-read in a later sitting and its close, a second `report.Assemble`
    leaves "minted because: R" under G in docket.md's closure index. Row: lens M's identical
    reading names G with `--reason R2`; M's Verify holds R2, and G's `mint_reason` stays R. Row: `verify
    --as refutes` with every gap flag and no `--reason` is refused naming `--reason`, and the record holds
    no Verify, Mint or Anchor from it. A `verify --url --as refutes` row repeats the first asserts. <!-- CHANGED F13 --> Partner
    §V.5.ee, ff.
  - `TestAReadingsGapNamesItsSource` (`cli/`, F12, driven through both forms of `verify`, `show
    evidence --json` and `report.Assemble`): lens L's `verify --anchor C --as refutes` (C cites url U1)
    mints G; lens N's identical reading names G; lens M's `verify --url U2 --as refutes --quote S2`
    mints G2; lens P's `verify --url U3 --as supports --quote S3` of an uncited sentence mints G3
    (`unsourced_claim`), and its row and docket.md's line for G3 name U3 — how blue finds the source a
    lens found (F13) <!-- CHANGED F13 -->; lens K's backing `verify` of another citation carries nothing. Asserts: `show evidence`'s
    row for L's reading under C's source has `carried_by` G, N's has G, M's row under `independent` has
    G2, K's row has no `carried_by` key; docket.md's open-gap block for G holds one source line per
    carrying reading — naming U1, the reader, the outcome and the reading's event, twice (L, N) — and
    G2's names U2; report.md holds none of them. Partner §V.5.ll.
  - `TestAReadingIsAttributedToTheVerbTheSeatTyped` (`cli/`, drives the verbs, then `seatprobe.Read` and
    `seatprobe.Check` over the run): lens L's refuting `verify` → `Used["verify"]` 1, `mint` absent from
    `Used` and present in `Unused`, and an expectation of `mint` is unmet, with the substitute `verify`;
    lens M's refuting `verify --url` → `Used["verify"]` 1, no `mint`, and no `corroborate` anywhere in the
    probe's verb set (F13) <!-- CHANGED F13 -->; L's later
    plain `mint` → `Used["mint"]` 1, and a reading of that sentence by lens N (`carrier_minted` false)
    leaves it 1. Partner §V.5.gg.
  - `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` (`cli/`, on the pattern of
    `TestAbsentIsRefusedOnACitationWhoseCopyIsOnlyTheAbstract`, `verifycopy_test.go:257`: the run's
    cache entry for the source is stored with `Completeness = fetchcache.CompletenessNotTheWork` — after
    the cite where the cite needs a copy that is the work — and the reading resolves to it). Rows, each
    driven through the verb: (a) `verify --as refutes` with the gap flags → refused naming
    `not_the_work`, nothing recorded; (b) the same without the gap flags → one Verify, `source_completeness`
    `not_the_work`, no Mint, no Anchor, `carried_by` and `carrier_minted` unset, `--json` has no `gap`;
    the citation's failed check has state `unread`, owner the reader, and the human text carries
    today's unread sentences ("stays unread", "you owe a gap about it"), opens on "this reading reads
    nothing of the source" and contains no "REFUTES" (GAP 2, `human()` reads before outcome); `show
    evidence --json`'s row for the reading has `source_completeness` `not_the_work`, and
    `EvidenceVerificationJSON.FindsAgainst()` is false on it; (c) `verify --as absent` →
    refused by the existing body-less-copy refusal, nothing recorded; (d) `verify --url --as refutes`
    with the gap flags → refused naming `not_the_work`; <!-- CHANGED F13 --> (e) the same without them → one Verify, no
    Mint, no Anchor, no carrier; (f) `verify --as unreachable` with the gap flags → refused. A seventh
    row writes, at the record layer outside `Migrating`, a `not_the_work` `refutes` Verify carrying
    `carried_by` and one carrying `carrier_minted` → each refused by the write-path rule. GAP 1 rows
    (RULED: every reading of a `not_the_work` copy is unread, a backing one included): (g) `verify --as
    supports` of the `not_the_work` copy → one Verify, its stored `source_completeness` `not_the_work`
    (the stamp `verify` now writes), no Mint, no Anchor, no label; the citation's failed check has state
    `unread`, owner the reader, and the human text carries "stays unread"; (h) `verify --url --as
    supports` of a `not_the_work` copy → one Verify, no Mint, no Anchor, no carrier — `record.CitedSources`
    and `show evidence`'s sources hold nothing for it — and the same seat's `verify --url --as supports`
    of ANOTHER claim from that url in the same sitting meets the duplicate-key refusal on the url (a
    reading that mints nothing keys on its url, §III.1), nothing recorded; the `abstract` and `not_asked`
    stamp rows moved from `cli/corroborate_test.go:301-321` run on `verify --url` <!-- CHANGED F13 -->; (i) `verify --as supports`
    of a cached `full` copy and of an uncached source → `source_completeness` `full` and `not_asked`.
    Partners §V.5.s and §V.5.w.
  - `TestAVerifyOfARepeatedSentencePlacesItsGapByTheAnchor` (`cli/`, driven through `verify`,
    `verify --url`, `show board` and `blue edit --accept` as the seats run them) <!-- CHANGED F13 -->: a report holding one
    sentence's words twice, in two sections, each copy cited (C-1 on the first, C-2 on the second), and
    a third sentence, unique, whose citation C-3 backs a clause in its middle (its token before the
    sentence's last words). Rows:
    (a) lens L's `verify --anchor C-2 --as refutes …` lands: one Verify, one Mint, one Anchor with
    `sentence_of` = C-2 and no `location`; the rendered report holds `Token(G)` at the end of the second
    copy and not in the first; `show board`'s sentence for G is the second section's; `gapSentenceIn`
    puts G in C-2's sentence;
    (b) a second lens's `verify --anchor C-2 --as refutes` with the same class names G (`carrier_minted`
    false) — the one-gap rule placed it by identity, not by text;
    (c) a `verify --anchor C-1 --as refutes` with the same class mints a second gap (a different
    sentence, the same words);
    (d) clause-cited: `verify --anchor C-3 --as refutes …` places its gap's token at the END of C-3's
    sentence — after the sentence's last word, not right after `Token(C-3)` — at the offset a plain
    `mint --quote <that whole sentence>` gives (asserted by placing both on copies and comparing);
    (e) `--new` on a reading over the repeated sentence (row (a) with `--new`) is verified (`fix_basis`
    `verified`); blue's `edit --accept G` then applies it to the SECOND occurrence only, records the edit
    with `exact_span`, and the first copy is byte-unchanged; `show board`'s pair for G is the pair the
    accept applied (the board and the accept read one `ProposalOver`);
    (f) `--new` on a reading of the unique clause-cited sentence (row (d) with `--new`) is verified and
    accepted, and the accept moves G's token to the end of `--new`;
    (g) a `--new` the accept would refuse (it drops C-2's token without keeping the sentence) on a
    `verify` is refused naming "the sentence your --anchor's token sits in", and the refusal contains no
    "--quote";
    (h) `verify --url U --as refutes --quote <the repeated sentence>` (with the gap flags) <!-- CHANGED F13 --> is refused naming
    both occurrences (`WhereAmbiguous`), nothing recorded;
    (i) a replay of the record (`reportproj.Replay` from the empty text) renders byte-identical to the
    write-time render, the exact-span accept included. Partner §V.5.t and §V.5.z.
  - `TestTheReadingsHeldGapIsDecidedInsideItsWrite` (`cli/` or `record/`, through the reading's placed
    append): lens A's refuting reading of S runs under `runtest.Overtaken` (`record/runtest/overtaken.go:25`,
    which commits other seats' acts at `record.WriteBegins`), whose acts are lens B's refuting reading
    of S with the same class — the Mint G, its Anchor and B's Verify carrying G — landing after A's verb
    has read the board and before A's transaction begins. A's act lands, not
    refused: one Verify by A with `carried_by` = G, `carrier_minted` false; the record holds one Mint
    on S of that class. Partner §V.5.u.
  - `TestAReadingOverAnAppliedPrescriptionMintsItsGap` (`cli/`, F10, driven through `mint`, `edit
    --accept`, both forms of `verify` and `show work` as the seats run them) <!-- CHANGED F13 -->: sentence S, cited by C-1
    whose token ends it; lens P mints G0 with `--quote S --new S′` (verified), where S′ holds C-1's token
    as `show report` prints it, and blue's `edit --accept G0` applies it verbatim. Rows: (a) control —
    lens L's plain `mint --quote S′ …` is refused as an estoppel naming G0, and one `LOG_TYPE_ESTOPPEL`
    entry is written; (b) L's `verify --anchor C-1 --as refutes --class claim_contradicts_source …` is
    admitted: Mint G1, Anchor `{G1, sentence_of: C-1}`, Verify `carried_by` G1, `carrier_minted` true;
    `record.EstoppelRejectionsOf` is unchanged from after (a); G1 is open; `PassBlockers` holds C-1's
    `failed_check`, `CheckState()` found_against, `AnsweredBy()` G1, and the chair's PASS is refused; (c)
    on a copy of the record after (a), lens M's `verify --url U2 --title T2 --as refutes --quote S′
    --class claim_contradicts_source …` is admitted, mints G2, writes no estoppel entry, and G2 is a
    material open gap holding the PASS (F13: a found-source `refutes` admits that class alone, so the row
    runs where (b)'s G1 does not hold the sentence) <!-- CHANGED F13 -->. Partner §V.5.cc.
  - `TestTheReadingVerbsDescribeTheirGapFlagsByTheirSubject` (`cli/`, renders `verify --help` and `mint
    --help`): `verify`'s page names no `unclassified`, says it takes exactly one of `--anchor` and `--url`,
    its `--new` names "the sentence your --anchor's token sits in, or your --quote", its `--class` lists
    exactly the four classes the two facets admit and a class outside them is refused at the parse; its
    `--reason` opens with its `ReasonIs` text and names the gap's argument, and the seven required flags
    open "REQUIRED unless"; `mint`'s flag text is byte-equal to the manual golden's; `lens corroborate`
    is an unknown command. Partner §V.5.dd, hh. <!-- CHANGED F13 -->
  - `TestEveryNegativeOutcomeNamesItsClasses` (F13: and `found_source_gap_classes` is non-empty exactly on
    `supports`, `supports_with_bridge`, `weak` and `refutes`; <!-- CHANGED F13 --> the `gap_classes` facet is non-empty exactly on
    `refutes` and `absent`, and every named class exists).
  - `TestAFoundAgainstCheckIsWordedByItsState` (live record, `record/`): a lens's `refutes` reading
    mints G. Asserts the found-against blocker's `CheckState()`, `AnsweredBy() == G`; `WorkItem()`,
    `why`, the PASS refusal, `blueFailedItem` and blue's `failedCheckReadiness` carry "gap G carries it"
    and the found-against remedy, and none of the owed-mint sentences; the reader's work list holds it as
    `may`; `lensOwesOf` does not name it; the reader's about-anchor mint on that anchor is counted. Its
    clear sentence (`failedClears`, the close refusal, the PASS refusal) names the reader's own later
    reading and says another lens's backing reading clears nothing, and none contains "when a lens's
    reading" or "a lens reads its source"; none says the gap "does not close". A second table row runs
    the `unread` state and asserts its owed-mint and remedy sentences are unchanged ("owes its gap about
    the anchor", "Retire the citation") — the state, not the carrier, picks the wording.

### III.11 Release fuzz harness — `releasegate/fuzz/` [MODIFY]; PR #1642 [SUPERSEDED]

- Delete the contradiction drives and census fields: `answerContradictions` (`fuzz_test.go:2513-2530`, called at `:5291, :5588, :8890`, named in `dischargeOwed`'s comment `:1128-1133`), the
  `ownAnswers`/`answerBudgetErr` fields (`:499-502`), the sweep gates at `sweeptally_test.go:922, :956`
  and every other site the census lists.
- **The main reading drive** (`fuzz_test.go:5406-5491`; the raise/`contradictionOf` half at
  `:5413-5425` is deleted with the census). A lens's reading draws the gap flags (class from the
  outcome's admitted set, grades, check) exactly when the reading will find against: the outcome has
  `gap_classes` AND the copy the reading resolves to is not `not_the_work` — the drive reads that from
  the run's cache entry for the source (`fetchcache` lookup by url, the entry the verb stamps from),
  never from the outcome alone; a `verify --url` reading draws them from `found_source_gap_classes`
  (a backing outcome on an uncited sentence of the stock, `refutes` anywhere), and its `absent` draws none
  (F5). <!-- CHANGED F13 --> Every drive that called `corroborate` (`fuzz_test.go`, `acts_test.go`, `copy_test.go`)
  calls `verify --url`, and no drive expects a label, a footnote or a red `C-`. <!-- CHANGED F13b --> The
  `verify --url` drive's oracle expects the budget (F6 amended): a minting reading by a lens at its cap
  refused for budget, nothing recorded, and below the cap admitted and counted; and the key: a second
  url against a sentence the drive already read in that sitting recorded, naming the standing gap, and
  an identical repeat in the sitting refused naming the stored reading. After the act it checks
  the result against the stored Verify: `gap` present iff `carried_by` is set; and where
  `carrier_minted` is true, the stored Mint's `mint_reason` equals the `--reason` the drive passed and its
  `fix_basis` is `verified` where `--new` was passed and `proposed` where it was not. Any mismatch is a
  `noteActErr`.
  - **Its found-against repeat** (round-4 note). The drive re-reads the same anchor across sittings
    with a random outcome, and its one repeat guard today is `ownUnreadGapOpen` (`:3735`, `:5429-5433`,
    the `unread` arm). A second reading by the same lens with the same outcome, for which `FindsAgainst`
    holds, of an unchanged sentence while a gap of its earlier reading's carrier tree is open is now
    refused (§III.2's repeat rule), where the drive expects it admitted. [NEW] `ownFoundAgainstGapOpen
    (seatID, anchor, outcome)` beside `ownUnreadGapOpen`, read from the record (the seat's last
    `FindsAgainst` reading of the anchor, its `carried_by` tree, the sentence's last edit): where it
    holds, the drive by coin either changes the outcome to a backing one (admitted, and on a later
    sitting it is the `foundAgainstClearedByOwnReading` drive below) or drives the repeat as refused —
    expecting a conflict naming the earlier reading's event and `AnsweredBy()`, nothing recorded — and
    counts `foundAgainstRepeats`. An admitted repeat, or a refusal naming another gap, is a
    `noteActErr`.
  - **A copy that is not the work** (GAP 1). The fuzz's source stock gains one source whose cache
    entry is stored with `Completeness = not_the_work` (`fetchcache.EntryFor` + `fetchcache.Store`, as
    `copy_test.go:36-37` stores a substituted copy), cited once. A lens reading it with a negative
    outcome by coin passes the gap flags (expects refused naming `not_the_work`, counted
    `notTheWorkFlagsRefused`) or omits them (expects admitted with no `gap`, no `carried_by`, and the
    citation's failed check in state `unread`; counted `notTheWorkReadingsUnread`); with a backing
    outcome (GAP 1, RULED) it expects the same unread state and, after the act, the stored Verify's
    `source_completeness` `not_the_work` (counted into `notTheWorkReadingsUnread`); a backing
    `verify --url` of that source expects no gap and nothing in `show evidence`'s sources <!-- CHANGED F13 --> — a
    mismatch is a `noteActErr`.
  - **A cited sentence the report holds twice** (F9). The fuzz's report stock gains one sentence whose
    words occur in two sections, each copy cited by its own citation. A lens's refuting `verify` of
    either citation is expected to mint (or name) a gap whose token renders at the END of that
    citation's sentence — after its last content character and any closing quotation mark, and BEFORE
    a citation token that ends the sentence (§III.2a: `of`'s token at the sentence's end stays after the
    gap's), in the occurrence holding that citation's token and in no other — the drive reads the
    rendered report — counted `repeatedSentenceGaps`; a
    `verify --url` whose `--quote` is that sentence is expected refused as ambiguous. <!-- CHANGED F13 --> Where the reading
    passed `--new`, blue's next sitting accepts that gap's fix by coin and expects it applied to the
    cited occurrence only, recorded with `exact_span` (counted `repeatedSentenceAccepts`).
  - **A reading over prescribed text** (F10). Where blue has applied a red `--new` verbatim (the
    estoppel drive's precondition, `fuzz_test.go:664-674`), a lens's negative reading of that text — a
    `verify --anchor` of a citation the accept carried onto it, or a `verify --url --quote` of it — is expected <!-- CHANGED F13 -->
    ADMITTED with its gap minted or named, and NO new `LOG_TYPE_ESTOPPEL` entry; an estoppel refusal of
    a reading is a `noteActErr`. Counted `readingsOverPrescriptions`.
- New gates (false-green guards, zero fails the sweep), each counted from the stored record, never from
  `--as`: readings that minted (`carrier_minted` true); readings refused a mint by the one-gap rule
  (`carried_by` set, `carrier_minted` false — the duplicate path is driven on purpose by re-reading a
  sentence already holding a reading's gap); found-against closes refused; `foundAgainstRepeats`;
  `notTheWorkFlagsRefused` and `notTheWorkReadingsUnread`; `foundSourceGaps` (F13: `verify --url`
  readings whose Mint is `unsourced_claim` or `claim_contradicts_source`) and `markedSentenceRefused`
  (a backing `--url` reading of a cited sentence, refused) <!-- CHANGED F13 -->; <!-- CHANGED F13b -->
  `foundSourceBudgetRefused` (a minting `--url` reading refused at the cap) and `foundSourceRetriesRefused`
  (an identical `--url` repeat in a sitting, refused); `repeatedSentenceGaps` and
  `repeatedSentenceAccepts`; `readingsOverPrescriptions`. Each new field is
  carried through `sweepTally`, the fold and its merge (the sites listed below for the found-against
  fields). New tally: max open gaps per (class, sentence) — must be ≤ 1.
- **Every fuzz branch that tells failed checks apart keys on `CheckState()`**, and the `unread*` tallies
  count the `unread` state only. Sweep, run at `9ee7ad95e` (80 hits, `new-fz.txt` beside the other lists):
  `grep -n -E 'BlockerFailedCheck|AnsweredBy\(\)|\.Detail\b|unread(Answers|Repeats|Retires|RetiresJoined|ClosesRefused)|unsound(Answers|Repeats|ClosesRefused)' releasegate/fuzz/*.go`,
  every hit read. Under a found-against check `Detail == ""` reads as "unread citation", so each of these
  would drive or count it as unread:
  - `fuzz_test.go:1229, :1238` (`fixFailed`, blue's remedy): `Detail == ""` → unread; it would call
    `envelopeRefusedEarly(… "check of citation "+C)` expecting the unread wording, retire the citation and
    count `unreadRetires`/`unreadRetiresJoined`. [REWRITE] a switch on the state: `unread` as today;
    `found_against` drives blue's found-against remedy — by coin, the edit that rewrites the sentence
    (`reopened` names C) or the retire `--answers` the carrying gap — with
    `envelopeRefusedEarly(… "reading of citation "+C+" at event")`, the found-against `failedFact`'s
    words, and counts `foundAgainstEdits` / `foundAgainstRetires`; `unsound` as today; any other state is
    a `noteActErr`, never a fall-through;
  - `:3026` (unread repeat: `verify --as unreachable` expected refused): `CheckState() == unread` only —
    an `unreachable` reading over a found-against check is admitted, so driving it as refused would
    fail the run; `:3036` (unsound repeat): `== unsound`;
  - `:3058, :3061` (the owed-gap mint, `citation := Detail == ""`, its early-envelope refusal and the
    `unreadAnswers` count): only `unread`/`unsound` with no carrier — `found_against` is never owed;
  - `:3115-3124` (`driven := unreadClosesRefused > 0`, the drive of a close refused at the mint) and
    `:4265-4289` (`refuseStandingClose`'s tally, `Detail == ""` → `unreadClosesRefused`): the tally
    switches on the state — `unread`, `unsound`, `found_against` → `foundAgainstClosesRefused` — and
    any other state is a `noteActErr`;
  - `:3434` (blue's oracle, "Retire the citation" for every `Detail == ""`): `unread` only; a
    `found_against` check's item must name the found-against remedy ("edits the sentence, or retires or
    replaces the citation");
  - `:3735` (`ownUnreadGapOpen`): `== unread`;
  - `:4208, :4253` (`standingFailedCheckOf` → `refuseStandingClose` in `closeGap`): state-agnostic
    selection by carrier, kept, and now the successor-following `AnsweredBy()` (§III.3); the tally it
    reaches switches as above;
  - KEEP: `:2972` (asserts the plan readies the owner or blue, true of every state), `:6990` (no
    VERIFIED over any failed check);
  - `sweeptally_test.go:843-854`: the `unread*` gates stay and now count only the unread path; NEW
    gates (zero fails the sweep, false-green guards) `foundAgainstClosesRefused`,
    `foundAgainstEdits + foundAgainstRetires`, and `foundAgainstClearedByOwnReading` (the drive below);
    each new field is carried through `sweepTally`, the fold and its merge (`sweeptally_test.go:189-190,
    :535-537, :657-659`, `fuzz_test.go:468-470, :5726-5728, :7057-7059`), which
    `TestSweepTallyRoundTripsEveryField` and `TestTheFoldReadsEveryRunFieldAndWritesEveryFoldField`
    already fail on a field left out.
  New drive: a refuting reading followed, in the next sitting a plan readies that lens for, by a backing re-read by the same lens, after
  which the gap's close is admitted (`foundAgainstClearedByOwnReading`), and by coin a backing reading
  by another lens first, after which the close is still refused.
- **The test of the keying, outside the gate** (`go test ./releasegate/fuzz/` runs it; the sweep does
  not run locally): `TestTheFailedCheckDrivesKeyOnTheState`, on the pattern of
  `registerbeforeappend_test.go` (`buildBinary`, `recordtest.TmpRun`, `newRunner`): a record holding one
  `unread` check by lens U carried by G-u, and one `found_against` check by lens L carried by G-f (L's
  `refutes` reading through the built binary). It calls `answerFailed(L)`, `answerFailed(U)`,
  `fixFailed("blue-respond")` and `refuseStandingClose` on both gaps, and asserts: no act error (an
  `unreachable` by L is never driven as a refused repeat; blue's envelope is checked against the
  found-against words); `unreadRepeats`, `unreadRetires`, `unreadClosesRefused` each count U's check
  alone; `foundAgainstClosesRefused` = 1 and `foundAgainstEdits + foundAgainstRetires` = 1. Reverting
  any one branch above to `Detail == ""` turns it red. Partner §V.5.q.
- PR #1642: closed as superseded (F7, ruled). Its four diagnostic tally fields are carried into this PR's
  `sweeptally_test.go`, each diagnosing a CEILING run independently of the contradiction:
  - `final_blockers` — the PASS blockers standing at each run's end, by kind;
  - `final_parties` — the parties the last dispatch readied;
  - `open_by_found_by` — open gaps at the end, re-keyed by what surfaced each now that contradictions are
    gone: `reading` (minted by a reading act: a Verify names it with `carrier_minted` true), `finding` (minted from a finding in
    `found_by`), `unsound_computation` (an about-anchor gap carrying an unsound judgement — #1642's
    `unsound_proof`, renamed by #1644; kept because it is the key #1642's data says holds the runs),
    `none`. Kept as F7 ruled: it reads `found_by` and makes no finding; the `finding` key is #1695's to
    re-key when `found_by` folds (§III.16) <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->;
  - `mints_by_epoch` / `closes_by_epoch` — mints and closes per epoch.
- Release gate: PR CI only. Never `FEOV_RELEASE_GATE=1` locally.

### III.12 Consumer census

Searches, run at `9ee7ad95e` over the tree as committed (raw hits kept at
`~/.claude/scratch/plan1647-r10/census-new.txt`, per-file counts in `cnt-new.txt` beside it):

```
git grep -I -n -i contradict 9ee7ad95e -- . ':!run-archive'
git grep -I -n -E 'ABOUT_KIND_CONTRADICTION|unanswered_contradictions|ContradictionOfReading|raisedBy|answeredBy|raised_by|opened_by|sources_refuted' 9ee7ad95e -- . ':!run-archive'
```

Result: **208 files**, 1,599 hits, each read for meaning. Against `0c57758e8`'s 207 files and 1,596
hits (`cnt-old.txt`, and a by-content diff of the two hit lists), what moved: `record/about.go` gains one
(`aboutShapes`' comment, #1674: "a contradiction's id is a number, which requireOpenContradiction reads"
— REWRITE, §III.6); `record/viewjson_test.go` is new to the census (2 hits, #1678's estopped-row test:
"a closer the board may contradict", ordinary English — KEEP); and lines rewritten in place, each re-read
— `cli/lens/finding.go:145-146` (#1674's `--about` text, still naming "a contradiction's id … under
unanswered_contradictions" — REWRITE as before), the inlined copies in the seven `agents/red-lens-*.md`
(REGENERATE), `debate.js:915` (#1677's chair prompt, still ending "the lens that read a contradicting
source no finding raises" — REWRITE), the chair golden (REGENERATE), `record/dispatch.go:22` and
`help/dispatch.md:19` (#1677 dropped `inquiry` from the gavel list; the contradiction clause stands —
REWRITE as before), `record/blockerowner_test.go` (a `want` reorder — REWRITE as before).
Files this plan changes beyond the word census (none carries the word), found by the sweeps of §III.1,
§III.2, §III.2a, §III.2b, §III.3, §III.7, §III.8 and §III.9: `help/close.md`, `help/reproduce.md`,
`fragments/blue-search-and-cite.md`, `docs/seat-command-triggers.md`, `recordpb/facets.go`,
`cli/keyedretry_test.go`, `unreadcitation_test.go`, `unreadb6_test.go`, `migrate/valuerenames.go`,
`migrate/remap.go`, `migrate/gapplace.go`, `anchortext/anchortext.go`, `reportproj/render.go`,
`recordsql/views.go`, `record/estoppel.go` (`ProposalOver` by identity; the rule's statement gains
estoppel's exemption, F10), `cli/blue/edit.go` (the accept's exact span), `record/queries.go` (the
`ReportOp` doc), `record/batch.go` (`Batch.AppendBodies`, §III.9 item 6), `help/edit.md` and
`record/changesjson.go` (estoppel's exemption, §III.2), `record/reopened_test.go` (the three-argument
`failedChecksOf`, §III.3), `seatprobe/seatprobe.go` (§III.8), `cli/requiredmarker_test.go` and
`cli/enum_help_test.go` (§III.2); and `enums_test.go`, `evidenceview_test.go`, `seatprobe/boards.go` and
`cli/seat/seat.go` (`ReasonIs`, F11) move from KEEP to REWRITE.
`migrate/plankinds.go` is not changed (§III.9 item 4 puts `TranslateReadingBlockers` in `planids.go`).
`capture/capture.go` (GAP 2: the assembly screen's count through `FindsAgainst()`) and
`report/assemble.go` (F12: docket.md's source line) move from KEEP to REWRITE.
Dispositions (a file counted once, MODIFY counted with REWRITE): **DELETE 4 · REWRITE 89 · REGENERATE
17 · KEEP 120** (45 of the KEEP are historical).
Paths are relative to `plugins/frank-exchange-of-views/tools/internal/` unless shown. "n" = lines
carrying the FEOV sense, at `9ee7ad95e`.

**Production Go — record**

| File | n | Carries | Disposition |
|---|---|---|---|
| record/citationid.go | 52 | the fold, `contradiction`, `contradictionActs`, `ContradictionOfReading`, `requireOpenContradiction`, `unansweredContradictions`, `idWord`; `contradicts()` (l.704) | REWRITE: fold deleted, predicate → `gap_classes` facet |
| record/passblockers.go | 33 | `BlockerContradiction`, `raisedBy`/`raiser`, four constants, `contradictionFact`, PASS text, producer l.707-708 (l.883 ordinary) | REWRITE |
| record/mintbudget.go | 17 | own-raise exemption, `Answering`, `mintAnswerOutsideBudget` | REWRITE (§III.5) |
| record/dispatch.go | 14 | `contradictionReadiness`, lens arm, `GapIDs`/`Subject` desc (feeds enginegen) | REWRITE |
| record/onegapperclass.go | 10 | `answersRaisedContradictionAt`, exemption | REWRITE |
| record/evidenceview.go | 11 | `UnansweredContradictions`, `EvidenceContradictionJSON` | REWRITE |
| record/anchorchecks.go | 8 | comments; `contradicts()` at l.261 | REWRITE (§III.3) |
| record/sitting.go | 4 | lens owes the blocker; raiser-at-cap rule (l.160-174 ordinary) | REWRITE |
| record/about.go | 6 | about-kind validation, gap refusal, `aboutShapes`' comment (#1674) | REWRITE |
| record/record.go | 4 | finding → `requireOpenContradiction`; mint about-kind refusal | REWRITE |
| record/enums.go | 2 | finding's set "less contradiction" | REWRITE; [NEW] `ReadingClassField` (§III.2) |
| record/retirement.go | 2 | "a contradiction it read" as a debt | REWRITE |
| record/narrowedview.go, record/convergence_refusal.go | 1 each | comments | REWRITE |
| record/recordpb/record.proto | 6 | `ABOUT_KIND_CONTRADICTION`, comments l.136-137, 906, 1172-1173 (other hits: class names, `refutes`, `refuted`) | REWRITE (§III.1) |
| record/recordpb/record.pb.go, record.proto.sha256, schema_gen.go | gen | — | REGENERATE protogen / schemagen |
| record/recordsql/testdata/schema.sql | 1 | `enum_about_kind` row | REGENERATE (golden) |
| record/migrate/replay.go | 7 | `contradictionRef` | REWRITE (§III.9) |
| record/migrate/planids.go | 6 | `"contradiction"` id phrase (archived `why` prose), blocker subject | MODIFY (§III.9): `TranslatePlanIDs`'s contradiction-blocker branch (`:91-103`) DELETE, `TranslateReadingBlockers` NEW; the `"contradiction"` phrase in `planIDPhrases` KEEPs (it reads archived prose) |
| record/migrate/valuerenames.go | 0 | `AboutKind` respellings | MODIFY (§III.9.1): `"contradiction": "anchor"` row |
| record/migrate/remap.go | 0 | `idFields` (`:108-116`) | MODIFY (§III.1): `Verify.carried_by`, `Anchor.sentence_of` |
| record/migrate/gapplace.go | 0 | Anchor placement on replay (`:78-79`, `:104-115`) | MODIFY (§III.2a) |
| record/migrate/entries.go, migrate/outcomes.go | 3, 1 | comments | REWRITE |
| record/migrate/plankinds.go | 0 | `planKindRenames`, `TranslatePlanKinds` | KEEP: unchanged. `ReadPlan` decodes a blocker's kind as a string (`planstore.go:51-62`), so an archived `contradiction` blocker passes `TranslatePlanKinds` untouched to `TranslateReadingBlockers` (`planids.go`, §III.9 item 4) |
| record/batch.go | 0 | `Batch.Append` | MODIFY: [NEW] `Batch.AppendBodies` (§III.9 item 6) |
| record/migrate/gapclass.go, enumof.go, outcomeclosed.go, spotcheck.go, failedcheck.go (`answeredBy` = failed check), recordsql/views.go (`opened_by` = gap view) | 0 | other sense | KEEP (failedcheck.go is MODIFIED by §III.3 and recordsql/views.go by §III.2a, not for this word) |

**Production Go — cli**

| File | n | Carries | Disposition |
|---|---|---|---|
| cli/lens/verify.go | 16 | result `contradiction{id,opened,opened_by}`, opened/joined text | REWRITE (§III.2) |
| cli/lens/finding.go | 2 | `--about-kind`/`--about` help | REWRITE |
| cli/chair/dispatch.go | 1 | comment | REWRITE |
| cli/seat/verbs.go | 3 | `unanswered_contradictions` manual text, budget text | REWRITE |
| cli/setup.go | 1 | `--mint-budget` help | REWRITE |
| cli/seat/help/corroborate.md, verify.md, dispatch.md, mint.md, finding.md | 2, 2, 3, 2, 1 | contradiction prose (finding.md l.9 "internal contradiction" stays) | REWRITE; `corroborate.md` DELETE (F13) <!-- CHANGED F13 --> |
| cli/lens/mint.go | 0 | ordinary English | MODIFY (shared flag helper, §III.2; `verifiedFix` places a `verify`'s gap with `AttachToSentenceOf` and names the verb's subject in its refusal, §III.2a) |
| cli/seat/verbs.go (placement) | 0 | `Places`, `AppendPlaced` (`:1384-1417`) | MODIFY (`PlacesToSentenceOf`, §III.2a) |
| cli/seat/seat.go | 0 | `ReasonIs` (`:1082`), ordinary at the census hit | MODIFY (the `verify` row, F11; no `corroborate` row, F13) <!-- CHANGED F13 --> |
| cli/seat/help/record.md, cli/bench/outcome.go | 0 | inquiry `refuted` / ordinary | KEEP |

**Placement and the accept by identity (§III.2a; none carries the word):** `anchortext/anchortext.go`
(`AttachToSentenceOf`, two sentinels) NEW; `reportproj/render.go` (`sentenceOfMut`, `mutationOf`)
MODIFY; `record/recordsql/views.go` (`report_op`'s `'sentence_of'` arm and comment) MODIFY;
`record/queries.go:251-256` (`ReportOp` doc) MODIFY; `record/recordsql/testdata/schema.sql`
REGENERATE; `record/record.go:664-668, :941-951` MODIFY; `record/recordpb/record.proto:1179-1187`
MODIFY; `record/onegapperclass.go:82-92` MODIFY; `record/estoppel.go:175-211` (`ProposalOver` by the
gap's token) MODIFY; `cli/blue/edit.go:335-349` (the accept's exact span) MODIFY.

**Other Go** — `terms/terms.json` (5: the entry, its ordinary-English collision, `lens retirement`): REWRITE
then vocabdoc. KEEP (0 FEOV-sense hits; source-outcome, class or ordinary sense): capture/computationrerun.go,
report/registrations.go, verify/verify.go, seatprobe/production.go, bluedoc/bluedoc.go, flags/names.go, feov/error.go, seatenv/identity.go,
seatenv/seatenv.go, tessocr/engine_stub.go, `scripts/decisions/main.go`,
`scripts/golden/review.go`, gray-area `cmd/gray-area-capture/main.go`, `internal/claims/coverage.go`,
`internal/claims/manifest.go`. REWRITE beyond the word: `seatprobe/seatprobe.go`, `seatprobe/boards.go`
(§III.8; `boards.go`'s census hit, `cross_section_contradiction`, stays); `capture/capture.go`
(`sources_refuted` and its found-against check keep their sense; `:483` calls `FindsAgainst()`, GAP 2,
§III.2b); `report/assemble.go` (its census hits keep their sense; docket.md gains F12's source line,
§III.2c).

**Tests**

- DELETE (4): record/contradictionref_test.go, record/contradictionquadratic_test.go,
  record/answerbudget_test.go (its one test is the own-raise exemption),
  `tools/integration/surface/contradictionanswer_test.go`.
- REWRITE (22): record/redcitation_test.go (3 of 7 tests; F13 replaces the whole file by `foundsource_test.go`) <!-- CHANGED F13 -->, blockerowner_test.go, passblockers_test.go
  (incl. the `contradictingVerify` helper), envelope_test.go, narrowedview_test.go, reopenedm19_test.go,
  planidphrases_test.go, export_test.go (`ContradictionReadiness` export), satthrough_test.go,
  lensowes_test.go, migrate/reproduceanchor_test.go, migrate/planids_internal_test.go;
  cli/verifyclaim_test.go, corroborate_test.go (DELETED by F13, below) <!-- CHANGED F13 -->, m23contradiction_test.go (the quote-refusal half stays),
  onegapperclass_test.go, mintabout_test.go, deadpath_test.go, badvalue_test.go, pagestates_test.go;
  `releasegate/fuzz/fuzz_test.go`, `sweeptally_test.go`.
- REWRITE beyond the word: cli/requiredmarker_test.go (`requiredProbes`' two reading rows, §III.2),
  cli/enum_help_test.go (`enforcedElsewhere`'s `verify --class`, §III.2), record/export_test.go
  (`BlockerWhy` per state, §III.9 item 4), record/reopened_test.go (`failedChecksOf`'s five calls, `:205,
  :210, :251, :262, :263`, take the open set, §III.3), record/enums_test.go (`narrowedBy`'s `gap_may_name` row, deleted with the
  facet), record/evidenceview_test.go:250 (the "FINDING HALF" comment, phrase sweep),
  record/unreadcitation_test.go:119 and record/unreadb6_test.go:21-22, :66 (the "does not close"
  comments, F8), cli/keyedretry_test.go (the retry test, §III.2).
- KEEP (other sense): record foundagainstm30_test (`sources_refuted`), cast_test,
  passagreement_test, registeronce_test, roster_test, spotcheck_test, viewjson_test (#1678, ordinary),
  recordpb/requiredfields_test, migrate/gapclass_test; cli readrunrefusal_test, agentbinding_test,
  identitywalk_test, counterparty_test, blueedit_test, verbguardrace_test, fetchsummary_test,
  fetchocr_test, verify_test, dashboard_serve_test, seat/seat_test; integration/surface/
  constitutiondirective_test; report assemble_test, conduct_test, registrations_test,
  spotcheckreaders_test, verdictbesideprose_test, verification_test; seatprobe_test,
  surfacecoverage_test, seatenv_test, bluedoc/proposal_test, capture_test, verify/verify_test,
  scripts/golden/review_test; gray-area main_test, coverage_test.

**Release fuzz** (`releasegate/fuzz/fuzz_test.go`): drive `answerContradictions`, `raisePrinted`,
`ownUnansweredRaise`, `raiseAndMint`, `contradictionAbout`; oracles `quoteRaisesNothing`,
`contradictionUnraised`; parser `contradictionOf`; call sites l.5278, 5400-5412 (`contradictionOf` at 5407), 5480-5501, 5575, 8873;
fields `ownAnswers`, `answerBudgetErr` (l.499-504, 3599-3640, 5732-5733, 6346-6347, 7075). `sweeptally_test.go`: `ownAnswers`
tally (l.534, 659, 943-947), gate wording l.913. All REWRITE (§III.11). PR #1642: superseded.

**Goldens / testdata** — REGENERATE: difftest/testdata/manual_runs_every_help_page_live.golden (8 lines),
`tests/simulator/testdata/prompt-red-chair.golden`. REWRITE: `tests/simulator/testdata/o3-chair5-relays.json`
(a captured plan with blocker kind `contradiction`; the schema test fails once enginegen drops the value).
KEEP: difftest error_catalogue.golden, migrate/testdata/prose-rewrites.golden (archived prose),
prompt-red-lens-logic/dark-side goldens, debate.test.mjs, prompts.test.mjs. gap_class_enum.golden
REGENERATES for the new class only.

**debate.js** (`skills/research-protocol/scripts/debate.js`): l.394 and l.415 sit inside the
`GENERATED ENVELOPE VOCABULARY` region (l.373-451) → REGENERATE (enginegen); l.915 (red-chair prompt,
rewritten by #1677, still ending "the lens that read a contradicting source no finding raises") is
authored → REWRITE; l.227, 253, 255 KEEP. #1677's `record_read` (`RECORD_READ`, l.448, 520, 561, 630;
`readOf`, l.687) carries no blocker, gap or reading: KEEP.

**Agents and skills** — REGENERATE (agentgen): all 12 `agents/*.md` (7 red-lens-*, red-chair, blue-lane,
blue-researcher, blue-synthesizer, lead-judge). REWRITE: `skills/research/SKILL.md` l.7,
`skills/research-protocol/references/run-directory.md` l.76. The two generated preloaded skills have no hits.

**agentgen sources** (`scripts/agentgen/src/`) — REWRITE: fragments/lens-audit-and-mint.md,
fragments/lens-duties.md, red-chair.md. KEEP: lead-judge.md, fragments/blue-researcher-frame.md,
fragments/blue-research-protocol.md (ordinary English).

**The `lane` role (#1663).** Is it a carrier of the concept? Swept at `9ee7ad95e`: `seatclass/roster.go:99`,
`cli/blue/command.go:40-47` (`LaneVerbs`), `record/envelope.go:111`, `agents/blue-lane.md` as rendered,
`scripts/agentgen/src/blue-lane.md` and its fragments.

| Site | Reads or teaches | Disposition |
|---|---|---|
| The lane's verbs (`LaneVerbs`: register, ingest, inquiry, log) | no `verify`, `mint`, `close`, `edit` or `retire` (F13: `corroborate` is deleted) <!-- CHANGED F13 -->: a lane performs no reading, mints no gap and is refused as the wrong seat on each | KEEP |
| Its work list (`WorkArrivesAsItems`, `envelope.go:111`) and the plan | work reaches it by dispatch, never as items, so it holds no failed-check item; a found-against check readies `blue-respond` alone (`dispatch.go`'s arm, §III.3) | KEEP |
| `cli/lens`' `seat.Places` guards; `PartStanding` (`queries.go:381`) | the first non-append row closes the door on a lane still writing, so a reading's Anchor, like a plain mint's, closes it. No lens sits before a report exists | KEEP |
| `scripts/agentgen/src/blue-lane.md` | no reading, failed-check or contradiction sentence. It includes neither `blue-research-protocol.md` nor `blue-search-and-cite.md`, the fragment holding blue's failed-check duty (§III.7) | KEEP |
| `fragments/blue-researcher-frame.md:40`, `fragments/blue-search.md` | "contradicted" (ordinary English); no reading | KEEP |
| `agents/blue-lane.md` | inlines the `manual`: `show evidence`'s `unanswered_contradictions` paragraphs (`:552`, `:562`), the `failed check` and `lens retirement` terms (`:171`, §III.7) | REGENERATE |
| `plugins/frank-exchange-of-views/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden`, the lane's page | the same | REGENERATE |
| `seatprobe.Roles` (`seatprobe.go:122`) | a lane's events are its own: no Verify | KEEP |
| The fuzz's lane drive (`releasegate/fuzz/parts_test.go`) | no reading | KEEP |

The phrase sweep and the word census hold no other hit under the role.

**Docs** — REGENERATE: `docs/vocabulary.md` (vocabdoc). REWRITE: `docs/citations-and-bibliography.md` l.192,
l.178 (every verdict, `verify`'s included, is stamped with the copy's completeness, GAP 1)
and l.158-167 (the failed-check bullet: two states, a gap that "cannot close", and verify without the gap
act → three states, the gap minted at the reading, the close refused while the check stands and a
closing ruling closing the gap while the check holds the PASS); `docs/seat-command-triggers.md` l.56 (the
`lens verify` row: a reading of the work that finds against mints its gap in the same act, with the mint's judgment flags, its
`--reason` the gap's argument, at the
sentence the citation's token sits in; a reading that reads nothing — unreachable, or a not_the_work
copy — mints nothing, and the unread citation's lens owes the gap about it), l.66-67 (the `lens mint` row: the owed failed-check gap is the unread or unsound
check's; the `lens close` row: a gap carrying a found-against check is refused too, until the reader's
own backing reading, an edit or a retire), l.86 (the `blue retire` row: an anchor found against is
blue's to edit, retire or replace).
KEEP: propagation-and-anchoring.md, seat-surface-naming.md, why-a-seat-stops.md; outside FEOV,
`plugins/gray-area/README.md`, `plugins/gray-area/skills/audit-seat-coverage/SKILL.md`,
`plugins/prosthetic-conscience/skills/resume/SKILL.md` (ordinary English).

**Phrase-sweep additions** (old model, no word; §III.7): `scripts/agentgen/src/skills/research-protocol/SKILL.md:17-19`
→ REWRITE (regenerates `skills/research-protocol/SKILL.md`); `cli/lens/verify.go:304, :702-705` → REWRITE.
`scripts/agentgen/src/lead-judge.md` → REWRITE (one added sentence, R4).

**Plans / ideas / law / feov-memory / research** — KEEP-historical (45 files: 33 under `plans/`, 7
under `research/`, 2 under `ideas/`, 2 under `feov-memory/`, 1 under `law/`). FEOV-sense mentions in
plans/reopened-evidence-is-owed.md, feov-lens-bar.md, gap-class-enum.md, pre-dry-run-batch.md,
one-item-per-dispatch.md:336, ocr-cite-derives-page.md:574, historical/red-citations.md,
historical/migrate-staged-publish.md:284, law/precedents.md:156 record past decisions; none is a live
instruction, none is rewritten.

**What #1677 and #1678 moved, and the dispositions that could have rested on it** (re-derivation at
`9ee7ad95e`, `git diff 0c57758e8 9ee7ad95e`, every file this plan cites):

| Change | Where this plan reads it | Disposition holds? |
|---|---|---|
| `Party.Occasions`, `Dispatch.occasions`, `benchRowSitting`'s per-occasion arm deleted (#1677); `Dispatch` field 4 reserved | §III.3 dispatch arms (`dispatch.go:546-561`), `failedCheckReadiness` (`:1290-1308`), `PlanBlocker` (`:29-33`), `contradictionReadiness` (`:1281-1288`) | Yes: no arm of this plan names an occasion; the lines moved up by the deletion and the map carried them. The bench is still readied on a docketed gap by the plan (`engage(benchSeat, g.id)`) |
| The bench hears a petition or position at any barrier where no seat runs, with no chair sitting (#1677) | `TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld` (the chair dockets G-1, the bench rules); F8 | Yes: a docket is still a plan party and a docket sitting (`benchRowSitting` reads `OCCASION_DOCKET`); a hearing answers no row and rules on no gap |
| `debate.js` learns the halt and the outcome only from `record_read`, which a PreToolUse hook writes on every envelope (#1677) | `debate.js:1087-1094` (`anchorsStanding`), §III.3 KEEP | Yes: the plan is still the chair's envelope (`const plan = chairEnv`, `:953-957`, validated by `requirePlan`); #1672, open, would move it to the record and is not derived against |
| `record_read`, `readOf` (new) | — | Carries no blocker, gap or reading: KEEP, outside the census |
| `TranslatePlanParties` runs before `TranslatePlanKinds` in replay (#1677, `replay.go:373-390`) | §III.9 item 4 (`TranslateReadingBlockers` after `TranslatePlanKinds`, before `TranslatePlanIDs`) | Yes: the order becomes Parties → Kinds → ReadingBlockers → IDs; the bench-row occasion fill-in (`:340-347` at the old base) is gone, which no step of §III.9 item 6's table named |
| Client floor 2.1.305 (#1677) | §V.8 | Stated there |
| `Verify.sha256 = 20` (#1678), in `remap.go`'s string list and on the evidence row | §III.1 field numbers; §III.2c | `carried_by = 21`, `carrier_minted = 22`; F13 deletes `sha256 = 20` with the footnote, reserved <!-- CHANGED F13 --> |
| `EstoppedJSON` names no closer (#1678 D17) | §III.2's estoppel sweep (94 lines) | Another sense (the `estopped` register): KEEP |
| `ResolveAbout` checks a reference's shape first (`aboutShapes`, #1674) | §III.6 `about.go` | Its comment names the contradiction (REWRITE); the map holds no contradiction row; migrate's raise test is unaffected (replay does not call `ResolveAbout`) |
| `verifiedFix`'s dropped-anchor clause names `--quote` (#1663 smoke D7, `mint.go:364-372`) | §III.2a `verifiedFix` | Takes the verb's subject, as the rest of the refusal does |

**Readers of a refutation that do not name the contradiction** (added at the Slim-fast peer's request;
each states how it reads once a refutation is a gap):

| Site | Today | After | Disposition |
|---|---|---|---|
| Anchor ledger `anchorStanding.state()` (`anchorchecks.go:152-178`, from #1601 / 90f3fa008) and `stateFoundAgainst` | a citation whose standing check refutes or is absent; its PASS hold was "the contradiction's" | Unchanged predicate: found-against is still the **citation's standing** (a standing negative reading of the sentence the report carries), NOT "has an open gap of the new class". The gap is a separate fact linked by `Verify.carried_by`; the state is what makes the gap unclosable (§III.3). Keying the state on the gap would let a gap closed into a successor, or an archived reading with no gap, read as holding | MODIFY (comment only; `contradicts()` → `gap_classes` facet) |
| `failed_check` PASS blocker (`passblockers.go:403`, `failedcheck.go`) | unread citations and unsound computations | gains the found-against state (§III.3): holds the PASS until its reader's own later reading backs the sentence, blue edits it, or blue retires the citation (another lens's backing reading clears nothing); readies blue alone; carrier = `Verify.carried_by` followed through its successors | MODIFY |
| `unchecked` PASS blocker | a citation or computation with no check of its current sentence | unchanged: an edit of a found-against sentence makes the citation unchecked, owed a fresh check by its owning lens; that check either holds (failed check cleared, the gap may close) or is negative again (names the standing gap, or mints one if it is closed) | KEEP |
| Evidence lens's gap duties (`fragments/lens-duties.md`, `lens-audit-and-mint.md`, `verify.md`) | read every citation; a negative reading opens a contradiction it must raise by finding, then answer by a gap minted from it, outside its budget | reads every citation; a reading of the work that finds the source against the sentence carries the gap's judgment (class among the outcome's admitted set, problem, check, grades) and its argument (`--reason`) and mints it in the same act, outside its budget, at the sentence its citation's token sits in; a reading that reads nothing (unreachable, a not_the_work copy) mints nothing; a second reading of a sentence holding that gap names it; the lens closes the gap with its ordinary close once its own later reading backs the sentence, or blue edits or retires it. Unread readings keep the owed-gap-about-the-anchor duty | REWRITE (agentgen) |
| Failed-check owed-ness readers (§III.3): dispatch lens arm `dispatch.go:546-553`, lens work item `sitting.go:401-408` (and `EnvelopeRefusal`, `envelope.go:44`, through it), retirement `retirement.go:164-172`, budget `failedcheck.go:109-119, :208-215` | owed iff `answeredBy == ""` | owed iff state is `unread`/`unsound` and no gap carries it; `found_against` never owed by its reader | MODIFY |
| Failed-check printed sentences (§III.3): `passblockers.go` `failedFact` :226, `failedClears` :234, `failedRemedy` :245, `failedOwed` :254, `blueFailedItem` :262, the kind's gloss/refusal/why :404-433; `dispatch.go:1293` `failedCheckReadiness`; `failedcheck.go` repeat and close refusals :135, :181 | branch on `Detail` (a computation's `bundle_sha`) and `answeredBy`; citation = unread | branch on the state; found-against wording names the reading and blue's remedy (edit, retire or replace), never an owed mint | MODIFY |
| Failed-check prose: `help/mint.md:15`, `help/close.md:11`, `help/verify.md:11`, `help/reproduce.md:21`, `fragments/blue-search-and-cite.md:22-29`, `fragments/lens-audit-and-mint.md:6`, `terms.json` `failed check`, `docs/citations-and-bibliography.md:158-167`, `docs/seat-command-triggers.md:56, :66-67, :86` | two states; the seat whose check failed owes the gap; "the gap does not close while the check stands" | three states; the owed gap is the unread/unsound check's; a found-against check's gap is minted at the reading, and the check clears by the ledger's rule; the seat's close is refused while the check stands, and a closing bench ruling closes the gap while the check holds the PASS (F8) | REWRITE |
| Fuzz failed-check drives and tallies `fuzz_test.go:1229, :1238, :3026, :3036, :3058, :3061, :3115-3124, :3434, :3735, :4208, :4253, :4285`; gates `sweeptally_test.go:843-854` | `Detail == ""` = unread citation; `AnsweredBy() == ""` = owed; unread tallies count every citation check | key on `CheckState()`; unread tallies count `unread` only; found-against has its own tallies and gates | REWRITE (§III.11) |
| `verify.go:530-542, :632-657` `failedCheckWords` and the JSON `failed_check` | guarded by the NEW reading's `!VerifyReads`, worded as unread whatever the check's state; JSON carries no state | after every verification of a citation, every standing failed check of the anchor, worded by its state; JSON `failed_checks[].state` | REWRITE (§III.2) |
| `verdict.go:141-162`; `debate.js:1087-1094`; `fuzz_test.go:2972, :6990` | anchor names only; owner-or-blue readied; no VERIFIED over a failed check | unchanged, true of every state | KEEP |
| Tests pinning unread/unsound sentences: `unreadcitation_test`, `unreadb6_test`, `blueanswered_test`, `failedcheck_test`, `reopened_test`, `tiersubstitution_test`, `unsoundm32_test`, `capture/dispatchparity_test` | those states' wording | their owed-mint and remedy wording does not move — they are the regression net for it; the "does not close" comments in `unreadcitation_test.go:119` and `unreadb6_test.go:21-22, :66` are rewritten (F8) | KEEP (two comment REWRITEs) |

**Readers of a failed check's owed-ness and its printed sentence** (the class the round-2 audit named;
run at `0c57758e8` and carried by the line map; the wider `new-fc.txt`, re-run at `9ee7ad95e`, adds only capture audits' `.Detail` (another sense), every hit read, `new-fc3.txt`; dispositions in the table above):

```
grep -rIn -E 'answeredBy|answered_by|AnsweredBy' plugins/frank-exchange-of-views/tools/internal plugins/frank-exchange-of-views/tools/cmd plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js scripts/agentgen/src | grep -v _test.go
grep -rIn -E 'BlockerFailedCheck|failedChecksOf|answersOwnFailedCheck|ownFailedCheckAnswers|failedRemedy|failedOwed|blueFailedItem|failedCheckReadiness|failedCheckWords|failed check' <same paths> | grep -v _test.go
grep -rIn -i -E 'check (of yours|that) failed|whose check failed|failed check|about the anchor|about that anchor' scripts/agentgen/src plugins/frank-exchange-of-views/tools/internal/cli/seat/help plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js
```

Hits outside the contradiction fold (which is deleted: `citationid.go:554-685`, `onegapperclass.go:145`,
`mintbudget.go:362-380`): `passblockers.go:84-98, :223-274, :403-433, :714`; `failedcheck.go:44-119,
:134-215`; `sitting.go:321-322, :401-408`; `retirement.go:142-172`; `dispatch.go:22, :546-561,
:1290-1308`; `mintbudget.go:308-309, :370, :394, :400-410`; `verdict.go:141-162`; `cli/lens/verify.go:530-542,
:632-657`; `cli/seat/verbs.go:308`; `terms.json:1079-1080`; `help/mint.md:15`, `help/close.md:11`,
`help/verify.md:11` (unread; its "does not close" clause REWRITEs, F8), `help/dispatch.md:19`; `fragments/lens-audit-and-mint.md:6`,
`fragments/blue-search-and-cite.md:22-29`; `debate.js:394` (enginegen, from `dispatch.go:22`),
`debate.js:1087-1094`; `onegapperclass.go:28` (a comment naming `answersOwnFailedCheck`, KEEP);
`migrate/plankinds.go:19` (respelling, KEEP); test files under `record/`, `capture/` and the fuzz (rows
above). `sitting.go:268, :676-678` (`answeredByActThisSitting`) is a motion reader, other sense, KEEP.

**Readers of the deleted `gap_may_name` facet** (`grep -rn -i 'gap_may_name\|GapMayName'` over the tools
tree, every hit read): `recordpb/record.proto:134-138, :1140-1144` (declaration and values) [DELETE];
`recordpb/facets.go:35` (`"gap_may_name": E_GapMayName`) [DELETE]; `record/enums.go:347-363`
(`gapAboutKinds`) [DELETE] and its two callers, `enums.go:300` (mint's `about_kind` values → the full
`AboutKind` set, as finding's) and `record.go:970` (the mint's about-kind refusal, §III.6) [DELETE];
`record/enums_test.go:339-341` (the `{"mint","about_kind"}` row of `narrowedBy`) [REWRITE: the row is
deleted, since mint's set no longer narrows the enum]; `cli/mintabout_test.go:73` (a comment naming the
facet; the file is a census REWRITE row already) [REWRITE]; `record.pb.go` [REGENERATE].

**Readers of the new Verify fields** (`carried_by`, `carrier_minted`): by name — `migrate/remap.go:108-116`
`idFields` (`Verify.carried_by` listed there; `carrier_minted` is a bool, which the three string lists do
not class — `TestEveryStringFieldIsInOneList` fails `carried_by` left out), `record.go:483`
`keyFields` (neither name is on it; <!-- CHANGED F13b --> after F13b a citation reading keys on `anchor` and a
found-source reading on the ordinal, its identity checked by `requireNotARetriedReading`; the reason the
field is not called `gap_id`, §III.1), `flags/names.go:402` `payloadFlag` (falls back to `--carried-by`,
never printed: both fields are tool-supplied, `seat.Supplies` in `verify.go`), and the
required-field sweep (`recordpb/requiredfields_test.go`, both optional); by meaning — `failedChecksOf`
(carrier), `requireMintWithinBudget`/`lensMintStanding` (`carrier_minted`, §III.5), the repeat rule
(§III.2), the reading's result (`gap.minted` read back), migrate (archived readings carry neither,
§III.9.2), the sweep tally `open_by_found_by` (§III.11), §V.8's query, and F12's two printers —
`EvidenceVerificationJSON.CarriedBy` and docket.md's source line (§III.2c) — and, after this plan,
#1657 Part 1a′'s widened L1 and re-read arm (§III.16), which read both fields by SQL. `references: "mint.gap_id"`
makes the column a foreign key; replay translates no gap id (G- ids are carried unchanged), so migrate
needs no translation for it beyond `idFields`' respelling of an archived id, which a record written at
this epoch never holds.

**Readers of `Anchor.sentence_of`** (by name and by meaning; the rows of §III.2a's table):
`record.go:941-951` (validation), `record.AppendPlaced` and the reading's placed append (writers),
`seat.PlacesToSentenceOf` (guard), `recordsql/views.go` `report_op` (`'sentence_of'` arm),
`reportproj` `sentenceOfMut` (replay), `onegapperclass.go` (placement for the sentence),
`migrate/gapplace.go` (replay check), `migrate/replay.go` (the reading act re-driven whole, §III.9 item
6), `migrate/remap.go` `idFields`, `cli/lens/mint.go` `verifiedFix` (a reading's `--new`). The SQL column
is new, so every SQL reader of `anchor` was read: `report_op` and `id.go:90`'s `NOT EXISTS (… "anchor"
WHERE "id" = ?1)` (id only, KEEP) are the only two.

**Readers of the reading's Mint's fields** are §III.2's field table (`new-mintfields.txt`).

**Carried by the concept, not found by the word:** `record.proto.sha256`, `requirements.json` (epoch),
`schema_gen.go`, `record/batch.go` (§III.9 item 6), `failedcheck.go` (§III.3), `cli/lens/mint.go` (flag helper),
`recordpb/facets.go` and `enums_test.go` (the facet), `help/reproduce.md:21` and
`docs/seat-command-triggers.md` (the failed-check prose, F8), `cli/keyedretry_test.go` (§III.2's retry
test), `unreadcitation_test.go` (the F8 bench test lands beside its unread mirror).

**F13 census — `corroborate`, the red footnote, and the label and bibliography paths** <!-- CHANGED F13 -->
(run at `6102f2b09`; raw hits in `~/.claude/scratch/plan1647-f13/`: `census-corr.txt`, per-file counts
`census-corr-files.txt`, identifier sweep `sweep-label.txt`, phrase sweep `sweep-phrase.txt`):

```
git grep -I -n -i corroborat HEAD -- . ':!run-archive'
git grep -n -E 'Corroborated\b|CorroboratedBy|corroborated_by|ExistingCorroboration|refuseCorroborationTitleVoice|"independent"|\.Independent\b|backsTheClaim|adjudates|mayCite|GetWorkStatus\(\)|GetSha256\(\)' HEAD -- plugins/frank-exchange-of-views scripts docs ':!*.pb.go'
git grep -n -i -E 'source (it|you|red|the lens) (went and )?f(ou|i)nd|found (it )?(itself|yourself)|source of (its|your) own|red.s (own )?(source|footnote|citation)|red-placed|placed by (red|a lens)' HEAD -- plugins/frank-exchange-of-views/{tools,agents,skills} scripts/agentgen docs ':!*.golden'
grep -n 'GetLabel()' over the tools tree, each hit read for a Verify receiver
```

Result: the word census **1,219 hits in 144 files** — this plan 137 (rewritten by this revision), 25
historical files under `plans/`, `research/`, `feov-memory/` (388 hits, KEEP-historical), and **118 live
files, 831 hits**; the identifier sweep **100 hits in 37 files** (<!-- CHANGED F13b --> re-run at `6102f2b09`
by the targeted round and here: 100 hits, 37 files; the F13 text said 38) and the phrase sweep **34 hits in 20
files**, every file of both already in the word census but `cli/blue/cite.go` and `cli/cite_test.go`
(`GetWorkStatus()`/`GetSha256()` of a `Cite`: KEEP) and `capture/capture_test.go` (a fixture's "source
red found" prose: REWRITE with the fixture). Every hit read; dispositions per file (n = word hits):

| Disposition | Files (n) |
|---|---|
| DELETE | `cli/seat/help/corroborate.md` (2); `cli/corroborate_test.go` (52; its stamp rows move, §III.10); `cli/corroborationsha_test.go` (23); `record/corroborationsource_test.go` (6); `record/redcitation_test.go` (41; replaced by `foundsource_test.go`) |
| REWRITE — production Go, record | `record/evidenceview.go` (24: `CorroboratedBy`, `Label`, the labelled source row, `independent`'s comments), `record/citationid.go` (17: `Corroborated`, `citedSource`'s Verify arm, `ExistingCorroboration`), `record/recordpb/record.proto` (16: fields 11, 15, 20 reserved, comments; the facet), `record/redvoice.go` (6), `record/reading.go` (5, phase 1's), `record/anchorchecks.go` (3), `record/supersede.go` (3), `record/refs.go` (3; <!-- CHANGED F13b --> and, outside the word census, its two SQL readers of `verify.label`, `:130` `requireAboutAnchor` and `:205` `HoldsAnchor`, §III.2a), `record/id.go` (<!-- CHANGED F13b --> outside the word census: `:88`, `unplacedLocationAt`'s `verify` arm), `consistency/consistency.go` (<!-- CHANGED F13b --> outside the word census: `:272`, a `Verify.GetLabel()`), `record/migrate/proserefs_test.go` (<!-- CHANGED F13b --> `:112-113`, a `Verify.GetLabel()`), `record/sitting.go` (2, comments), `record/record.go` (2: `:1751` comment, `:1785` call DELETE; `keyFields`), `record/failedcheck.go` (1, phase 1's), `record/viewjson.go` (1), `record/recordsql/views.go` (1, comment), `record/migrate/gapplace.go` (2: the Verify arm DELETE), `record/migrate/eraentries.go` (1: the archived `independent` reason, reworded to "a reading of a source the lens found names no anchor") |
| REWRITE — production Go, cli and others | `cli/lens/verify.go` (38), `cli/lens/command.go` (1, DELETE the registration), `cli/root.go` (1), `cli/seat/seat.go` (3), `cli/seat/verbs.go` (2), `cli/seat/source.go` (2), `cli/seat/answer.go` (1), `cli/blue/retire.go` (1, comment), `cli/seat/help/verify.md` (1), `cli/seat/help/finding.md` (2), `report/assemble.go` (7), `report/citations.go` (3), `report/layout.go` (2), `reportproj/render.go` (1, comment), `anchor/anchor.go` (1, comment), `flags/names.go` (1, comment), `seatprobe/boards.go` (1), `seatprobe/grant.go` (1, comment), `devcmd/seatprobe/main.go` (1, comment), `reportvoice/tells.go` (1, comment), `nonet/nonet.go` (1, comment naming the fixture) |
| REWRITE — tests | `cli/redvoice_test.go` (26: the title-voice cases DELETE), `cli/retryvalidates_test.go` (21), `cli/answers_test.go` (18), `cli/board_test.go` (14), `cli/verifycopy_test.go` (13), `cli/verbs_test.go` (9), `cli/gapanchor_test.go` (7), `cli/m23contradiction_test.go` (6), `cli/verifyconfidence_test.go` (4), `cli/deadpath_test.go` (4), `cli/bridgekind_test.go` (4), `cli/reopen_test.go` (3), `cli/unreadablerefuses_test.go` (2), `cli/requiredmarker_test.go` (2), `cli/vocabulary_test.go`, `cli/siblingverbs_test.go`, `cli/retiremidsentence_test.go` (1 each), `record/migrate/cutfields_test.go` (9), `record/migrate/gapplace_test.go` (8), `record/migrate/archive_test.go`, `record/migrate/entries_test.go` (sweep), `record/evidenceview_test.go` (1 + sweep), `record/idnamespace_test.go` (1), `record/recordpb/keycensus_test.go` (1), `report/assemble_cite_test.go` (5), `report/assemble_pages_test.go` (4), `report/showsthework_test.go` (3), `report/riskvoice_test.go` (2), `report/assemble_dedup_test.go` (1), `reportproj/render_insert_test.go` (2), `seatprobe/fidelity_test.go` (3), `diagnostics/seen_test.go` (5), `difftest/contract_test.go` (1), `consistency/consistency_test.go` (1), `integration/surface/promptverbs_test.go` (1), `capture/capture_test.go` (sweep), `releasegate/fuzz/fuzz_test.go` (21), `acts_test.go` (7), `sweeptally_test.go` (5), `copy_test.go` (2) |
| REWRITE — agent text and docs | `scripts/agentgen/src/red-lens-evidence.md` (1), `fragments/lens-duties.md`, `fragments/leaf-node-read.md`, `fragments/blue-search.md` (1 each), `fragments/lens-audit-and-mint.md` (3), `src/skills/research-protocol/SKILL.md` (1: "The bibliography is BOTH sides'"), `skills/research-protocol/references/report_template.md` (1), `docs/citations-and-bibliography.md` (6), `docs/seat-command-triggers.md` (3: the `lens corroborate` row DELETE), `docs/seat-surface-naming.md` (7: the verb lists; the dated measurement prose stays), `docs/why-a-seat-stops.md` (3), `tests/simulator/debate.test.mjs` (1, the footnote-title assertion DELETE) |
| REGENERATE | `record/recordpb/record.pb.go` (16), `recordsql/testdata/schema.sql` (4), the 12 `agents/*.md` (24, 23 ×6, 6 ×3, 4 ×2), `skills/research-protocol/SKILL.md` (1), `difftest/testdata/manual_runs_every_help_page_live.golden` (12), `error_catalogue.golden` (4), `help_contracts.golden`, `role_boundaries_and_help_contracts.golden`, `projections_debate_changelog_citations.golden`, `full_round_integration.golden` (2 each) |
| KEEP | `record/migrate/testdata/prose-rewrites.golden` (57, archived prose); `anchor/sentences_test.go` (3, subject text of archived reports); `fetchcache/pageread.go` (3, ordinary English); `scripts/decisions/main.go` (1, a dated decision note); `README.md`, `plugins/frank-exchange-of-views/README.md` (1 each, "corroboration confidence", ordinary English for the trust grade); `cli/blue/cite.go`, `cli/cite_test.go` (sweep, a `Cite`'s fields) |

Phrase sweep, beyond the word: `skills/research-protocol/scripts/debate.js:879` (1, the authored
`ledgerClause`: "A SOURCE YOU FOUND YOURSELF IS A DIFFERENT ACT — blue never cited it …" — REWRITE: a
source you found yourself is read against the sentence and, where it backs an uncited sentence or
refutes one, records the gap blue answers with a citation; the prompt names the act, never a verb or
flag), `record.pb.go` (REGENERATE), the rest
the files above.

### III.13 Directory tree of touched paths

```
plugins/frank-exchange-of-views/
  requirements.json                         [MODIFY] eventSchema 55
  tools/internal/record/
    recordpb/record.proto                   [MODIFY] AboutKind, GapClass, SourceOutcome facet, Verify.carried_by, Verify.carrier_minted, Anchor.sentence_of
    recordpb/facets.go                      [MODIFY] gap_may_name deleted
    recordpb/record.pb.go, schema_gen.go    [REGENERATE] protogen, schemagen
    recordsql/views.go                      [MODIFY] report_op 'sentence_of' arm
    recordsql/testdata/schema.sql           [REGENERATE]
    citationid.go                           [MODIFY] contradiction fold deleted
    successor.go                            [READ] successorsOf, reused by failedChecksOf
    anchorchecks.go, failedcheck.go         [MODIFY] found-against is a failed check
    onegapperclass.go, mintbudget.go        [MODIFY]
    passblockers.go, sitting.go, dispatch.go, retirement.go, evidenceview.go,
    narrowedview.go, about.go, record.go, enums.go, convergence_refusal.go   [MODIFY] incl. ReadingClassField
    batch.go                                [MODIFY] Batch.AppendBodies (a reading act on replay)
    changesjson.go                          [MODIFY] comment: estoppel's exemption
    migrate/replay.go, planids.go, entries.go, outcomes.go                   [MODIFY]
    migrate/valuerenames.go, remap.go, gapplace.go                          [MODIFY]
    estoppel.go, queries.go                 [MODIFY] ProposalOver by the gap's token; ReportOp doc
    evidenceview.go                         [MODIFY] FindsAgainst() (GAP 2); carried_by on the row (F12)
    migrate/*_test.go                       [NEW] m30, m23, a record at this epoch, sentence_of check
    *_test.go                               [NEW] m32 cut, migrated found-against checks
  tools/internal/anchortext/anchortext.go   [MODIFY] AttachToSentenceOf
  tools/internal/reportproj/render.go       [MODIFY] sentenceOfMut
  tools/internal/report/assemble.go         [MODIFY] docket.md's source line (F12)
  tools/internal/capture/capture.go         [MODIFY] the assembly screen counts through FindsAgainst()
  tools/internal/seatprobe/seatprobe.go, boards.go   [MODIFY] a reading's Mint counts under its verb
  tools/internal/cli/seat/verbs.go          [MODIFY] PlacesToSentenceOf
  tools/internal/cli/seat/seat.go           [MODIFY] ReasonIs rows for the reading verbs
  tools/internal/cli/lens/verify.go, mint.go, finding.go   [MODIFY]
  tools/internal/cli/blue/edit.go           [MODIFY] the accept's exact span
  tools/internal/cli/m23contradiction_test.go → m23corroboration_test.go  [REWRITE] the migrated state (GAP 2)
  tools/internal/cli/corroborate_test.go, corroborationsha_test.go         [DELETE] (F13) <!-- CHANGED F13 -->
  tools/internal/record/corroborationsource_test.go                        [DELETE] (F13)
  tools/internal/record/redcitation_test.go → foundsource_test.go          [REWRITE] (F13, §III.2d)
  tools/internal/cli/seat/help/corroborate.md                              [DELETE] (F13)
  tools/internal/record/{citationid,evidenceview,anchorchecks,redvoice,supersede,refs}.go  [MODIFY] the label paths (F13)
  tools/internal/report/{assemble,citations,layout}.go                     [MODIFY] blue's cites only (F13)
  tools/internal/record/migrate/{gapplace,remap,registry,replay}.go        [MODIFY] §III.9 items 1, 3, 7
  tools/internal/cli/seat/help/*.md         [MODIFY] incl. edit.md, mint.md (estoppel's exemption)
  tools/internal/terms/terms.json           [MODIFY] failed check (one sentence), reading (NEW), contradiction deleted
  docs/citations-and-bibliography.md, docs/seat-command-triggers.md   [MODIFY]
  tools/releasegate/fuzz/*                  [MODIFY]
  skills/research-protocol/scripts/debate.js  [REGENERATE + MODIFY]
  agents/*.md, skills/*/SKILL.md            [REGENERATE] agentgen
  docs/vocabulary.md                        [REGENERATE] vocabdoc
scripts/agentgen/src/…                      [MODIFY]
```

### III.14 Delivery — one PR, one phase per fresh agent (RULED, gblock 2026-10-30)

ONE PR on one branch, one epoch. No part is a complete concept on its own (deleting the contradiction
before the reading mints leaves a refuting reading holding no PASS; a class no act routes to is a dead
value; agent text cannot precede the flags it describes), so the bundle is the concept, not a default.
Its cost is dominated by the census rewrite (87 REWRITE/MODIFY incl. 31 tests, 4 DELETE, 17
REGENERATE), not machinery: one facet, two Verify fields, one Anchor field, one append path (and its
replay), one ledger state widened, one check state read, one locate by identity.

Implementation runs one phase per **fresh agent**, each starting from this plan and the branch's
`.claude/checkpoints/CHECKPOINT.md`, ending with a commit, the checkpoint rewritten (head, the phase's
loop result, the next phase), and the agent's exit. A phase ends where its loop is green; the full
`check` is owed only at phase 6. **A phase loop builds only what compiles at that phase.** Phase 1's
deletion of `ABOUT_KIND_CONTRADICTION` and `BlockerContradiction` leaves `migrate/replay.go:391-402` and
`planids.go:91-103` uncompiled until phase 2, and every package whose TESTS import `record/migrate`
fails to build its test binary until then. Sweep (`grep -rl 'internal/record/migrate"'` over the tools
module, at `9ee7ad95e`, the same importers as before #1663 but for `migrate`'s own tests): `internal/record` itself (`archivemigrate_test.go`, `planidphrases_test.go`,
package `record_test`), `internal/cli` (`rerunperact_test.go`, `composedsection_test.go`,
`m23contradiction_test.go`, `deadpath_test.go`; `migrateorder_test.go` through `internal/migratecli`),
`internal/report` (`migrateddrift_test.go`), `internal/migratecli`, `cmd/feov-migrate`,
`devcmd/archdump`, `devcmd/archiverun`. So `go test ./internal/record/` cannot run in phase 1 (its
external test package imports migrate): phase 1's loop BUILDS the record package and tests only
`reportproj` and `anchortext`, whose test graphs reach `record` but not `migrate`; the record package's
tests — phase 1's new ones included — run in phase 2's loop, with migrate's. Outside `record`, the
deleted symbols are used by `migrate` (phase 2), `cli` (phase 3) and the fuzz (phase 5) only
(`grep -rln` of the deleted identifiers), so no other package breaks between phases.

| Phase | Scope | Phase loop (must pass to end it) |
|---|---|---|
| 1. Record | §III.1 proto, epoch, facets, `FindsAgainst`/`ReadingGapClasses`, `Verify.carried_by`, `Anchor.sentence_of`; §III.2a's record half (`AttachToSentenceOf`, `report_op`'s `'sentence_of'` arm, `sentenceOfMut`, the Anchor validation, the one-gap placement, `ProposalOver` by identity); §III.3 the check state, the open set from the gap view, and every reader of it; §III.4; §III.5; §III.6 deletions; record-package tests of §III.10 (written, run in phase 2) | protogen/schemagen clean (§V.1); `go build ./internal/record/ ./internal/reportproj/ ./internal/anchortext/`; `go test -count=1 ./internal/reportproj/ ./internal/anchortext/` |
| 2. F13 record half, then migrate | FIRST (F13, §III.2d "Phase-1 code that must change"): the `found_source_gap_classes` facet, `Verify.label`/`work_status`/`sha256` reserved and regenerated, `ReadingGapClasses`' found-source arm, the repeat rule's unanchored arm, its early return (`:319`) and its refusal text, <!-- CHANGED F13b --> `keyFields` losing `url`, `requireNotARetriedReading`, the budget's `anchor` test (§III.5), the SQL readers of `verify.label` (`id.go:88`, `refs.go:130`, `:205`), `reading.go`'s two refusal strings, the record-package deletions of the label paths (`citationid.go`, `evidenceview.go`, `anchorchecks.go`, `redvoice.go`, `supersede.go`, `refs.go`), `record/foundsource_test.go` replacing `redcitation_test.go`, <!-- CHANGED F13b --> `consistency/consistency.go:272`'s and `migrate/proserefs_test.go:112-113`'s `Verify.GetLabel()` (which stop compiling when `label` is reserved; `go build ./internal/consistency/` joins the phase loop); THEN §III.9 translation (`valueRenames` row, the raise recognised by `anchor.Kind(ref)==""` and written as no event, `found_by` entries naming a raise removed, item 7's label translation, <!-- CHANGED F13 --> `TranslateReadingBlockers`, the reading act re-driven whole on replay with each member's own duties and `Batch.AppendBodies`, the deletions in `replay.go` and `planids.go` that phase 1's enum deletions leave uncompiled), `remap.go` `idFields`, `gapplace.go`'s sentence_of check, and the m30/m32/m23/sqlite-schema-source tests and item 7's six-archive test and its no-token test over every archive (13b) <!-- CHANGED F13b -->. `cli`, `report` and the fuzz, which read the deleted Verify fields, build in phases 3 and 5 | §V.6; `go test -count=1 ./internal/record/...` (the record package's tests, phase 1's included, and migrate's); `go build ./internal/migratecli/ ./cmd/feov-migrate/ ./devcmd/archdump/ ./devcmd/archiverun/` (not `./devcmd/...` or `./cmd/...`: `devcmd/seatprobe` and `cmd/feov-record` import `internal/cli`, which builds in phase 3) |
| 3. Reading act | F13's CLI half: `verify`'s `--url` form and flag groups, `corroborate`'s deletion (verb, help page, registrations), the marked-sentence guard, the footnote path's deletion in `verify.go`, `report`'s red-source arms, `seat/source.go`, `seatprobe` boards row <!-- CHANGED F13 -->; §III.2 flags (`gapFlags`, `ReadingClassField`, `RequiredUnless`, `ReasonIs`), the GAP 2 and GAP 3 result words, §III.2c (F12), the field table, the completeness stamp and the footnote gate, the reading's placed append, held gap inside the transaction, repeat, result words; §III.2a's CLI half (`PlacesToSentenceOf`, `verifiedFix`, the accept's exact span); §III.8 `seatprobe`; cli tests, the GAP 3 record-at-this-epoch test | `go test -count=1 ./internal/record/... ./internal/cli/... ./internal/migratecli/ ./internal/report/ ./internal/capture/ ./internal/seatprobe/ ./devcmd/seatprobe/ ./cmd/...` |
| 4. Agent text | §III.7 help, terms, agentgen sources (F13's included), debate.js (`:879` included); regenerate; read every changed golden line <!-- CHANGED F13 --> | §V.1 second run no diff; §V.4 census and phrase sweep |
| 5. Fuzz | §III.11 drives, gates, the four carried tally fields; every `corroborate` drive becomes `verify --url`, with `foundSourceGaps` and `markedSentenceRefused` <!-- CHANGED F13 --> | `go test -count=1 ./releasegate/...` with the gate unset (compiles; the sweep runs in CI) |
| 6. Verify | §V.2, §V.3, §V.5, §V.8, §V.9; open the PR; read PR CI's fuzz tally (§V.7) | all of §V |

### III.15 Required tests, in order, by the phase that owns each (approved 2026-11-01; F13's rows audited by the targeted round) <!-- CHANGED F13 -->

Each phase writes its tests before the code that turns them green, and does not end until each runs
and passes in its phase loop (§III.14) and its §V.5 partner, applied and reverted, turns it red. A
phase-1 test is written in phase 1 and first runs in phase 2 (the record package's tests import
`migrate`). NEW unless marked; "R8" marks a round-8 gap or note carried as a required test.

**Phase 1 — record** (written; run in phase 2)
1. `TestEveryNegativeOutcomeNamesItsClasses` (and `found_source_gap_classes`, F13; written in phase 1, extended in phase 2) — §III.1; partners a, nn. <!-- CHANGED F13 -->
2. `TestTheLensCloseOfAReadingsGapIsRefusedWhileTheCitationStandsFoundAgainst` — §III.3; partners d, r.
3. `TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld` — F8; partners d, n.
4. `TestASupersededReadingGapHandsTheCheckToItsSuccessor` — §III.3; partners o, aa.
5. `TestAFoundAgainstCheckIsWordedByItsState` — §III.3; partner k.
6. `evidenceview_test.go`'s `not_the_work` `refutes` row (REWRITE) — R8 GAP 2; partner jj.
7. `TestEveryIDTheVerbPrintsIsTranslated` (REWRITE: `BlockerWhy` per state) — §III.9 item 4; partner ii.
8. KEEP, the regression net, unchanged: `TestABenchRulingDoesNotClearAnUnreadCitation`, `unreadb6_test`,
   `blueanswered_test`, `failedcheck_test`, `reopened_test` (three-argument `failedChecksOf`),
   `tiersubstitution_test`, `unsoundm32_test`, `TestTheReportReplayIsAFold`.

**Phase 2 — F13 record half, then migrate** <!-- CHANGED F13 -->
9a. `TestReadingGapClassesCoversBothForms` (`record/foundsource_test.go`): every `SourceOutcome` ×
    `SourceCompleteness` × {anchor set, anchor empty} gives the classes §III.1 states, and `FindsAgainst`
    is false on every backing found-source reading — partner nn.
9b. <!-- CHANGED F13b --> `TestAFoundSourceReadingKeysOnSourceAndSentence` (record layer, `AppendReading`;
    renamed from F13's `TestAFoundSourceReadingKeysOnItsGap`): one seat, one sitting — (url1, s1) mints G1;
    (url1, s2) records and mints G2; (url2, s1) records naming G1 (`carrier_minted` false); an `absent` of
    url3 on s1 and on s2 records both; (url1, s1) again, any outcome → refused by
    `requireNotARetriedReading`, its text naming the first reading's `events.id`; the same (url1, s1) in
    the seat's NEXT sitting with another outcome → recorded — partners oo, vv.
9c. `TestARepeatedFoundSourceReadingIsRefused`: the same seat's second backing reading of the same
    sentence from the same url, its gap open, in a later sitting → refused naming the gap; the refusal
    holds no "finding" <!-- CHANGED F13b --> and none of "finds the source against" (the found-source form's
    text, §III.2) — partner pp.
9e. <!-- CHANGED F13b --> `TestTheRepeatRuleReachesABackingFoundSourceReading` (record layer): drives
    `requireNotARepeatReading` itself on a backing found-source Verify whose earlier twin's gap is open,
    and asserts the refusal — a change to the loop alone, with `failedcheck.go:319`'s early return left on
    `FindsAgainst`, leaves it red — partner xx.
9d. `TestNoReadingPlacesACitation` (record layer): after `verify`'s writes in both forms and every
    outcome, `record.CitedSources`, `show evidence`'s sources and the assembled Bibliography hold blue's
    cites only, and no Verify carries a label — partner qq.
9. `TestTheMigratedM30RaiseIsDropped` (renamed; its m32 row included) — §III.9 items 1, 3; partners f, v. <!-- FINDING-DEPENDENT -->
10. `TestAFindingAboutAnAnchorIsNoRaise` — §III.9 item 1; partner v.
11. `TestTheMigratedM32HoldsAFailedCheckPerStandingReading` — §III.9 item 4.
12. `TestAMigratedFoundAgainstCheckReadiesBlueAlone` (m32 cut and sqlite-schema-source rows) — §III.3; partners i, j, r.
13. `TestM23CorroborationReadingComesForward` — §III.9.
13a. `TestTheMigratedRedCitationsAreReadingsOfTheirSources` (six archives; <!-- CHANGED F13b --> the EXACT `reopened` lists, the 3 re-readings' url and title) — §III.9 item 7; partners rr, zz. <!-- CHANGED F13 -->
13b. <!-- CHANGED F13b --> `TestNoMigratedArchiveHoldsALensPlacedToken` (every archive; the 15 labels in both spellings, token and bare id, outside prose) — §III.9 item 7; partner yy.
14. KEEP, must stay green: `TestEveryArchivedRunMigrates`, `TestValueRenames`, `TestEveryStringFieldIsInOneList`, `reopenedm19_test`, `foundagainstm30_test`.

**Phase 3 — the reading act**
15. `TestARefutingReadingMintsItsGapInTheSameAct` — partners a, b.
16. `TestASecondRefutingReadingNamesTheStandingGap` — partner c.
17. `TestANegativeReadingWithoutGapFlagsIsRefused`, `TestABackingReadingWithGapFlagsIsRefused` — §III.2.
18. `TestAFoundSourceAbsentMintsNothing` (renamed) — F5. <!-- CHANGED F13 -->
19. `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` (rows (a)-(i) and the record-layer row; row (b) carries R8 GAP 2) — partners s, w, jj.
20. `TestAReadingsGapTakesNoMint` (<!-- CHANGED F13b --> `verify --anchor` rows only) — F6; partner e.
21. `TestAPlainMintALaterReadingNamesKeepsItsMint` — partners p, bb.
22. `TestTheReadingsHeldGapIsDecidedInsideItsWrite` — partner u.
23. `TestARetriedNegativeReadingIsRefusedNamingItsGap` (`cli/keyedretry_test.go`) — partner m.
24. `TestAReadingOfAFoundAgainstCitationIsWordedByItsState` — partner l.
25. `TestAVerifyOfARepeatedSentencePlacesItsGapByTheAnchor` (rows (a)-(i)) — F9; partners t, z.
26. `TestAReadingOverAnAppliedPrescriptionMintsItsGap` — F10; partner cc.
27. `TestAReadingsGapIsFilledAsAPlainMints` (its GAP 3 row is R8 GAP 3; the `mint_key` presence is R8's note) — F11; partners ee, ff, kk.
28. `TestAReadingsGapCarriesItsArgumentWhereverAMintsDoes` (reads docket.md; asserts report.md lacks R — R8's note) — F11; partners ee, ff.
29. `TestAReadingsGapNamesItsSource` — F12 (R8 fork); partner ll.
30. `TestTheReadingVerbsDescribeTheirGapFlagsByTheirSubject` — partners dd, hh.
31. `TestAReadingIsAttributedToTheVerbTheSeatTyped` — §III.8; partner gg.
32. `TestM23sArchivedCorroborationHoldsNothing` (`cli/m23corroboration_test.go`, renamed) — GAP 2 of round 3.
33. `TestARecordWrittenAtThisEpochMigratesWithNoRefusal` (<!-- CHANGED F13b --> with lens N's `verify --url` minting row) — §III.9 item 6; partner y.
34. `TestMigrateChecksAReadingsGapBySentenceOf` (`migrate/`; its stream is written by the verbs) — partner y.
35a. `TestAFoundSourceReadingMintsItsGap`, 35b. `TestOneSourceBacksManySentences` (<!-- CHANGED F13b --> its `absent` rows: two sentences both recorded, one sentence twice refused naming the first), 35c.
    `TestABackingFoundSourceReadingOfAMarkedSentenceIsRefused` (<!-- CHANGED F13b --> its no-gap-flags row meets the guard first), 35d. `TestVerifyTakesExactlyOneSubject` —
    §III.2d; partners ss, oo, tt, uu. <!-- CHANGED F13 -->
35e. <!-- CHANGED F13b --> `TestOneSentenceReadAgainstManySources` (b7's shape, events 152-155, through `verify --url`: four urls, one sentence, four Verifys, one gap; the retry refused naming the stored reading) — §III.2d, gblock's key ruling; partners oo, vv.
35f. <!-- CHANGED F13b --> `TestAFoundSourceGapTakesAMint` (both arms: the found-source mint refused past the budget, `unsourced_claim` and `claim_contradicts_source`; the anchor-reading mint admitted past it) — §III.5, gblock's budget ruling; partners e, ww.
    <!-- CHANGED F13b --> Name grep at `6102f2b09`, every name F13b coins (`requireNotARetriedReading`, 9b's, 9e's, 13b's, 35e's, 35f's tests, the two tally fields): 0 hits.
35. Gates REWRITTEN for the reading verbs: `TestHelpSaysRequiredExactlyWhereOmissionIsRefused` (`requiredProbes`; <!-- CHANGED F13b --> `oneOfAnother`'s `anchor`/`url` pair and `"verify"` in `verbsWithAOneOf`), `TestEverySetShapedFlagIsEitherDeclaredOrExempt` (three `enforcedElsewhere` sites, R8's note), `TestEveryFlagAHelpPageNamesIsOnThatPage` — partners dd, hh.

**Phase 4 — agent text**
36. `TestTheEmbeddedRegistryLoads` (the one-sentence `failed check`, R8's note) and `TestNoDefinitionUsesABannedVariant` over the new `reading` entry (R8 GAP 1) — partner mm.
37. `TestVocabularyProse`, `TestNoPromptGrowsItsCommandCatalogue`, `TestTheSeatPromptsNameNoVerb`, the manual golden (`difftest`), `agentgen -check`, `enginegen`'s region check, `vocabdoc -check` — §V.1, §V.2.

**Phase 5 — fuzz**
38. `TestTheFailedCheckDrivesKeyOnTheState` — partner q.
39. `TestSweepTallyRoundTripsEveryField` and `TestTheFoldReadsEveryRunFieldAndWritesEveryFoldField` over every new tally field and #1642's four (F7) — §III.11.

**Phase 6 — verify**
40. Every §V.5 partner applied and reverted, each turning its test red (a–mm); §V.4's sweeps; §V.6;
    §V.8 (or NOT MEASURED); §V.9 on real data; PR CI's fuzz tally read (§V.7), the four #1642 fields
    reported beside #1642's run (§I criterion 5); then the independent review (§V.10).

**Commit trailers.** `scripts/rulesweep` reads `Rule-Class:` / `Sibling-Sweep:` from each commit and
reports a protocol-surface edit (`skills/*/SKILL.md`, `agents/*.md`, `references/*.md`, `debate.js`,
`law/`, `hooks/hooks.json`) whose commit lacks them. Phase 4's commit touches every one of those but
`law/` and `hooks.json`; any other phase commit that regenerates an agent definition (a help page is
inlined into it) is in the same position. Every phase commit of this PR carries, in the body, before
`Co-Authored-By`:

```
Rule-Class: definition-forked-not-widened
Sibling-Sweep: every carrier of the contradiction and of a reading that finds against its sentence —
  record (proto, fold, blockers, dispatch, sitting, retirement, budget, one-gap rule), the reading
  verbs and their help, every reader and printed sentence of a failed check (§III.3), agentgen
  sources and the generated agents and skills, debate.js (enginegen regions and the chair prompt),
  terms and vocabulary, migrate, seatprobe, capture's assembly screen, docket.md, the release fuzz;
  and (F13) `corroborate`, the red footnote, label and Bibliography paths, and the archived labels;
  census §III.12 (its F13 census included) re-run as §V.4
```

(F13 widens the trailer's sweep; the class is the same: the labelled corroboration forked the citation's placing for a lens.) <!-- CHANGED F13 -->
The class: the gap's lifecycle was FORKED for a reading against its sentence (a contradiction, with
its own id space, exemptions and blocker kind) instead of WIDENED; this plan absorbs it into the gap
and the failed check (`feov-memory/protocol-class-registry.md`, `definition-forked-not-widened`). Each
phase names in its own trailer line what it swept, keeping the class.

### III.16 Seam with #1657 Part 1a′ (`plans/compute-and-show.md`, branch `plan/anchor-lifecycle`) (RULED, gblock 2026-11-01)

**The ruling.** After the two sessions discussed the conflict, gblock picked their joint
recommendation: this plan governs a refuting reading; #1657 Part 1a″'s L4 is dropped; contradictions
are deleted everywhere, corroborations included; #1657 Part 1a′ continues; this plan lands FIRST, then
1a′. (Read at `plan/anchor-lifecycle` `1bcda6a0c`, a local branch: no remote ref exists at this
writing.)

**What 1a′ implements on top of this plan; this plan implements none of it.** The `source` check kind
stays unenforced on main until 1a′: today only `computation` is demanded at close
(`cli/lens/close.go:63-67`, `requireComputationAnswered`).
- **L1 widens.** A gap's `source` demand is met when ANY retired anchor is its `about_ref`, or is the
  anchor of a Verify whose `carried_by` names this gap. A close stays refused by
  `requireFailedCheckAnswered` while another reading's check stands — the demand does not replace the
  check.
- **A backing re-read arm in `evidence_answered`**, defined off rows: the anchor of a carrying Verify
  (one whose `carried_by` names the gap) is not retired, and the latest Verify of that anchor by that
  lens is a backing outcome.
- **The closure demand stays on the `close` verb's write** (a `Close` event, `close` and `carry`), so a
  bench `closure_act` — a closing ruling, which writes no Close row — stands, and replay skips the
  demand (`Migrating`), as it skips `requireFailedCheckAnswered`.

**This plan's obligation to that seam.** `carried_by` and `carrier_minted` are shipped and documented
(the proto comments, the `reading` term, `show evidence`'s row, docket.md's line); several readings may
carry one gap (each later one with `carrier_minted` false, §III.2's held gap); and both stay readable by
SQL — columns of the `verify` table, `carried_by` a foreign key to `mint.gap_id` — so 1a′'s views read
them without a Go reader.

**The overlap, stated.** On a reading's gap the demand mostly duplicates the check: while a reading's
found-against check stands, the close is already refused (`requireFailedCheckAnswered`), and the check
clears by the acts the demand's two new arms read — a retire of the citation (the widened L1) and the
reader's own backing re-read (the re-read arm). The demand adds over the check exactly two cases:
- **an edit.** Blue's edit of the sentence clears the check (the citation becomes `unchecked`), and on
  main this plan admits the close then; under 1a′ a `source`-kind reading gap's close also needs a demand
  arm — the reader's backing re-read of the edited sentence, a retire, or a live answer (`cite
  --answers`). So this plan's close-after-edit rows — the edit subtest of
  `TestTheLensCloseOfAReadingsGapIsRefusedWhileTheCitationStandsFoundAgainst`, the edit arm of the fuzz's
  `foundAgainstEdits` drive, and §V.9's drive if it closes after an edit — are carriers of 1a′'s demand,
  green at this plan's merge, and 1a′'s census rewrites them;
- **a found-source reading's gap** (F13; `verify --url`, `unsourced_claim` or
  `claim_contradicts_source`). Its Verify has no anchor, so neither the widened L1 nor the re-read arm
  can reach its gap: under 1a′ its `source` demand is met by blue's live answer (`cite --answers G` on
  the sentence — the cite the ruling names) or by the bench, and the new citation is `unchecked`, so a
  lens's `verify --anchor` of it is owed before the PASS. 1a′'s corroboration arm — V1a′.10(6), "a
  refuting `corroborate` minted `--check-kind source` → met only by `cite --answers G`", and its rows at
  `compute-and-show.md:693` ("Red's corroborations: red places a `C-` it found") and `:758-759` — is
  DELETED: no `corroborate` exists and no reading places a `C-`; 1a′'s census rewrites (6) as "a
  refuting or backing `verify --url` minted `--check-kind source` → met only by `cite --answers G`". On
  main, with no check carrying a found-source gap (§III.3), its lens closes it on its own judgement. <!-- CHANGED F13 -->

**F13b's consequences for 1a′.** <!-- CHANGED F13b -->
- **A sentence 1a′ states is false after F13.** `compute-and-show.md` (scratch copy
  `~/.claude/scratch/plan1647-f13/cas.md:812`) says "every reading gap is material by its class". A
  backing `verify --url` reading's gap is `unsourced_claim`, `CLASS_MATERIAL_BY_GRADE` (O1 closed,
  §III.2d), so it is material by its grades, not its class; 1a′'s census rewrites that sentence.
- **Several readings may carry one gap from several sources** (F13b's key): one seat in one sitting reads
  several urls against one sentence and each later reading names the first's gap. 1a′'s demand over the
  carrying Verifys reads them all.
- **E1′, the cross-check once both land.** 1a′'s E1′ ("no verify row carries a label outside
  evidence_anchor") is the cross-check of item 7's drop of the 15 lens-placed anchors: after this plan no
  Verify has a `label` column at all (§III.1), so E1′ holds vacuously on every migrated record, and a
  failure of it then means a label came back. Until 1a′ lands, `TestNoMigratedArchiveHoldsALensPlacedToken`
  (§III.15 13b) is the gate.
- **Two 1a′ picks gblock made 2026-11-03 (relayed by the #1657 session; the option wording is that
  session's).** (a) A gap carried by several readings: any one reading that meets the demand is enough;
  after an edit, another reading's `unchecked` holds the PASS until it is re-read. (b) A claim retire
  that removes a citation and the computation citing it in the same act is allowed. Neither touches a
  row of this plan: (a) is the `source` demand's arm over carrying Verifys, which this plan does not
  build (it ships `carried_by` and admits several carriers, §III.2's held gap, and nothing it states
  bounds how many must meet a demand), and the `unchecked` state an edit leaves is main's ledger,
  unchanged here (§III.3); (b) is `blue retire`'s scope, which this plan does not touch. Both are 1a′'s.

**#1695 (findings fold into gaps).** Relayed by the peer session bridge-cse-01s7d2sa1pneo9oqzevmlwdg-4f
(2026-11-03): gblock ruled findings fold into gaps under a separate plan, issue #1695, landing after this
plan and 1a′. This plan makes and keeps no finding: it deletes the contradiction about-kind, the raise
and its credit (§III.6, §III.9 items 1 and 3), and nothing else of the finding verb, `found_by`, the `F-`
kind or docket.md's provenance; what the reading verbs' result and repeat refusal used to advise ("add
what it lacks as a finding about G") and the `found_by` field's future are #1695's. <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->

**Checked against this plan, for anything that contradicts the seam:** nothing writes `carried_by` from
`found_by` (§III.9 item 2, ruled); nothing makes `carried_by` unique per gap; the reading's Mint leaves
`about_kind`/`about_ref` unset (§III.2's field table), which is why L1 must widen through `carried_by`
rather than through `about_ref`; no sentence says a closing ruling is refused or that the bench's
closure falls. The only sentences that state when the lens's close is admitted (§III.0 row 3, §III.3,
§III.7's `close.md`, the `failed check` term) speak of the check, which 1a′ does not change; §III.0 row 3
names the seam.

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation (step) |
|---|---|---|---|
| R1 | A carrier keeps speaking the contradiction (prompt, help, golden, fuzz) — half-state that reads done | H × H × L | §III.12 census re-run as §V.4, identifiers grep must be empty; agentgen/enginegen/vocabdoc `-check` gates (§V.2) |
| R2 | Lenses stop recording negative readings because the act now costs a full mint's flags | M × H × M | Flags are the same set `mint` already asks; refusal names each; smoke §V.8 checks negative readings still occur; a citation reading's gap takes no budget (F6) so cap is not a reason to read `supports`; a found-source reading's gap takes a mint (F6 amended), as its plain mint would <!-- CHANGED F13b --> |
| R3 | One-gap rule + lens-picked class lets two lenses split one defect across two classes on one sentence | M × L × L | Ordinary behaviour for any gap; the admitted sets are two classes each; sweep tally reports per-sentence gaps by class |
| R4 | A bench `not_a_defect` ruling on a misread source closes the gap, and the run still cannot PASS while the reading stands | M × M × M | Ruled (F8): the ruling closes the gap through the bench arm, and the failed check alone holds the PASS, as an unread or unsound check does today; a failed check is not a dispute. It clears when the reader's own later reading backs the sentence (its reading replaced), or blue edits or retires — another lens's backing reading clears nothing. Stated in the bench's constitution (§III.7) and pinned by `TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld` |
| R5 | migrate drops something other than a raise, or leaves a `found_by` naming a dropped raise | L × H × L | A raise is recognised by `anchor.Kind(ref)==""` and a ref that names a Verify, else refused (§III.9 item 1); `TestTheMigratedM30RaiseIsDropped` (m30 and m32 rows); `TestEveryArchivedRunMigrates` <!-- CHANGED F13 --> |
| R6 | Epoch collision with another epoch PR landing first (#1644, #1652, #1660 and #1663 each took one while this plan was audited) | M × M × L | Re-derive at rebase; `schemaepoch_test.go` and setup's epoch refusal catch a stale number |
| R7 | Goldens conflict semantically with a help-changing PR merged meanwhile | H × M × L | Rebase, regenerate at `-count=1`, read each changed line, re-run CI before merge (CLAUDE.md) |
| R8 | The reading's three-event act (Mint, Anchor, Verify) leaves a Verify without its Mint on crash, or a retry records a second pair | L × H × L | One transaction, so a crash leaves nothing; a retry after the commit meets the repeat refusal naming the stored reading and its gap, nothing recorded (§III.2; `TestARetriedNegativeReadingIsRefusedNamingItsGap`) |
| R9 | Release fuzz loses coverage of the duplicate path | M × M × L | New zero-gates (§III.11) fail the sweep if the one-gap path in the reading act is never driven |
| R10 | A reader of a failed check keeps keying on `answeredBy == ""` or `Detail == ""`, so a found-against check is owed by its reader (never retires, envelope refused), prints an owed mint or "stays unread" that does not exist, or feeds the fuzz's unread tallies and keeps their false-green guards green | M × H × L | §III.3 state key on every reader, the verb's sentence included (§III.2); the fuzz keyed on the state with its own tallies (§III.11); the failed-check sweep re-run as §V.4; `TestAMigratedFoundAgainstCheckReadiesBlueAlone`, `TestAFoundAgainstCheckIsWordedByItsState`, `TestAReadingOfAFoundAgainstCitationIsWordedByItsState`, `TestTheFailedCheckDrivesKeyOnTheState`; delete-the-row checks §V.5 i–l, q |
| R11 | A carrier is closed into a successor, or named by a reading that did not mint it, and the close gate or the budget reads the wrong gap | L × M × L | The carrier follows `Close.successor` (§III.3); the budget exemption reads `carrier_minted`, set only in the act that minted (§III.5); §V.5 o, p |
| R12 | A site decides "negative reading" from the outcome alone, so a verdict on a `not_the_work` copy demands gap flags or mints a gap no failed check carries, beside the `unread` check's owed gap | M × H × L | One predicate, `ReadingGapClasses`, decided after the completeness stamp (§III.1, §III.2); the class sweep of §III.2b re-run as §V.4; `TestAReadingOfACopyThatIsNotTheWorkMintsNothing`; §V.5.s; the fuzz's `not_the_work` source (§III.11) |
| R13 | A reading's gap placed by text lands in, or is refused by, another copy of its sentence's words; or a sentence_of Anchor drops out of the render silently; or a clause citation pulls the gap's token into mid-sentence, where the accept and the board do not find it | M × M × L | Placement by `Anchor.sentence_of` at the sentence's end (§III.2a); `report_op`'s `'sentence_of'` arm and a loud replay error where the token is missing; `TestAVerifyOfARepeatedSentencePlacesItsGapByTheAnchor` (replay byte-identical, clause-cited row); §V.5.t; the fuzz's repeated sentence |
| R14 | A record written at this epoch is refused at its next migration — the reading's Mint refused for budget on replay, or its `carrier_minted` Verify refused outside its act — so every live run's record becomes un-migratable | M × H × L | migrate re-drives the reading act whole (§III.9 item 6); `TestARecordWrittenAtThisEpochMigratesWithNoRefusal`; §V.5.y |
| R15 | Locating a prescription by the gap's token changes the accept for every gap, and an exact-span accept edits text a typed edit would have refused as ambiguous | L × M × L | The exact span applies only to the accept's token-carrying `old`, which occurs once by construction; a typed `blue edit` keeps the ordinary locate (§III.2a); rows (e), (f), (i); §V.5.z; the board, the accept and estoppel read one `ProposalOver` |
| R16 | `verify` stamping completeness turns a backing reading of a sign-in page into an unread citation that holds the PASS and owes a gap, where today it cleared the citation | M × M × L | Ruled (GAP 1): that is the intended model, already stated in `help/verify.md:11`; the owed gap is the `unread` check's existing duty, outside the budget; `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` (g); §V.5.w; the fuzz's `not_the_work` source with a backing outcome |
| R17 | A lens uses a reading to reopen text it prescribed and blue applied, which `mint` would refuse as an estoppel | L × M × L | Ruled exempt (F10): the reading must read the work and find the source against the sentence, and the reading itself is on the record as the evidence (its source and outcome in `show evidence`, its `--reason` as the gap's `mint_reason` on the board and in docket.md's "minted because", open-gap block and closure index), where blue and the bench weigh it; `TestAReadingOverAnAppliedPrescriptionMintsItsGap`; §V.5.cc; the fuzz's `readingsOverPrescriptions` |
| R18 | A field the reading's Mint leaves blank reads as absent where a plain mint's reads: the lens's argument gone from the board and the report, or a `fix_basis` the board calls never omitted | M × M × L | Ruled (F11); §III.2's field table; one fill (`gapFlags.fill`); `TestAReadingsGapIsFilledAsAPlainMints` walks `Mint`'s descriptor, so a new field needs a disposition; `TestAReadingsGapCarriesItsArgumentWhereverAMintsDoes`; §V.5.ee, ff; the fuzz's stored-Mint check |
| R20 | #1657 Part 1a′ lands on top and a reading's gap meets a `source` demand this plan's close-after-edit rows do not meet, or 1a′ reads a field this plan stopped writing | M × M × L | §III.16 states what 1a′ implements and what it reads; `carried_by` and `carrier_minted` are shipped, documented, readable by SQL, and several readings may carry one gap; the close-after-edit rows are named there as 1a′'s to rewrite |
| R19 | A tool that maps an event to the act a seat typed reads the reading's Mint as a `mint` | L × L × L | §III.8's sweep; `TestAReadingIsAttributedToTheVerbTheSeatTyped`; §V.5.gg |
| R21 | A lens stops reading sources it found because the act now costs a gap's judgment flags and a mint of its budget, and `unsourced_claim` graded low holds no PASS | M × M × L | One act still; the flags are `mint`'s and the mint is the one a plain `unsourced_claim` mint would take (F6 amended); materiality by grade is the class's, as for every `unsourced_claim` gap (O1 closed); the smoke (§V.8) counts `verify --url` readings <!-- CHANGED F13 --> <!-- CHANGED F13b --> |
| R22 | migrate's removal of a dropped label's token leaves an archived edit, finding or anchor quote unplaceable, or an exact span off by the removed bytes | M × M × L | The token is removed from every field by the one respelling path (§III.9 item 7); `TestTheMigratedRedCitationsAreReadingsOfTheirSources` drives the six archives, b9's two edits over the label included <!-- CHANGED F13 --> |
| R23 | One seat's found-source readings in a sitting that differ in url or in sentence are refused as duplicates (one url against several sentences; several urls against one sentence, b7's shape) | M × M × L | Ruled (F13b): `url` leaves `keyFields`, no `carried_by` key; the identity is (seat, sitting, url, sentence), checked by `requireNotARetriedReading` (§III.1); `TestAFoundSourceReadingKeysOnSourceAndSentence`, `TestOneSourceBacksManySentences`, `TestOneSentenceReadAgainstManySources`; §V.5.oo, vv <!-- CHANGED F13 --> <!-- CHANGED F13b --> |
| R24 | migrate strips a dropped label's token from some fields and not others, and `proseRefs`' silent skip of an id the record does not hold (`proserefs.go:74`) hides the miss; or an id list keeps an empty element | L × M × L | The token is removed in every field; list elements are removed, not respelled; `TestNoMigratedArchiveHoldsALensPlacedToken` (both spellings), 13a's exact lists; §V.5.yy, zz <!-- CHANGED F13b --> |

## V. Verification Plan

All commands from the worktree root, after `unset $(env | grep -o '^FEOV_[A-Z_]*')`.

1. `go -C scripts run ./protogen && go -C scripts run ./schemagen && go -C scripts run ./agentgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen && go -C scripts run ./vocabdoc` → no diff on a second run. Re-armed by: `record.proto`, `requirements.json`, `scripts/agentgen/src/**`, `terms.json`, the roster/enums.
2. `go -C scripts run ./check` → `0 failed` (read the count, not the exit code). Re-armed by: any file in the diff. Includes the `-check` gates of every generator, `archaeology`, `golden`, `pluginparity`, vocabulary prose.
3. `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/... ./internal/cli/... ./integration/... ./releasegate/...` (release gate unset: the fuzz sweep does not run) → 0 FAIL. Goldens: `UPDATE_GOLDENS=1 go test -count=1 <pkg>` then read every changed line. Re-armed by: any Go or help change.
4. Census re-run (must print only KEEP lines of §III.12):
   (F13) `grep -rIn -i corroborat --exclude-dir=.git --exclude-dir=run-archive plugins scripts docs` prints only §III.12's F13 KEEP lines and migrate's archived-word reasons; `grep -rIn -E 'newCorroborate|ExistingCorroboration|refuseCorroborationTitleVoice|CorroboratedBy|corroborated_by|backsTheClaim|ReadingClassField\(corroboration' plugins scripts` prints nothing; every `GetLabel()` on a `Verify` is gone (`Verify` has no `label`); the phrase sweep of §III.12's F13 census prints no live sentence that has a lens's reading place a citation, a footnote or a Bibliography line; <!-- CHANGED F13 -->
   `grep -rIn -i contradict --exclude-dir=.git --exclude-dir=run-archive plugins scripts docs | grep -v -E 'cross_section_contradiction|claim_contradicts_(own_record|source)'`
   and identifiers, which must print nothing outside `migrate/` archived-word tables:
   `grep -rIn -E 'ABOUT_KIND_CONTRADICTION|BlockerContradiction|unanswered_contradictions|contradictionsOf|ContradictionOfReading|requireOpenContradiction|answersRaisedContradiction|answersOwnRaise|gap_may_name|raised_by' plugins scripts`;
   the phrase sweep of §III.7, over the whole tools tree — `tools/internal` (not only `cli/`) and `tools/releasegate` — and the agent-text paths (must print only its KEEP lines); the "does not close" sweep of §III.7 (every hit in the failed-check sense says the close is refused, never that the gap does not close); the failed-check sweep of §III.3 / §III.12 over `tools/internal`, `tools/cmd`, `tools/releasegate` and `debate.js`, every `answeredBy`/`AnsweredBy()`/`.Detail`/`BlockerFailedCheck` hit read for a branch that decides owed-ness, wording or a tally without the state (must find none); and `grep -rn 'carried_by\|carrier_minted'` over the tools tree, every hit one of §III.12's "Readers of the new Verify fields"; `grep -rn 'sentence_of\|SentenceOf\|sentenceOfMut\|ErrAnchorGone\|ErrAnchorBare\|AppendBodies'` over the tools tree, every hit a row of §III.2a's table (which names `migrate/replay.go` and `record/batch.go`) or a test; the proposal-span sweep of §III.2a (`ProposalOver`, `PlanAccept`, `proposalAt`, `ProposalAppliedVerbatim`, `PlanSplice`, `LocateUnique`), every locate of a gap's prescribed span by the gap's token and every typed edit's by its quote; the completeness sweep of §III.2b (every Verify writer stamps `source_completeness`, every reader's row as listed); the estoppel sweep of §III.2 (every hit that teaches or enforces the prescription estoppel states the reading's exemption, or is a KEEP row there); the replay-duty sweep of §III.9 item 6 (every per-source-event step of `replay.go`'s loop runs per member of a reading act, and no member's `newID` entry is read from `MAX(id)`); the field table of §III.2 (every `p.<Field> =` in `gapFlags.fill` and `newMint`, each a row, and every reader of an unset field a row); `grep -rn 'gapFlags\|ReadingClassField'` over the tools tree, every hit in `mint`, `verify`, `corroborate`, `record/enums.go` or a test; the `lane` table of §III.12 (`grep -rn 'Lane\.Role\|LaneVerbs\|blue-lane'`, each hit a row or other sense); the attribution sweep of §III.8 (`grep -rn 'verbOfEvent\|EVENT_TYPE_MINT\b' tools/internal tools/devcmd`, no tests or goldens); and the GAP-1 class sweep of §III.2b — `grep -rn -E 'SOURCE_OUTCOME_(REFUTES|ABSENT)|contradicts\(|VerifyReads|FindsAgainst|ReadingGapClasses'` over `tools/internal` and `tools/releasegate` — every hit that decides whether a reading mints, words its gap or counts it calls `ReadingGapClasses` (or `FindsAgainst` for the ledger), never the outcome alone.
   Also, at this revision: the outcome-STRING sweep of §III.2b (`grep -rn '"refutes"\|"absent"\|Refuted()'` over `tools/internal`, `tools/cmd`, `tools/releasegate`, no tests) — every hit that counts or words a found-against reading calls `FindsAgainst`, the rest are `hookprovenance`'s other sense; the seat-facing "reading" sweep of §III.7 — every hit the term's sense, the OCR collision or ordinary English as listed, and no new prose says "a verdict that reads nothing"; `grep -rn 'CarriedBy\|carried_by'` over `record/evidenceview.go` and `report/assemble.go`, each hit a reader of §III.2c; and `mintResult`'s fields against `readingGap`'s (GAP 3: every field of `mintResult` a reading's gap has is echoed).
   Re-armed by: any agent-facing or Go edit.
5. Delete-the-row checks (each must turn the suite red; revert after):
   a. delete `(gap_classes)` from `SOURCE_OUTCOME_REFUTES` → `TestARefutingReadingMintsItsGapInTheSameAct` and `TestEveryNegativeOutcomeNamesItsClasses` fail;
   b. delete the Mint body from the reading's append (keep the Verify) → the same test fails;
   c. delete the held-gap branch in the reading path → `TestASecondRefutingReadingNamesTheStandingGap` fails;
   d. delete the found-against branch of `failedChecksOf` → `TestTheLensCloseOfAReadingsGapIsRefusedWhileTheCitationStandsFoundAgainst` and `TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld` fail;
   e. delete the act-paired exemption in `requireMintWithinBudget` → `TestAReadingsGapTakesNoMint` fails (the reading's Mint is refused at the cap);
   f. delete migrate's raise drop (write the raise as decoded) → `TestTheMigratedM30RaiseIsDropped` fails (a refusal: its `about_ref` names no anchor); keep the raise's id in a `found_by` → its G-dc877265 / G-f417112d assertions fail; <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->
   g. delete the call site of the new flag-registration helper in `verify.go` → help/flag tests fail;
   h. delete the write-path rule that a reading whose `ReadingGapClasses` is non-empty names its gap (`record.go`, the Verify case) → a test writing such a reading with no gap fails;
   i. revert `retirement.go`'s failed-check skip to `answeredBy != ""` → `TestAMigratedFoundAgainstCheckReadiesBlueAlone` (retirement) fails;
   j. revert `sitting.go`'s lens arm to `answeredBy == ""` → the same test (work list, envelope) fails;
   k. delete the `found_against` branch of `failedOwed` / `blueFailedItem` → `TestAFoundAgainstCheckIsWordedByItsState` and the m32 test (printed sentence) fail;
   l. restore `verify.go`'s `!VerifyReads(stored)` guard and the unread-only `words()` → `TestAReadingOfAFoundAgainstCitationIsWordedByItsState` rows (a), (b), (d) fail;
   m. drop the negative-reading arm of `requireNotARepeatFailedCheck` → `TestARetriedNegativeReadingIsRefusedNamingItsGap` fails (a second Verify lands);
   n. make `failedChecksOf` skip a found-against check whose carrier is closed (the shape that would let a ruling clear the check) → `TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld` fails (PASS admitted);
   o. read the carrier from `carried_by` alone, without `successorsOf` → `TestASupersededReadingGapHandsTheCheckToItsSuccessor` fails (G2's close admitted);
   p. count the open-gap exemption from `carried_by` instead of `carrier_minted` → `TestAPlainMintALaterReadingNamesKeepsItsMint` fails;
   q. revert any one fuzz branch of §III.11 to `Detail == ""` → `TestTheFailedCheckDrivesKeyOnTheState` fails;
   r. let another lens's later backing reading replace a found-against check (`stands` keyed on any seat) → `TestTheLensCloseOfAReadingsGapIsRefusedWhileTheCitationStandsFoundAgainst` and `TestAMigratedFoundAgainstCheckReadiesBlueAlone` (subtest 6, another lens) fail;
   s. make `ReadingGapClasses` ignore `VerifyReads` (outcome's `gap_classes` alone) → `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` rows (a), (b), (d), (e) fail (flags demanded, or a gap minted on a `not_the_work` reading); and, separately, drop the write-path refusal of a carrier on a non-minting reading → its record-layer row fails;
   t. place a `verify`'s gap by `location` (the anchor-stripped sentence) instead of `sentence_of` → `TestAVerifyOfARepeatedSentencePlacesItsGapByTheAnchor` fails (refused as ambiguous); drop the `report_op` `'sentence_of'` arm → the same test fails (the gap's token absent from the render); place the gap's token right after the citation's token instead of at the sentence's end → its clause-cited row (d) and the accept row (f) fail; restore `return nil` on a sentence_of placement failure in the one-gap rule → its second-lens row fails (a second gap minted);
   u. decide the held gap before the transaction (read it in the verb, refuse in the Mint's validation) → `TestTheReadingsHeldGapIsDecidedInsideItsWrite` fails (A's act refused whole);
   v. delete the `"contradiction": "anchor"` row from `valueRenames` → `TestTheMigratedM30RaiseIsDropped` <!-- CHANGED F13b --> and `TestEveryArchivedRunMigrates` fail (decode refusal); replace the raise's `anchor.Kind(ref) == ""` test with "about_kind is anchor" alone → `TestAFindingAboutAnAnchorIsNoRaise` fails (its finding about `C-…`, and its archived `proof`-kind finding about `P-…`, respelled `K-…`, refused as unparsable events.ids).
   w. restore the completeness stamp's guard to `body.GetUrl() != ""` → `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` rows (a), (b), (g), (i) fail (a `verify` carries no stamp: its `refutes` demands gap flags, its backing reading clears the citation); drop `record.VerifyReads(body)` from the footnote decision → row (h) fails (a footnote minted from a `not_the_work` copy);
   y. replay the reading's three events one at a time (no re-driven act) → `TestARecordWrittenAtThisEpochMigratesWithNoRefusal` fails (the capped lens's Mint refused for budget, its Verify refused outside its act); write the act at the Mint's iteration and skip the other two (round-6 shape) → it fails (the plan refused: its `why` names the Verify's old id, which wrote no `newID` entry; the Stale through the Mint translated to the Verify's id); map each member from `MAX(id)` after the act → it fails (the Mint's Stale names the Verify); set `record.Now` once for the act → it fails (`ts` differ from the source's); count the act once in `res.Out` → it fails (`res.In`/`res.Out`); run `place.step` on the Anchor after the write → `TestMigrateChecksAReadingsGapBySentenceOf` fails (the act lands first: the migrated record holds G's Mint, Anchor and Verify where it asserts none); skip `gapplace`'s sentence_of check → `TestMigrateChecksAReadingsGapBySentenceOf` fails (admitted);
   z. locate `ProposalOver`'s span by text (`LocateOnce`) again → `TestAVerifyOfARepeatedSentencePlacesItsGapByTheAnchor` row (e) fails (the verified fix refused as ambiguous); drop the accept's exact-span locate → row (e) fails (the accept refused), and row (i) fails if the edit is recorded without `exact_span` (replay ambiguous);
   aa. read open-ness for the tie-break from Close events instead of the gap view → `TestASupersededReadingGapHandsTheCheckToItsSuccessor`'s bench-ruled branch fails (`AnsweredBy()` names G3, closed by the ruling);
   cc. attach `requireNotEstopped` to the reading's append → `TestAReadingOverAnAppliedPrescriptionMintsItsGap` rows (b) and (c) fail (refused as an estoppel, an estoppel entry logged);
   dd. register the reading verbs' `--new`, `--class` and `--problem` with `mint`'s descriptions → `TestTheReadingVerbsDescribeTheirGapFlagsByTheirSubject` fails (`verify`'s `--new` names `--quote` alone and its `--class` offers `unclassified`); <!-- CHANGED F13 -->
   ee. drop `MintReason` from the reading's Mint → `TestAReadingsGapIsFilledAsAPlainMints` fails (presence differs), and `TestAReadingsGapCarriesItsArgumentWhereverAMintsDoes` fails on `show board --json` and on `report.Assemble`'s gap block and closure index;
   ff. leave `fix_basis` unset on the reading's Mint → both fail (`fix_basis` empty on the board and the work item, no `fix_basis` line in the gap block);
   gg. delete `seatprobe.Read`'s skip of a reading's Mint → `TestAReadingIsAttributedToTheVerbTheSeatTyped` fails (`mint` used by a seat that never typed it); skip every Mint a Verify names, `carrier_minted` unread → its plain-mint row fails; let a `verify --url` reading count under any verb but `verify` → its `--url` row fails; <!-- CHANGED F13 -->
   hh. register `--class` on the reading verbs from `MustEnum("mint", "class")` → `TestTheReadingVerbsDescribeTheirGapFlagsByTheirSubject` fails (every class listed); drop `RequiredUnless` from `--class` → `TestHelpSaysRequiredExactlyWhereOmissionIsRefused` fails ("REFUSED without --class and its help does not say so"); restore a `supports` clean set in the verify probe → the same test fails ("says REQUIRED and the verb ran clean without it");
   ii. print the found-against sentence's event id after a word other than `event` → `TestEveryIDTheVerbPrintsIsTranslated` fails;
   jj. restore `Refuted()`'s outcome-only body (`Outcome == "refutes" || "absent"`) under the new name → `evidenceview_test.go`'s `not_the_work` row and `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` row (b) fail; restore `human()`'s outcome-first switch → row (b) fails ("REFUTES" printed);
   kk. drop `acceptance_check` from `readingGap`, or print the reading's gap tells under `verifyResult`'s note → `TestAReadingsGapIsFilledAsAPlainMints`' GAP 3 row fails;
   ll. leave `EvidenceVerificationJSON.CarriedBy` unset, or delete docket.md's source line → `TestAReadingsGapNamesItsSource` fails;
   mm. restore the `failed check` definition's four sentences → `TestTheEmbeddedRegistryLoads` fails ("the definition is ONE sentence"); delete the `reading` entry → the manual golden and `vocabdoc -check` fail;
   bb. let the successor of a reading's gap inherit the budget exemption (exclude by `answers` instead of `carrier_minted`) → `TestAPlainMintALaterReadingNamesKeepsItsMint`'s successor row fails (the over-cap successor admitted).
   nn. (F13) drop `found_source_gap_classes` from `SUPPORTS` → `TestReadingGapClassesCoversBothForms` and `TestEveryNegativeOutcomeNamesItsClasses` fail; gate `ReadingGapClasses`' found-source arm on `FindsAgainst` → the backing rows fail;
   oo. <!-- CHANGED F13b --> put `url` back on `keyFields` → `TestAFoundSourceReadingKeysOnSourceAndSentence`, `TestOneSourceBacksManySentences` and `TestOneSentenceReadAgainstManySources`' second-sentence rows fail (refused on the url key); add `carried_by` to `keyFields` (F13's key) → `TestOneSentenceReadAgainstManySources` fails (the second url refused, b7's shape);
   pp. key the repeat rule's unanchored arm on `FindsAgainst` again → `TestARepeatedFoundSourceReadingIsRefused` fails (a second backing reading admitted); <!-- CHANGED F13b --> restore the refusal's one text ("finds the source against this sentence") → the same test fails;
   qq. restore `citedSource`'s Verify arm → `TestNoReadingPlacesACitation` fails;
   rr. drop item 7's token removal (or restore `gapplace`'s Verify arm) → `TestTheMigratedRedCitationsAreReadingsOfTheirSources` fails (a refusal, or a label in the render);
   ss. let a backing `verify --url` record without a gap (`found_source_gap_classes` read as empty) → `TestAFoundSourceReadingMintsItsGap` fails;
   tt. delete the `uncitedSentence` guard → `TestABackingFoundSourceReadingOfAMarkedSentenceIsRefused` fails (an `unsourced_claim` minted on a cited sentence);
   uu. drop any one flag group from `verify` → `TestVerifyTakesExactlyOneSubject` fails (the call reaches RunE, or both subjects are recorded). <!-- CHANGED F13 -->
   vv. <!-- CHANGED F13b --> delete `requireNotARetriedReading` → `TestAFoundSourceReadingKeysOnSourceAndSentence` and `TestOneSentenceReadAgainstManySources` fail (the identical retry recorded); drop the stored event from its text → the same tests fail (the refusal names no reading);
   ww. <!-- CHANGED F13b --> restore the act-paired budget admission for every paired Mint (no `anchor` test) → `TestAFoundSourceGapTakesAMint` rows (a), (b) fail (admitted at the cap); restore the open-count exclusion for every `carrier_minted` Verify → its row (d) fails (the open count unchanged); narrow the admission to `--url` instead → its row (c) and `TestAReadingsGapTakesNoMint` fail;
   xx. <!-- CHANGED F13b --> leave `failedcheck.go:319`'s early return on `FindsAgainst` → `TestTheRepeatRuleReachesABackingFoundSourceReading` and `TestARepeatedFoundSourceReadingIsRefused` fail;
   yy. <!-- CHANGED F13b --> skip the token removal in one report-text field (`Retire.claim`) → `TestNoMigratedArchiveHoldsALensPlacedToken` fails on b9's retire (archived event 166);
   zz. <!-- CHANGED F13b --> respell a dropped label's list element to `""` instead of removing it → `TestTheMigratedRedCitationsAreReadingsOfTheirSources` fails (b9 164 and 167's `reopened` is `["F-84381c8e", ""]`); read the re-readings' url and title through the batch's querier → it fails (a refusal: no label column).
   Re-armed by: the function or table each names.
6. migrate on real archives: `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/migrate/ -run 'M30|M32|M23|EveryArchivedRun|M19|AboutAnAnchor|ValueRenames|StringField|BySentenceOf|RedCitations'` (F13: `RedCitations`) <!-- CHANGED F13 --> and `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/ ./internal/cli/ -run 'M32|M30|MigratedFoundAgainst|M23s|RecordWrittenAtThisEpoch|EveryIDTheVerb'` → pass, 0 refusals. Re-armed by: `migrate/**`, `record.proto`.
7. Release fuzz: PR CI only (sharded 4 + tally). Read the tally: new gates non-zero; max gaps per (class, sentence) = 1. NEVER run `FEOV_RELEASE_GATE=1` locally. Re-armed by: any record or flag change (CLAUDE.md requires a full pass before done).
8. **Universe smoke (real engine, cheap configuration).** The client is Claude Code 2.1.305 or later (#1677's floor, held by `sc-doctor` and the setup preflight; an older client is refused at setup, not a finding). `scripts/universe.sh build --dir ~/.claude/scratch/contradiction-is-a-gap/universe`; prove the build carries the change: `grep -rl claim_contradicts_source ~/.claude/scratch/contradiction-is-a-gap/universe/config/plugins/cache/special-circumstances/frank-exchange-of-views/` → at least the installed agents and the `feov-record` binary; and `scripts/universe.sh doctor` → schema epoch 55 (or the rebased number). Then
   `MODEL=haiku JUDGMENT_MODEL=haiku LANES=1 scripts/universe.sh --dir ~/.claude/scratch/contradiction-is-a-gap/universe run "Is the Great Wall of China visible from low Earth orbit with the naked eye?" --mint-budget 3 --k-max 2`
   (a topic whose popular sources misstate the claim, so an evidence lens is likely to read a refuting source). Observe with `universe.sh watch`; afterwards, on the run's record (read-only `file:` URI):
   `SELECT v.outcome, v.anchor, v.carried_by, m.class, m.class_material, m.mint_reason, m.fix_basis FROM verify v LEFT JOIN mint m ON m.gap_id = v.carried_by WHERE COALESCE(v.source_completeness, '') <> 'not_the_work' AND ((v.anchor IS NOT NULL AND v.anchor <> '' AND v.outcome IN ('refutes','absent')) OR (COALESCE(v.anchor, '') = '' AND v.outcome IN ('supports','supports_with_bridge','weak','refutes')));`
   (the readings that mint: they read the work, and either find a citation's source against its sentence or read a source the lens found that backs or refutes the sentence; a found-source `absent` mints nothing, F5).
   Expected: every row has a `carried_by`, class in the form's and outcome's admitted set, `class_material = always` for every class but `unsourced_claim` (by grade, O1), a non-blank `mint_reason` and a `fix_basis` of `proposed` or `verified`; any second reading of one sentence shares a `carried_by`. Also: `SELECT COUNT(*) FROM verify WHERE label IS NOT NULL` fails (no such column — the field is reserved), and no `anchor` row's id is a `C-` without a `cite` row of that label. The converse —
   `SELECT COUNT(*) FROM verify WHERE carried_by IS NOT NULL AND (COALESCE(source_completeness, '') = 'not_the_work' OR outcome = 'unreachable' OR (anchor IS NOT NULL AND anchor <> '' AND outcome NOT IN ('refutes','absent')) OR (COALESCE(anchor, '') = '' AND outcome = 'absent'));` <!-- CHANGED F13 -->
   — is 0. Every row is of an admitted act: refused acts write nothing. If no negative reading occurred, the smoke is reported **NOT MEASURED on this path** (not green), and check 9 carries it. Archive the run to `run-archive/`.
9. **Driveable check on real data.** m30 as archived holds its judge's `outcome` (events.id 467) and the chair's verdict (463), and `requireOpenToSeats` refuses every seat act after an outcome, so the archive is CUT before it is driven: copy m30's `records/record.db` to `~/.claude/scratch/contradiction-is-a-gap/m30/`, and in the copy delete every row with `event_id >= 463` from each table holding an `event_id` column, then `DELETE FROM events WHERE id >= 463` (a scratch step; nothing in the repository changes). Migrate the cut copy (`TestEveryArchivedRunMigrates`' path, or `feov-record migrate`), and confirm a seat act is admitted on it (`register` as `red-lens-evidence`, `show board`) before the drive — a refusal there means the cut is wrong, not the change. Then on the migrated copy as `red-lens-evidence`: `feov-record verify --anchor C-4f4b8cfa --as refutes --class claim_contradicts_source …` → a gap lands (or names G-dc877265 if it is open with that class); a second lens's identical reading → names the same gap, no mint; `close` of that gap → refused while the reading stands; a third lens's `--as supports` reading → recorded, and the close still refused (a disagreement clears nothing); the reader's own `--as supports` re-read a sitting later → the close admitted; `show evidence` has no `unanswered_contradictions` key, and its verification row for the reading
   carries `source_completeness` `not_asked` (the archive holds no fetched-source cache, so the stamp
   says the copy was never asked — the stamp `verify` now writes, read off a real record); `show board --json`
   lists the gap with the `--reason` as its `mint_reason` and its `fix_basis`. <!-- CHANGED F13b --> Then, on the
   same cut and migrated copy, the found-source form on real data: pick a sentence of the migrated m30
   render that holds no citation or computation token (named in the PR with its section), and as a lens
   under its mint cap (`show work` states the standing) run `feov-record verify --url <a url the run's
   cache holds> --title … --quote "<that sentence>" --as supports --class unsourced_claim …` → an
   `unsourced_claim` gap lands on the sentence, no `C-` anchor, no Bibliography line, and the lens's open
   count rises by one (F6 amended); the same lens's reading of a SECOND url against that sentence in the
   same sitting → recorded, naming that gap, no second gap (gblock's key ruling); the first call
   repeated → refused, naming the first reading's `events.id`; the same call by a lens AT its cap →
   refused for budget, nothing recorded. Re-armed by: the reading act, the ledger, the close gate,
   <!-- CHANGED F13b --> the retry rule and the budget's anchor test.
10. The audit loop ended at round 8 (gblock, 2026-11-01); the plan is approved and carries round 8's gaps as the required tests of §III.15. BEFORE merge: an INDEPENDENT review of the implementation — a reviewer who did not implement it reads the diff against this plan, re-runs §V.4's sweeps, and confirms each test of §III.15 exists, fails on its §V.5 partner and passes. The merge waits on that review and on a full PR CI fuzz pass (§V.7).

## Forks — all RULED (gblock, 2026-10-30; F10, F11 2026-10-31; F12, #1657, audit loop 2026-11-01)

Each ruling is recorded here; the alternatives it rejected are deleted from the design.

- **F1 — Class (RULED).** A NEW class `claim_contradicts_source` for `refutes`; `citation_figure_misattribution`
  narrowed to `absent`. The lens picks among the classes its outcome allows, declared once on
  `SourceOutcome` (`gap_classes`): `refutes` → `claim_contradicts_source` | `within_source_condition_misattribution`;
  `absent` → `citation_figure_misattribution` | `within_source_condition_misattribution` (§III.1). Cost
  carried as R3.
- **F2 — Judgment (RULED).** The reading act (`verify`, both forms since F13 <!-- CHANGED F13 -->) takes the mint's judgment flags,
  required when the reading finds against — it reads the work and its outcome has `gap_classes`
  (`ReadingGapClasses`, §III.2b) — and refused otherwise (§III.2). F11 extends it to `--reason`.
- **F3 — Duplicate (RULED).** The reading is kept and the mint refused; `Verify.carried_by` names the standing
  gap (§III.1, §III.2).
- **F4 — PASS hold (RULED).** Found-against is a third failed-check state. It holds the PASS until the
  citation holds — by its reader's own later backing reading, as the ledger replaces only a seat's own
  check — the sentence is edited, or the citation is retired. The lens's close of its gap is refused while
  it stands. No bench ruling clears the check (§III.3; R4; refined by F8).
- **F8 — A bench ruling on the carrying gap (RULED, gblock 2026-10-30: "Gap closes, PASS still held").**
  A closing docket ruling closes a gap that carries a found-against check, through the bench's existing
  `closure_act` arm; the failed check alone keeps holding the PASS until it is cleared by the reader's own
  later backing reading, blue's edit, or blue's retire — the same as today's unread and unsound checks.
  F4's intent, no PASS while the citation is refuted, holds through the check, not through the gap. Every
  sentence that said the gap does not close under a ruling is rewritten (§III.3, §III.7, R4); test
  `TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld` (§III.10).
- **F9 — Placement of a reading's gap (RULED, gblock 2026-10-30: "I think if only text is provided
  out [it] should be refused but it's fair to point at an anchor using the right flag").** A reading
  that names its citation by the reading act's anchor flag (`verify --anchor`) places its gap at the end
  of the sentence that anchor sits in, located by the anchor's identity (`Anchor.sentence_of`), never by
  searching for the sentence's text. Placement from text alone (`verify --url --quote`, F13) that is
  ambiguous stays refused (§III.2a). <!-- CHANGED F13 -->
- **`--new` over a sentence held twice (RULED, gblock 2026-10-30: locate by gap anchor).** Blue's accept
  and `ProposalOver` locate the prescribed span by the gap's anchor identity; the accept's edit is
  recorded with `exact_span`. A reading's `--new` over a repeated sentence is verified and accepted at the
  cited occurrence; nothing refuses it for the repeat (§III.2a).
- **GAP 1 — A reading of a `not_the_work` copy (RULED, gblock 2026-10-30: "verify stamps it").** `verify`
  stamps `source_completeness` from the cited url's cache entry; every reading of a `not_the_work` copy
  is unread — mints nothing (no gap, no footnote), holds the PASS through the citation's `unread` check,
  owes that check's gap — a backing one included (§III.2, §III.2b).
- **GAP 2 — Archived corroboration readings (RULED, gblock 2026-10-30: "accept the loss, stated").** An
  archived negative corroboration holds no PASS after the change; the four lost readings are named per
  archive, and the m23 test asserts the migrated state (§III.9). F13 adds the archived BACKING
  corroborations: translated, their labels and red citations removed (§III.9 item 7). <!-- CHANGED F13 -->
- **F10 — Estoppel and a reading's gap (RULED, gblock 2026-10-31: "Exempt, stated").** A reading's gap
  is exempt from estoppel: a source's refutation is new evidence, not red relitigating its own
  prescription. The gap opens, and the failed check holds the PASS (a `verify --url` refuting gap, material by
  its class, holds it while open). The reading verbs attach no `requireNotEstopped`; which of `lens
  mint`'s other checks the reading's Mint meets is §III.2's guard table; the exemption and its reason
  are stated wherever estoppel is taught or enforced (§III.2's sweep); test
  `TestAReadingOverAnAppliedPrescriptionMintsItsGap` (§III.10).
- **F11 — The reading's gap's argument (RULED, gblock 2026-10-31: "Reading takes --reason").** The reading
  act takes `--reason` exactly as plain mint does, required when the reading finds against, and
  `fix_basis` defaults to `proposed` as for mint: no copied text and no new reader path. The verbs' one
  `--reason` is read once and written at the act to `Mint.mint_reason`, as `mint` writes it, and to
  `Verify.text`, which is required and cannot be left blank; nothing derives one field from the other on
  a read. Every field `newMint` sets has a disposition (§III.2's field table). Tests
  `TestAReadingsGapIsFilledAsAPlainMints`, `TestAReadingsGapCarriesItsArgumentWhereverAMintsDoes` (§III.10).
- **F5 — Corroboration `absent` (RULED).** Records the reading and mints no gap (§III.1, §III.10); since F13, a `verify --url` reading's `absent`. <!-- CHANGED F13 -->
- **F6 — Budget (RULED; amended F13b).** A gap a reading of a CITATION mints (`verify --anchor`) is
  outside the mint budget (§III.5). <!-- CHANGED F13b --> **Amended (gblock, 2026-11-03; the option wording is
  the lead session's):** any gap a `verify --url` reading mints counts against the lens's mint budget —
  taken at the act, counted while open; the anchor readings' gaps stay outside it as F6 states. Ruled
  because F6, decided when only a reading that finds against minted, would otherwise have exempted
  `unsourced_claim`, a lens's most common class, whenever the lens attached a backing source. Tests
  `TestAFoundSourceGapTakesAMint`, `TestAReadingsGapTakesNoMint`; the replay measure in §III.9 item 7.
- **F7 — PR #1642 (RULED).** <!-- FINDING-DEPENDENT --> (its `open_by_found_by` reads `found_by`; §III.11) Closed as superseded; its four diagnostic tally fields — `final_blockers`,
  `final_parties`, `open_by_found_by` (re-keyed `reading | finding | unsound_computation | none`: #1642's
  `contradiction` key becomes `reading`, and its `unsound_proof` key keeps its meaning under #1644's name), `mints_by_epoch` /
  `closes_by_epoch` — are carried into this PR (§III.11).
- **§III.9 — Archived carrier-less checks (RULED).** Cleared by the reader's own backing re-read, an edit or a retire. migrate does
  NOT fill `carried_by` from the `found_by` chain; the residue is stated in §III.9.
- **§III.14 — Delivery (RULED).** ONE PR, implemented one phase per fresh agent.
- **F12 — The link from a reading's gap to its source (RULED, gblock 2026-11-01: "Print the link").**
  `found_by` is not set on a reading's gap (the round-8 verdict: "found_by goes; a corroboration's gap
  has no link to its reading"; the bullet's former "goes on" misread it, §III.2c) <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->; `carried_by` is printed on `show evidence`'s rows, and docket.md
  names the source of each carrying reading (§III.2c); test `TestAReadingsGapNamesItsSource`.
- **#1657 — the refuting reading (RULED, gblock 2026-11-01, picking the two sessions' joint
  recommendation).** This plan governs a refuting reading; #1657 Part 1a″'s L4 is dropped;
  contradictions are deleted everywhere, corroborations included; #1657 Part 1a′ continues; this plan
  lands FIRST, then 1a′ (§III.16).
- **Audit loop (RULED, gblock 2026-11-01: "Revise once, then implement").** Round 8's revision is the
  last; no further audit rounds; implementation is approved, carrying round 8's gaps as required tests
  (§III.15); independent review after implementation (§V.10).
- **F13 — `corroborate` dropped (RULED, gblock 2026-11-03, in the #1657 session; quoted in §III.2d).**
  A reading of a source the lens found mints `unsourced_claim` (backing outcome, uncited sentence) or
  `claim_contradicts_source` (`refutes`) in the same act; no reading places a citation, label or
  footnote; blue answers with a cite and a lens verifies it. The verb shape — ONE verb, `verify`, exactly
  one of `--anchor` and `--url`, `corroborate` deleted with no alias — is this revision's decision
  (§III.2d), as is the refusal of a backing `--url` reading on a marked sentence (<!-- CHANGED F13b --> a
  single right answer under the rulings, not a fork: the targeted round; residue quadratic `c-95969800`,
  b9 `c-d49a32c4`); the archived labels are translated (§III.9 item 7; gblock picked "drop them, state
  it", relayed 2026-11-03). <!-- CHANGED F13 -->
  - <!-- CHANGED F13b --> **O1 — CLOSED, not a fork** (the targeted round). Materiality is the class's and the
    ruled option names `unsourced_claim`; a backing found-source gap holds the PASS where its grades make
    it material. §III.16 tells 1a′ its "every reading gap is material by its class" is false for
    `unsourced_claim`.
  - <!-- CHANGED F13b --> **The key (RULED, gblock 2026-11-03).** A found-source reading keys on its source and
    its sentence: one seat's readings in one sitting that differ in url or in sentence are each
    recorded; an identical retry is refused naming the stored reading (§III.1). F13's `carried_by` key is
    not built.
  - <!-- CHANGED F13b --> **Scope (RULED, gblock 2026-11-03): F13 stays in #1694, not split.** Phase 1 already
    built the corroboration arms F13 deletes; splitting would build them out through phases 2-5 and then
    delete them (§III.2d).
- **#1695 — findings fold into gaps (RULED by gblock; relayed by the peer session
  bridge-cse-01s7d2sa1pneo9oqzevmlwdg-4f, 2026-11-03).** A separate plan, after this one and 1a′. This
  plan keeps only the finding rows that delete (the about-kind, the raise, its credit); migrate writes the
  five archived raises as no event and removes the two `found_by` entries naming one (§III.9 items 1, 3);
  the rest is §III.16's pointer. <!-- CHANGED F13 --> <!-- FINDING-DEPENDENT -->
