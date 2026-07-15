# Red round-4 — lens 2 (leaf-node citation verification, instance 2 of 4; slice 2: §2–§3)

Full living report re-read whole (1534 lines, three windowed reads). Pin equivalence re-run
first-hand: both `git diff --stat` checks empty (retrospective @ `bfa8a3b`; backlog + plugin @
`5396952`) — all carried pinned-corpus verifications remain drift-proof. Ledger clause applied:
§2/§3 external sources last leaf-verified round 1 (>2 rounds elapsed) were re-fetched this
round; round-2/3 verifications carried.

## Findings

### L2-F1 (NEW) — §3.3(v)'s restated operator surface overclaims its readership; the red-side watchman still has no read path

- **Location:** §3.3 "Mechanism (lane 3's design...)", clause (v). Challenged sentence:
  "pending-entry deltas are LISTED in the round's `### RED` debate entry — a git-tracked
  surface every seat and the human operator already read" (and the follow-on "any seat (blue,
  a lens, the lead, the operator) may docket a listed delta to the judge within the window").
- **Leaf evidence (debate.js @ `5396952`, prompt regions ll.195–265, read first-hand this
  round):** the lens prompt names `blue/report.md`, `CHANGELOG.md`, the citation ledger, and
  the candidates output path — **never debate.md** (live corroboration: this very dispatch's
  prompt; a lens is forbidden to *write* debate.md and never instructed to read it). The
  red-merge prompt reads `red/candidates/` + `red/findings.md` and **appends** the `### RED`
  section — no read mandate on debate.md. The blue prompt DOES read "the latest '### RED'
  section" every round. The judge reads debate.md in full **only when a contested docket
  fires that round** (ll.247–250).
