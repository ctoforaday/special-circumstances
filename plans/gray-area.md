# Gray Area — trajectory mining, continuity, and the hook surface

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
reported.** Context survival is one consumer of that substrate, not the point.

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
| Continuity across compaction | done **by hand**, six boundaries, twice reported on PR #3 |

Eight instances of one capability. Four of them are a person doing it manually.

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
| `PostToolUseFailure` | Deterministic strike counting for **anti-spinning** — the 3-strike limit currently lives only in a skill and is enforced by the model against itself |
| `StopFailure` | Matched on error type (`rate_limit`, `overloaded`, `billing_error`, …) — resilience for **sleeper-service**'s unattended loops |
| `TaskCreated` / `TaskCompleted` | Exit 2 prevents creation or completion — the enforcement point for PR #3's improvement **I1**, that a forward actionable must land in the durable queue and not only in the note |
| `InstructionsLoaded` | Observability-only: which `CLAUDE.md` or rule loaded, why (`session_start`, `nested_traversal`, `path_glob_match`, `include`, `compact`), and what triggered it. Answers "did the rules actually bind on this run?" — currently unanswerable |
| `ConfigChange` | Exit 2 blocks a settings or skills change mid-session — an **agent-guardrails** surface |
| `SessionEnd` | Reason-matched (`clear`, `logout`, `prompt_input_exit`, `other`) — the seal point for a session that ends without compacting |
| `PermissionRequest` / `PermissionDenied` | Programmatic allow/deny with `updatedInput`; `retry` on denial |
| `PostToolUse` → `updatedToolOutput` | Replaces tool output **before the model sees it** — a real lever for **context-efficiency** and for the silent-truncation problem |
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

## 4. The memento protocol, rewritten

PR #3's three-part structure — agent-authored note, deterministic seal, deterministic restore — is
correct and survives. What follows amends the mechanism, not the thesis.

| Stage | PR #3 | Corrected |
|---|---|---|
| **Author** | Agent maintains `CHECKPOINT.md` at breakpoints | Unchanged. A hook cannot reflect; the note must be agent-authored |
| **Seal** | `PreCompact` copies the note to a timestamped snapshot | Also: emit compact instructions on stdout naming what the summary must preserve verbatim — validation loop, ordered next actions, in-flight handles, beyond-plan flag |
| **Restore** | `SessionStart(source=compact)` injects a ~1.5 KB digest | `PostCompact` reads the actual summary and injects **only what the summary dropped**. `SessionStart` keeps the `resume`/`fork`/`startup` cases, where no summary exists |
| **Durable index** | Improvement I3: a `MEMORY.md` pointer, because no hook fires on a cold start | Unchanged and still load-bearing. This is the mechanism that actually carried session `6f24a6f4` across six boundaries |

**Fields the two live field reports promoted from candidate to mandatory**, carried as-is:

- **Ordered next actions**, each with a pointer to its home in the canonical queue (issue or plan
  task). A note-only actionable dies when the resumed workflow rebuilds its worklist from a
  different index — this is the one failure a human had to catch.
- **In-flight state handles.** Background work is invisible to the conversation after compaction;
  the checkpoint is the only thread back to it. Validated when a detached 1.28 GB download survived
  a boundary and was re-attached from its task id.
- **Invariants and foot-guns**, verbatim (*"NEVER change model on a resume — cache keys"*).
- **Each validation check's trigger surface** — what re-arms it — now backed by `FileChanged`
  rather than by discipline alone.
- **Not** summaries of completed work. Git history and the run directories already carry that.

**Two disciplines that are not optional:**

- **One block, rotated — never accumulated.** A new seal supersedes and prunes the previous one. The
  second field report caught the checkpoint file outgrowing the harness's auto-recall threshold, at
  which point the deterministic anchor silently degraded to a pointer the agent had to choose to
  follow. Bounded collections, explicit lifecycle.
- **Restore is read-only until the ordered next-actions list.** Post-compaction the harness
  re-presents previously-invoked skills, including ones originally invoked with mutating arguments.
  The interactive harness guards this; a headless restore that naively replays checkpoint content
  would re-run side-effectful steps. Nothing replayed from before the seam is executable.
