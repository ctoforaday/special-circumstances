# Cost audit

Measured from 15 per-agent API transcripts in `C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-special-circumstances/6f24a6f4-d0b6-4e8f-b64e-5b1c8c38c346/subagents/workflows/wf_bc10cbd8-9f3`. List-rate arithmetic; see the price table in cost-audit.mjs.

## Per seat-round

| round | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |
|---|---|---|---|---|---|---|---|---|---|
| — | assemble | haiku | 1 | 32 | 0.00M | 0.00M | 0.80M | 0.10M | $0.21 |
| — | blue-lane | haiku | 1 | 63 | 0.01M | 0.01M | 3.07M | 0.16M | $0.58 |
| — | blue-synthesize | haiku | 1 | 81 | 0.00M | 0.01M | 3.15M | 0.17M | $0.61 |
| — | frontier | haiku | 1 | 10 | 0.00M | 0.00M | 0.16M | 0.07M | $0.12 |
| 1 | blue-respond | haiku | 1 | 123 | 0.00M | 0.02M | 8.96M | 0.30M | $1.39 |
| 1 | red-lens | haiku | 3 | 228 | 0.00M | 0.04M | 10.26M | 0.41M | $1.74 |
| 1 | red-merge | haiku | 1 | 81 | 0.00M | 0.01M | 3.56M | 0.29M | $0.79 |
| 2 | blue-respond | haiku | 1 | 121 | 0.00M | 0.03M | 9.12M | 0.30M | $1.42 |
| 2 | judge | haiku | 1 | 44 | 0.00M | 0.01M | 1.29M | 0.14M | $0.35 |
| 2 | red-lens | haiku | 3 | 178 | 0.00M | 0.04M | 7.10M | 0.78M | $1.89 |
| 2 | red-merge | haiku | 1 | 104 | 0.00M | 0.02M | 5.13M | 0.34M | $1.02 |
| | **TOTAL** | | 15 | 1065 | 0.02M | 0.20M | 52.60M | 3.08M | **$10.12** |

## Notes

- Cache traffic is 100% of all tokens; harness panel counters (input+output only) understate real flow accordingly.
- Known physics (runs 3-4 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks the CUMULATIVE ARCHIVE of closed cases (countermeasure: the ledger/archive shard split); judgment-seat premium is cache-RATE-driven (~5x sonnet-intro cache rates at the session tier), not volume-driven; burn is spiky at the judgment seats.

## Tier check

- PASS — every seat ran on its configured tier (bulk: haiku, judgment: haiku).

## Board telemetry (per round)

| round | open | max severity | new mints | mass | realized_open | accepted deltas | mapping |
|---|---|---|---|---|---|---|---|
| 1 | 7 | high | 7 | 39.5 | 0 | 0 | v1 |
| 2 | 4 | medium | 3 | 18 | 0 | 0 | v1 |

Telemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.
