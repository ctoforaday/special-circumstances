# The gap class is a record enum (#1134)

Written against `be6c110d`. Implements gblock's ruling of 2026-10-18 on #1134 (the issue's last
comment); absorbs #1131 (definitions) and #1132 (truncated registry sample). Replaces the
superseded draft `plans/class-registry-ships.md`, which planned to ship the JSON registry inside
the binary.

## I. Summary & Goals

**Objective.** A gap's class is today a free slug validated against a JSON file
(`feov-memory/class-registry.json`) that `setup` copies from the *caller's* working directory. The
plugin does not ship that file, so in every consuming project `setup` prints `class registry: NOT
STAGED` and every `mint` of the run is refused — reproduced at `be6c110d` from an empty project:

```
$ feov-record --seat-id operator setup …/research/probe --model haiku --judgment-model haiku
  class registry: NOT STAGED — no class-registry.json in …/proj/feov-memory — nothing constrains `--class`, so every mint this run will be REFUSED rather than waved through
$ feov-record --seat-id red-lens-logic mint --class nope …
feov-record: mint: no gap-class registry is staged for this run — every --class would be accepted, so a mint is refused rather than waved through. run-setup stages it from feov-memory/class-registry.json; …
```

The class becomes an ordinary record enum, `GapClass`, declared in `record.proto` like every other
closed set: each value carries its definition (`(means)`), its materiality, its nearest neighbour and
the question that tells them apart, and `mint --help` prints all four. A catch-all value takes a gap
no value fits. Coining (`class new`), the class list (`class list`), the registry file, its per-run
staging, its generator and its harvest are deleted. The event-schema epoch moves 31 → 32 and
`migrate` carries every archived run onto the enum.

**Rulings this plan implements (settled — #1134 last comment).**

1. Ordinary record enum; each value's definition and materiality in the enum; `--help` prints them.
2. Catch-all `unclassified`: `by_grade`, requires `--reason` (why no class fits); its rate in telemetry.
3. Delete: registry JSON + per-run staging, `classgen` + both `_gen.go` files + its gate and CI step,
   `ValidateClassRegistry` / `StageClassRegistry` / `loadRegistry`, `lens class new` + the `class_new`
   event, `class list`, `HarvestClasses` + the class proposals, `universe.sh`'s feov-memory staging.
4. Membership: a class stays only if a definition and a distinguisher from its nearest neighbour can be
   written; synonyms merge; use counts inform, not decide (20 of 26 archived runs are "is 91 prime").
5. Epoch bump; `migrate` maps coined, merged and dropped classes onto the enum, tested on an archived run.

**Success criteria.**

1. `setup` from a directory with no `feov-memory/` prints no class-registry line and a lens `mint`
   in that run succeeds (§V.4, built binary).
2. `mint --class unclassified` without `--reason` is refused at the write; with `--reason` it records,
   and `inquest telemetry` reports `unclassified_rate` for that epoch (§V.2 tests, §V.4, §V.6).
3. `mint --help` (rendered, not grepped) lists every `GapClass` value with its materiality,
   definition, neighbour and distinguisher; the seven `red-lens-*` constitutions carry the same block
   via `agentgen` (§V.3).
4. All 27 archived run tarballs (26 with mints) migrate with zero refusals; a named test on
   `2026-10-18_is-91-prime-m28` and `2026-08-23_sleeper-service-plan` asserts every mint lands on a
   `GapClass` value, every recorded `class_material` is unchanged, no `class_new` event and no
   `records/class-registry.json` survives, and the manifest lists each class translation (§V.5).
5. A universe smoke with no `feov-memory` staged mints ≥1 gap, and one mint uses `unclassified`
   (§V.7).
6. `git grep -n -E 'class-registry\.json|classgen|ShippedClasses|ShippedMaterialDefaults|StageClassRegistry|ValidateClassRegistry|loadRegistry|StageForRun|ClassRoster|HarvestClasses|ClassNew|class_new|class new|class list'`
   over `plugins/ scripts/ .github/ docs/` returns nothing (§V.8).
7. `go -C scripts run ./check` green at `-count=1` with the fail count read; release-gate fuzz green.

**Value case.** Against the superseded design (embed the JSON, keep coining), the enum: removes the
class of defect rather than relocating it — a vocabulary the binary owns cannot be missing from a
consumer's tree; puts definitions where a lens already reads closed sets (39 of 41 registry rows had
none, and three runs logged that `class list` shows none); replaces a parallel validator
(`loadRegistryIn`), a generator, two generated files, a CI step, a per-run channel, a verb group, an
event type and a capture harvest with one proto enum the existing machinery (`EnumFields`, enum
tables, `enumhelp`, `migrate`'s value translation) already handles. The measured cost of what is
removed: 33 coinings across 13 runs, mostly single-use near-synonyms (inquiry-* in 5 runs, citation-*
in 4); none of 14 harvested proposals adopted. What the change gives up — a run extending its own
vocabulary — is replaced by `unclassified` + its telemetry rate, which says when a release should add
a value.

**Split.** No part ships alone with value: the enum without the deletions leaves two vocabularies;
the deletions without the enum leave none; migrate is required by the epoch bump. The one
independently reviewable part is the **membership table (§III.2)**, and it is a human review, not a
PR: it is the fork that must be ruled before implementation, because every definition is a seat-facing
sentence and every merge is a migrate translation. One PR, as ruled.

## II. Technical Context

- Go 1.27, `GOTOOLCHAIN=local` in CI. FEOV tools module `plugins/frank-exchange-of-views/tools`; dev
  tooling `scripts/`. Proto source `internal/record/recordpb/record.proto`, generated by
  `go -C scripts run ./protogen` (`record.pb.go`, `record.proto.sha256`).
- **How a closed set works today** (the pattern followed exactly):
  - The enum lives in `record.proto`; each value carries `(means)` (`EnumValueOptions` 50001), and
    optionally FACETS (`closes`, `mass`, `seat_may_file`, `gap_may_name`) — "what the record DOES
    about it" — interpreted only by `recordpb/facets.go` and refused at schema build when declared on
    some values and not others (`recordsql/schema.go` `enumTable`).
  - Spelling: `recordpb.Spelling` strips the type prefix and lowercases — underscores
    (`low_medium`, `defect_accepted`). `SameWord` folds case and separators, for near-miss refusals and
    for `migrate`'s archive reads.
  - Every enum a column uses becomes a vocabulary table `enum_<name>` (`value`, `means`, one column
    per declared facet) and the column a `REFERENCES` it; the database refuses an unknown word.
  - `record.EnumFields[<event>]` declares which payload key / flag carries the set; `validate`
    enforces it at the one write path, and `enumhelp` renders `--help`'s "Enumerated values" block
    and the refusal from the same `[]EnumValue` (`evsOf`, read off the descriptor).
- **The gap class today.** `Mint.class` is `optional string` (field 3, `required`). Materiality is
  `Mint.class_material` (`ClassMaterial`, field 25, required), stamped by `stampClassMaterial` from
  the staged registry row or the run's `class_new` row; a seat may not supply it. `knownClasses` =
  staged registry ∪ `class_new` slugs. `Mint` also carries `class_new` (4, bool, set by the mint verb
  from `ClassCoinedInRun`), `definition` (5), `neighbor` (6), `distinguisher` (7) — the last three
  written only by the pre-`class new` era and read by nothing but migrate and capture.
- **Materiality is the only behaviour that depends on class** (ruling's measurement; confirmed: no
  Go reader branches on a slug — `git grep` for each of the 41 slugs in non-test Go returns comments
  only, §III.14 C12). Readers of the class word are display and telemetry: `view.go` (`by_class`,
  `class_repeat_rate`), `report/assemble.go:863`, `graph/graph.go:223`, `record/viewjson.go`,
  `capture.go`.
- **Event-schema epoch** `record.EventSchema` = 31, generated from
  `plugins/frank-exchange-of-views/requirements.json` `eventSchema` by `scripts/schemagen`. Changing
  `Mint.class`'s type, deleting `EVENT_TYPE_CLASS_NEW`/`ClassNew`, and deleting Mint fields 4–7 change
  the event shape: **epoch 32**.
- **migrate** (`internal/record/migrate`): re-drives an archived record's events through the live
  write path. Old words translate through `Entries()` (a `Registry` of event words → `Translate`, or
  `Loss`); an archived enum word resolves through `valueRenames[<enum full name>]` then the schema's
  own spelling with `SameWord` (`enumValue`/`enumNumberOf`). A translation that produces no event is
  precedent (`revision`). Fills of a value the source predates are `StatedFill`s on the manifest. The
  staged registry is today a copied run channel plus `translateStagedRegistry`.
- **Archived data** (measured, `~/.claude/scratch/relocations/classuse/scripts/{m,cm,coined}.py` over
  the 27 run tarballs): 210 mints; 26 of 41 registry classes used; 38 mints under 30 coined slugs (33 coined
  in all). 16 runs (2026-08-22 … 2026-09-11) record **no** `class_material` on any of their 123 mints
  and their staged registries carry no `material_default`; the 10 runs with mints from 2026-10-09 record it on
  every mint.
- **Import graph.** The superseded design fought a `record`/`recordtest` cycle over a shipped Go
  table. An enum in `recordpb` is imported by both already; the constraint disappears.
- **How a lens sees a closed set today** (rendered from the built binary at `be6c110d`, scratch
  `~/.claude/scratch/relocations/gapenum/help-*.txt`): `mint --help` ends with an "Enumerated values"
  block — `--about-kind` and `--check-kind`, each value with its `(means)`; `--class` is a bare
  `slug` flag pointing at "the class new verb". The lens constitutions inline every help page of the
  lens surface (`scripts/agentgen` renders `manual`), so `agents/red-lens-*.md` carry `class list`,
  `class new` and `mint` pages verbatim (lines ~188–240, ~413).
- Constraints: no network, no secrets. Nothing consumer-configurable.

## III. Proposed Changes

### III.1 Design decisions

| # | Question | Options weighed | Decision and reason |
|---|---|---|---|
| D1 | Proto enum or string? | (a) keep `string` + a Go table + a Go validator; (b) **proto enum `GapClass`**; (c) a separate SQL vocabulary table outside the proto | **(b).** It is the pattern every other closed set uses, so the vocabulary table, the FK refusal, `EnumFields` validation, `--help` and `migrate`'s word translation come with it; (a) rebuilds `loadRegistryIn` in Go — the second validator the ruling deletes; (c) is a hand-kept second schema. The type change also makes the Go compiler the census of every `Mint.GetClass()` reader. |
| D2 | Spelling | (a) **underscores** (`report_voice`), the schema's spelling; (b) hyphen join like the retired grade spelling | **(a).** `Spelling` derives underscores and the grade hyphen exception no longer exists; a seat that types `report-voice` gets the existing `SameWord` near-miss refusal, and migrate reads hyphenated archives through the same fold with no authored row. Disclosed: prose naming `report-voice` changes (§III.12). |
| D3 | Where materiality lives | (a) string facet holding a word; (b) **an enum-typed value option `(material_default)` of type `ClassMaterial`**; (c) a Go map | **(b).** `protoc` refuses a misspelt word at generation; the name is the existing concept's (`material_default` on registry rows and `ClassNew`), not a rival term. It is a FACET (the stamp reads it): `facets.go` gains a third facet kind (enum-valued) and `enumTable` emits `"material_default" TEXT NOT NULL REFERENCES "enum_class_material"("value")`. Partly-annotated is refused as for every facet, so a value cannot be added without answering it. |
| D4 | Neighbour and distinguisher | (a) facets (table columns); (b) **value options read by help and a test, not facets**; (c) folded into `(means)` prose | **(b).** The catch-all has no neighbour, and a facet cannot be declared on all-but-one value; nothing in the machinery switches on them (precedent: `ruled_by` is an option read directly, not a facet). `TestEveryGapClassNamesItsNeighbour` requires both on every value but `unclassified` and requires `neighbor` to be another value's word. (c) would compose a checkable reference into prose. Names are the existing ones (`--neighbor`, `--distinguisher`). |
| D5 | `Mint.class_material` | (a) **keep stored, stamped from the value's `material_default`**; (b) delete and derive by join on `enum_gap_class` | **(a).** Merges and drops move archived gaps to values with a different default (e.g. `audited-artifact-sibling-halo` `by_grade` → `verification_scope_blindspot` `always`; anything → `unclassified` `by_grade`); CLAUDE.md requires the migrated record to hold the equivalent state, and a stored stamp is the gate's materiality at the act. The stamp is a different fact from the class's current default, so this is not two writers of one fact. |
| D6 | Catch-all name | `unclassified` (ruled) | Grep found `Unclassified` in `diagnostics/survey.go` (`Traversal.Unclassified`, JSON `unclassified`: invocations the traversal could not place in a group) and `migrate/manifest.go` (`unclassified_tables`: old tables no rule could place). Same sense — "could not be placed" — in namespaces no reader joins with a gap's class. **Kept**; the telemetry key is `unclassified_rate`, which collides with nothing (0 hits). |
| D7 | "`unclassified` requires `--reason`" | (a) Go `if` in validate; (b) **a message `(check)` on `Mint`** | **(b)**, precedent `DocketRuling`'s remand-direction check: `"class" <> 'unclassified' OR coalesce(trim("mint_reason", …), '') <> ''`, `why:` "a gap no class fits says why in --reason: the nearest class and what it does not cover — the rate of these is how a release learns which class is missing". The database refuses it, so no route writes around it. `--reason` (`mint_reason`) carries it, as ruled; its help line gains the duty. Alternative noted, not proposed: a refusable `--nearest <class>` enum flag would make "which value was nearly right" a field rather than prose. |
| D8 | Telemetry | by_class map vs repeated; where the rate goes | `NewMint.by_class` was a map "because the registry is extended at runtime by `class-new`" (its own comment) — the premise is gone, so it becomes `repeated ClassTally by_class` in schema order, the `by_severity` precedent (deterministic by construction). New `optional double unclassified_rate` (absent when nothing was minted, as `class_repeat_rate`). `class_repeat_rate` counts an `unclassified` mint in the denominator and never as a repeat: "no class fits" recurring is not one kind recurring, and counting it would raise the bench's stopping signal on noise. Per-epoch only; a cross-run rate belongs to #1454's archive-derived scorecards (F4). |
| D9 | How a lens sees the definitions | `mint --help`; the constitution; the refusal | `mint --help`'s "Enumerated values" block gains `--class`, one entry per value: `word  [material_default] definition` and, indented, `nearest: <word> — <distinguisher>`. `agentgen` inlines it into every `red-lens-*.md` (the lens's manual), so a lens reads it at sitting start without a call. A wrong word is refused with the same block (as `--check-kind banana` is today). Cost stated in F3. |

