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
2. **No percentages at all.** Not "none without a denominator" — none. A fraction needs a window,
   nothing in any payload or transcript carries one, and the only obtainable denominator exists on
   some sessions and not others. Every figure this design reports is absolute. Asserted as a test
   over the render, not as an intention.

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
   control is the **in-harness arm measured at the time**, never a quoted constant: spike §8 read 4
   entries clean / 20 looping on 2.1.235 and §13 read 35 on 2.1.240, so any number written here is
   already wrong for some client.
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

### The numerator is exact, and there is no denominator — so nothing is divided

Context size is exact: `usage.input_tokens + cache_read_input_tokens + cache_creation_input_tokens`
on the last `type:"assistant"` entry, measured against `jq` on a live transcript.

**The window is not available and the design stops looking for one.** No payload and no transcript
field carries it; `message.model` reads `claude-opus-5` on a session that is `claude-opus-5[1m]`;
`SessionStart`'s payload may carry a model name on some clients, which would still be a name and not
a size. A session's own `compactMetadata.preTokens` is the one real number in reach, and it exists
only after that session has compacted — so a percentage would be available on some sessions and
absent on others, which is the worst of both: a figure that looks comparable and is not.

**So: no percentages.** The measures are absolute — tokens grown, turns taken, commits landed — and a
reader who wants a ratio supplies the window knowingly. That is a smaller claim than this section
used to make, and it costs nothing the design was actually using: `preTokens` is no longer read at
all, and `cumulativeDroppedTokens` from the same boundary is kept, because growth needs it to stay
monotone across a compaction (**1,001,875 → 12,823** measured at one boundary — the raw counter
resets, so the naive difference goes negative and the stalest note reads as the freshest).

### Work done is a branch-line count, not a HEAD distance — measured

The restore **used to** compute staleness as `git rev-list --count <note-head>..HEAD`
(`internal/checkpointrestore/main.go`, fixed in `2eeebc7`). That counted every commit reachable from
HEAD and not from the note — which is dominated by *other people's work arriving*, not by work this session
did. On this repository, from the merge-base with `main` to `HEAD`:

```
git rev-list --count               <merge-base>..HEAD   →  109
git rev-list --count --first-parent <merge-base>..HEAD   →   24
```

**85 of those 109 are merged-in side branches.** A note written before a routine `main` merge
was reported as 100+ commits stale having done nothing. `--first-parent` walks this branch's own line
and answers the question that matters: how much work landed *here* since the note. **Fixed and
tested** — `commitsSince` had no test against a real repository at all, since `staleness` was tested
through an injected counter and the shelling-out half was never exercised.

**Do not filter by author.** Re-measured 2026-08-22 across the same range in two worktrees, which
agree: `--author=$(git config user.email)` (`gblock+agent@ctoforaday.com`) returns **1** of 109 —
the single commit `13332bc` — against `noreply@anthropic.com` (85) and `gblock@ctoforaday.com`
(23). The configured identity is not the identity the branch's work was committed under, and the
filter reports 1 for a branch that this pair did all 109 commits of.


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
│   ├── ctxusage/           [NEW]    transcript tail → {tokens, turns, dropped} · all tri-state
│   ├── freshness/          [NEW]    the gauge: three measures, bands, debounce, one-line render
│   ├── checkpointrestore/  [MODIFY] --first-parent defect; gauge in the restore digest
│   ├── checkpointseal/     [MODIFY] stamp note-age onto every seal record
│   ├── stopnudge/          [NEW]    the guarded single emission at the turn boundary
│   └── postcompactobserve/ [MODIFY] carry note-age-at-seal into the observation row
└── cmd/
    └── sc-stop/            [NEW]    Stop — the nudge channel AND the gauge read, once per turn
                             (no `sc-posttooluse` change: the per-tool tick was cut — §III)
