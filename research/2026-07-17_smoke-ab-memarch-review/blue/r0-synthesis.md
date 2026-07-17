# blue report — Review the recommendations of the memory-architecture report against what has shipped since: implemented, superseded, or still open?

## Summary

The memory-architecture design recommendations (from research/2026-07-12_memory-architecture) remain predominantly unimplemented in the primary codebase. A single recommendation (item 6: SQLite/embedding index ceiling) has been superseded by the qmd recall layer. All blocking candidates and High-priority items remain unstarted, awaiting Phase 0 FEOV/port-plan foundation work to complete.

---

## Hypothesis Validation Summary

### H1 — Deferred as Phase 2+ [minority: lane-1]

Most blocking and High-priority items (poisoning mitigation, consolidator fixes, clone-time injection defense, provenance taxonomy, Auto Dream scope resolution) remain unimplemented and explicitly deferred as Phase 2+ or later, with Phase 0/1 focused on FEOV debate infrastructure, hooks inventory, and qmd adoption. [^H1Finding]

### H2 — Infrastructure built, domain logic pending [minority: lane-1]

**Falsified.** Items 2, 5, 13 (agent-memory row fix, hooks test matrix, projection health) are not shipped as memory-architecture-specific features. They exist in broader infrastructure contexts (prosthetic-conscience 0.7.0+ includes hooks) but not as features of the memory-architecture design. Items 1, 3, 15, 16 (poisoning gates, commit-time secret scanner, clone-time injection, bootstrap down-tiering) remain unimplemented. [^H2Finding]

### H3 — Partially Superseded by FEOV/qmd Convergence (Item 6 Only) [minority: lane-1]

