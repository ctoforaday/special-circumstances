# gray-area

> *The GCU shunned for reading minds.*

Trajectory evidence for [Special Circumstances](../../README.md): **what a session actually did, as against what it reported.**

Named for the GCU *Gray Area* (*Excession*) — the Culture's mind-reader, the ship that establishes what happened by reading directly, and is shunned by other Minds for doing it. The capability is necessary and distasteful at once, and the name is meant to keep that uncomfortable.

**Status: Phase 2 — capture, plus a reader.** The plugin records where each seat's trajectory is, and `gray-area tools` lists what a seat actually invoked, with provenance on every row. Nothing yet judges what it finds.

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
