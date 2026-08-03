# prosthetic-conscience

> *The drone that keeps you honest.*

The base plugin of [Special Circumstances](../../README.md). It carries the shared rule substrate the other plugins preload, plus the adversarial-partner behaviour for interactive work.

**Status: shipping.** 20 skills, 5 commands, 2 agents, 8 hook binaries.

## The distinctive idea

The rules are **defense in depth, not duplication**. A skill states the *semantic* rule — what a regex cannot catch — and a hook enforces the *pattern-matchable* part in the places where prompts never fire: auto mode, headless `claude -p`, scheduled runs. Neither half is sufficient. A rule with no hook is advice; a hook with no rule is a regex nobody can argue with.

## Rules

Nine load on every session, the rest by description. `design-by-contract` is the authoring grammar: BEFORE / During / AFTER · YOU MUST.

`agent-guardrails` · `anti-spinning` · `complete-the-concept` · `context-checkpointing` · `context-efficiency` · `critical-stance` · `design-by-contract` · `git-proficiency` · `markdown-proficiency` · `pair-programming` · `plan-act-reflect` · `project-memory` · `qlty-proficiency` · `refactoring-safety` · `scratch-policy` · `semantic-consent` · `spec-driven-development` · `terse-communication` · `test-driven-development` · `think-around-problem` · `validation-loop`

## Hooks

| Binary | Event | What it does |
|---|---|---|
| `sc-secrets-gate` | PreToolUse | Blocks secrets leaving via web calls and shell |
| `sc-push-freeze-guard` | PreToolUse | Refuses pushes to pinned paths while a research run is live |
| `sc-posttooluse` | PostToolUse | One hook per event: the quality gate (`qlty fmt` + `qlty check`) and the recall index, as units over one shared context |
| `sc-toolchain-nudge` | SessionStart | One line when a recommended tool is missing, silence when healthy |
| `sc-checkpoint-seal` | PreCompact · SessionEnd · SubagentStop | Seals the note at every seam; tells the summarizer what to preserve, on PreCompact only |
| `sc-checkpoint-restore` | SessionStart | Hands the note back — every source, compaction included |
| `sc-postcompact-observe` | PostCompact | Scores what each summary kept; observation only |
| `sc-filechanged-rearm` | FileChanged | Marks a check stale when its trigger surface moves |

Every hook is wrapped in a bootstrap guard: a fresh plugin version ships from git *without* binaries, and an unguarded hook crash-storms every tool call in that window. The guard degrades to one stderr line pointing at `/prosthetic-conscience:doctor --fix`.

## Compaction survival — the Memento problem

Compaction replaces the transcript with a summary. The summary is good at what happened and worst at **what you were about to do**: the exact validation commands, what re-arms each one, the ordered next actions, and the handles to work still running in the background.

So the agent maintains one `CHECKPOINT.md`, overwritten in place. `SessionStart` hands it back on the far side of any seam. `/checkpoint` writes it, `/resume` prints it in full and re-verifies its claims.

**Compaction is not the only seam**, so the seal fires on three events. `PreCompact` snapshots the note and asks the summarizer to preserve what it already carries. `SessionEnd` catches a session that ends without ever compacting — on *every* reason, because a headless `claude -p` run reports `other` rather than one of the interactive reasons, and those are exactly the sessions with no human watching. `SubagentStop` seals a seat's note keyed by `agent_id`, since every subagent shares its parent's `session_id` and a name without it has concurrent seats overwriting each other.

Only `PreCompact` speaks. `SubagentStop` stdout reaches a seat still working, and an instruction there is a directive that seat never established — the same rule the restore hook ships a test for.

**And the note's own claims decay.** A compacted agent reading `last run: pass` off its checkpoint will believe a check is green; if the check's inputs moved after that run, the note is stale in the most dangerous way — it *looks* current. So `SessionStart` registers each check's trigger surface with the file watcher, and `FileChanged` marks the check stale when something under it changes, naming the file that did it. The next digest says so.

Measured limits, because they shape the design: `watchPaths` takes **paths — files or directories — and no pattern of any kind**, not globs and not regex. A pattern registers nothing, silently. So a surface written `manifest/*.yml` is watched as `manifest/`, directories are recursive and do fire for files created later, and a surface that is prose rather than a path (*"a human deciding to ship"*) is **reported as unresolvable** instead of quietly watching nothing.

Two constraints came from measurement rather than design, and both cost a cycle to learn:

**Restore is `SessionStart`, on every source including `compact`.** An earlier revision routed the compaction boundary through `PostCompact`, since that event receives `compact_summary` and could in principle inject only the delta. It cannot inject at all — absent from the client's `hookSpecificOutput` union, stdout routed to the user rather than the model, and `NOT-SEEN` in an end-to-end marker test. It also runs *after* `SessionStart`, so the digest is already written by the time the summary exists.

**The resumed agent treats the digest as a claim, and that is correct.** The digest names the file it came from, when it was written, and which session wrote it. That was designed to stop the agent reading it as a prompt-injection attempt — a reaction observed twice on other events. **It does not stop it.** In the acceptance run the agent recovered every value exactly, attributed them honestly to the hook, and *still* flagged the payload: *"untrusted file content … formatted to read as authoritative state … embeds an imperative instruction."* The imperative it named was the note's own foot-gun entry, and a section whose job is to carry foot-guns carries imperatives by definition.

What that measurement did change is narrower and worth keeping: **the hook adds no imperative of its own.** The first run's digest closed with *"verify each item before acting on it"*, and the agent cited that sentence as one of the directives making it injection-shaped. Removing it removed it from the reason. An instruction arriving inside injected text reads as foreign however reasonable it is, and one the hook invented is one the session never established — so the duty to verify lives in the skill, which the session already carries. The distrust itself is the posture the skill asks for; the agent reached it unprompted.

An explicit `/clear` gets a pointer rather than the digest. That carve-out is by intent — the human just wiped the context deliberately — which is precisely what the withdrawn `compact` carve-out was not.

## Commands

`/checkpoint` · `/resume` · `/doctor` · `/plan-audit` · `/probe`

## Honest limits

- **Hook events are version-unstable.** Several events this plugin uses postdate its own design. The load-bearing path is deliberately built on `SessionStart`, the oldest of them: an older client loses observability and the seal's instruction fold-in, never continuity. `/doctor` reports what the installed client supports.
- **The note is a claim, not a fact.** Restore hands back what the session *wrote down*, which may already be stale. `/resume` re-verifies before acting; the digest says so in its own text.
- **`sc-quality-gate` depends on qlty being installed.** Absent, it degrades to silence rather than to a false pass.

Design: [`plans/context-checkpointing.md`](../../plans/context-checkpointing.md) · [`plans/claude-port-plan.md`](../../plans/claude-port-plan.md) §3a. Measured hook payloads: [`plans/hook-surface-spike.md`](../../plans/hook-surface-spike.md).
