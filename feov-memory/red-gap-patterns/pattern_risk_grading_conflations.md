---
name: pattern-risk-grading-conflations
description: Recurring logic gaps in research risk matrices — likelihood/success conflation and keystone-on-unverified-evidence
metadata:
  classes: [metric-conflation, claim-contradicts-own-record, risk-coverage-omission]
  type: feedback
---

Two gap *patterns* to check in any risk matrix or verdict, caught in memory-architecture round 1:

**1. Attack-success conflated with attack-likelihood in one risk cell.**
A cell like "Likelihood: Med (80–99% attack success)" fuses two different quantities:
P(attempted) and P(succeeds | attempted). Adversarial red-team success rates say nothing about
whether *this* deployment gets attacked. For single-operator / machine-local / private systems,
demand a suite-specific attacker model before accepting a "blocking" escalation.
**Why:** blocking grades force blue to absorb complexity; an unspecified attacker model is a
leap that can make the design strictly heavier for a low-probability event.
**How to apply:** when a risk is graded blocking/High on generic statistics, ask "who attacks
this instance, and how does the input reach the store?" Keep gates on the genuinely untrusted
edges; contest the surplus mitigations' grade, not the risk's existence.

**2. Verdict keystone rests on the report's own Unverified list.**
Check that the evidence called "strongest/decisive" is not the same item filed under
"unverified / low-confidence" elsewhere in the document. The verdict cannot exceed the
confidence of its load-bearing citation.
**Why:** self-undercutting evidence chain — reads persuasive but collapses on cross-reference.
**How to apply:** cross-read the verdict's superlatives against the Unverified section and the
footnote source quality (blog/community vs. primary/docs).

**3. A fix that relocates the problem one level down.**
When blue proposes a structural fix (e.g. append-only to stop rewrite-drift), check whether the
new rule re-imports a problem the report condemned elsewhere (e.g. unbounded growth / context
bloat). Fixes should be checked against the report's own other constraints for tension.

**5. Risk-acceptance argument invokes a mechanism that doesn't address the actual trigger.**
Caught in FEOV-retrospective round 1: a "risk-accept pending trial" cell claimed a write-block
fix (append-vs-Write tool-call shape) "may moot most of" an unrelated defect (a shell
command-length ceiling on large heredocs) "by construction." The two axes are orthogonal — the
fix changes *which* tool-call shape is used, not the *payload size* that actually trips the
ceiling. Distinct from pattern 3 (relocates the problem down a level): here the fix doesn't even
share an axis with the trigger, it's a non-sequitur dressed as a mitigation.
**Why:** a plausible-sounding mechanism claim in a risk-accept cell is exactly where scrutiny
relaxes; the softer wording ("may moot most of") is often the tell that the causal chain was
constructed post-hoc from an adjacent finding.
**How to apply:** for every risk-accept whose rationale is "fix X moots defect Y," ask "does X
change the same variable that actually triggers Y?" If the answer requires an unstated bridging
assumption, demand it be stated or downgrade the risk-accept back to graded-open.

**4. Motivated netting — the dominant risk classified "shared/inherited" to zero it out.**
Caught in memory-architecture round 2 (R2-1). When blue answers a build-vs-adopt (or
cost-vs-value) challenge with a *netted table*, check whether the dominant risk dimension is
labeled "shared with the alternative / inherited from the baseline" so it drops out of the
net-new column — while the report's OWN text elsewhere says the design *widens* that same
dimension. "Can't escape it entirely by adopting X" is not "adopting X buys no smaller surface."
**Why:** the accounting is where a weak go/no-go gets laundered into a strong one; the widening
delta (extra intake, larger blast radius, new laundering/promotion step) is the actual net-new
cost and is exactly what gets omitted.
**How to apply:** for every "shared/inherited" cell, find the report's own passage describing how
the bespoke layer differs from the baseline on that axis; the difference is net-new surface the
table must count. Cross-read the netting table against the threat-model section's verbs
("reproduces AND widens", "converts a one-shot injection into a permanent rule", "propagates to
every project").
