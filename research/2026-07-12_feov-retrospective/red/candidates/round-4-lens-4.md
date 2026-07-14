# Red audit — round 4, lens 4 (logic and completeness)

Scope: full re-read of `blue/report.md` (832 lines) in context, end to end, plus `blue/CHANGELOG.md`
(336 lines, used only as a navigation hint, not as the audit surface) and `debate.md` (564 lines,
full transcript) for round-3 context. Focus: leaps of faith, missing counterarguments, unexplored
alternatives, template compliance. Verified all 20 round-1, 11 round-2, and 10 round-3 gap fixes
hold as described in the body text at their cited locations — no regressions found in this lens's
read on any of those 41 prior items. The three items below are new, not re-raisings.

Also checked: the top-level `report.md` (35 bytes, a stub) — correctly still a stub pending a PASS
verdict, consistent with round-3 lens-4's prior finding on this axis; no template-compliance
violation. All footnotes use semantic word-labels (`[^WordLabel]`), none numbered — compliant with
the research-protocol mandate.

## R4-1 — OPEN — MEDIUM-HIGH — certain (the ambiguity is present in the shipped text, not
hypothetical) x medium-high (the two undecided options have materially different safety
properties) x trivial (one clause) — corroboration: HIGH (self-evident from the quoted text,
cross-checked against every other occurrence of the same language in the document; no external
re-fetch needed)

