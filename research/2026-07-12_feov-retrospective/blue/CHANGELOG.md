# blue CHANGELOG — FEOV retrospective

## Round 0 — synthesis by union (2026-07-13)

Merged `candidates/lane-1.md` (H1-deep, 6 sections + proposal table), `lane-2.md` (H2-deep,
7 sections + 14-case simulator suite), `lane-3.md` (H3-deep, 8 sections + 13-item graded table)
into `blue/report.md`. Structural merge, no substantive content dropped. Concrete operations:

- RESOLVED LANE CONFLICT (new §0): lane 3 ("claimed fixes do not exist in source") vs lane 1
  ("fixes exist on unmerged PR #14"). Synthesizer leaf-node verification (git branch/ls-tree/show
  on both refs, new footnote [^MainVsBranch]): both true — guards/simulator/skeleton/catechism
  live on `feat/feov-dogfood-round-1`, absent from `main` @ 9ff0fad. Union headline: "a shipping
  question, not a research question" (L1) + "a backlog checkbox is not a diff" (L3).
- CORRECTED a lane-2 figure: its footnote attributed 2,972 lines to run-2 `blue/report.md`;
  wc -l shows blue/report.md = 2,145, assembled report.md = 2,972. Grep conclusions unaffected;
  correction noted inline in §1.2.
- §1.1 (H1): merged lane 1's convergence/erasure measurements + disconfirming literature
  (diversity collapse, wisdom-of-crowds, isolated-correction), lane 2's half-refutation of the
  frontier's CHANGELOG clause + convergence-as-replication + method-vs-headcount analysis +
  diminishing-returns caution, lane 3's structural-overlap and lens-assignment evidence
  (arXiv 2602.03794) and line-104 source reading. Kept all three run-2 content inventories
  (consensus / lane-1-only / lane-2-only), deduplicated.
- §1.2 (H2): merged all three lanes' grep measurements (deduped to one set), lane 2's
  negative-provenance exception + R2-10 distinction + cheap-half complexity check, lane 3's
  CHANGELOG-partial-attribution nuance + provenance-literature precedent, lane 1's
  pragmatist-grade.
- §1.3/§2 (H3): adopted the trimodal taxonomy all three lanes independently derived. §2.1 defect
  table = union of lane 3's 10-row table + lane 2's per-defect items (empty-candidates cascade,
  gap-id rollover, edge counts, malformed envelope) + lane 1's PR-14 classifications. §2.2 design
  merged lane 2's mechanism sketch + lane 3's two implementation paths, with lane 3's flagged
  open question (injection contract) RESOLVED by lane 1's PR-14 harness evidence (AsyncFunction).
  §2.3 = PR #14's existing 11 tests (lane 1) + 12 additions deduplicated from lane 2's 14-case
  and lane 3's 13-case lists.
- §3 (H4): merged three graded tables into 18 rows; reconciliations recorded in-row: write-block
  layered fix (skeleton primary [L1] / Bash-append belt-and-braces [L3] / rename risk-accepted
  [L2+L3]); advisory access risk-accept (L2's engineered-around evidence adopted over L1's
  fix-if-feasible); PDF extraction re-graded to low complexity via MCP adoption (L2 supersedes
  L1's grade); per-role models "defer" (L3) superseded by already-built (L1). Consolidated
  explicit risk-accepted list.
- §4 (H5): merged three friction rankings; preserved the three lanes' differing role-counts for
  write-block and advisory rounds as an instrumentation finding rather than silently picking one.
  Kept lane 2's round-4-asymmetry insight (engineered-around vs no-workaround) as the ranking
  rationale, lane 3's "already fixed is false" refutation of the frontier parenthetical, lane 1's
  four-item top tier correction.
- Kept lane 3's un-frontiered-doubts section (§1.4: duplication refuted, open_questions confirmed
  dropped, Heilmeier needs instrumentation) + red-memory-loop positive finding; added lane 1's
  PR-14 catechism evidence to the Heilmeier item.
- NEW §5: consolidated open-questions register (6 items) — practicing the very open_questions
  fix §1.4 confirms is missing from the engine's template.
