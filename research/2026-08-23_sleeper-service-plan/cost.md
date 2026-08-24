# Cost audit

Measured from 32 per-agent API transcripts in `/root/.claude/projects/-home-user-special-circumstances/83b38831-07aa-5901-b6fd-3614fe43a560/subagents/workflows/wf_98d8fa59-8cc`. List-rate arithmetic; see the price table in cost-audit.mjs.

## Per seat-round

| round | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |
|---|---|---|---|---|---|---|---|---|---|
| — | assemble | sonnet | 1 | 56 | 0.00M | 0.00M | 5.58M | 0.49M | $2.36 |
| — | blue-lane | opus | 3 | 222 | 0.00M | 0.02M | 22.67M | 1.46M | $21.00 |
| — | blue-synthesize | sonnet | 1 | 106 | 0.00M | 0.00M | 16.58M | 0.60M | $4.83 |
| — | frontier | opus | 1 | 67 | 0.00M | 0.01M | 4.75M | 0.51M | $5.77 |
| — | judge-terminal | sonnet | 1 | 73 | 0.00M | 0.00M | 7.69M | 0.55M | $2.91 |
| 1 | blue-respond | opus | 1 | 136 | 0.00M | 0.05M | 24.57M | 0.89M | $19.17 |
| 1 | red-lens | opus | 3 | 258 | 0.00M | 0.03M | 28.76M | 1.71M | $25.80 |
| 1 | red-merge | sonnet | 1 | 122 | 0.00M | 0.00M | 16.93M | 1.93M | $8.22 |
| 2 | blue-respond | opus | 1 | 80 | 0.00M | 0.02M | 10.93M | 0.62M | $9.87 |
| 2 | judge | sonnet | 1 | 47 | 0.00M | 0.00M | 3.46M | 0.45M | $1.83 |
| 2 | red-lens | opus | 3 | 151 | 0.00M | 0.03M | 12.75M | 1.26M | $14.91 |
| 2 | red-merge | sonnet | 1 | 85 | 0.00M | 0.01M | 11.81M | 0.55M | $3.84 |
| 3 | blue-respond | opus | 1 | 71 | 0.00M | 0.01M | 8.73M | 1.05M | $11.03 |
| 3 | judge | sonnet | 1 | 69 | 0.00M | 0.00M | 7.31M | 0.52M | $2.77 |
| 3 | red-lens | opus | 4 | 213 | 0.00M | 0.02M | 20.02M | 1.41M | $19.35 |
| 3 | red-merge | sonnet | 1 | 62 | 0.00M | 0.00M | 9.55M | 0.54M | $3.27 |
| 4 | blue-respond | opus | 1 | 113 | 0.00M | 0.00M | 18.42M | 0.75M | $13.94 |
| 4 | judge | sonnet | 1 | 51 | 0.00M | 0.00M | 4.83M | 0.53M | $2.30 |
| 4 | red-lens | opus | 4 | 226 | 0.00M | 0.03M | 21.67M | 1.54M | $21.25 |
| 4 | red-merge | sonnet | 1 | 68 | 0.00M | 0.01M | 8.90M | 0.54M | $3.24 |
| | **TOTAL** | | 32 | 2276 | 0.00M | 0.25M | 265.93M | 17.90M | **$197.67** |

## Notes

- Cache traffic is 100% of all tokens; harness panel counters (input+output only) understate real flow accordingly.
- Known physics (runs 3-4 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks the CUMULATIVE ARCHIVE of closed cases (countermeasure: the ledger/archive shard split); judgment-seat premium is cache-RATE-driven (~5x sonnet-intro cache rates at the session tier), not volume-driven; burn is spiky at the judgment seats.

## Tier check

- **WARN** — red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — blue-lane ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — red-lens ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — blue-respond ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted
- **WARN** — frontier ran on opus, CHEAPER than the configured bulk tier fable — verification may be discounted

## Board telemetry (per round)

| round | open | max severity | new mints | mass | realized_open | mapping |
|---|---|---|---|---|---|---|
| 1 | 9 | high | 9 | 49 | 0 | v2 |
| 2 | 4 | medium-high | 4 | 21 | 0 | v2 |
| 3 | 1 | medium | 5 | 5 | 0 | v2 |
| 4 | 2 | medium-high | 5 | 8 | 0 | v2 |

Telemetry is the convenience copy, never the evidence of record — actuation reviews recompute from the git-tracked ledger.
