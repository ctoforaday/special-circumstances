# Gray Area — trajectory mining and the hook surface

> Foundations for the **fourth plugin**. Seeded by PR #3 (the Memento proposal), retargeted by the
> operator on 2026-07-18 from "a checkpoint system" to "a trajectory-evidence plugin".
> Status: design research. No code. Companion: [`reasoning-telemetry.md`](reasoning-telemetry.md).

**Named for** the GCU *Gray Area* (*Excession*) — the Culture's mind-reader, the ship that
establishes what actually happened by reading directly, and is shunned by other Minds for doing it.
The capability being necessary and distasteful at once is the intuition to preserve.

**Client under test:** Claude Code **2.1.220** (`GIT_SHA 4073f595…`, built 2026-07-24). Hook facts
below are read out of that binary's own event catalogue, not from documentation.

---

## 1. What the plugin is for

**The spine is trajectory mining: a transcript holds what actually happened, as opposed to what was
reported.** That is the whole of it. Continuity across compaction — which seeded PR #3 and was
carried here through the retarget — has been split back out to `prosthetic-conscience`; §4 records
the argument.

The case for a plugin rather than another script is that the consumers already exist, each written
separately against the same JSONL:

| Consumer | Status |
|---|---|
| Bench integrity inspection (attestation vs actual tool calls) | shipped |
| `recordJoinAudit` — declared events vs transcript invocations | shipped |
| `contextUse` — per-seat peak context against the window | shipped |
| `cost-audit` — per-seat model and token accounting | shipped |
| Attestation audit (E0.5a) | done **by hand**, never tooled |
| Friction mining (E0.5e) | done **by hand**, never tooled |
| Wall-clock forensics (1015 API rounds × ~23.8s) | done **by hand**, never tooled |

Seven instances of one capability. Three of them are a person doing it manually.

**Wanted capabilities, from the operator:** mine trajectories for problems (human frustration, signs
of deception, seats going in circles); give the bench toolsets to judge red and blue behaviour, and
a symmetric one to judge the bench; carry commit and pull-request metrics on the same substrate.

---

## 2. The capability picture, corrected

PR #3 and the 2026-07-18 debate run concluded that reasoning is not persisted, and designed around
its absence. That conclusion was measuring a default. `--thinking-display summarized` retains
thinking summaries in transcripts, headless included, and propagates to subagents — measured, see
[`reasoning-telemetry.md`](reasoning-telemetry.md) §1.

**This does not change the adjudication design, and it must not.** Summaries are second-hand
self-report; they stay on the exploration side of *exploration may summarize, adjudication must
cite*. What changes is that the exploration side now has a channel it was assumed not to have.

Three constraints carry forward unaltered:

- **Acts are ground truth.** You cannot fake having made a tool call, nor fake not having made one.
- **Integrity findings are separated from merits.** Untidy reasoning that reached a sound conclusion
  is not a finding.
- **Every inspection is declared**, with what it relied on quoted, because a party must be able to
  answer a finding it could not watch being made.

---

## 3. The hook surface at 2.1.220

The suite currently uses **three** hook events. The client ships **thirty-one**. Enumerated from the
binary's own catalogue:

```text
PreToolUse  PostToolUse  PostToolUseFailure  PostToolBatch  PermissionRequest  PermissionDenied
UserPromptSubmit  UserPromptExpansion  SessionStart  SessionEnd  Stop  StopFailure
SubagentStart  SubagentStop  PreCompact  PostCompact  Setup  TeammateIdle
TaskCreated  TaskCompleted  Elicitation  ElicitationResult  ConfigChange  InstructionsLoaded
WorktreeCreate  WorktreeRemove  CwdChanged  FileChanged  DirectoryAdded  MessageDisplay
```

Hooks also come in three **types**, not one: `command` (a shell command), `prompt` (an LLM evaluates
a condition), and `agent` (an agent runs with tools). The latter two are available **only** for
`PreToolUse`, `PostToolUse` and `PermissionRequest`. This matters for a suite whose standing excuse
for a rule living in a skill rather than a hook is "a regex can't catch it" — on tool events, it no
longer has to be a regex.

### The five that change what we can build

**`SubagentStop` — input carries `agent_id`, `agent_type`, and `agent_transcript_path`.**
The single most consequential event for this plugin. The seat's own trajectory is handed over, by
path, at the moment the seat finishes, matched on `agent_type`. Every mining consumer in §1 that
currently sweeps `~/.claude/projects/` guessing at which files belong to which seat becomes a hook
that is *given* the file. Deterministic attribution, no glob, no race with a live store, and it
fires in headless runs where nobody is watching.

**`PreCompact` — exit 0, and stdout is appended as custom compact instructions.**
This refutes the central engineering constraint of PR #3, which reads:

> *"`PreCompact` cannot reflect, and cannot edit the summary… per the docs it **cannot inject
> content into the compacted summary**."*

It cannot *write* the summary, and it still cannot reflect — it is a script. But its stdout steers
the summarizer. A seal hook can therefore say *"preserve verbatim: the validation loop, the ordered
next actions, the beyond-plan flag"* and have the harness's own summary carry them, instead of
racing it with a second narrative injected afterwards. The plan's division of labour survives; the
mechanism improves.

**`PostCompact` — input carries `trigger` and `compact_summary`.**
PR #3 flagged this event as unconfirmed and refused to depend on it. It exists, and it receives the
summary.

> **Corrected 2026-07-29.** This paragraph continued: *"Restore stops being blind: a restore hook
> can compare the checkpoint against what the summary actually kept and inject only the delta …
> the direct fix for R4 and R3."* **Wrong, and measured wrong.** `PostCompact` cannot inject —
> it is absent from the client's `hookSpecificOutput` union, its stdout goes to the user rather
> than the model, and an end-to-end marker test found it `NOT-SEEN` where the same marker from
> `SessionStart` arrived as a `hook_additional_context` attachment. It also runs *after*
> `SessionStart(compact)`, so even a hypothetical injector would be too late to diff against.
> Restore is `SessionStart`, every source. R3 and R4 stay live. Evidence:
> [`hook-surface-spike.md`](hook-surface-spike.md) §3a. PR #3 was right to refuse to depend on it,
> for a reason better than the one it gave.

What survives: `PostCompact` is an **observation** point. It can record what each summary kept or
dropped against the seal that asked for it — a signal that accumulates across boundaries, even
though nothing can act on it within one.

