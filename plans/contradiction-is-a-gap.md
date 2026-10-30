# A source reading against the report is a gap — the contradiction is deleted

Written against `origin/main` **`8236e3a17`** (EventSchema 48, after Slim-fast's #1601 and #1643).
Every file:line below is read at that commit. Implements gblock's ruling of 2026-10-30.

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
  `GapClass` gains one value (Fork F1); `SourceOutcome` gains a facet naming the classes each negative
  outcome admits (Fork F1); `Verify` gains `gap_id` (references `mint.gap_id`); `Mint` is unchanged.
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
  - "GapClass has no class for the cited source refutes / does not contain the claim": `citation_figure_misattribution`'s `(means)` already reads "the figure, statement or support is absent from or contradicted by its text". F1 decides whether to keep that breadth or split it.
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
| 1. One class or two | Two: a new `claim_contradicts_source` for `refutes`, the existing `citation_figure_misattribution` (narrowed) for `absent`; the lens picks among the classes its outcome admits | **F1** |
| 1b. Who names class, quote, problem | Quote = the reading's claim (the tool already reads it for `verify`; `corroborate`'s `--quote`). Class = the lens's `--class`, refused outside the outcome's admitted set. Problem, check, check-kind, severity, likelihood, impact, fix = the lens's flags on the reading act, required on a negative outcome and refused on any other | **F2** |
| 1c. Materiality | The class's `material_default` (both `always`), stamped at the write as for every gap | — |
| 2. Duplicate | The one-gap rule refuses the *mint*; the reading is written and names the standing gap in `Verify.gap_id` | **F3** |
| 3. Backing re-check | A citation found against is a **failed check** (third state beside unread/unsound). It holds the PASS until the citation holds, its sentence is edited, or it is retired; the gap it names cannot close while it stands (`requireFailedCheckAnswered`). A later backing reading clears the failed check; the minting lens then closes the gap `repaired` with its own close. No new answer kind | **F4** |
| Corroboration `absent` | Records the reading only; mints nothing | **F5** |
| Mint budget | A reading's gap takes no mint of the budget and does not count while open (the failed-check rule, restated for readings); the "answers its own raise" exemption is deleted | **F6** |
| 4. Delete | Everything in the census (§III.12) | — |
| 5. migrate | One translation, §III.9 | — |
| 6. Epoch | 49, re-derived at rebase | — |

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
- `Verify`: [NEW] `optional string gap_id = 20 [(sql) = { references: "mint.gap_id" }]` (field numbers
  1-19 are taken; 18 is `copy_substitution`) — the gap that carries this reading against its sentence:
  the one the reading minted, or the open gap of the reading's class on that sentence that refused the
  mint. Set by the tool, on a negative reading only. The rule "a negative reading names its gap" (except a
  corroboration's `absent`, F5) is enforced **at the write path** (`record.go`, the Verify case), not as a
  table CHECK: archived negative readings carry none (§III.9.2), and replay (`Migrating`) admits them as
  every other new write rule does. A field the writer can refuse, not a relation recovered from sentence
  position at close time (facts-are-fields).
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
  and then the Verify carrying `gap_id` — Mint first, because `gap_id` references `mint.gap_id` and the
  store runs with `foreign_keys(1)`. `record.go` gains the two-body placed append beside `AppendPlaced`;
  the Mint passes every mint validation (`record.go:955-980`: unclassified reason, subject, one-gap,
  budget, about-kind) in the same `writeView`, so the one-gap read and the insert share the write lock
  (`_txlock=immediate`, as `onegapperclass.go:32-35` already relies on).
- One-gap refusal **inside the reading act** (F3): `requireOneGapPerClassAndSubject` returns the held
  gap as a value rather than only an error; the reading path writes the Verify with `gap_id` = the held
  gap and no Mint. The plain `mint` verb keeps its refusal unchanged.
- Repeat: `requireNotARepeatFailedCheck` (`failedcheck.go`) extends to a negative reading of the same
  anchor by the same lens, sentence unchanged, while the gap its earlier reading names is open — refused,
  nothing recorded, naming the gap. A reading with a *different* outcome is not a repeat.
- Text that states the old model without the word [REWRITE]: the `--as` flag help (`verify.go:304`,
  "refutes and absent are findings") and `human()` (`verify.go:698-701`, "that is a finding, and this event
  is the evidence for it"); see §III.7's phrase sweep.
- Result: `verifyResult.Contradiction` / `readingContradiction` / `contradictionOf` / `words()`
  (`verify.go:615-688`) [DELETE]; [NEW] `Gap *readingGap {id, class, minted}` and its sentence:
  "This reading minted gap G (class C)" / "Gap G, open on this sentence with class C, carries this
  reading; no second gap is minted — add what it lacks as a finding about G". The `failedCheckWords`
  block stays for unread readings.
- Crash-retry: the act's `--key` keys both bodies (the Mint's `mint_key` derived from the reading's
  key), so a retry returns the stored pair; a retry differing in any gap flag is refused, as a keyed
  retry is today (`keyedretry_test.go`).

### III.3 The anchor ledger and the failed check — `internal/record/anchorchecks.go`, `failedcheck.go` [MODIFY] (F4)

- `stateFoundAgainst` stays as a ledger state (it is the citation's standing, and `CitationsFoundAgainst`,
  `counts.sources_refuted` and capture's assembled-report check read it); its comment stops saying "the
  contradiction's".
- `failedChecksOf` adds the found-against state: each standing negative reading of the sentence the report
  carries is a failed check whose owner is its reader and whose carrier (`answeredBy`) is the reading's
  `gap_id` — read from the field, not from an about-anchor mint. The owner owes nothing (the gap exists
  at the act), so dispatch readies blue alone for it (`dispatch.go:531-546`: the lens branch fires only
  while `answeredBy == ""`, which a live found-against check never is).
- `requireFailedCheckAnswered` then refuses closing that gap while the citation stands found against —
  bench included, successor (`--superseded-by`) admitted — with a found-against sentence beside the
  unread and unsound ones. The backing re-check is what clears the failed check; the gap's close stays
  the minting lens's ordinary `close` (or the bench's), so no new answer kind exists.
- **A found-against failed check never makes its reader owe a gap**, whether or not it names a carrier.
  `failedCheck` gains its state (`unread | unsound | found_against`), and three readers key on the state,
  not on `answeredBy == ""`: the lens arm of the readiness (`dispatch.go:531-538`) and the lens's work
  item (`sitting.go`) fire for `unread`/`unsound` only, and `answersOwnFailedCheck`
  (`failedcheck.go:109-119`) admits an about-anchor mint outside the budget for `unread`/`unsound` only.
  So a carrier-less found-against check — which only a migrated archived reading can produce (§III.9) —
  readies blue alone, and is cleared by the acts that clear every found-against check: a backing re-read
  by its reader (its reading replaced), an edit of the sentence, or a retire. No mint answers it, so the
  budget exemption cannot repeat. Test: `TestAMigratedFoundAgainstCheckReadiesBlueAlone` on migrated m32.
- `failedRemedy` / `failedCheckReadiness` text for found-against: "blue edits the sentence, or retires or
  replaces the citation".
- `holds()` comment (`anchorchecks.go:104-108`) and `VerifyReads` comment (`:278-281`) [MODIFY].
- Corroborations place no anchor when negative, so they reach the board through their gap only.

### III.4 The one-gap rule — `onegapperclass.go` [MODIFY]

- `answersRaisedContradictionAt` and its call (`:68-72`) [DELETE]; the exemption bullet (`:40-42`)
  [DELETE]. The rule now has no exemption but `--supersedes` and replay.
- [NEW] the held-gap return used by the reading act (III.2).

### III.5 Mint budget — `mintbudget.go` [MODIFY] (F6)

- `answersOwnRaiseAt`, its call (`:299-303`), `ownAnswersAt`'s contradiction half (`:367-395`), the
  `Answering` wording at `:504`, `:560` [DELETE/REWRITE]; the comment block `:33-41` [REWRITE].
- [NEW] a gap whose minting event's act is a reading (the Mint written with a Verify naming it) takes no
  mint and does not count while open, read from `Verify.gap_id` joined to the seat's own Mint rows.
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
- `sitting.go:369-425` (the lens owes a contradiction it read; the at-cap special case) [DELETE].
- `dispatch.go`: the readiness case (`:520-526`), `contradictionReadiness` (`:1285-1292`), Plan JSON
  `desc` text (`:18-36`) [DELETE/REWRITE].
- `retirement.go:20, :142` (a lens is not retired while it owes a contradiction) [REWRITE].
- `evidenceview.go:306-373`: `UnansweredContradictions`, `EvidenceContradictionJSON` [DELETE];
  `:564`, `:584` comments [REWRITE]. `narrowedview.go:20` [REWRITE].
- `convergence_refusal.go:77` [REWRITE].

### III.7 Agent-facing text [MODIFY / REGENERATE] (contract: prompts and constitutions)

- Help pages (`internal/cli/seat/help/`): `verify.md`, `corroborate.md` — the two contradiction paragraphs
  are replaced by one paragraph: a negative reading mints its gap in the same act, the flags it needs,
  what the one-gap rule does to a second reading, and what clears the citation's failed check;
  `finding.md`, `mint.md`, `dispatch.md` [REWRITE] the contradiction sentences; `finding.go:144-145` flag
  descriptions [REWRITE]; `verbs.go`, `flags/names.go` per census.
- `terms.json`: the `contradiction` entry [DELETE] (with its ordinary-English collision); `failed check`
  [REWRITE] to include "or a lens's reading found the cited source refutes the sentence or does not
  contain it — whose gap that reading minted at the act"; `lens retirement` [REWRITE] ("no contradiction
  it read" goes). Then `go -C scripts run ./vocabdoc` regenerates `docs/vocabulary.md`.
- agentgen sources [REWRITE]: `scripts/agentgen/src/fragments/lens-duties.md`, `fragments/lens-audit-and-mint.md`,
  `src/red-chair.md` (census), `src/skills/research-protocol/SKILL.md:17-19` ("it goes to the board as a
  finding — a PASS is refused until one is raised for it", found by the phrase sweep below), and
  `src/lead-judge.md` gains one sentence for the bench: a gap a lens's reading minted does not close
  while that reading stands, whatever the ruling (R4). `fragments/blue-research-protocol.md` KEEPs.
  Then `go -C scripts run ./agentgen` regenerates `agents/*.md` and both preloaded skills. Never
  hand-edit the generated text.
- **Phrase sweep (the old model without the word).** The census grep finds the word; it cannot find a
  sentence that teaches "a negative reading is a finding to raise". Run, and read every hit for that model:
  `grep -rIn -E -i '(refut|absent|negative reading|found against|contradicting).{0,120}(finding|rais)|(finding|rais).{0,120}(refut|absent)' scripts/agentgen/src plugins/frank-exchange-of-views/tools/internal/cli plugins/frank-exchange-of-views/skills plugins/frank-exchange-of-views/docs docs`
  Known hits at `8236e3a17`: the research-protocol SKILL source above, `verify.go:304`, `verify.go:699`.
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
2. **An archived negative reading** carries no `gap_id`: its gap was minted later (forward reference) or
   never. Not filled. Its found-against state is still a failed check on the migrated record (read from the
   ledger), with no named carrier.
3. **A mint whose `found_by` names a raising finding** is carried unchanged — the finding exists.
4. **A stored plan's `contradiction` blocker** is respelled `failed_check` with subject = the reading's
   anchor (`plankinds.go`, `planids.go:29-31, :91-96`), re-keyed as `TranslatePlanKinds` already does.
5. Migrate's comments in `entries.go:64-66` (the credit entry) and `outcomes.go:19` [REWRITE] to state
   archived bytes without the live concept word where they describe the live model.

**Residue (stated, accepted in this plan):**
- Archived readings name no gap, so the close gate cannot tie an archived answering gap to its reading;
  that matters only to a run resumed after migration. Such a carrier-less found-against check readies blue
  alone, puts nothing on its reader's work list and admits no budget-exempt mint (§III.3); it clears by a
  re-read, an edit or a retire. Where the archive held the sentence on a contradiction, the migrated record
  holds it on a failed check.
- m32: each of its four negative readings of `C-a36dce3f` that still stands at the end is its own
  found-against failed check (up to four, one per reader) and its four raising findings become four
  findings about that anchor. m32 ended UNVERIFIED, so no verdict moves.
- Not taken: filling archived `gap_id` from the found_by chain — the answering gap is minted after the
  reading, a forward reference the FK and the write order refuse.
- Epoch-24/25/26 negative readings (m19, m23, 07d, wave-a) never had a raise; they come forward as
  readings on the ledger exactly as before.

Tests [NEW], in `migrate/`: `TestTheMigratedM30RaiseIsAFindingAboutItsCitation` (F-8410bec7 →
`about anchor C-4f4b8cfa`; G-dc877265 open/closed as archived; 0 refusals) and
`TestTheMigratedM32HoldsAFailedCheckPerStandingReading` (and `TestAMigratedFoundAgainstCheckReadiesBlueAlone`: dispatch readies no m32 lens for it, no lens work item, an about-anchor mint is counted); `TestM23CorroborationReadingComesForward`
(the unraised corroboration and citation readings, 0 refusals). `reopenedm19_test.go` and
`TestEveryArchivedRunMigrates` keep passing.

### III.10 Tests [MODIFY / DELETE / NEW]

- DELETE the four files that exist for the concept (§III.12): `integration/surface/contradictionanswer_test.go`,
  `record/contradictionquadratic_test.go`, `record/contradictionref_test.go`, `record/answerbudget_test.go`.
  `cli/m23contradiction_test.go` keeps its quote-refusal half and is renamed for what it then tests.
- REWRITE the rest per census; each rewritten assertion tests the new behaviour, not the absence of the old.
- NEW (each with its delete-the-row partner in §V.5):
  - `TestARefutingReadingMintsItsGapInTheSameAct` (verify and corroborate; class from the admitted set;
    `class_material` = always; `Verify.gap_id` = the gap).
  - `TestASecondRefutingReadingNamesTheStandingGap` (another lens, same sentence, same class: reading
    written, no mint, `gap.minted=false`; a different admitted class mints).
  - `TestANegativeReadingWithoutGapFlagsIsRefused` / `TestABackingReadingWithGapFlagsIsRefused`.
  - `TestTheReadingsGapDoesNotCloseWhileTheCitationStandsFoundAgainst` (close refused; edit, retire, or
    a backing re-read by the reader → close admitted).
  - `TestAReadingsGapTakesNoMint` (lens at cap records a refuting reading; its gap lands).
  - `TestACorroborationAbsentMintsNothing` (F5).
  - `TestEveryNegativeOutcomeNamesItsClasses` (the `gap_classes` facet is non-empty exactly on
    `refutes` and `absent`, and every named class exists).

### III.11 Release fuzz harness — `releasegate/fuzz/` [MODIFY]; PR #1642 [SUPERSEDED]

- Delete the contradiction drives and census fields: `answerContradictions` (`fuzz_test.go:965-966`), the
  `ownAnswers`/`answerBudgetErr` fields (`:494-497`), the sweep gates at `sweeptally_test.go:860, :894`
  and every other site the census lists.
- The lens's negative reading draws the gap flags (class from the outcome's admitted set, grades, check)
  in the same act. New gates (false-green guards, zero fails the sweep): negative readings that minted;
  negative readings refused a mint by the one-gap rule (the duplicate path is driven on purpose by
  re-reading a sentence already holding a reading's gap); found-against closes refused. New tally:
  max open gaps per (class, sentence) — must be ≤ 1.
- PR #1642 (on hold): closed as superseded once this plan's PR is opened; its `final_blockers`,
  `final_parties`, `mints_by_epoch`, `closes_by_epoch` tally fields are re-proposed here only if the
  human wants them (they are diagnostics independent of this concept), with `open_by_found_by`'s
  `contradiction` bucket replaced by `reading`.
- Release gate: PR CI only. Never `FEOV_RELEASE_GATE=1` locally.

### III.12 Consumer census

Searches, run at `8236e3a17` (raw hits kept at `~/.claude/scratch/contra-census/hits.txt`, context in
`ctx.txt` beside it):

```
grep -rIn -i contradict --exclude-dir=.git --exclude-dir=run-archive .
grep -rIn -E 'ABOUT_KIND_CONTRADICTION|unanswered_contradictions|ContradictionOfReading|raisedBy|answeredBy|raised_by|opened_by|sources_refuted' --exclude-dir=.git .
```

Result: **201 files** (198 carry the word, 3 only the identifiers). Each hit was read for meaning.
**DELETE 4 · REWRITE 60 · REGENERATE 16 · KEEP 121** (38 of the KEEP are historical). Paths are relative
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
- KEEP (other sense): record evidenceview_test, foundagainstm30_test (`sources_refuted`), cast_test,
  enums_test, passagreement_test, registeronce_test, roster_test, spotcheck_test,
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

**Docs** — REGENERATE: `docs/vocabulary.md` (vocabdoc). REWRITE: `docs/citations-and-bibliography.md` l.185.
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
| Anchor ledger `anchorStanding.state()` (`anchorchecks.go:152-178`, from #1601 / 90f3fa008) and `stateFoundAgainst` | a citation whose standing check refutes or is absent; its PASS hold was "the contradiction's" | Unchanged predicate: found-against is still the **citation's standing** (a standing negative reading of the sentence the report carries), NOT "has an open gap of the new class". The gap is a separate fact linked by `Verify.gap_id`; the state is what makes the gap unclosable (§III.3). Keying the state on the gap would let a gap closed into a successor, or an archived reading with no gap, read as holding | MODIFY (comment only; `contradicts()` → `gap_classes` facet) |
| `failed_check` PASS blocker (`passblockers.go:397`, `failedcheck.go`) | unread citations and unsound proofs | gains the found-against state (§III.3): holds the PASS until the citation holds, is edited or retired; readies blue alone; carrier = `Verify.gap_id` | MODIFY |
| `unchecked` PASS blocker | a citation or proof with no check of its current sentence | unchanged: an edit of a found-against sentence makes the citation unchecked, owed a fresh check by its owning lens; that check either holds (failed check cleared, the gap may close) or is negative again (names the standing gap) | KEEP |
| Evidence lens's gap duties (`fragments/lens-duties.md`, `lens-audit-and-mint.md`, `verify.md`) | read every citation; a negative reading opens a contradiction it must raise by finding, then answer by a gap minted from it, outside its budget | reads every citation; a negative reading carries the gap's judgment (class among the outcome's admitted set, problem, check, grades) and mints it in the same act, outside its budget; a second reading of a sentence holding that gap names it; the lens closes the gap with its ordinary close once the citation holds, is edited or retired. Unread readings keep the owed-gap-about-the-anchor duty | REWRITE (agentgen) |

**Carried by the concept, not found by the word:** `record.proto.sha256`, `requirements.json` (epoch),
`schema_gen.go`, `migrate/plankinds.go`, `failedcheck.go` (§III.3), `cli/lens/mint.go` (flag helper).

### III.13 Directory tree of touched paths

```
plugins/frank-exchange-of-views/
  requirements.json                         [MODIFY] eventSchema 49
  tools/internal/record/
    recordpb/record.proto                   [MODIFY] AboutKind, GapClass, SourceOutcome facet, Verify.gap_id
    recordpb/record.pb.go, schema_gen.go    [REGENERATE] protogen, schemagen
    recordsql/testdata/schema.sql           [REGENERATE]
    citationid.go                           [MODIFY] contradiction fold deleted
    anchorchecks.go, failedcheck.go         [MODIFY] found-against is a failed check
    onegapperclass.go, mintbudget.go        [MODIFY]
    passblockers.go, sitting.go, dispatch.go, retirement.go, evidenceview.go,
    narrowedview.go, about.go, record.go, enums.go, convergence_refusal.go   [MODIFY]
    migrate/replay.go, plankinds.go, planids.go, entries.go, outcomes.go     [MODIFY]
    migrate/*_test.go                       [NEW] m30, m32, m23
  tools/internal/cli/lens/verify.go, mint.go, finding.go   [MODIFY]
  tools/internal/cli/seat/help/*.md         [MODIFY]
  tools/internal/terms/terms.json           [MODIFY]
  tools/releasegate/fuzz/*                  [MODIFY]
  skills/research-protocol/scripts/debate.js  [REGENERATE + MODIFY]
  agents/*.md, skills/*/SKILL.md            [REGENERATE] agentgen
  docs/vocabulary.md                        [REGENERATE] vocabdoc
scripts/agentgen/src/…                      [MODIFY]
```

### III.14 Independently shippable parts

None that is a complete concept on its own; stated so the bundle is a choice, not a default:

- *Delete the contradiction first, add the reading's gap later* — truncated: between the two, a refuting
  reading holds no PASS and reaches no seat (the half-state the 10-06 ruling closed).
- *Add the class and the facet first* — a class no act routes to is a dead value; it would also take an
  epoch of its own for nothing a reader sees.
- *The fuzz change alone* (PR #1642) — patches the harness, not the engine (§I value case).
- The agent-text rewrite cannot precede the behaviour (the help would describe flags that do not exist).

So: one PR, one epoch. Its cost is dominated by the census rewrite (60 REWRITE incl. 22 tests, 4 DELETE, 16 REGENERATE), not
machinery; the machinery is one facet, one field, one append path, one ledger state widened.

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation (step) |
|---|---|---|---|
| R1 | A carrier keeps speaking the contradiction (prompt, help, golden, fuzz) — half-state that reads done | H × H × L | §III.12 census re-run as §V.4, identifiers grep must be empty; agentgen/enginegen/vocabdoc `-check` gates (§V.2) |
| R2 | Lenses stop recording negative readings because the act now costs a full mint's flags | M × H × M | Flags are the same set `mint` already asks; refusal names each; smoke §V.8 checks negative readings still occur; the reading's gap takes no budget (F6) so cap is not a reason to read `supports` |
| R3 | One-gap rule + lens-picked class lets two lenses split one defect across two classes on one sentence | M × L × L | Ordinary behaviour for any gap; the admitted sets are two classes each; sweep tally reports per-sentence gaps by class |
| R4 | A bench `not_a_defect` ruling on a misread source cannot close the gap while the reading stands | M × M × M | Same as an unsound proof today (failed check is not a dispute); the reader re-reads (its own reading replaced) or blue edits. Stated in the bench's constitution. Fork F4 offers the alternative |
| R5 | migrate mistranslates m30/m32 raises | L × H × L | Tests on the real archives (§III.9); refusal on corroboration raises is loud |
| R6 | Epoch collision with Slim-fast's epoch PR (50) or #1544 Part 1a | M × M × L | Re-derive at rebase; `schemaepoch_test.go` and setup's epoch refusal catch a stale number |
| R7 | Goldens conflict semantically with a help-changing PR merged meanwhile | H × M × L | Rebase, regenerate at `-count=1`, read each changed line, re-run CI before merge (CLAUDE.md) |
| R8 | The two-event act leaves a Verify without its Mint on crash | L × H × L | One transaction; keyed retry returns the stored pair (`keyedretry_test.go` extended) |
| R9 | Release fuzz loses coverage of the duplicate path | M × M × L | New zero-gates (§III.11) fail the sweep if the one-gap path in the reading act is never driven |

## V. Verification Plan

All commands from the worktree root, after `unset $(env | grep -o '^FEOV_[A-Z_]*')`.

1. `go -C scripts run ./protogen && go -C scripts run ./schemagen && go -C scripts run ./agentgen && go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen && go -C scripts run ./vocabdoc` → no diff on a second run. Re-armed by: `record.proto`, `requirements.json`, `scripts/agentgen/src/**`, `terms.json`, the roster/enums.
2. `go -C scripts run ./check` → `0 failed` (read the count, not the exit code). Re-armed by: any file in the diff. Includes the `-check` gates of every generator, `archaeology`, `golden`, `pluginparity`, vocabulary prose.
3. `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/... ./internal/cli/... ./integration/... ./releasegate/...` (release gate unset: the fuzz sweep does not run) → 0 FAIL. Goldens: `UPDATE_GOLDENS=1 go test -count=1 <pkg>` then read every changed line. Re-armed by: any Go or help change.
4. Census re-run (must print only KEEP lines of §III.12):
   `grep -rIn -i contradict --exclude-dir=.git --exclude-dir=run-archive plugins scripts docs | grep -v -E 'cross_section_contradiction|claim_contradicts_(own_record|source)'`
   and identifiers, which must print nothing outside `migrate/` archived-word tables:
   `grep -rIn -E 'ABOUT_KIND_CONTRADICTION|BlockerContradiction|unanswered_contradictions|contradictionsOf|ContradictionOfReading|requireOpenContradiction|answersRaisedContradiction|answersOwnRaise|gap_may_name|raised_by' plugins scripts`.
   Re-armed by: any agent-facing or Go edit.
5. Delete-the-row checks (each must turn the suite red; revert after):
   a. delete `(gap_classes)` from `SOURCE_OUTCOME_REFUTES` → `TestARefutingReadingMintsItsGapInTheSameAct` and `TestEveryNegativeOutcomeNamesItsClasses` fail;
   b. delete the Mint body from the reading's append (keep the Verify) → the same test fails;
   c. delete the held-gap branch in the reading path → `TestASecondRefutingReadingNamesTheStandingGap` fails;
   d. delete the found-against branch of `failedChecksOf` → `TestTheReadingsGapDoesNotCloseWhileTheCitationStandsFoundAgainst` fails;
   e. delete the reading exemption in `requireMintWithinBudget` → `TestAReadingsGapTakesNoMint` fails;
   f. delete migrate's raise translation → `TestTheMigratedM30RaiseIsAFindingAboutItsCitation` fails (refusal);
   g. delete the call site of the new flag-registration helper in `verify.go` → help/flag tests fail;
   h. delete the `Verify.gap_id` CHECK → a test writing a negative reading with no gap fails.
   Re-armed by: the function or table each names.
6. migrate on real archives: `go -C plugins/frank-exchange-of-views/tools test -count=1 ./internal/record/migrate/ -run 'M30|M32|M23|EveryArchivedRun|M19'` → pass, 0 refusals. Re-armed by: `migrate/**`, `record.proto`.
7. Release fuzz: PR CI only (sharded 4 + tally). Read the tally: new gates non-zero; max gaps per (class, sentence) = 1. NEVER run `FEOV_RELEASE_GATE=1` locally. Re-armed by: any record or flag change (CLAUDE.md requires a full pass before done).
8. **Universe smoke (real engine, cheap configuration).** `scripts/universe.sh build --dir ~/.claude/scratch/contradiction-is-a-gap/universe`; prove the build carries the change: `grep -rl claim_contradicts_source ~/.claude/scratch/contradiction-is-a-gap/universe/config/plugins/cache/special-circumstances/frank-exchange-of-views/` → at least the installed agents and the `feov-record` binary; and `scripts/universe.sh doctor` → schema epoch 49 (or the rebased number). Then
   `MODEL=haiku JUDGMENT_MODEL=haiku LANES=1 scripts/universe.sh --dir ~/.claude/scratch/contradiction-is-a-gap/universe run "Is the Great Wall of China visible from low Earth orbit with the naked eye?" --mint-budget 3 --k-max 2`
   (a topic whose popular sources misstate the claim, so an evidence lens is likely to read a refuting source). Observe with `universe.sh watch`; afterwards, on the run's record (read-only `file:` URI):
   `SELECT v.outcome, v.anchor, v.gap_id, m.class, m.class_material FROM verify v LEFT JOIN mint m ON m.gap_id = v.gap_id WHERE v.outcome IN ('refutes','absent');`
   Expected: every row has a `gap_id`, class in the outcome's admitted set, `class_material = always`; any second reading of one sentence shares a `gap_id`. If no negative reading occurred, the smoke is reported **NOT MEASURED on this path** (not green), and check 9 carries it. Archive the run to `run-archive/`.
9. **Driveable check on real data.** Migrate m30 into scratch (`~/.claude/scratch/contradiction-is-a-gap/m30/`), then on the migrated copy as `red-lens-evidence`: `feov-record verify --anchor C-4f4b8cfa --as refutes --class claim_contradicts_source …` → a gap lands (or names G-dc877265 if it is open with that class); a second lens's identical reading → names the same gap, no mint; `close` of that gap → refused while the reading stands; `show evidence` has no `unanswered_contradictions` key; `show board` lists the gap. Re-armed by: the reading act, the ledger, the close gate.
10. Auditor gate: `/prosthetic-conscience:plan-audit plans/contradiction-is-a-gap.md` — two consecutive PASS, or the human's approval with every fork ruled.

## Forks (for gblock — each with a recommendation)

**F1 — Which class a negative reading's gap takes.**
- *A. One class:* widen `citation_figure_misattribution` (it already says "absent from or contradicted by
  its text") to every negative reading. No enum change. But it says "a cited source", which a refuting
  corroboration is not, and it merges two repairs (fix the sentence vs fix the citation) into one class,
  so the one-gap rule folds a refutes and an absent on one sentence into one gap.
- *B. Two classes, fixed by outcome:* `refutes` → new `claim_contradicts_source`; `absent` →
  `citation_figure_misattribution` narrowed. Separates the repairs; the tool picks, the lens cannot
  misfile. But m30 and m32's lenses both filed their refutation as `within_source_condition_misattribution`
  — the source states the result under a condition the report drops — which B could not record.
- *C. (Recommended) Two classes plus a declared choice:* B's classes, and the lens picks among the
  classes its outcome admits (`refutes`: `claim_contradicts_source` | `within_source_condition_misattribution`;
  `absent`: `citation_figure_misattribution` | `within_source_condition_misattribution`), declared once on
  `SourceOutcome`. Keeps the archived choice expressible; cost: two lenses may file one defect under two
  classes on one sentence (R3).

**F2 — Where the gap's judgment comes from.**
- *A. (Recommended)* The reading act takes the mint's flags (problem, check, check-kind, grades, fix), required on a negative outcome. One act; nothing owed; no grade invented.
- *B.* The reading records and the lens owes a `mint` naming it (the failed check's owed-gap shape, readied until minted). Two acts, and the debt between them is the placeholder the ruling objects to.
- *C.* The tool mints with default grades. Rejected: the tool never makes it up — a fabricated grade reads exactly like a judged one.

**F3 — What the reading act does when the one-gap rule refuses its mint.**
- *A. (Recommended)* Keep the reading, refuse the mint, write `Verify.gap_id` = the standing gap, print it; the lens adds a finding about that gap only if it has more to say. The evidence stays on the ledger, so the citation is checked (no readiness loop).
- *B.* Refuse the whole act (nothing recorded), telling the lens to file a finding about the gap. Loses the evidence, and for a second citation in the same sentence leaves it unchecked: the evidence lens is readied again, reads again, is refused again — a loop.
- *C.* Keep the reading with no link; gate the gap's close by re-deriving the sentence's anchors at close time. A relation recovered from text position instead of a field the writer can refuse.

**F4 — What holds the PASS for a citation found against, and how a backing re-check relates to the gap.**
- *A. (Recommended)* Found-against becomes the third failed-check state: it holds the PASS until the citation holds, is edited or retired, no bench ruling clears it, and the gap its reading names cannot close while it stands. The backing re-check clears the failed check; the lens closes the gap with its ordinary close. Consistent with `unsound` (also a judgement) and closes the successor hole. Cost: a bench cannot rule a refutation wrong while the reading stands (R4).
- *B.* Only the gap holds the PASS; a close gate refuses `repaired` while a reading naming the gap stands, and the bench may close it. Simpler, but a gap closed into a successor drops the reading's hold (the successor is not named by the reading), and archived readings hold nothing.

**F5 — A corroboration that reads `absent`.**
- *A. (Recommended)* Recorded as a reading, no gap. A source red chose that is silent on a claim is not a defect of the report; if the claim is unsupported, that is `unsourced_claim`, minted by the lens's judgement like any gap.
- *B.* Mint as for a citation. There is no citation to misattribute, so no admitted class fits; it would need `unsourced_claim` (`by_grade`) or a third class, and today's behaviour (a PASS hold on red's own unlucky search) continues.

**F6 — Mint budget for a reading's gap.**
- *A. (Recommended)* Outside the budget, not counted while open (as a failed check's gap). The one-gap rule bounds it per sentence; counted, a lens at its cap could only record `supports` or nothing — the incentive PR #1642's stock built into the fuzz.
- *B.* Counted; at the cap the reading is refused. Bounds mints harder, but the citation then stays unchecked and the evidence lens is readied to re-read it every sitting.

**F7 — PR #1642's diagnostic tally fields** (`final_blockers`, `final_parties`, `mints_by_epoch`, `closes_by_epoch`): *(Recommended)* close #1642 as superseded and carry those four fields into this PR's fuzz change, since they diagnose CEILING runs independently of the contradiction; or drop them.
