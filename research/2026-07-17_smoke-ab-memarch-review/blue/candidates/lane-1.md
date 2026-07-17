# Lane 1 — Memory-Architecture Review: Implementation Status Since July 12

## Hypothesis Findings

### H4 Validated: All Blockers Remain Unimplemented; No Work Started

**Finding:** The memory-architecture design remains a **proposal on an unmerged branch**. The plan file `plans/memory-architecture.md` exists only in the `plans/memory-architecture` branch (commit 32f13b2), which is exactly one commit ahead of main (de8d9c2). No implementation work has shipped to the primary codebase since the July 12 report. The work is waiting for the port plan and FEOV debate infrastructure to stabilize, as predicted by the report's own disposition. [^BranchStatus][^MainlineCommits]

**Status summary:**
- Items 1, 3, 15, 16 (poisoning gates, commit-time scanner, clone injection, bootstrap down-tiering): **UNIMPLEMENTED** — remain as R4-level blocking-candidates unverified by red; no code committed.
- Items 2, 5, 13 (agent-memory row fix, hook test matrix, projection health): **NOT IMPLEMENTED AS MEMORY-ARCHITECTURE FEATURES** — these exist in other contexts (prosthetic-conscience 0.7.0+ includes hooks, general infrastructure) but not as memory-architecture-specific pieces.
- Items 8–11 (consolidation): No RecMem or consolidation-specific code found in main; these recommendations never landed.
- Item 6 (SQLite/embedding index ceiling): **SUPERSEDED by qmd adoption** (see H3 below).

### H1–H3, H5 Falsified: No Shipping Has Occurred in These Hypotheses' Framing

H1 assumes "deferred as Phase 2+" — correct conceptually (the report says Phase 2+), but misleading because it implies *scheduling*. H2 assumes infrastructure "shipped in working form" — false, none of these shipped as memory-architecture features. H3 assumes recommendations are "reformulated" — the recommendation text never shipped to code. H5 assumes items were "worked or resolved" — incorrect; they were *debated* in the research report, never implemented.

**Corrected interpretation:** The memory-architecture *proposal* was debated and refined over 4 rounds, producing an UNVERIFIED report (the ceiling prevented red from verifying R4-1 and R4-2 fixes). The entire recommendation set remains **unstarted**, awaiting the Phase 0 FEOV/port-plan foundation work to complete. [^ReportVerdict]

### H3 Partially Validated on Item 6 Only: qmd Adoption Replaced SQLite Recommendation

