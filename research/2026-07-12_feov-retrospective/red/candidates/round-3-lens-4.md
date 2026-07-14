# Red audit — round 3, lens 4 (logic and completeness)

Scope: full re-read of `blue/report.md` in context (not the CHANGELOG diff), against leaps of
faith, missing counterarguments, unexplored alternatives, template compliance. Verified all 20
round-1 gap fixes and all 11 round-2 gap fixes (R1-1..R1-20, R2-1..R2-11) hold as described —
no regressions found in this lens's re-read; the two items below are new, not re-raisings.

Also checked: `references/report_template.md` and `references/catechism_template.md` (blue's
report is not required to contain the Catechism — that belongs to the final assembled
`report.md`, correctly left a stub pending a PASS verdict; no template-compliance violation
found on that axis) and `debate.js` at `main` HEAD (`d164ab2`) directly, to check two of the
report's own internal-consistency claims against the live source (§2.3 item 1's
"blue-respond's reassignment... cannot crash" claim, and row 16b's bulk/judgment fallback
claim) — both verified true on direct read; not raised as gaps.

## R3-1 — OPEN — LOW-MEDIUM — certain x low-medium x trivial — corroboration: HIGH (self-evident from the quoted text, no external re-fetch needed)

**Location:** Footnote `[^DiminishingReturns]` (the round-2, R2-4 replacement text) — *"harder
tasks shift the breakeven higher and, per arXiv:2606.02646's actual finding, may show
diminishing returns arriving even earlier (practical knee ≈10, effective diversity saturating
well before nominal N=30) rather than later — the direction of the correction is toward *more*
caution about adding agents on hard tasks, not less."*

**Problem:** internal contradiction inside a single sentence, introduced by the R2-4 fix itself
(the report's second round of trouble in this exact footnote — R1-5 restated it qualitatively,
R2-4 then had to drop an uncited "7 agents" clause; this is a third, smaller defect in the same
spot, worth naming as a recurring-footnote pattern even though each instance is low-severity on
its own). "Shift the breakeven higher" means harder tasks tolerate *more* agents before returns
diminish — i.e., diminishing returns arrive *later* (at a higher N) than on moderate tasks. The
clause immediately joined to it by "and" claims the opposite: diminishing returns arrive "even
earlier... rather than later." Both cannot be true of the same quantity in the same direction.
The underlying source data (practical knee ≈10 > moderate-task breakeven of 2–4; but an
absolute ceiling of ~1.8 effective agents by N=30) *can* support a coherent claim — but only by
distinguishing two different senses of "diminishing returns" (the nominal-N breakeven point vs.
the eventual saturation ceiling reached well short of a large nominal pool) that the sentence
conflates instead of naming. As currently worded, a careful reader cannot extract a single
consistent claim from it — which is exactly what this footnote's own citation-discipline
history (R1-5, R2-4) exists to prevent.

**Required fix:** disambiguate the two senses explicitly, e.g.: "harder tasks shift the
breakeven to a higher nominal agent count (moderate ~2–4 vs. harder ~10, per the paper's
'practical knee'), but the *ceiling* those extra agents are chasing is itself lower than naive
scaling would suggest — effective diversity saturates around 1.8 agents by N=30 regardless of
task difficulty, so 'add more agents on hard tasks' buys less than the higher breakeven number
alone implies." Or, if the two clauses were meant as alternatives rather than a conjunction, use
"or" instead of "and" and say which one blue's own synthesis favors.

## R3-2 — OPEN — LOW — certain x low x trivial — corroboration: HIGH (grep-confirmed exact locations)

**Location:** §2.1, Tier A defect table, row "Uninitialized/stringified `runDir`/`topic`" —
*"run-1 `journal.jsonl`: every dispatch detected the defect and refused to fabricate; 252.9k
tokens, 11m48s, honest UNVERIFIED deadlock[^Run1Friction][^Run1Journal]"*.

**Problem:** incomplete propagation of round 2's own fix (R2-5). R2-5 added
`[^CostFigureProvenance]` — flagging that the report's headline token-cost figures (252.9k run
1, ~3M run 2) are self-reported and of unstated/possibly-undercounted provenance — to every
*other* place these numbers are quoted (§2.3's closing line, §2.4's likelihood x impact
argument). This row quotes the identical "252.9k tokens" figure, within the same §2 the R2-5 fix
was scoped to ("pre-cost-audit token figures throughout §2–§4"), and carries no such caveat — a
reader who meets the number here first (§2.1 precedes §2.3/§2.4 in reading order) sees it
presented as a plain fact, not a flagged-provenance one. Direction of risk is understatement
only (per R2-5's own finding), so this does not change any verdict — it is a consistency/
completeness gap in how far the round-2 repair actually reached, the same footnote-lag pattern
red has caught in this document before (R1-9/R1-10's write-block quote, R1-19's URL — a repair
applied at its point of origin but not at every point of citation).

**Required fix:** append `[^CostFigureProvenance]` to this row's citation, or add one clause to
the table's own preamble noting the same caveat applies to every token figure in the tier
tables.

## Noted, not raised (this lens)

- **§3 row 8's "(c) risk-accept/skip — superseded by (a), which is strictly cheaper for the same
  outcome"** initially read as a leap of faith (declaring (a)'s efficacy proven while the same
  row also says "(a) shipped — verify against the real skeleton in run 4"). On direct check
  against `debate.js`, this is better-evidenced than it looks: the corpus's own record (this
  retrospective's round-1 red-merge hit) shows `Edit` succeeding on the identical filename
  (`red/findings.md`) where `Write` was refused — since (a)'s entire mechanism is "subagents
  only append/Edit" against a pre-created file, the load-bearing half of the claim already has a
  positive data point, not merely an assumption. Not raised as a gap; flagged here per the
  stickler duty to show the near-miss was checked, not skipped.

## Disposition

Two new gaps, both low/low-medium severity, both trivial-complexity, both high-corroboration.
Neither disputes any H1–H5 conclusion or any round-1/round-2 disposition. Both are additive
one-line/one-tag fixes in the same spirit as the round's other mechanical corrections.
