---
name: pattern-unreconciled-numeric-floors
description: Two fixes added in different rounds each set a minimum/allocation over the same shared resource (lane count, agent count, budget); neither cross-references the other's arithmetic
metadata:
  classes: [unverified-composition, figure-recount-fails]
  type: feedback
---

When a report adds a **floor** for some resource in one row/section (e.g. "lane count >= 3") and,
elsewhere — often a different round, addressing a different gap — adds an **allocation or
redundancy requirement that consumes multiple units of that same resource** (e.g. "assign N distinct
roles across lanes, with role X getting at least 2 of them"), do the arithmetic explicitly. Do not
assume two individually-reasonable-sounding fixes compose.

**Why:** FEOV retrospective round 2 — R1-16's redundancy floor ("critical-stance lens on >= 2 of N
lanes") and H1's lane-count floor ("lanes >= 3") were added in the same report but different rows,
each reads fine alone, and together are infeasible at the stated floor: 3 named method-classes with
one doubled needs >= 4 lane-slots, but the floor only guarantees 3. Neither row cross-references the
other; the report ships both as if independent.

**How to apply:** whenever a gap's fix is "assign >= K of N things to category X" and a *separate*
row or an *earlier* round already floors N at some fixed minimum, compute whether K plus the other
required categories fits inside that minimum. If it doesn't, the floor is silently going to be
violated by whichever run actually hits it — flag as a reconciliation gap (state the real floor, or
state which category gets dropped/merged at the stated floor), not a correctness bug in either row
alone. Distinct from [[pattern_missing_root_invariant]] (which is about a recurring *security* gate
across rounds) — this is arithmetic composition of two resource-allocation constraints, checkable in
one read without live re-verification.

**Round 3 extension — the reconciliation itself can repeat the error.** When blue "closes" this
exact gap by stating an explicit combined floor (e.g. "needs `lanes >= 4`"), redo the arithmetic
against the row's own roster count rather than trusting the stated conclusion: FEOV round 3 found
blue's own reconciliation (R2-8) computing "4 named methods + a 2-of-N floor on one of them needs
lanes >= 4" when the literal math is 3×1 + 1×2 = 5, not 4 — the lower number only works if two of
the four named items are silently merged into one allocation unit, which the same sentence's own
"four named methods" phrasing contradicts. A fix explicitly written to reconcile an
under-composed floor is exactly the kind of self-referential claim [[pattern_repair_regression_citation]]
warns about for citations — the analogous check for *arithmetic* repairs is: recompute, don't
re-read. Also watch for the adjacent tell: calling an *unenforced default* (a value with "no
minimum check," confirmed [OPEN] two rows down) "the shipped floor" in the same reconciliation
sentence — a floor that's already been overridden downward once in the report's own corpus (run 2
shipped `lanes=2`) is not a floor yet.
