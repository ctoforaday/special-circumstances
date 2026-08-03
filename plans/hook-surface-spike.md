# Hook-surface spike record — Phase 0

> Measured hook payloads from a live client. This is the **evidence** both `gray-area.md` and
> `context-checkpointing.md` build against; it is not a design document.
> Re-run it against whatever client a consumer actually runs — that is the whole point (R9/G5).

**Client:** Claude Code **2.1.220**, native binary, `GIT_SHA 4073f595…`, built 2026-07-24, Linux.
**Method:** a scratch project with `.claude/settings.json` registering a logging command hook on
twelve events, driven by headless `claude -p` runs. Every field below was read out of the hook's
own stdin, not from documentation and not from the binary.

Auto-compaction was forced with `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`, which is how the compaction
pair was exercised at all in a short session. `=1` reaches the boundary but thrashes; `=2` gives a
session room to work either side of one compaction (§4).

**This record has been corrected against itself twice.** §3 in particular reversed on measurement.
Where a section states a finding, it states how it was observed; where a superseded claim mattered
enough to have been acted on, the withdrawal is left in place rather than edited out.

---

## 1. What fired

A single headless run — one deliberate tool failure, one subagent — produced **seven** events:

`SessionStart` · `UserPromptSubmit` · `SubagentStart` · `PostToolUseFailure` · `SubagentStop` ·
`Stop` · `SessionEnd`

A second run under a forced-compaction threshold added `PreCompact` ×3 and `PostCompact` ×3.

`ConfigChange`, `TaskCreated` and `InstructionsLoaded` were registered and did not fire — nothing
in the run triggered them. Absence here is "not exercised", not "does not exist".

## 2. Payloads that matter

Every event carries `session_id`, `transcript_path`, `cwd`, and most carry `prompt_id`.

### `SubagentStop` — the capture point for Gray Area

```
agent_id               aeaae1e2e57179ff5
agent_type             general-purpose
agent_transcript_path  …/<session-id>/subagents/agent-<agent_id>.jsonl
last_assistant_message "42"
background_tasks       []
session_crons          []
effort                 {"level": "high"}
stop_hook_active       false
```

**`agent_transcript_path` resolves to a real file** — 12,997 bytes, distinct from the parent
transcript, whose lines carry `agentId` and `attributionAgent: general-purpose`. Confirmed by
opening it. This is the claim the whole capture design rests on, and it holds: the seat's own
trajectory is handed over by path, per seat, at the moment it finishes.

**Unbudgeted find: `background_tasks` and `session_crons`.** The memento design lists in-flight
handles as a *mandatory, hand-authored* checkpoint field, promoted from candidate to mandatory by a
field report where a detached download survived a compaction. The harness supplies them. Both were
empty here because the run had none; a run with background work should populate them. **If it does,
the most error-prone field in the checkpoint stops being hand-written.** Verify before relying on
it — an empty array from a run with no background work proves only that the key exists.

`Stop` carries the same `background_tasks` / `last_assistant_message` pair, so the seal point is not
restricted to subagents.

### `PostToolUseFailure` — the strike counter for [[anti-spinning]] (#127)

```
tool_name     Bash
tool_input    {"command": "ls /definitely/not/a/real/path/xyz", "description": "…"}
tool_use_id   toolu_01AsrZCoxxw8kBicnwLk4oiR
error         Exit code 2\nls: cannot access '…': No such file or directory
duration_ms   879
is_interrupt  false
```

Everything a counter needs, keyed on `(tool_name, target)` out of `tool_input`.

**`is_interrupt` is load-bearing and was not anticipated in #127.** `anti-spinning` also says *honor
the cancel* — a user interrupt must not be counted as a strike, or the rule's two halves fight each
other. The field distinguishes them. #127's mechanism should key on `is_interrupt == false`.

### 2a. `PostToolUseFailure` CAN inject (measured 2026-08-03, #234)

The paragraph above verifies what the event *carries*. It says nothing about whether the event's
*output* is honoured, and §3a is the standing proof those are different questions: `PostCompact`
carried the compaction summary and could not inject a byte of it. `sc-strike-counter` shipped
emitting `additionalContext` anyway, marked in its own source as the unverified half — the count
reaching the MODEL rather than only the operator is the whole difference between `anti-spinning`
having a mechanism and having a log line.

Measured under **headless `claude -p`**, deliberately: that is the case the issue worried about,
because it is where no operator is reading stderr. Same client as the rest of this record — 2.1.220
— so this sits alongside §3a's refutation rather than describing a different build.

