# The geometry layer reads one coordinate space

Status: revision 2, IMPLEMENTED and measured 2026-09-20. Carries #1074. §VI holds the result.
Revision 1 proposed deriving ~20 constants per page and failed its audit on six gaps. The audit
also put a fork revision 1 never considered, and it is better than what revision 1 proposed; this
revision is built on it. **[r2]** marks what moved.

## I. Summary & Goals

**The defect.** #1058 renders each page at its own scan resolution — floored 300, capped 600 — and
derived the DETECTOR's thresholds from it (`GridFor(dpi)`). Nothing below the detector came with
it. The lattice geometry, the reconstruction's clustering, the header-band probe and the level
cropper are pixel constants measured at 300 DPI, compared against coordinates that now arrive at up
to 600.

**Measured.** `usfs-birds-markgrid` (corpus, native 501 — a dense species-by-habitat mark grid),
one physical table read at five resolutions with today's build:

| DPI | bands | rows | columns | cells | `must_contain` |
|---|---|---|---|---|---|
| 300 | 2 | 7 | 8 | 16 | 5/7 |
| 350 | 2 | 9 | 9 | 17 | 4/7 |
| 400 | 2 | 8 | 13 | 25 | 5/7 |
| **501 — what ships** | **3** | **10** | **13** | **37** | **4/7** |
| 600 | 2 | 18 | 11 | 21 | 3/7 |

One table, five shapes. Scaling five of the constants by `dpi/300` returns the 501 read to **2
bands, 7 rows** — the 300-DPI structure exactly; its columns stay at 11 against 300's 8.

`nbs602-form` is stable across the whole range (2 columns, 11 bands, 5/6 at every DPI), which is why
a spot-check on one page reads clean and why this survived #1058's review.

**The change. [r2]** Not twenty derived constants — **one normalization at one boundary.** Every
measurement crossing into the geometry layer is converted to the 300-DPI space it was fitted in:
lengths by `300/dpi`, area-like pixel counts by `(300/dpi)²`. The geometry layer is then unchanged,
every constant in it stays exactly as measured, and the coordinates that address the page image are
mapped back at the two sites that need them.

Revision 1 proposed a `Tune` struct carrying ~20 derived constants threaded through 8 functions.
Its audit found **twelve more per-DPI literals** the census had missed — `medianPitch` fallbacks,
`gtolC`/`gtolR`'s 16 px arm, four same-line tolerances, a second mark-minimum site — and named the
real defect in that shape: a hand-kept census with nothing that fails when the next unscaled literal
is written is [[facts-are-fields]]'s "guard whose own allowlist is hand-kept", which is how #1058
produced this defect one layer up. Normalization needs no census, because it does not enumerate.

**Goals.**

1. Every geometric decision below the detector is made in **one coordinate space**, so a constant
   fitted at 300 is compared against a 300-DPI quantity whatever the page rendered at.
2. A page rendered at 300 reads **byte-identically**, by construction: the transform is the identity
   at 300, not an approximation of it.
3. Differences in a page's reported structure across resolutions come from **recognition**, not
   arithmetic. **This is deliberately weaker than revision 1's goal 3, and the reason is a
   correction: [r2]** tesseract reads different pixels at 300 and 600 and returns different words, so
   the structure built from them legitimately differs. Revision 1 demanded exact equality of
   bands/rows/columns across resolutions and would have had an implementer weaken a red gate at the
   console. §V.2 pins the achievable property instead — the geometry layer given the SAME input at
   two resolutions returns the SAME structure — and §V.6 reports the end-to-end spread without
   asserting it.

**The value case, and the alternative that would make this unnecessary. [r2]** The audit put a
one-line fork: pin `RenderDPIFor` back to a fixed 300 and every 300-DPI constant is correct by
construction. Its stated price is at `pagerender.go:52-66` — native rendering took the corpus's
expect-failures from 13 to 10. **That figure is not trustworthy as a reason to keep native
rendering, and saying so is part of this plan:** it was measured with the geometry unscaled, so it
compares "300 everywhere" against "native with broken arithmetic", and on the one page whose native
resolution is far from 300 the change was a REGRESSION (`usfs-birds-markgrid`, 5/7 → 4/7). The
honest position is that native rendering's benefit is unmeasured in a correct regime. This plan
makes the regime correct, at which point the question can be asked properly; §V.6 records the
numbers to ask it with. Reverting to 300 first would discard #1058 on the strength of the same
untrustworthy figure, in the other direction.

