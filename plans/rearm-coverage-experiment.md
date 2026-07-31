# The re-arm coverage experiment — a runbook for a FRESH session

**Read this if you are starting a new session in this repo and the checkpoint note points you here.**
Everything below is designed to be executed by someone with no memory of the session that wrote it.

Two questions cannot be answered from inside a running session, because both are about what happens
*at* a session start. This is the runbook for both. It should take under ten minutes.

> **RUN 2026-07-31T03:43–03:50Z (session `307f5c51`, `/clear`-minted). Read this before re-running.**
>
> - **Q1: still confirmed, 19/19.** But a new defect appeared beside it: on the **first**
>   `SessionStart` of a newly minted session id the path is carried and **does not yet exist**, so the
>   row is `resolved: false` **permanently** and `gray-area checkpoint` refuses all session. That is a
>   genuine race. Evidence: gray-area §11.9.
> - **Q1b: the answer CHANGED. A seat row resolved `true`** — 16,442 bytes, real per-seat file with
>   `isSidechain: true` turns. #189's "never resolves / not written in this environment" is scoped to
>   the old session, not the environment. Report on #189, gray-area §11.9.
> - **Q2: NOT answered, and NOT answerable from a `/clear` session.** The `/clear` this runbook
>   prescribes is what disables the mechanism under test — see step 2's correction and
>   `plans/context-checkpointing.md` §20. **Next runner: do NOT `/clear` if Q2 is the goal.** The
>   two instructions are in conflict and Q2 loses.

---

## Q1 — Does `SessionStart` carry `transcript_path`? **ANSWERED: yes** (gray-area §11.7)

> Confirmed 2026-07-30T17:12Z: six firings, every row `resolved: true` with a real path. Keep the
> procedure below — it is the check to re-run if the harness changes — but it is no longer open.

### Q1b — does `agent_transcript_path` resolve? ~~**ANSWERED: no. Not a race**~~ **SOMETIMES — re-opened 2026-07-31** (#189, gray-area §11.8 → §11.9)

> **The "no" below held for 10 rows in session `937047bc` and is FALSE as a general claim.** Re-run
> on 2026-07-31 produced `resolved: true` on the first attempt in a new session: a real 16,442-byte
> per-seat transcript at exactly the recorded path. The 10 old failures are still failures — every
> path re-`stat`ed 10h41m later, still absent — so both observations stand and the mechanism that
> separates them is **unknown**. Do not write a cause here. Read resolvability off the row, per
> session. Detail: gray-area §11.9.

> **The race hypothesis below is REFUTED — read it only to see how it got here.** Eight seat rows,
> eight `resolved: false`; the named `subagents/` directory does not exist ~10 hours later; the two
> per-seat files that exist anywhere on disk both predate the wiring, sit under cwd-keyed project
> directories, and share no id with any seat row; and the parent transcript holds zero
> `isSidechain: true` entries, so the content is not there either. Full evidence on **#189**.
>
> **Phase 0's acceptance criterion cannot be met in this environment.** Do not add a fallback path
> search — a wrong file confidently attributed to a seat is the false citation the adjudicator exists
> to refuse.
>
> **Still worth re-running**, which is why the procedure stays: it is cheap, it is the check that
> fires if the harness changes back, and more evidence costs nothing. Spawn a subagent, then look for
> a `kind: "seat"` row in `.claude/gray-area/`. A `resolved: true` row would mean the behaviour
> returned — report it, don't assume it.

> ~~A `SubagentStop` at 17:05:38 produced `resolved: false` — *`stat …/subagents/agent-<id>.jsonl: no
> such file or directory`*. `hook-surface-spike.md` §3 measured that path resolving to a real file,
> and Phase 0's acceptance criterion requires it be readable. One observation, not yet a general
> claim; **it may be a timing race** with the seat's file being written after the hook fires.~~
>
> ~~**To settle it:** run a task that spawns a subagent, then check `.claude/gray-area/` for a
> `kind: "seat"` row. If `resolved: false`, re-`stat` the recorded path a minute later — if it
> exists by then, it is a race and the capture hook should retry or record later rather than
> declaring it unresolvable.~~
>
> The struck text is kept because of what it did: "it may be a timing race" was a guess offered
> beside a single observation, and it became the thing this runbook told the next session to test.
> The observation was sound; the mechanism was invented. **A guess written beside an observation is
> read as part of it** — the same failure as the 100ms probe below, one level up.

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

   > **CORRECTED 2026-07-31. This step used to hard-code `"source":"resume"` and call the result
   > "what is actually being watched this session". In a `/clear`-minted session that is FALSE: it
   > prints a populated list the session never registered.** Substitute the source this session
   > really started with.

   ```bash
   SRC=clear   # <-- the source THIS session started with: startup | resume | compact | clear
   CLAUDE_PROJECT_DIR=$PWD plugins/prosthetic-conscience/bin/sc-checkpoint-restore \
     <<< '{"session_id":"probe","cwd":"'"$PWD"'","hook_event_name":"SessionStart","source":"'"$SRC"'"}' \
     | python3 -c 'import json,sys; print(json.load(sys.stdin)["hookSpecificOutput"].get("watchPaths"))'
   ```

   **If this prints `None`, STOP — steps 3–6 cannot answer anything.** A `clear`-sourced session
   registers zero watchPaths deliberately (`TestPointerSessionsRegisterNoWatch`: "a pointer session
   is not resuming this work"). Any null you measure afterwards is caused by that, not by #165, and
   the two are indistinguishable from the outside. Q2 needs a session started with `startup`,
   `resume` or `compact`. Otherwise pick a surface from the list and continue.

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
- Q1b → a comment on **#189**, then `plans/gray-area.md` §11.8. Only a `resolved: true` row is news;
  another `resolved: false` confirms what is already recorded and needs no write-up beyond the count
- Q2 → a comment on #165, then `plans/context-checkpointing.md`
- Both → the checkpoint note's validation loop, so the next session inherits the answer rather than
  the question
