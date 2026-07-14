# Red audit — round 2, lens 1 (leaf-node citation verification, slice 1 of 3)

Scope: report.md §0 "Headline" and §1 "Doubts" (1.1 H1, 1.2 H2, 1.3 H3, 1.4 un-frontiered doubts),
plus the footnotes those sections cite that changed per `blue/CHANGELOG.md` Round 1 (R1-1, R1-2,
R1-4, R1-5, R1-6 as referenced in §0, R1-17, R1-18, R1-19 as referenced, R1-20). Per protocol,
claims already at HIGH confidence in `red/citation-ledger.md` whose section did not change this
round are not re-fetched. Sections 2-5 and their footnotes are out of scope (slices 2-3).

## Finding 1 — MEDIUM-HIGH severity: R1-17's rationale for deferring cross-provider model
diversity misreads its own cited source — the paper's "2 diverse agents match/exceed 16
homogeneous" result requires cross-model diversity, not lens/persona diversity alone

**Location:** §1.1, "Cross-provider model diversity — named and dispositioned (round 1, R1-17)" —
quote: *"the same paper's practical finding — '2 diverse [persona-lensed] agents match/exceed 16
homogeneous' — shows method/lens diversity within one provider already captures most of the
achievable gain without the infrastructure cost."*

**Verification performed:** direct fetch of arXiv:2602.03794 full HTML text (`arxiv.org/html/2602.03794`)
this round, targeted at the paper's own diversity taxonomy. The paper defines four levels: L1 (same
model, same prompt), **L2 (same model, different persona prompt — i.e., exactly the "lens/persona
diversity within one provider" blue's disposition rests on)**, L3 (different models, same prompt),
L4 "Full Diversity" (different models AND different persona prompts). The specific result the report
cites — "2 diverse agents match/exceed 16 homogeneous" — is **L4's result, not L2's**: per Table 2,
L4 needs only 2 agents to match L1's 16-agent baseline, while **L2 (persona-only, same base model)
needs 8 agents to match the same baseline** — a 4x efficiency gap between persona-only and full
diversity, not "most of the achievable gain."

**Why this matters:** the bracketed insertion `[persona-lensed]` is blue's own gloss, not the
paper's language, and it silently substitutes L2 for L4 in the very sentence built to justify not
building L3/L4 (cross-provider) diversity. The paragraph even quotes, two sentences earlier in the
same footnote-adjacent text, the paper's own correct qualitative finding — "same-base-model agents
remain more correlated than architecturally distinct ones" (this is accurate; see the already-cited
[^AgentDiversity] qualitative claim, verified clean in round 1) — and then draws the opposite
practical conclusion from it in the R1-17 paragraph. The report's own source, read past the
abstract, argues *against* the disposition it is cited to support. This is not a fabricated figure
(unlike R1-4's original miscitation) — it is a real citation being misapplied to a different
diversity level than the one it actually measures, immediately adjacent to the report's correctly-
hedged version of the same fact.

**Grading:** likelihood — realized (confirmed by direct source read, not inference). Impact —
medium-high: R1-17 is the report's own explicit, reasoned tradeoff decision ("defer, not adopt")
gating a real infrastructure question (§3 item 6's redundancy floor and the open question in §5
item 5 both depend on lens-diversity being "enough" to measure before revisiting); if lens-diversity
alone reaches only the L2 efficiency curve (8 agents to match baseline) rather than the L4 curve (2
agents), the "revisit if lens-assignment under-delivers" trigger is set against the wrong
comparison baseline. Complexity to mitigate — low: correct the sentence to state the finding
accurately (L4 requires both model- and persona-diversity; L2 alone is a smaller, real, but far
short of "most of the gain") and let the "defer, not adopt" disposition stand or fall on the
accurate number, not the inflated one — this may still be the right call, but not for the reason
currently given.

**Disposition:** raised, not closed.

## Finding 2 — LOW-MEDIUM severity: R1-5's replacement figure ("continued gains observed to 7
agents on the hardest") is a new unpinned precise number, sourced only to "independent re-search
this round" with no citation added — the same footnote-over-attribution failure R1-5 was raised to
fix, recurring in the fix itself

**Location:** footnote [^DiminishingReturns] (referenced from §1.1's bulleted disconfirming-evidence
list) — quote: *"Independent re-search this round (WebSearch, 2026-07-13) corroborates the
qualitative aggregate: accuracy plateaus around 2–3 debate rounds and 2–4 agents for
moderate-complexity tasks, with the breakeven at 3–4 agents for harder tasks and continued gains to
7 agents on the hardest — treat as a synthesis across sources, not a single citable number."*

**Verification performed:** search this round for the specific "7 agents on hardest tasks" figure.
It traces to a real, findable paper not in the footnote's four bundled sources (2603.20640,
2601.19921, VentureBeat, 2605.00914): **"The Ringelmann Effect in Multi-Agent LLM Systems: A Scaling
Law for Effective Team Size," arXiv:2606.02646** — which reports continued gains on harder tasks
(e.g. GSM-Plus) out to 7 agents and recommends size-7 as an optimal ensemble size for many use
cases. The number is real and directionally corroborated, not fabricated — but the footnote does
not cite it, attributing the figure only to an unlinked "independent re-search this round."

**Why this matters:** this is the identical failure pattern R1-5 exists to correct (a precise
numeric bound asserted without being pinned to the source that actually states it) reappearing in
R1-5's own replacement text, one round later. The hedge "treat as a synthesis across sources, not a
single citable number" reads as if it covers this, but a single citable number that supports the
"7 agents" clause specifically does exist and was simply not added.

