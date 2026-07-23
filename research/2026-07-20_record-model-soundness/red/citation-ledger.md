# red citation-ledger — Is the event-log-as-source-of-truth record model in frank-exchange-of-views sound, and where are its failure modes?

| Claim | Reference | Confidence | Round | Access Date | Notes |
|-------|-----------|------------|-------|-------------|-------|
| JSON numbers above 2^53 lose precision | Direct JavaScript testing + RFC 8259 | HIGH | 1 | 2026-07-20 | Nanosecond timestamps 1721479843359319700/01 collapse to same value |
| RFC 8259 Section 8.2 documents unpaired UTF-16 surrogates | pubs.opengroup.org/RFC8259 section 8.2 | HIGH | 1 | 2026-07-20 | Direct quote verified; section number accurate |
| POSIX O_APPEND seek-to-end is atomic | pubs.opengroup.org write(2) specification | HIGH | 1 | 2026-07-20 | Spec language clear; concurrent byte-interleave behavior implicit not explicit |
| Collision event (frontier + blue-lane-1 at 2026-07-20T08:10:43.359319700Z) | events-frontier-68011165.jsonl + events-blue-lane-1-64fa437e.jsonl | HIGH | 1 | 2026-07-20 | **DOES NOT EXIST** — timestamps differ by 2:22; no duplicate timestamps in logs |
| Frontier.md §H4 bench-closure quote | blue/frontier.md line 39 | HIGH | 1 | 2026-07-20 | Punctuation differs (comma vs. parenthesis); meaning unchanged |
| Map insertion order is deterministic (JavaScript) | GeeksforGeeks article exists; claim plausible | MEDIUM | 1 | 2026-07-20 | Article accessible; full content not extracted |
| Lamport clock quote (Fowler) | martinfowler.com/articles/.../lamport-clock.html | MEDIUM | 1 | 2026-07-20 | Article exists; specific quote not locatable via grep |
| Concurrent O_APPEND byte interleaving | linux-fsdevel.vger.kernel.narkive.com | LOW | 1 | 2026-07-20 | URL returns 503; underlying claim is correct but unverified at source |
