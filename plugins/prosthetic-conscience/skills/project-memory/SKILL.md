---
name: project-memory
description: Use for any long-running or multi-session workstream — the four-artifact projects/<name>/ discipline that lets work survive compaction and restarts.
---

# project-memory

Long work lives in artifacts, not in context. One directory per workstream: `projects/<name>/`.

## The four living artifacts

1. **`AGENTS.md`** — the context bootstrap: mission + success criteria, "how we got here" (key discoveries and decisions, chronological), and the **operational verification suite** (exact commands to test, lint, and run end-to-end).
2. **`implementation_plan.md`** — the approved implementation plan (see [[spec-driven-development]]).
3. **`task.md`** — the living TODO list (`[ ]` open, `[/]` in progress, `[x]` done).
4. **`walkthrough.md`** — verification outcomes: commands run, results observed.

## The contracts

- BEFORE starting a complex or multi-session workstream, YOU MUST create the project directory and seed the four artifacts.
- AFTER completing a phase or landing a change, YOU MUST update `task.md` and `walkthrough.md` — the files, not just the in-session todo list.
- AFTER a context compaction or session restart, YOU MUST read `AGENTS.md` and `task.md` FIRST — and the workstream's `CHECKPOINT.md` if one exists (see [[context-checkpointing]]) — and reconcile them against whatever you believe. The artifacts are the truth, your memory is the cache. Memory holds what this workstream *is*; the checkpoint holds where the cursor was, and it is the only one carrying in-flight handles.
- During the workstream, YOU MUST keep `AGENTS.md` current when direction changes — a stale bootstrap is worse than none.
