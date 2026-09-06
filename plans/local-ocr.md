# The reader becomes deterministic: Tesseract + Leptonica, statically linked, fully local

Ruled by gblock 2026-09-06: drop the Claude reading path; keep tesseract + leptonica at
300 DPI; fully local — including tables. Status: under review; this plan is the decision
document, and implementation follows its merge in the waves of §III.

## I. Summary & Goals

The model reading path measured excellently (#644: near-verbatim prose, 329/330 table
cells) and is being removed anyway, for reasons the measurements also support: it takes
25–35 s and real money per page, requires a second credential system beside the seat's
subscription (the split that drove this decision), and the platform's content filter
deterministically refuses ~21% of the first real document's pages on every model and
billing path. The build spike (2026-09-06, ~/scratch-tessspike/RESULT.md) proved the
replacement end to end: a fully static tesseract 5.5.3 + leptonica 1.87.0 cgo binary
builds from hash-pinned tarballs in 4m20s, and at this plan's operative 300 DPI OCRs a
page in ~1.6 s at 0.80% prose WER, detects ruled tables at 96 ms/page, and recovers all
11 columns and all 11 rotated headers of the corpus's hardest table. (The spike's
200-DPI figures — 1.3 s, 0.40% WER, 32 ms, and the recall-1.0 confusion matrix — are the
road not taken and the tune to redo; §III Wave 0 and §IV price the difference.)

Goals, in priority order:

1. `fetch` reads a scanned PDF with no model, no credentials, no network, no filter —
   every page, including the 17 the API refuses forever — at seconds per document.
2. Tables survive: ruled-table pages are detected and reconstructed into the same
   |-separated rows the corpus already uses, validated against the trusted model
   transcripts we already own.
3. The reading becomes REPRODUCIBLE, and the record says so: a deterministic engine keyed
   as `tesseract@5.5.3+leptonica@1.87.0` restores the re-derivation property (#636's
   extractor model) that the model reader had to disclaim.
4. Downloaded release binaries are fully linked: CI builds the C stack itself from pinned
   sources; no system dependency on any seat.

Non-goals: any LLM fallback or escalation path (ruled out — fully local); OCR of
non-ruled/borderless tables beyond what plain text gives (named future work); languages
beyond English tessdata (the corpus is English standards; adding traineddata is config,
not architecture).

## II. Design

