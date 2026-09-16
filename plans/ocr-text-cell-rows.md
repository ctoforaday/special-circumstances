# A ruled table of text cells keeps its rows

Status: draft for gblock, 2026-09-16. Carries #932 forward from #644's measurement.

## I. Summary & Goals

**The problem.** On a ruled table whose cells hold TEXT rather than tick marks, the grid
detector fires, mark reconstruction finds no marks, and the page falls back to tesseract's plain
reading. That reading serialises column by column within a row band, so the words survive and the
binding of cell to row does not. Measured on IEEE 1012: **34 of 35 grid pages fall back today**,
and only p54 (a mark grid) keeps a reconstruction. Two consequences the record cannot see:

- p34 (Table 1) emits three task cells, then all three tasks' inputs, then all their outputs. "What
  are the required inputs of task (2)?" has no answer in the reading.
- p57 (Annex A) emits row 5.5.5's cells after the next row's heading, so read in order the page
  maps ISO 6.4.1 to Maintenance V&V — which the page does not say.

**The measured basis for this plan** (spike on this branch, 2026-09-16, real engine at 300 DPI):
the rule geometry recovers the lattice. p34: 6 horizontal and 6 vertical rules → 5 bands × 3
columns, 533 of 556 words placed inside cells, the two section headings correctly one cell wide.
p57: 11 bands × 3 columns, and 5.5.5 reads with its own mapping. §V.1 records the whole-document
figures.

**Goals.**

1. A ruled table of text cells reconstructs as `|`-separated rows, cell by cell, row order kept.
2. The lattice is a SECOND reconstructor, tried only where mark reconstruction does not stand.
   p54 keeps the mark path it passes today, byte for byte.
3. What was measured is on the record as fields: cells, columns, words placed, words outside.
4. A page whose lattice does not explain its words still falls back to plain text, with the
   consequence stated as today.
5. No model, no network: leptonica's own openings plus the TSV the page already produces.

**Success criteria (measured, §V).**

- The 30 text-cell pages reconstruct; p25 and the four mark grids fall back, stated.
- p34 and p57 reconstruct, and each row's cells match the oracle's row in
  `~/ocr-runs/assembled/` (cell-level, whitespace-normalised, on a hand-checked sample).
- p54's text is unchanged from main's reading (its `text_sha` moves for no page that keeps the
  mark path).
- Every refusal and threshold added here fails a test when deleted (§V.4).
- `go test ./...`, the tagged engine suite and the goldens pass.

**Non-goals.** Tables with no ruling (whitespace-aligned) — the detector never fires on them.
Mark grids, which keep their own reconstructor. Merging a cell's text across a page break.
Reading a rotated text table (the rotation probe stays as it is).

## II. Technical Context

`internal/tessocr` is the engine (cgo, `tessocr` tag). `DetectGrid` counts rule pixels through two
leptonica openings; the counts decide table-or-not (`Grid300`) and say nothing about where the
rules are. `readGridPage` runs the TSV passes, the rotated-header band, `Reconstruct` (marks), and
`fallbackReason`, which is the whole acceptance decision and is a pure function of the
measurements. `PageResult.Reconstruction` carries `Stats`; `PageReading` on the reading record
carries `table`, `grid_intersections`, `reconstruction` and `reconstruction_fallback`.

**What this change adds, already spiked on the branch:** `tessocr_grid_lines` returns one line per
rule (`h|v x y w h`) from the connected components of the SAME two openings — so the geometry
explains the counts rather than being a second measurement of the page. `GridLines` parses it into
a `Lattice`; `Lattice.Cells` turns it into cell rectangles: bands between horizontal rules, and per
band the vertical rules that actually cross it, which is what makes a spanning heading one cell
rather than three empty ones.

## III. Proposed Changes

**D1. The lattice reconstructs, not the plain reader — and only where marks do not.**
`readGridPage` keeps its order: marks first, `fallbackReason` as today. Where that says fall back,
the page tries the lattice before plain text. p54 is untouched by construction.
*Beaten:* running the lattice first (it would regress p54's marks, measured: the lattice places X
marks into neighbouring columns), and replacing `Reconstruct` outright (two different inputs — one
infers its grid from marks, the other reads it off the rules).

**D2. A cell's text is the TSV words whose box centre falls inside it; the rest of the page is
kept around the table.** A page is not always all table: p15 is prose with a 5-row table in it (65
of its 519 words sit inside the lattice). The reading is the prose above the table's bounding box,
in reading order, then the table's rows, then the prose below. Only a word inside the box and
inside no cell is dropped, and that count is a field.
*Beaten:* dropping every word outside the lattice (it would delete 454 of p15's words — the page's
actual content), clipping by box overlap (a word touching a rule lands in two cells), and
assigning by nearest cell (the running head would be pulled into the first row).

