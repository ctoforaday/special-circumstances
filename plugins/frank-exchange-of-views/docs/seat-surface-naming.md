# What naming the surface in the constitution actually buys

Measured 2026-08-15. Instrument: `cmd/seatprobe`, model haiku, 24 dispatches — 4 arms × 2 boards
(`arithmetic`, blue seat; `audit`, merge seat) × 3 replicates. Raw cell reports are not committed;
regenerate with `seatprobe -naming <arm> [-help-directive]`.

## Why this was run

`internal/cli/seat/menu.go` makes the refusal the primary teaching channel on a measured basis:
"seats do not learn this tool from `--help`. Every one of them read it once or twice in twenty to
forty tool calls." The observation is real. The inference is not supported by it, because every one
of those nine sittings ran with a PARTIAL list of verbs already in front of the seat — 2 to 4 names,
counted against the constitutions the probe dispatches under:

| seat | constitution | verbs named | reachable |
|---|---|---|---|
| blue | `blue-researcher.md` | 2 | 18 |
| bench | `lead-judge.md` | 2 | 11 |
| merge | `red-auditor.md` | 4 | 16 |
| lens | `red-auditor.md` | 1 | 9 |

A partial list is a plausible answer to the question `--help` answers completely. "Seats do not read
`--help`" and "seats stop when the partial list runs out" produce the same number and want opposite
fixes.

## The arms

| arm | constitution |
|---|---|
| `none` | shipped text with every verb NAME redacted, situations left standing |
| `partial` | shipped text, unmodified — the condition every prior probe ran under |
| `complete` | shipped text plus the whole role surface, GENERATED from the cobra tree |
| `none+directive` | `none`, plus production's "read `--help` before your first act" clause |

## Result

| arm | n | surface reached (of 17) | range | `--help` reads | tool calls | refusals |
|---|---|---|---|---|---|---|
| `none` | 6 | **8.83** | 6–10 | 0.33 | 26.3 | 3.33 |
| `partial` | 6 | **6.83** | 5–9 | 0.67 | 11.7 | 2.17 |
| `complete` | 6 | **8.33** | 7–11 | 1.17 | 16.7 | 3.00 |
| `none+directive` | 6 | **7.50** | 7–9 | 2.33 | 24.7 | 3.33 |

Welch t on surface reached: `none` − `partial` = +2.00 (t = +2.35); `complete` − `partial` = +1.50
(t = +1.83); `none` − `complete` = +0.50 (t = +0.61).

## What it says

**The status quo is the worst arm.** A seat given the partial list reached fewer verbs than a seat
given no names at all, and did roughly half the work getting there (11.7 tool calls against 26.3).
The partial answer does not merely fail to help — it appears to terminate the search. That is the
mechanism the nine-sitting finding was read as evidence against, and it was never tested.

**Stating the complete surface buys nothing over stating nothing** (t = 0.61). Delivery of the verb
list is not what caps a seat at ~8 of 17.

**The `--help` directive moves the behaviour it targets and not the outcome.** Help reads rose 0.33
→ 2.33 and moved to the first or second tool call — the seat orients rather than recovers — and
surface reach did not follow. Making a seat read the surface does not make it use the surface.

So the discovery explanation is the weak one. Across every arm the ceiling sits near 8 of 17
whether the seat is told nothing, told a little, told everything, or ordered to go and read it.
Whatever caps the reach is not knowledge of what exists.

## The mechanism, read off the trajectories

The obvious reading of the `none` arm — take the names away and the seat falls back on `--help` —
is wrong, and the trajectories say so. The `none` arm read `--help` LESS than any other arm (0.33).
What it did instead, from `none-arithmetic-r1`:

    ERR · ERR · blue show · blue close-gap · ERR · blue manifest-row · blue position ·
    blue friction · blue revision · …

`blue close-gap` does not exist. The refusal answered with the surface — `RefuseUnknownVerb` embeds
cobra's own help — and the seat's next four acts are real verbs it had not used before. It learned
the tool by being turned away from it, never having asked. That is exactly the claim `menu.go`
makes, and this is the first measurement of it against a condition where the names were absent.

The `partial` arm, same board, same model:

    show board · ERR · --help · blue show · blue edit ×5 · blue manifest-row ×4 ·
    blue revision · blue prove ×2 · blue position · blue friction · stop

It read help once, worked the short list, and stopped at 22 calls with 2 refusals. It never reached
past what it had been told, so it never met the channel that teaches.

**The two design decisions are in conflict.** The refusal was built as the primary teaching channel;
the partial naming in the constitution is a bypass around it. A seat that has been handed four verbs
has no occasion to guess a fifth, and guessing is what opens the channel.

## What it does not say

n = 6 per arm, one weak model, two boards, and the ranges overlap on every comparison. The
`partial` deficit is suggestive and not established; it wants replicates before it carries a
decision on its own. The elicitation arm (`seatprobe -ask`) is the instrument for the question this
raises — whether the unreached verbs are unperceived or perceived-and-declined — and it has not
been run against these arms.

Reach is also a proxy. A seat that reaches for 11 verbs has not thereby done better work than one
that reached for 6; `internal/seatprobe` says so at length and it stays true here.
