# Checkpoint freshness — making the note's staleness measurable

> Phase 5 of [`context-checkpointing.md`](context-checkpointing.md) §13, which reads in full:
> *"Staleness **nudge** (non-blocking), preferring `PostToolUseFailure` over a mutation counter,
> only if Phase 4 shows stale seals."* That sentence has never been actionable, because nothing
> records how stale a seal was. This document builds the record first and the nudge on top of it.
>
> Evidence: [`hook-surface-spike.md`](hook-surface-spike.md) (measured payloads),
> [`gray-area.md`](gray-area.md) §3 (the client's 31-event catalogue), and the measurements in
> §II, taken 2026-08-22 against a live transcript and this repository's own history.
>
> **Scope is deliberately narrow, and the boundary is stated rather than implied.** Everything
> here rides channels already *measured* to reach the model. Every proposal that needs an
> unverified event, or an unmeasured behavioural claim, was moved out to a filed issue and is
> listed in §VI — `complete-the-concept`: scoping down is legitimate, scoping down silently is
> how the class survives.

---

## I. Summary & Goals

**The problem.** Every trigger in the checkpointing design is deterministic except the one that
decides whether the sealed note is worth anything. `PreCompact`, `SessionEnd` and `SubagentStop`
all seal reliably — and a reliable seal of a note written 300 turns ago preserves a stale cursor
with perfect fidelity. R2 named this and offered "add the nudge if staleness is observed." It has
never been observed, because no seal record carries the note's age.

**What this is not.** No automatic note writing: a script inventing a forward intent produces a
file that reads exactly like a real checkpoint, which is the `facts-are-fields` failure where the
absent case and the healthy case are the same bytes. No blocking — every path exits 0, matching
`sc-strike-counter`'s contract.

**Success criteria.**

1. **Gauge before nudge.** Every seal record carries the note's age in three units — context
   tokens grown, assistant turns, and commits landed on this branch. Baseline over ≥ 20 real
   boundaries **before any threshold is chosen**.
2. **No fabricated denominator.** Zero emissions containing a percentage-of-window on a session
   where the window is not known. Asserted as a test over the render, not as an intention.
3. **Cost.** ≤ 5 ms per gauge read at p95 on a 13 MB transcript (the size measured in §II).
4. **Nudge budget.** ≤ 200 bytes per emission, ≤ 4 emissions per session, counted in the
   observation row rather than intended.
5. **Zero extra turns.** A session with the nudge live shows no more assistant entries than the
   same session without it, beyond the one turn a single emission is intended to create. The
   control in §II is the baseline (4 entries clean, 20 looping).
6. **The falsification.** After the nudge ships, median note-age-at-seal falls against criterion
   1's baseline. If it does not, the nudge comes out — stated before the data so the outcome
   cannot be reinterpreted after it.

---

## II. Technical Context

Go hook binaries under `plugins/prosthetic-conscience/tools/`, one binary per event, units as
shared `internal/` packages (#201 step 3). No new dependency: `git`, already required.

### Injection is verified on six events — and `Stop` re-arms the turn

`SessionStart` (spike §3a), `PostToolUseFailure` (§2a), and — **measured 2026-08-22 for this plan,
on client 2.1.235, recorded as spike §8** — `UserPromptSubmit`, `PostToolUse`, `PostToolBatch` and
`Stop`. All four arrived as `hook_additional_context` attachments alongside a `SessionStart`
positive control. `PostCompact` remains the standing counter-example: it carries the whole
compaction summary and cannot inject a byte of it, which is why none of these was assumed.

**`Stop` injection re-invokes the model.** Nine `Stop` firings from one prompt; the transcript
shows the turn ending, the hook injecting, the model being re-invoked with nothing to do, and
`Stop` firing again. A null control — same project, same prompt, a `Stop` hook emitting nothing —
settles that the injection causes it:

| Run | `Stop` firings | assistant entries | outcome |
|---|---|---|---|
| Stop emits `additionalContext` | **9** | **20** | 8 filler turns, 1,186 wasted output tokens |
| Stop emits nothing | **1** | **4** | one answer, clean exit |

`PostToolUse` and `PostToolBatch` injected once each with no loop. The pathology is specific to
`Stop` — and it is also the reason `Stop` is the right nudge channel: **the hazard and the feature
are the same mechanism.** A guarded single emission hands the model exactly one extra turn in which
to act on what it was told. No other event does that: `PostToolUse` injects into a turn already
committed to its next action; `Stop` injects at a boundary and then creates a turn.

### The numerator is exact; the denominator is not available

Measured 2026-08-22 on a live 13 MB / 7,308-line transcript, client 2.1.235:

| Question | Answer | How |
|---|---|---|
| Current context size | `usage.input_tokens + cache_read_input_tokens + cache_creation_input_tokens` on the last `type:"assistant"` entry — **913,789** | `tail -c 200000 <t> \| grep '"usage"' \| tail -1` |
| Cost to read it | **14 ms** in shell; a Go seek of the last 200 KB is well under | `time` |
| The context window | **Absent from every field.** `message.model` is `claude-opus-5` on a session whose model is `claude-opus-5[1m]` — the string does not distinguish the 1M variant from the 200k one | `jq -r '.message.model' \| sort -u` |
| Any `*limit*` field | **none** | `grep -o '"[a-zA-Z_]*[Ll]imit[a-zA-Z_]*"'` |
| Where compaction fired | `type:"system", subtype:"compact_boundary"` → `compactMetadata.{trigger,preTokens,postTokens,cumulativeDroppedTokens}`; here **preTokens 1,001,875 → postTokens 12,823** | `jq -c 'select(.compactMetadata)'` |

So the fraction-of-window cannot be computed until a session has compacted once — after which
`preTokens` gives that session's trigger point exactly. A hardcoded 1M denominator would print
91% on one session and 457% on another, both rendered as confident numbers. `facts-are-fields`
clause 3 governs: **a percentage we cannot compute must not be printed.**

### Work done is a branch-line count, not a HEAD distance — measured

The shipped restore computes staleness as `git rev-list --count <note-head>..HEAD`
(`internal/checkpointrestore/main.go:293`). That counts every commit reachable from HEAD and not
from the note — which is dominated by *other people's work arriving*, not by work this session
did. On this repository, from the merge-base with `main` to `HEAD`:

```
git rev-list --count               <merge-base>..HEAD   →  109
git rev-list --count --first-parent <merge-base>..HEAD   →   24
```

**85 of those 109 are merged-in side branches.** A note written before a routine `main` merge
would be reported as 100+ commits stale having done nothing. `--first-parent` walks this branch's
own line and answers the question that matters: how much work landed *here* since the note.

**Do not filter by author.** Re-measured 2026-08-22 across the same range in two worktrees, which
agree: `--author=$(git config user.email)` (`gblock+agent@ctoforaday.com`) returns **1** of 109 —
the single commit `13332bc` — against `noreply@anthropic.com` (85) and `gblock@ctoforaday.com`
(23). The configured identity is not the identity the branch's work was committed under, and the
filter reports 1 for a branch that this pair did all 109 commits of.

> An earlier draft of this paragraph said the filter returns **0**. It does not, and did not: the
> range yields 1 in the worktree the original measurement was taken in. The claim is corrected
> rather than quietly restated — a number in this record that fails on re-run is the defect the
> record exists to prevent.

Near-zero rather than zero makes the case worse, not better: it is a plausible *number*, not even a
suspicious blank. Clause 3 governs — the miss and the honest "did no work" are the same bytes.
First-parent needs no identity and has no such miss.

**Commits are coarse, and are therefore secondary.** A session can do a great deal with nothing
committed — this worktree has uncommitted changes right now. Growth and turns stay primary;
commits corroborate. Comparing the note's prose "Files touched" section against `git status` was
considered and rejected: recovering a fact from prose beside a record that already holds it is the
defect this suite is named after.

### Transcript reading and the plugin boundary

`README.md`: *"Reading transcripts is a surveillance capability; keeping a note about your own work
is not."* This stays on the prosthetic-conscience side — `transcript_path` from the payload, for
this session only. Precedent: `internal/transcript`, `checkpointseal`, `postcompactobserve` all
already declare it. A cross-session sweep of `~/.claude/projects/` would have given the denominator
problem a better answer and is **rejected on this boundary alone**, recorded here so a later author
finds the refusal instead of re-deriving it.

---

## III. Proposed Changes (the spec)

```
plugins/prosthetic-conscience/tools/
├── internal/
│   ├── ctxusage/           [NEW]    transcript tail → {tokens, turns, ceiling|Unknown}
│   ├── freshness/          [NEW]    the gauge: three measures, bands, debounce, one-line render
│   ├── checkpointrestore/  [MODIFY] --first-parent defect; gauge in the restore digest
│   ├── checkpointseal/     [MODIFY] stamp note-age onto every seal record
│   ├── stopnudge/          [NEW]    the guarded single emission at the turn boundary
│   └── postcompactobserve/ [MODIFY] carry note-age-at-seal into the observation row
└── cmd/
    ├── sc-stop/            [NEW]    Stop — the nudge channel (one binary per event, #201 step 3)
    └── sc-posttooluse/     [MODIFY] fan in the gauge as a cheap tick (matcher union widens)
```

### `[NEW] internal/ctxusage` — the measurement

Bounded backward scan of the last N KB of `transcript_path` for the most recent `type:"assistant"`
entry and any `subtype:"compact_boundary"`. Widen once if no assistant entry is found, then report
`Unmeasured` — a hook must not read 13 MB on a tick, and must not hang on a rotated or truncated
file. Returns `Tokens` (exact), `Turns` (assistant entries after a given time), and `Ceiling`,
which is `compactMetadata.preTokens` from **this session's own** most recent boundary or the
distinct value `Unknown`.

`Unknown` is a case the type forces every caller to handle. It is never defaulted, never
substituted, never a zero. This is clause 3 made structural rather than remembered.

### `[NEW] internal/freshness` — the gauge

Three measures, independent because they fail independently: a session can burn 400k tokens in
twelve turns of bulk reading, or take 300 turns without moving the token count.

| Measure | Definition | Needs ceiling? |
|---|---|---|
| **Growth** | context tokens now − context tokens when the note was written | no |
| **Turns** | assistant turns since the note was written | no |
| **Branch work** | `git rev-list --count --first-parent <note.head>..HEAD` | no |
| *Proximity* | `Tokens / Ceiling` — **emitted only when `Ceiling` is known**, with its basis named | yes |

**Reference point is `CHECKPOINT.md`'s mtime** — a fact the filesystem holds, that no writer can
forget to update. The `updated:` frontmatter is read as a cross-check and a disagreement between
the two is itself reportable, but mtime is the authority. **No new record is created to hold "when
we last checkpointed"**; the note and the transcript already hold it.

**One new file, holding only debounce:** `.claude/checkpoints/freshness.json`, keyed by
`(session_id, agent_id)` — subagents share the parent's session id and two seats would otherwise
silence each other (R10, a defect in the original design). Deleting it costs at most one duplicate
emission. Bands are `NOTICE → WARN → URGENT`, at most once per band per session, all reset when
the note's mtime moves. **Thresholds are set from the Phase 1 baseline and ship as `TBD`**; a
threshold chosen before the distribution is known is a guess, and this plan will not launder one
as a default.

**Render:** one line, ≤ 200 bytes — the measure, the number, the note's path. No instruction. Spike
§3b twice observed an injected directive treated as a suspected prompt injection; the line reads
as the session's own recovered state.

### `[MODIFY] internal/checkpointrestore` — the `--first-parent` defect

`commitsSince` gains `--first-parent` and the surrounding prose changes from "commits ago" to
"commits on this branch". This is a **correctness fix to shipped behaviour**, evidenced in §II
(109 vs 24), and it stands independently of everything else here. The restore digest additionally
carries the gauge — `SessionStart` is the verified channel, and this is where a resumed session
already reads its own provenance.

### `[MODIFY] internal/checkpointseal` — the gauge becomes a record

Every seal record gains `note_age_turns`, `note_growth_tokens`, `note_branch_commits`,
`ceiling_known`. This is the whole of criterion 1 and the prerequisite for criterion 5.

**Plus two observation-only fields, folded in from the census (§VI-a).** Both are reads of a payload
the sealing binaries already receive; neither emits anything, and neither needs a channel verified to
reach the model — which is why they can ride Phase 1 while the issues they come from stay out of
Phase 2.

| Field | Source | Question it makes answerable |
|---|---|---|
| `live_handles` | `len(background_tasks) + len(session_crons)` on the `Stop`/`SubagentStop` payload — measured non-empty on a live task, `hook-surface-spike.md` §9c | Does a note sealed while background work was in flight turn out to have been wrong more often? #506's premise, currently assumed. |
| `seal_trigger` | which event sealed: `precompact` \| `sessionend` \| `seat_return` | How much staleness a *seat's return* actually represents. #507 has no threshold data, and this is the cheapest way to get some. |

`seat_return` is nearly free: `cmd/sc-subagentstop` is already a `checkpointseal` consumer, and §9d
established that its payload names the **parent's** `session_id` and `transcript_path` alongside the
seat's own — so the seal it writes is the parent's note, distinguished by field rather than inferred.

**Neither field licenses a nudge.** `SubagentStop` cannot inject at all (§10), and #506's strong
claim — *the note's handles are provably wrong* — needs handle ids in a field the note does not have
(§VI-b). These two rows exist to decide whether either is worth building, which is what Phase 1 is
for.

**Consumer census** — `grep -rn "checkpointseal" --include=*.go plugins/` → `cmd/sc-precompact`,
`cmd/sc-sessionend`, `cmd/sc-subagentstop`, and the package's own tests. No external reader; fields
are additive. Re-run at build time: a census pasted into a plan and not re-run is the defect §III
exists to prevent.

### `[MODIFY] internal/postcompactobserve` — the same fields on the row

**Consumer census** — `grep -rln "postcompact\|observe" --include=*.go --include=*.md plugins/
scripts/` → `cmd/sc-postcompact-observe`, the package and its tests, plus prose mentions in
`README.md` and three `SKILL.md` files. The row is JSONL and nothing parses it positionally.

**"The same fields" includes the two folded in above** — `live_handles` and `seal_trigger` — and
this is not a detail: §V's Phase 1 queries read `.claude/checkpoints/observe.jsonl`, so a field
stamped onto the seal but never carried to the row makes `jq` return an empty group. **An empty
group renders identically to "no difference between the two populations", which is the answer those
queries exist to test for** — the analysis would report the null result it was built to detect, and
be believed. A Phase 1 test MUST assert both fields are present on a written row, not merely on a
seal.

### `[NEW] internal/stopnudge` + `cmd/sc-stop` — the nudge channel

`Stop` is the turn boundary: the natural unit of the operator's "it's been 100 turns", ~10× cheaper
than a per-tool tick, and the only channel that gives the model a turn in which to respond.

**The guard is safety-critical, not politeness.** Measured (§II): an unconditional `Stop` injector
looped nine times and burned 1,186 output tokens on filler. Therefore:

- **Write before emit.** The band is recorded as spent **before** the context is returned, never
  after. A guard that writes afterwards, or crashes in between, re-emits — and a re-emission on
  `Stop` is not a duplicate nudge, it is a loop.
- **`stop_hook_active`** is present in the payload (spike §2) and is checked as a second,
  independent brake: if the turn is already a Stop-hook continuation, emit nothing regardless of
  band state. Belt and braces, because the two failure modes have different causes — one is our
  state file, one is the client's own re-entry.
- A **loop regression test** asserts the null case: given a band already spent, the emission is
  empty. This is the test that must never be deleted.

`sc-posttooluse` additionally fans in the gauge as a cheap tick — it updates the measurement so
`Stop` can decide quickly, and it does **not** emit. One writer of the nudge, one event.

The parent plan's Phase 5 preferred `PostToolUseFailure`. That is now demoted to a fallback and
was, on the evidence, the wrong instinct for a good reason: it was the only injector known at the
time. Staleness has nothing to do with a tool failing.

### `[MODIFY] skills/context-checkpointing/SKILL.md`

> - AFTER a freshness nudge, YOU MUST either write the note or state why the current note is still
>   accurate. The nudge measures the note's **age**, never its **truth**: a note can be 300 turns
>   old and exactly right because nothing changed, and 3 turns old and already wrong because the
>   last turn changed the objective. Silence is the failure this mechanism exists to remove; a
>   reasoned "still accurate" closes the band and is a valid answer.

---

## IV. Risk & Mitigation

| # | Risk | L×I×C | Mitigation | Step |
|---|---|---|---|---|
| F1 | **The nudge becomes wallpaper** — and a context-pressure warning that costs context is self-defeating at exactly the moment it matters. | H×H×M | Bands, once per band; ≤ 200 bytes; reset on write; **criterion 5 removes it** if the baseline median does not fall. | Ph. 3 |
| F2 | **Fabricated denominator.** | M×H×L | `Ceiling` tri-state with `Unknown` as a value callers must handle; criterion 2 is a test over the render. | Ph. 1 |
| F3 | **Commit count reads other people's work as mine.** | H×M×L | `--first-parent`; no author filter (§II measured both). | Ph. 1 |
| F4 | **Gauge cost on a hot path.** | M×M×L | Bounded tail; skip entirely when note mtime is unchanged and the band is current; criterion 3 is a measured gate. | Ph. 2 |
| F5 | **Concurrent seats share `session_id`.** | M×M×L | Debounce keyed on `(session_id, agent_id)`; R10 regression test. | Ph. 2 |
| F6 | **mtime lies** — a tool touches the note without rewriting it. | L×H×M | `updated:` cross-check; disagreement is reported rather than resolved silently. | Ph. 1 |
| F7 | **Short sessions get lectured.** | M×L×L | Gauge arms only above a floor (a note exists, or the session crossed a turn/token floor). Below it, silent. | Ph. 3 |
| F8 | **Age is read as truth.** A fresh note is not a correct note. | M×M×L | The skill clause says so; the render states a measure, never a verdict. gray-area's `/audit-checkpoint` remains the instrument for the note's *claims*, and the two are deliberately different tools. | Ph. 3 |
| **F10** | **`Stop` injection loops** — measured, not theorised: 9 firings, 8 filler turns, 1,186 wasted tokens from one unguarded emission. | **L×M×L** *(was H×H×L)* | Write-before-emit; `stop_hook_active` as an independent second brake; a loop regression test asserting the empty emission on a spent band. §III. | Ph. 2 |
| F9 | **Hook-surface churn** (R9, realised once already). | M×M×M | Every channel here is now measured on 2.1.235; each binary is inert rather than broken when its event never fires. | all |

---

## V. Verification Plan

### Phase 1 — gauge, defect fix, baseline (no nudge)

```bash
(cd plugins/prosthetic-conscience/tools && go test ./internal/ctxusage/... ./internal/freshness/... \
    ./internal/checkpointrestore/... ./internal/checkpointseal/... && go test ./...)
```

Coverage must include the negatives, which are the point: a tail containing **no** assistant entry;
a truncated final line; a session with **no** compact boundary → `Ceiling: Unknown` **and no
percentage in the render**; two boundaries → most recent wins; mtime and `updated:` disagreeing; a
note head that is unreachable; a branch whose history contains merges → first-parent and plain
counts differ, asserted against fixed numbers.

**Driveable check on real data.** Run the gauge against the live 13 MB transcript from §II and a
real `research/<slug>/CHECKPOINT.md`, and check the token figure by hand against
`tail -c 200000 … | jq '.message.usage'`. Then, in this repository:

```bash
git rev-list --count --first-parent <merge-base>..HEAD    # expect 24, not 109
```

Fixtures prove the parser; only the real file surfaces the data-shaped defects (ANSI escapes in
content, sidechain entries, `iterations[]` nesting, 13 MB of it).

**Then collect the baseline** — ≥ 20 real boundaries, seals stamped, nothing emitted:

```bash
jq -s 'map(select(.note_age_turns)) | {n:length,
        turns:(map(.note_age_turns)|sort), growth:(map(.note_growth_tokens)|sort),
        commits:(map(.note_branch_commits)|sort)}' .claude/checkpoints/observe.jsonl
```

And the two folded-in observations, which are read the same way and answer their own questions:

```bash
# #506: is a note sealed with live background work staler than one sealed without?
jq -s 'group_by(.live_handles > 0)
       | map({live_handles: .[0].live_handles > 0, n: length,
              median_age: (map(.note_age_turns) | sort | .[length/2|floor])})' \
   .claude/checkpoints/observe.jsonl

# #507: how stale is the note at a seat return, against the other seal triggers?
jq -s 'group_by(.seal_trigger)
       | map({trigger: .[0].seal_trigger, n: length,
              median_age: (map(.note_age_turns) | sort | .[length/2|floor])})' \
   .claude/checkpoints/observe.jsonl
```

**Both may come back "no difference", and that is a result.** If `live_handles > 0` does not track
with staler notes, #506's premise is wrong and its strong form should not be built at any price; if
`seat_return` is not staler than `sessionend`, #507 has no case. Stated before the data, so the
outcome cannot be reinterpreted after it — the same contract criterion 6 makes for the nudge itself.

The distribution sets the Phase 2 thresholds. This phase is shippable alone: the `--first-parent`
fix is a correctness change and the stamping is pure observation.

### Phase 2 — the nudge on `Stop`, guarded

Thresholds from Phase 1. Cost gate on the 13 MB transcript, p95 over 100 invocations, **fail above
5 ms**.

**The loop gate comes first, because it is the one that burns tokens.** Reproduce spike §8's null
control against the real binary: a scratch project, one prompt, band already spent → **exactly one
`Stop` firing and ≤ 4 assistant entries**. Then band unspent → **exactly two firings** (the
emission, and the boundary of the turn it creates), never more.

Injection is re-confirmed rather than assumed — the attachment is the leaf:

```bash
jq -c 'select(.attachment.type=="hook_additional_context")
       | {e:.attachment.hookEvent, c:.attachment.content}' <transcript>
```

Live acceptance: a session driven to compaction with **`--autocompact 100k`**. The sweep this
sentence used to prescribe — `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE=10` — was reproduced on 2026-08-22 as
this instruction demanded, and **the variable is inert on client 2.1.240 at every value tested**
(spike §11: 2/5/10/25, zero boundaries, ~102k context reached in each). The instruction to reproduce
rather than trust is what caught it; it stands, and now applies to this sentence too.

Two properties, both from §11, that this acceptance test MUST respect:

- **Assert the boundary from the transcript** — `jq -c 'select(.subtype=="compact_boundary")'` — never
  from a clean exit or from hook silence. A disconnected lever produces both, and Phase 2 would
  report a passing acceptance run in which no compaction ever occurred.
- **Boundary count follows the workload, not the flag.** Six files gave one boundary at `100k`,
  fifteen gave four. "One clean boundary" is tuned per workload and asserted, never assumed. Assert the nudge attachment precedes
the boundary, and assert the negative: a short session below the F7 floor emits **zero**.

### Phase 3 — the falsification

Twenty more boundaries with the nudge live.

```bash
jq -s '{before:[.[]|select(.nudge_enabled==false)|.note_age_turns]|sort,
        after: [.[]|select(.nudge_enabled==true )|.note_age_turns]|sort}' \
   .claude/checkpoints/observe.jsonl
```

If the median does not fall, the nudge is removed and the gauge stays.

### The gate

`/prosthetic-conscience:plan-audit` before any code. `versionguard` unaffected — plugin content,
and the version moves at a release boundary.

---

## VI. Deliberately not in this plan

Filed as issues, not built here. Each needs either an unverified event or a behavioural claim no
measurement supports yet; carrying them in the plan would let unverified work ride an approved
document.

> **Superseded in part, 2026-08-22.** All four have since been measured (§VI-a). The table below is
> kept as written — it is the reason each item was deferred, and a deferral whose stated reason has
> expired is worth seeing next to what replaced it. Read §VI-a for the current verdicts. None of
> them has been folded into §III.

**#505 was answered before this plan was audited, and folded in** — all four candidate channels
inject, `Stop` is the one worth having, and its loop hazard is now a named risk (F10) with a test
rather than a discovery waiting to happen in Phase 2. Evidence: `hook-surface-spike.md` §8.

| Filed | Why it is out |
|---|---|
| **#506** — `TaskCreated` / `TaskCompleted` → "In-flight handles" is provably wrong | The best semantic trigger in the catalogue — the note's section is not merely old, it is false. The event never fired in the spike and its injection is unverified. |
| **#507** — `SubagentStop` at the **parent** | A seat returned a large result and the lead is about to decide on it. The seat's own note is sealed; the parent's is not. |
| **#508** — remaining catalogue signals: `StopFailure`, `PermissionDenied`, `CwdChanged`/`WorktreeCreate`, `ConfigChange`, `InstructionsLoaded`, `FileChanged`-armed-but-not-re-run | Each invalidates a named part of the note. Enumerated so the omission is stated; none measured. |
| **#509** — model → context-window table for the denominator | Would answer F2 cheaply **if** `SessionStart`'s `model` field distinguishes `[1m]`, which is unmeasured. A hand-kept model→window table is a maintained copy of someone else's record, and the cost is not paid silently. |

### VI-a. The whole docket, measured — census of all 30 events on 2.1.240

The table above holds the reasons each item was deferred; **every one of those reasons was "not
measured", and all four are now measured.** Evidence: `hook-surface-spike.md` §9 and §10 — 30 events
registered, 16 headless sessions, four passes.

**What was folded in, and what was not.** Two observation-only fields reached §III — `live_handles`
and `seal_trigger`, both reads of a payload the sealing binaries already receive. **Nothing reached
Phase 2**: no new emission, no new channel, no threshold. The distinction is the plan's own
admission rule — observation needs no channel verified to reach the model, a nudge does — and it is
what keeps this from being a scope increase wearing a census as justification.

| # | Verdict | Basis | Cost if folded in |
|---|---|---|---|
| **#506** | **QUALIFIED in its weak form; the strong form is blocked** | `TaskCreated`/`TaskCompleted` never fire (§9b), but `Stop` carries `background_tasks[]` — `{id, type, status, description, command}` — and `session_crons[]`, verified against a live task (§9c). | **Easy to read, expensive to compare — and an earlier draft of this row said only "Easy".** Counting live handles is free. Deciding the note's handles are *false* means matching those ids against a `## In-flight handles` section written in **prose**, which is the move this plan already refused for `git status`. Weak form folded into Phase 1 (`live_handles`, §III); strong form blocked on §VI-b. |
| **#507** | **OBSERVATION ONLY — refuted as a nudge channel** | `SubagentStop` fires at the parent and names both transcripts (§9d), so it can *record*. It cannot *tell*: **9 firings, marker delivered nowhere**, against a live positive control (§10). Worse, the emission re-arms the **seat's** turn — 9 assistant entries for a one-word answer — so an injector here costs exactly what `Stop`'s loop costs and delivers nothing. | **Free as a stamp, unavailable as a nudge.** Sealing/gauging at a seat's return needs no injection and can ride Phase 1. Anything that wants to *reach the model* when a seat returns has no channel, and no threshold data either. |
| **#508** | **SPLIT — four fire, three refuted, six unreachable** | Fire with usable payloads: `ConfigChange` `{source, file_path}`, `InstructionsLoaded` `{file_path, memory_type, load_reason}`, `CwdChanged` `{old_cwd, new_cwd}`, `FileChanged` `{file_path, event}`. Refuted for the shapes tested: `StopFailure` (hook exit 1), `PermissionRequest`/`PermissionDenied` (hook-issued deny). Not reachable in a headless harness: `Elicitation`/`ElicitationResult`, `Setup`, `TeammateIdle`, `WorktreeCreate`/`WorktreeRemove`, `DirectoryAdded`, `UserPromptExpansion` (§9b). | **Hard, and mostly not worth it.** Four usable signals, each invalidating a *named* part of the note — but each is one more emission path competing for the same ≤ 4-per-session budget that F1 already calls the main risk. |
| **#509** | **CLOSED, negatively** | `SessionStart`'s payload has no model field at all; `message.model` reads `claude-opus-5` on a live `claude-opus-5[1m]` session; no `*limit*` field exists anywhere in a transcript (§9e). | **Nothing to fold in.** F2's tri-state `Ceiling` with `Unknown` is not a conservative choice pending better data — it is the only correct one, and `compactMetadata.preTokens` remains the sole denominator. |

### VI-b. Blocked: the note has no field for its handles

#506's strong claim — *the note's "In-flight handles" section is provably wrong* — needs handle ids
the note does not carry. `CHECKPOINT.md` **already is structured data**: versioned frontmatter
(`schema: 2`) with `updated`, `head`, `session_id`, `agent_id`, `objective`, `plan`, `beyond_plan`,
`status`, read by one parser (`internal/checkpoint.Parse`) and four binaries. `head` is the
precedent — it exists precisely to make the note's age falsifiable. Handles are the same kind of
fact and belong in the same place.

**A `checkpoint.json` sidecar is the wrong answer** and is rejected here so a later author finds the
refusal rather than re-deriving it: it invents a second record where one exists, breaks the skill's
"one block, overwritten" contract, and drifts silently — a fresh `.md` beside a stale `.json` reads
exactly like a healthy checkpoint.

**But the field cannot simply be added, for a reason in the parser.** `Parse` is *deliberately* not
YAML — flat scalars only, and **a line that does not look like `key: value` is skipped rather than
raised**, so that a restore never refuses a slightly malformed note. That forgiveness is correct for
scalars, where a skipped line loses one fact. For handles it **fabricates "no in-flight work"** —
the precise claim the check turns on, with the miss and the honest zero identical, which is clause 3
head-on. A comma-joined string in one scalar reproduces the same defect inside the field.

So the strong form costs: `schema: 3`, a `Parse` that can report a parse failure instead of skipping,
and carriers at `context-checkpointing/SKILL.md`, the four consumers, gray-area's `claims`/`manifest`
readers, and the goldens. **That is a project, not a field**, and Phase 1's `live_handles` column
exists to say whether it is worth starting.

**What the census changes about the plan's own risks.** **F10 has been re-rated H×H×L → L×M×L.**
Not because the measurement changed — 9 firings and 1,186 wasted tokens stand — but because the
hazard is cycle detection, and two independent brakes are already specified: `stop_hook_active` is a
field the client sets on every re-entry (false on the first firing, true on the eight that followed,
on `Stop` and `SubagentStop` alike), needing no state and no write, and the debounce file carries
band policy, whose worst failure is one duplicate emission. The original rating priced an *unguarded*
injector, which nothing here proposes. The mitigations and the regression test are unchanged; a
likelihood term that survives its own mitigation is a rating that never comes down. **F9 gains a measured instance rather than a theoretical one:** `SubagentStop`'s behaviour
on 2.1.240 contradicts this record's own 2.1.220 reading (§9e correction 2), so "each binary is
inert rather than broken when its event never fires" now has a case where the event fires
*differently* — inertness does not cover that, and no mitigation here does either.