## II. Technical Context

- **Where the boundary is.** `ReadPage` (`page.go:150`) receives the page PNG at its native
  resolution and produces everything downstream from two inputs: the rule geometry from the C
  detector (`GridLines`, page pixels) and tesseract's TSV word boxes (page pixels). Every function
  below it — `Lattice.Tables`, `Lattice.Cells`, `TextCells`, `LevelBandCells`, `Reconstruct`,
  `headerBand`, `fallbackReason` — consumes those two and nothing else resolution-bearing.
- **The detector is already correct and is NOT in scope.** #1031 derived `GridFor(dpi)` properly;
  the C morphology must run on real pixels at the real resolution. This plan changes nothing above
  the boundary.
- **What leaves the geometry layer as a coordinate.** `Cell{X0,Y0,X1,Y1}` is used by `CropCell`
  (`levels.go:151`) to crop the actual page image, and `headerBand` returns a page rectangle. Those
  are the sites that map back.
- **Build split.** The pixel work is C behind the `tessocr` tag; everything here is Go in the
  default build, so §V.2's gate runs on every leg — including Windows, at no engine cost.
- **`Identity()` hashes this package's source** (`tessocr.go:58-96`), so this change re-keys every
  stored reading: cached runs re-OCR on next read, self-healing through the receipt check at
  `pageread.go:260`. Stated so it is not discovered on a run. **[r2]**

## III. Proposed Changes

**D1. Parse once at the edge, normalize there, pass structs down. [r2.1]** `scale := 300.0 / dpi`.
Nothing below changes: `ruleJoin` stays 12, `cellPad` stays 6, `gtolC`'s 16 px arm stays 16, and
each is once again compared against the quantity it was fitted against.

**The mechanism, corrected against the code before any of it was written.** Revision 2 said the
transform sits on "two input structures at the top of `ReadPage`". The rule geometry is a struct
(`ParseLattice`, one call site) but the TSV is **a string that crosses the boundary unparsed and is
re-parsed at eight separate sites** (`parseTSVWords`). Normalizing "at the top" would therefore have
left seven of them reading native pixels — the half-applied transform R4 names, arrived at by
following the plan exactly. So: `readGridPage` parses the TSV ONCE into `[]tsvWord`, normalizes that
slice and the `Lattice`, and passes the SLICE down; the string survives only as evidence. This also
removes seven redundant parses of a page-sized table.

Three of the eight sites are unaffected either way and are named so the reader need not re-derive
it: `page.go:313` (cell text) and `page.go:359` (`confidentWordCount`) read only `.text` and
`.conf`, never geometry; `page.go:373` (`headerBand`) needs NATIVE pixels because it returns a
rectangle used to crop the page image, and takes the native slice explicitly.

**D2. What crosses, and how each kind transforms.**

| crossing the boundary | kind | transform |
|---|---|---|
| rule box `X0,Y0,X1,Y1` | length | `× 300/dpi` |
| TSV word `x, y, w, h` | length | `× 300/dpi` |
| `GridStats.Intersections` (D4) | area-like pixel count | `× (300/dpi)²` |
| TSV confidence, text | neither | unchanged |

There is no third case and no list to keep: a quantity is a length, an area, or neither, and the
transform is a property of the field's TYPE rather than of a constant somebody remembered to add.

**D3. What addresses the page image stays in page pixels.** Two sites: `headerBand`'s rectangle
(which takes the native slice, so nothing is converted) and the `Cell` handed to `CropCell`, which
is converted back by `× dpi/300`. That single map-back is where the rounding this design costs is
spent — sub-pixel at worst, against a crop already inset by `cellPad`.

