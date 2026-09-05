# Context Checkpointing — a "Memento" for Long-Running Agents

> The archaeology — superseded designs, the correction record, and the acceptance and measurement
> runs — is [`plans/historical/context-checkpointing.md`](historical/context-checkpointing.md).
> Section numbering is shared: §§1–11, 13 and 14 are here; §12 and §§15–20 are there.

> STATUS 2026-09-02: **built and shipped in `prosthetic-conscience`** — the
> `context-checkpointing` skill, `/checkpoint`, `/resume`, and the seal / restore / observe /
> re-arm hooks are all live. Phase 4 is partial: I1 enforcement is refuted (no event exists to
> carry it) and the sleeper-service wiring is not started (§13). Phase 5 graduated into
> `plans/checkpoint-freshness.md`.
>
> Home plugin: **prosthetic-conscience** (core cowork behaviour). Consumed by **gray-area**;
> required by **sleeper-service** (not yet wired).

---

## 1. The problem

In *Memento*, the protagonist cannot form new memories, so he tattoos the load-bearing
facts onto his body and leaves himself notes — because he knows the amnesia is coming and
that his future self will not remember deciding anything. A long-running Claude Code agent
has the same disability. When the context window fills, the harness **compacts**: it
replaces the running transcript with a model-generated summary and continues. Whatever the
summary drops is simply gone from the agent's working memory.

This is survivable while the agent is executing a written plan — the plan is on disk and
re-grounds it. The failure mode is **work that has drifted beyond the plan's scope**: the
interesting decisions made in the last hour, the three approaches tried and rejected, the
half-finished refactor, and — most insidiously — **the validation loop**
("re-run `qlty check` then the three integration tests, in that order"). Validation loops
are the first casualty, because they live in the conversation, not the plan, and once you
are past the plan nobody re-derives them. The agent post-compaction believes it is done, or
re-litigates a settled decision, or ships without re-running the checks it had agreed to run.

**The fix, stated as Memento states it:** leave a note *while you still remember*, in a
known place, and make sure your future self is handed it the moment the amnesia lifts.

---

## 2. Grounding: what Claude Code actually gives us

Verified against the shipped client's own hook-event catalogue — **Claude Code 2.1.220**, native
binary, `GIT_SHA 4073f595…`, built 2026-07-24 — by reading the thing itself rather than the docs.
An earlier revision of this section was verified against the hooks *documentation* and two of its
load-bearing claims were wrong; the corrections are tabulated in historical §15.

### The compaction lifecycle

- Compaction fires **manually** (`/compact`) or **automatically** when the context window
  approaches its limit. Auto-compaction gives essentially **no advance warning** to the
  agent — it happens between turns.
- **`PreCompact`** — fires *before* compaction, matched on `trigger` ∈ {`manual`, `auto`}. Input
  carries the common fields plus `custom_instructions`. **Exit 0, and its stdout is appended as
  the custom compact instructions**, so it can tell the summarizer what to preserve verbatim.
- **`PostCompact`** — matched on `trigger` ∈ {`manual`, `auto`}; its input carries
  **`compact_summary`: the summary that was just written.** It **cannot inject anything** (§3 C).
- **`SessionStart`** — `source` ∈ {`startup`, `resume`, `clear`, `compact`, `fork`}. Input also
  carries `agent_type`, `model` and `session_title`. Returns `hookSpecificOutput.additionalContext`,
  injected before the next model turn. Its output also supports **`watchPaths`** (register files
  with the `FileChanged` watcher) and **`reloadSkills`**.
- **`SubagentStart` / `SubagentStop`** — matched on `agent_type`. `SubagentStart` returns
  `additionalContext` to the subagent; `SubagentStop` receives `agent_id`, `agent_type` and
  **`agent_transcript_path`**. The checkpointing half needs `agent_id` from them (§4).
- **`FileChanged`** — matcher names the files to watch, fires on change/add/unlink. This is the
  mechanism improvement **I2** (§5) is built on.
- **`TaskCreated` / `TaskCompleted`** — present in the catalogue but **refuted in practice**:
  registered against a live background task and a live subagent, neither ever fired
  (`hook-surface-spike.md` §9b). I1 therefore has no enforcement point (§5).

### The constraint

> **`PreCompact` cannot reflect.** It is an external script. It receives ids and a path to the
> transcript; it has no channel to ask the model "what were you about to do?". It can read the
> transcript, steer the summary, and block compaction (exit 2) — but it cannot *think*.

The central engineering decision follows from it:

> The rich, semantic checkpoint — decisions, rejected options, next steps, the validation
> loop — **must be authored by the agent while it still has the context**, not manufactured
> by a hook at the moment of amnesia. The hook's job is only to *seal* the latest note and
> guarantee it exists; the *writing* is a discipline the agent runs continuously.

