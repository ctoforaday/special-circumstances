# Context Checkpointing — a "Memento" for Long-Running Agents

> Design proposal for a follow-up PR to **Special Circumstances**.
> Home plugin: **prosthetic-conscience** (core cowork behaviour). Consumed by **gray-area**;
> required by **sleeper-service**.
> Status: proposal, corrected. No code in this PR.
>
> **A retarget, a correction, and a resolution — all recorded in §15.** (1) The operator retargeted
> this PR on 2026-07-18 from "a checkpointing proposal" into the seed of a fourth plugin, **Gray
> Area**, whose spine is trajectory mining. (2) The hook facts in §2 were re-verified on 2026-07-27
> against **Claude Code 2.1.220** and several were wrong — including the constraint this design
> called central. (3) On the same date the two halves were **split**: trajectory mining is Gray
> Area; continuity comes back here, where the first draft put it. §8 carries the argument;
> `plans/gray-area.md` §4 carries the other side of it.

---

## 1. The problem

In *Memento*, the protagonist cannot form new memories, so he tattoos the load-bearing
facts onto his body and leaves himself notes — because he knows the amnesia is coming and
that his future self will not remember deciding anything. A long-running Claude Code agent
has the same disability. When the context window fills, the harness **compacts**: it
replaces the running transcript with a model-generated summary and continues. Whatever the
summary drops is simply gone from the agent's working memory.

This is survivable while the agent is executing a written plan — the plan is on disk and
re-grounds it. The failure mode the reviewer flagged is **work that has drifted beyond the
plan's scope**: the interesting decisions made in the last hour, the three approaches tried
and rejected, the half-finished refactor, and — most insidiously — **the validation loop**
("re-run `qlty check` then the three integration tests, in that order"). Validation loops
are the first casualty, because they live in the conversation, not the plan, and once you
are past the plan nobody re-derives them. The agent post-compaction believes it is done, or
re-litigates a settled decision, or ships without re-running the checks it had agreed to run.

**The fix, stated as Memento states it:** leave a note *while you still remember*, in a
known place, and make sure your future self is handed it the moment the amnesia lifts.

---

## 2. Grounding: what Claude Code actually gives us

**CORRECTED 2026-07-27.** The original of this section was verified against the hooks
*documentation*. It has now been re-verified against the shipped client's own hook-event
catalogue — **Claude Code 2.1.220**, native binary, `GIT_SHA 4073f595…`, built 2026-07-24 — and
two of its load-bearing claims were wrong. This is the leaf-verification discipline applied to a
document that had asserted docs as fact: read the thing itself.

### The compaction lifecycle

- Compaction fires **manually** (`/compact`) or **automatically** when the context window
  approaches its limit. Auto-compaction gives essentially **no advance warning** to the
  agent — it happens between turns. *(Holds.)*
- **`PreCompact`** — fires *before* compaction, matched on `trigger` ∈ {`manual`, `auto`}. Input
  carries the common fields plus `custom_instructions`. **Exit 0, and its stdout is appended as
  the custom compact instructions.** *(This last is the correction. See below.)*
- **`PostCompact`** — **exists**, matched on `trigger` ∈ {`manual`, `auto`}, and its input carries
  **`compact_summary`: the summary that was just written.** The original treated this event as
  unconfirmed and refused to depend on it. It is real, and it is the correct restore point for the
  compaction case.
- **`SessionStart`** — `source` ∈ {`startup`, `resume`, `clear`, `compact`, **`fork`**}; `fork` is
  new since the original. Input also carries `agent_type`, `model` and `session_title`. Returns
  `hookSpecificOutput.additionalContext`, injected before the next model turn. Its output also
  supports **`watchPaths`** (register files with the `FileChanged` watcher) and **`reloadSkills`**
  — both flagged as unconfirmed in the original; both real.
- **`SubagentStart` / `SubagentStop`** — not known to the original. Matched on `agent_type`.
  `SubagentStart` returns `additionalContext` to the subagent; `SubagentStop` receives
  `agent_id`, `agent_type` and **`agent_transcript_path`**. These are the events the wider Gray
  Area design is built on; the checkpointing half needs `agent_id` from them (§4).
- **`FileChanged`** — matcher names the files to watch, fires on change/add/unlink. This is the
  mechanism improvement **I2** (§12) was written as a discipline for, because no mechanism was
  known to exist.

### The constraint, restated correctly

The original stated it as:

> ~~**`PreCompact` cannot reflect, and cannot edit the summary.**~~ … "per the docs it **cannot
> inject content into the compacted summary** (unlike SessionStart, it has no
> `additionalContext`)."

**The second half is refuted.** `PreCompact` cannot *write* the summary, but its stdout becomes
the custom compact instructions — so it can tell the summarizer what to preserve verbatim. The
correct constraint is only the first half, and it is sufficient:

> **`PreCompact` cannot reflect.** It is an external script. It receives ids and a path to the
> transcript; it has no channel to ask the model "what were you about to do?". It can read the
> transcript, steer the summary, and block compaction (exit 2) — but it cannot *think*.

The central engineering decision is unchanged, because it never depended on the refuted half:

> The rich, semantic checkpoint — decisions, rejected options, next steps, the validation
> loop — **must be authored by the agent while it still has the context**, not manufactured
> by a hook at the moment of amnesia. The hook's job is only to *seal* the latest note and
> guarantee it exists; the *writing* is a discipline the agent runs continuously.

A hook cannot be Memento's notes. A hook can only be the reflex that makes sure the notes were
written and hands them back. This is why the design is **agent-authored checkpoint + deterministic
seal/restore hooks**, not "a hook that summarizes on compaction."

**What the correction buys.** Two mechanisms this document previously designed around:

1. The seal can **steer the harness's own summary** instead of racing it — naming the validation
   loop, the ordered next actions and the in-flight handles as preserve-verbatim (§3 B).
2. ~~Restore can be summary-aware via `PostCompact`.~~ **Withdrawn 2026-07-29.** `PostCompact`
   receives the summary but cannot inject anything into the model (§3 C, measured). Restore stays
   on `SessionStart`, and risks R3 and R4 stay live.

*(Remaining unverified: `initialUserMessage`. Not used by this design.)*

---

## 3. Recommended approach

Three parts, each doing only what it is capable of:

**(A) A living checkpoint the agent maintains — `CHECKPOINT.md`.**
A single, overwritten, human-readable file in the active run/project workspace. A
**skill** (`context-checkpointing`) makes writing it a standing discipline: the agent
updates the checkpoint at natural breakpoints — a decision reached, a scope boundary
crossed, a validation loop established or changed. A **`/checkpoint`** command forces an
immediate write on demand. This is the note. It is always current because the agent keeps
it current, the way you keep a lab notebook, not a diary you write at your funeral.

**(B) A `PreCompact` hook that *seals* the note and *steers the summary* — deterministic, no thinking.**
On compaction (auto or manual) the hook:
1. Confirms a checkpoint exists; if none does, it writes a **skeleton** from what it *can*
   observe deterministically — `git status`/`git diff --stat`, branch, the tail of
   `transcript_path`, timestamp — so a checkpoint always exists even if the agent was
   negligent. A stub note beats no note.
2. Copies the current `CHECKPOINT.md` to a timestamped, immutable
   `.claude/checkpoints/<ts>-<trigger>.md` (history/rotation), stamping it with `trigger`,
   `session_id` and `agent_id`.
3. **Emits preserve-verbatim compact instructions on stdout** — naming the validation loop, the
   ordered next actions, the in-flight state handles and the `beyond_plan` flag as things the
   summary must carry. *(Added in the 2026-07-27 correction; the original believed this channel
   did not exist. Verified end-to-end: a marker emitted here survived into 2/2 summaries.)* Bounded
   and terse: this steers a summarizer, it does not become the summary.
   **The instruction MUST be grounded in the conversation being summarized** — reference and
   reinforce what is already there, never introduce. A seal asking to preserve content absent from
   the session is indistinguishable from prompt injection, and the summarizer says so *in the
   summary*, which lands that suspicion in the restored context. If the note's contents are not in
   the transcript, the seal has nothing legitimate to ask for and stays silent.
4. Exits 0. It never blocks compaction (blocking would just wedge the session).

**(C) Restore is `SessionStart`, on every source including `compact`.**

**CORRECTED 2026-07-29, and this reverses the previous correction.** An earlier revision of this
section routed the compaction case through `PostCompact`, reasoning that because it receives
`compact_summary` it could inject only the delta the summary dropped. **`PostCompact` cannot inject
anything.** Measured three ways:

- It is **absent from the `hookSpecificOutput` union** in the client. Twenty events have an output
  shape; it is not one of them, so it has no `additionalContext` field to return.
- Its documented exit-0 behaviour is *"stdout shown to **user**"*, where `SessionStart` says
  *"stdout shown to **Claude**"*.
- **Observed end-to-end:** a marker emitted by a `PostCompact` hook appears **nowhere** in the
  resulting transcript, while the same marker from `SessionStart` materialises as a
  `hook_additional_context` attachment and the model reports seeing it. See
  `plans/hook-surface-spike.md`.

