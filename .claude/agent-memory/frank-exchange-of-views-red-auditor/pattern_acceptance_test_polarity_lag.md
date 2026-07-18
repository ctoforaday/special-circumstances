---
name: pattern-acceptance-test-polarity-lag
description: A repair inverts a mechanism's polarity but the standing acceptance test still asserts the OLD behavior — the gate becomes unsatisfiable and its cheapest pass is undoing the repair
metadata:
  type: feedback
---

A repair changes what a mechanism DOES (e.g. `/self-improve` becomes an inert trampoline;
the payload moves to the wrapper), but an acceptance test adopted earlier — often quoted
from an external plan — still asserts the old polarity ("headless `claude -p '/self-improve'`
produces a run dir"). Post-repair the test FAILS BY CONSTRUCTION, and the perverse incentive
is sharp: the cheapest way for a builder to make the printed gate pass is to re-inline the
payload, silently undoing the repair.

**Why:** round-5 sleeper-service (R5-2, sup R4-1). The R4-1 trampoline made the port plan's
Phase-4 verify step unsatisfiable; §3.3 still said it "must remain the acceptance test," and
§3.4's ladder cell still described the command markdown as the payload. Neither site was in
the repair's edit list or propagation greps.

**How to apply:** after any repair that inverts or relocates a mechanism, hunt the standing
ACCEPTANCE TESTS and gate definitions that reference it (grep the mechanism's old name and
the test's promised observable). The repaired design's correct test is usually TWO-legged:
(i) the new path produces the promised artifact; (ii) the OLD path now produces NOTHING —
the repair's own property becomes a verifiable assertion. A test that only exercises the new
path misses the regression channel; a test that still asserts the old path incentivizes
reverting the repair. Distinct from stale gate-survival tables: here the defect is the
test's POLARITY, not its rung coverage.