```

### `[NEW] internal/ctxusage` — the measurement · **BUILT**

Bounded backward scan of the last **256 KB** of `transcript_path` — the number is fixed here because
"N KB" is not a specification — for the most recent `type:"assistant"` entry and any
`subtype:"compact_boundary"`. Widen **once**, to 1 MB, if no assistant entry is found — **and only for that miss**, never for an
unmeasured turn count: for a genuinely old note a wider window does not reach back either, so the
second read cannot change the answer. Measured, when it did widen for turns: **p95 160 ms against a
5 ms budget** with a 2 ms floor; probing first still put p50 at 5.9 ms, because it is a second file
open on every call. Past one widen, report `Unmeasured` — a hook must not read 13 MB on a tick, and must not hang on a rotated or truncated
file. Returns `Tokens` (exact), `Turns`, and `Dropped`,
which is `compactMetadata.cumulativeDroppedTokens` from the most recent boundary, or unmeasured. **`preTokens` is deliberately not read**: it was the only available denominator and this design renders no percentages.

**Built, with two things the spec did not anticipate, both found by the cost gate.**

**Reachability is decided BEFORE the parse.** If the window does not reach back past `written_at` the
turn count will be discarded, so counting it is work whose result is thrown away — and on a real
transcript that is most of the parsing. In that case only the newest usage figure is needed, found by
scanning backward and stopping at the first hit. Measured on a live transcript: **p50 5.9 ms → 1.4 ms
against a raw-read floor of 1.3 ms**, i.e. measuring now costs about 0.1 ms over the read it cannot
avoid.

**And a prefilter is required to meet criterion 3.**
Unmarshalling every line in the window cost **p50 19 ms / p95 203 ms** on a live 3.1 MB transcript —
4× to 40× over budget. Parsing only lines that can carry a figure brings it to **p50 3.0 ms / p95
3.9 ms** against a raw-read floor of 1.3 ms. The prefilter is a performance device only: every line
it admits is parsed properly and every field is read from the parsed struct, never from the bytes.
**A first attempt matched bare `assistant` and admitted 47 of 113 lines**, because tool results carry
`sourceToolAssistantUUID` — a prefilter that admits everything is not a prefilter, and it was
silently not one. Measured admit rates over one real window: `assistant` 47, `"type":"assistant"` 43,
`"usage"` 43, `compact_boundary` 7.

**`Turns` is `Unmeasured` whenever the window does not reach `written_at`, and NEVER a partial
count.** This is the failure mode the whole design exists to catch, inverted: a 300-turn-old note is
precisely the one whose earliest turns fall outside a bounded scan, so a truncated count would report
the stalest notes as the freshest — and a small number is indistinguishable from an honestly small
one. The scan therefore checks whether it reached back past `written_at` before it reports at all:
if the oldest entry in the window is newer than `written_at`, the answer is "I could not see far
enough", not a number. 

Every tri-state here forces its caller to handle the miss: the flag is a separate field, so a zero cannot be mistaken for an answer.
substituted, never a zero. This is clause 3 made structural rather than remembered.

### `[NEW] internal/freshness` — the gauge · **BUILT**

Three measures, independent because they fail independently: a session can burn 400k tokens in
twelve turns of bulk reading, or take 300 turns without moving the token count.

| Measure | Definition |
|---|---|
| **Growth** | `(tokens_now + dropped_now) − (tokens_at_write + dropped_at_write)` |
| **Turns** | assistant turns since the note was written |
| **Branch work** | `git rev-list --count --first-parent <note.head>..HEAD` |

**Built.** `Gauge` is arithmetic over readings the caller already holds — it re-reads nothing, so it
stays callable on a tick — and `Observe` stamps the write-time reading **once per note**, keyed on
`written_at`. Once per note and not once per tick is the whole of it: re-stamping would move the
reference point to the present on every call, and growth would report the interval since the last
tick rather than since the note. That measure is always near zero and always looks healthy.

`Branch` is passed IN rather than computed here, because it costs a git subprocess (p50 33 ms, p95
170 ms measured on this repository) — one to two orders of magnitude above the transcript budget.

**Who owns the cache, and when there is one to own.** In Phase 1 there is no cache and none is
needed: the callers are the two record writers, which run at a **seal** and at a **compaction** —
events that happen a handful of times per session, where a 33 ms subprocess is invisible. The cache
becomes necessary only when something calls the gauge on a **turn boundary**, and that something is
`cmd/sc-stop` in Phase 2. **`sc-stop` owns it**, in `freshness.json`, as `branch_head_seen` +
`branch_commits_seen`: recomputed only when `HEAD` moves. Criterion 3's 200 ms branch budget and §V's
repeat-call gate are Phase 2 gates for that reason — an earlier draft placed the requirement with no
owner named, which is a gate over a component the spec never assigned.

**No thresholds and no bands are built.** Phase 1 collects the distribution they are supposed to come
from; a package that shipped a default would be laundering the guess §III refuses.

**Growth needs a write-time reading, and no record held one.** "Tokens now − tokens when the note
was written" was never computable: `ctxusage` returns the current figure, the note carries no token
count (an agent cannot read its own context size any more than it can hash its own bytes), and
`freshness.json` held only band state. **`freshness.json` therefore gains `tokens_at_write` and
`dropped_at_write`**, stamped by the gauge the first time it observes a `written_at` it has not seen
before. The note stays prose the agent writes; the numbers stay facts the machine reads.

**Growth must survive a compaction, and the naive subtraction goes NEGATIVE across one.** §II's own
measurement shows the counter resetting 1,001,875 → 12,823 at a boundary, so a note spanning one
would report growth of −989,052 and read as the freshest note in the file. `compactMetadata`
carries `cumulativeDroppedTokens`, which makes the true figure recoverable:

```
growth = (tokens_now + cumulative_dropped_now) − (tokens_at_write + cumulative_dropped_at_write)
```

Both terms are monotone, so growth is monotone. If either dropped-figure is unavailable — a session
that has never compacted has no boundary to read — both are 0 and the identity still holds. **If
`tokens_at_write` is absent** (a note first seen before this shipped), growth is `Unmeasured` and
**never zero**, which would read as "no work since the note".

### The reference point is a FIELD, not the file's mtime — schema 3

**mtime records that the file was TOUCHED, and the design
needs to know when its CONTENT last changed.** A "still accurate" re-affirmation moves mtime exactly
as a rewrite does, so an age measured from mtime resets either way, and the kill switch cannot tell a
fresher note from a touched one. Mitigating that downstream (three-valued `nudge_answered`,
segmented medians, F6's cross-check) was treating the symptom: the fact was never in a field.

**`CHECKPOINT.md` frontmatter goes to `schema: 3` and carries the facts the gauge reads:**

| Field | Type | Meaning |
|---|---|---|
| `written_at` | `<UTC ISO>` | when the note's **body last changed**. This is the age reference point |
| `reaffirmed_at` | `<UTC ISO\|null>` | when it was last confirmed still accurate without changing |
| `head` | `<short sha\|null>` | **existing field, with a new rule: it moves only when `written_at` moves.** A re-affirmation leaves it alone — see below |

**`body_sha` is deliberately NOT among them.** `CHECKPOINT.md` is agent-authored
prose (`commands/checkpoint.md`), so the only named writer would be typing a hex digest of the bytes
it is in the middle of typing, and a wrong digest is bytes-identical to a right one —
[[design-by-contract]]'s rule that a script must do what a script can do, inverted. **The hash belongs
to the machine**: `internal/checkpointseal` already snapshots the note at every seal, so it computes
`body_sha` there and writes it to the **seal row**, where drift is a comparison between consecutive
rows rather than a claim needing a prior value nobody stored.

**`head:` needed a rule, and its absence made §VI-c's claim false.** `SKILL.md`'s standing contract is
*"BEFORE writing the note, YOU MUST record `head:`"* — and a re-affirmation **is** a write. So branch
work (`note.head..HEAD`), one of criterion 1's three baseline distributions, would still have reset on
a touch, while growth and turns no longer did. §VI-c's "nothing for a touched note to score" was true
of two measures out of three. The contract therefore gains: **a re-affirmation sets `reaffirmed_at`
and touches nothing else** — not `head`, not `written_at`. Content facts move together or not at all.

**The three frontmatter fields are flat scalars, which is why they fit.** §VI-b refuses `handles:` because it is a
**list** and `internal/checkpoint.Parse` is deliberately not a YAML parser. That objection is about
shape, not about frontmatter, and it does not reach a timestamp or a hash — the existing parser reads
these with no change to its contract. §VI-b is narrowed accordingly rather than left to look like a
blanket refusal.

**What this retires.** Age is now `now − written_at`, immune to a touch; `reaffirmed_at` records the
other event as its own fact instead of being inferred from a mtime that moved. **mtime is demoted to
a cross-check** — the reverse of the earlier draft — and the seal row's `body_sha` makes it decidable:
a note whose `written_at` advanced between two seals while the body hash did not is a mis-written
note, reportable as an error rather than a disagreement someone has to adjudicate. **Both sides of
that comparison are computed by the machine, from snapshots it already takes** — neither is a value
the note claims about itself.

**Carriers of the schema bump**, censused 2026-08-23 — **plugin-wide, because a census scoped to
`tools/` cannot return its own table's rows**:

```bash
git grep -l "schema: 3"      -- plugins/          # 13 files: 11 migrated, plus commitssince_test.go and sealrow_test.go from the build
git grep -n -w "updated"     -- plugins/          # every reader of the retired field
git grep -n 'Get("schema")'  -- plugins/          # 1 hit
```

| Carrier | Change |
|---|---|
| `skills/context-checkpointing/SKILL.md:16-17` | `schema: 3`; **`updated:` removed**; `written_at`, `reaffirmed_at` added; and the contract that a re-affirmation sets `reaffirmed_at` and touches nothing else — not `head`, not `written_at` |
| `commands/checkpoint.md` | writes `written_at`/`body`-changes, or `reaffirmed_at` alone — that write IS the artifact; the sealer derives `nudge_answered` from which one moved, so no agent is asked to self-report whether it complied |
| **`commands/resume.md:13`** | *"Report the seam: the note's `updated` timestamp…"* — **a prompt contracted on the deleted field.** Becomes `written_at`, and reports `reaffirmed_at` beside it when set, since a note re-affirmed by another session is exactly the seam this step exists to surface |
| `internal/checkpointrestore` | `main.go:210` renders `Get("updated")` → re-point at `written_at`, render `reaffirmed_at` beside it; **3 `updated:` fixtures** in `main_test.go:16,149,193` |
| `internal/checkpoint` | no parser change (flat scalars), one accessor per field. **`checkpoint_test.go:17` hard-asserts `Get("schema") == "2"`** — the only consumer of the key anywhere, and it must move with the bump |
| `internal/checkpointseal` | computes `body_sha` from the snapshot it already takes, into the seal row |
| `internal/postcompactobserve`, `internal/filechangedrearm`, `internal/sessionstart` | read the note; additive fields, no behaviour change |
| `gray-area/tools/internal/claims/claims_test.go:18` | a `schema: 2` fixture — **outside `prosthetic-conscience` entirely**, which is why the census must be plugin-wide |
| **`schema: 3` fixtures — 13 files**: the two above plus `checkpoint_test.go`, `checkpointrestore/{compose,main}_test.go`, `checkpointseal/{drift,hook,main}_test.go`, `filechangedrearm/main_test.go`, `postcompactobserve/main_test.go`, `sessionstart/main_test.go` | inline string literals, not `testdata` — there is no such directory |

The remaining `git grep -n -w "updated"` hits are `frank-exchange-of-views` prose using the English
word, `hookcmd.go`'s `updatedInput` local, **and two prosthetic-conscience hits added by the build** —
`checkpointrestore/main.go` and `commitssince_test.go`, both comments explaining what `updated:` used
to mean. Adjudicated as unrelated rather than left unmentioned, because a census that lists only its
hits cannot be checked against a re-run.

**Schema 2 notes must not read as age zero.** Nothing consumes the `schema:` key today (verified: it
appears in comments and fixtures only), so the bump costs no migration — but a note without
`written_at` is **`Unmeasured`, not fresh**, and the gauge reports it as such. That is the same
tri-state discipline every other figure here uses, applied to the field this design turns on.

**TWO files, not one, and the split is a safety requirement rather than tidiness.**

`freshness.json` is written by the two RECORD writers (the sealer and the compaction observer) via
`json.Marshal` of a fixed struct — so **every write erases any key the struct does not declare**. Put
the band state in there and an unrelated hook, firing on an unrelated event, silently destroys the
record of which bands have been spent. **An erased band is a re-emission on `Stop`**, which is F10's
sixteen-firing loop: the highest-rated risk in this plan, reached through a file layout.

So: `freshness.json` holds the **gauge's** state, owned by the record writers. A second file,
`.claude/checkpoints/nudge.json`, holds the **nudge's** state and is owned by `sc-stop` alone. One
writer per file, no locking needed between them, and neither can erase the other. **The rule is about WRITES, not reads.** No process writes a file it does not own: the record
writers write `freshness.json` only, `sc-stop` writes `nudge.json` only. Reads cross freely and must
— the sealer reads `nudge.json` for `nudge_enabled` (its presence) and `answered_at_seen`
(`nudge_answered`'s derivation), and `sc-stop` reads `freshness.json` for the gauge. A read cannot
erase anything.

**`freshness.json` — the gauge's state, written by the record writers:**

| Field | Built? | Why |
|---|---|---|
| `written_at_seen` | **yes** | the note's `written_at` as last observed — the key deciding whether the reading below still belongs to this note |
| `tokens_at_write`, `dropped_at_write` | **yes** | growth's other term, which no record held. Both, so growth stays monotone across a compaction — **and both sides must be on the same footing**: if `dropped_at_write` is non-zero and the current read has no dropped figure, growth is `Unmeasured`, because subtracting a post-compaction stamp from a pre-compaction reading yields a large negative number that renders as the freshest note in the file |
| `has_write_reading` | **yes** | separates "stamped at zero" from "never stamped". The mechanism behind the "absent → `Unmeasured`, never 0" promise |
| `stamped_at` | **yes** | **when the baseline reading was taken, which is NOT when the note was written.** Nothing observes a note at its write: `Of`'s only callers are the sealer and the compaction observer, which run at boundaries. A note written at turn 10 and first sealed at turn 200 gets its baseline at turn 200, so growth afterwards is growth **since the seal**. That is a real interval and a useful one; it is not the one the field name suggests, and a reader cannot tell from the number. Both records therefore carry `growth_since`, so the gap is visible rather than assumed away — and criterion 1's growth distribution and F7's 50,000-token disjunct must be read against it |

**`nudge.json` — the nudge's state, written by `sc-stop` alone (Phase 2):**

| Field | Why |
|---|---|
| `session_id` | which session this state belongs to. **Without it the file outlives the session it describes**: `emissions_this_session` would be a lifetime counter, so four emissions ever would suppress the nudge permanently in that project, and "once per band per session" would have no session boundary at all. On a session_id that differs from the stored one, the counters reset and `bands_spent` is cleared — a new session has a new budget and has been told nothing yet |
| `bands_spent` | which of NOTICE/WARN/URGENT have fired **in this session**. **Not cleared on an answer** — see below |
| `answered_at_seen` | the later of `written_at`/`reaffirmed_at` at the moment a band was spent. This, not `bands_spent`, is what `nudge_answered` compares against |
| `emissions_this_session`, `emission_bytes_max` | criterion 4's cap, counted rather than intended. Never reset |
| `branch_head_seen`, `branch_commits_seen` | the per-`HEAD` branch cache criterion 3 requires, owned here because `sc-stop` is the only caller that runs per turn |

**Bands re-arm without being erased.** An earlier draft said `bands_spent` is "reset when the note is
answered", which made `nudge_answered` underivable: the derivation reads `bands_spent` to know a band
fired, so clearing it on an answer left `rewritten` and `reaffirmed` unreachable and only `ignored`
observable — the artifact that keeps the `SKILL.md` clause out of `unobservable-duty` would have
reported nothing, by construction. Instead the spent record **stays**, and a band re-arms when
`answered_at_seen` is older than the note's current `written_at`/`reaffirmed_at`. The fact is
"which bands fired against which version of the note", and that is one fact, not two.

**The hard cap is here, and it is separate from the band policy.** Bands are "at most once per band
per session, reset on write" — which permits three emissions per note-write cycle, so two checkpoint
writes in one session would exceed criterion 4's ≤ 4 by design. **`emissions_this_session >= 4`
therefore suppresses unconditionally, whatever the bands say**, and the seal row copies both counters
so the budget is *counted* rather than intended.

**Keyed on `session_id` alone**, because no per-seat emitter exists. `sc-stop` is the only writer, and `Stop`'s payload carries no `agent_*` fields at all
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


### `[MODIFY] internal/checkpointrestore` — the `--first-parent` defect

`commitsSince` gains `--first-parent` and the surrounding prose changes from "commits ago" to
"commits on this branch". This is a **correctness fix to shipped behaviour**, evidenced in §II
(109 vs 24), and it stands independently of everything else here. **The restore digest carries the BRANCH measure only, and that is a payload fact rather than a
choice.** `SessionStart`'s stdin is `{session_id, cwd, hook_event_name, source}` — measured, spike **§9e correction 3** —
**with no `transcript_path`**. Turns and growth both need the transcript, so at this event they
cannot be computed at all. What the digest can say about age it already says: `staleness()` renders
the branch count, which needs only git. The other two ride the seal record, where the sealing events
do carry a transcript path.

Deriving the transcript path from `session_id` and `cwd` was considered and **refused**: that recovers
a path by assembling a string, which is the shape this suite is named after, and it breaks silently
the moment the projects-directory convention moves. Recorded here so a later author finds the refusal
rather than re-deriving it — an earlier draft of this section assumed the digest could carry the whole
gauge, and building it is what found the payload.

**Consumer census for the digest TEXT** — the rendered line is injected at `SessionStart`, so it is a
contract, not a log message. `git grep -n "commits\? ago" -- plugins/`:

| Hit | Adjudication |
|---|---|
| `internal/checkpointrestore/main.go` (the two `staleness` writers) | **changed** — line numbers deliberately not pinned: they moved with the build, and a citation that rots is worse than one that names the function |
| `internal/checkpointrestore/compose_test.go:192,195` | hard-assert `"written 1 commit ago (abc1234)"` and `"written 12 commits ago (abc1234)"` — **change**; two shipped assertions that break on this edit, which an earlier draft called independently shippable without naming them |

No other reader: nothing greps the digest, and the injected text has no machine consumer.


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
(`sc-precompact`, `sc-sessionend`, `sc-subagentstop`). **BUILT** — `sealrow.go`, carrying the seal-side fields,
the three age figures and their measured flags, `written_at`, `nudge_enabled` and `nudge_answered`. Not pruned — a row is ~200 bytes and the
baseline needs history. The markdown snapshot and its comment stamp stay exactly as they are: they
serve human recovery, and [[facts-are-fields]] does not ask for prose to be stripped, only for the
machine-read fact to live in a field.

| Field | Meaning |
|---|---|
| `at`, `event`, `occasion`, `session_id`, `agent_id` | the stamp's five facts, as fields this time — **built** |
| `note_age_turns`, `note_growth_tokens`, `note_branch_commits` | criterion 1 — **built**, each with its own `*_measured` flag and **omitted when unmeasured**: a zero meaning "could not tell" would pull every median toward fresh |
| `seal_trigger` | `precompact` \| `sessionend` \| `seat_return` — which event sealed — **built** |
| `live_handles` | count of `background_tasks` entries with `type != "subagent"`, plus `session_crons` — **only when measurable**. Seats are excluded deliberately: at a `seat_return` seal the returning seat appears in the parent's own list (§12), and counting it reads high by exactly one while answering a different question from "did this note miss some background work" |
| `handles_measured` | `false` when the payload carries no `background_tasks` key |
| `nudge_enabled` | whether the nudge was live when this seal was written — **read from the presence of `.claude/checkpoints/nudge.json`**, which only `sc-stop` creates and which exists from its first run in a session. Not a build tag (invisible to a consumer, and wrong the moment a binary is swapped), not an env var (unset in the hook's environment as often as not), not `hooks.json` (registered ≠ running). The file is written by the thing whose liveness is the question, so its existence IS the answer — criterion 6's falsification groups on it, and it must be recorded per row rather than inferred from dates |
| `nudge_answered` | `rewritten` \| `reaffirmed` \| `ignored` \| **`n/a`** — three values plus the Phase 1 case, **derived by the sealer, not written by an agent**: if no band has fired the value is `n/a` (Phase 1 ships with no nudge, so every row carries this); otherwise compare the note's `written_at` and `reaffirmed_at` against `nudge.json`'s `answered_at_seen` — `written_at` newer → `rewritten`, `reaffirmed_at` newer → `reaffirmed`, neither → `ignored` |
| `body_sha` | hash of the note's body, **computed here** from the snapshot this package already writes — not read from the note. Drift is a comparison between consecutive seal rows, which needs no prior value held anywhere else |
| `emissions_this_session` | copied from `nudge.json`'s monotone counter at seal time, and `emission_bytes_max` beside it — **criterion 4's budget is stated as "counted in the observation row rather than intended", and until this field existed nothing counted it**. The debounce file already holds the number; the seal row is where it becomes checkable after the fact |

**`handles_measured` is not defensive padding; it is the whole of the column's honesty.** Measured
per event 2026-08-23 on 2.1.240 and recorded as `hook-surface-spike.md` **§12** — key sets read from
raw hook stdin:

| Sealing event | carries `background_tasks`? |
|---|---|
| `SubagentStop` | **yes, and populated** — measured with a shell task AND a seat in flight |
| `PreCompact` | **NO — the key is absent from the payload entirely** (4/4 firings) |
| `SessionEnd` | **NO — the key is absent** (payload: `cwd, hook_event_name, prompt_id, reason, session_id, transcript_path`) |


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
| `internal/checkpointseal/{main.go,main_test.go,drift_test.go,hook_test.go}` | the package | **YES** — new writer plus tests. **`hook_test.go`'s `snapshots()` helper listed the whole directory**, so it counted `seals.jsonl` as a seal and broke five unrelated assertions the moment the record appeared. Scoped to `*.md`, the same rule `prune()` uses — a helper that measures more than the code manages is a carrier, and this census did not name it until the build hit it |
| `commands/resume.md:7` | `--seals` tells the agent to list snapshots "with their sealed-at stamp, trigger and agent" — a **prompt-side contract against the comment string** | **No** — the stamp is unchanged. This is the carrier that makes the stamp load-bearing; retiring it starts here |
| `internal/checkpoint/checkpoint.go:239` (`NoteLoopProblems`) + `checkpoint_test.go:101` | the note body. The test fixture contains `<!-- sealed: trigger=auto -->` — **a stamp shape the writer never emits** (it writes `event=`/`occasion=`) | **No** — `Parse` strips scaffolding and never reads the stamp's keys. Flagged because a fixture asserting a format nothing writes is how a format's real readers get miscounted |
| `gray-area/tools/internal/claims/{claims_test.go, adjudicate_test.go}` | the two files the commands actually return — fixtures of a sealed note's body | **No** — snapshot format unchanged. (The package doc of `claims.go` describes reading a sealed note, but the file matches neither command and is listed here for the reader rather than as a census hit) |
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
  (spike §13: `false` on the first firing, `true` on all fifteen re-entries).** — evidence on the channel the design uses: if the turn is already a Stop-hook continuation, emit nothing regardless of
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

**Nothing gates hook registration.** `grep -c "hooks.json" scripts/pluginparity/main.go`
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

> **Three values, because two of them are indistinguishable from the outside.** `rewritten` and
> `reaffirmed` are both legitimate answers to a band, and only the note's author knows which one
> happened — the age fields record *when*, this records *which*. `ignored` is the one the mechanism
> exists to make visible: it is F1's wallpaper signal, and without it a band that fires into silence
> looks identical to one that was answered.

The value goes in the **seal row**, a record with fields — never composed into a timestamp field,
which is the shape §III indicts elsewhere.


**Consumer census** — `git grep -ln "context-checkpointing" -- plugins/ scripts/ docs/`, full output,
every hit adjudicated:

| Hit | Relationship | Changes? |
|---|---|---|
| `skills/context-checkpointing/SKILL.md` | the surface itself | **YES** — the clause above |
| `commands/checkpoint.md`, `commands/resume.md` | invoke the skill; `checkpoint.md` enumerates the note's sections; `resume.md:13` reports the seam from `updated` | **YES, both** — `checkpoint.md` writes `nudge_answered`'s trigger; `resume.md` reports `written_at` in place of the retired `updated` |
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
| F3 | **Commit count reads other people's work as mine.** | H×M×L | `--first-parent`; no author filter (§II measured both). | Ph. 1 |
| F4 | **Gauge cost on a hot path.** | M×M×L | Bounded tail; **cache the PARSED NOTE keyed on mtime** — re-read the file only when mtime moves. mtime can save a *parse*; it cannot save the transcript read, which is the cost criterion 3 budgets, and it must never gate the MEASUREMENT: growth, turns and branch work all advance while the file sits untouched, which is the case this design exists to catch. | Ph. 2 |
| F6 | **A writer advances `written_at` without changing the body** — the note claims to be fresh and is not. | L×M×L | `checkpointseal` computes `body_sha` from the snapshot it already takes; a `written_at` that advanced between two seals while the hash did not is reported as an **error**. Both sides are computed by the machine — neither is a value the note claims about itself. | Ph. 2 |
| F7 | **Short sessions get lectured.** | M×L×L | The gauge arms only above a **floor fixed here, not later**: a note exists AND (**≥ 20 assistant turns since it was written** OR **≥ 50,000 tokens grown**). Below that, silent — no band, no emission. The numbers are deliberately generous: a session that has done less than either has nothing a nudge could usefully say. Phase 2's acceptance asserts zero emissions below the floor, so the floor must exist by then. | **Ph. 2** |
| F8 | **Age is read as truth.** A fresh note is not a correct note. | M×M×L | The skill clause says so; the render states a measure, never a verdict. gray-area's `/audit-checkpoint` remains the instrument for the note's *claims*, and the two are deliberately different tools. | Ph. 3 |
| **F10** | **`Stop` injection loops** — measured, not theorised, and **worse on the current client**: 9 firings / 1,186 output tokens on 2.1.235, **16 firings / 35 assistant entries / 4,326 output tokens on 2.1.240** (spike §13). The cap is undocumented and moved between two patch versions. | **H×H×L** · residual after mitigation **L×M** | Write-before-emit; `stop_hook_active` as an independent second brake; a loop regression test asserting the empty emission on a spent band. §III. | Ph. 2 |
| F11 | **`seals.jsonl` grows without bound** — deliberately not pruned, reversing the `keepSnapshots = 10` precedent beside it, whose comment cites a resource problem this repository already hit once. | L×L×L | A row is ~200 bytes and the file is gitignored: 20 boundaries ≈ 4 KB, a year of heavy use is single-digit MB. **Pruning is what made the old seal unusable as a record** (§III) — the baseline needs history. Phase 3 sets a retirement point if the measured row size exceeds the estimate. | Ph. 3 |
| F9 | **Hook-surface churn** (R9, realised **three times across three client versions**: `SubagentStop` changed behaviour between 2.1.220 and 2.1.240; `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` went inert; and the `Stop` loop's cost went 9 firings/1,186 tokens → 16/4,326 between 2.1.235 and 2.1.240). | M×M×M | Channels are measured per client and each row says which: §8 on **2.1.235**; §9–§13 on **2.1.240**. Phase 2 re-confirms at build time as version-drift insurance. **Not counted here:** the `matcher`-key finding (§9e correction 1) is a measurement artifact on a single client — a dead driver blamed on a missing matcher — and folding it into a churn count would inflate the risk this row exists to size. | all |

---

## V. Verification Plan

### Phase 1 — gauge, defect fix, baseline (no nudge)

```bash
(cd plugins/prosthetic-conscience/tools && go test ./internal/ctxusage/... ./internal/freshness/... \
    ./internal/checkpointrestore/... ./internal/checkpointseal/... && go test ./...)