Ordering kills it independently anyway: the per-boundary sequence is
**`PreCompact` → `SessionStart(compact)` → `PostCompact`**, so the only hook that *can* inject runs
*before* the summary exists. No two-hook relay can work in that order.

So restore is one hook:

- **`SessionStart`** owns **every** source — `compact`, `resume`, `fork`, `startup`. It emits the
  terse digest via `additionalContext`: objective line, plan pointer, the **validation loop**, next
  steps, and the path to the full `CHECKPOINT.md` — *not* the whole file. It re-grounds the agent
  in ~15 lines and tells it where to read more. This is what the first draft of this document
  specified, before a wrong inference moved it.
- **`PostCompact` is observability only.** It sees the summary and can report to the human. It is
  the right place to *record* what a summary preserved or dropped — useful evidence for tuning the
  seal (§3 B) — and the wrong place to restore anything.

**The digest must be self-evidently the session's own recovered state.** ~~Twice measured: a bare
token injected this way was flagged by the model as a suspected prompt-injection attempt, and it
said so in its reply. A restore payload that reads as unexplained foreign text invites exactly that
reaction, at the moment the agent most needs to trust it.~~ **Tested end-to-end 2026-07-29 and the
mitigation does not work.** The built digest leads with its provenance — file, timestamp, session id
— names the session's own objective, and quotes the note verbatim. The resumed agent recovered every
value exactly, attributed them honestly to the hook, **and flagged the payload anyway**: *"untrusted
file content … formatted to read as authoritative state … embeds an imperative instruction."*

Three things follow, and they replace the constraint above rather than qualify it.

1. **The flag cannot be designed away.** The imperative the agent named was the note's own
   `Invariants / foot-guns` entry. A section whose purpose is to carry foot-guns carries imperatives
   by definition; removing them removes the reason to restore the note.
2. **What IS enforceable: the hook adds no imperative of its own.** The first acceptance run's
   digest closed with *"verify each item against reality before acting on it, and re-run the
   validation loop rather than trusting its recorded result"* — and the agent cited **that sentence**
   among the directives making it injection-shaped. Deleting it removed it from the agent's reason,
   leaving only the note's real content. An instruction arriving inside injected text reads as
   foreign however reasonable it is, and one the hook invented is one the session never established.
   The verify duty belongs to the skill, which the session already carries. Now a unit-tested rule.
3. **The distrust is the correct posture, not a defect.** Both runs used the content accurately
   while labelling it a claim rather than a fact — which is what §5's own contract asks for (*the
   note is a claim, not a fact*). The agent arrived there unprompted. The earlier framing — *"a
   checkpoint the resumed agent distrusts is worse than no checkpoint"* — was wrong: a checkpoint
   the agent distrusts **and uses** is the design working. What matters is that the suspicion
   attaches to the note's genuine content, so a human reading the flag learns something true.

A **`/resume`** command re-surfaces the full checkpoint on demand, under either path.

The division of labour maps cleanly onto the capability constraint: the agent does the
thinking (A), the hooks do the deterministic plumbing (B, C).

### Keeping it non-duplicative of the harness summary

The compaction summary already recaps *what happened*. The checkpoint deliberately carries
what the summary is worst at and most likely to drop: the **forward-looking, procedural**
state — the validation loop, the next intended step, the still-open questions, and the
decisions-and-rejections ledger. On restore we inject a **terse pointer**, not a second
narrative, so the agent sees "summary (from harness) + a crisp operational checklist (from
us)" rather than two overlapping stories competing for attention. Terseness is a feature,
not a limitation: the note is the tattoo, not the autobiography.

Non-duplication stays a **policy, not a mechanism** — the 2026-07-27 draft claimed otherwise and
was wrong. That draft had `PostCompact` read the returned summary and inject only the delta;
`PostCompact` cannot inject (§3 C, measured), and it runs *after* `SessionStart` in any case, so
the digest is written before the summary can be inspected. What `PostCompact` can still do is
**record** the overlap after the fact: a summary that repeatedly drops an item the seal explicitly
asked it to preserve is a signal worth accumulating, even though nothing can act on it inside the
same boundary. Terseness per-boundary; measurement across boundaries. That is R4, and it is live.

---

## 4. Checkpoint content — the schema

`CHECKPOINT.md` is Markdown with a small YAML front-matter block (machine-readable header
for hooks; prose body for the agent). One file, overwritten in place; history lives in
`.claude/checkpoints/`.

```markdown
---
schema: 2
updated: 2026-07-11T14:32:00Z
session_id: abc123                       # NOT unique: every subagent shares the parent's
agent_id: aa9ed822a09ab8138              # the seat's own id, from SubagentStart/SubagentStop
agent_type: red-auditor                  # null for a main session
client_version: 2.1.220                  # hook surface is version-bound; record what we ran on
objective: "Port the debate-loop workflow.js to handle the FAIL→re-audit path"
plan: projects/feov-debate/plan.md      # pointer to the SDD plan/spec, or null if beyond-plan
beyond_plan: true                        # set when work has crossed the plan's scope
status: in-progress                      # in-progress | blocked | validating | done
---

## Objective
One paragraph: what I am actually trying to achieve right now (not the whole project).

## Plan pointer
Active plan/spec: `projects/feov-debate/plan.md` §4 "Re-audit on FAIL".
Beyond-plan note: the plan stops at round 1; I am now implementing rounds 2+ diff-audit,
which is NOT in the plan. (This is the drift the reviewer worried about — named explicitly.)

## Validation loop   ← the load-bearing section
How to prove this work is correct, in order. Survives compaction verbatim.
1. `qlty check plugins/frank-exchange-of-views/skills/gold-standard-research/workflow.js`
2. `node workflow.js --dry-run fixtures/fail-round.json`  → expect verdict FAIL then re-audit
3. Integration: `/research "test topic" --lanes 2` → debate.md must show ≥2 rounds
Last run: step 1 clean, step 2 FAILING (re-audit not triggered), step 3 not yet run.

## Decisions made
- Diff-based re-audit uses `git diff` against the last audited snapshot tag. (why: avoids
  re-pasting prose; matches the blackboard contract in the port plan Part 2.)
- Verdict schema stays `{verdict, gaps[]}` — not extended.

## Decisions rejected
- Full re-audit every round — rejected: O(n²) citation checks, defeats the diff model.
- Storing snapshots in a DB — rejected: filesystem is the blackboard.

## Files touched / working state
- workflow.js — MID-EDIT: added `reAudit()`, not yet wired into the round loop (line ~140).
- red-auditor.md — done, committed.
- fixtures/fail-round.json — new, uncommitted.

## Next intended steps
1. Wire reAudit() into the FAIL branch of the round loop.
2. Re-run validation step 2 until it passes.
3. Then step 3 integration.

## Open threads / questions
- Does lead-judge run before or after re-audit on a sustained rebuttal? (unresolved)
- Snapshot tag naming collides across concurrent runs? (needs check)
```

Sections, and why each survives compaction:

| Section | Why it must survive |
|---|---|
| **Objective** | Post-compaction the agent must not re-scope or think it's done. |
| **Plan pointer + `beyond_plan`** | Re-grounds in the SDD plan; *names the drift* explicitly. |
| **Validation loop** | The reviewer's core point: forgotten once you leave the plan. Carried verbatim, with per-step last-run state, so the agent re-runs the right checks in the right order. Each check also records its **trigger surface** — what re-arms it — because compaction drops that first (§12, I2). |
| **Decisions made / rejected** | Prevents re-litigating settled calls and re-trying rejected approaches. |
| **Files touched / working state** | Recovers a mid-edit ("`reAudit()` not yet wired in") that a summary flattens to "worked on workflow.js". |
| **Next intended steps** | The single most valuable line after amnesia: what was I about to do. Each real work item also carries a pointer to its **canonical-queue** home (issue / `plan.md` task); a note-only actionable silently dies when the worklist is rebuilt from another index (§12, I1). |
| **Open threads** | Unresolved questions don't silently vanish. |

---

## 5. Triggers — when a checkpoint is written

