# frank-exchange-of-views

> *A heated argument, diplomatically put.*

The research debate engine of [Special Circumstances](../../README.md). Given a topic, it produces a verified research report with the full record of the debate preserved.

**Carries:** an *additive-only* blue side (union, never summary), a *subtractive* red side that owns the PASS/FAIL gate with graded trust (corroboration confidence) and graded risk (likelihood × impact × complexity), an engine — a deterministic workflow script — for the mechanics, and a bench — a `lead-judge` agent invoked only for the docket (gaps at impasse) and the final compromise. Best-of-N lanes, each working one method, the Catechism (a worth-our-time decision, adapted from Heilmeier), per-run artifacts (the report, the full record of every finding, citation and ruling, the three-party `debate.md` transcript). Termination is judged (red-PASS or deadlock/spinning), never counted; the epoch limit only bounds cost.

**Run it:** `/frank-exchange-of-views:research <topic> [--lanes N] [--lens-areas a,b,c] [--k-max N] [--mint-budget N] [--max-epochs N]` — the run lands in `research/<date>_<slug>/`: `report.md` is the research, and the debate's own documents are beside it, one per audience (`docket.md`, `debate.md`, `judgments.md`, `avenues.md`, `evidence.md`, `run.md`, `CHANGELOG.md`), indexed by `README.md` and rendered together as a tabbed `report.html`.

**Citations:** every footnote and Bibliography line is built by the tool from a fetched, hash-addressed copy of the source, stamped with what the tool knows about it (retracted, abstract only) and checked by red against the same bytes. How that works, step by step, and where it is known to fail: [`docs/citations-and-bibliography.md`](docs/citations-and-bibliography.md).

**Old records:** the record has no compatibility layer — a binary refuses, loudly, any record whose event vocabulary its schema does not declare. The way back is replay: `feov-record --seat-id operator migrate --from <oldRun> --to <freshDir>` re-drives every event, in order, through the current write path under the original clock, so the output is a record the current binary could have written. The source is never modified; `inputs/migration.json` records what was translated, refused, accepted as loss, and brought across — the per-turn measurements are brought over rather than replayed, since the transcripts they were read from are not part of a run — beside the epoch the source declared and the one this binary wrote. `capture` names a migrated run as one. Design: [`plans/replay-migration.md`](../../plans/replay-migration.md).

Depends on [prosthetic-conscience](../prosthetic-conscience) (rule-skills preloaded by the debate agents). Design: [`plans/claude-port-plan.md`](../../plans/claude-port-plan.md) §3b.
