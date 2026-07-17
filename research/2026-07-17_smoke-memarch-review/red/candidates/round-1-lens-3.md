# RED audit — round 1, lens 3 (dark-side & risk)

**Lens focus:** failure modes, likelihood × impact × complexity grading, security and tradeoff blindspots. This pass audits blue's evidence chains and confidence claims, then identifies gaps the blue report itself misses on risk.

---

## Preamble: Blue's Scope & Confidence

Blue performs a SMOKE-mode retrospective audit (2026-07-12 research completion → 2026-07-17 HEAD) on whether memory-architecture research recommendations were implemented. Five hypotheses tested; all marked DISCONFIRMED. High confidence on "zero implementation" claims (multiple search surfaces, structural evidence); medium-high on "deferral/abandonment" (plan document absent, zero post-research commits, competing priorities). Two gaps flagged by blue itself (live-source drift on Auto Dream status, unverified assumptions on qmd-as-substitute).

---

## Verified Core Claims

| Claim | Verification | Confidence |
|---|---|---|
| Plan document commit 32f13b2 exists and adds `plans/memory-architecture.md` | ✓ Verified: `git show 32f13b2:plans/memory-architecture.md` succeeds (290 lines) | HIGH |
| Plan document absent at HEAD (6f0f8bd) | ✓ Verified: `git show HEAD:plans/memory-architecture.md` returns fatal; `ls plans/` shows only `efficiency-phase.md`, `README.md` | HIGH |
| Zero implementation of `/dream`, `/remember`, `/ingest` commands in plugins/ | ✓ Verified: `grep -r "/dream\|/remember\|/ingest" plugins/` returns zero matches outside research runs | HIGH |
| Zero `knowledge/` directory in `.claude/` or project roots | ✓ Verified: `find . -type d -name "knowledge"` returns no results; `ls .claude/` contains only `rules/`, `projects/` | HIGH |
| Sleeper-service plugin is Phase-0 scaffold only | ✓ Verified: `ls plugins/sleeper-service/` shows only `plugin.json`, `README.md`; no `skills/`, `commands/`, `agents/` subdirs | HIGH |
| Research report §6.3 / §11 marks security gates as "blocking prerequisites" | ✓ Verified: Research report explicitly states "the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds" | HIGH |
| R4-1 (allowlist inversion) and R4-2 (git-ignore projection) marked "UNVERIFIED by red" | ✓ Verified: Research report §42 ("But the ceiling denied red the round to verify those fixes at the leaf node") and risk matrix row 51-52 explicitly label as "UNVERIFIED by red" | HIGH |

---

## L3-F1: Plan Branch Never Integrated Into Main

**Location:** "Evidence Chains / §H2 Memory-architecture build deferred"

**Challenged claim:** "Commit 32f13b2 (2026-07-04) added `plans/memory-architecture.md` (migrated from AgentOrange PR #3)."

**Finding:** The commit 32f13b2 exists and added the file, but it lives on an unmerged feature branch (`plans/memory-architecture`), not in the main branch history. Verification:
- `git merge-base --is-ancestor 32f13b2 6f0f8bd` returns FALSE
- `git branch --contains 32f13b2 -a` returns `plans/memory-architecture` (not merged)
- `git log --all --graph --decorate` shows 32f13b2 on a separate trunk diverging from Phase-0 bootstrap (0937317), never merged to main

**Implication of fix:**
- Blue's narrative ("proposal migrated to repo, then deleted") is misleading. The proposal was added to a feature branch that was never integrated into main.
- The correct narrative: "proposed but not adopted; the feature branch remained unmerged."
- This is a **deferral signal, not an abandonment signal**: the branch exists and could be revived, whereas deletion + removal from history signals a more final abandonment decision.

**Risk grading:**
- **Likelihood:** MED (common in fast-moving projects — branches de-prioritized and left unmerged)
- **Impact:** LOW-MED (does not change the fact that zero implementation landed in main; affects interpretation of reversibility, not current state)
- **Complexity:** LOW (metadata correction; no code change)

**Confidence:** HIGH (git history is authoritative; branch structure verifiable)

**Recommendation:** CLOSE-AS-REFRAME. Reword: "Proposal added on unmerged feature branch 2026-07-04; never integrated into main; zero subsequent commits on the branch since Phase-0 bootstrap." This is deferral, not deletion. Downstream teams cannot re-implement from this branch without reviving it — but the branch exists if the decision reverses.