- **The defect:** "every seat ... already read" is false at the pin for both red-side seat
  classes — and those are exactly the seats with adversarial standing against a blue-favoring
  accepted deflation. The actual in-run watchmen of the one-round window reduce to: blue (the
  deflation's beneficiary), the human operator, and the judge conditionally on an unrelated
  docket firing. The "any seat ... may docket" enumeration names **a lens** as a docketing
  actor that (a) cannot see the surface under the pinned read instructions and (b) has no
  named channel to the judge other than via red-merge — the seat that just accepted the delta.
  The listing mechanism itself is sound and fail-closed (unlisted deltas never enter mass;
  checkable against the telemetry delta record) — the hole is narrower than R3-3's
  "window-without-a-watchman" but the *stated basis* for closing R3-3 ("already read", zero
  new mechanism) is contradicted by the engine text it cites.
- **Chain note for the merge:** this is the fourth consecutive round the accepted-branch
  absorber rests on a mechanism defect (R1-2 → R2-6 → R3-3 → this); §4.5 condition 7's
  write-up tracks the analogous three-round chain explicitly — this one should be tracked the
  same way. Merge to decide lineage (descends from R3-3).
- **Grade:** likelihood **medium-high** (the textual claim is certainly false at the pin; the
  operative hole is conditional on actuation — the same conditionality blue's own §3.3
  likelihood-honesty paragraph applies to (v)–(vii)) × impact **medium** (the window retains
  some watchmen; the design's honesty clause "the accepted branch's absorbers are the
  OPERATED window in-run" overstates operation) × complexity **low** (restate the readership
  from the pinned prompts — blue always, judge on dispatch, operator; and give the red-side
  watchman a named read: e.g. next round's red-merge MUST read the prior round's `### RED`
  delta list before grading, or the delta list rides a surface lenses already read).
- **Corroboration confidence of my own evidence: HIGH** (direct read of pinned engine text +
  live corroboration from this dispatch's own prompt).

### LOW note (merge may fold or drop)

§3.3(v) cites "clause (vii)'s 5/round cap" as a fixed constant; (vii) itself offers 5 only as
"e.g." — a worked example cited as a spec value. Trivial wording nit, no successor warranted
on its own.

## Verifications (clean) — appended to citation ledger

| Claim (statement ↔ reference) | Result | Confidence |
|---|---|---|
| §3.6 [^NineJudges] arXiv:2605.29800 — "about 2 independent votes' worth", "8–22 percentage points short", "best single judge matches or outperforms the full panel across all conditions", "close at most 11% of this gap" | all four verbatim in abstract, re-fetched | HIGH |
| §3.6 [^PoLL] arXiv:2404.18796 — smaller-model panel outperforms single large judge, "over seven times less expensive", "less intra-model bias due to its composition of disjoint model families" | verbatim, re-fetched | HIGH |
| §3.6 [^PersuasiveDebate] arXiv:2402.06782 — "76% and 88% accuracy respectively (naive baselines obtain 48% and 60%)" | verbatim, re-fetched | HIGH |
| §3.6 [^WeakJudges] arXiv:2407.04622 — debate beats consultancy "across all tasks" (random-assignment condition); gains vs direct QA task-dependent ("mixed" without information asymmetry); stronger-debater effect "more modestly than in previous studies" | all three glosses match, re-fetched; R1-22 repair still faithful | HIGH |
| §2.2/§3.2 [^CvssInconsistent] arXiv:2308.15259 — "59 participants ... 68% of these users gave different severity ratings" | verbatim, re-fetched | HIGH |
| §2.3 [^Iso29119] arXiv:1905.10676 — title + 29119 risk-based-testing coverage | abstract matches; the 29119 normative gloss itself still unfetched (standard paywalled), self-labeled "Via" | HIGH title / MEDIUM gloss (unchanged) |
| §2.3 [^CaptureRecaptureEval] Briand quote "no model is sufficiently accurate and underestimation may be substantial" | ieeexplore page rendered empty at this seat this round; exact phrase corroborated via search record of the same paper (ResearchGate listing); R1 quote-match stands | HIGH (carried R1 leaf-match + R4 secondary corroboration) |
| §2.1 item 4 (R3-14 repair) — means 4.9/5.9/4.4/6.0/5.2, band 4.4–6.0, "two highest are rounds 2 and 4" | recomputed from the printed lane-1 row: 98/20, 65/11, 44/10, 30/5, 31/6 ✓; 3.5-vs-4 illustration arithmetic ✓ | HIGH |
| §2.4 (R3-10 repair) — "× 3 throttled rounds (the low-mass rounds 3–5 of §2.1's series)" basis now stated in-section; ~$6×3=$18, ~10% of ~$180 | present + arithmetic holds (recomputed) | HIGH |
| §2.5 item 1 (R3-7 repair) — presence-check honesty ("catches a missing line, never a wrong one") + mass/board independent re-derivation clause at a non-red-merge seat | present, faithful to the raise | HIGH |
| §2.5 item 1 + carried design (R3-8 repair) — mapping "pinned before the first logged round", version-stamped, changed mapping = NEW series, no cross-version actuation comparison; realized excluded; §2.1 series ring-fenced as historical; consistent with §8 Q6's pinned enumeration | present at all three sites, mutually consistent | HIGH |
| §3.3(v) (R3-3 repair) — "or, equivalently" dropped ("as a second guard — NOT equivalent"); threshold made cumulative-per-round with overflow batch-docket mirroring (vii) | present, faithful on these two heads; the operator-surface head is L2-F1 | HIGH (two heads) / defect (third) |
| §3.3(vi) (R3-12 repair) — terminal resolution set excludes `carried` (accepted-with-delta / rejected-recorded-as-contested / unresolved); exit-time dispatch priced marginal-unless-sole-member; §3.3 cost line updated | present, faithful | HIGH |
| §3.3 default-to-docket (R3-9 repair) — repriced via §6.4 item 6 cross-reference, sole-member rule | present, consistent with the one-dispatch-per-round mechanics re-read first-hand (ll.247–250) | HIGH |
| §2.3/§3.3 judge enum "closed \| rebuttal_sustained \| risk_accepted \| carried \| unresolved" lacks grade-wrong; citationPasses formula `min(4, max(1, ceil(claim_count/40)))` | both re-read first-hand in debate.js this round | HIGH |
| Pin equivalence (both corpora) | re-run, both diffs empty | HIGH |
| Carried without re-fetch (within 2 rounds or pinned): [^CathedralBazaar] (R2 ×3 fetches), [^Sprt] (R3 ×2), Fragmentation-403 (R2 first-hand), [^DalalMallows] via-secondary, [^ExpertCvss]/[^FentonOhlsson] as-labeled MEDIUM/MEDIUM-HIGH, §2.3 corpus reads (R4-1 4-of-5, R5-5 singleton, R5-1 3-of-5 — pinned files, pin re-verified), §3.1 zero-LEAD, §2.4 6-lens shape (rounds 1–3 candidate dirs hold 6 files each — live re-confirmed by `ls`, and this round again dispatched 4 citation instances + 2 fixed) | ledger cross-check | as ledgered |

## Slice verdict (lens-scoped)

§2 is clean at the leaf this round — every round-3 repair in §2 verified faithful, every
re-fetched external reproduced verbatim, no repair-regression found. §3 carries one new
finding (L2-F1, medium-high × medium × low): the R3-3 repair's operator-surface claim is
contradicted by the pinned engine prompts it rides on. No prior-round closure in this slice
regresses.