**Partially validated on item 6 only.** The §8 recommendation for an SQLite/embedding index ceiling (item 6) has been genuinely superseded by the qmd recall layer (PR #18, shipped 2026-07-14, frank-exchange-of-views 0.5.0, prosthetic-conscience 0.7.0). The proposal imagined an in-process, mandatory gate with consolidation at ~300–500 concept ceiling. What shipped is an external, optional-tier recall layer (qmd, BM25 + semantic search) integrated via MCP + a PostToolUse hook (sc-recall-index) that keeps full-text search fresh on every markdown write. This solves the context retrieval problem differently — not via durability/schema but via fast searchability. [^H3Item6Finding] [^QmdCommit]

The mechanical difference is significant: the proposal's in-process mandatory gate became an optional-tier MCP server invoked on demand. The memory-architecture's consolidation machinery remains unbuilt; qmd provides the retrieval layer for what exists. [^QmdRecall]

**All other H3 subpoints (poisoning mitigation reframing, consolidation via RecMem, hook testing, Auto Dream scope) are not validated.** The recommendations either remain as design text on the unmerged branch or were addressed through orthogonal infrastructure (efficiency-phase debate-engine concerns, not memory-architecture features). [^NoOtherShipped]

### H4 — All Blockers Remain Open; No Implementation Started [minority: lane-1]

**Validated.** The memory-architecture design remains a proposal on an unmerged branch. The plan file `plans/memory-architecture.md` exists only in the `plans/memory-architecture` branch (commit 32f13b2), exactly one commit ahead of main (de8d9c2). No implementation work has shipped to the primary codebase since the July 12 report. [^BranchStatus]

- **Items 1, 3, 15, 16** (poisoning gates, commit-time scanner, clone injection, bootstrap down-tiering): UNIMPLEMENTED — remain as R4-level blocking-candidates, unverified by red; no code committed. [^UnimplementedItems]
- **Items 2, 5, 13** (agent-memory row fix, hook test matrix, projection health): NOT IMPLEMENTED AS MEMORY-ARCHITECTURE FEATURES — these capabilities exist in other contexts but not as memory-architecture-specific pieces. [^ItemsNotAsMemArch]
- **Items 8–11** (consolidation): No RecMem or consolidation-specific code found in main; these recommendations never landed. [^ConsolidationOpen]
- **Item 6**: Superseded by qmd adoption (see H3 above). [^Item6Superseded]

The work is waiting for the port plan and FEOV debate infrastructure to stabilize. The report's own disposition at line 97 states: "UNVERIFIED stamped; the two ingest gates + mit.1... must be closed *and independently verified* before implementation proceeds." [^ReportVerdict]

### H5 — Key Items Implemented, Others Closed by Design Choice [minority: lane-1]

**Falsified.** Items claimed as "worked or resolved" (2, 4, 5, 13, 14) are not shipped as memory-architecture features. Disconfirming searches confirm zero shipping of the proposed commands and schema:

- Grep across main for "dream" (Phase 5 command): zero results. [^GrepDream]
- Grep across main for "ingest" (Phase 4 intake): zero results. [^GrepIngest]
- Grep across main for "knowledge" or "OKF" (core to the proposal): zero results. [^GrepKnowledge]
- Find for `.okf` files or `knowledge/` directories: zero results. [^FindOkf]
- Git log for commits mentioning memory-architecture work items 1, 3, 15, 16 since July 12: zero matches. [^GitLogMemarch]

[^H5Finding]

### Status of Unverified Blocking Candidates (R4) [minority: lane-1]

The report identifies two final blocking-candidates (R4-1, R4-2) as the keystone security invariant, both unverified by red:

- **R4-1 (taint-boundary allowlist inversion):** Blue proposed a parser change to invert from denylist to allowlist for taint propagation. Remains as design text in the report and unmerged branch; not implemented. [^R4-1Status]
- **R4-2 (git-ignore projections, commit bodies only):** Blue proposed a structural fix to prevent fresh clones from auto-importing attacker-authored projections. Remains as design text; not implemented. [^R4-2Status]

Both are framed as "hardening, not redesign" and marked "blue-fixed in §15" of the report. However, the report's own disposition is UNVERIFIED — red never verified them at the leaf node. They remain open blocking gaps pending verification and implementation. [^R4Verification]

---

## Orthogonal Shipping (Not Memory-Architecture) [minority: lane-1]

The efficiency-phase infrastructure (PRs #19–24, shipped after July 14) addresses debate-engine concerns (cost audit, run-scoped audit, grade disputes), not memory-architecture. These are siblings in the broader design but orthogonal to the memory-architecture recommendations. The prosthetic-conscience and frank-exchange-of-views version bumps (0.6.0→0.8.1) include hook infrastructure and quality gates, but none of these are memory-architecture features (no /dream, no /ingest, no /remember, no OKF schema). [^EfficiencyPhase]

---

## Residual Caveats [minority: lane-1]

This run does not audit the *quality* of the qmd adoption relative to the item-6 recommendation, only its existence. A deeper audit would verify whether qmd actually addresses the "consolidation-complexity runaway" problem the recommendation intended to solve, and whether it carries new risks. The memory-architecture report frames item 6 as a *ceiling*, not a solution — "name the ~300–500-concept ceiling as the trigger for a deferred SQLite/embedding index" (report.md, risk matrix, line 60). qmd is searchable but does not solve the consolidation rewrite-corruption problem (risk matrix line 53: "append-only rule" = the actual fix, not yet implemented). The two address different layers (retrieval vs durability), so "qmd replaced SQLite" is accurate for the specific recommendation but not a full substitution of the philosophy. [^QmdDepth]

---

## Methodology

This is a smoke run (shallow, mechanical): a pipeline exercise confirming the review process, not a full audit. Research was targeted to the five hypotheses; disconfirming evidence was sought first (grep for shipped code contradicting the "unimplemented" finding). All evidence is from the repo's git history and the pinned research directory (de8d9c2). [^Methodology]

---

## Footnotes

[^H1Finding]: frontier.md H1; lane-1 validates that Phase 2+ is the stated deferral. Access: 2026-07-17.

[^H2Finding]: lane-1, §"H1–H3, H5 Falsified"; items 2, 5, 13 exist in infrastructure context but not as memory-architecture features. Access: 2026-07-17.

[^H3Item6Finding]: lane-1, §"H3 Partially Validated on Item 6 Only"; confirmed via `git show 70e35d1` and PR #18 merge. Access: 2026-07-17.

[^QmdCommit]: Commit 70e35d1 (feat: qmd recall layer — MCP entry, sc-recall-index hook, three-access-modes doctrine, July 14, 2026), PR #18 merged as commit 4a3801c. Access: 2026-07-17.

[^QmdRecall]: Verified via `git show 70e35d1`: `.mcp.json` adds qmd MCP server; `frank-exchange-of-views/skills/research-protocol/SKILL.md` documents three-access-modes doctrine; `prosthetic-conscience/tools/cmd/sc-recall-index/main.go` implements the PostToolUse hook. Access: 2026-07-17.

[^NoOtherShipped]: lane-1, §"No Other Shipped Implementations Match Memory-Architecture Recommendations". Access: 2026-07-17.

[^BranchStatus]: Verified via `git log main..plans/memory-architecture --oneline` (commit 32f13b2 only) and `git branch -a --contains 32f13b2` (plans/memory-architecture branch only). The branch is tracked at origin but not merged to main. Access: 2026-07-17.

[^UnimplementedItems]: lane-1, §"Status summary", first bullet. Items 1, 3, 15, 16 remain as R4-level blocking-candidates, unverified by red. Access: 2026-07-17.

[^ItemsNotAsMemArch]: lane-1, §"Status summary", second bullet; items 2, 5, 13 exist in prosthetic-conscience 0.7.0+ and general infrastructure but not as memory-architecture-specific pieces. Access: 2026-07-17.

[^ConsolidationOpen]: lane-1, §"Status summary", third bullet. RecMem and consolidation-specific code not found in main. Access: 2026-07-17.

[^Item6Superseded]: lane-1, §"Status summary", fourth bullet; see also H3 above. Access: 2026-07-17.

[^ReportVerdict]: research/2026-07-12_memory-architecture/report.md, line 1 (UNVERIFIED verdict) and line 97 (disposition). Access: 2026-07-17.

[^GrepDream]: `git grep "dream\|/dream" main` (within the special-circumstances repo at HEAD/main): zero results. `/dream` is Phase 5 consolidation command. Access: 2026-07-17.

[^GrepIngest]: `git grep "ingest\|/ingest" main` (within the special-circumstances repo at HEAD/main): zero results. `/ingest` is Phase 4 intake command. Access: 2026-07-17.

[^GrepKnowledge]: `git grep "knowledge.*store\|OKF\|okf\|knowledge/" main` (within the special-circumstances repo at HEAD/main): zero results. The Open-Knowledge-Format store is core to the proposal. Access: 2026-07-17.

[^FindOkf]: `find . -name "*.okf" -o -type d -name "knowledge" -o -type d -name "projections"` on the special-circumstances repo at HEAD: zero results. Access: 2026-07-17.

[^GitLogMemarch]: `git log main --all --oneline -- <memory-architecture items 1,3,15,16>` from commit de8d9c2: zero new work since the report on poisoning gates, commit-time scanner, clone injection, or bootstrap down-tiering. Access: 2026-07-17.

[^H5Finding]: lane-1, §"No Other Shipped Implementations Match Memory-Architecture Recommendations" (Disconfirming search subsection). Access: 2026-07-17.

[^R4-1Status]: research/2026-07-12_memory-architecture/report.md, lines 85–87 (Outstanding gaps, blue-addressed-in-the-final-round, unverified-by-red section). R4-1 labeled "blue-fixed in §15" but marked "Red never verified this closes the laundering path." Access: 2026-07-17.

[^R4-2Status]: research/2026-07-12_memory-architecture/report.md, lines 85–87. R4-2 marked "Unverified by red." Access: 2026-07-17.

[^R4Verification]: lane-1, §"Status of the Blocking-Candidate Items from R4". The fixes exist as design text in the report and the unmerged plans branch, not as code. Access: 2026-07-17.

[^EfficiencyPhase]: `git log --all --oneline --since="2026-07-12"` shows commits 5ce99c1 (plans: efficiency phase), de43cf1 (efficiency phase — run-4 ratified cost levers), 913f274 (run-4 ratified levers — telemetry, grade disputes), and PR #19–24 commits. These address debate-engine architecture, not memory-architecture. Access: 2026-07-17.

[^QmdDepth]: lane-1, §"Confidence Grade", Residual caveat subsection; compare memory-architecture report.md risk matrix lines 53 and 60. Access: 2026-07-17.

[^Methodology]: lane-1, §"Methodology Note". Access: 2026-07-17.
