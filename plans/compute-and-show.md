# Compute and show — what we computed, what a source shows, and how either is seen

> STATUS: approved 2026-10-25 by gblock after seven audit rounds; remaining local gaps carried as
> required tests (§V carried-tests table). Not started. Written against `0f81e2db` (code as
> `a15aaaf7`: `EventSchema` 43; `feov-migrate` is its own binary; 30 archived runs). Forks and
> rulings by gblock; design gaps resolved by the lead are marked (Forks).
> Ledgers under `~/.claude/scratch/compute-plan/`: `audit-r7/probe*` (round-7 probes),
> `audit-r6/probe*-out.txt` (round-6 renderer
> probes), `round6/census/*.txt` (sweeps),
> `round6/archscan/*.txt` (archive scans), `round6/resize-numbers.txt` (vision sizing),
> `round6/probe/probe-out.txt` (renderer probe); `round2/census.tsv` (Go meanings),
> `round3/c3-ledger.tsv` (non-Go), `round3/c10-*.txt` (archive text).
>
> AMENDED 2026-10-30 (#1657): Part 1a shipped (PR #1644, epoch 49; main `a31a7b06`, epoch 50).
> "One anchor lifecycle" is inserted before 1b on gblock's rulings of 2026-10-30 (Forks:
> Lifecycle). Its forks L1–L6 and the `source` demand were ruled by gblock on 2026-10-31 (Forks:
> L1–L6, Source demand): L1–L4 and L6 as the lead recommended, L5 reversed (`reproduce` records the
> sentence it judged). Plan-audit round 1 (FAIL, `~/.claude/scratch/anchor-lifecycle/audit-r1/VERDICT.md`)
> was followed by four more rulings the same day (Forks: Split, Staleness, L2 history, L1×L2): the
> amendment is two parts, 1a′ (the CEILING fix, epoch E1′) and 1a″ (one failed-check carrier and the
> judged sentence, epoch E1″), each its own pull request. This revision answers round 1's gaps and
> notes. 1b, 1c and 2 are amended where they assumed two
> lifecycles (Parts 1a′ and 1a″ dependents). Ledgers are under `~/.claude/scratch/anchor-lifecycle/`
> (`census-verify.tsv`: the 1a-branch census re-verified on main; `audit-r1/`: the auditor's sweeps;
> `rev-r1/`: this revision's sweeps, the 33 migrated archives and `l4census.txt`) and
> `~/.claude/scratch/compute-1a/orphan/` (`divergence.md`, `timeline.txt`, `gaps.txt`).

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
footnote (a citation's footnote is its title and page; its `--reason` never prints, `cli/blue/cite.go:191–195`).
Every seat, program or source string inside a shown block is composed as literal text
(`report.Literal`). A shown block leaves the report as every evidence anchor does: `retire --anchor` takes the anchor
out and keeps its sentence, which an edit may then change or cut. No new lifecycle: Parts 1a′ and 1a″ make
the existing one a single lifecycle, which every evidence kind lives by (`anchor.Backs`).

**Parts (R-split).** Each ships alone, with its own epoch and `migrate` translation.

| Part | Scope | Depends on | Epoch |
|---|---|---|---|
| 1a | `prove` → `compute`; directory + argv; two-value basis; per-act store; one re-run per bundle, judged per act; `K-` anchors | — | E1 |
| 1a′ | One anchor lifecycle, the CEILING fix. `retire --anchor` alone and the bare-edit refusal apply to every evidence kind. A retired anchor is dropped from every document; `[^K<n>]` numbered per document. `source` and `computation` gaps close only on a live answer (`cite --answers`), under every closing disposition, enforced at the record write for `close` and `carry`. A closure resting on a retired answer falls, and every closure-history reader re-derives it. A `K-` is a claim. A citation a live computation names cannot be retired | 1a | E1′ |
| 1a″ | One failed-check carrier: a refuted citation is a failed check, as an unsound computation is; contradictions are corroborations'. `reproduce` records the sentence it judged; evidence.md and the consistency oracle surface a check of an earlier wording | 1a′ | E1″ |
| 1b | `--input` (raw cited bytes, read-only); package `archive`; operator `archive rerun` | 1a, 1a′ | E2 |
| 1c | `--show table\|image\|text` on `compute`; images exist only as shown blocks | 1a, #1569 (render-time allowlist) | E3 |
| 2 | `--show image\|text` on `cite`: a colour crop of a box drawn on a view the model processes unresized | 1b (`archive rerun` re-derives crops), 1c (placement, `Show`, `Literal`) | E4 |

E1–E4 are the next free epochs when each part lands (main was 43). 1a landed at 49. Main is 50 at
this amendment; E1′ and E1″ are the next free epochs when 1a′ and 1a″ land.

**Success criteria.**

| # | Part | Criterion | §V |
|---|---|---|---|
| S1 | 1a | `feov-record blue compute --dir D --quote S --reason R -- python3 main.py` records a computation: bundle = every regular file under `D`; command = the argv after `--`; toolchain = resolved path + sha256 of `python3`. `blue prove` is an unknown command | V1a.1 |
| S2 | 1a | Two runs in fresh copies outside the run tree, each copy's canonical root replaced by `<bundle>` in the output. Equal output, exit and `out/` → `REPRODUCIBLE`; different with no `--variance` → refused naming what moved; with `--variance` → `VARIANT` + drift; `--variance` over agreeing runs → refused | V1a.2, V1a.3 |
| S3 | 1a | A recorded act never changes after write. Every act's verdict compares the bundle's STANDING re-run — the latest `ran` reproduce, its result at `rerun/<its rerun_sha>/` — against the act's OWN output, exit and `out/`, whatever refusal, crash or concurrency preceded it; capture's audit and `archive rerun` execute fresh and neither read nor write the store | V1a.4, V1a.5 |
| S4 | 1a | A keyed retry restates every recorded flag or is refused; program, argv and inputs are in the sha | V1a.6 |
| S5 | 1a | Every archive migrates with zero refusals; III.1.6's assertions hold; `archive rerun` over every migrated archive gives each of the 159 stored scripts the outcome `archiverun` measured on the CI image and committed | V1a.7 |
| S11 | 1a′ | `retire --anchor <id>` alone takes any evidence anchor out of the report and keeps its sentence. An edit leaving one bare is refused. A retired anchor appears in no document, and `verify` and `reproduce` refuse it. Every surface naming a retire names the anchor's kind. Footnotes are numbered per document over live anchors. A sentence's `K-` counts one claim | V1a′.1, V1a′.5, V1a′.8 |
| S12 | 1a′ | A `Close` — red's `close` or the chair's `carry` — on a `source` or `computation` gap is refused at the record write unless a live anchor of the demanded kind answers it (`cite --answers`, `compute --answers`) or, for a gap about an anchor, that anchor is retired (L1), whatever the disposition, `moot`, `not_a_defect` and a close to a successor included. Retiring that answer leaves an open gap `awaiting_evidence`. A red closure falls at the retire that leaves none of the anchors answering it at the close live, unless L1 met the demand at the close; it stays fallen until a new answer and a new close; a bench closure stands. After a fall no reader counts the gap closed: the board, the work list, `board_counts`, `convergence_at` from the fall on, telemetry, the scorecard, the chart, the spot-check archive, the dashboard, the oracle's raw walk; a re-close is counted once | V1a′.2, V1a′.3 |
| S13 | 1a′ | A computation's keyed retry returns its id. `--cites` refuses a retired citation, and retiring a citation a live computation names is refused, both inside the write's transaction. Every reader of an anchor's answers reads both kinds | V1a′.4 |
| S14 | 1a′ | Every archive migrates at E1′ with zero refusals, and the flip census is committed. The 2026-10-30 CEILING scenario resolves on its own record (committed to `run-archive/`), in a scripted replay, and in a smoke | V1a′.6, V1a′.7, V1a′.9 |
| S15 | 1a″ | One failed-check carrier: a standing `refutes` or `absent` reading of a citation, like an `unsound` judgement of a computation, is a failed check carried by the reader's gap about the anchor; it opens no contradiction. `assembly-screen` fails on a failed live anchor of either kind. The 5 archived findings raising an anchored contradiction migrate as findings about the citation with zero refusals | V1a″.1, V1a″.3, V1a″.4 |
| S16 | 1a″ | `reproduce` records the sentence it judged (`Reproduce.claim`), read off the report by the tool as `verify` records `Verify.claim`. evidence.md shows, for every check of either kind, the sentence it was of when that is not the anchor's sentence now. A standing check of a sentence the report no longer carries is a consistency violation the release fuzz reports | V1a″.2, V1a″.6 |
| S6 | 1b | `--input C-…` materialises the cached bytes read-only at `sources/C-…`; the archive commits no source bytes; `archive rerun` on an extracted archive with no `cache/` re-fetches from where the bytes came, verifies sha256 and re-runs, with 1b's exit semantics | V1b.1–3 |
| S7 | 1c | `--show table` renders the CSV under "*Table n.* R"; `--show image` embeds PNG/SVG under "*Figure n.* R" in report.md and (as `data:` URI) report.html; `--show text` inlines markdown refused where `edit` refuses it | V1c.1 |
| S8 | 1c | Raw HTML in either output is #1569's: only the tool's shapes become HTML, decided on the whole shipped document's parse; report.md is written through the same decision. An image opens only from a shown block: `edit` and `ingest` refuse an image outside code in authored text, and composed seat text's `!` is escaped; no seat, program or source text inside a shown block opens an image, raw HTML or a weave-label footnote, or breaks the block (a `--show text` block is report text: a footnote of another label is its to write); every escape and refusal that skips code reads code as the renderer's parser does, and no unit leaves a fence open across its boundary | V1c.2, V1c.6 |
| S9 | 1c | Shown blocks appear in `show report --assembled --json`; the bare read carries the anchor only and says so | V1c.3 |
| S10 | 2 | `cite --show image --page N` returns a colour VIEW the model processes unresized (`ResizedSize(w, h)` = `(w, h)`, §II; its width and height read from the PNG); `--region x,y,w,h` in that view's pixels records the citation with the crop's sha256 and returns a colour crop of the same page area from the full-resolution raster; a keyed retry restating another page or box is a Conflict; the evidence lens verifies it against the view; `cite --show text` displays the source span; no crop is committed; an absent crop assembles as a loud placeholder | V2.1–3 |

**Value case — what goes away.**

| Removed | Replaced by |
|---|---|
| The extension→interpreter table, its Windows extra and `argvFor` (`interpreter.go`) | the argv after `--` |
| The grading run from the store (`executeStored`, `proof.go:202`) | two runs, each in a fresh copy at a different path |
| Three bases as free strings (`proof.go:72–74`) | a two-value enum |
| Silent `observed` for output that moved (`proof.go:157`) | refusal, or a declared invariant — the tool never infers |
| Working directory = the run tree (`proof.go:261`) | working directory = a copy of the bundle |
| One `output` per sha, overwritten by the next act (`proof.go:364`) | one output per act, written once |
| One re-run's verdict standing for every act on a script | one standing re-run per bundle, compared against each act's own output |
| Two lifecycles (`retire.go:191`, `edit.go:268`): a computation leaves only with its sentence, and a retired one stays in evidence.md "anchored to nothing" (1a′) | one lifecycle read off `anchor.Backs`; kind differences in two tables |
| A closure on a retired answer, and `source` enforced nowhere (`estoppel.go:280`) | a live answer of the demanded kind |

**Not in scope.** Interactive charts (`--show html`, a sandboxed frame): #1574, building on Part 1c and #1569.
Rich media (audio, video). A machine-checked basis (Lean): the argv holds
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
  `requireThisEpoch` (`recordsql/store.go`), pointing to `feov-migrate`.
- Execution posture: a seat could already run anything through Bash; `compute` adds recording, not
  reach; `archive rerun` is new reach for the OPERATOR (R12). No sandbox (`nonet` is test-only).
- Fetch cache: `fetchcache.Read` :382 by sha, `fetchcache.Resolve` :526 by URL; a cache entry
  carries `TDMReserved` (`fetchcache.go:204`), held in the run's fetch index (`cache/`, never
  archived); the rights premise is stated at `fetchcache/shell.go:213–230` and
  `cli/fetchsummary.go:110–112, :345–348, :354–357`. Page rendering (`fetchcache/pagerender.go`):
  resolution per page and per axis, fractional, from the scan (`RenderDPIFor` :75, floor 300
  `tessocr.RenderDPI`, cap 600; `PageRenderDPI` :420); `renderPagePNG` :480 rasterises with PDFium
  (`RenderPageInPixels`) and converts to grayscale (`toGray`), for three callers — `RenderPages`
  :376, `RenderOnePage` :455, the scanned-PDF OCR read `fetchcache/readscanned.go:204`; `pngSize`
  :516 reads a raster's size back from its bytes. `render-page` is a lens verb only
  (`cli/lens/command.go:32`); its result carries no width or height (`renderPageResult`,
  `cli/lens/renderpage.go:82–94`). No seat-facing surface takes image coordinates today (W2, III.9).
- Vision ([vision-coordinates](https://platform.claude.com/docs/en/build-with-claude/vision-coordinates),
  read 2026-10-24): the model resizes to the largest aspect-preserving size meeting BOTH edge ≤ 1568
  px and `⌈w/28⌉·⌈h/28⌉` ≤ 1568 visual tokens (standard tier; 2576 / 4784 high-resolution) and states
  coordinates in the resized image — "The token limit can also trigger a resize when neither side
  exceeds the edge limit … the most common cause of misaligned coordinates." Its reference
  `resizedSize` returns an image that fits unchanged (`1075×1520 → 924×1307`); for PDFs it directs
  rasterising pages yourself and pre-resizing. Its Go port (`round6/resize-numbers.txt`): letter at
  300 or 600 DPI → `952×1232`; `3000×2000 → 1344×896`; round 5's `1212×1568` and `1568×1045` are
  resized again. A standard-tier fit is a high-resolution fit. `resizedSize(w, h) = (w, h)` exactly
  when `fits(w, h)` (else the search returns a shorter long edge), and `fits` is integer arithmetic,
  so a SQL `CHECK` expresses it.
- Site: report.md and report.html are written at the run root (`report/docs.go:391`,
  `report/assemble.go:330`); report.html is one self-contained file rendered from the markdown
  bodies (`report/site.go:60`); composed text is escaped where it is composed (`report/notes.go:43`
  `EscapeNotes`, called at `layout.go:89`, `docs.go:129`, `:211`, `assemble.go:151`,
  `proofs.go:85`); evidence.md fences a script and its output in a fixed triple backtick
  (`proofs.go:183`, `:185`). Code is found by hand: `eachLine` (`layout.go:350–361`) toggles on any
  line opening ```` ``` ```` or `~~~` — no run length, no indented code — for its seven callers
  (`EscapeNotes`; `WovenNoteIn` :630; `composedClaims` :407, `stands` :532, `part` :658,
  `layoutview.go:102`, `:119`, which find headings), and `reCodeSpan` (`notes.go:32`) matches
  single-backtick spans; the renderer reads goldmark's tree (`md.go:42` `mdReader`; `:48`
  `lineReader`, one inline line). md.go passes an `ast.HTMLBlock` (`:160`) and an HTML-only
  paragraph (`htmlOnly`, `:138`) through raw. Each section is escaped alone while the renderer
  parses the whole document, so a fence one section leaves open flips the next one's parity
  (`audit-r6/probe-out.txt`: seat `<script>` reaches report.html; a four-backtick fence holding
  ```` ``` ```` leaves `[^K1]` open), and escaping only an HTML block's opening `<` turns it into a
  paragraph whose later `<img>` stays raw (`probe2-out.txt`). **#1569** (open, blocks 1c) fixes
  report.html on main: on the whole shipped document's parse, only the tool's shapes become HTML —
  evidence.md's `<a id>` (two inline `RawHTML` nodes in an `htmlOnly` paragraph, `proofs.go:122`),
  the board chart (`chart.go`), comments that are exactly the tool's tokens, and images whose
  destination is a placed store path; every other HTML node is escaped as text. Archived seat text holds raw HTML outside code (quadratic-formula `verify.text`
  row 683, `<img …>` inline); authored text, none (`round6/archscan/authored-html.txt`).
- SVG: an SVG shown as an image (markdown `![]()`, `<img>`) renders statically — no script, no
  external resource, no hover or zoom (to verify at implementation).
- Store writes: the fetch cache stages a sibling `.tmp-*` and renames (`fetchcache.go:462–476`).
  `migrate`'s replay renumbers `events.id` through `newID` (`record/migrate/replay.go:193`), so no
  store path is keyed by an event id.
- Run-archive tarball: written by `capture.go:2045` (`ArchiveRecord`, every entry `Mode: 0o644`
  :2159), read by two private copies — `internal/record/recordtest/archive.go:34` and
  `devcmd/archdump/main.go:149`, both via `os.Create` (umask).
- CI: the Windows pull-request leg runs FEOV's suite with `-short` (`.github/workflows/hooks.yml:693`).
  Every job holds `contents: read` (workflow-level :27–28) but `release` (:1581, `contents: write`
  :1594–1595). A scheduled run saves main-scoped Go caches that pull requests restore
  (`cache: ${{ github.event_name != 'push' }}`, e.g. `feov-record` :518); `ocr-corpus` (:315) is a
  Linux-only job with `cache: false` (:340).

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
//   rerun = 8 is THIS act's comparison: the bundle's standing re-run against the act's own output, exit, out/
//   recorded_output = 4, observed_output = 5 → reserved 4, 5 (both outputs are in the store; III.1.4)
//   + optional bool   ran        = 9;   // this act executed the bundle
//   + optional int32  rerun_exit = 10;  // that execution's exit; set only when ran
//   + optional string rerun_sha  = 12;  // the execution this act compared against — its own when ran, else the
//                                       // standing re-run's: rerun/<rerun_sha>/ (framing below)
//   + optional string drift      = 11;  // what differs between the act's recorded result and the re-run
//   checks: drift present ⇔ rerun = DIFFERS; rerun_exit present ⇒ ran and rerun_sha present
//   (one-way: migrated events carry neither rerun_exit nor rerun_sha; every migrated DIFFERS gets drift, III.1.6)
// Rerun [MODIFY] (means): REPRODUCED = "the bundle's re-run wrote this act's recorded output, exit and
//   out/, byte for byte"; DIFFERS = "… other than this act's recorded output, exit or out/; drift names
//   what"; CALLS_THE_TOOL: "bundle" for "script". RERUN_NOT_RUN = 4 → reserved 4, "RERUN_NOT_RUN":
//   `ran` = false carries "compared against the bundle's re-run, ran nothing".
// AboutKind [MODIFY]: ABOUT_KIND_PROOF = 5 → ABOUT_KIND_COMPUTATION = 5.
// Cite [MODIFY, 2]: + optional Show show = 27; optional int32 crop_page = 28;   // absent for an image source
//   (answers = 26 is 1a′'s, III.1′.3)
//   optional int32 crop_x = 29, crop_y = 30, crop_w = 31, crop_h = 32;   // the box, in VIEW pixels (--region)
//   optional int32 crop_view_w = 33, crop_view_h = 34;                  // the view's size, read from its PNG
//   optional string crop_sha256 = 35;   // the crop the seat was returned: sha256 of its DECODED pixels (Crop hash)
//   checks: show ∈ {IMAGE, TEXT}; crop_x…crop_sha256 present ⇔ IMAGE; crop_w, crop_h > 0;
//   crop_x + crop_w ≤ crop_view_w; crop_y + crop_h ≤ crop_view_h;
//   the view is one the model processes unresized — fits(crop_view_w, crop_view_h), in SQL integer arithmetic:
//   (w+27)/28*28 ≤ 1568 AND (h+27)/28*28 ≤ 1568 AND ((w+27)/28)*((h+27)/28) ≤ 1568
```

`rerun_sha` = sha256 over length-prefixed items (III.1.3's framing): `exit <n>`, `output <sha256>`,
`out <path> <sha256>` per `out/` file, sorted. The view check is the record's; the sizing function
`fetchcache.ResizedSize` (beside `pngSize`; a port of the guide's Go with its standard-tier
constants, the "vision ceiling" of `pagerender.go:38`) produces the view; V2.1 holds the two to
each other.

`inputs` is the first list with an `(sql) flag` (`grep -c 'repeated.*(sql)' record.proto` → 0):
`flagOf` (`record/supersede.go:287`) and `flagFor` (`recordpb/requiredfields.go:101`) read the flag
whatever the cardinality; schemagen's output is read in the `schema.sql` golden. Toolchain is two
scalars and the crop seven `int32`s because the generator has no message or double kind. The live
writer writes `rerun_sha` on every reproduce and `rerun_exit` with `ran` (V1a.4); migrated events
carry neither (residue). `awaiting_proof` →
`awaiting_computation`. Every `(means)` naming the proof concept (:278, :280, :344, :374, :404,
:449, :452, :455, :541, :555–558, :1127, :1130; :419's "script" is English) names a computation.
remap's field lists class every new field: `inputs` id; `invariant`, `drift` argument;
`Reproduce.drift` (migrated: program bytes, III.1.6), `argv`, `toolchain_*`, `ran`, `rerun_exit`,
`rerun_sha`, `crop_*` data;
`Reproduce.recorded_output`/`observed_output` leave the data list (`remap.go:126`). `keycensus_test.go`
`notFields` gains `proof_basis`, `proof_id`, `proof_key`, `proof_sha`, each naming its replacement,
and its `output` reason (`keycensus_test.go:85`, "the record carries recorded_output/observed_output
on Reproduce") is restated: an act's output is in the store, `acts/<K-id>/output`.

```
<runDir>/
├── computations/<bundle_sha>/
│   ├── dir/                the bundle — content-addressed, written once
│   ├── acts/<K-id>/        one per recorded act: output (each copy's root reads <bundle>), out/
│   ├── rerun/<rerun_sha>/  content-addressed: one `ran` execution's output, out/
│   └── .tmp-*/             staging for all three; never read, never archived
└── cache/                  fetched bytes; Part 2's views, keyed (source sha, page), and crops at
                            crops/<crop_sha256>.png, named by pixel sha (staged and renamed as
                            below) — never archived
<tmp>/…/<root>/             each execution's copy, outside the run tree, two at different depths:
                            dir/ contents + empty out/ + sources/C-… (1b, read-only)
```

**Store writes — one mechanism** (precedent `fetchcache.go:462`): stage into
`computations/<sha>/.tmp-*` (same filesystem), then `os.Rename`. Every reader resolves a directory
from the record — `acts/` by `computation_id`, `rerun/` by a reproduce's own `rerun_sha` —
never by listing, and `archive.Write` copies only directories the record names.

| Directory | Promoted | Order vs event | Target exists | Crash or refusal leaves | Read by |
|---|---|---|---|---|---|
| `dir/` | rename `.tmp-*` → `dir/` | before | rename fails → the present `dir/` must hash to `bundle_sha` (else refused); staging removed (two lanes on one bundle) | a `.tmp-*` | every execution |
| `acts/<K-id>/` | rename | before `AppendPlaced`; a refused append removes it | refused (random id) | an `acts/` no event names | `compare`, evidence, assembly |
| `rerun/<rerun_sha>/` | rename; name = sha256 of its own content | before the `ran` append | rename fails → the present directory holds the same bytes by construction; staging removed | refusal or crash before the append: a directory no event names (ignored, never archived); no event ever lacks its directory | a `ran` = false `reproduce`; evidence.md (III.1.4) |

Leftovers are ignored, not swept: a sweep races a concurrent writer's staging, and no reader or
archive reaches them. Every store name is a content hash or a `computation_id`, so a `migrate`
that renumbers `events.id` renames nothing.

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
| Recorded flags restate on a keyed retry (`seat.RepeatsKeyed`, `cli/seat/keyedretry.go:31`) | asserted set (`proveAsserted`, `prove.go:191`) + `--variance` → `invariant` differs; a different sha under the key → Conflict (`prove.go:84`). Class gate (W6): `seat.AssertedCovers` test — for each flag of the verb, a distinct value through the verb's own field writer (`proofFields` → `computeFields`, `prove.go:180`; `citeFields`, `cite.go:206`); every field that moves is `key` or asserted. `mint`, `finding`, `edit` have no field writer to drive (`mint.go:391`, `finding.go:153`, `edit.go:264`) and keep their own retry tests |
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
5. Write per III.0's store table: `dir/`; mint the `K-` id (`record.NewID`, random,
   `record/id.go:22`); `acts/<K-id>/` from the first run; then `AppendPlaced`. A keyed retry after a
   crash mints a fresh id; an idempotent keyed retry writes nothing.
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

#### III.1.4 Re-run — one standing re-run per bundle, judged per act [MODIFY] (ruling Re-run; Re-run store)

`compute.Rerun(run, bundle_sha)` executes the bundle ONCE — III.1.2 steps 1–3 and 6 with a single
copy — and returns that execution's output, exit and `out/` in scratch; it writes nothing to the
store. That one result is what `compare` reads and, for a `ran` reproduce, what is stored.
`compare(act, result)` sets the act's verdict from its own `acts/<K-id>/` — output, exit, `out/`;
an absent recorded `out/` equals an empty one; on a difference it writes `drift`. On `VARIANT` a
differing re-run is expected: the lens's reason says whether the output satisfies the invariant (F2).

| Caller | Executes | Compares against | Store |
|---|---|---|---|
| `reproduce`, bundle has a standing re-run S | nothing; `ran` = false; records S's `rerun_sha` | `rerun/<S.rerun_sha>/` and S's `rerun_exit` | writes nothing; the append is refused when S is no longer the standing re-run (re-read inside the write) |
| `reproduce`, none (never run, or only migrated runs, which carry no `rerun_sha`) | `compute.Rerun`; `ran` = true | its own execution | stage → `compare` → `refuseSound` (refused: staging removed, nothing recorded) → rename to `rerun/<rerun_sha>/` → append |
| capture's audit, `archive rerun` | `compute.Rerun`, always | its own execution | reads `dir/` and `acts/` only; never reads `rerun/`, writes nothing |

S is the LATEST `ran` reproduce on the bundle carrying `rerun_sha` (the rule `scriptRerunsOf`
keeps). Every write renames before it appends. Two concurrent reproduces with no standing re-run
both execute, rename and append: two directories, two events, the later stands; each `ran` act's
verdict is its own execution's. A reader between S's rename and S's append finds no standing event
and executes itself: one extra execution, never a comparison against another result.

**What an act displays.** Its own recorded output (`acts/<K-id>/output`) and the output of the
execution its comparison read (`rerun/<its rerun_sha>/output`) — never another act's. `Reproduce`
carries no copy of either (`recorded_output`/`observed_output` reserved, III.0).

| Carrier (sweep W1, W7, III.9) | Becomes |
|---|---|
| `record/proofverdict.go` header :11–23, `scriptRerun` :41–60 (`recorded`/`observed` :57 deleted), `ranTheScript` :63 | "THE RE-RUN belongs to the BUNDLE — the latest `ran` reproduce carrying `rerun_sha` stands; THE COMPARISON belongs to the ACT; SOUND belongs to the ANCHOR"; `bundleRerunsOf` + `actVerdictsOf` (latest reproduce per `computation_id`) |
| `validateReproduce` :70–100 (called on every append and by migrate's replay, `record.go:1007`, `replay.go:53`), `proofVerdict` :111, `proofVerdictOf` :137 | `ran` = false needs a `ran` reproduce on the bundle; its `rerun_sha`, when present, must be the standing re-run's (else refused: the standing moved); against a standing `CALLS_THE_TOOL` it records `rerun` = `CALLS_THE_TOOL`, no drift (no output to compare), and `sound` is refused; the verdict read is the act's. Migrated `ran` = false carry no `rerun_sha` and stand on the bundle's migrated `ran` event (residue) |
| `cli/lens/reproduce.go` :93–123 (comment + flow; the `RecordedOutput`/`ObservedOutput` writes :126–135 deleted), :144 (flag help) | the table above; help: "the tool executes the bundle unless it already has a standing re-run; either way it compares THIS act's own output, exit and out/" |
| `reproduce.go` `rerunOutcome`/`refuseSound` :151–200 | judges `sound` on THIS act's comparison, not the standing re-run's own verdict; the refusal names the act's drift |
| `reproduce.go` result :218–234, `Human()` :240–258 | `standing_rerun` removed; `rerun_by` names the seat whose execution was compared; Human states this act's comparison ("compared against the bundle's re-run by S: this act's output REPRODUCED / DIFFERS — drift") — never "stands for every proof" |
| `cli/seat/verbs.go:290–292` (`show evidence` help) | "`rerun` is THIS ACT's comparison against its bundle's standing re-run … `not_compared`: the bundle was re-run and nobody compared this act; `not_rerun`: the bundle was never re-run" |
| `record/evidenceview.go` :54–55, :162–164, :198–199, :215–216, :452–518 | per-act `rerun` ∈ {its comparison, `not_compared`, `not_rerun`}; `bundle_bases` on a row whose bundle backs acts of differing basis |
| `record/citationid.go` :310–311, :340–352, :388–390 (`RecordedProofs`, `ProofAt`) | per-act `Rerun` from `actVerdictsOf`, plus the standing re-run's seat |
| `record/available.go` :508–524, :574–596 (`proofChecksOwed`, `proofOwed.item`) | the act's own state; the item says whether the bundle has a standing re-run (then `reproduce` executes nothing and compares) |
| `record/passblockers.go:282`, `record/reopened.go:12` | "…which executes nothing once the bundle has a standing re-run, and compares this act against it" |
| `report/proofs.go:134–153` (evidence.md) | three states, comment :138 restated: "not re-run — no checker executed this bundle"; "not compared — the bundle was re-run and nobody compared this act's output"; the comparison (reproduced, differs + drift, calls the tool). "recorded / on re-run" (:151–152, today the standing event's copies) → this act's `acts/<K-id>/output` and `rerun/<its rerun_sha>/output`, each fenced (III.1.5's rule); a migrated act shows its own output and the drift |
| `capture/proofrerun.go` :61–110, :120–185 (check `proof-rerun` → `computation-rerun`) | one row per bundle holding every act (id, basis); one `compute.Rerun` per sampled bundle; each act graded by its own basis — `REPRODUCIBLE` differing FAIL, `VARIANT` moved; the detail names any bundle whose acts' bases disagree |
| `record/migrate/replay.go:506–538` (`reproduceAnchors`) | anchors after the first get `ran` = false (writes `RERUN_NOT_RUN` today, a compile error once 4 is reserved); the output clearing :532 goes with the fields |
| record.proto `Rerun` comment :545–548 ("which every proof anchor on that script shares"), :555–558, Reproduce comment :1441–1457; `help/reproduce.md` :7 :9 :15 :23 :25 | III.0; the bundle's standing re-run is executed once; the verdict is this act's comparison |
| debate.js :1008 ("RE-RUN IT ONCE — the stored script is what the re-run executes") and the simulator pin `tests/simulator/debate.test.mjs:416` | "reproduce it — the tool executes the bundle once and compares this act's own output"; the pin follows |

#### III.1.5 Rename sweep [MODIFY] — names only; III.1.4 carries the logic

| Site | Becomes |
|---|---|
| anchor kind `proof`/`P-`/tag `proof` (`anchor.go:63`) | `computation`/`K-`/tag `computation` |
| `proofNote` `[^P%d]` (`layout.go:591`), `wovenNote` `(?i)\[\^p?[0-9]+\]`, `WovenNotesRule` "P and digits" | `[^K%d]`, `k?`, "K and digits" |
| `weaveProofRefs` (`report/proofs.go:49`), evidence.md `## Proofs` | `weaveComputationRefs`, `## Computations` (argv, toolchain, basis, invariant, drift; each bundle file fenced, its language from `fenceLang`'s table (`proofs.go:231`), or one line — path, size, sha256 — when it is not UTF-8 or holds NUL; output and both drifts (`Computation.drift`, `Reproduce.drift`) fenced; every fence a backtick run longer than any in its body, replacing the fixed ``` at `proofs.go:183`, `:185`; the preamble `proofs.go:111–115` (`observed`, "rather than a proof", `proofs/<sha256>/`) restated for `REPRODUCIBLE`/`VARIANT` and `computations/<bundle_sha>/`). No seat- or program-chosen name enters markup raw: argv, the toolchain path and every bundle and `out/` path print through `report.Literal` (lands in 1a for this; 1c reuses it) |
| `awaiting_proof` (`recordsql/views.go:566`), `ProofAnswers` (`estoppel.go:258`), `proofChecksOwed`/`proofOwnerOf` (`available.go:522`/`:557`), "unjudged proof" (`retirement.go:161`), `proofs_unverified` (`evidenceview.go:329`), `unsoundProof` (`record/unsoundproof.go`), `BlockerUnsoundProof` "unsound_proof" (`passblockers.go:36`) | `…computation…` |
| `ABOUT_KIND_PROOF` (`about.go:70`, `mintbudget.go:393`, `:409`, `unsoundproof.go:45`, `:81`, `refs.go:126`; help `cli/lens/finding.go:144–145`) | `ABOUT_KIND_COMPUTATION`, `--about-kind computation` |
| `EvidenceProofAtJSON` `script`/`script_file`/`recorded_output` (`evidenceview.go:598`, `:614–616`, filled `cli/seat/verbs.go:725`; help `:297`) | `files` (path, body) and `output` (the act's own `acts/<K-id>/output`) |
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
| `Reproduce` that ran (`reproduced`, `differs`, `calls_the_tool`) / `not_run` | `ran` = true / `ran` = false with `rerun` = its bundle's standing re-run's value, and its `drift` when that is `DIFFERS` (so the two-way check holds; no archive holds the case — a synthetic one tests it): migrated acts on one sha hold one output, so that is each act's comparison |
| `Reproduce.recorded_output` / `observed_output` (13 events, all `differs`: quadratic-formula 6, is-91-prime-b 2, m20 5; era 0 — `round6/archscan/reproduce-outputs.txt`) | `recorded_output` dropped: its bytes are the act's `acts/<K-id>/output`, whole; `observed_output` → `drift` = "on re-run the program wrote: " + its bytes, set after id respelling (program bytes are not respelled). Every archived `differs` carries one, so every migrated `DIFFERS` has `drift` and the two-way check holds |
| `about_kind` `proof` | `computation` |
| `P-`/`p-` + 8 hex | `K-` + the same hex: `hexID` (:78) and `archivedToken` (:82) admit `P-` — today `currentID` (:76) matches `P-` and `id()` (:209) would otherwise hash a new hex |
| `<!--proof:…-->`; `[^Pn]` in argument fields | `<!--computation:K-…-->`; `[^Kn]` |
| `computations/<sha>/rerun/<rerun_sha>/` (a run written at E1 or later, migrated forward) | copied unchanged — content-addressed, no event id in the path; `.tmp-*` and directories the record does not name are not copied |

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
| store paths inside program output | 9 migrated `drift`s (from `observed_output`), 1 stored `output` | bytes the program wrote |
| acts sharing a sha receive the sha's one stored `output` | 110 acts on 27 shas, 20 runs | all recorded `reproducible`: the one output is each act's by its basis |
| scripts that reproduce neither from the store nor a copy | 50 of 159, 9 runs (17 quadratic-formula, 21 record-store-authority) | 48 migrate to `REPRODUCIBLE` and `archive rerun` prints `differs`, exit 1; 2 (`62d048c0`, `ab6aa892`) are `VARIANT` → `moved`. They were graded from the run tree; the committed outcomes (V1a.7) pin each |
| migrated reproduces hold no raw result (no `rerun/`, `rerun_exit` or `rerun_sha`); a migrated `ran` = false stands on its bundle's migrated `ran` event, and a migrated `drift` is ≤ 400 bytes (`truncateOutput`, `reproduce.go:207`) | every migrated reproduce | not recorded then; a migrated `ran` event is no standing re-run (no `rerun_sha`), so the next live `reproduce` on such a bundle executes it, once |
| `proofs/synth/` | 17 drafts, 1 run | each is also stored under its sha |
| author script path (`Proof.script`) | every proof | the argv replaces it |

#### III.1.7 Agent-facing text [MODIFY] (contract: prompts and constitutions)

Sources only, then `agentgen`, `enginegen`, `vocabdoc`: `help/prove.md` → `help/compute.md`;
`help/reproduce.md` (III.1.4); `help/{dispatch,close,mint,edit}.md`;
`scripts/agentgen/src/{blue-researcher,blue-synthesizer,red-lens-computation,red-lens-evidence}.md`,
`fragments/{blue-search-and-cite,lens-audit-and-mint,lens-duties}.md`, both preloaded-skill sources;
`skills/research/SKILL.md:7`; debate.js authored :229 :545 (comment) :933 :978 :1008 :1166 (:1007
"reproducible measurement" is English), generated :395 and `unsound_proof` :417 (the envelope block
:370–433, from `record.Plan` and the enums); `references/report_template.md:147`,
`references/run-directory.md` (:28 "with script", :77); terms.json (III.1.5); FEOV docs (ledger
§C3); `docs/local-development.md:226`; `CLAUDE.md:34`; `scripts/universe-stream.py:24`. Seat prompts
name the act — "compute it"; the prompt gate matches `<role> <verb>` (`seatprobe/naming.go:78`).
The `compute --help` page teaches: a directory and a command; output records `<bundle>` for the
copy's path; make output deterministic or state the invariant; a helper the program executes runs
through its interpreter (only argv[0] is made executable); a non-zero exit both runs agree on is
recorded; output and `out/` are unbounded and committed.

### Part 1a′ — one anchor lifecycle (epoch E1′) (ruling Lifecycle; #1657)

**Why.** The Part 1a smoke (2026-10-30_mersenne-31-prime-sum, one lane, `--smoke`) ended CEILING on
gaps nothing could close. Blue reworded a sentence and moved its `<!--computation:K-…-->` to a bare
position — refused for a `C-`, allowed for a `K-` — then retired the claim with `--quote`. The
computation stayed in evidence.md "anchored to nothing". Lenses minted six gaps against the stale
entries, and `close` refused the three `computation` ones: `requireComputationAnswered` asks only
whether any computation names the gap, retired or not, and a gap about the retired anchor had none
(`~/.claude/scratch/compute-1a/orphan/timeline.txt`, `gaps.txt`). The cause is that one lifecycle is
claimed and two are built. `anchor.go:6–9` says "every kind lives the same way"; `retire.go:191` and
`edit.go:268` implement two lifecycles. The census (`~/.claude/scratch/compute-1a/orphan/divergence.md`,
taken on the 1a branch, re-verified on main at `a31a7b06` in
`~/.claude/scratch/anchor-lifecycle/census-verify.tsv`: 100 of 120 refs at the same line, 20
moved, none changed in meaning). In its sections A–J it classes 36 table rows LIFECYCLE, wholly or in
part. Parts 1a′ and 1a″ remove them before 1b, so that §I's placement rule, and every shown block in
1c and 2, rests on one lifecycle.

**The rulings (gblock, 2026-10-30, binding).** Citations and computations are the same thing: "a
textual statement about an external source or reference material. Even a reference to a figure in
text is a citation." They "behave exactly like citations". A retired anchor of either kind is
dropped from evidence.md, and the record keeps it. A gap whose check kind is `source` or
`computation` closes only on a LIVE answer: `cite` gains `--answers G-…`, a non-retired anchor of
the demanded kind must answer the gap, and retiring that anchor reopens the debt. Retire never
auto-closes a gap. The tool refuses; it never infers.

**The rulings (gblock, 2026-10-31, binding; Forks L1–L6, Source demand).**

- **L1.** A gap about an evidence anchor (`mint --about-kind anchor`, the failed check's carrier)
  under check kind `source` or `computation` has its live-answer demand met when that anchor is
  retired, read from two fields, `mint.about_ref` and `retire.anchors`. Red still closes it.
- **L2.** A closure of a demand gap falls, derived, once no anchor that answered the gap when the
  close was recorded is live. The gap is then `open` and `awaiting_evidence` until a new answer AND a
  new close; nothing re-closes it on its own. A bench closure, and a closure that rested on no
  answer, stand.
- **L3.** A `K-` counts as a claim exactly as a `C-` does: one per distinct anchor per sentence.
  `CountsAsClaim` reads `backs`, and the `claim` column of `anchor.kinds` is deleted.
- **L4** (Part 1a″). A standing `refutes` or `absent` reading of a citation is a failed check,
  carried by the reader's gap about the anchor, as an `unsound` computation is. A contradiction keeps
  only red's corroborations. migrate translates the archived raise chains on anchored readings, from
  counts measured before the translation is written.
- **L5** (Part 1a″). `reproduce` records the sentence it judged (`Reproduce.claim`), exactly as
  `verify` records `Verify.claim`, and the K- surfaces read it as their citation counterparts do.
- **L6.** Retiring a citation that a live computation names (`--cites`; 1b's `--input`) is
  refused, and the refusal names those computations.
- **Source demand.** `source` is a hard live-answer demand on every closing disposition, `moot`
  and `not_a_defect` included, exactly as `computation` is today (`close.go:69–72` exempts no
  disposition).

**The rulings after plan-audit round 1 (gblock, 2026-10-31, binding; Forks Split, Staleness, L2
history, L1×L2).**

- **Split.** Part 1a′ is the CEILING fix: retire for every evidence kind, the bare-marker refusal,
  the closure demand and live answer (`cite --answers` included), render after retire, L1, L2, L3,
  L6 and the `source` demand. Part 1a″ is L4 (the failed-check carrier and its archive translation)
  and L5 (`Reproduce.claim` and the staleness oracle rule). Each is its own pull request at its own
  epoch; 1a″ builds on 1a′.
- **Staleness.** The consistency oracle gains the staleness rule over the evidence table (1a″, with
  L5).
- **L2 history.** Every closure-history reader is re-derived: `convergence_at` per point,
  `HasClosed`/`ClosedEpoch`/`Closure`, the scorecard, telemetry, the board chart, the spot-check
  archive, the dashboard, the repair denominator, the board JSON and the oracle's raw walk.
- **L1×L2.** A closure whose demand L1 met stands: the `about_ref` anchor's retirement is permanent,
  so the demand stays met whatever happens to the answering anchors. Only a closure that rested
  solely on answering anchors falls.

| Part | Holds | Epoch | PR |
|---|---|---|---|
| 1a′ | Phase 0; the kind rows; `retire --anchor` for every kind; the bare-marker refusal; `cite --answers`; the closure demand at the record write (`close` and `carry`), every disposition; L1; L2 and every closure-history reader re-derived; render after retire (every document, per-document `[^K<n>]`); L3; L6 inside the write transaction; `verify`/`reproduce` refuse an anchor with no live sentence; the work list's backing over both check verbs; `computation-rerun` over live computations | E1′ | its own |
| 1a″ | L4: `stateFailed`, one carrier, contradictions for corroborations only, the archived raise chains translated; L5: `Reproduce.claim`, `checkLines`, the judgement JSON's `claim`; the staleness oracle rule; blockers and owners read off the kind row; `assembly-screen` on a failed anchor of either kind | E1″ | its own, after 1a′ merges |

**Phase 0 — the fixture.** 1a′'s first commit adds the 2026-10-30 smoke's archive as
`run-archive/2026-10-30_mersenne-31-prime-sum.tar.gz`, from
`~/.claude/scratch/compute-1a/smoke/universe/src/run-archive/`. It is a real run, written by
`capture`, in `run-archive/`'s shape (`records/record.db`, `inputs/run-config.json`,
`computations/`, no `cache/`; CLAUDE.md, `docs/local-development.md:223`). 284,802 bytes, inside
the committed range (the 32 committed tarballs total 7.1 MB, the largest 1,120,483 bytes);
`*.tar.gz` is `binary` in `.gitattributes`. Its `record.db` is byte-identical to the scratch copy
(`~/.claude/scratch/compute-1a/smoke/copy/run/records/record.db`, `cmp`), at schema epoch 49, holding
17 computations, 19 reproduces, 23 verifies and 7 retires; `feov-migrate` at epoch 50 (`f5c09519`)
migrates it with exit 0 (`~/.claude/scratch/anchor-lifecycle/rev-r1/logs/`). Every reader of every
committed archive takes it up — census L15, two sweeps: the files that glob
`run-archive/*.tar.gz` and the callers of `archivedRuns` (`remaplists_test.go:96`). In
`internal/record/migrate`, twenty test files: `archivedgaps`, `argumentresidue`, `benchoccasion`,
`everyarchive`, `everyrender`, `inquirymethod`, `proserefs`, `proserewrites`,
`unmigratedarchive`, `remaplists`, `archiveonce`, `computationarchive`, `computationresidue`,
`correctionarchive`, `cutfields`, `dispatchpin`, `findingarea`, `gapclass_internal`,
`remanddirection`, `reproduceanchor`; and `devcmd/archdump`, `devcmd/archiverun` (`main.go:105`).
Per-archive tables in those tests (`computationresidue_test.go:261–286`'s `notRun` and `carried`
maps) read an absent row as 0, and the fixture holds no `not_run` and no pre-epoch-28 reproduce, so
neither gains a row. So the phase (a) runs `go test -count=1 ./internal/record/migrate/...` and
reads each per-archive line the 33rd archive adds; (b) dispatches the `archive-rerun` job
(`hooks.yml`, `workflow_dispatch`), downloads `archive-rerun-outcomes` and commits it, so
`devcmd/archiverun/testdata/outcomes.tsv` holds the archive's acts as MEASURED (V1a.7) — until then
the nightly check fails on the new acts by design. From this commit on, "every archive" in §V means
these 33.

**One mechanism, the kind row decides.** An **evidence anchor** is an anchor whose kinds-table row
has `backs: true`: `C-` and `K-` today, and Part 2's figures, which are citations. Every lifecycle
rule below reads `anchor.Backs(id)`, never `anchor.Kind(id) == "…"`. What differs between kinds is
looked up in a row:

- **`anchor.kinds`** (the vocabulary leaf, which imports no record package) holds the label, the
  token, the assembly (footnote target) and `backs`. Its `claim` column is deleted, and
  `CountsAsClaim` reads `backs` (ruling L3).
- **`recordsql.EvidenceTables`** [NEW] (`record/recordsql/evidencetables.go`) holds one row per
  backing kind, keyed by the `anchor.kinds` name. It lives in `recordsql`, not `record`: the
  `evidence_anchor` view (`recordsql/views.go`) is generated from it, `record` imports `recordsql`,
  and `recordsql` cannot import `record` (Fork Kind-row home). 1a′'s columns:
  - `table`, `idColumn`: the placing event's SQL table and id column (`cite`/`label`,
    `computation`/`computation_id`);
  - `check`: the check kind a live anchor answers (`source`, `computation`, the `enum_check_kind`
    word the SQL compares);
  - `checkTable`, `checkIDColumn`: the event that checks it (`verify`/`anchor`,
    `reproduce`/`computation_id`), which the work list's `backing` reads;
  - `noun`: "citation", "computation" — the word every refusal and result names the kind by.

  1a″ adds `owner` (replacing `anchorOwnerOf(computation bool)`, `available.go:431`) and
  `failedWord`. `TestEveryBackingKindHasAnEvidenceRow` (record) fails when a `backs` kind has no
  row, when a row names no `backs` kind, when a `check` value has no row, or when a row's
  `table`/`idColumn`/`checkTable` names no table or column of the generated DDL.

A new kind adds one row to each table. It adds no lifecycle code: the `evidence_anchor` view's UNION
arms are generated from `EvidenceTables` (III.1′.3), and every hand-written UNION over the evidence
tables reads that view instead (census L16: `record/id.go:86–87`, `record/refs.go:130`, `:204`,
`record/grounds.go:119–120`; `mintbudget.go:61`, `:64` are per-kind measures of the logic lens's
budget and stay per kind).

#### III.1′.1 The lifecycle (contract: CLI flag set, event schema, prompts)

| Stage | One rule (every evidence anchor) | Today (main `a31a7b06`) | Kind row supplies |
|---|---|---|---|
| Attach | `cite` and `compute` place the anchor at the end of `--quote` (`AppendPlaced`, `seat/verbs.go:1384`). Both take `--answers G-…`. A keyed idempotent retry returns the anchor id | `cite` has no `--answers`. `compute`'s idempotent result omits the `K-` id (`cli/blue/compute.go:117`, `computation_id` omitempty :276). A seat retrying a compute never gets back the id that `retire --anchor` needs | — |
| Edit, bare | An edit that would leave an evidence anchor bare is refused. The refusal names `retire --anchor <id> [--answers G]` as the way out (`refuseBaredCitations` → `refuseBaredEvidence`, `edit.go:261`, now on `anchor.Backs`) | `K-` may be left bare (`edit.go:268`). The smoke's orphan generator | `noun` in the message |
| Retire, sentence kept | `retire --anchor <id>` alone, for any evidence anchor: the anchor leaves, its sentence stays word for word, `claim_kept` is set, and `--answers` is optional (`retireCitation` → `retireAnchor`, `retire.go:183–216`). A finding or gap anchor alone is refused, as today, naming the claim retire | `K-` refused (`retire.go:191`), so an unsound `K-` on a correct sentence can leave only with the sentence. The help and the prompts already promise the act (III.1′.5) | `noun` |
| Retire, with the claim | `retire --quote` takes the bare anchors of a cut claim, any kind (unchanged, `retire.go:87–143`). An evidence anchor reaches it bare only from an archived record or the synthesizer's round-0 write | the same | — |
| Render after retire | A retired evidence anchor appears in no document: no woven footnote in report.md (its token is gone), no entry in evidence.md, no row in `show evidence`'s lists or counts. `show evidence --anchor <retired id>` still returns it and says `retired` (event, reason). Every surface that names a retire names it by its kind's `noun`: CHANGELOG's "Claims retired" (`withdrawnClaims`, `assemble.go:925–967`), `show changes` (`view/changes.go:194`), `retire`'s result (`retire.go:474`) | evidence.md lists every computation and calls a retired one "anchored to nothing" (`report/computations.go:133–135`). It lists every source as still cited (`report/citations.go:22–73`). The unverified counts fold differently (`evidenceview.go:563–572`). A `K-` retire would print "retired citation K-…" (`changes.go:194`, `retire.go:474`) | `noun` |
| Footnote numbering | Per document, by first appearance, over live anchors: `[^N]` and `[^K<n>]` alike. evidence.md's `#K-…` anchors carry identity across documents, so no number has to | `[^K<n>]` is run-wide record order, retired included, so a document shows `K1, K3` (`computations.go:32`) | `assembly` (target) |
| Check needs a live sentence | `verify` and `reproduce` refuse an anchor that is not in the report, is retired, or stands bare, through one function (`citedSentence` → `record.LiveSentence`, `verify.go:129–150`) | `reproduce` judges a retired or bare `K-` (`reproduce.go:68` → `ComputationAt`, `citationid.go:428`) and the ledger then ignores the verdict | — |
| Claim count | Every evidence anchor on prose counts, one claim per distinct anchor per sentence, as citations count today (`segmentLabels`, `claimcount.go:165`); `CountsAsClaim` reads `backs` (ruling L3). The logic lens's mint budget follows (`mintbudget.go:66`), and only grows, so no archived mint is refused | `K-` `claim: false` (`anchor.go:63`). The footnote says "a computation, not a claim" (`computations.go:87`) | `backs` |
| Gap answer | A gap's answers are the live evidence anchors whose `answers` names it (`evidence_anchor`, III.1′.3). Its check kind demands a live answer of the kind whose row's `check` is that check kind. An anchor may answer a gap of any check kind; only the demanded kind discharges a demand | `computation` only, retired anchors included (`estoppel.go:280`, `views.go:569–572`). `source` is enforced nowhere | `check`, `table`, `idColumn` |
| Closure demand, at the write | Every `Close` event — red's `close` and the chair's `carry`, the two writers of `recordpb.Close` (`cli/lens/close.go:39`, `cli/chair/carry.go:18`, both through `seat.ClosurePayload`) — is refused on a `source` or `computation` gap that is not `evidence_answered` (live answer, or L1), whatever the disposition (ruling Source demand), `moot`, `not_a_defect` and a close `--superseded-by` included. The check is `requireEvidenceAnswered(q, b)` in `record.validate`'s `Close` arm (`record.go:1104`), asked of the writing transaction as `requireFailedCheckAnswered` is, and skipped under `Migrating` as that one is (`failedcheck.go:179`). It leaves `cli/lens/close.go:69–73`. The bench's closing ruling (a `motion_rule` docket disposition that closes) is not a `Close`, is not held to the demand, and is the exit for a demand the lens no longer holds: the lens puts the gap before the bench (`motion docket`), as at impasse | checked in `cli/lens/close.go:69–73` only, outside the transaction; `carry` (`carry.go:17–27`) writes a `Close` with no demand check, and `record.go:1147–1162` admits a carry naming any prior closure epoch, a fallen one included | `check` |
| Retire the answer — a closure falls | Retiring an answering anchor removes it from the answer set, so an open gap is awaiting again. A red closure of a demand gap FALLS at the retire that leaves no anchor live of those that answered the gap, in the demanded kind, when the close was recorded (`placed` < close < `retired` or none) — unless L1 met the demand at the close (ruling L1×L2: the `about_ref` anchor's retirement is permanent). A closure resting on answers at the close whose `about_ref` is retired AFTER the close and before the last answer's retire stands too (ruling L1-late): at that point L1 meets the demand. A fallen closure is not a closure for any reader (III.1′.3 Closure history): the gap is `open` and `awaiting_evidence`. An answer recorded after the close never restores it; only a new answer AND a new close do. A bench closure, a closure that rested on no answer (archived; `document` gaps), and a closure whose demand L1 met at the close, stand. Retire closes nothing | a retired `K-` still answers, and a closure never falls | `check` |
| Gap about the anchor | A gap minted `--about-kind anchor` about an evidence anchor is the failed check's carrier. Under check kind `source` or `computation`, its live-answer demand is met when that anchor is retired, read from `mint.about_ref` and `retire.anchors`; nothing is inferred, and red still closes it (ruling L1) | refused at `close` (`requireComputationAnswered`) even when retiring the anchor was the fix (the smoke's G-342d3cdc, G-78817cfa) | — |
| Work-list backing | `backing[]` is read from the ledger: the latest standing check of each anchor that answers the gap, with outcome or soundness, its seat and its event (`backingOf`, `viewjson.go:1181`), through the row's `checkTable` | only `Verify` is indexed, so a `K-` shows an empty outcome | `checkTable` |
| Cross-kind reference | `compute --cites C-…` (and 1b's `--input`) refuse a retired citation (`requireLiveCitation`, replacing `requireCitation` at `record.go:1051`). `retire --anchor C-…` is refused while a live computation names it in `cites` (1b: or `inputs`), and the refusal names those computations; blue retires or recomputes them first. Both refusals are asked of the writing transaction in `record.validate` — the `Computation` arm and the `Retire` arm (`record.go:1451–1464`) — so a `compute --cites C` and a `retire --anchor C` racing each other cannot both land. The tool refuses and never cascades (ruling L6) | `requireCitation` accepts a retired `C-` (`refs.go:239`). evidence.md prints "applies the method cited at C-…" for a citation the report no longer holds | — |
| Capture | `computation-rerun` samples live computations only (`computationrerun.go:93`) | retired computations sampled | — |

**What stays kind-specific, and why**

| Difference | Why it is intrinsic | Where it lives |
|---|---|---|
| How evidence is made: fetch, sha256 of bytes, span, OCR versus bundle, two runs, basis | It is the evidence itself | `cite.go:92–110`, `compute.go:120–165` |
| How it is checked: a reading against the source versus a re-run and a soundness judgement. The `unread` state is citation-only | A source can be unreachable or not the work. A stored bundle always runs (`RERUN_CALLS_THE_TOOL` is refused at write) | `anchorchecks.go:160–176` |
| Owning lens | Different competence | `anchorOwnerOf` (1a″: `EvidenceTables.owner`) |
| Footnote target: `[^N]` to an external source with a bibliography, deduped by url and pages, versus `[^K<n>]` to `evidence.md#K-…`. Only the target differs; numbering is now one scheme | A url is a shared external target; a computation's record lives in evidence.md | `anchor.kinds.assembly` |
| Red's corroborations | Red places a `C-` it found. Red never authors a program into the report | `verify.go:277` |
| Caption text (1c, 2): the footnote's text, which is the `--reason` for a computation and the title and page for a citation | It follows the footnote target | Fork "Caption" |
| Absent-crop placeholder (2) | `cache/` is never archived (R-commit). `computations/` is | Part 2 |
| Capture re-runs computations and re-fetches no source | Re-execution is local. Re-fetching at every capture breaks the politeness contract; 1b's `archive rerun --citations` is the operator's re-fetch | `computationrerun.go` |
| Per-kind budget measures (`measureCitations`, `measureComputations`) | The logic lens's budget unit is per kind by design | `mintbudget.go:61`, `:64` |

#### III.1′.2 Verbs [MODIFY] (contract: CLI flag set)

| Verb | Change |
|---|---|
| `blue cite` | `--answers G-…` (`flags.GapID().WithCheck(record.GapExists)`, as on `compute`, `compute.go:225`), written to `Cite.answers`. It joins `citeAsserted` (`cite.go:197–203`), so a keyed retry naming another gap is refused; `seat.AssertedCovers` (III.1.1) holds it. An anchor may answer a gap of any check kind; only the demanded kind discharges a demand (Fork Answer of another kind) |
| `blue compute` | The idempotent keyed retry returns `computation_id` (`compute.go:117`). `--cites` refuses a retired citation through `requireLiveCitation` in the write's validate (`record.go:1051`) |
| `blue retire` | `--anchor` alone accepts any evidence anchor (`retireAnchor`). The flag help (`retire.go:161–163`), `seat.Supplies` (`:158`), the refusals (`:88`, `:185`, `:192`) and the result (`:474`) say "an evidence anchor — a citation or computation" and name the anchor by its row's `noun`. `retire --anchor C-…` is refused while a live computation names the citation in `cites` (1b: or `inputs`); the refusal names each such `K-` and says to retire or recompute it first (ruling L6). The `Retire` arm's `--answers` refusal (`record.go:1460–1461`) says "an evidence-anchor retire" |
| `blue edit` | `refuseBaredEvidence` on `anchor.Backs` (`edit.go:250–290`, its comment `:257` and message `:285`). Its per-kind "what each anchor marks" text (`:536–555`) is read through one `EvidenceAt` accessor (`evidenceview.go:644`) |
| `lens verify`, `lens reproduce` | Both call `record.LiveSentence(run, id)` [NEW, moved from `verify.go:129`]. `reproduce` gains the refusal "a computation the report holds no sentence for is refused", mirroring `verify`'s |
| `lens close` | `requireComputationAnswered` (`close.go:69–73`, `:114–131`) is deleted from the verb: the demand is `requireEvidenceAnswered` in the record's `Close` arm (III.1′.1 Closure demand). Its refusal names `cite --answers` for `source` and `compute --answers` for `computation`; for a gap about an anchor, `retire --anchor <about_ref>` as well; and for a demand the lens no longer holds, `motion docket --id G` — the bench's ruling, not a close, disposes of it |
| `chair carry` | Held to the same demand, at the same write. `priorClosureEpochs` (`replay.go:342`) reads the standing closures (`closure_act` with no `fell_event`), so a carry naming the epoch of a fallen closure is refused as naming no closure, with the epochs of the standing ones |
| `cite`/`compute` refusal wording | `citeRefusal` and `computeRefusal` (`cite.go:224`, `compute.go:273`) become one `placementRefusal` (MisQuote, InFence), worded from the row's `noun` |

#### III.1′.3 The record [MODIFY] (contract: event schema)

```proto
// Cite [MODIFY]: + optional string answers = 26;   // the gap this citation answers; flag "answers".
//   remap class: id (respelled). Part 2's fields move to 27–35 (III.0).
// Retire [MODIFY] comments only: claim_kept (:1679) and answers (:1688) "a retire of a CITATION" / "a citation retire" →
//   "of an evidence anchor"; anchors (:1673) "Each named c- anchor is one counted claim" →
//   "Each named evidence anchor".
```

| View (`recordsql/views.go`) | Change |
|---|---|
| `evidence_anchor` [NEW] | `(anchor_id, kind, check_kind, answers, placed_event, retired_event)`. One UNION arm per `EvidenceTables` row, generated from the row (`table`, `idColumn`, `check`). `retired_event` = the first `retire_anchors` row naming the anchor (a retired anchor cannot come back: an edit invents no anchor, `bluedoc`; no archive holds a retired anchor later reopened or retired twice, `~/.claude/scratch/anchor-lifecycle/audit-r1/reopen-probe.txt`) |
| `closure_act` (`:445`) | gains `fell_event`: NULL on the bench arm; on the close arm, for a gap whose `check_kind` is a row's `check`, the greatest `retired_event` among the `evidence_anchor` rows of that kind answering the gap that were live at the close (`placed_event` < close and `retired_event` NULL or > close) — when there is at least one, all of them have a `retired_event`, and L1 did not meet the demand at or before the fall point (the gap is not `about_kind` `anchor`, or its `about_ref`'s `retired_event` is NULL or after that greatest `retired_event`; rulings L1×L2, L1-late). Otherwise NULL. A closing act whose `fell_event` is set is a FALLEN closure |
| `gap` (`:479–636`) | Every closure column reads standing closing acts only (`fell_event` IS NULL): `cx`, `bc`, `close_event`, `bench_close_event`, `closed_by_bench` (the arm of the first STANDING act), `bench_closed_last`, `open`, `stranded`. `computation_answered` → `evidence_answered`: a live `evidence_anchor` of the demanded `check_kind` answers the gap, or the gap is `about_kind` `anchor` and its `about_ref` has a `retired_event` (ruling L1). `awaiting_computation` → `awaiting_evidence`: `open`, `check_kind` is a row's `check`, and not `evidence_answered` |
| `convergence_at` (`:660–717`) | the `closed` CTE carries `fell_event`, and a gap is open at point P unless a closing act at or before P had not fallen by P (`fell_event` NULL or > P) |

`ComputationAnswers` → `EvidenceAnswers` (`estoppel.go:280`), reading `gap.evidence_answered`.
`Gap.NeedsComputation` → `Gap.AnswerKind` (`replay.go:199`), which returns the demanded anchor kind,
"" for `document`. The JSON fields `awaiting_computation` → `awaiting_evidence` (`viewjson.go:153`,
`:839`, `:974`), plus `answer_kind` beside it, so the seat reads which verb it owes. Blue's sitting
debt (`sitting.go:257–285`) reads `AwaitingEvidence`, and its words come from the row. No
compatibility alias: the old names leave every surface (CLAUDE.md).

**Closure history, re-derived (rulings L2, L2 history).** What a closure-history field means once a
closure can fall: `HasClosed` = the gap has a standing closing act (so `HasClosed` = `!Open`, as
today); `ClosedEpoch` = the epoch of its last standing closing act; `Closure` = the body of its last
standing red close; `ClosedByBench` = the arm of its first standing closing act. A fallen closure
contributes to none of them, so after a re-close the gap is counted once, in the re-close's epoch,
and the repair denominator (`view.go:461`, `:571`) counts it once. Point-in-time readers ask "closed
at P": some closing act at or before P had not fallen by P. Census L10 is every reader; each one's
reading after a fall:

| Reader | Reads | After the retire that fells a closure |
|---|---|---|
| `gap` view, `board_counts`, `gapState`/`requireOpenGap` (`refs.go:298`, `:308`), `remand.go:165`, spot-check's query (`spotcheck.go:80`) | the view | open, `close_event` NULL |
| `gapStatesOf` (`gapstates.go:103`) → `HasClosed`, `ClosedEpoch`, `Closure`, and through them: board JSON closure payload (`viewjson.go:475–485`), work list, `scorecard.go:154`, telemetry `closedAtR` (`view/view.go:427–461`, `:571`), chart (`report/chart.go:36–54`), `docs.go:476`, spot-check archive (`spotcheck.go:51`), `dashboard/model.go:490`, declarations (`declarations.go:105`), capture's claims (`capture.go:851–864`), report fates (`assemble.go:1061–1145`), `view.go:347–385` | the view | not closed in any bucket or payload until a standing re-close |
| `convergence_at` | the view, per point | closed at points between the close and the fall; open from the fall |
| `engagementOf` (`manifestowed.go:145–160`: a gap blue's sitting "found closed") and `closingDutiesOf` (`closingowed.go:110–147`: a closing "not owed" because the gap closed) | `Close` events in a Go fold | each takes the standing set from `closure_act` (`fell_event`), handed in by its entry point as `motionsAt`'s fold is today: a close counts at P when it had not fallen by P |
| `priorClosureEpochs` (`replay.go:342`, `carry`'s prior closure) | the `close` table | standing closures only |
| the consistency oracle's raw walk (`consistency.go:184–241`, `holdGap` `:286–300`) | its own fold, deliberately | re-derives L2 itself from `Cite`/`Computation` answers, `Retire` anchors, `Mint` check kind and about, and closing events — a second opinion, not a reader of `closure_act` — and holds `Open`, `HasClosed`, `ClosedEpoch` and `ClosedByBench` to it |

**Readers of an anchor's answers (census L12).** Every reader of `Computation.answers` reads both
kinds off `evidence_anchor`:

| Reader | Change |
|---|---|
| `gapActsOf` (`gapacts.go:55–62`): a computation answering G, and a reproduce of it, are acts on G — the act that begins a remand's exchange (`impasse.go:160`, `remand.go:25`) | an anchor answering G, of either kind, and a check of that anchor (`verify` or `reproduce`), are acts on G |
| `groundsActSQL` (`grounds.go:114–120`, `:144`): a computation answering G concerns G; a citation only when an edit answering G wrote its location | an anchor of either kind concerns G when it answers G; the edit-wrote arm stays for a citation placed by an edit's text |
| `ComputationAnswers` (`estoppel.go:280–285`), `views.go:569–572` | `EvidenceAnswers` and `gap.evidence_answered` (above) |

Unchanged, and for both kinds alike: `blueAnswersOf` (`blueanswered.go:99–175`), the impasse
movement fold (`impasse.go:186–199`) and `manifestowed.go:51` read `BlueEdit.answers` and a
claim-kept `Retire.answers`, never an anchor's answers; `report/edits.go:36`, `view/changes.go:70`,
`changesjson.go:121` read the `change` view (edits and claim-kept retires).

#### III.1′.4 Render [MODIFY]

| Site | Change |
|---|---|
| `report/computations.go:111–218` `evidenceDoc` | live computations only. "anchored to nothing" stays only for a live computation no document references (a defect, not a retirement). Each entry gains "- **backs**: <the anchor's live sentence>", as a citation's has (`citations.go:47–49`) |
| `report/citations.go:22–73` `citationsSection` | live citations only |
| `computations.go:32` `computationNumbers`, `weaveComputationRefs` | per document, first appearance, live only (the shape of `weaveCitations`, `assemble.go:49`). Footnote text drops "a computation, not a claim" (`:87`) |
| `computations.go:192–196` "applies the method cited at" | prints the citation's footnote. It is never a retired one: retiring a citation a live computation names is refused (ruling L6) |
| `record/evidenceview.go:365–575` `EvidenceJSONOf` | `sources` and `computations` hold live anchors. Both unverified counts come from the ledger `state()` (`:563–572`). `--anchor` returns a retired anchor with `retired` (event id, reason) (Fork Show evidence after retire) |
| `computationverdict.go:224–231` ("a retired anchor keeps the verdicts it left with") | still true of the record. evidence.md no longer shows it, and the comment says so |
| `assemble.go:925–967` `withdrawnClaims` | a `claim_kept` row names the anchor by its `noun`: "…a citation it stood on was retired" or "…a computation it stood on was retired" |
| `view/changes.go:48`, `:157–158`, `:190–194`; `record/changesjson.go:30`, `:55`; `recordsql/views.go:307` (the `change` view's comment) | "each edit, and each evidence anchor retired from a sentence the report keeps"; a row names the retire by the anchor's `noun`: "retired computation K-…, its sentence kept" |

#### III.1′.5 Agent-facing text [MODIFY] (contract: prompts and constitutions)

Edit the sources, then run `agentgen`, `enginegen` and `vocabdoc`. Census rows L5 and L13 (III.9).
Help pages name commands; seat prompts and constitution prose name acts (CLAUDE.md).

| Carrier | Becomes |
|---|---|
| `help/retire.md:9`, `:11` | "EVIDENCE LEAVES ON ITS OWN, ITS SENTENCE KEPT": `--anchor` alone, a citation or a computation. A cited or computed sentence that is to go first loses its anchors this way, then is edited like any prose. A citation a live computation names is refused, naming the computation: retire or recompute it first (ruling L6). Retiring an anchor that answered a closed gap fells that closure unless the gap was about an anchor already retired: the gap reads open and awaiting evidence |
| `help/edit.md:28` | "EVIDENCE NEVER STANDS BARE" |
| `help/reproduce.md:21`; new paragraph | "A COMPUTATION THE REPORT HOLDS NO SENTENCE FOR IS REFUSED", mirroring `help/verify.md:19`. The text promising retire is now true |
| `help/close.md:11`; `help/mint.md:13`, `:15` | `source` and `computation` close only on a live answer, under every disposition, `moot` and `not_a_defect` included. Retiring the answer reopens the debt, and a closure resting on it falls until a new answer and a new close. A gap about an anchor is met by retiring that anchor, and you still close it. A demand you no longer hold is not closed: put the gap before the bench (`motion docket --id G`), whose closing ruling is not held to the demand. A `source` gap's fix that rewords the sentence leaves it unanswered: the sentence's citation answers nothing until blue cites on it with `--answers G` |
| `help/carry.md` | a carry restates a STANDING closure: one that fell when its answer was retired names no closure, and the carry is held to the demand as `close` is |
| `help/cite.md`, `help/compute.md` | `--answers`: the gap this anchor answers; a `source` gap is settled only by a citation that names it, so after a reword cite the source on the reworded sentence with `--answers G` (where it replaces the old citation, retire that one with `--answers G` first). How the anchor leaves (`retire --anchor`, sentence kept) |
| `cli/seat/verbs.go:289`, `:661` (`show changes` help and its `--id` flag) | "every edit and evidence-anchor retire" |
| `cli/seat/verbs.go:295` (`show evidence` help) | "…or of a retired citation, is in the rows" → retired anchors are not in the rows: `--anchor <id>` returns one with `retired` |
| `record/blueanswered.go:57`, `:78` | "(an edit or an evidence-anchor retire answering it)" |
| `scripts/agentgen/src/fragments/blue-search-and-cite.md:14–18`, `:28`, `:31–32` | one lifecycle for both. ":31–32 a new computation with the old one retired" is now an act the tool admits: retire the old computation as the answer to the gap, then compute again on the kept sentence as the answer to the gap. A gap that asks for a source is answered only by a citation placed as its answer: a reword leaves it unanswered until you cite on the reworded sentence as that gap's answer |
| `scripts/agentgen/src/blue-researcher.md:113` | `awaiting_evidence: true` with `answer_kind`: a program or a source is owed from YOU |
| `scripts/agentgen/src/red-lens-evidence.md`, `red-lens-computation.md` | a `source`/`computation` check kind is a demand for a live answer. Do not demand retirement through a `document` gap: mint about the anchor (L1). A demand you no longer hold goes before the bench; you do not close it |
| terms.json `retire` (:779), `anchor` | "retiring an evidence anchor alone — a citation or computation — keeps its sentence". Regenerate `docs/vocabulary.md` |
| FEOV `docs/citations-and-bibliography.md:146`, `:158`; `docs/propagation-and-anchoring.md` | the one lifecycle |
| `seatprobe/boards.go:288` (retire `Because`) | "deleting the sentence is refused while evidence stands on it; retire the anchor first" |

#### III.1′.6 migrate — one translation (contract: record migration)

| Change | Translation | Asserted on every archive |
|---|---|---|
| `Cite.answers` (field 26) | none: no archived cite holds one. remap classes the field `id` (`remaplists_test.go`) | zero refusals |
| DDL: the `cite.answers` column; the `evidence_anchor` view; `closure_act.fell_event`; the `gap` and `convergence_at` views | one epoch move, E1′ (`record.EventSchema` from `requirements.json` by `schemagen`). `TestEveryDDLChangeBumpsTheEpoch` fails without it; `schema.epoch.json` restamped with `UPDATE_GOLDENS=1`; the `schema.sql` golden read line by line | — |
| `Close` under the demand | none: replay sets `Migrating`, under which `requireEvidenceAnswered` is skipped, as `requireFailedCheckAnswered` is | zero refusals |
| Render and derived state | per-document `[^K<n>]` numbers, evidence.md without retired entries, claim counts, fallen closures | goldens and the flip census (V1a′.6) |

**Residue, accepted** (each counted by V1a′.6's flip census over the 33 archives):

| Residue | Why it stays |
|---|---|
| Archived edits that left a `K-` bare, and claim retires that took a `K-` with its sentence | what a seat did. Replay exempts refusals (`Migrating`) |
| Archived `source` and `computation` gaps closed on no live answer (prose; `computation` closes under `Migrating` exemptions) | a closure that rested on no answer stands (ruling L2) |
| Archived closures whose answer was later retired (the 2026-10-30 fixture's `G-221322e1`: red close 701 on `K-71a99b9a`, retired 722) | they re-derive fallen (ruling L2): open from the retire until a later standing closing act (`G-221322e1`: the bench at 792) |
| Two definitions of "live": the SQL `evidence_anchor` (not retired) and `record.LiveSentence` (in the report and not bare). An archived bare anchor that was never retired satisfies a demand in SQL and is refused by `verify`/`reproduce` | a bare unretired anchor is producible only by an archived edit (live edits refuse a bare anchor), and the flip census counts the archived bare anchors that answer a gap |
| Re-derived blockers and claim counts differ from what an archived run's seats saw | derived views; the archived verdict is the record's |

#### III.1′.7 The release fuzz drive (contract: release gate)

Census L6. The drive (`releasegate/fuzz/fuzz_test.go`) exercises the acts 1a′ adds and the
refusals it adds, and the tally (`sweeptally_test.go`) fails a sweep that never reaches them:

| Arm | Drive | Tally gate (`sweeptally_test.go`, judged over the four parts) |
|---|---|---|
| Retire a computation | the unsound-computation branch (`:1111–1129`) retires `--anchor K-` `--answers G` on a coin, then computes again on the kept sentence `--answers G` | `computationRetires` > 0 |
| Source demand | a `source` gap's close is driven unanswered first (refused, Conflict), then answered by `cite --answers G`, then closed | `sourceClosesRefused` > 0 and `sourceAnswers` > 0 |
| Closure falls | on a coin, blue retires the anchor answering a closed demand gap; the drive reads the board (gap `open`, `awaiting_evidence`), drives a `carry --carried-from <the fallen close's epoch>` to its refusal, answers again and closes again | `closuresFallen` > 0 and `fallenCarriesRefused` > 0 |
| L6 | `retire --anchor C-` of a citation a live computation cites → refused naming the `K-` | `citedRetiresRefused` > 0 |

The consistency oracle runs over every simulated debate (`internal/consistency.Check`), and its raw
walk re-derives L2 (III.1′.3), so a projection that counts a fallen closure fails the run.

### Part 1a″ — one failed-check carrier and the judged sentence (epoch E1″) (rulings L4, L5, Staleness)

Built on 1a′'s merged head: the kind rows, `LiveSentence`, `evidence_anchor` and the re-derived
closure history are 1a′'s. 1a″ adds two `EvidenceTables` columns (`owner`, `failedWord`), one
record field (`Reproduce.claim`) and one oracle rule, and moves a refuted citation from the
contradiction carrier to the failed-check carrier.

#### III.1″.1 The lifecycle rows 1a″ adds

| Stage | One rule (every evidence anchor) | Today (main `a31a7b06`) | Kind row supplies |
|---|---|---|---|
| Check records its sentence | Every check records the sentence it judged, read by the tool from `record.LiveSentence` at the write and never typed by the seat: `Verify.claim` today, `Reproduce.claim` [NEW] (ruling L5). Both verbs declare it with `seat.Supplies(c, "claim", …)` (`verify.go:117`) | `Reproduce` holds the anchor and no sentence (`record.proto` `Reproduce`), so a judgement cannot say which wording it was of | `checkTable` |
| Staleness | ONE rule decides whether a check stands, for both kinds, and it is the ledger's today: an edit that changes the anchor's sentence names it in `BlueEdit.reopened`, and only a check from a sitting opened after that edit stands (`anchorStanding.sinceEdit`, `anchorchecks.go:147–149`, fed by `verify` and `reproduce` alike, `:245–275`). The recorded sentence is how staleness SURFACES: (1) evidence.md prints, under each check of either kind, "of the sentence: …" when its recorded sentence is not the anchor's sentence now (III.1″.4); (2) the consistency oracle (`internal/consistency.Check`, run over every simulated debate of the release fuzz) gains one rule, iterated over `EvidenceTables` (ruling Staleness): a STANDING check whose recorded sentence differs from the anchor's live sentence is a violation, naming the anchor, the check's event and both sentences. A check with no recorded sentence (a migrated reproduce) is outside the rule and counted, never folded into "agrees" | citations: the ledger rule plus evidence.md's "of the sentence" line (`citations.go:100–107`); computations: the ledger rule (`SoundnessReopened`, "not judged since the sentence changed", `computations.go:187`) and no sentence to show; no oracle rule for either | `checkTable` |
| Failed check | One state, `stateFailed`: a standing check of the sentence that does not hold (unsound, refutes or absent), carried by the reader's gap about the anchor (`about_kind` `anchor`, minted by the check's seat from the check's sitting on, `failedcheck.go:65–80`), as an unsound computation is today (ruling L4). Its result names the failed check and the gap owed, as an unreadable reading's does (`failedCheckWords`, `verify.go:631–655`). A repeat failed reading while its gap is open is refused, as a repeat `unsound` is (`requireNotARepeatFailedCheck`). A contradiction keeps only readings with no anchor: red's corroborations. `stateUnread` stays citation-only | refuted citation → contradiction (`citationid.go:661`); unsound computation → failed-check gap (`failedcheck.go:56`) | `failedWord` |
| Blockers | The blocker carries the anchor id and its kind. Kind is read from the row, never inferred from `Detail` holding a sha (`passblockers.go:162`, `:186`, `:226`, `:234`, `:245`, `:266`, `:415`, `:428`; `dispatch.go:1299`; `sitting.go:413`) | inferred from `b.Detail != ""` | `noun`, `owner` |
| Owner | The anchor's row names its owning lens | `anchorOwnerOf(bool)` | `owner` |
| Capture | `assembly-screen` fails on any live evidence anchor in `stateFailed`, not only citations found against (`capture.go:434`) | computation counterpart missing | — |

#### III.1″.2 Verbs [MODIFY] (contract: CLI flag set)

| Verb | Change |
|---|---|
| `lens verify` | a `refutes`/`absent` reading of a citation opens no contradiction: its result carries `failed_check` (owner, the gap that carries it or the one owed) and its human line says "this reading finds against citation C-…: PASS is refused while the report cites it, and you owe a gap about it (`mint --about-kind anchor --about C-…`, which takes no mint of your budget)" (`failedCheckWords.words`, `verify.go:644–655`, worded from `failedWord`). A corroboration's reading (no anchor) opens or joins a contradiction as today (`verify.go:624–626`, `:676–686`) |
| `lens reproduce` | writes the sentence it read to `Reproduce.claim`, declared `seat.Supplies(c, "claim", "the sentence holding the --anchor token in the report as it stands")` as `verify.go:117` declares it (ruling L5). `reproduceResult` gains `claim`, and its human line quotes the sentence judged, as `verifyResult.human` does (`verify.go:696`) |
| `lens mint` | unchanged flags; a mint `--about-kind anchor` by the reader whose reading found against the citation answers its failed check outside the budget (`answersOwnFailedCheckAt`), as for an unsound computation |

#### III.1″.3 The record [MODIFY] (contract: event schema)

```proto
// Reproduce [MODIFY]: + optional string claim = 13;   // the sentence this judgement is of, read off the report
//   by the tool (record.LiveSentence) at the write, as Verify.claim is (ruling L5). Not `required`:
//   migrated reproduces hold none, and CheckRequired has no migration exemption; the live writer sets it
//   on every write (V1a″.2). remap class: data, beside Verify.claim and Retire.claim (`remap.go:135–138`).
// AboutKind [MODIFY] means: ABOUT_KIND_ANCHOR (:1130) "a citation or computation, by anchor, whose check
//   failed: a computation judged unsound, a citation found against or whose work went unread";
//   ABOUT_KIND_CONTRADICTION (:1129) "a contradiction red recorded on a source it corroborated …".
```

The ledger (`anchorchecks.go`): `anchorStanding.computation bool` → `kind string` (census L2).
`anchorCheck` gains `sentence`, from `Verify.claim` or `Reproduce.claim` ("" for a migrated
reproduce), the one field the Staleness row's evidence.md line and oracle rule read.
`stateFoundAgainst` and `stateUnsound` → `stateFailed` (ruling L4). `failedChecksOf`
(`failedcheck.go:56`) adds the failed state of any kind. `unansweredContradictions`
(`citationid.go:661`) and `contradictionsOf` (`:522`, which keys an open contradiction on
`Verify.claim`, `:533`) fold only readings with no anchor (corroborations); their citation-standing
arm (`:668`, "blue RETIRES the anchor … the anchor HOLDS") is deleted, since no anchored reading
opens a contradiction. `Reproduce.claim` never enters them. Census L7 is every reader.

`sitting.go:369–422`, read for duties lost: the lens's owed list takes `BlockerContradiction` (its
reading to raise) and `BlockerFailedCheck` (owed until its gap is minted, `:398–406`). An anchored
refuting reading moves from the first to the second, so the duty "raise it, and answer it with the
gap minted from that finding" becomes "mint the gap about the anchor" — the same debt, carried as an
unsound computation's is. The contradiction-at-mint-cap clause (`:420–425`) stays for corroborations.
No contradiction-only duty is lost; `seatprobe`'s board (R15) asserts the new expectation.

#### III.1″.4 Render [MODIFY]

| Site | Change |
|---|---|
| `report/citations.go:85–110` `readings` → `checkLines` [MODIFY, one function for both kinds] | renders every check of the anchor, one per list item (seat, epoch, outcome or soundness, page, text or note), from `EvidenceJSONOf`'s rows, which carry `claim` for both kinds, and from that `claim`: nothing when it equals the anchor's live sentence; "of the same sentence as the check above" when it repeats the one above; else "of the sentence: <recorded>" — the staleness surface. An absent sentence (a migrated reproduce) prints no sentence line, as an empty `Verify.claim` prints none today (`citations.go:102`); its standing still comes from the ledger |
| `report/computations.go:173–191` | a computation's judgements are listed one per line through `checkLines`, where today only their notes are folded into one line |
| `record/evidenceview.go:404` verification JSON `claim`; the computation judgements' JSON [MODIFY] | each judgement of `show evidence` gains `claim`, the sentence it was of, as each verification carries it |
| `counts.sources_refuted` (`evidenceview.go`) | counts live citations in `stateFailed` by a finding-against check, from the ledger |

#### III.1″.5 Agent-facing text [MODIFY] (contract: prompts and constitutions)

Census L13 (the 1a″ half) and L14.

| Carrier | Becomes |
|---|---|
| `help/verify.md:11`, `:15`, `:19` | a reading that finds against a citation (`refutes`, `absent`) is a failed check, carried by your gap about the citation, as an unsound computation's is. A contradiction is opened only by a reading of a source you corroborated |
| `help/reproduce.md`, new paragraph | the sentence holding the anchor is read off the report and recorded with your judgement; an edit that changes it leaves your judgement standing on an earlier wording, owed again |
| `help/mint.md:13`, `:15`; `help/close.md:11` | the failed-check paragraph covers a reading found against a citation; the contradiction exemption in `mint.md:13` covers corroborations only |
| `cli/seat/verbs.go:295–297` (`show evidence` help) | `counts.sources_refuted` reads the failed state; `unanswered_contradictions` are corroborations' only, each row's `anchor` is always empty and is removed with the "A row with an `anchor` is answered as well…" sentence |
| `cli/lens/verify.go:619–626` (`claim`, `Contradiction` comments), `:644–655` (`failedCheckWords.words`) | the found-against arm (III.1″.2) |
| `record/passblockers.go:139–155` (`contradictionCitationAnswer`), `:237`, `:248`, `:269–271`, `:343–366`, `:415–428` | `contradictionCitationAnswer` is deleted; the failed-check words name both failure kinds from `failedWord`; the contradiction words speak of a corroborated source |
| `record/dispatch.go:18–21` (`Party.GapIDs` desc → debate.js :393 by `enginegen`), `:1289–1291` | "…or a check of its that failed; empty for blue readied to remedy a failed check — fix or retire a computation judged unsound or a citation found against, retire a citation no check could read" |
| `record/anchorchecks.go:107`, `:143` (comments) | the failed state |
| `scripts/agentgen/src/fragments/lens-audit-and-mint.md:6`, `fragments/lens-duties.md:5` (the "contradiction your own finding raised" clause) | the failed check in both kinds, carried by the reader's gap about the anchor (ruling L4); the contradiction clause names a source you corroborated |
| `scripts/agentgen/src/red-lens-evidence.md` | a reading that finds against a citation is your failed check: mint the gap about the citation in that sitting |
| `scripts/agentgen/src/red-lens-computation.md` | your judgement records the sentence it was of; `show evidence --anchor K-…` shows that sentence beside each judgement, and one of an earlier wording no longer stands |
| terms.json `failed check` (:1072), `contradiction` | `failed check` gains "a reading finds against it" (ruling L4); `contradiction` is narrowed to corroborations. Regenerate `docs/vocabulary.md` |
| `seatprobe/boards.go` | a board with a refuted citation expects the lens's `mint --about-kind anchor` about it, not a contradiction finding (R15) |

#### III.1″.6 migrate — one translation (contract: record migration)

**The archived chain, measured** (`~/.claude/scratch/anchor-lifecycle/rev-r1/l4census.txt`, over
the 33 archives migrated to epoch 50 by `feov-migrate` at `f5c09519`, every one exit 0). No
archived mint is `about_kind` `contradiction` (`about.go:48` refuses one), and none is
`about_kind` `anchor` on a citation. The raise chain is: a reading with an `anchor` and outcome
`refutes`/`absent` opens contradiction `<its events.id>` → a Finding `about_kind` `contradiction`,
`about_ref` `<that id>`, raises it → a Mint whose `found_by` names that finding answers it.

| Measured | Count | Where |
|---|---|---|
| anchored `refutes`/`absent` readings | 16 | sqlite-schema-source 2, quadratic-formula 1, wave-a 6 (all six of one citation, retired by the archive's end), m19 1, m23 1, m30 1, m32 4 |
| contradictions they opened (the rest joined one) | 11 | — |
| findings raising an anchored contradiction | 5 | m30 `F-8410bec7`; m32 `F-d8329c9a`, `F-1c990edb`, `F-b59925d3`, `F-5e7dfd3d` (one per reader) |
| mints `found_by` such a finding | 2 | m30 `G-dc877265`, m32 `G-f417112d`: each minted by the reader, `--quote`-located (gap marker placed at ev433, ev100), `check_kind` `document`, closed `repaired` |
| readings carried in L4's shape (a mint by the reader, `about_kind` `anchor` on the citation, after the reading's sitting opened) | 0 | — |
| contradictions mixing anchored and unanchored readings | 0 | — |
| contradictions opened by a corroboration | 4 | unchanged |

Under L4 an anchored reading opens no contradiction, so replay's `requireOpenContradiction`
(`record.go:1062`) would refuse the 5 findings.

| Change | Translation | Asserted |
|---|---|---|
| A Finding `about_kind` `contradiction` whose `about_ref` names a contradiction opened by a reading with an `anchor` | `about_kind` `anchor`, `about_ref` = that reading's `anchor` (a citation `requireAboutAnchor` finds). One translation, at the site that already translates this reference (`migrate/replay.go:395–404`, which maps the reading's old `events.id` through `contradictionRef`): replay notes each reading it writes with an `anchor` and outcome `refutes`/`absent`, by its old id, with the anchor as written (after remap); a finding naming such a reading is written `about_kind` `anchor`, `about_ref` that anchor, and every other finding about a contradiction is translated by `contradictionRef` as today. Keyed on the archived reading, never on text | the 5 measured findings translated, and no other finding; m30 and m32 migrate with zero refusals |
| The 2 mints `found_by` a translated finding | none. A `--quote` mint placed a gap marker in the report (`anchor` rows at ev433, ev100), and archived edits replay over that text; moving its subject to `about_kind` `anchor` drops the marker the later acts were written against. Their `found_by` stays, naming a finding now about the citation | zero refusals; their location and marker unchanged |
| `Reproduce.claim` (field 13) | none: no archived reproduce holds one, and migrate reads no sentence back into it (the report an archived judgement saw is not re-derivable at its event). remap classes the field `data` (`remaplists_test.go`) | every migrated reproduce has no `claim` |
| DDL: the `reproduce.claim` column, the `AboutKind` means | one epoch move, E1″; `TestEveryDDLChangeBumpsTheEpoch`; `schema.epoch.json` restamped | — |

The minting lens at each answering mint held no open gap (m30 `red-lens-adversary` before ev432,
m32 `red-lens-evidence` before ev99: 0 open), so a mint that no longer answers its own raise is
admitted within the budget at replay.

**Residue, accepted** (counted by V1a″.4's flip census):

| Residue | Count | Why it stays |
|---|---|---|
| Anchored readings whose failed check no gap carries in L4's shape | 16 (every one) | the archived lens raised a contradiction and, twice, minted a `--quote` gap from the finding; neither is a gap about the anchor, and translating the two mints breaks the report text (above). The flip census counts, per archive, those still standing at the archive's end, which re-derive as an uncarried failed-check blocker |
| Archived reproduces carry no `claim` (19 in the 2026-10-30 archive; every archive's count in the census) | — | which wording an archived judgement judged is not on its event. evidence.md prints no sentence line for it; the ledger's edit rule still decides whether it stands, and the oracle rule skips it, counted |

#### III.1″.7 The release fuzz drive (contract: release gate)

Census L6 (the 1a″ half). The reading drive (`fuzz_test.go:5220–5340`) splits on what the reading
checked:

| Arm | Drive | Tally gate |
|---|---|---|
| A refuted citation (the `verify --anchor` branch, `:5247–5287`) | a `refutes`/`absent` reading that prints a contradiction fails the run; one that prints no `failed_check` fails the run (`names`, `:5234–5246`). `answerFailed` (`:2846–2900`) mints the gap about the citation, reading the blocker's kind field, not `Detail` (`:2857`, `:2867`); on a coin it reads the citation again `--as refutes` while that gap is open → refused (Conflict); on a coin blue retires the citation or rewords the sentence, and the blocker clears | `refutedAnswers` > 0, `refutedRepeats` > 0 |
| A corroboration (the `corroborate` branch, `:5288–5310`) | unchanged: `raisePrinted`, `answerContradictions`, `raiseAndMint` (`:2344–2427`) and `quoteRaisesNothing` run on corroborations' contradictions only | the existing contradiction gates, now over corroborations |
| Staleness | a test seam (never a seat path) drops the anchor from one edit's `BlueEdit.reopened` in one simulated debate per part; the oracle must report that debate | `stalenessForged` = `stalenessReported` |

### Parts 1a′ and 1a″ — what 1b, 1c and 2 now assume

Each later part's text below is amended in place. This table is the index.

| Part | Was | Now | From |
|---|---|---|---|
| §I placement rule | "No new lifecycle" (a premise two lifecycles falsified) | "1a′'s one lifecycle" | 1a′ |
| 1b `--input C-…` | refused when not on the record, not cached, or tool-composed | also refused when retired (`requireLiveCitation`). Retiring a citation a live computation reads is refused, naming it (ruling L6) | 1a′ |
| 1b `archive rerun` | every act | every act, retired included: it audits the record, not the report. A retired act's line says `retired` and keeps its outcome and exit semantics. `--citations` likewise | 1a′ |
| 1c assembly, figure and table numbering | numbered per kind in document order | over live anchors. A retired act's block is never placed | 1a′ |
| 1c voice lens | "edit the sentence, retire the `K-` anchor, compute again" | "retire the `K-` anchor (the sentence stays), compute again on the sentence" | 1a′ |
| 1c evidence.md | each entry adds its shown output | each live entry | 1a′ |
| 1c judging a shown block | the computation lens reads the image against "the sentence" | the sentence a shown block refers to is the one its judgement records (`Reproduce.claim`, ruling L5): the link from the report's text to the figure's anchor. `show evidence --anchor K-…` returns the shown output beside each judgement and the sentence it was of | 1a″ |
| 1c, 2 epochs | E3, E4 the next free | still the next free when each lands, after E1′ and E1″ | 1a′, 1a″ |
| 2 wrong crop | "`retire` + cite again (the existing lifecycle)" | `retire --anchor C-…` keeps the sentence, then cite again: the same act as a wrong figure | 1a′ |
| 2 `Cite` fields | `show = 26` … `crop_sha256 = 34` | `answers = 26`, so `show = 27` … `crop_sha256 = 35` | 1a′ |
| 2 a refuted crop | a contradiction | a failed check carried by the evidence lens's gap about the citation | 1a″ |

### Part 1b — `--input` and `archive rerun` (epoch E2)

| Change | Detail |
|---|---|
| `compute --input C-…` (repeatable) [MODIFY] | the citation's RAW cached bytes (`fetchcache.Read`) — never extracted text (R-input). Placed read-only at `sources/C-…` in each copy; id + sha256 enter `bundle_sha`. Refused: a citation not on the record, or retired (`requireLiveCitation`, 1a′); bytes the cache does not hold; a citation whose bytes the tool composed — `copy_substitution` `recovered` or `named_backend` with no `copy_url` (`record.proto:1237`, `:1239`): no document exists to re-fetch |
| `reproduce`, capture's audit [MODIFY] | `compute.Rerun` materialises `sources/` the same way |
| Package `archive` [NEW] (`internal/archive`) | the one writer and the one reader of the run-archive tarball: `archive.Write` (from `ArchiveRecord`, `capture.go:2045`; `capture` calls it) and `archive.Extract` (replacing `internal/record/recordtest/archive.go:34`'s loop and `archdump`'s `untar`, `main.go:149`). Every entry `0o644`, created with explicit permissions; identity holds no mode (III.1.3). Capture's transcript tarball (`capture.go:1748`) is another format and stays. It commits no `cache/` (R-commit) |
| `feov-record archive rerun <runDir>` [NEW] (N1) | input: an extracted archive brought to the current epoch by `feov-migrate --from <runDir> --to <freshDir>` (an older epoch is refused by `requireThisEpoch`, pointing to `feov-migrate`). Re-fetch each input from where its bytes came — `copy_url` when set, else `url` (`fetchcache.Resolve`, paced) — verify sha256, then `compute.Rerun` each bundle once and compare every act (III.1.4: never reads `rerun/`, writes nothing to the store). Writes no event. It audits the record, not the report, so a retired act is re-run as well and its line says `retired` (1a′). `--citations` verifies every citation the same way (N2); a tool-composed citation prints `refused` (no document). `--help` states: it runs every archived program with the operator's privileges, filesystem and network (R12) |
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
| `--citations` (2): a re-derived crop whose sha256 ≠ `crop_sha256` — the source verified; a renderer change can move pixels | `differs` (C-id, box, recorded and derived sha) — not refused | 1 |
| bundle absent, older epoch | `refused` (named) | 1 |

### Part 1c — `--show` on `compute` (epoch E3)

| Change | Detail |
|---|---|
| `compute --show table\|image\|text` [MODIFY] | `out/` decoded as its kind: `table` one CSV with a header row (a cell holding a newline refused); `image` one PNG or SVG; `text` one markdown file under `edit`'s refusals (`composedRefusal` chain, `cli/blue/ingest.go:147`, called by `edit.go:308` and `ingest.go:80`, plus the image refusal below) and its voice advisory (`cli/blue/edit.go:104`). Any other count or extension, or a file that does not decode → refused. Joins the asserted set |
| Assembly [MODIFY] (`report/layout.go`, `report/computations.go`) | block after the anchor's paragraph, numbered per kind in document order over live anchors (1a′), inserted after every escape and the weave (no escape reads a shown block): `TABLE` → markdown table, every cell `Literal`, then `*Table n.* Literal(R)`; `IMAGE` → `![Figure n](<placed>)` — alt is tool text only — then `*Figure n.* Literal(R)`, in report.md (run root: the link resolves in the run directory and in an `archdump` of its archive, which copies `computations/` beside the report set; the archive itself holds no report.md, `capture.go:2045–2160`); `<img src="data:image/…;base64,…">` in report.html (an `<img>` runs no SVG script); `TEXT` → verbatim, already held to `edit`'s refusals. `<placed>` = `computations/<sha>/acts/<K-id>/out/` + each path segment `url.PathEscape`d, joined by `/` (a nested `out/sub/f.png` resolves), from one function (`placedPath`): the program chose the name, and PathEscape leaves no space, `(`, `)`, `[`, `]`, `<`, `>`, `\` or line break in a segment |
| Composition rule [NEW in 1a, III.1.5] (`report/notes.go`) | `report.Literal(s)` backslash-escapes every ASCII punctuation character and folds each line break to a space, so the string renders as its own characters in report.md and report.html and opens no image, link, raw HTML, footnote, emphasis or table cell. It is the one composer of seat, program and source text in a shown block — 1c: table cells (replacing `cell`'s pipe-only escape for shown tables, `assemble.go:1525`), the caption; 2: the caption (source title, page when there is one), the placeholder, the `--show text` span. Sweep of composition sites a part adds: these; evidence.md (names per III.1.5; `docs.go:211`'s whole-document escape moves per entry, applied before the entry's figure line is appended, so no tool-composed image is under it); run.md's absent-crop line (tool text, C-id and box) |
| Code regions [MODIFY] (`layout.go:350`, `notes.go:32`) | serve markdown STRUCTURE only — the weave-label `[^` escape and the composed-image `!` escape; raw HTML is #1569's. `eachLine` marks a line fenced when goldmark's tree (`mdReader`) places it in an `ast.FencedCodeBlock` (fence lines included) or `ast.CodeBlock`, so its seven callers (§II) inherit the renderer's reading; code spans come from `ast.CodeSpan`; `reCodeSpan` is deleted. Each escape parses its unit as it ships — a whole document (`docs.go:129`, an evidence entry) or a section block (`layout.go:89`) with `mdReader`, a one-line title (`assemble.go:151`, `proofs.go:85`) with `lineReader`. Units join closed: assembly appends the closing fence to any unit whose parse ends inside a fenced block, so no fence state crosses a unit boundary and each unit's reading is the document's (`probe-out.txt`'s case) |
| Raw HTML and images [MODIFY] (#1569; `notes.go:43`) | Raw HTML is #1569's in both outputs: report.html in md.go; report.md is written (`docs.go:391`) through the same decision on the same whole-document parse — every `<` outside code, autolinks and the tool's shapes backslash-escaped (each `<`, not only a node's opener: `probe2-out.txt`). So 1c adds no raw-HTML escape at composition and no raw-HTML refusal at `edit`/`ingest`. Images, which #1569 keys on a destination a seat can type: `EscapeNotes` → `EscapeComposed`, which outside code also escapes the `!` of each `ast.Image`, at its call sites, so no composed seat text (mint, fix, inquiry, ruling, verify, evidence) opens one; `edit` and `ingest` refuse an `ast.Image` outside code in authored text — a new link in `composedRefusal`'s chain, named in `help/edit.md` and `help/ingest.md` (measured 0 outside code in 30 archives). 1c supplies the allowlist's image set — `placedPath` of each shown act; Part 2 adds `cache/crops/<crop_sha256>.png`. Every other shipped .md is written through the same decision (T13) |
| Seat-facing views [MODIFY] | `show report --assembled --json` carries each block in its section's `text`; bare `show report` (`report/seatread.go`) carries the anchor only, and `AssembledNote` (`layout.go:284`) says shown blocks are assembled content |
| `show evidence --anchor K-…` [MODIFY] | returns the act's shown output — the stored path for an image — beside each judgement and the sentence that judgement records (`Reproduce.claim`, 1a″): a shown figure's referring sentence is that sentence, the link from the report's text to the figure's anchor; `reproduce` help and `scripts/agentgen/src/red-lens-computation.md`: read the image, not the caption — a truncated axis is an unsound sentence; an SVG is read as its source, and the duty states what the source does not show (the rendered layout: overlap, clipping, font substitution) until #1573 gives the lens a renderer (T14) |
| Voice lens [MODIFY] `scripts/agentgen/src/red-lens-voice.md` | a shown text block is report text: read the report as assembled (R5). A finding's `--quote` must be report text, which a shown block is not: quote the anchor's sentence and name the block. The fix is `retire --anchor K-…` (the sentence stays), then compute again on the sentence (1a′) |
| evidence.md, help [MODIFY] | each live entry adds its shown output (1a′); tables are CSV, never printed markdown; a diagram can be a mermaid fence in `text`. `compute --help` (`image`): an SVG is shown as a static image (no script, hover or zoom); set `svg.fonttype: none`, so labels stay text the lens can read in the source |
| `devcmd/archdump` [MODIFY] | copies `computations/` beside the report set, so `diff -r` of two dumps compares shown images too |
| migrate | replay only: migrated computations are not shown |

### Part 2 — `--show` on `cite` (epoch E4) (rulings Crop, Colour)

Every image a seat states coordinates on — a PDF page, an image source (W2 finds no other) — is a
VIEW the model processes unresized: `(w, h) = ResizedSize(raster_w, raster_h)`, the largest view
the guide's algorithm leaves unchanged (§II), so box and recorded size are in the space the model
perceived on either tier.

| Change | Detail |
|---|---|
| `cite --show text` | the source span (`--source-span`) as a block quote, `Literal`; verification unchanged |
| `cite --show image --url U --page N` — phase 1 [MODIFY] | records nothing, and takes only what names the view: `--url`, `--page` (none for an image source). `--quote` and `--title` move from `seat.Require` (`cite.go:180`) to `seat.RequiredUnless(…, "unless --show image is given without --region")` (precedent `edit.go:239`). Returns the VIEW's path, width and height: a PDF page rendered in colour by `renderPagePNG` at `ResizedSize` of its full raster (`PageRenderDPI` :420) via `RenderPageInPixels`; an image source, the image box-averaged to `ResizedSize` of its own size when that differs (stdlib `image`, no new module). Width and height are read back from the PNG (`pngSize`, `pagerender.go:516`), never computed. Cached in `cache/`, keyed (source sha, page) |
| `… --region x,y,w,h` — phase 2 [MODIFY] | every phase-1 flag plus `--quote`, `--title`, `--region` (the box in the view's pixels). Re-derives the view from its key and records `crop_view_w`/`_h` from it; a box outside it is refused. Crops from the full-resolution colour raster (a PDF at `PageRenderDPI`; an image's own pixels), the box scaled per axis by raster ÷ view and rounded; `crop_sha256` = sha256 over the crop's width, height and decoded 8-bit RGBA pixels, row-major, in III.1.3's length-prefixed framing — not the PNG bytes, which an encoder change moves; stages and renames it to `cache/crops/<crop_sha256>.png`, records it, then returns its path for the seat to read; a wrong crop is `retire --anchor C-…`, which keeps the sentence, then a new cite: 1a′'s lifecycle, the same act as for a wrong figure. Never committed (R-commit); `archive rerun --citations` re-derives it and prints `differs` when its sha256 moved (1b table). `citeFields` writes `show`, `crop_page` and `crop_x…h`, and `citeAsserted` (`cite.go:198`) gains `--show` → `show`, `--page` → `crop_page`, `--region` → `crop_x`, `crop_y`, `crop_w`, `crop_h`, so a keyed retry with another page or box is a Conflict (W6); `crop_view_*` and `crop_sha256` derive from those |
| Colour | `renderPagePNG` (`pagerender.go:480`) takes a pixel size and the colour model in place of its DPI; its three callers — `RenderPages` :376, `RenderOnePage` :455, `readscanned.go:204` — pass the pixel size their DPI gives today and grayscale, and a view passes `ResizedSize` and colour, so OCR renders and `render-page` shas are unchanged. The comment `pagerender.go:298–308` is restated: OCR and `render-page` stay grayscale; views and crops are colour; the colour follows the consumer, so the record carries no colour field |
| Caption | `*Figure n.* Literal(<source title>, page N)`, or `Literal(<source title>)` for an image source — the citation's footnote text; `--reason` stays off the report (`cite.go:191–195`) |
| Absent crop | `cache/crops/<crop_sha256>.png` missing: report.md, report.html, `show report --assembled` and `devcmd/archdump` render "*Figure n.* <source title>[, page N]: image not held in this copy" (`Literal`, subject terms only — report-voice ruling); run.md's log names the absent crop (C-…, box) and `archive rerun`, which re-derives it |
| Evidence lens | `show evidence --anchor C-…` on a shown citation returns the crop and the view (derived into `cache/` on read; a derived crop whose sha256 is not `crop_sha256` is returned with both shas named); `verify` reads both — fidelity at the leaf, no seat program between |
| First test | `fetchcache` has no `image/*` branch (`grep 'image/' internal/fetchcache internal/cli/blue`: imports only); whether `cite` accepts an image source is measured before the image-source arm is built |
| migrate | replay only: migrated citations are not shown |

| Carrier (sweep W3, III.9) | Becomes |
|---|---|
| `cli/seat/help/cite.md` | adds SHOWING WHAT A SOURCE SHOWS: `--show text` displays your span; `--show image` is two calls — first with only the source and page, and the tool returns a view sized so you see it unresized, with its width and height; read it; then the same call with your quote, title and `--region x,y,w,h` in that view's pixels records the citation and returns the crop — read it; a wrong crop is retired with `--anchor` (the sentence stays) and cited again; a retry under one key restates the page and box |
| `cli/seat/help/verify.md` | a shown citation's crop is checked against the view of its page (`show evidence --anchor`): the same area, uncut, in colour |
| `scripts/agentgen/src/blue-researcher.md`, `fragments/blue-search-and-cite.md` | show a source's figure when the figure is the evidence the sentence rests on; state the box on the view the tool returned, in its pixels |
| `scripts/agentgen/src/red-lens-evidence.md` | fidelity duty: a shown crop that cuts a legend, axis or label that changes its reading is a finding |
| FEOV `docs/seat-command-triggers.md:85` (`blue cite` row) | names `--show` |
| `seatprobe/boards.go` | a board with a shown citation expects `red-lens-evidence` `show evidence` then `verify`; the OCR board (:563, `render-page` then `verify`) is unchanged |
| `help/render-page.md`, FEOV `docs/seat-command-triggers.md:57`, `cli/lens/verify.go:118` | unchanged: `render-page` stays the lens's grayscale OCR check, no part of `cite --show` (ruling Crop) |
| difftest `manual` golden, agent definitions | regenerated (`agentgen`, `UPDATE_GOLDENS=1`) |

### III.9 Consumer census

Pattern `P` = `(?i:proof|\bprov(?:e|ed|es|ing)\b|reproduc|not_?portable)|Prove(?=[A-Z(])|[a-z]Prove\b`
(PCRE; `provenance`, `improve`, `proven` do not match). Re-run at `e8d4fbfe` from the repository root
(`round6/census/<row>.txt`; C1–C3, W2 identical to the round-5 audit's lists; W5 supersedes C6);
every hit's meaning read and recorded in the ledgers. W rows: `F=plugins/frank-exchange-of-views`;
excluded `_test.go`, `.pb.go`, generated `agents/` and preloaded skills, goldens, testdata.

| # | Command | Hits | Consumers / false | Parts |
|---|---|---|---|---|
| C1 | `grep -rlP "$P" --include=*.go plugins/frank-exchange-of-views/tools \| grep -v _test.go` | 120 files | 76 / 44 (+`lens/finding.go`, `record/about.go`, `record/unsoundproof.go`, `migratecli/migratecli.go`; −`cli/migrate.go`) | 1a; also 1b 8, 1c 13, 2 2; III.1.4's carriers |
| C2 | same, `--include=*_test.go` | 203 files | 147 / 56 | 1a; `keycensus_test.go` `notFields`; new tests per part (§V) |
| C3 | `git ls-files \| grep -v '\.go$' \| grep -vE '^(run-archive\|plans\|research\|ideas\|projects)/' \| xargs grep -cP "$P"` (non-zero) | 95 files | 36 regenerated (11 agents, 2 preloaded skills, 7 difftest + 3 migrate goldens, `feov-migrate` help golden, `schema.sql`, 10 simulator, `vocabulary.md`); 34 authored (`round2/c3-nongo.tsv` + `fragments/lens-duties.md:5`); `site.html.tmpl:33`, terms.json; `payload-keys.txt` frozen; 22 English | 1a; 1b/1c/2 where marked |
| C4 | not `P` — rights: `git grep -niI 'tdm\|data-mining'` outside tests and archives | `shell.go:213–230`, `fetchsummary.go:110–112, :345–348, :354–357`, `fetchcache.go:110–117, :196–203` (facts, no premise — unchanged), `citations-and-bibliography.md:74–75`; plus `red-lens-voice.md` (1c), `help/reproduce.md:9`, `run-directory.md:28` (1a) | all | as marked |
| C5 | archives: `ls */proofs/`, `bases.py`, `grepdb.py`, NOT_RUN and sha-prose scans (`round4/m32/`) | 159 scripts; 246 events; 56 `not_run` in 7 runs (round-6 audit recount); III.1.6 residue counts | — | 1a |
| C7 | `grep -nP '(?i:\bprove\b\|reproduce\|proof)' releasegate/fuzz/fuzz_test.go` | 281 lines: prove :1899 :2720 :2739 :2745 :3265; reproduce :1959 :2069 :3768 :3785; `dirProveDrifts` :2934–:2940 (passes `--variance`); bases :5749–:5752; "unjudged proof " :2188 :2211 | all | 1a; a two-acts-one-bundle action 1a; `--input` 1b; `--show` 1c |
| C8 | `grep -rnP '"P-"\|P-…\|\bP-[0-9a-f]{4,8}\b\|\[\^P\|\bP-\b' --include=*.go tools` | 27 lines / 13 files non-test; 270 / 49 test | all | 1a (code: `anchor.go:63`, `flags/shapes.go:155,162,200`, `cli/seat/verbs.go:283,290,297,655,670`, `cli/lens/close.go:84`, `cli/lens/reproduce.go:144`, `record/refs.go:118`, `layout.go:580,591`, `report/notes.go:10–20`, `report/proofs.go:40,45`; the rest comments) |
| C9 | names across Go, md, js: `compute`, `computations`, `sources`, `rerun`, `rerun/`, `archive`, `acts`, `<bundle>`, `ran`, `rerun_exit`, `crop_*`, `crop_view_*`, `bundle_bases`, `not_compared`, `EscapeComposed`, `Literal`, `archiverun`, job `archive-rerun` | `"compute"` only as a refused `--check-kind` spelling (`cli/checkkind_test.go:146`); `sources`, `rerun`, `acts`, `archive` (tarball; also `fetchcache/backends.go:41`, `record/spotcheck.go:43`, neither seat-facing on an operator page) as round 3; `.tmp-*` reuses `fetchcache.go:462`'s meaning; `rerun_by` keeps its meaning (`reproduce.go:233`); the rest 0 hits | — | 1a–2 |
| C10 | `grep -rn 'tar\.NewReader' --include=*.go tools \| grep -v _test` | 2 (`recordtest/archive.go:66`, `archdump/main.go:159`) | both move onto `archive.Extract` | 1b |
| W1 | per-script re-run model: `git grep -nP '(?i)(re-?run\|rerun\|re-ran\|reran)[^\n]{0,100}(script\|every (proof\|one)\|stands for\|shares?\|standing)\|(script\|every proof)[^\n]{0,100}(re-?run\|rerun\|re-ran)\|scriptRerun\|standing_rerun\|RerunsOf' -- $F scripts/agentgen/src` | 125 lines (round 5's 116 + 9: `record.proto`, debate.js :1008, `debate.test.mjs:416`) | carriers in III.1.4's table (`record.proto:545–548`, debate.js :1008 and its pin new this round; `record.proto:536` English); rename-only (`prove.go:134`, `help/prove.md:13`, `report/docs.go:148`, `layout.go:171`, `red-lens-computation.md:11`, `:15`, `seatenv.go:74–80`, `cli/proofrerun.go`) in III.1.5/III.1.7; `migrate/entries.go:60`, `:283` describe the old input (unchanged); English (`capture` re-runnable, `debatejs.go:376`, `enums.go:306`, `record.go:994`, SKILL.md) | 1a |
| W2 | images a seat states coordinates on: `git grep -nIiP '\b(region\|bbox\|bounding.?box\|coordinat\w*\|crop\w*\|pixel box)\b' -- $F` | 63 lines | none seat-facing: `tessocr` internal geometry, enginegen/debate.js "region", research drafts — the class is Part 2's two views | 2 |
| W3 | Part 2 carriers: `git grep -nIP 'render-page\|renderpage\|\bpage image\|page.?s image' -- $F scripts/agentgen/src`; plus every `cite` row in help, agentgen sources, docs, seatprobe | 54 lines (the round-5 audit's 59 also counted `record.pb.go` ×3, `schema.sql` ×2 and a testdata label; this count adds debate.js :978, an OCR-check clause) | Part 2's carrier table; the OCR-check sites (`verify.go`, `render-page.md`, `ocr.go`, `pagerender.go`, `tessocr`, debate.js :978) unchanged | 2 |
| W4 | code read by hand: `` grep -rnE 'eachLine\(\|reCodeSpan\|HasPrefix\([^)]*"(```\|~~~)"' --include=*.go $F/tools \| grep -v _test `` | 13 lines: `eachLine`'s declaration, its toggle (`layout.go:354`) and 7 callers; `reCodeSpan` ×4 (comment, declaration, `notes.go:55`, `layout.go:634`) | all move onto goldmark's tree (1c Code regions) | 1c |
| W5 | what the renderer opens: `grep -rn 'ast\.Image\|ast\.HTMLBlock\|ast\.RawHTML' --include=*.go $F/tools \| grep -v _test` | 9 lines | `md.go:160` (HTMLBlock raw) and `:437` (Image; C6's one consumer) — #1569's allowlist, which 1c feeds its image set (1c Raw HTML and images); `md.go:260`, `:345`, `:442` print raw HTML as text; `footnotes.go:43` skips it; `anchor/sentences.go:128`, `:135`, `:252` find sentences (no rendering). Tool-written HTML in composed markdown, `grep -rnE '"[^"]*<(a\|div\|span\|sup\|br\|details\|summary\|img\|p)[ >/][^"]*"' --include=*.go $F/tools/internal/{report,view,anchor} \| grep -v _test`: `proofs.go:122` only (`md.go`, `chart.go`, `site.go` write the HTML page, not markdown) | 1c |
| W6 | keyed acts' asserted sets: `grep -rn 'RepeatsKeyed(' --include=*.go $F/tools \| grep -v _test \| grep -v 'func '` | 5 (`cite`, `prove`, `mint`, `finding`, `edit`) | `cite`, `compute` under `seat.AssertedCovers` (III.1.1); the other three have no field writer to drive | 1a, 2 |
| W7 | a reproduce's output copies: `grep -rn 'RecordedOutput\|ObservedOutput\|recorded_output\|observed_output' --include=*.go --include=*.proto $F/tools \| grep -v '_test\|\.pb\.go'` | 11 lines | `Reproduce` fields (`record.proto:1462–1463`, `reproduce.go:129–135`, `proofverdict.go:57`, `replay.go:532`, `remap.go:126`) → III.1.4, III.1.6; `EvidenceProofAtJSON.RecordedOutput` (`evidenceview.go:616`, `verbs.go:297`, `:725`) is the act's own stored output → `output` (III.1.5) | 1a |
| W8 | page rasterisers: `grep -rn 'renderPagePNG(' --include=*.go $F/tools \| grep -v _test \| grep -v 'func '` | 3 | all pass grayscale (Part 2 Colour) | 2 |
| L1 | kind-literal branches: `grep -rnE 'anchor\.Kind\([^)]*\) *[!=]= *"(citation\|computation)"\|Kind\(id\); k != "citation"' --include=*.go $F/tools \| grep -v _test` (re-run at `f5c09519`, `audit-r1/L1.txt`) | 3: `cli/blue/edit.go:268`, `cli/blue/retire.go:191`, `report/rawhtml.go:346` | edit and retire move to `anchor.Backs` (III.1′.1). `rawhtml.go:346` is the `<a id>` link target of the `WeaveComputation` assembly, which is intrinsic and unchanged | 1a′ |
| L2 | kind inferred from a side field or a bool: `grep -rnE '\bcomputation +bool\b\|\.computation\b\|Detail *[!=]= *""' --include=*.go $F/tools/internal/{record,capture,report} $F/tools/releasegate` (`rev-r1/L2.txt`) | non-test 24 lines / 9 files: `anchorchecks.go` 3, `available.go` 1 (`anchorOwnerOf`), `dispatch.go` 1 (`:1299`), `evidenceview.go` 1 (`:569`), `failedcheck.go` 4, `passblockers.go` 9, `sitting.go` 1 (`:413`), `unchecked.go` 3, `record.pb.go` 1 (false). The release fuzz infers a blocker's kind from `Detail` at 7 lines: `fuzz_test.go:1087`, `:2857`, `:2867`, `:2892`, `:3265`, `:3563`, `:4129` | `evidenceview.go:569` reads `state()` (1a′, III.1′.4); the rest read `anchorStanding.kind` and the `EvidenceTables` row, and the fuzz reads the blocker's kind field (1a″, III.1″.1, III.1″.7) | 1a′ (1), 1a″ (the rest) |
| L3 | the computation debt: `git grep -c 'awaiting_computation\|AwaitingComputation\|computation_answered\|ComputationAnswers\|NeedsComputation\|requireComputationAnswered' -- plugins scripts` (`rev-r1/L3.txt`) | 26 files / 76 lines. Non-test: `close.go` 4, `estoppel.go` 3, `views.go` 3, `replay.go` 2, `sitting.go` 2, `viewjson.go` 11, `agentgen/src/blue-researcher.md:113` 1. Test: `cli/lens/unreadable_test.go` 5, `record/checkkind_test.go` 5, `record/queries_parity_test.go` 4, `recordsql/viewfamilies_test.go` 4, `record/workstate_test.go` 1, `fuzz_test.go` 1. Generated: the manual golden 4, `schema.sql` 3, and the 11 `agents/*.md` (2 each, 3 in `blue-researcher.md`), which carry `awaiting_computation` in the OUTPUT shapes `agentgen` inlines from the tool's manual | renamed to the `evidence` forms, with no alias (III.1′.3); the 11 agent definitions and both goldens regenerate (`agentgen`, `UPDATE_GOLDENS=1`), and `agentgen -check` fails a stale one | 1a′ |
| L4 | `CHECK_KIND_SOURCE` readers: `grep -rn 'CHECK_KIND_SOURCE' --include=*.go $F/tools \| grep -v pb.go` | 0 | `Gap.AnswerKind` and `evidence_answered` through `EvidenceTables.check` become its readers | 1a′ |
| L5 | anchor retirement in agent text: `grep -n -i retire` over `cli/seat/help/*.md`, `scripts/agentgen/src/**`, debate.js (`skills/research-protocol/scripts/debate.js`; `audit-r1/L5-help.txt`, `L5-agentgen.txt`: 23 lines, none in debate.js's authored text) | 14 on anchors (`retire.md:9,11`; `edit.md:28`; `verify.md:11,15,19`; `reproduce.md:21`; `close.md:11`; `mint.md:15`; `blue-search-and-cite.md:14–18,28,31–32`; `lens-audit-and-mint.md:6`); 9 on claim or lens retirement (`dispatch.md:13`, `spot-check.md:7`, `blue-additive-only.md:1`, `blue-research-protocol.md:9`, both preloaded-skill sources, `red-chair.md:14`, `lens-duties.md:5`) | the 14 change (III.1′.5; `verify.md:11,15,19` and `lens-audit-and-mint.md:6` in III.1″.5); the 9 are unchanged | 1a′, 1a″ |
| L6 | the release fuzz drive: `grep -nE 'retire\|"cite"\|compute\|reproduce\|"close"\|contradiction' releasegate/fuzz/fuzz_test.go` (`fuzz-verbs.txt`, 48 lines; plus the contradiction drive) | `retire --anchor` only on the citation branch (:1095); an unsound computation is edited, never retired (:1111–1129); computation gaps are answered before close (:3920–3947, :3946, :3964, :3969). The contradiction drive: `verifyOutcomes` :2267, `answerContradictions` :2346, `raisePrinted` :2363–2369, `raiseAndMint` :2394, the reading's `names` :5234–5246 — under L4 it goes quiet on every anchored reading, so nothing would exercise a refuted citation's failed check | 1a′'s arms (III.1′.7); 1a″'s split of the reading drive, with a refuted citation driven through its failed check and its repeat refusal (III.1″.7). Each arm has a tally gate | 1a′, 1a″ |
| L7 | contradiction carrier: `grep -rn 'unansweredContradictions\|contradictionsOf\|ABOUT_KIND_CONTRADICTION' --include=*.go $F/tools \| grep -v '_test\|pb.go'` (`audit-r1/L7.txt`) | 26 lines / 10 files (`citationid.go` 11, `mintbudget.go` 4, `about.go` 2, `evidenceview.go` 2, `passblockers.go` 2, `verify.go`, `anchorchecks.go`, `migrate/replay.go`, `onegapperclass.go`, `record.go` 1 each) | each one narrowed to readings with no anchor; anchored readings move to `failedChecksOf` (III.1″.3); `migrate/replay.go:398` reads the translated findings (III.1″.6) | 1a″ |
| L8 | archives: the raise chain on an anchored reading (`rev-r1/tools/l4census.py` over the 33 archives migrated to epoch 50, `rev-r1/l4census.txt`) | 16 anchored `refutes`/`absent` readings in 7 archives; 11 contradictions they opened; 5 findings raising one (m30 1, m32 4); 2 mints `found_by` such a finding (`--quote`, `document`, closed); 0 mints about a contradiction; 0 mints about a citation anchor; 0 carried in L4's shape; 0 mixed contradictions | the translation and residue of III.1″.6, from these counts; V1a″.3 asserts them | 1a″ |
| L9 | readers of a check's recorded sentence: `grep -rn 'GetClaim()\|\.Claim\b' --include=*.go $F/tools/internal \| grep -v '_test\|pb.go'` (re-run at `f5c09519`, `audit-r1/L9.txt`) | 22 lines / 9 files. `Verify.claim`: `report/citations.go:28` (labelled corroboration filter), `:100` (the "of the sentence" line), `record/evidenceview.go:404` (verification JSON), `:434`, `:437` and `record/citationid.go:102` (a labelled corroboration's location), `citationid.go:533` (contradiction key), `anchorchecks.go:229`, `:264` (a corroboration's hold), `migrate/gapplace.go:81` (a corroboration's placement), `consistency/consistency.go:272` (placed ids), `cli/lens/verify.go:102`, `:459`, `:483`, `:510`, `:551`, `:696` (the verb's own write and output). `Retire.claim`: `report/assemble.go:933` (`withdrawnClaims`), `cli/blue/retire.go:207`, `:215`, `:474`, `:476` | K- counterparts (III.1″.2–III.1″.4): `:100` → `checkLines`; `evidenceview.go:404` → the judgement JSON's `claim`; `verify.go:102`, `:696` → `reproduce`'s write and human line; the ledger's `anchorCheck.sentence` (new reader of both). Corroboration-only, no counterpart (red never authors a computation): `citations.go:28`, `evidenceview.go:434–437`, `citationid.go:102`, `anchorchecks.go:229`, `:264`, `gapplace.go:81`, `consistency.go:272`, `verify.go:459–551`. `citationid.go:533`: corroborations only under ruling L4. `Retire.claim` is another field and unchanged | 1a″ |
| L10 | closure-history readers: `grep -rn 'closure_act\|close_event\|HasClosed\|ClosedEpoch\|closedAt\|\.Closure\b' --include=*.go $F/tools/internal \| grep -v '_test\|pb.go'` (`rev-r1/g1-sweep.txt`, 89 lines / 20 files), plus the Go folds over `*recordpb.Close`: `grep -rn '\*recordpb\.Close\b' --include=*.go $F/tools/internal \| grep -v '_test\|pb.go'` (16 lines) | the view and its columns (`views.go` 9); `gapStatesOf` and its 13 downstream readers; `convergence_at`; two Go folds deciding "closed" from `Close` events (`engagementOf`, `closingDutiesOf`); `priorClosureEpochs`; the oracle's raw walk. The other `Close` folds read the act, not the state: lineage (`successor.go:38`, `:96`), movement (`impasse.go:199`), acts on a gap (`gapacts.go:47`), migrate (`replay.go:429`, `proserefs.go:52`) — unchanged | every state reader re-derived (III.1′.3 Closure history) | 1a′ |
| L11 | writers of a closing `Close`: `grep -rn 'ClosurePayload(' --include=*.go $F/tools \| grep -v _test` | 2: `cli/lens/close.go:39`, `cli/chair/carry.go:18` (and migrate's replay, under `Migrating`) | the demand moves into `record.validate`'s `Close` arm (III.1′.1) | 1a′ |
| L12 | readers of an anchor's answers: `grep -rn 'GetAnswers()\|"answers"\|\.Answers\b' --include=*.go $F/tools/internal \| grep -v '_test\|pb.go'` (`rev-r1/g6-answers.txt`, 43 lines) | readers of `Computation.answers`: `gapacts.go:60–61`, `grounds.go:119`, `estoppel.go:284`, `views.go:569`, `:572`; the rest read `BlueEdit.answers` or a claim-kept `Retire.answers` (`blueanswered.go`, `manifestowed.go`, `impasse.go`, `estoppel.go:59`, `:122`, `edits.go`, `changes.go`, `changesjson.go`), or are the flag and its writers | the five read `evidence_anchor` (III.1′.3 Readers of an anchor's answers); the rest read neither kind's anchor answers, unchanged | 1a′ |
| L13 | seat-facing Go strings naming a retire or a failed check by one kind: `grep -rn 'citation retire\|retired citation\|citation retired\|retires the citation\|retire the citation\|Retire the citation\|retire a citation\|retires a citation\|retire of that citation\|retires one citation\|retire each citation' --include=*.go $F/tools/internal/{cli,record,view,report,capture} \| grep -v '_test\|pb.go'` (`rev-r1/g5-retire-strings.txt`, 42 lines; `audit-r1/go-strings-sweep.txt`, 44) | the evidence-anchor retire spoken as a citation's: `cli/seat/verbs.go:289`, `:295`, `:661`; `view/changes.go:48`, `:157–158`, `:190–194`; `record/changesjson.go:30`, `:55`; `record/blueanswered.go:57`, `:78`, `:168`; `record/impasse.go:195`; `record/record.go:1461`; `cli/blue/retire.go:80`, `:88`, `:158`, `:161–162`, `:167`, `:185`, `:192`, `:462–474`; `cli/blue/edit.go:257`, `:285`; `recordsql/views.go:307`. Citation-only by meaning (the `unread` state, which no computation has): `failedcheck.go:165`, `:199`; `passblockers.go:237`, `:248`, `:269–271`, `:427`; `cli/lens/verify.go:653`. `record/dispatch.go:21` names both kinds' remedies | the first set names the anchor by its row's `noun` (III.1′.2, III.1′.4, III.1′.5); the citation-only set is unchanged in 1a′ and gains the found-against arm in 1a″ (III.1″.5) | 1a′, 1a″ |
| L14 | `contradiction` in non-test Go: `grep -rn 'contradiction' --include=*.go $F/tools/internal \| grep -v '_test\|pb.go'` (`rev-r1/L14-contradiction-all.txt`, 167 lines / 31 files; seat-facing strings `rev-r1/g5-contradiction-strings.txt`, 38 lines) | the strings that describe an anchored reading's contradiction or its citation-standing answer: `cli/seat/verbs.go:296–297`; `record/passblockers.go:139–155`, `:343–366`; `record/dispatch.go:21`, `:1289–1291`; `cli/lens/verify.go:619–626`, `:676–686`; `anchorchecks.go:107`, `:143`; `record.proto` `AboutKind` means (`:1129`, `:1130`). The rest describe a contradiction as such and hold for a corroboration's | narrowed to corroborations, or moved to the failed check (III.1″.5) | 1a″ |
| L15 | readers of every committed archive: in `$F/tools/internal/record/migrate`, `(grep -l '"\*\.tar\.gz"' *_test.go; grep -l 'archivedRuns(t)' *_test.go) \| sort -u`, plus `grep -rln 'tar.gz' $F/tools/devcmd` | 20 migrate test files (Phase 0 lists them) and `devcmd/archdump`, `devcmd/archiverun`; `publish_test.go` names one archive | each takes up Phase 0's archive | 1a′ |
| L16 | hand-written UNIONs over the evidence tables: `grep -rn 'FROM "computation"' --include=*.go $F/tools/internal \| grep -v _test` and the same for `FROM "cite"` | `record/id.go:86–87`, `record/refs.go:130`, `:204`, `record/grounds.go:119–120`, `estoppel.go:284`, `views.go:569–572`; per-kind by design: `mintbudget.go:61`, `:64`, `viewjson.go:524`, `citationid.go:232`, `:278`, `refs.go:113` (a computation's bundle) | the first set reads `evidence_anchor` (One mechanism; III.1′.3) | 1a′ |
Named consumers, meaning checked: `capture.go:2035–2036` (archive contents, 1a; no `cache/`, kept
by R-commit); `pagerender.go:298–308` (restated in Part 2); `law/precedents.md:11–20`, `seat-surface-naming.md:806`
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
| R12 | **`archive rerun` runs archived programs with the operator's privileges, filesystem and network** | L × H × M | Stated in its `--help`; in CI it runs only in its own Linux job `archive-rerun`, modelled on `ocr-corpus` (:315), gated `if: github.event_name == 'schedule' \|\| github.event_name == 'workflow_dispatch'` (the workflow's `on:` also holds `push` and `pull_request`, :4–22): `contents: read`, `persist-credentials: false`, `setup-go` `cache: false`, `setup-node` `package-manager-cache: false` (as :569), `setup-python` with its `cache` input unset (its default caches nothing; it has no `package-manager-cache` input — its `action.yml`), no `actions/cache` — so nothing an archived program writes (`GOCACHE`, `GOMODCACHE`, `$HOME`) is saved for a pull request to restore; never the `release` job (`contents: write`) — ruling CI. Residue: a program can still read the runner's environment and reach the network |
| R13 | **Re-derived archived state (1a′, 1a″)**: L2 fells archived closures; L4 turns 16 archived anchored readings into failed checks no gap carries | L × L × L | Replay exempts refusals (`Migrating`). Each part's flip census (V1a′.6, V1a″.4) counts every change per archive and commits the counts |
| R14 | **Retire is cheap for a computation (1a′)**: blue retires an unsound `K-` and keeps a sentence nothing now backs | M × M × L | The same as for a citation today. The kept sentence is prose the logic lens audits, the claim count drops, and the CHANGELOG lists the retire |
| R15 | **One failed-check carrier changes the evidence lens's procedure (1a″, L4)** | M × M × M | the help, the `red-lens-evidence` source, a `seatprobe` board (a refuted citation expects the lens's gap about the anchor), the fuzz drive and its tally gates (III.1″.7) |
| R16 | **A generated SQL arm drifts from its table row (1a′)** | L × M × L | the `evidence_anchor` arms are built from `recordsql.EvidenceTables`; `TestEveryBackingKindHasAnEvidenceRow` checks each row against the DDL; the `schema.sql` golden |
| R17 | **A sentence change no edit names (1a″, L5)**: the ledger decides standing from `BlueEdit.reopened`, so an edit that changes an anchor's sentence without naming it leaves a check standing on an earlier wording | L × M × L | the consistency oracle's rule (III.1″.1 Staleness) fails such a record in every release-fuzz debate; in a real run, evidence.md prints the check's own sentence beside the anchor's, so the reader sees the mismatch. The oracle does not run in a live debate: residue |
| R18 | **The fall rule is stated twice (1a′)**: `closure_act.fell_event` in SQL, which every reader follows (the Go folds are handed its rows), and the oracle's raw walk, which re-derives it as a deliberate second opinion | L × M × L | the oracle's `replay-agreement` rule compares the two in every simulated debate of the release fuzz, whose drive fells closures (III.1′.7); V1a′.3's delete-the-row arms |
| R19 | **The demand moves to the write (1a′)**: `carry`, which wrote a `Close` with no demand check, is now refused on an unanswered demand gap; a close admitted by the CLI's check before a concurrent retire is refused inside the transaction | L × M × L | the refusal names every exit (`cite --answers`, `compute --answers`, `retire --anchor <about_ref>`, `motion docket`); `carry --help` states it; the fuzz tally's `sourceClosesRefused` and `fallenCarriesRefused` |
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
| the source body (`cache/`, `sources/C-…`, views, crops) | never | R-commit |
| a bundle file copied from a source (the only REPRODUCIBLE route in 1a) | yes, ≤ 1 MiB | accepted residue, as `proofs/` today |
| output and `out/` computed over a source (1b), TDM-reserved or not | yes, unbounded | accepted residue (R-rights), stated to the seat in the fetch summary's note |

## V. Verification Plan

Before any FEOV test: `unset $(env | grep -o '^FEOV_[A-Z_]*')`. Every part ends with
`go -C scripts run ./check` (fail count read), `go -C scripts run ./agentgen -check` and
`go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check`. Re-armed by: any edit
under `plugins/frank-exchange-of-views/tools`, `scripts/agentgen/src`, debate.js,
`.github/workflows/hooks.yml` (the `archive-rerun` job). Compute tests drive a fixture program the
test builds (`.exe` on Windows) and carry no Short gear, so they run on the Windows leg.

**The release fuzz is CI's, never a local run** (CLAUDE.md). Every part here changes the record, a
flag or an enum, so its pull request's path filter runs the four `release-fuzz` parts and
`release-fuzz-tally`. BEFORE calling a part done: read all five jobs' results on the PR's FINAL head
(the fail count in each log, not an exit code); a skipped fuzz step means the path filter did not see
the change, and that goes to the human. A red is compared with main's latest scheduled `hooks` run
(`gh run list --workflow hooks.yml --event schedule --limit 1`) before it is blamed on the branch;
when that run predates the branch's base, wait for the next one. Re-armed by: any push to the PR.

Tests of an act are driven the way the shipped path runs it: through the CLI verbs as `internal/cli`
tests drive a seat (`runAt`, `runMint`, `checkkind_test.go:20–60`), on a record the test builds by
those verbs, so the record's validate, its transaction and the seat's refusal envelope are all on the
path. A guard tested only as a function is not a test of the call (CLAUDE.md): each table below names
the call-site deletion that must fail it.

| # | Check | Fails if |
|---|---|---|
| V1a.1 | Rename commit alone: `protogen`, `schemagen`, `vocabdoc`, the shared gates; read every changed golden line and each `--help` before/after | a gate is red, or a help page lost a duty |
| V1a.2 | `internal/compute`: agreeing runs → `REPRODUCIBLE`; undeclared drift (output, exit, `out/` each) → refused; declared → `VARIANT` + drift; `--variance` over agreeing runs → refused; `../x` present at one depth → refused; a missing relative path printing lowercase `ENOENT`, exit 1 both runs → `REPRODUCIBLE`, exit 1 recorded; tool call → refused; symlink, > 1 MiB, `out`/`sources` entry, `__pycache__`/`.pyc` → refused; `./run.sh` (Unix) and `./prog.exe` (Windows) → run, toolchain absent; `./run.sh` on Windows → refused naming the start failure; under umask 077 and 022 the second act on the bundle finds `dir/` hashing to `bundle_sha`; an argv holding a newline and one splitting it differently → different shas; no copy remains, read-only `sources/` included | any arm passes the wrong way |
| V1a.3 | P-norm: a Python program printing `__file__` and `os.getcwd()` and raising, its copy under a symlinked temp dir → `REPRODUCIBLE`, recorded output holds `<bundle>` and neither copy's path; "No such file or directory" on stderr at exit 0 → refused. Delete the replacement, or the `EvalSymlinks` → the first fails | it is recorded otherwise |
| V1a.4 | Per act: a `VARIANT` bundle whose program prints and increments a counter file at an absolute path outside the bundle and both copies (in argv, so in identity); acts A (recorded 1) and B (recorded 3); counter reset to 3; `reproduce` A → `ran`, A `DIFFERS`, the counter now 4 (one execution); `reproduce` B → `ran` = false, B `REPRODUCED`, its `rerun_sha` A's; A's `acts/` byte-identical throughout; evidence.md shows A's own recorded output beside its re-run's; an act D recorded 2, compared against that standing re-run → `DIFFERS`, and evidence.md shows D's 2, never A's 1; a third act never compared reads `not_compared` in the evidence row and "not compared" in evidence.md, an unexecuted bundle `not_rerun`. An `out/` file changed with output equal → `DIFFERS`, drift names the file; exit changed → `DIFFERS`; absent recorded `out/` against an empty one → `REPRODUCED`; every live `DIFFERS` carries drift, every live reproduce `rerun_sha`, every `ran` `rerun_exit`; a `ran` = false whose `rerun_sha` is no longer the standing re-run's (a later `ran` appended between its read and its write) → refused. A `REPRODUCIBLE` act whose own comparison differs while the standing re-run's own verdict reproduced → `--as sound` refused. capture's audit over the same: one execution, A moved, B reproduced; mixed bases → `bundle_bases` and the audit detail. Delete `compare`'s act argument → B fails; read evidence.md's outputs from the standing event → B's display fails | B's verdict or display reads A's output, an act changes after write, or a no-comparison act reads as compared or unexecuted |
| V1a.5 | Store lifecycle (III.0, III.1.4): (a) a `REPRODUCIBLE` bundle whose re-run differs: `--as sound` refused → no `rerun/` entry, no `.tmp-*`; then `--as unsound` → one `ran` event and exactly one `rerun/<its rerun_sha>/` holding that execution's output and `out/`, `rerun_exit` on the event; (b) a test seam fails the append after the rename → an unnamed directory, no event; the next `reproduce` executes (`ran` = true), finds the directory present with the same bytes, and stands; (c) two `reproduce` processes released together on an unexecuted bundle → two events; a later `ran` = false act follows the LATER event's `rerun_sha` (give the two executions different outputs: tamper the earlier's directory → verdict unchanged; the later's → it moves); (d) capture's audit over a bundle without `rerun/` and with a tampered `rerun/<id>/output`, and `archive rerun` on an extracted archive holding a tampered `rerun/` → identical verdicts, the `computations/` listing hash unchanged; (e) two lanes writing one bundle → one `dir/`, both acts recorded; a pre-existing `acts/<K-id>/` → refused; a refused append leaves no `acts/<K-id>/`; `archive.Write` → `tar tzf` holds each recorded `rerun/<id>/` and no `.tmp-*` or unnamed directory; (f) migrate of a run whose replay renumbers a `ran` event → its `rerun/<rerun_sha>/` unchanged and still resolved. Resolve `rerun/` by listing (newest directory) → (c) fails; write in place without staging → (a) fails | any arm passes the wrong way |
| V1a.6 | Keyed retry: different `--variance` → refused; different argv / file bytes → Conflict; identical → idempotent, nothing written. `seat.AssertedCovers` over `compute` and `cite`: drop any one entry from `computeAsserted` or `citeAsserted` → it fails naming the flag. Record write: `VARIANT` without drift or invariant, `REPRODUCIBLE` with either, `show` = `SHOW_UNSPECIFIED`, `drift` on `REPRODUCED`, `DIFFERS` without `drift`, `rerun_exit` without `ran` or without `rerun_sha` → refused by the `(check)`; a migrated `ran` event with neither → accepted | any is recorded, or the migrated shape refused |
| V1a.7 | migrate tests on the archives of III.1.6, each assertion and residue row (m32's `not_run` rows included). Then `go -C plugins/frank-exchange-of-views/tools run ./devcmd/archiverun`: each of the 30 archives through `archive.Extract`, migrate and `archive rerun`, every act's outcome against `devcmd/archiverun/testdata/outcomes.tsv` (159 rows: archive, old sha, outcome). The file is MEASURED, not forecast: `archiverun -write` in the `archive-rerun` job on `workflow_dispatch` uploads it as an artifact and it is committed — interpreters pinned by `setup-python`/`setup-node`, their versions in its header. The dev-box probe's 109 / 48 / 2 had no pins and one copy; m19's `10ef5e65` reads `/home/…/scratch/universe-m19/…` by absolute path and is expected to differ on the CI image. Off the CI image the outcomes check fails by design (other interpreters, no pins): a dev box runs the migrate half only. The 13 migrated `DIFFERS` carry `drift` "on re-run the program wrote: …" with their program bytes unrespelled. Nightly in the same job on `schedule` (R12); never in `release` (a workflow edit: CI's `actionlint` reads it) | a refusal, a failed assertion, a header version mismatch, a missing interpreter, or any act's outcome other than its committed row — a reproducible act's exactly; a variant act's re-run may or may not write its recorded output, so for it `reproduced` and `moved` stand for each other |
| V1a.8 | Delete-the-row: each pinned env var, each `remap` admission (`P-` in `hexID`, `archivedToken`, `[^Pn]`, `not_run`, `about_kind proof`), each refusal arm, each `notFields` entry — a test fails for each | a deletion leaves the suite green |
| V1a.9 | Smoke: `scripts/universe.sh build`, grep the staged marketplace for `blue compute`, `--smoke` on a numeric question; capture's `computation-rerun` | the run recorded no computation, or a FAIL |
| V1a′.1 | Lifecycle, ONE table-driven test over every `backs` kind (`C-`, `K-`) running the same script through the CLI. Place with `--answers G`. An edit leaving the anchor bare is refused, naming `retire --anchor`. `retire --anchor` alone keeps the sentence: `claim_kept`, `anchors` = [id]; the retire's result, `show changes` and CHANGELOG name the anchor by its kind's noun ("retired computation K-…"). report.md, evidence.md and `show evidence`'s lists and counts then hold no entry, and `show evidence --anchor` says `retired`. `verify` and `reproduce` of it are refused, and so is a bare one from a migrated record. An edit carrying it to other prose reopens it. A sentence with one `K-` counts one claim (L3). Delete-the-row: remove `computation`'s `EvidenceTables` row → `TestEveryBackingKindHasAnEvidenceRow` fails; set `K-` `backs: false` → the script fails on its `K-` arm; restore `anchor.Kind(id) == "citation"` in edit or retire → the `K-` arm fails; drop the retire filter from `evidenceDoc` or `citationsSection` → the render arm fails; hard-code "citation" in `changes.go:194` → the `K-` label arm fails | any arm passes the wrong way for either kind |
| V1a′.2 | Closure demand at the write, per check kind, through `close` and `carry`. A close of an unanswered `computation` gap and of an unanswered `source` gap is refused under each closing disposition in turn — `repaired`, `repaired_with_regression` (with `--superseded-by`), `amends_prior`, `not_a_defect`, `defect_accepted`, `defect_owed_elsewhere`, `moot` — and the refusal names `cite --answers` / `compute --answers` and `motion docket`; a regrade of it is admitted. A live answer of the demanded kind admits the close; one of the other kind does not. **Gap 3's case:** close a `computation` gap answered by `K-a`, retire `K-a`, then the chair's `carry --carried-from <the close's epoch>` → refused, and the gap stays `open` and `awaiting_evidence`; a `carry` restating a standing closure whose answer is still live is admitted. The bench's closing ruling on an unanswered `source` gap (filed by `motion docket`) closes it. L1: a gap about `K-x` minted `--check-kind computation`, and one about `C-y` minted `--check-kind source`, each after `retire --anchor` of its anchor → red's close admitted, nothing auto-closed. Ruling L1-late: a close resting on an answer, then `retire --anchor` of the `about_ref`, then of the answer → the closure stands (`fell_event` NULL, gap closed); the reverse order, the answer retired first → it falls at that retire and stays fallen when the `about_ref` is retired later. Blue's sitting is held on an awaiting `source` gap as on a `computation` one. Delete-the-row: move `requireEvidenceAnswered` back into `cli/lens/close.go` only → the `carry` arms fail; drop the `Migrating` skip → V1a′.6 fails on its first archived prose closure; exempt `moot` → its disposition arm fails; drop the `about_ref` arm → the L1 arms fail; drop `priorClosureEpochs`'s standing filter → the carry arm's refusal no longer names the standing epochs and fails | any arm passes the wrong way |
| V1a′.3 | Closure history (rulings L2, L2 history), one script through the CLI. A lens mints computation gap G and source gap H; blue answers G with `compute --answers G` and H with `cite --answers H`; the lens closes both in epoch 2; blue retires both answers in epoch 3. Then every reader says G and H are open: `show board --json` (`open`, `awaiting_evidence`, no closure payload, no `closed_epoch`), blue's work list (both owed), `board_counts`, `convergence_at` at the newest point (their mass in `mass`, `material_open`), telemetry (`closedAtR` for epoch 2 holds neither), the scorecard's durable-repairs rows, the board chart's closed buckets, the spot-check archive, the dashboard model, `declarations`, capture's closure claims, `engagementOf` (a blue sitting dispatched on G before the fall and sitting after it lists G open, not found closed), `closingDutiesOf` (a docket filed on G before the fall owes its closing after it), and `consistency.Check` reports no violation. At the close's own point `convergence_at` counts both closed. A new answer → still open (no auto re-close); a new close in epoch 4 → closed, `ClosedEpoch` 4, telemetry counts it once, in epoch 4, and the repair denominator counts it once. With two answers live at a close, retiring one → still closed, both → open. A bench closure after a fallen red close stands (the fixture's G-221322e1 shape); a close whose demand L1 met at the close stands when its answer is retired (G-a0c6ca1e's shape). Delete-the-row: drop `fell_event IS NULL` from the `gap` view → the board arm fails; from `convergence_at`'s `closed` CTE → the newest-point arm fails; hand `engagementOf` or `closingDutiesOf` the raw `Close` events → its arm fails; remove the raw walk's L2 arm from `internal/consistency` → `consistency.Check` reports `replay-agreement` and the oracle arm fails; read "answered now" instead of "answered at the close" → the no-auto-reclose arm fails | any reader counts a fallen closure, or a re-close twice |
| V1a′.4 | Retry, cross-kind and answer readers. A keyed idempotent `compute` retry returns `computation_id`. `cite` keyed retries with another `--answers` → refused; `seat.AssertedCovers` over `cite` includes `answers`. `--cites` a retired citation → refused. `retire --anchor C-` named by two live `K-`s' `cites` → refused naming both (ruling L6); still refused after one is retired, naming the other; admitted after both are retired; nothing else is retired by it. The race: two `feov-record` processes, one `compute --cites C`, one `retire --anchor C`, released together twenty times on fresh records → each time exactly one admitted, and no record holds a live `K-` citing a retired `C-`. **Gap 6's case:** a `source` gap remanded at impasse, then a blue sitting whose only act on it is `cite --answers G` → the remand's exchange begins (`gapActsOf`, `impasse.go:160`), and a lens sitting after it is not refused for an unbegun exchange; a citation placed `--answers G` after a ruling on G is new grounds about G (`motion docket --grounds C-…`). Delete-the-row: move the L6 check out of `record.validate` into `cli/blue/retire.go` → the race arm fails; drop the `Cite` arm from `gapActsOf` → the remand arm fails; drop it from `groundsActSQL` → the grounds arm fails | any |
| V1a′.5 | Render goldens. Two documents referencing `K-a`, `K-b` in opposite orders → each numbers `[^K1]`, `[^K2]` by its own first appearance, and a retired `K-` takes no number. evidence.md holds no retired `C-` or `K-`. "anchored to nothing" appears only for a live computation no document references. `show changes` of a `K-` retire reads "retired computation K-…, its sentence kept". Read every changed golden line (difftest `manual`, report, evidence) | a retired anchor rendered, a gap in numbering, a retire named by the wrong kind, or a lost duty on a help page |
| V1a′.6 | Migrate at E1′. All 33 archives (the 32 committed before 1a′ and Phase 0's): zero refusals; `Cite.answers` absent. The flip census, written by a `devcmd` (`devcmd/flipcensus`) and committed as testdata: per archive, each fallen closure (gap, close event, fall event), each gap whose `open` or `awaiting_evidence` differs between E1 and E1′, the claim count before and after, and each never-retired bare anchor that answers a gap (the two-lives residue). Measured, not forecast; a change in a count fails the test | a refusal, a count off its committed row, an uncounted flip |
| V1a′.7 | Real data: the 2026-10-30 CEILING. (a) The smoke's own archive, `run-archive/2026-10-30_mersenne-31-prime-sum.tar.gz` (Phase 0), extracted with `recordtest.ExtractArchive` and brought through `feov-migrate` from epoch 49 to E1′. Its evidence.md holds none of `K-414ee920`, `K-72d5d71a`, `K-71a99b9a`, `K-bc8ac2b6` and no "anchored to nothing". `G-342d3cdc` and `G-78817cfa` (about `K-71a99b9a`, `K-bc8ac2b6`, retired at 722 and 779; bench-closed at 881, 882): on the final state `gap.evidence_answered` is true for both; on the record's prefix through event 880 (the migrated events before 881 replayed into a fresh record under `Migrating`), red's `close` of each through the CLI is admitted (L1). `G-a0c6ca1e` (about `K-71a99b9a`, retired 722; red `moot` close 752; its answer `K-bc8ac2b6` retired 779) is closed on the final state: L1 met the demand at the close, so the closure stands (ruling L1×L2). `G-221322e1` (red close 701 on `K-71a99b9a`; bench close 792) is closed on the final state; its red close has `fell_event` 722, and `convergence_at` counts it open at a point between 722 and 792. (b) A scripted replay through the CLI at E1′. `compute` `K1` on S; the lens mints a computation gap G on S. Blue's edit moving `K1` to a bare position → refused. `retire --anchor K1 --answers G`, then `compute --answers G` on S → red's close of G admitted, and no computation gap stays open on a retired `K-`. A `source` arm: `cite` `C1` on T; the lens mints a `source` gap H on T; blue rewords T (C1 stays) → red's close of H refused, naming `cite --answers`; blue cites on the reworded T `--answers H` → admitted. (c) `scripts/universe.sh build`, grep the staged marketplace (`$WORKDIR.marketplace/plugins/frank-exchange-of-views`) for `retire --anchor` on a computation; then `--smoke` on the same question. The run records no close refusal of a gap whose anchor was retired, evidence.md holds no retired anchor, and capture's `assembly-screen` and `computation-rerun` PASS | any; or (c) ends CEILING on a gap held by a retired anchor |
| V1a′.8 | Agent text. `agentgen -check`, `enginegen -check`, `vocabdoc`; each changed `--help` read before and after. A failing assertion, not a golden a person reads, over the rendered pages: `retire --help` (any evidence anchor; the L6 refusal; a retire fells a closure), `edit --help`, `reproduce --help` (no live sentence, refused), `close --help` (every disposition; `motion docket` as the exit for a demand no longer held; a reword leaves a `source` gap unanswered until a citation answers it), `carry --help` (a standing closure only), `cite --help` (`--answers`; cite again on a reworded sentence), `show changes --help` and `show evidence --help` (evidence-anchor retires; no retired anchor in the rows); and the generated `blue-researcher` definition holds the cite-as-the-answer act. Drop one sentence → it fails | a stale page or a lost duty |
| V1a′.9 | The release fuzz on 1a′'s PR, read as §V's preamble states: the four parts and the tally green on the final head, and the tally's 1a′ gates (III.1′.7: `computationRetires`, `sourceClosesRefused`, `sourceAnswers`, `closuresFallen`, `fallenCarriesRefused`, `citedRetiresRefused`) each above zero | a red part or tally, a skipped fuzz step, or a gate at zero |
| V1a″.1 | Failed check (ruling L4), through the CLI. A standing `refutes`, and a standing `absent`, on a `C-` → `stateFailed`, PASS held, no contradiction opens, and the result's `failed_check` names the reader and the gap owed; the reader's `mint --about-kind anchor --about C-…` in that sitting carries it, outside the budget; a repeat refuted reading while that gap is open is refused, as a repeat `unsound` is. A corroboration's `refutes` opens a contradiction as today. `assembly-screen` FAILs on a live unsound `K-` and on a live refuted `C-`, and PASSes once each is retired. Blocker words: a computation blocker whose `Detail` is empty still names `reproduce` and its owner. `seatprobe`'s refuted-citation board expects the mint about the anchor. Delete-the-row: let an anchored reading open a contradiction again → the no-contradiction arm fails; drop `refutes` from `stateFailed` → the PASS arm fails; read kind from `Detail` again → the blocker-words arm fails | any |
| V1a″.2 | Recorded sentence and staleness (rulings L5, Staleness), ONE table-driven test over `EvidenceTables`' `checkTable` rows (`verify`, `reproduce`) running the same script. A check of a live anchor records `claim` equal to `record.LiveSentence`; the seat has no flag to set it (`seat.Supplies`); `reproduce`'s result carries `claim` and its human line quotes it. evidence.md: the `K-` entry prints `backs` as a `C-` entry does, and each check appears through `checkLines`. An edit changing the sentence (naming the anchor in `reopened`) → the check no longer stands for either kind, evidence.md prints "of the sentence: <old>" under it, and a fresh check from a later sitting prints no sentence line. Oracle: a test seam drops the anchor from `BlueEdit.reopened` on that edit → `consistency.Check` reports one violation naming the anchor, the check's event and both sentences, for each kind. A migrated reproduce (no `claim`) prints no sentence line, stands or not by the ledger alone, and the oracle counts it as unmeasured. Delete-the-row: drop the `Reproduce.claim` write → the `K-` arm fails; drop the oracle rule, or its `computation` row → the forged arm fails for that kind; restore `readings` as citation-only → the `K-` render arm fails; class `Reproduce.claim` other than `data` in remap → `remaplists_test.go` fails | any arm passes the wrong way for either kind, or a disagreement passes the oracle |
| V1a″.3 | The L4 translation on the real chains. Over the 33 archives at E1″: zero refusals; exactly 5 findings translated — m30 `F-8410bec7` and m32 `F-d8329c9a`, `F-1c990edb`, `F-b59925d3`, `F-5e7dfd3d` — each now `about_kind` `anchor` naming its reading's citation (`C-4f4b8cfa`, `C-a36dce3f`), and no other finding moved; `G-dc877265` and `G-f417112d` keep their location, gap marker and `found_by`; no corroboration's contradiction changes. Delete the translation → m30 and m32 are refused at replay (`requireOpenContradiction`), naming the finding | a refusal, a count other than 5, a moved mint |
| V1a″.4 | The flip census at E1″, by V1a′.6's `devcmd`, committed: per archive, each anchored reading (16) and whether its failed check stands at the archive's end, uncarried (III.1″.6 residue); the reproduces with no recorded sentence; the contradictions that no longer open. A change in a count fails the test | an uncounted flip, a count off its row |
| V1a″.5 | Agent text. `agentgen -check`, `enginegen -check`, `vocabdoc`; each changed `--help` read before and after. Failing assertions over the rendered `verify --help` (a reading found against a citation is a failed check carried by your gap about it; contradictions are a corroborated source's), `mint --help`, `close --help`, `reproduce --help` (the sentence is recorded with your judgement), `show evidence --help` (`unanswered_contradictions` rows carry no anchor), the `dispatch` plan's `gap_ids` description in debate.js, and `verify`'s human line for a refuted citation. Drop one → it fails | a stale page or a lost duty |
| V1a″.6 | The release fuzz on 1a″'s PR, read as §V's preamble states, with the tally's 1a″ gates (III.1″.7: `refutedAnswers`, `refutedRepeats` above zero; `stalenessForged` = `stalenessReported`) and the drive's own failures (a refuted citation printing a contradiction, or no `failed_check`) | a red part or tally, a skipped fuzz step, or a gate off |
| V1a″.7 | Smoke: `scripts/universe.sh build`, grep the staged marketplace for the failed-check sentence of `verify --help`; `--smoke` on a question with sources. Read `show evidence`: no `unanswered_contradictions` row carries an anchor; any refuted citation has its reader's gap about it | a contradiction on an anchored reading, or a refuted citation with no gap after its reader's sitting |
| V1b.1 | `--input`: raw bytes at `sources/C-…`, read-only, in both copies; a citation not on the record / not cached / `recovered` with no `copy_url` → refused; a different input under the same key → Conflict; a zero-input bundle hashes to the sha pinned in 1a's test | any |
| V1b.2 | `archive rerun` on a fixture archive with no `cache/` against a local server: correct bytes → `reproduced`, exit 0; a `followed_link` citation re-fetched from its `copy_url`, not its `url`; a `VARIANT` that moved → `moved`, exit 0; a `REPRODUCIBLE` that differs → exit 1; changed bytes → refused naming url and both shas, exit 1; 404 → exit 1; each act line names its inputs verified. `archive.Write` → `archive.Extract` round-trips a `./run.sh` bundle that then reproduces; C10 finds `tar.NewReader` only in `internal/archive`; `archive rerun --help` names operator privileges and network | any outcome, exit or line differs |
| V1b.3 | Real data: V1a.9 on a question with a CSV source; extract its archive to a clean scratch dir, `feov-migrate --from … --to …`, `feov-record archive rerun` | the summary's count of computations with an input re-fetched, verified and re-run is below 1; or a non-zero exit |
| V1b.4 | Rights: a fetch summary of a TDM-reserved page prints the note stating computation is accepted and its output committed; each C4 comment carries the decision | the note or a carrier says otherwise |
| V1c.1 | Assembly goldens for each `Show`: report.html carries `data:image/` and no external `src`; delete each `Show` case → a test fails | a block missing, misnumbered, or external |
| V1c.2 | Images: `edit` and `ingest` adding `![x](y)` outside code → refused, naming the refusal `help/edit.md`/`help/ingest.md` teach; inside a code span or fence → accepted. `![x](y)` in a mint, fix, inquiry, ruling and verify text → escaped in report.md, text in report.html, also when `y` is a PLACED store path. Raw HTML (#1569's decision, carried into report.md): #1569's four probes through real assembly, read in report.md as goldmark parses it — no raw `<script>`, `<iframe>` or external `<img>`, every later section present; `<img src=…>` in authored and composed text → text in both outputs; quadratic-formula's `verify.text` row 683 migrates and renders as text; evidence.md's `<a id>` anchors survive and every `evidence.md#K-…` link lands, while a seat-typed `<a id="x"></a>` is text. Delete the `!` escape, the image refusal or report.md's `<` pass → its forged case fails; all 30 archives still migrate with zero refusals | accepted, embedded, raw, or a migration refusal |
| V1c.3 | `show report --assembled --json` holds the table after the anchor's paragraph; bare `show report` holds none; `show evidence --anchor K-…` of a shown act returns the shown output with each judgement and its `claim`, the sentence holding the anchor; the `red-lens-voice` agent golden holds the assembled-read and quote-the-sentence duties | either differs |
| V1c.4 | Smoke as V1b.3 on a question whose answer is a table and a figure; read report.md and report.html | no shown block |
| V1c.5 | Keyed retry with a different `--show` → refused; `out/` with 0 or 2 files (`table`, `image`, `text`), a CSV cell holding a newline, an undecodable PNG → refused | any is recorded |
| V1c.6 | Composition: each of `![x](computations/<sha>/acts/<K>/out/f.png)` naming a real placed file, an unbalanced `[`, `[^K1]`, `\|`, a line break, `*a*`, a backtick — as `--reason` (table and figure captions), as a CSV cell, as a source title (Part 2 caption and placeholder), as a source span (`cite --show text`) → each figure embedded exactly once, no footnote reference but the weave's, the table's column count intact, the rendered caption text equal to the input with line breaks folded; a bundle file and an output holding a ``` line → evidence.md's fence holds. Code regions: an output holding a ``` line, then a later entry whose `--reason` holds `![x](computations/<sha>/acts/<K>/out/f.png)` naming a real placed file and `[^K1]` → neither opens and the fenced bytes are unchanged (no backslash); a section ending inside an open fence, then a composed section holding a placed `![x](…)` and `[^K1]` → neither opens in report.md (units join closed); `compute --show text` and `edit` with a four-backtick fence holding ```` ``` ```` then an image after the close → refused; a `## Bibliography` heading after that close → refused by `ComposedIn`, the same heading inside it → accepted; an indented code line holding `[^K1]` → unescaped. Names: `out/` files `a b).png`, `x](y.png` and nested `sub/f.png` → each embedded exactly once at its `placedPath`, which resolves to the file; a bundle path and argv holding `](`, `` ` `` and `<` → literal in evidence.md. Remove `Literal` at any one site, or restore `eachLine`'s toggle → a test fails | a forged figure, link, raw HTML or footnote, a broken block, or a backslash inside code |
| V2.1 | `ResizedSize` reproduces the guide's `1075×1520 → 924×1307` and `1920×1080 → 1456×819`; over a grid around each limit (edges 1500–1600, token products 1500–1600) the record check accepts `(w, h)` exactly when `ResizedSize(w, h) = (w, h)`. A fixture letter PDF page with a red rectangle at known points, read at ≥ 300 DPI (raster ≥ 2550 × 3300). Phase 1 with `--url` and `--page` alone records nothing and returns a colour view of `952×1232` = `ResizedSize(raster)`, which `ResizedSize` leaves unchanged, its stated width and height equal to the view PNG's. A box computed from the rectangle's points in VIEW pixels, as the model states it → phase 2 records `crop_view_w`/`_h` equal to the view's and returns a crop from the full raster covering the rectangle (edges within ⌈raster ÷ view⌉ px), red channel dominant; `render-page`'s image stays single-channel with an unchanged sha; a box outside the view → refused. Image-source arm (after First test): a 3000 × 2000 PNG → view `1344×896`, same property. Phase 2 records `crop_sha256` equal to the returned crop's; a keyed phase-2 retry with another box or page → Conflict; `archive rerun --citations` with the crop renderer altered by a test seam → `differs`, exit 1. Size the view by long edge alone → the record refuses `1212×1568` and the test fails; record the raster's size or skip the scale → the crop misses and the test fails; `verify` reads crop and view | grayscale, a different area, a view the algorithm would resize, a size not read from the view, a missing or wrong crop sha, or a changed `render-page` sha |
| V2.2 | Cache entry deleted: report.md, report.html, `show report --assembled` and archdump each hold the placeholder — subject terms only, `Literal`: figure number, source title and page, "image not held in this copy" — and run.md's log names the absent crop and `archive rerun`; `run-archive/` holds no crop (`tar tzf`). Delete the placeholder branch → a test fails | silent omission, or a committed crop |
| V2.3 | Smoke: a question with a PDF source carrying a colour figure; read report.html and the seat's crop | no shown crop, a gray one, or a crop that cuts the figure its sentence names |
| V* | `/prosthetic-conscience:plan-audit plans/compute-and-show.md` | FAIL |

**Required tests carried from audit (rounds 6 and 7)** — local gaps, approved as tests (R-approve);
each lands with its part and fails if its property breaks.

| # | Part | Property (fix) | Test |
|---|---|---|---|
| T1 | 1a, 1c | Every program-derived string printed into evidence.md or report.md — `Computation.drift`, `Reproduce.drift` live and migrated (≤ 400 bytes of output holding `====` lines and indented tracebacks), `out/` names — goes through `Literal` or a fence longer than its longest backtick run; every placed destination resolves (`placedPath` escapes per segment, III 1c Assembly) | a `VARIANT` whose `out/` file name holds a backtick, a newline and `## X` and differs between the runs; a live `Reproduce.drift` naming an `out/` file whose name differs between the act and its re-run; a migrated quadratic-formula `Reproduce.drift`; evidence.md's heading set and `<a id>` anchors equal those of the same run without them; nested `out/sub/f.png` linked to a path that resolves |
| T2 | 1a | `keycensus_test.go:85`'s `output` reason restated (III.0) | the census test passes and asserts the entry's reason is the restated one (the act's output in `acts/<K-id>/output`); drop the entry → it fails naming `output` |
| T3 | 1a | evidence.md's preamble (`proofs.go:111–115`) speaks the two bases and `computations/<bundle_sha>/` (III.1.5) | evidence.md golden: no `observed`, `proof`, `proofs/` |
| T4 | 1a | `Reproduce.drift` is remap class data: migrated program bytes are not respelled (III.0) | a migrated drift holding a `P-` + 8-hex string from program output keeps it; class it argument → the test fails |
| T5 | 2 | `renderPagePNG` takes a pixel size across its three callers (`pagerender.go:480`; Part 2 Colour) | OCR and `render-page` goldens unchanged; a view is exactly `ResizedSize` pixels (V2.1) |
| T6 | 1c | evidence.md's `<a id>` — two inline `RawHTML` nodes in an `htmlOnly` paragraph — is a shape #1569 renders | V1c.2's anchor arm |
| T7 | 1c | a shown image in evidence.md is composed after the escape (Composition rule) | evidence.md golden: the tool's `![Figure n](placed)` embeds; a seat's `![x](placed)` in `--reason` is escaped |
| T8 | 1a | `ran` = false against a standing `CALLS_THE_TOOL` records `rerun` = `CALLS_THE_TOOL`, no drift; `sound` refused (III.1.4) | a V1a.4 arm whose standing execution is `CALLS_THE_TOOL`: a `ran` = false act records `rerun` = `CALLS_THE_TOOL`, no drift, and `--as sound` is refused |
| T9 | 1a | a migrated `ran` = false copying a standing `DIFFERS` copies its drift (III.1.6) | a synthetic archive (the 30 hold no case, so V1a.7 alone is vacuous) migrates with zero refusals and the check holds |
| T10 | 1c | `help/edit.md` and `help/ingest.md` name the one refusal 1c adds (an image outside code) | a failing assertion, not a golden a person reads: the rendered `edit --help` and `ingest --help` contain the image refusal's text; drop it from either page → the test fails |
| T11 | 1a | C5's `not_run` count is 56 (III.9) | V1a.7 counts migrated `ran` = false events per archive against the ledger |
| T12 | 1c | Unit-boundary parser state beyond fences — link reference definitions apply document-wide, before or after use; an indented unit after a list item continues it; definitions sit in footnote continuations. In the shipped document's parse, an image node at a placed destination and a weave-label footnote reference exist only where assembly put one; #1569's whole-document decision is the natural home: its image arm checks placement, not only destination | through real assembly, in authored and composed units: P1 (a definition `[r]: computations/…/out/f.png` in one unit, `![x][r]` in another); Q1 (`- item`, then a unit `    ![x](placed) [^K1]`); Q2 (the definition after its use); Q3 (the definition inside a footnote continuation) → nothing embedded and no weave-label reference in any output |
| T13 | 1c | Every shipped .md (`docs.go:386–393`: report, docket, debate, judgments, inquiries, evidence, run, CHANGELOG) is written through #1569's decision, or the plan states the cut | V1c.2's four probes, per document |
| T14 | 1c | The computation lens reads an SVG as its source and its duty states what it cannot see there, until #1573 gives the lens a renderer (1c `show evidence` row) | the `red-lens-computation` agent golden holds that sentence |

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
| R-html-split (gblock, 2026-10-24) | Who fixes seat raw HTML in report.html | #1569 on main: a render-time allowlist on the whole shipped document; 1c depends on it and drops its per-site HTML escaping | Per-site escaping in 1c (round 6: fence parity across units and a block's later lines defeat it, `audit-r6/probe*-out.txt`) |
| R-approve (gblock, 2026-10-24; approved 2026-10-25) | When the plan is approved | after round 7's targeted audit, with rounds 6–7's local gaps carried as required tests (§V T1–T14) | Another full audit round |
| R-1d (gblock, 2026-10-25) | Interactive charts | out of this plan: #1574, not blocking v2; its residue rulings (script policy, network) live there | Interactive charts in Part 1c |
| SVG (gblock, 2026-10-25) | SVG under `--show image` | allowed, shown statically; `compute --help` teaches `svg.fonttype: none`; the computation lens cannot see a rendered SVG — stated residue until #1573 gives the lens a renderer (T14) | PNG only |
| report.md raw HTML (lead, disclosed) | How report.md keeps seat raw HTML out, where #1569's renderer does not run | written through #1569's decision on the same whole-document parse: every `<` outside code, autolinks and the tool's shapes escaped — one pass at the write replaces round 6's `<` escape at five call sites, its passed-in-element exception and the `edit`/`ingest` raw-HTML refusal | Leave report.md's raw HTML to the viewer (an unclosed `<!--` hides later sections in any CommonMark viewer; `<img src=https://…>` loads in viewers that render HTML); keep round 6's per-site escape (defeated, `probe2-out.txt`) |
| R-fs | Filesystem sandbox | not now; residue stated (R7); #1560 | Sandbox in 1a |
| R-commit | Input bytes in `run-archive/` | none; the archive commits the citation; `archive rerun` re-fetches and verifies | Commit cited bytes |
| N1 | Home of the archive re-run | group `archive`, member `rerun <runDir>` on an extracted, migrated archive; `capture` stays top-level and writes through package `archive` | ~~`capture` moved into `archive`~~; ~~top-level `rerun`~~ |
| N2 | `archive rerun --citations` | ship in 1b, opt-in | Inputs only |
| P-norm | Absolute paths in output | each copy's canonical root → `<bundle>` before comparing and recording, at deposit and every re-run | Record raw; such programs stay `VARIANT` |
| R-rights | Computing over cited bodies | stated and accepted, as for citations; every C4 carrier, the seat note included, states it; §IV R8 grades every channel | Consult `TDMReserved` |
| R-input | What `--input` carries | raw bytes only | Extracted text |
| R-size | Limit on output and `out/` | unbounded, stated | A cap |
| Re-run (2026-10-24) | How many re-runs a bundle gets | ONE standing re-run per bundle; each act's verdict is its own comparison against it; disagreeing bases surfaced (`bundle_bases`) | A re-run per act; one verdict per bundle |
| Re-run store (lead, disclosed 2026-10-24) | Who writes the re-run's raw result, when, and which stands | executed ONCE into staging; renamed to `rerun/<rerun_sha>/` (content-addressed, so migrate renames nothing); then the `ran` event is appended — no event lacks its directory; the LATEST `ran` reproduce carrying `rerun_sha` stands; leftovers ignored and never archived; capture and `archive rerun` execute fresh and touch no store | Name by `events.id` (every migrate renumbering must rename); the earliest stands; sweep leftovers (races a writer's staging); refuse a concurrent second execution |
| Crop (gblock, 2026-10-24) | How a seat names a crop | `cite --show image` in two phases: a colour view the model processes unresized (View sizing), size read from its PNG; a box in that view's pixels; the crop from the full-resolution raster; phase 2 records first, a wrong crop is retired. No new blue verb; `render-page` stays the lens's | ~~`render-page` offered to blue~~; ~~a box on the file raster~~ (the model's coordinates are in the image it processed) |
| Caption (lead, disclosed) | A shown block's caption | the act's footnote text: `--reason` for a computation, title and page for a citation | A citation's `--reason` (run-voice, `cite.go:191–195`) |
| Literal (lead, disclosed) | Seat, program and source text inside a shown block | `report.Literal`: every ASCII punctuation escaped, line breaks folded | `EscapeComposed` alone (leaves `[`, `\|`, line breaks) |
| Colour (2026-10-24) | Colour of shown crops | colour, through PDFium's RGBA raster; the lens compares against the colour view; grayscale OCR renders unchanged | Grayscale crops |
| CI (2026-10-24; job and cache lead, disclosed) | Where archived programs run | a job with no write credential and no saved cache: its own scheduled `archive-rerun` job; `--help` states operator privileges; §IV R12 | The release job; a step of `feov-record` (its schedule saves the Go caches pull requests restore) |
| Per-act (lead, disclosed) | Store key for an act's artifacts | `dir/` by `bundle_sha`; output and `out/` under `acts/<K-id>/`, written once, before the event | One store per sha |
| Crop-absent (lead, disclosed) | Assembly when a shown crop is not cached | a loud placeholder in subject terms; the operator pointer to `archive rerun` goes to run.md, never report.md | Omit the figure; ~~an operator command in report.md~~ (report-voice ruling) |
| Identity (lead, disclosed) | Mode and framing in the listing | no mode (the tool sets argv[0] executable); length-prefixed fields; `__pycache__`/`.pyc` refused by name; `meta.json` not migrated | A mode field; refuse newlines; exclude caches silently |
| Composed input (lead, disclosed) | `--input` of bytes the tool composed (`recovered`/`named_backend`, no `copy_url`) | refused: no document exists to re-fetch | Accept as residue |
| Images (lead, disclosed; narrowed by R-html-split) | What may open an image | assembly only: composed text escapes `![` as it escapes `[^`; authored text refuses it outside code (measured 0); #1569's image arm embeds only a path the record places — raw HTML is #1569's, no longer this plan's | Key on the destination string alone (a seat can type a placed path); a hand-kept list of image-loading tags (misses CSS `url()`, `<picture>`, …) |
| View sizing (lead, 2026-10-24, serving the Crop ruling) | How large a view is | `ResizedSize(raster, 1568, 1568)` — the guide's reference algorithm, both edge and token limits, the largest view it leaves unchanged; the record checks `fits(view)` in SQL (equivalent, §II) | Long edge ≤ 1568 alone (the model resizes a letter view 1212×1568 to 952×1232); the high-resolution tier's limits (resized again on the standard tier) |
| Crop sha (gblock, 2026-10-24) | What the record holds of the crop | phase 2 records `crop_sha256` = sha256 of the DECODED RGBA pixels (with width and height), not the PNG bytes; the crop is stored as `cache/crops/<pixel sha>.png`; `archive rerun --citations` re-derives and prints `differs` loudly, exit 1, not refused — a renderer upgrade can move pixels | Hash the PNG bytes (an encoder change moves them); re-derive silently; refuse on a moved sha |
| Code regions (lead, 2026-10-24; narrowed by R-html-split) | How an escape or refusal finds code | goldmark's tree, the renderer's own parser (`mdReader`; `lineReader` for one inline line); `eachLine` rebuilt on it, `reCodeSpan` deleted; units join closed (an open fence closed at the unit's end); serves footnotes and the image `!` only | A second hand scanner taught run lengths and indented code (a rival reading of CommonMark); escape the whole document (loses which bytes are composed) |
| Reproduce outputs (lead, disclosed) | `Reproduce.recorded_output`/`observed_output` | reserved: both outputs are in the store; evidence.md reads the act's `acts/` and its comparison's `rerun/`; migrate drops `recorded_output` (whole in `acts/`) and moves `observed_output` into `drift` | Keep both, rewritten per act on every reproduce (a third copy of bytes the store holds) |
| Comparison basis (lead, disclosed) | What a `ran` = false act records of the re-run it read | its `rerun_sha`; the write refuses one that is no longer the standing re-run's | Record nothing (a standing that moves between read and write goes unseen) |
| Phase-1 flags (lead, disclosed) | What `cite --show image` without `--region` takes | `--url` and `--page` only; `--quote`/`--title` `RequiredUnless` | Require all three in both phases (a quote and title stated twice, recorded once) |
| Placed names (lead, disclosed) | A program-chosen `out/` name in a figure's destination | `url.PathEscape` per path segment in one `placedPath`, which the allowlist's image arm matches | A tool-named copy of the file (a second copy of `out/` bytes); refuse names by a character class; escape the whole relative path (`/` → `%2F` resolves nothing) |
| Lifecycle (gblock, 2026-10-30; #1657) | How citations and computations live | exactly alike: a computation, like a citation, is a textual statement about external or reference material ("even a reference to a figure in text is a citation"). Retired anchors are dropped from evidence.md, and the record keeps them. A `source` or `computation` gap closes only on a live answer: `cite --answers`, and retiring the answer reopens the debt. Retire never auto-closes a gap. Part 1a′ is built before 1b. No compatibility path | ~~retire auto-closes the gaps its anchor answered~~ (rejected); keep two lifecycles |
| L1 (gblock, 2026-10-31) | A gap minted `--about-kind anchor` about an evidence anchor (the failed check's carrier) under `--check-kind source`/`computation`: does that anchor's retirement meet its live-answer demand? | yes. The demand is met when the gap's `about_ref` anchor is retired, read from two fields, `mint.about_ref` and `retire.anchors`; nothing is inferred. Red still closes | ~~(b) strict: only a live anchor answering it closes it; red regrades or supersedes~~ (leaves the smoke's `G-342d3cdc`, `G-78817cfa` at CEILING); ~~(c) the retire must name the gap (`--answers G`)~~ (a retire without it leaves the gap unclosable, and a retire cannot be repeated) |
| L2 (gblock, 2026-10-31) | A CLOSED gap whose answering anchor is retired | the closure falls, derived, once no anchor that answered the gap when the close was recorded is live: `open` and `awaiting_evidence` until a new answer AND a new close. An answer recorded after the close never restores it; nothing auto-re-closes. A bench closure and a closure that rested on no answer stand. One definition of `open` that every reader (dispatch, PASS, work list) follows | ~~(b) the gap stays closed and a "closed on retired evidence" blocker holds PASS until red re-closes~~ (a second state machine); ~~(c) only an open gap's debt reopens~~ (a closed gap's evidence vanishes unseen) |
| L3 (gblock, 2026-10-31) | Claim count | a `K-` on prose counts as a claim exactly as a `C-` does: one per distinct anchor per sentence. `CountsAsClaim` reads `backs`; the `claim` column is deleted. The logic lens's budget only grows, so no archived mint is refused | ~~(b) count sentences carrying at least one evidence anchor~~ (changes the citation count too); ~~(c) keep `K-` `claim: false`~~ (the divergence the Lifecycle ruling removes) |
| L4 (gblock, 2026-10-31) | One failed-check carrier | a standing `refutes`/`absent` reading of a citation is a failed check, carried by the reader's gap about the anchor, as `unsound` is. A contradiction keeps only red's corroborations. migrate translates archived contradiction mints on anchored readings, from counts measured and committed before the translation is written (census L8; Fork L4 translation; V1a″.3) | ~~(b) an unsound computation becomes a contradiction~~ (adds a finding step to the computation lens); ~~(c) two carriers, stated kind-specific~~ ("exactly like citations" fails on the path that holds the PASS) |
| L5 (gblock, 2026-10-31; reverses the lead's recommendation) | Does `reproduce` record the sentence it judged, as `verify` records `claim`? | yes: `Reproduce.claim` (field 13), read by the tool from `record.LiveSentence`, never typed. Its readers are the K- counterparts of `Verify.claim`'s (census L9): evidence.md's per-check sentence line (`checkLines`), the judgement JSON's `claim`, `reproduce`'s own result, the ledger's `anchorCheck.sentence` and the oracle's staleness rule; in 1c, a shown block's referring sentence. Archived reproduces carry none (residue, III.1′.6) | ~~(a) record nothing new~~: its premise, that no surface would read the field, was wrong — `Verify.claim` is read at `citations.go:28`, `:100` (evidence.md's sentence line), `citationid.go:533` (the contradiction key) and `evidenceview.go:404`, and the first two have a computation counterpart. (The draft also cited `assemble.go:933`, which reads `Retire.claim`.) |
| L6 (gblock, 2026-10-31) | Retiring a citation a live computation names (`--cites`; 1b `--input`) | refused while such a computation is live, naming each; blue retires or recomputes it first. The tool refuses and never cascades | ~~(b) the computation reopens when its citation retires~~ (a derived cross-kind edge the ledger lacks); ~~(c) allowed~~ (evidence.md names a citation the report no longer holds) |
| Source demand (gblock, 2026-10-31) | Does `source` bind every closing disposition? | yes: a hard live-answer demand on every disposition that closes, `moot` and `not_a_defect` included, exactly as `computation` is (`close.go:69–72`). Stated in `close --help` and `mint --help`. The draft's gloss "a demand the lens no longer holds is regraded or superseded" was wrong (round 1: a regrade carries grades only, `record.proto:1069`, and a close to a successor is held to the demand): the exit is the bench, reached by `motion docket`, and the help names it | ~~exempt `moot` and `not_a_defect`~~; ~~carry it as a risk~~ |
| Staleness surface (lead, disclosed; serving L5) | How a check of a sentence the report no longer carries surfaces | the ledger's edit rule (`sinceEdit`) stays the ONE rule that decides standing, for both kinds; the recorded sentence surfaces it: evidence.md's "of the sentence" line through one `checkLines`, and one oracle rule over `recordsql.EvidenceTables` that fails a STANDING check whose recorded sentence is not the anchor's live sentence (III.1″.1; the oracle rule ruled BUILD, Fork Staleness) | the recorded sentence as a second standing rule in the ledger (the ledger is a fold over events and renders no report; and the sitting rule is the stricter, since it also catches a check written after the edit by a seat that read before it); a runtime refusal (nothing to refuse: the check was of the sentence when written) |
| Smoke fixture (gblock, 2026-10-31) | Where V1a′.7(a)'s record lives | committed to `run-archive/` by 1a′'s first commit, as a real run (Phase 0), with `outcomes.tsv` re-measured by the `archive-rerun` job | a testdata copy of `record.db` (a second home for a run's record) |
| Show evidence after retire (lead, disclosed; stands, gblock 2026-10-31) | What `show evidence` says of a retired anchor | its lists and counts hold live anchors only, as evidence.md does. `--anchor <retired>` returns it with `retired` (event, reason), so a lens can still read what left | list every anchor with a `retired` field (the lists and evidence.md would disagree) |
| Answer of another kind (lead, disclosed; stands, gblock 2026-10-31) | `cite --answers` on a `computation` gap, or `compute --answers` on a `source` gap | accepted; only the demanded kind discharges a demand (an anchor may answer part of a gap of any kind, as `compute --answers` on a `document` gap does today) | refused at write |
| Kind-row home (lead, disclosed; stands, gblock 2026-10-31; package moved in revision r1, lead, disclosed) | Where kind differences live | two tables. `anchor.kinds` holds the vocabulary (the leaf imports nothing). The record facts (table, id column, check kind, check table, noun; 1a″: owner, failure words) are `recordsql.EvidenceTables`, held level with `anchor.kinds` and the DDL by a test, with the SQL arms generated from it. It is in `recordsql`, not `record`, because the view generator in `recordsql` reads it and `record` imports `recordsql` (an import cycle the round-1 audit found) | all in `anchor.kinds` (the leaf would import `recordpb`); `record.evidenceTables` (`recordsql` cannot import `record`); a proto option on `Cite` and `Computation` read from the descriptors (a new option kind for six values); per-site `if computation` (today's 24 sites, L2) |
| Split (gblock, 2026-10-31, after plan-audit round 1) | One part or two for the lifecycle | two: 1a′ = the CEILING fix (retire for every evidence kind, the bare-marker refusal, the closure demand and live answer incl. `cite --answers`, render after retire, L1, L2, L3, L6, the `source` demand); 1a″ = L4 (the failed-check carrier and the archive translation) and L5 (`Reproduce.claim` and the staleness oracle rule). Each its own PR and epoch; 1a″ builds on 1a′ | one Part 1a′ at one epoch |
| Staleness (gblock, 2026-10-31, after plan-audit round 1) | Build the consistency oracle's staleness rule over the evidence table | BUILD, in 1a″ with L5 (III.1″.1) | leave staleness to the ledger and evidence.md's line |
| L2 history (gblock, 2026-10-31, after plan-audit round 1) | Which closure-history readers follow a fallen closure | every one is re-derived: `convergence_at` per point, `HasClosed`/`ClosedEpoch`/`Closure`, scorecard, telemetry, chart, spot-check archive, dashboard, repair denominator, board JSON, the oracle's raw walk (III.1′.3 Closure history) | re-derive only the gap view's `open` (the round-1 gap: a closure payload on an open gap, a double-counted re-close) |
| L1×L2 (gblock, 2026-10-31, after plan-audit round 1) | An L1-met closure whose answering anchor is later retired (the fixture's `G-a0c6ca1e`) | it STANDS: the `about_ref` anchor's retirement is permanent, so the demand stays met whatever happens to the answering anchors; only a closure that rested solely on answering anchors falls | it falls with its answer |
| L1-late (gblock, 2026-10-31) | A red closure that rested solely on answering anchors at the close, whose gap's `about_ref` anchor is retired AFTER the close and BEFORE the last answer is retired: does it fall at that last retire? No archive holds the case (`rev-r1/`: the fixture's only L1 closure, `G-a0c6ca1e`, had its `about_ref` retired before the close) | **It stands.** L2 fells a closure because no live answer meets the demand; at the fall point L1 meets it, and ruling L1×L2 gives the `about_ref`'s retirement as permanent. The SQL rule then reads "L1 met at or before the fall point", one arm wider than "at the close" | ~~it falls ("rested solely on answering anchors" read at the close)~~: the gap would be open with its demand already met |
| Fall in event folds (lead, disclosed; serving L2 history) | How the Go folds that decide "closed" from `Close` events (`engagementOf`, `closingDutiesOf`) learn a closure fell | their entry point hands them the standing set read from `closure_act.fell_event`, as `motionsAt`'s fold is handed in today: one statement of the rule, in SQL; the oracle's raw walk is the only second statement, deliberately | a Go re-implementation of the rule held level with the view by a parity test (two statements every reader could drift between) |
| Demand at the write (lead, disclosed; serving Source demand and the round-1 gap on `carry`) | Where the closure demand is checked | `record.validate`'s `Close` arm, in the writing transaction, for every `Close` writer (`close`, `carry`), skipped under `Migrating` as `requireFailedCheckAnswered` is | the CLI verb `close` alone (`carry` writes a `Close` on nothing) |
| L4 translation (lead, disclosed; serving L4) | What migrate translates, given the archived chain | the 5 findings `about_kind` `contradiction` that raise an anchored reading's contradiction become findings about that citation; the 2 mints `found_by` them stay as written. The ruling's "contradiction mints" do not exist (0 measured; `about.go:48` refuses one) | also move the 2 mints to `about_kind` `anchor` (they placed a gap marker the later archived acts were written against; moving the subject drops it); translate nothing and accept the 5 refusals |
