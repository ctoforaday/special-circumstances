# Tesseract is guessing our page resolution, and telling it the truth breaks determinism

Status: measured 2026-09-20, **NOT SHIPPED**. The diff is at `~/ocr-runs/dpi-declare/declare-dpi.diff`.

## The defect, measured

We never tell Tesseract what resolution a page is. `read_png` calls `pixReadMem` on a PNG, Go's
encoder writes no `pHYs` chunk, so the PIX arrives with a resolution of **0**. Tesseract then takes
`kMinCredibleResolution` (70) and re-estimates from body-text line height x 10 — **and the warning
it would print is suppressed when the resolution is precisely 0**, so nothing says it happened.

Its guess, against what we actually rendered, over the eleven corpus pages:

| page | rendered | tesseract guessed | error |
|---|---|---|---|
| gpo-record-threecolumn | 300.0 | 301 | 0% |
| doe-dotmatrix-listing | 300.0 | 296 | −1% |
| usgs-bibliography-jpeg | 400.0 | 394 | −2% |
| nbs602-form | 350.4 | 341 | −3% |
| usfs-birds-markgrid | 501.2 | 534 | +7% |
| doe-equations-figure | 300.0 | 325 | +8% |
| nbs602-dashed-matrix | 350.2 | 383 | +9% |
| nbs602-contents | 350.2 | 439 | +25% |
| **usgs-openfile-jbig2** | 300.0 | **521** | **+74%** |
| **nbs602-blank** | 350.5 | **616** | **+76%** |
| **nbs602-cover-photo** | 350.3 | **939** | **+168%** |

The guess comes from line height, so a page of sparse or unusual text fools it badly.

**It matters because layout analysis is denominated in INCHES and multiplied by that number.**
`kAlignedFraction` (0.03125 in) and `kMaxGutterWidthAbsolute` (2.00 in) in `tabfind.cpp` decide
where columns and TABLES are, and `table_finder.set_resolution()` takes the same value. Recognition
reads grayscale pixels and is indifferent; the structure a citation's row and column come from is
not. The Sauvola window and adaptive-Otsu tile are also `factor x yres_`, so the alternative
`thresholding_method` values cannot be evaluated at all while the resolution is a guess.

## What declaring it bought

`pixSetResolution(pix, dpi, dpi)` before `SetImage`, with the DPI we already compute exactly:

- The `Estimating resolution as N` line disappears from all eleven pages.
- On the `gridcrop` fixture, page text goes from **327 characters to 413** (+26%) and words from
  49 to 79.
- Across the corpus: net **+2 characters**, and **no verdict moves** — every table/not-table
  decision and rule source is identical.

## Why it is not shipped

**It makes the reader non-deterministic.** `TestTextCellGoldens` on the `levelgrid` fixture, same
binary, same input, fresh engine per subtest, no parallelism, sorted order:

| tree | runs | passes |
|---|---|---|
| without the change | 9 | **9** |
| with it | 5 | **1** |

Two consecutive failing runs produced byte-identical output, so the reading is stable *within* a
process and varies *across* processes. Passing no resolution for the `PSM_SINGLE_CHAR` cell crops
did not fix it, so it is the full-page passes.

A reading that cannot be re-derived byte for byte is the one thing this engine may not be: the
record keys every citation to the engine identity and the image hashes precisely so an audit can
reproduce it. A 26% text gain does not buy that.

## What is not known

Why declaring a resolution makes Tesseract's output vary between processes. Uninitialised state or
an ordering dependency inside a resolution-gated path is the obvious guess and it is only a guess —
nothing here has looked inside Tesseract to confirm it. **Stated as unexplained.**

## If this is picked up again

1. Find the non-determinism before anything else. `TestTextCellGoldens/levelgrid` reproduces it in
   about fifteen seconds a run, roughly four times in five.
2. Bisect the declaration by call site — full-page `PSMAuto`, the `PSMSparseText` second witness,
   `RotatedBand` — rather than all at once, which is what was tried here.
3. If it can be made stable, the rest is already measured: the guess is removed, no verdict moves,
   and `thresholding_method` 1 and 2 become evaluable for the first time.

Sources: `tesseract/src/ccmain/thresholder.cpp`, `baseapi.cpp:2013-2020`, `pagesegmain.cpp:319-325`,
`tabfind.cpp:47-49`, and `publictypes.h:36,38`, all in the pinned 5.5.3 tree.