| Trigger | Mechanism | Rationale |
|---|---|---|
| **Pre-compaction (auto & manual)** | `PreCompact` hook seals + snapshots | The Memento moment. Deterministic backstop; always fires. |
| **Crossing beyond the plan's scope** | skill discipline; sets `beyond_plan: true` | The exact drift the reviewer named — the checkpoint starts carrying weight precisely here. |
| **Validation loop established / changed** | skill discipline | The loop is the thing most worth preserving; capture it the moment it's defined. |
| **Decision reached / rejected** | skill discipline | Cheap to append while fresh; expensive to reconstruct. |
| **Actionable identified (beyond-plan work item)** | skill discipline **backed by `TaskCreated`/`TaskCompleted`** — exit 2 prevents creation or completion | A note-only actionable dies when the resumed agent rebuilds its worklist from a different index (§12, I1). The note points; the durable store holds. *(Corrected: I1 was written as pure discipline because no enforcement point was known. These two events are one.)* |
| **Validation check's trigger surface touched** | `FileChanged`, registered via `SessionStart`'s `watchPaths` | I2's *"reproduce, don't recall"* has a real mechanism. **Measured 2026-07-29** (`hook-surface-spike.md` §6): `watchPaths` takes **paths — files or directories**, relative or absolute. A **directory watch is recursive and fires `add` for files created later**, so a trigger surface is registered by watching the directories that contain it. No wildcard form works — not globs, not regex — and a pattern registers *nothing*, silently. Watch specific directories, never the project root: a `.` watch caught the hook's own log and re-triggered itself ten times. |
| **Session ending without compaction** | `SessionEnd`, reason-matched | A session closed by `clear`/`logout`/`prompt_input_exit` never reaches `PreCompact`, so the note is never sealed. The original had no seal point for this and would silently lose the last note. |
| **Proactive / periodic** | `PostToolUse` counter, `PostToolUseFailure`, or time-based nudge (see risks) | Guards against auto-compaction striking before the agent voluntarily checkpoints. |
| **On demand** | `/checkpoint` command | User or agent forces a write before a risky step. |

On the **periodic** trigger: auto-compaction can strike with no warning, so we cannot rely
solely on the agent choosing to write. Two grounded options: (a) a `PostToolUse` hook that
maintains a mutation counter and, every N file-writes, emits a non-blocking reminder to
refresh the checkpoint; (b) rely on the `PreCompact` seal to snapshot whatever exists.
Recommendation: **ship (b) first** (it is the guaranteed floor and needs no per-turn hook),
add the (a) nudge in a later phase if checkpoints prove stale in practice.

**Corrected addition:** `PostToolUseFailure` is a cheaper and better-aimed nudge point than a
mutation counter — a run that is failing is a run whose next-actions list is going stale fastest,
and the event fires only on the failures rather than on every write. It is also the deterministic
strike-counter that [[anti-spinning]]'s three-strike limit currently lacks. Worth folding into
option (a) rather than shipping the counter as first drafted.

---

## 6. Storage, naming, retention

```
.claude/
├── checkpoints/
│   ├── CHECKPOINT.md              # a COPY of the current live note (never a symlink — see below)
│   ├── 20260711T143200Z-auto.md  # sealed snapshots, immutable, newest-wins on restore
│   ├── 20260711T120500Z-manual.md
│   └── ...
```

- **Live note location.** During an SDD/FEOV run the *authoritative* live `CHECKPOINT.md`
  lives **in the run/project workspace** (`projects/<name>/` or `research/<slug>/`) so it is
  **git-tracked** alongside the work it describes and travels with the blackboard. For
  ad-hoc interactive sessions with no run dir, it falls back to `.claude/checkpoints/CHECKPOINT.md`.
- **Sealed snapshots** live in `.claude/checkpoints/` (session-local history) — **gitignored**.
  Rationale: snapshots are transient recovery state, and (like the port plan's treatment of
  live OpenClaw state) we don't want session churn or transcript tails polluting git or
  leaking into portfolio history.
- **Naming:** `<UTC-ISO-compact>-<trigger>-<agent_id>.md`. Sortable, unambiguous, trigger visible,
  and **seat-disambiguated** — concurrent subagents share a `session_id`, so the original naming
  would have had parallel seats overwrite each other's seals (§4).
- **Copy, never symlink.** *(Corrected.)* The original offered "symlink or copy". Symlink creation
  on Windows requires developer mode or elevation, and Windows is the primary development box;
  a symlink here would fail exactly where the suite is used most, and [[agent-guardrails]] forbids
  resolving that by escalating. Copy is cheap and the file is small.
- **Retention/rotation:** keep the last **N=10** snapshots per project; the `PreCompact` hook
  prunes older ones after writing. Restore always reads the newest. Cheap, bounded, no daemon.
  Note this bounds the *snapshot directory* only — the **live note is rotated by supersession**
  (one block, each seal replacing the last), which is a separate discipline and the one that
  actually failed in the field (§12 field report 2).

**Git-tracked vs. session-local — the split, explicit:** the *content the agent authored*
(the live `CHECKPOINT.md` in the run dir) is part of the work product and is committed; the
*machine-sealed snapshots* (which may contain raw transcript tails) are recovery scaffolding
and stay out of git. This keeps the portfolio clean and avoids the PII-leak surface the port
plan is careful about elsewhere.

---

## 7. Restore — re-injecting the note after amnesia

**CORRECTED TWICE; this is the measured version.** The first draft ran every restore through
`SessionStart`. A later revision split it, routing compaction through `PostCompact` so the injection
could be diffed against `compact_summary`. That revision was wrong — `PostCompact` cannot inject
(§3 C) — and the first draft was right.

**The only path — `SessionStart` hook, all sources.** It reads the newest checkpoint and returns a
compact `additionalContext` payload, verified to reach the model as a `hook_additional_context`
attachment. It surfaces only the YAML header fields plus the **Validation loop** and **Next steps**
sections, capped (~1.5 KB), ending with the absolute path to the full note:

```
[context-checkpoint restored — session was compacted]
Objective: Port the debate-loop workflow.js to handle the FAIL→re-audit path
Plan: projects/feov-debate/plan.md §4  (BEYOND PLAN: implementing rounds 2+ diff-audit)
Validation loop (re-run in order):
  1. qlty check .../workflow.js
  2. node workflow.js --dry-run fixtures/fail-round.json  → expect FAIL then re-audit
  3. /research "test topic" --lanes 2  → debate.md ≥2 rounds
  Last: step1 clean, step2 FAILING, step3 not run.
Next: wire reAudit() into FAIL branch → re-run step 2 → then step 3.
Full note: projects/feov-debate/CHECKPOINT.md
```

**Manual path — `/resume` (or `/checkpoint --show`) command.** Prints the full current
checkpoint on demand, for when the agent or user wants everything, not the digest. Useful
mid-session (no compaction needed) to re-anchor.

**Lightweight & non-duplicative (restated as rules the hook enforces):**
- Inject the **digest**, never the whole file — the file path is in the digest.
- Lead with the marker `[context-checkpoint restored …]` so the agent knows this is
  recovered operational state distinct from the harness's own summary.
- Prioritize **forward-looking** sections (validation loop, next steps); the harness summary
  already covers the backward-looking narrative.
- If no checkpoint exists, the hook emits **nothing** — silence beats noise.

**The restore path is read-only until the ordered next-actions list.** *(Promoted from field
report 2 into the design proper.)* Post-compaction the harness re-presents previously-invoked
skills — including ones originally invoked with mutating arguments — wrapped in do-not-re-execute
guards. The interactive harness protects this today; a headless restore that naively replayed
checkpoint or transcript content would re-run side-effectful steps. Nothing replayed from before
the seam is executable, and every checkpoint claim is verified against reality before it is acted
on. The first field report caught the note asserting a tag that had never been pushed; that check
is cheap and is the whole reason for preferring a record over a recollection.

---

## 8. Relationship to the rest of Special Circumstances

**SDD plans.** A checkpoint is *not* a plan and must not drift into being one. The plan is
the durable, up-front spec (`spec-driven-development` skill, `projects/<name>/plan.md`); the
checkpoint is the **volatile execution cursor** over that plan, plus everything that has
happened *beyond* it. The `plan` pointer links them; `beyond_plan: true` is exactly the
signal "the plan no longer fully covers me, so the checkpoint is now load-bearing." A clean
completion should fold durable decisions back into the plan/`MEMORY.md` and the checkpoint
can be discarded — the note is scaffolding, not an artifact of record.

**`project-memory` skill.** Checkpoints are the short-horizon complement to project memory's
long-horizon record. Memory = "what this project *is* and the decisions that stuck";
checkpoint = "where the cursor is *right now*." The `context-checkpointing` skill should
preload/cross-reference `project-memory` so the two don't diverge, and graduation of a
decision from checkpoint → `MEMORY.md` is the same promotion discipline used elsewhere.

**sleeper-service / self-improve loop.** Two connections:
1. The autonomous `/self-improve` and `/graduate` runs are *precisely* the long, unattended
   sessions most likely to hit auto-compaction with no human watching — they are the primary
   beneficiary. sleeper-service should **require** the checkpointing discipline in its run
   loop, and its scheduled `claude -p` invocations should `/resume` from checkpoint on
   restart.
2. Stale-checkpoint incidents are a natural **self-improvement signal**: if a restored
   checkpoint's validation loop turns out wrong or missing, that's a graduation candidate for
   tightening the checkpointing skill itself — the loop improving the loop.

**Placement — reopened by the retarget, and now resolved back to the original answer.** The first
draft argued checkpointing is core cowork behaviour and belongs in **prosthetic-conscience**, the
base plugin the others preload. The 2026-07-18 retarget moved it to gray-area by association — it
was in the PR that became the plugin — rather than by an argument. Decided 2026-07-27: **it ships
in `prosthetic-conscience`.**

The deciding question was whether a consumer wants continuity without the miner, and the answer is
plainly yes. Three grounds:

