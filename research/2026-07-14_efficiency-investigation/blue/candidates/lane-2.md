# Lane 2 — primary-literature lens: efficiency and termination levers for the frank-exchange-of-views debate engine

Run 4, 2026-07-14. Method lens: **primary-literature** (papers, specs, standards — leaf sources
over commentary). Assigned order: hypothesis H2 first, then breadth over H1, H3, H4, H5.
Evidence base per `inputs/PINNED.md`: run-3 retrospective @ `bfa8a3b` (report §3, cost.md,
friction.md), engine + backlog @ `5396952`. Winnow list honored — nothing below re-recommends
PR #14–#18 content.

**Search discipline:** 13 searches/fetches; 4 spent on disconfirming evidence (correlated-judge
panels, weak-judge debate limits, fault-clustering counter-intuition, hierarchical-summarization
loss) — above the 1-in-5 floor. Saturation signal: final searches re-surfaced already-seen
clusters (Rothermel/Harrold + Yoo-Harman, STADS, PoLL).

## Verdict summary (this lane's positions)

| Lever | Position | One-line basis |
|---|---|---|
| (1) Severity-floor termination | **REJECT as specified; RATIFY re-scoped** as a judge-routing trigger with a discovery-decay arm | The floor as specified **never fires anywhere in run 3** — the backlog's "$10 at round 3" savings claim is false against the pinned record |
| (2) Risk-mass-proportional spend | **RATIFY, narrowed**: mass throttles citation-instance redundancy only; distinct-lens floor (1 citation + logic + dark-side) is never scoped down | Mass input proved noise-robust in run 3; late high-value discoveries were 3–4× lens-redundant, so marginal instances are cheapenable redundancy, not adversary |
| (3a) Grade-dispute channel | **RATIFY conditionally — coupled to (2)** | Post-PR-#15 the whole-debate docket window already dockets persisting same-id disputes; the residual (grades on gaps whose fix blue accepts) becomes load-bearing exactly when grades drive spend |
| (3b) Best-of-N grading | **REJECT for now** (backlog's own deferral condition unmet) | Zero surviving-bias instances in the pinned corpus; same-family panels ≈ 2 effective votes of 9 in the primary literature — the precondition for panel gains (disjoint model families) does not exist in this harness |
| (4a) Sharded findings (open ledger / closed archive) | **RATIFY with reopen discipline** | The full-re-read MUST protects red-vs-BLUE (verified at pin); the ledger precedent held but needed drift triggers — the shard needs the archive kept as a leaf-node verification target |
| (4b) Collator seat | **REJECT as an agent seat; adopt the degenerate form** (mechanical concatenation, one prompt line at the merge) | Arithmetic: single-digit turns saved per round vs. a new failure surface; lens attribution IS the corroboration signal a normalizer would destroy |
| (5) Round-scoped audit | **KEEP HELD through run 4; ratify for run 5 only propagation-aware + spot-check floor + live evidence that PR #15's propagation clause holds** | Safe-selection theory: scoping is safe only under a complete impact map; unpropagated sites are by construction outside changed sections |

---

## H2 — Risk-mass-proportional spend (assigned first)

### What the primary literature says the lever is

Risk-proportional effort allocation is not an invention of this project's backlog — it is the
normative core of the international software-testing standard: ISO/IEC/IEEE 29119 makes all
testing risk-based, with resource allocation proportional to item risk exposure and formal
thresholds a tailoring option[^Iso29119]. The economics have a closed form in the optimal-stopping
literature: Dalal & Mallows derive the optimal rule for stopping testing as a trade-off between
the cost of continued testing and expected losses from remaining defects, with the asymptotic
form keyed to the **observed discovery rate against the cost ratio** — not to the severity of
already-known open items[^DalalMallows]. Wald's sequential-test result is the general form:
letting the spend adapt to the accumulating evidence buys 30–50% average savings at unchanged
error rates, **provided the stopping/throttling statistic is the right one**[^Sprt]. The lever's
shape is standard; everything turns on the statistic.

### Is the statistic right? (disconfirming test a — grade noise)

The general-case evidence says severity grading by a lone rater is noisy: 68% of surveyed users
scored the same vulnerabilities differently under CVSS[^CvssInconsistent]; dual-assessed CVE
scores disagree between NVD and CNAs in roughly a third of cases[^ConflictingScores]; an
expert-based CVSS study measured disagreement variance of ~4.5 on a 10-point scale[^ExpertCvss].
The risk-based-testing literature's own self-criticism is the same: risk estimates are subjective
expert opinion and "additional measures should be applied to triangulate"[^RbtTaxonomy].

But the pinned corpus shows FEOV's grades are not lone-rater grades — they are adversarially
argued, and the run-3 correction record shows the **noise landed in the evidence cells, not the
grade cells**. All three grade-correction chains the frontier names were checked at the pin:

- R2-1 (count inflation, §3 row 15): the occurrence count fell 3→2, and the disposition kept
  likelihood High on an explicit re-argument ("two independent hits... is a 2-for-2 rate on the
  triggering conditions")[^Retro3Docket].
- R3-7 (mechanism narrowing, same row): narrowed 2-confirmed to 1-confirmed + 1-plausible;
  grade kept, one-clause caveat added[^Retro3Docket].
- R5-1 (discarded enumeration, row 23): corrected a mis-traced lineage list inside the
  likelihood cell; the High grade was untouched[^Retro3Docket].

Computed mass over the open board would have been **stable through every grade correction run 3
produced**. Disconfirming test (a) does not fire: the throttle input is noisy in the general
literature but was measured-robust in this loop, because the adversarial loop is itself the
triangulation the literature asks for. (This is also why lever 3a is coupled here — see H3.)

### Does throttling cheapen the adversary? (disconfirming test b)

The decisive corpus fact: run 3's late high-value discoveries were **massively lens-redundant**.
R4-1 (lineage-blind docket, the corpus's highest-graded engine finding) was independently minted
by 4 of 5 lenses in round 4 — including all three citation instances, not only the always-on
logic/dark-side lenses (round-4 lens files 1, 2, 3, 5 each carry it; lens 2 labels it "Finding 1
(NEW)")[^Round4Lenses]. R5-1 was converged on by 3 of 5 lenses (findings.md's own corroboration
cell: "three lenses converged independently — lenses 1, 2, 4")[^Retro3Findings]. A round that had
run ONE citation instance plus the two always-on lenses would still have caught both. The
marginal citation instances in low-mass rounds bought convergence, not coverage — and convergence
is redundancy, which doctrine permits cheapening.

Two honest costs of cutting it:

1. **Corroboration signal.** Red grades corroboration partly BY lens convergence ("three lenses
   converged independently" is the HIGH-corroboration argument on R5-1). Fewer instances = weaker
   corroboration grading on what remains. Mitigation: corroboration language should distinguish
   "single lens, leaf-verified" from "unverifiable" — verification depth, not seat count, is the
   load-bearing part.
2. **Unknown unknowns.** Known-open risk mass is NOT residual risk. The capture-recapture
   inspection literature exists precisely because open-item counts don't estimate undiscovered
   defect content, and estimates from few reviewers are biased[^CaptureRecapture]; Fenton &
   Ohlsson's counter-intuition (most fault-prone pre-release modules are among the LEAST
   fault-prone post-release) warns against extrapolating from where defects were already
   found[^FentonOhlsson]. This is the primary-literature form of the doctrine constraint: the
   spot-check floor exists because mass measures the known board, not the residual — **the floor
   never reaches zero** is not a policy preference, it is the statistically forced hedge.

### Ratified shape (this lane's proposal)

- Script computes `mass = Σ likelihood×impact` over unadjudicated open gaps after each merge
  (enum→numeric map including compound grades; the data is already in `redEnv.gaps`).
- Mass throttles **citation-instance count only**: `citationPasses` scales down toward 1 (never
  0) as mass falls and the ledger's coverage of the claim surface rises. The three distinct
  lenses (≥1 citation, logic, dark-side) run full every round — that is the concrete spot-check
  floor.
- Mass NEVER narrows audit scope (that is lever 5, held separately) and never touches red-merge
  or the judge.
- Measured stake, honestly modest: red-lens totaled $49.48 of $149.95 in run 3 (33%). Dropping
  rounds 3–5 from 5 lenses to 3 saves ~$4/round ≈ **$12–18/run (~10%)** plus slightly cheaper
  merges (fewer candidate files). This is the second-biggest recurring line item after the merge
  seat, but the savings are real-not-dramatic; the case for ratifying is that the cut provably
  lands on redundancy (the 4-of-5 convergence evidence), not that it is large.

**Confidence: HIGH** on the corpus facts (all verified at pin), **MEDIUM** on the savings
estimate (extrapolated from one run's lens redundancy pattern).

---

## H1 — Severity-floor termination

### The corpus kills the lever as specified

The backlog's cost-benefit cell — "would have ended run 3 at round 3 for ~$10"[^BacklogLevers] —
is **contradicted by run 3's own round-by-round board**, read at the pin:

- End of round 3: red's convergence statement reads "round 3: 2 MEDIUM-HIGH, both code-trace —
  every prose gap is now ≤ MEDIUM"[^Round3Red]. Two open gaps above MEDIUM → the floor
  ("every open gap ≤ MEDIUM with trivial fix cost") does not arm.
- End of round 4: R4-1 is OPEN at HIGH[^Retro3Findings]. Does not arm.
- End of round 5: R5-5 is OPEN at MEDIUM-HIGH[^Retro3Findings]. Does not arm.

**The severity floor never fires at any round boundary of the only measured run.** Its claimed
saving is unrealizable on the evidence that motivated it. And weakening the threshold until it
DOES fire at round 3 would have terminated discovery before round 4 minted R4-1 (High×High — the
finding that became PR #15's flagship lineage fix) and round 5 minted R5-5 (the enforcement throw
that shipped) — i.e., the loosened floor deletes exactly the rounds that produced the corpus's
most valuable engine findings, for a saving of ~$53 (rounds 4–5 seat-round sum from
cost.md)[^Retro3Cost]. The frontier's disconfirming test fires in full.

### What the primary literature says the right stopping statistic is

Every principled stopping rule in the testing literature keys on the **discovery process**, not
on the severity of the known-open board: Dalal & Mallows stop when the observed discovery count
falls below a cost-ratio function[^DalalMallows]; Böhme's STADS framework estimates residual risk
from the discovery curve itself (Good-Turing: the singleton rate estimates the probability the
next probe finds something new)[^Stads]. A board of all-trivial KNOWN gaps says nothing about
undiscovered gaps — run 3 is the live demonstration: rounds whose open boards were converging
("severity is declining monotonically"[^Round3Red]) still minted HIGH-grade novel findings,
because the report's subject (the engine) kept yielding new seams. Known-open severity and
residual discovery risk are different quantities; the specified floor conflates them.

### Re-scoped shape worth ratifying

Arm the floor only when BOTH hold at a round boundary:

1. every unadjudicated open gap ≤ MEDIUM with low/trivial fix cost (the original condition), AND
2. the round just completed minted **zero new gaps above the floor** (the discovery-decay
   condition — the Good-Turing-shaped arm: recent discovery predicts undiscovered mass).

On arming, **route the board to the judge for disposition** — not terminate. The judge can still
carry gaps and continue the loop. This keeps the engine's own doctrine intact (termination is
judged, never counted; the floor is a dispatch trigger, not a counter) at low complexity: set
arithmetic over data already in the envelope, plus one judge-mandate extension, plus simulator
cases.

Honesty clause: even re-scoped, this lever **never fires on run 3's data** (every round minted
something above the floor). Its value is unproven on the measured corpus; the case is that it is
cheap, doctrine-clean, structurally incapable of the mis-fire the specified version invites, and
pays only on genuinely-converged runs (plausible for narrower topics than "audit the engine
itself" — run-3's topic endogeneity, where the audit surface WAS the engine, likely inflated
late-round yield). Prediction registered for red: if runs 4–5 end with a final round whose open
board is all-≤-MEDIUM AND whose new-mint list is empty, the re-scoped floor would have saved
exactly one round's spend (~$25–30) at zero verdict cost.

**Confidence: HIGH** on the never-fires finding (three board states read verbatim at pin),
**MEDIUM** on the re-scoped variant's expected value (no corpus instance where it arms).

---

## H3 — Blue grade-dispute channel; best-of-N grading

### The dispute channel: mostly subsumed, residually valuable, coupled to H2

Run-3 facts at the pin: grade disputes were real and rode the general gap loop as prose (row 15's
likelihood cycle R1-13→R2-1→R3-7; impact re-grades R2-9, R5-2)[^Retro3Docket]; corrections landed
in both directions (red conceded after independently re-verifying blue's rebuttal twice — the
backlog's own account[^BacklogLineage]); and the judge was dispatched ZERO times in five rounds —
zero `### LEAD` headers in debate.md, grep-confirmed by red itself (R5-3
corroboration)[^Retro3Findings].

But the zero-dispatch fact is a PRE-PR-#15 fact. The shipped detector's docket window is now the
whole debate (`allPriorGapIds` accumulates every round; `contested` matches any re-raised id or
supersedes-descendant — `debate.js` at `5396952`, contested filter)[^EngineSource]. Any gap that
stays open across two rounds now auto-dockets. A grade dispute that keeps its gap open therefore
ALREADY reaches the judge under shipped mechanics — the run-3 disputes (persisting same-id gaps)
would all have docketed. What remains uncovered is precisely one case: **blue accepts the fix but
disputes the grade** — the gap closes, so nothing persists, so nothing dockets, and a
systematically inflated grade survives to the record unadjudicated.

That residual is nearly worthless today — grades on closed gaps decorate the record. It becomes
**load-bearing the moment lever 2 ships**, because grades then drive next-round spend: an
inflated likelihood×impact buys unnecessary lens instances; a deflated one starves the round.
Wrong grades stop being cosmetic and start being budget errors. Hence: ratify the
`grade_disputes` envelope field (`{gap_id, dimension, proposed, evidence}`) + auto-docket on red's
re-rejection, as a rider on lever 2 — near-zero mechanism (one schema field, one filter clause,
simulator case), and the CVSS-grade noise literature says a structured contest path over severity
grades is hygiene any scoring system needs[^CvssInconsistent]. If lever 2 is rejected, this lane
would drop the channel too — without spend coupling, the shipped docket window covers everything
that matters.

### Best-of-N grading: the literature's precondition is absent here

The panel-evaluation literature is unambiguous about WHERE panel gains come from: PoLL beat a
single large judge with a panel of smaller models **from disjoint model families**, at 7×
lower cost, and names intra-model bias as the thing panels dilute[^PoLL]. The 2026 follow-up is
the direct disconfirmation for same-family panels: nine correlated judges delivered "about 2
independent votes' worth of information," panels underperformed the independent-voting ideal by
8–22 points, and **the single best judge matched or exceeded the full panel** — aggregation
tricks recover at most 11% of the gap[^NineJudges]. FEOV runs one provider; a best-of-N grading
panel in this harness is the correlated-errors case, not the PoLL case.

Meanwhile the adversarial-first alternative has primary support: debate raises judge accuracy
over consultancy and no-assistance baselines (76% vs 48% for LLM judges)[^PersuasiveDebate], with
the honest caveat that gains are task-dependent and shrink in some regimes[^WeakJudges]. And the
corpus shows the backlog's own deferral condition ("best-of-N panel grading only if runs 4–5 show
lone-voice bias survives"[^BacklogLevers]) is unmet: every grade error in run 3 was caught and
corrected inside the loop; no surviving-bias instance exists in the pinned record. **Reject
best-of-N now.** Revisit trigger stands as written in the backlog — per-gap records from runs 4–5
showing bias that survived the loop — plus a second precondition this lane adds from the
literature: a genuinely independent second grader (different model family) must be available,
or the panel is decorative.

**Confidence: HIGH** on subsumption-by-shipped-docket (source read at pin), **HIGH** on the
best-of-N rejection (convergent primary literature + zero corpus instances), **MEDIUM** on the
residual channel's value (contingent on H2's ratification).

---

## H4 — Sharded findings + collator stage

### The cost physics, spec-verified

The Anthropic prompt-caching documentation confirms the backlog's TURNS × CONTEXT mechanism at
the specification level: in multi-turn tool use the entire accumulated prefix is re-billed at
cache-read rate (0.1× base input) on every API turn, with new content written at 1.25×/2×
(5-minute/1-hour)[^PromptCaching]. cost.md's rate assumptions match (sonnet 0.2/2.5 on base 2;
session model 1/12.5 on base 10)[^Retro3Cost]. So a seat's cost ≈ turns × average-context ×
rate — and the judgment-seat rate premium (5× read, 12.5× write over bulk) multiplies the
red-merge seat specifically. Run-3 red-merge: $57.54 across five rounds, 38% of the whole run,
~129K tokens of context re-billed per turn at round 5 (7.87M cache reads / 61 turns)[^Retro3Cost].
Sharding attacks the CONTEXT factor at the priciest seat. This is the largest recurring line item
any docketed lever touches.

### The full-re-read principle does not cover red-vs-own-archive

Verified at the pin: `red-auditor.md`'s MUST reads "re-read the FULL living report in context" —
the object of the mandate is **blue's report** (the adversarial surface), and the corresponding
prompt line in `debate.js` names `blue/report.md`[^EngineSource]. No MUST anywhere in the agent
file or the dispatch prompts requires red to re-read its own closed cases. The frontier's H4
premise holds at the spec level: an open-ledger/closed-archive split does not touch the
full-re-read principle's letter or its purpose (decontextualized diffs of the AUDITED document
mislead; red's own closure records are not the audited document).

Long-context primary literature adds a quality argument, not just a cost one: models use
information at the edges of long contexts far better than the middle, and degradation on
mid-context material is substantial[^LostMiddle]. A merge seat whose working set is "open items +
this round" is not merely cheaper — its attention is concentrated on the material that decides
the verdict.

### The two disconfirming tests, taken seriously

**(a) R5-1's catch pathway.** R5-1 was caught by lenses reading `red/findings.md`'s closure
status lines verbatim against blue's row-23 lineage claims — and R1-13/R2-1's closure records
were CLOSED-archive material at that moment[^Retro3Findings]. Under sharding, that material is
out of the merge's default read. The catch survives only if lineage/closure claims in blue's
report are treated as **citations whose leaf source is the archive** — the lens's existing
leaf-node verification mandate then forces the on-demand archive read. This must be an explicit
prompt line ("a claim about a closed gap is verified against `red/archive.md`, never from
memory"), not an assumption. With it, the R5-1 class is covered by mode-3 leaf-node fetch; without
it, sharding is blind to exactly the stale-lineage class run 3 produced. Ratification is
conditional on this line.

**(b) Precedent honesty.** The citation ledger (the same closed-items-don't-reopen pattern) held
all prior confidences through run 3 with zero observed regressions (friction #11: "ledger
skip-rule held all prior-round confidences")[^Retro3Friction] — but the pattern was NOT free: its
prose-only skip-trigger suppressed source-drift re-checks and had to be repaired by adding
drift/time triggers (row 10, R2-9; the repaired clause is in the shipped
`ledgerClause`)[^Retro3Docket][^EngineSource]. The findings shard needs the analogous reopen
triggers designed in from day one: a closed case reopens into the ledger when (i) a new gap's
`supersedes` names it, (ii) blue's report cites it, or (iii) a spot-check samples it (nonzero
floor, as everywhere).

**(c) Write-path cost cell.** Red's every living-artifact write pays the filename-keyed
write-block copy detour (friction #4, #8, #10: `findings.md` refused even in a scratchpad;
neutral name + `cp` succeeded)[^Retro3Friction]. Sharding doubles red's living artifacts. Both
new files MUST be in the pre-created blackboard skeleton (append/Edit path proven), and neither
may be named `findings.md`/`report.md` — e.g. `red/open-ledger.md` + `red/archive.md`. Otherwise
the lever's cost cell inherits a per-round detour tax the docket never counted. (The still-open
backlog item "sanctioned write path for red's living artifacts" is the durable fix; the skeleton
is the shipping mitigation.)

### The collator: the arithmetic does not support a seat

What the collator saves: the merge currently reads ~5 lens candidate files in separate turns.
At round-5 merge rates (~$0.13/turn of context re-billing[^Retro3Cost]), collapsing 5–8 read
turns saves roughly **$1–2/round** — against which the collator seat costs its own agent run,
adds a null-return failure surface (the class that crashed run 2), and adds another
write-blocked filename to manage. Worse, normalization is not mechanics here: the merge grades
corroboration BY lens convergence (R5-1's HIGH rests on "three lenses converged independently"),
and lens-scoped labels (L2-F1) plus quoted anchors are the merge's dedup-vs-distinct raw
material. Run 3's merge did real judgment at exactly this boundary (findings.md carries a
"Merge-time verification and dedupe notes" section)[^Retro3Findings]. The inspection literature's
one directly-relevant result cuts the same way: Votta found collection meetings added few defects
over independent reviews BUT were significantly better at false-positive reduction — the
consolidation step's value is judgment (killing false positives), not collation
mechanics[^Votta]. And the hierarchical-summarization literature documents information loss and
hallucination at merge layers, with the standard mitigation being "keep the source text
alongside"[^HierSumm] — which degenerates the collator to concatenation.

Concatenation does not need a seat. **Adopt the degenerate form:** one added sentence in the
red-merge prompt — first action, `cat round-N-lens-*.md > round-N-all.md` via Bash, then read the
single file — captures the whole turn saving for zero new seats, zero normalization risk, zero
schema. Reject the collator as designed; log the degenerate form as the shipped alternative.

**Confidence: HIGH** on the doctrine-compatibility of sharding and on the collator arithmetic
(both source/spec-verified), **MEDIUM** on the sharding savings magnitude (depends on how much of
merge context is findings-file vs. transcript — needs one instrumented run; propose measuring in
run 4's cost.md before PR).

---

## H5 — Round-scoped audit (held for cause)

The hold was correct and should survive this run. Three findings:

**1. The safe-selection analogy is exact, and it cuts against naive scoping.** Regression test
selection is "safe" only when it excludes no fault-revealing test — and every safety proof rests
on a complete change-impact analysis[^SafeRTS][^YooHarman]. The prose analog of impact analysis
is "every site stating a corrected claim," and run 3's dominant blue regression class was
precisely incomplete propagation (5 chains in 5 rounds; R4-3's type-specimen sentence was one
R3-5 left UNEDITED — in a section unchanged that round, invisible to any changed-sections
audit)[^Retro3Docket]. A changed-sections scope is an UNSAFE selection by construction against
this measured failure class.

**2. The feasibility of propagation-aware scoping has materially improved since the hold.** The
mechanical expansion (for each accepted correction, add all sites stating the corrected claim to
the audit surface) was grep-cheap but paraphrase-blind when row 18 was held. PR #18's recall
layer (lex + vec + hyde retrieval over the corpus, hook-refreshed on every markdown write) is a
paraphrase-tolerant site-finder — the audit surface expansion can now be computed semantically,
not just lexically. This does not remove the need for the spot-check floor over unchanged
sections (unknown-unknowns again: Fenton & Ohlsson's counter-intuition warns specifically against
concentrating inspection where defects were previously found[^FentonOhlsson] — spot-checks must
sample unchanged AND previously-clean sections), but it converts the ratification condition from
"expensive judgment" to "cheap mechanics plus a floor," which is the doctrine-approved direction.

**3. The decision is gated on evidence this very run is generating.** PR #15's blue-side
propagation clause ("propagate every correction to ALL sites") ships untested — run 4 is its
first live trial[^AlreadyShipped]. Per the winnow rule, the shipped fix is auditable only on this
run's own evidence. Disposition this lane proposes: keep round-scoping HELD through run 4; at
run's end, read run 4's red rounds for propagation regressions. If the clause held (zero
unpropagated-site regressions), ratify round-scoped audit for run 5 **with all three conditions**
(propagation-aware surface expansion via grep + semantic retrieval; nonzero spot-check floor
including unchanged sections; archive/ledger reopen triggers per H4). If the clause failed even
once, reject round-scoping outright for run 5 — it would remove the only backstop for the
engine's measured dominant regression class while that class is demonstrably still live.

**Confidence: HIGH** on the unsafe-by-construction analysis, **HIGH** on the gating logic (it is
this run's own design), **LOW** on predicting which branch fires — genuinely open until run 4's
red rounds land.

---

## Cross-cutting: where the money actually is (priority order for the synthesis)

From cost.md at the pin[^Retro3Cost], per-run recurring spend: red-merge $57.54 (38%) > red-lens
$49.48 (33%) > blue-respond $18.21 (12%) > blue setup/synthesis+assemble (one-time). Levers
ranked by measured target × ratification confidence:

1. **H4 sharding** — targets the biggest line at the highest rate premium; ratify with the three
   conditions (leaf-node archive line, reopen triggers, skeleton coverage of both new files).
2. **H2 narrowed mass-throttle** — targets the second line; provably cuts redundancy (4-of-5
   convergence evidence); ~10% of run cost.
3. **H3a dispute channel** — near-zero cost, rides H2's PR; ships the guardrail H2's input needs.
4. **H1 re-scoped floor** — cheap and doctrine-clean but never fires on measured data; ship as
   set-arithmetic + judge-mandate extension, expect zero savings until a converging run occurs.
5. **H4b collator degenerate form** — one prompt sentence; the seat itself rejected.
6. **H5** — no action this run; decision rule registered for run's end.
7. **H3b best-of-N** — rejected; revisit condition unchanged plus the model-family-independence
   precondition.

Doctrine check, run against every position above: every cut lands on instance-redundancy,
re-read volume of red's OWN closed cases, or mechanical collation; no position reduces judge
strength, red-merge depth, distinct-lens coverage, or the spot-check floor, and two positions
(H2 floor, H5 conditions) make the never-zero floor arithmetic-explicit rather than hortatory.

## Open questions (carried to synthesis)

1. What fraction of red-merge's per-turn context is findings.md vs. debate.md vs. blue's report?
   The sharding savings estimate needs this split — run 4's cost audit could bin cache-read
   composition per seat (instrument before the PR, not after).
2. Does run 4 (this run) produce any propagation regression under PR #15's shipped clause? —
   the H5 gate. Register: red should specifically probe propagation completeness this run.
3. If H2 ships, what is the enum→numeric mapping for compound grades (and does `realized` count
   in open-gap mass at all, since realized risk is no longer a probability)?
4. Is a cross-model-family grader reachable from this harness at all (the best-of-N
   precondition)? If structurally impossible, the backlog's revisit clause should say so instead
   of implying a panel is one decision away.
5. The re-scoped H1 floor and the shipped degenerate-FAIL guard interact: a judge-disposition
   round that closes the whole board yields PASS via red next round, or via judge directly?
   The termination path needs one state-machine sentence to avoid a new degenerate shape.

## Pre-flight self-audit (protocol-required)

- Every substantive claim footnoted (corpus claims to pinned commits with file+location; external
  claims to primary sources with access dates): checked.
- Disconfirming budget: 4 of 13 searches (31%) — above floor: checked.
- Confidence self-graded per section: checked.
- Open questions declared: 5, above.
- Red gap-pattern memory: **not readable from this seat** (project memory dir for the
  special-circumstances project is empty from this lane's environment) — substituted the run-3
  findings' full gap-class inventory read at the pin (propagation chains, count inflation,
  enum rounding, unenforced-optional-field, stale-cell classes), which is the same content one
  generation older. Logged as friction.
- Known verification limits, labeled not laundered: the 34.1% NVD-vs-CNA disagreement figure and
  the −0.38/4.46 expert-CVSS moments are from search-result digests of the cited papers, not
  leaf-verified against the papers' tables (paywalled/PDF; not load-bearing — the 68% figure from
  the arXiv-open study carries the point). Graded MEDIUM and marked in the footnotes.

## Footnotes

[^DalalMallows]: "When Should One Stop Testing Software?", S.R. Dalal & C.L. Mallows, Journal of
the American Statistical Association 83(403):872–879 (1988). https://www.tandfonline.com/doi/abs/10.1080/01621459.1988.10478676 — optimal stopping trades testing cost against expected loss
from remaining bugs; asymptotic rule keyed to observed discovery count vs. cost ratio. Accessed
2026-07-14.

[^Iso29119]: ISO/IEC/IEEE 29119-2 (Test processes), risk-based testing as the normative strategy;
allocation proportional to risk exposure with optional formal thresholds. Via IEEE SA standard
page https://standards.ieee.org/ieee/29119-2/7498/ and "A Taxonomy to Assess and Tailor
Risk-based Testing in Recent Testing Standards" (arXiv:1905.10676). Accessed 2026-07-14.

[^Sprt]: A. Wald, "Sequential Tests of Statistical Hypotheses" (1945); Wald & Wolfowitz (1948)
optimality. Savings figures (30–50% typical; ≥36% for symmetric error bounds) per "The relative
efficiency of sequential tests" (arXiv:2603.00216) and the Springer introduction to Wald (1945).
Accessed 2026-07-14.

[^CvssInconsistent]: "Shedding Light on CVSS Scoring Inconsistencies: A User-Centric Study on
Evaluating Widespread Security Vulnerabilities" (arXiv:2308.15259) — 68% of 59 participants gave
different severity ratings for the same vulnerabilities. Accessed 2026-07-14.

[^ConflictingScores]: "Conflicting Scores, Confusing Signals: An Empirical Study of Vulnerability
Scoring Systems" (arXiv:2508.13644) — NVD-vs-CNA disagreement on dual-assessed CVEs (~34%).
Figure taken from search digest, not leaf-verified against the paper's tables — grade MEDIUM.
Accessed 2026-07-14.

[^ExpertCvss]: "An expert-based investigation of the Common Vulnerability Scoring System",
Computers & Security (2015), https://www.sciencedirect.com/science/article/abs/pii/S0167404815000620 —
expert disagreement mean −0.38, variance ~4.46 on 0–10. Moments from search digest (paywalled
abstract), not leaf-verified — grade MEDIUM; not load-bearing beyond "expert severity scores
disagree materially". Accessed 2026-07-14.

[^RbtTaxonomy]: "A taxonomy of risk-based testing", Felderer & Schieferdecker, STTT 16:559–568
(2014); arXiv:1912.11519 — risk estimates from subjective expert opinion; triangulation
recommended; long-term empirical ROI validation thin. Accessed 2026-07-14.

[^CaptureRecapture]: "Capture-recapture in software inspections after 10 years research — theory,
evaluation and application", Petersson, Thelin, Runeson & Wohlin, Journal of Systems and Software
72:249–264 (2004), https://wohlin.eu/jss04-1.pdf — estimating remaining defects from reviewer
overlap; estimator bias with few reviewers; defect-localization mismatch biases estimates.
Accessed 2026-07-14.

[^FentonOhlsson]: "Quantitative Analysis of Faults and Failures in a Complex Software System",
Fenton & Ohlsson, IEEE Transactions on Software Engineering 26(8):797–814 (2000) — Pareto fault
clustering; counter-intuitive result that modules most fault-prone pre-release are among the
least fault-prone post-release. Accessed 2026-07-14.

[^Stads]: "STADS: Software Testing as Species Discovery", M. Böhme, ACM TOSEM 27(2) (2018),
arXiv:1803.02130 — residual risk and undiscovered-species probability estimated from the
discovery curve (Good-Turing singleton rate); companion "Estimating Residual Risk in Greybox
Fuzzing" (ESEC/FSE 2021), https://mboehme.github.io/paper/FSE21.pdf. Accessed 2026-07-14.

[^PoLL]: "Replacing Judges with Juries: Evaluating LLM Generations with a Panel of Diverse
Models", Verga et al. (arXiv:2404.18796, 2024) — panel of smaller judges from DISJOINT model
families beats a single large judge at ~1/7 cost; names intra-model bias. Accessed 2026-07-14.

[^NineJudges]: "Nine Judges, Two Effective Votes: Correlated Errors Undermine LLM Evaluation
Panels" (arXiv:2605.29800, 2026; leaf-fetched abstract/body this run) — 9 correlated judges ≈ 2
independent votes; panels 8–22 points below independent-voting ideal; best single judge matches
or exceeds the panel; aggregation recovers ≤11% of the gap. Accessed 2026-07-14.

[^PersuasiveDebate]: "Debating with More Persuasive LLMs Leads to More Truthful Answers", Khan et
al., ICML 2024 (arXiv:2402.06782) — debate raises non-expert judge accuracy to 76% (LLM) / 88%
(human) vs 48%/60% naive baselines. Accessed 2026-07-14.

[^WeakJudges]: "On scalable oversight with weak LLMs judging strong LLMs", Kenton et al., NeurIPS
2024 (arXiv:2407.04622) — debate gains over consultancy are task-dependent and smaller than prior
studies; mixed outside information-asymmetry regimes. Disconfirming bound on debate enthusiasm.
Accessed 2026-07-14.

[^LostMiddle]: "Lost in the Middle: How Language Models Use Long Contexts", Liu et al., TACL
12:157–173 (2024) — U-shaped context use; mid-context material significantly degraded. Accessed
2026-07-14.

[^PromptCaching]: Anthropic prompt-caching documentation (platform.claude.com, leaf-fetched this
run) — cache read 0.1× base input; 5-minute write 1.25×; 1-hour write 2×; whole conversation
prefix re-billed at read rate every subsequent tool turn. Living source — volatility: pricing
can change; matches cost.md's assumed rate structure at the pin. Accessed 2026-07-14.

[^Votta]: "Does every inspection need a meeting?", L.G. Votta, ACM SIGSOFT '93,
https://dl.acm.org/doi/10.1145/167049.167070 (with Springer replication "Does Every Inspection
Really Need a Meeting?", Empirical Software Engineering) — meetings found few additional defects
vs. independent review but significantly reduced false positives, at higher cost. Accessed
2026-07-14.

[^SafeRTS]: "A safe, efficient regression test selection technique", Rothermel & Harrold, ACM
TOSEM 6(2):173–210 (1997), https://dl.acm.org/doi/10.1145/248233.248262 — safe selection excludes
no fault-revealing tests, conditional on sound change-impact analysis. Accessed 2026-07-14.

[^YooHarman]: "Regression testing minimization, selection and prioritization: a survey", Yoo &
Harman, STVR 22(2):67–120 (2012) — selection keyed to change relevance; unsafe selection risks
missing fault-revealing tests. Accessed 2026-07-14.

[^HierSumm]: "A systematic review of long document summarization methods", Neurocomputing (2025),
https://www.sciencedirect.com/science/article/pii/S0925231225019599 and hierarchical-merging
literature (e.g. NexusSum, arXiv:2505.24575) — hierarchical/map-reduce merging introduces
information loss and hallucination; mitigation is contextual augmentation with source text.
Accessed 2026-07-14.

[^Retro3Cost]: Run-3 cost audit, `research/2026-07-12_feov-retrospective/cost.md` @ `bfa8a3b` —
per-seat-round table (red-merge $7.52–$13.56/round; red-lens $9.22–$11.05/round; totals computed
this lane: red-merge Σ$57.54, red-lens Σ$49.48, rounds 4–5 Σ$53.00 of $149.95); findings list
(cache = 99% of tokens; lens cost tracks corpus size; merge cost tracks dispute size; judgment
rates 5×/12.5×). Accessed 2026-07-14.

[^Retro3Docket]: Run-3 report §3 graded docket, `research/2026-07-12_feov-retrospective/report.md`
@ `bfa8a3b` — rows 15 (R1-13/R2-1/R3-7 grade-correction chain, grade retained), 10 (R2-9 ledger
skip-trigger repair), 18 (round-scoped audit hold + sharding as first candidate scoping rule),
23 (R4-1 + R5-5, lineage detection and enforcement), risk-accepted list. Accessed 2026-07-14.

[^Retro3Findings]: Run-3 red findings (in-report full copy), same file @ `bfa8a3b` — R4-1 OPEN
HIGH; R5-1 OPEN MEDIUM with "three lenses converged independently — lenses 1, 2, 4"; R5-3
corroboration (zero `### LEAD` headers, grep-verified); R5-5 OPEN MEDIUM-HIGH; merge-time
verification and dedupe notes section. Accessed 2026-07-14.

[^Round3Red]: Run-3 debate transcript round-3 `### RED`,
`research/2026-07-12_feov-retrospective/debate.md` @ `bfa8a3b` — "severity is declining
monotonically (round 1: 2 HIGH; round 2: 5 MEDIUM-HIGH...; round 3: 2 MEDIUM-HIGH, both
code-trace — every prose gap is now ≤ MEDIUM)". Accessed 2026-07-14.

[^Round4Lenses]: Run-3 red lens candidates @ `bfa8a3b`,
`research/2026-07-12_feov-retrospective/red/candidates/round-4-lens-{1,2,3,5}.md` — the
lineage-blind-docket finding appears independently in lenses 1, 2, 3, and 5 (lens 2: "Finding 1
(NEW, round 4) — the contested-docket detector is lineage-blind by construction"; lens 5 grades
it R4-1 HIGH). Accessed 2026-07-14.

[^Retro3Friction]: Run-3 friction harvest, `research/2026-07-12_feov-retrospective/friction.md` @
`bfa8a3b` — entries #4/#8/#10 (filename-keyed write-block, copy detour every red merge), #11
(ledger skip-rule held all prior confidences), #15 (25k Read cap vs 54KB living report at the
merge seat). Accessed 2026-07-14.

[^BacklogLevers]: `ideas/backlog.md` @ `5396952`, "run-3 termination & fairness levers" item —
severity-floor spec ("would have ended run 3 at round 3 for ~$10"), risk-mass umbrella with
spot-check-floor caveat, grade-dispute channel with best-of-N deferral condition. Accessed
2026-07-14.

[^BacklogLineage]: `ideas/backlog.md` @ `5396952`, docket-detector item — "red even conceded an
error after independently re-verifying blue's rebuttal twice — adversarial self-correction
working". Accessed 2026-07-14.

[^AlreadyShipped]: `inputs/already-shipped.md` (this run dir) — PR #15 blue propagation clause
shipped 2026-07-14; run 4 is its first live trial. Accessed 2026-07-14.

[^EngineSource]: `plugins/frank-exchange-of-views/` @ `5396952` — `scripts/debate.js`: contested
filter over `allPriorGapIds` + `supersedes` (whole-debate window), `ledgerClause` with
drift/time triggers, lineage-enforcement throw, degenerate-FAIL throw, per-role model routing
and doctrine comment; `agents/red-auditor.md`: "re-read the FULL living report in context" (the
audited object is blue's report). Accessed 2026-07-14.
