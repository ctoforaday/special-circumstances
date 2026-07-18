---
name: pattern-missing-root-invariant
description: When successive rounds each patch one security gate and each patch spawns a next-order gap, the root cause is a missing stated invariant — surface the invariant, not just the Nth instance
metadata:
  classes: [missing-root-invariant]
  type: feedback
---

When a design accretes security through gate-by-gate patching and *each round's fix introduces the
next round's gap* (R1 fix → R2 gap → R3 gap on the same axis), the recurrence itself is the finding:
the design lacks a single stated **invariant** from which the individual gates would follow by
construction.

**Why:** memory-poisoning defense in the 2026-07 audit went three rounds — trust tiers → clone gate
→ authorship gate; provenance-of-record → provenance-of-content → turn-level taint. R3-3 (taint
under-propagates) and R3-4 (consolidator must read the surface it guards) both collapse into one
rule: *"external-touched ⇒ tainted, transitively, until a human clears it."* Grading the Nth patched
gate in isolation lets blue keep relocating the hole; naming the missing invariant forces the design
to derive the gates instead of bolting them on.

**How to apply:** when you catch a third-order failure of the same first-order risk, (1) still grade
the instance (red never soft-passes), but (2) additionally surface — as a design-coherence
recommendation to the lead, not a block — that the axis needs a stated invariant. Watch the severity
trend: *declining* severity across rounds = convergence, so frame it as a recommendation; *flat or
rising* severity = the patching is not converging and the invariant becomes a block. Distinguish this
from [[pattern_self_defeating_mitigation]] (per-instance) — this is the meta-pattern across instances.
