# The page's background is flattened before anything reads it

Status: measured and decided 2026-09-19. Carries #1032, and closes #1028 as a side effect —
the same change removes the cover's false table.

## I. Summary & Goals

**The defect.** `usfs-birds-markgrid` (#1004's corpus) is a species-by-habitat tick grid printed
with alternating grey shaded columns. A page of legible names read as **56 characters**, and they
were `Aduapisay aaaa NNNNNNANHNNMNGEA …` — "Residency" recovered upside down from a rotated header
and nothing else. The shading also gave it **1251479 grid intersections**, so the detector called it
a table for the wrong reason.

**Why the obvious knobs cannot fix it.** Read from the sources we build against:

- `tesseract 5.5.3`: `Tesseract::BestPix()` returns `pix_original_` whenever its width matches the
  page, and `SetImage` sets `pix_original_` from what we hand it — so **the LSTM reads OUR pixels**.
  `thresholding_method` assigns `pix_binary_`/`pix_grey_` only, which layout consumes. It cannot
  change what the recognizer sees.
- A global threshold cannot separate tint from ink page-wide, which is why raising our fixed 180 is
  not the alternative; leptonica says so directly for Otsu — it "does a very poor job when there are
  large amounts of dark background", and this page is 63.1% ink at 180.
- Sauvola alone does not save it either: inside a uniform grey band the local variance goes to zero
  and the threshold collapses below the band's own grey. Lazzara & Géraud name low contrast and
  contrast inversion as two of its three structural defects.

**The change.** `pixBackgroundNormSimple` in the shim's `read_png`, so the normalized pixels are
what `SetImage` receives AND what the grid detector binarizes. One call, both stages.

## II. Measured before it shipped

**The corpus (8 pages), each read twice:**

| page | chars | expect-failures | detector |
|---|---|---|---|
| `usfs-birds-markgrid` | **54 → 1069** | 7 → 2 | intersections 1251479 → 2250; now reconstructs 7 rows × 8 columns, 121 words placed |
| `nbs602-cover-photo` | 278 → 56 | 1 → 1 | intersections 1256009 → **0**, table **true → false** |
| `nbs602-dashed-matrix` | 1631 → 1812 | 4 → 4 | unchanged (that is #1027) |
| `nbs602-form` (control) | 2582 → 2577 | 1 → 1 | table true → true, intersections 252 → 319 |
| the other four | ±17 | unchanged | unchanged |

**The tuning document (IEEE 1012, 80 pages), where every constant was fitted:**

- **The detector does not move.** 35 pages called tables with normalization off, 35 with it on,
  **zero flips**. With normalization off this build also agrees with `grid300-sel151.txt` — the
  tune's own record — on all 80 pages, so the comparison is against the record and not a drifted
  build.
- **Recognition is a wash.** 18763 alphabetic words (4+ letters) become 18769: +0.0%. 63 pages
  byte-identical, 7 better, 10 worse. The losses are noise churn on sideways pages (`Aeuouy` becomes
  `Ajswouy`); the biggest gain is real — p0007's contents page read `1.1 1.2 1.3 1.4 15` with its
  titles eaten by dot leaders, and now reads `1.2 Field of application`, `1.4 Organization of the
  standard`.

**The generated mark grid, scored against the pattern its generator draws** (marks
`{0,1} {0,1,2,3} {0} {1,2} {0,1,2} {3}` rotated per row):

| | correct of 52 | invented | missed |
|---|---|---|---|
| before | 22 | 5 | 30 |
| after | **29** | **4** | 23 |

## III. The cost, stated

**`nbs602-cover-photo` reads less text.** It keeps a clean `NBS SPECIAL PUBLICATION 602` (before,
that line was `I. i NBS SPECIAL PUBLICATION | 602 % | a CReay 4 of 2` inside an invented table) and
loses two secondary lines: `U.S. DEPARTMENT OF COMMERCE / National Bureau of Standards` and
`Research Workshop`. Its `expect` block still demands the first of those, so the corpus records the
loss rather than absorbing it.

This is leptonica's documented failure mode arriving exactly where it was predicted: background
normalization assumes a real local background lighter than the print, and a photographic cover does
not have one. The trade taken — an honest verdict and a correct title line on that page, against two
lines of cover prose — is recorded here so that reversing it later is a decision with a reason
rather than a rediscovery.

**Not taken:** a guard that skips normalization on pages that look photographic. That is a threshold
on a threshold, and the evidence for it is one page. If a second page shows the same loss, the guard
gets a case; today it would be fitted to a single specimen.

## IV. What moved, and what a reviewer reads

- `internal/tessocr/shim.{h,cpp}` — `normalize()` in `read_png`. The measurement switch that made
  the A/B possible is REMOVED; there is one path.
- `internal/tessocr/engine_cgo_test.go` — the real crop's pinned counts, 23375/21440/367 →
  32364/25284/594. More rule pixels is what recovering a faint rule looks like; the verdict is
  unchanged.
- `testdata/levelgrid.golden`, `testdata/markgrid.golden` — 2 of the 9 generated pages move; the
  other 7 are byte-identical, which is what a near-identity on a clean white page should look like.
- `testdata/corpus/*/reading.golden` and `STATUS.md` — 6 of 8 corpus pages move.

The engine identity carries the shim's source hash, so every golden that names the engine moves with
it, loudly.

## V. Verification

1. `go test -count=1 ./...` in the tools module.
2. `eval "$(./third_party/pins/build-cstack.sh env linux-amd64 <cstack>)"` then
   `go test -tags tessocr -count=1 -ldflags '-linkmode external -extldflags "-static"' ./internal/tessocr/ ./internal/fetchcache/`.
3. `FEOV_OCR_CORPUS=1 … -run TestCorpusGoldens` — the corpus, read and pinned.
4. The three A/B measurements above, reproducible from `~/ocr-runs/1032/scratch-tests/`.
5. `go -C scripts run ./check`.
