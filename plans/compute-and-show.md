# Compute and show — what we computed, what a source shows, and how either is seen

> STATUS (2026-10-24): proposed, not started; revised after audit round 1. Written against
> `62fe337e` (`EventSchema` 37; 29 archived runs). Forks F1–F5 and rulings R-split, R-show, R-fs,
> R-commit, N1–N2 by gblock. Census ledger: `~/.claude/scratch/compute-plan/round2/census.tsv`.

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
   recorded `reproducible`, and 6 do not reproduce from the store (`proof.go:62`).
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
| Shown in the report | `--show image\|text` (Part 2) | `--show table\|image\|text` (1c) |
| Checked by | the evidence lens — **fidelity**: is this what the source shows? | the computation lens — **re-run and soundness** |
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

| Part | Scope | Epoch |
|---|---|---|
| 1a | `prove` → `compute`; directory + argv; two-value basis; `K-` anchors | E1 |
| 1b | `--input` (cited bytes, read-only); operator `archive rerun` re-fetches and verifies inputs | E2 |
| 1c | `--show table\|image\|text` on `compute`; images exist only as shown blocks | E3 |
| 2 | `--show image\|text` on `cite` | E4 |

E1–E4 are the next free epochs when each part lands.

**Success criteria.**

| # | Part | Criterion | §V |
|---|---|---|---|
| S1 | 1a | `compute --dir D --quote S --reason R -- python3 main.py` records a computation: bundle = every regular file under `D`; command = the argv after `--`; toolchain = resolved path + sha256 of `python3`. `feov-record prove` is an unknown command | V1a.1 |
| S2 | 1a | Two runs, in fresh copies outside the run tree. Equal output, exit and `out/` → `REPRODUCIBLE`; different with no `--variance` → refused naming what moved; with `--variance` → `VARIANT` + drift; `--variance` over agreeing runs → refused | V1a.2 |
| S3 | 1a | Recorded output is everything the program writes (stdout and stderr, one stream); an environment error on stderr alone at exit 0 is refused | V1a.3 |
| S4 | 1a | A keyed retry restates every recorded flag or is refused; program, argv and inputs are in the sha | V1a.4 |
| S5 | 1a | `reproduce` compares output, exit and `out/`; `RERUN_REPRODUCED` only when all three are equal | V1a.5 |
| S6 | 1a | All 29 archives migrate with zero refusals; per-archive assertions of III.1.6 hold; 0 computations that re-ran from the old store fail from the migrated one | V1a.6 |
| S7 | 1b | `--input C-…` materialises the cached bytes read-only at `sources/C-…`; the archive commits no source bytes; `archive rerun` on an extracted archive with no `cache/` re-fetches, verifies sha256, and re-runs — refusing loudly on mismatch | V1b.1–3 |
| S8 | 1c | `--show table` renders the CSV under "*Table n.* R"; `--show image` embeds PNG/SVG under "*Figure n.* R" in report.md and (as `data:` URI) report.html; `--show text` inlines markdown refused where `edit` would refuse it | V1c.1 |
| S9 | 1c | An image in authored text is refused at `edit` and `ingest` | V1c.2 |
| S10 | 1c | Shown blocks appear in `show report --assembled --json`; the bare read carries the anchor only and says so | V1c.3 |
| S11 | 2 | `cite --show image --page N --region …` displays a tool-rendered crop the evidence lens verifies against `render-page`; `cite --show text` displays the quoted span as a block quote; no crop is committed | V2.1–2 |

**Value case — what goes away.**

| Removed | Replaced by |
|---|---|
| The extension→interpreter table, its Windows extra and `argvFor` (`interpreter.go`) | the argv after `--`, resolved on PATH |
| The grading run from the store (`executeStored`, `proof.go:202`) | two runs, each in a fresh copy at a different path |
| Three bases as free strings (`proof.go:72–74`) | a two-value enum |
| Silent `observed` for output that moved (`proof.go:156`) | refusal, or a declared invariant — the tool never infers |
| Working directory = the run tree (`proof.go:256`) | working directory = a copy of the bundle |
| Typed tables that drift from their computation | the program's own output, placed by the tool |
| "Images render as links" | images exist only as shown blocks, embedded from the run's store |