```

Coverage must include the negatives, which are the point: a tail containing **no** assistant entry;
a truncated final line; a session with **no** compact boundary → `Dropped: Unknown` (growth then computed from the raw
counter alone, which is correct when nothing has been dropped); two boundaries → most recent wins;
**the render contains no `%` on any input** — asserted over a fully-measured gauge, not just an
empty one; **a schema-2 note lacking `written_at` → `Unmeasured`, NOT age zero**; **a `written_at` older than
the scan window → `Turns: Unmeasured`, never a partial count** (the stalest-note case, asserted
against a fixture whose note predates the window); **`tokens_at_write` absent → growth `Unmeasured`,
never 0**; **a transcript with a compact boundary between write and now → growth positive, computed
through `cumulativeDroppedTokens`**;  a note head that is
unreachable; a branch whose history contains merges → first-parent and plain
counts differ; **a payload with no `background_tasks` key → `handles_measured: false` and NO
`live_handles` field** (the `PreCompact`/`SessionEnd` case, measured §12); **and the case that
discriminates — `background_tasks: []` PRESENT but empty → `handles_measured: true`, `live_handles:
0`** (the normal `seat_return`, §9d). Without the second, the suite passes for a `[]Task` decode that
cannot tell absent from empty, and the only trigger that can measure handles is silently dropped from
the baseline; **a seat's own entry excluded from the count** (§12: a live seat appears as
`type: "subagent"` in the parent's list); **a seal row written on each of the three
trigger events**, asserted per event rather than per package.

**Driveable check on real data — and it MUST exercise the measure this plan moved.** Every note in
this repository today is schema 2, so a run against "a real `CHECKPOINT.md`" exercises only the
`Unmeasured` branch and never `now − written_at`. Both cases are required:

1. a real schema-2 note → **`Unmeasured`**, and no age rendered;
2. a note actually written under schema 3 → age hand-checked against its own `written_at`, and the
   branch-work count hand-checked against its `head`;
3. the token figure hand-checked against `tail -c 200000 … | jq '.message.usage'` on a live multi-MB
   transcript. **Done: 525,066 from a 3.1 MB transcript, matching `jq` exactly**, with `Turns`
   correctly `Unmeasured` (the file exceeds the window) and `Dropped` unmeasured (that session had
   never compacted). Kept as `TestAgainstARealTranscript`, driven by `SC_REAL_TRANSCRIPT` and skipped
   without it — a machine-specific path must not turn into a red that means "not applicable here".


Then, in this repository:

```bash
# Pinned to a fixed range: these numbers move with every commit, and an expectation
# that re-arms on each commit fails for reasons unrelated to the defect it guards.
R=24f8fc63622f39797e5b4103c003f5aa1465138b..1105a02
test "$(git rev-list --count --first-parent $R)" -lt "$(git rev-list --count $R)" || echo FAIL
```

The **relation** is the invariant — first-parent strictly less than total across a merge-bearing
range — not either number.

**Then collect the baseline** — ≥ 20 real boundaries, seals stamped, nothing emitted. The record is
`.claude/checkpoints/seals.jsonl`, written by `internal/checkpointseal` on all three sealing events:

```bash
# Each distribution filters on ITS OWN measured flag. Filtering once on turns and then
# building all three arrays lets rows with turns measured but growth absent contribute
# null, which sort places first and which moves the median index.
jq -s '{n: length,
        turns:   ([.[] | select(.turns_measured)  | .note_age_turns]    | sort),
        growth:  ([.[] | select(.growth_measured) | .note_growth_tokens]| sort),
        commits: ([.[] | select(.branch_measured) | .note_branch_commits] | sort),
        measured: {turns:  ([.[] | select(.turns_measured)]  | length),
                   growth: ([.[] | select(.growth_measured)] | length),
                   branch: ([.[] | select(.branch_measured)] | length)}}' \
   .claude/checkpoints/seals.jsonl
