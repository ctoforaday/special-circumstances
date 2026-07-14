# Red audit — round 2, lens 4: logic and completeness

Lens mandate: leaps of faith, missing counterarguments, unexplored alternatives, template
compliance. Full living report re-read in context (all 704 lines of `blue/report.md`, current
state — every Round 1 correction block included, not a diff against `blue/CHANGELOG.md`), plus
`blue/CHANGELOG.md`, `debate.md` (both rounds), `red/findings.md` (round 1, 20 gaps, all conceded
per blue's Round 1 response), `red/candidates/round-1-lens-4.md` (this lens's own prior pass, to
avoid re-litigating closed ground), and fresh live verification against the current
`special-circumstances` repo (`git log`, `git show`, `git ls-tree`, direct read of `debate.js` on
`main`) as of this audit — not reused from round 1's pin.

None of round 1's 20 gaps are reopened here; direct re-verification this round confirms the
judge-null-guard (line ~184) and `citationPasses`-outside-the-loop (line ~139) defects blue
conceded remain exactly as described, unchanged, on current `main`. This pass is new-ground only.

## GAP-R2L4-1 — The lane-diversity redundancy floor (added round 1, R1-16) is arithmetically incompatible with the report's own lane-count floor (row 7) at the stated defaults [severity MEDIUM-HIGH]

- **Location:** §3, row 6 — *"assign distinct method/source-class lenses (primary-literature /
  practitioner-production / adversarial-disconfirming-first / local-repo critical-stance), not
  persona text and not more headcount"* ... disposition: *"**Fix before run 4, scoped to
  source-class/method assignment, WITH an explicit redundancy floor**: assign the
  critical-stance/adversarial lens to at least 2 of N lanes (not 1-of-N), so a single null
  dispatch does not zero out a method's round coverage — cost is one more lane-dispatch, cheaper
  than losing the round's minority-report class entirely."* Against §3, row 7 — *"Lane-count floor
  (`lanes >= 3` or explicit justified override) ... Fix now."*
- **Problem:** the row's own roster names **four** distinct method/source-class lenses
  (primary-literature, practitioner-production, adversarial-disconfirming-first, local-repo
  critical-stance). Its redundancy-floor fix folds two of those four
  ("adversarial-disconfirming-first" and "local-repo critical-stance") into one bucket
  ("critical-stance/adversarial") and requires **2 of N** lanes for that bucket alone — without
  ever stating that this conflation is intentional, or that the other two named methods
  (primary-literature, practitioner-production) are being demoted to share the single remaining
  lane. At the report's own stated default (`lanes >= 3`, row 7), 2 lanes go to the merged
  critical-stance/adversarial bucket, leaving exactly **1** lane for both primary-literature
  *and* practitioner-production combined — a method-per-lane design that the same row's prose
  describes as four separate assignments. The "cost is one more lane-dispatch" phrase implies
  the fix is paid for by growing N by exactly one (e.g., 3→4), which is never stated and directly
  collides with §3's own "Explicitly risk-accepted" list: *"blanket lane-count raise as a
  diversity fix (method diversity is the cheaper, better-targeted lever ...)"* — the report
  argues against raising N as a fix in one place and quietly requires raising N to make its own
  redundancy floor arithmetically work in another, without reconciling the two.
- **Missing counterargument / unexplored alternative:** the report never states which of the
  three admissible resolutions it intends — (a) raise N to 4+ to fit all four methods plus the
  floor (reopens the risk-accepted headcount question), (b) treat
  adversarial-disconfirming-first and local-repo critical-stance as one method going forward
  (silently drops a named distinct lens the rest of §1.1 treats as separate, e.g. "lane C
  adversarial/disconfirming-first ... or lane N always runs a local-repo critical-stance pass"
  — phrased as alternatives, not requirements, elsewhere, which is itself inconsistent with row
  6 listing both as mandatory-sounding roster items), or (c) keep N=3 and let
  primary-literature/practitioner-production compete for the one remaining lane per round
  (silently reintroduces the "which hypothesis gets breadth-only coverage" under-provisioning
  problem H1 already flagged for run 2's `--lanes 2`).
- **Grading:** likelihood — high (the arithmetic is directly checkable from the row's own two
  numbers: 4 named methods, a 2-of-N floor on one bucket, N>=3 elsewhere); impact — medium-high
  (this is the report's own headline fix for its own most consequential confirmed finding, H1 —
  if it ships as literally written, the next run's lane-dispatch prompt cannot honor all four
  named assignments without either quietly dropping one or quietly growing N past the
  risk-accepted ceiling); complexity to mitigate — low (one sentence: either state N for this fix
  explicitly, e.g. "this fix requires `lanes >= 4`, superseding row 7's floor of 3," or state
  which two of the four named methods merge into the redundancy-floored bucket and why that
  merge is safe).
- **Disconfirming check performed:** considered whether "critical-stance/adversarial" was always
  meant as shorthand for a single bucket and the roster's four-way split is just descriptive
  color, not four separable assignments. Against this: §1.1's own sourcing paragraph attributes
  the two methods to different lane letters with "or" ("lane C ... or lane N ..."), i.e. as
  alternatives *within* the original four-item brainstorm, not a settled merge — the redundancy
  floor's silent 2-bucket conflation is a decision the report makes without saying it's making
  it.

## GAP-R2L4-2 — Row 19's risk-accept rests on a mitigation mechanism ("independent re-verification against a second source") that is not what the leaf-node citation lens actually does, and contradicts its own adjacent open question [severity MEDIUM-HIGH]

- **Location:** §3, row 19 — *"red's leaf-node citation lens (every claim traced to a source a
  skeptic can follow) is a real, already-existing structural defense against exactly this class —
  a poisoned page asserting a fake fact still has to survive **independent re-verification
  against a second source**."* Against §5, item 8 — *"Is a poisoning attack against the citation
  itself — a fabricated but internally-consistent secondary source, as opposed to a fabricated
  primary claim — covered by the leaf-node citation-verification lens, or does it require a
  distinct defense (e.g. **cross-referencing claimed sources against an independent index**)?"*