A hook cannot be Memento's notes. A hook can only be the reflex that makes sure the notes were
written and hands them back. This is why the design is **agent-authored checkpoint + deterministic
seal/restore hooks**, not "a hook that summarizes on compaction."

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
1. Confirms a checkpoint exists. *(The design called for a skeleton fallback — `git status`, branch,
   transcript tail, timestamp — when none does. Not built: the hook stays silent instead, which is
   the same outcome by a different route.)*
2. Copies the current `CHECKPOINT.md` to a timestamped, immutable
   `.claude/checkpoints/<ts>-<trigger>-<agent_id>.md` (history/rotation), stamping it with
   `trigger`, `session_id` and `agent_id`.
3. **Emits preserve-verbatim compact instructions on stdout** — naming the validation loop, the
   ordered next actions, the in-flight state handles and the `beyond_plan` flag as things the
   summary must carry. *(Verified end-to-end: a marker emitted here survived into 2/2 summaries.)*
   Bounded and terse: this steers a summarizer, it does not become the summary.
   **The instruction MUST be grounded in the conversation being summarized** — reference and
   reinforce what is already there, never introduce. A seal asking to preserve content absent from
   the session is indistinguishable from prompt injection, and the summarizer says so *in the
   summary*, which lands that suspicion in the restored context. If the note's contents are not in
   the transcript, the seal has nothing legitimate to ask for and stays silent.
4. Exits 0. It never blocks compaction (blocking would just wedge the session).

**Steering is `PreCompact`-only.** `SubagentStop` stdout reaches a live seat, so an instruction
there is a directive that seat never established.

**(C) Restore is `SessionStart`, on every source including `compact`.**

`PostCompact` cannot inject. Measured three ways: it is absent from the `hookSpecificOutput` union
in the client (twenty events have an output shape; it is not one of them); its documented exit-0
behaviour is *"stdout shown to **user**"*, where `SessionStart` says *"stdout shown to **Claude**"*;
and a marker emitted by a `PostCompact` hook appears **nowhere** in the resulting transcript, while
the same marker from `SessionStart` materialises as a `hook_additional_context` attachment and the
model reports seeing it (`plans/hook-surface-spike.md`). Ordering kills it independently anyway: the
per-boundary sequence is **`PreCompact` → `SessionStart(compact)` → `PostCompact`**, so the only
hook that *can* inject runs *before* the summary exists. No two-hook relay can work in that order.

So restore is one hook:

- **`SessionStart`** owns **every** source — `compact`, `resume`, `fork`, `startup`. It emits the
  terse digest via `additionalContext`: objective line, plan pointer, the **validation loop**, next
  steps, and the path to the full `CHECKPOINT.md` — *not* the whole file. It re-grounds the agent
  in ~15 lines and tells it where to read more.
- **`PostCompact` is observability only.** It sees the summary and can report to the human. It is
  the right place to *record* what a summary preserved or dropped — evidence for tuning the seal
  (§3 B) — and the wrong place to restore anything.

**The digest is flagged as untrusted by the resumed agent, and that is the design working.**
Tested end-to-end: the digest leads with its provenance — file, timestamp, session id — names the
session's own objective, and quotes the note verbatim. The resumed agent recovered every value
exactly, attributed them honestly to the hook, **and flagged the payload anyway**: *"untrusted file
content … formatted to read as authoritative state … embeds an imperative instruction."* Three
rules follow (evidence in historical §16 B):

1. **The flag cannot be designed away.** The imperative the agent named was the note's own
   `Invariants / foot-guns` entry. A section whose purpose is to carry foot-guns carries imperatives
   by definition; removing them removes the reason to restore the note.
2. **What IS enforceable: the hook adds no imperative of its own.** An instruction arriving inside
   injected text reads as foreign however reasonable it is, and one the hook invented is one the
   session never established. The verify duty belongs to the skill, which the session already
   carries. This is a unit-tested rule.
3. **The distrust is the correct posture, not a defect.** Both acceptance runs used the content
   accurately while labelling it a claim rather than a fact — which is what §5's own contract asks
   for (*the note is a claim, not a fact*). A checkpoint the agent distrusts **and uses** is the
   design working.

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

**Non-duplication is a policy, not a mechanism.** The digest is written before the summary can be
inspected — `SessionStart` runs before `PostCompact` — so overlap cannot be measured away
per-boundary. What `PostCompact` can do is **record** the overlap after the fact: a summary that
repeatedly drops an item the seal explicitly asked it to preserve is a signal worth accumulating.
Terseness per-boundary; measurement across boundaries. That is R4, and it is live.

---

## 4. Checkpoint content — the schema

`CHECKPOINT.md` is Markdown with a small YAML front-matter block (machine-readable header
for hooks; prose body for the agent). One file, overwritten in place; history lives in
`.claude/checkpoints/`.

