# Special Circumstances

> *An adversarial human/AI methodology suite. It argues with you on purpose — and it researches its own rules while you sleep.*

## What it is

An AI assistant that agrees with you is comfortable and dangerous. It ships your mistakes, launders your assumptions back to you as confirmation, and gives a confident wrong answer the same tone as a right one.

**Special Circumstances** is built to do the opposite. It is a suite of three Claude Code plugins that make the assistant *argue back — with reasons*: verify claims at their source instead of trusting them, refuse an under-specified instruction instead of guessing, break a repeating failure loop instead of spinning, and run structured adversarial debate over research before calling anything true. A good argument here is a courtesy, not an attack.

The suite is named for the [Culture](https://en.wikipedia.org/wiki/The_Culture) — Iain M. Banks's civilisation of humans and Minds who treat a good argument as an act of respect. *Special Circumstances* is the Culture's division for the hard, consequential work; the three plugins are named for its ships and drones.

| Plugin | Named for | Role |
|---|---|---|
| [**prosthetic-conscience**](plugins/prosthetic-conscience) | the drone that keeps you honest | Always-on working discipline for interactive and headless sessions |
| [**frank-exchange-of-views**](plugins/frank-exchange-of-views) | a heated argument, diplomatically put | A research debate engine — blue builds, red audits and gates, a bench rules |
| [**sleeper-service**](plugins/sleeper-service) | the GSV quietly running vast hidden projects | Autonomous self-improvement, always human-gated at promotion |

## The three plugins

### prosthetic-conscience — the working discipline

The base plugin. It carries a set of **always-on skills** that bind every session and encode how the assistant should behave under pressure, each a small contract in a `BEFORE / During / AFTER · YOU MUST` grammar:

- **terse-communication** — strict token economy; no filler, no self-congratulation, name things by function.
- **semantic-consent** — the human owns intent, the assistant owns syntax; ask before an irreversible or outward-facing action, don't guess an under-specified instruction.
- **plan-act-reflect** — separate planning, execution, and reflection; an approved plan before acting on complex work.
- **anti-spinning** — stop repeating a failed approach after three strikes; honor a cancel immediately.
- **validation-loop** — write down the exact commands that prove a change works, then actually run them; never claim success from a clean diff.
- **think-around-problem** — never satisfice; treat docs, comments, and the human's own claims as unverified until checked.
- **agent-guardrails** — single-purpose agents, no privileged mutations without approval, no secrets sent off-box — the backstop for when permission prompts don't fire (auto mode, headless, scheduled loops).

On top of the always-on set it ships **on-demand skills** (pair-programming, spec-driven-development, test-driven-development, refactoring-safety, project-memory, and proficiency guides for git / markdown / qlty), a **plan-audit gate** (`/plan-audit` runs an implementation plan through a binary PASS/FAIL auditor against the five-section spec standard), and **Go hook binaries** that enforce the pattern-matchable rules deterministically:

| Hook | Enforces |
|---|---|
| `sc-doctor` | Environment preflight — a READY / DEGRADED / BLOCKED verdict (`/prosthetic-conscience:doctor`) |
| `sc-secrets-gate` | Blocks secrets/tokens from leaving in web calls |
| `sc-quality-gate` | Quality checks on writes |
| `sc-push-freeze-guard` | Guards pushes |
| `sc-toolchain-nudge` | Nudges toward the installed toolchain |
| `sc-recall-index` | Indexes for recall |

**The distinctive idea:** the rules are defense in depth — a skill states the *semantic* rule for what a regex can't catch, and a hook enforces the *pattern-matchable* part where prompts never fire.

### frank-exchange-of-views — the research debate engine

Given a topic, it produces a verified research deliverable with the entire adversarial record preserved beside it. Three seats, with genuinely different mandates:

- **Blue** is *additive*. It researches, drafts, and synthesizes by **union, never summary** — its terminal goal is to be true *at the leaf node* (every claim traceable to a checkable source).
- **Red** is *subtractive* and **owns the PASS/FAIL gate**. It audits blue's living report at the leaf — verifying every citation at its source — grades **trust** (corroboration confidence) and **risk** (likelihood × impact × complexity), and mints each defect as a tracked *gap* in a ledger.
- **The bench** (a `lead-judge` agent) adjudicates only the contested docket with written opinions, holds the terminal values (correctness > thoroughness > economy; safety above all), and assembles the final report by union-copy. It never gates rounds — passing is red's call alone.

It runs as a **sandboxed workflow script** (`debate.js`) that drives the mechanics — best-of-N lanes and lenses, round sequencing, the **Catechism** (a "worth our time" decision adapted from Heilmeier), and a safety ceiling that bounds cost. **Termination is judged, never counted:** the run ends on a red PASS or on detected deadlock/spinning, and the round ceiling only caps spend.

Every act is recorded through **`feov-record`, a Go CLI that is the single contract**. Seats don't hand-write state; they call the tool, and the tool validates every write — required fields, cross-reference integrity checked at write time, legal state transitions, and tool-assigned unguessable identifiers — then writes an **append-only JSONL event log** plus **rendered markdown projections** of the board, ledger, and archive. The event stream is the authoritative, queryable, attributable record; the markdown is a view of it.

**The distinctive idea:** the seat that builds is not the seat that judges, and the judge of the *gate* is not the judge of the *dispute* — three separated mandates, with a machine-checked event log so no seat can quietly rewrite the record.

### sleeper-service — autonomous self-improvement

The learning plugin. It is designed to improve the suite while you sleep: a `/self-improve` loop (daily default cadence) that researches how one of the system's own rules should evolve, a `/graduate` pipeline, and a continuous-learning **promotion ladder** (insight → MEMORY → rule-skill → cheatsheet). It writes only to `research/` and `ideas/`; **promotion into an actual rule or skill always requires the human.** It invokes frank-exchange-of-views to do the research.

**Status: scaffold only (v0.1.0).** The design is settled; the commands are not yet built out.

**The distinctive idea:** the system dogfoods its own debate engine on its own rulebook, but a machine may only ever *propose* a rule change — the promotion step is a hard human gate.

## How a research debate works

Run `/frank-exchange-of-views:research <topic>` (or `/research <topic>`). One end-to-end run:

1. **Blue builds.** Parallel lanes and lenses research the topic; blue synthesizes them additively into a living `report.md`, carrying a frontier of hypotheses and preserving candidate drafts. Every claim aims to be true at the leaf.
2. **Red audits at the leaf.** Red opens each citation at its source, grades trust and risk, and mints every defect as a *gap* in the ledger through `feov-record`. Red owns the verdict: PASS only if the report survives audit.
3. **Blue repairs, round by round.** Red's gaps come back as obligations; blue revises; red re-audits the changed text. This repeats until red passes or the debate deadlocks.
4. **The bench rules.** On a contested docket, the `lead-judge` issues written opinions, closes or carries each gap, and assembles the final report by union-copy from the audited sources.
5. **The record captures all of it.** Every act — findings, gaps, opinions, closures, disputes, friction — is a validated event in the append-only JSONL log, projected to markdown.

A finished run lands in `research/<date>_<slug>/`:

| Artifact | Contents |
|---|---|
| `report.md` | The deliverable — verdict, TL;DR, the audited report. The one genuine file, not a projection |
| `records/*.jsonl` | The append-only event log, one file per seat-turn — the authoritative record |
| `debate.md` | The full three-party transcript (blue / red / bench), union not summary |
| `blue/`, `red/` | Living reports, ledger, archive, citation ledger — markdown projections of the events |
| `friction.md` | Every place a seat hit a missing verb or a rough edge — the backlog, self-reported |
| `cost.md` | Per-seat, per-round token and dollar accounting |

A real, complete run ships in the repo as evidence: [`research/2026-07-18_gray-area-telemetry/`](research/2026-07-18_gray-area-telemetry/). It is a *ceiling-terminated* run — it hit its round ceiling while still converging — and the report is explicit that this is not a judged failure. The debt it leaves (a final blue revision no red pass has audited) is stated as the finding, which is the point: the record makes the residual honesty checkable.

## Getting started

Install from the Claude Code plugin marketplace:

```text
/plugin marketplace add ctoforaday/special-circumstances
/plugin install prosthetic-conscience@special-circumstances
/plugin install frank-exchange-of-views@special-circumstances
/plugin install sleeper-service@special-circumstances
```

`prosthetic-conscience` is the base; the other two preload its rule-skills, and `sleeper-service` invokes `frank-exchange-of-views`. One marketplace install gets all three; each is individually useful.

After installing, check your environment:

```text
/prosthetic-conscience:doctor
```

This runs the `sc-doctor` binary for a deterministic READY / DEGRADED / BLOCKED verdict on the toolchain (git, gh, qlty) and the hook binaries. `--fix`, with your consent, builds or fetches any missing hook binaries.

<!-- TODO: document the manual build path for the Go hook + feov-record binaries (go build in plugins/*/tools) for users whose /doctor --fix can't fetch a prebuilt binary, and state the toolchain prerequisites (Go version, CGO for -race). -->

### A quick tour

- **`/prosthetic-conscience:doctor`** — verify the environment before anything assumes a toolchain.
- **`/research <topic>`** — run the debate engine; watch `research/<date>_<slug>/debate.md` grow as blue and red argue and the bench rules.
- **`/plan-audit <file>`** — put an implementation plan through the spec-driven-development gate.
- **`/self-improve`** — let the suite research how one of its own rules should evolve (writes only to `ideas/` and `research/`). *(sleeper-service — scaffold.)*

## Repository layout

> This repository is the *workshop*. `CLAUDE.md` governs work **on** the repo; nothing in it reaches an installing project. Consumer behaviour lives entirely under `plugins/`.

| Path | Role |
|---|---|
| `plugins/<name>/` | The product: everything a consumer installs — skills, agents, commands, hooks, Go tools |
| `.claude-plugin/marketplace.json` | Marketplace manifest listing the three plugins |
| `plans/` | Design artifacts under review — each arrives as a PR, graduates into the plugins |
| `research/` | Completed debate runs (the working corpus; seeded by `/research`) |
| `ideas/` | Proposals from `/self-improve`, pre-promotion |
| `README.md`, `plugins/*/README.md` | The shipped documentation |

## Roadmap

> **Outline, not a finished doc.** This sketches direction toward a future project microsite; items are drawn from [`plans/handoff.md`](plans/handoff.md) and [`plans/tool-is-the-contract.md`](plans/tool-is-the-contract.md). Grouped by sequence, not by date — **dates are unset.**

### Now — retire the file paths so the tool is the only contract

The debate engine's biggest open change. Seats already *write* every act through `feov-record`, but they still *read* the run by opening markdown at paths learned from a prompt — the asymmetry that let a hand-written board drift to 3-open/15-closed against the event log's 9-open/9-closed. Measured: **~44% of the seat-prompt corpus is tool-absorbable** (file-path instructions plus procedural ceremony the tool could enforce).

- **Read through the tool, not the filesystem.** Add `show` views (`report`, `lens-passes`, `gap-patterns`, `law`) so no prompt names a storage path. `report.md` stays the one genuine file — it is the deliverable, not a projection.
- **Enforce ceremony instead of asking for it.** `register` becomes a precondition every verb refuses before; `mint` returns near-matches and requires an explicit not-a-reopen; demanded lineage reads become impossible to skip.
- **Wire stdin through the verbs** so a seat writing rich prose stops fighting shell escaping (the last run paid an "escaping tax" of 68 escaped-quote commands, 9 heredocs, 37 temp-file stagings). Give the nine bare verbs a free-text channel.
- **Round-parity guard** — distinguish an evidentiary FAIL from a no-response FAIL, so telemetry stops reporting a missing blue turn as failed repairs.

<!-- TODO: write the TestSeatNeverNeedsAPath validation loop first and let it fail; re-measure the archive byte-parity gap on the first post-timestamp run (the 34,086-vs-7,527 figure is contaminated by legacy event ordering). -->

### Next — correctness and ergonomics the last run demanded

- **Assembly stops being an LLM.** The union-copy at the end of a run is a mechanical operation the model has been measured *corrupting* (one run regressed 6 of 7 Catechism answers). A script does the copy; a small model writes only the TL;DR and synopsis. This is a correctness lever, not a cost one — assembly is ~$2 of a ~$54 run.
- **Positional `--seat-id`.** Five failures per run, and it cannot be inferred (shell state doesn't persist, cwd resets, all subagents share the parent session id). Making it a positional argument removes a flag name to forget — but it breaks every seat prompt and 22 goldens, so it lands on its own.
- **Four verbs seats asked for by hitting their absence:** `mint --amend` (a first mint is currently permanent), `amends_prior` (usable by a seat that didn't enter the original closure), `withdraw` (for an archive block written in error), `closed_with_residue` (a check passes but a residue of the same class survives at an unnamed site).
- **Scorecard metric fixes** — several telemetry metrics parse hand-written prose while the data lives in events (`anchored_closures_pct`, `lines_of_inquiry`, `round_parity_failures`).

### Later — new capability

- **A SQLite-backed index for the event log.** Make the append-only JSONL queryable without re-rendering. <!-- TODO: flesh out from plans/record-tool.md — schema, what queries it serves, how it stays a projection (the JSONL remains authoritative). -->
- **Gray Area — the fourth plugin (trajectory-evidence mining).** Not yet shipped. The problem it solves is already captured in the evidence run: a seat logged a friction it *had already recovered from* — friction is unverified self-report, and only the execution trajectory can adjudicate the claim against its own refutation. Constraint for the miner: ~11% of tool calls are invisible to a naive string matcher because seats alias the binary to a shell variable — it must parse shell structure, not grep for a name. <!-- TODO: scope Gray Area's acceptance test and its relationship to friction.md as the raw feed. -->

<!-- TODO (microsite): expand each roadmap item into a page; add a "what a run costs" section grounded in the gray-area run; link the plans/ documents as the authoritative design record. -->

## Origins

Ported from an earlier "Antigravity Meta Brain" experiment (the private `AgentOrange` repo). The *methodology* came across — the rules, the debate protocol, the gates, the templates. The hand-rolled *harness* did not: a 787-line orchestrator, prompt/skill compilers, watchdogs, and a local-LLM proxy stack, all made redundant by Claude Code's native model. See [`plans/claude-port-plan.md`](plans/claude-port-plan.md) for the full teardown.

## License

MIT © 2026 Gregory Block
