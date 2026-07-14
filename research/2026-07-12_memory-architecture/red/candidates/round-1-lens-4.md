# Red audit — Round 1, Lens 4: logic & completeness

**Surface audited:** full `blue/report.md` (living, Round 0), read in context against
`inputs/memory-architecture-proposal.md`, `references/report_template.md`,
`references/heilmeier_template.md`, and `blue/frontier.md`.

**Lens verdict: CHANGES-REQUESTED.** The report is evidence-dense and its factual corrections
verify at the leaf node. But on *logic and completeness* it carries one structural omission
(the report never delivers the re-scoped build plan its own §3 argues is mandatory), one
self-undercutting evidence chain (the "strongest validation" rests on the report's own
weakest-verified item), one blocking-grade escalation that skips the suite-specific likelihood
argument, and several fixes advanced without their obvious counterargument. None of these are
"the finding is wrong" — they are "the argument has a hole a skeptic walks through." Graded
below.

---

## Gaps (graded, cumulative anchor for the round)

### L4-1 — The report diagnoses a re-scope but never delivers it (completeness, HIGH)

**Location:** §3 "The competitive landscape moved" — *"The bespoke layer's defensible remit
shrinks to what native does not do: cross-project global knowledge as a reviewable git repo;
typed/schema'd concepts; external-source ingest with provenance; human-gated promotion to
skills; the project store committed with the code. The build plan should be re-scoped so phases
duplicate nothing the harness ships."*

The single most consequential claim in the report is that native machinery (Auto Memory shipped,
Auto Dream rolling out) now covers per-project capture and consolidation "without building
anything," collapsing the proposal's remit to five items. §8 change #4 then says "re-scope against
native machinery" and "drop bespoke work duplicating native capture" — but **the report never
produces the re-scoped phase plan.** The proposal's §10 has six phases (0–5); the report leaves
the reader to guess which survive. If §3 is right, Phases 1 (single-trajectory capture) and 2's
`MEMORY.md` ingest are candidates for deletion — that is the actionable core of the audit, and it
is missing. A skeptic asks: "You told me to build less. What, exactly?" The report has no answer.
- Likelihood the omission bites: **High** (this is the decision the operator must make next).
- Impact: **High** (without it the audit is diagnostic, not actionable).
- Complexity to fix: **Low** (one re-mapped phase table).
- Disposition: **Fix.** Deliver the re-scoped phase plan, or state explicitly that re-scoping is
  deferred to a decision the lead/operator owns and why.

### L4-2 — "Strongest validation" leans on the report's own weakest-verified evidence (logic, HIGH)

**Location:** §3 Consequences — *"Anthropic independently building trajectory-signal-gathering +
scheduled consolidation is the strongest available evidence that the proposal's core loop is the
right shape."* Cross-read with §10 Unverified — *"Native Auto Dream availability — verified as
concept and community replication, unverified as a dependable API (server-side flag)."*

The report's headline validation argument (verdict line: *"Anthropic is independently converging
on the same loop natively"*) rests on Auto Dream, which the report itself files under **Unverified
items**. The two footnotes for it (`[^AutoDream]`, `[^DreamSkill]`) are third-party blog posts and
a community skill "replicating Anthropic's *unreleased* auto-dream feature" — not Anthropic
documentation. So the strongest load-bearing evidence for the verdict is corroborated at **low
confidence** by the report's own admission. This is not "the claim is false" — convergence is
plausible — but the report presents a low-confidence item as its keystone without flagging the
tension. Either downgrade the rhetoric ("suggestive, not strongest") or the verdict inherits the
unverified item's confidence.
- Corroboration confidence (statement↔reference): **Low** (blog/community sources; report concurs).
- Impact: **Med** (weakens the verdict's stated basis, not its direction).
- Disposition: **Fix by re-framing.** Do not let an item on the Unverified list carry the word
  "strongest."

### L4-3 — "Blocking" poisoning grade skips the suite-specific attacker model (logic / graded-risk, HIGH)

**Location:** §4 — *"one missing threat model severe enough to be blocking (memory poisoning...)"*
and §9 risk row — *"Memory poisoning via ingest/inbox (§4) | Med (single operator, but npm-CVE
precedent; 80–99% reported attack success)."*