### III.2 Membership (FOR HUMAN REVIEW — fork F1)

Sources: the 41 registry rows (2 with definitions: `structure-noncompliance`, `self-defeating-mitigation`),
`feov-memory/class-registry-seed.md` (original intent, tie-break triplets), every archived mint's
`problem` (`~/.claude/scratch/relocations/classuse/samples.txt`) and every coining's definition
(`coined.jsonl`). "Uses" = mints in the 26 archived runs. ⚑ = a call I am least sure of.

**Table A — the values (30).** Order is the enum's, grouped so neighbours sit together.

| # | value | md | uses (→ after migrate) | definition (`(means)`) | nearest | distinguisher |
|---|---|---|---|---|---|---|
| 1 | `cross_section_contradiction` | always | 22 (→25) | two passages of the report state things that cannot both be true — figures, verdicts, grades, or the reasons given for them | `claim_contradicts_own_record` | are both statements in the report's text, or is one of them the run's record — an inquiry's fate, an edit, a proof's output? |
| 2 | `claim_contradicts_own_record` | always | 23 (→29) | the report and the run's own record disagree: the report states what the record contradicts — an inquiry's fate, method or reason, an edit, a proof's result — or a record entry contradicts what the report shows was done | `cross_section_contradiction` | (as row 1) |
| 3 | `reader_channel_mismatch` ⚑ | by_grade | 2 (→3) | a fact the reader needs is held — in the record or elsewhere in the report — but not where the reader who needs it reads, so a reader acting on that section acts without it | `claim_contradicts_own_record` | do the two places disagree, or does one simply lack what the other holds? |
| 4 | `incomplete_repair_propagation` | always | 2 | a repair fixed a defect at some of its sites and left it standing at others — a sibling sentence, row or section still carries what the repair removed | `cross_section_contradiction` | does the disagreement exist because a repair reached only part of one defect's sites, or do two passages disagree with no repair behind them? |
| 5 | `missing_root_invariant` ⚑ | by_grade | 1 | one fact is restated in several places with nothing tying the copies to a single source, so they have drifted or can drift unnoticed | `incomplete_repair_propagation` | is the defect that no single source governs the copies, or that one repair missed some of them? |
| 6 | `report_voice` | by_grade | 32 (→33) | report prose speaks about the run that made it rather than the subject — process voice, a claim wearing the seat that produced it, the report's own draft history, or the machinery that checked it (the seed's promoted definition) | `claim_contradicts_own_record` | does the fix correct a FACT the report gets wrong, or the AUDIENCE the sentence is written for, whether or not the fact is true? |
| 7 | `structure_noncompliance` | never | 3 | (registry's promoted definition, unchanged) | `risk_coverage_omission` | that is a risk the report does not grade; this is material the report holds with no order a reader can follow it by (neighbour moves: `risk-cross-term-ungraded` merges into `risk_coverage_omission`) |
| 8 | `false_universal` | always | 14 | a claim stated universally — every, none, always, the only — that a case the evidence reaches, or one it never tested, contradicts | `exhaustive_sweep_omits_case` | is it a rule about the subject stated too broadly, or a list or search presented as complete that leaves a member out? |
| 9 | `exhaustive_sweep_omits_case` | always | 4 | a list, survey, search or sweep the report presents as complete omits a member it should hold | `false_universal` | (as row 8) |
| 10 | `risk_coverage_omission` | by_grade | 12 (→14) | the report's account of risk leaves out a risk, or a combination of risks, that its own analysis or its subject makes live | `exhaustive_sweep_omits_case` | is the missing member a risk the case against owes a grade, or any other member of a list presented as complete? |
| 11 | `derivation_status_overclaim` | always | 11 | a claim carries more certainty or status than how it was established supports — inferred stated as verified, hedged in one place and flat in another, assumed stated as shown | `verification_scope_blindspot` | is the claim's STATUS overstated, or is real evidence spent on a claim broader than what it covers? |
| 12 | `verification_scope_blindspot` | always | 8 (→10) | a check, experiment or proof covers less than the claim it is credited with — a read test cited for a write path, one artifact's check credited to its siblings, a proof that establishes half its sentence | `derivation_status_overclaim` | (as row 11) |
| 13 | `self_attestation` | always | 4 (→6) | the report asserts its own checking — that it verified, re-ran, searched or matched something — and the record does not bear the assertion out | `derivation_status_overclaim` | is the overstated claim about the report's own verification or research, or about the subject? |
| 14 | `unverified_composition` ⚑ | always | 1 | a property of a whole is claimed because its parts were each checked, and the combination itself was never checked | `derivation_status_overclaim` | does the claim rest on combining separately checked parts whose combination nobody examined, or overstate how one claim was established? |
| 15 | `invalid_inference` ⚑ NEW | by_grade ⚑ | 0 (→2) | a reason the report offers for a conclusion does not entail it — the step fails even where every stated fact is true, including a check that gives the same result on the case the conclusion must exclude | `derivation_status_overclaim` | is the reason irrelevant to the conclusion, or relevant and stated with more certainty than it earns? |
| 16 | `vote_laundering` ⚑ | by_grade | 0 | agreement among sources, programs or reviewers is counted as independent corroboration when they share an origin, so the count overstates the evidence | `derivation_status_overclaim` | is the overstatement in how many independent supports exist, or in how one claim was established? |
| 17 | `doctrine_vs_implementation` | by_grade | 1 | what a system, method or report says it does differs from what its implementation does — a stated rule its code, script or procedure does not follow | `claim_contradicts_own_record` | is the disagreement between a stated rule and the mechanism meant to carry it out, or between a report statement and the run's record of what happened? |
| 18 | `policy_without_mechanism` | by_grade | 9 | a rule, safeguard or guarantee is asserted with nothing that enforces it | `partial_control_coverage` | is there no enforcing mechanism at all, or one that covers only part of the surface? |
| 19 | `partial_control_coverage` | by_grade | 0 (→6) | a control credited with covering a surface works on part of it and leaves the rest — an axis, channel, verb or case — uncovered | `self_defeating_mitigation` | the registry's existing distinguisher (coverage vs. failing on its own terms) |
| 20 | `self_defeating_mitigation` | by_grade | 0 (→1) | (registry's promoted definition, unchanged) | `policy_without_mechanism` | (registry's promoted distinguisher, unchanged) |
| 21 | `unhandled_degenerate_case` ⚑ | by_grade | 0 | a procedure or design the report describes breaks or is undefined at an edge input — empty, zero, one, a boundary, a duplicate — that the report does not address | `false_universal` | is a mechanism failing at an edge input, or a claim stated too broadly? |
| 22 | `spec_underspecification` | by_grade | 2 | a term, rule, procedure or recommendation the report relies on is not specified enough for a reader to apply or check it — undefined, defined only by what it is not, or left as an undecided either/or | `unhandled_degenerate_case` | is the specification missing or left open, or present and failing at an edge? |
| 23 | `figure_recount_fails` | always | 9 (→10) | a figure, count or computed result the report states differs from what an independent recount, recomputation or re-run produces | `metric_conflation` | is the number wrong on its own definition, or right on it and compared or identified with a figure measured differently? |
| 24 | `metric_conflation` ⚑ | always | 1 (→2) | two quantities or categories that measure or mean different things are treated as one — compared, added or given one label — across different bases, methods or definitions | `figure_recount_fails` | (as row 23) |
| 25 | `citation_figure_misattribution` | always | 3 (→6) | a cited source does not say what the report attributes to it — the figure, statement or support is absent from or contradicted by its text, the cached text is a different document from the one named, or the anchor backs a sentence its source is not about | `unsourced_claim` | is a source recorded against the claim that says something else, or is there no source behind the claim at all? |
| 26 | `within_source_condition_misattribution` | always | 2 (→4) | the report carries a source's or model's result outside the conditions under which the source establishes it — a condition, bound, regime or date the source states is dropped | `citation_figure_misattribution` | does the source state the result, under a condition the report drops, or not state it at all? |
| 27 | `live_source_drift` | always | 1 (→2) | a claim or proof rests on a source's or tree's current state rather than the state the report cites or pins, so it reads differently now from when it was written | `artifact_preservation` | is the artifact present and changed or read live, or missing or broken where the report says it is kept? |
| 28 | `artifact_preservation` | by_grade | 2 | an artifact a reader needs to check the report — a proof, program, cached source or data file — is not kept, or does not run, where the report says it is | `live_source_drift` | (as row 27) |
| 29 | `unsourced_claim` ⚑ NEW | by_grade ⚑ | 0 (→6) | a claim of fact or a quotation the report states with no source or proof behind it on the record — no anchor, an anchor that resolves to nothing in the evidence, or an appeal to a result it never names | `citation_figure_misattribution` | (as row 25) |
| 30 | `unclassified` | by_grade | 0 (→1) | no other value fits this gap. The mint's reason names the nearest value and what it does not cover | — | — |

Count: 27 registry classes kept, 2 new (⚑ F2), 1 catch-all.

**Table B — registry classes that do not stay (14).**

| registry slug (md, uses) | verdict | why |
|---|---|---|
| `enumeration-non-exhaustive` (always, 0) | MERGE → `exhaustive_sweep_omits_case` | synonym in practice: one of the 4 `exhaustive-sweep-omits-case` mints is "the enumeration omits two facts" |
| `figure-miscomposition` (always, 1) | MERGE → `figure_recount_fails` | its one mint is a count that fails recount (35 of 89 coprime); same fix |
| `causal-narrative-fails-reproduction` (always, 0) ⚑ | MERGE → `figure_recount_fails` | "fails reproduction" is the shared kind; alternative target `derivation_status_overclaim` |
| `pricing-basis-drift` (always, 0) | MERGE → `metric_conflation` | figures on different bases compared |
| `measurement-methodology-drift` (always, 0) | MERGE → `metric_conflation` | measurements by different methods compared |
| `citation-status-drift` (always, 1) | MERGE → `live_source_drift` | source changed after citing; its one mint (b6) reads as a misattribution, and migrate maps by slug, so it lands here as recorded |
| `audited-artifact-sibling-halo` (by_grade, 1) | MERGE → `verification_scope_blindspot` | one artifact's check credited to its siblings is a scope gap; the m28 mint keeps its recorded `by_grade` stamp (D5) |
| `undecided-disjunction` (by_grade, 0) | MERGE → `spec_underspecification` | an undecided either/or is an unspecified recommendation |
| `negative-definition` (by_grade, 0) | MERGE → `spec_underspecification` | a definition by what it is not is an underspecification |
| `recurrence-detector-keying` (by_grade, 0) | DROP → `unclassified` | an instance about this engine's own escalator, not a kind a report on any subject has |
| `name-keying-vs-marker` (by_grade, 0) | DROP → `unclassified` | engine-internal instance (identity recovered from a name); no distinguisher from `missing_root_invariant` that a lens could apply |
| `cross-corpus-id-collision` (always, 0) | DROP → `unclassified` | a seed singleton instance |
| `config-semantics-error` (by_grade, 0) | DROP → `unclassified` | no distinguisher from `citation_figure_misattribution`/`doctrine_vs_implementation` writable |
| `reflexivity-blindspot` (by_grade, 0) | DROP → `unclassified` | no distinguisher from `self_attestation` writable |

No archived mint uses a dropped class, so DROP costs no archived gap its class.

**Table C — the 33 coined slugs.** (mints; run)

| coined | → | | coined | → |
|---|---|---|---|---|
| research-protocol-violation (2; is-7-prime) | `self_attestation` | | unsourced-quotation-or-appeal (2; quadratic) | `unsourced_claim` |
| uncited-claims (1; is-7-prime-run2) | `unsourced_claim` | | graded-warrant-contradicts-record (2; quadratic) | `cross_section_contradiction` |
| unproven-computation (1; run2) | `unsourced_claim` | | unrecorded-report-edit (1; quadratic) | `claim_contradicts_own_record` |
| unsettled-inquiry (1; run2) | `claim_contradicts_own_record` | | misplaced-citation-anchor (1; quadratic) ⚑ | `citation_figure_misattribution` |
| contradictory-framing (1; run2) | `cross_section_contradiction` | | unsurfaced-inquiry-finding (1; b) ⚑ | `reader_channel_mismatch` |
| hypothesis-framing-incomplete (1; run2) | `unclassified` (its mint carries a `mint_reason`, so D7's check holds) | | nonreproducible-cited-proof (0; b) | `artifact_preservation` |
| silent-no-match-probe (1; record-store) | `verification_scope_blindspot` | | process-voice-leak (1; b3) | `report_voice` |
| risk-taxonomy-conflation (1; research-loop) ⚑ | `metric_conflation` | | orphaned-edit-artifact (0; b3) | `unclassified` |
| guardrail-scope-overclaim (3; sleeper) | `partial_control_coverage` | | inquiry-fate-mismatch (1; b6) ⚑ | `claim_contradicts_own_record` |
| unflagged-inferential-transfer (1; sleeper) | `within_source_condition_misattribution` | | record-status-erasure (1; b6) | `claim_contradicts_own_record` |
| confinement-scope-gap (1; sleeper) | `partial_control_coverage` | | inquiry-reason-record-mismatch (2; b8) ⚑ | `claim_contradicts_own_record` |
| structure-noncompliance (3; sleeper) | `structure_noncompliance` (adopted already) | | dangling-citation-anchor (1; b9) | `unsourced_claim` |
| mitigation-coverage-mismatch (2; sleeper) | `partial_control_coverage` | | citation-identity-mismatch (1; m23) | `citation_figure_misattribution` |
| mitigation-self-contradiction (1; sleeper) | `self_defeating_mitigation` | | invalid-inference (2; m28) | `invalid_inference` |
| risk-cross-term-ungraded (2; sleeper) | `risk_coverage_omission` | | non-discriminating-check (0; m28) | `invalid_inference` |
| model-scope-overclaim (1; sleeper) | `within_source_condition_misattribution` | | | |
| assertion-without-inquiry (1; sleeper) | `unsourced_claim` | | | |
| citation-support-mismatch (1; sleeper) | `citation_figure_misattribution` | | | |

**The calls I am least sure of** (fold into F1):

1. **The inquiry-* family → `claim_contradicts_own_record`** (4 slugs, 5 mints). The alternative is a
   new value `inquiry_record_mismatch`; I merged because the distinguisher the coiners wrote ("the
   line-of-inquiry record vs the report") is exactly this class's distinguisher from
   `cross_section_contradiction` — the inquiry record is part of the run's record.
2. **Two NEW values, `unsourced_claim` and `invalid_inference`** (F2). The registry has no class for
   "no source at all" (6 coined mints across 4 runs land there) or for a non sequitur (m28 coined it
   twice in one run); and both are `by_grade`, though the registry's own rule is that claim and citation
   classes are `always`. `by_grade` because an unsourced well-known fact, or a bad reason beside a true
   verdict, need not change what a reader concludes.
3. **`reader_channel_mismatch` redefined** (its original meaning is unrecorded; its only two mints,
   m28, fit the definition above) and receiving `unsurfaced-inquiry-finding`.
4. **Zero-use classes kept** — `unhandled_degenerate_case`, `vote_laundering`, `unverified_composition`
   (one misused mint), `missing_root_invariant`. Each has a writable distinguisher; all four come from
   the seed's design-review runs, which the archive under-samples. Dropping them is the cheaper
   vocabulary.
5. **`metric_conflation` broadened to categories** (absorbs `risk-taxonomy-conflation`).

### III.3 Directory tree

```
plugins/frank-exchange-of-views/
  requirements.json                                   [MODIFY] eventSchema 31 → 32
  docs/seat-command-triggers.md                       [MODIFY] rows `lens class list`, `lens class new` deleted
  docs/seat-surface-naming.md                         [MODIFY] `class new` dropped from the merge row
  docs/vocabulary.md                                  [MODIFY] regenerated (vocabdoc)
  agents/red-lens-*.md (7), lead-judge.md, red-chair.md [MODIFY] regenerated (agentgen)
  skills/adversarial-audit/SKILL.md                   [MODIFY] regenerated (agentgen)
  skills/research-protocol/scripts/debate.js          [MODIFY] authored lens prompt line (~990)
  tests/simulator/debate.test.mjs                     [MODIFY] `class new` leaves the command-path regex
  tests/simulator/testdata/prompt-red-lens-*.golden (4) [MODIFY] regenerated
  tools/internal/record/recordpb/record.proto         [MODIFY] GapClass, options, Mint, Event, NewMint
  tools/internal/record/recordpb/{record.pb.go,record.proto.sha256} [MODIFY] protogen
  tools/internal/record/recordpb/schema_gen.go        [MODIFY] schemagen (epoch 32)
  tools/internal/record/recordpb/facets.go            [MODIFY] enum-valued facet kind
  tools/internal/record/recordpb/gapclass.go          [NEW]    MaterialDefault(GapClass), Neighbor, Distinguisher
  tools/internal/record/recordpb/gapclass_test.go     [NEW]
  tools/internal/record/recordpb/testdata/payload-keys.txt [MODIFY]
  tools/internal/record/recordsql/schema.go           [MODIFY] enum-valued facet column
  tools/internal/record/recordsql/testdata/schema.sql [MODIFY] regenerated
  tools/internal/record/{replay.go,record.go,enums.go,enumvalue.go,enumof.go,writeview.go,viewjson.go} [MODIFY]
  tools/internal/record/classroster.go (+_test)       [DELETE]
  tools/internal/record/shippedclasses_gen.go         [DELETE]
  tools/internal/record/recordtest/classes_gen.go     [DELETE]
  tools/internal/record/{fixtureclasses_test.go,classmaterial_test.go} [DELETE / MODIFY]
  tools/internal/record/migrate/gapclass.go           [NEW]    the archived-slug → GapClass table
  tools/internal/record/migrate/gapclass_test.go      [NEW]    (replaces classmaterial_test.go)
  tools/internal/record/migrate/{migrate.go,replay.go,entries.go,remap.go,manifest.go,valuerenames.go} [MODIFY]
  tools/internal/cli/lens/class.go                    [DELETE]
  tools/internal/cli/lens/mint.go                     [MODIFY]
  tools/internal/cli/enumhelp/enumhelp.go             [MODIFY] render material/neighbour/distinguisher
  tools/internal/flags/names.go                       [MODIFY] Definition, Neighbor, Distinguisher, MaterialDefault removed
  tools/internal/setup/{setup.go,run.go,run_test.go}  [MODIFY]
  tools/internal/capture/capture.go                   [MODIFY] HarvestClasses + call deleted
  tools/internal/capture/classharvest_test.go         [DELETE]
  tools/internal/view/view.go (+_test)                [MODIFY] by_class tally, unclassified_rate
  tools/internal/seatprobe/{build.go,boards.go,seatprobe.go,boardclasses_test.go} [MODIFY]
  tools/internal/terms/terms.json                     [MODIFY] `cross-section-contradiction` → `cross_section_contradiction`
  tools/internal/difftest/testdata/*.golden (18), class_registry.golden [MODIFY / DELETE scenario]
  tools/releasegate/fuzz/{fuzz_test.go,registerbeforeappend_test.go} [MODIFY]
  tools/integration/surface/{chairrefusal,promptverbs,vocabulary_prose}_test.go [MODIFY]
  ~128 test files minting placeholder classes        [MODIFY] mechanical (§III.10)
feov-memory/class-registry.json                       [DELETE]
feov-memory/class-registry-seed.md                    [MODIFY] one line: the vocabulary now lives in record.proto (seed kept as design record)
law/proposed/class-*.md (14)                          [DELETE]
scripts/classgen/                                     [DELETE]
scripts/check/gates.go                                [MODIFY] classgen gate removed
scripts/agentgen/src/{fragments/lens-audit-and-mint.md,red-lens-voice.md,skills/adversarial-audit/SKILL.md} [MODIFY]
scripts/universe.sh                                   [MODIFY] feov-memory staging block removed
.github/workflows/hooks.yml                           [MODIFY] classgen step removed
docs/local-development.md                             [MODIFY] `report-voice` → `report_voice`
plans/class-registry-ships.md                         [DELETE] (superseded draft, untracked)
```

### III.4 Schema — `record.proto` and `recordpb` [MODIFY]/[NEW] (contract: event schema — census C1–C8)

- `extend google.protobuf.EnumValueOptions`: `optional ClassMaterial material_default = 50011;`
  `optional string neighbor = 50012;` `optional string distinguisher = 50013;` — each with the comment
  stating what reads it (the stamp; help and its test).
- `enum GapClass { GAP_CLASS_UNSPECIFIED = 0; … }` — Table A, one value per row, e.g.

  ```proto
  GAP_CLASS_FALSE_UNIVERSAL = 8 [(means) = "a claim stated universally — every, none, always, the only — that a case the evidence reaches, or one it never tested, contradicts",
    (material_default) = CLASS_MATERIAL_ALWAYS, (neighbor) = "exhaustive_sweep_omits_case",
    (distinguisher) = "is it a rule about the subject stated too broadly, or a list or search presented as complete that leaves a member out?"];
  GAP_CLASS_UNCLASSIFIED = 30 [(means) = "no other value fits this gap. The mint's reason names the nearest value and what it does not cover",
    (material_default) = CLASS_MATERIAL_BY_GRADE];
  ```
  The enum's comment states the membership rule (a value stays only with a definition and a
  distinguisher from its neighbour; adding one is a release, an epoch bump and nothing in migrate,
  because no archive holds the new word).
- `Mint.class`: `optional GapClass class = 3`, `required`, `why:` rewritten ("the KIND of defect, from
  the closed set `mint --help` lists; `unclassified` when none fits, with --reason saying why").
  `Mint` 4–7 (`class_new`, `definition`, `neighbor`, `distinguisher`) → `reserved 4, 5, 6, 7;
  reserved "class_new", "definition", "neighbor", "distinguisher";`. New message `(check)` (D7).
  `class_material`'s comment: stamped from the class's `material_default`.
- `EVENT_TYPE_CLASS_NEW = 7` → `reserved 7; reserved "EVENT_TYPE_CLASS_NEW";` and the `Event` oneof
  arm `class_new = 31` → `reserved 31`; `message ClassNew` deleted.
- `ClassMaterial`'s comment: "the class's `material_default` carries it" (no registry, no coiner).
- `NewMint`: `map<string,int32> by_class = 3` → `reserved 3`; `repeated ClassTally by_class = 5;`
  `optional double unclassified_rate = 6;`; `message ClassTally { optional GapClass class = 1;
  optional int32 count = 2; }`. Name check: `ClassTally` 0 hits.
- `recordpb/facets.go`: an enum-valued facet kind (`enumFacets = {"material_default":
  ClassMaterial}`), read with a new `EnumFacet(vd, name) (protoreflect.EnumNumber, declared, err)`,
  same declared/undeclared split and refusals as `Facet`/`Number`.
- `recordpb/gapclass.go` [NEW]: `MaterialDefault(GapClass) ClassMaterial` (via `EnumFacet`),
  `Neighbor(GapClass) (GapClass, bool)`, `Distinguisher(GapClass) string`. `gapclass_test.go`:
  `TestEveryGapClassDeclaresItsMaterialDefault`, `TestEveryGapClassNamesItsNeighbour` (every value but
  `unclassified` declares both options; `neighbor` resolves by `BySpelling` to a different value),
  `TestTheCatchAllIsByGradeAndHasNoNeighbour`.
- `recordsql/schema.go` `enumTable`: emits an enum-valued facet as `TEXT NOT NULL REFERENCES
  "enum_<enum>"("value")`; `enumTables` orders `enum_class_material` before `enum_gap_class` (a
  referenced table first). `schema.sql` golden regenerated; every changed line read.
- `requirements.json` `eventSchema` 32; `go -C scripts run ./schemagen`; `go -C scripts run ./protogen`.

### III.5 The record [MODIFY]/[DELETE]

- `replay.go`: delete `registryClass`, `classRegistry`, `loadRegistry`, `loadRegistryIn`,
  `classMaterialFor`, `knownClasses`, `StageForRun`, `StageForRunWithDefaults`, `stageRegistry`,
  `ClassCoinedInRun`, `classCoinedAt`, `validateClass`, `validateClassNew`. `stampClassMaterial`
  becomes: refuse a seat-supplied value (unchanged); under `Migrating` keep a recorded value
  (unchanged); else `m.ClassMaterial = recordpb.MaterialDefault(m.GetClass()).Enum()`.
- `record.go:790` keeps the stamp-first ordering; `:823` `case *recordpb.ClassNew` deleted.
- `enums.go`: `EnumFields["class_new"]` deleted; `EnumFields["mint"]` gains `{Key: "class", Flag:
  flags.Class, Values: evsOf(recordpb.GapClass(0).Descriptor()), Why: "the class is the KIND of
  defect, and its material default decides whether an open gap of it holds the PASS gate"}`.
- `enumvalue.go`: `EnumValue` gains `Material, Neighbor, Distinguisher string`, filled by `evsOf` for
  an enum whose values declare them (read through `recordpb`), empty otherwise.
- `enumof.go:116` `ClassMaterialOf` comment rewritten; add `GapClassOf(word)` beside it (the write
  path's word → value, as `GradeOf`).
- `writeview.go:49` `registry()` deleted; `runFiles` loses the records-dir read if nothing else uses it.
- `classroster.go`, `classroster_test.go`, `shippedclasses_gen.go`, `recordtest/classes_gen.go`,
  `fixtureclasses_test.go` deleted. `recordtest/recordtest.go:123–130` (seeded coining default)
  deleted.
- Every `GetClass()` reader (compiler-enumerated; C1–C2 hits `viewjson.go`, `graph.go:223`,
  `report/assemble.go:863`, `view.go:507–527`, `capture.go:2128`, `motion.go:353`) reads
  `recordpb.Word(m.GetClass())`; JSON keeps the key `class` holding the word.
- `castcoverage.go:20`, `duplicatescreen.go:38`, `record.go:828,1414`: comments respelled to the
  underscore word where they name a value.

### III.6 The lens surface [MODIFY]/[DELETE] (contract: CLI flag set — census C4, C5, C8)

- `cli/lens/class.go` deleted (the `class` group, `class list`, `class new`).
- `cli/lens/mint.go`: `--class` registered through `enumhelp.Flag(c, flags.Class,
  record.MustEnum("mint", "class"), "REQUIRED — the KIND of defect, from the values below; its
  material default is recorded with the gap. `unclassified` only when none fits, and then --reason
  names the nearest value and what it does not cover")`. `p.ClassNew` (`:95`) and the coining comment
  (`:88–92`, `:259–262`) deleted. `seat.Supplies(c, "class_material", "stamped from the class's
  material default")`. `--reason`'s help on `mint` gains the `unclassified` duty.
- `cli/enumhelp/enumhelp.go` `Section`: a value with `Material` renders `word  [material] means`, and
  with `Neighbor` a second indented line `nearest: <word> — <distinguisher>`. Other enums render
  byte-identically (goldens prove it).
- `flags/names.go:166–168` `Definition`, `Neighbor`, `Distinguisher` and `MaterialDefault` deleted
  (each has no other user — re-check with the compiler).
- `cli/` tests: `unknownclass_test.go` rewritten to the enum refusal (a near-miss `report-voice` names
  `report_voice`); `verbguardrace_test.go:225–242` (coin race) deleted; `cli_test.go:825–866`
  (coining) deleted; `badvalue_test.go:115`, `help_test.go:149`, `requiredmarker_test.go:96`,
  `referencechecks_test.go:121–153,265`, `projectionshape_test.go:288` updated to the new surface;
  `fixtureclasses_test.go` deleted (fixtures mint enum words, §III.10). New
  `TestUnclassifiedRequiresAReason` (refused without `--reason`, recorded with it, `class_material` =
  `by_grade`).

### III.7 setup [MODIFY]

- `setup.go:366–431`: `ValidateClassRegistry`, `StageClassRegistry` and their comments deleted.
- `run.go:230–237`: the pre-state registry check deleted. `run.go:280` staging and `:341–346` summary
  lines deleted — there is no class line in setup's output.
- `run_test.go:260–341`: `TestClassRegistryIsStagedIntoTheRun`, `TestSetupRefusesRegistryWithoutMaterialDefault`,
  `TestAbsentRegistryIsAnnounced` deleted; new `TestSetupReadsNoFeovMemory` — a run set up from a
  directory with no `feov-memory/` and one with a garbage `feov-memory/class-registry.json` produce
  byte-identical stdout and an identical records directory listing.
- `unresolvablerun_test.go:27` comment only. After the change `git grep -n feov-memory --
  plugins/frank-exchange-of-views/tools/internal/setup` prints only `run.go:362` (the scorecard
  comment, #1454's).

### III.8 migrate — one translation [NEW]/[MODIFY] (census C1, C2, C4)

The translation, in `migrate/gapclass.go`, registered in `valueRenames["feov.record.v1.GapClass"]`:

- **A kept registry slug** resolves with no row: `enumNumberOf`'s `SameWord` fold reads
  `report-voice` as `report_voice`.
- **A merged or dropped registry slug, and every coined slug** has an authored row (Table B + C:
  14 + 31 rows; `structure-noncompliance` and `invalid-inference` resolve by fold). `TestEveryArchivedClassLandsOnAValue`:
  every key is NOT a current word (else the row is dead), every value IS one.
- **`mint`** gets an `Entries()` row, `mapped("mint", nil, drops{"class_new": …, "definition": …,
  "neighbor": …, "distinguisher": …}, nil)`: the four columns are dropped with the reason "a class is a
  value of the closed set; whether a run coined it, and the coiner's definition, are not carried — the
  value's own definition is". (The jsonl era's mint-carried trio and the sqlite era's `class_new`
  column both arrive here.)
- **`class_new`** (and the jsonl spelling the era table already folds to it) → `{Translate:
  noEvent}`, the `revision` precedent; the manifest counts each in `events_in` and none in
  `events_out`. No archived reference points at a `class_new` event id (C4: nothing reads one), so the
  renumbering this causes is absorbed by replay's existing id translation.
- **Materiality.** A mint the source records with `class_material` keeps it (`stampClassMaterial`
  under `Migrating`, unchanged) — all 10 runs from 2026-10-09. A mint the source records without one
  (the 16 earlier runs, 123 mints, whose staged registries carry no default) is stamped by the live
  rule from its translated value — the same result today's fill gives (`ShippedMaterialDefaults` IS
  that table), recorded as a `StatedFill` only where the archived slug was translated to a different
  value (`Why`: "the source predates material_default; the gap takes the default of the class its
  archived slug now maps to").
- **The manifest** gains `class_translations: [{from, to, mints}]`, one row per archived slug that is
  not a kept value's spelling — a merge is visible, not silent.
- **The staged registry**: `copyChannels` skips `records/class-registry.json`; `translateStagedRegistry`,
  `shippedDefault`, `byGradeFillWhy` deleted; the `migrate.go:89–95` call removed.
- `remap.go`: `Mint.definition Mint.neighbor Mint.distinguisher ClassNew.*` leave `argumentFields`,
  `ClassNew.slug` and `Mint.class` leave `dataFields` (an enum, not a string — `TestEveryStringFieldIsInOneList`
  holds the lists to the schema).
- `manifest.go:74` (`Where: "staged registry"|"class_new event"`) rewritten for the class fill.
- Tests: `classmaterial_test.go` → `gapclass_test.go` with
  `TestTheGapClassMigratesOntoTheEnum` on `2026-10-18_is-91-prime-m28` (current epoch, two
  `class_new` events, two `invalid-inference` mints) and `2026-08-23_sleeper-service-plan` (jsonl era,
  mint-carried coinings, no `class_material`): every mint's class is a `GapClass` value; m28's
  `class_material` per `gap_id` equals the source's; sleeper's three `guardrail-scope-overclaim` mints
  are `partial_control_coverage` with `by_grade` and a stated fill each; zero `class_new` events; no
  `records/class-registry.json` in the destination; `class_translations` lists
  `invalid-inference → invalid_inference (2)`. `publish_test.go:429,455`, `gapplace_test.go:42`,
  `reproduceanchor_test.go:158` drop their registry staging. `TestEveryArchivedRunMigrates` is the
  census over all 27 tarballs.

### III.9 Telemetry [MODIFY]

- `view/view.go:500–565`: `byClass` becomes `[]*recordpb.ClassTally` in `GapClass` schema order
  (zero counts omitted); `unclassified_rate` = unclassified mints / mints this epoch, absent when none;
  `class_repeat_rate` excludes `unclassified` from the repeat numerator (D8). `view_test.go:504–521`
  updated; new `TestUnclassifiedRateIsAbsentWhenNothingWasMinted` and `…CountsOnlyTheCatchAll`.
- `inquest telemetry --help`'s OUTPUT shape is generated from the proto, so it changes to
  `by_class:[{class,count}]` and gains `unclassified_rate` with no hand edit; `lead-judge.md:91,330`
  and `red-chair.md:255` follow through `agentgen`. `scripts/agentgen/src/lead-judge.md:73` ("`new_mint.by_class`
  moving, and `class_repeat_rate` rising") stays true and is not edited.

### III.10 Harnesses and fixtures [MODIFY]

- **Fixtures mint enum words.** 349 lines in 128 test files mint with placeholder slugs (`x` 67,
  `scope-creep` 53, `overclaim` 39, `self-attestation` 33, `safety` 18, …; measured by
  `git grep -h -o -E '"--class", "[^"]+"|--class [a-z0-9_-]+|Class: *proto\.String\("[^"]+"\)'`).
  One mechanical pass: a placeholder that only says "some class" → `false_universal` (`always`, the
  common case); a test whose subject is materiality picks the value with the materiality it asserts
  (`structure_noncompliance` for `never`, `report_voice` for `by_grade`); a test proving an unknown
  class is refused keeps an unknown word. `Class: proto.String("x")` → `Class:
  recordpb.GapClass_GAP_CLASS_FALSE_UNIVERSAL.Enum()`. The 25 files calling `StageForRun*` (C1) lose
  the call.
- **Fuzz** (`releasegate/fuzz/fuzz_test.go`): the coining path (`:1681–1730`), `classMu`/`classMade`
  (`:343`), `fuzzClasses` (`:6086`) and the three `StageForRun` calls deleted; `class_new` leaves the
  event-type list (`:4460`) and the coverage table (`:4660`). Each mint draws its class from the enum
  so every materiality is exercised every run, and ~10% of mints use `unclassified` — half of those
  first without `--reason`, asserting the refusal, then with it. `registerbeforeappend_test.go:47`
  drops its staging.
- **seatprobe**: `build.go:89,146–150,419–427` read `recordpb.MaterialDefault` for `BoardClasses`
  (now `[]recordpb.GapClass`); `boards.go` class fields respelled (C12); the two `class list`/`class
  new` duty rows (`:536–537`) deleted and `:371` comment updated; `seatprobe.go:150–156` drops
  `EVENT_TYPE_CLASS_NEW`.
- **surface tests**: `chairrefusal_test.go:269` drops staging; `promptverbs_test.go:44,130,959` drop the
  `class new` example; `vocabulary_prose_test.go:265` exemption for `recordtest`'s generated list
  deleted (the list is gone).
- **difftest**: the `class_registry` scenario (`scenarios_test.go`, `class_registry.golden`) is
  rewritten as `gap_class_enum` (mint each materiality, the catch-all with and without a reason, a
  near-miss word); `contract_test.go:94`, `golden_test.go:174,190` drop staging. The other 17 goldens
  in C4/C12 regenerate.

### III.11 Dev tooling, CI, universe [DELETE]/[MODIFY]

- `scripts/classgen/` deleted; `scripts/check/gates.go:247` gate deleted; `.github/workflows/hooks.yml:1504–1506`
  step deleted (`scripts/check/parity_test.go` reconciles the two).
- `scripts/universe.sh:519–539` deleted. Nothing else in the universe needs `feov-memory`: `setup`
  read only the class registry from it (C10), and `universe.sh` does not run `capture`, whose
  `WriteScorecards` is the only other reader and already returns a reason on a missing directory.
- `feov-memory/class-registry.json` deleted. `feov-memory/class-registry-seed.md` stays (design
  record) with one present-tense line naming `record.proto`'s `GapClass` as where the vocabulary
  lives; its move to `plans/historical` is separate.

### III.12 Agent-facing text [MODIFY] (contract: prompts and constitutions — census C11, C12)

Generated carriers are regenerated, never hand-edited; each changed line and each affected `--help`
is read before and after (CLAUDE.md).

- `scripts/agentgen/src/fragments/lens-audit-and-mint.md:8`: "the registry's `material_default` is
  `always` (claim, figure and citation classes), `never` (structure-noncompliance) or `by_grade`" →
  "each class's material default, printed beside it in your mint page, is `always`, `never` or
  `by_grade` (medium and above)"; `report-voice` → `report_voice`.
- `scripts/agentgen/src/red-lens-voice.md:13`: "Mint it as class `report_voice` — the one voice class,
  and it holds all four kinds … A voice defect it does not cover is `unclassified`, with the reason
  saying what it is." (the coining sentence goes).
- `scripts/agentgen/src/skills/adversarial-audit/SKILL.md:28`: `report-voice` → `report_voice`,
  "the registry's voice class" → "the voice class".
- `skills/research-protocol/scripts/debate.js:990` (authored lens prompt): "registering a new class
  first where the registry lacks one" → "under the class that fits, or the catch-all with the reason
  none does". Names an act, no command path (`TestTheSeatPromptsNameNoVerb`).
- `tests/simulator/debate.test.mjs:517–518`: `class new` leaves the command-path regex (the path no
  longer exists; a ban list of a deleted word is the archaeology the repo refuses).
- `tools/internal/terms/terms.json:727`: "the defect class `cross-section-contradiction`" →
  `cross_section_contradiction`; `go -C scripts run ./vocabdoc`.
- `go -C scripts run ./agentgen`: the seven lens constitutions lose the `class` group, `class list`
  and `class new` pages and gain the `--class` enumerated block; `lead-judge.md`/`red-chair.md` take the
  new telemetry shape. `go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen` is NOT
  needed (its inputs are `Plan`, `OccasionWords`, `GradeWords`, `MassMappingVersion` — no class).
- Prompt goldens `prompt-red-lens-{dark-side,evidence,evidence-engaged,logic}.golden` regenerate.
- Not changed: every `policy-without-mechanism` in `scripts/{archaeology,rulesweep,versionguard}`,
  `hooks.yml:1152` and `prosthetic-conscience/skills/design-by-contract/SKILL.md:27` — they name the
  PROTOCOL class registry (`feov-memory/protocol-class-registry.md`, `### policy-without-mechanism` at
  line 85), a different concept sharing a slug (facts-are-fields clause 4).

### III.13 Docs [MODIFY]

`docs/seat-command-triggers.md:73–74` (the two rows deleted), `docs/seat-surface-naming.md:252`
(`class new` dropped), `docs/local-development.md:207` (`report_voice`), `docs/vocabulary.md`
(generated). Plans are design records and are not edited.

### III.14 Consumer census

Commands run at `be6c110d` from the worktree root with `-- ':!run-archive' ':!plans/historical'
':!research'`. Full output: `~/.claude/scratch/relocations/gapenum/census/c*.txt`; the compact form
below is `file lines` (I/ = `plugins/frank-exchange-of-views/tools/internal/`, F/ =
`plugins/frank-exchange-of-views/`). `plans/` and `law/` hits are counted, not listed: plans are design
records (no change); `law/proposed/class-*.md` (14) are deleted (III.11).

Re-running each command after the change must print only the "no change" consumers below.

<details><summary>C1 — <code>git grep -nE 'ShippedClasses|ShippedMaterialDefaults|StageClassRegistry|ValidateClassRegistry|loadRegistry|StageForRun|ClassRoster|knownClasses|classMaterialFor|stampClassMaterial|HarvestClasses|translateStagedRegistry|shippedDefault|registry\(\)'</code> — 141 lines</summary>

```
feov-memory/class-registry.json 2
F/tools/integration/surface/chairrefusal_test.go 269
I/capture/capture.go 1830,2138,2155
I/capture/classharvest_test.go 42,72,86,105,106,125
I/cli/fixtureclasses_test.go 13,27,50
I/cli/lens/class.go 59,72
I/cli/referencechecks_test.go 121,143
I/cli/seatprobe_fixture_test.go 106
I/difftest/golden_test.go 174
I/record/classmaterial_test.go 23,142,145,175
I/record/classroster.go 17,24,36,41,42,46
I/record/classroster_test.go 35,44,75,128,155
I/record/fixtureclasses_test.go 12,24,41
I/record/migrate/classmaterial_test.go 103,190
I/record/migrate/gapplace_test.go 42
I/record/migrate/migrate.go 91,221,226,228,254
I/record/migrate/replay.go 55,57,58,318
I/record/migrate/reproduceanchor_test.go 158
I/record/record.go 790
I/record/recordpb/descriptions.go 75
I/record/recordtest/classes_gen.go 5,12
I/record/recordtest/recordtest.go 123
I/record/replay.go 418,433,445,453,458,460,…(+20)
I/record/replay_test.go 605
I/record/shippedclasses_gen.go 7,12
I/record/writeview.go 49,50
I/report/fixtureclasses_test.go 13,24,41
I/reportproj/mintbudget_test.go 46
I/seatprobe/boardclasses_test.go 22
I/seatprobe/build.go 89,146,148
I/setup/run.go 233,280
I/setup/run_test.go 260,280
I/setup/setup.go 366,369,384,388,394,406,…(+4)
I/setup/unresolvablerun_test.go 27
I/sittinghook/join_test.go 57
F/tools/releasegate/fuzz/fuzz_test.go 3560,4545,6133
F/tools/releasegate/fuzz/registerbeforeappend_test.go 47
scripts/classgen/main.go 8,169,176,194,199
scripts/rulesweep/main.go 57
scripts/rulesweep/sweep.go 112
scripts/rulesweep/sweep_test.go 182,185
plans: 17
```
</details>

| Consumer | Changes? |
|---|---|
| `classroster.go`/`_test`, `shippedclasses_gen.go`, `recordtest/classes_gen.go`, `fixtureclasses_test.go` ×3, `classharvest_test.go`, `scripts/classgen` | deleted |
| `replay.go`, `record.go:790`, `writeview.go`, `recordtest.go:123` | yes — III.5 |
| `setup.go`, `run.go`, `run_test.go` | yes — III.7; `unresolvablerun_test.go:27` comment only |
| `migrate.go`, `migrate/replay.go`, `migrate/classmaterial_test.go`, `gapplace_test.go`, `reproduceanchor_test.go` | yes — III.8 |
| `capture.go:1830,2138,2155` | yes — `HarvestClasses` and its call deleted (III.11 ruling) |
| `cli/lens/class.go` | deleted |
| `seatprobe/build.go`, `boardclasses_test.go`, `cli/seatprobe_fixture_test.go` | yes — III.10 |
| `chairrefusal_test.go`, `referencechecks_test.go`, `difftest/golden_test.go`, `classmaterial_test.go`, `replay_test.go:605`, `mintbudget_test.go`, `join_test.go`, fuzz ×2 | yes — staging calls removed (III.10) |
| `recordpb/descriptions.go:75` | no — matched `registry()` in a terms-registry sentence |
| `scripts/rulesweep/*` | no — `registry()` is the PROTOCOL class registry |

<details><summary>C2 — <code>git grep -nE 'class-registry'</code> — 94 lines</summary>

```
.github/workflows/hooks.yml 1065
CLAUDE.md 63
feov-memory/class-registry.json 2,3
feov-memory/protocol-class-registry.md 3
F/docs/seat-command-triggers.md 73
I/capture/capture.go 2140,2179
I/cli/lens/class.go 45
I/cli/referencechecks_test.go 136,153
I/difftest/contract_test.go 94
I/difftest/golden_test.go 190
I/difftest/scenarios_test.go 59
I/record/classmaterial_test.go 148,245,269
I/record/classroster.go 13,50
I/record/classroster_test.go 152
I/record/legacyformat_test.go 75
I/record/migrate/classmaterial_test.go 121,154,275,279
I/record/migrate/migrate.go 29,229
I/record/migrate/publish_test.go 429,455
I/record/recordtest/classes_gen.go 1
I/record/replay.go 463,586,661
I/record/replay_test.go 59,611
I/record/shippedclasses_gen.go 1
I/record/writeview.go 37
I/seatprobe/build.go 419
I/setup/run_test.go 260,270,278,302
I/setup/setup.go 370,389,392,403,412,431
plugins/prosthetic-conscience/skills/design-by-contract/SKILL.md 24
scripts/archaeology/sweep.go 14
scripts/classgen/main.go 8,11,58
scripts/rulesweep/main.go 25,49,53
scripts/rulesweep/sweep.go 16,241
scripts/universe.sh 522,536
law: 14   plans: 20
```
</details>

| Consumer | Changes? |
|---|---|
| `hooks.yml:1065`, `CLAUDE.md:63`, `protocol-class-registry.md`, `design-by-contract/SKILL.md:24`, `archaeology/sweep.go`, `rulesweep/*` | no — `protocol-class-registry.md`, another concept |
| `seat-command-triggers.md:73`, `capture.go`, `lens/class.go`, `classroster*`, `_gen.go`, `classgen`, `setup.go`, `run_test.go`, `universe.sh` | yes / deleted — III.6–III.13 |
| `replay.go`, `writeview.go`, `migrate.go`, `seatprobe/build.go:419` | yes — III.5, III.8, III.10 |
| `referencechecks_test.go`, `difftest/{contract,golden,scenarios}_test.go`, `classmaterial_test.go`, `legacyformat_test.go:75`, `migrate/{classmaterial,publish}_test.go`, `replay_test.go` | yes — staged-file assertions removed; `legacyformat_test.go:75` re-read at implementation (it names the staged path in an old-format fixture) |

<details><summary>C3 — <code>git grep -nE 'classgen'</code></summary>

```
.github/workflows/hooks.yml 1505
I/record/recordtest/classes_gen.go 1
I/record/shippedclasses_gen.go 1
scripts/check/gates.go 247
scripts/classgen/main.go 1,84,92,116,121,127,…(+3)
plans: 12
```
</details>

All change (deleted) — III.11.

<details><summary>C4 — <code>git grep -nE 'class_new|ClassNew|CLASS_NEW|classNew|"class new"|class new|ClassNewVerb'</code> — 238 lines</summary>

```
feov-memory/class-registry.json 4
F/agents/red-lens-{adversary,architecture,computation,dark-side,evidence,logic,voice}.md 2 each (~220, ~417)
F/docs/seat-command-triggers.md 74
F/docs/seat-surface-naming.md 252
F/tests/simulator/debate.test.mjs 517,518
F/tools/integration/surface/promptverbs_test.go 44,130
I/capture/capture.go 2106,2107,2118,2188
I/capture/classharvest_test.go 19,67,71
I/cli/badvalue_test.go 115
I/cli/cli_test.go 825,840,856,857,859
I/cli/help_test.go 149
I/cli/lens/class.go 35,103,107,120,131,137
I/cli/lens/mint.go 88,89,95,260,262
I/cli/referencechecks_test.go 265
I/cli/requiredmarker_test.go 96
I/cli/unknownclass_test.go 14,15,33,34,41
I/cli/verbguardrace_test.go 225,227,229,241,242
I/diagnostics/manual_test.go 28,33,122
I/difftest/testdata/{a_repeated_act_names_its_answer,carried_from_renders_as_carried,class_registry,close_validation_and_archive,docket_ruling_requires_each_unconditional_field,error_catalogue,full_round_integration,hostile_prose_via_file,ids_stay_distinct_across_rounds,mint_idempotency_on_crash_retry,mint_validation,missing_required_flags,multi_nonce_terminal_event_wins,projections_debate_changelog_citations,regrade_history_is_recoverable}.golden
I/record/classmaterial_test.go 180,182,183,195,196
I/record/classroster.go 11,17,27,91,95
I/record/classroster_test.go 25,66
I/record/enumof.go 117
I/record/enums.go 293,295
I/record/migrate/classmaterial_test.go 69,71,77,83,92,108,…(+5)
I/record/migrate/manifest.go 74
I/record/migrate/remap.go 84,85,96
I/record/migrate/replay.go 52,317,322
I/record/record.go 823,824
I/record/recordpb/body_test.go 25
I/record/recordpb/record.pb.go 64,105,143,…(+37)
I/record/recordpb/record.proto 237,579,717,722,724,824,833,834
I/record/recordpb/testdata/payload-keys.txt 20
I/record/recordsql/testdata/schema.sql 96,483,502,533
I/record/recordtest/recordtest.go 129,130
I/record/replay.go 508,560,586,670,679,696,700,708
I/record/replay_test.go 692,693,700,702,703,704,…(+6)
I/record/writerace_test.go 70,71
I/seatprobe/boards.go 371,537
I/seatprobe/seatprobe.go 81,150,156
I/setup/run.go 343
F/tools/releasegate/fuzz/fuzz_test.go 1681,1710,1725,4460,4660
law: 14   plans: 16
```
</details>

Every hit changes: generated (agents, `record.pb.go`, goldens, `schema.sql`, `payload-keys.txt`)
by regeneration; the rest by III.4–III.13. `diagnostics/manual_test.go:28,33,122,127` names the
`class` group as a traversal fixture and is updated to a group that still exists; `body_test.go:25`
drops the `ClassNew` row; `writerace_test.go:70` (a coining racer) is deleted.

<details><summary>C5 — <code>git grep -nE 'class list|classList|class_list'</code></summary>

```
F/agents/red-lens-*.md (7) one each (~200)
F/docs/seat-command-triggers.md 73
F/tests/simulator/debate.test.mjs 397
F/tools/integration/surface/vocabulary_prose_test.go 265
I/cli/lens/class.go 40,54,63,67,71,76,79
I/cli/projectionshape_test.go 288
I/diagnostics/manual_test.go 127
I/fetchcache/testdata/walls/acl.html 167
I/fetchcache/testdata/walls/arxiv-control.html 4
I/record/available.go 122
I/record/classroster_test.go 138
I/report/site.js 16,56,142
I/seatprobe/boards.go 536
F/tools/releasegate/fuzz/fuzz_test.go 1693,1695
scripts/rulesweep/sweep_test.go 94
plans: 2
```
</details>

No change: `debate.test.mjs:397` ("…NOT MATERIAL, BY CLASS…" — the chair's by-class listing of
non-material gaps, which survives), `record/available.go:122` (same listing), `fetchcache/testdata`,
`report/site.js` (CSS `classList`), `vocabulary_prose_test.go:265` (removed with III.10),
`rulesweep/sweep_test.go:94` (protocol registry). The rest change (III.6, III.10).

<details><summary>C6 — <code>git grep -nE 'by_class|class_repeat_rate|ByClass|ClassRepeat'</code></summary>

```
F/agents/lead-judge.md 91,330
F/agents/red-chair.md 255
I/difftest/testdata/manual_runs_every_help_page_live.golden 208
I/record/enums.go 159,165,169,195
I/record/enums_test.go 439
I/record/recordpb/record.pb.go 2822,…(+14)
I/record/recordpb/record.proto 674,676,678,680
I/record/recordpb/testdata/payload-keys.txt 12,21
I/reportvoice/census_test.go 23
I/reportvoice/tells.go 239,240
I/view/view.go 560,561,815,844,845,901,902
I/view/view_test.go 504,506,509,510,514,515,521
scripts/agentgen/src/lead-judge.md 73
plans: 1
```
</details>

No change: `enums.go`/`enums_test.go` (`artifactByClass` — closure classes), `reportvoice/*`
(`CountByClass` — voice-tell classes), `agentgen/src/lead-judge.md:73` (still true). Change:
`record.proto`, `view.go`, `view_test.go`, and the generated agents, golden, `payload-keys.txt`.

<details><summary>C7 — <code>git grep -nE 'material_default|MaterialDefault|class_material|ClassMaterial'</code> — 480 lines, by file</summary>

Non-test Go and prose: `record/replay.go`, `record/enumof.go`, `record/record.go`, `record/enums.go`,
`record/recordpb/record.proto`, `cli/lens/class.go`, `cli/lens/mint.go:285`, `flags/names.go`,
`setup/setup.go`, `migrate/{migrate,replay,manifest}.go`, `seatprobe/build.go`, `capture/capture.go`,
`record/recordsql/views.go` (reads `mint.class_material` — unchanged, D5), `feov-memory/class-registry.json`,
`scripts/classgen`, `scripts/agentgen/src/fragments/lens-audit-and-mint.md`; plus ~60 test files and the
generated `record.pb.go`/`schema.sql`.
</details>

`views.go` and every reader of `mint.class_material` do not change (D5); the writers change as
III.4–III.8. Tests that build a `Mint` with `ClassMaterial` set by hand are untouched unless they
also stage a registry.

<details><summary>C8 — <code>git grep -nE 'flags\.Class\b|flags.Definition|flags.Neighbor|flags.Distinguisher|"definition"|"neighbor"|"distinguisher"'</code></summary>

```
F/tools/integration/surface/promptverbs_test.go 959
I/cli/cli_test.go 864,865,866
I/cli/lens/class.go 121,…(+10)
I/cli/lens/mint.go 93,262
I/cli/motion/command.go 74
I/cli/motion/verbs.go 129,131,226,476
I/flags/names.go 166,167,168
I/record/classroster.go 95
I/record/classroster_test.go 145,146
I/record/recordsql/testdata/schema.sql 484,…(+5)
I/record/replay.go 424,425,426,716,717,718
I/record/replay_test.go 702,703,704
I/terms/terms.go 56, terms.json (40), terms_test.go (13), scripts/vocabdoc/main.go 44
```
</details>

No change: `cli/motion/*` (`flags.Class` is the PETITION class flag, `PetitionClass`), `terms/*` and
`vocabdoc` (the terms registry's own `definition` key). The rest change or are deleted.

<details><summary>C10 — <code>git grep -n 'feov-memory'</code> (beyond C2)</summary>

`capture.go:1632,1814,1819,2007,2223`, `capture/readwhen_test.go:40`, `cli/scorecard.go:61`,
`cli/scorecardunreadable_test.go:33`, `scorecard/scorecard.go:316`, `scorecard/inquiryzero_test.go:28`
(scorecards, #1454); `setup/run.go:362` (scorecard comment); `hooks.yml:1065`, `rulesweep`,
`archaeology`, `design-by-contract/SKILL.md:24` (protocol registry); `fixtureclasses_test.go` ×3
(deleted).
</details>

No change beyond the deletions: nothing else in `setup` or `universe.sh` reads `feov-memory`.

<details><summary>C11 — agent-facing prose: <code>git grep -n -i -E 'class(es)? (new|list)|coin|registry|gap.class|material_default|by_class|class_repeat' -- plugins/frank-exchange-of-views/skills '**/debate.js' scripts/agentgen/src terms.json plugins/frank-exchange-of-views/docs README.md docs</code> — 32 lines</summary>

```
docs/setup-script.md 80,139                         ("marketplace registry" — no change)
F/docs/seat-command-triggers.md 73,74               (change)
F/docs/seat-surface-naming.md 252 (change), 666 ("coin flip" — no change)
F/docs/vocabulary.md (11)                           ("REGISTRY-ONLY" — terms registry; regenerated only for III.12)
F/skills/adversarial-audit/SKILL.md 48              (generated; change via src)
F/skills/research-protocol/scripts/debate.js 990    (change)
I/terms/terms.json (9)                              ("REGISTRY-ONLY" kinds — no change)
scripts/agentgen/src/fragments/lens-audit-and-mint.md 8  (change)
scripts/agentgen/src/lead-judge.md 73               (no change)
scripts/agentgen/src/red-lens-voice.md 13           (change)
scripts/agentgen/src/skills/adversarial-audit/SKILL.md 28 (change)
```
</details>

<details><summary>C12 — any of the 41 registry slugs, outside generated agents/skills — 186 lines</summary>

Non-test, non-generated: `seatprobe/{boards,build}.go` (change), `record/{castcoverage,duplicatescreen,record}.go`
(comments respelled), `terms.json:727` (change), `agentgen/src` ×3 (change), `docs/local-development.md:207`
(change), `docs/vocabulary.md:422` (regenerated), `hooks.yml:1152`, `scripts/{archaeology,rulesweep,versionguard}`,
`design-by-contract/SKILL.md:27` (protocol registry — no change), `_gen.go` ×2 (deleted). Tests: 43 files
(`capture/*`, `cli/*`, `dashboard/*`, `record/*`, `record/migrate/*`, `setup/run_test.go`, `view/telemetryrounds_test.go`,
fuzz) — respelled by the fixture pass (III.10).
</details>

**Rendered, not grepped.** `mint --help`, `class --help`, `class new --help`, `class list --help` and
the `mint --class nope` refusal were rendered from the built `be6c110d` binary into
`~/.claude/scratch/relocations/gapenum/help-*.txt` (quoted in §I and §II). §V.3 renders them again
after the change.

**Census size:** 12 queries; ≈ 300 tracked files hit outside `plans/`; ≈ 200 change (≈ 128 by the
mechanical fixture pass, ~35 regenerated, the rest by hand), 14 law files and 9 code/tool files deleted, the
remainder verified no-change above.

### III.15 Interactions

- **#1452 (law retired).** This PR deletes `HarvestClasses`, its capture output lines and the 14
  `law/proposed/class-*.md`; `HarvestPrecedents` and `law/proposed/` itself stay for #1452, as do
  setup's `<cwd>/law` mirror and its `law:` summary line. Either order merges cleanly; if #1452 lands
  first, its deletion of `capture.go`'s harvest block makes III.11's capture edit a no-op.
- **#1454 (scorecards from the archive).** Untouched: `WriteScorecards` and `feov-memory/*-scorecard.md`
  stay. A cross-run `unclassified` rate (F4) is #1454's to add, reading the archived telemetry.
- `protocol-class-registry.md` and its gates are another concept and are not touched.

## IV. Risk & Mitigation

| # | Risk | L | I | Mitigation cost | Mitigation (step) |
|---|---|---|---|---|---|
| R1 | An archived run fails to migrate (an unmapped slug, a dropped column, the D7 check on a catch-all mint without a reason) | Med | High — the census test is red | Low | `TestEveryArchivedRunMigrates` over 27 tarballs; `TestEveryArchivedClassLandsOnAValue`; measured: the one catch-all mint carries a `mint_reason` (III.8, §V.5) |
| R2 | A migrated gap changes materiality | Med (merges move defaults) | High — an archived board reads differently | Low | D5 keeps every recorded stamp; m28 test compares `class_material` per `gap_id`; pre-field fills are stated (III.8) |
| R3 | The enum-valued facet breaks the schema build (FK order inside the DDL transaction) | Med | Build/test break, loud | Low | `enumTables` orders the referenced table first; `schema.sql` golden (III.4) |
| R4 | `enumhelp` change alters other enums' help | Low | Golden churn misread as intended | Low | other enums render byte-identically; `manual` golden diff read line by line (§V.3) |
| R5 | Help/constitution size: ~30 values × ~450 chars ≈ 13 KB added to `mint --help` and to each of 7 lens constitutions (~100 KB each today), and to every wrong-word refusal | High (certain) | Medium — per-sitting context | Low to state, a fork to reduce | F3; definitions held to one sentence; the near-miss refusal names the word |
| R6 | Definitions trip the vocabulary gate (`TestVocabularyProse` bans variants such as "the debate") | Med | Gate red | Low | run the gate early; Table A already avoids run-voice words |
| R7 | The fixture pass changes what a test asserts (a materiality test silently picks an `always` class) | Med | A test passes for the wrong reason | Med | rule in III.10; delete-the-row checks in §V.9 |
| R8 | A carrier still teaches coining (agent prose, prompt, docs) | Med | Seats look for a verb that is gone | Low | §V.8 sweep; C4/C5/C11 re-run; rendered constitutions grepped |
| R9 | The universe carries published content, not this checkout | Med (it has happened) | False pass of §V.7 | Low | §V.7 greps the installed binary's `mint --help` for `unclassified` and the epoch for 32 |
| R10 | Seat environment leaks into tests | High in an agent session | Reads as a code defect | Low | every §V step starts with the `FEOV_*` unset |
| R11 | Smoke lenses never choose `unclassified` | High | Criterion 5 unmet by the run alone | Low | §V.7 drives one catch-all mint by hand on a copy of the finished run |

## V. Verification Plan

`W=/home/gblock_ctoforaday_com/projects/special-circumstances/.claude/worktrees/bridge-cse_0154Z23DoVFryzFDGWWcWHxj`,
`T=$W/plugins/frank-exchange-of-views/tools`, scratch `~/.claude/scratch/relocations/gapenum/`. Every
step that runs Go or `feov-record` begins with `unset $(env | grep -o '^FEOV_[A-Z_]*')`. Long output
goes to a file; read the fail count, not the exit code. Check the box load before believing a red or a
timing.

1. **Generators.** `go -C $W/scripts run ./protogen && go -C $W/scripts run ./schemagen && go -C $W/scripts run ./agentgen && go -C $W/scripts run ./vocabdoc`
   then the `check` gates covering them (`agentgen -check`, `schemagen -check`, the protogen sha, `TestVocabularyProse`) → green. `feov-record --schema` → `32`. · re-armed by: `record.proto`,
   `requirements.json`, `agentgen/src/**`, `terms.json`.
2. **FEOV tests, uncached.** `go -C $T vet ./... && go -C $T test -count=1 ./... > ~/.claude/scratch/relocations/gapenum/test.out 2>&1; grep -c -- '--- FAIL' …/test.out`
   → 0; `grep -c '^ok' …/test.out` equals the package count. Includes `recordpb/gapclass_test.go`,
   `TestUnclassifiedRequiresAReason`, the telemetry tests, `TestSetupReadsNoFeovMemory`, the migrate
   tests. · re-armed by: anything under `$T`.
3. **Goldens, read.** `UPDATE_GOLDENS=1 go test -count=1 ./internal/difftest/... ./internal/record/recordsql/... ./internal/record/recordpb/...`
   in `$T`, and `go -C $W/scripts run ./golden -review` for the prompt goldens; then `git -C $W diff --stat -- '*.golden' '*.sql' '*testdata*'`
   and READ every changed line: the only changes are the `class` group's disappearance, `--class`'s
   enumerated block, underscore class words, the telemetry shape, `schema.sql`'s `enum_gap_class`/`mint`
   FK/check and `class_new` removal. Render, before and after, one at a time: `mint --help`,
   `inquest telemetry --help`, the lens top-level `--help`, `mint --class report-voice …` (near-miss)
   and `mint --class unclassified …` without `--reason`. · re-armed by: any help string, enum, golden.
4. **Setup from an empty project, built binary (driveable).** In `~/.claude/scratch/relocations/gapenum/proj`
   (fresh `git init`, no `feov-memory/`), with the binary under `bin/` and `requirements.json` beside it:
   `feov-record --seat-id operator setup <proj>/research/probe --model haiku --judgment-model haiku`
   → exit 0, no `class registry` line; then as `red-lens-logic`: `register`, a `finding`, and `mint
   --class false_universal …` → recorded with `class_material: always`; `mint --class unclassified …`
   without `--reason` → refused naming the check's `why`; with `--reason` → recorded `by_grade`;
   `inquest telemetry` as `red-chair` → the epoch's `unclassified_rate` equals the catch-all's share of that epoch's mints, and `by_class` lists both values. · re-armed by: setup,
   mint, view.
5. **migrate on archived runs.** `go -C $T test -count=1 -run 'TestTheGapClassMigratesOntoTheEnum|TestEveryArchivedClassLandsOnAValue|TestEveryArchivedRunMigrates' ./internal/record/migrate/`
   → ok, 27 subtests in the census. Then driveable on the real archive: extract
   `run-archive/2026-10-18_is-91-prime-m28.tar.gz` into scratch, `feov-record migrate` it with the built
   binary, and read the manifest's `class_translations` and `stated_fills` and `SELECT class,
   class_material, count(*) FROM mint GROUP BY 1,2` on the result (`file:` URI, `mode=ro`) against the
   same query on the source. · re-armed by: `migrate/**`, `record.proto`.
6. **Release-gate fuzz.** `cd $T && FEOV_RELEASE_GATE=1 go test -count=1 ./releasegate/fuzz/` → ok;
   the fuzz's own coverage table shows `unclassified` minted and refused-without-reason at least once.
   If red, run the same on `origin/main` before blaming the branch.
7. **Universe smoke — mints with no feov-memory, one catch-all.**
   ```
   $W/scripts/universe.sh build
   U=~/.claude/scratch/universe/plugins/cache/special-circumstances/frank-exchange-of-views
   ls $U/*/bin/feov-record && $U/*/bin/feov-record --schema            # → 32
   grep -c 'GAP_CLASS_UNCLASSIFIED\|unclassified' $U/*/bin/feov-record    # ≥1: this checkout
   grep -c 'class new' $U/*/agents/red-lens-logic.md                        # 0
   MODEL=haiku JUDGMENT_MODEL=haiku LANES=1 $W/scripts/universe.sh run "is 91 prime" --mint-budget 3 --k-max 2
   ```
   (stream-json via `universe.sh`, never `--output-format json`; `universe.sh watch` read-only; confirm
   the universe bills the subscription before launching — memory "universe billed API credits".)
   Then: `test ! -e ~/.claude/scratch/universe/src/feov-memory`; the run log has no `staged feov-memory`
   line; the run's own `.bin/feov-record` (not a PATH lookup) reports ≥1 `mint` and zero class
   refusals in the transcripts. If no lens chose `unclassified`, copy the finished run directory and,
   with its `.bin/feov-record` as `red-lens-logic`, `register` and mint once with `--class
   unclassified` without and then with `--reason`; record which path satisfied the criterion.
   · re-armed by: `universe.sh`, setup, mint, agentgen output.
8. **Carrier sweep.** Success criterion 6's `git grep` over `plugins/ scripts/ .github/ docs/` → nothing;
   `git -C $W grep -n -E "\b($(cat ~/.claude/scratch/relocations/gapenum/slugs.re))\b" -- plugins/frank-exchange-of-views/agents plugins/frank-exchange-of-views/skills scripts/agentgen/src plugins/frank-exchange-of-views/tools/internal/terms`
   → nothing (hyphenated class words are gone from seat-facing text). Re-run C1–C12 and compare to
   the no-change lists in III.14.
9. **Delete-the-row / delete-the-call** (each restored after): delete one `GapClass` value's
   `(material_default)` → schema build refuses (partly annotated); delete one `(neighbor)` →
   `TestEveryGapClassNamesItsNeighbour` fails; delete one row of `migrate/gapclass.go` →
   `TestEveryArchivedRunMigrates` fails for that run; delete the `Mint` `(check)` →
   `TestUnclassifiedRequiresAReason` fails; make `stampClassMaterial` stamp `by_grade` unconditionally →
   the m28 migrate test and a cli materiality test fail; exclude-nothing in `class_repeat_rate` →
   `…CountsOnlyTheCatchAll` fails. Any that stays green is a gap closed before done.
10. **All local gates.** `go -C $W/scripts run ./check > …/check.out 2>&1` → every gate PASS, fail count
    read; `-list` no longer lists `classgen`. A workflow edit (`hooks.yml`) needs CI's `actionlint`
    before a green claim.
11. **Auditor gate.** `/prosthetic-conscience:plan-audit plans/gap-class-enum.md` → PASS, after F1–F4
    are ruled.

## Forks — rulings (gblock, 2026-10-18)

- **F1 — Membership (§III.2): PENDING the human's review of Tables A–C** (gblock reviews before the gate).
- **F2 — Add both `unsourced_claim` and `invalid_inference`, `by_grade`: RULED, add both.**
- **F3 — Render all four fields (value, materiality, definition, nearest neighbour + distinguisher) in `mint --help` and, through `agentgen`, in every lens constitution: RULED, full.** The ≈ 13 KB per lens constitution (R5) is accepted.
- **F4 — `unclassified` rate per epoch now (D8); the cross-run rate goes with #1454: RULED.**
- Decided and disclosed (reversible, in scope): underscore spelling (D2); `material_default` as an
  enum-typed facet, `neighbor`/`distinguisher` as help-only options (D3, D4); `Mint.class_material`
  stays stored (D5); `unclassified` kept despite two same-sense hits (D6); the reason requirement as a
  database check (D7); `by_class` becomes a tally list and `class_repeat_rate` excludes the catch-all
  (D8); `class-registry-seed.md` stays as a design record.
