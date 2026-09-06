# A blocked page is a hole that names itself, and a retry never pays twice

> **Superseded by plans/local-ocr.md (Wave 2, 2026-09-06).** The content filter left with
> the API: the local engine refuses nothing, so there are no blocked pages, and the
> blocked machinery (marker, blocked fields, `ocr_blocked_pages` in both summaries) is
> deleted. What survives from this plan is the receipt design — per-page provenance rows
> the reading record projects from, with render-sha staleness validation — whose resume
> purpose retired with the dollars but whose never-serve-stale-pixels guarantee stands.

Design for #679, from the #644 measurement. Status: under review — this plan is the
decision document; no code lands until it merges. First draft FAILED the plan-auditor
gate on four completeness gaps (the receipt/clear contradiction, the `ocr read` summary
carrier, an impossible §V oracle, unpasted censuses); this revision answers each.

## I. Summary & Goals

The API's output content filter deterministically refuses some pages of some documents:
17 of IEEE 1012's 80, page-stable across four attempts and two models, unpredictable from
page role (the title page with the full copyright notice passes; ordinary body pages 11
and 12 do not). Today one refused page makes the whole automatic read permanently fail —
fetch reports `ocr_reason` honestly and writes no reading — and every retry re-spends
every page before the refusal. The feature is unusable on exactly the corpus it was built
for.

Goals, in priority order:

1. A document with refused pages yields a reading whose holes are LOUD — stated in the
   assembled text where a citing seat reads, and carried as fields on the record where
   machinery reads. Never a silent skip: the fatal-error design existed to prevent a
   document with an invisible hole, and that invariant survives the change.
2. A page already read is never paid for again — not across retries of a blocked
   document, and not across a crash mid-read (the page-79 problem: today a crash at page
   79 re-spends 78 calls).
3. A failure that is NOT the content filter keeps today's semantics: fatal, nothing
   half-written that reads as whole.

Non-goals: retrying blocked pages automatically (deterministic refusals do not yield to
retries — measured 4/4 on page 2); working around the filter; changing what the filter
blocks.

## II. Design

Technical context, in one place: Go module `plugins/frank-exchange-of-views/tools`;
model calls via `anthropic-sdk-go` v1.68.0, whose content-filter refusal arrives as a 400
`invalid_request_error` with message text "Output blocked by content filtering policy"
and a request id, no structured code; per-document artifacts live under
`<run>/cache/<sha>.pages/` (page texts) and `<run>/cache/<sha>.ocr.txt` (assembled),
with `reading.json` inside the pages directory.

**Classifying the refusal.** One function, `blockedByPolicy(err) bool`: a typed
`*anthropic.Error`, status 400, message containing the filter sentence. That is a
string-shaped key and [[facts-are-fields]] is owed an answer: the failure direction is
CLOSED. A reworded message stops classifying, the error falls through to the fatal path,
and the read refuses loudly — a wrong-way drift costs a usable document, never a silent
hole. The fragility is stated on the function; its test asserts both directions.

**The hole, in both registers.** A blocked page's transcription becomes the marker
`[page N: output blocked by content filtering policy]` in the per-page text and the
assembled document — prose for the human and the citing seat. The record carries the same
fact in fields: `PageReading` gains `blocked bool` and `blocked_reason string` (the API
message verbatim, request id included). No count field on `ReadingRecord` — the count is
derivable from the pages slice, and a second copy is free to disagree with it.

**Both summaries state partiality.** fetch's summary keeps `ocr_derived: true` (a reading
exists and is usable) and gains `ocr_blocked_pages: N` when N > 0, with its own line in
the human render. `ocrReadSummary` — the deliberate verb's summary, including its
reuse-guard path that returns a stored blocked record — gains the same field and line: a
summary-only reader on EITHER path learns the reading has holes before opening the text.
Exit semantics, stated: both `fetch` and `ocr read` exit 0 with blocked pages — the
reading succeeded and says exactly what it could not contain; only classifier-missed
(fatal-path) errors are non-zero.

**Per-page receipts are what resume is made of.** Each page writes a receipt beside its
text as the loop runs — `<run>/cache/<sha>.pages/p%04d.receipt.json`: render sha, dpi,
model, read_at, text sha, tokens, and the blocked/blocked_reason pair. The assembled
`ReadingRecord` is built FROM the receipts, a projection of per-page facts rather than a
parallel account. On a re-run, each page is rasterised first (deterministic, cheap), and
a receipt matching that page's fresh render sha and the requested model is reused without
a model call — including a blocked receipt: a deterministic refusal is a fact already
paid for. A malformed receipt is an error, never an absence.

