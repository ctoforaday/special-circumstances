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

**And *why* it was unresolvable is a field, not prose.** Every unresolved row carries a `capture_category` from a closed set, because two very different conditions used to produce byte-identical rows: an event that named no seat at all, and a seat that should have a transcript and does not. The second is worth an alarm — one transcript in sixteen arrived *after* its stat, so that race is real. `capture_error` still says it in English for a human reading one row; `capture_category` says it in a word a counter can add up.

## Seat coverage: `coverage`

```
gray-area coverage
```

Answers the question any seat-scoped inspection must ask before printing a number: **does the manifest name every seat transcript that exists?** #189 established there are no phantom seats — every `kind: "seat"` row names a file that is there. That is a statement about the *rows*, and it does not bound the rows that were never written.

```
19 distinct seat(s) named by 21 row(s), 20 transcript(s) on disk
2 seat(s) captured more than once — a seat that was continued stops again, and each stop is a true observation:
  REPEAT    agent-a703a4e… — 2 captures
  UNNAMED   agent-a9c4e78… — on disk, and no manifest row accounts for it
```

**Rows are not seats, and the count says which it is.** `SubagentStop` fires again for a seat that is *continued* — measured on this repository's own manifest, one seat captured twice three minutes apart with its transcript grown from 356 KB to 452 KB. Both rows are true, so the manifest keeps both; the reconciliation counts *distinct seats* and reports the repeat rather than erasing it. Counting rows put a per-row number beside a per-file one and called the pair a reconciliation, and a seat captured twice that was also absent from disk was listed under `MISSING` twice.

**A line that does not parse is counted, not skipped.** The manifest is appended to by a hook that can be killed mid-write, and what an interruption leaves is a final line with no newline on it. The writer closes such a tail before appending, so an interruption costs the row it cut and never the next one as well; the cut line stays in the file and `coverage` reports how many there are. Without that count, a seat lost to a torn write shows up as `UNNAMED` — which reads as a hook that never fired, a different fault with a different fix.

**This is not the glob `plans/gray-area.md` §3 refuses.** That rules out sweeping `~/.claude/projects/` *guessing which files belong to which seat* — attribution by guessing, whose failure mode is a false citation. This reads **one** directory, derived from the transcript path the harness itself handed over at SessionStart, and attributes nothing: its output is a set difference over ids, and it cannot emit a citation at all. Auditing the handover is not replacing it.

**It exits 1 when it could not measure and 0 when it did**, whatever it found. Unnamed transcripts are a finding for a human; an unmeasurable board is a broken instrument, and an instrument reporting a clean board when it cannot see is the failure this plugin is about. A missing seat directory in a session that recorded seats is loud; in a session that recorded none it is consistent, and the two are distinguished rather than collapsed.

## Pull request bodies: `pr`

```
gray-area pr <body.md> [transcript.jsonl]
```

A pull request body stops being testimony and becomes a record under inspection: its backticked commands are adjudicated against what the session actually ran, with the same `CITED` / `NO-EVIDENCE` / `UNCHECKABLE` verdicts the checkpoint audit uses.

**The trajectory records what was RUN, never what the run SAID.** So ``` `go run ./check` → 26 passed ``` splits into an act this record can check and an outcome it cannot see at all, and the row says so:

```
CITED       body.md:3  `go run ./check` -> 26 passed, 2 failed
    evidence: …/937047bc.jsonl:15484 389b0d59 2026-08-18T07:52:57Z  cd scripts && go run ./check
    searched: [go run ./check] across 3535 citable events
    NOT MEASURED: the asserted outcome (-> 26 passed, 2 failed) — the trajectory records invocations, not their output
```

Staying silent about the outcome would let `CITED` read as endorsing a number nothing checked. **Exit is 0 even with findings**: a body is read after the fact by a human, and failing on a `NO-EVIDENCE` row would turn "the tokens did not match" into "the pull request is wrong". A body with no backticked command reports that in those words rather than printing an empty list.

## Repetition: `rework` and `stalls`

Two inspections over the trajectory, needing no claim source at all.

```
gray-area rework [transcript.jsonl]   acts done more than once
gray-area stalls [transcript.jsonl]   acts repeated 3+ times BACK-TO-BACK
```

`stalls` counts **consecutive** repetition because [[anti-spinning]]'s 3-strike rule is about a repair loop, not about ordinary iteration — the same check run three times across a session is normal work. It exists alongside the strike-counting hook because that hook counts TOOL failures only, and is documented as blind to a command that exits 0 while the work still failed.

**A row reports repetition, not waste.** Five writes to one file may be five careful increments. Every occurrence is cited with file, line and uuid, and the reader decides.

**The trajectory records what was RUN, never what the run SAID** — result bodies are conversation content, and this plugin does not copy them. So whether a run of identical invocations was retries or deliberate re-runs is printed as **not measured** rather than guessed. That boundary is the answer to the question a miner like this has to get right: a verdict the record cannot support is a false positive wearing the tool's authority.

**Every listing ends with its own coverage** — how many citable tool uses were searched, how many could not be keyed, how many acts happened exactly once. An empty listing with that line is a clean session; without it, an empty listing and an unreadable transcript are the same bytes.

**`kind` says what the row IS, and it is not inferred from which hook delivered it.** `SubagentStop` does not only fire for subagents: measured across 165 such events in one session, 146 carried no `agent_type`, had no file at the path they predicted, and landed at the **main agent's turn end** — 0 of them inside any of 3406 mid-turn windows, where a real subagent completion must land. Those rows are written as `kind: "turn-end"`; a genuine subagent is `kind: "seat"`; the session's own trajectory is `kind: "session"`. Only the conjunction (no type **and** no file) reclassifies, so a typed row whose transcript is missing stays a seat and stays an alarm. **Count seats by the field, not by the hook** — reading every `SubagentStop` as a seat is what made this manifest look 72% blind when every seat in it resolved.

Rows are still stat'ed even when `agent_type` is empty. The correlation is one session's evidence and what makes a seat untyped is undetermined, so the row reports what it observed rather than what was predicted — a row that cannot contradict the expectation has stopped being a measurement.

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
re-arm mechanism recorded nothing for it ([#165](https://github.com/ctoforaday/special-circumstances/issues/165) — an earlier reading
called that mechanism dead; withdrawn, its coverage is merely unreliable),
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
