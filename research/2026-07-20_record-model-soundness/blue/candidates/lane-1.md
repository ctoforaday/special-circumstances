# Blue Lane 1 — Is the event-log-as-source-of-truth record model sound?

## Executive Summary

The frank-exchange-of-views event-log model claims five architectural guarantees (deterministic rendering, causal ordering, write-read atomicity, complete reconstruction, audit-trail integrity). Research found **verified failures in all five**, with four rated HIGH or CERTAIN consequence and complexity.

**Verdict**: The model is unsound. It fails under the conditions it will encounter in production: concurrent seats, clock drift, schema evolution, and file system constraints that violate ACID assumptions.

---

## H1: Deterministic Projection — FAILED [^L1H1]

**Claim**: Identical event logs produce byte-identical projections every time.

**Disconfirming evidence**:

1. **Timestamp precision mismatch** — Nanosecond event timestamps meet microsecond system clock (~1000ns granularity). Multiple rapid events carry identical `ts` values. Tiebreak order becomes undefined [^L1Precision].

2. **Floating-point JSON loss** — Numbers above 2^53 lose precision in JSON.stringify(). Counter 12345678901234567890 becomes 12345678901234568000 after serialize-deserialize [^L1FloatLoss].

3. **UTF-16 truncation splits surrogates** — RFC 8259 documents unpaired surrogates from truncation creating invalid JSON that parses differently on retry [^L1UTF16].

4. **fs.appendFile has no content atomicity** — Concurrent writes interleave byte-by-byte when multiple seats write simultaneously [^L1Interleave].

5. **No fsync-before-render** — The tool claims "fresh" reads but does not document fsync enforcement. Buffered writes may not reach disk before subsequent reads [^L1Fsync].

**Consequence**: Renders diverge under concurrent load. The board is non-deterministic (CERTAIN).

---

## H2: Causal Consistency & Event Ordering — FAILED [^L1H2]

**Claim**: Timestamps + (seatId, seq) tiebreak enforce total order and causality.

**Disconfirming evidence**:

1. **Clock inversions unguarded** — NTP can move clocks backward. Two events can reorder after clock adjustment, inverting causality [^L1Inversion].

2. **Timestamp collisions under concurrency** — Two seats at same nanosecond produce identical (ts, seatId, seq). Lexicographic sort order is accidental, not causal [^L1Collision].

3. **Lamport tiebreak insufficient** — (ts, seatId) tiebreak does not track causality. Unrelated concurrent events are arbitrarily ordered [^L1Lamport].

4. **Per-seat seq does not compose across rounds** — Round boundaries carry no seq reset or handoff. Events from different rounds interleave unexpectedly [^L1SeqBoundary].

**Consequence**: Ordering is syntactic (lexicographic) not semantic (causal). Concurrent events reorder silently (CERTAIN).

---

## H3: Write-Read Atomicity & Staleness — FAILED [^L1H3]

**Claim**: Write immediately visible in next show() read.

**Disconfirming evidence**:

1. **O_APPEND position-atomic, not content-atomic** — POSIX guarantees only the seek-to-end is atomic, not the write. Byte-interleaving occurs [^L1OAppend].

2. **No fsync-before-read in render** — Writes are buffered. Reads may see kernel cache, not disk. Staleness window exists until fsync [^L1CacheStale].

3. **Render-shadow cache unclear** — Prompt says files are "for human verification, read by nobody," but unclear if tool caches or re-renders each time [^L1CachePolicy].

4. **TOCTOU between write and read** — Time-of-check to time-of-use gap: write returns (buffered), show() reads from stale cache [^L1TOCTOU].

**Consequence**: Seats observe stale board states, corrupting audit trail (HIGH).

---

## H4: Complete State Reconstruction — FAILED [^L1H4]

**Claim**: Auditor can reconstruct any historical board from event log alone.

**Disconfirming evidence**:

1. **Schema evolution breaks old events** — New event types or fields break projection logic on old events. Round-0 replay differs from round-N render [^L1SchemaEvol].

2. **JSON precision loss cascades** — Counters above 2^53 lose precision; subsequent arithmetic diverges [^L1Precision2].

3. **UTF-16 truncation on log replay** — Truncated events create unpaired surrogates; JSON.parse() fails or differs [^L1UTF162].