```

And the two folded-in observations. **Both queries filter on measurability first**, because the
absent case and the honest zero are otherwise the same bytes:

```bash
# #506: is a note sealed with live background work staler than one sealed without?
# Rows where the payload could not tell us are EXCLUDED, not counted as zero.
jq -s '[.[] | select(.handles_measured)] | group_by(.live_handles > 0)
       | map({live_handles: .[0].live_handles > 0, n: length,
              median_age: ([.[] | select(.turns_measured) | .note_age_turns] | sort | .[length/2|floor]),
              turns_measured_n: ([.[] | select(.turns_measured)] | length)})' \
   .claude/checkpoints/seals.jsonl

# ...and how much of the baseline that excluded, reported rather than inferred from a small n:
jq -s '{total: length, measured: ([.[] | select(.handles_measured)] | length)}' \
   .claude/checkpoints/seals.jsonl

# #507: how stale is the note at a seat return, against the other triggers?
jq -s 'group_by(.seal_trigger)
       | map({trigger: .[0].seal_trigger, n: length,
              median_age: ([.[] | select(.turns_measured) | .note_age_turns] | sort | .[length/2|floor]),
              turns_measured_n: ([.[] | select(.turns_measured)] | length)})' \
   .claude/checkpoints/seals.jsonl
```

**A result of "no difference" is only readable if all three triggers are present.** Assert that
first — `map(.seal_trigger) | unique | length == 3` — because a query returning one group looks
identical to a query whose other groups genuinely did not differ. Expect `sessionend` to be the thin one: it did not fire in **either** of the two runs that ended with a background task still running (0 of 2, spike §12 — with the pipe-close confound named there and carried in §III).

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

**Both gates carry an UNMEASURABLE case, and it is not optional.** An absolute wall-clock budget on a
shared machine measures the neighbours: this gate read **p95 191 ms at load average 11**, against a
raw seek-and-read floor — no parsing at all — of **13 ms**, over the same 256 KB. Blaming the code
for that is how a gate teaches people to ignore it. So each run measures the unavoidable read
alongside the real one, and **when the floor alone exceeds the budget the gate SKIPS with that
number stated**, rather than failing. A budget that cannot be evaluated here reports "not measured",
which is the same tri-state the thing it is testing applies to everything else.

Note what was NOT done: the first fix attempt replaced the budget with a multiple of the raw read,
and the multiple would have been chosen after seeing that the result was 3.43×. That is the post-hoc
freedom §III forecloses for the nudge's own thresholds, and it applies to the gates too.

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

**It is a lower bound, and must be reported as one.** These rows exist only for sessions that
**sealed** — and `SessionEnd` fired in 0 of 2 runs that ended with live background work (§III). A
session that emitted five nudges and never sealed produces the same output as one that emitted three,
so the absent case and the honest pass are again the same bytes. Report the denominator beside it:

```bash
# emitting sessions observed, against sessions that started at all
jq -s '{sealed_sessions: (group_by(.session_id) | length),
        emitting: ([.[] | select(.emissions_this_session > 0)] | group_by(.session_id) | length)}' \
   .claude/checkpoints/seals.jsonl
