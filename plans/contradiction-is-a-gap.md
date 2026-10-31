# A source reading against the report is a gap — the contradiction is deleted

Written against `origin/main` **`0c57758e8`** (EventSchema 52: #1660, a withheld PASS is heard, took 51;
#1663, the report is the lanes' parts, took 52). #1663 made lanes their own role, `lane` (agent
`blue-lane`), and renamed `base_ingest` to `report_part`, so the report is the `report_op` stream folded
from the empty text; #1659 made an exchange begin with blue's act (`lensTurnOn`). Every file:line, symbol,
census entry and sweep hit below is read at that commit. The references of the round-7 text (read at
`a31a7b06d`) were carried by a `git diff -U0` line map (`~/.claude/scratch/plan1647-r8/mapref2.py`); every
reference the map could not carry unchanged — a line those PRs rewrote — was re-read by hand, and every
sweep and census was re-run at `0c57758e8` (raw output beside the script, `new-*.txt`, `census-new.txt`,
`phrase-new.txt`). #1644 renamed proof to computation: a `P-` id is a `K-` id, `proof_id` is
`computation_id`, and every sentence below that speaks of a proof's failed check speaks of a
computation's. Implements gblock's ruling of 2026-10-30; every fork is RULED (gblock, 2026-10-30 and
2026-10-31, §Forks). This plan takes EventSchema 53, re-derived at the implementation rebase.

> **Owner ruling (gblock, 2026-10-30, his words):** "why isn't the contradiction recorded as a gap in
> the first place like red would for any other problem it finds? We're letting it mint a weird
> placeholder." And: "Our standard behaviour is reject a duplicate, not dedup — why are you treating
> contradictions as unique?" He approved planning the redesign: a `refutes` or `absent` source reading
> is recorded as a gap of a real class, like any other problem red finds, and the contradiction
> concept is deleted.

## I. Summary & Goals

