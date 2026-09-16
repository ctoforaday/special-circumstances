# FEOV lens bar — materiality by class, per-seat retirement, all seven lenses cast

TinySpec, 2026-09-11, revised after plan audit rounds 2–5; implementation APPROVED by gblock after round 5, with the round-5 local gaps carried as required tests (§V, "Carried from audit"). Rulings: gblock 2026-09-11 (items 1–8, the blue
standard, the report-voice ruling, the Migrating exemption, the two round-3 forks, decisions in III.10). Worktree
`plan/feov-lens-bar` at main `1503f27b`; uncommitted. `origin/main` is now `1ef6338f`. Since the base it adds the
b6 archive and PRs #927 and #928, which changed 34 FEOV files: `dispatch.go` gains `DispatchGroups`,
`unopenedChairSitting` and `DispatchStands` (+177 lines); `capture/dispatchparity.go` is rebuilt on
`record.DispatchGroups`; `sitting.go` gains the chair's "register for this sitting" item; `merge/dispatch.go` and the
help pages move. The design was re-checked against that tree (extracted at
`~/.claude/scratch/feov-lens-review/audit/r4-om/`) and holds: the parity audit's grouping is the one fork (b) extends
(III.4), and `DispatchStands` compares recorded rows with the plan, which the retirement fold does not change.
**Every `file:line` below is verified against `origin/main` `6d89b380` (re-verified at the phase-0 rebase, 2026-09-15), the tree implementation is rebased on.** Between `1ef6338f` and `6d89b380`, main renamed the chair's role and CLI package `merge` to `chair` (#847), and `6a67e0bf` made the plan's `parties` and `docket` print `[]` and gave the chair's work list severity-based `Material` and `Stranded` items; §II states the result, and the design is unchanged by either. Paths are relative to `plugins/frank-exchange-of-views/`
unless they start with `feov-memory/`, `scripts/` or `law/`.
**[R]** = evidence from FEOV run records; **[P]** = extrapolated from plan-auditor verdicts.
Evidence: `~/.claude/scratch/feov-lens-review/{lens-quant.md, gh-lens-issues.md, work/*.py, gap1-*.txt, gap2-sweep.txt}`,
`~/.claude/scratch/auditor-review/results/*.md`.

## I. Summary & Goals

