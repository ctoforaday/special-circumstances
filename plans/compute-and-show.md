# Compute and show — what we computed, what a source shows, and how either is seen

> STATUS (2026-10-23): proposed, not started. Written against `380164a8` (`EventSchema` 37,
> 29 archived runs carrying 157 proofs). Forks F1–F5 ruled by gblock 2026-10-23.

## I. Summary & Goals

**Objective.** The report holds two kinds of evidence and can show neither.

- **What a source shows** — `cite` — backs a sentence with a footnote.
- **What we computed** — `prove` — backs a sentence with a footnote, and its output is a code fence
  in evidence.md.

Nothing a reader sees in the report body is a table, chart, image or quoted figure: a markdown image
renders as a link (`report/md.go:437`), and no code builds a table or image from evidence. And
`prove` carries three misfits:

1. **The name.** Of 157 archived "proofs", none is a proof in the sense a reader takes the word; the
   code already calls the `observed` basis "a MEASUREMENT, not a proof" (`proof/proof.go:53`). The
   record names the concept correctly elsewhere: `check_kind: computation`, the computation lens.
2. **One file, language by extension.** `proofs/<sha>/script.<ext>`, the extension picking an
   interpreter from a five-row table (`proof/interpreter.go:37`). A computation that needs a
   helper, a query file or data has nowhere to live; one that reads the run tree is graded
   `not_portable` (6 of 28 in `2026-09-02_quadratic-formula`).
3. **Inputs nobody can reach later.** A script that reads a cached source works in the run and fails
   from the archive, because the archive omits `cache/` (`capture/capture.go:2007`).

**The design is two mirrored verbs and one placement rule.**

| | `cite` — theirs | `compute` — ours (replaces `prove`) |
|---|---|---|
| Records | a source's bytes, by sha256 | a directory and a command, by sha256 |
| Store | `cache/` | `computations/` |
| Backs a sentence | `--quote` | `--quote` |
| Shown in the report | `--show image\|text` | `--show table\|image\|text` |
| Checked by | the evidence lens — **fidelity**: is this what the source shows? | the computation lens — **re-run and soundness** |
| Joined by | | `--input <citation>`: what we compute is built from what sources show |

**Why two verbs and not one.** "The source shows this" and "we computed this" are different claims,
checked by different tests, and a reader must be able to tell them apart. A program between a
source and its displayed image is a place for distortion that re-running cannot catch — a recolored
or rescaled crop reproduces byte for byte. So a source's image is extracted by the tool and checked
for fidelity; a chart we drew is re-run. The axis that matters is never "figure vs code"; it is
*theirs vs ours* × *backs a sentence vs shown*.

**The placement rule.** Everything backs a sentence; nothing floats. `--show` adds a display:
assembly places the content after the paragraph holding the anchor, numbered in document order
("Figure n" for an image, "Table n" for a table), captioned by the act's `--reason` — the same field
that is already its footnote. A shown block leaves the report the way every anchor does: the
sentence is edited, then `retire` takes the anchor out (`anchor/anchor.go:75`). No new lifecycle.

**Success criteria.**

1. `compute --dir D --quote S --reason R -- python3 main.py` records a computation whose stored
   bundle is every file under `D`, whose command is the argv after `--`, and whose recorded
   toolchain is the resolved path and sha256 of `python3`. `prove` does not exist.
2. A computation reading `inputs/C-…` re-runs from `run-archive/<slug>.tar.gz` alone, on a machine
   with no `cache/` (§V.5).
3. Two runs whose output differs are REFUSED at write unless `--variance` states the invariant that
   must hold; with it, the computation is recorded `variant`. The bases are exactly `reproducible`
   and `variant`, as a record enum.
4. `--show table` renders the program's CSV as a markdown table under "Table n. R"; `--show image`
   embeds its PNG or SVG under "Figure n. R" in report.md and report.html; `--show text` inlines its
   markdown, refused where the same text would be refused by `edit`.
5. An image in authored report text is refused at `edit`: the only images in a report are shown
   blocks, as the only numbered footnotes are the tool's.