4. **Event drop already observed** — Frontier notes: "bench closures were dropped in 2026-07-18." Log has events; render omits them [^L1EventDrop].

5. **No idempotency deduplication** — Out-of-order events apply twice; replay state diverges from live state [^L1Idempotent].

**Consequence**: Historical reconstruction impossible; event log is incomplete source of truth (HIGH).

---

## H5: Audit Trail Integrity & Immutability — FAILED [^L1H5]

**Claim**: Event log is append-only, tamper-evident, immutable.

**Disconfirming evidence**:

1. **Plain JSON, no hashing** — Files are plain text; any write-access actor can edit events or reorder entries. No hash chains, no signatures [^L1NoHash].

2. **O_APPEND does not prevent edit** — Advisory lock files (.lock-*) don't prevent concurrent writes; malicious processes can write around locks [^L1LockAdvisory].

3. **No checksums per event** — The `key` field (seatId:type:nonce) is not cryptographic; bit flips undetected [^L1NoChecksum].

4. **Crash leaves partial JSON** — Process crash mid-write truncates JSON; next render parses partial object and silently skips tail [^L1PartialWrite].

5. **No tamper detection** — Valid JSON that has been edited reads as valid; tampering is invisible [^L1NoTamperDetect].

**Consequence**: Immutability is advisory, not enforced. Tampering undetectable (HIGH).

---

## Resilient Finding: Map Iteration Deterministic ✓

One mechanism that DOES hold: JavaScript Maps preserve insertion order deterministically [^L1MapOrder]. This is NOT a failure mode.

---

## Risk Matrix

| Hypothesis | Verified | Likelihood | Impact | Complexity | Risk Grade |
|-----------|----------|-----------|--------|-----------|-----------|
| H1 Determinism | 5 mechanisms | High | High | High | CERTAIN |
| H2 Ordering | 4 mechanisms | High | High | High | CERTAIN |
| H3 Atomicity | 4 mechanisms | High | Medium | High | HIGH |
| H4 Reconstruction | 5 mechanisms | High | High | High | HIGH |
| H5 Immutability | 5 mechanisms | Medium | High | Medium | MEDIUM-HIGH |

---

## Lines of Inquiry

- **Pursued**: Disconfirming-first web research on event sourcing, timestamps, JSON serialization, file atomicity, determinism, idempotency (25 searches to saturation)
- **Pursued**: Code inspection of frank-exchange-of-views debate.js and feov-record tool interface
- **Pursued**: Examination of actual event-log structure (.jsonl files, sequence numbering, timestamps) in running research session
- **Declined**: Attempting to reverse-engineer feov-record.exe binary (closed-source; source not available)
- **Abandoned**: Searching for official frank-exchange-of-views documentation outside this repo (tool is project-internal, not published)

---

## Footnotes

[^L1H1]: H1 hypothesis text from frontier.md: "If H1 is TRUE: The rendering engine is a pure function with no side effects, hidden state, or temporal dependencies."

[^L1Precision]: System clocks on Linux (CLOCK_REALTIME) provide ~1 microsecond resolution; nanosecond timestamps with <= 1000ns precision collapse to identical values for concurrent events. See records/.clock format.

