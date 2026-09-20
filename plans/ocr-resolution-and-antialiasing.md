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

## Decision

Keep 1:1 native. Do not anti-alias. Do not upscale by default.

**What this does NOT close:** that one page improved markedly under exact 2× suggests a
page-shaped question rather than a document-shaped one — a dense mark grid may genuinely want
more pixels. Any future attempt should be a per-page decision made on a measured property, not a
global policy, and should use an INTEGER multiplier bounded by an absolute pixel budget rather
than a DPI cap. The old conclusion (upscaling is worthless) was right; the old reasoning was not,
and a reader of #1031 should come here before trusting its numbers.

Harness: `zz_res_cgo_test.go`, scratch, in `~/ocr-runs/resolution.log`.
