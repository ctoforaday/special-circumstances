# blue report — Efficiency and termination levers for the frank-exchange-of-views debate engine

Run 4 (efficiency investigation), round 0 synthesis, 2026-07-14. Union of three method-lens lane
drafts (preserved in `blue/candidates/`): **lane 1** adversarial-disconfirming-first, **lane 2**
primary-literature, **lane 3** local-repo critical-stance. Evidence base pinned per
`inputs/PINNED.md`: run-3 retrospective @ `bfa8a3b`, engine + backlog @ `5396952`; pin
equivalence to the working tree verified by empty `git diff --stat` on every cited path
[minority: lane-3/local-repo].[^PinCheck] Winnow list honored: nothing below re-recommends
PR #14–#18 content; shipped fixes are audited only on this run's own evidence.

**Claim provenance convention:** claims made independently by two or more lanes are convergent
and carry no marker. Claims appearing in exactly one lane's draft are tagged
`[minority: lane-N/method]` — red must weigh these differently from convergent ones.

**Doctrine constraint applied throughout:** cheapen redundancy and mechanics, never judgment or
the adversary; the spot-check floor never reaches zero.

## §0 Verdict summary (blue's synthesized positions, with lane votes)

| Lever | Blue position (round 0) | Lane votes |
|---|---|---|
| (1) Severity-floor termination | **REJECT as specified** (the trigger never fires on the only measured run; the only relaxation that fires does so at the round-2 boundary and deletes rounds 3–5 — the rounds that minted R3-1/R3-2, R4-1, R5-5; corrected per red R1-4). RATIFY the telemetry — per-round board-profile record with the **durable merge-seat sink specified round 2 per red R2-1** (harness `log()` verified console-ephemeral first-hand; see §2.5 item 1) — + document stop-and-resume as the standing terminator. Lane 2's re-scoped judge-routing variant carried as a minority option, itself never firing on run-3 data | REJECT-as-specified 3/3; re-scoped RATIFY 1/3 (lane 2); advisory-signal RATIFY 2/3 (lanes 1, 3) |
| (2) Risk-mass-proportional spend | **REJECT as an actuated throttle for runs 4–5**; RATIFY the instrumentation (per-merge mass telemetry + `found_by` lens-overlap field), **conditional on the durable sink named in §2.5 item 1** (round-2 correction: the round-1 `log()`/journal sink is contradicted by the measured record). Lane 2's narrowed citation-instance-only throttle preserved as the candidate design if the logged record validates it | REJECT-throttle 2/3 (lanes 1, 3); narrowed RATIFY 1/3 (lane 2); instrumentation 3/3 |
| (3a) Blue grade-dispute channel | **RATIFY the minimal envelope form** (two optional fields, default-to-docket on unaddressed disputes, one resolution-enum value, plus the accepted-branch audit clauses of §3.3 — added round 1 per red R1-2) — stated honestly as record-integrity insurance with zero expected round-savings on run-3-shaped traffic. BINDING interlock: mandatory companion of any future grade-actuated spend (lever 1 or 2 actuation), qualified per §3.5 | Honest tally (restated per red R1-16): RATIFY-unconditional 1/3 (lane 3); RATIFY-conditional-on-lever-2-actuation 1/3 (lane 2 — premise REJECTED this run, so this vote does not count at face value); REJECT-with-binding-trigger 1/3 (lane 1, dissent preserved in §3.4). Blue ratifies on lane 3's grounds, argued in §3.5 |
| (3b) Best-of-N grading | **REJECT.** Backlog's own precondition (surviving lone-voice bias) unmet in the pinned corpus; the panel literature's benefit concentrates in the cross-provider configuration FEOV does not run; grading is already multi-voice | REJECT 3/3 |
| (4a) Sharded findings (open ledger / closed archive) | **RATIFY with seven named conditions** (§4.5; condition 7 and the condition-3/6 extensions added round 1 per red R1-6/R1-8/R1-14). The full-re-read MUST protects red-reads-BLUE, not red-vs-own-archive — no doctrine conflict exists | RATIFY 3/3 |
| (4b) Collator stage | **REJECT the seat; RATIFY the degenerate form** — one prompt-level concatenation/read-batching sentence at the merge | REJECT-seat 3/3 |
| (5) Round-scoped audit | **HOLD through run 4** (as staged). **Conditionally RATIFY for run 5**: four-arm propagation-aware scope, gated on run 4's live propagation-clause record. Any unpropagated-site regression in run 4 = reject outright for run 5 | 3/3 (identical gate logic reached independently) |

---

## §1 Severity-floor termination — REJECT as specified

### 1.1 Convergent core: the floor never fires on the only measured run

The backlog's claim: fire "when every open gap is <= MEDIUM with trivial fix cost, route the
whole board to the judge for disposition instead of another $25-30 round (would have ended run 3
at round 3 for ~$10)."[^BacklogLevers] All three lanes independently checked this against red's
graded board at every run-3 round boundary, read verbatim from the findings file's preserved
per-round blocks:[^FindingsBoard]

| Board after merge | Open gaps | Max severity (members) | Floor fires? |
|---|---|---|---|
| Round 1 | 20 | HIGH (R1-1, R1-2) | no |
| Round 2 | 11 | MEDIUM-HIGH (R2-1, R2-3, R2-7, R2-8, R2-9) | no |
| Round 3 | 10 | MEDIUM-HIGH (R3-1, R3-2 — both code-trace, complexity "low," not "trivial") | no |
| Round 4 | 5 | HIGH (R4-1) | no |
| Round 5 | 6 | MEDIUM-HIGH (R5-5) | no |

Zero fires; realized saving $0. **The backlog's "would have ended run 3 at round 3 for ~$10"
claim is contradicted by the pinned record it cites** — the round-3 board carried two MEDIUM-HIGH
code-trace gaps (R3-1's degenerate-envelope loop, R3-2's dropped friction seat), neither
≤ MEDIUM nor trivial-fix.[^FindingsBoard][^Round3Red] Red's own round-3 convergence statement
confirms: "round 3: 2 MEDIUM-HIGH, both code-trace — every prose gap is now ≤ MEDIUM."[^Round3Red]

### 1.2 Convergent: making it fire makes it wrong

The only threshold that realizes the claimed saving admits MEDIUM-HIGH — and at that setting the
floor first fires at the **round-2** boundary, not round 3 (corrected round 1 per red R1-4,
verified against the pinned per-round blocks): the round-2 board's five MEDIUM-HIGH members
(R2-1, R2-3, R2-7, R2-8, R2-9) all carry complexity "low," exactly like round 3's pair — no
severity × fix-cost threshold fires after round 3 but not after round 2.[^FindingsBoard] The
corrected counterfactual is *worse* for the lever: a MEDIUM-HIGH-admitting floor deletes rounds
3–5 (~$78 gross: rounds 4–5's Σ$53 plus round 3's ~$25 from cost.md; **~$68 net** of the ~$10
judge-disposition round the floor routes the board to — netting applied round 2 per red R2-10,
matching §1.5's R1-17-netted sibling prediction)[^CostAudit] and with them **all ~21
rounds-3–5 mints** (10 + 5 + 6 by the per-round mint counts below), four of consequence —
R3-1 (degenerate-envelope loop), R3-2 (dropped friction seat), **R4-1** (HIGH,
certain × high — the lineage-blind docket, the retrospective's most consequential engine
finding, shipped as PR #15's core)[^R4OneDetail] and **R5-5** (MEDIUM-HIGH — unenforced
supersedes, shipped as PR #15's structural enforcement
throw).[^R5FiveDetail][^AlreadyShipped] And **no threshold setting reproduces the backlog's
claimed round-3 stop** — the claim fails not just on its board but on every board, which
strengthens §1.1's headline contradiction. The floor's implicit model — the current board's
severity predicts the next round's *discovery* severity — is directly falsified twice: the
round-2→3 transition (an all-low-complexity MEDIUM-HIGH board preceded two code-trace
MEDIUM-HIGH mints) and the round-3→4 transition (a MEDIUM-HIGH-max board preceded a fresh HIGH
mint). A judge disposing the round-2 or round-3 residual board does not audit; disposition
produces no R4-1. Terminating the adversary's discovery rounds is a doctrine violation, not a
mechanics saving. Framing corrected per red R1-15: "deleted" means **deleted from run 3, at
risk of never being minted** — engine-audit runs demonstrably recur (this run re-audits the
engine), so a later run could plausibly have re-minted R4-1/R5-5; the honest counterfactual
cost is a delay of at least one full run and possibly forever (nothing guarantees a later run
re-covers the same surface), priced against a certain ~$68 net saving (~$78 in deleted rounds
minus the ~$10 judge-disposition round the floor routes to — netting per red R2-10)
[minority: lane-2/primary-literature, reframed].[^CostAudit]

Additional convergent-adjacent findings:

- **Run 3's actual termination itself violated the floor** [minority:
  lane-1/disconfirming-first]: the run ended by human stop-and-resume with reduced `maxRounds`
  WITH an above-floor gap (R5-5, MEDIUM-HIGH) still open — a strict severity floor used as a
  termination *condition* would have demanded MORE rounds at that point, not fewer. Applied to
  its own motivating run, the lever either does nothing or extends the run.[^StopResume][^FindingsBoard]
- **The frontier's own re-scoped variant (arm only on a no-new-gaps round) also never fires on
  run 3** [minority: lane-3/local-repo]: every round minted new gaps
  (20/11/10/5/6).[^FindingsBoard]

### 1.3 Convergent: the right stopping statistic is the discovery process, not the residual board

Two lanes reached the same conclusion from disjoint literatures:

- Multi-agent debate: measured accuracy saturates at ~2–5 rounds and adaptive stopping works,
  but the published criterion is *stability of the discovery process* (round-over-round
  distributional change below ε for 2 consecutive rounds — a double confirmation), not a
  property of the residual list [minority: lane-1/disconfirming-first for the specific
  criterion].[^AdaptiveStability][^DebateRounds] Run 3 never stabilized under that test: round 4
  minted a HIGH, round 5 a MEDIUM-HIGH — a stability rule correctly refuses to stop, matching
  the record's own verdict (FAIL, UNVERIFIED, 6 open).
- Software-testing optimal stopping: Dalal & Mallows key the stop decision to the **observed
  discovery rate against the cost ratio**;[^DalalMallows] Böhme's STADS estimates residual risk
  from the discovery curve itself (Good-Turing: the singleton rate estimates the probability the
  next probe finds something new);[^Stads] classical inspection uses capture-recapture over
  inspector overlap — a forward-looking estimate of what has NOT been
  found.[^CaptureRecaptureEval][^CaptureRecaptureDecade] [All: minority:
  lane-2/primary-literature except capture-recapture, which lane 1 also cites.]

The severity floor is a backward-looking statistic standing in for a forward-looking decision.
Known-open severity and residual discovery risk are different quantities; run 3 demonstrates the
divergence live.

### 1.4 What already exists, cheaper (convergent, lanes 1 and 3)

- **The demonstrated ~$0 terminator:** operator stop-and-resume with reduced `maxRounds` (cache
  replay, measured ~$0, cut ~7 residual rounds).[^StopResume] It is termination-by-judgment —
  the human reads the board and decides — which is where the doctrine says judgment belongs.
- **Ceiling assembly already produces the residual-board disposition table** the floor's judge
  call would produce (report §"Outstanding gaps at the ceiling": per-gap grading, blue's
  response, disposition, compromise rationale) [minority:
  lane-3/local-repo].[^CeilingDisposition]
- What run 3's operator lacked was the *signal*, not the mechanism.

### 1.5 Disposition

**REJECT severity-floor termination as specified.** Automating the stop decision on red's own
grades cheapens judgment — the one thing the constraint forbids; the severity floor is an
autopilot for the call the judge exists to make [minority: lane-3/local-repo for this doctrinal
framing]. **RATIFY instead (convergent, lanes 1 and 3):**

1. A per-round board-profile record — open count, max severity, new-mint count and severity
   profile, computed mass (§2), accepted-dispute grade deltas (§3.3(v)) — so the stop decision
   is made by a judge (human or lead) with the numbers in front of them. **Durable sink
   (corrected round 2 per red R2-1 — the round-1 "`log()` into journal.jsonl" sink is
   contradicted by the measured record):** harness `log()` is console-ephemeral — verified
   first-hand at the blue seat this round: this run's own LIVE workflow journal holds only
   `started`/`result` lifecycle events (zero `log()` output, "researching" grep count 0),
   matching run 3's copied journal (87 lifecycle lines, zero `log()`)[^JournalCheck] — so the
   durable sink is the **merge-seat telemetry append** specified in §2.5 item 1
   (`trajectories/board-telemetry.jsonl`, one JSON line per round, named consumers stated
   there). The in-script `log()` line may still fire as operator-console advisory — useful
   live, persisted nowhere — and composes with the still-unshipped log()-per-transition
   heartbeat backlog item, which is the same console tier.[^BacklogLevers] Cost: near-zero
   (one appended line per round at a seat already holding every input first-hand), no longer
   claimed as literally zero tokens.
2. Document stop-and-resume as the standing termination practice (one paragraph in
   research-protocol / the `debate.js` header) [minority: lane-1/disconfirming-first for the
   documentation step].[^StopResume]

