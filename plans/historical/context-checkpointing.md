# Context Checkpointing — archaeology

> The current design is [`plans/context-checkpointing.md`](../context-checkpointing.md). This file
> is the record of what changed and why: the superseded designs, the corrections entered against
> them, the field evidence that forced them, and the acceptance and measurement runs.

**Section numbering is preserved from the original document.** §§1–11, 13 and 14 are the design and
live in the clean file; §12 and §§15–20 are here. Sections cited from elsewhere in the tree —
`plans/hook-surface-spike.md` §17 A, `plans/rearm-coverage-experiment.md` §20, `plans/README.md`
§15 — are all in this file and keep their numbers.

Nothing below is edited to match the tree as it stands. A superseded design is accurate about its
own moment; correcting it would destroy the record of the change, which is the only reason to keep
it.

---

## A. Superseded design statements, verbatim

Each block is what the document used to say, followed by what replaced it. The corrections
themselves are tabulated in §15 (C1–C8); this section preserves the passages that the clean
document no longer carries.

### A1. The central constraint, as originally stated (§2)

> ~~**`PreCompact` cannot reflect, and cannot edit the summary.**~~ … "per the docs it **cannot
> inject content into the compacted summary** (unlike SessionStart, it has no
> `additionalContext`)."

**Superseded by** the 2026-07-27 re-verification against Claude Code 2.1.220 (§15, C1). The second
half is refuted: `PreCompact` exits 0 and its stdout is appended as the custom compact
instructions, so it *can* steer the summarizer. Only the first half survives — it cannot reflect —
and it was always sufficient to carry the design. The clean §2 states the surviving half.

### A2. Restore split onto `PostCompact` — a revision made and then reversed (§3 C, §7)

The 2026-07-27 draft routed the compaction case through `PostCompact`, reasoning that because it
receives `compact_summary` it could inject only the delta the summary dropped. The 2026-07-29
revision reversed it, and recorded the reversal in place:

> **CORRECTED 2026-07-29, and this reverses the previous correction.** An earlier revision of this
> section routed the compaction case through `PostCompact`, reasoning that because it receives
> `compact_summary` it could inject only the delta the summary dropped. **`PostCompact` cannot inject
> anything.** Measured three ways:
>
> - It is **absent from the `hookSpecificOutput` union** in the client. Twenty events have an output
>   shape; it is not one of them, so it has no `additionalContext` field to return.
> - Its documented exit-0 behaviour is *"stdout shown to **user**"*, where `SessionStart` says
>   *"stdout shown to **Claude**"*.
> - **Observed end-to-end:** a marker emitted by a `PostCompact` hook appears **nowhere** in the
>   resulting transcript, while the same marker from `SessionStart` materialises as a
>   `hook_additional_context` attachment and the model reports seeing it. See
>   `plans/hook-surface-spike.md`.
>
> Ordering kills it independently anyway: the per-boundary sequence is
> **`PreCompact` → `SessionStart(compact)` → `PostCompact`**, so the only hook that *can* inject runs
> *before* the summary exists. No two-hook relay can work in that order.

§7 carried the same reversal in its own words:

> **CORRECTED TWICE; this is the measured version.** The first draft ran every restore through
> `SessionStart`. A later revision split it, routing compaction through `PostCompact` so the injection
> could be diffed against `compact_summary`. That revision was wrong — `PostCompact` cannot inject
> (§3 C) — and the first draft was right.

**Superseded by** the single-hook restore in the clean §3 C and §7: `SessionStart` owns every source
including `compact`; `PostCompact` is observability only.

### A3. Non-duplication as a mechanism (§3)

> That draft had `PostCompact` read the returned summary and inject only the delta.

**Superseded by** non-duplication as a *policy*: the digest is written before the summary can be
inspected, so terseness is enforced by rule per boundary and only *measured* across boundaries by
the observer. That is R4, and it is live.

### A4. The injection-suspicion mitigation, withdrawn (§3 C)

> ~~Twice measured: a bare token injected this way was flagged by the model as a suspected
> prompt-injection attempt, and it said so in its reply. A restore payload that reads as unexplained
> foreign text invites exactly that reaction, at the moment the agent most needs to trust it.~~

And the framing it rested on:

> *"a checkpoint the resumed agent distrusts is worse than no checkpoint"*

**Superseded by** the 2026-07-29 acceptance runs (§16 B): the digest led with its provenance, quoted
the note verbatim, and was flagged anyway. The flag cannot be designed away while the note carries
foot-guns. What replaced the mitigation is the enforceable residue — *the hook adds no imperative of
its own*, now a unit-tested rule — plus the finding that a checkpoint the agent distrusts **and
uses** is the design working.

### A5. Schema 2 — the front-matter and note as originally proposed (§4)

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

**Superseded 2026-09-02 (front-matter only)** by `schema: 3`: `updated:` is retired, replaced by
`written_at`/`reaffirmed_at`, per `plans/checkpoint-freshness.md` §III. The section list also grew
— `In-flight handles` and `Invariants / foot-guns` — and the normative copy is now
`skills/context-checkpointing/SKILL.md`, not this document.

