# Red audit — round 5, lens 4 (logic and completeness)

Scope: full re-read of `blue/report.md` (911 lines) end to end, in context — not a diff against
`blue/CHANGELOG.md` (used only as a navigation hint) — plus `debate.md`'s round-4 BLUE/RED
sections (lines 568–713) to confirm what blue's round-4 response actually changed and what red's
round-4 merge actually said. Verified all 20 round-1, 11 round-2, 10 round-3, and 5 round-4 gap
fixes hold as described at their cited locations — no regressions found in this lens's read on any
of those 46 prior items, with one exception (R5-1 below, which is itself a fresh instance of the
report's own repair-propagation-lag pattern surfacing inside the very row that names that pattern).
Template compliance checked: top-level `report.md` remains a correct stub pending PASS; all
footnotes are semantic word-labels, none numbered.

## R5-1 — OPEN — MEDIUM — certain (both text bodies read side by side in the same document; no
external fetch needed) x medium (this is the evidentiary cell backing the report's own
highest-graded still-open finding, and contains an outright wrong linkage, not merely a stale
numeral) x trivial (replace one parenthetical with text already sitting two sections earlier in
the same document) — corroboration: HIGH (cross-checked against `red/findings.md`'s own
round-2/round-3 closure entries for R2-1, R2-5, R2-7, R2-8, R3-5, R3-7, which state each chain's
actual descendant explicitly)

**Location:** §3, row 23 (*"Lineage-following contested-gap detection — the docket detector is
id-string-equal, not lineage-aware"*), Likelihood cell — *"this exact corpus contains four live
regression chains (R1-5→R2-4→R3-4/R3-9; R1-13→R2-1→R3-7; R1-16→R2-8→R3-5; R2-5→R3-8)"* — against
§2.1, Tier A, the gap-id-rollover row's sub-bullet (b) — *"plus at least three more same-shaped
chains in this corpus (per debate.md's round-4 RED merge-seat enumeration, cross-checked against
the cited rows/footnotes): R2-5 → R3-10 ... R2-7 → R3-6 ... and R2-8 → R3-5 → R4-3."*

**Problem:** these are two different enumerations of "the four regression chains in this corpus,"
and they disagree on three of the four entries. §2.1(b) is corrected and accurate — every link in
it matches `red/findings.md`'s own closure record (R2-5 "CLOSED WITH REGRESSION (round 3) →
R3-10"; R2-7 "CLOSED WITH REGRESSION (round 3) → R3-6"; R2-8 "CLOSED WITH REGRESSION (round 3) →
R3-5" and R3-5 "CLOSED WITH REGRESSION (round 4) → R4-3"). §3 row 23 is **not** corrected — it
still carries blue's own discarded first-pass guess, which blue's round-4 `debate.md` text
explicitly narrates replacing: *"I initially reconstructed the 'three more chains' independently
from the report's own gap history (R1-13→R2-1→R3-7, R1-16→R2-8→R3-5, R2-5→R3-8) before checking
this file's own round-4 RED section, which had already enumerated a different... set... Re-verified
red's list against the cited rows and adopted it in place of my own"* (debate.md, round-4 BLUE,
item 1). That substitution reached §2.1(b) but was never applied to row 23, where the discarded
list still stands, verbatim, as the cell's sole evidence.

Two of the three stale entries are not just superseded phrasing — they are factually wrong against
the corpus's own record:
- **"R2-5→R3-8" is incorrect.** `red/findings.md`'s own round-2 closure entry states plainly:
  "### R2-5 — CLOSED WITH REGRESSION (round 3) → **R3-10**." R3-8 is a different gap entirely (the
  `[^CostFigureProvenance]` footnote's stale re-pin, unrelated to R2-5's cost-figure-provenance
  *caveat-coverage* gap). Citing R3-8 as R2-5's regression descendant sends a verifier checking this
  row's own headline evidence to the wrong gap.
- **"R1-13→R2-1→R3-7" is not a regression chain by the corpus's own explicit disclaimer.**
  `red/findings.md`'s R2-1 closure entry states: *"A narrower, distinct follow-on — whether
  occurrence 2 is even the same mechanism — is R3-7, **not a reopening**."* The corpus itself says
  R2-1→R3-7 is not the closed-WITH-REGRESSION pattern row 23 is illustrating; including it as a
  three-link regression chain contradicts red's own findings file, which this report cites as its
  source elsewhere in the same paragraph's corrected form.

This is the report's own named failure class — repair reaching the conclusion but not every
instance of the source text (R3-4, R3-10, R4-4 all diagnose exactly this shape) — recurring inside
the paragraph that argues the debate engine cannot track exactly this shape of defect. The row's
Likelihood cell is the single most load-bearing piece of evidence for the report's current
highest-severity open finding (R4-1/row 23 itself, still graded HIGH going into this round); a
reader spot-checking that evidence hits an in-document contradiction with `red/findings.md`, which
this same cell cites as source material two clauses later ("per debate.md's round-4 RED merge-seat
enumeration").

**Required fix:** replace row 23's parenthetical list with §2.1(b)'s corrected one: "(R1-5→R2-4→
R3-4/R3-9; R2-5→R3-10; R2-7→R3-6; R2-8→R3-5→R4-3)" — the same four chains already stated correctly
two sections earlier. One clause; no new research.

