# Cost audit

Measured from 29 per-agent API transcripts in `/root/.claude/projects/-home-user-special-circumstances/83b38831-07aa-5901-b6fd-3614fe43a560/subagents/workflows/wf_5ca42dd4-f10`. List-rate arithmetic; see the price table in cost-audit.mjs.

## Per seat-round

| round | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |
|---|---|---|---|---|---|---|---|---|---|
| — | assemble | sonnet | 1 | 47 | 0.00M | 0.00M | 4.08M | 0.44M | $1.92 |
| — | blue-lane | opus | 3 | 227 | 0.00M | 0.03M | 25.36M | 1.82M | $24.89 |
| — | blue-synthesize | sonnet | 1 | 201 | 0.00M | 0.00M | 35.86M | 1.68M | $11.40 |
| — | frontier | opus | 1 | 58 | 0.00M | 0.01M | 4.20M | 0.33M | $4.29 |
| — | judge-terminal | sonnet | 1 | 70 | 0.00M | 0.00M | 6.77M | 0.40M | $2.39 |
| 1 | blue-respond | opus | 1 | 89 | 0.00M | 0.01M | 10.90M | 0.93M | $11.57 |
| 1 | red-lens | opus | 3 | 146 | 0.00M | 0.02M | 12.43M | 1.04M | $13.33 |
| 1 | red-merge | sonnet | 1 | 104 | 0.00M | 0.01M | 13.21M | 0.61M | $4.25 |
| 2 | blue-respond | opus | 1 | 76 | 0.00M | 0.04M | 10.91M | 1.41M | $15.33 |
| 2 | red-lens | opus | 3 | 172 | 0.00M | 0.02M | 16.06M | 1.04M | $14.95 |
| 2 | red-merge | sonnet | 1 | 92 | 0.00M | 0.00M | 11.40M | 0.66M | $3.93 |
| 3 | blue-respond | opus | 1 | 103 | 0.00M | 0.01M | 14.02M | 1.00M | $13.46 |
| 3 | red-lens | opus | 4 | 215 | 0.00M | 0.05M | 20.47M | 1.48M | $20.62 |
| 3 | red-merge | sonnet | 1 | 75 | 0.00M | 0.00M | 10.36M | 0.56M | $3.49 |
| 4 | blue-respond | opus | 1 | 87 | 0.00M | 0.01M | 12.43M | 1.04M | $12.86 |
| 4 | red-lens | opus | 4 | 208 | 0.00M | 0.01M | 19.13M | 1.46M | $19.02 |
| 4 | red-merge | sonnet | 1 | 71 | 0.00M | 0.01M | 9.87M | 0.54M | $3.47 |
| | **TOTAL** | | 29 | 2041 | 0.00M | 0.24M | 237.48M | 16.44M | **$181.18** |

## Notes

- Cache traffic is 100% of all tokens; harness panel counters (input+output only) understate real flow accordingly.
- Known physics (runs 3-4 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks the CUMULATIVE ARCHIVE of closed cases (countermeasure: the ledger/archive shard split); judgment-seat premium is cache-RATE-driven (~5x sonnet-intro cache rates at the session tier), not volume-driven; burn is spiky at the judgment seats.

## Tier check

- **WARN** — red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — blue-lane ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — frontier ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted

## Board telemetry (per round)

| round | open | max severity | new mints | mass | realized_open | mapping |
|---|---|---|---|---|---|---|
| 1 | 6 | medium | 6 | 22.5 | 0 | v2 |
| 2 | 4 | medium | 4 | 13.75 | 0 | v2 |
| 3 | 3 | medium | 3 | 11 | 0 | v2 |
| 4 | 0 | ? | 3 | 0 | 0 | v2 |

Telemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.
