# The record in SQLite

> Companion: [`historical/record-sqlite.md`](historical/record-sqlite.md) holds the archaeology — the sharded-JSONL storage this replaced, the defects the cutover surfaced and fixed, the decisions that were reversed, and the measurements that justified them.

> STATUS 2026-09-05: **shipped and in production.** `internal/record/recordsql` is the store (landed on main via #556, `record-protobuf-pr2`); the `Open` schema race is fixed (#632, `0fa2ce93`, pinned by `recordsql/openrace_test.go`); step 4's projections are queries (#700, #701, #703) and its follow-on views landed in `e9c94e36`. One thread is still open and it is named at the end: the remaining Go folds, `BoardState` above all.

This is the storage layer, not the schema. `record.Append(Identity, proto.Message)` did not change
signature, which is why the protobuf fixture work stood through the cutover.

## II. Decisions

1. **Driver: `modernc.org/sqlite`** — pure Go, no cgo, so cross-compilation and the release build
   stay as they are. Matches [[committed-tooling-is-go]].
2. **The schema is DERIVED FROM THE DESCRIPTORS AT DATABASE CREATION.** Not a committed `.sql`, not
   struct tags, not `protoc-gen-go-tags`. A DDL file is a second copy of the proto and would need a
   staleness gate; tags put the mapping on generated structs and pull in an ORM. Deriving it in
   process means the schema CANNOT drift from the schema, and there is no gate to write.
3. **No migrations, because a run directory IS the database.** Every run creates its own, so a
   schema change never meets an old database. This is the same ground `a12362c` stood on: a project
   in building mode whose every record is a test run.
4. **Class-table layout**: an `events` envelope table plus one table per body type, keyed on the
   event id. NOT a serialized body column — a blob would buy transactions and ordering while
   forfeiting joins, constraints and aggregate queries, which are the whole reason to move.
5. **`STRICT` tables, WAL, `foreign_keys=ON`.** SQLite's default affinity is loose enough to undo
   the point of typing the record at all.
6. **Append-only is enforced, not merely intended**: `BEFORE UPDATE`/`BEFORE DELETE` triggers that
   raise. The record is evidence in an adversarial process; nothing may rewrite it.
7. **Enums become `TEXT` with a `CHECK` over the schema's own spellings**; repeated fields become
   child tables so they stay joinable; oneofs become nullable columns with a `CHECK` that exactly
   one arm is present.

## III. Order of work

1. Derive the DDL from descriptors; test that every message and field is covered. — landed
2. The store: open, append in a transaction, read back. — landed
3. Rewire `record.Append` and `BoardState` onto it; delete the shard/nonce/anomaly machinery. — landed
4. Replace the hand-written projections with queries, one at a time, each with its test. — landed
   (#700, #701, #703); the thread that continues from it is the last section of this file.

`internal/record/queries.go` names this file and this step in its own header comment, and
`internal/view/view.go` and `internal/dashboard/model.go` name step 4 at their conversion sites —
so the path `plans/record-sqlite.md` is load-bearing from Go.

## The storage model as built, and what each piece replaced

| Was (sharded JSONL) | Now | Why |
|---|---|---|
| `shardPath` + one JSONL file per (seat, nonce) | one `record.db` per run | The shard-per-seat layout existed to avoid write contention. WAL plus `busy_timeout` is the same guarantee, held by the storage rather than by the naming scheme. |
| pointer file + `withLock` + `durableWrite` | the nonce of the seat's latest register row | A pointer file is a record standing outside the record. It needed a lock because two registers race; a query does not. |
| `seq` from `len(ReadShard(shard))` | `count(*)` for (seat_id, nonce) inside the insert transaction | Read-then-write was two steps with a gap; now it is one statement under SQLite's writer serialization. |
| `nextStamp` monotonic clock file | wall clock, informational | Ordering is `events.id`. The clock file existed because filename order was not event order — the bug that dropped a whole sitting's bench closures. |
| `appendLine` torn-line healing | — | A torn line is a JSONL failure mode. A transaction either commits or does not. |
| `ReadShard` + `ClassifyLine` stages | `recordsql.Events` | Undecodable-line classification has no analogue: a row is a row. |

## Two channels that could have become plausible zeros

- **`Merged.Discarded` is deleted.** It told apart a healthy re-dispatch from one seat id used for
  two sittings, where the losing shard held work that exists nowhere downstream, and `capture.go`
  gated a run on it. In SQLite BOTH sittings' events are rows: nothing is discarded, so the loss is
  **unrepresentable**, not merely absent. The field and its gate are gone with the reason recorded
  at `record/replay.go` and `record/agentbinding.go` (`DiscardedForSeat` is not ported, and its
  absence is the honest answer rather than an omission) — leaving it always-empty is exactly the
  shape this migration exists to remove, and it would read as "no loss detected" forever.
- **`Merged.Anomalies` is deleted too.** It had two producers. The replay-time ones (torn line,
  undecodable row, missing gap) are gone — the first two are unrepresentable under a transaction,
  and `missingGap` is now a foreign key — which left no producer at all, so the field went with
  them. `view.Counts` DROPPED its third return rather than reporting a constant 0, because every
  caller reading "0 anomalies" as a clean board would be reading it in the same words it used when
  the number meant something. The **projection-time** anomalies in `viewjson.go` are a different
  field with a different meaning and are untouched: a dropped mutation is still surfaced there,
  never swallowed.

  *(The cutover plan expected `Merged.Anomalies` to survive with only its replay-time producers
  removed. It did not, and the headline count was removed rather than re-meant.)*

## Validation loop

Re-arms on any edit under `internal/record/`:

    (cd plugins/frank-exchange-of-views/tools && go build -gcflags=-e ./... 2>&1 | grep -c '\.go:[0-9]')
    (cd plugins/frank-exchange-of-views/tools && go test ./internal/record/... ./internal/cli/...)
    (cd plugins/frank-exchange-of-views/tools && UPDATE_GOLDENS=1 go test ./internal/record/recordsql && git diff --stat)

**No `GOTOOLCHAIN` pin.** The cutover's loop carried `GOTOOLCHAIN=go1.25.0`; `go.mod` now asks for
`go 1.25.13`, so that pin REFUSES to run at all (`go: go.mod requires go >= 1.25.13`). Use the
toolchain go.mod asks for. Last run 2026-09-05: build 0 errors, `./internal/record/...` +
`./internal/cli/...` 10 packages green.

The third command regenerates `recordsql/testdata/schema.sql`. It is a golden to be READ, not
diffed away: it is the only place the derived DDL is visible as a document, and reading it is what
found arm tables with no foreign keys.

## Which questions are queries, and what deliberately stays a fold

**The hand-written projections are queries** (#700 board slices, #701 write-path lookups,
#703 motion joins — one PR per mechanism group, each conversion holding its fold's exact
contract and carrying a parity test that seeds the fold's edges through the real write
path). `board_counts` and `gap` have production readers; `convergence_vs_verdict`
already did. The write path no longer replays the whole record to mint an id or check a
reference — every such question is one indexed SELECT. Since `e9c94e36` the questions that had
landed as SQL strings inside Go functions are view columns instead: `gap` carries the regrade
overlay (`current_*`), `proof_answered`, `awaiting_proof`, `superseded_by`, `stranded` and
`minted_event`, and `motion_answers` and `line_of_inquiry` answer their whole families. The motion
guards read `motion_answers`; first-wins is stated ONCE, where a record carrying an illegal second
ruling cannot multiply anybody's join.

**What deliberately stays a fold, and why it is a decision rather than a leftover** (each
reason also lives at its site): the consistency oracle's raw walk (converting the ground
truth into a projection makes the oracle a tautology); `lastActivity` (SQL's date parsing
is a DIFFERENT timestamp parser than Go's RFC3339Nano and silently mis-sorts a
fraction-less stamp); `CitedSources`/`CitationLabels`/`RecordedProofs` (one Go rule shared
with board-shaped twins — a SQL fork recreates the two-copies defect); seatprobe (needs
full event bodies, which is `Events`'s job); `requirePassClosesAllGaps`'s motion arm (its
"unruled" is not "no rule row": an empty-armed ruling reads as unanswered and a direction
motion is created by its ruling; the gate runs once per verdict).

## Presence for repeated fields: measured, not reasoned

**The question.** A repeated field lands in a child table, so "does this gap supersede
anything" needs a join. Could an insert hook on the child set a boolean on the parent, so
presence comes for free with the row?

**The mechanism works.** An `AFTER INSERT` trigger on the child updating the parent does
what you would expect — measured: parent with a child reads 1, parent without reads 0.

**It cannot recover the distinction that bit us.** An explicitly-empty list and an absent
one are lost ABOVE SQL. Measured on the real messages:

    unset             Has=false  len=0  nil=true
    explicitly empty  Has=false  len=0  nil=false
    one entry         Has=true   len=1  nil=false
    wire bytes identical: true

`protoreflect` reports no presence for either, and the two marshal to the same bytes. By
the time a writer reaches the storage there is nothing left to tell apart, so no trigger
can recover it.

**Three costs, one of them fatal to the strongest use.**

1. It is a **second copy** of a fact the child table already holds — derivable by `EXISTS`,
   which is what `facts-are-fields` says to prefer generating over guarding.
2. It **cannot be used by a CHECK.** The parent row is written before any child, so a
   `CHECK (has_kids = 1)` fails at the moment the parent is inserted. Measured:
   `CHECK constraint failed: has_kids = 1`. That was the one thing a stored column could do
   that a view cannot, and it does not work.
3. It requires an **UPDATE on an append-only record.** Measured: the same guard that
   protects `events` refuses it (`constraint failed: append-only`).

**What was done instead.** The `gap` view carries `supersedes_count` and `found_by_count`.
Free, derived, cannot disagree with the rows it counts, and it answers the join several
readers were writing for themselves.

**Where the fact is a CLAIM rather than a consequence,** the schema already has the right
shape and it is a field the WRITER sets: `SpotCheck.none` — a bool with real presence,
meaning "I checked nothing, deliberately". Derived where derivable; declared where claimed.

## Where SQL earns its place, and where Go keeps it

The cutover is a good sample: nine defects, several constraints added, two of them wrong.
The pattern is sharp enough to state as a rule.

**Ask what the thing DOES, not where it could live.**

**Refusing a state that is unconditionally illegal → SQL.** This is where the whole move
paid. `gap_id` referencing `mint.gap_id` turned "a dangling reference is an ANOMALY
discovered on read, per reader, if any reader looks" into "the row cannot be written."
`events.key` UNIQUE turned a silent read-time dedup into a refusal the seat sees. The
transaction replaced torn-line healing. None of that is logic moved out of Go — it is the
DATA MODEL doing what Go was doing badly, and it holds against anything writing to the
file, which for a record that is EVIDENCE is the point.

**Refusing a state that is conditionally illegal → Go.** Both constraints that had to be
removed were this shape. `motion_rule.motion_id` referencing `motion.motion_id` is right
for two subjects and false for the third, because a direction motion has no motion row.
`Avenue.line` required is right for a proposal and wrong for a move. SQL cannot see the
condition: a CHECK holds no subquery, and a foreign key has no idea what subject the row
is. **A constraint that is wrong for a third of its cases is worse than none** — it
refuses correct work while reading as a guarantee.

**Explaining a refusal → Go, always.** `RAISE(ABORT, …)` takes a STATIC string. Every
refusal in this tool that earns its keep names the flag, the set, and what the omission
costs: "merge close: `carried` defers the gap instead of closing it, and deferring is the
BENCH decision". SQL can refuse; it cannot teach. Where a constraint and a message are both
wanted, the constraint goes in SQL as the wall and the message stays in Go as the door —
that is why `merge close`'s subset check exists twice on purpose.

**Computing a derived value → a VIEW, not a trigger.** Same language, opposite direction:
a view is SQL used as a QUERY, which is what it is good at. A trigger maintaining a
denormalised column is logic, and it buys nothing a view does not — see the presence
section above for the measurement.

**The one trigger that earns its place** is the append-only guard on `events`. It refuses
an unconditionally illegal state (editing a written event), and its static message is
adequate precisely because there is nothing conditional to explain: you cannot edit the
record, and that is the whole sentence.

### Presence for a list, if a verb ever needs it

The proto answer is a wrapper message (`optional Lineage supersedes` where `Lineage {
repeated string values = 1; }`) — message fields have always had presence, so set-but-empty
is distinguishable from absent. Confirmed: the wire bytes differ. `optional repeated` is a
syntax error in every proto version; the labels are one slot.

**We should not use the wrapper here.** It exists to work around proto's lack of
repeated-field presence at the LANGUAGE BINDING level — it buys `has_supersedes()` in every
language. This schema is DERIVED from descriptors with our own annotations, so we do not
have that constraint, and the wrapper costs us specifically: a message field becomes its own
table, so one list becomes TWO tables (a wrapper table whose only column is `event_id`, plus
the list table under it). Avoiding that needs a `presence_only` annotation AND generator
support to collapse the wrapper and re-key the list to the grandparent — real work to undo a
workaround we adopted for a limitation we do not have.

**The flat shape gets the same fact for nothing:**

    repeated string supersedes = 20;
    optional bool supersedes_stated = 21;   // "I considered lineage; there is none"

One column, no nesting, no second table, no generator change — and it is the idiom this
schema already uses well (`SpotCheck.none`).

**And it is the better model, not just the cheaper one.** A wrapper makes presence a
STRUCTURAL fact (is the message set); a sibling bool makes it a DECLARED fact (did the seat
say so). This record wants the second: a claim should be a field a writer can be REFUSED on.
An empty wrapper the CLI sets because a flag was registered is not a seat asserting
anything, and `--supersedes ""` should not become "I thought about lineage and there is
none" by accident.

**Not built.** No verb produces the distinction today, and inventing a record for a fact
nobody states is the cost facts-are-fields warns about. If a bench ever needs "red
considered lineage and found none" as distinct from "red did not address lineage", that is
the moment — and it is one field.

## The storage posture, measured (2026-08-23)

Four settings, each with the measurement that justifies it. Three were already right; two were
missing and one was actively expensive.

| setting | why | measured |
|---|---|---|
| `journal_mode(WAL)` | concurrent readers across processes | — |
| `busy_timeout(5000)` | SQLite resolves cross-process races internally before Go sees an error | — |
| `_txlock=immediate` | a deferred BEGIN upgrades read→write, and SQLite deliberately will NOT apply busy_timeout to an upgrade | 8 processes lost ~half their writes without it |
| `SetMaxOpenConns(1)` | **was missing.** Two goroutines on two pooled connections collide in SQLITE_BUSY; with one they queue on a Go mutex | — |
| schema in ONE transaction | **was 171 implicit transactions, each fsyncing** | **525ms → 39ms per fresh database** |

### The schema was the expensive one, and it cost nothing to fix

`Open` applied 171 DDL statements. SQLite autocommits any statement not already in a transaction
and every commit is an fsync, so creating a run directory paid 171 disk syncs — to write a schema
that is DERIVED and regenerated for free whenever `events` is absent.

Every alternative traded durability and was still slower:

    as-is (171 implicit transactions)   499ms   full durability
    synchronous=normal                   67ms   can drop recent commits
    synchronous=off                      46ms   corruption risk
    ONE transaction                      39ms   full durability, unchanged

One fsync instead of 171 is not a relaxed guarantee — it is the same guarantee, asked for once.
SQLite's own forum states the mechanism: an autocommitted statement fsyncs on its own, so
batching replaces one-fsync-per-statement with a single fsync at COMMIT.

### What is NOT there, stated so nobody assumes it is

**No retry, anywhere on the write path.** The only retry in the record layer is
`renameWithRetry`, a Windows rename workaround with LINEAR backoff and no jitter. Everything else
relies on `busy_timeout`, and SQLite's default busy handler is a FIXED schedule with no jitter —
`{1,2,5,10,15,20,25,25,25,50,50,100}`ms, 100ms thereafter — so contending waiters wake in
lockstep. Past five seconds it is a hard failure and the seat is told its write failed.

`SetMaxOpenConns(1)` removes the in-process half of that surface entirely, so what remains is
seat-vs-seat across processes, which the 8-process test exercises and currently passes. Adding
exponential backoff with full jitter (`avast/retry-go` has `FullJitterBackoffDelay`;
`cenkalti/backoff` is a port of Google's algorithm) is the standard answer IF that test starts
to show strain — but it is a dependency added on evidence, not on reasoning, and the evidence is
not there yet.

### One trade taken knowingly

`SetMaxOpenConns(1)` serializes the dashboard server's per-request renders, which run in
`http.Server`'s per-request goroutines. It is a local single-operator dashboard, not a
throughput-sensitive service, and the fix if it ever matters is a separate read-only handle for
the server rather than reverting the pool setting — the seat write path is what the setting
protects.

## Still open

**Re-verifying one source in one sitting is now refused.** A verify keys on its reference
(`url` is in `keyFields`), so two verifications of one source share a key. Under shards
both were written and the reader kept one — "idempotent, updates in place" was a read-time
illusion over an append-only log. `events.key` is UNIQUE now, so the second is refused.

The loss is real: a lens that re-reads a source mid-sitting and finds something different
cannot record the second reading. Three ways out, none obviously right:

1. **Leave it.** The refusal follows from append-only plus one-act-per-key, both
   deliberate, and it teaches. A seat that must revise says so in a later sitting.
2. **Drop `url` from `keyFields`** so verifies take an ordinal. Re-verification works, and
   a crash-retry of the same command now writes a SECOND event instead of being idempotent
   — which is what the key was for.
3. **Scope the key to the reading, not the source** (reference + outcome, say). Both cases
   work; the key stops being derivable from one field, which is how `keyFields` goes stale
   silently (record.go's own note on `reference`).

Pinned by `TestBoardCountsCiteEvents`, which asserts the refusal so the behaviour cannot
change without someone editing the assertion and reading this. Verified 2026-09-05: `url` is
still in `keyFields` (`record/record.go`) and the test still asserts it.

Related, and decided the other way for a reason worth reading: `ExistingCorroborationLabel`
(`record/citationid.go`) reached the SAME cost by another route — moving a corroboration's key off
`url` and onto a freshly-minted label meant nothing collided, so a crash-retry wrote a DUPLICATE.
That is the concrete price of option 2 above, already measured on a neighbouring verb.

**`--check ""` and friends.** Requiredness is present-and-non-empty, with `allow_empty` as the
narrow exception declared at the field. `principle` was TIGHTENED (`b37232cb`, 2026-08-22) — a
ruling with no stated rule is what `bench opinion` exists to prevent — and `tension`, `review_flag`
and `settled` still carry `allow_empty`, each with its own reason at the field: demanding prose
there produces invented tension, pro-forma flags and restatements of the disposition, which read as
reasoning and are worse than an honest blank. Whether `tension` should follow `principle` is the
live remainder of the question.

**`observe` has no verb at all.** The event type is in the schema and nothing in the
command tree writes it — the only Appends of a `recordpb.Observe` are in record's own
tests. Exempted with that stated, rather than driven (a drive would have to invent a verb)
and rather than dropped from the list (the type's homelessness should be a line somebody
reads). The `Observation` a board carries is built from FINDING events.

Verified 2026-09-05: the exemption is at `releasegate/fuzz/fuzz_test.go` (`"observe": true`) with
  that reason written beside it.

**The goal this line is still walking toward, and it is not a perf question:** the remaining Go
walks — `BoardState`'s fold above all. The point of step 4 is simplification: a question
about the record is authored ONCE, as SQL a reader can see and a test can hold, instead of
Go pulling tables and recomputing what the database can answer. Perf was the trigger that
paid for the plumbing; it is not the bar. Next: the board's own readers move onto views,
retiring the fold reader by reader.
