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

GOVERNANCE: setup mirrors law/ read-only into each run's inputs/; humans
promote, ignore, or reverse at review. The bench cannot make binding law alone —
an unappealable precedent is a covert legislature.

WHAT A PROPOSAL IS: A CHANGE TO SCOPE LIMITS, NOT A NEW RULE FROM NOTHING.

A holding is a rule plus the boundary it holds within, and `scope-limits` is
where almost all later work lands. A run rarely discovers a wholly new rule; it
discovers that an existing holding reaches further than it said, or not as far —
that an assumption named in scope-limits does not hold in the case just seen.
So a proposal names the holding, quotes the limit at issue, and states the case
that moves it:

    ## scope: <existing-holding-slug> [PROPOSED]
    limit-at-issue: <quote the scope-limits clause this case tests>
    case: <what happened, concretely, with record citations>
    proposed-change: <the limit as it should now read>
    effect: <which past rulings read differently under the new limit, if any>
    source: <run-slug, sitting, gap/petition id>

A genuinely NEW rule comes from `bench declare`, whose whole purpose is to state
one — its text IS the holding, and it is harvested as PERSUASIVE on that basis.

WHAT IS NOT A PROPOSAL. An ordinary ruling disposing of a gap is not a candidate
holding. Harvesting every ruling produced 76 entries of which 68 stated no rule
at all: 35 whose `holding` was a disposition word (`closed`, `carried`), and 33
carrying a placeholder asking a reviewer to supply the rule the harvest could not
invent. A queue in which nine of ten entries cannot be promoted is not a backlog,
it is noise that hides the entries that can.
