# Cost audit

Measured from 58 per-agent API transcripts in `C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-AgentOrange/44104cac-198f-40ac-84a1-3f30486755f7/subagents/workflows/wf_c9dc1f42-7bd`. List-rate arithmetic; see the price table in cost-audit.mjs.

## Per seat-round

| round | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |
|---|---|---|---|---|---|---|---|---|---|
| — | assemble | fable | 1 | 51 | 0.01M | 0.03M | 2.37M | 0.24M | $7.12 |
| — | blue-lane | fable | 3 | 180 | 0.09M | 0.11M | 15.29M | 2.07M | $47.73 |
| — | blue-synthesize | fable | 1 | 31 | 0.00M | 0.05M | 2.48M | 0.53M | $11.78 |
| — | frontier | fable | 1 | 24 | 0.00M | 0.01M | 0.58M | 0.11M | $2.47 |
| 1 | blue-respond | fable | 1 | 114 | 0.00M | 0.09M | 16.55M | 0.68M | $29.41 |
| 1 | red-lens | fable | 6 | 243 | 0.25M | 0.17M | 19.83M | 2.94M | $67.31 |
| 1 | red-merge | fable | 1 | 33 | 0.00M | 0.05M | 2.81M | 0.45M | $10.84 |
| 2 | blue-respond | fable | 1 | 121 | 0.01M | 0.07M | 17.35M | 0.43M | $26.21 |
| 2 | judge | fable | 1 | 24 | 0.00M | 0.02M | 0.76M | 0.47M | $7.54 |
| 2 | red-lens | fable | 3 | 78 | 0.00M | 0.09M | 6.14M | 1.02M | $23.24 |
| 2 | red-lens | opus | 3 | 69 | 0.00M | 0.05M | 4.91M | 0.74M | $8.27 |
| 2 | red-merge | fable | 1 | 35 | 0.00M | 0.06M | 4.11M | 0.41M | $12.26 |
| 3 | blue-respond | fable | 1 | 120 | 0.00M | 0.06M | 19.75M | 0.56M | $29.81 |
| 3 | judge | fable | 1 | 19 | 0.00M | 0.02M | 0.88M | 0.22M | $4.42 |
| 3 | red-lens | fable | 3 | 109 | 0.02M | 0.10M | 10.51M | 1.19M | $30.49 |
| 3 | red-lens | opus | 3 | 104 | 0.04M | 0.13M | 9.08M | 1.68M | $18.48 |
| 3 | red-merge | opus | 1 | 47 | 0.00M | 0.07M | 4.77M | 0.67M | $8.26 |
| 4 | blue-respond | opus | 1 | 147 | 0.00M | 0.05M | 22.66M | 0.57M | $16.15 |
| 4 | judge | opus | 1 | 15 | 0.00M | 0.02M | 0.66M | 0.33M | $2.90 |
| 4 | red-lens | fable | 12 | 369 | 0.24M | 0.37M | 37.94M | 4.96M | $120.71 |
| 4 | red-merge | fable | 1 | 1 | 0.00M | 0.00M | 0.00M | 0.00M | $0.00 |
| 4 | red-merge | opus | 1 | 38 | 0.00M | 0.08M | 3.93M | 0.48M | $6.87 |
| 5 | blue-respond | opus | 1 | 108 | 0.00M | 0.06M | 14.69M | 0.75M | $13.43 |
| 5 | judge | opus | 1 | 33 | 0.00M | 0.02M | 2.10M | 0.49M | $4.52 |
| 5 | red-lens | fable | 2 | 57 | 0.00M | 0.05M | 6.33M | 0.95M | $20.91 |
| 5 | red-lens | opus | 4 | 137 | 0.01M | 0.11M | 17.02M | 1.98M | $23.79 |
| 5 | red-merge | fable | 1 | 57 | 0.00M | 0.06M | 9.70M | 0.67M | $21.15 |
| 5 | red-merge | opus | 1 | 45 | 0.00M | 0.03M | 5.44M | 0.94M | $9.23 |
| | **TOTAL** | | 58 | 2409 | 0.68M | 2.00M | 258.63M | 26.55M | **$585.31** |

## Notes

- Cache traffic is 99% of all tokens; harness panel counters (input+output only) understate real flow accordingly.
- Known physics (runs 3-4 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks the CUMULATIVE ARCHIVE of closed cases (countermeasure: the ledger/archive shard split); judgment-seat premium is cache-RATE-driven (~5x sonnet-intro cache rates at the session tier), not volume-driven; burn is spiky at the judgment seats.

## Board telemetry (per round)

| round | open | max severity | new mints | mass | realized_open | accepted deltas | mapping |
|---|---|---|---|---|---|---|---|
| 1 | 30 | high | 30 | 118.75 | 0 | 0 | v1 |
| 2 | 23 | high | 22 | 81.5 | 0 | 0 | v1 |
| 3 | 17 | medium-high | 17 | 55 | 0 | 0 | v1 |
| 4 | 16 | medium-high | 16 | 46 | 0 | 0 | v1 |
| 5 | 10 | medium | 10 | 30 | 0 | 0 | v1 |

Telemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.
