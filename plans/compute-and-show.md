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
(`report.Literal`). A shown block leaves the report as every anchor does: `edit` the sentence,
then `retire` (`anchor/anchor.go:75`). No new lifecycle.

**Parts (R-split).** Each ships alone, with its own epoch and `migrate` translation.

| Part | Scope | Depends on | Epoch |
|---|---|---|---|
| 1a | `prove` → `compute`; directory + argv; two-value basis; per-act store; one re-run per bundle, judged per act; `K-` anchors | — | E1 |
| 1b | `--input` (raw cited bytes, read-only); package `archive`; operator `archive rerun` | 1a | E2 |
| 1c | `--show table\|image\|text` on `compute`; images exist only as shown blocks | 1a, #1569 (render-time allowlist) | E3 |
| 2 | `--show image\|text` on `cite`: a colour crop of a box drawn on a view the model processes unresized | 1b (`archive rerun` re-derives crops), 1c (placement, `Show`, `Literal`) | E4 |

E1–E4 are the next free epochs when each part lands (main is 43).

**Success criteria.**

| # | Part | Criterion | §V |
|---|---|---|---|
| S1 | 1a | `feov-record blue compute --dir D --quote S --reason R -- python3 main.py` records a computation: bundle = every regular file under `D`; command = the argv after `--`; toolchain = resolved path + sha256 of `python3`. `blue prove` is an unknown command | V1a.1 |
| S2 | 1a | Two runs in fresh copies outside the run tree, each copy's canonical root replaced by `<bundle>` in the output. Equal output, exit and `out/` → `REPRODUCIBLE`; different with no `--variance` → refused naming what moved; with `--variance` → `VARIANT` + drift; `--variance` over agreeing runs → refused | V1a.2, V1a.3 |
| S3 | 1a | A recorded act never changes after write. Every act's verdict compares the bundle's STANDING re-run — the latest `ran` reproduce, its result at `rerun/<its rerun_sha>/` — against the act's OWN output, exit and `out/`, whatever refusal, crash or concurrency preceded it; capture's audit and `archive rerun` execute fresh and neither read nor write the store | V1a.4, V1a.5 |
| S4 | 1a | A keyed retry restates every recorded flag or is refused; program, argv and inputs are in the sha | V1a.6 |
| S5 | 1a | Every archive migrates with zero refusals; III.1.6's assertions hold; `archive rerun` over every migrated archive gives each of the 159 stored scripts the outcome `archiverun` measured on the CI image and committed | V1a.7 |
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
// Cite [MODIFY, 2]: + optional Show show = 26; optional int32 crop_page = 27;   // absent for an image source
//   optional int32 crop_x = 28, crop_y = 29, crop_w = 30, crop_h = 31;   // the box, in VIEW pixels (--region)
//   optional int32 crop_view_w = 32, crop_view_h = 33;                  // the view's size, read from its PNG
//   optional string crop_sha256 = 34;   // the crop the seat was returned: sha256 of its DECODED pixels (Crop hash)
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

### Part 1b — `--input` and `archive rerun` (epoch E2)

