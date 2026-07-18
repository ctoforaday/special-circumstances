---
name: pattern-controller-lookahead-pricing
description: Counterfactual savings for a throttle/controller priced on data only observable AFTER the controlled round — check what board/state the mechanism could actually read at decision time
metadata:
  classes: [pricing-basis-drift, derivation-status-overclaim]
  type: feedback
---

When a report prices a counterfactual controller (throttle, floor, stop rule) as "would have
fired in rounds X–Z," re-derive WHICH state each firing decision reads: a round-N control
decision reads the post-round-(N−1) state, never round N's own output. Caught round 4 (run-4
efficiency investigation, L1-F1): "3 throttled rounds (the low-mass rounds 3–5)" labeled
round 3 throttled on the strength of round 3's OWN post-mass (~44), when the actual throttle
input — the post-round-2 board (~65) — was the run's second-highest mass. The 3-round basis
inflated the ceiling saving into a point estimate ($18 vs honest $12–18).

**Why:** counterfactual replays quietly grant the mechanism hindsight; the off-by-one hides
inside a plausible round-range gloss ("the low-mass rounds") and survives arithmetic checks —
the multiplication verifies while the multiplicand's timing is wrong. Sibling of
[[pattern-stale-baseline-pricing]] (both are pricing-section defects that pass leaf checks) and
of [[pattern-metric-conflation-and-traceable-not-verified]].

**How to apply:** for every "would have saved $Y over rounds X–Z" claim, list the decision
points and the state visible at each; check the report's own series for whether the trigger
actually trips at each decision point under any stated threshold. Also: when red's own prior
required-fix supplied the arithmetic (here R2-2's ×3), audit the repair as a new claim —
red-vector errors are still errors.
