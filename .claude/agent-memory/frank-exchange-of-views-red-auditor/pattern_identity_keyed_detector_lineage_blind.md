---
name: pattern-identity-keyed-detector-lineage-blind
description: An escalation/convergence detector keyed on stable identifiers never fires when the process's own bookkeeping convention renames the tracked object every cycle — check who mints the ids against what the detector matches on
metadata:
  type: feedback
---

When auditing any mechanism that escalates on *recurrence* (contested dockets, retry
escalators, flaky-test detectors, repeat-offender triggers), YOU MUST check whether the
identifier the detector matches on survives the process's own minting convention. If the
convention issues a fresh id per cycle (successor gaps, re-filed tickets, renamed retries),
an identity-equality detector is structurally inert — it will show zero-recurrence telemetry
forever while the underlying dispute recurs indefinitely.

**Why:** FEOV-retrospective round 4 (R4-1). `debate.js`'s contested-docket check is
`prevGapIds.has(g.id)` — pure id string-equality against the prior round. But red's own
closed-WITH-REGRESSION methodology mints a fresh id for every successor gap
(R1-5 → R2-4 → R3-4/R3-9: one footnote, four ids, three rounds). Result: `contested` was 0 in
every round, the judge was never dispatched once across three completed rounds (zero `### LEAD`
sections in the transcript), and the only brake on a spinning debate was the maxRounds cost
ceiling. The debate converged anyway — but because blue conceded in good faith, a property of
the actors, not one the mechanism enforced. The report's own coverage row described only the
narrower same-id-skips-a-round case, whose remedy (widen the id history) does NOT close the
fresh-id case — a fix for one variant masquerading as coverage of the class.

**How to apply:** (1) For any recurrence-triggered control, trace both ends: what key does the
detector compare, and who assigns that key each cycle? If the assigner is the audited process
itself (including RED — your own successor-id practice was the defeat mechanism here), the
audit must treat its own conventions as part of the system under test. (2) Absence-of-firing
telemetry ("judge never invoked," "escalation count 0") is evidence FOR this pattern, not
evidence of health — grep the transcript for the escalation artifact at header/structural
level, not plain text (a quoted phrase is not an invocation). (3) The fix shape is lineage,
not history-widening: a `supersedes: [prior-ids]` field plus chain-depth detection; verify a
proposed remedy closes the variant actually observed, not the adjacent one. (4) Convergence
observed under a broken convergence-enforcer must be attributed to actor behavior, and stated
as such in the disposition.

**Extension — "normalized" asserted, normalization unspecified (sleeper-service round 3,
L5-F3):** a per-cause escalator keyed on a "normalized" failure signature (exit class +
first abort-record line) is inert if the record format the design itself specifies embeds
per-cycle variable content (dated markers, fresh-minted run-dir paths, session ids) and
the normalization is never defined. The unspecified normalization IS the mechanism; demand
it spelled out (strip dates/paths/ids, match on template or error code) plus zero-firing
telemetry surfaced so an inert detector is visible.

Related: [[pattern-missing-root-invariant]] (multi-round gap chains signal a missing stated
invariant — the gap chain here IS the detector's blind spot),
[[pattern-self-defeating-mitigation]] (two controls starving each other's trigger),
[[pattern-policy-without-mechanism]] (docket doctrine present, enforcement never reachable),
[[pattern-repair-regression-citation]] (red-practice-becomes-system-defect sibling).
