# Local development — running the engine against a checkout

The gates tell you a change compiles and that its goldens moved. They do not tell you the research
engine still runs. For that the plugins have to be **installed** — content copied into a versioned
cache, hook binaries built there, seats dispatched as real subagents — and then driven the way a
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
scripts/universe.sh run "<topic>"    # one `claude -p`, invoking /research as a user would
scripts/universe.sh watch [secs]     # the BOARD, live, read-only from the record
scripts/universe.sh shell            # interactive claude inside the universe
scripts/universe.sh doctor           # what is installed, binaries per plugin, schema epoch
```

`build` is isolated by default: `CLAUDE_CONFIG_DIR` and `CLAUDE_CODE_PLUGIN_CACHE_DIR` point inside
`WORKDIR` (default `~/.claude/scratch/universe`), so nothing touches `~/.claude` and other sessions
on the box are unaffected. Use `--live` only when you want the change in *your* next session, and
remember it is shared state.

Tearing one down is **two** paths, not one: `$WORKDIR` and `$WORKDIR.marketplace`. The staged
manifest sits beside the universe rather than inside it because it holds a symlink into the
checkout, and a lead exploring its own `WORKDIR` under `bypassPermissions` will find such a door and
run the checkout's binaries through it. `build` prints both paths when it finishes.

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

**After any `build`, prove the universe carries your change** rather than trusting the summary:

```sh
U=~/.claude/scratch/universe/plugins/cache/special-circumstances
grep -c '<a phrase from your change>' $U/frank-exchange-of-views/*/skills/research-protocol/scripts/debate.js
```

### Binaries are built in the cache

`${CLAUDE_PLUGIN_ROOT}` resolves to the versioned cache, never to a checkout, so the hook binaries
must be built there — after install. A build in the checkout leaves the installed plugin with no
binaries and hooks degrading to a single stderr line, which from inside a run looks exactly like
hooks that were never configured. `scripts/bootstrap-plugins.sh` does the same thing for cloud
environments and is the older, supported path; `universe.sh` follows its ordering.

A cache build has no git context, so `feov-record --version` reports `unknown` where a release
reports its sha. **Version-skew detection is inert in a universe run** — `tool_version` and
`hook_version` on the record are not meaningful there.

## Seeing what is happening

`--output-format json` returns one object when the process exits. For a run of tens of minutes that
means no output at all until it ends, so a working run and a wedged run look identical. Two views,
answering different questions:

**The harness view** — `universe.sh run` streams `--output-format stream-json --verbose` through
`scripts/universe-stream.py`, which keeps the raw JSONL and prints dispatches, record verbs and tool
errors as they happen:

```
[  312s] DISPATCH  red-lens-voice #1
[  318s]   verb    mint
[  341s]   ERROR   refused: --quote is required
[  902s] DONE  success  turns=42  $1.23
```

**A run that dispatched no `Workflow` fails, and says so.** The engine cannot research anything
without one, so its absence is decidable from the stream — and `claude -p` exits 0 for it, because
the assistant did produce an answer. `run` returns the renderer's status when claude's is clean, so
"the plugins did not load" cannot arrive as a success with a quiet `0 dispatch(es)` line above it.

**The debate view** — `universe.sh watch` reads the record (SQLite `mode=ro`, so it cannot take a
write lock on a live run) and prints the board: seats and their sitting counts, gaps minted and
closed, the chair's verdict. It calls out **hook-written sitting spans** specifically, because that
count is zero by construction under a top-level-process harness and non-zero here — it is the
one-glance evidence that subagent dispatch is real.

The engine also ships its own instruments: `feov-record dashboard <runDir> <transcriptDir> --watch`
renders board-mass trend, live seats with a completion ETA, cost-so-far and the run's scorecards to
`dashboard.html` every 15s, and `--serve <port>` hosts it over HTTPS behind a per-run secret URL.

## The universe is a snapshot, so you can keep working

`claude plugin install` COPIES content into the versioned cache — proven by inode, not assumed:
the cache file and the checkout file are different inodes, and the symlink exists only in the
staged marketplace, which is the install *source*.

So once `build` has run you can keep editing, committing, rebasing and pushing the checkout while a
run is in flight, and the running engine is unaffected. It is holding a frozen copy. Re-run `build`
when you want the new code, and the next `run` picks it up.

It isolates the other way too: the run's source tree is `$WORKDIR/src`, so its `.claude/run-live.json`
sits there and cannot claim the subagents of the session you are developing in.

Hooks are universe-local, and completely so: `${CLAUDE_PLUGIN_ROOT}` resolves to *this* universe's
cache, and the `hooks.json` there points at that cache's own `bin/`. Nothing reaches outside the
`WORKDIR`.

**The exception is narrower than it first looks: `build` mutates the universe IN PLACE.** It
reinstalls over the same version directory, and hooks spawn fresh from that directory on every tool
call — so rebuilding under a live run swaps its binaries mid-sitting. Editing the checkout is
always safe; reinstalling over a running engine is not. `universe.sh build` now REFUSES when a run
is in flight in that `WORKDIR` rather than leaving it to this paragraph; use a second `WORKDIR` for
a second universe.

A rebuild does propagate — verified with a probe marker, installed and then changed and reinstalled,
both edits landing in the cache. There is no stale-universe trap.

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
| terms | `--mint-budget 1 --k-max 2` | defaults (5 / 6) |

`judgmentModel` is set EXPLICITLY, and that is the point of the mode. Before both tiers were
required, omitting it let the judgment seats inherit the session model — measured at 68% of one
run's wall clock, 92 of 136 minutes across seven serial seats — so a "smoke" was two-thirds full
strength and no faster to iterate on than a keeper.

```sh
MODEL=haiku JUDGMENT_MODEL=haiku LANES=1 \
  scripts/universe.sh run "<topic>" --mint-budget 1 --k-max 2
```

Measured for scale on one elementary topic: the development configuration cost **$42.96** and 99.5
minutes; a haiku smoke of the same topic reached four epochs, minted and audited gaps, and cost
**$7.01** for 18 sittings.

### Reading a smoke's result

- **A refused seat and a barren seat return the same nothing.** A seat whose request is declined
  produces zero output tokens; a lens that audited and found nothing also files nothing. The engine
  continues either way — it will synthesise a report over lane drafts that do not exist and pass
  every gate. Check seat exit codes and output-token counts, not just that the run finished.
- **`VERIFIED` does not mean every dimension was audited.** A cast narrowed with `--lens-area`
  seats fewer lenses, and the verdict now states which dimensions never sat. Read that line.
- **A `report-voice` gap is `by_grade`,** so a voice defect that grades below material does not hold
  the PASS gate: a run can ship VERIFIED with known voice defects open. That is a deliberate
  setting, not an accident — decide whether it is the one you want before reading the stamp as
  clean.
- Archive the run before the container is reclaimed. `research/` is gitignored, so `run-archive/`
  is the only part that outlives it — records and proofs only, gzipped; the fetched-source cache is
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

**A run-live marker claims every subagent in its project.** While `.claude/run-live.json` names a
run, the SubagentStart/Stop hooks attribute every subagent in that project to it — a dev session's
included. Run the engine from its own source tree (a detached worktree), never from the worktree you
develop in.

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
  tiers, `--mint-budget 1 --k-max 2`. Running it at the development configuration — three lanes,
  `sonnet` on both tiers — costs roughly an order of magnitude more and is not a smoke.