6. All 29 archived runs migrate with zero refusals; every proof becomes a computation that re-runs
   from the migrated store to its recorded output where it did before (§V.4).
7. Part 2: `cite --show image --page N --region …` displays a tool-rendered crop the evidence lens
   verifies against `render-page`; `cite --show text` displays the quoted span as a block quote.

**Value case — what goes away.**

| Removed | Replaced by |
|---|---|
| The extension→interpreter table, its Windows extra and `argvFor` (`interpreter.go`) | the argv after `--`, resolved on PATH |
| The second grading run from the store (`executeStored`, `proof.go:202`) | two runs, each in a fresh copy of the bundle at a different path |
| Three bases as free strings (`reproducible`, `observed`, `not_portable`) | a two-value enum, `reproducible` / `variant` |
| Silent `observed` for output that moved | refusal, or a declared invariant — the tool never infers |
| Language recovered from a stored filename | the recorded command and toolchain fields |
| "Images render as links" as a blanket rule | images exist only as shown blocks, embedded from the run's store |

**Not in scope.** Rich media and interactive charts (no home in a static report.md; the site would
have to run scripts). A machine-checked basis (Lean): the argv already holds `lake env lean X.lean`;
what it lacks is a checker for "type-checks with no `sorry`", and the Mathlib environment. Output
signature scanning (`ErrorSignature`, `--expect-error`) is untouched — it belongs to #1466.

## II. Technical Context

- Go, FEOV tools module; record is SQLite generated from `record.proto` (`Proof` :1366,
  `Reproduce` :1421, `Rerun` :552, `Soundness` :538, `CheckKind.COMPUTATION` :343).
- `EventSchema` 37 → **E** = the next free epoch at implementation. One `migrate` translation.
- Anchors: one mechanism (`anchor/anchor.go:63`); `P-` today names a proof and becomes `K-`
  (F1; `K` is held by no other id kind). No new anchor kind: a shown block is the same anchor with
  a `show` field on its event.
- Execution posture is unchanged: a seat could already run anything through Bash; `compute` adds
  recording, not reach. There is no execution sandbox (`nonet` is test-only), so "no network" is
  declared, never enforced — §IV R3.
- Page rendering exists (`fetchcache/pagerender.go`, lens verb `render-page`); Part 2 reuses it.
- HTML site: `report/md.go` renders markdown tables (:191) and mermaid fences (`site.go:181`).

## III. Proposed Changes

Two parts, in ship order. Part 1 is one pull request at epoch E.

### Part 1 — `compute` replaces `prove`; `--show` on `compute`

#### III.1 The record [MODIFY] (contract: event schema)

```proto
message Computation {                // replaces Proof
  string computation_id = 1;         // K-…  (F1)
  string key            = 2;
  string bundle_sha     = 3;         // sha256 over argv, input shas and the bundle tree
  string answers        = 4;
  string cites          = 5;         // the METHOD's citation, unchanged
  string text           = 6;         // --reason: footnote, and caption when shown
  string location       = 7;         // --quote
  repeated string argv  = 8;         // the command after --
  repeated Toolchain toolchain = 9;  // argv[0] resolved: path + sha256 of the binary
  repeated string inputs = 10;       // citation ids, materialised at inputs/<id>
  Basis basis           = 11;        // REPRODUCIBLE | VARIANT
  string invariant      = 12;        // --variance; required iff VARIANT
  Show show             = 13;        // NONE | TABLE | IMAGE | TEXT
  int32 exit            = 14;
  string drift          = 15;        // what moved between the two runs, iff VARIANT
}
message Toolchain { string name = 1; string path = 2; string sha256 = 3; }
enum Basis { BASIS_UNSPECIFIED = 0; REPRODUCIBLE = 1; VARIANT = 2; }
enum Show  { SHOW_UNSPECIFIED = 0; NONE = 1; TABLE = 2; IMAGE = 3; TEXT = 4; }
```

