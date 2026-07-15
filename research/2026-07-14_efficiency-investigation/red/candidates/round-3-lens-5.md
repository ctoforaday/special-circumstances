# Round 3 — lens 5: logic and completeness

Full re-read of `blue/report.md` (all 1365 lines, three windowed reads) performed before this
pass; CHANGELOG used as navigation only. All 18 round-2 repairs re-read in context; the five
findings below are regressions or composition defects inside round-2 (and one pair of round-1)
repair text, plus one new measured-table contradiction. Leaf verification performed first-hand
this pass: `debate.js` envelope schemas ll.60–144 (JUDGE_ENVELOPE has no spot-check/log field;
RED_ENVELOPE is red-merge's return, emitted before the judge dispatch). Lens-scoped ids only;
merge assigns stable R3-N ids and lineage.

---

## L5-F1 — measured-table claim contradicted by its own table

**Location:** §4.2 "Convergent: the cost case, measured" — *"Two bonus measured facts:
**blue/report.md is the largest merge-context component every round from round 2** (145KB
ingested at round 5 — run-3 friction #15's real referent, untouched by lever 4a)"* — and the
propagated site §8 Q2: *"blue/report.md is the largest merge component every round from
round 2."*

**Problem:** the round-4 row of the same table refutes the universal: blue/report.md 20% vs
lens candidates 33% and red/findings.md 32% (absolute: 190KB × 20% = 38KB vs 63KB / 61KB).
The claim is true for rounds 2, 3, 5 only. This is the exact internal-contradiction class the
report itself flags in cost.md finding 2 (§6.4 item 1(a): a finding contradicted by its own
table) — committed in the round-2 measurement text that closed R2-3. Arithmetic on the
surrounding figures checks out (145KB = 318×46%; 52–80KB lens ingest; 60–92KB findings;
≈$12 = 1.40+2.60+4.10+4.10), so this is the one cell where the prose overran the data; the
sharding-addressable sizing does not turn on it, but a skeptic quoting the "bonus fact"
mis-cites the measurement, and the fact feeds future lever targeting (which file is the
biggest merge-cost referent).

**Required fix:** restate at both sites: "the largest merge-context component in rounds 2, 3,
and 5 (and 36% in round 1); in round 4 lens candidates (33%) and findings (32%) exceeded it
(20%)." debate.md round-2 BLUE carries the same sentence — transcript, not repairable; the
report sites are the fix surface.

**Grading:** certain (static text vs its own table) × low-medium (measurement-integrity claim
in the #1-ranked lever's evidence base; no disposition flips) × trivial. **LOW-MEDIUM.**
**Corroboration:** HIGH (table and claim quoted side by side from the audit surface).
**Suggested lineage:** regression inside the R2-3 repair → supersedes [R2-3] if merge closes
R2-3 with regression.

## L5-F2 — registered prediction no longer matches the arm conditions it tests

**Location:** §1.5 carried minority option — arm condition (b): *"**two consecutive rounds**
minted zero new gaps above the floor"* — vs the registered prediction: *"if runs 4–5 end with
a final round whose open board is all-≤-MEDIUM AND whose new-mint list is empty, the re-scoped
floor would have saved one round's spend"*.

**Problem:** two round-1 repairs to the same paragraph do not compose. R1-12 hardened the arm
to double confirmation (two consecutive zero-above-floor-mint rounds + red-health term); R1-17
restated the prediction's dollar netting but left its trigger condition single-round. Under
the hardened arm, saving the final round R requires arming at the end of round R-1, i.e.
rounds R-2 and R-1 both zero-above-floor-mint plus an all-≤-MEDIUM board at R-1 — none of
which the registered condition checks. The prediction can settle TRUE (final round zero-mint,
low board) while the hardened variant would have saved $0 (e.g. round R-1 minted a HIGH).
Two further mismatches: "new-mint list is empty" (total) vs the arm's "zero new gaps above the
floor"; and the arm routes to judge disposition which "never terminates," so "saved one
round's spend … at zero verdict cost" assumes the disposition ends the run. Direction of
error: overstates the variant's value — the registered test could falsely validate the build
trigger for a lever blue itself rejects as unproven.

**Required fix:** restate the prediction to test the arm as hardened: "if the two rounds
preceding runs 4–5's final round minted zero above-floor gaps, the pre-final board was
all-≤-MEDIUM with low/trivial fix cost, and the red-health term held, the re-scoped floor
would have armed and substituted a ~$10 judge-disposition round for the ~$25–30 final round
(~$15–20 net), provided the judge disposed rather than carried."

**Grading:** medium (the run-end condition is plausibly realized — that is why the prediction
was registered; it settles mechanically at run end) × low-medium (a registered figure feeding
the build-trigger decision settles wrong; §1's REJECT itself is unaffected) × trivial.
**LOW-MEDIUM.** **Corroboration:** HIGH (both clauses quoted from the same §1.5 paragraph).
**Suggested lineage:** composition defect between the R1-12 and R1-17 repairs (both closed
clean round 2) → supersedes [R1-12, R1-17].

## L5-F3 — the judge's demanded reads are routed to an envelope the judge does not emit

**Location:** §4.5 condition 7 — *"Condition-2 demanded reads (red-merge's and the judge's)
are logged in the same field."* — where "the field" is `RED_ENVELOPE.archive_spot_checks`
(*"`RED_ENVELOPE` gains an `archive_spot_checks` field required non-empty from round 2"*).

**Problem:** mechanically impossible as written, verified first-hand at debate.js ll.60–144
this pass: the judge speaks only through `JUDGE_ENVELOPE` (`deadlock`, `resolutions`,
`friction` — no log field), and the judge dispatch runs AFTER red-merge's envelope is already
submitted (contested docket built from `redEnv`, ll.244–250). The judge's condition-2
demanded reads (added round 2 per R2-5 — rulings on supersedes-descended gaps MUST read
ancestors' archive records) therefore have no home in "the same field": contemporaneous
logging is impossible, and no alternative (a `JUDGE_ENVELOPE` field, the ruling `rationale`,
or next-round relay) is stated. The observability condition loses exactly the judge's half —
the half R2-5 established as the lineage-dense, gate-erosion-sensitive consumer — recreating
the policy-without-mechanism class (R2-4's) for one of the two named readers, in the clause
that was rewritten to kill it. Charitable readings fail: a same-named judge-envelope field is
a schema change the spec doesn't state; next-round red relay adds a round of lag and is not
what is written.

**Required fix:** one clause: the judge's demanded reads are logged in the judge's own
channel — either `JUDGE_ENVELOPE` gains the mirror field (spec addition, stated), or each
chain ruling's `rationale` MUST name the ancestor archive records read (zero schema change;
auditable post-hoc in debate.md's ### LEAD section by the same named post-hoc auditor).
Restate "the same field" to the two named homes.

**Grading:** medium-high once 4a ships (docket arming near-certain from round 2 per §6.1;
chain rulings are the docket norm) × low-medium (unobservable judge compliance on the
`carried`-vs-`risk_accepted` ruling class §4.5 condition 2 exists to protect) × trivial.
**LOW-MEDIUM.** **Corroboration:** HIGH (schema read first-hand this pass; both quoted
clauses from the audit surface). **Suggested lineage:** regression across the R2-4 and R2-5
repairs → supersedes [R2-4, R2-5].

## L5-F4 — the write-block preflight's executor seat-class is unverified

**Location:** §4.5 condition 6 — *"the preflight MUST issue real Write calls from a live seat
in the production harness (e.g. the next run's skeleton-creation step writes both shard
files)."*

**Problem:** the R2-17 repair moves the preflight out of the stub simulator but does not pin
the executor's seat class. Every observed guard firing is at a SUBAGENT seat (run-3 red-merge,
friction #4; this run's blue-synthesize refusal; red lens seats). Skeleton creation is
plausibly a lead/orchestrator step whose permission configuration may differ from subagent
seats — if the guard is seat-class-scoped rather than global, a lead-seat preflight passes
while the first red-merge shard write still blocks mid-merge: the vacuous-preflight failure
one level up, behind the same green checkmark R2-17 named. No evidence either way exists in
the corpus (no lead-seat Write of a guarded name has ever been attempted/recorded) — the
clause asserts guard-equivalence across seat classes that nothing has read first-hand, under
a round whose stated bar was "no new mechanism claims that have not been read first-hand."

**Required fix:** one clause: the preflight writer MUST be a seat of the same class as the
future shard writer (a subagent — e.g. the first sharded run's red-merge writes both skeleton
shards as its opening act), or the guard's seat-independence must be verified first and the
verification cited.

**Grading:** low-medium (guard is plausibly a global deny rule; seat-scoped permission config
is a real mechanism in this harness) × low-medium (mid-merge discovery, the exact failure the
condition exists to prevent) × trivial. **LOW.** **Corroboration:** MEDIUM (guard behavior at
subagent seats HIGH from three live instances; lead-seat behavior unobserved — the gap is the
absence of evidence, stated as such). **Suggested lineage:** residue of the R2-17 repair →
supersedes [R2-17].

## L5-F5 — false "equivalently," and the contest window has no named reviewer

**Location:** §3.3 clause (v) contest-window clause — *"an accepted-dispute delta enters the
mass computation only after a **one-round contest window** (or, equivalently, deltas above a
stated magnitude auto-docket to the judge before actuating)."*

**Problem:** two defects in one parenthesis. (a) The two branches are not equivalent: the
auto-docket branch names its reviewing mechanism (the judge); the window branch names no seat
tasked with reviewing accepted deltas during the window — the delta's only implicit reviewer
is next-round red, i.e. the seat class that just accepted it (the conflicted-executor problem
R2-6(a) itself named for the condition-5 floor). A window nobody is tasked to use is a delay,
not a check; "equivalently" launders a mechanism-free branch with a mechanism-bearing one.
(b) Vector honesty: the window phrasing originates in red's own R2-6 required fix ("enter
mass only after a one-round contest window") — the unnamed-reviewer hole was in red's
phrasing; the "equivalently" is blue's addition. Both defects are design-spec text on an
unbuilt channel under rejected actuation — low likelihood now, rising exactly when the
interlock becomes mandatory.

**Required fix:** either drop the window branch and keep the auto-docket branch (one
mechanism, named executor), or name the window's reviewer (e.g. accepted deltas above the
stated magnitude are listed in the next round's lens dispatch prompts as review-eligible) and
delete "equivalently."

**Grading:** low this run (channel unbuilt, zero dispute traffic) rising to medium under
actuation × low-medium (the interlock's accepted-branch check is the part §3.5/§6.2 declare
load-bearing for any future actuation) × trivial. **LOW.** **Corroboration:** HIGH
(report-internal, quoted; the R2-6 fix text compared side by side — red logs the vector
against itself). **Suggested lineage:** residue of the R2-6 repair → supersedes [R2-6].

---

## Checked and deliberately not raised

- §1.2 arithmetic (~$78 = $53 + $25; ~$68 net; 21 = 10+5+6) — recomputed, holds.
- §2.4/§6.1 rescale arithmetic ($49.48/25 lens-rounds ≈ $1.98; +$2/round; ~$18/run ≈ ~10% of
  a ~$200 rescaled baseline) — recomputed, holds within stated "~".
- §4.2 remaining cells: 145KB (318×46%), 52–80KB lens ingest, 60–92KB findings, ≈$12
  cache-weighted sum, "debate.md minor (2–10%)" — all recomputed, hold. Rows not summing to
  100% (80–96%) — residual is other ingest, not a defect.
- §2.5 item 1's mapping-stability condition vs open question 6 (undecided enum mapping with
  no named decider before run 4 round 1): the "pinned OR version-stamped" disjunct makes a
  mechanical default available without deciding Q6 — interesting, not of interest.
- §6.1 ranking of carried-ruling persistence ("plausibly second-largest") at item 3 behind
  batching: defensible under the map's stated "measured target × confidence" key (batching is
  measured, item 3 is "plausibly") — not raised.
- §0 row 4a's parenthetical citing only round-1 condition changes: not false, merely
  non-exhaustive — not raised.
- Blue's choice of the merge-seat file sink over the lead-ruling's envelope-aggregation
  alternative without stated reason: the alternative is arguably unexecutable (script has no
  fs), so the choice is right; a one-clause reason would be nice-to-have — not raised.
- Template compliance: §7 self-audit, §8 open-questions union, footnote access dates,
  minority tags, confidence grades per section — all present.

## Friction (also appended to friction.md)

- The 25k Read cap forced three windowed reads of the 1365-line living report at this lens
  seat — live recurrence of run-3 friction #15 at a lens seat, supporting §4.2's own finding
  that blue/report.md is the dominant unaddressed read cost.
