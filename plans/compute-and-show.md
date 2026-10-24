# Compute and show — what we computed, what a source shows, and how either is seen

> STATUS (2026-10-24): proposed, not started; revised after audit round 2. Written against
> `d9b2c823` (code identical to `62fe337e`; `EventSchema` 37; 29 archived runs). Forks and rulings
> by gblock; two design gaps resolved by the lead (Forks). Census ledgers:
> `~/.claude/scratch/compute-plan/round2/census.tsv` (Go), `round3/c3-ledger.tsv` (every tracked
> non-Go file), `round3/c10-*.txt` (archive text).

## I. Summary & Goals

**Objective.** The report holds two kinds of evidence and can show neither.

- **What a source shows** — `cite` — backs a sentence with a footnote.
- **What we computed** — `prove` — backs a sentence with a footnote; its output is a code fence in
  evidence.md.

No reader sees a table, chart or image built from evidence: a markdown image renders as a link
(`report/md.go:437`), and nothing builds one. `prove` carries three misfits:

1. **The name.** The record already calls the concept a computation (`check_kind: computation`, the
   computation lens); the code calls the `observed` basis "a MEASUREMENT, not a proof"
   (`proof/proof.go:53`). 29 archived runs hold 157 stored scripts: 198 proof events in the 23
   SQLite-era runs, 41 in the 6 era runs.
2. **One file, language by extension.** `proofs/<sha>/script.<ext>`, the interpreter picked from a
   five-row table (`proof/interpreter.go:37`), run with the RUN directory as working directory
   (`proof.go:256`). A helper, a query file or data has nowhere to live, and a script that reads
   relative to itself is graded on a path no re-runner uses: quadratic-formula's 28 proofs are all
   recorded `reproducible`; re-run from the archive, 17 of its 26 stored scripts reproduce neither
   from the store nor from a fresh copy (`round2/c9-archive-scan.txt`; the comment at
   `proof.go:62` says six).
3. **Inputs nobody can reach later.** A script reading a cached source works in the run and fails
   from the archive, which omits `cache/` (`capture/capture.go:2007`).

**Why `--show` (R-show).** A table typed into the report is a second copy of the computation's
output, and copies drift — the defect class `figure_recount_fails` (`record.proto:436`). An image
cannot be typed at all. `--show` makes the displayed content BE the output: the tool places what
the program wrote, so there is nothing to recount.

**The design is two mirrored verbs and one placement rule.**

| | `cite` — theirs | `compute` — ours (replaces `prove`) |
|---|---|---|
| Records | a source's bytes, by sha256 | a directory and a command, by sha256 |
| Store | `cache/` (never archived) | `computations/` (archived) |
| Backs a sentence | `--quote` | `--quote` |
| Shown in the report | `--show image\|text` (2) | `--show table\|image\|text` (1c) |
| Checked by | the evidence lens — **fidelity**: is this what the source shows? | the computation lens — **re-run and soundness** |
| Re-derived from the archive by | `archive rerun` re-fetches and verifies (1b) | `archive rerun` re-runs (1b) |
| Joined by | | `--input <citation>` (1b): what we compute is built from what sources show |

**Why two verbs.** "The source shows this" and "we computed this" are different claims, checked by
different tests. A program between a source and its displayed image is a place for distortion
re-running cannot catch — a recolored crop reproduces byte for byte — so a source's image is
extracted by the tool and checked for fidelity; a chart we drew is re-run.

**The placement rule.** Everything backs a sentence; nothing floats. `--show` adds a display:
assembly places the content after the paragraph holding the anchor, numbered in document order per
kind ("Figure n", "Table n"), captioned by the act's `--reason` — the field that is already its
footnote. A shown block leaves the report as every anchor does: `edit` the sentence, then `retire`
(`anchor/anchor.go:75`). No new lifecycle.

**Parts (R-split).** Each ships alone, with its own epoch and `migrate` translation.

| Part | Scope | Depends on | Epoch |
|---|---|---|---|
| 1a | `prove` → `compute`; directory + argv; two-value basis; per-act store; `K-` anchors | — | E1 |
| 1b | `--input` (raw cited bytes, read-only); package `archive`; operator `archive rerun` | 1a | E2 |
| 1c | `--show table\|image\|text` on `compute`; images exist only as shown blocks | 1a | E3 |
| 2 | `--show image\|text` on `cite` | 1b (`archive rerun` re-derives crops), 1c (placement, `Show`) | E4 |

E1–E4 are the next free epochs when each part lands.

**Success criteria.**

| # | Part | Criterion | §V |
|---|---|---|---|
| S1 | 1a | `feov-record blue compute --dir D --quote S --reason R -- python3 main.py` records a computation: bundle = every regular file under `D`; command = the argv after `--`; toolchain = resolved path + sha256 of `python3`. `blue prove` is an unknown command | V1a.1 |
| S2 | 1a | Two runs in fresh copies outside the run tree, each copy's root replaced by `<bundle>` in the output. Equal output, exit and `out/` → `REPRODUCIBLE`; different with no `--variance` → refused naming what moved; with `--variance` → `VARIANT` + drift; `--variance` over agreeing runs → refused | V1a.2, V1a.3 |
| S3 | 1a | A recorded act never changes after write: its output and `out/` live under its own id; `reproduce` compares the anchor's own | V1a.4, V1a.5 |
| S4 | 1a | A keyed retry restates every recorded flag or is refused; program, argv and inputs are in the sha | V1a.6 |
| S5 | 1a | Every archive migrates with zero refusals; III.1.6's assertions hold; every computation that re-ran equal from the old store (107) re-runs equal from the migrated one, by the shipped re-run | V1a.7 |
| S6 | 1b | `--input C-…` materialises the cached bytes read-only at `sources/C-…`; the archive commits no source bytes; `archive rerun` on an extracted archive with no `cache/` re-fetches, verifies sha256 and re-runs, with Part 1b's exit semantics | V1b.1–3 |
| S7 | 1c | `--show table` renders the CSV under "*Table n.* R"; `--show image` embeds PNG/SVG under "*Figure n.* R" in report.md and (as `data:` URI) report.html; `--show text` inlines markdown refused where `edit` refuses it | V1c.1 |
| S8 | 1c | An image in authored text is refused at `edit` and `ingest`; an image in any other seat text renders as a link | V1c.2 |
| S9 | 1c | Shown blocks appear in `show report --assembled --json`; the bare read carries the anchor only and says so | V1c.3 |
| S10 | 2 | `cite --show image --page N --region …` displays a tool-rendered crop the evidence lens verifies against `render-page`; `cite --show text` displays the quoted span; no crop is committed; an absent crop assembles as a loud placeholder | V2.1–3 |

