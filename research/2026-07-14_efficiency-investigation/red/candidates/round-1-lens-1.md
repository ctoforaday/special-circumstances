# red round-1 lens-1 — leaf-node citation verification, slice 1 of 4

Scope: preamble + §0 + §1 (severity-floor termination) of `blue/report.md` (round-0 synthesis),
plus every footnote those sections cite ([^PinCheck] [^BacklogLevers] [^FindingsBoard]
[^Round3Red] [^R4OneDetail] [^R5FiveDetail] [^AlreadyShipped] [^CostAudit] [^StopResume]
[^AdaptiveStability] [^DebateRounds] [^DalalMallows] [^Stads] [^CaptureRecaptureEval]
[^CaptureRecaptureDecade] [^CeilingDisposition]). Full report re-read in context (914 lines,
two windows). Ledger empty at start (round 1).

## Method

Every corpus citation re-verified first-hand against the working tree after re-running the pin
check myself (`git diff --stat bfa8a3b HEAD -- research/2026-07-12_feov-retrospective/` and
`git diff --stat 5396952 HEAD -- ideas/backlog.md plugins/frank-exchange-of-views/` — both
empty, so working-tree reads are pin-faithful). External citations leaf-fetched (arXiv HTML,
secondary abstract expositions where primaries are paywalled).

## Verified HIGH (statement ↔ reference corroboration)

- **Preamble / [^PinCheck]** — pin equivalence re-run first-hand; both diffs empty. HIGH.
- **§1.1 backlog quote / [^BacklogLevers]** — "when every open gap is <= MEDIUM with trivial
  fix cost... (would have ended run 3 at round 3 for ~$10)" is verbatim at `ideas/backlog.md`
  line 30. HIGH.
- **§1.1 board table / [^FindingsBoard]** — recounted from `red/findings.md` (1364 lines,
  matching the footnote): round 1 = R1-1..R1-20 (20 open), R1-1 and R1-2 both HIGH (lines
  1281, 1287); round 2 = R2-1..R2-11 (11 open), MEDIUM-HIGH members exactly {R2-1, R2-3, R2-7,
  R2-8, R2-9} (lines 1080, 1112, 1156, 1170, 1181); round 3 = R3-1..R3-10 (10 open), R3-1 and
  R3-2 both MEDIUM-HIGH, both code-trace, both complexity "low" (lines 717, 737); round 4 =
  R4-1..R4-5 (5 open), R4-1 HIGH (line 425); round 5 = R5-1..R5-6 (6 open), R5-5 MEDIUM-HIGH
  (line 200). Zero fires at "≤ MEDIUM + trivial" — confirmed at every boundary. HIGH.
- **§1.1 / [^Round3Red]** — "round 3: 2 MEDIUM-HIGH, both code-trace — every prose gap is now
  ≤ MEDIUM" verbatim at `debate.md` lines 491–492, inside the round-3 `### RED` section
  (headers at 43/225/379/568/716 — line 491 falls under 379). HIGH.