- Practiced this run's own H2 finding: every claim tagged [L1]/[L2]/[L3]/[all lanes].
- LIVE EVIDENCE ADDENDUM (§0, §2.1, §3 #8, §4 rows 4–5): the Write of this very report was
  refused by the write-block (third occurrence, third run), and the first chunked-heredoc
  workaround failed on shell parsing; produced via scratchpad-Write + Bash copy. Logged as
  friction in the envelope.
- Footnotes: union of 40 lane footnotes deduplicated to 36 semantic labels (merged
  ResearchCommand/ResearchCmd, WorkflowJs/Workflow, Run1Journal/Journal,
  Enametoolong/DriftFriction into Run2Friction; renamed run-2 candidate refs to
  Run2Lane1/Run2Lane2 to avoid collision with this run's lanes); added [^MainVsBranch].

## Round 1 — address red's 20 gaps (R1-1..R1-20), all additive (2026-07-13)

Live re-verification this round: `gh pr view 14` (state MERGED, `00018a5`, 2026-07-14T05:58:54Z,
diffstat +318/-48/18 files/11 commits), `git log --oneline origin/main` (HEAD now `47ae48d`),
direct full read of `debate.js` on `main` @ `47ae48d`, `wc -l`/entry-count on `run2-friction.md`
(21 entries), live WebFetch/WebSearch re-checks of four cited papers plus the WisdomCrowds PDF
URL. Nothing removed; every correction is an addition (struck/annotated inline, corrected
footnote, or new row) per the additive mandate.

- **R1-1/R1-2 (§0, both HIGH):** added a "Round 1 correction" block at the top of §0 re-verifying
  PR #14 merged to `main`, correcting the "unmerged" headline — WITHOUT naively flipping to
  "fixed": the judge call site (`debate.js:184`) is confirmed still unguarded. Reframed headline
  from "the fixes exist, but they have not shipped" to "the fixes shipped mid-debate — one exact
  defect class survived the merge." Original §0 prose preserved below the correction as the
  verified-at-the-time historical record, with the "Union finding" paragraph annotated superseded
  rather than deleted. Threaded the flip through §3 rows 1 (MERGED), 3 (MERGED, extend), 4
  (confirmed still OPEN — unaffected by the merge), 7 (confirmed still OPEN), 8 (skeleton MERGED,
  first live trial still owed), 13/14 (unaffected), 16/17 (MERGED), and §4 rows 3–4 and the shape
  verdict. New footnote [^Reverify47ae48d] carries the full re-verification trail; new
  [^PinnedRepoState] states the going-forward discipline.
- **R1-2 (§3 row 2, new row 2b):** reworded row 2 from conditional ("subsumed by #1 only if...")
  to factual — confirmed still open, pinned to `debate.js:184`. Added row 2b for the same-class
  `citationPasses` defect (R1-11, below) so it has its own graded row rather than living only
  inside §2.3's suite-case list.
- **R1-3 (§2.3 item 1):** rewrote the "null at every call site" claim to differentiate three
  failure classes (throws-then-recovers: judge only; silent-degrade-and-continue: frontier,
  red-lens, final-assembly; already-covered: blue-respond, guarded at the final ternary) instead
  of asserting uniform crash behavior across five sites.
- **R1-4 (§1.1, §3 row 6, footnotes):** corrected the "~19%/~95%" miscitation. Re-verified
  arXiv:2602.03794 live this round (abstract fetch + independent percentage search) — figures
  absent; paper's real citable claim is qualitative ("2 diverse agents match/exceed 16
  homogeneous"). Re-cited the 19% figure to its actual source, arXiv:2603.22103 (narrative-
  similarity annotation, narrower domain), with an explicit domain caveat; new footnote
  [^NarrativeSimilarity] carries the verified figures (r=0.388 vs. r=0.461; 76.0% vs. 75.3%
  ensemble accuracy). Dropped "~95%" — traces to no source found.
- **R1-5 (§1.1, footnote):** restated the "2–4 agents" plateau as qualitative aggregate synthesis
  rather than a single pinned figure; independent re-search this round corroborates the shape
  (moderate-complexity breakeven ~2 agents, harder tasks 3–4, continued gains to 7 on the
  hardest) without pinning to one of the four originally-bundled sources.
- **R1-6 (footnote [^PR14]):** corrected diffstat from "+2281/-46" to the live "+318/-48, 18
  files, 11 commits" (`gh pr view 14 --json additions,deletions,commits`, this round).
- **R1-7 (§4 counting-method note, two footnotes):** corrected "35-entry" to "21 entries" in both
  [^Run2Friction] and [^FrictionCount]; noted "35" is the file header's agent count, not the
  entry count.
- **R1-8 (§4 row 6):** corrected persistence from "run 2 r1–r2" to "run 2, round 1 only" — the
  cited friction file attests exactly one live-source-drift instance.
- **R1-9 (footnote [^SubagentWriteBug]):** softened the filename-independence inference — issue
  #13890 is a silent no-op write failure, this repo's block is an explicit worded refusal;
  named the signature difference rather than treating the analogy as load-bearing.
- **R1-10 (footnote [^WriteBlock]):** struck "report.md-specific" as this report's own conclusion
  — the same corpus's `run2-friction.md` line 3 and this retrospective's own round-1 red-merge
  hit on `red/findings.md` falsify it; the quote is now explicitly attributed to the backlog, not
  endorsed.
- **R1-11 (§2.3 item 4, new §3 row 2b):** flagged `citationPasses`'s never-rescaling as a live
  shipped defect (confirmed: `const` at line 139, outside the `while` loop at line 148), not a
  hypothetical test case — added the missing "known-failing until fixed" flag and a dedicated
  graded row.
- **R1-12 (§3 row 16, new row 16b):** named the doctrine-vs-routing tension (red-lens passes route
  to the cheap bulk tier; only red-merge gets judgment tier) and gave it an explicit two-option
  disposition (reclassify for keeper runs, or state the bounded tradeoff for dev/smoke runs)
  rather than leaving it silent.
- **R1-13 (§3 row 15, §4 row 5):** dropped the mechanically-unsound "skeleton fix may moot most of
  it by construction" clause (ENAMETOOLONG is a payload-length ceiling orthogonal to
  Write-vs-append); replaced with the mechanism-accurate mitigation (chunked-append helper);
  re-graded likelihood Medium to High given a third occurrence this round (red-merge seat, per
  debate.md's round-1 friction) and updated §4 row 5's persistence count from 2 to 3.
- **R1-14 (§3 row 13):** added the missing vetting-step sentence (pin version, review
  source/maintainer, scope permissions) before MCP-server adoption, closing by addition rather
  than blocking the adoption recommendation.
- **R1-15 (new §3 row 19, new §5 item 8):** added a graded row applying the report's own
  content-poisoning finding to FEOV's own WebFetch/WebSearch-driven research phase, with an
  explicit disposition (risk-accept, naming the existing leaf-node citation-verification lens as
  the real structural mitigation) instead of leaving the reflexivity gap silent. Flagged the
  citation-poisoning residual (a fabricated-but-consistent secondary source) as a new open
  question.
- **R1-16 (§3 row 6):** added an explicit redundancy floor to the lens-assignment proposal — the
  critical-stance/adversarial lens goes to at least 2 of N lanes, not 1-of-N — naming and closing
  the failure-concentration trade the original proposal left undiscussed.
- **R1-17 (§1.1, new paragraph):** added the missing disposition for cross-provider model
  diversity — named as the report's own citation's stronger lever, explicitly deferred (not
  adopted) because the harness's `model`/`judgmentModel` knobs select Claude aliases only and
  wiring a second provider is an infrastructure change an order of magnitude larger than the
  already-scoped lens-assignment fix.
- **R1-18 (§0 live addendum):** relabeled the write-block self-report "self-observed, not yet
  artifact-logged," explicitly downgraded to one data point rather than proof, and added the
  independent corroboration that does exist (red's own round-1 hit on `findings.md`, a different
  seat logging the class rather than the seat it vindicates) as the stronger evidence for "the
  class is alive," while being explicit that neither occurrence alone is a full forensic trace.
- **R1-19 (footnote [^WisdomCrowds]):** corrected the URL to the real path (verified live this
  round: 672KB PDF, not a 404) and removed the bare-domain form that 404s.
- **R1-20 (report header):** removed the false parallelism between lanes 1–2's quantified
  disconfirming-search ratios and lane 3's per-claim citation discipline (not a search-budget
  ratio); restated honestly rather than implying a third equivalent number.
