---
name: context-checkpointing
description: Use for long sessions at risk of compaction, before a risky step, and when resuming after compaction or a restart — the single overwritten CHECKPOINT.md carrying the validation loop, ordered next actions, and in-flight handles.
---

# context-checkpointing

Leave the note while you still remember. Compaction replaces the transcript with a summary; what the summary drops is gone. The summary is good at what happened and worst at what you were about to do.

## The note

One file, **overwritten in place** — `CHECKPOINT.md` in the active `projects/<name>/` or `research/<slug>/`, else `.claude/checkpoints/CHECKPOINT.md`.

```markdown
---
schema: 2
updated: <UTC ISO>
session_id: <id>          # NOT unique — every subagent shares the parent's
agent_id: <id|null>       # the seat's own id, when running as a subagent
objective: "<what I am trying to achieve right now>"
plan: <path §section|null>
beyond_plan: <true|false> # true once work has crossed the plan's scope
status: <in-progress|blocked|validating|done>
---
## Validation loop        ← load-bearing
1. <exact command>  → <expected>  · re-armed by: <file/condition that makes it fire>
   last run: <result>
## Next intended steps    ← each real work item carries its canonical-queue pointer (issue / plan task)
## In-flight handles      ← background task ids, PRs, long-running processes
## Invariants / foot-guns ← verbatim
## Decisions made / rejected
## Files touched          ← mid-edit state a summary flattens
## Open threads
```

## The contracts

- BEFORE a risky step, and on crossing beyond the plan's scope (`beyond_plan: true`), YOU MUST write the checkpoint. Auto-compaction gives no warning, so YOU MUST NOT defer this to the moment you need it.
- During the session, YOU MUST update the note when the **validation loop** is established or changes, when a decision is reached or rejected, and when background work starts — recording each check's **trigger surface**, because compaction drops that first.
- YOU MUST keep **one block, overwritten** — never accumulate. An accumulating note outgrows the harness's auto-recall budget and silently degrades from an anchor to a pointer you may not follow.
- BEFORE recording a forward actionable, YOU MUST file it where the resumed workflow will actually look (issue, or a task in the plan) and let the note carry the pointer. A note-only actionable dies when the worklist is rebuilt from another index (see [[project-memory]]).
- AFTER writing the note, YOU MUST register its path in the project's durable memory, so continuity survives a cold start where no hook fires.
- AFTER a compaction or restart, the restore path is **read-only until the ordered next-actions list** — YOU MUST NOT re-execute anything replayed from before the seam, and YOU MUST verify each checkpoint claim against reality before acting on it. The note is a claim, not a fact.
- AFTER the work completes, YOU MUST fold durable decisions into the plan or project memory and discard the note — it is scaffolding, not a record.

## Boundary with [[project-memory]]

Project memory is the long-horizon record — what this workstream *is*, and the decisions that stuck. The checkpoint is the **volatile cursor over it**: which validation step last ran and what it returned, what you were about to do next, what is still running. Memory holds the canonical verification suite; the checkpoint holds its live state. Where there is no project directory, the checkpoint carries continuity alone.