**D3. A cell's words are joined in reading order, and a line break inside a cell is a space.**
The oracle renders `<br>` for the same break; a plain-text reading cannot carry markup, so the
reading joins with a space and the record says the cell's line count.

**D4. Acceptance is three measured conditions: two columns, two bands, and cells that hold
WORDS.** Measured over the 35 grid pages (§V.1), words-per-cell is what separates a table of text
cells from a table of marks, and it does so with a wide margin:

| page | kind | cols | bands | words per cell | verdict |
|---|---|---|---|---|---|
| p15 | prose page with a table | 3 | 5 | 4.3 | reconstructs |
| p33, p34 | Table 1 | 3 | 3, 5 | 111, 59 | reconstructs |
| p44 | Table 1 | 4 | 4 | 41 | reconstructs |
| p57 | Annex A mapping | 3 | 11 | 8.3 | reconstructs |
| p28 | a FIGURE the detector fires on | 2 | 9 | 4.4 | reconstructs (see below) |
| p25 | a boxed paragraph (the known false positive) | 1 | 2 | — | falls back: one column |
| p51, p54, p55, p66 | mark grids | 12–20 | 4–41 | 0.43–0.96 | falls back: the cells hold marks |

So: `MinWordsPerCell = 2.0`, at least 2 columns, at least 2 bands. Otherwise the page falls back
to plain text with the reason stated, as today.

**p28 is accepted and that is the honest direction.** It is a figure of boxes, and the detector
already fires on it (`Grid300` documents over-detection as the failure it chooses). Its cells are
the figure's own boxes, so the rows it emits are the figure's text in its own arrangement —
compared against the oracle, which renders that figure as pseudo-table rows too, this is no worse
than the plain reading it replaces. What the record must not do is call it a clean table, so the
`text_cells` fields state what was measured and `columns` says 2.
*Beaten:* a shape test on how consistently bands repeat their column boundaries — measured, and it
does not separate the cases (p28 scores 1.00 on boundary agreement, p44 scores 0.50).

**D6. A column is its segments, and a band holds the rows its print separates by space.** Two
rules the measurement forced, each with the page that forced it:

- **p45** is a three-column Table 1 page where not one vertical component crosses a band whole:
  the scan breaks each printed rule where it meets a row line. Asking whether any single component
  spans the band answered no, and the page fell back. The segments at one x are one column, and
  their covered length is what crosses.
- **p34** has six horizontal rules and eight rows: IEEE 1012's Table 1 rules its section headings
  and separates the tasks inside a section by WHITE SPACE. Binding the ruled band alone still
  leaves "the required inputs of task (2)" unanswered, which is the issue's own example. So a band
  is cut at every y no word crosses with at least `minRowGap` (32 px at 300 DPI) of blank: measured
  on that page, the gaps between tasks are 45, 53 and 88 px and the gaps between lines inside a
  task's prose are 10-15 px. With the cut, p34 reads as the oracle's eight rows, each task beside
  its own inputs and outputs.
  - The rules are not faint: counted at four binarization thresholds and under an adaptive one, the
    page has exactly six horizontal rules. The missing rows are the print's, not the reader's.
- Acceptance counts the ROWS the reading emits, not the ruled bands, so a one-band table of two
  rows is a table.

**D7. Layout analysis drops a short word alone in a narrow column, and the sparse pass is the
second witness.** Measured on the generated fixture: "High" and "Major" are absent from the auto
TSV and present in the sparse one, exactly as isolated marks are. Cells the first pass left EMPTY
are filled from the sparse pass, and the count is a field (`words_from_sparse`) — so a cell never
takes two passes' readings of the same words.

**D8. A page may carry more than one table, and a column says where one ends.** The generated
two-table page read as ONE lattice at first: the two tables' columns merged and the sentence
between them landed in a cell. Height cannot separate them — this corpus has single rows 676 px
tall and two tables 416 px apart — so the test is whether a COLUMN crosses the gap. A column runs
through its table's rows and stops at its bottom rule. Each table is then bounded by its own
cells, not by the page's rules, so a word between two tables belongs to neither.

**D9. A rotation is adopted on a margin, not a majority.** The probe fired on the generated pages
and turned two of them sideways, which destroys a lattice: rules no longer bound the text they
bound. Measured at 300 DPI: the corpus's genuinely landscape tables read 9 words portrait and 97
and 94 rotated; the false positives read 15 → 24 and 0 → 8. So the rotated pass must read at least
3× the portrait one AND clear the probe's own floor of 25 confident words. This also narrows
#934's sideways false-accept.

**D10. The better pass fills the cells, and the other fills what it left empty.** Layout analysis
reads a dense page best and drops words on a sparse one (measured: 22 words where the sparse pass
read more). Whichever pass read more words fills first; the other fills only cells still empty, so
neither pass's dropout decides what the page says, and no cell takes two readings of one word.

