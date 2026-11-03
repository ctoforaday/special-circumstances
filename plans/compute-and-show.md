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
> was followed by rulings the same day (Forks: Split, Staleness): the amendment is two parts, 1a′ (the
> CEILING fix, epoch E1′) and 1a″ (one failed-check carrier and the judged sentence, epoch E1″), each
> its own pull request. On 2026-11-01 gblock ruled that closures are history (Fork Closures are
> history): a closure never falls, and a retire affects one-hop work lists only. L2 and the rulings
> that served it are rejected rows in Forks. The same day gblock ruled jointly with #1647 (Fork Joint
> ruling with #1647): #1647 governs a refuting reading, lands first and deletes contradictions; L4 is
> dropped; 1a″ is L5 and the staleness oracle rule only; 1a′ is built on #1647's shape and starts
> after it merges, its line references, epoch and field numbers re-derived then. This revision answers
> round 1's gaps and notes that depend on neither L2 nor L4. 1b, 1c and 2 are amended where they
> assumed two lifecycles (Parts 1a′ and 1a″ dependents). Ledgers are under `~/.claude/scratch/anchor-lifecycle/` (`census-verify.tsv`: the
> 1a-branch census re-verified on main; `audit-r1/`: the auditor's sweeps; `rev-r1/`: the 33 migrated
> archives and `l4census.txt`; `graph/`: the record's dependency census and the ruling;
> `rev-r2/`: the work-list measurements, `items.txt`; `p1647.md`: #1647's plan at `02cf1c189`) and
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
out and keeps its sentence, which an edit may then change or cut. No new lifecycle: Part 1a′ makes
the existing one a single lifecycle, which every evidence kind lives by (`anchor.Backs`).

**Parts (R-split).** Each ships alone, with its own epoch and `migrate` translation.

| Part | Scope | Depends on | Epoch |
|---|---|---|---|
| 1a | `prove` → `compute`; directory + argv; two-value basis; per-act store; one re-run per bundle, judged per act; `K-` anchors | — | E1 |
| 1a′ | One anchor lifecycle, the CEILING fix. `retire --anchor` alone and the bare-edit refusal apply to every evidence kind. A retired anchor is dropped from every document; `[^K<n>]` numbered per document. `source` and `computation` gaps close only on a live answer (`cite --answers`), under every closing disposition; a retire of the anchor a gap is about, or whose reading it carries, meets it, and so does a reading gap's own backing re-read. A retire moves no closure: it affects one-hop work lists only. A `K-` is a claim. A citation a live computation names cannot be retired. `assembly-screen` fails on a live anchor of either kind whose check fails | 1a, #1647 | E1′ |
| 1a″ | `reproduce` records the sentence it judged; evidence.md and the consistency oracle surface a check of an earlier wording | 1a′ | E1″ |
| 1b | `--input` (raw cited bytes, read-only); package `archive`; operator `archive rerun` | 1a, 1a′ | E2 |
| 1c | `--show table\|image\|text` on `compute`; images exist only as shown blocks | 1a, #1569 (render-time allowlist) | E3 |
| 2 | `--show image\|text` on `cite`: a colour crop of a box drawn on a view the model processes unresized | 1b (`archive rerun` re-derives crops), 1c (placement, `Show`, `Literal`) | E4 |

E1–E4 are the next free epochs when each part lands (main was 43). 1a landed at 49. #1678 took 54,
and #1647 holds the next; E1′ and E1″ are the next free epochs when 1a′ and 1a″ land, derived then.

**Success criteria.**

| # | Part | Criterion | §V |
|---|---|---|---|
| S1 | 1a | `feov-record blue compute --dir D --quote S --reason R -- python3 main.py` records a computation: bundle = every regular file under `D`; command = the argv after `--`; toolchain = resolved path + sha256 of `python3`. `blue prove` is an unknown command | V1a.1 |
| S2 | 1a | Two runs in fresh copies outside the run tree, each copy's canonical root replaced by `<bundle>` in the output. Equal output, exit and `out/` → `REPRODUCIBLE`; different with no `--variance` → refused naming what moved; with `--variance` → `VARIANT` + drift; `--variance` over agreeing runs → refused | V1a.2, V1a.3 |
| S3 | 1a | A recorded act never changes after write. Every act's verdict compares the bundle's STANDING re-run — the latest `ran` reproduce, its result at `rerun/<its rerun_sha>/` — against the act's OWN output, exit and `out/`, whatever refusal, crash or concurrency preceded it; capture's audit and `archive rerun` execute fresh and neither read nor write the store | V1a.4, V1a.5 |
| S4 | 1a | A keyed retry restates every recorded flag or is refused; program, argv and inputs are in the sha | V1a.6 |
| S5 | 1a | Every archive migrates with zero refusals; III.1.6's assertions hold; `archive rerun` over every migrated archive gives each of the 159 stored scripts the outcome `archiverun` measured on the CI image and committed | V1a.7 |
| S11 | 1a′ | `retire --anchor <id>` alone takes any evidence anchor out of the report and keeps its sentence. An edit leaving one bare is refused. A retired anchor appears in no document, and `verify` and `reproduce` refuse it. Every surface naming a retire names the anchor's kind. Footnotes are numbered per document over live anchors. A sentence's `K-` counts one claim | V1a′.1, V1a′.5, V1a′.8 |
| S12 | 1a′ | `close` on a `source` or `computation` gap is refused unless a live anchor of the demanded kind answers it (`cite --answers`, `compute --answers`); or the anchor it is about, or the anchor of a reading it carries (#1647's `carried_by`), is retired (L1, widened); or, on a `source` gap carrying a reading, that reader's latest reading of the citation backs it — whatever the disposition, `moot`, `not_a_defect` and a close to a successor included. Another lens's backing reading meets nothing, and a met demand leaves a standing failed check to refuse the close. Retiring the only answer of an open gap leaves it `awaiting_evidence`. The demand is asked in the `close` verb only: every archived close replays unrefused, and `carry` restates a closure the record holds and closes no open gap | V1a′.2, V1a′.6, V1a′.10 |
| S13 | 1a′ | A computation's keyed retry returns its id. `--cites` refuses a retired citation, and retiring a citation a live computation names is refused, both inside the write's transaction. Every reader of an anchor's answers reads both kinds | V1a′.4 |
| S14 | 1a′ | Every archive migrates at E1′ with zero refusals, and the flip census is committed. On the 2026-10-30 record (committed to `run-archive/`), in a scripted replay and in a smoke, no gap whose anchor was retired is left with no seat able to act on it: its close is admitted, or it reaches the bench. `assembly-screen` fails on a live unsound `K-` | V1a′.6, V1a′.7, V1a′.9, V1a′.11 |
| S15 | 1a′ | A retire moves no closure: a closed gap stays closed for every reader, whatever happens to its answers. An open gap whose demand a retire met owes blue nothing, is on its minting lens's list, gives that lens its turn, and closes (L1). A closed gap whose answers are all retired carries `answers_retired_event` on the board and an item on its minting lens's work list, which blocks nothing and does not hold PASS, and leaves that list once the lens has sat through a sitting opened after the retire, has minted a gap superseding it, or a later closing act has been recorded on it | V1a′.3, V1a′.7 |
| S16 | 1a″ | `reproduce` records the sentence it judged (`Reproduce.claim`), read off the report by the tool as `verify` records `Verify.claim`. evidence.md shows, for every check of either kind, the sentence it was of when that is not the anchor's sentence now. A standing check of a sentence the report no longer carries, in any state #1647's ledger gives a standing check, is a consistency violation the release fuzz reports | V1a″.1, V1a″.4 |
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

**Built on #1647.** #1647 (`plans/contradiction-is-a-gap.md`, branch `plan/contradiction-is-a-gap` at
`02cf1c189`) governs a reading that finds against a citation, lands first, and deletes the
contradiction everywhere, corroborations included (Fork Joint ruling with #1647). Its shape is this
part's baseline: a negative reading of the work mints its gap in the reading act, placed by quote at
the sentence, with `about_kind` and `about_ref` unset; `Verify.carried_by` names the gap that carries
the reading and `Verify.carrier_minted` whether the act minted it; `found_against` is a third
failed-check state beside `unread` and `unsound`, and `requireFailedCheckAnswered` refuses the close
of its carrying gap while it stands; a closing bench ruling closes that gap and the check still holds
the PASS. An `unreachable` or not-the-work reading mints nothing and keeps the `unread` check's owed
gap about the anchor (#1647 §III.0, §III.3). 1a′ starts after #1647 merges. Its file:line references
are read at main `a31a7b06`, and those in files #1647 rewrites (`sitting.go`, `dispatch.go`,
`failedcheck.go`, `passblockers.go`, `retirement.go`, `mintbudget.go`, `anchorchecks.go`,
`cli/lens/verify.go`, `record.go`, `recordsql/views.go`, `migrate/replay.go`) are re-verified on
#1647's merged head before work starts. The epoch E1′ and every new field number below are derived
from main at that point; none is carried from this document.

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
part. Part 1a′ removes them before 1b, so that §I's placement rule, and every shown block in 1c and
2, rests on one lifecycle; #1647 removes the contradiction rows, and 1a″ adds the judged sentence.

**The rulings (gblock, 2026-10-30, binding).** Citations and computations are the same thing: "a
textual statement about an external source or reference material. Even a reference to a figure in
text is a citation." They "behave exactly like citations". A retired anchor of either kind is
dropped from evidence.md, and the record keeps it. A gap whose check kind is `source` or
`computation` closes only on a LIVE answer: `cite` gains `--answers G-…`, a non-retired anchor of
the demanded kind must answer the gap, and retiring that anchor leaves an open gap's debt owed. Retire never
auto-closes a gap. The tool refuses; it never infers.

**The rulings (gblock, 2026-10-31, binding; Forks L1–L6, Source demand).**

- **L1**, widened by the joint ruling with #1647 (2026-11-01). A gap under check kind `source` or
  `computation` has its live-answer demand met when a retired anchor is its `about_ref` (a gap minted
  `--about-kind anchor`, the `unread` and `unsound` checks' carrier), or is the `anchor` of any Verify
  whose `carried_by` is this gap (#1647's reading gap). Read from `mint.about_ref`,
  `Verify.carried_by`, `Verify.anchor` and `retire.anchors`; nothing is inferred. Red still closes it.
- **L3.** A `K-` counts as a claim exactly as a `C-` does: one per distinct anchor per sentence.
  `CountsAsClaim` reads `backs`, and the `claim` column of `anchor.kinds` is deleted.
- **L5** (Part 1a″). `reproduce` records the sentence it judged (`Reproduce.claim`), exactly as
  `verify` records `Verify.claim`, and the K- surfaces read it as their citation counterparts do.
- **L6.** Retiring a citation that a live computation names (`--cites`; 1b's `--input`) is
  refused, and the refusal names those computations.
- **Source demand.** `source` is a hard live-answer demand on every closing disposition, `moot`
  and `not_a_defect` included, exactly as `computation` is today (`close.go:69–72` exempts no
  disposition).

**The rulings after plan-audit round 1 (gblock, 2026-10-31, binding; Forks Split, Staleness).**

- **Split.** Part 1a′ is the CEILING fix: retire for every evidence kind, the bare-marker refusal,
  the closure demand and live answer (`cite --answers` included), render after retire, L1, L3,
  L6 and the `source` demand. Each part is its own pull request at its own epoch; 1a″ builds on 1a′.
  The joint ruling with #1647 narrows 1a″ to L5 (`Reproduce.claim`, `checkLines`, the judgement
  JSON's `claim`) and the staleness oracle rule.
- **Staleness.** The consistency oracle gains the staleness rule over the evidence table (1a″, with
  L5).

**The joint ruling with #1647 (gblock, 2026-11-01, relayed by the Relocations session; binding).**
#1647 governs a refuting reading and lands first; L4 is dropped (Fork L4, rejected), and
contradictions are deleted everywhere, corroborations included. 1a′ lands second, on #1647's shape. Its
seam with #1647 is the widened L1 above and four rules (III.1′.3 The demand on a reading's gap): the
backing re-read arm of `evidence_answered`; the source reading-gap table; an unread citation's gap
staying on the about-anchor path; and the demand staying in the `close` verb. Part 2's refuted crop is
#1647's in-act gap, and `assembly-screen` reads one predicate, "this check fails", through the kind
table rather than a shared state (III.1′.1 Capture). #1647's own tests that close a reading's `source` gap after an edit, before any demand exists (its §III.15, plan head `21c4cb901`), are rewritten by 1a′ when it adds the demand: each such close then needs the reader's backing re-read or `cite --answers` first (the source reading-gap table), and is otherwise refused.

**Closures are history (gblock, 2026-11-01, binding; Fork Closures are history).** Closures never
fall. A retire affects only one-hop work lists; there is no general dependency graph, no walker and
no edge table.

| State | What holds |
|---|---|
| Open gap about a retired anchor (L1), or carried by a retired anchor's reading (L1 widened) | The demand is met. Inactionable for blue (not `awaiting_evidence`), on red's list, and red closes it. No dead end |
| Closed gap whose answers are ALL retired | The closure stands. The owning lens's work list shows an item: its answer was retired; mint if still a defect. It does not hold PASS. No new act |
| Open gap with a demand | A live answer is required; a retired answer does not count |
| Citation named by a live computation (L6) | Retire is refused, naming them |

| Part | Holds | Epoch | PR |
|---|---|---|---|
| 1a′ | Phase 0; the kind rows; `retire --anchor` for every kind; the bare-marker refusal; `cite --answers`; the closure demand at `close`, every disposition; L1 as widened; the demand on #1647's reading gaps (III.1′.3); the two work-list behaviours of a retire (III.1′.3 The work lists); render after retire (every document, per-document `[^K<n>]`); L3; L6 inside the write transaction; `verify`/`reproduce` refuse an anchor with no live sentence; the work list's backing over both check verbs; `assembly-screen` through the kind row's failed state; `computation-rerun` over live computations | E1′ | its own, after #1647 merges |
| 1a″ | L5: `Reproduce.claim`, `checkLines`, the judgement JSON's `claim`; the staleness oracle rule | E1″ | its own, after 1a′ merges |

**1a′ is one complete concept and ships alone.** Its concept is "an evidence anchor of either kind
lives one way, and a demand is discharged only by a live answer of its kind": every carrier of that
concept is in this part (the verbs, the record and its views, both work lists, render, agent text,
migrate, the fuzz drive), and nothing in it waits on 1a″. 1a″ adds a recorded field and a surface;
without it the ledger's edit rule still decides standing for both kinds, as today.

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
committed archive takes it up — census L12, two sweeps: the files that glob
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
  - `noun`: "citation", "computation" — the word every refusal and result names the kind by;
  - `failed`: the failed-check state word in which a check of this kind finds against its sentence
    (`found_against` for a citation, `unsound` for a computation: #1647's `Blocker.CheckState()`
    words), which `record.FailingAnchors` reads (III.1′.1 Capture). The two states stay two; the
    predicate "this check fails" is the one thing shared.

  `TestEveryBackingKindHasAnEvidenceRow` (record) fails when a `backs` kind has no row, when a row
  names no `backs` kind, when a `check` value has no row, when a row's `failed` is not a state word
  the ledger emits, or when a row's `table`/`idColumn`/`checkTable` names no table or column of the
  generated DDL.

A new kind adds one row to each table. It adds no lifecycle code: the `evidence_anchor` view's UNION
arms are generated from `EvidenceTables` (III.1′.3), and every hand-written UNION over the evidence
tables reads that view instead (census L13: `record/id.go:86–87`, `record/refs.go:130`, `:204`,
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
| Closure demand, at the close | `close` is refused on a `source` or `computation` gap that is not `evidence_answered` (live answer, L1 as widened, or a reading gap's backing re-read: III.1′.3), whatever the disposition (ruling Source demand), `moot`, `not_a_defect` and a close `--superseded-by` included. The check stays where it is today, in the verb: `requireComputationAnswered` → `requireEvidenceAnswered` (`cli/lens/close.go:69–73`, `:114–131`). Replay never reaches it: `internal/record/migrate` links no `internal/cli` (`go list -deps`: 236 packages, `internal/record` among them, no `internal/cli`), so an archived close, the 9 on `source` gaps with no answer included, migrates unrefused (V1a′.6). `carry` needs no demand check: it writes a `Close` only where a `close` row of the gap already exists (`priorClosureEpochs`, `record.go:1147–1162`; its flags hold no verification triple, so that arm always runs), a `close` row closes the gap for good, and so a carry can land on no open gap. The bench's closing ruling (a `motion_rule` docket disposition that closes) is not a `Close`, is not held to the demand, and is the exit for a demand the lens no longer holds: the lens puts the gap before the bench (`motion docket`), as at impasse | `computation` only, in the same verb (`close.go:69–73`), and any computation naming the gap satisfies it, retired or not | `check` |
| Retire the answer | Retiring an answering anchor removes it from the answer set, so an OPEN gap with a demand is `awaiting_evidence` again until a new answer. A CLOSED gap stays closed: a closure is history, and no reader's count of it changes. A closed gap whose answers are all retired carries `answers_retired_event` and an item on its minting lens's work list (III.1′.3 The work lists). Retire closes nothing and reopens nothing | a retired `K-` still answers (`estoppel.go:280`, `views.go:569–572`), and nothing tells the lens | `check` |
| Gap about the anchor, or carrying a reading of it | A gap minted `--about-kind anchor` (the `unread` and `unsound` checks' carrier), and a reading's gap that a Verify names in `carried_by` (#1647's `found_against` carrier), are the two carriers of a failed check. Under check kind `source` or `computation`, the demand is met when the `about_ref` anchor, or the `anchor` of any carrying Verify, is retired; nothing is inferred, and red still closes it (ruling L1, widened). Blue owes it nothing; its lens's list says so (III.1′.3 The work lists). An unread citation's gap is not a reading gap: an `unreachable` or not-the-work reading carries no `carried_by` (#1647 §III.2), so its owed gap is minted about the anchor and stays on the `about_ref` path | refused at `close` (`requireComputationAnswered`) even when retiring the anchor was the fix (the smoke's G-342d3cdc, G-78817cfa) | — |
| Work-list backing | `backing[]` is read from the ledger: the latest standing check of each anchor that answers the gap, with outcome or soundness, its seat and its event (`backingOf`, `viewjson.go:1181`), through the row's `checkTable` | only `Verify` is indexed, so a `K-` shows an empty outcome | `checkTable` |
| Cross-kind reference | `compute --cites C-…` (and 1b's `--input`) refuse a retired citation (`requireLiveCitation`, replacing `requireCitation` at `record.go:1051`). `retire --anchor C-…` is refused while a live computation names it in `cites` (1b: or `inputs`), and the refusal names those computations; blue retires or recomputes them first. Both refusals are asked of the writing transaction in `record.validate` — the `Computation` arm and the `Retire` arm (`record.go:1451–1464`) — so a `compute --cites C` and a `retire --anchor C` racing each other cannot both land. The tool refuses and never cascades (ruling L6) | `requireCitation` accepts a retired `C-` (`refs.go:239`). evidence.md prints "applies the method cited at C-…" for a citation the report no longer holds | — |
| Capture | `computation-rerun` samples live computations only (`computationrerun.go:93`). `assembly-screen` FAILs on any live evidence anchor whose check fails: `record.FailingAnchors(evs, win)` [NEW] lists each live anchor whose ledger state is its row's `failed` word, and replaces `CitationsFoundAgainst` in `capture.go:470`; its SKIP reads "no evidence anchors on the record". `counts.sources_refuted` keeps reading `CitationsFoundAgainst` (a citation count) | retired computations sampled. `assembly-screen` screens only citations found against (`capture.go:470`): a live unsound `K-` passes it | `failed` |

**What stays kind-specific, and why**

| Difference | Why it is intrinsic | Where it lives |
|---|---|---|
| How evidence is made: fetch, sha256 of bytes, span, OCR versus bundle, two runs, basis | It is the evidence itself | `cite.go:92–110`, `compute.go:120–165` |
| How it is checked: a reading against the source versus a re-run and a soundness judgement. The `unread` state is citation-only | A source can be unreachable or not the work. A stored bundle always runs (`RERUN_CALLS_THE_TOOL` is refused at write) | `anchorchecks.go:160–176` |
| Owning lens | Different competence | `anchorOwnerOf` (`available.go:431`) |
| The failed state's name: `found_against` versus `unsound` | A reading finds a source against the sentence; a judgement finds a program unsound. #1647 words each state's remedy | `EvidenceTables.failed` names it; the ledger keeps both states |
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
| `lens close` | `requireComputationAnswered` → `requireEvidenceAnswered` in place (`close.go:69–73`, `:114–131`), reading `gap.evidence_answered` through `record.EvidenceAnswers`; a record it cannot read refuses the closure, as today. Its refusal names `cite --answers` for `source` and `compute --answers` for `computation`; for a gap about an anchor, `retire --anchor <about_ref>` as well; for a reading's gap, `retire --anchor` of the read citation, and the reader's own backing re-read (III.1′.3); and for a demand the lens no longer holds, `motion docket --id G` — the bench's ruling, not a close, disposes of it. The verb asks the demand before the record's `requireFailedCheckAnswered`, so a close the demand admits can still be refused by a failed check that stands, in that check's words |
| `chair carry` | Unchanged. It restates a closure the record holds: `priorClosureEpochs` (`replay.go:342`) refuses a gap with no `close` row, and a gap with one is closed for good. A carry after the answers are retired restates history and removes nothing from the lens's list (III.1′.3 The work lists) |
| `cite`/`compute` refusal wording | `citeRefusal` and `computeRefusal` (`cite.go:224`, `compute.go:273`) become one `placementRefusal` (MisQuote, InFence), worded from the row's `noun` |

#### III.1′.3 The record [MODIFY] (contract: event schema)

```proto
// Cite [MODIFY]: + optional string answers = <next free Cite field; 26 at a31a7b06>;
//   the gap this citation answers; flag "answers". remap class: id (respelled). Part 2's fields
//   follow it (III.0).
// Retire [MODIFY] comments only: claim_kept (:1679) and answers (:1688) "a retire of a CITATION" / "a citation retire" →
//   "of an evidence anchor"; anchors (:1673) "Each named c- anchor is one counted claim" →
//   "Each named evidence anchor".
```

| View (`recordsql/views.go`) | Change |
|---|---|
| `evidence_anchor` [NEW] | `(anchor_id, kind, check_kind, answers, placed_event, retired_event)`. One UNION arm per `EvidenceTables` row, generated from the row (`table`, `idColumn`, `check`). `retired_event` = the first `retire_anchors` row naming the anchor (a retired anchor cannot come back: an edit invents no anchor, `bluedoc`; no archive holds a retired anchor later reopened or retired twice, `~/.claude/scratch/anchor-lifecycle/audit-r1/reopen-probe.txt`) |
| `gap` (`:479–636`) | `computation_answered` → `evidence_answered`, four arms, each off rows (III.1′.3 The demand on a reading's gap): (1) a live `evidence_anchor` of the demanded `check_kind` answers the gap; (2) the gap is `about_kind` `anchor` and its `about_ref` has a `retired_event`; (3) a `verify` row whose `carried_by` is the gap names an `anchor` with a `retired_event` (arms 2 and 3: ruling L1, widened); (4) the backing re-read arm. `awaiting_computation` → `awaiting_evidence`: `open`, `check_kind` is a row's `check`, and not `evidence_answered`. `answers_retired_event` [NEW]: on a closed gap with at least one `evidence_anchor` row answering it, and every such row retired, the greatest `retired_event`; NULL otherwise, and NULL when a closing act other than a carry (a `close` with no `carried_from`, or the bench arm of `closure_act`) follows that retire, since that closer had the retirement before it. The closure columns (`close_event`, `bench_close_event`, `open`, `stranded`) and `closure_act` itself are unchanged |
| `convergence_at` (`:660–717`) | unchanged: a closing act closes a gap from its point on |

`ComputationAnswers` → `EvidenceAnswers` (`estoppel.go:280`), reading `gap.evidence_answered`.
`Gap.NeedsComputation` → `Gap.AnswerKind` (`replay.go:199`), which returns the demanded anchor kind,
"" for `document`. The JSON fields `awaiting_computation` → `awaiting_evidence` (`viewjson.go:153`,
`:839`, `:974`), plus `answer_kind` beside it, so the seat reads which verb it owes. Blue's sitting
debt (`sitting.go:257–285`) reads `AwaitingEvidence`, and its words come from the row. No
compatibility alias: the old names leave every surface (CLAUDE.md).

**The demand on a reading's gap (joint ruling with #1647).** A reading's gap is an ordinary mint
with its own `check_kind` (#1647 §III.2: the lens's `--check-kind` on the reading act). Under
`document` no demand is asked of it. Under `source` or `computation` the demand is the one above, and
arms 3 and 4 are how it is met by the acts #1647 already clears the check by. Arm 4 reads rows, not the
ledger:

- **Arm 4, the backing re-read** (gap's `check_kind` = `source`, the `check` of the citation row): some
  Verify V with `carried_by` = the gap names an `anchor` A with no `retired_event`, and the latest
  `verify` of A recorded by V's own seat (`events.seat_id`) is a backing outcome. "Backing" is the
  ledger's `holds` written in SQL over the row: the reading reads the work (`outcome` is not
  `unreachable` and `source_completeness` is not `not_the_work`, `VerifyReads`) and finds nothing
  against (`outcome` is not `refutes` or `absent`, #1647's `FindsAgainst`), so `weak` backs. "By V's
  own seat" is the ledger's same-seat rule (`anchorStanding.stands`: only a seat's own later reading
  replaces its check), read as the seat of the carrying reading; another lens's backing reading meets
  nothing (Fork That lens). The arm is not asked of a `computation` gap: a reading is a check of a
  source, and only the demanded kind discharges a demand (Fork Answer of another kind). The SQL and the
  Go predicates are held level by `TestTheBackingArmIsTheLedgersHolds`, every `SourceOutcome` ×
  `SourceCompleteness` pair through both.
- **Several readings, one gap.** A gap may carry readings of several citations on its sentence (the
  one-gap rule names it on each later reading, #1647 F3). Arms 3 and 4 are met by ANY carrying
  reading. That is safe because `requireFailedCheckAnswered` still refuses the close while another
  reading's check stands: the demand and the check are two gates, asked in that order.
- **A corroboration's reading** names no anchor (#1647: a negative corroboration places none), so
  neither arm reads it. Its `source` gap's demand is met by `cite --answers G` or closed by the bench.

The source reading-gap table (`check_kind` `source`, carried by lens L's `refutes` of citation C):

| Act after the reading | The check (#1647) | The demand (1a′) | Close by L |
|---|---|---|---|
| Blue retires C | cleared (retired) | met, arm 3 | admitted |
| Blue edits the sentence | cleared (unchecked) | not met: L's latest reading of C is the `refutes` | refused by the demand, naming the re-read and `cite --answers` |
| …then L's own backing re-read of C | cleared | met, arm 4 | admitted |
| …or blue's `cite --answers G` on the edited sentence | cleared; C and the new citation are `unchecked`, each holding PASS until checked | met, arm 1 | admitted |
| L's own backing re-read, no edit | cleared (L's own check replaced) | met, arm 4 | admitted |
| Another lens's backing reading of C | stands (a disagreement, #1647 F4) | not met | refused by the demand |
| A closing bench ruling on G | stands, holds PASS (#1647 F8) | not asked: the bench's `closure_act` arm writes no `Close` and passes no verb | — (closed) |

Arm 4 reads the latest reading, not the wording it was of: `refutes`, backing, then an edit leaves
the demand met while the edit leaves C `unchecked`, which holds the PASS until a check of the new
sentence (the ledger's rule). The lens's close is its judgement of the report as it stands.

**The demand stays in the verb** (Fork Demand location). `close` asks it (`cli/lens/close.go`);
`record.validate` does not, so neither replay nor `carry` meets it:

- **Replay.** `internal/record/migrate` writes through `record`'s `Batch.Append` (and #1647's
  `Batch.AppendBodies`) and links no `internal/cli` package (`go list -deps`, `rev-r2/migdeps.txt`), so
  no archived close passes the verb. The 9 archived red closes on `source` gaps, none answered
  (`Cite.answers` is new), migrate unrefused. V1a′.6 proves it on the archives: it asserts the 9 closes
  per archive (so the arm is not vacuous) and their admission, and moving the check into the
  `Close` arm of `record.validate` fails the test at the first of them.
- **`carry`.** `cli/chair/carry.go` writes a `Close` with `carried_from` and no verification triple,
  so `record.validate` always runs its carry arm: `priorClosureEpochs` (`replay.go:342`) refuses a gap
  with no `close` row, and a carried epoch that names none of them (`record.go:1147–1162`). A `close`
  row closes a gap for good: the gap view's `open` is "no closing act", and no act reopens a gap
  (`closure_act` has a `close` arm and a closing-bench arm only). So a carry lands only on a gap that
  is already closed, and no demand is owed there. V1a′.2's carry arm shows it through the CLI, and
  deleting the `len(prior) == 0` refusal fails it.

**The work lists (ruling Closures are history).** A retire changes no closure column and no reader of
one (census L8): what it adds is one hop from the record to a work list, each read off fields the
record holds. "Owning lens" is the gap's `minted_by`, the one seat `requireOriginator` admits to
close it. A retire moves no report head (`reportHeadOf` reads edits and ingests), so a retire alone
readies nobody by Source 1.

*(a) An open gap whose demand a retire met.* The gap is `open`; its `about_ref`, or the `anchor` of a
Verify whose `carried_by` is the gap, has a `retired_event`. Under `source` or `computation`,
`evidence_answered` is true and `awaiting_evidence` false, however the retire named the gap
(`--answers` is optional, and a claim retire takes the anchor too). Call such a retire a **reaching
retire** of the gap: `retire.anchors` holds the gap's `about_ref` or a carrying Verify's `anchor`.
Under `document`, the same retire is the same act on the gap; only the demand columns do not apply.

| Seat | Reads | Sees |
|---|---|---|
| `blue-respond`, every other blue seat | `gap.awaiting_evidence` (`sitting.go:257–285`, renamed) | No item, blocking or not. The gap keeps its row in `open[]`, with `awaiting_evidence: false` and `yours_to_close: false`. A reading gap's `found_against` item (#1647's `blueFailedItem`) is gone with the check, since a retired anchor has no standing check |
| The owning lens | `gap.about_ref`; `verify.carried_by`, `verify.anchor`; `retire.anchors`; the gap's `minted_by` | `open[]` row with `yours_to_close: true`, `awaiting_evidence: false`. A new first arm of `blueAnswers` (`blueanswered.go:49`), ahead of "NO edit answers it", which would be false of this gap: a non-blocking item, "blue retired K-… (or C-…), which gap G is about (or whose reading G carries), at event N, so its demand is met — judge the report against the acceptance check you set at mint, then close it, or keep its grade with a regrade that says what still fails". A new row of `gapNewsOf` (`retirement.go:166`), `at` = the reaching retire's event, read off the stream (`Mint.about_kind`, `about_ref`; `Verify.carried_by`, `anchor`; `Retire.anchors`): so the lens owes it until it sits through a sitting opened after the retire (`lensOwesOf`, a retired lens included), and an act `AppendFresh` guards is told of it (`staleOf`) |
| Any other lens | `open[]` | the row, `yours_to_close: false` |
| Chair | `passBlockersOf`, as for every open gap | A material gap is `BlockerMaterialGap`, owner the lens (`another_seat_clears`), and holds PASS until closed (every reading gap is material by its class, #1647 §II). One that is not material is listed as holding nothing |
| Board reader | `show board --json` | `open: true`, `evidence_answered: true`, `awaiting_evidence: false`, and `about_ref` (an about-anchor gap); a reading gap's carrying reading is in `show evidence --anchor` |

**The lens's turn.** Since #1659 the plan engages a gap's minting lens only when its turn has come:
blue has recorded an act on the gap after the lens's last act there (`lensTurnOn`, `gapActsOf`;
`dispatch.go:331–343`). A retire is an act on a gap only through `Retire.answers`
(`retireAnswers`, `blueanswered.go:171`), so a retire with no `--answers` would meet the demand and
leave the lens's turn never coming: the plan would ready blue alone on the gap, and the gap would
reach the bench through blue's null turns with the one seat able to close it never readied for it.
So `gapActsOf` gains one arm: a reaching retire is blue's act on each gap it reaches, read off the
same fields as the `gapNewsOf` row (Fork A reaching retire is an act on the gap). Its readers follow
from the one function: `lensTurnOn` (the plan's `debate` and capture's `EarlyLensDispatches`), the
exchange fold's passed turn, a remand's exchange beginning (`impasse.go:160`), and `blueActed`. The
exchange fold's movement (`impasse.go`'s `Retire` case, today `retireAnswers` alone) counts a reaching
retire as movement too, as a named retire is. `ManifestOwed` keeps reading `Retire.answers` only: a
manifest row is owed for an answer blue named.

Its exits are the ordinary ones. The lens is readied (its turn has come) and its `close`, under any
disposition, is admitted (L1 as widened; `requireFailedCheckAnswered` passes, since a retired anchor's
check no longer stands). A lens that keeps its grade runs the exchange clock, and at impasse the gap
goes to the bench (`routeOf`). One that is not material readies nobody, and its lens still holds the
item and owes the news. Blue is still engaged on the gap by `debate` (Fork Blue on a met gap). No
state waits on an act no seat can make.

*(b) A closed gap whose answers are all retired.* `gap.answers_retired_event` is set (III.1′.3 view
table). "Answers" is every `evidence_anchor` row whose `answers` names the gap, of any kind, as the
ruling's words read (Fork Which answers). The gap is `open: false`, its `closed_epoch`, closure
payload, fate and every count are untouched, and PASS is untouched: no `BlockerKind`, no `lensOwesOf`
row and no `gapNewsOf` row reads the column, since `gapNewsOf` holds open gaps and an owed lens that is
behind the head is a ready lens, which is a `BlockerLensReady` and refuses `PassPermitted`.

| Seat | Reads | Sees |
|---|---|---|
| The owning lens | `WorkGapState.AnswersRetiredEvent`, `MintedBy`, `benchClosedLast` (the WORK query, `viewjson.go:1090`, widened); `satThroughOf(…).through[seat]`; `Mint.supersedes` | One `Item{Blocks: false}` from `answersRetiredItems`, called from `sittingOf`'s lens arm beside `closingsNotOwedItemsOf`: "the answers to gap G, which you closed, have all been retired, the last at event N (K-…) — the closure stands; if the defect it closed is still in the report, that is a new gap: mint it, naming G as the gap it supersedes; it takes a mint" (with `benchClosedLast`, "which the bench closed — what its ruling settled stays barred to you"). `complete` is computed from the blocking items, so it stays as it was |
| Every other seat's list | — | nothing: only the gap's minter may act on its lineage's news, and the chair mints nothing |
| Board reader | `GapJSON.answers_retired_event` (the BOARD query, `viewjson.go:436`, widened) | A closed row with its closure as ever, beside `answers_retired_event: N`, null on every other gap, never omitted. The board states the record's fact for as long as it holds; the item states the lens's news |

**What removes the item.** The existing reading of "what is new since it last sat" (`lensOwesOf`'s
`owe`), asked of the list instead of the debt, and two facts of the record. It needs no new event and
no seen-marker. The item is listed unless one of these holds:

| Removed by | Read off | Fixture |
|---|---|---|
| The lens has sat through a sitting opened after the retire: `firstAfter(ends.through[seat], N)`. It is listed for the whole of that sitting, so the sitting that clears it is one that read it. A sitting stopped at the call limit, or never closed, does not clear it (`satThrough`) | `sitting_open`, `sitting_close`, `sitting_limit` through `satThroughOf` | `G-221322e1`: retire 722, listed through the sitting opened 742, gone at its close 767 |
| The lens has minted a gap naming G in `supersedes` after the retire. A mint not naming G leaves it, and the sitting's end removes it | `Mint.supersedes` | `G-a0c6ca1e`'s successor `G-342d3cdc` was minted at 702, before the retire 779, so it does not remove the item |
| A closing act other than a carry follows the retire: the column is then NULL (III.1′.3 view table) | `closure_act`, `close.carried_from` | `G-221322e1`: the bench at 792 |

A lens that never sits again never reads it (Fork A lens that does not sit again); the board
still shows the column. Measured over the 33 archives at epoch 50 (`rev-r2/items.txt`): of 217 red
closes, 2 closed gaps ever have all their answers retired (`G-221322e1`, `G-a0c6ca1e`), and both of
their lenses sat again after the retire.

**The mint budget.** The item spends and frees nothing. A closed gap is outside the count of gaps a
lens holds open (`requireMintWithinBudget`), before the retire and after it, so a retire cannot push
a lens over its budget. A mint in answer takes a mint like any other, superseding or not: it answers
no failed check of the lens's, so `answersOwnFailedCheckAt` does not exempt it. At no mints left
(`sittingReads.mintsLeft`, `viewjson.go:1416`) the item stays listed, not blocking, with "you have no
mint left; a close of a gap you hold frees one". #1647 deletes `mintsLeft`'s one reader today (the
contradiction-at-cap arm, `sitting.go:436–437`); if its merge deletes the field too, 1a′ keeps it for
this item, as `mintBudgetOf` already computes the number for `show work`.
`requireOneGapPerClassAndSubject` does not refuse it: a closed gap blocks no mint, and a gap named in
`supersedes` is excluded besides. Superseding a closed gap is admitted (`record.go:1027`) and strands
nothing, because `stranded` needs an open ancestor. A superseding mint is not fresh
(`supersedes_count`), so it does not reset the lens's retirement count; a fresh mint does.

**Carriers.** The `gap` view's column list is hand-authored, so `answers_retired_event` and
`evidence_answered` widen each enumerating reader its comment names: the BOARD query (scan and
`GapJSON`), the WORK query (scan and `WorkGapState`), and `gapstates.go` for `AnswerKind`. `show board
--help`'s field tree (`cli/seat/verbs.go`) and `difftest`'s `manual` golden name the new field. It has
no report destination; its purpose is inter-seat (the lens's item).

**Readers of an anchor's answers (census L10).** Every reader of `Computation.answers` reads both
kinds off `evidence_anchor`:

| Reader | Change |
|---|---|
| `gapActsOf` (`gapacts.go:55–62`): a computation answering G, and a reproduce of it, are acts on G — the act that begins a remand's exchange (`impasse.go:160`, `remand.go:25`) and gives the lens its turn (`lensTurnOn`) | an anchor answering G, of either kind, and a check of that anchor (`verify` or `reproduce`), are acts on G. A `verify` is an act on G only through its citation's `answers`, never through `carried_by`: a reading that names its gap moves no turn (#1647 §III.3). A reaching retire is blue's act on each gap it reaches (III.1′.3 The lens's turn) |
| `groundsActSQL` (`grounds.go:114–120`, `:144`): a computation answering G concerns G; a citation only when an edit answering G wrote its location | an anchor of either kind concerns G when it answers G; the edit-wrote arm stays for a citation placed by an edit's text |
| `ComputationAnswers` (`estoppel.go:280–285`), `views.go:569–572` | `EvidenceAnswers` and `gap.evidence_answered` (above) |

Unchanged, and for both kinds alike: `blueAnswersOf` (`blueanswered.go:99–175`) and
`manifestowed.go:51` read `BlueEdit.answers` and a claim-kept `Retire.answers`, never an anchor's
answers (the impasse movement fold adds the reaching retire, above); `report/edits.go:36`, `view/changes.go:70`,
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

Edit the sources, then run `agentgen`, `enginegen` and `vocabdoc`. Census rows L5 and L11 (III.9).
Help pages name commands; seat prompts and constitution prose name acts (CLAUDE.md).

| Carrier | Becomes |
|---|---|
| `help/retire.md:9`, `:11` | "EVIDENCE LEAVES ON ITS OWN, ITS SENTENCE KEPT": `--anchor` alone, a citation or a computation. A cited or computed sentence that is to go first loses its anchors this way, then is edited like any prose. A citation a live computation names is refused, naming the computation: retire or recompute it first (ruling L6). Retiring an anchor that answers an open gap leaves that gap owed. Retiring one that answered a closed gap leaves the closure standing, and the gap's lens is told |
| `help/edit.md:28` | "EVIDENCE NEVER STANDS BARE" |
| `help/reproduce.md:21`; new paragraph | "A COMPUTATION THE REPORT HOLDS NO SENTENCE FOR IS REFUSED", mirroring `help/verify.md:19`. The text promising retire is now true |
| `help/close.md:11`; `help/mint.md:13`, `:15` | `source` and `computation` close only on a live answer, under every disposition, `moot` and `not_a_defect` included. Retiring the answer of an open gap leaves it owed; a closure stands whatever later happens to its answers. A gap about an anchor, or one carrying a reading of a citation, is met by retiring that anchor, and you still close it; a reading's `source` gap is met too by its reader's own later backing reading, and not by another lens's. A demand you no longer hold is not closed: put the gap before the bench (`motion docket --id G`), whose closing ruling is not held to the demand. A `source` gap's fix that rewords the sentence leaves it unanswered: the sentence's citation answers nothing until blue cites on it with `--answers G`. `help/mint.md`: when your list says the answers to a gap you closed have all been retired, a defect still in the report is a new gap, minted with `--supersedes G`, and it takes a mint |
| `show board --help` (`cli/seat/verbs.go`) | `answers_retired_event`: on a closed gap, the event of the retire that left none of its answers live; the closure stands, and its lens is told |
| `help/cite.md`, `help/compute.md` | `--answers`: the gap this anchor answers; a `source` gap is settled only by a citation that names it, so after a reword cite the source on the reworded sentence with `--answers G` (where it replaces the old citation, retire that one with `--answers G` first). How the anchor leaves (`retire --anchor`, sentence kept) |
| `cli/seat/verbs.go:289`, `:661` (`show changes` help and its `--id` flag) | "every edit and evidence-anchor retire" |
| `cli/seat/verbs.go:295` (`show evidence` help) | "…or of a retired citation, is in the rows" → retired anchors are not in the rows: `--anchor <id>` returns one with `retired` |
| `record/blueanswered.go:57`, `:78` | "(an edit or an evidence-anchor retire answering it)" |
| #1647's failed-check sentences that offer a close to a successor: `requireFailedCheckAnswered`'s found-against, unread and unsound messages (`failedcheck.go`), `failedClears` (`passblockers.go`), and the `failed check` term | where they say "a close that hands the defect on names its successor with `--superseded-by`", add: "on a `source` or `computation` gap that close is still held to the live-answer demand" (ruling Source demand). Re-read on #1647's merged head; every sentence that offers the successor close is in the sweep |
| `gapacts.go` header ("AN ACT ON A GAP IS A RECORDED EVENT THAT NAMES IT") | names the reaching retire: a retire of the anchor a gap is about, or whose reading it carries, is blue's act on that gap, read off the gap's own fields |
| `scripts/agentgen/src/fragments/blue-search-and-cite.md:14–18`, `:28`, `:31–32` | one lifecycle for both. ":31–32 a new computation with the old one retired" is now an act the tool admits: retire the old computation as the answer to the gap, then compute again on the kept sentence as the answer to the gap. A gap that asks for a source is answered only by a citation placed as its answer: a reword leaves it unanswered until you cite on the reworded sentence as that gap's answer |
| `scripts/agentgen/src/blue-researcher.md:113` | `awaiting_evidence: true` with `answer_kind`: a program or a source is owed from YOU |
| `scripts/agentgen/src/red-lens-evidence.md`, `red-lens-computation.md` | a `source`/`computation` check kind is a demand for a live answer. Do not demand retirement through a `document` gap: mint about the anchor, or let your reading's gap carry it (L1). Your own reading of a citation that now backs the sentence meets the `source` demand of the gap your earlier reading opened; another lens's reading does not. A demand you no longer hold goes before the bench; you do not close it. An item saying the answers to a gap you closed were retired is news, not debt: read the sentence, and if the defect is still there, mint it as a new gap superseding the old one |
| terms.json `retire` (:779), `anchor` | "retiring an evidence anchor alone — a citation or computation — keeps its sentence". Regenerate `docs/vocabulary.md` |
| FEOV `docs/citations-and-bibliography.md:146`, `:158`; `docs/propagation-and-anchoring.md` | the one lifecycle |
| `seatprobe/boards.go:288` (retire `Because`) | "deleting the sentence is refused while evidence stands on it; retire the anchor first" |

#### III.1′.6 migrate — one translation (contract: record migration)

| Change | Translation | Asserted on every archive |
|---|---|---|
| `Cite.answers` (its field number derived at #1647's merge) | none: no archived cite holds one. remap classes the field `id` (`remaplists_test.go`) | zero refusals |
| DDL: the `cite.answers` column; the `evidence_anchor` view; the `gap` view | one epoch move, E1′ (`record.EventSchema` from `requirements.json` by `schemagen`). `TestEveryDDLChangeBumpsTheEpoch` fails without it; `schema.epoch.json` restamped with `UPDATE_GOLDENS=1`; the `schema.sql` golden read line by line | — |
| `Close` under the demand | none: the demand is in the `close` verb, and replay writes through `record` with no verb on its path (`internal/record/migrate` links no `internal/cli`) | zero refusals, over the 9 archived red closes on `source` gaps, none with an answer (`Cite.answers` is new): sleeper-service-plan 2, quadratic-formula 3, is-91-prime-a 1, b6 1, wave-a 1, m23 1 |
| The widened L1 and the backing arm | none: no archived Verify carries `carried_by` (#1647 §III.9: migrate does not fill it from the `found_by` chain), so arms 3 and 4 flip no archived gap; the flip census counts them per archive to show it (0 expected, measured) | zero refusals |
| Render and derived state | per-document `[^K<n>]` numbers, evidence.md without retired entries, claim counts, `evidence_answered`, `answers_retired_event` | goldens and the flip census (V1a′.6). No gap's `open` changes |

**Residue, accepted** (each counted by V1a′.6's flip census over the 33 archives):

| Residue | Why it stays |
|---|---|
| Archived edits that left a `K-` bare, and claim retires that took a `K-` with its sentence | what a seat did. Replay exempts refusals (`Migrating`) |
| Archived `source` and `computation` gaps closed on no live answer (the 9 `source` closes; any `computation` close whose answer was later retired) | a closure is history (ruling Closures are history) |
| Archived closed gaps whose answers are all retired: `G-a0c6ca1e` (close 752, answer `K-bc8ac2b6` retired 779) holds `answers_retired_event` 779 on the final state; `G-221322e1` (close 701, answer retired 722, bench 792) holds it only on prefixes before 792 | each closure stands. No seat of an archived run saw the item |
| Two definitions of "live": the SQL `evidence_anchor` (not retired) and `record.LiveSentence` (in the report and not bare). An archived bare anchor that was never retired satisfies a demand in SQL and is refused by `verify`/`reproduce` | a bare unretired anchor is producible only by an archived edit (live edits refuse a bare anchor), and the flip census counts the archived bare anchors that answer a gap |
| Re-derived blockers and claim counts differ from what an archived run's seats saw | derived views; the archived verdict is the record's |

#### III.1′.7 The release fuzz drive (contract: release gate)

Census L6. The drive (`releasegate/fuzz/fuzz_test.go`) exercises the acts 1a′ adds and the
refusals it adds, and the tally (`sweeptally_test.go`) fails a sweep that never reaches them:

| Arm | Drive | Tally gate (`sweeptally_test.go`, judged over the four parts) |
|---|---|---|
| Retire a computation | the unsound-computation branch (`:1111–1129`) retires `--anchor K-` `--answers G` on a coin, then computes again on the kept sentence `--answers G` | `computationRetires` > 0 |
| Source demand | a `source` gap's close is driven unanswered first (refused, Conflict), then answered by `cite --answers G`, then closed | `sourceClosesRefused` > 0 and `sourceAnswers` > 0 |
| A closed gap's answer retired | on a coin, blue retires the anchor answering a closed gap; the drive reads the board (gap closed, closure unchanged, `answers_retired_event` set) and the lens's list (the item, `blocks: false`), and on a coin the lens mints a gap superseding it, after which the item is gone. A PASS the board otherwise permits lands with the item standing | `closedAnswersRetired` > 0, `retiredItemsCleared` > 0 |
| L1 | a computation or source gap about an anchor: blue retires the anchor with no `--answers`; blue's list holds no item for it; the lens's `close` is admitted | `retiredAboutCloses` > 0 |
| L6 | `retire --anchor C-` of a citation a live computation cites → refused naming the `K-` | `citedRetiresRefused` > 0 |
| A reading's `source` gap | #1647's refuting `verify` arm mints its gap; on a coin its `--check-kind` is `source`. Then one of: blue retires the citation (the lens's close admitted); blue edits the sentence (the lens's close refused by the demand, Conflict) and then the reader re-reads it backing (admitted); the reader re-reads it backing with no edit (admitted) | `readingDemandRetired` > 0, `readingDemandRefused` > 0, `readingDemandReread` > 0 |
| A reaching retire gives the lens its turn | after a reaching retire with no `--answers`, the next plan engages the minting lens on the gap; a plan that readies blue alone on it fails the run | `reachingRetireTurns` > 0 |
| Assembly screen | the capture step of a simulated debate that ends with a live unsound `K-` reports `assembly-screen` FAIL | `screenedUnsound` > 0 |

The consistency oracle runs over every simulated debate (`internal/consistency.Check`). Its raw walk is
unchanged, since a retire moves no closure, so a projection that lets a retire move one fails the run.

**Not claimed here.** #1642's sweep data (relayed by a peer session) shows reopened anchors and
unsound gaps holding 36 of 37 drawn fuzz runs. 1a′ does not claim to change that number: the arms
above are gates that the new acts run, not a forecast of how the sweep ends.

### Part 1a″ — the judged sentence (epoch E1″) (rulings L5, Staleness)

Built on 1a′'s merged head: the kind rows, `LiveSentence` and `evidence_anchor` are 1a′'s, and #1647's
check states are on main. 1a″ adds one record field (`Reproduce.claim`), one render function
(`checkLines`), the judgement JSON's `claim` and one oracle rule. **It is one complete concept:** a
check records the sentence it judged, for both kinds, and a check of an earlier wording surfaces in
evidence.md and fails the consistency oracle. Every carrier of that concept is here (the verb, the
record, the ledger's `anchorCheck`, both render surfaces, the JSON, the oracle, the help, the
computation lens's text, migrate, the fuzz). It moves no blocker, no work list and no closure, and
nothing in 1a′ waits on it.

#### III.1″.1 The lifecycle rows 1a″ adds

| Stage | One rule (every evidence anchor) | Today (main `a31a7b06`) | Kind row supplies |
|---|---|---|---|
| Check records its sentence | Every check records the sentence it judged, read by the tool from `record.LiveSentence` at the write and never typed by the seat: `Verify.claim` today, `Reproduce.claim` [NEW] (ruling L5). Both verbs declare it with `seat.Supplies(c, "claim", …)` (`verify.go:117`) | `Reproduce` holds the anchor and no sentence (`record.proto` `Reproduce`), so a judgement cannot say which wording it was of | `checkTable` |
| Staleness | ONE rule decides whether a check stands, for both kinds, and it is the ledger's: an edit that changes the anchor's sentence names it in `BlueEdit.reopened`, and only a check from a sitting opened after that edit stands (`anchorStanding.sinceEdit`, `anchorchecks.go:147–149`, fed by `verify` and `reproduce` alike, `:245–275`). The recorded sentence is how staleness SURFACES: (1) evidence.md prints, under each check of either kind, "of the sentence: …" when its recorded sentence is not the anchor's sentence now (III.1″.3); (2) the consistency oracle (`internal/consistency.Check`, run over every simulated debate of the release fuzz) gains one rule, iterated over `EvidenceTables` (ruling Staleness): a STANDING check whose recorded sentence differs from the anchor's live sentence is a violation, naming the anchor, the check's event, its state and both sentences | citations: the ledger rule plus evidence.md's "of the sentence" line (`citations.go:100–107`); computations: the ledger rule (`SoundnessReopened`, "not judged since the sentence changed", `computations.go:187`) and no sentence to show; no oracle rule for either | `checkTable` |

**The oracle rule reads #1647's check states.** "Standing" is the ledger's `stands`: a check that
reads, from a sitting after the latest edit naming the anchor, not replaced by its seat's own later
reading, of an anchor not retired. So the rule covers every standing check whatever state it puts the
anchor in: one that holds, an `unsound` judgement, and a `found_against` reading, carried or not
(#1647 §III.3; the carrier-less one is a migrated reading). An `unread` reading (`unreachable`, or any
outcome of a not-the-work copy) never stands, since it reads nothing, and is outside the rule; so is a
check with no recorded sentence (a migrated reproduce), which is counted, never folded into "agrees".
V1a″.1 forges the disagreement once per state the rule covers.

#### III.1″.2 Verbs and the record [MODIFY] (contract: CLI flag set, event schema)

| Site | Change |
|---|---|
| `lens reproduce` | writes the sentence it read to `Reproduce.claim`, declared `seat.Supplies(c, "claim", "the sentence holding the --anchor token in the report as it stands")` as `verify.go:117` declares it (ruling L5). `reproduceResult` gains `claim`, and its human line quotes the sentence judged, as `verifyResult.human` does (`verify.go:696`) |

```proto
// Reproduce [MODIFY]: + optional string claim = <next free Reproduce field; 13 at a31a7b06>;
//   the sentence this judgement is of, read off the report by the tool (record.LiveSentence) at the
//   write, as Verify.claim is (ruling L5). Not `required`: migrated reproduces hold none, and
//   CheckRequired has no migration exemption; the live writer sets it on every write (V1a″.1).
//   remap class: data, beside Verify.claim and Retire.claim (`remap.go:135–138`).
```

The ledger (`anchorchecks.go`): `anchorCheck` gains `sentence`, from `Verify.claim` or
`Reproduce.claim` ("" for a migrated reproduce), the one field the Staleness row's evidence.md line
and oracle rule read. `Reproduce.claim` has no other reader in the ledger: standing stays the edit
rule's.

#### III.1″.3 Render [MODIFY]

| Site | Change |
|---|---|
| `report/citations.go:85–110` `readings` → `checkLines` [MODIFY, one function for both kinds] | renders every check of the anchor, one per list item (seat, epoch, outcome or soundness, page, text or note), from `EvidenceJSONOf`'s rows, which carry `claim` for both kinds, and from that `claim`: nothing when it equals the anchor's live sentence; "of the same sentence as the check above" when it repeats the one above; else "of the sentence: <recorded>" — the staleness surface. An absent sentence (a migrated reproduce) prints no sentence line, as an empty `Verify.claim` prints none today (`citations.go:102`); its standing still comes from the ledger |
| `report/computations.go:173–191` | a computation's judgements are listed one per line through `checkLines`, where today only their notes are folded into one line |
| `record/evidenceview.go:404` verification JSON `claim`; the computation judgements' JSON [MODIFY] | each judgement of `show evidence` gains `claim`, the sentence it was of, as each verification carries it |

#### III.1″.4 Agent-facing text [MODIFY] (contract: prompts and constitutions)

| Carrier | Becomes |
|---|---|
| `help/reproduce.md`, new paragraph | the sentence holding the anchor is read off the report and recorded with your judgement; an edit that changes it leaves your judgement standing on an earlier wording, owed again |
| `cli/seat/verbs.go` (`show evidence` help) | each judgement carries `claim`, the sentence it was of |
| `scripts/agentgen/src/red-lens-computation.md` | your judgement records the sentence it was of; reading a computation's evidence shows that sentence beside each judgement, and one of an earlier wording no longer stands |

#### III.1″.5 migrate — one translation (contract: record migration)

| Change | Translation | Asserted |
|---|---|---|
| `Reproduce.claim` | none: no archived reproduce holds one, and migrate reads no sentence back into it (the report an archived judgement saw is not re-derivable at its event). remap classes the field `data` (`remaplists_test.go`) | every migrated reproduce has no `claim`; zero refusals on every archive |
| DDL: the `reproduce.claim` column | one epoch move, E1″, derived from main when 1a″ starts; `TestEveryDDLChangeBumpsTheEpoch`; `schema.epoch.json` restamped | — |

**Residue, accepted:** archived reproduces carry no `claim` (19 in the 2026-10-30 archive; every
archive's count in V1a″.2's census). Which wording an archived judgement judged is not on its event.
evidence.md prints no sentence line for it; the ledger's edit rule still decides whether it stands,
and the oracle rule skips it, counted.

#### III.1″.6 The release fuzz drive (contract: release gate)

| Arm | Drive | Tally gate |
|---|---|---|
| Recorded sentence | every `reproduce` the drive records carries `claim` equal to the anchor's live sentence at the write; one that does not fails the run | `reproducesWithClaim` = `reproduces` |
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
| 1c, 2 epochs | E3, E4 the next free | still the next free when each lands, after #1647's, E1′ and E1″ | 1a′, 1a″ |
| 2 wrong crop | "`retire` + cite again (the existing lifecycle)" | `retire --anchor C-…` keeps the sentence, then cite again: the same act as a wrong figure | 1a′ |
| 2 `Cite` fields | `show = 26` … `crop_sha256 = 34` | after `Cite.answers`, whose number is derived when 1a′ starts; Part 2's nine fields take the next nine | 1a′ |
| 2 a refuted crop | a contradiction | #1647's in-act gap: the evidence lens's `verify --as refutes` of the crop's citation mints its gap at the sentence, carried by `Verify.carried_by`, and the `found_against` check holds the PASS until the reader's own backing re-read, an edit, or a retire | #1647 |
| 1c, 2 `assembly-screen` on a shown block's anchor | not stated | fails on a live anchor whose check fails, through `record.FailingAnchors` and the row's `failed` (III.1′.1 Capture) | 1a′ |
| 1c, 2 a shown block's anchor answering a closed gap | not stated | retiring it leaves the closure standing, and its lens's list says so (III.1′.3 The work lists) | 1a′ |

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
| L2 | kind inferred from a side field or a bool: `grep -rnE '\bcomputation +bool\b\|\.computation\b\|Detail *[!=]= *""' --include=*.go $F/tools/internal/{record,capture,report} $F/tools/releasegate` (`rev-r1/L2.txt`) | non-test 24 lines / 9 files: `anchorchecks.go` 3, `available.go` 1 (`anchorOwnerOf`), `dispatch.go` 1 (`:1299`), `evidenceview.go` 1 (`:569`), `failedcheck.go` 4, `passblockers.go` 9, `sitting.go` 1 (`:413`), `unchecked.go` 3, `record.pb.go` 1 (false). The release fuzz infers a blocker's kind from `Detail` at 7 lines: `fuzz_test.go:1087`, `:2857`, `:2867`, `:2892`, `:3265`, `:3563`, `:4129` | `evidenceview.go:569` reads `state()` (1a′, III.1′.4). The blocker sites and the fuzz's `Detail` inferences are #1647's: its `Blocker.CheckState()` replaces `Detail != ""` as the kind test (#1647 §III.3), re-read on its merged head | 1a′ (1) |
| L3 | the computation debt: `git grep -c 'awaiting_computation\|AwaitingComputation\|computation_answered\|ComputationAnswers\|NeedsComputation\|requireComputationAnswered' -- plugins scripts` (`rev-r1/L3.txt`) | 26 files / 76 lines. Non-test: `close.go` 4, `estoppel.go` 3, `views.go` 3, `replay.go` 2, `sitting.go` 2, `viewjson.go` 11, `agentgen/src/blue-researcher.md:113` 1. Test: `cli/lens/unreadable_test.go` 5, `record/checkkind_test.go` 5, `record/queries_parity_test.go` 4, `recordsql/viewfamilies_test.go` 4, `record/workstate_test.go` 1, `fuzz_test.go` 1. Generated: the manual golden 4, `schema.sql` 3, and the 11 `agents/*.md` (2 each, 3 in `blue-researcher.md`), which carry `awaiting_computation` in the OUTPUT shapes `agentgen` inlines from the tool's manual | renamed to the `evidence` forms, with no alias (III.1′.3); the 11 agent definitions and both goldens regenerate (`agentgen`, `UPDATE_GOLDENS=1`), and `agentgen -check` fails a stale one | 1a′ |
| L4 | `CHECK_KIND_SOURCE` readers: `grep -rn 'CHECK_KIND_SOURCE' --include=*.go $F/tools \| grep -v pb.go` | 0 | `Gap.AnswerKind` and `evidence_answered` through `EvidenceTables.check` become its readers | 1a′ |
| L5 | anchor retirement in agent text: `grep -n -i retire` over `cli/seat/help/*.md`, `scripts/agentgen/src/**`, debate.js (`skills/research-protocol/scripts/debate.js`; `audit-r1/L5-help.txt`, `L5-agentgen.txt`: 23 lines, none in debate.js's authored text) | 14 on anchors (`retire.md:9,11`; `edit.md:28`; `verify.md:11,15,19`; `reproduce.md:21`; `close.md:11`; `mint.md:15`; `blue-search-and-cite.md:14–18,28,31–32`; `lens-audit-and-mint.md:6`); 9 on claim or lens retirement (`dispatch.md:13`, `spot-check.md:7`, `blue-additive-only.md:1`, `blue-research-protocol.md:9`, both preloaded-skill sources, `red-chair.md:14`, `lens-duties.md:5`) | the 14 change (III.1′.5), except `verify.md:11,15,19` and `lens-audit-and-mint.md:6`, which #1647 rewrites (re-read on its merged head for the one lifecycle); the 9 are unchanged | 1a′ |
| L6 | the release fuzz drive: `grep -nE 'retire\|"cite"\|compute\|reproduce\|"close"\|contradiction' releasegate/fuzz/fuzz_test.go` (`fuzz-verbs.txt`, 48 lines; plus the contradiction drive) | `retire --anchor` only on the citation branch (:1095); an unsound computation is edited, never retired (:1111–1129); computation gaps are answered before close (:3920–3947, :3946, :3964, :3969). The contradiction drive is #1647's to delete and replace (its §III.11) | 1a′'s arms (III.1′.7), on #1647's reading drive; 1a″'s two arms (III.1″.6). Each arm has a tally gate | 1a′, 1a″ |
| L7 | readers of a check's recorded sentence: `grep -rn 'GetClaim()\|\.Claim\b' --include=*.go $F/tools/internal \| grep -v '_test\|pb.go'` (re-run at `f5c09519`, `audit-r1/L9.txt`) | 22 lines / 9 files. `Verify.claim`: `report/citations.go:28` (labelled corroboration filter), `:100` (the "of the sentence" line), `record/evidenceview.go:404` (verification JSON), `:434`, `:437` and `record/citationid.go:102` (a labelled corroboration's location), `citationid.go:533` (contradiction key), `anchorchecks.go:229`, `:264` (a corroboration's hold), `migrate/gapplace.go:81` (a corroboration's placement), `consistency/consistency.go:272` (placed ids), `cli/lens/verify.go:102`, `:459`, `:483`, `:510`, `:551`, `:696` (the verb's own write and output). `Retire.claim`: `report/assemble.go:933` (`withdrawnClaims`), `cli/blue/retire.go:207`, `:215`, `:474`, `:476` | K- counterparts (III.1″.2, III.1″.3): `:100` → `checkLines`; `evidenceview.go:404` → the judgement JSON's `claim`; `verify.go:102`, `:696` → `reproduce`'s write and human line; the ledger's `anchorCheck.sentence` (new reader of both). Corroboration-only, no counterpart (red never authors a computation): `citations.go:28`, `evidenceview.go:434–437`, `citationid.go:102`, `anchorchecks.go:229`, `:264`, `gapplace.go:81`, `consistency.go:272`, `verify.go:459–551`. `citationid.go:533` is #1647's to delete with the contradiction. `Retire.claim` is another field and unchanged | 1a″ |
| L8 | closure readers, read to prove none moves: `grep -rn 'closure_act\|close_event\|HasClosed\|ClosedEpoch\|closedAt\|\.Closure\b' --include=*.go $F/tools/internal \| grep -v '_test\|pb.go'` (`rev-r1/g1-sweep.txt`, 89 lines / 20 files), plus the Go folds over `*recordpb.Close`: `grep -rn '\*recordpb\.Close\b' --include=*.go $F/tools/internal \| grep -v '_test\|pb.go'` (16 lines) | the view and its columns (`views.go` 9); `gapStatesOf` and its 13 downstream readers; `convergence_at`; `engagementOf`, `closingDutiesOf`; `priorClosureEpochs`; the oracle's raw walk | none changes: a retire moves no closure (ruling Closures are history). V1a′.3 holds them level | 1a′ |
| L9 | writers of a closing `Close`: `grep -rn 'ClosurePayload(' --include=*.go $F/tools \| grep -v _test` | 2: `cli/lens/close.go:39`, `cli/chair/carry.go:18` (and migrate's replay, which calls neither verb) | the demand stays in the verb `close` (III.1′.1); `carry` restates a closure and lands on no open gap | 1a′ |
| L10 | readers of an anchor's answers: `grep -rn 'GetAnswers()\|"answers"\|\.Answers\b' --include=*.go $F/tools/internal \| grep -v '_test\|pb.go'` (`rev-r1/g6-answers.txt`, 43 lines) | readers of `Computation.answers`: `gapacts.go:60–61`, `grounds.go:119`, `estoppel.go:284`, `views.go:569`, `:572`; the rest read `BlueEdit.answers` or a claim-kept `Retire.answers` (`blueanswered.go`, `manifestowed.go`, `impasse.go`, `estoppel.go:59`, `:122`, `edits.go`, `changes.go`, `changesjson.go`), or are the flag and its writers | the five read `evidence_anchor` (III.1′.3 Readers of an anchor's answers); the rest read neither kind's anchor answers, unchanged | 1a′ |
| L11 | seat-facing Go strings naming a retire or a failed check by one kind: `grep -rn 'citation retire\|retired citation\|citation retired\|retires the citation\|retire the citation\|Retire the citation\|retire a citation\|retires a citation\|retire of that citation\|retires one citation\|retire each citation' --include=*.go $F/tools/internal/{cli,record,view,report,capture} \| grep -v '_test\|pb.go'` (`rev-r1/g5-retire-strings.txt`, 42 lines; `audit-r1/go-strings-sweep.txt`, 44) | the evidence-anchor retire spoken as a citation's: `cli/seat/verbs.go:289`, `:295`, `:661`; `view/changes.go:48`, `:157–158`, `:190–194`; `record/changesjson.go:30`, `:55`; `record/blueanswered.go:57`, `:78`, `:168`; `record/impasse.go:195`; `record/record.go:1461`; `cli/blue/retire.go:80`, `:88`, `:158`, `:161–162`, `:167`, `:185`, `:192`, `:462–474`; `cli/blue/edit.go:257`, `:285`; `recordsql/views.go:307`. Citation-only by meaning (the `unread` state, which no computation has): `failedcheck.go:165`, `:199`; `passblockers.go:237`, `:248`, `:269–271`, `:427`; `cli/lens/verify.go:653`. `record/dispatch.go:21` names both kinds' remedies | the first set names the anchor by its row's `noun` (III.1′.2, III.1′.4, III.1′.5); the citation-only set is #1647's (its found-against arm); 1a′ adds the demand clause where it offers a close to a successor (III.1′.5) | 1a′ |
| L12 | readers of every committed archive: in `$F/tools/internal/record/migrate`, `(grep -l '"\*\.tar\.gz"' *_test.go; grep -l 'archivedRuns(t)' *_test.go) \| sort -u`, plus `grep -rln 'tar.gz' $F/tools/devcmd` | 20 migrate test files (Phase 0 lists them) and `devcmd/archdump`, `devcmd/archiverun`; `publish_test.go` names one archive | each takes up Phase 0's archive | 1a′ |
| L13 | hand-written UNIONs over the evidence tables: `grep -rn 'FROM "computation"' --include=*.go $F/tools/internal \| grep -v _test` and the same for `FROM "cite"` | `record/id.go:86–87`, `record/refs.go:130`, `:204`, `record/grounds.go:119–120`, `estoppel.go:284`, `views.go:569–572`; per-kind by design: `mintbudget.go:61`, `:64`, `viewjson.go:524`, `citationid.go:232`, `:278`, `refs.go:113` (a computation's bundle) | the first set reads `evidence_anchor` (One mechanism; III.1′.3) | 1a′ |
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
| R13 | **Re-derived archived state (1a′, 1a″)**: `evidence_answered` turns archived open demand gaps `awaiting_evidence`, or not; a reaching retire becomes blue's act on archived gaps, so re-derived turns and exchange counts differ from what the seats saw | L × L × L | No closure moves. Replay exempts refusals (`Migrating`) and passes no verb. Each part's flip census (V1a′.6, V1a″.2) counts every change per archive and commits the counts |
| R14 | **Retire is cheap for a computation (1a′)**: blue retires an unsound `K-` and keeps a sentence nothing now backs | M × M × L | The same as for a citation today. The kept sentence is prose the logic lens audits, the claim count drops, and the CHANGELOG lists the retire |
| R15 | **A generated SQL arm drifts from its table row (1a′)** | L × M × L | the `evidence_anchor` arms are built from `recordsql.EvidenceTables`; `TestEveryBackingKindHasAnEvidenceRow` checks each row against the DDL; the `schema.sql` golden |
| R16 | **A sentence change no edit names (1a″, L5)**: the ledger decides standing from `BlueEdit.reopened`, so an edit that changes an anchor's sentence without naming it leaves a check standing on an earlier wording | L × M × L | the consistency oracle's rule (III.1″.1 Staleness) fails such a record in every release-fuzz debate; in a real run, evidence.md prints the check's own sentence beside the anchor's, so the reader sees the mismatch. The oracle does not run in a live debate: residue |
| R17 | **A closure stands on answers no longer in the report (1a′)**: a defect it closed may still be there, and nothing reopens it. A close whose read of a live answer races a retire can land just after the retire, leaving no item, because the check is outside the write's lock | L × M × L | The ruling accepts it (closures are history). Measured: 2 such closed gaps in 33 archives, 1 of 217 red closes where the closure rested solely on the answer, and that gap was bench-closed anyway. An edit of an answering anchor has always left the closure standing, with `unchecked:K` holding PASS until the sentence is checked. Blue's retire leaves the sentence for every lens that audits the report at the next head. The lens's item names it. A racing close is the same closure a moment later and costs one item |
| R18 | **An item no seat reads (1a′)**: a lens that does not sit again after the retire never sees it, since readying it would hold PASS | M × L × L | Fork A lens that does not sit again. The board's `answers_retired_event` stays; both measured cases were read at the lens's next sitting; V1a′.6 counts them per archive |
| R19 | **The release fuzz still ends CEILING on other causes (1a′)**: #1642's sweep data, relayed by a peer session, shows reopened anchors and unsound gaps holding 36 of 37 drawn runs | H × M × L | Not claimed fixed. 1a′'s gates (V1a′.9) prove its acts run and its refusals fire; the sweep's end states are read from the tally's carried #1642 fields (`final_blockers`, #1647 F7), and a CEILING on another cause is that cause's issue |
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
| V1a′.2 | Closure demand at `close`, per check kind. A close of an unanswered `computation` gap and of an unanswered `source` gap is refused under each closing disposition in turn — `repaired`, `repaired_with_regression` (with `--superseded-by`), `amends_prior`, `not_a_defect`, `defect_accepted`, `defect_owed_elsewhere`, `moot` — and the refusal names `cite --answers` / `compute --answers` and `motion docket`; a regrade of it is admitted. A live answer of the demanded kind admits the close; one of the other kind does not; a retired one does not. **`carry` (round 1's gap, under this design):** `carry --carried-from <epoch>` on an open demand gap with no `close` row → refused as naming no closure, and the gap stays `open` and `awaiting_evidence`; on a closed gap whose answer has since been retired → admitted, the gap stays closed, and no demand is asked of it; on a bench-closed gap with no `close` row → refused. The bench's closing ruling on an unanswered `source` gap (filed by `motion docket`) closes it. L1: a gap about `K-x` minted `--check-kind computation`, and one about `C-y` minted `--check-kind source`, each after `retire --anchor` of its anchor → red's close admitted, nothing auto-closed. Blue's sitting is held on an awaiting `source` gap as on a `computation` one. Delete-the-row: delete the `requireEvidenceAnswered` call from `cli/lens/close.go` → every refusal arm fails; exempt `moot` → its disposition arm fails; drop the `about_ref` arm of `evidence_answered` → the L1 arms fail; drop `retired_event` from `evidence_answered` → the retired-answer arm fails; drop the `len(prior) == 0` refusal in `record.go` → the carry-on-an-open-gap arm fails | any arm passes the wrong way |
| V1a′.3 | Closures are history and the two work lists, one script through the CLI. **A closure stands.** A lens mints computation gap G and source gap H; blue answers G with `compute --answers G` and H with `cite --answers H`; the lens closes both in epoch 2; blue retires both answers in epoch 3. Then every reader says G and H are closed, as before the retire: `show board --json` (`open: false`, the closure payload, `closed_epoch` 2), `board_counts`, `convergence_at` at the newest point, telemetry, the scorecard, the board chart, the spot-check archive, `consistency.Check` (no violation). `answers_retired_event` is the later retire on each. With two answers live, retiring one → null and no item; both → set and the item. A close recorded after the last retire (the gap being open, on another live answer or L1) → null and no item. **(b) The closed gap's item.** The owning lens's `show work` lists it, `blocks: false`, naming G, the event and the anchors, and `complete` is true; blue, the chair, the bench and another lens list nothing for it. With no other blocker the plan's `pass_permitted` is true, nobody is readied by it, and the chair's PASS is admitted while the item stands. It is listed throughout the lens's next sitting and gone once that sitting closes; a sitting stopped at the call limit leaves it listed. A mint naming G in `supersedes` removes it at once, is admitted on the closed ancestor, strands nothing, and takes one mint of the budget (`mint_budget` moves by one); a mint not naming G leaves it. At no mints left it stays listed, unblocking, with the budget clause. A bench closing after the retire nulls the column and removes it; a chair `carry` after the retire does neither. On a gap the bench closed, the item says what its ruling settled stays barred. **(a) The open gap a retire met.** Three gaps: a `computation` gap about `K-x`; a `source` gap about `C-y`; and a `source` reading gap G′ that #1647's refuting `verify` of `C-z` minted (`carried_by` G′). Blue retires each anchor with no `--answers`. For each: blue's list holds no item (and no `found_against` item for G′), `complete` is unaffected, and its `open[]` row reads `awaiting_evidence: false`, `yours_to_close: false`; the lens's `open[]` row reads `yours_to_close: true`; its list holds the "blue retired … so its demand is met" item and not "NO edit answers it"; it owes it (`lensOwesOf`), a lens at its retirement count included, until it sits through a sitting opened after the retire; the next plan engages the lens on the gap (its turn has come: `lensTurnOn`), where before the retire it readied blue alone; its close is admitted. Left unclosed and material, the gap's exchanges reach impasse and the dispatch dockets it. A `document` gap about `K-w` takes the same turn and item, and no demand column. Delete-the-row: add the item to `lensOwesOf`'s blockers → the PASS arm fails; set `Blocks: true` → the `complete` arm fails; drop the `firstAfter` filter → the gone-after-the-sitting arm fails; drop the `supersedes` filter → the mint arm fails; let a carry count as a closing act in the column → the carry arm fails; drop the closing-act arm → the bench arm fails; drop `retired_event` from the column → the two-answers arm fails; let the retire fell the closure (`open` read from the column) → the "A closure stands" arm fails; drop the `gapNewsOf` row → the owes arm fails; drop the `blueAnswers` arm → the "NO edit answers it" assertion fails; drop the reaching-retire arm of `gapActsOf` → the plan-engages-the-lens arm fails for all three gaps; drop the `carried_by` half of it → the G′ arm alone fails | a closure moves, the item holds PASS, blocks, or never leaves, the lens is never readied on a met gap, or any gap is left with no seat able to act |
| V1a′.4 | Retry, cross-kind and answer readers. A keyed idempotent `compute` retry returns `computation_id`. `cite` keyed retries with another `--answers` → refused; `seat.AssertedCovers` over `cite` includes `answers`. `--cites` a retired citation → refused. `retire --anchor C-` named by two live `K-`s' `cites` → refused naming both (ruling L6); still refused after one is retired, naming the other; admitted after both are retired; nothing else is retired by it. The race: two `feov-record` processes, one `compute --cites C`, one `retire --anchor C`, released together twenty times on fresh records → each time exactly one admitted, and no record holds a live `K-` citing a retired `C-`. **Gap 6's case:** a `source` gap remanded at impasse, then a blue sitting whose only act on it is `cite --answers G` → the remand's exchange begins (`gapActsOf`, `impasse.go:160`), and a lens sitting after it is not refused for an unbegun exchange; a citation placed `--answers G` after a ruling on G is new grounds about G (`motion docket --grounds C-…`). Delete-the-row: move the L6 check out of `record.validate` into `cli/blue/retire.go` → the race arm fails; drop the `Cite` arm from `gapActsOf` → the remand arm fails; drop it from `groundsActSQL` → the grounds arm fails | any |
| V1a′.5 | Render goldens. Two documents referencing `K-a`, `K-b` in opposite orders → each numbers `[^K1]`, `[^K2]` by its own first appearance, and a retired `K-` takes no number. evidence.md holds no retired `C-` or `K-`. "anchored to nothing" appears only for a live computation no document references. `show changes` of a `K-` retire reads "retired computation K-…, its sentence kept". Read every changed golden line (difftest `manual`, report, evidence) | a retired anchor rendered, a gap in numbering, a retire named by the wrong kind, or a lost duty on a help page |
| V1a′.6 | Migrate at E1′. All 33 archives (the 32 committed before 1a′ and Phase 0's): zero refusals; `Cite.answers` absent. **Replay does not run the demand:** the test counts the archived red closes on `source` gaps, asserts 9 (per archive as III.1′.6, so the arm is not vacuous) and all 9 admitted, through `TestEveryArchivedRunMigrates`'s replay (`record`'s `Batch`, no verb). Delete-the-row: move the check into `record.validate`'s `Close` arm → this test fails on the first of the 9, which is what proves the location. The flip census, written by a `devcmd` (`devcmd/flipcensus`) and committed as testdata: per archive, each closed gap with `answers_retired_event` (the fixture's `G-a0c6ca1e`), each gap whose `awaiting_evidence` differs between E1 and E1′, the gaps met by arm 3 or arm 4 (0 measured: no archived Verify carries `carried_by`), each reaching retire that becomes blue's act on a gap, the claim count before and after, and each never-retired bare anchor that answers a gap (the two-lives residue). No gap's `open` differs between E1 and E1′. Measured, not forecast; a change in a count fails the test | a refusal, a count off its committed row, an uncounted flip, a gap whose `open` moved |
| V1a′.7 | Real data: the 2026-10-30 CEILING. (a) The smoke's own archive, `run-archive/2026-10-30_mersenne-31-prime-sum.tar.gz` (Phase 0), extracted with `recordtest.ExtractArchive` and brought through `feov-migrate` from epoch 49 to E1′. Its evidence.md holds none of `K-414ee920`, `K-72d5d71a`, `K-71a99b9a`, `K-bc8ac2b6` and no "anchored to nothing". Each prefix below is the migrated events through that id replayed into a fresh record under `Migrating`; the test finds each id from its row (the retire naming the anchor, the sitting's `sitting_close`, the closing act), and the numbers here are the archive's at epoch 50. All of the seat is `red-lens-computation`. `G-342d3cdc` (about `K-71a99b9a`, retired 722; bench 881) and `G-78817cfa` (about `K-bc8ac2b6`, retired 779; bench 882): at the final state `evidence_answered` is true and each is closed by the bench. At the prefix through 721, `G-342d3cdc` is `awaiting_evidence` and blue's list holds its item. At the prefix through 760 it is not, blue's list holds no item for it, and the lens's list holds the "blue retired K-71a99b9a … so its demand is met" item. The plan computed on the prefix through 723 engages `red-lens-computation` on `G-342d3cdc` (the reaching retire 722 is blue's act on it); the test first asserts that no other blue act on `G-342d3cdc` lies between its mint and 722, so the arm is the retire's alone, and dropping the reaching-retire arm of `gapActsOf` leaves the plan readying blue alone. On the prefix through 880, red's `close` of each through the CLI is admitted (L1). `G-a0c6ca1e` (about `K-71a99b9a`, retired 722; red `moot` close 752; its answer `K-bc8ac2b6` retired 779): closed at every prefix from 752 and at the final state, with `close_event` 752. `answers_retired_event` is 779 from the retire on. The lens's list holds its item at the prefix through 810, inside the sitting opened 803, and not at the prefix through 823, that sitting's close. `G-221322e1` (red close 701 on `K-71a99b9a`; bench close 792): closed at every prefix from 701, `convergence_at` counts it closed at every point from 701, and no reader moves at 722. `answers_retired_event` is 722 at the prefixes from 722 to 791 and null from 792. The lens's list holds its item at the prefix through 723 and through 760 (the sitting opened 742) and not at the prefix through 767 (its close) or through 791. So `G-221322e1`'s closure stands, and its lens's item shows at the record prefixes between 722 and 792 (through 723 and 760) until that sitting's close. On no prefix does either item appear in `passBlockersOf`. Delete `answers_retired_event` from the WORK query → the item arms fail; let a retire reopen → the `G-221322e1` arm fails at 722. (b) A scripted replay through the CLI at E1′. `compute` `K1` on S; the lens mints a computation gap G on S. Blue's edit moving `K1` to a bare position → refused. `retire --anchor K1 --answers G`, then `compute --answers G` on S → red's close of G admitted, and no computation gap stays open on a retired `K-`. A `source` arm: `cite` `C1` on T; the lens mints a `source` gap H on T; blue rewords T (C1 stays) → red's close of H refused, naming `cite --answers`; blue cites on the reworded T `--answers H` → admitted. (c) `scripts/universe.sh build`, grep the staged marketplace (`$WORKDIR.marketplace/plugins/frank-exchange-of-views`) for `retire --anchor` on a computation; then `--smoke` on the same question. What is asserted is the dead end, not the run's outcome: the run records no close refusal of a gap whose demand a retire met, no plan readies blue alone on such a gap after the retire, every such gap is closed by its lens or reaches the bench, evidence.md holds no retired anchor, and capture's `assembly-screen` and `computation-rerun` PASS. The run may still end CEILING on another cause (R19); that is read and reported, not asserted | any; or (c) shows a gap whose demand a retire met refused at close, or left with only blue readied |
| V1a′.8 | Agent text. `agentgen -check`, `enginegen -check`, `vocabdoc`; each changed `--help` read before and after. A failing assertion, not a golden a person reads, over the rendered pages: `retire --help` (any evidence anchor; the L6 refusal; an open gap left owed and a closure left standing), `edit --help`, `reproduce --help` (no live sentence, refused), `close --help` (every disposition; `motion docket` as the exit for a demand no longer held; a reword leaves a `source` gap unanswered until a citation answers it), `mint --help` (a retired answer's defect is a new gap, superseding, and takes a mint), `cite --help` (`--answers`; cite again on a reworded sentence), `show changes --help` and `show evidence --help` (evidence-anchor retires; no retired anchor in the rows), `show board --help` (`answers_retired_event`); and the generated `blue-researcher` definition holds the cite-as-the-answer act. Drop one sentence → it fails | a stale page or a lost duty |
| V1a′.9 | The release fuzz on 1a′'s PR, read as §V's preamble states: the four parts and the tally green on the final head, and the tally's 1a′ gates (III.1′.7: `computationRetires`, `sourceClosesRefused`, `sourceAnswers`, `closedAnswersRetired`, `retiredItemsCleared`, `retiredAboutCloses`, `citedRetiresRefused`, `readingDemandRetired`, `readingDemandRefused`, `readingDemandReread`, `reachingRetireTurns`, `screenedUnsound`) each above zero. The sweep's CEILING share is read and reported against main's latest scheduled run, not asserted (R19) | a red part or tally, a skipped fuzz step, or a gate at zero |
| V1a′.10 | The demand on a reading's gap (#1647's shape), through the CLI on #1647's verbs. (1) Readings on `C1` and `C2` of one sentence carry `source` gap G (the second reading names G, `carrier_minted` false); blue retires `C1` → `evidence_answered` true, and the lens's close is refused by `requireFailedCheckAnswered` naming `C2`, not by the demand; blue retires `C2` → admitted. (2) The source reading-gap table, one subtest per row on a copy: retire → met and admitted; edit → refused by the demand, naming the re-read and `cite --answers`; edit then the reader's backing re-read → admitted; edit then blue's `cite --answers G` → admitted; the reader's backing re-read with no edit → admitted; another lens's backing reading → refused by the demand; a closing bench ruling → G closed, the check still held in `passBlockersOf`, and no `Close` written. (3) Arm 4's ordering: `refutes`, backing, `refutes` again by the reader → not met; `refutes`, backing → met; `weak` backs; a `supports` of a not-the-work copy does not. (4) The same refuting reading minted `--check-kind computation`: the reader's backing re-read clears the check and does not meet the demand; a retire of the citation does. (5) An unread citation: an `unreachable` reading mints nothing and carries no `carried_by`; the reader's owed `mint --about-kind anchor --about C` under `source`, then blue's retire of C → met through arm 2, the gap carrying no reading. (6) A refuting `corroborate` minted `--check-kind source` → met only by `cite --answers G`. `TestTheBackingArmIsTheLedgersHolds`: every `SourceOutcome` × `SourceCompleteness` pair gives the same answer through arm 4's SQL and through `VerifyReads && !FindsAgainst`. Delete-the-row: drop arm 3 → (1) and the retire row fail; drop arm 4 → the re-read rows fail; drop "by V's own seat" → the other-lens row fails; drop the `source` restriction of arm 4 → (4) fails; read "any" Verify rather than the latest per seat → (3) fails; move the demand after `record.Append` (into the record) → V1a′.6 fails | any arm passes the wrong way |
| V1a′.11 | `assembly-screen` through the kind row. On an assembled report, a live `K-` with a standing `unsound` judgement → FAIL naming it; a live `C-` with a standing `refutes` → FAIL naming it; each retired → PASS; a record with computations and no citations → screened, not SKIP. Delete-the-row: clear the computation row's `failed` → the `K-` arm fails; restore `CitationsFoundAgainst` at `capture.go:470` → the `K-` arm fails | any |
| V1a″.1 | Recorded sentence and staleness (rulings L5, Staleness), ONE table-driven test over `EvidenceTables`' `checkTable` rows (`verify`, `reproduce`) running the same script. A check of a live anchor records `claim` equal to `record.LiveSentence`; the seat has no flag to set it (`seat.Supplies`); `reproduce`'s result carries `claim` and its human line quotes it. evidence.md: the `K-` entry prints `backs` as a `C-` entry does, and each check appears through `checkLines`. An edit changing the sentence (naming the anchor in `reopened`) → the check no longer stands for either kind, evidence.md prints "of the sentence: <old>" under it, and a fresh check from a later sitting prints no sentence line. Oracle, once per state #1647's ledger gives a standing check: a test seam drops the anchor from `BlueEdit.reopened` on that edit, over a check that holds, an `unsound` judgement, and a `found_against` reading carried by its gap → `consistency.Check` reports one violation each, naming the anchor, the check's event, its state and both sentences; over an `unread` reading → none, since it never stands. A migrated reproduce (no `claim`) prints no sentence line, stands or not by the ledger alone, and the oracle counts it as unmeasured. Delete-the-row: drop the `Reproduce.claim` write → the `K-` arm fails; drop the oracle rule, or its `computation` row → the forged arm fails for that kind; restrict the rule to checks that hold → the `unsound` and `found_against` arms fail; restore `readings` as citation-only → the `K-` render arm fails; class `Reproduce.claim` other than `data` in remap → `remaplists_test.go` fails | any arm passes the wrong way for either kind, or a disagreement passes the oracle |
| V1a″.2 | Migrate at E1″. All 33 archives: zero refusals; every migrated reproduce has no `claim`. The census by V1a′.6's `devcmd`, committed: per archive, the reproduces with no recorded sentence (19 in the 2026-10-30 archive). A change in a count fails the test | a refusal, a count off its row |
| V1a″.3 | Agent text. `agentgen -check`, `vocabdoc`; each changed `--help` read before and after. Failing assertions over the rendered `reproduce --help` (the sentence is recorded with your judgement; an edit leaves it standing on an earlier wording, owed again) and `show evidence --help` (each judgement carries `claim`), and the generated `red-lens-computation` definition. Drop one → it fails | a stale page or a lost duty |
| V1a″.4 | The release fuzz on 1a″'s PR, read as §V's preamble states, with the tally's 1a″ gates (III.1″.6: `reproducesWithClaim` = `reproduces`; `stalenessForged` = `stalenessReported`) | a red part or tally, a skipped fuzz step, or a gate off |
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
| Lifecycle (gblock, 2026-10-30; #1657) | How citations and computations live | exactly alike: a computation, like a citation, is a textual statement about external or reference material ("even a reference to a figure in text is a citation"). Retired anchors are dropped from evidence.md, and the record keeps them. A `source` or `computation` gap closes only on a live answer: `cite --answers`, and retiring the answer leaves an open gap's debt owed. Retire never auto-closes a gap. Part 1a′ is built before 1b. No compatibility path | ~~retire auto-closes the gaps its anchor answered~~ (rejected); keep two lifecycles |
| L1 (gblock, 2026-10-31; widened 2026-11-01 by Fork Joint ruling with #1647) | A gap minted `--about-kind anchor` about an evidence anchor (the failed check's carrier) under `--check-kind source`/`computation`: does that anchor's retirement meet its live-answer demand? | yes. The demand is met when the gap's `about_ref` anchor is retired, or the `anchor` of any Verify whose `carried_by` is the gap (#1647's reading gap), read from `mint.about_ref`, `Verify.carried_by`, `Verify.anchor` and `retire.anchors`; nothing is inferred. Red still closes. Safe with several readings on one gap: `requireFailedCheckAnswered` still refuses the close while another reading's check stands | ~~(b) strict: only a live anchor answering it closes it; red regrades or supersedes~~ (leaves the smoke's `G-342d3cdc`, `G-78817cfa` at CEILING); ~~(c) the retire must name the gap (`--answers G`)~~ (a retire without it leaves the gap unclosable, and a retire cannot be repeated) |
| L2 (gblock, 2026-10-31; REVERSED 2026-11-01) | A CLOSED gap whose answering anchor is retired | **Rejected:** ~~the closure falls, derived, once no anchor that answered the gap at the close is live~~. Measured: its trigger is 1 of 217 red closes on 33 archives (`G-221322e1`), and that gap was bench-closed anyway; it costs 28 KB of this plan and 89 lines across 20 files of readers (census L8); an edit of an answering anchor already leaves the closure standing with `unchecked:K` holding PASS (`graph/probeD.out`), so the fall was the one rule treating a retire otherwise. Fork Closures are history replaces it | ~~(b) the gap stays closed and a "closed on retired evidence" blocker holds PASS until red re-closes~~; ~~(c) only an open gap's debt reopens~~ (what is built, with the lens's item) |
| L3 (gblock, 2026-10-31) | Claim count | a `K-` on prose counts as a claim exactly as a `C-` does: one per distinct anchor per sentence. `CountsAsClaim` reads `backs`; the `claim` column is deleted. The logic lens's budget only grows, so no archived mint is refused | ~~(b) count sentences carrying at least one evidence anchor~~ (changes the citation count too); ~~(c) keep `K-` `claim: false`~~ (the divergence the Lifecycle ruling removes) |
| L4 (gblock, 2026-10-31; DROPPED 2026-11-01 by Fork Joint ruling with #1647) | One failed-check carrier | **Rejected:** ~~a standing `refutes`/`absent` reading of a citation is a failed check carried by the reader's gap about the anchor; a contradiction keeps only red's corroborations; migrate translates the 5 archived findings that raise an anchored contradiction~~. #1647 measured the contradiction itself as the defect — in release-fuzz run 40872228665, 22 of 45 mints per run answered a contradiction and all 37 drawn runs ended CEILING; m32's four lenses refuted one claim, raised four contradictions, and one gap answered one — and deletes it, corroborations included, with the reading's gap minted at the act and `found_against` as a failed-check state (#1647 §I, §III.0). L4 kept a second carrier for corroborations and an archive translation (5 findings, 16 readings of residue) that #1647's one translation replaces | ~~(b) an unsound computation becomes a contradiction~~; ~~(c) two carriers, stated kind-specific~~ |
| L5 (gblock, 2026-10-31; reverses the lead's recommendation) | Does `reproduce` record the sentence it judged, as `verify` records `claim`? | yes: `Reproduce.claim` (its field number derived when 1a″ starts), read by the tool from `record.LiveSentence`, never typed. Its readers are the K- counterparts of `Verify.claim`'s (census L7): evidence.md's per-check sentence line (`checkLines`), the judgement JSON's `claim`, `reproduce`'s own result, the ledger's `anchorCheck.sentence` and the oracle's staleness rule; in 1c, a shown block's referring sentence. Archived reproduces carry none (residue, III.1″.5) | ~~(a) record nothing new~~: its premise, that no surface would read the field, was wrong — `Verify.claim` is read at `citations.go:28`, `:100` (evidence.md's sentence line), `citationid.go:533` (the contradiction key, which #1647 deletes) and `evidenceview.go:404`, and the first two have a computation counterpart. (The draft also cited `assemble.go:933`, which reads `Retire.claim`.) |
| L6 (gblock, 2026-10-31) | Retiring a citation a live computation names (`--cites`; 1b `--input`) | refused while such a computation is live, naming each; blue retires or recomputes it first. The tool refuses and never cascades | ~~(b) the computation reopens when its citation retires~~ (a derived cross-kind edge the ledger lacks); ~~(c) allowed~~ (evidence.md names a citation the report no longer holds) |
| Source demand (gblock, 2026-10-31) | Does `source` bind every closing disposition? | yes: a hard live-answer demand on every disposition that closes, `moot` and `not_a_defect` included, exactly as `computation` is (`close.go:69–72`). Stated in `close --help` and `mint --help`. The draft's gloss "a demand the lens no longer holds is regraded or superseded" was wrong (round 1: a regrade carries grades only, `record.proto:1069`, and a close to a successor is held to the demand): the exit is the bench, reached by `motion docket`, and the help names it | ~~exempt `moot` and `not_a_defect`~~; ~~carry it as a risk~~ |
| Staleness surface (lead, disclosed; serving L5) | How a check of a sentence the report no longer carries surfaces | the ledger's edit rule (`sinceEdit`) stays the ONE rule that decides standing, for both kinds; the recorded sentence surfaces it: evidence.md's "of the sentence" line through one `checkLines`, and one oracle rule over `recordsql.EvidenceTables` that fails a STANDING check whose recorded sentence is not the anchor's live sentence (III.1″.1; the oracle rule ruled BUILD, Fork Staleness) | the recorded sentence as a second standing rule in the ledger (the ledger is a fold over events and renders no report; and the sitting rule is the stricter, since it also catches a check written after the edit by a seat that read before it); a runtime refusal (nothing to refuse: the check was of the sentence when written) |
| Smoke fixture (gblock, 2026-10-31) | Where V1a′.7(a)'s record lives | committed to `run-archive/` by 1a′'s first commit, as a real run (Phase 0), with `outcomes.tsv` re-measured by the `archive-rerun` job | a testdata copy of `record.db` (a second home for a run's record) |
| Show evidence after retire (lead, disclosed; stands, gblock 2026-10-31) | What `show evidence` says of a retired anchor | its lists and counts hold live anchors only, as evidence.md does. `--anchor <retired>` returns it with `retired` (event, reason), so a lens can still read what left | list every anchor with a `retired` field (the lists and evidence.md would disagree) |
| Answer of another kind (lead, disclosed; stands, gblock 2026-10-31) | `cite --answers` on a `computation` gap, or `compute --answers` on a `source` gap | accepted; only the demanded kind discharges a demand (an anchor may answer part of a gap of any kind, as `compute --answers` on a `document` gap does today) | refused at write |
| Kind-row home (lead, disclosed; stands, gblock 2026-10-31; package moved in revision r1, lead, disclosed) | Where kind differences live | two tables. `anchor.kinds` holds the vocabulary (the leaf imports nothing). The record facts (table, id column, check kind, check table, noun, failed state) are `recordsql.EvidenceTables`, held level with `anchor.kinds` and the DDL by a test, with the SQL arms generated from it. It is in `recordsql`, not `record`, because the view generator in `recordsql` reads it and `record` imports `recordsql` (an import cycle the round-1 audit found) | all in `anchor.kinds` (the leaf would import `recordpb`); `record.evidenceTables` (`recordsql` cannot import `record`); a proto option on `Cite` and `Computation` read from the descriptors (a new option kind for six values); per-site `if computation` (today's 24 sites, L2) |
| Split (gblock, 2026-10-31, after plan-audit round 1) | One part or two for the lifecycle | two: 1a′ = the CEILING fix (retire for every evidence kind, the bare-marker refusal, the closure demand and live answer incl. `cite --answers`, render after retire, L1, L3, L6, the `source` demand); 1a″ = L5 (`Reproduce.claim` and the staleness oracle rule), narrowed by Fork Joint ruling with #1647, which drops L4. Each its own PR and epoch; 1a″ builds on 1a′ | one Part 1a′ at one epoch |
| Staleness (gblock, 2026-10-31, after plan-audit round 1) | Build the consistency oracle's staleness rule over the evidence table | BUILD, in 1a″ with L5 (III.1″.1) | leave staleness to the ledger and evidence.md's line |
| L2 history (gblock, 2026-10-31; REVERSED 2026-11-01) | Which closure-history readers follow a fallen closure | **Rejected** with L2: ~~re-derive every reader (`convergence_at` per point, `HasClosed`/`ClosedEpoch`/`Closure`, scorecard, telemetry, chart, spot-check archive, dashboard, repair denominator, board JSON, the raw walk)~~. Measured: 89 lines / 20 files and 16 lines of Go folds, for 1 trigger in 217 closes. No closure moves, so no reader does | — |
| L1×L2 (gblock, 2026-10-31; REVERSED 2026-11-01) | An L1-met closure whose answering anchor is later retired (`G-a0c6ca1e`) | **Rejected** with L2 as moot: ~~an L1-met closure stands and only one resting solely on answering anchors falls~~. Every closure stands; `G-a0c6ca1e` also takes the lens's item | — |
| L1-late (gblock, 2026-10-31; REVERSED 2026-11-01) | A closure on answers, whose `about_ref` is retired after the close and before the last answer | **Rejected** with L2 as moot: no fall point exists to compare with. No archive holds the case (`rev-r1/`) | — |
| Fall in event folds (lead, disclosed; REVERSED with L2, gblock 2026-11-01) | How the Go folds that decide "closed" from `Close` events learn a closure fell | **Rejected**: nothing falls, so `engagementOf` and `closingDutiesOf` read `Close` events as ever | — |
| Demand at the write (lead, disclosed; REVERSED with L2, gblock 2026-11-01) | Where the closure demand is checked | **Rejected**: ~~`record.validate`'s `Close` arm, for every `Close` writer~~. It existed for `carry` re-closing a fallen demand gap. No closure falls, so a carry lands only on a closed gap (`priorClosureEpochs`). Fork Demand location replaces it | — |
| L4 translation (lead, disclosed; serving L4; REJECTED with L4, 2026-11-01) | What migrate translates, given the archived chain | **Rejected** with L4: ~~the 5 findings that raise an anchored reading's contradiction become findings about that citation~~. #1647's §III.9 translates the archived contradiction chains once, for every reading | — |
| Closures are history (gblock, 2026-11-01) | What a retire does to a closed gap | Closures never fall. A retire affects only one-hop work lists: no general dependency graph, no walker, no edge table (III.1′.3 The work lists; `graph/RULING.md`, `graph/census.md`). An open gap about a retired anchor: demand met, inactionable for blue, on red's list, red closes. A closed gap whose answers were all retired: the closure stands, and the owning lens's list shows an item that does not hold PASS; no new act. An open gap with a demand: a live answer required. L1, L3, L5, L6, the live-answer demand on every disposition for both kinds, the split and the staleness rule stand; L4 stands in this ruling and is dropped by Fork Joint ruling with #1647 | ~~L2 and its readers~~ (rows above); a general dependency graph over the record's edges (`graph/census.md`) |
| Demand location (lead, disclosed; serving Closures are history) | Where the closure demand is checked | the verb `close`, in place (`close.go:69–73`), bound by Fork Joint ruling with #1647, which keeps it out of the record write. No measured reason moves it: `carry` lands on no open gap, replay cannot reach it (`internal/record/migrate` links no `internal/cli`; V1a′.6), and a close racing a retire costs one item (R17) | `record.validate`'s `Close` arm with a `Migrating` skip (machinery for no measured defect); a `record.Check` guard passed to `close`'s `Append` (`landing.go:29`), which would close the race in the verb, for one lost item in 217 closes |
| Joint ruling with #1647 (gblock, 2026-11-01, relayed by the Relocations session) | How 1a′ and #1647 meet | #1647 governs a refuting reading and lands FIRST; L4 is DROPPED; contradictions are deleted everywhere, corroborations included; 1a″ is L5 and the staleness oracle rule only; 1a′ lands second, on #1647's shape (`found_against`, `Verify.carried_by`/`carrier_minted`, the reading gap placed by quote with `about_*` unset, a closing bench ruling that closes the gap while the check holds PASS). The seam: L1 widened to a carrying Verify's anchor; the backing re-read arm off rows; the source reading-gap table; an unread citation's gap on the about-anchor path; the demand stays in `close`. Part 2's refuted crop is #1647's in-act gap; `assembly-screen` reads a shared "this check fails" predicate through the kind table, not a shared state; census L14 and every contradiction string row go; epochs re-derived with 1a′ second to land (III.1′.3 The demand on a reading's gap) | the order reversed, with 1a′ translating contradictions it then deletes |
| A reaching retire is an act on the gap (gblock, 2026-11-03; serving Closures are history and L1) | How the lens is readied on a gap a retire met when the retire named no gap | `gapActsOf` counts a retire as blue's act on each gap it reaches — the gap's `about_ref`, or a carrying Verify's `anchor`, in `retire.anchors` — read off those fields; so `lensTurnOn` gives the lens its turn and the exchange fold its movement. Without it, since #1659, the plan readies blue alone on the gap and the closing lens only by its other causes (III.1′.3 The lens's turn) | require `--answers G` on the retire (L1's rejected (c): the retire cannot be repeated); a turn arm in `debate` for met gaps (`EarlyLensDispatches`, the exchange fold and a remand's exchange would each need their own copy) |
| That lens (lead, disclosed; serving the joint ruling's backing arm) | Whose backing re-read meets a reading gap's demand | the seat of the carrying Verify: the ledger replaces only a seat's own check, so arm 4 reads the latest `verify` of the anchor by that seat. A gap carrying two readers' readings is met by either's backing re-read, and the other's standing check still refuses the close | the gap's minting lens (misses a reading another lens carried into the gap, which that lens alone can replace) |
| Backing arm's kind (lead, disclosed; serving Answer of another kind) | Does a reading gap's backing re-read meet a `computation` demand? | no: arm 4 is asked of a `source` gap only. A reading checks a source, and only the demanded kind discharges a demand. A reading gap minted `--check-kind computation` is met by a live computation answering it or by retiring the read citation | ask arm 4 under every demand (a source re-read would discharge a demand for a computation) |
| Failed predicate (lead, disclosed; serving the joint ruling's assembly screen) | Where "this check fails" lives | one column, `EvidenceTables.failed` (`found_against`, `unsound`), and `record.FailingAnchors`, read by `assembly-screen`; the ledger keeps both states | one `stateFailed` (L4's shared state, which the joint ruling rejects); a per-site `if computation` in capture |
| Item removal (lead, disclosed; serving Closures are history) | What takes the closed gap's item off the lens's list | the existing reading `firstAfter(ends.through[seat], at)` (`lensOwesOf`'s), a mint naming the gap in `supersedes`, or a later closing act that is not a carry. No new event or seen-marker | a new acknowledgement act (the ruling: no new act); a permanent line (a nag); the sitting's first read (a sitting would lose it mid-audit) |
| Blue on a met gap (gblock, 2026-11-03) | An open material gap whose demand a retire met still routes to `debate`, which engages blue on it always and its lens once its turn has come (`dispatch.go:331–343`, #1659). Is blue dispatched on it? Measured: the fixture dispatched blue on `G-342d3cdc` (737) | **`debate` is unchanged.** The reaching retire gives the lens its turn (Fork A reaching retire is an act on the gap), so both are engaged; the lens sits first in an epoch and usually closes, leaving blue the found-closed row. Where the lens keeps its grade, blue's null turn is what completes the exchange and runs the clock to impasse and the bench. Costs blue a sitting per met gap the lens has not yet closed | (b) engage blue on the gap only after the lens has acted on it since the retire: saves blue's sitting in the common case and keeps the clock, for one more condition in `debate` and in `EarlyLensDispatches`' mirror of it; (c) never engage blue on it: no exchange counts, a lens that declines to close holds the run to CEILING with no impasse, unless a new bound is built |
| A lens that does not sit again (gblock, 2026-11-03) | The item for a lens with no further sitting (retired; or the run ends) goes unread. Does the retire ready it? | **No; the residue is accepted (R18).** A readied party refuses `PassPermitted` (`dispatch.go:566`), and the ruling says the item does not hold PASS. Both measured cases were read at the lens's next sitting. The board carries the column | Ready the lens once per cause (holds PASS one sitting); a chair item naming the unread gap (the chair mints nothing, and it is a second list of one fact) |
| Which answers (gblock, 2026-11-03) | Does "answers all retired" count anchors of any kind, or the demanded kind only? A `computation` gap closed on `K-a` with a `C-b` also naming it: `K-a` retired, `C-b` live | **Any kind, the ruling's words,** with the simplest predicate. No archive can measure it (`Cite.answers` is new). This plan is written to it | Demanded kind only: the item also shows in the case above, at the price of the `check_kind` join and a different rule for `document` gaps |
