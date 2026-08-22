# The record moves to SQLite

The storage layer, not the schema. `record.Append(Identity, proto.Message)` does not change
signature, so the ~500 fixture conversions already done on this branch stand.

## I. Why, in one paragraph

The protobuf migration fixed the TYPING. It does not fix the three problems that come from
sharded append-only JSONL, and those are where this branch's bugs actually live:

- **Ordering.** Each seat writes its own shard and replay merges by timestamp, so a ruling can
  replay before its filing. `record/motion.go:182-189` says so in capitals, the same bug shipped
  once in the function `compat.go` existed to be the legacy twin of, and `record.Motions()` exists
  only to paper over it. A foreign key makes the state unwritable.
- **The join, re-implemented per reader.** Eight readers key a bench disposition on `gap_id`, which
  `motion-rule` does not carry. Eight hand-written joins, each able to drift. In SQL it is one.
- **Concurrency.** Nonces, torn lines, `ReadShard` anomalies — all of it is multiple writers
  appending to files. WAL and a transaction delete the category.

Two more fall out: `NOT NULL` is declarative rather than a hand-kept `required.go` table (which had
already omitted `Outcome.verdict`), and derived counts stop being Go loops that can read zero
forever (`filed > ruled` computing `0 > 0`).

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

1. Derive the DDL from descriptors; test that every message and field is covered. ← here
2. The store: open, append in a transaction, read back.
3. Rewire `record.Append` and `BoardState` onto it; delete the shard/nonce/anomaly machinery.
4. Replace the hand-written projections with queries, one at a time, each with its test.

## The cutover (in progress)

**The seam is `MergedEvents`.** 28 call sites inside `internal/record` take `Merged` and
never touch a file. If `MergedEvents` returns events from SQLite, all 28 are untouched.
That is the whole reason this is one change and not thirty.

### What each piece becomes

| Today | After | Why |
|---|---|---|
| `shardPath` + one JSONL file per (seat, nonce) | one `record.db` per run | The shard-per-seat layout existed to avoid write contention. WAL plus `busy_timeout` is the same guarantee, held by the storage rather than by the naming scheme. |
| pointer file + `withLock` + `durableWrite` | the nonce of the seat's latest register row | A pointer file is a record standing outside the record. It needed a lock because two registers race; a query does not. |
| `seq` from `len(ReadShard(shard))` | `count(*)` for (seat_id, nonce) inside the insert transaction | Read-then-write was two steps with a gap; now it is one statement under SQLite's writer serialization. |
| `nextStamp` monotonic clock file | wall clock, informational | Ordering is `events.id`. The clock file existed because filename order was not event order — the bug that dropped a whole sitting's bench closures. |
| `appendLine` torn-line healing | — | A torn line is a JSONL failure mode. A transaction either commits or does not. |
| `ReadShard` + `ClassifyLine` stages | `recordsql.Events` | Undecodable-line classification has no analogue: a row is a row. |

### Two channels that must NOT become plausible zeros

- **`Merged.Discarded`** tells apart a healthy re-dispatch from one seat id used for two
  sittings, where the losing shard held work that exists nowhere downstream. `capture.go`
  gates a run on it. In SQLite BOTH sittings' events are rows: nothing is discarded, so the
  loss is **unrepresentable**, not merely absent. DELETE the field and the gate, with the
  reason recorded — leaving it always-empty is exactly the shape this migration exists to
  remove, and it would read as "no loss detected" forever.
- **`Merged.Anomalies`** has two producers. The replay-time ones (torn line, undecodable
  row, missing gap) go away — the first two are unrepresentable, and `missingGap` is now a
  foreign key. The **projection-time** ones in `viewjson.go` stay and are unrelated. So the
  field survives, its replay-time producers do not, and `view.go:124`'s headline count
  changes meaning: say so in the change.

### Validation loop

Re-arms on any edit under `internal/record/`:

    (cd plugins/frank-exchange-of-views/tools && GOTOOLCHAIN=go1.25.0 go build -gcflags=-e ./... 2>&1 | grep -c '\.go:[0-9]')
    (cd plugins/frank-exchange-of-views/tools && GOTOOLCHAIN=go1.25.0 go test ./internal/record/... ./internal/cli/...)
    (cd plugins/frank-exchange-of-views/tools && UPDATE_GOLDENS=1 GOTOOLCHAIN=go1.25.0 go test ./internal/record/recordsql && git diff --stat)

Known-red and NOT a regression: `internal/difftest`, `internal/flags`
(`TestGradeValueSetRejections`), `internal/proof` — all three fail identically at HEAD,
verified by stashing. The remaining build failures are the in-flight test-fixture
conversion (capture, verify, view, fuzz, report, cli, record).

## What the cutover actually found (2026-08-22)

Nine production defects, none of which the pre-migration suite could have found. Recorded
here because the pattern matters more than the list: **every one was invisible because the
thing that would have caught it was itself broken, absent, or reading a fact out of the
wrong shape.**

| Defect | Why nothing caught it |
|---|---|
| `merge close` wrote `successor = ''` for an absent flag — every close in the tool failed once successor became a reference | Before the reference, the row was written and read as a closure whose successor was the empty gap. "Never said" and "said nothing" were the same bytes. |
| A re-dispatched seat could not record at all: the ordinal was per-sitting, `events.key` is global | Impossible under shards — the retry wrote the same keys into a NEW FILE and replay picked a winner. The storage change turned a tolerated duplicate into a refusal. |
| The sqlite driver's blank import was in `schema_test.go` | Every test had a driver; the binary had none. A blank import is invisible to the unused check. |
| 8 concurrent seat processes lost ~half their writes to `SQLITE_BUSY` | A hazard the storage change INTRODUCED. The first regression test used goroutines and passed with the fix reverted. |
| The debate view and the report printed `DISPOSITION_CLOSED` | Typing the enum made `%s` silently wrong — no type error, no test failure. |
| The fuzz prose gate was inert for nine event types (`reason` vs `text`/`rationale`/`basis`) | A payload map returned `""` for a wrong key and the rule was skipped. The gate reported coverage over rules that could not fire. |
| The coverage census grepped `Append(..., "type")`, a call shape that no longer exists | An empty set of ungated types reads exactly like full coverage. |
| `--as supports-with-bridge` was advertised in `--help` and refused by the write path | Nothing compared the advertised set against the schema. A comment called it a caveat "not mine to fix". |
| A required prose field was satisfied by `""` | The annotation collapsed two flavours of requiredness (present, present-and-non-empty) into one. |

Plus one gate reporting a **false** failure: the mass-parity regex matched `'low_medium'`
as `medium`, carried a wrong value, and reported two keys absent from a file that declares
both.

### The method that found them

Not the migration itself — the migration only made them *reachable*. What surfaced them:

1. **Driving the real binary**, not the library. Four of the nine only appear in
   `cmd/feov-record`.
2. **Making a miss LOUD.** `fieldStr` fails on a field name the schema does not carry;
   the enum census fails on a declared word the record cannot hold. Both found their
   defect within one run of being written.
3. **Reading generated output with eyes** — the golden schema, which found the arm tables
   with no foreign keys, and the rendered `--help`, which found the doubled `REQUIRED —`
   and cobra eating a backquoted value as the flag's placeholder.
4. **Asking whether a test can fail.** Two of my own passes produced assertions that
   could not: a Grade comparison against an `any`-typed string field (always true), and a
   blanket fixture rewrite that filled in the very fields four cases existed to omit.