**D5. The fields.** `PageReading` gains `cells`, `columns`, `words_placed`, `words_outside` and
`cell_lines` under a `text_cells` object, written only on the lattice path. The fallback sentence
gains its own case, naming the consequence ("the rules bound N cells and M of the page's words
landed outside them, so the row binding could not be recovered").

**Components.**

- `[MODIFY] internal/tessocr/shim.{h,cpp}`: `tessocr_grid_lines` (spiked).
- `[NEW] internal/tessocr/lattice.go` + `lattice_test.go`: `Lattice`, `ParseLattice`, `Cells`,
  and the cell-filling and rendering, all pure Go and testable tag-less.
- `[MODIFY] internal/tessocr/engine_cgo.go` / `engine_stub.go`: `GridLines`.
- `[MODIFY] internal/tessocr/page.go`: the second reconstructor, the new stats, the fallback case.
- `[MODIFY] internal/fetchcache/pageread.go`: the `text_cells` fields on `PageReading` and the
  receipt, so a reused receipt carries them.
- `[MODIFY] internal/cli/fetchsummary.go`: the fetch summary says how many pages reconstructed as
  text cells, beside `table_pages`.
- `[MODIFY] internal/cli/fetch.go` Long + help goldens: one sentence on what a text-cell table
  reads as now, replacing the sentence that says its row binding is lost.
- `[NEW] internal/tessocr/testdata/*.png` + `testdata/gen`: EIGHT generated pages, one per layout
  case — a ruled table of text; rows separated by space inside a band; rules the scan broke; a grid
  of marks; a boxed paragraph; prose with a small table; two tables on one page; the same table at
  photocopy contrast. Generated rather than cropped, so nothing licensed is checked in and the
  bytes are reproducible from the generator.
- `[NEW] internal/tessocr/goldens_cgo_test.go` + `testdata/*.golden`: each page's WHOLE reading,
  byte for byte, with the path taken and what it measured above it. `-update` rewrites them. A
  regression is a diff a human reads, never a judgement a model makes. They pin a BUILD: an engine
  or traineddata pin bump moves them, and reading that diff is the review the bump deserves.
  - Two fixtures pin engine behaviour we do not control and did not choose: a grid of marks whose
    glyphs neither pass reads, and two rows of the two-table page the engine drops in both passes.
    The golden states what the reader does with them rather than pretending it does better.

## IV. Risk & Mitigation

- **R1. A rule that stops short splits a column (likelihood medium, impact medium).** `coverage`
  (0.8 of the band) is the guard, and §V.2 measures every band of the 35 grid pages.
- **R2. A doubled rule reads as two bands (low × medium).** `ruleJoin` and `minBandHeight` collapse
  them; both are pinned by a test on a real crop.
- **R3. The lattice accepts a page whose text it mangles (medium × high).** D4's ratio plus §V.3's
  oracle comparison; the failure direction is a fallback to today's reading, never a silent table.
- **R4. p54 regresses (low × high).** It never reaches the lattice path; §V.2 pins its text_sha.
- **R5. Per-page cost (low × low).** One more leptonica pass per grid page, on openings the
  detector already computes — measured in §V.5.

## V. Verification Plan

Run from `plugins/frank-exchange-of-views/tools`. Re-armed by anything under `internal/tessocr` or
`internal/fetchcache`.

1. **The spike's own figures, kept.** `~/ocr-runs/932-spike-all.txt` (all 35 grid pages) and
   `~/ocr-runs/932-spike-shape.txt` (the acceptance measures above): bands, columns, words placed
   and words outside for all 35 grid pages. The plan's claims are read off it, and the
   implementation is measured against it again at the end.
2. **Whole-document E2E**, tagged binary, fresh run dir: every page reads; p54's text is
   byte-identical to main's; the 30 text-cell pages carry `text_cells` fields; no page's reading
   is empty. `sha256sum p????.txt` against `~/ocr-runs/644-real/e2eCf`, with the changed set
   listed and read.
3. **Oracle comparison on a pre-registered sample** (p34, p57, p15, p44, p66, chosen before
   looking): each row's cells against `~/ocr-runs/assembled/p00NN.txt`, whitespace-normalised.
   Recorded as cells matching, cells differing, and the differences read by hand.
4. **Delete the row.** Each of D4's three acceptance conditions, `coverage`, `ruleJoin`,
   `minBandHeight`, and the words-outside count: removed one at a time under `go test -overlay`,
   each must turn a named test red.
5. **Cost.** The added leptonica pass, timed over the 35 grid pages, reported as ms/page.
6. `go vet ./...`, `go test -count=1 ./...`, the tagged engine suite, `go -C scripts run ./check`.

## Open questions (gblock's)

- **Q1.** The reading is plain text, so a cell's internal line break becomes a space (D3). The
  oracle writes `<br>`. Is the space right, or should the reading keep the break?
- **Q2.** Should the fetch summary count text-cell pages separately from `table_pages`, or is one
  count enough?
