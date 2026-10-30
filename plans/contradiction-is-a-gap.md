# A source reading against the report is a gap — the contradiction is deleted

Written against `origin/main` **`8236e3a17`** (EventSchema 48, after Slim-fast's #1601 and #1643).
Every file:line below is read at that commit. Re-checked against `origin/main` `eb73f6dc6` (round 4): the
commits since touch only `internal/tessocr/` and OCR corpus goldens, so every file:line holds. Implements
gblock's ruling of 2026-10-30; every fork is RULED (gblock, 2026-10-30, §Forks).

> **Owner ruling (gblock, 2026-10-30, his words):** "why isn't the contradiction recorded as a gap in
> the first place like red would for any other problem it finds? We're letting it mint a weird
> placeholder." And: "Our standard behaviour is reject a duplicate, not dedup — why are you treating
> contradictions as unique?" He approved planning the redesign: a `refutes` or `absent` source reading
> is recorded as a gap of a real class, like any other problem red finds, and the contradiction
> concept is deleted.

## I. Summary & Goals

**Objective.** Today a lens's `verify`/`corroborate` reading whose outcome is `refutes` or `absent`
opens a **contradiction** (`record/citationid.go:484` `contradictionsOf`): a record object with its own
id (the reading's `events.id`) that a finding must *raise* (`--about-kind contradiction --about <id>`)
and that only the gap minted from that finding (`found_by`) *answers*. Until then it is a PASS blocker
of its own kind (`passblockers.go:35`), a work-list debt of the reader (`sitting.go:369-425`), a
readiness reason (`dispatch.go:520-526`), and it is exempt from the one-gap-per-class rule
(`onegapperclass.go:68`) and from the mint budget (`mintbudget.go:299-303`). A reading after the raise
opens a *new* contradiction (`citationid.go:487-491`), so N readings of one sentence yield N
contradictions, each owed its own exempt gap: in release-fuzz run 40872228665 (PR #1642's finding),
22 of 45 mints per run answered a contradiction and all 37 drawn runs ended CEILING. m32 shows the
same shape on real seats: four lenses refuted one claim on `C-a36dce3f`, four findings raised four
contradictions, one gap answered one of them.

After this change a negative reading **is** the gap: the reading act mints a gap of a real class on the
sentence, through the ordinary mint path (class, quote, grades, check, one-gap rule). A second negative
reading of a sentence that already holds an open gap of that class is the ordinary duplicate: its mint
is refused by the one-gap rule, the reading is kept as evidence and names the standing gap. Blue answers
it like any gap. The contradiction — its about-kind, raise, fold, blocker kind, `unanswered_contradictions`,
work items, readiness reasons, exemptions, term, help, constitution and prompt text, goldens and fuzz
drives — is deleted, with no alias and no "formerly" prose.

**Success criteria (each a §V check).**

1. A `verify` or `corroborate` with `--as refutes` (or `verify --as absent`) writes the reading AND an
   open gap in one transaction; the reading's JSON result names the gap (`gap.id`, `gap.minted: true`).
2. A second negative reading of the same sentence with the same class, while that gap is open, writes the
   reading, writes no gap, and names the standing gap (`gap.minted: false`). Gaps per (class, sentence)
   open at once: **≤ 1**, asserted over every fuzz run by the sweep tally.
3. The identifiers grep of §V.4 returns **0** lines outside migrate's archived-word tables; the word grep
   returns only lines §III.12 lists as KEEP (ordinary English, the two self-contradiction classes,
   historical plans).
4. `migrate` brings every archive forward with **0** refusals (`TestEveryArchivedRunMigrates`), and the
   migrated m30 and m32 hold the dispositions §III.9 states.
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
  `scripts/schemagen` from `plugins/frank-exchange-of-views/requirements.json` (`"eventSchema": 48`).
- **Epoch.** This plan takes **EventSchema 49**, confirmed with the Slim-fast peer session (relayed by
  the coordinator, 2026-10-30). Slim-fast takes 50, or the next free above ours, for its epoch PR
  (`MotionRule.relief`; ruling rows read from the record; the bench grades by an ordinary regrade; a docket
  motion on a ruled gap needs `--grounds` as a locator — `--quote` or an F-/C-/P-/G- id). PR #1644
  (`feov/compute-1a`, #1544 Part 1a) will also claim one in its logic phase. Whichever PR lands second
  rebases and re-derives its number from `origin/main`; the number is never carried from this document.
- **Slim-fast's `--grounds` and this plan's gaps.** A reading's gap is an ordinary Mint with an ordinary
  `G-` id, so it qualifies as `--grounds`. migrate creates no gap for an archived refutation: an
  archived answering gap (m30 `G-dc877265`, m32 `G-f417112d`) is carried with its own `G-` id, and an
  archived reading no gap answered gets none — the tool does not invent a gap the seat never minted
  (its grades would be fabricated).
- **Record shapes touched** (contract): `AboutKind` loses `ABOUT_KIND_CONTRADICTION = 4` (number and name
  `reserved`) and the `gap_may_name` facet (every remaining kind admits a gap, so the facet is dead);
  `GapClass` gains one value (F1); `SourceOutcome` gains a facet naming the classes each negative
  outcome admits (F1); `Verify` gains `carried_by` (references `mint.gap_id`) and `carrier_minted`; `Mint` is unchanged.
- **Existing mechanisms reused, not invented:** the one-gap rule (`onegapperclass.go`, eb0c5a99f), the
  failed-check state and its close gate (`failedcheck.go`, `requireFailedCheckAnswered`), the class's
  `material_default` stamped at write (`Mint.class_material`, `recordpb/gapclass.go:18`), the
  placed-append that already writes two events in one act (`seat.AppendPlaced`, used by corroborate's
  footnote), and replay-time reference translation in migrate (`migrate/replay.go:351-362`).
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
  - "#1522 made a refutes|absent reading a contradiction": the contradiction dates from gblock's 2026-10-06 rule (`citationid.go:455`); #1522 (a9ab33f49) made the finding insufficient, so only the gap minted from it answers.
  - "m19 contains contradictions": m19 (epoch 24) holds one negative reading (ev 114, `C-706fda63`), but the `contradiction` about-kind did not exist before epoch 30/31, so nothing in it was ever raised. Only m30 and m32 hold raises.
  - "N sources refuting one sentence yield N gaps": that holds in the release fuzz. No archive shows it. m32 shows N raises with one gap: four readings, four raising findings, one answering gap, three raised and never answered.
- **Out of scope.** `unreachable` and not-the-work readings keep the failed check's *owed-gap* shape
  (the lens owes a gap about the anchor): nothing was read, so there is no claim to mint against at
  the act. Proofs judged `unsound` are unchanged. The `cross_section_contradiction` and
  `claim_contradicts_own_record` classes are different concepts and stay.

## III. Proposed Changes

### III.0 The design, decided

| Question (task item) | Decision | Fork |
|---|---|---|
| 1. One class or two | A new `claim_contradicts_source` for `refutes`; the existing `citation_figure_misattribution` narrowed to `absent`; the lens picks among the classes its outcome allows | **F1** RULED |
| 1b. Who names class, quote, problem | Quote = the reading's claim (the tool already reads it for `verify`; `corroborate`'s `--quote`). Class = the lens's `--class`, refused outside the outcome's admitted set. Problem, check, check-kind, severity, likelihood, impact, fix = the lens's flags on the reading act, required on a negative outcome and refused on any other | **F2** RULED |
| 1c. Materiality | The class's `material_default` (both `always`), stamped at the write as for every gap | — |
| 2. Duplicate | The reading is kept and the one-gap rule refuses the *mint*; `Verify.carried_by` names the standing gap | **F3** RULED |
| 3. Backing re-check | Found-against is a third **failed-check** state beside unread/unsound. It holds the PASS until its reader's own later reading backs the sentence, blue edits the sentence, or blue retires the citation; another lens's backing reading clears nothing. The lens's close of its gap is refused while it stands (`requireFailedCheckAnswered`); a closing docket ruling closes the gap and the check still holds the PASS. Once it clears, the minting lens closes the gap with its ordinary close. No new answer kind | **F4**, **F8** RULED |
| Corroboration `absent` | Records the reading; mints no gap | **F5** RULED |
| Mint budget | A gap a reading mints is outside the mint budget: it takes no mint and does not count while open; the "answers its own raise" exemption is deleted | **F6** RULED |
| 4. Delete | Everything in the census (§III.12) | — |
| 5. migrate | One translation, §III.9 | — |
| 6. Epoch | 49, re-derived at rebase | — |
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
  distinguisher stops saying "says something else". Name grep (CLAUDE.md naming rule) run at `8236e3a17`:
  `contradicts_source` has 0 hits; the sibling `claim_contradicts_own_record` is the parallel the name
  follows (report vs record → report vs source).
- `SourceOutcome`: [NEW] facet `(gap_classes)` on `REFUTES` (`claim_contradicts_source`,
  `within_source_condition_misattribution`) and `ABSENT` (`citation_figure_misattribution`,
  `within_source_condition_misattribution`); absent on every other value. This is the one declaration of
  "a negative reading mints"; `contradicts()` (`citationid.go:666`) is replaced by "has `gap_classes`", so
  a value added to the enum cannot fall between the reading rule and the anchor ledger. The `REFUTES`/
  `ABSENT` `(means)` lose "the strongest finding this verb can carry" in favour of "mints a gap".
  `within_source_condition_misattribution` is admitted because both archived answering gaps (m30
  `G-dc877265`, m32 `G-f417112d`) were minted in it by the lens that read the source.
- `Verify`: [NEW] `optional string carried_by = 20 [(sql) = { references: "mint.gap_id" }]` and
  `optional bool carrier_minted = 21` (field numbers 1-19 are taken; 18 is `copy_substitution`).
  `carried_by` is the gap that carries this reading against its sentence: the one the reading minted, or
  the open gap of the reading's class on that sentence that refused the mint. `carrier_minted` is true
  when this reading's own act minted it, false when it names a standing gap. Both are set by the tool,
  on a negative reading only (`seat.Supplies`, as `mint.go:288` declares the Mint's `gap_id`), so no
  flag sets them.
  - **Why not `gap_id`.** The idempotency key is matched by field NAME (`record.go:483` `keyFields`,
    `"gap_id"` first, ahead of `anchor` and `url`): a Verify carrying a field named `gap_id` would key
    on the gap instead of its citation or source, and `flags/names.go:393` would print it as `--id`, a
    flag the reading verbs do not have. Name grep run at `8236e3a17`: `carried_by` and `carrier_minted`
    have 0 hits. The names' other readers are swept in §III.12 ("Readers of the new Verify fields").
  - The write-path rules, in `record.go`'s Verify case, not table CHECKs (archived negative readings
    carry neither field, §III.9.2, and replay (`Migrating`) admits them as every other new write rule
    does): a negative reading names its gap (except a corroboration's `absent`, F5); `carrier_minted`
    true is admitted only inside the two-body append whose Mint has that `gap_id`; `carried_by` with
    `carrier_minted` false must name an open gap of a class in the outcome's `gap_classes` that quotes
    the reading's sentence. A field the writer can refuse, not a relation recovered from sentence
    position or event adjacency (facts-are-fields).
- `Mint`: comments at `record.proto:888-891` [MODIFY] (no kind is refused). No field change.
- Regenerate: `go -C scripts run ./protogen`, `go -C scripts run ./schemagen`; `requirements.json`
  `eventSchema` 48 → 49.

### III.2 The reading act — `internal/cli/lens/verify.go` [MODIFY] (contract: CLI flag set; prompt a seat is told)

- `verify` and `corroborate` gain the mint's judgment flags, the same `flags.*` names and descriptions
  `lens mint` registers (`mint.go:395-405`): `--class`, `--problem`, `--fix`, `--new`, `--check`,
  `--check-kind`, `--severity`, `--likelihood`, `--impact`, `--complexity`. Required with an outcome
  that has `gap_classes`, refused with any other ("a reading that backs the sentence mints nothing").
  `--class` is refused outside the outcome's `gap_classes`, and the refusal lists them. One flag
  registration helper shared by `mint` and the reading verbs, so the two cannot drift.
- `--reason` stays the reading's (`Verify.text`: what the source says). The gap's `mint_reason` is not
  written from it: `Mint.mint_reason` is red's argument, and `problem` is "never --reason"
  (`record.proto` Mint.problem `why`).
- The write: one transaction holding the Mint (placed at the claim, as `lens mint` places its quote)
  and then the Verify carrying `carried_by` — Mint first, because `carried_by` references `mint.gap_id` and the
  store runs with `foreign_keys(1)`. `record.go` gains the two-body placed append beside `AppendPlaced`;
  the Mint passes every mint validation (`record.go:955-980`: unclassified reason, subject, one-gap,
  budget) in the same `writeView`, so the one-gap read and the insert share the write lock
  (`_txlock=immediate`, as `onegapperclass.go:32-35` already relies on).
- **The budget exemption comes from the act, not the record (F6).** The Mint's budget check runs before
  the Verify exists in the transaction, so it cannot read a Verify naming the gap. The two-body append
  hands the Mint's validation the act's paired Verify body; `requireMintWithinBudget` admits a Mint so
  paired without counting it (§III.5). A plain `mint` has no paired body and is counted as today.
- One-gap refusal **inside the reading act** (F3): `requireOneGapPerClassAndSubject` returns the held
  gap as a value rather than only an error; the reading path writes the Verify with `carried_by` = the held
  gap and no Mint. The plain `mint` verb keeps its refusal unchanged.
- Repeat: `requireNotARepeatFailedCheck` (`failedcheck.go`) extends to a negative reading by the same
  lens of the same sentence from the same source (the anchor for `verify`; the url and claim for
  `corroborate`), sentence unchanged, while the gap its earlier negative reading's `carried_by` names
  (or that gap's open successor, §III.3) is open — refused, nothing recorded, naming the earlier reading's
  event and the gap: "your reading of citation C at event N finds the source against this
  sentence, the sentence has not changed, and gap G carries that reading and is open — nothing is
  recorded; the reading and its gap stand". A reading with a *different* outcome is not a repeat.
- Text that states the old model without the word [REWRITE]: the `--as` flag help (`verify.go:304`,
  "refutes and absent are findings") and `human()` (`verify.go:698-701`, "that is a finding, and this event
  is the evidence for it"); see §III.7's phrase sweep.
- Result: `verifyResult.Contradiction` / `readingContradiction` / `contradictionOf` / `words()`
  (`verify.go:615-688`) [DELETE]; [NEW] `Gap *readingGap {id, class, minted}` and its sentence:
  "This reading mints gap G (class C)" / "Gap G, open on this sentence with class C, carries this
  reading; no second gap is minted — add what it lacks as a finding about G". `gap.minted` is read back
  from the stored `carrier_minted`, never computed by the verb.
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
  (`verify.go:466-481`) runs only on a backing corroboration (the label it keys on is minted only
  there) and is unchanged. A negative corroboration `absent` (no gap, F5) retried in its sitting meets
  the duplicate-key refusal on its url, as today. Tested by
  `TestARetriedNegativeReadingIsRefusedNamingItsGap` (`cli/keyedretry_test.go`, §III.10).

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
  successor chain (`successorsOf`, `successor.go:29`, from `Close.successor`), and `answeredBy` is the
  chain's live end (the latest successor, `carried_by` itself when none). The close gate, the repeat
  rule and every printed sentence then name the gap that carries the check now. No field is added: the
  chain is the record's own `Close.successor`. Tested by
  `TestASupersededReadingGapHandsTheCheckToItsSuccessor` (§III.10).
- **What clears a found-against check is the ledger's rule** (`anchorStanding.stands`,
  `anchorchecks.go:84-94`): only the SAME seat's later reading that reads replaces its check. So a
  found-against check clears when its reader's own later reading of the sentence as it stands backs it,
  when blue edits the sentence (which leaves it unchecked), or when blue retires the citation. Another
  lens's backing reading beside it is a disagreement and clears nothing — as `failedClears` already says
  of a proof's `sound` beside a standing `unsound`. Every found-against sentence below that states a
  clear condition states this one; none says "a lens's reading" or "a lens reads its source" (the
  `unread` state's condition, which any reading of the work meets).
- **A closing docket ruling closes the carrying gap; the check alone keeps holding the PASS** (gblock,
  2026-10-30: "Gap closes, PASS still held"; the same as today's unread and unsound checks,
  `TestABenchRulingDoesNotClearAnUnreadCitation`). The bench's ruling closes a gap through
  `closure_act`'s bench arm (`recordsql/views.go:443-450`, read by the gap view at `:531-541`) and writes
  no Close, so `requireFailedCheckAnswered` (`record.go:1105`, Close only) does not run, and nothing is
  added to make it run. The failed check is not the gap: it stands until one of the three acts above,
  and no bench ruling clears it. No sentence says the gap does not close under a ruling.
- **OWED-NESS AND WORDING KEY ON THE STATE, NEVER ON `answeredBy == ""`.** `failedCheck` and `Blocker`
  gain the check's state (`unread | unsound | found_against`; `Blocker.CheckState()` exported for the CLI
  and the fuzz), and `Detail != ""` stops standing for "proof, else unread citation". The rule every
  reader below holds: a lens **owes** a gap for a failed check of its own only when the state is `unread`
  or `unsound` and no gap carries it; a `found_against` check is never owed by its reader, carried or not
  — its gap is minted at the act, and where none is named (a migrated archived reading, §III.9) no mint
  answers it either. `answeredBy` says only *which gap carries it*, for printing and for the close gate.
  The readers, each changed to that rule (census: §III.12, "Readers of a failed check's owed-ness"):
  - dispatch's lens arm (`dispatch.go:531-538`): readies the owner for `unread`/`unsound` with no carrier;
    never for `found_against`, which readies blue alone (`:540-546`, cause `max(since, answeredAt)`);
  - the lens's work item (`sitting.go:396-403`): `add` (owed) for `unread`/`unsound` with no carrier;
    `may` for every `found_against` check — so `EnvelopeRefusal` (`envelope.go:44`, which reads this
    list) never refuses a reader's envelope over one;
  - retirement (`retirement.go:164-172`, `lensOwesOf`): a failed check is owed only for
    `unread`/`unsound` with no carrier; `found_against` is skipped, so it never holds its reader from
    retiring. Its doc comment (`:142-144`) says "a check of its that read nothing or judged a proof
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
  - `failedFact` (`:220-225`): `found_against` → "L's reading of citation C at event N finds the source
    against the sentence (`refutes` or `absent`), and the report still carries that sentence";
  - `failedClears` (`:228-234`): `found_against` → "it clears when L's own later reading of the sentence
    as it stands backs it, or when blue edits the sentence, which leaves it unchecked, or retires the
    citation; another lens's backing reading beside it is a disagreement and clears nothing. The lens's
    close of the gap that carries it is refused while it stands; a closing ruling of the bench closes that
    gap, and the check still holds the PASS — no bench ruling clears it";
  - `failedRemedy` (`:239-244`): `found_against` → "edits the sentence, or retires or replaces the citation";
  - `failedOwed` (`:248-253`): the "owes its gap about the anchor (`mint --about-kind anchor …`), which
    takes no mint of its budget" branch is `unread`/`unsound` with no carrier only; `found_against` with a
    carrier → "gap G carries it to blue, which …"; without one → "no gap names it — blue …" (no seat is
    told it owes a mint);
  - `blueFailedItem` (`:256-268`): "the lens whose check failed has minted no gap about it yet" is the
    `unread`/`unsound` no-carrier sentence only; `found_against` without a carrier → "no gap names it";
    the citation remedy sentence takes `failedRemedy`'s found-against wording, and "Retire the citation" is
    the `unread` sentence's;
  - the kind's `gloss` (`:398`), `refusal` (`:402-420`: its hard-coded "The lens whose check failed mints
    its gap about the anchor …" names the `unread`/`unsound` duty only, and `how` gains "found against"
    (participle, as "judged unsound")) and `why` (`:421-427`, `until`);
  - `failedCheckReadiness` (`dispatch.go:1297-1312`): `what, failed` and blue's remedy by state; the lens
    arm's "minted no gap about it" stays, as only `unread`/`unsound` reach it.
  - the verb's failed-check sentence (`verify.go:529-541`, `:631-656`): by state, after every
    verification of a citation, §III.2 — this site was judged state-safe in round 2 and is not: its
    guard is the new reading's outcome, not the check's state.
- Read and judged state-safe: `verdict.go:141-162` (`anchorsStanding`) and `debate.js:1220-1227` name
  the anchor only, whatever its state.
- The close gate: `requireFailedCheckAnswered` refuses the minting lens's (or any seat's) `close` of a
  gap in a found-against check's `answers` — successor (`--superseded-by`) admitted — until the check
  clears by the ledger's rule above. A closing docket ruling closes the gap and the check still holds
  the PASS (above). The gap's close stays the minting lens's ordinary `close`, so no new answer kind
  exists.
- **The mechanism's own sweep** — every site that SELECTS a failed check (by Subject, Owner, AnsweredBy
  or Detail) and decides or prints, re-run at `8236e3a17` with
  `grep -rn -E 'BlockerFailedCheck|AnsweredBy\(\)|failedChecksOf\(|\.answeredBy|\.answers\b|\.Detail\b'`
  over `tools/internal`, `tools/cmd`, `tools/releasegate` and `debate.js`, every hit read (dispositions in
  §III.12's table "Readers of a failed check's owed-ness"): the record readers above; `verify.go:538`
  (§III.2); `sitting.go:321` (blue's arm, prints `blueFailedItem`, by state); `passblockers.go:685` (the
  producer, which now sets the state); the fuzz sites of §III.11 (`fuzz_test.go:1063, :1072, :2814, :2824,
  :2846, :2849, :2903-2912, :3210, :3503, :3995, :4040, :4072`, and the gates at `sweeptally_test.go:788-799`);
  KEEP — `verdict.go:148`, `debate.js:1226`, `fuzz_test.go:2760` (asserts the owner or blue is readied,
  true of every state), `fuzz_test.go:6700` (no VERIFIED over any failed check), `migrate/plankinds.go:19`.
  The contradiction fold's hits (`citationid.go:516-647`, `onegapperclass.go:145`, `mintbudget.go:360-378`)
  are deleted. `PlanBlocker` (`dispatch.go:34-38`) carries kind, subject and owner only, and stays so:
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
- `holds()` comment (`anchorchecks.go:104-108`) and `VerifyReads` comment (`:278-281`) [MODIFY].
- Corroborations place no anchor when negative, so they reach the board through their gap only.

### III.4 The one-gap rule — `onegapperclass.go` [MODIFY]

- `answersRaisedContradictionAt` and its call (`:68-72`) [DELETE]; the exemption bullet (`:40-42`)
  [DELETE]. The rule now has no exemption but `--supersedes` and replay.
- [NEW] the held-gap return used by the reading act (III.2).

### III.5 Mint budget — `mintbudget.go` [MODIFY] (F6)

- `answersOwnRaiseAt`, its call (`:299-303`), `ownAnswersAt`'s contradiction half (`:367-395`), the
  `Answering` wording at `:504`, `:560` [DELETE/REWRITE]; the comment block `:33-41` [REWRITE].
- [NEW] a gap the reading act minted takes no mint and does not count while open. Two readers, two
  sources, neither inferred from "some Verify names this gap":
  - **at the write** (`requireMintWithinBudget`, `mintbudget.go`): the Mint's validation runs before
    the Verify exists in the transaction, so the exemption is read from the ACT — the two-body append
    passes its paired Verify body to the Mint's validation (§III.2), and a Mint so paired is admitted
    without counting it;
  - **while open** (`lensMintStanding`): a gap is excluded from the seat's count when a Verify by that
    seat names it with `carrier_minted` true — the Mint written in the reading act. A gap the lens
    minted with plain `mint` and that a later reading names as the open gap refusing the mint
    (`carrier_minted` false) keeps its mint. Tested by `TestAReadingsGapTakesNoMint` and
    `TestAPlainMintALaterReadingNamesKeepsItsMint` (§III.10).
- `mintAnswerOutsideBudget` (`:345`) [REWRITE]: "A gap your reading mints, or about an anchor whose check
  of yours read nothing or judged a proof unsound, takes no mint; …".

### III.6 PASS blockers, work list, readiness, retirement [MODIFY/DELETE]

- `passblockers.go`: `BlockerContradiction` (`:35`), `contradictionReader`, `contradictionCitationAnswer`,
  `contradictionAnswerBudget`, `contradictionGapSubject`, `contradictionFact` (`:138-155`), the kind's
  table row (`:334-361`), the fold call (`:678-679`), `raisedBy`/`raiser` fields (`:80`) [DELETE].
  `failed_check`'s row (`:397`) gains the found-against sentence.
- `citationid.go:440-670`: `contradictionsOf`, `ReadingContradiction`, `ContradictionOfReading`,
  `requireOpenContradiction`, `unansweredContradictions`, `contradiction`, `contradictionActs`, `idWord`
  [DELETE]; `contradicts` → replaced by the `gap_classes` facet reader (III.1). Comment at `:74` [MODIFY].
- `about.go`: the contradiction case (`:47-50`, `:82-88`) and the doc sentence (`:15-16`) [DELETE].
- `record.go`: the finding case (`:1059-1063`) and the mint's about-kind refusal (`:966-971`) [DELETE] —
  with every kind admitted the check has nothing to refuse; `enums.go:294-300` (finding's set "less
  contradiction") [REWRITE]: finding and mint carry one set.
- `sitting.go:369-425` (the lens owes a contradiction it read; the at-cap special case) [DELETE]; the
  failed-check arm (`:396-403`) keys on the state (§III.3).
- `dispatch.go`: the readiness case (`:520-526`), `contradictionReadiness` (`:1285-1292`), Plan JSON
  `desc` text (`:18-36`) [DELETE/REWRITE].
- `retirement.go:20, :142` (a lens is not retired while it owes a contradiction) [REWRITE]; the
  failed-check skip (`:164-172`) keys on the state, so a found-against check never holds its reader
  from retiring (§III.3).
- `evidenceview.go:306-373`: `UnansweredContradictions`, `EvidenceContradictionJSON` [DELETE];
  `:564`, `:584` comments [REWRITE]. `narrowedview.go:20` [REWRITE].
- `convergence_refusal.go:77` [REWRITE].

### III.7 Agent-facing text [MODIFY / REGENERATE] (contract: prompts and constitutions)

- Help pages (`internal/cli/seat/help/`): `verify.md`, `corroborate.md` — the two contradiction paragraphs
  are replaced by one paragraph: a negative reading mints its gap in the same act, the flags it needs,
  what the one-gap rule does to a second reading, and what clears the citation's failed check (your own
  later reading of the sentence as it stands that backs it, blue's edit, or blue's retire; another lens's
  backing reading is a disagreement and clears nothing);
  `finding.md`, `mint.md`, `dispatch.md` [REWRITE] the contradiction sentences; `finding.go:144-145` flag
  descriptions [REWRITE]; `verbs.go`, `flags/names.go` per census. The failed-check paragraphs name the
  third state: `mint.md:15` ("A CHECK OF YOURS THAT FAILED … you owe its gap") says the owed gap is an
  unread or unsound check's, and that a reading that finds the source against its sentence mints its gap in the same act;
  `close.md:11` [REWRITE] opens "YOUR CLOSE OF A GAP THAT CARRIES A FAILED CHECK IS REFUSED WHILE THE
  CHECK STANDS", adds the found-against case, states what clears each state by the ledger's rule (a
  found-against check: your own later reading of the sentence as it stands that backs it, blue's edit of
  the sentence, or blue's retire of the citation — another lens's backing reading clears nothing), and
  says a closing ruling of the bench closes the gap while the check still holds the PASS.
- **No page says a gap carrying a failed check does not close** (gblock's ruling, F8). Sweep, run at
  `8236e3a17` over `scripts/agentgen/src`, `tools/internal`, `tools/releasegate`, `skills`, `docs`:
  `grep -rIn -i -E "(does not|cannot|can't|never) close"` plus the multi-line form
  `perl -0777 -ne 'while(/(gap|it)[^.]{0,60}?(does not|cannot|can.t|never) close[^.]{0,80}?(while|until|stand)/gis){…}'`,
  every hit read. Hits in this sense, each [REWRITE] to "the close is refused while the check stands; a
  closing ruling of the bench closes the gap, and the check still holds the PASS" in the page's own
  voice: `help/close.md:11`, `help/mint.md:15` ("it does not close while the check stands"),
  `help/verify.md:11` ("That gap does not close while the citation stands unread" — the rest of that
  unread paragraph is unchanged), `help/reproduce.md:21` ("That gap does not close while the judgement
  stands"), `terms.json:1042`, `fragments/lens-audit-and-mint.md:6` ("The gap does not close while the
  check it carries still stands failed"), `passblockers.go:232` (`failedClears`, unread) and `:417` (the
  refusal: "the gap does not close while the check stands, and no bench ruling clears it" → "a seat's
  close of the gap is refused while the check stands; a closing ruling of the bench closes that gap, and
  no bench ruling clears it" — the last clause is the substring `TestABenchRulingDoesNotClearAnUnreadCitation`
  asserts), `failedcheck.go:193, :198` (the close refusals' "so the gap does not close" → "so this close
  is refused"), `docs/citations-and-bibliography.md:155-159`; code and test comments in the same sense:
  `failedcheck.go:30`, `unreadcitation_test.go:119`, `unreadb6_test.go:21-22, :66`, `fuzz_test.go:2906,
  :3992`, `sweeptally_test.go:794`; REGENERATE `docs/vocabulary.md:605` and the help golden. Every other
  hit is another sense (a sitting the record cannot close, a quote or fence that does not close, the
  `remanded` disposition) and KEEPs.
- `terms.json`: the `contradiction` entry [DELETE] (with its ordinary-English collision); `failed check`
  [REWRITE]: "A failed check is a standing check of a citation or proof that fails — a lens's standing
  judgement calls the proof's sentence unsound; no reading of the citation since its sentence last
  changed reads the cited work; or a lens's standing reading finds that the cited source refutes the
  sentence or does not contain it, and that reading mints the gap that carries it in the same act. It
  holds the PASS and is cleared by no
  bench ruling. An unread or unsound check obliges the seat whose check failed to mint a gap about the
  anchor, outside its mint budget. A seat's close of the gap that carries a failed check is refused while
  the check stands; a closing ruling of the bench closes that gap, and the check still holds the PASS.
  It clears when the anchor is checked and holds — for a proof or a found-against citation, by the
  checking lens's own later check, since another lens's check beside it is a disagreement; for an unread
  citation, by any lens's reading of the work — when an edit changes its sentence, or when blue retires
  it." `lens retirement` [REWRITE] ("no contradiction it read" goes). Then `go -C scripts run ./vocabdoc` regenerates `docs/vocabulary.md`.
- agentgen sources [REWRITE]: `scripts/agentgen/src/fragments/lens-duties.md`, `fragments/lens-audit-and-mint.md`,
  `src/red-chair.md` (census), `src/skills/research-protocol/SKILL.md:17-19` ("it goes to the board as a
  finding — a PASS is refused until one is raised for it", found by the phrase sweep below), and
  `src/lead-judge.md` gains one sentence for the bench (F8, R4): "A closing ruling on a gap that carries a
  failed check — a citation no reading reads, a proof judged unsound, a citation whose source a lens's
  reading finds against it — closes the gap and leaves the check standing: the PASS stays held until a
  check that clears it holds, blue edits the sentence, or blue retires the anchor. Your ruling decides
  the gap, never the record's state of the anchor." `fragments/blue-search-and-cite.md:28-35` (blue's failed-check duty) [REWRITE]: a citation found against
  is fixed by an edit of its sentence, or by retiring or replacing the citation, and "the lens's gap about
  the anchor, once minted" becomes "the gap that carries it". `fragments/lens-audit-and-mint.md:6` (the
  lens's failed-check duty) [REWRITE] with it: the owed gap is the unread or unsound check's; a reading
  that finds the source against the sentence records its gap in the same act; a found-against citation
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
  At `8236e3a17` it prints 57 lines (list kept at `~/.claude/scratch/plan1647-r4/phrase-wide.txt`). The
  30 beyond `cli/` and the agent-text paths, each read:
  REWRITE — `record/anchorchecks.go:279` (`VerifyReads` comment, §III.3), `record/evidenceview.go:245`
  (comment: "at high confidence it is a finding to act on" → the gap its reading minted),
  `record/recordpb/record.proto:509` (`REFUTES` means, §III.1), `record/evidenceview_test.go:251`
  (comment "THE FINDING HALF" → the negative half), `fuzz_test.go:840` (the quote comment: a negative
  corroboration's quote is the sentence its gap is minted on), and the census rows already listed —
  `blockerowner_test.go:202`, `narrowedview_test.go:34`, `redcitation_test.go:347`,
  `migrate/reproduceanchor_test.go:130`, `fuzz_test.go:5242`; DELETE — `record.proto:1127`,
  `dispatch.go:1291`, `contradictionquadratic_test.go:17`, `terms.json:765`; REGENERATE —
  `record.pb.go:737, :8243, :8301`; KEEP — `record.proto:501` (why the outcome field exists, as
  `verify.go:56`), `evidenceview.go:238` (the same history), `evidenceview.go:559` (`sources_refuted`),
  `unchecked.go:27` (the reopened-evidence history), `narrowedview_test.go:49` (fixture text),
  `capture/dispatchparity.go:208`, `debatejs/debatejs.go:388`, `changesjson.go:32`, `viewjson.go:878`,
  `report/assemble.go:1183`, `report/assemble_test.go:997, :1000`, `seatprobe/naming.go:191` (ordinary
  English).
  Known hits at `8236e3a17` (27; the 9 under `cli/*_test.go` are the census's test rows or other-sense
  KEEPs — `missingflag_test.go:119`, `stategraph_test.go:91`, `verbs_test.go:114`):
  REWRITE — `scripts/agentgen/src/skills/research-protocol/SKILL.md:17` (the sentence spanning l.17-19,
  which a line grep cannot match), `fragments/lens-audit-and-mint.md:8`, `red-chair.md:12`,
  `verify.go:304`, `verify.go:699`, `help/corroborate.md:13`, `help/finding.md:11`, `help/dispatch.md:15`
  and `:19`, `debate.js:1046`, `docs/citations-and-bibliography.md:166` ("`refutes` and `absent` are
  findings, not failures to grade"); REGENERATE — `skills/research-protocol/SKILL.md:19`,
  `docs/vocabulary.md:455`; KEEP — `verify.go:56` (a code comment on why the outcome field exists),
  `fragments/lens-audit-and-mint.md:9` (materiality), `debate.js:227` (inquiry `refuted`),
  `docs/seat-surface-naming.md:749`, `docs/why-a-seat-stops.md:103` (ordinary English).
  The property each rewritten sentence holds: a negative reading's gap is minted in the same act.
- `debate.js`: enginegen regions regenerate (`go -C plugins/frank-exchange-of-views/tools run
  ./devcmd/enginegen`); authored hits per census [REWRITE]; simulator goldens regenerate.
- Seat prompts name acts, never verbs or flags (`TestNoPromptGrowsItsCommandCatalogue`,
  `TestTheSeatPromptsNameNoVerb`): the new prompt text says "a reading that finds the source against the
  sentence records its gap in the same act", never `--class`.
- AFTER regeneration, every changed golden line and each affected `--help` page (verify, corroborate,
  finding, mint, dispatch, show) is read before and after, one command at a time, for semantic loss.

### III.8 Other Go consumers [MODIFY]

- The census (§III.12) found no FEOV-sense hit in `verify/`, `report/`, `capture/`, `seatprobe/`,
  `bluedoc/`, `flags/` or `scripts/decisions/`: they read the source-outcome sense (`refutes`,
  `sources_refuted`) or the self-contradiction classes, and KEEP. `capture.go:469-475` (found-against
  citations in the assembled report) keeps its check. `cli/chair/dispatch.go`, `cli/setup.go`,
  `cli/seat/verbs.go`, `skills/research/SKILL.md`, `references/run-directory.md` and
  `docs/citations-and-bibliography.md` are REWRITTEN.

### III.9 migrate — one translation [NEW] (contract: record migration)

Archive survey (all 32 runs in `run-archive/`, read at `8236e3a17`): the `contradiction` about-kind
exists from epoch 30/31; only **m30** and **m32** hold raising findings, all about **citation** readings;
no archive holds a raised corroboration reading.

| Archived run | Negative readings | Raises | Answering gaps |
|---|---|---|---|
| m30 (epoch 33) | 1 citation `refutes` (ev 401, `C-4f4b8cfa`) | F-8410bec7 | G-dc877265 (`within_source_condition_misattribution`) |
| m32 (epoch 41) | 4 citation `refutes` of one claim (`C-a36dce3f`, ev 80/100/147/169) | 4 | 1 (G-f417112d); 3 raised, never answered |
| m19, m23, 07d, wave-a, quadratic, sqlite-schema-source | 1–6 each, citation and corroboration, `refutes` and `absent` | 0 (kind did not exist) | 0 |

Translation (in `migrate/replay.go`, beside `contradictionRef`, which it replaces; `contradictionRef`
[DELETE]):

1. **A finding about a contradiction** becomes a finding about the reading's citation: `about_kind
   anchor`, `about_ref` = the anchor of the reading its `about_ref` named, read through `newID` from the
   migrated record. `requireAboutAnchor` (`refs.go:126`) admits any citation anchor. A raise naming a
   **corroboration** reading (no anchor) is refused with the finding's id named — no archive holds one,
   so the refusal is the loud miss, not a silent zero.
2. **An archived negative reading** carries no `carried_by`: its gap was minted later (forward reference) or
   never. migrate does NOT fill `carried_by` from the `found_by` chain (ruled, gblock 2026-10-30). Its
   found-against state is still a failed check on the migrated record (read from the ledger), with no
   named carrier.
3. **A mint whose `found_by` names a raising finding** is carried unchanged — the finding exists.
4. **A stored plan's `contradiction` blocker** becomes a `failed_check` blocker whose subject is the
   reading's anchor and whose owner is unchanged (the reader, as a live found-against blocker's is).
   `TranslatePlanKinds` cannot do it: it is a pure kind rename (`planKindRenames`) with no record
   access, and the archived subject is the reading's `events.id`, not an anchor. So [NEW]
   `TranslateReadingBlockers(pl, newID, q)` in `migrate/planids.go`, called in `replay.go:337-350`
   after `TranslatePlanKinds` and before `TranslatePlanIDs`: for each blocker whose kind is the archived
   word `contradiction` (a literal in migrate's archived-word table; the live `BlockerContradiction` is
   deleted), it carries the subject through `newID`, reads the Verify the migrated record holds at that
   `events.id` through the batch's querier (`batch.Querier()`, as `CarryPlanVerdict` reads it), and sets
   kind `failed_check`, subject = that Verify's `anchor`; then re-keys and re-prints the plan. A subject
   that wrote no event, or names a Verify with no anchor (a corroboration), is a refusal naming the plan's
   event and the reading — the loud miss. `TranslatePlanIDs`'s contradiction branch (`planids.go:91-103`)
   [DELETE]; the `contradiction` id phrase in `planIDPhrases` stays — it translates ids inside the
   archived `why` prose, which carries the words it was printed with. m32 holds two such plans (events
   223 and 310, each with three `contradiction` blockers, subjects 100, 147, 169, owners
   red-lens-architecture, red-lens-computation, red-lens-adversary).
5. Migrate's comments in `entries.go:64-66` (the credit entry) and `outcomes.go:19` [REWRITE] to state
   archived bytes without the live concept word where they describe the live model.

**Residue (stated, accepted in this plan; ruled, gblock 2026-10-30):**
- Archived readings name no gap, so the close gate cannot tie an archived answering gap (m30
  `G-dc877265`, m32 `G-f417112d`) to its reading: in a run resumed after migration that gap may close
  while its reading still stands. Accepted. A carrier-less found-against check readies blue alone, sits
  on its reader's work list as not owed, holds no reader from retiring and admits no budget-exempt mint
  (§III.3); it clears by the ledger's rule — its reader's own later reading of the sentence as it stands
  that backs it, an edit, or a retire. Where the archive held the sentence on a contradiction, the
  migrated record holds it on a failed check.
- m32: each of its four negative readings of `C-a36dce3f` that still stands at the end is its own
  found-against failed check (up to four, one per reader) and its four raising findings become four
  findings about that anchor. m32 ended UNVERIFIED, so no verdict moves.
- Epoch-24/25/26 negative readings (m19, m23, 07d, wave-a) never had a raise; they come forward as
  readings on the ledger exactly as before.

Tests [NEW], in `migrate/`: `TestTheMigratedM30RaiseIsAFindingAboutItsCitation` (F-8410bec7 →
`about anchor C-4f4b8cfa`; G-dc877265 open/closed as archived; 0 refusals) and
`TestTheMigratedM32HoldsAFailedCheckPerStandingReading`, which also asserts the two migrated stored plans
(the events m32 wrote as 223 and 310, read through `newID`): each holds `failed_check` blockers with
subject `C-a36dce3f` and owners red-lens-architecture, red-lens-computation and red-lens-adversary, no
blocker of kind `contradiction`, and a key that verifies (`requirePrintedPlan` admitted it at the
write); a delete of `TranslateReadingBlockers`'s call turns it red (a refusal, or a numeric subject);
`TestAMigratedFoundAgainstCheckReadiesBlueAlone`, on
migrated m32, asserting for each found-against failed check of `C-a36dce3f` (state `found_against`,
`AnsweredBy() == ""`):
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
`TestM23CorroborationReadingComesForward`
(the unraised corroboration and citation readings, 0 refusals). `reopenedm19_test.go` and
`TestEveryArchivedRunMigrates` keep passing.

### III.10 Tests [MODIFY / DELETE / NEW]

- DELETE the four files that exist for the concept (§III.12): `integration/surface/contradictionanswer_test.go`,
  `record/contradictionquadratic_test.go`, `record/contradictionref_test.go`, `record/answerbudget_test.go`.
  `cli/m23contradiction_test.go` keeps its quote-refusal half and is renamed for what it then tests.
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
    L's close of G2 is refused naming C; L's repeat negative reading is refused naming G2. Partner §V.5.o.
  - `TestAReadingsGapTakesNoMint` (`cli/`, driven through the `verify` verb as a seat runs it, so the
    Mint's budget check meets the act's paired Verify before the Verify is written): a lens at its cap
    records a refuting reading; the reading and its gap land in one transaction; the lens's count is
    unchanged after; a plain `mint` at that cap is still refused. Partner §V.5.e.
  - `TestAPlainMintALaterReadingNamesKeepsItsMint` (`cli/`): L mints G with plain `mint`, class
    `claim_contradicts_source`, quoting sentence S; a sitting later L records a `refutes` reading of S
    with that class — the one-gap rule refuses the reading's mint, `carried_by` = G, `carrier_minted`
    false; L's open-gap count still counts G (at cap, a further plain mint is refused). Partner §V.5.p.
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

- Delete the contradiction drives and census fields: `answerContradictions` (`fuzz_test.go:965-966`), the
  `ownAnswers`/`answerBudgetErr` fields (`:494-497`), the sweep gates at `sweeptally_test.go:860, :894`
  and every other site the census lists.
- The lens's negative reading draws the gap flags (class from the outcome's admitted set, grades, check)
  in the same act. New gates (false-green guards, zero fails the sweep): negative readings that minted;
  negative readings refused a mint by the one-gap rule (the duplicate path is driven on purpose by
  re-reading a sentence already holding a reading's gap); found-against closes refused. New tally:
  max open gaps per (class, sentence) — must be ≤ 1.
- **Every fuzz branch that tells failed checks apart keys on `CheckState()`**, and the `unread*` tallies
  count the `unread` state only. Sweep, run at `8236e3a17`:
  `grep -n -E 'BlockerFailedCheck|AnsweredBy\(\)|\.Detail\b|unread(Answers|Repeats|Retires|RetiresJoined|ClosesRefused)|unsound(Answers|Repeats|ClosesRefused)' releasegate/fuzz/*.go`,
  every hit read. Under a found-against check `Detail == ""` reads as "unread citation", so each of these
  would drive or count it as unread:
  - `fuzz_test.go:1063, :1072` (`fixFailed`, blue's remedy): `Detail == ""` → unread; it would call
    `envelopeRefusedEarly(… "check of citation "+C)` expecting the unread wording, retire the citation and
    count `unreadRetires`/`unreadRetiresJoined`. [REWRITE] a switch on the state: `unread` as today;
    `found_against` drives blue's found-against remedy — by coin, the edit that rewrites the sentence
    (`reopened` names C) or the retire `--answers` the carrying gap — with
    `envelopeRefusedEarly(… "reading of citation "+C+" at event")`, the found-against `failedFact`'s
    words, and counts `foundAgainstEdits` / `foundAgainstRetires`; `unsound` as today; any other state is
    a `noteActErr`, never a fall-through;
  - `:2814` (unread repeat: `verify --as unreachable` expected refused): `CheckState() == unread` only —
    an `unreachable` reading over a found-against check is admitted, so driving it as refused would
    fail the run; `:2824` (unsound repeat): `== unsound`;
  - `:2846, :2849` (the owed-gap mint, `citation := Detail == ""`, its early-envelope refusal and the
    `unreadAnswers` count): only `unread`/`unsound` with no carrier — `found_against` is never owed;
  - `:2903-2912` (`driven := unreadClosesRefused > 0`, the drive of a close refused at the mint) and
    `:4052-4076` (`refuseStandingClose`'s tally, `Detail == ""` → `unreadClosesRefused`): the tally
    switches on the state — `unread`, `unsound`, `found_against` → `foundAgainstClosesRefused` — and
    any other state is a `noteActErr`;
  - `:3210` (blue's oracle, "Retire the citation" for every `Detail == ""`): `unread` only; a
    `found_against` check's item must name the found-against remedy ("edits the sentence, or retires or
    replaces the citation");
  - `:3503` (`ownUnreadGapOpen`): `== unread`;
  - `:3995, :4040` (`standingFailedCheckOf` → `refuseStandingClose` in `closeGap`): state-agnostic
    selection by carrier, kept, and now the successor-following `AnsweredBy()` (§III.3); the tally it
    reaches switches as above;
  - KEEP: `:2760` (asserts the plan readies the owner or blue, true of every state), `:6700` (no
    VERIFIED over any failed check);
  - `sweeptally_test.go:788-799`: the `unread*` gates stay and now count only the unread path; NEW
    gates (zero fails the sweep, false-green guards) `foundAgainstClosesRefused`,
    `foundAgainstEdits + foundAgainstRetires`, and `foundAgainstClearedByOwnReading` (the drive below);
    each new field is carried through `sweepTally`, the fold and its merge (`sweeptally_test.go:171-172,
    :509-511, :620-622`, `fuzz_test.go:463-465, :5470-5472, :6765-6767`), which
    `TestSweepTallyRoundTripsEveryField` and `TestTheFoldReadsEveryRunFieldAndWritesEveryFoldField`
    already fail on a field left out.
  New drive: a refuting reading followed, a sitting later, by a backing re-read by the same lens, after
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

Searches, run at `8236e3a17` (raw hits kept at `~/.claude/scratch/contra-census/hits.txt`, context in
`ctx.txt` beside it):

```
grep -rIn -i contradict --exclude-dir=.git --exclude-dir=run-archive .
grep -rIn -E 'ABOUT_KIND_CONTRADICTION|unanswered_contradictions|ContradictionOfReading|raisedBy|answeredBy|raised_by|opened_by|sources_refuted' --exclude-dir=.git .
```

Result: **201 files** (198 carry the word, 3 only the identifiers). Each hit was read for meaning.
**DELETE 4 · REWRITE 62 · REGENERATE 16 · KEEP 121** (38 of the KEEP are historical). The two REWRITEs
beyond the word census are `help/close.md` and `fragments/blue-search-and-cite.md`, found by the
failed-check sweep below; its other hits are in files the word census already holds. Round 4 adds, from
the sweeps in §III.3, §III.7 and below: REWRITE `help/reproduce.md`, `docs/seat-command-triggers.md`,
`recordpb/facets.go`, `cli/keyedretry_test.go`, `unreadcitation_test.go`, `unreadb6_test.go` (none
carries the word), and moves `enums_test.go` and `evidenceview_test.go` from KEEP to REWRITE — totals
**DELETE 4 · REWRITE 70 · REGENERATE 16 · KEEP 119** over 207 files. Paths are relative
to `plugins/frank-exchange-of-views/tools/internal/` unless shown. "n" = lines carrying the FEOV sense.

**Production Go — record**

| File | n | Carries | Disposition |
|---|---|---|---|
| record/citationid.go | 52 | the fold, `contradiction`, `contradictionActs`, `ContradictionOfReading`, `requireOpenContradiction`, `unansweredContradictions`, `idWord`; `contradicts()` (l.666) | REWRITE: fold deleted, predicate → `gap_classes` facet |
| record/passblockers.go | 33 | `BlockerContradiction`, `raisedBy`/`raiser`, four constants, `contradictionFact`, PASS text, producer l.678 (l.846 ordinary) | REWRITE |
| record/mintbudget.go | 17 | own-raise exemption, `Answering`, `mintAnswerOutsideBudget` | REWRITE (§III.5) |
| record/dispatch.go | 14 | `contradictionReadiness`, lens arm, `GapIDs`/`Subject` desc (feeds enginegen) | REWRITE |
| record/onegapperclass.go | 10 | `answersRaisedContradictionAt`, exemption | REWRITE |
| record/evidenceview.go | 11 | `UnansweredContradictions`, `EvidenceContradictionJSON` | REWRITE |
| record/anchorchecks.go | 8 | comments; `contradicts()` at l.261 | REWRITE (§III.3) |
| record/sitting.go | 4 | lens owes the blocker; raiser-at-cap rule (l.160-174 ordinary) | REWRITE |
| record/about.go | 5 | about-kind validation, gap refusal | REWRITE |
| record/record.go | 4 | finding → `requireOpenContradiction`; mint about-kind refusal | REWRITE |
| record/enums.go | 2 | finding's set "less contradiction" | REWRITE |
| record/retirement.go | 2 | "a contradiction it read" as a debt | REWRITE |
| record/narrowedview.go, record/convergence_refusal.go | 1 each | comments | REWRITE |
| record/recordpb/record.proto | 6 | `ABOUT_KIND_CONTRADICTION`, comments l.136-137, 890, 1156-1157 (other hits: class names, `refutes`, `refuted`) | REWRITE (§III.1) |
| record/recordpb/record.pb.go, record.proto.sha256, schema_gen.go | gen | — | REGENERATE protogen / schemagen |
| record/recordsql/testdata/schema.sql | 1 | `enum_about_kind` row | REGENERATE (golden) |
| record/migrate/replay.go | 7 | `contradictionRef` | REWRITE (§III.9) |
| record/migrate/planids.go | 6 | `"contradiction"` id phrase, blocker subject | REWRITE (§III.9) |
| record/migrate/entries.go, migrate/outcomes.go | 3, 1 | comments | REWRITE |
| record/migrate/plankinds.go | 0 | — | MODIFY (new respelling, §III.9) |
| record/migrate/gapclass.go, enumof.go, outcomeclosed.go, spotcheck.go, failedcheck.go (`answeredBy` = failed check), recordsql/views.go (`opened_by` = gap view) | 0 | other sense | KEEP (failedcheck.go is MODIFIED by §III.3, not for this word) |

**Production Go — cli**

| File | n | Carries | Disposition |
|---|---|---|---|
| cli/lens/verify.go | 16 | result `contradiction{id,opened,opened_by}`, opened/joined text | REWRITE (§III.2) |
| cli/lens/finding.go | 2 | `--about-kind`/`--about` help | REWRITE |
| cli/chair/dispatch.go | 1 | comment | REWRITE |
| cli/seat/verbs.go | 3 | `unanswered_contradictions` manual text, budget text | REWRITE |
| cli/setup.go | 1 | `--mint-budget` help | REWRITE |
| cli/seat/help/corroborate.md, verify.md, dispatch.md, mint.md, finding.md | 2, 2, 3, 2, 1 | contradiction prose (finding.md l.9 "internal contradiction" stays) | REWRITE |
| cli/lens/mint.go | 0 | ordinary English | MODIFY (shared flag helper, §III.2) |
| cli/seat/help/record.md, cli/bench/outcome.go, cli/seat/seat.go | 0 | inquiry `refuted` / ordinary | KEEP |

**Other Go** — `terms/terms.json` (5: the entry, its ordinary-English collision, `lens retirement`): REWRITE
then vocabdoc. KEEP (0 FEOV-sense hits; source-outcome, class or ordinary sense): capture/capture.go
(`sources_refuted`, found-against check), capture/proofrerun.go, report/assemble.go,
report/registrations.go, verify/verify.go, seatprobe/boards.go (`cross_section_contradiction`),
seatprobe/production.go, bluedoc/bluedoc.go, flags/names.go, feov/error.go, seatenv/identity.go,
seatenv/seatenv.go, tessocr/engine_stub.go, proof/proof.go, `scripts/decisions/main.go`,
`scripts/golden/review.go`, gray-area `cmd/gray-area-capture/main.go`, `internal/claims/coverage.go`,
`internal/claims/manifest.go`.

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
- REWRITE beyond the word: record/enums_test.go (`narrowedBy`'s `gap_may_name` row, deleted with the
  facet), record/evidenceview_test.go:251 (the "FINDING HALF" comment, phrase sweep),
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
  proof/portability_test; scripts/golden/review_test; gray-area main_test, coverage_test.

**Release fuzz** (`releasegate/fuzz/fuzz_test.go`): drive `answerContradictions`, `raisePrinted`,
`ownUnansweredRaise`, `raiseAndMint`, `contradictionAbout`; oracles `quoteRaisesNothing`,
`contradictionUnraised`; parser `contradictionOf`; call sites l.5040, 5162-5174, 5242-5260, 5337, 8567;
fields `ownAnswers`, `answerBudgetErr` (l.494-499, 3384-3413, 3425). `sweeptally_test.go`: `ownAnswers`
tally (l.512, 628, 892-893), gate wording l.860. All REWRITE (§III.11). PR #1642: superseded.

**Goldens / testdata** — REGENERATE: difftest/testdata/manual_runs_every_help_page_live.golden (6 lines),
`tests/simulator/testdata/prompt-red-chair.golden`. REWRITE: `tests/simulator/testdata/o3-chair5-relays.json`
(a captured plan with blocker kind `contradiction`; the schema test fails once enginegen drops the value).
KEEP: difftest error_catalogue.golden, migrate/testdata/prose-rewrites.golden (archived prose),
prompt-red-lens-logic/dark-side goldens, debate.test.mjs, prompts.test.mjs. gap_class_enum.golden
REGENERATES for the new class only.

**debate.js** (`skills/research-protocol/scripts/debate.js`): l.398 and l.420 sit inside the
`GENERATED ENVELOPE VOCABULARY` region (l.373-443) → REGENERATE (enginegen); l.1046 (red-chair prompt, "the
lens that read a contradicting source no finding raises") is authored → REWRITE; l.227, 253, 255 KEEP.

**Agents and skills** — REGENERATE (agentgen): all 11 `agents/*.md` (7 red-lens-*, red-chair, blue-researcher,
blue-synthesizer, lead-judge). REWRITE: `skills/research/SKILL.md` l.7,
`skills/research-protocol/references/run-directory.md` l.77. The two generated preloaded skills have no hits.

**agentgen sources** (`scripts/agentgen/src/`) — REWRITE: fragments/lens-audit-and-mint.md,
fragments/lens-duties.md, red-chair.md. KEEP: lead-judge.md, blue-researcher.md,
fragments/blue-research-protocol.md (ordinary English).

**Docs** — REGENERATE: `docs/vocabulary.md` (vocabdoc). REWRITE: `docs/citations-and-bibliography.md` l.185
and l.151-160 (the failed-check bullet: two states, a gap that "cannot close", and verify without the gap
act → three states, the gap minted at the reading, the close refused while the check stands and a
closing ruling closing the gap while the check holds the PASS); `docs/seat-command-triggers.md` l.56 (the
`lens verify` row: a negative reading mints its gap in the same act; an unread reading owes the gap
about the citation), l.66-67 (the `lens mint` row: the owed failed-check gap is the unread or unsound
check's; the `lens close` row: a gap carrying a found-against check is refused too, until the reader's
own backing reading, an edit or a retire), l.87 (the `blue retire` row: an anchor found against is
blue's to edit, retire or replace).
KEEP: propagation-and-anchoring.md, seat-surface-naming.md, why-a-seat-stops.md; outside FEOV,
`plugins/gray-area/README.md`, `plugins/gray-area/skills/audit-seat-coverage/SKILL.md`,
`plugins/prosthetic-conscience/skills/resume/SKILL.md` (ordinary English).

**Phrase-sweep additions** (old model, no word; §III.7): `scripts/agentgen/src/skills/research-protocol/SKILL.md:17-19`
→ REWRITE (regenerates `skills/research-protocol/SKILL.md`); `cli/lens/verify.go:304, :698-701` → REWRITE.
`scripts/agentgen/src/lead-judge.md` → REWRITE (one added sentence, R4).

**Plans / ideas / law / feov-memory** — KEEP-historical (38 files). FEOV-sense mentions in
plans/reopened-evidence-is-owed.md, feov-lens-bar.md, gap-class-enum.md, pre-dry-run-batch.md,
one-item-per-dispatch.md:336, ocr-cite-derives-page.md:574, historical/red-citations.md,
historical/migrate-staged-publish.md:284, law/precedents.md:156 record past decisions; none is a live
instruction, none is rewritten.

**Readers of a refutation that do not name the contradiction** (added at the Slim-fast peer's request;
each states how it reads once a refutation is a gap):

| Site | Today | After | Disposition |
|---|---|---|---|
| Anchor ledger `anchorStanding.state()` (`anchorchecks.go:152-178`, from #1601 / 90f3fa008) and `stateFoundAgainst` | a citation whose standing check refutes or is absent; its PASS hold was "the contradiction's" | Unchanged predicate: found-against is still the **citation's standing** (a standing negative reading of the sentence the report carries), NOT "has an open gap of the new class". The gap is a separate fact linked by `Verify.carried_by`; the state is what makes the gap unclosable (§III.3). Keying the state on the gap would let a gap closed into a successor, or an archived reading with no gap, read as holding | MODIFY (comment only; `contradicts()` → `gap_classes` facet) |
| `failed_check` PASS blocker (`passblockers.go:397`, `failedcheck.go`) | unread citations and unsound proofs | gains the found-against state (§III.3): holds the PASS until its reader's own later reading backs the sentence, blue edits it, or blue retires the citation (another lens's backing reading clears nothing); readies blue alone; carrier = `Verify.carried_by` followed through its successors | MODIFY |
| `unchecked` PASS blocker | a citation or proof with no check of its current sentence | unchanged: an edit of a found-against sentence makes the citation unchecked, owed a fresh check by its owning lens; that check either holds (failed check cleared, the gap may close) or is negative again (names the standing gap, or mints one if it is closed) | KEEP |
| Evidence lens's gap duties (`fragments/lens-duties.md`, `lens-audit-and-mint.md`, `verify.md`) | read every citation; a negative reading opens a contradiction it must raise by finding, then answer by a gap minted from it, outside its budget | reads every citation; a negative reading carries the gap's judgment (class among the outcome's admitted set, problem, check, grades) and mints it in the same act, outside its budget; a second reading of a sentence holding that gap names it; the lens closes the gap with its ordinary close once its own later reading backs the sentence, or blue edits or retires it. Unread readings keep the owed-gap-about-the-anchor duty | REWRITE (agentgen) |
| Failed-check owed-ness readers (§III.3): dispatch lens arm `dispatch.go:531-538`, lens work item `sitting.go:396-403` (and `EnvelopeRefusal`, `envelope.go:44`, through it), retirement `retirement.go:164-172`, budget `failedcheck.go:109-119, :208-215` | owed iff `answeredBy == ""` | owed iff state is `unread`/`unsound` and no gap carries it; `found_against` never owed by its reader | MODIFY |
| Failed-check printed sentences (§III.3): `passblockers.go` `failedFact` :220, `failedClears` :228, `failedRemedy` :239, `failedOwed` :248, `blueFailedItem` :256, the kind's gloss/refusal/why :398-427; `dispatch.go:1297` `failedCheckReadiness`; `failedcheck.go` repeat and close refusals :134, :178 | branch on `Detail` (proof) and `answeredBy`; citation = unread | branch on the state; found-against wording names the reading and blue's remedy (edit, retire or replace), never an owed mint | MODIFY |
| Failed-check prose: `help/mint.md:15`, `help/close.md:11`, `help/verify.md:11`, `help/reproduce.md:21`, `fragments/blue-search-and-cite.md:28-35`, `fragments/lens-audit-and-mint.md:6`, `terms.json` `failed check`, `docs/citations-and-bibliography.md:151-160`, `docs/seat-command-triggers.md:56, :66-67, :87` | two states; the seat whose check failed owes the gap; "the gap does not close while the check stands" | three states; the owed gap is the unread/unsound check's; a found-against check's gap is minted at the reading, and the check clears by the ledger's rule; the seat's close is refused while the check stands, and a closing bench ruling closes the gap while the check holds the PASS (F8) | REWRITE |
| Fuzz failed-check drives and tallies `fuzz_test.go:1063, :1072, :2814, :2824, :2846, :2849, :2903-2912, :3210, :3503, :3995, :4040, :4072`; gates `sweeptally_test.go:788-799` | `Detail == ""` = unread citation; `AnsweredBy() == ""` = owed; unread tallies count every citation check | key on `CheckState()`; unread tallies count `unread` only; found-against has its own tallies and gates | REWRITE (§III.11) |
| `verify.go:529-541, :631-656` `failedCheckWords` and the JSON `failed_check` | guarded by the NEW reading's `!VerifyReads`, worded as unread whatever the check's state; JSON carries no state | after every verification of a citation, every standing failed check of the anchor, worded by its state; JSON `failed_checks[].state` | REWRITE (§III.2) |
| `verdict.go:141-162`; `debate.js:1220-1227`; `fuzz_test.go:2760, :6700` | anchor names only; owner-or-blue readied; no VERIFIED over a failed check | unchanged, true of every state | KEEP |
| Tests pinning unread/unsound sentences: `unreadcitation_test`, `unreadb6_test`, `blueanswered_test`, `failedcheck_test`, `reopened_test`, `tiersubstitution_test`, `unsoundm32_test`, `capture/dispatchparity_test` | those states' wording | their owed-mint and remedy wording does not move — they are the regression net for it; the "does not close" comments in `unreadcitation_test.go:119` and `unreadb6_test.go:21-22, :66` are rewritten (F8) | KEEP (two comment REWRITEs) |

**Readers of a failed check's owed-ness and its printed sentence** (the class the round-2 audit named;
run at `8236e3a17`, every hit read; dispositions in the table above):

```
grep -rIn -E 'answeredBy|answered_by|AnsweredBy' plugins/frank-exchange-of-views/tools/internal plugins/frank-exchange-of-views/tools/cmd plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js scripts/agentgen/src | grep -v _test.go
grep -rIn -E 'BlockerFailedCheck|failedChecksOf|answersOwnFailedCheck|ownFailedCheckAnswers|failedRemedy|failedOwed|blueFailedItem|failedCheckReadiness|failedCheckWords|failed check' <same paths> | grep -v _test.go
grep -rIn -i -E 'check (of yours|that) failed|whose check failed|failed check|about the anchor|about that anchor' scripts/agentgen/src plugins/frank-exchange-of-views/tools/internal/cli/seat/help plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js
```

Hits outside the contradiction fold (which is deleted: `citationid.go:516-647`, `onegapperclass.go:145`,
`mintbudget.go:360-378`): `passblockers.go:83-97, :217-268, :397-427, :685`; `failedcheck.go:44-119,
:134-215`; `sitting.go:321-322, :396-403`; `retirement.go:142-172`; `dispatch.go:21, :531-546,
:1294-1312`; `mintbudget.go:306-307, :368, :392, :398-408`; `verdict.go:141-162`; `cli/lens/verify.go:529-541,
:631-656`; `cli/seat/verbs.go:297`; `terms.json:1041-1042`; `help/mint.md:15`, `help/close.md:11`,
`help/verify.md:11` (unread; its "does not close" clause REWRITEs, F8), `help/dispatch.md:19`; `fragments/lens-audit-and-mint.md:6`,
`fragments/blue-search-and-cite.md:28-35`; `debate.js:398` (enginegen, from `dispatch.go:21`),
`debate.js:1220-1227`; `onegapperclass.go:28` (a comment naming `answersOwnFailedCheck`, KEEP);
`migrate/plankinds.go:19` (respelling, KEEP); test files under `record/`, `capture/` and the fuzz (rows
above). `sitting.go:268, :662-664` (`answeredByActThisSitting`) is a motion reader, other sense, KEEP.

**Readers of the deleted `gap_may_name` facet** (`grep -rn -i 'gap_may_name\|GapMayName'` over the tools
tree, every hit read): `recordpb/record.proto:134-138, :1124-1128` (declaration and values) [DELETE];
`recordpb/facets.go:35` (`"gap_may_name": E_GapMayName`) [DELETE]; `record/enums.go:347-363`
(`gapAboutKinds`) [DELETE] and its two callers, `enums.go:300` (mint's `about_kind` values → the full
`AboutKind` set, as finding's) and `record.go:970` (the mint's about-kind refusal, §III.6) [DELETE];
`record/enums_test.go:339-341` (the `{"mint","about_kind"}` row of `narrowedBy`) [REWRITE: the row is
deleted, since mint's set no longer narrows the enum]; `record.pb.go` [REGENERATE].

**Readers of the new Verify fields** (`carried_by`, `carrier_minted`): by name — `record.go:483`
`keyFields` (neither name is on it, so a Verify keys on `label`/`anchor`/`url` as today; the reason the
field is not called `gap_id`, §III.1), `flags/names.go:393` `payloadFlag` (falls back to `--carried-by`,
never printed: both fields are tool-supplied, `seat.Supplies` in `verify.go` and `corroborate`), and the
required-field sweep (`recordpb/requiredfields_test.go`, both optional); by meaning — `failedChecksOf`
(carrier), `requireMintWithinBudget`/`lensMintStanding` (`carrier_minted`, §III.5), the repeat rule
(§III.2), the reading's result (`gap.minted` read back), migrate (archived readings carry neither,
§III.9.2), the sweep tally `open_by_found_by` (§III.11) and §V.8's query. `references: "mint.gap_id"`
makes the column a foreign key; replay translates no gap id (G- ids are carried unchanged), so migrate
needs no translation for it.

**Carried by the concept, not found by the word:** `record.proto.sha256`, `requirements.json` (epoch),
`schema_gen.go`, `migrate/plankinds.go`, `failedcheck.go` (§III.3), `cli/lens/mint.go` (flag helper),
`recordpb/facets.go` and `enums_test.go` (the facet), `help/reproduce.md:21` and
`docs/seat-command-triggers.md` (the failed-check prose, F8), `cli/keyedretry_test.go` (§III.2's retry
test), `unreadcitation_test.go` (the F8 bench test lands beside its unread mirror).

### III.13 Directory tree of touched paths

```
plugins/frank-exchange-of-views/
  requirements.json                         [MODIFY] eventSchema 49
  tools/internal/record/
    recordpb/record.proto                   [MODIFY] AboutKind, GapClass, SourceOutcome facet, Verify.carried_by, Verify.carrier_minted
    recordpb/facets.go                      [MODIFY] gap_may_name deleted
    recordpb/record.pb.go, schema_gen.go    [REGENERATE] protogen, schemagen
    recordsql/testdata/schema.sql           [REGENERATE]
    citationid.go                           [MODIFY] contradiction fold deleted
    successor.go                            [READ] successorsOf, reused by failedChecksOf
    anchorchecks.go, failedcheck.go         [MODIFY] found-against is a failed check
    onegapperclass.go, mintbudget.go        [MODIFY]
    passblockers.go, sitting.go, dispatch.go, retirement.go, evidenceview.go,
    narrowedview.go, about.go, record.go, enums.go, convergence_refusal.go   [MODIFY]
    migrate/replay.go, plankinds.go, planids.go, entries.go, outcomes.go     [MODIFY]
    migrate/*_test.go                       [NEW] m30, m32, m23
  tools/internal/cli/lens/verify.go, mint.go, finding.go   [MODIFY]
  tools/internal/cli/seat/help/*.md         [MODIFY]
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
Its cost is dominated by the census rewrite (70 REWRITE incl. 27 tests, 4 DELETE, 16 REGENERATE), not
machinery: one facet, two Verify fields, one append path, one ledger state widened, one check state read.

Implementation runs one phase per **fresh agent**, each starting from this plan and the branch's
`.claude/checkpoints/CHECKPOINT.md`, ending with a commit, the checkpoint rewritten (head, the phase's
loop result, the next phase), and the agent's exit. A phase ends where its loop is green; the full
`check` is owed only at phase 6.

| Phase | Scope | Phase loop (must pass to end it) |
|---|---|---|
| 1. Record | §III.1 proto, epoch, facet, `Verify.carried_by`; §III.3 the check state and every reader of it; §III.4; §III.5; §III.6 deletions; record-package tests of §III.10 | protogen/schemagen clean (§V.1); `go test -count=1 ./internal/record/` (excluding `migrate/`) |
| 2. Reading act | §III.2 flags, two-body append, held gap, repeat, result words; cli tests | `go test -count=1 ./internal/cli/...` |
| 3. migrate | §III.9 translation and the m30/m32/m23 tests | §V.6 |
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
| R6 | Epoch collision with Slim-fast's epoch PR (50) or #1544 Part 1a | M × M × L | Re-derive at rebase; `schemaepoch_test.go` and setup's epoch refusal catch a stale number |
| R7 | Goldens conflict semantically with a help-changing PR merged meanwhile | H × M × L | Rebase, regenerate at `-count=1`, read each changed line, re-run CI before merge (CLAUDE.md) |
| R8 | The two-event act leaves a Verify without its Mint on crash, or a retry records a second pair | L × H × L | One transaction, so a crash leaves nothing; a retry after the commit meets the repeat refusal naming the stored reading and its gap, nothing recorded (§III.2; `TestARetriedNegativeReadingIsRefusedNamingItsGap`) |
| R9 | Release fuzz loses coverage of the duplicate path | M × M × L | New zero-gates (§III.11) fail the sweep if the one-gap path in the reading act is never driven |
| R10 | A reader of a failed check keeps keying on `answeredBy == ""` or `Detail == ""`, so a found-against check is owed by its reader (never retires, envelope refused), prints an owed mint or "stays unread" that does not exist, or feeds the fuzz's unread tallies and keeps their false-green guards green | M × H × L | §III.3 state key on every reader, the verb's sentence included (§III.2); the fuzz keyed on the state with its own tallies (§III.11); the failed-check sweep re-run as §V.4; `TestAMigratedFoundAgainstCheckReadiesBlueAlone`, `TestAFoundAgainstCheckIsWordedByItsState`, `TestAReadingOfAFoundAgainstCitationIsWordedByItsState`, `TestTheFailedCheckDrivesKeyOnTheState`; delete-the-row checks §V.5 i–l, q |
| R11 | A carrier is closed into a successor, or named by a reading that did not mint it, and the close gate or the budget reads the wrong gap | L × M × L | The carrier follows `Close.successor` (§III.3); the budget exemption reads `carrier_minted`, set only in the act that minted (§III.5); §V.5 o, p |

## V. Verification Plan

All commands from the worktree root, after `unset $(env | grep -o '^FEOV_[A-Z_]*')`.

1. `go -C scripts run ./protogen && go -C scripts run ./schemagen && go -C scripts run ./agentgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen && go -C scripts run ./vocabdoc` → no diff on a second run. Re-armed by: `record.proto`, `requirements.json`, `scripts/agentgen/src/**`, `terms.json`, the roster/enums.
2. `go -C scripts run ./check` → `0 failed` (read the count, not the exit code). Re-armed by: any file in the diff. Includes the `-check` gates of every generator, `archaeology`, `golden`, `pluginparity`, vocabulary prose.
3. `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/... ./internal/cli/... ./integration/... ./releasegate/...` (release gate unset: the fuzz sweep does not run) → 0 FAIL. Goldens: `UPDATE_GOLDENS=1 go test -count=1 <pkg>` then read every changed line. Re-armed by: any Go or help change.
4. Census re-run (must print only KEEP lines of §III.12):
   `grep -rIn -i contradict --exclude-dir=.git --exclude-dir=run-archive plugins scripts docs | grep -v -E 'cross_section_contradiction|claim_contradicts_(own_record|source)'`
   and identifiers, which must print nothing outside `migrate/` archived-word tables:
   `grep -rIn -E 'ABOUT_KIND_CONTRADICTION|BlockerContradiction|unanswered_contradictions|contradictionsOf|ContradictionOfReading|requireOpenContradiction|answersRaisedContradiction|answersOwnRaise|gap_may_name|raised_by' plugins scripts`;
   the phrase sweep of §III.7, over the whole tools tree — `tools/internal` (not only `cli/`) and `tools/releasegate` — and the agent-text paths (must print only its KEEP lines); the "does not close" sweep of §III.7 (every hit in the failed-check sense says the close is refused, never that the gap does not close); the failed-check sweep of §III.3 / §III.12 over `tools/internal`, `tools/cmd`, `tools/releasegate` and `debate.js`, every `answeredBy`/`AnsweredBy()`/`.Detail`/`BlockerFailedCheck` hit read for a branch that decides owed-ness, wording or a tally without the state (must find none); and `grep -rn 'carried_by\|carrier_minted'` over the tools tree, every hit one of §III.12's "Readers of the new Verify fields".
   Re-armed by: any agent-facing or Go edit.
5. Delete-the-row checks (each must turn the suite red; revert after):
   a. delete `(gap_classes)` from `SOURCE_OUTCOME_REFUTES` → `TestARefutingReadingMintsItsGapInTheSameAct` and `TestEveryNegativeOutcomeNamesItsClasses` fail;
   b. delete the Mint body from the reading's append (keep the Verify) → the same test fails;
   c. delete the held-gap branch in the reading path → `TestASecondRefutingReadingNamesTheStandingGap` fails;
   d. delete the found-against branch of `failedChecksOf` → `TestTheLensCloseOfAReadingsGapIsRefusedWhileTheCitationStandsFoundAgainst` and `TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld` fail;
   e. delete the act-paired exemption in `requireMintWithinBudget` → `TestAReadingsGapTakesNoMint` fails (the reading's Mint is refused at the cap);
   f. delete migrate's raise translation → `TestTheMigratedM30RaiseIsAFindingAboutItsCitation` fails (refusal);
   g. delete the call site of the new flag-registration helper in `verify.go` → help/flag tests fail;
   h. delete the write-path rule that a negative reading names its gap (`record.go`, the Verify case) → a test writing a negative reading with no gap fails;
   i. revert `retirement.go`'s failed-check skip to `answeredBy != ""` → `TestAMigratedFoundAgainstCheckReadiesBlueAlone` (retirement) fails;
   j. revert `sitting.go`'s lens arm to `answeredBy == ""` → the same test (work list, envelope) fails;
   k. delete the `found_against` branch of `failedOwed` / `blueFailedItem` → `TestAFoundAgainstCheckIsWordedByItsState` and the m32 test (printed sentence) fail;
   l. restore `verify.go`'s `!VerifyReads(stored)` guard and the unread-only `words()` → `TestAReadingOfAFoundAgainstCitationIsWordedByItsState` rows (a), (b), (d) fail;
   m. drop the negative-reading arm of `requireNotARepeatFailedCheck` → `TestARetriedNegativeReadingIsRefusedNamingItsGap` fails (a second Verify lands);
   n. make `failedChecksOf` skip a found-against check whose carrier is closed (the shape that would let a ruling clear the check) → `TestABenchRulingClosesAFoundAgainstGapAndThePassStaysHeld` fails (PASS admitted);
   o. read the carrier from `carried_by` alone, without `successorsOf` → `TestASupersededReadingGapHandsTheCheckToItsSuccessor` fails (G2's close admitted);
   p. count the open-gap exemption from `carried_by` instead of `carrier_minted` → `TestAPlainMintALaterReadingNamesKeepsItsMint` fails;
   q. revert any one fuzz branch of §III.11 to `Detail == ""` → `TestTheFailedCheckDrivesKeyOnTheState` fails;
   r. let another lens's later backing reading replace a found-against check (`stands` keyed on any seat) → `TestTheLensCloseOfAReadingsGapIsRefusedWhileTheCitationStandsFoundAgainst` and `TestAMigratedFoundAgainstCheckReadiesBlueAlone` (subtest 6, another lens) fail.
   Re-armed by: the function or table each names.
6. migrate on real archives: `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/migrate/ -run 'M30|M32|M23|EveryArchivedRun|M19'` → pass, 0 refusals. Re-armed by: `migrate/**`, `record.proto`.
7. Release fuzz: PR CI only (sharded 4 + tally). Read the tally: new gates non-zero; max gaps per (class, sentence) = 1. NEVER run `FEOV_RELEASE_GATE=1` locally. Re-armed by: any record or flag change (CLAUDE.md requires a full pass before done).
8. **Universe smoke (real engine, cheap configuration).** `scripts/universe.sh build --dir ~/.claude/scratch/contradiction-is-a-gap/universe`; prove the build carries the change: `grep -rl claim_contradicts_source ~/.claude/scratch/contradiction-is-a-gap/universe/config/plugins/cache/special-circumstances/frank-exchange-of-views/` → at least the installed agents and the `feov-record` binary; and `scripts/universe.sh doctor` → schema epoch 49 (or the rebased number). Then
   `MODEL=haiku JUDGMENT_MODEL=haiku LANES=1 scripts/universe.sh --dir ~/.claude/scratch/contradiction-is-a-gap/universe run "Is the Great Wall of China visible from low Earth orbit with the naked eye?" --mint-budget 3 --k-max 2`
   (a topic whose popular sources misstate the claim, so an evidence lens is likely to read a refuting source). Observe with `universe.sh watch`; afterwards, on the run's record (read-only `file:` URI):
   `SELECT v.outcome, v.anchor, v.carried_by, m.class, m.class_material FROM verify v LEFT JOIN mint m ON m.gap_id = v.carried_by WHERE v.outcome IN ('refutes','absent');`
   Expected: every row has a `carried_by`, class in the outcome's admitted set, `class_material = always`; any second reading of one sentence shares a `carried_by`. If no negative reading occurred, the smoke is reported **NOT MEASURED on this path** (not green), and check 9 carries it. Archive the run to `run-archive/`.
9. **Driveable check on real data.** Migrate m30 into scratch (`~/.claude/scratch/contradiction-is-a-gap/m30/`), then on the migrated copy as `red-lens-evidence`: `feov-record verify --anchor C-4f4b8cfa --as refutes --class claim_contradicts_source …` → a gap lands (or names G-dc877265 if it is open with that class); a second lens's identical reading → names the same gap, no mint; `close` of that gap → refused while the reading stands; a third lens's `--as supports` reading → recorded, and the close still refused (a disagreement clears nothing); the reader's own `--as supports` re-read a sitting later → the close admitted; `show evidence` has no `unanswered_contradictions` key; `show board` lists the gap. Re-armed by: the reading act, the ledger, the close gate.
10. Auditor gate: `/prosthetic-conscience:plan-audit plans/contradiction-is-a-gap.md` — two consecutive PASS, or the human's approval with every fork ruled.

## Forks — all RULED (gblock, 2026-10-30)

Each ruling is recorded here; the alternatives it rejected are deleted from the design.

- **F1 — Class (RULED).** A NEW class `claim_contradicts_source` for `refutes`; `citation_figure_misattribution`
  narrowed to `absent`. The lens picks among the classes its outcome allows, declared once on
  `SourceOutcome` (`gap_classes`): `refutes` → `claim_contradicts_source` | `within_source_condition_misattribution`;
  `absent` → `citation_figure_misattribution` | `within_source_condition_misattribution` (§III.1). Cost
  carried as R3.
- **F2 — Judgment (RULED).** The reading act (`verify`/`corroborate`) takes the mint's judgment flags,
  required when the outcome is negative and refused otherwise (§III.2).
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
- **F5 — Corroboration `absent` (RULED).** Records the reading and mints no gap (§III.1, §III.10).
- **F6 — Budget (RULED).** A gap a reading mints is outside the mint budget (§III.5).
- **F7 — PR #1642 (RULED).** Closed as superseded; its four diagnostic tally fields — `final_blockers`,
  `final_parties`, `open_by_found_by` (re-keyed `reading | finding | none`), `mints_by_epoch` /
  `closes_by_epoch` — are carried into this PR (§III.11).
- **§III.9 — Archived carrier-less checks (RULED).** Cleared by the reader's own backing re-read, an edit or a retire. migrate does
  NOT fill `carried_by` from the `found_by` chain; the residue is stated in §III.9.
- **§III.14 — Delivery (RULED).** ONE PR, implemented one phase per fresh agent.
