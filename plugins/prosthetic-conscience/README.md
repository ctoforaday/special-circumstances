# prosthetic-conscience

> *The drone that keeps you honest.*

The base plugin of [Special Circumstances](../../README.md). It carries the shared rule substrate the other plugins preload, plus the adversarial-partner behaviour for interactive work.

**Status: shipping.** 27 skills (5 of them slash commands), 2 agents, 10 hook binaries.

## The distinctive idea

The rules are **defense in depth, not duplication**. A skill states the *semantic* rule — what a regex cannot catch — and a hook enforces the *pattern-matchable* part in the places where prompts never fire: auto mode, headless `claude -p`, scheduled runs. Neither half is sufficient. A rule with no hook is advice; a hook with no rule is a regex nobody can argue with.

## Rules

Eleven load on every session, the rest by description. `design-by-contract` is the authoring grammar: BEFORE / During / AFTER · YOU MUST.

`agent-guardrails` · `anti-spinning` · `complete-the-concept` · `context-checkpointing` · `context-efficiency` · `critical-stance` · `design-by-contract` · `facts-are-fields` · `git-proficiency` · `markdown-proficiency` · `pair-programming` · `plan-act-reflect` · `project-memory` · `qlty-proficiency` · `refactoring-safety` · `scratch-policy` · `semantic-consent` · `spec-driven-development` · `terse-communication` · `test-driven-development` · `think-around-problem` · `validation-loop`

## Hooks

**One binary per event**, each composing the units that belong to it, so one process parses the payload once and emits one answer.

| Binary | Event | Units it runs |
|---|---|---|
| `sc-pretooluse` | PreToolUse | Secrets gate (fails closed, can deny) · push-freeze guard (warns, never blocks) |
| `sc-posttooluse` | PostToolUse | The quality gate (`qlty fmt` + `qlty check`), over one shared context |
| `sc-strike-counter` | PostToolUseFailure | Counts (tool, target) repeats for anti-spinning; skips interrupts |
| `sc-sessionstart` | SessionStart | Toolchain nudge · checkpoint restore — every source, compaction included, where it also says to resume from the note |
| `sc-precompact` | PreCompact | Seals the note, and tells the summarizer what to preserve |
| `sc-sessionend` | SessionEnd | Seals the note on every reason, headless `other` included |
| `sc-subagentstop` | SubagentStop | Seals a seat's note, keyed by `agent_id` |
| `sc-postcompact-observe` | PostCompact | Scores what each summary kept; observation only |
| `sc-filechanged-rearm` | FileChanged | Marks a check stale when its trigger surface moves |
| `sc-stop` | Stop | The checkpoint-freshness nudge — once per band, how stale the note is; with no note, how heavy the context is |

`sc-doctor` is the eleventh binary and the one you invoke yourself, via `/prosthetic-conscience:doctor`.

Every hook is wrapped in a bootstrap guard: a fresh plugin version ships from git *without* binaries, and an unguarded hook crash-storms every tool call in that window. The guard hands a missing binary to `hooks/fetch-bin.sh`, which installs the plugin's binaries from its own release in the background and tells you it is doing so. If that fails it says why, and `/prosthetic-conscience:doctor --fix` installs them by hand.

### If you launch sessions whose reply is parsed — `SC_FINAL_MESSAGE_CONTRACTED`

A headless session is still a *main* session, so `Stop` fires in it and the freshness nudge asks the agent to write its note and talk to the human. Where the agent's final message is an envelope your program parses, that request is an interruption with nowhere to go: the agent answers it in prose and your envelope is lost.

**Set `SC_FINAL_MESSAGE_CONTRACTED` in the session's environment** and `sc-stop` declines — exit 0, nothing said on either channel, no state written. Any non-empty value asserts it; leave it unset to mean no. Nothing in the hook payload distinguishes such a session from one a person is reading, so this is a signal you send rather than a condition the plugin can detect: a launcher that does not set it gets the nudge. `hooks/README.md` carries the measurement behind it.

Every other session keeps the nudge, agent included — the agent is the one who has to write the note.

## Compaction survival — the Memento problem

Compaction replaces the transcript with a summary. The summary is good at what happened and worst at **what you were about to do**: the exact validation commands, what re-arms each one, the ordered next actions, and the handles to work still running in the background.

So the agent maintains one `CHECKPOINT.md`, overwritten in place. `SessionStart` hands it back on the far side of any seam. `/prosthetic-conscience:checkpoint` writes it, `/prosthetic-conscience:resume` prints it in full and re-verifies its claims.