### A6. "Symlink or copy" for the live note (§6)

The original offered symlink *or* copy for the live-note location.

**Superseded by** *copy, never symlink* (§15, C8): symlink creation on Windows requires developer
mode or elevation, Windows is the primary development box, and [[agent-guardrails]] forbids
resolving that by escalating.

### A7. The rejection of "hook-only, no agent discipline" as originally argued (§10)

The original rejected it because `PreCompact` "cannot reflect **and cannot edit the summary**."

**Superseded by** the narrower rejection: the second clause was wrong (C1), and the rejection
survives on the first clause alone — a script's note is a mechanical transcript slice, and the
ability to say *"keep the validation loop"* is worthless if nothing wrote the validation loop down.

### A8. The estimate (§13)

> **Estimate: withdrawn.** The original said ~2–3 working days on the reading that this was a
> Markdown file plus two thin hooks. It is now seven hooks across five events, one of them
> enforcing a cross-index invariant, inside a plugin that does not yet exist. Re-estimate after
> Phase 0, against the placement decision in §8. An estimate carried forward unchanged through two
> retargets is not an estimate.

**Superseded by** the build itself: Phases 0–3 shipped, and the plugin the estimate said did not
exist is `prosthetic-conscience`, which it shipped into. No re-estimate was ever made.

### A9. Phase 5, as originally written (§13)

`plans/checkpoint-freshness.md` opens by quoting this row *"in full"*, so it is preserved here
verbatim:

> Staleness *nudge* (non-blocking), preferring `PostToolUseFailure` over a mutation counter,
> only if Phase 4 shows stale seals.

with the verify column:

> Nudge fires on failure runs; no interference with agent writes; measurably fresher seals.

**Superseded 2026-09-02** by `plans/checkpoint-freshness.md`, which demoted `PostToolUseFailure` to
a fallback — staleness has nothing to do with a tool failing — and built a staleness *gauge* first,
on the ground that the original sentence *"has never been actionable, because nothing records how
stale a seal was."* The gauge and a `Stop`-channel nudge are merged; thresholds are unset, so the
nudge is inert.

### A10. Phase 2 and Phase 3, as built before the per-event re-cut (§13)

Phase 2 shipped as **one** `sc-checkpoint-seal` binary serving three events, with `-event` passed
by `hooks.json` rather than inferred from the payload (the spike never verified `hook_event_name`
exists, and this plan had already spent a cycle on a capability inferred from an input field).
Phase 3 shipped `sc-checkpoint-restore` as its own binary.

**Superseded by #201 step 3**, which re-cut the binaries to **one per event**: the seal became
`sc-precompact` / `sc-sessionend` / `sc-subagentstop`, and the restore became the
`internal/checkpointrestore` unit composed into `sc-sessionstart`. The `-event` flag existed only
to work around the rule it inverted; each binary now knows its event by construction, and the drift
the merged binary avoided is prevented by the shared unit instead.

### A11. Cross-platform hooks as shell variants (§9)

> Hooks are cross-platform per the suite convention (PowerShell + POSIX variants; the port
> plan already establishes `Get-Command`/`command -v` capability gating).

**Superseded by** compiled Go binaries plus a single POSIX bootstrap guard. No `.ps1` variant
exists anywhere in the suite; `hooks.json` records that the guard shell is Git Bash on every
platform, verified live — the 0.7.0 failures came from `/usr/bin/bash`.

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
- ~~**`CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` is not a usable lever.**~~ ~~**Wrong; it is, at `PCT=10`.
  See §17 A**~~ — the failure was using it at 1 and 2, where the headroom above the fixed ~30.6k floor
  is a few thousand tokens and the next boundary is one turn away whatever the workload does.
  **Both readings are now obsolete: on client 2.1.240 the variable is inert at every value tested
  (2/5/10/25 — zero boundaries with ~102k of context reached in each run).** The lever is
  `--autocompact <auto|tokens>`; evidence in `hook-surface-spike.md` §11.
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

> **OBSOLETE ON 2.1.240 (measured 2026-08-22, `hook-surface-spike.md` §11).** The variable does
> nothing at any value now: the same sweep shape run at 2/5/10/25 produced **zero** compactions while
> every run reached ~102k tokens of context. Use `--autocompact <tokens>` instead — a flag that did
> not exist when this section was written.
>
> **The section is kept in full, because its method is what caught this.** Its own lesson — *"a clean
> exit says the run finished; it says nothing about where the threshold was"* — is exactly why §11
> asserts the boundary from the transcript rather than from an exit code. This is the third revision
> of this claim, and the failure moved: it was the value twice, and now it is the lever. A control
> surface can disappear between patch versions while every test built on it keeps exiting 0.

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

**RESOLVED 2026-09-02:** #165's cause turned out to be concurrent events destroying each other's
records — fixed on main in `f32c2815` (`fix(rearm): concurrent events were destroying each other's
records (#165)`), with the transient-miss reader hardened in `a1d3a52f` (#567). The experiment below
is no longer needed.

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
