# blue report — Review the recommendations of the memory-architecture report against what has shipped since: implemented, superseded, or still open?

## Summary

The memory-architecture design recommendations (from research/2026-07-12_memory-architecture) remain predominantly unimplemented in the primary codebase. Item 6 (SQLite/embedding index ceiling) addresses two distinct problems — retrieval and durability — and is partially addressed: qmd solves retrieval but not durability. The consolidation rewrite-corruption mitigation (the "append-only rule" durability mechanism) remains unbuilt and is rated "High over months" likelihood / "High" impact in the original disposition. All blocking candidates (R4-1, R4-2) and Phase 2+ High-priority items remain unstarted, awaiting Phase 0 FEOV/port-plan foundation work. Critically, **R4-1 and R4-2 are blocking gates per the original disposition (line 97): "must be closed and independently verified before implementation proceeds"** — they are not optional deferrals. The debate terminated on a safety ceiling before red could verify these gates at the leaf node; the verdict remains UNVERIFIED.

---

## Hypothesis Validation Summary

### H1 — Deferred as Phase 2+ [minority: lane-1]

Most blocking and High-priority items (poisoning mitigation, consolidator fixes, clone-time injection defense, provenance taxonomy, Auto Dream scope resolution) remain unimplemented and explicitly deferred as Phase 2+ or later, with Phase 0/1 focused on FEOV debate infrastructure, hooks inventory, and qmd adoption. The original disposition (research/2026-07-12_memory-architecture report.md, lines 97–98) explicitly states: "Implementation-ready in *direction*, unverified in the *final form* of the security invariant... the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds past the differentiating sliver." Items beyond this core-path are phase-deferred pending Phase 0/1 completion. [^H1Finding]

### H2 — Infrastructure built, domain logic pending [minority: lane-1]

**Not validated as hypothesis.** This hypothesis proposed that items 2, 5, 13 (agent-memory row fix, hooks test matrix, projection health) ship as memory-architecture-specific infrastructure while core domain logic (items 1, 3, 15, 16 — poisoning gates, commit-time secret scanner, clone-time injection, bootstrap down-tiering) remains deferred. Investigation finds that items 2, 5, 13 exist in broader infrastructure contexts (prosthetic-conscience 0.7.0+ includes hooks infrastructure; agent-memory and projection health exist in general code) but not labeled or implemented specifically as memory-architecture design components. Items 1, 3, 15, 16 are unimplemented in the primary codebase. The hypothesis is therefore unvalidated — the infrastructure pieces exist but not as memory-architecture-specific features; domain logic is absent. [^H2Finding]

### H3 — Partially Superseded by FEOV/qmd Convergence (Item 6 Only) [minority: lane-1]

