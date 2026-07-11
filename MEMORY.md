# Special Circumstances — design-rationale log

A running log of *why* the suite is shaped the way it is. PII-scrubbed. One entry per decision.

## 2026-07-11 — Founding

- **Three plugins, one marketplace**, named for the Culture: prosthetic-conscience (core), frank-exchange-of-views (research debate engine), sleeper-service (autonomous learning). prosthetic-conscience is the base; sleeper-service requires frank-exchange-of-views.
- **Ported methodology, deleted machinery.** From the origin `AgentOrange`/Antigravity experiment, only the distinctive methodology came across (rules, debate protocol, gates, templates). The local-LLM stack and hand-rolled harness are out of scope — not ported, not archived.
- **Clean start.** `ideas/`, `research/`, `projects/` ship empty; the corpus is grown by the suite itself.
- Full teardown, first-principles harness comparison, and phased build plan: `plans/claude-port-plan.md`.
