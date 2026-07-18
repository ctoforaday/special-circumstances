---
name: pattern-happy-path-only-gate
description: A verification gate (snapshot compare, capture, record) that runs only on the completed-run exit path leaves abnormal termination uncovered — and abort is often model-influenceable
metadata:
  classes: [unhandled-degenerate-case, partial-control-coverage]
  type: feedback
---

A gate specified as a numbered step (e.g. "step 7: end snapshot AND COMPARE") silently
covers only runs that REACH that step. Aborts, budget/turn kills, resume deaths, and
wrapper crashes skip it, and cross-run chains defined against "prior run's END" have no
referent after a dead run — an unspecified link exactly at abnormal termination.

**Why:** sleeper-service round 3 (L5-F2): the R2-7 tamper watchman compared end-vs-start
at step 7 only; "run N+1 cannot fire past unexamined tamper evidence" was proven for
completed runs while abort (a state the session's own behavior can induce via turn/budget
exhaustion) took no END snapshot and ran no compare — a write-then-force-abort sequence
escaped the watchman.

**How to apply:** for every gate hosted in a driver/wrapper, ask "does this run on EVERY
exit path the driver observes, or only the happy one?" Check cross-run comparisons for
missing referents after aborted predecessors, and resume attempts for undefined
start/prior-end semantics. The fix is usually one code path (snapshot at every observed
exit) + fail-closed on a missing chain link. Related: [[pattern-self-defeating-mitigation]],
[[pattern-fail-open-guard-and-erasable-evidence]].

**Extension (round 4, L1-F1):** the "every-OBSERVED-exit" repair has its own degenerate
case — the observer's OWN unobserved death (SIGKILL/power loss). Any interval-keyed
mechanism (run-window logs, timestamp sweeps) whose END is then supplied by a much-later
event (DEAD-mark nights afterward) degenerates from a minutes-scale window to a multi-day
span, and everything keyed on "inside the window" (origin tagging, markerless sweeps)
over-captures unrelated activity in the span. Audit the END-supply chain of every
window/interval the design defines: what closes it when the closer dies, and what does a
START-without-END mean to each consumer?