**Compaction is not the only seam**, so the seal fires on three events. `PreCompact` snapshots the note and asks the summarizer to preserve what it already carries. `SessionEnd` catches a session that ends without ever compacting — on *every* reason, because a headless `claude -p` run reports `other` rather than one of the interactive reasons, and those are exactly the sessions with no human watching. `SubagentStop` seals a seat's note keyed by `agent_id`, since every subagent shares its parent's `session_id` and a name without it has concurrent seats overwriting each other.

Only `PreCompact` speaks. `SubagentStop` stdout reaches a seat still working, and an instruction there is a directive that seat never established — the same rule the restore hook ships a test for.

**And the note's own claims decay.** A compacted agent reading `last run: pass` off its checkpoint will believe a check is green; if the check's inputs moved after that run, the note is stale in the most dangerous way — it *looks* current. So `SessionStart` registers each check's trigger surface with the file watcher, and `FileChanged` marks the check stale when something under it changes, naming the file that did it. The next digest says so.

Measured limits, because they shape the design: `watchPaths` takes **paths — files or directories — and no pattern of any kind**, not globs and not regex. A pattern registers nothing, silently. So a surface written `manifest/*.yml` is watched as `manifest/`, directories are recursive and do fire for files created later, and a surface that is prose rather than a path (*"a human deciding to ship"*) is **reported as unresolvable** instead of quietly watching nothing.

Two constraints came from measurement rather than design, and both cost a cycle to learn:

**Restore is `SessionStart`, on every source including `compact`.** An earlier revision routed the compaction boundary through `PostCompact`, since that event receives `compact_summary` and could in principle inject only the delta. It cannot inject at all — absent from the client's `hookSpecificOutput` union, stdout routed to the user rather than the model, and `NOT-SEEN` in an end-to-end marker test. It also runs *after* `SessionStart`, so the digest is already written by the time the summary exists.

**The resumed agent treats the digest as a claim, and that is correct.** The digest names the file it came from, when it was written, and which session wrote it. That was designed to stop the agent reading it as a prompt-injection attempt — a reaction observed twice on other events. **It does not stop it.** In the acceptance run the agent recovered every value exactly, attributed them honestly to the hook, and *still* flagged the payload: *"untrusted file content … formatted to read as authoritative state … embeds an imperative instruction."* The imperative it named was the note's own foot-gun entry, and a section whose job is to carry foot-guns carries imperatives by definition.

What that measurement did change is narrower: **the digest itself states facts.** The first run's digest closed with *"verify each item before acting on it"*, and the agent cited that sentence as one of the directives making it injection-shaped. Removing it removed it from the reason. The distrust itself is the posture the skill asks for; the agent reached it unprompted.

**After a compaction, and only then, the hook adds one instruction.** The turn continues anyway, so the only question is whether it continues from the summary or from the note — and the skill that says "from the note" loads by description, so a consumer session often lacks it. The line tells the agent to read the full note, check its head, handles and queue pointers, redo nothing the summary reports as done, and take the first next action; when the note's status is `blocked`, to tell the human what it waits on and stop. On startup and resume the note may be stale or another session's, so there the digest stays a claim with no instruction. Measured on a manual `/compact` followed by *"Continue."*, one run per cell:

| Note | haiku, old | haiku, with the line | sonnet, old | sonnet, with the line |
|---|---|---|---|---|
| in progress | finished the work | finished the work | finished the work | read the full note, checked it, then finished |
| blocked on a human decision | **made the decision and edited** | stopped | stopped and asked | stopped and asked |

None of the eight runs called the digest or the line an injection.

An explicit `/clear` gets a pointer rather than the digest. That carve-out is by intent — the human just wiped the context deliberately — which is precisely what the withdrawn `compact` carve-out was not.

## Slash commands

`/prosthetic-conscience:checkpoint` · `/prosthetic-conscience:resume` · `/prosthetic-conscience:doctor` · `/prosthetic-conscience:plan-audit` · `/prosthetic-conscience:probe`

## Honest limits

- **Hook events are version-unstable.** Several events this plugin uses postdate its own design. The load-bearing path is deliberately built on `SessionStart`, the oldest of them: an older client loses observability and the seal's instruction fold-in, never continuity. `/prosthetic-conscience:doctor` reports what the installed client supports.
- **The note is a claim, not a fact.** Restore hands back what the session *wrote down*, which may already be stale. `/prosthetic-conscience:resume` re-verifies before acting; the digest says so in its own text.
- **The quality gate depends on qlty being installed.** Absent, the `sc-posttooluse` unit degrades to silence rather than to a false pass.

Design: [`plans/context-checkpointing.md`](../../plans/context-checkpointing.md) · [`plans/claude-port-plan.md`](../../plans/claude-port-plan.md) §3a. Measured hook payloads: [`plans/hook-surface-spike.md`](../../plans/hook-surface-spike.md).
