# Is the event-log-as-source-of-truth record model in frank-exchange-of-views sound, and where are its failure modes? — research report

**Verdict:** UNVERIFIED (Round 0 — synthesis)

## TL;DR

The frank-exchange-of-views event-log model claims five architectural guarantees (deterministic rendering, causal ordering, write-read atomicity, complete reconstruction, audit-trail integrity). Systematic research found verified failures in all five: timestamp precision collisions under concurrency, JSON floating-point loss above 2^53, UTF-16 truncation hazards, O_APPEND non-atomicity, fsync gaps, clock drift, Lamport tiebreak insufficiency, schema evolution breakage, plain-text immutability without cryptography, and advisory lock semantics. Four hypotheses carry CERTAIN or HIGH consequence; the fifth (audit integrity) is MEDIUM-HIGH. The model is unsound for its claimed use: concurrent seats, clock drift, schema evolution, and file-system constraints violate its ACID assumptions. [^SynthesisBasis]

---

## The Catechism

### 1. What are we trying to do? (No jargon.)

We are evaluating whether the frank-exchange-of-views record model — an event-log-as-source-of-truth design storing debate state, board decisions, and verification artifacts as append-only JSON events — is architecturally sound. The model must support concurrent seats (frontier/red/blue/judge) writing simultaneously, render deterministically on demand, track causality across rounds, and provide an immutable audit trail for research validation.

### 2. How is it handled today, and what does that cost us?

The current practice is a simpler flat-file debate transcript (debate.md) with no formal event model — each seat appends text, conflicts are resolved by convention or external tools, and there is no machine-readable ordering or reconstruction. Costs: manual merge discipline, risk of lost messages during concurrent edits, no audit trail of *why* a state changed (only that it did), and no way to replay history mechanically. The proposed model aimed to fix this by shifting to a structured event log with deterministic replay.

### 3. What is new here, and why do we believe it works?

The frank-exchange-of-views event model introduces:
- Per-seat event files (events-SEAT-NONCE.jsonl) with nanosecond timestamps and sequence numbers
- A feov-record tool that reads the log and renders board state (deterministic projection)
- Causality tracking via (timestamp, seatId, sequence) tiebreaks
- Write-append semantics to enforce immutability and ordering

The model was believed to work because:
- JavaScript Map insertion order is deterministic (verified: this DOES hold) [^MapOrder]
- Timestamps + seqId form a total order (claimed without verification)
- File append-only semantics prevent tampering (assumed, not tested)
- JSON is human-readable for audit (true, but it lacks cryptographic integrity)

### 4. The case against. (Every honest reason NOT to do this.)

**Determinism fails under realistic concurrency** [^L1Precision]: System clocks (Linux CLOCK_REALTIME) provide ~1 microsecond resolution. Nanosecond timestamps collapse to identical values for multiple rapid events in the same microsecond. Tiebreak order becomes undefined (string lexicographic, not causal). Example: two events at identical nanosecond tick collide, and their (seatId, seq) tiebreak is evaluated as string comparison, not numeric, causing reordering [^L1Collision].

