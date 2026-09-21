# Upscaling and anti-aliasing, measured in a correct regime

Status: measured 2026-09-20, REJECTED. Revised 2026-09-21 with the published record, which
explains what the measurement alone could not. Recorded so it is not run a fourth time.

## Why it was asked again

Every previous measurement of "does more resolution help" was taken through the defect #1105
fixed. #1031 compared 300, 400 and 600 DPI renders and found 830, 816 and 800 words — and
concluded upscaling buys nothing. A Real-ESRGAN test found a neural upscale indistinguishable
from cubic and from the original.

**Every one of those re-rendered the page at a different DPI, which changed the pixel COUNT and
the PIXELS at once — and always through a fractional resample.** So "more pixels" was never
separated from "interpolation damage", and a conclusion about the world was drawn from data our
own defect produced. That is the same error the confidence finding in #1105 turned on, and it is
why gblock asked whether the resolution question was bound together in ways nobody had untangled.

## The design that separates them

The source raster is held FIXED at 1:1 — which is only possible since #1105 — and only the
enlargement changes:

- **1× native.** The page as scanned.
- **2× exact.** Each pixel replicated into a 2x2 block. Lossless and invents nothing; edges stay
  hard. An INTEGER multiplier is the point: 350 → 700 is exact where 350 → 400 is fractional.
- **2× smooth.** Bilinear. Invents the intermediate values an antialiased glyph edge carries,
  which is what tesseract's models were trained on — the anti-aliasing proposal, in its only
  meaningful form, since anti-aliasing at the SOURCE resolution is just blur.

Ten corpus pages, 45 transcribed expectations, read through `PageText` so the comparison is
recognition and not table reconstruction.

## Result

| arm | expect hits | characters | mean confidence |
|---|---|---|---|
| 1× native | **36 of 45** | 20796 | 74.6 |
| 2× exact | **36 of 45** | 21308 | 71.8 |
| 2× smooth | **33 of 45** | **21469** | **76.2** |

**Anti-aliasing loses, and the shape of the loss is the finding.** The smooth arm reads the MOST
characters and reports the HIGHEST confidence while getting the FEWEST expectations right. That is
precisely the pathology #1105 removed: an operation that invents plausible detail raises the
engine's certainty along with its error rate, so the wrong answer stops being detectable.

The clearest single case is `nbs602-blank`, a blank page: 1× and 2× exact read 0 characters,
and 2× smooth reads **16**. It invents text on an empty page.

**Exact upscale ties and is not worth 4× the pixels.** Per page it is not a wash — the mark grid
`usfs-birds-markgrid` gains two expectations (4/7 → 6/7) while `nbs602-contents` and
`nbs602-dashed-matrix` lose one each — but the total is unchanged and the cost is four times the
raster on every page to find out.

## The mechanism, as far as it is established

**Why exact ties, and this part is verified rather than argued.** Tesseract's LSTM does not read the
image at the resolution it is handed. `lstm/input.cpp` in the pinned source rescales every text line
to a fixed height before recognition:

    int target_height = network->NumInputs();
    image_data.PreScale(target_height, kMaxInputHeight, ...);

`NumInputs()` is the network's own input height, read out of the VGSL spec string in the
`eng.traineddata` THIS REPOSITORY SHIPS: `[1,36,0,1 Ct3,3,16 Mp3,3 Lfys48 ...]` — **36 px**.
(`kMaxInputHeight = 48` is only a cap for variable-height input; an earlier draft of this plan
quoted 48 as the target and that was wrong.)

So a 2x image is normalised straight back down, and replication carries no information the 1x raster
did not. A tie is what theory predicts, and a tie is what the measurement shows. This also explains
why #1031 found the DPI sweep flat — scale per se was never the variable.

It also sharpens what #1105 was about. That fix did not add resolution; it stopped a fractional
resample DESTROYING a sub-pixel feature (a 1 px counter inside a 5). No downstream normalisation
recovers a feature that is gone, so the prediction is a large LOCAL effect and a negligible global
one — measured as one cell flipping and 135 characters across ten pages.

**Why smooth loses IS established, and it was in the literature the whole time.** Three strands,
none of them mine:

1. **Two low-passes where the pipeline budgets one.** Leptonica's `pixScale` is not one filter but
   five, chosen by the scale factor (`scale1.c`): 0.2-0.7 is area-map subsampling with sharpening,
   0.7-1.4 is linear interpolation with sharpening, >=1.4 is linear interpolation with none. A
   300-DPI 10 pt line strip is ~45-55 px, so the reduction to 36 runs at ~0.7 — LI with sharpening.
   Pre-upscale 2x and the strip is ~100 px, so the reduction runs at ~0.36 — area-map. The bilinear
   arm therefore pays a low-pass of ours AND a low-pass of Leptonica's, where the 1x arm pays one.
2. **The documented LSTM x-height ceiling.** Tesseract's own FAQ, scoped explicitly to LSTM: *"there
   seems also to be a maximum x-height somewhere around 30 px. Above that, Tesseract doesn't produce
   accurate results. The legacy engine seems to be less prone to this."* At 300 DPI / 10 pt the
   x-height is ~20 px; doubling it overshoots. The figure is hedged and its only citation is a
   Google Groups thread that no longer resolves, so treat it as a maintainer's prior, not a
   benchmark.