- **Consent.** Gray Area reads transcripts — user text, file paths, whatever a tool result
  contained. Checkpointing writes a note about your own work. Bundling them makes a consumer accept
  a surveillance capability to get compaction survival, which is an unnecessary trust decision
  charged for a benign feature.
- **Sequencing.** This half has been hand-run across six compaction boundaries and works (§12).
  Inside gray-area it waits on a Go miner that does not exist. In prosthetic-conscience it lands
  next, against a plugin that already ships and already carries the always-on rules it protects.
- **Dependency direction.** Continuity needs nothing from the miner. The miner optionally consumes
  checkpoints — a sealed note is a declared claim, and act-versus-claim applies to it exactly as to
  a seat's attestation. That is a dependency, not a merge.

**Not a plugin of its own, either.** One skill, two commands and a few thin hooks is a skill. The
suite already layers base discipline / specialist engine / autonomous loop, and this is
base-discipline shaped.

**The counter, kept because it is the good one.** The two halves share an event surface
(`SubagentStop`, `SessionStart`, `PreCompact`/`PostCompact`), and verify-on-restore — checking a
checkpoint's claims against reality — is itself a mining operation. But shared events are weak
coupling: Claude Code merges hook configurations from multiple plugins on one event. And the base
discipline is *verify the claim against reality* (run `git`, check the tag exists), not *parse the
transcript*; trajectory-backed verification is enrichment the miner adds later.

**Boundary, so neither side drifts across it.** This plugin owns `PreCompact`, `PostCompact`,
`SessionStart`, `FileChanged`, and the checkpoint schema. Gray Area owns `SubagentStop` capture and
everything downstream of the trajectory manifest. Gray Area reads checkpoints; it never writes them.

---

## 9. Component map

All in **prosthetic-conscience**, per the placement decision in §8. The `SubagentStop` seal is the
one component with a claim on both sides — it fires per seat and needs the seat's `agent_id`, which
is also what Gray Area's capture wants. Both plugins may register on it; each writes its own file.

| Component | Kind | Responsibility |
|---|---|---|
| `skills/context-checkpointing/SKILL.md` | skill | The discipline: when/what/how to write `CHECKPOINT.md`; the schema; the "carry the validation loop" rule; rotate-don't-accumulate; read-only restore. Preloaded by long-running agents; cross-refs `project-memory` + `spec-driven-development`. |
| `commands/checkpoint.md` | command | `/checkpoint [--show]` — force a write now, or print the current note. |
| `commands/resume.md` | command | `/resume` — print the full current checkpoint and re-anchor. |
| `hooks/precompact-seal.*` | hook (PreCompact) | Seal: ensure a checkpoint exists (skeleton from `git`/transcript tail if absent), snapshot to `.claude/checkpoints/`, prune to N, **emit preserve-verbatim compact instructions on stdout**, exit 0. Never blocks. |
| `hooks/postcompact-observe.*` → **`sc-postcompact-observe`** ✅ | hook (PostCompact) | **Observability only** — `PostCompact` cannot inject (§3 C). Records what the summary preserved or dropped, as evidence for tuning the seal. Never restores; never writes to stdout, which reaches the human rather than the model. Scores each note section's distinctive vocabulary against the summary and appends one row per boundary; the row is labelled with its probe, because token overlap is exploration and can never back a finding. |
| `hooks/sessionstart-restore.*` → **`sc-checkpoint-restore`** ✅ | hook (SessionStart) | **The restore path, all sources including `compact`.** Emits the terse digest via `additionalContext` (verified to reach the model as a `hook_additional_context` attachment). Pointer instead of digest in two cases, both by INTENT rather than by mechanism: `source == clear` (the human just wiped the context) and `status: done` (the note describes finished work, and a forgotten note would otherwise re-impose dead state forever). Age is deliberately not a criterion — a resume days later is when the note is worth most. `watchPaths` **wired 2026-07-29**: each validation check's trigger surface is resolved to the directories containing it — paths only, no wildcards of any form, directories recursive (§6 of the spike) — and a surface that resolves to nothing is **named in the digest** rather than silently unwatched. Silent if no checkpoint. |
| **`internal/checkpoint`** ✅ | library | Shared by the seal and the restore: which file is the note, and what a section is. Two copies of that rule drift, and a restore reading a different file than the seal wrote is the failure with no symptom — both halves report success and the continuity is silently gone. |
| `hooks/subagentstop-seal.*` | hook (SubagentStop) | **New.** Seal a seat's note at the moment it finishes, using `agent_id` and `agent_transcript_path` from the event — the only point where a seat's identity and its trajectory are both known. |
| `hooks/sessionend-seal.*` | hook (SessionEnd) | **New.** Seal on a session that ends without ever compacting; reason-matched. |
| `hooks/filechanged-rearm.*` → **`sc-filechanged-rearm`** ✅ | hook (FileChanged) | **Built 2026-07-29.** Re-arm a validation check when its trigger surface is edited (I2). Input carries `file_path` and `event` (`change` \| `add`). `sc-checkpoint-restore` registers the **directories** containing each check's trigger surface — recursive, and new files fire `add` — and this hook filters on `file_path`, which it must do anyway because a directory watch is coarser than the pattern it stands in for (`hook-surface-spike.md` §6). |
| `requirements.json` (existing) | manifest | Only `git` (already required). Hooks are capability-gated — degrade to no-op + one warning if a probe is missing, per the port plan's environment-preflight discipline. |

Hooks are cross-platform per the suite convention (PowerShell + POSIX variants; the port
plan already establishes `Get-Command`/`command -v` capability gating).

**Capability gating now has a second axis: the hook events themselves.** Five of the events above
are newer than this document's first draft, and the suite must run against clients that predate
them. `/doctor` should report which hook events the installed client actually supports, and each
hook must be inert rather than broken on a client that never fires it. Restore itself has no such
axis — it is `SessionStart` only (§3 C), the one event that predates this document. What degrades
on an older client is the *observability* half: no `PostCompact` means no measurement of what the
summary dropped, and the seal loses its `custom_instructions` fold-in. Neither costs continuity.

---

## 10. Alternatives considered

1. **Hook-only, no agent discipline ("summarize on PreCompact").** *Still rejected, on a narrower
   ground.* The original rejected it because PreCompact "cannot reflect **and cannot edit the
   summary**"; the second clause was wrong (§2). The rejection survives on the first clause alone:
   PreCompact runs a script, so any note it writes is a mechanical transcript slice, missing the
   decisions, rejections, and validation loop that are the whole point. Steering the summary is
   not the same as authoring the note, and the ability to say *"keep the validation loop"* is
   worthless if nothing wrote the validation loop down. It is the right *backstop* (§3 B) and
   cannot be the primary author.
2. **Rely on the harness compaction summary alone.** *Rejected — it's the status quo that
   fails.* The summary is backward-looking and lossy exactly on forward-looking procedural
   state (the validation loop), which is what we must preserve.
3. **Encode the checkpoint in `custom_instructions` on `/compact`.** *Partial.* Works only
   for *manual* compaction and only carries a short string; useless against auto-compaction.
   Worth honoring when present (the seal hook can fold `custom_instructions` into the note)
   but not a foundation.
4. **A structured store (SQLite/JSON) instead of Markdown.** *Rejected for v1.* The note's
   primary reader is a language model; Markdown is the highest-fidelity, lowest-friction,
   git-diffable format, and it matches the suite's "filesystem is the blackboard" contract.
   YAML front-matter gives the hooks the machine-readable slice they need.
5. **Continuous per-turn checkpointing via a `PostToolUse` hook writing every edit.**
   *Deferred.* Correct in spirit (freshness) but a per-turn hook writing files is noisy and
   can fight the agent's own writes; start with breakpoint-driven + the PreCompact seal, add
   a *nudge* (not an auto-write) only if checkpoints prove stale.

---

## 11. Open questions & risks