The lenses cost the most where they catch the least. Trifles hold runs open, and closures are attested
against stand-ins.
- Round-era chair gates: 21 FAIL, 0 PASS. Roundless re-sittings: 27 sat, 2 minted. b5's chair readied the voice
  lens alone eleven times at head 144 (#911). [R §8]
- Lineage fallout: 20 of 105 gaps. `repaired_with_regression`: 8. [R §9.1]
- Proxy closures:
  - record-store G5 was closed on an exit status;
  - quadratic G22 on "legs failing: 0";
  - quadratic G24 on a list counted against itself;
  - quadratic G10 and is-91-b G2 on proofs run where they were authored rather than from the store. [R §9.2]
- Quadratic rounds 4–5 cost $104.59 for two trivial mints (#554). [R]

Goals:
- **G1** A gap names the decision it changes; otherwise it is a finding.
- **G2** Materiality defaults by class, carried on the record, and read through one definition by every reader
  and every surface that states or decides what holds the PASS gate.
- **G3** "Verified" means the object itself.
- **G4** A class fix is closed by the command that enumerates the class.
- **G5** Every re-sitting on already-read text must say why a fresh gap was missed.
- **G6** Per-seat retirement with a single re-arm, enforced by the dispatcher.
- **G7** Blue writes to the reader's question, and complexity must pay for itself.
- **G8** All seven lenses are cast by default.
- **G9** Red's report-bound text obeys the report-voice rules.

Non-goals:
- A "read changed text first" duty (gblock judged that finding biased).
- A new round cap. MaxEpochs and the no-progress valve stay (D2).
- Any change to impasse or the bench.
- Backwards compatibility. A pre-change run is brought forward with `feov-record migrate`. There are no compat
  shims, aliases, tolerance paths or read-time fallbacks, and shipped surfaces carry no prose about the old model
  (gblock).

## II. Technical Context (verified in this tree)

**The sweeps.** Each runs from `plugins/frank-exchange-of-views/`; outputs are saved in the scratch directory.
| # | Command | Lines | What it finds |
|---|---|---|---|
| S1 | `grep -rnE 'GradeMass\(\|MASS\[\|>= ?2\.0\|\bmaterial\b' tools --include=*.go \| grep -v _test` | 70 | threshold text in Go |
| S2 | `grep -rnEi "PASS is refused\|refuses? (a )?PASS\|PASS (is )?refused\|holds? the gate\|hold(s\|ing)? (the )?PASS\|pass_permitted\|PassPermitted\|permits? a PASS\|PASS requires\|blocks? (a )?PASS\|before a PASS\|gap(s)? (still )?open.{0,40}PASS" tools agents skills docs commands --include=*.go --include=*.md --include=*.js --include=*.proto \| grep -v '\.pb\.go' \| grep -v /testdata/` | 83 (43 outside tests) | every site that states what holds the PASS gate |
| S3 | `grep -rnE '\bg\.Open\b\|\.Open &&\|WHERE g\."open"\|WHERE "open"\|"open" AND\|g\."open"' tools --include=*.go \| grep -v _test \| grep -v '\.pb\.go'` | 26 | every code reader of gap openness, to find the deciders S2's wording misses |
| S4 | `grep -rnEi 'material' tools agents skills docs --include=*.go --include=*.proto --include=*.md --include=*.js \| grep -Ei 'medium\|mass\|grade\|severity\|2\.0'` | 46 | old-model carriers on surfaces no gate reads |
| S5 | `grep -rnEi '\bpins?\b\|head moved\|sat against\|sits? (again\|when)\|spot-check' tools/internal/cli/seat/help agents skills docs tools/internal/record/recordpb/record.proto tools/internal/record/dispatch.go` | 53 | readiness carriers (III.5) |

S1–S4 are classified line by line in III.1; S5 in III.5.

- **Materiality today is severity alone.** Every decider below compares current severity mass with 2.0:
  - `record/dispatch.go:139` (readiness, and `materialOpen` for `PassPermitted` at `:191`), over `openGaps` (`:503`, query at `:509`);
  - `record/refs.go:322-350` (the PASS gate, `requirePassClosesAllMaterialGaps`);
  - `record/convergence_refusal.go:41-62` (the FAIL refusal);
  - `record/recordsql/views.go:439,454,466` (`convergence_vs_verdict`);
  - `verify/verify.go:371`, check `pass-closes-all-gaps`, on the Go `Family` fold (`record/replay.go:147`). Its
    output reaches `run.md` (`report/assemble.go:1389`) and `feov-record verify`.
  - **The chair's work list, which S1 missed because it names no threshold.** `record/sitting.go:153-166`
    (`SittingOf`, role `chair`, which is `red-chair`: `roles.go:52`) adds a BLOCKING item "gap X is open and material — PASS is
    refused while it is" for every open gap whose current severity is medium or above, and "gap X is open and superseded — PASS is refused until its minter closes it" for every stranded gap (since `6a67e0bf`), and `Complete = !Blocked()` (`:241`). The gaps come from
    `workGapStatesOfRun` (`viewjson.go:757-797`), which selects `stranded` and computes `Material` from the severity mass, not the class, through `WorkJSONBytes`
    (`viewjson.go:927-938`). `cli/seat/help/verdict.md:9` tells the chair "`sitting.complete` is true exactly when
    nothing blocking is left". Today that agrees with the severity gate; it names no unraised contradiction and no lens condition, and nothing on it is class-aware.
  - **The work view's open-gap rows** (`viewjson.go:825`, `workJSONOfGaps`, into `WorkGapJSON`): the rows the chair
    lists at PASS. They carry no materiality today.
  - **The chair's docket affordance** (`record/available.go:112-136`, role `chair`, the loop at `:113`): it offers
    `motion docket file` on EVERY open gap that has no pending docket motion. A docket motion holds PASS until
    the bench rules it (`refs.go:370`; `dispatch.go:142-150`, `unruledDocket`), so docketing a gap that is not
    material creates a block that the gap itself does not.
  - S3's other 17 lines count or render openness without deciding the gate: `verify.go:199,552`,
    `graph/graph.go:205,219,265,271`, `view/view.go:273,380,468`, `report/docs.go:339`, `report/site.go:169`,
    `report/assemble.go:712`, `record/nearmatch.go:93`, `record/estoppel.go:290`, `record/replay.go:193`,
    `record/available.go:181`, `recordsql/views.go:473`.
  - **Not readers** (S1): ordering (`view/view.go:556`, `report/assemble.go:469`); the mass table and definition
    (`record/record.go:108,119`, `recordpb/facets.go:163`); subject prose (`seatprobe/boards.go:622-657`,
    `enums.go:318`).
- **Fresh** is `supersedes_count = 0` (`views.go:462-466`).
- **Required fields.** `(sql).required` on a proto field makes the column NOT NULL and is what the write path
  refuses on (`record.proto:178-180`). `recordpb/requiredfields.go` builds the refusal and appends the field's
  `why` (`because`, `:108-113`); `record/required.go:48` (`RequiredFields`) feeds the help's REQUIRED marks. A
  field the verb fills itself is declared with `seat.Supplies` (`cli/seat/seat.go:395`), which the vocabulary gate
  honours (`cli/vocabulary_test.go:162-185`).
- **The class registry:**
  - It lives at `feov-memory/class-registry.json` (39 classes, `{slug}` only).
  - Staged by `setup/setup.go:662-683`.
  - Read by `record/replay.go:470-570`.
  - Written for hand-built runs by `StageForRun` (`replay.go:579`). Its callers: `seatprobe/build.go`,
    `releasegate/fuzz/{fuzz,registerbeforeappend}_test.go`, `difftest/golden_test.go`,
    `{record,report,cli}/fixtureclasses_test.go`, `reportproj/mintbudget_test.go`, `cli/seatprobe_fixture_test.go`.
  - Hand-written registry JSON appears in `difftest/{scenarios,contract}_test.go` and
    `cli/referencechecks_test.go:136`.
  - Generated into tests by `scripts/classgen`.
  - `structure-noncompliance` exists only as `law/proposed/class-structure-noncompliance--2026-08-23_sleeper-service-plan.md`,
    and as a `ClassNew` event in the sleeper-service-plan archive.
  - `lawqueue` counts only `[PERSUASIVE]` holdings and has no "adopted" form.
- **Migration:** `record/migrate/migrate.go:21` copies the run's channels, including the staged registry, then
  replays the record through the current write path with `record.Migrating` set (`record.go:504`). At the verdict
  (`record.go:1172-1195`), two refusals are gated `!Migrating`: the FAIL convergence refusal (`:1181-1185`) and
  every-cast-lens-sat (`:1190-1194`). **`requirePassClosesAllMaterialGaps` runs under `Migrating` today**
  (`:1186-1189`). The migration source is opened with its own driver, never through `recordsql.Open`
  (`migrate/sqlitesource.go:14,63`). A word with no registry entry translates by identity: fields matched by
  NAME (`migrate/registry.go:27-44`, `identityBody` `:77-112`, `fillFields` `:119-139`). So a run written by a
  binary that already has a field replays that field as recorded, and a pre-change run lacks the column, so the
  field arrives absent. The event-schema epoch is compared once, at setup (`setup.go:467`).
  `record/store.go:82-95` refuses a former record FORMAT (event shards, no database) by content.
- **A pre-change database keeps its creation-time schema.** `recordsql.Open` (`recordsql/store.go:250`) calls
  `ensureSchema` (`:339`), which returns at once when an `events` table exists (`hasEvents`, `:365`), so tables and
  views are those of the binary that created the run. The older-run diagnoses, `olderSchema` and `olderColumns`
  (`recordsql/read.go:490,526`, advice at `:513`), fire only on the ERROR path of a body-table read or write. A
  view read that names a new column (`gap."material"`) fails first with SQLite's "no such column", which names
  neither the cause nor `migrate`.
- **Dispatch** is `PlanDispatch` (`dispatch.go:59-206`), which has three sources:
  1. a lens is ready when `pins[l] < Head` (`:115-121`). With `Head == 0` (nothing ingested) no lens is ready,
     and `allLensesSat := plan.Head > 0` (`:114`) keeps `PassPermitted` false;
  2. every open material gap engages its minter and `blue-respond`;
  3. docketed gaps engage the bench.

  `PassPermitted` needs every lens to have sat at the head (`:191`, `requireEveryCastLensSatAgainstHead` `:530`,
  which also refuses at head 0, `:562-563`). `Party` carries only `{seat_id, gap_ids}`, so `debate.js:847` tells
  every no-gap lens that "the head moved", whether or not it did.
- **The relay.** `debate.js` reads no record (`:338, :415, :485, :679`); it calls only `agent`, `parallel` and
  `log`. The plan reaches it as the chair model's copy, `chairEnv.plan` (`:884`), checked only for
  `Array.isArray(plan.parties)` (`:885`). The `PLAN` schema (`:444-466`) requires `head`, `parties`,
  `pass_permitted` and `ceiling`, and `seat_id` and `gap_ids` per party. `docket`, `why`, `max_epochs` and
  `epoch_limit_reached` are optional and read with fallbacks (`plan.docket || []` at `:927`). Go's `Plan`
  (`dispatch.go:22-39`) emits every field, with no `omitempty`. Capture's `DispatchParityAudit`
  (`capture/dispatchparity.go:12-29`) checks registrations against dispatch rows, not party fields. Capture already
  reads every envelope result from the workflow journal (`ReadJournal`, `capture.go:98-131`, called at `:1818`),
  so the relayed plans are there to compare. `PlanDispatch` builds `plan := Plan{Parties: []Party{}, Docket: []string{}}` (`dispatch.go:63`, since `6a67e0bf`): with
  nothing to say, `Parties` and `Docket` print `[]`, but `Why` is nil, so `--json` emits `null`
  for it. The fuzz relays the verb's JSON as a live chair does and patches only `parties`
  (`releasegate/fuzz/fuzz_test.go:581-599`).
- **"Has this seat sat" is one predicate.** `sittingFor` (`dispatch.go:260`) counts the seat's first register
  after a dispatch as its sitting for that dispatch, and `register` is every seat's first act
  (`cli/seat/help/register.md`). So a fact read from INSIDE a sitting sees that sitting as already sat. The
  readers: `PlanDispatch` (`:98`, `lensPins` `:461`), `benchSatFor` (`:151,478`), `exchangesOf` (`impasse.go:74-122`, via
  `:125`; its exported `Exchanges`, `impasse.go:39`, has no caller outside tests), the PASS gate's lens check
  (`:552`) — all read by the chair between sittings — and `owedSitting` (`dispatch.go:443`, read at `sitting.go:113`), read inside the
  sitting, where "not yet sat for" is the intended meaning (the register discharges it). #927 adds three more,
  none reading inside a sitting in the wrong sense: `DispatchGroups` (`dispatch.go:286`; `Sat` off the LAST row
  naming each party, read by the parity audit after the run and by the two below); `unopenedChairSitting`
  (`:339`, read inside the chair's sitting at `sitting.go:122-125` and by `RequireChairSittingOpened`, where it asks
  exactly whether the chair's CURRENT sitting has been opened); and `DispatchStands` (`:399`, the chair's
  `dispatch next`, which asks only whether anyone registered since the latest rows).
- **Default cast** is four areas: `record/cast.go:15` and `debate.js:563`. Concurrency is capped at about 2 (#788).
  computation, adversary and architecture were never seated in any archived run [R §1].
- **Report voice:**
  - The rules: `agents/blue-researcher.md:184`, `blue-synthesizer.md:109`, `red-lens-voice.md:11-17`,
    `adversarial-audit/SKILL.md:62`.
  - The tells: `reportvoice/tells.go`. They are advisory and wired only into blue (`cli/blue/voice.go`, `ingest.go:125`).
  - **Red's paths into `report.md`** are three (III.9 names each):
    - the risk matrix (`report/assemble.go:482-497`, `docs.go:121`), which lifts the lead sentence of each OPEN gap's
      `problem` and `required_fix`;
    - the source's note and the Bibliography: a labelled corroboration is a cited source (`record/citationid.go`),
      and `weaveCitations` (`report/assemble.go`, called from `docs.go`) prints a note `[^N]: <title>. <url>
      (accessed <date>)` per URL and PDF page, titled by the first label under it, then a Bibliography of one
      `- <title>. <url> (accessed <date>)` line per URL, titled by a blue cite of that URL where one exists
      (`bibliographyEntry`; main's `c408a944`);
    - a `fix_new` prescription, which reaches the report only when blue accepts it (`cli/blue/edit.go:88`).
  - Findings go to `docket.md`, verdict prose to `debate.md`/`run.md`, and retirements to `CHANGELOG.md`. No lens
    reads the matrix.
- **Carriers no list above names** (phase 7's census of S1–S5 against the implemented tree; each speaks the
  concept and is rewritten with it):
  - `cli/chair/dispatch.go:17`, the verb's doc: readiness as the head against each lens's pin. S5 does not search
    `cli/chair`, and S1's row reads only its material clause.
  - `record/record.go:498-506`, the `Migrating` doc: it lists the exemptions, and names neither the stale-area
    refusal nor the admission stamp.
  - `record/sitting.go:119`: "Dispatch enforces it — the lens's pin does not move".
  - `record/record.go:1201`, `cli/cli_test.go` (the PASS-over-an-open-gap tests), `cli/rolethroughgroup_test.go`:
    the PASS gate as refused over ANY open gap.
  - `record/passagreement_test.go:67`: "two below-material gaps" (b9's G3 is `always`, so material).
  - `docs/seat-command-triggers.md`, the `chair verdict` row: refused while a gap is open (III.5 names only the
    dispatch row beside it).
  - `releasegate/fuzz/fuzz_test.go`'s verdict oracle message, which counts every open gap.
  - The terms registry (`internal/terms/terms.json`, and `docs/vocabulary.md` generated from it): `gap`,
    `material`, `lens retirement`, `stale area`, and the collision of "retired" (a lens) with retiring a claim.
  - `debate.js`'s assemble prompt, which repeats `refs.go`'s claim that the report lists open gaps "below material,
    not certified against"; nothing produces that listing.
  - Comments and test wording in `record/mintbudget.go`, `record/roster_bind_test.go` and `verify/verify_test.go`.
- **Seat prompt ceilings.** `seatprobe/promptsize_test.go` holds each dispatched prompt under a ceiling and fails on
  growth. The chair's retirement paragraph (III.5) is its job, so the chair's ceiling is 10,200 (from 9,500; the
  prompt measures 10,059) and the paragraph drops the refusal the verdict help carries. The evidence lens carries
  the `last_sitting` clause (III.4) beside main's OCR page-check duty (`c408a944`): 12,166 against a ceiling of
  12,300 (11,846 before main's clause).
- **Archives:** 16 on `origin/main` `6d89b380` (`run-archive/*.tar.gz`). All 16 migrate with zero refusals (b6: 194
  events in, 194 out). Measured in `work/classmaterial.py` and `work/passopen.py`; see §V #11.

## III. Proposed Changes

### III.1 Gap vs finding; materiality by class — items 1–2 [R]

**Skill.** `skills/adversarial-audit/SKILL.md` gets a new bullet after `:51`:
> **A GAP CHANGES SOMETHING A READER DECIDES.** BEFORE minting, YOU MUST name, in the problem, the conclusion or
> reader decision the defect changes; one that changes none is a FINDING. Materiality starts from the class: the
> registry's `material_default` is `always` (claim, figure and citation classes), `never` (structure-noncompliance)
> or `by_grade` (medium and above). A `report-voice` gap is graded medium or above only when the process voice is
> PERVASIVE — a pattern across sections, not one sentence.

**Registry values** (adopted by gblock, D3/D7):
- **`always` (21):**
  - citation: citation-figure-misattribution, citation-status-drift, live-source-drift,
    within-source-condition-misattribution, cross-corpus-id-collision;
  - figure: figure-recount-fails, figure-miscomposition, metric-conflation, pricing-basis-drift,
    measurement-methodology-drift, causal-narrative-fails-reproduction;
  - claim: claim-contradicts-own-record, derivation-status-overclaim, cross-section-contradiction, false-universal,
    self-attestation, verification-scope-blindspot, unverified-composition, exhaustive-sweep-omits-case,
    enumeration-non-exhaustive, incomplete-repair-propagation.
- **`never`:** `structure-noncompliance`, PROMOTED from its `law/proposed/` file (definition, neighbour and
  distinguisher copied). The row carries `"_promoted": "gblock, 2026-09-11"`, because gblock is the human
  `_extending` requires. The proposal file is **deleted in the same commit**: lawqueue has no adopted form, the
  row names its source, and git keeps the text.
- **`by_grade` (18):** the rest, including report-voice.

**Every writer of, and instruction for, a shipped registry row** now carries `material_default`:
- **Writers:**
  - `feov-memory/class-registry.json` gains the field on all 40 rows.
  - `scripts/classgen` parses it and FAILS on a row with a missing or unknown value, so a hand-added row without
    it is a CI failure, not a runtime surprise.
  - `setup/setup.go:662` (`StageClassRegistry`) validates it (below); `seatprobe/build.go` stages
    `ShippedMaterialDefaults`; `migrate.go` copies and translates it; `StageForRun`/`StageForRunWithDefaults`
    (`replay.go:579`); and the hand-written JSON in `difftest/{scenarios,contract}_test.go` and
    `cli/referencechecks_test.go:136`.
- **Instructions:**
  - The registry's `_extending` text ("A slug that earns its keep across runs is promoted into this file by a
    human") gains "with its `material_default` — `always`, `never` or `by_grade`".
  - `capture/capture.go:2173`, the adoption text written into every `law/proposed/class-*.md` ("Adopting it means
    adding the slug to `feov-memory/class-registry.json` by hand"), gains "with its `material_default`; this run
    coined it as `<value>`". Capture reads the value from the run's `ClassNew.material_default`.
  - The 14 committed `law/proposed/class-*.md` files that remain after the promotion predate the field. Their
    sentence is edited in place to "with its `material_default` (`always`, `never` or `by_grade`)", because their
    records carry no coined value.
- `feov-memory/*.md` carries no row instruction (`grep -ln class-registry.json feov-memory/*.md` is empty).

**On the record** (facts-are-fields):
- `record.proto` `Mint` gains `ClassMaterial class_material` with `(sql).required` and a `why`; the requiredness is
  read by `recordpb/requiredfields.go` (the refusal) and `record/required.go` (the help).
  - The write path stamps it from the staged registry, or from the run's `ClassNew`. No seat can set it: outside
    `Migrating`, `Append` refuses a `Mint` that arrives with `class_material` already set, and the `mint` verb
    declares the field with `seat.Supplies(c, "class_material", "stamped from the class registry at the write
    path")`, so no flag exists and the vocabulary gate does not demand one.
  - Under `Migrating` the source is the record, not a seat. A `Mint` that arrives WITH `class_material` (a run
    written by this binary) keeps it as recorded, after checking it is a known value. One that arrives without
    it (a pre-change run) is stamped from the registry the migration translated (below).
- `ClassNew` gains `material_default`, and `class new` gains `--material-default` (enum, default `by_grade`).
- **One definition:** material = `always`, or `by_grade` with current severity mass ≥ 2.0; never for `never`.
  - It is carried twice: the `gap` view's `"material"` column (`views.go:~240`), and `Gap.Material` in the Go
    `Family` fold (`replay.go:147`), set on mint and on every regrade.
  - `const material` stays as the `by_grade` threshold, read only by those two carriers.

**Every decider and every statement of the PASS gate** switches to that definition (S2 and S3, §II):
- **Deciders:**
  - `dispatch.go:139` reads `material` (still OR-ing `stranded`); `openGaps` (`:503`, query at `:509`) selects it.
  - `refs.go:322-350` selects `g."material"` in place of the mass join.
  - `convergence_refusal.go:41-62`: "nothing open is material" and `FreshMaterialMints` both read the column;
    `MaxSeverityMass` stays as a reported figure and no longer decides.
  - `views.go:439,454,466`: `convergence_vs_verdict` reads the column.
  - `verify.go:371` reads `g.Material`.
  - **The chair's work list.** `WorkGapState` (`viewjson.go:733-753`) reads `Material` from the view's class-aware column in place of `6a67e0bf`'s severity mass, selected and scanned at
    `viewjson.go:778-797`. `WorkGapJSON` (`viewjson.go:635`) gains `"material"`, so the open-gap rows the chair
    lists at PASS say which gaps hold it. `sitting.go:156-166` becomes:
    - an open MATERIAL gap is BLOCKING: "gap X is open and material — PASS is refused while it is";
    - an open gap that is not material goes on the same list with `Blocks: false`: "gap X is open and not material
      (its class is never material | graded <severity>) — it does not hold PASS; your PASS lists it by class with
      why it changes no reader decision". `Complete` then agrees with the gate.
- **Prose, code comments and help** (S2's gate statements):
  - `cli/seat/help/verdict.md:9`: "an open gap" becomes "an open material gap (material by its class, else graded
    medium or above)".
  - `cli/seat/help/dispatch.md:13` (III.5) and `:15`: "A gap below material readies nobody" becomes "A gap that is
    not material — by its class, or graded below medium where its class goes by grade — readies nobody".
  - `record/available.go:84-88,118-122`: "the open-gap row in sitting.go already refuses this seat's PASS" becomes
    "…refuses PASS over a material gap"; the matching message at `available_test.go:321` follows.
  - **The docket affordance (`available.go:112-136`) is offered only on open MATERIAL gaps** (D12), the carried
    arm (`:128-134`) included. An open gap that is not material gets the sitting row above instead, which says
    it does not hold PASS. The verb itself is not refused: a party may still docket such a gap
    (`dispatch.go:140-143`, "a trifle may be escalated too"). The work list just does not offer an act that
    would create a PASS block the chair's by-class listing exists to avoid. Tested in §V #2. A stranded gap is held
    as material (`dispatch.go:133-164`), so the docket offer stays on it.

**Every Gate write-path refusal has a statement on the chair's work list, and the reverse** (round-5 gap 1).
`sitting.complete` agrees with the gate only if every refusal has a blocking item and no item blocks what the gate
admits. The census is the `*recordpb.Gate` case (`record.go:1172-1195`) and every refusal it reaches:
| Gate refusal | Refuses | Work-list statement (role `chair`) |
|---|---|---|
| `requireSupersededAreClosed` (`refs.go:272-305`), the stranded arm | ANY verdict, whatever the grade | **today (`6a67e0bf`) "gap X is open and superseded — PASS is refused until its minter closes it"; reworded.** `WorkGapState` gains `SupersededBy` beside `Stranded` (the view's `stranded`/`superseded_by`). A BLOCKING item: "gap X is open and superseded by Y — every verdict is refused while it is; close it with `--superseded-by`". A stranded gap is never listed as "not material — does not hold PASS", whatever its grade. |
| material gaps (`refs.go:322-350`) | PASS | `sitting.go:156-166`, class-aware (above). |
| unruled motions (`refs.go:370`) | PASS | `sitting.go:186` — agrees. |
| no line-of-inquiry review (`refs.go:378`) | PASS | `sitting.go:197` — agrees. |
| unraised contradictions (`unansweredContradictions`, `refs.go:387-389`) | PASS | **none today; added:** a BLOCKING item per claim, from the same function. |
| `requireNoCastLensReady` (III.5), head 0 included | PASS | **added:** "lens L is ready (active \| re-arm owed) — PASS is refused while it is", and "no report has been ingested — PASS is refused", both from the same fold. |
| `requirePassCoversStaleAreas` (III.5) | PASS | **added:** "area A is behind its pin P — PASS is refused until a spot-check this sitting names it". |
| `requireFailIsNotConvergent` (`record.go:1181-1185`) | FAIL only | none: the list claims nothing about a FAIL. The chair prompt states the refusal. |

Each added item is computed by the function the refusal calls, so the two cannot disagree. Pinned by
`TestChairWorkListStatesEveryGateRefusal` and the stranded row of `TestChairWorkListAgreesWithPassGate` (§V #2).
  - `recordsql/views.go:300-305` quotes the sitting sentence; it quotes the new one.
  - `dispatch.go:29` (`PassPermitted`'s comment) and `:144`: "material" is defined by reference to the class
    definition.
  - `agents/red-chair.md:13-14` and `debate.js:858-859` (III.5).
- **Non-carriers in S2** (another gate, or they read the plan's own flag): `sitting.go:186,197` and
  `refs.go:370,378,389` (motions, the inquiry read, unraised contradictions); `research-protocol/SKILL.md:21`
  (a contradiction finding); `cli/seat/help/inquiry-support.md:3`; `seatprobe/boards.go:503,527`,
  `recordpb/descriptions.go:164`, `cli/seat/help.go:22`, `cli/seat/verbs.go:230` (motions); `record/verdict.go:77-79`
  and `cli/chair/dispatch.go:102,105` ("every open material gap", `r.PassPermitted`, both read the plan);
  `seatprobe/production.go:61`, `releasegate/fuzz/{termination_test.go:106, fuzz_test.go:585,592,644,1742,2898}`
  (plan fixtures and the oracle over `plan.PassPermitted`); `debate.js:446,461,826,887,1033` (the relay fields), and `debate.js:802` (the termination comment, which reads the plan's `pass_permitted`).

**S1, every line classified** (70 lines, paths under `tools/internal/`):
| Lines | Disposition |
|---|---|
| `record/dispatch.go:139`, `refs.go:324`, `convergence_refusal.go:59,62`, `recordsql/views.go:466`, `verify/verify.go:371` | Deciders: read the class-aware column or `Gap.Material` (above). |
| `record/dispatch.go:29,43,46,160`, `refs.go:312-316,350`, `convergence_refusal.go:12,77-79`, `views.go:435`, `verify.go:367` | Rewrite (S4 table below). |
| `record/recordpb/record.pb.go:2794,2795,2797,6723,7258,7260` | Generated; regenerate. |
| `record/verdict.go:45,67,79`; `record/enums.go:243`; `record/recordpb/record.pb.go:7273`; `report/assemble.go:398,414`; `cli/chair/dispatch.go:18,105`; `record/dispatch.go:31,57` | Stay true: CEILING and readiness text, "every open material gap", where materiality is whatever the one definition says. |
| `record/dispatch.go:133,136,164,176` | Stay true: a stranded gap is held as material; the reason strings say "material" by the definition. |
| `record/params.go:22`; `record/convergence_refusal.go:8,54,66,68`; `cli/setup.go:94`; `verify/verify.go:376-377`; `refs.go:347` | Stay true: they say "material" and define nothing. (The audit's range `convergence_refusal.go:54-68` also covers `:59,62`, which are deciders.) |
| `view/view.go:556`; `report/assemble.go:469`; `record/record.go:108,117,119,754`; `recordpb/facets.go:163`; `recordsql/views.go:421`; `record.pb.go:6536` | Not materiality: ordering, the mass table and the mass formula. |
| `seatprobe/boards.go:622,638,644,649,657`; `record/enums.go:318`; `record.pb.go:7280`; `diagnostics/seen.go:81`; `cli/log.go:16`; `cli/seat/verbs.go:89`; `cli/lens/verify.go:157` | Another sense ("material" as content, subject prose). |

**S4, every line classified** (46 lines):
| Lines | Disposition |
|---|---|
| `dispatch.go:43,46,139,160` | Rewrite. `:43-45`: "`material` is the severity mass at or above which a `by_grade` gap is material; the class decides for `always` and `never`". `:160`: the reason names why ("never-material class" or the grade). |
| `refs.go:308-309` (with `:308-316`, one doc comment since `6a67e0bf` merged the two, and `:350`) | Rewrite as one comment stating the class-aware gate: an open gap that is not material stays open on the board. The rewrite drops the claim at `:314-316,350` that "the report lists it as open, below material, not certified against": nothing produces that listing. |
| `convergence_refusal.go:12` (with `:8-12`), `:62`, `:77-79` | Rewrite. The message's "(top severity mass %.1f, material is %.1f)" and "a finding graded medium or above" become "a material finding — by its class, else graded medium or above". The seat-facing clause at `:78`, "the sub-material gaps stay on the board and the report lists them as not certified against", becomes "the gaps that are not material stay open on the board". |
| `verify.go:366-367` | Rewrite the comment to the class-aware definition, without the same report-listing claim. |
| `views.go:426,434-435` (with `:462-463`) | Rewrite the comments; the SQL reads the column. |
| `record.proto:736-737` and `:742` (the `(sql).why` a lens sees when it mints without `--severity`) | Rewrite. New `why`: "materiality reads it wherever the class goes by grade (medium and above holds the gate), so an absent grade would score as TRIVIAL and the gap would read as harmless rather than ungraded". |
| `record.proto:295-297` (`(means)` on low_medium, medium, medium_high: "between minor and material", "material", "between material and serious"), seat-facing through `enums.go:125` and `enumvalue.go:225` | Rewrite: "below the by-grade material floor", "the by-grade material floor", "above the by-grade material floor, below serious". |
| `record.pb.go:2794-2795,6723,7258-7260`; `recordsql/testdata/schema.sql:136-138,1129` | Generated. Regenerate with `go -C $W/scripts run ./protogen` and `./schemagen`; §V #7's `schemagen` gate fails a stale copy. |
| `debate.js:849` | Rewrite: "a gap that is not material — by its class, or graded below medium — does not hold the gate". |
| `debate.js:869` | Rewrite "material begins at medium" to "materiality is the class's default: always, never, or by grade from medium". |
| `view/view_test.go:490` | Rewrite the comment: "materiality reads it wherever the class goes by grade". |
| `dispatch_test.go:385,423,453-454,468`; `cli/cli_test.go:1285`; `recordsql/convergence_test.go:47-49,63` | Stay true: these fixtures stage `by_grade` classes (`StageForRun`), where medium is the floor. |
| `cli/setup.go:94` | Stays true: "nothing material" names no definition. |
| `blue-synthesizer.md:110`, `lead-judge.md:121`, `adversarial-audit/SKILL.md:38,67`, `debate.js:870` | Another sense ("material" as content). |

`cli/lens/mint.go:250` (`--class` help) adds "its material default is recorded with the gap".

**Staging, loading, migration, pre-change runs:**
- **Setup refuses a registry without `material_default`, before any run state exists** (fork 2, gblock).
  Setup stages the registry from the CALLER's `<cwd>/feov-memory/class-registry.json` (`setup/run.go:296`), which
  can be stale. So the registry is parsed and validated beside setup's existing pre-run gate (`run.go:244-256`,
  `return 2`), before `record.NewRun` (`:266`). Today a registry that fails to stage only prints "NOT STAGED"
  (`:392-395`). A row with a missing or unknown value is refused, naming the file, the slug and the real remedy:
  "add `material_default` (`always` | `never` | `by_grade`) to that row, or re-stage from the plugin's shipped
  `feov-memory/class-registry.json`". It says nothing about `migrate`, because the run does not exist yet.
- `loadRegistry` still refuses a missing or unknown `material_default`; there is no tolerance path. Its text now
  says what can actually have happened. A pre-change run never reaches it: its database is refused at open first
  (below), with the `migrate` advice. A current run can only get here if its staged registry was hand-written or
  altered after setup. So the refusal names the file and slug, says the registry setup staged has been changed,
  and gives the remedy of restoring the row's value. It does not mention `migrate`.
- **A pre-change database is refused at open, before any view is read.** On the `hasEvents` path
  (`recordsql/store.go:365`), `recordsql.Open` checks that every column the binary declares for each body table
  exists, reusing `olderColumns`' two readers (`columnsOf`, `declaredColumns`: `read.go:554,571`). A missing
  column is refused naming the table, the column and `olderRunAdvice` (`read.go:513`, which names `migrate`).
  For a pre-change run that reads "this run's "mint" table has no "class_material" column — …migrate…". It is
  content, not an epoch: it names what the database lacks, and no version is compared. A fingerprint of the
  whole DDL was rejected because it is an epoch under another name. The cost is one `PRAGMA table_info` per
  declared table per process open; the handle is cached (`recordsql.Open`); §V #2 records the measured cost.
  Migration is unaffected, because its source never passes through `recordsql.Open` (§II).
- `StageForRun(run, slugs...)` writes `by_grade` for every slug, so all ten callers listed in §II keep working
  unchanged. `StageForRunWithDefaults(run, map)` serves the class-default tests.
- **Migrate's source is compiled in.** `classgen` (`scripts/classgen`, which today writes only the test package's
  `recordtest/classes_gen.go`) also writes `record/shippedclasses_gen.go`: `ShippedMaterialDefaults`, slug →
  `material_default`, from `feov-memory/class-registry.json`. `classgen -check` (run by `check -only classgen`) regenerates BOTH outputs in memory and fails if either committed file has drifted, and it fails on a row with a missing or unknown value. Migrate,
  the fuzz and the seat probe read it. Why not a flag: `feov-record migrate` takes only `--from`, `--to` and
  `--accept-loss` (`cli/migrate.go:51-54`, `flags/names.go:281-282`), and no path reaches it. A new `--memory-dir`
  would make one archive migrate differently on two machines. A compiled table makes the translation a property of
  the binary, testable on the archives and reviewable in the diff. Its cost is stated: a class adopted into the
  memory directory after the build takes `by_grade` until the next build, the same answer an unknown slug gets.
- The hand-written JSON bodies in `difftest/{scenarios,contract}_test.go` and `cli/referencechecks_test.go:136`
  gain the field.
- `feov-record migrate` gains exactly one translation, and **it supplies a value ONLY where the source lacks the
  field**: the class's `material_default` is the binary's shipped value for that slug (`record.ShippedMaterialDefaults`,
  below), else `by_grade`. A value the source carries is kept as recorded. Every field this plan adds that the
  write path stamps or migrate touches:
  | Field | Source has it (run by this binary) | Source lacks it (pre-change run) |
  |---|---|---|
  | staged `records/class-registry.json` row `material_default` | kept as recorded | rewritten after `copyChannels`: the shipped value; a slug the table lacks takes `by_grade` and is recorded as a STATED FILL |
  | `ClassNew.material_default` (declared `optional`, so absence is visible) | kept: a run-coined `class new --material-default always` stays `always` | a `class_new` registry entry runs identity, then the same rule: the shipped value if the table holds the slug, else `by_grade` as a STATED FILL |
  | `Mint.class_material` | kept, checked as a known value (above) | stamped by the write path from the translated registry |
  | `SpotCheck.areas` | kept by identity; validated against the cast, which replays first | empty; the stale-area gate is `!Migrating` |
  | `Gate.migration_admitted_gap_ids` | kept when it equals the open material gaps at the PASS; any other value refused | stamped by the write path on a PASS from the open material gaps at it (below) |

  **The stated fill** (fork 1, gblock). Migrate already has a stated-miss mechanism for fields a record predates:
  `migrate/entries.go:147-163` ("THE STATED MISS") fills them, only where absent, with a value that says it was
  never recorded. `material_default` is an enum and cannot say that in its own value. So the fill takes
  `by_grade`, and the manifest records it: `Manifest` (`migrate/manifest.go:23-47`, assembled at `:50-66`) gains
  `StatedFills []StatedFill{Where ("staged registry" | "class_new <event>"), Slug, Value, Why}` beside
  `Refusals`. Nothing is refused for a missing slug. So sleeper-service-plan's replayed `ClassNew` of
  `structure-noncompliance`, which has no `material_default`, takes `never` from the table with no fill, and the
  archives' other 27 class-creation slugs take `by_grade` as 27 stated fills (§V #11). Migrating a run written by this binary refuses nothing and changes none of these fields
  (§V #2).
- **The material PASS gate is exempt under `Migrating`, and a PASS the exemption admits says so on the record**
  (gblock, 2026-09-11; fork (a) BUILD, 2026-09-15). The `*recordpb.Gate` case in `record.go` wraps
  `requirePassClosesAllMaterialGaps` in `!Migrating`, like the convergence refusal and every-lens-sat. The code
  comment carries the reason: migrate translates history, and it does not re-judge an archived PASS under a rule
  that did not exist when the PASS was issued. The admission is a field, so `verify` tells it from a violation:
  - **The fact.** `Gate` carries `repeated string migration_admitted_gap_ids`: the open gaps that are material,
    by the class-aware definition, at the PASS a migration writes.
  - **The writer.** `stampMigrationAdmission` (`refs.go`) runs first in the `Gate` case. Under `Migrating`, on a
    PASS, it stamps the field from `openMaterialGaps`, the one query `requirePassClosesAllMaterialGaps` also
    reads, sorted, and empty when nothing open is material. A source PASS that already carries the field (a run
    this binary migrated) must carry exactly that set; any other value is refused, because the gaps'
    `class_material` is kept as recorded and the two differ only on an altered record. A non-PASS carrying the
    field is refused.
  - **The refusal.** Outside `Migrating`, a `Gate` carrying the field is refused, naming it, before any other gate
    reads the board: a live PASS cannot admit itself by claiming a migration. No flag sets the field.
  - **The reader.** `verify`'s `pass-closes-all-gaps` reads the last verdict's field. An open material gap it
    names is not a violation: the check holds, its detail says the PASS was admitted by migration, and
    `Check.admitted_by_migration` names the gaps. An open material gap it does not name FAILs the check.
  - **The surfaces.** `feov-record verify` (text: a `· <gap> — admitted by migration` line under the check;
    `--json`: `admitted_by_migration`) and run.md's Record verification section (`report/assemble.go`,
    `recordVerification`) print the admitted gaps. Nothing else states the outcome of a PASS over open gaps:
    `DeriveVerdict`'s VERIFIED reads only that a PASS exists, `convergence_vs_verdict` judges FAIL verdicts, and
    the chair's work list and `pass_permitted` are live-run gates that migration does not consult.
  - **Measured on the 16 archives.** b7's PASS carries G2 (`derivation-status-overclaim`, `low_medium`) and b9's
    carries G3 (`cross-section-contradiction`, `low_medium`), both classes `always` in the shipped table. The
    PASSes of is-91-prime-a, -b3, -b6 and -b8 carry none. b6's two open gaps (`inquiry-fate-mismatch`,
    `record-status-erasure`, both `low_medium`, both `ClassNew`) stay `by_grade`, so neither is material.
- `EventSchema` bumps (`requirements.json` → `schemagen`), because the event shape changes. Setup's existing
  start-of-run check (`setup.go:467`) covers new runs. No epoch is compared on resume or read (PR #794).

### III.2 Closure targets the object — item 2 [R]

`adversarial-audit/SKILL.md:56`, clause (1), is replaced by:
> (1) every "verified" clause carries seat + tool + target, and the TARGET IS THE OBJECT THE CLAIM IS ABOUT — the
> stored proof run FROM THE STORE, the cited bytes, the named path — never a stand-in. A check whose output would read
> the same had the thing never run (an exit status, a zero-failing summary, a list counted against itself) verifies
> nothing. A proof is reviewed as CODE AND OUTPUT.

Also:
- `debate.js:851` gains "run it from the proof store; a clean exit or '0 failing' is not an output".
- `cli/lens/close.go:72` (`--verified-against`) becomes "the object itself (the stored proof, the cited bytes, the
  named path), never a stand-in".

### III.3 Class fixes close on the enumerating command — item 3 [R]

- `adversarial-audit/SKILL.md:55`, after "declare the enumeration open —", gains: "and for a class, the acceptance
  check NAMES THE COMMAND THAT ENUMERATES THE CLASS; closure RUNS THAT COMMAND, never the instance list".
- `:61` appends: "For a class defect, the check is the enumerating command and its pass condition."
- `cli/lens/mint.go:258` (`--check` help) appends the same.

### III.4 The lens's sitting clause and the missed-then duty — item 4 [R]

**Where the fact comes from.** Every field `debate.js` reads from the relay (`chairEnv.plan`, `:884`) to build a
seat prompt:
| Field | Read at | Missing today |
|---|---|---|
| `parties[].seat_id` | `roleOfParty` `:907`, agent choice | schema-required |
| `parties[].gap_ids` | `lensPrompt` `:917`, `bluePrompt` `:927`, `benchPrompt` `:959`, `found_closed` filter `:930-931` | schema-required |
| `head` | `lensPrompt` `:847` | schema-required |
| `docket` | `bluePrompt` `:927` (the closing-arguments clause) | optional; `plan.docket \|\| []` silently reads as "nothing docketed" |
| `last_pin`, `stale_areas` (proposed in round 1) | would be `lensPrompt`, the chair | would be absence-encoded |

The rest (`pass_permitted`, `ceiling`, `epoch_limit_reached`, `max_epochs`, `why`) steer termination and the
outcome text (`:887-896, :988-996, :1022, :1033`), not a seat prompt.

**Decision: the lens's sitting fact leaves the relay and comes from the record.** Alternatives weighed:
1. Keep `last_pin` in the relay, schema-required, with a guard in `debate.js`. Rejected. A present but wrong value
   copied by the chair model passes every check, and `DispatchParityAudit` checks registrations, not fields.
2. Have `debate.js` read the record. Impossible: it reads no record and has no tool (§II).
3. Let the lens derive it from the changes projection. Rejected: that is arithmetic every sitting redoes, and it
   has no explicit never-sat value.
4. **Chosen:** the lens's work view, "the read a seat already does first" (`viewjson.go:587-588`), carries it.
   `SittingJSON` gains, for role `lens`, `last_sitting: {"kind": "first" | "behind" | "unchanged" | "undispatched",
   "pin": P, "head": H}`. `kind` is an explicit enum, never an absent field.

**`last_sitting` describes the latest sitting STRICTLY BEFORE the dispatch now being sat for.** It is NOT read
from `lensPins`. The lens reads its work view after `register`, and `sittingFor` counts that register as having
sat, so `lensPins` would report the current sitting: every lens would read `unchanged`, and G5 would fire on a
first sitting (§II). A new `lastSittingBefore(evs, seatID)` in `dispatch.go`, beside `owedSitting` and on the
same `dispatchLedger`/`sittingFor`:
- `D` is the latest dispatch row naming the seat: the one it is sitting for, or owes. With none, `kind` is
  `undispatched` (pin 0, head 0).
- A **prior sitting** is a dispatch `d` naming the seat with `d.at < D.at`, whose sitting register `r` (by
  `sittingFor`) also satisfies `r < D.at`. The second bound keeps an unsat earlier dispatch from borrowing the
  current register.
- The latest prior sitting gives `pin = d.pin`, and `head = D.pin` (the text this sitting audits). The kind is
  `first` with no prior sitting (pin 0), `behind` when `pin < head`, and `unchanged` when `pin == head`.

The sweep of every fact computed from `lensPins`, `sat`, `sittingFor` or `owedSitting` (§II) finds one other
reader inside the sitting it describes. That is `owedSitting` (`dispatch.go:443`, read at `sitting.go:113`), whose "not yet sat for" is
correct there. The facts this plan adds — retirement states, productive-sitting attribution and `stale_areas` —
are computed by `PlanDispatch`, which the chair runs between sittings.

**The prompt.** `lensPrompt` (`debate.js:842-847`) drops the relayed head clause and carries one static clause:
> YOUR SITTING IS ON YOUR WORK LIST: read `sitting.last_sitting` first. `first`: audit the report in full.
> `behind`: the report head moved past your last sitting (head H; you last sat at P) — audit the report as it now
> stands, in full. `unchanged`: the report is unchanged since your last sitting (head H). `undispatched`: the
> record holds no dispatch for you — log that and end the sitting. Whenever the kind is `behind` or `unchanged`,
> a FRESH gap on text you already read and passed at an earlier sitting states, in its mint reason, why you
> missed it then.

No "read changed text first" clause.

**Every required relayed field is refused loud when missing or mistyped.** The `PLAN` schema makes `docket`,
`why`, `max_epochs`, `epoch_limit_reached` and `stale_areas` required, alongside `head`, `parties`,
`pass_permitted` and `ceiling`. After `:885`, `debate.js` checks EVERY required field against its type:
- `head` and `max_epochs` are integers;
- `pass_permitted`, `ceiling` and `epoch_limit_reached` are booleans;
- `parties`, `docket`, `why` and `stale_areas` are arrays;
- every party has a string `seat_id` and an array `gap_ids`;
- every `stale_areas` entry has a string `seat_id` and an integer `pin`.

A failed check throws, naming the field and the sitting. `plan.docket || []` at `:927` becomes `plan.docket`.

**Go emits arrays, never `null`, at the source.** `PlanDispatch` (`dispatch.go:63`) initializes `Parties`,
`Docket`, `Why` and `StaleAreas` to empty slices. The fuzz's `parties` patch (`fuzz_test.go:594-596`) is deleted,
so the fuzz relays exactly what the verb prints, and §V #9 fails if any array comes back `null`.

**A relayed party field that disagrees with the dispatch row is caught** (fork (b), gblock: build the check).
`DispatchParityAudit` (`capture/dispatchparity.go:29`) takes the journal results (`capture.go:1818`) and the
journal's presence. It pairs each relayed plan with non-empty `parties`, in order, with the chair sitting's
dispatch group, in order; a count mismatch is a FAIL. It then compares every relayed field that has a
dispatch-row counterpart:
- the set of `seat_id`s against the group's rows;
- each party's `gap_ids` (as a set) against the `gap_ids` of the LAST row naming that seat in the group
  (`Dispatch.gap_ids`, `record.proto:1798`);
- `head` against that last row's `pin`.

A group can hold more than one row per seat: a plan with a docket is never "standing" (`DispatchStands`,
`dispatch.go:400-401`), and the archived B5 and B6 chairs wrote the plan twice. The last row is the one
`DispatchGroup.Sat` already reads. `DispatchGroup` exports it per party (`PartyRows map[string]PartyRow{Pin, GapIDs}`; `DispatchGroup` already has `Last int`, the group's last stream position),
built from the map `DispatchGroups` already keeps (`dispatch.go:308-314`), so capture reads no second copy of the
rule.

A mismatch FAILs, naming the sitting, the seat, and the relayed and recorded values. **Stated assumption:** the
journal holds ONE chair result per chair sitting, including across stop-and-resume (`debate.js:42-48`). If a
resume re-journals a cached chair result, the pairing is off by one and the audit FAILs on the count mismatch. That
is loud, not silent. `TestRelayPairingOnDuplicatedChairResult` (§V #15) pins that outcome. The first resumed live
run's capture (`run-record-audit.md`, the `dispatch-parity` line; §V #10) is what exposes whether resumes do this. With no journal the
registration half still runs, and the detail states that party fields were NOT compared — never a silent pass.
Carriers:
- `dispatchparity.go:12-28` (the doc comment);
- the call at `capture.go:1853`;
- `debate.js:440` (comment) and `debate.js:858` / `red-chair.md:13` ("a party you drop or add is a finding
  against you" becomes "a party, gap id or head you drop, add or alter is a finding against you");
- `docs/seat-command-triggers.md:72`.

### III.5 Per-seat retirement — item 5 [R]

**Definitions** (`record/dispatch.go`):
- **Sittings** are the lens's registers following a dispatch naming it (`lensPins`' `sat`), including sittings
  engaged on its own gaps.
- A mint belongs to the sitting whose register most recently precedes it.
- A sitting is **productive** if it minted a fresh mint that is material now.
- **States**, folded in order (gblock D1: re-arm ONCE per retirement):
  - **active**: fewer than 2 trailing barren sittings. A productive sitting returns the lens here, with its re-arm
    unspent.
  - **retired**: 2 barren sittings, re-arm unspent.
  - **retired for good**: its re-arm sitting was barren.

**Readiness** replaces Source 1, and **keeps today's head precondition**: no lens is ready while `Head == 0`,
because nothing has been ingested and there is nothing to audit. A never-sat lens at `Head > 0` is active.
- active → ready.
- retired with `Head > pin` → ready once (the re-arm).
- retired at the head, or retired for good → not ready.
- Sources 2 and 3 are unchanged. A lens retired for good still answers its own open gaps.
- The `why` strings name the state, so the valve (`debate.js:821`) reads a change of state as progress.
- Cost bound: at most 2 barren sittings per active spell, plus 1 per retirement.

**Finish:**
- **One retirement fold feeds both halves** (gblock, round 4: MIRROR `PassPermitted`). `lensStates(evs, ids, head)`
  in `dispatch.go` folds each cast lens to its state and whether it is READY: active, or retired with its re-arm
  owed (`Head > pin`). `PlanDispatch` readies from it, and the Gate write path reads the same fold.
- `PassPermitted` requires `Head > 0`, no material gap open, no unruled docket, no parties, and NO CAST LENS READY:
  each is retired at the head (re-arm not owed) or retired for good. It is no longer "at the head".
- `requireEveryCastLensSatAgainstHead` becomes `requireNoCastLensReady`, at the same site
  (`record.go:1190-1194`, `!Migrating`). Outside `Migrating` it refuses a PASS whenever the fold leaves any cast
  lens ready, naming each lens and why it is ready. So a chair that ignores `pass_permitted` cannot record a PASS
  over an owed re-arm, which today's gate (`pins[s] < head`, `dispatch.go:530-565`) also refuses.
  It keeps the head-0 refusal (`dispatch.go:562-563`).

**Stale-area spot-check** (N3, gblock):
- The plan gains `stale_areas`: every lens retired for good whose pin is behind the head, with that pin. The chair
  reads it from its own `dispatch next` output; the relay carries it (schema-required), and `debate.js` builds
  no prompt from it.
- At the PASS sitting the chair reads the record's changes since each pin against the area's duties, and names the
  areas: `chair spot-check --areas <csv>`, a field on the SpotCheck event, validated against the cast.
- `requirePassCoversStaleAreas` refuses a PASS until a spot-check in that sitting names every stale area. It is
  gated `!Migrating`, because is-91-prime-a, -b3 and -b6 carry PASS gates.
- A defect found there goes in the spot-check's prose and the chair records no verdict (§IV, residue).

**Interactions:** the chair's verdict, CEILING, MaxEpochs, the valve, the bench and the mint budget are all
unchanged. A lens at budget retires within two sittings. There is no chair route to ready a lens (D6).

**Carriers that no gate reads** (goldens diff only the flag list, so each is named). S5 gives 53 lines. Four are new since #927 and are not carriers: `dispatch.go:391` (the chair's owed-register
message names the dispatch row's pin) and `:417,420,424` (`DispatchStands`' key is a row's seat, pin and gaps).
The rest change as follows:
- **Help pages:**
  - `cli/seat/help/dispatch.md:11` becomes: "Three things make a party ready. A lens that is active (it minted fresh
    material within its last two sittings), or retired and re-armed once because the head moved since it sat —
    never before a report is ingested. The minter of an open material gap, and blue, below the gap's limits. The
    bench, at impasse."
  - `dispatch.md:13`: "every cast lens has sat against the head" becomes "no cast lens is ready — each is retired
    with no re-arm owed, or retired for good — and `stale_areas` names what the chair spot-checks before a PASS".
  - `cli/seat/help/spot-check.md` gains the `--areas` paragraph.
- **Code and schema comments:**
  - `dispatch.go:14,29,56,87,116-118,459-473,530-565` are rewritten to the states.
  - `record.proto:1787-1791` (the Dispatch comment; regenerate `record.pb.go:5933`): "a lens engaged with no gaps is
    ready by its retirement state".
  - `releasegate/fuzz/fuzz_test.go:1898`: "or FRESH, because the head moved past its pin" becomes "or by its
    retirement state".
  - `releasegate/fuzz/scriptedchair_test.go:32`: the fixture reason "every lens sat against the head and nothing
    material is open" becomes "no lens is ready and nothing material is open".
  - `releasegate/fuzz/termination_test.go:26`: "the evidence lens (head moved)" becomes "the evidence lens (active)".
- **Tests outside S5's directories that carry the old readiness model.** Each is rewritten to the states: the
  assertion as well as the wording, where the behaviour changes.
  - `record/dispatch_test.go:165` ("still ready under rule 1 when the head moved") and `:180` ("while a cast lens has
    not sat against the head"): ready because active; PASS refused while a lens is not retired.
  - `dispatch_test.go:251,255` (`TestAnEmptyDispatchIsTermination`): "every cast lens sat against the head"
    becomes "every cast lens retired" (the fixture gains the barren sittings).
  - `dispatch_test.go:345,356,366` (`TestPassIsRefusedUntilEveryCastLensSatAgainstTheHead`): renamed
    `TestPassIsRefusedWhileAnyCastLensIsReady`, with the messages to match.
  - `record/owedsitting_test.go:40`: "its pin never moved and dispatch readied it again" becomes "it never sat, so
    its state never moved and dispatch readied it again".
  - `tests/simulator/prompts.test.mjs:87`: "three lenses whose pin the head moved past" becomes "three active
    lenses".
  - `tests/simulator/debate.test.mjs:66` (test title "is told the head moved") becomes "is told to read its last
    sitting from its work list".
  - `debate.test.mjs:152` (`b5Plan`'s why "head 144 is past its pin 60") takes the state wording `PlanDispatch`
    now writes.
- **Prompts and constitutions:**
  - `debate.js:456` (the `gap_ids` description), `debate.js:858` and `red-chair.md:13`: "the lenses whose pin the
    report head moved past" becomes "active lenses, and retired lenses a head move re-arms once".
  - `red-chair.md:14` and `debate.js:859` gain:
    > A LENS RETIRES WHEN IT STOPS FINDING MATERIAL. A lens seat retires after two sittings with no fresh material
    > mint, is re-armed ONCE when the head moves, and retires for good if that sitting is barren. The plan permits a
    > PASS only when no lens is ready — every lens retired with no re-arm owed, or retired for good — and nothing
    > material is open. BEFORE a PASS, read the changes since each stale area's pin and name those areas in your
    > spot-check; a defect you find there goes in that spot-check, and you record no verdict. YOUR PASS LISTS EVERY
    > OPEN GAP THAT IS NOT MATERIAL, BY CLASS — your work list marks each — with one line on why it changes no reader
    > decision, on the record.

    The paragraph states no refusal: the stopping-judgment clause beside it says the tool refuses a PASS the board
    does not permit, and the verdict help carries each refusal.
  - `red-chair.md:16` (the spot-check duty) and its prompt twin `debate.js:861` gain the stale areas.
  - `dispatch.go:256`, `sittingFor`'s comment listing its readers, gains `lensStates` and `lastSittingBefore`.
  - `dispatch.go:98` (`lensPins` in `PlanDispatch`) is replaced by the retirement fold's call.
- **Printers and schema:**
  - `cli/chair/dispatch.go:113` ("— the head moved past its pin") becomes "— audits the report (its state is in the
    reasons below)", and the printer lists `stale_areas`.
  - The `CHAIR_ENVELOPE` plan schema (`debate.js:444-466`) gains `stale_areas`, required (III.4).
  - `docs/seat-command-triggers.md:72` gains one clause.
- **Non-carriers** (other senses of the words):
  - `debate.js:215` ("pins" means tests pin);
  - `docs/{propagation-and-anchoring,seat-surface-naming}.md`;
  - `lead-judge.md:118`, `report_template.md:91`, `research-protocol/SKILL.md:61`;
  - `record.proto:883`, `docs/seat-command-triggers.md:66,169-207`;
  - `vocabulary.md:117,181,190`, which stay true;
  - `debate.js:795` (the section banner) and `debate.js:855` (a comment on what the chair does), `red-chair.md:3`
    (the description: "spot-checks the archive");
  - `dispatch.go:222,239` and `record.proto:1789`: "pin" is the dispatch row's head, which stays true.

### III.6 Blue: the reader's question, focus, complexity that pays — item 6 and gblock's standard [R/P]

`agents/blue-researcher.md` gets new points after `:57`; `blue-synthesizer.md` gets the same after `:54`, phrased
for authorship.
> 4. **Material to the question asked — of interest, not merely interesting.** YOU MUST make every section, table and
>    caveat name the reader's question it answers. A tangent stays only if it illuminates something central to it.
>    Allegory and metaphor are welcome where they make complex material easier for a lay reader.
> 5. **DEFEND FOCUS.** When a gap asks for immaterial detail or complexity without value, YOU MUST answer with that
>    argument — rebut, or argue `defect_accepted` with the reason it changes no decision — rather than comply. The
>    bench arbitrates.
> 6. **COMPLEXITY MUST PAY.** Kept complexity — explanation, method or implementation — MUST beat the naive version by
>    more than the cognitive load and cost it adds; state that trade where you keep it.

**The drop rule** (D4), at `blue-researcher.md:125-126` and `blue-synthesizer.md:59-60`, becomes:
> YOU MUST NOT drop substantive content — a claim, its evidence, or its qualification. Material that answers no
> question the reader asked may leave the report for the record: retire it there, with the reason.

**Other carriers:**
- The Pragmatist bullets point to point 6.
- `debate.js:870` gains "…or, where the gap changes no reader decision or asks for complexity that does not pay,
  argue `defect_accepted` with that reason".
- `debate.js:779` gains "every section names the reader's question it answers".
- The red mirror, at `adversarial-audit/SKILL.md:51`: "YOU MUST NOT file a gap whose fix adds complexity that fails
  blue's test".

**Unchanged:**
- `README.md:7` and the blue `description`s ("union, never summary" is still true of substance).
- `catechism_template.md:21` (Q5 asks about the topic's worth, not the text).
- `report_template.md:56`.

### III.7 The half-landed telos — item 7 [R]

- `agents/blue-synthesizer.md:46-54` ("**Red's PASS is your win condition**…") is replaced by `blue-researcher.md`'s
  text ("AND THE PASS IS EVIDENCE, NOT THE GOAL…").
- `agents/blue-researcher.md:109-111` loses its history ("An earlier version of this constitution said 'Red's PASS
  is your win condition'. That frame is wrong and it invites…") and reads in the present tense:
  > **AND THE PASS IS EVIDENCE, NOT THE GOAL.** Treating the PASS as the target makes the target *satisfying the
  > auditor* rather than *being right*, and every dodge below is a rational move once the goal is a verdict.
- Both copies are the same present-tense text. `grep -rn "win condition" plugins/` then returns nothing.

### III.8 Default cast = all seven areas — item 8 [R]

- `record/cast.go:13-15`: `DefaultCastAreas` holds all seven areas.
- `debate.js:560-563`: `DEFAULT_AREAS = RED_AREAS.slice()`.
- Help and docs: `cli/setup.go:93` help, `commands/research.md:6` ("narrow it only with a reason"), `setup/run.go:43`.
- `adversarial-audit/SKILL.md:62`: "the voice seat is not always cast" becomes "the voice seat sits unless an
  operator narrows the cast".
- Fixtures:
  - `releasegate/fuzz/fuzz_test.go:740-745,2443`: `fuzzLensSeats` comes from `record.CastFor(nil, …)`;
  - `record/cast_test.go:31`;
  - `tests/simulator/debate.test.mjs`.
- `seatprobe/build.go:54` stays at four (D5); only its comment is corrected.
- Cost: the first sitting is 7 lens sittings instead of 4 (4 waves instead of 2 at the cap). After that the III.5
  bound applies, where today each lens sits on every head move.

### III.9 Red's report-bound text obeys report voice — gblock [R]

Red text that reaches `report.md` is addressed to the reader, about the subject. It reaches the report three ways,
and each is held:
- **The risk matrix** prints the first sentence of an open gap's `problem` and `required_fix`, and nothing else the
  mint carries: not `mint_reason`, the acceptance check, the class, the gap id or the seat. Held by the refusal and
  advice in (c); #13 pins the row.
- **The source's note and the Bibliography** print a labelled corroboration's `--title` beside its URL and access date. The label is
  minted only for `supports`, `supports_with_bridge` and `weak` (`backsTheClaim`, `cli/lens/verify.go`). The anchor
  becomes a tool marker `[^N]`, and the claim is blue's sentence. Red's reason, outcome, confidence and seat stay on
  the record. Red's title prints in the note when red's anchor is the first under it, and in the Bibliography line
  when no blue cite names the URL. A source's own title is subject matter, so the title is shown. **It is held as the mint
  fields are** (gblock, 2026-09-15): the `*recordpb.Verify` validation in `record.go` refuses the unambiguous tells
  in the title of a LABELLED corroboration (`refuseCorroborationTitleVoice`, `record/redvoice.go`; the label is the
  field that makes it a footnote), outside `Migrating`; `corroborate` returns the ambiguous rest as `voice_tells`.
  An unlabelled corroboration — refutes, absent, unreachable — prints nowhere, so its title is not checked.
  `TestCorroborationCarriesOnlyItsSourceIntoTheReport` pins the note and the Bibliography line to exactly title, URL and date. Measured with
  the implemented checks over the 16 archives: **11 labelled corroboration titles, 0 refused, 0 advised** (5
  unlabelled corroborations, not checked).
- **A `fix_new` prescription** becomes report prose when blue accepts it. **It is held at the write where red
  authors it** (gblock, 2026-09-15): `fix_new` is set only by `lens mint --new` (`cli/lens/mint.go`), so
  `refuseMintReportVoice` reads it beside `problem` and `required_fix` — the unambiguous tells refused outside
  `Migrating`, the ambiguous rest returned in the mint's `voice_tells`. Blue's edit advisory on the accepted text
  (`cli/blue/edit.go:102`) stays as a second net, and red's voice lens audits it like any edit. Measured with the
  implemented checks: **10 prescriptions across the 16 archives, 0 refused, 3 advised** (record-store-authority G4
  and G7, "this report"; research-loop-counterparts G1, "this run").
- **(a) Skill.** `adversarial-audit/SKILL.md`, after `:62`:
  > **WHAT YOU WRITE THAT REACHES THE REPORT IS HELD TO ITS VOICE.** The first sentence of a gap's problem and of its
  > fix ship in the report's risk matrix while the gap is open; replacement text you prescribe becomes the report's own
  > text when blue accepts it; and a corroborating source's title ships in the source's note and Bibliography entry. Write all of them
  > to a reader of the SUBJECT: no lens, seat or side names, no gap or finding ids, no epochs or sittings, no tool
  > verbs, no account of the run. Say "the analysis claims…", not "this run claimed…". The tool refuses the
  > unambiguous tells — a seat or lens id, a finding label, a gap id beside a process word, a lane tag — and flags the
  > rest. The run's part goes in the mint reason, and how you found a source goes in the corroboration's reason.

  Each red lens constitution carries the same duty under "What a lens may not do" ("WHAT YOU WRITE THAT PRINTS IS
  WRITTEN TO THE READER"), naming the problem, the fix, the prescribed replacement and the corroborated title.
- **(b) The chair's PASS listing stays on the record** (D8). The risk row keeps red's reader-facing remedy.
- **(c) Refusal of the unambiguous, advice on the rest** (gblock's fork ruling). `tells.go` gains a `Refuse bool` on
  each tell.
  - **Refused:**
    - seat and lens ids: `\b(red-lens-[a-z-]+|red-chair|blue-(respond|synthesize|lane-\d+)|judge-terminal)\b`;
    - finding labels: `\b[a-z-]+-F\d+\b`;
    - gap or finding ids joined to process words: `\b(?i:gaps?|findings?) G\d+\b` (the id carries its `G`: "a gap
      2 metres wide" is subject prose, and no archived problem or fix differs between the two spellings),
      `\bG\d+['’]s (fix|repair|closure|mint)\b`, `\b(minted|closed|regraded|superseded) (as )?G\d+\b`;
    - lane tags: `\[(minority|lane-\d)[^\]]*\]`.
  - **Advised:**
    - "this run/round/report/sitting", "the debate", "research lanes";
    - `\b(red|blue) (team|side|lens|chair|seat)\b` and `\b(epoch|sitting) #?\d+\b`;
    - draft history and apparatus.
  - The check runs on the write path (the `*recordpb.Mint` validation in `record.go`, beside `validateClass`). It
    covers **`problem`, `required_fix` and `fix_new`, never `mint_reason`**; the same refusal runs on a labelled
    corroboration's `title` in the `*recordpb.Verify` validation, never on its reason. A refusal names the flag and
    the term; advice returns in the verb result's `voice_tells`.
  - Under `Migrating` neither check runs.
  - Bare ids stay a written duty under (a). Blue's advisory is unchanged, and it gains the new advisory tells.
  - `cli/seat/help/mint.md` adds the refusal beside its write-time refusals (`:11-13`), naming the replacement text;
    `cli/seat/help/corroborate.md` says the title is printed, refused on the unambiguous tells and advised on the rest.
- **Measured** with the implemented `reportvoice.Refused` and `reportvoice.Advised` over the 16 archives' 123 gaps
  (phase 6: a throwaway Go test over the migrated `record.db` files, not committed). The refused set hits **4 of 123**
  problem statements and **1 of 123** fixes, on 4 gaps:
  - quadratic G8: `blue-lane-2` and `[minority: lane-2/primary-literature]`;
  - quadratic G25: `blue-respond`;
  - b8 G2: `blue-synthesize`, in both the problem and the fix;
  - b9 G4: `red-lens-evidence`.
  The advised set hits **28 of 123** problem statements and **22 of 123** fixes, on 41 gaps. Most are "this run" (25),
  "this round" (11) and "this report" (9).

### III.10 Decisions (2026-09-11)

| # | Question | Ruling | By |
|---|---|---|---|
| D1 | Re-arm on every head move changed nothing versus today | Re-arm once per retirement; a barren re-arm retires the lens for good; PASS needs every lens retired, plus the stale-area spot-check | gblock |
| D2 | Round cap | MaxEpochs and the valve stay; the plan-audit "no round cap" ruling does not apply to FEOV | lead, disclosed |
| D3 / D7 | Class list; "pervasive" | Adopt the 21 `always`; promote `structure-noncompliance` as `never`; report-voice is `by_grade` plus the lens duty | gblock |
| D4 | "process record" is a gated term | "the record — retire it there" | lead |
| D5 | Seat-probe staging lenses | Stay at four (a budget fixture) | lead |
| D6 | A chair route to ready a lens | Reversed by the lead; residue accepted by gblock (§IV) | lead; gblock |
| D8 | The "changes no decision" line in the report? | Record only | gblock |
| D9 / fork | Refuse or advise at mint | Refuse unambiguous tells; advise on ambiguous ones | gblock |
| N3 | Changed text behind a lens retired for good | `spot-check --areas`, `stale_areas`, PASS refusal | gblock |
| N4 | Refusal reach | `problem` and `required_fix`; bare ids are a written duty. Extended by the two phase-7 rulings below | gblock |
| phase 7 (title) | A red corroboration's title reaches the Bibliography | `corroborate` REFUSES the unambiguous tells in a labelled corroboration's title — the `reportvoice.Refused` set `mint` uses — and advises on the ambiguous rest, skipped under `Migrating`. At ruling: none of 11 labelled titles in the 16 archives carries a tell (III.9) | gblock, 2026-09-15 |
| phase 7 (`fix_new`) | A lens's `fix_new` becomes report text when blue accepts it, and the mint check never read it | Held to the same check at the write where red authors it: refuse unambiguous tells, advise ambiguous, skipped under `Migrating`; blue's edit warning on accepted text stays as a second net. At ruling: 10 archived prescriptions, 3 with advised tells only (III.9) | gblock, 2026-09-15 |
| fork | Regrade fold | Residue accepted (§IV) | gblock |
| fork | Pre-change runs | No tolerance: `migrate` writes the field; a pre-change database is refused at open by the columns it lacks, and a pre-change registry by content, both naming `migrate`. No epoch comparison on read (PR #794). | gblock; open-time placement by the lead |
| fork | The material PASS gate under `Migrating` | Exempt, like the convergence refusal and every-lens-sat: migrate does not re-judge an archived PASS under a rule it predates | gblock |
| D10 | Where a lens's last-sitting fact comes from | The record, on the lens's work view, with an explicit `kind`; not the chair's relay (III.4) | lead |
| D11 | A lens ready before any ingest | No: the `Head > 0` precondition is kept | lead |
| D12 | The chair's docket affordance on an open gap that is not material | Not offered; the verb is not refused (III.1) | lead |
| D13 | What `last_sitting` describes | The latest sitting strictly before the dispatch now being sat for (III.4) | lead |
| fork (a) | `verify`'s `pass-closes-all-gaps` on a migrated archived PASS over a gap class defaults now make material | Decided BUILD (reopened): 2 of the 16 archives on the rebased tree have such a PASS, b7 over G2 (`derivation-status-overclaim`, `low_medium`) and b9 over G3 (`cross-section-contradiction`, `low_medium`), and `verify` FAILed both. The migrated PASS carries `Gate.migration_admitted_gap_ids`, stamped only under `Migrating` and refused on a live write; `verify` reports it as admitted by migration and not as a violation (III.1) | gblock, 2026-09-15 |
| round 4 | The PASS write path's lens condition | MIRROR `PassPermitted`: one retirement fold feeds the dispatcher and the Gate write path, which refuses a PASS outside `Migrating` while the fold leaves any cast lens ready (III.5) | gblock |
| fork 1 (round 5) | A staged or class-creation slug absent from the compiled-in table | `by_grade`, recorded in the manifest's `stated_fills`; never a refusal (III.1) | gblock |
| fork 2 (round 5) | A caller's registry without `material_default` | Setup refuses it before any run state, naming the registry remedy; `migrate` advice only where the run predates the change (III.1) | gblock |
| D15 | Gate refusals vs the chair's work list | Every refusal has a blocking item from the same function; the stranded arm and unraised contradictions gain one (III.1) | lead, from round-5 gap 1 |
| D14 | Where migrate reads class defaults | The compiled-in `ShippedMaterialDefaults` (classgen), not a new flag (III.1) | lead |
| fork (b) | A wrong id in a relayed `gap_ids` | Build the check: `DispatchParityAudit` compares every relayed party field that has a dispatch-row counterpart (III.4) | gblock |

## IV. Risk & Mitigation

| Risk | Source | Kept by |
|---|---|---|
| Late real catches are lost | sleeper G16, G17, G19 (e3/e4, fresh, high or medium_high); quadratic G18 (e3) and its successor G26 (e5) [R] | Class-pinned materiality: G18 and G26 are `always`; the rest grade ≥ medium_high. The minting lenses were productive every sitting, so none retired. Successor mints ride Source 2. |
| Dead checks | #910: every fuzz arm dead, passing because both sides failed alike [P] | III.2: output that reads the same had the thing never run verifies nothing. |
| A late catch where earlier text claimed impossibility | catalogue-shape R6.1, the foreign-file write [P] | The single re-arm, then the chair's stale-area spot-check. `false-universal` is `always`. |
| A late change breaks the area of a lens retired for good (ACCEPTED, D1) | D1 | The chair must read and name each stale area before a PASS; a chair read is not a lens sitting. |
| A regrade down retroactively retires a lens for good without its re-arm (ACCEPTED, fork) | material is read NOW | The stale-area spot-check covers that area before a PASS. |
| The chair finds a defect in a stale area (ACCEPTED, D6 reversed) | N1/N2 | The chair records it in the spot-check and records no verdict, so the run ends UNVERIFIED: loud. A FAIL is refused on a converged board (`requireFailIsNotConvergent`), so the bench's `outcome --reason` cites the spot-check; the engine's "neither PASS nor CEILING held" (`debate.js:996`) understates it. Rare: the defect must fall in an area whose lens already spent its re-arm. |
| A reader of "material" disagrees with the rest | the deciders and gate statements (§II, III.1) | One definition carried on the record: the view column and the `Family` fold. #2 tests the `verify` PASS case and the chair's work list against the gate; #14 re-runs S1–S5. |
| The chair's work list says a PASS is blocked while the gate admits it, or the reverse | `sitting.go:156-166` (audit round 2) | `WorkGapState.Material`; #2's table test holds `sitting.complete` equal to the gate's answer for `always`, `never` and `by_grade` gaps. |
| A migrated PASS over a gap class defaults now make material reads as a `pass-closes-all-gaps` violation, or a live PASS claims a migration's admission | the `Migrating` exemption (gblock); fork (a) | The migrated PASS carries `Gate.migration_admitted_gap_ids`, which `verify` reports as admitted by migration; a live `Gate` carrying the field is refused (III.1). #2's `TestMigrationRecordsAdmittedPassOnArchives` (b7, b9) and `TestLiveGateCarryingMigrationAdmissionRefused`. |
| A pre-change run dies on SQLite "no such column" | `recordsql/store.go:339-372` | The open-time column check fires first and names `migrate` (#2). |
| A relayed prompt input is dropped, mistyped or copied wrong | the chair model's copy (`debate.js:884`) | The sitting fact is off the relay (III.4). Every required field is type-checked and throws. Go emits arrays, never `null`. `DispatchParityAudit` FAILs a relayed seat set, `gap_ids` or `head` that disagrees with the dispatch rows (fork (b)). |
| A chair records a PASS over an owed re-arm (a regression on today's gate) | round-4 audit | `requireNoCastLensReady` reads the same fold as `pass_permitted` (III.5); #1's `TestPassRefusedWhileReArmOwed` and #4's delete-the-row. |
| The lens reads its current sitting as its last one | `sittingFor` counts the register the lens has just made (§II) | `lastSittingBefore` bounds both the dispatch and its register strictly before the current dispatch; #1's fixture registers before reading, and #4 drops the bound. |
| Migrating a run written by this binary refuses or rewrites its own fields | identity translation by name (`migrate/registry.go:27-44`) | The preset-`class_material` refusal is gated `!Migrating`; the translation supplies values only where absent (III.1 table); #2 round-trips such a run. |
| Friction from the red report-voice refusal | the implemented refusals over the 16 archives (III.9) | 4/123 problem statements and 1/123 fixes refused (4 gaps), 0/10 `fix_new` prescriptions and 0/11 labelled corroboration titles; the rest are advised (28/123, 22/123, 3/10, 0/11). Lenses rephrase run and round references in reader terms. `mint_reason` and a corroboration's reason are never checked. |
| structure-noncompliance becomes invisible | sleeper G7/G18 (medium, bench-repaired) [R] | It stays on the board, listed by class at PASS on the record. Migrated, sleeper's replayed `ClassNew` takes `never` (III.1). |
| Blue uses "focus" to shed evidence | the item-6 narrowing | A claim, its evidence and its qualification cannot be dropped; retirement is on the record; the bench arbitrates; `always` classes cannot be argued sub-material. |
| An unmigrated run, or a hand-built fixture, breaks | no tolerance path | The open-time refusal names `migrate`; setup refuses a stale caller registry, naming the registry fix; `StageForRun` writes `by_grade`; #11 migrates all 16 archives with zero refusals. |
| Seven lenses cost too much | #788, #554 | The first sitting costs 7; after that the III.5 bound; MaxEpochs over all. |

## V. Verification Plan

`T=plugins/frank-exchange-of-views/tools`, `S=scripts`, `W=` the worktree. Absolute `-C` paths; no `cd`.
Scratch goes under `~/.claude/scratch/feov-lens-review/`. **Step 0:** `git -C $W fetch origin && git -C $W rebase
origin/main` (`1ef6338f` or later), with a clean tree; implementation starts there. Any `file:line` a later
`origin/main` moves is re-verified before it is edited.

| # | Command | Proves | Re-armed by |
|---|---|---|---|
| 1 | `go -C $W/$T test -count=1 ./internal/record/...` running `TestLensRetiresAfterTwoBarrenSittings`, `TestRetiredLensReArmedOnceByHeadMove`, `TestBarrenReArmRetiresForGood`, `TestRetiredForGoodIgnoresLaterHeadMoves`, `TestFreshMintOnReArmMakesLensActiveAgain`, `TestActiveLensReadyAtSameHead`, `TestNoLensReadyBeforeIngest`, `TestLineageMintIsNotFresh`, `TestNeverClassMintIsNotFresh`, `TestPassPermittedWithLensRetiredForGoodBehindHead`, `TestPassRefusedWhileReArmOwed` (a lens retired with `Head > pin`: `pass_permitted` is false AND a chair's `verdict --as PASS` is refused at the write path, naming the lens), `TestPassPermittedAndGateShareTheFold` (table over active, retired-at-head, retired-with-re-arm-owed and retired-for-good: the plan's `pass_permitted` and the write path's answer agree in every row), `TestBudgetExhaustedLensRetires`, `TestWhyCarriesState`, `TestLensWorkCarriesLastSitting` (the fixture REGISTERS the lens for the dispatch before reading its work view, as a live seat does: its first sitting reads `first`; after a mint and a blue edit, the next sitting reads `behind` with the earlier pin; a re-dispatch at the same head reads `unchanged`; a lens whose earlier dispatch went unsat reads `first`; a lens never dispatched reads `undispatched`; the key is always present) | retirement (D1), D11, D10 | `dispatch.go`, `refs.go`, `sitting.go` |
| 1b | same package: `TestPlanListsStaleAreas`, `TestPassRefusedUntilSpotCheckCoversStaleAreas`, `TestSpotCheckAreasValidatedAgainstCast`, `TestMigratingSkipsStaleAreaGate` | N3 | `dispatch.go`, `refs.go`, `cli/chair/spot_check.go` |
| 2 | `go -C $W/$T test -count=1 ./internal/record/... ./internal/verify/... ./internal/cli/... ./internal/setup/...` running `TestMaterialAlwaysHoldsLowGrade`, `TestMaterialNeverReleasesHighGrade`, `TestFamilyGapMaterialMatchesView`, `TestChairWorkListOverOpenNeverClassGap` (an open `never` gap graded high: the item is present with `blocks: false`, `sitting.complete` is true, and the PASS appends), `TestChairWorkListAgreesWithPassGate` (the fixture registers the chair AFTER the parties sit, because #927's blocking "register for this sitting" item, `sitting.go:122-125`, would otherwise hold `complete` false for a reason unrelated to materiality; table over `always`-low, `never`-high, `by_grade`-low, `by_grade`-medium and a STRANDED `by_grade`-low ancestor — blocking, `complete` false while it is open, never listed as "does not hold PASS", with the docket offer present: `complete` equals "the PASS appends", and `open[].material` matches), `TestPreChangeRegistryRefusedNamingMigrate`, `TestPreChangeDatabaseRefusedAtOpen` (a database whose `mint` table lacks `class_material`: `recordsql.Open` refuses naming the table, the column and `migrate`, and `dispatch next` over the same run prints that refusal, not "no such column"; the test logs the open's added cost), `TestMigrateSourceOpensPreChangeDatabase`, `TestStageForRunWritesByGrade`, `TestClassNewCarriesMaterialDefault`, `TestMintRefusesPresetClassMaterial`, `TestSeverityRefusalWhyIsClassAware` (minting without `--severity` shows the new `why`), `TestMigratingAdmitsArchivedPassOverNowAlwaysGap` (migrating a fixture whose archived PASS stood over an open low-graded gap of a class the shipped table makes `always` succeeds, and the PASS gate is on the migrated record), `TestLivePassOverOpenAlwaysGapRefused` (the same board, a live `Append`, is refused), `TestMigrateClassNewTakesCurrentRegistryDefault` (a pre-change `ClassNew` with no `material_default`), `TestMigrateRunWrittenByThisBinaryKeepsNewFields` (a run written by the new binary with a run-coined `class new --material-default always`, a `Mint` of that class, a `Mint` of a registry `never` class and a `spot-check --areas`: migration refuses nothing, and `class_material`, `material_default`, the staged registry rows and `areas` in the destination equal the source's), `TestPlanJSONEmitsArraysNeverNull` (an empty plan marshals `parties`, `docket`, `why` and `stale_areas` as `[]`), `TestChairWorkListStatesEveryGateRefusal` (for each Gate refusal in III.1's table, a board where only that refusal holds: the chair's list carries exactly one blocking item naming it, and `complete` is false; the FAIL-only convergence refusal adds none), `TestSetupRefusesRegistryWithoutMaterialDefault` (a `<cwd>/feov-memory/class-registry.json` row lacking the field, and one with an unknown value: setup exits 2 before any run directory exists, naming the file, the slug and the registry remedy, never `migrate`), `TestLoadRegistryRefusalNamesRegistryNotMigrate` (a current run whose staged registry was altered after setup), `TestMigrateStagedSlugAbsentFromTableFillsByGrade` (the row takes `by_grade`; the manifest's `stated_fills` names it; nothing is refused), `TestMigrateClassNewWithoutDefaultFillsFromTable` (a slug the table holds takes the table's value with no fill; one it lacks takes `by_grade` with a stated fill), `TestMigrateClassNewWithDefaultKeepsIt`, `TestChairDocketAffordanceOnlyOnMaterialGaps` (an open `never` gap and an open low `by_grade` gap get no `motion docket file` item; an open material gap and a stranded one do; a carried material gap gets the bench-remanded row in PLACE of the offer, since docketing it again is what the row tells the chair to do), `TestMigrationRecordsAdmittedPassOnArchives` (the b7 and b9 archives migrated: each PASS carries `migration_admitted_gap_ids` [G2] and [G3], `verify` reports `pass-closes-all-gaps` as admitted by migration naming the gap, and re-migrating keeps the field), `TestMigratedPassOverOnlyNonMaterialGapsCarriesNoAdmission` (b6), `TestLiveGateCarryingMigrationAdmissionRefused` (a PASS on a board a plain PASS clears, a PASS over an open material gap, and a FAIL, each carrying the field: refused for the field), and in verify `TestPassOverOpenNeverClassGapIsNotA67Violation` and `TestPassOverOpenMaterialGapWithoutAdmissionFails` (without the field, or with one naming another gap, the PASS FAILs naming the gap) | one class-aware definition (G2), the work list (audit gap 1), the pre-change refusals, the exemption | `replay.go`, `views.go`, `viewjson.go`, `sitting.go`, `verify.go`, `record.proto`, `recordsql/store.go`, `migrate.go`, `record.go` |
| 3 | `go -C $W/$T test -count=1 -run TestDefaultCastListsAllSevenAreas ./internal/record/` | III.8 | `cast.go` |
| 4 | Delete-the-row check: invert `streak >= 2`; delete the retired-for-good transition; delete the stale-area refusal; delete one `always` row; delete the refused lane-tag tell; drop the `material` arm in `sitting.go`; drop the `!Migrating` around the material gate; drop the open-time column check; drop the `r < D.at` bound in `lastSittingBefore`; drop the `!Migrating` gate on the preset-`class_material` refusal; make the migrate translation unconditional; drop the material filter on the docket affordance; drop the `gap_ids` comparison in `DispatchParityAudit`; compare against the group's FIRST row instead of its last; make `requireNoCastLensReady` admit a lens retired with its re-arm owed; delete one row's `material_default` from the registry (`classgen` must fail); hand-edit one value in `record/shippedclasses_gen.go` (`go -C $W/$S run ./classgen -check` must fail); drop the stranded item from the work list; drop the setup refusal; drop the stated-fill record; drop the admission stamp; drop the live refusal of a `Gate` carrying `migration_admitted_gap_ids`; stamp every open gap instead of the material ones; make `verify` ignore the field, or admit every open material gap whatever it names. Re-run #1, #1b, #2, #12 and #15; each MUST fail. Then restore. | the tests notice | the same files |
| 5 | `go -C $W/$T test -count=1 ./...` | whole module, including `TestVocabularyProse` (D4) and the vocabulary gate over `Supplies` | any `$T` file, agents, skills |
| 6 | `node --test $W/plugins/frank-exchange-of-views/tests/simulator/{debate,prompts}.test.mjs`. The assertion at `debate.test.mjs:73` ("HEAD MOVED PAST YOUR LAST SITTING (head 7)") becomes: every lens prompt carries the static `sitting.last_sitting` clause with all three kinds and the missed-then sentence, and no head clause built from the relay. New tests: for EVERY required relayed field (`head`, `parties`, `pass_permitted`, `ceiling`, `docket`, `why`, `max_epochs`, `epoch_limit_reached`, `stale_areas`, and a party's `seat_id` or `gap_ids`, and a `stale_areas` entry's `seat_id` (a string) or `pin` (an integer)), a plan missing it or carrying the wrong type (including `null` for an array) throws, naming the field — one table-driven test, so the check and III.4's list cannot drift; blue's prompt carries the closing-arguments clause exactly when the relayed `docket` is non-empty; seven default lenses; the chair prompt carries re-arm-once, the no-lens-ready PASS condition, the stale areas and the by-class listing. The prompt goldens that carry the old wording — `tests/simulator/testdata/prompt-red-chair.golden:41` and `prompt-red-lens-{evidence:16,logic:12,dark-side:12}.golden` ("HEAD MOVED PAST YOUR LAST SITTING") — are regenerated by #8 and reviewed line by line, together with `prompt-red-lens-evidence-engaged.golden` (the engaged lens prompt now carries the static `last_sitting` clause) and `seat-roster.golden` (reviewed for the seven-area default wherever the simulator uses it). | III.4, III.5, III.8 | `debate.js`, `tests/simulator/` |
| 7 | `go -C $W/$S run ./classgen`, then `go -C $W/$S run ./protogen`, then `go -C $W/$S run ./check -only classgen,schemagen,massgen,vocabdoc,validatejson,frontmatter,pluginparity,fixtureparity,mjsparity,archaeology,rulesweep,lawqueue,golden,feov-record` | generated carriers (`record.pb.go`, `testdata/schema.sql`), parity, vocabulary, the law queue after the deletion | registry, `record.proto`, agents, `debate.js`, `law/` |
| 8 | `go -C $W/$S run ./golden -update` once after the prompt and help changes (the command `tests/simulator/golden.mjs:65-70` names), committed on its own; then `go -C $W/$S run ./golden` (uncached), reviewing each changed page, including the four simulator prompt goldens #6 lists | the flag lists of `dispatch`, `spot-check`, `mint` and `class new`, and `verdict.md`/`dispatch.md` text | `cli/chair/*`, `cli/lens/*`, `cli/seat/help/*` |
| 9 | `FEOV_RELEASE_GATE=1 go -C $W/$T test -count=1 -timeout 45m ./releasegate/fuzz/` | seven-lens fuzz runs end in a terminal verdict, and none settles empty (#637/#870); with the `parties` patch deleted, the fuzz relays the verb's JSON unaltered, so a `null` array fails it; `outcome --as VERIFIED` is driven on seed 3 (`verifiedSeed`, APPLY and COUNTER gaps only) and asserted, because the draw reaches VERIFIED on about one run in ten at seven lenses: every gap left "at impasse, ruled remanded" holds the run at CEILING, about a quarter of minted gaps end there under the sweep's k 1 (153 of 558; 63 of 250 on `origin/main`), and seven lenses mint 7.0 per run against 3.1 (8 of 80 VERIFIED against 29 of 80). Before the forced seed the word rode on the estoppel seed's all-APPLY run. With every drawn seed forced, 39 of 39 reach VERIFIED | `dispatch.go`, `cast.go`, `fuzz_test.go` (the directives, `verifiedSeed`) |
| 10 | `go -C $W/$T build -o ~/.claude/scratch/feov-lens-review/bin/feov-record ./cmd/feov-record && go -C $W/$T run ./cmd/seatprobe -bin ~/.claude/scratch/feov-lens-review/bin/feov-record -board all -dir ~/.claude/scratch/feov-lens-review/seatprobe` (once, at the end; costs tokens). The seats load THIS tree's plugin: `-plugin-dir` defaults to `$W/plugins/frank-exchange-of-views`, and each board's own session `init` event is checked for that path, so a board scored against the installed release fails rather than reporting. **RUN 2026-09-16, haiku, 11 boards, 2 at a time, $3.39 over 13 minutes.** Ten boards dispatched and reported; `lens-ocr-verify` NOT BUILT (its citation needs the OCR engine build tag). Verbs recorded per board: `adjudicate` register; `arithmetic` edit, log, manifest-row, position, prove, register, revision; `audit` closing, inquiry-support, log, motion docket file, position, register, spot-check, verdict; `blocked` closing, edit, log, manifest-row, register, revision; `boundary` certify, halt, log, outcome, register; `docket` edit, line-of-inquiry, log, manifest-row, motion grade file, position, register, revision; `lens-audit` finding, log, mint, register, reproduce, verify; `lens-dispose` close, log, mint, register; `sitting` log, motion docket rule, motion petition rule, register; `sources` cite, edit, log, manifest-row, register, retire, revision. The chair reached `spot-check` and `verdict`, and a lens reached `mint` and `close`: the lens bar's own surfaces were read and acted on. 28 expectations went unmet and stand as the friction corpus — the probe reports, it does not gate. The run also found a defect in the instrument: the report read the run through the handle the BUILD opened, before the record leaves the run directory, so `recorded:` printed empty and every act a seat had recorded was filed under "invoked, no event". The run is re-opened for the report now; the figures above are that read over the same dispatch, and the broken read scored 19 more expectations unmet than the record supports. | real seats read the changed constitutions and reach their verbs | `agents/*.md`, the skill |
| 11 | `feov-record --seat-id operator migrate --from <run> --to <fresh>` for each of the **16** archives (extracted from `origin/main:run-archive/`) into scratch; then `work/classmaterial.py`, `work/passopen.py`, `work/gapagg.py` and `work/refusalrate2.py`. Expected, measured over the 16 archives on the rebased tree: zero refusals, and no stated fill for any staged row (every staged registry holds 38–39 slugs, all in the shipped table); exactly 29 stated fills, one per archived class-creation slug absent from the table (30 class-creation events; `structure-noncompliance` is in the table); every staged registry carries `material_default`; 123 gaps, material 82 today and **101** under class defaults (21 raised by an `always` class, 2 lowered: sleeper's `structure-noncompliance`), open-at-end material 3 → 9; b7's PASS carries `migration_admitted_gap_ids` [G2] and b9's [G3], the other four PASSes carry none, and `feov-record verify` exits 0 on all 16, reporting b7 and b9's `pass-closes-all-gaps` as admitted by migration. Each per-run difference against lens-quant §2 is explained by a class default. Re-run 2026-09-16 against the rebased tree, whose migration registry has since gained main's cite-origin entry: every figure above holds unchanged, and `verify` exits 0 on all 16. | the migration translation, on real data | `migrate.go`, `views.go`, `replay.go` |
| 12 | `go -C $W/$T test -count=1 ./internal/reportvoice/ ./internal/record/ ./internal/cli/...` running `TestRefusedTellsAreUnambiguous` (seat ids, finding labels, gap ids with process words, lane tags), `TestOrdinarySubjectProseIsClean` (plus "the G20 summit", "the chair of the committee", "a sitting judge", "a gap 2 metres wide", and "the red team exercise", which is advised and not refused), `TestMintRefusesSeatIdInProblem`, `TestMintRefusesLaneTagInRequiredFix`, `TestMintAcceptsProcessWordsInMintReason`, `TestMintAdvisesOnThisRun`, `TestMigratingReplaySkipsVoiceRefusal`, `TestMintRefusesAnUnambiguousTellInFixNewThroughTheRealVerb`, `TestMintAdvisesOnAnAmbiguousTellInFixNewThroughTheRealVerb`, `TestMigratingReplayKeepsAVoicedFixNew`, `TestMigratingReplayKeepsAVoicedCorroborationTitle`, and the blue advisory tests unchanged | III.9(c) | `tells.go`, `record/redvoice.go`, the Mint and Verify validations in `record.go`, `cli/lens/mint.go` |
| 13 | `go -C $W/$T test -count=1 -run TestRiskMatrixCarriesOnlyWhatMintWrote ./internal/report/`; blanking the `RequiredFix` cell must fail it. Then `go -C $W/$T test -count=1 -run 'TestCorroborationCarriesOnlyItsSourceIntoTheReport|TestCorroborateAdvisesOnItsTitleThroughTheRealVerb|TestCorroborateRefusesAnUnambiguousTellInItsTitleThroughTheRealVerb' ./internal/cli/`; printing anything beyond title, URL and date in the source's note or its Bibliography line, dropping the title advice, refusing an ambiguous tell, or dropping the title refusal must fail them | III.9 names the whole red path into the report | `report/assemble.go`, `docs.go`, `cli/lens/verify.go` |
| 14 | Re-run S1–S5 (§II). S1's only thresholds are the `by_grade` constant and its two carriers; S2 and S3 show every decider and gate statement on the class-aware definition, each remaining line in III.1's non-carrier list; S4 shows only III.1's "stays true" and "another sense" rows; S5 shows no old-model text outside III.5's non-carriers. A new line in any sweep is classified before merge — the counts on the implemented tree are S1 141, S2 146 (81 outside tests), S3 26, S4 205, S5 55, and they are recorded here because a sweep with no previous figure cannot say which of its lines is new. Re-run 2026-09-16: one line since phase 7's census, `releasegate/fuzz/fuzz_test.go`'s "dispute-lost docket holds the PASS, and the estoppel run exists for its refused mint" — a comment on the fuzz scenario, not a materiality decider, and in a test file, so S2's outside-tests count is unmoved. | no carrier still speaks the old model | any file the sweeps reach |
| 15 | `go -C $W/$T test -count=1 ./internal/capture/...` running `TestDispatchedPartiesAndRegistersAgree` (unchanged) plus `TestRelayedGapIDsDisagreeingWithDispatchRowFail` (a record whose dispatch engages `blue-respond` on G1, and a journal whose relayed plan carries `blue-respond` on G2: FAIL naming the sitting, the seat, G2 relayed and G1 recorded), `TestRelayedHeadDisagreeingWithPinFails`, `TestRelayCountMismatchFails`, `TestRelayPairingOnDuplicatedChairResult` (a journal carrying one chair result twice, the resume shape: FAIL on the count, naming it), `TestAdoptionTextNamesMaterialDefault` (capture's `law/proposed/class-*.md` text names the field and the run's coined value), `TestRelayComparedAgainstLastRowOfDuplicatedGroup` (a group whose seat is named by two rows — a docket plan recorded twice, as B5 and B6 did — with different `gap_ids`: a relay matching the last row PASSes, one matching only the first FAILs), `TestFaithfulRelayFieldsPass`, and `TestNoJournalStatesFieldsNotCompared` | fork (b) | `capture/dispatchparity.go`, `capture.go` |

| 16 | `go -C $W/$S test -count=1 ./classgen/` running `TestClassgenRefusesRowWithoutMaterialDefault`, `TestClassgenRefusesUnknownMaterialDefault` and `TestClassgenCheckCoversShippedTable` | registry rows carry the field; `-check` covers both generated files | `scripts/classgen`, `feov-memory/class-registry.json` |

**Implementation order.** Each phase ends green on the named rows before the next starts.
0. Rebase (Step 0) and `go -C $W/$T test -count=1 ./...` on the untouched tree, to record the baseline.
1. **Registry and schema.** The registry values and promotion, the `classgen` table and validation, the
   proto fields, `EventSchema`, the setup refusal, `loadRegistry`, the `Mint` stamp, the view column, the `Family`
   fold, and the open-time column check. → #7, #16, #2 (materiality, registry, setup and pre-change tests), #5.
2. **Deciders and the work list.** `dispatch.go`, `refs.go`, `convergence_refusal.go`, `views.go`, `verify.go`,
   `sitting.go`, `viewjson.go`, `available.go`. → #2 (work-list and gate tests), #5.
3. **Retirement, the PASS lens gate, stale areas, `last_sitting`, and the seven-area cast.** → #1, #1b, #3, #5.
4. **Migration.** The translation, stated fills and the `Migrating` exemptions. → #2 (migrate tests), #11.
5. **Relay and capture.** The `debate.js` checks, `Plan` arrays, the parity field check. → #6, #15, #9.
6. **Red's report voice.** → #12, #13.
7. **Text carriers.** Skills, agents, help, prompts, docs, `law/` files, goldens. → #8, #6, #7, #14.
8. **Whole.** → #4, #5, #9, then #10 once.

**Carried from audit** — each is a required test, discharged in the PR body by name:
- Round 2: `TestChairWorkListOverOpenNeverClassGap`, `TestChairWorkListAgreesWithPassGate`,
  `TestPreChangeDatabaseRefusedAtOpen`, `TestMigratingAdmitsArchivedPassOverNowAlwaysGap`,
  `TestLivePassOverOpenAlwaysGapRefused`.
- Round 3: `TestLensWorkCarriesLastSitting` (the fixture registers first), `TestMigrateRunWrittenByThisBinaryKeepsNewFields`,
  `TestPlanJSONEmitsArraysNeverNull`, #6's required-field table test, `TestChairDocketAffordanceOnlyOnMaterialGaps`,
  `TestRelayedGapIDsDisagreeingWithDispatchRowFail`.
- Round 4: `TestPassRefusedWhileReArmOwed`, `TestPassPermittedAndGateShareTheFold`,
  `TestRelayComparedAgainstLastRowOfDuplicatedGroup`, `TestRelayPairingOnDuplicatedChairResult`.
- Round 5, gap 1: the stranded row of `TestChairWorkListAgreesWithPassGate`, and `TestChairWorkListStatesEveryGateRefusal`.
- Round 5, gap 2 / fork 1: `TestMigrateStagedSlugAbsentFromTableFillsByGrade`, `TestMigrateClassNewWithoutDefaultFillsFromTable`,
  `TestMigrateClassNewWithDefaultKeepsIt`.
- Round 5, fork 2: `TestSetupRefusesRegistryWithoutMaterialDefault`, `TestLoadRegistryRefusalNamesRegistryNotMigrate`.
- Round 5, gap 3: `TestClassgenRefusesRowWithoutMaterialDefault`, `TestClassgenRefusesUnknownMaterialDefault`,
  `TestClassgenCheckCoversShippedTable`, `TestAdoptionTextNamesMaterialDefault`.
- fork (a) rebuilt (gblock, 2026-09-15): `TestMigrationRecordsAdmittedPassOnArchives`, `TestLiveGateCarryingMigrationAdmissionRefused`,
  `TestPassOverOpenMaterialGapWithoutAdmissionFails`, `TestMigratedPassOverOnlyNonMaterialGapsCarriesNoAdmission`.
- phase 6 (red's report voice, the corroboration path): `TestCorroborationCarriesOnlyItsSourceIntoTheReport`,
  `TestCorroborateAdvisesOnItsTitleThroughTheRealVerb` (its premise is now an ambiguous tell only).
- phase 7 rulings (gblock, 2026-09-15; the corroboration title and `fix_new`): `TestCorroborateRefusesAnUnambiguousTellInItsTitleThroughTheRealVerb`,
  `TestCorroborateAdvisesOnItsTitleThroughTheRealVerb`, `TestMintRefusesAnUnambiguousTellInFixNewThroughTheRealVerb`,
  `TestMintAdvisesOnAnAmbiguousTellInFixNewThroughTheRealVerb`, `TestMigratingReplayKeepsAVoicedFixNew`,
  `TestMigratingReplayKeepsAVoicedCorroborationTitle`; each fails with its fix reverted.
- phase 8 (the release sweep): the forced-VERIFIED seed's assertion in `TestFuzzDebate`, which fails with the seed's scenario override removed.
- Every item in #4's delete-the-row list fails its test when applied.

Done when #1–#9 and #11–#16 are green and observed, and #10 has run once.
