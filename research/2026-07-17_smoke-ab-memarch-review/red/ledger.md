# red ledger — Round 1 open gaps + closure index

## Open Gaps (Round 1)

### R1-1: Qmd supersession conflates retrieval and durability layers
- **Location**: § "H3 — Partially Superseded by FEOV/qmd Convergence (Item 6 Only)" (lines 19–25)
- **Quoted**: "The §8 recommendation for an SQLite/embedding index ceiling (item 6) has been genuinely superseded by the qmd recall layer..."
- **Problem**: Report frames qmd as direct replacement for durability/append-only layer; actually solves retrieval (searchability) only. Body verdict conflicts with admitted caveat (line 69).
- **Required fix**: Reframe as "qmd addresses retrieval; consolidation durability (append-only rule, risk matrix line 53) remains unimplemented" and clarify which risk is closed vs. still open.
- **Severity**: medium
- **Likelihood**: high
- **Impact**: medium
- **Complexity**: low
- **Found by**: ["L2", "L3"]
- **Supersedes**: (none)

### R1-2: Disconfirmation via command-name grep incomplete
- **Location**: § "H5 — Key Items Implemented, Others Closed by Design Choice" (lines 38–48)
- **Quoted**: "Grep across main for 'dream' (Phase 5 command): zero results... Grep across main for 'ingest' (Phase 4 intake): zero results..."
- **Problem**: Lexical absence of proposed command names does not prove functional absence; consolidation logic could exist under different names or in non-markdown layers.
- **Required fix**: Search for functional equivalents (deduplication, conflict resolution, merging logic under any name) or audit code layer directly to prove consolidation logic absent; or soften verdict from "falsified" to "no evidence of shipping in primary channels."
- **Severity**: medium
- **Likelihood**: medium
- **Impact**: medium
- **Complexity**: medium
- **Found by**: ["L2"]
- **Supersedes**: (none)

### R1-3: "Memory-architecture feature" vs "infrastructure capability" distinction undefined
- **Location**: § "H2 — Infrastructure built, domain logic pending" (lines 15–17)
- **Quoted**: "Items 2, 5, 13 (agent-memory row fix, hooks test matrix, projection health) are not shipped as memory-architecture-specific features. They exist in broader infrastructure contexts..."
- **Problem**: Binary (feature vs infrastructure) never defined; risk of post-hoc redefining "not implemented" to fit any evidence.
- **Required fix**: Define what counts as a "memory-architecture feature" (location in codebase? labeling? functional scope?) vs infrastructure implementation; apply definition consistently.
- **Severity**: low
- **Likelihood**: low
- **Impact**: low-medium
- **Complexity**: low
- **Found by**: ["L2"]
- **Supersedes**: (none)

### R1-4: Blue pre-judges red's verification authority
- **Location**: § "Status of Unverified Blocking Candidates (R4)" (lines 50–58)
- **Quoted**: "Both are framed as 'hardening, not redesign' and marked 'blue-fixed in §15' of the report. However, the report's own disposition is UNVERIFIED — red never verified them..."
- **Problem**: Blue asserts "red never verified" conflates blue's self-review with red's gatekeeping; blue cannot predict red's judgment.
- **Required fix**: Rephrase as "remain open pending red verification" not "red never verified"; red decides acceptance, not blue.
- **Severity**: low
- **Likelihood**: high
- **Impact**: low
- **Complexity**: low
- **Found by**: ["L2"]
- **Supersedes**: (none)

