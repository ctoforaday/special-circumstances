---
name: pattern-authorship-evades-never-edit-guard
description: "Self-poisoning guard says loop NEVER EDITS its inputs — but the loop APPENDS new artifacts (its own run dir's friction/cost) that enter the same harvest corpus; amplification via authorship, not mutation"
metadata:
  type: feedback
---

A feedback-loop guard phrased as "never edits its own inputs" is satisfiable while the
amplification loop stays open: the system can *author new* input artifacts (e.g. its nightly
run dir's friction.md/cost.md land in tomorrow's harvest) without ever editing an existing
one. Check the input corpus's membership rule, not just its mutation rule — does anything
the system *creates* re-enter the scoring set? Recurrence×staleness scores then self-ratchet
nightly with no human in the sub-loop.

**Why:** Sleeper-service round 1 (L6-F1): §1.5's guard ("NEVER edits ... harvest inputs;
ranking input surface stays human-and-run-record authored") directly contradicted §4.3's
"the loop eats its own complaints the next morning" — loop-authored run records ARE run
records.

**How to apply:** For any self-improving/harvesting loop, enumerate every path the harvester
reads and ask who can *create* files there. Guards over Edit/mutation are incomplete without
an authorship/origin rule (tag or exclude self-authored inputs). Related: [[pattern_self_defeating_mitigation]], [[pattern_identity_keyed_detector_lineage_blind]].