**The normative schema is `plugins/prosthetic-conscience/skills/context-checkpointing/SKILL.md`,
not this document.** It is `schema: 3`. The front-matter carries `written_at` / `reaffirmed_at`
(both set by *reading a clock*, per `plans/checkpoint-freshness.md` §III), `session_id`,
`agent_id`, `agent_type`, `client_version`, `objective`, `plan`, `beyond_plan` and `status`. The
body sections are Objective · Plan pointer · **Validation loop** · Next intended steps · In-flight
handles · Invariants / foot-guns · Decisions made / rejected · Files touched · Open threads.
The schema-2 form this document originally proposed is preserved in historical §A5.

Two of the header fields exist for reasons worth restating. `agent_id` is in the schema *and* in
the snapshot filename because every subagent shares the parent's `session_id`, so a
`session_id`-keyed snapshot would have parallel seats overwrite each other (R10).
`client_version` is recorded because the hook surface is version-bound (R9).

Sections, and why each survives compaction:

| Section | Why it must survive |
|---|---|
| **Objective** | Post-compaction the agent must not re-scope or think it's done. |
| **Plan pointer + `beyond_plan`** | Re-grounds in the SDD plan; *names the drift* explicitly. |
| **Validation loop** | The core point: forgotten once you leave the plan. Carried verbatim, with per-step last-run state, so the agent re-runs the right checks in the right order. Each check also records its **trigger surface** — what re-arms it — because compaction drops that first (I2). |
| **Decisions made / rejected** | Prevents re-litigating settled calls and re-trying rejected approaches. |
| **In-flight handles** | Background task ids, pull requests, long-running processes — the state a summary has no slot for. |
| **Invariants / foot-guns** | Carried verbatim. This is the section that makes the digest read as imperative, and it is not removable (§3 C). |
| **Files touched / working state** | Recovers a mid-edit ("`reAudit()` not yet wired in") that a summary flattens to "worked on workflow.js". |
| **Next intended steps** | The single most valuable line after amnesia: what was I about to do. Each real work item also carries a pointer to its **canonical-queue** home (issue / `plan.md` task); a note-only actionable silently dies when the worklist is rebuilt from another index (I1). |
| **Open threads** | Unresolved questions don't silently vanish. |

---

## 5. Triggers — when a checkpoint is written

| Trigger | Mechanism | Rationale |
|---|---|---|
| **Pre-compaction (auto & manual)** | `PreCompact` hook seals + snapshots | The Memento moment. Deterministic backstop; always fires. |
| **Crossing beyond the plan's scope** | skill discipline; sets `beyond_plan: true` | The exact drift this design exists for — the checkpoint starts carrying weight precisely here. |
| **Validation loop established / changed** | skill discipline | The loop is the thing most worth preserving; capture it the moment it's defined. |
| **Decision reached / rejected** | skill discipline | Cheap to append while fresh; expensive to reconstruct. |
| **Actionable identified (beyond-plan work item)** | skill discipline **only**. The intended enforcement was `TaskCreated`/`TaskCompleted` with exit 2 preventing creation or completion; **both events are REFUTED** — registered against a live background task and a live subagent, neither ever fired (`hook-surface-spike.md` §9b) — so **I1 has no enforcement point and remains pure discipline** | A note-only actionable dies when the resumed agent rebuilds its worklist from a different index. The note points; the durable store holds. |
| **Validation check's trigger surface touched** | `FileChanged`, registered via `SessionStart`'s `watchPaths` | I2's *"reproduce, don't recall"*, and it is built. `watchPaths` takes **paths — files or directories**, relative or absolute. A **directory watch is recursive and fires `add` for files created later**, so a trigger surface is registered by watching the directories that contain it. No wildcard form works — not globs, not regex — and a pattern registers *nothing*, silently. Watch specific directories, never the project root: a `.` watch caught the hook's own log and re-triggered itself ten times (`hook-surface-spike.md` §6). |
| **Session ending without compaction** | `SessionEnd`, matched on **every** reason — headless reports `other` | A session closed by `clear`/`logout`/`prompt_input_exit` never reaches `PreCompact`, so the note would never be sealed. |
| **Subagent finishing** | `SubagentStop`, keyed by `agent_id` | The only point where a seat's identity and its trajectory are both known. |
| **Proactive / periodic** | Staleness gauge + `Stop`-channel nudge, in `plans/checkpoint-freshness.md` | Auto-compaction can strike with no warning, so we cannot rely solely on the agent choosing to write. Shipped there, inert until thresholds are chosen. |
| **On demand** | `/checkpoint` command | User or agent forces a write before a risky step. |

### The three improvements in force

Field-evidenced on a live six-boundary session; the evidence is historical §12, and I2's build
record is historical §18.