- **Problem:** the disposition asserts, as an already-existing mitigation, exactly the mechanism
  ("second source," "independent" cross-reference) that the report's own very next open question
  treats as a *hypothetical, not-yet-built* defense ("does it require a distinct defense... e.g.
  cross-referencing... against an independent index" — phrased as a proposal, not a description
  of current practice). These cannot both be true: either the lens already checks a claim against
  a second, independent source (in which case item 8's question is answered, not open), or it
  doesn't (in which case row 19's risk-accept rationale for the *easier* case — primary-source
  fabrication — is resting on a mechanism that doesn't exist). Direct check against this run's own
  described practice settles it against row 19: the protocol's stated leaf-node method (used
  identically by every lens pass in this corpus, including this one) is "follow the citation to
  the source; confirm the source actually corroborates the statement" — singular source,
  no independent second fetch mandated. A lane that fetches a poisoned page and cites it as its
  source would have that exact same page re-read by red's leaf-node check, which would correctly
  confirm "yes, this source supports the claim" — the poisoning would pass leaf-node verification
  undetected, because there is no second-source cross-check built into the practice as specified
  anywhere in this corpus (the one place a second source is invoked — the WisdomCrowds footnote's
  "search synthesis across... related literature" — is ad hoc research breadth, not a stated,
  repeatable verification step applied to every claim).
- **Why this matters for the lens (leap of faith):** the risk-accept disposition is only sound if
  the claimed mitigation is real. As written, row 19 both overstates the current defense (for the
  primary-fabrication half it claims is "covered") and is internally inconsistent with the
  report's own next paragraph about the same defense's actual scope. This is not a request to
  reopen the underlying risk-accept as wrong — a poisoned-primary-source attack surviving
  single-source leaf-node tracing is a real, still-open exposure the report should name plainly,
  not paper over with a "second source" claim that isn't practiced.
- **Grading:** likelihood — high (directly checkable against the protocol's own stated method,
  used throughout this very corpus, including by this lens pass); impact — medium-high (a
  security-adjacent risk-accept resting on an inaccurate description of its own mitigation is
  exactly the kind of gap that should not survive to a judge's docket un-flagged — the actual
  residual risk is larger than row 19 currently states, even though the disposition's ultimate
  call — risk-accept for now — may still be the right one once restated accurately); complexity
  to mitigate — low (drop "independent re-verification against a second source" and replace with
  an accurate description: single-source leaf-node tracing catches a claim that misstates its
  cited source, but does not catch a claim whose cited source is itself fabricated/poisoned and
  internally consistent — collapsing rows 19 and the §5 item 8 open question into one accurately-scoped
  residual-risk statement rather than two inconsistent ones).

## GAP-R2L4-3 — Live-source drift since round 1's pin: `main` has advanced two more commits, "run 3" evidently already executed live with zero artifact trail, and a fresh instrumentation-reliability defect surfaced that the report's own headline token-cost figures are never checked against [severity MEDIUM]

- **Location:** §0, footnote [^Reverify47ae48d] and [^PinnedRepoState] — *"This report's own
  repo-state citations from this round forward carry a SHA + UTC timestamp + 're-verify before
  acting' note"* (the discipline the report itself now claims to hold); §2.3 — *"Entire suite:
  zero tokens ... against historical incident costs of 252.9k (run 1) and ~3M tokens (run 2's
  quota crashes)"*; §2.4 — *"simulator = high likelihood x high impact (253k–3M tokens per
  historical incident) x low complexity"*; §3 item 11 — *"Formalize trajectory capture
  (`journal.jsonl` into `<run>/trajectories/`, gzip) after every `/research` run."*
