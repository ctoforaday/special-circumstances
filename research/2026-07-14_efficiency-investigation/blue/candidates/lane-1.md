# Lane 1 — adversarial-disconfirming-first: the five deferred levers, tested against their own evidence base

Run 4 (efficiency investigation), 2026-07-14. Method lens: hunt evidence AGAINST each frontier
hypothesis before evidence for it. Evidence base per `inputs/PINNED.md`: run-3 retrospective @
`bfa8a3b` (report §3, cost.md, friction.md, red/findings.md, debate.md, red/candidates/),
engine + backlog @ `5396952`. Hypothesis 1 taken first, then breadth. Winnow list honored.

**Lane verdicts at a glance** (each argued below, disconfirming evidence first):

| Lever | Lane verdict | One-line reason |
|---|---|---|
| 1. Severity-floor termination | **REJECT as specified** | The floor's trigger condition is never satisfied at any run-3 round boundary; the backlog's own "$10 at round 3" claim is contradicted by the pinned grades |
| 2. Risk-mass-proportional spend | **REJECT as auto-throttle; RATIFY as logged instrumentation** | Low-mass boards preceded the run's highest-value discoveries twice; mass measures known rework, not undiscovered defects |
| 3. Grade-dispute channel / best-of-N | **REJECT for run 4, named revisit trigger; best-of-N REJECT per the backlog's own condition** | PR #15's whole-debate docket already escalates persisting disputes; the residual case has zero observed instances |
| 4. Sharded findings + collator | **RATIFY sharding (conditions named); REJECT collator-as-seat, adopt prompt-level batching** | The full-re-read MUST binds red to blue's report, not red's own archive; the collator is dominated by a one-line prompt instruction |
| 5. Round-scoped audit | **HOLD for run 4 (as staged); conditionally RATIFY for run 5** — four-arm propagation-aware scope, gated on run 4's live propagation-clause evidence | Every run-3 full-re-read catch maps to a scope arm, but the mapping is post-hoc; the gate evidence arrives this run |

Doctrine constraint applied throughout: cheapen redundancy and mechanics, never judgment or the
adversary; the spot-check floor never reaches zero.

---

## 1. Severity-floor termination — REJECT as specified

### 1.1 The disconfirming find: the floor never fires on the only measured run

The backlog's claim: "when every open gap is <= MEDIUM with trivial fix cost, route the whole
board to the judge for disposition instead of another $25-30 round (would have ended run 3 at
round 3 for ~$10)."[^BacklogItem30]

Checked against red's own graded board at every run-3 round boundary (severities from the
findings file's preserved per-round blocks, read verbatim):[^FindingsBoard]

| Boundary (after red-merge round N) | Open board | Above-MEDIUM members | Floor fires? |
|---|---|---|---|
| Round 3 | R3-1..R3-10 | **R3-1 MEDIUM-HIGH, R3-2 MEDIUM-HIGH** (both complexity "low," not "trivial") | **No** |
| Round 4 | R4-1..R4-5 | **R4-1 HIGH** (certain × high) | **No** |
| Round 5 (actual termination) | R5-1..R5-6 | **R5-5 MEDIUM-HIGH** (medium × high) | **No** |

Three consequences, in increasing order of damage to the lever:

1. **The claimed saving is unsubstantiated by the lever's only cited evidence.** At the pinned
   grades, run 3 contains no round boundary where the floor's condition holds — the "$25-30/round
   saved" figure has no round it could have saved. The "would have ended run 3 at round 3" claim
   is contradicted by two MEDIUM-HIGH open gaps (R3-1, R3-2) sitting on the round-3
   board.[^FindingsBoard]
2. **A floor relaxed far enough to fire would terminate exactly the wrong rounds.** To fire at
   round 3 the threshold must admit MEDIUM-HIGH — but then rounds 4 and 5, which minted R4-1
   (HIGH, certain × high: the lineage-blind docket, the corpus's highest-graded engine
   finding)[^R4OneDetail] and R5-5 (MEDIUM-HIGH: unenforced supersedes)[^R5FiveDetail], never
   run. Both of those findings shipped as PR #15 mechanism (the `supersedes`/`closures`
   enforcement throw now live at `debate.js` lines 227–235).[^EngineLineage] Terminating the
   adversary's discovery rounds is a doctrine violation, not a mechanics saving.
3. **Run 3's actual termination violated the floor.** The run ended by human stop-and-resume with
   a reduced `maxRounds` (measured ~$0 via cache replay, cutting ~7 residual rounds)[^StopResume]
   — WITH an above-floor gap (R5-5, MEDIUM-HIGH) still open. A strict severity floor used as a
   termination condition would have demanded MORE rounds at that point, not fewer. The lever,
   applied to its own motivating run, either does nothing or extends the run.

### 1.2 Breadth: what a well-formed stopping rule looks like (and that this isn't one)

The stopping-rule literature in two adjacent fields agrees on the shape of a valid criterion, and
it is not an open-board severity floor:

- **Multi-agent debate:** measured accuracy saturates at ~2–5 rounds on QA-style tasks, and
  adaptive stopping works — but the published criterion is *stability of the discovery process*
  (distributional change below ε for **2 consecutive rounds**, a double confirmation), not a
  property of the residual list.[^AdaptiveStability][^DebateRounds] Run 3 never stabilized under
  that test: round 4 minted a HIGH, round 5 minted a MEDIUM-HIGH — a stability rule correctly
  refuses to stop, matching the record's own verdict (FAIL, UNVERIFIED, 6 open).
