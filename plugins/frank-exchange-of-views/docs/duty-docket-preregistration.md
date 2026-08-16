# Pre-registration: does the shipped duty list reduce a blue seat's work on `docket`?

**Written and committed BEFORE the run. No data existed when this file was authored.** That is the
entire point of it — the hypothesis it tests was a subgroup that survived a failed pooled test, and
a subgroup chosen after the fact is not evidence. This file exists so the choices cannot move once
the numbers arrive.

## Where the hypothesis comes from, and why it is not yet a finding

`seat-duty-channel.md` measured `shipped` − `off` at n = 6 and found expectations MET at −1.33
(t = −2.48) — the only |t| > 2 either naming or duty matrix produced. The n = 16 replication
collapsed it to −0.12 (t = −0.50). Pooled, the effect is gone.

Split by board, one half survived: `docket` alone ran −0.50 (t = −2.26) while `lens-audit` ran
+0.25 (t = +0.60). There is a reason to prefer `docket` that does not depend on its result —
delivery. The `shipped` treatment produced a non-empty duty list in 7 of 8 `docket` cells and only
2 of 8 `lens-audit` cells, because a lens seat's sole duty is the friction line and its seats rarely
open the projection carrying it. `lens-audit` was a weak instrument and diluted both prior runs.

That is still a subgroup that survived a failed test. It is a hypothesis, and this run is the test.

## Declared before the data

| | |
|---|---|
| **Hypothesis (directional)** | On `docket`, seats under the `shipped` duty arm meet FEWER expectations than under `off`. |
| **Primary outcome** | Expectations MET, per the probe's own `- MET` lines. ONE primary, declared so the winner cannot be chosen later. |
| **Arms** | `off` and `shipped` only. Naming held at `partial`. |
| **Board** | `docket` only. |
| **n** | 20 per arm, 40 dispatches. Fixed in advance; no interim peeking, no stopping early, no adding cells if the result is close. |
| **Model** | haiku, as every prior run. |
| **Analysis** | Welch's t on MET, one-sided (the hypothesis is directional). |
| **Decision threshold** | t ≤ −1.69 (one-sided, α = 0.05, df ≈ 38). Stated now so it cannot be revised to fit. |
| **Secondary, reported but NOT decisive** | reach, tool calls, refusals, worklist reads. |

## Delivery, and the exclusion rule — declared now

A cell where the `shipped` seat never received a non-empty duty list is a treatment that did not
happen, and prior runs showed that is ~1 in 8. The rule, fixed in advance:

- The primary analysis is **intention-to-treat**: every dispatched cell counts, whatever it received.
- A secondary analysis excludes `shipped` cells with no non-empty `outstanding` list.
- **Both are reported**, whichever way they fall. Reporting only the one that agrees with the
  hypothesis is the same defect this file exists to prevent, one level down.

## Prior data is NOT pooled

The `docket` cells from the n = 6 and n = 16 runs are prior evidence and stay out of this analysis.
Pooling them after seeing them is another degree of freedom, and this run is powered to answer on
its own.

## What a null means

If t > −1.69, the hypothesis is not supported and the duty-channel line of inquiry is closed —
not "needs more data". Three runs at increasing n with a shrinking effect is an answer. The claim
was already withdrawn in `seat-duty-channel.md`; a null here retires the subgroup that survived it,
and nothing about the naming, the duty list or the gate changes on the strength of any of it.

## Reproduction

```
seatprobe -bin <feov-record> -board docket -dir <scratch> \
  -naming partial -duty off|shipped -constitutions <plugin>/agents
```

40 cells, 20 per arm, `-parallel 4`.

`-constitutions` was NOT in this line when the design was committed, and adding it is the only
edit made to anything above the results. `constitutionFor` defaults to `<cwd>/../agents`, which
resolves correctly only when the harness is run from `tools/`; run from the repository root it
looked in `/home/user/agents` and every cell died in `dispatch`. That is recorded below rather than
quietly fixed, because editing a pre-registration after the fact is exactly what needs a paper
trail — and no field of the design changed.


---

# RESULT — the hypothesis is NOT supported (2026-08-16)

Run against the design fixed at `1b5398c`. Nothing above this line was altered except the
reproduction command, for the reason given there.

## The void first attempt

The first 40 dispatches all died in `dispatch` before a seat existed: `constitutionFor` resolves
`<cwd>/../agents` and the harness was run from the repository root. **Zero outcomes were produced**,
which is the one condition where re-running is the fix rather than a second bite — nothing was
measured, so nothing could be selected by measuring again. Arms, n, primary outcome and threshold
were untouched. A single cell was smoke-tested before committing the re-run.

**And the check that caught it nearly hid it.** The delivery script reported `cells=0`, which reads
as "the treatment never reached the seats" — a finding — when it meant "no run happened at all".
Same number, two states. The analysis now separates a MISSING TRAJECTORY from a trajectory showing
no delivery, and reports each.

## Delivery

| arm | n | cells with a non-empty `outstanding` |
|---|---|---|
| `off` | 20 | 0 (correct: `off` emits no duties) |
| `shipped` | 19 | **18** |

One `shipped` cell (`r3`) failed to build its board and produced no data; one (`r16`) ran without
receiving a duty list and is excluded from the secondary, per the rule declared in advance.

## Primary — intention-to-treat, every dispatched cell

| arm | n | MET |
|---|---|---|
| `off` | 20 | 2.05 |
| `shipped` | 19 | 1.89 |

diff **−0.16**, **t = −1.11**. Declared threshold t ≤ −1.69. **NOT SUPPORTED.**

## Secondary — per-protocol, undelivered cell excluded

| arm | n | MET |
|---|---|---|
| `off` | 20 | 2.05 |
| `shipped` | 18 | 1.83 |

diff **−0.22**, **t = −1.65**. **NOT SUPPORTED**, by 0.04.

That near-miss is the entire reason the threshold was written down first. −1.65 is not −1.69, and
the pre-registration forbade adding cells if the result came out close. It came out close. No cells
were added.

## Secondary outcomes, reported and not decisive

| arm | reach | unmet | tool calls | refusals | worklist reads |
|---|---|---|---|---|---|
| `off` | 8.30 | 2.95 | 21.6 | 3.90 | 1.60 |
| `shipped` | **9.16** | 3.11 | 22.7 | 4.37 | 2.26 |

Reach runs the OTHER WAY here — `shipped` seats reached for *more* verbs, not fewer. The original
story was that the duty list makes a seat do less; on the board that story was built from, at
n ≈ 20, it does slightly more.

## The effect shrinks as n grows

| run | n per arm (docket) | MET diff | t |
|---|---|---|---|
| first matrix | 3 | −1.33 (pooled across boards) | −2.48 |
| replication | 8 | −0.50 | −2.26 |
| **this, pre-registered** | **~20** | **−0.16 / −0.22** | **−1.11 / −1.65** |

Three runs at increasing n with a monotonically shrinking effect that never clears its threshold.

## What this closes

Per the rule declared in advance: **the duty-channel line of inquiry is closed, not "needs more
data".** The cross-channel claim was already withdrawn; this retires the subgroup that survived it.

Nothing about the constitution's naming, the duty list, or
`TestEveryRecordingVerbIsNamedInAPrompt` changes on the strength of any of it. The question that
started this work — whether naming verbs in a constitution helps or hurts a seat — remains
unanswered, now across three experiments and 120 dispatches. What the work produced instead is seven
fixed defects of one class, and that was never what it was looking for.
