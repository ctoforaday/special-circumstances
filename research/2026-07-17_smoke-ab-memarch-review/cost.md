# Cost audit

Measured from 10 per-agent API transcripts in `C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-AgentOrange/44104cac-198f-40ac-84a1-3f30486755f7/subagents/workflows/wf_bfb5cee2-a15`. List-rate arithmetic; see the price table in cost-audit.mjs.

## Per seat-round

| round | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |
|---|---|---|---|---|---|---|---|---|---|
| — | assemble | haiku | 1 | 32 | 0.00M | 0.03M | 0.64M | 0.21M | $0.45 |
| — | blue-lane | haiku | 1 | 59 | 0.00M | 0.01M | 2.71M | 0.19M | $0.56 |
| — | blue-synthesize | haiku | 1 | 34 | 0.00M | 0.01M | 1.10M | 0.20M | $0.42 |
| — | frontier | haiku | 1 | 40 | 0.00M | 0.01M | 1.24M | 0.16M | $0.35 |
| 1 | blue-respond | fable | 1 | 19 | 0.00M | 0.00M | 0.46M | 0.29M | $4.28 |
| 1 | blue-respond | haiku | 1 | 51 | 0.00M | 0.02M | 2.47M | 0.24M | $0.65 |
| 1 | red-lens | haiku | 3 | 94 | 0.00M | 0.03M | 3.22M | 0.44M | $1.03 |
| 1 | red-merge | haiku | 1 | 31 | 0.00M | 0.02M | 1.04M | 0.16M | $0.39 |
| | **TOTAL** | | 10 | 360 | 0.00M | 0.12M | 12.88M | 1.89M | **$8.13** |

## Notes

- Cache traffic is 99% of all tokens; harness panel counters (input+output only) understate real flow accordingly.
- Known physics (runs 3-4 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks the CUMULATIVE ARCHIVE of closed cases (countermeasure: the ledger/archive shard split); judgment-seat premium is cache-RATE-driven (~5x sonnet-intro cache rates at the session tier), not volume-driven; burn is spiky at the judgment seats.

## Board telemetry (per round)

| round | open | max severity | new mints | mass | realized_open | accepted deltas | mapping |
|---|---|---|---|---|---|---|---|
| 1 | 17 | high | 17 | 73.5 | 0 | 0 | v1 |

Telemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.