- **Software inspection:** the classical stopping decision is an estimate of *remaining defect
  content* (capture-recapture over inspector overlap), i.e. a forward-looking estimate of what
  has NOT been found — not a severity summary of what has.[^CaptureRecapture]

The severity floor is a backward-looking statistic standing in for a forward-looking decision.
Run 3's record demonstrates the divergence live: the round-3 board's severity profile said
"mostly medium trivia," and the next two rounds produced a HIGH and a MEDIUM-HIGH.

**Generalization caveat (kept honest):** run 3's topic was the engine itself, so lens passes
tracing `debate.js` could mint HIGH *engine* findings late; on an ordinary external topic the
late-round population may genuinely be textual trivia. One run, on a self-referential topic, is
thin evidence in both directions — which is itself an argument for instrumentation before
mechanism (§1.3).

### 1.3 Disposition

**REJECT severity-floor termination as specified.** Do not build the trigger. Instead:

- **Document the demonstrated cheap terminator:** human stop-and-resume with reduced `maxRounds`
  (cache replay, ~$0 measured).[^StopResume] It is termination-by-judgment (the human reads the
  board and decides), which is exactly where the doctrine says judgment belongs. One paragraph in
  the research-protocol skill / `debate.js` header comment; zero mechanism.
- **Instrument for a future stability criterion:** log per round (the script already holds
  `redEnv`) the new-gap count and severity profile — `log()` is available and costs nothing. Two
  consecutive rounds minting nothing above a threshold is the discovery-stability shape the
  literature validates;[^AdaptiveStability] whether FEOV ever automates it should be decided on
  runs 4–5 data, not on a run the rule never fits.

---

## 2. Risk-mass-proportional spend — REJECT as auto-throttle; RATIFY as instrumentation

### 2.1 The disconfirming find: mass fails its own correlation prediction at the two boundaries that matter

H2 predicts computed mass at each merge correlates with next-round realized value. Computed from
the pinned per-round boards (mapping disclosed: low=1, low-medium=1.5, medium=2, medium-high=2.5,
high=3, certain/realized=3.5; compound cells read verbatim from the findings
blocks):[^FindingsBoard]

| Board after round | Open gaps | sum(L × I) | Mean per gap | Next round minted |
|---|---|---|---|---|
| 1 | 20 | ~98 | ~4.9 | 11 gaps, top MEDIUM-HIGH |
| 2 | 11 | ~65 | ~5.9 | 10 gaps, top MEDIUM-HIGH |
| 3 | 10 | ~44 | ~4.4 | 5 gaps, **top HIGH (R4-1)** |
| 4 | 5 | ~30 | ~6.0 | 6 gaps, **top MEDIUM-HIGH (R5-5)** |
| 5 | 6 | ~31 | ~5.2 | (terminated) |

Two failures, one structural insight:

1. **Low mass did not predict low next-round value — twice, at exactly the boundaries a throttle
   would have acted on.** The two lowest-mass boards (rounds 3 and 4) preceded the run's
   highest-graded discovery (R4-1, certain × high)[^R4OneDetail] and a MEDIUM-HIGH
   (R5-5).[^R5FiveDetail]
2. **The metric cannot even discriminate the trivia it exists to detect.** Mean mass per gap
   stays ~5 across all five rounds because late-round textual nits carry `certain` likelihood by
   construction (a text defect, once found, is certain) — a certain × low nit scores ~3.5 against
   a medium × medium real risk's 4. The cost.md narrative "rounds 3–5 closed ~15 mostly-trivial
   gaps"[^CostAudit] is true of fix cost but invisible to sum(L × I).
3. **Why (structural):** risk mass measures the *known open rework* — backward-looking. A spend
   decision needs an estimate of the *undiscovered* gap population — forward-looking. Those are
   different quantities, and run 3 exhibits the divergence live. The inspection literature's
   estimator for the forward-looking quantity is capture-recapture over *inspector overlap* —
   and run 3 already produces exactly that data shape for free: R4-1 was found by 4 of 5 lenses
   independently, R5-1 by 3 of 5, R5-5 by 1 of 5.[^LensConvergence][^CaptureRecapture]

### 2.2 The doctrine test: the throttle's target is partly the adversary, not only redundancy

The frontier's disconfirming test (b) bites. The lens-count throttle assumes lens passes are
redundancy; the pinned corpus shows the redundancy is partial and the singletons are valuable:

