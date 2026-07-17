# RED audit — round 1, lens 2 (logic and completeness)

## Findings

### L2-F1: Supersession claim conflates retrieval and durability layers

**Location**: § "H3 — Partially Superseded by FEOV/qmd Convergence (Item 6 Only)" (lines 19–25)

**Quoted sentence**: "The §8 recommendation for an SQLite/embedding index ceiling (item 6) has been genuinely superseded by the qmd recall layer (PR #18, shipped 2026-07-14...)"

**Issue**: The conclusion that qmd "superseded" the design is a layer-confusion leap. The report itself admits in Residual Caveats (line 69) that "qmd is searchable but does not solve the consolidation rewrite-corruption problem." The original recommendation addressed DURABILITY (in-process, mandatory gate with append-only semantics); qmd solves RETRIEVAL (searchability). Solving retrieval does not supersede a durability design—it provides an alternative layer that makes durability *less urgent*, but the report frames it as a direct replacement. "Partially validated on item 6 only" glosses over this distinction. The honest reading: qmd is orthogonal infrastructure that alleviates pressure on item 6, but does not technically supersede it.

**Risk**: medium. The conclusion doesn't change (item 6 is addressed or deferred), but the reasoning mischaracterizes what was shipped, which could lead to false confidence that the consolidation rewrite-corruption risk is closed.

---

### L2-F2: Disconfirmation via artifact names assumes unique naming

**Location**: § "H5 — Key Items Implemented, Others Closed by Design Choice" (lines 38–48)

**Quoted sentence**: "Grep across main for 'dream' (Phase 5 command): zero results... Grep across main for 'ingest' (Phase 4 intake): zero results..."

