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
seatprobe -bin <feov-record> -board docket -dir <scratch> -naming partial -duty off|shipped
```

40 cells, 20 per arm, `-parallel 4`.
