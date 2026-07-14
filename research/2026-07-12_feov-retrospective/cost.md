# Cost audit — FEOV retrospective (run 3)

Measured from per-agent API transcripts. Pricing assumptions ($/MTok): sonnet 2/10 in/out (intro), cache-read 0.2, cache-write 2.5; session model (judgment seats) 10/50, cache-read 1, cache-write 12.5. List-rate arithmetic — the plan meter observed ~0.6x of these figures.

## Per seat-round

| seat | round | model | agents | api-turns | input | output | cache-read | cache-write | $ |
|---|---|---|---|---|---|---|---|---|---|
| assemble | — | fable | 1 | 31 | 0.00M | 0.02M | 1.55M | 0.22M | $5.57 |
| blue-lane | — | sonnet | 3 | 237 | 0.18M | 0.11M | 16.59M | 1.04M | $7.38 |
| blue-synthesize | — | fable | 1 | 42 | 0.00M | 0.07M | 3.27M | 0.29M | $10.58 |
| other | — | sonnet | 1 | 30 | 0.03M | 0.01M | 0.75M | 0.11M | $0.58 |
| blue-respond | 1 | sonnet | 1 | 112 | 0.02M | 0.05M | 11.57M | 0.43M | $3.95 |
| red-lens | 1 | sonnet | 5 | 284 | 0.04M | 0.13M | 19.84M | 1.58M | $9.28 |
| red-merge | 1 | fable | 1 | 39 | 0.00M | 0.03M | 2.73M | 0.27M | $7.52 |
| blue-respond | 2 | sonnet | 1 | 104 | 0.00M | 0.04M | 12.38M | 0.44M | $3.96 |
| red-lens | 2 | sonnet | 5 | 224 | 0.00M | 0.12M | 19.10M | 1.67M | $9.22 |
| red-merge | 2 | fable | 1 | 51 | 0.00M | 0.04M | 5.64M | 0.44M | $13.22 |
| blue-respond | 3 | sonnet | 1 | 94 | 0.00M | 0.03M | 9.64M | 0.29M | $2.98 |
| red-lens | 3 | sonnet | 5 | 235 | 0.08M | 0.13M | 21.61M | 1.46M | $9.46 |
| red-merge | 3 | fable | 1 | 53 | 0.00M | 0.03M | 5.44M | 0.44M | $12.64 |
| blue-respond | 4 | sonnet | 1 | 91 | 0.00M | 0.03M | 9.56M | 0.33M | $3.05 |
| red-lens | 4 | sonnet | 5 | 233 | 0.09M | 0.14M | 23.49M | 1.66M | $10.47 |
| red-merge | 4 | fable | 1 | 62 | 0.00M | 0.02M | 5.48M | 0.31M | $10.60 |
| blue-respond | 5 | sonnet | 1 | 109 | 0.00M | 0.03M | 14.88M | 0.38M | $4.27 |
| red-lens | 5 | sonnet | 5 | 213 | 0.02M | 0.11M | 24.35M | 2.01M | $11.05 |
| red-merge | 5 | fable | 1 | 61 | 0.00M | 0.03M | 7.87M | 0.34M | $13.56 |
| red-lens | 6 | sonnet | 5 | 18 | 0.00M | 0.00M | 0.26M | 0.22M | $0.61 |
| **TOTAL** | | | 46 | 2323 | 0.47M | 1.20M | 216.01M | 13.93M | **$149.95** |

## Findings

- Cache traffic is 99% of all tokens; the workflow panel counter (input+output only) showed under 2% of real flow.
- Lens cost tracks CORPUS size (rose $1.86 -> $2.21/agent avg r1->r5 while the gap board shrank 20 -> 6); merge cost tracks DISPUTE size (peaked r2, fell after). Two opposite cost physics in one seat color.
- Judgment seats (session model) pay 5x cache-read and 12.5x cache-write rates vs sonnet bulk — the premium is rate-driven, not volume-driven.
- Rounds 1-2 closed 31 gaps ($60-ish); rounds 3-5 closed ~15 mostly-trivial gaps for a similar spend — diminishing returns motivate risk-mass-proportional spend (backlog).
- Stop-and-resume with a reduced maxRounds (cache replay) cost ~$0 and cut ~7 residual rounds; five round-6 lenses were killed mid-spawn for pennies.