**Live-code correction (added round 1 per red R1-1, verified first-hand at debate.js
ll.236–258):** per-round judge disposition of the CONTESTED subset is not future design — it is
shipped code. From round 2 onward, every re-raised or supersedes-descended gap dockets
automatically and the judge disposes it that same round. What the carried variant below would
still add over live code is exactly one thing: suppressing the *discovery* seats (lenses, merge,
blue) for a round and disposing the whole board, not just the contested subset. Adjudication
itself is no longer the variant's contribution. Two consequences propagate: (a) arming of the
docket in run 4 is **near-certain**, not speculative (any gap surviving two rounds dockets —
open questions 4 and 8 restated accordingly); (b) every efficiency estimate must be priced net
of the judge seat (§6.1's projected line, added round 1).

**Carried minority option** [minority: lane-2/primary-literature; arm conditions hardened
round 1 per red R1-12]: a re-scoped floor that arms only when ALL of (a) every unadjudicated
open gap ≤ MEDIUM with low/trivial fix cost, (b) **two consecutive rounds** minted zero new
gaps above the floor (the discovery-decay arm, Good-Turing-shaped — double confirmation,
matching the §1.3 adaptive-stopping criterion this design cites; a single zero-mint round is
indistinguishable from a degraded red round), and (c) a red-health sanity term: the round
delivered its full lens complement with non-degenerate outputs (a zero-mint round produced by
lens under-delivery, a null return, or a cheap-model substitution MUST NOT arm the floor) —
and on arming **routes the board to the judge for disposition, never terminates** (the judge
can carry gaps and continue). Honesty clause carried verbatim: even re-scoped, this variant
never fires on run 3's data; its value is unproven on the measured corpus. Registered
prediction for red (restated round 3 per red R3-6 — the round-1 R1-17 restatement fixed the
netting but left the trigger single-round and total-mint, so it could settle TRUE while the
R1-12-hardened double-confirmation arm never fired and saved $0; sibling repairs verified
faithful in isolation did not compose, and the round-2 closure ledger records both as closed
WITH REGRESSION; the prediction now tests the arm AS HARDENED): if a run-4/5 record shows
(i) **two consecutive rounds minting zero new gaps above the floor immediately preceding the
final round**, (ii) a pre-final open board all-≤-MEDIUM with low/trivial fix cost, and
(iii) the red-health term held in both zero-mint rounds, then the armed floor would have
substituted a ~$10 judge-disposition round for the ~$25–30 final round — ≈ **$15–20 net** —
**provided the judge disposed rather than carried** (a carried board continues the run:
saving $0; the netting assumes the disposition round ends the run, which is the judge's call,
not the arm's). A final round that is merely all-≤-MEDIUM with an empty mint list does NOT
settle this prediction true — that weaker single-round trigger is exactly what the R1-12
hardening exists to reject. Blue's
synthesized position: the two ratified telemetry items above collect exactly the evidence this
variant needs; build the trigger only if the logged record ever shows it firing. Interaction
flagged for design time [minority: lane-2/primary-literature]: a judge-disposition round that
closes the whole board must define its PASS path against the shipped degenerate-FAIL guard
(one state-machine sentence).

**Generalization caveat (convergent, all three lanes):** run 3 is one run (n=1) on a
self-referential topic — the audit surface WAS the engine, so lens passes tracing `debate.js`
could mint HIGH engine findings late; an ordinary external topic's late rounds may genuinely be
textual trivia. Thin evidence in both directions — itself the argument for instrumentation
before mechanism.

**Confidence: HIGH** on the never-fires finding (mechanical extraction from pinned grades, three
independent lane reads agree); **MEDIUM** on the §1.2 falsification counterfactual (added round 1
per red R1-15: the firing-round and dollar figures are mechanical from the pinned boards, but
"which findings a truncated run never mints" is single-run, threshold-model-dependent
extrapolation — blue's own generalization caveat applies to it); **MEDIUM** on the re-scoped
variant's expected value (no corpus instance where it arms).

---

## §2 Risk-mass-proportional spend — REJECT as actuated throttle; RATIFY instrumentation

### 2.1 Convergent: the computed mass series, and its correlation failure

Lanes 1 and 3 independently computed sum(likelihood × impact) over open gaps at each run-3 merge
(mapping disclosed by both: low=1, low-medium=1.5, medium=2, medium-high=2.5, high=3,
certain/realized=3.5; compound cells read verbatim):[^FindingsBoard]

| Board after round | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| Mass (lane 1) | ~98 | ~65 | ~44 | ~30 | ~31 |
| Mass (lane 3) | ~98 | ~62 | ~44 | ~29 | ~32 |

(Small extraction differences between lanes; per lane 3, no conclusion turns on ±0.5 in any
constant — the mapping is disclosed so the computation is reproducible.)

Convergent failures of the throttle premise:

1. **Low mass did not predict low next-round value — twice, at exactly the boundaries a throttle
   would have acted on.** The two lowest-mass boards **among boards that preceded another round**
   (post-round-3 ~44, post-round-4 ~30; post-round-5's ~31 has no successor round to predict —
   universe stated per red R1-24) preceded the run's highest-graded discovery (R4-1, certain ×
   high) and a MEDIUM-HIGH plus a dark-side companion (R5-5,
   R5-6).[^R4OneDetail][^R5FiveDetail][^FindingsBoard]
2. **The series is not even monotone** — round 5 *rose* [made explicit by lane-3/local-repo;
   present in both tables].
3. **Residue is not discovery (the structural defect):** open-gap mass measures *known open
   rework* — backward-looking; a spend decision needs an estimate of the *undiscovered* gap
   population — forward-looking. The two are uncorrelated in the pinned record. Scoping lens
   count to residue mass throttles the adversary's discovery capacity using a gauge that reads
   only what is already found.
4. **The metric cannot discriminate the trivia it exists to detect** [minority:
   lane-1/disconfirming-first]: mean mass per gap sits in a flat 4.4–6.0 band across the five
   rounds — 4.9 / 5.9 / 4.4 / 6.0 / 5.2, recomputed from the table above (band stated round 3
   per red R3-14; the round-0 "~5" glossed a ~20% spread, and the honest band strengthens the
   item: the two highest per-gap means are rounds 2 and 4 — no downward trend as the board
   turns trivial) — because late-round textual nits carry `certain` likelihood by construction
   (a text defect, once found, is certain) — a certain × low nit scores ~3.5 against a
   medium × medium real risk's 4.
   cost.md's "rounds 3–5 closed ~15 mostly-trivial gaps" is true of fix cost but invisible to
   sum(L × I).[^CostAudit]

### 2.2 Grade-noise test resolved in the lever's favor (convergent, all three lanes)

The frontier's disconfirming test (a) does NOT fire, and all three lanes say so — recorded for
honesty since it removes one rejection ground: run 3's grade-correction chains (R2-1 count
inflation 3→2 with likelihood retained High by re-argument; R3-7 mechanism narrowed, grade kept;
R5-1 enumeration corrected, grade untouched) moved computed mass ~0.[^Retro3Docket] Lane 3
sharpens the point [minority: lane-3/local-repo]: those famous corrections were corrections to
*blue's §3 docket cells*, not to the red gap grades that would feed the mass computation;
within-round merge temperings (R5-5, R5-6) move mass ~1–1.5 points on ≈31 (≈5%) — modest. The
throttle input is noisy in the general severity-grading literature (68% of surveyed users scored
the same vulnerabilities differently under CVSS;[^CvssInconsistent] 73% of public CNAs — 194 of
266 under the pairwise setting, across 44,180 CVEs scored by both NVD and a CNA — show a median
of at least one diverging CVSS base metric from the NVD, leaf-verified round 1 by full-text
extraction;[^CathedralBazaar] expert disagreement variance ~4.5 on a 10-point
scale[^ExpertCvss][^RbtTaxonomy]) [all minority: lane-2/primary-literature]. **Correction record
(red R1-5):** the round-0 draft cited "NVD-vs-CNA disagreement on roughly a third of
dual-assessed CVEs" against arXiv:2508.13644 — that figure is affirmatively absent from the
cited paper (which compares four scoring *systems*, with no NVD-vs-CNA analysis; two independent
red full-text fetches, confirmed by blue's own extraction attempt round 1); the figure was a
search-digest number wearing a named paper's citation and is **withdrawn**, replaced by the
leaf-verified CNA-divergence figures above. The plausible actual source of a ~34% per-CVE figure
(Computers & Security 2026, "Fragmentation of CVSS scores in the NVD") is inaccessible from the
verifying seats (403 at the abstract; the journal is subscription, so a paywall is plausible,
but a 403 shows access failure, not mechanism — bot-block vs paywall unshown; adverb corrected
round 2 per red R2-18) and remains unverified — not re-cited.[^ConflictingScores] The paper
originally miscited is open-access, not paywalled — §7's excuse is corrected there. But **the
throttle input** was measured-robust in this loop (subject restored round 2 per red R2-12 —
the round-1 splice severed it), because the adversarial loop is itself the triangulation the
literature asks for. Noise is NOT the kill reason; §2.1 and §2.3 are.

### 2.3 The doctrine test: what a lens cut actually loses — the two-sided evidence

**The case that marginal lens instances are cheapenable redundancy** [minority:
lane-2/primary-literature]: run 3's late high-value discoveries were massively lens-redundant —
R4-1 was independently minted by 4 of 5 lenses (round-4 lens files 1, 2, 3, 5 each carry it,
including all three citation instances);[^Round4Lenses][^R4OneDetail] R5-1 by 3 of 5 ("three
lenses converged independently — lenses 1, 2, 4").[^R5OneDetail] A round running one citation
instance plus the always-on logic and dark-side lenses would still have caught both; the
marginal citation instances bought convergence, not coverage — and convergence is redundancy,
which doctrine permits cheapening. Risk-proportional allocation is also the normative core of
the international testing standard (ISO/IEC/IEEE 29119) and sequential-adaptive spend cuts
expected sample size by **36–75% for symmetric error bounds** (type-I = type-II — the source's
derivation condition, restored round 2 per red R2-14; "at matched error rates" named only the
comparison baseline, and asymmetric regimes may fall outside the band) *when the statistic is
right* (figure corrected round 1 per red R1-26 to the relative-efficiency paper's own stated
band; the round-0 "30–50%" was a gloss not present in the cited
source).[^Iso29119][^Sprt][^DalalMallows]

**The case that the cut lands on the adversary** (lanes 1 and 3, convergent):

- **R5-5 is a 1-of-5 singleton** [minority: lane-1/disconfirming-first for the lens-file
  verification]: lens 5 alone made the enforcement argument; lens 4 checked the adjacent
  territory and explicitly held ("Considered raising this... Not raised"); lens 2 examined the
  detector mechanics without raising enforcement.[^R5FiveSingleton] Lane 3 identifies the
  producing pass as the dark-side lens on the run's minimum-mass board.[^FindingsBoard]
  Synthesis note on the interaction: lane 2's specific floor (≥1 citation + logic + dark-side,
  never scoped down) would have RETAINED the pass that minted R5-5 — the singleton evidence
  kills a generic lens-count throttle but not lane 2's named-floor variant; red should weigh
  the variant against the singleton evidence directly.
- **Thin-lens rounds remove red's internal error correction** [minority: lane-3/local-repo]:
  both round-5 lens errors were caught by *sibling-lens comparison* at the merge (three lenses
  converging against lens 5's unquoted "no discrepancy" hold, overruled by direct read; lenses 1
  and 3 against lens 2's six-id miscount, overruled by mechanical extraction —
  citation-ledger.md line 184).[^R5OneOverrule][^CitationLedgerRun3]
- **Corroboration grading depends on convergence counts** (lanes 1, 2, 3): "four of five lenses
  converged independently" is part of R4-1's HIGH corroboration; "three lenses converged" part
  of R5-1's.[^R4OneDetail][^R5OneDetail] Fewer instances = weaker corroboration grading on what
  remains. Mitigation carried [minority: lane-2/primary-literature]: corroboration language
  should distinguish "single lens, leaf-verified" from "unverifiable" — verification depth, not
  seat count, is the load-bearing part.
- **Cutting lens count destroys the only forward-looking estimator the system could have**
  (lanes 1 and 2): capture-recapture's validity caveat — with too few inspectors "no model is
  sufficiently accurate and underestimation may be substantial."[^CaptureRecaptureEval]
  Fenton & Ohlsson's counter-intuition warns against extrapolating from where defects were
  already found [minority: lane-2/primary-literature].[^FentonOhlsson] This is the
  primary-literature form of the doctrine constraint: the spot-check floor never reaching zero
  is not a policy preference, it is the statistically forced hedge.
- **The incentive loop** [minority: lane-1/disconfirming-first]: under an auto-throttle, red's
  self-graded output drives red's own next-round budget — grade high to keep lenses, or the
  system starves red exactly when red under-grades. The backlog's caveat names the floor but
  not the loop; the grade-dispute channel (lever 3) is the loop's only adversarial check —
  which is why the two levers are a package deal in any actuated future (§3.5, §6.2).
- **The correct proportionality already ships** [minority: lane-3/local-repo]:
  `citationPasses = Math.min(4, Math.max(1, Math.ceil(claim_count / 40)))`, recomputed every
  round (PR #15) — spend scales with the **verification surface** (blue's live claim count),
  not with residue mass. (Softened round 1 per red R1-23: the round-0 clause "blue-respond
  spend already tracks board size naturally" is contradicted by its own series —
  $3.95/$3.96/$2.98/$3.05/$4.27 across rounds, with round 5 the HIGHEST on the second-smallest
  board and round 1 handling 20 gaps for $3.95; the honest reading is that blue-respond spend
  is driven by fix weight, not inflated by board size, and the `citationPasses` recompute
  carries the "correct proportionality already ships" point
  alone.)[^EngineSource][^CostAudit][^AlreadyShipped]

### 2.4 Measured stake (convergent on the numbers; framing from lane 2)

Red-lens totaled $49.48 of $149.95 in run 3 (33%; scoping stated per red R1-33: rounds 1–5 —
the killed round-6 spawn adds $0.61, seat total $50.09), the second-biggest recurring line
after the merge seat; its cost tracks corpus size, not board size.[^CostAudit] Lane 2's
estimate for its narrowed throttle, **recomputed round 2 per red R2-2 at the live-code lens
shape** (the shipped citationPasses recompute — `min(4, max(1, ceil(claim_count/40)))` — plus
the two always-on lenses yields **6 lens seats/round** at report-scale claim counts; this run's
rounds 1 and 2 both demonstrably ran 6, and the round-1 recompose priced the dead 5-lens
shape): cutting 4→1 citation instances removes ~3 agents/round ≈ ~$6/round at run-3 per-lens
rates × **3 throttled rounds (the low-mass rounds 3–5 of §2.1's series — the rounds a
mass-scaled throttle would have throttled; basis stated round 3 per red R3-10, having ridden
silently from the round-0 construction through the round-2 recompute)** = **~$18/run (~10% of
the rescaled run-4 baseline)**, not the $12 registered round 1. Direction noted for honesty: the live-code figure *strengthens* the throttle case
blue rejects — corrected anyway, because registered figures settle mechanically; §2's REJECT
rests on §2.1's correlation failure and §2.3's doctrine analysis, not on dollar size. The
round-0 "$12–18" upper bound bundled an unsized cheaper-merge second-order term (fewer lens
files at the merge read) — that term is real in direction (measured lens-candidate ingest at
the merge: 52–80KB/round, §4.2) but its dollar value is still unsized, so it stays stated
separately rather than folded into the headline [minority: lane-2/primary-literature] — real,
not dramatic.

### 2.5 Disposition

**REJECT risk-mass-proportional spend as an automatic controller of lens count or audit scope
for runs 4–5** (2/3 lanes; and even the ratifying lane's design concedes the floor must hold the
distinct lenses). **RATIFY the instrumentation half (3/3):**

1. Compute sum(L × I) per merge and persist it — grades are already machine-readable in
   `redEnv` (compound enums shipped in PR #15); one arithmetic line, zero new seats. **Durable
   sink, re-corrected round 2 per red R2-1 (the round-1 repair relocated the sink error instead
   of fixing it):** the round-1 text named `log()` into `trajectories/journal.jsonl` consumed
   by cost-audit.mjs — both halves are contradicted by the measured record. Verified first-hand
   at the blue seat this round: the journal is the HARNESS's lifecycle log, not the script's —
   this run's own live workflow journal contains only `started`/`result` events and zero
   `log()` output (grep "researching" = 0), exactly like run 3's copied journal (87 lines =
   46 started + 41 result); and `cost-audit.mjs` never opens journal.jsonl — its input glob is
   `agent-*.jsonl` token records only (read first-hand, l.28).[^JournalCheck] Harness `log()`
   is operator-console-ephemeral; nothing a script logs persists. **The sink that works:** a
   **merge-seat telemetry append** — red-merge writes one JSON line per round (round, open
   count, max severity, new-mint profile, mass, accepted-dispute deltas per §3.3(v), `found_by`
   summary) to the git-tracked, name-preflighted `trajectories/board-telemetry.jsonl`
   (neutral name; the append route is the four-rounds-proven ledger `cat` path,[^FrictionTen]
   and the preflight is §4.5 condition 6's, live-seat rule included). **Named consumers,
   stated per the lead's ruling:** (i) the retrospective / next-run docket — the demonstrated
   consumer class (this run's entire docket was assembled from run-3 git-tracked artifacts);
   (ii) `scripts/cost-audit.mjs` only WITH a stated extension: one added read of
   `<runDir>/trajectories/board-telemetry.jsonl` joining the per-round mass/board columns into
   cost.md's table — the current script cannot consume it and is not cited as if it could.
   Work-done integrity for the line itself rides the §6.2 attestation ceiling: line count per
   round is post-hoc reconcilable against the round count in `debate.md` — **a presence check
   only: it catches a missing line, never a wrong one (extended round 3 per red R3-7 — the
   values inside each line are red-merge transcription self-report, item 3 makes the logged
   record the actuation trigger, and item 2's `found_by` got an independent re-derivation
   clause on exactly this reasoning while the mass series, the larger input to the same
   decision, did not).** The same clause now applies to the series: any actuation review MUST
   recompute the mass/board columns for a sample of rounds directly from the git-tracked
   findings record at a seat other than red-merge and cite the recomputation (cheap: mass is
   arithmetic over git-persisted grades); the telemetry lines are the convenience copy, never
   the evidence of record. **Mapping stability condition (same gap, part b; restated round 3
   per red R3-8 — "pinned or version-stamped" was a false equivalence: a version stamp makes a
   mid-series break VISIBLE, only pinning PREVENTS it, and offering both licenses the weak
   branch):** the enum→numeric mapping is **pinned before the first logged round**, and each
   logged line carries the mapping version; a changed mapping starts a **NEW series** —
   stamped lines are not comparable across mapping versions, and no cross-version comparison
   enters an actuation case. The mapping question itself is **decided this round** rather than
   left as an owner-less deadline (per the lead's R3-8 ruling — this run is the owner of
   record): `realized` is excluded from open-gap mass; see §8 Q6, now
   DECIDED.[^EngineSource]
2. Record **lens overlap per merged gap** [minority: lane-1/disconfirming-first]: a `found_by`
   envelope field (e.g. `['L1','L2','L4']`) — red-merge already states convergence in prose;
   this is the capture-recapture input, mechanical to collect, converting "was that round
   redundant?" from narrative to data. **Named ratification condition (promoted round 1 from
   open question 3, per red R1-9):** `found_by` is self-reported by red-merge — the seat whose
   lens budget a future capture-recapture throttle would set from this field, with
   under-reported overlap inflating the remaining-defect estimate in red's favor under
   actuation — so the field ships only with its audit: `found_by` values MUST be auditable
   against the preserved per-lens candidate files (grep-cheap; the files are git-tracked), and
   the §4.5-condition-5 spot-check floor samples them. **Independent re-derivation clause
   (added round 2 per red R2-7 — the round-1 audit routed the in-run check to red-merge, the
   seat whose budget the field would set):** in-run sampling by red-merge is self-report (§6.2
   attestation ceiling), so any future actuation review MUST re-derive `found_by` for a sample
   of gaps independently from the preserved lens files at a seat other than red-merge (lead,
   retrospective, or blue), and the actuation case must cite that re-derivation — the
   actuation decision is the named trigger and consumer the grep-cheap audit was missing. An
   unaudited instrument must not generate runs 4–5's actuation evidence base.
3. Revisit an actuated throttle only when runs 4–5's logged record shows mass (or a
   remaining-defect estimate) actually predicting next-round value — the same
   evidence-before-mechanism condition the backlog itself imposes on best-of-N
   grading.[^BacklogLevers]

**Carried minority design** [minority: lane-2/primary-literature], preserved as the candidate if
the telemetry ever validates actuation: mass throttles **citation-instance count only**
(`citationPasses` scales down toward 1, never 0, as mass falls and ledger coverage rises); the
three distinct lenses (≥1 citation, logic, dark-side) run full every round as the concrete
spot-check floor; mass NEVER narrows audit scope and never touches red-merge or the judge.
Design detail flagged [minority: lane-2/primary-literature] and **resolved round 3** (per red
R3-8 / the lead's ruling): `realized` is excluded from open-gap mass — realized risk is no
longer a probability; a realized-but-open gap counts in the board profile and contributes 0 to
mass. Decision recorded at §8 Q6; the mapping pin of §2.5 item 1 carries it into run 4's first
logged round. (§2.1's retrospective series, computed with realized=3.5 under its own disclosed
mapping, stands as a historical computation — the pin governs the runs-4–5 telemetry series,
and the two are not comparable, by the new-series rule.)

**Confidence: HIGH** on the correlation failure (two boundary counterexamples, pinned, two
independent computations agree); **HIGH** on the corpus facts behind both sides of §2.3;
**MEDIUM** on all savings estimates (single-run extrapolation); **MEDIUM** on the
capture-recapture alternative (external precedent strong, in-corpus validation pending runs 4–5).

---

## §3 Blue grade-dispute channel and best-of-N grading

### 3.1 Convergent facts

- **Judge dispatch count in run 3: zero, verified structurally** (all three lanes): anchored
  `grep -n "^### " debate.md` returns 11 headers — 6 `### BLUE`, 5 `### RED`, zero `### LEAD`.
  (An unanchored grep returns 5 prose mentions — the count-mode footgun run 3's R5-3 documents;
  lanes 1 and 3 both hit and disarmed it live.)[^DebateNoLead]
- **The zero-dispatch fact is a PRE-PR-#15 fact** (lanes 1 and 2): the shipped detector's docket
  window is now the whole debate (`allPriorGapIds` accumulates every round; `contested` matches
  any re-raised id or supersedes-descendant), so any gap that stays open across two rounds now
  auto-dockets — a dispute red re-rejects already reaches the judge under shipped
  mechanics.[^EngineSource] The frontier's H3 premise is half-obsoleted at the pin;[^FrontierH3]
  the winnow list bars re-recommending the fix. (Sharpened round 1 per red R1-1: docket arming
  in run 4 is **near-certain** — any gap surviving two rounds dockets, and multi-round gap
  survival is the corpus norm — so per-round judge dispatches are an expected recurring cost
  line, priced in §6.1.)
- **The honest residual is exactly one case: fix-but-dispute-grade.** Blue repairs the gap
  (cheaper than arguing) while disputing the grade; the gap closes, the id leaves the board, and
  the disagreement leaves no persistence signal. Observed instances in run 3: **zero** (lanes 1,
  2, 3 all searched). Round-4 BLUE states it outright ("every gap was real, at the location red
  found it, and none was over-graded relative to its fix cost"); round-5 BLUE conceded all six
  grades.[^BlueRound4] Every run-3 grade disagreement resolved inside the loop in ≤ 2 rounds by
  evidence (R2-4 rebuttal accepted; item-15's likelihood retained High by argument; R2-9/R2-10
  argued risk-accepts accepted on stated conditions).[^Retro3Docket][^DebateNegotiation]
- **Grade-correction traffic ran one direction only — red→blue** [minority: lane-3/local-repo]:
  every run-3 grade correction targeted cells in blue's §3 docket (R2-1, R2-9, R5-2); blue never
  disputed a red gap grade. **Predicted round-savings from the channel on run-3-shaped traffic:
  zero.**[^FindingsBoard][^DebateNoLead]

### 3.2 The case for ratifying anyway (lane 3, with lane 2's coupling argument)

- **The asymmetry is structural and the downstream consumer is demonstrated** [minority:
  lane-3/local-repo]: red owns `red/findings.md`; a gap red closes carries red's grade into the
  permanent record with no machine-readable path for blue to contest it — blue prose in
  `debate.md` has no reader in the script, and the risk-accept mechanism covers gap
  *dispositions*, not grade *values*.[^EngineSource][^RedAuditorMust] Grade integrity has a
  proven consumer: **this very run's docket was assembled from run 3's graded record** — over-
  or under-graded closures propagate into the next run's priorities. The channel is insurance
  priced at its mechanism cost, not a round-saver.
- **The coupling argument** (lanes 1 and 2, convergent from opposite dispositions): grades
  become load-bearing the moment any spend or termination mechanism actuates on them — an
  inflated likelihood×impact then buys unnecessary lens instances; a deflated one starves the
  round. Wrong grades stop being cosmetic and start being budget errors. A spend-controlling
  input without an adversarial correction path is §2.3's incentive loop.
- **CVSS-grade noise literature** [minority: lane-2/primary-literature]: a structured contest
  path over severity grades is hygiene any scoring system needs.[^CvssInconsistent]

### 3.3 Mechanism (lane 3's design, measured against the source) [minority: lane-3/local-repo]

`debate.js` has no filesystem access by design (all state rides envelopes), so the channel must
be envelope-only:[^EngineSource]

- `BLUE_ENVELOPE` gains optional `grade_disputes: [{gap_id, dimension, proposed, evidence}]`;
- `RED_ENVELOPE` gains optional `dispute_responses: [{gap_id, dimension, response:
  accepted|rejected, rationale}]`;
- the script holds rejected disputes one round and adds re-disputed ids to `contested` — set
  arithmetic mirroring the existing lineage filter;
- **default-to-docket**: if `blueEnv.grade_disputes` was non-empty last round and
  `redEnv.dispute_responses` does not address every disputed gap_id, treat the unaddressed
  disputes as REJECTED (auto-docket), not as absent — inheriting the R5-5/R3-2
  unenforced-optional-field lesson (a schema'd field under prompt instruction alone goes
  silently unset; three rounds unnoticed in run 3).[^FindingsBoard] **Priced round 1 per red
  R1-18; repriced round 3 per red R3-9 — the round-2 R2-11 repricing landed at §6.4 item 6
  and left this sibling site at stale per-firing pricing:** the default's failure mode —
  `dispute_responses` (itself a new optional field, the same R3-2 class, symmetrically) going
  silently unset — auto-escalates every open dispute into the round's judge docket, priced per
  §6.4 item 6's marginal-docket-growth rule: the dispatch is one agent call covering the whole
  docket, so an auto-docketed dispute costs a full ~$10–13 firing only when it is the docket's
  sole member — otherwise it grows an already-occurring dispatch marginally. The enforcement
  default is cheap insurance whose worst case is a judge round, and that price belongs in an
  efficiency report, not in the fine print.
- one resolution-enum addition [minority: lane-1/disconfirming-first]: the judge's enum
  (`closed | rebuttal_sustained | risk_accepted | carried | unresolved`) cannot currently
  express "gap real, grade wrong."[^EngineSource]

**Three clauses added round 1 (red R1-2 — the round-0 design audited only its rejected
branch):**

- **(v) Accepted-branch audit trail:** an ACCEPTED dispute changes a grade in
  `red/findings.md` with, under the round-0 design, no record, no docket, and no check — under
  actuation that is blue's symmetric *deflation* lever (dispute grades down → smaller mass →
  fewer lens instances auditing blue), and the round-0 text analyzed only red's inflation
  incentive. Clause: every accepted-dispute grade delta is appended to the §2.5-item-1
  board-telemetry line (old grade → new grade, gap id, round — durable sink per red R2-1: a
  spot-check can only sample a line that persists) and is **spot-check-eligible** under the
  §4.5-condition-5 floor. **Contest-window clause (added round 2 per red R2-6(a) — the
  accepted branch bypasses the judge by construction, so under actuation an accepted deflation
  would buy blue a thinner audit before any detector fires: the telemetry consumers are
  post-run, and the condition-5 floor is executed by the seat that just accepted):** under
  actuation, an accepted-dispute delta enters the mass computation only after a **one-round
  contest window with a named operator** (restated round 3 per red R3-3 — the round-2 window
  had no seat positioned to watch it: lenses do not read that surface, red-merge had just
  accepted, the judge never sees non-docketed disputes; a delay nobody watches is not an
  absorber. Vector honesty, mirrored from red's own log: the window phrasing was red's R2-6
  fix text — the hole was still a hole). The operator mechanism is a read-surface change:
  **pending-entry deltas are LISTED in the round's `### RED` debate entry** — a git-tracked
  surface every seat and the human operator already read — and any seat (blue, a lens, the
  lead, the operator) may docket a listed delta to the judge within the window; an unlisted
  delta never enters mass (the listing is the precondition, checkable against the telemetry
  line's delta record). **And, as a second guard — NOT equivalent (the round-2 "or,
  equivalently" equated a mechanism-bearing branch with a mechanism-free one; dropped per red
  R3-3(b)):** accepted deltas auto-docket to the judge before actuating when the round's
  **cumulative** accepted-delta magnitude crosses a stated threshold — cumulative per round,
  not per-delta (per red R3-3(c): a per-delta threshold under clause (vii)'s 5/round cap makes
  salami-slicing in-spec, compounded deflation bounded only by 5 × threshold × rounds) — with
  overflow batch-docketed as a single judge item, mirroring clause (vii)'s overflow rule.
  Stated honestly: the judge is the absorber of contested deltas only on the contested branch;
  the accepted branch's absorbers are the OPERATED window in-run and the post-hoc audit
  after — blue's deflation incentive and red's inflation incentive are symmetric, and the
  channel's design must say so.
- **(vi) Disposition at every exit (restated exit-agnostically round 2 per red R2-6(b) — the
  round-1 clause closed one loop exit of three):** the loop has three exits, each verified
  first-hand this round — the PASS break (l.236), the judged-deadlock break (l.256), and
  `while (round < maxRounds)` expiry (l.192) — and blue speaks last every round, so a
  final-round dispute structurally has no next red round at ANY exit. A PASS, deadlock, or
  maxRounds exit under the round-0/round-1 design would moot pending disputes, shipping the
  contested grade in the permanent record: the channel would fail the terminal case it is most
  needed for (lane 1's §3.4 final-round-grades observation, applied to the channel itself).
  Clause: pending or held disputes at ANY loop exit — PASS, deadlock, or maxRounds —
  auto-docket for judge disposition before assembly. **Terminal resolution set (added round 3
  per red R3-12(a)):** at a terminal exit there is no next round to carry into, so the
  resolution set for this docket class EXCLUDES `carried` — available terminal resolutions
  are accepted-with-delta, rejected-recorded-as-contested, or unresolved (recorded as such in
  the assembled report); without the exclusion a carried-at-exit dispute exits looking
  disposed while the contested grade ships as if this clause did not exist. **Priced (per red
  R3-12(b) — the terminal firing class was absent from this section's cost enumeration):**
  the exit-time dispatch is one judge firing at the terminal boundary when pending disputes
  exist — a full ~$10–13 only when the terminal docket is otherwise empty; marginal docket
  growth where a terminal contested docket already fires (§6.4 item 6's rule).
- **(vii) Dispute cap / batch rule:** `grade_disputes` is otherwise uncapped, and each
  disputed id×dimension obligates a red response on pain of auto-docket — a cost lever pointed
  at red, rational for blue exactly under the actuation regime the interlock exists for.
  Clause: a per-round cap (e.g. 5 disputes) with overflow batch-docketed as a single judge
  item rather than N obligations.

Cost: two optional fields, one filter clause, one ordering change, one log line, one cap
constant, one judge-prompt sentence, one enum value, one simulator case, plus the priced
terminal-exit firing class of clause (vi) (added round 3 per red R3-12). Complexity: low.
Doctrine: clean — routes to the existing judge, cheapens nothing. Likelihood honesty (red's
own grading, accepted): the three holes are low-likelihood this run (no actuation, zero
observed dispute traffic) rising to medium-high under the actuation the interlock is mandatory
for — which is exactly when the clauses must already be in the spec.

### 3.4 Lane 1's dissent (preserved in full) [minority: lane-1/disconfirming-first]

REJECT both halves for run 4: no exhibited need (zero instances of the residual case), the
persistence path is covered by shipped mechanism, and grades currently control nothing but the
record. Named revisit trigger (binding, not decorative): the moment any mechanism makes red's
grades load-bearing for spend or termination, the grade-dispute channel MUST ship with it.
Related structural observation carried regardless of disposition [minority:
lane-1/disconfirming-first]: **final-round grades are never adversarially audited** (the run
ends; R5-1..R5-6's grades shipped un-reviewed) — the cheap fix is one more blue response or a
stop-resume, not a standing panel.

### 3.5 Blue's synthesized position on 3a

**RATIFY the minimal envelope form (§3.3).** Honest tally, restated round 1 per red R1-16 (the
round-0 "two of three lanes ratify in some form" counted a conditional vote at face value):
**1/3 unconditional** (lane 3), **1/3 conditional on lever 2's actuation** — a premise blue
REJECTS this run, so that vote lends no support to ratifying now — and **1/3
reject-with-binding-trigger** (lane 1, §3.4). The ratification therefore stands or falls on
lane 3's grounds alone, argued explicitly: (a) the asymmetry is structural (red owns the
permanent record; blue has no machine-readable contest path); (b) the downstream consumer is
demonstrated, not hypothetical (this run's docket was assembled from run 3's graded record);
(c) the mechanism cost is near-zero and shrinks further only if never built; (d)
default-to-docket makes the field self-enforcing rather than R3-2-class decorative. Blue
weighs (a)+(b) as sufficient: record-integrity defects compound across runs, and the cheapest
time to ship a correction path is before the record it protects becomes load-bearing. Stated
honestly per all three lanes: expected round-savings zero; this is record-integrity insurance.
**BINDING interlock (all three lanes converge on this), qualified round 1 per red R1-2:** the
channel is a mandatory companion of any future lever-1/2 actuation — and the interlock claim
holds **only with §3.3 clauses (v)–(vii) included**: an interlock whose accepted branch is
dark is not a safety mechanism but a one-sided one (blue's deflation lever unaudited).
Rejecting lever 2's actuation this run is part of what makes the channel's low urgency
acceptable, and shipping the channel now is what makes a future actuation debate honest.

### 3.6 Best-of-N grading — REJECT (3/3)

- **The backlog's own precondition is unmet:** it defers best-of-N until "runs 4–5 show
  lone-voice bias survives."[^BacklogLevers] All three lanes searched the pinned corpus for a
  surviving-bias instance: none exists — every identified grade error was caught and corrected
  within the adversarial loop (blue caught red's transposed count R2-1; red caught its own stale
  framing R3-7; red's merge overruled a lens's wrong hold at R5-1 and logged the error against
  itself).[^Retro3Docket][^R5OneOverrule]
- **The "lone-voice" premise is itself partially false structurally** [minority:
  lane-3/local-repo]: run-3 grading already passed through at least two voices per gap (lens
  grade → merge temper), and the merge exercised that power against its own lenses four times in
  round 5 alone — R5-5 tempered HIGH→MEDIUM-HIGH, R5-6 MEDIUM-HIGH→MEDIUM, both downward, plus
  two outright lens-error overrules.[^FindingsBoard][^CitationLedgerRun3] Multi-voice grade
  correction is observed working; a graded panel would add a judgment-seat cost multiple
  ($7.52–$13.56/round is what the one existing judgment merge costs — full verified range per
  red R1-25; the round-0 band quoted rounds 2–5 only, excluding round 1's $7.52 without stated
  reason)[^CostAudit] against a defect class with zero occurrences.
- **The panel literature's benefit concentrates where FEOV isn't** (lanes 1 and 2): PoLL's
  panel-beats-judge result (~7× cheaper) rests on panels of smaller models from **disjoint model
  families**, and names intra-model bias as the thing panels dilute;[^PoLL] the direct
  disconfirmation for same-family panels: nine correlated judges delivered "about 2 independent
  votes' worth of information," panels underperformed the independent-voting ideal by 8–22
  points, the single best judge matched or exceeded the full panel, and aggregation tricks
  recover ≤11% of the gap.[^NineJudges] FEOV grading runs single-provider by construction — the
  correlated-errors case, not the PoLL case. FEOV's grades are also not one-shot scalar
  judgments but evidence-anchored, multi-round, adversarially contested records — the panel
  literature's single-shot setting is the weaker analogue [minority:
  lane-1/disconfirming-first for this framing].
- **Adversarial-first has primary support** [minority: lane-2/primary-literature]: debate raises
  judge accuracy over consultancy and no-assistance baselines (76% vs 48% for LLM
  judges),[^PersuasiveDebate] with the honest caveat that gains are task-dependent and shrink in
  some regimes.[^WeakJudges]

**Revisit trigger:** the backlog's condition as written (per-gap records from runs 4–5 showing
bias that survived the loop — which the `grade_disputes` records of §3.3 exist to collect),
**plus** one precondition added from the literature [minority: lane-2/primary-literature]: a
genuinely independent second grader (different model family) must be available, or the panel is
decorative. Open question carried: if a cross-family grader is structurally unreachable from
this harness, the backlog's revisit clause should say so instead of implying a panel is one
decision away.

**Confidence: HIGH** on the best-of-N rejection (convergent primary literature + zero corpus
instances + observed multi-voice correction); **MEDIUM-HIGH** on the channel ratification (the
zero-instances fact cuts both ways: absence of need vs. cheapness of insurance; the dissent is
preserved and the interlock is binding either way).

---

## §4 Sharded findings + collator — RATIFY sharding with conditions; REJECT the collator seat

### 4.1 Convergent: the doctrinal question is answerable from the text — no conflict exists

The full-re-read MUST, quoted from the pinned agent contract: "BEFORE auditing, YOU MUST re-read
the FULL living report in context — a change-summary is a navigation hint, never the audit
surface."[^RedAuditorMust] The object is **blue's living report** — the lens prompt makes it
explicit ("the FULL living report ${runDir}/blue/report.md").[^EngineSource] No mandate anywhere
in the agent file or the dispatch prompts requires red to re-read its own closed cases (verified
by direct read of the agent file and both red prompts at `5396952` — all three lanes). Sharding
red's own findings file (open-items ledger + closed archive; merge reads open + this round;
archive readable on demand) narrows no audit read the doctrine names.

**Correction to this run's own frontier** [minority: lane-3/local-repo]: run-3 friction #15 (the
25k Read cap forcing three windowed reads) is about **`blue/report.md` at the merge seat**, not
findings.md — sharding does not fix that entry, and the frontier's attribution must not survive
into later rounds. (Both files did outgrow the Read cap — findings.md ended at 106,772 bytes,
blue/report.md at 159,394[^FrictionFifteen] — but the friction entry's complaint names the
report.)

### 4.2 Convergent: the cost case, measured

- Merge seats are the priciest recurring line: red-merge $7.52/$13.22/$12.64/$10.60/$13.56
  across rounds 1–5 = $57.54 of the run's $149.95 (38%), rate-driven at the judgment tier (5×
  cache-read, 5× cache-write — corrected round 1 per red R1-30: the session tier's cache-write
  rate is 12.5 $/MTok *absolute* against sonnet's 2.5, so the multiplier is 5×, same as
  cache-read; cost.md's own finding 3 carries the 12.5× multiplier error internally — logged
  as cost.md defect (b) in §6.4).[^CostAudit]
- The measured driver is TURNS × CONTEXT: an agent re-reads its whole accumulated context every
  tool call; red-merge-r1 alone held ~100–150K of material across 2.7M+ cache
  reads.[^Backlog28d] The Anthropic prompt-caching documentation confirms the mechanism at the
  specification level [minority: lane-2/primary-literature]: the entire prefix is re-billed at
  cache-read rate (0.1× base input) every API turn; cost.md's rate assumptions
  match.[^PromptCaching]
- **cost.md finding 2 is internally contradicted — a live defect in a pinned artifact**
  [minority: lane-3/local-repo]: "merge cost tracks DISPUTE size (peaked r2, fell after)" is
  contradicted by its own table — round 5's merge was the run's most expensive ($13.56 > r2's
  $13.22; 7.87M cache reads > r2's 5.64M; 61 turns) on the run's *second-smallest* dispute
  board (6 open gaps, vs round 4's 5 — superlative corrected round 1 per red R1-32; the
  contradiction argument is unchanged, since the actual smallest board had a $10.60 merge),
  while findings.md reached 1364 lines.[^CostAudit][^FindingsBoard] The growth term
  is the cumulative archive plus the growing transcript dragged through every merge turn at
  judgment-tier rates — which flips the finding's lesson into H4's whole case. Needs a
  correction in the pinned artifact's successor.
- **The quality argument runs the same direction as the cost argument** (lanes 1 and 2,
  disjoint sources): long-context performance degrades substantially with input length even at
  perfect retrieval (13.9%–85% across models/tasks),[^ContextLength] and models use mid-context
  material far worse than edges ("lost in the middle").[^LostMiddle] A merge seat holding 100K+
  of mostly-own-closed-history is operating in the measured degradation regime; shrinking the
  resident archive is judgment-STRENGTHENING, not only cheaper — the rare lever the doctrine
  favors twice over.
- **Measured sizing (corrected AND measured round 2 per red R2-3 — the round-1 text's
  "gitignored and absent at the pin" was a false unavailability premise):** gitignored ≠
  absent. `agent-transcripts.tar.gz` is untracked in the git object store at the pin but
  PRESENT in the pinned run's working tree (7,040,514 bytes, 46 per-agent transcripts —
  lane 3 checked git alone and never ran `ls`; availability is not git-trackedness). Blue
  extracted it and ran the decomposition this round: tool-result bytes ingested per red-merge
  transcript, classified by source file, with a cache-weighting (bytes × remaining turns)
  approximating prefix re-billing:[^MergeDecomposition]

  | red-merge round | tool-result ingest | blue/report.md | red/findings.md | lens candidates | debate.md |
  |---|---|---|---|---|---|
  | 1 | 174KB | 36% | 0.1% | 46% | 4% |
  | 2 | 247KB | 36% | 9% | 28% | 7% |
  | 3 | 250KB | 43% | 17% | 21% | 10% |
  | 4 | 190KB | 20% | 32% | 33% | 6% |
  | 5 | 318KB | 46% | 29% | 19% | 2% |

  The findings-file share of merge ingest grows 0.1% (r1) → 32% (r4) / 29% (r5)
  (~60–91KB/round by rounds 4–5 ≈ 15–23K tokens at ~4 bytes/token).

  **Instrument and derivation status (corrected round 3 per red R3-1 — the round-2 dollar
  series and its "measured" label do not survive reproduction from the committed
  instrument):** the round-2 parser ran from a session scratchpad and was not retained —
  under this report's own §6.2 attestation ceiling, that made the dollar attribution an
  unreproducible work-done claim whose audit artifacts were exactly the ones not git-tracked.
  The instrument is now **committed as `trajectories/decompose-merge.mjs`** in this run's
  directory (tarball-retention assumption stated in the footnote) and was re-run at round 3
  against the pinned tarball. Three results, separated by derivation status:

  - **The raw decomposition table above REPRODUCES exactly** — all five rounds within ±1
    point, round-5 blue/report.md ingest 145.7KB, per-round totals 173/246/249/188/317KB.
    Status: measured, reproduced from a committed instrument. HIGH.
  - **The round-2 cache-weighted dollar series does NOT reproduce under the method's own
    stated 4-bytes/token conversion.** Strict application prices findings.md at ≈$0.26 (r2) /
    $0.53 (r3) / $0.89 (r4) / $1.16 (r5) — Σ≈**$2.8/run** findings-attributable, not the
    round-2 ≈$12. The printed ~$1.40/2.60/4.10/4.10 series — and lens 3's independent
    recompute (Σ$12.4) — reproduces only if cache-weighted BYTES are priced as tokens: the
    byte→token conversion dropped at the pricing step. Two implementations agreeing was
    consistency of convention, not correctness. Cross-check against the measured ceiling:
    round 5's entire cache-read spend was $7.87 (7.87M cache-read tokens, cost.md); a $4.10
    findings share would be 52% of it while findings was 28.8% of file ingest — impossible
    once the system prompt and harness overhead (in no file bucket) are added; the corrected
    $1.16 (≈15%) fits. A proportional-share upper bound (findings' ingest share × each
    round's measured cache-read dollars) gives ≈$6/run. Honest findings-attributable band:
    **≈$3–6/run at run-3 scale** (direct-attribution floor $2.8; proportional-share ceiling
    ≈$6, which charges findings for its share of overhead re-billing
    too).[^MergeDecomposition][^CostAudit]
  - **The archive fraction is now DERIVED, not asserted** (the round-2 "clear majority" had
    no documented derivation — red R3-1(a)): at the pinned round-5 state, findings.md's
    "superseded … preserved" blocks (l.340 through EOF l.1364) hold 76,356 of 105,223
    LF-normalized body bytes = **72.6%** (one awk line, quoted in the footnote); the split is
    conservative — closure-status records above the l.340 boundary would also move under
    sharding — and earlier-round file states carry progressively smaller archive shares,
    since the archive is what accumulates. Sharding-addressable ≈ archive share ×
    findings-attributable ≈ **$2–4/run at run-3 scale** — replacing the round-2 "$7–10,
    squarely in lane 1's $5–15 band," which inherited the 4× pricing overstatement. The
    corrected figure falls BELOW lane 1's dollar band.[^MergeDecomposition]

  Comparison to lane 1's token estimate, restated like-with-like (round 3 per red R3-5 — the
  round-2 clause compared the WHOLE findings file to lane 1's estimate of the ARCHIVE SHARE,
  a metric conflation whose honest reading inverts the claim): lane-1.md l.281 estimated the
  archive's share of merge context at 20–30K tokens by round 5; the measured round-5
  whole-file ingest is ~23K tokens, so lane 1's band is bounded above by the whole-file
  figure — its upper half excluded — and the measured archive sub-fraction (72.6% × 23K ≈
  16–17K) sits below the band's floor. On both the token and dollar axes, lane 1's
  directional estimates ran HIGH; §4.2 validates sharding's direction, not lane 1's
  magnitudes.

  Two bonus measured facts, restated from the table as printed (round 3 per red R3-4 — the
  round-2 universal was refuted by its own round-4 row under both raw and cache-weighted
  measures, the its-own-table contradiction class this report flags in cost.md finding 2):
  **blue/report.md is the largest merge-context component in rounds 2, 3, and 5** (145KB
  ingested at round 5 — run-3 friction #15's real referent, untouched by lever 4a); **in
  round 4 the lens candidates (33%) and findings.md (32%) both exceeded it (20%)**. And
  debate.md is minor at the merge (2–10%) though the judge reads it in full (§6.1).
  **Residual disclosed (round 3 per red R3-13):** the four named columns sum to 80–96% per
  round (86/80/91/91/96); the ~4–20% remainder is the "other" bucket — citation-ledger.md
  reads, red-auditor memory files, git/wc command output, and scratchpad staging, composition
  varying by round (the committed parser prints it: 13.6/19.5/9.0/8.6/4.0%) — so the table is
  a four-file decomposition plus a disclosed residual, not a complete partition. Caveats,
  stated: single run; bytes→tokens ≈ 4:1 assumed; the weighting ignores the system prompt and
  harness overhead — measured-directional, not exact; run 4 re-measures from its own
  transcripts before the PR (`node trajectories/decompose-merge.mjs <extracted-dir>` — one
  command now, not a bespoke parse).

### 4.3 Convergent: the disconfirming tests resolve in sharding's favor

**(a) R5-1 (the red-reads-own-closed-cases catch) survives sharding — with a condition.** The
catch's *trigger* (a discarded lineage enumeration in blue's §3 row 23, written in round 4) sits
in blue's report — the always-read audit surface, untouched by sharding; the *verification*
(every chain link checked against red's own closure entries, some closed 2–3 rounds prior) is a
targeted on-demand read into the archive — exactly the protocol's mode-3 leaf-node
fetch.[^R5OneDetail] Lane 2's sharpening: the closure records consulted WERE closed-archive
material at that moment, so the on-demand read must be a demanded read, not a discretionary one —
an explicit prompt line ("a claim about a closed gap is verified against the archive, never from
memory"). Without that line, sharding is blind to exactly the stale-lineage class run 3
produced; with it, the catch replays identically. Ratification is conditional on the line
(condition 2 below). Contrast noted with lever 5's hard case [minority:
lane-1/disconfirming-first]: R5-1's trigger is *visible* in the always-read surface; R4-4's
trigger (a stale numeral in an unchanged paragraph) is invisible — that asymmetry is why blue
ratifies lever 4 while holding lever 5.

**(b) The dedupe function is the real archive dependency — a condition, not a rejection**
[minority: lane-3/local-repo]: the merge assigns "fresh R{round}-N ids to genuinely new gaps
only" and runs merge-time dedupe notes every round — knowing a candidate gap is *new* requires
comparison against ALL prior gaps, including closed ones.[^EngineSource][^FindingsBoard] Under
naive sharding the merge either re-reads the archive (savings vanish) or mints duplicate ids for
re-litigated closed ground. Hence the compact closure index (condition 3). **Sharpened round 1
per red R1-8 — the dichotomy must not be resolved by assertion:** whether a one-line index
suffices as a dedupe key IS the judgment-heavy comparison the dichotomy worries about; run-3
dedupe compared full prose, and semantic near-misses (same defect, different framing) are the
observed norm in this corpus. If the key is insufficient, the design silently falls into the
duplicate-id horn — the identity-keyed-detector failure class this repo already documented
(R4-1's whole lesson). Hence the near-match trigger now written into condition 3: an index
near-match to a candidate gap forces a targeted archive read BEFORE a fresh id is minted — a
mode-3 demanded read, mirroring condition 2 — so the index is a *screen*, never the final
comparator.

**(c) The write path is safe if the shards are skeleton-born and neutrally named** (all three
lanes): run 3's accidental controlled experiment isolated the write-block guard as
filename-keyed and path-independent (`findings.md` refused even in a scratchpad; a neutral
filename succeeded);[^FrictionFour] Edit on pre-created files worked every round, and the
citation ledger was appended via `cat` across four rounds with zero
incidents.[^FrictionTen] The frontier's undercounted-cost-cell worry does not materialize under
naming discipline; lane 3 adds that a small open ledger is possibly one detour *cheaper* to Edit
than the ~107KB findings.md monolith measured at the pin (figure corrected round 1 per red
R1-34: the round-0 "54KB" matches neither pinned file — findings.md is 106,772 bytes,
blue/report.md 159,394; the 54KB figure is a stale number carried by run-3 friction #15 and
backlog 31(g), logged as a pinned-artifact defect in §6.4 item 5). (The still-open backlog item
"sanctioned write path for red's living
artifacts" remains the durable fix; the skeleton is the shipping mitigation.) Live confirmation
from this very round: blue's own synthesis Write of `blue/report.md` was refused by a
filename-keyed guard at the synthesis seat and required the neutral-name + copy detour — the
write-block class is still active and still filename-keyed (logged to friction.md).

**(d) The citation-ledger precedent holds, with one inherited qualifier** (all three lanes cite
the precedent; the qualifier is convergent lanes 2 and 3): the skip-rule held all prior-round
confidences through run 3 with zero closed-citation regressions,[^LedgerPrecedent] and R5-2 was
NOT a ledger regression — the stale MA-status claim was never a ledgered pair (verified against
citation-ledger.md; first MA-status entries appear in round-5 blocks) [minority:
lane-3/local-repo for the verification].[^CitationLedgerRun3] But the pattern was not free: its
prose-only skip-trigger suppressed source-drift re-checks and had to be repaired by adding
drift/time triggers (run-3 row 10 / R2-9; the repaired clause is in the shipped
`ledgerClause`).[^Retro3Docket][^EngineSource] The findings shard needs the analogous reopen
triggers designed in from day one (condition 4).

### 4.4 Convergent: what sharding gets — one authoritative site

Single source of truth for status [minority: lane-1/disconfirming-first for this framing]:
closure status lives ONLY in the open ledger (items move ledger → archive on closure; archive
blocks are immutable). R5-1's failure class — status lines contradicting each other across
sections of one growing file — gets one authoritative site instead of five preserved-verbatim
historical blocks.[^R5OneDetail]

### 4.5 Ratification conditions (union of all lanes' conditions)

1. **Single source of truth for status** — ledger is authoritative; archive immutable (lane 1).
2. **A named MUST for archive verification** — any lineage/closure claim (in blue's text or
   red's own docket) is verified against the archive by targeted read; grep-cheap, mode-3
   discipline, written into the red-auditor contract so R5-1-class catches are demanded reads
   (lanes 1 and 2). **Extended to the judge (round 2 per red R2-5):** the judge is the SECOND
   full-read consumer of findings.md (each dispatch reads `debate.md` + `red/findings.md` in
   full, prompt l.249), and its docket is the lineage-dense subset by construction (contested
   = re-raised ids + supersedes-descendants, ll.244–245) — under sharding the closed ancestors
   it rules against move to the archive. Clause: the demanded-read MUST is written into the
   lead-judge prompt/agent as well — a ruling on a supersedes-descended gap MUST read the
   named ancestors' archive records — because the ruling class most sensitive to missing
   ancestor context is `carried` vs `risk_accepted`, the §6.4-item-6 gate-erosion path this
   report itself graded MEDIUM.
3. **Compact closure index resident in the ledger** — one line per closed gap: id | closure
   class | one-line summary | supersedes; full prose in the archive; dedupe stays resident at
   ~10% of archive bytes (lane 3). **Near-match trigger (added round 1 per red R1-8):** an
   index near-match to a candidate gap forces a targeted archive read BEFORE a fresh id is
   minted (mode-3 demanded read, mirroring condition 2) — the index screens, it never decides.
4. **Reopen + drift triggers** — a closed case reopens into the ledger when (i) a new gap's
   `supersedes` names it, (ii) blue's report cites it, or (iii) a spot-check samples it; archived
   closures whose evidence cites volatile living sources inherit the ledger's drift/time
   re-check triggers (lanes 2 and 3).
5. **Archive spot-check floor** — N sampled closed cases re-verified per round, never zero
   (replacing round-4's full 41-closure sweep, which caught nothing that round but is the class
   of check the floor must keep alive) (lanes 2 and 3).
6. **Skeleton-born, neutrally named shards, git-tracked** — both files pre-created in the
   blackboard skeleton (PR #14 pattern) with non-report-semantic names (e.g. `red/ledger.md`,
   `red/archive.md`); the archive stays readable on demand — this is a read-default change, not
   a retention change; blue-additive doctrine intact (all lanes). **Two additions round 1 per
   red R1-14:** (i) a **name preflight** — the write-guard's key set is unenumerated and
   demonstrably wider than `findings.md` (it refused `blue/report.md` at this run's own
   round-0 synthesis seat), so the proposed shard names MUST be test-written before the first
   sharded run, not discovered mid-merge — **and the preflight MUST issue real Write calls
   from a live seat in the production harness, of the same seat class as the future shard
   writer** (pinned round 3 per red R3-11 — the round-2 clause left the executor's seat class
   open, and its own example, "the next run's skeleton-creation step writes both shard
   files," is plausibly a lead/orchestrator step: every corpus-observed write-block firing is
   at a SUBAGENT seat, nothing in the corpus has read the guard's seat-(in)dependence
   first-hand, and if the guard is seat-class-scoped a lead-seat preflight passes while the
   first red-merge shard write still blocks mid-merge — the vacuous-preflight failure one
   level up. Concretely: the first sharded run's red-merge writes both skeleton shards as its
   opening act; alternatively, verify the guard's seat-independence first-hand and cite the
   verification in the PR). Hardened round 2 per red R2-17, simulator internals verified at
   the judge seat and re-read first-hand at this one: the PR-#14 simulator is "a stub world"
   running "canned envelopes and no live agents" (tests/simulator/harness.mjs) — the
   write-block guard lives in the live-seat harness/permission layer, so a simulator-only
   preflight passes vacuously and the block is still discovered mid-merge, behind a green
   checkmark; (ii) the **engine prompt edits are
   part of the change**: `debate.js` hardcodes `red/findings.md` at both the red-merge prompt
   (l.216) and the judge prompt (l.249) — renaming without both edits strands the judge's full
   read; verified first-hand at the pin-equal working tree.
7. **Observability for the prompt-MUST conditions (added round 1 per red R1-6; executor and
   guarantee restated round 2 per red R2-4 — the round-1 clause committed the
   policy-without-mechanism class it was written to kill):** conditions 2, 4, and 5 as stated
   are prompt-level MUSTs with no observable — the exact R5-5/R3-2 unenforced-field class
   §3.3's default-to-docket exists to kill, and this report cannot coherently demand schema
   enforcement for lever 3a while leaving 4a's load-bearing conditions hortatory (red's
   symmetry argument, accepted). Clause, with each check's executor named and its guarantee
   stated to exactly what it delivers (§6.2 attestation ceiling): `RED_ENVELOPE` gains an
   `archive_spot_checks` field required non-empty from round 2 — the script IS the executor of
   the shape check, so an UNSET or empty field fails structurally; a VACUOUS one (plausible
   ids with no work behind them) is beyond any in-script check, because unlike the lineage
   throw — which cross-references two independent envelope structures (`closures` vs
   `gaps[].supersedes`) — `archive_spot_checks` has no independent counterpart in the script's
   sight. Vacuity's named auditor is post-hoc: the run retrospective / next-run red docket
   over the git-tracked shards (the demonstrated consumer class). The closure-index line count
   and archive block count both ride the envelope as merge-reported integers; the script's
   arithmetic comparison of them catches only a self-inconsistent self-report — the true
   counts are verified by the same named post-hoc auditor, or, if built, by a hook of the
   `sc-recall-index` class, which has filesystem access and already fires on markdown writes
   (the one in-run executor with eyes on the files; optional, not assumed). **Condition-2
   demanded reads are logged in TWO named homes, one per writer (restated round 3 per red
   R3-2 — "logged in the same field" was mechanically impossible for the judge's half:
   `JUDGE_ENVELOPE` is required `{deadlock, resolutions[{gap_id, resolution, rationale}]}`
   with no log field, and the judge dispatch fires after `RED_ENVELOPE` is already
   submitted — the third consecutive round a condition-7-class observable rode an unavailable
   mechanism, R1-6 → R2-4 → R3-2):** red-merge's demanded reads ride `archive_spot_checks` as
   before; the judge's demanded reads live in the judge's own channel — **each chain ruling's
   rationale MUST name the ancestor archive records read**, git-tracked in the round's
   `### LEAD` debate entry (zero schema change; this run's round-3 LEAD entry demonstrates
   the form by construction). Writer-capability audit of this condition, run end-to-end after
   the edit per the lead's R3-2 instruction — every named observable and its writer:
   `archive_spot_checks` → red-merge, author of the envelope that carries it (can write it);
   closure-index line count and archive block count → red-merge, same envelope (can write
   it); judge demanded-read log → the judge, in its own `rationale`/`### LEAD` prose (can
   write it); vacuity audit → post-hoc retrospective/next-run docket over git-tracked shards
   (the demonstrated consumer class, post-hoc by design). No observable in this condition now
   names a writer that cannot physically write it. In-corpus base
   rate justifying the cost: a schema'd-but-unset field ran three rounds unnoticed in run 3.

### 4.6 The collator — REJECT the seat (3/3); RATIFY the degenerate form

Independent kills, any one sufficient:

1. **Dominated alternative at zero mechanism (all three lanes):** the script cannot concatenate
   (no filesystem access by design),[^EngineSource] but the merge agent can: one added prompt
   sentence — first action, `cat <runDir>/red/candidates/round-N-lens-*.md >
   <scratchpad>/round-N-all.md`, then read the single file — captures the entire
   one-big-read/fewer-turns benefit with no new seat, no dispatch cost, no handoff. (Output
   path corrected round 1 per red R1-13: the round-0 sentence's cwd-relative redirect target
   is a footgun twice over — agent cwd resets between bash calls, and a concatenation landing
   in `red/candidates/` becomes a sixth file for every downstream glob consumer: the merge's
   own read instruction, future `found_by` audits, mechanical convergence counts where a
   singleton would read as 2-of-6 and convergence counts feed corroboration grades, and the
   concatenation being swept into qmd's index at the next `qmd update` — triggered by any
   seat's subsequent markdown write; mechanism corrected round 2 per red R2-16 and verified
   first-hand at this seat: the `sc-recall-index` hook matches `Write|Edit` (hooks.json), so a
   Bash `cat >` redirect never fires it at write time — the hazard is the later sweep, one
   step downstream and still real. The batching sentence MUST
   name an absolute output path outside `red/candidates/` and outside the recall index's
   markdown watch surface — the seat's session scratchpad is the natural home.) This is the
   backlog's own lever 3, "prompt-level read batching,"[^Backlog28d] and it strictly dominates
   a seat whose dispatch overhead is the same order as its plausible saving.
2. **The arithmetic does not support a seat** (lanes 2 and 3): collapsing 6–8 read turns
   (6 lens files minimum at live-code claim counts — swept round 2 per red R2-2; measured
   lens-candidate ingest at the merge: 52–80KB/round, §4.2) saves
   roughly $1–2/round at round-5 merge rates, against which the collator costs its own agent
   run — a bulk dispatch that itself reads all five files, roughly the spend it saves, to
   produce a lossier input — plus a null-return failure surface (the class that crashed run 2)
   [minority: lane-2/primary-literature for the null-return point] and another write-blocked
   filename to manage.[^CostAudit]
3. **Normalization is the documented in-corpus failure class** [minority:
   lane-1/disconfirming-first for the in-corpus instance]: run 3's envelope enum rounded red's
   compound grades every round — "every compound grade above was rounded; the authoritative
   grading lives in red/findings.md" (friction #6) — a live instance of normalization destroying
   grading nuance, fixed in PR #15 by widening the enum.[^FrictionSix] The
   hierarchical-summarization and multi-agent-handoff literatures say the same thing generally:
   intermediate compression drops edge-case detail and introduces errors that compound
   downstream; the standard mitigation is "keep the source text alongside" — which degenerates
   the collator to concatenation.[^HandoffLoss][^HierSumm]
4. **Clustering is judgment, not mechanics** (all three lanes): deciding two lens findings are
   "the same gap" sets the convergence count, and convergence counts fed corroboration grades
   ("four of five lenses converged" is part of R4-1's HIGH; "three lenses converged" part of
   R5-1's); run 5's two lens errors were caught precisely because the merge held conflicting
   lens outputs side by side and re-derived first-hand.[^R4OneDetail][^R5OneDetail][^CitationLedgerRun3]
   A bulk-tier collator pre-deciding convergence (or pre-resolving lens conflicts) cheapens the
   adversary's evidentiary input — the doctrine's named prohibition. PR #15's lens-scoped labels
   already removed the renumbering mechanics that was the collator's honest
   half.[^AlreadyShipped]
5. **The inspection literature's directly-relevant result** [minority:
   lane-2/primary-literature; attribution split round 1 per red R1-29]: Votta (1993) found
   collection meetings added few defects over independent reviews; the *Empirical Software
   Engineering* replication — not Votta's own paper — found meetings significantly better at
   false-positive reduction. The combined lesson stands: the consolidation step's value is
   judgment (killing false positives), not collation mechanics. (Verification honesty: the ACM
   primary returned 403 to red and to blue; the split rests on search-level corroboration of
   the replication record — grade MEDIUM.)[^Votta]

**Confidence: HIGH** on the doctrine-scope argument, the write path, the precedent, the R5-1
replay, and the collator rejection (dominance + in-corpus demonstration); **HIGH** on the raw
decomposition table (reproduced at round 3 from the committed instrument, all five rounds);
**MEDIUM** on the dollar magnitude of sharding savings (restated round 3 per red R3-1 — the
round-2 "now MEASURED" claim overstated its own derivation status: the raw ingest shares are
measured; the dollar attribution is MODELED — bytes→tokens 4:1, remaining-turns weighting —
and was corrected ≈4× downward at round 3 when the committed parser applied the model's own
stated conversion, §4.2; the archive fraction is derived from one pinned file-state; run 4
re-measures from its own transcripts before the PR).

---

## §5 Round-scoped audit — HOLD through run 4; conditionally RATIFY for run 5, gated

### 5.1 Convergent: the blindness is structural, and the type specimens are R4-4 and R5-2

An unpropagated correction site sits, by definition, in a section blue did not change; a
changed-sections audit excludes it. Blue's dominant run-3 failure mode was exactly this class —
5 incomplete-propagation chains in 5 rounds (R1-5→R2-4→R3-4/R3-9; R2-5→R3-10; R2-7→R3-6;
R2-8→R3-5→R4-3; plus R4-4's unpropagated fifth numeral), each costing a full audit round at
$25–30.[^PropagationChains][^CostAudit]

Clean specimens from the record [minority: lane-3/local-repo for the specimen selection]:
**R4-4** — the fifth "4th occurrence" numeral sat in §3's risk-accepted paragraph, untouched
since round 2's correction missed it, caught in round 4 by a report-wide grep at the merge seat
("grep '4th|fourth' at merge: exactly one uncorrected instance");[^R4FourGrep] **R5-2** — §4 row
1's status, unchanged since round 2, went stale by cross-corpus drift and was caught only by a
lens re-reading the other corpus first-hand.[^FindingsBoard] Both catches came from audit
surface a changed-sections rule excludes.

**Correction to this run's own frontier** [minority: lane-3/local-repo]: the frontier's R4-3
type specimen is *weaker* than claimed — the unedited sentence sat in the same row 6 that
R3-5's fix edited, so row-granularity scoping would have included it. The structural-blindness
argument survives on R4-4/R5-2; it must not ride on R4-3. (Lane 1's independent mapping agrees:
R4-3 is covered by a section-granular contested-lineage arm.)

### 5.2 The catch-to-arm mapping (lane 1's audit of the full catch record) [minority: lane-1/disconfirming-first]

Every late-round catch red actually made by full re-read in run 3, traced to whether a four-arm
scope would have surfaced it:

| Catch | What found it (per the pinned record) | Scope arm that covers it |
|---|---|---|
| R4-3 (unedited ambiguous sentence) | Sat in the SAME CELL as R3-5's contested fix; lenses 2+4 read the cell | Contested-lineage arm, section-granular |
| R4-4 (fifth stale numeral, paragraph unchanged since round 1) | Report-wide grep at merge | Propagation-grep arm — **the catch was ALREADY made by grep, not by linear re-read** |
| R3-6 (zero "independent" hits) | Repo-wide grep | Propagation-grep arm |
| R3-10 (untagged cost figures) | Direct read of both instances | Propagation-grep arm (grep the corrected figure) |
| R5-1 (discarded enumeration in row 23) | Row 23 was WRITTEN in round 4 — a changed section at the round-5 audit | Changed-sections arm |
| R5-2 (stale MA-status, §4 row 1 unchanged since round 2, gone stale by cross-corpus drift) | A lens re-reading the OTHER corpus first-hand | **None — no arm covers the cross-corpus-drift class** (not changed-section, not contested-lineage, not propagation-greppable); covered only by PR #16 pinning, which has zero live evidence (§5.4). Row added round 1 per red R1-3 — the round-0 table silently omitted one of §5.1's own two named clean specimens |
| R1-1 (stale §0 headline) | Round-1 full audit | Round 1 is always unscoped |

The load-bearing observation (convergent, lanes 1 and 3): **the poster-child full-re-read catch
(R4-4) was not made by reading — it was made by grep.** The propagation-aware expansion is not a
hypothetical mitigation; it is a description of how the catch actually happened at the merge
seat. **Completeness correction (round 1, red R1-3):** with R5-2 in the table, the four-arm
rule is demonstrably NOT complete against run 3's catch record on its own — it is safe only
JOINTLY with PR #16 pinning; the §5.5 gate now conditions on both load-bearing mitigations,
not one.

### 5.3 External evidence (each lane's, preserved)

- **Safe-selection theory is exact and cuts against naive scoping** [minority:
  lane-2/primary-literature]: regression test selection is "safe" only when it excludes no
  fault-revealing test, and every safety proof rests on a complete change-impact
  analysis.[^SafeRTS][^YooHarman] The prose analog of impact analysis is "every site stating a
  corrected claim" — a changed-sections scope is an UNSAFE selection by construction against
  run 3's measured dominant failure class.
- **Diff-vs-context review evidence** [minority: lane-1/disconfirming-first]: diff-only review
  reliably catches local issues and misses global-invariant violations; the field's answer is
  hybrid scoping (diff + selective context), not full re-reads forever.[^DiffReview]
- **Spot-checks must sample unchanged AND previously-clean sections** [minority:
  lane-2/primary-literature]: Fenton & Ohlsson's counter-intuition warns specifically against
  concentrating inspection where defects were previously found.[^FentonOhlsson]
- **Feasibility has materially improved since the hold** [minority: lane-2/primary-literature]:
  the propagation expansion was grep-cheap but paraphrase-blind when row 18 was held; PR #18's
  recall layer (lex + vec + hyde retrieval, hook-refreshed on every markdown write) is a
  paraphrase-tolerant site-finder — the audit-surface expansion can now be computed
  semantically, not just lexically, converting the ratification condition from "expensive
  judgment" to "cheap mechanics plus a floor."[^AlreadyShipped]
- **Context degradation cuts FOR scoping too — but the doctrine clause outranks** [minority:
  lane-1/disconfirming-first]: a lens re-reading a 159KB report every round is also in the
  measured degradation regime,[^ContextLength][^FrictionFifteen] but the protocol ranks
  full-read-of-the-audit-surface above token savings explicitly ("this clause outranks any token
  saving") — overriding a named doctrine clause needs run-4 evidence, not an inference from
  run 3. Unlike lever 4, this lever narrows the audit surface itself.[^RedAuditorMust]

### 5.4 Convergent: three shipped mitigations cover the observed miss classes — none has live evidence

1. **Blue propagation clause** (PR #15, in `blue-researcher.md` and the blue-respond prompt) —
   prevention at the source for the R4-4 class. Live rounds of evidence: zero; run 4 is the
   first trial.[^BlueMandate][^EngineSource][^AlreadyShipped]
2. **Corpus pinning** (PR #16) — kills the R5-2 cross-corpus-drift class structurally; this
   run's own PINNED.md is the first live trial.[^AlreadyShipped]
3. **Ledger drift triggers** (PR #15).[^EngineSource]

Narrowing red's full re-read before mitigation 1 has a single live run of evidence removes the
backstop and the belt in the same release — the exact compound-change class this project's own
review doctrine flags [minority: lane-3/local-repo for this framing].

### 5.5 Disposition (identical gate logic reached by all three lanes)

**HOLD for run 4** (already staged that way; this run audits at full re-read). **For run 5:
RATIFY conditionally** — round 1 always full; rounds 2+ audit scope = (a) changed sections in
full context ∪ (b) contested/lineage locations, section-granular ∪ (c) **propagation
expansion** — for every correction accepted this round, grep (and, per lane 2, semantically
retrieve) the corrected strings/figures report-wide and add every hit site to the surface ∪
(d) a nonzero random spot-check floor that includes unchanged and previously-clean sections.
Lens grades from scoped rounds carry a stated confidence discount, mirroring the row-16b
documented-tradeoff pattern [minority: lane-3/local-repo].[^Row16b]

**The gate (extended round 1 per red R1-3, R1-7, R1-11 — three conditions, all mechanical):**

1. **Propagation-clause record, with positive evidence the detector RAN:** run 4's record must
   show zero unpropagated-site regressions AND the propagation greps exercised and logged per
   accepted correction — "zero regressions observed" and "zero regressions" differ by detector
   sensitivity, which is unmeasured; a run-4 red that never runs the propagation greps produces
   a clean record that would falsely ratify scoping (blue's own §5.3 safe-selection framing:
   unsafe selection on a false-clean record). Absence of findings alone does not satisfy the
   gate.
2. **Pinning's run-4 record:** the four-arm rule is safe only jointly with PR #16 pinning
   (§5.2's R5-2 row) — any cross-corpus-drift regression in run 4 takes the same
   reject-outright branch. The gate tests both load-bearing mitigations or it tests neither.
3. **Doctrine-amendment prerequisite:** a run-5 ratification cuts red's read of blue's report —
   none of §6.3's three safe categories — and therefore requires amending the red-auditor
   full-re-read MUST and the research-protocol mode-2 clause ("this clause outranks any token
   saving") as part of the run-5 PR. The doctrine text must change with the behavior, or the
   first scoped run runs in violation of its own standing doctrine; this prerequisite belongs
   in the conditions, the cost accounting, and §6.3's honesty scope, and as of round 1 it is
   in all three.

Any condition failing = reject round-scoping outright for run 5 and re-dock the lever — it
would remove the only check that catches the engine's measured dominant regression class while
that class is demonstrably still live. The decision point and its evidence are one run away;
nothing is gained by deciding early on thinner evidence.

**Honest residual (convergent, lanes 1 and 3):** the catch-to-arm mapping is post-hoc — the
four arms were drawn looking at run 3's catch classes; a novel regression class may evade all
four. That residual is what the spot-check floor exists for, and the doctrine pins it above
zero — but the floor's catch probability is unmeasured. The rule is plausible, not proven.
**Cost at stake, stated for the tradeoff** [minority: lane-3/local-repo]: red-lens is the
largest recurring multi-agent line ($9.22–$11.05/round at run 3's 5 agents; the run-4 live-code
shape is 6 lens seats/round at report-scale claim counts — swept round 2 per red R2-2), and its
cost tracks corpus size, not board size — the one half of cost.md finding 2 the table does
support.[^CostAudit]

**Confidence: HIGH** on the structural-blindness analysis and the gate logic; **MEDIUM** on
predicting which branch fires — genuinely open until run 4's red rounds land; **MEDIUM** on the
four-arm rule's sufficiency (post-hoc by construction).

---

## §6 Cross-cutting

### 6.1 Where the money actually is (priority order for this run's dispositions — item 6 is HOLD, not ratified; retitled round 1 per red R1-35(b)) [minority: lane-2/primary-literature for the ranking]

From cost.md at the pin: red-merge $57.54 (38%) > red-lens $49.48 (33%; rounds 1–5 — the
killed r6 spawn adds $0.61; run-3 measured at 5 lenses/round — the run-4 live-code shape is
**6 lens seats/round** at report-scale claim counts, per red R2-2) > blue-respond $18.21 (12%)
> blue setup/synthesis + assembly (one-time).[^CostAudit]

**Projected run-4 judge-seat line (added round 1 per red R1-1 — the round-0 money map priced
run 4 against a dead baseline):** run 3's distribution had ZERO judge dispatches, but PR #15
made that distribution unreachable — the shipped contested-docket dockets every re-raised or
supersedes-descended gap from round 2 on (debate.js ll.244–245), and each dispatch reads
`debate.md` + `red/findings.md` IN FULL at the judgment tier (l.249): the same read profile as
a merge, so ~$10–13 per dispatched round is the honest planning figure (the backlog's own
judge-round estimate is ~$10). Expected docket rounds in run 4: near-certain ≥1 from round 2
onward for as long as any gap survives two rounds — with dispatch frequency scaling with the
CONTESTED subset (chained/re-raised gaps), not the raw board (red's merge tempering, accepted).
Every savings estimate in this report is therefore stated against a run-4 baseline of
(run-3 seat costs **with red-lens rescaled to the citationPasses-implied 6 seats/round —
~+$2/round, ~+$10/run at run-3 per-lens rates; the round-1 concession stopped at the judge
seat and is swept round 2 per red R2-2**) + (judge dispatches × ~$10–13), and the error term
of any estimate that ignores either line is plausibly the size of the savings being ranked —
which is red's point, conceded in full, twice. **Judge read profile under 4a (added round 2
per red R2-5), both directions:** sharding shrinks the judge's resident findings read (the
open ledger replaces the full file — unpriced benefit) while per-chain archive demanded-reads
add targeted read cost back (unpriced); neither direction is measured, and the ~$10–13 figure
assumes the unsharded profile. The carried-gap re-docket loop compounds this line (§6.4
item 6).

Levers ranked by measured target × confidence:

1. **Sharding (4a)** — targets the biggest line at the highest rate premium. Dollar sizing
   corrected round 3 per red R3-1 (§4.2): ≈**$3–6/run** findings-attributable
   (direct-attribution $2.8 from the committed parser; proportional-share ceiling ≈$6 — the
   round-2 "≈$12 measured" priced cache-weighted bytes as tokens), archive fraction **72.6%**
   derived from the pinned file's structure → **sharding-addressable ≈$2–4/run at run-3
   scale**, comparable to item 2's batching saving rather than dollar-dominant. The #1 rank
   now rests on what the dollars never carried alone: the unpriced judge-read benefit (the
   open ledger replaces the judge's full findings read — §6.1 above), the quality argument
   (the merge operates in the measured long-context degradation regime, and shrinking the
   resident archive is judgment-strengthening — §4.2), and growth direction (the archive is
   the component that compounds every round). Seven conditions.
2. **Prompt-level read batching (4b degenerate form)** — one prompt sentence.
3. **Carried-ruling persistence (engine fix, §6.4 item 6)** — inserted round 2 per red R2-8
   (the map's scope includes non-docketed items — the HOLD lever is listed — so omitting the
   report's own engine fix was an inconsistency, not a scoping choice): carrying rulings
   forward eliminates the re-docket loop's repeat component — marginal docket growth every
   round a carried gap stays open, a full ~$10–13 only when it is the docket's sole member
   (priced per red R2-11) — and closes the carried→risk_accepted gate-erosion path; plausibly
   the second-largest actionable saving this report identifies. Engine-successor work, not
   this run's PR.
4. **Grade-dispute channel (3a minimal form)** — near-zero cost; the guardrail any future
   grade-actuated mechanism needs.
5. **Instrumentation (1 + 2)** — board-profile + mass telemetry via the merge-seat append
   (durable sink per §2.5 item 1 — corrected round 2 per red R2-1), `found_by` field;
   near-zero tokens; collects the evidence every deferred actuation decision needs.
6. **Stop-and-resume documentation (1)** — one paragraph.
7. **Round-scoped audit (5)** — **HOLD, not ratified**; no action this run; decision rule
   registered for run's end (listed for completeness of the money map, not as ratified work).

Cross-cutting prior kept from the frontier: termination levers (1, 2) attack round COUNT
(~$25–30/round whole-round elimination); structure levers (4) attack per-round turns × context
at the priciest seats; lever 5 attacks per-round read volume — the ratify/reject calculus
differed accordingly, and the record bears the frontier's prediction out.

### 6.2 The interlocks (stated so no later round loses them) [minority: lane-1/disconfirming-first for the explicit interlock framing]

- Lever 3a is a **mandatory companion** of any future lever-1/2 actuation: spend-controlling
  grades need an adversarial correction path (the §2.3 incentive loop) — valid only with the
  §3.3 accepted-branch clauses (v)–(vii) included; a channel whose accepted branch is dark is
  blue's unaudited deflation lever, not a safety interlock (qualified round 1 per red R1-2,
  propagated from §3.5).
- Lever 4b's batching sentence **replaces** the collator; they are not independent items.
- Lever 5's run-5 disposition is **decided by run 4's propagation record**, mechanically, per
  the winnow list's audit trigger.
- Lever 1/2 instrumentation is the **evidence supply** for every rejected actuation's revisit
  trigger (including best-of-N's per-gap records via 3a) — **conditional on the durable
  merge-seat sink of §2.5 item 1** (verified round 2 per red R2-1: harness `log()` persists
  nowhere; without the named sink and consumer, runs 4–5 produce no mass series and the
  deferred actuation decisions arrive with no evidence base).

**The attestation ceiling (root invariant, stated once — adopted round 2 per red R2-9, whose
four instances this round were all ad-hoc rediscoveries of it inside round-1 repairs):** the
engine has no attestation primitive stronger than self-report for work-done claims. In-run
enforcement reaches exactly two tiers: **shape** (schema checks — required, non-empty, valid
ids — catch omission) and **consistency** (cross-references between independent envelope
structures, like the shipped lineage throw's `closures` vs `gaps[].supersedes`, catch
self-contradiction). Nothing in-run catches **vacuity** — a seat asserting work it did not do.
The enforcement tier for vacuity is post-hoc and named: independent seats over git-tracked
artifacts — the run retrospective and the next run's docket, the demonstrated consumer class
(this run's docket was assembled from run 3's git-tracked record). Conditions 5 and 7 of §4.5,
§2.5 item 2's `found_by` re-derivation clause, and §3.3 clause (v)'s delta trail are
applications of this rule and claim no more than it allows; any future condition reaching for
"required envelope field" or "fails structurally" must state which tier it is buying.

### 6.3 Doctrine check (run against every position ACTUATED THIS RUN — scope stated honestly round 1 per red R1-7) [minority: lane-2/primary-literature for the systematic check]

Every cut ratified for actuation this run lands on instance-redundancy, residency of red's OWN
closed cases, or mechanical collation; no actuated position reduces judge strength, red-merge
depth, distinct-lens coverage, or the spot-check floor; two positions (the §2.5 carried floor
design, §5.5's arm (d)) make the never-zero floor arithmetic-explicit rather than hortatory.
**The named exception the round-0 check silently skipped:** §5.5's conditional run-5 RATIFY
would cut red's read of blue's report — none of the three safe categories, and blue's own §5.3
quotes the controlling doctrine ("this clause outranks any token saving"). That ratification
is future, gated, and now carries an explicit doctrine-amendment prerequisite (§5.5 gate
condition 3): the doctrine text changes with the behavior or the ratification does not
proceed. A doctrine check that claims universal coverage while excluding its hardest case is
worth less than one that names the case.

### 6.4 Defects found in pinned artifacts and this run's own inputs

1. **cost.md carries two internal defects:** (a) finding 2 is internally contradicted ("merge
   cost tracks DISPUTE size" vs its own table: r5 $13.56 > r2 $13.22 on the second-smallest
   board — superlative per red R1-32); (b) finding 3 states the judgment-tier cache-write
   multiplier as 12.5× when its own pricing header (sonnet cache-write 2.5 $/MTok, session
   tier 12.5) makes it 5× — 12.5 is the absolute rate, not the multiplier (added round 1 per
   red R1-30). Both flagged for correction in the artifact's successor; the corrected lesson
   of (a) (merge cost tracks the cumulative archive) is lever 4a's case [minority:
   lane-3/local-repo].[^CostAudit]
2. **The backlog's severity-floor savings claim fails audit** against the round-3 board it cites
   (two MEDIUM-HIGH open gaps) — convergent, all three lanes; and per the round-1 correction
   (§1.2), NO threshold setting reproduces its claimed round-3
   stop.[^BacklogLevers][^FindingsBoard]
3. **This run's frontier carries three misattributions** (count corrected round 1 per red
   R1-19 — the round-0 "two" had a live counterexample): friction #15 concerns blue/report.md,
   not findings.md (corrected §4.1); R4-3 is a weak type specimen for lever 5 (corrected
   §5.1); and frontier H1 grades R5-5 "HIGH" and R4-1 "High likelihood × High impact" against
   the pinned MEDIUM-HIGH (merge-tempered) and certain × high — stale pre-temper lens grades,
   verified first-hand at frontier.md H1 this round, corrected here so the completeness claim
   no longer has a counterexample [minority: lane-3/local-repo for the first two].
4. **The Grep count-mode footgun recurred live** (unanchored '### LEAD' count = 5 prose mentions
   vs 0 anchored headers) — second documented recurrence, supports a lens-prompt documentation
   line [minority: lane-1/disconfirming-first; also logged to friction.md].[^DebateNoLead]
5. **A stale "54KB" size figure circulates in pinned artifacts** (run-3 friction #15 and
   backlog 31(g)): it matches neither pinned file — findings.md is 106,772 bytes,
   blue/report.md 159,394 at the pin — and the round-0 draft reused it attached to a third
   referent before correction (§4.3(c)); flagged for the artifacts' successors (added round 1
   per red R1-34).[^FrictionFifteen]
6. **Engine defect at the pin — the carried-gap re-docket loop (added round 1 per red R1-1(c),
   verified first-hand at debate.js ll.252–253):** `carried` rulings never enter
   `adjudicated`, so a carried gap re-enters the contested docket every subsequent round it
   stays open — (i) repeat judge spend, **priced round 2 per red R2-11 as marginal docket
   growth**: the dispatch is ONE agent call per round covering the whole docket (verified
   ll.247–250), so a re-docketed carried gap costs a full ~$10–13 only when it is the docket's
   sole member — otherwise it grows an already-occurring dispatch — and (ii) a gate-erosion
   drift path: each re-docket is a fresh chance the ruling drifts carried→risk_accepted, and
   `risk_accepted` removes the gap from red's verdict permanently — a gap red keeps re-raising
   can exit the gate by judge attrition rather than by evidence. Graded MEDIUM (likelihood
   medium once dockets arm — near-certain in run 4; impact medium-high on the gate-erosion
   branch; complexity of the fix low: carry the ruling forward without re-dispatch **unless
   red's GRADE changed — the script-visible trigger in `redEnv`; "evidence changed" lives in
   findings prose the script cannot read, so new evidence routes through red re-raising under
   a successor id, the existing lineage path, which re-dockets it by construction — trigger
   restated round 2 per red R2-11, killing the self-report fallback that would let red force
   re-dispatch every round**). The same loop applies to the dispute-resolution value §3.3
   adds to the judge's enum — the resolution-enum bullet, NOT clause (v), which contains no
   enum (pointer corrected round 3 per red R3-9; the round-2 pointer copied red's own R2-11
   phrasing — vector logged against red, error still an error): a dispute ruled `carried`
   re-dockets identically — both traffic classes named. In-run, clause (vi)'s terminal
   exclusion (round 3, per red R3-12) caps the dispute branch of this loop at the exit
   boundary: a terminal dispute cannot be `carried` into a round that does not exist. Flagged for the engine's successor; not a re-recommendation of shipped work
   — the shipped detector is correct, the loop is its unpriced interaction with the `carried`
   ruling class.

---

## §7 Pre-flight self-audit (protocol-required)

- Every substantive claim footnoted (corpus claims to pinned commits with file + location;
  external claims to primary sources): checked; access dates present on all footnotes.
- Minority-claim provenance: every single-lane claim tagged inline with its lane marker;
  convergent claims unmarked by convention (see preamble).
- Disconfirming budget: satisfied at the lane level (lane 1's entire method is
  disconfirming-first; lane 2 logged 4 of 13 searches disconfirming (31%); lane 3's method is
  first-hand verification against the pin). The synthesis added no new claims requiring new
  searches.
- Confidence self-graded per lever section: checked.
- Open questions declared: §8, union of all lanes.
- Red's gap-pattern memory: **not readable from any blue seat this run** (project memory
  directory empty from the lane environments — logged to friction.md by the lanes).
  Substituted per lanes 2/3: the pinned run-3 findings' full gap-class inventory (incomplete
  propagation, count inflation, enum rounding, unenforced-optional-field, stale-cell,
  unquoted-hold classes) — the same content one generation older; this report addresses each
  class explicitly where relevant (propagation: §5 throughout and §6.4.3; enum rounding:
  §4.6.3; unenforced-optional-field: §3.3's default-to-docket; stale-cell/lineage: §4.3(a),
  §4.4).
- Known verification limits, labeled not laundered (rewritten round 1 per red R1-5; the
  enumeration and the "paywalled" adverb corrected round 2 per red R2-18 — the round-1 "ONLY"
  had two in-report counterexamples, one of them two sentences later in the same bullet):
  **four** sources are access-blocked from the verifying seats, and "403" is an access status,
  not a mechanism (bot-block vs paywall unshown): [^ExpertCvss] (ScienceDirect abstract-only;
  moments from search digest), the unverified Computers & Security 2026 "Fragmentation" paper
  (403 at the abstract, attempted round 1; not cited for any figure), the Votta ACM primary
  (403 to both seats — the false-positive-reduction attribution rests on search-level
  corroboration of the EMSE replication, §4.6 item 5), and [^DalalMallows]'s tandfonline
  primary (403 at red's round-1 verifying seat; carried via the Höhle 2016 secondary
  exposition per red's citation ledger — via-secondary disclosure added to the footnote
  round 2); the round-0 draft's ~34% NVD-vs-CNA figure claimed a paywall
  excuse that was FALSE — the paper it was pinned to (arXiv:2508.13644) is open-access, and
  the figure is affirmatively absent from it (misattribution, withdrawn and replaced in §2.2
  with leaf-verified figures from arXiv:2607.05670, extracted round 1 via WebFetch + pdftotext
  — the extraction path the protocol's MUST-try names, skipped at round 0 and recorded as
  friction). Lane 1's PoLL citation was via secondary summaries (primary arXiv id given;
  lane 2 cites the primary); the Votta false-positive attribution rests on search-level
  corroboration of the EMSE replication (ACM primary 403 — §4.6 item 5). All graded MEDIUM in
  the footnotes where applicable.
- Frontier misattributions found during lane work are corrected here, not inherited (§6.4.3).
- **Claim count, echoed into a tracked artifact (per red R2-2 / friction):** ~166 declared
  this round (round 0 ≈132, round 1 ≈148, round 2 ≈160; round-3 growth is the R3-1
  instrument-correction record and the §3.3/§4.5 design clauses). This is the input to the
  live `citationPasses` recompute — `ceil(166/40) = 5`, capped at 4 by the shipped
  `min(4, …)`, so citation instances stay 4 + 2 fixed lenses = the 6-seat red-lens shape
  §2.4/§6.1 price — recorded here and in `blue/CHANGELOG.md` so the rescale is auditable
  against something other than the envelope.

## §8 Open questions carried past this round (union, deduplicated)

1. **Does run 4's blue propagation clause hold at zero unpropagated sites?** Decides lever 5's
   run-5 disposition — the record is being generated by this run; red should specifically probe
   propagation completeness. (All three lanes.)
2. **Merge-seat context decomposition — ANSWERED for run 3 (raw shares round 2; dollar
   attribution corrected round 3, red R3-1):** measured from the extracted run-3 transcripts
   (§4.2's table — findings share grows 0.1%→~30%; blue/report.md the largest merge component
   in rounds 2, 3, and 5, with round 4's lens candidates (33%) and findings.md (32%) both
   exceeding it (20%) — restated round 3 per red R3-4; dollar sizing ≈$3–6/run
   findings-attributable and ≈$2–4/run sharding-addressable — the round-2 $7–10 inherited a
   pricing-step 4× overstatement, §4.2). The instrument is now committed
   (`trajectories/decompose-merge.mjs`), so run 4's re-measurement is one command against its
   own tarball. Remaining open: does run 4's decomposition reproduce the shape, and should
   `scripts/cost-audit.mjs` grow the per-agent timeline (backlog 28(d)) so the measurement is
   standard run output instead of a separate parse? (All three lanes; measurement round 2;
   correction round 3.)
3. **Does the capture-recapture estimate track next-round discovery on runs 4–5?** Decides
   whether any future spend throttle has a valid input. (Lane 1.) (Narrowed round 1 per red
   R1-9: the honesty half — can red-merge attribute lens overlap without inflating
   convergence — is no longer parked here; it is a named ratification condition in §2.5
   item 2, with the audit mechanism specified.)
4. **Does any natural grade dispute occur in run 4?** Recalibrates lever 3a's zero-traffic
   estimate. (Lane 3.) If a dispute reaches the judge, does the "gap real, grade wrong"
   resolution case occur? (Lane 1.) (Restated round 1 per red R1-1: the round-0 form also
   asked whether the lineage docket arms — that is no longer an open question; arming is
   near-certain live code, and the open questions are its per-round COST and the carried-gap
   re-docket loop's observed behavior, §6.4 item 6.)
5. **Generalization:** does an ordinary (non-self-referential) topic show the same late-round
   high-severity discovery pattern, or was run 3's tail an artifact of auditing the engine with
   the engine? (Lanes 1–3.)
6. **Enum→numeric mapping for compound grades — DECIDED round 3** (per red R3-8 and the
   lead's ruling; an owner-less deadline was the defect, and this run is the owner of
   record): `realized` is EXCLUDED from open-gap mass — realized risk is no longer a
   probability; a realized-but-open gap counts in the board profile's open/severity columns
   and contributes 0 to mass. The pinned mapping for runs 4–5's telemetry series: low=1,
   low-medium=1.5, medium=2, medium-high=2.5, high=3, certain=3.5, realized=excluded —
   version-stamped into each logged line; a changed mapping starts a new series (§2.5
   item 1). §2.1's run-3 retrospective series (computed with realized=3.5 under its own
   disclosed mapping) stands as historical and is not comparable to the new series. (Lane 2;
   decided round 3.)
7. **Is a cross-model-family grader reachable from this harness at all** (the best-of-N
   precondition)? If structurally impossible, the backlog's revisit clause should say so.
   (Lane 2.)
8. **The re-scoped floor × degenerate-FAIL guard interaction:** a judge-disposition round that
   closes the whole board yields PASS via red next round, or via judge directly? One
   state-machine sentence needed before any such variant ships. (Lane 2.) (Corrected round 1
   per red R1-1: per-round judge disposition of the CONTESTED subset is live code, not future
   design — this question now concerns only the whole-board variant's remaining increment,
   §1.5.)
9. **Is qmd MCP `get`/`multi_get` the intended archive on-demand path for lever 4a**, and is it
   approved in run environments? Plain Read suffices; qmd makes the targeted fetch cheaper.
   (Lane 3.)

---

## Footnotes

Corpus sources (all pinned per `inputs/PINNED.md`):

[^PinCheck]: Pin equivalence check — `git diff --stat bfa8a3b HEAD -- research/2026-07-12_feov-retrospective/` (empty) and `git diff --stat 5396952 HEAD -- ideas/backlog.md plugins/frank-exchange-of-views/` (empty), run first-hand at the lane-3 seat; working-tree reads are pin-faithful. Local repo `C:/Users/gbloc/Projects/special-circumstances`. Accessed 2026-07-14.
[^BacklogLevers]: "run-3 termination & fairness levers," `ideas/backlog.md` item 30 @ `5396952` — severity-floor spec ("would have ended run 3 at round 3 for ~$10"), risk-mass umbrella + spot-check-floor caveat, grade-dispute channel + best-of-N deferral condition; also the docket-detector item's "red even conceded an error after independently re-verifying blue's rebuttal twice — adversarial self-correction working." The log()-per-transition heartbeat lives in the ADJACENT item (line 31, STILL OPEN list), not item 30 — attribution split round 1 per red R1-21; §1.5's claims about the heartbeat are true at the real location. Accessed 2026-07-14.
[^Backlog28d]: `ideas/backlog.md` item 28(d) @ `5396952` — merge-seat cost analysis: "the driver is TURNS x CONTEXT... red-merge-r1: ~100-150K of material, 2.7M+ cache reads"; levers (1) shard, (2) collator, (3) prompt-level read batching, (4) tooling step-up (beads vs a tiny sc-gaps Go tool); plus a closing note that cost.md should show a per-agent timeline (lever list corrected round 1 per red R1-28 — the round-0 footnote promoted the trailing timeline note to lever (4)). Accessed 2026-07-14.
[^FindingsBoard]: `research/2026-07-12_feov-retrospective/red/findings.md` @ `bfa8a3b` — per-round gap blocks read verbatim by lanes 1 and 3 independently: round-5 block (lines ~135–253), round-4 originals (~425–531), round-3 board (~717–893), round-2 originals (~1080–1210), round-1 originals (~1279–1360); closure records rounds 2–5; R3-2 friction-field history (three rounds unnoticed); merge temperings R5-5 (HIGH→MEDIUM-HIGH) and R5-6 (MEDIUM-HIGH→MEDIUM); R5-2 stale MA-status catch; file length 1364 lines. Accessed 2026-07-14.
[^CostAudit]: `research/2026-07-12_feov-retrospective/cost.md` @ `bfa8a3b` — per-seat-round table (red-merge $7.52/$13.22/$12.64/$10.60/$13.56, Σ$57.54 of $149.95; red-lens $9.22–$11.05/round, Σ$49.48 rounds 1–5, +$0.61 killed r6 spawn per red R1-33; blue-respond $3.95/$3.96/$2.98/$3.05/$4.27, Σ$18.21; rounds 4–5 Σ≈$53); findings: cache traffic 99% of tokens; judgment-tier rates 5× cache-read and 5× cache-write (the file's own finding 3 says "12.5×" but its pricing header — sonnet cache-write 2.5 $/MTok, session tier 12.5 — makes the multiplier 5×; 12.5 is the absolute rate; corrected round 1 per red R1-30, §6.4 item 1(b)); "rounds 1–2 closed 31 gaps ($60-ish); rounds 3–5 closed ~15 mostly-trivial gaps for a similar spend"; finding 2's dispute-size claim (contradicted by its own table — §4.2); round-5 merge 7.87M cache reads / 61 turns vs round-2 5.64M. Accessed 2026-07-14.
[^StopResume]: `cost.md` @ `bfa8a3b`, finding 5: "Stop-and-resume with a reduced maxRounds (cache replay) cost ~$0 and cut ~7 residual rounds; five round-6 lenses were killed mid-spawn for pennies." Accessed 2026-07-14.
[^R4OneDetail]: `red/findings.md` @ `bfa8a3b`, R4-1 original grading (line ~425): "HIGH — certain (already realized in this corpus, not projected) × high × low-medium — corroboration: HIGH (... four of five lenses converged independently, each tracing the code first-hand)". Accessed 2026-07-14.
[^Round4Lenses]: `research/2026-07-12_feov-retrospective/red/candidates/round-4-lens-{1,2,3,5}.md` @ `bfa8a3b` — the lineage-blind-docket finding appears independently in lenses 1, 2, 3, and 5 (lens 2: "Finding 1 (NEW, round 4) — the contested-docket detector is lineage-blind by construction"; lens 5 grades it R4-1 HIGH); verified independently by lanes 1 and 2. Accessed 2026-07-14.
[^R5FiveDetail]: `red/findings.md` @ `bfa8a3b`, R5-5 header (line ~200): "MEDIUM-HIGH — medium × high (telemetry-invisible: an unset or vacuous supersedes leaves contested.length at 0...)". Accessed 2026-07-14.
[^R5FiveSingleton]: `red/candidates/round-5-lens-5.md` @ `bfa8a3b` (lines ~36–103, the full enforcement argument) vs `round-5-lens-4.md` (lines ~110–120: "checked whether 'WITH REGRESSION' is a documented protocol state... Considered raising this... Not raised") and `round-5-lens-2.md` (detector-logic analysis, no enforcement claim). Accessed 2026-07-14.
[^R5OneDetail]: `red/findings.md` @ `bfa8a3b`, R5-1 (line ~135): "certain (static text, read side by side at the merge seat) × medium... three lenses converged independently — lenses 1, 2, 4... every chain link checked against this file's own closure entries." Accessed 2026-07-14.
[^R5OneOverrule]: `debate.md` @ `bfa8a3b`, round-5 RED (line ~738): "one lens's contrary 'no discrepancy' hold was overruled at the merge seat by direct read of report lines 496/727 — logged against red below." Accessed 2026-07-14.
[^R4FourGrep]: `red/findings.md` @ `bfa8a3b`, R4-4 corroboration: "report-wide grep '4th|fourth' at merge: exactly one uncorrected instance, §3 risk-accepted paragraph." Accessed 2026-07-14.
[^Round3Red]: `research/2026-07-12_feov-retrospective/debate.md` @ `bfa8a3b`, round-3 `### RED` — "severity is declining monotonically (round 1: 2 HIGH; round 2: 5 MEDIUM-HIGH...; round 3: 2 MEDIUM-HIGH, both code-trace — every prose gap is now ≤ MEDIUM)". Accessed 2026-07-14.
[^EngineSource]: `plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js` @ `5396952` (288 lines, read in full by lanes 1 and 3) — no-filesystem doctrine comment (lines 32–34), doctrine comment (24–31), envelope schemas incl. `supersedes`/`closures` (62–144, RED at 89–107), compound GRADE enum (line 60), judge dispatch + resolution enum (125–144, 247–257), lineage-enforcement throw (227–235), whole-debate contested window (186, 238–246, 258; lineage filter at 244–245), citationPasses recompute (198), ledger drift clause (205), lens full-re-read prompt naming blue/report.md (212), blue-respond propagation sentence (263). Accessed 2026-07-14.
[^RedAuditorMust]: `plugins/frank-exchange-of-views/agents/red-auditor.md` @ `5396952`, line 13: "BEFORE auditing, YOU MUST re-read the FULL living report in context — a change-summary is a navigation hint, never the audit surface." The named object is blue's living report. Accessed 2026-07-14.
[^BlueMandate]: `plugins/frank-exchange-of-views/agents/blue-researcher.md` @ `5396952`, line 14 — the shipped propagation clause (corrections propagate to ALL sites). Accessed 2026-07-14.
[^Retro3Docket]: Run-3 report §3 graded docket, `research/2026-07-12_feov-retrospective/report.md` @ `bfa8a3b` — rows 15 (R1-13→R2-1→R3-7 grade-correction chain: count 3→2, likelihood retained High by argument — "two independent hits... is a 2-for-2 rate on the triggering conditions"; mechanism narrowed with grade kept), 10 (R2-9 ledger skip-trigger repair), 18 (round-scoped audit hold + sharding as first candidate scoping rule), 23 (R5-1 corrected enumeration, grades untouched; R4-1 + R5-5 lineage detection and enforcement), risk-accepted list. Accessed 2026-07-14.
[^Row6Roster]: `report.md` §3 row 6 @ `bfa8a3b`: "this run's own highest-value catch (false-premise repo verification) came from exactly one lane doing exactly one method"; shipped lane roster + 2-of-N floor per `debate.js` LANE_METHODS @ `5396952`. Accessed 2026-07-14.
[^Row16b]: `report.md` §3 row 16b @ `bfa8a3b` — lens passes are the leaf-node audit work; keeper runs omit `model` so the adversary runs at full strength; the `debate.js` header's KNOWN TRADEOFF note @ `5396952`. Accessed 2026-07-14.
[^PropagationChains]: `report.md` §3 row 23 corrected enumeration + §2.1(b) @ `bfa8a3b`: R1-5→R2-4→R3-4/R3-9; R2-5→R3-10; R2-7→R3-6; R2-8→R3-5→R4-3 — plus R4-4's unpropagated fifth numeral (the retrospective's own count at l.1541 is "four chains in this corpus"; 4 + R4-4 = 5). The quoted phrase is NOT in the retrospective — its verbatim home is `blue-researcher.md` l.14 ("5 chains in run 3's 5 rounds") @ `5396952` ONLY; `debate.js`'s blue-respond prompt (l.263) carries a PARAPHRASE ("5 regressions in 5 rounds"), not the quotation — attribution split round 2 per red R2-15, both lines re-read first-hand at this seat (the round-1 over-attribution copied red's own R1-27 phrasing; a round-2 red lens graded that repair HIGH for matching red's instruction and was overruled by the source read). Re-sourced round 1 per red R1-27; the retrospective is cited for the enumeration only. Accessed 2026-07-14.
[^CeilingDisposition]: "Outstanding gaps at the ceiling — disposition and compromise rationale" + TL;DR, `research/2026-07-12_feov-retrospective/report.md` @ `bfa8a3b`, lines 3–20. Accessed 2026-07-14.
[^DebateNoLead]: `research/2026-07-12_feov-retrospective/debate.md` @ `bfa8a3b` — anchored `grep -n "^### "` returns 11 headers: 6 BLUE (rounds 0–5), 5 RED (rounds 1–5), zero LEAD; an unanchored grep returns 5 prose mentions (the count-mode footgun class, run-3 friction #12 / R5-3), re-verified live by lanes 1 and 3 independently. Accessed 2026-07-14.
[^BlueRound4]: `debate.md` @ `bfa8a3b`, round-4 `### BLUE` closing (lines ~705–712): "No rebuttals this round — every gap was real, at the location red found it, and none was over-graded relative to its fix cost"; round-5 BLUE conceded all six grades. Accessed 2026-07-14.
[^DebateNegotiation]: `debate.md` @ `bfa8a3b`, round-2 `### RED` (lines ~287–289): "Red will accept argued risk-accepts on R2-9 and R2-10 if..." — grade/disposition negotiation resolving in-loop. Accessed 2026-07-14.
[^FrictionFifteen]: `research/2026-07-12_feov-retrospective/friction.md` @ `bfa8a3b`, entry 15 (red-merge-r5): the 25k Read cap vs the living report forced "three windowed reads plus targeted greps" — the named file is blue/report.md (lane-3 correction of this run's frontier). File sizes measured at the pin: findings.md 106,772 bytes; blue/report.md 159,394 bytes. Accessed 2026-07-14.
[^FrictionSix]: `friction.md` @ `bfa8a3b`, entry 6 (red-merge-r3): "gap grades are forced into low/medium/high enums... every compound grade above was rounded; the authoritative grading lives in red/findings.md." Accessed 2026-07-14.
[^FrictionFour]: `friction.md` @ `bfa8a3b`, entry 4 (red-merge-r2): the accidental control condition — `findings.md` refused at a scratchpad path, neutral filename succeeded; "the guard is filename-keyed regardless of directory." Accessed 2026-07-14.
[^FrictionTen]: `friction.md` @ `bfa8a3b`, entry 10 (red-merge-r4): "debate.md and the ledger appended via cat" — four rounds of citation-ledger appends without a write-block incident; Edit on pre-created files worked every round. Accessed 2026-07-14.
[^LedgerPrecedent]: `friction.md` entry 11 @ `bfa8a3b` ("ledger skip-rule held all prior-round confidences") and PR #14's citation-ledger entry per `inputs/already-shipped.md`. Accessed 2026-07-14.
[^CitationLedgerRun3]: `research/2026-07-12_feov-retrospective/red/citation-ledger.md` @ `bfa8a3b`, lines 159–187 — round-5 MA-status entries (first appearance; R5-2 was never a ledgered pair) and line 184's merge-seat overrule of lens 2's six-id claim by mechanical extraction. Accessed 2026-07-14.
[^AlreadyShipped]: `inputs/already-shipped.md` (this run dir) — PR #14 (skeleton, ledger, model knobs, simulator), PR #15 (lineage docket + enforcement throw, compound grades, lens-scoped labels, lane roster + floor, minority tagging, blue propagation clause — merged 2026-07-14, run 4 is its first live exercise), PR #16 (pinning, run-record capture, smoke mode, blue pre-flight), PR #17 (PDF MCPs), PR #18 (qmd recall layer — the source's verify-first instruction discharged round 1 per red R1-31: verified merged pre-pin, merge commit `4a3801c` ancestor of `5396952`, hook + `.mcp.json` artifacts present — red lens 3's first-hand git check, accepted). Accessed 2026-07-14.
[^FrontierH3]: `blue/frontier.md` (this run), H3 — plus the frontier's three misattributions corrected in §4.1/§5.1/§6.4.3: H4/H5's (friction #15 referent; R4-3 specimen) and H1's stale pre-temper grades (R5-5 "HIGH," R4-1 "High likelihood × High impact" — vs pinned MEDIUM-HIGH and certain × high; third item added round 1 per red R1-19). Accessed 2026-07-14.

[^JournalCheck]: Harness `log()` persistence, determined first-hand at the blue seat, round 2 (per red R2-1 / the lead's ruling): (a) this run's own LIVE workflow journal (`~/.claude/projects/.../subagents/workflows/wf_5cefd2a4-35f/journal.jsonl`, 43 lines mid-run) contains only `{"type":"started"}` (22) and `{"type":"result"}` (21) lifecycle events — grep "researching" (debate.js l.52's `log()` string) returns 0; (b) run 3's copied journal @ `bfa8a3b` shows the same shape (87 lines = 46 started + 41 result, zero `log()` output — red's finding, reproduced); (c) `scripts/cost-audit.mjs` read in full first-hand — input glob `agent-*.jsonl` (l.28), zero journal.jsonl references. Conclusion: harness `log()` is operator-console-ephemeral; the journal is the harness's lifecycle log, not a script sink. Accessed 2026-07-14.
[^MergeDecomposition]: Run-3 red-merge context decomposition. Round-2 measurement: `agent-transcripts.tar.gz` @ the pinned run's working tree (7,040,514 bytes, 46 members) extracted to a session scratchpad; each red-merge transcript (identified by cost-audit.mjs's own "Red merge, round N" head match) parsed with a ~70-line node script — tool_result bytes attributed to source files via their tool_use `file_path`/command, cache-weighting = bytes × remaining assistant turns (approximates prefix re-billing at cache-read rate; ignores system prompt and harness overhead). **Instrument correction record (round 3, per red R3-1):** the round-2 script lived in the session scratchpad and was NOT retained — an unreproducible work-done claim under §6.2's own attestation ceiling. The method is now committed as **`trajectories/decompose-merge.mjs`** (this run's directory) and was re-run 2026-07-14 against the pinned tarball: the raw table reproduces exactly (±1 point, all five rounds; per-round totals 173/246/249/188/317KB; round-5 blue ingest 145.7KB; parser validation: 61 assistant turns found for round 5, matching cost.md's 61 turns); the round-2 dollar series does NOT reproduce under the stated 4-bytes/token conversion — it, and lens 3's independent round-3 recompute (Σ$12.4), priced cache-weighted BYTES as tokens, a ≈4× overstatement corrected in §4.2 (strict series ≈$0.26/0.53/0.89/1.16, Σ≈$2.8/run; ceiling cross-check against cost.md's measured per-round cache-read tokens in §4.2). **Archive-fraction derivation (round 3):** pinned findings.md's superseded-preserved blocks (l.340 "Verdict (round 4): FAIL — superseded by round 5, preserved" through EOF l.1364) = 76,356 of 105,223 LF-normalized body bytes = 72.6%, via `awk 'NR<340{a+=length($0)+1} NR>=340{b+=length($0)+1} END{print a, b}' red/findings.md` @ `bfa8a3b`; conservative — closure-status records above the boundary would also archive under sharding. **Retention assumption, stated:** the tarball is gitignored (`**/trajectories/agent-transcripts.tar.gz` — the only .gitignore entry, lens-6 leaf read) and exists only in a run's working tree, never in the git object store; the parser reproduces only where a tarball survives, so each run's decomposition OUTPUT must be committed to the git-tracked record (this footnote + §4.2) at measurement time. Measured 2026-07-14 (round 2); re-derived from the committed instrument 2026-07-14 (round 3).

External sources:

[^AdaptiveStability]: "Multi-Agent Debate for LLM Judges with Adaptive Stability Detection," arXiv:2510.12697 — stopping when the round-over-round KS statistic stays < 0.05 for 2 consecutive rounds; adaptive stops typically fall at rounds 4–7, **full range 2–8** across the 22 reported configurations (Table 2: the three JudgeAnything rows stop at 2, 2, and 8), losing ≈1% or less accuracy (max delta −1.03%) vs fixed 10 rounds. Correction record, amended round 2 per red R2-13 (whose round-1 vector was red's own R1-20 phrasing, copied verbatim — logged by red against red): the round-0 gloss failed only on its "<1%" half (max loss −1.03%); its "rounds 2–8" half MATCHED the table's min–max, and round 1's "~4–7 in the reported configurations" was false as a universal (3 of 22 configurations fall outside, at both ends). The §1.3 body claim — criterion = distributional stability with double confirmation — is unaffected and verified HIGH across three fetches. https://arxiv.org/html/2510.12697v1. Accessed 2026-07-14.
[^DebateRounds]: "Literature Review of Multi-Agent Debate for Problem-Solving," arXiv:2506.00066 — saturation ~2–5 rounds, task-dependent; degradation documented past round 2 on some tasks. https://arxiv.org/html/2506.00066v1. Accessed 2026-07-14. Volatility: preprint; QA-task-shaped, weaker analogue to audit loops.
[^DalalMallows]: "When Should One Stop Testing Software?", Dalal & Mallows, Journal of the American Statistical Association 83(403):872–879 (1988). https://www.tandfonline.com/doi/abs/10.1080/01621459.1988.10478676 — optimal stopping trades testing cost against expected loss from remaining bugs; asymptotic rule keyed to observed discovery count vs. cost ratio. Access note (added round 2 per red R2-18): the tandfonline primary returned 403 at red's round-1 verifying seat; the stated result is carried via the Höhle 2016 secondary exposition (red citation-ledger line 63, graded HIGH-via-detailed-secondary) — via-secondary, disclosed. Accessed 2026-07-14.
[^Stads]: "STADS: Software Testing as Species Discovery," M. Böhme, ACM TOSEM 27(2) (2018), arXiv:1803.02130 — residual risk estimated from the discovery curve (Good-Turing singleton rate); companion "Estimating Residual Risk in Greybox Fuzzing" (ESEC/FSE 2021), https://mboehme.github.io/paper/FSE21.pdf. Accessed 2026-07-14.
[^Sprt]: A. Wald, "Sequential Tests of Statistical Hypotheses" (1945); Wald & Wolfowitz (1948) optimality. Savings band per "The relative efficiency of sequential tests" (arXiv:2603.00216): "for symmetric error bounds, the sequential test reduces the average sample size by at least 36% and by at most 75%" — quoted in full round 2 per red R2-14 (the round-1 quotation dropped the source's second "by" inside quotation marks and omitted the symmetric-error-bounds derivation condition; corrected round 1 per red R1-26 from the round-0 "30–50% typical" gloss not present in the source). The band is conditional on type-I = type-II error bounds; asymmetric regimes may fall outside it. Accessed 2026-07-14.
[^Iso29119]: ISO/IEC/IEEE 29119-2 (Test processes) — risk-based testing as the normative strategy; allocation proportional to risk exposure with optional formal thresholds. Via https://standards.ieee.org/ieee/29119-2/7498/ and "A Taxonomy to Assess and Tailor Risk-based Testing in Recent Testing Standards" (arXiv:1905.10676). Accessed 2026-07-14.
[^CvssInconsistent]: "Shedding Light on CVSS Scoring Inconsistencies: A User-Centric Study on Evaluating Widespread Security Vulnerabilities" (arXiv:2308.15259) — 68% of 59 participants gave different severity ratings for the same vulnerabilities. Accessed 2026-07-14.
[^ConflictingScores]: CORRECTION RECORD (round 1, red R1-5): the round-0 footnote attributed "NVD-vs-CNA disagreement on dual-assessed CVEs (~34%)" to "Conflicting Scores, Confusing Signals: An Empirical Study of Vulnerability Scoring Systems" (arXiv:2508.13644). The figure is affirmatively absent from that paper, which compares four scoring SYSTEMS and contains no NVD-vs-CNA analysis (two independent red full-text fetches; paper is open-access, contrary to the round-0 "paywalled" grade note). Claim withdrawn; replaced in §2.2 by [^CathedralBazaar]. The plausible actual home of a ~34% per-CVE figure — "Fragmentation of CVSS scores in the NVD," Computers & Security (2026), sciencedirect.com/science/article/abs/pii/S0167404826001549 — returned 403 to blue's round-1 fetch and remains unverified; not cited for any figure. Accessed 2026-07-14.
[^CathedralBazaar]: "The Cathedral and the Bazaar of Software Vulnerabilities: From the NVD to the CNAs," Zhang, Massacci & Zhang, arXiv:2607.05670 — leaf-verified round 1 via WebFetch + pdftotext full-text extraction: 44,180 CVEs scored by both NVD and a CNA (Pairwise setting; 72,122 under Consumer-View); "194/266 (73%) ... of CNAs have a median of at least 1 vector divergent" from the NVD (Pairwise; 139/288 = 48% Consumer-View); divergence concentrates in Attack Complexity, User Interaction, and Impact; cross-source model transfer accuracy can drop by 40%. https://arxiv.org/abs/2607.05670. Accessed 2026-07-14.
[^ExpertCvss]: "An expert-based investigation of the Common Vulnerability Scoring System," Computers & Security (2015), https://www.sciencedirect.com/science/article/abs/pii/S0167404815000620 — expert disagreement mean −0.38, variance ~4.46 on 0–10. Moments from search digest (paywalled) — grade MEDIUM; load-bearing only for "expert severity scores disagree materially." Accessed 2026-07-14.
[^RbtTaxonomy]: "A taxonomy of risk-based testing," Felderer & Schieferdecker, STTT 16:559–568 (2014); arXiv:1912.11519 — risk estimates from subjective expert opinion; triangulation recommended; long-term empirical ROI validation thin. Accessed 2026-07-14.
[^CaptureRecaptureEval]: "A Comprehensive Evaluation of Capture-Recapture Models for Estimating Software Defect Content," Briand et al., IEEE TSE, https://ieeexplore.ieee.org/document/852741/ — inspector-overlap estimation of remaining defects as the stop/reinspect decision; "when the number of inspectors is too small, no model is sufficiently accurate and underestimation may be substantial." Accessed 2026-07-14. (Label disambiguated at synthesis: lanes 1 and 2 used the same footnote label for different capture-recapture papers.)
[^CaptureRecaptureDecade]: "Capture-recapture in software inspections after 10 years research — theory, evaluation and application," Petersson, Thelin, Runeson & Wohlin, Journal of Systems and Software 72:249–264 (2004), https://wohlin.eu/jss04-1.pdf — estimator bias with few reviewers; defect-localization mismatch biases estimates. Accessed 2026-07-14.
[^FentonOhlsson]: "Quantitative Analysis of Faults and Failures in a Complex Software System," Fenton & Ohlsson, IEEE TSE 26(8):797–814 (2000) — Pareto fault clustering; modules most fault-prone pre-release are among the least fault-prone post-release. Accessed 2026-07-14.
[^PoLL]: "Replacing Judges with Juries: Evaluating LLM Generations with a Panel of Diverse Models," Verga et al., arXiv:2404.18796 (2024) — panel of smaller judges from DISJOINT model families beats a single large judge at ~1/7 cost; names intra-model bias. Lane 2 cites the primary; lane 1 reached the same result via secondary summaries (https://www.comet.com/site/blog/llm-juries-for-evaluation/ — graded accordingly). Accessed 2026-07-14.
[^NineJudges]: "Nine Judges, Two Effective Votes: Correlated Errors Undermine LLM Evaluation Panels," arXiv:2605.29800 (2026; leaf-fetched by lane 2) — 9 correlated judges ≈ 2 independent votes; panels 8–22 points below independent-voting ideal; best single judge matches or exceeds the panel; aggregation recovers ≤11% of the gap. https://arxiv.org/pdf/2605.29800. Accessed 2026-07-14.
[^PersuasiveDebate]: "Debating with More Persuasive LLMs Leads to More Truthful Answers," Khan et al., ICML 2024 (arXiv:2402.06782) — debate raises non-expert judge accuracy to 76% (LLM) / 88% (human) vs 48%/60% naive baselines. Accessed 2026-07-14.
[^WeakJudges]: "On scalable oversight with weak LLMs judging strong LLMs," Kenton et al., NeurIPS 2024 (arXiv:2407.04622) — debate beats consultancy across ALL tested scenarios; gains over DIRECT question-answering baselines are task-dependent; the stronger-debater effect (better debaters → higher judge accuracy) is more modest than prior studies reported (gloss corrected round 1 per red R1-22 — the round-0 footnote attached task-dependence to the wrong comparison; §3.6's body text was already compatible with the source). Disconfirming bound on debate enthusiasm. Accessed 2026-07-14.
[^ContextLength]: "Context Length Alone Hurts LLM Performance Despite Perfect Retrieval," EMNLP 2025 Findings / arXiv:2510.05381 — 13.9%–85% degradation with input length even at perfect retrieval, persisting when irrelevant tokens are whitespace or masked. https://arxiv.org/pdf/2510.05381. Accessed 2026-07-14.
[^LostMiddle]: "Lost in the Middle: How Language Models Use Long Contexts," Liu et al., TACL 12:157–173 (2024) — U-shaped context use; mid-context material significantly degraded. Accessed 2026-07-14.
[^PromptCaching]: Anthropic prompt-caching documentation (platform.claude.com, leaf-fetched by lane 2) — cache read 0.1× base input; 5-minute write 1.25×; 1-hour write 2×; whole conversation prefix re-billed at read rate every tool turn. Living source — volatility: pricing can change; matches cost.md's assumed rate structure at the pin. Accessed 2026-07-14.
[^Votta]: "Does every inspection need a meeting?", L.G. Votta, ACM SIGSOFT '93, https://dl.acm.org/doi/10.1145/167049.167070 — meetings found few additional defects vs. independent review, at higher cost. The significant false-positive-reduction result traces to the Empirical Software Engineering (Springer) replication, NOT to Votta's own paper (attribution split round 1 per red R1-29; ACM primary 403 to both seats — the split rests on search-level corroboration of the replication record, grade MEDIUM). Accessed 2026-07-14.
[^HandoffLoss]: Multi-agent handoff/compression failure literature — lossy intermediate summaries drop edge-case detail and introduce paraphrase errors that compound downstream (https://www.zartis.com/the-compounding-errors-problem-why-multi-agent-systems-fail-and-the-architecture-that-fixes-it/; https://galileo.ai/blog/why-multi-agent-systems-fail). Accessed 2026-07-14. Volatility: practitioner sources; used for the failure-class shape, not figures — the in-corpus friction #6 instance is the load-bearing evidence.
[^HierSumm]: "A systematic review of long document summarization methods," Neurocomputing (2025), https://www.sciencedirect.com/science/article/pii/S0925231225019599 and hierarchical-merging literature (e.g. NexusSum, arXiv:2505.24575) — hierarchical/map-reduce merging introduces information loss and hallucination; mitigation is contextual augmentation with source text. Accessed 2026-07-14.
[^DiffReview]: Diff-vs-context code-review evidence — diff-only review catches local issues, misses global-invariant/architectural defects; hybrid diff+selective-context is the effective pattern (https://graphite.com/guides/ai-code-review-context-full-repo-vs-diff; https://www.coderabbit.ai/guides/code-context; "Towards Practical Defect-Focused Automated Code Review," arXiv:2505.17928). Accessed 2026-07-14.
[^SafeRTS]: "A safe, efficient regression test selection technique," Rothermel & Harrold, ACM TOSEM 6(2):173–210 (1997), https://dl.acm.org/doi/10.1145/248233.248262 — safe selection excludes no fault-revealing tests, conditional on sound change-impact analysis. Accessed 2026-07-14.
[^YooHarman]: "Regression testing minimization, selection and prioritization: a survey," Yoo & Harman, STVR 22(2):67–120 (2012) — selection keyed to change relevance; unsafe selection risks missing fault-revealing tests. Accessed 2026-07-14.
