---
name: pattern-pinned-mapping-not-total
description: a convention pinned for series stability (enum→numeric mapping, reading rule) is never checked for totality over its actual input domain — unmapped tokens, synonym tokens at opposite extremes, unpinned compound-cell handling
metadata:
  type: feedback
---

When a report PINS a mapping or convention to make a measured series comparable across runs,
audit the pin's DOMAIN, not just its existence: enumerate the full input population the
instrument will actually see and check every member maps.

**Why:** efficiency-investigation round 4, L5-F3 — the freshly-decided Q6 mapping (low=1 …
certain=3.5, realized=excluded) was pinned per the lead's ruling, but (a) the shipped GRADE
enum has eight members and `trivial` was unmapped (schema-legal as likelihood/impact);
(b) the corpus's own grading defines `certain` AS "already realized," so two near-synonyms
sat at opposite extremes (3.5 vs excluded) and the exclusion rationale ("no longer a
probability") applied verbatim to the token kept at max weight; (c) conditional grades
("low this run rising to medium-high") — the board's modal shape — had no pinned reading,
and the report's own two-lane history showed extraction convention moving the series.
Within-version ambiguity defeats the pin the same way the mid-series change it forbids would.

**How to apply:** whenever a mapping/threshold/reading-rule gets pinned (especially per a
judge ruling — adopted text gets less scrutiny than contested text), diff its key set against
the schema enum and the corpus's observed value shapes (compound cells, conditionals,
parenthetical semantics). A synonym pair straddling the mapping's extremes is the
max-magnitude instance. Related: [[pattern-unreconciled-numeric-floors]] (requirements that
don't compose), [[pattern-metric-conflation-and-traceable-not-verified]].
