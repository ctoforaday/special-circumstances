# Context Checkpointing — a "Memento" for Long-Running Agents

> Design proposal for a follow-up PR to **Special Circumstances**.
> Home plugin: **prosthetic-conscience** (core cowork behavior). Touches **sleeper-service**.
> Status: proposal. No code in this PR.

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

Verified against the Claude Code hooks documentation
(`https://code.claude.com/docs/en/hooks`). Where the docs were thin I say so rather than
invent.

### The compaction lifecycle

- Compaction fires **manually** (`/compact`) or **automatically** when the context window
  approaches its limit. Auto-compaction gives essentially **no advance warning** to the
  agent — it happens between turns.
- **`PreCompact` hook** — fires *before* compaction. Matchers: `manual` and `auto`. Input
  JSON includes the common fields (`session_id`, `transcript_path`, `cwd`,
  `hook_event_name`) plus, for the manual case, `custom_instructions` (whatever the user
  typed after `/compact`), and a `trigger` field distinguishing manual from auto.
- **`SessionStart` hook** — fires when a session begins or resumes. The `source` field takes
  `startup`, `resume`, `clear`, and — critically — **`compact`**: SessionStart fires again
  *after* a compaction completes. SessionStart can return
  `hookSpecificOutput.additionalContext` (a string), which the harness injects into the
  fresh context as a system reminder before the next model turn.

### The hard constraint that shapes the whole design

> **`PreCompact` cannot reflect, and cannot edit the summary.**

PreCompact runs an external script. That script receives IDs and a path to the transcript
file — it does **not** get a channel to ask the model "what were you about to do?", and per
the docs it **cannot inject content into the compacted summary** (unlike SessionStart, it
has no `additionalContext`). It can read the raw transcript from `transcript_path` and it
can block compaction (exit 2), but it cannot *think*.

The consequence is the central engineering decision of this proposal:

> The rich, semantic checkpoint — decisions, rejected options, next steps, the validation
> loop — **must be authored by the agent while it still has the context**, not manufactured
> by a hook at the moment of amnesia. The hook's job is only to *seal* the latest note and
> guarantee it exists; the *writing* is a discipline the agent runs continuously.

A hook cannot be Memento's notes. A hook can only be the reflex that makes sure the notes
were written and hands them back. This is why the design is **agent-authored checkpoint +
deterministic seal/restore hooks**, not "a hook that summarizes on compaction."

*(Uncertainty flagged: the fetched docs also referenced a `PostCompact` event and several
SessionStart output fields — `initialUserMessage`, `watchPaths`, `reloadSkills`. `PreCompact`,
`SessionStart(source=compact)`, and `additionalContext` are well-established and load-bearing
here; the others I treat as unconfirmed and the design does **not** depend on them. Phase 0
verifies exact field names on the target Claude Code build.)*

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

**(B) A `PreCompact` hook that *seals* the note — deterministic, no thinking.**
On compaction (auto or manual) the hook:
1. Confirms a checkpoint exists; if none does, it writes a **skeleton** from what it *can*
   observe deterministically — `git status`/`git diff --stat`, branch, the tail of
   `transcript_path`, timestamp — so a checkpoint always exists even if the agent was
   negligent. A stub note beats no note.
2. Copies the current `CHECKPOINT.md` to a timestamped, immutable
   `.claude/checkpoints/<ts>-<trigger>.md` (history/rotation), stamping it with `trigger`
   and `session_id`.
3. Exits 0. It never blocks compaction (blocking would just wedge the session).

**(C) A `SessionStart` hook that *restores* the note — lightweight, pointer-first.**
On `source ∈ {compact, resume}` the hook emits, via `additionalContext`, a **short**
system reminder: the checkpoint's objective line, the plan pointer, the **validation loop**,
next steps, and the path to the full `CHECKPOINT.md` — *not* the whole file. It re-grounds
the agent in ~15 lines and tells it where to read more. A **`/resume`** command re-surfaces
the full checkpoint on demand.

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

---

## 4. Checkpoint content — the schema

`CHECKPOINT.md` is Markdown with a small YAML front-matter block (machine-readable header
for hooks; prose body for the agent). One file, overwritten in place; history lives in
`.claude/checkpoints/`.