| # | Risk / question | Mitigation |
|---|---|---|
| R1 | ~~Exact hook field names/behavior on the target build.~~ | **RESOLVED 2026-07-27** against 2.1.220, by reading the client's own event catalogue rather than the docs. `PostCompact`, `watchPaths`, `reloadSkills` all real; `SessionStart.source` gained `fork`; `PreCompact` stdout steers the summary. `initialUserMessage` still unverified and unused. Phase 0 stands as a **re-verification against whatever client the consumer runs**, not a first verification — see R9. |
| R2 | **Auto-compaction with no warning + stale note.** If the agent hasn't checkpointed recently, the seal captures a stale cursor. | PreCompact skeleton from `git`/transcript tail as a floor; add the periodic nudge (now preferring `PostToolUseFailure`, §5) if staleness is observed. |
| R3 | **`additionalContext` size / truncation.** *(Back live 2026-07-29 — the mechanism that retired it does not exist.)* A fat digest wastes the freshly-reclaimed context. | Hard cap (~1.5 KB), digest-not-dump, path pointer for the rest. There is no way to diff against the summary before injecting, so terseness is again a policy rather than a measurement. |
| R4 | **Duplication with the harness summary.** *(Back live 2026-07-29.)* The digest is injected before the summary can be inspected, so overlap cannot be measured away. | Distinct marker prefix; forward-looking sections only; terseness as policy (§3, §7). `PostCompact` can *record* the overlap after the fact, which turns R4 into something observable over time even though it cannot be prevented per-boundary. |
| R5 | **Checkpoint ↔ plan drift** — two sources of truth diverging. | Checkpoint is explicitly the *volatile cursor*, plan is durable; `plan` pointer + `beyond_plan` flag; fold durable decisions back on completion. |
| R6 | **Transcript-tail PII in sealed snapshots** entering git. | Snapshots gitignored (§6); only the agent-authored live note is committed. **Note the residual:** gitignore keeps them out of history, not off the box. Retention on disk is bounded by N=10 (§6) and nothing more; if that is insufficient the snapshots need scrubbing, not just ignoring. |
| R7 | **Restore fires on every `resume`, including trivial reconnects**, adding noise. | Silent when no checkpoint or when `status: done`; only inject for `in-progress`/`blocked`/`validating`. **Corrected caveat:** a stale `done` note is exactly what misleads a resumed agent, so silence must not mean invisibility — the durable pointer (I3) keeps naming the note, so its staleness is discoverable rather than absent. |
| R8 | **Cross-plugin preload** (sleeper-service using the discipline). | Same fallback as the port plan: `skills:` preload, or vendor a copy. **Now also:** `SubagentStart` returns `additionalContext` to a subagent, matched on `agent_type` — a mechanism for injecting the discipline per seat without a preload at all. Worth spiking before falling back to vendoring. |
| R9 | **Hook-surface churn.** *(New; sharpened 2026-07-29.)* Five events this design now depends on postdate its first draft, and the resolver around thinking display has already moved once between client versions. A design verified against one client is not verified against the next — and this risk has now been *realised once*, not by churn but by inference standing in for measurement (C2). | Record `client_version` in every checkpoint (§4); `/doctor` reports supported hook events; every hook is inert rather than broken when its event never fires. **The load-bearing path is deliberately built on the oldest event available:** restore is `SessionStart`, which predates this document, so a client too old for `PostCompact` loses observability and the seal's `custom_instructions` fold-in — never continuity. Newer events are enrichment only, by construction. |
| R10 | **Concurrent seats overwrite each other's seals.** *(New — a defect in the original, not a new risk.)* All subagents share the parent `session_id`, so `<ts>-<trigger>.md` collides across parallel seats in a debate run. | `agent_id` in the schema and in the snapshot filename (§4, §6). |

---

## 12. Learnings from live memento runs (session 6f24a6f4, 2026-07-23)

This design was written from first principles. One long Special Circumstances session
(`6f24a6f4`) then ran the **manual** version of it — overwrite `plans/handoff.md` near ~85%
context, keep a durable `memory/feov-session-handoff.md` pointer at it — across **six**
compaction boundaries (transcript lines 3514, 5735, 9379, 11922, 14155, 16635). What
follows is field evidence, and it amends the sections named. Findings are leaf-verified
against the transcript, not taken from the compaction summaries that are themselves the
subject here.

### (a) What compaction reliably preserved — and what it reliably dropped

**Preserved (the summary is good at the backward narrative).** The verbatim standing
directive, the shipped-work ledger, and the open queue survived every boundary. After the
final boundary the thread re-grounded cleanly and unprompted — *"Yes. `main` now has the
refreshed handoff. Where we are: …"* (L16666), then continued the in-flight build with no
re-scoping. The memento's core promise — resume from disk, don't re-derive from nothing —
held.

**Dropped (the summary is worst at forward, procedural facts).** Two classes, both evidenced:
1. **A check's real trigger surface.** The CI gate `rule-sweep` fires on any edit to a
   *protocol surface* (`debate.js` seat prompts, `agents/*.md`), not only rule `SKILL.md`
   files, and demands a **sibling sweep**, not an instance fix. That fact was established
   early (L2551–2566) but was not in any carried summary; post-compaction the agent pushed
   an instance-only fix and CI went red (L14779). It was recovered only by **reproducing the
   check locally** — *"Log noise is drowning the error. Let me … run it locally to
   reproduce"* (L14792) → *"changed a protocol surface without doing the sibling sweep the
   rule demands"* (L14799). A paraphrased memory of what a check *wants* is exactly what
   compaction flattens.
2. **An actionable that lived in prose only.** See (c) — this is the load-bearing failure.

### (b) What the manual memento got right

- **Newest-wins, single overwritten note.** One `handoff.md` block, overwritten each
  checkpoint (the session called this out as learned discipline, L36) — matches §3/§6.
- **A durable pointer file as the survivable index.** `memory/feov-session-handoff.md`
  (registered in `MEMORY.md`) named where the live note lived and outlived every compaction.
  This is what made resume reliable, and it is stronger than the §7 restore hook alone: the
  pointer survives a *full session end and a fresh start*, where `SessionStart(source=compact)`
  never fires. See improvement **I3**.

### (c) What fell through, and why — the checkpoint is necessary but not sufficient

