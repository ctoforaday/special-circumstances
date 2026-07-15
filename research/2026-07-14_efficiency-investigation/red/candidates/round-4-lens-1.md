# red round-4 — lens 1 (leaf-node citation verification), instance 1 of 4

Slice 1: preamble + §0 + §1 + §2 (matching the slice-1 precedent of rounds 2–3).
Full report re-read whole (both pages) before auditing; CHANGELOG used as navigation only.
Round-3 changes in slice: §1.5 (R3-6), §2.4 (R3-10), §2.5 item 1 + carried design (R3-7/R3-8),
§2.1 item 4 (R3-14).

## Verdict (lens-scoped)

One NEW gap (L1-F1, LOW-MEDIUM). Every round-3 repair landing in this slice verified clean at
the leaf. All ledger-stale claims (3 rounds since last verification) re-verified first-hand;
none drifted.

## Findings

### L1-F1 — NEW — LOW-MEDIUM — certain × low-medium × trivial — §2.4 throttle-timing conflation in the R3-10-stated basis

**Location:** §2.4 ("Measured stake"), the sentence: "× **3 throttled rounds (the low-mass
rounds 3–5 of §2.1's series — the rounds a mass-scaled throttle would have throttled; basis
stated round 3 per red R3-10, having ridden silently from the round-0 construction through the
round-2 recompute)** = **~$18/run (~10% of the rescaled run-4 baseline)**".

**Problem:** a throttle sets round N's spend from the board it can read — the post-round-(N−1)
board. Throttling rounds 3–5 therefore requires the post-round-2 (~65), post-round-3 (~44), and
post-round-4 (~30) boards to trip the threshold. §2.1's own item 1 names only post-round-3 and
post-round-4 as "the two lowest-mass boards among boards that preceded another round";
post-round-2's ~65 is the run's second-HIGHEST mass (66% of peak). Calling rounds 3–5 "the
low-mass rounds" reads each round's OWN post-merge mass (44/30/31) as its throttle input — a
lookahead the mechanism cannot perform — or silently admits a mid-mass board as "low-mass."
Only rounds 4–5 are clearly throttled by §2.1's series as printed; round 3 is threshold-dependent
at best. Honest figure: **$12–18/run (~7–10%)**, with $18 the ceiling (fires only if the
threshold admits mass 65), not the point estimate.

**Vector honesty, logged against red:** the ×3 multiplier originated in red's own R2-2 rescale
arithmetic (~$6/round × 3 = $18), carried through the round-3 ledger's arithmetic check (l.163)
which verified the multiplication and flagged only that the basis was unstated (L2-F2 → R3-10).
Blue stated the basis as instructed — and the stated basis exposes the timing defect. Error
still an error; the repair is audited as a new claim (repair-regression class, red-vector
sub-class).

**Impact bound:** blue's §2 REJECT is unaffected — §2.4 itself notes the figure *strengthens*
the throttle case blue rejects, and the REJECT rests on §2.1/§2.3. The damage is
record-accuracy: ~$18/~10% is a headline planning figure echoed into §6.1's baseline framing.