```markdown
---
schema: 1
updated: 2026-07-11T14:32:00Z
session_id: abc123
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
| **Validation loop** | The reviewer's core point: forgotten once you leave the plan. Carried verbatim, with per-step last-run state, so the agent re-runs the right checks in the right order. |
| **Decisions made / rejected** | Prevents re-litigating settled calls and re-trying rejected approaches. |
| **Files touched / working state** | Recovers a mid-edit ("`reAudit()` not yet wired in") that a summary flattens to "worked on workflow.js". |
| **Next intended steps** | The single most valuable line after amnesia: what was I about to do. |
| **Open threads** | Unresolved questions don't silently vanish. |

---

## 5. Triggers — when a checkpoint is written

| Trigger | Mechanism | Rationale |
|---|---|---|
| **Pre-compaction (auto & manual)** | `PreCompact` hook seals + snapshots | The Memento moment. Deterministic backstop; always fires. |
| **Crossing beyond the plan's scope** | skill discipline; sets `beyond_plan: true` | The exact drift the reviewer named — the checkpoint starts carrying weight precisely here. |
| **Validation loop established / changed** | skill discipline | The loop is the thing most worth preserving; capture it the moment it's defined. |
| **Decision reached / rejected** | skill discipline | Cheap to append while fresh; expensive to reconstruct. |
| **Proactive / periodic** | `PostToolUse` counter *or* time-based nudge (see risks) | Guards against auto-compaction striking before the agent voluntarily checkpoints. |
| **On demand** | `/checkpoint` command | User or agent forces a write before a risky step. |

On the **periodic** trigger: auto-compaction can strike with no warning, so we cannot rely
solely on the agent choosing to write. Two grounded options: (a) a `PostToolUse` hook that
maintains a mutation counter and, every N file-writes, emits a non-blocking reminder to
refresh the checkpoint; (b) rely on the `PreCompact` seal to snapshot whatever exists.
Recommendation: **ship (b) first** (it is the guaranteed floor and needs no per-turn hook),
add the (a) nudge in a later phase if checkpoints prove stale in practice.

---

## 6. Storage, naming, retention

```
.claude/
├── checkpoints/
│   ├── CHECKPOINT.md              # symlink or copy of the current live note (convenience)
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
- **Naming:** `<UTC-ISO-compact>-<trigger>.md`. Sortable, unambiguous, trigger visible.
- **Retention/rotation:** keep the last **N=10** snapshots per project; the `PreCompact` hook
  prunes older ones after writing. Restore always reads the newest. Cheap, bounded, no daemon.

**Git-tracked vs. session-local — the split, explicit:** the *content the agent authored*
(the live `CHECKPOINT.md` in the run dir) is part of the work product and is committed; the
*machine-sealed snapshots* (which may contain raw transcript tails) are recovery scaffolding
and stay out of git. This keeps the portfolio clean and avoids the PII-leak surface the port
plan is careful about elsewhere.

---

## 7. Restore — re-injecting the note after amnesia

**Primary path — `SessionStart` hook.** On `source ∈ {compact, resume}` the hook reads the
newest checkpoint and returns a compact `additionalContext` payload. It surfaces only the
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
- If no checkpoint exists, the hook emits **nothing** — silence beats noise.

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

**prosthetic-conscience placement.** Checkpointing is core cowork behavior (it protects any
long interactive session, not just research), so the skill, both hooks, and the two commands
**ship in prosthetic-conscience**, the base plugin the other two preload. sleeper-service
merely *depends on and invokes* it.

---

## 9. Component map

All in **prosthetic-conscience** unless noted.

| Component | Kind | Responsibility |
|---|---|---|
| `skills/context-checkpointing/SKILL.md` | skill | The discipline: when/what/how to write `CHECKPOINT.md`; the schema; the "carry the validation loop" rule. Preloaded by long-running agents; cross-refs `project-memory` + `spec-driven-development`. |
| `commands/checkpoint.md` | command | `/checkpoint [--show]` — force a write now, or print the current note. |
| `commands/resume.md` | command | `/resume` — print the full current checkpoint and re-anchor. |
| `hooks/precompact-seal.*` | hook (PreCompact) | Seal: ensure a checkpoint exists (skeleton from `git`/transcript tail if absent), snapshot to `.claude/checkpoints/`, prune to N, exit 0. Never blocks. |
| `hooks/sessionstart-restore.*` | hook (SessionStart) | Restore: on `source ∈ {compact,resume}`, emit the terse digest via `additionalContext`. Silent if none. |
| `requirements.json` (existing) | manifest | Only `git` (already required). Hooks are capability-gated — degrade to no-op + one warning if a probe is missing, per the port plan's environment-preflight discipline. |

Hooks are cross-platform per the suite convention (PowerShell + POSIX variants; the port
plan already establishes `Get-Command`/`command -v` capability gating).

---

## 10. Alternatives considered

