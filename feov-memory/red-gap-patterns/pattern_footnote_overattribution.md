---
name: pattern-footnote-overattribution
description: A footnote's claim-list bundles several specifics but only some trace to the primary; multi-citing one figure to N footnotes hides which one carries it
metadata:
  classes: [citation-figure-misattribution, derivation-status-overclaim]
  type: feedback
---

When a footnote lists several distinct specifics (e.g. "summarization drift; semantic
intensification; cross-version score drift; ~29-day half-life") treat each as a **separate**
statement↔source pair — verify them independently. Frequently only one leg (the generic
qualitative one) is corroborable at the primary; the specific quantitative legs are asserted.

**Why:** caught in memory-architecture Round 3 — `[^MemorySurvey]` (arXiv 2603.07670) claimed four
things; leaf-node fetch confirmed only summarization drift. The load-bearing ~29-day half-life
(sole prop for "decay windows are in the evidenced band") was uncorroborable.

**Compounding trick — multi-citation laundering:** a figure cited to three footnotes
(`[^A][^B][^C]`) reads as heavily-sourced, but often only ONE of the three actually carries it and
the other two are topical padding. Check WHICH footnote's text asserts the number; the number's
real support may be a single un-pinnable source wearing a crowd.

**How to apply:** for any bundled-claim footnote or multi-cited figure, name which specific source
must carry the load-bearing number, follow only that one, and grade the number on that source
alone. Distinguish "unable-to-corroborate (lossy fetch)" from "contradicted" — the former is a
graded low-confidence gap under the stickler rule, not a pass. Relates to
[[citation_status_and_misattribution_patterns]] and [[repair-regression-citation]].