- **Verify every checkpoint claim against reality before acting on it.** The first field report
  caught a memory-file claim about a tag that had never been pushed. Cheap to check, and the check
  is the whole point of a plugin that reads trajectories rather than trusting reports.

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

Plugin `gray-area`, depending on `prosthetic-conscience`.

| Component | Kind | Responsibility |
|---|---|---|
| `skills/context-checkpointing/` | skill | The authoring discipline: schema, when to seal, the mandatory fields, rotate-don't-accumulate, read-only restore |
| `commands/checkpoint.md` | command | Force a seal now; `--show` prints the current note |
| `commands/resume.md` | command | Print the full note and re-anchor |
| `hooks` → `PreCompact` | hook | Seal the note; emit preserve-verbatim compact instructions on stdout; snapshot and prune |
| `hooks` → `PostCompact` | hook | Diff checkpoint against `compact_summary`; inject only the delta |
| `hooks` → `SessionStart` | hook | Restore for `resume`/`fork`/`startup`; register `watchPaths` for validation trigger surfaces |
| `hooks` → `SubagentStop` | hook | Capture `agent_transcript_path` per seat into the run's trajectory index |
| `hooks` → `FileChanged` | hook | Re-arm a validation check when its trigger surface is touched |
| `tools/gray-area` | Go CLI | The miner: act-vs-claim, rework, stalls, frustration surface — every answer provenance-stamped |
| `commands/inspect.md` | command | Run an inspection; declared, with what it relied on quoted |

Placement question left open: the checkpointing half is core cowork behaviour and has a claim to
living in `prosthetic-conscience`, where PR #3 originally placed it. Deciding factor is whether a
consumer wants continuity without the miner. Not resolved here.

---

## 7. Phased build

| Phase | Work | Verify |
|---|---|---|
| **0. Hook reality spike** | Register no-op hooks for `SubagentStop`, `PreCompact`, `PostCompact`, `SubagentStart`, `FileChanged` that log full input JSON. Force a compaction and a subagent run. | Logged JSON matches §3. Specifically: `agent_transcript_path` present and readable; `compact_summary` non-empty; `PreCompact` stdout demonstrably reaches the summarizer |
| **1. Capture** | `SubagentStop` writes a per-seat trajectory manifest for a run | A completed `/research` run yields one manifest row per seat with a resolvable transcript path, and no glob of the projects directory anywhere |
| **2. The miner** | Go CLI over the manifest: act-vs-claim, rework, stalls. Provenance on every answer | Re-derive by machine the two hand analyses (attestation audit, friction mining) from the 2026-07-18 run and reconcile against the hand results |
| **3. Continuity** | Checkpoint skill, seal and restore hooks, the two commands | A run survives a forced compaction and resumes on the correct validation step; the injected delta contains nothing the summary already carried |
| **4. Instrumented reasoning** | Launch runs with `--thinking-display summarized`; summaries feed exploration queries only | Summary text is present in seat transcripts **and** the miner refuses to return it on an adjudication query |
| **5. Bench symmetry** | The same inspections aimed at the bench | An inspection of the bench produces the same declared, cited output shape as one aimed at a seat |

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
| G6 | **Reading trajectories is a surveillance capability.** The ship's name is the warning. Transcripts carry user text, paths, and whatever a tool result contained. | Inspections are declared and scoped; snapshots stay out of git; nothing leaves the box |

---

## 9. What this document does not settle

- Whether the continuity half ships here or in `prosthetic-conscience` (§6).
- The relationship between `friction.md` and the miner: friction is unverified self-report and is
  the obvious first mining target, but whether it stays a seat-authored artifact with the miner
  adjudicating it, or becomes miner-derived, is undecided.
- The acceptance test for the plugin as a whole.
- Commit and pull-request metrics on the same substrate — named as in scope by the operator,
  unscoped here. Note that `attributionAgent` is populated for seats and **empty in main sessions**,
  which is the immediate obstacle to attributing a commit to a session.
