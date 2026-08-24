# Operator brief — sleeper-service-plan

Tracking: FR ctoforaday/special-circumstances#534 (this prompt, recorded) · takes over bug #429.

## The operator's prompt (verbatim)

> ok.  install this plugin then use the FEOV deep research skill - using fable for research and sonnet for the rest.
>
> I want you to come up with a plan for sleeper service.  you'll choose your own research topics, and run them all the way to tracked reports in the target repo, creating tracking bugs and feature requests to do implementations.  We will stop shy of *requiring* full implementations as part of the PRs but the process blue goes through should always require prototyping and testing as verification of its work and recommendations.  one worktrees per research project.
>
> this is most of the way to a fully automated research loop built on FEOV so compare and contrast with other known systems and compare to its parallel counterparts, dreaming and memory.
>
> record this prompt for later reference in a FR to do this research, then generate the research report.  you may use anything previously written on the subject in the repo to provide color, guidance, fill in gaps, and provide alternative approaches.
>
> go nuts.
>
> if we already have a bug on this you may take it over and update it, if not you'll be starting a new one.

## This project's scope (one of two parallel runs)

THIS run produces **the plan**: what the first shipped increment of sleeper-service should be.
In scope: the loop's inputs (the friction corpus — `friction`/`friction-none` events across runs,
envelope harvests; run records; telemetry; `law/proposed/` rulings), topic-selection policy, cadence
and scheduling (headless `claude -p`, cloud routines), the promotion ladder
(insight → MEMORY → rule-skill → cheatsheet) and its human gates, what `/self-improve` and
`/graduate` concretely do, and the ship-vs-withdraw resolution for issue #429.

OUT of scope here (the sibling run `2026-08-23_research-loop-counterparts` owns it): the survey of
other automated-research/self-improving-agent systems and the deep comparison with the dream loop
and the memory architecture. Reference them where the plan needs them; do not duplicate the survey.

## Standing process requirement (from the operator, binding on every recommendation)

Full implementations are NOT required as part of resulting PRs — but blue's process MUST verify its
work and recommendations by **prototyping and testing**: where a recommendation rests on a claim a
computation can settle (a hook fires, a flag exists, a schedule syntax parses, a binary behaves),
prototype it and anchor the result as a `proof:` (rerunnable script + output), not as prose.
Recommendations the run could have cheaply falsified and didn't are audit findings.

## Deliverable expectations

The report's recommendations should be **implementable as tracked issues**: concrete enough that
each becomes a bug or feature request with acceptance criteria, without requiring the
implementation itself in this run.
