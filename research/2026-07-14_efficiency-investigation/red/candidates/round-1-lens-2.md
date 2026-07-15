# Red round 1 — lens 2 of 4: leaf-node citation verification, slice 2 (§2 risk-mass-proportional spend; §3 grade-dispute channel + best-of-N)

Lens seat: findings carry lens-scoped ids (L2-F*) only; stable R1-N ids are the merge's to assign.
Access date for every check below: 2026-07-14. Pin re-verified at this seat: `git diff --stat
bfa8a3b HEAD -- research/2026-07-12_feov-retrospective/` and `git diff --stat 5396952 HEAD --
ideas/backlog.md plugins/frank-exchange-of-views/` both empty.

## Verdict contribution: §2 and §3 are substantially sound at the leaf node

The load-bearing corpus claims in both sections verified first-hand at the pin, and the four
load-bearing external panel/debate sources verified verbatim against their abstracts (one via
full-text HTML). The §2.5/§3.5 dispositions' evidentiary bases hold. Five graded defects below,
all LOW or LOW-MEDIUM — none disturbs a disposition; all are trivial fixes.

## Verified at HIGH confidence (summary; full lines appended to red/citation-ledger.md)

- **Backlog item 30** carries verbatim: the severity-floor spec + "$10" claim, the risk-mass
  umbrella + "spot-check floor never reaches zero" caveat, the `grade_disputes:
  [{gap_id, dimension, proposed, evidence}]` schema, and best-of-N deferred until "runs 4-5 show
  lone-voice bias survives (needs per-gap records to grade blind from)". [^BacklogLevers] HIGH.
- **debate.js @ 5396952**: `citationPasses = Math.min(4, Math.max(1, Math.ceil((blueEnv.claim_count
  || 20) / 40)))` inside the round loop (recomputed every round — §2.3's quote omits only the
  `|| 20` fallback); judge resolution enum exactly `closed | rebuttal_sustained | risk_accepted |
  carried | unresolved` — §3.3's "cannot express 'gap real, grade wrong'" is correct; contested
  window is the whole debate (`allPriorGapIds` accumulates; supersedes-descendant match), so
  §3.1's "any gap open across two rounds auto-dockets" is correct. [^EngineSource] HIGH.
- **Zero judge dispatches, reproduced live**: anchored `grep "^### "` on run-3 debate.md = 11
  headers (6 BLUE, 5 RED, 0 LEAD); unanchored `grep -c "### LEAD"` = 5 — both the fact and the
  footgun reproduce exactly as §3.1 states. [^DebateNoLead] HIGH.
- **Grade headers verbatim** in run-3 findings.md: R4-1 (HIGH, certain × high × low-medium,
  "four of five lenses converged independently"), R5-5 (MEDIUM-HIGH, medium × high,
  telemetry-invisible clause), R5-1 (certain × medium, "three lenses converged independently —
  lenses 1, 2, 4"). [^R4OneDetail][^R5FiveDetail][^R5OneDetail] HIGH.
- **R5-5 singleton evidence** checked in the round-5 lens files: lens 5 carries the enforcement
  argument; lens 4's "Considered raising this... Not raised" sits at lines 115–120 (footnote says
  ~110–120 — within tolerance); lens 2's supersedes mentions are detector-mechanics only, no
  enforcement raise. [^R5FiveSingleton] HIGH.
- **R4-1 4-of-5 convergence**: lineage-blind/docket content dense in round-4 lenses 1/2/3/5;
  lens 2's header "Finding 1 (NEW, round 4) — the contested-docket detector is lineage-blind by
  construction" verbatim at its line 19; lens 4's three "docket" hits are unrelated framing
  checks, not a mint. [^Round4Lenses] HIGH.
- **Run-3 citation-ledger lines ~184–185**: both merge-seat overrules present as described
  (lens 2's six-id claim overruled by mechanical extraction — five ids; lens 5's no-discrepancy
  hold overruled by direct read of report lines 496 + 727). [^CitationLedgerRun3] HIGH.
