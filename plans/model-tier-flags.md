# Model-tier flags — group seats by cognitive demand, not by pipeline stage

Status: DESIGN / queued (non-smoke runs). Not built. Smoke stays all-cheap by design (the
baseline). Derived in the 2026-07-20 debug-run conversation; this file is the permanent
rationale the eventual flags' help text should cite.

## The problem

The research engine exposes two model knobs:

- `model` — the BULK tier: blue lanes, red lenses (all L1–L6), blue-respond.
- `judgmentModel` — blue-synthesize, red-merge, judge, assemble.

Both group by PIPELINE STAGE. The decision that actually matters groups by COGNITIVE
DEMAND, and the two cross-cut. `model` alone carries three different demands; `judgmentModel`
still carries `assemble`, which after the assemble-from-record change needs no model at all.
So an operator cannot cheapen the mechanical seats without also dulling the inventive ones.

## The axis: recoverable error vs unrecoverable absence

Sort every seat by its FAILURE MODE and whether anything downstream recovers it.

- **Recoverable error** — a wrong output a later seat catches or repairs. Tolerates a
  cheaper model; audit or the next round restores it.
- **Unrecoverable absence** — something that was never produced: an idea nobody had, a
  counterexample nobody constructed, a gap nobody caught. Nothing downstream restores an
  absence — you cannot audit your way to an idea, and you cannot verify a flaw no one saw.
  Capability-bound.

The hard part on BOTH sides of the debate is the absence. Blue's frontier is the unmade
idea; red's frontier is the unseen counterexample. The mechanical parts of each — blue
drafting known material, red verifying a quote that is right there — are the cheap-able
parts. Construction is not lookup.

Red splits inside its own sentence: "is blue making shit up" is a LOOKUP (grep the quote,
recompute the number — tool-backed, cheap-able, and red is handed the data). "Does the
argument have logical flaws" is CONSTRUCTION — the decisive object (the counterexample, the
unstated assumption, the unenumerated universal, the sibling case) is by definition NOT on
the page, so finding it is building it. A weak model reads a slick argument and passes it;
that false PASS is the absence-of-a-catch nothing recovers.

## The receipt

`citation_yield_by_round` from prior runs: round 2 was citation **1**, logic **3**,
dark-side **16**; round 3 citation **0**, logic **9**, dark-side **6**. Citation-checking
collapses to near-zero yield after round 1; the real gaps come from logic and dark-side. The
cheap-able lens and the capability-bound lenses have OPPOSITE value profiles yet today share
the one `model` flag. This is measurable: cheapen the bulk tier and dark-side yield falls
while citation yield barely moves — that delta is the cheap-red damage made visible.

## The groupings

| Group | Seats | Demand | Failure mode | Model |
|---|---|---|---|---|
| **Construction** | blue lanes, blue-synthesize, red logic (L5), red dark-side (L6) | invent the idea / construct the flaw | unrecoverable absence | big |
| **Lookup** | red citation lenses (L1–L4), blue-respond | verify given data / reactive repair | recoverable (re-audited) | cheap OK |
| **Judgment** | red-merge, judge | adjudicate the contested docket | mixed | capable |
| **Modelless** | assemble | mechanical (the tool composes) | n/a | none |

The consequential move that today's flags cannot make: **split the red lenses.** L1–L4
(citation) ride the cheap tier; L5–L6 (logic, dark-side) ride the construction tier — same
`model` flag today, opposite needs. The lens dispatch in `debate.js` assigns per lens number,
so the split is a dispatch change, not a new mechanism.

## Proposed flags (sketch, to refine when built)

Three tiers keyed to the groups, plus assemble taking none:

- `constructionModel` — blue lanes, blue-synthesize, red L5/L6. Defaults to the session
  model (the current keeper-run default of "omit `model`" exists because the bulk tier
  secretly carries BOTH construction frontiers — the inventor and the skeptic).
- `lookupModel` — red L1–L4, blue-respond. The safe place to save money.
- `judgmentModel` — red-merge, judge. (Drop `assemble` from it; assemble is modelless.)

Names/shape are a sketch; the load-bearing content is the GROUPING and its rationale above.
Whatever the flags are called, mis-tiering a construction seat to the cheap model is a
SILENT quality loss — absence leaves no error to trip a check — so the help text must say so.

## Scope

Non-smoke runs. `--smoke` stays all-cheap (haiku everywhere) as the all-cheap baseline —
its whole value is being the cheap extreme to compare tuned runs against.