3. **The confidence inversion is a measured, named effect.** Gilbey & Schoenlieb (arXiv:2105.04515)
   measured Tesseract 4.1.1's word confidence against correctness over 20,000 lines and found the
   relationship breaks under resampling — their recalibration curve puts a REPORTED 80 at about 71%
   actually correct. The general law is calibration degradation under dataset shift (Ovadia et al.,
   NeurIPS 2019, whose corruption set includes blur). The reason this recogniser is the worst case
   is CTC: the loss drives posteriors to be peaked as a corollary of training, not as evidence
   (Zeyer et al., arXiv:2105.14849). **"Most characters, highest confidence, fewest correct" is the
   predicted signature of a smoothing operation, not a paradox.**

**The consequence for how we choose anything.** Mean reported confidence is the metric that picked
the losing arm here, and the literature says it will keep doing so off-distribution. It must not be
used as a preprocessing selection signal.

**What remains unexplained is narrower than it was.** The near-blank page reading 16 characters
under smooth and 0 under the other two survived every mechanism proposed for it locally:

1. *An edge bug in the bilinear.* Real, and found: clamping the index at the border while leaving
   the weight negative extrapolates instead of interpolating. **Refuted as the cause:** these pages
   have flat white margins, and on a flat field the bad weights cancel exactly (1.25 - 0.25 = 1).
   Fixed anyway; the re-run is byte-identical on all thirty reads.
2. *Bridging.* That smooth joins nearby marks into a stroke big enough to pass a blob filter — the
   page carries a faint dashed scanner edge-shadow down its margin, which is the shape that would
   do it. **Refuted by measurement:** on a synthetic dashed line, exact keeps 8 marks and so does
   smooth. A 2x bilinear blends across half a source pixel and cannot close a 3 px gap.
3. *More ink.* That smooth darkens the faint mark past a threshold. **Refuted by measurement** on
   the real margin strip: at thresholds 160/200/230 exact reads 0.00%/4.65%/42.08% ink in 0/6/1
   runs, and smooth reads 0.00%/4.39%/43.35% in 0/6/1. Indistinguishable, and slightly LESS ink
   where it would have to be more.

What is left is inside tesseract's own response to soft versus hard edges at equal scale. The
calibration literature above explains why a smoothed input yields confident errors; it does not
explain why THIS page yields characters from a margin artefact. **Stated as unexplained rather than
given a fourth story.**

**One published result points the other way, and it should be on the record.** Gilbey & Schoenlieb,
same engine, 990 pages of real 300-DPI scans, report in passing that *"reducing the 300 dpi images
by a factor of 2 and then enlarging them back to 300 dpi using bicubic interpolation resulted in
slightly better recognition results"*, and call an analysis of it out of scope. That is a low-pass
HELPING, the opposite sign to this plan's result. Reasons it may not transfer: their scans were
printed at 600 and scanned at 300, so they carry sensor noise a low-pass removes; their operation is
down-then-up (one band-limiting filter) where ours is up-then-tesseract's-down (two); and it is an
aside with no numbers. **The cheap test, if anyone wants to close it, is the one they ran: down-2x
then bicubic-up on these ten pages against the 1x baseline, on the same 45 expectations.**

**The effect is also small.** Three expectations out of forty-five separate smooth from the other
two arms. That is enough to decline a change that costs four times the raster and buys nothing, and
it is not enough to support a claim about anti-aliasing in general.

## Decision

Keep 1:1 native. Do not anti-alias. Do not upscale by default.

**If a sub-300 page is ever upscaled, use bicubic, not bilinear.** The one study that compared seven
kernels for Tesseract (Gilbey & Schoenlieb, 20,000 lines) found nearest-neighbour plus a Gaussian
best at 60 DPI and bicubic best at 75; bilinear won nothing at any resolution. Our 300-DPI floor
already sits inside the regime they measured as recoverable — they report 150 DPI upscaled to 300 as
"still excellent" — so the floor is defensible and a 300-DPI TARGET is not.

**What this does NOT close:** that one page improved markedly under exact 2× suggests a
page-shaped question rather than a document-shaped one — a dense mark grid may genuinely want
more pixels. Any future attempt should be a per-page decision made on a measured property, not a
global policy, and should use an INTEGER multiplier bounded by an absolute pixel budget rather
than a DPI cap. The old conclusion (upscaling is worthless) was right; the old reasoning was not,
and a reader of #1031 should come here before trusting its numbers.

## Sources

- Tesseract 5.5.3 pinned source: `src/lstm/input.cpp`, `src/ccstruct/imagedata.cpp`,
  `src/ccmain/linerec.cpp`; the VGSL spec string inside this repository's own `eng.traineddata`.
- Leptonica 1.87.0 `src/scale1.c`, the `pixScale` doc block, for which kernel runs at which factor.
- Tesseract FAQ, "Is there a Minimum / Maximum Text Size?", for the LSTM x-height ceiling.
- Gilbey & Schoenlieb, arXiv:2105.04515 — Tesseract 4.1.1, 990 pages, 20,000 lines: the confidence
  calibration curve, the seven-kernel comparison, the 150-DPI-upscaled result, and the down-then-up
  aside.
- Ovadia et al., NeurIPS 2019, arXiv:1906.02530 — calibration under dataset shift.
- Zeyer, Schlueter & Ney, arXiv:2105.14849 — why CTC posteriors are peaked by construction.

Harness: `zz_res_cgo_test.go`, scratch, in `~/ocr-runs/resolution.log`.