**D4. The intersection ratio, in the same idea.** `MaxIntersectionRatio = 4.0` compares
`GridStats.Intersections` (crossing PIXELS, ∝ DPI²) against `Stats.ExpectedIntersections()`
(lattice POINTS, `(rows+1)×(subcolumns+1)`, DPI-invariant). `page.go:96-110` says the ratio is not
unitless and names that as its limit — it was measured on one document at 300, so at 501 the same
healthy reconstruction scores ~2.8× higher and the gate silently tightens. Normalizing
`Intersections` by `(300/dpi)²` at the boundary puts numerator and denominator back in the units the
4.0 was measured in. **The constant does not move**; the measurement handed to it does. **[r2:
revision 1 loosened the limit to `4.0 × (dpi/300)²`, which is the same arithmetic in the more
dangerous direction — a safety gate whose written value no longer matches the one in force.]**

**D5. What is NOT normalized, stated as a decision.** Confidences (`bandMinConf`,
`confidentWordMinConf`), counts of things (`MinTextCellColumns`/`MinTextCellRows`,
`MinWordsPerCell`, `bandMinBoxes`, `rotAdoptRatio`, `confidentWordMinRunes`,
`rotProbeMaxConfidentWords`, `levelCaptionMin`), ratios (`coverage`, `bandMinAspect`,
`MinMarkPlacement`) and text lengths (the 24-rune cap in `isStrongMark`) are not geometry and are
untouched — and under this design they need no invariance test, because nothing scales them.
**[r2: revision 1 promised `TestTuneInvariantsDoNotScale`, which would have compared fields that its
own D2 never put on the struct — a test with nothing to compare, passing forever.]**

**D6. Alternatives considered and rejected. [r2]**

- *Derive every constant per page* (revision 1). Rejected: its census was missing twelve literals
  when audited, and nothing in the design would fail when the thirteenth is written.
- *Pin the render back to 300.* Rejected as premature, not as wrong — see §I's value case. It
  remains the fallback if §V.6 shows native rendering buys nothing once the arithmetic is correct,
  and that is a decision for a measurement, not for this plan.
- *Widen `GridThresholds` to carry the geometry.* Moot under normalization: no new carrier exists,
  so the naming collision revision 1 created — a `Tune` type beside seven existing uses of "tune"
  for the detector's threshold set — does not arise. **The word keeps one referent.**

**Components.**

