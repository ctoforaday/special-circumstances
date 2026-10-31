# Local development — running the engine against a checkout

The gates tell you a change compiles and that its goldens moved. They do not tell you the research
engine still runs. For that the plugins have to be **installed** — content copied into a plugin
root, hook binaries built there, seats dispatched as real subagents — and then driven the way a
user drives them.

This is how to do that without inventing a harness.

## The rule that this page exists to enforce

**Exercise the engine through `claude`, not through a harness that stands in for it.**

A scratch harness that reimplements the Workflow tool over a pool of `claude -p` processes looks
equivalent and is not. Measured, each of these cost a day:

- **Seats spawned as top-level processes never fire `SubagentStart`/`SubagentStop`.** No sitting
  span is written, and `gray-area` capture has no transcript to bind. (The per-sitting tool-call cap
  survives only because `sittingcap` keys on the register event — its package comment says why.)
- **A harness that pastes each seat's JSON schema into the prompt and then accepts any parseable
  object is not enforcing it.** The real harness enforces it. An arm died twelve sittings in on a
  relayed plan missing a field the tool always writes; the defect was in the instrument.
- **`parallel()` swallowing a rejected seat into `null`** hands the engine the same value a lens
  that audited and found nothing returns. A wave where seats died reads as a wave that passed.

A rig that is easier than the thing it stands for does not give a conservative reading. It runs a
different experiment.

## `scripts/universe.sh`

```sh
scripts/universe.sh build            # build an ISOLATED environment from this checkout
scripts/universe.sh build --live     # install into ~/.claude instead (shared with every session)
scripts/universe.sh run "<topic>"    # one DETACHED `claude -p`, invoking /research as a user would, and follow its log
scripts/universe.sh run --no-follow "<topic>"   # the same, and return at once
scripts/universe.sh resume           # restart a run whose process died with no outcome
scripts/universe.sh watch [secs]     # the BOARD, read-only from the record, until the process ends
scripts/universe.sh shell            # interactive claude inside the universe
scripts/universe.sh doctor           # what is installed, binaries per plugin, schema epoch
```

`build` is isolated by default: `CLAUDE_CONFIG_DIR` and `CLAUDE_CODE_PLUGIN_CACHE_DIR` point inside
`WORKDIR` (default `~/.claude/scratch/universe`), so nothing touches `~/.claude` and other sessions
on the box are unaffected. Use `--live` only when you want the change in *your* next session, and
remember it is shared state.

Tearing one down is **two** paths, not one: `$WORKDIR` and `$WORKDIR.marketplace`. The staged
marketplace is a copy of the checkout's `plugins/` beside the universe, and it is the plugin root a
universe session loads (below). `build` prints both paths when it finishes.

### The manifest has to be restaged, and the reason is not obvious

`.claude-plugin/marketplace.json` pins three of the four plugins to a **release tag**, via a
`git-subdir` source against the GitHub URL. That is right for consumers and wrong here: pointing
`claude plugin marketplace add` at the checkout still resolves content by the pinned ref, so
`install` fetches the **published** plugin and the working tree is never consulted.

Measured the first time this was run: the universe came up reporting four plugins installed and
twenty binaries built, carrying released prompts, with none of the checkout's changes in it — and
every sign of success.

`universe.sh` therefore stages its own manifest whose every source is `./plugins/<name>`, over a
symlink into the checkout (`scripts/universe-manifest.py`). Same four plugins, same names; only the
content resolves differently.

The path that manifest is staged at is recorded in the universe's `settings.json` and re-read at
**every session start**, not just at install. Staged in a temp directory and cleaned up after the
build, it left a config naming a directory that no longer existed: cache fully populated, every
binary present, `doctor` clean — and the plugins silently not loading. What that looks like from
outside is a lead that receives `/frank-exchange-of-views:research <topic>` as ordinary prose,
because the command does not exist in that session, and **answers the topic**: one turn, five
cents, `DONE success`. `build` now asserts that the recorded path exists and is the one it staged.

**After any `build`, prove the universe carries your change** rather than trusting the summary —
and prove it in the staged marketplace, because that is the tree a universe session runs:

```sh
M=~/.claude/scratch/universe.marketplace/plugins
grep -c '<a phrase from your change>' $M/frank-exchange-of-views/skills/research-protocol/scripts/debate.js
```

### A directory marketplace runs its plugins in place

The universe's marketplace is a `directory` source, and Claude Code loads a directory source's
plugins from that directory: `${CLAUDE_PLUGIN_ROOT}`, each skill's base directory, the agent
definitions, `hooks.json` and `bin/` all resolve under `$WORKDIR.marketplace/plugins/<name>`. The
stream's `init` event names it — every entry of `plugins[].path` is a staged-marketplace path.
`claude plugin install` still copies each plugin into the versioned cache
(`$WORKDIR/plugins/cache/…`) and names that copy in `installed_plugins.json`, but a session does not
load from it: a run's stream and transcripts reference the cache nowhere.