```

If the sealed count is far below the sessions actually run, `worst_count` is measuring the sessions
that ended tidily, which are not the ones the budget is protecting against.

**F6's drift check, which is what makes `body_sha` a measurement rather than a stored string.**
Nothing read it before; a hash with no reader is a column, not a check:

```bash
# A note whose written_at ADVANCED between two seals while its body hash did NOT is
# claiming to be fresh without having changed. Reported as an error, not adjudicated.
jq -s 'sort_by(.at) | [ .[] | {at, written_at, body_sha} ]
       | . as $r | range(1; length) as $i
       | select($r[$i].written_at > $r[$i-1].written_at and $r[$i].body_sha == $r[$i-1].body_sha)
       | {at: $r[$i].at, written_at: $r[$i].written_at, note: "written_at advanced, body did not"}' \
   .claude/checkpoints/seals.jsonl
```

Empty output is the healthy state — and, this being the shape this plan is about, an empty result is
also what a query over an empty file returns, so it is run beside the row count above.

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
**relation, not a constant**. A constant copied from one client's run would be wrong on the next: §13 measured this channel's counts moving 3.6× between two patch versions.

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
# EVERY array filters on turns_measured. A row with turns unmeasured omits the key, jq
# yields null, sort places nulls FIRST, and the median index moves — on the statistic
# that decides whether the nudge is removed.
jq -s '[.[] | select(.turns_measured)] |
       {before:           [.[]|select(.nudge_enabled==false)|.note_age_turns]|sort,
        after_all:        [.[]|select(.nudge_enabled==true )|.note_age_turns]|sort,
        after_rewritten:  [.[]|select(.nudge_enabled==true and .nudge_answered=="rewritten")|.note_age_turns]|sort,
        after_reaffirmed: [.[]|select(.nudge_enabled==true and .nudge_answered=="reaffirmed")|.note_age_turns]|sort}' \
   .claude/checkpoints/seals.jsonl
# and the count this excluded, reported rather than inferred:
jq -s '{total: length, turns_measured: ([.[]|select(.turns_measured)]|length)}' \
   .claude/checkpoints/seals.jsonl
```