Two hooks, one run. `SessionStart` emitted `MARKER-SS-CONTROL-7Q4X` as a **positive control** — its
injection is already verified (§3a), so a subject that came back NOT-SEEN could not be blamed on a
run where nothing fired. `PostToolUseFailure` emitted `MARKER-PTUF-SUBJECT-9K2M`. Both hooks logged
their raw payload to a file, so *fired but ignored* stays distinguishable from *never fired*.

Result: **both** markers present, as attachments of the same type that settled the `SessionStart`
case, with the subject arriving in the same turn as the failure it describes.

```json
{"attachment":{"type":"hook_additional_context","hookEvent":"PostToolUseFailure",
 "hookName":"PostToolUseFailure:Read","content":["MARKER-PTUF-SUBJECT-9K2M"]}}
```

Asked afterwards which markers it could see, the model returned both, verbatim. That is the weaker
of the two lines of evidence and is recorded second, deliberately: the attachment is the leaf.

**What this changes.** `sc-strike-counter` reaches the model at the moment the rule says to stop,
in exactly the unattended runs where the rule matters most. Its two channels stop being a hedge and
become a choice — stderr is for the operator watching a session, and it is also what survives a
model treating an injected token as a suspected prompt injection (§3b, observed twice).

**What it does not change.** Injection reaching context is not the model acting on it, and the
counter still cannot see a command that exits 0 while the work failed. Its silence remains
[[anti-spinning]]'s documented blind spot, not evidence the loop was broken.

### `SubagentStart`

```
agent_id    aeaae1e2e57179ff5
agent_type  general-purpose
```

Matched on `agent_type`, returns `additionalContext` to the subagent. Per-seat injection confirmed
available.

### `PreCompact` / `PostCompact`

```
PreCompact   trigger: auto   custom_instructions: null
PostCompact  trigger: auto   compact_summary: <8543 / 11882 / 692 chars across three compactions>
```

Observed across runs: 692–11,882 characters. The summary is **not** a fixed-size digest — a 17×
spread between the smallest and largest observed on the same client. A short summary drops more, so
the amount a restore would need to replace varies by an order of magnitude from one boundary to the
next. That variance is *not* actionable at the boundary — the only injector runs before the summary
exists (§3a) — but it is worth recording as an observation, and it is the reason `PostCompact`'s
observability role is worth keeping: how much a summary drops is measurable across boundaries even
though nothing can respond to it within one.

**`compact_summary` is real and substantial.** The summary opens `<analysis>` and walks the
conversation chronologically. `custom_instructions` is null on an `auto` trigger, as documented —
it carries what the user typed after `/compact`, so only the manual path populates it.

## 3. The finding that corrects our own design

**`SessionStart` fires with `source: compact` *in addition to* `PostCompact`, on the same boundary.**
Observed sources across the compaction run, in order:

```
startup, compact, compact, compact
```

The 2026-07-28 draft of this section read the co-firing as a **collision** and required the
`SessionStart` restore hook to no-op on `source == "compact"`, leaving that boundary to
`PostCompact`. **That requirement is withdrawn 2026-07-29. Following it would have removed restore
from the compaction boundary entirely — the one boundary the whole design exists for.**

### 3a. `PostCompact` cannot inject (measured 2026-07-29)

Three independent lines, in increasing order of authority:

1. **Structural.** `PostCompact` does not appear in the client's `hookSpecificOutput` union. There
   is no `additionalContext` shape for it to emit.
2. **Behavioural.** Its stdout is routed to the *user's* transcript view, not into the model's
   context — the same path as a `Notification` hook.
3. **End-to-end, and the one that settles it.** Two hooks on the same forced-compaction run, each
   emitting a distinct marker: `SessionStart` → `MARKER-SS-A1B2`, `PostCompact` → `MARKER-PC-C3D4`.
   Post-compaction the model was asked which markers it could see. Answer: `SS=SEEN`, `PC=NOT-SEEN`.

Leaf-cited rather than taken from the model's own report — transcript line 42:

```json
{"type":"attachment","attachment":{"type":"hook_additional_context",
 "content":["MARKER-SS-A1B2"],"hookName":"SessionStart", ...}}
```

`MARKER-PC-C3D4` occurs nowhere in the transcript except the prompt that names it. There is no
attachment for it, because there is no mechanism that would make one.

