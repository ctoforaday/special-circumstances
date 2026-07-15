# Round 3 — lens 6: dark-side and risk (failure modes, likelihood × impact × complexity, security and tradeoff blindspots)

Audit surface: FULL re-read of `blue/report.md` (1365 lines, three windowed reads — the 25k
Read cap forced windowing again; friction noted below). CHANGELOG used for navigation only.
Focus: the round-2 repairs as new mechanism claims — each new control audited for its own
failure modes, operator, and incentive interactions, per the lead's round-3 bar ("every
regression this round came from repair text asserting something unverified").

Leaf verifications performed at this seat:

- `C:/Users/gbloc/Projects/special-circumstances/.gitignore` read first-hand: the only
  trajectories rule is `**/trajectories/agent-transcripts.tar.gz`; run-3's
  `trajectories/journal.jsonl` is tracked (`git ls-files` confirms). The §2.5-item-1 premise
  that `trajectories/board-telemetry.jsonl` CAN be git-tracked is sound — checked, no
  discrepancy (quote: gitignore contains no `*.jsonl` or blanket `trajectories/` rule).
- `debate.js` judge-envelope schema read first-hand: `required: ['deadlock', 'resolutions']`
  (l.127); `resolutions` items require exactly `['gap_id', 'resolution', 'rationale']`
  (l.135); resolution enum at l.138. **No field in the judge envelope can carry a
  demanded-read log.** Load-bearing for L6-F2.
- cost-audit glob `agent-*.jsonl` cannot sweep `board-telemetry.jsonl` (name mismatch) — no
  accidental-consumer hazard; checked, not raised.
- The R2-11 repair (grade-change-only re-dispatch trigger, evidence-change routed via
  successor-id lineage) was probed for relocation-of-problem: the successor-id route is
  red-controlled but VISIBLE (lineage machine-checked, successor prose read by the judge) —
  it converts an invisible self-report fallback into a judged path. Sound; checked, not
  raised.

## Findings

### L6-F1 — the contest window is a control without an operator, and its "equivalently" is false