`Reproduce` keeps its shape, renamed `proof_sha`→`bundle_sha`, `proof_id`→`computation_id`. The
`Citation` event gains `show` (Part 2 adds `region`). `awaiting_proof` → `awaiting_computation`.
Event type `PROOF` → `COMPUTATION`.

#### III.2 The verb [MODIFY → rename] (contract: CLI flag set)

```
feov-record compute --dir <bundle> --quote <sentence> --reason <note>
                    [--input C-…]… [--variance <invariant>] [--show table|image|text]
                    [--answers G-…] [--cites C-…] [--key k] [--expect-error]
                    -- <command> [args…]
```

- `--dir` replaces `--script`. A one-file computation is a one-file directory. Symlinks leaving it,
  and bundles over 1 MiB, are refused (data arrives by `--input`, not by copying it in).
- The argv after `--` is the command, run with the bundle as working directory. A missing argv[0]
  is refused naming it.
- `--input C-…` places the cited source's cached bytes read-only at `inputs/C-…`. A citation the
  cache does not hold is refused.
- `--show` names what the program writes into `out/`: exactly one file, decoded as its kind —
  `table` a CSV with a header row; `image` a PNG or SVG; `text` markdown passing the same checks
  `edit` applies to report text. Zero files, two files, or a file that does not decode is refused.
- Standard output stays the recorded output and evidence.md's fenced block, as now.

#### III.3 Execution [MODIFY/DELETE] — `internal/proof` → `internal/compute`

1. Hash the bundle, argv and input shas → `bundle_sha`; store at `computations/<sha>/` as `dir/`
   (the bundle), `out/`, `stdout`.
2. Copy the bundle twice, to two distinct paths; materialise inputs into each; run each under the
   checker environment (`FEOV_PROOF_RERUN` → `FEOV_COMPUTE_RERUN`), so a tool call is refused in
   both and the write is refused — no `not_portable` grade survives.
3. Pin the environment: `SOURCE_DATE_EPOCH=0` (byte-stable matplotlib SVG/PDF), `PYTHONHASHSEED=0`,
   `TZ=UTC`, `LC_ALL=C`, `PYTHONUTF8=1`.
4. Compare stdout, exit and `out/` bytes. Equal → `REPRODUCIBLE`. Different and no `--variance` →
   refused, naming what moved and the two remedies. Different with `--variance` → `VARIANT`, drift
   recorded.
5. Delete: `interpreter.go`'s `candidates`/`windowsExtra`/`argvFor`, `executeStored`, `NotPortable`,
   `Observed`, `StoredScript`'s extension spelling.

`reproduce` (lens) re-runs from `computations/<sha>/` the same way and records the rerun as now. On
`VARIANT`, a `sound` verdict over a differing rerun is permitted, and the lens's reason MUST say
whether its output satisfies the invariant (fork F2). On an `image` computation, `show evidence
--anchor` returns the stored image path and the lens help says to read the image, not the caption:
a truncated axis is an unsound sentence.

#### III.4 Assembly and rendering [MODIFY]

- `weaveProofRefs` → `weaveComputationRefs`: footnote `[^Kn]` (F1).
- A computation or citation with `show ≠ NONE` emits its block after the anchor's paragraph:
  - `TABLE`: CSV → markdown table, pipes escaped, then `*Table n.* R`.
  - `IMAGE`: `![Figure n. R](computations/<sha>/out/<file>)` in report.md; `<img>` in the site
    (an `<img>` runs no SVG script).
  - `TEXT`: the markdown verbatim.
- `md.go`: an image embeds iff its destination is in the run's store and assembly emitted it; every
  other image is impossible, because `edit` refuses one in authored text (success criterion 5).
- evidence.md `## Proofs` → `## Computations`: per entry, the argv, toolchain, inputs, basis,
  invariant, every bundle file fenced by its own extension, stdout fenced, and the shown output.

#### III.5 Gates and views [MODIFY]

