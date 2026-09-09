# frank-exchange-of-views

> *A heated argument, diplomatically put.*

The research debate engine of [Special Circumstances](../../README.md). Given a topic, it produces a verified research deliverable with the full adversarial record preserved.

**Carries:** an *additive-only* blue team (union, never summary), a *subtractive* red team that owns the PASS/FAIL gate with graded trust (corroboration confidence) and graded risk (likelihood × impact × complexity), and a lead — a deterministic workflow script for mechanics plus a `lead-judge` agent invoked only for the docket — gaps at impasse — and the final compromise. Best-of-N lanes/lenses, the Catechism (a worth-our-time decision, adapted from Heilmeier), per-run artifacts (living blue/red reports, the full event record of every finding, citation and ruling, the three-party `debate.md` transcript). Termination is judged (red-PASS or deadlock/spinning), never counted; the safety ceiling only bounds cost.

**Run it:** `/frank-exchange-of-views:research <topic> [--lanes N] [--lens-areas a,b,c] [--k-max N] [--mint-budget N]` — the deliverable lands in `research/<date>_<slug>/`: `report.md` is the research, and the entire adversarial record is beside it, one document per audience (`docket.md`, `debate.md`, `judgments.md`, `evidence.md`, `run.md`, `CHANGELOG.md`), indexed by `README.md` and rendered together as a tabbed `report.html`.

**Old records:** the record has no compatibility layer — a binary refuses, loudly, any record whose event vocabulary its schema does not declare. The way back is replay: `feov-record --seat-id operator migrate --from <oldRun> --to <freshDir>` re-drives every event, in order, through the current write path under the original clock, so the output is a record the current binary could have written. The source is never modified; `inputs/migration.json` records what was translated, refused, and accepted as loss, and `capture` names a migrated run as one. Design: [`plans/replay-migration.md`](../../plans/replay-migration.md).

Depends on [prosthetic-conscience](../prosthetic-conscience) (rule-skills preloaded by the debate agents). Design: [`plans/claude-port-plan.md`](../../plans/claude-port-plan.md) §3b.
