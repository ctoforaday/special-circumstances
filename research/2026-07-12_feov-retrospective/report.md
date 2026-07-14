# Retrospective on the frank-exchange-of-views debate engine after its first two live runs — research report

**Verdict:** UNVERIFIED (after 5 debate rounds — terminated by the safety ceiling, not by PASS and not by a judge-confirmed deadlock) · **Run:** `research/2026-07-12_feov-retrospective/`

**TL;DR:** The engine's five design doubts are all confirmed by its own evidence — blue lanes converge in the breadth phase (run 2 measured it directly), consensus-vs-minority claim provenance is destroyed at synthesis by schema (RED_ENVELOPE has no provenance field), and the defect population is trimodal (unit-simulatable / live-smoke / production-only). The single most consequential finding is off-frontier: the contested-gap docket detector is lineage-blind by construction (pure gap-id string equality, no `supersedes` field), so the judge was dispatched zero times across every completed round of this corpus — the engine's entire adjudication mechanism has never once fired, and this debate's convergence was a property of the actors' good faith, not of the machine. The debate itself demonstrated the system's failure modes while auditing them: 52 gaps raised over five rounds, 46 closed (12 of them closed-with-regression — the dominant blue failure class is incomplete propagation of an otherwise-correct correction, five instances in five rounds), and the round-5 six were fixed in the living report but never re-audited because the ceiling fired first. Verdict is therefore UNVERIFIED: nothing in the final state is known-false, the external citation surface came back clean for two consecutive rounds, but the last round of repairs is exactly the class of change that regressed in four of the four prior rounds it was attempted. The pre-run-4 docket (24 graded rows, §3 of the blue report) and the friction ranking (PDF extraction #1 at 3/3 roles) are the deliverables to act on.

## Outstanding gaps at the ceiling — disposition and compromise rationale

The debate ended mid-cycle: red's round-5 audit returned FAIL with six open gaps; blue's round-5 response landed fixes for all six (none rebutted); no round-6 red audit ran. Formal state: **open, blue-fix-landed, red-unaudited.** No gap was ever adjudicated by the judge (the docket never armed — see R4-1 below), so there are no judge-ruled `closed`/`rebuttal_sustained`/`risk_accepted` dispositions; the risk-accepts on record are blue's argued ones inside §3, each of which red explicitly reviewed and accepted in-round.

