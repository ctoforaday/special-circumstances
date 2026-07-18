---
name: pattern-metric-conflation-and-traceable-not-verified
description: Two citation-lens patterns — a "success band" whose endpoints are two DIFFERENT metrics, and "closed-as-traceable" conflating paper-traceability with digit-verification
metadata:
  type: feedback
---

Two leaf-node citation patterns caught in FEOV round-4 lens-3 (memory-architecture report).

**Pattern A — metric-conflation into a false range.** A source reports two *distinct* measures
(e.g. MINJA: 98.2% *injection* success rate vs 76.8% *attack* success rate). One section states
them correctly and separately; other sections collapse them into a single band ("succeeds
~76.8–98.2%"). The band's two endpoints are different quantities, so the upper bound reads as a
higher *attack*-success observation when it is actually a different metric. **Why:** blue repairs
the primary occurrence but propagates a lossy paraphrase elsewhere. **How to apply:** when a cited
"range" spans two round numbers, check whether both endpoints measure the *same thing*; a band built
from ISR-low + ASR-high (or precision-vs-recall, latency-vs-token) is imprecise even when both
digits are individually correct. Grade LOW if it doesn't move the disposition; require relabel.

**Pattern B — "traceable" ≠ "digit-verified."** A citation marked *closed-as-traceable* in a prior
round may rest on paper-level traceability (right paper, right title) while its *specific number*
still sits behind PDF-table friction. Re-verify the digit, not the paper. Corollary: **arXiv
abstract-version drift** — the `/abs/` page and v1 HTML may lack numbers that appear in the v2 HTML
abstract or a results table. If a fetch returns "no percentages in abstract," try `/html/<id>v2`
before concluding the number is unsourced. In the case that spawned this note the digit *did* check
out via v2 HTML after `/abs/` and the 7.7MB PDF both failed — friction delayed, didn't falsify.

Related: [[citation_status_and_misattribution_patterns]] (lossy arXiv-HTML fetch),
[[pattern_footnote_overattribution]], [[pattern_repair_regression_citation]].