### R1-5: Phase 0/1 deferral stated as explicit but source unspecified
- **Location**: § "H1 — Deferred as Phase 2+" and § "H4 — All Blockers Remain Open..." (lines 11–13, 27–36)
- **Quoted**: "All blocking and High-priority items remain unimplemented and explicitly deferred as Phase 2+ or later..."
- **Problem**: "Explicitly" suggests documentation, but source missing; footnote cites blue's own frontier.md (circular) not external source (July 12 report, plans file, git commit).
- **Required fix**: Cite actual source document (July 12 report §, plans/memory-architecture.md line #, or git commit) where Phase 2+ deferral is explicitly stated; remove frontier.md circular reference.
- **Severity**: low
- **Likelihood**: high
- **Impact**: low
- **Complexity**: low
- **Found by**: ["L2"]
- **Supersedes**: (none)

### R1-6: Branching verification assumes main is sole source of truth
- **Location**: § "H4 — All Blockers Remain Open; No Implementation Started" (lines 27–36)
- **Quoted**: "No implementation work has shipped to the primary codebase since the July 12 report... Verified via `git log main..plans/memory-architecture --oneline`..."
- **Problem**: Verification checks only unmerged branch, not whether work was: (a) on a third branch, (b) shipped incrementally on main under different names, or (c) shipped then reverted.
- **Required fix**: Sweep all branches for memory-architecture work (git branch -a, grep all for consolidation/memory keywords); verify no incremental commits on main implementing pieces without memory-architecture label.
- **Severity**: low
- **Likelihood**: low
- **Impact**: low
- **Complexity**: low
- **Found by**: ["L2"]
- **Supersedes**: (none)

### R1-7: Functional alternative (qmd as durability workaround) not explored
- **Location**: § "Residual Caveats" (lines 67–70)
- **Quoted**: "qmd is searchable but does not solve the consolidation rewrite-corruption problem..."
- **Problem**: Report correctly notes qmd ≠ durability fix, but doesn't explore: If qmd keeps retrieval fast and deduplicated at scale, does the need for consolidation-durability machinery change? Could faster retrieval make consolidation optional/deferrable?
- **Required fix**: Analyze whether retrieval layer (qmd) functionally addresses the underlying problem (consolidation-complexity runaway) by obviating the need for consolidation machinery, or if durability layer is still required for ground-truth integrity.
- **Severity**: low
- **Likelihood**: low
- **Impact**: low
- **Complexity**: low
- **Found by**: ["L2"]
- **Supersedes**: (none)

### R1-8: Methodology scope label ("smoke run") does not match depth of work
- **Location**: § "Methodology" (lines 73–76)
- **Quoted**: "This is a smoke run (shallow, mechanical): a pipeline exercise confirming the review process, not a full audit... Research was targeted to the five hypotheses..."
- **Problem**: Report labeled "smoke run (shallow)" but uses deep verification (leaf-node git checks, line-by-line cite-to-source). Smoke test would spot-check 1–2 hypotheses; this audits all 5 + seeks disconfirming evidence.
- **Required fix**: Either relabel scope as "targeted deep audit" or perform shallower spot-check (sample 1–2 hypotheses); align label with actual methodology to avoid false confidence (readers trust "shallow" process but get deep verification, or vice versa).
- **Severity**: low
- **Likelihood**: high
- **Impact**: low
- **Complexity**: low
- **Found by**: ["L2"]
- **Supersedes**: (none)

### R1-9: Circular evidence — blue cites blue's frontier as disconfirming source
- **Location**: Multiple footnotes (e.g., [^H1Finding], [^H2Finding], [^H5Finding])
- **Quoted**: "[^H1Finding]: frontier.md H1; lane-1 validates that Phase 2+ is the stated deferral. Access: 2026-07-17."
- **Problem**: Blue uses its own hypothesis document (frontier.md) and its own methodology lane (lane-1) as evidence for its own findings. Hypothesis is not evidence for itself.
- **Required fix**: Replace all frontier.md and lane-1 citations with external sources (July 12 memory-architecture report, plans files, git commits); remove self-references as evidence. Any claim sourced only to blue's own work remains unverified.
- **Severity**: medium
- **Likelihood**: high
- **Impact**: medium
- **Complexity**: low
- **Found by**: ["L2"]
- **Supersedes**: (none)

### R1-10: Missing counterargument — plan may have changed (items cancelled, not deferred)
- **Location**: § "H1 — Deferred as Phase 2+" and § "H4 — All Blockers Remain Open..." (lines 11–36)
- **Quoted**: "Most blocking and High-priority items remain unimplemented and explicitly deferred as Phase 2+ or later..."
- **Problem**: Report states deferral as established fact but doesn't verify current plan status; items could be cancelled, reprioritized, or moved to different phases. Absence of code doesn't prove deferral is still in force.
- **Required fix**: Read current plan (not just July 12 report) to verify Phase 2+ deferral is still the active status; if plan changed, reframe findings.
- **Severity**: low
- **Likelihood**: medium
- **Impact**: low
- **Complexity**: low
- **Found by**: ["L2"]
- **Supersedes**: (none)

### R1-11: R4 blocking-candidates unverified and security-critical; gate blocks implementation
- **Location**: § "Status of Unverified Blocking Candidates (R4)" (lines 50–58)
- **Quoted**: "Both [R4-1 and R4-2] are framed as 'hardening, not redesign' and marked 'blue-fixed in §15' of the report. However, the report's own disposition is UNVERIFIED — red never verified them..."
- **Problem**: R4-1 (taint-boundary allowlist inversion) and R4-2 (git-ignore projections, commit bodies only) are load-bearing for security invariant. Original disposition (2026-07-12 report line 97) states: "R4-1/R4-2 structural fixes must be closed and independently verified before implementation proceeds." Report identifies them as unverified but buries the severity; they are not nice-to-have caveats, they are gates.
- **Required fix**: Either verify R4-1 and R4-2 at leaf node (trace allowlist inversion through all taint channels; confirm no auto-import path in fresh clones) or escalate to Phase-0 blocker or explicitly document that implementation cannot proceed until they are verified and closed.
- **Severity**: high
- **Likelihood**: high
- **Impact**: high
- **Complexity**: high
- **Found by**: ["L3"]
- **Supersedes**: (none)

### R1-12: Phase 2+ deferral lacks timeline and risk acceptance
- **Location**: § "H1 — Deferred as Phase 2+" (lines 12–13); § "H4 — All Blockers Remain Open..." (lines 27–29)
- **Quoted**: "All blocking candidates and High-priority items remain unstarted, awaiting Phase 0 FEOV/port-plan foundation work to complete."
- **Problem**: Deferral has no implementation timeline, no compensating controls, and no explicit risk-acceptance decision. Original disposition requires R4-1/R4-2 closed before implementation (gate, not deferral with expiration). System will operate in known-vulnerable state (memory poisoning, clone-time injection risks HIGH impact) for undefined period if Phase 0/1 slips.
- **Required fix**: Record risk-acceptance decision naming completion date for Phase 0/1 (e.g., "Phase 0/1 must complete by DATE before R4-1/R4-2 implementation begins") or escalate R4-1/R4-2 to Phase-0 gates, or implement compensating controls (restricted untrusted input, offline-only mode, additional review gates).
- **Severity**: medium-high
- **Likelihood**: medium
- **Impact**: medium-high
- **Complexity**: low
- **Found by**: ["L3"]
- **Supersedes**: (none)

### R1-13: Consolidation machinery is foundational, not peripheral — absence creates slow-burn risk
- **Location**: § "H4 — All Blockers Remain Open..." (line 23); § "Residual Caveats" (line 69)
- **Quoted**: "The memory-architecture's consolidation machinery remains unbuilt; qmd provides the retrieval layer for what exists."
- **Problem**: Framed as "residual caveat" but is architectural gap. Original design (2026-07-12 report Heilmeier §3) premises append-only durability on nightly consolidation pass (re-derive tiers, deduplicate, write immutable claims). Without it: no consolidation loop, no append-only enforcement, no tier re-derivation. Failure mode is "High over months" (silent knowledge loss). This is slow-burn control failure, not feature gap; will manifest as memory store grows.
- **Required fix**: Acknowledge as architectural blocker (Phase 2–5 component, not optional); either accelerate consolidation machinery to Phase 0/1 or explicitly accept slow-burn silent-data-loss risk over months with documented acceptance and monitoring plan.
- **Severity**: high
- **Likelihood**: high
- **Impact**: high
- **Complexity**: high
- **Found by**: ["L3"]
- **Supersedes**: (none)

### R1-14: Smoke-run verdicts over-confident; methodology not calibrated to confidence
- **Location**: § "Methodology" (lines 74–76)
- **Quoted**: "This is a smoke run (shallow, mechanical)... Disconfirming evidence was sought first (grep for shipped code contradicting the 'unimplemented' finding)."
- **Problem**: Report asserts high-confidence "falsified" verdicts (H2: "Falsified," H5: "Falsified") on methodology that is admittedly shallow (grep only). Grep for command names doesn't prove functional absence; implementation could use different naming or exist in non-markdown layer.
- **Required fix**: Either soften verdicts to "no evidence of shipping in primary channels" or perform deeper verification (search for functional equivalents, audit code layer, not just lexical grep). For security-critical misses, high confidence requires deeper evidence.
- **Severity**: medium
- **Likelihood**: medium
- **Impact**: low-medium
- **Complexity**: medium
- **Found by**: ["L3"]
- **Supersedes**: (none)

### R1-15: Unmerged branch carrying all design work at risk of divergence
- **Location**: § "H4 — All Blockers Remain Open..." (line 29)
- **Quoted**: "The plan file `plans/memory-architecture.md` exists only in the `plans/memory-architecture` branch (commit 32f13b2), exactly one commit ahead of main (de8d9c2)."
- **Problem**: Long-lived unmerged branch at risk of becoming stale/diverged without merge timeline or keep-in-sync policy. If main evolves in memory layer, hooks, or MCP infrastructure, branch will diverge and integration rework required.
- **Required fix**: Commit to merge timeline within Phase 0/1, or establish "keep-in-sync" policy (rebase against main on Phase-boundary commits) to prevent divergence.
- **Severity**: medium
- **Likelihood**: medium
- **Impact**: medium
- **Complexity**: low
- **Found by**: ["L3"]
- **Supersedes**: (none)

### R1-16: qmd quality not audited; failure modes unexplored
- **Location**: § "Residual Caveats" (line 69); § "Methodology" (line 75)
- **Quoted**: "This run does not audit the quality of the qmd adoption relative to the item-6 recommendation, only its existence."
- **Problem**: Report verifies qmd exists but not whether it solves consolidation-complexity runaway or carries new risks (MCP server failure, refresh latency, scale ceiling). Caveat is footnoted but not carried into confidence grade for H3 verdict.
- **Required fix**: Audit qmd failure modes (MCP crash, stale embeddings, scale limits, silent search loss); add telemetry/recovery if necessary. Or lower confidence in H3 "Partially Validated" verdict.
- **Severity**: medium
- **Likelihood**: medium
- **Impact**: medium
- **Complexity**: medium
- **Found by**: ["L3"]
- **Supersedes**: (none)

### R1-17: Original disposition gate not emphasized — R4-1/R4-2 block implementation
- **Location**: § "Status of Unverified Blocking Candidates (R4)" (lines 50–58); cross-reference 2026-07-12 report line 97
- **Quoted**: "Both [R4-1 and R4-2] are framed as 'hardening, not redesign' and marked 'blue-fixed in §15' of the report... They remain open blocking gaps pending verification and implementation."
- **Problem**: Report correctly identifies R4 as unverified but doesn't emphasize original disposition requirement. Original line 97: "R4-1/R4-2 structural fixes must be closed and independently verified BEFORE IMPLEMENTATION PROCEEDS." This is a gate, not a deferral. Report's framing ("remain open pending verification") could be misread as "nice-to-have caveat that can be deferred to Phase 2+."
- **Required fix**: Explicitly state "This blocks implementation per the original disposition (2026-07-12 report line 97)" and escalate to Phase-0 blocker or document that Phase 2/3 gates must close them before shipping memory-architecture to production.
- **Severity**: high
- **Likelihood**: high
- **Impact**: high
- **Complexity**: low
- **Found by**: ["L3"]
- **Supersedes**: (none)

---

## Closure Index

(none yet — all gaps are OPEN)

---

## Round-1 Summary

**Total gaps:** 17  
**Open:** 17  
**Closed:** 0  
**Severity distribution:** 7 HIGH, 6 MEDIUM/MEDIUM-HIGH, 4 LOW/LOW-MEDIUM  
**Max severity:** HIGH (R1-11, R1-13, R1-17)  

**Blocking verdict:** FAIL — R1-11 (unverified security gates), R1-13 (missing architectural component), R1-17 (gate requirement not emphasized) must be closed before implementation can proceed per original disposition.
