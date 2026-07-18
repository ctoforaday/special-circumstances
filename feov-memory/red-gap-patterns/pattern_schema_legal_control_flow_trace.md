---
name: pattern-schema-legal-control-flow-trace
description: dark-side/risk lens finds its best gaps by hand-tracing control flow for schema-legal-but-semantically-incoherent envelope shapes, not by re-checking citations
metadata:
  classes: [unhandled-degenerate-case, partial-control-coverage, false-universal]
  type: feedback
---

When auditing a script/harness whose keystone facts (guards, routing, ledger text) have already
been re-verified clean across multiple rounds by leaf-node citation lenses, the dark-side/risk
lens's marginal value shifts from "re-check the same facts again" to **hand-tracing runtime control
flow for inputs that satisfy the schema but are semantically degenerate**. Two recurring
sub-patterns, both caught this way in the FEOV retrospective round 3 (`round-3-lens-5.md`):

1. **Schema-legal-but-incoherent envelope shapes.** A schema requires fields to exist and have the
   right type, but rarely enforces cross-field coherence (e.g. `verdict: 'FAIL'` with `gaps: []`).
   Trace what the *loop*, not the *schema validator*, does with that shape by hand, for several
   rounds. Look especially for paths where a guard/branch is gated on a derived condition (e.g.
   `contested.length > 0`) that a degenerate-but-valid input makes permanently false — this
   silently disables an entire adjudication/escalation path rather than throwing.
2. **"Never dropped" / "always captured" claims about telemetry plumbing** (friction arrays, gap
   ids, provenance tags) are asserted from *design intent*, not from grepping every call site that
   *should* invoke the aggregator. Grep for the aggregator function itself (e.g. `takeFriction(`)
   and enumerate every call site against every schema'd seat that *could* produce that signal — an
   asymmetry (3 of 4 schema'd seats wired, 1 silently omitted) is exactly the kind of omission a
   passing regression suite won't catch if the suite's own test list mirrors the same incomplete
   site list.

**Why this works when citation lenses have already gone clean**: citation lenses re-verify facts
*about the world* (a paper says X, a repo state is Y); this class of gap lives entirely inside the
system's own internal data flow and requires no external reference at all — it is caught by
reading the script like an interpreter would, not by fetching anything. Best done specifically in
round 3+ of a debate, after the obvious citation/doctrine gaps are exhausted and the marginal
citation-recheck yield has dropped.

See also [[gap_live_source_drift]] (re-pin against the *current* HEAD before starting — in this
case the repo had advanced to a 3rd commit past the report's last pin, still docs-only, confirmed
via diff before trusting any prior-round claim) and [[pattern_self_defeating_mitigation]] (a
control whose own trigger condition can be permanently disabled by a degenerate input is the same
family as a mitigation defeating itself).
