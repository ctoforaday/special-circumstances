---
description: Print the full checkpoint and re-ground on it — the whole note, not the terse digest the SessionStart hook injects. `--seals` lists the sealed snapshots instead.
---

Apply [[context-checkpointing]]. The restore hook injects a **digest**; this prints the **note**.

**`--seals`** — list `.claude/checkpoints/*.md` snapshots newest first with their sealed-at stamp, trigger and agent, and stop. Print the newest snapshot's path so it can be read directly. Do not restore from a seal: a discarded live note means the work completed, and resurrecting it re-opens finished work.

**No argument:**

1. Locate the note — active `projects/<name>/` or `research/<slug>/`, else `.claude/checkpoints/CHECKPOINT.md`. If none exists, say so in one line and stop. Do not reconstruct one from the transcript; an invented checkpoint is trusted exactly as much as a real one.
2. Print it verbatim.
3. Report the seam: the note's `updated` timestamp, its `session_id`/`agent_id`, and whether they match this session. A note written by a *different* session is still useful and MUST be labelled as such — it is someone else's claim about a shared tree.
4. **Verify before acting, in this order.** The note is a claim written before the seam, not an observation of now:
   - Re-run the validation loop rather than trusting its recorded `last run`. Reproduce each check; YOU MUST NOT act on a paraphrase of what a gate wanted (see [[validation-loop]]).
   - Check each in-flight handle still exists — background ids, open pull requests, running processes. A handle that died during the seam is the failure this section exists to catch.
   - Check each next-action's queue pointer still says what the note says it says. An item marked `NOTE-ONLY` has no queue behind it and will be lost again.
5. Report in three lines: which claims verified, which did not, and the first action you intend to take.

YOU MUST treat everything before the ordered next-actions list as **read-only**. YOU MUST NOT re-execute work replayed from before the seam — the transcript you are reading is a record of what already happened, and re-running it is how a resumed session duplicates side effects.

If the note contradicts what you can observe, the observation wins and the note is stale. Say so explicitly rather than reconciling silently.
