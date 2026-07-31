# The re-arm coverage experiment — a runbook for a FRESH session

**Read this if you are starting a new session in this repo and the checkpoint note points you here.**
Everything below is designed to be executed by someone with no memory of the session that wrote it.

Two questions cannot be answered from inside a running session, because both are about what happens
*at* a session start. This is the runbook for both. It should take under ten minutes.

---

## Q1 — Does `SessionStart` carry `transcript_path`? **ANSWERED: yes** (gray-area §11.7)

> Confirmed 2026-07-30T17:12Z: six firings, every row `resolved: true` with a real path. Keep the
> procedure below — it is the check to re-run if the harness changes — but it is no longer open.

### Q1b — NEW AND OPEN: does `agent_transcript_path` resolve?
>
> A `SubagentStop` at 17:05:38 produced `resolved: false` — *`stat …/subagents/agent-<id>.jsonl: no
> such file or directory`*. `hook-surface-spike.md` §3 measured that path resolving to a real file,
> and Phase 0's acceptance criterion requires it be readable. One observation, not yet a general
> claim; it may be a timing race with the seat's file being written after the hook fires.
>
> **To settle it:** run a task that spawns a subagent, then check `.claude/gray-area/` for a
> `kind: "seat"` row. If `resolved: false`, re-`stat` the recorded path a minute later — if it
> exists by then, it is a race and the capture hook should retry or record later rather than
> declaring it unresolvable.

`plans/gray-area.md` §11.3 records this as an **explicitly unverified assumption**.
`hook-surface-spike.md` §3 says every hook event carries `session_id`, `transcript_path` and `cwd`,
but that was never re-measured for `SessionStart`, and `gray-area checkpoint` depends on it to
resolve a session's own trajectory without globbing `~/.claude/projects/`.

The hook is built to make the answer visible either way: a payload with no `transcript_path` writes
**no row** and prints one line to stderr.

**Check, in order:**

```bash
ls -la .claude/gray-area/                     # a trajectories-<session>.jsonl should exist
cat .claude/gray-area/trajectories-*.jsonl | python3 -m json.tool | head -20
```

| What you see | What it means | What to do |
|---|---|---|
| a row with `"kind": "session"` and a non-empty `transcript_path` | **CONFIRMED.** The assumption holds. | Mark §11.5 item 5 done in `plans/gray-area.md`; drop the §11.3 caveat |
| no manifest at all | the hook did not run — check the wiring below before concluding anything | see *Wiring* |
| a manifest with only `"kind": "seat"` rows | `SubagentStop` fired, `SessionStart` did not | wiring problem, not an assumption problem |
| stderr said `SessionStart carried no transcript_path` | **REFUTED.** The spike's claim is wrong for this event. | `gray-area checkpoint` needs another way to resolve; record it in §11.3 and reopen the design question |

Then confirm the whole path works end to end:

```bash
plugins/gray-area/bin/gray-area checkpoint .claude/checkpoints/CHECKPOINT.md
```

It should print `resolved this session's trajectory from …` and then adjudicate. Exit 1 on
`STALE`/`NO-EVIDENCE` is normal and expected, not a failure of this experiment.

---

## Q2 — Why does `FileChanged` coverage stop mid-session? (#165)

**Read #165 first.** Its original claim — that `FileChanged` never fires — is **withdrawn**. Events
demonstrably do arrive, for both `git checkout` churn and the session's own Write-tool edits. What
is unexplained is that re-arm records *stop advancing* partway through a session even though editing
continues under a registered surface.

Background you need: `sc-filechanged-rearm` keys `rearmed.json` by check index and **overwrites**, so
each check holds only its most recent re-arm. Few records is expected. A record whose timestamp is
old while its surface is being actively edited is **not**.

**The experiment:**

1. Note the baseline.
   ```bash
   cat .claude/checkpoints/rearmed.json 2>/dev/null | python3 -m json.tool
   ```
   Record which check indices are present and their `at` timestamps.

2. Confirm what is actually being watched this session — do not assume.
   ```bash
   CLAUDE_PROJECT_DIR=$PWD plugins/prosthetic-conscience/bin/sc-checkpoint-restore \
     <<< '{"session_id":"probe","cwd":"'"$PWD"'","hook_event_name":"SessionStart","source":"resume"}' \
     | python3 -c 'import json,sys; print(json.load(sys.stdin)["hookSpecificOutput"].get("watchPaths"))'
   ```
   This shows what the restore hook *would* register from the current note. Pick a surface from it.

3. Edit a real file under that surface with the **Edit or Write tool** (not Bash — the point is to
   test the session's own edits).

4. **WAIT AT LEAST 60 SECONDS.** This is the whole methodological point: the original wrong
   conclusion came from checking ~100ms after the write. The known-good records at `04:38:10` all
   share one timestamp, which looks like batching, so latency is plausible and untested.

5. Re-read `rearmed.json`. Did the relevant check's `at` advance?

| Result | Reading |
|---|---|
| timestamp advanced | coverage is fine and the original stop was something transient — say so and close the question |
| unchanged after 60s, then advances later | there is a debounce; measure it and record the figure |
| still unchanged after several minutes | the failure is real. Go to step 6 |

6. If it did not advance: rewrite the checkpoint note (add or remove a validation-loop entry so the
   indices shift), edit under the surface again, wait, re-check. If it works *before* a note rewrite
   and not after, hypothesis (1) in #165 is confirmed — registration goes stale relative to a
   rewritten note, and the fix is to re-register rather than to change the watcher.

---

## Wiring — check this before concluding anything from an absence

**The plugin cache is empty in this environment, so `plugins/*/hooks/hooks.json` never loads.** Hooks
only fire because `.claude/settings.local.json` points at locally built binaries. That file is
gitignored, so a fresh container will not have it, and a missing row proves nothing about the
harness.

```bash
python3 -c 'import json;d=json.load(open(".claude/settings.local.json"));print(sorted(d["hooks"]))'
grep -c gray-area-capture .claude/settings.local.json     # expect 2 (SessionStart + SubagentStop)
ls -la plugins/gray-area/bin/ plugins/prosthetic-conscience/bin/
```

Rebuild the binaries if they are missing:

```bash
for d in plugins/gray-area/tools/cmd/*/;              do go build -C plugins/gray-area/tools -o "../bin/$(basename "$d")" "./cmd/$(basename "$d")"; done
for d in plugins/prosthetic-conscience/tools/cmd/*/;  do go build -C plugins/prosthetic-conscience/tools -o "../bin/$(basename "$d")" "./cmd/$(basename "$d")"; done
```

---

## The rule this runbook exists to enforce

Both questions were previously answered wrongly, in the same way: **one observation, taken as
settled.** The re-arm mechanism was declared dead on a single 100ms probe; that judgement then
propagated into an issue, a plan section, a code comment and a checkpoint note before anyone looked
again.

So when you run this: if you see nothing, **wait and look again** before writing it down. Record the
negative *with how long you waited and what you searched* — the same standard `gray-area checkpoint`
holds itself to when it reports `NO-EVIDENCE` rather than "did not run".

## Where to write the answer

- Q1 → `plans/gray-area.md` §11.5 item 5 and §11.3, then the plugin README if the answer changes it
- Q2 → a comment on #165, then `plans/context-checkpointing.md`
- Both → the checkpoint note's validation loop, so the next session inherits the answer rather than
  the question