**The verdict is `after_all` against `before`** — §VI-c closes the population question in prose and
§V must not reopen it at the point of execution, in the one loop that is the kill switch. If *that*
median does not fall, the nudge is removed and the gauge stays. `after_rewritten` and
`after_reaffirmed` are **diagnostic**: they say how the duty was discharged, and a large
`after_reaffirmed` population with an unchanged `before`/`after_all` comparison is the F1 wallpaper
signal, not a second verdict to weigh against the first.

### The gate

`/prosthetic-conscience:plan-audit` before any code. `versionguard` unaffected — plugin content,
and the version moves at a release boundary.

---

## VI. Deliberately not in this plan

Filed as issues, not built here. Every one has been **measured** — `hook-surface-spike.md` §9 (the 30-event
catalogue census, 15 sessions), §10, §12 and §13 — so what follows is a verdict on evidence, not a
deferral pending it.
**#505 is answered and folded in**: four channels inject, `Stop` is the one worth having, and its loop
hazard is F10 with a test rather than a discovery waiting to happen in Phase 2.

### VI-a. The docket, decided on measurement

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

### VI-c. Criterion 6 needs no population choice

Age is `now − written_at`, which moves only when the body changes. A re-affirmation sets
`reaffirmed_at` and leaves the age alone, so criterion 6's median is computed over **every seal** and
a touched note cannot score as a fresher one.