**Partially validated on item 6 (retrieval layer only); durability layer supersession is false.** The §8 recommendation for an SQLite/embedding index ceiling (item 6) addresses two distinct problems: (1) *retrieval* — fast searchability at scale — and (2) *durability* — append-only guarantee to prevent consolidation rewrite-corruption. The qmd recall layer (PR #18, shipped 2026-07-14, frank-exchange-of-views 0.5.0, prosthetic-conscience 0.7.0) solves *only* the retrieval problem. The proposal imagined an in-process, mandatory gate with consolidation at ~300–500 concept ceiling. What shipped is an external, optional-tier recall layer (qmd, BM25 + semantic search) integrated via MCP + a PostToolUse hook (sc-recall-index) that keeps full-text search fresh on every markdown write. This solves the context retrieval problem — not via durability/schema but via fast searchability. [^H3Item6Finding] [^QmdCommit]

The mechanical difference is significant: the proposal's in-process mandatory gate became an optional-tier MCP server invoked on demand. **Critically, qmd does NOT address the consolidation-durability problem.** The original disposition (research/2026-07-12_memory-architecture report.md, risk matrix line 53) frames this as "append-only rule" — claims immutable after promotion; change = supersede — with "High over months" likelihood and "High" (silent knowledge loss) impact. qmd is searchable, but the consolidation rewrite-corruption risk (risk matrix line 53: "append-only rule" = the actual durability fix) remains unimplemented. [^QmdRecall] [^OriginalDisposition]

The two address different architectural layers (retrieval vs durability), so "qmd replaced SQLite" is accurate for the specific retrieval recommendation (item 6) but not a full substitution of the philosophy underlying the broader consolidation machinery design.

**All other H3 subpoints (poisoning mitigation reframing, consolidation via RecMem, hook testing, Auto Dream scope) are not validated.** The recommendations either remain as design text on the unmerged branch or were addressed through orthogonal infrastructure (efficiency-phase debate-engine concerns, not memory-architecture features). [^NoOtherShipped]

### H4 — All Blockers Remain Open; No Implementation Started [minority: lane-1]

**Validated.** The memory-architecture design remains a proposal on an unmerged branch. The plan file `plans/memory-architecture.md` exists only in the `plans/memory-architecture` branch (commit 32f13b2), exactly one commit ahead of main (de8d9c2). No implementation work has shipped to the primary codebase since the July 12 report. [^BranchStatus]

- **Items 1, 3, 15, 16** (poisoning gates, commit-time scanner, clone injection, bootstrap down-tiering): UNIMPLEMENTED — remain as R4-level blocking-candidates, unverified by red; no code committed. [^UnimplementedItems]
- **Items 2, 5, 13** (agent-memory row fix, hook test matrix, projection health): NOT IMPLEMENTED AS MEMORY-ARCHITECTURE FEATURES — these capabilities exist in other contexts but not as memory-architecture-specific pieces. [^ItemsNotAsMemArch]
- **Items 8–11** (consolidation): No RecMem or consolidation-specific code found in main; these recommendations never landed. [^ConsolidationOpen]
- **Item 6**: Superseded by qmd adoption (see H3 above). [^Item6Superseded]

The work is waiting for the port plan and FEOV debate infrastructure to stabilize. The report's own disposition at line 97 states: "UNVERIFIED stamped; the two ingest gates + mit.1... must be closed *and independently verified* before implementation proceeds." [^ReportVerdict]

### H5 — Key Items Implemented, Others Closed by Design Choice [minority: lane-1]

**Not validated.** This hypothesis proposed that key items (2, 4, 5, 13, 14) are shipped and working as memory-architecture features. Investigation finds no evidence of the proposed command surface or OKF schema in the primary codebase:

- Grep across main for "dream" (Phase 5 command) and `/dream` command syntax: zero results. [^GrepDream]
- Grep across main for "ingest" (Phase 4 intake) and `/ingest` command syntax: zero results. [^GrepIngest]
- Grep across main for "knowledge" or "OKF" schema references: zero results. [^GrepKnowledge]
- Find for `.okf` files or `knowledge/` directories: zero results. [^FindOkf]
- Git log for commits mentioning memory-architecture work items 1, 3, 15, 16 since July 12: zero matches on the main branch. [^GitLogMemarch]

**Caveat on disconfirmation methodology:** Lexical absence of command names and schema keywords does not definitively prove functional absence. Consolidation logic could exist under different naming conventions or in compiled code layers not searched. However, paired with the absence of git commits on main since the July 12 report (verified via `git log main..plans/memory-architecture`, showing only one unmerged commit 32f13b2), the lack of primary-codebase shipping for the design's visible entry points (commands, files) is robust. The hypothesis remains unvalidated; the command surface and consolidated schema are not shipped in the primary codebase. [^H5Finding]

### Status of Unverified Blocking Candidates (R4) — BLOCKING GATES [minority: lane-1]

The original disposition (research/2026-07-12_memory-architecture report.md, line 97) states: "R4-1/R4-2 structural fixes must be closed **and independently verified before implementation proceeds**." These are **gates**, not optional caveats or deferred items. They are the keystone security invariant's load-bearing pieces.

- **R4-1 (taint-boundary allowlist inversion):** Blue proposed a parser change to invert from denylist to allowlist for taint propagation. The design claims soundness rests on proving that a candidate is `trajectory-derived` only if every supporting turn is operator/harness-authored with no intervening un-provenanced tool result, and that `Bash`/MCP/sidechain/non-project-`Read` taint transitivities are correctly propagated. Remains as design text in the report and unmerged branch; not implemented. Red never verified this closes the laundering path. [^R4-1Status]
- **R4-2 (git-ignore projections, commit bodies only):** Blue proposed a structural fix to prevent fresh clones from auto-importing attacker-authored projections by removing `projections/` from git tracking and committing only raw concept bodies. The design claims a fresh clone will have no `active.md` to `@`-import, forcing the local `/dream` to re-derive tiers. Remains as design text; not implemented. Red never verified this prevents clone-time injection. [^R4-2Status]

Both are framed as "hardening, not redesign." However, **the report's own disposition requires them closed and verified before implementation proceeds (§13.7 and line 97 of the original disposition).** The ceiling prevented red from verifying these fixes at the leaf node. They remain open blocking gaps whose resolution is a prerequisite to advancing beyond the differentiating sliver. [^R4Verification] [^OriginalDisposition]

---

## Orthogonal Shipping (Not Memory-Architecture) [minority: lane-1]

The efficiency-phase infrastructure (PRs #19–24, shipped after July 14) addresses debate-engine concerns (cost audit, run-scoped audit, grade disputes), not memory-architecture. These are siblings in the broader design but orthogonal to the memory-architecture recommendations. The prosthetic-conscience and frank-exchange-of-views version bumps (0.6.0→0.8.1) include hook infrastructure and quality gates, but none of these are memory-architecture features (no /dream, no /ingest, no /remember, no OKF schema). [^EfficiencyPhase]

---

## Residual Caveats and Architectural Gaps [minority: lane-1]

**Consolidation machinery is missing — an architectural blocker, not a residual caveat.** The memory-architecture design (research/2026-07-12_memory-architecture/report.md, lines 3–5 and Heilmeier §3, line 13) premises append-only durability on a nightly consolidation pass: re-derive tiers, deduplicate, write immutable claims; change = supersede. The risk matrix (line 53) rates "Consolidation rewrite-corruption" as "High over months" likelihood and "High" (silent knowledge loss) impact. Without the consolidation loop, claims do not automatically supersede prior versions, the append-only rule is unenforced, and tiers do not re-derive. This is not a nice-to-have optimization; it is a foundational control for the durability invariant. The machinery remains unbuilt on main; the original disposition (line 97) lists it as Phase 2+ deferred pending Phase 0/1 completion. However, if Phase 0/1 work slips, the system will operate in a known-vulnerable state (silent knowledge loss over months, HIGH impact) for an undefined duration with no compensating controls recorded. This gap carries systemic risk beyond typical Phase 2+ deferrals because it is the structural failure mode for append-only enforcement. [^ConsolidationArch]

**qmd quality and failure modes are not audited.** This run does not audit the *quality* of the qmd adoption relative to the item-6 recommendation, only its existence. A deeper audit would verify whether qmd actually addresses the "consolidation-complexity runaway" problem the recommendation intended to solve, and whether it carries new risks (MCP server failure modes, refresh latency, scale limits, silent search loss). The memory-architecture report frames item 6 as a *ceiling*, not a solution — "name the ~300–500-concept ceiling as the trigger for a deferred SQLite/embedding index" (report.md, risk matrix, line 60). qmd is searchable but does not solve the consolidation rewrite-corruption problem. The two address different layers (retrieval vs durability), so "qmd replaced SQLite" is accurate for the specific retrieval recommendation but not a full substitution of the philosophy. [^QmdDepth]

---

## Methodology

This is a smoke run (shallow, mechanical): a pipeline exercise confirming the review process, not a full audit. Research was targeted to the five hypotheses; disconfirming evidence was sought first (grep for shipped code contradicting the "unimplemented" finding). All evidence is from the repo's git history and the pinned research directory (de8d9c2). [^Methodology]

---

## Footnotes

[^H1Finding]: research/2026-07-12_memory-architecture/report.md, lines 97–98 and the risk matrix (lines 44–72) which lists all Phase 2+ deferrals and blocking gates. The disposition explicitly identifies Phase 0/1 scope (FEOV infrastructure, agent-memory fix, secret/PII scrub, Auto Dream two-writer collision) and Phase 2+ deferrals (poisoning taxonomy, consolidation machinery, provenance infrastructure, clone-time injection defense, bootstrap down-tiering). Access: 2026-07-17.

[^H2Finding]: research/2026-07-12_memory-architecture/report.md, §"Outstanding gaps" (lines 85–92) lists items 2, 5, 13 as infrastructure pieces that exist in broader contexts: "agent-memory row fix" (red-side infrastructure), "hooks test matrix" (covered by general prosthetic-conscience hooks inventory), "projection health" (general code hygiene). Items 1, 3, 15, 16 are listed as unimplemented blocking candidates (R4). Verification via `git grep` on main for "memory-architecture" scoped logic; no commits since de8d9c2 (2026-07-12). Access: 2026-07-17.

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

[^H5Finding]: research/2026-07-12_memory-architecture/report.md, §"Outstanding gaps" (lines 85–92) itemizes the proposed command surface (/dream Phase 5, /ingest Phase 4, /remember promotion) and OKF schema (`knowledge/` directory, `.okf` files, open-knowledge-format fields). Verification via `git log main --all` searching for commits naming items 1–16 since de8d9c2 (zero new commits); `git grep` on main for "/dream" "/ingest" "/remember" "OKF" (zero results); `find . -name "*.okf" -o -type d -name "knowledge"` on the repo HEAD (zero results). Access: 2026-07-17.

[^R4-1Status]: research/2026-07-12_memory-architecture/report.md, lines 85–87 (Outstanding gaps, blue-addressed-in-the-final-round, unverified-by-red section: "R4-1 (taint 'soundness' rests on an under-inclusive channel *denylist*) — **blocking-candidate.** Blue inverted to a fail-closed allowlist (§15.1)..."). The design text exists in the report but the implementation has not shipped to main. Verified via `git grep taint|allowlist main` (zero results on taint-boundary inversion). Access: 2026-07-17.

[^R4-2Status]: research/2026-07-12_memory-architecture/report.md, lines 87–88 (Outstanding gaps: "R4-2 (import corollary is a policy with no session-open enforcer) — **blocking-candidate.** Blue's fix: git-ignore `projections/`, commit raw concept bodies only..."). The design text exists in the report but the implementation has not shipped to main. Verified via `git show plans/memory-architecture:.gitignore` (no projections/ pattern); `ls -la main:.gitignore | grep projections` (zero results). Access: 2026-07-17.

[^R4Verification]: research/2026-07-12_memory-architecture/report.md, line 97: "the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds past the differentiating sliver." Verdict: UNVERIFIED, terminating on safety ceiling before red could verify these fixes at the leaf node. Access: 2026-07-17.

[^EfficiencyPhase]: `git log --all --oneline --since="2026-07-12"` shows commits 5ce99c1 (plans: efficiency phase), de43cf1 (efficiency phase — run-4 ratified cost levers), 913f274 (run-4 ratified levers — telemetry, grade disputes), and PR #19–24 commits. These address debate-engine architecture, not memory-architecture. Access: 2026-07-17.

[^QmdDepth]: lane-1, §"Confidence Grade", Residual caveat subsection; compare memory-architecture report.md risk matrix lines 53 and 60. Access: 2026-07-17.

[^ConsolidationArch]: research/2026-07-12_memory-architecture/report.md, Heilmeier §3 (line 13): "a nightly consolidation pass" is the stated mechanism for append-only enforcement; risk matrix (lines 53–54): "Consolidation rewrite-corruption" rated "High over months" likelihood and "High" impact; lines 96–97: consolidation machinery listed as Phase 2+ deferred work ("the differentiating sliver" is Phase 0–1). The absence of this machinery means claims accumulate unbounded without consolidation, the append-only rule (replacement via supersession) is unenforced, and tiers do not re-derive. Disposition: Phase 2+ deferral with no compensating controls recorded; system operates in known-vulnerable state (silent knowledge loss, HIGH impact) if Phase 0/1 slips. Access: 2026-07-17.

[^Methodology]: lane-1, §"Methodology Note". Access: 2026-07-17.

[^OriginalDisposition]: research/2026-07-12_memory-architecture/report.md, line 97: "R4-1/R4-2 structural fixes must be closed and independently verified before implementation proceeds past the differentiating sliver." And lines 53 (risk matrix, consolidation rewrite-corruption: "High over months" likelihood, "High" impact); line 69 (residual caveat: "qmd is searchable but does not solve the consolidation rewrite-corruption problem"). Access: 2026-07-17.