So a check that reads the cache proves what was installed, not what runs. Grep the staged
marketplace for a prompt, skill or constitution change; for a binary, `cmp` the staged copy against
the cache copy, or ask the staged one directly (`$M/frank-exchange-of-views/bin/feov-record --schema`).

### Binaries are built in the cache and mirrored

`build` compiles each plugin's binaries in the cache, where the install put its `tools/` — the
ordering `scripts/bootstrap-plugins.sh` uses for cloud environments, which install from a git
source and do run from the cache — and copies them into the staged marketplace's `bin/`. A build in
the checkout leaves both roots with no binaries and hooks degrading to a single stderr line, which
from inside a run looks exactly like hooks that were never configured.

A cache build has no git context, so `feov-record --version` reports `unknown` where a release
reports its sha. **Version-skew detection is inert in a universe run** — `tool_version` and
`hook_version` on the record are not meaningful there.

## Seeing what is happening

`--output-format json` returns one object when the process exits. For a run of tens of minutes that
means no output at all until it ends, so a working run and a wedged run look identical. Two views,
answering different questions:

**The harness view** — `universe.sh run` has claude write `--output-format stream-json --verbose` to
`stream-<stamp>.jsonl`, and tails that file through `scripts/universe-stream.py` into
`run-<stamp>.log`, which prints dispatches, record verbs and tool errors as they happen. The view
cannot harm the capture: the renderer reads a file claude writes, so nothing about it can stop claude
or the raw stream.

```
[  312s] DISPATCH  red-lens-voice #1
[  318s]   verb    mint
[  341s]   ERROR   refused: --quote is required
[  902s] RESULT  success  turns=42  $1.23
```

A `RESULT` line ends a turn, not the run, and there can be several. A resumed session's first one
comes from a queued notification with `turns=0`, before the lead has read its prompt. Only `EXIT` says
the run ended.

**The debate does not outlive the `claude -p` process.** Print mode waits a bounded time for
background tasks and then terminates them — `Background tasks still running after 600s;
terminating` — and a research run is tens of minutes, so the workflow is killed mid-debate. Measured
on `is 91 prime`: last event a lens's `mint` with no `sitting_close` after it, nine sitting spans
opened and eight closed, no chair verdict, no outcome, three gaps left open. An earlier measurement
found the workflow alive nine minutes after exit and concluded it survives; nine minutes is 540s,
*inside* the ceiling, so it never reached the boundary it was taken to have tested. `universe.sh`
sets `CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS=0` to wait indefinitely.

**A run outlasts the tool that launched it.** A debate can run for hours, longer than a launcher's
own ceiling: m37 was started as an agent's Bash background task, which was killed at its 7200000 ms
limit in epoch 11 of 12 and took the debate with it. So `run` detaches the debate into its own session
(`setsid`, sharing no stdin or stdout with the caller) and what the caller holds is only a `tail` of
the log. Killing it, or its whole process group, leaves the run running. `run` and `resume` print the
paths; they are `$WORKDIR/run.pid`, `$WORKDIR/run-<stamp>.log` (with `run-current.log` pointing at the
latest) and `$WORKDIR/stream-<stamp>.jsonl`.

**To wait for a run, wait for its `EXIT` line.** The log's last line is `EXIT <rc>`, written when
claude has ended, whichever way it ended. It is the only thing in the log that means the run is over.
No line at all means the supervisor itself was killed, so a wait has to stop on that too, which a
`tail` bound to the pid does:

```sh
W=~/.claude/scratch/universe
tail -n +1 -f --pid="$(cat $W/run.pid)" $W/run-current.log | grep -m1 '^EXIT '
```

**`watch` reads the pid file, not the silence.** A truncated run and a running one look the same on
the board, because it just stops. `watch` therefore checks whether the session named by `run.pid` has
a live process and prints that beside the board. Once the process is gone, `watch` prints a last board
and stops: `process: ended` with an outcome on the record, or `!! TRUNCATED` with none, which is exit
status 3. The board's sitting arithmetic tells you what the kill interrupted: every sitting the
harness opens it also closes, so `opens > closes` means a seat was cut off.

**`resume` restarts a truncated run.** It resumes the lead's session with `claude -p --resume`, in the
universe's environment, and tells it to invoke the Workflow tool with `resumeFromRunId` and the SAME
args. The session, script, args and run id are read from the first Workflow call in the run's first
stream (`run-origin.jsonl`) and that call's result, so nothing is retyped. It gets its own log and
stream, and is waited on and watched the same way.