`awaiting_computation` (`recordsql/views.go:547`), `ProofAnswers` → `ComputationAnswers`
(`estoppel.go:255`), `proofOwnerOf`/`proofChecksOwed` (`available.go:522–564`), retirement hold
(`retirement.go:154`), `BlockerReopened` (`passblockers.go:34`), `proofs_unverified`
(`evidenceview.go:314`) — renamed, logic unchanged. A shown block reopens with its sentence, by the
existing rule.

#### III.6 Archive [MODIFY]

`ArchiveRecord` (`capture/capture.go:2016`) archives `computations/` in place of `proofs/`, plus
`cache/<sha>` for every sha a computation names as input — only those. This is the one growth in
the archive, bounded by the fetch cap × inputs (fork F3).

#### III.7 migrate — one translation [NEW]

Per `Proof` event: argv from the extension (`.py`→`python3`, `.sh`→`bash`, `.js`/`.mjs`→`node`,
`.go`→`go run`), the script moved to `computations/<new sha>/dir/<script name>`, toolchain empty,
`observed` → `VARIANT` with the recorded drift as invariant, `not_portable` → `VARIANT` with its
drift (F4), `show` NONE. `proof_sha` references in `Reproduce` and the manifest are rewritten
to the new sha; every `P-` id becomes `K-` with the same hex, and every `<!--proof:P-…-->` token
`<!--computation:K-…-->`, in the report text (`migrate/remap.go`), and in every event field that names one (F1). Test on `2026-09-02_quadratic-formula` (26
proofs, the not_portable six) and `2026-09-11_is-91-prime-b6` (the `.go` proof).

#### III.8 Agent-facing text [MODIFY] (contract: prompts and constitutions)

Sources only, then `agentgen` and `enginegen`:
`scripts/agentgen/src/{blue-researcher,blue-synthesizer,red-lens-computation,red-lens-evidence}.md`,
`fragments/{blue-search-and-cite,lens-audit-and-mint}.md`, both preloaded skills, debate.js
(:229, :969, :991, :1134), `references/report_template.md:147`, `references/run-directory.md:77`,
terms.json (`proof` :586 → `computation`; then `vocabdoc`). Seat prompts name the act —
"compute it", "show it" — never the verb. The `compute --help` page teaches: a directory and a
command; inputs are citations; make output deterministic or state the invariant; tables are CSV,
never printed markdown; a diagram can be a mermaid fence in `text`, needing no plotting library.

#### III.9 Consumer census

Run at `380164a8`; output in `~/.claude/scratch/compute-plan/census.txt`.

| # | Command | Hits | Changes |
|---|---|---|---|
| C1 | `grep -rlE '\bproof\.\|Proof\|prove\|Reproduce\|awaiting_proof\|not_portable\|NotPortable' tools --include=*.go \| grep -v _test` | 84 files | every one (rename; logic in III.3–III.5) |
| C2 | same, `--include=*_test.go` | 150 files | rename; R/proof's 21 tests rewritten for bundles |
| C3 | `grep -rlE '\bproofs?\b\|\bprove\b\|reproduce' scripts/agentgen/src skills/*/references debate.js` | 11 files | III.8 |
| C4 | `grep -n '"proof' tools/internal/terms/terms.json` | 1 term | III.8 |
| C5 | archived runs with `proofs/` | 29 runs, 157 proofs | III.7 |
| C6 | `grep -rn 'ast.Image' tools/internal/report` | `md.go:437` | III.4 |
| C7 | `tools/releasegate/fuzz/fuzz_test.go` prove/reproduce sites | :533, :1757, :1817, :2436, :2981, :3463 | rename + a `--show` and an `--input` action |
| C8 | `grep -rn '"P-"' tools --include=*.go` (the prefix, F1) | run at implementation | `anchor.go:63` and every literal |

### Part 2 — `--show` on `cite`

- `cite --show text`: the quoted span, displayed as a block quote. Verification unchanged.
- `cite --show image`: for a PDF, `--page N --region x0,y0,x1,y1` (fractions of the page);
  the tool renders the page with `pagerender` at its fixed resolution and crops. For a source whose
  cached bytes are an image, the whole image, no region. The crop is stored as a derived cache
  entry keyed by (source sha, page, region, renderer) and archived with the run.