---

## L3-F2: Blocking Security Prerequisites Abandoned Without Risk-Acceptance

**Location:** "Evidence Chains / §H1 Blocking security recommendations closed"

**Challenged claim:** "The research report (§11) explicitly states 'the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds past the differentiating sliver.'"

**Finding:** The research report's disposition is correct and explicitly marks these as prerequisites, not optional. However, post-research (2026-07-12 → 2026-07-17), these prerequisites were not adopted as implementation items:
- Zero commits touching `mit.1`, `ingest`, `taint`, `allowlist`
- No implementation branch or task created for R4-1/R4-2 fixes
- Competing priorities (efficiency-phase PRs #14–22) took engineering cycles
- No documented risk-acceptance or waiver for shipping without these gates

**Implication of risk:**
- The research identified a **security gate** (two ingest-edge gates + mit.1) as "blocking core" (risk matrix row 50: HIGH impact, MED likelihood, LOW-MED complexity to fix).
- The blocking-candidate fixes (R4-1, R4-2) are load-bearing pieces of the security invariant, explicitly marked UNVERIFIED.
- **If the memory-architecture (or any similar inbox/consolidation system) ships without these gates, the compromise/injection risk remains open and unmitigated.**

**Risk grading:**
- **Likelihood:** MED-HIGH (if a simplified memory/consolidation system ships, it may omit these gates as "out of scope")
- **Impact:** HIGH (poisoned memory can corrupt active directives across projects; precedent: CVE-class in-the-wild attacks on similar inboxes)
- **Complexity:** LOW-MED (gates are "parser change" per research; mit.1 is "zero separable cost" per research)

**Confidence:** HIGH on non-adoption (commit history, plan scope, zero implementation); MEDIUM-HIGH on risk (depends on whether a simplified memory system ships anyway)

**Recommendation:** FLAG AS OPEN. This is not a blue finding (blue correctly identified non-implementation); it's a system-level risk that the research's blocking prerequisites were deprioritized. Needs explicit risk-acceptance or a committed implementation plan with owner/deadline before any memory-consolidation system ships.

---

## L3-F3: Live-Source Drift — Auto Dream Status Unverified; Two-Writer Collision Never Addressed

**Location:** "Methodology & Confidence Notes"

**Challenged claim:** "MEDIUM (Auto Dream non-integration): ... the upstream status of Auto Dream itself (reportedly 'flag-gated, rolling out' at research time) is unverified here and could have changed."

**Finding:** Blue correctly identifies live-source drift as a risk pattern but does not resolve it. The research report (2026-07-12) noted Auto Dream as a mitigating factor for memory-architecture's two-writer collision risk (marked HIGH likelihood in risk matrix row 57). Post-research, Auto Dream integration was never attempted, and its current status is not verified.

**Implication of risk:**
- Research risk matrix row 57: "Agent-memory `memory:` row wrong / bidirectional write collision" — HIGH likelihood if Auto Dream does not consolidate.
- If Auto Dream later ships and is used in conjunction with memory-architecture (or a replacement memory system), the collision risk activates: MEMORY.md could be written simultaneously by an agent and by Auto Dream's native consolidation, corrupting state or losing data.
- **The research flagged this as needing resolution (§6.3: "Native Auto Dream two-writer collision on `MEMORY.md`"); no fix or investigation followed.**

**Risk grading:**
- **Likelihood:** MED-HIGH (Auto Dream may be in use now; unverified)
- **Impact:** HIGH (bidirectional write collision ≈ data loss under race; corrupt agent learning across projects)
- **Complexity:** MED (depends on Auto Dream's API; coordination may require changes to both systems)

**Confidence:** MEDIUM (absence of integration code is clear; upstream status unknowable without live verification)

**Recommendation:** OPEN BLOCKER. Before shipping any memory-consolidation system or before Auto Dream integration: verify Auto Dream's current status, test two-writer collision scenario empirically, and document the mitigation (locking, merge semantics, or version-pinning to pre-Auto-Dream Anthropic runtime).

---

## L3-F4: Project-Memory as Implicit Architectural Substitute — No Decision Record

**Location:** "Disposition Summary / Superseded ✗"

**Challenged claim:** "This is **not** the OKF-based global cross-project knowledge store proposed in memory-architecture; it is project-scoped, manually-maintained, and has no lifecycle/decay/promotion machinery. **Interpretation:** The project adopted a scaled-down memory approach rather than implementing the full architecture. **Confidence grading:** MEDIUM-HIGH on 'lightweight alternative adopted' (project-memory is a real alternate pattern, but its relationship to the deferred memory-architecture is not explicitly documented in the repo)."

**Finding:** Project-memory skill (prosthetic-conscience Phase 2) is a real artifact: four-artifact discipline (AGENTS.md, implementation_plan.md, task.md, walkthrough.md) per project in `projects/<name>/`. It is simpler and project-scoped, unlike memory-architecture's global typed-concept promotion ladder. However:
- No commit message or documentation states "adopting project-memory as substitute for memory-architecture"
- Project-memory and memory-architecture serve overlapping but distinct functions (project-local ephemeral state vs. global consolidated knowledge)
- **Risk:** Future work may treat them as orthogonal, leading to re-scoping or double-implementation of memory-architecture despite project-memory already filling the simpler use case

**Implication of risk:**
- **Rework risk:** A future planning cycle might prioritize "global memory-architecture build" without realizing project-memory already de-facto closed the simpler requirement.
- **Scope creep risk:** The gap between what project-memory delivers (manual per-project artifacts) and what memory-architecture promised (global + lifecycle + typed concepts) may attract new feature requests. Future scopes might not articulate the trade-off clearly.
- **Architectural coherence risk:** No explicit decision record means future teams cannot reason about the trade-off (lightweight vs. globally-discoverable; manual vs. automated lifecycle).

**Risk grading:**
- **Likelihood:** HIGH (common in rapid-iteration projects; scope decisions made implicitly by implementation choices)
- **Impact:** MED (causes rework cycles or scope bloat; not data/safety loss)
- **Complexity:** LOW (documentation / decision record only)

**Confidence:** HIGH on project-memory existence and function; MEDIUM on its status as an intentional substitute (inferred from adoption pattern, not from explicit documentation).

**Recommendation:** DOCUMENT-REQUIRED. Create a decision record (in `docs/` or a backlog item) explicitly stating: "Project-memory (Phase 2) was adopted as the memory-discipline for this project, replacing the memory-architecture proposal (2026-07-12) in scope. Trade-off: simpler, project-scoped, manual lifecycle; no global cross-project consolidation or typed-concept promotion. If global consolidation is needed in future, the memory-architecture branch remains available for revamp." Assign owner and flag as architectural decision for the next planning cycle.

---

## L3-F5: Qmd Recall vs. Memory-Architecture Consolidation — Conflation Risk

**Location:** "Disposition Summary / Implemented ✓ / Superseded ✗"

**Challenged claim:** "Qmd recall layer (PR #18): MCP-based BM25+semantic search over markdown corpus; orthogonal to memory-architecture. ... Qmd is a **DIFFERENT** recall mechanism — it provides retrieval search over markdown documents, not semantic dedup of memory concepts."

**Finding:** Blue correctly identifies qmd as a retrieval layer, not a consolidation layer. However, a downstream team reading "we shipped memory improvements in PR #18" might conflate the two:
- Qmd: **retrieval** (search, what's in the corpus?)
- Memory-architecture: **consolidation** (lifecycle, dedup, promotion, taint gates)
- Both address "memory" but solve different problems

**Implication of risk:**
- **Scope confusion:** A future feature request ("better memory reuse") might reference qmd as already addressing the need, when actually consolidation (memory-architecture) is the blocker.
- **Budget miscalculation:** Project planners might believe "we've invested in memory" (qmd shipped), missing that consolidation's lifecycle/decay machinery is still unbuilt and may be necessary for scale.
- **Integration risk:** If auto-dream consolidation ships, having qmd retrieval but no bespoke consolidation layer means qmd search returns both fresh and stale candidates indiscriminately (no freshness/confidence decay).

**Risk grading:**
- **Likelihood:** MED-HIGH (overlapping terminology; easy to conflate "search" and "memory consolidation")
- **Impact:** MED (scope creep or delayed architecture decisions; not data loss)
- **Complexity:** LOW (naming/documentation; qmd is functioning correctly)

**Confidence:** HIGH (qmd and memory-architecture are distinct; terminology overlap is real)

**Recommendation:** SCOPE-TRIAGE. Label qmd explicitly as a **retrieval/search layer** in all documentation and planning. Do not mention it as a substitute for or advancement of memory-architecture. If memory-architecture is re-scoped, ensure the consolidation layer is designed independently from qmd, with clear handoff points (qmd fetches; consolidation deduplicates/promotes; projections serve to clients).

---

## L3-F6: Verification File-Type Blindspot — Search Scopes Incomplete

**Location:** "Methodology & Confidence Notes / Gap-pattern pre-flight check"

**Challenged claim:** "Mitigated — this audit searched three file-type scopes (Go, TypeScript, Markdown) + directory structure + git history, not single-scope grep."

**Finding:** Three-file-type scope is better than single-scope; however, several implementation surfaces were not searched:
- **Configuration files:** `.mcp.json` (MCP server configuration), YAML/TOML deployment configs, `.claude/` rule files beyond `rules/`
- **Shell/bash scripts:** `/scripts/` directory, setup files, hook scripts (these can implement gates without source code)
- **Compiled artifacts / binaries:** pre-built tools, plugin bundles (unlikely but unverified)
- **Plugin manifest fields beyond commands/:** nested schema in plugin.json could define gate behaviors without CLI commands

**Implication of risk:**
- **LOW likelihood:** Security gates typically live in code (not config), but if a gate is implemented as a pre-commit hook script or in `.mcp.json` server config, it would be invisible to the grep on source code.
- **If ingest gates ARE implemented elsewhere:** Blue's "zero implementation" claim would regress from HIGH to MEDIUM confidence.

**Risk grading:**
- **Likelihood:** LOW-MED (gates usually in code; but MCP-server gates in config would be easy to miss)
- **Impact:** MED (if missed, the security-gate risk analysis weakens)
- **Complexity:** LOW (expand grep scope to include .mcp.json and /scripts/)

**Confidence:** MEDIUM (file-type scope is broad but not exhaustive)

**Recommendation:** EXPAND-SEARCH-SURFACE. Re-run grep with `--include="*.json"` (specifically for `.mcp.json` and hook configs) and search `/scripts/` for ingest/gate/taint-related bash functions. If zero results, confidence on "zero implementation" moves back to HIGH. If matches found, re-grade the gap and potentially close L3-F2 (blocking security) with evidence of partial mitigation.

---

## L3-F7: Research-Stage Fixes R4-1 / R4-2 Never Became Implementation Responsibility

**Location:** "Evidence Chains / §H1 Blocking security recommendations"

**Challenged claim:** "The R4-1 and R4-2 fixes were research-stage proposals never committed post-debate."

**Finding:** VERIFIED. However, the **transition from research-stage fix to implementation responsibility is missing**. The research report explicitly states these are unverified blocking candidates; blue correctly identified they were never committed. But the gap is not just that they're unimplemented — it's that:
1. They were never **formally assigned** to an implementation team
2. No issue/task was created tracking them
3. No decision record states "defer R4-1/R4-2 to Phase 2" or "accept risk, skip R4-1/R4-2"
4. The research's "UNVERIFIED" verdict was never addressed with a follow-up verification plan

**Implication of risk:**
- **Orphaned fixes:** These are concrete, scoped fixes (parser change, git-ignore rule) that could be implemented in 1–2 days. But they languish because no one owns them.
- **Debt accumulation:** Each research round that passes without addressing them increases the likelihood they'll be forgotten or re-discovered as new vulnerabilities.
- **Knowledge loss:** The research context (why allowlist inversion is necessary, why git-ignore is sufficient) may fade, requiring re-research.

**Risk grading:**
- **Likelihood:** HIGH (untracked fixes commonly fall through the cracks)
- **Impact:** HIGH (security fixes left unimplemented = active vulnerability)
- **Complexity:** LOW-MED (fixes are scoped; implementation straightforward if assigned)

**Confidence:** HIGH (absence of task/issue in tracking system is verifiable; git log confirms no commits)

**Recommendation:** OPEN BLOCKER — CREATE TASKS. For each of R4-1 and R4-2, create a GitHub issue (or equivalent) with:
- Research citation (link to memory-architecture report §15.1–15.2)
- Unverified status and required verification step (red must verify at leaf node after implementation)
- Acceptance criteria (allowlist parser correctly rejects laundering paths for R4-1; git-ignore projection verified in a fresh clone for R4-2)
- Assign owner and target date before any memory-consolidation system ships

---

## L3-F8: Measurement-Methodology Drift in "Zero Implementation" Claim

**Location:** "Methodology & Confidence Notes / Searches conducted"

**Challenged claim:** "Confidence grading: HIGH on zero-implementation claim (multiple independent search surfaces, grep on both code and prose, plugin directory structure verified)."

**Finding:** Blue's claim is HIGH confidence, but it rests on structural evidence (absence of directories, commands, files) which is strong, combined with grep evidence (absence of keywords) which is strong but **text-based only**. A potential gap: what if a component WAS implemented but under different terminology?

For example:
- "ingest" might be called "intake" or "import" elsewhere
- "taint" might be called "untrusted" or "origin" in a different system
- A "gate" might be implemented as a "filter" or "check"

**Implication of risk:**
- **Term-blindness:** If the implementation uses different vocabulary (e.g., "source-screening" instead of "injection-screening"), grep for "injection" would miss it.
- **Aliasing:** Multiple names for the same concept (e.g., `/dream` consolidation also called `/sync` or `/commit`) would require exhaustive synonymy search.

**Risk grading:**
- **Likelihood:** LOW-MED (the research's own terminology is specific; if implementers chose different terms, they likely didn't read the proposal)
- **Impact:** MED (if undetected, security-gate risk analysis weakens)
- **Complexity:** MED (requires semantic code search or manual review of plugin implementations)

**Confidence:** MEDIUM-HIGH (structural evidence is strong; keyword search is comprehensive but text-based)

**Recommendation:** SECONDARY-VERIFICATION. Manual code review of the main plugin implementations (frank-exchange-of-views, prosthetic-conscience, sleeper-service) for any gate/check/screening logic that *might* provide ingest safety, even if not explicitly labeled. This is a deep-dive; if zero findings, can close L3-F6/L3-F8 and raise "zero implementation" confidence to VERY-HIGH.

---

## Summary of Open Gaps & Risk-Acceptance Calls

| ID | Class | Status | Recommendation |
|---|---|---|---|
| L3-F1 | Metadata | Verified but misleading narrative | CLOSE-AS-REFRAME: "unmerged branch, not deleted" |
| L3-F2 | Security | OPEN BLOCKER | Risk-accept or create tasks for R4-1, R4-2, mit.1 before shipping any memory system |
| L3-F3 | Security / Live-source | OPEN BLOCKER | Verify Auto Dream status; test two-writer collision empirically |
| L3-F4 | Architectural coherence | OPEN | Document decision: project-memory as intentional substitute for memory-architecture |
| L3-F5 | Scope confusion | OPEN | Clarify qmd as retrieval-only; don't conflate with memory-consolidation |
| L3-F6 | Verification completeness | OPEN | Expand search to .mcp.json, /scripts/ to verify "zero implementation" claim |
| L3-F7 | Ownership / implementation | OPEN BLOCKER | Create tracking tasks for R4-1, R4-2 with verification acceptance criteria |
| L3-F8 | Terminology blindness | OPEN | Secondary manual review of plugin implementations for aliased gate/check logic |

---

## Verdict

**FAIL (CHANGES-REQUIRED).**

Blue's core finding is correct: zero implementation of memory-architecture recommendations occurred post-research. However, two blocking security gaps remain open without documented risk-acceptance:

1. **R4-1 & R4-2 (taint-boundary allowlist inversion, git-ignore projection)** are research-stage fixes marked UNVERIFIED by red, never committed, and never assigned as implementation responsibilities. These are load-bearing pieces of the security invariant. They must be either implemented+verified or explicitly risk-accepted.
2. **Auto Dream two-writer collision** (research risk matrix row 57, HIGH likelihood) was never investigated or mitigated post-research. If Auto Dream ships, this collision activates.

Additionally, L3-F6 (file-type search blindspot) and L3-F8 (terminology drift) warrant secondary verification before closing the "zero implementation" claim.

The frame issues (L3-F1, L3-F4, L3-F5) are scoping/documentation gaps that do not block ship but create future rework risk if unaddressed.

---

## Friction

None. The audit harness, protocol, and blue's report structure accommodated the dark-side lens cleanly.
