# Special Circumstances

> *An adversarial human/AI methodology suite. It argues with you on purpose — and it researches its own rules while you sleep.*

**Special Circumstances** is a Claude Code marketplace of three cooperating plugins, named for the [Culture](https://en.wikipedia.org/wiki/The_Culture) — Iain M. Banks's civilisation of humans and Minds who treat a good argument as an act of respect.

| Plugin | Named for | What it does |
|---|---|---|
| [**prosthetic-conscience**](plugins/prosthetic-conscience) | the drone that keeps you honest | Core adversarial + cowork behaviour: Design-by-Contract rules (shipped as skills), pair programming, spec-driven development + a plan-audit gate, project memory, proficiency skills, quality hooks, and an environment preflight (`/prosthetic-conscience:doctor`). |
| [**frank-exchange-of-views**](plugins/frank-exchange-of-views) | a heated argument, diplomatically put | The research debate engine: an *additive* blue team builds, a *subtractive* red team audits and owns the pass/fail gate, a lead runs the mechanics and the final compromise. Best-of-N, a Heilmeier Catechism, and the full adversarial record preserved — union, never summary. |
| [**sleeper-service**](plugins/sleeper-service) | the GSV quietly running vast hidden projects | Autonomous learning: a daily self-improvement loop, a graduation pipeline, and a continuous-learning promotion ladder. Always human-gated at promotion. |

## Install

```text
/plugin marketplace add ctoforaday/special-circumstances
/plugin install prosthetic-conscience@special-circumstances
/plugin install frank-exchange-of-views@special-circumstances
/plugin install sleeper-service@special-circumstances
```

`prosthetic-conscience` is the base; the other two preload its rule-skills. `sleeper-service` invokes `frank-exchange-of-views`. One marketplace install gets all three; each remains individually useful.

## Dogfood tour

- **`/prosthetic-conscience:doctor`** — check your environment is set up (qlty, git, gh) before anything assumes a toolchain.
- **`/research <topic>`** — run the debate engine; watch `research/<date>_<slug>/debate.md` grow as blue and red argue and the lead adjudicates.
- **`/plan-audit <file>`** — put an implementation plan through the spec-driven-development gate.
- **`/self-improve`** — let the suite research how one of its own rules should evolve (writes only to `ideas/` and `research/`).

## Status

Early bootstrap (Phase 0): this repo currently holds the marketplace skeleton and the design under review in [`plans/`](plans) — the master port plan plus proposals for the memory architecture and context-compression checkpointing. The plugins are built out phase by phase from that plan.

## Origins

Ported from an earlier "Antigravity Meta Brain" experiment (the private `AgentOrange` repo). The *methodology* came across — the rules, the debate protocol, the gates, the templates. The hand-rolled *harness* did not: a 787-line orchestrator, prompt/skill compilers, watchdogs, and a local-LLM proxy stack, all made redundant by Claude Code's native model. See [`plans/claude-port-plan.md`](plans/claude-port-plan.md) for the full teardown and rationale.

## License

MIT © 2026 Gregory Block