**The clears become per-page validation on the automatic path — the first draft's
contradiction, resolved.** Today `renderAndReadPages` clears `PagesDir` wholesale before
reading (readscanned.go), which would delete receipts before resume could fire. The
wholesale clear's job was staleness: never serve a 72-DPI reading for 200-DPI pixels.
Per-page receipt validation does that job strictly better — a receipt is reused only when
its render sha equals the sha of the page as rendered NOW, so a DPI change mismatches
every receipt and replaces every page, which is the old clear's outcome without the old
clear's amnesia. The document is content-addressed, so the page count cannot drift under
a sha. What IS still cleared on the automatic path: a page whose receipt mismatches gets
its text and receipt overwritten via the existing replace-not-atomic writer. The
deliberate `ocr pages` re-render keeps its wholesale clear unchanged — its product is
images at a stated resolution, and a re-render legitimately wipes readings, receipts
included; `ocr read --force` discards receipts and reads everything again, blocked pages
included (the operator's lever for "the filter may have changed").

**What does not change.** The page cap and its refusal wording; the (document, DPI) reuse
guard for whole stored readings; `RenderPages` and the `ocr pages` verb; the prompt.

## III. Scope — carriers and consumers, censuses run

Artifact locations after the change (receipts are [NEW], the rest [MODIFY]):

    <run>/cache/<sha>.pages/p0001.txt            per-page text (existing)
    <run>/cache/<sha>.pages/p0001.receipt.json   [NEW] per-page provenance row
    <run>/cache/<sha>.pages/reading.json         projection of receipts (existing name)
    <run>/cache/<sha>.ocr.txt                    assembled text with markers (existing)

- [MODIFY] `internal/fetchcache/pageread.go` — classifier [NEW func], receipt read/write
  [NEW funcs], `PageReading` fields, `ReadRenderedPages` loop.
- [MODIFY] `internal/fetchcache/readscanned.go` — `renderAndReadPages` gains the same
  per-page step (shared with pageread's loop, not copied); wholesale clear replaced by
  per-page validation as designed above.
- [MODIFY] `internal/cli/fetchsummary.go` — `ocr_blocked_pages` field + render line.
- [MODIFY] `internal/cli/ocr.go` — `ocrReadSummary` (type at ocr.go:173, render at :189,
  built at :282) gains the same field + line; long-help prose updated.
- [MODIFY] `internal/cli/fetch.go` — long-help prose ("a fetch failure is a non-zero
  error (pick another source)" and the refused-not-truncated sentence must speak the
  hole-and-resume model).
- Consumer census for the summary fields, run 2026-09-05 —
  `grep -rlnE "ocr_derived|ocr_reason|OCRDerived|OCRReason" --include=*.go internal/ releasegate/`
  (the first draft pasted this without `-E`, a command whose no-match reads as a clean
  board — the auditor's catch). Full results, one disposition each:
  - `internal/cli/fetch.go`, `internal/cli/fetchsummary.go`, `internal/cli/ocr.go` —
    [MODIFY], named above.
  - `internal/fetchcache/pageread.go` — [MODIFY], named above.
  - `internal/cli/ocr_test.go` — asserts `ocr_derived: true` in the `ocr read` render
    (line 228). UNCHANGED, and the plan commits to what makes that true: the
    `ocr_blocked_pages` line is emitted only when N > 0, so a hole-free reading renders
    exactly as today. A new test asserts the N > 0 render.
  - `internal/cli/fetchocr_test.go` — stub-scan-reader plumbing tests; gains the
    blocked-count assertions listed in the test census below, existing assertions
    unchanged.
  - `internal/fetchcache/extractor.go`, `internal/fetchcache/fetchcache.go`,
    `internal/fetchcache/pdfextract_test.go` — extraction-level `text_extracted` reasons,
    a different concept sharing the string family. NO CHANGE.
  - `internal/consistency` — zero references; `releasegate/fuzz` fetches with
    `--ocr=false` always (fuzz_test.go:1566) and never reads these fields.
- Consumer census for the record (`grep -rln 'ReadingRecord|ReadReadingRecord'` outside
  fetchcache): `internal/cli/fetchsummary.go`, `internal/cli/ocr.go`,
  `internal/cli/fetchocr_test.go` — all in scope above.
- Tests [NEW]: blocked mid-document via stub reader (holes + fields + both summaries'
  counts + exit 0); non-filter error still fatal; resume spends only unread pages
  (call-count asserted); blocked receipt reused with zero calls; malformed receipt
  refuses; DPI change replaces every receipt (the old clear's guarantee, restated as a
  test); `--force` re-reads blocked pages; classifier both directions.
- #679 closes on merge of the implementation; the #644 thread gets the outcome.

## IV. Risks, graded (likelihood × impact × complexity)

- Substring classification drifts on an API rewording — likely eventually × low impact
  (fails closed to today's fatal behavior) × trivial to re-point. Stated on the function.
- Receipt/record dualism confuses a future reader — medium × medium (a writer that
  updates one and not the other forks the account) × contained: the record is BUILT from
  receipts in one function, and nothing else writes it. The rejected alternative (one
  incremental record file) turns a mid-append crash into a malformed record, which blocks
  the honest-absence path receipts preserve.
- A reading with holes is quoted as whole — medium × medium × cheap: marker in the text,
  field on the record, count in both summaries. A consumer ignoring all three had no
  reading at all under today's semantics — 0 of 80 pages instead of 63 with 17 named
  holes.
- Blocked-receipt reuse ossifies a transient block — low (measured page-deterministic,
  4/4) × low (63/80 pages still served) × `--force` exists.

## V. Verification plan

Written before implementation; re-run at every stage:

1. `cd plugins/frank-exchange-of-views/tools && go build ./... && go vet ./... &&
   go test -count=1 ./internal/fetchcache/ ./internal/cli/` — re-arms on any change under
   those packages.
2. The §III test list, plus mutation-mindedness on the classifier: a `blockedByPolicy`
   substring flip is exactly the survivor shape the release gate now hunts, so its test
   kills both directions (classifies the real message; refuses a reworded one).
3. Live, once, with credentials (~$2 at opus rates, needs gblock's go-ahead at
   implementation time): re-fetch IEEE 1012 through the automatic path. Expected: exit 0;
   `ocr_blocked_pages: 17`; 17 markers in the assembled text; 63 per-page texts whose
   page-length profile matches the #644 corpus in `~/ocr-runs/assembled/` within
   transcription variance (byte-equality is NOT expected — a model re-read returns
   different bytes by the record's own doctrine; the earlier duplicate-read measurement
   showed word-identical text with whitespace drift, so a spot-diff of three pages is the
   check); and a second fetch of the same URL spending zero model calls (receipt reuse,
   asserted from the summary's token fields).