`resume` refuses, with exit status 3, while the process is alive, when no pid file says it is dead,
when the record already has an outcome or cannot be read, and when the stream holds no Workflow call.
A process alive in the universe also makes `build` and `run` refuse, and `--force` does not override
that. A session that holds only a stuck leftover is ended with the `pkill -s` the refusal prints.

**A run that dispatched no `Workflow` fails, and says so.** The engine cannot research anything
without one, so its absence is decidable from the stream — and `claude -p` exits 0 for it, because
the assistant did produce an answer. `EXIT` carries the renderer's status when claude's is clean, so
"the plugins did not load" cannot arrive as a success with a quiet `0 dispatch(es)` line above it. A
resume is held to the same bar: a resumed lead that dispatches no `Workflow` did not restart the
debate, and says that instead of blaming the plugins.

**The debate view** — `universe.sh watch` reads the record (SQLite `mode=ro`, so it cannot take a
write lock on a live run) and prints the board: seats and their sitting counts, gaps minted and
closed, the chair's verdict. It calls out **hook-written sitting spans** specifically, because that
count is zero by construction under a top-level-process harness and non-zero here — it is the
one-glance evidence that subagent dispatch is real.

The engine also ships its own instruments: `feov-record dashboard <runDir> <transcriptDir> --watch`
renders board-mass trend, live seats with a completion ETA, cost-so-far and the run's scorecards to
`dashboard.html` every 15s, and `--serve <port>` hosts it over HTTPS behind a per-run secret URL.

## The universe is a snapshot, so you can keep working

`build` COPIES the checkout's plugins into the staged marketplace (`cp -r`), and the install copies
them again into the cache. Neither is a link back to the working tree.

So once `build` has run you can keep editing, committing, rebasing and pushing the checkout while a
run is in flight, and the running engine is unaffected. It is holding a frozen copy. Re-run `build`
when you want the new code, and the next `run` picks it up.

It isolates the other way too: the run's source tree is `$WORKDIR/src`, so its `.claude/run-live.json`,
its pins and its push freeze sit there, away from the tree you develop in.

Hooks are universe-local: `${CLAUDE_PLUGIN_ROOT}` resolves to *this* universe's staged marketplace,
and the `hooks.json` there points at that root's own `bin/`. Nothing reaches outside `$WORKDIR` and
`$WORKDIR.marketplace`.

**The exception is narrower than it first looks: `build` mutates the universe IN PLACE.** It
re-stages the marketplace and reinstalls over the same version directory, and hooks spawn fresh
from the plugin root on every tool call — so rebuilding under a live run swaps its binaries
mid-sitting. Editing the checkout is
always safe; reinstalling over a running engine is not. `universe.sh build` now REFUSES when a run
is in flight in that `WORKDIR` rather than leaving it to this paragraph; use a second `WORKDIR` for
a second universe.

A rebuild propagates: it re-copies the staged marketplace from the checkout and reinstalls the same
edits into the cache. There is no stale-universe trap.

### Run the plugin commands from outside the checkout

`universe.sh` runs every `claude plugin …` call with cwd set to `$WORKDIR`, and that is load-bearing.
Run from inside the repo, `claude plugin marketplace remove` resolved the PROJECT's
`.claude/settings.json` — a **tracked** file — and stripped its `extraKnownMarketplaces` and
`enabledPlugins`. A tool built to leave the tree alone silently edited committed state, and it
showed up only as an unexplained entry in `git status`.

## Smoke runs

A smoke exercises the pipeline. It is not a small keeper, and the difference is an order of
magnitude:

| | smoke | development |
|---|---|---|
| lanes | 1 (+ `laneFloorOverride`) | 3 |
| bulk tier | `haiku` | `sonnet` |
| judgment tier | `haiku` | `sonnet` |
| terms | `--mint-budget 3 --k-max 2` | defaults (15 / 6) |

`judgmentModel` is set EXPLICITLY, and that is the point of the mode. Before both tiers were
required, omitting it let the judgment seats inherit the session model — measured at 68% of one
run's wall clock, 92 of 136 minutes across seven serial seats — so a "smoke" was two-thirds full
strength and no faster to iterate on than a keeper.

```sh
MODEL=haiku JUDGMENT_MODEL=haiku LANES=1 \
  scripts/universe.sh run "<topic>" --mint-budget 3 --k-max 2
```

Measured for scale on one elementary topic: the development configuration cost **$42.96** and 99.5
minutes; a haiku smoke of the same topic reached four epochs, minted and audited gaps, and cost
**$7.01** for 18 sittings.

### Reading a smoke's result