**Required fix (trivial):** one sentence in §2.4 stating throttle-input timing (boards after
rounds 2–4 gate rounds 3–5's spend; post-round-2 mass ~65 is mid-band) and restating the
saving as $12–18/run with $18 conditional on a threshold that fires at mass 65 — or name a
concrete threshold and show which rounds it fires on.

## Round-3 repairs in slice — verified clean

- **R3-6 (§1.5 registered prediction):** composes with the hardened arm — prediction (i) = arm
  (b) (two consecutive zero-above-floor-mint rounds), (ii) is STRICTER than arm (a) (whole open
  board vs unadjudicated subset — settling TRUE implies the arm fired; correct test direction),
  (iii) = arm (c); single-round total-mint trigger explicitly disavowed; netting arithmetic
  consistent ($25–30 − ~$10 ≈ $15–20), matching §1.2's R2-10 netting. The sibling-repair
  composition defect does not recur in the restatement. HIGH.
- **R3-6's provenance sentence** ("the round-2 closure ledger records both as closed WITH
  REGRESSION"): corroborated first-hand — this run's red/findings.md ll.11/335/412 record the
  round-3 AMENDMENT reclassifying R1-12/R1-17 from closed-clean to closed-with-regression in
  the round-2 ledger; the report's parenthetical names round 3 for the restatement, so the
  timeline is not misstated. HIGH.
- **R3-7/R3-8 (§2.5 item 1 + carried design):** presence-check honesty stated; mass/board
  recompute clause names a non-red-merge executor and mirrors the found_by clause; mapping
  pinned-before-first-logged-round with version stamps marking (not preventing) breaks;
  new-series rule stated; §8 Q6 DECIDED text consistent at both sites (realized=excluded;
  mapping enumerated low=1 … certain=3.5); §2.1's historical realized=3.5 series ring-fenced
  as non-comparable. No false-equivalence residue. HIGH.
- **R3-14 (§2.1 item 4):** band recomputed exactly from the printed tables — 98/20=4.9,
  65/11=5.9, 44/10=4.4, 30/5=6.0, 31/6=5.2; "two highest are rounds 2 and 4" holds; the
  certain×low=3.5 vs medium×medium=4 example is faithful to the disclosed mapping. HIGH.

## Staleness discharge (ledger rule: >2 rounds since last verification)

- Pin equivalence re-run this round: both `git diff --stat` empty (bfa8a3b;
  5396952). HIGH.
- §1.1 board/grade quotes re-read at the pinned lines: R5-1 (l.135, certain×medium, three
  lenses), R5-5 (l.200, MEDIUM-HIGH medium×high, telemetry-invisible), R4-1 (l.425,
  certain×high, four of five lenses), R3-1/R3-2 (ll.717/737, both MEDIUM-HIGH code-trace,
  complexity low), R2-1/R2-3/R2-7/R2-8/R2-9 (ll.1080–1181, all MEDIUM-HIGH, all complexity
  low — §1.2's premise holds). Backlog item 30 present at l.30. All verbatim-consistent with
  the round-1 ledger entries. HIGH.
- [^DebateRounds] re-fetched (arXiv /html/2506.00066v1): saturation quotes at three/four/five
  rounds ("performance improvements saturate around three rounds"; "accuracy plateaus around
  four rounds"; "performance saturation after five rounds") and decline after round 2 ("only
  until the second round, after which accuracy declines") — §1.3's "~2–5 rounds" and the
  footnote's "degradation past round 2 on some tasks" both hold. HIGH.
- [^CvssInconsistent] re-fetched (arXiv:2308.15259 abs): "for the same vulnerabilities from
  the main study, 68% of these users gave different severity ratings" — 59 participants in the
  follow-up. Verbatim. HIGH.
- [^Stads] re-fetched (arXiv:1803.02130 abs): residual-risk-from-discovery framing +
  ecological-biostatistics basis confirmed at the abstract; Good-Turing/singleton specifics
  are body-level. Stays MEDIUM-HIGH exactly as self-labeled — no gap.
- [^RbtTaxonomy] leaf-resolved for the FIRST time (never fetched in rounds 1–3, accepted
  as-labeled at R1): arXiv:1912.11519 IS "A taxonomy of risk-based testing," Felderer &
  Schieferdecker — identity HIGH; the subjective-expert-opinion/triangulation gloss remains
  MEDIUM as the report self-grades it. Citation identity confirmed; no misattribution.

## Carried without re-fetch (in-window per ledger rule)

§1.2 arithmetic ($78/$68 net, 21 mints = 10+5+6, series 52 = 31+15+6) — R2/R3 recomputed;
[^JournalCheck] — R3 first-hand ×2; [^AdaptiveStability]/[^Sprt] — R3 re-fetched ×2;
[^CathedralBazaar] — R2 ×3; [^CaptureRecaptureEval]/[^CaptureRecaptureDecade] — R2 substance;
[^Iso29119] — R2; Fragmentation-403 access-failure — R2 reproduced; §0 rows 1–2/3a — R3
propagation verified; [^ExpertCvss]/[^FentonOhlsson] — access-blocked, carried as-labeled
MEDIUM/MEDIUM-HIGH per §7's disclosed limits.

## Friction

None this pass — memory readable, both pins clean, all four external fetches succeeded on the
first route (abs or /html), no write-block firing on this filename.
