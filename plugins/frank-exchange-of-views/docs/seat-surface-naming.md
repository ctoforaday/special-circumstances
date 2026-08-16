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

## A third channel was uncontrolled, and it co-varied with the arm

**Read this before the section below.** `record.SittingOf` derives a `Duty{What, How}` for each live
circumstance — the situation, its consequence, and the exact command that discharges it. It carries
the same fact the constitution's naming carries, arriving only when it applies, and it rides on ONE
projection: `worklist`. `show board` does not carry it.

That channel was never varied, never measured, and is not a constant:

| arm | `worklist` reads/cell | cells with ≥1 | `board` reads/cell | reach |
|---|---|---|---|---|
| `none` | 2.00 | 5/6 | 4.00 | 8.83 |
| `partial` | 0.67 | 3/6 | 3.17 | 6.83 |
| `complete` | 0.33 | 2/6 | 4.33 | 8.33 |
| `none+directive` | 1.83 | 4/6 | 2.67 | 7.50 |

A covariate that moves threefold with the treatment is not an oversight to be waved off — it is a
rival explanation wearing the result's clothes. **The `none` over `partial` advantage cannot be
attributed to the constitution's naming.**

It is not a clean mediator either, and the opposite over-claim is refuted by the same table:
`complete` had the FEWEST worklist reads and the second-highest reach, so "reading the worklist
raises reach" does not hold.

`ReadViewReads` now prints this beside `HelpUse` on every probe report, so no future run states a
naming effect without the channel that competes with it.

**The design finding here is independent of the experiment.** `board` is described in this tool's
own words as "the form a seat acts on" and is read 2.7–4.3 times a sitting. `worklist` carries every
duty and is read 0.33–2.00 times. The one channel that delivers situation-plus-verb at the moment it
applies is the one the tool steers seats away from — and `SittingOf` can name only 4 of blue's 17
verbs, 5 of merge's 17, 3 of bench's 14, 2 of lens's 10, because a duty is derived only where
omission already carries a mechanical consequence (a refusal, or a capture score). Every verb whose
omission is merely a quality loss — `line-of-inquiry`, `manifest-row`, `spot-check`, `verify`, `reproduce`,
`closing`, `regrade` — gets no line, and those are the verbs the probe boards were built to bait.

## What it says

**The status quo is the worst arm — CONFOUNDED, see above.** A seat given the partial list reached
fewer verbs than a seat given no names at all, and did roughly half the work getting there (11.7
tool calls against 26.3). The partial answer does not merely fail to help; it appears to terminate
the search. But those same cells read the duty-carrying projection a third as often, so the effect
is not separable from the constitution's naming on this data. What is established is that the
shipped condition came LAST on reach; why it did is open.

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
cobra's own help — and the seat's next four acts are real verbs it had not used before.

**The attribution to the refusal is withdrawn.** Two of those four — `friction` and `revision` — are
exactly the verbs `SittingOf` names, and this seat had already called `blue show` before it guessed.
The refusal and the duty list are both live at that moment, and the trace cannot separate them. What
the trace does establish is narrower and still worth having: the seat never asked for `--help`, so
whatever taught it was not the channel the constitution tells it to use.

The `partial` arm, same board, same model:

    show board · ERR · --help · blue show · blue edit ×5 · blue manifest-row ×4 ·
    blue revision · blue prove ×2 · blue position · blue friction · stop

It read help once, worked the short list, and stopped at 22 calls with 2 refusals. It never reached
past what it had been told, so it never met the channel that teaches.

**The two design decisions are in conflict, and that much does not depend on the attribution.** The
refusal was built as the primary teaching channel; the partial naming in the constitution is a
bypass around it. A seat that has been handed four verbs has no occasion to guess a fifth, and
guessing is what opens the channel. The `partial` arm's 2.17 refusals against `none`'s 3.33 is the
same story in the one number the confound does not touch.

## What it does not say

n = 6 per arm, one weak model, two boards, and the ranges overlap on every comparison. The
`partial` deficit is suggestive, confounded by the duty channel above, and not established; it
wants replicates AND a controlled duty channel before it carries a decision on its own. The next
run varies the duty lines and holds the naming fixed, which is the inverse of this one: this matrix
varied the channel seats barely read and held fixed the channel they do. The elicitation arm (`seatprobe -ask`) is the instrument for the question this
raises — whether the unreached verbs are unperceived or perceived-and-declined — and it has not
been run against these arms.

Reach is also a proxy. A seat that reaches for 11 verbs has not thereby done better work than one
that reached for 6; `internal/seatprobe` says so at length and it stays true here.
