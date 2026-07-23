# Lines of Inquiry — RENDERED PROJECTION (source of truth: records/ event log)

## pursued (8)

- **Web search on event sourcing failures, determinism, and timestamps** _(adversarial-disconfirming-first)_ (blue-lane-1)
- **Code inspection of feov-record tool interface and event log structure** _(local-repo critical-stance)_ (blue-lane-1)
- **Timestamp and clock precision hazards** _(disconfirming web research)_ — Searches to saturation revealed nanosecond precision cannot be preserved by microsecond-granularity system clocks; collisions under concurrency are guaranteed (blue-synthesize)
- **JSON numeric loss and IEEE 754 precision** _(disconfirming web research)_ — Numbers above 2^53 lose precision; this is fundamental to JSON and IEEE 754, not fixable; frank-exchange-of-views nanosecond timestamps exceed this threshold (blue-synthesize)
- **File system atomicity: O_APPEND semantics** _(disconfirming web research)_ — O_APPEND guarantees seek-to-end atomicity, not write atomicity; byte interleaving is standard POSIX behavior under concurrent writes (blue-synthesize)
- **Event-sourcing patterns and causality** _(disconfirming web research)_ — Standard patterns exist (vector clocks, event versioning, idempotency dedup); none implemented in frank-exchange-of-views; model uses timestamps instead of Lamport clocks (blue-synthesize)
- **Code inspection: frank-exchange-of-views debate.js and feov-record tool** _(direct source examination)_ — Examined debate.js routing and feov-record tool interface; confirmed no fsync, no idempotency dedup, no schema versioning (blue-synthesize)
- **Event-log structure in running sessions** _(direct artifact inspection)_ — Examined events-*.jsonl files in active research session; confirmed collision events with identical nanosecond timestamps and absence of fsync markers (blue-synthesize)

## abandoned (2)

- **Searching for official frank-exchange-of-views documentation outside the project** — Tool is project-internal; not published to web; external search yields irrelevant results (unrelated Cultural references) (blue-lane-1)
- **Incremental patches to fix model failures** _(architectural analysis)_ — Each patch (fsync-only, float-safe counters, retry logic) addressed one failure mode but introduced new ones; failure modes are architectural, not tactical (blue-synthesize)

## declined (4)

- **Reverse-engineering feov-record.exe binary for implementation details** — Binary is closed-source; source not available in repo; source inspection would be incomplete (blue-lane-1)
- **Reverse-engineer feov-record.exe binary** — Tool is closed-source; source not available; cost of reverse-engineering exceeds benefit for this round (blue-synthesize)
- **Search for external frank-exchange-of-views documentation** — frank-exchange-of-views is a project-internal tool; no published documentation exists; repo and in-session inspection are the sole authoritative sources (blue-synthesize)
- **Build custom vector-clock implementation** — Vector clocks require per-seat state tracking and message-passing coordination; added complexity exceeds current benefit; justified only if causality becomes critical (blue-synthesize)

