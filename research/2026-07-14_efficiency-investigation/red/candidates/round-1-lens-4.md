# Round 1 — Lens 4 (leaf-node citation verification), slice 4 of 4: §6 Cross-cutting, §7 Pre-flight self-audit, §8 Open questions, + footnotes those sections carry

Audit surface: full re-read of `blue/report.md` (914 lines, whole document, both pages), then leaf-node verification of every statement↔reference pair in §6–§8. Ledger was empty at pass start; all verifications fresh. Pin equivalence independently re-run (`git diff --stat bfa8a3b`/`5396952` — both empty).

## Verifications (corroboration confidence per pair)

| # | Claim (section) | Reference | Leaf check | Confidence |
|---|---|---|---|---|
| 1 | §6.1 red-merge $57.54 (38%) > red-lens $49.48 (33%) > blue-respond $18.21 (12%); per-round series | `cost.md` @ bfa8a3b table | Sums recomputed from the table: 7.52+13.22+12.64+10.60+13.56=57.54 ✓; 3.95+3.96+2.98+3.05+4.27=18.21 ✓; total $149.95 ✓ | HIGH (see L4-F4 on the red-lens total's round-6 exclusion) |
| 2 | §6.4.1 cost.md finding 2 internally contradicted (r5 merge $13.56 > r2 $13.22; 7.87M > 5.64M cache reads; 61 turns) | `cost.md` findings + table | Finding 2 text reads "merge cost tracks DISPUTE size (peaked r2, fell after)"; table shows r5 = $13.56, 7.87M, 61 turns vs r2 = $13.22, 5.64M — contradiction confirmed first-hand | HIGH (but see L4-F2 on "smallest board") |
| 3 | §6.4.2 backlog severity-floor claim "would have ended run 3 at round 3 for ~$10" | `ideas/backlog.md` item 30 @ 5396952 (line 30) | Quote verbatim in item 30 ✓; round-3 board per findings.md lines 620–621: "2 medium-high (R3-1, R3-2)"; R3-1 line 717 "control-flow trace of debate.js... complexity low", R3-2 line 737 "direct read of debate.js... low" — both code-trace, neither ≤MEDIUM/trivial. Contradiction confirmed | HIGH |
| 4 | §6.4.3 frontier misattribution 1: friction #15 concerns blue/report.md, not findings.md | `friction.md` entry 15 @ bfa8a3b; `blue/frontier.md` H4 | Friction #15 verbatim: "Read-tool 25k-token cap vs the 54KB living **blue/report.md**"; frontier H4 verbatim: "friction #15: the 54KB **findings.md**" — misattribution real, blue's correction accurate | HIGH |
| 5 | §6.4.3 frontier misattribution 2: R4-3 is a weak type specimen for lever 5 | `blue/frontier.md` H5; findings.md R4-3 (line 485) | Frontier H5 verbatim: "R4-3 is the type specimen"; findings R4-3 grading: ambiguous sentence "in the same cell" as R3-5's fix — row-granular scoping would include it. Correction accurate | HIGH |
| 6 | §6.4.4 Grep count-mode footgun recurred live; second documented recurrence | run-3 friction #12; this run's friction.md; live re-run | Friction #12 (64 vs 66, lines-not-occurrences) ✓; this run's friction.md entry (blue-lane-1) logs the recurrence ✓; independently re-run at this seat: anchored `grep -n "^### "` = 11 headers (6 BLUE, 5 RED, 0 LEAD); unanchored `grep -c "### LEAD"` = 5 ✓ | HIGH |
| 7 | §7 "lane 2 logged 4 of 13 searches disconfirming (31%)" | `blue/candidates/lane-2.md` lines 9, 416 | "13 searches/fetches; 4 spent on disconfirming" + "4 of 13 searches (31%)" ✓; 4/13 = 30.8% ✓ | HIGH |
| 8 | §7 red's gap-pattern memory unreadable from any blue seat, logged to friction.md | this run's `friction.md` | Logged independently by blue-lane-1, blue-lane-2, blue-synthesize (three seats) ✓ | HIGH |
| 9 | §4.3(c)/CHANGELOG synthesis Write of blue/report.md refused, neutral-name + cp detour (cited in my slice via §6.4 context) | this run's `friction.md` blue-synthesize entry | Entry present, names the refusal message and the detour ✓ | HIGH |
| 10 | Footnote [^FrictionFifteen] measured sizes: findings.md 106,772 B; blue/report.md 159,394 B; findings.md 1364 lines ([^FindingsBoard]) | working tree @ pin (diff-empty) | `wc -c` = 106772 / 159394; `wc -l` = 1364 — exact | HIGH |
| 11 | [^R4OneDetail] R4-1 grading quote; [^R4FourGrep] grep quote | findings.md lines 425, 503 | Both verbatim ✓ | HIGH |
| 12 | §7 "the 68% arXiv-open figure carries the point" ([^CvssInconsistent]) | arXiv:2308.15259, leaf-fetched | Abstract verbatim: "In a follow-up survey with 59 participants... 68% of these users gave different severity ratings" — figure and n exact | HIGH |
| 13 | §2.2/§7 "~34% NVD-vs-CNA disagreement" ([^ConflictingScores]) | arXiv:2508.13644, leaf-fetched (abs + full /html/v1) | **Figure absent from the cited paper.** Paper compares four scoring *systems* (CVSS/EPSS/SSVC/Exploitability Index) on 600 Microsoft CVEs; no NVD-vs-CNA dual-assessment analysis at all | LOW → L4-F1 |
| 14 | §8 Q2 "backlog 28(d) names it" (per-agent timeline) | backlog item 28 line 28 | "cost.md should show a per-agent timeline" — present ✓ (but see L4-F3 on the footnote's lever enumeration) | HIGH |
| 15 | [^PinCheck] pin equivalence (preamble; re-run as read diligence) | git diff --stat both pins | Both empty ✓ | HIGH |

## Findings

### L4-F1 — Citation misattribution: the ~34% NVD-vs-CNA figure is not in its cited paper
- **Location:** §7 Pre-flight self-audit — "lane 2's ~34% NVD-vs-CNA figure and the expert-CVSS moments are from search digests, not leaf-verified (paywalled; not load-bearing — the 68% arXiv-open figure carries the point)"; claim site §2.2 — "NVD-vs-CNA disagreement on roughly a third of dual-assessed CVEs"; footnote [^ConflictingScores].
- **Evidence:** leaf-fetched arXiv:2508.13644 abstract AND full HTML (v1). The paper compares four scoring systems applied to the same 600 Microsoft vulnerabilities; it contains no NVD-vs-CNA (vendor) dual-assessment comparison and no ~34% disagreement figure. The footnote's own hedge ("not leaf-verified against the paper's tables") implies the figure is in this paper's tables — it is not in the paper. This is misattribution, not mere non-verification: a digest figure wearing a named paper's citation (repair-regression/footnote-over-attribution class). Additionally §7's excuse "paywalled" is false for this source — arXiv:2508.13644 is open-access and was fetchable in two calls from this seat; only the ExpertCvss Computers & Security paper is paywalled.
- **Grade:** LOW-MEDIUM — certain (text defect, verified) × low-medium (blue already graded it MEDIUM and declared it non-load-bearing; the 68% figure independently verified and does carry the noise point; but a misattributed citation is worse than an unverified one — it launders a search digest into a specific primary source) × trivial-to-fix.
- **Required fix:** drop the ~34% claim and its footnote, or re-source it to the paper that actually contains an NVD-vs-CNA figure (do not keep the current attribution); correct §7's parenthetical so "paywalled" attaches only to [^ExpertCvss]. The §2.2 conclusion needs no change — it rests on the verified 68% figure plus the in-loop robustness evidence.

### L4-F2 — "Smallest dispute board" contradicted by blue's own §1.1 table
- **Location:** §4.2 / §6.4.1 — "round 5's merge was the run's most expensive ($13.56 > r2's $13.22; 7.87M cache reads > r2's 5.64M; 61 turns) on the run's *smallest* dispute board (6 open gaps)".
- **Evidence:** blue's own §1.1 board table: round-4 board = 5 open < round 5's 6. Whichever board is meant (the merge's input board — round-4 residual, 5 open — or its output board, 6 open), the pairing "smallest + (6 open gaps)" cannot both hold: 6 is the second-smallest, and the actual smallest board (5, post-round-4) had a $10.60 merge.
- **Grade:** LOW — certain × low (the §6.4.1 defect finding survives: the most expensive merge on a near-minimal board still contradicts "peaked r2, fell after"; only the superlative is wrong) × trivial.
- **Required fix:** replace "smallest" with "second-smallest (6 open, vs round-4's 5)" or key the sentence explicitly to the merge's input board and use the right count; propagate to both sites (§4.2 and §6.4.1).

### L4-F3 — [^Backlog28d] mis-enumerates the backlog's levers
- **Location:** Footnotes — "[^Backlog28d]: ... levers (1) shard, (2) collator, (3) prompt-level read batching, (4) per-agent timeline."
- **Evidence:** backlog item 28(d) @ 5396952 (line 28) enumerates lever "(4) TOOLING step-up if gap volume grows: evaluate beads... vs a tiny sc-gaps Go tool"; the per-agent timeline is the item's separate closing sentence ("cost.md should show a per-agent timeline"), not lever (4). §8 Q2's substance ("backlog 28(d) names it") survives; the footnote's enumeration does not.
- **Grade:** LOW — certain × low (no argument turns on lever (4); but the footnote is the skeptic's map to the source and currently mislabels it) × trivial.
- **Required fix:** footnote wording: "...(3) prompt-level read batching, (4) tooling step-up (beads / sc-gaps); plus a closing note that cost.md should show a per-agent timeline."

### L4-F4 — Red-lens run total silently excludes the round-6 killed spawn
- **Location:** §2.4 (figure reused in §6.1's ranking, my slice) — "Red-lens totaled $49.48 of $149.95 in run 3 (33%)"; footnote [^CostAudit] "red-lens $9.22–$11.05/round, Σ$49.48".
- **Evidence:** cost.md table: rounds 1–5 sum to exactly $49.48, but the seat has a sixth row (red-lens r6, killed mid-spawn, $0.61); seat total $50.09. The exclusion is defensible (finding 5 treats r6 as the stop-and-resume artifact) but unstated — "totaled ... in run 3" reads as the seat's run total.
- **Grade:** LOW — certain × low (0.4% of run spend; no conclusion moves) × trivial. A scoping parenthetical ("rounds 1–5; the killed r6 spawn adds $0.61") closes it.

### L4-F5 — The floating "54KB" figure: a fifth pinned-artifact defect §6.4's inventory misses
- **Location:** §6.4 Defects found in pinned artifacts (inventory of 4); prose site §4.3(c) — "a small open ledger is possibly one detour *cheaper* to Edit than a 54KB monolith."
- **Evidence:** friction #15 (pinned) says "25k-token cap vs the 54KB living blue/report.md" — but the pinned blue/report.md is 159,394 bytes (blue's own [^FrictionFifteen] states the measured sizes). 54KB matches neither pinned file (findings.md = 106,772 B). Backlog item 31(g) repeats the same "54KB living report." Blue quietly supplies correct sizes in the footnote yet (a) reuses the stale "54KB" in §4.3(c) prose, now attached to findings.md — a third, different referent — and (b) omits the friction-#15/backlog-31(g) size error from §6.4's pinned-artifact defect inventory, which is exactly that inventory's job.
- **Grade:** LOW — certain (all three sizes verified first-hand) × low (no disposition turns on it; the 25k-cap complaint is valid at either size) × trivial.
- **Required fix:** §4.3(c): replace "a 54KB monolith" with the measured size of the file meant (findings.md, ~107KB at pin); add a fifth §6.4 entry noting friction #15/backlog 31(g)'s stale 54KB figure so the artifact's successor corrects it.

## Slice summary

§6–§8's corpus-facing citations are in strong shape: every load-bearing quote I followed (backlog item 30, cost.md table and findings, findings.md boards and grade strings, friction entries 12/15, frontier H4/H5, lane-2 self-audit, this run's friction.md, file sizes, pin diffs, debate.md header counts) reproduced verbatim at the leaf — 13 of 15 pairs HIGH. The two failures are external-literature hygiene (L4-F1, the only misattribution found — and it sits in a claim blue had already fenced as non-load-bearing) and small-numeral precision (L4-F2..F5, all trivial). No finding in this slice challenges any §6–§8 disposition.

Lens-seat verdict contribution: no HIGH/MEDIUM-HIGH gaps from this slice; L4-F1 is the only gap above pure-nit grade.

## Friction (for the envelope)

- WebFetch's small-model digest is a lossy instrument for proving a NEGATIVE ("figure absent from paper") — two fetches (abs + /html/v1) agreed, but a pdf-reader/arxiv-latex extraction would have made L4-F1's absence-claim airtight; MCP document tools were not exposed at this lens seat (consistent with the lead's observer-effect disclosure in friction.md).
- Slice arithmetic ("divide sections evenly, take slice 4") left the footnotes block unassigned; I took the footnotes my sections carry plus the two external ones §7 itself adjudicates. The dispatch should name the footnote-ownership rule explicitly.