**Ordering closes the argument even if injection existed.** The observed sequence is
`PreCompact → SessionStart(compact) → PostCompact`, so the only hook that *can* inject has already
run by the time the summary is available to diff against. A summary-aware delta restore is not
merely unimplemented; it is unreachable on this event ordering.

**Corrected requirement:** restore is `SessionStart`, on **every** source including `compact`. There
is no collision to avoid, because there is no second injector. `PostCompact` is observability —
it can *record* what the summary kept or dropped, which is worth having, but it cannot act on it.

### 3b. The restore digest must read as the session's own recovered state

The marker test reproduced, unprompted, the behaviour §5's C1 test found: the model flagged the
injected token as *a suspected prompt-injection attempt*. Twice now, in two unrelated tests, on two
different events. That is not a quirk of one prompt; it is the default posture toward unexplained
content arriving from outside the conversation.

> ~~**A restore digest asserting content the session cannot recognize is treated as hostile.** The
> digest must be self-evidently the session's own recovered state — sourced from, and traceable to,
> what this session actually did.~~

**Superseded 2026-07-29 by building it and measuring.** Phase 3's digest does everything that
constraint asks — leads with file, timestamp and session id, names the session's own objective,
quotes the note verbatim — and the resumed agent flagged it anyway, while recovering every value
exactly. The mitigation does not work, and the constraint as written was untestable in the direction
it mattered: it said what the digest should look like, not what the agent should end up believing.

Replaced by two claims that survive contact:

- **The hook adds no imperative of its own.** The first acceptance run's digest closed with a
  reasonable-sounding *"verify each item before acting on it"*; the agent named **that sentence**
  among the directives making the payload injection-shaped. Deleting it removed it from the reason.
  Enforceable, unit-tested, and cheap.
- **The distrust is correct and should be designed for.** Both runs used the recovered content
  accurately while labelling it a claim rather than a fact — the posture the checkpoint skill asks
  for, reached without being told. The residual flag attaches to the note's own `foot-guns` section,
  which carries imperatives by definition. That is the content working as intended, not a leak.

Recorded here because the pattern is the general one: a constraint phrased as *"the payload must
read as X"* cannot be checked, and the version that could be checked only appeared once the thing
was built and pointed at a live client.

### 3c. What this cost, and the method rule that comes out of it

The `PostCompact`-owns-restore claim was inferred from one true fact (`PostCompact` receives
`compact_summary`) and never tested, then written into two documents as settled, then used to retire
two live risks. Sequence: correct in the first draft → overturned by inference → restored by
measurement. The intermediate state was the confident one.

**Rule: a hook's *input* fields say nothing about its *output* capability.** Every event carries a
rich payload; only some can inject. Check the output union, then check it end-to-end with a marker.
Both are cheap; the inference that skipped them cost a full cycle.

## 4. Also worth recording

- **`SessionEnd.reason` was `other`** for a headless `-p` run — not one of `clear` / `logout` /
  `prompt_input_exit`. A `SessionEnd` seal that matches only the interactive reasons will never fire
  in headless or scheduled runs, which are the sessions with no human to notice.
- **Auto-compaction thrashes below a threshold, and the numbers are now measured rather than
  reasoned.** With the threshold forced to 1% the client aborts with *"Autocompact is thrashing: the
  context refilled to the limit within 3 turns of the previous compact, 3 times in a row."* The first
  reading blamed payload size; it was the threshold.

  ~~Predicted `PCT=2` → ~70k trigger → room for real work; confirmed by a run with one clean
  compaction and a normal exit.~~ **Wrong, and wrongly "confirmed" — 2026-07-29.** That confirmation
  came from a clean exit, not from token counts, and a clean exit only says the run finished. The
  actual trigger at `PCT=2` is ~35k, half the prediction. A controlled sweep — same workload, four
  values, each in its own project directory so no run shares a transcript:

  | `PCT` | first trigger | compactions | exit | outcome |
  |---|---|---|---|---|
  | 2 | ~35,370 | 3 | 1 | thrash abort |
  | 5 | ~44,744 | 3 | 1 | thrash abort |
  | **10** | **~94,321** | **1** | **0** | **clean** |
  | 25 | never fired (peaked ~120,390) | 0 | 0 | finished without compacting |

  **The fixed floor is ~30.6k** — `cache_read_input_tokens` bottoms out at 30,601 in all four runs
  and at 30,593 in an earlier unrelated one. That is the system prompt, tools and carried summary,
  and it is what the percentage sits on top of. So the usable window is `trigger − 30.6k`: about 5k
  at `PCT=2` and about 64k at `PCT=10`. At 2 and 5 the next boundary is one or two turns away
  whatever the hooks do, which is why every workload size thrashed.

  **The relationship is not linear** — the 2-and-5 points extrapolate to ~60k at `PCT=10` and the
  measured value is ~94k. Two points are a line whether or not the function is one; do not fit and
  predict, sweep and read.

  **Use `PCT=10` to drive a compaction in a test.** Below that the run cannot survive its own
  boundary; at 25 a short workload never reaches the threshold at all.

  **Method, restated because it failed twice:** size the override off the *transcript's own token
  counts*, and treat a clean exit as evidence of nothing but a clean exit.