- **cost.md figures**: red-lens $9.28/$9.22/$9.46/$10.47/$11.05, Σ rounds 1–5 = $49.48 exactly
  (§2.4's 33% of $149.95 ✓; excludes the killed round-6 $0.61 — defensible, killed mid-spawn);
  blue-respond $3.95/$3.96/$2.98/$3.05/$4.27 exact; red-merge $7.52/$13.22/$12.64/$10.60/$13.56
  exact; "rounds 3-5 closed ~15 mostly-trivial gaps" verbatim. [^CostAudit] HIGH.
- **§2.1 mass series spot-recompute**: round-5 board under the disclosed mapping = R5-1
  (3.5×2=7) + R5-2 (3×2=6) + R5-3 (3.5×1=3.5) + R5-4 (3.5×1=3.5) + R5-5 (2×3=6) + R5-6
  (2×2.5=5) = **31.0 exactly** — brackets lane 1's ~31 and lane 3's ~32; the disclosed-mapping
  reproducibility claim holds on the round I recomputed. "Round 5 rose" confirmed. HIGH on the
  recomputed round; the other four rounds accepted on two-lane independent agreement (not
  re-derived here).
- **Debate quotes verbatim**: round-4 BLUE "No rebuttals this round — every gap was real, at the
  location red found it, and none was over-graded relative to its fix cost"; round-5 BLUE
  concedes/builds all six of R5-1..R5-6 with zero grade contests; round-2 RED "Red will accept
  argued risk-accepts on R2-9 and R2-10 if reasoned". [^BlueRound4][^DebateNegotiation] HIGH.
- **Run-3 report §3 rows 15/23**: R2-1 count 3→2 with High retained by the "2-for-2 rate on the
  triggering conditions" argument; R3-7 mechanism narrowed, grade kept; R5-1's enumeration
  corrected with grades untouched. §2.2's characterization is faithful. [^Retro3Docket] HIGH.
- **External, fetched at the leaf**: NineJudges (arXiv:2605.29800) — all four quoted results
  ("about 2 independent votes' worth", "8-22 percentage points short", "best single judge matches
  or outperforms the full panel", "at most 11% of this gap") verbatim in the abstract. PoLL
  (arXiv:2404.18796) — "outperforms a single large judge, exhibits less intra-model bias due to
  its composition of disjoint model families... over seven times less expensive" verbatim.
  CvssInconsistent (arXiv:2308.15259) — "59 participants... 68% of these users gave different
  severity ratings" verbatim. PersuasiveDebate (arXiv:2402.06782) — "76% and 88%... (naive
  baselines obtain 48% and 60%)" verbatim. All HIGH.
- **Capture-recapture few-reviewers caveat** (feeds §2.3's "statistically forced hedge"):
  IEEE Xplore returned a blank page for Briand et al., so the verbatim quote in
  [^CaptureRecaptureEval] is MEDIUM on quote-fidelity — but the substance verified HIGH via the
  Petersson JSS 2004 PDF (fetched, extracted with pdftotext): "at least four to five reviewers
  should participate in order to make the accuracy acceptable (Briand et al., 2000)"; "most
  models underestimate"; Mh-JK "non-robust in the case with two reviewers and produces
  underestimates." The §2.3 argument stands on corroborated substance.
- **Iso29119 companion** (arXiv:1905.10676) exists and is on-topic (title verbatim). HIGH for
  the taxonomy citation; the 29119 standard itself was not fetched (paywalled standard) — its
  "risk-based allocation is normative" gloss rests on the taxonomy paper + general knowledge:
  MEDIUM, not load-bearing (one clause in §2.3's pro-throttle side, which blue rejects anyway).
- **Frontier H3 and red-auditor.md line 13**: both match the report's characterizations. HIGH.
- **Accepted as-labeled without re-fetch**: [^ExpertCvss] (genuinely paywalled ScienceDirect;
  blue self-graded MEDIUM, honest), [^RbtTaxonomy], [^DalalMallows] and [^FentonOhlsson]
  (classical results, characterizations match the known literature — MEDIUM-HIGH; neither is
  load-bearing alone for any disposition in my slice).

## Findings (graded; every one anchored)

### L2-F1 — [^ConflictingScores] is a source misattribution, and §7's excuse for not verifying it is false — LOW-MEDIUM
**Location:** §2.2 ("The throttle input is noisy in the general severity-grading literature...")
— *"NVD-vs-CNA disagreement on roughly a third of dual-assessed CVEs"* — and §7 — *"lane 2's
~34% NVD-vs-CNA figure and the expert-CVSS moments are from search digests, not leaf-verified
(paywalled; not load-bearing...)"*.
**Problem:** I fetched the FULL text (arXiv HTML, 2508.13644v1 — the paper is open-access, not
paywalled). It contains **no NVD-vs-CNA disagreement rate and no ~34% figure anywhere**: it
compares four scoring SYSTEMS (CVSS/EPSS/SSVC/Exploitability Index) on the same Microsoft Patch
Tuesday CVEs — inter-system disagreement, not intra-system assessor disagreement. The claim is
not merely unverified; it is affirmatively outside the cited paper's scope. The real NVD-vs-CNA
~34% figure exists in the literature but lives in a different study. Two defects: (a) real
figure, wrong source (the recurring misattribution class); (b) §7's "paywalled" label is
factually wrong for this source — the verification blue skipped was available.
**Grading:** certain (text vs fetched full text) × low-medium (blue already flags the figure
non-load-bearing and the 68% figure carries §2.2's point; but a wrong-source citation surviving
inside a self-audited "known verification limits" section is a process defect, and §2.2's
"roughly a third" clause currently reads as sourced fact) × trivial (drop the NVD-vs-CNA clause,
or re-source it to the actual study; correct §7's "paywalled" to "not checked").
**Required fix:** remove or re-source the NVD-vs-CNA claim; amend §7's parenthetical — the
excuse must match reality ("figure from a search digest; actual source not identified" — NOT
"paywalled").

### L2-F2 — [^WeakJudges] footnote gloss reassigns the source's condition — LOW
**Location:** Footnotes — *"debate gains over consultancy are task-dependent and smaller than
prior studies"* (vs §3.6 body: *"gains are task-dependent and shrink in some regimes"*).
**Problem:** the abstract (fetched) says debate beats consultancy **across all tested
scenarios**; what is task-dependent is debate vs **direct question answering** ("in extractive
QA tasks with information asymmetry debate outperforms direct question answering... in other
tasks without information asymmetry the results are mixed"), and "more modestly than in previous
studies" attaches to the stronger-debater effect. The §3.6 body sentence is compatible with the
source; the footnote's specific "over consultancy" attribution is not. Within-source condition
misattribution, footnote-only.
**Grading:** certain × low (body text and the §3.6 disposition are unaffected; the footnote is
the reference layer future rounds will trust) × trivial (reword the gloss: "debate beats
consultancy across tasks; gains over direct QA are task-dependent; stronger-debater gains more
modest than prior studies").

### L2-F3 — §2.3's "blue-respond spend already tracks board size naturally" is under-supported by its own cited series — LOW
**Location:** §2.3, final bullet — *"blue-respond spend already tracks board size naturally
($3.95/$3.96/$2.98/$3.05/$4.27 across rounds)"*.
**Problem:** figures verified exact against cost.md — but they do not show natural
board-tracking: round 5 is the run's HIGHEST blue-respond spend ($4.27) on the second-smallest
board (6 open), while round 1 handled 20 gaps for $3.95. Boards 20/11/10/5/6 against spends
3.95/3.96/2.98/3.05/4.27 is a weak, round-5-broken correlation (round 5's spend was fix-weight
— building R5-5's structural check — not board-count). The bullet's first half (citationPasses
scales with claim_count) is verified and carries the "correct proportionality already ships"
point alone.
**Grading:** certain × low (one supporting clause in a section whose disposition rests on §2.1's
verified counterexamples) × trivial (soften to "blue-respond spend is driven by fix weight, not
inflated by board size" or delete the clause).

### L2-F4 — §2.1's "two lowest-mass boards" contradicts its own table on a strict read — LOW
**Location:** §2.1 item 1 — *"The two lowest-mass boards (post-round-3, post-round-4) preceded
the run's highest-graded discovery..."*.
**Problem:** per the section's own table the two lowest masses overall are post-round-4 (~30)
and post-round-5 (~31/32); post-round-3 is ~44. The intended universe — boards a throttle would
have acted on, i.e. boards with a successor round — is named later in the sentence ("at exactly
the boundaries a throttle would have acted on"), under which the claim is true. As written, the
noun phrase and the table disagree.
**Grading:** certain × low (the argument is right; the sentence is loose) × trivial ("the two
lowest-mass boards that preceded another round").

### L2-F5 — §3.6's judgment-merge cost band starts at the wrong end — LOW
**Location:** §3.6 — *"a judgment-seat cost multiple ($10.60–$13.56/round is what the one
existing judgment merge costs)"*.
**Problem:** the verified red-merge series is $7.52/$13.22/$12.64/$10.60/$13.56 — the full-run
range is $7.52–$13.56. The quoted band is rounds 2–5 only; round 1's $7.52 is excluded without a
stated reason. Direction of error inflates the anti-panel cost argument slightly; the rejection
has three independent verified legs (unmet precondition, PoLL configuration mismatch, NineJudges)
so nothing turns on it.
**Grading:** certain × low × trivial ("$7.52–$13.56/round" or "typically $10–14/round").

### L2-F6 — [^Sprt]'s "30–50% typical" band does not match the cited paper's band — LOW
**Location:** §2.3 — *"sequential-adaptive spend buys 30–50% savings at unchanged error rates
when the statistic is right (Wald)"* and the footnote's *"Savings figures (30–50% typical) per
'The relative efficiency of sequential tests' (arXiv:2603.00216)..."*.
**Problem:** the fetched paper states the sequential test "reduces the average sample size by at
least 36% and by at most 75%" — a 36–75% band, not 30–50%. The report's figure may descend from
the footnote's second, vaguer source (the Springer introduction to Wald), which I could not pin.
Direction is conservative at the low end (claim understates the source), but the number as
cited is pinned to a source that says something else.
**Grading:** certain × low (the claim supports the pro-throttle side blue rejects; understated,
not inflated) × trivial (either quote 36–75% from the arXiv paper, or attribute 30–50% to the
Wald introduction alone).

## Non-findings, recorded for the merge

- §2.2's "measured-robust in this loop" claim (grade corrections moved mass ~0) is consistent
  with row 15/row 23 as verified: those corrections changed counts and mechanisms in blue's
  docket cells while likelihood grades were retained by argument — no mass-feeding red grade
  moved. Corroboration MEDIUM-HIGH (verified for the three named chains; not exhaustively for
  all corrections).
- §3.1's "blue never disputed a red gap grade" holds on sampling: round-4 BLUE's no-rebuttals
  statement and round-5's six concessions verified verbatim; rounds 1–3 rebuttals sampled (R1-13
  "partially rebutted on urgency", R1-18 "conceded on evidentiary weight") are dispositional
  contests, not likelihood×impact grade contests. MEDIUM-HIGH — exhaustive check would require
  re-reading all five BLUE sections clause by clause; recommend the merge treat as corroborated
  unless a sibling lens hit a counterexample.
- The §2.3 citationPasses quote omits the `|| 20` claim_count fallback — cosmetic, not raised.

## Friction

- PDF MCP tools (`pdf-reader`, `arxiv-latex`) were not exposed at this lens seat (no ToolSearch
  available); the Petersson PDF verification fell back to WebFetch's saved binary + mingw
  `pdftotext`, which worked — but the protocol's "try the document-extraction MCPs before
  grading down" step is unfulfillable as tooled at a lens seat.
- IEEE Xplore returns an empty page to WebFetch (Briand et al. abstract unverifiable at the
  publisher) — recurring paywalled-verification gap; the Read tool's PDF page-rendering path is
  also broken on this box (`pdftoppm is not installed`), so saved-PDF reading needs the
  text-extraction detour every time.