- New §5 items 7–10: skeleton-fix-untested-against-real-path (corollary of R1-1/R1-2), the
  citation-poisoning residual (R1-15), whether reclassifying lens routing changes catch rates
  (R1-12), and the ENAMETOOLONG fourth-occurrence trigger (R1-13).
- New footnotes: [^Reverify47ae48d], [^JudgeUnguarded], [^SmokeAbsent], [^PinnedRepoState],
  [^NarrativeSimilarity]. All prior footnotes retained; five annotated in place with corrections
  rather than replaced, so the original citations remain visible alongside the fix.
- Nothing substantive was deleted: every correction above is either a struck/annotated inline
  note, a corrected footnote (original claim still quoted, correction appended), or a new
  row/paragraph/footnote. §0's original analysis is preserved verbatim below its Round 1
  correction block.

## Round 2 — address red's 11 gaps (R2-1..R2-11), all additive; one rebutted-with-evidence (2026-07-14)

Live re-verification this round: direct read of `debate.md`'s round-1 merge-seat friction section
(confirms no ENAMETOOLONG event there), `run1-friction.md`/`run2-friction.md` (confirms 2, not 3,
documented ENAMETOOLONG occurrences), `debate.js` verbatim (ledger clause lines 152-156, doctrine
comment, header `--smoke` description), `commands/research.md` (no `--smoke` string anywhere),
`ideas/backlog.md` item 28 at `main` @ `88eb57f` (cost-audit finding), `git ls-tree` at `88eb57f`
(confirms run 3 left no artifact trail), and independent WebFetch of arXiv:2602.03794 Table 2 (L4
vs. L2) and arXiv:2606.02646 (full abstract + HTML). Nothing removed; every correction is additive
per protocol; one required-fix (R2-4's proposed citation) is rebutted in writing with the fetch
evidence rather than applied as specified, since applying it as specified would have introduced a
new, worse miscitation.

- **R2-1 (§3 row 15, §4 row 5, §5 item 10, §4 shape-verdict paragraph):** corrected the
  ENAMETOOLONG recurrence count from the fabricated "3 times across 3 runs, per debate.md's
  merge-seat friction" to the honest "2 documented occurrences across 2 runs" (`run2-friction.md`
  line 4 + this retrospective's own round-0 heredoc failure); dropped the false debate.md citation;
  renumbered the §5 item 10 build-trigger from fourth to third occurrence; re-argued the High
  likelihood grade on the honest 2/2 rate rather than re-asserting it on the wrong count.
- **R2-2 (§0 live addendum, §4 row 4, §5 item 7):** corrected "this same round" (and its two
  inheritors) to the accurate "across rounds 0 and 1" / "one round after it hit blue" — blue's hit
  is dated round 0, red's is dated round 1, per the CHANGELOG and debate.md respectively.
- **R2-3 (§1.1 cross-provider paragraph, §3 row 6 note, §5 item 5):** corrected the "2 diverse
  agents match/exceed 16 homogeneous" misattribution — that result is arXiv:2602.03794's **L4**
  (different models AND personas) condition, not **L2** (persona-only, same base model), the
  condition the disposition's bracketed "[persona-lensed]" gloss actually substitutes in. L2's real
  curve needs **8 agents** to match the same 16-agent baseline (Table 2, independently re-fetched
  at the merge seat this round: L4 N=2 67.71% vs. L1 N=16 65.34%; L2 N=8 65.44%). The "defer, not
  adopt" disposition for cross-provider diversity now rests on the infrastructure-cost argument
  alone, not the mis-cited figure; §5 item 5's revisit-trigger baseline recalibrated to the honest
  L2 curve. Added a cross-reference note to §3 row 6 so the same figure is not misread twice.
- **R2-4 (footnote [^DiminishingReturns]) — CONCEDED ON THE GAP, REBUTTED ON THE PROPOSED FIX:**
  red correctly flagged the "continued gains to 7 agents on the hardest" clause as an uncited
  precise figure — the identical over-attribution failure R1-5 was raised to fix, recurring inside
  its own fix. Red's required fix proposed citing arXiv:2606.02646 for this exact figure. **Direct
  verification this round (WebFetch of the abstract and full HTML, 2026-07-14) shows this paper
  does not state it**: its hardest benchmark is GSM-Hard, not GSM-Plus; it states "on harder tasks,
  the practical knee is N≈10," not 7; and its headline finding is that effective team size
  plateaus around 1.8 agents by N=30 on free-form math, with a single N≤5 pilot sufficient to
  predict that ceiling. Citing this paper for "7 agents" would have repeated the exact miscitation
  pattern it was meant to fix. **Resolution: dropped the unpinned "7 agents" clause rather than
  re-cite it to an unverified source**; restated the qualitative synthesis without a specific
  agent-count ceiling for the hardest tier, and noted (accurately, per the paper's real finding)
  that harder tasks may show diminishing returns arriving earlier, not later, than the moderate-
  complexity case. This is a written rebuttal with evidence, not a silent skip of red's gap — the
  gap (uncited figure) is conceded and fixed; the specific proposed source is contested and shown
  not to support the number.
- **R2-5 (§2.3, §2.4, new footnote [^CostFigureProvenance]):** flagged that the report's headline
  token-cost figures (252.9k run 1, ~3M run 2) are self-reported and of unstated/possibly mixed
  provenance, per `ideas/backlog.md` item 28's live finding that the project's own panel token
  counter excludes cache traffic (~92% of real flow, 610K reported vs. 47.7M transcript-derived).
  Direction of risk is understatement only — the simulator/`--smoke` cost case is unweakened;
  pointed at the now-existing cost-audit tool for comparable figures before quoting again.
- **R2-6 (§3 row 11, §5 items 4 and 7):** noted that run 3 evidently already executed (two backlog
  commits cite its live measurements) with zero artifact trail in the tree — a live instance of
  exactly the gap row 11 argues for, and a silent mooting of §5 items 4/7's "pending live trial"
  framing, updated to say the answer may exist only as unlogged operator experience.
- **R2-7 (§3 row 19, §5 item 8):** rewrote the poisoning risk-accept's mitigation description from
  "independent re-verification against a second source" (a capability the protocol does not grant
  — same-source re-reading only, confirmed by a repo-wide grep for "independent" returning zero
  hits outside the ledger clause itself) to an honestly-scoped statement: the existing leaf-node
  lens catches source-misstatement (as it did for R1-4/R1-6), not self-consistent fabrication.
  Collapsed the previously-contradicting §5 item 8 into the same corrected statement. The
  risk-accept disposition itself stands, now on an accurate mechanism description.
- **R2-8 (§3 row 6 disposition):** added a reconciling sentence — the four-method roster plus a
  2-of-N redundancy floor arithmetically needs `lanes >= 4`, which is stated as a scoped exception
  to row 7's `lanes >= 3` floor and to the risk-accepted "blanket lane-count raise" (which always
  meant headcount-only raises, not a raise required by a stated method roster — the two claims are
  now distinguished rather than left to silently collide).
- **R2-9 (§3 row 10):** re-graded the impact cell to account for the shipped citation-ledger's
  actual skip-trigger (prose-change-keyed, verbatim at `debate.js:152-156`, confirmed live), which
  suppresses exactly the re-verification the row's "usually caught by re-verification" claim leans
  on; added the concrete fix (a time/access-date condition added to the same `ledgerClause` string)
  rather than leaving the two controls silently starving each other.
- **R2-10 (new §5 item 11):** added the missing dependency statement — recurrence-triggered
  risk-accepts (rows 14, 15) have no durable counter equivalent to the citation ledger, so treat
  "Nth occurrence triggers a build" as advisory until an equivalent ledger exists; recurrence counts
  must be re-derived from primary sources each time (R2-1 is the demonstration of what happens when
  they aren't).
- **R2-11 (§3 row 4):** reworded from "no `--smoke` flag in `commands/research.md` or `debate.js`"
  (over-reaching its own footnote, which checked only `commands/research.md`) to the accurate
  claim: no functional argument-parsing path exists; the string appears only in `debate.js`'s
  descriptive header comment.
- New round-2 summary paragraph added directly under the report's provenance note, previewing the
  two most load-bearing repairs (R2-3, R2-7) and the one rebutted-not-applied fix (R2-4) before the
  section-by-section detail.
- New footnote: [^CostFigureProvenance]. All prior footnotes retained; [^DiminishingReturns]
  annotated in place with a second round of correction (R2-4) rather than replaced.
- Nothing substantive was deleted: every round-2 correction above is a struck/annotated inline
  note, a corrected footnote/row (original text preserved, correction appended), or a new
  paragraph/item. Round 1's corrections remain intact beneath round 2's.

## Round 3 — address red's 10 gaps (R3-1..R3-10), all additive, none rebutted (2026-07-14)

Live re-verification this round: direct read of `debate.js` on `origin/main` @ `d164ab2` (HEAD;
confirmed docs-only commits since `47ae48d` — `debate.js` byte-identical, so `47ae48d` and
`d164ab2` are cited interchangeably for source claims), hand trace of the `{verdict:'FAIL',
gaps:[]}` control-flow path (lines 56–91 schema, 148–198 loop, 200–218 terminal return), direct
read of `tests/simulator/debate.test.mjs` lines 114–123 (friction test), a fresh `git grep -ni
"independen"` across the plugin (zero hits), direct read of `run2-friction.md` line 4 and
`blue/CHANGELOG.md` Round 0's own heredoc-failure sentence, and a live re-fetch of `ideas/backlog.md`
item 28 at `d164ab2` (gained sub-item (d) since the round-2 pin at `88eb57f`). Nothing removed;
every correction is additive per protocol; no required fix was rebutted this round.

- **R3-1 (§2.1 Tier A round-loop row, §2.3 item 8, new §3 row 20, new §2.3 addition 13):** added
  the schema-legal `{verdict: 'FAIL', gaps: []}` degenerate shape — hand-traced to show it never
  reaches the judge branch (`contested` is always empty), loops silently to `maxRounds`, and
  returns `verdict: 'UNVERIFIED'` paired with `gaps_outstanding: 0`, a self-contradictory terminal
  state. Corrected §2.1's round-loop row description (previously implied full coverage), added a
  distinguishing note to §2.3 item 8 (a different degenerate-input case, not a substitute), added
  founding-suite addition 13 (known-failing until the new §3 row 20 guard ships), and graded the
  fix medium-high (low-medium likelihood x medium-high impact x low complexity).
- **R3-2 (§2.1 Tier A friction-aggregation row, §2.3 addition 14, new §3 row 21):** corrected
  "never dropped" — `takeFriction` is called at exactly three sites (red-merge, judge,
  blue-respond) and never for `blue-synthesize`, despite `BLUE_ENVELOPE` declaring `friction` on
  that identical schema; this report's own §0 write-block addendum (a round-0, blue-synthesize-seat
  complaint) is a live instance of the dropped-event class. Corrected the "every seat" reading of
  the merged simulator test (direct read: it stubs only `red`-merge and `blueRespond` friction,
  never `blueSynthesize`) rather than letting 11/11 green stand as unqualified confidence. Added
  founding-suite addition 14 and new §3 row 21 (one-line fix: `takeFriction('blue-synthesize',
  blueEnv)` after the line-136 null-guard).
- **R3-3 (§2.3 item 5, new §3 row 22):** corrected item 5's "with its required-fix intact" framing
  — true for the red gap object, not for the judge's carried-gap rationale, which the judge prompt
  explicitly requests ("for carried, state what further research blue owes") but which is never
  read again by the script or by `blue-respond`'s prompt. Added §3 row 22 with two fix options
  (fold into `openGaps`, or read `debate.md`'s latest `### LEAD` section) and adopted the cheaper
  option as a fix-before-run-4 rather than a risk-accept, since low complexity does not clear the
  bar for risk-acceptance.
- **R3-4 (§1.1 body clause):** corrected the §1.1 body's surviving "continued gains observed to 7
  agents on the hardest" clause — the exact figure its own footnote had already retracted in
  round 2 (R2-4) — to match the footnote's corrected synthesis; named the body-lags-footnote
  pattern explicitly (the reverse of the usual footnote-lags-body direction) since a majority-surface
  reader meets the body text first.
- **R3-5 (§3 row 6 R2-8 reconciliation):** corrected the round-2 reconciliation's arithmetic — three
  unfloored methods (1 lane each) plus one floored-to-2 method is 3 + 2 = 5 lane-assignments
  minimum, not the stated `lanes >= 4`, which holds only if two of the four named methods silently
  merge (contradicted by the same sentence's "four named methods"). Also corrected "row 7 floors N
  at 3" — row 7 is graded [OPEN], an unbuilt proposal, not a shipped floor. Restated: item 6's full
  roster needs `lanes >= 5`; row 7's unbuilt proposal targets `lanes >= 3` for runs not adopting
  item 6's full roster.
- **R3-6 (§3 row 19):** corrected "a repo-wide grep for 'independent' ... returns zero hits outside
  this ledger clause's own text" (implying one hit inside the clause) to the live-reverified
  "returns zero hits anywhere, including inside the ledger clause itself" — stronger for the row's
  own point, not weaker; noted the imprecise phrasing originated in red's own round-2 merge text.
- **R3-7 (§3 row 15, §4 row 5):** narrowed the "2/2 rate" claim — occurrence 1 (`run2-friction.md`
  line 4) is confirmed as the Windows command-length-ceiling mechanism; occurrence 2
  (`blue/CHANGELOG.md` Round 0) is confirmed only as "failed on shell parsing," which a quoting or
  CRLF issue (a separately documented fragility class in this corpus) could also produce.
  Restated as 1-confirmed + 1-same-family-plausible, not 2-confirmed-identical; kept the High
  likelihood grade as an argued risk-accept (transcript likely unrecoverable) rather than silently
  asserting a confirmed match.
- **R3-8 (footnote [^CostFigureProvenance], §3 row 18):** re-pinned the footnote from `88eb57f` to
  `d164ab2` (three commits forward), adding the backlog item's new sub-item (d) — a merge-seat cost
  analysis naming TURNS x CONTEXT (not report length) as the actual driver, with a concrete
  shard-the-findings lever. Added this as new, directly relevant evidence to §3 row 18's
  audit-narrowing hold, including the sharding proposal as a first candidate scoping rule for a
  future revisit rather than a blank-slate redesign.
- **R3-9 (footnote [^DiminishingReturns]):** disambiguated the R2-4 replacement sentence's internal
  contradiction ("breakeven higher... arriving even earlier... rather than later") by naming two
  distinct quantities — the nominal-N practical knee (higher on harder tasks) and the
  effective-diversity saturation ceiling (reached earlier in nominal-N terms regardless) — as
  non-contradictory claims about different curves, rather than alternative readings of one number.
  Third consecutive round of defects in this exact footnote (R1-5, R2-4, R3-9), noted explicitly as
  a standing argument for the claim manifest applying to blue's own footnotes.
- **R3-10 (§2.1 Tier A first row, §3 row 4):** propagated `[^CostFigureProvenance]` to the two
  reading-order-first instances of the token-cost figures that the R2-5 caveat had not yet reached
  (§2.1's opening row, §3 row 4's impact cell) — direction understatement-only, no verdict change.
- New round-3 summary paragraph added directly under the round-2 summary paragraph, previewing all
  ten corrections' shape before the section-by-section detail.
- No new footnotes; all edits are in-place corrections to existing footnotes/rows/list items or
  new rows/list items (§3 rows 20–22, §2.3 additions 13–14). Rounds 1–2's corrections remain intact
  beneath round 3's; nothing substantive was deleted.

## Round 4 — address red's 5 gaps (R4-1..R4-5), all additive, none rebutted (2026-07-14)

Live re-verification this round: `git log --oneline -5 origin/main` (HEAD `42dba2d`, one commit
past the round-3 pin `d164ab2`, docs-only), direct read of `debate.js` on `main` @ `42dba2d`
(byte-identical to `d164ab2`/`47ae48d` — confirmed no code drift since round 1), hand trace of
`RED_ENVELOPE`'s schema (lines 56–91, no `supersedes` property) and the contested-detection line
(`contested = redEnv.gaps.filter(g => prevGapIds.has(g.id))`, pure id-string-equality), `grep -n
"^### " debate.md` (8 matches, all `### BLUE`/`### RED`, zero `### LEAD`), `git show 42dba2d --
ideas/backlog.md` (confirms the docket-detector-tracks-IDs-not-lineages entry, its own worked
example, and its own drafted `supersedes`-field fix, landed 2026-07-14T00:49:14-07:00 — 25 minutes
after this report's round-3 pin), and cross-checked debate.md's round-4 RED merge-seat chain
enumeration against this corpus's own `red/findings.md` round-by-round gap ids (confirmed:
R1-5→R2-4→R3-4/R3-9, R2-5→R3-10, R2-7→R3-6, R2-8→R3-5→R4-3 — four chains, matching red's count,
the fourth continuing live into this very round via R4-3). Also verified
the memory-architecture corpus's own `red/findings.md` gap-id range extends to R4-12 (lines
115–118), confirming the cross-corpus id-collision claim. Nothing removed; every correction is
additive per protocol; no required fix was rejected, one disjunction was resolved by decision
(R4-2) rather than left open.

- **R4-1 (§2.1 Tier A gap-id-rollover row split into two sub-rows; §2.3 addition 3 unchanged +
  new addition 15; new §3 row 23; new §5 item 12; new round-4 summary paragraph):** the single
  "gap-id rollover" row previously described only the narrower same-id-skips-a-round failure
  (fixed by widening `prevGapIds` to full adjudicated history). Added the second, independent
  failure class: the detector is pure id-string equality with no `supersedes`/lineage field, so
  red's own closed-WITH-REGRESSION practice — minting a fresh id per successor gap — means a
  multi-round dispute lineage never arms the docket regardless of window width. Confirmed live:
  zero `### LEAD` headers across this corpus's three completed rounds despite four full-length
  regression chains; independently confirmed by the project's own backlog naming this
  retrospective's own chain as its worked example. New §3 row 23 grades the `supersedes`-field fix
  high/high/medium and scopes it as additive to, not a substitute for, the `prevGapIds`-widening
  fix already scoped at row 2/addition 3. New §5 item 12 states the two fixes' independence
  explicitly, closing the risk that one fix is mistaken for subsuming the other.
- **R4-2 (§3 row 20; §2.3 addition 13):** row 20 previously shipped red's own R3-1 required-fix
  text as an unresolved "either/or" (logged-warning-PASS vs. throw) — the only round-3 fix left as
  a disjunction where its siblings (rows 21, 22) shipped a decided choice. Decided: throw, with a
  stated reason (a degenerate `{verdict:'FAIL', gaps:[]}` return is evidence of a broken merge
  lens, not evidence the report is clean; converting it toward `PASS` manufactures a false-positive
  verdict, which the report's own anti-silent-degradation argument — used at row 19's poisoning
  finding, §2.3 item 1's throws-vs-degrades taxonomy, and R2-7's honest-mitigation-scoping —
  already argues against everywhere else). Added the concrete thrown-error message. Extended §2.3
  addition 13 with the matching positive assertion (the round throws and terminates without
  reaching `maxRounds`), while noting the negative/option-agnostic assertions were already writable
  before this decision — red's "cannot be written" framing for the test itself was overstated,
  even though the guard's positive behavior genuinely was undecided until now.
- **R4-3 (§3 row 6, disposition's operative sentence):** the sentence instructing "assign the
  critical-stance/adversarial-disconfirming lens to at least 2 of N lanes" used its slash as a
  synonym-joiner (one method, two labels) while the four-method roster two sentences later uses
  the identical slash-separated form as a list separator (four distinct methods) — nothing flagged
  the switch, so a reader stopping at the operative sentence alone could reconstruct the exact
  `lanes >= 4` misreading R3-5 closed downstream. Edited the sentence itself: "the
  adversarial-disconfirming-first lens (a distinct method from local-repo critical-stance, named
  separately below)" — fixed at the source rather than only at R2-8/R3-5's downstream arithmetic,
  closing the repair-reaches-the-conclusion-not-the-source class red flagged as a regression of
  R3-5's own closure.
- **R4-4 (risk-accepted closing paragraph before §4):** corrected the fifth, previously-uncorrected
  instance of R2-1's retracted "4th occurrence" figure to "third occurrence (corrected R2-1)",
  matching the three already-corrected instances (row 15, §4 row 5, §5 item 10).
- **R4-5 (§1.2's R2-10 mention; §3 row 13; §4 rank-1 row; §2.1 Tier C lossy-fetch bullet; new
  footnote [^GapIdScheme]):** this report and the memory-architecture corpus both use a bare
  `R<round>-<n>` gap-id scheme, and both are now into round 4 (this report: R4-1..R4-5;
  memory-architecture: confirmed live to run to R4-12) — the schemes collide in form. Four prior
  bare cross-references into the memory-architecture corpus's own gap ids (not this report's)
  are now prefixed `MA-` at all four locations; this report's own ids remain bare, unchanged, per
  every prior round's usage. Chose the four-location prefix fix over a full-document rename of
  this report's own (far more numerous) ids, which would be a high-blast-radius edit for a
  collision that has not caused an actual misread within this document to date.
- New round-4 summary paragraph added directly under the round-3 summary paragraph, previewing all
  five corrections' shape before the section-by-section detail.
- New footnote: [^GapIdScheme]. All prior footnotes retained; no footnote replaced, only extended
  or added to. Rounds 1–3's corrections remain intact beneath round 4's; nothing substantive was
  deleted.

## Round 5 — address red's 6 gaps (R5-1..R5-6), all additive, none rebutted (2026-07-14)

Live re-verification this round: direct read of `debate.js` (unchanged, byte-identical to the
round-4 pin — `friction` array at line 145, throw sites at lines 36/136/171, `takeFriction` call
sites and their absence for `blue-synthesize`); direct read of `commands/research.md` step 5
(friction.md write is envelope-gated, confirming it never fires on a throw); direct read of
`debate.md`'s round-4 BLUE section (item 1's "adopted it in place of my own" sentence, confirming
the substitution's scope); `grep -n "^### " debate.md` (10 headers: 5 BLUE rounds 0–4, 5 RED rounds
1–5, zero LEAD); and a full first-hand read of the memory-architecture corpus's current
`red/findings.md` (round-4 state) to check every one of the six previously-cited MA gap ids'
present status rather than re-trusting the round-2/round-4 citation. Nothing removed; every
correction is additive; no required fix was rebutted this round.

- **R5-1 (§3 row 23's likelihood cell):** the row's chain enumeration was still shipping blue's own
  discarded first-pass list (`R1-13→R2-1→R3-7; R1-16→R2-8→R3-5; R2-5→R3-8`) even though round 4's
  own debate record shows the synthesizer explicitly comparing it against red's more precise
  enumeration and adopting red's in its place — a substitution that reached §2.1(b) only, not this
  row. Two of the three discarded entries are independently contradicted by `red/findings.md`'s own
  status lines (R2-5's successor is R3-10, not R3-8; R2-1 closed clean, no R3-7 reopening); the
  third truncated the chain's live R4-3 tip. Corrected to §2.1(b)'s verbatim list, one clause, no
  new research.
- **R5-2 (§2.1 Tier C bullet, §3 row 13, §4 row 1; new footnote [^MAStatusR5]):** all three
  locations cited a stale round-2 snapshot of six memory-architecture gap ids as currently blocked
  by lossy PDF/HTML fetch, reasserted without a diff at rounds 4 and 5. Direct re-read of the
  current corpus: MA-R1-28 and MA-R2-8 CLOSED round 3 by ordinary live re-fetch (no PDF tool
  involved — falsifying the citation's own implied prediction that one would be needed); MA-R3-14
  and MA-R3-15 CLOSED round 4, same mechanism; MA-R4-9 is open but a diagnosed miscitation (wrong
  paper), not a lossy-fetch case. Only MA-R1-19 remains genuinely open and friction-blocked. All
  three locations reconciled to this one corrected reading (previously 6/5/6-member lists that
  disagreed with each other); §4 row 1's build-priority disposition is explicitly preserved as
  unaffected, since it rests on the friction recurring every round for four rounds (unchanged
  historical fact) and this retrospective's own #1 friction ranking, not on the memory-architecture
  corpus's current open-gap count.
- **R5-3 (front matter, §2.1(b), §3 row 23 — three instances):** "the judge was never dispatched...
  across three completed rounds" was one round stale — round 4 also completed judge-free. Reworded
  round-agnostically ("every completed round to date") at all three locations rather than re-dating
  to "four," so the phrasing does not go stale again next round; `grep -c "three completed"` against
  the corrected file returns zero raw (unquoted) instances.
- **R5-4 (§2.3 addition 15):** the simulator case's canned-chain description claimed both mirrored
  closures read "WITH REGRESSION" uniformly — the real chain it claims to mirror (R1-5→R2-4) closes
  its second link "REBUTTAL ACCEPTED WITH EVIDENCE," a different label. Loosened to label-agnostic
  phrasing ("whatever its actual closure label") with the real labels quoted; noted explicitly that
  the detector logic under test is label-independent, so test validity was never affected — this is
  a framing correction, not a test-design correction.
- **R5-5 (§3 row 23's complexity cell, §2.3 addition 15 extended):** the originally-scoped
  `supersedes` fix (add the schema field, follow it in `contested` detection) relies on exactly the
  unenforced good-faith prompt compliance the row's own likelihood cell indicts two sentences
  earlier — nothing validates that a regression-flagged closure actually names its successor, and
  this corpus has already demonstrated that unenforced-instruction failure mode twice (the friction
  field silently unset for three rounds; an undecided disjunction shipped verbatim). Added a fourth,
  script-level structural check: after `red-merge` returns, throw if a "WITH REGRESSION"-class
  closure names no successor in any gap's `supersedes` array — mirroring the R4-2 precedent of
  throw-over-silent-acceptance. Extended §2.3 addition 15 with the matching known-failing assertion.
  Not risk-accepted: the check is a few lines reusing the same fields the original fix already adds.
- **R5-6 (§2.1 Tier A friction row, new §3 row 24, new §2.3 addition 16):** the friction-aggregation
  array (`debate.js` line 145) is script-local and read out only at the terminal return and the
  final-assembly prompt — never at any of the loop's throw sites (args guard, `blueEnv`/`redEnv`
  null-guards, and soon row 20's guard), and `commands/research.md` step 5 only writes `friction.md`
  on a successful return, confirmed by direct read. A mid-run throw discards every seat's
  accumulated friction, losing the self-improvement signal for exactly the runs that crashed. Added
  new §3 row 24, scoped as a prompt-only fix (each schema'd seat also appends its friction directly
  to `friction.md` via Bash, reusing the already-proven row-8 write-block-workaround pattern) rather
  than a script-side write (which would violate the script's own stated no-filesystem-access
  doctrine) or a bare risk-accept (the fix is cheap enough that silence was not warranted). Added
  founding-suite addition 16.
- New footnote: [^MAStatusR5]. All prior footnotes retained; [^GapIdScheme] not modified (its own
  four-location correction stands; R5-2 is a distinct correction, to currency of status, not to
  naming discipline). Rounds 1–4's corrections remain intact beneath round 5's; nothing substantive
  was deleted.
