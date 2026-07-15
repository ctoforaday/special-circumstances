# Red round 2 — lens 5: logic and completeness

Full re-read of `blue/report.md` (1178 lines, read whole in three windows — Read cap; see
friction note). CHANGELOG used as navigation only. Round-1 repairs audited as new claims, per
the repair-regression discipline. Six findings; none verdict-flipping; two are regressions or
unswept siblings of round-1 repairs.

## L5-F1 — MEDIUM — high (textual: the check as written names no executor that can exist) × medium (the reconciliation half of condition 7 is what keeps index/archive drift detectable; unexecutable, it re-opens the R1-6 observability hole for conditions 3/4) × low — corroboration: HIGH (report-internal, both halves quoted; debate.js re-verified first-hand this round: no fs import anywhere, "the script has no filesystem access by design" comment at ll.32–34)

**Location:** §4.5 condition 7 — *"and the closure index's line count is reconciled against
the archive's block count each round (one arithmetic check; a mismatch throws)."*

**Problem:** policy without a mechanism, inside a round-1 repair. "A mismatch throws" names
the engine script as executor (it is the only component that throws — the lineage throw and
degenerate-FAIL guard it "mirrors" are script code), but the script cannot count lines in
`red/ledger.md` or blocks in `red/archive.md`: it has no filesystem access by design — a fact
this report itself states three times (§3.3, §2.5 item 1, §4.6 item 1) and used in round 1 to
correct the exact same class (R1-10: "emit into cost.md" impossible). The shipped throws all
validate *envelope-internal* consistency; the reconciliation needs filesystem facts. The only
ways to execute it as written: (a) both counts ride `RED_ENVELOPE` — self-reported by
red-merge, the seat under audit, which is the provenance-self-report class the same
condition's spot-check half exists to avoid; (b) an unnamed hook with fs access — not
specified anywhere; or (c) it degrades to a prompt-level MUST — precisely the hortatory class
R1-6 was raised to kill. The `archive_spot_checks`-non-empty half of condition 7 is genuinely
schema-enforceable and unaffected; the gap is scoped to the reconciliation clause.

**Required fix:** name the executor. Either: counts ride the envelope as declared fields AND
the condition-5 spot-check floor audits them against the files (self-report + audit, matching
the §2.5 found_by pattern); or the reconciliation runs in a hook (sc-recall-index pattern);
or the "throws" claim is withdrawn and condition 7's text admits the reconciliation is
prompt-level and argues why that is acceptable here.

**Lineage note for the merge:** this challenges the repair that closed R1-6. If the merge
closes R1-6 WITH REGRESSION, this finding is the successor and must carry
`supersedes: [R1-6]`.

**found_by:** ['L5']

## L5-F2 — MEDIUM — high (mechanically certain from shipped code + this run's own record) × medium (pricing error the size of the savings being ranked — the report's own R1-1 error-term argument, unapplied to a second seat) × low — corroboration: HIGH (citationPasses formula re-verified first-hand at debate.js l.198 this round: `Math.min(4, Math.max(1, Math.ceil(claim_count/40)))` + two fixed lenses appended at l.210; blue's declared claim count 132→≈148 → 4 citation instances; this run's round 1 demonstrably ran SIX lenses — `red/candidates/round-1-lens-{1..6}.md` exist, findings.md header says "six lens passes")

**Location:** §6.1 — *"Every savings estimate in this report is therefore stated against a
run-4 baseline of (run-3 seat costs) + (judge dispatches × ~$10–13)"* — and §2.4 — *"dropping
rounds 3–5 from 5 lenses to 3 saves ~$4/round × 3 rounds = **$12/run (~8%)**."*

