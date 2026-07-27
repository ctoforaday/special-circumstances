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
2. Restore can be **summary-aware** — `PostCompact` sees what the summary actually kept, so the
   injection can carry only the delta instead of a second narrative (§3 C, §7). This directly
   retires risks R3 and R4.

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
   did not exist.)* Bounded and terse: this steers a summarizer, it does not become the summary.
4. Exits 0. It never blocks compaction (blocking would just wedge the session).

**(C) Restore, split across two events — summary-aware where a summary exists.**

- **`PostCompact`** owns the compaction case. It receives `compact_summary`, so the hook compares
  the checkpoint against what the summary actually kept and surfaces **only the delta**. When the
  summary already carried the validation loop — which the field evidence in §12(a) says it
  sometimes does — the hook stays quiet rather than repeating it. *(The original routed this
  through `SessionStart(source=compact)` and injected the digest blind, because `PostCompact` was
  believed unconfirmed.)*
- **`SessionStart`** owns `source ∈ {resume, fork, startup}` — the cases where the context is
  fresh and no summary exists to diff against. Here it emits the full terse digest via
  `additionalContext`: objective line, plan pointer, the **validation loop**, next steps, and the
  path to the full `CHECKPOINT.md` — *not* the whole file. It re-grounds the agent in ~15 lines
  and tells it where to read more.

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

The correction makes this stronger than a policy. Non-duplication is now **measured, not
intended**: the seal asks the summarizer to keep the load-bearing items, and `PostCompact` then
reads the summary it produced and injects only what is missing from it. Where the original could
only promise terseness, the corrected design can observe whether the promise was kept — and a
summary that repeatedly drops an item the seal explicitly asked it to preserve is itself a signal
worth recording.

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
| **Validation check's trigger surface touched** | `FileChanged`, matcher = the surface | I2's *"reproduce, don't recall"* now has a mechanism: the check re-arms when its surface is edited, rather than depending on the agent remembering what arms it. |
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

**CORRECTED.** The original ran every restore through `SessionStart` and injected the digest
without knowing what the summary had kept. With `PostCompact` confirmed, restore splits by whether
a summary exists to diff against.

**Primary path (compaction) — `PostCompact` hook.** Input carries `compact_summary`. The hook
reads the newest checkpoint, **diffs it against the summary**, and injects only the items the
summary failed to carry. If the summary already holds the validation loop and the next actions,
the hook says nothing. This is the mechanism that makes non-duplication measurable rather than
merely intended, and it retires R3 (digest size) and R4 (double narrative) as design risks.

**Cold path (no summary) — `SessionStart` hook.** On `source ∈ {resume, fork, startup}` the hook
reads the newest checkpoint and returns a compact `additionalContext` payload. It surfaces only the
YAML header fields plus the **Validation loop** and **Next steps** sections, capped (~1.5 KB),
ending with the absolute path to the full note:

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
- On the `PostCompact` path, inject **only the delta** against `compact_summary`.
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
| `hooks/postcompact-restore.*` | hook (PostCompact) | **New.** Diff the checkpoint against `compact_summary`; inject only the delta. Silent when the summary already carried it. |
| `hooks/sessionstart-restore.*` | hook (SessionStart) | Restore for `source ∈ {resume, fork, startup}`: emit the terse digest via `additionalContext`; register validation trigger surfaces via `watchPaths`. Silent if no checkpoint. |
| `hooks/subagentstop-seal.*` | hook (SubagentStop) | **New.** Seal a seat's note at the moment it finishes, using `agent_id` and `agent_transcript_path` from the event — the only point where a seat's identity and its trajectory are both known. |
| `hooks/sessionend-seal.*` | hook (SessionEnd) | **New.** Seal on a session that ends without ever compacting; reason-matched. |
| `hooks/filechanged-rearm.*` | hook (FileChanged) | **New.** Re-arm a validation check when its trigger surface is edited (I2). |
| `requirements.json` (existing) | manifest | Only `git` (already required). Hooks are capability-gated — degrade to no-op + one warning if a probe is missing, per the port plan's environment-preflight discipline. |

Hooks are cross-platform per the suite convention (PowerShell + POSIX variants; the port
plan already establishes `Get-Command`/`command -v` capability gating).