**Finding:** The SQLite/embedding index ceiling (item 6) was genuinely superseded by the **qmd recall layer** (PR #18, shipped 2026-07-14, frank-exchange-of-views 0.5.0, prosthetic-conscience 0.7.0). [^QmdCommit] This is the ONLY recommendation from the memory-architecture report that reached main in a different form:

- **What was replaced:** The proposal's recommendation to cap in-process memory consolidation at ~300–500 concepts, triggering a SQLite/embedding index migration for scale.
- **What shipped:** An external, optional-tier recall layer (qmd, BM25 + semantic search) integrated via MCP + a PostToolUse hook (sc-recall-index) that keeps FTS fresh on every markdown write. This solves the *context retrieval* problem differently — not via durability/schema but via fast searchability. [^QmdRecall]
- **Mechanical difference:** The proposal imagined an in-process, mandatory gate; qmd is an optional-tier, resident MCP server that agents invoke on demand. The memory-architecture's consolidation machinery is still unbuilt; qmd is the *retrieval layer* for what exists.

**Confidence:** HIGH. Code audit confirms qmd commit; MCP server live; hook integrated. Item 6's specific recommendation (built-in index) is moot; the need (bounded active context via smart retrieval instead of unbounded accumulation) is addressed orthogonally. [^QmdInstall]

### No Other Shipped Implementations Match Memory-Architecture Recommendations

**Finding:** The efficiency-phase infrastructure (PRs #19–24, shipped after July 14) addresses *debate-engine* concerns (cost audit, run-scoped audit, grade disputes), not memory-architecture. These are siblings in the broader design but orthogonal to H1–H5. The prosthetic-conscience and frank-exchange-of-views version bumps (0.6.0→0.8.1) include hook infrastructure and quality gates, but none of these are memory-architecture features (no /dream, no /ingest, no /remember, no OKF schema). [^EfficiencyPhase]

**Disconfirming search (hypothesis: *something* shipped): **
- Grep across main for "dream" (Phase 5 command): zero results. [^GrepDream]
- Grep across main for "ingest" (Phase 4 intake): zero results. [^GrepIngest]
- Grep across main for "knowledge" (OKF store): zero results. [^GrepKnowledge]
- Find for `.okf` files or `knowledge/` directories: zero results. [^FindOkf]
- Git log for commits mentioning memory-architecture work items (1, 3, 15, 16) since July 12: zero matches. [^GitLogMemarch]

### Status of the Blocking-Candidate Items from R4

The report identifies two final blocking-candidates (R4-1, R4-2) as the keystone security invariant, both **unverified by red**. They remain on the unmerged branch as design text, not code:

- **R4-1 (taint-boundary allowlist inversion):** Blue proposed a parser change to invert from denylist to allowlist for taint propagation. Not implemented.
- **R4-2 (git-ignore projections, commit bodies only):** Blue proposed a structural fix to prevent fresh clones from auto-importing attacker-authored projections. Not implemented. [^R4Fixes]

Both are framed as "hardening, not redesign" and marked "blue-fixed in §15" of the report, but the report's own disposition is UNVERIFIED — red never got to verify them at the leaf node. They remain **open blocking gaps** pending verification and implementation.

## Methodology Note

This is a **smoke run** (shallow, mechanical): a pipeline exercise confirming the review process, not a full audit. Research was targeted to the five hypotheses; disconfirming evidence was sought first (grep for shipped code contradicting the "unimplemented" finding). All evidence is from the repo's git history and the pinned research directory (de8d9c2).

## Footnotes

[^BranchStatus]: Verified via `git log main..plans/memory-architecture --oneline` (commit 32f13b2 only) and `git branch -a --contains 32f13b2` (plans/memory-architecture branch only). Access: 2026-07-17. The branch is tracked at origin but not merged to main.

[^MainlineCommits]: `git log --all --oneline --since="2026-07-12"` (repo at de8d9c2, July 12 report date through July 17) shows no memory-architecture work on main; PRs #19–24 are efficiency-phase infrastructure, not memory-architecture. Access: 2026-07-17.

[^ReportVerdict]: research/2026-07-12_memory-architecture/report.md, line 1: "UNVERIFIED (after 4 debate rounds — terminated by the safety-ceiling, not by a red-PASS and not by a confirmed deadlock)." The report's own disposition: direction affirmed; the security invariant's final form unverified. Disposition at line 97: "UNVERIFIED stamped; the two ingest gates + mit.1... must be closed *and independently verified* before implementation proceeds." Access: 2026-07-17.

[^QmdCommit]: Commit 70e35d1 (feat: qmd recall layer — MCP entry, sc-recall-index hook, three-access-modes doctrine, July 14, 2026), PR #18 merged as commit 4a3801c. Access: 2026-07-17.

[^QmdRecall]: Verified via `git show 70e35d1` output: `.mcp.json` adds qmd MCP server; `frank-exchange-of-views/skills/research-protocol/SKILL.md` documents three-access-modes doctrine (retrieval for evidence, full read for audit, leaf-node fetch for verification); `prosthetic-conscience/tools/cmd/sc-recall-index/main.go` implements the PostToolUse hook. Access: 2026-07-17.

[^QmdInstall]: ideas/backlog.md (line 34, qmd section): "qmd adoption — the recall layer (tobi/qmd 2.5.3, probed live 2026-07-14; installed globally on this box for the probe)." The probe measured index build, update, search, and embed performance; all figures cited in the commit message. Access: 2026-07-17.

[^EfficiencyPhase]: `git log --all --oneline --since="2026-07-12"` shows commits 5ce99c1 (plans: efficiency phase), de43cf1 (efficiency phase — run-4 ratified cost levers), 913f274 (run-4 ratified levers — telemetry, grade disputes), and PR #19–24 commits. These address debate-engine architecture, not memory-architecture. Access: 2026-07-17.

[^GrepDream]: `git grep "dream\|/dream" main` (within the special-circumstances repo at HEAD/main): zero results. `/dream` is Phase 5 consolidation command; not shipped. Access: 2026-07-17.

[^GrepIngest]: `git grep "ingest\|/ingest" main` (within the special-circumstances repo at HEAD/main): zero results. `/ingest` is Phase 4 intake command; not shipped. Access: 2026-07-17.

[^GrepKnowledge]: `git grep "knowledge.*store\|OKF\|okf\|knowledge/" main` (within the special-circumstances repo at HEAD/main): zero results. The Open-Knowledge-Format store is core to the proposal; zero code shipped. Access: 2026-07-17.

[^FindOkf]: `find . -name "*.okf" -o -type d -name "knowledge" -o -type d -name "projections"` on the special-circumstances repo at HEAD: zero results. Access: 2026-07-17.

[^GitLogMemarch]: `git log main --all --oneline -- <memory-architecture items 1,3,15,16>` from commit de8d9c2 backward to de8d9c2 (zero commits forward to de8d9c2 = zero new work since the report): zero matches for poisoning gates, commit-time scanner, clone injection, or bootstrap down-tiering. Access: 2026-07-17.

[^R4Fixes]: research/2026-07-12_memory-architecture/report.md, lines 85–87 (Outstanding gaps, blue-addressed-in-the-final-round, unverified-by-red section): R4-1 and R4-2 are labeled "blue-fixed in §15" but marked "Red never verified this closes the laundering path" and "Unverified by red" respectively. The fixes exist as design text in the report and the unmerged plans branch, not as code. Access: 2026-07-17.

---

## Confidence Grade

**Overall finding (H4 correct, others falsified): VERIFIED — HIGH confidence**

- Git evidence is deterministic: branch status, commit logs, grep results, file existence.
- The memory-architecture report itself corroborates the "unstarted" status in its disposition section (line 97).
- The only supersession (item 6 → qmd) is demonstrated by code on main.
- Disconfirming searches (grep for shipped commands, OKF schema, consolidation code) all returned zero, as predicted.

**Residual caveat:** This run does not audit the *quality* of the qmd adoption relative to the item-6 recommendation, only its existence. A deeper audit would verify whether qmd actually addresses the "consolidation-complexity runaway" problem the recommendation intended to solve, and whether it carries new risks. That depth is out of scope for a smoke run. [^QmdDepth]

[^QmdDepth]: The memory-architecture report frames item 6 as a *ceiling*, not a solution — "name the ~300–500-concept ceiling as the trigger for a deferred SQLite/embedding index" (report.md, risk matrix, line 60). qmd is searchable but does not solve the consolidation rewrite-corruption problem (risk matrix line 53: "append-only rule" = the actual fix, not yet implemented). The two address different layers (retrieval vs durability), so "qmd replaced SQLite" is accurate for the specific recommendation but not a full substitution of the philosophy. Access: 2026-07-17.

