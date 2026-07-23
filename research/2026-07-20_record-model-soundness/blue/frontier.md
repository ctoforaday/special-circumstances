# blue frontier — Is the event-log-as-source-of-truth record model in frank-exchange-of-views sound, and where are its failure modes?

## H1: Deterministic Projection Hypothesis

**Claim**: If the event-log model is sound, identical event logs produce byte-identical projections every time, regardless of render timing, invocation order, or system state.

**If H1 is TRUE**: The rendering engine is a pure function with no side effects, hidden state, or temporal dependencies. Renders are idempotent and can be safely parallelized or repeated without risk of divergence. The tool's contract holds: `BoardState(log) → Projection` is deterministic.

**If H1 is FALSE**: Renders diverge due to non-determinism (floating-point formatting quirks, `jsNum` edge cases, non-stable map iteration, UTF-16 truncation logic, insertion-order sensitivity, or external clock/filesystem dependencies). A projection is only trustworthy for one specific render window, violating the claim that the event log is *the* source of truth—the projection becomes co-authoritative.

---

## H2: Causal Consistency & Event Ordering Hypothesis

**Claim**: If the model is sound, the event log's ordering guarantees happened-before semantics: no seat reads an event it hasn't written, all events are causally ordered, and replay always produces the same board state independent of read timing.

**If H2 is TRUE**: The timestamp/sequence logic enforces a total order. Event causality is preserved through every render and read. A seat's writes are atomic: an event enters the log or doesn't; partial writes and torn reads are impossible.

**If H2 is FALSE**: Timestamp collisions, NTP clock inversions, or seat-sequence interleaving produce ambiguous orderings. The tiebreak (SeatID, Seq) is insufficient or incorrect under concurrent writes. Renders replay events in different orders, yielding different boards. Or: `ts` field is absent/optional in old runs, making ordering undefined and renders non-reproducible.

---

## H3: Write-Read Atomicity & Staleness Hypothesis

**Claim**: If the model is sound, a seat's write through `feov-record` is immediately visible in the next `show --view` read. No staleness window exists where the event is on disk but not yet rendered into a projection a seat sees.

**If H3 is TRUE**: The tool's `show` command always renders fresh from the current event log, atomically returning the result. Write-then-read is linearizable: the writer's own read cannot see an older state than what it just wrote.

**If H3 is FALSE**: Renders are cached and lag behind the event log. A seat writes event E, then calls `show`; the render happened before E was committed, or the cached projection is stale. Or: multiple seats issue concurrent reads and each sees a different projection snapshot, violating the single-source-of-truth property. Or: the materialized `render-shadow/*.md` files (§VIII, tool-is-the-contract) are stale if a render lagged after a mint.

---

## H4: Complete State Reconstruction Hypothesis

**Claim**: If the model is sound, any auditor with only the event log can perfectly reconstruct the complete board state that any seat observed at any round, with no information loss and no hidden dependencies on markdown files, external state, or prior renders.

**If H4 is TRUE**: The rendering engine is pure and reversible. All board state flows through the event log. No computed values are stored outside the log and then read back. Red's hand-written `ledger.md`/`archive.md` are exact projections, not semi-authoritative summaries.

**If H4 is FALSE**: Rendering logic contains bugs (e.g., the `truncate` UTF-16 slicing edge case that can mangle text; `jsNum`/`jsText` formatting that diverges from JavaScript on certain values). Or: the `BoardState` computation drops events (e.g., bench closures were dropped in 2026-07-18 because `ts` was absent and seat-sequence ordering was wrong, §III tool-is-the-contract). Or: a computed field in a projection (e.g., `max_severity`, `repair_regression` ratio) depends on unserialized state and cannot be re-derived from the log alone.

---

## H5: Audit Trail Integrity & Immutability Hypothesis

**Claim**: If the model is sound, the event log is append-only, tamper-evident, and forms an immutable audit trail. No event can be deleted, modified, or reordered after-the-fact without invalidating checksums or breaking timestamps. Audits can trust the log.

**If H5 is TRUE**: The event log is protected by file permissions, append-only semantics, and logical monotonic timestamps that forbid inversions. Any mutation is detected; any gap in the sequence is visible. Red's audits of the board are final.

**If H5 is FALSE**: Events can be deleted or edited in the raw JSON files outside the tool. The ordering is contingent on soft-state (loose timestamp comparisons, fallback tiebreaks). Or: concurrent writes can corrupt the log (missing lock discipline, race conditions in `nextStamp`). Or: a render from a corrupted log silently produces an invalid board, and no downstream consumer knows.
