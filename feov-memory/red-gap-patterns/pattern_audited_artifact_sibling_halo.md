---
name: pattern-audited-artifact-sibling-halo
description: Blue audits ONE finding of a pinned artifact as defective, then transcribes a SIBLING finding's own error from the same artifact unchecked — the partial audit confers unearned trust on the rest of the source
metadata:
  classes: [audited-artifact-sibling-halo, figure-miscomposition]
  type: feedback
---

When a report demonstrates it audited a pinned artifact (e.g. flags cost.md finding 2 as
internally contradicted), do NOT let that demonstrated diligence stand in for verification of the
artifact's OTHER findings quoted elsewhere — recompute each transcribed figure independently.

**Instance (run 4, round 1, L3-F4):** blue's §6.4 correctly flagged run-3 cost.md finding 2
("merge cost tracks dispute size" vs its own table), but §4.2 transcribed finding 3's "12.5×
cache-write" as a multiplier — the artifact's own pricing header (2.5 vs 12.5 $/MTok) makes the
ratio 5×; 12.5 is the absolute rate. Units-vs-ratio confusion inherited verbatim from the source.

**Why:** a partial audit reads as "this source was checked"; the halo suppresses re-derivation of
sibling claims. Related to [[measurement-methodology-drift]] (citation-faithful ≠ sound) and
[[pattern-risk-grading-conflations]].

**How to apply:** whenever a report flags defect(s) in a source it also cites approvingly,
enumerate every OTHER figure the report takes from that same source and recompute each from the
source's own primitives (pricing headers, tables, totals) — the flagged defect raises, not
lowers, the prior that siblings are wrong.