**Capability gating now has a second axis: the hook events themselves.** Five of the events above
are newer than this document's first draft, and the suite must run against clients that predate
them. `/doctor` should report which hook events the installed client actually supports, and each
hook must be inert rather than broken on a client that never fires it. The seal/restore pair must
degrade to the original single-`SessionStart` design when `PostCompact` is unavailable.

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
| R3 | ~~`additionalContext` size / truncation.~~ | **Largely retired.** The `PostCompact` diff injects only the delta, so the payload shrinks to what the summary actually dropped. Cap retained as a floor. |
| R4 | ~~Duplication with the harness summary confusing the agent.~~ | **Largely retired** by the same mechanism — non-duplication is now measured against `compact_summary` rather than promised. Marker prefix retained. |
| R5 | **Checkpoint ↔ plan drift** — two sources of truth diverging. | Checkpoint is explicitly the *volatile cursor*, plan is durable; `plan` pointer + `beyond_plan` flag; fold durable decisions back on completion. |
| R6 | **Transcript-tail PII in sealed snapshots** entering git. | Snapshots gitignored (§6); only the agent-authored live note is committed. **Note the residual:** gitignore keeps them out of history, not off the box. Retention on disk is bounded by N=10 (§6) and nothing more; if that is insufficient the snapshots need scrubbing, not just ignoring. |
| R7 | **Restore fires on every `resume`, including trivial reconnects**, adding noise. | Silent when no checkpoint or when `status: done`; only inject for `in-progress`/`blocked`/`validating`. **Corrected caveat:** a stale `done` note is exactly what misleads a resumed agent, so silence must not mean invisibility — the durable pointer (I3) keeps naming the note, so its staleness is discoverable rather than absent. |
| R8 | **Cross-plugin preload** (sleeper-service using the discipline). | Same fallback as the port plan: `skills:` preload, or vendor a copy. **Now also:** `SubagentStart` returns `additionalContext` to a subagent, matched on `agent_type` — a mechanism for injecting the discipline per seat without a preload at all. Worth spiking before falling back to vendoring. |
| R9 | **Hook-surface churn.** *(New.)* Five events this design now depends on postdate its first draft, and the resolver around thinking display has already moved once between client versions. A design verified against one client is not verified against the next. | Record `client_version` in every checkpoint (§4); `/doctor` reports supported hook events; every hook is inert rather than broken when its event never fires; the seal/restore pair degrades to the original single-`SessionStart` shape when `PostCompact` is absent. |
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
per R1/R8 it may be absent or degraded. *(Editorial note added 2026-07-27: §7's mechanism has
since moved — the compaction case is now `PostCompact` and the cold case is
`SessionStart(resume/fork/startup)`. I3's argument is unaffected and slightly strengthened: there
are now two hooks that can be absent or degraded instead of one, and neither fires on a cold
fresh session. The finding below is left exactly as recorded.)* Make the **project-memory pointer file** (a stable
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
| **2. Seal hooks** | `PreCompact` seal: ensure-exists (skeleton fallback), snapshot, prune to N, `custom_instructions` fold-in, **preserve-verbatim instructions on stdout**. Plus `SubagentStop` and `SessionEnd` seals. Capability-gated on `git`. | Force `/compact` with a stub-only session → a skeleton snapshot appears; with a live note → sealed, timestamped, `agent_id`-tagged; old snapshots pruned. Two concurrent seats produce two distinct seals (R10). A session ended without compacting still leaves a seal. |
| **3. Restore + `/resume`** | `PostCompact` delta restore; `SessionStart` digest for `resume`/`fork`/`startup`; `/resume` command. | Post-compaction session shows only what the summary dropped; a summary that already carried the validation loop produces **silence**; `/resume` prints the full note. |
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
because the hook itself cannot reflect, and tells the summarizer what to preserve; a `PostCompact`
hook reads the summary that came back and hands over only what it dropped. So work that has
drifted beyond the plan is never lost to compression — and the note only says what the summary
failed to.

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
| C2 | `PostCompact` "unconfirmed", design "does not depend on" it | Exists, matched on `trigger`, input carries **`compact_summary`** | Restore split (§3 C, §7); R3 and R4 largely retired |
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