| Change | Detail |
|---|---|
| `compute --input C-…` (repeatable) [MODIFY] | the citation's RAW cached bytes (`fetchcache.Read`) — never extracted text (R-input). Placed read-only at `sources/C-…` in each copy; id + sha256 enter `bundle_sha`. Refused: a citation not on the record; bytes the cache does not hold; a citation whose bytes the tool composed — `copy_substitution` `recovered` or `named_backend` with no `copy_url` (`record.proto:1237`, `:1239`): no document exists to re-fetch |
| `reproduce`, capture's audit [MODIFY] | `compute.Rerun` materialises `sources/` the same way |
| Package `archive` [NEW] (`internal/archive`) | the one writer and the one reader of the run-archive tarball: `archive.Write` (from `ArchiveRecord`, `capture.go:2045`; `capture` calls it) and `archive.Extract` (replacing `internal/record/recordtest/archive.go:34`'s loop and `archdump`'s `untar`, `main.go:149`). Every entry `0o644`, created with explicit permissions; identity holds no mode (III.1.3). Capture's transcript tarball (`capture.go:1748`) is another format and stays. It commits no `cache/` (R-commit) |
| `feov-record archive rerun <runDir>` [NEW] (N1) | input: an extracted archive brought to the current epoch by `feov-migrate --from <runDir> --to <freshDir>` (an older epoch is refused by `requireThisEpoch`, pointing to `feov-migrate`). Re-fetch each input from where its bytes came — `copy_url` when set, else `url` (`fetchcache.Resolve`, paced) — verify sha256, then `compute.Rerun` each bundle once and compare every act (III.1.4: never reads `rerun/`, writes nothing to the store). Writes no event. `--citations` verifies every citation the same way (N2); a tool-composed citation prints `refused` (no document). `--help` states: it runs every archived program with the operator's privileges, filesystem and network (R12) |
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
| Assembly [MODIFY] (`report/layout.go`, `report/computations.go`) | block after the anchor's paragraph, numbered per kind in document order, inserted after every escape and the weave (no escape reads a shown block): `TABLE` → markdown table, every cell `Literal`, then `*Table n.* Literal(R)`; `IMAGE` → `![Figure n](<placed>)` — alt is tool text only — then `*Figure n.* Literal(R)`, in report.md (run root: the link resolves in the run directory and in an `archdump` of its archive, which copies `computations/` beside the report set; the archive itself holds no report.md, `capture.go:2045–2160`); `<img src="data:image/…;base64,…">` in report.html (an `<img>` runs no SVG script); `TEXT` → verbatim, already held to `edit`'s refusals. `<placed>` = `computations/<sha>/acts/<K-id>/out/` + each path segment `url.PathEscape`d, joined by `/` (a nested `out/sub/f.png` resolves), from one function (`placedPath`): the program chose the name, and PathEscape leaves no space, `(`, `)`, `[`, `]`, `<`, `>`, `\` or line break in a segment |
| Composition rule [NEW in 1a, III.1.5] (`report/notes.go`) | `report.Literal(s)` backslash-escapes every ASCII punctuation character and folds each line break to a space, so the string renders as its own characters in report.md and report.html and opens no image, link, raw HTML, footnote, emphasis or table cell. It is the one composer of seat, program and source text in a shown block — 1c: table cells (replacing `cell`'s pipe-only escape for shown tables, `assemble.go:1525`), the caption; 2: the caption (source title, page when there is one), the placeholder, the `--show text` span. Sweep of composition sites a part adds: these; evidence.md (names per III.1.5; `docs.go:211`'s whole-document escape moves per entry, applied before the entry's figure line is appended, so no tool-composed image is under it); run.md's absent-crop line (tool text, C-id and box) |
| Code regions [MODIFY] (`layout.go:350`, `notes.go:32`) | serve markdown STRUCTURE only — the weave-label `[^` escape and the composed-image `!` escape; raw HTML is #1569's. `eachLine` marks a line fenced when goldmark's tree (`mdReader`) places it in an `ast.FencedCodeBlock` (fence lines included) or `ast.CodeBlock`, so its seven callers (§II) inherit the renderer's reading; code spans come from `ast.CodeSpan`; `reCodeSpan` is deleted. Each escape parses its unit as it ships — a whole document (`docs.go:129`, an evidence entry) or a section block (`layout.go:89`) with `mdReader`, a one-line title (`assemble.go:151`, `proofs.go:85`) with `lineReader`. Units join closed: assembly appends the closing fence to any unit whose parse ends inside a fenced block, so no fence state crosses a unit boundary and each unit's reading is the document's (`probe-out.txt`'s case) |
| Raw HTML and images [MODIFY] (#1569; `notes.go:43`) | Raw HTML is #1569's in both outputs: report.html in md.go; report.md is written (`docs.go:391`) through the same decision on the same whole-document parse — every `<` outside code, autolinks and the tool's shapes backslash-escaped (each `<`, not only a node's opener: `probe2-out.txt`). So 1c adds no raw-HTML escape at composition and no raw-HTML refusal at `edit`/`ingest`. Images, which #1569 keys on a destination a seat can type: `EscapeNotes` → `EscapeComposed`, which outside code also escapes the `!` of each `ast.Image`, at its call sites, so no composed seat text (mint, fix, inquiry, ruling, verify, evidence) opens one; `edit` and `ingest` refuse an `ast.Image` outside code in authored text — a new link in `composedRefusal`'s chain, named in `help/edit.md` and `help/ingest.md` (measured 0 outside code in 30 archives). 1c supplies the allowlist's image set — `placedPath` of each shown act; Part 2 adds `cache/crops/<crop_sha256>.png`. Every other shipped .md is written through the same decision (T13) |
| Seat-facing views [MODIFY] | `show report --assembled --json` carries each block in its section's `text`; bare `show report` (`report/seatread.go`) carries the anchor only, and `AssembledNote` (`layout.go:284`) says shown blocks are assembled content |
| `show evidence --anchor K-…` [MODIFY] | returns the act's shown output — the stored path for an image; `reproduce` help and `scripts/agentgen/src/red-lens-computation.md`: read the image, not the caption — a truncated axis is an unsound sentence; an SVG is read as its source, and the duty states what the source does not show (the rendered layout: overlap, clipping, font substitution) until #1573 gives the lens a renderer (T14) |
| Voice lens [MODIFY] `scripts/agentgen/src/red-lens-voice.md` | a shown text block is report text: read the report as assembled (R5). A finding's `--quote` must be report text, which a shown block is not: quote the anchor's sentence and name the block; the fix is `edit` the sentence, `retire` the `K-` anchor, compute again |
| evidence.md, help [MODIFY] | each entry adds its shown output; tables are CSV, never printed markdown; a diagram can be a mermaid fence in `text`. `compute --help` (`image`): an SVG is shown as a static image (no script, hover or zoom); set `svg.fonttype: none`, so labels stay text the lens can read in the source |
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
| `… --region x,y,w,h` — phase 2 [MODIFY] | every phase-1 flag plus `--quote`, `--title`, `--region` (the box in the view's pixels). Re-derives the view from its key and records `crop_view_w`/`_h` from it; a box outside it is refused. Crops from the full-resolution colour raster (a PDF at `PageRenderDPI`; an image's own pixels), the box scaled per axis by raster ÷ view and rounded; `crop_sha256` = sha256 over the crop's width, height and decoded 8-bit RGBA pixels, row-major, in III.1.3's length-prefixed framing — not the PNG bytes, which an encoder change moves; stages and renames it to `cache/crops/<crop_sha256>.png`, records it, then returns its path for the seat to read; a wrong crop is `retire` + cite again (the existing lifecycle). Never committed (R-commit); `archive rerun --citations` re-derives it and prints `differs` when its sha256 moved (1b table). `citeFields` writes `show`, `crop_page` and `crop_x…h`, and `citeAsserted` (`cite.go:198`) gains `--show` → `show`, `--page` → `crop_page`, `--region` → `crop_x`, `crop_y`, `crop_w`, `crop_h`, so a keyed retry with another page or box is a Conflict (W6); `crop_view_*` and `crop_sha256` derive from those |
| Colour | `renderPagePNG` (`pagerender.go:480`) takes a pixel size and the colour model in place of its DPI; its three callers — `RenderPages` :376, `RenderOnePage` :455, `readscanned.go:204` — pass the pixel size their DPI gives today and grayscale, and a view passes `ResizedSize` and colour, so OCR renders and `render-page` shas are unchanged. The comment `pagerender.go:298–308` is restated: OCR and `render-page` stay grayscale; views and crops are colour; the colour follows the consumer, so the record carries no colour field |
| Caption | `*Figure n.* Literal(<source title>, page N)`, or `Literal(<source title>)` for an image source — the citation's footnote text; `--reason` stays off the report (`cite.go:191–195`) |
| Absent crop | `cache/crops/<crop_sha256>.png` missing: report.md, report.html, `show report --assembled` and `devcmd/archdump` render "*Figure n.* <source title>[, page N]: image not held in this copy" (`Literal`, subject terms only — report-voice ruling); run.md's log names the absent crop (C-…, box) and `archive rerun`, which re-derives it |
| Evidence lens | `show evidence --anchor C-…` on a shown citation returns the crop and the view (derived into `cache/` on read; a derived crop whose sha256 is not `crop_sha256` is returned with both shas named); `verify` reads both — fidelity at the leaf, no seat program between |
| First test | `fetchcache` has no `image/*` branch (`grep 'image/' internal/fetchcache internal/cli/blue`: imports only); whether `cite` accepts an image source is measured before the image-source arm is built |
| migrate | replay only: migrated citations are not shown |

| Carrier (sweep W3, III.9) | Becomes |
|---|---|
| `cli/seat/help/cite.md` | adds SHOWING WHAT A SOURCE SHOWS: `--show text` displays your span; `--show image` is two calls — first with only the source and page, and the tool returns a view sized so you see it unresized, with its width and height; read it; then the same call with your quote, title and `--region x,y,w,h` in that view's pixels records the citation and returns the crop — read it; a wrong crop is retired and cited again; a retry under one key restates the page and box |
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
`go -C scripts run ./check` (fail count read), `go -C scripts run ./agentgen -check`,
`go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen -check`, and
`FEOV_RELEASE_GATE=1 go -C plugins/frank-exchange-of-views/tools test ./releasegate/fuzz/` (record
and flag change). Re-armed by: any edit under `plugins/frank-exchange-of-views/tools`,
`scripts/agentgen/src`, debate.js, `.github/workflows/hooks.yml` (the `archive-rerun` job). Compute tests drive a fixture program the test builds (`.exe` on
Windows) and carry no Short gear, so they run on the Windows leg.

| # | Check | Fails if |
|---|---|---|
| V1a.1 | Rename commit alone: `protogen`, `schemagen`, `vocabdoc`, the shared gates; read every changed golden line and each `--help` before/after | a gate is red, or a help page lost a duty |
| V1a.2 | `internal/compute`: agreeing runs → `REPRODUCIBLE`; undeclared drift (output, exit, `out/` each) → refused; declared → `VARIANT` + drift; `--variance` over agreeing runs → refused; `../x` present at one depth → refused; a missing relative path printing lowercase `ENOENT`, exit 1 both runs → `REPRODUCIBLE`, exit 1 recorded; tool call → refused; symlink, > 1 MiB, `out`/`sources` entry, `__pycache__`/`.pyc` → refused; `./run.sh` (Unix) and `./prog.exe` (Windows) → run, toolchain absent; `./run.sh` on Windows → refused naming the start failure; under umask 077 and 022 the second act on the bundle finds `dir/` hashing to `bundle_sha`; an argv holding a newline and one splitting it differently → different shas; no copy remains, read-only `sources/` included | any arm passes the wrong way |
| V1a.3 | P-norm: a Python program printing `__file__` and `os.getcwd()` and raising, its copy under a symlinked temp dir → `REPRODUCIBLE`, recorded output holds `<bundle>` and neither copy's path; "No such file or directory" on stderr at exit 0 → refused. Delete the replacement, or the `EvalSymlinks` → the first fails | it is recorded otherwise |
| V1a.4 | Per act: a `VARIANT` bundle whose program prints and increments a counter file at an absolute path outside the bundle and both copies (in argv, so in identity); acts A (recorded 1) and B (recorded 3); counter reset to 3; `reproduce` A → `ran`, A `DIFFERS`, the counter now 4 (one execution); `reproduce` B → `ran` = false, B `REPRODUCED`, its `rerun_sha` A's; A's `acts/` byte-identical throughout; evidence.md shows A's own recorded output beside its re-run's; an act D recorded 2, compared against that standing re-run → `DIFFERS`, and evidence.md shows D's 2, never A's 1; a third act never compared reads `not_compared` in the evidence row and "not compared" in evidence.md, an unexecuted bundle `not_rerun`. An `out/` file changed with output equal → `DIFFERS`, drift names the file; exit changed → `DIFFERS`; absent recorded `out/` against an empty one → `REPRODUCED`; every live `DIFFERS` carries drift, every live reproduce `rerun_sha`, every `ran` `rerun_exit`; a `ran` = false whose `rerun_sha` is no longer the standing re-run's (a later `ran` appended between its read and its write) → refused. A `REPRODUCIBLE` act whose own comparison differs while the standing re-run's own verdict reproduced → `--as sound` refused. capture's audit over the same: one execution, A moved, B reproduced; mixed bases → `bundle_bases` and the audit detail. Delete `compare`'s act argument → B fails; read evidence.md's outputs from the standing event → B's display fails | B's verdict or display reads A's output, an act changes after write, or a no-comparison act reads as compared or unexecuted |
| V1a.5 | Store lifecycle (III.0, III.1.4): (a) a `REPRODUCIBLE` bundle whose re-run differs: `--as sound` refused → no `rerun/` entry, no `.tmp-*`; then `--as unsound` → one `ran` event and exactly one `rerun/<its rerun_sha>/` holding that execution's output and `out/`, `rerun_exit` on the event; (b) a test seam fails the append after the rename → an unnamed directory, no event; the next `reproduce` executes (`ran` = true), finds the directory present with the same bytes, and stands; (c) two `reproduce` processes released together on an unexecuted bundle → two events; a later `ran` = false act follows the LATER event's `rerun_sha` (give the two executions different outputs: tamper the earlier's directory → verdict unchanged; the later's → it moves); (d) capture's audit over a bundle without `rerun/` and with a tampered `rerun/<id>/output`, and `archive rerun` on an extracted archive holding a tampered `rerun/` → identical verdicts, the `computations/` listing hash unchanged; (e) two lanes writing one bundle → one `dir/`, both acts recorded; a pre-existing `acts/<K-id>/` → refused; a refused append leaves no `acts/<K-id>/`; `archive.Write` → `tar tzf` holds each recorded `rerun/<id>/` and no `.tmp-*` or unnamed directory; (f) migrate of a run whose replay renumbers a `ran` event → its `rerun/<rerun_sha>/` unchanged and still resolved. Resolve `rerun/` by listing (newest directory) → (c) fails; write in place without staging → (a) fails | any arm passes the wrong way |
| V1a.6 | Keyed retry: different `--variance` → refused; different argv / file bytes → Conflict; identical → idempotent, nothing written. `seat.AssertedCovers` over `compute` and `cite`: drop any one entry from `computeAsserted` or `citeAsserted` → it fails naming the flag. Record write: `VARIANT` without drift or invariant, `REPRODUCIBLE` with either, `show` = `SHOW_UNSPECIFIED`, `drift` on `REPRODUCED`, `DIFFERS` without `drift`, `rerun_exit` without `ran` or without `rerun_sha` → refused by the `(check)`; a migrated `ran` event with neither → accepted | any is recorded, or the migrated shape refused |
| V1a.7 | migrate tests on the archives of III.1.6, each assertion and residue row (m32's `not_run` rows included). Then `go -C plugins/frank-exchange-of-views/tools run ./devcmd/archiverun`: each of the 30 archives through `archive.Extract`, migrate and `archive rerun`, every act's outcome against `devcmd/archiverun/testdata/outcomes.tsv` (159 rows: archive, old sha, outcome). The file is MEASURED, not forecast: `archiverun -write` in the `archive-rerun` job on `workflow_dispatch` uploads it as an artifact and it is committed — interpreters pinned by `setup-python`/`setup-node`, their versions in its header. The dev-box probe's 109 / 48 / 2 had no pins and one copy; m19's `10ef5e65` reads `/home/…/scratch/universe-m19/…` by absolute path and is expected to differ on the CI image. Off the CI image the outcomes check fails by design (other interpreters, no pins): a dev box runs the migrate half only. The 13 migrated `DIFFERS` carry `drift` "on re-run the program wrote: …" with their program bytes unrespelled. Nightly in the same job on `schedule` (R12); never in `release` (a workflow edit: CI's `actionlint` reads it) | a refusal, a failed assertion, a header version mismatch, a missing interpreter, or any act's outcome other than its committed row |
| V1a.8 | Delete-the-row: each pinned env var, each `remap` admission (`P-` in `hexID`, `archivedToken`, `[^Pn]`, `not_run`, `about_kind proof`), each refusal arm, each `notFields` entry — a test fails for each | a deletion leaves the suite green |
| V1a.9 | Smoke: `scripts/universe.sh build`, grep the staged marketplace for `blue compute`, `--smoke` on a numeric question; capture's `computation-rerun` | the run recorded no computation, or a FAIL |
| V1b.1 | `--input`: raw bytes at `sources/C-…`, read-only, in both copies; a citation not on the record / not cached / `recovered` with no `copy_url` → refused; a different input under the same key → Conflict; a zero-input bundle hashes to the sha pinned in 1a's test | any |
| V1b.2 | `archive rerun` on a fixture archive with no `cache/` against a local server: correct bytes → `reproduced`, exit 0; a `followed_link` citation re-fetched from its `copy_url`, not its `url`; a `VARIANT` that moved → `moved`, exit 0; a `REPRODUCIBLE` that differs → exit 1; changed bytes → refused naming url and both shas, exit 1; 404 → exit 1; each act line names its inputs verified. `archive.Write` → `archive.Extract` round-trips a `./run.sh` bundle that then reproduces; C10 finds `tar.NewReader` only in `internal/archive`; `archive rerun --help` names operator privileges and network | any outcome, exit or line differs |
| V1b.3 | Real data: V1a.9 on a question with a CSV source; extract its archive to a clean scratch dir, `feov-migrate --from … --to …`, `feov-record archive rerun` | the summary's count of computations with an input re-fetched, verified and re-run is below 1; or a non-zero exit |
| V1b.4 | Rights: a fetch summary of a TDM-reserved page prints the note stating computation is accepted and its output committed; each C4 comment carries the decision | the note or a carrier says otherwise |
| V1c.1 | Assembly goldens for each `Show`: report.html carries `data:image/` and no external `src`; delete each `Show` case → a test fails | a block missing, misnumbered, or external |
| V1c.2 | Images: `edit` and `ingest` adding `![x](y)` outside code → refused, naming the refusal `help/edit.md`/`help/ingest.md` teach; inside a code span or fence → accepted. `![x](y)` in a mint, fix, inquiry, ruling and verify text → escaped in report.md, text in report.html, also when `y` is a PLACED store path. Raw HTML (#1569's decision, carried into report.md): #1569's four probes through real assembly, read in report.md as goldmark parses it — no raw `<script>`, `<iframe>` or external `<img>`, every later section present; `<img src=…>` in authored and composed text → text in both outputs; quadratic-formula's `verify.text` row 683 migrates and renders as text; evidence.md's `<a id>` anchors survive and every `evidence.md#K-…` link lands, while a seat-typed `<a id="x"></a>` is text. Delete the `!` escape, the image refusal or report.md's `<` pass → its forged case fails; all 30 archives still migrate with zero refusals | accepted, embedded, raw, or a migration refusal |
| V1c.3 | `show report --assembled --json` holds the table after the anchor's paragraph; bare `show report` holds none; the `red-lens-voice` agent golden holds the assembled-read and quote-the-sentence duties | either differs |
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
