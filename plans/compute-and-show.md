# Compute and show — what we computed, what a source shows, and how either is seen

> STATUS (2026-10-24): proposed, not started; revised after audit round 3. Written against
> `a15aaaf7` (origin/main merged; `EventSchema` 43; `feov-migrate` is its own binary; 30 archived
> runs). Forks and rulings by gblock; design gaps resolved by the lead are marked (Forks). Census
> ledgers: `~/.claude/scratch/compute-plan/round4/census.txt` (this round),
> `round2/census.tsv` (Go meanings), `round3/c3-ledger.tsv` (non-Go), `round3/c10-*.txt` (archive text).

## I. Summary & Goals

**Objective.** The report holds two kinds of evidence and can show neither.

- **What a source shows** — `cite` — backs a sentence with a footnote.
- **What we computed** — `prove` — backs a sentence with a footnote; its output is a code fence in
  evidence.md.

No reader sees a table, chart or image built from evidence: a markdown image renders as a link
(`report/md.go:437`), and nothing builds one. `prove` carries three misfits:

1. **The name.** The record already calls the concept a computation (`check_kind: computation`, the
   computation lens); the code calls the `observed` basis "a MEASUREMENT, not a proof"
   (`proof/proof.go:53`). 30 archived runs hold 159 stored scripts: 205 proof events in the 24
   SQLite-era runs, 41 in the 6 era runs.
2. **One file, language by extension.** `proofs/<sha>/script.<ext>`, the interpreter picked from a
   table (`proof/interpreter.go:37`), run with the RUN directory as working directory
   (`proof.go:261`). A helper, a query file or data has nowhere to live, and a script that reads
   relative to itself is graded on a path no re-runner uses: quadratic-formula's 28 proofs are all
   recorded `reproducible`; re-run from the archive, 17 of its 26 stored scripts reproduce neither
   from the store nor from a fresh copy (`audit-r1/probe-results.tsv`; the comment at `proof.go:63`
   says six).
3. **Inputs nobody can reach later.** A script reading a cached source works in the run and fails
   from the archive, which omits `cache/` (`capture/capture.go:2036`).

**Why `--show` (R-show).** A table typed into the report is a second copy of the computation's
output, and copies drift — the defect class `figure_recount_fails` (`record.proto:437`). An image
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
| 1a | `prove` → `compute`; directory + argv; two-value basis; per-act store; one re-run per bundle, judged per act; `K-` anchors | — | E1 |
| 1b | `--input` (raw cited bytes, read-only); package `archive`; operator `archive rerun` | 1a | E2 |
| 1c | `--show table\|image\|text` on `compute`; images exist only as shown blocks | 1a | E3 |
| 2 | `--show image\|text` on `cite`: a colour crop of a pixel box | 1b (`archive rerun` re-derives crops), 1c (placement, `Show`) | E4 |

E1–E4 are the next free epochs when each part lands (main is 43).

**Success criteria.**

| # | Part | Criterion | §V |
|---|---|---|---|
| S1 | 1a | `feov-record blue compute --dir D --quote S --reason R -- python3 main.py` records a computation: bundle = every regular file under `D`; command = the argv after `--`; toolchain = resolved path + sha256 of `python3`. `blue prove` is an unknown command | V1a.1 |
| S2 | 1a | Two runs in fresh copies outside the run tree, each copy's canonical root replaced by `<bundle>` in the output. Equal output, exit and `out/` → `REPRODUCIBLE`; different with no `--variance` → refused naming what moved; with `--variance` → `VARIANT` + drift; `--variance` over agreeing runs → refused | V1a.2, V1a.3 |
| S3 | 1a | A recorded act never changes after write. A bundle is re-run once; every act's verdict compares that re-run against the act's OWN recorded output, exit and `out/` | V1a.4, V1a.5 |
| S4 | 1a | A keyed retry restates every recorded flag or is refused; program, argv and inputs are in the sha | V1a.6 |
| S5 | 1a | Every archive migrates with zero refusals; III.1.6's assertions hold; `archive rerun` over every migrated archive gives each of the 159 stored scripts the outcome committed in the repo (109 reproduced) | V1a.7 |
| S6 | 1b | `--input C-…` materialises the cached bytes read-only at `sources/C-…`; the archive commits no source bytes; `archive rerun` on an extracted archive with no `cache/` re-fetches from where the bytes came, verifies sha256 and re-runs, with 1b's exit semantics | V1b.1–3 |
| S7 | 1c | `--show table` renders the CSV under "*Table n.* R"; `--show image` embeds PNG/SVG under "*Figure n.* R" in report.md and (as `data:` URI) report.html; `--show text` inlines markdown refused where `edit` refuses it | V1c.1 |
| S8 | 1c | An image in authored text is refused at `edit` and `ingest`; an image in any other seat text renders as a link, whatever its destination | V1c.2 |
| S9 | 1c | Shown blocks appear in `show report --assembled --json`; the bare read carries the anchor only and says so | V1c.3 |
| S10 | 2 | `cite --show image --page N --region x,y,w,h` displays a tool-rendered COLOUR crop of that box on the image the seat viewed, the same page area at any resolution; the evidence lens verifies it against a colour render of the page; `cite --show text` displays the quoted span; no crop is committed; an absent crop assembles as a loud placeholder | V2.1–3 |

**Value case — what goes away.**

| Removed | Replaced by |
|---|---|
| The extension→interpreter table, its Windows extra and `argvFor` (`interpreter.go`) | the argv after `--` |
| The grading run from the store (`executeStored`, `proof.go:202`) | two runs, each in a fresh copy at a different path |
| Three bases as free strings (`proof.go:72–74`) | a two-value enum |
| Silent `observed` for output that moved (`proof.go:157`) | refusal, or a declared invariant — the tool never infers |
| Working directory = the run tree (`proof.go:261`) | working directory = a copy of the bundle |
| One `output` per sha, overwritten by the next act (`proof.go:364`) | one output per act, written once |
| One re-run's verdict standing for every act on a script | one re-run per bundle, compared against each act's own output |
| Typed tables that drift from their computation | the program's own output, placed by the tool |
| "Images render as links" | images exist only as shown blocks, embedded from the run's store |