**Grading:** likelihood — realized. Impact — low-medium: the figure turns out to be accurate, so no
reader is misled about the world, only about the report's own sourcing discipline — which is exactly
the property this section's citation apparatus is supposed to guarantee. Complexity to mitigate —
trivial: add arXiv:2606.02646 to the bundle with the GSM-Plus/size-7 figures it actually supports.

**Disposition:** raised, not closed.

## Finding 3 — LOW severity, informational: the report's newly-added pinned-SHA discipline
([^PinnedRepoState], added this round specifically to fix R1-1) is already one live-state
advance behind current `main` — confirmed harmless this time, but the drift recurred immediately

**Location:** §0, "Round 1 correction" — quote: *"`main` has since advanced again to `47ae48d`
(its commit message references 'run 3')."*

**Verification performed:** `git log --oneline -3 origin/main` (fetched live this round) shows
current HEAD is `88eb57f` ("docs(backlog): run cost audit..."), two commits past `47ae48d`
(`47ae48d` → `88eb57f` adds one intervening docs-only backlog commit not shown in this trace before
the HEAD I checked). `git diff 47ae48d 88eb57f -- .../debate.js` returns **empty** — no functional
change to the file the report's claims about line 136/139/148/181/184 depend on, so §0's specific
technical claims (judge site unguarded, citationPasses never rescaled) remain accurate against
current HEAD, independently re-confirmed this round by direct read of `debate.js` at current `main`.

**Why this matters:** this is not a correctness failure this time — it is the same live-source-drift
class the report itself just named and built a discipline for ([^PinnedRepoState]: "pin load-bearing
repo-state claims to a SHA + timestamp with a 're-verify before acting' note"), recurring against the
discipline itself within one audit round. The mitigation is working as designed (the "re-verify
before acting" note is exactly what caught this, and re-verification found no semantic drift) — this
is recorded for the pattern, not as an open defect requiring action beyond what's already specified.

**Grading:** likelihood — near-certain to recur every round given the repo is under active
development throughout this retrospective's own writing. Impact — low, this instance (content
unaffected). Complexity to mitigate — none beyond what [^PinnedRepoState] already prescribes.

**Disposition:** noted, risk-accepted as inherent to the discipline already adopted — not a new
required fix, just confirmation the "re-verify before acting" note earns its keep.

## Verified clean this round (§0/§1 scope)

- **R1-1/R1-2 §0 correction block**, all technical claims (`blueEnv` guard line 136, `redEnv` guard
  line 171, `citationPasses` `const` line 139 outside `while` line 148, `judge` assignment line 181
  dereferenced unguarded at line 184) — re-confirmed by direct read of `debate.js` at current `main`
  (`88eb57f`), not merely carried forward from the round-1 ledger entry. **HIGH.**
- **R1-6, footnote [^PR14] diffstat** ("+318/-48, 18 files, 11 commits") — re-confirmed live via
  `gh pr view 14 --json additions,deletions,changedFiles,commits`: returns exactly
  additions:318, deletions:48, changedFiles:18, and an 11-entry commit array. **HIGH**, exact match.
- **R1-4, footnote [^NarrativeSimilarity]** (arXiv:2603.22103, Table 5: r=0.388 vs r=0.461 [19%
  correlation gap], 76.0% vs 75.3% majority-vote accuracy, 71.0% vs 71.7% individual accuracy) —
  the original round-1 verification reached only the abstract; this round's direct fetch of
  `arxiv.org/html/2603.22103v1` reached Table 5 and confirms all four figures exactly as quoted in
  the footnote. **Upgraded to HIGH** (was implicitly HIGH per round-1 ledger but not independently
  re-derived from full text until now).
- **R1-19, footnote [^WisdomCrowds] URL** — `alexanderakm.github.io/projects/wisdom-of-llm-crowd.pdf`
  fetched live this round: resolves to a genuine 672KB PDF (binary/compressed, could not extract the
  quoted sentence from the stream, consistent with the existing MEDIUM grade on the quote itself —
  this round only re-confirms the URL no longer 404s). **HIGH** on "URL resolves"; **MEDIUM**
  unchanged on the quoted sentence's verbatim accuracy (carried forward, self-labeled "search
  synthesis").
- **R1-17's factual premise** ("the harness's `model`/`judgmentModel` knobs select Claude aliases
  only") — confirmed via `commands/research.md` (`--model sonnet|haiku|opus`) and `debate.js` lines
  34/38-39 (`bulk`/`judgment` objects built from bare `model`/`judgmentModel` strings passed straight
  through as `agent()` options, no provider-selection wiring anywhere in the file). **HIGH** — this
  part of R1-17's argument holds even though Finding 1 above shows the conclusion drawn from it is
  miscalibrated.
- **R1-20 header fix** — current header text makes no numeric-ratio claim for lane 3, consistent
  with the round-1 finding of no quantified ratio in `lane-3.md`. **HIGH**, no new claim introduced.
- §1.1's already-existing correct qualitative citation ("same-base-model agents remain more
  correlated than architecturally distinct ones") — unchanged this round, carried forward at its
  existing MEDIUM-HIGH grade per the round-1 ledger.

## Not flagged (considered and set aside)

- §1.4's "15 well-formed gap-pattern files" in the red-auditor's memory store: the live directory
  now contains 18 files (this auditor's own memory has grown across rounds 1 and 2 of this very
  retrospective). This is the same live-source-drift class as Finding 3, but the underlying
  mechanism it demonstrates (the memory loop accreting entries as gaps are found) is the thing the
  sentence is praising, so a stale count here is close to unavoidable churn from the process being
  described rather than a citation error — not raised as a gap distinct from the general
  drift-in-a-live-corpus risk already covered by Finding 3 and [^PinnedRepoState].
