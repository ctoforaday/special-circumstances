# Cost audit

Measured from 8 per-agent API transcripts in `C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-special-circumstances/6f24a6f4-d0b6-4e8f-b64e-5b1c8c38c346/subagents/workflows/wf_013be4ed-3bc`. List-rate arithmetic; see the price table in cost-audit.mjs.

## Per seat-round

| round | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |
|---|---|---|---|---|---|---|---|---|---|
| — | assemble | haiku | 1 | 15 | 0.00M | 0.00M | 0.21M | 0.07M | $0.11 |
| — | blue-lane | haiku | 1 | 78 | 0.05M | 0.02M | 4.69M | 0.35M | $1.06 |
| — | blue-synthesize | haiku | 1 | 63 | 0.00M | 0.02M | 3.35M | 0.21M | $0.71 |
| — | frontier | haiku | 1 | 41 | 0.00M | 0.01M | 1.07M | 0.10M | $0.26 |
| 1 | red-lens | haiku | 3 | 193 | 0.00M | 0.03M | 7.34M | 0.42M | $1.44 |
| 1 | red-merge | haiku | 1 | 64 | 0.00M | 0.02M | 3.27M | 0.20M | $0.68 |
| | **TOTAL** | | 8 | 454 | 0.05M | 0.11M | 19.94M | 1.35M | **$4.26** |

## Notes

- Cache traffic is 99% of all tokens; harness panel counters (input+output only) understate real flow accordingly.
- Known physics (runs 3-4 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks the CUMULATIVE ARCHIVE of closed cases (countermeasure: the ledger/archive shard split); judgment-seat premium is cache-RATE-driven (~5x sonnet-intro cache rates at the session tier), not volume-driven; burn is spiky at the judgment seats.

## Board telemetry (per round)

| round | open | max severity | new mints | mass | realized_open | accepted deltas | mapping |
|---|---|---|---|---|---|---|---|
| 1 | 9 | high | 9 | 46.25 | 0 | 0 | v1 |

Telemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.