**`SubagentStart` — `agent_id` and `agent_type` in, `additionalContext` out, matched on `agent_type`.**
Per-seat context injection at spawn, deterministically, without touching a prompt. The debate
engine's seat mandates, the read-only restore protocol, a per-seat trajectory marker — all can be
injected by type rather than trusted to a prompt that a seat may or may not have been given.

**`FileChanged` — matcher names the files to watch; fires on change/add/unlink.**
This is the mechanism for PR #3's improvement **I2**, which was written as a discipline because no
mechanism existed. The live failure it came from: a CI gate fires on any *protocol surface* edit
(`debate.js` seat prompts, `agents/*.md`), a post-compaction agent pushed an instance-only fix, and
CI went red because the rule's trigger surface had been flattened by the summary. A trigger surface
is a file glob. A hook can hold it, and re-arm the check when the surface is touched — no memory of
the rule required.

### The rest, briefly

| Event | What it buys |
|---|---|
| `PostToolUseFailure` | Deterministic strike counting for **anti-spinning** — the 3-strike limit currently lives only in a skill and is enforced by the model against itself *(filed: #127)* |
| `StopFailure` | Matched on error type (`rate_limit`, `overloaded`, `billing_error`, …) — resilience for **sleeper-service**'s unattended loops |
| `TaskCreated` / `TaskCompleted` | Exit 2 prevents creation or completion — the enforcement point for PR #3's improvement **I1**, that a forward actionable must land in the durable queue and not only in the note |
| `InstructionsLoaded` | Observability-only: which `CLAUDE.md` or rule loaded, why (`session_start`, `nested_traversal`, `path_glob_match`, `include`, `compact`), and what triggered it. Answers "did the rules actually bind on this run?" — currently unanswerable |
| `ConfigChange` | Exit 2 blocks a settings or skills change mid-session — an **agent-guardrails** surface, and the one path around `sleeper-service`'s human promotion gate *(filed: #128)* |
| `SessionEnd` | Reason-matched (`clear`, `logout`, `prompt_input_exit`, `other`) — the seal point for a session that ends without compacting |
| `PermissionRequest` / `PermissionDenied` | Programmatic allow/deny with `updatedInput`; `retry` on denial |
| `PostToolUse` → `updatedToolOutput` | Replaces tool output **before the model sees it** — a real lever for **context-efficiency**, provided our own truncation is marked rather than silent *(filed: #129)* |
| `MessageDisplay` | `displayContent` replaces the delta on screen; display-only, stored message untouched |
| `CwdChanged`, `DirectoryAdded`, `WorktreeCreate`/`Remove`, `Setup` | Environment lifecycle; `CLAUDE_ENV_FILE` lets a hook export env into subsequent Bash calls |
| `Elicitation` / `ElicitationResult` | Programmatic response to MCP elicitation |
| `TeammateIdle` | Exit 2 prevents idle — relevant only if teams are adopted |

`SessionStart` also gained since PR #3 was written: `source` now includes **`fork`** alongside
`startup`/`resume`/`clear`/`compact`, and the input carries `agent_type`, `model` and
`session_title`. Its output supports `watchPaths` (register files with the `FileChanged` watcher)
and `reloadSkills` (re-scan skill directories so skills installed by the hook are live in the same
session).

---

## 4. Continuity is not in this plugin

**Decided 2026-07-27.** Earlier drafts of this document carried the memento protocol as Gray Area's
second half. It is split out. **Continuity ships in `prosthetic-conscience`** — which is where PR #3
originally placed it, before the retarget moved it here by association rather than by argument. The
corrected design lives in `plans/context-checkpointing.md`; this section records why it is not here.

**The deciding argument is consent, not architecture.** Gray Area reads transcripts — user text,
file paths, whatever a tool result happened to contain. That is a real surveillance surface, and
the ship's name is the warning. Checkpointing writes a note about your own work. Bundling them
would force a consumer who wants *"don't lose my validation loop after compaction"* to also accept
*"a tool that reads all my transcripts"* — an unnecessary trust decision, charged for a benign
feature. A suite that makes the distinctive claim this one makes about the miner cannot also
smuggle the miner in with the safety belt.

**The practical argument is sequencing.** Continuity has been hand-run across six compaction
boundaries and works. Inside Gray Area it would wait on a Go miner that does not exist. In
`prosthetic-conscience` — which ships today, is already installed, and already carries the
always-on rules continuity exists to protect — it can land next, on its own timescale.

**Why not a fourth plugin of its own:** one skill, two commands and a few thin hooks is a skill, not
a marketplace entry. The suite already layers this correctly — `prosthetic-conscience` is base
discipline, `frank-exchange-of-views` is a specialist engine, `sleeper-service` is the autonomous
loop. Continuity is base-discipline shaped.

**The strongest counter, stated fairly.** The two halves share events (`SubagentStop`,
`SessionStart`, `PreCompact`/`PostCompact`), and verify-on-restore — checking a checkpoint's claims
against reality — is itself a mining operation. But shared events are weak coupling: Claude Code
merges hook configurations from multiple plugins on the same event. And the base discipline is
*verify the claim against reality* (run `git`, check the tag exists), not *parse the transcript*.
Trajectory-backed verification is enrichment the miner adds later, which is a dependency rather
than a merge.

**What Gray Area keeps of it.** Checkpoints become one input among many: a sealed note is a
declared claim about what a session was doing, and act-vs-claim applies to it exactly as it applies
to a seat's attestation. A checkpoint that says *"validation step 2 failing"* against a trajectory
showing step 2 never ran is a finding. That is a consumer relationship, and it is the only one.

---

## 5. Architecture of the mining substrate

Two mechanisms, one available now and one the durable answer. Both were settled on PR #3 and are
recorded here so the plugin does not relitigate them.

**Now — spend an agent to hold the corpus, then query it live.** A subagent reads the trajectory once
into its own context and is queried repeatedly. No build required, keeps a 6 MB corpus out of the
lead's context. Adequate for exploration.

**Later — a persistent index behind an MCP server.** Events keyed by `uuid`, parented by
`parentUuid`, carrying `timestamp`, `agentId`, `attributionAgent`, `effort`, and for tool calls the
name, input and linked `tool_result`. Precedent exists in-repo and its lessons are already paid for:
per-user index location, absolute-path collections config, staleness discipline, and the caching
trap where a stale index answers confidently.

**The distinction that decides the schema:** an agent-as-index is a summarizer — non-deterministic,
unreproducible, uncitable. Fine for "what patterns are in here?". **Disqualifying for a finding**,
because this suite spent an entire cycle removing self-report from the evidence chain, and
"an agent says the transcript shows" reintroduces the defect one layer up, harder to spot because
the summarizer is on our side. Two properties follow for the index:

- **Every answer carries its provenance** — uuid, line offset, file — so a consumer can drop to the
  raw trajectory and check.
- **Staleness fails loudly.** An index answering confidently from stale data is the golden-runner
  failure ("recorded" while the test cache meant nothing was written) with more leverage.

Queries the existing consumers want: every tool call by seat X touching target Y; rework (same tool,
same target, repeated); gaps and stalls between timestamps; user messages with surrounding context;
the steps between two points in the causal chain.

**A measured constraint on the miner:** ~11% of tool calls are invisible to a naive string matcher
because seats alias the binary to a shell variable. It must parse shell structure, not grep for a
name.

---

## 6. Component map

Plugin `gray-area`, depending on `prosthetic-conscience`. Mining only — the continuity components
moved out with §4.

| Component | Kind | Responsibility |
|---|---|---|
| `hooks` → `SubagentStop` | hook | Capture `agent_id`, `agent_type` and `agent_transcript_path` per seat into the run's trajectory manifest, at the moment the seat finishes |
| `hooks` → `SessionEnd` | hook | Close the manifest for a main session; reason-matched |
| `tools/gray-area` | Go CLI | The miner: act-vs-claim, rework, stalls, frustration surface — every answer provenance-stamped, refusing to answer without one |
| `commands/inspect.md` | command | Run an inspection; declared, with what it relied on quoted |
| `commands/trawl.md` | command | Exploration queries over a run — summarizing permitted, findings not |
| `requirements.json` | manifest | `git`; the miner is a Go binary on the same doctor/fetch path as `feov-record` |

**Boundary with `prosthetic-conscience`, stated so neither side drifts across it:** continuity owns
`PreCompact`, `PostCompact`, `SessionStart` and `FileChanged`, and owns the checkpoint schema. Gray
Area owns `SubagentStop` capture and everything downstream of the trajectory manifest. Both may
register on `SessionEnd` — Claude Code merges hook configurations across plugins on one event, so
this is a real division rather than a contested one. Gray Area **reads** checkpoints as declared
claims; it never writes them.

**Capability gating has a second axis: the hook events themselves.** The events this plugin depends
on postdate much of the suite, and a consumer may run a client that never fires them. `/doctor`
should report which hook events the installed client supports, and each hook must be inert rather
than broken where its event is absent.

---

## 7. Phased build

| Phase | Work | Verify |
|---|---|---|
| **0. Hook reality spike** | Register no-op hooks for `SubagentStart`, `SubagentStop` and `SessionEnd` that log full input JSON. Run a `/research` run with subagents. | Logged JSON matches §3. Specifically: `agent_transcript_path` present, readable, and pointing at the seat's own file — not the parent's |
| **1. Capture** | `SubagentStop` writes a per-seat trajectory manifest for a run | A completed `/research` run yields one manifest row per seat with a resolvable transcript path, and no glob of the projects directory anywhere |
| **2. The miner** ✅ | Go CLI over the manifest: act-vs-claim, rework, stalls. Provenance on every answer | ~~Re-derive by machine the two hand analyses from the 2026-07-18 run and reconcile against the hand results~~ — **not runnable, see below.** Substitute: parse a live-generated transcript, resolve the aliased-invocation case, and refuse to emit any row without provenance. **All three shipped 2026-08-18** (#471): `gray-area rework`, `gray-area stalls`, `gray-area pr`. Act-vs-claim over checkpoint notes shipped earlier still, under Phase 5. Spec and what the real data falsified: `plans/gray-area-phase-2-build.md`. NOT shipped and not claimed: the *frustration surface* named in §6's table, and any seat-scoped coverage number — #469 leaves the missed-seat direction unproven |
| **3. Instrumented reasoning** | Launch runs with `--thinking-display summarized`; summaries feed exploration queries only | Summary text is present in seat transcripts **and** the miner refuses to return it on an adjudication query |
| **4. Bench symmetry** | The same inspections aimed at the bench | An inspection of the bench produces the same declared, cited output shape as one aimed at a seat |
| **5. Checkpoints as a mined input** ✅ | Read sealed notes as declared claims and adjudicate them against the trajectory | A checkpoint asserting a validation step that the trajectory shows never ran is reported as a finding, with provenance — **met**, see §10.7 |

**Continuity is not a phase here.** It ships in `prosthetic-conscience` on its own schedule (§4) and
does not gate any phase above. Phase 5 depends on it having shipped, and is the only one that does.

**Phase 2's original acceptance test cannot be run, and this is worth knowing before anyone plans
around it.** It named the 2026-07-18 run's two hand analyses — the attestation audit and friction
mining — as the ground truth to reconcile against. Those analyses read **seat transcripts**, and the
transcripts were never committed. `research/2026-07-18_gray-area-telemetry/trajectories/` holds the
Workflow runtime's journal (25 `started` + 25 `result` records carrying `agentId`, a spawn ledger
with no tool calls) and per-round board metrics. The run record mentions a 25-transcript tarball; it
is not in the repository, and correctly so — the files are large and carry conversation content.

So the historical re-derivation is unavailable until a run is captured *with Phase 1 in place*,
which is the first thing that will preserve the trajectory paths deliberately rather than by
accident. Two consequences:

- The substitute criterion above is weaker: it proves the reader works on real transcripts, not
  that it reproduces a known-good human result. **Do not treat a green Phase 2 as evidence that the
  miner agrees with a human auditor.** That check is deferred, not passed.
- The first `/research` run after Phase 1 ships is worth treating as the real acceptance fixture.
  Capture it, keep the manifest, and reconcile then.

**Validation loop, written before implementation** — the commands that prove the plugin:

1. `go test ./...` in `plugins/gray-area/tools` — clean.
2. Phase-0 hook logs contain every field §3 claims, on the installed client version.
3. Miner re-derivation of the 2026-07-18 attestation audit matches the hand result, or every
   divergence is explained. **Trigger surface:** any change to the trajectory parser or to the
   record schema re-arms this check, and it demands a sibling sweep across all consumers, not an
   instance fix.
4. A forced compaction on a live run resumes on the correct validation step.
5. An adjudication query returns provenance (uuid, file, offset) or returns nothing.

---

## 8. Risks

| # | Risk | Mitigation |
|---|---|---|
| G1 | **Transcript format is vendor-internal and version-unstable.** The vendor steers integrators to `/export` and the Agent SDK instead. | Version-pinned parser, recorded client version per run, graceful degradation. Never silently parse an unrecognized shape |
| G2 | **The JSONL is not tamper-evident.** Append-only is not the same as signed. | State it in every inspection. Gray Area establishes what the record says, not that the record is authentic |
| G3 | **Summaries get promoted to evidence** because they are now available and readable. | Structural refusal in the tool, not a convention (T2 in `reasoning-telemetry.md`) |
| G4 | **The miner becomes a trusted component that can lie silently.** | Provenance on every answer; staleness fails loudly |
| G5 | **Hook surface churn.** Five of the events this design depends on are newer than PR #3. | Phase 0 exists precisely to re-verify against the installed client, and `/doctor` reports the hook events the client actually supports |
| G6 | **Reading trajectories is a surveillance capability.** The ship's name is the warning. Transcripts carry user text, paths, and whatever a tool result contained. | Inspections are declared and scoped; snapshots stay out of git; nothing leaves the box. **This is also why continuity is not bundled here** (§4) — a consumer must be able to take compaction survival without taking the miner |

---

## 9. What this document does not settle

- The relationship between `friction.md` and the miner: friction is unverified self-report and is
  the obvious first mining target, but whether it stays a seat-authored artifact with the miner
  adjudicating it, or becomes miner-derived, is undecided.
- The acceptance test for the plugin as a whole.
- Commit and pull-request metrics on the same substrate — named as in scope by the operator,
  unscoped here. Note that `attributionAgent` is populated for seats and **empty in main sessions**,
  which is the immediate obstacle to attributing a commit to a session.

---

## 10. Phase 5 — checkpoints as a mined input

**Written before implementation.** Phase 5 was gated on continuity shipping; it shipped in #164,
so this is the design and the loop that proves it.

### 10.1 The claim this phase adjudicates

A sealed checkpoint note carries a validation loop, and each entry is a **declared claim about
work that was done**:

```
1. `go test -C plugins/gray-area/tools ./...`  → 3 packages ok  · re-armed by: plugins/gray-area/tools/
   last run: pass 2026-07-30T04:41Z
```

That asserts three separable things: a command was run, it passed, and it was last run at a stated
time. All three are self-report. The trajectory is the only independent record of what a session
actually invoked, so the two can be put against each other — which is the whole of this phase.

### 10.2 Verdicts, and why the negative one is the delicate one

| Verdict | Meaning | Provenance emitted |
|---|---|---|
| `CITED` | a matching invocation is in the trajectory | note file:line **and** transcript uuid+line+timestamp |
| `STALE` | a write under the check's trigger surface happened AFTER the claimed run time | both, plus the citing write |
| `NO-EVIDENCE` | nothing matched | note file:line, the exact tokens searched, and the number of events searched |
| `UNCHECKABLE` | the claim names no command to look for | note file:line, and the reason |

`NO-EVIDENCE` is deliberately **not** called "did not run", and this is the load-bearing decision of
the section. A miner that reports absence as fact is G4 — a trusted component that lies silently —
and the failure is one-directional: a matcher too narrow produces confident false accusations, while
a matcher too broad produces a citation the reader can see is wrong. So the tool states what it
searched for and how much it searched, and lets the reader convict. `UNCHECKABLE` exists for the
same reason: a prose check ("the continuity loop itself") has no command, and guessing at one would
manufacture exactly the false negative this row refuses to produce.

### 10.3 Matching, stated rather than implied

The note's command and the trajectory's command are not the same string, and cannot be made so —
`go test -C x ./...` and `cd x && go test ./... 2>&1 | grep FAIL` are the same check. Matching is
therefore **token containment**: the claimed command reduces to a signature (its binary, its
subcommand, and any path-shaped argument), and a Bash event matches when its command contains all of
them. The signature travels in the output. An approximate match that shows its working is honest;
an exact match that silently misses the piped form is not.

### 10.4 Why `STALE` is the row that pays for this phase

It is the verdict this repo's own history demanded. On 2026-07-30 a note asserted
`last run: PASS 2026-07-30T03:23Z` for the continuity check while the mechanism that would have
re-armed it produced nothing for that check (#165) — and the note went on presenting that pass as
current for the rest of the session. Nothing caught it; it took a hand audit. `STALE` is that audit, mechanically: the
trajectory holds every `Write`/`Edit` with its target path and timestamp, so "a claim older than the
last write to its own trigger surface" is computable without any cooperation from the mechanism that
failed. **It does not depend on `FileChanged` firing**, which is precisely why it is worth having.

> **Correction, 2026-07-30.** An earlier draft of this section called the re-arm mechanism *dead*.
> Withdrawn — see #165. `FileChanged` does fire, including for the session's own edits; what is
> unexplained is that coverage stops advancing partway through a session. The argument for `STALE`
> is unchanged and arguably stronger, since it depends on that coverage not at all: it holds whether
> the coverage is absent, patchy, or perfect. The runbook for settling it is
> `plans/rearm-coverage-experiment.md`.

### 10.5 Scope held back, on purpose

- No scoring, no ranking, no aggregate "trust" number. Rows and provenance only (the §5 line).
- No reading of thinking/summary content — Phase 3 owns that, with its refusal.
- The note parser is **restated, not shared**: `prosthetic-conscience`'s
  `checkpoint.ParseValidationLoop` is canonical, and the plugins are separate Go modules by design.
  The drift risk is real and is handled by failing loudly — a note with a validation-loop heading
  and zero parsed claims is an error, never an empty result.

### 10.6 Validation loop for this phase

1. `go test ./...` in `plugins/gray-area/tools` — clean.  · **re-armed by:** `plugins/gray-area/tools/`
2. A fixture note whose claim IS in the fixture trajectory yields `CITED` with a uuid that appears
   in the transcript.  · **re-armed by:** the note parser or the matcher
3. A fixture note whose claim is NOT in the trajectory yields `NO-EVIDENCE` **and prints the tokens
   searched** — a bare "not found" fails this check.  · **re-armed by:** the matcher
4. The #165 shape: a claim timestamped BEFORE a write under its own trigger surface yields `STALE`,
   citing the write.  · **re-armed by:** the staleness rule
5. A note with a `## Validation loop` heading and no parsable entries EXITS NON-ZERO. Silent success
   on an unrecognized format is the failure mode this phase would otherwise ship.
   · **re-armed by:** the note parser
6. Every row carries provenance or is not emitted — the package contract, unchanged.

### 10.7 Built, and what running it on real data taught

Shipped as `gray-area checkpoint <note.md> <transcript.jsonl>`, plugin `0.3.0 → 0.4.0`. The loop in
§10.6 passes, and the phase's acceptance criterion is met by construction: `STALE` reports exactly
"a checkpoint asserting a validation step the trajectory contradicts", with both documents cited.

**It was run against this repo's own note and this session's own 14 MB transcript** — 1232 citable
events, 9 claims — rather than only against fixtures. That found two defects fixtures had not:

1. **Heredoc bodies were being matched as commands.** A `python3 - <<'PY'` script that merely
   *mentioned* `plugins/prosthetic-conscience/tools` was reported `CITED` as evidence that the tests
   there had run. Two of nine claims were cited to heredocs that only quoted their paths. Fixed by
   stripping heredoc bodies before matching, and the reasoning is recorded in the code: **a false
   citation is worse than a miss.** A miss shows its search and invites a check; a wrong citation
   looks settled.
2. **The evidence line shown was not the line matched on.** A compound command that edited a file
   and then ran the check displayed as `python3 - <<'PY'` — a correct citation presented as an
   obviously wrong one, which costs precisely the trust provenance exists to build.

Both are the same underlying lesson, and it is worth stating because it will recur across every
later phase: **the fixtures could not have found either.** A fixture is written by whoever wrote the
matcher and inherits their idea of what a command looks like. Real transcripts contain heredocs,
compound commands, and pipes nobody thought to fixture. Phase 2's remaining inspections should be
run against a real trajectory before they are believed, not only against tests that pass.

**On the note parser's own near-miss.** The first draft took the command from any line of an entry,
and a real note's prose entry continues "…re-armed by an `add` event" — so it searched the
trajectory for `add` and would have reported `NO-EVIDENCE` against a check it had invented a command
for. Caught by the parser test, and fixed by the same asymmetry: the command comes from the entry's
first line only, because a missing command yields `UNCHECKABLE` (honest) while an invented one
yields an accusation that reads as fact.

**Still deferred:** Phase 2's re-derivation against a human result. Nothing here is evidence that the
miner agrees with a human auditor — only that it reads real transcripts and cites what it finds.

---

## 11. The surface layer — making Phase 5 reachable

**Written before implementation.** Phase 5 shipped a capability nothing can invoke: `gray-area
checkpoint` is a binary, and a session has no way to know it exists or when to run it. This is the
smallest change from *built* to *used*.

### 11.1 What the command layer is, and is not

| plugin | commands | skills | agents |
|---|---|---|---|
| prosthetic-conscience | 5 | 20 | 2 |
| frank-exchange-of-views | 1 | 1 | 3 |
| **gray-area** | **0** | **0** | **0** |
| sleeper-service | 0 | 0 | 0 |

Gray Area is not an outlier for lacking one — sleeper-service does too. The pattern is narrower:
the two plugins with **user-facing capability** ship a command layer, and Gray Area only acquired
user-facing capability with Phase 5.

**It is NOT wired into `prosthetic-conscience`'s seal hook, and must not be.** §4 and G6 are
explicit that continuity cannot depend on the miner: reading trajectories is a surveillance
capability, and a consumer has to be able to take compaction survival without taking it. A
cross-plugin hook would quietly reverse that. The cost is real and is accepted here: a command a
human or a session *chooses* to run is a weaker mechanism than a hook that always fires. Naming the
weakness is the point — this narrows #166, it does not close it.

### 11.2 The blocker, and why the obvious fix is forbidden

The command needs **this session's transcript path**. The manifest only ever holds rows written by
`SubagentStop`, so a main session that spawned no subagents has no manifest at all. The obvious
answer — glob `~/.claude/projects/` for the newest file — is exactly what Phase 1's acceptance
criterion forbids ("*no glob of the projects directory anywhere*"), and §3 explains why: the whole
design principle is that the harness **hands over** the path, deterministically, rather than the
tool guessing which file belongs to whom.

So the enabler is a `SessionStart` hook that records what the harness hands it. Same shape as
`SubagentStop`, one event earlier.

### 11.3 An assumption that is NOT measured, and is built to fail loudly

`hook-surface-spike.md` §3 states every event carries `session_id`, `transcript_path` and `cwd`. That
is a recorded measurement, but it was **not re-measured for `SessionStart` here**, and this suite's
own rule is to treat documentation as unverified until checked. There is no way to fire a
`SessionStart` from inside a running session, so it cannot be closed in the same turn that writes it.

Therefore the hook is built so the assumption's failure is *visible*: a payload with no
`transcript_path` writes **no row** and says so on stderr, rather than writing a row whose path is
empty. A manifest that silently cannot resolve is the failure this whole plugin exists to avoid.
**The first session after this ships confirms or refutes it** — see §11.5.

**Confirmed, 19/19 (§11.7, §11.9).** The assumption holds and this caveat is discharged. What
replaced it is narrower and was not anticipated here: the path is carried but may not yet *exist* at
`SessionStart`, so presence and resolvability are two different checks. §11.9 has the evidence.

### 11.4 Resolution, with provenance

`gray-area checkpoint <note.md>` with no transcript argument reads `.claude/gray-area/` — *our own
manifest directory*, not the projects store — and takes the newest session row. Listing a directory
this plugin writes is categorically different from guessing inside one it does not own.

Which row was used is **printed**, because "the tool picked a transcript for you" is a claim like
any other and gets cited like one.

### 11.5 Validation loop

1. `go test ./...` in `plugins/gray-area/tools` — clean.  · **re-armed by:** `plugins/gray-area/tools/`
2. A `SessionStart` payload carrying `transcript_path` writes exactly one session row, resolvable.
   · **re-armed by:** the capture hook
3. A `SessionStart` payload with **no** `transcript_path` writes **no row** and says so on stderr.
   · **re-armed by:** the capture hook — this is §11.3's alarm, and it is the check that matters
4. `gray-area checkpoint <note>` with no transcript argument resolves from the manifest and PRINTS
   which row it used. With no manifest it says so and exits non-zero rather than guessing.
5. ~~**OPEN, needs a fresh session:**~~ **CONFIRMED 2026-07-30T17:12Z**, re-confirmed **19/19** on
   2026-07-31. See §11.7 — and §11.9 for the separate defect this does *not* cover: the path is
   always carried, but on the first `SessionStart` of a newly minted session id it does not yet
   exist, the row records `resolved: false` permanently, and `checkpoint` then refuses all session.
   · **re-armed by:** the capture hook

### 11.6 Built, and the one check that is still open

Shipped as `/gray-area:audit-checkpoint` plus a `SessionStart` capture mode, plugin `0.4.0 → 0.5.0`.
Loop items 1–4 pass; **item 5 remains open by construction** and is the honest state of §11.3.

Verified end to end by feeding the hook a real `SessionStart` payload and then resolving from what
it wrote — a session row, then `checkpoint` with no transcript argument printing the row it used and
adjudicating 9 claims (5 `CITED`, 3 `STALE`, 1 `UNCHECKABLE`). That exercises every link except the
one that cannot be exercised from inside a running session: **whether the harness actually puts
`transcript_path` in a real `SessionStart` payload.** The hook is wired, so the next session in this
repo answers it — look for a `kind: "session"` row in `.claude/gray-area/`, or for the stderr line
naming the missing field.

**A test that was measuring its own name.** The first draft of the no-manifest test asserted the
error message did not contain "guess", and failed — because `t.TempDir()` embeds the test's name in
the path it returns, and the name ended `...RatherThanGuessing`. The assertion was matching its own
label. It is worth recording next to §10.7's lesson because it is the same family: a check that can
be satisfied or broken by incidental text is not a check. It now asserts on the two things a reader
needs — what is missing, and what to do about it.

**What this does NOT close.** #166 asked for a mechanism, and a command is not one: it fires when a
human or a session chooses to run it. The hook that would always fire belongs in
`prosthetic-conscience`'s seal, and §4/G6 forbid putting it there. So this narrows the gap and
leaves it open, deliberately, rather than buying enforcement with a coupling the design rules out.


### 11.7 §11.3's assumption, CONFIRMED — and one Phase 0 claim that did not survive

Once gray-area's hooks were wired locally, six real `SessionStart` firings each wrote a row:

```
kind='session' resolved=True at=2026-07-30T11:55:14Z … 17:12:35Z
transcript_path='/root/.claude/projects/-home-user-.../<session-id>.jsonl'
```

Every one `resolved: true`, every one carrying a real path. **`SessionStart` does carry
`transcript_path`**, so §11.3's assumption holds and `hook-surface-spike.md` §3 is right for this
event. The stderr alarm never fired, which is the outcome that would have refuted it.

**A wiring fact worth keeping, because it nearly produced a false negative.** The plugin cache is
empty in this environment, so `plugins/*/hooks/hooks.json` **never loads** — hooks fire only because
the gitignored `.claude/settings.local.json` points at locally built binaries, and gray-area's entry
was not in it. The first "end to end" check in §11.6 fed the hook a payload by hand and proved only
that the binary works. An absent manifest would have been evidence about the *wiring*, not about the
harness. `plans/rearm-coverage-experiment.md` states this first, for exactly that reason.

**Phase 0's `agent_transcript_path` claim does NOT hold here.** A `SubagentStop` fired at 17:05:38
and its row reads:

```
kind='seat' resolved=False
capture_error='agent_transcript_path did not resolve: stat …/subagents/agent-a9adb953cea6b9265.jsonl: no such file or directory'
```

`hook-surface-spike.md` §3 records that path resolving to a real 12,997-byte file, and Phase 0's
acceptance criterion demands it be "present, **readable**, and pointing at the seat's own file".
Here it is present and **not readable**. That is one observation and is not yet a general claim —
it may be a timing race (the seat's file written after the hook fires), a cleanup, or a change in
the harness. It is recorded rather than resolved.

Two things follow. The capture hook behaved correctly: it wrote the row with `resolved: false` and
the reason, rather than a row that would silently resolve to nothing — the §11.3 design applied to a
different field. And **Phase 1's acceptance criterion is not currently met**, so any Phase 2
inspection that assumes a readable seat transcript needs this settled first. Added to the runbook.

### 11.8 The `agent_transcript_path` question, SETTLED — and the race hypothesis refuted (#189)

The paragraph above is right that the criterion is not met and right to hold one observation as one
observation. Its guess at the cause was wrong, and the sentence "it may be a timing race" survived
into the runbook as the leading hypothesis. It is refuted.

**Eight seat rows now, eight `resolved: false`.** Every one carries the same `capture_error` shape.
The named directory `…/-home-user-special-circumstances/937047bc…/` exists and holds only
`tool-results/` — **no `subagents/` directory at all**, ~10 hours after the earliest of those rows.
A race that has not resolved in ten hours is not a race.

**The spike was not wrong; the behaviour changed under it.** §3's "real 12,997-byte file" still
exists, byte-for-byte:

```
2026-07-28 11:37  12997  …/-tmp-claude-0--…-scratchpad-spike/937047bc…/subagents/agent-aeaae1e2e57179ff5.jsonl
2026-07-27 19:35  15419  …/-tmp-claude-0--…-scratchpad-td4/937047bc…/subagents/agent-aa9ed822a09ab8138.jsonl
```

Those are the only two on disk anywhere under `/root/.claude`, both predate the wiring, and **neither
id appears in any seat row**. Same session id, different working directory — the project directory is
keyed to cwd, and the spike's cwd was a scratchpad. So the harness *did* write per-seat files on
27–28 July and does not now, for the same session, under the repo's own project directory. What
changed between those dates is unknown.

**And the content is nowhere else.** The parent transcript (5,970 lines, 16.5 MB) holds **zero
entries with `isSidechain: true`**. Each subagent id appears exactly twice — the `Agent` tool_use and
its result. A seat's prompt and its return envelope are recoverable; nothing between them is.

The consequence for the design is narrow and firm: **Phase 0's acceptance criterion cannot be met in
this environment**, and any inspection over a seat's own turns has no input. Inspections over the
parent trajectory are untouched. No fallback path search — a wrong file confidently attributed to a
seat is precisely the false citation §10 exists to refuse.

The methodological point is the one this section already had to learn once: the earlier paragraph did
what it warned against, offering a mechanism ("a timing race") for a single unexplained observation.
The mechanism then travelled into the runbook and became the thing to test. **A guess written beside
an observation is read as part of it.**

### 11.9 Re-run in a fresh session (2026-07-31T03:43–03:50Z): §11.8 is no longer general, and §11.3's assumption holds while its *resolution* does not

The runbook was re-run cold after a `/clear`, which minted session `307f5c51`. Both prior answers
moved, in opposite directions.

**Q1 — `SessionStart` carries `transcript_path`: CONFIRMED, now 19/19.** Every session row ever
written carries a non-empty path: 18 in session `937047bc`, 1 in `307f5c51`. §11.3's assumption is
not in doubt.

**But the path does not always EXIST when the hook fires, and that is new.** All 18 rows in
`937047bc` are `resolved: true`; `307f5c51`'s single row is `resolved: false` — captured
`03:43:31Z`, naming a file that was absent at that instant and present by `03:45:18Z` (331,017 bytes
by `03:50Z`). The difference is not the source: those 18 firings all landed in a session whose
transcript already existed, and this one was the **first `SessionStart` of a newly minted session
id**, where the file is created after the hook runs.

Nothing ever re-resolves a row. So the `resolved: false` is permanent, and `gray-area checkpoint`
refuses for the entire life of that session:

```
gray-area: …/trajectories-307f5c51….jsonl:1 names a transcript that did not resolve (… no such file
or directory) — adjudicating against it would report NO-EVIDENCE for every claim, which is a lie
about the session rather than about the note        EXIT=1
```

**The refusal is correct and the state it refuses on is wrong.** The tool is right not to adjudicate
against a manifest row it cannot trust; the defect is upstream, in a row that records a one-time
`stat` as a permanent property. A capture that re-`stat`s on later firings, or a reader that
re-checks a `resolved: false` path before refusing, would close this. **This is a real race** —
unlike Q1b's, where the file never appears at all — and it is the one case where §11.8's "a race
that has not resolved in ten hours is not a race" does not transfer.

> **FIXED 2026-08-01, in the READER.** Of the two options above, the reader-side one is right and
> the capture-side one is not sufficient: a later firing does write a fresh row, but a session whose
> only `SessionStart` is its first has no later row to write. Re-checking at read time covers both.
>
> `claims.ResolveSession` now takes a `stat` and treats the row's `resolved` as what it is — an
> observation made at capture time, not a property of the transcript:
>
> - a row recorded `resolved: false` whose transcript is readable now is **`Recovered`** and usable,
>   and the command says so rather than proceeding silently against a row that still reads `false`;
> - a row recorded `resolved: true` whose transcript has since vanished is **demoted**, with a reason
>   that distinguishes it from one that never resolved — the same principle in the other direction;
> - selection prefers a row that resolves **now** over a newer one that does not, because the
>   question is which transcript can be read, not which claim about one was written last.
>
> Verified end to end against the binary by reconstructing this exact condition — a row stamped
> `03:43:31Z` with `resolved: false` naming a file that exists — which previously refused with
> `EXIT=1` and now resolves, prints the correction, and adjudicates.
>
> **The capture hook is deliberately unchanged.** It records what the harness handed it and what it
> saw at that instant; that record is accurate and is the evidence this section was written from.
> Making the writer retroactively edit its own observation would destroy exactly the trail that made
> the defect findable.

**Q1b — `agent_transcript_path`: §11.8's general claim is REFUTED. Per-seat transcripts ARE written
in this environment.** A trivial `Explore` subagent produced a seat row that resolved:

```
captured 2026-07-31T03:45:23Z  resolved: true  16,442 bytes  agent_type: Explore
…/-home-user-special-circumstances/307f5c51…/subagents/agent-a9c4e7847f2201f3a.jsonl
```

The file is on disk at exactly the recorded path, with a `.meta.json` beside it, and its first line
is a real seat turn (`"isSidechain":true`, the seat's prompt verbatim). **Phase 0's acceptance
criterion is met here.** §11.8's "cannot be met in this environment" must be read as scoped to
session `937047bc`, not to the environment.

What did *not* change:

- The 10 failures are all `937047bc` (`2026-07-30T17:05:38Z` → `2026-07-31T03:42:51Z`). Every one of
  those paths was re-`stat`ed at `03:46Z` and **all 10 are still absent** — the oldest 10h41m later.
  That session's project directory still holds only `tool-results/`, never a `subagents/`.
- Both parent transcripts still hold **zero** `isSidechain: true` entries. Seat content is still not
  in the parent.
- The two 27–28 July files under scratchpad-keyed project directories are unchanged and still match
  no seat row.

**The correlate is the session, not the clock.** The last failure (`03:42:51`) and the first success
(`03:45:23`) are 2m32s apart, under two different session ids. A harness change inside that window
is not excluded, but nothing observed here supports it, and **the mechanism remains unknown.**
Recorded as an unexplained divergence, not as a cause — the capture hook composes nothing, it reads
`agent_transcript_path` straight from the payload and `stat`s it, so both paths were supplied by the
harness and only one of them was ever written.

The operational consequence: **do not act on §11.8's blanket "not written here".** Whether a seat is
recoverable has to be read off the row, per session. The refusal to add a fallback path search stands
unchanged — it was right for the reason given (a wrong file is a false citation), not because seats
were unrecoverable.

### 11.10 §11.9's correlate is wrong too, and the real one is the SEAT (measured 2026-08-15, #189)

> **SUPERSEDED BY §11.11 (2026-08-18).** Every measurement below holds; the UNIT does not. The
> "typeless seats" are not seats — `SubagentStop` fires at the main agent's turn end. So "blind to
> 72% of seats" counted turn ends in the denominator, and seat coverage is 19/19. Kept unrewritten,
> because a wrong axis that reads as an answer is the thing worth being able to recognise again.

**Third axis, same method as the first two failed on: a property nobody had checked.** §11.8 said
"not written in this environment"; §11.9 narrowed it to "the correlate is the session, not the
clock". Both are wrong — and §11.9's own evidence was compatible with the right answer all along,
because the row it quotes carries `agent_type: Explore` and every failing row it compares against
carries none.

Session `937047bc` — the same session §11.8 and §11.9 both call uniformly broken — now holds **both**
populations, seconds apart:

```
2026-08-15T00:09:54Z  RESOLVED
2026-08-15T00:09:55Z  unresolved
2026-08-15T00:09:57Z  RESOLVED
```

All 69 `kind: "seat"` rows in that manifest:

| | `agent_type` populated | `agent_type` empty |
|---|---|---|
| `resolved: true` | **19** | 0 |
| `resolved: false` | 0 | **50** |

**Zero exceptions.** So the unit is neither the environment nor the session. It is the seat.

**Checked against disk, not against the flag.** The resolved ids were exactly the `agent-*.jsonl`
files present — **0 resolved-but-missing**. One file was on disk whose own row recorded unresolved:
it landed after its `stat`. The write race §11.8 refuted is therefore real, and accounts for **1 of
16, not for the 50.** Both of §11.8's conclusions survive in scoped form: a race exists but explains
almost nothing, and the missing files really are missing.

Every row in both populations carries a non-empty `agent_id` — 69 distinct, 0 duplicates, 0 empty.
The typeless events are not anonymous. They have an identity and no type, pointing at a file nobody
wrote.

**What Phase 0's acceptance criterion now means.** It is met for seats that carry a type and unmet
for seats that do not, in the same session, at the same minute. "Read resolvability off the row" from
§11.9 stands — but **per seat**, not per session, and it can now be read *without* a `stat`:
`agent_type` empty ⟹ no transcript. Capture can classify a seat as uncapturable when the row is
built, instead of emitting a filesystem error that reads like a transient one.

**The refusal to add a fallback path search stands and is reinforced.** With 50 of 69 seats having no
file at all, a fallback would be guessing three times in four.

**What produced the typeless population is NOT determined.** *(ANSWERED in §11.11: they are main
agent turn ends, not a subagent population at all.)* Six probe agents — explicit
`general-purpose`, explicit `claude`, explicit `Explore`, and `subagent_type` omitted — all carried a
type and all resolved. An attempt to identify the others by searching the parent transcript for their
ids was **contaminated and discarded**: printing the ids into diagnostics put them into the
transcript, so the counts measured the analysis rather than the harness. Recorded rather than quietly
dropped — a contaminated count that agrees with the hypothesis is exactly how §11.8 and §11.9 each
got their axis wrong.

Evidence and method: `plans/hook-surface-spike.md` §7a. The same measurement session closed #290's
gate — `PreToolUse` carries `agent_id`, and it joins to `SubagentStop` on the same key (spike §7).

### 11.11 ANSWERED: they are not seats. `SubagentStop` fires at the MAIN agent's turn end (measured 2026-08-18, #189)

**§11.10's table is right and its unit is wrong.** Every row in it is real; the population it calls
"seats with no `agent_type`" is not a population of seats. It is the main agent's own turn boundary,
arriving on the `SubagentStop` hook.

Four measurements on session `937047bc` — 165 `SubagentStop` rows, 19 carrying a type. Method and
data: the manifest, plus the session's own transcript, read in aggregate.

**1 — IDENTITY.** 19 typed rows carry 19 distinct `agent_id`s, and **19 of 19** are present as
`agent-*.jsonl` under `<session>/subagents/`. 146 typeless rows carry 146 distinct `agent_id`s, and
**0 of 146** are present. Every id in both populations fires exactly once. Zero exceptions.

**2 — POSITION.** 139 of the 146 typeless captures (95%) land within 120s **after a main-agent turn
end**. The nearest preceding transcript record is the assistant's closing text (96) or the Stop
hook's own summary (43). Going the other way: **166 of the 197 turn ends** inside the hook's active
window are followed by one — 84%, and between 80% and 90% on every one of the eight days.

**3 — THE FALSIFIER, and it is the one that settles it.** Across **3406** mid-turn windows — an
assistant `tool_use` through its matching `tool_result` — **0 of 146** typeless captures land inside
one. 2 of 19 typed captures do. A subagent completing is *by construction* mid-turn: the parent is
blocked on the `Agent` call, waiting. **Nothing that fires only at turn boundaries is a subagent
finishing.** This is the measurement §11.8, §11.9 and §11.10 each lacked — a prediction the wrong
answer could not survive, rather than a correlation the right one also fits.

**4 — MECHANISM, supporting and not load-bearing.** 20 subagent transcripts are on disk; 19 carry a
`.meta.json` holding `{"agentType": …}`, written for `Agent`-tool spawns. 19 metas, 19 typed rows.
So `agent_type` reads off that sidecar, and an event with no sidecar has no type to report. The three
findings above stand without this one. (The 20th transcript — on disk, no meta — produced **no
manifest row at all**, and is the one on-disk id never named by any row. Left open below.)

**So the harness fires this hook at the MAIN agent's Stop**, with a freshly minted `agent_id`, no
meta sidecar, and an `agent_transcript_path` it *predicts* and nothing ever writes to. The path was
never a broken pointer. It is a forecast, and `stat` was being asked to confirm a file that was never
promised.

#### What this overturns

**"The substrate is blind to 72% of seats" measured the wrong denominator.** 146 turn ends were in
it. Every seat this repository has ever spawned resolved: **19 of 19**. §11.10's operational
advice — "`agent_type` empty ⟹ no transcript, so classify the seat uncapturable" — was a correct
prediction about the wrong object, and acting on it would have hard-coded the miscount into the
tool.

The consequence runs past this file. `plans/gray-area-phase-2.md` scoped Phase 2 around that number:
its §3 table gives seat-scoped inspection "28%" coverage, and Options A, B and C are three ways to
live with it. **None of them is needed.** Phase 2's seat-scoped half is not blocked, and #189 is not
a substrate defect to fix before building — see that file's §0.1.

#### The class, and why three investigations missed it

`misattributed-enforcement`, in the instrument rather than in a message: **the row's `kind` was
decided by which hook delivered it.** `SubagentStop` fired, so the row said `seat` — a
classification with no field behind it and nothing that could refuse it. Every later question was
then asked about seats, and the answer to *"why do 72% of seats have no transcript"* is that 72% of
them are not seats. This is `facts-are-fields` at the point where the fact enters the record: the
event's *name* was read as the event's *nature*.

It survived three rounds because the miscount is invisible from inside the manifest. A reader has
`kind: "seat"` and no way to disagree with it. The disagreement had to come from outside — the
session transcript, which is where the mid-turn falsifier lives.

#### What changed (schema 4)

`gray-area-capture` no longer hard-codes the kind. A `SubagentStop` carrying **no `agent_type` AND
no file at its predicted path** is written as `kind: "turn-end"`. The conjunction, and only the
conjunction: a typed row that did not `stat` stays a seat and stays the alarm; an untyped row that
*did* `stat` stays a seat, because a real trajectory whose name is missing is a surprise, and a
surprise must not be filed under an explanation measured on a different population. Neither has ever
been observed, which is exactly why neither may be folded in silently. Both cases are asserted in
`main_test.go`.

The schema bump is load-bearing: `kind` changed meaning, and a reader counting seats across the
boundary would mix 19 with 165.

#### What this does NOT claim

- **84% is not 100%.** 31 of 197 turn ends produced no row. Windows where the hook binary was absent
  (they are gitignored and rebuilt after every merge) are the known cause and are *not* separately
  measured. Stated rather than rounded away.
- **Why the harness labels this event `SubagentStop`** is not established here — only that it does,
  and that `declared_event` and `hook_event_name` agree, so the wiring is not lying (#462).
- **The one on-disk transcript with no meta and no row** is unexplained. It is a real subagent that
  went entirely unrecorded, and it is the opposite failure from the one this section closes: not a
  phantom counted as a seat, but a seat counted as nothing.

**Contamination, recorded rather than dropped.** Four `agent_id`s were printed into this session's
own transcript while reading `capture_error` text. Small, and this session is itself measured data —
the same class of self-contamination that voided §11.10's identification attempt, caught later this
time rather than not at all.