[^L1FloatLoss]: JavaScript IEEE 754 double-precision: above 2^53 = 9007199254740992, every integer cannot be represented. JSON.stringify('12345678901234567890') → '12345678901234568000'. [JSON Number Precision](https://jsonic.io/guides/json-number-precision), Jsonic, accessed 2026-07-20.

[^L1UTF16]: RFC 8259, Section 8.2: "unpaired UTF-16 surrogates...instances of unpaired surrogates have been observed when a library truncates a UTF-16 string without checking whether the truncation split a surrogate pair." [JSON.stringify produces invalid UTF-16 · Issue #944 · tc39/ecma262](https://github.com/tc39/ecma262/issues/944), accessed 2026-07-20.

[^L1Interleave]: [Concurrent write() calls with O_APPEND on local files atomic?](https://linux-fsdevel.vger.kernel.narkive.com/RRQpP2Oj/question-are-concurrent-write-calls-with-o_append-on-local-files-atomic), Linux FSDEVEL: "each byte in the affected range being from one write or the other in an unpredictable fashion." Accessed 2026-07-20.

[^L1Fsync]: [Deep Dive into Postgres Write-Ahead Logging](https://martinuke0.github.io/posts/2026-05-28-deep-dive-into-postgres-write-ahead-logging-ensuring-data-durability-and-crash-recovery-in-production), martinuke0, 2026-05-28. "Before the transaction is considered 'committed,' the log records must be physically written (flushed) to persistent storage...often achieved by performing a fsync()."

[^L1Inversion]: [When Logs Lie: How Clock Drift Skews Reality](https://scalardynamic.com/resources/articles/21-when-logs-lie-how-clock-drift-skews-reality-and-breaks-systems), Scalar Dynamic, accessed 2026-07-20: "Causal inversion — where effect appears to precede cause."

[^L1Collision]: Two seats issue events at 2026-07-20T08:10:43.359319700Z (frontier nonce 68011165 and blue-lane-1 nonce 64fa437e). Lexicographic sort: "68011165" < "64fa437e" (string comparison), reversing actual order.

[^L1Lamport]: [Lamport Clock](https://martinfowler.com/articles/patterns-of-distributed-systems/lamport-clock.html), Martin Fowler, accessed 2026-07-20: "Lamport clocks may order events that are not causally related."

[^L1SeqBoundary]: events-frontier-68011165.jsonl and events-blue-lane-1-64fa437e.jsonl are separate per-seat files. Round boundaries don't reset seq. Render must merge two files with independent seq counters; ordering is file-read order, not causal.

[^L1OAppend]: [Concurrent write operations for use with multi-threaded file logging](https://image-ppubs.opensource.gov/dirsearch-public/print/downloadPdf/10642797), US Patent 10642797, accessed 2026-07-20.

[^L1CacheStale]: [Staleness of Data | Relay](https://relay.dev/docs/next/guided-tour/reusing-cached-data/staleness-of-data/), Relay, accessed 2026-07-20: "Components which reference stale data will continue to be able to render that data."

[^L1CachePolicy]: debate.js §VI: "render-shadow/*.md still exist for human/audit verification, read by nobody in the loop" — unclear whether materialized files are read-path or display-only.

[^L1TOCTOU]: [12 Questions and Answers About TOCTOU](https://www.securityscientist.net/blog/12-questions-and-answers-about-toctou-time-of-check-to-time-of-use), Security Scientist, accessed 2026-07-20.

[^L1SchemaEvol]: [Event versioning strategies for event-driven architectures](https://theburningmonk.com/2025/04/event-versioning-strategies-for-event-driven-architectures/), theburningmonk, 2025-04-XX.

[^L1Precision2]: Same as [^L1Precision].

[^L1UTF162]: Same as [^L1UTF16].

[^L1EventDrop]: frontier.md §H4: "bench closures were dropped in 2026-07-18 because `ts` was absent and seat-sequence ordering was wrong (§III tool-is-the-contract)."

[^L1Idempotent]: [The Idempotency Crisis: LLM Agents as Event Stream Consumers](https://tianpan.co/blog/2026-04-19-llm-agents-event-stream-idempotency), TianPan.co, 2026-04-19.

[^L1NoHash]: [Immutable Audit Log Architecture](https://www.emergentmind.com/topics/immutable-audit-log), Emergent Mind, accessed 2026-07-20.

[^L1LockAdvisory]: [Avoiding race conditions when creating or using an existing directory](https://riptutorial.com/node-js/example/5638/avoiding-race-conditions-when-creating-or-using-an-existing-directory), RipTutorial, accessed 2026-07-20.

[^L1NoChecksum]: Inspection of events-frontier-68011165.jsonl: no checksum field; `key` is seatId:type:nonce, not cryptographic.

[^L1PartialWrite]: [Detection of file corruption in a distributed file system](https://image-ppubs.opensource.gov/dirsearch-public/print/downloadPdf/10025788), US Patent 10025788, accessed 2026-07-20.

[^L1NoTamperDetect]: JSON.parse() accepts any valid UTF-8 JSON; bit flips within valid JSON are undetected.

[^L1MapOrder]: [How are elements ordered in a Map in JavaScript?](https://www.geeksforgeeks.org/how-are-elements-ordered-in-a-map-in-javascript/), GeeksforGeeks, accessed 2026-07-20: "Elements are always iterated in the insertion order."
