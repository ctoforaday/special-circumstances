---
name: pattern-policy-without-mechanism
description: An invariant/policy asserted as self-enforcing ("just a removal of trust") while the concrete enforcing artifact was withdrawn in a prior round and never replaced
metadata:
  classes: [policy-without-mechanism, self-attestation]
  type: feedback
---

**Pattern: policy-without-mechanism (invariant asserted as self-enforcing after its enforcer was withdrawn).**

When a design converges on a clean *organizing invariant* to replace N spot-patches, verify the
invariant names a concrete enforcement mechanism at the exact moment it must hold — do not accept the
policy statement as the fix.

**Why:** In the memory-architecture debate, three rounds hollowed out the clone-injection enforcer:
§12.2 had a concrete git-ignored ratification marker → §13.2 replaced it with a git-authorship check
→ §14.2 demoted authorship to "nudge-convenience, not activation" → §14.1 asserted the *outcome*
("foreign clones load clamped to reference tier") with **no** stated gate, framed as "not new
machinery — a removal of trust." But the committed `active.md` projection is loaded by **native
`@`-import at session open**, before any bespoke process runs. Nothing bespoke can clamp a native
import; a generator-side property (de-authorized voice) does not touch attacker-authored committed
bytes; a SessionStart hook adds context, it cannot un-import. Enforcing the invariant actually
REQUIRES new machinery (git-ignore the projection + regenerate locally). "Removal not machinery" was
the leap of faith.

**How to apply:**
- When blue adopts a unifying invariant, trace it to the leaf: what artifact enforces it, and does
  that artifact run *before* the untrusted bytes reach context? If enforcement happens "at next local
  re-derivation" but the untrusted artifact is loaded natively before then, the invariant is policy,
  not mechanism.
- Check the *withdrawal chain*: if a prior round had a concrete (even flawed) enforcer that got
  withdrawn/demoted, confirm a replacement mechanism was carried forward — not just the *goal* the
  old enforcer served. Enforcers get hollowed round-over-round while the goal-language persists.
- "Removal of trust" / "cheaper than machinery" framing is a tell — verify the removal is
  self-enforcing and doesn't silently require a new gate.

**Extension (efficiency-investigation run 4, round 2):** the class recurs *within the round that
fixed it*. Blue's R1-10 repair correctly named the no-filesystem constraint ("emit into cost.md" is
impossible) in §2.5 — while the same round's R1-6 repair in the adjacent §4.5 specified a
reconciliation check that "throws" on a line-count mismatch the script cannot compute (no fs access).
Heuristic: when a round's repair correctly applies constraint X at site A, grep the SAME round's
other new mechanisms for X-violations — shared authorship + time pressure elevate the base rate, and
the fixed site creates a sibling halo over the unfixed one. Companion root invariant surfaced same
pass (the *attestation ceiling*): an engine whose state rides self-reported envelopes has no
primitive stronger than self-report for work-done claims — schema checks catch omission,
cross-referenced independent structures (lineage-throw style) catch inconsistency, nothing in-run
catches vacuity ("required non-empty" ≠ "work performed"); the honest enforcement tiers are
in-run shape/cross-ref checks + post-run independent audit over git-tracked artifacts, and a repair
claiming "fails structurally, not silently" for a bare non-empty field overclaims. Also watch:
repairs that route the audit of a conflicted seat's self-report back to the same seat's own
spot-check floor (found_by sampled by red-merge; accepted-dispute deltas spot-checked by the
accepting merge) — "auditable" doing the work "audited by a named independent consumer" should do.

**Extension (run 4, round 2 — assumed-durable logging):** telemetry/instrumentation ratified with
an ASSUMED sink is the same class. Blue's repaired §2.5 named "log() into trajectories/journal.jsonl,
consumed by cost-audit.mjs" — but the prior run's journal.jsonl (the only measured one) holds ONLY
started/result lifecycle events (zero log() lines; grep the script's own log strings against the
journal), and cost-audit.mjs has zero journal references (it parses harness transcripts). Check
recipe: (1) grep the emitting script's literal log strings in the prior run's journal; (2) grep the
named consumer for the sink's filename. A "zero-token instrumentation" recommendation whose lines
persist nowhere produces no evidence base — and the repair that introduced the claim was itself the
fix for a prior sink error (repair relocated the defect: [[pattern_repair_regression_citation]]).

**Related, same pass:** *supersession-accounting drift* — a headline count ("5 blocking") goes stale
when a superseding row changes a grade (item 29 Blocking supersedes item 22 High → true count ~6);
re-derive counts from the operative rows, don't trust the verdict tally. And *headline-lag* — a
template section (Heilmeier §0) keeps marketing a feature a downstream concession (auto-promotion →
near-empty-set convenience) has since gutted. Links: [[pattern_self_defeating_mitigation]],
[[pattern_missing_root_invariant]] (this is the failure mode of *adopting* the root invariant red
asked for — the invariant is right but its enforcement is unspecified).