**Value case — what goes away.**

| Removed | Replaced by |
|---|---|
| The extension→interpreter table, its Windows extra and `argvFor` (`interpreter.go`) | the argv after `--` |
| The grading run from the store (`executeStored`, `proof.go:202`) | two runs, each in a fresh copy at a different path |
| Three bases as free strings (`proof.go:72–74`) | a two-value enum |
| Silent `observed` for output that moved (`proof.go:156`) | refusal, or a declared invariant — the tool never infers |
| Working directory = the run tree (`proof.go:256`) | working directory = a copy of the bundle |
| One `output` per sha, overwritten by the next act (`proof.go:345`) | one output per act, written once |
| Typed tables that drift from their computation | the program's own output, placed by the tool |
| "Images render as links" | images exist only as shown blocks, embedded from the run's store |

**Not in scope.** Rich media and interactive charts. A machine-checked basis (Lean): the argv holds
`lake env lean X.lean`; a checker for "no `sorry`" and Mathlib are missing. A filesystem or network
sandbox (R-fs; #1560). Extracted text as an input (R-input). Output-signature scanning
(`ErrorSignature`, `--expect-error`) keeps its behaviour (#1466).

## II. Technical Context

- Go, FEOV tools module. Record: SQLite generated from `record/recordpb/record.proto` (`Proof`
  :1366, `Reproduce` :1421, `Soundness` :538, `Rerun` :552, `CHECK_KIND_COMPUTATION` :343,
  `EVENT_TYPE_PROOF` :278, event oneof :718). Epoch: `requirements.json` → `scripts/schemagen` →
  `recordpb/schema_gen.go:18`. The DDL generator types string, bool, int32/64, bytes and enum only
  (`recordsql/schema.go:390`); a list becomes a child table of scalars (`:617`). Two-field rules are
  `option (check)` (precedent `:843`, `:983`); removed fields leave `reserved` names (`:869`).
- Anchors: one table (`anchor/anchor.go:63`). `P-` becomes `K-` (F1); `K` is held by no id kind.
- `migrate`: channel copy (`record/migrate/migrate.go:170`), id respelling (`remap.go`: `currentID`
  :84, `hexID` :86, `archivedToken` :90, `id()` :217, field lists :109–140), era mapping
  (`eraentries.go:54`; era proof events carry their own `output`), plan-id translation
  (`planids.go`).
- Execution posture is unchanged: a seat could already run anything through Bash; `compute` adds
  recording, not reach. No sandbox (`nonet` is test-only).
- Fetch cache: `fetchcache.Read` :374 by sha, `fetchcache.Resolve` :518 by URL; `cite` events carry
  `url` and `sha256`; a cache entry carries `TDMReserved` (`fetchcache.go:204`), which nothing
  consults; the rights note is `fetchcache/shell.go:215–226`. Page rendering:
  `fetchcache/pagerender.go`, lens verb `render-page`.
- Site: report.md and report.html are written at the run root (`report/docs.go:388`,
  `report/assemble.go:296`); report.html is one self-contained file (`report/site.go:57`); markdown
  tables render (`md.go:188`); mermaid is the one external fetch (`site.go:180`).
- Run-archive tarball: written by `capture.go:2016` (`ArchiveRecord`), read by two private copies —
  `recordtest/archive.go:34` and `devcmd/archdump/main.go:149`.
- FEOV's whole suite runs on the Windows pull-request leg with `-short` (`.github/workflows/hooks.yml:693`).

## III. Proposed Changes

### III.0 The record and the store (contract: event schema, store layout)

```proto
// Computation is a program blue ran twice on a bundle, and what it wrote. EVENT_TYPE_COMPUTATION = 24.
message Computation {                                     // [MODIFY] replaces Proof (oneof :718, = 42)
  optional string computation_id   = 1;                   // K-… (F1)
  optional string key              = 2 [(sql) = { flag: "key" }];
  optional string bundle_sha       = 3;                   // III.1.3
  optional string answers          = 4;
  optional string cites            = 5;                   // the METHOD's citation
  optional string drift            = 6;                   // what moved between the two runs
  optional string text             = 7 [(prose) = true, (sql) = { flag: "reason" }];
  optional int32  exit             = 10;
  optional string location         = 11 [(prose) = false, (sql) = { flag: "quote" }];
  repeated string argv             = 12;                  // the command after --
  optional string toolchain_path   = 13;                  // argv[0] resolved; absent for a bundle file
  optional string toolchain_sha256 = 18;
  optional Basis  basis            = 14;
  optional string invariant        = 15 [(prose) = true, (sql) = { flag: "variance" }];
  repeated string inputs           = 16 [(sql) = { flag: "input" }];   // 1b
  optional Show   show             = 17 [(sql) = { flag: "show" }];    // 1c; absent = not shown
  option (check) = { expr: "VARIANT ⇔ drift and invariant present; REPRODUCIBLE ⇔ both absent" };
  option (check) = { expr: "show absent or not SHOW_UNSPECIFIED" };
  reserved 8, 9;
  reserved "proof_basis", "script";
}
enum Basis {
  BASIS_UNSPECIFIED  = 0;
  BASIS_REPRODUCIBLE = 1 [(means) = "both runs wrote the same output, exit and out/ files, byte for byte"];
  BASIS_VARIANT      = 2 [(means) = "the two runs differed, and blue stated what holds across them"];
}
enum Show {                                               // 1c (compute), 2 (cite: IMAGE, TEXT)
  SHOW_UNSPECIFIED = 0;
  SHOW_TABLE = 1 [(means) = "the program's CSV, rendered as a numbered table after the anchor's paragraph"];
  SHOW_IMAGE = 2 [(means) = "an image, embedded as a numbered figure after the anchor's paragraph"];
  SHOW_TEXT  = 3 [(means) = "markdown, placed after the anchor's paragraph"];
}
// Reproduce [MODIFY]: proof_sha = 1 → bundle_sha = 1; proof_id = 7 → computation_id = 7;
//   + optional string drift = 9;  check: present ⇔ rerun = DIFFERS
// Rerun (means) [MODIFY]: REPRODUCED = "re-ran the stored bundle; output, exit and out/ equal the act's recorded ones";
//   DIFFERS = "… output, exit or out/ differ; drift names what moved". CALLS_THE_TOOL, NOT_RUN: "bundle" for "script".
// Cite [MODIFY, 2]: + optional Show show = N; optional int32 page = N+1;
//   optional int32 region_x0 = N+2, region_y0 = N+3, region_x1 = N+4, region_y1 = N+5;  // render-page pixels
//   checks: show ∈ {IMAGE, TEXT}; page and region ⇔ IMAGE; x0 < x1, y0 < y1
```

`inputs` is the first list with an `(sql) flag`: `flagOf` (`record/supersede.go:285`) and `flagFor`
(`recordpb/requiredfields.go:105`) read the flag whatever the cardinality; schemagen's output is read
in the `schema.sql` golden. Toolchain is two scalars and the region four `int32`s because the
generator has no message or double kind. `awaiting_proof` → `awaiting_computation`. Every `(means)`
naming a proof (:280, :343, :372–375, :403, :448, :451, :454, :575, :1125) names a computation.
remap's field lists class every new field: `inputs` id, `invariant`/`drift`/`Reproduce.drift`
argument, `argv`/`toolchain_*` data. `keycensus_test.go` `notFields` gains `proof_basis`,
`proof_id`, `proof_key`, `proof_sha`, each naming its replacement.

```
<runDir>/
├── computations/<bundle_sha>/
│   ├── dir/                 the bundle — content-addressed, written once, never rewritten
│   └── acts/<K-id>/         one per recorded act, written once at its write
│       ├── output           everything the program wrote; each copy's root reads <bundle>
│       └── out/             the files it wrote there (1c shows one)
└── cache/                   fetched bytes; Part 2's derived crops — never archived
<tmp>/…/<root>/              each run's copy, outside the run tree, two at different depths:
                             dir/ contents + empty out/ + sources/C-… (1b, read-only)
```

### Part 1a — `compute` replaces `prove` (epoch E1)

#### III.1.1 The verb [MODIFY → rename] (contract: CLI flag set)

```
feov-record blue compute --dir <bundle> --quote <sentence> --reason <note>
                         [--variance <invariant>] [--answers G-…] [--cites C-…] [--key k] [--expect-error]
                         -- <command> [args…]
```

| Rule | Refused when |
|---|---|
| `--dir` replaces `--script`; a one-file computation is a one-file directory | a symlink, a non-regular file, a bundle over 1 MiB, or an entry named `out` or `sources` (the tool's mounts) |
| argv[0]: with a separator and relative (`./run.sh`) → a bundle file, run from the copy, its sha already in the bundle, no toolchain; absolute → that file; bare → resolved on PATH; toolchain = resolved path + sha256 | argv empty; a relative argv[0] that is not a bundle file with its executable bit; a bare argv[0] not on PATH (named) |
| `--variance` states what holds when the runs differ | given and the runs agree (precedent: `--expect-error` with no failure, `cli/blue/prove.go:128`) |
| Recorded flags restate on a keyed retry (`seat.RepeatsKeyed`, `cli/seat/keyedretry.go:31`) | asserted set (`prove.go:191`) + `--variance` → `invariant` differs; a different sha under the key → Conflict (`prove.go:83`) |
| Output and `out/` have no size limit (R-size) | — |

#### III.1.2 Execution [MODIFY/DELETE] — `internal/proof` → `internal/compute`

1. Copy the bundle twice, to two fresh directories at different depths, neither inside the run tree;
   each gets an empty `out/`.
2. Run each with the copy as working directory, under the checker environment
   (`FEOV_PROOF_RERUN` → `FEOV_COMPUTE_RERUN`, `seatenv/seatenv.go:87`) and the pins
   `SOURCE_DATE_EPOCH=0`, `PYTHONHASHSEED=0`, `TZ=UTC`, `LC_ALL=C`, `PYTHONUTF8=1`,
   `MATPLOTLIBRC=<a tool-written file outside both copies holding svg.hashsalt>` (matplotlib salts SVG
   ids with uuid4 unless `svg.hashsalt` is set).
3. Take everything the program writes — stdout and stderr as one stream (`CombinedOutput`,
   `proof.go:258`) — and replace each occurrence of that copy's root path, the exact string the tool
   chose, with `<bundle>` (P-norm). `ErrorSignature` scans the result. `out/` files are compared and
   stored as bytes, untouched.
4. Refuse at write: a tool call in either run (today graded `not_portable`, `proof.go:173`); an
   environment error without `--expect-error`; differing output, exit or `out/` without `--variance`.
5. Write `dir/` if absent (present → its listing must hash to `bundle_sha`), then
   `acts/<K-id>/output` and `out/` — refused if `acts/<K-id>/` exists. An idempotent keyed retry
   writes nothing.
6. Remove both copies, clearing read-only bits first (Windows `RemoveAll` refuses read-only files).
7. Delete: `interpreter.go`, `executeStored`, `NotPortable`, `Observed`, `StoredScript`, `store`'s
   per-sha `output`.

#### III.1.3 Identity

`bundle_sha` = sha256 over a sorted canonical listing, one line per item: `argv <i> <arg>`; `input
<C-id> <sha256>` (1b); `file <path> <mode> <sha256>`. A computation with no inputs has no `input`
lines, so 1b leaves every 1a `bundle_sha` unchanged. Computed without running, so a keyed retry
tells a crash-retry from a collision before executing (as `ScriptSha` does today).

#### III.1.4 Re-run [MODIFY]

`reproduce` and capture's audit (`capture/proofrerun.go`, check `proof-rerun` → `computation-rerun`)
call one entry, `compute.Rerun(run, K-id)`: III.1.2 steps 1–3 and 6 from `dir/`, compared against
`acts/<K-id>/` — output, exit and `out/`; an absent recorded `out/` equals an empty one.
`Reproduce.drift` names what moved. On `VARIANT` a differing re-run is expected: the lens's reason
says whether the output satisfies the invariant (F2), and capture counts it `moved`, never FAIL
(`proofrerun.go:155`).

#### III.1.5 Rename sweep [MODIFY] — logic unchanged

| Site | Becomes |
|---|---|
| anchor kind `proof`/`P-`/tag `proof` (`anchor.go:63`) | `computation`/`K-`/tag `computation` |
| `proofNote` `[^P%d]` (`layout.go:591`), `wovenNote` `(?i)\[\^p?[0-9]+\]`, `WovenNotesRule` "P and digits" | `[^K%d]`, `k?`, "K and digits" |
| `weaveProofRefs` (`report/proofs.go:49`), evidence.md `## Proofs` | `weaveComputationRefs`, `## Computations` (argv, toolchain, basis, invariant, drift, each bundle file fenced by extension, output fenced) |
| `awaiting_proof` (`recordsql/views.go:550`), `ProofAnswers` (`estoppel.go:255`), `proofChecksOwed`/`proofOwnerOf` (`available.go:522`/`:557`), "unjudged proof" (`retirement.go:156`), `proofs_unverified` (`evidenceview.go:314`) | `…computation…` |
| `EvidenceProofAtJSON` `script`/`script_file` (`evidenceview.go:588`; `verbs.go:299`) | `files` (path, body) and `output` |
| `ArchiveRecord` `proofs/` (`capture.go:2066`, comment :2006); `site.html.tmpl:33` "`records/` and `proofs/` are what both were derived from" | `computations/` |
| terms.json :574 :586 :587 :965 :976 :980 ("a proof", "citation or proof") | "a computation", "citation or computation"; `vocabdoc` regenerates `docs/vocabulary.md` |

#### III.1.6 migrate — one translation [NEW] (contract: record migration)

| Old | New |
|---|---|
| `Proof` / era `proof` event (`eraentries.go:54`) | `Computation` |
| `proofs/<sha>/script.<ext>` (sha dirs only) | `computations/<new>/dir/script.<ext>`; `<new>` from the old store alone (argv + tree) |
| `proofs/<sha>/output` | `acts/<K-id>/output` for each act on that sha (SQLite era); the era event's own `output` (JSONL era) |
| extension | argv: `.py` `python3`, `.sh` `bash`, `.js`/`.mjs` `node`, `.go` `go run` — each + `script.<ext>`; toolchain absent |
| `proof_basis` `reproducible` / `observed` / `not_portable` | `REPRODUCIBLE` / `VARIANT`, invariant = drift (F4) |
| `proof_sha` in `Reproduce` and the manifest | `bundle_sha` through the old→new map |
| `P-`/`p-` + 8 hex | `K-` + the same hex: `hexID` (:86) and `archivedToken` (:90) admit `P-` — today `currentID` (:84) matches `P-` and `id()` (:217) would otherwise hash a new hex |
| `<!--proof:…-->`; `[^Pn]` in argument fields | `<!--computation:K-…-->`; `[^Kn]` |

| Archive (measured; `round2/c9-archive-scan.txt`) | Holds | Asserted after migrate |
|---|---|---|
| all 29 | 157 stored scripts | zero refusals; no `\bP-[0-9a-f]{8}\b` outside `Plan.printed`, no `<!--proof:`, no `[^P`; every `K-` id's hex equals its source id's; every act has `acts/<K-id>/output` |
| 2026-09-09_is-91-prime-b | `not_portable` `p-808ae840` | `VARIANT`, invariant = drift "runs from the proof store and says something else — exit code moved 0 -> 1 between runs" |
| 2026-08-22_is-7-prime-run2 (era) | `observed` `p-568517ec` | `VARIANT`, invariant = drift "output differs from byte 131 between runs"; its output is its event's |
| 2026-10-12_is-91-prime-wave-b | `[^P5]` in a finding | `[^K5]` |
| 2026-10-10_is-91-prime-m19 | 125 `P-` ids | `K-`, same hex |
| 2026-09-11_is-91-prime-b6 | the `.go` script | argv `go run script.go` |
| 2026-10-17_is-91-prime-07d | 9 acts over 6 scripts | 6 `dir/`, 9 `acts/` |

**Residue, accepted** (`round3/c10-*.txt`, `outputs-with-paths.txt`):

| Residue | Size | Why it stays |
|---|---|---|
| `P-` ids and "unjudged proof" in `Plan.printed` | 34 ids, 3 runs (m28, m29, m30) | printed plans keep the wording they were printed in (`planids.go:15–17`) |
| `proofs/<sha-prefix>` paths in seat text (`finding.text`, `reproduce.note`, `close.*`, `log.text`, …) | 51, 10 SQLite runs; 0 era | a seat's prose account of where it read, abbreviated 8–64 hex with `…`/`...`; recovering a path from prose is a pattern over free text — the act it describes keeps its id |
| bare old-sha prefixes (≥ 8 hex) in seat text | 155, 17 SQLite runs | same |
| store paths inside program output | 9 `Reproduce.observed_output`, 1 stored `output` | bytes the program wrote; that one script reproduces neither way |
| acts sharing a sha receive the sha's one stored `output` | 104 acts on 26 shas, 19 runs | all 104 recorded `reproducible`: the one output is each act's by its basis |
| `proofs/synth/` | 17 drafts, 1 run | each is also stored under its sha |
| `meta.json` in sha dirs | 37 files, 6 era runs | its fields are the era event's |
| author script path (`Proof.script`); `Reproduce.drift` | every proof; migrated re-runs | the argv replaces it; not recorded then |

The lens help says a migrated `VARIANT`'s invariant describes what moved, not what must hold.

#### III.1.7 Agent-facing text [MODIFY] (contract: prompts and constitutions)

Sources only, then `agentgen`, `enginegen`, `vocabdoc`: `help/prove.md` → `help/compute.md`;
`help/reproduce.md` (`:9` "runs the stored script … in the run directory", `observed`/`not_portable`
:21); `help/{dispatch,close,mint,edit}.md`;
`scripts/agentgen/src/{blue-researcher,blue-synthesizer,red-lens-computation,red-lens-evidence}.md`,
`fragments/{blue-search-and-cite,lens-audit-and-mint}.md`, both preloaded-skill sources;
`skills/research/SKILL.md:7`; debate.js :229 :538 :923 :969 :991 :1134;
`references/report_template.md:147`, `references/run-directory.md` (:28 "with script", :77);
terms.json (III.1.5); FEOV docs (ledger §C3); `docs/local-development.md:226`; `CLAUDE.md:34`;
`scripts/universe-stream.py:24`. Seat prompts name the act — "compute it"; the prompt gate matches
`<role> <verb>` (`seatprobe/naming.go:78`), so "blue compute" fails it. The `compute --help` page
teaches: a directory and a command; output records `<bundle>` for the copy's path; make output
deterministic or state the invariant; output and `out/` are unbounded and committed.

### Part 1b — `--input` and `archive rerun` (epoch E2)

| Change | Detail |
|---|---|
| `compute --input C-…` (repeatable) [MODIFY] | the citation's RAW cached bytes (`fetchcache.Read`) — never extracted text: OCR is non-deterministic, so `archive rerun` could not reproduce an extraction (R-input). Placed read-only at `sources/C-…` in each copy; id + sha256 enter `bundle_sha`. Refused: a citation not on the record, or bytes the cache does not hold |
| `reproduce`, capture's audit [MODIFY] | `compute.Rerun` materialises `sources/` the same way |
| Package `archive` [NEW] | the one writer and the one reader of the run-archive tarball: `archive.Write` (from `ArchiveRecord`, `capture.go:2016`; `capture` calls it) and `archive.Extract` (replacing `recordtest/archive.go:34`'s loop and `archdump`'s `untar`, `main.go:149`). Capture's transcript tarball (`capture.go:1718`) is another format and stays. It commits no `cache/` (R-commit): the record holds each citation's URL and sha256, never its bytes |
| `feov-record archive rerun <runDir>` [NEW] (N1) | input: an extracted archive brought to the current epoch by `migrate` (an older epoch is refused by `requireDeclaredSchema`, pointing to `migrate`). Re-fetch each input by its `url` (`fetchcache.Resolve`, paced), verify sha256, then `compute.Rerun` every computation. Writes no event. `--citations` verifies every citation the same way (N2) |
| Rights [MODIFY] `fetchcache/shell.go:223–226` | the comment states the decision: a seat's program computing over a cited body is accepted, whatever `TDMReserved` says; its output and `out/` are committed to `run-archive/` and may carry more of a source than a quotation; the bodies themselves are never committed, kept as a corpus beyond the cache, or trained on. `TDMReserved` is not copied onto `Computation`: it is a fact about the source, kept beside the citation's other source facts, and no reader acts on it (R-rights) |
| Help, prompts [MODIFY] | `compute --help`: inputs are citations, raw bytes, read at `sources/<id>`; blue-researcher source: compute from what you cited |
| migrate | replay only: migrated computations hold no inputs |

**`archive rerun` per computation:**

| Outcome | Printed | Exit |
|---|---|---|
| re-run equal | `reproduced` | 0 |
| `VARIANT`, re-run moved | `moved` (drift) | 0 |
| `REPRODUCIBLE`, re-run differs | `differs` (drift) | 1 |
| input 404, or fetched sha ≠ recorded | `refused` (url, recorded sha, fetched sha) | 1 |
| bundle absent, older epoch | `refused` (named) | 1 |

### Part 1c — `--show` on `compute` (epoch E3)

| Change | Detail |
|---|---|
| `compute --show table\|image\|text` [MODIFY] | exactly one file in `out/`, decoded as its kind: `table` CSV with a header row (a cell holding a newline refused); `image` PNG or SVG; `text` markdown under `edit`'s refusals and its voice advisory (`cli/blue/edit.go:104`). Zero files, two, or one that does not decode → refused. Joins the asserted set |
| Assembly [MODIFY] (`report/layout.go`, `report/computations.go`) | block after the anchor's paragraph, numbered per kind in document order: `TABLE` → markdown table, pipes escaped, `*Table n.* R`; `IMAGE` → `![Figure n. R](computations/<sha>/acts/<K-id>/out/<file>)` in report.md (run root, so the link resolves in the run and in an extracted archive), `<img src="data:image/…;base64,…">` in report.html (an `<img>` runs no SVG script); `TEXT` → verbatim |
| Image rule [MODIFY] (`md.go:437`) | assembly hands the renderer the store paths it placed; an image node embeds iff its destination is one of them, else renders as a link — in authored text, mint/fix, inquiry and ruling text alike. `edit` and `ingest` refuse an image in authored text (no archive holds one — measured 0) |
| Seat-facing views [MODIFY] | `show report --assembled --json` carries each block in its section's `text`; bare `show report` (`report/seatread.go`) carries the anchor only, and `AssembledNote` (`layout.go:284`) says shown blocks are assembled content |
| `show evidence --anchor K-…` [MODIFY] | returns the act's shown output — the stored path for an image; `reproduce` help: read the image, not the caption — a truncated axis is an unsound sentence |
| Voice lens [MODIFY] `scripts/agentgen/src/red-lens-voice.md` | a shown text block is report text: read the report as assembled (R5) |
| evidence.md, help [MODIFY] | each entry adds its shown output; tables are CSV, never printed markdown; a diagram can be a mermaid fence in `text` |
| migrate | replay only: migrated computations are not shown |

### Part 2 — `--show` on `cite` (epoch E4)

| Change | Detail |
|---|---|
| `cite --show text` | the quoted span as a block quote; verification unchanged |
| `cite --show image --page N --region x0,y0,x1,y1` | PDF: `pagerender` at its fixed resolution, cropped to the region in its pixels; an image source: the whole image. Stored as a derived cache entry keyed (source sha, page, region, renderer) — in `cache/`, never committed (R-commit); `archive rerun` re-derives it from the re-fetched source |
| Absent crop | report.md, report.html, `show report --assembled` and `devcmd/archdump` render "*Figure n.* R — <source title>, page N: image not held in this copy" (subject terms only — report-voice ruling); run.md's log names the absent crop (C-…, region) and `archive rerun`, which re-derives it. Never a silent omission |
| Evidence lens `verify` | reads the crop beside `render-page`'s render of the same page — fidelity at the leaf, no seat program between |
| First test | `fetchcache` has no `image/*` branch (`grep 'image/' internal/fetchcache internal/cli/blue`: imports only); whether `cite` accepts an image source is measured before the verb is built |
| migrate | replay only: migrated citations are not shown |

### III.9 Consumer census

Pattern `P` = `(?i:proof|\bprov(?:e|ed|es|ing)\b|reproduc|not_?portable)|Prove(?=[A-Z(])|[a-z]Prove\b`
(PCRE; `provenance`, `improve`, `proven` do not match). Run at `d9b2c823` from the repository root;
every hit's meaning read and recorded in the ledgers.

| # | Command | Hits | Consumers / false | Parts |
|---|---|---|---|---|
| C1 | `grep -rlP "$P" --include=*.go plugins/frank-exchange-of-views/tools \| grep -v _test.go` | 117 files | 73 / 44 | 72 in 1a, of them also 1b 8, 1c 13, 2 2; `cli/lens/verify.go` (an English hit) changes in 2 only |
| C2 | same, `--include=*_test.go` | 195 files | 139 / 56 | 1a; `keycensus_test.go` `notFields`; new tests per part (§V) |
| C3 | `git ls-files \| grep -v '\.go$' \| grep -vE '^(run-archive\|plans\|research\|ideas\|projects)/' \| xargs grep -cP "$P"` | 93 files | 35 regenerated (11 agents, 2 preloaded skills, 7 difftest + 3 migrate goldens, `schema.sql`, 10 simulator, `vocabulary.md`); 33 authored, per line in `round2/c3-nongo.tsv`; `site.html.tmpl:33` and terms.json's six lines change; `payload-keys.txt` frozen; 22 English | 1a; 1b/1c/2 where marked |
| C4 | not `P`: `red-lens-voice.md` (1c duty), `fetchcache/shell.go:223–226` (1b rights), `help/reproduce.md:9`, `run-directory.md:28` (1a) | 4 | all | as marked |
| C5 | archives: `ls */proofs/`, `bases.py`, `grepdb.py`, `grepdb2.py`, shared-sha and sha-prose scans | 157 scripts; 239 events; III.1.6 residue counts | — | 1a |
| C6 | `grep -rn 'ast\.Image\|"!\[\|<img' --include=*.go tools \| grep -v _test` | `md.go:437` (+ a comment, `pdfextract.go:28`) | 1 | 1c |
| C7 | `grep -nP '(?i:\bprove\b\|reproduce\|proof)' releasegate/fuzz/fuzz_test.go` | prove :1757 :2436 :2455 :2461 :2981 :6604; reproduce :1817 :3463 :3480 :6465; store :4577; bases :5390–:5391; `dirProveDrifts` :2656 (now passes `--variance`); "unjudged proof " :1919 :1942 | all | 1a; an `--input` action 1b; a `--show` action 1c |
| C8 | `grep -rnP '"P-"\|P-…\|\bP-[0-9a-f]{4,8}\b\|\[\^P\|\bP-\b' --include=*.go tools` | 27 lines / 13 files non-test; 215 / 46 test | all | 1a (`anchor.Token("P-…")` in `flags/shapes.go:155,162,194`, `cli/seat/verbs.go:286,293,299`, `cli/lens/close.go:86`, `cli/lens/reproduce.go:144`, `record/refs.go:118`; `P-…` prose at `verbs.go:656,671`; `[^P` at `layout.go:580,591`, `report/notes.go:10–20`, `report/proofs.go:40,45`) |
| C9 | names across Go, md, js: `compute`, `computations`, `sources`, `rerun`, `archive`, `acts`, `<bundle>` | `"compute"` only as a refused `--check-kind` spelling (`cli/checkkind_test.go:146`); `sources` = cited sources, the same sense; `rerun` = the `Rerun` enum and capture's audit, the same concept; `archive` — the run-archive tarball (the group's and the package's sense), the web-archive backend (`fetchcache/backends.go:41`) and the chair's closed-gap archive (`record/spotcheck.go:43`), neither seat-facing sense on an operator page; `acts` = recorded acts, the same sense; `<bundle>` 0 hits | — | 1a–1b |
| C10 | tar readers: `grep -rn 'tar\.NewReader' --include=*.go tools \| grep -v _test` | 2 (`recordtest/archive.go:66`, `archdump/main.go:159`) | both move onto `archive.Extract` | 1b |

Named consumers, meaning checked: `capture.go:2006` (archive contents — changes, 1a); `capture.go:2007`
(no `cache/` — R-commit keeps it true); `capture/proofrerun.go:155` (`Observed`/`NotPortable` → moved
— becomes `VARIANT`, 1a); `fetchcache/shell.go:223–226` (changes, 1b); `law/precedents.md:11–20` (a
past run's facts and a human-promoted holding — unchanged); `seat-surface-naming.md:806` (a past
run's measurement transcript — unchanged).

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation |
|---|---|---|---|
| R1 | The 1a rename (212 Go files) hides a logic change | M × H × L | Rename commit first, alone, gates green; logic commits after (V1a.1) |
| R2 | Plotting libraries absent in a seat's environment | H × M × L | Help names mermaid-in-`text` and stdlib SVG; a missing module is a refused run, loud |
| R3 | Network reach is declared, not enforced | M × M × H | Two runs catch most moving output; the rest is the lens's soundness duty; #1560 |
| R4 | Image bytes flap across machines (fonts, float, salted ids) | M × M × M | the pins and `MATPLOTLIBRC` (III.1.2); toolchain sha explains a re-run's difference; `VARIANT` exists |
| R5 | Inlined `text` carries process voice | M × M × L | the voice lens's duty to read the assembled report (1c). **Residue:** `edit`'s voice check is advisory (`edit.go:104`, `:171`) and refuses nothing; the lens's read is judgement, not a gate |
| R6 | A caption overclaims its chart | M × M × M | Caption = `--reason`, judged against the image (1c help) |
| R7 | **Filesystem residue (R-fs)** — table below | M × M × H | Stated; #1560 |
| R8 | **Cited bytes leave the cache (R-commit, R-rights)** — table below | M × M × L | Source bodies never enter `run-archive/`; every other channel is accepted residue, stated in 1b's comment and `compute --help` |
| R9 | **Link rot / moving sources** | M × M × M | `archive rerun` refuses loudly per input and exits 1 — never a silent pass |
| R10 | `archive rerun` breaks the fetch politeness contract | L × M × L | `fetchcache.Resolve`'s pacing; `--citations` is opt-in (N2) |
| R11 | Unbounded output and `out/` (R-size) bloat `run-archive/` | M × L × L | Stated in `compute --help`; the bundle stays ≤ 1 MiB |

**R7 — what the two-path copy catches:**

| A program that … | Run 1 vs run 2 | Graded |
|---|---|---|
| reads a bundle file, or prints its own path | same (`<bundle>`) | correct |
| reads `../x`, escaping the copy | different depths → different targets | moved → refused |
| reads the run tree by relative path and fails on it | missing in both | environment error or exit → refused |
| reads a missing path and tolerates it | same | **`REPRODUCIBLE`, vacuous — residue**: the computation lens judges the bundle and output |
| reads `sources/C-…` (1b) | materialised in both | correct; verified from the archive by `archive rerun` |
| reads the run tree by absolute path | the same file in both | **`REPRODUCIBLE` — residue**: re-runs equal on the run's own machine; elsewhere `archive rerun` exits 1. 27 of 157 archived scripts name an absolute path, 11 read `cache/` |
| reads the network | may agree | residue (R3) |

**R8 — channels for a cited source's bytes:**

| Channel | Committed | Graded |
|---|---|---|
| the source body (`cache/`, `sources/C-…`, crops) | never | R-commit |
| a bundle file copied from a source (the only REPRODUCIBLE route in 1a) | yes, ≤ 1 MiB | accepted residue, as `proofs/` today |
| output and `out/` computed over a source (1b), TDM-reserved or not | yes, unbounded | accepted residue (R-rights): may carry more of a source than a quotation |

## V. Verification Plan

Before any FEOV test: `unset $(env | grep -o '^FEOV_[A-Z_]*')`. Every part ends with
`go -C scripts run ./check` (fail count read), `go -C scripts run ./agentgen -check`,
`go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check`, and
`FEOV_RELEASE_GATE=1 go -C plugins/frank-exchange-of-views/tools test ./releasegate/fuzz/` (record
and flag change). Re-armed by: any edit under `plugins/frank-exchange-of-views/tools`,
`scripts/agentgen/src`, debate.js. Compute tests drive a fixture program the test builds (`.exe` on
Windows) and carry no Short gear, so they run on the Windows leg.

| # | Check | Fails if |
|---|---|---|
| V1a.1 | Rename commit alone: `protogen`, `schemagen`, `vocabdoc`, the shared gates; read every changed golden line and each `--help` before/after | a gate is red, or a help page lost a duty |
| V1a.2 | `internal/compute`: agreeing runs → `REPRODUCIBLE`; undeclared drift (output, exit, `out/` each) → refused; declared → `VARIANT` + drift; `--variance` over agreeing runs → refused; `../` escape → refused; relative run-tree read → refused; tool call → refused; symlink, > 1 MiB, `out`/`sources` entry → refused; `./run.sh` with its executable bit → runs, toolchain absent; without it → refused; no copy remains after a run, read-only `sources/` included | any arm passes the wrong way |
| V1a.3 | P-norm: a Python program printing `__file__` and raising → `REPRODUCIBLE`, recorded output holds `<bundle>` and neither copy's path; "No such file or directory" on stderr only at exit 0 → refused. Delete the replacement → the first fails | it is recorded otherwise |
| V1a.4 | Per act: two acts over one `VARIANT` bundle with different outputs → two `acts/`, the first's output and `out/` byte-identical after the second; `reproduce` of the first compares its own. A pre-existing `acts/<K-id>/` → refused. Delete the write-once guard → a test fails | an act changes after write |
| V1a.5 | `reproduce`: an `out/` file changed with output equal → `RERUN_DIFFERS`, drift names the file; exit changed → `RERUN_DIFFERS`; absent recorded `out/` with an empty re-run `out/` → `RERUN_REPRODUCED` | wrong verdict |
| V1a.6 | Keyed retry: different `--variance` → refused; different argv / file bytes → Conflict; identical → idempotent, nothing written. Record write: `VARIANT` without drift or invariant, `REPRODUCIBLE` with either, `show` = `SHOW_UNSPECIFIED` → refused by the `(check)` | any is recorded |
| V1a.7 | migrate tests on the archives of III.1.6, each assertion and residue row; then a test calling `compute.Rerun` on every migrated computation of all 29 (run by hand before E1 merges and on the release gate; a missing interpreter fails it) | a refusal, a failed assertion, or any of the 107 old-store matches (`audit-r1/probe-results.tsv`) not reproducing |
| V1a.8 | Delete-the-row: each pinned env var, each `remap` admission (`P-` in `hexID`, `archivedToken`, `[^Pn]`), each refusal arm, each `notFields` entry — a test fails for each | a deletion leaves the suite green |
| V1a.9 | Smoke: `scripts/universe.sh build`, grep the staged marketplace for `blue compute`, `--smoke` on a numeric question; capture's `computation-rerun` | the run recorded no computation, or a FAIL |
| V1b.1 | `--input`: raw bytes at `sources/C-…`, read-only, in both copies; a citation not on the record / not cached → refused; a different input under the same key → Conflict; a zero-input bundle hashes to the sha pinned in 1a's test | any |
| V1b.2 | `archive rerun` on a fixture archive with no `cache/` against a local server: correct bytes → `reproduced`, exit 0; a `VARIANT` that moved → `moved`, exit 0; a `REPRODUCIBLE` that differs → exit 1; changed bytes → refused naming url and both shas, exit 1; 404 → exit 1. `archive.Write` → `archive.Extract` round-trips; C10 finds `tar.NewReader` only in `internal/archive` | any outcome or exit differs |
| V1b.3 | Real data: V1a.9 on a question with a CSV source; extract its archive to a clean scratch dir, `feov-record migrate`, `feov-record archive rerun` | fewer than one computation with an input re-fetched, verified and re-run; or a non-zero exit |
| V1c.1 | Assembly goldens for each `Show`; report.html carries `data:image/` and no external `src`; delete each `Show` case → a test fails | a block missing, misnumbered, or external |
| V1c.2 | `edit` and `ingest` adding `![x](y)` → refused; `![x](y)` in a mint, fix, inquiry and ruling text → a link in report.html, no `<img>`; all 29 archives still migrate with zero refusals | accepted, embedded, or a migration refusal |
| V1c.3 | `show report --assembled --json` holds the table after the anchor's paragraph; bare `show report` holds none; the `red-lens-voice` agent golden holds the assembled-read duty | either differs |
| V1c.4 | Smoke as V1b.3 on a question whose answer is a table and a figure; read report.md and report.html | no shown block |
| V1c.5 | Keyed retry with a different `--show` → refused; `out/` with 0 or 2 files, a CSV cell holding a newline, an undecodable PNG → refused | any is recorded |
| V2.1 | `cite --show image` crop equals the same region of `render-page`'s raster; `verify` reads both | differs |
| V2.2 | Cache entry deleted: report.md, report.html, `show report --assembled` and archdump each hold the placeholder — in the subject's terms only: the figure's number, its source's title and page, and that the image is not held in this copy; no command, path or run term in the report (report-voice ruling) — and run.md's log names the absent crop and `archive rerun`, which re-derives it; `run-archive/` holds no crop (`tar tzf`). Delete the placeholder branch → a test fails | silent omission, or a committed crop |
| V2.3 | Smoke: a question with a PDF source carrying a figure; read report.html | no shown crop |
| V* | `/prosthetic-conscience:plan-audit plans/compute-and-show.md` | FAIL |

## Forks

| # | Question | Ruling | Alternative |
|---|---|---|---|
| F1 | Anchor letter and footnote label | `K-` / `[^Kn]`, remapped in migrate | ~~Keep `P-` / `[^Pn]`~~ |
| F2 | Who compares a variant's outputs | the computation lens, against the invariant; a disagreement reaches the bench by the existing contest | Auto-docket every variant re-run |
| F3 | Archive the cached inputs | superseded by R-commit | — |
| F4 | Migrated `not_portable` / `observed` | `VARIANT`, drift as invariant, in the manifest | A retained basis no live writer produces |
| F5 | Verb name | `compute` | `run`; `exhibit` |
| R-split | One PR or several | 1a / 1b / 1c / 2, each its own epoch and translation | One PR at one epoch |
| R-show | Build `--show` | yes — the displayed content is the output | Typed markdown tables |
| R-fs | Filesystem sandbox | not now; residue stated (R7); #1560 | Sandbox in 1a |
| R-commit | Input bytes in `run-archive/` | none; the archive commits the citation; `archive rerun` re-fetches and verifies | Commit cited bytes |
| N1 | Home of the archive re-run | group `archive`, member `rerun <runDir>` on an extracted, migrated archive; `capture` stays top-level and writes through package `archive` | ~~`capture` moved into `archive`~~; ~~top-level `rerun`~~ |
| N2 | `archive rerun --citations` | ship in 1b, opt-in | Inputs only |
| P-norm | Absolute paths in output | each copy's exact root → `<bundle>` before comparing and recording, at deposit and every re-run | Record raw; such programs stay `VARIANT` |
| R-rights | Computing over cited bodies | stated and accepted, as for citations; 1b's comment records the decision; §IV R8 grades every channel | Consult `TDMReserved` |
| R-input | What `--input` carries | raw bytes only | Extracted text |
| R-size | Limit on output and `out/` | unbounded, stated | A cap |
| Per-act (lead, disclosed) | Store key for an act's artifacts | `dir/` by `bundle_sha`; output and `out/` under `acts/<K-id>/`, written once | One store per sha |
| Crop-absent (lead, disclosed) | Assembly when a shown crop is not cached | a loud placeholder in subject terms (figure number, source title, page, "image not held in this copy"); the operator pointer to `archive rerun` goes to run.md, never report.md | Omit the figure; ~~an operator command in report.md~~ (report-voice ruling) |
