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
summary. Restore stops being blind: a restore hook can compare the checkpoint against what the
summary actually kept and inject **only the delta**. That is the direct fix for the risk PR #3
logged as R4 (duplication with the harness summary confusing the agent) and R3 (digest size).

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
| **2. The miner** | Go CLI over the manifest: act-vs-claim, rework, stalls. Provenance on every answer | Re-derive by machine the two hand analyses (attestation audit, friction mining) from the 2026-07-18 run and reconcile against the hand results |
| **3. Instrumented reasoning** | Launch runs with `--thinking-display summarized`; summaries feed exploration queries only | Summary text is present in seat transcripts **and** the miner refuses to return it on an adjudication query |
| **4. Bench symmetry** | The same inspections aimed at the bench | An inspection of the bench produces the same declared, cited output shape as one aimed at a seat |
| **5. Checkpoints as a mined input** | Read sealed notes as declared claims and adjudicate them against the trajectory | A checkpoint asserting a validation step that the trajectory shows never ran is reported as a finding, with provenance |

**Continuity is not a phase here.** It ships in `prosthetic-conscience` on its own schedule (§4) and
does not gate any phase above. Phase 5 depends on it having shipped, and is the only one that does.

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
