# Cost audit

Measured from 25 per-agent API transcripts in `C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-special-circumstances/6f24a6f4-d0b6-4e8f-b64e-5b1c8c38c346/subagents/workflows/wf_908dfd30-baf`. List-rate arithmetic; see the price table in cost-audit.mjs.

## Per seat-round

| round | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |
|---|---|---|---|---|---|---|---|---|---|
| — | assemble | opus | 1 | 58 | 0.00M | 0.02M | 4.07M | 0.31M | $4.51 |
| — | blue-lane | haiku | 3 | 253 | 0.07M | 0.05M | 14.22M | 0.72M | $2.62 |
| — | blue-synthesize | opus | 1 | 71 | 0.00M | 0.07M | 5.56M | 0.44M | $7.24 |
| — | frontier | haiku | 1 | 30 | 0.00M | 0.00M | 0.63M | 0.07M | $0.18 |
| 1 | blue-respond | haiku | 1 | 78 | 0.00M | 0.02M | 5.27M | 0.24M | $0.92 |
| 1 | red-lens | haiku | 4 | 223 | 0.01M | 0.05M | 10.12M | 0.74M | $2.19 |
| 1 | red-merge | opus | 1 | 58 | 0.00M | 0.04M | 5.36M | 0.47M | $6.65 |
| 2 | blue-respond | haiku | 1 | 112 | 0.00M | 0.02M | 8.32M | 0.25M | $1.25 |
| 2 | judge | opus | 1 | 37 | 0.00M | 0.03M | 1.68M | 0.28M | $3.30 |
| 2 | red-lens | haiku | 3 | 241 | 0.01M | 0.04M | 14.45M | 0.67M | $2.51 |
| 2 | red-merge | opus | 1 | 66 | 0.00M | 0.07M | 7.70M | 0.54M | $9.04 |
| 3 | blue-respond | haiku | 1 | 92 | 0.00M | 0.02M | 5.65M | 0.22M | $0.93 |
| 3 | judge | opus | 1 | 37 | 0.00M | 0.02M | 1.83M | 0.20M | $2.60 |
| 3 | red-lens | haiku | 4 | 221 | 0.03M | 0.06M | 12.93M | 0.93M | $2.77 |
| 3 | red-merge | opus | 1 | 67 | 0.00M | 0.07M | 9.10M | 0.49M | $9.40 |
| | **TOTAL** | | 25 | 1644 | 0.13M | 0.57M | 106.87M | 6.59M | **$56.11** |

## Notes

- Cache traffic is 99% of all tokens; harness panel counters (input+output only) understate real flow accordingly.
- Known physics (runs 3-4 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks the CUMULATIVE ARCHIVE of closed cases (countermeasure: the ledger/archive shard split); judgment-seat premium is cache-RATE-driven (~5x sonnet-intro cache rates at the session tier), not volume-driven; burn is spiky at the judgment seats.

## Board telemetry (per round)

| round | open | max severity | new mints | mass | realized_open | accepted deltas | mapping |
|---|---|---|---|---|---|---|---|
| 1 | 10 | medium-high | 10 | 37 | 0 | 0 | v2 |
| 2 | 7 | medium-high | 6 | 25.5 | 0 | 0 | v2 |
| 3 | 3 | medium-high | 2 | 14 | 0 | 3 | v2 |

Telemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.
