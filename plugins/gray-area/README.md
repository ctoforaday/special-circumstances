# gray-area

> *The GCU shunned for reading minds.*

Trajectory evidence for [Special Circumstances](../../README.md): **what a session actually did, as against what it reported.**

Named for the GCU *Gray Area* (*Excession*) — the Culture's mind-reader, the ship that establishes what happened by reading directly, and is shunned by other Minds for doing it. The capability is necessary and distasteful at once, and the name is meant to keep that uncomfortable.

**Status: Phase 5 — capture, a reader, and one adjudication.** The plugin records where each seat's trajectory is; `gray-area tools` lists what a seat actually invoked; and `gray-area checkpoint` puts a sealed checkpoint's validation-loop claims against what the session actually ran. It still concludes nothing on its own — every row cites both documents so a reader can check it.

## What it does today

Every subagent writes its own transcript. When a seat finishes, `SubagentStop` hands over that file **by path** — along with the seat's `agent_id` and `agent_type`. The `gray-area-capture` hook records one manifest row per seat:

```
.claude/gray-area/trajectories-<session-id>.jsonl
```

Keyed by session, because every subagent shares its parent's `session_id` — which makes it exactly the right grouping key: one manifest per run, one row per seat.

That replaces the thing every existing consumer of this data does by hand: sweeping `~/.claude/projects/` and guessing which file belongs to which seat. The manifest is *handed* the path.

## Reading a trajectory

```
gray-area tools <transcript.jsonl> [-binary <name>] [-json]
```

Lists every tool invocation as `file:line uuid seat tool target` — so a reader quotes the line, not the tool's opinion of it. An event that cannot be cited is **never emitted**; it is counted and reported as suppressed. If any line failed to parse, the tool warns that the listing is over a *subset*, because an answer over part of the record that reads like an answer over all of it is the failure this exists to prevent.

`-binary <name>` resolves shell-aliased invocations. A seat that runs `REC=./feov-record ; "$REC" finding` is invisible to a matcher that greps for the binary followed by a verb — a measured ~11% of real invocations. The resolver tracks single-level variable assignments so the verb is attributed to the right binary. It is not a shell: command substitution, arrays and indirect expansion are out of scope, and it returns nothing rather than guessing, because a wrong attribution is worse than a missed one when the point is checking a claim against what was actually run.

## What the manifest is, and is not

**It is an index.** Each row records where a trajectory is, and what was true of it at capture time — resolved or not, size, the seat's identity, any background task handles still in flight, the reasoning effort level.

**It is not a copy.** `SubagentStop` also carries `last_assistant_message`, and that is deliberately not recorded. The transcript already holds the content, and duplicating it into a second file buys no evidentiary value while spreading conversation text somewhere new.

**An unresolvable path is data, not an error.** A seat whose trajectory could not be found is recorded as `resolved: false` with the reason. A miner must be able to see that a trajectory was missing, rather than see nothing and assume the seat never ran.

## The line this plugin will not cross

> **Exploration may summarize. Adjudication must cite.**

An agent asked to read a transcript and report what it sees is a summarizer — cheap, useful for finding where to look, and non-deterministic, unreproducible and uncitable. Fine for a hypothesis. Disqualifying for a finding.

This suite spent a full cycle removing self-report from the evidence chain; *"an agent says the transcript shows"* would put it straight back, one layer up and harder to catch. So any query backing a finding must return the primary evidence — the event id, the line, the tool call — and cite the trajectory rather than the index's opinion of it.

## What it deliberately does not do

**Compaction survival.** The "Memento" problem — leaving yourself a note before the context window fills — ships in [prosthetic-conscience](../prosthetic-conscience), not here. Reading transcripts is a surveillance capability; keeping a note about your own work is not, and a consumer who wants the second should not have to accept the first to get it.

## Honest limits

- **The transcript format is vendor-internal and version-unstable.** The parser is version-pinned and degrades rather than guessing.
- **The JSONL is append-only but not signed.** Gray Area establishes what the record says, not that the record is authentic.
- **Reading trajectories is a surveillance capability.** Transcripts carry user text, file paths, and whatever a tool result happened to contain. Inspections are declared and scoped; the manifest is gitignored; nothing leaves the box.

Design: [`plans/gray-area.md`](../../plans/gray-area.md). Measured hook payloads: [`plans/hook-surface-spike.md`](../../plans/hook-surface-spike.md).

## Adjudicating a checkpoint

A sealed checkpoint note's validation loop is a set of **declared claims**: a command was run, it
passed, and it was last run at a stated time. All three are self-report. The session transcript is
the only independent record of what was actually invoked, so the two can be put side by side:

```
gray-area checkpoint .claude/checkpoints/CHECKPOINT.md ~/.claude/projects/<project>/<session>.jsonl
```

Four verdicts, and the negative one is deliberately weak:

| | |
|---|---|
| `CITED` | a matching invocation is in the trajectory — uuid, file, line, timestamp |
| `STALE` | a write under the check's own trigger surface happened **after** the claimed run time |
| `NO-EVIDENCE` | nothing matched. **Not** "did not run" — it prints the tokens searched and how many events were searched, so a reader can see whether the note or the matcher is at fault |
| `UNCHECKABLE` | the claim names no command, so there was nothing to look for |

Exit is non-zero when anything is `STALE` or `NO-EVIDENCE`, so it composes into a gate.

`STALE` is the row this exists for. On 2026-07-30 a note asserted a pass for a check whose
re-arm mechanism was dead ([#165](https://github.com/ctoforaday/special-circumstances/issues/165)),
and went on presenting that pass as current for the rest of the session; it took a hand audit to
notice. This computes the same thing from the transcript alone — **it does not depend on the
mechanism that failed.**

### What it will not do

It does not decide whether a check *should* have run, score a note's trustworthiness, or read
thinking content. A `NO-EVIDENCE` row is an absence, stated as one. The transcript is append-only,
not signed: this establishes what the record says, never that the record is authentic.

## The command

`/gray-area:audit-checkpoint` runs the adjudication above and relays the rows. With no argument it
resolves this session's trajectory from gray-area's own manifest and prints which row it used.

That resolution is why the plugin registers a `SessionStart` hook: `SubagentStop` only ever hands
over a *seat's* transcript, so without a session row there is no non-guessing way to find the
session's own. Searching `~/.claude/projects/` for a likely-looking file is the attribution failure
this plugin is built to remove — a guessed transcript produces confident findings about the wrong
session — so the tool refuses and says what is missing instead.

The command is a weaker mechanism than a hook, on purpose. Wiring the adjudication into
`prosthetic-conscience`'s seal would make continuity depend on the miner, and a consumer has to be
able to take compaction survival without taking a surveillance capability.