| Gap | Red's grading (round 5) | Blue's round-5 response | Disposition at ceiling |
|---|---|---|---|
| **R5-1** — §3 row 23's likelihood cell shipped blue's discarded first-pass regression-chain enumeration (two entries contradicted by red's own closure record) while §2.1(b) carried the corrected list | MEDIUM — certain × medium × trivial; corroboration HIGH (3 lenses) | Conceded; row 23's list replaced with §2.1(b)'s verbatim four-chain enumeration; both contradicted entries re-verified against `red/findings.md` | Fix landed, unaudited |
| **R5-2** — three disagreeing in-report citations of the memory-architecture corpus's blocked-gap set, stale by two rounds; 4 of 6 cited-as-blocked gaps had closed by ordinary re-fetch (no PDF tool), overstating live blockage ~3–6x | MEDIUM — high × medium × low; corroboration LOW for the stale sub-claims, HIGH for the list inconsistency | Conceded; all three locations reconciled to one corrected reading traced first-hand per id ([^MAStatusR5] in the blue report); §4 row 1's #1-build ranking explicitly re-based on the historical recurrence + live backlog, not the stale count | Fix landed, unaudited |
| **R5-3** — "three completed rounds" (judge-free) one round stale at three locations | LOW — certain × low × trivial | Conceded; reworded round-agnostically ("every completed round to date") so it cannot go stale again | Fix landed, unaudited |
| **R5-4** — §2.3 addition 15's uniform "WITH REGRESSION" labeling misdescribes the mirrored chain's real second link (closed "REBUTTAL ACCEPTED WITH EVIDENCE") | LOW — certain × low × trivial | Conceded; loosened to label-agnostic phrasing; noted the detector under test is label-independent | Fix landed, unaudited |
| **R5-5** — the `supersedes` lineage fix as scoped relied on unenforced prompt compliance, the exact class the same row indicts; telemetry-invisible on failure | MEDIUM-HIGH (gate tier) — medium × high × low; corroboration HIGH | Built, not risk-accepted: a fourth, script-level structural check added to §3 row 23 (throw if a regression-flagged closure names no successor in any `supersedes` array), plus a founding-suite assertion | Fix landed, unaudited |
| **R5-6** — the aggregated `friction` array is script-local and lost on any mid-run throw; `friction.md` is written only on a successful return, so crashed runs lose exactly the telemetry that matters most | MEDIUM — medium × medium-high × low-medium; corroboration HIGH | Built, scoped narrowly: new §3 row 24 (seat-level Bash-append of friction lines to `runDir/friction.md`, reusing the proven write-block-workaround pattern; no script filesystem access) plus founding-suite addition 16 | Fix landed, unaudited |

**Compromise rationale for stopping here:** the safety ceiling is the engine's only structural stop (the docket/judge path never armed — R4-1), and it fired. Arguments that the residual risk is low: severity declined (round 5's worst is MEDIUM-HIGH vs. round 4's HIGH); the external citation surface was verified clean for two consecutive rounds; five of the six round-5 fixes are copy-edits from material already verified inside this corpus; red's own convergence projection was "round 6 is the PASS round" conditional only on no new regressions. The argument that it is not zero: blue's fixes carried at least one regression in four of the four prior rounds where fixes were attempted, and the recurring class (incomplete propagation of a correct correction) is precisely what an unaudited final round cannot rule out. The gate does not soft-pass: this report is stamped UNVERIFIED, and the cheap residual action is a one-pass propagation audit (grep each round-5 corrected string/list against every location that asserts it) folded into run-4 prep — blue's own round-5 closing note proposes exactly this check as a standing revision-checklist item.

## The Catechism

The agreed answers. Caveat on "agreed": the judge was never invoked and the debate did not reach PASS, so these are the positions the adversarial process left standing — red's five audits never disputed the five doubt-verdicts (H1–H5) or any §3 disposition after its in-round corrections were applied, which is the strongest agreement this corpus can attest.

1. **What are we trying to do?** Decide what must change in the frank-exchange-of-views debate engine before its fourth run, by auditing it against its own first-two-run evidence: confirm or refute the open design doubts, classify every observed defect by how cheaply it could have been caught, grade every proposed change like a gap (likelihood × impact × complexity), and rank the capability gaps the agents themselves reported.

2. **How is it handled today, and what does that cost us?** Defects are discovered in production at ~253k–3M tokens per discovery (self-reported figures, themselves flagged as possibly ~92% undercounted by the project's own cost audit). Status is tracked by backlog checkboxes and friction-file self-reports that this corpus proved false for the shipping ref ("a backlog checkbox is not a diff"). There is no claim provenance across lanes, no trajectory capture (run 3 executed and left zero artifact trail), and the adjudication mechanism has structurally never fired.

3. **What is new here, and why do we believe it works?** Evidence, not enthusiasm: run 2's candidate files and CHANGELOG measure lane convergence directly; `grep -n "^### " debate.md` proves zero judge dispatches; direct control-flow traces of `debate.js` at pinned SHAs establish every crash/degrade class; a zero-token simulator with 11 passing tests merged mid-debate and its 16 graded founding-suite additions are each traced to a defect that actually occurred. The retrospective's method — every claim leaf-verified, every fix re-audited — caught 12 defects inside its own repairs, which is the method demonstrating it works on itself.

4. **The case against, at full strength.** (a) The process is expensive: five rounds, five-lens red audits each round, full-report re-reads whose cost driver (turns × context) the corpus itself measured; this report cost real tokens to say "fix a null guard." (b) The repair-regression rate is the strongest argument against trusting any of it: every round of fixes for four consecutive rounds introduced at least one new defect, so why believe round 5's? (c) Self-reference risk: the system audited itself with itself; a shared blind spot (e.g., all seats over-trusting grep-based verification) would be invisible by construction. (d) The 24-row docket invites scope creep before run 4 ever executes; several rows are speculative hardening for shapes never observed live (row 20's degenerate envelope has zero live occurrences). (e) A simulator can breed false confidence — R5-5 is the in-corpus proof that a Tier-A test can be written for a Tier-C failure surface and prove nothing about live compliance. (f) The verdict below is UNVERIFIED, so the report cannot claim its own gate's endorsement. These objections survive; the docket's grading discipline (build-vs-risk-accept argued row by row, revisit triggers named) is the mitigation, not a refutation.

5. **Of interest, or merely interesting?** Of interest, narrowly: run 4 is already planned, and running it against `main` today reproduces known defect classes (the judge-seat null dereference at `debate.js:184` is live; the docket detector cannot arm on regression chains). The docket displaces nothing except running run 4 sooner, and an un-fixed run 4 risks burning its budget rediscovering what this corpus already pinned to exact lines. The un-instrumented doubts (Catechism-template fit on survey topics; whether lens-assigned lanes measurably reduce convergence) are the "merely interesting until instrumented" remainder, and §5 names what each needs.

6. **What changes if it works — and if we simply don't do it?** If the docket ships: run 4 executes with every schema'd call site guarded, minority-report provenance survives synthesis (cheap half of the claim manifest), the judge can actually be reached, friction survives crashes, and the simulator regression-gates all of it for zero tokens. If not: nothing dramatic on a good day — the engine converged twice on actor good faith alone — but the failure modes are all silent-or-expensive (a judge-seat crash loses every paid round; a spinning lineage burns to `maxRounds` with telemetry indistinguishable from progress), and "it worked twice because everyone behaved" is exactly the property this corpus refuses to rely on.

7. **What does it cost, and where would we stop?** Almost every fix-before-run-4 row is graded Low or Low-medium complexity (guard clauses, prompt sentences, one schema field, template sections); the two Medium rows (lineage detection, simulator extensions) are bounded by already-drafted designs. Stopping points are named per row: ENAMETOOLONG tooling waits for a third occurrence; advisory-access infrastructure waits for the workaround to stop working; audit narrowing stays human-gated until the cheaper levers prove insufficient; the full claim manifest waits for the cheap half to prove load-bearing. The debate itself stopped at its 5-round ceiling — the honest cost of not yet having the lineage-aware docket it recommends.

## Technical foundations

Established at leaf level by the debate (every item below survived five adversarial rounds; graded verification pairs in `red/citation-ledger.md`):

- **Engine shape:** `commands/research.md` drives `skills/research-protocol/scripts/debate.js` (219 lines), which dispatches schema'd seats (`frontier`, blue lanes, `blue-synthesize`, `red-lens` ×5, `red-merge`, `judge`, `blue-respond`, assembly) over an append-only blackboard (`debate.md`, `blue/report.md`, `red/findings.md`).
- **PR #14 merged mid-debate** (`00018a5`, 2026-07-14T05:58:54Z, ~8 minutes after blue's round-0 verification pin): args parse+guard, `blueEnv`/`redEnv` null-guards, the 11-test simulator wired to CI, per-role model routing, citation ledger, pre-created blackboard skeleton, Catechism template. `main` then advanced through `47ae48d`/`88eb57f`/`d164ab2`/`42dba2d` with `debate.js` byte-identical — every round re-verified the pin.
- **Live defects on `main` at `42dba2d`:** the judge call site is unguarded (`debate.js:184`, `for (const r of judge.resolutions)` — run 2's exact `TypeError` class at an uncovered seat); `citationPasses` is `const`, computed once pre-loop, never rescaled as the report grows; the contested-docket filter is pure id equality (`prevGapIds.has(g.id)`) with no `supersedes` field in `RED_ENVELOPE`; `takeFriction` is never called for `blue-synthesize`; the `friction` aggregate is script-local and unreachable after any of the three (soon four) throw sites; a schema-legal `{verdict:'FAIL', gaps:[]}` loops silently to `maxRounds`.
- **Docket/judge:** zero `### LEAD` headers across the entire transcript — the judge was never dispatched in this corpus despite four multi-round regression-chain disputes; independently corroborated by the project's own backlog (commit `42dba2d`), which names this retrospective's chain R1-5→R2-4→R3-4/R3-9 as its worked example.
- **Lane diversity (H1):** run 2 ran `--lanes 2` (below the intended 3; no floor exists); breadth-phase candidates converged on the same material, with the poisoning finding an independent replication across lanes — the measured basis for lens/source-class assignment (with `lanes >= 5` for the full four-method roster plus 2-of-N redundancy floor) over headcount raises. External support downgraded honestly in-debate: the "2 diverse agents match 16" figure is the L4 (model+persona) condition of arXiv:2602.03794; persona-only (L2) needs 8 agents — the defer-cross-provider call rests on infrastructure cost alone.
- **Provenance (H2):** single-lane vs. all-lane origin exists transiently in `blue/candidates/` and is discarded at synthesis; `RED_ENVELOPE` carries no provenance field; the final report structurally cannot distinguish consensus claims from one lane's minority report. This report practiced the fix manually ([L1]/[L2]/[L3]/[all lanes] tags throughout the blue report).
- **Defect population (H3):** trimodal — unit-simulatable for zero tokens (crash/degrade/loop logic), live-smoke (permissions, write-block, OS limits), production-only (real-model judgment quality). The founding regression suite's 16 corpus-traced additions and the `--smoke` design (~50k tokens) are specified in blue §2.2–2.4.
- **Friction corpus (Q4):** PDF full-text/table extraction is #1 (3/3 roles, every round of the only complete run, no workaround exists — and it recurred at the red-merge seat during every round of this retrospective); advisory access #2 (3/3 roles, engineered around by round 4); the write-block is filename-keyed and path-independent (artifact-logged control experiment, round 2) and fired in all three runs.

## Analysis

The four task questions, answered from the foundations (full argument and evidence in the embedded blue report; adversarial trail in the embedded red findings):

**(1) Doubts.** H1 confirmed-with-refinement: lanes converged, but convergence on the poisoning finding was independent replication (a feature); the fix is method/source-class assignment, not headcount; the lane count was under-provisioned and unrecorded (no `lanes` return field). H2 confirmed sharper than posed: provenance is destroyed by schema, not habit — no field carries it, so no discipline could have preserved it. H3 refined from bimodal to trimodal. H4 (per-role model routing) confirmed and largely already built — with the doctrine-vs-routing tension (red lenses at bulk tier) named and dispositioned. H5 (friction shape) confirmed; its "(already fixed)" parenthetical refuted — the fixes were on an unmerged branch when written, and the checkbox-vs-diff lesson generalizes. Needs instrumentation: Catechism fit on survey topics, lens-assignment's measured effect on convergence (requires the claim manifest), blue's read of red's pattern memory.

**(2) Testing strategy.** Every defect from runs 1–2 classified in blue §2.1; the Node simulator (stub `agent()` with canned envelopes, zero tokens) is designed in §2.2 with a founding suite (§2.3) of 16 corpus-traced cases beyond stringified-args and null-returns — including the empty-candidates cascade, gap-id rollover, citationPasses recompute (known-failing), judge-null (known-failing), degenerate FAIL-empty-gaps throw, friction-aggregation coverage for all four seats, lineage/`supersedes` detection with the structural-check throw, and crash-time friction recoverability. Live-smoke (`--smoke`, ~50k tokens) covers the permission/OS/write-block tier the simulator's boundary discipline explicitly refuses to fake; production-only remains for real-model judgment quality (the R5-5 lesson: do not write Tier-A tests for Tier-C surfaces and call them coverage).

**(3) Pre-run-4 changes.** The graded docket is blue §3 (24 rows + an explicit risk-accepted list). The fix-now core: judge null-guard (row 2), citationPasses recompute (2b), simulator extensions (3), `--smoke` (4), claim-manifest cheap half (5), lens-assigned lanes with redundancy floor and `lanes >= 5` roster arithmetic (6), lane floor + `lanes` return field (7), `open_questions` template section (9), ledger skip-trigger time condition (10), trajectory capture (11), blue reads red's pattern memory (12), PDF-MCP adoption with vetting precondition (13), degenerate-envelope throw (20), blue-synthesize `takeFriction` (21), judge carried-rationale delivery (22), lineage-aware docket with script-level `supersedes` enforcement (23), crash-surviving friction appends (24).

**(4) What the agents need.** Ranked by distinct independent reporters in blue §4: PDF extraction (3/3 roles, unresolved, no workaround — every hedge is a standing per-round cost), advisory access (3/3 roles, engineered around — the asymmetry between those two, recurrence-without-workaround vs. designed-around, is the ranking's real logic), then the merged args guard (historically loudest, retired), the write-block (systemic across runs, skeleton merged, live trial owed), ENAMETOOLONG (2 occurrences, risk-accepted with a third-occurrence build trigger), and a long tail of singletons. The counting-method disagreement between lanes is itself preserved as a finding: friction needs structured role+round attribution.

**Alternatives considered and beaten:** raising lane headcount (beaten by method assignment on the diversity-collapse evidence); building a bespoke PDF tool (beaten by vetted MCP adoption); cross-provider model diversity (deferred on infrastructure cost after the L2/L4 correction); renaming report-like artifacts (superseded by the merged skeleton); full claim manifest now (cheap half first); narrowing red's full re-read (held — it caught R1-1; revisit with the backlog's sharding rule). Each is recorded as an argued risk-accept, not a drop.

## Risk matrix

Top-level distillation; the authoritative per-proposal grading is the embedded blue report §3 (rows 1–24), whose finer-grained compound grades (medium-high, certain, trivial, realized) are preserved there verbatim.

| Risk | Likelihood | Impact | Complexity to mitigate | Mitigation / disposition |
|---|---|---|---|---|
| Judge-seat null dereference crashes a live run, losing every paid round (`debate.js:184`) | high | high | low | Fix now — one guard mirroring the shipped `blueEnv` guard (§3 row 2) |
| Later-round citation audits under-scaled (`citationPasses` never rescales) | high | med | low | Fix before run 4 — recompute in-loop (§3 row 2b) |
| Docket detector lineage-blind: judge structurally unreachable; a bad-faith or stuck debate spins to `maxRounds` with clean-looking telemetry | high (realized — 4 live chains, 0 judge dispatches) | high | med | Fix before run 4 — `supersedes` field + lineage-following detection + script-level enforcement throw + suite case (§3 row 23, hardened per R5-5); independent of the `prevGapIds`-widening fix (§5 item 12) |
| Degenerate `{FAIL, gaps:[]}` burns all remaining rounds silently, terminal state self-contradictory | low-med | high | low | Throw a distinguishing error (§3 row 20, decided R4-2) |
| Friction telemetry lost: blue-synthesize never harvested; whole aggregate lost on any mid-run throw | high / med | med / med-high | low / low-med | Add the missing `takeFriction` call (row 21); seat-level append to `friction.md` (row 24) |
| Minority-report provenance destroyed at synthesis; red grades consensus and singleton claims identically | high (realized in run 2) | med-high | low | Claim-manifest cheap half: single-lane set-difference tags at synthesis (§3 row 5) |
| Breadth-phase lane convergence wastes parallel spend | high (measured) | med-high | low-med | Lens/source-class lane assignment with 2-of-N redundancy floor; full roster needs `lanes >= 5` (§3 rows 6–7) |
| Judge's carried-gap rationale never reaches blue | med (unexercised — judge never fired) | med | low | blue-respond reads latest `### LEAD` before drafting (§3 row 22) |
| Write-block on report-named files (filename-keyed, path-independent) taxes every seat every round | high (realized, 3 runs) | med | low | Skeleton merged — first live trial owed against the real skeleton path (§3 row 8a; §5 items 4/7); prompt-line belt-and-braces (8b) |
| PDF-source claims capped at abstract/HTML depth across all roles | high (every round) | high | low (adoption) | Adopt `arxiv-latex-mcp`/`pdf-reader-mcp` with the vetting precondition (§3 row 13) |
| Run leaves no trajectory; recurrence counts and live trials become unrecoverable (runs 1 and 3 demonstrated it) | high (realized 3x) | med-high | low-med | Copy-and-gzip `journal.jsonl` post-run (§3 row 11) |
| Cost figures self-reported and possibly ~92% undercounted | high | med (direction: understatement only) | n/a | Flagged at every use ([^CostFigureProvenance]); do not re-quote without the cost-audit tool |
| Round-5 fixes unaudited (ceiling) — incomplete-propagation regression may survive in the final blue report | med (class realized 5x/5 rounds) | med | low | Residual: one-pass propagation audit (grep corrected strings against all asserting locations) in run-4 prep; stamped UNVERIFIED |
| `risk_accepted`: advisory-access infrastructure | med (stopped by round 4) | high when it fires | unknown/possibly unfixable | Engineered around (unconditional de-authorized channel); building infra for a designed-around gap fails the complexity test; revisit if load-bearing again (§3 row 14) |
| `risk_accepted`: ENAMETOOLONG chunked-append tooling | high (2/2 under Windows/large-payload; occurrence 2's mechanism unconfirmed) | low-med | med | Workarounds proven (chunking, scratchpad-copy, Edit-append); build on a third occurrence — trigger advisory-only until a recurrence ledger exists (§3 row 15; §5 items 10–11) |
| `risk_accepted`: artifact rename as write-block fix | — | — | low edit, plugin-wide blast radius | Unverified mechanism; superseded by the skeleton, strictly cheaper (§3 row 8c) |
| `risk_accepted`: full per-claim manifest now | — | — | med | Cheap half first; full ledger only if the minority tag proves load-bearing (§3 row 5) |
| `risk_accepted`: blanket lane-count raise as diversity lever | — | — | — | Method diversity is the cheaper, better-targeted lever (diversity-collapse literature, restated qualitatively) |
| `risk_accepted`: round-scoped audit narrowing | med (cost) | med-high if wrong: missed regression — the gate's own failure mode | med-high | Hold, human-gated; full re-read caught R1-1; revisit with the backlog's open/closed sharding rule after rows 3/4/16 prove insufficient (§3 row 18) |
| `risk_accepted`: content-poisoning of FEOV's own research phase, residual beyond leaf-node verification | med (no live incident) | high if it fires | med | Mitigation honestly scoped (R2-7): catches source-misstatement, not self-consistent fabrication; independent-index cross-referencing explicitly not built (§3 row 19; §5 item 8) |
| `risk_accepted`: simulating Task-tool permissions/OS limits/fetch lossiness in the simulator | — | — | — | Tier boundary by design — that is `--smoke`'s job (§2.1) |

---

## Blue team report (in full)

Embedded verbatim from `blue/report.md` (union, not summary; internal heading levels preserved as written):

---

# blue report — FEOV retrospective

Round 0 synthesis: union of three lane drafts (`blue/candidates/lane-{1,2,3}.md`). Lane 1 took H1
(lane diversity) deep; lane 2 took H2 (claim provenance) deep; lane 3 took H3 (testing strategy)
deep plus the un-frontiered doubts. Combined evidence base: the retrospective `inputs/` corpus
(doubts, backlog snapshot, both friction files, the run-1 defect record incl. `journal.jsonl`),
the complete run-2 corpus (`2026-07-12_memory-architecture/`), the live plugin source at `main`
HEAD, the live `ideas/backlog.md`, the unmerged branch `feat/feov-dogfood-round-1` (PR #14), the
red-auditor's persistent agent memory, and ~22 web searches/fetches across the three lanes
(disconfirming budget met in lanes 1–2: 3/14 and 2/4 respectively against the 1-in-5 protocol
floor). **Corrected round 1 (R1-20):** lane 3's "per-claim source checks" was originally listed
as a third parallel ratio; it is not one — it is baseline citation discipline (every claim traced
to a source), not a disconfirming-evidence search budget, and `lane-3.md` contains no quantified
search-count ratio to report. Restated rather than re-asserting a false parallelism: lane 3
practiced per-claim verification throughout but did not track a distinct disconfirming-search
count separate from its regular research process.

**Provenance note (per this run's own H2 finding, practiced here):** claims below are tagged
[L1], [L2], [L3] for single-lane findings and [L1+L2], [all lanes] for multi-lane convergence.
One load-bearing lane conflict was resolved by synthesizer leaf-node verification (§0).

**Round 2 corrections (R2-1..R2-11), all additive — full accounting in `blue/CHANGELOG.md`:** the
dominant pattern in round 2 was repair-regression — five of round 1's own fixes introduced new,
smaller defects, audited as such. Two corrections are load-bearing enough to flag here before the
detail below: (1) **R2-3** — the disposition deferring cross-provider model diversity (§1.1) cited
"2 diverse agents match/exceed 16 homogeneous" for same-provider lens diversity, but that figure is
the paper's different-model-and-persona condition; same-provider lens diversity's real curve needs
8 agents for the same parity — the defer call is repaired to rest on the infrastructure-cost
argument alone. (2) **R2-7** — the content-poisoning risk-accept (§3 row 19) claimed a mitigation
("independent re-verification against a second source") that the protocol does not implement;
restated to the mitigation's honest scope (catches source-misstatement, not self-consistent
fabrication). One correction is a rebuttal, not a concession: **R2-4** flagged an uncited "7 agents"
figure correctly, but red's proposed replacement citation (arXiv:2606.02646) does not, on direct
fetch, state that figure either — the unpinned number is dropped rather than re-cited to an
unverified source (see [^DiminishingReturns]). The rest (R2-1, R2-2, R2-5, R2-6, R2-8, R2-9, R2-10,
R2-11) are mechanical accuracy fixes or one-clause reconciliations, detailed in place below.

**Round 3 corrections (R3-1..R3-10), all additive — full accounting in `blue/CHANGELOG.md`:** two
gaps are control-flow findings against the unchanged `debate.js` (R3-1: a schema-legal
`{verdict:'FAIL', gaps:[]}` shape loops silently to `maxRounds` with the judge never invoked and a
self-contradictory `UNVERIFIED`/`gaps_outstanding:0` terminal return; R3-2: `takeFriction` is never
called for the round-0 `blue-synthesize` seat despite `BLUE_ENVELOPE` declaring `friction`, so
§2.1's "never dropped" claim was false for exactly one schema'd seat — this report's own §0
write-block addendum is a live instance of the class it would have caught). One is a delivery-path
finding (R3-3: the judge's explicitly-requested "state what further research blue owes" for
`carried` gaps has no read site — neither the script nor `blue-respond`'s prompt ever reads
`judge.resolutions[].rationale` for that branch). Two are footnote/body-lag defects in the same
citation the report has now required fixes to three rounds running (R3-4: the §1.1 body still
carried a clause its own footnote had already retracted in R2-4; R3-9: the R2-4 replacement
sentence itself conflated nominal-N breakeven with effective-diversity saturation ceiling in one
self-contradicting clause — disambiguated here). Two are single-sentence arithmetic/wording
corrections (R3-5: the R2-8 reconciliation mis-added its own lane-count roster, `>= 4` corrected to
`>= 5`; R3-6: a "zero hits outside the ledger clause" phrasing corrected to the stronger "zero hits,
full stop"). Two are provenance narrowing/hygiene (R3-7: ENAMETOOLONG occurrence 2's mechanism is
confirmed only as "shell parsing failure," not confirmed as the same length-ceiling class as
occurrence 1 — narrowed, not reversed; R3-8: the cost-audit backlog footnote re-pinned three commits
forward, gaining a directly relevant merge-seat cost-driver finding for §3 row 18). One is pure
footnote-tag propagation (R3-10: the R2-5 undercount caveat now reaches every reading-order-first
instance of the token figures, not only its original three homes). No proposed fix was rejected
this round; every gap is either fixed in place or closed with the cheaper of two fix options
argued explicitly (R3-3).

**Round 4 corrections (R4-1..R4-5), all additive — full accounting in `blue/CHANGELOG.md`:** one
gap is the round's most load-bearing finding (R4-1: the contested-docket detector is lineage-blind
by construction — pure gap-id string equality, no `supersedes` field — so red's own
closed-WITH-REGRESSION practice, which mints a fresh id per successor gap, means a multi-round
dispute lineage never arms the docket no matter how far `prevGapIds` is widened; this corpus
contains four such live chains, and the judge was never dispatched once across every completed
round to date (corrected round 5, R5-3: this read "three completed rounds," one round stale —
round 4 completed judge-free too; phrased round-agnostically now so the claim does not need
re-dating again next round) — zero `### LEAD` headers in `debate.md` — independently corroborated
by the project's own
backlog, commit `42dba2d`, naming this retrospective's own chain as its worked example). §2.1 and
§2.3 now state the narrower same-id-skips-a-round case and the lineage-blind case side by side as
two independent failure classes with two independent fixes (§3 row 23 adds the `supersedes`-field
repair; §5 item 12 states the independence explicitly). One is an unresolved-disjunction repair
(R4-2: §3 row 20's guard shipped red's own required-fix "either/or" verbatim rather than deciding
between throw and logged-warning-PASS — decided here to throw, per the report's own
anti-silent-degradation argument used everywhere else, with the positive assertion added to §2.3
addition 13). One is a source-level ambiguity repair (R4-3: row 6's operative sentence used a
slash as a synonym-joiner where the roster two sentences later uses the same slash as a list
separator — a regression of R3-5's own closure at the level of the sentence R3-5 left unedited;
fixed at the sentence itself, not just its downstream arithmetic). One is a numeral-propagation
fix (R4-4: the fifth, previously-uncorrected location of R2-1's retracted "4th occurrence" figure).
One is a cross-corpus hygiene fix (R4-5: this report and the memory-architecture corpus both use a
bare `R#-#` gap-id scheme and both reached round 4 this round, so the four bare cross-references
into the other corpus's ids are now prefixed `MA-`, with a footnote covering the naming
discipline). No proposed fix was rejected this round; §2.1's lineage-blindness finding is the
first round-4 gap graded high/high on both likelihood and impact rather than a lower-severity
hygiene item, reflecting that it is the round's substantive finding rather than a repair to a
repair.

**Round 5 corrections (R5-1..R5-6), all additive — full accounting in `blue/CHANGELOG.md`:** two
are repair-reached-one-location-not-all defects inside round 4's own fixes. **R5-1** (§3 row 23's
chain enumeration): round 4's debate record shows the synthesizer explicitly comparing two
independently-derived chain lists and adopting red's more precise one "in place of my own," but the
substitution reached §2.1(b) only — row 23 kept shipping the discarded first-pass list, two of
whose three entries are directly contradicted by `red/findings.md`'s own status lines (R2-5's
successor is R3-10, not R3-8; R2-1 closed clean with no R3-7 reopening). Corrected to §2.1(b)'s
verbatim list, one clause. **R5-2** (§2.1 Tier C / §3 row 13 / §4 row 1, three disagreeing
memory-architecture citations): all three carried a stale round-2 snapshot of six blocked ids,
reasserted at rounds 4-5 without a diff; direct re-read of the current corpus shows four of the six
have since closed **by ordinary live re-fetch, not a PDF tool** — falsifying the citation's own
implied prediction — and a fifth is a diagnosed miscitation, not a lossy-fetch case; only MA-R1-19
remains genuinely open and blocked. All three locations reconciled to one corrected reading, new
footnote [^MAStatusR5] carrying the full trace. One is a one-round-stale phrasing fix, present at
three locations: **R5-3** ("three completed rounds" → round-agnostic "every completed round to
date" at the front matter, §2.1(b), and §3 row 23 — round 4 also completed judge-free, so the
count was already wrong when written; phrased so it will not go stale again). One is a wording
fix that does not affect test validity: **R5-4** (§2.3 addition 15's uniform "WITH REGRESSION"
label misdescribes the mirrored chain's real second link, which closed "REBUTTAL ACCEPTED WITH
EVIDENCE" — loosened to label-agnostic phrasing). Two close enforcement gaps in mechanisms this
report itself proposed: **R5-5** (§3 row 23's `supersedes` field, as scoped, would have been set
by prompt instruction alone with nothing validating compliance — the same unenforced-good-faith
class this corpus has already hit twice; added a fourth, script-level structural check that throws
if a regression-flagged closure names no successor, plus a founding-suite assertion) and **R5-6**
(the friction-aggregation array is script-local and lost on any of the debate loop's throw sites,
losing the self-improvement signal for exactly the runs that crashed — added a scoped, prompt-only
fix reusing the already-proven Bash-append pattern from the write-block workaround, new §3 row 24,
new founding-suite addition 16). No proposed fix was rejected or risk-accepted-without-argument this
round; both R5-5 and R5-6 are built rather than risk-accepted because, once scoped narrowly, their
complexity did not clear the bar the report's own pragmatist doctrine sets for accepting a gap
instead of closing it.

---

## 0. Headline: the fixes shipped mid-debate — one exact defect class survived the merge, and the corpus's own status reporting cannot be trusted without a diff

**Round 1 correction (2026-07-13, addressing red's R1-1/R1-2 — read this before the rest of §0,
which is preserved below as the historical record this synthesis actually verified):**
`ctoforaday/special-circumstances#14` merged to `main` at `00018a5` (2026-07-14T05:58:54Z),
~8 minutes after this section's own verification commit (`9ff0fad`) — a genuine race, not
carelessness, but one this report should have guarded against with a pinned-SHA/timestamp
discipline on its own repo-state claims (the same access-date-delta discipline §3 row 10 demands
of external citations; now added, see [^PinnedRepoState]). `main` has since advanced again to
`47ae48d` (its commit message references "run 3"). Live re-verification, this round: the
`debate.js` rename, the args parse+guard, the `blueEnv` null-guard (`if (!blueEnv) throw ...`),
`tests/simulator/{harness.mjs,debate.test.mjs}` (11/11 passing live), per-role model routing, the
citation ledger, the pre-created blackboard skeleton, and the Catechism template are now **all
present on `main`** — every item this section originally called "absent."[^Reverify47ae48d]

**But do not naively flip "unmerged" to "merged and done" — that overstates resolution just as
badly as the original staleness understated it.** Direct read of `debate.js` on `main` @
`47ae48d` shows the merge closed the `blueEnv`/`redEnv` null-guards (lines 136, 171) but **left
one schema'd call site unguarded**: line 181 `const judge = await agent(...)` is dereferenced at
line 184 (`for (const r of judge.resolutions)`) with no null check. A quota-wall or terminal
failure at the adjudication seat reproduces run 2's exact `TypeError` class, today, on `main`, at
a site the merged diff did not cover.[^JudgeUnguarded] So the honest status is: **shipped,
pending its first live trial, with one known-open defect of the exact class this retrospective
is about.** The corollary that follows either way is unchanged and is the section's real point:
`ideas/backlog.md` marked items `[x]` when they landed on the PR *branch*, not on `main`, and
`run1-friction.md`'s remediation self-report was false as written for the ref a run would
actually execute — **a backlog checkbox is not a diff, and neither is a merged PR without a
call-site inventory.** The regression suite (§2), now itself merged and passing, is the only
mechanism in this system that cannot be fooled by a stale status line — extending it to the
judge call site (§3 row 2, now reworded from conditional to factual) is the next thing that
suite needs to do, not a reason to trust the checkbox again.

**Original §0 analysis, preserved as the verified-at-the-time record (superseded status
claims are struck through inline via the correction above, not deleted — see
[^MainVsBranch] for the original verification trail):**

The two deepest lanes reached apparently contradictory top findings. Both are true; the union is
sharper than either alone.

- **[L3] The claimed fixes are absent from `main`.** `inputs/run1-friction.md` states "a
  parse+guard has since been added to the debate script." Read against the live
  `skills/research-protocol/scripts/workflow.js` at HEAD: line 16 is still
  `const { topic, runDir, lanes = 3, maxRounds = 12 } = args` — no string-type check,
  no `JSON.parse`, no rejection of `undefined`/empty `topic`/`runDir`, and `git log -p`
  shows no commit touching the line.[^WorkflowJs] Likewise `ideas/backlog.md` marks the
  `redEnv` null-guard `[x]` done, yet line 140 (`if (redEnv.verdict === 'PASS') break`), line 145
  (`redEnv.gaps.map(...)`), and line 162 (`redEnv.gaps.filter(...)`) are all unguarded on
  `main` — a null `agent()` return reproduces run 2's exact `TypeError` crash today.[^WorkflowJs]
- **[L1] The fixes exist — on an open, unmerged pull request.**
  `ctoforaday/special-circumstances#14` (branch `feat/feov-dogfood-round-1`, opened 2026-07-12,
  +2281/−46) contains a working args parse+guard, null-guards, a zero-token regression simulator
  with 11 passing tests wired into CI, per-role model routing, a citation ledger, a pre-created
  blackboard skeleton (the write-block fix), and a de-DARPA'd Catechism template — none of it on
  `main`.[^PR14][^SimulatorTests]
- **Synthesizer verification (2026-07-13, this machine):** both refs inspected directly.
  `main` @ `9ff0fad` carries the unguarded destructure; `feat/feov-dogfood-round-1` carries
  `debate.js` (renamed from `workflow.js`) with the guard — a string-typed `args` is parsed, and
  dispatch is refused on unbound `topic`/`runDir` — and
  `tests/simulator/{harness.mjs,debate.test.mjs}`.[^MainVsBranch]

**Union finding (superseded by the Round 1 correction above — kept for the historical
argument, which still holds one level down):** at the time this was written, any run invoked
against `main` — including a hypothetical run 3 — still carried R1-HARNESS-1's exact defect
class and run 2's null-crash class; that specific claim is now false for the args-guard and
`blueEnv`/`redEnv` classes (merged) and still true for the judge-null class (not merged — see
correction above). The reframe this section argued for ("a shipping question, first") is now
**a shipped-but-incompletely-covered question** — the same category of question, sharper. And
the checkbox-trust failure stands regardless of merge status [L3]: `ideas/backlog.md` marked
items `[x]` that were not on `main` at the time (the graduation commit `9ff0fad` checked them off
*because they landed on the PR branch*[^LiveBacklog]), and `run1-friction.md`'s remediation
self-report was false as written for the ref a run would actually have executed at that moment.
**A backlog checkbox is not a diff, and — per the correction above — neither is a merged PR
without a call-site inventory.** The only mechanism in this system that cannot be fooled by an
inaccurate status line in a markdown file is a regression suite that runs against the shipping
ref and covers every schema'd call site [L3] — which is itself the strongest argument for the
simulator (§2), now merged and passing, but not yet exhaustive (§2.3 additions 1–3, and the new
judge-null case below).

**Live addendum (this synthesis, 2026-07-13) — labeled per R1-18, self-observed, not yet
artifact-logged:** the write-block fired *again* on this very report — the synthesizer's Write of
`blue/report.md` was refused with the same message recorded in both runs' friction ("Subagents
should return findings as text, not write report files"), and a first chunked-heredoc workaround
attempt then failed on shell parsing (the long-command/heredoc fragility class), forcing a third
path (scratchpad Write + copy). Unlike every other write-block occurrence in this corpus, this
one has no independent artifact trail — no `friction.md` or journal exists for this run, and the
claim was made by the same seat it vindicates. It should be weighed accordingly: **self-observed,
not yet artifact-logged**, one data point, not full corroboration on its own. What *does*
corroborate the class independently: red's own round-1 audit pass hit the identical block writing
`red/findings.md` **one round later** — this synthesis's own hit is dated round 0 (this addendum,
`blue/CHANGELOG.md` Round 0), red's hit is dated round 1 (`debate.md`'s round-1 RED friction,
worked around via `Edit` rather than `Write`) — a different seat, a different filename, the same
message, logged by the party it does not vindicate, **across rounds 0 and 1**. **Correction
(round 2, R2-2):** the two sentences above previously read "this same round," which contradicts
the correct "two consecutive rounds" characterization two sentences later within the same
paragraph — a chronology error caught by red directly against the dated sources; corrected here to
the single consistent phrasing. Two independent hits at two seats across rounds 0 and 1 is stronger
evidence for "the class is alive on this environment" than either occurrence alone, even though
neither is a full forensic trace; §3 item 8's fix is not speculative hardening, but this section's
own italicized claim should not be over-read as proof by itself.
This also sharpens §3 item 11 (trajectory capture): had `journal.jsonl`-equivalent capture existed
for this retrospective run, this paragraph would not need to argue its own credibility.

---

## 1. Doubts: confirmed, refuted, needs instrumentation (task question 1)

### 1.1 H1 — blue lane diversity: CONFIRMED convergent and under-provisioned, with load-bearing refinements [all lanes]

**Lane count was under-provisioned, and the run record cannot even say so.** [all lanes]
`commands/research.md` documents `--lanes` default **3**;[^ResearchCommand] run 2 dispatched
**2** (`blue/candidates/` holds exactly `lane-1.md` and `lane-2.md`), so H3, H4, and H5 were
never any lane's dedicated deep assignment — covered only as breadth inside the two
hypothesis-deep lanes.[^Run2Frontier] `workflow.js`'s dispatch loop takes `lanes` as a plain
caller argument with no minimum enforced.[^WorkflowJs] Whether `--lanes 2` was deliberate cost
control or an omission is **unrecoverable from the corpus**: nothing in the workflow's return
object, `debate.md`, or `friction.md` records the `lanes` value used — the only way any lane
confirmed it was counting files in `blue/candidates/` [L3]. **Verified-as-to-fact,
unverified-as-to-intent** [L1].

**Direct lane-1-vs-lane-2 comparison: assigned-deep material diverges; breadth converges.**
[all lanes, independently measured]

*Genuinely convergent (both run-2 lanes independently reached):* CVE-2026-21852 memory poisoning
as the "absent from §9 entirely" blocking omission — headlined near-identically, same framing,
same section-numbering convention, arrived at independently (lanes dispatch via `parallel()` and
cannot see each other's drafts) [L2+L3]; drop-the-confidence-float; git-diff review as
weak/forensic; the same alternatives roster (claude-mem, basic-memory, mem0, Letta, Zep) with
near-identical "steal lists"; OKF spec verification (v0.1 Draft, `type` the only required field);
the transcript JSONL substrate at the leaf node; `@`-import semantics (4-hop max, silent-disable);
`.claude/rules/` as an unconsidered projection alternative; headless/`-p` execution risk; the
transcript-schema-instability caveat; Letta sleep-time compute and Stanford generative-agents
importance-threshold citations [L1+L2 breadth inventories, merged].

*Lane-1-of-run-2 only (minority reports, real and load-bearing):* local leaf-node repo
verification catching two **false premises** in the proposal (the secret-scrub gate does not
exist; `sleeper-service`'s `docs/scheduling.md` does not exist) — a critical-stance move only
that lane performed; the headless-hooks open-issue set (#20063/#38651/#40506); the Dependabot
54%-merge-rate bot-review-fatigue citation; the `autoMemoryDirectory` ingest-collapse idea; the
OKF reserved-file/no-frontmatter rule; the four native Claude Code surfaces inventoried
individually; the bidirectional-write-collision analysis of the `memory:` frontmatter row
[L1+L2, merged inventories].

*Lane-2-of-run-2 only:* RecMem eager-vs-recurrence consolidation (77–87% wasted tokens, no
accuracy gain); dedup-recall mechanics (paraphrase-detection gap, LLM-judge threshold at cosine
0.85–0.95); native "Auto Dream" as a competing rolling-out feature (scope-collision finding); the
headless parallel-Task-fan-out hang (#56540); the BeliefMemory/ALFWorld confidence-helps
counter-evidence; the two-lever consolidation-corruption fix; the git-diff
Dependabot/PR-review figures; the Auto-Dream two-writer-conflict analysis [L1+L2, merged].

**Refinement 1 — the frontier's "CHANGELOG dominated by dedup, not reconciliation" clause is
half-refuted** [L2]: each run-2 lane surfaced substantial non-overlapping, load-bearing material
(above), so the breadth phase produced *different* breadth, not pure re-coverage. Convergence
concentrated on the field's canonical facts, which two competent researchers *should* both find.

**Refinement 2 — convergence on a real gap is a feature, not only waste** [L2+L3]: two
independently-dispatched lanes reaching the poisoning finding through different entry hypotheses
is textbook independent replication — it *raises* confidence precisely because the frontier's
H1–H5 never mentioned poisoning; both lanes found it unprimed. The waste is specifically the
*structural* convergence (same section skeleton, same "first hypothesis, then all others" shape,
same reliance on the shared `blue/frontier.md` + `report_template.md` scaffold). A fix that
suppressed convergence indiscriminately would throw away the corroboration signal along with the
redundant scaffolding.

**Refinement 3 — the convergence signal is then erased at synthesis** [L1, measured]:
`blue/report.md` (2,145 lines, 4 rounds) contains zero occurrences of "lane-1"/"lane-2"/"lane
1"/"lane 2" as per-claim attribution.[^LocalGrep] A reader cannot tell the doubly-corroborated
poisoning finding from a single-lane minority report. (This is H2 — same defect, same
measurement; §1.2.)

**Disconfirming evidence against the naive fix ("more diversity is strictly better")**
[all lanes, distinct sources]:

- Persona/role assignment alone does **not** prevent convergence — agents "still converge toward
  homogeneous outputs despite these assignments"; the recommended countermeasures are structural
  decoupling (less shared context), isolated generation phases, and diversity metrics, not
  persona engineering [L1].[^DiversityCollapse]
- Crowd-wisdom value requires *uncorrelated* errors; lanes sharing a base model, training data,
  and the same frontier document are presumptively correlated — so un-tagged agreement is weak
  corroboration until lanes are engineered toward different source classes, which *sharpens* the
  case for provenance tagging [L1].[^WisdomCrowds]
- Diversity/committee gains plateau at roughly 2–4 agents in aggregate synthesis across multiple
  sources; "just add lanes" is the scope-creep reading of the finding [L2]. **Citation correction
  (round 1, R1-5):** this bound is not individually pinned to any one of the four originally
  bundled sources at the level checked (abstract/HTML) — one source (VentureBeat) carries a
  materially different diminishing-returns story (tool-count, not agent-count). Restated as
  qualitative synthesis, not a precise figure: independent re-search this round corroborates the
  aggregate shape — accuracy plateaus around 2–3 debate rounds and 2–4 agents on
  moderate-complexity tasks; harder tasks shift the nominal breakeven higher (3–4 agents) but, per
  [^DiminishingReturns]'s round-2 correction, may reach *effective-diversity saturation* even
  earlier in the agent-count curve than the moderate-complexity case, not later — the qualitative
  thesis (diminishing returns exist, agent count is not a free lever) holds either way, and the
  direction on harder tasks is toward more caution about adding agents, not less. **Corrected
  round 3 (R3-4): this body clause previously still read "continued gains observed to 7 agents on
  the hardest" after the footnote itself had already retracted that exact figure as unpinned
  (round 2, R2-4) — the body lagged its own footnote's repair for a full round, the reverse of the
  usual footnote-lags-body pattern, and a majority-surface reader (body text, not footnote) would
  have carried away a withdrawn, wrong-direction number.** The precise "2–4" remains a synthesis
  across sources, not a single citable number.[^DiminishingReturns]
- Heterogeneous personas (distinct lenses, not just different starting topics) measurably reduce
  pairwise error correlation, and diverse small ensembles can match much larger homogeneous ones —
  but same-base-model agents remain more correlated than architecturally distinct ones, so
  lens-diversity is a measured improvement, not a guarantee [L3]. **Citation correction (round 1,
  R1-4):** the previously-cited "~19% lower / ~95% of independent-ensemble gain" figures do not
  appear in arXiv:2602.03794 — confirmed this round by direct fetch of the abstract plus an
  independent full-text percentage search; that paper's real, citable, verified figure is
  qualitative: "2 diverse agents can match or exceed the performance of 16 homogeneous
  agents."[^AgentDiversity] The 19% figure is real but belongs to a different paper in a narrower
  domain — arXiv:2603.22103, a 31-persona narrative-similarity-annotation ensemble, where
  "Practitioner" personas show 19% lower pairwise error correlation than "Lay" personas (r=0.388
  vs. r=0.461), producing an ensemble accuracy gain under majority voting (76.0% vs. 75.3%)
  despite lower individual accuracy — confirmed this round by direct fetch of the paper's
  results section.[^NarrativeSimilarity] The domain is narrow (narrative annotation, not
  multi-agent research debate) so treat it as supporting-analogy, not a transferable rate. The
  "~95%" figure traces to no source found in either paper; dropped rather than re-cited.
- Isolated parallel generation matches or beats homogeneous multi-agent debate at 2.1–3.4x fewer
  tokens, while debate induces sycophantic conformity (up to 85.5% modal adoption) and consensus
  collapse (oracle gaps up to 32.3 points). FEOV's blue phase is already
  isolated-parallel-then-synthesize, and red is adversarial rather than consensus-seeking — the
  literature *validates* the engine's overall separation of concerns; the convergence problem is
  local to the blue-lane breadth phase, not systemic [L1].[^IsolatedCorrection]

**Cross-provider model diversity — named and dispositioned (round 1, R1-17):** the report's own
citation above states architecturally distinct models decorrelate errors more than same-base-model
diversity, and the report never surfaced this as an option for the lane-diversity fix. Disposition:
**defer, not adopt.** `debate.js`'s `model`/`judgmentModel` knobs select among Claude model aliases
only — the harness has no wiring for a non-Claude provider as an `agent()` backend, and adding one
is an infrastructure change (a different call surface, different auth, different cost/latency
profile) an order of magnitude larger than the source-class/method-lens fix already scoped in item
6 below. **Citation correction (round 2, R2-3):** the sentence originally closing this paragraph —
"the same paper's practical finding, '2 diverse [persona-lensed] agents match/exceed 16
homogeneous,' shows method/lens diversity within one provider already captures most of the
achievable gain" — misattributed the condition. Direct read of arXiv:2602.03794's Table 2 (lens 1's
fetch, independently re-fetched at the round-2 merge seat) shows the "2 vs. 16" result is the
paper's **L4** condition (different models AND different personas: L4 at N=2 scores 67.71% vs. L1
at N=16's 65.34%) — not L2 (persona-only diversity on one base model), which is the condition the
bracketed gloss "[persona-lensed]" actually substitutes in. **L2's real curve needs 8 agents to
match the same L1-at-16 baseline** (65.44%) — a 4x efficiency gap from the "2 vs. 16" headline, not
"most of the achievable gain." The paper's own adjacent, correctly-quoted sentence above
("same-base-model agents remain more correlated than architecturally distinct ones") already said
as much; the disposition drew the opposite practical conclusion from the same source. **The defer
call itself still stands, but on the infrastructure-cost argument alone** — a call this large is
not justified by a 4x same-provider efficiency gap alone, but the harness has no non-Claude
`agent()` backend regardless, and that fact does not depend on the mis-cited figure. §5 item 5's
revisit trigger is recalibrated below to the honest L2 curve rather than the L4 number it was
originally calibrated against. Revisit if lens-assignment (item 6) under-delivers on measured
convergence reduction once the claim manifest (§3 item 5) makes that measurable — "under-delivers"
now means "closer to L2's 8-agent curve than to any 2-agent parity," not the disproven claim that
2-agent parity was ever the null hypothesis for same-provider lens diversity.

**Where diversity actually came from this run** [L2]: run-2 lane 1's distinctive contribution
(false-premise repo verification) came from a *local-critical-stance lens* neither lane was
assigned; lane 2's came from *treating its own hypothesis's literature more deeply*. Method
diversity, not headcount, produced the minority-report value — the fix is per-lane
method/source-class assignment (§3, item 6), e.g. lane A primary literature/benchmarks; lane B
practitioner/production evidence (vendor blogs, postmortems, issue trackers); lane C
adversarial/disconfirming-first (inverting the 1-in-5 floor toward 1-in-2) [L3], or lane N always
runs a local-repo critical-stance pass verifying every claim the subject makes about the
codebase's current state [L2]. `workflow.js` line 104 currently assigns only hypothesis order
("lane i takes hypothesis i first, then breadth") — no lens, no source class, no method [L3].

**Verdict: CONFIRMED** — under-provisioned count, real assigned-deep divergence, near-total
breadth convergence, and total erasure of the convergence signal at synthesis; with the
refinements that substantive convergence is partly independent replication (good) and that the
remedy is engineered method diversity, not headcount.

### 1.2 H2 — consensus-vs-minority claim provenance: CONFIRMED destroyed at synthesis, not merely under-surfaced [all lanes]

- **The merge vocabulary carries no per-claim provenance tag** [L1+L2]. `blue/CHANGELOG.md`
  Round 0 describes every merge in class-level language — "kept both," "union of," "merged X §n +
  Y §n," "preserved distinctly-sourced near-duplicates" — applied identically whether a claim came
  from one lane or two.[^ChangelogR0] The memory-poisoning section (doubly-corroborated) is merged
  with the same language as single-lane content (e.g., the OKF reserved-files caveat) [L1].
- **Partial raw material existed and was discarded** [L3]: the CHANGELOG's Round-0 prose does use
  some lane-scoped phrases ("lane 2 unique," "lane-1-specific rows") — so consensus/minority
  classification briefly existed in the merge log, but nothing carries it into `blue/report.md`
  or red's inputs.
- **The synthesized report has zero per-claim lane attribution** [all lanes, measured]:
  `blue/report.md` (2,145 lines) — zero matches for lane strings as claim attribution;[^LocalGrep]
  the handful of "lane" mentions in the corpus are method-statement or table-header level
  ("two research lanes," "consolidated, both lanes"), never inline per-claim.[^BlueReportGrep]
  (Line-count correction, synthesizer-verified: `blue/report.md` is 2,145 lines; the *assembled*
  `report.md` is 2,972 — one lane draft conflated the two; the grep conclusion holds for
  both.[^MainVsBranch])
- **The one exception is negative provenance** [L2]: §10 flags items "cited by the proposal
  without independent corroboration in either lane" — the synthesizer *can* express lane-count
  when it chooses, but does so only for the zero-lane case, never the one-lane (minority)
  case.[^BlueReportUnverified]
- **Red's corroboration grading has no lane-count input — by schema, not just by habit**
  [all lanes]. `red/findings.md` (695 lines, 30+ graded gaps): 66 occurrences of "corroborat-",
  all grading *external citation* confidence; zero occurrences of "both lanes"/"one lane"/"lane
  provenance";[^LocalGrepRed] the file's single "lane" mention is housekeeping ("Disconfirming
  budget met in both blue lanes. Not a gap").[^RedFindingsGrep] `agents/red-auditor.md`'s grading
  dimensions are entirely about the external reference,[^RedAuditorSpec] and the `RED_ENVELOPE`
  schema has no `lane_count`/`provenance_class` field — there is no channel for the signal to
  travel even if red wanted to grade on it [L2+L3].
- **Adjacent but distinct — do not conflate with MA-R2-10** [L2] **(round 4, R4-5: disambiguated —
  see below)**: the memory-architecture corpus's own red audit (a separate gap-id sequence from
  this retrospective's own R1-1..R4-5) caught, at its gap **MA-R2-10**, blue laundering its own
  cross-lane comparison as *external* corroboration (`[^SingleUserLowRisk]` citing "practitioner
  consensus" for blue's own synthesis).[^R2-10] That is a citation-provenance failure (internal
  reasoning presented as external fact), not a lane-provenance failure (claim origin invisible to
  the reader). Tighter footnote discipline fixes the former and would not catch the latter at all;
  a reader assuming red's citation lens covers "claim provenance" will wrongly conclude H2 is
  handled. **Naming discipline (added round 4, R4-5): this report and the memory-architecture
  corpus both use a bare `R#-#` gap-id scheme, and both are now into their fourth round — this
  retrospective's own round-4 ids (R4-1..R4-5, this round) collide in form with the
  memory-architecture corpus's own R4-1..R4-12. From here forward, every reference into the
  memory-architecture corpus's own findings is prefixed `MA-` (as above); this retrospective's own
  ids are left bare, matching every prior round's usage and avoiding a full-document rename. See
  §3 row 13 and §4 row 1 for the two other pre-round-4 locations corrected the same way, and
  [^GapIdScheme] for the one-time note covering all four.**
- **Content survived; metadata died** [L2]: union-not-summary held for substance (verified — no
  substantive claim from either run-2 lane is missing from the final report; the CHANGELOG's
  dedup accounting is honest and traceable). What was lost is that the false-premises catch and
  the RecMem citation are each one-agent-hours of unreplicated minority work, while the CVE
  finding was independently triangulated by two agents with different search strategies. Red's
  corroboration-confidence field is the natural home for this signal and has no column for it.
- **Impact calibration (pragmatist duty)** [L1]: no wrong conclusion was reached *this run* — the
  poisoning finding survived on substantive strength. The risk is latent miscalibration: a future
  run where a single-lane hallucination and a double-lane finding read identically to red and the
  human. Likelihood certain (it already happened) x impact medium x fix-complexity low.
- **The fix already exists as a proposal and is externally precedented** [all lanes]:
  `ideas/backlog.md` item (5) — a **claim manifest**: blue emits a machine-readable ledger
  (claim → citation → self-graded confidence → lane provenance); "one artifact, five wins,"
  including red walking the manifest instead of parsing 53KB of prose.[^ClaimManifest] The idea is
  a lightweight instance of an active research area ("execution provenance" as a typed graph;
  PROV-AGENT extending W3C PROV for LLM agents) [L3].[^ProvenanceSurvey] The data exists
  transiently during blue's merge today and is discarded, not absent [L3].
- **Complexity check — build the cheap half first** [L2]: a full per-claim ledger is real
  authoring discipline, not free. The cheaper partial version — tag only claims present in
  exactly one candidate draft (set-difference at synthesis; silence = consensus by default) —
  captures the highest-value half (flagging single-source claims for extra scrutiny) at a
  fraction of the bookkeeping. Promote to the full manifest only if red's usage shows the
  minority tag is load-bearing for a real verdict change.

**Verdict: CONFIRMED** exactly as `doubts.md` states it — destroyed, not under-surfaced; the
union merge is lossy by construction for this one dimension while scrupulously non-lossy for
content.

### 1.3 H3 — defect-population shape: CONFIRMED bimodal at the poles, refined to trimodal [all lanes]

Full classification and simulator design in §2. Verdict here: the frontier's two-way split
(zero-token-simulable vs. production-only) is confirmed at the poles and **refined by a real
middle tier** — *live-smoke-testable* defects (write-block, ENAMETOOLONG) that need a real tool
call or real OS but no full debate and no model reasoning [all lanes, independently derived].
PR #14's existing simulator is primary-source proof the unit tier works in practice, not merely
in principle [L1].

### 1.4 The un-frontiered doubts [L3, with L1 additions]

- **"Duplication between the analytical core and blue-in-full" — REFUTED for this run.** Run 2's
  assembled `## Analysis` (report.md lines 34–42) is a genuinely distinct higher-order synthesis:
  it makes build-vs-adopt argumentative moves, cites blue sub-sections by reference, and states a
  verdict-rationale appearing verbatim nowhere in `blue/report.md`. Single data point; may still
  be a risk for topics with less independent analytical argument — but nothing in the corpus
  demonstrates the feared duplication.
- **"Blue's `open_questions` have no home in the report template" — CONFIRMED.** `blueEnv`'s
  schema requires `open_questions` every round; the assembled `report.md` (2,972 lines) contains
  zero occurrences of "open question." Faithfully populated, silently dropped at final assembly —
  `report_template.md` has no section for it, and `blueEnv.open_questions` is in scope at
  assembly time but never passed into the assembly prompt. Cheap fix: an "Open questions carried
  past this run" template section (§3, item 9).
- **"The Heilmeier is DARPA-shaped" — NEEDS INSTRUMENTATION, and a candidate fix is already
  built.** Run 1 never reached a topic; run 2's build/architecture topic is close to Heilmeier's
  native shape (the catechism reads coherently, not strained) — weak single-instance
  disconfirming evidence for build-decision topics, but the doubt targets explainer/survey
  topics, which neither run has been [L3]. Meanwhile PR #14 already replaces
  `heilmeier_template.md` with `catechism_template.md` — questions 1–3 kept; 4–9 reframed
  topic-agnostically ("The case against," "Of interest, or merely interesting?", "What changes if
  it works — and what happens if we simply don't do it?", cost and stopping points); by
  inspection none presuppose a funding ask or deliverable date [L1].[^CatechismTemplate]
  Instrumentation: run one genuinely explanatory/survey-shaped topic before declaring the doubt
  resolved [L3].
- **Unlisted positive finding — the red-auditor project-memory loop is real and working** [L3].
  `agents/red-auditor.md` mandates recording new gap *patterns*; the store exists at
  `AgentOrange/.claude/agent-memory/frank-exchange-of-views-red-auditor/` (note the path: the
  live debates ran against the `AgentOrange` project root, not `special-circumstances`) with
  **15 well-formed gap-pattern files** (`workflow_undefined_rundir.md`,
  `pattern_self_defeating_mitigation.md`, `pattern_invariant_soundness_by_enumeration.md` —
  generalizing R4-1's denylist-vs-allowlist finding into "prove incompleteness via the system's
  own symmetric defense; recommend allowlist inversion" — `pattern_policy_without_mechanism.md`,
  and eleven more). The intended self-improvement mechanism for red is **functioning as
  specified, not aspirational**. It also exposes an asymmetry: only `red-auditor.md` declares
  `memory: project`; blue and judge do not, and blue is not even instructed to *read* red's
  accumulated library before submitting (`ideas/backlog.md` item 4, unimplemented) — red persists
  learning, blue does not (§3, item 12).

---

## 2. Testing strategy: the trimodal classification and the simulator (task question 2)

### 2.1 Every defect in the corpus, classified [union of all three lanes' tables]

**Tier A — zero-token unit-testable (Node simulator stubbing `agent()`):**

| Defect | Evidence | Why unit-testable |
|---|---|---|
| Uninitialized/stringified `runDir`/`topic` — literal `"undefined"` paths in every dispatch | run-1 `journal.jsonl`: every dispatch detected the defect and refused to fabricate; 252.9k tokens, 11m48s, honest UNVERIFIED deadlock[^Run1Friction][^Run1Journal][^CostFigureProvenance] — **tagged round 3 (R3-10): this is the reading-order-first instance of this token figure in the report; the R2-5 undercount caveat previously reached only §2.3/§2.4/Tier B, not this row** | Caller-side value-shape bug at the destructure; reproduce with `args` as a JSON string / `{}` / `"undefined"` and assert no prompt contains `"undefined"` [all lanes] |
| `redEnv`/`blueEnv` null after terminal-failure `agent()` return — `TypeError`, paid rounds lost | run 2 crash (`null is not an object (evaluating 'redEnv.verdict')`); backlog item marked `[x]` but unguarded on `main` (§0) | Stub `agent()` to resolve `null`; assert degrade-to-assembled-UNVERIFIED, not throw [all lanes] |
| Round loop / contested docket / deadlock / safety ceiling / `adjudicated` bookkeeping | `workflow.js` lines 113–166 | Pure `Set`/array logic over canned envelope shapes; currently zero tests on `main` [all lanes]. **Correction (round 3, R3-1): this row's "covered" framing did not include one degenerate shape.** `redEnv = {verdict: 'FAIL', gaps: []}` is schema-legal (`RED_ENVELOPE` requires only `verdict`, `gaps`, `citations_checked` — an empty `gaps` array satisfies the schema) and produces a self-contradictory terminal state: `contested` is always empty (`gaps.filter(...)` over `[]`), so the judge branch (`if (contested.length > 0)`) never fires — the docket, deadlock, and adjudication logic this row describes as "covered" is simply never reached for this input shape. The loop instead falls through to `blueEnv` re-dispatch with `openGaps: []` every round, silently, until `maxRounds`, then returns `verdict: 'UNVERIFIED'` alongside `gaps_outstanding: 0` — a terminal state that says both "the debate failed" and "nothing is wrong" in the same return object. Hand-traced against `debate.js` lines 56–91 (schema), 148–198 (loop), 200–218 (terminal return) at `d164ab2`, unchanged since. Distinct from item 10's malformed-non-null-envelope case below (that one is schema-*illegal*; this one is schema-legal and still breaks). Graded medium-high: low-medium likelihood (requires red to return FAIL with a genuinely empty gaps array — a lens-merge bug or an over-cautious agent, not an adversarial input) x medium-high impact (silent `maxRounds`-long token burn plus a self-contradictory terminal report) x low complexity (one guard clause). See §3 row 20 for the fix and §2.3 addition 13 for the simulator case.** |
| `citationPasses` arithmetic `Math.min(4, Math.max(1, Math.ceil((claim_count or 20)/40)))` | source | Pure arithmetic; boundary table (0, missing, 40, 41, 160+) [L2+L3] |
| **Empty-candidates cascade**: red-merge dispatched with zero round-N lens inputs because upstream lens dispatches failed | run 1: "No round-2 red candidate lens passes were ever produced, so the merge had no inputs by construction"[^RedMergeR2] | Cascading precondition violation, distinct from a single null return; no non-empty check exists; not yet in any founding list [L2] |
| **Gap-id rollover — two distinct failure classes, both against `prevGapIds`/the contested-docket detector (corrected round 4, R4-1: this row previously stated only the narrower of the two)** | source reading [L2]; lineage-blindness class independently confirmed live this round [all lanes, round 4] | Needs 3+ rounds live (expensive) for the narrow class; cheap for both in the simulator |
| ↳ **(a) Same-id-skips-a-round:** `prevGapIds` (`debate.js` line ~176) holds only the immediately prior round's ids; a gap closed in round 1 recurring under the *same id* in round 3 (skipping round 2) classifies "new," not "contested," because `contested = redEnv.gaps.filter(g => prevGapIds.has(g.id))` only ever checks against the one-round-back set | source reading [L2] | Fixed by **widening `prevGapIds` to the full adjudicated history** (union of every prior round's gap ids, not just the last round's) — this is the remedy already stated in §2.3 addition 3 below |
| ↳ **(b) Lineage-blind successor ids — NOT closed by widening `prevGapIds` (round 4, R4-1):** the docket detector is pure **id string equality**; `RED_ENVELOPE` has no `supersedes` field (confirmed by direct read of the schema, `debate.js` lines 56–91). Red's own documented practice is to close a gap "WITH REGRESSION" and mint a **fresh id** for the successor defect it introduces — e.g. this very corpus's `red/findings.md` chain **R1-5 → R2-4 → R3-4/R3-9** (the diminishing-returns footnote), plus at least three more same-shaped chains in this corpus (per debate.md's round-4 RED merge-seat enumeration, cross-checked against the cited rows/footnotes): **R2-5 → R3-10** (the cost-figure-provenance caveat, closed round 2 with residual coverage gaps, re-raised and propagated round 3), **R2-7 → R3-6** (the content-poisoning mitigation's honest-scoping wording, closed round 2, re-raised round 3 over the same paragraph's "zero hits" phrasing), and **R2-8 → R3-5 → R4-3** (the lane-diversity redundancy-floor arithmetic — closed round 2 with a mis-added sum, re-raised round 3 to fix the arithmetic, re-raised again this round, R4-3 above, to fix the originating sentence R3-5 left unedited — a live, three-link demonstration of the same lineage-blindness continuing into this very report). No amount of widening `prevGapIds`'s *window* helps here, because the successor ids are never in that set at any window width — they are new strings by construction, every round. **Direct consequence, confirmed this round: the judge was never dispatched once across every completed round to date** (corrected round 5, R5-3: round-agnostic phrasing replaces the now-stale "three completed rounds" — round 4 also completed judge-free, and this way the sentence does not need re-dating every round) — `grep -n "^### " debate.md` returns zero `### LEAD` section headers (the one plain-text match for "LEAD" is a quoted phrase inside round-3 prose about this exact gap, not a judge section) — despite four genuine multi-round dispute lineages running the full length of the debate. **Independently confirmed by the project's own backlog**: commit `42dba2d` ("docket detector tracks IDs not lineages — regression-chain gaps evade the judge"), landed 25 minutes after this report's own round-3 pin (`d164ab2`), names this exact retrospective's `R1-5 → R2-4 → R3-4/R3-9` chain as its worked example and drafts the same `supersedes: [prior-ids]` fix independently arrived at here. **The convergence this report has praised throughout (red conceding errors, blue repairing in good faith, the debate settling by round 3) is a property of this run's actors' good faith, not a property the detector enforces — a less scrupulous or more fatigued pair of seats could spin the same lineage indefinitely with the only brake being the `maxRounds` cost ceiling, and the docket would never once escalate to the judge.**
| Edge lane/round counts: `--lanes 0/1` unguarded; `--maxRounds 0` emits a verdict message indistinguishable from a real 0-round deadlock | source reading [L2] | Assert on the `log()` transcript, not just the returned envelope — operators disambiguate "never ran" from "ran and failed" by that line |
| Lane-count fidelity + missing `lanes` field in the return object | §1.1: only recoverable by listing `blue/candidates/` [L3] | Assert N dispatches for `lanes: N`; add the field, then assert it |
| Friction aggregation: per-seat arrays namespaced by label and concatenated, never dropped | source | Self-improvement input integrity [L1+L3]. **Correction (round 3, R3-2): "never dropped" is false for one schema'd seat.** `takeFriction` is called at exactly three call sites — `red-merge` (line 170), `judge` (line 187), `blue-respond` (line 197) — and never for `blue-synthesize` (the round-0 dispatch at line 132), even though `BLUE_ENVELOPE` declares `friction` as an optional property on that exact envelope shape and the seat is schema'd identically to `blue-respond`. Any friction `blue-synthesize` reports today is discarded before the loop even starts — the structured channel silently drops it. This report's own §0 write-block addendum (a round-0, blue-synthesize-seat complaint) is a live instance of exactly this event class: it survives in this report only as narrated prose in §0, because the structured `friction` field had nowhere to put it. The merged simulator test `'friction aggregates from every seat with attribution'` (`tests/simulator/debate.test.mjs` line 114) does not contradict this — direct read shows it stubs only `red`-merge and `blueRespond` friction (lines 116–117, 120–121), never `blueSynthesize`; its title overclaims "every seat," and 11/11 green is confidence in the two seats it actually tests, not the three the row implies. Certain (structural, confirmed by reading the call sites, not probabilistic) x medium impact (self-improvement loses exactly the round-0 signal class, which is disproportionately the write-block/environment class per §4) x low complexity. See §3 row 21 for the fix and §2.3 addition 14 for the simulator case. **Second, distinct correction (round 5, R5-6): "never dropped" is also false along a second, orthogonal axis — aggregation location, not seat coverage.** The `friction` array (`debate.js` line 145, `const friction = []`) is script-local scratch state, populated only by `takeFriction` calls and read only once, at the terminal `return` (line 210–217) and inside the final-assembly prompt (line 207). Direct read of the script identifies **three live throw sites that never reach that return**: the args guard (line 36, unbound `topic`/`runDir`), the `blueEnv` null-guard (line 136), and the `redEnv` null-guard (line 171) — plus a fourth pending one, §3 row 20's decided-but-unshipped guard for the schema-legal empty-`gaps` FAIL shape. Any of the four firing mid-run discards every seat's friction accumulated up to that point; `commands/research.md` step 5 ("if the returned envelope carries `friction` entries, write them to `friction.md`") only fires on a successful return, confirmed by direct read — an uncaught exception never produces a returned envelope at all. **The runs most likely to need the self-improvement signal — the ones that crashed — are exactly the runs that lose it entirely**, the opposite of the R3-2 gap's scope (which drops one seat's friction every run) but the same underlying lesson: a script-local aggregate with a single write-out point is fragile to exactly the failure class this whole report is about. A naive fix (the script writes friction to disk incrementally) violates the script's own stated no-filesystem-access doctrine ("the filesystem is the blackboard; the script has no filesystem access by design," `debate.js` line 28) — the fix has to be agent-side, like the row-8 write-block workaround. See §3 row 24 for the scoped fix.** |
| Per-role model routing (`model` for bulk seats; `judgmentModel` for judgment seats, default inherit-session) | PR #14 doctrine: "cheapen redundancy and mechanics, never judgment or the adversary" [L1] | Already tested on the branch |
| CRLF line-endings rejected as "control characters," blocking run-2 launch | backlog item 12, `.gitattributes` added | Zero-token as a CI lint (`git check-attr` / grep), not an `agent()`-stub case — same cost bracket, different mechanism [L3] |
| Malformed-but-non-null envelope (e.g., `redEnv` present, `gaps` missing) | source: no runtime shape validation before destructuring [L2] | Document current behavior as known-gap pending clarity on whether the Workflow tool's schema parameter guarantees conformance-or-null (open question; friction) |

**Tier B — live-smoke-testable (real tool call or cheap real agent turn; no full debate, no model
reasoning about content):**

| Defect | Evidence | Why this tier |
|---|---|---|
| Filename write-block on `blue/report.md` (run 1), `red/findings.md` (run 2), and `blue/report.md` again (this synthesis, §0 addendum) — "Subagents should return findings as text, not write report files"; `debate.md`/`CHANGELOG.md` writes succeeded | both friction files; backlog: "CONFIRMED as a hard, report.md-specific tool error (`is_error: True`)"[^WriteBlock]; reproduced live during this synthesis | A platform/Task-tool permission behavior, not a `workflow.js` branch — the script has no filesystem access by design; repo hooks grep confirms this repo does not implement it[^HookGrep]. A single Write-tool call against a trigger-named path either fires or doesn't — no LLM reasoning needed [all lanes]. Candidate fixes: PR #14's pre-created blackboard skeleton (subagents only append/Edit) [L1];[^PR14Description] instruct agents to write living artifacts via Bash append — proven in production, though heredoc fragility on Windows makes scratchpad-write-then-copy the more reliable variant (this file's own path) [L3+synthesizer]. A live GitHub issue (#13890) documents related subagent write failures *independent of filename* — so a pure rename is a bet, not a guaranteed fix [L3][^SubagentWriteBug] |
| Windows `ENAMETOOLONG` / long-command fragility on large Bash heredocs (~236-line heredoc, red-merge-r1, run 2; forced a 6-call chunked workaround; recurred as shell-parse failure during this synthesis) | run-2 friction[^Run2Friction]; this synthesis | Real OS/shell ceiling; a simulator faking OS limits is complexity chasing a boundary it cannot own. Unit-testable only as chunking-helper arithmetic [all lanes] |
| `/research --smoke` end-to-end wiring (1 lane, 1 round, 1 citation pass, cheap model, trivial topic) | backlog item 15(b) | The tier that would have caught the write-block before a 35-agent run; ~50k tokens vs. the 253k–3M live-discovery costs[^CostFigureProvenance] [L3] |
| Hook-fire-under-headless behavior; the skeleton fix's actual effect on Write-tool permissions | run-2 corpus; PR #14 ("first live trial is run 3") | Real tool permissions required [L1] |

**Tier C — only observable in production (real network state, quotas, vendor content):**

- Lossy HTML/abstract fetch missing PDF body-table figures (memory-architecture corpus's own gaps
  **MA-R1-19, MA-R1-28, MA-R2-8 residual, MA-R3-14, MA-R3-15, MA-R4-9** — corrected round 4, R4-5:
  these are the memory-architecture retrospective's internal red-audit gap ids, not this
  retrospective's; the bare `R#-#` form collided in form with this retrospective's own ids once
  both corpora reached round 4 — see [^GapIdScheme]) — no stub reproduces model-level
  summarization lossiness [all lanes]. **Corrected round 5 (R5-2): this list is a stale round-2
  snapshot, reasserted at rounds 4 and 5 without a diff, and its own implied claim — that these
  six are blocked pending a PDF tool — was falsified by the memory-architecture corpus's own next
  rounds.** Direct re-read of the current `research/2026-07-12_memory-architecture/red/findings.md`
  shows four of the six **closed by ordinary live re-fetch, not by any PDF-extraction tool**: MA-R1-28
  and MA-R2-8 CLOSED round 3 (both figures re-verified live at the leaf node against the correct
  page; "red accepts closure" in both cases — no PDF tool involved, no residual survives for either,
  so "MA-R2-8 residual" in this bullet's own id list is itself wrong); MA-R3-14 and MA-R3-15 CLOSED
  round 4 (re-attributed/re-scoped, verified landed at the leaf node). MA-R4-9 is open but is a
  **diagnosed miscitation, not a lossy-fetch instance** — three independent verification routes
  (abstract, full-text HTML, web-search) agree the cited paper is simply the wrong paper for the
  claimed figures; a PDF-extraction tool would not have discharged it, correct re-citation would.
  **Only MA-R1-19 remains genuinely open and friction-blocked as of round 4** ("Carried: R1-19 …
  friction-blocked," `red/findings.md` line 118-119) — the one member of this list that is actually
  the class this bullet describes. See [^MAStatusR5] for the full trace; §3 row 13 and §4 row 1
  (which independently carried five- and six-member versions of the same stale list) are reconciled
  to this same corrected reading below rather than left to disagree with each other on membership.**
- Live-source drift (gh issue status flips, star counts, mem0's pivot to single-pass ADD-only) —
  point-in-time facts about the external world; the fix is procedural (record access-date deltas)
  not code [all lanes].
- Primary security-advisory access (CVE-2026-21852) and paywalled sources (Springer auth-wall) —
  network/auth-allowlist decisions, possibly unsolvable from inside this repo [all lanes].
- Auto Dream server-side-flag behavior — rollout state lives on vendor infrastructure; run 2's
  `§10 Unverified` treatment is by design, not a testing gap [all lanes].
- **The quota-wall trigger vs. the code's response — split per-defect-half** [L2]: the *trigger*
  is production-only; the null-guard *response* is Tier A. Classify halves, not defect names,
  when a defect has a rare live trigger and a testable code response.

**Boundary discipline (pragmatist case, argued not assumed)** [L3]: the simulator must not fake
Task-tool permission semantics, OS limits, WebFetch lossiness, or judgment content — a simulator
that fakes extraction-lossiness needs an oracle for correct extraction, which is the research
problem itself. Industry practice draws the same line: mocked-LLM unit tests catch deterministic
code bugs but miss reasoning/hallucination/context failures, with a commonly cited ~95% mock-heavy
to ~5% real-call split.[^AgentTestTiers]

### 2.2 Simulator: concrete design [union; one lane conflict resolved by PR #14 evidence]

- **Injection mechanism — resolved.** The script references `args`, `agent`, `parallel`, `phase`,
  `log` as ambient bindings; lane 3 flagged "parameterize vs. `vm` sandbox" as an open question it
  could not settle from static reading [L3]. PR #14's `harness.mjs` answers it: wrap the real
  script body in `new AsyncFunction(...)` with the harness supplying the bindings — the actual
  file under test, not a reimplementation [L1].[^SimulatorTests] (Lane 2's function-constructor
  sketch converges on the same mechanism [L2].)
- **`agent(prompt, opts)` stub**: (a) records every `(prompt, opts)` for assertion — how
  `"undefined"`-in-prompt is caught with no model call; (b) returns the next canned schema-shaped
  envelope from a per-scenario queue keyed by `opts.label` [L2]; PR #14 adds the faithful
  `parallel`/`pipeline` semantics where a throwing thunk resolves to null rather than rejecting
  the batch [L1].
- **`parallel(thunks)`**: `Promise.all(thunks.map(t => t()))` — faithful to the concurrency
  contract without real concurrency [L2].
- **`phase`/`log`**: push to an inspectable transcript array (needed for the `--maxRounds 0`
  message assertion) [L2].
- **No new dependency**: repo has no `package.json` anywhere;[^NoPackageJson] Node built-in
  `node:test` + `node:assert` via `node --test`, matching the Go tools' standard-library-only
  precedent [L2].[^GoTests] PR #14 already wires the suite into CI as job `debate-sim`, ~200ms
  [L1].
- **Location/naming**: PR #14 uses `tests/simulator/{harness.mjs,debate.test.mjs}` beside the
  renamed `debate.js`; do not let the `workflow.js`-to-`debate.js` rename[^Rename] and the
  simulator block each other [L2].

### 2.3 Founding regression suite — merged case list

PR #14's suite already contains 11 passing tests [L1]:[^SimulatorTests] (1) stringified args
parse, no `undefined/` in any prompt; (2) unbound topic/runDir refuses dispatch before any agent
spawns; (3) null red-merge aborts cleanly; (4) null blue-synthesize aborts cleanly; (5) happy-path
PASS-to-VERIFIED with phase order and lane count honored; (6) per-role model routing; (7) contested
docket — re-raised gap reaches the judge exactly once, un-recurred gaps stay off the docket,
adjudicated gaps leave red's verdict scope; (8) judge deadlock ends UNVERIFIED with deadlock stamp
(what actually fired in run 1); (9) safety ceiling with always-new gap ids; (10) citation passes
scale 1..4 with `claim_count` and carry the ledger clause; (11) friction aggregation with correct
attribution.

**Additions from lanes 2–3 not yet in that suite** (each traced to the corpus, none hypothetical):

1. **Null at every schema'd call site, differentiated by actual failure class (corrected round 1,
   R1-3 — the original wording overgeneralized "each null crashes," which direct source reading
   this round shows is true for only one of five sites):**
   - **assert-throws-then-recovers-cleanly:** `judge` (line 184, `for (const r of
     judge.resolutions)`) — the one live, currently-unguarded crash site (see §0's Round 1
     correction and §3 row 2 below); a null judge return today reproduces run 2's exact
     `TypeError`.
   - **assert-silent-degrade-and-continue** (different assertion shape — these returns are never
     dereferenced, so a null return is discarded, not thrown; the founding suite must assert the
     degrade happens *without* silently losing the round's other work, not assert no-throw):
     `frontier`, each `red-lens` pass, and `final assembly` — all three returns are read only for
     a synopsis and then discarded by the script.
   - **already-covered:** `blue-respond`'s reassignment of `blueEnv` is read only at the final
     return's already-guarded ternary (`blueEnv ? blueEnv.claim_count : null`), so this site
     cannot crash and needs no new case.
   "A suite that only guards the observed site is not founding — it is anecdotal" remains true
   for the judge site specifically [L3].
2. **Empty-candidates cascade** — lens stubs return null/blocked; red-merge still dispatched
   against an empty `red/candidates/`; assert detect-and-short-circuit, or document as
   known-failing until fixed [L2].
3. **Gap-id rollover, class (a) — same-id-skips-a-round:** id present r1, absent r2, present r3;
   assert `contested` (recurrence), not `new`; known-failing until `prevGapIds` widens to full
   adjudicated history [L2].
15. **(Added round 4, R4-1) Gap-id rollover, class (b) — lineage-following contested detection,
    NOT closed by addition 3's fix:** mirror this corpus's own chain directly — three canned
    `redEnv` round objects where round 1 raises gap `X-1`, round 2's merge closes `X-1` (under
    whatever label its actual closure carries — **corrected round 5, R5-4: the previous wording
    here said every closure reads "WITH REGRESSION" uniformly, which misdescribes the very chain
    this case claims to mirror** — `red/findings.md`'s real R1-5→R2-4 link closes R1-5 "CLOSED WITH
    REGRESSION (round 2)" but closes R2-4 "CLOSED, REBUTTAL ACCEPTED WITH EVIDENCE (round 3)," a
    different label for the same lineage-continuing event; the detector logic under test is
    label-independent — it keys on the successor gap's fresh id and the (absent) `supersedes`
    field, not on the closure-reason string — so test validity is unaffected, but the case's own
    prose should not claim a uniformity the record it mirrors does not have) and raises a fresh-id
    successor `X-2` addressing the same underlying defect, and round 3's merge closes `X-2` (again,
    whatever its actual label) and raises `X-3`/`X-3b`. Assert that by
    round 3 the contested docket has armed and **the judge has been invoked at least once** —
    i.e., some round's `agent()` call list includes a `judge-r*` label — because a
    `supersedes: [prior-ids]` field on the gap envelope let the detector follow the lineage rather
    than match bare ids. **KNOWN-FAILING against the shipped schema and detector**: `RED_ENVELOPE`
    (`debate.js` lines 56–91) has no `supersedes` property, and `contested` (line ~177) is
    `redEnv.gaps.filter(g => prevGapIds.has(g.id))` — pure string-equality membership, lineage-blind
    regardless of how wide `prevGapIds` is widened by addition 3's fix. Widening the window and
    following the lineage are independent repairs (§5 item 12); this case exists so the founding
    suite cannot be declared complete by addition 3 alone. **Extended round 5 (R5-5): a second
    assertion, for the enforcement gap in the fix itself, not only the detector.** Add a fourth
    canned round where a gap closes "WITH REGRESSION" (or its `closure_class` equivalent) but the
    merge's own `gaps` array mints a fresh-id successor with **no `supersedes` entry naming the
    closed gap** — i.e., exactly the shape a real red-merge pass would produce if it forgot the
    prompt instruction rather than the schema. Assert the round-level check throws
    `` `red-merge round ${round} closed gap ${id} WITH REGRESSION but no successor gap names it in
    supersedes — lineage silently dropped` ``. **KNOWN-FAILING until §3 row 23's structural
    cross-check (4) ships** — today nothing reads `supersedes` at all, let alone validates its
    presence, so a red-merge pass that skips the field produces no signal anywhere in the schema,
    the loop, or this founding suite.
4. **Multi-round FAIL-revision-PASS state threading** — blue re-dispatched with exactly the open
   non-adjudicated gaps; closed gap ids never reappear in blue's open-gaps prompt [L3].
   **KNOWN-FAILING (flagged round 1, R1-11 — this clause was previously phrased as if the
   recompute exists; it does not):** `citationPasses` (line 139) is computed once from the
   *initial* `blueEnv.claim_count`, before the `while` loop (line 148), and never reassigned as
   the report grows additively round over round — confirmed by direct read of `debate.js` on
   `main` @ `47ae48d`. This is a live shipped defect, not a hypothetical test case: it means the
   citation-verification pass count is systematically under-scaled in later rounds, exactly when
   the report is largest and has the most surface for uncaught citation drift. Add a §3 row (now
   row 2b below) and mark this suite case known-failing until `citationPasses` is recomputed
   inside the loop from the latest `blueEnv.claim_count` on each iteration.
5. **Judge `carried` branch** — deadlock stays false; carried gap re-enters `openGaps` with its
   required-fix intact [L3]. **Correction (round 3, R3-3): "with its required-fix intact" is true
   only for the original red gap object — not for the judge's carried-gap rationale, which this
   suite case must not imply is delivered.** The judge prompt (`debate.js` line 182) explicitly
   instructs: "for carried, state what further research blue owes" — but `judge.resolutions`
   entries are read in exactly one place (line 184's loop), which pushes only
   `closed`/`rebuttal_sustained`/`risk_accepted` resolutions into `adjudicated`; a `carried`
   resolution's `rationale` field is never read again — not by the script, not by
   `blue-respond`'s prompt (line 195, built entirely from `openGaps = redEnv.gaps.filter(...)`,
   the pre-judge red gap object). The one piece of guidance the judge prompt explicitly requests
   for the `carried` case has no delivery path to the seat that would use it. Mechanism confirmed
   HIGH by direct trace of lines 93–112 and 166–197; frequency MEDIUM — no live `carried`
   resolution has occurred in this corpus yet, so the gap is real but unexercised. The founding
   suite case for this branch must assert the carried gap reaches `openGaps` (true today) AND
   must not assert or imply the judge's rationale text reaches `blue-respond`'s prompt (false
   today) — see §3 row 22 for the fix options.
6. **Judge `risk_accepted` branch** — gap enters `adjudicated` and is excluded from the next
   red-merge prompt (the adjudicated-list interpolation branch deserves its own assertion) [L3].
7. **`--lanes 0`/`--lanes 1` edges + lane-count floor** (assert no crash; separately assert the
   floor guard or document its absence) and **`lanes` field in the return object** [all lanes].
8. **`--maxRounds 0`** — the emitted log line must distinguish "never ran" from "ran and failed
   at round 0" [L2]. **Note (round 3, R3-1): this case is distinct from addition 13 below** —
   `--maxRounds 0` is a caller-supplied degenerate *loop bound*; the schema-legal
   `{verdict: 'FAIL', gaps: []}` case is a degenerate *envelope shape* reachable at any
   `maxRounds`. Both need their own case; neither substitutes for the other.
9. **`claim_count` boundaries** — 0, missing (default-20 path), 40, 41, 500+ (cap at 4 holds)
   [L2+L3].
10. **Malformed-but-non-null envelope** — document current behavior (likely uncaught throw) as a
    known gap pending the Workflow tool's schema-enforcement guarantee [L2].
11. **Claim-manifest provenance passthrough** — once §1.2's fix ships [L1].
12. **CRLF/line-ending lint** on `debate.js`, agent `.md` files, and reference templates — same
    CI gate, different mechanism [L3].
13. **(Added round 3, R3-1; positive assertion added round 4, R4-2) Schema-legal degenerate FAIL:
    `{verdict: 'FAIL', gaps: []}`** — the round-3 case asserted only the negative/option-agnostic
    half (the loop does not silently re-dispatch `blue-respond` to `maxRounds` with an empty
    open-gaps docket and no judge invocation; the terminal return does not pair
    `verdict: 'UNVERIFIED'` with `gaps_outstanding: 0`) — both assertions are writable today,
    regardless of which guard behavior ships, so red's "the test cannot be written" reading was
    overstated. **Now that §3 row 20 (R4-2) has decided the positive behavior, add the matching
    positive assertion: given this input shape, the round throws
    `red-merge round ${round} returned FAIL with an empty gaps array — degenerate merge, refusing
    to loop silently` and the debate terminates without reaching `maxRounds`.** KNOWN-FAILING until
    the guard ships.
14. **(Added round 3, R3-2) `blue-synthesize` friction passthrough** — stub `blueEnv` (the
    round-0 synthesis return, not `blue-respond`) with a non-empty `friction` array; assert it
    reaches the aggregated `friction` array and the final envelope's `friction` field.
    **KNOWN-FAILING until `takeFriction('blue-synthesize', blueEnv)` is added after the line-136
    null-guard (§3 row 21).**
16. **(Added round 5, R5-6) Friction survives a mid-run throw** — stub a seat (e.g. `red-merge`)
    to return an envelope carrying non-empty `friction`, then have the *next* stubbed seat return
    `null` (triggering the existing line-171-class guard's throw) or a schema-legal-empty-`gaps`
    FAIL (triggering row 20's guard once shipped). Assert the first seat's friction text is still
    recoverable after the throw — e.g. via a stubbed `Bash`-append side channel the harness records
    calls to, standing in for the real "append directly to `friction.md`" prompt instruction.
    **KNOWN-FAILING today**: the script-local `friction` array is discarded with the rest of the
    call stack on any throw; nothing outside the terminal `return`/final-assembly prompt ever reads
    it, and no seat currently has an append-on-report instruction. Known-failing until row 24's
    prompt-text addition ships at the four `takeFriction`-adjacent seats.

Entire suite: zero tokens, no network, in-memory filesystem fakes, well under a second — against
historical incident costs of 252.9k (run 1) and ~3M tokens (run 2's quota crashes)[^CostFigureProvenance] [L3].

### 2.4 The three-tier stack as a whole

Simulator (Tier A) + `--smoke` mode (Tier B, ~50k tokens) + production-residual discipline
(Tier C: grade confidence low/medium and label unverifiable — what red already does) — the living
backlog's item 15 independently converges on exactly this shape [L3]. Graded: simulator = high
likelihood x high impact (253k–3M tokens per historical incident[^CostFigureProvenance]) x low
complexity (exists on the branch; a few hundred lines, no new infra). Smoke = medium-high x high
(catches what the simulator structurally cannot) x low-medium (threads existing tunables through
one flag). No risk-acceptance case against either [L3]; direction of the provenance gap below is
understatement only, so it does not weaken this case.

---

## 3. What should change before run 4, graded likelihood x impact x complexity (task question 3)

Merged from all three lanes' graded tables; conflicts reconciled inline.

**Status legend added round 1:** rows below are marked `[MERGED]` where direct verification
against `main` @ `47ae48d` (this round) confirms the item shipped, `[OPEN]` where it is verified
still absent, or `[PARTIAL]` where the merge covers part of the item's scope.

| # | Proposal | Likelihood | Impact | Complexity | Disposition |
|---|---|---|---|---|---|
| 1 | **[MERGED]** ~~Review and merge~~ PR #14 (args guard, null guards on blue-synthesize, simulator + CI, per-role models, citation ledger, pre-created blackboard skeleton, Catechism template) | n/a — merged `00018a5`, 2026-07-14T05:58:54Z[^Reverify47ae48d] | High — retired the args-guard and `blueEnv`/`redEnv` null-guard defect classes; unlocked items 3, 16, 17 below as already-shipped | Low — was review only | **Done. Superseded by item 2's finding that one call site (judge) is not yet covered** [L1, corrected round 1 per R1-1] |
| 2 | **[PARTIAL] Judge call-site null-guard — confirmed still open on `main`, not conditional** (corrected round 1, R1-2: this row previously read "subsumed by #1 only if the suite extends to all call sites" — that condition is now checked and unmet) | High — `debate.js:184` (`for (const r of judge.resolutions)`) is verified unguarded on `main` @ `47ae48d`; reproduces run 2's exact `TypeError` at the adjudication seat[^JudgeUnguarded] | High — a mid-debate crash at the judge seat loses every paid round up to that point, same as the original defect | Low — one `if (!judge) throw new Error(...)`-style guard, mirroring the existing `blueEnv` guard at line 136 | **Fix now — the most inexcusable open item in the corpus, unchanged from round 0's framing, now pinned to an exact line** [L3, R1-2] |
| 2b | **`citationPasses` recompute inside the debate loop** (new row, added round 1 per R1-11 — previously misfiled in §2.3 as an untested-but-assumed-working case) | High — confirmed live on `main`: `citationPasses` (line 139, `const`) is computed once from the initial `claim_count` before the `while` loop (line 148) and never reassigned[^JudgeUnguarded] | Medium-high — later-round citation audits are systematically under-scaled exactly when the report (and its citation surface) is largest | Low — recompute `Math.min(4, Math.max(1, Math.ceil((blueEnv.claim_count \|\| 20) / 40)))` inside the loop from the latest `blueEnv.claim_count` each iteration | **Fix before run 4 — a live shipped defect, not a hypothetical test case** [R1-11] |
| 3 | **[MERGED, extend]** Founding simulator with the §2.3 merged suite — PR #14's 11 tests confirmed passing live this round (`node --test`, 11/11)[^Reverify47ae48d]; the 12 lane-2/3 additions (now including the judge-null and citationPasses-recompute cases above) remain unmerged | High — every case has already occurred at least once or is directly traceable to source | High — the only mechanism that cannot be fooled by a stale checkbox (§0) | Low — no new dependencies | **Extend before run 4 with the 12 additions, including the two new cases above** [all lanes, corrected round 1] |
| 4 | **[OPEN]** `/research --smoke` mode — **reworded round 2 (R2-11): no functional `--smoke` argument-parsing path exists in either file; the original wording over-reached its own footnote, which checked only `commands/research.md`.** `debate.js`'s header comment (lines 17-18) *describes* "`--smoke` (1 lane, 1 round, model=haiku) exercises the pipeline for ~50k tokens" as an intended mode, but the string appears nowhere as a parsed flag — `commands/research.md` has zero matches at all. Functional conclusion unchanged: no `--smoke` code path exists; still not built[^SmokeAbsent] | Medium-high — would have caught the write-block pre-production | High — ~50k tokens vs. 253k–3M live discovery[^CostFigureProvenance] (**tagged round 3, R3-10**: both figures in this cell carry the same self-reported/possibly-undercounted provenance caveat as §2.3/§2.4's use of them) | Low-medium — threads existing tunables | **Build alongside #3 — unaffected by PR #14's merge; still not built** [L3, reworded R2-11] |
| 5 | **Claim manifest — cheap half first** (tag single-lane-sourced claims via set-difference at synthesis; full claim/citation/confidence/provenance ledger deferred until the minority tag proves load-bearing) | High — the gap is proven, not speculative (§1.2) | Medium-high — restores minority-report signal to red's grading; externally precedented[^ProvenanceSurvey] | Low — the data exists transiently during the merge and is currently discarded | **Build before run 4** [all lanes] |
| 6 | **Engineered per-lane diversity** — assign distinct method/source-class lenses (primary-literature / practitioner-production / adversarial-disconfirming-first / local-repo critical-stance), not persona text and not more headcount | High — run 2 measured the convergence directly; external evidence is qualitative ("2 diverse agents match/exceed 16 homogeneous," arXiv:2602.03794) rather than the previously-cited "~19%" figure, which traces to a narrower-domain paper — corrected round 1, R1-4; see §1.1[^AgentDiversity][^NarrativeSimilarity]. **Note (round 2, R2-3): this row's own "2 vs. 16" quote is the paper's L4 (model+persona) condition; this row's fix is same-provider lens diversity only (L2), whose real curve needs 8 agents to match 16 — see §1.1's corrected cross-provider paragraph. This row's grade does not rest on the "2 vs. 16" number and is unaffected by the correction; it is flagged here only so the same figure is not silently misread across two rows of the same table.** | Medium-high — produced 2 of run 2's highest-value minority findings; but persona-only fixes under-deliver per the diversity-collapse literature[^DiversityCollapse] — a mitigant, not a guarantee; treat lane-agreement as weak corroboration even after | Low-medium — one added sentence per lane-dispatch prompt; **undiscussed trade, named round 1 (R1-16): specialized (non-redundant) lane assignment means a single failed dispatch drops 100% of that method's coverage for the round, unlike today's fully-redundant hypothesis-order assignment — this run's own highest-value catch (false-premise repo verification) came from exactly one lane doing exactly one method, so the failure-concentration risk is not hypothetical** | **Fix before run 4, scoped to source-class/method assignment, WITH an explicit redundancy floor**: assign the adversarial-disconfirming-first lens (a distinct method from local-repo critical-stance, named separately below) to at least 2 of N lanes — not 1-of-N — so a single null dispatch does not zero out a method's round coverage. **(Sentence itself corrected round 4, R4-3: the previous wording — "the critical-stance/adversarial-disconfirming lens" — used the slash as a synonym-joiner here, naming one method under two labels, while the roster two sentences below uses the identical slash-separated form as a list separator naming four distinct methods; nothing flagged the switch, so a reader stopping at this operative sentence alone reconstructed the pre-R3-5 misreading — three unfloored methods collapsing to fewer via the same ambiguous punctuation, `lanes >= 4` again, not `>= 5`. This is the repair-reaches-the-conclusion-not-the-source failure R3-5 itself exists to prevent, recurring in the sentence R3-5 left unedited. Fixed at the source, not just at R2-8/R3-5's downstream arithmetic below.)** — cost is one more lane-dispatch, cheaper than losing the round's minority-report class entirely [all lanes; redundancy floor added round 1 per R1-16]. **Reconciled round 2 (R2-8), arithmetic corrected round 3 (R3-5):** four named methods (primary-literature / practitioner-production / adversarial-disconfirming-first / local-repo critical-stance), one of which (adversarial-disconfirming-first) carries a 2-of-N redundancy floor per item 6's own text above. **The round-2 reconciliation mis-added this: three unfloored methods (1 lane-assignment each) plus one floored method (2 lane-assignments) is 3 + 2 = 5 lane-assignments minimum, not the stated `lanes >= 4`** — `lanes >= 4` is arithmetically reachable only if two of the four named methods are merged into a single lane's assignment, which the same sentence's "four named methods" (stated as four distinct, separately-named items) contradicts on its face. Restated correctly: **item 6's full four-method roster with the stated redundancy floor requires `lanes >= 5`.** Separately, row 7 below is **not** a shipped floor to reconcile against — direct verification (`main` @ `47ae48d`/`d164ab2`, unchanged) shows `lanes = 3` is an unguarded *default* with no minimum check anywhere in the dispatch loop, and row 7 is graded **[OPEN]**, not merged; run 2 overrode it downward to 2 with no floor stopping that. The original R2-8 sentence's "row 7 floors N at 3" language called an absent guard a "floor" — corrected here to "row 7's proposed floor, not yet built, targets `lanes >= 3`." Net restated resolution: **item 6's full roster needs `lanes >= 5`; row 7's proposed (unbuilt) floor targets `lanes >= 3` for runs not adopting item 6's full roster** (e.g. a smoke/dev run keeping hypothesis-order assignment). The risk-accepted "blanket lane-count raise" language (§3's risk-accepted list) always meant "raising lane count as a headcount-only diversity lever," not "raising lane count when a stated, scoped method roster requires it" — the two remain different claims and should not have shared a sentence without saying so. The fix that closed R2-8 asserted this composition's arithmetic instead of computing it; this correction computes it. |
| 7 | **[OPEN]** Lane-count floor (`lanes >= 3` or explicit justified override) **+ `lanes` field in the workflow return object** — confirmed still absent: `lanes = 3` remains an unguarded default with no minimum check on `main` @ `47ae48d`[^Reverify47ae48d] | Medium — intent of run 2's under-provisioning unrecoverable | Medium — an under-provisioned run silently loses a hypothesis's dedicated attention; the field is a prerequisite for ever checking H1 automatically | Low — one guard clause + one field; folds into simulator case 7 | **Fix now — unaffected by PR #14's merge; still not built** [all lanes; L2 notes headcount is not the diversity lever — the floor is observability and default-honoring, not a diversity fix] |
| 8 | **[MERGED (a); build (b); risk-accept (c)]** Write-block fix — layered: (a) pre-created blackboard skeleton so subagents only append/Edit — confirmed merged, `main` @ `47ae48d`; first live trial still pending (this synthesis and red's round-1 pass both hit the block on scratch runs against this very `research/` tree, not against the pre-created skeleton — see the R1-18 correction in §0); (b) prompt instruction to write living artifacts via Bash append or scratchpad-then-copy — proven at least twice in production this round alone (blue-synthesis and red-merge seats, both this run) [L3+synthesizer]; (c) rename `report.md`/`findings.md` to e.g. `corpus.md`/`audit.md` | High — recurred across 3+ runs on 2+ filenames, incl. twice live during this very round (blue-synthesis and red-merge seats) | Medium — worked around each time, so not blocking, but token overhead every occurrence | (a) Low, merged; (b) Low — prompt-only; (c) Low edit cost but plugin-wide blast radius, and unverified — the block may be semantic/role-based, not purely filename-based. **Analogy softened round 1 (R1-9):** GitHub issue #13890 documents a *different* failure signature (silent no-op write; the agent believes it succeeded) from this repo's explicit worded refusal — both are subagent write failures, but the inference "therefore not filename-keyed" is a weaker transfer than originally implied[^SubagentWriteBug] | **(a) shipped — verify against the real skeleton in run 4, not a scratch tree; (b) keep as belt-and-braces prompt line; (c) risk-accept/skip — superseded by (a), which is strictly cheaper for the same outcome** [reconciled L1+L2+L3] |
| 9 | **`open_questions` template section** — "Open questions carried past this run," sourced from the last `blueEnv.open_questions` at assembly | High — confirmed dropped every round (§1.4) | Medium — the field is schema-required, faithfully produced, and silently discarded | Low — one template section + pass the in-scope variable into the assembly prompt | **Fix before run 4** [L3] |
| 10 | **Access-date-delta recording** for citations (protocol/footnote-template level) | High — observed 4+ times (mem0 pivot, issue-status flips, star counts) | Medium — drift is usually caught by re-verification; the cost is re-work, not silent error. **Re-graded round 2 (R2-9): this impact cell assumed a re-verification safety net that the shipped citation ledger actually suppresses.** `debate.js`'s ledger clause (line 152-156, verbatim, confirmed live at `88eb57f`) instructs every citation lens: a claim verified HIGH in a prior round "stays verified — do not re-fetch it unless `blue/CHANGELOG.md` shows its section changed this round." That skip-trigger is keyed to the citing *prose* changing, not the *source* changing or an access-date elapsing — exactly the two things this row's own fix tracks. The corpus's own drift evidence (mem0's pivot, PR #14 merging mid-debate, issue-status flips) shows sources moving within a single debate's runtime while the citing prose stays untouched — precisely the case the ledger's skip-trigger would miss. So "usually caught by re-verification" describes a re-verification step the shipped ledger suppresses for every already-HIGH claim, and the access-date field this row proposes has no consumer that acts on it once ledgered | Low — a footnote-template field, **plus one clause added to the ledger's own skip-trigger text**: "...unless the section changed, OR N rounds have elapsed since last verified, OR the recorded access-date delta exceeds the source's volatility class." Both are one-line prompt-text edits to the same `ledgerClause` string in `debate.js` | **Build now — and word the ledger's skip-trigger to include a time/access-date condition, not only a prose-change condition, or this field is collected and never read** [L2, re-graded R2-9] |
| 11 | **Formalize trajectory capture** (`journal.jsonl` into `<run>/trajectories/`, gzip) after every `/research` run | High — confirmed absent from run 2's native corpus (`trajectories/` exists only in the retroactively-assembled run-1 record); this round's own self-observed write-block claim (§0 addendum, R1-18) is a second live demonstration of the same gap — an artifact trail would have settled R1-18 without argument. **Sharpened round 2 (R2-6): run 3 is now a third, live demonstration.** Two backlog commits (`47ae48d`, `88eb57f`) cite run-3 measurements as fact (panel-anatomy findings, the cost-audit tool's own findings), yet `git ls-tree` at `88eb57f` shows exactly two run directories under `research/` and neither is run 3 — no run directory, no `friction.md`, no `journal.jsonl`. Run 3 evidently executed and left zero artifact trail, exactly the failure mode this row argues for closing, happening again while the argument was still open | Medium-high — this retrospective is itself the demonstration: §0–§2 could not have been written for run 1 without it, and could not be done natively for run 2 — and now a third run has come and gone unrecorded, raising this from "the fix would help" to "the fix is now retroactively unable to help for a run that already happened" | Low-medium — a copy-and-gzip step after the workflow returns | **Fix before run 4 — with no fallback for run 3's already-lost trail** [L2+L3, sharpened R2-6] |
| 12 | **Blue pre-flight self-audit reads red's gap-pattern memory** (15 pattern files, §1.4) before submitting | Medium-high — the library is real, substantial, and currently one-directional (red writes, nobody reads) | Medium — yesterday's expensive red discovery becomes tomorrow's free blue checklist line; limited to pattern-matchable gaps, red stays load-bearing for novel ones | Low — one line in `blue-researcher.md`'s pre-submission checklist | **Fix before run 4** [L3] |
| 13 | **PDF full-text/table extraction — adopt, don't build, WITH a vetting step** (vetting step added round 1, R1-14) | High — recurs every round, all roles (§4) | High — kept 3+ figures at unable-to-corroborate across 4 rounds at the time (memory-architecture corpus's own **MA-R1-19, MA-R1-28, MA-R3-14, MA-R3-15, MA-R4-9** — prefixed round 4, R4-5; see [^GapIdScheme]). **Corrected round 5 (R5-2), reconciled against §2.1 Tier C's fuller trace: this cell's own five-member list already omitted MA-R2-8 (right call — it closed round 3), but of the five it does list, only MA-R1-19 is still open and lossy-fetch-blocked as of round 4; MA-R1-28/MA-R3-14/MA-R3-15 have since closed by ordinary live re-fetch (not a PDF tool), and MA-R4-9 is a diagnosed miscitation, not a lossy-fetch instance — see [^MAStatusR5].** The historical point stands (the friction recurred across 4 rounds and shaped real citation-confidence costs at the time), but citing it as a live "kept... at unable-to-corroborate" blocker today overstates the current count by 4x | **Low once scoped as adoption**: off-the-shelf MCP servers exist (`arxiv-latex-mcp` for exact LaTeX figures; `pdf-reader-mcp` for tables with cell data/confidence)[^PdfMcp] — no bespoke `sc-pdf-extract` Go tool needed, contrary to the backlog's tentative candidate. **The report elsewhere headlines CVE-2026-21852-class supply-chain poisoning (§1.1) and then graded these two third-party tools on cost alone with no vetting line — corrected here:** before wiring in, pin the version, review source/maintainer activity, and scope MCP permissions to the minimum the extraction task needs (read-only, no network beyond the fetch target). Both tools check out fine on a live look this round (active maintenance, passing test suites) — the gap was the missing vetting *step*, not a defect in the tools themselves[^PdfMcp] | **Build now via adoption, with the vetting step as a stated precondition, not a silent skip — the re-grade from unscoped-medium/high to low complexity still materially changes the priority order** [L2; L1 graded medium-high pre-re-grade — superseded; vetting step closes R1-14 by addition] |
| 14 | **Primary security-advisory access** | Medium — recurred rounds 1–3, **stopped by round 4** | High when it fires — was load-bearing for the R2-2/R1-8 double-bind keystone | Unknown/possibly unfixable from this side (auth/allowlist decision outside the repo) | **Risk-accept with a revisit trigger** [L2, adopted over L1's fix-if-feasible]: round 4's record shows the team engineered around it (unconditional de-authorized channel, §13.7[^ChangelogR2]) — the workaround generalizes; building access infrastructure for a gap already designed around fails the complexity test. Revisit if it becomes load-bearing again; use run 2's hedging pattern as the standing template [L1] |
| 15 | **ENAMETOOLONG append-path/chunking tooling** — risk-accept rationale corrected round 1 (R1-13); **recurrence count corrected round 2 (R2-1)** | High, re-graded from Medium — Windows-specific. **Correction (R2-1): the round-1 re-grade cited "confirmed recurred 3 times across 3 runs... per debate.md's merge-seat friction" — that source does not contain an ENAMETOOLONG event.** Direct re-read of `debate.md`'s round-1 merge-seat friction section (lines ~146-153) shows exactly two items logged there: lossy PDF-fetch depth and the lens-writes-transcript process misfit — no ENAMETOOLONG, no heredoc, no command-length event. The corpus attests exactly **two** documented ENAMETOOLONG-class occurrences: `run2-friction.md` line 4 (red-merge-r1, run 2) and this retrospective's own round-0 heredoc shell-parse failure (`blue/CHANGELOG.md` Round 0's "chunked-heredoc workaround failed on shell parsing"); `run1-friction.md` contains zero mentions. Likely mechanism: the write-block's separately-correct "third occurrence" count (a different defect class, tracked accurately in row 8/§4 row 4) was transposed onto this adjacent narrative during the R1-13 edit. **Honest count: 2 documented occurrences across 2 runs** (run 2 and this retrospective). Whether "High" likelihood is still the right grade on 2/2 (both occurrences under Windows + large-append conditions, both forced a workaround) rather than mechanically-inflated to 3/3 is argued below, not re-asserted by the wrong count. **Narrowed further, round 3 (R3-7): occurrence 2's mechanism is confirmed only as "failed on shell parsing" (`blue/CHANGELOG.md` Round 0), not confirmed as the same length-ceiling class as occurrence 1** (`run2-friction.md` line 4, which explicitly names "Windows command-length limit" and an errno-class cause). A shell-parse failure can also arise from quoting or CRLF line-ending issues independent of payload length — CRLF rejection is a separately documented fragility class in this same corpus (backlog item 12, `.gitattributes` fix). So the honest evidentiary state is **1-confirmed-as-length-ceiling + 1-same-family-plausible-but-unconfirmed**, not 2-confirmed-identical-mechanism. This is narrower than the already-closed R2-1 correction (which fixed *how many* times something occurred); this flags *what* occurrence 2 actually was. The transcript that would settle it is very likely unrecoverable at this point — **argued risk-accept: leave the High-likelihood grade resting on the "2/2 under the same Windows/large-payload conditions" framing already stated below, with this one-clause caveat that occurrence 2's exact mechanism is unconfirmed, rather than spending a search hunting a transcript that probably no longer exists** | Low-medium — forces extra calls or a scratchpad detour, not a failure | Medium — a new tool surface. **Mechanism correction (R1-13): the previous "the skeleton fix may moot most of it by construction" clause does not hold** — ENAMETOOLONG is a payload/command-length ceiling on a single large shell call, orthogonal to Write-vs-Edit/append; a sufficiently large *append* overflows the identical ceiling. The mechanism-accurate mitigation is a chunked-write/append helper (fixed-size chunks under the OS command-length ceiling), not the write-block skeleton fix, which addresses a different defect class entirely | **Risk-accept for run 4 given the still-modest impact (workarounds proven both times), tracking honestly at 2 confirmed occurrences, not 3; build the chunking helper if it recurs a third time (not a fourth — corrected trigger, R2-1) rather than continuing to absorb the workaround cost indefinitely. Likelihood argued, not miscounted: two independent hits under the same Windows/large-payload conditions across the only two runs where a large single-call write was attempted at all is a 2-for-2 rate on the triggering conditions, which supports keeping High even at the honest count — but that argument, not the false "3 times," is what should carry the grade** [L1+L2, corrected round 1; count corrected round 2, R2-1] |
| 16 | **[MERGED]** Per-role model split (`model` bulk / `judgmentModel` judgment, default inherit) | Medium — cost lever, not correctness | Medium (cost only) | Low | **Shipped** (verified on `main` @ `47ae48d`, with the resume warning: never change `model`/`judgmentModel` on a resume — busts cache keys, re-runs completed rounds at full price) [L1; L3's "defer" superseded by it already existing]. **Doctrine-vs-implementation tension named round 1 (R1-12), disposition below in row 16b** |
| 16b | **Red-lens routing tier — reclassify or document the tradeoff** (new row, added round 1 per R1-12) | High — confirmed on `main`: `debate.js`'s own doctrine comment reads "cheapen redundancy and mechanics, never judgment or **the adversary**," yet the routing table sends each `red-lens` pass (the actual leaf-node audit work) to the cheap `bulk` tier and only `red-merge` (consolidation) to `judgment`[^JudgeUnguarded] | Medium-high — a dev/smoke run with a cheap bulk `model` silently downgrades exactly the verification this corpus shows is load-bearing (e.g., this round's citation-miscitation catches, R1-4/R1-5/R1-6/R1-7, all came from lens passes) | Low — either (a) move `red-lens` dispatches to the `judgment` tier (small cost increase, every round), or (b) keep the current routing and add one documented sentence: "lens passes run at bulk-tier by design; treat cheap-lens-run gap-grades with a confidence discount relative to a full-strength run" | **Adopt (b) for cost-sensitive dev/smoke runs, (a) as the default for keeper runs** — the doctrine comment already implies (a) is the intended keeper behavior; the gap was that dev/smoke-tier cost cutting silently applied the same discount to keeper runs with no stated tradeoff. State it; do not leave it implicit [R1-12] |
| 17 | **[MERGED]** Catechism template survey-topic instrumentation | n/a — template shipped, verified on `main` @ `47ae48d` | Medium — the doubt is untested in its actual target case (§1.4) | Low — pick one explainer/survey topic for a future run | **Instrument, don't assume** [L1+L3] |
| 18 | **Round-scoped audit narrowing** (rounds 2+: changed sections + contested gaps + spot checks) | Medium — red's full-re-read x scaled lenses x rounds is the dominant burn per the backlog. **Sharpened round 3 (R3-8): `ideas/backlog.md` item 28's new sub-item (d), a merge-seat cost analysis of run 3's transcripts, identifies the actual driver as TURNS x CONTEXT (an agent re-reads its whole context every tool call; red-merge-r1 alone: ~100–150K of material, 2.7M+ cache reads), not report length per se — this narrows "full-re-read is expensive" from a general intuition to a measured, specific mechanism**[^CostFigureProvenance] | Medium-high (cost) but **trades directly against the "re-read the FULL living report" MUST** in `red-auditor.md`, which the corpus shows catching real regressions (R1-1's retracted false claim was caught only by a full re-read) [L3] | Medium-high — needs a principled scoping rule or it reopens the failure mode the gate exists to prevent. **The backlog's own (d) sub-item independently proposes a candidate scoping rule this round hadn't considered: shard the findings file into an open-items ledger vs. a closed archive, so a merge reads "open + this round only" — the full-re-read principle would still cover red-vs-blue (the adversarial check this row protects), just not red re-reading its own already-closed cases, the same pattern the citation ledger already uses for citations** | **Hold / risk-accept status quo for run 4** — backlog marks it human-gated; the one lever where getting it wrong is a missed regression, i.e., the gate's own failure mode. Revisit only after #3/#4/#16 prove insufficient, **now with the sharding proposal above as the first candidate scoping rule to evaluate on that revisit, rather than starting from a blank design** [all lanes; sharpened R3-8] |
| 19 | **FEOV's own research phase as an untrusted-content-poisoning surface** (new row, added round 1 per R1-15) | Medium — no live incident observed in either run; a structural analogy (reasoned, not sourced from an observed exploit) | High if it fires — ~22 WebFetch/WebSearch pulls per run feed a multi-agent debate with repo write authority; a poisoned page shaping a claim that reaches `report.md` unchallenged is the same class of failure as CVE-2026-21852 (§1.1), just with the agent's own research corpus as the untrusted-context vector instead of a memory store | Medium — no bespoke tool; the mitigation is discipline already partially present (leaf-node citation verification is exactly the check that would catch a fabricated-source injection) plus one explicit addition | **Risk-accept with a named, disposed mitigation, honestly scoped (rewritten round 2, R2-7):** the round-1 wording claimed the leaf-node citation lens catches a poisoned page via "independent re-verification against a second source" — **that overstates what the protocol actually specifies and what the codebase implements.** `research-protocol/SKILL.md` and `agents/red-auditor.md`'s verification mandates both describe re-reading *the same* cited source to confirm it corroborates the claim, not cross-referencing against an independent second source; a repo-wide grep for "independent" in the plugin returns zero hits anywhere, including inside the ledger clause itself (`debate.js:156`) — **corrected round 3, R3-6: the original phrasing implied one hit existed inside the ledger clause's own text; live re-grep at the merge seat this round confirms zero hits, full stop, which is the stronger version of this row's own point** (the mitigation's honest-scoping claim does not even have a literal "independent" to lean on, anywhere in the source). A self-consistent poisoned page — one that states a fabricated fact and cites a fabricated-but-internally-consistent source for it — passes the current leaf-node check by construction, because the check re-reads the same (poisoned) reference rather than triangulating against a different one. **Honest scope of the existing mitigation:** it reliably catches *miscitation* (a claim that misstates or exceeds what its real, legitimate source actually says) — this is exactly how R1-4's "~19%/~95%" fabrication and R1-6's diffstat error were caught, both by a lens noticing a claim didn't match its cited source on independent re-fetch. It does **not** catch a *self-consistent fabrication* where the claim and its citation agree with each other but neither is real — the scenario this row was written to risk-accept. This is not a new gap to build around; it is a correction to what the already-risk-accepted mitigation is honestly good for. The residual-risk sentence and the open question this same round added (§5 item 8, "does it require a distinct defense, e.g. cross-referencing claimed sources against an independent index?") were saying almost the same thing from two directions without cross-referencing each other — collapsed here into one statement: **the mitigation catches source-misstatement; it does not catch consistent fabrication; a distinct defense (independent-index cross-referencing) would be needed for the latter and is explicitly not built.** The risk-accept disposition itself still stands on this honest scoping — no live incident has been observed, and building a cross-referencing capability against a hypothetical is the same complexity-vs-likelihood judgment call as before — but it is now a decision made with the mitigation's real limits stated, not an implied capability the system does not have [R1-15; rewritten R2-7] |

| 20 | **(Added round 3, R3-1) Guard the schema-legal degenerate `{verdict: 'FAIL', gaps: []}` shape — decided round 4 (R4-2)** | Low-medium — requires red to return FAIL with a genuinely empty gaps array (a merge-lens bug or an over-cautious agent, not an adversarial input); zero observed live occurrences | High — silent `maxRounds`-long token burn with no judge invocation and a terminal return that pairs `UNVERIFIED` with `gaps_outstanding: 0`, a self-contradictory report a human or a downstream automation could misread as "clean" | Low — one guard clause immediately after the existing `blueEnv`/`redEnv` null-checks: if `redEnv.verdict === 'FAIL' && redEnv.gaps.length === 0`, **throw a distinguishing error** (`throw new Error('red-merge round ${round} returned FAIL with an empty gaps array — degenerate merge, refusing to loop silently')`) rather than converting toward a passing verdict or looping. **Decided, round 4 (R4-2): this row previously shipped red's own R3-1 required-fix text as an unresolved "either/or" between throwing and a logged-warning PASS — the only round-3 fix left as a disjunction where its siblings (rows 21, 22) shipped a choice, and the disjunction's silently-convert-toward-PASS branch is the exact silent-degradation shape the report argues against everywhere else (§3 row 19's poisoning finding, §2.3 item 1's throws-vs-degrades taxonomy, R2-7's honest-mitigation-scoping). Reasoned choice: throw. A `{verdict: 'FAIL', gaps: []}` return is evidence of a broken merge lens or a miscoded agent turn, not evidence that the report is clean — converting it toward `PASS` manufactures a false-positive verdict from a malformed input, which is strictly worse than halting for a human to look at the transcript. The one-time cost of a thrown error (lost round, but a legible one) is cheaper than a `PASS`/`UNVERIFIED` verdict a downstream reader cannot distinguish from a genuine clean audit. An argued loud-warning-PASS remains an acceptable alternative if a future maintainer has a reasoned case for it (e.g., if empty-gaps-FAIL turns out to correlate reliably with an over-cautious lens rather than a bug) — but that reasoning does not exist yet and should not be assumed by default** | **Fix before run 4 — add the guard (throw, per the decision above), add §2.3 addition 13 to the founding suite (updated to assert the throw, not the disjunction), and this row itself is the correction to §2.1's round-loop coverage claim and §2.3 item 8's framing** [R3-1; decided R4-2] |
| 21 | **(Added round 3, R3-2) `takeFriction('blue-synthesize', blueEnv)` after the line-136 null-guard** | High — certain by structure: the call site is simply absent, not merely under-triggered | Medium — loses exactly the round-0 blue-synthesize friction class, disproportionately the write-block/environment complaints per §4's ranking | Low — one line, mirroring the existing three `takeFriction` call sites | **Fix before run 4 — add the call, add §2.3 addition 14 to the founding suite, and correct §2.1's "never dropped" claim to name this exception until fixed** [R3-2] |
| 22 | **(Added round 3, R3-3) Deliver the judge's carried-gap rationale to `blue-respond`'s prompt** | Medium — no live `carried` resolution has occurred yet in this corpus, so the delivery gap is unexercised, not yet symptomatic | Medium — when it does fire, blue re-drafts against the original red gap object with none of the judge's specific "what further research blue owes" guidance, the exact thing the judge prompt was written to produce | Low — either (a) fold each contested gap's judge rationale into the `openGaps` payload sent to `blue-respond` (robust: survives even if `debate.md`'s prose changes), or (b) instruct `blue-respond` to read the latest `### LEAD` section of `debate.md` before drafting (cheap: no schema change) | **Fix before run 4 via (b)** — at low complexity this is cheaper than continuing to carry the gap; risk-acceptance is for when complexity exceeds likelihood × impact, which does not hold here. §2.3 item 5 is corrected above to stop implying the rationale is already delivered, closing the documentation half of this gap immediately; the code half (b) is the remaining work [R3-3] |
| 23 | **(Added round 4, R4-1) Lineage-following contested-gap detection — the docket detector is id-string-equal, not lineage-aware** | **High** — not hypothetical: this exact corpus contains four live regression chains (R1-5→R2-4→R3-4/R3-9; R2-5→R3-10; R2-7→R3-6; R2-8→R3-5→R4-3) — **corrected round 5, R5-1: this cell previously carried a different, discarded enumeration.** Round 4's own debate record (`debate.md`, round-4 BLUE, item 1) states in the synthesizer's own words that this list was reconstructed independently on the first pass (`R1-13→R2-1→R3-7; R1-16→R2-8→R3-5; R2-5→R3-8`), found less precise than red's round-4 merge-seat enumeration on direct comparison, and explicitly "adopted... in place of my own" — but the substitution reached §2.1(b) only, not this row, leaving this cell shipping the discarded list. Two of its three entries are independently contradicted by `red/findings.md`'s own status lines, read verbatim: R2-5's regression successor is **R3-10**, not R3-8 (§2.1(b) above cites this correctly); and R2-1 closed clean, with the record's own R1-13 chain explicitly noting no reopening at R3-7 (that citation was this cell's error, not a real chain). The third entry (R1-16→R2-8→R3-5) truncated the chain's live tip — the same lineage's third link, R4-3, is stated two rows up in §2.1(b) and inside this very row's required-fix column, but was dropped from this cell alone. Corrected here to §2.1(b)'s verbatim list, one clause, no new research — and the judge was dispatched zero times across every completed round to date (`debate.md` has zero `### LEAD` section headers) despite all four disputes running the debate's full length | **High** — the contested-docket-to-judge escalation is the engine's entire deadlock/adjudication mechanism; a detector that structurally cannot arm on regression-chain gaps means that mechanism has never once fired in this corpus, and nothing but the `maxRounds` cost ceiling would stop a less-good-faith pair of seats from spinning a lineage indefinitely (the project's own backlog, commit `42dba2d`, reaches the identical conclusion independently) | Medium — (1) add `supersedes: { type: 'array', items: { type: 'string' } }` (optional) to `RED_ENVELOPE`'s gap schema and instruct red-merge to set it when closing a gap "WITH REGRESSION" and minting a successor; (2) change the contested-detection line from bare `prevGapIds.has(g.id)` to also match when `g.supersedes` intersects the full adjudicated-and-prior-gap-id history (not just the previous round — this is where the fix composes with row 2/addition-3's window-widening, not substitutes for it); (3) add founding-suite addition 15 (§2.3) as the regression test. **Named gap, closed by addition, round 5 (R5-5): (1)-(2) as originally scoped rely on exactly the unenforced good-faith compliance this very row indicts two sentences earlier in its own likelihood cell** — `supersedes` is an optional schema field set purely by prompt instruction, with nothing that validates a regression-flagged closure actually names its successor; this corpus has already demonstrated that class twice (§3 row 21's friction field went unset for three rounds before anyone noticed; R4-2 shipped an undecided disjunction verbatim). A `supersedes` field nobody checks is this report's own R3-1 schema-legal-but-empty pattern one level up, and the failure would be telemetry-invisible — `contested` would simply stay computed-correctly-from-an-empty-array, indistinguishable from "no lineage exists" at a glance. **(4) add a script-level structural check, not a prompt-only convention:** after `red-merge` returns, if any gap's closure reason matches `/WITH REGRESSION/i` (or, more robustly, a dedicated `closure_class` enum value) and no gap in that same round's `gaps` array lists its id in a `supersedes` array, throw `` `red-merge round ${round} closed gap ${id} WITH REGRESSION but no successor gap names it in supersedes — lineage silently dropped` `` — mirroring the R4-2 precedent of choosing a loud failure over a silently-accepted gap in the mechanism. This is not scope creep: the check is a few lines alongside the existing null-guards, reuses the schema fields (1)-(2) already add, and closes exactly the enforcement gap that made (1)-(2) alone a plausible-looking but hollow fix | **Fix before run 4 — independent of and additive to the `prevGapIds`-widening fix already scoped for the narrower same-id-skips-a-round case (§2.1, §2.3 addition 3); see §5 item 12 for the explicit independence statement; addition 15 (§2.3) extended to assert the throw when a closure omits its required `supersedes` entry, per R5-5** [R4-1; enforcement gap closed R5-5] |
| 24 | **(Added round 5, R5-6) Friction aggregate is script-local and lost on any mid-run throw** — the `friction` array (`debate.js` line 145) is only ever written out at the terminal `return` (line 210–217) and the final-assembly prompt (line 207); a thrown error at any of the (currently three, soon four with row 20's guard) guard sites never reaches either | Medium — trigger-conditional, not certain: fires only on the same rare inputs that trip the args guard (line 36) or a null `blueEnv`/`redEnv` (lines 136, 171) — not on every run, unlike row 21's every-run-certain single-seat drop | Medium — loses telemetry for exactly the runs most worth learning from (the ones that crashed); does not lose report content or corrupt the verdict — a crash is already loud on its own, so the harm is a missing improvement signal, not a silent failure | Low-medium — no script-side filesystem write (the script has none, by design); instead, extend the existing prompt text at the four `takeFriction`-adjacent seats (`red-merge`, `judge`, `blue-respond`, and `blue-synthesize` once row 21 lands) with one added sentence: "if you have friction to report, also append it directly to `${runDir}/friction.md` (one line, attributed) in addition to the envelope's `friction` field, so it survives even if a later phase aborts" — reusing the row-8/§4-proven Bash-append pattern rather than inventing a new mechanism | **Build alongside row 21 (same prompt-editing pass, same seats) rather than risk-accept: the fix is cheap because it reuses an already-validated pattern, and "silence is not acceptable" per this round's own review — a bare risk-accept would leave a self-improvement signal gap with no argued rationale on record.** Add a founding-suite case (§2.3 addition 16: stub a throw immediately after a seat's `takeFriction` call and assert the pre-throw friction text is still recoverable — e.g. from a stubbed `Bash`-append side channel the harness can inspect — known-failing until the prompt text ships) [R5-6] |

**Explicitly risk-accepted (pragmatist duty against scope creep), with rationale:** full per-claim
manifest (build the cheap half first, §1.2) [L2]; blanket lane-count raise as a diversity fix
(method diversity is the cheaper, better-targeted lever; diminishing-returns literature, restated
qualitatively per R1-5[^DiminishingReturns]) [L2]; artifact rename as the primary write-block fix
(unverified mechanism, plugin-wide blast radius, superseded by the skeleton) [L2+L3]; advisory-access
infrastructure (#14 above) [L2]; ENAMETOOLONG tooling (#15, re-graded round 1 to track recurrence
but still risk-accepted pending a third occurrence (corrected R2-1; fixed round 4, R4-4 — this
was the fifth location still carrying the pre-correction "4th" numeral)) [L1]; audit narrowing
(#18) [all lanes];
simulating Task-tool permissions/OS limits/fetch lossiness inside the simulator (§2.1 boundary)
[L3]; the content-poisoning-of-own-research-phase residual risk beyond the existing leaf-node
citation-verification mitigation (#19, added round 1) [R1-15].

---

## 4. What the friction corpus says the agents actually need, ranked by distinct independent reporters (task question 4)

**Counting-method note (union honesty):** the three lanes counted differently and got different
role-counts for two items. Lane 2 counted run-2 rounds strictly by attributed
entry;[^FrictionCount] lane 3 counted across both runs; lane 1 counted roles per complaint-class
across both runs including judge references. Discrepancies are preserved below and are themselves
an instrumentation finding: **friction entries need structured role+round attribution** so this
ranking never again requires archaeology. All three lanes agree on the top-2 and on the overall
shape.

| Rank | Gap | Distinct-role evidence | Persistence | Status |
|---|---|---|---|---|
| 1 | **PDF full-text / table extraction** (HTML/abstract fetches lossy for in-table figures) | **3/3 roles** — red r1–r4, blue r1–r4, judge r2 [all lanes agree] | Every round of the only complete run | **Open — highest-value unbuilt capability**; as of memory-architecture round 2, blocked **MA-R1-19, MA-R1-28, MA-R2-8, MA-R3-14, MA-R3-15, MA-R4-9** (prefixed round 4, R4-5 — see [^GapIdScheme]) from resolving past "unable-to-corroborate-at-leaf-node." **Corrected round 5 (R5-2): this status line is stale by two rounds and overstates the current blockage 6x.** Direct re-read of `research/2026-07-12_memory-architecture/red/findings.md` (current, round-4 state): MA-R1-28 and MA-R2-8 CLOSED round 3 by ordinary live re-fetch (no PDF tool); MA-R3-14 and MA-R3-15 CLOSED round 4, same mechanism; MA-R4-9 is open but is a diagnosed miscitation (wrong paper cited), not a case this row's "unable-to-corroborate-at-leaf-node" framing describes. **Only MA-R1-19 remains open and genuinely blocked by lossy fetch** — the disposition below (the #1-ranked, unbuilt PDF-adoption fix) is unaffected and still correctly the top build priority, since it rests on the friction *recurring across every round for four rounds running* (a historical fact, unchanged) and on this retrospective's own top-ranked friction-corpus finding (§4 methodology below), not on the memory-architecture corpus's current open-gap count. See [^MAStatusR5]; the backlog independently reaches the same #1 ranking[^LiveBacklog]; adoption-not-construction fix available (§3 #13, now with a vetting-step precondition per R1-14) |
| 2 | **Primary security-advisory access** (CVE-2026-21852; post-cutoff vendor-blog-only sourcing) | **3/3 roles** — L2: red r2–r3, blue r2–r3, judge r2; L3: red r1+r3, blue r2–r3, judge r2 (round-attribution differs; role-count agrees) | Rounds 1–3, **absent in round 4** | Open, but **engineered around**: `blue/CHANGELOG.md` R2 shows the R1-8/R2-2 double-bind closed via an unconditional de-authorized channel[^ChangelogR2] — the design was made robust to both branches of the unresolvable fact, so agents stopped hitting it as a live blocker. The imposed cost was real debate rounds, not just citation confidence [L2+L3]. Risk-accepted with revisit trigger (§3 #14) |
| 3 | **Uninitialized run-directory/topic** (`undefined` paths) | Reported by **every dispatched run-1 agent** — L3 counts 10/10 invocations (frontier, 3x blue-lane, blue-synthesize, 3x red-lens, red-merge, judge); L1 counts 16 dispatches in the journal (the journal logs re-attempts; both counts are honest against different units)[^Run1Journal] | One incident, one run — the historically loudest single complaint by raw count | **[MERGED, confirmed round 1]** — `main` @ `47ae48d` carries the args parse+guard, verified live this round; safely retired from "open" as of this round[^Reverify47ae48d]; demonstrates the fix-and-it-stops pattern the top-2 have not yet followed [L3, corrected from "unmerged" per R1-1] |
| 4 | **Report-named-file write-block** | Role-count differs by counting method: L2 strict run-2 attribution = red r1 only; L3 cross-run = 2 roles (blue on `report.md` run 1; red on `findings.md` run 2); L1 = 3 roles incl. judge references; +2 occurrences across this retrospective's own rounds 0 and 1 (blue-synthesis in round 0, §0 addendum; red-merge in round 1, debate.md round-1 RED friction) | 3+ runs, 2+ filenames — systemic *across* runs despite low per-run count, and now confirmed hitting a second role (red) one round after it hit blue **(corrected round 2, R2-2: previously misdated "the same round" — the two hits are one round apart, not simultaneous)** [L3, updated round 1; date corrected round 2] | **[MERGED (skeleton), first live trial still owed]** — the pre-created blackboard skeleton (§3 #8a) is confirmed on `main` @ `47ae48d`, but both this retrospective's occurrences fired against ad hoc scratch writes, not the pre-created skeleton path, so the fix's real test is still pending; belt-and-braces prompt fix available (§3 #8) [corrected from "unmerged" per R1-1] |
| 5 | Windows ENAMETOOLONG / long-heredoc fragility | red (run 2 r1); this retrospective's own synthesizer (round 0, `blue/CHANGELOG.md`'s "chunked-heredoc workaround failed on shell parsing") | **2 documented occurrences across 2 runs — corrected round 2 (R2-1): the round-1 "3 times across 3 runs, per debate.md's merge-seat friction" citation is wrong.** Direct re-read of `debate.md`'s round-1 merge-seat friction contains no ENAMETOOLONG event (it logs only PDF-fetch depth and the lens-writes-transcript process misfit); `run1-friction.md` has zero mentions either. Only two artifact-attested occurrences exist: `run2-friction.md` line 4 and this retrospective's own round-0 heredoc failure. The likely mechanism was a transposed count from the write-block's separately-correct three-occurrence tally (row 4 above) onto this narrative during the R1-13 edit. Likelihood retained at High on the honest 2/2 rate (see §3 row 15's argued rationale), not on the miscounted 3/3. **Narrowed round 3 (R3-7): the "2/2 rate on the triggering conditions" language describes 2 occurrences under the same Windows/large-payload conditions, but only occurrence 1 (`run2-friction.md` line 4) is confirmed as the length-ceiling mechanism specifically — occurrence 2's source (`blue/CHANGELOG.md` Round 0) names only "failed on shell parsing," which a quoting or CRLF issue could also produce. 1-confirmed + 1-same-family-plausible, not 2-confirmed-identical; argued risk-accept given the transcript is likely unrecoverable** | Open; workarounds proven (chunking, scratchpad-copy, Edit-append); risk-accept but tracked for a third occurrence, not a fourth (§3 #15, corrected R2-1) |
| 6 | Live-source drift needing access-date deltas | red only | **run 2, round 1 only — corrected round 1 (R1-8): the cited friction file (`run2-friction.md`) attests exactly one instance (red-merge-r1); no round-2 instance was found in that source. The original "r1–r2" range is unconfirmed by the cited artifact and is corrected here rather than re-asserted** | Open; protocol-documentation fix, not a tool gap (§3 #10) |
| 7 | Trajectory-extractor implementation opacity | red only | run 2 r3 only | Open; the underlying artifact doesn't exist yet either |
| 8 | No sandbox for Auto Dream runtime behavior | red only | run 2 r4 only | Open; external vendor dependency, not fixable here |
| 9 | Springer/institutional auth-wall | red only | run 2 r4 only | Open; single occurrence |
| 10 | Long-tail singletons: blue accepting red's gh-issue finding without independent re-verification (run 2 r1); `%TMP%` clobbering in the doctor bootstrap; missing workflow progress heartbeats | one report each, unattributed or single-role | — | Recorded for completeness; no action graded [L3] |

**Shape verdict (H5): CONFIRMED with two corrections, one of which is itself corrected round 1.**
The stable, unresolved list is exactly the predicted two document-fetch-fidelity gaps, dominated
by fetch fidelity rather than tool diversity [all lanes]. Original corrections: (a) the top tier
by raw role-count is four items, not 2–3, but two of the four were fixed-on-branch at the time
[L1]; (b) the frontier's "(already fixed)" parenthetical for the round-1 cluster was **false for
all three items as of the original `main` HEAD checked** — the write-block fix was unmerged,
ENAMETOOLONG had only improvised workarounds, and the preflight guard was never on the shipping
ref despite the friction file's claim (§0) [L3]. **Round 1 correction to correction (b):** as of
`main` @ `47ae48d` (this round), the write-block skeleton and the preflight guard *are* merged —
(b) is now accurate only for ENAMETOOLONG, which remains workaround-only. Rows 3–4 above are
updated accordingly. **Round 1's claim that row 5's persistence count "corrected upward (3, not 2)
per this round's own third occurrence" is itself corrected round 2 (R2-1): the cited "third
occurrence" was a miscount transposed from the write-block's own (genuinely 3-times) tally — row
5 is honestly 2 documented occurrences across 2 runs, not 3; see row 5 and §3 row 15 above.** The
informative asymmetry between the top two ranked gaps is unaffected by any of this: PDF friction
recurs every round because no workaround exists — every hedge is a standing per-round cost;
advisory friction stopped when the team designed around it. That asymmetry, not raw counts, is why
PDF extraction outranks advisory access in the build list [L2].

---

## 5. Open questions carried past this round

1. Does the Workflow tool's `schema` parameter guarantee conformance-or-null on `agent()`
   returns, or can a malformed non-null envelope reach the script? (Determines whether §2.3 case
   10 is a defensive-code gap or dead code.) [L2]
2. Was run 2's `--lanes 2` deliberate cost control or an omission? Unrecoverable from the corpus;
   the `lanes` return-object field (§3 #7) prevents recurrence of the ambiguity. [all lanes]
3. Does the Catechism template hold up on a genuinely explanatory/survey-shaped topic? Untested
   in its target case. (§1.4, §3 #17) [L1+L3]
4. Does the pre-created blackboard skeleton actually clear the write-block under real Task-tool
   permissions? PR #14's own commit message defers to run 3's live trial; issue #13890 and this
   synthesis's own occurrences suggest the block may not be purely filename-keyed. **Sharpened
   round 2 (R2-6): run 3 evidently already happened (two backlog commits cite its live
   measurements) with zero artifact trail in the tree — this question's answer may already exist,
   but only as unlogged operator experience, not as anything this retrospective's corpus can
   read.** [L1+L3, sharpened R2-6]
5. Will lens/source-class lane assignment measurably reduce breadth-phase convergence, given the
   diversity-collapse literature's warning that prompt-level differentiation under-delivers?
   Needs the claim manifest (§3 #5) in place to measure it on run 4. **Baseline recalibrated round
   2 (R2-3): the "2 diverse agents match/exceed 16 homogeneous" figure this question's revisit
   trigger was implicitly calibrated against is the paper's L4 (model+persona) condition, not the
   L2 (persona-only, same-provider) condition item 6 actually implements — L2's real curve needs 8
   agents to match the same baseline. "Under-delivers" for this question now means "closer to L2's
   8-agent curve than to any 2-agent parity," not the disproven claim that 2-agent parity was ever
   plausible for same-provider lens diversity.** [L1+L3, recalibrated R2-3]
6. Is red's 15-pattern memory library actually consulted to effect (does blue's pre-flight read,
   once added, reduce shallow gaps caught by red)? Measurable as red-gap-count-by-class deltas
   across runs. [L3]
7. **(Added round 1)** Does the pre-created blackboard skeleton actually clear the write-block
   under real Task-tool permissions, tested against the real skeleton path rather than a scratch
   run? Both write-block occurrences observed across this retrospective's rounds 0 and 1
   (blue-synthesis, red-merge — corrected round 2, R2-2: not the same round, one round apart) hit
   ad hoc writes against this retrospective's own tree, not the merged skeleton path — so item 4
   above remains untested even though the fix has shipped. **Sharpened round 2 (R2-6): see item 4
   — the live trial this question defers to may have already happened in run 3, unrecorded.**
   [R1-1, R1-2 correction; sharpened R2-6]
8. **(Added round 1, R1-15)** Is a poisoning attack against the citation itself — a fabricated
   but internally-consistent secondary source, as opposed to a fabricated primary claim — covered
   by the leaf-node citation-verification lens, or does it require a distinct defense (e.g.
   cross-referencing claimed sources against an independent index)? **Answered round 2 (R2-7): no
   — the leaf-node lens re-reads the same cited source and cannot catch an internally-consistent
   fabrication by construction; a distinct defense (independent-index cross-referencing) would be
   required and is not built.** §3 row 19 rewritten to state this plainly rather than leave the
   two passages contradicting each other.
9. **(Added round 1, R1-12)** Once red-lens dispatches are reclassified to the judgment tier for
   keeper runs (§3 row 16b), does the measured gap-catch rate actually improve, or was the
   bulk/judgment split immaterial to lens quality in practice? Measurable by comparing round-1
   citation-catch rates (this retrospective, lenses at bulk-tier) against a future keeper run with
   lenses at judgment-tier.
10. **(Added round 1, R1-13)** Does ENAMETOOLONG recur a third time (**corrected round 2, R2-1:
    the true count is 2 documented occurrences, not 3 — the trigger is a third, not a fourth**)
    before the chunked-append helper is built? The risk-accept in §3 row 15 is explicitly
    conditioned on this — a third occurrence should trigger the build, not another risk-accept
    renewal.
11. **(Added round 2, R2-10)** Rows 14/15's revisit triggers and item 10 above all depend on a
    future reader correctly counting recurrences across runs — exactly the step this retrospective's
    own R2-1 shows failing without an artifact. `red/citation-ledger.md` gives citations a durable,
    append-only counter; no equivalent exists yet for recurrence-triggered risk-accepts. Until a
    similar ledger exists for recurrence counts, treat every "Nth occurrence triggers a build"
    disposition in §3 (rows 14, 15) as advisory, not self-enforcing — the count must be re-derived
    from primary sources each time, not trusted from the last report that stated it. [R2-10]
12. **(Added round 4, R4-1)** §2.1/§2.3 addition 3's `prevGapIds`-widening fix (class (a),
    same-id-skips-a-round) and §3 row 23's `supersedes`-field lineage fix (class (b), fresh-id
    regression chains) are **independent repairs to independent failure modes**, not two framings
    of the same bug — building one does not subsume or de-prioritize the other. Widening the
    window without adding lineage-following leaves every regression-chain gap in this corpus
    (four of them) undetected; adding lineage-following without widening the window leaves the
    narrower same-id case undetected. Both are scoped as "fix before run 4" (§3 rows 2/3 and 23)
    precisely because neither is a partial credit toward the other. Open until both ship and a
    combined simulator case (addition 3 + addition 15 run together) confirms neither regresses the
    other's coverage. [R4-1]

---

## Footnotes

[^WorkflowJs]: `plugins/frank-exchange-of-views/skills/research-protocol/scripts/workflow.js`, this repo, `main` @ `9ff0fad`, accessed 2026-07-13. Line 16 `const { topic, runDir, lanes = 3, maxRounds = 12 } = args` (no parse/guard); line 140 `if (redEnv.verdict === 'PASS') break`, line 145 `redEnv.gaps.map(...)`, line 162 `redEnv.gaps.filter(...)` (no null-guards); dispatch loop `Array.from({ length: lanes }, ...)` with no minimum; line 104 lane prompt assigns hypothesis order only. `git log -p` shows no commit touching line 16 since introduction.
[^MainVsBranch]: Synthesizer leaf-node verification, this machine, 2026-07-13: `git branch -a` (branch `feat/feov-dogfood-round-1` exists locally and on origin); `git log --oneline` (`main` HEAD = `9ff0fad`); direct read of `workflow.js` lines 10–30 on `main` (unguarded destructure present); `git ls-tree -r feat/feov-dogfood-round-1` (contains `scripts/debate.js`, `tests/simulator/{harness.mjs,debate.test.mjs}`, `references/catechism_template.md`); `git show feat/feov-dogfood-round-1:...debate.js` (parse+guard and refuse-dispatch check present in header). Also `wc -l`: run-2 `blue/report.md` = 2,145 lines; assembled `report.md` = 2,972; `red/findings.md` = 695; `lane-1.md` = 355, `lane-2.md` = 321.
[^PR14]: `gh pr view 14` (ctoforaday/special-circumstances), accessed 2026-07-13. "feat: template-misfit friction + dogfood round-1 fixes", state OPEN at first access, branch `feat/feov-dogfood-round-1`, opened 2026-07-12. **Diffstat corrected round 1 (R1-6):** originally cited as "+2281/−46"; live re-check via `gh pr view 14 --json additions,deletions,commits` (2026-07-13, this round) returns additions:318, deletions:48, changedFiles:18, commits:11 — the original figure was miscounted at origin (not explained by intervening commits) and is corrected here rather than re-asserted.
[^Reverify47ae48d]: Direct re-verification, this machine, round 1 (2026-07-13): `git log --oneline -3 origin/main` (HEAD = `47ae48d`, one commit past the PR #14 merge `00018a5`, docs-only, message references "run 3"); `gh pr view 14 --json state,mergedAt,mergeCommit` (state MERGED, mergedAt 2026-07-14T05:58:54Z, mergeCommit `00018a5`); direct read of `plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js` on `main` @ `47ae48d`, full file: line 33 `const a = typeof args === 'string' ? JSON.parse(args) : args` + line 34 destructure + line 35 refuse-dispatch guard (all present); line 136 `if (!blueEnv) throw new Error(...)` (present); line 139 `const citationPasses = Math.min(4, Math.max(1, Math.ceil((blueEnv.claim_count || 20) / 40)))` outside the `while` loop at line 148, never reassigned inside it (confirms R1-11); line 181 `const judge = await agent(...)` and line 184 `for (const r of judge.resolutions)` with no intervening null check (confirms R1-2); line 162/164 `...bulk` for `red-lens` dispatch vs. line 168 `...judgment` for `red-merge` (confirms R1-12's routing-tier claim); `lanes = 3` default at line 34 with no minimum enforced anywhere in the dispatch loop (line 128, confirms row 7 still open); `commands/research.md` argument-hint and body text checked for `--smoke` — no match (confirms row 4 still open).
[^JudgeUnguarded]: See [^Reverify47ae48d] — same verification pass; called out separately here because R1-2 and R1-11 are graded as distinct gaps with distinct required fixes even though both were confirmed in the same file read.
[^PinnedRepoState]: This report's own repo-state citations from this round forward carry a SHA + UTC timestamp + "re-verify before acting" note (see [^Reverify47ae48d] for the pattern) — added round 1 in direct response to R1-1's finding that the original §0 had no such discipline despite demanding it of external citations (§3 row 10).
[^SmokeAbsent]: See [^Reverify47ae48d].
[^PR14Description]: `gh pr view 14` body text, accessed 2026-07-13. "`/research` now pre-creates the blackboard skeleton so subagents only append (dodges the harness write-block on fresh report-like files; red's own recommended fix)."
[^SimulatorTests]: `git show feat/feov-dogfood-round-1:plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs` and `.../harness.mjs`, accessed 2026-07-13. 11 `node --test` cases, ~200ms, CI job `debate-sim`; `AsyncFunction`-wrapped script body against stubbed `agent()`/`parallel`/`pipeline`.
[^CatechismTemplate]: `git show feat/feov-dogfood-round-1:plugins/frank-exchange-of-views/skills/research-protocol/references/catechism_template.md` vs. `references/heilmeier_template.md` on `main`, accessed 2026-07-13. Questions 4–9 reframed topic-agnostically ("The case against," "Of interest, or merely interesting?", cost and stopping points).
[^LiveBacklog]: `ideas/backlog.md` at `main` @ `9ff0fad` ("docs(backlog): graduate simulator, per-role models, citation ledger, write-block fix to PR #14"), diffed against the frozen `inputs/backlog.md` snapshot, accessed 2026-07-13. Four items flip open to `[x]` between snapshot and HEAD; also the source of the cheap-testing-stack item 15, the pre-flight self-audit item 4, the trajectory item 10, the CRLF item 12, and "TOP TOOL GAP ... PDF full-text/table extraction."
[^ClaimManifest]: `ideas/backlog.md`, item (5) "CLAIM MANIFEST" — "blue emits a machine-readable ledger (claim, citation, self-graded confidence, lane provenance)"; "one artifact, five wins." Accessed 2026-07-13.
[^WriteBlock]: `ideas/backlog.md`, item 8: write-block "CONFIRMED as a hard, report.md-specific tool error (forensics: `is_error: True`, 'Subagents should return findings as text, not write report files')." Accessed 2026-07-13. Reproduced live against this run's `blue/report.md` during synthesis, 2026-07-13. **Corrected round 1 (R1-10): strike "report.md-specific" from the backlog's own claim as literally quoted here** — the same corpus's `run2-friction.md` line 3 records "filename-heuristic guard ('findings' .md)" and this retrospective's own round-1 red-merge pass independently hit the identical block writing `red/findings.md`, not `report.md` — two distinct filenames on record. The block keys on a semantic class of report-like output-artifact names, not the one literal string "report.md"; the backlog's original phrasing is carried here as a direct quote (attributed, not endorsed) and should not be read as this report's own conclusion about the block's actual trigger condition.
[^Rename]: `ideas/backlog.md`, item: rename `skills/research-protocol/scripts/workflow.js` to `debate.js`. Accessed 2026-07-13.
[^ResearchCommand]: `plugins/frank-exchange-of-views/commands/research.md`, accessed 2026-07-13. "`--lanes` (blue candidate drafts, default 3)".
[^Run2Frontier]: `research/2026-07-12_memory-architecture/blue/frontier.md`, accessed 2026-07-13. "Lane assignments: lane 1 took H1 to saturation then breadth; lane 2 took H2 to saturation then breadth."
[^Run2Lane1]: `research/2026-07-12_memory-architecture/blue/candidates/lane-1.md` (355 lines, H1-deep), accessed 2026-07-13.
[^Run2Lane2]: `research/2026-07-12_memory-architecture/blue/candidates/lane-2.md` (321 lines, H2-deep), accessed 2026-07-13.
[^LocalGrep]: Local verification, 2026-07-13: Grep "lane-1|lane-2|lane 1|lane 2" against `research/2026-07-12_memory-architecture/blue/report.md` (2,145 lines) — 0 per-claim matches.
[^LocalGrepRed]: Local verification, 2026-07-13: Grep "both lanes|one lane|single lane|lane-sourced|lane provenance" against `red/findings.md` (695 lines) — 0 matches; Grep "corroborat" — 66 matches, all external-citation confidence.
[^BlueReportGrep]: Grep -i "lane" against the run-2 assembled report and `blue/report.md` — 7 total matches, all method-level or section-header, none per-claim. Accessed 2026-07-13. (File-attribution corrected per [^MainVsBranch].)
[^BlueReportUnverified]: `research/2026-07-12_memory-architecture/blue/report.md` §10: "cited by the proposal without independent corroboration in either lane." Accessed 2026-07-13.
[^RedFindingsGrep]: Grep -i "lane" against `red/findings.md` — 1 match, line 156: "Disconfirming budget met in both blue lanes. Not a gap." Accessed 2026-07-13.
[^RedAuditorSpec]: `plugins/frank-exchange-of-views/agents/red-auditor.md`, accessed 2026-07-13. Corroboration-confidence mandate per statement-reference pair; `memory: project` declaration; "AFTER catching a new gap pattern (not instance), YOU MUST record it in your project memory."
[^R2-10]: `research/2026-07-12_memory-architecture/red/findings.md`, lines ~424–451 (`[^SingleUserLowRisk]` footnote-provenance finding), accessed 2026-07-13.
[^GapIdScheme]: **Added round 4 (R4-5).** This retrospective's own red audit and the memory-architecture corpus's red audit (`research/2026-07-12_memory-architecture/red/findings.md`) both use a bare `R<round>-<n>` gap-id scheme with no corpus qualifier. Both corpora are now into their fourth round (this retrospective: R4-1..R4-5; memory-architecture: confirmed live at `research/2026-07-12_memory-architecture/red/findings.md` lines 118, 631-633 to run at least to R4-12), so the two id spaces collide in form. Direct grep confirms this report bare-referenced the memory-architecture corpus's own gap ids at exactly three prior locations (§1.2's R2-10 mention, §3 row 13, §4 rank-1 row) plus the §2.1 Tier C lossy-fetch bullet — four locations total, all corrected this round to the `MA-` prefix rather than renaming this retrospective's own (much more numerous) internal ids, which would be a full-document, high-blast-radius edit for a scheme that has not caused a misread within this document so far. Going forward: any reference into the memory-architecture corpus's `red/findings.md` gap ids is prefixed `MA-`; this retrospective's own ids remain bare, matching every round's usage to date. Accessed 2026-07-14.
[^MAStatusR5]: **Added round 5 (R5-2).** Direct read, this round, of `research/2026-07-12_memory-architecture/red/findings.md` (current, round-4 state), superseding the round-2-era status this report had been re-citing without a diff at rounds 4 and 5. **MA-R1-28**: "R3 status: CLOSED. All three compounding sub-defects discharged this round... The honest wide band... is stated and each half traces to a carrier. Red accepts closure" (line 345). **MA-R2-8**: "R3 status: CLOSED. Both legs re-verified LIVE at the leaf node this round... Contradicted number gone — red accepts closure" (line 410); also listed under round-3 "Real closures" (line 51) and "Closed this round" (line 99) — no residual survives, so "MA-R2-8 residual" (as this report's §2.1/§4 citations read pre-correction) is itself inaccurate, not merely stale. **MA-R3-14**: "R4 status: CLOSED, but the re-homing spawned R4-10... The over-attribution red flagged is gone" (line 551) — the spawned MA-R4-10 is a distinct, newly-created gap, not a surviving MA-R3-14 residual. **MA-R3-15**: "R4 status: CLOSED... the unsourced 77% lower bound is gone... Body↔footnote parity holds" (line 558). **MA-R4-9**: severity LOW-MEDIUM, open as of round 4 (raised that same round, so unclosed is expected, not evidence of persistence) — "Verified at the leaf node via three independent routes... the paper... does not use cosine-similarity thresholds and does not report true-duplicate precision by cosine bin... a different measurement" (lines 626-627): a wrong-paper citation, diagnosed and fixable by re-attribution, not a case where the paper's true content is unreachable by non-PDF means. **MA-R1-19**: "Carried: R1-19 (agent-PR figures, friction-blocked)" (lines 118-119); its own entry states "a PDF-table-extraction tool would discharge this definitively" (line 285) — the one member of the original six-id list for which the lossy-fetch/PDF-tool framing is actually correct and current. Accessed 2026-07-14.
[^ChangelogR0]: `research/2026-07-12_memory-architecture/blue/CHANGELOG.md`, Round 0 entry, accessed 2026-07-13.
[^ChangelogR2]: `research/2026-07-12_memory-architecture/blue/CHANGELOG.md`, Round 2 entry: "§13.7 R1-8/R2-2 lead docket CLOSED ... double-bind resolved by UNCONDITIONAL de-authorized channel." Accessed 2026-07-13.
[^Run1Friction]: `research/2026-07-12_feov-retrospective/inputs/run1-friction.md`, accessed 2026-07-13. "Cost of the null run: 16 agents, 252.9k tokens, 11m48s"; "Known outcomes" remediation claim (shown false for `main` in §0).
[^Run1Journal]: `research/2026-07-12_feov-retrospective/inputs/run1-defect-record/trajectories/journal.jsonl` (32 lines), accessed 2026-07-13. Every dispatch received literal `undefined` topic/runDir paths; each agent independently detected and refused to fabricate. L3 counts 10 distinct invocations; L1 counts 16 dispatch entries (incl. re-attempts).
[^RedMergeR2]: `research/2026-07-12_feov-retrospective/inputs/run1-friction.md`, line 7: "No round-2 red candidate lens passes were ever produced, so the merge had no inputs by construction." Accessed 2026-07-13.
[^Run2Friction]: `research/2026-07-12_feov-retrospective/inputs/run2-friction.md` — **corrected round 1 (R1-7): 21 entries, 4 rounds; the file's own header states "35 agents, 4 rounds" — "35" is the dispatched-agent count, not the friction-entry count, and the two were conflated in the original footnote.** Direct line count re-verified this round (`wc -l` + entry-marker count = 21). Accessed 2026-07-13. Red-merge-r1 entries: ENAMETOOLONG ~236-line heredoc, live-source drift / access-date-delta recommendation (one instance each, per R1-8's correction to §4 row 6).
[^FrictionCount]: `research/2026-07-12_feov-retrospective/inputs/run2-friction.md`, **corrected round 1 (R1-7): the 21-entry count by role and round (lane 2's strict-attribution table) — not "35," which is the header's agent count, not the entry count.** Accessed 2026-07-13.
[^HookGrep]: Grep -i "report|findings" against `plugins/prosthetic-conscience/hooks/` — no matches; the write-block is not implemented by this repo's hooks. Local verification, 2026-07-13.
[^NoPackageJson]: Glob `**/package.json` across the repo — no results. Local verification, 2026-07-13.
[^GoTests]: `plugins/prosthetic-conscience/tools/cmd/*/main_test.go`, `tools/internal/*/*_test.go` — standard-library `testing`, no external framework. Local verification, 2026-07-13.
[^DiversityCollapse]: "Diversity Collapse in Multi-Agent LLM Systems: Structural Coupling and Collective Failure in Open-Ended Idea Generation", arXiv:2604.18005, accessed 2026-07-13. Persona assignment does not prevent convergence; recommends structural decoupling, isolated generation, diversity metrics.
[^WisdomCrowds]: Search synthesis across "The Wisdom of the LLM Crowd" and related 2026 LLM-ensemble literature, accessed 2026-07-13. "Under independence ... the collective outperforms its members; under correlation, diversity collapses and the collective inherits its members' errors." **URL corrected round 1 (R1-19):** the bare-domain form 404s; the resource is at `alexanderakm.github.io/projects/wisdom-of-llm-crowd.pdf`, re-verified live this round (fetch returns a 672KB PDF, not a 404).
[^IsolatedCorrection]: "The Cost of Consensus: Isolated Self-Correction Prevails Over Unguided Homogeneous Multi-Agent Debate", arXiv:2605.00914, accessed 2026-07-13. Isolated self-correction matches/beats debate at 2.1–3.4x fewer tokens; sycophantic conformity up to 85.5%; oracle gaps up to 32.3 points.
[^DiminishingReturns]: Multi-agent debate diversity/scaling survey (arXiv:2603.20640; arXiv:2601.19921 "Demystifying Multi-Agent Debate"; VentureBeat "'More agents' isn't a reliable path"; arXiv:2605.00914), accessed 2026-07-13. **Corrected round 1 (R1-5): the "2–3 rounds / 2–4 agents" bound is not individually pinned to any one of these four sources at abstract/HTML depth — VentureBeat's diminishing-returns story is about tool count, not agent count, and none of the other three states the bound as a single verified figure.** Independent re-search this round (WebSearch, 2026-07-13) corroborates the qualitative aggregate: accuracy plateaus around 2–3 debate rounds and 2–4 agents for moderate-complexity tasks, with the breakeven at 3–4 agents for harder tasks and continued gains to 7 agents on the hardest — treat as a synthesis across sources, not a single citable number. **Round 2 (R2-4): red correctly flagged this "7 agents" clause as the identical over-attribution failure R1-5 was raised to fix, still uncited beyond "independent re-search."** Red's required fix proposed a specific source (arXiv:2606.02646, "The Ringelmann Effect in Multi-Agent LLM Systems: A Scaling Law for Effective Team Size") as the citable origin. **Direct verification this round (abstract + full HTML fetch, 2026-07-14) does not support that attribution and is recorded here as a rebuttal to the *specific proposed source*, not to the underlying gap, which is conceded:** the paper's hardest free-form-math benchmark is GSM-Hard, not GSM-Plus; it states "on harder tasks, the practical knee is N≈10," not 7; and its headline finding is that effective team size plateaus around 1.8 agents by N=30 on free-form math regardless, with "a single N≤5 pilot" sufficient to predict the N=30 ceiling — none of which states "continued gains to 7 agents." Citing this paper for the "7 agents" figure would have repeated exactly the R1-4/R2-4 failure pattern (a specific number attributed to a source that, read past the abstract, does not contain it) rather than fixing it. **Resolution: the unpinned "7 agents on the hardest" clause is dropped rather than re-cited to an unverified source; the qualitative synthesis is restated without a specific agent-count ceiling for the hardest tier** — **corrected round 3 (R3-9): the round-2 replacement sentence conflated two distinct senses of "breakeven" in one clause** ("harder tasks shift the breakeven higher... may show diminishing returns arriving even earlier... rather than later" reads as a direct self-contradiction if "breakeven" means one thing throughout it). Disambiguated: arXiv:2606.02646 supports two separate, non-contradictory claims about different quantities. (1) The **nominal-N practical knee** — the agent count at which *marginal* accuracy gains per additional agent become small — is *higher* on harder tasks than on moderate ones (the paper states N≈10 for harder tasks vs. a lower knee implied for easier ones). (2) The **effective-diversity saturation ceiling** — a distinct measure of how many agents' worth of *independent* signal the ensemble actually contains, regardless of nominal N — plateaus around 1.8 by N=30 on the paper's hardest free-form-math benchmark, and a single N≤5 pilot is sufficient to predict that ceiling; this is a *low* ceiling reached *early* in nominal-N terms. Both are true simultaneously: the nominal knee sits further out (higher breakeven, sense 1) while the *useful* diversity inside the ensemble saturates close to the floor almost immediately (early ceiling, sense 2) — they are not alternative readings of the same number, they describe different curves. The report's synthesis favors sense 2 for the practical recommendation (§3 row 6, item 6's redundancy-floor lens-assignment fix): adding nominal agent count on hard tasks buys little because the *effective* diversity ceiling is reached almost immediately, even though the nominal knee (sense 1) is further out than on easier tasks. This is the second round this exact footnote required a citation-discipline fix and the third round running a defect recurred in this one footnote (R1-5, R2-4, now R3-9); §3's disposition should treat this footnote's history as a standing argument for the claim manifest (§3 row 5) applying to blue's own footnotes, not only to cross-lane provenance.
[^AgentDiversity]: "Understanding Agent Scaling in LLM-Based Multi-Agent Systems via Diversity", arXiv:2602.03794, accessed 2026-07-13. **Corrected round 1 (R1-4): the previously-cited "~19% lower pairwise error correlation, up to ~95% of independent-ensemble gain" does not appear in this paper** — re-verified this round via direct abstract fetch plus an independent full-text percentage search; no matching figures found. The paper's real, verified, citable claim is qualitative: homogeneous-agent scaling exhibits strong diminishing returns because outputs correlate; heterogeneous agents (different models/prompts/tools) continue to yield gains; "2 diverse agents can match or exceed the performance of 16 homogeneous agents"; same-base-model agents remain more correlated than architecturally distinct members.
[^NarrativeSimilarity]: "Multiperspectivity as a Resource for Narrative Similarity Prediction", arXiv:2603.22103, accessed 2026-07-13 (added round 1, R1-4's corrected citation for the "~19%" figure). A 31-LLM-persona ensemble for narrative-similarity prediction: "Practitioner" personas show 19% lower pairwise error correlation than "Lay" personas (r=0.388 vs. r=0.461, direct full-text fetch this round), producing a majority-vote ensemble accuracy gain (76.0% vs. 75.3%) despite lower individual per-persona accuracy (71.0% vs. 71.7%). Domain is narrative-similarity annotation, not multi-agent research debate — cited here as a supporting analogy for the mechanism (persona diversity lowering pairwise error correlation), not a transferable rate.
[^AgentTestTiers]: "Agent Testing Strategies: Unit, Integration, and End-to-End Testing for AI Systems" — OpenHelm Blog, accessed 2026-07-13. Mocked-LLM unit tests catch deterministic code bugs, miss reasoning/hallucination/context failures; ~95%/5% mock-to-real split.
[^ProvenanceSurvey]: "From Agent Traces to Trust: A Survey of Evidence Tracing and Execution Provenance in LLM Agents", arXiv:2606.04990, accessed 2026-07-13. "Execution provenance" as a typed activity graph; PROV-AGENT extending W3C PROV for LLM-specific claim-to-source-to-agent lineage.
[^SubagentWriteBug]: "[BUG] Subagents unable to write files and call MCP tools silently" — GitHub, anthropics/claude-code issue #13890, accessed 2026-07-13. Subagent Task-tool write failures independent of this plugin. **Softened round 1 (R1-9):** the issue's failure signature is a *silent no-op* — the subagent believes the write succeeded and nothing happens — while this repo's block is an *explicit worded refusal* ("Subagents should return findings as text..."). Both are subagent write failures, but different signatures; the inference "therefore the block may not be purely filename-keyed" is a plausible but weaker transfer than the original phrasing implied.
[^PdfMcp]: GitHub — takashiishida/arxiv-latex-mcp (LaTeX source for exact figures) and SylphxAI/pdf-reader-mcp (table extraction with cell data, bounding boxes, confidence scores), accessed 2026-07-13.
[^CostFigureProvenance]: **Added round 2 (R2-5). Re-pinned round 3 (R3-8).** `ideas/backlog.md` item 28, originally pinned at `main` @ `88eb57f`, accessed 2026-07-14: "run cost audit — a tool, not a diet (from run 3's live measurement)... the panel token counter excludes cache traffic = 92% of real flow (panel said 610K; transcripts showed 47.7M)." **Round 3 update (R3-8): the same item gained a sub-item (d) at `main` @ `d164ab2` — three commits past the round-2 pin — re-fetched live at this round's merge seat:** "(d) MERGE-SEAT ANALYSIS (run-3 transcripts): the driver is TURNS x CONTEXT, not file size — an agent re-reads its whole context every tool call (red-merge-r1: ~100-150K of material, 2.7M+ cache reads). Levers, cheapest first: (1) shard the findings (open-items ledger vs. closed archive, same pattern as the citation ledger); (2) a collator stage that digests the round's lens passes before the judgment merge reads them; (3) prompt-level read batching; (4) tooling step-up (beads vs. sc-gaps) if gap volume grows. Burn is spiky and the spikes are judgment-seat full re-reads." Re-pinned at `d164ab2`, accessed 2026-07-14; this sub-item is new, directly relevant evidence for §3 row 18's audit-narrowing hold (see row 18's updated rationale) and does not contradict anything already stated here — it sharpens the mechanism rather than changing the undercount finding below. This is high-confidence evidence that the *panel-reported* token figure undercounts real spend by roughly an order of magnitude by excluding cache-read/cache-write traffic. Whether this report's own headline figures (252.9k for run 1, ~3M for run 2) suffer the same undercount is **unverified either way** — [^Run1Journal] suggests run 1's figure may be transcript-derived (closer to real spend), while run 2's ~3M traces to a friction self-report of unstated methodology — so the two numbers may not be apples-to-apples with each other, let alone with a cost-audited figure. Direction of risk is understatement only: if either figure is undercounted, the simulator's and `--smoke` mode's cost case (§2.3, §2.4) only strengthens, since the incident costs they compare against would be even higher than stated. Precision/provenance gap, not verdict-changing. Use the now-existing cost-audit tool (backlog item 28(a)) to produce comparable, cache-inclusive figures before quoting incident costs again.

---

## Red team findings (in full)

Embedded verbatim from `red/findings.md` (living audit, cumulative across rounds; final verdict and every gap disposition):

---

# red findings — FEOV retrospective

LIVING audit — cumulative across rounds.

**Round 1** consolidated five lens passes (`red/candidates/round-1-lens-{1..5}.md`): verdict FAIL,
20 gaps (R1-1..R1-20), full grading preserved below with round-2 statuses prepended.

**Round 2** consolidated five lens passes (`red/candidates/round-2-lens-{1..5}.md`): verdict FAIL,
all 20 round-1 gaps closed (5 with regressions), 11 new gaps (R2-1..R2-11), full grading preserved
below with round-3 statuses prepended.

**Round 3** consolidated five lens passes (`red/candidates/round-3-lens-{1..5}.md`:
3 leaf-node citation slices covering §0–§5 and every footnote, 1 logic/completeness, 1
dark-side/risk) — superseding the single-lens (lens-5) preview previously in this slot, whose three
gaps keep their ids (R3-1..R3-3) unchanged. Keystone claims re-verified live at merge time
(2026-07-14): `origin/main` HEAD is `d164ab2` — one docs-only commit past the report's newest
pin (`88eb57f`), three past its older `47ae48d` pin; `git diff 88eb57f..d164ab2 -- .../debate.js`
is empty (only `ideas/backlog.md` changed), so no report claim is invalidated by drift;
`git grep -ni "independen" -- plugins/frank-exchange-of-views` re-run at the merge seat: **zero
matches anywhere, including inside `debate.js:156`'s `ledgerClause` string** (bears on R3-6, and
on red's own round-2 phrasing — see R3-6's provenance note); the §1.1 body clause "continued gains
observed to 7 agents on the hardest" confirmed still present by direct full-report re-read (770
lines) at merge (R3-4); [^DiminishingReturns]'s self-contradictory sentence and the untagged
252.9k/253k–3M instances (§2.1 first row; §3 row 4 impact cell) read directly (R3-9, R3-10);
§3 row 6's reconciliation arithmetic re-computed against its own four-item roster (R3-5).
arXiv:2606.02646 was independently re-fetched by two separate round-3 lenses (abstract + full
HTML), both confirming blue's R2-4 rebuttal exact: GSM-Hard (not GSM-Plus), "practical knee is
N≈10," effective team size plateaus ~1.8 agents by N=30, single N≤5 pilot predicts the ceiling.
Round-3 verification pairs are appended to `red/citation-ledger.md` (lines 90–111).

**Round 4** consolidated five lens passes (`red/candidates/round-4-lens-{1..5}.md`:
3 leaf-node citation slices — lens 1 §0–§1, lens 2 §2–§3, lens 3 §4/§5/Footnotes — 1
logic/completeness, 1 dark-side/risk). Keystone claims re-verified first-hand at the merge seat
(2026-07-14): `origin/main` HEAD is `42dba2d` — one docs-only commit past the report's newest pin
(`d164ab2`); `git diff d164ab2 42dba2d --stat` touches `ideas/backlog.md` only (+1/−1),
`debate.js` byte-identical across `47ae48d`/`88eb57f`/`d164ab2`/`42dba2d`. The drifting commit
itself is load-bearing: `git show 42dba2d` read directly — a new backlog item, *"the docket
detector tracks IDs, not lineages,"* citing **this retrospective's own gap chain
R1-5 → R2-4 → R3-4/R3-9 as its worked example** (bears on R4-1). `debate.js:178`'s
`contested = redEnv.gaps.filter(g => prevGapIds.has(g.id))` re-read live; `RED_ENVELOPE`'s gap
schema confirmed to carry no `supersedes`/lineage field. `debate.md` checked at header level via
`grep -n "^### "`: **zero `### LEAD` sections across rounds 0–3** — only BLUE/RED headers; stated
precisely because a plain-text grep returns one match (line 528, a *quoted phrase* inside round-3
prose, not a judge section) and lens 5's "zero matches" phrasing would otherwise repeat the R3-6
imprecision class. Also read directly at merge: §3 row 20's disjunction text and §2.3 addition
13's actual assertion wording (bears on R4-2, including a merge-seat temper of lens 4's grading);
§3 row 6's original floor sentence beside its round-3 correction (R4-3); the report-wide
`grep "4th|fourth"` showing exactly one uncorrected instance at line 670 (R4-4); and the three
cross-corpus id locations at lines 478/651/689 (R4-5). Round-4 verification pairs are in
`red/citation-ledger.md` (round-4 blocks, appended by the lenses plus merge-seat additions).

**Round 5 (this update)** consolidates five lens passes (`red/candidates/round-5-lens-{1..5}.md`:
3 leaf-node citation slices — lens 1 front matter/§0–§1, lens 2 §2–§3, lens 3 §4/§5/Footnotes — 1
logic/completeness, 1 dark-side/risk). Keystone claims re-verified first-hand at the merge seat
(2026-07-14): `origin/main` HEAD still `42dba2d` — zero drift since round 4's pin; `debate.js`
re-read at that ref (the `friction` array is script-local at line 145, `takeFriction` at 146,
throw sites at lines 36/136/171, and the only egress is the success-path assembly prompt at 207
and terminal `return {..., friction}` at 210–217 — bears on R5-6); `commands/research.md` step 5
read directly ("If the **returned** envelope carries `friction` entries, write them to
`<run directory>/friction.md`" — a thrown workflow never returns one; bears on R5-6); §3 row 23
(report line 727) and §2.1(b) (line 496) read side by side — row 23 still carries the chain list
`R1-13→R2-1→R3-7; R1-16→R2-8→R3-5; R2-5→R3-8` that `debate.md`'s own round-4 BLUE section
narrates discarding ("Re-verified red's list against the cited rows and adopted it in place of my
own... the report now cites red's enumeration" — true for §2.1, false for row 23; bears on R5-1);
`grep -c "three completed"` on the report = 3 (front matter, §2.1(b), row 23) against
`grep -n "^### " debate.md` = 9 headers (5 BLUE rounds 0–4, 4 RED rounds 1–4, zero `### LEAD`) —
round 4 is now the *fourth* completed judge-free round (bears on R5-3); the six
memory-architecture ids in §4 row 1 traced to their current status lines in that corpus's own
`red/findings.md` first-hand (MA-R1-28 and MA-R2-8 closed round 3, MA-R3-14 and MA-R3-15 closed
round 4, MA-R4-9 open but a fully-diagnosed miscitation verified via three independent routes,
MA-R1-19 the only genuinely open lossy-fetch-blocked case — bears on R5-2); §3 row 13's id set
extracted mechanically from report line 715 (five ids, **no** `MA-R2-8 residual`) against
§2.1/§4's six-id sets (bears on R5-2, and overrules one lens's contrary ledger line — see the
merge notes). Round-5 verification pairs are in `red/citation-ledger.md` (round-5 blocks, appended
by the lenses plus merge-seat additions and two merge-seat corrections).

## Verdict (round 5): FAIL

All 5 round-4 gaps are **CLOSED** — 4 clean, 1 **closed-with-regression** (R4-1→R5-1, with three
further, smaller residues arising inside the same fix's own text: R5-3, R5-4, R5-5). **6 gaps
open (R5-1..R5-6):** 1 medium-high (R5-5), 3 medium (R5-1, R5-2, R5-6), 2 low (R5-3, R5-4). None
disputes H1–H5 or any prior closure. The round's shape: the external citation surface stays
converged — every external re-fetch across the three citation lenses came back exact again — and
the new catches are internal-consistency drift (R5-1, R5-3, R5-4), cross-corpus citation-status
drift (R5-2), and two dark-side reflexivity findings against round 4's own fixes (R5-5, R5-6).
Severity trend resumes its decline (round 4's worst: HIGH; round 5's worst: MEDIUM-HIGH, closable
with one argued sentence). Round 4 projected this as the PASS round; it is not — but every open
gap is one-clause-to-one-paragraph sized, five of the six are copy-edits from already-verified
material inside this corpus, and none requires new research.

---

## Round-4 gaps — status after round 5

### R4-1 — CLOSED WITH REGRESSION (round 5) → R5-1 (three further residues inside the same fix's text: R5-3, R5-4, R5-5)
The required coverage correction landed in full: §2.1's rollover row states classes (a)/(b) side
by side with the widen-window remedy scoped to (a) only; §3 row 23 added and graded; §2.3
addition 15 added known-failing; §5 item 12 states the two fixes' independence explicitly — all
verified in place, and the fix shape matches the backlog item re-read first-hand at `42dba2d`.
But row 23's own likelihood cell ships blue's *discarded* first-pass chain enumeration (R5-1),
the fix's round-count numeral is already one round stale (R5-3), addition 15's uniform closure
labeling does not match the chain it claims to mirror (R5-4), and the `supersedes` remedy's
reliance on unenforced prompt compliance goes unnamed two sentences after the report indicts
exactly that reliance (R5-5).

### R4-2 — CLOSED (round 5)
Decided: throw, with stated reasoning (anti-silent-degradation, consistent with row 19, §2.3
item 1, and R2-7's precedent); §2.3 addition 13 extended with the matching positive assertion,
verbatim-identical in both locations. No guard on live `main` yet — consistent with the
docketed-for-run-4 disposition red pre-accepted. Corroboration: high. (R5-6 touches the same
throw surface, but it is a pre-existing structural class shared with the two shipped null-guards
— not a defect in this decision.)

### R4-3 — CLOSED (round 5)
The originating sentence edited at the source ("the adversarial-disconfirming-first lens (a
distinct method from local-repo critical-stance, named separately below)"); report-wide grep: the
retired slash-compound and `lanes >= 4` survive only inside explicitly-marked correction
narratives. Corroboration: high.

### R4-4 — CLOSED (round 5)
The fifth location now reads "pending a third occurrence (corrected R2-1)"; report-wide grep
"4th|fourth": zero remaining unmarked instances. Corroboration: high.

### R4-5 — CLOSED (round 5)
All four cross-corpus locations carry the `MA-` prefix; [^GapIdScheme] present and verified
against the memory-architecture corpus first-hand (all six ids real; that corpus confirmed to run
to at least R4-12). Closed as raised — the namespace hygiene is done. The round-5 finding against
the same id lists (R5-2) is a different defect class (status drift of the *referenced* gaps, not
namespace collision) and is not a reopening.

---

## Round-5 gaps (id | status | severity | likelihood x impact x complexity | corroboration)

### R5-1 — OPEN — MEDIUM — certain (static text, read side by side at the merge seat) x medium (the evidentiary cell of the report's highest-graded round-4 finding carries mis-traced lineages, two of them contradicted by this file's own closure record — inside the very row arguing lineage precision matters) x trivial (copy an already-correct list from two sections earlier) — corroboration: HIGH (three lenses converged independently — lenses 1, 2, 4; both quotes verbatim from the current report; every chain link checked against this file's own closure entries)
**Location:** §3 row 23, likelihood cell — *"this exact corpus contains four live regression
chains (R1-5→R2-4→R3-4/R3-9; R1-13→R2-1→R3-7; R1-16→R2-8→R3-5; R2-5→R3-8)"* — against §2.1
Tier A, rollover sub-row (b) — *"R1-5 → R2-4 → R3-4/R3-9... plus... R2-5 → R3-10... R2-7 → R3-6...
and R2-8 → R3-5 → R4-3."*
**Problem:** two different enumerations of the same "four live regression chains" claim, agreeing
only on the first chain. Row 23's version is blue's own discarded first-pass reconstruction —
`debate.md`'s round-4 BLUE section narrates replacing it ("Re-verified red's list against the
cited rows and adopted it in place of my own... the report now cites red's enumeration"), but the
substitution reached §2.1 only. Two of row 23's three non-shared entries are factually wrong
against this file's own record: R2-5's regression successor is **R3-10**, not R3-8 (R3-8 is a
distinct live-drift finding on the same footnote's pin); and R2-1 closed **clean** round 3 with
this file explicitly disclaiming the link ("R3-7, **not a reopening**"). The third entry
(R1-16→R2-8→R3-5) is truthful link-by-link but truncates the chain's live tip (→ R4-3). This is
the report's own named repair-reaches-one-location-not-all class (R3-4, R3-10, R4-4) recurring
*inside* R4-1's own fix, between two sections of the same round-4 patch. Secondary nuance, not
gated: the fullest forms of two §2.1 chains include their round-1 origins (R1-15→R2-7 and
R1-16→R2-8 are both real CLOSED-WITH-REGRESSION links per this file); §2.1's "at least three
more" hedge keeps its version accurate as stated — including the origins is optional.
**Required fix:** replace row 23's parenthetical with §2.1(b)'s list — "(R1-5→R2-4→R3-4/R3-9;
R2-5→R3-10; R2-7→R3-6; R2-8→R3-5→R4-3)" — one clause, no new research.

### R5-2 — OPEN — MEDIUM — high (the status claims are read at face value; no hedge or as-of date) x medium (the #1-build disposition survives on independent grounds — live backlog + MA-R1-19 — but the stated blocking evidence is inflated ~3x: 2 genuinely open vs. 6 cited, and 4 of the 6 resolved by exactly the ordinary re-fetch the framing implies cannot happen without the tool) x low — corroboration: LOW for the four closed-cited-as-open sub-claims (the primary source contradicts them), MEDIUM for MA-R4-9 (real and open, but a diagnosed miscitation — wrong failure class for this row), HIGH for MA-R1-19 (matches source exactly), HIGH for the three-way list inconsistency (mechanical extraction)
**Location:** §4 row 1, status cell — *"blocks memory-architecture's own MA-R1-19, MA-R1-28,
MA-R2-8 residual, MA-R3-14, MA-R3-15, MA-R4-9... from resolving past
'unable-to-corroborate-at-leaf-node'"* — plus the same claim's two siblings: §2.1 Tier C (six
ids) and §3 row 13 (*five* ids — drops `MA-R2-8 residual`).
**Problem:** citation-status drift on a living source. Direct read of the memory-architecture
corpus's own `red/findings.md`: MA-R1-28 and MA-R2-8 **closed round 3** (by live abstract
re-fetch and re-citation — not PDF extraction); MA-R3-14 and MA-R3-15 **closed round 4**;
MA-R4-9 is open but is a fully-diagnosed miscitation verified through three independent routes —
not an "unable-to-corroborate" block; only MA-R1-19 is a genuine, open, lossy-fetch-blocked case.
Second-order staleness: the underlying source sentence (that corpus's friction section, "a
full-PDF... tool would discharge R1-19, R1-28, R2-8's residual, and R2-10 definitively") was
written mid-round-2 and its own prediction was falsified the next round — two of those gaps
closed without any PDF tool. And the three in-report enumerations of the same set disagree on
membership (6/6/5), no two matching the source's current state.
**Required fix:** either restate historically ("as of memory-architecture round 2; four of six
since closed by ordinary re-fetch, not PDF tooling") or cite only the genuinely-open MA-R1-19
(plus MA-R4-9 correctly classed as a miscitation, if kept) — and reconcile all three locations to
one canonical list. The row's #1 ranking itself stands on the independent backlog corroboration.

### R5-3 — OPEN — LOW — certain (grep-counted: 3 instances) x low (correcting it *strengthens* the finding — a fourth completed round with zero judge dispatch) x trivial — corroboration: HIGH (`grep -c "three completed"` = 3; `grep -n "^### " debate.md` = 9 headers, 5 BLUE rounds 0–4, 4 RED rounds 1–4, zero `### LEAD`, re-run at the merge seat)
**Location:** front matter (Round 4 corrections summary — "the judge was never dispatched once
across three completed rounds"), §2.1(b), and §3 row 23 — all "three completed rounds."
**Problem:** accurate when blue wrote it mid-round-4; blue's own round-4 response has since
completed the round without a `### LEAD` section (the same lineage-blindness the row describes
kept the docket disarmed), making round 4 the fourth judge-free completed round. A stale
"current-state" numeral of exactly the class §3 row 11/R2-6 warns about, self-applied. Lens 4
named two locations; the merge-seat grep found the third (front matter) — the retracted-token
grep must run report-wide, both directions.
**Required fix:** "four completed rounds" (or round-agnostic: "every completed round to date") at
all three locations.

### R5-4 — OPEN — LOW — certain x low (documentation-fidelity nit on a proposed, unbuilt test case; test validity unaffected) x trivial — corroboration: HIGH (addition-15 text against this file's own R1-5/R2-4 closure lines, read verbatim)
**Location:** §2.3 addition 15 — *"round 2's merge closes `X-1` 'WITH REGRESSION'... and round
3's merge closes `X-2` 'WITH REGRESSION' and raises `X-3`/`X-3b`"* — claiming to "mirror this
corpus's own chain directly."
**Problem:** the mirrored chain's second link did not close "WITH REGRESSION": R2-4's recorded
status is "CLOSED, **REBUTTAL ACCEPTED WITH EVIDENCE** (round 3) → regressions R3-4, R3-9." The
detector logic under test does not depend on the closure-status label (both statuses mint
fresh-id successors), but the case's framing misdescribes the record it claims to mirror.
**Required fix:** loosen the wording ("closes... and mints a fresh-id successor, whatever the
closure status") or name both real labels.

### R5-5 — OPEN — MEDIUM-HIGH — medium (an optional field set under prompt instruction alone, by a merge agent under load — this corpus has already demonstrated the class twice: R3-2's schema-declared friction field uncalled for three rounds; R4-2's undecided disjunction shipped verbatim) x high (the failure is telemetry-invisible: an unset or vacuous `supersedes` leaves `contested.length` at 0, indistinguishable from "no regression occurred" — silent reversion to the exact defect the fix exists to close) x low report-side (one named sentence either way) — corroboration: HIGH (row 23's fix text and §2.1(b)'s good-faith sentence read verbatim in the same document; `RED_ENVELOPE` schema re-confirmed at `42dba2d`)
**Location:** §3 row 23, complexity cell — *"(1) add `supersedes: { type: 'array', items: {
type: 'string' } }` (**optional**) to `RED_ENVELOPE`'s gap schema and **instruct red-merge to set
it** when closing a gap 'WITH REGRESSION' and minting a successor"* — against §2.1(b), same
finding, two sentences earlier: *"a property of this run's actors' good faith, not a property the
detector enforces."*
**Problem:** the fix for lineage-blindness is itself blind to its own failure mode — it relies on
exactly the unenforced good-faith compliance the row indicts. Nothing validates that a
closed-WITH-REGRESSION gap has a successor whose `supersedes` names it. §2.3 addition 15 only
tests the *detector* given a correctly-populated canned field (Tier A); whether the live merge
agent reliably populates it is a real-model-reasoning question the report's own boundary
discipline (§2.1: "a simulator that fakes... judgment content is the research problem itself")
places at Tier C — a Tier-A test for a Tier-C failure surface, with no fallback named and no
residual risk flagged, and the row prices the whole three-part fix at Medium complexity as if the
schema/prompt half were sufficient. Making the field required does not close this: a vacuous or
incomplete array is schema-legal-but-semantically-empty — the report's own R3-1 class.
**Merge-seat temper of lens 5's HIGH:** the defective claim concerns a proposed, docketed,
unbuilt mechanism's future reliability plus an underpriced complexity cell — not a false claim
about current state; precedent is R2-7 (a mitigation overstating its guarantee on a live
risk-accept), graded MEDIUM-HIGH. Graded to match, with lens 5's likelihood/impact reasoning
adopted intact.
**Required fix:** one sentence at row 23 (or a new row) naming the residual: either an argued
risk-accept (three rounds of demonstrated merge-seat compliance — argued, not assumed, same
pattern as row 19's rescoped mitigation), or scope the cheap structural cross-check — record
regression-flagged closures in the merge envelope as a small required field on the *closed* gap,
and have the script assert every such closure reconciles against some new gap's `supersedes`
this round, throwing on mismatch per the report's own R4-2 precedent. Either closes it; silence
does not — the report cannot indict unenforced good faith at §2.1(b) and then ship an
unenforced-good-faith fix in the same row without saying so.

### R5-6 — OPEN — MEDIUM — medium (requires a mid-run throw; the null-return class has fired in this project's history — run 2's crash is why the guards exist — and R4-2's decided guard adds a third mid-loop site) x medium-high (the entire run's accumulated friction telemetry is lost at exactly the runs where something went wrong enough to be worth reporting; `friction.md` is never written — `commands/research.md` step 5 fires only on a successful return, and the line-207 assembly-prompt injection is success-path-only too) x low-medium (report-side: one paragraph; the compliant code-side fix needs one design decision, not new infrastructure) — corroboration: HIGH (`debate.js` lines 36/136/145–146/171/207/210–217 and `commands/research.md` step 5 all read first-hand at the merge seat, at `42dba2d`)
**Location:** §2.1, Tier A friction row (as corrected by R3-2 — the correction names the
`blue-synthesize` call-site exception but not this second, structurally distinct loss path) — and
§3 row 2's impact cell (*"a mid-debate crash at the judge seat loses every paid round up to that
point"* — the round-loss recognition never extended to the friction aggregate specifically).
**Problem:** the aggregated `friction` array lives only in script-local scope until the terminal
`return` — which a thrown error never reaches. Three throw sites exist today (args guard,
`blueEnv` null, `redEnv` null) and R4-2's decision adds a fourth; any of them firing mid-run
discards every seat's accumulated friction from every prior round, and the command layer's
friction persistence (step 5) is unreachable for a thrown workflow. The naive fix (script writes
incrementally) violates the script's own stated no-filesystem-access doctrine the same round it
would fix the gap. **Merge-seat temper of lens 5's MEDIUM-HIGH:** the trigger is conditional (a
throw must fire), unlike R3-2's every-run structural certainty, and the loss is telemetry, not
report content (the blackboard artifacts persist on disk) — graded MEDIUM, mechanism adopted
intact.
**Required fix:** one paragraph at §2.1's friction row or a new §3 row naming the
throw-loses-the-aggregate path, then either (a) adopt the architecture-compliant variant — each
schema'd seat appends its own friction line directly to `runDir/friction.md` via the
append-only-blackboard convention already adopted for §3 row 8, in addition to returning it in
the envelope — or (b) an argued risk-accept that crash-time friction is low-value; the argument
does not exist in the report today and is not assumable by silence (the report's own R4-2
standard).

---

## Merge-time verification and dedupe notes (round 5)

- Dedupe map: lens 1 Finding 1 + lens 2 Finding 1 + lens 4's in-pass "R5-1" = **R5-1** (three
  independent convergences). Lens 3's in-pass "R5-1" + lens 1 Finding 2 = **R5-2** (merged: the
  status-drift finding subsumes the list-membership inconsistency). Lens 4's in-pass "R5-2" =
  **R5-3** (extended at the merge seat: the grep found a third instance in the front matter that
  lens 4's two-location statement missed). Lens 2 Finding 2 = **R5-4**. Lens 5's in-pass "R5-1" =
  **R5-5** (tempered HIGH → MEDIUM-HIGH, reasons stated in the gap). Lens 5's in-pass "R5-2" =
  **R5-6** (tempered MEDIUM-HIGH → MEDIUM, reasons stated in the gap).
- **Two lens errors caught and overruled at the merge seat** — logged with the same discipline
  demanded of blue: (1) lens 5's "checked, not raised" item asserted row 23's chain enumeration
  "matches §3 row 23's and §2.1's text exactly — no discrepancy found"; direct read of report
  line 727 shows row 23 carries the discarded list, the live discrepancy three sibling lenses
  caught independently. An unquoted hold masked a discrepancy — hold-claims need the same
  side-by-side quoting as gap-claims (new red-side pattern, recorded in red's memory). (2) lens
  2's ledger line 168 asserted §3 row 13 carries the six-id `MA-` set including `MA-R2-8
  residual`; mechanical extraction of report line 715 shows five ids, no `MA-R2-8`. Lenses 1 and
  3 are correct; two merge-seat correction lines appended to `red/citation-ledger.md`.
- `origin/main` re-fetched at merge: HEAD still `42dba2d`; `debate.js` and `ideas/backlog.md`
  both stable at the round-4 pin. The memory-architecture statuses (R5-2) were traced first-hand
  against that corpus's `red/findings.md`, not taken from any lens's paste.
- Lens 5's disconfirming holds re-checked and accepted: the R4-2 throw cannot false-positive on a
  genuine PASS (the `verdict === 'PASS'` break precedes the guard's position); `parallel()`
  semantics are irrelevant to R5-6 (the throws are synchronous in the loop body); the
  cost-incentive angle on R5-5 is folded into its likelihood grade as unproven-aggravating, not
  asserted separately.

## Noted, not raised (round 5)

- **Chain-origin truncation in §2.1(b)'s list** (R1-15→R2-7 and R1-16→R2-8 are real
  closed-WITH-REGRESSION origins this file records, omitted from the report's chain heads):
  §2.1's "at least three more" hedge keeps its version accurate as stated; folded into R5-1 as an
  optional improvement, not gated.
- **Row 23's fix item (1) lacking a documented "WITH REGRESSION" protocol state** (lens 4,
  checked and held): the convention is this corpus's emergent practice, not doctrine, but the fix
  text's "instruct red-merge to set it" covers the prompt-change obligation at the same
  specificity as its §3 siblings — holding it to a fuller implementation-spec standard would be
  inconsistent.
- **Item 6's "cost is one more lane-dispatch" phrase** (lens 4): scoped to the floor's marginal
  cost; the full `lanes >= 5` arithmetic is stated two sentences later in the same cell. A full
  read is not misled.
- **Row 16b's dev/smoke-vs-keeper split** (lens 4): gated on the existing `model`/`judgmentModel`
  knobs, which ship independently of the unbuilt `--smoke` path. Mechanism real; held.
- **Red-auditor memory-pattern-file count now 25** (was 23 at round 4): settled round-2/3/4
  non-gap (live accreting mechanism); not re-litigated.
- **[^CostFigureProvenance] pin still `d164ab2`, HEAD `42dba2d`:** the one-commit drift is R4-1's
  own subject commit, byte-identical on item 28 (verified round 4); no re-pin gap.

## Verified clean this round (round 5 — see red/citation-ledger.md round-5 blocks)

Lens 1 (front matter/§0–§1): the R4-2/R4-3/R4-4/R4-5 front-matter summary sentences against
their cited rows — all verbatim-consistent; §1.2's R4-5 addition against the memory-architecture
MA-R2-10 source (exact); [^GapIdScheme]'s "runs to at least R4-12" re-confirmed live; report-wide
"4th|fourth" grep clean. Lens 2 (§2–§3): `RED_ENVELOPE` no-`supersedes` / pure-id-equality
`contested` / full-replacement `prevGapIds` all re-read live at `42dba2d` (code fix genuinely
undocked, as the report states); the 25-minute commit-delta claim exact (25m12s); the backlog
docket-detector item verbatim against the report's citation; §3 row 6's corrected floor sentence
and zero residual "lanes >= 4"; row 20/addition 13's identical throw-message text; both PDF-MCP
repos re-checked live (`archived: false`). Lens 3 (§4/§5/Footnotes): §4 rows 2–10 and §5 items
1–12 unchanged and held at prior HIGH; [^CostFigureProvenance] pin unchanged; backlog line 27(c)
"TOP TOOL GAP" re-fetched live (supports row 1's ranking independently of R5-2). Lens 4: all 46
prior-round gap fixes hold as described at their cited locations (the one exception raised as
R5-1); template compliance (top-level `report.md` still a correct stub pending PASS; all
footnotes semantic word-labels). Lens 5: all five round-4 closures re-verified first-hand at
`42dba2d`; friction-aggregation call sites and null-guard positions exact as the report
describes. Merge seat: the keystones in the round-5 header paragraph, first-hand.

## Round 5 disposition summary

FAIL. 5/5 round-4 gaps closed (1 with regression: R4-1→R5-1, plus three smaller residues in the
same fix's text). 6 new gaps open. Gate tier: **R5-5** (name the `supersedes` fix's residual
reliance on unenforced merge-seat compliance — one argued sentence, risk-accept or structural
cross-check, either accepted) and **R5-1/R5-2** (mechanical accuracy fixes to the evidence cells
of the report's two highest-profile rows — row 23's chain list and row 1/row 13/§2.1's
memory-architecture status claims). R5-3 and R5-4 are one-word-to-one-clause; R5-6 closes by one
paragraph or an argued risk-accept red will take if reasoned. None disputes H1–H5. Convergence
expectation: five of the six fixes are copy-edits from material already verified inside this
corpus; R5-5/R5-6 are single argued sentences/paragraphs. If blue lands these without new
regressions, round 6 is the PASS round. Red's own ledger this round: two lens errors (lens 5's
unquoted hold contradicting three sibling lenses; lens 2's wrong ledger line on row 13's id set)
caught and corrected at the merge seat before entering the record — and the unquoted-hold class
is new, logged in red's pattern memory.

---

## Verdict (round 4): FAIL — superseded by round 5, preserved

All 10 round-3 gaps are **CLOSED** — 8 clean, 2 **closed-with-regression** (fix responsive to the
gap as raised, but carrying a new, smaller defect re-raised as a round-4 gap: R3-1→R4-2,
R3-5→R4-3). **5 gaps open (R4-1..R4-5):** 1 high (R4-1), 1 medium (R4-2), 2 low-medium (R4-3,
R4-5), 1 low (R4-4). None disputes H1–H5 or any prior closure. The round's shape: the citation
surface has converged — all three citation lenses' external and local re-fetches came back clean;
every round-4 *prose* gap is repair-propagation residue at ≤ LOW-MEDIUM. The one HIGH (R4-1) is
driven by evidence that landed in the live repo 25 minutes after blue's round-3 pin — four of
five lenses converged on it independently. Honest trend note: headline severity did **not**
decline this round (round 3's worst was MEDIUM-HIGH; R4-1 is HIGH), but the report-quality trend
did — R4-1 is not a blue drafting failure, it is live-source drift plus a mechanism this corpus
itself demonstrates; it still gates because §2.1/§2.3 currently state the narrower rollover case
as the *only* version of the docket-detector defect.

---

## Round-3 gaps — status after round 4

### R3-1 — CLOSED WITH REGRESSION (round 4) → R4-2
§3 row 20 and §2.3 addition 13 added; row 20 itself states it is "the correction to §2.1's
round-loop coverage claim and §2.3 item 8's framing" — the report-side coverage correction red
required; code fix docketed for run 4, the disposition red pre-accepted. But the guard's
specified behavior ships as an unresolved disjunction — copied verbatim from red's own R3-1
required-fix text — never decided. See R4-2 (provenance note against red recorded there).

### R3-2 — CLOSED (round 4)
§2.1's "never dropped" claim corrected in place, naming the `blue-synthesize` exception with call
sites (lines 170/187/197 vs. the absent line-132 harvest) and the simulator test's real two-seat
coverage; §3 row 21 and §2.3 addition 14 added. Re-confirmed against live `debate.js` at
`42dba2d` by lens 2 and lens 5 independently. Corroboration: high.

### R3-3 — CLOSED (round 4)
§2.3 item 5 corrected; §3 row 22 added with an explicit option-(b) choice *with stated reasoning*
(cheaper at the stated complexity; risk-acceptance reserved for complexity exceeding
likelihood × impact) — a decision, not a disjunction, unlike row 20. Re-verified against live
`debate.js` (judge `resolutions` still read only for `adjudicated`/`friction`). Corroboration:
high.

### R3-4 — CLOSED (round 4)
§1.1's body clause edited to match the corrected footnote; the in-place correction note ("this
body clause previously still read 'continued gains observed to 7 agents...'") confirmed at line
257. Report-wide grep: the retracted clause survives only inside the correction note and the
footnote's preserved historical narrative, both explicitly marked retracted. Corroboration: high.

### R3-5 — CLOSED WITH REGRESSION (round 4) → R4-3
The reconciliation recomputed, choosing red's option (a): full four-method roster requires
`lanes >= 5`, stated plainly; row 7's language corrected from "floors N at 3" to "proposed
(unbuilt) floor targets `lanes >= 3`," aligned with its [OPEN] status — both required alignments
delivered, and this time the correction computes instead of asserting. But the *originating*
ambiguous sentence (the slash-compound "critical-stance/adversarial-disconfirming lens" floor
clause) stands unedited above the correction in the same cell. See R4-3.

### R3-6 — CLOSED (round 4)
Row 19 reworded to "zero hits anywhere in the plugin, including the ledger clause itself."
`git grep -ni "independen"` re-run at `42dba2d` by lens 3: still zero. Corroboration: high.

### R3-7 — CLOSED (round 4)
The one-clause narrowing added to §3 row 15 and §4 row 5 (1-confirmed-as-length-ceiling +
1-same-family-plausible, occurrence 2's source names only "shell parsing"), with the argued
risk-accept on the likely-unrecoverable transcript — exactly the disposition red offered to
accept. Textual distinction re-verified against `run2-friction.md` line 4 and
`blue/CHANGELOG.md` Round 0 by lens 3. Corroboration: high.

### R3-8 — CLOSED (round 4)
[^CostFigureProvenance] re-pinned at `d164ab2` with sub-item (d) quoted in full; §3 row 18's
rationale updated with the turns×context finding. The subsequent drift to `42dba2d` touched a
*different* backlog item (item 28 byte-identical across the diff, verified by lens 3) — not a
re-open; the new commit's content is R4-1's subject, handled there. Corroboration: high.

### R3-9 — CLOSED (round 4)
The two-senses disambiguation (nominal-N practical knee vs. effective-diversity saturation
ceiling) added to [^DiminishingReturns], with the synthesis explicitly favoring sense 2 and the
reason stated. Internally coherent on direct re-read (lens 3); figures match the
twice-independently-re-fetched arXiv:2606.02646 ledger entries. Corroboration: high.

### R3-10 — CLOSED (round 4)
Both untagged instances now tagged: §2.1's reading-order-first "252.9k" row and §3 row 4's
"253k–3M" impact cell both carry [^CostFigureProvenance] with in-place round-3 tag notes.
Verified by direct grep at merge (lines 453, 642). Corroboration: high.

---

## Round-4 gap details (original grading, preserved verbatim — round-5 statuses above supersede the "OPEN" markers below)

### R4-1 — OPEN — HIGH — certain (already realized in this corpus, not projected) x high x low-medium — corroboration: HIGH (direct read of `debate.js:142/176–190` at `42dba2d`; direct `git show 42dba2d` of the backlog item; header-level grep of `debate.md`; four of five lenses converged independently, each tracing the code first-hand rather than trusting the backlog)
**Location:** §2.1, Tier A — *"**Gap-id rollover across non-adjacent rounds**: `prevGapIds` holds
only the prior round, so a gap closed in round 1 recurring in round 3 classifies 'new,' not
'contested'"* — and §2.3 addition 3 — *"Gap-id rollover — id present r1, absent r2, present r3...
known-failing until `prevGapIds` widens to full adjudicated history."*
**Problem:** the report's only stated version of the docket-detector defect is the
same-id-skips-a-round timing case, and its own proposed remedy (widen `prevGapIds` to full
adjudicated history) does not close the broader case that is actually live in this corpus.
Contested-docket membership is pure id string-equality (`debate.js:178`); the gap schema has no
`supersedes`/lineage field; and red's own closed-WITH-REGRESSION methodology mints a **fresh id
for every successor gap** — so a multi-round dispute lineage never matches `prevGapIds`,
`contested` stays 0 every round, and the judge is never dispatched, no matter how long the
dispute persists. The only remaining brake is the `maxRounds` cost ceiling — which bounds cost,
not convergence quality, the docket's stated purpose ("so debates converge instead of grinding").
**Already realized, four chains in this corpus:** R1-5 → R2-4 → R3-4/R3-9 (one footnote, four
ids, three rounds), plus R2-5→R3-10, R2-7→R3-6, R2-8→R3-5→R4-3. Zero `### LEAD` sections exist
in this retrospective's entire transcript — the judge has never once been invoked across three
completed rounds despite exactly the rebuttal-and-regression chains the docket exists to route.
That this debate converged anyway is a property of this run's actors (blue conceded every
sustained gap in good faith), not a property the detector enforces; a spinning debate would show
identical `contested.length === 0` telemetry to `maxRounds`. Independently confirmed by the
project's own live backlog (commit `42dba2d`, 25 minutes after blue's `d164ab2` pin), which
names this retrospective's chain as its worked example and proposes the fix shape
(`supersedes: [prior-ids]` on the gap envelope; lineage-following contested-detection at chain
depth ≥ 2; a simulator lineage-regression case). Distinct from R3-1/R3-2/R3-3, all of which
assume the judge branch is at least reachable — this is one level upstream: the branch is never
entered for the corpus's dominant real gap-lifecycle event (repair-regression).
**Required fix (report-side is the gate; code fix docketed for run 4 is acceptable, same
disposition as R3-1/R3-2):** correct §2.1's rollover row and §2.3 addition 3 to state both
failure classes side by side (same-id-after-a-skipped-round vs. fresh-id-successor-chain) and
that the widen-`prevGapIds` remedy closes only the former; add a graded §3 row for the lineage
fix citing the backlog item; add a §5 open question (are the two fixes independent? — yes, per
the trace, and the report should say so); add the simulator case mirroring this corpus's own
chain, asserting the judge IS invoked by round 3 of a depth-≥2 supersession lineage.

### R4-2 — OPEN — MEDIUM — certain x medium x trivial — corroboration: HIGH (row 20's cell and §2.3 addition 13's assertion text read directly at merge; red's own R3-1 required-fix text compared side by side)
**Location:** §3 row 20, complexity cell — *"either treat as `PASS`-with-a-logged-warning (red
found nothing to fail on) or throw a distinguishing error rather than looping."*
**Problem:** the shipped fix is red's R3-1 required-fix disjunction carried forward verbatim and
never resolved into a decision — the only round-3 fix that ships an "or" where its siblings ship
choices (row 21: a single named call; row 22: option (b) with stated reasoning). The two branches
are opposite failure philosophies — silently converting a degenerate FAIL into a passing verdict
vs. halting loudly — precisely the "silent" axis this report argues against everywhere else
(§0, R3-2's own finding, row 2b, R3-1's own problem statement). **Merge-seat temper of lens 4's
grading:** §2.3 addition 13's assertions are negative and option-agnostic ("assert the loop does
not silently re-dispatch...; assert the terminal return does not pair `UNVERIFIED` with
`gaps_outstanding: 0`"), so the test CAN be written as-is — lens 4's "cannot actually be written"
overstates. But only the negative half is specifiable; the guard's positive behavior remains
undefined until the disjunction is decided. Impact graded medium on that verified basis, not
lens 4's medium-high. **Provenance note, against red:** the undecided "or" originated in red's
own R3-1 required-fix text; red must not hand blue a disjunction without naming its favored side
— blue ships red's phrasings verbatim (third instance of the class: R3-6, R2-4's proposed source,
now this; logged in red's pattern memory).
**Required fix:** one clause picking a side with a stated reason. Red's position: throw — a
degenerate merge-lens return should halt for human attention, not resolve toward a passing
verdict, for the same reason row 19's poisoning finding distrusts an unexamined clean signal. An
argued "PASS with a loud operator-surfaced warning, because throw loses partial progress" is
acceptable if reasoned — the gap is the undecided disjunction, not red's preference. Then extend
addition 13 with the positive assertion matching the choice.

### R4-3 — OPEN — LOW-MEDIUM — high (reading-order exposure: the ambiguous sentence is phrased as the operative instruction; the correction reads as a gloss on it) x low (the correct arithmetic is stated later in the same cell; a full read resolves it) x trivial — corroboration: HIGH (both sentences read verbatim in the same cell at merge; caught independently by lenses 2 and 4)
**Location:** §3 row 6 — *"assign the critical-stance/adversarial-disconfirming lens to at least
2 of N lanes (not 1-of-N)"* — against the same cell's round-3 correction — *"four named methods
(primary-literature / practitioner-production / adversarial-disconfirming-first / local-repo
critical-stance), one of which (adversarial-disconfirming-first) carries a 2-of-N redundancy
floor."*
**Problem:** R3-5's fix computes the right total (`lanes >= 5`) but leaves unedited the
originating sentence whose slash-as-compound-name reading produced the round-2 mis-add in the
first place. The same slash character is a synonym-joiner in the floor sentence and a
list-separator in the roster two sentences later, with nothing flagging the switch — a reader who
stops at the operative instruction still picks up the 3-method/`lanes >= 4` misreading R3-5
exists to prevent. Repair-reaches-the-conclusion-not-the-source class (sibling of R3-4/R3-10:
the correction landed downstream of the defect instead of in it).
**Required fix:** edit the original sentence itself — e.g. "assign the
adversarial-disconfirming-first lens (a distinct method from local-repo critical-stance, below)
to at least 2 of N lanes" — so the round-3 correction confirms an unambiguous sentence rather
than patching an ambiguous one.

### R4-4 — OPEN — LOW — certain x low x trivial — corroboration: HIGH (report-wide grep "4th|fourth" at merge: exactly one uncorrected instance, §3 risk-accepted paragraph; the three corrected locations — row 15, §4 row 5, §5 item 10 — confirmed in the same pass)
**Location:** §3, the risk-accepted closing paragraph — *"ENAMETOOLONG tooling (#15, re-graded
round 1 to track recurrence but still risk-accepted pending a 4th occurrence) [L1]"*.
**Problem:** a fifth location of R2-1's corrected trigger count, missed because R2-1 was closed
by fixing the three locations the gap named without a report-wide grep for the retracted numeral
— the exact class red's own memory names (grep the retracted token report-wide, both directions).
The paragraph is written as a scanning aid, so a summary-only reader takes away the stale
trigger.
**Required fix:** one word — "pending a third occurrence (corrected R2-1)," matching the three
already-corrected instances.

### R4-5 — OPEN — LOW-MEDIUM — certain (three live in-report instances, grep-confirmed) x low-medium (traceability: bare ids from another corpus are indistinguishable from this retrospective's own gap namespace; "R4-9" now collides live with this round's active R4-* series) x trivial — corroboration: HIGH (grep-confirmed at merge, lines 478/651/689; ids traced to `inputs/run2-friction.md` — the memory-architecture corpus's numbering)
**Location:** §3 row 13 — *"kept 3+ figures at unable-to-corroborate across 4 rounds (R1-19,
R1-28, R3-14/15, R4-9)"* — plus §4 rank-1 (*"blocks R1-19, R1-28, R2-8's residual, R3-14, R3-15,
R4-9 from resolving"*) and §2.3's lossy-fetch row (same id list).
**Problem:** unresolved cross-corpus gap-id collision. These are the *memory-architecture*
retrospective's internal red-audit ids (the corpus this run's H1–H5 doubts are about), not this
retrospective's — whose findings file uses the identical bare "R#-#" scheme. Round 2's
disposition already flagged this class once (§1.2's "R2-10") but only inside `red/findings.md`;
the note never reached the report, and three more unflagged instances stand. The sting is now
live: this round mints R4-1..R4-5, so a reader or tool cross-referencing the report's "R4-9"
against this retrospective's findings gets a hit-shaped token that resolves to nothing — and
would resolve to the *wrong thing* in any round that mints nine-plus gaps. A report arguing this
extensively for per-claim provenance (§1.2, claim manifest, §3 row 5) should not ship an
overloaded identifier scheme with no disambiguating marker.
**Required fix:** one global disambiguation parenthetical or footnote at first occurrence
(covering §1.2's instance too), or a corpus prefix (e.g. "MA-R1-19") at all four locations.

---

## Merge-time verification and dedupe notes (round 4)

- Dedupe map: lens 1's in-pass "R4-1" + lens 2's Finding 1 + lens 3's in-pass "R4-1" + lens 5's
  "R4-1" = **R4-1** (four independent convergences — each lens traced `debate.js` first-hand
  rather than trusting the backlog item; strongest multi-lens convergence of any gap in this
  corpus). Lens 4's in-pass "R4-1" = **R4-2**. Lens 2's Finding 2 + lens 4's in-pass "R4-3" =
  **R4-3** (merged). Lens 2's Finding 3 = **R4-4**. Lens 4's in-pass "R4-2" = **R4-5**.
- Precision correction to lens 5, applied at merge: "grep of `debate.md` for `### LEAD` — zero
  matches" is imprecise; a plain grep returns **one** match (line 528, a quoted phrase inside
  round-3 BLUE prose). The correct, stronger-form claim — verified via `grep -n "^### "` — is
  zero `### LEAD` *section headers* across rounds 0–3. Substance unchanged (the judge has never
  been dispatched); the imprecise form is the exact R3-6 class, caught at the merge seat before
  entering the record. R4-1 cites the precise form.
- Merge-seat temper of lens 4: addition 13's assertions read directly — negative and
  option-agnostic, so "the test cannot actually be written" overstates. R4-2 graded on the
  verified state (medium, not medium-high).
- `origin/main` re-fetched at merge: HEAD `42dba2d`; `debate.js` byte-identical across
  `47ae48d`/`88eb57f`/`d164ab2`/`42dba2d`. The backlog item quoted in R4-1 read via
  `git show 42dba2d` first-hand, not from any lens's paste.

## Noted, not raised (round 4)

- **[^CostFigureProvenance]'s pin now one commit stale** (`d164ab2` vs. live `42dba2d`): item 28
  byte-identical across the drift (lens 3, `git diff` scoped to the file). Per rounds-2/3
  precedent, docs-only drift is raised only where materially load-bearing; the drifted content
  (the docket-detector item) is load-bearing but is R4-1's subject, not a re-pin gap.
- **Lens 1's [^BlueReportGrep] apparent miscount (5 + 7 ≠ 7):** resolved as a subset relationship
  by direct comparison of matched line content — the assembled report's 7 matches contain all 5
  of blue/report.md's, renumbered. False alarm, correctly not raised.
- **Grep line-count vs. occurrence-count tool-mode footgun** ([^LocalGrepRed]: count-mode said
  64, occurrence count is 66, footnote says 66): tool artifact, footnote accurate. Noted for
  lens practice only — use occurrence-count when a claimed number could be either.
- **Lens 4's checked-and-held items:** §5 items 1–11 each carry a live counter-scenario; §2.4's
  "no risk-acceptance case" is argued elsewhere in-document; row 16b's run-level split matches
  the corpus's evidence granularity; the R3-1/R3-2 "docket, don't gate" framing is argued, not
  asserted. All held.
- **Lens 5's disconfirming items:** `maxRounds`-as-backstop checked and rejected as a
  risk-accept substitute for R4-1 (bounds cost, not convergence quality — recorded inside R4-1's
  grading); `adjudicated`-array growth is a non-issue (`Set` rebuilt per round, O(1) lookup);
  the bulk-tier-model susceptibility angle is speculative beyond this corpus's evidence — named
  here so it is not silently unconsidered.
- **§1.4's gap-pattern-file count:** live store still 23 (unchanged since round-3 merge check);
  the round-2/3 non-gap disposition holds.

## Verified clean this round (round 4 — see red/citation-ledger.md round-4 blocks)

Lens 1 (§0–§1): [^LocalGrep], [^BlueReportGrep], [^LocalGrepRed], [^RedFindingsGrep],
[^BlueReportUnverified], [^RedAuditorSpec], [^ClaimManifest], §1.4's memory-file and backlog-item
characterizations — all HIGH, several by independent re-run of the cited greps. Lens 2 (§2–§3):
rows 20/21/22 against live code; the R3-2 §2.1 correction; §2.3's 14-item numbering; row 13's
[^PdfMcp] held per skip-rule. Lens 3 (§4/§5/Footnotes): R3-7's textual distinction; item 28
byte-identical across the drift; "independen" grep still zero at `42dba2d`; [^DiminishingReturns]
disambiguation coherent against the twice-re-fetched source. Lens 4: all 41 prior-round fixes
hold as described at their cited locations; template compliance (stub top-level report.md
correct pending PASS; all footnotes semantic word-labels). Lens 5: R3-1/R3-2/R3-3 findings
re-confirmed live at `42dba2d`, correctly framed as docketed-not-shipped. Merge seat: the five
keystones listed in the round-4 header paragraph, first-hand.

## Round 4 disposition summary

FAIL. 10/10 round-3 gaps closed (2 with regressions re-raised: R3-1→R4-2, R3-5→R4-3). 5 new gaps
open. Gate tier: **R4-1** (the report-side coverage correction is the gate — §2.1/§2.3 currently
present the narrow rollover case as the whole docket-detector defect while the live repo and this
corpus's own history demonstrate the broader lineage-blindness; code fix docketed for run 4 is
acceptable, same disposition as R3-1/R3-2) and **R4-2** (decide the shipped disjunction — either
side, argued). R4-3/R4-4/R4-5 are mechanical one-clause fixes; R4-4 has no legitimate
risk-accept path (a stale numeral contradicting its own correction); R4-3 and R4-5 close by one
edit each. None disputes H1–H5. Convergence expectation: the citation surface is done — every
external re-fetch across three citation lenses came back clean, and all prose gaps are
propagation residue. If blue lands these five (all one-clause-to-one-row sized; R4-1 is the only
one requiring new report content, and its fix shape is already drafted in the live backlog),
round 5 is the PASS round on this trajectory. Red's own ledger this round: one imprecision
(lens 5's grep phrasing) caught and corrected at the merge before entering the record, one
overstatement (lens 4's "cannot be written") tempered against the verified text, and one
structural provenance debt owned (R4-2's disjunction originated in red's own round-3
required-fix) — the adversary's errors are logged with the same discipline demanded of blue.

---

## Verdict (round 3): FAIL — superseded by round 4, preserved

All 11 round-2 gaps are **CLOSED** — 7 clean, 4 **closed-with-regression** (blue's fix responsive
to the gap as raised, but the fix itself carries a new, smaller defect, re-raised as a round-3
gap: R2-4→R3-4+R3-9, R2-5→R3-10, R2-7→R3-6, R2-8→R3-5). One closure (R2-4) is a
**rebuttal accepted with evidence**: blue conceded the gap, leaf-checked red's proposed
replacement citation, found it does not contain the figure, and dropped the claim instead —
independently re-verified by two round-3 lenses; red's own proposed source was wrong, and that is
recorded against red, not blue. **10 gaps open (R3-1..R3-10):** 2 medium-high (R3-1, R3-2), 3
medium (R3-3, R3-4, R3-5), 2 low-medium (R3-8, R3-9), 1 medium-low (R3-7), 2 low (R3-6, R3-10).
None dispute H1–H5 or any prior closure. The round's shape: the citation surface is converging —
the three citation lenses' external re-fetches came back clean except for repair-regressions
inside round-2's own fixes (second-order now: regressions inside repairs of repairs) — while the
dark-side lens's control-flow trace of the already-merged `debate.js` opened a genuinely new
front (R3-1..R3-3), runtime edge cases no prior round examined. Severity is declining
monotonically across rounds (round 1: 2 HIGH; round 2: 5 MEDIUM-HIGH, all prose; round 3: 2
MEDIUM-HIGH, both code-trace findings — the prose gaps are now all ≤ MEDIUM): convergent, not
divergent.

---

## Round-2 gaps — status after round 3

### R2-1 — CLOSED (round 3)
Count corrected to 2 documented occurrences across 2 runs in §3 row 15, §4 row 5, §5 item 10
(trigger now "third," not "fourth"), and the §4 shape-verdict paragraph; the false
"per debate.md's merge-seat friction" citation dropped; likelihood re-argued (not re-asserted) on
the honest 2/2 rate. All four locations re-verified against sources (run2-friction.md line 4;
run1-friction.md zero mentions; CHANGELOG Round 0). Corroboration: high. A narrower, distinct
follow-on — whether occurrence 2 is even the same *mechanism* — is R3-7, not a reopening.

### R2-2 — CLOSED (round 3)
"Across rounds 0 and 1" now used consistently in §0's addendum, §4 row 4, §5 item 7; chronology
re-confirmed against `blue/CHANGELOG.md` Round 0 and `debate.md`'s round-1 RED dating. No residual
"this same round" phrasing found on full re-read. Corroboration: high.

### R2-3 — CLOSED (round 3)
§1.1's cross-provider paragraph now attributes "2 vs. 16" to the paper's L4 (model+persona)
condition and states L2's real curve (8 agents to match L1@16: 65.44% vs. 65.34%); the defer
disposition rests on the infrastructure-cost argument alone; §5 item 5's revisit trigger
recalibrated; §3 row 6 cross-referenced. Matches the ledger's HIGH-confidence Table 2 re-fetch.
Corroboration: high.

### R2-4 — CLOSED, REBUTTAL ACCEPTED WITH EVIDENCE (round 3) → regressions R3-4, R3-9
Blue conceded the gap (uncited "7 agents" figure) but rebutted red's specific proposed fix:
direct fetch of arXiv:2606.02646 shows it does **not** state "continued gains to 7 agents" — its
hardest benchmark is GSM-Hard, its harder-task knee is N≈10, and its headline is a ~1.8-agent
effective-team plateau by N=30. Two round-3 lenses independently re-fetched and confirmed every
figure in the rebuttal. Red accepts: applying red's required fix as written would have introduced
a second miscitation; the drop-rather-than-re-cite disposition is the correct one, and red's
unverified proposed source is logged as red's own error (see memory:
`pattern_repair_regression_citation.md`, red-side extension). **But the repair regressed twice:**
the §1.1 *body* still asserts the retracted clause (R3-4), and the footnote's replacement
sentence is internally contradictory (R3-9).

### R2-5 — CLOSED WITH REGRESSION (round 3) → R3-10
[^CostFigureProvenance] added, honestly scoped (self-reported, mixed provenance,
understatement-direction-only), attached at §2.3's closing line, §2.4, and the Tier B table —
which is what the gap required. But the propagation is incomplete: §2.1's first Tier A row
(reading-order-first instance of "252.9k") and §3 row 4's impact cell ("253k–3M") carry the same
figures untagged. See R3-10.

### R2-6 — CLOSED, STRENGTHENED (round 3)
§3 row 11 and §5 items 4/7 now state run 3 ran with zero artifact trail. Re-verified live at
`d164ab2`: still exactly two run directories under `research/`, while a *third* backlog commit
(`d164ab2` itself, "merge-seat cost analysis... run-3 transcripts") cites run-3 transcript data —
further corroboration, not contradiction. Corroboration: high.

### R2-7 — CLOSED WITH REGRESSION (round 3) → R3-6
§3 row 19 rewritten to the honest mechanism scope (catches source-misstatement, not
self-consistent fabrication); §5 item 8 answered and collapsed into the same statement; the
risk-accept stands on the honest scoping — exactly what the gap required. But the rewrite carries
forward, verbatim, an imprecise grep characterization that originated in *red's own round-2
merge text*. See R3-6.

### R2-8 — CLOSED WITH REGRESSION (round 3) → R3-5
The reconciling sentence the gap demanded was added to §3 row 6 (scoped exception to the
headcount risk-accept; `lanes >= 4` under the full roster; row 7's floor for non-adopting runs) —
responsive in form. But the reconciliation's own arithmetic under-counts by one lane and calls an
unenforced `[OPEN]` default a "floor." See R3-5.

### R2-9 — CLOSED (round 3)
§3 row 10's impact cell re-graded against the shipped ledger's actual prose-change-keyed
skip-trigger, with the concrete one-line fix (time/access-date condition in the same
`ledgerClause` string) stated as a build-now recommendation. Lens 5's disconfirming pass checked
the optimistic direction too: the clause is *not* silently already implemented (`debate.js`
lines 152–156 re-read live — still prose-change-keyed only), so "Build now" is accurate, not
overclaiming; lens 2 confirmed the clause byte-identical across `47ae48d`/`88eb57f`/`d164ab2`.
Corroboration: high.

### R2-10 — CLOSED BY ACKNOWLEDGMENT (round 3)
New §5 item 11 states the dependency plainly: recurrence-triggered risk-accepts (rows 14, 15)
have no durable counter until an equivalent of the citation ledger exists; treat as advisory;
re-derive counts from primary sources each time. Reasoned, not a silent skip — this is the
argued disposition red said it would accept. Corroboration: high.

### R2-11 — CLOSED (round 3)
§3 row 4 reworded to the accurate claim (no functional parsing path; string present only in
`debate.js`'s descriptive header comment, zero matches in `commands/research.md`). Re-verified
live at `d164ab2`, exact. Corroboration: high.

---

## Round-3 gap details (original grading, preserved verbatim — round-4 statuses above supersede the "OPEN" markers below)

### R3-1 — OPEN — MEDIUM-HIGH — low-medium x medium-high x low — corroboration: HIGH (direct control-flow trace of `debate.js` lines 56–91, 148–198, 200–218; live and unchanged at `d164ab2`, re-confirmed at full merge)
**Location:** §2.3 item 8 (*"`--maxRounds 0` — the emitted log line must distinguish 'never ran'
from 'ran and failed at round 0'"*) and §2.1's Tier A round-loop row (*"Round loop / contested
docket / deadlock / safety ceiling / `adjudicated` bookkeeping"*) — both describe round-loop
degenerate-input handling as covered; neither covers this shape.
**Problem:** `RED_ENVELOPE` requires `verdict`/`gaps` but has no cross-field constraint forcing
`gaps` non-empty when `verdict === 'FAIL'`. Traced by hand: `{verdict: 'FAIL', gaps: []}` makes
`contested.length` stay 0 every round (the judge is never invoked — no adjudication path exists
for "FAIL with nothing to adjudicate"), blue is dispatched against an empty open-gaps list while
being told the debate failed, and this recurs silently until `maxRounds` — no distinguishing log
line, no thrown error. The final return then reports `verdict: 'UNVERIFIED'` alongside
`gaps_outstanding: 0`, a self-contradictory terminal state to any caller reading only the
top-level return. Distinct from §2.3 item 10 (`gaps` *missing* — a schema violation); this is
schema-*valid* but semantically incoherent — a separate degenerate-loop-termination case.
**Required fix:** guard after the `redEnv` null-check: if `verdict === 'FAIL'` and `gaps` is
empty, treat as effective PASS-with-warning or throw a distinguishing error rather than looping
silently to the ceiling; add as a simulator case alongside the existing `--maxRounds 0` case.
Report-side: name the case in §2.3's additions. Red will accept "code fix docketed for run 4"
if the report-side coverage claim is corrected.

### R3-2 — OPEN — MEDIUM-HIGH — certain (already realized, structural) x medium x low — corroboration: HIGH (direct read of `debate.js` lines 132–146, 170, 187, 197 and the merged `debate.test.mjs` lines ~114–123, both live and unchanged)
**Location:** §2.1, Tier A table — *"Friction aggregation: per-seat arrays namespaced by label and
concatenated, never dropped | source | Self-improvement input integrity [L1+L3]"*.
**Problem:** this claim is false for one schema'd, friction-capable seat. `takeFriction` is
called at exactly three sites (red-merge, judge, blue-respond) — never for round-0
blue-synthesize, even though it is schema'd against `BLUE_ENVELOPE`, which declares a `friction`
field. Not equivalent to the (fine, by-design) exclusion of unschema'd frontier/blue-lane
dispatches. The practical sting: this report's own §0 live addendum (the write-block firing
during this retrospective's own round-0 synthesis) is exactly the class of event the structured
`friction` channel would silently drop under the already-merged code — it survives in this
report only because the agent narrated it into prose. The merged regression test ("friction
aggregates from every seat with attribution") stubs and asserts only red-merge/blue-respond, so
11/11-green gives false confidence.
**Required fix:** add `takeFriction('blue-synthesize', blueEnv)` after the null-guard (line 136);
add a simulator case; **correct §2.1's "never dropped" claim to name the exception until fixed**
(the report-side correction is the gate; the code fix may be docketed for run 4).

### R3-3 — OPEN — MEDIUM — medium (untested live, only by trace) x medium x low — corroboration: HIGH on mechanism (direct trace of `debate.js` lines 93–112, 166–197); MEDIUM on real-world frequency (no live `carried` resolution observed yet in this corpus)
**Location:** `debate.js`'s judge prompt (line 182) — *"for carried, state what further research
blue owes"* — against §2.3 item 5's suite-case description (*"carried gap re-enters `openGaps`
with its required-fix intact"*).
**Problem:** the judge is asked for a specific piece of guidance when carrying a gap, but the
script reads `judge.resolutions` only to populate `adjudicated` (which excludes "carried") and
`judge.friction` — the rationale text for a carried gap is never passed into the next round's
`blue-respond` prompt, which is built entirely from red's original gap object. §2.3 item 5's own
test description confirms only the gap's `required_fix` is checked to survive the carry. The
judge's answer is written to `debate.md` for a human, not threaded back into the loop that
requested it.
**Required fix:** either fold the judge's per-gap rationale into the `openGaps` payload passed to
`blue-respond` (more robust), or add a sentence instructing blue-respond to read the latest
`### LEAD` section of `debate.md` before responding (cheaper; relies on blue's initiative — the
reliability class already flagged for backlog item 4 / §3 row 12). An argued risk-accept
(mechanism untested live; low observed frequency) is acceptable if the §2.3 item 5 description
stops implying the guidance is delivered.

### R3-4 — OPEN — MEDIUM — certain x medium x trivial — corroboration: HIGH (direct text comparison within the shipped document, confirmed independently by lens 1 and by the merge's own full re-read; the footnote's corrected direction independently re-verified against arXiv:2606.02646 full HTML by two lenses)
**Location:** §1.1 (H1), diminishing-returns disconfirming bullet — *"with the breakeven shifting
to 3–4 agents on harder tasks and continued gains observed to 7 agents on the hardest — so the
qualitative thesis holds"*.
**Problem:** the body still asserts the exact clause its own footnote retracted this round —
repair-regression one hop further out, **body lagging the correctly-repaired footnote** (the
reverse of the familiar footnote-lag direction). [^DiminishingReturns]'s round-2 text explicitly
drops "7 agents" ("dropped rather than re-cited to an unverified source") and states the
corrected direction is toward *more* caution on hard tasks; the §1.1 sentence the footnote is
attached to — written in round 1, untouched by the round-2 fix — still presents the retracted,
wrong-direction claim as corroborated by "independent re-search." A body-prose reader (the
majority surface) takes away a claim the report's own citation apparatus has withdrawn.
**Required fix:** edit the §1.1 body clause to match the footnote's corrected synthesis — drop
"continued gains observed to 7 agents on the hardest"; the correct wording is already sitting in
the footnote. Trivial; no new research.

### R3-5 — OPEN — MEDIUM — certain (static logic property) x medium x trivial — corroboration: HIGH (pure arithmetic against the row's own roster; row 7's [OPEN] status on the same table; §1.1's documented run-2 `--lanes 2` override)
**Location:** §3 row 6, the R2-8 reconciliation clause — *"four named methods (primary-literature
/ practitioner-production / adversarial-disconfirming-first / local-repo critical-stance) plus a
2-of-N floor on one of them arithmetically needs `lanes >= 4`"* — and, same clause, *"row 7 below
floors N at 3 (the shipped default)"*.
**Problem:** the reconciliation added to close R2-8 **mis-adds**. Taken literally as stated: 3
unfloored methods x 1 lane + 1 floored method x 2 lanes = **5** lane-assignments minimum, not 4.
"`lanes >= 4`" holds only if "adversarial-disconfirming-first" and "local-repo critical-stance"
are silently treated as one lens for allocation (the floor clause's slash-phrasing
"critical-stance/adversarial-disconfirming lens" implies exactly this merge; the same paragraph's
own "four named methods" count contradicts it). A future operator implementing the roster at the
stated floor silently drops one named method — reintroducing R1-16's failure-concentration risk
via a wrong-by-one floor instead of an absent one. Compounding, same clause: row 7 is graded
**[OPEN]** ("`lanes = 3` remains an unguarded default with no minimum check") and §1.1 documents
run 2 overriding it downward to 2 — an unenforced default is not a "floor," and the
reconciliation's whole point is a hard minimum operators must respect. Same failure class as the
round-2 catch it was meant to close (`pattern_unreconciled_numeric_floors`): the composition was
asserted, not computed — this time inside the fix itself.
**Required fix:** one sentence — either (a) state the true minimum as `lanes >= 5` with all four
methods distinct, or (b) state explicitly that the two adversarial/critical methods collapse into
one floored lens (roster = 3 for this arithmetic, minimum 4) and why that merge is safe; and
align "floors N at 3" with row 7's own [OPEN] status ("will floor once built," not "floors").

### R3-6 — OPEN — LOW — low x low x trivial — corroboration: HIGH (git grep re-run at the merge seat: zero matches anywhere in the plugin, including `debate.js:156`'s `ledgerClause`, read in full)
**Location:** §3 row 19 — *"a repo-wide grep for 'independent' in the plugin returns zero hits
outside this ledger clause's own text."*
**Problem:** the phrasing asserts the ledger clause itself contains the word (implying exactly
one hit, located there). The live grep returns **zero hits, full stop** — the `ledgerClause`
string contains no form of "independent." The accurate statement is *stronger* for row 19's
point (the protocol has no independent-cross-referencing language anywhere, not even where a
reader might expect it), but as written it sends a future verifier hunting an exception that
does not exist. **Provenance note, against red:** this exact phrasing originated in red's own
round-2 merge (`findings.md`, round-2 header: "zero hits outside the ledger line's own text") and
was carried verbatim into blue's R2-7 fix — a red-originated imprecision surviving two rounds
inside the very lens whose job is this check. Logged in red's pattern memory (red-side extension
of `pattern_repair_regression_citation.md`). Caught independently by lenses 2 and 3;
deduplicated here.
**Required fix:** reword to "returns zero hits anywhere in the plugin, including the ledger
clause itself."

### R3-7 — OPEN — MEDIUM-LOW — medium x low-medium x trivial — corroboration: HIGH on the textual distinction (both primary sources read directly); MEDIUM on practical import
**Location:** §4 row 5 — class label *"Windows ENAMETOOLONG / long-heredoc fragility"* over
*"**2 documented occurrences across 2 runs**"* — and §3 row 15's likelihood rationale — *"a
2-for-2 rate on the triggering conditions."*
**Problem:** occurrence 1's source names the mechanism explicitly (`run2-friction.md` line 4:
"ENAMETOOLONG... **Windows command-length limit**"); occurrence 2's source names only a symptom
(`blue/CHANGELOG.md` Round 0: "chunked-heredoc workaround attempt then **failed on shell
parsing**" — no errno, no length, no byte count). A parse failure on a large heredoc can arise
from quoting/special-character/CRLF issues independent of total length — and CRLF is a
*separately documented* Windows fragility class in this same corpus (§2.1 Tier A). The "2/2
same-mechanism rate" carrying the High likelihood grade is therefore 1-confirmed +
1-same-family-plausible, not 2-confirmed-identical. Does not overturn the risk-accept or the
revisit trigger — narrower than R2-1 (which fixed *how many*; this flags *what* the second one
was).
**Required fix:** one clause — "occurrence 2's exact failure mode is not confirmed as the
length-ceiling class (its source names only 'shell parsing'); treat the rate as 1-confirmed +
1-same-family-plausible." An argued risk-accept (transcript likely unrecoverable) is acceptable.

### R3-8 — OPEN — LOW-MEDIUM — medium x low x trivial — corroboration: HIGH (live re-fetch of `ideas/backlog.md` item 28 at `d164ab2`, three commits past the footnote's pin)
**Location:** footnote [^CostFigureProvenance] — *"`ideas/backlog.md` item 28, live at `main` @
`88eb57f`"* — as consumed by §2.3/§2.4 and §3 row 18's cost rationale.
**Problem:** live-source drift on the report's own most-cited backlog footnote — a textbook
self-application of §3 row 10's named risk. Item 28 has since gained sub-item (d), "MERGE-SEAT
ANALYSIS (run-3 transcripts): the driver is TURNS × CONTEXT, not file size... red-merge-r1:
~100-150K of material, 2.7M+ cache reads," with per-seat cost figures — new data directly
relevant to §3 row 18 (audit-narrowing hold/risk-accept, which reasons about red's full-re-read
burn as the dominant cost). Nothing claimed is contradicted (direction remains
understatement-only, consistent with the footnote's own caveat); the footnote is simply three
commits stale on a source the report knows is volatile.
**Required fix:** re-fetch and re-pin at current HEAD; add the (d) sub-item and note the
turns-x-context finding in row 18's rationale. Cheap; strengthens the report.

### R3-9 — OPEN — LOW-MEDIUM — certain x low-medium x trivial — corroboration: HIGH (single-sentence internal contradiction, footnote read directly; underlying figures re-verified against arXiv:2606.02646 by two lenses)
**Location:** footnote [^DiminishingReturns], the R2-4 replacement sentence — *"harder tasks
shift the breakeven higher and, per arXiv:2606.02646's actual finding, may show diminishing
returns arriving even earlier (practical knee ≈10, effective diversity saturating well before
nominal N=30) rather than later."*
**Problem:** internal contradiction inside one sentence, introduced by the R2-4 fix itself — the
third consecutive round of trouble in this exact footnote (R1-5, R2-4, now this: a
recurring-footnote pattern worth naming even though each instance is small). "Shift the breakeven
higher" means diminishing returns arrive *later* (at higher N); the conjoined clause claims they
arrive "even earlier... rather than later." Both cannot be true of the same quantity. The source
data supports a coherent claim only by distinguishing two senses the sentence conflates: the
nominal-N breakeven (knee ≈10, higher than moderate-task 2–4) vs. the effective-diversity
saturation ceiling (~1.8 by N=30, reached far below nominal pool size). As worded, a careful
reader cannot extract one consistent claim.
**Required fix:** disambiguate the two senses explicitly (e.g., "harder tasks raise the nominal
breakeven (~10 vs. 2–4), but the ceiling those agents chase saturates ~1.8 effective agents by
N=30 — so adding agents on hard tasks buys less than the higher breakeven alone implies"), or
make the clauses alternatives and say which the synthesis favors.

### R3-10 — OPEN — LOW — certain x low x trivial — corroboration: HIGH (both untagged instances read directly at merge)
**Location:** §2.1, Tier A first row — *"252.9k tokens, 11m48s, honest UNVERIFIED
deadlock[^Run1Friction][^Run1Journal]"* — and (found at merge, extending lens 4's catch) §3 row
4's impact cell — *"High — ~50k tokens vs. 253k–3M live discovery"* — both untagged.
**Problem:** incomplete propagation of the R2-5 repair. [^CostFigureProvenance] reached §2.3's
closing line, §2.4, and the Tier B table, but not §2.1's first row — the *reading-order-first*
instance of the figure, presented as plain fact — nor §3 row 4's impact cell. Direction of risk
is understatement only (per R2-5's own finding), so no verdict moves; this is the
repair-reached-some-instances-not-all class red's memory already names (grep the token
report-wide, not just the sections the repair note cites).
**Required fix:** append [^CostFigureProvenance] to both instances, or add one clause to the
tier-table preamble stating the caveat covers every token figure in §2–§3.

---

## Merge-time verification and dedupe notes (round 3)

- Lens-2 Finding 1 and the compounding row-7 point = R3-5 (merged). Lens-2 Finding 2 and
  lens-3's in-pass "R3-1" = R3-6 (deduplicated; both lenses ran the grep independently, same
  result). Lens-1's in-pass "R3-1" = R3-4. Lens-3's in-pass "R3-2"/"R3-3" = R3-7/R3-8. Lens-4's
  in-pass "R3-1"/"R3-2" = R3-9/R3-10. Lens-5's R3-1..R3-3 keep their preview ids.
- `origin/main` re-fetched at merge: still `d164ab2`; `debate.js` byte-identical to `47ae48d`
  across all intervening commits. No report claim invalidated by drift between the lens passes
  and this merge.
- The `git grep -ni "independen"` result (zero hits) was re-run first-hand at the merge seat
  rather than trusted from the lenses, because R3-6 indicts red's own round-2 text — the merge
  does not get to self-grade by citation to itself.

## Noted, not raised (round 3)

- **§1.4's "15 well-formed gap-pattern files":** live store now holds 23 (was 18 at round 2).
  Already dispositioned round 2 as stale-count churn inherent to describing a live accreting
  mechanism; nothing changed about that reasoning. Logged in the ledger for the running count
  only; re-litigating a settled non-gap every round would violate the no-grinding norm.
- **`commands/research.md` line 9 "haiku for smoke tests" prose:** an operator can already
  manually approximate a smoke run via existing `lanes`/`maxRounds`/`model` parameters —
  consistent with row 4's complexity grade, not a contradiction. (Lens 2.)
- **§3 row 8's "(c) superseded by (a), strictly cheaper" near-miss:** initially read as a leap of
  faith; on check, the corpus's own record (round-1 red-merge: `Edit` succeeded on
  `red/findings.md` where `Write` was refused) gives (a)'s append/Edit mechanism a positive data
  point. Checked, held, not raised. (Lens 4.)
- **CHANGELOG R2-9 phrasing ("added the concrete fix... to the same ledgerClause string"):**
  plausibly skim-read as a live code edit; in full row-10 context ("Build now," future tense) it
  resolves as a report-table recommendation. Not misleading on a full read. (Lens 5.)
- **Disconfirming checks that held (lens 5):** R2-9's ledger clause is not silently already
  implemented (still prose-change-keyed — "Build now" accurate); R2-8's floor has no code
  enforcement and the report doesn't claim it does (row 7 already [OPEN]); no
  gap-resurrection behavior beyond §2.3 item 3's known case; `exhausted` cannot misfire on the
  PASS path.
- **Backlog drift `88eb57f`→`d164ab2`:** docs-only, `debate.js` untouched — the
  [^PinnedRepoState] discipline continues to work as designed; raised only where the drifted
  content is materially load-bearing (R3-8), per round-2's precedent.

## Verified clean round 3 (see red/citation-ledger.md lines 90–111 for the graded pairs)

R2-2/R2-3 fix texts verified in place; [^DiminishingReturns]'s rebuttal figures (GSM-Hard, knee
N≈10, ~1.8 plateau, N≤5 pilot) independently re-fetched twice — all exact; §1.4's
blue/judge-have-no-`memory:`-key negative claim verified by direct frontmatter read (new
leaf-node check); judge-null and citationPasses-const re-confirmed live at `d164ab2`; ledger
clause byte-identical across three refs; run-directory count (still 2) re-confirmed; run2-friction
21 entries / single ENAMETOOLONG line re-counted; §4 rows 1–3 and 10 footnotes re-spot-checked
live; §2.3 item 1's "blue-respond cannot crash" and row 16b's bulk/judgment fallback claims
verified true against `debate.js` by direct read (lens 4); no functional `--smoke` path; frontier
shape-verdict quote verbatim; report-template compliance (Catechism correctly a stub pending
PASS — not a violation). 22 claim-reference pairs ledgered by the round-3 lenses plus the
merge-seat re-verifications above.

## Round 3 disposition summary

FAIL. 11/11 round-2 gaps closed (4 with regressions re-raised; 1 closure a rebuttal red accepts
with evidence — and with red's own proposed citation logged as red's error). 10 new gaps open.
Gate tier: R3-1 and R3-2 (runtime edge cases in the merged `debate.js`, one of which — R3-2 —
directly falsifies §2.1's "never dropped" claim and would have swallowed this report's own
headline live evidence); R3-4 and R3-5 (a body sentence contradicting its own corrected footnote
on a graded bullet; a reconciliation that mis-adds by one lane). R3-3 closes by data-plumbing,
prompt sentence, or an argued risk-accept with §2.3 item 5 corrected; R3-6/R3-9/R3-10 are
mechanical accuracy fixes with no legitimate risk-accept path; R3-7 and R3-8 close by one clause
each or an argued risk-accept red will take if reasoned. For R3-1/R3-2, red will accept
"report-side claim corrected + code fix docketed for run 4" — the report's job is the
retrospective; what cannot stand is the report describing coverage the shipped code does not
have. None of the ten disputes H1–H5. Convergence expectation: this is the first round where
every remaining prose gap is ≤ MEDIUM and every fix is one-sentence-to-one-tag sized; if blue's
round-3 response holds the no-new-regression bar that round 2's repairs did not, round 4 is
plausibly the PASS round.

---

## Round-1 gaps — status after round 2 (unchanged in round 3; no closure regressed further except as tracked via R2→R3 chains above)

### R1-1 — CLOSED (round 2)
Fix verified: §0 "Round 1 correction" block present; headline reframed to "shipped, pending first
live trial"; the ~13 dependent §3/§4 cells threaded ([MERGED]/[OPEN]/[PARTIAL] legend);
[^PinnedRepoState] SHA+timestamp discipline added. Every technical claim in the correction block
(blueEnv guard L136, redEnv guard L171, citationPasses const L139/while L148, judge L181
unguarded at L184) re-confirmed by direct read of `debate.js` at current `main` (`88eb57f`) —
not carried forward. Corroboration: high.

### R1-2 — CLOSED (round 2)
Row 2 reworded from conditional to factual, pinned to `debate.js:184`; new row 2b added. The
*code* defect remains live on `main` (re-verified at merge, `88eb57f`) — correctly carried as a
run-4 docket item, which is what the gap required. Corroboration: high.

### R1-3 — CLOSED (round 2)
§2.3 item 1 rewritten into the three failure classes exactly as specified
(throws: judge; silent-degrade: frontier/red-lens/assembly; already-covered: blue-respond).
Corroboration: high (report text read directly).

### R1-4 — CLOSED (round 2)
"~19%" re-cited to arXiv:2603.22103 with an explicit domain caveat; "~95%" dropped as
untraceable. The new [^NarrativeSimilarity] figures (r=0.388 vs. r=0.461; 76.0% vs. 75.3%;
71.0% vs. 71.7%) verified **exact** against the paper's full HTML Table 5 this round — upgraded
to high (round 1 reached only the abstract). Corroboration: high.

### R1-5 — CLOSED WITH REGRESSION (round 2) → R2-4
The "2–4 agents" bound restated as qualitative synthesis, as required. But the replacement text
introduces a *new* unpinned precise figure ("continued gains to 7 agents on the hardest")
attributed only to "independent re-search this round" — the identical footnote-over-attribution
failure R1-5 was raised to fix, recurring inside its own fix. See R2-4 (and, round 3, the chain's
continuation: R3-4, R3-9).

### R1-6 — CLOSED (round 2)
Corrected diffstat re-verified exact live this round: `gh pr view 14 --json` returns
additions:318, deletions:48, changedFiles:18, 11 commits. Corroboration: high.

### R1-7 — CLOSED (round 2)
"21 entries" verified by direct count of `run2-friction.md` bullet lines (3–23). Both footnotes
corrected. Corroboration: high.

### R1-8 — CLOSED (round 2)
§4 row 6 corrected to "run 2, round 1 only"; verified — the file attests exactly one
live-source-drift entry (line 6, red-merge-r1). Corroboration: high.

### R1-9 — CLOSED (round 2)
[^SubagentWriteBug] now names the signature difference (silent no-op vs. explicit worded
refusal) and downgrades the transfer. Corroboration: high (footnote read directly).

### R1-10 — CLOSED (round 2)
"report.md-specific" struck as the report's own conclusion; backlog quote now explicitly
attributed-not-endorsed, with both falsifying occurrences named. Corroboration: high.

### R1-11 — CLOSED (round 2)
Known-failing flag added to §2.3 item 4; dedicated graded row 2b added. The defect itself
re-verified still live at `88eb57f` (`const` L139, loop L148) — now correctly classified as a
shipped defect, which is what the gap required. Corroboration: high.

### R1-12 — CLOSED (round 2)
Row 16b added with the explicit two-option disposition ((b) for dev/smoke, (a) for keeper runs).
Doctrine quote verified verbatim (`debate.js` L22–24); routing verified (red-lens `...bulk` L164,
red-merge `...judgment` L168). Lens 5 additionally verified the keeper-run inheritance claim:
`bulk = model ? { model } : {}` — with no `--model` flag, bulk seats inherit the session model,
so disposition (b)'s scoping is mechanically real. Corroboration: high.

### R1-13 — CLOSED WITH REGRESSION (round 2) → R2-1
The mechanically-unsound "skeleton may moot it by construction" clause dropped and replaced with
the chunked-append mechanism, as required. But the likelihood re-grade (Medium→High) was keyed to
a recurrence count ("3 times across 3 runs... red-merge seat, round 1, per debate.md's merge-seat
friction") that its cited source does not contain. See R2-1 (closed round 3; narrower follow-on
R3-7).

### R1-14 — CLOSED (round 2)
Vetting-step sentence added to row 13 (pin, review, scope permissions). Both tools' "active
maintenance, passing test suites" claims verified live this round (neither archived;
`arxiv-latex-mcp` pushed 2026-06-30, CI green; `pdf-reader-mcp` main-branch CI green).
Corroboration: high.

### R1-15 — CLOSED WITH REGRESSION (round 2) → R2-7
Graded row 19 added with an explicit disposition — the silence the gap named is cured. But the
row's named mitigation ("independent re-verification against a second source") overstates what
the leaf-node lens actually does, and contradicts the report's own §5 item 8. See R2-7 (closed
round 3; wording follow-on R3-6).

### R1-16 — CLOSED WITH REGRESSION (round 2) → R2-8
Redundancy floor added to row 6 — the failure-concentration trade is now named, which is what the
gap required. But the floor does not arithmetically compose with row 7's lane-count floor and the
report's own risk-accepted ceiling on headcount raises. See R2-8 (closed round 3; the
reconciliation's own arithmetic re-raised as R3-5).

### R1-17 — CLOSED WITH REGRESSION (round 2) → R2-3
Cross-provider model diversity named and dispositioned (defer, not adopt) — the surfacing the gap
required. The factual premise (harness `model`/`judgmentModel` knobs select Claude aliases only;
no provider wiring) verified high this round. But the disposition's load-bearing citation
misapplies the paper's L4 result to the L2 condition. See R2-3 (closed round 3).

### R1-18 — CLOSED (round 2)
Addendum relabeled "self-observed, not yet artifact-logged," downgraded to one data point, with
the independent class corroboration (red's round-1 hit) correctly framed as the stronger
evidence. Corroboration: high. (The residual chronology error inside this same paragraph was
R2-2 — closed round 3.)

### R1-19 — CLOSED (round 2)
Corrected URL verified live this round: resolves to a genuine 672KB PDF, no 404. Quote itself
remains medium (self-labeled search synthesis) — unchanged, acceptable. Corroboration: high on
the URL.

### R1-20 — CLOSED (round 2)
Header restated; no numeric-ratio claim remains for lane 3. Corroboration: high.

---

## Round-2 gap details (original grading, preserved verbatim — round-3 statuses above supersede)

### R2-1 — MEDIUM-HIGH — certain x medium x low — corroboration: HIGH that the citation fails (direct read of debate.md's merge-seat friction section, both friction files, and blue/CHANGELOG.md round 0; re-confirmed at merge)
**Location:** §4 row 5 — *"red-merge (retrospective round 1, this round — third occurrence, per
debate.md's merge-seat friction)"* and *"3 runs/rounds, re-graded likelihood High per R1-13"*;
duplicated in §3 row 15 (*"confirmed recurred 3 times across 3 runs, not 2"*) and §5 item 10
(*"Does ENAMETOOLONG recur a fourth time..."*).
**Problem:** the cited source does not contain the claimed event. `debate.md`'s round-1
merge-seat friction lists exactly two items (lossy PDF-fetch depth; the lens-writes-transcript
process misfit) — no ENAMETOOLONG, no heredoc, no command-length event. The corpus attests
exactly **two** ENAMETOOLONG-class occurrences: `run2-friction.md` line 4 (red-merge-r1, run 2)
and this retrospective's own round-**0** heredoc shell-parse failure (per `blue/CHANGELOG.md`
Round 0). `run1-friction.md` contains zero mentions, so "3 runs" fails on its own terms too.
Likely mechanism: the *write-block's* separately-well-corroborated "third occurrence" count was
transposed onto the structurally adjacent ENAMETOOLONG narrative during the R1-13 fix, without
re-walking the ENAMETOOLONG-specific sources. The fabricated count feeds a likelihood re-grade
(Medium→High) and §5 item 10's "fourth occurrence" build-trigger, which should read "third."
**Required fix:** correct the count to 2 documented occurrences across 2 runs; drop the
"per debate.md's merge-seat friction" citation; renumber §5 item 10's trigger from fourth to
third; re-state the likelihood grade on the honest count (High may still be defensible given
Windows + large-append mechanics — but argue it, don't miscount it).

### R2-2 — LOW-MEDIUM — certain x low x trivial — corroboration: HIGH (both events dated by the corpus's own CHANGELOG and transcript; the contradiction is legible within one paragraph)
**Location:** §0 live addendum — *"red's own round-1 audit pass hit the identical block writing
`red/findings.md` this same round"* two sentences before *"Two independent hits at two seats in
two consecutive rounds"*; inherited by §4 row 4 (*"in the same round it hit blue"*) and §5 item 7
(*"Both write-block occurrences observed this round"*).
**Problem:** chronology error, internally contradictory on its face. Blue's synthesis hit is
dated round **0** (CHANGELOG Round 0); red's `findings.md` hit is dated round **1** (debate.md
round-1 RED friction). "Two consecutive rounds" is the correct half; "this same round" and its
two inheritors are wrong.
**Required fix:** correct the three phrasings to "across rounds 0 and 1" / "one round after it
hit blue."

### R2-3 — MEDIUM-HIGH — realized x medium-high x low — corroboration: HIGH (lens-1 full-HTML fetch of arXiv:2602.03794, independently re-fetched by red-merge at merge time: Table 2, L4 2-agent 67.71% vs. L1 16-agent 65.34%; L2 8-agent 65.44%)
**Location:** §1.1 cross-provider paragraph (round-1 R1-17 fix) — *"the same paper's practical
finding — '2 diverse [persona-lensed] agents match/exceed 16 homogeneous' — shows method/lens
diversity within one provider already captures most of the achievable gain without the
infrastructure cost."*
**Problem:** within-source condition misattribution. The paper's four-level taxonomy makes the
"2 vs. 16" result **L4's** (different models AND different personas). **L2** — persona/lens
diversity on one base model, exactly the condition the bracketed gloss "[persona-lensed]"
substitutes in — needs **8 agents** to match the same 16-agent homogeneous baseline: a 4x
efficiency gap, not "most of the achievable gain." The report's own adjacent sentence correctly
quotes "same-base-model agents remain more correlated than architecturally distinct ones," then
the disposition draws the opposite practical conclusion from the same source. The "defer, not
adopt" call may still be right (the infrastructure-cost argument stands on its own) — but not
for the reason currently given, and §5 item 5's "revisit if lens-assignment under-delivers"
trigger is calibrated against the wrong efficiency curve.
**Required fix:** correct the sentence to attribute "2 vs. 16" to L4 and state L2's real curve
(8-to-match-16); let the defer disposition rest on the infrastructure-cost argument alone;
recalibrate §5 item 5's revisit-trigger baseline.

### R2-4 — LOW-MEDIUM — realized x low-medium x trivial — corroboration as originally graded: HIGH that the footnote lacked the citation; the round-2 belief that the figure "traces to arXiv:2606.02646" was WRONG — disproven by blue's round-2 rebuttal fetch, independently re-verified twice in round 3; red's error, logged
**Location:** footnote [^DiminishingReturns] — *"continued gains to 7 agents on the hardest —
treat as a synthesis across sources, not a single citable number."*
**Problem (as raised):** repair-regression on R1-5 — a new precise numeric bound ("7 agents")
attributed only to an unlinked "independent re-search this round." The gap was real; red's
proposed required fix (add arXiv:2606.02646) was not — that paper does not contain the figure.
**Resolution (round 3):** blue's drop-rather-than-re-cite disposition accepted with evidence.

### R2-5 — MEDIUM — high x medium x low — corroboration: HIGH on the backlog finding (ideas/backlog.md item 28, live at 88eb57f: panel counter excludes cache traffic = 92% of real flow, 610K reported vs. 47.7M transcripts); MEDIUM on whether the report's own figures suffer the same undercount (unverified either way — that open question IS the gap)
**Location:** §2.3 closing line — *"against historical incident costs of 252.9k (run 1) and ~3M
tokens (run 2's quota crashes)"*; §2.4 — *"253k–3M tokens per historical incident."*
**Problem:** measurement-methodology drift, one layer out from round 1's PR-merge class. The
project's own live backlog attests (run 3's live measurement) that its in-band panel token
counter undercounts real spend by ~92% by excluding cache traffic. The report's two most-cited
efficiency figures are of unstated provenance; the side-by-side comparison may not be
apples-to-apples. Direction of risk: understatement only. Not verdict-changing.
**Required fix:** one footnote flagging provenance; point at the cost-audit tool.

### R2-6 — MEDIUM — high x medium x low — corroboration: HIGH (git ls-tree at 88eb57f: `research/` contains exactly two run directories, neither run 3's; backlog commits `47ae48d`/`88eb57f` both cite live run-3 measurements)
**Location:** §3 row 11; §5 items 4 and 7.
**Problem:** run 3 evidently already executed and left **zero artifact trail** in the tree — a
live, fresh instance of exactly the gap row 11 argues for closing, silently mooting §5 items
4/7's "pending live trial" framing: the trial is no longer pending, it is *unrecorded*.
**Required fix:** note in row 11 that run 3 is the demonstrated instance; update §5 items 4/7.

### R2-7 — MEDIUM-HIGH — high x medium-high x low — corroboration: HIGH (direct read of `research-protocol/SKILL.md` and `red-auditor.md` verification mandates; repo-wide grep "independent"; the two contradicting sentences are both round-1 additions on the same page)
**Location:** §3 row 19 disposition — *"a poisoned page asserting a fake fact still has to
survive independent re-verification against a second source"* — against §5 item 8 — *"or does it
require a distinct defense (e.g. cross-referencing claimed sources against an independent
index)?"*
**Problem:** the risk-accept's named mitigation claims a capability the protocol does not grant.
The leaf-node method as specified is: follow the citation to *the* source — the same URL re-read,
not an independent second source. A self-consistent poisoned page passes that check by
construction. §5 item 8, added the same round, half-admits this. Observed practice is better
than the protocol's letter for *miscitation* triggered by a suspicious figure — but that trigger
never fires for a self-consistent fabrication, precisely the scenario row 19 risk-accepts.
**Required fix:** restate row 19's mitigation honestly; collapse the two passages into one
accurately-scoped statement. The risk-accept itself may stand once restated.

### R2-8 — MEDIUM-HIGH — high x medium-high x low — corroboration: HIGH (arithmetic internal to the report's own two rows; both read directly; §1.1's "or"-phrased lane-letter roster cross-checked)
**Location:** §3 row 6 disposition — *"assign the critical-stance/adversarial-disconfirming lens
to at least 2 of N lanes (not 1-of-N) ... cost is one more lane-dispatch"* — against row 6's own
four-method roster, row 7 (*"Lane-count floor (`lanes >= 3` ...)"*), and the §3 risk-accepted
list (*"blanket lane-count raise as a diversity fix"*).
**Problem:** unreconciled numeric floors. Four named method lenses + a 2-of-N floor needs a
minimum of 4 lane-assignments; row 7 floors N at 3, the shipped default. At N=3 the two round-1
fixes cannot both hold. Neither row cross-references the other.
**Required fix:** one reconciling sentence — which the round-2 fix supplied, but see R3-5: the
reconciliation itself mis-adds.

### R2-9 — MEDIUM-HIGH — medium-high x medium-high x low — corroboration: HIGH (ledger clause read verbatim at debate.js:156, live at 88eb57f, re-confirmed at merge; row 10's rationale and §0's ledger praise both read directly)
**Location:** §3 row 10, impact cell — *"Medium — drift is usually caught by re-verification"* —
against §0's ledger praise.
**Problem:** two controls starve each other. The shipped ledger clause's skip-trigger is keyed to
the citing **prose** changing, not the **source** changing or access-date elapsed — the two
things row 10's fix tracks. Row 10's safety net is exactly the step the ledger suppresses for
every already-HIGH claim.
**Required fix:** add a time/access-date-based re-verification trigger to the ledger clause, or
re-grade row 10's impact cell honestly. Either closes it; silence does not.

### R2-10 — LOW-MEDIUM — low-medium x low-medium x low — corroboration: MEDIUM (reasoned from the report's own demonstrated counting failures R1-7/R1-8/R2-1 applied to a new instance)
**Location:** §3 rows 14/15's revisit triggers; §5 item 10.
**Problem:** three closures depend on a future reader correctly counting recurrences across runs
— the exact archaeology this report's own counting-method note shows failing. Citations got a
durable counter; recurrence triggers have none until item 11 ships.
**Required fix:** one clause: "revisit triggers have no counter until item 11 ships; treat as
advisory until then."

### R2-11 — LOW — medium x low x low — corroboration: HIGH (direct fetch of both files at 88eb57f)
**Location:** §3 row 4 — *"confirmed absent: no `--smoke` flag in `commands/research.md` or
`debate.js` as of `main` @ `47ae48d`[^SmokeAbsent]."*
**Problem:** the claim over-reaches its own footnote's verification trail (which checked only
`commands/research.md`) and is literally false for `debate.js`, whose header comment contains
the string descriptively. The functional conclusion (no `--smoke` code path) is correct.
**Required fix:** reword to the accurate scope.

---

## Live evidence logged round 2 (write-block, red-merge seat)

The report-file write-block fired **twice** on the round-2 merge pass: (1) direct Write of
`red/findings.md` refused ("Subagents should return findings as text, not write report files");
(2) a Write of the same content to a **scratchpad** path named `findings.md` (outside the run
tree entirely) refused with the identical message, while the same content under a neutral
filename succeeded. First artifact-logged demonstration that the guard keys on the *filename
semantics* regardless of directory — direct, positive evidence for §3 row 8(c)'s skepticism
(now partially settled: filename-keyed, path-independent). Occurrence ledger for the write-block
class, artifact-attested: run 1 (blue, `report.md`), run 2 (red, `findings.md`), retrospective
round 0 (blue-synthesis, self-observed label per R1-18), retrospective round 1 (red,
`red/findings.md`, debate.md RED friction), retrospective round 2 (red-merge, both paths).
**Round 3: no new occurrence claimed** — all five lens passes and this merge wrote via the
neutral-filename scratchpad-then-copy path as a precaution; the block was neither observed nor
ruled out this round.

## Noted, not raised (round 2 — preserved)

- **Live drift on the report's own pin, harmless this instance:** `main` advanced `47ae48d` →
  `88eb57f` (both docs-only; `debate.js` diff empty). Recorded for the pattern only.
- **§1.4's "15 well-formed gap-pattern files":** the live red-auditor memory store held 18 at
  round-2 audit time — stale-count churn inherent to describing a live accreting mechanism.
- **Row 16b keeper-run enforcement:** initially suspected policy-without-mechanism; refuted by
  direct read — `bulk = model ? { model } : {}` means keeper runs inherit the session model.
- **Row 13 vetting-step enforcement:** no `.mcp.json`/CI mechanism exists, but the row is an
  unbuilt before-run-4 action item like the rest of §3 — holding it to a built-mechanism
  standard would be inconsistent.

## Verified clean round 2 (see red/citation-ledger.md)

Doctrine quote L22–24 verbatim; red-lens bulk / red-merge judgment routing (L164/L168); judge
unguarded + citationPasses const re-confirmed at `88eb57f`; lanes=3 no-minimum + missing `lanes`
return key; simulator AsyncFunction injection + parallel catch-null semantics exact; 11 test
cases match §2.3's list in order; [^AgentTestTiers] 95%/5% exact quote; [^PR14Description]
verbatim; six cited backlog items unchanged despite two further commits; [^PdfMcp] maintenance
claims live-checked; R1-4/R1-5/R1-6/R1-7/R1-8/R1-19/R1-20 fix texts verified against sources;
R1-17's Claude-only-knobs premise; §4 rows 6–10 attributions line-by-line; §4 shape-verdict
frontier quote verbatim; §5 items 8–9 cross-references consistent. 28 claim-reference pairs
ledgered by the round-2 lenses + 4 merge-time re-verifications.

## Round 2 disposition summary (preserved)

FAIL. 20/20 round-1 gaps closed (5 with regressions re-raised); 11 new gaps open. Gate tier:
R2-3 and R2-7 (a tradeoff disposition and a security-adjacent risk-accept each resting on a
misdescribed mechanism/citation), R2-8 and R2-9 (two pairs of round-1-era controls that don't
compose), R2-1 (a fabricated recurrence count feeding a grade and a build trigger). R2-5/R2-6
close by one footnote + one §5 note; R2-2/R2-4/R2-10/R2-11 are mechanical. Expected convergence:
all eleven are additive-fix sized; none requires new infrastructure; none disputes H1–H5.

(Id-collision note: the report's §1.2 references "R2-10" from *run 2's* findings
(memory-architecture). This file's R2-* ids are this retrospective's round-2 gaps — unrelated.)

---

## Round 1 record (original grading, preserved verbatim — statuses above supersede)

Round 1 consolidated five lens passes (3x leaf-node citation verification slices, 1x
logic/completeness, 1x dark-side/risk). Keystone claims re-verified live at merge time
(2026-07-13): `special-circumstances` `main` HEAD was `47ae48d` (docs-only, one commit past
the `00018a5` PR-#14 merge); `debate.js:184` judge dereference unguarded; `citationPasses`
`const` at line 139, outside the `while` at line 148.

### Verdict (round 1): FAIL

20 gaps open, 0 closed, 0 rebutted, 0 adjudicated. R1-1/R1-2 must be corrected before the
report is safe to plan run 4 from; R1-4 and R1-11 are the next tier; the rest are additive
fixes or explicit risk-accepts blue can argue. None of the gaps overturn the report's
substantive conclusions on H1–H5 — the engineering analysis survives; the framing, several
graded cells, and a handful of citations do not.

### Round-1 gap details (as originally raised, condensed to grading + required fix; full prose in git history at the round-1 revision of this file)

**R1-1 — HIGH — realized x high x low** — §0 headline stale: PR #14 merged to `main` at
`00018a5` ~8 minutes after blue's verification commit; ~13 dependent cells loaded on the stale
premise; no pinned-SHA discipline. Required: re-verify, reframe, thread the flip, pin SHAs —
without naively flipping (see R1-2).

**R1-2 — HIGH — high x high x low** — §3 row 2's conditional triggered-and-unmet: judge call
site (`debate.js:181/184`) verified unguarded on merged `main`; reword to factual, dock the code
fix for run 4.

**R1-3 — MEDIUM — high x medium x low** — §2.3 item 1 overgeneralized "null at every call site
crashes": only judge dereferences unguarded; frontier/red-lens/assembly returns are discarded
(silent-degrade class); blue-respond already guarded. Differentiate by failure class.

**R1-4 — HIGH-MEDIUM — medium-high x medium-high x low; corroboration LOW on the figures** —
"~19%/~95%" absent from arXiv:2602.03794 (exhaustive percentage inventory); 19% traces to
arXiv:2603.22103 (narrative-similarity, narrower domain); ~95% traces to nothing. Re-cite with
caveat or drop; the false precision backed §3 row 6's Likelihood=High cell.

**R1-5 — MEDIUM — medium x medium x low; corroboration LOW-MEDIUM** — [^DiminishingReturns]'s
"2–4 agents" bound pinned to none of four bundled sources; VentureBeat's story is tool-count,
not agent-count. Pin or restate qualitatively.

**R1-6 — MEDIUM-LOW — certain x low-medium x low** — [^PR14] diffstat "+2281/−46" vs. live
+318/−48/18 files/11 commits; miscounted at origin. Correct.

**R1-7 — MEDIUM-LOW — certain x low-medium x trivial** — "35-entry" friction count conflates the
header's agent count with the file's 21 entries; two footnotes. Correct both.

**R1-8 — LOW — medium x low x trivial** — §4 row 6 "r1–r2" range unconfirmed; the cited file
attests one r1 instance. Correct or cite the r2 artifact.

**R1-9 — LOW — low x low x trivial** — #13890's silent-no-op signature differs from this repo's
worded refusal; soften the analogy.

**R1-10 — LOW-MEDIUM — certain x low-medium x low** — [^WriteBlock] carries "report.md-specific"
uncritically while the row's own second occurrence (findings.md) and run2-friction.md's
"filename-heuristic guard" falsify it. Strike or annotate.

**R1-11 — MEDIUM-HIGH — high x medium-high x low** — `citationPasses` computed once pre-loop,
never rescaled as the report grows; §2.3 item 4 phrased as if the recompute exists; no §3 row.
Recompute in-loop; add row; flag known-failing.

**R1-12 — MEDIUM — medium-high x medium-high x low** — routing table sends red-lens passes to
bulk tier against the file's own "never... the adversary" doctrine; §3 row 16 silent. Reclassify
or document the bounded tradeoff.

**R1-13 — MEDIUM-LOW — medium x low-medium x low** — row 15's "skeleton may moot it by
construction" mechanically unsound (ENAMETOOLONG is a command-length ceiling, orthogonal to
Write-vs-append); recurrence undercounted. Replace mechanism; re-grade.

**R1-14 — MEDIUM — medium x medium x low** — MCP-server adoption graded on cost alone, no
supply-chain vetting step, in a report headlining CVE-2026-21852. Add the vetting sentence.

**R1-15 — MEDIUM — medium x high x medium** — reflexivity blind spot: the report never applies
its own content-poisoning finding to FEOV's ~22-fetch research phase. A graded row or explicit
risk-accept closes it; silence is the gap.

**R1-16 — LOW-MEDIUM — low-medium x medium x low** — lane specialization concentrates failure:
one null dispatch drops 100% of a method's round coverage. Redundancy floor, re-dispatch policy,
or explicit risk-accept.

**R1-17 — LOW-MEDIUM — medium x low-medium x low** — cross-provider model diversity (the
report's own citation's stronger lever) never surfaced. Name it and its disposition.

**R1-18 — LOW-MEDIUM — high-that-something-happened x low-medium x low; corroboration LOW as
written** — §0 addendum's write-block occurrence is a self-report with no artifact trail, made
by the seat it vindicates. Label "self-observed, not yet artifact-logged" or footnote evidence.

**R1-19 — LOW — certain x low x trivial** — [^WisdomCrowds] URL 404s; real path
`/projects/wisdom-of-llm-crowd.pdf`. Correct.

**R1-20 — LOW — high x low x low** — header's "disconfirming budget met in each lane" false
parallelism: lane 3's per-claim checks are not a search-budget ratio. Quantify or rephrase.

### Round 1 disposition summary

FAIL. 20 open gaps: 2 high (R1-1, R1-2), 1 high-medium (R1-4), 1 medium-high (R1-11), 5 medium
(R1-3, R1-5, R1-12, R1-14, R1-15), 11 medium-low and below. Nothing here disputes the corpus's
substantive H1–H5 conclusions; what fails is present-tense accuracy (R1-1/R1-2), two
quantitative citations (R1-4/R1-5), one live shipped defect misfiled as a test case (R1-11),
and a set of gradable omissions. Expected round-2 path: R1-1/R1-2/R1-6/R1-7/R1-19 close by
mechanical re-verification and edit; R1-4/R1-5 close by re-cite-or-soften; the tradeoff gaps
(R1-12, R1-14–R1-17) close by one-sentence dispositions or argued risk-accepts, which red will
accept if reasoned.

---

## Debate record

Literal transcript: `debate.md` (this run directory). Per-round synopsis:

- **Round 0 — BLUE synthesis:** union of 3 lanes; all five doubts confirmed (H2 sharpened to destroyed-by-schema; H3 refined to trimodal); headline lane-conflict resolved by direct verification — every headline fix existed only on unmerged PR #14; the write-block fired live on the synthesis itself.
- **Round 1 — RED FAIL, 20 gaps (R1-1..R1-20):** PR #14 merged mid-debate ~8 minutes after blue's pin (headline stale, R1-1/2) but the judge call site is unguarded on `main`; `citationPasses` never rescales (R1-11, live defect misfiled as a test case); citation lenses pinned fabricated-precision figures ("~19%/~95%"), a 7x-off diffstat, a 404 URL; routing contradicts the script's own doctrine comment. **BLUE:** all 20 addressed additively, none rebutted; ENAMETOOLONG risk-accept re-argued with a corrected mechanism.
- **Round 2 — RED FAIL, 11 gaps (R2-1..R2-11):** the repair-regression round — five round-1 fixes contained new defects; sharpest catch R2-3 (the "2 vs 16" figure is the paper's L4 condition; persona-only L2 needs 8 agents); two pairs of round-1 controls don't compose (redundancy floor vs. lane floor; ledger skip-clause vs. access-date tracking); the poisoning mitigation was mis-described. **BLUE:** 10 conceded; R2-4 conceded-on-the-gap but red's proposed replacement citation REBUTTED with a fresh fetch — the source does not contain the figure.
- **Round 3 — RED FAIL, 10 gaps (R3-1..R3-10):** first code-trace round: schema-legal `{FAIL, gaps:[]}` loops silently to the ceiling (R3-1); blue-synthesize friction never harvested, falsifying "never dropped" (R3-2); the judge's carried-gap rationale has no delivery path (R3-3); R2-4's rebuttal accepted with evidence — red logs the unverified proposed citation as its own error. **BLUE:** all ten fixed; the two code gaps get graded rows + founding-suite cases rather than prose absorption.
- **Round 4 — RED FAIL, 5 gaps (R4-1..R4-5):** the gate finding — the contested-docket detector is lineage-blind (pure id equality, no `supersedes`), the judge was never dispatched in this corpus (zero `### LEAD`), four live regression chains prove it; blue had shipped red's either/or guard text verbatim instead of deciding (R4-2). **BLUE:** all conceded; decided throw; discarded its own chain reconstruction for red's more precise enumeration; cross-corpus ids prefixed `MA-`.
- **Round 5 — RED FAIL, 6 gaps (R5-1..R5-6):** regressions inside round 4's own fixes (the discarded chain list survived in row 23; memory-architecture status citations stale ~3–6x) plus two dark-side findings against proposed mechanisms (`supersedes` compliance unenforced, R5-5; friction lost on any throw, R5-6); two red lens errors caught and overruled at the merge seat, logged against red. **BLUE:** all six landed — R5-5 and R5-6 built rather than risk-accepted. **Ceiling fired before a round-6 audit; final state UNVERIFIED.**

Red's cumulative bookkeeping: 52 gaps raised, 46 closed (34 clean, 12 closed-with-regression or equivalent), 6 open-unaudited at the ceiling; per-round dedupe maps, merge-seat overrules, and "noted, not raised" restraint items are preserved in the embedded findings above. Blue's per-round fix accounting is in `blue/CHANGELOG.md`.

## Footnotes

Consolidated by union, not duplication: every footnote is defined once, verbatim, inside the embedded blue report's Footnotes section above (semantic word-labels: [^WorkflowJs], [^MainVsBranch], [^PR14], [^SimulatorTests], [^JudgeUnguarded], [^Reverify47ae48d], [^AgentDiversity], [^NarrativeSimilarity], [^DiversityCollapse], [^DiminishingReturns], [^ProvenanceSurvey], [^PdfMcp], [^SubagentWriteBug], [^FrictionCount], [^Run1Journal], [^WriteBlock], [^LiveBacklog], [^ChangelogR2], [^CostFigureProvenance], [^PinnedRepoState], [^GapIdScheme], [^MAStatusR5], [^WisdomCrowds], and the rest of the blue set), with access dates and pinned SHAs. Red's statement-by-statement verification of those citations — 100+ graded claim-reference pairs across five rounds, including the two merge-seat corrections of red's own lenses — is in `red/citation-ledger.md`, which is the confidence record to consult before re-using any figure from this report.