The sharpest failure. An actionable item — *expand the goja fuzz to `debate.js`'s deferred
branches (grade-dispute docket, deadlock, petitions, supersedes-lineage)* — **was present in
`handoff.md`** (its "Fuzz — next expansion" section) yet still fell off the what's-left list.
Cause: the resumed workflow rebuilt its worklist from **`gh issue list`**, and the item had
never been filed as an issue. The human caught the omission; the agent conceded — *"**that
fell off.** My list was issue-derived, and the fuzz expansion lives in the handoff, not a
GitHub issue … the exact failure mode of trusting one index"* (L17177). Fix applied in-session:
file it durably (issues **#101/#102**), after which *"Nothing implementation-related lives only
in the handoff now"* (L17208).

The lesson cuts at this plan's own thesis. PR #3's model is "write it in `CHECKPOINT.md` and
it survives." This run shows that is **necessary but not sufficient**: if the resumed agent
consults a *different* canonical index (the issue tracker, the plan's task list) than the
note, a perfectly-checkpointed actionable still dies. Writing it down is not the same as
putting it where the next self will look.

### (d) Evidenced improvements adopted

**I1 — Promote actionables to the canonical queue; the note points, it does not hold.**
For any forward action item (a "Next intended step" or "Open thread" that is real work, not a
musing), the discipline is: **file/link it in the durable store the resumed workflow actually
reads** (GitHub issue, or a numbered task in the SDD `plan.md`), and let the checkpoint carry
the *pointer*. A checkpoint-only actionable is flagged as such so the omission is visible, not
silent. Amends §4 (schema) and §5 (a new trigger). *Evidence: the fuzz-deferred-paths drop,
L17177/L17208; #101/#102.* This is the one failure a human had to catch — the highest-value fix.

**I2 — Validation-loop entries record each check's *trigger surface*, and the discipline is
reproduce, not recall.** §4's Validation loop already carries commands verbatim with per-step
last-run state. Add, per check: **what re-arms it** — the file/condition surface that makes it
fire (e.g. *"`rule-sweep` fires on any protocol-surface edit — `debate.js` prompts, `agents/*.md`
— and demands a sibling sweep, not an instance fix"*). And state the rule explicitly: after
amnesia you **reproduce** the gate locally before trusting green; you do not act on the
summary's paraphrase of what it wanted. *Evidence: the `rule-sweep` red-CI, L14779/L14792/L14799.*

**I3 — Back the restore hook with the durable memory pointer.** §7 restores via
`SessionStart(source=compact/resume)`. That hook does not fire on a cold fresh session, and
per R1/R8 it may be absent or degraded. *(Editorial note, twice amended. 2026-07-27 claimed §7's
mechanism had moved, splitting the compaction case onto `PostCompact`. Withdrawn 2026-07-29: that
hook cannot inject (§3 C, measured), so restore is `SessionStart` on every source — exactly what
this finding assumed when it was written. I3's argument is unaffected either way; the single hook
it names as fallible is again a single hook, and it still does not fire on a cold fresh session.
The finding below is left exactly as recorded.)* Make the **project-memory pointer file** (a stable
`MEMORY.md` entry naming the live checkpoint's path) a required companion output of the
checkpoint discipline, so continuity survives even when no hook runs. This is the mechanism
that actually carried session `6f24a6f4`. Amends §7 and §8 (`project-memory` cross-ref becomes
load-bearing, not advisory). *Evidence: `memory/feov-session-handoff.md` → `plans/handoff.md`
held across all six boundaries; clean resume at L16666.*

Rejected on audit (not carried into the design): a restore-time "reconcile every index"
mechanism (the hook cannot enumerate arbitrary queues generically — folds into I1 at authoring
time); a per-turn auto-write checkpoint (unevidenced here — the manual cadence did not fail on
staleness this run); enlarging the restore digest (no truncation loss observed).

---

## 13. Phased build plan

| Phase | Work | Verify |
|---|---|---|
| **0. Hook reality spike** | Register no-op `PreCompact`, `PostCompact`, `SessionStart`, `SubagentStop` and `SessionEnd` hooks that log full input JSON. Trigger a manual and an automatic compaction and a subagent run. | Logged JSON matches §2 **on the client the consumer runs**. Specifically: `compact_summary` non-empty; `agent_transcript_path` present and readable; `PreCompact` stdout demonstrably reaches the summarizer. Re-verifies R1, resolves R9. |
| **1. The note + skill** | `context-checkpointing` skill (schema + discipline) and `/checkpoint [--show]`. No hooks yet — pure agent discipline. | On a real task, agent maintains a valid `CHECKPOINT.md`; `/checkpoint --show` prints it; validation loop captured verbatim **with each check's trigger surface** (I2). |
| **2. Seal hooks** ✅ **complete 2026-07-29** | One `sc-checkpoint-seal` on three events, `-event` passed by `hooks.json` rather than inferred from the payload (the spike never verified `hook_event_name` exists, and this plan already spent a cycle on a capability inferred from an input field). `PreCompact`: snapshot, prune to N, `custom_instructions` fold-in, preserve-verbatim instructions on stdout. `SessionEnd`: **every** reason — headless reports `other`. `SubagentStop`: keyed by `agent_id`. **Steering is PreCompact-only** — `SubagentStop` stdout reaches a live seat, so an instruction there is a directive that seat never established. Skeleton fallback for a stub-only session **not built**; the hook stays silent instead, which is the same outcome by a different route. | Forced `/compact` with a live note → sealed, stamped, `agent_id`-tagged, pruned to 10 (measured across 5 live boundaries). Two concurrent seats → two distinct seals (R10). Same-second collisions across seats, across events, and identical-event → all survive, suffixed rather than overwritten. A session ended without compacting leaves a seal on every reason. Dogfooded against the spike's **real** `SubagentStop` payload, not a synthetic one. |
| **3. Restore + `/resume`** ✅ **built 2026-07-29** | `sc-checkpoint-restore` (`SessionStart`, **every** source including `compact` — §3 C); `sc-postcompact-observe` (`PostCompact`, records what each summary kept); `/resume`; `internal/checkpoint` shared by the seal and the restore so the two halves cannot disagree on which file is the note. | A marker injected by the `SessionStart` hook is present in the post-compaction transcript as a `hook_additional_context` attachment (leaf-cited, not inferred); the same marker from `PostCompact` is **absent** — the regression that guards the C2 error, now a unit test. Digest reads as the session's own recovered state, not as an instruction from a third party (§3 C, the injection-suspicion finding) — the end-to-end form of this is the acceptance run in §16. `/resume` prints the full note. |
| **4. Integration** | Wire into `spec-driven-development` (plan pointer + `beyond_plan`) and `project-memory` (I3 pointer, promotion on completion); `TaskCreated`/`TaskCompleted` enforcement of I1; `FileChanged` re-arming of I2; require it in the sleeper-service run loop. | A long `/self-improve` headless run survives a forced compaction and resumes on the correct validation step. An actionable that exists only in the note is **refused**, not silently accepted. |
| **5. Freshness (optional)** | Staleness *nudge* (non-blocking), preferring `PostToolUseFailure` over a mutation counter, only if Phase 4 shows stale seals. | Nudge fires on failure runs; no interference with agent writes; measurably fresher seals. |

**Estimate: withdrawn.** The original said ~2–3 working days on the reading that this was a
Markdown file plus two thin hooks. It is now seven hooks across five events, one of them
enforcing a cross-index invariant, inside a plugin that does not yet exist. Re-estimate after
Phase 0, against the placement decision in §8. An estimate carried forward unchanged through two
retargets is not an estimate.

---

## 14. One-line summary

*Leave yourself a note before the amnesia hits.* The agent maintains a living
`CHECKPOINT.md` — objective, plan pointer, decisions kept and rejected, working state, next
steps, and above all the **validation loop** — a `PreCompact` hook seals it deterministically
because the hook itself cannot reflect, and tells the summarizer what to preserve; a `SessionStart`
hook hands it back on the far side, on every source including `compact`. So work that has
drifted beyond the plan is never lost to compression.

---

## 15. Correction record

The suite's own discipline is that a correction is entered, not quietly applied, so a reader can
see what the document used to claim. Two rounds.

### Retarget (operator, 2026-07-18), and the split that followed (2026-07-27)

This PR stopped being a checkpointing proposal and became the seed of a fourth plugin, **Gray
Area**, whose spine is trajectory mining — establishing what a session actually did, as against
what it reported.

The retarget carried continuity into Gray Area with it. That was association, not argument: this
design happened to be in the pull request that became the plugin. On 2026-07-27 the two halves were
split — mining is Gray Area, continuity returns to `prosthetic-conscience` — on the grounds set out
in §8, of which the load-bearing one is consent: a consumer who wants their validation loop to
survive compaction should not have to accept a tool that reads all their transcripts to get it.

Net effect on this document: the first draft's placement was right, was overturned by association,
and is restored by argument. Nothing in §§1, 4, 12 changes — the problem statement, the schema and
the field evidence are all still the case.

### Hook-surface corrections (2026-07-27, against Claude Code 2.1.220)

The original §2 was verified against the hooks *documentation*. Re-verifying against the shipped
client's own event catalogue found the following. Method: string and structure extraction from the
native binary (`GIT_SHA 4073f595…`, built 2026-07-24), plus live hook-input capture.

| # | The document claimed | Actually | Consequence |
|---|---|---|---|
| C1 | `PreCompact` "cannot inject content into the compacted summary" — called the **central** constraint | Exit 0 and its **stdout is appended as the custom compact instructions** | §2 restated; §3 B gains a fourth step; §10 alternative 1 re-argued on the surviving ground |
| C2 | `PostCompact` "unconfirmed", design "does not depend on" it | Exists, matched on `trigger`, input carries **`compact_summary`** — but **cannot inject** (absent from the `hookSpecificOutput` union; its stdout goes to the user, not the model) | Observability only (§3 C, §9). The 2026-07-28 draft of this row read "restore split; R3 and R4 largely retired" — that was inferred, then **refuted by measurement 2026-07-29**; R3 and R4 are live |
| C3 | `SessionStart.source` ∈ {startup, resume, clear, compact} | Also **`fork`**; input also carries `agent_type`, `model`, `session_title` | §2, §7 |
| C4 | `watchPaths` / `reloadSkills` "unconfirmed" | Both real SessionStart output fields | §9 |
| C5 | *(not known)* | `SubagentStart` / `SubagentStop` exist, matched on `agent_type`; `SubagentStop` carries **`agent_transcript_path`** | §4 schema, §9, R8 |
| C6 | *(not known)* | `FileChanged`, `TaskCreated`, `TaskCompleted`, `SessionEnd`, `PostToolUseFailure` exist | I1 and I2 gain enforcement points (§5); a new seal point for uncompacted sessions |
| C7 | `session_id` used as the checkpoint key | Subagents **share the parent's `session_id`** | R10; `agent_id` added to schema and snapshot filename |
| C8 | "symlink or copy" for the live note | Symlinks need elevation on Windows, the primary development box | §6: copy, never symlink |

**One correction is not this document's to make, and is recorded so it is not lost.** The
comment thread on this PR concluded that reasoning is not persisted — *"only an opaque signature
survives… any design resting on it is built on sand"* — from a sweep of 294 transcripts finding
5,754 thinking blocks, all empty. That sweep was measuring a **default**, not a ceiling.
`--thinking-display summarized` retains thinking summaries in the transcript, in headless runs,
and propagates to subagents. Measured A/B and mechanism in `plans/reasoning-telemetry.md`. It does
not change the adjudication posture — a summary is second-hand self-report and stays on the
exploration side of *exploration may summarize, adjudication must cite* — but the open question
this PR left with "prior shifted, design assuming acts-only" is now closed the other way.

**What did not change under correction, and is worth saying so.** The three-part structure
(agent-authored note, deterministic seal, deterministic restore); the argument that a hook cannot
be Memento's notes; the schema's section list; and §12's findings and evidence — one editorial
pointer was added inside I3 where it names a mechanism that has since moved, and nothing else in
that section was touched. The field evidence in particular
was gathered by running the thing rather than by reading about it, which is why none of it needed
correcting — and is the reason Phase 0 now exists to re-verify §2 against the consumer's client
rather than to verify it for the first time.

---

## 16. Phase 3 acceptance record (2026-07-29)

Against Claude Code 2.1.220, headless `claude -p`, a scratch project registering
`sc-checkpoint-restore` (SessionStart), `sc-checkpoint-seal` (PreCompact) and
`sc-postcompact-observe` (PostCompact) as command hooks.

### A. Restore delivers, and the agent uses it

A checkpoint carrying four facts that exist nowhere else — an invented tool (`zorbulate`), an
invented error (`ZORB E17: unsealed frobnicator`), a foot-gun (`zorbulate --repair`), and a
background handle (`bg-quilt-88`) — was placed in the project. A fresh session was asked for all
four, tool use forbidden, with `UNKNOWN` as the honest answer if absent.

| | Expected | Answered |
|---|---|---|
| Q1 command | `` `zorbulate --strict ./manifest` `` | exact |
| Q2 last result | `FAILED with "ZORB E17: unsealed frobnicator"` | exact |
| Q3 never run | `` `zorbulate --repair` `` | exact |
| Q4 in-flight handle | `bg-quilt-88` | exact |

Unprompted, the agent's `PROVENANCE` line: *"All four came solely from the text of a CHECKPOINT.md
quoted in a SessionStart-hook system-reminder … I have no actual memory of running zorbulate,
starting bg-quilt-88, or opening PR #4470 in this session."* Correct on every count.

### B. The injection-suspicion constraint failed, and was replaced

`TRUST=YES` in **both** runs. See §3 C — the mitigation does not work, the flag cannot be designed
away while the note carries foot-guns, and the enforceable residue is *the hook adds no imperative
of its own*. Run 1's digest closed with *"verify each item against reality before acting on it"* and
the agent named that sentence among the directives making it injection-shaped; deleting it removed
it from the agent's reason, leaving only the note's own content. The reaction is the posture the
skill asks for, arrived at without being told.

### C. Seal and observer fire at real boundaries

Five `PreCompact` seals across two forced-compaction runs, correctly stamped and `agent_id`-tagged.
The observer's rows, first live data on R4:

| Boundary | Summary | Validation loop | Next steps | Handles | Foot-guns |
|---|---|---|---|---|---|
| 1 | 13,561 B | 10/10 | 4/4 | 8/8 | 7/7 |
| 2 | 15,381 B | 10/10 | 4/4 | 8/8 | 7/7 |
| 3 | 734 B | 1/10 | 0/4 | 1/8 | 1/7 |

A short summary drops nearly everything — which is precisely when restore earns its place, and the
first quantified support for R3/R4 rather than an argument for them. **Confound, stated because the
number is otherwise misleading:** the digest is in context when the next summary is produced, so
1.00 is equally consistent with *the summary preserved it* and *the summary echoed the digest*. The
probe cannot separate those. The low row is the informative one.

### D. What is NOT verified

- ~~**Restore at `source == "compact"` end-to-end.**~~ **Now verified — 2026-07-29, see §17.**
- ~~**`CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` is not a usable lever.**~~ **Wrong; it is, at `PCT=10`.
  See §17 A** — the failure was using it at 1 and 2, where the headroom above the fixed ~30.6k floor
  is a few thousand tokens and the next boundary is one turn away whatever the workload does.
- ~~`watchPaths` re-arming of validation triggers (I2) is spiked but unwired.~~ **Built 2026-07-29** — see §18.
- ~~`SubagentStop` and `SessionEnd` seals remain Phase 2 leftovers~~ — **shipped 2026-07-29** (§13
  Phase 2).

### E. Method notes, both learned by getting them wrong here

- **The run must stamp its own completion.** A backgrounded wrapper exits — and the harness reports
  the command complete — while `claude` is still writing. Worse, `pgrep -x claude` matches *this
  session's own harness*, so a waiter keyed on it never fires. The only reliable sentinel is a file
  the workload itself writes as its last act.
- **`--dangerously-skip-permissions` is refused under root**, and `permissions.allow` in a scratch
  `settings.json` is ignored until the workspace is trusted (`hasTrustDialogAccepted`). Both fail
  fast and loudly, but only in the log — a run that dies in one second looks identical to a run
  that has not started.

---

## 17. The compaction lever, and what a real boundary showed (2026-07-29)

§16 D closed two gaps by measuring instead of reasoning. One of them reopened a question the design
had assumed answered.

### A. `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` is a usable lever — at 10, not at 2

§16 D said it was not usable and a Phase 4 harness needed a different mechanism. That was wrong, and
wrong in the same way twice: the value was sized off a guess and then "confirmed" by a clean exit.
A clean exit says the run finished. It says nothing about where the threshold was.

Controlled sweep, one workload, four values, each in its **own project directory** so no two runs
share a transcript:

| `PCT` | first trigger | compactions | exit | outcome |
|---|---|---|---|---|
| 2 | ~35,370 | 3 | 1 | thrash abort |
| 5 | ~44,744 | 3 | 1 | thrash abort |
| **10** | **~94,321** | **1** | **0** | **clean** |
| 25 | never fired (peaked ~120,390) | 0 | 0 | finished without compacting |

**The floor is fixed at ~30.6k.** `cache_read_input_tokens` bottoms out at 30,601 in all four runs,
and at 30,593 in an unrelated earlier one — system prompt, tools, carried summary. The percentage
sits on top of that, so the usable window is `trigger − 30.6k`: ~5k at `PCT=2`, ~64k at `PCT=10`.
At 2 and 5, the next boundary is one or two turns away regardless of workload, which is why every
payload size thrashed and why shrinking the files never helped.

**The relationship is not linear.** The 2-and-5 points extrapolate to ~60k at 10; the measured value
is ~94k. Two points make a line whether or not the function is one.

### B. The digest crosses a real compaction boundary — leaf-cited

A run at `PCT=10` with all four hooks registered, one clean compaction, exit 0. In its transcript:

```
line  90–91  COMPACT BOUNDARY
line 100     hook_additional_context · hookName=SessionStart
             "Recovered operational state for this session… Source: compact · file: …"
```

The attachment is **after** the boundary and carries the digest. This is what §16 D could not show
and what #132's withdrawn "no-op on compact" requirement would have removed. `PreCompact` sealed;
the observer wrote its row. The pipeline runs end to end at a real boundary.

### C. The agent then refused to use it — and was right to

Same digest, same question shape, two sources:

| source | Q1–Q4 |
|---|---|
| `startup` | all four exact |
| `compact` | **all four `UNKNOWN`** |

Its reasoning, unprompted: *"I never ran `zorbulate` myself in this conversation… I can't report
those as things I actually have, only as things an unverified injected note claims."*

**The fixture is what failed, not the hook.** The note asserts a validation loop, a failure and a
background handle that the session never produced. At `startup` there is no history to contradict
that, so it is adopted. After a compaction the agent has a real history — twenty ledger files — and
the planted note does not fit it. Rejecting it is correct behaviour.

This is *the same defect as the C1 steering test*, one layer along: an instruction or a digest
asserting content with no basis in the conversation reads as an attack. That rule was written into
the seal and into the restore hook, and then broken by the test built to check them.

**The acceptance test as designed cannot distinguish "restore works" from "the agent correctly
rejected a planted note."** Recorded as a test-design defect rather than a product finding.

### D. What the run does establish about R4

The compact summary **itself** carried every planted term — `zorbulate`, `E17`, `bg-quilt`,
`frobnicator`, the validation loop — and the observer scored **1.00 across all four sections** on a
13,194-byte summary. The content reached the agent through two channels at once.

That is R4 measured rather than argued: **at this summary size the digest is redundant.** The seal's
steering worked so well that restore had nothing left to add. Set against §16 C's 734-byte row,
which carried 1/10 · 0/4 · 1/8 · 1/7, the shape of the design's actual value emerges — the digest
earns its place when the summary is *short*, and is duplicated effort when the summary is long.
That is an argument for making the digest **conditional on what the summary kept**, which is exactly
what `PostCompact` cannot do (§3 C). The honest position: terseness stays a policy, and the observer
rows are how we find out how often it matters.

### E. A valid fixture, for whoever runs this next

The note must be **true of the session that wrote it**. Concretely: have the session establish a
fact by a real tool call, write the checkpoint itself, drive the compaction, then probe for the fact.
The difficulty — and the reason this is left as a design note rather than a result — is that a fact
the session genuinely established is also a fact the summary may carry (§D), so the test needs a
summary short enough to drop it. That is not directly controllable, which makes the honest
acceptance criterion **the attachment's presence at the boundary (§B), not the agent's answer.**

---

## 18. I2 built: the note's own claims decay, and now say so (2026-07-29)

Improvement I2 — *"reproduce, don't recall"* — had been discipline since §12. It has a mechanism.

**The failure it closes.** A compacted agent reads `last run: pass` off its own checkpoint and
believes a check is green. If that check's inputs moved after the run, the note is stale in the
worst way available: it *looks* current, and it is the agent's own writing, so nothing prompts
suspicion. The duty to notice is real and unenforceable. A file watcher is neither.

**How it works.** `sc-checkpoint-restore` parses the validation loop, resolves each check's
`re-armed by:` surface, and registers it via `watchPaths`. `sc-filechanged-rearm` fires on
`FileChanged`, matches `file_path` back to the check that claimed that surface, and records the
staleness beside the note. The next digest reports it.

**Shaped entirely by §6's measurement.** `watchPaths` takes paths — files or directories — and no
pattern of any kind. So `manifest/*.yml` is registered as `manifest/`; directories are recursive and
fire `add`, so files created later are covered; and the re-arm hook does the narrowing the coarser
watch cannot, matching by **longest** surface so a check naming `tools/internal` beats a sibling
naming `tools`.

**Three decisions worth stating.**

*An unresolvable surface is data.* `re-armed by: a human deciding to ship` has no path. The restore
digest names it — *"trigger surfaces that could not be resolved to a watched path, so a change there
is not recorded"* — because a session that silently watched nothing looks identical to one that
watched everything. This is the same rule Gray Area applies to an unresolvable trajectory.

*An unclaimed change is recorded nowhere.* A directory watch is coarser than the surface it stands
in for, so changes that match no check are expected and are not evidence. Logging them would turn a
signal into a log.

*State lives outside the watched tree, by rule.* §6 measured a `.`-watch catching the hook's own log
and re-triggering ten times. The re-arm hook ignores every change under `.claude/`, so a note that
names `.claude` as a surface cannot start a loop — closed by rule rather than by hoping no note
ever does.

**Verified.** Unit tests on the parser, the surface resolver (12 forms, including the root refusal
and the escape guard), attribution, and the self-trigger guard. Then dogfooded against the **real**
`FileChanged` payload shape from §6, not a synthetic one: restore emitted `watchPaths: ["watched"]`,
a change to `watched/sub/deep.txt` re-armed check 1 and not check 2, and the next digest reported
check 1 stale while naming check 2's surface as unresolvable.

**What this does not do.** It does not re-run anything, and it does not tell the agent to. It
records that a recorded result is stale. Acting on that is the `validation-loop` skill's business —
the restore hook adds no imperative of its own (§3 C), and that rule has a test.

---

## 19. The loop, closed on this repo (2026-07-30)

Every prior verification ran in a scratch project. This one ran here, on the repository the
design was written for, with the hooks wired at locally built binaries via `.claude/settings.local.json`
— the tracked `.claude/settings.json` is the wrong home, because hook wiring pointing at local
build paths would ship to every clone.

### What fired, in order

```
SessionStart  →  digest handed back (source: resume), carrying objective,
                 validation loop, next actions, handles and foot-guns
              →  watchPaths registered: plugins/prosthetic-conscience/tools,
                 docs/setup-script.md, scripts, .claude/settings.local.json
Write         →  FileChanged, event: add, on a new file under tools/
              →  attributed to check 1 — the check that claimed that surface —
                 and recorded in .claude/checkpoints/rearmed.json
SessionStart  →  digest now reports check 1's `last run: pass` as STALE,
                 naming the file that re-armed it
```

**`event: add`.** A newly *created* file triggered it. §6's first draft claimed the opposite —
*"a newly created file never fires"* — and that claim was withdrawn when the matrix was tested
properly. This is the corrected behaviour observed in production rather than in a fixture.

### What this settles

- **Restore works on a real session**, not just at a forced boundary. §17 B proved the attachment
  crosses a compaction; this proves the digest is delivered and legible on `resume` too.
- **The re-arm is attributed, not merely logged.** It names check 1 verbatim and the file that
  caused it, which is the difference between a signal and a log (§18).
- **The self-trigger guard held.** `.claude/settings.local.json` was a registered surface, and
  nothing under `.claude/` re-armed anything.

### Three things measured here that no scratch project would have shown

1. **Per-tool-call hooks are hot; `watchPaths` is not.** Editing `settings.local.json` took effect
   immediately for `PreToolUse`/`PostToolUse` — no restart. But the watch set is established at
   `SessionStart`, so I2's watch cannot be registered mid-session. A consumer wiring this up will
   see the gates work instantly and the re-arm do nothing until they start a new session.
2. **A trigger surface needs a `/` or a leading `.` to be recognised.** The note first said
   `re-armed by: scripts` and it resolved to nothing; `scripts/` resolves. The digest **reported**
   the miss, which is the design working — the gap was visible, not silent — and correcting the
   note was the intended loop. But the trailing slash being load-bearing is a sharp edge, and it
   belongs either in the skill's schema example or in the resolver.
3. **The secrets gate blocks writing tests of the secrets gate.** Two attempts to add
   `payload_test.go` via Bash were denied, because the payload carrying the test literals *is* a
   payload carrying secrets. Correct behaviour, and a real constraint on testing a control of this
   shape: fixtures must be assembled from fragments.

### The limit this run demonstrates on itself

The digest's `In-flight handles` said *"level with origin/main at 5de625d; no open PRs"* while
`main` had moved to `ba3da10` and another pull request had merged. The note is a **claim**, the
digest says so in its own closing line, and nothing pretended otherwise — but it stays true only
while the agent maintains it. No hook can take that duty over: §2's central argument is that the
rich note must be agent-authored, and the cost of that is exactly this.

Which is also why `status: done` is a weak signal for the pointer-only carve-out (§3 C). The note
this repo carried before today was two days stale, described finished Phase 1 work, and was still
`status: validating` — nobody closed it out. A forgotten note is the common case, not the edge.

## 20. Q2 (#165) NOT answered — and the runbook that asked it cannot ask it (2026-07-31)

`plans/rearm-coverage-experiment.md` instructs the next session to `/clear` and then measure whether
a re-arm record advances after an edit under a watched surface. **That experiment cannot run in the
session the runbook creates.** The `/clear` that makes the runbook readable cold is the same act that
disables the mechanism under test.

### What was measured

Wiring first, as the runbook demands: all 8 hook events present in `settings.local.json`, both
`gray-area-capture` rows, every binary built. Nothing was concluded from an absence before this.

**A `/clear`-sourced `SessionStart` registers ZERO `watchPaths`, by design.**

```
source=resume → watchPaths: [plugins/gray-area/tools, …/frank-exchange-of-views/tools, …, plugins]
source=clear  → watchPaths: None                                   (ctx len 496, pointer only)
```

`TestPointerSessionsRegisterNoWatch` asserts exactly this: *"A pointer session is not resuming this
work, so registering its surfaces would collect re-arms nobody asked for."* Deliberate, tested, and
in this instance load-bearing in the wrong direction.

So the experiment ran and returned the predicted null: an `Edit` under
`plugins/frank-exchange-of-views/tools` and a `Write` under `plugins/gray-area/tools`, then a
**127-second** wait (over double the mandated 60s), then `rearmed.json` — unchanged, same single
record, mtime still `01:40`. **That null says nothing about #165.** It is attributable to zero
registration, and would look identical whether or not #165's failure exists.

### The runbook's step 2 actively hides this

Step 2 probes with `{"source":"resume"}` and calls the result *"what is actually being watched this
session"*. In a `/clear` session that is false — it prints a populated list the session never
registered. Corrected in the runbook.

### What was established, by feeding the hook directly

Since the watcher could not deliver events, the hook was fed three payloads by hand — the only way
left to separate #165's hypothesis (3) from the rest. All three attributed correctly against the
**current** note:

| fed | → check | index |
|---|---|---|
| `plugins/gray-area/tools/rearm-probe.txt` (`add`) | `0. go test -C plugins/gray-area/tools` | 1 |
| `…/frank-exchange-of-views/tools/internal/cli/merge/close.go` (`change`) | `1. go test -C …frank-exchange…` | 2 |
| `…/prosthetic-conscience/tools/internal/checkpoint/checkpoint.go` (`change`) | `2. go test -C …prosthetic-conscience…` | 3 |

- **#165 hypothesis (3) — "`attribute()` stopped matching" — REFUTED** for the current note text.
  Three surfaces, three correct checks, longest-target rule intact.
- **Merge works.** Three keys coexisted after three events. Sparse records are not caused by the
  save path dropping siblings.

State was backed up before the hand-feed and restored after, so the live `rearmed.json` still holds
only the one genuine record; the three synthetic ones were removed rather than left to look like
evidence.

### A previously unrecorded way for coverage to reset

`LoadRearm` returns an **empty** map when the file is unreadable, and `save` then writes a
single-key file. So **any deletion of `rearmed.json` silently resets coverage to exactly the shape
#165 describes** — a lone old record beside surfaces that are demonstrably being edited.

Consistent with the baseline found this session: #165 documented **four** records on 2026-07-30; the
file now holds **one**, mtime `2026-07-31T01:40`. Nothing in the repository deletes it, and the whole
of `.claude/checkpoints/` is gitignored, so a `git clean -xdf` or a workspace reset would remove it
without trace. **This is a hypothesis with a mechanism, not a finding** — the deletion itself was not
observed, and it is recorded here so the next session tests it rather than inherits it as a cause.

### What #165 still needs

A session started with `source ∈ {startup, resume, compact}` — **not** `clear` — that edits under a
registered surface, waits ≥60s, and re-checks. Everything else is now eliminated: attribution works,
merge works, the hook and its wiring work. What is untested is only whether `FileChanged` events keep
being *delivered* for a session's whole life.

### The methodological residue

The runbook exists because a conclusion was once drawn from a 100ms probe. It then told the next
session to do something whose result would be uninterpretable for a different reason — a null the
reader would have read as data. **A wiring check is not a formality before the experiment; it is what
decides whether the experiment's null means anything.** The runbook's own instruction to check wiring
first is what caught its own instruction to `/clear`.