**I1 — Promote actionables to the canonical queue; the note points, it does not hold.**
For any forward action item (a "Next intended step" or "Open thread" that is real work, not a
musing), the discipline is: **file/link it in the durable store the resumed workflow actually
reads** (GitHub issue, or a numbered task in the SDD `plan.md`), and let the checkpoint carry
the *pointer*. A checkpoint-only actionable is flagged as such so the omission is visible, not
silent. **Discipline only — the enforcement events do not fire.**

**I2 — Validation-loop entries record each check's *trigger surface*, and the discipline is
reproduce, not recall.** Per check: **what re-arms it** — the file/condition surface that makes it
fire. After amnesia you **reproduce** the gate locally before trusting green; you do not act on the
summary's paraphrase of what it wanted. Mechanised: `sc-checkpoint-restore` parses the loop,
resolves each check's `re-armed by:` surface, and registers it via `watchPaths`;
`sc-filechanged-rearm` matches `file_path` back to the check that claimed that surface — by
**longest** surface, so a check naming `tools/internal` beats a sibling naming `tools` — and
records the staleness beside the note. The next digest reports it. Three rules hold it up:

- *An unresolvable surface is data.* `re-armed by: a human deciding to ship` has no path. The
  digest **names** it, because a session that silently watched nothing looks identical to one that
  watched everything.
- *An unclaimed change is recorded nowhere.* A directory watch is coarser than the surface it
  stands in for, so changes matching no check are expected and are not evidence. Logging them
  would turn a signal into a log.
- *State lives outside the watched tree, by rule.* The re-arm hook ignores every change under
  `.claude/`, so a note naming `.claude` as a surface cannot start a loop.

It records that a recorded result is stale. It does **not** re-run anything and does not tell the
agent to — acting on it is the `validation-loop` skill's business.

**A trigger surface needs a `/` or a leading `.` to be recognised.** `re-armed by: scripts`
resolves to nothing; `scripts/` resolves. The digest reports the miss, so the gap is visible rather
than silent, but the trailing slash is a sharp edge.

