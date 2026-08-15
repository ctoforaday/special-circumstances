# What the duty channel buys, and what the first attempt to measure it found instead

Measured 2026-08-15. Instrument: `cmd/seatprobe -duty <arm>`, naming held at `partial` (the shipped
constitution), model haiku, 24 dispatches — 4 arms × 2 boards (`docket`/blue, `lens-audit`/lens) × 3
replicates.

## Why this run exists

The naming matrix (`seat-surface-naming.md`) varied the channel seats barely read — the
constitution — and held fixed the one they do: the board. `record.SittingOf` carries a
`Duty{What, How}` for each live circumstance, which is the same fact the constitution's naming
carries, arriving only when it applies. This run inverts the first: vary the board, hold the
constitution.

## The first attempt was void, and the check caught it

The matrix came back a flat null — all four arms within 0.5 verbs, every t under 1.0. The delivery
check written after the naming run's confound showed why: **the `available` list was empty in all 24
dispatches.**

`roleOf` was `cmd.Parent().Name()`, which is the role only while every verb's parent IS its role.
`show` became a GROUP, so the parent of `blue show worklist` is `show`, and every projection read
since had reported its role as the string `"show"`. `SittingOf` switches on that role; a switch that
matches nothing falls through in silence. **The only duty any seat had received since that refactor
was the friction line, which sits above the switch.**

Blue was never told about a computation gap it had not proved or a round record it had not filed.
The merge was never told about an open gap or an unruled motion. The bench was never told about an
unruled petition. And because `complete` is `len(Outstanding) == 0`, a seat that filed friction was
told `complete: true` while `verdict --as PASS` went on refusing it over gaps the same view had just
declined to mention — the exact disagreement `sitting.go`'s own doctrine forbids.

A null from a treatment that never happened is what that check exists to refuse. It refused it.

## The arms, on a tree where they can differ

| arm | what the board tells the seat |
|---|---|
| `off` | nothing — no duties, no affordances |
| `shipped` | the enforced duties only (what the tool sends) |
| `available` | plus what the board affords, on `worklist` |
| `available+board` | and carried on `show board` too |

Delivery verified per cell before reading any outcome: `available` non-empty in 3/3 `docket` cells
for both affordance arms.

## Result

| arm | n | reach (of 17 / 10) | range | expectations MET | `worklist` | `board` | tool calls | refusals |
|---|---|---|---|---|---|---|---|---|
| `off` | 6 | 7.67 | 3–11 | **2.67** | 1.50 | 1.50 | 20.2 | 2.83 |
| `shipped` | 6 | 6.83 | 5–8 | **1.33** | 1.17 | 1.50 | 15.8 | 3.67 |
| `available` | 6 | 7.67 | 5–10 | 1.83 | 1.83 | 1.33 | 18.2 | 3.50 |
| `available+board` | 6 | 7.83 | 5–10 | 2.00 | 1.00 | 2.33 | 20.2 | 4.17 |

| contrast | reach | expectations MET |
|---|---|---|
| `shipped` − `off` | −0.83 (t = −0.62) | **−1.33 (t = −2.48)** |
| `available` − `shipped` | +0.83 (t = +0.74) | +0.50 (t = +1.10) |
| `available+board` − `shipped` | +1.00 (t = +0.94) | +0.67 (t = +1.35) |
| `available+board` − `available` | +0.17 (t = +0.13) | +0.17 (t = +0.35) |

## What it says

**The shipped duty list is associated with LESS work, not more.** It is the only contrast in the
table with |t| > 2, and it runs the wrong way: a seat given the enforced-duty list met 1.33 fewer
expectations than a seat given no duty list at all, and made 4.4 fewer tool calls. On `docket` the
pattern is uncomfortably clean — `shipped` produced reach 8, 8, 8 and MET 1, 1, 1 across three
replicates, against `off` at 10, 11, 9 and MET 3, 3, 2.

**Affordances recover part of it and do not beat `off`.** `available` returns reach to `off`'s level
(7.67 both) and expectations to 1.83 against 2.67. So the affordance list is better than the
enforced list alone and is not better than silence.

**Carriage is not the constraint.** Moving the sitting onto `board`, the projection seats actually
read, adds +0.17 verbs over content alone. The list being in the wrong place was a real defect and
was not the binding one.

## The cross-cutting observation, stated as a hypothesis

Two independent channels, measured separately, point the same way: **a short authoritative list is
associated with the seat doing less than no list.** The constitution's partial naming came last of
four arms; the board's enforced-duty list comes last of four arms. Neither result is individually
significant at n = 6. Together they are the same shape twice, in channels that share nothing but the
property of handing a seat a bounded set and implying it is the relevant one.

## What it does not say

n = 6 per arm, one weak model, two boards, ranges that overlap on every contrast but one. `off` also
reports `complete: true` unconditionally, which is a second difference from the other arms and is
not controlled — a seat in that arm is told nothing is outstanding AND told nothing is available,
and this data cannot separate those.

Reach and MET are proxies. A seat reaching for more verbs has not thereby done better work.

## Adjacent: the missing `lens` case is the rule holding, not a gap in it

`SittingOf` has no `lens` arm, so after the role fix a lens seat still receives exactly one duty —
friction. An earlier draft of this document called that "a duty set nobody wrote". That was wrong,
and checking it rather than repeating it is the correction: nothing refuses a sitting over a missing
lens act, and the scorecard scores no lens parity duty. Under this file's governing rule — every
duty is enforced at a write path or scored at capture — a lens duty would be an invented obligation,
and `complete: false` on a seat no gate would hold is precisely the disagreement that teaches a seat
to trust neither surface.

The acts a lens genuinely has open to it — verifying a citation nobody checked, re-running a proof
nobody re-ran — are affordances, and they live in `AvailableOf`, where they carry no claim about
being finished. The comment now says so at the switch, because the absence reads as an oversight to
anyone who has just fixed the `roleOf` defect and is hunting for more of it.