The 80–99% attack-success figures ([^MemoryPoisonSurvey]) are from adversarial red-team studies of
agent memory systems; the CVE ([^MemoryPoisonCve]) is a real in-the-wild npm vector. Both verify.
But the report escalates to **"blocking before Phase 1"** for a **single-operator, machine-local,
optionally-private** store without ever building the *who-attacks-this* argument. Attack-*success-if-
attempted* is not attack-*likelihood*; the report conflates them in the same cell ("Med ... 80–99%
success"). The genuinely load-bearing exposure is narrow and the report already names it: `/ingest
<url>` and mid-session web reads (the untrusted-input edges). A skeptic accepts *those two edges
need a gate* while rejecting the full apparatus (trust tiers + injection screening at capture AND
promotion + de-authorized projection voice + independent-source corroboration) as complexity
priced against a low-probability targeted attack on a private solo store. Per the mandate:
surface it, grade it, let the tradeoff be argued — do not force blue to absorb five mitigations to
satisfy one attacker model that was never specified.
- Likelihood (this suite, targeted): **Low–Med**, not the "Med backed by 80–99%" the cell implies.
- Impact: **High** (persistent compromise) — undisputed.
- Complexity to mitigate: report says Med; the *full* set is arguably High and makes the design
  strictly heavier.
- Disposition: **Contest the grade, not the risk.** Keep the two ingest-edge gates as blocking;
  require blue to justify each additional mitigation against a stated attacker model, or demote
  the surplus to High/Med. Candidate for the lead's docket if re-raised.

### L4-4 — Append-only expansion trades rewrite-drift for unbounded concept-file growth (missing counterargument, MED)

**Location:** §2.3 — *"corroboration appends to the Evidence section and bumps counters in
frontmatter... This turns every consolidation diff into additions + frontmatter bumps..."*

The fix for continuous-rewrite corruption is sound, but the report advances it without its obvious
cost: an append-only Evidence section **grows without bound** as a concept is re-corroborated over
months. The report elsewhere (§6.1) makes context-bloat a *measured performance regression*
([^ContextRotChroma]) and mandates a hard cap on `active.md` (§6.2) — yet the concept files that
feed the projection have no equivalent cap under the append-only rule. The two sections are in
tension and the report does not reconcile them. A skeptic notes: you solved drift by moving the
bloat one level down.
- Likelihood: **Med over months.** Impact: **Low–Med** (projection ranker may mask it).
  Complexity: **Low** (cap Evidence entries; keep last-N + count).
- Disposition: **Fix.** Add an Evidence-section cap (e.g. keep N most-recent corroborations +
  a total counter) so append-only does not re-import the pile problem.

### L4-5 — Dropping the confidence float leaves merge/precedence tie-breaks undefined (completeness, MED)

**Location:** §6.2 — *"Drop the stored confidence float in v1... Derive activation from observables
(status: active AND review_count ≥ 2 AND last_seen within window AND trust tier sufficient)."*

The simplification is well-argued for *activation*. But the proposal uses `confidence` in two other
places the report does not re-home: §8 *"Confidence breaks intra-scope ties"* and *"higher
confidence + review_count wins the merge."* If the float is deleted, what breaks a merge tie when
`review_count` is equal? The report deletes the input to two decision rules without supplying the
replacement. Incomplete, not wrong.
- Likelihood: **Med** (ties occur). Impact: **Low** (arbitrary-but-deterministic fallback exists).
  Complexity: **Low**.
- Disposition: **Fix.** Name the replacement tie-breaker (e.g. `last_seen` recency, then
  provenance tier) wherever the proposal cited `confidence`.

### L4-6 — Git-diff review demotion rests on a setting-mismatched evidence transfer (graded corroboration, MED)

**Location:** §2.4 — *"a single operator reviewing nightly dream diffs will decay to LGTM within
weeks"* citing [^BotReviewFatigue][^UnreviewedPRs][^AIApprovingPRs].

The cited data (Dependabot ~54% merge, 61.4% of agent PRs unreviewed, 71.6% of review comments
agent-authored) is from **multi-contributor OSS with bot-noise queues** — a materially different
setting from a solo operator reviewing his own system's output, where personal investment and low
volume cut the other way. The conclusion (demote git-diff to forensic, add structural guards) is
likely still correct and the structural guards are cheap, so this is not a reason to keep git-diff
as the sole preventive guard. But the *evidentiary bridge* to "solo operator will LGTM within
weeks" is an extrapolation, not a measurement.
- Corroboration confidence (statement↔reference): **Medium** — sources are real and on-topic but
  from a different reviewer population; the "single operator" conclusion is inferred.
- Disposition: **Accept the recommendation, flag the claim.** Relabel as reasoned inference, not
  measured, or find solo-maintainer review-fatigue evidence.

### L4-7 — The "no alternative dominates" verdict never runs the "wait and build nothing" branch (unexplored alternative, MED)

**Location:** §7 — *"Bespoke remains justified for the shrunken remit; no external adoption
dominates."* and §11-adjacent framing.

H5 surveyed products (claude-mem, basic-memory, mem0, Letta, Zep) and native-surfaces-plus-thin-
skill. Given §3's own claim that native Auto Memory + Auto Dream are converging on the loop, the
logically forced alternative is **"defer the build 3–6 months and let native mature, building only
the irreducible git-repo/typed-concept/ingest layer when native gaps are confirmed."** The report
gestures at this ("scope transfer... without building anything") but never evaluates *timing* as a
decision — build-now vs. build-later-thinner. For a single operator this is arguably the dominant
option and it is unexamined.
- Likelihood the omission matters: **Med.** Impact: **Med** (could change what gets built now).
  Complexity: **Low** (one paragraph of decision analysis).
- Disposition: **Fix.** Add the timing/defer branch to §7 or §11 and say why build-now beats it
  (or that it does not).

### L4-8 — Heilmeier cost/schedule dimension is absent; 14 changes graded by priority, not effort (completeness / template, LOW–MED)

**Location:** §8 change table (grades Blocking/High/Med/Low) and the report as a whole.

The final `report.md` must carry the Heilmeier Catechism, whose Q7 ("What does it cost?") and Q8
("How long?") have **no answer anywhere in blue's living report.** §8 grades are *priority*, not
*effort*: change #1 (poisoning threat model) and change #3 (build a secret scanner) are non-trivial
engineering, graded identically to change #14 (reframe a doc sentence). A reader cannot tell whether
the 3 blocking + 11 other changes are a week or a quarter. This is the living report, not the final
assembly, so it is not strict non-compliance yet — but the cost/effort axis is the one dimension
the report never supplies, and the risk matrix's "Fix cost" column (Low/Med) is the closest proxy
and is not aggregated.
- Impact: **Med** (operator cannot sequence the work). Complexity: **Low** (annotate effort).
- Disposition: **Fix at assembly.** Add per-change effort or at least aggregate the blocking set's
  cost so the Heilmeier Q7/Q8 answers are derivable.

### L4-9 — Verdict optimism vs. blocking-defect count is unreconciled framing (logic, LOW)

**Location:** Verdict — *"The architecture is directionally right and better-supported by external
evidence than the proposal itself knows"* — immediately followed by three blocking defects
(poisoning, wrong agent-memory row, non-existent secret-scrub).

A design with an unmitigated blocking security gap and a factually wrong load-bearing mapping row
is defensibly "directionally right in shape" — but the verdict leads with praise and buries the
blockers below the fold. This is a framing choice, not an error; the content is all present. Flagged
only so the final report's verdict stamp (VERIFIED/UNVERIFIED) does not inherit the optimistic lead
without foregrounding that Phase 1 is gated on three blockers.
- Disposition: **Fix at assembly.** The verdict stamp should read as gated, not endorsed.

---

## Not gaps (checked, cleared)

- Disconfirming budget: frontier records lane-1 7/21 and lane-2 explicit disconfirming searches —
  meets the 1-in-5 floor. No finding.
- Factual corrections (§1.2 agent-memory row, §6.3 secret-scrub, §6.3 `docs/scheduling.md`): these
  are corroborated by local-verification footnotes and match the proposal text I read; out of this
  lens's scope (logic/completeness), left to the leaf-node/citation lens.
- Unverified items (§10) are labeled, not laundered — compliant with the protocol's honesty rule.

## Friction

None for this task. All source artifacts (proposal, report, templates, frontier, debate) were
readable locally at the leaf node.
