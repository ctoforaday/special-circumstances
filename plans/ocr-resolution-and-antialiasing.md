# Upscaling and anti-aliasing, measured in a correct regime

Status: measured 2026-09-20, REJECTED. Record so it is not re-run.

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

    const int kMaxInputHeight = 48;
    int target_height = network->NumInputs();
    image_data.PreScale(target_height, kMaxInputHeight, ...);

So a 2x image is normalised straight back down, and replication carries no information the 1x raster
did not. A tie is what theory predicts, and a tie is what the measurement shows. This also explains
why #1031 found the DPI sweep flat — scale per se was never the variable.

It also sharpens what #1105 was about. That fix did not add resolution; it stopped a fractional
resample DESTROYING a sub-pixel feature (a 1 px counter inside a 5). No downstream normalisation
recovers a feature that is gone, so the prediction is a large LOCAL effect and a negligible global
one — measured as one cell flipping and 135 characters across ten pages.

**Why smooth loses is NOT established, and three explanations were tested and refuted.** The
direction is consistent with bilinear low-passing the glyph before tesseract's own downscale, which
costs high-frequency detail the 1x path would have kept. But the sharpest single observation — the
near-blank page reading 16 characters under smooth and 0 under the other two — survived every
mechanism proposed for it:

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

What is left is inside tesseract's own response to soft versus hard edges at equal scale, which
cannot be attributed without instrumenting it. **Stated as unexplained rather than given a fourth
story.**

**The effect is also small.** Three expectations out of forty-five separate smooth from the other
two arms. That is enough to decline a change that costs four times the raster and buys nothing, and
it is not enough to support a claim about anti-aliasing in general.

## Decision

Keep 1:1 native. Do not anti-alias. Do not upscale by default.

**What this does NOT close:** that one page improved markedly under exact 2× suggests a
page-shaped question rather than a document-shaped one — a dense mark grid may genuinely want
more pixels. Any future attempt should be a per-page decision made on a measured property, not a
global policy, and should use an INTEGER multiplier bounded by an absolute pixel budget rather
than a DPI cap. The old conclusion (upscaling is worthless) was right; the old reasoning was not,
and a reader of #1031 should come here before trusting its numbers.

Harness: `zz_res_cgo_test.go`, scratch, in `~/ocr-runs/resolution.log`.