- **R5-5 is a 1-of-5 singleton.** Lens 5 alone made the enforcement argument; lens 4's candidate
  shows it *checked the adjacent territory and explicitly held* ("Considered raising this... Not
  raised"), and lens 2 examined the detector mechanics without raising
  enforcement.[^R5FiveSingleton] A mass-scoped round 5 running 2 lenses has a substantial chance
  of losing the finding that became PR #15's enforcement throw.
- The retrospective's own §3 row 6 documents the same phenomenon on the blue side: "this run's
  own highest-value catch came from exactly one lane doing exactly one method" — which is why the
  shipped lane roster carries a 2-of-N redundancy floor.[^Row6Roster]
- Capture-recapture's own validity caveat: with too few inspectors "no model is sufficiently
  accurate and underestimation may be substantial"[^InspectorCount] — cutting lens count destroys
  the only forward-looking estimator the system could have.
- The engine already classifies lens passes as the adversary's leaf-node work, not mechanics: §3
  row 16b moved keeper-run lenses toward full strength for exactly this reason, and the
  `debate.js` header carries the tradeoff note.[^Row16b]

Grade-noise test (frontier (a)) resolved in the lever's favor, for honesty: the three run-3 grade
corrections moved computed mass ~0 (R2-1 corrected the evidence count but retained High by
argument; R3-7 narrowed mechanism, grade retained; R5-1 corrected an enumeration, no grade
change).[^GradeCorrections] The input is stable; noise is NOT the kill reason. §2.1–2.2 are.

One further structural objection the backlog's caveat only half-covers: under an auto-throttle,
**red's self-graded output drives red's own next-round budget** — an incentive loop (grade high
to keep lenses; or the system starves red exactly when red under-grades). The caveat "grades are
red's own estimates — the spot-check floor never reaches zero"[^BacklogItem30] names the floor
but not the loop; H3's dispute channel was the loop's only adversarial check, and it is deferred.

### 2.3 Disposition

**REJECT risk-mass-proportional spend as an automatic controller of lens count or audit scope.**
**RATIFY the instrumentation half:**

- Compute and `log()` sum(L × I) per merge — the grades are already machine-readable in
  `redEnv` (compound enums shipped in PR #15);[^EngineLineage] one arithmetic line, zero new
  seats. Emit it into `cost.md` per round.
- Record **lens overlap** per merged gap (red-merge already states convergence in prose —
  "four of five lenses converged"[^R4OneDetail]; make it a per-gap envelope field, e.g.
  `found_by: ['L1','L2','L4']`). This is the capture-recapture input, mechanical to collect,
  and it converts the "was that round redundant?" question from narrative to data.
- Revisit an actuated throttle only when runs 4–5's logged record shows mass (or a
  remaining-defect estimate) actually predicting next-round value — the same
  evidence-before-mechanism condition the backlog itself imposes on best-of-N
  grading.[^BacklogItem30]

This is the doctrine-safe split: the costly thing (turns × context at the judgment seats) is
attacked by lever 4, which cuts mechanics; lever 2's automatic form cuts adversarial coverage on
the strength of a metric that failed its own correlation test on the pinned record.

---

## 3. Grade-dispute channel and best-of-N grading — REJECT for run 4, with a named revisit trigger

### 3.1 The disconfirming find: the frontier's premise is half-obsoleted at the pinned engine

The frontier's case: disputes rode the general gap loop "invisible to the docket detector, and a
red re-rejection could persist without judge escalation (run 3: judge dispatched ZERO
times)."[^FrontierH3] True of the engine run 3 ran on. **No longer true at `5396952`:** the
shipped detector arms on any gap whose id appears in ANY prior round (whole-debate window) or
whose `supersedes` chain descends from one — so a dispute red re-rejects keeps its id on the
board and auto-dockets to the judge the following round.[^EngineLineage] The zero-LEAD record
(verified: `grep -n "^### " debate.md` = 6 BLUE / 5 RED / zero LEAD headers)[^DebateNoLead] was
the R4-1 defect; PR #15 is its shipped fix, and the winnow list bars re-recommending it.

What the shipped mechanism still cannot see — the honest residual:

1. **Fix-but-dispute-grade.** Blue repairs the gap (cheaper than arguing) while disputing the
   grade; the gap closes, the id leaves the board, and the disagreement leaves no persistence
   signal. Observed instances in run 3: **zero.** Blue's round-5 position states it checked and
   found none: "every gap was real, at the location red found it, and none was over-graded
   relative to its fix cost."[^BlueRound5] Every run-3 grade disagreement resolved inside the
   loop in ≤ 2 rounds by evidence (R2-4: rebuttal accepted with evidence; item-15's likelihood:
   retained High by argument, not assertion; R2-9/R2-10: argued risk-accepts red accepted on
   stated conditions).[^GradeCorrections][^DebateNegotiation]
2. **Resolution vocabulary.** If a grade dispute does reach the judge via the persistence path,
   the judge's enum (`closed | rebuttal_sustained | risk_accepted | carried | unresolved`) cannot
   express "gap real, grade wrong."[^EngineLineage] A one-value enum addition, IF the case ever
   arises.

### 3.2 Best-of-N grading: the backlog's own condition is unmet, and the panel literature cuts against FEOV's configuration

The backlog defers best-of-N "only if runs 4–5 show lone-voice bias survives."[^BacklogItem30]
Searched for a surviving-bias instance in the pinned corpus per the frontier's disconfirming
test: none found — every identified grade error was caught and corrected within the adversarial
loop (blue caught red's transposed count R2-1; red caught its own stale framing R3-7; red's
merge overruled a lens's wrong "no discrepancy" hold at R5-1 and logged the error against
itself).[^GradeCorrections][^R5OneOverrule] The one structural blind spot is that **final-round
grades are never adversarially audited** (the run ends; R5-1..R5-6's grades shipped un-reviewed)
— but the cheap fix for that is one more blue response or a stop-resume, not a standing panel.

External evidence, weighed both ways per the lens: diverse judge panels beat single judges at
lower cost (PoLL: heterogeneous small-model juries, ~7x cheaper)[^PoLLJuries] — but the effect
rests on *cross-provider* diversity, and the correlated-errors result shows same-family panels
collapse to a few effective votes ("Nine Judges, Two Effective Votes").[^CorrelatedJudges] FEOV
grading runs single-provider by construction, i.e. in exactly the regime where a panel buys
least. And FEOV's grades are not one-shot scalar judgments — they are evidence-anchored,
multi-round, adversarially contested records; the panel literature's single-shot setting is the
weaker analogue.

### 3.3 Disposition

**REJECT both halves for run 4.** No exhibited need (zero instances of the residual case), the
persistence path is already covered by shipped mechanism, and grades currently control nothing
but the record. **Named revisit trigger (binding, not decorative):** the moment any mechanism
makes red's grades load-bearing for spend or termination (an actuated lever 1 or 2), the
grade-dispute channel MUST ship with it — a spend-controlling input without an adversarial
correction path is the incentive loop §2.2 names. The two levers are a package deal in that
future; rejecting lever 2's actuation this run is what makes rejecting lever 3 safe.

---

## 4. Sharded findings + collator — RATIFY sharding with named conditions; REJECT the collator seat

### 4.1 The doctrinal question is answerable from the text: no conflict exists

The full-re-read MUST, quoted from the pinned agent contract: "BEFORE auditing, YOU MUST re-read
the FULL living report in context — a change-summary is a navigation hint, never the audit
surface."[^RedAuditorMust] The object is **blue's living report** — the audit surface. Red's own
`findings.md` is not the audit surface; it is red's ledger of its own past work. Sharding it
(open-items ledger vs closed archive; merge reads open + this round; archive readable on demand)
narrows no audit read the doctrine names. The citation ledger is the shipped precedent for the
identical closed-items-don't-reopen pattern, and it held every prior confidence through round 4
with zero observed regressions.[^LedgerPrecedent]

### 4.2 The cost and quality case, measured

- Merge seats are the priciest recurring line: red-merge $7.52/$13.22/$12.64/$10.60/$13.56
  across rounds 1–5 = $57.54 of the run's $149.95 (38%), rate-driven at the judgment tier (5×
  cache-read, 12.5× cache-write).[^CostAudit]
- The measured driver is TURNS × CONTEXT: an agent re-reads its whole context every tool call;
  red-merge-r1 alone held ~100–150K of material across 2.7M+ cache reads.[^Backlog28d]
- The living files outgrew the Read tool: `red/findings.md` ended at 106,772 bytes and
  `blue/report.md` at 159,394 (both multiples of the 25k-token Read cap); round 5's merge needed
  three windowed reads plus greps for the full-re-read mandate — friction #15's exact
  complaint.[^FrictionFifteen]
- **The quality argument runs the same direction as the cost argument:** long-context research
  shows LLM performance degrades substantially (13.9%–85% across models/tasks) as input grows,
  *even when retrieval of the relevant span is perfect*.[^ContextLength] A merge seat holding a
  100K+ resident context that is mostly its own closed history is operating in the measured
  degradation regime. Shrinking the resident archive is judgment-STRENGTHENING, not only
  cheaper — the rare lever the doctrine favors twice over.

Honest sizing: the archive's share of merge context is maybe 20–30K tokens by round 5; at
judgment-tier cache-read rates across ~60 turns that is roughly $1–3/round of the $10–13, i.e.
~$5–15 at run-3 scale — bounded, but it grows monotonically with rounds (the archive only ever
grows) and the turn reduction from fewer windowed reads compounds it. No measured decomposition
of context composition exists; run 4 should measure before anyone quotes a bigger number.

### 4.3 The frontier's disconfirming tests, resolved

**(a) R5-1 — does the catch need the closed archive in-context?** The catch's trigger (a lineage
enumeration in blue's §3 row 23) sits in blue's report — the always-read audit surface, untouched
by sharding. The verification (checking each chain link against red's own closure entries, some
closed 2–3 rounds prior)[^R5OneDetail] is a leaf-node fetch into the archive — exactly the
protocol's mode-3 access, which "readable on demand" preserves. The catch survives sharding
**provided the demand is disciplined, not discretionary** — hence condition (ii) below. Note the
contrast with lever 5's hard case: R5-1's trigger is *visible* in the always-read surface; R4-4's
trigger (a stale numeral in an unchanged paragraph) is invisible — that asymmetry is why this
lane ratifies lever 4 and holds lever 5.

**(b) Collator nuance loss** — resolved against the collator; see §4.5.

**(c) Write-block on the new shard files** — resolved favorably: run 3's accidental control
condition isolated the guard as filename-keyed and path-independent (`findings.md` refused even
in a scratchpad; a neutral filename succeeded).[^FrictionFour] Neutral shard names plus PR #14's
pre-created skeleton (Edit/append path) plus the citation-ledger precedent (appended via `cat`
across four rounds, zero incidents)[^FrictionTen] cover it. The undercounted-cost-cell concern
does not materialize under naming discipline.

### 4.4 Ratification conditions (the shape that keeps the catch classes)

1. **Single source of truth for status:** closure status lives ONLY in the open-ledger (items
   move ledger → archive on closure; archive blocks are immutable). R5-1's failure class —
   status lines contradicting each other across sections of one growing file — gets one
   authoritative site instead of five preserved-verbatim historical blocks.[^R5OneDetail]
2. **A named MUST for archive verification:** any lineage/closure claim (in blue's text or red's
   own docket) is verified against the archive by targeted read — grep-cheap, mode-3 discipline,
   written into the red-auditor contract so R5-1-class catches are demanded reads, not
   discretionary ones.
3. **Skeleton + neutral filenames** per §4.3(c); shard files pre-created like the citation
   ledger.
4. **The archive stays git-tracked and readable on demand** — nothing is summarized away; this is
   a read-default change, not a retention change (blue-additive doctrine intact).

### 4.5 The collator: REJECT the seat — it is dominated, and its failure class is documented

Three independent kills, any one sufficient:

1. **Dominated alternative at zero mechanism:** the script cannot concatenate (it has no
   filesystem access by design — `debate.js`'s own comment),[^EngineLineage] but the merge agent
   can: one added prompt sentence — "first `cat` the round's lens candidates into a single
   scratch file and read that" — captures the entire one-big-read/fewer-turns benefit with no new
   seat, no new dispatch cost, no handoff. This is the backlog's own lever (3), "prompt-level
   read batching,"[^Backlog28d] and it strictly dominates a collator seat whose dispatch overhead
   is the same order as its plausible saving.
2. **A normalizing collator is the documented failure class, already demonstrated in-corpus:**
   run 3's envelope enum rounded red's compound grades every round — "every compound grade above
   was rounded; the authoritative grading lives in red/findings.md" (friction #6)[^FrictionSix] —
   a live instance of normalization destroying grading nuance, fixed in PR #15 by widening the
   enum. The multi-agent handoff literature says the same thing generally: intermediate
   compression drops edge-case detail and introduces paraphrase errors that compound
   downstream.[^HandoffLoss]
3. **Clustering is judgment, not mechanics:** deciding two lens findings are "the same gap" sets
   the convergence count, and convergence counts fed corroboration grades in run 3 ("four of
   five lenses converged independently" is part of R4-1's HIGH corroboration; "three lenses
   converged" part of R5-1's).[^R4OneDetail][^R5OneDetail] A bulk-tier collator pre-deciding
   convergence cheapens the adversary's evidentiary input — the doctrine's named prohibition.
   (PR #15's lens-scoped labels already removed the renumbering mechanics that was the collator's
   honest half.)[^AlreadyShipped]

---

## 5. Round-scoped audit — HOLD for run 4; conditionally RATIFY for run 5, propagation-aware, gated on run-4 evidence

### 5.1 The lens applied to the frontier's own rejection-lean: the catch record maps better than expected

The frontier predicts round-scoping is structurally blind to the unpropagated-site class. Hunting
against that (this lane's assigned direction), each late-round catch red actually made by full
re-read was traced to whether a four-arm scope would have surfaced it:

| Catch | What found it (per the pinned record) | Scope arm that covers it |
|---|---|---|
| R4-3 (unedited ambiguous sentence) | Sat in the SAME CELL as R3-5's contested fix; lenses 2+4 read the cell | Contested-lineage arm, section-granular (audit the whole section containing any fix, not the diff hunk) |
| R4-4 (fifth stale "4th occurrence" numeral, in a paragraph unchanged since round 1) | "report-wide grep '4th\|fourth' at merge: exactly one uncorrected instance"[^R4FourGrep] | **Propagation-grep arm — the catch was ALREADY made by grep, not by linear re-read**; mechanical, whole-file, cache-cheap |
| R3-6 (zero "independent" hits) | Repo-wide grep | Propagation-grep arm |
| R3-10 (untagged cost figures) | Direct read of both instances | Propagation-grep arm (grep the corrected figure) |
| R5-1 (discarded enumeration in row 23) | Row 23 was WRITTEN in round 4 — a changed section at the round-5 audit | Changed-sections arm |
| R1-1 (stale §0 headline) | Round-1 full audit | Round 1 is always unscoped |

The load-bearing observation: **the poster-child full-re-read catch (R4-4) was not made by
reading — it was made by grep.** The propagation-aware expansion the frontier names as the
ratification condition ("for every correction accepted this round, ALL sites stating the
corrected claim enter the audit surface — a grep-cheap mechanical expansion") is not a
hypothetical mitigation; it is a description of how the catch actually happened at the merge
seat. External corroboration runs the same way: diff-only review reliably catches local issues
and misses global-invariant violations; the field's answer is hybrid scoping (diff + selective
context), not full re-reads forever.[^DiffReview]

### 5.2 What still gates ratification

1. **The mapping above is post-hoc.** The four arms (changed sections ∪ contested lineages ∪
   propagation-grep expansion ∪ nonzero random spot-check including unchanged sections) were
   drawn looking at run 3's catch classes; a novel regression class in a future run may evade
   all four. That residual is exactly what the spot-check floor exists for, and the doctrine
   pins it above zero — but the floor's catch probability is unmeasured. Grade: the rule is
   plausible, not proven.
2. **PR #15's blue-side propagation clause has zero live evidence.** It shipped 2026-07-14 and
   run 4 (this run) is its first exercise.[^AlreadyShipped] The frontier's condition — one live
   run of evidence before red's backstop is narrowed — is satisfiable only after this run's
   record exists. If run 4's own record shows the clause failing (the winnow list's audit
   trigger), round-scoping must be rejected outright for run 5: it would remove the only check
   that catches the engine's measured dominant regression class (5 propagation chains in 5
   rounds).[^PropagationChains]
3. **Unlike lever 4, this lever narrows the audit surface itself.** The full-re-read MUST names
   blue's report;[^RedAuditorMust] round-scoping reads less of it. The context-degradation
   evidence (§4.2) cuts FOR scoping here too — a lens re-reading a 159KB report every round is
   also in the degradation regime — but the protocol ranks full-read-of-the-audit-surface above
   token savings explicitly ("this clause outranks any token saving"). Overriding a named
   doctrine clause needs the run-4 evidence, not an inference from run 3.

### 5.3 Disposition

**HOLD for run 4** (already staged that way; this run audits at full re-read). **For run 5:
RATIFY conditionally** — rounds 2+ audit scope = changed sections ∪ contested lineages
(section-granular) ∪ propagation-grep expansion of every correction accepted that round ∪ a
nonzero random spot-check floor that includes unchanged sections — **contingent on run 4's
propagation-clause record showing zero unpropagated-site regressions.** Any unpropagated site in
run 4 = reject for run 5 and re-dock the lever. The decision point and its evidence are one run
away; nothing is gained by deciding early on thinner evidence.

---

## 6. Cross-cutting: the shape of the ratified set

**What this lane ratifies spends nothing on judgment and nothing on the adversary:**

| Ratified | Class | Cost |
|---|---|---|
| Sharded findings (conditions §4.4) | Mechanics (red's own archive residency) | File discipline + skeleton entries + one MUST |
| Prompt-level read batching at the merge (collator's dominated replacement) | Mechanics (turn count) | One prompt sentence |
| Risk-mass + new-gap-severity logging per round | Instrumentation | One arithmetic `log()` line |
| Lens-overlap (`found_by`) per merged gap | Instrumentation (capture-recapture input) | One envelope field |
| Stop-and-resume documented as standing termination practice | Documentation | One paragraph |
| Round-scoped audit, run 5, four-arm | Mechanics — **gated on run-4 evidence** | Prompt/agent-contract edits |

**What it rejects, and the single sentence each rejection stands on:**

- Severity-floor termination: its trigger never fires on the only measured run, and every
  relaxation that fires terminates the rounds that produced the corpus's highest-graded
  findings.
- Risk-mass auto-throttle: low-mass boards preceded high-value discoveries twice; the metric
  measures the wrong quantity (known rework, not undiscovered defects) and its actuated form
  cuts adversarial coverage.
- Grade-dispute channel now: the persistence case is covered by shipped mechanism and the
  residual case has zero observed instances — revisit trigger bound to any future grade-actuated
  spend.
- Best-of-N grading: the backlog's own precondition (surviving lone-voice bias) is unmet in the
  pinned corpus, and the panel literature's benefit concentrates in the cross-provider
  configuration FEOV doesn't run.
- Collator seat: dominated by a one-sentence prompt instruction; its normalization half is the
  in-corpus-demonstrated nuance-loss class; its clustering half is judgment wearing a mechanics
  costume.

**The interlocks (stated so the synthesis can't lose them):** lever 3 is a mandatory companion
of any future lever-1/2 actuation (spend-controlling grades need an adversarial correction
path); lever 4's batching sentence replaces the collator; lever 5's run-5 disposition is decided
by run 4's propagation record, mechanically, per the winnow list's audit trigger.

## Confidence self-grades

- Lever 1 reject: **HIGH** — the floor-never-fires table is mechanical extraction from pinned
  grades; the doctrine argument follows from the record. Residual: one run, self-referential
  topic (§1.2 caveat).
- Lever 2 reject-throttle/ratify-instrumentation: **HIGH** on the correlation failure (two
  boundary counterexamples, pinned); **MEDIUM** on the capture-recapture alternative (external
  precedent strong, in-corpus validation pending runs 4–5).
- Lever 3 reject-with-trigger: **MEDIUM-HIGH** — the zero-instances claim rests on run 3's
  transcript read to saturation, but absence-of-instance is weaker than presence; the
  half-obsoleted-premise claim is HIGH (code read at pin).
- Lever 4 ratify-sharding: **MEDIUM-HIGH** — doctrine-scope argument is textual and HIGH; the
  dollar sizing is estimated, not measured (named as such). Collator reject: **HIGH** (dominance
  argument + in-corpus demonstration).
- Lever 5 hold-then-conditional: **MEDIUM** — the catch-to-arm mapping is post-hoc by
  construction; the gate evidence does not exist yet. That is why the verdict is a gate, not a
  ratification.

## Open questions carried

1. Does run 4's blue propagation clause hold at zero unpropagated sites? (Decides lever 5's
   run-5 disposition — the record is being generated by this run.)
2. Can red-merge attribute lens overlap honestly (the `found_by` field) without inflating
   convergence — and does the resulting capture-recapture estimate track next-round discovery on
   runs 4–5? (Decides whether any future spend throttle has a valid input.)
3. What fraction of merge-seat resident context is actually the closed archive? (The $1–3/round
   sizing in §4.2 is an estimate; run 4's transcripts can measure it before run 5 quotes it.)
4. If a grade dispute ever reaches the judge via the persistence path, the resolution enum
   cannot express "gap real, grade wrong" — does that case occur in runs 4–5?
5. Generalization: does an ordinary (non-self-referential) topic show the same late-round
   high-severity discovery pattern, or was run 3's tail an artifact of auditing the engine with
   the engine?

## Footnotes

[^BacklogItem30]: "frank-exchange-of-views — run-3 termination & fairness levers," `ideas/backlog.md` item 30 @ pinned `5396952` — severity-floor text ("would have ended run 3 at round 3 for ~$10"), risk-mass umbrella + caveat, grade-dispute channel + best-of-N condition. Accessed 2026-07-14.
[^FindingsBoard]: `research/2026-07-12_feov-retrospective/red/findings.md` @ `bfa8a3b` — per-round gap blocks read verbatim: round-3 board headers (lines ~717–893: R3-1 MEDIUM-HIGH low-medium×medium-high×low; R3-2 MEDIUM-HIGH certain×medium×low; R3-3..R3-10), round-4 originals (lines ~425–531: R4-1 HIGH certain×high), round-5 block (lines ~135–253: R5-5 MEDIUM-HIGH medium×high), round-1 originals (lines ~1281–1353), round-2 originals (lines ~1080–1188). Accessed 2026-07-14.
[^CostAudit]: `research/2026-07-12_feov-retrospective/cost.md` @ `bfa8a3b` — per-seat-round table (red-merge $7.52/$13.22/$12.64/$10.60/$13.56; total $149.95); findings: cache traffic 99% of tokens; judgment-tier 5×/12.5× rates; "rounds 1–2 closed 31 gaps ($60-ish); rounds 3–5 closed ~15 mostly-trivial gaps for a similar spend." Accessed 2026-07-14.
[^StopResume]: `cost.md` @ `bfa8a3b`, final finding: "Stop-and-resume with a reduced maxRounds (cache replay) cost ~$0 and cut ~7 residual rounds; five round-6 lenses were killed mid-spawn for pennies." Accessed 2026-07-14.
[^R4OneDetail]: `red/findings.md` @ `bfa8a3b`, R4-1 original grading (line ~425): "HIGH — certain (already realized in this corpus, not projected) × high × low-medium — corroboration: HIGH (... four of five lenses converged independently, each tracing the code first-hand)"; lens convergence re-verified this lane: 4 of 5 round-4 candidate files contain the lineage/contested analysis. Accessed 2026-07-14.
[^R5FiveDetail]: `red/findings.md` @ `bfa8a3b`, R5-5 header (line ~200): "MEDIUM-HIGH — medium × high (telemetry-invisible: an unset or vacuous supersedes leaves contested.length at 0...)". Accessed 2026-07-14.
[^R5FiveSingleton]: `red/candidates/round-5-lens-5.md` @ `bfa8a3b` (lines ~36–103, the full enforcement argument) vs `round-5-lens-4.md` (lines ~110–120: "checked whether 'WITH REGRESSION' is a documented protocol state... Considered raising this... Not raised") and `round-5-lens-2.md` (detector-logic analysis, no enforcement claim). Accessed 2026-07-14.
[^R5OneDetail]: `red/findings.md` @ `bfa8a3b`, R5-1 (line ~135): "certain (static text, read side by side at the merge seat) × medium... three lenses converged independently — lenses 1, 2, 4... every chain link checked against this file's own closure entries." Accessed 2026-07-14.
[^R5OneOverrule]: `debate.md` @ `bfa8a3b`, round-5 RED (line ~738): "one lens's contrary 'no discrepancy' hold was overruled at the merge seat by direct read of report lines 496/727 — logged against red below." Accessed 2026-07-14.
[^EngineLineage]: `plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js` @ pinned `5396952` — read in full this lane: RED_ENVELOPE `supersedes`/`closures` (lines 89–107), compound GRADE enum (line 60), lineage-enforcement throw (lines 227–235), whole-debate contested window (lines 186, 238–246, 258), judge dispatch + resolution enum (lines 125–144, 247–257), no-filesystem-by-design comment (lines 32–34), doctrine comment (lines 24–31). Accessed 2026-07-14.
[^RedAuditorMust]: `plugins/frank-exchange-of-views/agents/red-auditor.md` @ `5396952`, line 13: "BEFORE auditing, YOU MUST re-read the FULL living report in context — a change-summary is a navigation hint, never the audit surface." Accessed 2026-07-14.
[^Backlog28d]: `ideas/backlog.md` item 28(d) @ `5396952` — "the driver is TURNS x CONTEXT... red-merge-r1: ~100-150K of material, 2.7M+ cache reads"; levers (1) shard, (2) collator, (3) prompt-level read batching. Accessed 2026-07-14.
[^FrictionFifteen]: `research/2026-07-12_feov-retrospective/friction.md` @ `bfa8a3b`, entry 15 (red-merge-r5): Read-tool 25k cap vs the living report forced "three windowed reads plus targeted greps." File sizes measured this lane at `bfa8a3b` working tree: findings.md 106,772 bytes; blue/report.md 159,394 bytes. Accessed 2026-07-14.
[^FrictionSix]: `friction.md` @ `bfa8a3b`, entry 6 (red-merge-r3): "gap grades are forced into low/medium/high enums... every compound grade above was rounded; the authoritative grading lives in red/findings.md." Accessed 2026-07-14.
[^FrictionFour]: `friction.md` @ `bfa8a3b`, entry 4 (red-merge-r2): the accidental control condition — `findings.md` refused at a scratchpad path, neutral filename succeeded; "the guard is filename-keyed regardless of directory." Accessed 2026-07-14.
[^FrictionTen]: `friction.md` @ `bfa8a3b`, entry 10 (red-merge-r4): "debate.md and the ledger appended via cat" — four rounds of citation-ledger appends without a write-block incident. Accessed 2026-07-14.
[^LedgerPrecedent]: `friction.md` entry 11 @ `bfa8a3b` ("ledger skip-rule held all prior-round confidences") and PR #14's citation-ledger entry per `inputs/already-shipped.md`. Accessed 2026-07-14.
[^GradeCorrections]: `report.md` §3 rows 15 (R1-13→R2-1→R3-7: count corrected 3→2, likelihood retained High by argument; mechanism narrowed R3-7, grade kept as argued risk-accept) and 23 (R5-1: enumeration corrected, grades untouched) @ `bfa8a3b`. Accessed 2026-07-14.
[^Row6Roster]: `report.md` §3 row 6 @ `bfa8a3b`: "this run's own highest-value catch (false-premise repo verification) came from exactly one lane doing exactly one method"; shipped roster + 2-of-N floor per `debate.js` LANE_METHODS @ `5396952`. Accessed 2026-07-14.
[^Row16b]: `report.md` §3 row 16b @ `bfa8a3b` (lens passes are the leaf-node audit work; keeper runs omit `model` so the adversary runs at full strength) and the `debate.js` header's KNOWN TRADEOFF note @ `5396952`. Accessed 2026-07-14.
[^FrontierH3]: `blue/frontier.md` (this run), H3. Accessed 2026-07-14.
[^DebateNoLead]: `research/2026-07-12_feov-retrospective/debate.md` @ `bfa8a3b` — re-verified this lane: `grep -n "^### "` returns 6 BLUE + 5 RED headers, zero LEAD. (An unanchored grep returns 5 prose mentions of "### LEAD" — friction #12's count-mode footgun class, encountered live again this lane.) Accessed 2026-07-14.
[^BlueRound5]: `debate.md` @ `bfa8a3b`, round-4 BLUE closing (lines ~705–712): "No rebuttals this round — every gap was real, at the location red found it, and none was over-graded relative to its fix cost." Accessed 2026-07-14.
[^DebateNegotiation]: `debate.md` @ `bfa8a3b`, round-2 RED (lines ~287–289): "Red will accept argued risk-accepts on R2-9 and R2-10 if..." — grade/disposition negotiation resolving in-loop. Accessed 2026-07-14.
[^PropagationChains]: `report.md` §3 row 23 corrected enumeration + §2.1(b) @ `bfa8a3b`: R1-5→R2-4→R3-4/R3-9; R2-5→R3-10; R2-7→R3-6; R2-8→R3-5→R4-3 — plus R4-4's unpropagated fifth numeral; "5 chains in 5 rounds" per this run's task statement and blue-researcher contract. Accessed 2026-07-14.
[^R4FourGrep]: `red/findings.md` @ `bfa8a3b`, R4-4 corroboration: "report-wide grep '4th|fourth' at merge: exactly one uncorrected instance, §3 risk-accepted paragraph." Accessed 2026-07-14.
[^AlreadyShipped]: `inputs/already-shipped.md` (this run dir) — PR #15 entries: lineage docket + enforcement throw, compound grades via envelope, lens-scoped labels, blue propagation clause (merged 2026-07-14; run 4 is its first live exercise). Accessed 2026-07-14.
[^AdaptiveStability]: "Multi-Agent Debate for LLM Judges with Adaptive Stability Detection," arXiv:2510.12697 — stopping when the round-over-round KS statistic stays < 0.05 for 2 consecutive rounds; adaptive stops at rounds 2–8 lose <1% accuracy vs fixed 10 rounds. https://arxiv.org/html/2510.12697v1. Accessed 2026-07-14.
[^DebateRounds]: "Literature Review of Multi-Agent Debate for Problem-Solving," arXiv:2506.00066 (and survey results: saturation ~2–5 rounds, task-dependent; degradation documented past round 2 on some tasks). https://arxiv.org/html/2506.00066v1. Accessed 2026-07-14. Volatility: preprint literature, findings are QA-task-shaped, weaker analogue to audit loops.
[^CaptureRecapture]: "A Comprehensive Evaluation of Capture-Recapture Models for Estimating Software Defect Content," IEEE TSE (Briand et al.) — inspector-overlap estimation of remaining defects as the post-inspection stop/reinspect decision. https://ieeexplore.ieee.org/document/852741/. Accessed 2026-07-14.
[^InspectorCount]: Same evaluation literature: "when the number of inspectors is too small, no model is sufficiently accurate and underestimation may be substantial" (models strongly affected by inspector count). https://ieeexplore.ieee.org/document/852741/. Accessed 2026-07-14.
[^PoLLJuries]: "Replacing Judges with Juries" (PoLL, Cohere) — diverse panel of smaller heterogeneous models outperforms a single frontier judge at ~7x lower cost. Summarized via https://www.comet.com/site/blog/llm-juries-for-evaluation/ and https://medium.com/@techsachin/replacing-judges-with-juries-llm-generation-evaluations-with-panel-of-llm-evaluators-d1e77dfb521e. Accessed 2026-07-14. Volatility: secondary summaries; primary is arXiv:2404.18796 (not re-fetched this lane).
[^CorrelatedJudges]: "Nine Judges, Two Effective Votes: Correlated Errors Undermine LLM Evaluation Panels," arXiv:2605.29800 — same-family panels collapse to few effective votes; panel benefit concentrates in cross-provider diversity. https://arxiv.org/pdf/2605.29800. Accessed 2026-07-14.
[^ContextLength]: "Context Length Alone Hurts LLM Performance Despite Perfect Retrieval," EMNLP 2025 Findings / arXiv:2510.05381 — 13.9%–85% degradation with input length even at perfect retrieval, persisting when irrelevant tokens are whitespace or masked. https://arxiv.org/pdf/2510.05381. Accessed 2026-07-14.
[^HandoffLoss]: Multi-agent handoff/compression failure literature: lossy intermediate summaries drop edge-case detail and introduce paraphrase errors that compound downstream (e.g., https://www.zartis.com/the-compounding-errors-problem-why-multi-agent-systems-fail-and-the-architecture-that-fixes-it/; https://galileo.ai/blog/why-multi-agent-systems-fail). Accessed 2026-07-14. Volatility: practitioner sources; used for the failure-class shape, not figures — the in-corpus friction #6 instance is the load-bearing evidence.
[^DiffReview]: Diff-vs-context code-review evidence: diff-only review catches local issues, misses global-invariant/architectural defects; hybrid diff+selective-context is the effective pattern (https://graphite.com/guides/ai-code-review-context-full-repo-vs-diff; https://www.coderabbit.ai/guides/code-context; "Towards Practical Defect-Focused Automated Code Review," arXiv:2505.17928). Accessed 2026-07-14.
