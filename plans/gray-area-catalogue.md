# The catalogue: gray-area as a trajectory miner with a SQL surface

> Status: **under review** — this plan is the decision document; no code lands until it merges.
> Companion to [`gray-area.md`](gray-area.md), whose Phase 3 this implements.

**Client under test:** Claude Code 2.1.260. Every figure below was measured on this box on
2026-09-08 with the command named beside it. Three of them overturned the design that preceded
them — hook frequency, the `ATTACH` cap, and the meaning of an absent `is_error` — and the one
figure that is an estimate says so.

---

## I. Summary & Goals

**Gray Area is a trajectory miner: it answers questions about the behaviour and intent of agents
from their recorded actions, words and thoughts** (gblock, 2026-09-08). Today it records only
*where* a trajectory is, in a per-project manifest, and nothing can read across sessions.

The cost of that is measured, not hypothetical. On 2026-09-06 two sessions argued to a standstill
over a census drawn from **1 of 28 project directories** and reported as a total; both conclusions
were wrong and the artifact could not arbitrate. On 2026-09-08, with a release tag held and six
agents live on this box, **every coordination failure of the day was a query nobody could run** —
three claims relayed from memory that the record would have refuted, a shared blocker board gone
stale because its rows were hand-copied, and four messaging round-trips to ask what was
outstanding.

So the surface is a **switchboard**: *which agents are active, who is working on what, is anyone
else editing this file, what did that agent actually do — and what was it trying to do.*
`SendMessage` asks an agent what it **believes** and needs it alive and willing. This asks the
record what **happened**.

Goals, in priority order:

1. **A queryable surface, not a fixed menu.** Baked verbs for the common questions, and **raw
   read-only SQL** for everything else, because the questions worth asking are not knowable in
   advance. The schema is therefore a **published contract**, not an implementation detail.
2. **Ingest stays inside stated latency budgets: 20 ms per turn, 50 ms at `SessionStart`
   steady-state, 500 ms on the one `SessionStart` per UTC day that runs the 325 ms sweep, and
   500 ms at `SessionEnd`.** The last is bounded from above by the vendor: all `SessionEnd` hooks
   share a 1.5 s budget (per the hooks reference; raisable only by a per-hook `timeout`), and
   gray-area shares it with prosthetic-conscience's `sc-sessionend`. gray-area does **not** raise
   `timeout`; it stays at a third of the shared budget. Four numbers, because each is a different run — an earlier draft stated the middle
   one alone, which its own measurement refutes once a day.
   Not "zero added latency" — an earlier draft said that while also making `SessionStart` do work,
   so the goal could not be traced to any passing check. A budget with a number can be. §II shows
   the budget still rules out the obvious design by an order of magnitude.
3. **Actions, words and thoughts** — all three, because intent is not recoverable from acts alone.
4. **Answers are conclusions, not dumps** ([[context-efficiency]]).

**Non-goals.** Cross-machine aggregation; semantic search; a resident daemon (§II); an MCP surface
(§III names why it is scoped out); replacing `SendMessage` for intent and negotiation.

---

## II. Design

### The measurements that decide it

| measurement | result |
|---|---|
| bare Go process spawn — the floor under any hook | **4.33 ms** |
| tool calls per turn / turns per session | **8.8** / ~131 |
| `PostToolUse` ingest — busiest measured session, 2,187 calls (`5627f39a`) × spawn | **9.5 s**; at the ~1,150-call average (8.8 × 131) **5.0 s** |
| **post-turn (`Stop`) ingest — 131 turns × spawn** | **0.57 s per session** |
| SQLite insert on an open connection | ~4 µs (251k rows/s) |
| indexed query at 1M acts — *who touched this path* | **1.00 ms** |
| sequential open of 30 daily partitions | **14.6 ms** |
| `ATTACH` limit | **10 per connection** (main + 10 = 11), **not per process** — 60 separate connections opened fine |
| all three tiers, 3 weeks of corpus | **11.8 MB = 4.2%** of 280 MB |
| ripgrep over the whole corpus (real binary) | **33–54 ms** |
| `DELETE` one day (33k rows) from a 1M-row table | **325 ms**; file plateaus, no `VACUUM` |
| `is_error` absent on a `tool_result` | **18.4%** of calls — the normal SUCCESS shape for every tool but Bash |
| FTS5 index cost | **1–2 GB, ESTIMATED — not measured** |

### Ingest is a post-turn hook, not a daemon and not per-tool-call

`PostToolUse` fails goal 2 by an order of magnitude no store choice can fix: at the 4.33 ms spawn
floor it is **5.0 s per average session** (8.8 calls × 131 turns ≈ 1,150) and **9.5 s on the
busiest measured** (2,187 calls, session `5627f39a`), before the hook does anything. An earlier
draft quoted "2,000 calls" against the 131-turn average, which do not describe the same session.

Post-**turn** is a different frequency entirely — 8.8 tool calls per turn, ~131 turns per session,
so **0.57 s per session**. The `Stop` hook ingests incrementally from a per-file byte offset, so
steady-state work is proportional to new bytes, not corpus size.

**The transcript may not yet hold the turn that just ended — and that is UNMEASURED here, by
necessity.** The vendor's hooks reference states that at `Stop`/`SubagentStop` the transcript file
can lag, and that a hook needing the final assistant text should read `last_assistant_message` from
the payload instead. Byte-offset ingest recovers a lagged turn on the *next* `Stop` — but a
session's final turn has no next `Stop`, so its `word` row would never be hook-ingested. This
cannot be measured from anything on this box: gray-area has never bound `Stop`, and an attempt to
infer it from `SubagentStop` rows produced "24 of 41 sessions" with deltas up to 46 hours — which
is sessions *continuing* after their last subagent, not a transcript trailing a hook. That number
is recorded here so nobody re-derives it as lag. The design therefore assumes the lag is real and
closes the final-turn hole two ways: **(a)** `Stop` ingests `last_assistant_message` from its own payload as a **provisional** `word`
row keyed `(session, prompt_id)`, where `prompt_id` comes from the payload.

**And the two ids are the same namespace — verified, not assumed.** The payload's `prompt_id` and
the transcript's `promptId` sharing a name is not evidence they share a value ([[facts-are-fields]]
clause 4: a carrier is a site that speaks the same CONCEPT, not one that shares a STRING). Measured
across the 13 sessions where gray-area's manifest recorded a `prompt_id` and the transcript
survives: **434 of 434 hook-recorded values appear in that same session's transcript, 0 missing.**
The join is on a real key.

**One leg of that remains untested and is named rather than assumed.** Every one of those 434 rows
is a `SubagentStop` — gray-area has never bound `Stop`, so *that a `Stop` payload's `prompt_id`
names the turn which just ENDED, rather than the next one, is not established here.* If it named
the next turn the provisional would be filed against the wrong key. §V.17 asserts it at fire time,
alongside the lag measurement, because both need the hook to exist first.

**Supersession is by id, and an earlier draft's claim that it could not be was a fourth unmeasured
negative.** That draft said transcript assistant records "carry `uuid`, `requestId` and
`parentUuid` and **no `prompt_id`**", and built a byte-positional rule on it. True of an assistant
record read alone; false of the transcript as a graph. Re-measured over the 40 most recently
modified transcripts: `promptId` is carried by **`user`** records — 8,837 in that 40-file sample, **26,934 across the whole corpus** — and
**46,043 of 46,046 assistant records — 99.9935% — reach one by walking `parentUuid` to the nearest
ancestor that has it**, measured over the whole corpus (414 files). The 3 that do not are handled by
rule 6; an earlier draft said "15,675 of 15,675, 100.000%" from a 40-file sample that happened to
contain none of them. So the transcript side yields the same key the payload
carries, and the positional rule is withdrawn.

The rule. **An earlier draft called this "a total function" and it was not** — three measured
facts break the simple version, and each is stated here because each removed an option:

- **66.2% of turns carry MORE than one assistant text block** under a single `promptId` — 1,091 of
  1,648, whole corpus (416 files), up to 138. An earlier draft gave 492 of 709 from a 40-file sample
  with no scope named, the class round fourteen failed on. So "delete the provisional when any text for that key arrives" loses
  the final text on the majority case: the documented lag shape is *earlier blocks flushed, final
  block not yet*, so that ingest writes texts 1..k, deletes the provisional, and the tail is never
  stored. **That is loss, not over-keeping**, and it reopens the hole (a) exists to close.
- **Subagent transcripts share the parent's `sessionId`** — 60 of 60 measured, 263 on this box —
  and the parent's `promptId`. So `(session, prompt_id)` is not unique per agent, and a subagent's
  text would delete the parent turn's provisional.