**Not in scope.** Rich media and interactive charts. A machine-checked basis (Lean): the argv holds
`lake env lean X.lean`; a checker for "no `sorry`" and Mathlib are missing. A filesystem or network
sandbox (R-fs; #1560). Extracted text as an input (R-input). Output-signature scanning
(`ErrorSignature`, `proof.go:343`, `--expect-error`) keeps its behaviour (#1466).

## II. Technical Context

- Go, FEOV tools module. Record: SQLite generated from `record/recordpb/record.proto` (`Proof`
  :1404, `Reproduce` :1459 and its comment :1441–1457, `Soundness` :539, `Rerun` :553,
  `CHECK_KIND_COMPUTATION` :344, `EVENT_TYPE_PROOF` :278, oneof :700 with `Proof proof = 42` :719,
  `ABOUT_KIND_PROOF` :1130, `Cite` :1247 with `copy_substitution`/`copy_url` :1301–1302). Epoch:
  `requirements.json` → `scripts/schemagen` → `recordpb/schema_gen.go:18`. The DDL generator types
  string, bool, int32/64, bytes and enum only (`recordsql/schema.go:390`); a list becomes a child
  table (`listTable` :617). Two-field rules are `option (check)` (precedent :845, :985).
- Anchors: one table (`anchor/anchor.go:63`). `P-` becomes `K-` (F1); `K` is held by no id kind.
- `migrate` links into one binary, `feov-migrate` (`cmd/feov-migrate`; `scripts/pluginparity`
  fails any other shipped binary importing it, `record/migrate/source.go:11`). Channel copy
  (`migrate.go:171`), id respelling (`remap.go`: `currentID` :76, `hexID` :78, `archivedToken` :82,
  `id()` :209, field lists :101–131), era mapping (`eraentries.go:52`), value renames
  (`valuerenames.go`), plan-id translation (`planids.go`). An older epoch is refused at open by
  `requireDeclaredSchema` (`recordsql/store.go:594`), pointing to `feov-migrate`.
- Execution posture: a seat could already run anything through Bash; `compute` adds recording, not
  reach; `archive rerun` is new reach for the OPERATOR (R12). No sandbox (`nonet` is test-only).
- Fetch cache: `fetchcache.Read` :382 by sha, `fetchcache.Resolve` :526 by URL; a cache entry
  carries `TDMReserved` (`fetchcache.go:204`), held in the run's fetch index (`cache/`, never
  archived); the rights premise is stated at `fetchcache/shell.go:213–230` and
  `cli/fetchsummary.go:110–112, :345–348, :354–357`. Page rendering (`fetchcache/pagerender.go`):
  resolution per page and per axis, fractional, from the scan (`RenderDPIFor` :75, floor 300
  `tessocr.RenderDPI`, cap 600; `PageRenderDPI` :420); `renderPagePNG` :480 rasterises with PDFium
  (`RenderPageInPixels`) and converts to grayscale (`toGray`). `render-page` is a lens verb only
  (`cli/lens/command.go:32`).
- Site: report.md and report.html are written at the run root (`report/docs.go:391`,
  `report/assemble.go:330`); report.html is one self-contained file rendered from the markdown
  bodies (`report/site.go:59`); composed text is escaped where it is composed (`report/notes.go:43`
  `EscapeNotes`, called at `layout.go:89`, `docs.go:129`, `:211`, `assemble.go:151`,
  `proofs.go:85`).
- Run-archive tarball: written by `capture.go:2045` (`ArchiveRecord`, every entry `Mode: 0o644`
  :2159), read by two private copies — `internal/record/recordtest/archive.go:34` and
  `devcmd/archdump/main.go:149`, both via `os.Create` (umask).
- CI: the Windows pull-request leg runs FEOV's suite with `-short` (`.github/workflows/hooks.yml:693`).
  The `feov-record` job (:370) holds `contents: read` (workflow-level :27–28) and
  `persist-credentials: false` (:388); the `release` job (:1581) holds `contents: write`
  (:1594–1595).

## III. Proposed Changes

Code tree after 1b: `internal/proof` → `internal/compute` (execution, identity, `Rerun`);
`internal/archive` [NEW] (tarball `Write`/`Extract`); `devcmd/archiverun` [NEW] (V1a.7).

### III.0 The record and the store (contract: event schema, store layout)

```proto
// Computation is a program blue ran twice on a bundle, and what it wrote. EVENT_TYPE_COMPUTATION = 24.
message Computation {                                     // [MODIFY] replaces Proof (oneof = 42)
  optional string computation_id   = 1;                   // K-… (F1)
  optional string key              = 2 [(sql) = { flag: "key" }];
  optional string bundle_sha       = 3;                   // III.1.3
  optional string answers          = 4;
  optional string cites            = 5;                   // the METHOD's citation
  optional string drift            = 6;                   // what moved between the two runs
  optional string text             = 7 [(prose) = true, (sql) = { flag: "reason" }];
  optional int32  exit             = 10;                  // the first run's
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
//   rerun = 8 is THIS act's comparison: the bundle's re-run against the act's own output, exit, out/
//   + optional bool   ran        = 9;   // this act executed the bundle: it is the bundle's re-run
//   + optional int32  rerun_exit = 10;  // that re-run's exit; its output and out/ are in rerun/
//   + optional string drift      = 11;  // what differs between the act's recorded result and the re-run
//   checks: drift present ⇒ rerun = DIFFERS; rerun_exit present ⇒ ran
// Rerun [MODIFY] (means): REPRODUCED = "the bundle's re-run wrote this act's recorded output, exit and
//   out/, byte for byte"; DIFFERS = "… other than this act's recorded output, exit or out/; drift names
//   what"; CALLS_THE_TOOL: "bundle" for "script". RERUN_NOT_RUN = 4 → reserved 4, "RERUN_NOT_RUN":
//   `ran` = false carries "compared against the bundle's re-run, ran nothing".
// AboutKind [MODIFY]: ABOUT_KIND_PROOF = 5 → ABOUT_KIND_COMPUTATION = 5.
// Cite [MODIFY, 2]: + optional Show show = 26; optional int32 crop_page = 27;
//   optional int32 crop_x = 28, crop_y = 29, crop_w = 30, crop_h = 31, crop_image_w = 32, crop_image_h = 33;
//   checks: show ∈ {IMAGE, TEXT}; crop_x…crop_image_h present ⇔ IMAGE; crop_w, crop_h > 0;
//   crop_x + crop_w ≤ crop_image_w; crop_y + crop_h ≤ crop_image_h
```

`inputs` is the first list with an `(sql) flag` (`grep -c 'repeated.*(sql)' record.proto` → 0):
`flagOf` (`record/supersede.go:287`) and `flagFor` (`recordpb/requiredfields.go:101`) read the flag
whatever the cardinality; schemagen's output is read in the `schema.sql` golden. Toolchain is two
scalars and the crop seven `int32`s because the generator has no message or double kind. The two
one-way checks hold for migrated events, which carry no drift and no exit (residue); the live
writer always writes `drift` on `DIFFERS` and `rerun_exit` with `ran` (V1a.5). `awaiting_proof` →
`awaiting_computation`. Every `(means)` naming the proof concept (:278, :280, :344, :374, :404,
:449, :452, :455, :541, :555–558, :1127, :1130; :419's "script" is English) names a computation.
remap's field lists class every new field: `inputs` id; `invariant`, `drift`, `Reproduce.drift`
argument; `argv`, `toolchain_*`, `ran`, `rerun_exit`, `crop_*` data. `keycensus_test.go`
`notFields` gains `proof_basis`, `proof_id`, `proof_key`, `proof_sha`, each naming its replacement.

```
<runDir>/
├── computations/<bundle_sha>/
│   ├── dir/            the bundle — content-addressed, written once, never rewritten
│   ├── rerun/          the bundle's one re-run: output, out/ — written once, by the act with ran
│   └── acts/<K-id>/    one per recorded act, written once at its write
│       ├── output      what the first run wrote; each copy's root reads <bundle>
│       └── out/        the files it wrote there (1c shows one)
└── cache/              fetched bytes; Part 2's derived crops and colour pages — never archived
<tmp>/…/<root>/         each run's copy, outside the run tree, two at different depths:
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
| `--dir` replaces `--script`; a one-file computation is a one-file directory | a symlink, a non-regular file, a bundle over 1 MiB, an entry named `out` or `sources` (the tool's mounts), or a `__pycache__` directory or `.pyc` file (named — an interpreter's cache changes the sha between attempts) |
| argv[0]: with a separator and relative (`./run.sh`) → a bundle file, run from the copy, its sha already in the bundle, no toolchain; absolute → that file; bare → resolved on PATH; toolchain = resolved path + sha256 | argv empty; a relative argv[0] that is not a bundle file; a bare argv[0] not on PATH (named); a program the OS cannot start (`./run.sh` on Windows) |
| `--variance` states what holds when the runs differ | given and the runs agree (precedent: `--expect-error` with no failure, `cli/blue/prove.go:128`) |
| Recorded flags restate on a keyed retry (`seat.RepeatsKeyed`, `cli/seat/keyedretry.go:31`) | asserted set (`proveAsserted`, `prove.go:191`) + `--variance` → `invariant` differs; a different sha under the key → Conflict (`prove.go:84`) |
| Output and `out/` have no size limit (R-size) | — |

#### III.1.2 Execution [MODIFY/DELETE] — `internal/proof` → `internal/compute`

1. Copy the bundle twice, to two fresh directories at different depths, neither inside the run tree;
   each gets an empty `out/`. Every file is written `0o644`, argv[0]'s bundle file `0o755`, by
   explicit `Chmod` (umask-independent). Each copy's root is its canonical OS path
   (`filepath.EvalSymlinks`; on Windows the long-path form of an 8.3 `TEMP`).
2. Run each, `Timeout` (60 s) apiece, with the copy's root as working directory, under the checker
   environment (`FEOV_PROOF_RERUN` → `FEOV_COMPUTE_RERUN`, `seatenv/seatenv.go:87`) and the pins
   `SOURCE_DATE_EPOCH=0`, `PYTHONHASHSEED=0`, `TZ=UTC`, `LC_ALL=C`, `PYTHONUTF8=1`,
   `MATPLOTLIBRC=<a tool-written file outside both copies holding svg.hashsalt>`.
3. Take everything the program writes — stdout and stderr as one stream (`proof.go:263`) — and
   replace each occurrence of that copy's canonical root with `<bundle>` (P-norm). `ErrorSignature`
   scans the result. `out/` files are compared and stored as bytes, untouched.
4. Refuse at write: a tool call in either run; an `ErrorSignature` line without `--expect-error`;
   differing output, exit or `out/` without `--variance`. A non-zero exit both runs agree on is
   recorded, not refused (R7).
5. Write artifacts first, the event last: `dir/` if absent (present → its listing must hash to
   `bundle_sha`); mint the `K-` id (`record.NewID`, random, `record/id.go:22`); write
   `acts/<K-id>/output` and `out/` from the first run — refused if `acts/<K-id>/` exists; then
   `AppendPlaced`. A refused append removes `acts/<K-id>/`. A crash between the two leaves an
   `acts/` directory no event names: every reader enumerates acts from the record, never the
   directory, so it is inert, and a keyed retry mints a fresh id. An idempotent keyed retry writes
   nothing.
6. Remove both copies, clearing read-only bits first (Windows `RemoveAll` refuses read-only files).
7. Delete: `interpreter.go`, `executeStored`, `NotPortable`, `Observed`, `StoredScript`, `store`'s
   per-sha `output`.

#### III.1.3 Identity

`bundle_sha` = sha256 over a sorted listing of items, each item a sequence of length-prefixed fields
(`<len>:<bytes>`, so a newline in an argument or path forges nothing): `argv <i> <arg>`; `input
<C-id> <sha256>` (1b); `file <path> <sha256>`. No mode is identity: the one mode that matters,
argv[0]'s executable bit, is set by the tool in every copy (III.1.2), so `dir/`, `archive.Write`
and `archive.Extract` need preserve none. A computation with no inputs has no `input` items, so 1b
leaves every 1a `bundle_sha` unchanged. Computed without running, so a keyed retry tells a
crash-retry from a collision before executing.

#### III.1.4 Re-run — one per bundle, judged per act [MODIFY] (ruling Re-run, 2026-10-24)

`compute.Rerun(run, bundle_sha)` runs III.1.2 steps 1–3 and 6 from `dir/` and writes `rerun/`
once. `compare(act, rerun)` sets the act's verdict from its own `acts/<K-id>/` — output, exit and
`out/`; an absent recorded `out/` equals an empty one. On `VARIANT` a differing re-run is expected:
the lens's reason says whether the output satisfies the invariant (F2).

| Carrier | Becomes |
|---|---|
| `record/proofverdict.go` header :11–23 | "THE RE-RUN belongs to the BUNDLE, run once; THE COMPARISON belongs to the ACT; SOUND belongs to the ANCHOR" |
| `scriptRerunsOf` :50 | `bundleRerunsOf` (latest `ran` reproduce per `bundle_sha`) + `actVerdictsOf` (latest reproduce per `computation_id`, whose `rerun` is that act's comparison); `proofVerdictOf` :137 reads the act's |
| `validateReproduce` :70 | `ran` = false → the record holds a `ran` reproduce on the bundle, else refused; a standing `CALLS_THE_TOOL` refuses `sound` (unchanged) |
| `cli/lens/reproduce.go:93–123` | a standing re-run with its raw result → compare this act against `rerun/` and its `rerun_exit`, `ran` = false; none (or a migrated one with no `rerun/`) → `compute.Rerun`, `ran` = true |
| `reproduce.go:144` flag help | "the tool re-runs the computation's bundle unless the bundle already has a re-run; either way it compares THIS act's own recorded output, exit and out/ against that re-run" |
| `RecordedProofs`/`ProofAt` (`citationid.go:346`, `:390`) | per-act `Rerun` from `actVerdictsOf`, plus the standing re-run's seat |
| `proofChecksOwed` + item (`available.go:522`, `:584`) | the act's own state; the item says whether the bundle's re-run exists (then `reproduce` runs nothing) |
| `record/evidenceview.go:454` | per act; a row whose bundle backs acts of differing basis carries `bundle_bases` (each act with its basis) — and evidence.md's entry states the same |
| `capture/proofrerun.go:60–110` (check `proof-rerun` → `computation-rerun`) | one row per bundle holding every act (id, basis); one `compute.Rerun` per sampled bundle; each act compared against its own output and graded by its own basis — `REPRODUCIBLE` differing FAIL, `VARIANT` moved (`:150`); the detail names any bundle whose acts' bases disagree |
| record.proto :555–558, Reproduce comment :1441–1457 | III.0 |
| `help/reproduce.md` :7 :9 :15 :23 | the bundle is re-run once; the verdict is this act's own comparison |

#### III.1.5 Rename sweep [MODIFY] — names only; III.1.4 carries the logic

| Site | Becomes |
|---|---|
| anchor kind `proof`/`P-`/tag `proof` (`anchor.go:63`) | `computation`/`K-`/tag `computation` |
| `proofNote` `[^P%d]` (`layout.go:591`), `wovenNote` `(?i)\[\^p?[0-9]+\]`, `WovenNotesRule` "P and digits" | `[^K%d]`, `k?`, "K and digits" |
| `weaveProofRefs` (`report/proofs.go:49`), evidence.md `## Proofs` | `weaveComputationRefs`, `## Computations` (argv, toolchain, basis, invariant, drift; each bundle file fenced by extension, or one line — path, size, sha256 — when it is not UTF-8 or holds NUL; output fenced) |
| `awaiting_proof` (`recordsql/views.go:566`), `ProofAnswers` (`estoppel.go:258`), `proofChecksOwed`/`proofOwnerOf` (`available.go:522`/`:557`), "unjudged proof" (`retirement.go:161`), `proofs_unverified` (`evidenceview.go:329`), `unsoundProof` (`record/unsoundproof.go`), `BlockerUnsoundProof` "unsound_proof" (`passblockers.go:36`) | `…computation…` |
| `ABOUT_KIND_PROOF` (`about.go:70`, `mintbudget.go:393`, `:409`, `unsoundproof.go:45`, `:81`, `refs.go:126`; help `cli/lens/finding.go:144–145`) | `ABOUT_KIND_COMPUTATION`, `--about-kind computation` |
| `EvidenceProofAtJSON` `script`/`script_file` (`evidenceview.go:598`, `:614–615`; help `cli/seat/verbs.go:297`) | `files` (path, body) and `output` |
| `ArchiveRecord` `proofs/` (`capture.go:2095`, comment :2035); `migratecli.go:53` "run channels (inputs, proofs)"; `site.html.tmpl:33` | `computations/` |
| terms.json :574 :586 :587 :780 :965 :976 :980 | "a computation", "citation or computation"; `vocabdoc` regenerates `docs/vocabulary.md` |

#### III.1.6 migrate — one translation [NEW] (contract: record migration)

| Old | New |
|---|---|
| `Proof` / era `proof` event (`eraentries.go:52`) | `Computation` |
| `proofs/<sha>/script.<ext>` (sha dirs only; `copyChannels` skips `proofs/`) | `computations/<new>/dir/script.<ext>`; `<new>` from the old store alone (argv + tree); `meta.json` is not copied — its fields are the era event's |
| `proofs/<sha>/output` | `acts/<K-id>/output` for each act on that sha (SQLite era); the era event's own `output` (JSONL era) |
| extension | argv: `.py` `python3`, `.sh` `bash`, `.js`/`.mjs` `node`, `.go` `go run` — each + `script.<ext>`; toolchain absent |
| `proof_basis` `reproducible` / `observed` / `not_portable` | `REPRODUCIBLE` / `VARIANT`, invariant = drift (F4) |
| `proof_sha` in `Reproduce` and the manifest | `bundle_sha` through the old→new map |
| `Reproduce` that ran (`reproduced`, `differs`, `calls_the_tool`) / `not_run` | `ran` = true / `ran` = false with `rerun` = its bundle's standing re-run's value: migrated acts on one sha hold one output, so that is each act's comparison |
| `about_kind` `proof` | `computation` |
| `P-`/`p-` + 8 hex | `K-` + the same hex: `hexID` (:78) and `archivedToken` (:82) admit `P-` — today `currentID` (:76) matches `P-` and `id()` (:209) would otherwise hash a new hex |
| `<!--proof:…-->`; `[^Pn]` in argument fields | `<!--computation:K-…-->`; `[^Kn]` |

| Archive (measured; `round2/c9-archive-scan.txt`, `round4/census.txt`) | Holds | Asserted after migrate |
|---|---|---|
| all 30 | 159 stored scripts | zero refusals; no `\bP-[0-9a-f]{8}\b` outside `Plan.printed`, no `<!--proof:`, no `[^P`; every `K-` id's hex equals its source id's; every act has `acts/<K-id>/output` |
| 2026-09-09_is-91-prime-b | `not_portable` `p-808ae840` | `VARIANT`, invariant = drift "runs from the proof store and says something else — exit code moved 0 -> 1 between runs" |
| 2026-08-22_is-7-prime-run2 (era) | `observed` `p-568517ec` | `VARIANT`, invariant = drift "output differs from byte 131 between runs"; its output is its event's |
| 2026-10-12_is-91-prime-wave-b | `[^P5]` in a finding | `[^K5]` |
| 2026-10-10_is-91-prime-m19 | 125 `P-` ids | `K-`, same hex |
| 2026-09-11_is-91-prime-b6 | the `.go` script | argv `go run script.go` |
| 2026-10-17_is-91-prime-07d | 9 acts over 6 scripts | 6 `dir/`, 9 `acts/` |
| 2026-10-24_is-91-prime-m32 | 7 acts on 2 scripts; 9 `not_run` reproduces | 2 `dir/`, 7 `acts/`; each `not_run` → `ran` = false, `rerun` = `reproduced` |

**Residue, accepted** (`round3/c10-*.txt`, `round4/census.txt`):

| Residue | Size | Why it stays |
|---|---|---|
| `P-` ids and "unjudged proof" in `Plan.printed` | 50 ids, 4 runs (m28, m29, m30, m32) | printed plans keep the wording they were printed in (`planids.go:14–18`) |
| `proofs/<sha-prefix>` paths and bare old-sha prefixes in seat text (`finding.text`, `reproduce.note`, `close.*`, `log.text`, …) | 51 paths, 10 runs; 169 prefixes, 18 runs; 0 era | a seat's prose account of where it read; recovering a path from prose is a pattern over free text — the act it describes keeps its id |
| store paths inside program output | 9 `Reproduce.observed_output`, 1 stored `output` | bytes the program wrote |
| acts sharing a sha receive the sha's one stored `output` | 110 acts on 27 shas, 20 runs | all recorded `reproducible`: the one output is each act's by its basis |
| scripts that reproduce neither from the store nor a copy | 50 of 159, 9 runs (17 quadratic-formula, 21 record-store-authority) | 48 migrate to `REPRODUCIBLE` and `archive rerun` prints `differs`, exit 1; 2 (`62d048c0`, `ab6aa892`) are `VARIANT` → `moved`. They were graded from the run tree; the committed outcomes (V1a.7) pin each |
| migrated standing re-runs hold no raw result (no `rerun/`, no `rerun_exit`); migrated `DIFFERS` hold no `drift` | every migrated `ran` reproduce | not recorded then; the next live `reproduce` on such a bundle re-runs it, once |
| `proofs/synth/` | 17 drafts, 1 run | each is also stored under its sha |
| author script path (`Proof.script`) | every proof | the argv replaces it |

#### III.1.7 Agent-facing text [MODIFY] (contract: prompts and constitutions)

Sources only, then `agentgen`, `enginegen`, `vocabdoc`: `help/prove.md` → `help/compute.md`;
`help/reproduce.md` (III.1.4); `help/{dispatch,close,mint,edit}.md`;
`scripts/agentgen/src/{blue-researcher,blue-synthesizer,red-lens-computation,red-lens-evidence}.md`,
`fragments/{blue-search-and-cite,lens-audit-and-mint,lens-duties}.md`, both preloaded-skill sources;
`skills/research/SKILL.md:7`; debate.js authored :229 :978 :1008 :1166 (`unsound_proof` :417 is in
the generated envelope block :370–433); `references/report_template.md:147`,
`references/run-directory.md` (:28 "with script", :77); terms.json (III.1.5); FEOV docs (ledger
§C3); `docs/local-development.md:226`; `CLAUDE.md:34`; `scripts/universe-stream.py:24`. Seat prompts
name the act — "compute it"; the prompt gate matches `<role> <verb>` (`seatprobe/naming.go:78`).
The `compute --help` page teaches: a directory and a command; output records `<bundle>` for the
copy's path; make output deterministic or state the invariant; a helper the program executes runs
through its interpreter (only argv[0] is made executable); a non-zero exit both runs agree on is
recorded; output and `out/` are unbounded and committed.

### Part 1b — `--input` and `archive rerun` (epoch E2)

| Change | Detail |
|---|---|
| `compute --input C-…` (repeatable) [MODIFY] | the citation's RAW cached bytes (`fetchcache.Read`) — never extracted text (R-input). Placed read-only at `sources/C-…` in each copy; id + sha256 enter `bundle_sha`. Refused: a citation not on the record; bytes the cache does not hold; a citation whose bytes the tool composed — `copy_substitution` `recovered` or `named_backend` with no `copy_url` (`record.proto:1237`, `:1239`): no document exists to re-fetch |
| `reproduce`, capture's audit [MODIFY] | `compute.Rerun` materialises `sources/` the same way |
| Package `archive` [NEW] (`internal/archive`) | the one writer and the one reader of the run-archive tarball: `archive.Write` (from `ArchiveRecord`, `capture.go:2045`; `capture` calls it) and `archive.Extract` (replacing `internal/record/recordtest/archive.go:34`'s loop and `archdump`'s `untar`, `main.go:149`). Every entry `0o644`, created with explicit permissions; identity holds no mode (III.1.3). Capture's transcript tarball (`capture.go:1748`) is another format and stays. It commits no `cache/` (R-commit) |
| `feov-record archive rerun <runDir>` [NEW] (N1) | input: an extracted archive brought to the current epoch by `feov-migrate --from <runDir> --to <freshDir>` (an older epoch is refused by `requireDeclaredSchema`, pointing to `feov-migrate`). Re-fetch each input from where its bytes came — `copy_url` when set, else `url` (`fetchcache.Resolve`, paced) — verify sha256, then `compute.Rerun` each bundle once and compare every act. Writes no event. `--citations` verifies every citation the same way (N2); a tool-composed citation prints `refused` (no document). `--help` states: it runs every archived program with the operator's privileges, filesystem and network (R12) |
| Rights [MODIFY] — every carrier (C4) | the decision, in each: computing over a cited body is accepted whatever `TDMReserved` says; output and `out/` are committed to `run-archive/` and may carry more of a source than a quotation; the bodies are never committed, kept as a corpus beyond the cache, or trained on. `fetchcache/shell.go:213–230` (comment); `cli/fetchsummary.go:110–112` and `:345–348` (comments); `:354–357` (the seat-printed note: "…What it forbids is keeping a corpus of it or training on it, which this tool does not do. A program you run over it is accepted; its output is committed"); `docs/citations-and-bibliography.md:74–75`. `TDMReserved` stays in the run's fetch index (`cache/`), not copied onto `Computation` (R-rights) |
| Help, prompts [MODIFY] | `compute --help`: inputs are citations, raw bytes, read at `sources/<id>`; blue-researcher source: compute from what you cited |
| migrate | replay only: migrated computations hold no inputs |

**`archive rerun` output.** One line per act, naming each input and how it was verified
(`K-… reproduced — inputs: C-a re-fetched, sha256 verified`), then a summary: computations by
outcome, and how many had an input re-fetched, verified and re-run.

| Outcome | Printed | Exit |
|---|---|---|
| re-run equals the act's own result | `reproduced` | 0 |
| `VARIANT`, re-run moved | `moved` (drift) | 0 |
| `REPRODUCIBLE`, re-run differs | `differs` (drift) | 1 |
| input 404, fetched sha ≠ recorded, or tool-composed | `refused` (url, recorded sha, fetched sha) | 1 |
| bundle absent, older epoch | `refused` (named) | 1 |

### Part 1c — `--show` on `compute` (epoch E3)

| Change | Detail |
|---|---|
| `compute --show table\|image\|text` [MODIFY] | exactly one file in `out/`, decoded as its kind: `table` CSV with a header row (a cell holding a newline refused); `image` PNG or SVG; `text` markdown under `edit`'s refusals and its voice advisory (`cli/blue/edit.go:104`). Zero files, two, or one that does not decode → refused. Joins the asserted set |
| Assembly [MODIFY] (`report/layout.go`, `report/computations.go`) | block after the anchor's paragraph, numbered per kind in document order: `TABLE` → markdown table, pipes escaped, `*Table n.* R`; `IMAGE` → `![Figure n. R](computations/<sha>/acts/<K-id>/out/<file>)` in report.md (run root, so the link resolves in the run and in an extracted archive), `<img src="data:image/…;base64,…">` in report.html (an `<img>` runs no SVG script); `TEXT` → verbatim |
| Image rule [MODIFY] (`notes.go:43`, `md.go:437`) | ONLY ASSEMBLY OPENS AN IMAGE, as only the weave opens a footnote: `EscapeNotes` → `EscapeComposed`, which also escapes `![` outside code at its five call sites, so no seat text assembly composes (mint, fix, inquiry, ruling, evidence) parses as an image in report.md or report.html. `edit` and `ingest` refuse an image in authored text (no archive holds one — measured 0). An `ast.Image` then comes only from a shown block; md.go embeds one whose destination is a placed store path, and renders any other as a link |
| Seat-facing views [MODIFY] | `show report --assembled --json` carries each block in its section's `text`; bare `show report` (`report/seatread.go`) carries the anchor only, and `AssembledNote` (`layout.go:284`) says shown blocks are assembled content |
| `show evidence --anchor K-…` [MODIFY] | returns the act's shown output — the stored path for an image; `reproduce` help: read the image, not the caption — a truncated axis is an unsound sentence |
| Voice lens [MODIFY] `scripts/agentgen/src/red-lens-voice.md` | a shown text block is report text: read the report as assembled (R5). A finding's `--quote` must be report text, which a shown block is not: quote the anchor's sentence and name the block; the fix is `edit` the sentence, `retire` the `K-` anchor, compute again |
| evidence.md, help [MODIFY] | each entry adds its shown output; tables are CSV, never printed markdown; a diagram can be a mermaid fence in `text` |
| `devcmd/archdump` [MODIFY] | copies `computations/` beside the report set, so `diff -r` of two dumps compares shown images too |
| migrate | replay only: migrated computations are not shown |

### Part 2 — `--show` on `cite` (epoch E4) (rulings Crop, Colour)

| Change | Detail |
|---|---|
| `cite --show text` | the quoted span as a block quote; verification unchanged |
| `cite --show image --page N --region x,y,w,h` | a PIXEL box on the image the seat viewed: for a PDF, `render-page`'s image of page N (offered to blue as well as the lens), whose size the tool records as `crop_image_w × crop_image_h` from the same `PageRenderDPI`; for an image source, the image itself (`crop_page` absent). The event is self-describing — no resolution field. The result returns the crop's path for the seat to read before the citation stands |
| Colour crop | renderer: PDFium `RenderPageInPixels`, as `renderPagePNG` (`pagerender.go:480`), with the RGBA raster encoded to PNG as rendered — `renderPagePNG` gains the colour model as an argument, and `RenderPages`/`RenderOnePage` keep passing grayscale, so OCR renders and `render-page` shas are unchanged. Any derivation renders at its own resolution and scales the box by its raster ÷ `crop_image_w`/`crop_image_h` (rounded) → the same page area. Stored as a derived cache entry keyed (source sha, page, box, image size) — in `cache/`, never committed (R-commit); `archive rerun --citations` re-derives it from the re-fetched source |
| Absent crop | report.md, report.html, `show report --assembled` and `devcmd/archdump` render "*Figure n.* R — <source title>, page N: image not held in this copy" (subject terms only — report-voice ruling); run.md's log names the absent crop (C-…, box) and `archive rerun`, which re-derives it. Never a silent omission |
| Evidence lens | `show evidence --anchor C-…` on a shown citation returns the crop and a colour render of its page (both derived into `cache/` on read, as `render-page` writes); `verify` reads both — fidelity at the leaf, no seat program between |
| First test | `fetchcache` has no `image/*` branch (`grep 'image/' internal/fetchcache internal/cli/blue`: imports only); whether `cite` accepts an image source is measured before the verb is built |
| migrate | replay only: migrated citations are not shown |

### III.9 Consumer census

Pattern `P` = `(?i:proof|\bprov(?:e|ed|es|ing)\b|reproduc|not_?portable)|Prove(?=[A-Z(])|[a-z]Prove\b`
(PCRE; `provenance`, `improve`, `proven` do not match). Run at `a15aaaf7` from the repository root;
every hit's meaning read and recorded in the ledgers (new files since round 3: `round4/census.txt`).

| # | Command | Hits | Consumers / false | Parts |
|---|---|---|---|---|
| C1 | `grep -rlP "$P" --include=*.go plugins/frank-exchange-of-views/tools \| grep -v _test.go` | 120 files | 76 / 44 (+`lens/finding.go`, `record/about.go`, `record/unsoundproof.go`, `migratecli/migratecli.go`; −`cli/migrate.go`) | 1a; also 1b 8, 1c 13, 2 2; III.1.4's carriers |
| C2 | same, `--include=*_test.go` | 203 files | 147 / 56 | 1a; `keycensus_test.go` `notFields`; new tests per part (§V) |
| C3 | `git ls-files \| grep -v '\.go$' \| grep -vE '^(run-archive\|plans\|research\|ideas\|projects)/' \| xargs grep -cP "$P"` (non-zero) | 95 files | 36 regenerated (11 agents, 2 preloaded skills, 7 difftest + 3 migrate goldens, `feov-migrate` help golden, `schema.sql`, 10 simulator, `vocabulary.md`); 34 authored (`round2/c3-nongo.tsv` + `fragments/lens-duties.md:5`); `site.html.tmpl:33`, terms.json; `payload-keys.txt` frozen; 22 English | 1a; 1b/1c/2 where marked |
| C4 | not `P` — rights: `git grep -niI 'tdm\|data-mining'` outside tests and archives | `shell.go:213–230`, `fetchsummary.go:110–112, :345–348, :354–357`, `fetchcache.go:110–117, :196–203` (facts, no premise — unchanged), `citations-and-bibliography.md:74–75`; plus `red-lens-voice.md` (1c), `help/reproduce.md:9`, `run-directory.md:28` (1a) | all | as marked |
| C5 | archives: `ls */proofs/`, `bases.py`, `grepdb.py`, NOT_RUN and sha-prose scans (`round4/m32/`) | 159 scripts; 246 events; 55 `not_run` in 7 runs; III.1.6 residue counts | — | 1a |
| C6 | `grep -rn 'ast\.Image\|"!\[\|<img' --include=*.go tools \| grep -v _test` | `md.go:437` (+ a comment, `pdfextract.go:28`) | 1 | 1c |
| C7 | `grep -nP '(?i:\bprove\b\|reproduce\|proof)' releasegate/fuzz/fuzz_test.go` | 281 lines: prove :1899 :2720 :2739 :2745 :3265; reproduce :1959 :2069 :3768 :3785; `dirProveDrifts` :2934–:2940 (passes `--variance`); bases :5749–:5752; "unjudged proof " :2188 :2211 | all | 1a; a two-acts-one-bundle action 1a; `--input` 1b; `--show` 1c |
| C8 | `grep -rnP '"P-"\|P-…\|\bP-[0-9a-f]{4,8}\b\|\[\^P\|\bP-\b' --include=*.go tools` | 27 lines / 13 files non-test; 270 / 49 test | all | 1a (code: `anchor.go:63`, `flags/shapes.go:155,162,200`, `cli/seat/verbs.go:283,290,297,655,670`, `cli/lens/close.go:84`, `cli/lens/reproduce.go:144`, `record/refs.go:118`, `layout.go:580,591`, `report/notes.go:10–20`, `report/proofs.go:40,45`; the rest comments) |
| C9 | names across Go, md, js: `compute`, `computations`, `sources`, `rerun`, `rerun/`, `archive`, `acts`, `<bundle>`, `ran`, `rerun_exit`, `crop_*`, `bundle_bases`, `EscapeComposed`, `archiverun` | `"compute"` only as a refused `--check-kind` spelling (`cli/checkkind_test.go:146`); `sources`, `rerun`, `acts`, `archive` (tarball; also `fetchcache/backends.go:41`, `record/spotcheck.go:43`, neither seat-facing on an operator page) as round 3; the rest 0 hits | — | 1a–2 |
| C10 | `grep -rn 'tar\.NewReader' --include=*.go tools \| grep -v _test` | 2 (`recordtest/archive.go:66`, `archdump/main.go:159`) | both move onto `archive.Extract` | 1b |

Named consumers, meaning checked: `capture.go:2035–2036` (archive contents, 1a; no `cache/`, kept
by R-commit); `pagerender.go:298–308` ("the day a caller needs colour, the option and the field
arrive together" — Part 2 brings both); `law/precedents.md:11–20`, `seat-surface-naming.md:806`
(past runs' facts — unchanged).

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation |
|---|---|---|---|
| R1 | The 1a rename (223 Go files) hides a logic change | M × H × L | Rename commit first, alone, gates green; III.1.4 and the logic after (V1a.1) |
| R2 | Plotting libraries absent in a seat's environment | H × M × L | Help names mermaid-in-`text` and stdlib SVG; a missing module is a refused run, loud |
| R3 | Network reach is declared, not enforced | M × M × H | Two runs catch most moving output; the rest is the lens's soundness duty; #1560 |
| R4 | Image bytes flap across machines (fonts, float, salted ids) | M × M × M | the pins and `MATPLOTLIBRC` (III.1.2); toolchain sha explains a re-run's difference; `VARIANT` exists |
| R5 | Inlined `text` carries process voice | M × M × L | the voice lens's duty to read the assembled report (1c). **Residue:** `edit`'s voice check is advisory (`edit.go:104`, `:171`); the lens's read is judgement, not a gate |
| R6 | A caption overclaims its chart | M × M × M | Caption = `--reason`, judged against the image (1c help) |
| R7 | **Filesystem residue (R-fs)** — table below | M × M × H | Stated; #1560 |
| R8 | **Cited bytes leave the cache (R-commit, R-rights)** — table below | M × M × L | Source bodies never enter `run-archive/`; every other channel is accepted residue, stated in every C4 carrier and `compute --help` |
| R9 | **Link rot / moving sources** | M × M × M | `archive rerun` refuses loudly per input and exits 1 — never a silent pass |
| R10 | `archive rerun` breaks the fetch politeness contract | L × M × L | `fetchcache.Resolve`'s pacing; `--citations` is opt-in (N2) |
| R11 | Unbounded output and `out/` (R-size) bloat `run-archive/` | M × L × L | Stated in `compute --help`; the bundle stays ≤ 1 MiB |
| R12 | **`archive rerun` runs archived programs with the operator's privileges, filesystem and network** | L × H × M | Stated in its `--help`; in CI it runs only in a job holding no write credential — the nightly `feov-record` leg (`contents: read`, `persist-credentials: false`), never the `release` job (`contents: write`) — ruling CI. Residue: a program can still read the runner's environment and reach the network |

**R7 — what step 4 grades** (`ErrorSignature` matches case-sensitive `No such file or directory`,
`Traceback`, `command not found`, `Permission denied`, `cannot access`, `: not found`,
`ModuleNotFoundError` — `proof.go:333–341`):

| A program that … | Run 1 vs run 2 | Graded |
|---|---|---|
| reads a bundle file, or prints its own path | same (`<bundle>`) | correct |
| reads `../x` where `x` exists at one depth only | different | moved → refused (or `VARIANT` if declared) |
| reads a missing path (`../x` at both depths, the run tree by relative path) and the output carries a signature (Python traceback, shell message) | same | refused without `--expect-error` |
| reads a missing path and says so unmatched (Node `ENOENT: no such file or directory`, Go `open x: no such file or directory`, Windows `The system cannot find the file specified`) or tolerates it — exit 0 or a non-zero exit both runs share | same | **`REPRODUCIBLE`, vacuous — residue**: the computation lens judges bundle and output; #1466 |
| reads `sources/C-…` (1b) | materialised in both | correct; verified from the archive by `archive rerun` |
| reads the run tree by absolute path | the same file in both | **`REPRODUCIBLE` — residue**: re-runs equal on the run's own machine; elsewhere `archive rerun` exits 1. 27 of 157 round-2 scripts name an absolute path, 11 read `cache/` |
| reads the network | may agree | residue (R3) |
| prints its root spelled otherwise than canonical (forward slashes on Windows) | each copy's own | moved → refused, loud |

**R8 — channels for a cited source's bytes:**

| Channel | Committed | Graded |
|---|---|---|
| the source body (`cache/`, `sources/C-…`, crops, colour pages) | never | R-commit |
| a bundle file copied from a source (the only REPRODUCIBLE route in 1a) | yes, ≤ 1 MiB | accepted residue, as `proofs/` today |
| output and `out/` computed over a source (1b), TDM-reserved or not | yes, unbounded | accepted residue (R-rights), stated to the seat in the fetch summary's note |

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
| V1a.2 | `internal/compute`: agreeing runs → `REPRODUCIBLE`; undeclared drift (output, exit, `out/` each) → refused; declared → `VARIANT` + drift; `--variance` over agreeing runs → refused; `../x` present at one depth → refused; a missing relative path printing lowercase `ENOENT`, exit 1 both runs → `REPRODUCIBLE`, exit 1 recorded; tool call → refused; symlink, > 1 MiB, `out`/`sources` entry, `__pycache__`/`.pyc` → refused; `./run.sh` (Unix) and `./prog.exe` (Windows) → run, toolchain absent; `./run.sh` on Windows → refused naming the start failure; under umask 077 and 022 the second act on the bundle finds `dir/` hashing to `bundle_sha`; an argv holding a newline and one splitting it differently → different shas; no copy remains, read-only `sources/` included | any arm passes the wrong way |
| V1a.3 | P-norm: a Python program printing `__file__` and `os.getcwd()` and raising, its copy under a symlinked temp dir → `REPRODUCIBLE`, recorded output holds `<bundle>` and neither copy's path; "No such file or directory" on stderr at exit 0 → refused. Delete the replacement, or the `EvalSymlinks` → the first fails | it is recorded otherwise |
| V1a.4 | Per act, per bundle: a `VARIANT` bundle whose program prints and increments a counter file; acts A (recorded 1) and B (recorded 3); counter reset to 3; `reproduce` A → `ran`, `rerun/` written, A `DIFFERS`; `reproduce` B → runs nothing (`ran` = false), B `REPRODUCED`; A's `acts/` byte-identical throughout. capture's audit over the same: one re-run, A moved, B reproduced; a bundle with a `REPRODUCIBLE` and a `VARIANT` act → `bundle_bases` on the evidence row and named in the audit detail. A pre-existing `acts/<K-id>/` → refused; a refused append leaves no `acts/<K-id>/`. Delete `compare`'s act argument (compare against the first act) → B's verdict fails | B's verdict reads A's output, or an act changes after write |
| V1a.5 | `reproduce`: an `out/` file changed with output equal → `DIFFERS`, drift names the file; exit changed → `DIFFERS`; absent recorded `out/` with an empty re-run `out/` → `REPRODUCED`; every live `DIFFERS` carries drift and every `ran` carries `rerun_exit`; a migrated `ran` with no `rerun/` → the next reproduce re-runs once | wrong verdict or a missing field |
| V1a.6 | Keyed retry: different `--variance` → refused; different argv / file bytes → Conflict; identical → idempotent, nothing written. Record write: `VARIANT` without drift or invariant, `REPRODUCIBLE` with either, `show` = `SHOW_UNSPECIFIED`, `drift` on `REPRODUCED`, `rerun_exit` without `ran` → refused by the `(check)` | any is recorded |
| V1a.7 | migrate tests on the archives of III.1.6, each assertion and residue row (m32's `not_run` rows included). Then `go -C plugins/frank-exchange-of-views/tools run ./devcmd/archiverun`: each of the 30 archives through `archive.Extract`, migrate and `archive rerun`, every act's outcome against `devcmd/archiverun/testdata/outcomes.tsv` (159 rows: archive, old sha, outcome — 109 `reproduced`, 48 `differs`, 2 `moved`). Run by hand before E1 merges and nightly as a step of the `feov-record` Ubuntu leg on `schedule` (`contents: read`, `persist-credentials: false`); never in the `release` job (a workflow edit: CI's `actionlint` reads it); a missing interpreter fails it | a refusal, a failed assertion, or any act's outcome other than its committed row |
| V1a.8 | Delete-the-row: each pinned env var, each `remap` admission (`P-` in `hexID`, `archivedToken`, `[^Pn]`, `not_run`, `about_kind proof`), each refusal arm, each `notFields` entry — a test fails for each | a deletion leaves the suite green |
| V1a.9 | Smoke: `scripts/universe.sh build`, grep the staged marketplace for `blue compute`, `--smoke` on a numeric question; capture's `computation-rerun` | the run recorded no computation, or a FAIL |
| V1b.1 | `--input`: raw bytes at `sources/C-…`, read-only, in both copies; a citation not on the record / not cached / `recovered` with no `copy_url` → refused; a different input under the same key → Conflict; a zero-input bundle hashes to the sha pinned in 1a's test | any |
| V1b.2 | `archive rerun` on a fixture archive with no `cache/` against a local server: correct bytes → `reproduced`, exit 0; a `followed_link` citation re-fetched from its `copy_url`, not its `url`; a `VARIANT` that moved → `moved`, exit 0; a `REPRODUCIBLE` that differs → exit 1; changed bytes → refused naming url and both shas, exit 1; 404 → exit 1; each act line names its inputs verified. `archive.Write` → `archive.Extract` round-trips a `./run.sh` bundle that then reproduces; C10 finds `tar.NewReader` only in `internal/archive`; `archive rerun --help` names operator privileges and network | any outcome, exit or line differs |
| V1b.3 | Real data: V1a.9 on a question with a CSV source; extract its archive to a clean scratch dir, `feov-migrate --from … --to …`, `feov-record archive rerun` | the summary's count of computations with an input re-fetched, verified and re-run is below 1; or a non-zero exit |
| V1b.4 | Rights: a fetch summary of a TDM-reserved page prints the note stating computation is accepted and its output committed; each C4 comment carries the decision | the note or a carrier says otherwise |
| V1c.1 | Assembly goldens for each `Show`; report.html carries `data:image/` and no external `src`; delete each `Show` case → a test fails | a block missing, misnumbered, or external |
| V1c.2 | `edit` and `ingest` adding `![x](y)` → refused; `![x](y)` in a mint, fix, inquiry and ruling text → a link in report.html and escaped in report.md, no `<img>`; the same text naming a PLACED store path → still a link; delete the `![` escape → the forged case embeds and fails; all 30 archives still migrate with zero refusals | accepted, embedded, or a migration refusal |
| V1c.3 | `show report --assembled --json` holds the table after the anchor's paragraph; bare `show report` holds none; the `red-lens-voice` agent golden holds the assembled-read and quote-the-sentence duties | either differs |
| V1c.4 | Smoke as V1b.3 on a question whose answer is a table and a figure; read report.md and report.html | no shown block |
| V1c.5 | Keyed retry with a different `--show` → refused; `out/` with 0 or 2 files, a CSV cell holding a newline, an undecodable PNG → refused | any is recorded |
| V2.1 | A fixture PDF page with a red rectangle at known points: a box drawn tight on `render-page`'s image → the crop is colour (red channel dominant) while `render-page`'s image stays single-channel with an unchanged sha; re-derive at a different resolution → the crop covers the same page area (edges within one pixel after scaling); the event holds the box and `crop_image_w`/`_h`; `verify` reads crop and colour page | grayscale, a different area, or a changed `render-page` sha |
| V2.2 | Cache entry deleted: report.md, report.html, `show report --assembled` and archdump each hold the placeholder — subject terms only: figure number, source title and page, "image not held in this copy" — and run.md's log names the absent crop and `archive rerun`; `run-archive/` holds no crop (`tar tzf`). Delete the placeholder branch → a test fails | silent omission, or a committed crop |
| V2.3 | Smoke: a question with a PDF source carrying a colour figure; read report.html | no shown crop, or a gray one |
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
| P-norm | Absolute paths in output | each copy's canonical root → `<bundle>` before comparing and recording, at deposit and every re-run | Record raw; such programs stay `VARIANT` |
| R-rights | Computing over cited bodies | stated and accepted, as for citations; every C4 carrier, the seat note included, states it; §IV R8 grades every channel | Consult `TDMReserved` |
| R-input | What `--input` carries | raw bytes only | Extracted text |
| R-size | Limit on output and `out/` | unbounded, stated | A cap |
| Re-run (2026-10-24) | How many re-runs a bundle gets | ONE standing re-run per bundle, its raw result stored once; each act's verdict is its own comparison against it; disagreeing bases surfaced (`bundle_bases`) | A re-run per act; one verdict per bundle |
| Crop (lead, delegated by gblock) | How a crop region is named | a pixel box on the image the seat viewed, with that image's width and height; re-derivation scales; the crop is returned to the seat | Points; a DPI field |
| Colour (2026-10-24) | Colour of shown crops | colour, through PDFium's RGBA raster; the lens compares against a colour render; grayscale OCR renders unchanged | Grayscale crops |
| CI (2026-10-24) | Where archived programs run | a job with no write credential (nightly `feov-record` leg); `--help` states operator privileges; §IV R12 | The release job |
| Per-act (lead, disclosed) | Store key for an act's artifacts | `dir/` by `bundle_sha`; output and `out/` under `acts/<K-id>/`, written once, before the event | One store per sha |
| Crop-absent (lead, disclosed) | Assembly when a shown crop is not cached | a loud placeholder in subject terms; the operator pointer to `archive rerun` goes to run.md, never report.md | Omit the figure; ~~an operator command in report.md~~ (report-voice ruling) |
| Identity (lead, disclosed) | Mode and framing in the listing | no mode (the tool sets argv[0] executable); length-prefixed fields; `__pycache__`/`.pyc` refused by name; `meta.json` not migrated | A mode field; refuse newlines; exclude caches silently |
| Composed input (lead, disclosed) | `--input` of bytes the tool composed (`recovered`/`named_backend`, no `copy_url`) | refused: no document exists to re-fetch | Accept as residue |
| Images (lead, disclosed) | What may open an image | assembly only: composed text escapes `![` as it escapes `[^` | Key on the destination string (forgeable) |
