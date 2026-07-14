# frank-exchange-of-views

> *A heated argument, diplomatically put.*

The research debate engine of [Special Circumstances](../../README.md). Given a topic, it produces a verified research deliverable with the full adversarial record preserved.

**Carries:** an *additive-only* blue team (union, never summary), a *subtractive* red team that owns the PASS/FAIL gate with graded trust (corroboration confidence) and graded risk (likelihood × impact × complexity), and a lead — a deterministic workflow script for mechanics plus a `lead-judge` agent invoked only for deadlock checks and the final compromise. Best-of-N lanes/lenses, the Catechism (a worth-our-time decision, adapted from Heilmeier), per-run artifacts (living blue/red reports, preserved candidates, the full three-party `debate.md` transcript). Termination is judged (red-PASS or deadlock/spinning), never counted; the safety ceiling only bounds cost.

**Run it:** `/frank-exchange-of-views:research <topic> [--lanes N] [--lenses N] [--max-rounds N]` — the deliverable lands in `research/<date>_<slug>/report.md` with the entire adversarial record beside it.

Depends on [prosthetic-conscience](../prosthetic-conscience) (rule-skills preloaded by the debate agents). Design: [`plans/claude-port-plan.md`](../../plans/claude-port-plan.md) §3b.
