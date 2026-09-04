# Hook-surface spike record — Phase 0

> STATUS 2026-09-02: shipped — historical measurement record. The findings graduated into the
> shipped hooks (`sc-strike-counter`, the consolidated `sc-posttooluse`, `sc-filechanged-rearm`,
> the stop/subagent guards); the two forward-looking recommendations are marked DONE inline.
> The measurements stand as history, per client (§5's index names which client each row is from).

> Measured hook payloads from a live client. This is the **evidence** both `gray-area.md` and
> `context-checkpointing.md` build against; it is not a design document.
> Re-run it against whatever client a consumer actually runs — that is the whole point (R9/G5).

**Client:** Claude Code **2.1.220**, native binary, `GIT_SHA 4073f595…`, built 2026-07-24, Linux —
for §1–§7. **§8 was taken on 2.1.235 and §9 on 2.1.240**, each stating its own version, because this
record has now moved under a version bump three times. Read a row's section for the client it was
measured on; §5's index says so where two clients disagree.
**Method:** a scratch project with `.claude/settings.json` registering a logging command hook on
twelve events, driven by headless `claude -p` runs. Every field below was read out of the hook's
own stdin, not from documentation and not from the binary.

Auto-compaction was forced with `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`, which is how the compaction
pair was exercised at all in a short session. `=1` reaches the boundary but thrashes; `=2` gives a
session room to work either side of one compaction (§4).

> **That lever is dead on 2.1.240 — see §11.** The variable now does nothing at any value; the
> replacement is the `--autocompact <auto|tokens>` flag. Everything §3/§3a/§4 *observed* stands; only
> the means of provoking it has changed. Rebuild any compaction harness on §11 before running it, and
> assert the boundary from the transcript rather than from a clean exit.

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

Every event carries `session_id` and `cwd`, and most carry `prompt_id`.

> **`transcript_path` is NOT universal, and this line said it was.** Measured key sets (§12):
> `SessionEnd` carries it, `PreCompact` carries it, `SubagentStop` carries it and a per-seat one
> besides — but **`SessionStart` does not** (§9e correction 3: its stdin is
> `{session_id, cwd, hook_event_name, source}`). The exception matters more than the rule, because
> `SessionStart` is the verified injection channel: a design that reads the transcript at restore
> has no path to read. Corrected here rather than only where it was discovered, since this sentence
> is what a reader checks first.

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

> **SCOPED by §7a (2026-08-15).** It holds *for a seat whose `SubagentStop` payload carries an
> `agent_type`* — as this one did, which is why the spike saw only the good case from a sample of
> one. Measured across 69 seats: the file exists for all 19 that carry a type and for **none** of
> the 50 that do not. Read the resolvability off the row, per seat; the path is announced either
> way. #189.

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
**DONE (2026-09-02):** `sc-strike-counter` decodes `is_interrupt` and does not count an
interrupt as a strike (`internal/strikecounter/main.go`).

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

> **RE-MEASURED 2026-08-23 (#500).** Registered for real from a gitignored `settings.local.json`
> and the payload captured whole. Three results:
>
> **The full key set is:** `agent_id`, `agent_type`, `cwd`, `hook_event_name`, `prompt_id`,
> `session_id`, `transcript_path`. The section above listed the two that mattered at the time; the
> rest matter now, and one absence matters most.
>
> **Nothing the workflow supplies reaches the hook.** The probe agent's prompt carried a unique
> sentinel and it appears nowhere in the payload. So a dispatcher cannot hand `SubagentStart` a
> token, and the hook can learn the run only the way `PreToolUse` already does — from the
> singleton marker via `cwd`, which is identical for every agent.
>
> **`agent_type` is a ROLE and cannot discriminate a seat.** `debate.js` maps four types onto
> every seat in a run: `red-auditor` covers every `red-lens-*` *and* `red-merge-rN`; `lead-judge`
> covers `judge-rN`, `judge-petition-*`, `judge-terminal` *and* `assemble`. Thirteen seats, four
> types. A hook cannot know which seat it is looking at.
>
> **Injection is EXERCISED, not merely available.** A hook emitting
> `{"hookSpecificOutput":{"hookEventName":"SubagentStart","additionalContext":"…"}}` produced a
> subagent that returned the marker verbatim from its own context with no tools used.

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
  compaction in tests. **SUPERSEDED for 2.1.240 (§11): the percentage override no longer scales
  anything either. Use `--autocompact <tokens>`.**
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

**DONE (2026-09-02):** shipped as the consolidated `sc-posttooluse` — one process whose units
RETURN their log lines and one writer appends them in order; the recall-index unit was retired
2026-08-04 with the retrieval layer. `internal/posttooluse/main.go` cites this section (§4b) as
the reason parity, not speedup, was the merge's floor.

## 5. Status of the load-bearing claims

| Claim | Status |
|---|---|
| ~~`SubagentStop` hands over a usable per-seat transcript path~~ | **VERIFIED for seats carrying `agent_type`; the file does not exist for seats without one** — 19/19 vs 50/50, §7a. The original reading below held for the one agent this spike ran, which carried a type |
| **`PreToolUse` carries `agent_id`** | **VERIFIED** — 9/9 subagent calls, 6 agents, 3 tool types (§7). This row previously read NOT MEASURED and had once been asserted as measured; it is the gate on #290 |
| **`agent_id` is stable per agent and distinct across concurrent agents** | **VERIFIED** — attributed by marker string, not by timing (§7) |
| **`agent_id` joins `PreToolUse` to `SubagentStop`** | **VERIFIED** — byte-identical for 6/6 (§7) |
| **`session_id` / `prompt_id` can namespace a seat** | **REFUTED** — identical across the main session and all concurrent subagents (§7) |
| **A hook change in `settings.local.json` needs a session restart** | **REFUTED** — live on the next tool call (§7) |
| `PostCompact` receives the summary | **VERIFIED** — 3 summaries, 692–11,882 chars |
| **`PostCompact` can inject that summary back into context** | **REFUTED** — absent from the output union; stdout goes to the user; marker `NOT-SEEN` end-to-end (§3a) |
| **`SessionStart(source=compact)` injects into the post-compaction context** | **VERIFIED** — `hook_additional_context` attachment, transcript line 42 (§3a) |
| **`SubagentStart` can inject per-seat context** | **VERIFIED TWICE, INDEPENDENTLY.** #500 (2026-08-23) registered from a real `settings.local.json` and had the subagent return the marker verbatim; §10 found the `hook_additional_context` attachment in the SEAT's transcript and absent from the parent's. The two runs agree, and this row previously read "injection not exercised" while §2 already called it "confirmed available" — both measurements settle which was right. #500 adds what §10 did not ask: the payload carries nothing workflow-supplied, and `agent_type` names a ROLE, not a seat |
| **`SubagentStop` can inject** | **REFUTED** — 9 firings, marker absent from parent AND seat, against a live control (§10) |
| **Injection re-arms the turn of the context the event belongs to** | **VERIFIED, and independent of delivery** — `Stop` re-arms the session (§8), `SubagentStop` re-arms the SEAT for 9 turns while delivering nothing (§10) |
| `PostToolUseFailure` carries what a strike counter needs | **VERIFIED**, plus `is_interrupt` |
| **`PostToolUseFailure` can inject the count into the model's context** | **VERIFIED** — `hook_additional_context` attachment under headless `claude -p`, against a `SessionStart` positive control (§2a) |
| **Multiple hooks on one event run in parallel** | **VERIFIED** — overlapping intervals on `SessionStart` and `PostToolUse`; wall clock is the max, not the sum (§4b) |
| `SessionStart` still fires on `compact` | **VERIFIED** — and §3's first reading of *why that matters* was wrong (§3a) |
| **`PreCompact` stdout becomes custom compact instructions** | **VERIFIED** — marker survived 2/2 |
| **`UserPromptSubmit` / `PostToolUse` / `PostToolBatch` / `Stop` can inject** | **VERIFIED** — all four, against a `SessionStart` control (§8) |
| **`Stop` injection re-invokes the model** | **VERIFIED** — 9 firings vs 1 on a null control, 8 filler turns (§8). Hazard and feature are one mechanism |
| **`stop_hook_active` flips on a `Stop` re-entry** | **VERIFIED on 2.1.240 (§13)** — `false` on the 1st firing, `true` on all 15 that followed. A guard can detect its own loop with no state |
| **The `Stop` loop's SIZE is client-specific** | **MEASURED** — 9 firings / 1,186 output tokens on 2.1.235; **16 / 4,326 on 2.1.240** (§13). Read §8's figures as that client's, not as the hazard's |
| **`Stop` carries in-flight handles** | **VERIFIED** — `background_tasks[]` with `{id,type,status,description,command}` and `session_crons[]`, on a live task (§9c) |
| **Which events carry `background_tasks`** | **MEASURED per event (§12)** — `Stop` and `SubagentStop` carry and populate it; **`PreCompact` and `SessionEnd` do not carry the key at all**. Two of the three SEALING events therefore cannot measure handles |
| **`background_tasks` includes subagents** | **VERIFIED** — a live seat appears as `{type:"subagent", agent_type:…}` in the parent's list (§12) |
| **`SubagentStop` fires at the PARENT, carrying both transcripts** | **VERIFIED on 2.1.240** — parent `session_id`/`transcript_path` plus `agent_transcript_path`, `agent_id`, `agent_type`, `last_assistant_message` (§9d). **Disagrees with §7a's 2.1.220 reading — re-measure per client** |
| **Tool events need a `matcher` key to fire** | **REFUTED** — 15/15 and 12/12 matched pairs, NOMATCHER alongside STAR (§9e) |
| **`TaskCreated` / `TaskCompleted` fire for a background task or a subagent** | **REFUTED** for both shapes, the task proven live in the `Stop` payload (§9b) |
| **`StopFailure` fires when a `Stop` hook exits non-zero** | **REFUTED** for that shape; a blocking hook is untested (§9b) |
| **`PermissionRequest` / `PermissionDenied` fire on a hook-issued deny** | **REFUTED**; the interactive shape is untested (§9b) |
| **Any payload or transcript field carries the context window** | **REFUTED** — `SessionStart` has no model field; `[1m]` is invisible in `message.model`; no `*limit*` field exists (§9e) |

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

---

## 7. `PreToolUse` carries `agent_id` — measured 2026-08-15

**This closes the gate on #290**, which named exactly one unmeasured row as blocking the whole
identity design, and which this record had previously asserted as measured when it was not.

**Method.** A `PreToolUse` hook wired in `.claude/settings.local.json` with `matcher: "*"`, writing
raw stdin to a file named by `mktemp` — unique **by construction**, never assembled from a timestamp
and a pid, because concurrent hook processes destroying each other's records is a measured failure
in this repo (#213) and a probe that loses payloads to a race reports a plausible zero. The probe
always exits 0: a `PreToolUse` hook exiting 2 blocks the call, and a probe must not influence what it
measures. Settings were backed up first and restored after; the probe is not in the tree.

Session `937047bc…`, Claude Code on the web, managed container. **The hook took effect mid-session
with no restart** — a change to `settings.local.json` was live on the next tool call.

### The payload splits in two

| field | main session | subagent |
|---|---|---|
| `agent_id` | **absent** | **present** |
| `agent_type` | absent | present |
| `effort` | present | absent |
| `session_id` | `937047bc…` | **same** |
| `prompt_id` | `844da5c2…` | **same** |
| `transcript_path` | parent `.jsonl` | **same** (the parent's, not the seat's) |

Main-session key set in full: `session_id`, `transcript_path`, `cwd`, `prompt_id`,
`permission_mode`, `effort`, `hook_event_name`, `tool_name`, `tool_input`, `tool_use_id`. A subagent
payload swaps `effort` for `agent_id` + `agent_type`.

### The three properties the design needs, checked separately

Six probe agents, each given commands carrying a unique marker string so payloads could be
attributed by content rather than by timing.

1. **Present on every subagent tool call, not only `Bash`** — 9 of 9, across `Bash`, `Read` and
   `Grep`. The one-tool version of this claim would have been cheap and wrong.
2. **Stable within an agent** — two agents given three separate commands each; all three of one
   carried one id, all three of the other carried the other.
3. **Distinct across concurrent agents, and equal to the parent's own handle** — those two ran
   concurrently with no cross-contamination, and each id matched what the `Agent` tool returned to
   the caller. The dispatcher therefore knows the key *before* the seat's first call, which is what
   lets #290's step 1 write a settings file that step 3 can find.
4. **The join to `SubagentStop` closes on the same key** — `gray-area-capture` records `agent_id` at
   stop; for all six probes it is byte-identical to the one seen at `PreToolUse`.

### The corollary, which matters more than the headline

**`session_id` and `prompt_id` are identical across the main session and every concurrent subagent.**
Neither can namespace a seat. `agent_id` is the only field on the payload that discriminates one seat
from another — the empirical form of #290's *"`agent_id` is a collision-free identifier; if you need a
UUID, it's that."*

`agent_transcript_path` is literally `…/subagents/agent-<agent_id>.jsonl`. The harness derives it from
the same key.

### 7a. `agent_transcript_path`: the correlate is the SEAT, not the session (#189)

§11.8 of `gray-area.md` said "not written in this environment". #189's first correction narrowed that
to "the correlate is the session, not elapsed time". **Both are wrong**, and the third axis was found
the same way as the first two — by checking a property nobody had checked.

All 69 `kind: "seat"` rows of this session's trajectory manifest:

| | `agent_type` populated | `agent_type` empty |
|---|---|---|
| `resolved: true` | **19** | 0 |
| `resolved: false` | 0 | **50** |

**Zero exceptions.** Both populations appear inside one session seconds apart (`00:09:54` resolved,
`00:09:55` unresolved, `00:09:57` resolved), so neither "this environment" nor "this session" is the
unit — the seat is.

Checked against disk rather than trusting the flag: the resolved ids were exactly the `agent-*.jsonl`
files present, **0 resolved-but-missing**. One file was on disk whose own row said unresolved — it
landed after its `stat`. **So the write race is real and accounts for 1 of 16, not for the 50.**

Every row in both populations carries a non-empty `agent_id` (69/69 distinct, 0 duplicates). The
typeless events are not anonymous: they have an identity and no type, pointing at a file that was
never written.

**Operationally: `agent_type` empty ⟹ no transcript, predictable without a `stat`.** Capture can
classify a seat as uncapturable when the row is built instead of reporting a filesystem error that
reads like a transient one. The refusal to add a fallback path search **stands and is reinforced** —
with 50 of 69 seats having no file, a fallback would be guessing three times in four.

**What produced the typeless population is NOT determined.** *(ANSWERED 2026-08-18 —
`plans/gray-area.md` §11.11. They are not a population of seats: `SubagentStop` fires at the MAIN
agent's turn end, with a minted id, no meta sidecar and a PREDICTED transcript path nothing writes
to. 0 of 146 land mid-turn across 3406 windows; a subagent completing is by construction mid-turn.
The "50 of 69" and "guessing three times in four" figures above therefore count turn ends in their
denominator — seat coverage is 19/19.)* All six probes — explicit
`general-purpose`, explicit `claude`, explicit `Explore`, and `subagent_type` omitted — carried a type
and resolved. An attempt to identify the others by looking for their ids in the parent transcript was
**contaminated and discarded**: printing the ids into diagnostics put them into the transcript, so the
counts measured the analysis, not the harness. Recorded rather than quietly dropped, because a
contaminated count that agrees with the hypothesis is how this question got its axis wrong twice.

§5's table carries these claims; it is the index, and this section is the evidence behind the rows
added there. Kept in one place deliberately — a status duplicated in two tables drifts, and the
stale copy is the one that gets read.

---

## 8. Four more events CAN inject — and `Stop` re-arms the turn (measured 2026-08-22, #505)

**Client 2.1.235**, not the 2.1.220 the rest of this record was taken against — so this is also a
re-verification under R9, on a client two patch versions newer.

`SessionStart` (§3a) and `PostToolUseFailure` (§2a) were the only events verified to INJECT.
`checkpoint-freshness.md` needed a mid-session cadence and had none: one fires once, the other only
on failure. Four candidates were untested.

**Method**, following §2a. Scratch project at `/tmp/sc-inject-spike`, headless `claude -p --model
haiku --permission-mode bypassPermissions`, one prompt that forces a tool call
(`Read sample.txt and tell me what it says.`). Five hooks, each logging its raw stdin to a file and
emitting `{"hookSpecificOutput":{"hookEventName":"<E>","additionalContext":"MARKER-<E>-7Q4X"}}`.
`SessionStart` is the **positive control** — its injection is already settled, so a subject coming
back NOT-SEEN could not be blamed on a run where nothing fired.

**Result: all four inject.** Every marker arrived as a `hook_additional_context` attachment.

```
{"e":"SessionStart",   "n":"SessionStart",     "c":["MARKER-SessionStart-7Q4X"]}     <- control
{"e":"UserPromptSubmit","n":"UserPromptSubmit","c":["MARKER-UserPromptSubmit-7Q4X"]}
{"e":"PostToolUse",    "n":"PostToolUse:Read", "c":["MARKER-PostToolUse-7Q4X"]}
{"e":"PostToolBatch",  "n":"PostToolBatch",    "c":["MARKER-PostToolBatch-7Q4X"]}
{"e":"Stop",           "n":"Stop",             "c":["MARKER-Stop-7Q4X"]}  x9
```

`PostToolUse`'s `hookName` carries the matched tool (`PostToolUse:Read`), as `PostToolUseFailure`'s
did. `PostToolBatch` fires once per batch and is real — it had never been exercised here.

### The finding that matters more than the headline: `Stop` injection re-arms the turn

`Stop` fired **nine times on one prompt**. The transcript shows why — this is the trace, not an
inference:

```
USER:      Read sample.txt and tell me what it says.
ASSISTANT: <tool:Read> → "The file contains just the word "hello"…"
  [inject Stop]
ASSISTANT: Ready when you are. What would you like me to do?
  [inject Stop]
ASSISTANT: I'm standing by. What's next?
  [inject Stop]
ASSISTANT: I'm here and ready to help…
  … six more …
ASSISTANT: Ready.
```

The turn ends, the hook injects, the injected context re-invokes the model, it emits a filler line
with nothing to do, `Stop` fires again. **A controlled null run disambiguates it** — same project,
same prompt, same model, a `Stop` hook that consumes stdin and emits *nothing*:

| Run | `Stop` firings | assistant entries | outcome |
|---|---|---|---|
| **subject** (Stop emits `additionalContext`) | **9** | **20** | 8 filler turns, 1,186 wasted output tokens |
| **control** (Stop emits nothing) | **1** | **4** | one answer, clean exit |

So the repetition is caused by the injection, not by `Stop` firing repeatedly on its own. No cap or
loop-abort message appeared on stderr; it stopped at nine for a reason this run did not establish.

`PostToolUse` and `PostToolBatch` injected **once each with no loop** — the pathology is specific to
`Stop`.

### What this means for a design

**The hazard and the feature are the same mechanism.** An unconditional injector on `Stop` is a
token-burning loop. But a *guarded* one — at most one emission, the guard written before the
context is returned — hands the model exactly one extra turn in which to act on what it was told,
which is precisely what a nudge wants and what no other event provides: `PostToolUse` injects into
a turn that is already committed to its next action, `Stop` injects into a turn boundary and then
creates a turn.

Any `Stop` injector MUST therefore record its emission as spent **before** returning the context,
never after. A guard that writes afterwards, or crashes between, re-emits — and re-emitting on
`Stop` is not a duplicate nudge, it is a loop.

---

## 9. Full-catalogue census on 2.1.240 — and three corrections, one of them to §8's own run (measured 2026-08-22)

**Client 2.1.240**, a third version for this record (2.1.220 for §1–§7, 2.1.235 for §8). **Method:**
scratch project `/tmp/sc-spike-B`, all **30 catalogue events** registered to one logging hook that
writes its raw stdin and exits 0; **15 headless `claude -p` sessions** across three passes, `haiku`
except where a scenario needed a stronger model to actually perform the act being measured.

This census was rebuilt from a transcript after the machine it was first running on rebooted mid-run,
which is why §9 exists at all rather than an amendment to §8.

### 9a. What fired

15 of 30. `PreToolUse` / `PostToolUse` / `PostToolUseFailure` / `PermissionRequest` were each
registered **twice** — once with no `matcher` key, once with `matcher: "*"`.

| Event | Firings | Exercised by |
|---|---|---|
| `SessionStart` `UserPromptSubmit` `Stop` `SessionEnd` | 15 / 15 / 15 / 14 | every session |
| `InstructionsLoaded` | 15 | `CLAUDE.md` present — `{file_path, memory_type:"Project", load_reason:"session_start"}` |
| `MessageDisplay` | 18 | every assistant message — `{turn_id, message_id, index, final, delta}` |
| `PreToolUse` | 15 NOMATCHER **and** 15 STAR | every tool call |
| `PostToolUse` | 12 NOMATCHER **and** 12 STAR | every successful tool call |
| `PostToolBatch` | 17 | once per batch, carrying `tool_calls[]` and `permission_mode` |
| `PostToolUseFailure` | 1 NOMATCHER **and** 1 STAR | `ls /definitely/not/here/xyz` |
| `FileChanged` | 1 | a `SessionStart`-registered `watchPaths` dir — `{file_path, event:"add"}` |
| `CwdChanged` | 1 | `cd sub && pwd` — `{old_cwd, new_cwd}` |
| `ConfigChange` | 1 | writing `.claude/settings.local.json` — `{source:"local_settings", file_path}` |
| `SubagentStart` / `SubagentStop` | 1 / 1 | one `general-purpose` seat |

### 9b. The negatives, split by whether the scenario actually happened

**The split is the point.** An event that did not fire because nothing exercised it is not evidence
about the event; recording those two rows the same way is how a census produces a plausible zero.

| Event | Verdict | Basis |
|---|---|---|
| `TaskCreated` `TaskCompleted` | **REFUTED for both shapes tested** | A background shell task was created and *proven live* — id `baac0h503` appears in the `Stop` payload (§9c) — and a subagent was launched. Neither event fired for either. |
| `StopFailure` | **REFUTED for this shape** | The `Stop` hook exited **1** with stderr. `Stop` fired; `StopFailure` did not. A hook that *blocks* rather than errors is untested. |
| `PermissionRequest` `PermissionDenied` | **REFUTED for the hook-denial shape** | A `PreToolUse` hook returned `permissionDecision:"deny"`; the call was denied and no `PostToolUse` followed, so the denial demonstrably happened. Neither event fired. An interactive user-facing denial is a different shape and remains untested. |
| `Elicitation` `ElicitationResult` | **NOT MEASURABLE HERE** | `AskUserQuestion` is not available under `claude -p`; two models independently reported the tool absent. Needs an interactive harness. |
| `PreCompact` `PostCompact` | **not exercised** | No compaction forced in this run; measured in §3/§4. |
| `Setup` `TeammateIdle` `WorktreeCreate` `WorktreeRemove` `DirectoryAdded` `UserPromptExpansion` | **not exercised** | No scenario in this harness reaches them. Enumerated so the omission is stated. |

### 9c. `Stop` carries the in-flight handles as a field

The `Stop` payload includes `background_tasks[]` and `session_crons[]`. With a live task:

```json
{"stop_hook_active":false,"last_assistant_message":"launched",
 "background_tasks":[{"id":"baac0h503","type":"shell","status":"running",
                      "description":"sleep 120; echo late","command":"sleep 120; echo late"}],
 "session_crons":[]}
```

`TaskCreated`/`TaskCompleted` were wanted so a hook could know when a note's "In-flight handles"
section had gone false. **The events do not fire and are not needed**: the handles are a structured
field on the channel §8 already chose, with `id`, `type`, `status`, `description` and `command`.
**This discharges the condition §2 left open.** That section spotted `background_tasks` on 2.1.220 and
refused to build on it: *"an empty array from a run with no background work proves only that the key
exists."* The array above is from a run with a task still running, so the key is now known to carry
content and not merely to be present.

This is the `facts-are-fields` shape arriving for free — a record with fields, not a string to
recover — and it removes the reason #506 was held out of the plan.

### 9d. `SubagentStop` fires at the parent, and carries both transcripts

```json
{"session_id":"90e68dee…","transcript_path":"…/90e68dee….jsonl",
 "agent_id":"aaaacc3a224927897","agent_type":"general-purpose",
 "agent_transcript_path":"…/90e68dee…/subagents/agent-aaaacc3a224927897.jsonl",
 "last_assistant_message":"hello","effort":{"level":"high"},
 "stop_hook_active":false,"background_tasks":[],"session_crons":[]}
```

`session_id` and `transcript_path` are the **parent's**; the seat's own file is a separate field. A
hook that wants to gauge the *parent's* note at the moment a seat returns has both, distinguished by
field rather than by inferring which path is which — which is what #507 asked for and assumed was
unavailable.

### 9e. Three corrections

**1. Tool events DO fire without a `matcher` key.** 15/15 `PreToolUse` and 12/12 `PostToolUse`
firings came in matched pairs, NOMATCHER alongside STAR, never one without the other. The
interrupted run that preceded this one inferred the opposite from a pass in which its driver had
died of double-backgrounding: nothing ran, nothing fired, and the missing `matcher` took the blame
for an absence that had a different cause. Recorded because the inference was one edit away from
becoming a registration rule in shipped hooks — [[facts-are-fields]] clause 2, *find what actually
produced the number*.

**2. `SubagentStop` did not fire at main-agent turn ends.** §7a's parenthetical (ANSWERED
2026-08-18) says it fires at the MAIN agent's turn end with a minted id and a predicted transcript
path nothing writes to. **Across 15 sessions here it fired exactly once — for the one real seat**,
carrying `agent_type`, a resolvable `agent_transcript_path`, and the seat's own
`last_assistant_message`. Scope of this contradiction, stated rather than generalised: 2.1.240,
headless `claude -p`, single-prompt sessions. It does not retract §7a's population analysis on
2.1.220; it does mean **any design keying off "SubagentStop implies a real seat" must re-measure on
its own client**, because the two clients disagree. R9 again, and the third time this record has
moved under a version bump.

**3. The context window is not in any payload.** `SessionStart`'s stdin is
`{session_id, cwd, hook_event_name, source}` — **no model field at all**, so the question of whether
it distinguishes `claude-opus-5[1m]` from `claude-opus-5` does not arise. Independently, on a live
`[1m]` session, `message.model` in the transcript reads `claude-opus-5` and **no `*limit*` field
exists anywhere in the file**. `compactMetadata.preTokens` after a session's first compaction stays
the only source of a denominator. This closes #509 negatively rather than leaving it open.

---

## 10. `SubagentStop` loops without delivering; `SubagentStart` injects into the seat (measured 2026-08-22, #507)

**Client 2.1.240.** §9d established that `SubagentStop` reaches the parent and names both transcripts.
That made it a candidate nudge channel for #507, on one unmeasured assumption: that it can *inject*.
**It cannot** — and the way it fails is worse than a plain refusal.

**Method** as §8: scratch project, one hook per event emitting
`{"hookSpecificOutput":{"additionalContext":"MARKER-<E>-9Z2K"}}`, `SessionStart` as the positive
control, one `general-purpose` seat launched by `sonnet` under `claude -p`.

| Event | Firings | Marker delivered | Where |
|---|---|---|---|
| `SessionStart` (control) | 1 | **yes** | parent `hook_additional_context` |
| `SubagentStart` | 1 | **yes** | **the SEAT's** context, not the parent's |
| `SubagentStop` | **9** | **no** | nowhere — absent from parent and seat transcripts |

### `SubagentStop` re-arms the seat and discards what it said

All nine firings carry the **same** `agent_id`, the same `agent_type`, and the same
`agent_transcript_path` — one seat, nine turn boundaries — with `stop_hook_active` **false on the
first and true on the remaining eight**. The seat's own transcript holds **9 assistant entries** for
a task whose answer was the word `hello`; the parent's holds 5, unchanged.

So the emission re-invokes the **seat**, its turn ends, the hook fires again, nine times, and the
context it returned is thrown away at every step. `Stop` (§8) at least pays for its loop with a
delivered message. **This is the loop with the payload removed: cost identical, benefit zero.**

Under a log-only hook the same launch fires `SubagentStop` exactly **once** (§9a) — the loop is
caused by the emission, as it was for `Stop`, and is not a property of the event.

### What this settles

- **`SubagentStop` is not a nudge channel.** Anything wanting to *tell the model* something when a
  seat returns cannot use it. It remains sound for **observation** — sealing, stamping, recording —
  which needs no injection, and §9d's parent-identity finding stands unchanged.
- **`SubagentStart` is a per-seat injection channel, VERIFIED** — §5's "injection not exercised" row
  is now measured. It reaches the seat only, which is the correct shape for handing a subagent its
  own context and the wrong shape for anything about the parent's note.
- **`stop_hook_active` is present and correct on `SubagentStop`**, so a guard could break the loop —
  but there is nothing to guard, because nothing arrives.

**Keep the loop in proportion: it is cycle detection, and it is cheap.** Nine firings is the
*unguarded* case, which no design here proposes — `checkpoint-freshness.md` F10 already requires
write-before-emit. Two independent brakes exist, and they do different jobs:
`stop_hook_active` is a field on the payload from the authority itself (false on the first firing,
true on all eight re-entries), needing no state and no write; a debounce file is for **band
policy** — at most one emission per band per session, which must survive across turns — and its
worst failure is one duplicate emission. Do not read §8 or §10 as an argument that these channels
are dangerous. The measured cost of an *unguarded* injector is the argument for the guard, nothing
more.

**What the guard does not fix.** A perfect brake turns `SubagentStop`'s cost from nine seat turns
into one, and the marker is still absent from both transcripts. #507 is refuted on **delivery**;
the loop was never the reason.
- **The general rule, third time it has held:** injection re-arms the turn of whichever context the
  event belongs to. `Stop` → the session's turn. `SubagentStop` → the seat's turn. The delivery
  question and the re-arm question are **independent**, and this is the first measured case where a
  channel re-arms and delivers nothing. Any future "can X inject?" must ask both.

---

## 11. The compaction lever changed: `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` is inert on 2.1.240 (measured 2026-08-22)

Every compaction measurement in this record — §3, §3a, §4, and the header's own method paragraph —
was taken by forcing auto-compaction with `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`. **On 2.1.240 that
variable does nothing.** `checkpoint-freshness.md` §V drives Phase 2's live acceptance test with it,
which is why this was reproduced rather than trusted; §V says so in as many words.

### The sweep: four values, one workload, zero compactions

Scratch project, six 39 KB prose files read in order by `claude -p --model haiku`, one run per value,
each in its own project directory. `PreCompact`, `PostCompact`, `SessionStart` and `Stop` logged.

| `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` | `PreCompact` | `PostCompact` | `compact_boundary` in transcript | peak context |
|---|---|---|---|---|
| 2 | 0 | 0 | **0** | 102,671 |
| 5 | 0 | 0 | **0** | 102,456 |
| 10 | 0 | 0 | **0** | 102,609 |
| 25 | 0 | 0 | **0** | 102,707 |

**The last column is the proof, and it is why the hook counts alone would not have been enough.**
Zero firings is also what a run that never grew would produce. Every run reached ~102k tokens of
context: at `PCT=2` the trigger should sit near 4k on a 200k window, and near 20k even if the window
were 1M. Nothing fired at any value. This is not a mis-sized value — the lever is disconnected.

### `--autocompact` is the replacement, and it works

`claude --help` on 2.1.240 documents `--autocompact <auto|tokens>` — *"Auto-compact window size
(auto, or 100k–1M tokens)"* — a flag that did not exist when §4 was written. Same harness, fifteen
files, `--autocompact 100k`:

| | count |
|---|---|
| `compact_boundary` in transcript | **4** |
| `PreCompact` / `PostCompact` | **4 / 4** |
| `SessionStart` | **5** — one `startup`, **four `source: "compact"`** |

```
preTokens 72,747 → postTokens 13,283   (dropped  59,464)
preTokens 71,839 → postTokens 14,136   (dropped 117,167)
preTokens 74,853 → postTokens 14,856   (dropped 177,164)
preTokens 76,361 → postTokens 16,116   (dropped 237,409)
```

`PostCompact` carried a 7,150-character summary, inside §3's observed 692–11,882 band. **§3a's
post-compaction injection channel still fires on this client** — the `source: "compact"` SessionStart
is present at every boundary.

### Two properties worth carrying into any harness built on this

1. **The flag's number is not the trigger point.** A stated `100k` fired at 71.8k–76.4k, four times,
   tightly clustered — roughly 72–76% of the stated size. The wording *"window size"* suggests the
   flag sets a window and compaction fires at a fraction of it; that reading is a **hypothesis, not
   measured**, and no run here varied the flag to test it.
2. **Boundary count is a property of the workload, not the value.** Six files gave one boundary;
   fifteen gave four. Any test needing *exactly one* boundary must size the workload against the
   trigger and assert the count from the transcript — §V's "one clean boundary" is a recipe to be
   tuned, never a constant to be quoted.

### The method rule this earns

`context-checkpointing.md` §17 A already says this claim was *"wrong in the same way twice: the value
was sized off a guess and then 'confirmed' by a clean exit. A clean exit says the run finished. It
says nothing about where the threshold was."* **This is the third time, and the failure mode moved:**
it is no longer the value, it is the lever. A control surface can be removed between patch versions
while every test built on it keeps exiting 0 — the run finishes, the hook does not fire, and nothing
distinguishes "the compaction was handled" from "no compaction ever happened."

So: **a test that forces compaction MUST assert the boundary from the transcript**
(`select(.subtype=="compact_boundary")`), never infer it from a clean exit and never from hook
silence. Both are what a disconnected lever looks like.

---

## 12. Which events carry `background_tasks` — the key sets, measured (2026-08-23, client 2.1.240)

§2 spotted `background_tasks`/`session_crons` on `SubagentStop` and §9c confirmed `Stop` populates
them on a live task. **Neither says anything about the other sealing events**, and a design that
stamps in-flight handles at a seal needs to know per event, not in general. `checkpoint-freshness.md`
briefly cited §9c/§11 for a `PreCompact` result those sections do not contain; this section is that
measurement, taken rather than inferred.

**Method.** One `claude -p` run that launches a background shell task and then a subagent, so a seat
returns while a task is live; plus a clean run for the events the first cannot reach. Raw hook stdin,
keys read directly.

| Event | Seals? | `background_tasks` | Full key set |
|---|---|---|---|
| `PreCompact` | **yes** | **ABSENT** (4/4 firings) | `custom_instructions, cwd, hook_event_name, prompt_id, session_id, transcript_path, trigger` |
| `SessionEnd` | **yes** | **ABSENT** | `cwd, hook_event_name, prompt_id, reason, session_id, transcript_path` |
| `SubagentStop` | **yes** | **PRESENT, populated** | + `agent_id, agent_type, agent_transcript_path, background_tasks, effort, last_assistant_message, permission_mode, session_crons, stop_hook_active` |
| `Stop` | no | **PRESENT, populated** | as `SubagentStop`, minus the `agent_*` fields |

**Two of the three sealing events cannot measure handles at all.** Only `SubagentStop` can — so a
`live_handles` column stamped at a `PreCompact` or `SessionEnd` seal is not a measurement, and a `0`
written there is manufactured. The absent key and the honest empty list must be kept distinct by the
decoder, not by the field's zero value.

### `background_tasks` includes subagents, not just shells

With a shell task and a seat both in flight, `SubagentStop` carried **two** entries:

```json
[{"id":"bqpl0vjaw","type":"shell","status":"running",
  "description":"sleep 120; echo late","command":"sleep 120; echo late"},
 {"id":"ae91554853e3a83aa","type":"subagent","status":"running",
  "description":"Reply with hello","agent_type":"general-purpose"}]
```

`type` discriminates, and a seat appears in the parent's own handle list while it runs. Anything
counting "in-flight work" gets subagents for free and must decide whether it wants them — a count
that silently includes the seat whose return triggered the seal will read high by exactly one.

### `SessionEnd` did not fire in either run that ended with live background work

**0 of 2** runs ending with a `sleep 120` still running produced a `SessionEnd`; it fired in **14 of
15** census sessions, all of which ended clean, and in the clean control here. Two observations are a
pattern, not a finding, and the confound is named: both live-task runs also hit the driver's pipe
close, so "the session did not end the way the hook counts as ending" is not excluded. **Recorded
because the case it touches is the one a checkpoint design cares most about** — a session going away
with work still running is exactly when the sealed note matters — and because a seal that silently
never fires is indistinguishable from a seal that fired and found nothing.

---

## 13. `stop_hook_active` measured on `Stop`, and the loop is 3.6× worse on 2.1.240 (2026-08-23)

§8 measured the `Stop` injection loop on **2.1.235** and recorded `stop_hook_active`'s presence from
`SubagentStop`'s payload table. Two things were therefore never verified on the channel the design
actually uses: whether the flag flips on a `Stop` **re-entry**, and whether §8's numbers still hold.
Both now measured on **2.1.240**, same method as §8 — unguarded `Stop` hook emitting
`additionalContext`, `SessionStart` positive control, one prompt forcing a tool call.

### The brake is real

| Firing | `stop_hook_active` |
|---|---|
| 1st | **`false`** |
| 2nd–16th | **`true`** — all fifteen |

The client marks every re-entry. A guard reading this flag can refuse to emit inside a loop it
caused, **on the channel that matters**, with no state of its own and nothing to write. That closes
the second of F10's two brakes in `checkpoint-freshness.md`, which until now cited a `SubagentStop`
table for a `Stop` property.

### The loop got worse, and F9 gains its fourth instance

| | 2.1.235 (§8) | **2.1.240 (here)** |
|---|---|---|
| `Stop` firings | 9 | **16** |
| assistant entries | 20 | **35** |
| output tokens burned | 1,186 | **4,326** |

Same shape, **3.6× the cost**. Whatever caps the run at 9 or 16 is not a documented constant and
moved between two patch versions; neither run produced a cap message on stderr. **An unguarded
injector is more expensive on the current client than this record's headline figure**, so §8's
numbers must be read as that client's, not as the hazard's size.

Injection itself re-verified here: **16 `hook_additional_context` attachments**, one per firing,
alongside the `SessionStart` control — `Stop` injection was measured on 2.1.235 and had not been
re-run on 2.1.240 until now.

**The design consequence is unchanged and better founded.** Guard on `stop_hook_active` first — it is
the authority's own flag, needs no file, and cannot be lost — with the debounce record for band
policy, which is a different question from cycle detection.