`nudge_answered`'s three values remain as a **diagnostic**, not a second verdict: they say *how* the
duty was discharged, and a high `ignored` rate is F1's wallpaper signal whatever the medians do.

### VI-b. Blocked: the note has no field for its handles


#506's strong claim — *the note's "In-flight handles" section is provably wrong* — needs handle ids
the note does not carry. **This section is about a LIST**, and its objection is about shape:
`internal/checkpoint.Parse` is deliberately not a YAML parser. It does not reach the flat scalars
§III adds. `CHECKPOINT.md` **already is structured data**: versioned frontmatter
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

So the strong form costs, **on top of the schema-3 change §III already makes**: a `Parse` that can report a parse failure instead of skipping,
and carriers at `context-checkpointing/SKILL.md`, the four consumers, gray-area's `claims`/`manifest`
readers, and the goldens. **That is a project, not a field**, and Phase 1's `live_handles` column
exists to say whether it is worth starting.

**What the census changes about the plan's own risks.** **F10 keeps its `H×H×L` rating and gains a
stated residual of `L×M`** (§IV, where the table's pre-mitigation convention is now written down; an
earlier draft of this paragraph announced a re-rating to `L×M×L`, which would have made F10 the one
row not comparable with the others). The measurement is undisturbed — 9 firings and 1,186 wasted
tokens stand. What is low is the **residual**, because the hazard is cycle detection and two
independent brakes are already specified: `stop_hook_active` is a field the client sets on every re-entry — measured on `Stop` itself (§13: `false` on the 1st firing, `true` on all **fifteen** that followed), needing no state and no write, and the debounce file carries
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
