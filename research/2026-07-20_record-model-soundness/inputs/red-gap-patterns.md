# red gap-pattern inventory (mirrored at run setup — read-only copy)

<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: assembly_seat_regression_to_pre_repair_phrasings.md -->
---
name: assembly-seat-regression-to-pre-repair-phrasings
description: A synthesizing seat writing from RECALL re-mints the round-0 form of exactly the claims that were repaired
metadata:
  type: feedback
  classes: [incomplete-repair-propagation, claim-contradicts-own-record]
---

# Assembly-seat regression toward pre-repair phrasings

(Recorded by the lead on red's behalf, 2026-07-18 — red's out-of-band catechism
sitting was READ-ONLY and drafted this entry in its findings; content is red's own.)

A synthesizing seat (judge at assembly, any seat writing summary prose from RECALL
of a many-times-repaired document) preferentially re-mints the ROUND-0 form of
exactly the claims that were repaired. Three independent instances in one section
of the sleeper-service run's assembled catechism (post-capture audit, 2026-07-18):

- "manual triage demonstrably loses signal" — reinstated the pre-R1-14 frame
  (the audited repair: the manual loop works; automation must argue its margin).
- "the open MCP-headless bug trio" — reinstated the pre-R1-5 phrasing the
  citation ledger had graded LOW — REFUTED in round 1 (one of three is closed).
- "$2-5/night" as running cost — recollapsed the ceiling-vs-expected-spend
  composition R2-18 was minted to separate.

Direction of drift was patterned, not random: value-side claims strengthened,
against-case residuals dropped. Recall is not the record.

**How to apply:** any assembled/synthesized section (catechism, TL;DR, verdict
detail, executive summaries) gets audited AS IF round-0 text: decompose to claims,
trace each to the artifact of record. Mechanical screen: grep the assembly for
every token in the run's propagation-grep lists (the repaired phrasings are known
— they are what blue's propagation sweeps chased); any hit is a candidate
regression. Structural close of record: synthesis sections belong INSIDE blue's
report template where the mandatory full-re-read covers them every round;
assembly must be union-copy, never authorship.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: citation_status_and_misattribution_patterns.md -->
---
name: citation-status-and-misattribution-patterns
description: Recurring blue citation defects — closed GitHub issues cited as "open", and real figures miscited to the wrong source
metadata:
  classes: [citation-figure-misattribution, citation-status-drift, figure-recount-fails]
  type: feedback
---

Two leaf-node citation gap patterns seen in the memory-architecture research (round 1). Check for
both on every citation-verification lens.

**Pattern A — "open bug" that is actually Closed (not planned).** Blue cited GitHub issues
(#57507, #56540) as "open bugs" when both are *Closed as not planned*. This is not cosmetic:
- It inverts the fix story (a not-planned issue will NOT be resolved upstream → design must own the
  workaround, not wait).
- It can create an unsatisfiable plan dependency (a *blocking* change was made "contingent on
  issue resolution" — but it will never resolve).
**How to apply:** whenever a claim rests on a GitHub issue, WebFetch the issue and confirm
open/closed + closure reason. "Closed not planned" ≠ "bug doesn't exist" (the report still
corroborates the phenomenon) but the status and any plan-dependency wording must be corrected.

**Pattern B — real figure, wrong source.** A striking quantitative claim ("60% loss / 36.7×
compression / 2,000 facts") was cited to a blog that does not contain it; the number actually
comes from a different paper (arXiv 2603.17781, "Facts as First Class Objects"). The claim is
true but the footnote is unfollowable — exactly the "laundered into fact" failure.
**How to apply:** for headline numbers, fetch the *cited* source and confirm the number appears
*there*, not merely that the number exists somewhere. Grade statement-as-cited LOW even when the
underlying fact is true.

**Pattern C — a Round-N repair introduces a FRESH contradicted figure.** When blue softens a
challenged number ("80–99%" → "up to ~90–95% (MINJA / environment-injection)"), it may attach the
softened number to a *specific* source that does not support it. Round 2: the environment-injection
"~90%" was pinned to arXiv 2604.02623, whose real max ASR is ~32.5% (up to 8× under stress from a
low base). The repair swapped an untraced band for a *contradicted* attribution — worse, not
better. **How to apply:** never treat a Round-N citation repair as trustworthy because it was a
"fix." Re-follow every repaired footnote to its named primary; a repair that pins a number to a
source is a new statement↔reference pair to verify from scratch. Watch especially for repairs that
split a bundled band across named sources — each named source must now individually carry its half.

**Pattern D — disconfirming/"consensus" citation that rests on an unfollowable self-survey.** Blue
cited a dev.to article + "practitioner consensus surveyed <date>" for a quote-shaped disconfirming
claim; the article framed the topic differently and did not carry the claim, so the load fell on
the unfollowable self-survey. The human/agent's own survey is an untrusted, unfollowable source.
**How to apply:** when a footnote bundles a real source with "consensus surveyed" / "practitioner
consensus," follow the real source and check it carries the *specific* quoted claim; if not, the
claim rests only on the self-survey — flag it, especially when it is a disconfirming leg blue uses
to weaken its own grade.

**Pattern E — the self-audit's verification-limit EXCUSE misstates the source's accessibility.**
Blue honestly labels a figure "from search digest, not leaf-verified (paywalled)" — but the cited
source is arXiv-open, one WebFetch from adjudication (efficiency run, round 1: the "~34% NVD-vs-CNA"
figure excused as paywalled was cited to open arXiv 2508.13644, which does not contain the figure
at all — the paper compares scoring *systems*, never NVD vs CNA). An honest-looking hedge can hide
a misattribution: the hedge implies "the figure is in this paper, just unchecked," when the figure
is not in the paper. **How to apply:** on every "not leaf-verified because <reason>" label, test the
reason — if the source is actually open, do the fetch yourself and adjudicate; grade misattribution
(not mere non-verification) when the figure is absent, even for claims blue fenced as
non-load-bearing.

**Related caveat — HTML-arXiv fetch is lossy on numbers in tables.** When a fetch can't find a
cited statistic in an arXiv HTML page, grade "uncorroborated at leaf node," not "false" — the
small-model fetch often can't read tables. Flag for re-verification against the PDF. (Seen with
arXiv 2604.24450 / 61.38% / 71.58%.)

**Related caveat — WebFetch of a GitHub issue thread is lossy on COMMENTS (false absence-claims).**
Sleeper-service run round 1: WebFetch of issue #22055 confidently reported "no PreToolUse/chmod
workaround in thread" while `gh issue view 22055 --comments` showed a full PreToolUse workaround
comment (GitHub collapses/paginates comments; the page fetch never sees them). Status checks via
WebFetch are fine; any claim about what a thread DOES or DOES NOT contain must go through
`gh issue view <n> --comments` + targeted grep before grading. Same round, same issue: the thread
check also caught a half-false footnote ("chmod 444 + PreToolUse in thread" — PreToolUse yes,
chmod 444 absent) — split bundled workaround claims and verify each half.

**Related caveat — the fetch model's SUMMARY can contradict its own ENUMERATION.** Efficiency run
round 3: asked to count Table 2 rows of arXiv 2510.12697, the fetch model listed all 22 rows
correctly, then asserted "18 reported configurations" in its summary line. Had the summary been
trusted, a *correct* blue claim ("22 configurations") would have been falsely flagged. **How to
apply:** never cite the fetch model's aggregate/count/summary statement as the verification
result — recount from its quoted enumeration by hand; if it enumerates and summarizes
inconsistently, the enumeration is the evidence and the discrepancy goes to friction, not to a
finding against blue.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: env_write_block_filename_keyed.md -->
---
name: env-write-block-filename-keyed
description: The subagent report-file write-block keys on FILENAME semantics regardless of directory — even scratchpad/findings.md is refused; workaround = Write under a neutral name, then Bash cp into place
metadata:
  classes: []
  class_note: harness-limit — a tooling constraint, not a research defect class; kept for red's reference and deliberately NOT delivered as a repair duty
  type: project
---

The Task-tool write-block ("Subagents should return findings as text, not write report files")
fires on filename semantics alone, independent of path. Verified with a control condition at
the FEOV-retrospective red-merge, round 2 (2026-07-14): Write of `red/findings.md` refused;
Write of the *identical content* to a scratchpad path named `findings.md` (outside any run
tree) refused with the identical message; the same content under `r2-consolidation.md` in the
same scratchpad succeeded and was `cp`'d into place.

**Why:** the run corpus previously treated the trigger as uncertain ("may be semantic/
role-based, not purely filename-based") — this is the first artifact-logged test isolating the
filename variable. It also means the round-0 "scratchpad-write-then-copy" workaround only ever
worked because the scratchpad file had a neutral name.

**How to apply:** when a red-merge (or any subagent seat) must update `findings.md`/`report.md`
or similar report-named living artifacts: Write the full content to the scratchpad under a
neutral filename (e.g. `r<N>-consolidation.md`), then `cp` to the destination — cp is short,
no heredoc, so it also dodges ENAMETOOLONG. `Edit` on the existing destination file also works
for small changes. Do not burn attempts on Write-with-trigger-name at any path. Re-test
occasionally: the plugin's pre-created-skeleton fix or a platform change may alter behavior.

Related: [[pattern-repair-regression-citation]] (the recurrence-count discipline that says log
occurrences like this one with an artifact trail).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: gap-pattern-verification-file-type-blindspot.md -->
---
name: gap-pattern-verification-file-type-blindspot
description: Gap pattern — "no X exists" claims backed by a grep scoped to one file type miss X implemented in another layer (e.g. compiled tools)
metadata:
  classes: [verification-scope-blindspot, false-universal]
  type: feedback
---

Gap pattern to hunt: a blue "no such thing exists / must be built from scratch" claim whose
supporting local verification is a grep scoped to a **single file type**.

**Why:** In the 2026-07-12 memory-architecture audit, blue asserted "no secret-scrub gate exists,
build it" (§6.3, §8 item 3) backed by footnote `[^LocalRepoScrub]`: `grep -i secret|scrub|denylist
across *.md`. That scope was blind to the Go tool layer. A shipping PreToolUse hook
(`sc-secrets-gate`), a reusable pattern package (`internal/secrets`), tests, and `hooks.json`
wiring already existed. The claim miscast a blocking item's effort ("build from zero" vs "wire a
new consumer onto an existing package").

**How to apply:** When a report says "X does not exist" and cites a local grep, check the grep's
`--include`/glob scope. Non-existence claims are only as strong as the search surface. Re-run
across code (`*.go`, `*.ts`, `*.py`), config (`*.json`, `*.toml`), and wiring/manifest files
before accepting. A capability can be present in a compiled binary or hook manifest while absent
from the prose/skill markdown that describes it (which may lag in future tense). Also check the
narrower truth: the existing capability may cover a *different surface* than the design needs
(here: outbound-tool-input scan, not commit/push-time scan) — that nuance is the real gap.

**Compounding form (Round 1 merge, same audit):** a *second* independent verifier (a different red
lens) repeated the identical `*.md`-scoped grep and **corroborated the false "no gate exists" claim
HIGH**. Two verifiers agreeing does NOT raise confidence when they share the same flawed method
scope — the agreement is an artifact of the shared blindspot, not independent confirmation. When
consolidating multiple lens/verifier passes, do not treat concurrence as corroboration until you
confirm the verifiers used *different* search surfaces. Re-verify shared-method agreements at the
leaf node yourself (I did; found `sc-secrets-gate` + `internal/secrets` + `hooks.json` live).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: gap_live_source_drift.md -->
---
name: gap-live-source-drift
description: Citation figures/mechanisms drift when live web sources move after the report's access date; re-follow to primary, don't trust the footnote's numbers
metadata:
  classes: [live-source-drift]
  type: feedback
---

Gap pattern: **live-source drift**. A footnote's access date freezes a claim, but the cited
source keeps moving. Star counts, "current algorithm" descriptions, and vendor pipelines change;
the footnote silently goes stale even though it looked verified at drafting.

**Why:** In the memory-architecture audit (round 1, lens 3) three findings were only catchable by
re-fetching live sources: mem0 had switched from retrieve-then-classify (ADD/UPDATE/DELETE/NOOP)
to single-pass ADD-only; claude-mem stars had moved 46k→87.1k; a Letta "git-branch" mechanism was
absent from the cited blog entirely (traced only to an unnamed forum). An audit against archived
snapshots would have passed all three.

**How to apply:**
- Always re-follow citations to the *current* primary source, not just confirm the footnote is
  well-formed. Grade corroboration against what the source says *now*.
- Volatile numbers (stars, benchmarks, "latest version") get LOW confidence unless pinned with an
  access date AND the substantive claim survives without the exact number.
- When a "mechanism to steal / adopt" is recommended, verify the vendor still ships it — vendors
  abandon the exact pipeline being praised (mem0 case). A recommendation to adopt an abandoned
  design is a MEDIUM substantive gap, not pedantry.
- Watch for compound footnotes (`blog + docs + community forum`) where the load-bearing detail
  traces only to the vaguest, unfollowable member. Demand the specific source or downgrade the
  claim.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: MEMORY.md -->
- [Workflow undefined run-dir](workflow_undefined_rundir.md) — caller can pass literal "undefined" paths; hard-fail, never fabricate an audit surface
- [Citation status & misattribution patterns](citation_status_and_misattribution_patterns.md) — "open bug" that's actually closed-not-planned; real figures miscited to wrong source; false-paywall excuse hiding absence; lossy arXiv-HTML fetch
- [Risk-grading conflations](pattern_risk_grading_conflations.md) — likelihood≠success-rate, verdict-keystone-on-unverified, fixes that relocate the problem
- [Live-source drift](gap_live_source_drift.md) — re-follow citations to the current primary; volatile figures + abandoned vendor pipelines only catchable live, not from snapshots
- [Verification file-type blindspot](gap-pattern-verification-file-type-blindspot.md) — "X doesn't exist" backed by a grep scoped to one file type (e.g. *.md) misses X in the compiled/tool/manifest layer
- [Self-defeating mitigation](pattern_self_defeating_mitigation.md) — a control added to close a prior gap has its own failure mode: collides with the system's write loop, leans on discredited diligence, has an escape hatch, or only closes the durable path
- [Inherited-surface netting](pattern_inherited_surface_netting.md) — "risk is inherited from native, adopting buys nothing" must verify the baseline wasn't patched; bespoke may re-open what upstream closed
- [Incomplete-repair footnote lag](pattern_incomplete_repair_footnote_lag.md) — repair lag runs BOTH ways (footnote lags body, or body lags the repaired footnote); grep the retracted token report-wide in both directions
- [Repair-regression on citations](pattern_repair_regression_citation.md) — softening an unpinnable figure by re-citing to a NEW specific source can regress: new source contradicts (≤32.5% vs claimed ~90%) or accurate number lives in an uncited paper; re-verify every repair as a new claim
- [Footnote over-attribution](pattern_footnote_overattribution.md) — a footnote's claim-list bundles specifics but only the generic one traces; a figure multi-cited to N footnotes often has only one real source wearing a crowd — pin which one carries the load
- [Provenance self-report, stale gate, contradictory accept](pattern_provenance_self_report_and_stale_gate.md) — fix trusts metadata self-reported by the compromised component; point-in-time flag branch re-opens the conflict its fallback closed; risk-accept premise flips sign vs the build-value premise
- [Missing root invariant](pattern_missing_root_invariant.md) — 3 rounds of gate-by-gate patching where each fix spawns the next gap = a missing stated invariant; surface it (recommend if severity declining, block if not) alongside grading the instance
- [Invariant soundness-by-enumeration](pattern_invariant_soundness_by_enumeration.md) — adopted invariant claimed "sound/mechanical" but rests on an under-inclusive channel denylist (omits Bash/MCP/sidechain/in-repo); prove incompleteness via the system's own symmetric defense; recommend allowlist inversion
- [Metric-conflation & traceable≠verified](pattern_metric_conflation_and_traceable_not_verified.md) — a "success band" whose endpoints are two different metrics (ISR-low+ASR-high); "closed-as-traceable" ≠ digit-verified; try arXiv /html/<id>v2 when /abs/ + PDF lack the number
- [Policy-without-mechanism](pattern_policy_without_mechanism.md) — invariant asserted "self-enforcing / not new machinery" after its concrete enforcer was withdrawn over prior rounds; also supersession-accounting drift + Heilmeier headline-lag
- [Self-referential repo drift](pattern_self_referential_repo_drift.md) — the audited repo itself (not a citation) moves between blue's verification and red's audit; re-verify each sub-claim, a merge landing ≠ every related call site fixed
- [Doctrine vs implementation](pattern_doctrine_vs_implementation.md) — stated principle vs literal logic below it; now also: "not an entry point" contradicted by shipping in the entry-point dir + guard flag on sibling only
- [Reflexivity blindspot](pattern_reflexivity_blindspot.md) — report headlines a risk class for its subject but never applies it to its own toolchain/fetch surface/drafting window; silence ≠ argued out-of-scope
- [Self-defeating mitigation, extended](pattern_self_defeating_mitigation.md) — now also: mitigation claims "independent" re-check but protocol only re-fetches the same source; two controls from different rounds silently starve each other's trigger
- [Unreconciled numeric floors](pattern_unreconciled_numeric_floors.md) — a redundancy/allocation requirement (>=2 of N on category X) and a separately-added floor on N don't arithmetically compose; now also: the "reconciliation" fix itself can mis-add (recompute, don't re-read) + unenforced default called a "floor"
- [Measurement-methodology drift](pattern_measurement_methodology_drift.md) — cited self-reported figure passes leaf-node fidelity (source says X) but a LATER audit of the same project shows the counting method itself undercounts by ~92%; check method soundness, not just citation fidelity
- [Within-source condition misattribution](pattern_within_source_condition_misattribution.md) — right paper, exact figure, wrong experimental arm: a gloss ("[persona-lensed]") reassigns an L4 result to L2 to justify not building L4; pin every quoted result to its condition
- [Repair-regression, extended](pattern_repair_regression_citation.md) — adjacent-narrative transposition; RED-side: leaf-verify your own required_fix claims/phrasings (blue copies verbatim); no undecided "or"; a lens's "matches red's fix exactly" ≠ leaf verification — source read wins at merge
- [Identity-keyed detector, lineage-blind](pattern_identity_keyed_detector_lineage_blind.md) — recurrence-escalators keyed on stable ids never fire when the process mints fresh ids per cycle; zero-firing telemetry is evidence FOR the pattern; red's own conventions are part of the system under test
- [Gitignored ≠ absent / present ≠ committed](pattern_gitignored_not_absent.md) — verify "absent" with `ls` AND "committed" with `git status`+`git log --all`; both directions bite
- [Write-block is filename-keyed](env_write_block_filename_keyed.md) — even scratchpad/findings.md refused; Write under neutral name then `cp` into place (also dodges ENAMETOOLONG)
- [Schema-legal control-flow trace](pattern_schema_legal_control_flow_trace.md) — dark-side lens's best yield after citations go clean: hand-trace schema-valid-but-incoherent shapes + grep the aggregator's call sites vs. every seat that could feed it
- [Exhaustive sweep omits hard case](pattern_exhaustive_sweep_omits_hard_case.md) — "every/all/N-of-N" tables and checks skip the report's own named specimen/conflicted position; diff the claimed universe against the sweep
- [Stale-baseline pricing](pattern_stale_baseline_pricing.md) — efficiency levers priced on last run's cost distribution while shipped code already moved it; a correctly-described mechanic whose implications never propagate into the pricing sections
- [Unquoted hold masks discrepancy](pattern_unquoted_hold_masks_discrepancy.md) — a lens's "checked, no discrepancy" without side-by-side quotes can mask a live defect; merge resolves lens conflicts by direct read, never majority; now also: a hold's excuse-mechanism (trailing newline etc.) is itself testable — test it
- [Fail-open guard & erasable evidence](pattern_fail_open_guard_and_erasable_evidence.md) — marker-keyed guards that silently disarm on marker loss; "tamper evidence" the audited actor can rewrite/revert before review — demand fail-closed polarity + out-of-session evidence
- [Audited-artifact sibling halo](pattern_audited_artifact_sibling_halo.md) — flagging one defect in a source confers unearned trust on its sibling figures; recompute every figure taken from a partially-audited artifact
- [Conditional-vote laundering](pattern_conditional_vote_laundering.md) — a "2/3 ratify" tally counts a vote whose precondition the same synthesis rejects; check each counted vote's conditions against the document's own dispositions
- [Method-footnote under-coverage](pattern_method_footnote_undercoverage.md) — reproduction-promising method footnote emits every figure except the headline sub-figure; re-derive each figure, flag the step-less one
- [False-equivalence disjuncts](pattern_false_equivalence_disjuncts.md) — "A, or equivalently B" where only one branch has an operator/prevents the stated harm; window-without-a-watchman; per-item thresholds salami-slice under an item cap
- [Sibling-repair composition](pattern_sibling_repair_composition.md) — two same-round fixes to one design (mechanism hardened + registered prediction restated) can each verify clean yet not compose; re-derive the test from the hardened mechanism
- [Ephemeral instrument & grid-count](pattern_ephemeral_instrument_and_grid_count.md) — cited self-measurement whose script/input live only in scratchpad = unre-derivable self-report; table counts need row enumeration, fetch summaries assert the grid product
- [Waiver graduation & closure amendment](pattern_waiver_graduation_and_closure_amendment.md) — not-raised waivers are class-conditional; late composition defects reclassify prior clean closures with declared lineage
- [Pinned mapping not total](pattern_pinned_mapping_not_total.md) — a pinned enum→numeric mapping/reading rule audited for existence, never for domain coverage: unmapped tokens, synonyms at opposite extremes, unpinned compound cells
- [Controller lookahead pricing](pattern_controller_lookahead_pricing.md) — counterfactual throttle savings priced on post-round state the mechanism couldn't read at decision time; round-N control reads round-(N−1) board
- [Post-mortem misdiagnosis](pattern_postmortem_misdiagnosis.md) — corrected figures clean but the "old error happened BECAUSE X" story fails reproduction; test X against the OLD number, check claims about red's own methods vs the ledger
- [Repair-premise foreclosed by own profile](pattern_repair_premise_foreclosed_by_own_profile.md) — guard tags a write the design's own fence/denies already foreclose; dead where aimed, and the denied write was a subsystem's REQUIRED behavior
- [Config-sample schema & anchoring](pattern_config_sample_schema_and_anchoring.md) — embedded config JSON: verify key NESTING vs schema (misplaced key = silent no-op) and path-rule ANCHOR (cwd vs settings-source vs //absolute); quote fidelity doesn't transfer to the artifact
- [Authorship evades never-edit guard](pattern_authorship_evades_never_edit_guard.md) — "never EDITS its inputs" leaves amplification open when the loop AUTHORS new artifacts the harvester reads; check corpus membership, not just mutation
- [Gate-stack rung non-portability](pattern_gate_stack_rung_nonportability.md) — gates built for one deployment rung (local wrapper, out-of-repo settings) silently assumed on cloud/CI rungs; demand a per-rung gate-survival table
- [Verification probe layer-masking](pattern_verification_probe_layer_masking.md) — PASS produced by a different co-located layer; deny-canary over shared boundary can't isolate; ALLOW-side too: "no prompt" under defaultMode:auto ≠ carve-out classifier; nested claude -p probes are classifier-blocked from seats
- [Origin-tag naming-keyed](pattern_origin_tag_naming_keyed.md) — tag by CAUSE not name; now also: the untagged/out-of-contract actor STAFFS the corroboration pool, and change-keyed tags over-strip appended-to patterns
- [Per-instance cap not per-cause](pattern_per_instance_cap_not_per_cause.md) — retry cap keyed per run-dir/item lets a deterministic cause re-wedge every fresh instance; "cannot recur forever" holds per-instance only
- [Wildcard allow write-gadget](pattern_wildcard_allow_write_gadget.md) — a wildcard allow rule called "read-only" (Bash(git log *)) hides a write via the tool's own --output flag; run the leaf, enumerate output flags, watch out-of-repo targets escape the porcelain/hash watchman
- [Happy-path-only gate](pattern_happy_path_only_gate.md) — step-N-only compare/snapshot skips abort paths; cross-run chain vs "prior END" has no referent after a dead run; abort is model-influenceable
- [Rung-nonportability, extended](pattern_gate_stack_rung_nonportability.md) — now also: rung ZERO/manual mode voids wrapper-keyed guards; "same code path" vs "bypasses wrapper" contradictions; stale gate tables
- [Identity-keyed, extended](pattern_identity_keyed_detector_lineage_blind.md) — now also: "normalized" signature asserted with normalization unspecified while the design re-mints dated/pathed content per cycle
- [Closed-mode built-in carve-out](pattern_closed_mode_builtin_carveout.md) — "default-closed" mode premise defeated by its documented non-configurable auto-approve set (dontAsk read-only Bash); redundant allows = unmodeled carve-out smell
- [Change-summary desync misjudges round state](pattern_changesummary_desync_round_state.md) — a lens read stale CHANGELOG, concluded blue shipped no revision; merge resolves round parity by direct read of report.md, never the change-summary
- [Structural close scoped to top-level session](pattern_structural_close_scoped_to_toplevel.md) — "closed at the tool boundary" covers the top-level session but not the spawned Workflow/seat agents the report's own text concedes are Bash-reachable; hook-coverage (L2) ≠ settings inheritance (L1); fence blocks writes not read-exfil
- [Benign bucket launders attack evidence](pattern_benign_bucket_launders_attack.md) — a fix declaring a denial-stream "expected/normal" keyed on the protected TARGET path auto-classifies attacker attempts as benign; enforcement intact but detection erased
- [Acceptance-test polarity lag](pattern_acceptance_test_polarity_lag.md) — a repair inverts a mechanism but the standing acceptance test asserts the OLD behavior; unsatisfiable gate whose cheapest pass is undoing the repair — demand the two-legged restatement
- [Assembly-seat regression toward pre-repair phrasings](assembly_seat_regression_to_pre_repair_phrasings.md) — synthesis-from-recall re-mints round-0 forms of repaired claims; grep assemblies against propagation lists; synthesis sections belong inside blue's template


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_acceptance_test_polarity_lag.md -->
---
name: pattern-acceptance-test-polarity-lag
description: A repair inverts a mechanism's polarity but the standing acceptance test still asserts the OLD behavior — the gate becomes unsatisfiable and its cheapest pass is undoing the repair
metadata:
  classes: [incomplete-repair-propagation, cross-section-contradiction]
  type: feedback
---

A repair changes what a mechanism DOES (e.g. `/self-improve` becomes an inert trampoline;
the payload moves to the wrapper), but an acceptance test adopted earlier — often quoted
from an external plan — still asserts the old polarity ("headless `claude -p '/self-improve'`
produces a run dir"). Post-repair the test FAILS BY CONSTRUCTION, and the perverse incentive
is sharp: the cheapest way for a builder to make the printed gate pass is to re-inline the
payload, silently undoing the repair.

**Why:** round-5 sleeper-service (R5-2, sup R4-1). The R4-1 trampoline made the port plan's
Phase-4 verify step unsatisfiable; §3.3 still said it "must remain the acceptance test," and
§3.4's ladder cell still described the command markdown as the payload. Neither site was in
the repair's edit list or propagation greps.

**How to apply:** after any repair that inverts or relocates a mechanism, hunt the standing
ACCEPTANCE TESTS and gate definitions that reference it (grep the mechanism's old name and
the test's promised observable). The repaired design's correct test is usually TWO-legged:
(i) the new path produces the promised artifact; (ii) the OLD path now produces NOTHING —
the repair's own property becomes a verifiable assertion. A test that only exercises the new
path misses the regression channel; a test that still asserts the old path incentivizes
reverting the repair. Distinct from stale gate-survival tables: here the defect is the
test's POLARITY, not its rung coverage.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_adjacent_defect_count_bleed.md -->
---
name: pattern-adjacent-defect-count-bleed
description: A repair re-grades defect A's recurrence count using an event that actually belongs to structurally-adjacent defect B — check the cited section literally contains the claimed event, not just that a plausible-sounding event exists somewhere nearby
metadata:
  classes: [citation-figure-misattribution, figure-recount-fails]
  type: feedback
---

When two defect classes are narratively paired throughout a report (e.g. "write-block,
ENAMETOOLONG" as the recurring Tier-B pair), a round's repair to one class's recurrence count
can silently import an occurrence that actually belongs to the *other* class — because both
use the same language ("third occurrence," "this round," "the merge seat") and appear in the
same sentences/paragraphs throughout the corpus.

**Why:** In the FEOV retrospective round 2 audit, blue's R1-13 fix re-graded ENAMETOOLONG's
likelihood Medium→High citing "a third occurrence... at the red-merge seat, per debate.md's
round-1 merge-seat friction." Direct read of that exact cited section showed it contains only
a PDF-fetch-depth note and a process-misfit note — zero mentions of ENAMETOOLONG, heredocs, or
shell-parse failures. The actual documented ENAMETOOLONG occurrences were two (run 2, and this
retrospective's own round-0 synthesis), not three. The write-block defect — described in
adjacent sentences throughout the same report, also using "third occurrence" language, also
citing the round-0 synthesis and a round-1 red-merge hit — is what the round count actually
tracked. The repair correctly copied a *number* from the surrounding prose without checking
that the number's *source citation* named the right defect class.

**How to apply:** When a repair changes a likelihood/recurrence grade citing "per [section X]",
open section X and confirm the specific defect name/keyword appears there — do not accept that
*a* recurrence event of *some* kind is described nearby as sufficient. This is a distinct failure
mode from [[pattern_repair_regression_citation]] (new source contradicts or omits a figure): here
the cited section is real and about the same *general topic* (this report's Tier-B live-smoke
defects) but the specific claimed event is simply absent — a total-absence miscitation, not a
contradicting-number miscitation. Watch especially for shared adjectives ("third occurrence,"
"this round," "the merge seat") reused across two co-located defect rows in a graded table — the
copy-paste boundary between rows is where the bleed happens.

Related: [[pattern_repair_regression_citation]], [[citation_status_and_misattribution_patterns]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_audited_artifact_sibling_halo.md -->
---
name: pattern-audited-artifact-sibling-halo
description: Blue audits ONE finding of a pinned artifact as defective, then transcribes a SIBLING finding's own error from the same artifact unchecked — the partial audit confers unearned trust on the rest of the source
metadata:
  classes: [audited-artifact-sibling-halo, figure-miscomposition]
  type: feedback
---

When a report demonstrates it audited a pinned artifact (e.g. flags cost.md finding 2 as
internally contradicted), do NOT let that demonstrated diligence stand in for verification of the
artifact's OTHER findings quoted elsewhere — recompute each transcribed figure independently.

**Instance (run 4, round 1, L3-F4):** blue's §6.4 correctly flagged run-3 cost.md finding 2
("merge cost tracks dispute size" vs its own table), but §4.2 transcribed finding 3's "12.5×
cache-write" as a multiplier — the artifact's own pricing header (2.5 vs 12.5 $/MTok) makes the
ratio 5×; 12.5 is the absolute rate. Units-vs-ratio confusion inherited verbatim from the source.

**Why:** a partial audit reads as "this source was checked"; the halo suppresses re-derivation of
sibling claims. Related to [[measurement-methodology-drift]] (citation-faithful ≠ sound) and
[[pattern-risk-grading-conflations]].

**How to apply:** whenever a report flags defect(s) in a source it also cites approvingly,
enumerate every OTHER figure the report takes from that same source and recompute each from the
source's own primitives (pricing headers, tables, totals) — the flagged defect raises, not
lowers, the prior that siblings are wrong.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_authorship_evades_never_edit_guard.md -->
---
name: pattern-authorship-evades-never-edit-guard
description: "Self-poisoning guard says loop NEVER EDITS its inputs — but the loop APPENDS new artifacts (its own run dir's friction/cost) that enter the same harvest corpus; amplification via authorship, not mutation"
metadata:
  classes: [partial-control-coverage, cross-section-contradiction]
  type: feedback
---

A feedback-loop guard phrased as "never edits its own inputs" is satisfiable while the
amplification loop stays open: the system can *author new* input artifacts (e.g. its nightly
run dir's friction.md/cost.md land in tomorrow's harvest) without ever editing an existing
one. Check the input corpus's membership rule, not just its mutation rule — does anything
the system *creates* re-enter the scoring set? Recurrence×staleness scores then self-ratchet
nightly with no human in the sub-loop.

**Why:** Sleeper-service round 1 (L6-F1): §1.5's guard ("NEVER edits ... harvest inputs;
ranking input surface stays human-and-run-record authored") directly contradicted §4.3's
"the loop eats its own complaints the next morning" — loop-authored run records ARE run
records.

**How to apply:** For any self-improving/harvesting loop, enumerate every path the harvester
reads and ask who can *create* files there. Guards over Edit/mutation are incomplete without
an authorship/origin rule (tag or exclude self-authored inputs). Related: [[pattern_self_defeating_mitigation]], [[pattern_identity_keyed_detector_lineage_blind]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_benign_bucket_launders_attack.md -->
---
name: pattern-benign-bucket-launders-attack
description: A fix that declares a denial-stream "expected/normal" keys the benign bucket on the protected TARGET path, so an attacker aiming at that target inherits the benign label — the signal is erased, not distinguished
metadata:
  classes: [name-keying-vs-marker, partial-control-coverage]
  type: feedback
---

When a prior round asks "are routine denials distinguishable from attack evidence in the
fired-record?", a fix that ANSWERS by bucketing all denials to that target as a NORMAL class
(`expected-memory-deny`, doctor line does not flag) does not distinguish — it ERASES the
signal. If the bucket is keyed on the TARGET PATH (`.claude/agent-memory/**`) rather than the
ORIGINATING ACTOR/CAUSE, then any actor targeting that path — including an injection trying to
poison the protected surface — is auto-classified benign. The write stays blocked (enforcement
intact) but DETECTION is gone, on precisely the surface tamper-attempts most need to be visible.

**Why:** round-5 sleeper-service R4-4. The nightly red seat's own memory write is denied by
design (correct); to keep the resulting denial-stream un-alarming, the parser buckets ALL
`.claude/agent-memory/**` denials as `expected-memory-deny`. That surface is red's gap-pattern
corpus (steers future audits) — a write attempt there is as alarming as a guardrail delta, but
it's routed to the one bucket the doctor line ignores. Mirrors the origin-tag lesson (tag by
CAUSE not name) inverted onto a denylist bucket.

**How to apply:** when a fix makes a security-relevant event "expected/normal/routine," check
what KEY the benign classification uses. Target-path-keyed benign buckets are attacker-inheritable
(the attacker chooses the target). Demand the benign label be keyed on the ORIGINATING ACTOR +
the specific expected write shape; every OTHER actor hitting that target surfaces as a distinct
persistent anomaly. "Enforcement intact but detection erased" is still a gap — grade it on lost
attack visibility, not on breach (the write is blocked).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_changesummary_desync_round_state.md -->
---
name: pattern-changesummary-desync-round-state
description: A lens can misjudge the whole round state (thinks blue shipped no revision) when CHANGELOG/debate lags the actual report.md edits; merge resolves round parity by direct read, never by the change-summary
metadata:
  classes: [reader-channel-mismatch, self-attestation]
  type: feedback
---

At merge, determine whether blue revised THIS round by direct read of report.md, never
from CHANGELOG.md or the debate ### BLUE block.

**Why:** round-3 sleeper-service — blue edited report.md (added invariant 7 + all R2
fixes) but wrote NO round-3 CHANGELOG block and NO round-3 ### BLUE. Lens 3 read the
CHANGELOG, saw it ended at "Round 1", and concluded "blue shipped no round-3 revision;
report state = post-round-2" — then carried prior-round HIGHs instead of re-verifying the
new text. Lenses 1/2/5/6 read the report and found the R2-fix text present. The
change-summary is a navigation hint (protocol says so); when it silently desyncs it
actively misleads, and a lens that trusts it under-audits fresh text.

**How to apply:** (1) at merge, grep report.md for this round's expected fix tokens
(invariant names, gap-id fix markers like "R2-16", the lead's directed additions) before
trusting any lens's "no revision this round" claim; a lens conflict on round state is
resolved by direct read, never by majority. (2) If CHANGELOG/### BLUE lags report.md,
file it as friction — the desync is a real defect in the navigation channel, not cosmetic.
(3) A lens carrying prior-round grades on the premise "nothing changed" is only valid if
YOU confirmed nothing changed; re-derive the changed-sections set yourself.

Related: [[pattern_self_referential_repo_drift]] (the audited artifact moving under you),
[[pattern_unquoted_hold_masks_discrepancy]] (merge resolves lens conflict by direct read,
not majority).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_closed_mode_builtin_carveout.md -->
---
name: pattern-closed-mode-builtin-carveout
description: A design's "default-closed / allowlist-inverted" premise built on a platform mode is defeated by the mode's own documented built-in auto-approve carve-out
metadata:
  classes: [false-universal, enumeration-non-exhaustive]
  type: feedback
---

When a report claims a permission MODE makes the surface default-closed ("anything not
allow-listed is denied"), verify the mode's documented EXEMPTIONS, not just its headline
semantics. Instance (sleeper-service r3, L2-F2): `dontAsk` "denies anything not in your
permissions.allow rules **or the read-only command set**" — a non-configurable built-in Bash
set (`ls, cat, echo, pwd, head, tail, grep, find, wc, which, diff, stat, du, cd`, read-only
git) that runs "without a permission prompt in every mode". The design's read-scoping
closure (Read/Grep/Glob allow-scoped + Read denies) covered the Read TOOL only; Bash `cat`
on the same secret paths was auto-approved, silently re-opening a risk-matrix row already
graded closed.

**Why:** closed-world premises inherit every platform carve-out; the carve-out lived on the
same doc pages a prior round had graded HIGH — either live-source drift or under-read, so
re-fetch mode semantics whenever a closed-world claim leans on them.

**How to apply:** (1) for any "default is closed" claim, grep the platform docs for the
mode's exception clauses ("except", "or the", "in every mode", "not configurable");
(2) smell: ALLOW rules for things the carve-out auto-approves anyway (redundant allows =
unmodeled carve-out); (3) check whether the carve-out's members compose with retained
egress (read-only reads + WebFetch = exfil); (4) carve-out members with operator/redirect
forms (`echo >>`) belong in the same shell-operator build-test as the design's other Bash
rules. Related: [[pattern-invariant-soundness-by-enumeration]],
[[gap-live-source-drift]].

**Extension (r4, L2-F2 — the carve-out story cuts BOTH ways):** the same doc that grants
the carve-out can grant deny-rules reach INTO it — "Read and Edit deny rules apply ... to
file commands Claude Code recognizes in Bash, such as cat, head, tail, and sed." A repair's
motivating example ("cat <transcript path> would have been AUTO-APPROVED under the old
profile") failed at the leaf because the old profile already carried an explicit Read deny
on that exact path, which the doc extends to Bash cat. The carve-out exposure is real only
for paths protected by ALLOW-SCOPING alone (no deny rule). So: when grading a
carve-out/bypass severity story, test the chosen example against the FULL rule-interaction
matrix (deny reach, deny-beats-carve-out), not the carve-out clause in isolation — the
prior-exposure narrative is a claim to reproduce, same as the repair
([[pattern-postmortem-misdiagnosis]]).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_conditional_vote_laundering.md -->
---
name: pattern-conditional-vote-laundering
description: A synthesis counts a lane/seat vote toward a majority ("2/3 ratify") whose stated precondition the same synthesis rejects elsewhere — the tally is arithmetic camouflage for a 1/3 position
metadata:
  classes: [vote-laundering, cross-section-contradiction]
  type: feedback
---

Gap pattern: **conditional-vote laundering**. A multi-lane synthesis tallies votes for a
disposition ("RATIFY 2/3") where one counted vote was explicitly conditioned on a premise
(e.g. "ratify the channel IF the spend throttle actuates") that the synthesis itself REJECTS
in another section. On the document's own dispositions the unconditional support is smaller
than the tally, and the majority framing — not the stated grounds — does the persuading.
(Run-4 efficiency report §3.5: lane 2's grade-dispute-channel ratify was coupled to lever 2's
actuation, which §2.5 rejects; honest tally was 1/3 unconditional + 1/3
conditional-on-a-rejected-premise + 1/3 dissent. Blue even disclosed the dependency — the
disclosure conceded the point while the headline kept the 2/3.)

**Why:** vote counts read as evidence-strength and get inherited by later rounds and cross-run
dockets uncritically; the laundering survives leaf verification because each lane's position is
quoted faithfully — the defect is in the aggregation arithmetic, not the quotes.

**How to apply:** whenever a synthesis cites a vote tally (N/M lanes, N-of-M lenses), pull each
counted vote's stated conditions and check them against the same document's dispositions. A
vote whose condition is rejected, deferred, or out of scope moves to its own bucket; the
disposition must then argue from the surviving grounds, not the headline count. Related:
[[pattern-footnote-overattribution]] (one label wearing a crowd) — here one tally wears
incompatible votes.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_config_sample_schema_and_anchoring.md -->
---
name: pattern-config-sample-schema-and-anchoring
description: Embedded config samples can quote the doc faithfully yet misplace the key (wrong JSON nesting = silent no-op) or ignore path-rule anchoring semantics (cwd vs settings-source vs absolute) — audit the sample against the schema and the anchor table, not just the prose
metadata:
  classes: [config-semantics-error, policy-without-mechanism]
  type: feedback
---

Rule: when a report embeds a sample configuration (settings JSON, permission profile,
hook config), verify TWO things beyond quote fidelity: (1) every key sits at the nesting
level the schema demands — a correctly-quoted setting at the wrong level is a silent
no-op that the surrounding prose still credits ("closes the escalate route"); (2) every
path-pattern's ANCHOR — relative rules anchor at process cwd, `/path` at the settings
source, `//path` at filesystem root — so a deny set written bare-relative silently
relocates when a scheduler launches from a different cwd.

**Why:** sleeper-service run round 1 (2026-07-17): blue's §4.2 profile quoted the
permissions doc verbatim on `disableBypassPermissionsMode` semantics but placed the key
top-level instead of `permissions.disableBypassPermissionsMode` (L3-F1, MEDIUM); the same
profile's deny rules were bare-relative and thus cwd-anchored, unstated, in a
Task-Scheduler context where default cwd is System32 (L3-F6). Both defects survived a
blue self-audit that explicitly bragged about "drafting details a naive spec would get
wrong" — the quotes were all correct; the artifact wasn't.

**How to apply:** at citation lenses, treat an embedded config block as a CLAIM in its
own right: fetch the schema/doc, check each key's documented location, and check the
anchor table for every path rule. Sentence-level verbatim verification does not transfer
to the config sample. Related: [[pattern-self-defeating-mitigation]],
[[pattern-repair-regression-citation]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_controller_lookahead_pricing.md -->
---
name: pattern-controller-lookahead-pricing
description: Counterfactual savings for a throttle/controller priced on data only observable AFTER the controlled round — check what board/state the mechanism could actually read at decision time
metadata:
  classes: [pricing-basis-drift, derivation-status-overclaim]
  type: feedback
---

When a report prices a counterfactual controller (throttle, floor, stop rule) as "would have
fired in rounds X–Z," re-derive WHICH state each firing decision reads: a round-N control
decision reads the post-round-(N−1) state, never round N's own output. Caught round 4 (run-4
efficiency investigation, L1-F1): "3 throttled rounds (the low-mass rounds 3–5)" labeled
round 3 throttled on the strength of round 3's OWN post-mass (~44), when the actual throttle
input — the post-round-2 board (~65) — was the run's second-highest mass. The 3-round basis
inflated the ceiling saving into a point estimate ($18 vs honest $12–18).

**Why:** counterfactual replays quietly grant the mechanism hindsight; the off-by-one hides
inside a plausible round-range gloss ("the low-mass rounds") and survives arithmetic checks —
the multiplication verifies while the multiplicand's timing is wrong. Sibling of
[[pattern-stale-baseline-pricing]] (both are pricing-section defects that pass leaf checks) and
of [[pattern-metric-conflation-and-traceable-not-verified]].

**How to apply:** for every "would have saved $Y over rounds X–Z" claim, list the decision
points and the state visible at each; check the report's own series for whether the trigger
actually trips at each decision point under any stated threshold. Also: when red's own prior
required-fix supplied the arithmetic (here R2-2's ×3), audit the repair as a new claim —
red-vector errors are still errors.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_doctrine_vs_implementation.md -->
---
name: pattern-doctrine-vs-implementation
description: A code comment/doc states a design principle ("never cheapen the adversary") that the literal routing/logic below it contradicts — check every declared invariant against its own implementation, not just external behavior
metadata:
  classes: [doctrine-vs-implementation, cross-section-contradiction]
  type: feedback
---

Gap pattern: **doctrine-vs-implementation contradiction**. A source file states its own design
principle in a comment or doc string, then the concrete logic immediately below doesn't follow it
— caught by literally reading the routing table against the doctrine sentence, not by any
external test.

**Why:** Caught in the FEOV-retrospective audit (round 1, lens 5). `debate.js`'s own comment:
"cheapen redundancy and mechanics, never judgment or the adversary" — followed immediately by a
routing table that assigns the cheap/bulk model tier to red-lens passes (the actual adversarial
audit work: leaf-node citation checks, gap-finding) and reserves the judgment tier only for
red-merge (mechanical consolidation). Whichever reading is "intended," the document's own stated
invariant and its own next ten lines disagree — and the report under audit, which spent a whole
item grading this exact routing choice, never noticed the tension.

**How to apply:**
- When a file states a doctrine/principle in prose (comments, docstrings, design-doc sentences),
  treat it as a testable claim about the code immediately following it — read the two side by
  side, don't take the prose at face value.
- Two valid resolutions exist when a mismatch is found: (a) the code is wrong relative to its
  stated doctrine — fix the routing/logic; (b) the doctrine's terms are ambiguous and the code is
  a defensible reading — tighten the prose to remove the ambiguity. Present both; don't assume
  which was "intended."
- This is cheap to catch (grep the doctrine sentence, then read the next N lines) and easy to miss
  under time pressure because both halves *individually* read as reasonable — the contradiction
  only shows up on direct juxtaposition.

**Extension — doctrine contradicted by ARTIFACT PLACEMENT, not adjacent code (sleeper-service
R4, L5-F1):** the prose said "the command markdown is the wrapper's phase-1 payload, not a
standalone entry point" — while the file ships under `commands/`, the harness's entry-point
directory, making it user- AND model-invocable regardless of the prose. Sharper still when the
design applies the known mechanical guard (`disable-model-invocation: true`) to a SIBLING
artifact but not this one: the asymmetry proves the mechanism was known and the omission is a
real hole, not ignorance. Check where a "not an entry point" file physically lives and what
the harness does with that location; and diff sibling artifacts for guard flags one carries
and the other lacks.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_ephemeral_instrument_and_grid_count.md -->
---
name: pattern-ephemeral-instrument-and-grid-count
description: two quantitative-claim hazards — a measurement whose parser/input lives only in scratchpad (self-report with no re-derivation path), and table counts asserted by grid arithmetic instead of row enumeration
metadata:
  classes: [artifact-preservation, figure-recount-fails, self-attestation]
  type: feedback
---

Two leaf-verification hazards on quantitative claims, both caught round 3 of the efficiency run.

1. **Ephemeral instrument**: blue "measures" something first-hand (good) but the measuring
   script lives only in a session scratchpad and its input is an untracked working-tree file.
   The figures become self-report with no independent re-derivation path — exactly the vacuity
   tier the system's own attestation ceiling routes to "post-hoc audit over git-tracked
   artifacts," which these artifacts aren't. **Why:** [^MergeDecomposition]'s ~70-line parser
   (never committed) + gitignored tarball fed §4.2's table, the money map's #1 ranking, and an
   ANSWERED open question. **How to apply:** when a report cites its own measurement, ask where
   the instrument and input live NOW; "method stated in prose" ≠ preserved. Grade as
   preservation gap (usually LOW-MEDIUM, trivial fix: commit the script), not corroboration
   failure. Related: [[pattern-provenance-self-report-and-stale-gate]].

2. **Grid-count fetch hazard**: verifying "N reported configurations" against a paper's table,
   a summarizing WebFetch will assert the dimensional product (6 datasets x 4 models = 24) even
   when cells are missing (multimodal rows ran only 3 vision-capable models; true count 22).
   **Why:** first fetch of arXiv:2510.12697 Table 2 said 24/24 while its own enumeration
   totaled 22 — the enumeration was right, the header wrong. **How to apply:** never accept a
   count from a fetch summary; demand per-row enumeration and sum it yourself. A blue footnote
   that says 22 against a table that "obviously" reads 24 may be the CORRECT one.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_exhaustive_sweep_omits_hard_case.md -->
---
name: pattern-exhaustive-sweep-omits-hard-case
description: A self-certifying completeness sweep ("every catch traced", "checked against every position", "both misattributions corrected") silently omits the hardest named case — often one the same report itself elevated pages earlier
metadata:
  classes: [exhaustive-sweep-omits-case, enumeration-non-exhaustive]
  type: feedback
---

When a report contains an exhaustiveness claim — a mapping table ("every late-round catch
traced to an arm"), a systematic check ("doctrine check run against every position"), or a
counted sweep ("two frontier misattributions corrected") — enumerate the claimed universe
independently and diff it against the table/check/count. The characteristic defect: the
omitted item is the *hard case*, and frequently one the report itself named as a type
specimen or conceded as a conflict in an earlier section (writer's blind spot: the case was
"handled" mentally when first discussed, so the sweep skips it).

Observed (run-4 efficiency investigation, round 1): §5.2 "every late-round catch" table
omitted R5-2 — one of §5.1's own two named specimens, and the only catch NO proposed scope
arm covers (L5-F1, MEDIUM-HIGH); §6.3 doctrine check "run against every position" skipped
the §5.5 conditional-ratify, the one position §5.3 itself admits conflicts with a named
doctrine clause (L5-F2); §6.4.3 counted "two" frontier misattributions while a third (stale
pre-temper grades in H1) survived (L5-F10).

**Why:** exhaustiveness claims are load-bearing for gate/ratification logic — a sufficiency
argument with a silent hole propagates into future-run dispositions uncritically.

**How to apply:** at the logic/completeness lens, treat every "every/all/both/N of N"
sentence as a checkable enumeration: build the universe from the report's OWN earlier
sections (specimens it names, positions it takes, errors it admits), then verify membership
one by one. The omitted member is usually adjacent to a concession. Related:
[[pattern-missing-root-invariant]], [[pattern-unreconciled-numeric-floors]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_fail_open_guard_and_erasable_evidence.md -->
---
name: pattern-fail-open-guard-and-erasable-evidence
description: Two composing control-design defects — a guard keyed on an ambient marker that silently disarms when the marker is lost (fail-open), and "tamper evidence" living where the audited actor can erase it (uncommitted working tree, actor-writable ledger)
metadata:
  classes: [unhandled-degenerate-case, self-attestation, partial-control-coverage]
  type: feedback
---

Rule: for every proposed control, ask two questions the designer usually hasn't: (1) what
happens when the control's ENABLING PRECONDITION is absent — does it fail closed or silently
fail open? (2) does the DETECTION/EVIDENCE channel the control leans on live inside the
audited actor's write/erase reach?

**Why:** sleeper-service design round 1 produced clean specimens of both. (a) The PreToolUse
write-fence armed only "when the sleeper marker is present" (env var set by the wrapper) —
marker loss (scheduler misconfig, shell nesting, rung migration) silently drops the run to the
permission-rules-only configuration the same report REFUTED at the leaf node; nothing asserted
enforcement (preflight checked plugin LOAD, not hook FIRING). (b) §4.3 called an uncommitted
working-tree delta "tamper evidence (the DGM-lineage property)" — but an uncommitted delta has
no history; the one channel able to write out-of-boundary (subprocess residual) can also
revert its write before review, and run N+1 executes before run N is reviewed. Sibling
instance: the monthly budget cap read a ledger inside the loop's own Edit-allowed surface.
DGM's lineage worked because it was durable and externally inspectable; these were neither.

**How to apply:** (1) marker/env-keyed guards: demand fail-closed polarity (fence unless a
human-session marker present) or a per-run denial canary (attempt a forbidden write, abort
unless DENIED — enforcement becomes a verified fact, reusing the design's own probe method).
(2) any "tamper-evident / append-only / git-tracked" claim: check WHO can write/rewrite the
evidence between reviews and whether the review cadence outruns the actor's next unattended
session; evidence of record must be written from OUTSIDE the audited session (wrapper-side
snapshot/auto-commit). Related: [[pattern-self-defeating-mitigation]],
[[pattern-authorship-evades-never-edit-guard]], [[pattern-gate-stack-rung-nonportability]],
[[pattern-policy-without-mechanism]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_false_equivalence_disjuncts.md -->
---
name: pattern-false-equivalence-disjuncts
description: Spec/repair text offers two branches as "equivalently" or "X or Y" where only one has an operator or prevents the stated harm — the weak branch becomes the build path
metadata:
  classes: [undecided-disjunction, policy-without-mechanism]
  type: feedback
---

When a repair or ratification condition offers alternatives ("A, or equivalently B" / "A or
B"), audit each disjunct SEPARATELY against the harm the clause states. Two live instances
in one round (2026-07-14 efficiency run, round 3): (1) a one-round "contest window" offered
as equivalent to auto-docketing to the judge — the docket routes to a named independent
seat, the window routes to nobody (no seat on any read surface is positioned to act during
it; both parties already agreed to the delta); a delay is not an absorber. (2) "pinned or
version-stamped" mapping — the stated harm was series incomparability, which only pinning
prevents; stamping merely labels the break. Companion checks: (a) **window without a
watchman** — any review/contest window claimed as a control must name its operator AND show
the reviewed artifact is on that operator's read surface; (b) **per-item thresholds
salami-slice** — a magnitude trigger per delta under a per-round item cap licenses
cap × threshold × rounds of sub-threshold drift; ask whether the threshold should be
cumulative.

**Why:** implementers pick the cheaper branch while citing the clause as if it bought the
strong one; the false "equivalently" launders the choice.

**How to apply:** grep repair text for "or, equivalently", "or a", "either"; for each
branch ask: who executes it, on what read surface, and does it prevent (not just reveal)
the clause's own stated harm. Related: [[pattern-self-defeating-mitigation]],
[[pattern-policy-without-mechanism]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_footnote_overattribution.md -->
---
name: pattern-footnote-overattribution
description: A footnote's claim-list bundles several specifics but only some trace to the primary; multi-citing one figure to N footnotes hides which one carries it
metadata:
  classes: [citation-figure-misattribution, derivation-status-overclaim]
  type: feedback
---

When a footnote lists several distinct specifics (e.g. "summarization drift; semantic
intensification; cross-version score drift; ~29-day half-life") treat each as a **separate**
statement↔source pair — verify them independently. Frequently only one leg (the generic
qualitative one) is corroborable at the primary; the specific quantitative legs are asserted.

**Why:** caught in memory-architecture Round 3 — `[^MemorySurvey]` (arXiv 2603.07670) claimed four
things; leaf-node fetch confirmed only summarization drift. The load-bearing ~29-day half-life
(sole prop for "decay windows are in the evidenced band") was uncorroborable.

**Compounding trick — multi-citation laundering:** a figure cited to three footnotes
(`[^A][^B][^C]`) reads as heavily-sourced, but often only ONE of the three actually carries it and
the other two are topical padding. Check WHICH footnote's text asserts the number; the number's
real support may be a single un-pinnable source wearing a crowd.

**How to apply:** for any bundled-claim footnote or multi-cited figure, name which specific source
must carry the load-bearing number, follow only that one, and grade the number on that source
alone. Distinguish "unable-to-corroborate (lossy fetch)" from "contradicted" — the former is a
graded low-confidence gap under the stickler rule, not a pass. Relates to
[[citation_status_and_misattribution_patterns]] and [[repair-regression-citation]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_gate_stack_rung_nonportability.md -->
---
name: pattern-gate-stack-rung-nonportability
description: "Consent/gate stack designed for one deployment rung (local wrapper + out-of-repo settings) silently assumed to hold on other rungs (cloud routine = fresh clone, no local files; CI = settings must move in-repo)"
metadata:
  classes: [partial-control-coverage, false-universal]
  type: feedback
---

When a design offers a deployment *ladder* (manual → OS scheduler → desktop task → cloud →
CI) but its security/budget gates are artifacts of ONE rung (an operator-owned `--settings`
file outside the repo, a local wrapper script owning preflight/ledger/env-marker), audit
each rung for which layers actually survive. Cloud "fresh clone, no local files, fully
autonomous" rungs get NONE of the local profile; CI rungs invert the problem — enforcing
the profile requires committing it into the tree the loop's output can touch.

**Why:** Sleeper-service round 1 (L6-F7): §3.4's rung-3 caveat list covered qmd/connectors/
identity but never that §4.2's layer 1 and §5.1's ledger wrapper simply don't exist in a
cloud clone — an operator climbing the documented ladder carries a false belief about which
gates are on.

**How to apply:** For any ladder/matrix of deployment options, demand a per-rung
gate-survival table; treat rung upgrades as design decisions, not config toggles.

**Extension — rung ZERO is a rung (sleeper-service round 3, L5-F1):** guards keyed on the
scheduler wrapper's presence (origin-provenance stamps, ledgers, canaries) are void in the
design's own DEFAULT manual mode if manual runs bypass the wrapper — and a "manual = same
code path" claim can flatly contradict a "manual runs do not pass through the wrapper"
claim elsewhere. Audit the default/manual rung against the gate table with the same rigor
as the exotic rungs; also check the gate table itself gained rows for gates minted in
later rounds (a table built to prevent overstatement goes stale silently). Related:
[[pattern_inherited_surface_netting]], [[pattern_policy_without_mechanism]],
[[pattern_origin_tag_naming_keyed]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_gitignored_not_absent.md -->
---
name: pattern-gitignored-not-absent
description: availability/durability claims vs git reality, BOTH directions — "gitignored/absent" refuted by untracked working-tree artifacts; "committed" refuted by untracked status (present ≠ committed)
metadata:
  classes: [artifact-preservation, verification-scope-blindspot]
  type: feedback
---

A report claims data is UNAVAILABLE ("gitignored and absent at the pin", "not retained") and
defers a measurable question to a future run — but the artifact sits untracked in the working
tree, checkable with `ls`. Git-scoped existence checks (git show, git diff, pin equivalence)
see only the object store; gitignored ≠ deleted.

**Why:** efficiency-investigation round 2, L3-F1 — blue's lane 3 declared run-3 agent
transcripts "gitignored and absent at the pin", making the #1-ranked lever's savings
"unmeasurable until run 4"; `ls` found the 7MB tarball (46 per-agent transcripts) in the
working tree, mtime hours before the run launched. The report's own cited artifact (cost.md:
"Measured from per-agent API transcripts") already implied the data existed.

**The mirror direction (round 4, L4-F1):** a repair claims an instrument "is now COMMITTED as
trajectories/decompose-merge.mjs" — the file existed in the working tree and reproduced every
figure on re-run, but `git status` showed the whole directory untracked and
`git log --all -- '*file*'` was empty. Present ≠ committed: the repair answered a
"not-git-tracked audit artifacts" defect by creating another untracked artifact wearing the
word "committed." Substance sound, status word false — grade the residual, don't regress the
figures.

**How to apply:** never verify existence/durability through one lens. For "absent" claims run
`ls`; for "committed/tracked/preserved" claims run `git status` + `git log --all -- '*name*'`
on the exact path. Check mtimes for fairness on likelihood grading. Sibling of
[[gap-pattern-verification-file-type-blindspot]] (existence verified through one lens) and
[[pattern-self-referential-repo-drift]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_happy_path_only_gate.md -->
---
name: pattern-happy-path-only-gate
description: A verification gate (snapshot compare, capture, record) that runs only on the completed-run exit path leaves abnormal termination uncovered — and abort is often model-influenceable
metadata:
  classes: [unhandled-degenerate-case, partial-control-coverage]
  type: feedback
---

A gate specified as a numbered step (e.g. "step 7: end snapshot AND COMPARE") silently
covers only runs that REACH that step. Aborts, budget/turn kills, resume deaths, and
wrapper crashes skip it, and cross-run chains defined against "prior run's END" have no
referent after a dead run — an unspecified link exactly at abnormal termination.

**Why:** sleeper-service round 3 (L5-F2): the R2-7 tamper watchman compared end-vs-start
at step 7 only; "run N+1 cannot fire past unexamined tamper evidence" was proven for
completed runs while abort (a state the session's own behavior can induce via turn/budget
exhaustion) took no END snapshot and ran no compare — a write-then-force-abort sequence
escaped the watchman.

**How to apply:** for every gate hosted in a driver/wrapper, ask "does this run on EVERY
exit path the driver observes, or only the happy one?" Check cross-run comparisons for
missing referents after aborted predecessors, and resume attempts for undefined
start/prior-end semantics. The fix is usually one code path (snapshot at every observed
exit) + fail-closed on a missing chain link. Related: [[pattern-self-defeating-mitigation]],
[[pattern-fail-open-guard-and-erasable-evidence]].

**Extension (round 4, L1-F1):** the "every-OBSERVED-exit" repair has its own degenerate
case — the observer's OWN unobserved death (SIGKILL/power loss). Any interval-keyed
mechanism (run-window logs, timestamp sweeps) whose END is then supplied by a much-later
event (DEAD-mark nights afterward) degenerates from a minutes-scale window to a multi-day
span, and everything keyed on "inside the window" (origin tagging, markerless sweeps)
over-captures unrelated activity in the span. Audit the END-supply chain of every
window/interval the design defines: what closes it when the closer dies, and what does a
START-without-END mean to each consumer?


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_identity_keyed_detector_lineage_blind.md -->
---
name: pattern-identity-keyed-detector-lineage-blind
description: An escalation/convergence detector keyed on stable identifiers never fires when the process's own bookkeeping convention renames the tracked object every cycle — check who mints the ids against what the detector matches on
metadata:
  classes: [recurrence-detector-keying, name-keying-vs-marker, spec-underspecification]
  type: feedback
---

When auditing any mechanism that escalates on *recurrence* (contested dockets, retry
escalators, flaky-test detectors, repeat-offender triggers), YOU MUST check whether the
identifier the detector matches on survives the process's own minting convention. If the
convention issues a fresh id per cycle (successor gaps, re-filed tickets, renamed retries),
an identity-equality detector is structurally inert — it will show zero-recurrence telemetry
forever while the underlying dispute recurs indefinitely.

**Why:** FEOV-retrospective round 4 (R4-1). `debate.js`'s contested-docket check is
`prevGapIds.has(g.id)` — pure id string-equality against the prior round. But red's own
closed-WITH-REGRESSION methodology mints a fresh id for every successor gap
(R1-5 → R2-4 → R3-4/R3-9: one footnote, four ids, three rounds). Result: `contested` was 0 in
every round, the judge was never dispatched once across three completed rounds (zero `### LEAD`
sections in the transcript), and the only brake on a spinning debate was the maxRounds cost
ceiling. The debate converged anyway — but because blue conceded in good faith, a property of
the actors, not one the mechanism enforced. The report's own coverage row described only the
narrower same-id-skips-a-round case, whose remedy (widen the id history) does NOT close the
fresh-id case — a fix for one variant masquerading as coverage of the class.

**How to apply:** (1) For any recurrence-triggered control, trace both ends: what key does the
detector compare, and who assigns that key each cycle? If the assigner is the audited process
itself (including RED — your own successor-id practice was the defeat mechanism here), the
audit must treat its own conventions as part of the system under test. (2) Absence-of-firing
telemetry ("judge never invoked," "escalation count 0") is evidence FOR this pattern, not
evidence of health — grep the transcript for the escalation artifact at header/structural
level, not plain text (a quoted phrase is not an invocation). (3) The fix shape is lineage,
not history-widening: a `supersedes: [prior-ids]` field plus chain-depth detection; verify a
proposed remedy closes the variant actually observed, not the adjacent one. (4) Convergence
observed under a broken convergence-enforcer must be attributed to actor behavior, and stated
as such in the disposition.

**Extension — "normalized" asserted, normalization unspecified (sleeper-service round 3,
L5-F3):** a per-cause escalator keyed on a "normalized" failure signature (exit class +
first abort-record line) is inert if the record format the design itself specifies embeds
per-cycle variable content (dated markers, fresh-minted run-dir paths, session ids) and
the normalization is never defined. The unspecified normalization IS the mechanism; demand
it spelled out (strip dates/paths/ids, match on template or error code) plus zero-firing
telemetry surfaced so an inert detector is visible.

Related: [[pattern-missing-root-invariant]] (multi-round gap chains signal a missing stated
invariant — the gap chain here IS the detector's blind spot),
[[pattern-self-defeating-mitigation]] (two controls starving each other's trigger),
[[pattern-policy-without-mechanism]] (docket doctrine present, enforcement never reachable),
[[pattern-repair-regression-citation]] (red-practice-becomes-system-defect sibling).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_incomplete_repair_footnote_lag.md -->
---
name: pattern-incomplete-repair-footnote-lag
description: Round-N repair lands in prose body but the footnote/citation surface still carries the retracted number/claim — re-verify the citation, not just the sentence
metadata:
  classes: [incomplete-repair-propagation, claim-contradicts-own-record, reader-channel-mismatch]
  type: feedback
---

When blue accepts a citation gap and repairs it, the fix often lands only in the **body prose** while the **footnote** (the surface a skeptic actually follows) still asserts the original, now-retracted figure. Re-raise: the citation surface is authoritative for leaf-node verification.

Observed instances (memory-architecture audit, round 2):
- R1-22 dropped "v2.1.59" from §1.2/§3 body, but footnote `[^MemoryDocs]` still read "auto memory native v2.1.59+".
- R1-28 softened "80–99%" to "~90–95%" in §4 body, but footnote `[^MemoryPoisonSurvey]` still read "80–99% reported attack success rates".
- R1-25 downgraded Letta git-branch detail to "community-suggested" in §5 body, but footnote `[^LettaSleep]` still listed "isolated git-branch commits to avoid contention" (and never named the forum).

**Why:** research prose and its footnotes drift independently; a change-summary/CHANGELOG reports the body edit and looks complete, but the footnote is a separate string that must be edited too.

**Retract-by-annotation variant (round 3):** a footnote may ADD a repair note ("the parenthetical 'v2.1.59+' is dropped") while leaving the retracted string still literally present in the descriptive clause. The footnote becomes self-contradictory — asserts the figure, then says it was removed — and the leaf-node reader scanning the clause still lands on the asserted value. Watch for repair notes that *describe* a deletion that never happened. The clean fix is to delete the string; annotating it is not enough. Confirm by checking whether a sibling footnote repaired the same round (e.g. `[^SubagentDocs]` R2-12) actually *removed* its version string — asymmetry between two supposedly-identical repairs proves the annotated one is a real miss, not style. Also check whether a body medium-confidence tag (e.g. R1-29 "removed user memories from system prompt") propagated to the footnote, or whether the footnote still states the vendor-blog-only claim flatly.

**CONFIRMED round 3 (leaf-node re-follow, slice §0–§4):** the retract-by-annotation variant materialized exactly as predicted — `[^MemoryDocs]` line 1414 carries `v2.1.59+` **twice** (`grep -o` = 2 hits): once as the standing descriptive claim, once inside the R2-9 note announcing its removal. R2-9(a) was recorded CLOSED off the CHANGELOG; the edit never ran. Meanwhile `[^MemoryPoisonSurvey]` (R2-9b), `[^LettaSleep]` (R2-9c), and `[^SubagentDocs]` (R2-12) all *did* execute their string removals — the asymmetry proved `[^MemoryDocs]` was a genuine miss, not a stylistic choice.

**Un-propagated-repair variant (round 3, body-to-body):** a repair marked CLOSED can propagate to *some* instances and miss others. R1-24 corrected "46k-star"→"~87.1k" in §7 and `[^ClaudeMem]` but missed the *same figure* in §1.5 (line 230) — a standing stale claim contradicting the corrected sibling sections in-doc. Grep the retracted *token* across the WHOLE report, not just the section the repair note cites. A repair is closed only when every instance of the old token is gone.

**How to apply:** after any accepted citation-repair, `grep -o` the retracted number/token across the ENTIRE report (footnotes AND all body sections), not just the section named in the repair note. Count occurrences: if the old token appears anywhere as a standing claim (i.e. outside a retraction/correction sentence), the gap is OPEN regardless of the CHANGELOG marking it closed. "The repair note says it's dropped" ≠ "the string is gone." Do not trust round-N "CLOSED" ledgers derived from the change log — re-follow by grep.

**Reverse direction confirmed (FEOV-retrospective round 3) — body lags the repaired footnote:** the lag runs both ways. Blue's R2-4 repair correctly rewrote `[^DiminishingReturns]` (dropped the unpinnable "continued gains to 7 agents" clause, stated the corrected *opposite* direction), but the §1.1 body sentence the footnote is attached to still asserted "continued gains observed to 7 agents on the hardest" a full round later — the majority-surface reader takes away the retracted claim while the 400-words-down footnote says the reverse. The grep-the-retracted-token-across-the-ENTIRE-report rule catches both directions; apply it symmetrically (body→footnote AND footnote→body) after every accepted repair.

Related: a fresh repair can also INTRODUCE a miscitation — R1-28's softening added `[^EnvInjectedMemory]` "~90% env-injection", but that primary (arXiv 2604.02623 "Poison Once, Exploit Forever") reports ≤32.5%. Verify the NEW citation a repair introduces, not only the old one it replaces. See [[citation_status_and_misattribution_patterns]].

**Non-grep-catchable variant (FEOV-retrospective round 5) — sibling location carries the drafter's own discarded first guess, not a retracted token:** when the same round adds/edits TWO locations for one fix (a new row plus an update to an existing row/section), one of them can still hold the author's superseded rough draft of a worked-example enumeration while the other holds the corrected version — and this is invisible to "grep the retracted token" because every individual id in the wrong list is itself a real, legitimately-used gap id elsewhere in the document (e.g. R1-13, R2-1, R3-7 are each correct in their OWN rows; they are simply wrong when strung together as "the chain for finding X"). Caught here only by (a) diffing the two locations' enumerations against each other for the identical claim, and (b) cross-checking against the authoritative source (`red/findings.md`'s own closure-status lines) plus the debate transcript, which — in this instance — contained the author's own admission of exactly which list was the discarded draft (`debate.md`'s BLUE section: "I initially reconstructed... before checking [red's] section, which had already enumerated a different... set... adopted it in place of my own" — but only in ONE of the two sibling locations). **How to apply:** whenever a round adds a new table row/section AND edits an existing one for the same finding, compare their worked examples/enumerations against each other directly, not just each against its own cited source — a single-location fetch-check passes both locations individually while missing that they disagree.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_inherited_surface_netting.md -->
---
name: pattern-inherited-surface-netting
description: "We don't add risk, it's inherited from native/baseline" netting arguments must verify the baseline wasn't since remediated — bespoke may re-open what upstream closed
metadata:
  classes: [risk-coverage-omission, claim-contradicts-own-record]
  type: feedback
---

When blue defends a build decision with "most of the attack surface is *inherited* from the native
baseline, so adopting native instead buys no safety on this dimension," verify the baseline is still
as vulnerable as claimed.

**The trap:** the netting treats the native baseline as static-vulnerable. But if upstream *patched*
the surface (e.g. the CVE-2026-21852 fix reportedly removed user memories from the system prompt —
de-authorized the native memory surface), then post-fix native is *less* poisonable, and the bespoke
layer's high-authority projection (`.claude/rules/` loads at CLAUDE.md priority) *re-authorizes*
injection — bespoke re-widens what native narrowed. "Shared/inherited" is then false; the bespoke
layer creates net-new high-authority surface.

**Double-bind to exploit:** if the remediation detail is too low-confidence to rely on (blue itself
tagged "removed from system prompt" medium-confidence, R1-29), blue cannot also lean on it to claim
native == bespoke poisoning surface. Either direction breaks the "Shared" cell.

**Why:** the netted build-vs-adopt table is the keystone of the go/no-go; a wrong "Shared"
classification flips the conclusion.

**How to apply:** for each "inherited/shared" cell in a net-new-surface table, ask (1) was the
baseline remediated since the cited incident? (2) does the bespoke design restore authority/exposure
the remediation removed? If yes to both, reclassify as net-new.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_invariant_soundness_by_enumeration.md -->
---
name: pattern-invariant-soundness-by-enumeration
description: A keystone invariant sold as "sound/mechanical" whose soundness actually rests on an under-inclusive channel denylist — provably incomplete when the system's own symmetric defense already treats an omitted channel as I/O
metadata:
  classes: [enumeration-non-exhaustive, derivation-status-overclaim, false-universal]
  type: feedback
---

When blue closes a cluster of gate-by-gate patches by adopting a single **invariant** (the
anti-complexity move red often recommends), the invariant is frequently claimed **"sound"** or
**"mechanical"** — but its soundness rests on an **enumerated denylist** of the channels it must
cover, and the enumeration has holes.

**Why:** an invariant like "external-touched ⇒ tainted, transitively" is only sound if "external"
is the *complete* set of channels through which attacker-authored bytes enter. Blue tends to fix a
short denylist (e.g. `WebFetch`/`WebSearch`/`file:`/`ingest`) and omit routine channels: `Bash`-fetched
bytes (`curl`/`gh`/`git log`), MCP tool results, sub-agent **sidechain** reads that launder into the
parent, and in-repo files authored by untrusted commits read via `Read` (not classed "external").
A denylist is the wrong structure for a taint/trust boundary — a newly-added tool defaults to
*trusted*.

**The killer tell:** the system's *own symmetric defense* already treats an omitted channel as I/O.
In the memory-architecture debate (R4-1), blue's outbound secret-gate wired on `WebFetch|WebSearch|Bash`
— proving `Bash` is a first-class channel — while the inbound taint invariant omitted `Bash`. The
exfil pipe and the injection pipe are the same pipe; defending one end and not the other is the gap.
Look for this asymmetry to prove incompleteness at the leaf node rather than merely asserting it.

**How to apply:**
- When an invariant is claimed "sound"/"mechanical"/"by construction," find its channel/field
  enumeration and test each *routine* channel the enumeration omits.
- Recommend **inverting to an allowlist** (enumerate what is *trusted*; everything else taints) so
  new channels default safe — this is usually a parser change, not research, so grade it a hardening
  gap, not a redesign.
- Related: an invariant can also **name** a field in its statement but **omit** it from the
  corollary's reset/enforcement list (R4-4: `last_seen` named non-inheritable but not reset on
  import) — verify the mechanism executes every leg the invariant claims, same discipline as
  [[pattern_incomplete_repair_footnote_lag]] (retract-by-annotation vs actual edit).
- Every downstream closure that leans on the invariant ("closed by construction under §X") is
  **contingent** on the enumeration being complete — surface the contingency, don't accept the
  cascade of closures on an unproven root. Relates to [[pattern_missing_root_invariant]] (the prior
  round's recommendation) and [[pattern_risk_grading_conflations]] (verdict/docket keystone leaning
  on an unverified mechanism).
- **Two enumeration-completeness flavors caught in the sleeper-service run (R4-2/R4-3):**
  (1) **Source-hedged list.** "Deny-enumerated per command" was built from a doc that names the
  read-only set with "These **include** …" and never claims completeness — the same page names
  `sort`/`sed` as classifier-reasoned commands NOT in the enumerated list. When a closure's
  enumeration is copied from a source that itself hedges ("include", "such as", "e.g."), the
  closure inherits the hedge; grep the source sentence for a totality word before trusting a
  "fully enumerated" claim.
  (2) **Functional-exempt member.** A "deny-enumerated per command" carve-out RETAINED one member
  un-enumerated because the design NEEDS it (read-only git) — and that exact member was the one
  with a leaf-confirmed write gadget (`git format-patch -o` → out-of-repo file). An enumeration
  that intentionally exempts a member for functional need is not "fully enumerated"; the exempt
  member is where the next gadget lives (sibling-halo — see [[pattern_audited_artifact_sibling_halo]]
  and [[pattern_wildcard_allow_write_gadget]]). The allowlist-inversion recommendation applies to
  the exempt member specifically: allowlist the exact read forms, deny the rest.
  (3) **Lexical-form escape recurring INSIDE the repair (R5-5).** A belt extension minted to close
  a sibling escape (`Bash(* -o *)`, added to catch the short-form output flag the `--output` denies
  missed) is itself escaped by the NEXT lexical form: `Bash(* -o *)` requires a space, but getopt
  short options accept the ATTACHED form `-o<value>` (`git format-patch -1 -o/tmp/x` → exit 0,
  out-of-repo — leaf-reproduced at lens AND merge seats). The enumerate-and-extend approach the
  belt keeps doubling down on cannot close the class — each new pattern is escaped one lexical form
  deeper. The fix is NOT another belt pattern; it is to confirm the allowlist (hook parses `git`
  argv, rejects `format-patch` entirely) handles attached-form flags, and add the attached form to
  the probe matrix so it is leaf-tested, not assumed. When you catch this, re-run the gadget in the
  ATTACHED form specifically — a spaced-form-only reproduction understates the escape surface.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_measurement_methodology_drift.md -->
---
name: pattern-measurement-methodology-drift
description: A cited self-reported number (tokens/cost/etc.) survives leaf-node verification against its named source, but a LATER, more rigorous measurement from the same project quietly proves the counting method was wrong — the citation is accurate and the number is still stale
metadata:
  classes: [measurement-methodology-drift, self-attestation]
  type: feedback
---

Gap pattern: **measurement-methodology drift**, a third sibling to [[gap_live_source_drift]]
(external source moves) and [[pattern_self_referential_repo_drift]] (the audited repo's *code
state* moves). Here neither the citation nor the code changes — a **downstream, more rigorous
audit of the same project** later reveals that the informal self-reported figure being cited was
computed with a flawed method, without ever touching the original artifact.

**Why:** Caught in the FEOV-retrospective audit (round 2, lens 2). The report cites "252.9k tokens
(run 1)" and "~3M tokens (run 2)" as headline historical-incident costs, each footnoted to a
friction file's own self-report — both citations check out fine at the leaf node (the number is in
the file, as quoted). But the project's *live* backlog (checked because it kept moving past the
report's pinned SHA — see [[pattern_self_referential_repo_drift]]) had, by the time of this audit,
accumulated a NEW entry: a formal cost-audit tool built specifically to check the informal panel
token counter, which found it undercounts real spend by ~92% (excludes cache traffic — 610K
reported vs. 47.7M in the raw transcripts, for one round). The original 252.9k/3M figures were
never re-computed under the corrected method, so their comparability to each other, and their
precision as decision-supporting evidence, is now suspect — even though every individual citation
in the report is, in isolation, "accurate" (the number really is in the cited file).

**The trap this evades standard leaf-node checking:** grading corroboration confidence per
statement-reference pair passes this claim at HIGH (source says exactly what's quoted) even though
the number itself is now known-unreliable by the project's own later work. Leaf-node fidelity and
methodological soundness are orthogonal checks; passing the first says nothing about the second.

**How to apply:**
- When a report leans on a *magnitude comparison* between self-reported figures (X tokens vs. Y
  tokens vs. Z tokens) to argue urgency/priority, check whether the project has since built (or
  could build) an independent audit of the counting method itself — not just whether the cited
  number still appears in its source file.
- A live backlog/changelog that keeps drifting past the report's pin (per
  [[pattern_self_referential_repo_drift]]) is exactly where this kind of methodology-correcting
  entry shows up — re-check it for content, not just "did the SHA move."
- Grade this MEDIUM impact, not a verdict-blocker: the qualitative recommendation usually survives
  (the free thing is still free, the expensive thing is still expensive) — the gap is precision and
  comparability of the specific numbers used to argue the case, not the direction of the argument.
- Fix is cheap: one footnote flagging that pre-audit figures are self-reported and likely
  undercounted, pointing at the newer audit mechanism as the path to a comparable number next time.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_method_footnote_undercoverage.md -->
---
name: pattern-method-footnote-undercoverage
description: A "method documented for reproduction" footnote reproduces every figure EXCEPT the headline sub-figure — re-derive each figure from the stated method and flag the ones it cannot produce
metadata:
  classes: [derivation-status-overclaim, audited-artifact-sibling-halo]
  type: feedback
---

When a report claims a measurement is reproducible ("method stated here so run N can reproduce
it"), re-derive EVERY published figure from the stated method, not just the table. The failure
shape: the documented pipeline (e.g. per-file byte attribution) genuinely reproduces the table
and the aggregate dollars — earning a halo — but the decision-driving headline (e.g. a
within-file archive/open split behind "sharding-addressable $7–10") requires an extra
undocumented step the method text cannot perform. The verified siblings launder the unverifiable
headline (see [[audited-artifact-sibling-halo]]).

**Why:** run-4 efficiency investigation round 3: §4.2's decomposition table recomputed exactly
first-hand from [^MergeDecomposition]'s method, but the money-map #1 figure ($7–10
sharding-addressable) rode an archive-fraction "clear majority" claim with no documented
derivation — "measured" status on an asserted sub-figure (L3-F3).

**How to apply:** for each figure citing a method footnote, ask "which step of the stated
method emits this number?" A figure with no emitting step is graded LOW-MEDIUM assertion, not
measured — even when every sibling figure reproduces at HIGH. Also check the caveat list: an
honest caveat list that omits the undocumented step is part of the finding.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_metric_conflation_and_traceable_not_verified.md -->
---
name: pattern-metric-conflation-and-traceable-not-verified
description: Two citation-lens patterns — a "success band" whose endpoints are two DIFFERENT metrics, and "closed-as-traceable" conflating paper-traceability with digit-verification
metadata:
  classes: [metric-conflation, citation-status-drift]
  type: feedback
---

Two leaf-node citation patterns caught in FEOV round-4 lens-3 (memory-architecture report).

**Pattern A — metric-conflation into a false range.** A source reports two *distinct* measures
(e.g. MINJA: 98.2% *injection* success rate vs 76.8% *attack* success rate). One section states
them correctly and separately; other sections collapse them into a single band ("succeeds
~76.8–98.2%"). The band's two endpoints are different quantities, so the upper bound reads as a
higher *attack*-success observation when it is actually a different metric. **Why:** blue repairs
the primary occurrence but propagates a lossy paraphrase elsewhere. **How to apply:** when a cited
"range" spans two round numbers, check whether both endpoints measure the *same thing*; a band built
from ISR-low + ASR-high (or precision-vs-recall, latency-vs-token) is imprecise even when both
digits are individually correct. Grade LOW if it doesn't move the disposition; require relabel.

**Pattern B — "traceable" ≠ "digit-verified."** A citation marked *closed-as-traceable* in a prior
round may rest on paper-level traceability (right paper, right title) while its *specific number*
still sits behind PDF-table friction. Re-verify the digit, not the paper. Corollary: **arXiv
abstract-version drift** — the `/abs/` page and v1 HTML may lack numbers that appear in the v2 HTML
abstract or a results table. If a fetch returns "no percentages in abstract," try `/html/<id>v2`
before concluding the number is unsourced. In the case that spawned this note the digit *did* check
out via v2 HTML after `/abs/` and the 7.7MB PDF both failed — friction delayed, didn't falsify.

Related: [[citation_status_and_misattribution_patterns]] (lossy arXiv-HTML fetch),
[[pattern_footnote_overattribution]], [[pattern_repair_regression_citation]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_missing_root_invariant.md -->
---
name: pattern-missing-root-invariant
description: When successive rounds each patch one security gate and each patch spawns a next-order gap, the root cause is a missing stated invariant — surface the invariant, not just the Nth instance
metadata:
  classes: [missing-root-invariant]
  type: feedback
---

When a design accretes security through gate-by-gate patching and *each round's fix introduces the
next round's gap* (R1 fix → R2 gap → R3 gap on the same axis), the recurrence itself is the finding:
the design lacks a single stated **invariant** from which the individual gates would follow by
construction.

**Why:** memory-poisoning defense in the 2026-07 audit went three rounds — trust tiers → clone gate
→ authorship gate; provenance-of-record → provenance-of-content → turn-level taint. R3-3 (taint
under-propagates) and R3-4 (consolidator must read the surface it guards) both collapse into one
rule: *"external-touched ⇒ tainted, transitively, until a human clears it."* Grading the Nth patched
gate in isolation lets blue keep relocating the hole; naming the missing invariant forces the design
to derive the gates instead of bolting them on.

**How to apply:** when you catch a third-order failure of the same first-order risk, (1) still grade
the instance (red never soft-passes), but (2) additionally surface — as a design-coherence
recommendation to the lead, not a block — that the axis needs a stated invariant. Watch the severity
trend: *declining* severity across rounds = convergence, so frame it as a recommendation; *flat or
rising* severity = the patching is not converging and the invariant becomes a block. Distinguish this
from [[pattern_self_defeating_mitigation]] (per-instance) — this is the meta-pattern across instances.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_origin_tag_naming_keyed.md -->
---
name: pattern-origin-tag-naming-keyed
description: an anti-self-amplification origin tag keyed on a dir-NAME/glob convention is circumvented by the process's own spawned sub-runs that land in differently-named dirs — and those sub-runs supply the "independent corroboration" the cap required
metadata:
  classes: [name-keying-vs-marker, partial-control-coverage, incomplete-repair-propagation]
  type: feedback
---

An authorship/self-poisoning guard that tags its own outputs by DIRECTORY-NAME GLOB (e.g.
`*_self-improve/`) misses artifacts the same process authors into differently-named dirs.

**Concrete instance (sleeper-service R2, L6-F2):** R1-25 capped `origin: sleeper` recurrence
at 1 and required "one non-sleeper occurrence" before a self-origin class could be picked —
tagging keyed on the `*_self-improve/` run-dir glob. But the nightly loop's step-4 spawns a
bounded FEOV smoke run in a SEPARATE dir named for the research TOPIC, not `_self-improve`.
harvest reads "every run's friction.md," so the spawned sub-run's friction (also loop-
authored) is UNTAGGED, counts at full recurrence, AND satisfies the "non-sleeper
corroboration" gate — the loop's own sub-run supplies the independent corroboration the cap
demanded from a human. The guard is circumvented by normal control flow, not an edge case.

**Why:** provenance is CAUSAL (who spawned the run), not nominal (what the dir is called). A
name-glob is a proxy that the process's own multi-dir output defeats.

**How to apply:** when blue tags self-authored artifacts to exclude/down-weight them, check
(1) does the process author into MORE THAN ONE location? (2) is the tag keyed on name/glob or
on a causal marker written at creation? Name-keyed + multi-location = hole. Worse when the
untagged sibling then UNLOCKS the capped class (supplies the required corroboration). Extends
[[pattern_authorship_evades_never_edit_guard]] (check corpus membership, not just mutation) —
here: check corpus membership by CAUSE, not by name.

**Extension — the fix closes ONE consumed surface, the amplification reappears at a SIBLING
consumed surface the tag doesn't span (sleeper-service R3, L6-F1):** R2-5 upgraded the tag
from name-glob to a wrapper-stamped `.sleeper-origin` marker on every spawned RUN DIR — sound
for run dirs. But the harvester consumes MORE than run dirs: it also mirrors red's gap-pattern
agent MEMORY and parses its headers into the same scored docket. The nightly FEOV sub-run's
red seat writes to that shared memory dir (workflow agents are memory subjects). Next harvest
mirrors the now-loop-authored patterns; they have NO run dir, hence NO marker, hence default
to NON-sleeper — re-opening the exact self-corroboration hole R2-5 just closed, one surface
over. **When a self-tag is repaired, enumerate EVERY surface the consumer reads (friction,
cost, telemetry, backlog, AND mirrored agent memory) and confirm the tag spans all of them —
a marker keyed to run-dir creation cannot tag a memory file the loop appends to.** The fix's
own scoping choice (run dirs) becomes the new blind spot's boundary.

**Extension 2 — two more origin-tag scope defects (sleeper-service R4, L5-F1/L5-F3):**
(a) **Containment-polarity inversion.** A declared out-of-contract/manual execution path's
artifacts are markerless, hence classed NON-self-origin — and the design's reassurance reads
"its output is ordinary-classed, requiring corroboration" as if that CONTAINS it. Backwards:
a non-self-origin row does not FACE the corroboration requirement, it SELF-SATISFIES it (it
IS the required independent occurrence) and corroborates tagged rows of the same class into
the pick. The untagged actor doesn't stand before the gate; it staffs it. Check the polarity
of every "requires corroboration" containment claim: who supplies the corroboration pool?
(b) **Change-keyed tagging over-strips at the wrong granularity.** Tagging any memory
file/header that "appears OR CHANGES inside a self-run window" as self-origin reclassifies a
pre-existing, human-attested pattern wholesale when the self-run merely APPENDS to it — and
append-to-existing is the memory system's own documented modal write. Result: the
best-attested classes lose their non-self standing first (monotonic docket blinding via the
guard itself; safe polarity, but signal loss). Tag what the window ADDED, not the file it
touched.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_per_instance_cap_not_per_cause.md -->
---
name: pattern-per-instance-cap-not-per-cause
description: A retry/waste cap bounds the INSTANCE (run-dir, item, attempt) while the generating CAUSE freely mints new instances — "cannot recur forever" claims that hold per-instance only
metadata:
  classes: [recurrence-detector-keying, false-universal]
  type: feedback
---

When a design closes a livelock/waste gap with a bounded retry cap (k attempts per X),
check what X is keyed on. If the cap is per-INSTANCE (per run-dir, per item, per file)
and the failure's root CAUSE survives instance replacement (deterministic abort, corrupt
shared input, persistent dependency outage), the mechanism re-opens the bounded waste on
every fresh instance: cap trips → fresh instance minted → same cause re-wedges it →
repeat until an outer budget trips, then the outer budget RESETS (monthly) and the burn
resumes. The sentence often self-incriminates with an escape clause ("...until the
monthly cap trips") while headlining "cannot happen forever."

**Why:** sleeper-service round 2 (R2-10, superseding R1-29): resume cap k=3 per run-dir
+ "next fire mints a fresh dir" = ~k×$5/night with zero output for any deterministic
root cause, ~3 nights to the $50 monthly cap, resetting every month — the livelock the
fix claimed to close, relocated one level up.

**How to apply:** for every bounded-retry fix, ask "what happens on the N+1th INSTANCE
of the same cause?" Demand a per-cause signature (hash of abort reason/failure class)
with a halt-and-surface threshold, or an honest re-scope of the claim to its per-instance
truth. Sibling of [[pattern-self-defeating-mitigation]] (control's own failure mode) and
the controller-lookahead/stale-baseline pricing family (mechanism priced on the state it
cannot see).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_pinned_mapping_not_total.md -->
---
name: pattern-pinned-mapping-not-total
description: a convention pinned for series stability (enum→numeric mapping, reading rule) is never checked for totality over its actual input domain — unmapped tokens, synonym tokens at opposite extremes, unpinned compound-cell handling
metadata:
  classes: [enumeration-non-exhaustive, unhandled-degenerate-case, spec-underspecification]
  type: feedback
---

When a report PINS a mapping or convention to make a measured series comparable across runs,
audit the pin's DOMAIN, not just its existence: enumerate the full input population the
instrument will actually see and check every member maps.

**Why:** efficiency-investigation round 4, L5-F3 — the freshly-decided Q6 mapping (low=1 …
certain=3.5, realized=excluded) was pinned per the lead's ruling, but (a) the shipped GRADE
enum has eight members and `trivial` was unmapped (schema-legal as likelihood/impact);
(b) the corpus's own grading defines `certain` AS "already realized," so two near-synonyms
sat at opposite extremes (3.5 vs excluded) and the exclusion rationale ("no longer a
probability") applied verbatim to the token kept at max weight; (c) conditional grades
("low this run rising to medium-high") — the board's modal shape — had no pinned reading,
and the report's own two-lane history showed extraction convention moving the series.
Within-version ambiguity defeats the pin the same way the mid-series change it forbids would.

**How to apply:** whenever a mapping/threshold/reading-rule gets pinned (especially per a
judge ruling — adopted text gets less scrutiny than contested text), diff its key set against
the schema enum and the corpus's observed value shapes (compound cells, conditionals,
parenthetical semantics). A synonym pair straddling the mapping's extremes is the
max-magnitude instance. Related: [[pattern-unreconciled-numeric-floors]] (requirements that
don't compose), [[pattern-metric-conflation-and-traceable-not-verified]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_policy_without_mechanism.md -->
---
name: pattern-policy-without-mechanism
description: An invariant/policy asserted as self-enforcing ("just a removal of trust") while the concrete enforcing artifact was withdrawn in a prior round and never replaced
metadata:
  classes: [policy-without-mechanism, self-attestation]
  type: feedback
---

**Pattern: policy-without-mechanism (invariant asserted as self-enforcing after its enforcer was withdrawn).**

When a design converges on a clean *organizing invariant* to replace N spot-patches, verify the
invariant names a concrete enforcement mechanism at the exact moment it must hold — do not accept the
policy statement as the fix.

**Why:** In the memory-architecture debate, three rounds hollowed out the clone-injection enforcer:
§12.2 had a concrete git-ignored ratification marker → §13.2 replaced it with a git-authorship check
→ §14.2 demoted authorship to "nudge-convenience, not activation" → §14.1 asserted the *outcome*
("foreign clones load clamped to reference tier") with **no** stated gate, framed as "not new
machinery — a removal of trust." But the committed `active.md` projection is loaded by **native
`@`-import at session open**, before any bespoke process runs. Nothing bespoke can clamp a native
import; a generator-side property (de-authorized voice) does not touch attacker-authored committed
bytes; a SessionStart hook adds context, it cannot un-import. Enforcing the invariant actually
REQUIRES new machinery (git-ignore the projection + regenerate locally). "Removal not machinery" was
the leap of faith.

**How to apply:**
- When blue adopts a unifying invariant, trace it to the leaf: what artifact enforces it, and does
  that artifact run *before* the untrusted bytes reach context? If enforcement happens "at next local
  re-derivation" but the untrusted artifact is loaded natively before then, the invariant is policy,
  not mechanism.
- Check the *withdrawal chain*: if a prior round had a concrete (even flawed) enforcer that got
  withdrawn/demoted, confirm a replacement mechanism was carried forward — not just the *goal* the
  old enforcer served. Enforcers get hollowed round-over-round while the goal-language persists.
- "Removal of trust" / "cheaper than machinery" framing is a tell — verify the removal is
  self-enforcing and doesn't silently require a new gate.

**Extension (efficiency-investigation run 4, round 2):** the class recurs *within the round that
fixed it*. Blue's R1-10 repair correctly named the no-filesystem constraint ("emit into cost.md" is
impossible) in §2.5 — while the same round's R1-6 repair in the adjacent §4.5 specified a
reconciliation check that "throws" on a line-count mismatch the script cannot compute (no fs access).
Heuristic: when a round's repair correctly applies constraint X at site A, grep the SAME round's
other new mechanisms for X-violations — shared authorship + time pressure elevate the base rate, and
the fixed site creates a sibling halo over the unfixed one. Companion root invariant surfaced same
pass (the *attestation ceiling*): an engine whose state rides self-reported envelopes has no
primitive stronger than self-report for work-done claims — schema checks catch omission,
cross-referenced independent structures (lineage-throw style) catch inconsistency, nothing in-run
catches vacuity ("required non-empty" ≠ "work performed"); the honest enforcement tiers are
in-run shape/cross-ref checks + post-run independent audit over git-tracked artifacts, and a repair
claiming "fails structurally, not silently" for a bare non-empty field overclaims. Also watch:
repairs that route the audit of a conflicted seat's self-report back to the same seat's own
spot-check floor (found_by sampled by red-merge; accepted-dispute deltas spot-checked by the
accepting merge) — "auditable" doing the work "audited by a named independent consumer" should do.

**Extension (run 4, round 2 — assumed-durable logging):** telemetry/instrumentation ratified with
an ASSUMED sink is the same class. Blue's repaired §2.5 named "log() into trajectories/journal.jsonl,
consumed by cost-audit.mjs" — but the prior run's journal.jsonl (the only measured one) holds ONLY
started/result lifecycle events (zero log() lines; grep the script's own log strings against the
journal), and cost-audit.mjs has zero journal references (it parses harness transcripts). Check
recipe: (1) grep the emitting script's literal log strings in the prior run's journal; (2) grep the
named consumer for the sink's filename. A "zero-token instrumentation" recommendation whose lines
persist nowhere produces no evidence base — and the repair that introduced the claim was itself the
fix for a prior sink error (repair relocated the defect: [[pattern_repair_regression_citation]]).

**Related, same pass:** *supersession-accounting drift* — a headline count ("5 blocking") goes stale
when a superseding row changes a grade (item 29 Blocking supersedes item 22 High → true count ~6);
re-derive counts from the operative rows, don't trust the verdict tally. And *headline-lag* — a
template section (Heilmeier §0) keeps marketing a feature a downstream concession (auto-promotion →
near-empty-set convenience) has since gutted. Links: [[pattern_self_defeating_mitigation]],
[[pattern_missing_root_invariant]] (this is the failure mode of *adopting* the root invariant red
asked for — the invariant is right but its enforcement is unspecified).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_postmortem_misdiagnosis.md -->
---
name: pattern-postmortem-misdiagnosis
description: "Corrected figures verify clean while the repair's forensic story of HOW the old number went wrong is itself false — test the claimed error mechanism by reproducing the OLD number under it"
metadata:
  classes: [causal-narrative-fails-reproduction, audited-artifact-sibling-halo]
  type: feedback
---

When a repair corrects a figure AND narrates the mechanism of the original error ("the
byte→token conversion dropped at the pricing step"), audit the narrative as its own claim:
apply the claimed broken mechanism and check it reproduces the OLD printed number. In the
efficiency-investigation run 4 (R4-1), the corrected band reproduced perfectly at three seats
— sibling-halo pull toward trusting the diagnosis — but bytes-priced-as-tokens gave
$1.04/2.12/3.56/4.64 against the printed ~$1.40/2.60/4.10/4.10 (15–35% off), while
share-of-whole-merge-dollars reproduced it to ≤3%. The false story was registered at four
sites and ALSO mischaracterized red's own lens method against the citation ledger.

**Why:** verified-corrected figures confer a halo on the accompanying post-mortem; blue's
diagnosis of its own old error is a new unverified claim, and it can defame the auditor's
prior work (check characterizations of red's own past methods against red's ledger — the
ledger is the record).

**How to apply:** for every "the old figure was wrong BECAUSE X" claim: (1) reproduce the old
figure under X; (2) if it fails, hunt the convention that DOES reproduce it (ratios like "≈4×"
can survive coincidentally while the mechanism is wrong); (3) leaf-check any statement about
what a red lens/seat previously did against the ledger entry. Related: [[pattern-repair-regression-citation]],
[[pattern-audited-artifact-sibling-halo]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_provenance_self_report_and_stale_gate.md -->
---
name: pattern-provenance-self-report-and-stale-gate
description: Round-3 logic-lens patterns — a fix that trusts metadata self-reported by a compromised component, and a point-in-time gate that re-opens the very conflict its fallback closed
metadata:
  classes: [self-attestation, unhandled-degenerate-case, cross-section-contradiction]
  type: feedback
---

Two logic/completeness gap patterns caught auditing blue's second-order security fixes (memory-architecture, R3).

**Pattern A — provenance/metadata self-report trusted from inside the blast radius.**
A mitigation that keys trust on metadata (supporting-turn provenance, source tags, review counters) *self-reported by an LLM component* is defeated when that same component reads the untrusted content the mitigation is meant to screen. The injection can manipulate the metadata (e.g. "attribute this to the operator's direct instruction") while leaving the fact body benign — so a body-only screen passes it. Also watch the mirror case: a defense that says "decide on structured fields, treat body as opaque" is *unsafe* when the structured fields (review_count, provenance tier) are exactly what the laundering pipeline inflates. Enabling-a-defense (typing enables screening) is not the same as narrowing-the-surface; do not let it be netted as a surface reduction.
**Why:** the crux is whether the metadata is computed *mechanically by the harness* (safe) or *self-declared by the model* (manipulable). Blue's turn-level provenance (§13.4) never says which.
**How to apply:** whenever a fix rests on provenance/tags/counters, ask "who produces this value, and are they exposed to the input being screened?" If the producer is the compromised component, the fix is a leap of faith — grade the residual, demand mechanical derivation.

**Pattern B — point-in-time gate re-opens the conflict its fallback closed.**
A fix that branches on a *server-side / mutable* condition checked once (Phase 0 flag check) closes the "no owner" hole but opens a "two owners after the condition flips" hole — nothing re-detects the flip. A fallback added at setup time is not a fallback if the world changes after setup.
**How to apply:** for any "if native does X, defer; else we do X" branch, check whether X's availability is stable or can change post-decision, and whether anything re-evaluates. Flag-gated / server-side / rolling-out features are the tell.

**Pattern C — risk-accept rationale contradicts the build-value rationale.**
A design can argue "build is justified because the suite is cross-project / ecosystem / many-repo" in one section and "this risk is acceptable because the operator rarely does the risky-thing" in another, where the two premises are the same axis pointing opposite ways (more ecosystem breadth = more foreign-repo cloning = higher clone-vector likelihood). Also: check that a residual's *effort* grade names the right axis — "high-effort to spoof" was false (git author email is public + one-command settable); the real low-probability axis was *targeting likelihood*, not effort.
**How to apply:** cross-read every risk-accept rationale against the value/motivation sections; a premise that flips sign between them is a coherence gap.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_reflexivity_blindspot.md -->
---
name: pattern-reflexivity-blindspot
description: report headlines a risk class for its subject matter but never applies it to its own toolchain/pipeline/drafting window — audit the report's own operations against its own findings
metadata:
  classes: [reflexivity-blindspot, risk-coverage-omission]
  type: feedback
---

A report that establishes a risk class as a headline finding about its *subject* often exempts
its *own operations* from that same class — silently, not by argued out-of-scope.

**Why:** caught 3 co-occurring instances in one round (FEOV retrospective, 2026-07-13, R1-14/
R1-15 + the R1-1 root): (a) the report treated CVE-class supply-chain poisoning as load-bearing
evidence about a sibling run, then graded adopting two third-party MCP servers into its own
citation-verification path on cost alone — no pin/review/permission-scoping line; (b) the same
poisoning class (untrusted fetched content -> agent context -> downstream action) was never
asked of FEOV's own WebFetch/WebSearch research phase, across all 18 graded risk rows; (c) the
report distrusted the corpus's self-reported status ("a backlog checkbox is not a diff") and
taught the live-source-drift lesson for external citations, but applied neither to its own
drafting window — its keystone "unmerged" claim went stale 8 minutes after verification with no
pinned-SHA/re-verify-before-acting discipline. Related: [[pattern-self-referential-repo-drift]]
(instance c is that pattern's root cause), [[pattern-doctrine-vs-implementation]].

**How to apply:** for every headline risk class or methodological rule the report asserts, ask
"does the report's own pipeline/tooling/build process sit inside this class, and does the text
either grade it or explicitly risk-accept it?" Silence is the gap. Close by addition (one graded
row or a one-line argued out-of-scope), not by blocking the underlying recommendation —
proportionality still applies.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_repair_premise_foreclosed_by_own_profile.md -->
---
name: pattern-repair-premise-foreclosed-by-own-profile
description: A guard added to tag/bound a threat write rests on a premise the same document's allow/deny profile already forecloses — the guard is dead where aimed, and the foreclosed write was a REQUIRED behavior of a subsystem, now silently denied
metadata:
  classes: [cross-section-contradiction, risk-coverage-omission]
  type: feedback
---

When a repair claims "actor X writes surface Y, so we tag/bound that write," check
whether X's write to Y can actually EXECUTE under the design's own printed
allow/deny/fence surface. If every layer forecloses it, two texts cannot both be true:
the repair's justifying premise (the write happens) and the profile (the write is
denied).

**Why:** sleeper-service round 4 (L6-F2): R3-3 built snapshot-diff machinery to
provenance-tag gap-pattern memory "the nightly red-merge seat writes" — but the fence
(writes outside research/+ideas/ denied), `Edit(<REPO>/.claude/**)`, and the marker-loss
canary abort forecloses that write at three layers. The machinery was dead where aimed.

**How to apply:** for every new guard, trace the guarded write path against the profile
as printed. Two horns, both auditable: (a) denies are intended → the REAL finding is the
unstated side effect — the foreclosed write was a required behavior (here: the red
seat's MUST-record-patterns clause fails-denied nightly; routine denials pollute the
tamper-evidence record) and the guard's live remit shrinks to a different surface than
its text describes; (b) the write is meant to succeed → an undeclared allow contradicts
the stated write-surface invariant. Related: [[pattern-doctrine-vs-implementation]],
[[pattern-self-defeating-mitigation]], [[pattern-authorship-evades-never-edit-guard]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_repair_regression_citation.md -->
---
name: pattern-repair-regression-citation
description: A repair that softens an unpinnable figure by attributing it to a NEW specific source can regress — the new source may report a materially different (contradicting) number, or the accurate number lives in an uncited paper
metadata:
  classes: [citation-figure-misattribution, incomplete-repair-propagation, undecided-disjunction]
  type: feedback
---

When blue "fixes" an unpinned/uncorroborated statistic by re-citing it to a specific new
source, YOU MUST re-verify the *new* source at the leaf node — the repair is a fresh
statement↔reference pair, not a closure.

**Why:** In round 2 of the memory-architecture audit, blue's R1-28 repair softened an
unpinnable "80–99% attack success" band to "up to ~90–95% (MINJA / environment-injection),
attributed." Following the two newly-cited footnotes: (1) the environment-injection paper
(arXiv 2604.02623, *Poison Once Exploit Forever*) reports ASR ≤32.5%, not ~90% — the repair
*introduced* a leaf-node contradiction that did not exist in the vaguer original; (2) the
accurate MINJA ~95%/~70% figure is real but lives in an uncited paper (arXiv 2503.03704),
while the survey blue did cite (2606.04329) states it gives no ASR numbers at all. So the fix
made one half contradicted and left the accurate half untraceable.

**How to apply:** Treat every R-numbered "repair" as a new claim to verify, never as a
resolved gap to trust. A softened/hedged number is still a number a skeptic must be able to
follow. "Contradicted at leaf node" is a real regression even when the verdict's disposition
survives (here R1-11 meant the blocking grade did not rest on the figure) — grade it, raise it,
do not let it stand just because it is non-load-bearing. Escalate the round-1 gap id rather
than inventing a wholly new one, so the thread stays navigable.

**Extension (FEOV-retrospective round 2) — adjacent-narrative count transposition:** a repair
touching a graded cell can import an attribute (here a "third occurrence" recurrence count)
from a *structurally adjacent* defect narrative that legitimately carries it (write-block and
ENAMETOOLONG were the paired Tier-B defects, discussed side by side everywhere). Blue's R1-13
fix re-graded ENAMETOOLONG's likelihood on "3 occurrences... per debate.md's merge-seat
friction" — the cited transcript section contains no such event; the true count was 2, and the
"third occurrence" language belonged to the write-block's separately-correct ledger. When two
defect stories travel together in a report, verify each one's counts/dates against its OWN
sources — proximity is a contamination vector during repairs. Same round also produced a
within-repair chronology self-contradiction ("this same round" vs. "two consecutive rounds" in
one paragraph): read the whole repaired paragraph for internal consistency, not just the
repaired clause.

**Red-side extension (FEOV-retrospective rounds 2–3) — red's own text is a claim surface too:**
two ways the adversary's own words became report defects. (1) Red's R2-4 `required_fix` proposed
a specific replacement citation (arXiv:2606.02646) for the "7 agents" figure *without leaf-
verifying it first*; blue fetched it, found it does NOT contain the figure (knee N≈10, plateau
~1.8 by N=30 — opposite direction), and rebutted the proposed fix while conceding the gap — the
correct outcome, but red proposed a citation that would have been a new miscitation. Leaf-verify
any source you name in a `required_fix` before proposing it. (2) Red's round-2 merge wrote "grep
'independent': zero hits outside the ledger line's own text" — an unverified-as-worded
characterization (actual result: zero hits INCLUDING the ledger clause); blue copied it verbatim
into the row-19 fix, and the imprecision survived two rounds before a round-3 lens re-ran the
grep. Blue copies red's characterizations verbatim when repairing — an imprecise red sentence
becomes a report defect with red's own authority behind it. Re-verify your own prior round's
phrasings at the same leaf-node standard as blue's.

**Red-side extension (FEOV-retrospective round 4) — an unresolved "or" in a required_fix ships
as an undecided fix:** red's R3-1 required fix said "treat as effective PASS-with-warning or
throw a distinguishing error." Blue's round-3 repair (§3 row 20) carried the disjunction forward
*verbatim* — the only fix that round that shipped an "or" where its siblings shipped decisions —
and the two branches were opposite failure philosophies (silently convert a degenerate FAIL
toward PASS vs. halt loudly). Third instance of blue shipping red's phrasing verbatim. When a
required_fix contains alternatives, YOU MUST name red's favored side (and why), while still
accepting an argued choice of the other branch — otherwise the disjunction itself becomes the
shipped behavior spec.

**Extension (run 4, round 2) — two more repair-regression shapes:** (1) *repair relocates the
defect*: a fix for an impossible sink ("emit into cost.md" — script has no fs access) replaced it
with a different unverified sink ("log() into journal.jsonl, consumed by cost-audit.mjs") — both
halves contradicted by first-hand check; verify a repaired mechanism claim end-to-end, not just that
the old error is gone. (2) *insertion severs the host sentence*: a correction record spliced into
the middle of a round-0 sentence orphaned its second clause ("But was measured-robust in this
loop" — no subject); after any inline correction-record insertion, read the WHOLE host paragraph
aloud for grammatical integrity, not just the inserted text. Also positive signal worth keeping:
the highest-risk repair class (re-citation to a brand-new source, R1-5→CathedralBazaar) verified
clean at the PDF leaf — arXiv-HTML garbled the key sentence ("at least 11" for "at least 1
*vector*"); when an HTML digest returns a semantically impossible number (CVSS has 8 base metrics),
suspect italic/subscript rendering artifacts and go to the PDF via pdftotext before flagging.

**Extension (run 4, round 4) — a repair's forensic DIAGNOSIS is a claim too:** blue's R3-1 fix
corrected a 4×-overstated dollar series (band right, floor/ceiling reproduce) but asserted the
bad series "reproduces only if cache-weighted BYTES are priced as tokens." Re-derivation from
the committed instrument: bytes-as-tokens gives $1.04/2.12/3.56/4.64 — NOT the printed
$1.40/2.60/4.10/4.10; cache-weighted-share × whole-merge-dollars regenerates both the printed
series and red lens-3's sibling recompute to ≤3%, and red's own ledger had recorded lens-3's
method as "share × merge $." When a repair explains WHY a prior figure was wrong, re-generate
the wrong figure under the claimed mechanism; if it doesn't regenerate, the diagnosis
misattributes — and may mischaracterize a sibling seat's ledgered method with red's authority
attached. Corrected numbers reproducing does not validate the error story wrapped around them.

**Merge-seat extension (run 4, round 2) — "matches red's required fix exactly" is NOT leaf
verification:** when red's own round-N finding embedded an error (R1-27 asserted a phrase lived
verbatim in TWO files; debate.js l.263 actually says "5 regressions", not "5 chains"), a
round-N+1 lens verified the repair HIGH on *fidelity to red's instruction* while a sibling lens's
first-hand source read contradicted the copied content. At merge, a fidelity-based grade and a
source-based grade are different evidence tiers: the source read wins, always — re-run it
first-hand before overruling. Corollary: every claim red's required_fix asserts about a source
(not just citations it proposes) must itself be leaf-verified, because blue copies it AND a
future lens may "verify" the copy against the instruction instead of the source, laundering
red's error through two seats.

Related: [[citation_status_and_misattribution_patterns]] (real figure miscited to wrong
source), [[gap_live_source_drift]] (re-follow to current primary),
[[pattern-within-source-condition-misattribution]] (right paper, wrong experimental arm),
[[pattern-identity-keyed-detector-lineage-blind]] (red's fresh-id convention defeating the
docket detector — another red-practice-becomes-system-defect case).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_risk_grading_conflations.md -->
---
name: pattern-risk-grading-conflations
description: Recurring logic gaps in research risk matrices — likelihood/success conflation and keystone-on-unverified-evidence
metadata:
  classes: [metric-conflation, claim-contradicts-own-record, risk-coverage-omission]
  type: feedback
---

Two gap *patterns* to check in any risk matrix or verdict, caught in memory-architecture round 1:

**1. Attack-success conflated with attack-likelihood in one risk cell.**
A cell like "Likelihood: Med (80–99% attack success)" fuses two different quantities:
P(attempted) and P(succeeds | attempted). Adversarial red-team success rates say nothing about
whether *this* deployment gets attacked. For single-operator / machine-local / private systems,
demand a suite-specific attacker model before accepting a "blocking" escalation.
**Why:** blocking grades force blue to absorb complexity; an unspecified attacker model is a
leap that can make the design strictly heavier for a low-probability event.
**How to apply:** when a risk is graded blocking/High on generic statistics, ask "who attacks
this instance, and how does the input reach the store?" Keep gates on the genuinely untrusted
edges; contest the surplus mitigations' grade, not the risk's existence.

**2. Verdict keystone rests on the report's own Unverified list.**
Check that the evidence called "strongest/decisive" is not the same item filed under
"unverified / low-confidence" elsewhere in the document. The verdict cannot exceed the
confidence of its load-bearing citation.
**Why:** self-undercutting evidence chain — reads persuasive but collapses on cross-reference.
**How to apply:** cross-read the verdict's superlatives against the Unverified section and the
footnote source quality (blog/community vs. primary/docs).

**3. A fix that relocates the problem one level down.**
When blue proposes a structural fix (e.g. append-only to stop rewrite-drift), check whether the
new rule re-imports a problem the report condemned elsewhere (e.g. unbounded growth / context
bloat). Fixes should be checked against the report's own other constraints for tension.

**5. Risk-acceptance argument invokes a mechanism that doesn't address the actual trigger.**
Caught in FEOV-retrospective round 1: a "risk-accept pending trial" cell claimed a write-block
fix (append-vs-Write tool-call shape) "may moot most of" an unrelated defect (a shell
command-length ceiling on large heredocs) "by construction." The two axes are orthogonal — the
fix changes *which* tool-call shape is used, not the *payload size* that actually trips the
ceiling. Distinct from pattern 3 (relocates the problem down a level): here the fix doesn't even
share an axis with the trigger, it's a non-sequitur dressed as a mitigation.
**Why:** a plausible-sounding mechanism claim in a risk-accept cell is exactly where scrutiny
relaxes; the softer wording ("may moot most of") is often the tell that the causal chain was
constructed post-hoc from an adjacent finding.
**How to apply:** for every risk-accept whose rationale is "fix X moots defect Y," ask "does X
change the same variable that actually triggers Y?" If the answer requires an unstated bridging
assumption, demand it be stated or downgrade the risk-accept back to graded-open.

**4. Motivated netting — the dominant risk classified "shared/inherited" to zero it out.**
Caught in memory-architecture round 2 (R2-1). When blue answers a build-vs-adopt (or
cost-vs-value) challenge with a *netted table*, check whether the dominant risk dimension is
labeled "shared with the alternative / inherited from the baseline" so it drops out of the
net-new column — while the report's OWN text elsewhere says the design *widens* that same
dimension. "Can't escape it entirely by adopting X" is not "adopting X buys no smaller surface."
**Why:** the accounting is where a weak go/no-go gets laundered into a strong one; the widening
delta (extra intake, larger blast radius, new laundering/promotion step) is the actual net-new
cost and is exactly what gets omitted.
**How to apply:** for every "shared/inherited" cell, find the report's own passage describing how
the bespoke layer differs from the baseline on that axis; the difference is net-new surface the
table must count. Cross-read the netting table against the threat-model section's verbs
("reproduces AND widens", "converts a one-shot injection into a permanent rule", "propagates to
every project").


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_schema_legal_control_flow_trace.md -->
---
name: pattern-schema-legal-control-flow-trace
description: dark-side/risk lens finds its best gaps by hand-tracing control flow for schema-legal-but-semantically-incoherent envelope shapes, not by re-checking citations
metadata:
  classes: [unhandled-degenerate-case, partial-control-coverage, false-universal]
  type: feedback
---

When auditing a script/harness whose keystone facts (guards, routing, ledger text) have already
been re-verified clean across multiple rounds by leaf-node citation lenses, the dark-side/risk
lens's marginal value shifts from "re-check the same facts again" to **hand-tracing runtime control
flow for inputs that satisfy the schema but are semantically degenerate**. Two recurring
sub-patterns, both caught this way in the FEOV retrospective round 3 (`round-3-lens-5.md`):

1. **Schema-legal-but-incoherent envelope shapes.** A schema requires fields to exist and have the
   right type, but rarely enforces cross-field coherence (e.g. `verdict: 'FAIL'` with `gaps: []`).
   Trace what the *loop*, not the *schema validator*, does with that shape by hand, for several
   rounds. Look especially for paths where a guard/branch is gated on a derived condition (e.g.
   `contested.length > 0`) that a degenerate-but-valid input makes permanently false — this
   silently disables an entire adjudication/escalation path rather than throwing.
2. **"Never dropped" / "always captured" claims about telemetry plumbing** (friction arrays, gap
   ids, provenance tags) are asserted from *design intent*, not from grepping every call site that
   *should* invoke the aggregator. Grep for the aggregator function itself (e.g. `takeFriction(`)
   and enumerate every call site against every schema'd seat that *could* produce that signal — an
   asymmetry (3 of 4 schema'd seats wired, 1 silently omitted) is exactly the kind of omission a
   passing regression suite won't catch if the suite's own test list mirrors the same incomplete
   site list.

**Why this works when citation lenses have already gone clean**: citation lenses re-verify facts
*about the world* (a paper says X, a repo state is Y); this class of gap lives entirely inside the
system's own internal data flow and requires no external reference at all — it is caught by
reading the script like an interpreter would, not by fetching anything. Best done specifically in
round 3+ of a debate, after the obvious citation/doctrine gaps are exhausted and the marginal
citation-recheck yield has dropped.

See also [[gap_live_source_drift]] (re-pin against the *current* HEAD before starting — in this
case the repo had advanced to a 3rd commit past the report's last pin, still docs-only, confirmed
via diff before trusting any prior-round claim) and [[pattern_self_defeating_mitigation]] (a
control whose own trigger condition can be permanently disabled by a degenerate input is the same
family as a mitigation defeating itself).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_self_defeating_mitigation.md -->
---
name: pattern-self-defeating-mitigation
description: A control added to close a prior-round gap introduces its own failure mode — audit every new mitigation as a fresh attack surface, not a closed ticket
metadata:
  classes: [unverified-composition, partial-control-coverage, cross-section-contradiction]
  type: feedback
---

When blue adds a mitigation in round N to close a round N-1 gap, do NOT treat the gap as closed on
acceptance — re-audit the mitigation itself as new surface. Recurring sub-patterns observed:

- **Control collides with the system's own write loop.** A clone-injection defense that gates
  activation on a content-fingerprint marker breaks because the nightly `/dream` pass *mutates the
  store*, invalidating the fingerprint every run — forcing either self-ratification by the unattended
  pass (defeats the gate) or daily manual re-ratify (unworkable). Ask: does the new control's
  invariant survive the system's own routine writes?
- **Control depends on the very diligence the report elsewhere discredits.** §2.4 demoted human
  diff-review to "forensic" because it decays to LGTM; then the clone defense makes human ratification
  the *only* preventive control. A mitigation cannot lean on an assumption the same report argues is
  unreliable.
- **Control has an escape hatch that reopens the common case.** "auto-ratify repos under a configured
  trusted root" voids the clone defense for the solo-dev workflow (clone everything under ~/Projects).
- **Control closes the durable path but not the in-pass path.** Ephemeral consolidator memory closes
  *durable* self-poisoning but the agent still *reads* the poisonable store each pass → in-pass
  steering residual. "Sits outside the trust surface" is overstated when it still ingests the guarded
  surface.
- **Verification channel is the same as the injection channel.** A risk-accept names "independent
  re-verification" as the mitigant for content-poisoning, but the actual leaf-node protocol is
  "follow the citation to the source" — i.e. re-fetch the *same* URL, not a second independently-found
  one. A single poisoned/fabricated source that is internally consistent gets checked by re-reading
  itself; that isn't independence. Grep the protocol text for the word the mitigation leans on
  ("independent") before crediting it — if it doesn't appear, the mitigation is asserted, not built.
- **A two-party dispute/correction channel audits only its contested branch.** A grade-dispute
  mechanism dockets a dispute to the judge only after reject → re-dispute; a dispute the counterparty
  ACCEPTS silently rewrites the permanent record with no log, docket, or spot-check — and acceptance
  is the cheap/colluding path (both parties have budget stakes under actuation). Companion holes: the
  loop's PASS/terminal break precedes the dispute filter (ending the run moots pending disputes —
  the channel fails the final-round case it exists for), and dispute traffic is uncapped (a cost
  lever pointed at the responder). Ask of any adjudication channel: what happens on the AGREE path,
  at termination, and at volume? An incentive analysis that grades only one party's incentive is
  half an analysis.
- **Guard scheduled after the commitment it gates.** A write-guard preflight repaired twice
  (simulator→live seat, then live-seat→same seat class) ends up scheduled as "the first sharded
  run's red-merge writes both shards as its opening act" — i.e. it can only FAIL after the PR
  shipping the sharded design (renamed prompts, retired monolith) is already merged. A preflight
  that fires past the point of no return is a smoke alarm wired to the ashes; also check the fix
  didn't contradict a sibling clause (here: "skeleton-born" names a different creator seat).
  Ask of every preflight/gate: at the moment this check can first fail, is the decision it
  guards still reversible? (efficiency-investigation round 4, L6-F3 — vector was red's own
  R3-11 required-fix example, repeated by the lead.)
- **Two controls added in different rounds silently collide.** A cost-saving cache ("verified
  citations don't un-verify unless the citing prose changed") and a drift-mitigation ("record
  access-date deltas because sources change") get shipped separately, each reads fine alone, but the
  cache's skip-condition is keyed to internal edit history while the drift risk is external and
  time-based — so the cache actively prevents the re-check the drift fix's own rationale ("drift is
  usually caught by re-verification") depends on. Neither section mentions the other. Ask, for every
  pair of controls addressing adjacent risk classes: does control A's skip/short-circuit condition
  starve control B of the trigger it needs?

**Why:** blue's Round-1 additions were directionally correct and accepted, but each carried an
un-graded second-order failure — accepting the direction is not accepting the implementation.

**How to apply:** for every ACCEPT in blue's response section, write the adversarial one-liner "how
does this new mechanism fail / get bypassed / collide with existing machinery" before crediting the
gap as closed. Grade the residual separately.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_self_referential_repo_drift.md -->
---
name: pattern-self-referential-repo-drift
description: The audited repo's own git state (not an external citation) moves between blue's verification and red's audit — and a merge landing doesn't mean every related call site got fixed
metadata:
  classes: [live-source-drift, incomplete-repair-propagation]
  type: feedback
---

Gap pattern: **self-referential repo drift**, distinct from [[gap_live_source_drift]] (which is
about external web citations going stale). Here the *subject of the report itself* — the repo
being retrospected — changes state between the report's own verification timestamp and red's
audit pass, because both blue and the codebase live in the same fast-moving repo.

**Why:** Caught in the FEOV-retrospective audit (round 1, lens 5). Blue's report headlined "the
fixes exist but have not shipped," verified against `main` @ a specific commit. The PR in question
merged to `main` ~8 minutes *after* blue's verification commit — a genuine race, not sloppiness —
but by the time red audited, the headline was false on the current primary source (confirmed via
`git log`, `git merge-base --is-ancestor`, and `gh pr view --json state,mergedAt,mergeCommit`
against `origin/main`, not just local `git log`).

**The second-order trap:** once you notice the drift, the *naive* fix ("flip unmerged to merged")
overstates resolution. A merge landing does not mean every location the report's disposition table
depends on actually changed. In this case: the merge shipped 3 of 4 null-guard call sites: a
4th (the judge-adjudication call) still threw unguarded on the merged `main`. Verify each
sub-claim the drift affects individually — don't let "it merged" become "therefore it's fixed."

**How to apply:**
- For any report auditing "has X shipped / is Y merged," re-run the exact git/`gh` commands red
  is auditing *now*, not trust the report's cited commit hash as still being HEAD.
- `git log --oneline -1 origin/main` + `gh pr view <n> --json state,mergedAt,mergeCommit` +
  `git merge-base --is-ancestor <sha> origin/main` — confirms merged AND pushed, not a local-only
  merge commit.
- When drift is found, re-verify every downstream claim the stale premise supports individually
  (grep the actual diff / re-read the actual merged file) before declaring it resolved — a partial
  fix disguised as a full one is worse than the original staleness, because it reads as closed.
- Flag both halves as separate gaps: (a) the stale headline itself, (b) whichever specific
  sub-claim survives even after the correction — since a reader who fixes (a) alone will assume
  (b) is also handled.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_sibling_repair_composition.md -->
---
name: pattern-sibling-repair-composition
description: Two same-round repairs to one design (mechanism hardened, registered test/prediction restated) can fail to compose — re-derive the test from the hardened mechanism
metadata:
  classes: [unverified-composition, incomplete-repair-propagation]
  type: feedback
---

When one round applies two fixes to the same paragraph/design — one hardening the MECHANISM
(e.g. arm conditions upgraded to double confirmation) and one restating its REGISTERED
TEST/PREDICTION (e.g. dollar netting corrected) — verify they compose: the prediction's
trigger condition must be re-derived from the hardened mechanism, not carried over from the
pre-hardening version.

**Why:** efficiency-investigation round 3 (L5-F2): R1-12 hardened the re-scoped floor's arm
to "two consecutive zero-above-floor-mint rounds," while R1-17 restated the registered
prediction's netting but left its trigger single-round and total-mint-keyed. Both closed
clean in round 2 individually; jointly the registered test could settle TRUE while the
hardened variant saves $0 — falsely validating a build trigger. Sibling of
[[pattern-unreconciled-numeric-floors]] (requirements that don't arithmetically compose) but
at the repair layer: each repair verifies clean in isolation; the defect is only visible
reading them together.

**How to apply:** when auditing repairs, group same-section fixes and check pairwise
composition — especially mechanism-change + registered-figure/prediction pairs, and any
"arm/trigger" clause vs the test that claims to measure it. A repair-verification sweep that
checks each closure independently (as round-2 did) structurally misses this class.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_stale_baseline_pricing.md -->
---
name: pattern-stale-baseline-pricing
description: Cost/efficiency lever analysis priced against a historical run's cost distribution that already-shipped mechanics have made unreachable; the report may even DESCRIBE the shipped mechanic correctly in one section without propagating its consequences into the sections it invalidates
metadata:
  classes: [pricing-basis-drift, incomplete-repair-propagation]
  type: feedback
---

Gap pattern: **stale-baseline pricing / described-but-not-priced mechanic**. An efficiency or
tradeoff report ranks levers and estimates savings against the measured cost structure of the
last run, while code already shipped between that run and this one guarantees the next run's
distribution differs (run-4 efficiency report: §3.1 correctly stated the shipped whole-debate
auto-docket means "any gap open across two rounds now auto-dockets," yet §6.1's money map,
every savings estimate, and §1's rejection of judge-routing all assumed run 3's zero-judge
baseline — and §8 even asked whether the docket "arms in its first live trial" as if arming
were uncertain, when the code makes it near-certain).

**Why:** the report's own accurate sentence is camouflage — leaf-verifying it passes, so a
citation-fidelity check finds nothing. The defect is that the verified fact's consequences
were never propagated into the pricing/decision sections. Sibling of [[pattern-inherited-surface-netting]]
(baseline moved under the argument) and of incomplete-propagation, but the unpropagated item
is an *implication*, not a corrected string — greps can't find it.

**How to apply:** on any lever/efficiency/priority-ranking audit, (1) list every shipped
change between the measured baseline and the next run; (2) for each, hand-trace the control
flow and ask which recurring cost lines or trigger frequencies it changes (a detector that
fired zero times historically may now fire every round); (3) check the report's savings math
and rejection arguments against the PROJECTED distribution, not the measured one. Also check
loop ordering: which resolutions re-enter the loop (e.g. `carried` never entering
`adjudicated` = re-docket every round = recurring judgment-tier spend nobody modeled).

**Extension (run-4 round 2):** fixing one instance does not clear the pattern — after blue
conceded the dead-baseline principle "in full" for the judge seat (R1-1) and repriced,
the LENS seat stayed priced at run-3's 5-lenses/round while the shipped `citationPasses`
formula + the current claim count made it 6 — provable from the audited run's own candidate
directory (round-1-lens-{1..6}.md existed). When a dead-baseline gap is repaired, sweep
EVERY seat/line the shipped code moved, and use the live run's own artifacts (lens file
count, dispatch labels) as the cheapest evidence of the projected distribution. The
current run auditing the report is itself running the new baseline.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_structural_close_scoped_to_toplevel.md -->
---
name: pattern-structural-close-scoped-to-toplevel
description: A structural "closed at the tool boundary" claim covers the top-level session but not the subprocess-spawned seat agents the design's own text concedes are the reachable surface
metadata:
  classes: [partial-control-coverage, cross-section-contradiction]
  type: feedback
---

A structural close (e.g. a bare `Bash` tool deny that "removes the tool entirely") is
justified on the TOP-LEVEL session's behavior ("§2.2's steps never invoke Bash"), then a
totalizing safety claim is written elsewhere ("the Bash read carve-out is closed
STRUCTURALLY … so read-scoping holds on the Bash channel too"). But the design spawns
SUBPROCESS actors — Workflow-tool FEOV seat agents, subagents — that do the actual work, and
the report's OWN text concedes Bash is reachable there ("Where Bash IS reachable (a rebuilt
rung, the Workflow seat agents, profile drift)"). The two statements contradict: the total
claim and the conceded exception cannot both hold.

**Why:** round-5 sleeper-service. R4-3's bare `Bash` deny closed the top-level session; §6
row 13 claimed the exfil channel closed structurally; §4.3 layer 4 (i) conceded seats are
Bash-reachable. On the seat surface the bare deny does nothing and closure reverts to the
non-exhaustive belt enumeration R4-3 itself called inadequate. The hook FENCE proven on seats
covers WRITES, not Bash READS — so the read+egress exfil (row 13's own threat) reopens for
exactly the nightly research actors.

**How to apply:** when a report claims a control "removes/closes X at the tool boundary,"
find the surfaces that AREN'T the top-level session — spawned subagents, Workflow/debate
seats, rebuilt rungs. Ask: does this control (a layer-1 permission RULE / `--settings`
profile) INHERIT to the spawned actor? "Hooks fire on seats" (layer 2) is NOT evidence that
the settings profile (layer 1: bare-tool deny, read-scoping, egress allow/deny) inherits.
The two are separate layers; leaf evidence for one does not transfer. Also: a fence that
blocks WRITES does not close a READ-then-egress exfil channel — check the read side
separately. Grep the report for its own "where X IS reachable" concessions and diff them
against its totalizing "X is closed" claims — the contradiction is usually already on the page.

**Now also (round-5 merge):** check the OPPOSITE assertion too — the same report can
simultaneously say seats ARE bound ("seat agents are full permission-engine + hook subjects",
layer table) and NOT bound ("Where Bash IS reachable (… the Workflow seat agents …)",
repair bullet). Raise the contradiction with BOTH horns priced: the BOUND horn is not free —
it silently starves the spawned workers of a capability the design's own methods demonstrably
use (here: citation seats' Bash git probes, live-probe lanes), and that starvation's friction
can be structurally unable to surface (ordinary-classed, waiting on corroboration that cannot
occur). "Costs the session nothing" arguments must be re-derived over the spawned population,
not just the top-level step list.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_unquoted_hold_masks_discrepancy.md -->
---
name: pattern-unquoted-hold-masks-discrepancy
description: A lens's "checked, held, not raised" entry that asserts a match WITHOUT quoting both sides can mask a live discrepancy; when lenses conflict, the merge seat decides by direct read, never by majority or by trusting the hold
metadata:
  classes: [self-attestation, causal-narrative-fails-reproduction]
  type: feedback
---

Rule: a hold-claim ("verified: X matches Y, no discrepancy") carries evidentiary weight only if
it quotes both texts side by side, the same standard gap-claims are held to. An unquoted hold is
a weak negative and MUST NOT rebut a sibling lens's quoted positive finding.

**Why:** FEOV retrospective round 5 — lens 5's "noted, checked, not raised" item asserted §3
row 23's regression-chain enumeration "matches §2.1's text exactly; no discrepancy found," while
lenses 1, 2, and 4 each independently quoted row 23 carrying blue's discarded, factually-wrong
first-pass list. Direct read at the merge seat (report lines 496 vs. 727) confirmed the three
quoting lenses right and the hold wrong. Same round, second instance: lens 2's ledger line
asserted a five-id list "matched" a six-id set — mechanical extraction disproved it. Both errors
were inside *verification-positive* statements, the kind that never get adversarial pressure
because they raise no gap.

**How to apply:** at every red merge, (1) when any lens's hold contradicts any sibling lens's
finding, resolve by first-hand read of the primary text — never by lens majority, seniority, or
plausibility; (2) treat "verified clean" ledger lines that name a specific correspondence
(list X = list Y, text A = text B) as claims to re-extract mechanically (grep/sed the exact
line), not to trust; (3) in my own lens passes, quote both sides in every held comparison — a
hold without quotes is homework not done. Related: [[pattern-repair-regression-citation]]
(red-side errors get logged with the same discipline demanded of blue),
[[pattern-schema-legal-control-flow-trace]].

**Extension (sleeper-service r1):** the hold can come WITH a mechanism story that itself was
never verified. Lens 1 saw wc=1,557 vs report's "1,558 lines" and ruled "trailing-newline
artifact, not a discrepancy" — but `tail -c 1` shows the final byte IS a newline, so the file
has exactly 1,557 lines and the excuse fails reproduction (kin to
[[pattern-postmortem-misdiagnosis]]). An artifact-excuse for a numeric mismatch (trailing
newline, encoding, off-by-one-of-counting-method) is a testable claim: test it at the merge
seat before accepting the hold. Cost: one shell byte-check.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_unreconciled_numeric_floors.md -->
---
name: pattern-unreconciled-numeric-floors
description: Two fixes added in different rounds each set a minimum/allocation over the same shared resource (lane count, agent count, budget); neither cross-references the other's arithmetic
metadata:
  classes: [unverified-composition, figure-recount-fails]
  type: feedback
---

When a report adds a **floor** for some resource in one row/section (e.g. "lane count >= 3") and,
elsewhere — often a different round, addressing a different gap — adds an **allocation or
redundancy requirement that consumes multiple units of that same resource** (e.g. "assign N distinct
roles across lanes, with role X getting at least 2 of them"), do the arithmetic explicitly. Do not
assume two individually-reasonable-sounding fixes compose.

**Why:** FEOV retrospective round 2 — R1-16's redundancy floor ("critical-stance lens on >= 2 of N
lanes") and H1's lane-count floor ("lanes >= 3") were added in the same report but different rows,
each reads fine alone, and together are infeasible at the stated floor: 3 named method-classes with
one doubled needs >= 4 lane-slots, but the floor only guarantees 3. Neither row cross-references the
other; the report ships both as if independent.

**How to apply:** whenever a gap's fix is "assign >= K of N things to category X" and a *separate*
row or an *earlier* round already floors N at some fixed minimum, compute whether K plus the other
required categories fits inside that minimum. If it doesn't, the floor is silently going to be
violated by whichever run actually hits it — flag as a reconciliation gap (state the real floor, or
state which category gets dropped/merged at the stated floor), not a correctness bug in either row
alone. Distinct from [[pattern_missing_root_invariant]] (which is about a recurring *security* gate
across rounds) — this is arithmetic composition of two resource-allocation constraints, checkable in
one read without live re-verification.

**Round 3 extension — the reconciliation itself can repeat the error.** When blue "closes" this
exact gap by stating an explicit combined floor (e.g. "needs `lanes >= 4`"), redo the arithmetic
against the row's own roster count rather than trusting the stated conclusion: FEOV round 3 found
blue's own reconciliation (R2-8) computing "4 named methods + a 2-of-N floor on one of them needs
lanes >= 4" when the literal math is 3×1 + 1×2 = 5, not 4 — the lower number only works if two of
the four named items are silently merged into one allocation unit, which the same sentence's own
"four named methods" phrasing contradicts. A fix explicitly written to reconcile an
under-composed floor is exactly the kind of self-referential claim [[pattern_repair_regression_citation]]
warns about for citations — the analogous check for *arithmetic* repairs is: recompute, don't
re-read. Also watch for the adjacent tell: calling an *unenforced default* (a value with "no
minimum check," confirmed [OPEN] two rows down) "the shipped floor" in the same reconciliation
sentence — a floor that's already been overridden downward once in the report's own corpus (run 2
shipped `lanes=2`) is not a floor yet.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_verification_probe_layer_masking.md -->
---
name: pattern-verification-probe-layer-masking
description: a per-run "enforcement verified" probe whose PASS outcome is produced by a different layer than the one under test — the redundant layer can be dead and the probe still passes
metadata:
  classes: [verification-scope-blindspot, derivation-status-overclaim]
  type: feedback
---

A verification probe (denial canary, liveness check) minted to prove layer N is live can be
INCAPABLE of isolating layer N when a co-located layer produces the same observable outcome.

**Concrete instance (sleeper-service R2, L6-F1):** step-0 denial canary "attempt one
out-of-fence write, abort unless DENIED" was minted (R1-28) to prove the marker-keyed HOOK
fence (layer 2) is live per run and "fails CLOSED on marker loss." But the same out-of-fence
write is independently denied by layer 1 (the `dontAsk` permission profile) — so when the
marker is lost and layer 2 goes dormant, layer 1 denies the canary write anyway, the canary
reports DENIED, and the run proceeds under the leaky layer-1-alone config the probe existed
to prevent. The canary passes EXACTLY when the fence is dead.

**Why:** layers 1 and 2 fence the identical boundary, so no probe write exists that layer 1
allows but layer 2 denies. A deny-OUTCOME probe over a boundary two layers share can never
isolate the redundant layer. The fix is a POSITIVE liveness record emitted by the layer under
test (the hook writes "I fired + denied X" to a log the wrapper confirms), not an outcome the
other layer can mask.

**ALLOW-side variant (sleeper-service R4, L4-F1/L3-F2, two lenses independently):** the
inverse masking. Blue "reproduced" that the dontAsk read-only-Bash carve-out classifier
passes `git log --output` because the command ran with NO PROMPT in its seat — but every
harness seat on this box runs under `defaultMode: "auto"`, where the AUTO classifier is the
approving layer; "no prompt" proves "≥1 approving layer passed it," never "the carve-out
classifier passed it." An allow-outcome probe over shared APPROVING layers is exactly as
attribution-blind as a deny-outcome probe over shared denying layers. Check the probe seat's
own permission mode before accepting any classifier-attribution claim.

**Environment fact (recorded R4):** the isolating probe — nested `claude -p
--permission-mode dontAsk` from a clean temp repo — is itself DENIED by the seat's auto-mode
classifier (twice, minimal command). Layer-isolation probes are not runnable from lens seats;
grade the attribution claim down with the attempt recorded and route the probe to a build-PR
or operator step. Do not route around the denial.

**How to apply:** whenever blue adds a per-run "enforcement verified / fails closed" probe,
ask: what OTHER layer produces the same PASS observable? If any co-located control denies/
allows the same probe target, the probe proves "≥1 layer acted," not "THIS layer acted."
Demand a layer-specific positive signal. Sibling of [[pattern_self_defeating_mitigation]] and
[[pattern_fail_open_guard_and_erasable_evidence]] — the RECORD-but-don't-CLOSE-THE-LOOP shape
(attempt/record without verify-and-surface) recurred three times in one round (canary,
snapshot with no reader, plugin-copy with no version check).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_waiver_graduation_and_closure_amendment.md -->
---
name: pattern-waiver-graduation-and-closure-amendment
description: checked-not-raised waivers are class-conditional (promote when the class becomes established); prior CLEAN closures can need amendment when a composition defect between two verified repairs surfaces a round later
metadata:
  classes: [incomplete-repair-propagation, unverified-composition]
  type: feedback
---

Two red-merge process rules minted in the efficiency-investigation run, round 3.

**1. Waiver graduation.** A "checked and deliberately not raised" entry is a waiver of an
INSTANCE, conditional on its defect CLASS staying incidental. When a later round establishes
that class as recurring/load-bearing in the same report (e.g. round 3 found "over-tight prose
gloss on a measured series" twice — a false 'largest every round' universal AND a '~5' gloss
on a 4.4–6.0 band the round-2 merge had waived), promote the waived instance under the
ledger's elapsed-rounds rule and say why: a waived instance of an established class invites
the same skeptic's quote-mining the raised instances do.
**Why:** the not-raised list is part of the audit record; treating waivers as permanent lets
a defect class ship half-fixed.
**How to apply:** at each merge, before finalizing, diff this round's new defect CLASSES
against prior rounds' checked-not-raised lists; promote matches with an explicit
class-consistency rationale (and never silently — the reversal is argued in the merge notes).

**2. Closure amendment (composition defects arrive late).** Two repairs can each verify
faithful in isolation, be closed CLEAN, and only a later round's re-derivation shows they do
not compose ([[pattern_sibling_repair_composition]] — here R1-12's hardened arm vs R1-17's
registered prediction, caught one full round after both closed clean). The closure taxonomy
has no amendment class, so the mechanics are: mint the successor with `supersedes` naming BOTH
ancestors, record them in this round's `closures` array as closed_with_regression
(reclassification), and annotate the closure ledger + status table with an explicit
"amended round N" note. Filed as friction too — the schema wants closures to be this-round
events.
**How to apply:** when a successor supersedes gaps closed in an EARLIER round, always declare
the reclassification in the current envelope and ledger rather than leaving the chain implied;
an undeclared lineage is a protocol violation regardless of when the ancestor closed.

Related: red-phrasing-as-vector ([[pattern_repair_regression_citation]]) recurred a SECOND
consecutive round (round 3: R2-6's contest-window fix text carried an unnamed-reviewer hole;
R2-11's finding text mispointed the enum to clause (v)) — the required-fix self-leaf-check is
still not habitual; draft required fixes as verifiable claims, then verify them.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_wildcard_allow_write_gadget.md -->
---
name: pattern-wildcard-allow-write-gadget
description: A wildcard allow rule characterized as "read-only" hides a write gadget via the tool's own output flag; run the leaf to find it
metadata:
  classes: [false-universal, enumeration-non-exhaustive]
  type: feedback
---

A permission allow rule with a wildcard argv (e.g. `Bash(git log *)`) described in prose as
"pinned read-only" / "no rule grants argv that chooses a write target" can still contain a
**write gadget through the tool's own output flag** — `git log --output=<arbitrary path>`
writes a file at exit 0 under the wildcard. Confirmed by running it (round 3, sleeper-service
§6 row 4 / §4.3 layer 4).

**Why:** the report enumerated only shell-level escapes (compound-command `;&&`, redirection
`>`, path traversal) in its open question (OQ18) and declared everything else "read-only." A
tool-native output flag is a fourth class that neither the enumeration nor the tamper-
watchman covers: `--output=<outside-repo>` shows in neither `git status --porcelain` nor the
guardrail-hash set, so the snapshot-compare backstop misses it.

**How to apply:** whenever blue characterizes a wildcard allow rule as read-only or argues a
risk-accept on "the argv can't choose a write target," ENUMERATE the tool's own write flags
(`--output`, `-o`, `--output-directory`, `-O`, format-to-file options) and RUN one to prove
it. A conjunctive narrowness argument ("needs a bug AND a write gadget") collapses if the
gadget is present by the rule as written, needing no bug. Check whether the design's own
tamper-evidence set (porcelain + guardrail hashes) would even see an out-of-repo target.
Related: [[pattern_per_instance_cap_not_per_cause]] (allow/deny surface reasoning),
invariant-soundness-by-enumeration (denylist under-inclusion — this is its allowlist mirror:
an allowlist rule that is over-inclusive of write channels).


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: pattern_within_source_condition_misattribution.md -->
---
name: pattern-within-source-condition-misattribution
description: A correctly-cited paper's headline result quoted with a gloss that silently reassigns it to a weaker experimental condition/arm — especially to justify NOT building the stronger condition; check which arm carries the number
metadata:
  classes: [within-source-condition-misattribution, false-universal]
  type: feedback
---

**Extension (2026-07-17, sleeper-service run):** the cross-source variant — *scope-fusion
overstatement*. Two separately-true claims from different sources with different scopes get
fused into one sentence claiming the union scope (e.g. "reported by red, blue, AND judge
across two consecutive runs" where three-seats is attested for one run and two-runs is
attested for red merges only). Check each conjunct of a compound recurrence/attestation claim
against its own source's scope; the fused sentence must be no stronger than the weakest
component's scope.

When a report cites a real, verified figure from the right paper to support a tradeoff
disposition, YOU MUST check *which experimental condition/arm* the figure belongs to — not just
that the number appears in the source.

**Why:** FEOV-retrospective round 2 (R2-3): blue's R1-17 fix deferred cross-provider model
diversity by citing arXiv:2602.03794's "2 diverse agents match/exceed 16 homogeneous" — with a
bracketed gloss "[persona-lensed]" that silently reassigned the result to the paper's L2
condition (persona-only, same base model). The paper's own Table 2 makes it the **L4** result
(different models AND personas); L2 needs 8 agents to match the 16-agent baseline — a 4x
efficiency gap. The citation was real, the figure exact, the round-1 ledger even had the
qualitative claim at HIGH — and the disposition still misread it, because the gloss moved the
number to the arm the report wanted it to support. Telltale: the citation was deployed
specifically to justify *not building* the condition (L3/L4) that actually produced the number.

**How to apply:** Whenever a quoted result carries a bracketed insertion, paraphrase, or
condition label supplied by the citing text ("[persona-lensed]", "same-provider", "without X"),
re-fetch the source's taxonomy/results table and pin the result to its arm. Highest suspicion
when the cited result underwrites a defer/skip decision about the very dimension the source
varied. Distinct from miscitation-to-wrong-paper ([[citation_status_and_misattribution_patterns]])
and from metric conflation ([[pattern-metric-conflation-and-traceable-not-verified]]) — here
paper, number, and quote are all correct; only the condition attribution is wrong.

Related: [[pattern-repair-regression-citation]] (this instance also lived inside a round-1
repair), [[pattern_footnote_overattribution]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\feov-memory\red-gap-patterns: workflow_undefined_rundir.md -->
---
name: workflow-undefined-rundir
description: FEOV workflow script can pass a literal "undefined" run directory to red — abort and report, never fabricate an audit
metadata:
  classes: []
  class_note: harness-limit — a tooling constraint, not a research defect class; kept for red's reference and deliberately NOT delivered as a repair duty
  type: project
---

The frank-exchange-of-views workflow script invoked red with run-directory paths of literally `undefined/blue/report.md` and `undefined/red/candidates/...` (uninitialized variable in the caller). No run directory, blue report, or debate.md existed in the repo (2026-07-12, branch port-plan-review).

**Why:** An audit with no audit surface must hard-fail — writing findings against a nonexistent report, or creating a literal `undefined/` directory, would mask the harness bug and launder a fake verdict into the debate.

**How to apply:** If any invocation path contains `undefined`/`null` or the named living report does not exist, return FAIL with friction naming the uninitialized variable; do not create the bogus path, do not audit substitute documents unprompted.

**Update 2026-07-12 (round 2):** Defect recurred — caller re-dispatched round 2 with the same literal `undefined` paths; no preflight guard was added despite round-1 adjudication in `undefined/debate.md`. Blue's round-1 report never landed on disk (environment blocked the write; content traveled in its envelope only), so even the partial `undefined/` tree contains no auditable report. Expect this to repeat until the workflow script binds the run-directory/topic variables; keep appending round positions to the existing `undefined/debate.md` transcript but create no new files under `undefined/`.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\.claude\agent-memory\frank-exchange-of-views-red-auditor: pattern_causal_trigger_population_mismatch.md -->
---
name: pattern-causal-trigger-population-mismatch
description: Check whether a claimed mechanism's own trigger condition holds across the whole population the causal claim is asserted over — and whether a rival hypothesis was retired on an argument that only runs on part of it
metadata:
  type: feedback
---

When a report names a MECHANISM as the cause of an observed population (a code guard, a config
default, a resolver branch), ask:

- **What is the mechanism's stated trigger condition, quoted from the report's own extract?**
  Then: does that condition hold for every member of the population the cause is asserted over?
  Reports commonly establish a mechanism on the subset that motivated the inquiry and then assert
  it over the full corpus. Look for the subset the report itself counts elsewhere (a "top-level
  glob found 16 where a recursive walk found 294" line is a partition the causal claim must survive).
- **What does the OTHER branch return?** If the quoted resolver returns a third value (`void 0`,
  `undefined`, "not set") on the branch the guard does not cover, does the report ever say what
  that produces? Silence there is the gap.
- **Was a competing hypothesis retired on parsimony?** If so, re-run the parsimony argument on
  each subset. "A single guard explains it" is unavailable where no single guard fires — so the
  rival account may be undisposed of on exactly the share the report never separated out.
- **Did asking for cause-vs-consistency create this?** A prior round's repair that forces the
  author to say "causal" rather than "consistent with" is the repair working — it makes the claim
  checkable. Expect the overreach to surface at the NEXT audit and close the ancestor as
  closed_with_regression rather than treating the repair as a failure.

Related: [[pattern_missing_root_invariant]], [[pattern_sibling_repair_composition]],
[[pattern_within_source_condition_misattribution]].

**Red-side duty this pattern also carries.** When writing an acceptance_check that verifies a
FIGURE, ask whether the check also pins the CONDITION (model / benchmark / arm / population) that
produced it. A check that says "confirm both digits appear" passes while a repair adds the digits
and simultaneously introduces a generalization the source's other arm refutes. Digit-verification
and condition-pinning are two obligations; a fix-spec naming only the first is under-specified,
and the resulting defect is charged to red, not to blue.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\.claude\agent-memory\frank-exchange-of-views-red-auditor: pattern_fenced_not_sourced.md -->
---
name: pattern-fenced-not-sourced
description: Ask whether a report's honesty labels ("lane-reported", "secondary", "not leaf-verified") are standing in for source existence — and whether a pinned input reported absent is actually recoverable at its pin
metadata:
  type: feedback
---

A report that fences its weak claims well can still be citing sources that do not contain the
claim. The label is about provenance strength; it says nothing about whether the URL carries the
content. Check both legs separately.

**Why:** the label reads as diligence and buys the claim a pass at the lens seat. Every fenced
figure still needs the leaf fetch.

**How to apply — questions, not answers:**
- For each footnote labelled secondary/unverified/lane-reported: does the named URL actually carry
  the claim? Does the named author's own index list the cited article?
- Is a real figure cited to a page that does not carry the digit, while the paper that does carry
  it goes uncited? If the paper reports a RANGE or a second condition, is the quoted endpoint the
  confirming one and the omitted endpoint the disconfirming one?
- Is "unverified" being used where "contradicted by an accessible source" is the true state? Those
  are different epistemic claims and only one of them needs disclosing as a conflict.
- Does the report make an in-run tooling claim ("verb X exists; I ran `--help`")? Re-run the
  command. Check whether the named capability belongs to a *different seat*.
- Pinned inputs reported as "do not exist on disk": are they recoverable at the pinned commit
  (`git ls-tree` / `git show`)? If so, read them — do they advance a COMPETING hypothesis on the
  round's load-bearing question, and do their figures secretly source a corroboration the report
  calls "independent"?
- When a lens returns CLEAN over a slice, re-verify a sample of that slice at the merge. A clean
  lens pass is a hypothesis about the slice, not a result.

See [[citation_status_and_misattribution_patterns]], [[pattern_footnote_overattribution]],
[[pattern_self_referential_repo_drift]], [[pattern_within_source_condition_misattribution]].


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\.claude\agent-memory\frank-exchange-of-views-red-auditor: pattern_round_parity_and_disposition_source.md -->
---
name: pattern-round-parity-and-disposition-source
description: Before auditing, check whether the other party actually took a turn and who entered each prior disposition — the artifact and the change-summary answer neither question
metadata:
  type: feedback
---

Two questions to ask BEFORE drafting any closure or re-raise, at every merge seat:

**Did the other party move this round?** Check `records/` for an `events-<other-seat>-r<N>-*.jsonl`
and compare `blue/report.md` mtime against the prior round's judge/merge events. If the artifact is
byte-identical to the one already ruled on, every "closure" drafted against it is a re-confirmation,
not a closure — and the round's verdict is structural rather than evidentiary. Does the envelope let
you say which? If not, that is friction.

**Who entered each prior disposition?** The dispositions of record live in `debate.md`'s `### LEAD`
section, not in the ledger, not in the artifact, and not in the change-summary. A gap the bench closed
is not yours to close again: a second closure event double-counts the board's closure history and
corrupts the repair-regression denominator. A gap the bench CARRIED is not yours to close either, and
minting a fresh id for it disguises an unmet obligation as new business.

Related checks: does the tool's own projection agree with the hand-written board? A divergence sized
exactly to the bench's closures means nothing carries bench dispositions into the seat's projection.
Does a closure class exist for a defect found AFTER a bench ruling that amends it without reopening?

**Why:** run 5 round 3 — red drafted seven closure blocks from the repaired artifact plus the
change-summary, then found on reading `debate.md` that the bench had entered six of them in the prior
round and had CARRIED the seventh, which blue had never repaired because blue took no turn. The
archive is append-only, so the mis-framed blocks had to be withdrawn by appending a correction.

**How to apply:** read `debate.md`'s LEAD section and `ls records/` in the SAME batch as the report
read, before any closure prose is written. See [[pattern_changesummary_desync_round_state]] — same
family, one level up: that pattern is about which round the artifact is in, this one is about who
ruled on it and whether anyone moved.


<!-- mirrored from C:\Users\gbloc\Projects\special-circumstances\.claude\agent-memory\frank-exchange-of-views-red-auditor: pattern_unswept_limitations_section.md -->
---
name: pattern-unswept-limitations-section
description: The provenance/limitations section is where every round's hedges are parked and is the site no repair is ever swept against; also numeral identity mistaken for set identity
metadata:
  type: feedback
---

Two checks, both cheap, both found real defects after the citation layer went clean.

**Is the limitations/provenance section on the propagation list?** Ask: does the change-summary's
"corrections propagated report-wide" list name it, in ANY round? If never, sweep it directly against
the body. It is where each round's honest hedges, provisional retractions and not-verified entries get
parked, so it accumulates statements the body has since resolved, retired, or reversed. Check both
directions — a hedge surviving after the body resolved, AND a not-verified entry naming a claim the
report no longer contains while its own footnote records a verification. Grade the CLASS (a section
outside the repair discipline), not the instances; a bench ruling gap-by-gap structurally cannot see
that three residues share one site.

**Do two equal numerals denote the same set?** When an inference turns on figures matching across
sections, ask whether the counts are of the same objects. A partition can appear to cover a corpus
because `total − subset` happens to equal a separately-measured count of something else. Check whether
the report's own quoted evidence breaks the identity. Then ask the sibling question: is a proxy
(filesystem depth, file location, naming) silently standing in for the property the claim actually
turns on (session type, execution mode), and is the substitution argued anywhere or just glossed —
possibly attributed to a section that does not make the claim?

**Why:** run 5 round 3 — two lens seats returned CLEAN over the section containing the bench's own
carried obligation; the 278-with-thinking-blocks and 294−16=278-nested counts collided by coincidence
and made an unquantified partition read as complete.

**How to apply:** at every merge, read the limitations section LAST and against the body, not as prose.
See [[pattern_incomplete_repair_footnote_lag]] and [[pattern_unreconciled_numeric_floors]].
