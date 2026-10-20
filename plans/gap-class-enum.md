# The gap class is a record enum (#1134)

Written against `be6c110d`; the census (§III.14) re-run at `61107ba4` and again at **`b283d1e8`** (`main`
after Slim-fast's #1481, the fuzz's #1495 and migrate's #1497; `EventSchema` 32; 28 archived runs), the
base every line number below is read at. Implements gblock's ruling of
2026-10-18 on #1134 (the issue's last comment); absorbs #1131 (definitions) and #1132 (truncated
registry sample). Replaces the superseded design that shipped the JSON registry inside the binary
(never committed as a plan).

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
staging, its generator and its harvest are deleted. The event-schema epoch moves to **E = 35** (the next free epoch at implementation; `main` is 34 after #1522), one
above `main`'s 32 (§II), and `migrate` carries every archived run onto the enum.

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
   via `agentgen`; a wrong `--class` word is refused with the near miss and the 30 words only (§V.3).
4. `inquest telemetry --help`'s OUTPUT tree and an emitted line agree at every depth, `by_severity`
   included, and every enum-valued leaf is a word (§V.2 `TestTelemetryHelpTreeMatchesAnEmittedLine`).
5. Every fixture keeps the materiality it had: a placeholder class is a `by_grade` value (§III.10).
6. All 28 archived run tarballs (27 with mints) migrate with zero refusals; a named test on
   `2026-10-18_is-91-prime-m28` and `2026-08-23_sleeper-service-plan` asserts every mint lands on a
   `GapClass` value, every recorded `class_material` is unchanged, no `class_new` event and no
   `records/class-registry.json` survives, and the manifest lists each class translation; on m28,
   all 20 stored plans migrate, the 9 naming events the dropped `class_new` rows renumber carry the
   migrated numbers and a re-computed key (§III.8a); every other archive's stored plans (`final`,
   `n1`, `m29`) migrate byte-identical (§III.8a); on `2026-09-11_is-91-prime-b6`, the pre-field
   mint `G2` takes `always` from its nearest value and the PASS records it as admitted (§III.8, F9)
   (§V.5).
7. A universe smoke with no `feov-memory` staged mints ≥1 gap, and one mint uses `unclassified`
   (§V.7).