Technical context in one place. Engine: tesseract 5.5.3 C API + leptonica 1.87.0, built
static with `zig cc/c++ -target <triple>-musl` (linux) from tarballs pinned by sha256
(spike's PINS.txt, lifted into the repo); our own thin cgo shim — no third-party binding
— exposing init, OCR-to-text, OCR-to-TSV, and grid detection; `eng.traineddata`
(tessdata_fast, pinned) embedded via `go:embed` and loaded through a ~10-line C++
wrapper over `TessBaseAPI::Init(data, size, …)` since the C API lacks init-from-memory.
Render DPI: **300 globally** (ruled; the measured cost is stated in §IV). Existing
carriers: the fused loop in `readscanned.go`, the deliberate verbs in `cli/ocr.go`, the
summary surfaces, and the #757 receipt/holes machinery.

**The per-page pipeline.** Render page at 300 DPI (existing pdfium path) → grid
detection (leptonica morphology, 96 ms at the operative 300 DPI; thresholds are per-DPI constants from Wave 0's 300-DPI tune, seeded by the spike's
values) → if no grid: `GetUTF8Text`, done → if grid: TSV with word geometry, plus the
rotated-header recovery (detect tall-narrow high-confidence boxes and low-coverage
header bands; crop, `pixRotate90` clockwise, re-OCR the band) → deterministic
reconstruction: cluster x-centres into columns, y-centres into rows, place marks and
labels, emit |-separated rows. Reconstruction confidence is a FIELD on the page record
(columns found, marks placed/unplaced); a page whose reconstruction places fewer than a
threshold of its marks falls back to plain text for that region WITH the failure stated
on the record — a bad reconstruction must never read as a good one.

**What the record becomes.** `ReadingRecord`/`PageReading` keep their shape; `Model` is
replaced by `Engine` (`tesseract@5.5.3+leptonica@1.87.0`, the #636 identity key), token
fields go, per-page fields gain `table bool` and reconstruction stats. `ocr_derived`
stays true — the text is still machine-read off pixels, and that fact is what the field
carries — but the attestation language ("a re-read returns different bytes") is replaced
by the stronger claim it retires: this reading re-derives, and `reproduce` can check it.

**What is DELETED, named in full** ([[complete-the-concept]]; this is the subtraction
half of the concept, most of it merged only days ago):

- `AnthropicPageReader`, `readPrompt`, `PageReader`'s network implementation, the
  anthropic-sdk-go dependency (go.mod), credentials prose in errors and help.
- `blockedByPolicy`, `blockedMarker`, blocked fields, `ocr_blocked_pages` in both
  summaries, and plans/ocr-blocked-page-resilience.md's machinery (#679/#757): with no
  API there is no filter and no blocked page. The receipts survive ONLY as the
  per-page provenance rows the record projects from — their resume purpose (protecting
  dollars) retires with the dollars; their DPI-staleness validation stays because it
  replaced the wholesale clear.
- `MaxReadPages`: the cap was denominated in MODEL CALLS ("a page is a model call") and
  its denominator is gone. Replaced by a disk-denominated refusal on render (the 534-page
  CJK chart at 300 DPI is a disk question now, not a billing one).
- `--model` on `ocr read`, `defaultReadModel`, model/dpi/token summary fields' model
  parts. `--force` keeps its meaning (discard receipts, re-derive) and becomes cheap.

**Release builds.** The release job builds the C stack from pinned tarballs before
cross-compiling (4m20s measured for linux; the cache-less-on-tags rule stands and the
cost is acceptable at release cadence). Targets: linux amd64/arm64 via zig musl —
proven; windows amd64/arm64 via zig/mingw — high confidence, verified in Wave 0;
**darwin amd64/arm64 is the open build question**: Wave 0 spikes zig's darwin C++
cross-target, and if it fails the fallback is two native macos runner legs with
SHA256SUMS aggregated across jobs — the one-runner shape yields before the fully-local
ruling does. The feov-record race leg's `deps:` expansion adapts automatically (the
graph is derived, not listed), but race-with-cgo now needs the C toolchain present in
the race leg — Wave 1 wires zig there too.

## III. Scope — waves, carriers, censuses

Wave 0 (spike extensions, no repo changes): windows and darwin cross-build proofs;
reconstruction prototype validated against the oracle below; **the 300-DPI detector
tune** — the measured confusion matrix (precision 0.909, recall 1.0) is a 200-DPI
result, and the spike's 300-DPI spot-check flipped a borderline page (p0025) with its
own note saying re-tune before reuse — so Wave 0 re-tunes thresholds at 300 and re-runs
the full 80-page matrix; those constants are what Wave 1 ships and §V.5 pins.

Wave 1 (engine in-tree): `internal/tessocr` package [NEW] (cgo shim, embed, detector,
TSV, reconstruction), `third_party/pins/` [NEW] (tarball URLs + sha256s + build script
lifted from the spike), CI [MODIFY hooks.yml] with two decisions made here rather than
discovered: **(a) the cross-GOOS vet matrix survives cgo** — cross-vet runs with cgo
implicitly disabled (no per-target CC), and `internal/tessocr` carries a `!cgo` stub
file so the package type-checks under CGO_ENABLED=0; the matrix's header comment gains
that stated exception; **(b) PR jobs build the C stack behind an actions/cache keyed on
the sha256 of PINS.txt** (~4m20s cold, seconds warm; the engine compiles and tests on
every PR, never build-tagged out of them), while the RELEASE job builds from pinned
source cache-free — the release-reads-no-cache poisoning invariant stands untouched and
is restated beside the new cache. Race leg toolchain [MODIFY: zig present for cgo
-race].

Wave 2 (the swap + subtraction): `internal/fetchcache/{pageread,readscanned}.go`
[MODIFY: engine swap, deletions above], `internal/cli/{fetch,ocr,fetchsummary}.go`
[MODIFY: flags, fields, help prose], tests [MODIFY: stub reader becomes a fake engine;
blockedpage tests deleted with their feature; new reconstruction-oracle tests],
`plans/ocr-blocked-page-resilience.md` [MODIFY: superseded-by note], README/agent
surfaces (census below).

Carrier censuses, run 2026-09-06:

- Model-path code census
  (`grep -rlnE "AnthropicPageReader|readPrompt|blockedByPolicy|MaxReadPages|defaultReadModel|ant auth|ANTHROPIC" --include=*.go internal/ releasegate/`):
  `internal/cli/fetch.go`, `internal/cli/ocr.go`, `internal/cli/fetchocr_test.go`,
  `internal/fetchcache/{pageread,readscanned}.go`,
  `internal/fetchcache/{pageread,readscanned,blockedpage}_test.go` — all Wave 2
  [MODIFY/DELETE]. `releasegate/fuzz`: zero hits (fetches `--ocr=false` always).
- Summary-field consumers: the nine-file census from plans/ocr-blocked-page-resilience.md
  §III still holds (re-run, identical); the extraction-level trio again NO CHANGE.
- Agent-facing prose, run 2026-09-06 —
  `grep -rilE "ocr" README.md plugins/frank-exchange-of-views/README.md plugins/frank-exchange-of-views/{skills,agents,commands,tests}/ docs/`
  (the first draft pasted `grep "ocr|OCR"` without `-E`, a command whose zero cannot be
  trusted — the same class the previous plan's gate caught; corrected and widened): two
  hits. `docs/git-hooks.md` is the substring inside `autocrlf` — no concept, NO CHANGE.
  `docs/setup-script.md:66` states PDF/OCR tooling is deliberately NOT a system
  dependency — a claim static linking preserves; its rationale paragraph is checked and
  updated in Wave 2 [MODIFY]. Nothing else agent-facing speaks the model path.
- Version surface: this reshapes the record (`Model`→`Engine`), removes CLI flags, and
  changes the release artifact contract — it ships at the next feov RELEASE BOUNDARY,
  and the version bump + tag is the human's release act (CLAUDE.md's rule), named here
  so the thread is tracked rather than remembered.
- go.mod: anthropic-sdk-go v1.68.0 becomes removable
  (`grep -rl anthropic-sdk-go --include=*.go` → exactly `internal/fetchcache/pageread.go`
  and `internal/fetchcache/blockedpage_test.go`, both Wave 2 files).

## IV. Risks, graded (likelihood × impact × complexity)

- **Reconstruction quality on unruled/odd tables** — high × medium × the reason for the
  confidence FIELD and stated fallback: a reconstruction that cannot place its marks
  says so on the record instead of emitting a plausible grid. The oracle (§V) bounds the
  known corpus; unknown corpora get honesty rather than accuracy.
- **Prose WER doubles at 300 DPI** (0.40%→0.80% measured) — certain × low × zero: ruled
  cost, stated here so it is a decision and not a discovery. Adaptive per-page DPI is
  the rejected-for-now alternative (render twice, OCR prose at 200); it re-enters only
  with measurement showing the 0.4 points matter downstream.
- **Darwin cross-build fails** — medium × medium (release shape changes) × contained by
  the Wave 0 spike + native-runner fallback, decided before Wave 1 lands.
- **Quality regression vs the model on degraded scans** — the operative 0.80% WER is one clean-ish
  1998 scan; worse scans will fare worse with no escalation valve. Accepted by the
  fully-local ruling; the reproduce property and confidence fields are the mitigation:
  failures are visible, re-derivable, and arguable.
- **cgo spreads into the dev loop** — every `go test ./internal/fetchcache/` now needs
  the C objects. Mitigated: the engine lives in its own package behind an interface;
  fetchcache tests keep a pure-Go fake engine; only `internal/tessocr`'s own tests and
  release builds need the toolchain (build-tagged, documented in sc-doctor).

## V. Verification plan

1. Standing loop, re-armed on any change under the named packages:
   `go build ./... && go vet ./... && go test -count=1 ./internal/tessocr/ ./internal/fetchcache/ ./internal/cli/`.
2. **The reconstruction oracle**: rebuild p0054's Table 3 from 300-DPI TSV and compare
   against the trusted model transcript (`~/ocr-runs/assembled/p0054.txt`, measured
   329/330 against pixels) — cell-level agreement reported as n/330, target ≥300 with
   every miss listed; plus p0051/p0052 (the two heaviest ruled pages) at
   headers-and-shape level. Runs as a normal test against checked-in TSV fixtures, so
   CI needs no OCR run to enforce it.
3. Static proof per release target: `file`/`ldd` assertions in the build script (linux/
   windows: fully static; darwin: no non-libSystem dylibs), plus a smoke OCR of a
   checked-in fixture page by every built binary that can execute on the build host.
4. End-to-end, local and free: fetch IEEE 1012 through the new path — expect all 80
   pages read (17 formerly-blocked included), wall time under ~3 minutes, table pages
   flagged with reconstruction stats, prose WER spot-checked against the model corpus.
5. Mutation-mindedness: the detector thresholds and reconstruction clustering are
   operator-flip-rich; their tests assert both directions on the measured boundary
   pages OF THE 300-DPI TUNE from Wave 0 (the 200-DPI boundaries — p0028/p0040/p0044,
   plus the p0025 flip — are the candidates; the pinned set is whatever the 300-DPI
   matrix names).
