# A corpus of pages that beat the reader

Status: shape ruled by gblock 2026-09-18 — keep the COMPLEX one-shots, one folder per page, a
README in each saying where it came from, pages committed and served gzipped from loopback during
the test; every failing image is a fixture, and the net widens each round. Revision 2, after a FAIL
audit: the defects it found are fixed in D2, D4, D5, D6, D7 and D10, and the value fork is answered
in §I. Carries part of #1004 — §I says which part.

## I. Summary & Goals

**The defect.** Every threshold in `internal/tessocr` was fitted to IEEE 1012, and #934 measured the
second document: on NBS SP 602 the detector found 2 of 60 pages — one of them the cover, falsely —
and missed all 12 dashed-rule tables. Those pages exist today only in `~/ocr-runs/934/`, a scratch
directory on one container. **Nothing in the repository holds a page the reader gets wrong**, so no
gate notices when a fix breaks one, and the next round starts by re-finding the same scans.

Generated goldens (#932) prove the rules we wrote. They cannot produce the thing that beat us: a
1981 typewriter rule of 20 px dashes, a cover whose photograph clears every grid threshold by 375×.

**Goals.**

1. A page that beat the reader is committed as a fixture, with its provenance, and read on every
   run of the tagged suite.
2. A page's licence is a **field drawn from a closed set**, refused at the WRITE, before bytes enter
   the tree.
3. The reading of every corpus page is pinned byte for byte, INCLUDING the pages we still read
   wrong — a fix shows up as a golden diff that is read line by line and accepted.
4. Each page records what a CORRECT reader would do (D6's `expect` block), the harness checks the
   reading against it, and the meter is **derived from that check** — never from a hand-set status.

**Quantitative targets, and what was actually landed.** The target was 10 pages over 5 documents in
≤ 90 s. Landed: **8 pages over 3 documents** (NBS SP 602, a Forest Service bird checklist, a DOE
report), 11 defect classes, and the tagged harness measured at **95–190 s** for the 8 — 12–24 s a
page, not the 4–8 s §II estimated from a single-page read, and it varies with box load. Two
candidates were refused rather than forced: a NARA census frame (5.7 MB, over fetch's own cap) and
an Internet Archive journal whose content turned out to BE its text layer. Both are in
`references.json`. The meter's opening number is **7 of 8 pages do not yet read as they should**.

**What of #1004 this discharges, and what it does not.** #1004 asked for a pinned corpus harness
that FETCHES by hash at test time, and for breadth across UNLV/ISRI, PubTables-1M, DocLayNet and
tesseract's own test images. gblock's ruling of 2026-09-18 **supersedes the fetch-at-test-time
half**: pages live in the repo and are served from loopback. The breadth half is NOT discharged
here — this plan lands 8 pages from 3 documents plus the admission rule that grows them. #1004
stays open for breadth, and `testdata/corpus/references.json` (D8) carries the leads already found
with the reason each is or is not committable.

**What this enables next, and is not part of this plan.** `Grid300`'s constants are raw pixels
(SEL 151, MinHPix 15000, MinVPix 4500) tuned at one resolution; `ReadRenderedPages` REFUSES a render
at any other DPI rather than misapplying them, so the system is single-resolution by refusal, not by
accident — and that refusal also blocks #934's 400/600 DPI experiment. Whether the constants can be
DPI-derived (SEL proportional to DPI, the pixel minima to DPI squared) is answerable only by running
a labeled set at two resolutions and comparing verdicts, which is what this corpus is the instrument
for.

**Non-goals.** A broad OCR benchmark. Character-accuracy scoring against transcriptions. Fetching
anything at test time. Pages that only repeat a kind the corpus already holds (D9).

## II. Technical Context

Measured 2026-09-18 while sizing this:

- A single page pulled out of a scan with PDFium (`FPDF_ImportPages`), TEXT objects removed, is
  **11–60 KB** and carries the publisher's own image bytes.
- That extracted page renders **byte-identically** to the same page inside the full document — SP
  602 p44 gives render sha `1e6a4c31…` either way. A corpus page therefore measures the pixels the
  whole-document run measured.
- **gzip buys 1–9%** on these files (60451 → 59815 on the cover; 11312 → 10310 on the blank page):
  the page images are JPEG 2000 streams, already compressed. They are stored gzipped by gblock's
  ruling, for one convention with `run-archive/` — this line exists so nobody later reads the `.gz`
  as a saving.
- A page read costs 12–24 s of tagged-suite time (8 pages ran in 95 s once and 190 s twice; the
  first estimate of 4–8 s came from a single-page read and was wrong).

Facts the audit established, which the design now rests on rather than assuming:

- **Nothing in the module refuses an outside host.** `internal/fetchcache/httpfetcher.go` refuses
  non-http(s) schemes only; every offline test injects a stub fetcher. D5 must make the isolation
  structural, not conventional.
- **A main package may live outside `tools/cmd/`** — `internal/tessocr/testdata/gen/main.go` and
  `third_party/pins/staticproof/main.go` already do, and `sc-doctor` reads only `tools/cmd`. The
  curation step does not have to be a test (D7).
- `internal/tessocr` already declares `var update = flag.Bool("update", …)` in
  `goldens_cgo_test.go`; a second declaration does not compile.
- `scripts/validatejson` sweeps every tracked `.json`, so a malformed record is already loud; only
  semantic refusal is ours to build.
- Three comments in this very directory state the policy this plan reverses — "nothing licensed is
  checked in" (`goldens_cgo_test.go:17`, `testdata/gen/main.go:7`, `testdata/gen/pages.go:9`).
  `internal/seatprobe/testdata/gen/main.go:5` says the same about ITS OWN generated fixture and is
  NOT a carrier of this change.

## III. Proposed Changes

**D1. A corpus page is a one-page, pixels-only PDF, gzipped.** Not a PNG: the PDF is 4–10× smaller
and exercises the render path the reader uses. The text layer is removed so the reader cannot
extract its way around the pixels, and the count removed is recorded.

**D2. One directory per page, and the set is enumerated ONCE with nothing unaccounted.**

```
internal/tessocr/testdata/corpus/<slug>/
  page.pdf.gz       the pixels, one page, gzipped — decompressed in memory when served
  provenance.json   the fields, including expect (D6)
  README.md         GENERATED from provenance.json
  reading.golden    what the reader produces today, byte for byte
```

The enumeration is the directory listing, read once by `corpusCases(t)` and shared by every gate.
Three refusals make a miss loud, because a glob that matches nothing is otherwise indistinguishable
from a clean board:

- **an empty set fails** — the neighbouring `TestTextCellGoldens` already states this property for
  itself (`if cases == 0 { t.Fatal(…) }`) and this harness takes it;
- **every directory must hold exactly the four names**, and a missing one fails naming the slug;
- **every file under `corpus/` must be accounted for** — an orphan (a `page.pdf.gz` with no record,
  a stray crop) fails rather than being skipped.

`provenance.json` fields: `slug`, `source_url`, `source_page`, `publisher`, `date`, `rights`,
`rights_evidence`, `local_file_sha256`, `page_sha256`, `text_objects_removed`, `added_at`,
`defect_class[]`, `why`, `expect` (D6).

`page_sha256` is the sha of the DECOMPRESSED page — the bytes the reader is handed and the golden
was pinned against; the `.gz` is a container, so the field names the page.

`local_file_sha256` is the sha of the file the operator downloaded, and its name says exactly that.
Revision 1 claimed the tool "verifies the sha it computes", which verified nothing: this plan cannot
fetch `source_url` (a non-goal), so nothing here can attest the bytes came from that URL. The field
records what was hashed; the attestation is the operator's.

**D3. The README is generated and gated, never hand-kept.** `TestCorpusREADMEs` regenerates each one
from `provenance.json` and fails on any difference — which is also how a new page gets its README.

**D4. Rights are a closed set, refused at the write.** `rights` must be one of
`us-government-work`, `public-domain-expired`, `cc0-1.0`, `cdla-permissive-1.0`,
`cdla-permissive-2.0`, `apache-2.0`; anything else, including prose like "probably public domain",
is refused. `rights_evidence` must be a URL or a quoted licence line. **The curation command (D7)
refuses before it writes a byte**, so an unrightsed page never enters a commit; `TestCorpusRecords`
re-checks in CI, which catches a directory placed by hand. Residue, stated: a hand-placed directory
is caught only at CI, after the bytes exist in a local commit — CI refuses the merge, but the
commit happened.

**D5. The harness is structurally offline.** `TestCorpusGoldens` (tagged `tessocr`) decompresses
`page.pdf.gz`, checks it against `page_sha256`, serves those BYTES from an `httptest` server, and
hands the reader that server's URL. It never reads `source_url` — provenance is for humans, not for
the fetcher — and the test asserts the URL handed to the reader is the loopback one. The reading is
pinned into `reading.golden` with the receipt's structural fields (table, grid_intersections,
reconstruction, text-cell stats). The golden carries **no run path, no temp dir, no timestamp and no
`host:port`** (httptest picks a random port every run); the run is a `t.TempDir()`. Running the test
twice and diffing is the gate on that.

**D6. What a correct reader would do is a field, and the meter is derived from it.** Each record
carries an `expect` block — structural claims (`table`, `tables`, `rows`, `columns`) and a short
list of `must_contain` strings transcribed FROM THE PAGE IMAGE by a human. The harness checks the
reading against `expect` and reports, per page, `meets` or `fails` with the claim that failed.

This replaces revision 1's hand-set `status`, which the audit correctly called an unrefusable fact:
a fix would land, the golden would be regenerated, the hand-set status would stay `known-defect`,
and the meter would keep counting a page that now passes — the number nothing can refuse, which is
the very defect §I opens by naming. With `expect`, a fix flips the meter by itself, and a regression
on a page that used to meet its `expect` fails the suite instead of quietly changing a count.

**D7. Curation is a command, not a test.** `internal/tessocr/testdata/corpusadd/main.go`, run as
`go run ./internal/tessocr/testdata/corpusadd -file … -url … -page 44 -slug … -rights … -why …`,
following the `testdata/gen` precedent already in this module. Every input is a flag with a loud
refusal naming the missing or malformed field — revision 1's env-var test would have read a
misspelled variable as a SKIP, and a skip is green. It takes a file the operator has already
downloaded, extracts the page, strips text objects, writes `page.pdf.gz` and `provenance.json` with
an empty `expect` for the operator to fill, and applies D4's refusal before writing anything.

**D8. The leads that cannot be committed are a record too.** `testdata/corpus/references.json` holds
what was found and rejected — `url`, `why_wanted`, `rights`, `blocked_reason` — and `REFERENCES.md`
is generated from it under the same staleness gate. A slug may not appear in both `references.json`
and `corpus/`; the harness refuses that.

**D9. A page enters only if it is a specimen.** It broke something, or it is the only example of a
kind the corpus lacks; `defect_class` names what it is a specimen OF, and a page whose classes are
all covered needs a reason in review. gblock's rule, kept in `testdata/corpus/README.md`: keep the
complex one-shots, not the simplistic ones.

**D10. The old policy statement is swept.** The three comments saying "nothing licensed is checked
in" state the new rule: pages are generated OR admitted under D4's closed licence set. No gate reads
a comment and `archaeology` does not cover Go comments, so this is a manual sweep — listed in the
components so review checks it.

**The first entries.** Five from documents already measured, five from the licence-checked hunt of
2026-09-18:

| slug | source | defect class |
|---|---|---|
| `nbs602-dashed-matrix` | NBS SP 602 p44 (US Gov) | rules drawn as 20 px dashes — invisible to a 151 px opening at any DPI |
| `nbs602-cover-photo` | NBS SP 602 p1 | a photograph reads as a table: 1256009 intersections, 375× the tuning document's maximum |
| `nbs602-contents` | NBS SP 602 p5 | a contents page with no leaders: every page number lost |
| `nbs602-form` | NBS SP 602 p56 | the CONTROL — solid rules, ticked boxes, reconstructed correctly, gate refused it at ratio 84 |
| `nbs602-blank` | NBS SP 602 p57 | a blank page: `length: 0`, indistinguishable from a collapsed read |
| `usfs-birds-markgrid` | Forest Service bird checklist (Wikimedia Commons, `Copyrighted: False`) | **a tick MARK GRID that is not IEEE 1012**, headers set rotated — the gap #934 could not fill |
| `doe-dotmatrix-listing` | DOE report, OSTI p9 (US Gov) | dot-matrix impact print, broken strokes, photo-reduced |
| `commons-faint-typescript` | Commons manuscript scan p14 | faint, speckled, uneven density — no rules, stresses recognition |
| `nara-census-handwriting` | NARA 1950 census frame (AWS Open Data: "US Government work") | handwriting in ruled form fields, microfilm |
| `unlv-mag-multicolumn` | `tesseract-ocr/test` UNLV page (Apache-2.0 repo) | multi-column magazine page with halftones. **Caveat recorded in the record:** the grant is the tesseract project's; UNLV states none |

IEEE 1012's mark-grid pages stay OUT — a copyrighted standard — and go in `references.json` naming
the run that holds them. So do FUNSD (non-commercial, no image clearance), RVL-CDIP (no licence of
its own) and the UNLV SourceForge sets (no licence stated). Still unfilled after the hunt:
**a landscape/rotated table** and **a skewed or warped scan**, both recorded in `references.json`
with the leads that were tried.

**Components.**

- `[NEW] internal/tessocr/testdata/corpus/` — pages, records, generated READMEs, goldens,
  `README.md` (the admission rule), `references.json` + `REFERENCES.md`, `STATUS.md`.
- `[NEW] internal/tessocr/corpus_test.go` (default build) — `TestCorpusRecords` (D2's three
  refusals, D4's closed set, D8's disjointness), `TestCorpusREADMEs`, `TestCorpusStatus`.
- `[NEW] internal/tessocr/corpus_cgo_test.go` (`tessocr` tag) — `TestCorpusGoldens` (D5, D6), reusing
  the package's existing `-update` flag rather than declaring a second.
- `[NEW] internal/tessocr/testdata/corpusadd/main.go` — D7.
- `scripts/check/gates.go` — NOT MODIFIED, and the plan was wrong to expect it: CI's tagged step
  (`hooks.yml:558`) already runs `go test -tags tessocr … ./internal/tessocr/ ./internal/fetchcache/`,
  which is where these goldens live, so the corpus is covered by the job that exists.
- `[MODIFY] .gitattributes` — `*.gz binary` and `*.pdf binary`; the repo's one committed PDF
  survives on Git's auto-detection alone today.
- `[MODIFY] internal/tessocr/goldens_cgo_test.go:17`, `testdata/gen/main.go:7`,
  `testdata/gen/pages.go:9` — D10's sweep. NOT `internal/seatprobe/testdata/gen/main.go`.
- `[MODIFY] plugins/frank-exchange-of-views/README.md` — one line on where corpus pages come from.

## IV. Risk & Mitigation

- **R1. A pinned golden makes a defect look correct (medium × high).** D6 is the mitigation that
  works: `expect` states what the page SHOULD produce, the harness checks it, and the meter counts
  failures found rather than a hand-set flag. The golden makes a CHANGE visible; the `expect` block
  says which direction is right.
- **R2. Licence error (low × high).** D4's closed set refused at the write, plus `references.json`
  for what may not be committed. Residue stated in D4.
- **R3. Suite cost (medium × low), measured higher than estimated.** 12–24 s per page, tagged suite
  only: 8 pages ran in 95 s once and 190 s twice on a loaded box. The ≤ 90 s target is already
  missed, and the corpus will grow. Past ~20 pages it takes its own gate rather than losing pages.
- **R4. The corpus becomes a junk drawer (medium × medium).** D9's admission rule and
  `defect_class`.
- **R5. A page's `expect` is wrong (medium × medium).** It is transcribed from the image by a human
  and can be mistaken, pinning a wrong target. Mitigation: `expect` holds only claims readable off
  the page — does it hold a table, how many rows and columns, five strings printed on it — and the
  record carries `why` so a reviewer can check it against the README. A wrong `expect` fails loudly
  the first time the reader gets that page right, which is when someone is looking.

## V. Verification Plan

Every command is runnable as written, from `plugins/frank-exchange-of-views/tools`.

1. `go build ./... && go vet ./... && go test -count=1 ./...` → the default-build corpus gates pass.
2. `eval "$(./third_party/pins/build-cstack.sh env linux-amd64 ~/ocr-runs/644-real/cstack)"` then
   `go test -tags tessocr -count=1 -ldflags '-linkmode external -extldflags "-static"' ./internal/tessocr/ ./internal/fetchcache/`
   → corpus goldens pass. Regenerate with the same command plus
   `./internal/tessocr/ -run TestCorpusGoldens -update` (flags AFTER the package).
3. Run step 2 twice and diff the goldens: byte-identical, proving no port, path or timestamp leaked
   into one.
4. Refusal drills, each restored after: `rights: "probably public domain"` → `TestCorpusRecords`
   fails; delete a `provenance.json` → the orphan check fails naming the slug; empty the corpus
   directory → the empty-set check fails; hand-edit a README → `TestCorpusREADMEs` fails; put one
   slug in both `references.json` and `corpus/` → the disjointness check fails.
5. `go run ./internal/tessocr/testdata/corpusadd` with `-rights` missing → a refusal naming the flag,
   non-zero exit, and NOTHING written (checked with `git status`).
6. Delete-the-row on the harness: the rights check, the orphan check, the empty-set check, the README
   comparison, the `expect` comparison and the meter count — one at a time, each must fail the suite.
7. Add one page end to end with D7 and confirm `page.pdf.gz`, `provenance.json`, README, golden and
   the STATUS line appear without hand editing.
8. `go -C <repo>/scripts run ./check`.
