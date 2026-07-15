# red candidates — round 1, lens 5 (logic & completeness)

Audit surface: full read of `blue/report.md` (914 lines, two windows, whole document in
context), plus `blue/frontier.md` and `inputs/PINNED.md` for template/frontier compliance.
Lens scope: leaps of faith, missing counterarguments, unexplored alternatives, template
compliance. Citation leaf-verification is out of lens scope; where a finding touches a
source, it is flagged for the citation lens.

## Template compliance (checked, mostly clean)

- Frontier hypotheses H1–H5 recorded before search and each disposed in §1–§5: compliant.
- Minority-claim provenance convention declared and applied throughout: compliant.
- Semantic word-based footnotes with access dates: compliant.
- Confidence self-grades per section: present (one gap — see L5-F5).
- Open questions (§8) and pre-flight self-audit (§7): present.
- One protocol-step compliance defect found (L5-F9).

## Findings

### L5-F1 — The catch-to-arm table omits its own type specimen (R5-2); no arm covers it; the §5.5 gate doesn't gate on the only mitigation that does

- **Location:** §5.2, "Every late-round catch red actually made by full re-read in run 3,
  traced to whether a four-arm scope would have surfaced it:"
- **Challenge:** §5.1 names exactly two "clean specimens": R4-4 and **R5-2** ("caught only
  by a lens re-reading the other corpus first-hand... Both catches came from audit surface a
  changed-sections rule excludes"). The §5.2 table — whose completeness claim ("Every
  late-round catch") is the load-bearing evidence for the four-arm rule's sufficiency —
  lists R4-3, R4-4, R3-6, R3-10, R5-1, R1-1 and **silently omits R5-2**. Trace it against
  the four arms: not a changed section (§4 row 1 unchanged since round 2), not contested
  lineage, not reachable by propagation-grep (no correction was accepted whose string could
  be grepped — the drift was in the *other corpus*), leaving only the random spot-check
  floor, whose catch probability blue itself calls unmeasured. The honest answer is: **no
  arm covers the R5-2 class; only PR #16 pinning does** — and §5.4 states pinning has zero
  live evidence ("this run's own PINNED.md is the first live trial"). Yet the §5.5 gate
  conditions run-5 ratification **only** on the propagation-clause record ("contingent on
  run 4's propagation-clause record showing zero unpropagated-site regressions"), not on
  pinning surviving its first trial. The four-arm rule is safe only jointly with pinning;
  the gate tests one of the two load-bearing mitigations.
- **Required fix:** add R5-2 to the §5.2 table with the honest arm assignment ("none —
  covered only by PR #16 pinning"), and extend the §5.5 gate to condition on pinning's run-4
  record as well (any cross-corpus drift regression in run 4 = same reject-outright branch).
- **Grade:** MEDIUM-HIGH — likelihood high (the omission is textual, certain; the gate gap
  follows directly) × impact medium (run-5 disposition rests on a completeness argument with
  a named hole; the HOLD for run 4 is unaffected) × complexity-to-mitigate low (one table
  row, one gate clause).
- **Corroboration confidence:** high (direct side-by-side read of §5.1, §5.2, §5.4, §5.5).

### L5-F2 — The §6.3 doctrine check skips the one position with a genuine doctrine conflict, and lever 5's conditions omit the required doctrine-text amendment

- **Location:** §6.3, "Every ratified cut lands on instance-redundancy, residency of red's
  OWN closed cases, or mechanical collation; no position reduces judge strength, red-merge
  depth, distinct-lens coverage, or the spot-check floor"
- **Challenge:** the check claims to be "run against every position," but the §5.5
  conditional RATIFY for run 5 cuts **red's read of blue's report** — which falls in none of
  the three named safe categories. Blue's own §5.3 quotes the controlling doctrine ("the
  protocol ranks full-read-of-the-audit-surface above token savings explicitly ('this clause
  outranks any token saving')") and concedes "Unlike lever 4, this lever narrows the audit
  surface itself." A run-5 ratification therefore requires amending the red-auditor
  contract's full-re-read MUST and the research-protocol mode-2 clause — a doctrine-text
  change that appears nowhere in §5.5's conditions, the lever's cost accounting, or §6.3.
  §6.3 achieves its clean verdict by silently excluding the conflicted position (presumably
  as "HOLD this run" — but the run-5 RATIFY is a position taken in this report and §6.2
  calls its trigger mechanical).
- **Required fix:** either scope §6.3's claim honestly ("every position *actuated this
  run*") and add the doctrine-amendment prerequisite to §5.5's conditions, or argue
  explicitly why a gated future ratification needs no doctrine edit.
- **Grade:** MEDIUM — likelihood high (textual) × impact medium (an unstated prerequisite
  for a registered future decision; the systematic doctrine check is advertised as
  systematic and is not) × complexity low (one scoping sentence + one condition).
- **Corroboration confidence:** high (report-internal; §5.3 vs §5.5 vs §6.3 read side by
  side; doctrine quote confirmed against the research-protocol skill text).

### L5-F3 — Condition 3's dedupe sufficiency is asserted, not argued: a one-line closure index is presumed an adequate dedupe key

- **Location:** §4.3(b), "Under naive sharding the merge either re-reads the archive
  (savings vanish) or mints duplicate ids for re-litigated closed ground. Hence the compact
  closure index (condition 3)."
- **Challenge:** blue correctly poses the dichotomy, then resolves it by assertion. Whether
  "id | closure class | one-line summary | supersedes" suffices to recognize that a fresh
  candidate gap re-litigates closed ground is exactly the judgment-heavy comparison the
  dichotomy worries about — run 3's own dedupe work (merge-time dedupe notes every round)
  compared full findings prose, and semantic near-misses (same defect, different framing —
  the R5-5-vs-lens-4 "adjacent territory" case shows how differently two seats frame one
  area) are the expected hard case. If the one-line key is insufficient, the design silently
  falls into the dichotomy's second horn (duplicate ids) — the recurrence-escalator failure
  class this repo already documented (identity-keyed detectors never fire on re-minted ids).
  Condition 4's reopen triggers (supersedes-names-it, blue-cites-it, spot-check) do not
  cover "index line looks similar to a candidate gap."
- **Required fix:** add a reopen/read trigger to condition 3 or 4: an index near-match to a
  candidate gap forces a targeted archive read **before** a fresh id is minted (mode-3
  demanded read, mirroring condition 2's pattern). Cheap, and it converts the sufficiency
  assumption into a checked step.
- **Grade:** MEDIUM — likelihood medium (semantic near-misses are the observed norm in this
  corpus, but the index may often suffice) × impact medium (duplicate ids poison lineage,
  the docket detector, and the found_by/capture-recapture instrumentation §2.5 ratifies) ×
  complexity low (one trigger clause).
- **Corroboration confidence:** high for the textual gap (conditions re-read verbatim);
  medium for the near-miss likelihood (inference from run-3 dedupe behavior, not a counted
  rate).

### L5-F4 — The grade-dispute channel is presented as the incentive loop's check without analyzing blue's symmetric incentive

- **Location:** §2.3 (incentive-loop bullet), "the grade-dispute channel (lever 3) is the
  loop's only adversarial check — which is why the two levers are a package deal in any
  actuated future"; reinforced at §3.2 and §6.2's BINDING interlock.
- **Challenge:** the analyzed loop is red inflating grades to protect red's budget. Under
  the same actuated-throttle future, blue acquires the mirror incentive: **dispute grades
  downward to shrink the lens count auditing blue's own report**. The channel routes
  disputes to the judge, which may be an adequate answer — but the report never states the
  symmetric incentive, never argues the judge's adequacy against it, and sells the interlock
  as sufficient ("the guardrail any future grade-actuated mechanism needs," §6.1). A
  two-party dispute mechanism whose both parties have budget stakes is an adjudication load
  question, not a pure correction path; the interlock's sufficiency claim is one-sided.
- **Required fix:** one paragraph in §3.5 or §6.2 naming the symmetric incentive and the
  judge as its intended absorber (plus, if honest, a note that dispute *volume* under
  actuation is an open cost question — cf. L5-F8).
- **Grade:** MEDIUM — likelihood medium (only materializes if actuation ever ships — blue
  rejects it this run) × impact medium (the interlock is BINDING and designed now precisely
  so "a future actuation debate [is] honest"; a one-sided incentive analysis baked in now is
  the kind of framing later rounds inherit uncritically) × complexity low (one paragraph).
- **Corroboration confidence:** high (report-internal logic; no external source involved).

### L5-F5 — §1.2's "deleted value" conflates deleted with delayed, and the §1 confidence grading skips the load-bearing counterfactual

- **Location:** §1.2, "Deleted value for a saving of ~$53 (rounds 4–5 seat-round sum from
  cost.md)"
- **Challenge:** two leaps. (1) "A judge disposing the round-3 residual board does not
  audit; disposition produces no R4-1" — granted for the disposition round itself, but the
  argument then treats R4-1/R5-5 as *deleted* rather than *delayed*: nothing argues the
  findings were unreachable by a later engine-auditing run (this very run 4 re-audits the
  engine; run 3 was itself the retrospective that existed to catch engine defects). The
  honest framing is "deleted from run 3, at risk of never being minted" — a weaker claim
  whose strength depends on whether engine-audit runs recur, which they demonstrably do.
  (2) §1.5 grades HIGH only "on the never-fires finding" and MEDIUM on the re-scoped
  variant; the §1.2 falsification argument — a single-transition (n=1) counterexample that
  carries the REJECT beyond mere never-fires — receives no confidence grade at all, despite
  blue's own generalization caveat ("thin evidence in both directions").
- **Required fix:** reword "deleted" to the delay-vs-delete framing with the recurrence
  argument made explicit; add a confidence grade for the §1.2 counterfactual.
- **Grade:** LOW-MEDIUM — likelihood high (the conflation is textual) × impact low-medium
  (the REJECT verdict survives on the never-fires finding alone, which is mechanical and
  convergent; this trims overstatement, not the disposition) × complexity trivial.
- **Corroboration confidence:** high (report-internal).

### L5-F6 — The registered prediction's savings figure ignores the judge-disposition round's own cost

- **Location:** §1.5, "the re-scoped floor would have saved exactly one round's spend
  (~$25–30) at zero verdict cost"
- **Challenge:** the re-scoped variant "routes the board to the judge for disposition, never
  terminates" — so the counterfactual saving is (full round) − (judge-disposition round),
  not "exactly one round's spend." The backlog's own figure for the judge round is ~$10
  ("would have ended run 3 at round 3 for ~$10"), making the honest registered prediction
  ~$15–20 net. Registered predictions exist to be settled mechanically later; an overstated
  one settles wrong.
- **Required fix:** restate as net-of-judge-round.
- **Grade:** LOW — likelihood high × impact low (a prediction's magnitude, no current
  actuation) × complexity trivial.
- **Corroboration confidence:** high (report-internal arithmetic; backlog figure quoted in
  the report's own §1.1).

### L5-F7 — Vote accounting counts a ratify vote whose stated condition blue itself rejects

- **Location:** §3.5, "RATIFY the minimal envelope form (§3.3) — two of three lanes ratify
  in some form" (and §0's "RATIFY 2/3 (lane 3 unconditional-minimal; lane 2 coupled to
  lever 2)")
- **Challenge:** lane 2's ratification was coupled to lever 2 — whose actuation blue
  REJECTS this run. A vote conditioned on a premise the synthesis rejects should not be
  counted at face value toward a 2/3 majority framing. Blue discloses the dependency
  honestly ("lane 3's grounds... are independent of lever 2's fate, which matters because
  blue rejects lever 2's actuation this run") — but the disclosure concedes the point: on
  this run's actual dispositions the unconditional ratify support is 1 of 3, and the
  ratification rests on lane 3's grounds alone plus the convergent binding interlock.
- **Required fix:** restate the tally as "1/3 unconditional + 1/3 conditional-on-a-rejected
  premise + 1/3 reject-with-binding-trigger," and let the ratification argue from lane 3's
  grounds explicitly rather than from a majority.
- **Grade:** LOW — likelihood high (textual) × impact low (the grounds are stated and
  independently sufficient or not; only the majority framing misleads) × complexity trivial.
- **Corroboration confidence:** high (report-internal).

### L5-F8 — Default-to-docket's failure path creates unpriced judge-dispatch spend in an efficiency report

- **Location:** §3.3, "treat the unaddressed disputes as REJECTED (auto-docket), not as
  absent — inheriting the R5-5/R3-2 unenforced-optional-field lesson"
- **Challenge:** the default is the right enforcement lesson, but its failure mode — red's
  optional `dispute_responses` field going silently unset (the *exact* R3-2 class the clause
  cites: "three rounds unnoticed in run 3") — now auto-escalates every open dispute to a
  judge dispatch. The report ratifies this in an efficiency investigation without a cost
  cell for a judge round anywhere (run 3 had zero dispatches; the only figure in the corpus
  is the backlog's ~$10). Expected traffic is zero on run-3-shaped records, so this is a
  completeness nit, not a design flaw — but a lever ratified for its "mechanism cost" should
  price its enforcement default's firing cost.
- **Required fix:** one sentence in §3.3 or §3.5 pricing the auto-docket path (~$10/firing
  per the backlog's only estimate) and noting the R3-2-class risk applies to the new
  optional field symmetrically.
- **Grade:** LOW — likelihood low (requires blue to dispute AND red to drop the field) ×
  impact low-medium (bounded per-firing cost) × complexity trivial.
- **Corroboration confidence:** high (report-internal).

### L5-F9 — Protocol-mandated PDF-extraction step skipped for an open arXiv source, with a misattributed "paywalled" excuse

- **Location:** §7, "lane 2's ~34% NVD-vs-CNA figure and the expert-CVSS moments are from
  search digests, not leaf-verified (paywalled; not load-bearing...)"
- **Challenge:** the expert-CVSS paper (Computers & Security 2015) is genuinely paywalled;
  **arXiv:2508.13644 is not** — it is an open arXiv preprint, and the research-protocol
  skill mandates trying `arxiv-latex`/`pdf-reader` before grading down on a lossy fetch ("a
  claim capped at 'unable to corroborate' without trying these is an incomplete audit"). The
  §7 parenthetical launders one source's paywall across both. Template compliance defect on
  blue's side; the actual figure verification belongs to the citation lens — flagged for
  handoff there.
- **Required fix:** attempt the extraction tools on arXiv:2508.13644 (or record the
  attempt's failure as friction); correct §7's parenthetical to apply "paywalled" only to
  the Computers & Security source.
- **Grade:** LOW-MEDIUM — likelihood high (textual; the protocol step is a MUST) × impact
  low (blue already grades the figure MEDIUM and calls it non-load-bearing; the 68% figure
  carries the point) × complexity trivial (one tool attempt, one wording fix).
- **Corroboration confidence:** high for the compliance gap (protocol text vs §7 read side
  by side); the figure itself unverified at this lens.

### L5-F10 — Frontier-correction sweep is incomplete: a third frontier error (stale pre-temper grades) survives

- **Location:** §6.4 item 3, "This run's frontier carries two misattributions, corrected in
  §4.1 and §5.1 so they do not propagate"
- **Challenge:** the sweep found two but the frontier carries at least a third: H1 grades
  R5-5 as "HIGH" and R4-1 as "High likelihood × High impact," while the pinned findings
  record R5-5 as MEDIUM-HIGH (merge-tempered HIGH→MEDIUM-HIGH — the tempering the report
  itself cites in §3.6) and R4-1 as *certain* × high. The frontier quotes pre-temper lens
  grades — a stale-cell instance of exactly the class this report corrects elsewhere. The
  body's own §1 uses the correct grades, so nothing downstream breaks; but "two
  misattributions, corrected... so they do not propagate" is itself now a completeness claim
  with a counterexample.
- **Required fix:** amend §6.4.3 to three, or reword to "the two that could propagate."
- **Grade:** LOW — likelihood certain (textual) × impact low (body grades are correct;
  only the frontier artifact and §6.4's count are stale) × complexity trivial.
- **Corroboration confidence:** high (frontier.md H1 vs report §1.2/§3.6 read side by side).

### L5-F11 — Minor numeric/label nits (bundled)

- **Location (a):** §2.4, "dropping rounds 3–5 from 5 lenses to 3 saves ~$4/round ≈
  $12–18/run (~10%)" — ~$4/round × 3 rounds = $12; the $18 upper bound is unreconciled
  (if it is the "slightly cheaper merges" term, say so and size it). Recompute, don't
  re-read: this corpus has a documented history of ranges that don't recompose.
- **Location (b):** §6.1 header, "Where the money actually is (priority order for ratified
  work)" — item 6 (round-scoped audit) is HOLD/"no action this run," not ratified; the list
  label contradicts its own entry.
- **Grade:** LOW — both textual, both trivial fixes, neither verdict-touching.
- **Corroboration confidence:** high (report-internal arithmetic and labels).

## Lens summary

No verdict-flipping defect found at this lens. The dispositions themselves (REJECT 1, REJECT
2's actuation, RATIFY 3a-minimal, REJECT 3b, RATIFY 4a-conditional, REJECT 4b-seat, HOLD/gate
5) each rest on at least one leg this lens could not knock over. The two findings that matter
most for the run's registered future decisions: L5-F1 (the four-arm sufficiency case omits
its own hard specimen and the gate under-conditions) and L5-F2 (the doctrine check skips the
one conflicted position; a doctrine-text amendment is an unstated prerequisite). Both are
cheap to fix and both sit exactly where later rounds would inherit them uncritically.

Flagged for citation lens: arXiv:2508.13644 figure (L5-F9); §2.4 savings arithmetic inputs
(L5-F11a).
