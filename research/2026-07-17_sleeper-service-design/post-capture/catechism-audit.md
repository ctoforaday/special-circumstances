# Post-capture red audit: The Catechism (out-of-band sitting, 2026-07-18)

Provenance: user-directed test of the assembly's own disclosure ("the catechism as a
unit was never itself a red target"). Red-auditor seat, READ-ONLY sitting over the
captured record; findings relayed verbatim below. The run record above this directory
is unchanged.

## Per-answer verdict table

| Answer | Claims checked | Findings (severity) | Clean? |
|---|---|---|---|
| 1 — what are we trying to do | 8 | F5a "picks by deterministic ranking" (medium, shared w/ A3); trivial input-list omission | NO |
| 2 — how handled today | 5 | F1 novel synthesis "dominant waste class" (medium); F2a/b strengthening "unactioned"/"without curation" (medium) | NO |
| 3 — what is new | 4 | F5b "full-strength model" (medium, shared w/ A1) | NO |
| 4 — case against | 14 | F3 "open bug trio" regression (low-medium); F4 STOP condition misattribution (low); F8 omission (low-medium); F9 chain-undercount + DGM plural (low) | NO |
| 5 — of interest | 4 | F6 "$2-5/night under $50/mo" composition drift (low-medium); F2c "demonstrably loses signal" (shared) | NO |
| 6 — what changes | 4 | none | YES |
| 7 — cost & stop | 9 | F7 unaudited R5-2 test stated flat (low); F9 "one-line"/"retire" (trivial) | MOSTLY (minor) |

## Findings

**F1 — NOVEL SYNTHESIS (medium: certain x medium x trivial-to-fix). Answer 2, line 36:** "captured-but-never-triaged friction is the dominant waste class in the suite's own telemetry (§1.1-§1.3)." No waste-class claim, dominant or otherwise, exists anywhere in the audited body (grep: only hits are the catechism itself plus unrelated uses at lines 832, 1824, 1841). §1.1-§1.3 establish recurrence, signal evaporation, and Dependabot volume-fatigue — never a comparative measurement of waste classes. A judge-authored superlative wearing a section citation.

**F2 — STRENGTHENING cluster contradicting the embedded §1.4 (medium: certain x medium x trivial).** (a) Line 36 "complaints recurring unactioned" vs §1.4 lines 595-597: the by-hand loop "has run after every run and works." (b) Line 36 "backlog accretes without curation" vs §1.4 lines 588-593: backlog is "the human-curated intermediate." (c) Line 42 "manual triage demonstrably loses signal (R1-14)" vs §1.4's actual R1-14 repair: the manual loop works and "daily automation must argue its margin over it"; the nearest loss evidence is CAPTURE loss, a different mechanism. The catechism un-prices the null alternative R1-14 forced blue to price.

**F3 — REGRESSION of a corrected claim (low-medium). Answer 4(f), line 40:** "the open MCP-headless bug trio." R1-5's round-1 correction (archive lines 3004-3011; citation-ledger line 8 graded "LOW — REFUTED"): TWO open (#76239, #68375), ONE closed-as-duplicate (#32191). The catechism reinstates the refuted pre-repair phrasing.

**F4 — CONDITION MISATTRIBUTION (low). Answer 4(a), line 40:** "STOP measured 0.42% ... even with warnings." Artifact of record (lines 2374-2381, 2053-2056): 0.42% is the NO-warning arm; with-warning is 0.46%. Direction survives; figure bound to the wrong experimental arm against a ledger that pinned both arms four rounds running.

**F5 — MECHANICS/JUDGMENT TRANSPOSITION, internally contradictory between answers 1 and 3 (medium).** (a) Line 34 "picks ... by deterministic ranking" vs §1.4: RANKING is deterministic, the PICK is model judgment ("the pick is judgment," line 476). (b) Line 38 "full-strength model" vs §5.2 (lines 1824-1840): nightly pick and lanes pinned to BULK tier; protected judgment "is never exercised unattended, and therefore never cheapened"; "full-strength" belongs only to /graduate and occurs nowhere in the audited body.

**F6 — NUMBER/COMPOSITION DRIFT reinstating what R2-18 closed (low-medium). Answer 5, line 42:** "~$2-5" is the per-run anomaly CEILING, not running cost (expected ~$0.10-0.50/night, §5.2 line 1826, archive lines 3377-3382); nightly $2-5 under $50/month is arithmetically impossible (30 x $2-5 = $60-150 trips the cap mid-month by design).

**F7 — UNAUDITED ROUND-5 REPAIR STATED AS ESTABLISHED (low). Answer 7, line 46:** "The Phase-4 acceptance test IS two-legged" — R5-2's repair of record, marked UNAUDITED at lines 16 and 96. Answer 4(b) carries the unaudited caveat for R5-3; answer 7 omits it for R5-2 — inconsistent caveat discipline in the section a builder reads for the acceptance gate.

**F8 — OMISSION from "the case against — at full strength" (low-medium). Answer 4:** absent are (i) the layer-4 subprocess residual — the design's own "honest gap" (§4.3 line 1738; row 4: Low-Med x High, risk-accepted) and (ii) read+egress exfiltration (row 13, residual risk-accepted) + injection-via-retrieval (row 14). The against-case omits the design's highest-impact accepted residuals.

**F9 — minor cluster (trivial-to-low):** "THREE rounds of red pressure" is four counting R1-25 (whose heading the same sentence quotes); "DGM agents removed their guardrails" pluralizes one marker-removal incident; "two one-line cross-plugin edits" — the /doctor delta is multi-line; "or retire the loop" appears nowhere in the audited body; answer 1's consumed-sources list omits §1.3's run-records row.

## What checks out

Answer 6 fully clean. Backlog 25-at-pin exact (R1-1 recount). Gap-mass/open-count trajectories exact match to board-telemetry.jsonl all five rounds. Probe P2 claims exact. #22055, AI Scientist, Dependabot, H2 falsifier, quota-precheck requalification, seven layers, artifact counts, k=3/M=3/N=7, and the OQ probe list all faithful. The assembly disclosure itself (line 32) is honest and matches friction.md line 29.

## Closing judgment

**DEFECTIVE** — classes by mass: strengthening (F2, F5, F9), novel synthesis (F1), number/condition drift (F4, F6), repair-regression (F3), omission (F8), one internal contradiction (answers 1 vs 3). Six of seven answers carry at least one defect existing nowhere in the audited body. No fabricated citation, no inverted safety property — but the defects are directionally patterned: value-side answers strengthened pro-build; the against-case drops the two highest-impact accepted residuals; three defects reinstate exact pre-repair phrasings the debate spent rounds R1-14, R1-5, and R2-18 correcting. The catechism is the unaudited-synthesis problem realized on its own document: the judge, assembling from memory of the debated text, regressed toward the round-0 shapes red had already burned down.

**Structural fix (the run's own doctrine applied to itself):** make the Catechism a required section of blue's round-0 report.md — inside red's mandatory full-re-read every round — so assembly becomes union-copy, not authorship (friction.md line 29 named this). An assembly-time quotes-with-anchors rule is worth stating as belt, but it is an instructional gate of exactly the class this report refutes; the in-debate placement is the close of record.

**New gap pattern (for red's memory):** *assembly-seat regression toward pre-repair phrasings* — a synthesizing seat writing from recall of a many-times-repaired document preferentially re-mints the round-0 form of exactly the claims that were repaired (three independent instances: F2c/R1-14, F3/R1-5, F6/R2-18). Mechanical screen: grep the assembly for every token in the run's propagation-grep lists.