**Severity: LOW-MEDIUM now (channel unbuilt; actuation rejected this run) rising to
MEDIUM-HIGH under the actuation the interlock is mandatory for — likelihood low →
medium-high under actuation × impact medium-high (unreviewed deflation enters the mass
computation that sets red's lens budget) × complexity trivial-to-low. Corroboration: HIGH
(report-internal, quoted verbatim; the salami arithmetic uses clause (vii)'s own cap).**
**Lineage note for the merge: successor-candidate to R2-6 — the defect lives in the R2-6(a)
repair text itself.**

**Location:** §3.3 clause (v) — *"under actuation, an accepted-dispute delta enters the mass
computation only after a **one-round contest window** (or, equivalently, deltas above a
stated magnitude auto-docket to the judge before actuating)"* — and — *"the accepted
branch's absorbers are the contest window in-run and the post-hoc audit after."*

**Problem:** three defects in the repair's own text.

(a) **The window has no named contester.** An accepted delta is, by construction, a change
both parties agree on: blue proposed it, red-merge accepted it ("because accepting is
cheaper than arguing — the report's own documented behavior," per R2-6's accepted analysis).
During the one-round window, no seat is positioned to contest: the red lens seats read
blue's report (the delta lives in red's findings and the telemetry line, neither on the lens
read surface); red-merge is the seat that just accepted (the conflict R2-6 itself named for
the condition-5 floor); the judge sees only the docket, and an accepted dispute never
dockets. A delay with no reviewer is a delay, not an absorber — calling the window an
"absorber" repeats, one round later, the exact overclaim R2-6(a) was raised to kill.

(b) **"Or, equivalently" is false.** The auto-docket branch routes the delta to a named
independent seat before it actuates; the window branch routes it to nobody. The two branches
are equivalent only in delaying actuation by a round; they are not equivalent in review.
Offering them as interchangeable lets an implementer pick the weak branch while citing the
clause as if it bought the strong one.

(c) **The magnitude threshold is per-delta and clause (vii) caps disputes at 5/round —
salami-slicing is in-spec.** Under actuation, blue can submit up to five sub-threshold
deflations per round, every round; each is individually below the auto-docket magnitude,
each is rationally accepted by a red-merge for whom arguing is dearer, and each enters mass
after a window nobody watches. The compounded deflation across a run is bounded only by
5 × threshold × rounds.

**Required fix (one sentence each):** (i) name the window's operator with a read-surface
change that makes it real — pending-entry deltas MUST be listed in the round's `### RED`
debate entry (already on blue's and the judge's read surfaces), and any seat may docket them
within the window — or delete the window branch and make auto-docket the only preventive
path; (ii) drop "equivalently"; (iii) make the auto-docket magnitude cumulative per round
(sum of accepted deltas this round above threshold → batch-docket, mirroring clause (vii)'s
own overflow rule).

**found_by:** L6

### L6-F2 — condition 7 logs the judge's demanded reads in a field the judge cannot write

**Severity: LOW-MEDIUM — likelihood medium-high (the spec is ratified and would be built as
written; this is the third consecutive round a condition-7-class observable rides an
unavailable mechanism: R1-6 → R2-4 → this) × impact low-medium (the judge's condition-2
compliance — the ruling class most sensitive to missing ancestor context is `carried` vs
`risk_accepted`, the §6.4-item-6 gate-erosion path — becomes unobservable exactly where §4.5
claims it is logged) × complexity trivial. Corroboration: HIGH — judge envelope schema
leaf-verified first-hand this round: `required: ['deadlock','resolutions']`, resolution
items carry `gap_id | resolution | rationale` only; there is no judge-side field.**
**Lineage note for the merge: successor-candidate to R2-4 (and touches the R2-5 repair) —
the defect is in the round-2 rewrite of §4.5 condition 7.**

**Location:** §4.5 condition 7 — *"Condition-2 demanded reads (red-merge's and the judge's)
are logged in the same field."* — where "the field" is `RED_ENVELOPE`'s
`archive_spot_checks`, per the same condition's opening clause.

**Problem:** the judge is a separate dispatch returning a separate envelope
(`{deadlock, resolutions[]}` — verified at debate.js ll.127–138); it cannot write
`RED_ENVELOPE.archive_spot_checks`, and its own envelope has no analogous field. As written,
the judge half of condition 2's demanded-read MUST (added round 2 per R2-5) has no
observable channel — prompt-level MUST with a logging clause that names an impossible
mechanism. This is the attestation-ceiling discipline (§6.2: "any future condition reaching
for 'required envelope field' ... must state which tier it is buying") violated by the very
condition that was rewritten to honor it.

**Required fix:** one clause — the judge's demanded reads are logged in its ruling record
(the rationale text of the affected resolution, or the `### LEAD` debate entry), which is
git-tracked and post-hoc auditable; red-merge's stay in `archive_spot_checks`. Do not claim
a shared field.

**found_by:** L6

### L6-F3 — the actuation evidence base: found_by got an independent re-derivation clause; the mass series itself did not

**Severity: LOW-MEDIUM — likelihood low-medium (mechanical transcription at a seat holding
every input; but the in-corpus base rate for arithmetic defects in cost artifacts is not
zero — cost.md shipped two internal numeric defects, §6.4 item 1) × impact medium (the
logged mass series is the PRIMARY input to the §2.5-item-3 actuation decision; a
mistranscribed series poisons exactly the evidence the deferred decision needs — the same
argument R2-7 made for found_by, accepted) × complexity trivial (one sentence, symmetric
with the existing clause). Corroboration: HIGH (report-internal: §2.5 items 1–3 side by
side; R2-7's accepted rationale applies verbatim).**

**Location:** §2.5 item 1 — *"Work-done integrity for the line itself rides the §6.2
attestation ceiling: line count per round is post-hoc reconcilable against the round count
in `debate.md`."* — vs §2.5 item 2's round-2 clause — *"any future actuation review MUST
re-derive `found_by` for a sample of gaps independently from the preserved lens files at a
seat other than red-merge, and the actuation case must cite that re-derivation."*

**Problem:** line-count reconciliation is a presence check — it catches a MISSING line,
never a WRONG one. The values inside each line (mass, open count, max severity, new-mint
profile, accepted-dispute deltas) are red-merge transcription self-report, and §2.5 item 3
makes the logged record the trigger for revisiting actuation ("only when runs 4–5's logged
record shows mass ... actually predicting next-round value"). The round-2 repair gave
`found_by` an independent re-derivation clause on precisely this reasoning (self-reported
instrument feeding the actuation decision) but left the mass series — the larger input to
the same decision — covered by presence-reconciliation only. The asymmetry is unargued.
Mitigating fact, stated honestly: unlike found_by (clustering judgment), mass is arithmetic
over grades that persist in git-tracked findings, so it is cheaply re-derivable — which is
why the fix is one sentence, not a design.

**Required fix:** extend the §2.5-item-2 re-derivation clause to the series: the actuation
review MUST recompute the mass/board columns for a sample of rounds directly from the
git-tracked findings record (at a seat other than red-merge) and cite the recomputation;
telemetry lines are the convenience copy, never the evidence of record.

**found_by:** L6

### L6-F4 — a `carried` ruling at a terminal exit is disposition-in-name-only, and the exit dispatch is an unpriced firing site

**Severity: LOW — likelihood low (channel unbuilt; requires pending disputes at exit AND a
carried ruling) rising under actuation × impact low-medium (record integrity at exactly the
terminal case clause (vi) was rebuilt for: a carried dispute at exit ships the contested
grade with a judge stamp and no follow-up obligation — cosmetically stronger than mooting,
substantively identical) × complexity trivial. Corroboration: HIGH (resolution enum
leaf-verified at debate.js l.138 includes `carried`; the structural point mirrors the
lead's own round-2 note that `closed`/`rebuttal_sustained` were "unavailable by
construction" for first-raise successors).** **Lineage note for the merge:
successor-candidate to R2-6 (clause (vi) residual).**

**Location:** §3.3 clause (vi) — *"pending or held disputes at ANY loop exit — PASS,
deadlock, or maxRounds — auto-docket for judge disposition before assembly."*

**Problem:** (a) at a terminal exit there is no subsequent round to carry into, so the
judge's `carried` resolution is incoherent for this docket class — yet nothing excludes it;
a carried-at-exit dispute exits the record looking disposed while the contested grade ships
exactly as if clause (vi) did not exist. (b) The exit-time dispatch is a new judge-firing
site: R1-18 priced the default-to-docket firing (~$10/firing) but §3.3's cost enumeration
("two optional fields, one filter clause, ...") does not carry the terminal firing class,
and it fires at every exit with any pending dispute.

**Required fix:** one sentence — at a terminal exit the resolution set for docketed disputes
excludes `carried` (available: accepted-with-delta, rejected-recorded-as-contested,
`unresolved`); one clause adding the exit firing to the ~$10/firing price already stated.

**found_by:** L6

### L6-F5 — "pinned or version-stamped" — only one disjunct prevents the harm the condition states

**Severity: LOW — likelihood medium (open question 6 is undecided and run 4 is the first
logged round — the weak branch is the path of least resistance) × impact low-medium (the
stated harm: "runs 4–5's mass series incomparable, poisoning exactly the evidence the
deferred actuation decision needs") × complexity trivial. Corroboration: HIGH
(report-internal, quoted verbatim).**

**Location:** §2.5 item 1, mapping stability condition — *"the enum→numeric mapping must be
pinned or version-stamped into each logged line *before* the first logged round — an
undecided mapping (open question 6) changed mid-series makes runs 4–5's mass series
incomparable."*

**Problem:** (a) version-stamping a mid-series mapping change does not make the series
comparable — it makes the incomparability visible. The condition's own stated harm is
prevented only by the pinning disjunct; offering the two as alternatives licenses the weak
one (the same false-equivalence shape as L6-F1(b)). (b) The condition implies a hard
deadline — open question 6 (`does realized count in open-gap mass?`) must be DECIDED before
run 4's first logged round — but §8 Q6 carries it with no owner and no decision path; a
condition with a deadline and no assignee defaults to unmet.

**Required fix:** "pinned before the first logged round; a changed mapping starts a NEW
series (stamped lines are not comparable across versions)" — and either decide Q6 in this
run's disposition (simplest: `realized` excluded from open-gap mass, since realized risk is
no longer a probability — blue's own §2.5 design note already leans there) or name the
decider and the deadline in Q6 itself.

**found_by:** L6

### L6-F6 — the §4.2 decomposition table's rows sum to 80–96%; the residual is undisclosed

**Severity: LOW — likelihood certain (arithmetic on the printed table) × impact low (the
sharding-addressable figure derives from the findings column, which is measured; but a
reader summing the table cannot see that 4–20% of merge ingest is attributed to files
outside the four columns, and the caveat list names other limits — bytes→tokens, system
prompt — while omitting this one) × complexity trivial. Corroboration: HIGH (the table's
own cells: r1 36+0.1+46+4 = 86.1%; r2 = 80%; r3 = 91%; r4 = 91%; r5 = 96%).**

**Location:** §4.2, the round-2 measured table — *"| red-merge round | tool-result ingest |
blue/report.md | red/findings.md | lens candidates | debate.md |"* — and its caveat
sentence — *"Caveats, stated: single run; bytes→tokens ≈ 4:1 assumed; the weighting ignores
the system prompt and harness overhead."*

**Problem:** the four named columns leave an unlabeled residual of ~4–20% per round
(presumably PINNED.md, cost.md, engine source, backlog reads — the merge seat's other tool
results). The table reads as a complete decomposition; the caveats disclose two other
approximations but not this one. The #1-ranked money-map lever's sizing is quoted from this
table; its undisclosed residual is the exhaustive-sweep-omits class this corpus has already
documented.

**Required fix:** add an "other files" column (or one sentence stating the residual range
and its main constituents); no figure changes — the findings-attributable and
sharding-addressable numbers ride the findings column, which is unaffected.

**found_by:** L6

## Checked and deliberately not raised

- **Board-telemetry gitignore hazard:** none — gitignore leaf-read; only the transcripts
  tarball is ignored under trajectories/. The "git-tracked" premise holds.
- **§5.5 gate condition 1's detector-ran evidence is blue self-report:** the gate decision
  is post-hoc, red's zero-regression record is independent, and the retrospective auditor
  can grep the logged sweeps — within the stated attestation ceiling; conjunction
  acceptable. Interesting, not of interest.
- **Write-preflight vs cat-append channel mismatch** (condition 6 preflights the Write tool;
  the telemetry append rides Bash `cat`): error direction is conservative only (a false
  refusal, never a false pass). Not raised.
- **R2-11's successor-id route as red-controlled re-dispatch lever:** visible and judged,
  unlike the self-report fallback it replaced — the repair holds.
- **Clause (vii) cap arithmetic and the overflow batch-docket:** marginal dispatch cost ≈ 0
  confirmed against the single-dispatch-per-round granularity (ll.247–250, prior-round
  verification stands).

## Friction

- The 25k Read cap forced three windowed reads of `blue/report.md` (~50k tokens) at this
  lens seat — the run-3 friction #15 class, recurring live at a round-3 seat. The full
  re-read MUST is honored but the tooling taxes it; same referent as §4.2's measured
  "largest merge-context component" fact.
