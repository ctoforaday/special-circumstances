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

**Render:** one line, ≤ 200 bytes — the measure, the number, the note's path. No instruction. Spike
§3b twice observed an injected directive treated as a suspected prompt injection; the line reads
as the session's own recovered state.

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
| `live_handles` | `len(background_tasks) + len(session_crons)`, **only when measurable** |
| `handles_measured` | `false` when the payload carries no `background_tasks` key |
| `nudge_enabled` | whether the nudge was live when this seal was written — criterion 6's falsification groups on it, and it must be recorded per row rather than inferred from dates |

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
seal the returning seat can appear in its own handle list. `live_handles` counts entries with
`type != "subagent"` plus `session_crons`; a seat-inclusive figure answers a different question from
"did this note miss some background work", and reads high by exactly one.

> **A loose end, recorded rather than smoothed:** in a run ending with a background task still
> running, **`SessionEnd` did not fire at all** (0/1; it fired in 14 of 15 census sessions, all of
> which ended clean). One observation is not a finding, but it points at the same case this plan
> cares about, so Phase 1 MUST count `sessionend` rows against sessions rather than assume one each.

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
| `internal/checkpointseal/{main,drift,hook,main_test}.go` | the package | **YES** — new writer plus tests |
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
- **`stop_hook_active`** is present in the payload (spike §2) and is checked as a second,
  independent brake: if the turn is already a Stop-hook continuation, emit nothing regardless of
  band state. Belt and braces, because the two failure modes have different causes — one is our
  state file, one is the client's own re-entry.
- A **loop regression test** asserts the null case: given a band already spent, the emission is
  empty. This is the test that must never be deleted.

`sc-posttooluse` additionally fans in the gauge as a cheap tick — it updates the measurement so
`Stop` can decide quickly, and it does **not** emit. One writer of the nudge, one event.

**Carrier census for the two commands — three of these are hard-gated and were previously omitted.**
`scripts/pluginparity` fails the build on any of them going stale, and §V Phase 2 must run it:

| Carrier | What changes | Gate |
|---|---|---|
| `plugins/prosthetic-conscience/hooks/hooks.json` | a **new `Stop` registration** for `sc-stop`; `PostToolUse`'s matcher goes from `Write\|Edit` to **`Write\|Edit\|Read\|Bash`** | `pluginparity` |
| `requirements.json` `_hook_binaries.binaries[]` | `sc-stop` added | `pluginparity` |
| `docs/setup-script.md:99` | reads "15 at the time of writing"; a new `cmd/` directory makes it 16 | `pluginparity` `main.go:130-138` |
| `hooks.json` `_comment` on `PostToolUse` | states the union policy this widens — the prose is the policy's only statement | review |

**The matcher value is stated because a matcher is a contract, not an intention.** `Write|Edit|Read|Bash`
is chosen as the smallest set that tracks context growth: `Read` and `Bash` are what actually move the
token count, `Write|Edit` are already registered for the quality gate. It is deliberately **not** `*` —
the `_comment` at `hooks.json:18` records that each unit "still checks its own applicability so merging
cannot widen what it acts on", and criterion 3's ≤ 5 ms p95 is a **per-invocation** budget, so the
matcher is also the sizing input for it. A tick on every tool call would multiply that budget by the
tool count and could not be reviewed against a set nobody wrote down.

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
| F4 | **Gauge cost on a hot path.** | M×M×L | Bounded tail; skip entirely when note mtime is unchanged and the band is current; criterion 3 is a measured gate. | Ph. 2 |
| F5 | **Concurrent seats share `session_id`.** | M×M×L | Debounce keyed on `(session_id, agent_id)`; R10 regression test. | Ph. 2 |
| F6 | **mtime lies** — a tool touches the note without rewriting it. | L×H×M | `updated:` cross-check; disagreement is reported rather than resolved silently. | Ph. 1 |
| F7 | **Short sessions get lectured.** | M×L×L | Gauge arms only above a floor (a note exists, or the session crossed a turn/token floor). Below it, silent. | Ph. 3 |
| F8 | **Age is read as truth.** A fresh note is not a correct note. | M×M×L | The skill clause says so; the render states a measure, never a verdict. gray-area's `/audit-checkpoint` remains the instrument for the note's *claims*, and the two are deliberately different tools. | Ph. 3 |
| **F10** | **`Stop` injection loops** — measured, not theorised: 9 firings, 8 filler turns, 1,186 wasted tokens from one unguarded emission. | **H×H×L** · residual after mitigation **L×M** | Write-before-emit; `stop_hook_active` as an independent second brake; a loop regression test asserting the empty emission on a spent band. §III. | Ph. 2 |
| F9 | **Hook-surface churn** (R9, realised **three times**: `SubagentStop` changed behaviour between 2.1.220 and 2.1.240, and `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` stopped working entirely). | M×M×M | Channels are measured per client and each row says which: §8 on **2.1.235**, §9–§11 on **2.1.240**. `Stop` injection has NOT been re-run on 2.1.240 — Phase 2 re-confirms the attachment at build time rather than assuming it. Inertness does not cover an event that fires *differently*, and no mitigation here does. | all |

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
counts differ; **a payload with no `background_tasks` key → `handles_measured: false` and NO
`live_handles` field** (the `PreCompact` case, measured); **a seal row written on each of the three
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

**Registration parity, which Phase 2 cannot pass without:**

```bash
(cd scripts && go run ./pluginparity)
```

`sc-stop` is a new `cmd/` directory, so `hooks.json`, `requirements.json` `_hook_binaries.binaries[]`
and `docs/setup-script.md`'s binary count all move together; `pluginparity` fails on any one going
stale. §V previously ran only `go test` inside `tools/`, which cannot see any of them.

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

**What the census changes about the plan's own risks.** **F10 keeps its `H×H×L` rating and gains a
stated residual of `L×M`** (§IV, where the table's pre-mitigation convention is now written down; an
earlier draft of this paragraph announced a re-rating to `L×M×L`, which would have made F10 the one
row not comparable with the others). The residual is low — 9 firings and 1,186 wasted tokens stand — but because the
hazard is cycle detection, and two independent brakes are already specified: `stop_hook_active` is a
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