**Issue**: The disconfirmation searches grep for `/dream`, `/ingest`, `/remember`, OKF schema as proxies for "consolidation was not implemented." But the absence of these *named* artifacts does not prove the absence of the *capability*. Consolidation could be implemented under different command names, embedded in other features, or built into a different layer (e.g., inside qmd's indexing, inside the hook system). The searches prove the specific lexicon wasn't used, not that the function wasn't built. For H5 falsification to hold, the report should have searched for consolidation BEHAVIOR (e.g., deduplication, conflict resolution, merging, alias tracking) under any name, not just the proposed command names.

**Risk**: medium. The conclusion (H5 falsified) may still be correct, but the evidence is incomplete. A thorough disconfirmation would either (a) search for functional equivalents under other names or (b) audit the code to prove consolidation logic is absent, not just the lexicon.

---

### L2-F3: "Memory-architecture feature" vs. "infrastructure capability" distinction undefined

**Location**: § "H2 — Infrastructure built, domain logic pending" (lines 15–17)

**Quoted sentence**: "Items 2, 5, 13 (agent-memory row fix, hooks test matrix, projection health) are not shipped as memory-architecture-specific features. They exist in broader infrastructure contexts (prosthetic-conscience 0.7.0+ includes hooks) but not as features of the memory-architecture design."

**Issue**: The claim rests on a binary that is never defined. If hooks testing is built into prosthetic-conscience 0.7.0+, and the memory-architecture recommendation called for hooks testing, then either (a) the capability is shipped and satisfies the recommendation (making H2 "validated, infrastructure-first") or (b) it's not, in which case "broader infrastructure" doesn't matter. The report uses "feature-specific" vs "infrastructure-context" as a watertight distinction, but doesn't justify it. Is a capability only a "memory-architecture feature" if it's under the memory-architecture directory? Or if it's called out with that label in the plan? The carve-out is unclear and risks post-hoc redefining "not implemented" to mean "implemented but in a context I didn't expect."

**Risk**: low-medium. The conclusion (items not shipped as intended) may be sound, but the distinction between "feature" and "infrastructure implementation" needs definition to avoid moving the goalposts.

---

### L2-F4: Blue pre-judges red's verification authority

**Location**: § "Status of Unverified Blocking Candidates (R4)" (lines 50–58)

**Quoted sentence**: "Both are framed as 'hardening, not redesign' and marked 'blue-fixed in §15' of the report. However, the report's own disposition is UNVERIFIED — red never verified them at the leaf node."

**Issue**: Blue is asserting "red never verified them," which is blue claiming authority over red's verification status. In the adversarial structure, blue's role is to propose and defend; red's role is to verify or reject. Blue cannot know whether red will accept or deny a claim—that's red's decision to make. The proper framing is: "These remain open pending red verification" or "Red has not yet audited these claims." Saying "red never verified" is either (a) a prediction of red's future work (which blue shouldn't make) or (b) a claim about what red did in prior rounds (which should be sourced to red's archive, not asserted). This conflates blue's self-review with red's gatekeeping.

**Risk**: low. The factual conclusion (items need verification, are currently unverified) is correct; the error is in the framing. This is a structural misalignment, not a factual gap.

---

### L2-F5: Phase 0/1 deferral stated as explicit but source is unspecified

**Location**: § "H1 — Deferred as Phase 2+ [minority: lane-1]" and § "H4 — All Blockers Remain Open..." (lines 11–13, 27–36)

**Quoted sentence**: "All blocking and High-priority items (poisoning mitigation, consolidator fixes, clone-time injection defense, provenance taxonomy, Auto Dream scope resolution) remain unimplemented and explicitly deferred as Phase 2+ or later, with Phase 0/1 focused on FEOV debate infrastructure, hooks inventory, and qmd adoption."

**Issue**: The word "explicitly" suggests this deferral is documented somewhere (plan, decision record, code comment), but the report provides no source. Footnote [^H1Finding] cites "frontier.md H1; lane-1 validates that Phase 2+ is the stated deferral," which means blue's own frontier document confirms it—a circular reference. Is this deferral in the memory-architecture report (the July 12 artifact)? Is it in the plans file on the unmerged branch? Is it inferred from the absence of these items on main? The specific source document and location must be cited to verify "explicitly."

**Risk**: low-medium. The deferral is probably accurate (evidenced by the lack of shipping), but "explicitly" overstates the confidence if it's an inference rather than a cited statement.

---

### L2-F6: Branching verification assumes main is sole source of truth

**Location**: § "H4 — All Blockers Remain Open; No Implementation Started" (lines 27–36)

**Quoted sentence**: "No implementation work has shipped to the primary codebase since the July 12 report... Verified via `git log main..plans/memory-architecture --oneline` (commit 32f13b2 only) and `git branch -a --contains 32f13b2` (plans/memory-architecture branch only). The branch is tracked at origin but not merged to main."

**Issue**: The verification checks that work is on an unmerged branch, not on main. This proves it hasn't reached production, but doesn't rule out: (a) work on a third branch (not main, not plans/memory-architecture) that is tracking these recommendations; (b) incremental commits on main that implement pieces without being labeled "memory-architecture"; (c) work that was started, shipped, then deleted/reverted. The report assumes the branch is the ONLY locus of memory-architecture work. To close this, a sweep of all branches for commits touching the relevant files/features would be needed.

**Risk**: low. The assumption (main = authoritative) is reasonable for a shipped-vs-unshipped distinction, but the verification is narrower than stated.

---

### L2-F7: Functional alternative (qmd as durability workaround) not explored

**Location**: § "Residual Caveats" (lines 67–70)

**Quoted sentence**: "qmd is searchable but does not solve the consolidation rewrite-corruption problem (risk matrix line 53: 'append-only rule' = the actual fix, not yet implemented). The two address different layers (retrieval vs durability), so 'qmd replaced SQLite' is accurate for the specific recommendation but not a full substitution of the philosophy."

**Issue**: The report correctly notes that qmd doesn't solve durability. But it doesn't explore the counterargument: *Does the durability problem matter if you never need to consolidate?* If qmd keeps full-text search fast and accessible, does the original need for an "SQLite ceiling" still apply? The problem the recommendation was trying to solve was "consolidation-complexity runaway" (line 69). But if you have searchable, indexed memories that don't need merging, does runaway complexity still threaten the system? The report acknowledges the difference in layers but doesn't ask whether the retrieval layer functionally solves the underlying problem by making consolidation optional/deferrable. This is a missing alternative explanation for why item 6 might be *addressed* (not by the letter, but by the intent).

**Risk**: low. The finding (item 6 is superseded or deferred) is likely correct either way, but the reasoning for how/why is incomplete.

---

### L2-F8: Methodology scope unclear ("smoke run" + deep verification)

**Location**: § "Methodology" (lines 73–76)

**Quoted sentence**: "This is a smoke run (shallow, mechanical): a pipeline exercise confirming the review process, not a full audit. Research was targeted to the five hypotheses; disconfirming evidence was sought first (grep for shipped code contradicting the 'unimplemented' finding). All evidence is from the repo's git history and the pinned research directory (de8d9c2)."

**Issue**: The report is labeled "smoke run (shallow, mechanical)" but uses leaf-node verification (git log, git show, grep, file audits, footnote citations with specific line numbers and commit SHAs). These are hallmarks of *deep* verification, not shallow smoke testing. A smoke test would spot-check 1–2 hypotheses; this tests all 5 and seeks disconfirming evidence. Either (a) the scope is deeper than labeled and should say so, or (b) the verification is shallower than it appears (e.g., trusting lane-1's spot checks without re-doing them). The mismatch could lead to false confidence in a "shallow" process that is actually deep, or false doubt in a "deep" process that is only spot-checking.

**Risk**: low. This is primarily a labeling issue, not a gap in evidence. But it confuses the trust model for consumers of this report.

---

### L2-F9: Circular evidence (blue cites blue's frontier as disconfirming source)

**Location**: Multiple footnotes, exemplified by [^H1Finding] and [^H5Finding]

**Quoted sentence**: "[^H1Finding]: frontier.md H1; lane-1 validates that Phase 2+ is the stated deferral. Access: 2026-07-17." and similar patterns throughout.

**Issue**: The report uses `frontier.md` (blue's own hypothesis document) as the source for facts. For example, blue wrote a hypothesis "H1: items deferred as Phase 2+," then when auditing, blue cites that same hypothesis file as evidence that H1 is validated. This is circular: the hypothesis is not evidence for itself. The footnotes should cite external sources (the July 12 report, the plan file, git commits, the port-plan document) to verify that Phase 2+ deferral is real, not just that blue hypothesized it. The same issue applies to "lane-1" citations—lane-1 is blue's own methodology lane; blue can't use its own methodology as evidence for its findings.

**Risk**: medium. The findings are probably correct, but the citation model is broken for any claim that rests solely on blue's own prior work. Independent verification from the source material (July 12 report, plans, git history) is needed.

---

### L2-F10: Missing counterargument (plan obsolescence, priority shifts)

**Location**: § "H1 — Deferred as Phase 2+ [minority: lane-1]" and § "H4 — All Blockers Remain Open..." (lines 11–36)

**Quoted sentence**: "Most blocking and High-priority items (poisoning mitigation, consolidator fixes, clone-time injection defense, provenance taxonomy, Auto Dream scope resolution) remain unimplemented and explicitly deferred as Phase 2+ or later..."

**Issue**: The report states items are deferred to Phase 2+ as a fact, but doesn't explore the counterargument: *What if the plan changed?* What if the team decided to tackle Phase 2 items now, or to descope them entirely? The report verifies that code wasn't shipped, but doesn't verify that the deferral is still in force. A valid alternative hypothesis: "Phase 0/1 work is complete, Phase 2 priorities were re-evaluated, and these items are now cancelled or reprioritized." The report's disconfirming evidence (no code shipped) would still hold, but the *interpretation* (deferred to Phase 2) would be wrong. To close this, check the current plan, not just the July 12 report and the unmerged branch.

**Risk**: low. The conclusion (items are not shipped) is solid; the interpretation (deferred, not cancelled) is assumed. A current plan read would resolve this.

---

## Summary grading

| Finding | Confidence | Severity | Class |
|---------|-----------|----------|-------|
| L2-F1   | high      | medium   | layer-conflation in conclusion |
| L2-F2   | high      | medium   | incomplete disconfirmation |
| L2-F3   | high      | low      | undefined term, not yet exploited |
| L2-F4   | high      | low      | structural (blue/red role misalignment) |
| L2-F5   | high      | low      | missing source citation |
| L2-F6   | high      | low      | narrow verification, reasonable assumption |
| L2-F7   | medium    | low      | missing alternative explanation |
| L2-F8   | high      | low      | labeling/scope mismatch |
| L2-F9   | high      | medium   | circular evidence (blue citing blue) |
| L2-F10  | medium    | low      | missing disconfirming hypothesis (plan change) |

**Open for red verification**: 
- L2-F1, L2-F2, L2-F9 warrant deeper leaf-node checks (does qmd actually supersede in any sense? are there hidden consolidation features? what does blue's frontier.md actually say?)
- L2-F5, L2-F10 need a current plan read to resolve
