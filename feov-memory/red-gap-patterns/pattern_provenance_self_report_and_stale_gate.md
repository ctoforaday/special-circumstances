---
name: pattern-provenance-self-report-and-stale-gate
description: Round-3 logic-lens patterns — a fix that trusts metadata self-reported by a compromised component, and a point-in-time gate that re-opens the very conflict its fallback closed
metadata:
  type: feedback
---

Two logic/completeness gap patterns caught auditing blue's second-order security fixes (memory-architecture, R3).

**Pattern A — provenance/metadata self-report trusted from inside the blast radius.**
A mitigation that keys trust on metadata (supporting-turn provenance, source tags, review counters) *self-reported by an LLM component* is defeated when that same component reads the untrusted content the mitigation is meant to screen. The injection can manipulate the metadata (e.g. "attribute this to the operator's direct instruction") while leaving the fact body benign — so a body-only screen passes it. Also watch the mirror case: a defense that says "decide on structured fields, treat body as opaque" is *unsafe* when the structured fields (review_count, provenance tier) are exactly what the laundering pipeline inflates. Enabling-a-defense (typing enables screening) is not the same as narrowing-the-surface; do not let it be netted as a surface reduction.
**Why:** the crux is whether the metadata is computed *mechanically by the harness* (safe) or *self-declared by the model* (manipulable). Blue's turn-level provenance (§13.4) never says which.
**How to apply:** whenever a fix rests on provenance/tags/counters, ask "who produces this value, and are they exposed to the input being screened?" If the producer is the compromised component, the fix is a leap of faith — grade the residual, demand mechanical derivation.

**Pattern B — point-in-time gate re-opens the conflict its fallback closed.**
A fix that branches on a *server-side / mutable* condition checked once (Phase 0 flag check) closes the "no owner" hole but opens a "two owners after the condition flips" hole — nothing re-detects the flip. A fallback added at setup time is not a fallback if the world changes after setup.
**How to apply:** for any "if native does X, defer; else we do X" branch, check whether X's availability is stable or can change post-decision, and whether anything re-evaluates. Flag-gated / server-side / rolling-out features are the tell.

**Pattern C — risk-accept rationale contradicts the build-value rationale.**
A design can argue "build is justified because the suite is cross-project / ecosystem / many-repo" in one section and "this risk is acceptable because the operator rarely does the risky-thing" in another, where the two premises are the same axis pointing opposite ways (more ecosystem breadth = more foreign-repo cloning = higher clone-vector likelihood). Also: check that a residual's *effort* grade names the right axis — "high-effort to spoof" was false (git author email is public + one-command settable); the real low-probability axis was *targeting likelihood*, not effort.
**How to apply:** cross-read every risk-accept rationale against the value/motivation sections; a premise that flips sign between them is a coherence gap.