8. `git grep -n -E 'class-registry\.json|classgen|ShippedClasses|ShippedMaterialDefaults|StageClassRegistry|ValidateClassRegistry|loadRegistry|StageForRun|ClassRoster|HarvestClasses|ClassNew|class_new|\bclass (new|list)\b' -- plugins scripts .github docs ':!scripts/rulesweep' ':!plugins/frank-exchange-of-views/tools/internal/record/migrate' ':!*.pb.go' | grep -v 'reserved "'`
   returns nothing, and `git grep -n -E 'ShippedMaterialDefaults|shippedDefault|byGradeFillWhy|translateStagedRegistry|StageForRun|ClassNew\b|planRenumbered|revisionEntry|TestARefusalNamesToNotTheBuild' -- plugins/frank-exchange-of-views/tools/internal/record/migrate`
   and the rename's two greps — `git grep -n -E '\bkeyed\(|debatejs\.Keyed|\bKeyed\(' -- plugins/frank-exchange-of-views/tools plugins/frank-exchange-of-views/tests`
   (every call and declaration, Go and the JS harness) and `git grep -n -E 'const keyed\b|\bkeyed,|// [kK]eyed is' -- plugins/frank-exchange-of-views/tests/simulator plugins/frank-exchange-of-views/tools/internal/record/planstore.go plugins/frank-exchange-of-views/tools/internal/debatejs`
   (the JS export, its import list and the three doc comments; scoped by path because `keyed,` is a
   local variable in `cli/gapanchor_test.go:430` and a word in `cli/stategraph_test.go:59`) —
   return nothing (§V.8). `\bKeyed\(` does not match `NewKeyed(`. `\b` keeps "by-class listing" (`debate.test.mjs:397`, `record/available.go:122`)
   out; `scripts/rulesweep` is excluded because its "class list" (`sweep_test.go:94`) is the PROTOCOL
   class registry's, another concept (§III.14 C1/C2), and nothing else in that directory names the gap class.
   `migrate/` is excluded from the first grep because it MUST keep the archived words — the `class_new`
   `Entries()` row, the `drops{"class_new": …}` column, `copyChannels`' skip of `records/class-registry.json`
   — and gets the second, narrower grep; `record.proto`'s `reserved "class_new"` / `reserved "EVENT_TYPE_CLASS_NEW"`
   lines and the descriptor `record.pb.go` embeds carry the reserved names by design.
9. `go -C scripts run ./check` green at `-count=1` with the fail count read; release-gate fuzz green.

**Value case.** Against the superseded design (embed the JSON, keep coining), the enum: removes the
class of defect rather than relocating it — a vocabulary the binary owns cannot be missing from a
consumer's tree; puts definitions where a lens already reads closed sets (39 of 41 registry rows had
none, and three runs logged that `class list` shows none); replaces a parallel validator
(`loadRegistryIn`), a generator, two generated files, a CI step, a per-run channel, a verb group, an
event type and a capture harvest with one proto enum the existing machinery (`EnumFields`, enum
tables, `enumhelp`, `migrate`'s value translation) already handles. The measured cost of what is
removed: 34 coinings across 14 runs, mostly single-use near-synonyms (inquiry-* in 5 runs, citation-*
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
  `class_repeat_rate`, the ledger row at `:326`), `report/assemble.go:863`, `graph/graph.go:223`,
  `record/viewjson.go` (reads the SQL column, a word), `capture.go:2067`. `record/motion.go:353` reads
  the PETITION's class (`PetitionClass`), not a mint's.
- **Event-schema epoch** `record.EventSchema` = 32 on `main` at `b283d1e8` (Slim-fast's #1481 took
  32), generated from `plugins/frank-exchange-of-views/requirements.json` `eventSchema` by
  `scripts/schemagen`. Changing `Mint.class`'s type, deleting `EVENT_TYPE_CLASS_NEW`/`ClassNew`, and
  deleting Mint fields 4–7 change the event shape: **epoch E = 35**.
- **migrate** (`internal/record/migrate`): re-drives an archived record's events through the live
  write path. Old words translate through `Entries()` (a `Registry` of event words → `Translate`, or
  `Loss`); an archived enum word resolves through `valueRenames[<enum full name>]` then the schema's
  own spelling with `SameWord` (`enumValue`/`enumNumberOf`). A translation that produces no event is
  precedent (`revision`). Fills of a value the source predates are `StatedFill`s on the manifest. The
  staged registry is today a copied run channel plus `translateStagedRegistry`.
- **Archived data** (measured, `~/.claude/scratch/relocations/classuse/scripts/{m,cm,coined}.py` over
  the 27 run tarballs at `61107ba4`, and the 28th, `2026-10-18_is-91-prime-m29`, archived since —
  `~/.claude/scratch/relocations/audit5/rev5/`): 224 mints; 26 of 41 registry classes used; 39 mints under 31
  coined slugs (34 coined in all — m29 coined `summary-scope-omission`, one mint). 16 runs (2026-08-22 …
  2026-09-11) record **no** `class_material` on any of their 123 mints and their staged registries carry no
  `material_default`; the 11 runs with mints from 2026-10-09 record it on every mint. Today those 123 take `ShippedMaterialDefaults[slug]`, else `by_grade` (`migrate/replay.go:56–62`
  `shippedDefault`), so every coined slug lands `by_grade`; under F9 each takes its nearest value's default, and
  17 of them, in 8 runs, move `by_grade` → `always` (§III.8, measured from `coined.jsonl` against Tables A and C;
  `~/.claude/scratch/relocations/audit3/rev3/flip.py`).
- **Stored plans and event ids** (measured, `~/.claude/scratch/relocations/audit3/rev3/`, m29 in
  `~/.claude/scratch/relocations/audit5/rev5/`): a record stores plans from epoch 31; of the 28 archives, 4 do
  (`final` 3, `n1` 3, `m28` 20, `m29` 17), and m28 and m29 also hold `class_new` events (564, 572; 566). 9 of
  m28's 20 plans name an id ≥ 564 — in `head`, a `stale_areas[].pin`, and the `why` lines that print them
  (`head 579`, `pin 507`, `changed at edit 579 … last check is event 390`). m29's two plans after 566 (events
  581, 587) name no id above it, so none of its plans changes. m23's single `class_new` (373) precedes no
  stored plan.
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
| D1 | Proto enum or string? | (a) keep `string` + a Go table + a Go validator; (b) **proto enum `GapClass`**; (c) a separate SQL vocabulary table outside the proto | **(b).** It is the pattern every other closed set uses, so the vocabulary table, the FK refusal, `EnumFields` validation, `--help` and `migrate`'s word translation come with it; (a) rebuilds `loadRegistryIn` in Go — the second validator the ruling deletes; (c) is a hand-kept second schema. The type change makes the compiler the census of every TYPED reader (a `string` comparison, map key or assignment) and of NO untyped one: a value passed to an `any` parameter or a `fmt` verb compiles and renders wrong in silence. Two exist (§III.5): `report/assemble.go:863` (`grade(v any)` asserts `string`, so a `GapClass` prints `—`) and `graph/graph.go:223` (`%s` prints `GAP_CLASS_FALSE_UNIVERSAL`). Each converts with `recordpb.Word` and gets a test asserting the rendered class word. |
| D2 | Spelling | (a) **underscores** (`report_voice`), the schema's spelling; (b) hyphen join like the retired grade spelling | **(a).** `Spelling` derives underscores and the grade hyphen exception no longer exists; a seat that types `report-voice` gets the existing `SameWord` near-miss refusal, and migrate reads hyphenated archives through the same fold with no authored row. Disclosed: prose naming `report-voice` changes (§III.12). |
| D3 | Where materiality lives | (a) string facet holding a word; (b) **an enum-typed value option `(material_default)` of type `ClassMaterial`**; (c) a Go map | **(b).** `protoc` refuses a misspelt word at generation; the name is the existing concept's (`material_default` on registry rows and `ClassNew`), not a rival term. It is a FACET (the stamp reads it): `facets.go` gains a third facet kind (enum-valued) and `enumTable` emits `"material_default" TEXT NOT NULL REFERENCES "enum_class_material"("value")`. Partly-annotated is refused as for every facet, so a value cannot be added without answering it. |
| D4 | Neighbour and distinguisher | (a) facets (table columns); (b) **value options read by help and a test, not facets**; (c) folded into `(means)` prose | **(b).** The catch-all has no neighbour, and a facet cannot be declared on all-but-one value; nothing in the machinery switches on them (precedent: `ruled_by` is an option read directly, not a facet). `TestEveryGapClassNamesItsNeighbour` requires both on every value but `unclassified` and requires `neighbor` to be another value's word. (c) would compose a checkable reference into prose. Names are the existing ones (`--neighbor`, `--distinguisher`). |
| D5 | `Mint.class_material` | (a) **keep stored, stamped from the value's `material_default`**; (b) delete and derive by join on `enum_gap_class` | **(a).** Merges and drops move archived gaps to values with a different default (e.g. `audited-artifact-sibling-halo` `by_grade` → `verification_scope_blindspot` `always`; anything → `unclassified` `by_grade`); CLAUDE.md requires the migrated record to hold the equivalent state, and a stored stamp is the gate's materiality at the act. The stamp is a different fact from the class's current default, so this is not two writers of one fact. A mint whose source records no stamp (16 pre-field runs) has no act-time materiality to keep, and takes its translated value's default by the live rule (F9, §III.8). |
| D6 | Catch-all name | `unclassified` (ruled) | Grep found `Unclassified` in `diagnostics/survey.go` (`Traversal.Unclassified`, JSON `unclassified`: invocations the traversal could not place in a group) and `migrate/manifest.go` (`unclassified_tables`: old tables no rule could place). Same sense — "could not be placed" — in namespaces no reader joins with a gap's class. **Kept**; the telemetry key is `unclassified_rate`, which collides with nothing (0 hits). |
| D7 | "`unclassified` requires `--reason`" | (a) Go `if` in validate; (b) **a message `(check)` on `Mint`** | **(b)**, precedent `DocketRuling`'s remand-direction check: `"class" <> 'unclassified' OR coalesce(trim("mint_reason", …), '') <> ''`, `why:` "a gap no class fits says why in --reason: the nearest class and what it does not cover — the rate of these is how a release learns which class is missing". The database refuses it, so no route writes around it; the write's own refusal (`requireUnclassifiedReason`, which names the duty to the seat) trims the check's character set — space, tab, LF, CR — and no other, so the two refuse the same reasons (`TestTheWriteAndTheDatabaseAgreeOnABlankReason`). `--reason` (`mint_reason`) carries it, as ruled; its help line gains the duty. The check requires a non-empty reason and nothing more: that the reason names the nearest value is prose no write can refuse — **accepted residue (F5)**; no `--nearest` field. |
| D8 | Telemetry | by_class map vs repeated; where the rate goes; help/wire agreement | `NewMint.by_class` was a map "because the registry is extended at runtime by `class-new`" (its own comment) — the premise is gone, so it becomes `repeated ClassTally by_class` in schema order, the `by_severity` precedent (deterministic by construction). New `optional double unclassified_rate` (unset — the wire's `null` — when nothing was minted, as `class_repeat_rate`). `class_repeat_rate` counts an `unclassified` mint in the denominator and never as a repeat: "no class fits" recurring is not one kind recurring, and counting it would raise the bench's stopping signal on noise. Per-epoch only; a cross-run rate belongs to #1454's archive-derived scorecards (F4). **The wire is hand-written, not the proto:** help renders `jsonshape.Tree(view.TelemetryLineShape())` (`view.go:827`) over the hand-written `telemetryLineJSON`, filled by `telemetryWire`, and `by_severity` already disagrees — help says `[{grade,count}]`, `severityObject.MarshalJSON` (`view.go:864`, the tree's only custom marshaller) emits an object keyed by grade word. Options: (a) **both tallies become arrays of plain structs (`{grade,count}`, `{class,count}`) with word-valued fields, and `severityObject` is deleted** — the reflected type IS the marshalled shape, so tree and bytes cannot diverge; (b) keep the object and teach `jsonshape` an override interface — a second hand-written shape, the defect itself; (c) keep the object and hand-write the help — prose describing a struct. **(a)**, ruled with the test (F6): `TestTelemetryHelpTreeMatchesAnEmittedLine` (§III.9). An EMPTY tally is `[]`, never `null` (F10): on this wire `null` means "not measured" (`view.go:147–160`), and an epoch with no mints measured zero of each — today's `severityObject` writes `{}` for it. |
| D9 | How a lens sees the definitions | `mint --help`; the constitution; the refusal | **Help:** `mint --help`'s "Enumerated values" block (`enumhelp.Section`) gains `--class`, one entry per value: `word  [material_default] definition` and, indented, `nearest: <word> — <distinguisher>`. `agentgen` inlines it into every `red-lens-*.md` (the lens's manual), so a lens reads it at sitting start without a call. **Refusal (F7):** a wrong word is refused through `record.Refuse` with the near miss and the 30 words (`takes one of a\|b\|…`) and the field's `why`, and WITHOUT the `Menu` block — `Refuse` appends `Menu(vs)` only for an enum whose values carry no long form (`Material`/`Neighbor` empty), so every other enum's refusal is byte-identical. The two need not share a block. Tests: `TestMintHelpListsEveryGapClass` (rendered `mint --help`: every value's two lines) and `TestAWrongClassIsRefusedWithTheNearMissAndTheWordsOnly` (rendered refusal of `--class report-voice`: names `report_voice` as the punctuation near miss, holds each of the 30 words, holds no value's `(means)` and no `nearest:`); `--check-kind banana`'s refusal golden unchanged. Cost stated in F3. |
| D10 | A stored plan naming events a dropped event renumbers (F8) | (a) **translate inside migrate: every events.id the printed plan names — `head`, `stale_areas[].pin`, a `contradiction` blocker's `subject`, and the ids its `why` lines print — through `newID`, then re-key and re-print**; (b) re-compute the plan with the live verb at that place in the replay; (c) refuse, as today (`planRenumbered`); (d) keep a `class_new`-shaped event so nothing renumbers | **(a)**, ruled (F8), general over any event that writes none. (b) replaces the decision the chair relayed with what this binary's rules would decide, which can differ; (c) fails criterion 6 on m28; (d) is a live-record mechanism kept so something survives migration, which CLAUDE.md forbids. The `why` lines are prose the verb composed from those ids and from seat words (§III.8a's producer census), so the translation reads them through a closed table of the verb's id-introducing phrases, carries every seat-authored span as written (a `%q` span, and the remand's direction after its fixed prefix — F12), and makes every other miss LOUD (§III.8a). The key rule is exported from `record` rather than copied into migrate: one rule, one home. Its name is `Plan.WithKey()`, not `Keyed` — `record.Answer.Keyed` (`supersede.go:112`) already means "the key field a supersede row is for", another sense in the same package — and `debatejs.Keyed`, the same concept over a stub map, is renamed `debatejs.WithKey` with it, and the JS harness's `keyed` (`harness.mjs:92`) `withKey`, so one term names a plan carrying its key in both languages (`withKey`: 0 hits). |

### III.2 Membership (fork F1 — ruled as proposed)

Sources: the 41 registry rows (2 with definitions: `structure-noncompliance`, `self-defeating-mitigation`),
`feov-memory/class-registry-seed.md` (original intent, tie-break triplets), every archived mint's
`problem` (`~/.claude/scratch/relocations/classuse/samples.txt`) and every coining's definition
(`coined.jsonl`). "Uses" = mints in the 27 archived runs that mint (m29's 14 included). ⚑ = a call I am least sure of.

**Table A — the values (30).** Order is the enum's, grouped so neighbours sit together.

| # | value | md | uses (→ after migrate) | definition (`(means)`) | nearest | distinguisher |
|---|---|---|---|---|---|---|
| 1 | `cross_section_contradiction` | always | 24 (→27) | two passages of the report state things that cannot both be true — figures, verdicts, grades, or the reasons given for them | `claim_contradicts_own_record` | are both statements in the report's text, or is one of them the run's record — an inquiry's fate, an edit, a proof's output? |
| 2 | `claim_contradicts_own_record` | always | 24 (→30) | the report and the run's own record disagree: the report states what the record contradicts — an inquiry's fate, method or reason, an edit, a proof's result — or a record entry contradicts what the report shows was done | `cross_section_contradiction` | (as row 1) |
| 3 | `reader_channel_mismatch` ⚑ | by_grade | 3 (→5) | a fact the reader needs is held — in the record or elsewhere in the report — but not where the reader who needs it reads, so a reader acting on that section acts without it | `claim_contradicts_own_record` | do the two places disagree, or does one simply lack what the other holds? |
| 4 | `incomplete_repair_propagation` | always | 2 | a repair fixed a defect at some of its sites and left it standing at others — a sibling sentence, row or section still carries what the repair removed | `cross_section_contradiction` | does the disagreement exist because a repair reached only part of one defect's sites, or do two passages disagree with no repair behind them? |
| 5 | `missing_root_invariant` ⚑ | by_grade | 1 | one fact is restated in several places with nothing tying the copies to a single source, so they have drifted or can drift unnoticed | `incomplete_repair_propagation` | is the defect that no single source governs the copies, or that one repair missed some of them? |
| 6 | `report_voice` | by_grade | 35 (→36) | report prose speaks about the run that made it rather than the subject — process voice, a claim wearing the seat that produced it, the report's own draft history, or the machinery that checked it (the seed's promoted definition) | `claim_contradicts_own_record` | does the fix correct a FACT the report gets wrong, or the AUDIENCE the sentence is written for, whether or not the fact is true? |
| 7 | `structure_noncompliance` | never | 4 | (registry's promoted definition, unchanged) | `risk_coverage_omission` | that is a risk the report does not grade; this is material the report holds with no order a reader can follow it by (neighbour moves: `risk-cross-term-ungraded` merges into `risk_coverage_omission`) |
| 8 | `false_universal` | always | 17 | a claim stated universally — every, none, always, the only — that a case the evidence reaches, or one it never tested, contradicts | `exhaustive_sweep_omits_case` | is it a rule about the subject stated too broadly, or a list or search presented as complete that leaves a member out? |
| 9 | `exhaustive_sweep_omits_case` | always | 4 | a list, survey, search or sweep the report presents as complete omits a member it should hold | `false_universal` | (as row 8) |
| 10 | `risk_coverage_omission` | by_grade | 12 (→14) | the report's account of risk leaves out a risk, or a combination of risks, that its own analysis or its subject makes live | `exhaustive_sweep_omits_case` | is the missing member a risk the case against owes a grade, or any other member of a list presented as complete? |
| 11 | `derivation_status_overclaim` | always | 13 | a claim carries more certainty or status than how it was established supports — inferred stated as verified, hedged in one place and flat in another, assumed stated as shown | `verification_scope_blindspot` | is the claim's STATUS overstated, or is real evidence spent on a claim broader than what it covers? |
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

**Table C — the 34 coined slugs.** (mints; run)

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
| model-scope-overclaim (1; sleeper) | `within_source_condition_misattribution` | | summary-scope-omission (1; m29) | `reader_channel_mismatch` |
| assertion-without-inquiry (1; sleeper) | `unsourced_claim` | | | |
| citation-support-mismatch (1; sleeper) | `citation_figure_misattribution` | | | |

**The calls marked ⚑** (ruled as proposed under F1):

1. **The inquiry-* family → `claim_contradicts_own_record`** (4 slugs, 5 mints). The alternative is a
   new value `inquiry_record_mismatch`; I merged because the distinguisher the coiners wrote ("the
   line-of-inquiry record vs the report") is exactly this class's distinguisher from
   `cross_section_contradiction` — the inquiry record is part of the run's record.
2. **Two NEW values, `unsourced_claim` and `invalid_inference`** (F2). The registry has no class for
   "no source at all" (6 coined mints across 4 runs land there) or for a non sequitur (m28 coined it
   twice in one run); and both are `by_grade`, though the registry's own rule is that claim and citation
   classes are `always`. `by_grade` because an unsourced well-known fact, or a bad reason beside a true
   verdict, need not change what a reader concludes.
3. **`reader_channel_mismatch` redefined** (its original meaning is unrecorded; its mints in m28 and
   m29 fit the definition above) and receiving `unsurfaced-inquiry-finding` and m29's
   `summary-scope-omission` (round 5, not covered by F1 — disclosed below). The coiner named
   `risk-coverage-omission` as its neighbour; its own definition ("a summary section states a conclusion
   … without the scope, or the strongest limit, that the body gives it") is a fact the body holds and the
   summary's reader does not get, which is this value's definition, and its distinguisher ("is it a risk
   the reader is never shown at all") rules `risk_coverage_omission` out. The stamp is recorded
   (`by_grade`, D5), so the call moves no materiality.
4. **Zero-use classes kept** — `unhandled_degenerate_case`, `vote_laundering`, `unverified_composition`
   (one misused mint), `missing_root_invariant`. Each has a writable distinguisher; all four come from
   the seed's design-review runs, which the archive under-samples. Dropping them is the cheaper
   vocabulary.
5. **`metric_conflation` broadened to categories** (absorbs `risk-taxonomy-conflation`).

### III.3 Directory tree

```
plugins/frank-exchange-of-views/
  requirements.json                                   [MODIFY] eventSchema 34 → 35
  docs/seat-command-triggers.md                       [MODIFY] rows `lens class list`, `lens class new` deleted
  docs/seat-surface-naming.md                         [MODIFY] `class new` dropped from the merge row
  docs/vocabulary.md                                  [MODIFY] regenerated (vocabdoc)
  agents/red-lens-*.md (7), lead-judge.md, red-chair.md [MODIFY] regenerated (agentgen)
  skills/adversarial-audit/SKILL.md                   [MODIFY] regenerated (agentgen)
  skills/research-protocol/scripts/debate.js          [MODIFY] authored lens prompt line (~989)
  tests/simulator/debate.test.mjs                     [MODIFY] `class new` leaves the command-path regex (:517–518); :408's prompt phrase; :1033 comment; `keyed` → `withKey` (:4 import, :1089 call)
  tests/simulator/harness.mjs                         [MODIFY] `keyed` → `withKey` (:89 comment, :92 export, :93 call) — the JS stub of the same concept as `debatejs.Keyed` (D10)
  tests/simulator/testdata/prompt-red-lens-*.golden (4) [MODIFY] regenerated
  tools/internal/record/recordpb/record.proto         [MODIFY] GapClass, options, Mint, Event, NewMint
  tools/internal/record/recordpb/{record.pb.go,record.proto.sha256} [MODIFY] protogen
  tools/internal/record/recordpb/schema_gen.go        [MODIFY] schemagen (epoch 35)
  tools/internal/record/recordpb/facets.go            [MODIFY] enum-valued facet kind
  tools/internal/record/recordpb/gapclass.go          [NEW]    MaterialDefault(GapClass), Neighbor, Distinguisher
  tools/internal/record/recordpb/gapclass_test.go     [NEW]
  tools/internal/record/recordpb/testdata/payload-keys.txt [MODIFY]
  tools/internal/record/recordsql/schema.go           [MODIFY] enum-valued facet column; referenced vocabulary first
  tools/internal/record/recordsql/schema_test.go      [MODIFY] TestAVocabularyFollowsTheVocabulariesItReferences
  tools/internal/record/recordsql/testdata/schema.sql [MODIFY] regenerated
  tools/internal/record/{replay.go,record.go,enums.go,enumvalue.go,enumof.go,writeview.go,viewjson.go} [MODIFY]
  tools/internal/record/{planstore.go,planstore_test.go,dispatch.go,dispatchnext.go} [MODIFY] `keyed` exported as the method `Plan.WithKey` (migrate re-keys a translated plan, D10); `dispatch.go` exports `RemandDirectionPrefix` and `:365` formats with it (F14)
  tools/internal/record/remand_test.go                [MODIFY] TestARemandLineCarriesItsDirectionAsWritten (§III.8a)
  tools/internal/record/{retirement.go,dispatch.go}    [MODIFY] the lens fold's reason is `lensFold.readiness`, its unsat-sitting clause `unsatSitting`, a debt's blocker `owedBlocker`, a contradiction's readiness `contradictionReadiness` — each producer of an id-bearing `why` line one function the guard renders (§III.8a)
  tools/internal/record/{export_test.go,planidphrases_test.go} [MODIFY / NEW] TestEveryIDTheVerbPrintsIsTranslated, TestTheOpenedAtLineCrossesARenumbering (§III.8a)
  tools/internal/debatejs/{debatejs.go,petition_test.go}, seatprobe/production.go, integration/surface/fragmentdispatch_test.go, releasegate/fuzz/{fuzz_test.go,termination_test.go} [MODIFY] `debatejs.Keyed` → `debatejs.WithKey` (D10)
  tools/internal/record/recordtest/recordtest.go      [MODIFY] ClassFor(ClassMaterial); coining default deleted
  tools/internal/report/assemble.go (+ test)          [MODIFY] class word through recordpb.Word
  tools/internal/graph/graph.go (+ test)              [MODIFY] class word through recordpb.Word
  tools/internal/record/classroster.go (+_test)       [DELETE]
  tools/internal/record/shippedclasses_gen.go         [DELETE]
  tools/internal/record/recordtest/classes_gen.go     [DELETE]
  tools/internal/record/{fixtureclasses_test.go,classmaterial_test.go} [DELETE / MODIFY]
  tools/internal/record/migrate/gapclass.go           [NEW]    the archived-slug → GapClass table
  tools/internal/record/migrate/gapclass_test.go      [NEW]    (replaces classmaterial_test.go)
  tools/internal/record/migrate/{migrate.go,replay.go,entries.go,remap.go,manifest.go,valuerenames.go} [MODIFY]
  tools/internal/record/migrate/planids.go            [NEW]    a stored plan's event ids through newID (`TranslatePlanIDs`, exported for `record`'s guard) (§III.8a)
  tools/internal/record/migrate/planrenumbered_internal_test.go → planids_internal_test.go [MODIFY] (+ TestMigrateKnowsTheVerbsRemandPrefix, F14)
  tools/internal/record/migrate/{storedplan_test.go,benchoccasion_test.go,publish_test.go} [MODIFY] storedplan's prose and benchoccasion's byte-for-byte census (C15); TestARefusalNamesToNotTheBuild deleted (F11)
  tools/internal/record/migrate/testdata/prose-rewrites.golden [MODIFY] regenerated: sleeper's one `ClassNew.distinguisher` row leaves (a `class_new` writes no event)
  tools/internal/cli/lens/class.go                    [DELETE]
  tools/internal/cli/seat/help/{list,new}.md          [DELETE] class list / class new pages (help_test.go's bijection fails an orphan)
  tools/internal/cli/seat/help/mint.md                [MODIFY] the reason's `unclassified` duty
  tools/internal/cli/seat/help.go                     [MODIFY] comment example
  tools/internal/cli/lens/mint.go                     [MODIFY]
  tools/internal/cli/enumhelp/enumhelp.go             [MODIFY] render material/neighbour/distinguisher
  tools/internal/flags/names.go                       [MODIFY] Definition, Neighbor, Distinguisher, MaterialDefault removed; `Class` comment
  tools/internal/setup/{setup.go,run.go,run_test.go}  [MODIFY]
  tools/internal/capture/capture.go                   [MODIFY] HarvestClasses + call deleted
  tools/internal/capture/classharvest_test.go         [DELETE]
  tools/internal/view/view.go (+_test)                [MODIFY] by_class tally, unclassified_rate, telemetry wire, ledger class
  tools/internal/cli/telemetryshape_test.go           [NEW]    TestTelemetryHelpTreeMatchesAnEmittedLine
  tools/internal/seatprobe/{build.go,boards.go,seatprobe.go} [MODIFY]; boardclasses_test.go [DELETE] (C16)
  tools/internal/terms/terms.json                     [MODIFY] `cross-section-contradiction` → `cross_section_contradiction`
  tools/internal/difftest/testdata/*.golden (18), class_registry.golden [MODIFY / DELETE scenario]
  tools/releasegate/fuzz/{fuzz_test.go,registerbeforeappend_test.go} [MODIFY]
  tools/integration/surface/{chairrefusal,promptverbs,vocabulary_prose}_test.go [MODIFY]
  ~110 test files minting placeholder classes        [MODIFY] mechanical (§III.10)
  test comments and fixtures naming the registry      [MODIFY] the §III.5 widened sweep's list (C16)
feov-memory/class-registry.json                       [DELETE]
feov-memory/class-registry-seed.md                    [MODIFY] one line: the vocabulary now lives in record.proto (seed kept as design record)
law/proposed/class-*.md (14)                          [DELETE] — the directory's only files; law/proposed/ leaves the tree
scripts/classgen/                                     [DELETE]
scripts/check/gates.go                                [MODIFY] classgen gate removed
scripts/agentgen/src/{fragments/lens-audit-and-mint.md,red-lens-voice.md,skills/adversarial-audit/SKILL.md} [MODIFY]
scripts/universe.sh                                   [MODIFY] feov-memory staging block removed
scripts/universe-stream.py                            [MODIFY] `class` leaves INTERESTING (the verb is gone)
.github/workflows/hooks.yml                           [MODIFY] classgen step removed
docs/local-development.md                             [MODIFY] `report-voice` → `report_voice`
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
  same declared/undeclared split and refusals as `Facet`/`Number`. `material_default` lives in its own
  map `enumFacets = map[string]protoreflect.ExtensionType{"material_default": E_MaterialDefault}`, never
  in `facets`: every caller that ranges over `facets` and calls `Facet` gets the option asserted `.(bool)`
  (`facets.go:69–89`; `numericFacets` is the flag that keeps `mass` out of that path), and a third
  kind behind a second flag is one more caller to forget. `FacetNames()` returns the union (the
  refusal's "have …" list and `enumTable`'s column walk read it), `Facet`/`Number` refuse an
  `enumFacets` name with "read it with EnumFacet"; and `EnumFacetTargets(ed)`, the
  enums an enum's enum-valued facets reference, which `enumTables` visits first.
- `recordpb/gapclass.go` [NEW]: `MaterialDefault(GapClass) ClassMaterial` (via `EnumFacet`),
  `Neighbor(GapClass) (GapClass, bool)`, `Distinguisher(GapClass) string`. `gapclass_test.go`:
  `TestEveryGapClassDeclaresItsMaterialDefault`, `TestEveryGapClassNamesItsNeighbour` (every value but
  `unclassified` declares both options; `neighbor` resolves by `BySpelling` to a different value),
  `TestTheCatchAllIsByGradeAndHasNoNeighbour`.
- `recordsql/schema.go` `enumTable`: emits an enum-valued facet as `TEXT NOT NULL REFERENCES
  "enum_<enum>"("value")`.
- **Vocabulary order.** `enumTables` (`schema.go:753`) emits tables in first-reach order of a field
  walk, and each table's `INSERT`s follow its `CREATE` with foreign keys on (`store.go:80`). The walk
  reaches `Mint.class` (field 3) before `Mint.class_material` (field 25), and once `ClassNew` is
  deleted nothing reaches `ClassMaterial` earlier — so first-reach order would emit
  `enum_gap_class` and insert rows referencing `enum_class_material` before that table exists. The
  walk's append becomes a dependency-first visit: before appending an enum, it appends every enum
  its enum-valued facets reference (`recordpb.EnumFacetTargets(ed)`, read off `facets.go`'s table).
  The order then holds whatever the field numbering. Wrong order fails loudly — every record open
  fails at the `INSERT` with "no such table" — and is pinned by
  `TestAVocabularyFollowsTheVocabulariesItReferences` (`recordsql/schema_test.go`: in the built DDL,
  every `REFERENCES "enum_x"` inside `CREATE TABLE "enum_y"` has `enum_x`'s `CREATE` earlier). The
  `schema.sql` golden shows the position (`enum_class_material` ahead of `enum_gap_class`); it is
  regenerated and every changed line read.
- `requirements.json` `eventSchema` 35; `go -C scripts run ./schemagen`; `go -C scripts run ./protogen`.

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
- `enumof.go:111` `ClassMaterialOf` comment rewritten; add `GapClassOf(word)` beside it (the write
  path's word → value, as `GradeOf`).
- `writeview.go:49` `registry()` deleted; `runFiles` loses the records-dir read if nothing else uses it.
- `classroster.go`, `classroster_test.go`, `shippedclasses_gen.go`, `recordtest/classes_gen.go`,
  `fixtureclasses_test.go` deleted. `recordtest/recordtest.go:127–130` (seeded coining default)
  deleted; `:121–126` (a seeded mint without a class material gets `by_grade`) STAYS, its comment
  rewritten to "`by_grade` is the default of every placeholder class (§III.10), so a seeded fixture
  and a written one agree". New `recordtest.ClassFor(cm ClassMaterial) GapClass` returns the
  fixture value for each materiality (`spec_underspecification`, `false_universal`,
  `structure_noncompliance`), with `TestClassForHasTheMaterialityItNames` (`MaterialDefault(ClassFor(cm)) == cm`
  for all three).
- **The mint-class readers** (census C13, §III.14): `git grep -n -E 'GetClass\(\)|\b(m|p|mint)\.Class\b'`
  over FEOV's non-test, non-generated Go (`-- 'plugins/frank-exchange-of-views/*.go' ':!*_test.go' ':!*.pb.go'`;
  unscoped it also prints prosthetic-conscience's `secrets.go`), each hit read for which class it is.
  - Typed — the compiler refuses them once the type changes: `view.go:507,517,527` (map keys) read
    `recordpb.Word(...)`; `view.go:326` (`class, check = m.Class, m.AcceptanceCheck` into
    `undefStr(*string)`) becomes `undefStr` of the word when `m.Class != nil`, so an absent class still
    prints `undefined`; `capture.go:2067`, `mint.go:93,95` and `replay.go:537–697` are deleted or
    rewritten (§III.5, §III.6, §III.11).
  - Untyped — compile and render wrong (D1): `report/assemble.go:863` passes
    `recordpb.Word(g.Mint.GetClass())` to `grade`; `graph/graph.go:223` formats
    `recordpb.Word(g.Mint.GetClass())`. Tests driven through the shipped renderers:
    `TestTheOpenGapRowNamesItsClassWord` (`report`: assemble a board holding one open
    `false_universal` gap; the row carries `class false_universal`, not `class —`),
    `TestTheGapNodeNamesItsClassWord` (`graph`: the label carries `false_universal` and no `GAP_CLASS_`),
    and `TestTheLedgerRowNamesItsClassWord` (`view`: `class false_universal`; a mint with no class
    prints `class undefined`). Deleting the `recordpb.Word` at each site fails its test.
  - Not a mint's class: `record/motion.go:353` (`fil.Petition.GetClass()`, a `PetitionClass`) — unchanged;
    `seatprobe/build.go:260` (`"--class", m.Class` of a probe board's petition motion) — unchanged.
  - A mint's class as a flag word, not a typed read: `seatprobe/build.go:166` (`"--class", g.Class`,
    the string field of `boards.go`'s gaps passed to `mint`) — found by the wider
    `git grep -n -E '\.Class\b'` (audit4's `dotclass.txt`; the C13 pattern names `m|p|mint` only);
    its words are respelled under §III.10 rule 4, and the probe's own mint refuses one that is not a value.
  - `record/viewjson.go` reads the SQL column, which holds the word; unchanged. JSON keeps the key
    `class` holding the word.
- **Comments that describe the registry or coining**, rewritten present-tense to the enum:
  `record.go:743` ("class registry" → "the gap class's enum check"), `record.go:788` ("an unstaged
  registry" dropped), `report/assemble.go:854` ("a registry SLUG" → "a `GapClass` word"),
  `view/view.go:504` ("validated against the registry" → "a closed enum"), `cli/seat/help.go:22`
  (the example "coin the class first" → "name the nearest class when none fits"),
  `seatprobe/build.go:127,135,149` (the staging block goes; the comment with it) and `:419`
  (`BoardClasses` are `GapClass` values), `seatprobe/seatprobe.go:81` (the two-word subgroup example
  becomes `blue inquiry propose` alone) and `:150` (the `class-new` entry goes), `flags/names.go:158–163`
  (`Class`: "a gap's class from the closed `GapClass` set, a petition's from a fixed four"), and
  `record.proto:334–335` (§III.4). The sweep that finds them, re-run after the change (§V.8):
  `git grep -n -i -E 'registry|coin(ed|ing|s)?\b' -- plugins/frank-exchange-of-views/tools ':!*_test.go' ':!*.pb.go' ':!*/internal/terms/*'`
  prints only the terms registry (`cli/manual.go:173`, `flags/names.go:233`, `recordpb/descriptions.go`,
  `record.proto:73–76`), `migrate`'s translation `Registry` (`migrate/*`, `cli/migrate.go`, `flags/names.go:308`),
  `tessocr/testdata` ("AWS Registry"), `record.proto:1809` ("coin" flip), `migrate/testdata/prose-rewrites.golden`
  (archived seat prose naming the sleeper run's "proof registry", carried as written), and the tracked
  `tools/research/` fixture.
- **The same sweep over tests and non-Go carriers.** The non-test sweep above cannot see a test comment,
  a `.mjs` comment or a Python verb list, and five such carriers still spoke the deleted surface after
  round 2 (`universe-stream.py:25`, `debate.test.mjs:1033`, `boardbuild_test.go:47`,
  `m23contradiction_test.go:100`, `reopenedm19_test.go:44`). Two commands, each hit read for which
  concept it names (§III.14 C16 holds the classified list and the replacement for every hit that changes):
  `git grep -n -i -E 'class registry|gap-class registry|\bclass\b.*verb|"class"' -- scripts '*.py' '*.mjs' '*_test.go'`
  and `git grep -n -i -E 'registry|\bcoin(ed|ing|s)?\b' -- 'plugins/frank-exchange-of-views/tools/**_test.go' plugins/frank-exchange-of-views/tests scripts '*.py' '*.mjs' ':!scripts/rulesweep' ':!scripts/classgen'`.
  After the change both print only C16's no-change hits (another concept: petition `class`, the
  protocol, terms and migrate registries, the regenerated prompt goldens' other lines).
- `castcoverage.go:20`, `duplicatescreen.go:38`, `record.go:828,1414`: comments respelled to the
  underscore word where they name a value.

### III.6 The lens surface [MODIFY]/[DELETE] (contract: CLI flag set — census C4, C5, C8)

- `cli/lens/class.go` deleted (the `class` group, `class list`, `class new`), and with it the two help
  documents it claims, `cli/seat/help/list.md` and `cli/seat/help/new.md`: `help_test.go:73–92`
  fails a document no verb claims. (`list`/`new` are claimed by no other verb — `git grep -n
  'NewRead("list"\|New("new"'` hits only `class.go`; re-checked at implementation by that test.)
- `cli/seat/help/mint.md`: "The run's part of the finding goes in your reason for the gap, which the
  report never prints and nothing checks." → "The run's part of the finding goes in your reason for
  the gap, which the report never prints. Under `unclassified` the reason is required — name the
  nearest class and what it does not cover." (the database refuses an empty one, D7).
- `cli/lens/mint.go` (lines at `b283d1e8`): `--class` registered through `enumhelp.Flag(c, flags.Class,
  record.MustEnum("mint", "class"), "REQUIRED — the KIND of defect, from the values below; its
  material default is recorded with the gap. `unclassified` only when none fits, and then --reason
  names the nearest value and what it does not cover")`. `p.ClassNew` (`:95`) and the coining comment
  (`:88–92`, `:258–261`) deleted. `seat.Supplies(c, "class_material", "stamped from the class's
  material default")`. `--reason`'s help on `mint` gains the `unclassified` duty.
- `cli/enumhelp/enumhelp.go` `Section`: a value with `Material` renders `word  [material] means`, and
  with `Neighbor` a second indented line `nearest: <word> — <distinguisher>`. Other enums render
  byte-identically (goldens prove it).
- `record/enumvalue.go` `Refuse` (D9, F7): unchanged sentence — near miss, `takes one of <30 words>`,
  the `why` — and `Menu(vs)` appended only when no value carries `Material` or `Neighbor`. Tests
  `TestMintHelpListsEveryGapClass` and `TestAWrongClassIsRefusedWithTheNearMissAndTheWordsOnly`
  (`cli/`, driven through `run(t, …)` as a seat); appending the `Menu` unconditionally fails the
  second, dropping the long form from `Section` fails the first.
- `flags/names.go:166–168` `Definition`, `Neighbor`, `Distinguisher` and `MaterialDefault` deleted
  (each has no other user — re-check with the compiler).
- `cli/` tests: `unknownclass_test.go` rewritten to the enum refusal (a near-miss `report-voice` names
  `report_voice`); `verbguardrace_test.go:225–242` (coin race) deleted; `cli_test.go:825–866`
  (coining) deleted; `badvalue_test.go:115`, `help_test.go:149`, `requiredmarker_test.go:96`,
  `referencechecks_test.go:25,45–60,119–165,231–234,265` (the `--class` row checks against the enum;
  `stageClassRegistry` and `TestAnUnreadableClassRegistryIsRefusedRatherThanIgnored` deleted),
  `refusalshape_test.go:230` (its row names the deleted `replay.go:696`), `projectionshape_test.go:288`
  updated to the new surface; `fixtureclasses_test.go` deleted (fixtures mint enum words, §III.10) and
  the `newRun` it defines moves, unstaged, into `cli_test.go`, which its callers already share (C16). New
  `TestUnclassifiedRequiresAReason` (refused without `--reason`, recorded with it, `class_material` =
  `by_grade`); new `TestAMintTakesItsClassMaterialDefault` (one `mint` per materiality through the
  verb — `false_universal`, `structure_noncompliance`, `spec_underspecification` — each stored with
  its value's default); new `TestALowGradePlaceholderMintDoesNotHoldTheGate` (a lens mints
  `spec_underspecification` at `low` through the verb; the chair's PASS is accepted).

### III.7 setup [MODIFY]

- `setup.go:366–431`: `ValidateClassRegistry`, `StageClassRegistry` and their comments deleted.
- `run.go:230–237`: the pre-state registry check deleted. `run.go:280` staging and `:341–346` summary
  lines deleted — there is no class line in setup's output.
- `run_test.go:260–341`: `TestClassRegistryIsStagedIntoTheRun`, `TestSetupRefusesRegistryWithoutMaterialDefault`,
  `TestAbsentRegistryIsAnnounced` deleted; new `TestSetupReadsNoFeovMemory` — two `runCfg` fixtures,
  one with no `feov-memory/` and one with a garbage `feov-memory/class-registry.json`, each run
  through `Run(cfg, &out, &errb)`, exit 0 both. Stdout differs by the fixtures' own paths (`run-setup:
  <runDir>`, `record binary:`, `run-live marker:`, the wrapper `binDir`). `runCfg` roots all of them
  in one `t.TempDir()` (`Cwd` = `Home` = it; `RunDir` and `BinDir` beneath it), so each stdout is
  normalised by `strings.ReplaceAll(out, cfg.Cwd, "<cwd>")` and the two compared byte for byte;
  stderr likewise. The records directories compare as sorted lists of names relative to
  `RunDir/records` (no sizes, no times). The billing line reads the same environment in both. The test fails if setup reads anything in `feov-memory/` that changes its
  output or its records.
- `unresolvablerun_test.go:27` comment only. After the change `git grep -n feov-memory --
  plugins/frank-exchange-of-views/tools/internal/setup` prints only `run.go:362` (the scorecard
  comment, #1454's).

### III.8 migrate — one translation [NEW]/[MODIFY] (census C1, C2, C4)

The translation, in `migrate/gapclass.go`, registered in `valueRenames["feov.record.v1.GapClass"]`:

- **A kept registry slug** resolves with no row: `enumNumberOf`'s `SameWord` fold reads
  `report-voice` as `report_voice`.
- **A merged or dropped registry slug, and every coined slug** has an authored row (Table B + C:
  14 + 32 rows; `structure-noncompliance` and `invalid-inference` resolve by fold and have none). `TestEveryArchivedClassLandsOnAValue`:
  every key is NOT a current word under `SameWord` (else the row is dead, and the manifest would list a
  fold), every value IS one.
- **`mint`** gets an `Entries()` row, `mapped("mint", nil, drops{"class_new": …, "definition": …,
  "neighbor": …, "distinguisher": …}, nil)`: the four columns are dropped with the reason "a class is a
  value of the closed set; whether a run coined it, and the coiner's definition, are not carried — the
  value's own definition is". (The jsonl era's mint-carried trio and the sqlite era's `class_new`
  column both arrive here.)
- **`class_new`** (and the jsonl spelling the era table already folds to it) → `{Translate:
  toNoEvent}`: `revisionEntry` (`migrate/entries.go:250–251`, named by the `revision` row at `:47`) renamed
  `toNoEvent` ("translates an archived event to none, whatever its era's columns"), the one function both
  the `revision` and `class_new` rows name; the manifest counts each in `events_in` and none in
  `events_out`. No reference names a `class_new` event itself (a pin names a report head, a
  contradiction a source reading, a plan a head, pin, edit or check), but every event after one is
  written one number lower. Every reference migrate carries by events.id is translated through
  `newID` (§III.8a); on m28 that is 9 stored plans, which `planRenumbered` refuses today. m29's one
  `class_new` (566) precedes two stored plans that name no id above it; they migrate byte-identical.
- **Materiality (F9).** A mint the source records with `class_material` keeps it (`stampClassMaterial`
  under `Migrating`, unchanged) — all 11 runs from 2026-10-09. A mint the source records without one
  (the 16 earlier runs, 123 mints) is stamped by the live rule: its translated value's
  `material_default`, the same `stampClassMaterial` path a live mint takes. There is no second table
  and no special fill — `shippedDefault`, `byGradeFillWhy` and `record.ShippedMaterialDefaults` are
  deleted. Each such mint is a `StatedFill` (`Where: "mint <gap_id>"`, `Slug`: the archived slug,
  `Value`: the stamped word, `Why`: "the source predates class_material; the gap takes the material
  default of the value its archived slug translates to"), every one of the 123 and not only those
  whose result moved: which moved can be computed only against the deleted table. The fill is
  recorded once the write lands, as `class_translations` counts only landed mints: a mint the write
  refuses stamped nothing (`TestARefusedPreFieldMintStatesNoFill`).
  **Residue, accepted (F9):** 17 archived gaps in 8 runs move `by_grade` → `always` against today's
  fill (the ruling's "9 runs" names the same 17 gaps; the list holds 8 run names):

  | run | archived slug (mints) → value |
  |---|---|
  | `2026-08-22_is-7-prime` | `research-protocol-violation` (2) → `self_attestation` |
  | `2026-08-22_is-7-prime-run2` | `contradictory-framing` (1) → `cross_section_contradiction`; `unsettled-inquiry` (1) → `claim_contradicts_own_record` |
  | `2026-08-22_record-store-authority` | `silent-no-match-probe` (1) → `verification_scope_blindspot` |
  | `2026-08-23_research-loop-counterparts` | `risk-taxonomy-conflation` (1) → `metric_conflation` |
  | `2026-08-23_sleeper-service-plan` | `unflagged-inferential-transfer` (1), `model-scope-overclaim` (1) → `within_source_condition_misattribution`; `citation-support-mismatch` (1) → `citation_figure_misattribution` |
  | `2026-09-02_quadratic-formula` | `graded-warrant-contradicts-record` (2) → `cross_section_contradiction`; `unrecorded-report-edit` (1) → `claim_contradicts_own_record`; `misplaced-citation-anchor` (1) → `citation_figure_misattribution` |
  | `2026-09-11_is-91-prime-b6` | `inquiry-fate-mismatch` (G2), `record-status-erasure` (G7) → `claim_contradicts_own_record` |
  | `2026-09-11_is-91-prime-b8` | `inquiry-reason-record-mismatch` (G1, G2) → `claim_contradicts_own_record` |

  What it changes downstream: six of the eight runs end FAIL or CEILING, where a gap's materiality
  moves only derived views (the board, the dispatch readiness a reader re-folds). **b6** ended
  VERIFIED on a PASS (event 189) over open `G2` and `G7` (`low_medium`, `by_grade`, not material
  then); as `always` they are material, the PASS gate is exempt under `Migrating`
  (`passblockers.go:551`), and `stampMigrationAdmission` records both on the migrated PASS's
  `migration_admitted_gap_ids`, which `verify` reports — the residue is visible, not silent. b8 also
  ended VERIFIED, and its two were closed `repaired` before its PASS (event 199). No mapping moves a gap to `never`
  (`structure-noncompliance` was `never` in the shipped table already).
- **The manifest** gains `class_translations: [{from, to, mints}]`: one row per archived slug that some
  mint carries and that resolves through an AUTHORED row of `migrate/gapclass.go` — a merge, a drop or
  a coined slug onto a different word. A slug that resolves by the `SameWord` fold (a kept registry
  class, and the two coined slugs that became values, `invalid-inference` and the sleeper run's
  `structure-noncompliance`) gets no row: it is the same word respelled, which the fold already
  states. So the rows are exactly the authored-table rows a mint used; a merge is visible, not
  silent. m29's is `[summary-scope-omission → reader_channel_mismatch (1)]`.
- **The staged registry**: `copyChannels` skips `records/class-registry.json`; `translateStagedRegistry`,
  `shippedDefault`, `byGradeFillWhy` and the `ClassNew` fill arm (`migrate/replay.go:317–328`) deleted;
  the `migrate.go:89–95` call removed.
- `remap.go`: `Mint.definition Mint.neighbor Mint.distinguisher ClassNew.*` leave `argumentFields`,
  `ClassNew.slug` and `Mint.class` leave `dataFields` (an enum, not a string — `TestEveryStringFieldIsInOneList`
  holds the lists to the schema).
- `manifest.go:82` (`Where: "staged registry"|"class_new event"`) rewritten for the class fill.
- Tests: `classmaterial_test.go` → `gapclass_test.go` with
  `TestTheGapClassMigratesOntoTheEnum` on `2026-10-18_is-91-prime-m28` (current epoch, two
  `class_new` events, two `invalid-inference` mints) and `2026-08-23_sleeper-service-plan` (jsonl era,
  mint-carried coinings, no `class_material`), each driven as the shipped path runs — `migrate.Migrate`
  on the extracted tarball (`recordtest.RequireArchive`), the archive's channels and all, not a
  hand-built `Source`: every mint's class is a `GapClass` value; m28's `class_material` per `gap_id`
  equals the source's and m28 has zero refusals; sleeper's `citation-support-mismatch` mint is
  `citation_figure_misattribution` stamped `always` (today's fill gives it `by_grade`, so this row
  fails if the special fill survives) and its three `guardrail-scope-overclaim` mints
  `partial_control_coverage` `by_grade`, each of the four with a `StatedFill`; zero `class_new` events; no
  `records/class-registry.json` in the destination. m28's stored plans are `TestAStoredPlanCrossesARenumbering`'s
  (§III.8a). `TestAPreFieldMintTakesItsNearestValuesDefault` (F9's named test) on
  `2026-09-11_is-91-prime-b6`: `G2` (`inquiry-fate-mismatch`, `low_medium`, no recorded stamp) migrates
  as `claim_contradicts_own_record` stamped `always`, a `StatedFill` names it, and the migrated PASS's
  `migration_admitted_gap_ids` holds `G2`'s and `G7`'s migrated ids and nothing else; restoring
  `shippedDefault` fails all three assertions. This INVERTS today's
  `TestMigratedPassOverOnlyNonMaterialGapsCarriesNoAdmission` (`migrate/classmaterial_test.go:365–`),
  which asserts b6's PASS admits nothing because its two gaps stayed `by_grade`: under F9 that b6 case
  is the named test above, and the no-admission rule it pinned keeps a test of its own on a seeded
  source (a lens mint `spec_underspecification` at `low`, open under a PASS, migrated → an empty
  `migration_admitted_gap_ids`). The file's other PASS-admission tests
  (`TestMigratingAdmitsArchivedPassOverNowAlwaysGap`, `TestLivePassOverOpenAlwaysGapRefused`,
  `TestMigrationRecordsAdmittedPassOnArchives`, `TestMigrateRunWrittenByThisBinaryKeepsNewFields` with
  `thisBinaryRun`, its `ClassNew` dropped and its slugs → `false_universal`/`structure_noncompliance`)
  move to `gapclass_test.go` unchanged in subject; `:275–284`'s registry byte-compare becomes "no
  `records/class-registry.json` at `--to`". **F11:** `publish_test.go`'s `TestARefusalNamesToNotTheBuild`
  (`:427–491`, the file's only registry staging, `:429,455` included) triggered its refusal by
  corrupting the staged registry, which no longer exists; it is deleted, and with it the refusal
  rewrite in `migrate.go:170–176` (the comment and the `b.atTo(m.Refusals[i].Err)` loop).
  `build.atTo` stays: the warnings (`migrate.go:69`) and `atToError` (`publish.go:293`) call it.
  **Residue, accepted (F11):** no refusal names a run path once the registry goes; a later one that
  does publishes the build path into `inputs/migration.json`, and nothing fails — no lens reads the
  manifest, and migrate runs outside a debate. `:213`'s `files < 4` floor is re-measured without the registry. `class_translations`: m28's is exactly
  `[audited-artifact-sibling-halo → verification_scope_blindspot (1)]` (its mints, measured on the
  archive: 11 slugs, of which `invalid-inference` (2) folds and carries no row); sleeper's holds
  `guardrail-scope-overclaim → partial_control_coverage (3)` and no `structure-noncompliance` row. `gapplace_test.go:42`,
  `reproduceanchor_test.go:158` drop their registry staging. `TestEveryArchivedRunMigrates` is the
  census over all 28 tarballs.

### III.8a migrate — event ids across a renumbering (F8) [NEW]/[MODIFY] (census C15)

The replay moves events.ids in both directions. A source event that writes none moves every later id
down: one that translates to none (`class_new` here, `revision` already — both through `toNoEvent`), a
correction or replacement, written with its chain's root (`migrate/replay.go:212`), and an accepted loss
(`:236–238`). The cast written first and an anchor placed after an act move later ids up. `newID`
(`:205`) records where each source event that wrote one landed, whatever moved it, so every translation
below reads it and none assumes a direction. Every place an archived record names an event by number,
and what migrate does with it:

| reference | today | decision |
|---|---|---|
| `Dispatch.pin` (`migrate/replay.go:286–293`) | through `newID`; unlanded → refusal | unchanged |
| `Finding.about_ref` on a `contradiction` (`contradictionRef`, `:306–316,530`) | through `newID` | unchanged |
| stored `Plan`: `head`, each `stale_areas[].pin` | refused when moved (`planRenumbered`, `:295–304,510–528`) | through `newID` |
| stored `Plan`: a `contradiction` blocker's `subject` (`passblockers.go:434`, `c.idWord()` — an events.id as a word) | not read; carried as printed | through `newID` |
| stored `Plan`: the ids its `why` lines print — `never sat — audits the report at head N`, `has sat through no sitting — … — audits the report at head N`, `head N moved past its last sitting's pin N`, `sat at the head (pin N)`, `retired (pin N)` (`lensFold.readiness`, `retirement.go`); `its sitting opened at N …` (`unsatSitting`, `retirement.go`, in the `has sat through no sitting` and `head moved` lines); `contradiction N` among a lens's debts (`retirement.go:133`, through `retiresWhy`); `changed at edit N … its last check is event N` (`reopenedFact`, `passblockers.go:124`, through the reopened why `:237`); `contradiction N` (`contradictionFact`, `:118`, through the contradiction why `:203` and `dispatch.go:394,398`) | carried as printed | through `newID`, by the phrase table below |
| stored `Plan`: the seat words its `why` lines carry — a `%q` claim (`dispatch.go:394,398`, `passblockers.go:203`), the bench ruling's direction (`dispatch.go:365`, bare `%s`) — and `remand_owed[].direction` | carried as printed | **carried as written (F12)** — neither translated nor refused |
| stored `Plan`: `key` | over every field above but `epoch`/`final_sitting` | recomputed (`Plan.WithKey`) |
| stored `Plan`: other `subject`s, `parties[].gap_ids`, `docket`, `remand_owed[].gap_id`, `exchanges`; `Gate.migration_admitted_gap_ids` | content ids (gap, motion, anchor, seat), current shape at epoch ≥ 31, which `remap.id` keeps (`remap.go:202–205`) | unaffected |
| `events.key`, `events.sitting_id` | `seat:slug:#ordinal`; the sitting is the write's own derivation | unaffected |
| seat prose on the event tables that names an event by number | carried as written | **residue, accepted** (below) |

**The `why` producers, every operand classified** (census C15): each `%s`, `%q`, `%d` and
concatenated operand of a line that reaches `Plan.Why`.

| producer | operands | class |
|---|---|---|
| `dispatch.go:284` `f.why` | the lens fold's line | → `lensFold.readiness` |
| `lensFold.readiness` (`retirement.go`) | `seat`; `head`, `f.pin` (`%d`); `f.barren` (`%d`); `retiresWhy()`; `short` | seat id; **ids**; count; → `retiresWhy`; → `unsatSitting` |
| `unsatSitting` (`retirement.go`, from `1ac3ab5d`) | `r`, the sitting's opening (`%d`); `ends.short[r]` (`%s`, satThrough's three fixed reasons, `dispatch.go`) | **id**, after `opened at`; fixed text |
| `retirement.go:67–75` `retiresWhy` | `retiresAfter` (`%d`); `f.owes` joined — `lensOwesOf`'s items `:110,114,125,129,133,138`: `blue's answer to <gap>`, `the accepted grade motion on <gap>`, `the bench's remand of <gap> in <motion>`, `<seat>'s finding <finding> about <gap>`, `<blocker kind> <subject>`, `unjudged proof <anchor>` | count; verb text over content ids, seat ids and a kind word — the one events.id is `contradiction <subject>` (`owedBlocker`) |
| `dispatch.go:322` | `g.id`, `g.supersededBy` | content ids |
| `dispatch.go:337,341,369` | `g.id`, `g.unruledMotion` | content ids |
| `dispatch.go:346` | `g.id`; `notMaterialBecause` (`:1192`) | content id; verb text over a severity enum word |
| `dispatch.go:350,352,359,361` | `g.id`; `x.Counted()` (`impasse.go:60`) | content id; verb text over `%d` counts |
| `dispatch.go:365` | `g.id`; `x.Counted()`; **`g.direction`** | content id; counts; **seat prose, bare `%s`** — the bench ruling's direction, after the fixed `on the ruling's direction: ` |
| `dispatch.go:391` `benchReadiness` (`:1036–1054`) | `b.Subject`, `b.About`; `b.motionSubject` | content ids; enum word |
| `dispatch.go:394,398` `contradictionReadiness` | `b.Owner`; `b.Detail` (`%q`); `contradictionFact(b)` | seat id; **seat prose, `%q`** (the contradicted claim); `contradiction <events.id>` |
| `dispatch.go:416` → `passblockers.go:203` | `contradictionFact(b)`; `ownerOr`; `b.Detail` (`%q`); `planHold` | `contradiction <events.id>`; seat id or fixed text; **seat prose, `%q`**; fixed text |
| `dispatch.go:416` → `passblockers.go:237` | `reopenedFact(b)` (`:124`: `b.Subject`, `edit %d`, `event %d`); `planHold`; `reopenedOwed` | anchor id, **ids**; fixed text; seat id or fixed text |
| `dispatch.go:416` → `passblockers.go:297` | `b.Subject`; `b.Detail` (the gavel's `phrase`, `:463–469`); `planHold` | motion id; verb text (the gavel table's phrase, or a schema error's); fixed text |
| `dispatch.go:416` → `passblockers.go:328` | `planHold` | fixed text |
| `dispatch.go:431` | `params.MaxEpochs`, `len(plan.Parties)` (`%d`) | counts |

Not why producers: `passblockers.go:268,273` are `BlockerStaleArea`'s item and refusal wordings
(`pin %s` of `b.Detail`); the kind has no `why` func, so a stale area reaches the plan only as
`StaleAreas[].Pin` and the lens fold's `retired (pin N)`. So a `why` line carries seat words in
exactly two shapes — a `%q` span, and the remand's direction, the one bare `%s` of seat prose, which
the verb always prints last on its line. Content ids (`G-bc663c4b`) and seat ids (`red-lens-logic-2`)
are words holding digits, never numerals under step 4's definition.

The table is the producers at the branch head, re-read after the independent review of PR #1523
found `unsatSitting`'s `opened at %d` (added by `1ac3ab5d`, after the first census at `b283d1e8`); an
archived plan was printed by whichever producer ran it, so the census is re-read over the 12 commits
that touched `dispatch.go`, `retirement.go`, `passblockers.go` or `impasse.go` since plan storage began
(`ca32752a`, 2026-10-17; `~/.claude/scratch/relocations/audit5/producer-{history,diff}.txt`) and the two
since (`1ac3ab5d`, `5061518d`): every seat word any of them printed is `%q`; the remand wording is
unchanged since `dd15137f` (2026-10-04); every events.id they printed follows `head`, `pin`, `opened at`,
`edit`, `event` or `contradiction` (the removed lines print `edit %d`, `head %d`, `pin %d`), and their
other numerals are counts (`retires at %d`, `%d claim(s)`). `opened at` is two tokens: a bare `at` would
read the count in `retires at %d` as an id. The census is held by a test, not by this paragraph:
`TestEveryIDTheVerbPrintsIsTranslated` (`record`) renders every producer that prints an events.id with a
distinct sentinel at each id operand, in each wording, and requires migrate's translation to move every
one — a phrase the verb begins printing an id after, missing from the table, is a refusal there.

- **`migrate/planids.go`** [NEW] `TranslatePlanIDs(pl *recordpb.Plan, newID map[int64]int64) error`,
  replacing `planRenumbered` at the same call site:
  1. `record.ReadPlan` the printed bytes (a refusal as today if they do not read).
  2. `to(n)`: `0` stays `0`; a landed id maps through `newID`; an id that wrote no event is a refusal
     ("this stored plan names event N, which wrote no event on the migrated record before it").
  3. `Head`, each `StaleAreas[].Pin`, and each `Blockers[]` of kind `contradiction`'s `Subject`
     (parsed as an integer; not one → refusal) through `to`. `RemandOwed[].Direction` is not read.
  4. Each `Why` line, scanned left to right:
     - **A numeral** is a maximal run of `0-9` bounded on each side by the line's end or a character
       outside `[A-Za-z0-9_-]` — Go's `\b`, with `-` joining the word class. Implemented by splitting
       the line into maximal `[A-Za-z0-9_-]+` tokens: a numeral is an all-digit token. So `663` in
       `G-bc663c4b`, `7` in `G-7b637cca` and `2` in `red-lens-logic-2` are not numerals.
     - **A `"`** opens a Go-quoted span, read with `strconv.QuotedPrefix` — a `%q` claim: the seat's
       words, skipped whole, neither translated nor checked. A `"` that opens no valid span is a
       refusal naming the line ("a quote that does not close — the translation cannot tell the seat's
       words from the verb's").
     - **The first `remandDirectionPrefix` (`"on the ruling's direction: "`, the prefix `dispatch.go:365`
       formats with) outside a quoted span ends the scan (F12):** the rest of the line is the bench
       ruling's direction, carried as written — its quotes, numerals and phrases included. Searched
       outside quoted spans, so a claim quoting the prefix does not hide the `(contradiction N)` after
       it (`dispatch.go:394,398` print the id after the claim).
     - Outside those spans, a numeral is an id, and goes through `to`, when the text immediately before
       it is one of the verb's id-introducing phrases — `planIDPhrases = {"head", "pin", "opened at",
       "edit", "event", "contradiction"}`, each one or more whole tokens a single space apart, the
       comment naming each format string in the table above — and exactly one space separates the two.
       A phrase begins at a token boundary: `ahead 412`, `spin 412` and `reopened at 412` are not
       `head 412`, `pin 412` and `opened at 412`, and their numeral is "any other numeral" below. Measured: no phrase match on the four plan-storing archives sits at a non-token
       boundary (`~/.claude/scratch/relocations/audit5/step4.txt`, m29 in `rev5/`), so the rule refuses
       nothing an archive holds.
     - **Any other numeral whose value is a MOVED id** (`newID[n]` landed and `≠ n`) is a refusal
       naming the line — "prints N, a moved event's number, outside every id phrase; the translation
       cannot tell whether it names that event". So a phrase the table lacks fails LOUD the first time
       an id it prints moves, never silently. A count that happens to equal a moved id is refused the
       same way (residue below).

     `planIDPhrases` and `remandDirectionPrefix` describe ARCHIVED bytes, so they are migrate's own
     table, not the live verb's constant: a later rewording is a row migrate adds for the records printed
     after it, keeping the old one for the records printed before.
     - **The remand prefix is compared with the live one (F14).** `record` exports
       `RemandDirectionPrefix = "on the ruling's direction: "`, and `dispatch.go:365` formats with it
       (`… the one more exchange the remand grants, " + RemandDirectionPrefix + "%s"`). Name checked:
       `RemandDirectionPrefix` has 0 hits; it extends the existing `remandDirection` (the remand's
       direction — `migrate/entries.go:243`'s carrier, `record/remand_test.go:16`'s fixture), the same
       concept. `TestMigrateKnowsTheVerbsRemandPrefix` [NEW, `migrate/planids_internal_test.go`] asserts
       `remandDirectionPrefix == record.RemandDirectionPrefix`, failing with "the remand prefix the verb
       prints changed — keep the old wording for archived bytes and add the new one"; rewording the
       constant fails it. Once a second wording exists, migrate's prefix becomes a list and the test
       asserts the live prefix is one of its entries.
     - **The id phrases have no live constant** — they are spread across the producers' format strings
       (`retirement.go:133,206–216`, `passblockers.go:118,124`) — so
       nothing compares them with the live verb. The m28 test pins migrate's table to ARCHIVED bytes,
       which are fixed and cannot show a later rewording. A live rewording surfaces loudly instead, the
       first time an archive printed with it migrates across a renumbering: its moved id follows a word
       the table lacks and the plan is refused (the moved-id rule above).

     Measured on the four plan-storing archives (`final` 3 plans, `n1` 3, `m28` 20, `m29` 17;
     `~/.claude/scratch/relocations/audit4-rev/py/{scan,quotes}.py`, `audit5/py/step4.py`): no stored line
     holds a quoted span or the remand prefix; m28's lines hold 223 numerals after an id phrase and m29's
     161, and the largest numeral outside every phrase is 7 (m28) and 6 (m29), against moved ids ≥ 565
     and ≥ 567 — so none is refused. Read as bare
     digit runs instead, the hex inside `G-bc663c4b` and `G-7b637cca` would refuse 14 lines in 7 plans
     (527, 530, 604, 608, 723, 726, 745; `~/.claude/scratch/relocations/audit4/m28-numerals.txt`):
     the token rule is what stops it.
  5. Nothing changed → `pl` untouched (a record that keeps its numbering migrates byte for byte —
     `TestAStoredPlanMigratesAsPrinted` on a seeded run, `TestAnArchivedRunStoresOnlyThePlansItStored` on
     the archives). Changed → `p = p.WithKey()`, `pl.Printed = p.Printed()`.
     The write's `requirePrintedPlan` (`planstore.go:69`) then admits it: a printing, its own key,
     relayable.
- **`record/planstore.go`**: `keyed` → the method `Plan.WithKey()` (exported; `dispatch.go:195`,
  `dispatchnext.go:134`, `requirePrintedPlan` (`planstore.go:77`) and `planstore_test.go:272,295,299,311`
  call it), and `debatejs.Keyed` → `debatejs.WithKey` with the JS harness's stub of the same concept
  (`tests/simulator/harness.mjs:92` `keyed` → `withKey`, its `:89` comment and `:93` call, and
  `debate.test.mjs:4,1089`) (D10). Exporting the live key rule keeps one rule; it adds no field, id or
  mechanism to the record.
- **Prose that says how migrate treats a plan-storing record's numbering or bytes** (census C15,
  `git grep -n -i -E 'numbering|renumber|carried as printed|migrates as printed|byte for byte|byte-for-byte|byte-identical' -- plugins/frank-exchange-of-views/tools/internal/record/migrate/`,
  every hit read; `~/.claude/scratch/relocations/audit5/rev5/c15-migrate.txt`):
  - `storedplan_test.go:16–18` [MODIFY]: "A STORED PLAN MIGRATES AS PRINTED WHERE NOTHING RENUMBERS
    IT. A seeded run drops no event, so each plan's head and pins name the same events after migrate
    and the bytes migrated are the bytes stored; a record whose replay drops an event has its plans
    translated through `newID` (`TestAStoredPlanCrossesARenumbering`)." `:52`'s message keeps "as
    printed" — it states this test's case.
  - `benchoccasion_test.go:60–126` `TestAnArchivedRunStoresOnlyThePlansItStored` [MODIFY] — today it
    asserts every archive's migrated plans equal the source's byte for byte (`:64`, `slices.Equal` at
    `:112`, message `:113`), which m28's 9 translated plans fail. Rewritten: per archive, the migrated
    record stores as many plans as the source, in stream order; each pair is byte-identical where the
    replay moved no event the plan names; a pair that differs must be equal with every numeral (step
    4's token rule) and the key masked, and is `TestAStoredPlanCrossesARenumbering`'s to check number
    by number. Comment `:60–66`: "… An archive written at epoch 31 or later carries the plans it
    stored — byte for byte where the replay moved no event a plan names, and otherwise with only those
    numbers and the key changed (`TestAStoredPlanCrossesARenumbering`)"; message `:113`: "… stores
    exactly the source's plans, byte for byte where nothing a plan names moved". As measured, `final`,
    `n1` and `m29` are byte-identical and m28's 9 are the only differing pairs. Mapping a numeral that
    did not move (every phrase numeral `n` → `n−1`) fails it on `final`, `n1` and `m29`; carrying a
    plan with a changed word fails the masked comparison.
  - `replay.go:295–299` [MODIFY]: the comment states the translation — "a stored plan names events by
    events.id inside bytes a relay is compared with, and its key is over them; `TranslatePlanIDs`
    carries each through `newID` and re-keys the plan, and refuses an id it cannot place".
  - `replay.go:507–528` (`planRenumbered`, "carried as printed") and `planrenumbered_internal_test.go`
    [DELETE] with the function.
  - Already true, unchanged: `dispatchpin_test.go:12,20` (every archive renumbers; a seeded run shifts
    nothing), `reproduceanchor_test.go:125,190`, `reopenedm19_test.go:15` (a migration renumbers rows),
    `replay.go:75` (dispatch numbering — the key ordinal, another sense).
  - Another sense, unchanged: `archive_test.go:119` (byte-identical DATABASE files), `inquiryfate_test.go:195`
    (an inquiry's text and reasons, carried prose), `replay_test.go:22` (the original stamps),
    `testdata/prose-rewrites.golden:653,5138,6101,6102,6154,6159,6371` (archived seat prose).
  Outside `migrate/` the sweep's hits are other senses (proto field numbers, proof and citation
  numbering, validation-loop labels; audit4's `c15-prose.txt`), and `record.proto:1019` already says
  `migrate` translates `about_ref`.
- **Residue, accepted (F13): numbers in seat prose may be stale after any migration that drops an
  event; no list of shifts is kept.** Seat prose is carried as written — the words are the seat's, and
  migrate re-authors no prose — and the replay moves ids both ways (above), so a number a seat wrote can
  name another event on the migrated record. The source tarball under `run-archive/` keeps the original
  numbering. What this PR adds to it: dropping `class_new` renumbers every archive that holds one —
  sqlite era: `b` (79, 117), `b3` (85, 149), `b6` (85, 162), `b8` (124), `b9` (138), `quadratic-formula`
  (187, 349, 355, 532), `m23` (373), `m28` (564, 572), `m29` (566); jsonl era (`class-new`): `is-7-prime`
  (1), `is-7-prime-run2` (5), `record-store-authority` (1), `research-loop-counterparts` (1),
  `sleeper-service-plan` (10). Of these `m28` and `m29` store plans, and only m28's name a moved id;
  every dispatch pin and contradiction ref in all of them goes through `newID`, as `revision` drops
  already make it. Measured (`~/.claude/scratch/relocations/audit4-rev/py/prose.py`: every text column of
  every body table, `(head|pin|edit|event|contradiction) N` with `N` above the run's first dropped id):
  `m28` 23 (`position.text` 6, `close.anchor_target` 4, `close.prose` 3, `spot_check.reason` 3,
  `inquiry_review.reason` 3, `mint.acceptance_check`, `closing.text`, `regrade.basis`, `reproduce.note`
  1 each), `m23` 8, `quadratic-formula` 3, `b8` 2, `b`, `b3`, `b9` 1 each (`log.text`), `b6` 0, `m29` 0.
  A jsonl source carries no events.id — each shard's `seq` is its own — so its prose names no number the
  replay assigns: not measured, and nothing to measure.
- **Residue, accepted: a count equal to a moved id is refused.** Step 4 cannot tell a count from an id
  outside every phrase, so a numeral that is not an id but equals a moved id refuses the plan. Safe on
  every archive held: the largest numeral outside every phrase is 7 (m28) and 6 (m29), against moved
  ids ≥ 565 and ≥ 567. On a future archive the refusal is loud and names the line, and is ruled like any
  refusal.
- **Residue, accepted: a translated plan no longer equals its relay.** `capture/dispatchparity.go`'s
  `relayDepartures` (`:248`) compares the workflow journal's relays byte for byte with the stored plans
  when a journal is present (`:209–217`); a migrated m28's 9 translated plans would each report a
  departure. No archive carries `trajectories/` (every `run-archive/*.tar.gz` listed), so no audit
  of a migrated run compares relays, and the relays were of the source numbering by construction (R15).
- **Tests.**
  - `TestAStoredPlanCrossesARenumbering` [NEW, `migrate`], on the real m28 through `migrate.Migrate`
    on the extracted tarball: zero refusals; 20 stored plans out, in stream order beside the source's
    20; the 11 that name no id > 564 byte-identical; for the 9 others, an oracle that does not use
    `planIDPhrases`, numerals by step 4's token rule: (i) source and migrated printings with every
    numeral and the key masked are equal — only numbers changed; (ii) for `head`, each `pin`, and every
    numeral position where the two printings differ, the migrated event at the migrated number is the
    source event at the source number — same `type`, `seat_id`, `key` and `ts`, read from both
    databases (`file:` URI, `mode=ro`), not from `newID`; (iii) no numeral position outside a seat span
    keeps a source number > 564 unchanged (each such source event sits at another number on the
    migrated record); (iv) the key equals `WithKey()` of the decoded migrated plan; (v) every seat span —
    a Go-quoted span, the tail after the remand prefix, `remand_owed[].direction` — is byte-identical
    between the two printings (vacuous on m28, which holds none; the internal test below carries it).
    Deleting the `why` translation fails (iii) at event 604's `head 579`; deleting the pin translation
    fails (ii); deleting the re-key makes the write refuse each of the 9 (a refusal count of 9);
    reading numerals as bare digit runs refuses 14 lines in 7 plans.
  - `planids_internal_test.go` (replaces `TestAStoredPlanWhoseEventsWereRenumberedIsRefused`):
    `TestAStoredPlanNamingMovedEventsIsTranslated` on built `record.Plan`s, `newID` moving 412 → 411 as
    a dropped event before it would.
    - **One translated plan** whose head is 412, so it is re-keyed and re-printed through
      `WithKey()` → `Printed()`: a pin 412, a `contradiction` blocker `"412"`, and one `why` line per
      phrase word, each moved to 411; a quoted claim `"head 412"`, unchanged; a line `red-lens-evidence
      read a source contradicting "x on the ruling's direction: pin 412" (contradiction 412) …`, back
      with the claim unchanged and `(contradiction 411)`; **a remand line `G-1: at impasse (2
      exchange(s) (0 stalled)) and remanded by the bench — … on the ruling's direction: re-read pin 412
      and the claim "head 412` (an unclosed quote in the tail), and `remand_owed[0].direction` holding
      the same words**; and a line `G-7b637cca: open, material, …` with 637 moved. Decoded after the
      re-print, the remand line, its direction, both quoted claims and the `G-7b637cca` line are
      byte-identical to the input — the seat spans survive the re-print path, not only the scan.
    - **Refused plans, one each:** a line `… spin 412 …` (a phrase word that is not a whole token:
      its numeral is outside every phrase and 412 moved); a line printing a moved id after no phrase
      word; a `"` outside the remand tail that does not close; an unlanded head.
    - **A plan naming no moved id** comes back with the same bytes.
    Dropping the remand rule fails the translated plan (`pin 412` in the tail becomes `pin 411`, or the
    tail's unclosed quote is refused); searching the prefix inside quoted spans leaves `(contradiction
    412)` untranslated; reading bare digit runs refuses the translated plan at the `G-7b637cca` line;
    matching a phrase as a suffix rather than a whole token translates `spin 412` to `spin 411` and the
    refusal the test expects does not come.
  - `TestMigrateKnowsTheVerbsRemandPrefix` [NEW, `migrate/planids_internal_test.go`] (F14):
    `remandDirectionPrefix == record.RemandDirectionPrefix`, failing with "the remand prefix the verb
    prints changed — keep the old wording for archived bytes and add the new one".
  - `TestARemandLineCarriesItsDirectionAsWritten` [NEW, `record/remand_test.go`], driven through the
    shipped verb: `remandedAtImpasse`'s stage with the bench ruling's direction `re-read pin 412 and
    the claim "head 412`, then `DispatchNext` as the chair; the stored plan's line for `G1` ends with
    `RemandDirectionPrefix` + the direction, byte for byte, and `remand_owed[0].direction` is the
    direction. It pins what step 4's tail rule relies on: the direction is printed last on its line,
    after the prefix, as written — formatting it with `%q`, or anywhere but last, fails it.
  The m28 test pins migrate's table to the archived id phrasings; the internal test proves
  `contradiction`, which no archived plan carries, the seat spans, which none holds, and the
  token-boundary rule.

### III.9 Telemetry [MODIFY]

- `view/view.go:500–565`: `byClass` becomes `[]*recordpb.ClassTally` in `GapClass` schema order
  (zero counts omitted; no mints → an empty list); `unclassified_rate` = unclassified mints / mints this epoch, unset when none (the wire's `null`, as `class_repeat_rate`);
  `class_repeat_rate` excludes `unclassified` from the repeat numerator (D8). `view_test.go:504–521`
  updated — its `one["by_class"].(map[string]any)` assertion (`:504`) becomes a decode of `[]{class,count}` in schema order, and `:505`'s slugs become `false_universal`/`spec_underspecification`; new `TestUnclassifiedRateIsUnsetWhenNothingWasMinted` and `…CountsOnlyTheCatchAll`.
- **The wire (D8, F6).** `inquest telemetry`'s bytes and its help come from the hand-written
  `telemetryLineJSON`/`newMintJSON` (`view.go:829–846`), filled by `telemetryWire` (`:885`); the help
  tree is `jsonshape.Tree(view.TelemetryLineShape())` (`:827`, registered at `cli/seat/verbs.go:301`).
  Nothing is generated from the proto. Edits:
  - `newMintJSON.BySeverity` → `[]severityTallyJSON` (`{Grade string "grade"; Count *int32 "count"}`,
    `Grade` = `gradeText(t.GetGrade())`); `severityObject` and its `MarshalJSON` (`:859–881`) deleted.
    The wire becomes `by_severity:[{"grade":"high","count":2}]`, which is what the help has always said.
  - `newMintJSON.ByClass` → `[]classTallyJSON` (`{Class string "class"; Count *int32 "count"}`,
    `Class` = `recordpb.Word(t.GetClass())`), in schema order.
  - `newMintJSON.UnclassifiedRate *float64 "unclassified_rate"`, filled from `nm.UnclassifiedRate`.
  - **An empty tally is `[]` (F10).** The slice-typed fields of the wire — swept by reflecting over
    `telemetryLineJSON` and every struct it holds: `new_mint.by_severity` and `new_mint.by_class`, and
    no other (`repair_regression`, `edge_deltas` hold pointers only) — are filled by
    `make([]…, 0, len(src))` in `telemetryWire`, unconditionally, including when the row carries no
    `new_mint` at all, so `encoding/json` never sees a nil slice and never writes `null` for them.
    `null` stays what it means on this wire: a pointer field nobody measured.
  - The boundary comment (`view.go:147–160`) loses its `by_severity`-is-an-object property and states
    the rule instead: the wire types carry no custom marshaller, so the reflected tree is the bytes;
    an unmeasured scalar is `null`, a tally is a list and an empty one is `[]`.
    `severityObject` was the tree's only `MarshalJSON` (`git grep -n 'MarshalJSON()' -- '*.go' ':!*_test.go'`
    prints it alone), so the class has one instance.
  - Help renders `new_mint:{count,by_severity:[{grade,count}],by_class:[{class,count}],class_repeat_rate,unclassified_rate}`;
    `lead-judge.md:91,330` and `red-chair.md:255` follow through `agentgen`.
    `scripts/agentgen/src/lead-judge.md:73` ("`new_mint.by_class` moving, and `class_repeat_rate`
    rising") stays true and is not edited.
  - `dashboard/render.go:192` reads the typed `TelemetryLine`, not the JSON — unchanged.
- **`TestTelemetryHelpTreeMatchesAnEmittedLine`** [NEW, `cli/telemetryshape_test.go`], driven as the
  bench reads it: a run where one epoch mints at `high` and `medium` under `false_universal` and
  `unclassified` (with `--reason`), then `inquest telemetry --seat-id judge` and `inquest telemetry
  --help`. It parses the help's OUTPUT tree into nodes (object / array-of-object / leaf) and decodes
  the epoch's line, and asserts at every depth: the same keys, an array where the tree says `[…]`, an
  object where it says `{…}`. For every field with a custom marshaller or a proto-enum source —
  `max_severity`, `by_severity[].grade`, `by_class[].class` — the value is a JSON string that is a
  word of its enum (`recordpb.BySpelling` resolves it; never a number, never `GAP_CLASS_*`/`GRADE_*`).
  `unclassified_rate` is present and equals 0.5. The run also holds a SECOND epoch with no mint
  (a chair sitting opened over an unchanged board), and the test decodes that line too and asserts,
  at every node the tree marks `[…]`, a JSON array — `[]` here, never `null` — while
  `class_repeat_rate` and `unclassified_rate` are `null`. It fails if `severityObject` is restored, if
  `Class` is filled with the number, if `UnclassifiedRate` is dropped from `newMintJSON`, or if either
  tally is filled by plain assignment from a nil getter.
  `TestProjectionShapeMatchesEmittedKeys` checks top-level keys only and missed `by_severity`; this
  test is the deep check for the one JSONL view.

### III.10 Harnesses and fixtures [MODIFY]

- **Fixtures mint enum words, and each keeps its materiality.** Today every placeholder is
  `by_grade`: `StageForRun` stages each slug `CLASS_MATERIAL_BY_GRADE` (`replay.go:599–620`) and
  `recordtest.Seed` defaults a seeded mint to `by_grade` (`recordtest.go:121–126`). The pass keeps
  that. **Census at `b283d1e8`** (C14, §III.14; unchanged from `61107ba4` but one `mintGap` call):
  - literal words — `git grep -h -o -E '"--class", "[^"]+"|--class [a-z0-9_-]+|Class: *proto\.String\("[^"]+"\)' -- '*.go'`:
    304 lines in 109 Go files (`x` 67, `scope-creep` 53, `overclaim` 39, `self-attestation` 33,
    `safety` 18, `c` 13, …; 42 words; some hits are prose such as `--class would`, read and left);
  - helper-carried words the literal regex misses — `git grep -n -E '^func [a-zA-Z]+\([^)]*\bclass\b[^)]*\)'`
    and their call sites: `cli.mintGap(t, runDir, key, class)` (68 calls with a literal word, 22 words: `overclaim` 16,
    `read-surface` 10, `G1` 6, …), `capture.mintEvent` (deleted with `classharvest_test.go`),
    `record.liveMint`, `migrate`'s mint builder, the stage builders' fixed `"x"`
    (`dispatch_test.go:48`, `worklist_test.go:29` `cmMint`), `view_test.go:464`, the word list
    `difftest/fuzz_test.go:282`, and the two fixed `--class fuzzcls` mint sites `releasegate/fuzz/fuzz_test.go:1851,1992`
    (`fuzzClasses` at `:6221` is staging-only — read by `StageForRun` alone, here and in
    `registerbeforeappend_test.go:47` — and goes with its staging);
  - generated carriers (18 difftest goldens, the agents) regenerate and are read, not edited.

  The rule, one pass:
  1. **A placeholder** — a word that only says "some class" — becomes **`spec_underspecification`**
     (`by_grade`; its definition, "a term, rule, procedure or recommendation the report relies on is
     not specified enough for a reader to apply or check it", fits any fixture's "the defect"). A
     test that needs two or three distinct classes (`by_class`, `class_repeat_rate`, id-distinctness)
     adds `policy_without_mechanism`, then `risk_coverage_omission` — both `by_grade`. This holds for
     placeholders that happen to spell a registry class (`self-attestation`, `metric-conflation`):
     they were staged `by_grade`, and `self_attestation`/`metric_conflation` are `always`.
  2. **A test whose subject is materiality** picks the value with the materiality it asserts:
     `false_universal` (`always`), `structure_noncompliance` (`never`), `spec_underspecification`
     (`by_grade`). A seeded mint that sets `ClassMaterial` by hand (`cmMint`, `liveMint`,
     `verify/classmaterial_test.go`) takes its class from `recordtest.ClassFor(cm)` (§III.5), so the
     class and the stamp agree.
  3. **A test proving an unknown class is refused** keeps an unknown word (`no-such-class-slug`,
     `banana`, difftest's `propagation-incomplete`).
  4. **A fixture standing in for a real sitting** keeps real words with their real defaults:
     `seatprobe.BoardClasses` and `boards.go`'s gaps (respelled — the merged slug at `boards.go:262`,
     `citation-status-drift`, becomes `live_source_drift`, Table B's target; the staged default came from
     `ShippedMaterialDefaults`, which equals Table A's `md` for every kept class and for
     `citation-status-drift` → `live_source_drift`, both `always`).

  `Class: proto.String("x")` → `Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum()`.
  `recordtest.Seed`'s default stays `by_grade`. The 13 test files calling `StageForRun*` and
  `seatprobe/build.go` (C1) lose the call; `recordtest.go:123` names it only in the comment rewritten above.
  Re-running both census commands after the pass prints only enum words, rule-3 unknown words, and
  prose hits.
- **Fuzz** (`releasegate/fuzz/fuzz_test.go`, lines at `b283d1e8`): the coining path (`:1802–1850`), `classMu`/`classMade`
  and their comment (`:345–351`; `:215` drops `classMu` from the guarded-fields list) and the three
  `StageForRun` calls (`:3692`, `:4680`, `:6268`, with `:6261–6262`'s "without the class registry every
  `mint` is refused" comment) deleted, with `classDefaultTurn` (`:117–119`, the always/never coining alternation); `class_new` leaves
  the event-type list (`:4603–4605`, "these three" → "these two") and the coverage table (`:4795`). The class a fuzz mint carries is set at the TWO mint sites, not by `fuzzClasses`:
  `:1851` (`args = append(args, "--class", "fuzzcls")`, every drive's mint) and the estoppel probe
  `:1992`. The estoppel probe becomes `spec_underspecification`; every drive's mint draws from the
  `by_grade` values (the catch-all excepted) and each `always`/`never` value is minted ONCE per sweep,
  deliberately, as closeGap's `moot` (F16) — both lists read off the enum's `material_default` facet,
  never hand-listed; the forced runs take no deliberate mint, and `TestFuzzDebate` fails naming any
  deliberate value no mint landed under. A pooled `always`/`never` class "would move every gap's
  materiality, and with it the boards, the epochs and the terminal verdicts the sweep's other gates
  count on", and the release gate's enum-coverage census requires every `mint --class` value. Left as
  `fuzzcls`, every mint is refused as an unknown word, the estoppel probe never lands, and
  `estoppels == 0` fails the gate (`:5435`) — the loud form of the miss. At `:1851`, ~10% of mints use
  `unclassified` (also `by_grade`): half of those first without `--reason`, asserting the D7 refusal,
  then with it. `fuzzClasses` (`:6217–6221`) was read only by the three staging calls and is deleted
  with them. `registerbeforeappend_test.go:47` drops its staging.
- **seatprobe**: `build.go:89,146–150,419–427` read `recordpb.MaterialDefault` for `BoardClasses`
  (now `[]recordpb.GapClass`); `boards.go` class fields respelled (C12); the two `class list`/`class
  new` duty rows (`:536–537`) deleted and `:371` comment updated; `seatprobe.go:150–156` drops
  `EVENT_TYPE_CLASS_NEW`.
- **surface tests**: `chairrefusal_test.go:269` drops staging; `promptverbs_test.go:44,130,959` drop the
  `class new` example; `vocabulary_prose_test.go:265` exemption for `recordtest`'s generated list
  deleted (the list is gone).
- **difftest**: the `class_registry` scenario (`scenarios_test.go`, `class_registry.golden`) is
  rewritten as `gap_class_enum` (mint each materiality, the catch-all with and without a reason, a
  near-miss word); `contract_test.go:94`, `golden_test.go:174,190` drop staging, and `goldenClasses` (`golden_test.go:233–242`,
  read only by that `StageForRun`) is deleted with it; `contract_test.go:308` drops `"class"` from
  `defaultSeat`'s lens-verb list (the verb is gone). The other 17 goldens
  in C4/C12 regenerate.

### III.11 Dev tooling, CI, universe [DELETE]/[MODIFY]

- `scripts/classgen/` deleted; `scripts/check/gates.go:247` gate deleted; `.github/workflows/hooks.yml:1504–1506`
  step deleted (`scripts/check/parity_test.go` reconciles the two).
- `scripts/universe.sh:519–539` deleted. Nothing else in the universe needs `feov-memory`: `setup`
  read only the class registry from it (C10), and `universe.sh` does not run `capture`, whose
  `WriteScorecards` is the only other reader and already returns a reason on a missing directory.
- `scripts/universe-stream.py:25`: `"class"` leaves `INTERESTING`, the verbs the stream prints (the
  verb is gone). `universe-board.py:68` (`SELECT "gap_id", "class" FROM "mint"`) reads the column,
  which holds the word — unchanged.
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
- `skills/research-protocol/scripts/debate.js:989` (authored lens prompt): "registering a new class
  first where the registry lacks one" → "under the class that fits, or the catch-all with the reason
  none does". Names an act, no command path (`TestTheSeatPromptsNameNoVerb`).
- `tests/simulator/debate.test.mjs:517–518`: `class new` leaves the command-path regex (the path no
  longer exists; a ban list of a deleted word is the archaeology the repo refuses). `:408` asserts the
  evidence lens's prompt holds `'registering a new class first'`, the phrase the `debate.js:989` edit
  above removes: it becomes `'or the catch-all with the reason none does'`, so the test still pins the
  prompt's class duty and fails if the edit is reverted. `:1033`'s comment ("the class registry is the
  vocabulary, it is enforced at mint") → "the gap class's closed set is the vocabulary, enforced at mint".
- `difftest/testdata/help_contracts.golden:100` and `role_boundaries_and_help_contracts.golden:180`
  render the lens group list's `class       the gap-class registry: what KINDS of defect this run
  recognises`; with the group deleted the row leaves both on regeneration, and each is read.
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

Commands run at `be6c110d`, re-run at `61107ba4`, and re-run at **`b283d1e8`** — the base every line
below is read at — from the worktree root with `-- ':!run-archive' ':!plans/historical' ':!research'`.
`61107ba4..b283d1e8` (45 commits: Slim-fast's #1481 cut fields from `record.proto` and bumped the epoch
to 32 with a migrate translation; #1495 touched the fuzz; #1497 touched migrate; the m29 run was
archived) touched 154 files, among them `capture.go`, `cli/lens/mint.go`, `flags/names.go`,
`record/{enums,enumof,record,viewjson}.go`, `record.proto`, `record.pb.go`, `schema.sql`,
`migrate/{entries,manifest,migrate,remap,replay}.go`, `migrate/benchoccasion_test.go`, the migrate goldens,
`seatprobe/build.go`, `debate.js` and `releasegate/fuzz/fuzz_test.go`. Every line this plan cites was
mapped from `61107ba4` to `b283d1e8` by a line diff and refreshed
(`~/.claude/scratch/relocations/audit5/rev5/{linemap,census-diff,c16-diff}.txt`); every census below
was re-run at `b283d1e8` and its delta read: the same files and counts but for shifted lines, the new
`migrate/testdata/prose-rewrites.golden` hits (archived seat prose — C4, C6, C12, the §III.5 sweep),
`scripts/secretsgen/gitleaks/gitleaks.toml` ("registry" in a token rule — C16 N) and one more
`mintGap` call (C14). The producers of a stored plan's `why` (`dispatch.go`, `retirement.go`,
`passblockers.go`, `impasse.go`) and `planstore.go` are untouched by the range. Full output:
`~/.claude/scratch/relocations/gapenum/census/c*.txt` (`be6c110d`),
`~/.claude/scratch/relocations/audit2/c*.txt` (`61107ba4`); the compact form
below is `file lines` (I/ = `plugins/frank-exchange-of-views/tools/internal/`, F/ =
`plugins/frank-exchange-of-views/`). `plans/` and `law/` hits are counted, not listed: plans are design
records (no change); `law/proposed/class-*.md` (14) are deleted (III.11).

Re-running each command after the change must print only the "no change" consumers below.

<details><summary>C1 — <code>git grep -nE 'ShippedClasses|ShippedMaterialDefaults|StageClassRegistry|ValidateClassRegistry|loadRegistry|StageForRun|ClassRoster|knownClasses|classMaterialFor|stampClassMaterial|HarvestClasses|translateStagedRegistry|shippedDefault|registry\(\)'</code> — 124 lines outside <code>plans/</code></summary>

```
feov-memory/class-registry.json 2
F/tools/integration/surface/chairrefusal_test.go 269
I/capture/capture.go 1769,2077,2094
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
I/record/migrate/replay.go 56,58,59,321
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
F/tools/releasegate/fuzz/fuzz_test.go 3692,4680,6268
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
| `capture.go:1769,2077,2094` | yes — `HarvestClasses` and its call deleted (III.11 ruling) |
| `cli/lens/class.go` | deleted |
| `seatprobe/build.go`, `boardclasses_test.go`, `cli/seatprobe_fixture_test.go` | yes — III.10 |
| `chairrefusal_test.go`, `referencechecks_test.go`, `difftest/golden_test.go`, `classmaterial_test.go`, `replay_test.go:605`, `mintbudget_test.go`, `join_test.go`, fuzz ×2 | yes — staging calls removed (III.10) |
| `recordpb/descriptions.go:75` | no — matched `registry()` in a terms-registry sentence |
| `scripts/rulesweep/*` | no — `registry()` is the PROTOCOL class registry |

<details><summary>C2 — <code>git grep -nE 'class-registry'</code> — 74 lines outside <code>plans/</code></summary>

```
.github/workflows/hooks.yml 1065
CLAUDE.md 63
feov-memory/class-registry.json 2,3
feov-memory/protocol-class-registry.md 3
F/docs/seat-command-triggers.md 73
I/capture/capture.go 2079,2118
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

<details><summary>C4 — <code>git grep -nE 'class_new|ClassNew|CLASS_NEW|classNew|"class new"|class new|ClassNewVerb'</code> — 224 lines outside <code>plans/</code></summary>

```
feov-memory/class-registry.json 4
F/agents/red-lens-{adversary,architecture,computation,dark-side,evidence,logic,voice}.md 2 each (~220, ~417)
F/docs/seat-command-triggers.md 74
F/docs/seat-surface-naming.md 252
F/tests/simulator/debate.test.mjs 517,518
F/tools/integration/surface/promptverbs_test.go 44,130
I/capture/capture.go 2045,2046,2057,2127
I/capture/classharvest_test.go 19,67,71
I/cli/badvalue_test.go 115
I/cli/cli_test.go 825,840,856,857,859
I/cli/help_test.go 149
I/cli/lens/class.go 35,103,107,120,131,137
I/cli/lens/mint.go 88,89,95,259,261
I/cli/referencechecks_test.go 265
I/cli/requiredmarker_test.go 96
I/cli/unknownclass_test.go 14,15,33,34,41
I/cli/verbguardrace_test.go 225,227,229,241,242
I/diagnostics/manual_test.go 28,33,122
I/difftest/testdata/{a_repeated_act_names_its_answer,carried_from_renders_as_carried,class_registry,close_validation_and_archive,docket_ruling_requires_each_unconditional_field,error_catalogue,full_round_integration,hostile_prose_via_file,ids_stay_distinct_across_rounds,mint_idempotency_on_crash_retry,mint_validation,missing_required_flags,multi_nonce_terminal_event_wins,projections_debate_changelog_citations,regrade_history_is_recoverable}.golden
I/record/classmaterial_test.go 180,182,183,195,196
I/record/classroster.go 11,17,27,91,95
I/record/classroster_test.go 25,66
I/record/enumof.go 112
I/record/enums.go 285,287
I/record/migrate/classmaterial_test.go 69,71,77,83,92,108,…(+5)
I/record/migrate/manifest.go 82
I/record/migrate/remap.go 115,116,127
I/record/migrate/replay.go 53,320,325
I/record/migrate/testdata/prose-rewrites.golden 2516,3627
I/record/record.go 823,824
I/record/recordpb/body_test.go 25
I/record/recordpb/record.pb.go 64,104,141,…(+37)
I/record/recordpb/record.proto 237,581,718,723,725,825,834,835
I/record/recordpb/testdata/payload-keys.txt 20
I/record/recordsql/testdata/schema.sql 96,474,493,524
I/record/recordtest/recordtest.go 129,130
I/record/replay.go 508,560,586,670,679,696,700,708
I/record/replay_test.go 692,693,700,702,703,704,…(+6)
I/record/writerace_test.go 70,71
I/seatprobe/boards.go 371,537
I/seatprobe/seatprobe.go 81,150,156
I/setup/run.go 343
F/tools/releasegate/fuzz/fuzz_test.go 1802,1831,1846,4605,4795
law: 14   plans: 16
```
</details>

Every hit changes: generated (agents, `record.pb.go`, goldens, `schema.sql`, `payload-keys.txt`)
by regeneration; the rest by III.4–III.13. `diagnostics/manual_test.go:28,33,122,127` names the
`class` group as a traversal fixture and is updated to a group that still exists; `body_test.go:25`
drops the `ClassNew` row; `writerace_test.go:70` (a coining racer) is deleted. Two exceptions, both
archived seat prose in `migrate/testdata/prose-rewrites.golden`: `:2516` (sleeper's `ClassNew.distinguisher`
id rewrite) leaves on regeneration, because a `class_new` writes no event and so no prose is rewritten on
it; `:3627` (quadratic-formula's `Log.text` naming `class new`) stays — seat words carried as written.

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
F/tools/releasegate/fuzz/fuzz_test.go 1814,1816
scripts/rulesweep/sweep_test.go 94
plans: 2
```
</details>

No change: `debate.test.mjs:397` ("…NOT MATERIAL, BY CLASS…" — the chair's by-class listing of
non-material gaps, which survives), `record/available.go:122` (same listing) — both out of
criterion 8's grep by its `\b`, `fetchcache/testdata`,
`report/site.js` (CSS `classList`), `vocabulary_prose_test.go:265` (removed with III.10),
`rulesweep/sweep_test.go:94` (protocol registry; out of criterion 8's grep by path). The rest change
(III.6, III.10).

<details><summary>C6 — <code>git grep -nE 'by_class|class_repeat_rate|ByClass|ClassRepeat'</code></summary>

```
F/agents/lead-judge.md 91,330
F/agents/red-chair.md 255
I/difftest/testdata/manual_runs_every_help_page_live.golden 208
I/record/enums.go 159,165,169,195
I/record/enums_test.go 439
I/record/migrate/testdata/prose-rewrites.golden 1153,1156,1157
I/record/recordpb/record.pb.go 2737,…(+14)
I/record/recordpb/record.proto 675,677,679,681
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
(`CountByClass` — voice-tell classes), `agentgen/src/lead-judge.md:73` (still true),
`migrate/testdata/prose-rewrites.golden` (archived seat prose quoting `class_repeat_rate`, carried as written). Change:
`record.proto`, `view.go`, `view_test.go`, and the generated agents, golden, `payload-keys.txt`.

<details><summary>C7 — <code>git grep -nE 'material_default|MaterialDefault|class_material|ClassMaterial'</code> — 480 lines, by file</summary>

Non-test Go and prose: `record/replay.go`, `record/enumof.go`, `record/record.go`, `record/enums.go`,
`record/recordpb/record.proto`, `cli/lens/class.go`, `cli/lens/mint.go:284`, `flags/names.go`,
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
I/cli/lens/mint.go 93,261
I/cli/motion/command.go 74
I/cli/motion/verbs.go 129,131,226,476
I/flags/names.go 166,167,168
I/record/classroster.go 95
I/record/classroster_test.go 145,146
I/record/recordsql/testdata/schema.sql 475,…(+5)
I/record/replay.go 424,425,426,716,717,718
I/record/replay_test.go 702,703,704
I/terms/terms.go 56, terms.json (40), terms_test.go (13), scripts/vocabdoc/main.go 44
```
</details>

No change: `cli/motion/*` (`flags.Class` is the PETITION class flag, `PetitionClass`), `terms/*` and
`vocabdoc` (the terms registry's own `definition` key). The rest change or are deleted.

<details><summary>C10 — <code>git grep -n 'feov-memory'</code> (beyond C2)</summary>

`capture.go:1571,1753,1758,1946,2162`, `capture/readwhen_test.go:40`, `cli/scorecard.go:61`,
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
F/skills/research-protocol/scripts/debate.js 989    (change)
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
fuzz) — respelled by the fixture pass (III.10). At `b283d1e8` the census adds 66 lines in one file,
`migrate/testdata/prose-rewrites.golden` — archived seat prose naming slugs, carried as written (no change).
</details>

<details><summary>C13 — mint-class readers: <code>git grep -n -E 'GetClass\(\)|\b(m|p|mint)\.Class\b' -- 'plugins/frank-exchange-of-views/*.go' ':!*_test.go' ':!*.pb.go'</code></summary>

Scoped to FEOV: unscoped, the command also prints `plugins/prosthetic-conscience/tools/internal/secrets/secrets.go`
(`p.Class`, a secret pattern's class), another concept in another plugin.

```
I/capture/capture.go 2067            (deleted, III.11)
I/cli/lens/mint.go 93,95             (rewritten / deleted, III.6)
I/graph/graph.go 223                 (untyped %s — Word + test, III.5)
I/record/motion.go 353               (PETITION class — no change)
I/seatprobe/build.go 166             (g.Class, boards.go gap word as --class — III.10 rule 4; wider .Class sweep)
I/seatprobe/build.go 260             (m.Class, a petition's class — no change)
I/record/replay.go 537,540,689,697   (deleted / rewritten, III.5)
I/report/assemble.go 863             (untyped grade(any) — Word + test, III.5)
I/view/view.go 326,507,517,527       (typed — compiler; 326 keeps `undefined`, III.5)
```
</details>

<details><summary>C14 — fixture class words (§III.10)</summary>

Literal: `git grep -h -o -E '"--class", "[^"]+"|--class [a-z0-9_-]+|Class: *proto\.String\("[^"]+"\)' -- '*.go'`
→ 304 lines, 109 files, 42 words (`~/.claude/scratch/relocations/audit2/fixtureclasses.txt`).
Helper-carried: `git grep -n -E '^func [a-zA-Z]+\([^)]*\bclass\b[^)]*\)' -- '*.go'` → `mintGap`,
`mintEvent`, `liveMint`, `migrate`'s builder; `git grep -h -o -E 'mintGap\(t, [^,]+, [^,]+, "[^"]+"'` →
68 calls, 22 words (`boardminter_test.go`'s `read-surface` call is new since `61107ba4`); fixed `"x"` in `dispatch_test.go:48`, `worklist_test.go:29`; word list
`difftest/fuzz_test.go:282`; fixed `--class fuzzcls` at `releasegate/fuzz/fuzz_test.go:1851,1992` (`fuzzClasses`, `:6221`, is staging-only); `Class: proto.String(class)` at
`view_test.go:464`. Seed's default: `recordtest.go:121–126`.
</details>

<details><summary>C15 — event ids migrate carries: <code>git grep -n -i 'events\.id' -- '*.proto'</code>, every <code>int64</code> field of <code>record.proto</code> and of <code>record.Plan</code> and its parts, every operand of a line that reaches <code>Plan.Why</code>, and the prose that says how migrate treats the numbering or the bytes (<code>git grep -n -i -E 'numbering|renumber|carried as printed|migrates as printed|byte for byte|byte-for-byte|byte-identical'</code>)</summary>

The two tables in §III.8a are this census: `Dispatch.pin` (the proto's one `int64`), `Finding.about_ref`
(`record.proto:1017–1019`), `Plan.head`, `StaleArea.pin`, `PlanBlocker.subject` of kind
`contradiction`, `RemandOwed.direction` (seat prose); the `why` producers, every operand classified —
`dispatch.go:284–431` (its inline `Sprintf`s at `:322,337,341,346,350,352,359,361,365,369,394,398,431`,
`benchReadiness` `:1036–1054` through `:391`, the blocker `why` funcs through `:416`),
`retirement.go:67–75,110–138,206–216`, `passblockers.go:118,124,203,237,297,328`, `impasse.go:60`,
`dispatch.go:1192`; one bare `%s` of seat prose (`dispatch.go:365`), every other seat word `%q`.
`passblockers.go:268,273` are item and refusal wordings, not why producers. The producer census was
re-checked across the 12 producer commits since plan storage began (`ca32752a`): every seat word `%q`,
the remand wording unchanged since `dd15137f` (`~/.claude/scratch/relocations/audit5/producer-diff.txt`;
§III.8a). The prose carriers in `migrate/` are listed and classified in §III.8a (the `byte for byte`
hits included); the hits outside it are other senses (audit4's `c15-prose.txt`).
Measured on the 28 archives: stored plans in `final` (3), `n1` (3), `m28` (20), `m29` (17); `class_new`
events in b, b3, b6, b8, b9, quadratic-formula, m23 (373, before any plan), m28 (564, 572) and m29 (566,
before two plans that name no moved id), and `class-new` in five jsonl runs; 9 plans move, all m28's; no
stored line holds a quoted span or a remand line.
</details>

<details><summary>C16 — tests and non-Go carriers of the registry and coining (§III.5's two commands) — 62 + 443 lines at <code>b283d1e8</code></summary>

Every hit classified in `~/.claude/scratch/relocations/audit3/rev3/widesweep-classified.txt` (D deleted
182, C changes 97, N another concept 159, G regenerated 4) and `g4.txt`. Changes (I/ = `…/tools/internal/`,
Z/ = `…/tools/releasegate/fuzz/`):

```
F/tests/simulator/debate.test.mjs 408,1033          (§III.12)
scripts/universe-stream.py 25                       ("class" leaves INTERESTING)
scripts/universe.sh 519–539                         (block deleted)
scripts/agentgen/src/{fragments/lens-audit-and-mint.md 8, red-lens-voice.md 13, skills/adversarial-audit/SKILL.md 28} (§III.12)
I/cli/blue/edit_fuzz_test.go 18                     ("the class the registry calls silent-no-match-probe" → "a verification_scope_blindspot")
I/cli/boardbuild_test.go 47                         (the "Staging a class registry here…" sentences go; the bare-directory rule stays)
I/cli/cli_test.go 609, 823–866                      (staging comment; TestClassNewCoinsTheSlugInClass deleted); newRun moves here from the deleted fixtureclasses_test.go, unstaged
I/cli/m23contradiction_test.go 100, I/record/reopenedm19_test.go 44  ("inputs and class registry" → "inputs")
I/cli/referencechecks_test.go 25,45–46,56–60,119–165,231–234  (the --class row checks against the GapClass enum; needsRegistry, stageClassRegistry and TestAnUnreadableClassRegistryIsRefusedRatherThanIgnored deleted)
I/cli/refusalshape_test.go 230                      (the "use a registry slug" row names deleted replay.go:696 → the enum refusal's head)
I/cli/seatprobe_fixture_test.go 104–108, I/reportproj/mintbudget_test.go 46–48, Z/registerbeforeappend_test.go 47–49  (staging dropped)
I/cli/unknownclass_test.go 13–55                    (§III.6: the near miss and the 30 words)
I/cli/verbguardrace_test.go 224–244                 (coin race deleted)
I/diagnostics/manual_test.go 22,28,33,121–127       (the class group → a two-word path that stays, e.g. `show changes`)
I/difftest/{contract_test.go 94,308; golden_test.go 164,169–176,190,233–242; harness_test.go 63; scenarios_test.go 19,58–59}  (§III.10)
I/record/classmaterial_test.go 16–18,83–106,208–226 (MODIFY: liveRegistryRun → a lens run with no registry; slugs → ClassFor; its four registry/coining tests deleted)
I/record/redvoice_test.go 15,34,50,69,85,102         (the same helper, spec_underspecification)
I/record/replay_test.go 59,605–775                  (TestValidateClassRegistry deleted; class-registry.json leaves the non-shard example list)
I/record/legacyformat_test.go 75                    (the same example list)
I/record/writelookups_parity_test.go 122–123        (the ClassCoinedInRun assertion goes)
I/record/writerace_test.go 70–71,225                (coin racer deleted)
I/record/migrate/classmaterial_test.go → gapclass_test.go 180–211,241,275–284,367  (§III.8)
I/record/migrate/publish_test.go 30,214,425–457     (§III.8)
I/record/migrate/reproduceanchor_test.go 153,158–160 (translateStagedRegistry call and comment)
I/seatprobe/boardclasses_test.go 11–34              (deleted: BoardClasses is []GapClass and ShippedClasses is gone, so it asserts what the type guarantees)
I/seatprobe/separatedbuild_test.go 35               (the staged-registry item leaves the MEASURED block)
I/setup/unresolvablerun_test.go 27                  (comment: a "cannot resolve the record directory" reason)
Z/fuzz_test.go 117–119,215,345–351,1802–1851,1992,3683–3694,4603–4605,4680–4682,6217–6221,6261–6270  (§III.10)
```
No change (N): petition `class` (`promptverbs_test.go:958`, `debatejs/petition_test.go`, `enums_test.go`,
`termination_test.go`, `fuzz_test.go:894–972`, `replay_test.go:934`), the terms registry (`vocabulary_prose_test.go`,
`manualwords_test.go`, `descriptions_test.go:68–112`, `terms_test.go`, `scripts/vocabdoc`, `check/gates.go:262`),
the protocol registry (`dashboard_serve_test.go:112`, `promptverbs_test.go:744`, `archaeology`, `check/gates.go:215`,
`pluginparity/main.go:308`), migrate's `Registry`, `protoregistry`, the narrowed-view map, the SQL driver
registry, agentgen's fragment registry, `r.coin` and other coin flips, `cli_test.go:1067` (mint without `--class`
stays refused), `universe-board.py:68` (reads the column's word), `scripts/secretsgen/gitleaks/gitleaks.toml:1676,2213`
(a container registry in a token rule, new since `61107ba4`).
</details>

**Rendered, not grepped.** `mint --help`, `class --help`, `class new --help`, `class list --help` and
the `mint --class nope` refusal were rendered from the built `be6c110d` binary into
`~/.claude/scratch/relocations/gapenum/help-*.txt` (quoted in §I and §II). §V.3 renders them again
after the change.

**Census size:** 16 queries; ≈ 300 tracked files hit outside `plans/`; ≈ 200 change (≈ 110 by the
mechanical fixture pass, ~35 regenerated, the rest by hand), 14 law files, 9 code/tool files and 2 help documents deleted, the
remainder verified no-change above.

### III.15 Interactions

- **#1452 (law retired).** This PR deletes `HarvestClasses`, its capture output lines and the 14
  `law/proposed/class-*.md` — the directory's only files, so `law/proposed/` leaves the tree (git
  tracks no empty directory). Both readers handle that with no change: `scripts/lawqueue/main.go:96–99`
  reads a missing `law/proposed/` as "nothing has been harvested yet" and returns (its `check` gate
  stays green), and `HarvestPrecedents` creates it on write (`capture.go`, `os.MkdirAll` before its
  first file). `HarvestPrecedents` stays for #1452, as do setup's `<cwd>/law` mirror and its `law:`
  summary line. Either order merges cleanly; if #1452 lands
  first, its deletion of `capture.go`'s harvest block makes III.11's capture edit a no-op.
- **#1454 (scorecards from the archive).** Untouched: `WriteScorecards` and `feov-memory/*-scorecard.md`
  stay. A cross-run `unclassified` rate (F4) is #1454's to add, reading the archived telemetry.
- `protocol-class-registry.md` and its gates are another concept and are not touched.

## IV. Risk & Mitigation

| # | Risk | L | I | Mitigation cost | Mitigation (step) |
|---|---|---|---|---|---|
| R1 | An archived run fails to migrate (an unmapped slug, a dropped column, the D7 check on a catch-all mint without a reason) | Med | High — the census test is red | Low | `TestEveryArchivedRunMigrates` over 28 tarballs; `TestEveryArchivedClassLandsOnAValue`; measured: the one catch-all mint carries a `mint_reason` (III.8, §V.5) |
| R2 | A migrated gap changes materiality | Certain for 17 pre-field gaps (F9) | High — an archived board reads differently | Low | D5 keeps every recorded stamp (m28 test compares `class_material` per `gap_id`); the 17 moves are ruled residue, listed per run in §III.8, each pre-field mint a `StatedFill`, b6's PASS records its two as `migration_admitted_gap_ids`; `TestAPreFieldMintTakesItsNearestValuesDefault` pins one |
| R3 | The enum-valued facet breaks the schema build (FK order inside the DDL transaction) | Med | Build/test break, loud | Low | `enumTables` orders the referenced table first; `schema.sql` golden (III.4) |
| R4 | `enumhelp` change alters other enums' help | Low | Golden churn misread as intended | Low | other enums render byte-identically; `manual` golden diff read line by line (§V.3) |
| R5 | Help/constitution size: ~30 values × ~450 chars ≈ 13 KB added to `mint --help` and to each of 7 lens constitutions (~100 KB each today) | High (certain) | Medium — per-sitting context | Low | F3 accepts it; definitions held to one sentence; the refusal carries the near miss and the 30 words only (F7) |
| R6 | Definitions trip the vocabulary gate (`TestVocabularyProse` bans variants such as "the debate") | Med | Gate red | Low | run the gate early; Table A already avoids run-voice words |
| R7 | The fixture pass changes what a test asserts (a placeholder becomes an `always` class and masks severity) | Med | A test passes for the wrong reason | Med | III.10's rule keeps every placeholder `by_grade`; `ClassFor` ties a hand-set stamp to its class; §V.9's "stamp `always` unconditionally" mutation |
| R8 | A carrier still teaches coining (agent prose, prompt, docs) | Med | Seats look for a verb that is gone | Low | §V.8 sweep; C4/C5/C11 re-run; rendered constitutions grepped |
| R9 | The universe carries published content, not this checkout | Med (it has happened) | False pass of §V.7 | Low | §V.7 checks the installed binary's `--schema` equals 35, an epoch no published release carries; the installed `red-lens-logic.md` is grepped for the deleted verb |
| R10 | Seat environment leaks into tests | High in an agent session | Reads as a code defect | Low | every §V step starts with the `FEOV_*` unset |
| R11 | Smoke lenses never choose `unclassified` | High | Criterion 7 unmet by the run alone | Low | §V.7 drives one catch-all mint by hand on a copy of the finished run |
| R12 | `by_severity` changes from an object keyed by grade to `[{grade,count}]` (D8) — a seat reading the object breaks | Low (the help has always documented the array; `dashboard` reads the typed row) | A bench `jq` returns null | Low | the constitutions inline the regenerated help; `TestTelemetryHelpTreeMatchesAnEmittedLine` pins help = bytes |
| R13 | The stored-plan translation misses an id a `why` line prints, or mistakes a count for one | Med | High — a migrated plan names the wrong event, or a run fails to migrate | Low | A moved id outside every id phrase is a refusal, so a miss is loud (§III.8a step 4); a phrase word counts only as a whole token; `TestAStoredPlanCrossesARenumbering` checks each translated number against the event both databases hold there, not against `newID`; a count equal to a moved id is refused, loud — residue, none on any archive held (largest 7 vs moved ≥ 565) |
| R14 | Numbers in seat prose may be stale after any migration that drops an event; no list of shifts is kept (F13). This PR's `class_new` drops add m28 23, m23 8, quadratic-formula 3, b8 2, b, b3, b9 1 each, b6 and m29 0 (jsonl runs number no event) | Certain on those runs | Low — a reader following "head 579" lands on another event | Low | Accepted residue (§III.8a, F13); the source tarball under `run-archive/` keeps the original numbering; a bench direction and a quoted claim in a stored plan are carried as written too (F12) |
| R15 | A translated plan differs from the relay the workflow journal recorded, so `capture/dispatchparity.go`'s `relayDepartures` (`:248`) would report a departure per translated plan | None today: no `run-archive/*.tar.gz` carries `trajectories/`, so no migrated run has a journal to compare | Low — a dispatch-parity FAIL on an archived run whose relays name the source numbering | Low | Accepted residue (§III.8a); the relay was of the source numbering, the source tarball keeps both |

## V. Verification Plan

`W=/home/gblock_ctoforaday_com/projects/special-circumstances/.claude/worktrees/bridge-cse_0154Z23DoVFryzFDGWWcWHxj`,
`T=$W/plugins/frank-exchange-of-views/tools`, scratch `~/.claude/scratch/relocations/gapenum/`. Every
step that runs Go or `feov-record` begins with `unset $(env | grep -o '^FEOV_[A-Z_]*')`. Long output
goes to a file; read the fail count, not the exit code. Check the box load before believing a red or a
timing.

1. **Generators.** `go -C $W/scripts run ./protogen && go -C $W/scripts run ./schemagen && go -C $W/scripts run ./agentgen && go -C $W/scripts run ./vocabdoc`
   then the `check` gates covering them (`agentgen -check`, `schemagen -check`, the protogen sha, `TestVocabularyProse`) → green. `feov-record --schema` → 35 (`requirements.json`'s `eventSchema`, one above `main`'s 34 after #1522). · re-armed by: `record.proto`,
   `requirements.json`, `agentgen/src/**`, `terms.json`.
2. **FEOV tests, uncached.** `go -C $T vet ./... && go -C $T test -count=1 ./... > ~/.claude/scratch/relocations/gapenum/test.out 2>&1; grep -c -- '--- FAIL' …/test.out`
   → 0; `grep -c '^ok' …/test.out` equals the package count. Includes `recordpb/gapclass_test.go`,
   `TestAVocabularyFollowsTheVocabulariesItReferences`, `TestClassForHasTheMaterialityItNames`,
   `TestUnclassifiedRequiresAReason`, `TestAMintTakesItsClassMaterialDefault`,
   `TestALowGradePlaceholderMintDoesNotHoldTheGate`, `TestMintHelpListsEveryGapClass`,
   `TestAWrongClassIsRefusedWithTheNearMissAndTheWordsOnly`, `TestTelemetryHelpTreeMatchesAnEmittedLine`,
   the three class-word renderer tests (§III.5), the telemetry tests, `TestSetupReadsNoFeovMemory`,
   the migrate tests (`TestAStoredPlanCrossesARenumbering`, `TestAStoredPlanNamingMovedEventsIsTranslated`,
   `TestAPreFieldMintTakesItsNearestValuesDefault`, `TestMigrateKnowsTheVerbsRemandPrefix`,
   `TestAnArchivedRunStoresOnlyThePlansItStored`), `TestARemandLineCarriesItsDirectionAsWritten`, the
   zero-mint epoch case of
   `TestTelemetryHelpTreeMatchesAnEmittedLine`. · re-armed by: anything under `$T`.
3. **Goldens, read.** `UPDATE_GOLDENS=1 go test -count=1 ./internal/difftest/... ./internal/record/recordsql/... ./internal/record/recordpb/...`
   in `$T`, and `go -C $W/scripts run ./golden -review` for the prompt goldens; then `git -C $W diff --stat -- '*.golden' '*.sql' '*testdata*'`
   and READ every changed line: the only changes are the `class` group's disappearance, `--class`'s
   enumerated block, underscore class words, the telemetry shape, `schema.sql`'s `enum_gap_class`/`mint`
   FK/check and `class_new` removal. Render, before and after, one at a time: `mint --help`,
   `inquest telemetry --help`, the lens top-level `--help`, `mint --class report-voice …` (near-miss:
   names `report_voice`, lists the 30 words, carries no definition) and `mint --class unclassified …`
   without `--reason`; and one `inquest telemetry` line beside its help tree (`by_severity`/`by_class`
   arrays of words). · re-armed by: any help string, enum, golden, `view.go`'s wire types.
4. **Setup from an empty project, built binary (driveable).** In `~/.claude/scratch/relocations/gapenum/proj`
   (fresh `git init`, no `feov-memory/`), with the binary under `bin/` and `requirements.json` beside it:
   `feov-record --seat-id operator setup <proj>/research/probe --model haiku --judgment-model haiku`
   → exit 0, no `class registry` line; then as `red-lens-logic`: `register`, a `finding`, and `mint
   --class false_universal …` → recorded with `class_material: always`; `mint --class unclassified …`
   without `--reason` → refused naming the check's `why`; with `--reason` → recorded `by_grade`;
   `inquest telemetry` as `judge` (an `inquest` verb) → the epoch's `unclassified_rate` equals the catch-all's share of that epoch's mints, and `by_class` lists both values. · re-armed by: setup,
   mint, view.
5. **migrate on archived runs.** `go -C $T test -count=1 -run 'TestTheGapClassMigratesOntoTheEnum|TestEveryArchivedClassLandsOnAValue|TestEveryArchivedRunMigrates|TestAStoredPlanCrossesARenumbering|TestAPreFieldMintTakesItsNearestValuesDefault|TestAStoredPlanNamingMovedEventsIsTranslated|TestMigrateKnowsTheVerbsRemandPrefix|TestAnArchivedRunStoresOnlyThePlansItStored|TestEveryProseRewriteInTheArchiveIsListed|TestAStoredPlanMigratesAsPrinted' ./internal/record/migrate/` and `go -C $T test -count=1 -run 'TestARemandLineCarriesItsDirectionAsWritten' ./internal/record/`
   → ok, 28 subtests in the census, zero refusals in each. Then driveable on the real archive: extract
   `run-archive/2026-10-18_is-91-prime-m28.tar.gz` into scratch, `feov-record migrate` it with the built
   binary, and read the manifest's `class_translations` and `stated_fills` and `SELECT class,
   class_material, count(*) FROM mint GROUP BY 1,2` on the result (`file:` URI, `mode=ro`) against the
   same query on the source; then `SELECT e.id, p.printed FROM plan p JOIN events e ON e.id = p.event_id`
   on both, and read event 604's pair by eye: `head 579` → `head 578`, `pin 507` kept, `edit 579` →
   `edit 578`, a new key. The same on
   `2026-09-11_is-91-prime-b6`: `G2`/`G7` `always`, the PASS's `migration_admitted_gap_ids` holding
   both, a `stated_fills` row per mint. · re-armed by: `migrate/**`, `record.proto`, `record/planstore.go`,
   the `why` producers (`retirement.go`, `passblockers.go`, `dispatch.go:284–431,1036–1054`, `impasse.go`).
6. **Release-gate fuzz.** `cd $T && FEOV_RELEASE_GATE=1 go test -count=1 ./releasegate/fuzz/` → ok;
   the fuzz's own coverage table shows `unclassified` minted and refused-without-reason at least once,
   and its `classes:` line every `always`/`never` value landed (F16). · mutation: drop one value from
   the deliberate drive → the enum census and the named-drive gate both fail.
   If red, run the same on `origin/main` before blaming the branch.
7. **Universe smoke — mints with no feov-memory, one catch-all.**
   ```
   $W/scripts/universe.sh build
   U=~/.claude/scratch/universe/plugins/cache/special-circumstances/frank-exchange-of-views
   ls $U/*/bin/feov-record && $U/*/bin/feov-record --schema            # → 35: no published release carries it, so this is the checkout
   grep -c 'class new' $U/*/agents/red-lens-logic.md                        # 0
   MODEL=haiku JUDGMENT_MODEL=haiku LANES=1 $W/scripts/universe.sh run "is 91 prime" --mint-budget 3 --k-max 2
   ```
   (stream-json via `universe.sh`, never `--output-format json`; `universe.sh watch` read-only; confirm
   the universe bills the subscription before launching — memory "universe billed API credits".)
   `universe.sh` keeps `$WORKDIR/src` between runs, so a `feov-memory/` an earlier build staged
   survives the deletion of the staging block: BEFORE `run`, `rm -rf ~/.claude/scratch/universe/src/feov-memory`
   (scratch, inside the universe's own work directory). Then: `test ! -e ~/.claude/scratch/universe/src/feov-memory`; the run log has no `staged feov-memory`
   line; the run's own `.bin/feov-record` (not a PATH lookup) reports ≥1 `mint` and zero class
   refusals in the transcripts. If no lens chose `unclassified`, copy the finished run directory and,
   with its `.bin/feov-record` as `red-lens-logic`, `register` and mint once with `--class
   unclassified` without and then with `--reason`; record which path satisfied the criterion.
   · re-armed by: `universe.sh`, setup, mint, agentgen output.
8. **Carrier sweep.** Success criterion 8's `git grep` (with its `\b` and its `scripts/rulesweep` exclusion) → nothing;
   `git -C $W grep -n -E "\b($(cat ~/.claude/scratch/relocations/gapenum/slugs.re))\b" -- plugins/frank-exchange-of-views/agents plugins/frank-exchange-of-views/skills scripts/agentgen/src plugins/frank-exchange-of-views/tools/internal/terms`
   → nothing (hyphenated class words are gone from seat-facing text). The comment sweep of §III.5 →
   only its listed no-change hits; the two test/non-Go sweeps of §III.5 → only C16's N hits. Re-run
   C1–C16 and compare to the no-change lists in III.14.
   · re-armed by: any edit to a carrier, prompt, constitution source, doc or comment.
9. **Delete-the-row / delete-the-call** (each restored after): delete one `GapClass` value's
   `(material_default)` → schema build refuses (partly annotated); delete one `(neighbor)` →
   `TestEveryGapClassNamesItsNeighbour` fails; delete one row of `migrate/gapclass.go` →
   `TestEveryArchivedRunMigrates` fails for that run; delete the `Mint` `(check)` →
   `TestUnclassifiedRequiresAReason` fails; make `stampClassMaterial` stamp `by_grade` unconditionally →
   the m28 migrate test and `TestAMintTakesItsClassMaterialDefault` fail; **make it stamp `always`
   unconditionally → `TestALowGradePlaceholderMintDoesNotHoldTheGate` (a `by_grade` fixture) fails**;
   exclude-nothing in `class_repeat_rate` → `…CountsOnlyTheCatchAll` fails; drop the dependency-first
   visit in `enumTables` → `TestAVocabularyFollowsTheVocabulariesItReferences` fails (and record opens
   fail at the `INSERT`); restore `severityObject` on `by_severity`, or fill `by_class[].class` with the
   number → `TestTelemetryHelpTreeMatchesAnEmittedLine` fails; delete `UnclassifiedRate` from
   `newMintJSON` → the same; pass `g.Mint.GetClass()` unconverted at `assemble.go:863` or
   `graph.go:223` → `TestTheOpenGapRowNamesItsClassWord` / `TestTheGapNodeNamesItsClassWord` fails;
   make `Refuse` append the `Menu` unconditionally → `TestAWrongClassIsRefusedWithTheNearMissAndTheWordsOnly`
   fails; return `ClassFor(cm)` = `spec_underspecification` for every `cm` →
   `TestClassForHasTheMaterialityItNames` fails; drop the `why` translation in `TranslatePlanIDs` →
   `TestAStoredPlanCrossesARenumbering` fails at event 604; drop the pin translation → the same; drop
   the re-key → 9 refusals on m28; drop the moved-id-outside-a-phrase refusal →
   `TestAStoredPlanNamingMovedEventsIsTranslated` fails; drop the remand-tail rule, or search the
   remand prefix inside quoted spans, or read numerals as bare digit runs → the same test fails (the
   remand line, the `(contradiction 412)` line, the `G-7b637cca` line), and bare digit runs also refuse
   14 m28 lines; reword `RemandDirectionPrefix` → `TestMigrateKnowsTheVerbsRemandPrefix` fails with its keep-the-old-wording message; print the direction with `%q`, or before other text → `TestARemandLineCarriesItsDirectionAsWritten` fails; match a phrase word as a suffix instead of a whole token → `TestAStoredPlanNamingMovedEventsIsTranslated` fails at the `spin 412` plan; translate a numeral that did not move (every phrase numeral `n` → `n−1`) → `TestAnArchivedRunStoresOnlyThePlansItStored` fails on `final`, `n1` and `m29`; restore `shippedDefault` →
   `TestAPreFieldMintTakesItsNearestValuesDefault` and sleeper's `citation-support-mismatch` row fail;
   fill `by_class` by plain assignment from the nil getter → the zero-mint case of
   `TestTelemetryHelpTreeMatchesAnEmittedLine` fails. Any that stays green is a gap closed before done.
   · re-armed by: each named mechanism.
10. **All local gates.** `go -C $W/scripts run ./check > …/check.out 2>&1` → every gate PASS, fail count
    read — the `node:` simulator gates included (`debate.test.mjs` after the `withKey` rename); `-list` no longer lists `classgen`. A workflow edit (`hooks.yml`) needs CI's `actionlint`
    before a green claim.
11. **Auditor gate.** `/prosthetic-conscience:plan-audit plans/gap-class-enum.md` → PASS; F1–F14 are
    ruled.

## Forks — rulings (gblock, 2026-10-18; F13–F14 2026-10-19)

- **F1 — Membership (§III.2): RULED, Tables A–C as proposed, the ⚑ calls included** (gblock, 2026-10-18: "wholly agreed").
- **F2 — Add both `unsourced_claim` and `invalid_inference`, `by_grade`: RULED, add both.**
- **F3 — Render all four fields (value, materiality, definition, nearest neighbour + distinguisher) in `mint --help` and, through `agentgen`, in every lens constitution: RULED, full.** The ≈ 13 KB per lens constitution (R5) is accepted.
- **F4 — `unclassified` rate per epoch now (D8); the cross-run rate goes with #1454: RULED.**
- **F5 — D7's residue: RULED, accepted.** The nearest value lives in `--reason` prose; the check
  requires only a non-empty reason. No `--nearest` field.
- **F6 — D8's help/wire mismatch: RULED, fix both in this PR.** `by_severity` and `by_class` alike;
  one test (`TestTelemetryHelpTreeMatchesAnEmittedLine`: the rendered help tree against an emitted
  telemetry line) covers every `telemetryLineJSON` field with a custom `MarshalJSON` or a proto enum.
- **F7 — R5's refusal: RULED, near miss + the 30 words only.** Full definitions stay in `mint --help`
  and the constitutions; the refusal and the help need not share the block (D9).
- **F8 — Stored plans across a renumbering: RULED, translate inside migrate** (round 2). Each stored
  plan whose printed head or pins a dropped event renumbers is rewritten to the new ids with its key
  recomputed, so the migrated record holds the equivalent state; general over any event that writes
  none, not class-specific; no field added to the live record (D10, §III.8a).
- **F9 — Pre-field coined mints: RULED, "clean up implementation and choose nearest for historical."**
  Each historical slug takes its nearest value (Table C's target) and that value's `material_default` by
  the live rule; `shippedDefault` / `ShippedMaterialDefaults`' fill path is deleted. Residue stated: 17
  archived gaps move `by_grade` → `always` (§III.8 lists them — 8 runs by measurement; the ruling's
  "9 runs" covers the same 17); `TestAPreFieldMintTakesItsNearestValuesDefault` asserts one (b6 `G2`).
- **F10 — The empty tally: RULED, `[]`, never `null`**; the help-agreement test also decodes a
  zero-mint epoch line (§III.9).
- **F11 — The refusal path rewrite: RULED, delete.** `migrate.go:170–176`'s `b.atTo` over refusal
  text and `TestARefusalNamesToNotTheBuild` go; `build.atTo` stays for the warnings (`migrate.go:69`)
  and `atToError` (`publish.go:293`). Residue stated: no refusal names a run path today; a later one
  that does publishes the build path into `inputs/migration.json` and nothing fails — no lens reads
  the manifest, and migrate runs outside a debate (§III.8).
- **F12 — The remand direction in a stored plan's `why`: RULED, carry as written.** The span after
  the fixed `on the ruling's direction: ` prefix is seat prose, neither translated nor refused, as a
  `%q` span is (§III.8a step 4).
- **F13 — `events_dropped`: RULED, delete the manifest field** (gblock, 2026-10-19: no reader, no report
  destination, no seat purpose). R14 states the residue on its own: numbers in seat prose may be stale
  after any migration that drops an event; no list of shifts is kept (§III.8a, R14).
- **F14 — `remandDirectionPrefix`: RULED, `record` exports the prefix `dispatch.go:365` formats with**
  (`RemandDirectionPrefix`); `TestMigrateKnowsTheVerbsRemandPrefix` asserts migrate's copy equals it,
  failing with "keep the old wording for archived bytes and add the new one" (§III.8a step 4).
- **F15 — m29's coined `summary-scope-omission` → `reader_channel_mismatch`: RULED** (gblock, 2026-10-19).
- **F16 — Fuzz class coverage: RULED** (gblock, 2026-10-20): random mints draw from the `by_grade`
  GapClass values (materiality unchanged); each `always`/`never` value is minted once per sweep,
  deliberately, as closeGap's `moot`; `unclassified` keeps its landed/refused drive. The enum splits 14
  `by_grade` (the catch-all excepted) and 15 `always`/`never` (14 `always`, `structure_noncompliance`
  `never`), read off the enum; the ruling stated 15 and 14 (§III.10).
- Decided and disclosed (reversible, in scope), round 5: m29's coined `summary-scope-omission` →
  `reader_channel_mismatch` (Table C; a membership call F1 did not cover, because m29 was archived after
  it — its recorded `by_grade` stamp is kept, so the call moves no materiality); an id phrase word counts
  only as a whole token, so `ahead 5`/`spin 5` are not ids (§III.8a step 4); `RemandDirectionPrefix` is
  the exported name (0 hits; extends `remandDirection`); `TestAnArchivedRunStoresOnlyThePlansItStored`
  holds byte identity where nothing a plan names moved and the masked comparison elsewhere (§III.8a);
  harness.mjs's `keyed` is renamed `withKey` with `debatejs.Keyed` (D10).
- Decided and disclosed (reversible, in scope), round 4: a numeral is an all-digit
  `[A-Za-z0-9_-]+` token, so a content id's hex digits are never read as an id (§III.8a step 4);
  `planIDPhrases` and `remandDirectionPrefix` are migrate's own table, because they describe archived
  bytes; a `"` that opens no Go-quoted span outside the remand tail is a refusal; the exported key rule
  is `Plan.WithKey` (not `Keyed`, which `Answer.Keyed` already means otherwise), and `debatejs.Keyed`
  takes the same name; `revisionEntry` becomes the shared `toNoEvent`.
- Decided and disclosed (reversible, in scope), round 3: the plan's `why` ids are read through a closed
  phrase table with every miss a refusal, and a quoted claim is left as prose (§III.8a step 4);
  `record.keyed` exported for migrate's re-key (named `Plan.WithKey` in round 4); seat prose naming a moved event number is
  accepted residue (§III.8a, F13); every pre-field mint, not
  only the 17 whose default moved, is a `StatedFill` (which moved is computable only against the deleted
  table); b6's no-admission test is inverted by F9 and its rule moves to a seeded source (§III.8);
  `seatprobe/boardclasses_test.go` is deleted rather than reduced to what the type guarantees.
- Decided and disclosed (reversible, in scope): underscore spelling (D2); `material_default` as an
  enum-typed facet, `neighbor`/`distinguisher` as help-only options (D3, D4); `Mint.class_material`
  stays stored (D5); `unclassified` kept despite two same-sense hits (D6); the reason requirement as a
  database check (D7); `by_class` becomes a tally list and `class_repeat_rate` excludes the catch-all
  (D8); both tallies as arrays of plain structs rather than an override in `jsonshape` (D8 (a));
  `spec_underspecification` as the fixture placeholder (§III.10); `class-registry-seed.md` stays as a
  design record.