- **§1.2 / [^R4OneDetail]** — R4-1 "HIGH — certain (already realized in this corpus, not
  projected) x high x low-medium... four of five lenses converged independently" verbatim,
  findings.md line 425. "Most consequential engine finding" matches run-3 report TL;DR
  ("single most consequential finding"). "PR #15's core" matches `inputs/already-shipped.md`
  (lineage-aware contested docket headlines PR #15). HIGH.
- **§1.2 / [^R5FiveDetail]** — R5-5 "MEDIUM-HIGH — medium... x high (telemetry-invisible: an
  unset or vacuous supersedes leaves contested.length at 0..." verbatim, findings.md line 200;
  "PR #15's structural enforcement throw" matches already-shipped.md. HIGH.
- **§1.2 ~$53 / [^CostAudit]** — recomputed from cost.md's table: round 4 = 3.05 + 10.47 +
  10.60 = 24.12; round 5 = 4.27 + 11.05 + 13.56 = 28.88; sum = $53.00. Also re-verified:
  red-merge 7.52/13.22/12.64/10.60/13.56 (Σ 57.54 = 38% of 149.95), red-lens Σ 49.48 (rounds
  1–5, r6's $0.61 correctly excluded), blue-respond 3.95/3.96/2.98/3.05/4.27 (Σ 18.21),
  "rounds 1–2 closed 31 gaps ($60-ish); rounds 3–5 closed ~15 mostly-trivial gaps" verbatim
  (finding 4). HIGH.
- **§1.2 new-mint counts (20/11/10/5/6)** — id-range recount matches. HIGH.
- **§1.2 / [^StopResume]** — cost.md finding 5 verbatim ("cost ~$0 and cut ~7 residual rounds;
  five round-6 lenses were killed mid-spawn for pennies"). R5-5 OPEN MEDIUM-HIGH at
  termination confirmed (findings.md line 200 + verdict line 77 "FAIL"). Consistency check
  passed: run-3 report says "terminated by the safety ceiling"; the ceiling that fired was the
  operator-reduced maxRounds — the two accounts are one event. HIGH.
- **§1.3 "FAIL, UNVERIFIED, 6 open"** — findings.md line 77 (FAIL), run-3 report.md line 3
  (UNVERIFIED), R5-1..R5-6 open. HIGH.
- **§1.3 / [^AdaptiveStability] body claim** — arXiv 2510.12697v1 leaf-fetched: criterion is
  KS statistic < 0.05 for 2 consecutive rounds ("halts once D_t < 0.05 for 2 consecutive
  rounds"), a distributional-stability property of the process. Blue's characterization
  ("stability of the discovery process... double confirmation, not a property of the residual
  list") corroborates. HIGH for the body sentence; see L1-F2 for the footnote gloss.
- **§1.3 / [^DebateRounds]** — arXiv 2506.00066v1 leaf-fetched: saturation ~2–5 rounds
  (Xu et al. decline after round 2; Chen et al. ~3; Liu/Li peak ~4), degradation documented.
  HIGH.
- **§1.3 / [^DalalMallows]** — primary paywalled (tandfonline 403); bibliographic identity
  (JASA 83(403):872–879, 1988) and the asymptotic rule confirmed via a detailed secondary
  exposition (Höhle 2016): stop when f/c · (e^{μt}−1)/μ ≥ k, i.e. observed discovery count
  against the cost ratio — exactly blue's phrasing. HIGH (claim), with the caveat the check
  ran through a secondary exposition.
- **§1.3 + §2.3 quote / [^CaptureRecaptureEval]** — the exact sentence "when the number of
  inspectors is too small, no model is sufficiently accurate and underestimation may be
  substantial" confirmed against the Briand et al. IEEE TSE abstract (ieeexplore 852741) via
  search-verified quote match; adjacent context ("at least four to five reviewers") also
  supports §2.3's use. HIGH.
- **§1.4 / [^CeilingDisposition]** — "Outstanding gaps at the ceiling — disposition and
  compromise rationale" exists at run-3 report.md lines 7–20 with per-gap grading / blue
  response / disposition columns + compromise-rationale paragraph, as claimed. HIGH.
- **§1.5 ratify-1 "still-unshipped" heartbeat** — backlog line 31 lists
  "log()-per-transition heartbeats" under STILL OPEN. Substance HIGH; attribution see L1-F3.
- **[^BacklogLevers]'s "conceded an error" quote** — lives in the docket-detector item
  (backlog line 29), and the footnote correctly attributes it there. HIGH.

## Verified MEDIUM (needs more evidence, not failure)

- **§1.3 / [^Stads]** — arXiv abs page confirms residual-risk-from-discovery-curve framing;
  the specific Good-Turing/singleton sentence corroborated only via search digest (FSE21 PDF
  fetch lossy; arxiv-latex / pdf-reader MCPs not reachable from this lens seat — see
  friction). The claim matches the well-documented content of STADS. MEDIUM-HIGH.
- **§1.3 / [^CaptureRecaptureDecade]** — wohlin.eu/jss04-1.pdf fetch lossy; bibliographic
  identity confirmed via search listing only. Non-load-bearing in §1.3 (the Briand quote
  carries the point). MEDIUM.

## Findings (lens-scoped ids)

### L1-F1 — §1.2 firing-round arithmetic: the relaxed floor fires after round 2, not round 3

- **Location:** §1.2 "Convergent: making it fire makes it wrong" — "The only threshold that
  realizes the claimed saving admits MEDIUM-HIGH — and at that setting the floor fires after
  round 3 and terminates before round 4 minted R4-1... Deleted value for a saving of ~$53
  (rounds 4–5 seat-round sum from cost.md)."
- **Problem:** by the pinned grades blue itself tabulates, a floor admitting MEDIUM-HIGH first
  fires at the **round-2** boundary: the round-2 board's five MEDIUM-HIGH members (R2-1, R2-3,
  R2-7, R2-8, R2-9) all carry complexity "low" — indistinguishable, under the floor's
  severity + fix-cost terms, from round 3's two MEDIUM-HIGH/low gaps. Verified first-hand at
  findings.md lines 1080/1112/1156/1170/1181 vs 717/737. No severity/complexity threshold
  exists that fires after round 3 but not after round 2. Consequences: (i) the relaxed floor
  deletes rounds 3–5 (~$78 by the same cost.md table), not rounds 4–5 (~$53), and additionally
  deletes R3-1/R3-2's discovery; (ii) the sentence's premise — that some threshold "realizes
  the claimed saving" (a round-3 stop) — is false: no setting reproduces the backlog's round-3
  stop at all, which *strengthens* §1.1's headline contradiction of the backlog.
- **Direction:** the error is conservative — corrected, it makes blue's REJECT stronger. But a
  verdict-supporting counterfactual with a wrong firing round and a wrong dollar figure cannot
  ship as-is.
- **Grade:** LOW-MEDIUM — certain (mechanical from pinned grades) × low (conclusion
  unchanged; the numbers and the "only threshold" premise are wrong) × trivial (rewrite one
  passage: "first fires after round 2, deleting rounds 3–5, ~$78 and four findings including
  R3-1/R3-2; no setting reproduces the backlog's claimed round-3 stop").
- **Corroboration:** HIGH (two findings.md blocks read side by side; cost.md recomputed).

### L1-F2 — [^AdaptiveStability] footnote gloss overstates the adaptive-stopping result

- **Location:** Footnotes — "[^AdaptiveStability]: ...adaptive stops at rounds 2–8 lose <1%
  accuracy vs fixed 10 rounds."
- **Problem:** leaf fetch of arXiv 2510.12697v1: reported stops fall at rounds ~4–7 in the
  paper's examples, and at least one reported delta is −1.03% (BIG-Bench/Gemma-3-4B, 70.07%
  vs 71.10%) — breaching "<1%". "Rounds 2–8" and "<1%" are not the fetched table's numbers.
  The §1.3 body sentence this footnote supports (criterion = distributional stability, double
  confirmation) verifies HIGH and is unaffected.
- **Grade:** LOW — certain × low (footnote-gloss precision; body claim intact) × trivial
  (soften to "stops at rounds ~4–7 losing ≈1% or less in the reported configurations").
- **Corroboration:** HIGH for the discrepancy (direct fetch); the full table may contain
  stops outside 4–7 — if blue re-fetches and pins stops at 2 and 8 with all deltas <1%, the
  finding closes as rebutted.

### L1-F3 — [^BacklogLevers] over-attributes the heartbeat item to backlog item 30

- **Location:** Footnotes — "[^BacklogLevers]: 'run-3 termination & fairness levers,'
  `ideas/backlog.md` item 30 @ 5396952 — severity-floor spec..., log()-heartbeat item; ..."
- **Problem:** the log()-per-transition heartbeat item lives in the *adjacent* backlog item
  (line 31, "run-3 friction paper cuts," STILL OPEN list), not in item 30. Everything §1.5
  claims about it ("still-unshipped," composes with the board-profile log()) is true at the
  real location; only the footnote's bundling is wrong. Known pattern: footnote
  over-attribution — one label wearing a crowd of specifics.
- **Grade:** LOW — certain × low (navigability of the citation, not truth of the claim) ×
  trivial (split the footnote or name line 31).
- **Corroboration:** HIGH (both backlog lines read first-hand at the pin).

### L1-F4 — two §1.3 externals not leaf-verifiable from this seat (verification limit, not defect)

- **Location:** §1.3 — "Böhme's STADS estimates residual risk from the discovery curve itself
  (Good-Turing: the singleton rate...)" and "capture-recapture... [^CaptureRecaptureDecade]".
- **Problem:** both primaries are PDF-only from this seat; WebFetch returned lossy binary for
  FSE21.pdf and jss04-1.pdf; the protocol's mandated arxiv-latex / pdf-reader MCP fallback is
  not reachable from this lens seat (no ToolSearch/MCP tools in the seat's toolset). Claims
  corroborated at MEDIUM(-HIGH) via abstracts + search digests; both are secondary supports
  (Dalal-Mallows and Briand carry §1.3's load at HIGH).
- **Grade:** LOW — medium (drift/misquote risk on unverified specifics) × low
  (non-load-bearing) × low (one MCP-equipped pass closes it). Flagged so the merge can route
  the two PDFs to a seat with the extraction tools rather than grading blue down.
- **Corroboration:** n/a (this is a verification-coverage note).

## Out-of-slice observations (for the merge to route, not slice-1 findings)

1. **§4.2 (slice 3): "5× cache-read, 12.5× cache-write" repeats a rate-vs-ratio conflation in
   the pinned cost.md finding 3.** cost.md's own header prices sonnet cache-write at 2.5 and
   judgment-tier at 12.5 $/MTok — the *multiplier* is 5×, not 12.5×; 12.5 is the absolute
   rate. Blue's §6.4 catches cost.md finding 2's internal contradiction but inherits finding
   3's. Route to instance 3.
2. Incidentally re-confirmed while in the sources: debate.md anchored header count = 11
   (6 BLUE / 5 RED / 0 LEAD) supporting §3.1 (slice 2); findings.md = 106,772 bytes and
   blue/report.md = 159,394 bytes supporting [^FrictionFifteen] (slice 3).

## Slice verdict

FAIL (four open findings, none above LOW-MEDIUM; L1-F1 needs a rewrite, L1-F2/F3 trivial
footnote repairs, L1-F4 a tooled re-check). §1's core argument — never-fires table, backlog
contradiction, cost arithmetic, the stopping-statistic literature — verifies at HIGH
throughout; the defects found are in one counterfactual's numbers and two footnote glosses,
not in the disposition.
