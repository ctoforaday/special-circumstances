# Storage: the record IS an embedded SQLite database (trigger = #62)

> STATUS 2026-09-02: shipped — historical record. `internal/record/recordsql` is production; `export --jsonl` was never built (marked below) and #68 was answered by the event-schema epoch (#597), not migrations.

Status: **LANDED** (2026-08-28). First cut said "DECIDED: keep JSONL" on three hard
constraints; two did not survive review and are struck below. What survived ruled out a
*Postgres/framework* store but not an *embedded SQLite* one — and the embedded store is what
shipped: `modernc.org/sqlite` (pure Go, no cgo), `internal/record/recordsql`, and one
`record.db` per run directory. JSONL is not a fallback and not a reader: a pre-database run
directory is **refused by name** rather than read as an empty board
(`internal/record/legacyformat_test.go`) — which is the one thing the file format could not do,
because six archived runs read under the new binary reported a clean zero with every invariant
passing vacuously. The sections below are kept as the reasoning that got here; where a
paragraph has since been falsified by the record it is corrected in place, not deleted.

> **EVIDENCE ADDED 2026-08-23.** Two findings from the record-store debate and this session's
> work, both bearing on what a SQL backend has to be responsible for.
>
> **The cross-run authority question was argued and answered.** The
> `2026-08-22_record-store-authority` run — 5 lanes, opus judgment, 2 rounds, both red verdicts
> FAIL, $74.07 — concluded: *the log is a disposable per-run cache today, but the binary is the
> wrong shape, because cross-run authority already exists downstream in human-reviewed
> `law/precedents.md`.* So the store is not being asked to become the authority; the authority
> has a home and it is reviewed by a person. That narrows what indexing has to serve.
>
> ~~**A whole class of cross-run machinery is built and has never executed.** No archived run has
> ever reached round 3, and the bench sits at the END of a round, so `adjudicated` is empty at
> every seat in all four captured runs.~~
>
> **FALSIFIED 2026-08-28, by the record it appealed to.** The struck paragraph was mine and it was
> wrong within hours of being written. `2026-08-23_sleeper-service-plan` reached **round 4** —
> `red-lens-r4-{L1,L2,L5,L6}`, `red-merge-r4`, `blue-respond-r4`, and `judge-r2`, `judge-r3`,
> `judge-r4`, `judge-terminal` — and the estoppel machinery ran:
>
> - **10 opinions, 10/10 carrying `settled`**; 7 `reopens_on` + 3 `final` = 10, so every single
>   opinion answered the estoppel question. #502's stated mechanism risk — *"a new required field
>   is a field the bench can default"* — did not materialise.
> - The `settled` sentences are real propositions, not filler: *"Blue may not be held to an
>   uncaveated medium-high impact label for R3-2; medium is the settled figure by unanimous
>   agreement of bench, red and blue."*
> - Three mints name superseded gaps — `R3-1←[R1-1]`, `R3-3←[R1-2,R2-3]`, `R4-3←[R3-1]` — a
>   three-generation lineage, which is #499's designed successor path running for real.
>
> What remains unexercised is narrower than the struck claim and is #524's thesis, now with a live
> instance instead of a hypothetical: **two of those rulings went blue's way and blue never saw
> them.** Under current code blue receives a bare subtraction from the docket, not the sentence
> that says why. That raises #524's priority; it does not lower it.
>
> The claim also under-counted the corpus: **six** archived runs, not four.
>
> **Adjacent, landed since this plan was written:** the run-live marker is now a LIST
> (#529/#530), so more than one run can be open at once and `internal/runlive` owns the file.
> Any store that assumes one live run per project is assuming something that stopped being true.

## TL;DR of the correction

- Git commits a `.db` fine → **provenance never required plain text**; only the *projections*
  (report.md, etc.) need to be plain text, and they stay markdown regardless of the store.
- SQLite is embedded → **no server, no "ephemeral writer can't hold a connection" problem**;
  a one-shot CLI opens the file in a transaction. (MCP is an *optional* stronger form:
  a persistent server that collapses cross-process locking into in-process serialization.)
- SQLite's `BEGIN IMMEDIATE` transaction **fixes the `MintGapID` race for free** (correct
  cross-process write serialization, no proceed-unlocked-on-timeout hole) — so it is *more*
  correct under concurrency, not less. Storage is therefore **coupled to #62**, not decoupled.
- **Indexing is the real driver.** `MintGapID`/`MergedEvents` are O(all-events) full scans;
  an index is O(log n). Files can't give this; it's the one capability we genuinely lack.
- Remaining costs are thin: a driver dependency, loss of native `git diff` on raw events
  (mitigate with `export --jsonl`), and inertia. No *urgent* forcing function at today's scale.

