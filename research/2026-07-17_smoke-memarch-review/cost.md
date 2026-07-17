# Cost audit

Measured from 9 per-agent API transcripts in `C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-AgentOrange/44104cac-198f-40ac-84a1-3f30486755f7/subagents/workflows/wf_5a9b945d-200`. List-rate arithmetic; see the price table in cost-audit.mjs.

## Per seat-round

| round | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |
|---|---|---|---|---|---|---|---|---|---|
| — | assemble | haiku | 1 | 70 | 0.00M | 0.02M | 4.26M | 0.26M | $0.87 |
| — | blue-lane | haiku | 1 | 94 | 0.00M | 0.01M | 5.29M | 0.21M | $0.84 |
| — | blue-synthesize | haiku | 1 | 33 | 0.00M | 0.01M | 1.29M | 0.19M | $0.43 |
| — | frontier | haiku | 1 | 29 | 0.00M | 0.00M | 0.87M | 0.15M | $0.30 |
| 1 | blue-respond | haiku | 1 | 50 | 0.00M | 0.02M | 2.53M | 0.27M | $0.68 |
| 1 | red-lens | haiku | 3 | 113 | 0.00M | 0.01M | 3.56M | 0.26M | $0.74 |
| 1 | red-merge | haiku | 1 | 30 | 0.00M | 0.02M | 1.04M | 0.17M | $0.43 |
| | **TOTAL** | | 9 | 419 | 0.00M | 0.10M | 18.85M | 1.51M | **$4.28** |

## Notes

- Cache traffic is 99% of all tokens; harness panel counters (input+output only) understate real flow accordingly.
- Known physics (runs 3-4 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks the CUMULATIVE ARCHIVE of closed cases (countermeasure: the ledger/archive shard split); judgment-seat premium is cache-RATE-driven (~5x sonnet-intro cache rates at the session tier), not volume-driven; burn is spiky at the judgment seats.

## Board telemetry (per round)

| round | open | max severity | new mints | mass | realized_open | accepted deltas | mapping |
|---|---|---|---|---|---|---|---|
| 1 | 11 | high | 11 | 68.5 | 0 | 0 | v1 |

Telemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.
