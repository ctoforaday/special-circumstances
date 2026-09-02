# Special Circumstances

> *An adversarial human/AI methodology suite. It argues with you on purpose.*

## What it is

An AI assistant that agrees with you is comfortable and dangerous. It ships your mistakes, launders your assumptions back to you as confirmation, and gives a confident wrong answer in the same tone as a right one.

**Special Circumstances** is a set of [Claude Code](https://claude.com/claude-code) plugins built to do the opposite. Installed, they make the assistant behave like a good colleague rather than an eager one:

- **Check instead of trusting.** Documentation, comments, and your own claims are treated as unverified until opened at the source.
- **Ask instead of guessing.** An under-specified instruction that cannot be undone gets a question, not a plausible interpretation.
- **Stop instead of spinning.** Three failed attempts at the same fix end the loop and escalate to you.
- **Prove instead of asserting.** "It works" has to come from a command that was actually run, not from a clean-looking diff.
- **Argue before concluding.** Research goes through a structured debate — one agent builds the case, another tries to break it — before anything is called true.

A good argument here is a courtesy, not an attack.

## The four plugins

| Plugin | What it does | Status |
|---|---|---|
| [**prosthetic-conscience**](plugins/prosthetic-conscience) | The working discipline — always-on rules for how the assistant behaves under pressure. The base plugin. | Shipping |
| [**frank-exchange-of-views**](plugins/frank-exchange-of-views) | A research debate engine — one team builds, one team audits and gates, a judge rules. | Shipping |
| [**gray-area**](plugins/gray-area) | Trajectory evidence — what a session *actually did*, as against what it *reported* doing. | Early, usable |
| [**sleeper-service**](plugins/sleeper-service) | Autonomous self-improvement — the suite researching its own rules, human-gated at promotion. | Scaffold only |

The names come from the [Culture](https://en.wikipedia.org/wiki/The_Culture), Iain M. Banks's civilisation of humans and Minds who treat a good argument as an act of respect. *Special Circumstances* is the Culture's division for the hard, consequential work; the plugins are named for its ships and drones — the drone that keeps you honest, a heated argument diplomatically put, the mind-reading ship the others shun, and the vast quiet vessel that works while you sleep.

## Install

```text
/plugin marketplace add ctoforaday/special-circumstances
/plugin install prosthetic-conscience@special-circumstances
/plugin install frank-exchange-of-views@special-circumstances
/plugin install gray-area@special-circumstances
/plugin install sleeper-service@special-circumstances
```

`prosthetic-conscience` is the base — the other three preload its rules — but each plugin is individually useful, so install only what you want.

Then check your environment:

```text
/prosthetic-conscience:doctor
```

That gives a deterministic READY / DEGRADED / BLOCKED verdict on your toolchain (git, gh, qlty) and on the plugin's own helper binaries. `--fix`, with your consent, builds or downloads anything missing. If it can reach neither a release nor a network, [docs/setup-script.md](docs/setup-script.md#building-the-hook-binaries-by-hand) has the manual build.

### On Claude Code on the web

The `/plugin` commands above are command-line only. In web sessions use the `claude` equivalents:

```text
claude plugin marketplace add ctoforaday/special-circumstances
claude plugin install prosthetic-conscience@special-circumstances
```

Plugins resolve when a session starts, so a mid-session install applies to the *next* session — and a cloud container is discarded when the session ends. To have them present from the first prompt, install them from the environment's **Setup script** field: see [docs/setup-script.md](docs/setup-script.md).

## Try it

- **`/prosthetic-conscience:doctor`** — check the environment before anything assumes a toolchain.
- **`/research <topic>`** — run a full research debate and watch the argument happen.
- **`/plan-audit <file>`** — put an implementation plan through a PASS/FAIL gate before you build from it.
- **`/checkpoint`** and **`/resume`** — write and re-read the note that carries your work across a context compaction.
- **`/gray-area:audit-checkpoint`** — check that note's claims against what the session actually ran.

Everything else is ambient: once installed, the rules apply to every session without being invoked.

## What each plugin does

### prosthetic-conscience — the working discipline

The base plugin: 22 skills, of which **ten load in every session**. Each is a short contract written in a `BEFORE / During / AFTER · YOU MUST` grammar, so it says exactly when it applies.

| Rule | In one line |
|---|---|
| **terse-communication** | Strict token economy — no filler, no self-congratulation, no play-by-play. |
| **semantic-consent** | You own intent, the assistant owns syntax; ask before anything irreversible. |
| **plan-act-reflect** | Plan, act, reflect — as separate steps, with an approved plan before complex work. |
| **anti-spinning** | Three strikes on a failing approach, then stop and escalate. Honor a cancel immediately. |
| **validation-loop** | Write down the commands that prove a change works, then actually run them. |
| **think-around-problem** | Never satisfice; explore alternatives in proportion to the stakes. |
| **agent-guardrails** | No privileged mutations without approval, no secrets sent off-box. |
| **context-efficiency** | Shield the context — delegate bulk reading, leave artifacts that survive compression. |
| **complete-the-concept** | A change is finished when the concept is, not when the first commit merges. |
| **facts-are-fields** | Facts other parties act on belong in a field something can refuse, not in a filename or a regex. |

The other twelve load on demand by description: pair-programming, spec-driven-development, test-driven-development, refactoring-safety, project-memory, critical-stance, context-checkpointing, scratch-policy, design-by-contract, and proficiency guides for git, markdown and qlty.

It also ships `/plan-audit` — a binary PASS/FAIL auditor that puts an implementation plan against a five-section standard — and a set of Go hook binaries that enforce the mechanically checkable rules even where prompts never fire. See [Under the hood](#under-the-hood) for how those two halves fit together.

### frank-exchange-of-views — the research debate engine

Give it a topic; it gives back a verified research deliverable with the entire adversarial record preserved beside it. Three seats with genuinely different mandates:

- **Blue** is *additive*. It researches, drafts, and synthesizes by **union, never summary** — its terminal goal is to be true *at the leaf node*, meaning every claim traces to a source somebody can open.
- **Red** is *subtractive*, and **owns the PASS/FAIL gate**. It opens blue's citations at the source, grades **trust** (corroboration confidence) and **risk** (likelihood × impact × complexity), and files every defect as a tracked *gap*.
- **The bench** (a `lead-judge` agent) rules only on the contested docket, in writing. It holds the terminal values — correctness > thoroughness > economy, safety above all — and assembles the final report. It never gates a round; passing is red's call alone.

**Termination is judged, never counted.** A run ends on a red PASS or on detected deadlock. The round ceiling is a spending limit, not a definition of done — and a run that hits it is stamped CEILING-TERMINATED so nobody mistakes "ran out of budget" for "verified".

### gray-area — trajectory evidence

Every session writes a transcript: every tool call, every result, the causal chain, the timing. That record holds **what a session actually did**, which is a different thing from what it reported doing. The rest of this suite is built on separating those two — red audits blue's citations at the source rather than believing them. Gray Area applies the same move to the session itself.

What it can do today:

| Command | Question it answers |
|---|---|
| `/gray-area:audit-checkpoint` | Did the checks a checkpoint claims to have run actually run — and are they still fresh? |
| `/gray-area:audit-pr-body` | Do the commands in this pull request body appear in the session that wrote it? |
| `/gray-area:audit-repetition` | What was done more than once, and what was repeated three times back to back? |
| `/gray-area:audit-seat-coverage` | Does the record name every subagent transcript that exists? |
| `gray-area tools` | What did a given agent actually invoke? (command line) |

Each row cites both documents — the claim and the evidence — so a reader can check it rather than trust it. Verdicts are deliberately weak where the record is weak: `NO-EVIDENCE` means *nothing matched*, printed with the tokens searched, never *it did not happen*.

Reading transcripts is a surveillance capability, and the plugin is scoped accordingly: the manifest is an index of where trajectories are, never a copy of their contents, and nothing leaves the box.

### sleeper-service — autonomous self-improvement

The learning plugin, designed to improve the suite while you sleep: a `/self-improve` loop that researches how one of the system's own rules should evolve, a `/graduate` pipeline, and a promotion ladder (insight → memory → rule → cheatsheet). It will write only to `research/` and `ideas/`; **promoting anything into an actual rule always requires a human.**

**Status: scaffold only (v0.1.0).** The design is settled; the commands are not built yet.

## Anatomy of a research run

Run `/research <topic>` (or `/frank-exchange-of-views:research <topic> [--lanes N] [--lenses N] [--max-rounds N]`). One end-to-end run:

1. **Blue builds.** Parallel lanes and lenses research the topic; blue synthesizes them additively into a living report, carrying a frontier of hypotheses and preserving the candidate drafts.
2. **Red audits at the leaf.** Red opens each citation at its source, grades trust and risk, and files every defect as a gap. Red owns the verdict.
3. **Blue repairs, round by round.** Gaps come back as obligations; blue revises; red re-audits the changed text. Repeat until red passes or the debate deadlocks.
4. **The bench rules** on whatever is still contested, in writing, and assembles the final report from the audited sources.
5. **The record captures all of it** — findings, gaps, opinions, closures, disputes, friction — as validated events in an append-only log.

The run lands in `research/<date>_<slug>/`:

| Artifact | Contents |
|---|---|
| `README.md` | The run's front door — verdict, gaps, and what each document below holds |
| `report.md` | The research — verdict, TL;DR, the Catechism, foundations, analysis, risks, open questions |
| `docket.md`, `debate.md`, `judgments.md` | The adversarial record: red's board, the round-by-round transcript, the motions and their rulings |
| `evidence.md`, `run.md`, `CHANGELOG.md` | The computations in full; the friction, record check and cost; the report's own revisions |
| `report.html` | The whole set with real tabs and cross-document links — one self-contained file, no server |
| `records/*.jsonl` | The append-only event log — the authoritative record of every act |
| `blue/`, `red/` | The seats' working surfaces — blue's living report and its candidate drafts |
| `inputs/` | The brief, the pinned configuration, the scorecards — what the run was asked to do |
| `trajectories/` | Per-seat execution journals |
| `cost.md` | Per-seat, per-round token and dollar accounting |

The debate transcript and the gap board are **not files**: they are projections rendered from the event log on demand (`show debate`, `show board`), so there is never a second copy to drift from the record.

A real run ships in the repo as evidence: [`research/2026-08-23_research-loop-counterparts/`](research/2026-08-23_research-loop-counterparts/). It is *ceiling-terminated* — it hit its round budget while still converging — and the report says so in its own verdict line, naming the debt it leaves (a final blue revision no red pass audited). That is the point: the record makes the residual honesty checkable.

## Under the hood

Detail from here down. Skip it unless you want to know why the parts are shaped the way they are.

### Defense in depth, not duplication

A skill states the *semantic* rule — the part no regex can catch. A hook enforces the *mechanically checkable* part, in the places where prompts never fire: auto mode, headless `claude -p`, scheduled runs. Neither half is sufficient. A rule with no hook is advice; a hook with no rule is a regex nobody can argue with.

One binary handles each event, so a single process sees one parse of the payload and emits one answer:

| Event | What fires |
|---|---|
| `PreToolUse` | Secrets gate (fails closed, can deny) and push-freeze guard (warns, never blocks) |
| `PostToolUse` | The quality gate — `qlty fmt` and `qlty check` on what was just written |
| `PostToolUseFailure` | The anti-spinning strike counter, keyed on (tool, target) |
| `SessionStart` | Toolchain nudge, and the checkpoint handed back |
| `PreCompact` · `SessionEnd` · `SubagentStop` | The checkpoint sealed at each seam |
| `FileChanged` | A validation check marked stale when its trigger surface moves |
| `PostCompact` | Observation only — scoring what each summary kept |

Every hook is wrapped in a bootstrap guard: a fresh plugin version ships from git *without* binaries, and an unguarded hook crash-storms every tool call in that window. The guard degrades to one line of stderr pointing at `/prosthetic-conscience:doctor --fix`.

### Compaction survival — the Memento problem

Compaction replaces the transcript with a summary. The summary is good at what happened and worst at **what you were about to do**: the exact validation commands, the ordered next actions, the handles to work still running in the background.

So the agent keeps one `CHECKPOINT.md`, overwritten in place and sealed at every seam. `SessionStart` hands it back on the far side; `/checkpoint` writes it and `/resume` prints it in full. Two things measurement settled, not design:

- **Restore runs on `SessionStart`, on every source including `compact`.** Routing the compaction boundary through `PostCompact` looked better — that event actually receives the summary — but it cannot inject anything into the model at all, and it runs *after* `SessionStart` besides.
- **The resumed agent treats the restored note as a claim, not a fact.** In testing it recovered every value exactly, attributed them honestly to the hook, and flagged the payload as injection-shaped anyway. Provenance framing does not prevent that, and the design no longer pretends to. What it buys is that the suspicion attaches only to the note's real content, because the hook adds no instruction of its own.

It lives in prosthetic-conscience rather than gray-area on purpose: keeping a note about your own work asks far less of you than a tool that reads your transcripts, and you should not have to accept the second to get the first.

### The tool is the contract

In a debate run, seats never hand-write state. Every act goes through `feov-record`, a Go command-line tool that validates each write — required fields, cross-references checked at write time, legal state transitions, unguessable identifiers it assigns itself — and appends to a JSONL event log. Everything a seat reads back is a view over that log. This is why a hand-written board once drifted to 3-open/15-closed against an event log that said 9-open/9-closed, and why it cannot now.

### The line gray-area will not cross

> **Exploration may summarize. Adjudication must cite.**

An agent asked to read a transcript and report what it sees is a summarizer — cheap, useful for finding where to look, and non-deterministic, unreproducible, uncitable. Fine for a hypothesis. Disqualifying for a finding. This suite spent a full cycle removing self-report from the evidence chain; *"an agent says the transcript shows"* would put it straight back, one layer up and harder to catch. So any query behind a finding returns primary evidence — the event id, the line, the tool call — and cites the trajectory, never the index's opinion of it.

## Roadmap

> **Outline, not a finished document.** Grouped by sequence; dates are unset. Drawn from [`plans/handoff.md`](plans/handoff.md) and [`plans/tool-is-the-contract.md`](plans/tool-is-the-contract.md).

**Now — retire the file paths, so the tool is the only contract.** Seats already *write* every act through `feov-record`, but they still *read* the run by opening markdown at paths learned from a prompt. Measured, ~44% of the seat-prompt corpus is absorbable into the tool. That means `show` views for everything a seat reads, preconditions the tool refuses rather than ceremony a prompt asks for, and a free-text channel through the verbs so a seat writing prose stops fighting shell escaping (one run paid 68 escaped-quote commands, 9 heredocs and 37 temp-file stagings for the lack of it).

**Next — correctness and ergonomics the last run demanded.**

- **Assembly stops being an LLM.** The union-copy at the end of a run is a mechanical operation the model has been measured *corrupting* (one run regressed 6 of 7 Catechism answers). A script does the copy; a small model writes only the summary. A correctness lever, not a cost one — assembly is ~$2 of a ~$54 run.
- **Positional `--seat-id`.** Five failures per run, and it cannot be inferred. Making it positional removes a flag name to forget, but it breaks every seat prompt and 22 goldens, so it lands on its own.
- **Four verbs seats asked for by hitting their absence:** amending a mint, amending a prior closure from a seat that did not enter it, withdrawing an entry written in error, and closing with a named residue.
- **Scorecard metric fixes** — several telemetry metrics still parse hand-written prose while the data lives in events.

**Later — new capability.** A SQLite-backed index so the event log is queryable without re-rendering (the JSONL stays authoritative), and the rest of gray-area's mining: friction claims adjudicated against the trajectory that refutes them, act-versus-claim discrepancy, stall forensics from the timing.

<!-- TODO (microsite): expand each roadmap item into a page; add a "what a run costs" section grounded in the evidence run; link the plans/ documents as the authoritative design record. -->
<!-- TODO: write the TestSeatNeverNeedsAPath validation loop first and let it fail; re-measure the archive byte-parity gap on the first post-timestamp run (the 34,086-vs-7,527 figure is contaminated by legacy event ordering). -->

## Working on this repository

> This repository is the *workshop*. `CLAUDE.md` governs work **on** the repo and reaches no installing project; consumer behaviour lives entirely under `plugins/`.

| Path | Role |
|---|---|
| `plugins/<name>/` | The product: everything a consumer installs — skills, agents, commands, hooks, Go tools |
| `.claude-plugin/marketplace.json` | Marketplace manifest listing the four plugins |
| `plans/` | Design artifacts under review — each arrives as a pull request, graduates into the plugins |
| `research/` | Completed debate runs |
| `ideas/` | Proposals from `/self-improve`, pre-promotion |
| `docs/` | Setup and operational notes for this repo |

Versions move at a **release boundary**, not per pull request: an ordinary change leaves `version` alone, and a release bumps the manifest and tags the same commit. A guard refuses a tag whose manifest disagrees, so the two cannot drift.

## Origins

Ported from an earlier "Antigravity Meta Brain" experiment. The *methodology* came across — the rules, the debate protocol, the gates, the templates. The hand-rolled *harness* did not: a 787-line orchestrator, prompt and skill compilers, watchdogs, and a local-LLM proxy stack, all made redundant by Claude Code's native model. See [`plans/claude-port-plan.md`](plans/claude-port-plan.md) for the teardown.

## License

MIT © 2026 Gregory Block
