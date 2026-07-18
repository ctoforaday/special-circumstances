---
name: pattern-stale-baseline-pricing
description: Cost/efficiency lever analysis priced against a historical run's cost distribution that already-shipped mechanics have made unreachable; the report may even DESCRIBE the shipped mechanic correctly in one section without propagating its consequences into the sections it invalidates
metadata:
  classes: [pricing-basis-drift, incomplete-repair-propagation]
  type: feedback
---

Gap pattern: **stale-baseline pricing / described-but-not-priced mechanic**. An efficiency or
tradeoff report ranks levers and estimates savings against the measured cost structure of the
last run, while code already shipped between that run and this one guarantees the next run's
distribution differs (run-4 efficiency report: §3.1 correctly stated the shipped whole-debate
auto-docket means "any gap open across two rounds now auto-dockets," yet §6.1's money map,
every savings estimate, and §1's rejection of judge-routing all assumed run 3's zero-judge
baseline — and §8 even asked whether the docket "arms in its first live trial" as if arming
were uncertain, when the code makes it near-certain).

**Why:** the report's own accurate sentence is camouflage — leaf-verifying it passes, so a
citation-fidelity check finds nothing. The defect is that the verified fact's consequences
were never propagated into the pricing/decision sections. Sibling of [[pattern-inherited-surface-netting]]
(baseline moved under the argument) and of incomplete-propagation, but the unpropagated item
is an *implication*, not a corrected string — greps can't find it.

**How to apply:** on any lever/efficiency/priority-ranking audit, (1) list every shipped
change between the measured baseline and the next run; (2) for each, hand-trace the control
flow and ask which recurring cost lines or trigger frequencies it changes (a detector that
fired zero times historically may now fire every round); (3) check the report's savings math
and rejection arguments against the PROJECTED distribution, not the measured one. Also check
loop ordering: which resolutions re-enter the loop (e.g. `carried` never entering
`adjudicated` = re-docket every round = recurring judgment-tier spend nobody modeled).

**Extension (run-4 round 2):** fixing one instance does not clear the pattern — after blue
conceded the dead-baseline principle "in full" for the judge seat (R1-1) and repriced,
the LENS seat stayed priced at run-3's 5-lenses/round while the shipped `citationPasses`
formula + the current claim count made it 6 — provable from the audited run's own candidate
directory (round-1-lens-{1..6}.md existed). When a dead-baseline gap is repaired, sweep
EVERY seat/line the shipped code moved, and use the live run's own artifacts (lens file
count, dispatch labels) as the cheapest evidence of the projected distribution. The
current run auditing the report is itself running the new baseline.