- **A refused seat and a barren seat return the same nothing.** A seat whose request is declined
  produces zero output tokens; a lens that audited and found nothing also files nothing. The engine
  continues either way — it will synthesise a report over lane drafts that do not exist and pass
  every gate. Check seat exit codes and output-token counts, not just that the run finished.
- **A `report_voice` gap is `by_grade`,** so a voice defect that grades below material does not hold
  the PASS gate: a run can ship VERIFIED with known voice defects open. That is a deliberate
  setting, not an accident — decide whether it is the one you want before reading the stamp as
  clean.
- Archive the run before the container is reclaimed. `research/` is gitignored, so `run-archive/`
  is the only part that outlives it — records and computations only, gzipped; the fetched-source cache is
  re-fetchable and about 100× larger.

## Environment that will bite you

**Clear the seat identity before anything.** A shell that has touched a run has `FEOV_RUN` set, and
it is inherited by everything — including `go test`. The tool then refuses a run directory it was
handed, and the failure reads exactly like a code defect. Measured: a full gate run failed a refusal
assertion for this reason and the diagnosis cost an afternoon.

```sh
env -u FEOV_RUN -u FEOV_RUN_FROM_WRAPPER -u FEOV_AGENT_ID -u FEOV_AGENT_TYPE -u FEOV_HOOK_VERSION …
```

`universe.sh` does this itself. Do it by hand for `go -C scripts run ./check` and for any
`feov-record` call against an archived run.

**A run's members are its lead session's agents.** `setup` records the session it runs in
(`CLAUDE_CODE_SESSION_ID`) on the run's `.claude/run-live.json` row as the run's lead session. While
that row is live — the run's directory exists and its outcome is not yet recorded — the
SubagentStart/Stop, PreToolUse, PostToolUse and PostToolUseFailure hooks attribute a call to the run only when the
payload's `session_id` is the run's lead session; every subagent, the Workflow tool's seats included,
carries its dispatching session's id. A dev session in the same checkout or a worktree nested under
it is not attributed and gets no `FEOV_RUN`. Between the outcome and the capture the run's lead session's
`lead-judge` subagents are still attributed (the bench's assembly sitting), and once capture has run
nothing is. A payload with no `session_id`, or one of this plugin's agents dispatched by a session
that leads no live run there, is attributed to nothing and recorded on the hooks' failure record
(stage `run-membership`). A run resumed in a NEW session, once the run's lead session has ended, is
attributed only after the new session runs
`feov-record take-over <runDir> --from-session <the run's lead session> --seat-id operator`; the
take-over names the session it replaces and cannot see whether that session is still running, so it
is for a resume only. `setup` refuses a second live run in one project, whatever session leads it —
two runs share the project's run directories and the memory capture writes — a re-setup of a run
another session leads, and a run whose outcome is recorded and that was not captured; it prunes the
rows of runs that were captured or whose directory is gone. Run a second engine from its own
checkout.

**The box is shared.** Other sessions run their own gates here. Check `uptime` before timing
anything: a load average of 30 on four cores makes `go test` hit its 600s timeout and look like a
hang. Gates are void if the tree changes or the branch switches under them, and two gates writing
one output file interleave into nonsense.

## Reading a gate or a CI result

- `go -C scripts run ./check` is every CI gate that can run locally, at `-count=1`. **Read the fail
  count, not the exit code**: `cmd | tail` reports `tail`'s status, and a wrapped
  `gh pr checks --watch` exits 0 even when a check is red. PR #967 was merged with a red `qlty`
  for exactly that reason.
- `check` skips `qlty`, and a local `qlty` does not run `actionlint` — so a workflow edit needs CI's
  `qlty` before any green claim.
- A red under load, or with a dirty environment, is **unproven** rather than real. Reproduce it on a
  quiet box with a cleared environment before believing it.

## Cost discipline

Spend scales with **context size × number of turns**, not with output. Each turn of a long session
re-reads the whole context from cache. Measured over one week: dev sessions and their subagents read
6,004M cached tokens against the smoke runs' 408M.

- Prefer short, fresh sessions per workstream, handing off through a checkpoint file.
- Batch work into fewer, larger turns. **Wait for a background job's notification instead of
  polling it** — peeking is one of the most expensive habits available.
- Delegate bulk reading to a subagent that returns conclusions and pointers, not dumps.
- Do not resume a subagent across a long gap: subagent cache writes are 5-minute TTL (the 1-hour TTL
  is the main session's only), so a resume re-writes the whole grown transcript. Hand a fresh agent
  the artifacts instead.
- A smoke run is meant to be cheap. The engine's own `--smoke` is one lane, `haiku` on **both**
  tiers, `--mint-budget 3 --k-max 2`. Running it at the development configuration — three lanes,
  `sonnet` on both tiers — costs roughly an order of magnitude more and is not a smoke.