1. **Hook-only, no agent discipline ("summarize on PreCompact").** *Rejected.* PreCompact
   cannot reflect and cannot edit the summary — it only runs a script with transcript
   access. Any note it writes is a mechanical transcript slice, missing the decisions,
   rejections, and validation loop that are the whole point. It is the right *backstop*
   (§3 B) but cannot be the primary author.
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
| R1 | **Exact hook field names/behavior on the target build.** The fetched docs asserted fields (`PostCompact`, `initialUserMessage`, `watchPaths`, `reloadSkills`) I could not independently confirm. | Design depends only on `PreCompact` + `SessionStart(source=compact/resume)` + `additionalContext`, all well-established. **Phase 0 spike verifies field names empirically** before building. |
| R2 | **Auto-compaction with no warning + stale note.** If the agent hasn't checkpointed recently, the seal captures a stale cursor. | PreCompact skeleton from `git`/transcript tail as a floor; add the periodic `PostToolUse` nudge if staleness is observed. |
| R3 | **`additionalContext` size / truncation.** A fat digest wastes the freshly-reclaimed context or gets clipped. | Hard cap (~1.5 KB), digest-not-dump, path pointer for the rest. |
| R4 | **Duplication with the harness summary** confusing the agent. | Distinct marker prefix; inject only forward-looking sections; terseness as policy (§3, §7). |
| R5 | **Checkpoint ↔ plan drift** — two sources of truth diverging. | Checkpoint is explicitly the *volatile cursor*, plan is durable; `plan` pointer + `beyond_plan` flag; fold durable decisions back on completion. |
| R6 | **Transcript-tail PII in sealed snapshots** entering git. | Snapshots gitignored (§6); only the agent-authored live note is committed. |
| R7 | **Restore fires on every `resume`, including trivial reconnects**, adding noise. | Silent when no checkpoint or when `status: done`; only inject for `in-progress`/`blocked`/`validating`. |
| R8 | **Cross-plugin preload** (sleeper-service using the discipline). | Same fallback as the port plan: `skills:` preload, or vendor a copy — de-risked by the existing Phase 1 harness spike. |

---

## 12. Phased build plan

| Phase | Work | Verify |
|---|---|---|
| **0. Hook reality spike** | Register no-op `PreCompact` + `SessionStart` hooks that log their full input JSON. Trigger `/compact` and an auto-compaction; capture actual `source`, `trigger`, `custom_instructions`, and confirm `additionalContext` injects. | Logged JSON matches assumptions; `additionalContext` visibly re-grounds the agent post-compaction. Resolves R1. |
| **1. The note + skill** | `context-checkpointing` skill (schema + discipline) and `/checkpoint [--show]`. No hooks yet — pure agent discipline. | On a real task, agent maintains a valid `CHECKPOINT.md`; `/checkpoint --show` prints it; validation loop captured verbatim. |
| **2. Seal hook** | `PreCompact` seal: ensure-exists (skeleton fallback), snapshot to `.claude/checkpoints/`, prune to N, `custom_instructions` fold-in. Capability-gated on `git`. | Force `/compact` with a stub-only session → a skeleton snapshot appears; with a live note → it's sealed + timestamped; old snapshots pruned. |
| **3. Restore hook + `/resume`** | `SessionStart` restore digest (size-capped, forward-first, silent-when-absent) + `/resume` command. | Post-compaction session shows the terse digest with the validation loop and next steps; `/resume` prints the full note; no double-narrative with the harness summary. |
| **4. Integration** | Wire into `spec-driven-development` (plan pointer + `beyond_plan`) and `project-memory` (promotion on completion); require it in the sleeper-service run loop; `/self-improve` restarts `/resume` from checkpoint. | A long `/self-improve` headless run survives a forced compaction and resumes on the correct validation step. |
| **5. Freshness (optional)** | Add `PostToolUse` staleness *nudge* (non-blocking) only if Phase 4 shows stale seals. | Nudge appears every N mutations; no interference with agent writes; measurably fresher seals. |

**Estimate:** ~2–3 working days. Phase 0 is the only real unknown; everything after is
plumbing around a Markdown file plus two thin hooks. Ship 1–4; treat 5 as demand-driven.

---

## 13. One-line summary

*Leave yourself a note before the amnesia hits.* The agent maintains a living
`CHECKPOINT.md` — objective, plan pointer, decisions kept and rejected, working state, next
steps, and above all the **validation loop** — a `PreCompact` hook seals it deterministically
because the hook itself cannot reflect, and a `SessionStart` hook hands the digest back the
moment context is compacted, so work that has drifted beyond the plan is never lost to
compression.