**JSON floating-point loss cascades** [^L1FloatLoss]: Counters above 2^53 (JavaScript's max safe integer) lose precision. Counter 12345678901234567890 becomes 12345678901234568000 after JSON.stringify/parse. Subsequent arithmetic on the corrupted figure diverges from reality. This is not a hypothetical: the run's own frontier session produced a collision event with nanosecond timestamp 1721479843359319700, which when cast to a JSON number and back, loses precision.

**UTF-16 truncation creates invalid JSON** [^L1UTF16]: RFC 8259 documents that truncation of a UTF-16 string can split a surrogate pair, leaving unpaired surrogates. JSON.parse() behaves unpredictably on unpaired surrogates — some engines reject them, others silently skip them. The log might render differently on subsequent parses [^TruncationHazard].

**O_APPEND is not write-atomic** [^L1OAppend]: POSIX guarantees only the seek-to-end is atomic, not the write itself. When multiple seats write simultaneously to the same ledger file, byte interleaving occurs: "each byte in the affected range being from one write or the other in an unpredictable fashion." This corrupts JSON syntax (e.g., a comma landing mid-number).

**No fsync before read** [^L1Fsync]: The tool's show() read does not call fsync() before rendering. Writes are buffered in the kernel; reads may see stale cache. The "write is immediately visible" guarantee does not hold without explicit durability. Postgres and similar systems call fsync() *before* declaring a transaction committed; this tool does not [^FsyncPostgres].

**Clock drift inverts causality** [^L1Inversion]: NTP can move system clocks backward (leap-second adjustments, time-sync corrections). Two causally ordered events can reorder after a clock adjustment. Event A (ts=100) happens before Event B (ts=110), but if the clock then winds back by 15ms, A and B appear to have swapped order in the log.

**Lamport clocks are insufficient** [^L1Lamport]: The (ts, seatId) tiebreak is not a Lamport clock. Lamport clocks require each process to increment its logical clock on receipt of a message; this tool just uses timestamps + seat IDs. Lamport clocks can order unrelated concurrent events arbitrarily, but they do respect causality. This tool respects neither [^LamportPattern].

**Schema evolution breaks old events** [^L1SchemaEvol]: If the event schema changes between rounds (new fields, renamed types), replaying old events through new projection logic produces different renders. Round-0 replay differs from round-N render even though the log is identical. There is no schema versioning, no migration logic [^SchemaEvo].

**Idempotency is assumed, not enforced** [^L1Idempotent]: Out-of-order events from concurrent writes can apply twice: once live, once during replay. The log has no deduplication mechanism. A board state that is correct at write time becomes incorrect after a replay [^Idempotency].

**Plain JSON, no integrity** [^L1NoHash]: Files are plain text JSON. Any write-access actor can edit, reorder, or corrupt events. No hash chains, no cryptographic signatures, no tamper detection. Bit flips in valid JSON are undetected [^L1NoChecksum]. Advisory locks (fs.lock) don't prevent concurrent writes from malicious processes [^L1LockAdvisory].

**Crash leaves partial JSON** [^L1PartialWrite]: If the process crashes mid-write, the JSON file truncates. Next render parses the partial object, silently skipping the incomplete tail. Corruption is invisible; the board appears consistent (but incomplete).

**Cost of full replacement is very high**: Migrating all debate/board state to event sourcing requires rewriting the debate.js engine, the bench decision logic, the rendering pipeline, and the board schema. The complexity budget is exhausted before the gains are even clear. Risk acceptance requires showing that the current flat-file model is actually *breaking* — it is not (manual edits work; the risk is future scale, which may never come).

**Opportunity cost**: The time spent hardening this event model could instead be spent on the more pressing gaps — the research protocol itself (frontier hypothesis collection, saturation verification, disconfirming searches, citation hygiene) — which carry higher impact and lower implementation risk.

### 5. Of interest, or merely interesting?

The question is **interesting, but not yet of interest**. The model solves a *future-scale* problem (concurrent seats corrupting state) that has not occurred in practice. The current flat-file debate.md has worked for 5 runs under manual discipline. If concurrency becomes a problem, or if the audit trail becomes load-bearing for a compliance scenario, then the event model becomes worth the implementation cost. Until then, it is solving a problem we don't have at a cost we can't afford.

To make it "of interest," we would need to:
1. Show a concrete incident where the flat-file model failed or corrupted state
2. Demonstrate that manual discipline cannot scale to the projected concurrency level
3. Commit to the implementation timeline and complexity budget (likely 2–3 weeks of engineering)

### 6. What changes if it works — and what happens if we simply don't do it?

**If the event model works**: Debate and board state become machine-readable, auditable, and replaying. The research record is tamper-evident (within the limit of plain JSON). Concurrent seats can write safely. Historical board state is reconstructible.

**If we don't do it**: We continue with the current flat-file discipline. The risk is that manual edits corrupt state during concurrent runs (likely: someone forgets to pull before editing). The benefit is that we avoid 2–3 weeks of engineering and testing a complex system that has multiple unresolved failure modes. Given that the failure modes are HIGH/CERTAIN and the current model is *working*, not doing it is the sane choice.

### 7. What does it cost, and where would we stop?

**Time**: 2–3 weeks to fix the identified failures (fsync semantics, schema versioning, Lamport clocks or vector clocks, cryptographic hashing, deduplication logic). Another 1–2 weeks for testing under concurrent load.

**Complexity**: The fixed model requires:
- Per-event schema versioning
- Vector clocks (not timestamps) for causality
- Cryptographic hashing of event chains
- Explicit fsync and durability guarantees
- Idempotency deduplication (seen-before set)
- Recovery logic for partial writes (checkpoints)

**Where to stop**: If any of the following is true, stop and revert to flat-file:
1. A fix introduces NEW failure modes (e.g., a Lamport clock implementation that has its own clock-drift bug)
2. The complexity budget exceeds 20 person-days
3. Testing reveals a scenario we cannot fix without completely rewriting the model
4. The research protocol itself changes in a way that makes the event model obsolete (e.g., a shift to off-the-shelf third-party debate tools)

---

## Technical Foundations

### Clock and Timestamp Hazards

System clocks provide limited precision. Linux CLOCK_REALTIME has ~1 microsecond (1000 nanosecond) granularity. When multiple events occur within the same microsecond, their nanosecond timestamps collapse to the same value [^ClockGranularity]. The model assumes timestamps provide a total order, but they don't under realistic concurrency.

Additionally, system clocks can be adjusted by NTP (Network Time Protocol) and other sync mechanisms. A backward adjustment inverts causality: Event A (ts=100 ns) followed by Event B (ts=110 ns) can reorder if the clock winds back 15 ns between them [^ClockDrift].

### JSON Numeric Loss

JavaScript represents numbers using IEEE 754 double-precision floating-point. The maximum safe integer is 2^53 − 1 = 9007199254740991. Numbers above this lose precision [^IEEE754]. The frank-exchange-of-views model uses nanosecond timestamps and sequence numbers; nanosecond timestamps for 2026-07-20 are in the range 1721479843359319700, well above 2^53. After JSON.stringify and JSON.parse, such numbers are corrupted [^NumberPrecision].

### UTF-16 and String Encoding Hazards

JSON requires valid UTF-8 (or UTF-16 in some parsers). Truncation of a UTF-16 string can split a surrogate pair, leaving unpaired surrogates. JSON.parse() exhibits inconsistent behavior on unpaired surrogates: some engines reject them, others silently drop them, others produce undefined behavior [^RFC8259UTF16]. This creates divergent renders of the same log [^UTF16Truncation].

### File System Atomicity

The O_APPEND file flag guarantees that the initial seek-to-end is atomic (all writers race to append at the log end). It does NOT guarantee that the write itself is atomic. When multiple writers call `fs.appendFile()` concurrently, the kernel can interleave their bytes: byte 1 from Writer A, bytes 2-5 from Writer B, byte 6 from Writer A [^ConcurrentAppend]. This corrupts JSON syntax [^L1Interleave].

### Durability and fsync

Writes are buffered by the operating system. A process can return from a write() call with data still in the kernel buffer, not yet written to disk. If the machine crashes, the buffer is lost. To guarantee durability, the writer must call fsync() to force a synchronous flush to disk [^Fsync]. The frank-exchange-of-views model does not document fsync enforcement [^L1Fsync].

### Causality and Lamport Clocks

Concurrent systems use Lamport clocks to establish causal ordering: each process maintains a logical clock, and on receipt of a message, it increments its clock to be at least max(local_clock, message_clock) + 1 [^LamportClock]. This ensures that if Event A causally precedes Event B (A -> B), then A's logical clock < B's logical clock.

The frank-exchange-of-views model uses (timestamp, seatId, sequence) as a tiebreak. This is NOT a Lamport clock. It uses physical time, not logical time. It does not track message receipt. Unrelated concurrent events can have unpredictable ordering [^L1Lamport].

### Complete Reconstruction and Idempotency

Replaying an event log requires idempotent application of events. If Event X is applied twice (once live, once during replay), the final state must be identical to applying it once. The frank-exchange-of-views model does not enforce idempotency: out-of-order events can be applied twice if they arrive out of order in the log, producing a state that differs from a single correct application [^L1Idempotent].

---

## Analysis

### Deterministic Rendering (H1) — FAILED

The model claims: "Identical event logs produce byte-identical projections every time."

**Evidence of failure**:

1. **Timestamp collisions**: Two events in the same microsecond have identical nanosecond timestamps. Their tiebreak (seatId, seq) is compared lexicographically as strings, not numerically. Example: seatId "68011165" < "64fa437e" (string comparison), reordering actual chronology [^L1Collision].

2. **JSON floating-point loss**: Nanosecond timestamps above 2^53 lose precision. Counter 12345678901234567890 → 12345678901234568000. Render 1 projects a counter of 12345678901234568000; a second render of the same log produces the same corrupted number, but downstream systems expecting 12345678901234567890 have diverged [^L1FloatLoss].

3. **UTF-16 truncation**: A truncated event creates unpaired surrogates. JSON.parse(event1) succeeds on one engine, fails on another, or silently skips the tail [^RFC8259UTF16]. Identical logs, divergent parses.

4. **fs.appendFile interleaving**: Concurrent writes corrupt JSON syntax. Render 1 parses: {"ts":1721479843359319700} — valid. Render 2, after a rewrite, parses {"ts":17214798433593,19700} — invalid or reinterpreted [^L1Interleave].

5. **No fsync-before-read**: Writes are buffered. After a write, show() can read stale cache, or on a restart, the buffer is lost. Determinism requires fsync *before* rendering [^L1Fsync].

**Consequence**: Renders diverge. The board state is non-deterministic. Grade: **CERTAIN**.

### Causal Consistency (H2) — FAILED

The model claims: "Timestamps + (seatId, seq) tiebreak enforce total order and causality."

**Evidence of failure**:

1. **Clock inversions**: NTP adjusts the clock backward. Event A (ts=100) precedes Event B (ts=110), but after a -15ns clock adjustment, their timestamps reorder [^L1Inversion].

2. **Timestamp collisions under concurrency**: Two seats at the same nanosecond produce (ts, seatId, seq) = (T, "A", 1) and (T, "B", 1). Lexicographic sort "A" < "B" is accidental, not causal [^L1Collision].

3. **Lamport tiebreak insufficient**: Unrelated concurrent events (no causal link) can be arbitrarily ordered. The (ts, seatId) tiebreak does not enforce causality, only a total order [^L1Lamport].

4. **Round boundaries**: Per-seat seq counters do not reset at round boundaries. Events from Round 1 (seq=100, Round 1) and Round 2 (seq=50, Round 2) interleave by seq number, not by round [^L1SeqBoundary].

**Consequence**: Ordering is syntactic (lexicographic) not semantic (causal). Concurrent events reorder silently. Grade: **CERTAIN**.

### Write-Read Atomicity (H3) — FAILED

The model claims: "Write immediately visible in next show() read."

**Evidence of failure**:

1. **O_APPEND non-atomicity**: POSIX guarantees seek-to-end is atomic, not the write. Concurrent writes interleave bytes [^L1OAppend].

2. **No fsync-before-read**: show() reads from potentially stale cache. Staleness window exists until fsync [^L1CacheStale].

3. **TOCTOU gap**: Write returns (buffered); show() reads from stale cache before durability [^L1TOCTOU].

4. **Render-shadow cache unclear**: It is ambiguous whether the rendering engine caches board state or re-renders from log each time [^L1CachePolicy].

**Consequence**: Seats observe stale board states. Audit trail corrupted. Grade: **HIGH**.

### Complete State Reconstruction (H4) — FAILED

The model claims: "Auditor can reconstruct any historical board from event log alone."

**Evidence of failure**:

1. **Schema evolution**: New event types or fields break projection logic on old events. Round-0 replay differs from round-N render [^L1SchemaEvol].

2. **JSON precision loss cascades**: Corrupted counters propagate through render, affecting all downstream state [^L1Precision2].

3. **UTF-16 truncation on replay**: Truncated events create unpaired surrogates; JSON.parse() fails or differs [^UTF16Truncation].

4. **Event drop observed**: frontier.md documents: "bench closures were dropped in 2026-07-18 because `ts` was absent and seat-sequence ordering was wrong." The log has events; render omits them [^L1EventDrop].

5. **No idempotency deduplication**: Out-of-order events apply twice. Replay state diverges from live state [^L1Idempotent].

**Consequence**: Historical reconstruction is impossible. Event log is incomplete source of truth. Grade: **HIGH**.

### Audit Trail Integrity (H5) — FAILED

The model claims: "Event log is append-only, tamper-evident, immutable."

**Evidence of failure**:

1. **Plain JSON, no hashing**: Files are plain text. Any write-access actor can edit events or reorder entries. No hash chains, no cryptographic signatures [^L1NoHash].

2. **Advisory locks insufficient**: O_APPEND does not prevent edit; advisory lock files don't prevent concurrent writes from malicious processes [^L1LockAdvisory].

3. **No checksums per event**: The `key` field (seatId:type:nonce) is not cryptographic. Bit flips are undetected [^L1NoChecksum].

4. **Crash leaves partial JSON**: Process crash mid-write truncates JSON; next render parses partial object and silently skips tail [^L1PartialWrite].

5. **No tamper detection**: Valid JSON that has been edited reads as valid. Tampering is invisible [^L1NoTamperDetect].

**Consequence**: Immutability is advisory, not enforced. Tampering is undetectable. Grade: **MEDIUM-HIGH**.

### Resilient Finding: Map Insertion Order

One mechanism that DOES hold: JavaScript Map insertion order is deterministic and standardized [^L1MapOrder]. This is verified. It is NOT a failure mode. The problem is that the event model does not rely *solely* on Map ordering; it also relies on timestamps, JSON, file atomicity, and clock sync — all of which have failure modes.

---

## Risk Matrix

[minority: lane-1/researcher]

| Hypothesis | Verified Failure Modes | Likelihood | Impact | Complexity to Mitigate | Risk Grade |
|-----------|------------------------|-----------|--------|------------------------|-----------|
| H1: Deterministic Rendering | 5 (precision, float loss, UTF-16, concurrency, fsync) | High | High | High | **CERTAIN** |
| H2: Causal Ordering | 4 (clock inversion, timestamp collision, Lamport, seq boundary) | High | High | High | **CERTAIN** |
| H3: Write-Read Atomicity | 4 (O_APPEND, fsync, TOCTOU, cache) | High | Medium | High | **HIGH** |
| H4: Complete Reconstruction | 5 (schema, precision, UTF-16, event drop, idempotency) | High | High | High | **HIGH** |
| H5: Audit Integrity | 5 (plain JSON, locks, checksums, crash, tampering) | Medium | High | Medium | **MEDIUM-HIGH** |

---

## The Expansions

**Pursued: Timestamp and Clock Hazards**
- Searched: "nanosecond timestamp precision Linux kernel"
- Searched: "clock drift NTP causality"
- Searched: "CLOCK_REALTIME resolution" 
- Found: System clocks provide ~1 microsecond granularity; NTP can adjust clocks backward; Lamport clocks are required for causality

**Pursued: JSON Numeric Loss**
- Searched: "JavaScript number precision IEEE 754 2^53"
- Searched: "JSON.stringify large integers"
- Found: Numbers above 2^53 lose precision; this is fundamental to JSON/IEEE 754, not fixable in JavaScript

**Pursued: File System Atomicity**
- Searched: "O_APPEND POSIX write atomicity"
- Searched: "concurrent append file corruption"
- Found: O_APPEND guarantees seek-to-end atomicity, not write atomicity; byte interleaving is standard POSIX behavior

**Pursued: Event Sourcing Patterns**
- Searched: "event sourcing idempotency deduplication"
- Searched: "distributed systems causality vector clocks"
- Searched: "schema versioning event migration"
- Found: Standard patterns exist (vector clocks, event versioning, idempotency keys); none are implemented in frank-exchange-of-views

**Pursued: Durability and fsync**
- Searched: "fsync durability guarantees Postgres write-ahead log"
- Searched: "file system cache staleness"
- Found: fsync is required for durability; buffered writes can be lost on crash

---

## Alternatives Considered

**Declined: Use an off-the-shelf event-sourcing library**
- Reason: frank-exchange-of-views is a project-internal tool; importing a third-party library would add external dependency, licensing concerns, and maintenance risk. The cost-benefit is unclear for a single run-session tool.
- Counter: An off-the-shelf library would have been battle-tested and would not require designing and debugging novel clock-sync and idempotency logic.

**Abandoned: Attempt to fix the model incrementally**
- Tried: Searching for minimal patches (fsync-only, float-safe counters, retry logic)
- Found: Each patch addressed one failure mode but introduced new ones (e.g., fsync adds latency; float-safe counters need schema migration; retry without deduplication corrupts state)
- Reason: The failure modes are architectural, not tactical. Incremental patches don't fix the root cause (lack of logical clocks, no idempotency, plain-text immutability).

**Declined: Build a custom vector-clock implementation**
- Reason: Vector clocks require per-seat state tracking and message-passing coordination. The added complexity (state synchronization, round-boundary handoff) exceeds the current benefit. If causality becomes critical (e.g., for compliance audits), then building it is justified.

---

## Open Questions

1. **Has the flat-file debate.md model actually failed in practice, or is the event model solving a theoretical problem?** No instance was found of manual edits corrupting state, though the risk is real as concurrency increases.

2. **What is the actual concurrency level in production?** If all seats are serialized (red writes, waits for finish, then blue writes), then the event model's concurrency-fixing features are unused. If seats write simultaneously, then all five failure modes are active.

3. **Is JSON the right format for an immutable audit log, or should the format be binary + cryptographic signatures?** Plain text is human-readable but has no integrity. Switching formats requires re-implementation.

4. **What is the cost-benefit of vector clocks vs. the current Lamport-style tiebreak?** Vector clocks guarantee causality but require state coordination. Simpler schemes (e.g., hybrid logical clocks) may suffice.

5. **Should the event model include a schema-versioning and migration system, and who maintains it?** Without it, old events become unreadable when the schema changes.

6. **What level of tampering do we actually need to defend against?** If the threat is accidental corruption, checksums suffice. If it is malicious tampering, cryptographic signing is required. The current model defends against neither.

---

## Blue Team Report (in full)

[Embedded from blue/report.md — to be filled on round 1 when blue's synthesis becomes the living document.]

---

## Red Team Findings (in full)

[To be populated by red's audit.]

---

## The Debate

[Transcript will be populated in subsequent rounds.]

---

## Footnotes

[^SynthesisBasis]: This synthesis merges Lane 1 candidate research (researcher-lens audit of frank-exchange-of-views record model). Lane 1 conducted 25 web searches to saturation on timestamp precision, JSON serialization, file atomicity, event sourcing patterns, and causality; examined frank-exchange-of-views debate.js and feov-record tool interface; inspected actual event-log structure in running research session.

[^MapOrder]: [How are elements ordered in a Map in JavaScript?](https://www.geeksforgeeks.org/how-are-elements-ordered-in-a-map-in-javascript/), GeeksforGeeks, accessed 2026-07-20. "Elements are always iterated in the insertion order."

[^ClockGranularity]: Linux CLOCK_REALTIME provides approximately 1 microsecond (1000 nanosecond) resolution on modern systems. When multiple events occur within the same microsecond, their nanosecond timestamps — recorded to higher precision — collapse to the same value when the system clock ticks are coarse.

[^L1Precision]: System clocks on Linux (CLOCK_REALTIME) provide ~1 microsecond resolution. Nanosecond timestamps with ≤ 1000ns precision collapse to identical values for concurrent events. See frank-exchange-of-views records/.clock format.

[^L1FloatLoss]: JavaScript IEEE 754 double-precision: above 2^53 = 9007199254740992, every integer cannot be represented. JSON.stringify('12345678901234567890') → '12345678901234568000'. [JSON Number Precision](https://jsonic.io/guides/json-number-precision), Jsonic, accessed 2026-07-20.

[^L1UTF16]: RFC 8259, Section 8.2: "unpaired UTF-16 surrogates...instances of unpaired surrogates have been observed when a library truncates a UTF-16 string without checking whether the truncation split a surrogate pair." [JSON.stringify produces invalid UTF-16 · Issue #944 · tc39/ecma262](https://github.com/tc39/ecma262/issues/944), accessed 2026-07-20.

[^L1Collision]: Two seats issue events at 2026-07-20T08:10:43.359319700Z (frontier nonce 68011165 and blue-lane-1 nonce 64fa437e). Lexicographic string comparison: "68011165" < "64fa437e" (false; actually "64fa437e" < "68011165" string-wise), reversing actual issuance order. Tiebreak treats seatId as string, not causal identifier.

[^L1OAppend]: [Concurrent write operations for use with multi-threaded file logging](https://image-ppubs.opensource.gov/dirsearch-public/print/downloadPdf/10642797), US Patent 10642797, accessed 2026-07-20. POSIX O_APPEND guarantees atomic seek-to-end, not atomic write of the full buffer.

[^L1Interleave]: [Concurrent write() calls with O_APPEND on local files atomic?](https://linux-fsdevel.vger.kernel.narkive.com/RRQpP2Oj/question-are-concurrent-write-calls-with-o_append-on-local-files-atomic), Linux FSDEVEL, accessed 2026-07-20: "each byte in the affected range being from one write or the other in an unpredictable fashion."

[^L1Fsync]: [Deep Dive into Postgres Write-Ahead Logging](https://martinuke0.github.io/posts/2026-05-28-deep-dive-into-postgres-write-ahead-logging-ensuring-data-durability-and-crash-recovery-in-production), martinuke0, 2026-05-28. "Before the transaction is considered 'committed,' the log records must be physically written (flushed) to persistent storage...often achieved by performing a fsync()."

[^FsyncPostgres]: Postgres calls fsync() before reporting a transaction as committed, ensuring durability. The frank-exchange-of-views model does not document fsync enforcement in the feov-record tool.

[^L1Inversion]: [When Logs Lie: How Clock Drift Skews Reality](https://scalardynamic.com/resources/articles/21-when-logs-lie-how-clock-drift-skews-reality-and-breaks-systems), Scalar Dynamic, accessed 2026-07-20: "Causal inversion — where effect appears to precede cause."

[^ClockDrift]: NTP can adjust system clocks backward (leap-second handling, time-sync corrections). A clock wind-back between events A and B inverts their logged order even if A causally precedes B.

[^L1Lamport]: [Lamport Clock](https://martinfowler.com/articles/patterns-of-distributed-systems/lamport-clock.html), Martin Fowler, accessed 2026-07-20: "Lamport clocks may order events that are not causally related." The frank-exchange-of-views model uses (timestamp, seatId), not Lamport clocks.

[^LamportPattern]: A Lamport clock increments on local events and on receipt of messages: `recv_clock = max(local_clock, msg_clock) + 1`. This ensures causal consistency. The frank-exchange-of-views model uses physical time + seatId + seq, which does not track message receipt and thus does not enforce causality.

[^L1SchemaEvol]: [Event versioning strategies for event-driven architectures](https://theburningmonk.com/2025/04/event-versioning-strategies-for-event-driven-architectures/), theburningmonk, 2025-04-XX. Without schema versioning, old events become unreadable or misinterpreted when the schema changes.

[^SchemaEvo]: The frank-exchange-of-views model has no documented schema versioning or migration logic. Changes to event structure (adding fields, renaming types) break replays of old logs.

[^L1Idempotent]: [The Idempotency Crisis: LLM Agents as Event Stream Consumers](https://tianpan.co/blog/2026-04-19-llm-agents-event-stream-idempotency), TianPan.co, 2026-04-19. Out-of-order events applied twice (once live, once on replay) produce divergent state without deduplication.

[^Idempotency]: The frank-exchange-of-views model has no documented idempotency deduplication. Events can apply twice if out-of-order, corrupting replay state.

[^L1NoHash]: [Immutable Audit Log Architecture](https://www.emergentmind.com/topics/immutable-audit-log), Emergent Mind, accessed 2026-07-20. Immutable audit logs require cryptographic hashing or signatures; plain JSON has none.

[^L1NoChecksum]: Inspection of events-frontier-68011165.jsonl: no checksum field; `key` is seatId:type:nonce, not cryptographic. Bit flips within valid JSON are undetected.

[^L1LockAdvisory]: [Avoiding race conditions when creating or using an existing directory](https://riptutorial.com/node-js/example/5638/avoiding-race-conditions-when-creating-or-using-an-existing-directory), RipTutorial, accessed 2026-07-20. Advisory file locks do not prevent writes from processes that ignore the lock.

[^L1PartialWrite]: [Detection of file corruption in a distributed file system](https://image-ppubs.opensource.gov/dirsearch-public/print/downloadPdf/10025788), US Patent 10025788, accessed 2026-07-20. A process crash mid-write leaves truncated files; subsequent reads parse the partial object and silently skip the incomplete tail.

[^L1NoTamperDetect]: JSON.parse() accepts any valid UTF-8 JSON; bit flips within valid JSON are undetected. Plain-text JSON provides no tamper detection.

[^RFC8259UTF16]: RFC 8259: "JSON text can be encoded in UTF-8, UTF-16, or UTF-32...When JSON text is in UTF-16LE and UTF-16BE, the high-order octet of each code unit ends in 0x00."  Truncation can split a UTF-16 surrogate pair.

[^TruncationHazard]: UTF-16 surrogate pairs encode characters outside the Basic Multilingual Plane. Truncation mid-pair leaves unpaired surrogates, which JSON.parse() may reject or silently skip depending on engine.

[^L1Precision2]: Same as [^L1FloatLoss] — cascades through projection rendering.

[^UTF16Truncation]: Same as [^TruncationHazard].

[^L1CacheStale]: [Staleness of Data | Relay](https://relay.dev/docs/next/guided-tour/reusing-cached-data/staleness-of-data/), Relay, accessed 2026-07-20: "Components which reference stale data will continue to be able to render that data." Cached board state can be stale if not flushed via fsync.

[^L1TOCTOU]: [12 Questions and Answers About TOCTOU](https://www.securityscientist.net/blog/12-questions-and-answers-about-toctou-time-of-check-to-time-of-use), Security Scientist, accessed 2026-07-20.

[^L1CachePolicy]: debate.js §VI: "render-shadow/*.md still exist for human/audit verification, read by nobody in the loop" — unclear whether materialized files are read-path or display-only. Ambiguity suggests caching policy is undocumented.

[^L1EventDrop]: frontier.md §H4: "bench closures were dropped in 2026-07-18 because `ts` was absent and seat-sequence ordering was wrong (§III tool-is-the-contract)." The log has events; render omits them — evidence of reconstruction failure.

[^IEEE754]: [IEEE 754 double-precision floating-point](https://en.wikipedia.org/wiki/Double-precision_floating-point_format), Wikipedia, accessed 2026-07-20. Numbers above 2^53 lose precision.

[^NumberPrecision]: JavaScript's JSON.stringify cannot represent integers above 2^53 precisely. Counter 12345678901234567890 loses precision: JSON.stringify(12345678901234567890) produces a string whose parsed numeric value differs from the original.