**I3 — Back the restore hook with the durable memory pointer.** `SessionStart` does not fire on a
cold fresh session, and on an older client it may be degraded. Make the **project-memory pointer
file** (a stable `MEMORY.md` entry naming the live checkpoint's path) a required companion output
of the checkpoint discipline, so continuity survives even when no hook runs. This is the mechanism
that actually carried the field session. It makes §8's `project-memory` cross-reference
load-bearing rather than advisory.

**Rejected on audit:** a restore-time "reconcile every index" mechanism (the hook cannot enumerate
arbitrary queues generically — folds into I1 at authoring time); a per-turn auto-write checkpoint
(unevidenced — the manual cadence did not fail on staleness); enlarging the restore digest (no
truncation loss observed).

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
  Rationale: snapshots are transient recovery state, and we don't want session churn or transcript
  tails polluting git or leaking into portfolio history.
- **Naming:** `<UTC-ISO-compact>-<trigger>-<agent_id>.md`. Sortable, unambiguous, trigger visible,
  and **seat-disambiguated** — concurrent subagents share a `session_id` (R10). Same-second
  collisions across seats, across events, and identical-event all survive: suffixed, never
  overwritten.
- **Copy, never symlink.** Symlink creation on Windows requires developer mode or elevation, and
  Windows is the primary development box; a symlink here would fail exactly where the suite is used
  most, and [[agent-guardrails]] forbids resolving that by escalating. Copy is cheap and the file
  is small.
- **Retention/rotation:** keep the last **N=10** snapshots per project; the `PreCompact` hook
  prunes older ones after writing. Restore always reads the newest. Cheap, bounded, no daemon.
  This bounds the *snapshot directory* only — the **live note is rotated by supersession**
  (one block, each seal replacing the last), which is a separate discipline and the one that
  actually failed in the field (historical §12).

**Git-tracked vs. session-local — the split, explicit:** the *content the agent authored*
(the live `CHECKPOINT.md` in the run dir) is part of the work product and is committed; the
*machine-sealed snapshots* (which may contain raw transcript tails) are recovery scaffolding
and stay out of git. This keeps the portfolio clean and avoids a PII-leak surface.

---

## 7. Restore — re-injecting the note after amnesia

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

**Lightweight & non-duplicative (rules the hook enforces):**
- Inject the **digest**, never the whole file — the file path is in the digest.
- Lead with the marker `[context-checkpoint restored …]` so the agent knows this is
  recovered operational state distinct from the harness's own summary.
- Prioritize **forward-looking** sections (validation loop, next steps); the harness summary
  already covers the backward-looking narrative.
- Add **no imperative of its own** (§3 C). Unit-tested.
- If no checkpoint exists, the hook emits **nothing** — silence beats noise.

**Pointer instead of digest in exactly two cases, both by INTENT rather than by mechanism:**
`source == clear` (the human just wiped the context) and `status: done` (the note describes
finished work, and a forgotten note would otherwise re-impose dead state forever). **Age is
deliberately not a criterion** — a resume days later is when the note is worth most.
Note the cost of the `clear` carve-out: a pointer session registers **zero** `watchPaths`, so I2's
re-arm collects nothing there. That is deliberate and tested
(`TestPointerSessionsRegisterNoWatch`), and it is a trap for anyone measuring re-arm behaviour in a
`/clear`-started session (historical §20).

**The restore path is read-only until the ordered next-actions list.** Post-compaction the harness
re-presents previously-invoked skills — including ones originally invoked with mutating arguments —
wrapped in do-not-re-execute guards. The interactive harness protects this today; a headless
restore that naively replayed checkpoint or transcript content would re-run side-effectful steps.
Nothing replayed from before the seam is executable, and every checkpoint claim is verified against
reality before it is acted on. The first field report caught the note asserting a tag that had
never been pushed; that check is cheap and is the whole reason for preferring a record over a
recollection.

**The note is a claim, and stays true only while the agent maintains it.** Measured on this
repository: a digest's `In-flight handles` said *"level with origin/main at 5de625d; no open PRs"*
while `main` had moved and another pull request had merged. No hook can take that duty over — §2's
central argument is that the rich note must be agent-authored, and this is the cost of it. It is
also why `status: done` is a weak carve-out signal: the note this repo carried before that run was
two days stale, described finished work, and was still `status: validating`. **A forgotten note is
the common case, not the edge** (historical §19).

---

## 8. Relationship to the rest of Special Circumstances

**SDD plans.** A checkpoint is *not* a plan and must not drift into being one. The plan is
the durable, up-front spec (`spec-driven-development` skill, `projects/<name>/plan.md`); the
checkpoint is the **volatile execution cursor** over that plan, plus everything that has
happened *beyond* it. The `plan` pointer links them; `beyond_plan: true` is exactly the
signal "the plan no longer fully covers me, so the checkpoint is now load-bearing." A clean
completion should fold durable decisions back into the plan/`MEMORY.md` and the checkpoint
can be discarded — the note is scaffolding, not an artifact of record.
*(Not yet wired: `spec-driven-development/SKILL.md` carries no checkpoint reference — §13 Phase 4.)*

**`project-memory` skill.** Checkpoints are the short-horizon complement to project memory's
long-horizon record. Memory = "what this project *is* and the decisions that stuck";
checkpoint = "where the cursor is *right now*." Wired: `project-memory/SKILL.md` names
`CHECKPOINT.md` and [[context-checkpointing]], and graduation of a decision from checkpoint →
`MEMORY.md` is the same promotion discipline used elsewhere.

**sleeper-service / self-improve loop.** Two connections, **neither built** — that plugin is still
scaffold-only (§13 Phase 4):
1. The autonomous `/self-improve` and `/graduate` runs are *precisely* the long, unattended
   sessions most likely to hit auto-compaction with no human watching — they are the primary
   beneficiary. sleeper-service should **require** the checkpointing discipline in its run
   loop, and its scheduled `claude -p` invocations should `/resume` from checkpoint on
   restart.
2. Stale-checkpoint incidents are a natural **self-improvement signal**: if a restored
   checkpoint's validation loop turns out wrong or missing, that's a graduation candidate for
   tightening the checkpointing skill itself — the loop improving the loop.

**Placement — it ships in `prosthetic-conscience`.** Checkpointing is core cowork behaviour and
belongs in the base plugin the others preload. (It was briefly retargeted into gray-area by
association; the retarget and the argued reversal are historical §15.) The deciding question was
whether a consumer wants continuity without the miner, and the answer is plainly yes. Three
grounds:

- **Consent.** Gray Area reads transcripts — user text, file paths, whatever a tool result
  contained. Checkpointing writes a note about your own work. Bundling them makes a consumer accept
  a surveillance capability to get compaction survival, which is an unnecessary trust decision
  charged for a benign feature.
- **Sequencing.** This half was hand-run across six compaction boundaries and works (historical
  §12). Inside gray-area it would wait on a Go miner that did not exist. In prosthetic-conscience it
  lands against a plugin that already ships and already carries the always-on rules it protects.
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

**Naming note.** The design named one binary per role; #201 step 3 re-cut them to **one binary per
EVENT**, with the roles as units composed inside. `sc-checkpoint-seal` (one binary, three events,
behind an `-event` flag) became `sc-precompact` / `sc-sessionend` / `sc-subagentstop`;
`sc-checkpoint-restore` became the `internal/checkpointrestore` unit composed into
`sc-sessionstart` alongside the toolchain nudge. The role names below are the ones the rest of this
document and the Go comments use.

| Component | Kind | Responsibility |
|---|---|---|
| `skills/context-checkpointing/SKILL.md` ✅ | skill | The discipline: when/what/how to write `CHECKPOINT.md`; the schema; the "carry the validation loop" rule; rotate-don't-accumulate; read-only restore. Cross-refs `project-memory` + `spec-driven-development`. |
| `commands/checkpoint.md` ✅ | command | `/checkpoint [--show]` — force a write now, or print the current note. |
| `commands/resume.md` ✅ | command | `/resume` — print the full current checkpoint and re-anchor. |
| **`sc-precompact`** ✅ | hook (PreCompact) | Seal: snapshot to `.claude/checkpoints/`, prune to N, fold in `custom_instructions`, **emit preserve-verbatim compact instructions on stdout**, exit 0. Never blocks. |
| **`sc-postcompact-observe`** ✅ | hook (PostCompact) | **Observability only** — `PostCompact` cannot inject (§3 C). Records what the summary preserved or dropped, as evidence for tuning the seal. Never restores; never writes to stdout, which reaches the human rather than the model. Scores each note section's distinctive vocabulary against the summary and appends one row per boundary; the row is labelled with its probe, because token overlap is exploration and can never back a finding. |
| **`internal/checkpointrestore`** in `sc-sessionstart` ✅ | hook (SessionStart) | **The restore path, all sources including `compact`.** Emits the terse digest via `additionalContext`. Pointer instead of digest on `source == clear` and `status: done` (§7). Registers `watchPaths`: each validation check's trigger surface resolved to the directories containing it — paths only, no wildcards, directories recursive — and a surface that resolves to nothing is **named in the digest** rather than silently unwatched. Silent if no checkpoint. |
| **`internal/checkpoint`** ✅ | library | Shared by the seal and the restore: which file is the note, and what a section is. Two copies of that rule drift, and a restore reading a different file than the seal wrote is the failure with no symptom — both halves report success and the continuity is silently gone. |
| **`sc-subagentstop`** ✅ | hook (SubagentStop) | Seal a seat's note at the moment it finishes, using `agent_id` and `agent_transcript_path` from the event. |
| **`sc-sessionend`** ✅ | hook (SessionEnd) | Seal on a session that ends without ever compacting; **every** reason, because headless reports `other`. |
| **`sc-filechanged-rearm`** ✅ | hook (FileChanged) | Re-arm a validation check when its trigger surface is edited (I2). Input carries `file_path` and `event` (`change` \| `add`). Filters on `file_path`, which it must do anyway because a directory watch is coarser than the pattern it stands in for. |
| `requirements.json` (existing) | manifest | Only `git` (already required). Hooks are capability-gated — degrade to no-op + one warning if a probe is missing. |

Every hook command is wrapped in the bootstrap guard: a fresh plugin-cache version ships from git
WITHOUT binaries (they arrive via `doctor --fix`), and an unguarded hook crash-storms every tool
call in that window. The guard degrades to one stderr line pointing at the fix.

**Capability gating has a second axis: the hook events themselves.** Several of these events are
newer than this document's first draft, and the suite must run against clients that predate them.
`/doctor` should report which hook events the installed client actually supports, and each hook
must be inert rather than broken on a client that never fires it. Restore itself has no such
axis — it is `SessionStart` only, the one event that predates this document. What degrades on an
older client is the *observability* half: no `PostCompact` means no measurement of what the summary
dropped, and the seal loses its `custom_instructions` fold-in. Neither costs continuity.

---

## 10. Alternatives considered

1. **Hook-only, no agent discipline ("summarize on PreCompact").** *Rejected.* `PreCompact` runs a
   script, so any note it writes is a mechanical transcript slice, missing the decisions,
   rejections, and validation loop that are the whole point. Steering the summary is not the same
   as authoring the note, and the ability to say *"keep the validation loop"* is worthless if
   nothing wrote the validation loop down. It is the right *backstop* (§3 B) and cannot be the
   primary author.
2. **Rely on the harness compaction summary alone.** *Rejected — it's the status quo that
   fails.* The summary is backward-looking and lossy exactly on forward-looking procedural
   state (the validation loop), which is what we must preserve.
3. **Encode the checkpoint in `custom_instructions` on `/compact`.** *Partial.* Works only
   for *manual* compaction and only carries a short string; useless against auto-compaction.
   Honored when present (the seal hook folds `custom_instructions` into the note) but not a
   foundation.
4. **A structured store (SQLite/JSON) instead of Markdown.** *Rejected for v1.* The note's
   primary reader is a language model; Markdown is the highest-fidelity, lowest-friction,
   git-diffable format, and it matches the suite's "filesystem is the blackboard" contract.
   YAML front-matter gives the hooks the machine-readable slice they need.
5. **Continuous per-turn checkpointing via a `PostToolUse` hook writing every edit.**
   *Deferred.* Correct in spirit (freshness) but a per-turn hook writing files is noisy and
   can fight the agent's own writes. Freshness landed instead as a gauge plus a `Stop`-channel
   nudge in `plans/checkpoint-freshness.md`.

---

## 11. Open questions & risks

| # | Risk / question | Status / mitigation |
|---|---|---|
| R1 | Exact hook field names/behavior on the target build. | **Resolved** against 2.1.220 by reading the client's own event catalogue rather than the docs (historical §15). `initialUserMessage` still unverified and unused. Phase 0 stands as a **re-verification against whatever client the consumer runs** — see R9. |
| R2 | **Auto-compaction with no warning + stale note.** If the agent hasn't checkpointed recently, the seal captures a stale cursor. | **Live, and demonstrated on this repo** (historical §19). The skeleton fallback was not built. Freshness gauge + nudge shipped in `plans/checkpoint-freshness.md`, inert until thresholds are chosen. |
| R3 | **`additionalContext` size / truncation.** A fat digest wastes the freshly-reclaimed context. | **Live.** Hard cap (~1.5 KB), digest-not-dump, path pointer for the rest. There is no way to diff against the summary before injecting, so terseness is a policy rather than a measurement. |
| R4 | **Duplication with the harness summary.** | **Live.** The digest is injected before the summary can be inspected, so overlap cannot be prevented per-boundary. Distinct marker prefix; forward-looking sections only; terseness as policy (§3, §7). `PostCompact` *records* the overlap after the fact — measured, the digest is redundant against a long summary and load-bearing against a short one (historical §16 C, §17 D). |
| R5 | **Checkpoint ↔ plan drift** — two sources of truth diverging. | Checkpoint is explicitly the *volatile cursor*, plan is durable; `plan` pointer + `beyond_plan` flag; fold durable decisions back on completion. |
| R6 | **Transcript-tail PII in sealed snapshots** entering git. | Snapshots gitignored (§6); only the agent-authored live note is committed. **Residual:** gitignore keeps them out of history, not off the box. Retention on disk is bounded by N=10 and nothing more; if that is insufficient the snapshots need scrubbing, not just ignoring. |
| R7 | **Restore fires on every `resume`, including trivial reconnects**, adding noise. | Pointer-only when `status: done` or `source == clear`. **Caveat:** a stale `done` note is exactly what misleads a resumed agent, so silence must not mean invisibility — I3's durable pointer keeps naming the note, so its staleness is discoverable rather than absent. |
| R8 | **Cross-plugin preload** (sleeper-service using the discipline). | `skills:` preload, or vendor a copy. **Also:** `SubagentStart` returns `additionalContext` to a subagent, matched on `agent_type` — a mechanism for injecting the discipline per seat without a preload at all. Unspiked; open. |
| R9 | **Hook-surface churn.** A design verified against one client is not verified against the next — and this risk has been **realised twice**: once by inference standing in for measurement (historical §15, C2), and once by a control surface vanishing between patch versions (`CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` was a working lever on 2.1.220 and is inert on 2.1.240; the lever is now `--autocompact <auto\|tokens>` — `hook-surface-spike.md` §11). | Record `client_version` in every checkpoint (§4); `/doctor` reports supported hook events; every hook is inert rather than broken when its event never fires. **The load-bearing path is deliberately built on the oldest event available:** restore is `SessionStart`, which predates this document, so a client too old for `PostCompact` loses observability and the seal's `custom_instructions` fold-in — never continuity. Newer events are enrichment only, by construction. |
| R10 | **Concurrent seats overwrite each other's seals.** All subagents share the parent `session_id`. | **Resolved:** `agent_id` in the schema and in the snapshot filename (§4, §6); same-second collisions are suffixed, not overwritten. |
| R11 | **`FileChanged` delivery over a session's whole life** is untested. | Attribution, merge, the hook and its wiring are all eliminated as causes (historical §20). What remains untested is only whether events keep being *delivered* for a session's whole life. Needs a session started with `source ∈ {startup, resume, compact}` — **not** `clear` — that edits under a registered surface and re-checks. |

---

## 13. Phased build plan

| Phase | Work | Verify |
|---|---|---|
| **0. Hook reality spike** ✅ (recorded in `plans/hook-surface-spike.md`, re-run through 2.1.240) | Register no-op `PreCompact`, `PostCompact`, `SessionStart`, `SubagentStop` and `SessionEnd` hooks that log full input JSON. Trigger a manual and an automatic compaction and a subagent run. | Logged JSON matches §2 **on the client the consumer runs**. Specifically: `compact_summary` non-empty; `agent_transcript_path` present and readable; `PreCompact` stdout demonstrably reaches the summarizer. Re-verifies R1, resolves R9. |
| **1. The note + skill** ✅ (shipped: `skills/context-checkpointing/SKILL.md`, `commands/checkpoint.md`, `commands/resume.md`) | `context-checkpointing` skill (schema + discipline) and `/checkpoint [--show]`. No hooks yet — pure agent discipline. | On a real task, agent maintains a valid `CHECKPOINT.md`; `/checkpoint --show` prints it; validation loop captured verbatim **with each check's trigger surface** (I2). |
| **2. Seal hooks** ✅ **complete 2026-07-29** | Seal on `PreCompact` (snapshot, prune to N, `custom_instructions` fold-in, preserve-verbatim instructions on stdout), `SessionEnd` (**every** reason) and `SubagentStop` (keyed by `agent_id`). Steering is `PreCompact`-only. Skeleton fallback for a stub-only session **not built** — the hook stays silent instead. | Forced `/compact` with a live note → sealed, stamped, `agent_id`-tagged, pruned to 10 (measured across 5 live boundaries). Two concurrent seats → two distinct seals (R10). Same-second collisions across seats, across events, and identical-event → all survive, suffixed rather than overwritten. A session ended without compacting leaves a seal on every reason. Dogfooded against the spike's **real** `SubagentStop` payload, not a synthetic one. |
| **3. Restore + `/resume`** ✅ **built 2026-07-29** | Restore on `SessionStart`, **every** source including `compact` (§3 C); `sc-postcompact-observe` (`PostCompact`, records what each summary kept); `/resume`; `internal/checkpoint` shared by the seal and the restore so the two halves cannot disagree on which file is the note. | A marker injected by the `SessionStart` hook is present in the post-compaction transcript as a `hook_additional_context` attachment (leaf-cited, not inferred); the same marker from `PostCompact` is **absent** — the regression that guards the C2 error, now a unit test. `/resume` prints the full note. Acceptance record: historical §16–§17. |
| **4. Integration** ⚠️ **partial** | Wire into `spec-driven-development` (plan pointer + `beyond_plan`) and `project-memory` (I3 pointer, promotion on completion); `TaskCreated`/`TaskCompleted` enforcement of I1; `FileChanged` re-arming of I2; require it in the sleeper-service run loop. **Done (2026-09-02):** project-memory wired (its `SKILL.md` names `CHECKPOINT.md` and [[context-checkpointing]]); I2 re-arming built (§5; historical §18–19). **NOT BUILT:** I1 enforcement — `TaskCreated`/`TaskCompleted` never fire, REFUTED against a live task and a live seat (`hook-surface-spike.md` §9b); **there is no known enforcement point, so I1 stays discipline.** **NOT DONE:** the sleeper-service requirement — that plugin is still scaffold-only (`plugins/sleeper-service/` holds a manifest and README, no run loop); `spec-driven-development/SKILL.md` carries no checkpoint reference. | A long `/self-improve` headless run survives a forced compaction and resumes on the correct validation step. An actionable that exists only in the note is **refused**, not silently accepted. |
| **5. Freshness** ✅ **graduated** | Staleness *nudge* (non-blocking). **Graduated 2026-09-02 into `plans/checkpoint-freshness.md`**, which demoted `PostToolUseFailure` to a fallback (staleness has nothing to do with a tool failing) and built the gauge plus a `Stop`-channel nudge — merged, but inert until thresholds are chosen there. | Owned by `plans/checkpoint-freshness.md`. |

**The two live threads out of this plan:** Phase 4's sleeper-service and
`spec-driven-development` wiring, and R11's delivery question. I1 enforcement is not a thread —
it is refuted, and reopening it needs a new event, not a new attempt.

---

## 14. One-line summary

*Leave yourself a note before the amnesia hits.* The agent maintains a living
`CHECKPOINT.md` — objective, plan pointer, decisions kept and rejected, working state, next
steps, and above all the **validation loop** — a `PreCompact` hook seals it deterministically
because the hook itself cannot reflect, and tells the summarizer what to preserve; a `SessionStart`
hook hands it back on the far side, on every source including `compact`. So work that has
drifted beyond the plan is never lost to compression.

---

## Where §12 and §§15–20 went

They are in [`plans/historical/context-checkpointing.md`](historical/context-checkpointing.md),
under their original numbers:

- **§12** — field evidence from the six-boundary manual run that produced I1, I2 and I3.
- **§15** — the correction record: the gray-area retarget and its reversal, and hook-surface
  corrections C1–C8.
- **§16** — Phase 3 acceptance record.
- **§17** — the compaction lever and the first digest across a real boundary. (Cited by
  `plans/hook-surface-spike.md`.)
- **§18** — I2's build and verification record.
- **§19** — the loop closed on this repository.
- **§20** — the #165 re-arm experiment that could not run, and why. (Cited by
  `plans/rearm-coverage-experiment.md`.)
