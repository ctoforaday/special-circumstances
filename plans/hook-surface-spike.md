# Hook-surface spike record — Phase 0

> Measured hook payloads from a live client. This is the **evidence** both `gray-area.md` and
> `context-checkpointing.md` build against; it is not a design document.
> Re-run it against whatever client a consumer actually runs — that is the whole point (R9/G5).

**Client:** Claude Code **2.1.220**, native binary, `GIT_SHA 4073f595…`, built 2026-07-24, Linux.
**Method:** a scratch project with `.claude/settings.json` registering a logging command hook on
twelve events, driven by headless `claude -p` runs. Every field below was read out of the hook's
own stdin, not from documentation and not from the binary.

Auto-compaction was forced with `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE=1`, which is how the compaction
pair was exercised at all in a short session.

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

Observed across runs: 692–11,882 characters. The summary is **not** a fixed-size digest, so a
`PostCompact` hook that diffs the checkpoint against it must handle a summary an order of magnitude
smaller than the one before it — a short summary drops more, so it needs *more* injected, not less.

**`compact_summary` is real and substantial.** The summary opens `<analysis>` and walks the
conversation chronologically. `custom_instructions` is null on an `auto` trigger, as documented —
it carries what the user typed after `/compact`, so only the manual path populates it.

## 3. The finding that corrects our own design

**`SessionStart` fires with `source: compact` *in addition to* `PostCompact`, on the same boundary.**
Observed sources across the compaction run, in order:

```
startup, compact, compact, compact
```

`plans/context-checkpointing.md` §3(C)/§7 splits restore as *"`PostCompact` owns the compaction
case; `SessionStart` owns `resume`/`fork`/`startup`."* That split is correct as an intention and
**unsafe as an implementation**: both hooks fire at a compaction, so a naive pair injects twice and
reproduces R4 — the double narrative the correction claimed to retire.

**Required:** the `SessionStart` restore hook MUST no-op on `source == "compact"` and leave that
boundary to `PostCompact`. Not an optimization — without it the design is worse than the single-hook
version it replaced.

This is exactly what Phase 0 exists for. The correction was right about the mechanism and wrong
about the wiring, and only running it showed the difference.

## 4. Also worth recording

- **`SessionEnd.reason` was `other`** for a headless `-p` run — not one of `clear` / `logout` /
  `prompt_input_exit`. A `SessionEnd` seal that matches only the interactive reasons will never fire
  in headless or scheduled runs, which are the sessions with no human to notice.
- **Auto-compaction can thrash.** With the threshold forced to 1% the client aborted with
  *"Autocompact is thrashing: the context refilled to the limit within 3 turns of the previous
  compact, 3 times in a row."* A seal hook that does real work on every `PreCompact` will be invoked
  repeatedly in quick succession under context pressure. Keep it cheap and idempotent.
- **Adaptive thinking skipped a trivial subagent turn entirely** — the seat transcript had zero
  thinking blocks, where a forced budget produced a non-empty summary in earlier measurement. This
  confirms `reasoning-telemetry.md` risk **T3**: absence of a summary is not evidence of absence of
  reasoning, and "was thinking configured, at what budget" must be recorded alongside any mining
  result.

## 5. Status of the load-bearing claims

| Claim | Status |
|---|---|
| `SubagentStop` hands over a usable per-seat transcript path | **VERIFIED** — file opened and parsed |
| `PostCompact` receives the summary | **VERIFIED** — 3 summaries, 692–11,882 chars |
| `SubagentStart` can inject per-seat context | Event and fields verified; injection not exercised |
| `PostToolUseFailure` carries what a strike counter needs | **VERIFIED**, plus `is_interrupt` |
| `SessionStart` still fires on `compact` | **VERIFIED** — and it forces a design change (§3) |
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
the runs settled, the pairs were present. Two rules follow for anyone re-running this spike: give
each run its own log file, and confirm the process has exited before parsing. A concurrency artifact
is exactly the kind of thing that reads as a causal finding.