- **The payload key can be absent — but the rate an earlier draft gave was mostly my own test
  noise.** It said "51 of 672 `SubagentStop` rows carry no `prompt_id`". Of those 51, **50 arrive in
  two consecutive seconds with `session_id` empty as well, every one written by a `+dirty` build** —
  synthetic payloads from this plan's own experiments, not a vendor shape. Sharper still, on the
  auditor's own scoping: **all 51 come from two scratch manifests** under `~/scratch/`, and of **583
  rows from real manifests, 0 lack `prompt_id`**. So the rate is not evidence of anything about the
  payload.

  **The rule therefore rests on the vendor's documentation, not on this rate:** `prompt_id` is
  *"absent until the first user input"*. That is a stated shape rather than a local frequency, and it
  is enough to need the rule — which is why the rule is unchanged while the number that appeared to
  justify it is withdrawn.

  Worth recording as a side-effect: the only reason those 50 are separable from real data is
  `capture_build` (#818). Without it they would be 50 indistinguishable rows inflating a rate by
  50×, which is the defect that field was added to prevent, catching a case it was not designed
  for.

So the rule is:

1. **Key is `(session, agent_id, prompt_id)`** — the agent dimension is required, not optional,
   because sessionId and promptId are both shared with subagents. `agent_id` is `""` for the main
   agent, which is itself a distinct key.
2. **Whether a provisional is written is decided by the fire-time comparison in rule 5** — not
   unconditionally. An earlier draft said "written at every `Stop` that carries a `prompt_id`",
   a universal rule 5 contradicts. When the payload has
   none, **no provisional is written and the omission is recorded** as `provisional_skip` with
   the reason — never keyed on `""`, which would collide every such turn onto one row.
3. **An unclosed provisional IS READABLE, and that is what keeps the hole shut.** It appears in the
   `word` view with `source='payload'` and `provisional=1`, so a lagged final turn is visible from
   the moment its `Stop` fires. Closure decides whether that row is deleted or made permanent —
   never whether it can be read. An earlier draft left this unstated, which is exactly what made
   "(a) is sufficient alone" look unsupported: it argued the write and not the visibility.

4. **CLOSURE INGESTS BEFORE IT DECIDES**, and an earlier draft that omitted this had promotion
   backwards. It said a provisional promotes to "a final block that never reached the transcript" —
   but §II:89 says such a turn is *never hook-ingested*, **not** never written, and the file does get
   it: measured over **all** completed transcripts (>30 min untouched) — **406 of 409, 99.3%** (3 have
   assistant records for the final turn and no text block; 5 more have no resolvable final key). An
   earlier draft said "150 of 150" because its scan took the first 150 files, the narrow-sample class
   this plan names elsewhere — and the direction matters: a true 100% would make §V.3's second addend
   always zero, the condition §V.3 itself treats as the failing case. A closing sweep comparing only against already-*ingested* text would therefore
   promote a row whose text sits in the file unread, on the routine path, and §V.3's arithmetic would
   be wrong in both halves — passing only when its second addend is 0, i.e. only when R11 has no
   evidence. That is the inverse of round seventeen's error, not its repair.

   So the sweep **runs incremental ingest from the stored offset first**, and only then decides.
   **Which file, precisely — an earlier draft said "that session's transcript", which is wrong for a
   subagent.** A session maps to many files (measured: 324 of 420 transcripts here are
   `agent-*.jsonl`), offsets are recorded **per file**, and a provisional's `agent_id` selects the
   one: `agent_id = ""` reads the session's own `transcript_path`, and a non-empty `agent_id` reads
   that seat's `agent_transcript_path` — the value the payload already carries and the manifest
   already records. Reading the main transcript for an `agent_id`-keyed provisional would find the
   text in the wrong file, never satisfy it, and promote every subagent turn. **That is transcript I/O for another session inside `SessionStart`, so it
   was measured against that hook's budget rather than assumed into it:** a tail read beyond a stored
   offset costs **2.4 ms**, and even a cold full parse of the largest transcript on this box (36.2 MB)
   is **108 ms** — against the 500 ms rollover budget, with 6 sessions live. The stored offset is what
   keeps the steady state in the low milliseconds; the cold case is the bound, and it fits. Promotion is thereby restricted to text genuinely absent from the file, and
   rule 4's justification becomes true rather than assumed. **A provisional is deleted the moment it
   is SATISFIED** — an ingested block equal to it arrives for its key — not on first sight of any
   sibling text. If it is still unsatisfied *after* that ingest, it is retained until its
   turn is known closed, by any of **four** triggers: a later `prompt_id` ingested for that same
   `(session, agent_id)`; `SessionEnd`; or — the one that cannot be missed — **the session no
   longer being live**, per the `~/.claude/sessions/<pid>.json` check §II already specifies for
   liveness; and a **fourth, age — any provisional older than 24 hours**.

   **Closure is a database WRITE, so triggers 3 and 4 need a writer named, and an earlier draft's
   "no hook has to fire for it to resolve" was false.** The query path is read-only by design
   (§V.9 asserts writes are refused there), so nothing resolves without some hook running. **There are TWO sweeps and an earlier draft conflated them**, which made the resolution claim and
   the retention claim contradict: the **retention** sweep (the 325 ms `DELETE`) is a no-op except on
   the first `SessionStart` of a new UTC day, while **closure runs on EVERY `SessionStart`** — that is
   what makes "the next `SessionStart` on the box resolves it" true. Closure is cheap enough to
   belong there: **2.4 ms** for a tail read beyond a stored offset, so even every live session
   outstanding at once sits inside the 50 ms steady-state budget, and the 500 ms rollover budget
   covers the day the two coincide.

   **And closure ingests for any session that is no longer live, whether or not a provisional is
   outstanding.** Without that, §V.3's equality is one-sided by construction: the 4.1% of fires
   carrying no `last_assistant_message` and the 0.12% duplicate-tail residue both leave final-turn
   text in the file that the catalogue never reads, so the equality would go red on real data for a
   correct implementation. Ingesting unconditionally makes it exact.

   **But "any session no longer live" is unbounded, and my first attempt to bound it capped the
   wrong unit with wrong numbers.** That attempt said "415 of 421 transcripts / 202 MB are non-live"
   and capped **8 sessions per run** at "8 × 2.4 ms". Both halves were wrong:

   - The population counted every `agent-*.jsonl` as non-live, because an agent file's basename never
     equals a live `sessionId` — **the parent/subagent conflation rule 1 exists to prevent, for the
     third time in this plan.** Attributing files by the vendor's layout, which has **three** tiers and not two —
     `<project>/<sid>.jsonl`, `<project>/<sid>/subagents/agent-*.jsonl`, **and
     `<project>/<sid>/subagents/workflows/<wf_id>/agent-*.jsonl`**: live = **245 files / 217.7 MB**,
     non-live = **177 files / 85.5 MB**, across 94 sessions.

     **The third tier is a fourth instance of the same conflation, and it hid inside a correct
     total.** An earlier draft's rule named only the first two patterns while its *measurement* used
     a `**` glob that swept all three, so the number looked right and the rule an implementer would
     write does not: measured, the two-pattern rule attributes 189 files / 181.8 MB and leaves **56
     files unattributed**, every one of them presently under live session `5627f39a`. When that
     session ends those 56 enter the non-live population, an implementation built to the stated rule
     never enumerates them, their final turns go uncatalogued, and the pending queue reads empty —
     the plausible zero, produced by a rule that disagreed with the glob that validated it. The
     pending-queue enumeration walks all three tiers.
   - A session is not the unit that costs. Offsets are per **file** (line 182), and the largest
     non-live session here holds **25 files**; the 8 largest hold **94 files / 57.0 MB ≈ 226 ms**,
     4.5× the 50 ms budget. Capping sessions bounds nothing.

   So the cap is on **the unit that costs: at most 12 file-tails or 4 MB read per invocation,
   whichever binds first**, remainder deferred. **Derived, on one named model, and marked as the
   estimate it is:** the per-tail figure (2.4 ms, §II:189) is dominated by open/seek/parse rather
   than bytes, so the worst invocation is 12 × 2.4 ms ≈ **29 ms**, plus 4 MB at the 335 MB/s
   partial-parse rate ≈ 12 ms only if the tails are large — the two halves are alternatives, not
   addends, so the bound is ~29 ms against the 50 ms budget, leaving ~21 ms for the promotion writes
   and the queue query. An earlier draft said 16 tails, which is 38.4 ms and leaves 11.6 ms — too
   little for the writes it does not price. **ESTIMATED**: the combined worst case has not been run,
   and §V.2 is the check that runs it. Plus the two bounds that keep the queue finite: only
   sessions the catalogue already holds an offset record for (one it never ingested is `backfill`'s
   job — the pre-hook sessions), and each session closed once and never revisited.

   The steady-state population is sessions that ended since the last `SessionStart`, normally zero or
   one; the cap exists for the day it is not.

   Both closers are
   performed by **`gray-area-capture` at `SessionStart`**: it promotes or deletes any provisional whose session is no longer live, and any older
   than 24 hours regardless. The true property is narrower than the draft claimed and still enough:
   **no hook of the *originating* session has to fire** — the next `SessionStart` on the box
   resolves it, and on a machine running agents that is the next session to start. "Older than the
   window's granularity" was also residue of the day-partitioned store §II rejects; 24 hours is a
   value, not a granularity.

   The third and fourth exist because the first two both fail on the same case: a session whose
   **last** turn is the lagged one and which ends with live background work, where `SessionEnd` was
   measured firing 0 of 2 times. (Liveness is *observed* rather than delivered, which is why it
   resolves without the originating session's cooperation — but a hook must still run the write, as
   stated above.) At close: if the provisional's text is already among the ingested texts for that
   key, delete it; otherwise **promote it to a real `word` row**, because it is a final block that
   never reached the transcript.
5. That comparison is an **exact string equality** against `last_assistant_message`, scoped to one
   turn's blocks. An earlier draft claimed "no content matching anywhere" as a virtue; that claim
   is withdrawn rather than defended — there is no id for an individual text block.

   **THE DECISION MOVES TO FIRE TIME, because at close it is undecidable.** Two earlier drafts tried
   to decide it afterwards and each broke a different case, exhaustively:

   - *multiset multiplicity* needed "the blocks that turn should hold" — a quantity with **no
     source**: no assistant record carries more than one text block (measured: 39,206 carry zero,
     7,004 exactly one, none more), so a turn's blocks arrive one record at a time and its total is
     unknowable until the last one lands.
   - *offset ≥ `offset_at_fire`* fixed `[X, X]` and broke the majority path: on a turn that did not
     lag, the final block is already in the file at `Stop` time and is ingested at a position
     *before* the offset, so it never satisfies and **every non-lagged turn promotes a duplicate**.

   The discriminator is a **future** event — whether another identical block arrives — and no
   recorded position can encode it. But the hook does not need the future: **at the moment `Stop`
   fires it holds the payload's `last_assistant_message` and the transcript as it then stands,
   simultaneously.** So:

   **At fire, compare `last_assistant_message` to the last text block present in the transcript for
   that `(session, agent_id, prompt_id)`.** Equal ⇒ the text has landed and **no provisional is
   written**. Unequal, or absent on the transcript side ⇒ write the provisional, satisfied later by
   any ingested block equal to it for that key.

   **Two inputs can be missing, both measured, both given a rule rather than left to fall through:**

   - **The payload carries no `last_assistant_message` at all — 24 of 585 real rows, 4.1%.** There
     is nothing to compare and nothing to preserve, so **no provisional is written** and the
     omission is recorded as `provisional_skip` with reason `no-last-assistant-message`, the same
     channel as the absent-`prompt_id` case. Note this is **not** "the turn produced none": for 18
     of the 24 the seat's own transcript does contain assistant text. So on those fires closure (a)
     has no source at all, which is stated again at R11 rather than left here.
   - **The transcript side is empty — 95 of 1,745 turns, 5.4%** (75 subagent, 20 main-agent): a turn
     with assistant records and no text block. Nothing equal can ever land, so a provisional written
     here is promoted at close, and **whether that is correct depends on something not yet
     measured**: if the payload's text belongs to this `prompt_id` the promotion recovers a real
     final turn; if it belongs to an earlier one, it misattributes. §V.17 measures exactly that leg.
     Until it has, such a promotion is marked **`attribution='unverified'`** rather than asserted
     correct — the honest state, and visible to a reader.

     **An earlier draft put this at 1.6% (20 of 1,238) and was wrong** for a reason this plan should
     have caught: it keyed turns by `promptId` **alone**, which merges a parent turn with its
     subagents' — the very collision rule 1 exists to prevent. Keyed by `(file, prompt_id)` it is
     5.4%, and three quarters of the population is subagent turns the wrong key had hidden. The id alone suffices; `offset_at_fire` is **not** needed for supersession
   and remains only for incremental ingest (§V.8).

   **The residue, measured rather than argued away.** The rule is wrong exactly when a turn's final
   block **verbatim repeats an earlier block in the same turn** *and* lags: at fire the transcript
   shows `X`, the payload says `X`, so it concludes the text landed and the true tail is lost. Whole
   corpus, 416 files, 1,648 turns carrying text: **2 turns have that shape — 0.12%**; 7 (0.42%)
   contain any duplicate pair at all, and only the lagged subset of the 2 loses anything. The trade
   is 0.12% under-write against the offset rule's ~100% duplication and the multiset rule's
   non-executability, and is chosen on those numbers.

6. **The state this introduces is bound to the schema, and the promotion transition is stated** —
   an earlier draft invented `provisional_skipped` and `attribution` with no table and no column,
   and never said what promotion does to the flag, which left rules 3–5 undecidable:

   - The `word` view's column set — the published contract §V.6 pins — carries **`source`**
     (`'payload'` | `'transcript'`), **`provisional`** (1 = outstanding, 0 = settled) and
     **`attribution`** (`NULL` | `'unverified'` | `'ambiguous'`).
   - **At promotion `provisional` flips 1 → 0 while `source` stays `'payload'`.** That pair is what
     identifies a permanent row with no transcript counterpart, which §V.3's arithmetic needs.
   - **The `word` row's identity is `(session, agent_id, prompt_id, block_seq)`** — `block_seq` the
     block's ordinal within its turn, derived at ingest by counting text blocks already stored for
     that key, which makes it **stable across a `backfill` that starts from zero**: the same
     transcript replayed in the same order yields the same ordinals. It is a position within a turn,
     not a byte offset, so it does not move when a file is re-read. This had to be stated: `(session, seq)` is the `action` tier's
     key and backfill's, and `(session, agent_id, prompt_id)` is **not** unique per word row, since
     66.2% of turns carry more than one block. Without an identity the replacement below has nothing
     to replace on.
   - **A promoted row stays replaceable**, the dedup the flag transition would otherwise lose: if a
     transcript-sourced block equal to it later arrives — a `backfill` re-read, say — it **replaces**
     the row rather than inserting beside it (`source` → `'transcript'`, `attribution` → `NULL`), and
     **a provisional or promoted row carries `block_seq = NULL`** — it has no ordinal until a
     transcript block gives it one — so the replacement cannot match on the 4-tuple and instead
     matches on **`(session, agent_id, prompt_id)` with `source='payload'` and the text equal**, then
     **adopts the transcript block's `block_seq`**. Stating both halves is what gives §V.15's
     promoted-then-replaced assertion a defined pre-state; without them a mismatched ordinal makes
     the next `backfill` insert beside the row instead of finding it. Without both halves the same text double-stores.
   - **`backfill` writes an offset record and leaves `closed_at` NULL** — stating this is what makes
     §V.3 and §V.18 consistent, and an earlier draft left it out so both readings broke something.
     Because it writes the offset, a backfilled session becomes eligible for closure and is drained
     by the queue at the capped rate (177 non-live file-tails ≈ 15 invocations at 12 per run), so
     §V.3's pre-hook exclusion applies only *until* that drain completes and §V.18's month window has
     rows to read. And because `closed_at` stays NULL, backfill never asserts a session is settled
     when it has only been read once — the settling is closure's judgement, made against liveness.
     §II's "steady state is normally zero or one" is therefore true only **after** the initial drain,
     which is stated here rather than implied.
   - **Closure's own state is bound too, rather than invented as the last two were:**
     `session.closed_at` (`NULL` = not closed) is the marker, and **the pending queue is a query, not
     a table** — non-live sessions holding an offset record with `closed_at IS NULL`, oldest first. So
     "deferred, not dropped" is observable without new storage, and a missing marker reads as *not yet
     closed*, the safe direction: it gets revisited. `closed_at` sits on the `session` view, so it is
     on §V.6's contract.
   - **Skips get a carrier**: `provisional_skip(session, agent_id, prompt_id, reason, at)`, `reason`
     from a closed set — `no-prompt-id`, `no-last-assistant-message`. It sits behind a **`skip`
     view**, so the skip population is on the §V.6 contract rather than off it: a turn that was
     skipped is a fact a reader needs as much as a stored one. Prose elsewhere says
     `provisional_skipped`; **the table is `provisional_skip` and the prose follows it.** A promotion
     reached from the **unequal** path carries `attribution = NULL` — nothing about it is ambiguous
     or unverified; it is a final turn the transcript never received. **When both are absent the
     reason is `no-prompt-id`**, checked first, since without a key nothing could be written at all.

7. **A transcript record whose ancestry yields no `promptId` is ingested with a NULL key** — and an
   earlier draft stopped there, which double-stores: a NULL-keyed block can never be counted "for
   that key", so its turn's provisional is always promoted and the final text is stored **twice**.
   So a NULL-keyed block is attributed to the **open provisional for its `(session, agent_id)`**
   when exactly one is outstanding, and is then eligible to satisfy that provisional by the fire-time
   comparison above. **In a `backfill` pass no provisionals exist** — nothing was skipped and nothing
   is outstanding — so this attribution step is a no-op there and every NULL-keyed block is simply
   stored, which is why a from-zero `backfill` is reproducible while a live ingest is order-dependent. With **several** outstanding it stays NULL-keyed, those provisionals are
   promoted, and the rows are marked `attribution='ambiguous'` so the duplicate reads as deliberate.
   With **zero** outstanding there is nothing to promote and the block is simply stored NULL-keyed —
   an earlier draft's "its turn's provisional is promoted" was a no-op in that case.

   **The rate, corrected — an earlier draft claimed "100.000%, zero unresolved" from too narrow a
   sample.** Over the **whole** corpus (414 files, 46,046 assistant records): **3 unresolved,
   99.9935%.** Two are record 0 of a resumed/forked file whose `parentUuid` points outside it; one is
   a subagent record with no `parentUuid` at all. The 40-file scan that produced "zero" contained
   none of them — narrow evidence, universal claim.

**Provisionals are written at `SubagentStop` as well as `Stop`.** The vendor caveat names both —
*"Hooks that need the final assistant text of the current turn should use `last_assistant_message`
on Stop and SubagentStop"* — gray-area already binds `SubagentStop`, and 263 subagent transcripts
exist on this box. A subagent's final text has no other closure: there is no later `SubagentStop`
for that agent, so byte-offset ingest can never recover it. Rule 1's `agent_id` dimension already
keys it correctly. An earlier draft answered a README sentence specifically about the
**`SubagentStop`** payload with a `Stop`-only mechanism.

**(b)** gray-area binds **`SessionEnd`** (new;
prosthetic-conscience binds it, gray-area does not) for one final sweep after the last `Stop` —
**best-effort, and this repo has already measured why.** `plans/hook-surface-spike.md:1066`:
*"`SessionEnd` did not fire in either run that ended with live background work"* — 14 of 15
sessions overall, **0 of 2** with background tasks still live. A session ending with a `Monitor` or
a backgrounded command outstanding is exactly the shape this one has had all day, so (b) cannot be
the primary closure and is not treated as one: **(a) alone is sufficient**, because the provisional
is written at the `Stop` that precedes the end rather than at the end. (b) narrows the residue —
a turn whose `Stop` never fired at all — and an earlier draft presented it as a co-equal closure
without citing the measurement sitting in this repository.

§V.17 measures the lag once a `Stop` hook exists, and decides whether either half is
redundant.

**A retention claim in an earlier draft of this section was WRONG, and it is withdrawn.** It said
transcripts are cleaned up at ~2 weeks, citing "428 rows name a vanished transcript against 373
live". The correct measurement, over all 51 manifests / 1,240 rows:

    resolved rows                                     365
      of which the path is missing NOW                  0     <- nothing has been deleted
    unresolved rows                                   875
      no capture_category (predates schema 2)         447
      event-names-no-seat            (#189 turn ends) 378
      hook-input-carried-no-path                       50
      typed-seat-transcript-missing                     0     <- the alarming population is EMPTY

**Zero transcripts have been deleted since capture.** Every unresolved row was unresolved *at
capture*, and says so in `capture_category` — the field #398 added precisely so these populations
could not be confused. A second draft of this paragraph then got the split wrong in the other
direction, reporting 447 as "the rows that do not resolve" when 447 is only the uncategorized
subset, and double-counting the 378. Written out in full above so the third attempt is checkable.

The corpus is intact back to 2026-08-16 and no `cleanupPeriodDays` is set.

**So the daemon is rejected on the grounds that survive, which are not correctness.** A polling
daemon must be running, must be started, and must be version-matched to the hooks that feed it —
this repository has already been bitten by three capture binaries installed at once whose rows were
indistinguishable (#818), and a resident daemon is that failure with a longer fuse. The `Stop` hook
is event-driven, has no lifecycle, starts nothing on a consumer's machine, and costs 0.57 s per
session. That is sufficient without the urgency argument, and the urgency argument was false.

### Liveness is observed, not inferred

`~/.claude/sessions/<pid>.json` already holds, for every live session on the box:

```json
{"pid":2716,"sessionId":"67082c17-…","cwd":"…/worktrees/bridge-cse_019wLrgs…",
 "startedAt":1788512788587,"procStart":"5677","version":"2.1.260","kind":"interactive",
 "pidDomain":"linux:885fdabadec87be82908a4d724a5fe6e:pid:[4026531836]"}
```

Measured: six entries, six live sockets in `/run/user/<uid>/cc-socks/`. `sessionId` links to the
transcript and `cwd` names the worktree, so *which agents are active* and *who is working on what*
are answerable from this alone.

**`procStart` is what makes it sound.** A pid alone is unsafe — pids are reused, so a recycled one
would report a dead agent as alive, a wrong answer indistinguishable from a right one. pid plus
process start time is a durable process identity. So a session is live when its file exists, the
pid exists, and the start time matches; otherwise **`ended`** — **but only inside the same
`pidDomain`.** The vendor writes that field as `linux:<machine-id>:<pid namespace>` (verified:
`/etc/machine-id` + `readlink /proc/self/ns/pid`), and it is their own guard that a pid is only
meaningful within one namespace on one machine. An earlier draft compared pid and `procStart`
against local `/proc` unconditionally, so a session file written from a container or another host
sharing `$HOME` would have resolved to `ended` — the confident-wrong answer the third value exists
to refuse. The rule: **`pidDomain` must equal the querier's own before pid or `procStart` are
compared; a mismatch is `unknown`, never `ended`.**

**And that comparison is per-OS code this repository does not have** — `grep` finds no
process-liveness anywhere under `plugins/` or `scripts/`, while the release ships six GOOS/GOARCH
pairs. Rather than hand-roll three probes, **liveness is scoped to Linux** and the other platforms
report **`unknown`**, explicitly, as a third value beside `live` and `ended`. On Linux the start
time comes from field 22 of `/proc/<pid>/stat`, which needs no dependency. `unknown` is a real
answer — a board that renders an unmeasurable agent as `ended` would be the confident-wrong output
this plugin exists to refuse — and extending it is a later change with its own files. That is a
scoped-down **complete** concept, not a truncated one: the full concept additionally needs
`liveness_darwin.go` and `liveness_windows.go`, named here so the omission is tracked rather than
discovered.

The hook payload carries **no** pid (verified across 1,228 real rows: `agent_id`, `agent_type`,
`session_id`, `transcript_path`, `agent_transcript_path`, `cwd`, `effort`, `permission_mode`,
`prompt_id`, `stop_hook_active`, `background_tasks`, `session_crons`, `hook_event_name`,
`last_assistant_message` — nothing process-shaped). Identity comes from the payload; liveness comes
from `sessions/`. They are different questions with different sources.

### The three tiers, and why keeping them is free

| tier | contents | measured |
|---|---|---|
| **`action`** | `(session, seq, ts, tool, target, outcome)` | 23,076 calls, ~2.8 MB |
| **`word`** | assistant text and user prompts | 9,837 texts, 9.0 MB |
| **`thought`** | thinking summaries | 248, 0.06 MB *(capture was off until 2026-09-08)* |

**11.8 MB — 4.2% of the 280 MB corpus.** The other 96% is tool *results*: file contents, command
output, base64 images. Bulky, reproducible, and not what you mine for behaviour or intent. So the
discipline is not "index, not copy" — it is **keep the signal, drop the bulk**, and the bulk was
always the expensive part.

**`outcome` is an enum, and the absent case is SUCCESS — which an earlier draft got backwards.**
Measured across 20,947 invocations paired by `tool_use_id`, four ways rather than three:

| | share | |
|---|---|---|
| result present, `is_error=false` | 79.7% | `ok` |
| result present, **`is_error` absent entirely** | **18.4%** | `ok` |
| result present, `is_error=true` | 1.9% | `error` |
| **no `tool_result` block at all** | 0.1% (12 calls) | `unresolved` |

The convention is **per tool**, and that is the whole finding (re-measured 2026-09-08; the corpus
is live, so absolute counts drift and only the RATES are stable):

    tool          calls   false   true   absent   no result
    Bash          17,318 16,969    338        0          11
    Read           1,342      0     23    1,319           0
    Edit           1,325      0     21    1,304           0
    SendMessage      107      0      0      107           0

**Bash never omits the flag: 0 absent, on any call that produced a result.** An earlier draft wrote
"17,027 of 17,027", which did not sum — 16,682 + 335 left 10 calls unaccounted, and those 10 were
Bash calls with no `tool_result` at all. The denominator matters and is now named: of Bash calls
that **produced a result**, the explicit-flag rate is **100.00%**; over *all* Bash calls it is
99.94%, and the difference is the `unresolved` population, not a flag that went missing.

So absence of `is_error` is the *normal success shape* for every tool except Bash. A rule mapping
absent → `unresolved` — which a previous draft of §V asserted — would have mislabelled **3,844
successful calls**. `unresolved` means exactly one thing: no `tool_result` was ever recorded,
because the call was killed or interrupted mid-flight. A boolean would force those 12 into `ok` or
`error`, and both are lies.

**And that is why R2's detector is the Bash invariant, not the absent case.** `is_error` is a
vendor field; if it were renamed, every Bash result would read as absent-therefore-`ok`, the error
rate would fall silently to zero, and a check asserting "absent ⇒ unresolved" would pass
identically before and after — firing on the healthy path and silent on the failure, the exact
[[facts-are-fields]] clause-3 defect this plan cites elsewhere. What *does* move is the shape:
**of Bash calls that produced a `tool_result`, 100.00% carry an explicit `is_error`.** So the check
asserts that rate — over that denominator, not over all calls — and that Bash's error count over
the window is non-zero. A rename breaks both: every Bash result would become `absent`, dropping the
rate to 0% and the error count to zero.

**Reasoning is mined, not withheld.** Recovering intent is an explicit goal of the plugin. The
guard against summaries becoming evidence lives in **FEOV's citation ledger**, not here (gblock,
2026-09-08; landed as #844) — an anchor requires a `finding_id` and a location, and a summary has
neither, so it cannot form one. That is a property of the schema rather than a rule anyone must
remember.

### Store, retention, cleanup

    ~/.local/state/special-circumstances/catalogue/catalogue.db

**ONE DATABASE, NOT A PARTITION PER DAY — and this overrides an earlier instruction, so it is
flagged rather than swapped quietly.** gblock specified "a db with a timestamp for a name and a
ttl, and the hook as the cleanup". That predates the raw-SQL ruling, and the two are in direct
tension: **SQLite cannot join across separate connections**, and `ATTACH` is capped at **10 per
connection** (measured: refused at the 11th). So a day-partitioned store cannot serve arbitrary
user SQL over a month — a query would silently see one partition, or eleven days, and Goal 1 is
the whole point of the surface. Partitioning loses.

The cost of giving up `unlink` is measured and small. On a 1M-row table:

| | |
|---|---|
| `DELETE` one day (33,333 rows), indexed on `day` | **325 ms** |
| indexed query afterwards | **0.18 ms** |
| ingest a fresh day back | 211 ms |
| file size across a full delete+insert cycle | **69.2 MB → 69.2 MB** |

**No `VACUUM` is needed**: SQLite reuses freed pages, so at steady state the file plateaus rather
than growing.

**The sweep needs its own budget, because 325 ms does not fit in 50 ms.** An earlier draft claimed
both in one sentence, which is a contradiction its own measurement refutes. The sweep is a no-op on
every `SessionStart` but the first of a new UTC day; on that one run it costs 325 ms. So the budget
is stated as two numbers — **50 ms steady-state, 500 ms on the day-rollover run** — and §V.2
asserts both, with the rollover case forced rather than waited for.

Under `~/.local/state/`, **not** `~/.claude/`: writing our store into the vendor's directory is
squatting in a namespace we do not own.

**Archiving is not observable and nothing keys off it.** `ArchiveSession` is an HTTP POST to a
remote API (`[bridge:api] POST … 409 (already archived)`) operating on cloud sessions; it writes no
local state. Whether mined signal should outlive the window at all is **#845**, deferred.

### The SQL surface

Raw read-only SQL, plus verbs (`agents`, `session`, `touched`, `find`, `backfill`) that are ordinary queries
with names.

**Read-only is enforced and measured.** On `mode=ro&_pragma=query_only(1)`: `INSERT`, `UPDATE`,
`DELETE`, `DROP`, `CREATE` all refused. A runaway `act a, act b, act c` cross join was cancelled by
context deadline at **exactly 300 ms**, so a bad query cannot wedge the box.

**Two holes, and the gate that closes them is named rather than gestured at.** `ATTACH` and
`PRAGMA writable_schema` are both permitted on a read-only connection (measured). On this box
neither is an escalation — every caller already has `Read` and `Bash`, so attaching a file it could
simply open grants nothing — but it becomes one the moment the surface is less trusted than the
caller.

**Both holes are closed, with two calls.** An earlier draft of this plan asserted that neither a
statement classifier nor an authorizer callback existed in `modernc.org/sqlite`, and shipped the
holes on that basis. **That was an unmeasured negative claim and it was wrong** — the plan-auditor
caught it, and re-measuring confirms (on **v1.57.0**; note this is *feov's* pin and does not constrain this module — `go mod tidy` here resolves **v1.58.0**, on which all four mitigations were re-checked and hold):

| mitigation | measured |
|---|---|
| `_defensive=1` in the DSN | `PRAGMA writable_schema=1` becomes a **silent no-op** — reads back `0` with it, `1` without, on the same connection |
| `sqlite.Limit(conn, sqlite3.SQLITE_LIMIT_ATTACHED, 0)` (exported at `sqlite.go:1302`, constant at `lib/sqlite.go:3876`) | `ATTACH` fails: **`too many attached databases - max 0`**. Previous value was 10 |

**But `Limit` binds a CONNECTION, not the pool — and a draft of this section got that wrong too.**
`func Limit(c *sql.Conn, …)` takes a single connection. Measured on v1.57.0: after
`Limit(c1, ATTACHED, 0)`, `c1` refuses `ATTACH`, while **a second connection from the same
`*sql.DB` reads the limit back as 10 and `ATTACH` succeeds** — as does `db.Exec("ATTACH …")`
straight through the pool. `_defensive` is different: it is applied at connection open, so it holds
on every pooled connection (measured: `writable_schema` reads 0 on both).

So the query path **must pin its own connection**: acquire a `*sql.Conn`, apply
`Limit(ATTACHED, 0)` to it, run the user's statement on that connection, release it. Never
`db.Query` through the pool, because the pool may hand out a connection the limit was never
applied to. §II's own measurement row — *"10 per connection … not per process"* — was the evidence
for this all along, and an earlier draft printed it without drawing the conclusion.

So the surface is: `mode=ro` + `query_only(1)` + `_defensive=1` at the DSN, plus a **pinned
connection** carrying `Limit(ATTACHED, 0)` per query. An ordinary `SELECT` works under all four.

Two honest qualifications:

- **`RegisterPreUpdateHook` is NOT a veto** and is not proposed as defence in depth. Its callback is
  `func(SQLitePreUpdateData)` with no error return (`pre_update_hook.go:35`), so it can *observe* a
  write and cannot *refuse* one. Writes are refused by `mode=ro`, which is measured; the hook would
  only tell us afterwards.
- A **text pre-filter** over the SQL — "must start with SELECT" — remains deliberately **not**
  proposed. It is a pattern standing in for a schema, defeated by a leading comment or an
  unanticipated form, and now unnecessary: the two mechanisms above are SQLite's own.

**The schema is a contract, so callers query VIEWS.** If agents write their own SQL, a renamed
column breaks work that is not in this repo and cannot be swept. #819 is the warning — 0 of 7 run
archives are readable by any current binary, because words were retired and an epoch moved. So:
stable views named `action`, `word`, `thought`, `session` and `skip`; base tables free to evolve beneath
them; views evolve **additively only**; and `capture_build` on every row (#818's pattern) so a
reader can tell which binary wrote it.

### Full text stays unindexed

Ripgrep searches the whole corpus in **33–54 ms**, measured through a real binary. An FTS5 index is
**estimated** at 1–2 GB — not measured — and would need invalidating on every append and rebuilding
whenever the client deletes a session, to beat 40 ms. `find` shells to `rg` over the paths the
catalogue names and joins each hit back to who/when/where, passing the search term as **a single
argv element, never through a shell**.

Note for anyone re-measuring: inside Claude Code `rg` is a shell **function** the client injects,
not a binary, so a hand-run measurement is taken through a path no Go program can use. ripgrep is
declared in `requirements.json` as `recommended` (#843) and is **promoted to `required` in the
change that ships `find`**, because from that moment a missing binary returns zero hits that read
exactly like an honest zero.

---

## III. Scope — carriers and consumers, censuses run and adjudicated

**No new command directories.** Ingest is a new `Stop` event on the existing `gray-area-capture`;
queries are new verbs on the existing `gray-area`. So `requirements.json`'s `_hook_binaries` and
`docs/setup-script.md:148`'s hook-binary count (18) are **untouched**, and `scripts/pluginparity`
stays green. This was checked because two new directories would have turned it red.

- [NEW] `plugins/gray-area/tools/internal/catalogue/` — schema, views, incremental projection,
  query path (pinned connection + `Limit`).
- [NEW] `internal/catalogue/liveness_linux.go` and `liveness_other.go` (build-tagged) — the
  `/proc/<pid>/stat` start-time comparison, and the `unknown` answer everywhere else. Named
  because §II scopes liveness to Linux and an unnamed omission is one nobody tracks.
- [NEW] `internal/catalogue/testdata/corpus/` — **synthesized** transcripts (see §V).
- [NEW] `internal/catalogue/backfill.go` — **the component an earlier draft omitted entirely.** A
  `Stop`-hook-only store contains nothing until it has run for a month, so §V's checks over a
  month window would have had nothing to read. Backfill projects the existing corpus once, is
  triggered by the `gray-area backfill` verb — **the fifth new verb, counted in §III's dispatch-table and README carriers** — explicit, never from a hook, and is **exempt from goal 2's
  budgets because it is not on any agent's turn** — basis: a scan-and-classify pass over the 280 MB corpus
  measured **545 MB/s** (`python3` line-scan over `~/.claude/projects/**/*.jsonl`, 2026-09-08) — an
  **upper bound on throughput, not a measurement of this component**, which does not exist yet. A
  cold seed should be around a second; the number is marked as the estimate it is. It is idempotent by
  `(session, seq)` so a re-run cannot double-count.
- [MODIFY] `cmd/gray-area-capture/main.go` — handles `Stop` (ingest) and sweeps rows past the window
  at `SessionStart`. **`Stop` is an explicit new branch, NOT a fall-through.** Today
  `main.go:522–537` routes every non-`SessionStart` event to `appendRow(buildRow(...))`, i.e. a
  seat/turn-end row; a `Stop` binding that merely adds ingest without leaving that path would write
  ~131 `event-names-no-seat` rows per session into the manifest. So `Stop` **and `SessionEnd`** are each explicit branches that
  ingest and write **zero manifest rows** — an earlier draft gave `Stop` the branch and left
  `SessionEnd` to the same fall-through it had just named. The existing `SessionStart` and
  `SubagentStop` manifest writes are unchanged. The `-event` flag help at `main.go:488`
  (`SubagentStop | SessionStart`) gains both events.
- [MODIFY] `cmd/gray-area/main.go`, `inspect.go` — five verbs (`agents`, `session`, `touched`, `find`, `backfill`) plus `sql`. The existing six are
  untouched.
- [MODIFY] `plugins/gray-area/hooks/hooks.json` — two new bindings, `Stop` and `SessionEnd`.
- [MODIFY] `plugins/gray-area/requirements.json` — **three** changes, the third a carrier an earlier
  draft missed: `:8`'s `_tier_comment` reads *"RECOMMENDED, AND DELIBERATELY AHEAD OF ITS CONSUMER …
  PROMOTE TO `required` in the same change that ships the verb"*. Flipping `tier` without rewriting
  that leaves a carrier instructing the change this plan is making. The other two: The `gray-area-capture` entry's
  `event` field says `SubagentStop` and already omits `SessionStart`; it becomes accurate and
  gains `Stop` and `SessionEnd`. **And
  ripgrep is promoted `recommended` → `required`**, because §II commits to promoting it "in the
  change that ships `find`" and this is that change. On `main` today it is `recommended` (#843,
  merged); an earlier version of this branch carried `required` already, from a superseded commit
  that has been dropped — so the delta is real and is this change's to make. The consumer cost is
  graded as R10.
- [MODIFY] `plugins/gray-area/README.md` — the miner's new surface, the store, the month window.
- [MODIFY] **`README.md`** (repo root) — its verb table at `README.md:116`, **and `README.md:120`**,
  which says *"the manifest is an index of where trajectories are, never a copy of their
  contents."* That sentence is the plugin's consumer-facing surveillance disclosure and it becomes
  false the moment the `word` and `thought` tiers exist. It is rewritten to the new scoping
  statement: the manifest stays an index; the catalogue at
  `~/.local/state/special-circumstances/catalogue/catalogue.db` copies **tool names and targets,
  assistant text, user prompts and thinking summaries** — never tool results — for **one month**,
  on this machine only, never in the repository.
- [MODIFY] `plugins/gray-area/README.md:37` — *"`SubagentStop` also carries
  `last_assistant_message`, and that is deliberately not recorded… duplicating it into a second
  file buys no evidentiary value."* The direct negation of the design above. Rewritten: the
  **manifest row** still refuses it, for the reason stated; the **catalogue** ingests it at `Stop`
  as a provisional `word` row, because it is the only source of a final turn the transcript has not
  yet written.
- [MODIFY] `plugins/gray-area/README.md:124` — *"the manifest is gitignored; nothing leaves the
  box."* Gitignore is no longer the privacy mechanism once the store lives under
  `~/.local/state/`, outside any repository. Rewritten to the same scoping statement: nothing
  leaves the box because the store is outside the tree, not because of a gitignore line.
- [MODIFY] `plugins/gray-area/requirements.json:34` — `_hook_binaries.gray-area-capture.purpose`
  says *"an index, never a copy of conversation content."* Rewritten: the manifest row is an
  index; the same binary now also writes the catalogue, which copies the three tiers named above.
- `plugin.json` / `marketplace.json` descriptions ("Captures each seat's own transcript by path
  at SubagentStop") are **partial, not false** — unchanged, noted so the omission is deliberate.
- [MODIFY] `plugins/gray-area/tools/go.mod` + [NEW] `go.sum` — **not "one dependency".** Measured
  in a scratch module: **pinned at `modernc.org/sqlite v1.57.0`** to match the version §II's mitigations were measured on, rather than whatever `tidy` resolves (currently v1.58.0). The `go` directive is set to a **current patch release** (`go 1.25.13` today) rather than the `go 1.25.0` that `tidy` writes — see the qlty entry below for the measured reason. `go mod tidy` adds **1 direct + 9 indirect** requires (`modernc.org/libc`,
  `mathutil`, `memory`, `golang.org/x/sys`, `go-humanize`, `uuid`, `go-isatty`, `go-strftime`,
  `bigfft`), 25 modules in `go.sum`, **and rewrites the directive `go 1.25` → `go 1.25.0`.**
  **cgo rejected**: the release job cross-compiles six platforms and a cgo SQLite needs a C
  toolchain per target — a cost this repository already pays once for tesseract and should not pay
  twice for a store whose switchboard queries measure 1.00 ms and 2.38 ms at 1M rows. All six pairs
  cross-compile pure-Go today.
- **NO qlty suppression is needed — measured, and this reverses both an earlier draft and a ruling
  made on its framing.** The draft said `go 1.25` → 0 findings and `go 1.25.0` → 39, quoting
  `.qlty/qlty.toml:88-99`, and gblock ruled "add a real suppression" on that basis. Running the
  scanner instead of quoting the comment:

      go 1.25.0    ->  46 findings      (the comment says 39; CVEs have accumulated since)
      go 1.25.13   ->   0 findings
      go 1.25      ->   0 findings, but `go build` and `go test` REFUSE once a dependency
                        exists: "updates to go.mod needed; to update it: go mod tidy"

  So the fix is **not** a suppression, which would permanently hide real findings on that file. It
  is to set the directive to a **current patch release**, which is exactly what
  `frank-exchange-of-views/tools/go.mod` already does (`go 1.25.13`, scanning clean today). This
  module does the same. The cost is that the directive needs bumping as CVEs accumulate against
  whatever patch it names — a maintenance task, not a hidden finding.

  **`.qlty/qlty.toml:101` is a [MODIFY] CARRIER for this change**, which an earlier draft missed by
  treating that file as already-corrected. Its roster line reads *"a module with no dependencies
  keeps the short form (gray-area, prosthetic-conscience); a module WITH dependencies pins a current
  patch release (frank-exchange-of-views, scripts)"* — and this change moves gray-area across that
  line, making the sentence false about the post-change tree. `f6795bac` (the correction) is already
  an ancestor of this branch, so moving gray-area into the pinned group is *this* change's job.
  Separately, **`.qlty/qlty.toml:88-99` was STALE** — it said feov carried 39 accepted findings when
  feov reports clean — and **was corrected in #850, already merged and an ancestor of this branch** —
  so the in-tree comment is right about today and the roster line above is what this change still
  owes. An earlier draft described the correction as pending. All four `go.mod`
  files were scanned before rewriting it: 0 findings each.
- [MODIFY] `.github/workflows/hooks.yml` — `cache-dependency-path` names
  `plugins/gray-area/tools/go.mod` at **:95, :828, :1469**; with a `go.sum` present the key must
  track it, as feov's already does. `:673` (not `:663`) records that Windows `t.TempDir` SQLite tests cost ~73 s for
  `recordsql` — but that block already installs the ImDisk RAM disk which takes it to ~5 s, so the
  cost this plan imports into the gray-area leg is seconds, not a minute. Corrected rather than
  left overstated.

**Census 1 — manifest-format readers.** `grep -rlnE 'trajectories-|SessionRow|Seats\('` with **no
`--include` filter** — an earlier run of this census used one and returned 18, missing two `.mjs`
files. **21 entries, every one adjudicated:**

| # | file | disposition |
|---|---|---|
| 1 | `plans/gray-area-catalogue.md` | this plan |
| 2–3 | `plans/gray-area.md`, `plans/rearm-coverage-experiment.md` | design documents, not consumers |
| 4 | `feov/tests/simulator/debate.test.mjs` | **string sharer** — `citationSeats(`. No change |
| 5 | `feov/tests/simulator/prompts.test.mjs` | **string sharer** — `captureSeats(`. No change |
| 6–11 | `feov/tools/internal/{cli/diagnostics.go, record/planguard/planguard_test.go, record/queries.go, record/writelookups_parity_test.go, seatclass/seatclass.go, seatclass/seatclass_test.go}` | **string sharers** — `KnownSeats()` / `RegisteredSeats(`, FEOV's *debate* seats, a different concept sharing a token ([[facts-are-fields]] clause 4). No change |
| 12 | `plugins/gray-area/README.md` | **carrier** — [MODIFY], in scope |
| 13–14 | `gray-area/tools/cmd/gray-area-capture/{main.go,main_test.go}` | **carriers** — [MODIFY] |
| 15–16 | `gray-area/tools/cmd/gray-area/{inspect.go,main.go}` | **carriers** — [MODIFY] |
| 17–20 | `gray-area/tools/internal/claims/{coverage,manifest}{,_test}.go` | **carriers**, read-side. Unchanged by this plan: the catalogue is built *from* them |
| 21 | `prosthetic-conscience/.../checkpointseal/main_test.go` | **string sharer** — `ConcurrentSeats(`. No change |

**Census 2 — verb-surface consumers.**
`grep -rln -E 'gray-area (tools|checkpoint|rework|stalls|pr|coverage)' .` — **17 entries**, pasted
in full (an earlier draft claimed 18 and enumerated 17, and gave no command at all):

    1  README.md                                                   CHANGES - repo verb table
    2  plans/gray-area-phase-2-build.md                            unchanged - design doc
    3  plans/gray-area-phase-2.md                                  unchanged - design doc
    4  plans/gray-area.md                                          unchanged - design doc
    5  plans/rearm-coverage-experiment.md                          unchanged - design doc
    6  plugins/gray-area/README.md                                 CHANGES - plugin verb table
    7  plugins/gray-area/commands/audit-checkpoint.md              unchanged - names `checkpoint`
    8  plugins/gray-area/commands/audit-pr-body.md                 unchanged - names `pr`
    9  plugins/gray-area/commands/audit-repetition.md              unchanged - names `rework`
    10 plugins/gray-area/commands/audit-seat-coverage.md           unchanged - names `coverage`
    11 plugins/gray-area/tools/cmd/gray-area/main.go               CHANGES - the dispatch table
    12 .../internal/claims/conformance_test.go                     unchanged - names `checkpoint`
    13 .../internal/claims/manifest.go                             unchanged - error text
    14 .../internal/claims/recheck_test.go                         unchanged
    15 .../internal/prbody/prbody.go                               unchanged - names `pr`
    16 .../internal/prbody/prbody_test.go                          unchanged
    17 prosthetic-conscience/.../checkpoint/loopproblems_test.go   unchanged - golden text

**Three change; fourteen do not.** The six existing verbs keep their names and behaviour, so every
document that names only those stays correct. `cmd/gray-area/main.go` changes because it *is* the
dispatch table (the `case` at `main.go:180`).

**`cmd/gray-area/inspect.go` is [MODIFY] in the change list and correctly absent from this census**
— it implements verb bodies but contains no `gray-area <verb>` string, so the census's pattern does
not and should not match it. Noting the reconciliation because a file that is changing while
missing from a census is exactly the shape that should be questioned.

**Scoped out, per [[complete-the-concept]]:** an **MCP surface**. No plugin in this repository
declares an MCP server and no plugin-level `.mcp.json` exists, so the declaration mechanism is
unverified; naming a file that does not exist is what the first audit caught. The full concept
would additionally touch that mechanism, `plugin.json`, and a per-session stdio shim. The CLI plus
raw SQL is a complete concept without it.

**Not in the v2.0.0 boundary**: additive, no shipped schema changes, manifest keeps schema 5.

---

## IV. Risks, graded — each with the §V check that would notice it

| # | risk | grade | check |
|---|---|---|---|
| R1 | **Projection drifts from the vendor's transcript format.** Fails toward under-counting. | likely eventually × medium × contained | §V.4 — a corrupt line yields `unparsed` with a count, never a session with zero acts |
| R2 | **`is_error` is renamed and every call reads `ok`.** The failure rate silently becomes zero. The naive check — "absent field ⇒ `unresolved`" — CANNOT notice this: absence is already the normal success shape for every tool but Bash, so it passes identically before and after. | medium × high × cheap | §V.5 — asserts the invariant that actually moves: **of Bash calls that produced a `tool_result`, 100.00% carry an explicit `is_error`** (0 absent of 17,307 measured), and its error count is non-zero. A rename drives the rate to 0% and the count to zero |
| R3 | **A renamed column breaks callers' SQL** outside this repo. | certain over time × high × cheap | §V.6 — a test asserts the view columns, so renaming one fails here rather than in someone's query |
| R4 | **The month TTL deletes acts that could not be recovered.** An earlier draft graded this *certain*, on a retention claim now measured false: **nothing has been deleted**, so a swept row remains reprojectable from the transcript for as long as the client keeps it. The risk is real but **conditional** — it bites only if and when the client begins expiring transcripts, which it does not do today and which no setting on this box enables. | conditional × medium × **accepted** | **No check — accepted.** A check that the sweep works is not a check that the loss is acceptable, and the loss does not currently occur. #845 owns whether mined signal should outlive the window; the trigger to revisit is the client gaining a retention policy |
| R5 | **Byte offsets desynchronise** (a transcript truncated rather than appended). | low × medium × cheap | §V.8 — offset record stores size and the sha of the first 4 KiB; a mismatch reprojects from zero |
| R6 | **A malicious or clumsy query wedges the box.** | medium × medium × closed | §V.9 — writes refused; cross join cancelled at a deadline; `ATTACH` and `writable_schema` refused on the pinned connection |
| R7 | **Liveness reports a dead agent as live** via pid reuse. | low × high × closed | §V.10 — a stale `procStart` must resolve to `ended`, asserted with a fabricated mismatch |
| R8 | **First dependency in a zero-dependency plugin**, and its consumer-side install cost — `scripts/bootstrap-plugins.sh:83-99` builds hook binaries **on the consumer's machine**, and its own comment names dependency download as the part most likely to overrun the ~5 minute budget. Measured cold against an empty module cache: **3.7 s download + 23.0 s build, 348 MB cache.** Inside the budget, stated rather than assumed. | certain × low × **low to mitigate** — one-time and measured; nothing to do unless that budget tightens | named as a decision at §III's `go.mod` bullet (pin, cgo rejection, 1+9 requires), not absorbed |
| R11 | **The transcript lags the `Stop` hook and a session's final turn is never ingested.** Documented by the vendor, **not measured on this box** — it cannot be, until a `Stop` hook exists here; the one inference attempted was confounded and is recorded in §II so it is not repeated. | likely × medium × **one closure plus a best-effort narrowing, with two measured residues** — (a) covers the lagged final turn except: the 0.12% duplicate-tail under-write (rule 5), and the **4.1% of fires carrying no `last_assistant_message`, where (a) has no source at all**. "Sufficient alone" is claimed against those two rates, not absolutely; (b) `SessionEnd` did not fire in 2 of 2 runs ending with live background work | §V.3's `catalogue = transcript + promoted` arithmetic, whose second addend IS the R11 population, plus §V.17 at fire time. An earlier draft named §V.3 while §V.3 excluded exactly those rows |
| R10 | **Promoting ripgrep to `required` BLOCKS consumers** who lack the binary. The mechanism is `toolchain.MergeStrictest` + `doctor.verdict` — a missing `required` tool returns BLOCKED — **not** a manifest key: `_tier_comment` is documentation on the entry, and an earlier draft cited it as though it were the enforcement. Deliberate: from the moment `find` ships, a missing `rg` returns zero hits indistinguishable from an honest zero, and a wrong answer is worse than a blocked install. | certain × medium × accepted | §V.14 — asserts the verb and the tier move in the same change, so neither can ship alone |
| R9 | **The board reads as complete when reasoning capture was off.** | medium × medium × mitigated | §V.11 — a session with zero non-empty thoughts is distinguishable from one never asked |

---

## V. Verification plan

**FIXTURE-BACKED VS LOCAL-ONLY, because CI has no corpus.** `hooks.yml:652` runs the gray-area job
on ubuntu and windows with `go test ./...`; no runner has a `~/.claude` transcript tree. A check
written against "the live corpus" therefore passes for absence of data, or skips and reports green
— the plausible zero this plan cites [[facts-are-fields]] clause 3 against, reproduced inside its
own verification plan.

So every check below is one of two kinds, and says which:

- **[FIXTURE]** runs everywhere, against `internal/catalogue/testdata/corpus/` — a tree of
  **SYNTHESIZED** transcripts written to the shapes measured in §II: one session with a Bash error,
  one with an absent `is_error`, one tool call with no `tool_result`, one corrupt line, one
  truncated file. **Not captured transcripts.** This repository is PUBLIC, and real transcripts
  carry user prompts, assistant text and `cwd` worktree paths — committing them is exactly the
  disclosure #845 was filed about, which an earlier draft of this very section walked into one page
  after citing it. Synthesis is not a compromise here: every shape the checks need is named in §II
  and can be written by hand, and a fixture that reproduces a measured shape is a better test than
  one that happens to contain it.
- **[LOCAL]** needs the real corpus and is **skipped in CI by an explicit guard that prints why**,
  never by silence. A `[LOCAL]` check that finds no corpus reports `NOT MEASURED — no transcript
  tree at ~/.claude/projects` and fails if `GRAY_AREA_REQUIRE_CORPUS=1`, which the pre-merge run
  sets. Absence of data is reported as absence, not as a pass.

Written before implementation. **Re-arms on:** any change under `internal/catalogue/`,
`cmd/gray-area-capture/`, or the view definitions.

1. **[FIXTURE]** `cd plugins/gray-area/tools && go build ./... && go vet ./... && GOOS=windows go vet ./... &&
   go test -count=1 ./... && go test -race ./...` — `-race` because both
   `scripts/check/gates.go:152` and CI run it over this module, and this change is the first to
   give it a connection pool. The Windows vet is in the loop because a hand-rolled path match cost a
   red CI leg on #839.
2. **[LOCAL]** — retagged: a wall-clock budget measured on one Linux box cannot be asserted on
   windows-latest or under `-race`, both of which CI runs (`hooks.yml:837-838`). It is guarded to
   linux, non-race, and reports `NOT MEASURED` elsewhere rather than passing quietly.
   `go test -run TestIngestBudget ./internal/catalogue/` — **execs the BUILT
   `gray-area-capture` binary** rather than benchmarking library work in-process, because the
   budget is stated as spawn floor plus ingest and an in-process benchmark cannot measure a spawn.
   Four cases are forced rather than waited for: the day-rollover (seed a row dated outside the
   window), a **closure with an outstanding provisional and a nonempty unread tail** (the new work
   goal 2 previously had no case that could see — assert against the 50 ms steady-state budget), a
   **backlog exceeding both halves of the cap** — ≥40 non-live file-tails totalling ≥10 MB unread —
   asserting exactly **≤12 tails and ≤4 MB read per invocation**, the invocation inside 50 ms, and the
   remainder still queued rather than dropped; and **closure coinciding with the day-rollover sweep**
   (assert against 500 ms). An earlier draft asserted "the 8-per-run cap" over "20 non-live sessions"
   — the cap this section repudiated, in the unit it says bounds nothing, and a case an
   implementation with no cap at all would pass. —
   **goal 2, measured**: the `Stop` path stays under **20 ms per turn** (4.33 ms spawn floor +
   ingest); `SessionStart` under **50 ms** steady-state and under **500 ms** on the forced
   day-rollover run that performs the 325 ms sweep; **`SessionEnd` under 500 ms**, a third of the
   1.5 s the vendor shares across every `SessionEnd` hook. Four numbers, because each is a
   different run and an earlier draft asserted a single budget its own measurement refuted.
3. **[LOCAL]** `go test -run TestProjectionMatchesIndependentCount ./internal/catalogue/` — reads
   **the hook-populated catalogue** (not a fresh projection) and compares against a transcript-derived
   count taken by a different code path: session counts, per-session **act** counts, and per-session
   **`word`** counts. The arithmetic is **stated, not gestured at** — an earlier
   draft said promoted rows have no transcript counterpart "by construction", which §II refutes:
   byte-offset ingest *does* recover a lagged turn on the next `Stop`, so only a **final** lagged
   turn's text never reaches the file. So:

       FOR sessions closure has actually closed (session.closed_at IS NOT NULL):
         catalogue words  ==  transcript words  +  count(source='payload' AND provisional=0)

   **The scope is not the corpus; an unscoped `==` would be guaranteed red** for a correct
   implementation, three ways: live sessions have bytes past their last `Stop` that closure by
   definition has not read (the majority of the corpus here — 244 files / 217.3 MB); sessions
   predating the hook have transcript words and zero catalogue words by design, since closure only
   touches sessions it holds an offset for; and the deferred remainder has unread tail bytes at
   measurement time on purpose. So the test enumerates `session.closed_at IS NOT NULL` and compares
   over that set alone. Round nineteen fixed this equality on one side; leaving it unscoped reopened
   it on the other.

   Each addend is asserted rather than only the total: every `source='payload' AND
   provisional=0` row has its text asserted **absent** from the transcript for its key — that is
   what makes it a genuine residue rather than a duplicate — and `provisional=1` rows are counted
   and reported as **outstanding** rather than silently dropped. An earlier draft excluded them and
   called that "asserted separately" while naming no assertion, which is not executable, and could
   not see R11 either way since the excluded rows are exactly where R11's evidence lives. Words are compared because a final turn lost to transcript lag
   (R11) changes the word count and leaves the act count untouched.
4. **[FIXTURE]** `go test -run TestCorruptLineIsUnparsedNotZero ./internal/catalogue/`
5. **[FIXTURE]** for the model, **[LOCAL]** for the ratio: `go test -run 'TestAbsentIsErrorIsSuccess|TestBashAlwaysStatesTheFlag' ./internal/catalogue/` —
   the first pins the corrected model (absent ⇒ `ok`, only a missing `tool_result` ⇒ `unresolved`);
   the second is R2's real detector: Bash's explicit-flag rate is 100.00% **over calls that produced
   a result** (the denominator an earlier draft left unnamed, making the assertion fail on the live
   corpus at 99.94%), and its error count is non-zero. An earlier draft asserted the opposite of the first and would have mislabelled
   3,844 successful calls.
6. **[FIXTURE]** `go test -run TestViewColumnsAreTheContract ./internal/catalogue/` — asserts the exact column
   set of each view.
7. **[FIXTURE]** `go test -run TestSweepDeletesExactlyExpiredRows ./internal/catalogue/`
8. **[FIXTURE]** `go test -run TestTruncatedTranscriptReprojectsFromZero ./internal/catalogue/`
9. **[FIXTURE]** `go test -run 'TestWritesAreRefused|TestRunawayQueryIsCancelled|TestAttachRefusedOnQueryPath|TestWritableSchemaIsNoOp' ./internal/catalogue/` —
   asserts each mitigation §II claims, **through the connection the `sql` verb actually uses**, not
   a hand-held one: writes refused; a cross join cancelled by deadline; `ATTACH` refused; and
   `PRAGMA writable_schema=1` reading back 0. The `ATTACH` case is the one that would regress
   silently — a pooled `db.Query` instead of a pinned `*sql.Conn` reopens the hole and every other
   test still passes, so this check exercises the verb's own path deliberately. **And it must hold
   a SECOND live connection while it does so** (`SetMaxOpenConns(>1)`, the query issued while
   another conn is checked out): measured, an implementation that pins, sets the limit, releases,
   then queries through the pool is refused single-goroutine — because the pool hands back that
   same connection — and ships the hole under concurrency. A check that passes for that
   implementation is not a check.
10. **[FIXTURE]** (fabricated `sessions/` dir, not the real one) `go test -run 'TestStaleProcStartIsEnded|TestForeignPidDomainIsUnknown' ./internal/catalogue/` —
    the second case writes a session file whose `pidDomain` differs from the querier's and asserts
    the answer is `unknown`, not `ended`, even when a local pid happens to match.
11. **[FIXTURE]** `go test -run TestCaptureOffSessionIsDistinguishable ./internal/catalogue/`
12. **[PROCESS]** Revert-check every new test** against the mutation it exists to catch — one at a time, from a
    **file backup, never `git checkout`**: on #839 that silently failed to restore untracked files,
    four mutations accumulated, and the "KILLED" lines were damage piling up rather than tests
    working. The tell was three mutations reporting the same two failures.
13. **[FIXTURE]** Wiring pinned apart from mechanism — **and the fall-through pinned shut**:
    `go test -run 'TestStopWritesNoManifestRow|TestSessionEndWritesNoManifestRow' ./cmd/gray-area-capture/`
    asserts `-event Stop` **and** `-event SessionEnd` each leave the manifest at zero rows, because a wiring test that only proves the ingest CALL is
    present cannot see a manifest write that should not fire.** The `Stop` ingest call and the sweep call each get a
    test that fails when the CALL is removed, not only when the function breaks — #839's
    `boxWarnings` seam is the pattern.
14. **[FIXTURE]** `go test -run TestFindShipsWithRequiredRipgrep ./internal/catalogue/` — asserts
    the `find` verb and `requirements.json`'s `required` tier move together (R10), so neither can
    ship without the other.
15. **[FIXTURE]** `go test -run 'TestBackfillIsIdempotent|TestBackfillVerbDispatches|TestProvisionalWordIsSupersededOnce' ./internal/catalogue/` —
    the third is a **multi-turn, multi-block** fixture, because a one-block-per-turn fixture cannot
    exercise the rule at all — 66.2% of real turns carry more than one block. Each case names one
    outcome, an earlier draft having asserted two contradictory ones for the no-lag path. It
    carries: a **no-lag** turn (asserts **no provisional is written**, the fire-time comparison
    finding the text already present — the case an offset rule duplicated on nearly every turn); a
    lagged turn with **three blocks of which only the tail is missing** (asserts a provisional IS
    written and is satisfied when the tail lands, deleting it); two turns lagging at once (asserts
    no cross-turn consumption); a turn whose blocks are `[X, X]` with the tail lagged (asserts the
    **measured residue** — the rule under-writes here, so the fixture pins the known-wrong outcome
    rather than a fiction, at 0.12% of turns); a **subagent** record sharing the parent's
    `sessionId` and `promptId` (asserts it does not touch the parent's provisional, that
    `SubagentStop` writes its own, and that closure reads the **agent's** transcript rather than the
    session's — the file-set error that would otherwise promote every subagent turn); a `Stop` payload with **no `prompt_id`** (asserts
    `provisional_skip`, not a `""` key); a payload with **no `last_assistant_message`** (asserts
    `provisional_skip` with reason `no-last-assistant-message`, 4.1% of real rows); a **zero-text turn whose payload DOES carry text** (asserts the
    provisional is promoted and marked `attribution='unverified'`, 5.4% of turns — not "no row",
    which an earlier draft asserted from a wrongly-keyed 1.6%);
    a **final** lagged turn with no `SessionEnd` (asserts the provisional is readable while
    outstanding, and that the **next `SessionStart` sweep** promotes it — naming the closer that
    fires, since the fixture controls which); and a
    record with unresolvable ancestry (asserts single-outstanding attribution, and
    `attribution='ambiguous'` when not). §V.3
    cannot see any of this: it counts a live corpus where a +1 is invisible. Also —
    a second `backfill` over the same corpus leaves **both act and `word`** counts unchanged — word
    counts over a corpus containing a **promoted-then-replaced** row, since that is the case rule 6's
    `block_seq` adoption exists to make idempotent and an act-only assertion cannot see, and the verb reaches its handler.
16. **[PROCESS]** `qlty check --no-progress --no-fix --filter=osv-scanner
    plugins/gray-area/tools/go.mod` — **expected: 0 findings.** Run on the branch before review,
    because the directive `go mod tidy` writes (`go 1.25.0`) scans at 46 findings and the one this
    change pins scans at 0. A non-zero result means the pinned patch release has accumulated CVEs
    and needs bumping — the maintenance this approach trades for not suppressing anything.
17. **[LOCAL]** `go test -run TestStopLagMeasured ./cmd/gray-area-capture/` — with the `Stop`
    hook bound, at fire time record transcript size, whether the payload's
    `last_assistant_message` is already present in the transcript tail, **and whether the payload's
    `prompt_id` resolves to the turn that just ended rather than the next one** — the untested leg
    of the identity claim in §II. Across a real session.
    **Expected: a number, reported either way** — "0 of N turns lagged" is a result; an absent
    measurement is not. This is the check the plan could not run before the hook existed.
18. **[LOCAL] End-to-end, observed.** Against the real corpus: `gray-area agents` names this session and
    the other five live ones; `gray-area touched <a path this session edited>` returns it;
    `gray-area sql 'SELECT tool, count(*) FROM action GROUP BY tool'` returns a distribution
    matching item 3's independent count **over the whole month window** — asserted against rows
    older than 11 days, because that is the bound a partitioned store would have imposed and the
    single-database design exists to remove.