**Not in scope.** Rich media and interactive charts. A machine-checked basis (Lean): the argv holds
`lake env lean X.lean`; a checker for "no `sorry`" and Mathlib are missing. A filesystem or network
sandbox (R-fs; #1560). Output-signature scanning (`ErrorSignature`, `--expect-error`) keeps its
behaviour (#1466).

## II. Technical Context

- Go, FEOV tools module. Record: SQLite generated from `record/recordpb/record.proto` (`Proof`
  :1366, `Reproduce` :1421, `Soundness` :538, `Rerun` :552, `CHECK_KIND_COMPUTATION` :343,
  `EVENT_TYPE_PROOF` :278, event oneof :718). Epoch: `requirements.json` → `scripts/schemagen` →
  `recordpb/schema_gen.go:18`.
- Anchors: one table (`anchor/anchor.go:63`). `P-` becomes `K-` (F1); `K` is held by no id kind.
- `migrate`: channel copy (`record/migrate/migrate.go:170`), id respelling (`remap.go`: `currentID`
  :84, `hexID` :86, `archivedToken` :90, `id()` :217, field lists :109–140), era mapping
  (`eraentries.go:54`), plan-id translation (`planids.go`).
- Execution posture is unchanged: a seat could already run anything through Bash; `compute` adds
  recording, not reach. No sandbox (`nonet` is test-only).
- Fetch cache: `fetchcache.Read` :374 by sha, `fetchcache.Resolve` :518 by URL; `cite` events carry
  `url` and `sha256`. Page rendering: `fetchcache/pagerender.go`, lens verb `render-page`.
- Site: `report.html` is one self-contained file (`report/site.go:56`); markdown tables render
  (`md.go:188`); mermaid is the one external fetch (`site.go:180`).

## III. Proposed Changes

### III.0 The record (contract: event schema) — one sketch, fields tagged by part

```proto
// Computation is a program blue ran twice on a bundle, and what it wrote. EVENT_TYPE_COMPUTATION = 24.
message Computation {                                   // [MODIFY] replaces Proof (oneof :718, = 42)
  optional string computation_id = 1;                   // K-… (F1)
  optional string key            = 2 [(sql) = { flag: "key" }];
  optional string bundle_sha     = 3;                   // III.1.3
  optional string answers        = 4;
  optional string cites          = 5;                   // the METHOD's citation
  optional string drift          = 6;                   // what moved between the two runs; iff VARIANT
  optional string text           = 7 [(prose) = true, (sql) = { flag: "reason" }];
  optional int32  exit           = 10;
  optional string location       = 11 [(prose) = false, (sql) = { flag: "quote" }];
  repeated string argv           = 12;                  // the command after --
  repeated Toolchain toolchain   = 13;                  // argv[0], resolved
  optional Basis  basis          = 14;
  optional string invariant      = 15 [(prose) = true, (sql) = { flag: "variance" }];  // iff VARIANT
  repeated string inputs         = 16 [(sql) = { flag: "input" }];                     // 1b
  optional Show   show           = 17 [(sql) = { flag: "show" }];                      // 1c; absent = not shown
  reserved 8, 9;
}
message Toolchain { optional string name = 1; optional string path = 2; optional string sha256 = 3; }
enum Basis {
  BASIS_UNSPECIFIED  = 0;
  BASIS_REPRODUCIBLE = 1 [(means) = "both runs wrote the same output, exit and out/ files, byte for byte"];
  BASIS_VARIANT      = 2 [(means) = "the two runs differed, and blue stated what holds across them"];
}
enum Show {                                             // 1c (compute), 2 (cite: IMAGE, TEXT)
  SHOW_UNSPECIFIED = 0;
  SHOW_TABLE = 1 [(means) = "the program's CSV, rendered as a numbered table after the anchor's paragraph"];
  SHOW_IMAGE = 2 [(means) = "an image, embedded as a numbered figure after the anchor's paragraph"];
  SHOW_TEXT  = 3 [(means) = "markdown, placed after the anchor's paragraph"];
}
// Reproduce [MODIFY]: proof_sha = 1 → bundle_sha = 1; proof_id = 7 → computation_id = 7;
//   + optional string drift = 9;  // what moved between the recorded run and this re-run (tool-written)
// Rerun (means) [MODIFY]: REPRODUCED = "re-ran the stored bundle; output, exit and out/ equal the recorded ones";
//   DIFFERS = "… output, exit or out/ differ; drift names what moved". CALLS_THE_TOOL, NOT_RUN: "bundle" for "script".
// Cite [MODIFY, Part 2]: + optional Show show = N; optional int32 page = N+1; repeated double region = N+2;
```

`awaiting_proof` → `awaiting_computation`. Every `(means)` naming a proof (:280, :343, :372–375,
:403, :448, :451, :454, :575, :1125) names a computation. remap's field lists class every new field:
`Computation.inputs` id, `invariant`/`drift`/`Reproduce.drift` argument, `argv`/`Toolchain.*` data.

### Part 1a — `compute` replaces `prove` (epoch E1)

#### III.1.1 The verb [MODIFY → rename] (contract: CLI flag set)

```
feov-record compute --dir <bundle> --quote <sentence> --reason <note>
                    [--variance <invariant>] [--answers G-…] [--cites C-…] [--key k] [--expect-error]
                    -- <command> [args…]
```

| Rule | Refused when |
|---|---|
| `--dir` replaces `--script`; a one-file computation is a one-file directory | a symlink, a non-regular file, a bundle over 1 MiB, or an entry named `out` or `sources` (the tool's mounts) |
| The argv after `--` is the command, working directory = the copy | argv empty, or argv[0] not on PATH (named) |
| `--variance` states what holds when the runs differ | given and the runs agree (precedent: `--expect-error` with no failure, `cli/blue/prove.go:128`) |
| Recorded flags restate on a keyed retry (`seat.RepeatsKeyed`, `cli/seat/keyedretry.go:31`) | asserted set (`prove.go:191`) + `--variance` → `invariant` differs; a different sha under the key → Conflict (`prove.go:83`) |

#### III.1.2 Execution [MODIFY/DELETE] — `internal/proof` → `internal/compute`

1. Copy the bundle twice, to two fresh directories at different depths, neither inside the run tree;
   each gets an empty `out/`.
2. Run each with the copy as working directory, under the checker environment
   (`FEOV_PROOF_RERUN` → `FEOV_COMPUTE_RERUN`, `seatenv.go:87`) and the pins:
   `SOURCE_DATE_EPOCH=0`, `PYTHONHASHSEED=0`, `TZ=UTC`, `LC_ALL=C`, `PYTHONUTF8=1`, and
   `MATPLOTLIBRC=<a tool-written file outside both copies holding svg.hashsalt>` (matplotlib salts SVG
   ids with uuid4 unless `svg.hashsalt` is set; `SOURCE_DATE_EPOCH` alone fixes only dates).
3. Record everything the program writes — stdout and stderr as one stream, as `CombinedOutput` does
   today (`proof.go:258`) — and `ErrorSignature` scans all of it.
4. Refuse at write: a tool call in either run (today graded `not_portable`, `proof.go:173`); an
   environment error without `--expect-error`; differing output, exit or `out/` without `--variance`.
5. Store `computations/<sha>/` = `dir/` (the bundle), `output`, `out/`.
6. Delete: `interpreter.go`, `executeStored`, `NotPortable`, `Observed`, `StoredScript`.

#### III.1.3 Identity

`bundle_sha` = sha256 over a sorted canonical listing: the argv; each input's citation id and sha256
(1b); each file's relative path, executable bit and sha256. Computed without running, so a keyed
retry tells a crash-retry from a collision before executing (as `ScriptSha` does today).

#### III.1.4 Re-run [MODIFY]

`reproduce` and capture's audit (`capture/proofrerun.go`, check `proof-rerun` → `computation-rerun`)
re-run from `computations/<sha>/` by III.1.2 steps 1–3 and compare output, exit and `out/` bytes.
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
| `ArchiveRecord` `proofs/` (`capture.go:2066`, comment :2006) | `computations/` |

#### III.1.6 migrate — one translation [NEW] (contract: record migration)

| Old | New |
|---|---|
| `Proof` / era `proof` event (`eraentries.go:54`) | `Computation` |
| `proofs/<sha>/script.<ext>` + `output` (sha dirs only) | `computations/<new>/dir/script.<ext>` + `output`; `<new>` from the old store alone (argv + tree) |
| extension | argv: `.py` `python3`, `.sh` `bash`, `.js`/`.mjs` `node`, `.go` `go run` — each + `script.<ext>`; toolchain empty |
| `proof_basis` `reproducible` / `observed` / `not_portable` | `REPRODUCIBLE` / `VARIANT`, invariant = drift (F4) |
| `proof_sha` in `Reproduce` and the manifest | `bundle_sha` through the old→new map |
| `P-`/`p-` + 8 hex | `K-` + the same hex: `hexID` (:86) and `archivedToken` (:90) admit `P-` — today `currentID` (:84) matches `P-` and `id()` (:217) would otherwise hash a new hex |
| `<!--proof:…-->` | `<!--computation:K-…-->` |
| `[^Pn]` in argument fields | `[^Kn]` |

| Archive (measured; `round2/c9-archive-scan.txt`) | Holds | Asserted after migrate |
|---|---|---|
| all 29 | 157 stored scripts | zero refusals; no `\bP-[0-9a-f]{8}\b` outside `Plan.printed`, no `<!--proof:`, no `[^P`; every `K-` id's hex equals its source id's |
| 2026-09-09_is-91-prime-b | `not_portable` `p-808ae840` | `VARIANT`, invariant = drift "runs from the proof store and says something else — exit code moved 0 -> 1 between runs" |
| 2026-08-22_is-7-prime-run2 (era) | `observed` `p-568517ec` | `VARIANT`, invariant = drift "output differs from byte 131 between runs" |
| 2026-10-12_is-91-prime-wave-b | `[^P5]` in a finding | `[^K5]` |
| 2026-10-10_is-91-prime-m19 | 125 `P-` ids | `K-`, same hex |
| 2026-09-11_is-91-prime-b6 | the `.go` script | argv `go run script.go` |

**Residue, accepted:**

| Residue | Size | Why it stays |
|---|---|---|
| `P-` ids and "unjudged proof" in `Plan.printed` | 34 ids, 3 runs (m28, m29, m30) | printed plans keep the wording they were printed in (`planids.go:15–17`) |
| `proofs/synth/` | 17 drafts, 1 run (record-store-authority) | each is also stored under its sha |
| `meta.json` in sha dirs | 37 files, 6 era runs | its fields are the era event's |
| author script path (`Proof.script`) | every proof | the argv replaces it |
| `Reproduce.drift` | migrated re-runs | not recorded then |

The lens help says a migrated `VARIANT`'s invariant describes what moved, not what must hold.

#### III.1.7 Agent-facing text [MODIFY] (contract: prompts and constitutions)

Sources only, then `agentgen`, `enginegen`, `vocabdoc`: `help/prove.md` → `help/compute.md`;
`help/reproduce.md` (`observed`/`not_portable` :21); `help/{dispatch,close,mint,edit}.md`;
`scripts/agentgen/src/{blue-researcher,blue-synthesizer,red-lens-computation,red-lens-evidence}.md`,
`fragments/{blue-search-and-cite,lens-audit-and-mint}.md`, both preloaded-skill sources;
`skills/research/SKILL.md:7`; debate.js :229 :538 :923 :969 :991 :1134;
`references/report_template.md:147`, `references/run-directory.md:77`; terms.json :574 :586 :965;
FEOV docs (ledger §C3); `docs/local-development.md:226`; `CLAUDE.md:34`;
`scripts/universe-stream.py:24`. Seat prompts name the act — "compute it" — never the verb. The
`compute --help` page teaches: a directory and a command; make output deterministic or state the
invariant.

### Part 1b — `--input` (epoch E2)

| Change | Detail |
|---|---|
| `compute --input C-…` (repeatable) [MODIFY] | the citation's cached bytes (`fetchcache.Read`) placed read-only at `sources/C-…` in each copy; id + sha256 enter `bundle_sha`. Refused: a citation not on the record, or bytes the cache does not hold |
| `reproduce`, capture's audit [MODIFY] | materialise `sources/` the same way |
| Archive [MODIFY] | `ArchiveRecord` moves from `capture` into a package `archive` that owns the tarball format; `capture` calls it and still writes the tarball. It commits no `cache/` (`capture.go:2007`, R-commit): the record holds each citation's URL and sha256, never its bytes |
| `feov-record archive rerun <runDir>` [NEW] (N1) | operator command group `archive` — acts on an archived run; `rerun` its one member. Input: an extracted archive brought to the current epoch by `migrate` (an older epoch is refused by `requireDeclaredSchema`, pointing to `migrate`): re-fetch each input citation by its `url` (`fetchcache.Resolve`, paced), verify sha256 against the record — refuse loudly naming url, recorded sha and fetched sha — then re-run every computation (III.1.4) and print reproduced / differs / moved per computation; exit non-zero on any refusal or difference. Writes no event. `--citations` verifies every citation the same way (N2) |
| Help, prompts [MODIFY] | `compute --help`: inputs are citations, read at `sources/<id>`; blue-researcher source: compute from what you cited |
| migrate | replay only: migrated computations hold no inputs |

### Part 1c — `--show` on `compute` (epoch E3)

| Change | Detail |
|---|---|
| `compute --show table\|image\|text` [MODIFY] | exactly one file in `out/`, decoded as its kind: `table` CSV with a header row (a cell holding a newline refused); `image` PNG or SVG; `text` markdown passing `edit`'s checks. Zero files, two, or one that does not decode → refused. Joins the asserted set (`show`) |
| Assembly [MODIFY] (`report/layout.go`, `report/computations.go`) | block after the anchor's paragraph, numbered per kind in document order: `TABLE` → markdown table, pipes escaped, `*Table n.* R`; `IMAGE` → `![Figure n. R](computations/<sha>/out/<file>)` in report.md, `<img src="data:image/…;base64,…">` in report.html (one self-contained file, `site.go:56`; an `<img>` runs no SVG script); `TEXT` → verbatim |
| Image rule [MODIFY] (`md.go:437`) | an image embeds iff assembly emitted it from the run's store; `edit` and `ingest` refuse an image in authored text (no archive holds one — measured 0) |
| Seat-facing views [MODIFY] | `show report --assembled --json` carries each block in its section's `text`; bare `show report` (`report/seatread.go`) carries the anchor only, and `AssembledNote` (`layout.go:284`) says shown blocks are assembled content |
| `show evidence --anchor K-…` [MODIFY] | returns the shown output — the stored path for an image; `reproduce` help: read the image, not the caption — a truncated axis is an unsound sentence |
| evidence.md [MODIFY] | each entry adds its shown output |
| Help [MODIFY] | tables are CSV, never printed markdown; a diagram can be a mermaid fence in `text` |
| migrate | replay only: migrated computations are not shown |

### Part 2 — `--show` on `cite` (epoch E4)

| Change | Detail |
|---|---|
| `cite --show text` | the quoted span as a block quote; verification unchanged |
| `cite --show image --page N --region x0,y0,x1,y1` | PDF: `pagerender` at its fixed resolution, cropped (fractions of the page); an image source: the whole image. Stored as a derived cache entry keyed (source sha, page, region, renderer) — in `cache/`, never committed (R-commit); `archive rerun` re-derives it from the re-fetched source |
| Evidence lens `verify` | reads the crop beside `render-page`'s render of the same page — fidelity at the leaf, no seat program between |
| First test | `fetchcache` has no `image/*` branch (`grep 'image/' internal/fetchcache internal/cli/blue`: imports only); whether `cite` accepts an image source is measured before the verb is built |
| migrate | replay only: migrated citations are not shown |

### III.9 Consumer census

Pattern `P` = `(?i:proof|\bprov(?:e|ed|es|ing)\b|reproduc|not_?portable)|Prove(?=[A-Z(])|[a-z]Prove\b`
(PCRE, word-bounded; `provenance`, `improve`, `proven` do not match). Run at `62fe337e` from the
repository root; every hit's meaning read and recorded, one line per consumer, in the ledger.

| # | Command | Hits | Consumers / false | Parts |
|---|---|---|---|---|
| C1 | `grep -rlP "$P" --include=*.go plugins/frank-exchange-of-views/tools \| grep -v _test.go` | 117 files | 73 / 44 | 72 in 1a, of them also 1b 8, 1c 13, 2 2; `cli/lens/verify.go` (an English hit) changes in 2 only |
| C2 | same, `--include=*_test.go` | 195 files | 139 / 56 | 1a; new tests per part (§V) |
| C3 | `grep -rcP "$P"` over `tools/internal/**/help/*.md` (8), `record.proto` (55 lines), `plugins/frank-exchange-of-views/docs` (9), `docs/` (1), debate.js (7 lines), `scripts/agentgen/src` (8), `skills/*/references` (2), `skills/research/SKILL.md` (1), `law/precedents.md` (4 lines), `scripts/universe-stream.py` (1), `CLAUDE.md` (4 lines), `tests/simulator` (10), `scripts/**/*.go` (17, all English or racescope's own `prove()`), `.github/workflows` (15 lines, all English) | — | per line, ledger §C3 | 1a; 1b/1c/2 where marked |
| C4 | `grep -nP "$P" tools/internal/terms/terms.json` | :574 :586 :587 :965 :976 :980 | :574 :586 :965 change | 1a |
| C5 | archives: `ls */proofs/`, `bin/bases.py`, `bin/grepdb.py` | 157 scripts; 239 events; 1 `not_portable`, 1 `observed` | — | 1a (III.1.6) |
| C6 | `grep -rn 'ast\.Image\|"!\[\|<img' --include=*.go tools \| grep -v _test` | `md.go:437` (+ a comment, `pdfextract.go:28`) | 1 | 1c |
| C7 | `grep -nP '(?i:\bprove\b\|reproduce\|proof)' releasegate/fuzz/fuzz_test.go` | prove :1757 :2436 :2455 :2461 :2981 :6604; reproduce :1817 :3463 :3480 :6465; store :4577; bases :5390–:5391; `dirProveDrifts` :2656 (now passes `--variance`); "unjudged proof " :1919 :1942 | all | 1a; an `--input` action 1b; a `--show` action 1c |
| C8 | `grep -rnP '"P-"\|P-…\|\bP-[0-9a-f]{4,8}\b\|\[\^P\|\bP-\b' --include=*.go tools` | 27 lines / 13 files non-test; 215 / 46 test | all | 1a (`anchor.Token("P-…")` in `flags/shapes.go:155,162,194`, `cli/seat/verbs.go:286,293,299`, `cli/lens/close.go:86`, `cli/lens/reproduce.go:144`, `record/refs.go:118`; `P-…` prose at `verbs.go:656,671`; `[^P` at `layout.go:580,591`, `report/notes.go:10–20`, `report/proofs.go:40,45`) |
| C9 | names: `compute`, `computations`, `sources`, `rerun`, `archive` across Go, md, js | `"compute"` only as a refused `--check-kind` spelling (`cli/checkkind_test.go:146`); `sources` = cited sources (`evidenceview.go:278`), the same sense; `rerun` = the `Rerun` enum and capture's re-run audit, the same concept ; `archive` has three senses — the run-archive tarball (`capture.go`, `run-archive/`: the group's sense), the web-archive fetch backend (`fetchcache/backends.go:41`, seat-facing `--via archive`) and the closed gaps the chair spot-checks (`record/spotcheck.go:43`). The group is operator-only and renders on no seat page (no seat constitution names `capture`), so neither seat-facing sense meets it | — | 1a–1b |

Named consumers, meaning checked: `capture/capture.go:2006` (archive contents — changes, 1a);
`capture.go:2007` (no `cache/` — R-commit keeps it true); `capture/proofrerun.go:155` (`Observed`/
`NotPortable` → moved — becomes `VARIANT`, 1a); `fetchcache/shell.go:225` (nothing redistributed —
kept true, §IV R8); `law/precedents.md:11–20` (a past run's facts and a human-promoted holding; its
"proofs deposited" reads true of computations — unchanged, the human's text);
`seat-surface-naming.md:806` (a past run's measurement transcript — unchanged).

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation |
|---|---|---|---|
| R1 | The 1a rename (212 Go files) hides a logic change | M × H × L | Rename commit first, alone, gates green; logic commits after (V1a.1) |
| R2 | Plotting libraries absent in a seat's environment | H × M × L | Help names mermaid-in-`text` and stdlib SVG; a missing module is a refused run, loud |
| R3 | Network reach is declared, not enforced | M × M × H | Two runs catch most moving output; the rest is the lens's soundness duty; #1560 |
| R4 | Image bytes flap across machines (fonts, float, salted ids) | M × M × M | `SOURCE_DATE_EPOCH`, `MATPLOTLIBRC` with `svg.hashsalt` (III.1.2); toolchain sha explains a re-run's difference; `VARIANT` exists |
| R5 | Inlined `text` smuggles process voice | M × M × L | It passes `edit`'s checks; the voice lens reads assembled text |
| R6 | A caption overclaims its chart | M × M × M | Caption = `--reason`, judged against the image (1c help) |
| R7 | **Filesystem residue (R-fs)** — see table below | M × M × H | Stated; `archive rerun` surfaces it from the archive; #1560 |
| R8 | **Redistribution (R-commit)** | L × H × L | No source bytes enter `run-archive/` (public): citations are URL + sha256; crops re-derived. Bundles and their `out/` are committed, as `proofs/` is today — a bundle could embed copied source text (≤ 1 MiB); the lens reads bundles |
| R9 | **Link rot / moving sources** — a re-fetch 404s or returns other bytes | M × M × M | `archive rerun` refuses loudly per input; the computation is then not re-runnable from the archive, and says so — never a silent pass |
| R10 | `archive rerun` re-fetching a whole archive breaks the fetch politeness contract | L × M × L | `fetchcache.Resolve`'s pacing; `--citations` is opt-in (N2) |

**What the two-path copy catches (R7):**

| A read of … | Run 1 vs run 2 | Graded |
|---|---|---|
| a bundle file | same | correct |
| `../x` escaping the copy | different depths → different targets | moved → refused |
| the run tree by relative path (`cache/…`) | missing in both | environment error or exit → refused |
| `sources/C-…` (1b) | materialised in both | correct; verified from the archive by `archive rerun` |
| the run tree by absolute path | the same file in both | **`REPRODUCIBLE` — residue**: fails from the archive, where `archive rerun` reports it |
| the network | may agree | residue (R3) |

## V. Verification Plan

Before any FEOV test: `unset $(env | grep -o '^FEOV_[A-Z_]*')`. Every part ends with
`go -C scripts run ./check` (fail count read), `go -C scripts run ./agentgen -check`,
`go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check`, and
`FEOV_RELEASE_GATE=1 go -C plugins/frank-exchange-of-views/tools test ./releasegate/fuzz/` (record
and flag change). Re-armed by: any edit under `plugins/frank-exchange-of-views/tools`,
`scripts/agentgen/src`, debate.js.

| # | Check | Fails if |
|---|---|---|
| V1a.1 | Rename commit alone: `go -C scripts run ./protogen`, `./schemagen`, `./vocabdoc`, then the shared gates; read every changed golden line and each `--help` before/after | a gate is red, or a help page lost a duty |
| V1a.2 | `internal/compute` tests: agreeing runs → `REPRODUCIBLE`; undeclared drift (stdout, exit, `out/` each) → refused; declared → `VARIANT` + drift; `--variance` over agreeing runs → refused; `../` escape → refused; relative run-tree read → refused; tool call in either run → refused; symlink, >1 MiB, `out`/`sources` entry → refused | any arm passes the wrong way |
| V1a.3 | A script writing "No such file or directory" to stderr only, exit 0 → refused | it is recorded |
| V1a.4 | Keyed retry: same key + different `--variance` → refused; different argv / file bytes → Conflict; identical → idempotent | any is recorded as a second act |
| V1a.5 | `reproduce`: an `out/` file changed with stdout equal → `RERUN_DIFFERS`, drift names the file; exit changed → `RERUN_DIFFERS` | `RERUN_REPRODUCED` |
| V1a.6 | migrate tests on the archives of III.1.6, each assertion row; then the round-1 probe (`audit-r1/probe.sh`) with its NEW arm pointed at each migrated `computations/<sha>/dir/` | a refusal, a failed assertion, or fewer than 107 reproducing |
| V1a.7 | Delete-the-row: each pinned env var, each `remap` admission (`P-` in `hexID`, `archivedToken`, `[^Pn]`), each refusal arm — a test fails for each | a deletion leaves the suite green |
| V1b.1 | `--input`: bytes at `sources/C-…`, read-only, in both copies; a citation not on the record / not cached → refused; a different input under the same key → Conflict | any |
| V1b.2 | `archive rerun` on a fixture archive with no `cache/` against a local server: correct bytes → re-runs; changed bytes → refused naming url and both shas; 404 → refused | a pass on mismatch |
| V1b.3 | Real data: `scripts/universe.sh build`, grep the staged marketplace for `compute --input`, a smoke (`--smoke`) on a question with a CSV source; extract its archive to a clean scratch dir, `feov-record migrate`, `feov-record rerun` | any computation fails from the archive |
| V1c.1 | Assembly goldens for each `Show`; report.html carries `data:image/` and no external `src`; delete each `Show` case → a test fails | a block missing, misnumbered, or external |
| V1c.2 | `edit` and `ingest` adding `![x](y)` → refused; all 29 archives still migrate with zero refusals | accepted, or a migration refusal |
| V1c.3 | `show report --assembled --json` holds the table text after the anchor's paragraph; bare `show report` holds none | either differs |
| V1c.4 | Smoke as V1b.3 with `--show table` and `--show image`; read report.md and report.html | a figure or table absent |
| V1c.5 | Keyed retry with a different `--show` → refused; `out/` with 0 or 2 files, a CSV cell holding a newline, an undecodable PNG → refused | any is recorded |
| V2.1 | `cite --show image` crop equals a crop of `render-page`'s raster; `verify` reads both | differs |
| V2.2 | `archive rerun` re-derives the crop from the re-fetched source; `run-archive/` holds no crop (`tar tzf`) | a crop is committed |
| V* | `/prosthetic-conscience:plan-audit plans/compute-and-show.md` | FAIL |

## Forks

| # | Question | Recommendation | Alternative |
|---|---|---|---|
| F1 | Anchor letter and footnote label | **Ruled: `K-` / `[^Kn]`**, remapped in migrate | ~~Keep `P-` / `[^Pn]`~~ |
| F2 | Who compares a variant's outputs | **Ruled:** the computation lens, against the invariant; a disagreement reaches the bench by the existing contest | Auto-docket every variant re-run |
| F3 | Archive the cached inputs | **Superseded by R-commit** | ~~Archive cited inputs~~ |
| F4 | Migrated `not_portable` / `observed` | **Ruled:** `VARIANT`, drift as invariant, in the manifest | A retained basis no live writer produces |
| F5 | Verb name | **Ruled:** `compute` | `run`; `exhibit` |
| R-split | One PR or several | **Ruled:** 1a / 1b / 1c / 2, each its own epoch and translation | One PR at one epoch |
| R-show | Build `--show` | **Ruled:** yes — the displayed content is the output | Typed markdown tables |
| R-fs | Filesystem sandbox | **Ruled:** not now; residue stated (R7); #1560 | Sandbox in 1a |
| R-commit | Input bytes in `run-archive/` | **Ruled:** none; the archive commits the citation; `archive rerun` re-fetches and verifies | Commit cited bytes |
| N1 | Home of the archive re-run (no operator command exists: `cli/proofrerun.go` is the in-re-run refusal) | **Ruled:** a group `archive`, member `archive rerun <runDir>`; `capture` stays top-level and writes the tarball through package `archive`, the format's one owner | ~~`capture` moved into `archive`~~ (capture closes a live run from its transcripts); ~~top-level `rerun`~~ |
| N2 | `archive rerun --citations`: verify every citation, not only inputs | **Ruled:** ship in 1b, opt-in | Inputs only; citations later |