**NOT BUILT (2026-09-02):** no `export` verb exists anywhere in the tool (`grep -rn '"export"' tools/` finds nothing outside prose) — the `export --jsonl` mitigation this file cites four times was never implemented. The diffable-artifact need was met by projections instead.

## Context

The record model stores events as append-only per-seat JSONL shards
(`events-<seat>-<nonce>.jsonl`), read by `feov-record` into a replayed board. Red flagged
that off-the-shelf event stores were dismissed without a cost-benefit. A Go survey offered
three paths (Watermill messaging; goes/Chronicle frameworks; dynamic-streams/hallgren
DB-backed). We had earlier floated "maybe a SQL db."

## Constraints, corrected

The first cut rested on three. Two fall:

1. ~~Writers are one-shot sandboxed CLI processes, so no store needing a held connection.~~
   **STRUCK.** SQLite is embedded — a one-shot CLI opens the file transactionally, no server,
   no held connection. (And MCP would give a persistent server if we ever wanted one.) This
   even *helps*: a SQL transaction is the correct fix for the `MintGapID` race.
2. Concurrency is low and sharded by seat. **Holds, but weaker than it reads** — #62 adds
   concurrent writers, and SQLite serializes them correctly where our `flock`-with-unlocked-
   fallback does not.
3. ~~The run artifact must be git-tracked plain text.~~ **STRUCK as stated.** Git tracks a
   `.db` (provenance = commit + pin, intact). Plain text buys *inspectability*, which is (a)
   only load-bearing for the **projections** (they stay markdown), (b) being retired for red
   (#62 moves it to tool queries, not `grep`), and (c) recoverable on demand via
   `export --jsonl`. The real requirement is "projections are plain text," which no event
   store touches.
4. **The aggregate is domain-specific** (mint / close-with-regression / supersede /
   risk-accept). **Holds** — a generic framework is a worse fit than the `BoardState` replay
   we have. This rules out goes/Chronicle/Watermill, not an event *table* under our own code.

## Options and verdicts

| Path | Verdict |
|---|---|
| Watermill (messaging/sagas) | **Reject** — routes between running handlers; we have none. |
| goes / Chronicle / thefabric-io (frameworks) | **Reject** — generic aggregates; ours is the board, already built. |
| dynamic-streams-eventstore-go (Postgres) | **Reject** — a managed server for one-shot CLI spawns; wrong operational shape. |
| **Embedded SQLite** (hallgren adapter, or our own table) | **LEADING TARGET** — no server, ACID, correct cross-process write serialization (fixes the mint race), and **indexed queries** (fixes the O(all-events) scans). Cost: a driver dep + JSONL export for the diffable artifact. |

## Verified at the leaf (2026-07-20) — corrects red's guesses

- **Durability is already correct.** `durableAppend` does `O_APPEND` → `fsync` file → `fsync`
  dir inside a SIGINT critical section. Red's "does not document fsync" was a guess about a
  binary it couldn't read (#64). No fsync work needed either way.
- **`MintGapID` is an unguarded read-modify-write on a shared counter** — safe only because
  minting is single-process today; concurrent minters (which #62 introduces) → duplicate IDs.
  On JSONL the fix is a fail-closed `withLock(runDir, "mint", …)` or the #62 design (findings
  not concurrent mints, merge as single serializer). On SQLite the fix is a transaction — free.

## Decision (provisional)

**Stay on JSONL until #62; do #62's concurrency correctly; treat embedded SQLite as the
migration that #62 justifies.** Nothing forces a move at today's scale (small runs,
single-minter, red's failures were fabricated). But when #62 lands concurrent writers, the
choice is "hand-roll fail-closed locking around a shared counter on JSONL" vs "one SQLite
transaction" — and the DB also retires the full-scan cost and hands us #68 (schema migrations)
on a paved road. Embedded SQLite + `export --jsonl` for provenance is the front-runner; decide
for real when #62 is designed, not before.

## Remaining actionable

- **Schema evolution / versioning** (#68): version + upcasters on read. On SQLite this is a
  migration; on JSONL a read-time upcast registry. Latent until the first schema change with
  historical events on disk.
  **SUPERSEDED 2026-09-02:** answered by the event-schema epoch (#597) plus no-migrations-by-design — every run creates its own database, `setup` refuses a binary whose epoch differs (`internal/record/schema_gen.go`), and a pre-epoch run is refused by name rather than upcast.

## Coupling (corrects the first cut's "decoupling")

Storage sits behind `feov-record`, so no *seat* sees it — but the **concurrency semantics #62
needs are exactly what the store provides.** So the storage choice should be made *as part of*
#62's design, not deferred as independent. `export --jsonl` keeps the plain-text artifact
whatever we pick.
