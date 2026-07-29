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

> **A restore digest asserting content the session cannot recognize is treated as hostile.** The
> digest must be self-evidently the session's own recovered state — sourced from, and traceable to,
> what this session actually did. Same constraint as the seal (§5): reference and reinforce, never
> introduce.

This is a design constraint on Phase 3, not a caveat. A checkpoint the restored agent distrusts is
worse than no checkpoint: it costs the tokens *and* poisons the anchor.

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
- **Auto-compaction can thrash — and the arithmetic is predictable.** With the threshold forced to
  1% the client aborted with *"Autocompact is thrashing: the context refilled to the limit within 3
  turns of the previous compact, 3 times in a row."* The first reading blamed payload size. It was
  the threshold. Read off the baseline transcript's own token counts: `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE=1`
  puts the trigger at ~35k while a *post*-compaction session already starts at ~30k — 5k of headroom,
  so the next boundary is two turns away, guaranteed, regardless of what the hooks do. Predicted
  `PCT=2` → ~70k trigger → room for real work; confirmed by a run with one clean compaction and a
  normal exit. **Method:** size the override off measured token counts in a baseline transcript, not
  off a guess about how big the session feels.
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

## 5. Status of the load-bearing claims

| Claim | Status |
|---|---|
| `SubagentStop` hands over a usable per-seat transcript path | **VERIFIED** — file opened and parsed |
| `PostCompact` receives the summary | **VERIFIED** — 3 summaries, 692–11,882 chars |
| **`PostCompact` can inject that summary back into context** | **REFUTED** — absent from the output union; stdout goes to the user; marker `NOT-SEEN` end-to-end (§3a) |
| **`SessionStart(source=compact)` injects into the post-compaction context** | **VERIFIED** — `hook_additional_context` attachment, transcript line 42 (§3a) |
| `SubagentStart` can inject per-seat context | Event and fields verified; injection not exercised |
| `PostToolUseFailure` carries what a strike counter needs | **VERIFIED**, plus `is_interrupt` |
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
