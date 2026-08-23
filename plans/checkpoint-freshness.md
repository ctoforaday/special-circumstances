# Checkpoint freshness — making the note's staleness measurable

> Phase 5 of [`context-checkpointing.md`](context-checkpointing.md) §13, which reads in full:
> *"Staleness **nudge** (non-blocking), preferring `PostToolUseFailure` over a mutation counter,
> only if Phase 4 shows stale seals."* That sentence has never been actionable, because nothing
> records how stale a seal was. This document builds the record first and the nudge on top of it.
>
> Evidence: [`hook-surface-spike.md`](hook-surface-spike.md) (measured payloads),
> [`gray-area.md`](gray-area.md) §3 (the client's **30**-event catalogue — that section's prose says
> 31, its own enumeration lists 30, and §VI-a's census registered 30; the prose count is stale), and the measurements in
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
3. **Cost, budgeted per source because they differ by an order of magnitude.**
   - **Transcript read: ≤ 5 ms p95** on a 13 MB transcript (the size measured in §II).
   - **Branch work: ≤ 200 ms p95, and cached per `HEAD`** — recomputed only when `HEAD` moves, never
     on a per-tool tick. A single `git rev-list --count --first-parent` measured on this repository
     costs **median 33 ms, p95 170 ms** (30 runs, 2026-08-23, loaded machine; an independent audit
     measured 6.6 ms / 132 ms on the same range). Either way it is 1–2 orders of magnitude above the
     transcript budget, and an earlier draft of this criterion put both under one 5 ms gate — a
     number the design could not have met and that would have been discovered in Phase 2 as a
     mysterious failure of a component that was working correctly.
   - The gauge MUST be able to answer with branch work **stale or absent** rather than block on it.
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

### Injection: seven events verified, two refuted — and `Stop` re-arms the turn

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

`SubagentStart` injects too, into the **seat's** context rather than the parent's (spike §10) —
seven verified in total. Two refusals bound the set and are why none of these was assumed:
`PostCompact` carries the whole compaction summary and cannot inject a byte of it, and `SubagentStop`
injects nothing at all while still re-arming the seat's turn nine times.

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
    └── sc-stop/            [NEW]    Stop — the nudge channel AND the gauge read, once per turn
                             (no `sc-posttooluse` change: the per-tool tick was cut — §III)
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

### The reference point is a FIELD, not the file's mtime — schema 3

An earlier draft made **mtime the authority** — "a fact the filesystem holds, that no writer can
forget to update" — with `updated:` as a cross-check. That is the wrong authority, and the audit's
criterion-6 finding is what exposed it: **mtime records that the file was TOUCHED, and the design
needs to know when its CONTENT last changed.** A "still accurate" re-affirmation moves mtime exactly
as a rewrite does, so an age measured from mtime resets either way, and the kill switch cannot tell a
fresher note from a touched one. Mitigating that downstream (three-valued `nudge_answered`,
segmented medians, F6's cross-check) was treating the symptom: the fact was never in a field.

**`CHECKPOINT.md` frontmatter goes to `schema: 3` and carries the facts the gauge reads:**

| Field | Type | Meaning |
|---|---|---|
| `written_at` | `<UTC ISO>` | when the note's **body last changed**. This is the age reference point |
| `reaffirmed_at` | `<UTC ISO\|null>` | when it was last confirmed still accurate without changing |
| `body_sha` | `<hex>` | hash of the body below the frontmatter — makes "did the content change" **checkable** rather than claimed |

**All three are flat scalars, which is why they fit.** §VI-b refuses `handles:` because it is a
**list** and `internal/checkpoint.Parse` is deliberately not a YAML parser. That objection is about
shape, not about frontmatter, and it does not reach a timestamp or a hash — the existing parser reads
these with no change to its contract. §VI-b is narrowed accordingly rather than left to look like a
blanket refusal.

**What this retires.** Age is now `now − written_at`, immune to a touch; `reaffirmed_at` records the
other event as its own fact instead of being inferred from a mtime that moved. **mtime is demoted to
a cross-check** — the reverse of the earlier draft — and `body_sha` makes the cross-check decidable:
a note whose `written_at` moved while `body_sha` did not is a mis-written note, and is reportable as
an error rather than a disagreement someone has to adjudicate.

**Carriers of the schema bump**, enumerated because a version surface changing is exactly what
[[complete-the-concept]] sweeps for:

| Carrier | Change |
|---|---|
| `skills/context-checkpointing/SKILL.md` | the schema block: `schema: 3`, three new fields, and the contract that a writer sets `written_at` + `body_sha` on a content change and `reaffirmed_at` otherwise |
| `commands/checkpoint.md` | writes them; this is where "still accurate" becomes an artifact |
| `internal/checkpoint` | no parser change (flat scalars); a `Note` accessor per field, and **schema-2 notes lack all three** |
| `internal/checkpointrestore`, `checkpointseal`, `postcompactobserve`, `filechangedrearm` | read the note; additive fields, no behaviour change |
| `gray-area/tools/internal/claims` | reads a sealed note's body — unaffected, the body is unchanged |
| goldens under `internal/checkpoint*/testdata` | fixtures gain the fields |

**Schema 2 notes must not read as age zero.** Nothing consumes the `schema:` key today (verified: it
appears in comments and fixtures only), so the bump costs no migration — but a note without
`written_at` is **`Unmeasured`, not fresh**, and the gauge reports it as such. That is the same
tri-state discipline `Ceiling` already uses, applied to the field this design turns on.

**One new file, and its fields are specified because a query reads them:**
`.claude/checkpoints/freshness.json`, holding

| Field | Why |
|---|---|
| `bands_spent` | which of NOTICE/WARN/URGENT have fired, **reset when the note is ANSWERED** — i.e. when `written_at` OR `reaffirmed_at` moves. Bands close on an answer; age tracks content. The skill clause makes a reasoned "still accurate" a valid answer, so it must close the band while leaving the age alone |
| `emissions_this_session` | a monotone counter, **NOT reset by an answer** — criterion 4's cap is per session, and a counter that resets cannot enforce it |
| `emission_bytes_max` | the largest render emitted, for criterion 4's ≤ 200 bytes |
| `answered_at_seen` | the latest of `written_at`/`reaffirmed_at` this file has observed — what the band reset is keyed on. A field, so a filesystem timestamp is not load-bearing anywhere in the mechanism |

**The hard cap is here, and it is separate from the band policy.** Bands are "at most once per band
per session, reset on write" — which permits three emissions per note-write cycle, so two checkpoint
writes in one session would exceed criterion 4's ≤ 4 by design. **`emissions_this_session >= 4`
therefore suppresses unconditionally, whatever the bands say**, and the seal row copies both counters
so the budget is *counted* rather than intended. An earlier draft stated the budget in §I, described
the file as "holding only debounce", and left nothing to enforce or read it.

**Keyed on `session_id` alone.** An earlier draft keyed it on `(session_id, agent_id)` with F5
covering "two seats would otherwise silence each other" — **that scenario cannot occur in this
design.** `sc-stop` is the only writer, and `Stop`'s payload carries no `agent_*` fields at all
(spike §13, §12): the seat-bearing channel is `SubagentStop`, which cannot emit anything (§10). A
composite key would have made `agent_id` permanently empty and the promised "R10 regression test"
would have asserted a fiction over a constant. **F5 is retired**, and the reason is recorded here
rather than the row being deleted quietly — it was a real risk for a design that had a per-seat
emitter, and it stops being one at the moment that emitter was cut.

**The number is deferred; the RULE is not, and is fixed here before any data is collected.** Each of
the three measures gets its own distribution from the Phase 1 baseline — `note_age_turns`,
`note_growth_tokens`, `note_branch_commits` — and its own band edges at the **same** percentiles of
**its own** distribution: `NOTICE = P50`, `WARN = P75`, `URGENT = P90`. Naming one field while
claiming three would leave the choice of field open, which is the freedom this paragraph exists to
close.

**Combinator, fixed here too: ANY-OF, taking the MAX band.** A session that burned 400k tokens in
twelve turns and one that took 300 turns without moving the token count are both stale, and §III's own
argument for three measures is that they fail independently — an all-of rule would silence exactly the
lopsided cases the measures were chosen to catch. The cost is a higher emission rate, which is
budgeted (≤ 4 per session) and falsified in Phase 3. Deferring the number without fixing the rule leaves the
bands free to be chosen after seeing the data, which is precisely the post-hoc freedom criterion 6
forecloses for the falsification — and a plan that guards one and not the other has only moved where
the guess lives. If the distribution turns out to make these percentiles absurd (e.g. P50 and P90
within two turns of each other), that is a **finding to report and re-decide with the human**, not a
licence to pick different percentiles quietly.

**Render:** one line, ≤ 200 bytes — the measure, the number, the note's path. **No instruction**,
which is spike §3b's *surviving* claim: "the hook adds no imperative of its own".

> An earlier draft justified this with §3b's other sentence — that a digest asserting content the
> session cannot recognise is "treated as hostile", so the line should read as the session's own
> recovered state. **That sentence is struck through in §3b**, superseded 2026-07-29 by building it
> and measuring: *"The mitigation does not work, and the constraint as written was untestable."*
> Citing a retracted claim to support a live rule is how a withdrawal gets un-withdrawn by a reader
> who never opened the source.

### `[MODIFY] internal/checkpointrestore` — the `--first-parent` defect

`commitsSince` gains `--first-parent` and the surrounding prose changes from "commits ago" to
"commits on this branch". This is a **correctness fix to shipped behaviour**, evidenced in §II
(109 vs 24), and it stands independently of everything else here. The restore digest additionally
carries the gauge — `SessionStart` is the verified channel, and this is where a resumed session
already reads its own provenance.

### `[MODIFY] internal/checkpointseal` — the gauge becomes a record, and the record has to exist first

**There is no seal record today, and this plan previously assumed one.** A seal is currently an HTML
comment prepended to a markdown snapshot — `<!-- sealed: event=%s occasion=%s session=%s agent=%s
at=%s -->` (`internal/checkpointseal/main.go:397-405`) — in a directory pruned to `keepSnapshots = 10`
(`main.go:93`). Two consequences, both fatal to criterion 1 as it was written:

1. **A ≥ 20-boundary baseline cannot be reconstructed from seals**, because the eleventh-oldest is
   deleted.
2. **The stamp is itself the defect this suite is named after** — five facts composed into a string and
   recovered by parsing. Adding `note_age_turns` and friends to it would widen that defect, not use it.

So the first change is a **record**: `.claude/checkpoints/seals.jsonl`, append-only, one JSON object per
seal, written by `internal/checkpointseal` and therefore by all three sealing commands
(`sc-precompact`, `sc-sessionend`, `sc-subagentstop`). Not pruned — a row is ~200 bytes and the
baseline needs history. The markdown snapshot and its comment stamp stay exactly as they are: they
serve human recovery, and [[facts-are-fields]] does not ask for prose to be stripped, only for the
machine-read fact to live in a field.

| Field | Meaning |
|---|---|
| `at`, `event`, `occasion`, `session_id`, `agent_id` | the stamp's five facts, as fields this time |
| `note_age_turns`, `note_growth_tokens`, `note_branch_commits`, `ceiling_known` | criterion 1 |
| `seal_trigger` | `precompact` \| `sessionend` \| `seat_return` — which event sealed |
| `live_handles` | count of `background_tasks` entries with `type != "subagent"`, plus `session_crons` — **only when measurable**. Seats are excluded deliberately: at a `seat_return` seal the returning seat appears in the parent's own list (§12), and counting it reads high by exactly one while answering a different question from "did this note miss some background work" |
| `handles_measured` | `false` when the payload carries no `background_tasks` key |
| `nudge_enabled` | whether the nudge was live when this seal was written — criterion 6's falsification groups on it, and it must be recorded per row rather than inferred from dates |
| `nudge_answered` | `rewritten` \| `reaffirmed` \| `ignored` — **three values, not a boolean, and the distinction is load-bearing** (§III, `SKILL.md`) |
| `emissions_this_session` | copied from `freshness.json`'s monotone counter at seal time, and `emission_bytes_max` beside it — **criterion 4's budget is stated as "counted in the observation row rather than intended", and until this field existed nothing counted it**. The debounce file already holds the number; the seal row is where it becomes checkable after the fact |

**`handles_measured` is not defensive padding; it is the whole of the column's honesty.** Measured
per event 2026-08-23 on 2.1.240 and recorded as `hook-surface-spike.md` **§12** — key sets read from
raw hook stdin:

| Sealing event | carries `background_tasks`? |
|---|---|
| `SubagentStop` | **yes, and populated** — measured with a shell task AND a seat in flight |
| `PreCompact` | **NO — the key is absent from the payload entirely** (4/4 firings) |
| `SessionEnd` | **NO — the key is absent** (payload: `cwd, hook_event_name, prompt_id, reason, session_id, transcript_path`) |

> An earlier draft of this table cited §9c/§11 for the `PreCompact` result. **Those sections do not
> contain it** — §9c is the `Stop` payload, §11 records only firing counts — and it claimed
> `SubagentStop` was measured "populated on a live task" when the only recorded `SubagentStop` payload
> (§9d) shows an empty array from a run with nothing running. The claims turned out true; the
> citations did not exist, which made them assertions wearing a reference. §12 is the measurement,
> taken afterwards.

**Two of the three sealing events cannot measure handles at all**, so `live_handles` is a real column
only on `seat_return` rows — `Stop` carries it and does not seal. #506 is therefore answerable in
Phase 1 **only over seat returns**, which the total-vs-measured query reports explicitly rather than
leaving it to be inferred from a small `n`.

**`live_handles` MUST be omitted, and `handles_measured: false` written, whenever the key is absent.
A zero is never substituted.** The Phase 1 query filters on `handles_measured`, never on
`live_handles > 0`.

**This forces the decoder's shape, and the shape is the mechanism.** `checkpointseal`'s `hookInput`
has no `background_tasks` field today, and a plain `[]Task` decodes **absent** and **`[]`** to the
same nil — collapsing "the event cannot tell me" into "there was no background work". The field MUST
be a pointer or `json.RawMessage`, so presence is observable independently of contents:

| Payload | `handles_measured` | `live_handles` |
|---|---|---|
| key absent (`PreCompact`, `SessionEnd`) | `false` | **omitted** |
| `background_tasks: []` — a seat return with nothing else running | **`true`** | `0` |
| `background_tasks: [...]` | `true` | count |

The middle row is what a `[]Task` decode gets wrong, and it is not a corner case: it is the **normal**
`seat_return`, since §9d's measured payload is exactly `background_tasks: []`. A test covering only
the absent case passes for an implementation that cannot tell the two apart — and the only trigger
able to measure handles would be silently dropped from the baseline.

**A count must also decide about seats.** `background_tasks` includes subagents —
`{"type":"subagent","agent_type":"general-purpose","status":"running"}` (§12) — so at a `seat_return`
seal the returning seat can appear in its own handle list. The field table above carries the single definition; this is why it excludes them.

> **A loose end, recorded rather than smoothed:** `SessionEnd` did **not fire in either** run that
> ended with a background task still running — **0 of 2** — while it fired in 14 of 15 census sessions
> and in the clean control (spike §12). **The confound is named there and is carried here rather than
> dropped:** both live-task runs also hit the driver's pipe close, so "the session did not end the way
> the hook counts as ending" is not excluded. Two observations with an unexcluded confound are a
> pattern, not a finding. Phase 1 MUST therefore count `sessionend` rows **against sessions** rather
> than assume one each, because a seal that silently never fires is indistinguishable from one that
> fired and found nothing.

**Consumer census — of the RECORD's readers, not of the package's importers.** The previous census
ran `grep -rn "checkpointseal" --include=*.go plugins/` and concluded "no external reader": that
enumerates Go files importing the package, a different question, and the conclusion was false. It also
pasted **no command**, so the standard's completion test — re-running surfaces nothing the list omits
— could not be applied by anyone, including its author. Both commands and their full output,
2026-08-23:

```bash
git grep -ln "sealed:\|--seals\|checkpointseal\|snapshotName" -- plugins/ scripts/ docs/
git grep -ln "CHECKPOINT\.md\|checkpoints/"                    -- plugins/ scripts/ docs/
```

Every hit adjudicated, including the ones that do not change — a census that lists only the hits it
acts on cannot be checked against a re-run:

| Hit | Reads what | Changes? |
|---|---|---|
| `cmd/sc-precompact`, `cmd/sc-sessionend`, `cmd/sc-subagentstop` | call the sealer | **YES** — each now also writes a `seals.jsonl` row |
| `internal/checkpointseal/{main.go,main_test.go,drift_test.go,hook_test.go}` | the package | **YES** — new writer plus tests |
| `commands/resume.md:7` | `--seals` tells the agent to list snapshots "with their sealed-at stamp, trigger and agent" — a **prompt-side contract against the comment string** | **No** — the stamp is unchanged. This is the carrier that makes the stamp load-bearing; retiring it starts here |
| `internal/checkpoint/checkpoint.go:239` (`NoteLoopProblems`) + `checkpoint_test.go:101` | the note body. The test fixture contains `<!-- sealed: trigger=auto -->` — **a stamp shape the writer never emits** (it writes `event=`/`occasion=`) | **No** — `Parse` strips scaffolding and never reads the stamp's keys. Flagged because a fixture asserting a format nothing writes is how a format's real readers get miscounted |
| `gray-area/tools/internal/claims/claims.go` + its tests | *"reads a sealed checkpoint note as a set of DECLARED CLAIMS"* — the note body | **No** — snapshot format unchanged |
| `gray-area/commands/audit-checkpoint.md`, `gray-area/README.md` | the note's claims, agent-side | **No** |
| `internal/checkpointrestore/*`, `internal/filechangedrearm/*`, `sessionstart`/`transcript`/`postcompactobserve` tests | the live note or the checkpoints directory | **No** — none reads a seal |
| `commands/checkpoint.md`, `skills/{context-checkpointing,project-memory,validation-loop}/SKILL.md`, `README.md`, `requirements.json` | the note's path and schema, agent-side prose | **No** — `seals.jsonl` is machine-only; no agent must read it |
| `frank-exchange-of-views/.../durability_test.go` | its own record; matched on the word "sealed" | **No** — unrelated |

`keepSnapshots = 10` prunes `*.md` only, so an append-only `.jsonl` beside them is safe;
`.gitignore:19` already covers `.claude/checkpoints/`.

### `[MODIFY] internal/postcompactobserve` — the compaction row keeps its own job

`compaction-observations.jsonl` (`internal/postcompactobserve/main.go:156`) stays what it is: one row
per compaction, written by the PostCompact hook. **It is NOT the baseline record** — §V's Phase 1
queries read `seals.jsonl`, because a row written only at compaction can never carry a `sessionend` or
`seat_return` seal, and grouping by `seal_trigger` there would return `precompact` and nothing else,
which renders exactly like "the other triggers show no difference".

It gains the same age fields for its own purpose — what the note looked like at each compaction — and
`nudge_enabled`, which criterion 6's falsification query reads.

**Consumer census — re-run 2026-08-23, narrowed to the question.** The previous census pasted the
result of `grep -rln "postcompact\|observe" --include=*.go --include=*.md plugins/ scripts/` as eight
files. **That command returns 51.** It matches the word "observe" anywhere, including 30 files in
`frank-exchange-of-views`, `internal/strikecounter`, `scripts/mutate` and `commands/resume.md` — none
of which read this row. The command answering the actual question is:

```bash
grep -rln "compaction-observations" --include=*.go --include=*.md plugins/ scripts/
```

→ `internal/postcompactobserve/main.go` and its test. **Two readers, both in-package, both ours.** The
substantive conclusion the plan relied on survives; the pasted evidence for it did not, and a census
whose command returns 51 while the plan says 8 is the defect §III exists to prevent, committed by §III.

### `[NEW] internal/stopnudge` + `cmd/sc-stop` — the nudge channel

`Stop` is the turn boundary: the natural unit of the operator's "it's been 100 turns", ~10× cheaper
than a per-tool tick, and the only channel that gives the model a turn in which to respond.

**The guard is safety-critical, not politeness.** Measured (§II): an unconditional `Stop` injector
looped nine times and burned 1,186 output tokens on filler. Therefore:

- **Write before emit, and FAIL CLOSED.** The band is recorded as spent **before** the context is
  returned, never after. A guard that writes afterwards, or crashes in between, re-emits — and a
  re-emission on `Stop` is not a duplicate nudge, it is a loop.
  **This inverts the package's house posture, deliberately.** `checkpointseal`'s `seal` and
  `postcompactobserve`'s `appendRow` are best-effort: they report to stderr and continue, which is
  right for a recorder, because a lost row costs one observation. `stopnudge` is not a recorder. An
  unresolvable project directory, an unwritable debounce path, or a marshal error MUST produce **no
  emission at all** — silence is a missed nudge; emitting on an unrecorded band is the nine-firing
  loop measured in spike §8. The loop regression test asserts the failure path explicitly, not just
  the spent-band path.
- **`stop_hook_active`** is checked as a second, independent brake — **measured on `Stop` itself
  (spike §13: `false` on the first firing, `true` on all fifteen re-entries).** An earlier draft
  cited spike §2 for this, which is a `SubagentStop` payload table; the property now has evidence on
  the channel the design uses: if the turn is already a Stop-hook continuation, emit nothing regardless of
  band state. Belt and braces, because the two failure modes have different causes — one is our
  state file, one is the client's own re-entry.
- A **loop regression test** asserts the null case: given a band already spent, the emission is
  empty. This is the test that must never be deleted.

**The `PostToolUse` gauge tick is CUT, and this is a design change, not a wording fix.** It was
specified as a cheap tick that "updates the measurement so `Stop` can decide quickly" — an
optimisation with no measurement behind it, bought at a price the plan had not counted:

- `internal/posttooluse/main_test.go:133` `TestRealUnitsKeepTheirOwnMatchers` iterates **every** unit
  in `Units()` and asserts each one *rejects* `{"tool_name":"Bash"}` and *accepts* `Edit`. A gauge
  unit that accepts `Read`/`Bash` **fails a shipped test** — one that exists precisely to stop a
  merged binary widening what its units act on (`hooks.json:18`'s `_comment` states the same policy).
  Widening would have meant editing an invariant to fit a convenience.
- `requirements.json:68` states the matcher as `"event": "PostToolUse(Write|Edit)"` and would go stale.
- Criterion 3's ≤ 5 ms is a **per-invocation** budget; a per-tool tick multiplies it by the tool count.

`Stop` fires at every turn boundary and can read the gauge itself — **once per turn instead of once
per tool**, which is cheaper than the thing the tick was meant to make cheap. One writer, one event,
one registration. `sc-posttooluse` is untouched by this plan.

**Carrier census for `cmd/sc-stop`** — the one new command:

| Carrier | What changes | Actually gated by |
|---|---|---|
| `hooks.json` | a **new `Stop` registration** | **NOTHING — see below** |
| `requirements.json` `_hook_binaries.binaries[]` | `sc-stop` added | `pluginparity` (`main.go:120-169`, names against `cmd/` dirs) |
| `docs/setup-script.md:99` | "15 at the time of writing" → 16 | `pluginparity` (`main.go:130-138`) |

**An earlier draft of this table said `pluginparity` "fails on any one going stale". That is false,
and the false part is the row that matters.** `grep -c "hooks.json" scripts/pluginparity/main.go`
returns **0**: it gates the marketplace/bootstrap/docs plugin lists, the docs binary count, and
`requirements.json` binary *names* — never registration. The only `hooks.json` check in CI
(`.github/workflows/hooks.yml:428-455`) tests bootstrap-guard degradation, not registration or
matchers. **So a `Stop` hook that is built, declared and documented but never registered passes every
command §V names**, and the binary sits on disk doing nothing while the gates read green — the
plausible-zero shape, at the level of the gate set.

Phase 2 therefore carries a **stated manual review** of `hooks.json`, and this plan files the gap
rather than pretending a gate covers it: extending `pluginparity` with `hooks.json` ↔ `cmd/`
registration parity is the real fix and belongs to the parity tool, not here.

The parent plan's Phase 5 preferred `PostToolUseFailure`. That is now demoted to a fallback and
was, on the evidence, the wrong instinct for a good reason: it was the only injector known at the
time. Staleness has nothing to do with a tool failing.

### `[MODIFY] skills/context-checkpointing/SKILL.md`

> - AFTER a freshness nudge, YOU MUST either write the note or state why the current note is still
>   accurate. The nudge measures the note's **age**, never its **truth**: a note can be 300 turns
>   old and exactly right because nothing changed, and 3 turns old and already wrong because the
>   last turn changed the objective. Silence is the failure this mechanism exists to remove; a
>   reasoned "still accurate" closes the band and is a valid answer.

**This clause is `unobservable-duty`, and naming its class is what exposed the hole.** The registry's
sweep question for that class is *"what artifact would a reviewer look at to tell whether this was
done?"* — and for a spoken "still accurate" the answer is **nothing**. The duty would be discharged
into the conversation and leave no trace, so criterion 6's falsification could not distinguish a
session that answered the nudge from one that ignored it. **The seal row therefore carries `nudge_answered`** — and it MUST be three-valued.

> **A boolean here would have broken criterion 6 — and chasing that is what exposed the real defect.**
> The age reference point *was* `CHECKPOINT.md`'s mtime (it is now `written_at`, §III). A "still accurate" re-affirmation moves that mtime
> **exactly as a real rewrite does**, so both reset note-age to ~0 and a boolean marks both `true`.
> Criterion 6 — *median note-age-at-seal falls* — could then not distinguish **"the nudge made notes
> fresher"** from **"the nudge trained agents to touch the note"**, and would report success for the
> second. F6's `updated:`-vs-mtime cross-check does not catch it: the two agree, and only the
> *content* is unchanged. The kill switch would have been disarmed by the mechanism it polices.
>
> **Schema 3 removes the cause**, so this three-valued field is no longer load-bearing for criterion
> 6 — it survives as the diagnostic that says *how* the duty was discharged (§VI-c), and because an
> `ignored` rate is F1's wallpaper signal whatever the medians do.

The value goes in the **seal row**, a record with fields — not into `updated:`, which is `<UTC ISO>`
under schema 2 and parsed by a deliberately flat-scalar reader, so a judgement composed into it would
be the exact shape §III indicts elsewhere. An earlier draft said the judgement was "recorded in
`updated:`"; that field cannot hold it.

**Which population criterion 6 is computed over is a DECISION, not an editing fix**, and is left open
in §VI-c rather than chosen quietly.

**Consumer census** — `git grep -ln "context-checkpointing" -- plugins/ scripts/ docs/`, full output,
every hit adjudicated:

| Hit | Relationship | Changes? |
|---|---|---|
| `skills/context-checkpointing/SKILL.md` | the surface itself | **YES** — the clause above |
| `commands/checkpoint.md`, `commands/resume.md` | invoke the skill; `checkpoint.md` enumerates the note's sections | **YES for `checkpoint.md`** — it must write `nudge_answered`'s trigger; `resume.md` unchanged |
| `skills/validation-loop/SKILL.md`, `skills/project-memory/SKILL.md`, `skills/complete-the-concept/SKILL.md` | sibling rules that cite `[[context-checkpointing]]` | **No** — swept (below), none carries a rival nudge duty |
| `README.md` | describes the skill | **YES** — one line, if the nudge ships |
| `internal/checkpoint/checkpoint.go`, `checkpointrestore/{main,main_test}.go`, `checkpointseal/main.go`, `filechangedrearm/main.go` | reference the skill in comments; parse the note | **No** — none reads the clause; the note's schema is unchanged by it |

**Rule-Class and sibling sweep, for the commit that lands this** — `rulesweep` is a mandatory
pull-request gate (`.github/workflows/hooks.yml:204`) and its protocol-surface list includes both
`/skills/[^/]+/SKILL\.md$` and `/hooks/hooks\.json$`, which this plan modifies. The implementation
commit MUST carry:

```
Rule-Class: unobservable-duty
Sibling-Sweep: checked validation-loop, project-memory, complete-the-concept and commands/{checkpoint,resume}.md —
  none carries a rival post-nudge duty; checkpoint.md gains the artifact that makes this one observable
```

The gate rejects prose *describing* the trailers rather than carrying them (`sweep.go:97`), so this
block is a specification of the commit, not a substitute for it.

## IV. Risk & Mitigation

**Convention, stated because F10's re-rating exposed that it never was.** `L×I×C` is
**pre-mitigation** — F3 is `H×M×L` for a defect present in shipped code and mitigated in its own row.
A row rated post-mitigation is not comparable with the rest of the table, so where mitigation changes
the picture enough to be worth recording, it appears as a separate **residual** term rather than by
editing L or I.

| # | Risk | L×I×C | Mitigation | Step |
|---|---|---|---|---|
| F1 | **The nudge becomes wallpaper** — and a context-pressure warning that costs context is self-defeating at exactly the moment it matters. | H×H×M | Bands, once per band; ≤ 200 bytes; reset on write; **criterion 6 removes it** if the baseline median does not fall. | Ph. 3 |
| F2 | **Fabricated denominator.** | M×H×L | `Ceiling` tri-state with `Unknown` as a value callers must handle; criterion 2 is a test over the render. | Ph. 1 |
| F3 | **Commit count reads other people's work as mine.** | H×M×L | `--first-parent`; no author filter (§II measured both). | Ph. 1 |
| F4 | **Gauge cost on a hot path.** | M×M×L | Bounded tail; skip entirely when note mtime is unchanged and the band is current — mtime survives here as a **cheap short-circuit**, which is a performance question, not the measurement; criterion 3 is a measured gate. | Ph. 2 |
| ~~F5~~ | ~~**Concurrent seats share `session_id`.**~~ **RETIRED** — no per-seat emitter exists: `sc-stop` is the only writer and `Stop` carries no `agent_*` fields (spike §12/§13); the only seat-bearing channel cannot emit (§10). Left in place rather than deleted, because it was a live risk until the `PostToolUse` tick was cut. | — | — | — |
| F6 | ~~**mtime lies** — a tool touches the note without rewriting it.~~ **LARGELY RETIRED**: age is read from `written_at`, not mtime (§III, schema 3), so a touch no longer resets it. What remains is a **writer** that sets `written_at` without changing the body — caught by `body_sha`, and reported as an error rather than adjudicated. | L×M×L | `body_sha` disagreement is an error; mtime demoted to cross-check. | Ph. 1 |
| F7 | **Short sessions get lectured.** | M×L×L | Gauge arms only above a floor (a note exists, or the session crossed a turn/token floor). Below it, silent. | Ph. 3 |
| F8 | **Age is read as truth.** A fresh note is not a correct note. | M×M×L | The skill clause says so; the render states a measure, never a verdict. gray-area's `/audit-checkpoint` remains the instrument for the note's *claims*, and the two are deliberately different tools. | Ph. 3 |
| **F10** | **`Stop` injection loops** — measured, not theorised, and **worse on the current client**: 9 firings / 1,186 output tokens on 2.1.235, **16 firings / 35 assistant entries / 4,326 output tokens on 2.1.240** (spike §13). The cap is undocumented and moved between two patch versions. | **H×H×L** · residual after mitigation **L×M** | Write-before-emit; `stop_hook_active` as an independent second brake; a loop regression test asserting the empty emission on a spent band. §III. | Ph. 2 |
| F9 | **Hook-surface churn** (R9, realised **four times**: `SubagentStop` changed behaviour between 2.1.220 and 2.1.240; `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` stopped working entirely; and the `Stop` loop's cost went 9 firings/1,186 tokens → **16/4,326** between 2.1.235 and 2.1.240). | M×M×M | Channels are measured per client and each row says which: §8 on **2.1.235**; **§9–§13 on 2.1.240**. `Stop` injection **has** been re-run on 2.1.240 — 16 attachments (§13); an earlier draft of this cell still said it had not, after §13 measured it. Phase 2 re-confirms at build time as **version-drift insurance**, not because the channel is unmeasured: this row's own history is four surprises across three client versions. | all |

---

## V. Verification Plan

### Phase 1 — gauge, defect fix, baseline (no nudge)

```bash
(cd plugins/prosthetic-conscience/tools && go test ./internal/ctxusage/... ./internal/freshness/... \
    ./internal/checkpointrestore/... ./internal/checkpointseal/... && go test ./...)
```

Coverage must include the negatives, which are the point: a tail containing **no** assistant entry;
a truncated final line; a session with **no** compact boundary → `Ceiling: Unknown` **and no
percentage in the render**; two boundaries → most recent wins; **a schema-2 note lacking `written_at` → `Unmeasured`, NOT age zero**; `written_at`
moved while `body_sha` did not → reported as an **error**, not adjudicated; a note head that is
unreachable; a branch whose history contains merges → first-parent and plain
counts differ; **a payload with no `background_tasks` key → `handles_measured: false` and NO
`live_handles` field** (the `PreCompact`/`SessionEnd` case, measured §12); **and the case that
discriminates — `background_tasks: []` PRESENT but empty → `handles_measured: true`, `live_handles:
0`** (the normal `seat_return`, §9d). Without the second, the suite passes for a `[]Task` decode that
cannot tell absent from empty, and the only trigger that can measure handles is silently dropped from
the baseline; **a seat's own entry excluded from the count** (§12: a live seat appears as
`type: "subagent"` in the parent's list); **a seal row written on each of the three
trigger events**, asserted per event rather than per package.

**Driveable check on real data.** Run the gauge against a live multi-MB transcript and a real
`CHECKPOINT.md`, and check the token figure by hand against `tail -c 200000 … | jq '.message.usage'`.
Then, in this repository:

```bash
# Pinned to a fixed range: these numbers move with every commit, and an expectation
# that re-arms on each commit fails for reasons unrelated to the defect it guards.
# (An earlier draft hardcoded "expect 24, not 109"; by the next commit it was 25 and 110.)
R=24f8fc63622f39797e5b4103c003f5aa1465138b..1105a02
test "$(git rev-list --count --first-parent $R)" -lt "$(git rev-list --count $R)" || echo FAIL
```

The **relation** is the invariant — first-parent strictly less than total across a merge-bearing
range — not either number.

**Then collect the baseline** — ≥ 20 real boundaries, seals stamped, nothing emitted. The record is
`.claude/checkpoints/seals.jsonl`, written by `internal/checkpointseal` on all three sealing events:

```bash
jq -s 'map(select(.note_age_turns)) | {n:length,
        turns:(map(.note_age_turns)|sort), growth:(map(.note_growth_tokens)|sort),
        commits:(map(.note_branch_commits)|sort)}' .claude/checkpoints/seals.jsonl
```

And the two folded-in observations. **Both queries filter on measurability first**, because the
absent case and the honest zero are otherwise the same bytes:

```bash
# #506: is a note sealed with live background work staler than one sealed without?
# Rows where the payload could not tell us are EXCLUDED, not counted as zero.
jq -s '[.[] | select(.handles_measured)] | group_by(.live_handles > 0)
       | map({live_handles: .[0].live_handles > 0, n: length,
              median_age: (map(.note_age_turns) | sort | .[length/2|floor])})' \
   .claude/checkpoints/seals.jsonl

# ...and how much of the baseline that excluded, reported rather than inferred from a small n:
jq -s '{total: length, measured: ([.[] | select(.handles_measured)] | length)}' \
   .claude/checkpoints/seals.jsonl

# #507: how stale is the note at a seat return, against the other triggers?
jq -s 'group_by(.seal_trigger)
       | map({trigger: .[0].seal_trigger, n: length,
              median_age: (map(.note_age_turns) | sort | .[length/2|floor])})' \
   .claude/checkpoints/seals.jsonl
```

**A result of "no difference" is only readable if all three triggers are present.** Assert that
first — `map(.seal_trigger) | unique | length == 3` — because a query returning one group looks
identical to a query whose other groups genuinely did not differ. Expect `sessionend` to be the thin
one: it did not fire at all in the one measured run that ended with a background task still running.

**Both may come back "no difference", and that is a result.** If `live_handles > 0` does not track
with staler notes, #506's premise is wrong and its strong form should not be built at any price; if
`seat_return` is not staler than `sessionend`, #507 has no case. Stated before the data, so the
outcome cannot be reinterpreted after it — the same contract criterion 6 makes for the nudge itself.

The distribution sets the Phase 2 thresholds **at the percentiles fixed in §III**, not at percentiles
chosen once the shape is visible. This phase is shippable alone: the `--first-parent` fix is a
correctness change and the stamping is pure observation.

### Phase 2 — the nudge on `Stop`, guarded

Thresholds from Phase 1, at §III's preregistered percentiles. **Two cost gates, not one** (criterion
3): the transcript read p95 over 100 invocations **fails above 5 ms**; branch work p95 **fails above
200 ms** and must be served from the per-`HEAD` cache on a repeat call.

**The protocol-rule gate, which the implementation commit fails without:**

```bash
(cd scripts && go run ./rulesweep -base origin/main)
```

`rulesweep` is a mandatory pull-request gate and its protocol surfaces include
`/skills/*/SKILL.md` and `/hooks/hooks.json` — this plan touches both. The commit carries
`Rule-Class: unobservable-duty` and a `Sibling-Sweep:` trailer naming the surfaces checked (§III).
Prose describing the trailers does not satisfy it.

**Criterion 4's budget, counted rather than intended:**

```bash
# No session may exceed 4 emissions, and no render may exceed 200 bytes.
jq -s 'map(select(.nudge_enabled)) | {sessions: (group_by(.session_id) | length),
        worst_count: (map(.emissions_this_session) | max),
        worst_bytes: (map(.emission_bytes_max) | max)}' .claude/checkpoints/seals.jsonl
```

`worst_count > 4` or `worst_bytes > 200` fails Phase 2. A criterion whose number appears only in §I is
an intention; this is the query that makes it a check.

**Did the duty get discharged, and which way** — the artifact that keeps the `SKILL.md` clause out
of the `unobservable-duty` class. A field no check reads leaves the class alive one level up:

```bash
jq -s '[.[] | select(.nudge_answered)] | group_by(.nudge_answered)
       | map({outcome: .[0].nudge_answered, n: length})' .claude/checkpoints/seals.jsonl
```

A high `ignored` rate means the nudge is wallpaper (F1) whatever the age medians say. A high
`reaffirmed` rate with falling age medians is the failure criterion 6 must not report as success —
which is why Phase 3 segments on this field rather than aggregating over it.

**Registration parity — and what it does NOT cover:**

```bash
(cd scripts && go run ./pluginparity)
```

`sc-stop` is a new `cmd/` directory, so `requirements.json` `_hook_binaries.binaries[]` and
`docs/setup-script.md`'s binary count both move; `pluginparity` fails on either going stale, and §V
previously ran only `go test` inside `tools/`, which cannot see them.

**It does not cover `hooks.json`.** `pluginparity` never reads that file (§III), and no other gate
checks registration, so Phase 2 also requires a **stated manual review**: confirm `sc-stop` is
registered on `Stop`, and that no other event's matcher moved. Written down as a manual step because
naming a gate that does not check the thing is worse than admitting there is no gate — the run goes
green either way, and only one of the two tells you why.

**The loop gate comes first, because it is the one that burns tokens** — and it asserts a
**relation, not a constant**. An earlier draft required "≤ 4 assistant entries", which is §8's clean
control **on 2.1.235**; §13 then measured this channel's counts moving 3.6× between two patch
versions. A constant copied from one client's run is the same mistake §V already refuses for the git
range and spike §11 earned for compaction boundaries.

Run **both arms in one harness**, same project, same prompt, same model:

| Arm | Assertion |
|---|---|
| nudge disabled (control) | record `stop_firings` and `assistant_entries`; these are the baseline, **measured, not quoted** |
| band already spent | **exactly equal** to the control on both counts |
| band unspent | `stop_firings` = control + 1, `assistant_entries` ≤ control + 1 — the single turn the emission is *intended* to create, and no more |

That is criterion 5 stated as criterion 5 states it. Any absolute number in this gate is a number
that will be wrong on the next client.

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
jq -s '{before:      [.[]|select(.nudge_enabled==false)|.note_age_turns]|sort,
        after_all:   [.[]|select(.nudge_enabled==true )|.note_age_turns]|sort,
        after_rewritten: [.[]|select(.nudge_enabled==true and .nudge_answered=="rewritten")|.note_age_turns]|sort,
        after_reaffirmed:[.[]|select(.nudge_enabled==true and .nudge_answered=="reaffirmed")|.note_age_turns]|sort}' \
   .claude/checkpoints/seals.jsonl
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

### VI-c. Criterion 6's ambiguity — resolved by removing its cause, not by choosing a population

This section previously offered the human three ways to compute criterion 6's median, because
`rewritten` and `reaffirmed` both moved `CHECKPOINT.md`'s mtime and the age therefore reset either
way. All three were workarounds for a measurement taken from the wrong place.

**Resolved instead by putting the fact in a field** (§III, schema 3): age is `now − written_at`,
which moves only when the **body** changes. A re-affirmation sets `reaffirmed_at` and leaves the age
alone, so criterion 6's median is computed over **every seal**, with no population choice to make and
nothing for a "touched" note to score.

| Was | Now |
|---|---|
| A — rewrites only | unnecessary: re-affirmation no longer inflates the median |
| B — all answered | unnecessary: a touched note is not a fresher note by construction |
| C — segmented, report both | **kept, but as a diagnostic rather than the verdict** — `nudge_answered`'s three values still say *how* the duty was discharged, and a high `ignored` rate is F1's wallpaper signal whatever the medians do |

**The lesson is worth more than the fix.** Three rounds of audit hardened a measurement built on
mtime — a three-valued field, a segmented median, a cross-check, a risk row — and none of them
questioned the reference point. The plan's own §I says a fact another party acts on belongs in a
field on a record; the age reference was a fact recovered from a filesystem timestamp, and the
document arguing that principle did not apply it to its own primary measure.

### VI-b. Blocked: the note has no field for its handles

> **Narrowed by §III's schema-3 change.** This section is about `handles:`, a **list**, and its
> objection is about SHAPE: `internal/checkpoint.Parse` is deliberately not a YAML parser. It is not
> an argument against frontmatter, and it does not reach the flat scalars `written_at`,
> `reaffirmed_at` and `body_sha`, which the existing parser reads unchanged. What follows stands for
> lists; it never stood for scalars, and an earlier draft let it read as a blanket refusal.

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

**What the census changes about the plan's own risks.** **F10 keeps its `H×H×L` rating and gains a
stated residual of `L×M`** (§IV, where the table's pre-mitigation convention is now written down; an
earlier draft of this paragraph announced a re-rating to `L×M×L`, which would have made F10 the one
row not comparable with the others). The measurement is undisturbed — 9 firings and 1,186 wasted
tokens stand. What is low is the **residual**, because the hazard is cycle detection and two
independent brakes are already specified: `stop_hook_active` is a
field the client sets on every re-entry (false on the first firing, true on the eight that followed,
on `Stop` and `SubagentStop` alike), needing no state and no write, and the debounce file carries
band policy, whose worst failure is one duplicate emission. The original rating priced an *unguarded*
injector, which nothing here proposes. The mitigations and the regression test are unchanged.

**Why the rating stays H×H×L anyway:** the table is pre-mitigation throughout (F3 is `H×M×L` for a
defect present in shipped code and fixed in its own row), so lowering F10's terms would have priced
its mitigation twice — once in the rating and once in the mitigation column — and made it the only
row not comparable with the rest. The reduction is real and belongs in the residual, which is where
it now is. **F9 gains a measured instance rather than a theoretical one:** `SubagentStop`'s behaviour
on 2.1.240 contradicts this record's own 2.1.220 reading (§9e correction 2), so "each binary is
inert rather than broken when its event never fires" now has a case where the event fires
*differently* — inertness does not cover that, and no mitigation here does either.