**Location:** §3 row 20 (*"(Added round 3, R3-1) Guard the schema-legal degenerate
`{verdict: 'FAIL', gaps: []}` shape"*) — Complexity cell: *"one guard clause immediately after the
existing `blueEnv`/`redEnv` null-checks: if `redEnv.verdict === 'FAIL' && redEnv.gaps.length === 0`,
either treat as `PASS`-with-a-logged-warning (red found nothing to fail on) or throw a
distinguishing error rather than looping"* — and §2.3 addition 13 (*"KNOWN-FAILING until the guard
below (§3 row 20) ships"*).

**Problem:** the round-3 fix for R3-1 is not actually a disposition — it is red's own original
required-fix text (`red/findings.md` R3-1: *"treat as effective PASS-with-warning or throw a
distinguishing error"*) copied forward verbatim, with the "or" never resolved. Compare the two
sibling round-3 fixes in the same table: R3-2's disposition states a single call to add
(`takeFriction('blue-synthesize', blueEnv)`); R3-3's disposition explicitly picks option (b) over
option (a) and gives the reason ("at low complexity this is cheaper... risk-acceptance is for when
complexity exceeds likelihood x impact, which does not hold here"). Row 20 is the only round-3 fix
that ships a disjunction instead of a decision, and the disjunction is not cosmetic: "treat as
PASS-with-a-logged-warning" and "throw a distinguishing error" are opposite failure philosophies.
The first converts an ambiguous/degenerate red-merge return into a *passing* verdict; the second
halts loudly. This is exactly the axis this retrospective's own recurring villain class runs
along — "silent" is the word this report uses against nearly every defect it has found (§0's
"a backlog checkbox is not a diff," R3-2's own finding that friction is silently dropped for one
seat, R1-11/row 2b's "systematically under-scaled... exactly when the report is largest," R3-1's
own problem statement one column over: *"no distinguishing log line, no thrown error"*). The report
has already built, in its own words, a strong argument for "throw, don't silently convert to
PASS" — and then declines to apply that argument to close the one place it would matter most: a
degenerate/malfunctioning (or, per §3 row 19's own content-poisoning finding, potentially
adversarially-induced) red-merge return silently resolving to a clean bill of health is the worst
member of the "silent" failure family this report catalogues, not a neutral tossup between two
equally-acceptable options. §2.3 addition 13 inherits the same problem one level down: a test case
titled "KNOWN-FAILING until the guard ships" cannot actually be written — the assertion depends on
which behavior "the guard" produces, and that is exactly what is undetermined.

**Required fix:** one clause picking a side, consistent with the report's own established
position against silent degradation: "throw a distinguishing error, not silent PASS-with-warning —
a merge-lens return this degenerate should halt the run for human attention, not resolve to a
passing verdict, for the same reason §3 row 19's poisoning finding argues against trusting an
unexamined clean signal." Then §2.3 addition 13's assertion can actually be written (assert throw,
not assert pass-with-warning). Red will accept an argued choice either way — including "PASS with
a loud warning surfaced to the operator, because throw loses partial progress" — the gap is the
undecided disjunction, not a preference for one answer over the other.

## R4-2 — OPEN — LOW-MEDIUM — certain (three live instances, confirmed by direct grep) x
low-medium (traceability/reader-confusion; the report's own stated value is provenance and
per-claim traceability) x trivial (one disambiguating clause, or a namespace prefix) —
corroboration: HIGH (grep-confirmed against `inputs/run2-friction.md`, the source corpus these ids
actually belong to)

**Location:** §3 row 13 (*"kept 3+ figures at unable-to-corroborate across 4 rounds (R1-19, R1-28,
R3-14/15, R4-9)"*) and §4 rank-1 row (*"Directly blocks R1-19, R1-28, R2-8's residual, R3-14,
R3-15, R4-9 from resolving past 'unable-to-corroborate-at-leaf-node.'"*, quoted verbatim from
`inputs/run2-friction.md` line 117 and reproduced in §4's ranking table).

**Problem:** unresolved cross-corpus gap-id collision, now recurring in at least two more places
than the one instance red already caught. `git grep`-confirmed: "R1-19," "R1-28," "R3-14," "R3-15,"
and "R4-9" are gap-ids from `inputs/run2-friction.md` and the *memory-architecture* retrospective's
own internal red-audit numbering (the corpus this run's H1–H5 doubts are *about*), not this
retrospective's own red-audit ids (whose namespace this report also uses — `red/findings.md` R1-1
through R1-20, R2-1 through R2-11, R3-1 through R3-10). Round 2's disposition summary already
flagged exactly this collision class once, for a single occurrence in §1.2 (*"the report's §1.2
references 'R2-10' from run 2's findings (memory-architecture). This file's R2-* ids are this
retrospective's round-2 gaps — unrelated"*) — but that note lives only in `red/findings.md`'s
disposition summary, was never folded into the report text itself, and covered only that one
occurrence. §3 row 13 and §4's rank-1 row carry the identical collision pattern, unflagged in
either the report or a prior red disposition. The practical sting is sharper for this round
specifically: "R4-9" is, right now, a live, real id-prefix in this retrospective's own numbering —
this very lens pass is producing R4-1, R4-2, R4-3 in this file — so a reader or an automated tool
grepping `blue/report.md` for "R4-" to cross-reference this retrospective's own round-4 gaps
against `red/findings.md` gets a false-positive hit that resolves to nothing in this retrospective's
findings file (there is no R4-9 in `red/findings.md`, and there will not be unless round 4
coincidentally produces nine or more gaps, at which point the collision becomes silently
indistinguishable rather than merely absent). A report that argues so extensively for per-claim
provenance and traceability (§1.2, the claim-manifest proposal, §3 row 5) should not itself ship an
overloaded identifier scheme with no disambiguating marker.

**Required fix:** one of (a) a single global disambiguation footnote or parenthetical the first
time any memory-architecture-corpus id appears, covering all instances at once (e.g., "ids in this
paragraph and elsewhere prefixed bare 'R#-#' without a corpus name belong to the memory-architecture
retrospective's own red-audit numbering, distinct from this retrospective's `red/findings.md` R#-#
series"); or (b) rename the memory-architecture-corpus citations to a distinct prefix (e.g.,
"MA-R1-19") wherever they appear, including the two new instances found here and the one round 2
already noted but never fixed in-report.

## R4-3 — OPEN — LOW — certain (textual ambiguity present in the shipped sentence) x low (a
reader who reads the full table cell, not just its first sentence, is not actually misled — the
correction three sentences later does resolve it) x trivial (one clause) — corroboration: HIGH
(self-evident from the quoted text)

**Location:** §3 row 6, the original (round-1, R1-16) disposition sentence still standing
unedited — *"assign the critical-stance/adversarial-disconfirming lens to at least 2 of N lanes
(not 1-of-N)"* — against the same row's own round-3 (R3-5) correction three sentences later —
*"four named methods (primary-literature / practitioner-production / adversarial-disconfirming-first
/ local-repo critical-stance), one of which (adversarial-disconfirming-first) carries a 2-of-N
redundancy floor."*

**Problem:** R3-5 fixed the arithmetic (3 + 2 = 5, not 4) but did not fix the wording that produced
the miscount in the first place. The original disposition sentence names "the
critical-stance/adversarial-disconfirming lens" with a slash, which reads as one compound-named
method — exactly the reading that caused the round-2 reconciliation (R2-8) to under-add by
treating four methods as if the floor sat on a combined third-and-fourth item. R3-5's own
correction resolves the ambiguity in prose ("one of which... carries...") but leaves the ambiguous
source phrase itself unedited, immediately above it in the same cell. This is the same
propagation-gap shape this report has already named twice this round for other footnotes
(R3-4: body lagging a corrected footnote; R3-10: a caveat not reaching every reading-order
instance) — here the *originating* sentence, not a downstream copy, is the one left stale. A reader
who stops at the first (round-1) sentence — plausible, since it is phrased as the operative
instruction and the correction reads as a gloss on it rather than a replacement — still risks the
exact misreading R3-5 exists to prevent.

**Required fix:** edit the original sentence directly: "assign the adversarial-disconfirming-first
lens (a distinct method from local-repo critical-stance, below) to at least 2 of N lanes." Then the
R3-5 correction becomes confirmation of an already-unambiguous sentence rather than a patch bolted
onto an ambiguous one.

## Noted, not raised (this lens)

- **§5 items 1–11, checked individually for one-sided argument or missing counterfactual:** no
  item asserts a conclusion without a stated alternative or an explicit "unrecoverable from the
  corpus" honesty flag; items 4, 5, and 7 in particular each carry a live counter-scenario
  ("the answer may already exist, but only as unlogged operator experience") rather than assuming
  the trigger condition is still open. Checked, held, not raised.
- **§2.4's "No risk-acceptance case against either [simulator or --smoke]"** — read as a possible
  unexamined-alternative leap (is there really no case for skipping `--smoke`?). On check: the
  case against building `--smoke` would have to be "the simulator alone suffices," which §2.1's own
  boundary-discipline paragraph already refutes (a simulator that fakes Task-tool permissions needs
  an oracle for correct extraction — the exact reason Tier B exists as a separate tier). Not a leap;
  the alternative is considered and correctly refuted elsewhere in the same document. Not raised.
- **Row 16b's "(b) for dev/smoke, (a) for keeper runs" split** — checked for an unexamined third
  option (e.g., a per-round rather than per-run split, given round-4 is plausibly the last round).
  No evidence in the corpus that lens quality varies within a run's rounds in a way that would
  justify mid-run tier switching; the run-level split is the coarsest granularity that matches the
  corpus's own evidence (R1-4/R1-5/R1-6/R1-7 citation catches, all from bulk-tier lens passes in a
  single run). Not raised.
