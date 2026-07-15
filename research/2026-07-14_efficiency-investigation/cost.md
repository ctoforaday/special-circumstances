# Cost audit

Measured from 39 per-agent API transcripts in `C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-AgentOrange/05c428b1-bc0d-4ff8-87d9-05c5682b86fc/subagents/workflows/wf_5cefd2a4-35f`. List-rate arithmetic; see the price table in cost-audit.mjs.

## Per seat-round

| round | seat | model | agents | api-turns | input | output | cache-read | cache-write | $ |
|---|---|---|---|---|---|---|---|---|---|
| — | blue-lane | fable | 3 | 195 | 0.03M | 0.12M | 14.32M | 1.03M | $33.36 |
| — | blue-synthesize | fable | 1 | 34 | 0.00M | 0.09M | 2.59M | 0.56M | $13.99 |
| — | frontier | fable | 1 | 20 | 0.00M | 0.00M | 0.54M | 0.14M | $2.52 |
| 1 | blue-respond | fable | 1 | 98 | 0.00M | 0.05M | 9.69M | 0.34M | $16.69 |
| 1 | red-lens | fable | 6 | 255 | 0.04M | 0.15M | 17.32M | 1.70M | $46.62 |
| 1 | red-merge | fable | 1 | 46 | 0.00M | 0.05M | 4.20M | 0.80M | $16.86 |
| 2 | blue-respond | fable | 1 | 126 | 0.00M | 0.04M | 16.85M | 0.39M | $23.72 |
| 2 | judge | fable | 1 | 42 | 0.00M | 0.02M | 1.86M | 0.13M | $4.38 |
| 2 | red-lens | fable | 6 | 247 | 0.00M | 0.15M | 19.79M | 1.70M | $48.33 |
| 2 | red-merge | fable | 1 | 44 | 0.00M | 0.04M | 4.89M | 0.40M | $11.82 |
| 3 | blue-respond | fable | 1 | 97 | 0.00M | 0.04M | 10.87M | 0.29M | $16.53 |
| 3 | judge | fable | 1 | 33 | 0.00M | 0.02M | 1.66M | 0.27M | $5.85 |
| 3 | red-lens | fable | 6 | 176 | 0.00M | 0.13M | 14.89M | 1.64M | $41.89 |
| 3 | red-merge | fable | 1 | 49 | 0.00M | 0.03M | 5.76M | 0.42M | $12.70 |
| 4 | judge | fable | 1 | 11 | 0.00M | 0.00M | 0.22M | 0.21M | $2.98 |
| 4 | red-lens | fable | 6 | 245 | 0.00M | 0.17M | 26.48M | 2.13M | $61.46 |
| 4 | red-merge | fable | 1 | 45 | 0.00M | 0.07M | 5.49M | 0.49M | $14.94 |
| | **TOTAL** | | 39 | 1763 | 0.07M | 1.17M | 157.43M | 12.62M | **$374.63** |

## Notes

- Cache traffic is 99% of all tokens; harness panel counters (input+output only) understate real flow accordingly.
- Known physics (run-3 baseline): lens cost tracks CORPUS size (full re-read x additive growth); merge cost tracks DISPUTE size; judgment-seat premium is cache-RATE-driven, not volume-driven; burn is spiky at the judgment seats.