- `[MODIFY] internal/tessocr/page.go` — D1's transform at the top of `ReadPage`; D3's map-back for
  `headerBand`; `fallbackReason` reads the normalized intersections. The file header comment
  ("Every pixel constant in this file is a 300-DPI empirical fit … the read path refuses other
  resolutions", false since #1058) says what is true.
- `[MODIFY] internal/tessocr/levels.go` — D3's map-back at `CropCell`.
- `[MODIFY] internal/tessocr/tessocr.go:104-110` — `RenderDPI`'s doc ("Every pixel constant below is
  a 300-DPI fact; reusing one at another DPI is exactly the mistake") becomes the statement of what
  the normalization guarantees.
- `[MODIFY] internal/tessocr/{lattice,reconstruct}.go` — comments only (`lattice.go:99`,
  `reconstruct.go:14,27` say "px at 300 DPI (RenderDPI)"); the constants themselves do not move,
  which is the point of this design.
- `[MODIFY] internal/fetchcache/pageread.go:44,145-150,441` — the doc comments, one of which names
  `tessocr.MaxIntersectionRatio` and its units.
- `[NEW] internal/tessocr/normalize_test.go` — §V.2 and §V.3's gate.
- `[MODIFY] testdata/corpus/*/reading.golden`, `STATUS.md` — regenerate; the five pages above 300
  change. Every changed line is read.
- Signatures that change from `tsv string` to `words []tsvWord`: `TextCells`, `LevelBandCells`,
  `Reconstruct`, `reconstruct`. `ReadPage`'s own signature does NOT change, so the five test files
  that pass `Grid300` to it are untouched. **[r2.1: revision 2 claimed no signature changes at all,
  which the eight parse sites make false.]**

## IV. Risk & Mitigation

Graded likelihood × impact × complexity-to-mitigate.

- **R1. A 300-DPI page stops reading identically (low × high × low).** `scale = 300.0/300 = 1.0`, so
  the transform is the identity and the three 300-DPI corpus pages plus every generated fixture must
  be byte-identical. §V.2 asserts the identity case directly rather than inferring it from a golden.
- **R2. Rounding at the map-back changes a crop (medium × low × low).** `CropCell`'s rectangle is
  already inset by `cellPad` and the error is sub-pixel; `levelgrid.golden` is the check.
- **R3. D4 changes the dropout gate at every resolution but 300 — MEASURED, no corpus verdict moves
  (medium × high × medium).** Read out of the committed goldens, only two of eight pages reach the
  gate: `nbs602-form` (350) falls back on the PLACEMENT threshold, which D4 does not touch, and
  `usfs-birds-markgrid` (501) measures **338.2**, which normalizes to 338.2 × (300/501)² = **121**,
  still far above 4.0. The other six never reconstruct. **The residue, stated:** the corpus does not
  exercise the band this changes, so D4 rests on the units argument, not on a page. §V.3 is the
  check that the argument holds; §V.5 re-reads the two pages after.
- **R4. The transform is applied to some inputs and not others (low × high × low).** A half-applied
  normalization is worse than none — a constant compared against a mixture of spaces. The mitigation
  is structural: the conversion happens once, on the two input structures, at the top of `ReadPage`,
  so "did this field get normalized" is answered by where the value came from rather than by
  inspection.
- **R5. The corpus goldens move on five of eight pages (certain × low × low).** That is the change
  working. Every changed line is read; `STATUS.md`'s meter is a per-page yes/no and 5 of 8 already
  read "no", so **the meter cannot show improvement on those five** — the "what still fails" column
  and the goldens are what carry it. **[r2: revision 1 claimed the meter was the check.]**
- **R6. Recognition still differs across resolutions and always will (certain × low × none).** Not a
  risk this plan mitigates — a fact goal 3 is written around.

## V. Verification Plan

Each step in its own shell; the cstack environment leaking into a golden step makes it fail. Run
from `plugins/frank-exchange-of-views/tools`.

1. `go build ./... && go vet ./... && go test -count=1 ./...` → 57 packages, **0 failed**.
   Re-armed by any `.go`.
2. **The gate this plan exists for, and it is PURE GO. [r2]** `TestGeometryIsResolutionInvariant`:
   take one fixed set of rule boxes and TSV words, express it at 300 DPI and at 600 (coordinates
   doubled, the same page), run the geometry layer on both, and demand **identical** structure —
   bands, rows, columns, cells, and the emitted text. The identity case (300 → 300) is asserted in
   the same test, which is goal 2. Re-armed by any change in `lattice.go`, `reconstruct.go`,
   `levels.go`, `page.go`.
   **Why this and not two engine reads:** revision 1's gate read one real page at two resolutions
   and demanded equal structure, which tesseract's own variation makes false — the implementer would
   have met a red gate and weakened it. This pins exactly what the change delivers, runs on every CI
   leg including Windows, costs no engine time, and needs no workflow edit. Revision 1's gate would
   have run in **no CI job at all**: `hooks.yml:366` filters `-run '^TestCorpusGoldens$'` and the
   tagged Linux leg does not set `FEOV_OCR_CORPUS`. **[r2]**
3. **The D4 exponent, checked rather than asserted.** In the same pure-Go test, one lattice at two
   resolutions must produce the same normalized `Intersections / ExpectedIntersections()` within a
   stated tolerance. If the exponent were 1 instead of 2 this fails; today nothing would.
4. **Delete-the-row, and it is not self-certifying. [r2]** Remove the length transform, remove the
   area transform, and remove the map-back, one at a time — each must turn §V.2 or §V.3 red.
   Revision 1's equivalent let the implementer choose which of twenty fields to revert, and only
   fields the chosen page exercised could go red.
5. `eval "$(./third_party/pins/build-cstack.sh env linux-amd64 ~/ocr-runs/644-real/cstack)"` then
   `go test -tags tessocr -count=1 -ldflags '-linkmode external -extldflags "-static"'
   ./internal/tessocr/ ./internal/fetchcache/` → ok ok. Then re-read the two dropout lines of R3
   and name any verdict that moved. **Accept/reject, decided now rather than after the number is
   known: a page that moves from refused to PUBLISHED blocks this change until the reading is
   checked against the page image by eye; a page that moves from published to refused is acceptable
   and recorded.** **[r2: revision 1 required only that a moved verdict be "named", which a flip to
   published passes by being written down.]**
6. **The end-to-end spread, REPORTED and not asserted.** `usfs-birds-markgrid` at 300/350/400/501/600
   after the change, beside §I's table. Expected: the arithmetic-driven variation gone, a residue
   from recognition. This is also the measurement §I's value case needs to ask whether native
   rendering earns its keep. Scratch harness: `~/ocr-runs/1070/zz_dpiab_cgo_test.go`.
7. `FEOV_OCR_CORPUS=1` + the tagged command on `./internal/fetchcache/ -run TestCorpusGoldens`;
   `-update` regenerates goldens AND `STATUS.md`. Read every changed line.
8. `go -C scripts run ./check` → read the FAIL COUNT.

## VI. What it did, measured

**`usfs-birds-markgrid` (native 501), the page the defect was found on.** Same five resolutions,
after the change — §I's table is the before:

| DPI | bands | rows | columns | cells | chars | `must_contain` |
|---|---|---|---|---|---|---|
| 300 | 2 | 7 | 8 | 16 | 1069 | 5/7 |
| 350 | 2 | 9 | 11 | 22 | 1062 | 4/7 |
| 400 | 2 | 7 | 9 | 17 | 1077 | 5/7 |
| 501 | 2 | 7 | 11 | 21 | 1057 | 4/7 |
| 600 | 2 | 8 | 12 | 20 | 1063 | 3/7 |

**The headline is the last column but one.** The page's text length spanned **987–1377 characters**
across the band and now spans **1057–1077**: a spread of 390 characters became 20. Bands are
constant at 2 where they were 2/2/2/3/2; rows are 7–9 where they were 7–18, and the 600-DPI reading
that invented 18 rows out of a 7-row table now reports 8.

**`nbs602-form` (native 350), the page that looked fine.** Rows were 18/17/19/19/22 and are now
18/17/18/18/18; its cells were already 15 at every resolution and still are. A page that a
spot-check called stable was drifting by four rows at the top of the band.

**What did NOT come back, as §I said it must be reported rather than quietly dropped:** columns
still range 8–12 across the band, against 8 at 300. So a column count is not the invariant it
looks like, and this plan does not deliver one. The evidence says it is recognition rather than
arithmetic — `must_contain` moves in step with it (5/4/5/4/3) and the pure-Go gate shows the
geometry itself is exactly invariant on fixed input — but that is an inference, not a measurement,
and it is **#1094** rather than a line quietly softened in §V.

**The dropout gate, as R3 predicted, moved no verdict.** `usfs-birds-markgrid` read 338.2 against a
limit of 4.0 and now reads **121.3** against the same 4.0 — refused either way. `nbs602-form` still
falls back on the placement threshold. The other six pages never reach the gate.

**Corpus goldens:** eight files, seven changed only their engine identity hash (`Identity()` hashes
this package's source). The eighth is the birds page above. `STATUS.md` did not move — the meter is
a per-page yes/no and that page was already failing, which is R5 and is why the meter is not the
instrument for this change.

**Delete-the-row:** four mutations, three killed — removing the length transform, squaring the
crossings by `dpi` instead of `dpi²`, and making the cell map-back the identity all turn the gate
red. The fourth survived and is recorded at its source: scaling a rule's width directly instead of
deriving it from the scaled edges differs by at most one pixel, and nothing downstream has a
tolerance that tight. The comment there says so rather than claiming a defect it does not fix.
