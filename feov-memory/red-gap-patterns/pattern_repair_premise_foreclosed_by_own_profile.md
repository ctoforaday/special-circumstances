---
name: pattern-repair-premise-foreclosed-by-own-profile
description: A guard added to tag/bound a threat write rests on a premise the same document's allow/deny profile already forecloses — the guard is dead where aimed, and the foreclosed write was a REQUIRED behavior of a subsystem, now silently denied
metadata:
  type: feedback
---

When a repair claims "actor X writes surface Y, so we tag/bound that write," check
whether X's write to Y can actually EXECUTE under the design's own printed
allow/deny/fence surface. If every layer forecloses it, two texts cannot both be true:
the repair's justifying premise (the write happens) and the profile (the write is
denied).

**Why:** sleeper-service round 4 (L6-F2): R3-3 built snapshot-diff machinery to
provenance-tag gap-pattern memory "the nightly red-merge seat writes" — but the fence
(writes outside research/+ideas/ denied), `Edit(<REPO>/.claude/**)`, and the marker-loss
canary abort forecloses that write at three layers. The machinery was dead where aimed.

**How to apply:** for every new guard, trace the guarded write path against the profile
as printed. Two horns, both auditable: (a) denies are intended → the REAL finding is the
unstated side effect — the foreclosed write was a required behavior (here: the red
seat's MUST-record-patterns clause fails-denied nightly; routine denials pollute the
tamper-evidence record) and the guard's live remit shrinks to a different surface than
its text describes; (b) the write is meant to succeed → an undeclared allow contradicts
the stated write-surface invariant. Related: [[pattern-doctrine-vs-implementation]],
[[pattern-self-defeating-mitigation]], [[pattern-authorship-evades-never-edit-guard]].