- The evidence lens's `verify` of a shown citation reads the crop beside `render-page`'s render of
  the same page — fidelity at the leaf, with no seat program between them.
- Census before Part 2 starts: how `cite` treats an `image/*` content type today (unmeasured).

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation |
|---|---|---|---|
| R1 | The rename (234 files) hides a logic change in a mechanical diff | M × H × L | Rename commit first, alone, gates green; logic commits after (§V.1) |
| R2 | Plotting libraries absent in a seat's environment | H × M × L | Help names mermaid-in-`text` and Go/stdlib SVG; a missing module is a refused run, loud |
| R3 | "Idempotent inputs" is declared, not enforced — a script can still reach the network | M × M × H | Two runs catch most moving output; the rest is the lens's soundness duty. Stated, not hidden |
| R4 | Byte comparison of images flaps across machines (fonts, float) | M × M × M | `SOURCE_DATE_EPOCH`; the toolchain record explains a re-run's difference; `variant` exists |
| R5 | Inlined `text` smuggles process voice or structure into the report | M × M × L | It passes `edit`'s checks; the voice lens reads assembled text |
| R6 | Archive growth from inputs | L × M × L | Only cited inputs of computations; F3 |
| R7 | A shown block's caption overclaims what the chart shows | M × M × M | Caption = `--reason`, judged for soundness against the image itself (III.3) |

## V. Verification Plan

1. Rename commit alone: `unset $(env | grep -o '^FEOV_[A-Z_]*'); go -C scripts run ./check` green;
   `go -C scripts run ./agentgen -check` and `enginegen` clean.
2. New tests (R/compute): two-path copy catches `__file__`-relative reads; undeclared drift refused;
   declared drift recorded `VARIANT`; tool call refused at write; `--input` materialised read-only;
   `out/` with 0 or 2 files refused per kind; bundle symlink escape refused.
3. Delete-the-row checks: remove each `Show` case from assembly and each pinned env var — a test
   fails for each.
4. migrate: all 29 archives migrate with zero refusals; on quadratic-formula every computation that
   reproduced from its store before reproduces now; b6's `.go` proof re-runs under `go run`.
5. **Real data:** `scripts/universe.sh build`, prove the staged marketplace carries `compute`
   (grep the staged plugin), then a cheap smoke on a question with a data source — blue cites a CSV,
   computes a table and a chart from it with `--show`; red reproduces both. Then extract the
   archive on a clean directory with no `cache/` and `reproduce` from it.
6. Render: read report.md and report.html for the smoke — the table, the figure, their numbering,
   and an `edit` adding `![x](y)` refused.
7. `FEOV_RELEASE_GATE=1 go test ./releasegate/fuzz/` (record and flag change).
8. Auditor gate: `/prosthetic-conscience:plan-audit plans/compute-and-show.md`.

## Forks

| # | Question | Recommendation | Alternative |
|---|---|---|---|
| F1 | Anchor letter and footnote label for computations | **Ruled: `K-` / `[^Kn]`**, remapped in migrate | ~~Keep `P-` / `[^Pn]`~~ |
| F2 | Who compares a variant computation's outputs | **Ruled:** the computation lens, against the declared invariant — today's `observed` path, no new machinery; a disagreement reaches the bench by the existing contest | Auto-docket every variant re-run for the bench: a new docket basis, dispositions, and a mid-debate bench dispatch |
| F3 | Archive the cached inputs of computations | **Ruled:** yes, only those | No: archived computations with inputs become un-rerunnable |
| F4 | Migrated `not_portable` proofs | **Ruled:** `VARIANT`, drift as invariant, listed in the migrate manifest | A retained `NOT_PORTABLE` basis the live writer can never produce |
| F5 | Verb name | **Ruled:** `compute` — the record's own word (`check_kind: computation`, the computation lens) | `run`; `exhibit` (free of collisions, but a third noun for an existing concept) |