**Problem:** stale-baseline pricing on the lens seat — the same dead-baseline class R1-1
established for the judge seat, conceded "in full" in §6.1, then not swept across the other
seats shipped code moved. Run-3's red-lens line (5 lenses/round, $9.22–$11.05) is not run 4's:
the shipped `citationPasses` recompute — which §2.3 itself quotes as "the correct
proportionality already ships" — yields 4 citation instances at blue's own declared claim
count (148/40 → 4), plus the always-appended logic and dark-side lenses = **6 lens seats per
round**, and this run's rounds 1 and 2 are actually running 6. Consequences: (i) §6.1's
"run-3 seat costs" term understates the run-4 red-lens line by ~1 lens-agent/round (~$2/round
at run-3 per-lens rates, ~$10/run) — an error of the same order as ranked items 2–5; (ii)
§2.4's recomposed R1-35(a) figure is computed on the dead 5-lens shape — under live code the
narrowed throttle cuts 4→1 citation instances = ~3 agents/round ≈ **$18/run**, not $12, so
the repair *understates the case for the throttle blue rejects* (conservative in direction,
but a registered figure that settles wrong); (iii) §4.6 item 2's "collapsing 5–8 read turns"
is now 6 lens files minimum at the merge read. No disposition flips — §2's REJECT rests on
correlation failure and doctrine, not the dollar size — but the report's own stated standard
("the error term of any estimate that ignores this line is plausibly the size of the savings
being ranked") now applies to its own lens line.

**Required fix:** restate the run-4 baseline as (run-3 seat costs, with red-lens rescaled to
the citationPasses-implied lens count at the current claim count) + (judge line); recompute
§2.4's throttle saving at the 6-lens baseline; sweep §4.6 item 2 and §6.1's red-lens share.

**found_by:** ['L5']

## L5-F3 — LOW-MEDIUM — high (textual: the list is offered "for completeness of the money map") × low-medium (priority signal to the run-5 PR omits plausibly the second-largest actionable saving the report itself identifies) × trivial — corroboration: HIGH (report-internal: §6.1 vs §6.4 item 6 side by side)

**Location:** §6.1 — *"Levers ranked by measured target × confidence:"* (six-item list) — vs
§6.4 item 6's own grading — *"complexity of the fix low: carry the ruling forward without
re-dispatch unless red's grade or evidence changed."*

**Problem:** the money map is incomplete against the report's own findings. §6.1 projects the
judge seat at ~$10–13 per dispatched round, near-certain from round 2, and §6.4 item 6 finds
a low-complexity engine fix that eliminates the *repeat* component of that line (carried gaps
re-docketing every round). Expected saving: ~$10–13 × each avoided re-docket round — plausibly
larger than ranked items 3–5 combined (item 3 is stated zero-savings insurance, item 4 zero
tokens, item 5 a paragraph). Yet the ranked list omits it: item 6's HOLD lever is included
"for completeness of the money map," so the map's declared scope is not limited to docketed
levers, and the omission is an inconsistency, not a scoping choice. The winnow list does not
bar it (it is a new defect found this run, not shipped-PR content).

**Required fix:** add the §6.4-item-6 fix as a ranked line (or state explicitly why engine
fixes found this run are out of the map's scope while a HOLD lever is in it).

**found_by:** ['L5']

## L5-F4 — LOW — certain (arithmetic; the netting rule is applied three paragraphs later at the sibling site) × low (conservative: netting shrinks the floor's claimed saving, strengthening blue's REJECT) × trivial — corroboration: HIGH (report-internal: §1.2 vs §1.5's R1-17-corrected prediction; backlog's ~$10 judge-round figure quoted in §1.1)

**Location:** §1.2 — *"a MEDIUM-HIGH-admitting floor deletes rounds 3–5 (~$78: rounds 4–5's
Σ$53 plus round 3's ~$25 from cost.md)"* and *"priced against a certain ~$78 saving."*

**Problem:** inconsistent netting across sibling counterfactuals. The floor's mechanism is
"route the whole board to the judge for disposition" (~$10, the backlog's own figure, quoted
in §1.1) — so a floor firing at the round-2 boundary buys rounds 3–5's ~$78 *minus* the
judge-disposition round it substitutes ≈ **~$68 net**. Round 1's R1-17 forced exactly this
netting on §1.5's registered prediction ("~$25–30 minus ~$10 ≈ $15–20 net"); the §1.2
headline counterfactual — reworked the same round under R1-4/R1-15 — never received it. A
repair applied at one site and not its sibling is this corpus's named propagation class.

**Required fix:** restate as "~$68 net (~$78 in deleted rounds minus the ~$10
judge-disposition round the floor routes to)"; propagate to §0 row 1 if the figure appears
there in future edits.

**found_by:** ['L5']

## L5-F5 — LOW — certain (the sentence has no subject; unparseable as it stands) × low (one sentence; the argument is recoverable from context) × trivial — corroboration: HIGH (report-internal, quoted verbatim)

**Location:** §2.2 — *"The paper originally miscited is open-access, not paywalled — §7's
excuse is corrected there. But was measured-robust in this loop, because the adversarial loop
is itself the triangulation the literature asks for."*

**Problem:** repair-severed prose. "But was measured-robust in this loop" has no subject —
the round-1 splice of the R1-5 correction record cut the clause off from its antecedent (the
round-0 subject was the throttle input / grade noise, from "The throttle input is noisy in
the general severity-grading literature..."). As it stands the paragraph's pivot — noisy in
the literature BUT robust in this loop — is grammatically orphaned; a future reader (or a
future run's docket assembly) cannot tell what "was measured-robust." Incomplete-repair
class: the repair was verified for what it added, not for what it severed.

**Required fix:** restore the subject: "But the grade input was measured-robust in this
loop, ..." (or equivalent).

**Lineage note for the merge:** textual regression introduced by the R1-5 repair; if R1-5 is
closed WITH REGRESSION this is the successor (supersedes: [R1-5]). A plain-closed + new-LOW
treatment is also defensible — merge's call; the lineage must just be declared whichever way.

**found_by:** ['L5']

## L5-F6 — LOW — high (textual) × low (conservative: the true deleted-discovery count is ~5× larger than stated, so correcting it strengthens the REJECT) × trivial — corroboration: HIGH (report-internal: §1.5's own mint counts)

**Location:** §1.2 — *"deletes rounds 3–5 (~$78 ...) and with them four findings — R3-1
(degenerate-envelope loop), R3-2 (dropped friction seat), **R4-1** ... and **R5-5**."*

**Problem:** "with them four findings" reads as exhaustive, but rounds 3–5 minted ~21 gaps by
the report's own numbers (§1.5: "every round minted new gaps (20/11/10/5/6)" — 10+5+6). The
four named are the notable casualties, not the total; as written the sentence understates the
deleted discovery by ~5× and a skeptic quoting it onward will mis-cite the counterfactual's
size. One clause fixes it.

**Required fix:** "and with them all ~21 rounds-3–5 mints, four of consequence — R3-1, ...".

**found_by:** ['L5']

## Checked clean at this lens (so the merge knows coverage)

- §0 verdict table vs section dispositions: consistent, including the R1-16-corrected 3a
  tally (no conditional-vote laundering remains; the conditional vote is explicitly
  discounted and the ratification argued from lane 3's grounds alone).
- §1.2 corrected counterfactual arithmetic: $53 + $25 = $78 ✓ (netting aside, L5-F4);
  round-2-boundary firing logic sound against the pinned board table.
- §1.5 live-code correction: verified first-hand this round — PASS break l.236 precedes
  contested block l.244; carried excluded from adjudicated (ll.252–253); the correction's
  claims match the code.
- §1.5 carried-variant arm conditions now internally consistent with §1.3's cited
  double-confirmation criterion (R1-12 repair holds).
- §2.1 universe restatement (R1-24 repair) holds; mass series self-consistent; run-3 mint
  counts reconcile (20+11+10+5+6 = 52 = 46 closures + 6 residual — cross-checked against
  cost.md's closure counts).
- §2.4 recomputed $12 = $4 × 3 ✓ and 8% = 12/149.95 ✓ *on its stated (stale) baseline* —
  the baseline itself is L5-F2.
- §2.5 item 1 executable-sink repair (R1-10) holds: mass computes from envelope-borne grades,
  log() is script-side — no fs needed, unlike L5-F1's reconciliation.
- §3.3 clauses (v)–(vii): (v) and (vi) are envelope/ordering changes the script can execute;
  no L5-F1-class defect found in them.
- §5.2 table with R5-2 row + §5.5 three-condition gate + §6.3 honesty scope (R1-3/R1-7/R1-11
  repairs): internally consistent; the gate now tests both load-bearing mitigations.
- §6.4 item counts and cross-references (six items, all anchored); §8 restatements of Q3/Q4/Q8
  match their sections.
- Template compliance: living-report structure, §7 self-audit, §8 open questions, semantic
  footnote labels with access dates — all present; no misfit found.

## Friction (lens seat)

The 25k-token Read cap forced this lens's mandatory full re-read of `blue/report.md`
(~43k tokens) into three windowed reads — a live recurrence of run-3 friction #15's class on
this run's own report, at a lens seat. The full-re-read MUST outranks token savings, so the
windows were read contiguously and completely, but the cap turns one read into three tool
turns per lens per round (× 6 lenses × N rounds at the bulk tier), and the truncation banner
invites a lazy lens to answer from page one. Wanted: a Read mode (or higher cap) sized to
"the whole audit surface, in order, once."
