# Round 2 — lens 4 (leaf-node citation verification), slice 4: §6 Cross-cutting + §7 Pre-flight self-audit + §8 Open questions + Footnotes

Full living report re-read whole (three windowed reads forced by the 25k Read cap — friction, see below).
Ledger honored: claims verified HIGH in round 1 and untouched by round-1 edits were not re-fetched; every
round-1 repair in this slice was re-verified as a new claim (repair-regression discipline). Pin equivalence
re-run first-hand this round: both `git diff --stat` checks empty (retrospective @ bfa8a3b; backlog + plugin
@ 5396952).

## Findings

### L4-F1 — [^AdaptiveStability] repair-regression: the round-1 "~4–7" stopping band is contradicted by the source's own table (and the correction record's "matched neither" is half-false)

- **Location:** Footnotes, `[^AdaptiveStability]`: "adaptive stops fall at rounds ~4–7 losing ≈1% or less
  accuracy in the reported configurations vs fixed 10 rounds (softened round 1 per red R1-20's leaf
  re-fetch: at least one reported delta is −1.03%, and the round-0 'rounds 2-8 / <1%' gloss matched
  neither)."
- **Leaf verification (two independent fetches of arxiv.org/html/2510.12697v1, second targeted at Table 2):**
  stopping rounds per configuration span **2 to 8** — JudgeAnything stops at round 2 for Gemma-3-4B and
  Qwen-2.5-VL-7B and at round **8** for Gemini-2.0-Flash; the remaining 19 configurations fall at 4–7.
  So: (a) "stops fall at rounds ~4–7 ... in the reported configurations" is false as a universal — three
  of 22 reported configurations fall outside the band, at both ends; (b) the correction record's claim
  that the round-0 "rounds 2-8" gloss "matched neither" is itself wrong on the rounds half — **2–8 is
  exactly the table's min–max**; the only half that failed was "<1%" (max loss −1.03%, BIG-Bench/Gemma,
  confirmed verbatim). The honest phrasing: "typically 4–7, full range 2–8, losing ≈1% or less (max
  −1.03%)."
- **Lineage hint for the merge:** this defect was *introduced by the round-1 repair* — the CHANGELOG
  attributes the "~4–7" phrasing to red R1-20's own leaf re-fetch, and blue copied it verbatim. If the
  merge treats R1-20 as closed, this is a closure WITH REGRESSION; the successor should name R1-20 in
  `supersedes`. (Red's own required-fix phrasing was the vector — the pattern memory
  `pattern_repair_regression_citation` names exactly this.)
- **Grade:** likelihood **certain** (static text vs. fetched table) × impact **low** (the load-bearing §1.3
  body claim — KS < 0.05 for 2 consecutive rounds as the criterion — is verified HIGH and unaffected;
  this is footnote-fidelity plus a false statement inside a correction record, which is reputationally
  worse than an ordinary gloss error in a report whose §7 promises "labeled not laundered") × complexity
  **trivial** (one clause).
- **Corroboration confidence for the statement↔reference pair as written: LOW** (band contradicted;
  delta figure confirmed).

### L4-F2 — [^Sprt] drops the source's scoping condition on the 36–75% band

- **Location:** Footnotes, `[^Sprt]`: "expected sample size reduced 'by at least 36% and at most 75%' —
  quoted as 36–75%."
- **Leaf verification (arxiv.org/abs/2603.00216):** exact sentence confirmed — "for **symmetric error
  bounds**, the sequential test reduces the average sample size by at least 36% and by at most 75%." The
  band is condition-scoped to symmetric error bounds (α = β); the footnote quotes the band without the
  condition, and the §2.3 body (slice 2's section — propagation site noted for the merge) renders it "at
  matched error rates," which is the adjacent-but-different condition (same error levels between the two
  tests, not symmetry between the two error types).
- **Grade:** likelihood **certain** (condition omitted) × impact **low** (the figure is illustrative;
  §2.3's argument survives any width of band) × complexity **trivial** (add "for symmetric error
  bounds"). Within-source condition-scoping class (`pattern_within_source_condition_misattribution`,
  mild form). **Confidence: MEDIUM** as written (quote exact, condition dropped).

### L4-F3 — §7's "paywalled attaches ONLY to ..." enumeration has in-report counterexamples

- **Location:** §7 Pre-flight self-audit, verification-limits bullet: "'paywalled' attaches ONLY to
  [^ExpertCvss] (ScienceDirect abstract) and to the unverified Computers & Security 2026 'Fragmentation'
  paper (403 at the abstract, attempted round 1)."
