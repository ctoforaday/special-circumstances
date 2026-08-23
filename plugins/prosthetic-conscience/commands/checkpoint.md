---
description: Write the session checkpoint now — the validation loop, ordered next actions, and in-flight handles, before compaction takes them. `--show` prints the current note instead of writing.
---

Apply [[context-checkpointing]]. Model [[terse-communication]]: the note is the tattoo, not the autobiography.

**`--show`** — locate the current `CHECKPOINT.md` (active `projects/<name>/` or `research/<slug>/`, else `.claude/checkpoints/CHECKPOINT.md`), print it verbatim, and stop. If none exists, say so in one line. Do not write.

**No argument** — write the note:

1. Locate the note as above. If none exists, create it; if one does, **overwrite it in place** — one block, never appended.
2. **If the note is still accurate, do not rewrite it — set `reaffirmed_at` and stop.** That is a
   complete, valid answer, and it is a different fact from a rewrite: `written_at` and `head` describe
   the CONTENT and must not move when the content did not. A re-affirmation that also touches them
   reports the note as fresh and the branch as re-established when neither happened, and no reader can
   tell afterwards which of the two occurred.
3. Fill every section of the schema in [[context-checkpointing]]. The **validation loop** is load-bearing: exact commands, in order, each with its expected result, what re-arms it, and what it returned when last run. Carry commands verbatim — a paraphrase of what a check wants is what compaction produces, and reproducing the gate is the only way back from that.
4. Set `beyond_plan: true` if the work has crossed the scope of its plan. That flag is the signal the note is now load-bearing rather than convenient.
5. For each forward actionable that is real work rather than a musing: it MUST already exist in the durable queue the resumed workflow reads — a GitHub issue, or a numbered task in the plan. If it does not, file it first (see [[semantic-consent]] before creating anything outward-facing), then record the pointer. If it stays note-only, mark it `NOTE-ONLY` so the omission is visible rather than silent.
6. Record in-flight handles — background task ids, open pull requests, long-running processes. After compaction these are invisible to the conversation and the note is the only thread back to them.
7. Register the note's path in the project's durable memory if it is not already there.
8. Report in two lines: where the note was written, and the count of validation steps, next actions, and in-flight handles it carries.

YOU MUST NOT infer the validation loop from what a check is *probably* called. If you do not have the exact commands, say so and ask — an invented loop is worse than an absent one, because the resumed session will trust it.
