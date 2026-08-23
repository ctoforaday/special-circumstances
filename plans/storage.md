# Storage: JSONL now; embedded SQLite is the leading target (trigger = #62)

Status: **REOPENED** (2026-07-20). First cut said "DECIDED: keep JSONL" on three hard
constraints; two did not survive review and are struck below. What survives rules out a
*Postgres/framework* store but not an *embedded SQLite* one — which is now the leading
target, with the #62 concurrency work as its trigger.

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
> **A whole class of cross-run machinery is built and has never executed.** No archived run has
> ever reached round 3, and the bench sits at the END of a round, so `adjudicated` is empty at
> every seat in all four captured runs. The estoppel delivery on both sides (#499, #517, #524) is
> correct in code and unexercised in the record. If a backend changes what a run can read across
> runs, that becomes a live question rather than a settled one — and the first run to reach round
> 3 is what proves the existing half either way.
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

## Coupling (corrects the first cut's "decoupling")

Storage sits behind `feov-record`, so no *seat* sees it — but the **concurrency semantics #62
needs are exactly what the store provides.** So the storage choice should be made *as part of*
#62's design, not deferred as independent. `export --jsonl` keeps the plain-text artifact
whatever we pick.