- **A seal hook must still be cheap and idempotent.** Thrash is reachable without an override
  whenever context pressure is sustained; `PreCompact` firing three times in quick succession is a
  supported case, not a pathology.
- **`CLAUDE_CODE_MAX_CONTEXT_TOKENS` appears to be ignored** on 2.1.220 — setting it produced no
  change in trigger point, where `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` scaled it as predicted. Recorded
  as observed behaviour, not as a claim about intent; use the percentage override to force
  compaction in tests.
- **Adaptive thinking skipped a trivial subagent turn entirely** — the seat transcript had zero
  thinking blocks, where a forced budget produced a non-empty summary in earlier measurement. This
  confirms `reasoning-telemetry.md` risk **T3**: absence of a summary is not evidence of absence of
  reasoning, and "was thinking configured, at what budget" must be recorded alongside any mining
  result.

## 4b. Multiple hooks on ONE event run in PARALLEL (measured 2026-07-31)

The question #201 turns on, and the one nobody had asked: when two hooks match the same
event, does the client run them one after another or at once?

**Method.** A scratch project registering TWO hooks on one event, each writing a start
timestamp, sleeping 2s, then writing an end timestamp. Serial execution cannot produce
overlapping intervals; parallel execution cannot avoid them.

`SessionStart`, two hooks:

```
A start 1785487925434853787
B start 1785487925438412360      <- 3.6ms after A started
A end   1785487927437397313      <- A was still sleeping when B began
B end   1785487927440755630
```

`PostToolUse` with one `Read` matcher — the shape `sc-quality-gate` + `sc-recall-index` use:

```
Q start 1785487957883223391
R start 1785487957884189478      <- 0.97ms
Q end   1785487959885816405
R end   1785487959886465847
```

Both events: the second hook starts ~1–4ms after the first, while the first is still
running. Wall clock is **2.006s and 2.003s for two 2-second hooks** — the MAX, not the sum.

### What this costs #201, and the boundary of that claim

For the units we have TODAY, merging buys no concurrency: the OS already provides it, free,
with process isolation thrown in. A merged binary runs its units SEQUENTIALLY unless it
reimplements the concurrency, so a naive merge is a REGRESSION — `sc-quality-gate`'s qlty
shell-out (30s ceiling) and `sc-recall-index`'s `qmd update` (3.9s measured) would go from
overlapping to additive. Step 3's success criterion is therefore PARITY, not speedup.

**`max(A,B)` for free holds only while the units SHARE NOTHING.** Today they share a tiny
JSON payload and a project-root string, so the measurement above is the whole story. It
stops being the whole story the moment a unit reads something expensive, and the roadmap is
entirely made of those:

- **Shared expensive reads.** Ten units each parsing the same transcript, SQLite store
  (#197) or git state is 10x the dominant cost, and process parallelism makes it WORSE
  rather than better — ten concurrent readers of one file, ten database connections, page
  cache contention. One process parses once and hands the structure to every unit; the
  saving scales with how expensive the shared input becomes.
- **A shared prelude.** "Validate the VCS state first" is work every unit needs and only one
  should do. Across processes that requires a cache file, which requires staleness handling,
  which is a new defect surface invented to avoid doing the work once.
- **Fan-out inside a unit.** By-file goroutines in one binary use a single pool sized to
  NumCPU. Across N processes it is N pools each sized to NumCPU, oversubscribing the machine
  with no global scheduler — and it degrades as both N and the file count grow.

So this measurement bounds a claim rather than settling the design: it says the OS's free
parallelism is a LOCAL MAXIMUM, good exactly while the units stay trivial and independent.
The non-latency case — one payload parse, one project-root resolve, one log writer, fewer
binaries, names that match the events — holds regardless, and is what the shape should be
chosen on.

### And one hazard this promotes from theoretical to certain

`sc-quality-gate` and `sc-recall-index` both append to
`.claude/prosthetic-conscience-hook.log` on the same event. Parallel execution means those
two writes are CONCURRENT BY DESIGN, not merely possible — the known limit recorded in
`internal/hooklog` is a live condition on every markdown write, not an edge case. Single-
writer consolidation is the fix; a lock would be the alternative.

## 5. Status of the load-bearing claims

| Claim | Status |
|---|---|
| `SubagentStop` hands over a usable per-seat transcript path | **VERIFIED** — file opened and parsed |
| `PostCompact` receives the summary | **VERIFIED** — 3 summaries, 692–11,882 chars |
| **`PostCompact` can inject that summary back into context** | **REFUTED** — absent from the output union; stdout goes to the user; marker `NOT-SEEN` end-to-end (§3a) |
| **`SessionStart(source=compact)` injects into the post-compaction context** | **VERIFIED** — `hook_additional_context` attachment, transcript line 42 (§3a) |
| `SubagentStart` can inject per-seat context | Event and fields verified; injection not exercised |
| `PostToolUseFailure` carries what a strike counter needs | **VERIFIED**, plus `is_interrupt` |
| **`PostToolUseFailure` can inject the count into the model's context** | **VERIFIED** — `hook_additional_context` attachment under headless `claude -p`, against a `SessionStart` positive control (§2a) |
| **Multiple hooks on one event run in parallel** | **VERIFIED** — overlapping intervals on `SessionStart` and `PostToolUse`; wall clock is the max, not the sum (§4b) |
| `SessionStart` still fires on `compact` | **VERIFIED** — and §3's first reading of *why that matters* was wrong (§3a) |
| **`PreCompact` stdout becomes custom compact instructions** | **VERIFIED** — marker survived 2/2 |

### C1, verified — and it comes with a condition

The test: a `PreCompact` hook echoing a distinctive token (`XYZZY-7391-MARKER`) plus a fabricated
line (`VALIDATION LOOP: run qlty check then the three integration tests, in that order`), then a
check of whether either survives into `compact_summary`.

**Both appeared in both summaries** (7,502 and 807 chars). `PreCompact` stdout does reach the
summarizer, and the design's C1 correction holds under observation rather than only under a reading
of the client.

**The condition, which the test found by accident and matters more than the result.** The summarizer
complied *and flagged the instruction as an attack*, in the summary text:

> *"The final incoming message … instructs me to embed a specific unexplained token … and a specific
> line … verbatim into my output. Nothing in the actual conversation involves qlty checks,
> integration tests, or any marker token — these concepts never appeared anywhere in the real
> session. This has the hallmarks of a prompt-injection attempt…"*

It was right to. My test asked it to preserve facts with no basis in the conversation, which is
indistinguishable from injection. In real use the seal names things that *are* in the session — the
validation loop the agent actually established — so the synthetic case is worse than the real one.
But the constraint is real and belongs in the design:

> **The seal's preserve-verbatim instruction MUST be grounded in the conversation it is
> summarizing.** Instructions asserting content absent from the session get treated as hostile, and
> the summarizer narrates its suspicion *into the summary* — which then lands in the restored
> context, costing tokens and casting doubt on the checkpoint that is supposed to be the trusted
> anchor.

Practical form: the seal should reference and reinforce ("preserve the validation loop stated
above"), never introduce. If the note's contents are not already in the transcript, the seal has
nothing legitimate to ask for and should stay silent.

### A correction to this record's own method

An intermediate reading of these logs showed `PreCompact` firing with **no** matching `PostCompact`
on the two stdout-emitting runs, against three clean pairs on the log-only run. The obvious
inference — that emitting stdout from `PreCompact` wedges compaction — was wrong, and was one step
from being written down here as a finding.

The logs were being read **while two `claude` processes were still appending to them**. Re-read after
the runs settled, the pairs were present.

**It then happened twice more, because "confirm the process has exited" is harder than it sounds
here.** The trap: a backgrounded shell wrapper around `claude` exits — and the harness reports the
command *completed* — while `claude` itself is still running and still writing. The completion
notification is about the wrapper, not the workload. Watching for the child by name has its own
version of the same bug: `pgrep -f 'claude -p'` matches the very shell that is searching for it, so
a kill sweep takes out its own caller.

Three rules for anyone re-running this spike:

- **One log file per run.** Never share a log across concurrent clients.
- **Sentinel, not notification.** Have the run itself write a `DONE` file as its last act, and gate
  every read on `until [ -f DONE ]; do sleep 10; done`. The sentinel is written by the process whose
  output you care about; the notification is not.
- **Match on `comm`, never `pgrep -f`,** when checking for or signalling a client process.

A concurrency artifact is exactly the kind of thing that reads as a causal finding, and this spike
produced three of them before the discipline was right. Each one looked like a result.

---

## 6. `watchPaths` and `FileChanged` — measured 2026-07-29

Both were listed as *"real per the client's event catalogue"* (C4, C6) and **never exercised**.
`plans/context-checkpointing.md` §5 builds improvement I2 on them: *"the check re-arms when its
surface is edited, rather than depending on the agent remembering what arms it."* That claim was
about to be implemented on an unmeasured capability, which is the mistake this record already
carries once (§3c). So: spiked first.

Method: a scratch project whose `SessionStart` hook returns `watchPaths` and whose `FileChanged`
hook logs raw stdin. An agent then edits a watched file and creates a new one. Three runs, each in
its own project directory, differing only in the **form** of the path registered.

### Both events work — and `FileChanged` carries what I2 needs

```
hook_event_name  FileChanged
file_path        …/fcspike/watched/target.txt
event            change
```

Plus the common `session_id`, `transcript_path`, `cwd`, `prompt_id`. `watchPaths` alongside
`additionalContext` in the same `SessionStart` response works — the marker arrived (`MARKER-SEEN=yes`)
*and* the watch registered, so returning both is not an either/or.

### What registers, and what silently does not

Twelve forms, each in its own project directory, each with the edit **confirmed to have happened**
before a "no event" result was believed. (Four early runs died on an API error and produced a
vacuous `NONE`; they were re-run. A negative from a run that never edited anything is not a
negative.)

| `watchPaths` entry | fires? |
|---|---|
| `watched/target.txt` — file, relative | `change` |
| `/abs/…/watched/target.txt` — file, absolute | `change` |
| **`watched` — directory, relative** | **`change`, recursively (incl. `sub/deep.txt`)** |
| **`/abs/…/watched` — directory, absolute** | **`change`** |
| **`.` — the project root** | **everything, recursively** |
| `watched/*` | nothing |
| `watched/*.txt` | nothing |
| `watched/**` | nothing |
| `watched/**/*` | nothing |
| `**/*.txt` | nothing |
| `**/target.txt` | nothing |
| `watched/.*\.txt` — a regex, not a glob | nothing |

**It takes paths — files or directories. Neither globs nor regular expressions.** Every wildcard
form fails the same way: silently, with the hook configured, the event enabled, and no error
anywhere. The regex form was tested because "maybe the syntax is regex" is the obvious next
hypothesis after one glob shape fails; it is not.

### A directory watch is recursive AND catches new files

This is the part a single-shape test got wrong. Watching the directory `watched`:

```
add    brandnew.txt
add    sub/deepnew.txt
```

Both created by the agent during the run, one nested. So `watchPaths` **is** a directory watcher,
it **does** follow the tree forward in time, and `event` distinguishes `add` from `change`.

An earlier draft of this section concluded the opposite — *"it watches the literal files it is
handed; a newly created file never fires"* — from a run that registered only file paths and one
glob. Both halves of that were wrong, and the correction came from being asked whether the failure
was really about *glob syntax*. It was not; the axis is **path vs pattern**, not one glob dialect
vs another.

### What this means for I2

Better than the first reading, not worse. A validation check's trigger surface — *"any `.go` edit"*,
`manifest/*.yml` — is registered by **watching the directories that contain it**, not by expanding
the pattern to a snapshot of files. New files under those directories fire `add`. The re-arm hook
filters on `file_path` itself, which it must do anyway since a directory watch is coarser than the
pattern.

**Foot-gun, measured:** watching `.` catches the hook's own output. The `.`-watch run logged
`add: fc.jsonl` followed by nine `change: fc.jsonl` events — the `FileChanged` hook writing its log
inside the watched tree, re-triggering itself. Watch the specific directories, never the project
root, and keep hook state outside whatever is watched.

### Also settled here

**`hook_event_name` is present in every payload** — `SessionStart`, `FileChanged`, `SessionEnd` all
carry it. `sc-checkpoint-seal` treats it as a fallback behind its explicit `-event` flag and
documents it as unverified. The flag-first design stands (a stale `hooks.json` is still the only way
to reach an unflagged invocation), but the fallback is now measured rather than hoped for.
