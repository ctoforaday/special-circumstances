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
| **2. The miner** | Go CLI over the manifest: act-vs-claim, rework, stalls. Provenance on every answer | ~~Re-derive by machine the two hand analyses from the 2026-07-18 run and reconcile against the hand results~~ — **not runnable, see below.** Substitute: parse a live-generated transcript, resolve the aliased-invocation case, and refuse to emit any row without provenance |
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
re-armed it was dead (#165) — and the note went on presenting that pass as current for the rest of
the session. Nothing caught it; it took a hand audit. `STALE` is that audit, mechanically: the
trajectory holds every `Write`/`Edit` with its target path and timestamp, so "a claim older than the
last write to its own trigger surface" is computable without any cooperation from the mechanism that
failed. **It does not depend on `FileChanged` firing**, which is precisely why it is worth having.

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
5. **OPEN, needs a fresh session:** the wired hook actually produces a session row on a real
   `SessionStart`. Until that is observed, §11.3's assumption is unconfirmed and this row is the
   evidence to look for.

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