## R5-2 — OPEN — LOW — certain (debate.md's own header count, re-checked live) x low (does not
change the substantive finding — if anything it strengthens it by one more round of zero-dispatch
evidence — only the round-count numeral is stale) x trivial (two-word edit, two locations) —
corroboration: HIGH (`grep -n "^### " debate.md` re-run this round: 9 headers — 5 BLUE, 4 RED,
zero LEAD — covering rounds 0 through 4, all complete)

**Location:** §2.1, Tier A, gap-id-rollover sub-bullet (b) — *"the judge was never dispatched once
across this corpus's **three completed rounds**"* — and §3 row 23, Likelihood cell — *"the judge
was dispatched zero times across **three completed rounds**."*

**Problem:** both instances of "three completed rounds" were accurate at the moment blue wrote them
during round 4 — at that point in the transcript, red's round-4 merge (the event citing "zero
`### LEAD` sections across rounds 0–3") had just been appended to `debate.md`, and blue's own
round-4 response, which would close out round 4, had not yet been written. Blue's round-4 response
*has* since been written and appended (`debate.md`, lines 647–713) — and it did not add a `### LEAD`
section either, because the same undetected-lineage bug this row describes meant the round-4
docket, like every round before it, never armed. `debate.md` now contains 9 section headers (5
BLUE: rounds 0–4; 4 RED: rounds 1–4) and still zero `### LEAD` headers — round 4 is the *fourth*
completed round with no judge dispatch, not the third. The number is one round stale in exactly the
way this report's own §3 row 11 / R2-6 finding warns about (a claim about "current" state that time
has already moved past by the point a later round reads it) — here self-applied to this report's
own most recently added row.

**Required fix:** update both instances from "three completed rounds" to "four completed rounds"
(or phrase round-agnostically: "every completed round to date"), reflecting that round 4 itself
closed without a judge dispatch, extending rather than contradicting the finding.

## Noted, not raised (this lens)

- **§5's twelve open questions, re-checked for one-sided argument:** none assert a conclusion
  without a stated alternative; items 4/5/7/11 each still carry an explicit "unrecoverable" or
  "advisory, not self-enforcing" honesty flag rather than assuming their trigger condition is
  settled. Held.
- **§3 row 20/R4-2's "throw" decision, re-checked against §2.3 addition 13's wording:** the decided
  behavior (throw a specific, quoted error string) and the addition-13 assertion text agree
  verbatim. No regression.
- **Row 16b's dev/smoke-vs-keeper split, re-checked for a missing mechanism:** initially suspected
  this depends on an unbuilt `--smoke` flag (row 4, still OPEN); on check, the split is gated on the
  existing `--model`/`judgmentModel` CLI knobs (confirmed live per R1-12's closure: "no `--model`
  flag, bulk seats inherit the session model"), which already exist independently of `--smoke`'s
  argument-parsing path. Not a leap; the mechanism the disposition needs is already shipped. Not
  raised.
- **§3 row 23's own required-fix item (1), "instruct red-merge to set [supersedes] when closing a
  gap 'WITH REGRESSION'":** checked whether "WITH REGRESSION" is a documented protocol state
  anywhere red-merge is actually instructed from. Direct grep of `agents/red-auditor.md`,
  `research-protocol/SKILL.md`, and `debate.js`'s red-merge prompt (line 167, "Gap ids are stable
  across rounds (R1-1 stays R1-1)") for "regression": zero hits in all three — the convention is
  this corpus's own emergent authoring practice, not doctrine. Considered raising this as a
  completeness gap on row 23's fix (the schema addition needs an accompanying prompt change, not
  just a schema change) but the row's fix text already says "instruct red-merge to set it," which
  covers the prompt-change obligation at the same specificity the rest of §3's docketed-for-run-4
  items use (e.g. row 21's single-call fix, row 22's two-option fix) — holding this to a fuller
  implementation-spec standard than its siblings would be inconsistent. Not raised.
- **Item 6's "cost is one more lane-dispatch" phrase (§3 row 6), re-checked against the row's own
  later-stated `lanes >= 5` minimum (vs. today's default of 3):** read in isolation this could
  mislead a reader into thinking the total cost from today's baseline is +1 lane, when the full
  fix's total cost from the shipped default is +2. On check, the phrase is scoped to the marginal
  cost of the redundancy floor specifically (4 unfloored lane-assignments → 5 with the floor), not
  the total cost from the current default — and the full `lanes >= 5` arithmetic is stated
  explicitly two sentences later in the same cell. A reader who reads the full cell is not misled;
  narrower than R5-1/R5-2 and already effectively covered by the cell's own later text. Not raised.

## Disposition

Two new gaps, both inside the same row (§3 row 23) that this retrospective's round-4 audit made
its gate-tier finding: one medium (R5-1, a factually wrong regression-chain enumeration standing
uncorrected beside its own already-corrected counterpart two sections earlier — two of its four
links do not match `red/findings.md`'s own closure record, one of them explicitly disclaimed by
that record as "not a reopening"), one low (R5-2, a round-count numeral one round stale now that
round 4 has itself completed without a judge dispatch — substance unaffected, arguably
strengthened). Neither disputes H1–H5, the round-4 verdict, or any prior round's closure; both are
one-clause propagation fixes of exactly the shape this report has repeatedly diagnosed in other
rows (R3-4, R3-10, R4-4) — the irony being that this round's catch is inside the very row
documenting that failure class in the debate engine itself. No template-compliance violation found
on this pass.
