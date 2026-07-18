---
name: pattern-within-source-condition-misattribution
description: A correctly-cited paper's headline result quoted with a gloss that silently reassigns it to a weaker experimental condition/arm — especially to justify NOT building the stronger condition; check which arm carries the number
metadata:
  type: feedback
---

**Extension (2026-07-17, sleeper-service run):** the cross-source variant — *scope-fusion
overstatement*. Two separately-true claims from different sources with different scopes get
fused into one sentence claiming the union scope (e.g. "reported by red, blue, AND judge
across two consecutive runs" where three-seats is attested for one run and two-runs is
attested for red merges only). Check each conjunct of a compound recurrence/attestation claim
against its own source's scope; the fused sentence must be no stronger than the weakest
component's scope.

When a report cites a real, verified figure from the right paper to support a tradeoff
disposition, YOU MUST check *which experimental condition/arm* the figure belongs to — not just
that the number appears in the source.

**Why:** FEOV-retrospective round 2 (R2-3): blue's R1-17 fix deferred cross-provider model
diversity by citing arXiv:2602.03794's "2 diverse agents match/exceed 16 homogeneous" — with a
bracketed gloss "[persona-lensed]" that silently reassigned the result to the paper's L2
condition (persona-only, same base model). The paper's own Table 2 makes it the **L4** result
(different models AND personas); L2 needs 8 agents to match the 16-agent baseline — a 4x
efficiency gap. The citation was real, the figure exact, the round-1 ledger even had the
qualitative claim at HIGH — and the disposition still misread it, because the gloss moved the
number to the arm the report wanted it to support. Telltale: the citation was deployed
specifically to justify *not building* the condition (L3/L4) that actually produced the number.

**How to apply:** Whenever a quoted result carries a bracketed insertion, paraphrase, or
condition label supplied by the citing text ("[persona-lensed]", "same-provider", "without X"),
re-fetch the source's taxonomy/results table and pin the result to its arm. Highest suspicion
when the cited result underwrites a defer/skip decision about the very dimension the source
varied. Distinct from miscitation-to-wrong-paper ([[citation_status_and_misattribution_patterns]])
and from metric conflation ([[pattern-metric-conflation-and-traceable-not-verified]]) — here
paper, number, and quote are all correct; only the condition attribution is wrong.

Related: [[pattern-repair-regression-citation]] (this instance also lived inside a round-1
repair), [[pattern_footnote_overattribution]].