**Objective.** Today a lens's `verify`/`corroborate` reading whose outcome is `refutes` or `absent`
opens a **contradiction** (`record/citationid.go:522` `contradictionsOf`): a record object with its own
id (the reading's `events.id`) that a finding must *raise* (`--about-kind contradiction --about <id>`)
and that only the gap minted from that finding (`found_by`) *answers*. Until then it is a PASS blocker
of its own kind (`passblockers.go:35`), a work-list debt of the reader (`sitting.go:374-439`), a
readiness reason (`dispatch.go:547-553`), and it is exempt from the one-gap-per-class rule
(`onegapperclass.go:68`) and from the mint budget (`mintbudget.go:301-305`). A reading after the raise
opens a *new* contradiction (`citationid.go:525-529`), so N readings of one sentence yield N
contradictions, each owed its own exempt gap: in release-fuzz run 40872228665 (PR #1642's finding),
22 of 45 mints per run answered a contradiction and all 37 drawn runs ended CEILING. m32 shows the
same shape on real seats: four lenses refuted one claim on `C-a36dce3f`, four findings raised four
contradictions, one gap answered one of them.

Throughout this plan, a **negative reading** is a reading that finds against: it reads the work
(`VerifyReads` — not `unreachable`, not any outcome of a copy the tool stamped `not_the_work`) and its
outcome is `refutes` or `absent` (`record.FindsAgainst`, §III.1). A verdict that reads nothing is never
a negative reading, whatever its `--as`, and mints nothing (§III.2b).

After this change a negative reading **is** the gap: the reading act mints a gap of a real class on the
sentence. Its Mint meets every write-path rule a mint meets (class, subject, grades, check kind, one-gap
rule, report voice; the budget through F6's act-paired exemption) and, of `lens mint`'s verb-level
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

1. A `verify` or `corroborate` with `--as refutes` (or `verify --as absent`) of a copy that is the work
   writes the reading AND an open gap in one transaction; the reading's JSON result names the gap
   (`gap.id`, `gap.minted: true`). A `verify`'s gap stands at the end of the sentence its `--anchor`
   token sits in, placed by that token's identity, also where the report holds the sentence's words
   twice, and blue's `--accept` of its verified fix edits that occurrence. A reading that reads
   nothing — `unreachable`, or ANY outcome of a copy the tool stamped `not_the_work`, a backing one
   included — writes no gap, no footnote and no `carried_by`, and leaves a `verify`'s citation on the
   `unread` failed check; its gap flags are refused. Every reading, `verify`'s included, carries the
   `source_completeness` of the copy it read. A reading's gap carries its argument and its `fix_basis`
   wherever a plain mint's appear: `show board --json`, the blue work item, the gap view, and the
   assembled report's gap block and closure index.
2. A second negative reading of the same sentence with the same class, while that gap is open, writes the
   reading, writes no gap, and names the standing gap (`gap.minted: false`). Gaps per (class, sentence)
   open at once: **≤ 1**, asserted over every fuzz run by the sweep tally.
3. The identifiers grep of §V.4 returns **0** lines outside migrate's archived-word tables; the word grep
   returns only lines §III.12 lists as KEEP (ordinary English, the two self-contradiction classes,
   historical plans).
4. `migrate` brings every archive forward with **0** refusals (`TestEveryArchivedRunMigrates`), and the
   migrated m23, m30, m32 and sqlite-schema-source hold the dispositions §III.9 states; a record written at
   this epoch — a capped lens's refuting reading in it — migrates forward with **0** refusals.
5. `go -C scripts run ./check`: 0 failed. PR CI's release fuzz (sharded, CI only): 0 failed, with the
   new gates of §III.11 non-zero.
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
  `scripts/schemagen` from `plugins/frank-exchange-of-views/requirements.json` (`"eventSchema": 52`).
- **Epoch.** `origin/main` is at EventSchema 52 (#1660 took 51, #1663 took 52). This plan takes
  **EventSchema 53**, and re-derives the number from `origin/main` at the implementation rebase; the
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
  outcome admits (F1); `Verify` gains `carried_by` (references `mint.gap_id`) and `carrier_minted`; `Anchor` gains
  `sentence_of` (a marker placed at the end of the sentence another anchor's token sits in, found by
  that token's identity, §III.2a); `Mint` is unchanged.
- **Existing mechanisms reused, not invented:** the one-gap rule (`onegapperclass.go`, eb0c5a99f), the
  failed-check state and its close gate (`failedcheck.go`, `requireFailedCheckAnswered`), the class's
  `material_default` stamped at write (`Mint.class_material`, `recordpb/gapclass.go:18`), the
  placed-append that already writes two events in one act (`seat.AppendPlaced`, used by corroborate's
  footnote), the one-gap rule's `heldGap` (`onegapperclass.go:100`, the open gap it names), the
  exact-span edit that replay locates byte for byte (`BlueEdit.exact_span`, `reportproj/render.go:89-97`),
  and replay-time reference translation in migrate (`migrate/replay.go:395-406`).
- **Materiality** is the class's: `material_default` on the `GapClass` value, stamped by the write path
  onto `Mint.class_material`, never set by a seat. Every class this plan admits for a reading
  (`claim_contradicts_source`, `citation_figure_misattribution`, `within_source_condition_misattribution`)
  is `CLASS_MATERIAL_ALWAYS`, so every reading's gap holds the PASS whatever it is graded; grades still
  order it (mass). No seat picks materiality.
- **Constraints.** No backwards compatibility (CLAUDE.md change discipline): no alias for the about-kind,
  no tolerance path for an archived `contradiction` word in live code, no "formerly" prose. Archived
  records are brought forward by one `migrate` translation. The release gate
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
| 1a. Which readings mint | A reading that FINDS AGAINST: it reads the work (`VerifyReads`: not `unreachable`, not a `not_the_work` copy) AND its outcome has `gap_classes`, less a corroboration's `absent` (F5). One predicate, `record.ReadingGapClasses`, decided after the tool stamps the copy's completeness; every site in §III.2b's sweep calls it, none decides by outcome alone | — |
| 1a′. The completeness stamp | `verify` stamps `source_completeness` from the run's cache entry for the cited url, as `corroborate` does for its `--url`; today it stamps nothing (`verify.go:494`, guarded by `body.GetUrl()`), so a `verify` of a `not_the_work` copy counts as a reading. EVERY reading of a `not_the_work` copy is unread — mints no gap, no footnote, holds the PASS through the citation's `unread` check and owes that check's gap — a BACKING one included | **GAP 1** RULED |
| 1b. Who names class, quote, problem | Quote = the reading's claim (the tool already reads it for `verify`, by its anchor; `corroborate`'s `--quote`). Class = the lens's `--class`, refused outside the outcome's admitted set. Problem, check, check-kind, severity, likelihood, impact = the lens's flags on the reading act, required on a reading that finds against (1a) and refused on any other; fix, new text and complexity optional as on `mint`, and refused on any other | **F2** RULED |
| 1b′. The gap's argument and its other fields | The reading act takes `--reason` as plain mint does: its text is the gap's `mint_reason` as well as the reading's `Verify.text`, and required when the reading finds against. `fix_basis` defaults to `proposed`, as for mint. Every field `lens mint` sets has a disposition (§III.2's field table); no reader path is added | **F11** RULED |
| 1c. Where the gap stands | `verify`: at the end of the sentence its `--anchor` token sits in — where every gap's token stands, at its quote's end — found by that token's identity (`Anchor.sentence_of`), never by searching the sentence's text. `corroborate` names no anchor: its gap stands at its `--quote`, and a quote the report holds more than once is refused, as every text placement is | **F9** RULED |
| 1c′. `--new` on a reading | Blue's `--accept` and `record.ProposalOver` locate a gap's prescribed span by the gap's anchor identity — the occurrence of its location whose anchor run holds its token — and the accept's edit is recorded with `exact_span`, located byte for byte on replay. So `--new` on a `verify` over a sentence held twice is verified and accepted at the cited occurrence; nothing refuses it for the repeat | **Fork** RULED (locate by gap anchor) |
| 1d. Materiality | The class's `material_default` (both `always`), stamped at the write as for every gap | — |
| 1e. Estoppel | A reading's gap is EXEMPT: the reading verbs attach no `requireNotEstopped`, also where the sentence is text blue applied verbatim from a red `--new`. The gap opens; a `verify`'s failed check, and a `corroborate`'s material gap, hold the PASS. The exemption and its reason are stated wherever estoppel is taught or enforced (§III.2) | **F10** RULED |
| 2. Duplicate | The reading is kept and the one-gap rule refuses the *mint*; `Verify.carried_by` names the standing gap | **F3** RULED |
| 3. Backing re-check | Found-against is a third **failed-check** state beside unread/unsound. It holds the PASS until its reader's own later reading backs the sentence, blue edits the sentence, or blue retires the citation; another lens's backing reading clears nothing. The lens's close of its gap is refused while it stands (`requireFailedCheckAnswered`); a closing docket ruling closes the gap and the check still holds the PASS. Once it clears, the minting lens closes the gap with its ordinary close. No new answer kind | **F4**, **F8** RULED |
| Corroboration `absent` | Records the reading; mints no gap | **F5** RULED |
| Archived corroboration readings | An archived negative corroboration has no anchor, so after the change it holds no failed check and no PASS: the loss is accepted and stated per archive (§III.9) | **GAP 2** RULED |
| Mint budget | A gap a reading mints is outside the mint budget: it takes no mint and does not count while open; the "answers its own raise" exemption is deleted | **F6** RULED |
| 4. Delete | Everything in the census (§III.12) | — |
| 5. migrate | One translation, §III.9 | — |
| 6. Epoch | 53, re-derived at rebase | — |
| 7. PR #1642 | Closed as superseded; its four diagnostic tally fields are carried into this PR (§III.11) | **F7** RULED |
| 8. Delivery | ONE PR, implemented one phase per fresh agent (§III.14) | RULED |

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
  distinguisher stops saying "says something else". Name grep (CLAUDE.md naming rule) run at `0c57758e8`:
  `contradicts_source` has 0 hits; value 31 is free (`GAP_CLASS_UNCLASSIFIED = 30` is the last); the sibling `claim_contradicts_own_record` is the parallel the name
  follows (report vs record → report vs source).
- `SourceOutcome`: [NEW] facet `(gap_classes)` on `REFUTES` (`claim_contradicts_source`,
  `within_source_condition_misattribution`) and `ABSENT` (`citation_figure_misattribution`,
  `within_source_condition_misattribution`); absent on every other value. The facet says which classes
  an OUTCOME admits; it does not say a reading mints, because an outcome of a copy that is not the work
  reads nothing (`VerifyReads`, `anchorchecks.go:284-289`, on the `source_completeness` the tool stamps
  from the run's cache — on every reading once §III.2 makes `verify` stamp it; today only a
  corroboration carries it, `verify.go:494`).
  So the record declares two predicates over the stored `Verify`, and no site decides from `outcome`
  alone:
  - [NEW] `record.FindsAgainst(v) bool` = `VerifyReads(v)` and the outcome has `gap_classes`. It
    replaces `contradicts()` (`citationid.go:704`) and is the ledger's found-against test
    (`anchorchecks.go:260-261`: `holds = reads && !FindsAgainst`, found against = `FindsAgainst`), so a
    value added to the enum cannot fall between the reading rule and the anchor ledger. Name grep at
    `0c57758e8`: `FindsAgainst` 0 hits; it extends the ledger's own term (`stateFoundAgainst`,
    `CitationsFoundAgainst`) in the same sense.
  - [NEW] `record.ReadingGapClasses(v) []recordpb.GapClass` = the classes this reading's gap may take:
    nil unless `FindsAgainst(v)`; nil for a corroboration's `absent` (no `anchor`, F5); else the
    outcome's `gap_classes`. "This reading mints" is `len(ReadingGapClasses(v)) > 0`, everywhere. Name
    grep at `0c57758e8`: `ReadingGapClasses` and `gap_classes` 0 hits (`archivedGapClasses` in
    `migrate/gapclass.go` is the archived-slug table, another sense). The `REFUTES`/
  `ABSENT` `(means)` lose "the strongest finding this verb can carry" in favour of "mints a gap".
  `within_source_condition_misattribution` is admitted because both archived answering gaps (m30
  `G-dc877265`, m32 `G-f417112d`) were minted in it by the lens that read the source.
- `Verify`: [NEW] `optional string carried_by = 20 [(sql) = { references: "mint.gap_id" }]` and
  `optional bool carrier_minted = 21` (field numbers 1-19 are taken; 18 is `copy_substitution`).
  `carried_by` is the gap that carries this reading against its sentence: the one the reading minted, or
  the open gap of the reading's class on that sentence that refused the mint. `carrier_minted` is true
  when this reading's own act minted it, false when it names a standing gap. Both are set by the tool,
  on a reading whose `ReadingGapClasses` is non-empty only (`seat.Supplies`, as `mint.go:288` declares the Mint's `gap_id`), so no
  flag sets them.
  - **Why not `gap_id`.** The idempotency key is matched by field NAME (`record.go:483` `keyFields`,
    `"gap_id"` first, ahead of `anchor` and `url`): a Verify carrying a field named `gap_id` would key
    on the gap instead of its citation or source, and `flags/names.go:402` would print it as `--id`, a
    flag the reading verbs do not have. Name grep run at `0c57758e8`: `carried_by` and `carrier_minted`
    have 0 hits. The names' other readers are swept in §III.12 ("Readers of the new Verify fields").
  - The write-path rules, in `record.go`'s Verify case, not table CHECKs (archived negative readings
    carry neither field, §III.9.2, and replay (`Migrating`) admits them as every other new write rule
    does — and a record written at this epoch replays its reading act whole, §III.9 item 6), each keyed on `ReadingGapClasses` of the body as stamped — never on `outcome` alone: a
    reading whose `ReadingGapClasses` is non-empty names its gap; a reading whose `ReadingGapClasses` is
    empty — a backing reading, `unreachable`, ANY outcome of a `not_the_work` copy, a corroboration's
    `absent` — carrying either field is refused ("this reading reads nothing of the work, or backs the
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
  not hold. Name grep at `0c57758e8`: `sentence_of`/`SentenceOf` 0 hits.
- migrate's field lists (`migrate/remap.go:108-144`, gated by `TestEveryStringFieldIsInOneList`):
  `Verify.carried_by` and `Anchor.sentence_of` are ids → `idFields` (an archived id equal to the
  value is respelled, as `Verify.anchor` and `Anchor.id` are); `carrier_minted` is a bool and is in no
  list (the lists class string fields only).
- `Mint`: the `mint_reason` comment (`record.proto:968-978`, "`lens mint` stores --reason here and
  nowhere else") [REWRITE]: `lens mint` and a reading that mints store --reason here. Comments at `record.proto:904-907` [MODIFY] (no kind is refused). No field change.
- Regenerate: `go -C scripts run ./protogen`, `go -C scripts run ./schemagen`; `requirements.json`
  `eventSchema` 52 → 53.

### III.2 The reading act — `internal/cli/lens/verify.go` [MODIFY] (contract: CLI flag set; prompt a seat is told)

- **Every reading stamps the completeness of the copy it read (GAP 1, RULED: "verify stamps it").**
  Today the stamp (`verify.go:491-500`) runs only where `body.GetUrl() != ""`, and `newVerify`
  (`verify.go:97`) builds `Verify{Anchor}` with no url — the source is `citedURL`'s, passed beside the
  body as `source` — so a `verify` never carries `source_completeness`, `VerifyReads` reads its
  `not_the_work` copy as a reading, and `verify --as refutes` of a sign-in page would mint. [MODIFY]: the
  stamp keys on `source` (the cited url's cache entry for `verify`, `--url`'s for `corroborate`, the one
  `held` entry `fetchcache.Lookup` already read at `:366`), never on the body's url, and runs before the
  footnote decision, so every decision below reads the stamped body. [MODIFY] the footnote decision
  (`verify.go:426`, `mayCite == cites && backsTheClaim(...)`) adds `record.VerifyReads(body)`: a backing
  corroboration of a `not_the_work` copy reads nothing, so it mints no label, no footnote and no Anchor —
  a footnote whose own check reads nothing would be a citation born unread — and records the reading
  alone. A backing `verify` of such a copy leaves its citation on the `unread` failed check, owned by
  its reader, who owes that check's gap about the anchor (the ledger already reads it so,
  `anchorchecks.go:259-266`, once the stamp is there). `backsTheClaim` (`:346-350`) keeps its
  outcome-only meaning — which outcomes back a sentence — and its doc comment (`:336-345`, "They owe a
  finding instead") [REWRITE]: a reading against the sentence mints its gap. Tested by
  `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` rows (g)-(i) (§III.10); partner §V.5.w.
- `verify` and `corroborate` gain the mint's judgment flags, the same `flags.*` names and descriptions
  `lens mint` registers (`mint.go:265-280`): `--class`, `--problem`, `--fix`, `--new`, `--check`,
  `--check-kind`, `--severity`, `--likelihood`, `--impact`, `--complexity`. Where `ReadingGapClasses` of
  the body is non-empty, the seven `mint` requires — `--class`, `--problem`, `--check`, `--check-kind`,
  `--severity`, `--likelihood`, `--impact` — are required, each refusal naming its flag, and `--fix`,
  `--new` and `--complexity` are optional, as on `mint`. Where it is empty, all ten are refused (§III.1) — decided in
  `writeVerify` AFTER the copy's completeness is stamped (`verify.go:491-500`, keyed on `source` above), never from `--as` at
  flag parse. The refusal names why nothing is minted: "a reading that backs the sentence mints
  nothing"; "this copy is not the work (`not_the_work`), so the reading reads nothing of it and mints
  no gap — the citation stands unread, and the gap owed is about its anchor"; "an unreachable source
  was not read and mints no gap"; "a corroboration's `absent` records the reading and mints no gap".
  `--class` is refused outside `ReadingGapClasses`, and the refusal lists them. One flag registration
  helper shared by `mint` and the reading verbs, so the two cannot drift — and it registers each flag
  with the description true ON THAT VERB'S PAGE. [NEW] `gapFlags` in `cli/lens/mint.go` holds the four
  `flags.GradeValue`s, `register(c, subject)` registers the ten flags, and `fill(cmd, text)` builds the
  body (the field table below), so one function fills a plain mint's Mint and a reading's. Name grep at
  `0c57758e8`: `gapFlags` 0 hits. Swept: the descriptions `mint.go:265-281` gives these
  flags, each read for a word the reading verbs do not have or a duty they do not carry:
  - `--new` ("the exact text --quote should become", `:274`) names `--quote`, which `verify` does not
    register — `TestEveryFlagAHelpPageNamesIsOnThatPage` (`cli/deadpath_test.go:356`) fails it. The
    helper takes the verb's SUBJECT phrase: `mint` and `corroborate` "--quote"; `verify` "the sentence
    your --anchor's token sits in";
  - `--class` ("unclassified only when none fits, and then --reason names the nearest value", `:265-266`):
    no reading admits `unclassified`. On the reading verbs: "the KIND of defect this reading finds, from
    the values below". `enumhelp.Flag` (`enumhelp.go:55`) renders and parses the `record.EnumField` it is
    handed, and `record.MustEnum` reads `EnumFields`, keyed by event type and payload key — `Verify` has
    no class field, so `MustEnum("verify", "class")` panics, and a row there would declare a field the
    body does not hold. [NEW] `record.ReadingClassField(corroboration bool) EnumField`: `MustEnum("mint",
    "class")` (its `Key`, `Flag`, `Why`) with `Values` narrowed to the union of the `gap_classes` of the
    outcomes that verb mints on, read off the `SourceOutcome` descriptor (`verify`: three; `corroborate`:
    `refutes`' two, since its `absent` mints nothing, F5). `enumhelp` is unchanged, so the parse refuses a
    class outside the union and the page lists it; `ReadingGapClasses` then narrows to the outcome's.
    `TestEverySetShapedFlagIsEitherDeclaredOrExempt` (`cli/enum_help_test.go:200`) walks the site
    `verify --class`, which is in no `EnumFields` row: it enters `enforcedElsewhere` with the reason
    "the reading's Mint's class, from `record.ReadingClassField`; a Verify has no class field". Name grep
    at `0c57758e8`: `ReadingClassField` 0 hits;
  - `--problem` ("Your argument for the gap goes in --reason, which the report never prints", `:272`): the
    first clause is true on the reading verbs, whose `--reason` is that argument (below). They stop at
    "goes in --reason": the assembled report's gap block prints the argument as "minted because"
    (`assemble.go:1124`), so the second clause is not repeated on a page that teaches a reading. `mint`'s
    own text, and `mint.md:23` and its `ReasonIs` row, which say the same, are byte-unchanged: a mismatch
    this plan finds and does not widen;
  - `--fix`, `--check`, `--check-kind`, `--severity`, `--likelihood`, `--impact`, `--complexity`: true on
    every verb, one text.
  **The conditional requirement is marked.** `--class` and the six others above carry
  `seat.RequiredUnless(c, name, cond)` on the reading verbs (`seat.go:763`; `mint`'s are marked from the
  schema by `markRequired` and do not move), with `cond` = "unless this reading backs the sentence or reads
  nothing of the work" on `verify` and "unless this reading backs the sentence, reads nothing of the work
  or is `absent`" on `corroborate`, so the page opens "REQUIRED unless … — " and no seat loses a call to
  learn it. `TestHelpSaysRequiredExactlyWhereOmissionIsRefused` reads a marked flag the clean set never
  passed as "says REQUIRED and the verb ran clean without it", so its `requiredProbes` entries for `lens
  verify` and `lens corroborate` (`cli/requiredmarker_test.go:98-115`) [REWRITE] to a clean set that
  finds against: `--as refutes`, `--confidence`, `--reason`, the seven flags, on the fixture's citation
  or `--quote`. Each omission is then refused naming its flag; `--fix`, `--new`, `--complexity` stay out
  of the set and unmarked. The backing path is driven by `TestABackingReadingWithGapFlagsIsRefused`.
  `mint`'s rendered flag text is byte-unchanged. Tested by
  `TestTheReadingVerbsDescribeTheirGapFlagsByTheirSubject` (§III.10) and the gate above.
- **`--reason` is the reading's account and the gap's argument (F11, RULED: "Reading takes --reason").**
  `writeVerify` reads the one `seat.Reason(cmd)` (`verify.go:405-409`) and writes it to the two fields
  whose readers ask for it, at the act: `Verify.text`, as today, and, on the Mint the reading builds,
  `p.MintReason` as `mint.go:69-71` does. What the source says against the sentence IS why the gap is
  worth one. It is one text from one flag. No second flag is added, no reader joins `Verify.text` onto a
  gap, and no field is copied from a stored field or from `problem`. `Verify.text` is `(sql).required`
  with flag `reason`, so it cannot be the field left blank, and `--reason` is already REQUIRED on both
  reading verbs: a reading that finds against is refused without it, nothing recorded. `problem` stays
  the only statement of the defect. [NEW] `ReasonIs` rows (`cli/seat/seat.go:1082`) for `verify` and
  `corroborate`: "what the source says, in your words. Where this reading finds the source against the
  sentence, the same text is your argument for the gap it mints — kept on the reading and on the gap, the
  case blue answers and a bench weighs. " + `flags.ReasonNotProcess`. Where the one-gap rule names a
  standing gap, no Mint is written: the reason stays on the Verify, and the standing gap keeps its own
  `mint_reason`.
- The write: one transaction holding the Mint, the Anchor that places the gap's marker (§III.2a), and
  then the Verify carrying `carried_by` — Mint first, because `carried_by` references `mint.gap_id` and
  the store runs with `foreign_keys(1)`. The Mint's `location` is the reading's claim, anchors stripped
  (`citedSentence` for `verify`, `seat.Location` for `corroborate`), so every reader of a gap's quote
  (the board, the gap view, `report/assemble.go`) reads the sentence as today. `record.go` gains the
  reading's placed append beside `AppendPlaced` (one `appendOwn` over the three bodies); the Mint passes
  every mint validation (`record.go:955-980`: unclassified reason, subject, one-gap, budget) in the same
  `writeView`, so the one-gap read and the insert share the write lock (`_txlock=immediate`, as
  `onegapperclass.go:32-35` already relies on).
- **The budget exemption comes from the act, not the record (F6).** The Mint's budget check runs before
  the Verify exists in the transaction, so it cannot read a Verify naming the gap. The reading's placed append
  hands the Mint's validation the act's paired bodies (its Anchor and its Verify); `requireMintWithinBudget`
  admits a Mint so paired without counting it (§III.5), and the one-gap rule places it by the paired
  Anchor (§III.2a). A plain `mint` has no paired body and is counted and placed as today. Replay
  writes events one at a time (`migrate/replay.go:452`), so it would hand the Mint no pair: migrate
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
  | placement: `seat.AppendPlaced` with `mintPlacementRefusal` (`:212`, the `seat.Places` guard) | MET by `corroborate`, through `claimRefusal` (the same `anchortext.Refusal` arms); `verify` meets `seat.PlacesToSentenceOf` instead (§III.2a) |
  | `verifiedFix` on `--new` (`:136-139`, `record.CheckOnReport`) | MET, placed by the verb's subject (§III.2a) |
  | `--check-kind` through `record.CheckKindOf` (`:147-153`); the four grades (`:154-157`) | MET (the shared flag helper) |
  | `--supersedes`, `--found-by` (`:158-159`) | NOT TAKEN: the reading verbs register neither; the gap supersedes none, and its reading is named by `carried_by` |
  | `--key`: `record.ExistingMintByKey`, `seat.RepeatsKeyed`, `seat.PlaceOwed` (`:164-180`) | NOT TAKEN: the reading verbs take no `--key`; a retry meets the repeat refusal (crash-retry, below) |
  | `record.NewID("gap")` (`:182`) | MET (one id source) |
  | `requireNotEstopped` (`:208`, `record.EstoppelConflict`), and on its refusal the `LOG_TYPE_ESTOPPEL` entry and `record.ToolLogged` (`:211-247`) | **EXEMPT** (F10). Nothing is checked and nothing is logged, so `EstoppelRejectionsOf` and the change log's "Estoppel rejections" (`view/changes.go:93-103`) count lens mints only, and stay true |
  | `reportvoice.Advised` tells on the result (`:250-260`) | MET: the reading's `gap` result carries `voice_tells` over problem, fix and `--new`, as `mintResult` does |
  | `seat.Supplies("gap_id")`, `seat.Supplies("class_material")` (`:288-289`) | MET: the reading verbs declare both, beside `carried_by` and `carrier_minted` |

  **Every field `lens mint` sets on its body, and every reader of a field the reading leaves unset (F11).**
  Sweep at `0c57758e8`: every `p.<Field> =` in `newMint` and `verifiedFix` (`mint.go:68-183, :375-376`), the
  write path's stamp, and `grep -rnE '<field>|<Getter>'` over the tools tree for each, no tests, every hit
  read (`new-mintfields.txt`). Both verbs fill through `gapFlags.fill`, so these rows are one code path:

  | Field (`mint.go`) | The reading's Mint | Readers, where it is unset or defaulted |
  |---|---|---|
  | `mint_reason` (`:69-71`) | `--reason` | Carried, so `GapJSON.MintReason` (`viewjson.go:129, :448`), the gap view (`views.go:529`) and the report's "minted because" (`assemble.go:1124`, `:1142`) read it as a plain mint's. `requireUnclassifiedReason` (`record.go:1947`): no admitted class is `unclassified`. `redvoice.go:26`: never voice-checked, as on a mint |
  | `mint_key` (`:72`) | `""`, as a plain mint given no `--key` | `ExistingMintByKey` (`replay.go:368`), whose `keyedAct` returns nil on an empty key (`id.go:47`), so no retry finds it |
  | `class` (`:75-81`) | `--class`, narrowed (§III.1) | set |
  | `location` (`:113`) | the claim, anchors stripped | set |
  | `about_kind`, `about_ref` (`:114-115`) | UNSET, as a plain `--quote` mint: the subject is the sentence | `requireMintSubject` (`record.go:1937`) reads `location`; `requireAboutAnchor` (`record.go:973`) has no kind to check; the one-gap rule takes its sentence path (`onegapperclass.go:54`); `mintbudget.go:386, :403` and the about-anchor scan (`failedcheck.go:63`) pass it by, its carrier being `carried_by` (§III.3); `gapacts.go:41, :52`, `blueanswered.go:103`, `grounds.go:117`, `viewjson.go:415, :1066`, `views.go:526-527` read a sentence gap |
  | `problem` (`:116`) | `--problem`, required | set |
  | `required_fix` (`:117`) | `--fix`, optional | `proto.String` of the flag, `""` when absent, as a plain mint's: the report prints it blank alike |
  | `fix_basis` (`:135`) | `"proposed"` by default, `"verified"` through `verifiedFix` (`:376`) | Set. Without it `viewjson.go:811`'s "never omitted" is false and `fixProposal` (`assemble.go:1274-1284`) prints nothing; `estoppel.go:131`, `changesjson.go:113`, `view/changes.go:134` read it as a plain mint's |
  | `fix_new` (`:375`) | set by `verifiedFix` on `--new`, else UNSET | As a plain mint with no `--new`: `locateGap` and the board give no pair, `assemble.go:1277` prints none, estoppel skips it (`estoppel.go:131`) |
  | `acceptance_check`, `check_kind` (`:145-153`) | `--check`, `--check-kind`, required | set |
  | `severity`, `likelihood`, `impact` (`:154-156`) | flags, required | set |
  | `complexity_cost` (`:157`) | `--complexity`, optional | `nullWord` on the board, `grade` in the report, as a plain mint's |
  | `supersedes` (`:158`) | UNSET: the reading verbs take no `--supersedes` | A gap with no ancestor, as a plain mint's: `onegapperclass.go:64` (nothing exempt), `benchgrade.go:89, :198, :206`, `impasse.go:188`, `lensturn.go:52`, `gapacts.go:44`, `successor.go:41`, `record.go:1027`, `viewjson.go:166, :442`, `verify/verify.go:235`, `view/view.go:304, :457, :467`, `graph/graph.go:232, :280`, `dashboard/model.go:500`, `consistency.go:162`. A later gap naming it is §III.3's successor |
  | `found_by` (`:159`) | UNSET: the reading verbs take no `--found-by`; the reading is named by `carried_by` | A gap no finding surfaced, as a plain mint's: `provenance` prints no "surfaced by" (`assemble.go:982`), and the argument and the source's reading stand on `mint_reason` and `show evidence`; `requireFindings` (`record.go:1024`), `verify/verify.go:165, :442`, `view/view.go:273, :308`, `assemble.go:1345`, `viewjson.go:165, :442, :859, :1125, :1251, :1555`, `views.go:570` (no reader) |
  | `gap_id` (`:183`) | `record.NewID("gap")` | set |
  | `class_material` | stamped by the write path | set |

  The property: a reading's gap carries its argument and its `fix_basis` wherever a plain mint's appear,
  through the readers that exist. Tested by `TestAReadingsGapIsFilledAsAPlainMints` and
  `TestAReadingsGapCarriesItsArgumentWhereverAMintsDoes` (§III.10); partners §V.5.ee, ff.

  **Why estoppel does not reach a reading (gblock, 2026-10-31: "Exempt, stated").** Estoppel binds red
  to its own prescription: a gap whose quote is text blue applied verbatim from a red `--new` is red
  relitigating its words (`record/estoppel.go:14-29`). A reading is a source read against the sentence:
  its refutation is new evidence, which the prescription never weighed. Reachable today: `corroborate
  --as refutes --quote` of text blue applied verbatim, and `verify` of a citation whose token a red
  `--new` carried onto its rewording. The gap opens; a `verify`'s found-against check holds the PASS
  until it clears by the ledger's rule (§III.3), and a `corroborate`'s gap, material by its class, holds
  it while open. Tested by `TestAReadingOverAnAppliedPrescriptionMintsItsGap` (§III.10), partner §V.5.cc.

  **Sweep — where estoppel is taught or enforced**, run at `0c57758e8`:
  `grep -rIn -E '[Ee]stopp|ESTOPP' plugins scripts docs | grep -v -E '\.golden:|testdata/|record\.pb\.go|_test\.go|\.test\.mjs|/agents/'`
  (95 lines; `~/.claude/scratch/plan1647-r8/new-est.txt`), every hit read. Two senses: the
  PRESCRIPTION estoppel (the mint guard), and the work list's `estopped` register of bench-ruled
  closures. Dispositions:
  - enforced — `cli/lens/mint.go:185-208` (the guard's comment) and `:412-430` (`requireNotEstopped`'s
    doc) [REWRITE]: add that the reading verbs attach no such guard, and why; `cli/lens/verify.go`'s
    reading append [NEW comment] says the same at the write;
  - taught to the record's readers — `record/estoppel.go:14-29` (the rule's statement) [REWRITE]: a
    reading's gap is exempt, a source's refutation being new evidence; `record/recordpb/record.proto:1611`
    (`applied_verbatim`: "red may not open a clean slate against its own words") and
    `record/changesjson.go:74` ("Estoppel is enforced on the first") [REWRITE]: on a lens's mint, never on
    a reading's gap;
  - taught to seats — `help/mint.md:19` [REWRITE] adds: "A gap your reading of a source mints in its own
    act is not refused: a source that finds against the sentence is new evidence, not your prescription
    relitigated"; `help/verify.md` and `help/corroborate.md` (the new reading paragraph, §III.7) [NEW
    sentence] state it in the same words; `help/edit.md:13` (blue's: the estoppel "that stops red
    re-opening its own words") [REWRITE] adds "— a lens's reading of a source that finds against that
    text still mints its gap"; the generated `agents/*.md` (each inlines the help pages through the
    tool's `manual`) and the help golden REGENERATE;
  - counted, true as is — `record/estoppel.go:71-103, :238-257` (`EstoppelConflict`,
    `EstoppelRejectionsOf`), `view/changes.go:17, :93-103`, `record.proto:109, :1759` (`LOG_TYPE_ESTOPPEL`
    means "the TOOL refused a mint"), `record/record.go:1257`, `cli/seat/verbs.go:172`,
    `record/toollogged.go:5`, `recordpb/descriptions.go:202`, `cli/blue/edit.go:68, :188-201`,
    `migrate/entries.go` (archived log entries), `seatenv/identity.go:116`, `cli/seat/seat.go:819`,
    `cli/motion/verbs.go:95`, `report/assemble.go:1265`, `scripts/agentgen/src/lead-judge.md:101` — KEEP;
  - the rest of the guard's own lines, true as is — `record/estoppel.go:31-48` (`minEstoppelOverlap`, what the
    guard compares), `cli/lens/mint.go:341-345, :380` (the `estopped` error and `contains`),
    `record.proto:1732, :1780, :1783` (a `Log` comment and the reserved `estopped_by`) — KEEP;
  - another sense — `record/outcomeclosed.go:24` (`cli/enginestopped.go`, the engine's stop) — KEEP;
  - the `estopped` register (another sense: a bench ruling's bar, read as advice by every seat, and
    refused at no write) — `record/viewjson.go:650-1277`, `cli/seat/verbs.go:327, :343`,
    `record/sitting.go:589`, `record/declarations.go:28`, `record/enums.go:166`,
    `cli/motion/verbs.go:26`, `recordsql/views.go:487` — KEEP; a reading's gap meets it as a plain mint
    does;
  - the fuzz — the estoppel drive (`fuzz_test.go:576-668, :1396-1400`) mints with `lens mint` and KEEPs;
    its gates (`sweeptally_test.go:1036-1061`) count tool estoppel entries, which no reading writes, and
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
  `FindsAgainst` holds, by the same lens, of the same sentence from the same source (the anchor for
  `verify`; the url and claim for `corroborate`), with the same outcome as an earlier reading of its
  own for which `FindsAgainst` held, sentence unchanged, while a gap of that earlier reading's carrier
  tree (`carried_by` and its successors, §III.3) is open — refused, nothing recorded, naming the earlier
  reading's event and the carrier (`AnsweredBy()`): "your reading of citation C at event N finds the
  source against this sentence, the sentence has not changed, and gap G carries that reading and is
  open — nothing is recorded; the reading and its gap stand". A reading with a *different* outcome is
  not a repeat. A reading for which `FindsAgainst` does not hold (`unreachable`, any outcome of a
  `not_the_work` copy) never meets this arm: over a citation it meets the `unread` arm as today, keyed
  on that state (§III.3).
- Text that states the old model without the word [REWRITE]: the `--as` flag help (`verify.go:304`,
  "refutes and absent are findings") and `human()` (`verify.go:701-704`, "that is a finding, and this event
  is the evidence for it"); see §III.7's phrase sweep.
- Result: `verifyResult.Contradiction` / `readingContradiction` / `contradictionOf` / `words()`
  (`verify.go:615-691`) [DELETE]; [NEW] `Gap *readingGap {id, class, minted}` and its sentence:
  "This reading mints gap G (class C)" / "Gap G, open on this sentence with class C, carries this
  reading; no second gap is minted — add what it lacks as a finding about G". `gap.minted` is read back
  from the stored `carrier_minted`, never computed by the verb, and `Gap` is set exactly when the
  stored Verify carries `carried_by`: a reading whose `ReadingGapClasses` is empty prints no `Gap`
  sentence, whatever its `--as` (its citation's words are its failed check's, by state, below).
- **The failed-check words key on the state** (`verify.go:529-541`, `:631-656`) [REWRITE]. Today the
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
- **Crash-retry and the repeat refusal.** `verify` and `corroborate` take no `--key`: a reading keys on
  its `KeysOn` fields (`verify.go:116` anchor; `:296` url and label), per seat and sitting
  (`record.go:540-552`). A crash before the commit leaves nothing (one transaction, R8), and the retry
  writes the pair afresh. A retry after the commit is the same seat, sentence and source in the same
  sitting, and meets the repeat refusal above before any write — its validation runs first, and the
  one-gap read of the retry finds the gap its first write minted — so it is refused, nothing recorded,
  naming the stored reading's event and its gap. That refusal is how a crash-retried negative reading
  learns its first write landed; it is not an idempotent return. `ExistingCorroboration`
  (`verify.go:459-480`) runs only on a backing corroboration (the label it keys on is minted only
  there) and is unchanged. A negative corroboration `absent` (no gap, F5) retried in its sitting meets
  the duplicate-key refusal on its url, as today. So does a BACKING corroboration of a `not_the_work`
  copy, a new shape: the footnote gate (first bullet) mints it no `label`, so it keys on `url`
  (`keyFields`: label, then anchor, then url), not on its label — and a second claim the same seat
  corroborates from that url in the same sitting meets the duplicate-key refusal on the url, nothing
  recorded, where a labelled backing corroboration would have been admitted under its own label. Stated,
  not changed: the copy reads nothing, so a second claim from it reads nothing either. Tested by row (h)
  of `TestAReadingOfACopyThatIsNotTheWorkMintsNothing`. Tested by
  `TestARetriedNegativeReadingIsRefusedNamingItsGap` (`cli/keyedretry_test.go`, §III.10).

### III.2a Where a reading's gap stands — by the anchor's identity (F9) [NEW] (contract: event schema)

gblock's ruling (2026-10-30, his words): "I think if only text is provided out [it] should be refused
but it's fair to point at an anchor using the right flag".

**How each reading verb names its subject today** (read at `0c57758e8`):
- `verify` takes no text. `--anchor` is required (`verify.go:113-114`), checked to name a citation
  (`record.CitationExists`), and the claim is read by the anchor's identity:
  `bluedoc.SentenceAround(report, anchor.Token(id))` (`verify.go:129-150`). It places nothing today
  (`quote = ""` for a non-citing verb, `verify.go:510-513`).
- `corroborate` takes text and no anchor: `--quote` is the report's sentence (`verify.go:291`), placed
  by text (`seat.AppendPlaced` → `Places` → `anchortext.Attach` → `LocateOnce`, `verbs.go:1384-1417`,
  `anchortext.go:315-336`), which refuses a quote the report holds more than once (`AmbiguousError`).
  It has no anchor flag: the source it read is one blue never cited.

**So the path that places is `verify`'s, and the path that refuses is `corroborate`'s.** A `verify`
names its citation by `--anchor`, the right flag; its gap is placed at the end of the sentence that
anchor sits in, by the anchor's identity. A `corroborate` names text only; its gap is placed at its
`--quote` by text, as its citation's anchor is placed today, and a quote the report holds more than once
is refused, nothing recorded — `claimRefusal` (`verify.go:716-727`) words it through `anchortext.Refusal`
(`anchortext.go:508-528`), which names each occurrence's section (`WhereAmbiguous`) and tells the seat
to quote the sentence with the text before it in its paragraph so the quote occurs once. Since #1654
`Refusal` also words `ErrInQuotation` ("an anchor sits after the quotation it ends on: quote through
the closing mark", `:524-525`), raised where `Attach` (doc `:547`, func `:554-568`) finds the quote stopping short of a
closing double quotation mark; `claimRefusal` inherits that arm for a negative corroboration's gap as
for its footnote, and it is the same sentence the seat already gets. This plan adds no anchor flag to
`corroborate`.

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
  searches text. Name grep at `0c57758e8`: `AttachToSentenceOf`, `ErrAnchorGone`, `ErrAnchorBare` 0 hits.
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
  since guards run before validation, `record.go` `Append` doc). `gapSentenceIn` (`viewjson.go:2020`)
  then reads the sentence the marker sits in — the cited sentence, whatever text repeats elsewhere.
- [MODIFY] migrate's placer (`migrate/gapplace.go:78-79`, the Anchor arm, and `places`, `:104-115`): an
  Anchor with `sentence_of` is checked by `AttachToSentenceOf` on the migrated render, never by
  `places(id, location)`, whose empty location would pass it unchecked. Its Mint arm (`:55-70`) skips a
  gap the stream anchors itself (`p.anchored`), which every reading's gap is. No archive holds one; the
  next migration of a record written at this epoch does. Tested by
  `TestMigrateChecksAReadingsGapBySentenceOf` (§III.9).
- `verify`'s Anchor for its gap is `{id: G, sentence_of: <--anchor>}`. `corroborate`'s is
  `{id: G, location: <--quote>}`, exactly as `lens mint` places one.

**`--new` on a reading — blue's accept locates the span by the gap's anchor (RULED, gblock
2026-10-30).** A verified fix is one blue's accept applies over the report with the gap placed
(`mint.go:350-378`, `verifiedFix` → `blue.PlanAccept`, `cli/blue/edit.go:335-341` →
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
- The board's pair (`viewjson.go:1996`, `locateGap`) and estoppel's comparison (`proposalAt`,
  `estoppel.go:159-164`, read by `ProposalAppliedVerbatim`, `:220-232`, `landing.go:109`, `edit.go:206`)
  read the same `ProposalOver`, so the pair the board serves, the pair the accept applies and the pair
  estoppel compares are one.
- [MODIFY] `verifiedFix` (`mint.go:350-378`) builds `placed` by the gap's placement — `Attach` for a
  `--quote` (mint, corroborate), `AttachToSentenceOf` for a `verify` — and its refusal (`:366-376`,
  "with this gap's anchor at the end of --quote", "keep the sentence …") names the subject the verb took:
  `--quote` for `mint` and `corroborate`, "the sentence your --anchor's token sits in" for `verify`,
  which has no `--quote`. `help/verify.md` says it in the same words. "`--new` on a repeated sentence is
  refused" is not part of this design: the accept edits the cited occurrence.

**Sweep — every reading-act site that places a gap or reads where one stands, and every text-locate of a
proposal span**, run at `0c57758e8` (45 hits, `~/.claude/scratch/plan1647-r8/new-place1.txt, new-place2.txt`):
`grep -rn -E 'anchortext\.(Attach|LocateOnce|InsertAnchor)|seat\.(AppendPlaced|Places|PlaceOwed)|record\.AppendPlaced|reportproj\.Place|gapSentenceIn|ProposalOver|SentenceAround' plugins/frank-exchange-of-views/tools/internal plugins/frank-exchange-of-views/tools/releasegate | grep -v _test.go`,
plus `grep -rn -E 'PlanAccept\(|proposalAt\(|ProposalAppliedVerbatim\(|PlanSplice\(|LocateUnique(Replacing)?\(' plugins/frank-exchange-of-views/tools/internal | grep -v _test.go`,
every hit read; the reading-act ones, the proposal-span ones, and those the new placement reaches:

| Site | Today | After | Disposition |
|---|---|---|---|
| `cli/lens/verify.go:95-120` `newVerify` / `:129-151` `citedSentence` | anchor required; claim read by the anchor's identity | unchanged; the anchor is also what places the gap | KEEP |
| `cli/lens/verify.go:501-514` `writeVerify` placement | `quote=""` for verify; corroborate's quote placed by text | verify: the reading's append writes `Anchor{G, sentence_of}`; corroborate: `Anchor{G, location}` by text, ambiguity refused | REWRITE |
| `cli/lens/verify.go:716-727` `claimRefusal` | corroborate's placement refusal through `anchortext.Refusal` (ambiguity, crossing, word split, heading, `ErrInQuotation`) | the same refusal for a negative corroboration's gap | KEEP |
| `cli/seat/verbs.go:1384-1417` `Places`, `AppendPlaced` | text placement guard | gains `PlacesToSentenceOf`; text path unchanged | MODIFY |
| `cli/seat/verbs.go:1452-1474` `PlaceOwed`; its callers `verify.go:480`, `cite.go:83`, `finding.go:92`, `mint.go:177`, `compute.go:114` | re-places a migrated act's stored location | not reached by a reading's gap (its Anchor is in its own write) | KEEP |
| `record/record.go:664-668` `AppendPlaced`, `:941-951` Anchor validation and its comment `:942-943` | location required; "at its quote's end" | location XOR sentence_of, comment names both | MODIFY |
| `record/recordpb/record.proto:1179-1187` `Anchor` and its comment | "at the end of its quote `location`" | `sentence_of = 7`, comment names both placements | MODIFY |
| `record/onegapperclass.go:82-92` | `Attach` by text; nil on failure | sentence_of placement for a paired Anchor, failure refused | MODIFY |
| `record/recordsql/views.go:793-821` `report_op`; `record/queries.go:251-256` `ReportOp` doc | four kinds; insert rows need a location | `'sentence_of'` arm; doc names five kinds | MODIFY |
| `record/queries.go:381` `PartStanding`; `record/migrate/order.go:118`; `record/landingreport.go:28` | read `"kind"` against `'append'`, or count rows | a sentence_of row is a change, as an insert is | KEEP |
| `reportproj/render.go:113-131, :295-308` | insert by location | `sentenceOfMut` | MODIFY |
| `record/migrate/gapplace.go:55-70, :78-79, :104-115` | places archived anchors by location | checks sentence_of placements by identity | MODIFY |
| `record/migrate/replay.go:222-505` (the replay loop) | writes each source event alone | re-drives a reading act's Mint, `{G, sentence_of}` Anchor and Verify as one act, each member keeping its own per-event duties (§III.9 item 6) | MODIFY |
| `record/batch.go:82-114` `Batch.Append` | one body per savepoint | [NEW] `Batch.AppendBodies` beside it, the act's bodies in one savepoint (§III.9 item 6) | MODIFY |
| `record/id.go:65-91` `unplacedLocationAt` | the location of an act no Anchor places | a reading's gap always has its Anchor; reads `anchor.id` only | KEEP |
| `consistency/consistency.go:254-256, :521` | reads `Anchor.id` | id only | KEEP |
| `record/estoppel.go:175-211` `ProposalOver` | text-located span | located by the gap's token | MODIFY |
| `record/estoppel.go:159-164, :220-232` `proposalAt`, `ProposalAppliedVerbatim`; `record/landing.go:109`; `cli/blue/edit.go:206` | read `ProposalOver` | unchanged code, the identity-located pair | KEEP |
| `record/viewjson.go:1996-2032` `locateGap`, `gapSentenceIn` | `ProposalOver`; sentence by token | unchanged code, the identity-located pair | KEEP |
| `cli/blue/edit.go:145` (`--accept`), `:335-349` `PlanAccept`/`planTyped`, `:401-445` `validateEdit` | ordinary locate of `old` | exact-span locate of the token-carrying `old` where the ordinary one is ambiguous; `validateEdit`'s stay-put check (`:418`) is skipped on an exact span, as today | MODIFY (`PlanAccept`/`planTyped`) |
| `reportproj/plan.go:55-115` `PlanSplice`; `bluedoc.go:45-74` `LocateUnique`, `LocateUniqueReplacing`; `cli/blue/edit.go:418, :465` | a typed edit's locate | unchanged: a typed edit's quote is the seat's | KEEP |
| `reportproj/render.go:89-105` (an edit's replay); `migrate/gapplace.go:159-166` `locate` | exact span byte for byte | unchanged; the accept's exact span replays through it | KEEP |
| `cli/lens/mint.go:350-378` `verifiedFix` | `Attach`; refusal speaks of `--quote` | placement by the verb's subject; refusal names it | MODIFY |
| `bluedoc.go:204, :343-370` `SentenceAround`, `:400`; `cli/blue/retire.go:180-200` | sentence by token; an edit's reopen check | identity already | KEEP |
| `cli/motion/verbs.go:170` (#1652, `seat.Places("", q, groundsQuoteRefusal)`) | a docket motion's grounds quote, checked against the report | another sense: a quote of report text, no gap placed | KEEP |
| `cli/blue/compute.go:114, :200`; `cli/blue/cite.go:83, :160`; `cli/lens/finding.go:92, :120`; `cli/lens/mint.go:111, :177, :212` | their own acts' text placement | not reached | KEEP |
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
corroboration carries it: the stamp (`verify.go:491-500`) runs where `body.GetUrl() != ""`, and a
`verify`'s body has no url (`newVerify`, `verify.go:97`; its source is `citedURL`'s, `:153-172`), so a
`verify` of a `not_the_work` copy reads as a reading and its citation leaves `unread`. `help/verify.md:11`
already states the model the code does not implement; the only test is a synthetic `not_the_work`
Verify at the record layer (`record/reopened_test.go:260`), and no CLI test drives it. A site that
decides "negative reading" from `refutes`/`absent` alone, or reads an unstamped `verify`, would demand
gap flags on such a reading and mint a `claim_contradicts_source` gap no failed check carries —
closable freely — while the `unread` check separately owes its about-anchor gap. `verify --as absent`
of such a copy is already refused at `verify.go:379-389` (`WithoutBody`, read from the cache entry the
verb looks up by `source`, so it works for `verify` today), so the live hole is `refutes` (both verbs)
and every backing outcome; `absent` of an uncached copy is a reading (`not_asked` reads).

**The ruling (gblock, 2026-10-30): `verify` stamps it.** §III.2's first bullet: every reading stamps the
`source_completeness` of the cache entry for the source it names — the cited url for `verify` — and
every reading of a `not_the_work` copy is unread, a backing one included.

**The property:** a reading for which `VerifyReads` is false — `unreachable`, or ANY outcome of a copy
stamped `not_the_work`, `supports`/`supports_with_bridge`/`weak` included, by `verify` or
`corroborate` — mints no gap and no footnote, carries no `carried_by` or `carrier_minted`, has its gap
flags refused, and holds the PASS the way an unread reading does: for `verify`, the citation's `unread`
failed check, owned by its reader, and the about-anchor gap that check owes; for `corroborate`, the
reading alone (it cites nothing, so nothing is held).

**Sweep**, run over this plan and the code it specifies — every site that decides, counts or words a
negative reading — with `grep -n -i -E 'negative (reading|outcome|half|corroboration)|refutes|absent|gap_classes|found.against|finds.against'`
over this plan and `grep -rn -E 'SOURCE_OUTCOME_(REFUTES|ABSENT)|contradicts\(|VerifyReads'` over
`tools/internal`, `tools/releasegate` (no test files) at `0c57758e8` (22 hits, `~/.claude/scratch/plan1647-r8/new-cls.txt`), and — for the stamp the ruling adds — every
Verify writer (`grep -rn 'recordpb.Verify{' tools --include=*.go | grep -v _test`) and every reader of
`source_completeness` (`grep -rn -i 'source_completeness\|SourceCompleteness' plugins/frank-exchange-of-views scripts docs`, less
`fetchcache/` and the cite's own readers), every hit read:

| Site | Decided by | Disposition |
|---|---|---|
| Gap-flag requirement (§III.2) | `--as` | REWRITE: `ReadingGapClasses` after the completeness stamp |
| Write-path rule "a negative reading names its gap" (§III.1, `record.go` Verify case) | outcome | REWRITE: `ReadingGapClasses` of the stored body; either field on a non-minting reading refused |
| Repeat rule's negative arm (§III.2) | outcome | REWRITE: `FindsAgainst` of both readings |
| Reading result's `Gap` words (§III.2) | outcome | REWRITE: set from the stored `carried_by` only |
| Held-gap / one-gap path in the reading append (§III.2) | reached only from a minting reading | KEEP (downstream of the requirement) |
| Budget exemption (§III.5) | the act's paired bodies; `carrier_minted` | KEEP: no Mint is paired with a non-minting reading, and the write rule refuses `carrier_minted` on one |
| Anchor ledger `anchorchecks.go:260-261`, `CitationsFoundAgainst`, `stateFoundAgainst` | `reads && contradicts` | REWRITE: `FindsAgainst` (same truth table; `contradicts` deleted) |
| `failedChecksOf` found-against arm (§III.3) | ledger state | KEEP (reads-gated by the ledger) |
| The footnote decision (`verify.go:426`, `mayCite == cites && backsTheClaim(...)`) and `backsTheClaim` (`:336-350`) | outcome, positive half | REWRITE (GAP 1, RULED): the decision adds `record.VerifyReads(body)`, so a backing corroboration of a `not_the_work` copy mints no label, footnote or Anchor; `backsTheClaim` keeps its outcome-only question and its comment's "They owe a finding instead" goes |
| Live Verify writers: `verify.go:97` (`newVerify`) and `:280` (`newCorroborate`), both through `writeVerify` | — | MODIFY: the stamp in `writeVerify` keys on `source`, so both verbs stamp; migrate's archived Verify (`migrate/entries.go`) carries what the archive held — an archived `verify` has none and keeps reading (residue, §III.9) |
| The stamp `verify.go:491-500` | `body.GetUrl() != ""` | REWRITE: `source != ""`, before the footnote decision |
| `anchorchecks.go:259-266` ledger Verify arm; `VerifyReads` `:284-289` | the stored stamp | KEEP (code): reads the stamp `verify` now writes; the `VerifyReads` comment `:278-283` [REWRITE] (§III.3) |
| `evidenceview.go:270-273, :422` (a verification row's `source_completeness`) | the stored stamp; empty on every `verify` today | KEEP (code): every verification row now carries it; `show evidence` goldens that print a `verify` row REGENERATE, each changed line read |
| `evidenceview.go:499`, `citationid.go:46-49, :89, :106`, `report/assemble.go:155, :202` (a source row's, a footnote's, the assembly's abstract note) | a cite's or a labelled corroboration's stamp | KEEP: a footnote is minted only from a reading of the work now, so no source row carries `not_the_work` from a corroboration |
| `record.proto:1392-1394` (`Verify.source_completeness` comment) | "which part of the work red's copy is" | KEEP (true of every reading now) |
| `docs/citations-and-bibliography.md:178` ("Every verdict naming a url is stamped") | a corroboration only, in fact | REWRITE: every verdict is stamped, `verify`'s from its citation's url; any verdict on a copy not the work reads nothing and leaves the citation unread |
| `help/verify.md:11` | states the model | KEEP its unread paragraph's model (now true); its "does not close" clause REWRITEs (F8, §III.7) |
| `record/reopened_test.go:260` (synthetic `not_the_work` Verify) | record layer only | KEEP; the CLI rows of `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` drive the verb |
| `cli/corroborate_test.go:301-321` (a corroboration's stamp, `abstract`, `not_asked`) | corroborate only | KEEP; the verify rows (i) are new |
| `verify.go:379-389` `absent` on a copy without its body | outcome + completeness | KEEP |
| `verify.go:532` failed-check words guard | `!VerifyReads(stored)` | REWRITE by state (§III.2), already reads-aware |
| Fuzz reading drive (§III.11) | outcome | REWRITE: draws gap flags from the record's stamp of the copy the reading resolves to, and drives a `not_the_work` copy |
| Fuzz gates "negative readings that minted / refused a mint" (§III.11) | outcome | REWRITE: counted from stored `carried_by`, never from `--as` |
| §V.8 smoke query | outcome | REWRITE: `FindsAgainst` rows only, plus the converse query |
| Success criterion 1 (§I) | outcome | REWRITE: "of a copy that is the work" |
| Help `verify.md`, `corroborate.md`, `mint.md`; agentgen fragments; `terms.json` `failed check` (§III.7) | prose "a negative reading mints" | REWRITE: "a reading of the work that finds the source against the sentence mints"; `verify.md` states that any outcome of a `not_the_work` copy, like `unreachable`, mints nothing and leaves the citation unread |
| Migrate (§III.9) archived negative readings | ledger | KEEP: archived readings carry no carrier; their state is the ledger's, reads-gated |
| `TestEveryNegativeOutcomeNamesItsClasses` | the facet itself | KEEP: it tests the declaration, which is per outcome by design |

Tested by `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` (§III.10).

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
  - dispatch's lens arm (`dispatch.go:558-565`): readies the owner for `unread`/`unsound` with no carrier;
    never for `found_against`, which readies blue alone (`:567-573`, cause `max(since, answeredAt)`);
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
  - `failedCheckReadiness` (`dispatch.go:1326-1341`): `what, failed` and blue's remedy by state; the lens
    arm's "minted no gap about it" stays, as only `unread`/`unsound` reach it.
  - the verb's failed-check sentence (`verify.go:529-541`, `:631-656`): by state, after every
    verification of a citation, §III.2 — this site was judged state-safe in round 2 and is not: its
    guard is the new reading's outcome, not the check's state.
- **The lens's turn (#1659) reads a reading's gap as it reads any mint.** `lensTurnOn` and capture's
  `LensTurnAudit` take the Mint's place from `mintPlaces`, so blue is readied first on a reading's open
  material gap, and its minting lens once blue has acted on it. A Verify is no act on a gap
  (`gapActsOf`, `gapacts.go`, has no Verify arm), so a second lens's reading naming G, or the reader's
  own re-read, moves no turn. Both stay unchanged, and the check's readiness above is by state. The
  reader's backing re-read happens in any sitting a plan readies it for.
- Read and judged state-safe: `verdict.go:141-162` (`anchorsStanding`) and `debate.js:1050-1057` name
  the anchor only, whatever its state.
- The close gate: `requireFailedCheckAnswered` refuses the minting lens's (or any seat's) `close` of a
  gap in a found-against check's `answers` — successor (`--superseded-by`) admitted — until the check
  clears by the ledger's rule above. A closing docket ruling closes the gap and the check still holds
  the PASS (above). The gap's close stays the minting lens's ordinary `close`, so no new answer kind
  exists.
- **The mechanism's own sweep** — every site that SELECTS a failed check (by Subject, Owner, AnsweredBy
  or Detail) and decides or prints, re-run at `0c57758e8` (`~/.claude/scratch/plan1647-r8/new-fc.txt`) with
  `grep -rn -E 'BlockerFailedCheck|AnsweredBy\(\)|failedChecksOf\(|\.answeredBy|\.answers\b|\.Detail\b'`
  over `tools/internal`, `tools/cmd`, `tools/releasegate` and `debate.js`, every hit read (dispositions in
  §III.12's table "Readers of a failed check's owed-ness"): the record readers above; `verify.go:538`
  (§III.2); `sitting.go:321` (blue's arm, prints `blueFailedItem`, by state); `passblockers.go:714` (the
  producer, which now sets the state); the fuzz sites of §III.11 (`fuzz_test.go:1218, :1227, :3007, :3017,
  :3039, :3042, :3096-3105, :3415, :3719, :4192, :4237, :4269`, and the gates at `sweeptally_test.go:834-845`);
  KEEP — `verdict.go:148`, `debate.js:1056`, `fuzz_test.go:2953` (asserts the owner or blue is readied,
  true of every state), `fuzz_test.go:6974` (no VERIFIED over any failed check), `migrate/plankinds.go:19`; another kind's `.Detail`
  (#1660's `upheld_position`: `dispatch.go`'s lens arm, `passblockers.go`'s kind row, `positionheard.go`).
  The contradiction fold's hits (`citationid.go:554-685`, `onegapperclass.go:145`, `mintbudget.go:362-380`)
  are deleted. `PlanBlocker` (`dispatch.go:36-40`) carries kind, subject and owner only, and stays so:
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
- Corroborations place no anchor when negative, so they reach the board through their gap only.

### III.4 The one-gap rule — `onegapperclass.go` [MODIFY]

- `answersRaisedContradictionAt` and its call (`:68-72`) [DELETE]; the exemption bullet (`:40-42`)
  [DELETE]. The rule now has no exemption but `--supersedes` and replay.
- [NEW] the held-gap return used by the reading act (III.2).

### III.5 Mint budget — `mintbudget.go` [MODIFY] (F6)

- `answersOwnRaiseAt`, its call (`:301-305`), `ownAnswersAt`'s contradiction half (`:369-397`), the
  `Answering` wording at `:506`, `:562` [DELETE/REWRITE]; the comment block `:33-41` [REWRITE].
- [NEW] a gap the reading act minted takes no mint and does not count while open. Two readers, two
  sources, neither inferred from "some Verify names this gap":
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
- `mintAnswerOutsideBudget` (`:347`) [REWRITE]: "A gap your reading mints, or about an anchor whose check
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
- `about.go`: the contradiction case (`:47-50`, `:82-88`) and the doc sentence (`:15-16`) [DELETE].
- `record.go`: the finding case (`:1062-1066`) and the mint's about-kind refusal (`:966-971`) [DELETE] —
  with every kind admitted the check has nothing to refuse; `enums.go:294-300` (finding's set "less
  contradiction") [REWRITE]: finding and mint carry one set.
- `sitting.go:374-439` (the lens owes a contradiction it read; the at-cap special case) [DELETE]; the
  failed-check arm (`:401-408`) keys on the state (§III.3).
- `dispatch.go`: the readiness case (`:547-553`), `contradictionReadiness` (`:1314-1321`), Plan JSON
  `desc` text (`:19-38`) [DELETE/REWRITE].
- `retirement.go:20, :142` (a lens is not retired while it owes a contradiction) [REWRITE]; the
  failed-check skip (`:164-172`) keys on the state, so a found-against check never holds its reader
  from retiring (§III.3).
- `evidenceview.go:306-373`: `UnansweredContradictions`, `EvidenceContradictionJSON` [DELETE];
  `:578`, `:598` comments [REWRITE]. `narrowedview.go:20` [REWRITE].
- `convergence_refusal.go:77` [REWRITE].

### III.7 Agent-facing text [MODIFY / REGENERATE] (contract: prompts and constitutions)

- Help pages (`internal/cli/seat/help/`): `verify.md`, `corroborate.md` — the two contradiction paragraphs
  are replaced by one paragraph: a reading of the work that finds the source against the sentence
  mints its gap in the same act, the flags it needs, and that a verdict that reads nothing —
  `unreachable`, or any verdict on a copy the tool stamped not_the_work, a backing one included — mints
  nothing (no gap, and on `corroborate` no footnote), its gap flags are refused, and the citation stands
  unread (`verify.md:11`'s unread paragraph keeps that duty, now true of `verify`, which stamps the
  copy's completeness); where the gap stands — `verify`: "the gap stands at the end of the sentence your
  --anchor's token sits in, found by that token, so a sentence whose words the report holds twice takes
  its gap where you cited it, and a --new you give is checked, and accepted by blue, at that
  occurrence";
  `corroborate`: "the gap stands at your --quote, which the report must hold once; a quote it holds
  more than once is refused"; what the one-gap rule does to a second reading, and what clears the citation's failed check (your own
  later reading of the sentence as it stands that backs it, blue's edit, or blue's retire; another lens's
  backing reading is a disagreement and clears nothing); the gap's flags and its argument: "the --reason
  you give says what the source says, and it is also your argument for the gap: it stays on the reading
  and on the gap, where the board shows it to blue and the bench"; and estoppel's exemption (F10), in the words
  `help/mint.md:19` gains (§III.2): "A gap your reading of a source mints in its own act is not
  refused: a source that finds against the sentence is new evidence, not your prescription
  relitigated";
  `corroborate.md:17`'s "How you found the source goes in your reason, which the report never prints" keeps its
  footnote case and is REWRITTEN to say so: on a reading that mints, the reason is the gap's argument. `mint.md:23` keeps its text: `mint`'s `ReasonIs` row is byte-unchanged.
  `finding.md`, `mint.md`, `dispatch.md` [REWRITE] the contradiction sentences; `mint.md:19` and
  `edit.md:13` [REWRITE] for estoppel's exemption (§III.2); `finding.go:144-145` flag
  descriptions [REWRITE]; `verbs.go`, `flags/names.go` per census. The failed-check paragraphs name the
  third state: `mint.md:15` ("A CHECK OF YOURS THAT FAILED … you owe its gap") says the owed gap is an
  unread or unsound check's, and that a reading that finds the source against its sentence mints its gap in the same act;
  `close.md:11` [REWRITE] opens "YOUR CLOSE OF A GAP THAT CARRIES A FAILED CHECK IS REFUSED WHILE THE
  CHECK STANDS", adds the found-against case, states what clears each state by the ledger's rule (a
  found-against check: your own later reading of the sentence as it stands that backs it, blue's edit of
  the sentence, or blue's retire of the citation — another lens's backing reading clears nothing), and
  says a closing ruling of the bench closes the gap while the check still holds the PASS.
- **No page says a gap carrying a failed check does not close** (gblock's ruling, F8). Sweep, run at
  `0c57758e8` (123 hits; list at `~/.claude/scratch/plan1647-r8/new-dnc.txt`) over `scripts/agentgen/src`, `tools/internal`, `tools/releasegate`, `skills`, `docs`:
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
  `failedcheck.go:30`, `unreadcitation_test.go:119`, `unreadb6_test.go:21-22, :66`, `fuzz_test.go:3099,
  :4189`, `sweeptally_test.go:840`; REGENERATE `docs/vocabulary.md:629` and the help golden. Every other
  hit is another sense (a sitting the record cannot close, a quote or fence that does not close, the
  `remanded` disposition, `record/viewjson.go:1006`, `positionowed.go`, `remand_test.go`,
  `positionheard_test.go`) and KEEPs.
- `terms.json`: the `contradiction` entry [DELETE] (with its ordinary-English collision); `failed check`
  [REWRITE]: "A failed check is a standing check of a citation or computation that fails — a lens's standing
  judgement calls the computation's sentence unsound; no reading of the citation since its sentence last
  changed reads the cited work; or a lens's standing reading finds that the cited source refutes the
  sentence or does not contain it, and that reading mints the gap that carries it in the same act. It
  holds the PASS and is cleared by no
  bench ruling. An unread or unsound check obliges the seat whose check failed to mint a gap about the
  anchor, outside its mint budget. A seat's close of the gap that carries a failed check is refused while
  the check stands; a closing ruling of the bench closes that gap, and the check still holds the PASS.
  It clears when the anchor is checked and holds — for a computation or a found-against citation, by the
  checking lens's own later check, since another lens's check beside it is a disagreement; for an unread
  citation, by any lens's reading of the work — when an edit changes its sentence, or when blue retires
  it." `lens retirement` [REWRITE] ("no contradiction it read" goes). Then `go -C scripts run ./vocabdoc` regenerates `docs/vocabulary.md`.
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
  At `0c57758e8` it prints 56 lines (list kept at `~/.claude/scratch/plan1647-r8/phrase-new.txt`; the same 56
  as at `a31a7b06d`, moved by the line map, and no file #1663 added prints one: `blue-lane.md`,
  `blue-researcher-frame.md` and `blue-search.md` hold none). The 29 beyond `cli/` and the
  agent-text paths, each read:
  REWRITE — `record/anchorchecks.go:279` (`VerifyReads` comment, §III.3), `record/evidenceview.go:245`
  (comment: "at high confidence it is a finding to act on" → the gap its reading minted),
  `record/recordpb/record.proto:509` (`REFUTES` means, §III.1), `record/evidenceview_test.go:250`
  (comment "THE FINDING HALF" → the negative half), `fuzz_test.go:996` (the quote comment: a negative
  corroboration's quote is the sentence its gap is minted on), and the census rows already listed —
  `blockerowner_test.go:202`, `narrowedview_test.go:34`, `redcitation_test.go:347`,
  `migrate/reproduceanchor_test.go:158`, `fuzz_test.go:5480`; DELETE — `record.proto:1143`,
  `dispatch.go:1320`, `contradictionquadratic_test.go:17`, `terms.json:763`; REGENERATE —
  `record.pb.go:737, :8643, :8705`; KEEP — `record.proto:501` (why the outcome field exists, as
  `verify.go:56`), `evidenceview.go:238` (the same history),
  `unchecked.go:29` (the reopened-evidence history), `narrowedview_test.go:49` (fixture text),
  `capture/dispatchparity.go:208`, `debatejs/debatejs.go:388`, `changesjson.go:32`, `viewjson.go:886`,
  `report/assemble.go:1183`, `report/assemble_test.go:997, :1000`, `seatprobe/naming.go:191` (ordinary
  English).
  Known hits at `0c57758e8` (27; the 9 under `cli/*_test.go` are the census's test rows or other-sense
  KEEPs — `missingflag_test.go:119`, `stategraph_test.go:91`, `verbs_test.go:114`):
  REWRITE — `scripts/agentgen/src/skills/research-protocol/SKILL.md:17` (the sentence spanning l.17-19,
  which a line grep cannot match), `fragments/lens-audit-and-mint.md:8`, `red-chair.md:12`,
  `verify.go:304`, `verify.go:702`, `help/corroborate.md:13`, `help/finding.md:11`, `help/dispatch.md:15`
  and `:19`, `debate.js:889`, `docs/citations-and-bibliography.md:173` ("`refutes` and `absent` are
  findings, not failures to grade"); REGENERATE — `skills/research-protocol/SKILL.md:19`,
  `docs/vocabulary.md:463`; KEEP — `verify.go:56` (a code comment on why the outcome field exists),
  `fragments/lens-audit-and-mint.md:9` (materiality), `debate.js:227` (inquiry `refuted`),
  `docs/seat-surface-naming.md:749`, `docs/why-a-seat-stops.md:103` (ordinary English).
  The property each rewritten sentence holds: a reading of the work that finds the source against the
  sentence mints its gap in the same act, and a verdict that reads nothing mints none.
- `debate.js`: enginegen regions regenerate (`go -C plugins/frank-exchange-of-views/tools run
  ./devcmd/enginegen`); authored hits per census [REWRITE]; simulator goldens regenerate.
- Seat prompts name acts, never verbs or flags (`TestNoPromptGrowsItsCommandCatalogue`,
  `TestTheSeatPromptsNameNoVerb`): the new prompt text says "a reading that finds the source against the
  sentence records its gap in the same act", never `--class`.
- AFTER regeneration, every changed golden line and each affected `--help` page (verify, corroborate,
  finding, mint, dispatch, show) is read before and after, one command at a time, for semantic loss.

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
  `0c57758e8`: `grep -rn 'verbOfEvent\|EVENT_TYPE_MINT\b' tools/internal tools/devcmd`, no tests and no
  goldens, every hit read:
  - `seatprobe/seatprobe.go:153, :212` — REWRITE. A pass over the seat's events first: the `gap_id`s a
    Verify of this seat names with `carried_by` and `carrier_minted` true. `Read` skips a Mint of those
    ids, read from the fields, never from adjacency. A plain mint a later reading names (`carrier_minted`
    false) still counts as `mint`. The Anchor's comment (`:207-208`, "it places the marker of the act
    before it, whose own event names the verb") [REWRITE]: an Anchor places the marker of the act it
    belongs to, whose own event names the verb; a reading's Anchor stands between its Mint and its Verify.
    A Verify counts under the verb its body names: `verify` where it has an `anchor`, `corroborate` where
    it has none, read in `Read` as `motionSubject` reads a motion's. Today every corroboration counts as
    `verify`, so `boards.go:535`'s `corroborate` expectation cannot be met, and the reading's property
    holds for neither verb without this.
  - `seatprobe/boards.go:532` (the lens-audit board's `mint` expectation) [REWRITE] its `Because`: it
    argues for a defect no source reading carries, since a defect a reading finds against a source is
    minted by the reading's act and counts under the reading's verb. `:534`, `:535`, `:563` KEEP.
  - `record/viewjson.go:1606, :1918`, `record/evidenceview.go:586` (a narrowed view's event families),
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

1. **A finding about a contradiction** becomes a finding about the reading's citation: `about_kind
   anchor`, `about_ref` = the anchor of the reading its `about_ref` named. Two steps, because the
   archived word cannot reach replay as itself:
   - **Decode.** An archived event's enum words are resolved at decode (`registry.go:172` →
     `valuerenames.go` `enumValue` → `enumNumberOf`), before any replay translation runs, and the live
     `AboutKind` has no `contradiction` value once §III.1 deletes it — so without a row the archived
     finding is refused at decode and replay never sees it. [MODIFY] `migrate/valuerenames.go`: the
     `feov.record.v1.AboutKind` row gains `"contradiction": "anchor"`, beside `"proof": "anchor"`, with
     a comment in the same voice ("a finding about a reading that found the source against its
     sentence is a finding about that reading's citation"). The archived finding then decodes as
     `about_kind anchor` with its archived `about_ref`, a reading's `events.id` (e.g. `"401"`).
   - **Replay.** `replay.go:395-406` can no longer key on `ABOUT_KIND_CONTRADICTION` (deleted). It
     recognises the raise by its reference: a Finding whose `about_kind` is `anchor` and whose
     `about_ref` names no anchor — `anchor.Kind(ref) == ""` (`internal/anchor`, the same test
     `record.go:944` puts to an Anchor's id) — is a raise. Never a regex over the ref: a shipped-tool
     regex is ruled out (gblock, 2026-10-18), and the anchor package is the one reader of what an id
     is. The ref is parsed as an `events.id` (`strconv.ParseInt`, as `contradictionRef` does), carried
     through `newID`, and the Verify the migrated record holds at that id is read through the batch's
     querier (`batch.Querier()`); `about_ref` becomes its `anchor`. A ref that is no `events.id`, an id
     the replay wrote nothing for, or one naming a Verify with no anchor (a **corroboration** reading)
     is refused with the finding's id named — no archive holds one, so the refusal is the loud miss,
     not a silent zero. A Finding about an anchor whose ref IS an anchor id (`C-…`, `K-…`, including
     every archived `proof` respelled to `anchor`, whose `P-` ref `remap.apply` has already respelled
     `K-` with the same hex, `replay.go:303`, before this step) is not a raise and is untouched.
     `requireAboutAnchor` (`refs.go:126`) then admits the citation anchor.
   Both gates: `TestEveryArchivedRunMigrates`, and the m30 test's `about anchor C-4f4b8cfa` assertion
   below; a delete of the `valueRenames` row turns both red (a decode refusal on m30 and m32).
   `TestValueRenames` (`valuerenames_internal_test.go:15`) holds the row to its three conditions (a
   real enum, an old word the live lookup refuses, a new word it accepts). The row's readers are every
   decode of an `AboutKind` field: `Finding.about_kind` (the raises) and `Mint.about_kind`, which no
   archive holds as `contradiction` (the mint refused that kind when it was live, `record.go:966-971`),
   so the respelling reaches findings only, and the replay test above is the one place it is
   interpreted.
2. **An archived negative reading** carries no `carried_by`: its gap was minted later (forward reference) or
   never. migrate does NOT fill `carried_by` from the `found_by` chain (ruled, gblock 2026-10-30). Its
   found-against state is still a failed check on the migrated record (read from the ledger), with no
   named carrier.
3. **A mint whose `found_by` names a raising finding** is carried unchanged — the finding exists.
4. **A stored plan's `contradiction` blocker** becomes a `failed_check` blocker whose subject is the
   reading's anchor and whose owner is unchanged (the reader, as a live found-against blocker's is).
   `TranslatePlanKinds` cannot do it: it is a pure kind rename (`planKindRenames`) with no record
   access, and the archived subject is the reading's `events.id`, not an anchor. So [NEW]
   `TranslateReadingBlockers(pl, newID, q)` in `migrate/planids.go`, called in `replay.go:381-394`
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
5. Migrate's comments in `entries.go:64-66` (the credit entry) and `outcomes.go:19` [REWRITE] to state
   archived bytes without the live concept word where they describe the live model.
6. **A record written at this epoch replays its reading acts whole, and every source event of the act keeps its own duties (GAP 3; round-6 GAP 1).** Replay appends one event
   at a time (`replay.go:452`, `batch.Append`), and three write rules of this plan read the reading's
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
   (`replay.go:222-505`) owes every source event a set of duties, and a reading act's three members
   keep each one, in their own iterations, as any event does. Swept — every per-source-event step of
   the loop, each read at `0c57758e8`:

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
   whole (as `Append`'s does), and a `[]record.WrittenBody` of the events written. Name grep at `0c57758e8`:
   `AppendBodies`, `StampedBody`, `WrittenBody`, `readingActs` 0 hits; `appendAct` (one body's write) is the sense it
   composes, not a rival.

   Property: every source event of a reading act maps through `newID` to the event written from it
   (Mint→Mint, Anchor→Anchor, Verify→Verify), each is written and counted once at its own source
   stamp, an act is written whole or each member refused, and a record written at this epoch migrates
   with 0 refusals. `TestARecordWrittenAtThisEpochMigratesWithNoRefusal` (tests below) drives it; its
   partners (§V.5.y) turn it red.

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
- m32's four raising findings become four findings about `C-a36dce3f`. m32 ended UNVERIFIED, so no
  verdict moves.

Tests [NEW / REWRITE], each driven through `migrate.Migrate` on the real archive, as `feov-migrate`
runs it:
- `TestTheMigratedM30RaiseIsAFindingAboutItsCitation` (`migrate/`): F-8410bec7 → `about anchor
  C-4f4b8cfa`; G-dc877265 open/closed as archived; 0 refusals.
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
  record before event 400, the same `corroborate` with the gap flags a negative reading needs
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
  whose `gap_id` is G, the Anchor whose `id` is G, L's Verify). As specified before round 6 — L, M and
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
  `cli/m23contradiction_test.go` is rewritten to assert the migrated state the GAP 2 ruling accepts
  and renamed — the file `cli/m23corroboration_test.go`, the test `TestM23sArchivedCorroborationHoldsNothing`
  (§III.9), neither carrying the census word; its live quote-refusal half stays, driven with the gap
  flags a negative reading now needs.
- REWRITE the rest per census; each rewritten assertion tests the new behaviour, not the absence of the old.
- NEW (each with its delete-the-row partner in §V.5):
  - `TestARefutingReadingMintsItsGapInTheSameAct` (verify and corroborate; class from the admitted set;
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
    unchanged after; a plain `mint` at that cap is still refused. Partner §V.5.e.
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
    The same for `corroborate --as refutes`; `corroborate --as absent` twice meets the duplicate-key
    refusal and the record holds one Verify and no Mint. Partner §V.5.m.
  - `TestACorroborationAbsentMintsNothing` (F5).
  - `TestAReadingsGapIsFilledAsAPlainMints` (`cli/mintreason_test.go`, driven through `mint`, `verify` and
    `corroborate`): one lens gives the same judgment flags and `--reason` to a plain `mint --quote S1`, a
    `verify --anchor C --as refutes` of S2 and a `corroborate --as refutes --quote S3`, and the test reads
    the three stored Mints. It walks `Mint`'s descriptor (`ProtoReflect().Descriptor().Fields()`): for
    every field, presence and value equal the plain mint's, or the field is in the test's two named
    sets — differs by subject (`gap_id`, `location`) or by presence only (`fix_new`) — so a field added
    to `Mint` fails until it has a disposition here and in §III.2's table. Rows: no `--new`, no
    `--complexity`, no `--fix` (each absent alike, `fix_basis` `proposed`); all three given (`fix_basis`
    `verified`, `fix_new` set). Partners §V.5.ee, ff.
  - `TestAReadingsGapCarriesItsArgumentWhereverAMintsDoes` (`cli/mintreason_test.go`, beside
    `TestAGapCarriesRedsArgumentAndNotOnlyItsProblem`): lens L's `verify --anchor C --as refutes --reason R`
    with no `--new` mints G. Asserts: `show board --json`'s row for G has `mint_reason` R and `fix_basis`
    `proposed`; blue's `show work --json` item for G has `fix_basis` `proposed` (never omitted); the gap
    view's row agrees; `show report` holds `Token(G)` at the sentence's end and not R, the live text
    carrying no argument; `report.Assemble` over that run — what the bench's `assemble` verb runs
    (`cli/bench/assemble.go:43`), the one path to the report's gap block — holds "minted because: R" and
    "fix_basis: **proposed**" in G's block. After L's own backing re-read in a later sitting and its
    close, `report.Assemble`'s closure index holds "minted because: R" under G. Row: lens M's identical
    reading names G with `--reason R2`; M's Verify holds R2, and G's `mint_reason` stays R. Row: `verify
    --as refutes` with every gap flag and no `--reason` is refused naming `--reason`, and the record holds
    no Verify, Mint or Anchor from it. A `corroborate --as refutes` row repeats the first asserts. Partner
    §V.5.ee, ff.
  - `TestAReadingIsAttributedToTheVerbTheSeatTyped` (`cli/`, drives the verbs, then `seatprobe.Read` and
    `seatprobe.Check` over the run): lens L's refuting `verify` → `Used["verify"]` 1, `mint` absent from
    `Used` and present in `Unused`, and an expectation of `mint` is unmet, with the substitute `verify`;
    lens M's refuting `corroborate` → `Used["corroborate"]` 1, `Used["verify"]` 0, no `mint`; L's later
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
    today's unread sentences ("stays unread", "you owe a gap about it"); (c) `verify --as absent` →
    refused by the existing body-less-copy refusal, nothing recorded; (d) `corroborate --as refutes`
    with the gap flags → refused naming `not_the_work`; (e) the same without them → one Verify, no
    Mint, no Anchor, no carrier; (f) `verify --as unreachable` with the gap flags → refused. A seventh
    row writes, at the record layer outside `Migrating`, a `not_the_work` `refutes` Verify carrying
    `carried_by` and one carrying `carrier_minted` → each refused by the write-path rule. GAP 1 rows
    (RULED: every reading of a `not_the_work` copy is unread, a backing one included): (g) `verify --as
    supports` of the `not_the_work` copy → one Verify, its stored `source_completeness` `not_the_work`
    (the stamp `verify` now writes), no Mint, no Anchor, no label; the citation's failed check has state
    `unread`, owner the reader, and the human text carries "stays unread"; (h) `corroborate --as
    supports` of a `not_the_work` copy → one Verify with no `label`, no Anchor, no footnote —
    `record.CitedSources` and `show evidence`'s sources hold nothing for it — and the same seat's
    `corroborate --as supports` of ANOTHER claim from that url in the same sitting meets the
    duplicate-key refusal on the url (no label to key on), nothing recorded; (i) `verify --as supports`
    of a cached `full` copy and of an uncached source → `source_completeness` `full` and `not_asked`.
    Partners §V.5.s and §V.5.w.
  - `TestAVerifyOfARepeatedSentencePlacesItsGapByTheAnchor` (`cli/`, driven through `verify`,
    `corroborate`, `show board` and `blue edit --accept` as the seats run them): a report holding one
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
    (h) `corroborate --as refutes --quote <the repeated sentence>` (with the gap flags) is refused naming
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
    --accept`, `verify`, `corroborate` and `show work` as the seats run them): sentence S, cited by C-1
    whose token ends it; lens P mints G0 with `--quote S --new S′` (verified), where S′ holds C-1's token
    as `show report` prints it, and blue's `edit --accept G0` applies it verbatim. Rows: (a) control —
    lens L's plain `mint --quote S′ …` is refused as an estoppel naming G0, and one `LOG_TYPE_ESTOPPEL`
    entry is written; (b) L's `verify --anchor C-1 --as refutes --class claim_contradicts_source …` is
    admitted: Mint G1, Anchor `{G1, sentence_of: C-1}`, Verify `carried_by` G1, `carrier_minted` true;
    `record.EstoppelRejectionsOf` is unchanged from after (a); G1 is open; `PassBlockers` holds C-1's
    `failed_check`, `CheckState()` found_against, `AnsweredBy()` G1, and the chair's PASS is refused; (c)
    lens M's `corroborate --as refutes --quote S′ --class within_source_condition_misattribution …` (a
    class G1 does not hold) is admitted, mints G2, writes no estoppel entry, and G2 is a material open gap
    holding the PASS. Partner §V.5.cc.
  - `TestTheReadingVerbsDescribeTheirGapFlagsByTheirSubject` (`cli/`, renders `verify --help`,
    `corroborate --help` and `mint --help`): `verify`'s page names no `--quote` and no `unclassified`, its
    `--new` names "the sentence your --anchor's token sits in", its `--class` lists exactly the three
    classes `verify`'s outcomes admit and a class outside them is refused at the parse; `corroborate`'s
    `--new` names `--quote` and its `--class` lists the two `refutes` admits; both pages' `--reason` opens
    with its `ReasonIs` text and names the gap's argument, and the seven required flags open "REQUIRED
    unless"; `mint`'s flag text is byte-equal to the manual golden's. Partner §V.5.dd, hh.
  - `TestEveryNegativeOutcomeNamesItsClasses` (the `gap_classes` facet is non-empty exactly on
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

- Delete the contradiction drives and census fields: `answerContradictions` (`fuzz_test.go:2494-2511`, called at `:5278, :5575, :8873`, named in `dischargeOwed`'s comment `:1117-1122`), the
  `ownAnswers`/`answerBudgetErr` fields (`:499-502`), the sweep gates at `sweeptally_test.go:913, :947`
  and every other site the census lists.
- **The main reading drive** (`fuzz_test.go:5393-5478`; the raise/`contradictionOf` half at
  `:5400-5412` is deleted with the census). A lens's reading draws the gap flags (class from the
  outcome's admitted set, grades, check) exactly when the reading will find against: the outcome has
  `gap_classes` AND the copy the reading resolves to is not `not_the_work` — the drive reads that from
  the run's cache entry for the source (`fetchcache` lookup by url, the entry the verb stamps from),
  never from the outcome alone; a corroboration's `absent` draws none (F5). After the act it checks
  the result against the stored Verify: `gap` present iff `carried_by` is set; and where
  `carrier_minted` is true, the stored Mint's `mint_reason` equals the `--reason` the drive passed and its
  `fix_basis` is `verified` where `--new` was passed and `proposed` where it was not. Any mismatch is a
  `noteActErr`.
  - **Its found-against repeat** (round-4 note). The drive re-reads the same anchor across sittings
    with a random outcome, and its one repeat guard today is `ownUnreadGapOpen` (`:3719`, `:5416-5420`,
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
    `corroborate` of that source expects no label and no footnote in `show evidence`'s sources — a
    mismatch is a `noteActErr`.
  - **A cited sentence the report holds twice** (F9). The fuzz's report stock gains one sentence whose
    words occur in two sections, each copy cited by its own citation. A lens's refuting `verify` of
    either citation is expected to mint (or name) a gap whose token renders at the END of that
    citation's sentence — after its last content character and any closing quotation mark, and BEFORE
    a citation token that ends the sentence (§III.2a: `of`'s token at the sentence's end stays after the
    gap's), in the occurrence holding that citation's token and in no other — the drive reads the
    rendered report — counted `repeatedSentenceGaps`; a
    `corroborate` whose `--quote` is that sentence is expected refused as ambiguous. Where the reading
    passed `--new`, blue's next sitting accepts that gap's fix by coin and expects it applied to the
    cited occurrence only, recorded with `exact_span` (counted `repeatedSentenceAccepts`).
  - **A reading over prescribed text** (F10). Where blue has applied a red `--new` verbatim (the
    estoppel drive's precondition, `fuzz_test.go:658-668`), a lens's negative reading of that text — a
    `verify` of a citation the accept carried onto it, or a `corroborate --quote` of it — is expected
    ADMITTED with its gap minted or named, and NO new `LOG_TYPE_ESTOPPEL` entry; an estoppel refusal of
    a reading is a `noteActErr`. Counted `readingsOverPrescriptions`.
- New gates (false-green guards, zero fails the sweep), each counted from the stored record, never from
  `--as`: readings that minted (`carrier_minted` true); readings refused a mint by the one-gap rule
  (`carried_by` set, `carrier_minted` false — the duplicate path is driven on purpose by re-reading a
  sentence already holding a reading's gap); found-against closes refused; `foundAgainstRepeats`;
  `notTheWorkFlagsRefused` and `notTheWorkReadingsUnread`; `repeatedSentenceGaps` and
  `repeatedSentenceAccepts`; `readingsOverPrescriptions`. Each new field is
  carried through `sweepTally`, the fold and its merge (the sites listed below for the found-against
  fields). New tally: max open gaps per (class, sentence) — must be ≤ 1.
- **Every fuzz branch that tells failed checks apart keys on `CheckState()`**, and the `unread*` tallies
  count the `unread` state only. Sweep, run at `0c57758e8` (80 hits, `new-fz.txt` beside the other lists):
  `grep -n -E 'BlockerFailedCheck|AnsweredBy\(\)|\.Detail\b|unread(Answers|Repeats|Retires|RetiresJoined|ClosesRefused)|unsound(Answers|Repeats|ClosesRefused)' releasegate/fuzz/*.go`,
  every hit read. Under a found-against check `Detail == ""` reads as "unread citation", so each of these
  would drive or count it as unread:
  - `fuzz_test.go:1218, :1227` (`fixFailed`, blue's remedy): `Detail == ""` → unread; it would call
    `envelopeRefusedEarly(… "check of citation "+C)` expecting the unread wording, retire the citation and
    count `unreadRetires`/`unreadRetiresJoined`. [REWRITE] a switch on the state: `unread` as today;
    `found_against` drives blue's found-against remedy — by coin, the edit that rewrites the sentence
    (`reopened` names C) or the retire `--answers` the carrying gap — with
    `envelopeRefusedEarly(… "reading of citation "+C+" at event")`, the found-against `failedFact`'s
    words, and counts `foundAgainstEdits` / `foundAgainstRetires`; `unsound` as today; any other state is
    a `noteActErr`, never a fall-through;
  - `:3007` (unread repeat: `verify --as unreachable` expected refused): `CheckState() == unread` only —
    an `unreachable` reading over a found-against check is admitted, so driving it as refused would
    fail the run; `:3017` (unsound repeat): `== unsound`;
  - `:3039, :3042` (the owed-gap mint, `citation := Detail == ""`, its early-envelope refusal and the
    `unreadAnswers` count): only `unread`/`unsound` with no carrier — `found_against` is never owed;
  - `:3096-3105` (`driven := unreadClosesRefused > 0`, the drive of a close refused at the mint) and
    `:4249-4273` (`refuseStandingClose`'s tally, `Detail == ""` → `unreadClosesRefused`): the tally
    switches on the state — `unread`, `unsound`, `found_against` → `foundAgainstClosesRefused` — and
    any other state is a `noteActErr`;
  - `:3415` (blue's oracle, "Retire the citation" for every `Detail == ""`): `unread` only; a
    `found_against` check's item must name the found-against remedy ("edits the sentence, or retires or
    replaces the citation");
  - `:3719` (`ownUnreadGapOpen`): `== unread`;
  - `:4192, :4237` (`standingFailedCheckOf` → `refuseStandingClose` in `closeGap`): state-agnostic
    selection by carrier, kept, and now the successor-following `AnsweredBy()` (§III.3); the tally it
    reaches switches as above;
  - KEEP: `:2953` (asserts the plan readies the owner or blue, true of every state), `:6974` (no
    VERIFIED over any failed check);
  - `sweeptally_test.go:834-845`: the `unread*` gates stay and now count only the unread path; NEW
    gates (zero fails the sweep, false-green guards) `foundAgainstClosesRefused`,
    `foundAgainstEdits + foundAgainstRetires`, and `foundAgainstClearedByOwnReading` (the drive below);
    each new field is carried through `sweepTally`, the fold and its merge (`sweeptally_test.go:186-187,
    :531-533, :651-653`, `fuzz_test.go:468-470, :5713-5715, :7041-7043`), which
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
    `found_by`), `none`;
  - `mints_by_epoch` / `closes_by_epoch` — mints and closes per epoch.
- Release gate: PR CI only. Never `FEOV_RELEASE_GATE=1` locally.

### III.12 Consumer census

Searches, run at `0c57758e8` over the tree as committed (raw hits kept at
`~/.claude/scratch/plan1647-r8/census-new.txt`, per-file counts in `cnt-new.txt` beside it):

```
git grep -I -n -i contradict 0c57758e8 -- . ':!run-archive'
git grep -I -n -E 'ABOUT_KIND_CONTRADICTION|unanswered_contradictions|ContradictionOfReading|raisedBy|answeredBy|raised_by|opened_by|sources_refuted' 0c57758e8 -- . ':!run-archive'
```

Result: **207 files**, 1,596 hits, each read for meaning. Against `a31a7b06d`'s 206 files, per-file counts
differ in three places (`cnt-old.txt`): `agents/blue-lane.md` is new (7 hits: the inlined manual's
`show evidence` page and ordinary English, REGENERATE); the manual golden moves from 11 hits to 17 (the
lane's page, REGENERATE); `scripts/agentgen/src/blue-researcher.md`'s one hit moved to
`fragments/blue-researcher-frame.md:40` ("contradicted", ordinary English, KEEP).
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
Dispositions (a file counted once, MODIFY counted with REWRITE): **DELETE 4 · REWRITE 87 · REGENERATE
17 · KEEP 121** (45 of the KEEP are historical).
Paths are relative to `plugins/frank-exchange-of-views/tools/internal/` unless shown. "n" = lines
carrying the FEOV sense, at `0c57758e8`.

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
| record/about.go | 5 | about-kind validation, gap refusal | REWRITE |
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
| record/migrate/plankinds.go | 0 | `planKindRenames`, `TranslatePlanKinds` | KEEP: unchanged. `ReadPlan` decodes a blocker's kind as a string (`planstore.go:52-63`), so an archived `contradiction` blocker passes `TranslatePlanKinds` untouched to `TranslateReadingBlockers` (`planids.go`, §III.9 item 4) |
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
| cli/seat/help/corroborate.md, verify.md, dispatch.md, mint.md, finding.md | 2, 2, 3, 2, 1 | contradiction prose (finding.md l.9 "internal contradiction" stays) | REWRITE |
| cli/lens/mint.go | 0 | ordinary English | MODIFY (shared flag helper, §III.2; `verifiedFix` places a `verify`'s gap with `AttachToSentenceOf` and names the verb's subject in its refusal, §III.2a) |
| cli/seat/verbs.go (placement) | 0 | `Places`, `AppendPlaced` (`:1384-1417`) | MODIFY (`PlacesToSentenceOf`, §III.2a) |
| cli/seat/seat.go | 0 | `ReasonIs` (`:1082`), ordinary at the census hit | MODIFY (`verify` and `corroborate` rows, F11) |
| cli/seat/help/record.md, cli/bench/outcome.go | 0 | inquiry `refuted` / ordinary | KEEP |

**Placement and the accept by identity (§III.2a; none carries the word):** `anchortext/anchortext.go`
(`AttachToSentenceOf`, two sentinels) NEW; `reportproj/render.go` (`sentenceOfMut`, `mutationOf`)
MODIFY; `record/recordsql/views.go` (`report_op`'s `'sentence_of'` arm and comment) MODIFY;
`record/queries.go:251-256` (`ReportOp` doc) MODIFY; `record/recordsql/testdata/schema.sql`
REGENERATE; `record/record.go:664-668, :941-951` MODIFY; `record/recordpb/record.proto:1179-1187`
MODIFY; `record/onegapperclass.go:82-92` MODIFY; `record/estoppel.go:175-211` (`ProposalOver` by the
gap's token) MODIFY; `cli/blue/edit.go:335-349` (the accept's exact span) MODIFY.

**Other Go** — `terms/terms.json` (5: the entry, its ordinary-English collision, `lens retirement`): REWRITE
then vocabdoc. KEEP (0 FEOV-sense hits; source-outcome, class or ordinary sense): capture/capture.go
(`sources_refuted`, found-against check), capture/computationrerun.go, report/assemble.go,
report/registrations.go, verify/verify.go, seatprobe/production.go, bluedoc/bluedoc.go, flags/names.go, feov/error.go, seatenv/identity.go,
seatenv/seatenv.go, tessocr/engine_stub.go, `scripts/decisions/main.go`,
`scripts/golden/review.go`, gray-area `cmd/gray-area-capture/main.go`, `internal/claims/coverage.go`,
`internal/claims/manifest.go`. REWRITE beyond the word: `seatprobe/seatprobe.go`, `seatprobe/boards.go`
(§III.8; `boards.go`'s census hit, `cross_section_contradiction`, stays).

**Tests**

- DELETE (4): record/contradictionref_test.go, record/contradictionquadratic_test.go,
  record/answerbudget_test.go (its one test is the own-raise exemption),
  `tools/integration/surface/contradictionanswer_test.go`.
- REWRITE (22): record/redcitation_test.go (3 of 7 tests), blockerowner_test.go, passblockers_test.go
  (incl. the `contradictingVerify` helper), envelope_test.go, narrowedview_test.go, reopenedm19_test.go,
  planidphrases_test.go, export_test.go (`ContradictionReadiness` export), satthrough_test.go,
  lensowes_test.go, migrate/reproduceanchor_test.go, migrate/planids_internal_test.go;
  cli/verifyclaim_test.go, corroborate_test.go, m23contradiction_test.go (the quote-refusal half stays),
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
  passagreement_test, registeronce_test, roster_test, spotcheck_test,
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

**debate.js** (`skills/research-protocol/scripts/debate.js`): l.393 and l.415 sit inside the
`GENERATED ENVELOPE VOCABULARY` region (l.371-438) → REGENERATE (enginegen); l.889 (red-chair prompt, "the
lens that read a contradicting source no finding raises") is authored → REWRITE; l.227, 253, 255 KEEP.

**Agents and skills** — REGENERATE (agentgen): all 12 `agents/*.md` (7 red-lens-*, red-chair, blue-lane,
blue-researcher, blue-synthesizer, lead-judge). REWRITE: `skills/research/SKILL.md` l.7,
`skills/research-protocol/references/run-directory.md` l.76. The two generated preloaded skills have no hits.

**agentgen sources** (`scripts/agentgen/src/`) — REWRITE: fragments/lens-audit-and-mint.md,
fragments/lens-duties.md, red-chair.md. KEEP: lead-judge.md, fragments/blue-researcher-frame.md,
fragments/blue-research-protocol.md (ordinary English).

**The `lane` role (#1663).** Is it a carrier of the concept? Swept at `0c57758e8`: `seatclass/roster.go:99`,
`cli/blue/command.go:40-47` (`LaneVerbs`), `record/envelope.go:107`, `agents/blue-lane.md` as rendered,
`scripts/agentgen/src/blue-lane.md` and its fragments.

| Site | Reads or teaches | Disposition |
|---|---|---|
| The lane's verbs (`LaneVerbs`: register, ingest, inquiry, log) | no `verify`, `corroborate`, `mint`, `close`, `edit` or `retire`: a lane performs no reading, mints no gap and is refused as the wrong seat on each | KEEP |
| Its work list (`WorkArrivesAsItems`, `envelope.go:107`) and the plan | work reaches it by dispatch, never as items, so it holds no failed-check item; a found-against check readies `blue-respond` alone (`dispatch.go`'s arm, §III.3) | KEEP |
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
sentence the citation's token sits in; a verdict that reads nothing — unreachable, or a not_the_work
copy — mints nothing, and the unread citation's lens owes the gap about it), l.66-67 (the `lens mint` row: the owed failed-check gap is the unread or unsound
check's; the `lens close` row: a gap carrying a found-against check is refused too, until the reader's
own backing reading, an edit or a retire), l.86 (the `blue retire` row: an anchor found against is
blue's to edit, retire or replace).
KEEP: propagation-and-anchoring.md, seat-surface-naming.md, why-a-seat-stops.md; outside FEOV,
`plugins/gray-area/README.md`, `plugins/gray-area/skills/audit-seat-coverage/SKILL.md`,
`plugins/prosthetic-conscience/skills/resume/SKILL.md` (ordinary English).

**Phrase-sweep additions** (old model, no word; §III.7): `scripts/agentgen/src/skills/research-protocol/SKILL.md:17-19`
→ REWRITE (regenerates `skills/research-protocol/SKILL.md`); `cli/lens/verify.go:304, :701-704` → REWRITE.
`scripts/agentgen/src/lead-judge.md` → REWRITE (one added sentence, R4).

**Plans / ideas / law / feov-memory / research** — KEEP-historical (45 files: 33 under `plans/`, 7
under `research/`, 2 under `ideas/`, 2 under `feov-memory/`, 1 under `law/`). FEOV-sense mentions in
plans/reopened-evidence-is-owed.md, feov-lens-bar.md, gap-class-enum.md, pre-dry-run-batch.md,
one-item-per-dispatch.md:336, ocr-cite-derives-page.md:574, historical/red-citations.md,
historical/migrate-staged-publish.md:284, law/precedents.md:156 record past decisions; none is a live
instruction, none is rewritten.

**Readers of a refutation that do not name the contradiction** (added at the Slim-fast peer's request;
each states how it reads once a refutation is a gap):

| Site | Today | After | Disposition |
|---|---|---|---|
| Anchor ledger `anchorStanding.state()` (`anchorchecks.go:152-178`, from #1601 / 90f3fa008) and `stateFoundAgainst` | a citation whose standing check refutes or is absent; its PASS hold was "the contradiction's" | Unchanged predicate: found-against is still the **citation's standing** (a standing negative reading of the sentence the report carries), NOT "has an open gap of the new class". The gap is a separate fact linked by `Verify.carried_by`; the state is what makes the gap unclosable (§III.3). Keying the state on the gap would let a gap closed into a successor, or an archived reading with no gap, read as holding | MODIFY (comment only; `contradicts()` → `gap_classes` facet) |
| `failed_check` PASS blocker (`passblockers.go:403`, `failedcheck.go`) | unread citations and unsound computations | gains the found-against state (§III.3): holds the PASS until its reader's own later reading backs the sentence, blue edits it, or blue retires the citation (another lens's backing reading clears nothing); readies blue alone; carrier = `Verify.carried_by` followed through its successors | MODIFY |
| `unchecked` PASS blocker | a citation or computation with no check of its current sentence | unchanged: an edit of a found-against sentence makes the citation unchecked, owed a fresh check by its owning lens; that check either holds (failed check cleared, the gap may close) or is negative again (names the standing gap, or mints one if it is closed) | KEEP |
| Evidence lens's gap duties (`fragments/lens-duties.md`, `lens-audit-and-mint.md`, `verify.md`) | read every citation; a negative reading opens a contradiction it must raise by finding, then answer by a gap minted from it, outside its budget | reads every citation; a reading of the work that finds the source against the sentence carries the gap's judgment (class among the outcome's admitted set, problem, check, grades) and its argument (`--reason`) and mints it in the same act, outside its budget, at the sentence its citation's token sits in; a verdict that reads nothing (unreachable, a not_the_work copy) mints nothing; a second reading of a sentence holding that gap names it; the lens closes the gap with its ordinary close once its own later reading backs the sentence, or blue edits or retires it. Unread readings keep the owed-gap-about-the-anchor duty | REWRITE (agentgen) |
| Failed-check owed-ness readers (§III.3): dispatch lens arm `dispatch.go:558-565`, lens work item `sitting.go:401-408` (and `EnvelopeRefusal`, `envelope.go:44`, through it), retirement `retirement.go:164-172`, budget `failedcheck.go:109-119, :208-215` | owed iff `answeredBy == ""` | owed iff state is `unread`/`unsound` and no gap carries it; `found_against` never owed by its reader | MODIFY |
| Failed-check printed sentences (§III.3): `passblockers.go` `failedFact` :226, `failedClears` :234, `failedRemedy` :245, `failedOwed` :254, `blueFailedItem` :262, the kind's gloss/refusal/why :404-433; `dispatch.go:1326` `failedCheckReadiness`; `failedcheck.go` repeat and close refusals :135, :181 | branch on `Detail` (a computation's `bundle_sha`) and `answeredBy`; citation = unread | branch on the state; found-against wording names the reading and blue's remedy (edit, retire or replace), never an owed mint | MODIFY |
| Failed-check prose: `help/mint.md:15`, `help/close.md:11`, `help/verify.md:11`, `help/reproduce.md:21`, `fragments/blue-search-and-cite.md:22-29`, `fragments/lens-audit-and-mint.md:6`, `terms.json` `failed check`, `docs/citations-and-bibliography.md:158-167`, `docs/seat-command-triggers.md:56, :66-67, :86` | two states; the seat whose check failed owes the gap; "the gap does not close while the check stands" | three states; the owed gap is the unread/unsound check's; a found-against check's gap is minted at the reading, and the check clears by the ledger's rule; the seat's close is refused while the check stands, and a closing bench ruling closes the gap while the check holds the PASS (F8) | REWRITE |
| Fuzz failed-check drives and tallies `fuzz_test.go:1218, :1227, :3007, :3017, :3039, :3042, :3096-3105, :3415, :3719, :4192, :4237, :4269`; gates `sweeptally_test.go:834-845` | `Detail == ""` = unread citation; `AnsweredBy() == ""` = owed; unread tallies count every citation check | key on `CheckState()`; unread tallies count `unread` only; found-against has its own tallies and gates | REWRITE (§III.11) |
| `verify.go:529-541, :631-656` `failedCheckWords` and the JSON `failed_check` | guarded by the NEW reading's `!VerifyReads`, worded as unread whatever the check's state; JSON carries no state | after every verification of a citation, every standing failed check of the anchor, worded by its state; JSON `failed_checks[].state` | REWRITE (§III.2) |
| `verdict.go:141-162`; `debate.js:1050-1057`; `fuzz_test.go:2953, :6974` | anchor names only; owner-or-blue readied; no VERIFIED over a failed check | unchanged, true of every state | KEEP |
| Tests pinning unread/unsound sentences: `unreadcitation_test`, `unreadb6_test`, `blueanswered_test`, `failedcheck_test`, `reopened_test`, `tiersubstitution_test`, `unsoundm32_test`, `capture/dispatchparity_test` | those states' wording | their owed-mint and remedy wording does not move — they are the regression net for it; the "does not close" comments in `unreadcitation_test.go:119` and `unreadb6_test.go:21-22, :66` are rewritten (F8) | KEEP (two comment REWRITEs) |

**Readers of a failed check's owed-ness and its printed sentence** (the class the round-2 audit named;
run at `0c57758e8`, every hit read, `new-fc3.txt`; dispositions in the table above):

```
grep -rIn -E 'answeredBy|answered_by|AnsweredBy' plugins/frank-exchange-of-views/tools/internal plugins/frank-exchange-of-views/tools/cmd plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js scripts/agentgen/src | grep -v _test.go
grep -rIn -E 'BlockerFailedCheck|failedChecksOf|answersOwnFailedCheck|ownFailedCheckAnswers|failedRemedy|failedOwed|blueFailedItem|failedCheckReadiness|failedCheckWords|failed check' <same paths> | grep -v _test.go
grep -rIn -i -E 'check (of yours|that) failed|whose check failed|failed check|about the anchor|about that anchor' scripts/agentgen/src plugins/frank-exchange-of-views/tools/internal/cli/seat/help plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js
```

Hits outside the contradiction fold (which is deleted: `citationid.go:554-685`, `onegapperclass.go:145`,
`mintbudget.go:362-380`): `passblockers.go:84-98, :223-274, :403-433, :714`; `failedcheck.go:44-119,
:134-215`; `sitting.go:321-322, :401-408`; `retirement.go:142-172`; `dispatch.go:22, :558-573,
:1323-1341`; `mintbudget.go:308-309, :370, :394, :400-410`; `verdict.go:141-162`; `cli/lens/verify.go:529-541,
:631-656`; `cli/seat/verbs.go:308`; `terms.json:1079-1080`; `help/mint.md:15`, `help/close.md:11`,
`help/verify.md:11` (unread; its "does not close" clause REWRITEs, F8), `help/dispatch.md:19`; `fragments/lens-audit-and-mint.md:6`,
`fragments/blue-search-and-cite.md:22-29`; `debate.js:393` (enginegen, from `dispatch.go:22`),
`debate.js:1050-1057`; `onegapperclass.go:28` (a comment naming `answersOwnFailedCheck`, KEEP);
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
`keyFields` (neither name is on it, so a Verify keys on `label`/`anchor`/`url` as today; the reason the
field is not called `gap_id`, §III.1), `flags/names.go:402` `payloadFlag` (falls back to `--carried-by`,
never printed: both fields are tool-supplied, `seat.Supplies` in `verify.go` and `corroborate`), and the
required-field sweep (`recordpb/requiredfields_test.go`, both optional); by meaning — `failedChecksOf`
(carrier), `requireMintWithinBudget`/`lensMintStanding` (`carrier_minted`, §III.5), the repeat rule
(§III.2), the reading's result (`gap.minted` read back), migrate (archived readings carry neither,
§III.9.2), the sweep tally `open_by_found_by` (§III.11) and §V.8's query. `references: "mint.gap_id"`
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

### III.13 Directory tree of touched paths

```
plugins/frank-exchange-of-views/
  requirements.json                         [MODIFY] eventSchema 53
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
    migrate/*_test.go                       [NEW] m30, m23, a record at this epoch, sentence_of check
    *_test.go                               [NEW] m32 cut, migrated found-against checks
  tools/internal/anchortext/anchortext.go   [MODIFY] AttachToSentenceOf
  tools/internal/reportproj/render.go       [MODIFY] sentenceOfMut
  tools/internal/seatprobe/seatprobe.go, boards.go   [MODIFY] a reading's Mint counts under its verb
  tools/internal/cli/seat/verbs.go          [MODIFY] PlacesToSentenceOf
  tools/internal/cli/seat/seat.go           [MODIFY] ReasonIs rows for the reading verbs
  tools/internal/cli/lens/verify.go, mint.go, finding.go   [MODIFY]
  tools/internal/cli/blue/edit.go           [MODIFY] the accept's exact span
  tools/internal/cli/m23contradiction_test.go → m23corroboration_test.go  [REWRITE] the migrated state (GAP 2)
  tools/internal/cli/seat/help/*.md         [MODIFY] incl. edit.md, mint.md (estoppel's exemption)
  tools/internal/terms/terms.json           [MODIFY]
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
deletion of `ABOUT_KIND_CONTRADICTION` and `BlockerContradiction` leaves `migrate/replay.go:395-406` and
`planids.go:91-103` uncompiled until phase 2, and every package whose TESTS import `record/migrate`
fails to build its test binary until then. Sweep (`grep -rl 'internal/record/migrate"'` over the tools
module, at `0c57758e8`, the same importers as before #1663 but for `migrate`'s own tests): `internal/record` itself (`archivemigrate_test.go`, `planidphrases_test.go`,
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
| 2. migrate | §III.9 translation (`valueRenames` row, the raise by `anchor.Kind(ref)==""`, `TranslateReadingBlockers`, the reading act re-driven whole on replay with each member's own duties and `Batch.AppendBodies`, the deletions in `replay.go` and `planids.go` that phase 1's enum deletions leave uncompiled), `remap.go` `idFields`, `gapplace.go`'s sentence_of check, and the m30/m32/m23/sqlite-schema-source tests | §V.6; `go test -count=1 ./internal/record/...` (the record package's tests, phase 1's included, and migrate's); `go build ./internal/migratecli/ ./cmd/feov-migrate/ ./devcmd/archdump/ ./devcmd/archiverun/` (not `./devcmd/...` or `./cmd/...`: `devcmd/seatprobe` and `cmd/feov-record` import `internal/cli`, which builds in phase 3) |
| 3. Reading act | §III.2 flags (`gapFlags`, `ReadingClassField`, `RequiredUnless`, `ReasonIs`), the field table, the completeness stamp and the footnote gate, the reading's placed append, held gap inside the transaction, repeat, result words; §III.2a's CLI half (`PlacesToSentenceOf`, `verifiedFix`, the accept's exact span); §III.8 `seatprobe`; cli tests, the GAP 3 record-at-this-epoch test | `go test -count=1 ./internal/cli/... ./internal/migratecli/ ./internal/report/ ./internal/seatprobe/ ./devcmd/seatprobe/ ./cmd/...` |
| 4. Agent text | §III.7 help, terms, agentgen sources, debate.js; regenerate; read every changed golden line | §V.1 second run no diff; §V.4 census and phrase sweep |
| 5. Fuzz | §III.11 drives, gates, the four carried tally fields | `go test -count=1 ./releasegate/...` with the gate unset (compiles; the sweep runs in CI) |
| 6. Verify | §V.2, §V.3, §V.5, §V.8, §V.9; open the PR; read PR CI's fuzz tally (§V.7) | all of §V |

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation (step) |
|---|---|---|---|
| R1 | A carrier keeps speaking the contradiction (prompt, help, golden, fuzz) — half-state that reads done | H × H × L | §III.12 census re-run as §V.4, identifiers grep must be empty; agentgen/enginegen/vocabdoc `-check` gates (§V.2) |
| R2 | Lenses stop recording negative readings because the act now costs a full mint's flags | M × H × M | Flags are the same set `mint` already asks; refusal names each; smoke §V.8 checks negative readings still occur; the reading's gap takes no budget (F6) so cap is not a reason to read `supports` |
| R3 | One-gap rule + lens-picked class lets two lenses split one defect across two classes on one sentence | M × L × L | Ordinary behaviour for any gap; the admitted sets are two classes each; sweep tally reports per-sentence gaps by class |
| R4 | A bench `not_a_defect` ruling on a misread source closes the gap, and the run still cannot PASS while the reading stands | M × M × M | Ruled (F8): the ruling closes the gap through the bench arm, and the failed check alone holds the PASS, as an unread or unsound check does today; a failed check is not a dispute. It clears when the reader's own later reading backs the sentence (its reading replaced), or blue edits or retires — another lens's backing reading clears nothing. Stated in the bench's constitution (§III.7) and pinned by `TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld` |
| R5 | migrate mistranslates m30/m32 raises | L × H × L | Tests on the real archives (§III.9); refusal on corroboration raises is loud |
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
| R17 | A lens uses a reading to reopen text it prescribed and blue applied, which `mint` would refuse as an estoppel | L × M × L | Ruled exempt (F10): the reading must read the work and find the source against the sentence, and the reading itself is on the record as the evidence (its source and outcome in `show evidence`, its `--reason` as the gap's `mint_reason` on the board, in the report's "minted because" and in the closure index), where blue and the bench weigh it; `TestAReadingOverAnAppliedPrescriptionMintsItsGap`; §V.5.cc; the fuzz's `readingsOverPrescriptions` |
| R18 | A field the reading's Mint leaves blank reads as absent where a plain mint's reads: the lens's argument gone from the board and the report, or a `fix_basis` the board calls never omitted | M × M × L | Ruled (F11); §III.2's field table; one fill (`gapFlags.fill`); `TestAReadingsGapIsFilledAsAPlainMints` walks `Mint`'s descriptor, so a new field needs a disposition; `TestAReadingsGapCarriesItsArgumentWhereverAMintsDoes`; §V.5.ee, ff; the fuzz's stored-Mint check |
| R19 | A tool that maps an event to the act a seat typed reads the reading's Mint as a `mint` | L × L × L | §III.8's sweep; `TestAReadingIsAttributedToTheVerbTheSeatTyped`; §V.5.gg |

## V. Verification Plan

All commands from the worktree root, after `unset $(env | grep -o '^FEOV_[A-Z_]*')`.

1. `go -C scripts run ./protogen && go -C scripts run ./schemagen && go -C scripts run ./agentgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen && go -C scripts run ./vocabdoc` → no diff on a second run. Re-armed by: `record.proto`, `requirements.json`, `scripts/agentgen/src/**`, `terms.json`, the roster/enums.
2. `go -C scripts run ./check` → `0 failed` (read the count, not the exit code). Re-armed by: any file in the diff. Includes the `-check` gates of every generator, `archaeology`, `golden`, `pluginparity`, vocabulary prose.
3. `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/... ./internal/cli/... ./integration/... ./releasegate/...` (release gate unset: the fuzz sweep does not run) → 0 FAIL. Goldens: `UPDATE_GOLDENS=1 go test -count=1 <pkg>` then read every changed line. Re-armed by: any Go or help change.
4. Census re-run (must print only KEEP lines of §III.12):
   `grep -rIn -i contradict --exclude-dir=.git --exclude-dir=run-archive plugins scripts docs | grep -v -E 'cross_section_contradiction|claim_contradicts_(own_record|source)'`
   and identifiers, which must print nothing outside `migrate/` archived-word tables:
   `grep -rIn -E 'ABOUT_KIND_CONTRADICTION|BlockerContradiction|unanswered_contradictions|contradictionsOf|ContradictionOfReading|requireOpenContradiction|answersRaisedContradiction|answersOwnRaise|gap_may_name|raised_by' plugins scripts`;
   the phrase sweep of §III.7, over the whole tools tree — `tools/internal` (not only `cli/`) and `tools/releasegate` — and the agent-text paths (must print only its KEEP lines); the "does not close" sweep of §III.7 (every hit in the failed-check sense says the close is refused, never that the gap does not close); the failed-check sweep of §III.3 / §III.12 over `tools/internal`, `tools/cmd`, `tools/releasegate` and `debate.js`, every `answeredBy`/`AnsweredBy()`/`.Detail`/`BlockerFailedCheck` hit read for a branch that decides owed-ness, wording or a tally without the state (must find none); and `grep -rn 'carried_by\|carrier_minted'` over the tools tree, every hit one of §III.12's "Readers of the new Verify fields"; `grep -rn 'sentence_of\|SentenceOf\|sentenceOfMut\|ErrAnchorGone\|ErrAnchorBare\|AppendBodies'` over the tools tree, every hit a row of §III.2a's table (which names `migrate/replay.go` and `record/batch.go`) or a test; the proposal-span sweep of §III.2a (`ProposalOver`, `PlanAccept`, `proposalAt`, `ProposalAppliedVerbatim`, `PlanSplice`, `LocateUnique`), every locate of a gap's prescribed span by the gap's token and every typed edit's by its quote; the completeness sweep of §III.2b (every Verify writer stamps `source_completeness`, every reader's row as listed); the estoppel sweep of §III.2 (every hit that teaches or enforces the prescription estoppel states the reading's exemption, or is a KEEP row there); the replay-duty sweep of §III.9 item 6 (every per-source-event step of `replay.go`'s loop runs per member of a reading act, and no member's `newID` entry is read from `MAX(id)`); the field table of §III.2 (every `p.<Field> =` in `gapFlags.fill` and `newMint`, each a row, and every reader of an unset field a row); `grep -rn 'gapFlags\|ReadingClassField'` over the tools tree, every hit in `mint`, `verify`, `corroborate`, `record/enums.go` or a test; the `lane` table of §III.12 (`grep -rn 'Lane\.Role\|LaneVerbs\|blue-lane'`, each hit a row or other sense); the attribution sweep of §III.8 (`grep -rn 'verbOfEvent\|EVENT_TYPE_MINT\b' tools/internal tools/devcmd`, no tests or goldens); and the GAP-1 class sweep of §III.2b — `grep -rn -E 'SOURCE_OUTCOME_(REFUTES|ABSENT)|contradicts\(|VerifyReads|FindsAgainst|ReadingGapClasses'` over `tools/internal` and `tools/releasegate` — every hit that decides whether a reading mints, words its gap or counts it calls `ReadingGapClasses` (or `FindsAgainst` for the ledger), never the outcome alone.
   Re-armed by: any agent-facing or Go edit.
5. Delete-the-row checks (each must turn the suite red; revert after):
   a. delete `(gap_classes)` from `SOURCE_OUTCOME_REFUTES` → `TestARefutingReadingMintsItsGapInTheSameAct` and `TestEveryNegativeOutcomeNamesItsClasses` fail;
   b. delete the Mint body from the reading's append (keep the Verify) → the same test fails;
   c. delete the held-gap branch in the reading path → `TestASecondRefutingReadingNamesTheStandingGap` fails;
   d. delete the found-against branch of `failedChecksOf` → `TestTheLensCloseOfAReadingsGapIsRefusedWhileTheCitationStandsFoundAgainst` and `TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld` fail;
   e. delete the act-paired exemption in `requireMintWithinBudget` → `TestAReadingsGapTakesNoMint` fails (the reading's Mint is refused at the cap);
   f. delete migrate's raise translation → `TestTheMigratedM30RaiseIsAFindingAboutItsCitation` fails (refusal);
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
   v. delete the `"contradiction": "anchor"` row from `valueRenames` → `TestTheMigratedM30RaiseIsAFindingAboutItsCitation` and `TestEveryArchivedRunMigrates` fail (decode refusal); replace the raise's `anchor.Kind(ref) == ""` test with "about_kind is anchor" alone → `TestAFindingAboutAnAnchorIsNoRaise` fails (its finding about `C-…`, and its archived `proof`-kind finding about `P-…`, respelled `K-…`, refused as unparsable events.ids).
   w. restore the completeness stamp's guard to `body.GetUrl() != ""` → `TestAReadingOfACopyThatIsNotTheWorkMintsNothing` rows (a), (b), (g), (i) fail (a `verify` carries no stamp: its `refutes` demands gap flags, its backing reading clears the citation); drop `record.VerifyReads(body)` from the footnote decision → row (h) fails (a footnote minted from a `not_the_work` copy);
   y. replay the reading's three events one at a time (no re-driven act) → `TestARecordWrittenAtThisEpochMigratesWithNoRefusal` fails (the capped lens's Mint refused for budget, its Verify refused outside its act); write the act at the Mint's iteration and skip the other two (round-6 shape) → it fails (the plan refused: its `why` names the Verify's old id, which wrote no `newID` entry; the Stale through the Mint translated to the Verify's id); map each member from `MAX(id)` after the act → it fails (the Mint's Stale names the Verify); set `record.Now` once for the act → it fails (`ts` differ from the source's); count the act once in `res.Out` → it fails (`res.In`/`res.Out`); run `place.step` on the Anchor after the write → `TestMigrateChecksAReadingsGapBySentenceOf` fails (the act lands first: the migrated record holds G's Mint, Anchor and Verify where it asserts none); skip `gapplace`'s sentence_of check → `TestMigrateChecksAReadingsGapBySentenceOf` fails (admitted);
   z. locate `ProposalOver`'s span by text (`LocateOnce`) again → `TestAVerifyOfARepeatedSentencePlacesItsGapByTheAnchor` row (e) fails (the verified fix refused as ambiguous); drop the accept's exact-span locate → row (e) fails (the accept refused), and row (i) fails if the edit is recorded without `exact_span` (replay ambiguous);
   aa. read open-ness for the tie-break from Close events instead of the gap view → `TestASupersededReadingGapHandsTheCheckToItsSuccessor`'s bench-ruled branch fails (`AnsweredBy()` names G3, closed by the ruling);
   cc. attach `requireNotEstopped` to the reading's append → `TestAReadingOverAnAppliedPrescriptionMintsItsGap` rows (b) and (c) fail (refused as an estoppel, an estoppel entry logged);
   dd. register the reading verbs' `--new`, `--class` and `--problem` with `mint`'s descriptions → `TestTheReadingVerbsDescribeTheirGapFlagsByTheirSubject` and `TestEveryFlagAHelpPageNamesIsOnThatPage` fail (`verify`'s page names `--quote`);
   ee. drop `MintReason` from the reading's Mint → `TestAReadingsGapIsFilledAsAPlainMints` fails (presence differs), and `TestAReadingsGapCarriesItsArgumentWhereverAMintsDoes` fails on `show board --json` and on `report.Assemble`'s gap block and closure index;
   ff. leave `fix_basis` unset on the reading's Mint → both fail (`fix_basis` empty on the board and the work item, no `fix_basis` line in the gap block);
   gg. delete `seatprobe.Read`'s skip of a reading's Mint → `TestAReadingIsAttributedToTheVerbTheSeatTyped` fails (`mint` used by a seat that never typed it); skip every Mint a Verify names, `carrier_minted` unread → its plain-mint row fails; delete the anchor test on a Verify's verb → its `corroborate` row fails;
   hh. register `--class` on the reading verbs from `MustEnum("mint", "class")` → `TestTheReadingVerbsDescribeTheirGapFlagsByTheirSubject` fails (every class listed); drop `RequiredUnless` from `--class` → `TestHelpSaysRequiredExactlyWhereOmissionIsRefused` fails ("REFUSED without --class and its help does not say so"); restore a `supports` clean set in the verify probe → the same test fails ("says REQUIRED and the verb ran clean without it");
   ii. print the found-against sentence's event id after a word other than `event` → `TestEveryIDTheVerbPrintsIsTranslated` fails;
   bb. let the successor of a reading's gap inherit the budget exemption (exclude by `answers` instead of `carrier_minted`) → `TestAPlainMintALaterReadingNamesKeepsItsMint`'s successor row fails (the over-cap successor admitted).
   Re-armed by: the function or table each names.
6. migrate on real archives: `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/migrate/ -run 'M30|M32|M23|EveryArchivedRun|M19|AboutAnAnchor|ValueRenames|StringField|BySentenceOf'` and `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/ ./internal/cli/ -run 'M32|M30|MigratedFoundAgainst|M23s|RecordWrittenAtThisEpoch|EveryIDTheVerb'` → pass, 0 refusals. Re-armed by: `migrate/**`, `record.proto`.
7. Release fuzz: PR CI only (sharded 4 + tally). Read the tally: new gates non-zero; max gaps per (class, sentence) = 1. NEVER run `FEOV_RELEASE_GATE=1` locally. Re-armed by: any record or flag change (CLAUDE.md requires a full pass before done).
8. **Universe smoke (real engine, cheap configuration).** `scripts/universe.sh build --dir ~/.claude/scratch/contradiction-is-a-gap/universe`; prove the build carries the change: `grep -rl claim_contradicts_source ~/.claude/scratch/contradiction-is-a-gap/universe/config/plugins/cache/special-circumstances/frank-exchange-of-views/` → at least the installed agents and the `feov-record` binary; and `scripts/universe.sh doctor` → schema epoch 53 (or the rebased number). Then
   `MODEL=haiku JUDGMENT_MODEL=haiku LANES=1 scripts/universe.sh --dir ~/.claude/scratch/contradiction-is-a-gap/universe run "Is the Great Wall of China visible from low Earth orbit with the naked eye?" --mint-budget 3 --k-max 2`
   (a topic whose popular sources misstate the claim, so an evidence lens is likely to read a refuting source). Observe with `universe.sh watch`; afterwards, on the run's record (read-only `file:` URI):
   `SELECT v.outcome, v.anchor, v.carried_by, m.class, m.class_material, m.mint_reason, m.fix_basis FROM verify v LEFT JOIN mint m ON m.gap_id = v.carried_by WHERE v.outcome IN ('refutes','absent') AND COALESCE(v.source_completeness, '') <> 'not_the_work' AND NOT (v.anchor IS NULL AND v.outcome = 'absent');`
   (the readings that find against: they read the work, and a corroboration's `absent` mints nothing, F5).
   Expected: every row has a `carried_by`, class in the outcome's admitted set, `class_material = always`, a non-blank `mint_reason` and a `fix_basis` of `proposed` or `verified`; any second reading of one sentence shares a `carried_by`. The converse —
   `SELECT COUNT(*) FROM verify WHERE carried_by IS NOT NULL AND (outcome NOT IN ('refutes','absent') OR COALESCE(source_completeness, '') = 'not_the_work' OR (anchor IS NULL AND outcome = 'absent'));`
   — is 0. Every row is of an admitted act: refused acts write nothing. If no negative reading occurred, the smoke is reported **NOT MEASURED on this path** (not green), and check 9 carries it. Archive the run to `run-archive/`.
9. **Driveable check on real data.** m30 as archived holds its judge's `outcome` (events.id 467) and the chair's verdict (463), and `requireOpenToSeats` refuses every seat act after an outcome, so the archive is CUT before it is driven: copy m30's `records/record.db` to `~/.claude/scratch/contradiction-is-a-gap/m30/`, and in the copy delete every row with `event_id >= 463` from each table holding an `event_id` column, then `DELETE FROM events WHERE id >= 463` (a scratch step; nothing in the repository changes). Migrate the cut copy (`TestEveryArchivedRunMigrates`' path, or `feov-record migrate`), and confirm a seat act is admitted on it (`register` as `red-lens-evidence`, `show board`) before the drive — a refusal there means the cut is wrong, not the change. Then on the migrated copy as `red-lens-evidence`: `feov-record verify --anchor C-4f4b8cfa --as refutes --class claim_contradicts_source …` → a gap lands (or names G-dc877265 if it is open with that class); a second lens's identical reading → names the same gap, no mint; `close` of that gap → refused while the reading stands; a third lens's `--as supports` reading → recorded, and the close still refused (a disagreement clears nothing); the reader's own `--as supports` re-read a sitting later → the close admitted; `show evidence` has no `unanswered_contradictions` key, and its verification row for the reading
   carries `source_completeness` `not_asked` (the archive holds no fetched-source cache, so the stamp
   says the copy was never asked — the stamp `verify` now writes, read off a real record); `show board --json`
   lists the gap with the `--reason` as its `mint_reason` and its `fix_basis`. Re-armed by: the reading act, the ledger, the close gate.
10. Auditor gate: `/prosthetic-conscience:plan-audit plans/contradiction-is-a-gap.md` — two consecutive PASS, or the human's approval with every fork ruled.

## Forks — all RULED (gblock, 2026-10-30; F10, F11 2026-10-31)

Each ruling is recorded here; the alternatives it rejected are deleted from the design.

- **F1 — Class (RULED).** A NEW class `claim_contradicts_source` for `refutes`; `citation_figure_misattribution`
  narrowed to `absent`. The lens picks among the classes its outcome allows, declared once on
  `SourceOutcome` (`gap_classes`): `refutes` → `claim_contradicts_source` | `within_source_condition_misattribution`;
  `absent` → `citation_figure_misattribution` | `within_source_condition_misattribution` (§III.1). Cost
  carried as R3.
- **F2 — Judgment (RULED).** The reading act (`verify`/`corroborate`) takes the mint's judgment flags,
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
  searching for the sentence's text. Placement from text alone (`corroborate --quote`) that is
  ambiguous stays refused. No anchor flag is added to `corroborate` (§III.2a).
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
  archive, and the m23 test asserts the migrated state (§III.9).
- **F10 — Estoppel and a reading's gap (RULED, gblock 2026-10-31: "Exempt, stated").** A reading's gap
  is exempt from estoppel: a source's refutation is new evidence, not red relitigating its own
  prescription. The gap opens, and the failed check holds the PASS (a `corroborate`'s gap, material by
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
- **F5 — Corroboration `absent` (RULED).** Records the reading and mints no gap (§III.1, §III.10).
- **F6 — Budget (RULED).** A gap a reading mints is outside the mint budget (§III.5).
- **F7 — PR #1642 (RULED).** Closed as superseded; its four diagnostic tally fields — `final_blockers`,
  `final_parties`, `open_by_found_by` (re-keyed `reading | finding | none`), `mints_by_epoch` /
  `closes_by_epoch` — are carried into this PR (§III.11).
- **§III.9 — Archived carrier-less checks (RULED).** Cleared by the reader's own backing re-read, an edit or a retire. migrate does
  NOT fill `carried_by` from the `found_by` chain; the residue is stated in §III.9.
- **§III.14 — Delivery (RULED).** ONE PR, implemented one phase per fresh agent.
