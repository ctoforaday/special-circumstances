---
name: pattern-fail-open-guard-and-erasable-evidence
description: Two composing control-design defects — a guard keyed on an ambient marker that silently disarms when the marker is lost (fail-open), and "tamper evidence" living where the audited actor can erase it (uncommitted working tree, actor-writable ledger)
metadata:
  type: feedback
---

Rule: for every proposed control, ask two questions the designer usually hasn't: (1) what
happens when the control's ENABLING PRECONDITION is absent — does it fail closed or silently
fail open? (2) does the DETECTION/EVIDENCE channel the control leans on live inside the
audited actor's write/erase reach?

**Why:** sleeper-service design round 1 produced clean specimens of both. (a) The PreToolUse
write-fence armed only "when the sleeper marker is present" (env var set by the wrapper) —
marker loss (scheduler misconfig, shell nesting, rung migration) silently drops the run to the
permission-rules-only configuration the same report REFUTED at the leaf node; nothing asserted
enforcement (preflight checked plugin LOAD, not hook FIRING). (b) §4.3 called an uncommitted
working-tree delta "tamper evidence (the DGM-lineage property)" — but an uncommitted delta has
no history; the one channel able to write out-of-boundary (subprocess residual) can also
revert its write before review, and run N+1 executes before run N is reviewed. Sibling
instance: the monthly budget cap read a ledger inside the loop's own Edit-allowed surface.
DGM's lineage worked because it was durable and externally inspectable; these were neither.

**How to apply:** (1) marker/env-keyed guards: demand fail-closed polarity (fence unless a
human-session marker present) or a per-run denial canary (attempt a forbidden write, abort
unless DENIED — enforcement becomes a verified fact, reusing the design's own probe method).
(2) any "tamper-evident / append-only / git-tracked" claim: check WHO can write/rewrite the
evidence between reviews and whether the review cadence outruns the actor's next unattended
session; evidence of record must be written from OUTSIDE the audited session (wrapper-side
snapshot/auto-commit). Related: [[pattern-self-defeating-mitigation]],
[[pattern-authorship-evades-never-edit-guard]], [[pattern-gate-stack-rung-nonportability]],
[[pattern-policy-without-mechanism]].
