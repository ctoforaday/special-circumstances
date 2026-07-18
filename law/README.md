# law/ — the debate's legal system

The bench keeps NO private memory. Its continuity across runs is entirely
constituted by reviewable text, under an explicit authority hierarchy:

    STATUTE  >  PRECEDENT  >  case-local argument

- STATUTE: the seat constitutions (plugins/frank-exchange-of-views/agents/*.md)
  and engine law (debate.js's mechanical gates). Human-written; amended only by
  ordinary reviewed commits.
- PRECEDENT: the bench's published holdings, in precedents.md. Two-tier
  authority (the sleeper-service promotion gate applied to law):
    * PERSUASIVE — a fresh holding. Citable as ARGUMENT in later sittings,
      never as authority. Every capture-harvested holding starts here, in
      proposed/<run-slug>.md, awaiting review.
    * AFFIRMED — a human reviewed the holding and promoted it into
      precedents.md. Binding on later sittings UNTIL a leaf conflicts (the
      leaf always wins; the conflict is flagged for review) or a human
      REVERSES it (strike it with a dated note — never delete; reversal is
      itself precedent).
- Case-local argument: closings, rebuttals, and party citations of law.
  PRECEDENT IS ARGUMENT, NOT EVIDENCE: the only evidence is the artifact and
  the leaf. A cited precedent MUST be addressed in the ruling opinion; both
  parties may cite and contest law.

PRECEDENT FORM (defeasible by construction — a holding without its factual
predicate is not citable):

    ## <slug> [PERSUASIVE|AFFIRMED <date>|REVERSED <date>]
    facts: <the run-local situation, concretely>
    question: <what the bench had to decide>
    holding: <the rule applied>
    rationale: <principle + values in tension + why one won>
    scope-limits: <what this holding assumed; where it is distinguishable>
    source: <run-slug, round, gap/petition id>

GOVERNANCE: setup mirrors law/ read-only into each run's inputs/; capture
harvests the run's rulings into proposed/ as PERSUASIVE; humans promote,
ignore, or reverse at review. The bench cannot make binding law alone —
an unappealable precedent is a covert legislature.
