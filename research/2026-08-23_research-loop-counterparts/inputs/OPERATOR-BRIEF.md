# Operator brief — research-loop-counterparts

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

THIS run produces **the positioning**: sleeper-service understood as a (mostly) automated research
loop built on FEOV, compared and contrasted with:

1. **Known external systems** — automated-research and self-improving-agent systems in the
   literature and in products (e.g. AI-Scientist-style closed-loop paper generators, evolutionary
   program-search systems, agent-architecture-search, deep-research products, self-refinement
   loops, Gödel-machine-lineage proposals — whatever the evidence actually supports; verify at the
   leaf, do not take this list as the answer). What does each automate, where does each put its
   verification, where its human gate, and what does the FEOV loop do differently (adversarial
   gate owned by red, judged termination, tool-mediated record, human-gated promotion)?
2. **Its in-repo parallel counterparts** — the **dream loop** (memory consolidation,
   `plans/memory-architecture.md` §7.4–7.6) and the **memory architecture** (the OKF store,
   promotion/decay lifecycle). All three are scheduled, human-gated, git-native background loops:
   self-improve evolves the *rules*, dream consolidates the *knowledge*, memory is the *substrate*.
   Where do they genuinely share machinery (scheduling, promotion ladders, human gates, git as
   substrate), where must they stay separate, and what does each borrow from the others?

OUT of scope here (the sibling run `2026-08-23_sleeper-service-plan` owns it): the concrete build
plan for the plugin — its commands, phases, and issue-ready increments.

## Standing process requirement (from the operator, binding on every recommendation)

Blue's process MUST verify its work and recommendations by **prototyping and testing** where a
computation can settle a claim — anchored as `proof:` (rerunnable script + output), not prose.
For external-system claims, verification is at the leaf: the paper, the repo, the changelog —
never a secondary summary.

## Deliverable expectations

A compare-and-contrast that yields actionable convergence/divergence findings — each borrowable
mechanism or required separation concrete enough to file as a tracked issue.