- **Problem — live re-verification, this round, direct check:** `git log --oneline -5 main`
  shows two commits past the report's own pinned `47ae48d`: `88eb57f`
  ("docs(backlog): run cost audit — tool + findings + efficiency investigation"). Its body
  states, as a live finding from an actual "run 3": *"the panel token counter excludes cache
  traffic = 92% of real flow (panel said 610K; transcripts showed 47.7M); blue-synthesize on the
  session model was the single priciest agent ($10.58 — cache RATES, not output volume); red ≈
  $20/round at this corpus size; full run projects $80-120 at list rates."* Two things follow,
  neither addressed anywhere in the report:
  1. **Run 3 evidently already happened**, yet `git ls-tree -d main -- research/` shows exactly
     two run directories in the repo (`2026-07-12_feov-retrospective`,
     `2026-07-12_memory-architecture`) — **no run-3 directory, no friction.md, no
     journal.jsonl for it exists anywhere in the tree.** This is a live, fresh instance of the
     exact gap §3 item 11 argues for building (trajectory capture "after every `/research` run")
     — except this time the missing artifact isn't hypothetical, it's the report's own §5 open
     questions 4 and 7 (*"does the pre-created blackboard skeleton actually clear the
     write-block under real Task-tool permissions?"* / *"tested against the real skeleton path
     rather than a scratch run"*) going unanswerable in practice even though the run that could
     answer them has already occurred and left no trace to check.
  2. **The report's own headline token-cost figures are never re-examined in light of this.**
     `run1-friction.md`'s "252.9k tokens" and the "~3M tokens" run-2 quota-crash figure are used
     repeatedly (§2.3, §2.4) as the quantitative baseline the simulator/`--smoke` proposals are
     graded against ("high impact"). If those figures were produced by the same class of
     panel-display tooling the cost-audit commit just found undercounts real usage by
     omitting cache-read/write traffic (610K reported vs. 47.7M in the transcripts — roughly a
     78x gap), the report's own comparison baseline is potentially understated by a similar
     order of magnitude — which would *strengthen*, not weaken, the case for the simulator, but
     the report currently asserts a specific number without flagging that its provenance
     (panel-reported vs. transcript-derived) is now in live, sourced doubt.
- **Why this is a logic/completeness gap and not just more drift-hygiene:** the report explicitly
  adopted a going-forward discipline for exactly this class of problem after R1-1
  ([^PinnedRepoState]: *"re-verify before acting"*) — this round's re-verification is that
  discipline being exercised, and it surfaces something the discipline was designed to catch:
  the ground moved again, in a way that's substantively load-bearing (an artifact-trail gap the
  report's own recommendations argue matters, plus a fresh, quantified doubt over the report's
  own comparison figures), not merely another stale timestamp.
- **Grading:** likelihood — high that `main` has drifted and that the cited commit content is as
  quoted (direct `git log`/`git show`, reproduced above); high that no run-3 artifact exists in
  `research/` (direct `git ls-tree`); **medium** on whether the report's own 252.9k/~3M figures
  specifically suffer the same undercount (unverified either way — this is the gap, not an
  established fact, and should be stated as an open question, not an assertion); impact — medium
  (does not change the report's qualitative conclusions — the simulator/`--smoke` case only gets
  stronger if the true costs are higher — but a report that just adopted a
  re-verify-before-acting discipline should apply it to its own two most-cited efficiency
  figures once a live, sourced reason to doubt their provenance surfaces); complexity to
  mitigate — low (one footnote: flag that [^Run1Friction]'s and the run-2 quota-crash figure's
  measurement provenance is unconfirmed against the newly-discovered panel-vs-transcript
  undercount, note the direction of the risk (understatement, not overstatement), and add a §5
  open question tying it to item 11's trajectory-capture fix — which, once built, would settle
  this for future runs by construction).

---

## Round-2 lens-4 synopsis

Three new gaps, none disputing H1–H5's substantive conclusions and none reopening round 1's 20
closed items (spot-re-verified live: judge-null-guard and `citationPasses`-outside-loop remain
exactly as blue described, unchanged, on current `main`). One MEDIUM-HIGH (the lane-diversity
redundancy floor's own arithmetic doesn't reconcile against the report's stated lane-count floor
and its own risk-accepted ceiling on headcount raises — three admissible resolutions, none
chosen), one MEDIUM-HIGH (the content-poisoning risk-accept's stated mitigation mechanism —
"independent re-verification against a second source" — is not what the leaf-node lens actually
does anywhere in this corpus, and contradicts the report's own adjacent open question about the
same defense), one MEDIUM (fresh live-source drift discovered this round: `main` has advanced
past the report's own pin, "run 3" evidently ran with zero artifact trail — a live instance of
the exact gap item 11 argues for — and surfaced a quantified token-accounting reliability defect
the report's own headline cost comparisons are never checked against). All three are additive-fix
sized: one-sentence-to-one-paragraph corrections, not blocking, consistent with round 1's
convergence pattern.