- **The check:** the same report records an ACM **403** on the Votta primary ("ACM primary 403 to both
  seats," [^Votta] — named two sentences later in the same bullet, so the bullet contradicts its own
  "ONLY"), and red's round-1 verification of [^DalalMallows] found the tandfonline primary **403'd**,
  verified via the Höhle 2016 exposition (ledger, round-1 lens-1) — yet the [^DalalMallows] footnote
  carries a bare primary URL + access date with no secondary-provenance disclosure, and §7's enumeration
  omits it entirely. An exhaustive-sounding sweep ("ONLY") that skips two access-blocked primaries is the
  `pattern_exhaustive_sweep_omits_hard_case` class; for [^DalalMallows] the access date on a URL that
  403'd at the verifying seat borders on the laundering §7 forswears.
- **Caveat weighed:** blue's seats may have had different access results on tandfonline than red's seat;
  the claim-source pair itself remains verified (HIGH via detailed secondary). This is a
  disclosure-completeness defect, not a corroboration failure.
- **Grade:** likelihood **medium** (turns on whose fetch 403'd; the Votta self-contradiction inside the
  bullet is certain) × impact **low** (§7 bookkeeping) × complexity **trivial** (rephrase "ONLY" to name
  all four; add the via-secondary note to [^DalalMallows]).

## Verified clean this round (slice 4)

| Claim | Reference | Confidence |
|---|---|---|
| [^CathedralBazaar] all five figure clusters: 44,180 Pairwise / 72,122 Consumer-View / 194/266 = 73% / 139/288 = 48% / AC-UI-Impact concentration / "accuracy can drop by 40%" | arXiv:2607.05670 abs + /html/v1, leaf-fetched twice | **HIGH** (new-citation repair, fully corroborated) |
| [^CostAudit] round-1 additions: killed r6 spawn = $0.61 (table row red-lens r6); "rounds 1–2 closed 31 gaps ($60-ish); rounds 3–5 closed ~15 mostly-trivial" verbatim (finding 4); pricing header sonnet 2.5 / session 12.5 → 5× multiplier confirmed; red-lens Σ$49.48 = 9.28+9.22+9.46+10.47+11.05 recomputed; %s recomputed (38.4/33.0/12.1) | cost.md @ bfa8a3b, direct read + recompute | **HIGH** |
| §6.4 item 5: "54KB" present verbatim in run-3 friction #15 ("the 54KB living blue/report.md") AND backlog 31(g) ("25k cap vs 54KB living report"); matches neither pinned file size | friction.md entry 15 + ideas/backlog.md line 31, direct read | **HIGH** |
| [^BacklogLevers] R1-21 repair: log()-per-transition heartbeat lives in backlog line 31's STILL OPEN list ("NEXT PR: emit log() at every state transition..."), adjacent to item 30 | ideas/backlog.md line 31, direct read | **HIGH** |
| §6.1 projected judge line's code facts (contested docket ll.244–245; judge reads debate.md + red/findings.md in full l.249) and §6.4 item 6 (carried never enters adjudicated, ll.252–253) | ledger round-1 merge re-verifications, pinned file, pin re-confirmed this round | **HIGH** (carried forward) |
| [^Sprt] band figures 36/75 as numbers | arXiv:2603.00216, exact sentence quoted | **HIGH** on the numbers (see L4-F2 for the dropped condition) |
| [^AdaptiveStability] −1.03% delta + KS<0.05/2-rounds criterion + 10-round baseline | arXiv:2510.12697v1 | **HIGH** on these three (see L4-F1 for the band) |
| [^PropagationChains] / [^Backlog28d] / [^Votta] split / [^WeakJudges] gloss / [^FrontierH3] H1 grades / [^ConflictingScores] correction record / [^AlreadyShipped] PR #18 discharge | each matches the round-1 red-verified facts recorded in the ledger (L3-F1/L3-F2/L3-F3/L2-F2/R1-19/L2-F1+L4-F1(r1)/L3-F5) | **HIGH/MEDIUM as ledgered** — repairs faithful to red's findings |
| §7 lane-2 disconfirming budget 4/13 (31%); memory-unreadable-from-blue-seats | ledgered round 1, sections unchanged in substance | **HIGH** (carried) |
| Pin equivalence (both corpora) | git diff --stat re-run this round, empty | **HIGH** |

## Lens verdict contribution

Slice 4 is substantially clean: 10 of 13 round-1 repairs touching this slice verify faithfully at the
leaf, including the one wholly new citation ([^CathedralBazaar]). Three graded defects, none above
LOW-impact, one of them (L4-F1) a repair-regression whose vector was red's own round-1 phrasing —
lineage must be declared if merged as a successor to R1-20.

## Friction

- The 25k Read cap again forced three windowed reads of the 1178-line living report at a seat whose
  contract mandates a full re-read — the same friction the audited corpus itself documents (friction #15
  class). A section-indexed or capped-file-aware full-read path for the designated audit surface remains
  the shape the work wants.
- WebFetch digests are small-model summaries; confirming Table 2 of arXiv:2510.12697 required a second
  targeted fetch to trust the min/max stopping rounds. A table-extraction path (the `arxiv-latex` MCP the
  protocol names) was not available at this seat's tool surface this round; the two-fetch workaround
  sufficed but is the documented lossy-PDF/HTML gap.