- **The R3-1/R3-2 "docket for run 4, not stop-the-line" framing** — checked whether treating two
  live, confirmed, code-level control-flow defects as additive-fix-sized (rather than a gate) is
  itself a leap of faith. Both are genuinely low observed-frequency (R3-1: zero live occurrences,
  requires a merge-lens bug to trigger; R3-2: certain-by-structure but low-impact, loses only a
  prose-narrated complaint that still reaches this report via narration) — the "docket, don't gate"
  call is argued, not asserted, in both cases. Not raised as a fresh gap (R4-1 above addresses the
  narrower, undecided sub-question inside R3-1's docketed fix, not the docket-vs-gate framing
  itself).

## Disposition

Three new gaps: one medium-high (R4-1, an operationally consequential design decision left as an
unresolved disjunction inside a shipped "fix"), one low-medium (R4-2, a documentation-integrity
gap — cross-corpus id collision recurring unaddressed in two more locations, now live-colliding
with this very round's own id namespace), one low (R4-3, a source sentence left ambiguous after its
own correction). None dispute any H1–H5 conclusion or any round-1/round-2/round-3 disposition; all
three are additive, one-clause-to-one-sentence fixes in the same spirit as the round's prior
mechanical corrections. No template-compliance violation found on this pass (stub top-level
report.md is correct-as-is; footnote format compliant).
