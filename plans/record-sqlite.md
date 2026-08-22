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
