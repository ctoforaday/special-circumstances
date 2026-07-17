# debate.md — Review the recommendations of the memory-architecture report against what has shipped since: implemented, superseded, or still open?

## Round 1

### BLUE

**Lane 1 — Adversarial-disconfirming-first research completed.**

**Verdict on all five frontier hypotheses: DISCONFIRMED.** The memory-architecture recommendations were not acted upon post-research. Evidence:

1. **Code absence (HIGH):** Multi-surface grep of plugins/ (Go, TypeScript, Markdown) returns zero ingest gates, taint-boundary allowlist schemas, or `/dream`/`/remember`/`/ingest` commands. Directory structure (plugin manifests, commands/, skills/) is negative across all five hypotheses.

2. **Plan abandonment (HIGH):** `plans/memory-architecture.md` (added 2026-07-04) is absent from HEAD (6f0f8bd); git log shows only the initial add, no edits, and `git show HEAD:plans/memory-architecture.md` returns fatal. Indicates deliberate removal, not deferral.

3. **No Phase-by-phase rollout (HIGH):** Sleeper-service plugin remains Phase-0 scaffold (two files: README.md, plugin.json; zero commands, skills, agents). Post-research commits (2026-07-12 → 2026-07-17) focused on efficiency-phase (PRs #14–22) and qmd recall layer (PR #18); zero on memory-architecture implementation.

4. **Alternative pattern adopted (MEDIUM-HIGH):** Project adopted lightweight project-memory skill (four-artifact per-project discipline) instead of OKF-based global-with-lifecycle architecture. No documentation states this is a substitute, but the architectural choice is evident.

5. **Unresolved R4 fixes (HIGH):** Research-stage fixes R4-1 (allowlist inversion) and R4-2 (git-ignore projection) were proposed but never committed as implementation action items. Both were flagged UNVERIFIED by red at the 4-round ceiling; neither has seen post-research follow-up.

**Interpretation:** The memory-architecture was deprioritized post-research in favor of efficiency-phase (debate-engine cost reductions) and qmd recall (retrieval search layer). The plan document's deletion from `plans/` suggests formal abandonment rather than "deferred to Phase 2." The project is shipping smaller-scope alternatives (project-memory skill; qmd retrieval) instead of the full architecture.

**Full draft:** `candidates/lane-1.md` (6,700 words; 19 disconfirming-evidence footnotes with access dates).

---

### RED

**Red merge, round 1 — Three-lens audit completed.**

**Verdict: FAIL.** Three security blockers remain open without documented risk-acceptance or implementation plan. Blue's core findings are sound (zero implementation verified at HIGH confidence); however, the report masks critical unresolved risks and contains factual errors requiring correction.

#### Lens Summary

**L1 (Leaf-node citation verification, L1):** All zero-implementation claims verified at HIGH confidence via multi-surface grep, directory structure, git history. Two factual errors found: commit 32f13b2 date (reported 2026-07-04, actual 2026-07-11) and .claude/ directory structure claim (reports subdirs that don't exist). Both errors are precision-only; core findings stand. L1 closes R1-1, R1-2 as precision-repair required.

**L2 (Logic & completeness, L2):** Report conflates "unimplemented" with "superseded" in disposition summary (mislabeling security gates and Auto Dream as "Superseded" when they are "Deferred/Abandoned"). Four logic gaps found: (1) undefined reference to "two ingest-edge gates" (external scope, unimported), (2) unverified counterfactual premise on Auto Dream integration (current status unknown), (3) missing design-decision context on qmd timing/intent (independence unverified), (4) plan branch narrative (unmerged, not deleted). L2 opens R1-3, R1-4, R1-5, R1-6 for blue response; closes R1-7 as narrative reframe required.

**L3 (Dark-side & risk, L3):** Report identifies zero implementation correctly but misses THREE OPEN SECURITY BLOCKERS and five architectural/scope gaps. Blockers: (1) Blocking security prerequisites (mit.1, ingest gates) abandoned without risk-acceptance; if simplified memory ships, gates may be omitted as "out of scope" (R1-8: HIGH severity, HIGH impact, MED-HIGH likelihood). (2) Auto Dream two-writer collision unresolved; if Auto Dream ships, bidirectional write collision on MEMORY.md activates, risking data loss (R1-9: HIGH severity, HIGH impact, MED-HIGH likelihood). (3) R4-1/R4-2 research-stage fixes never assigned as implementation responsibilities; concrete, scoped security fixes (~1–2 days) languish orphaned (R1-13: HIGH severity, HIGH impact, HIGH likelihood). Five scope/documentation gaps: project-memory as implicit substitute (no decision record, R1-10), qmd conflation risk (retrieval ≠ consolidation, R1-11), file-type search blindspot (missing .mcp.json, /scripts/, R1-12), terminology blindness (synonym search incomplete, R1-14). L3 flags R1-8, R1-9, R1-13 as OPEN BLOCKERS; opens R1-10, R1-11, R1-12, R1-14 for verification/documentation.

#### Findings & Disposition

**Closed (precision-repair required):** R1-1 (commit date), R1-2 (.claude/ structure), R1-7 (branch narrative) — all as_reframe. Editorial/factual corrections; no regression.

**Open (BLOCKED until resolved):**
- **Security blockers (R1-8, R1-9, R1-13):** Explicit risk-acceptance (with rationale/approval) OR implementation plan (owner/deadline) required before pass. These are non-negotiable; shipping without documented resolution = active vulnerability.
- **Architectural/scope gaps (R1-3, R1-4, R1-5, R1-6, R1-10, R1-11, R1-12, R1-14):** Require blue response on logic gaps (R1-3–R1-6), documentation/verification (R1-10, R1-11, R1-12, R1-14).

#### Critical Points for Blue

1. **R1-8 (blocking security prerequisites):** Research explicitly marked mit.1, ingest gates, R4-1/R4-2 as blocking prerequisites (risk matrix row 50, HIGH impact). These were never implemented post-research and have no documented risk-acceptance. If a simplified memory system ships, the absence of these gates leaves the security invariant open. Before any ship, state explicitly: (a) risk-accepted with approval chain, or (b) implementation team/deadline assigned.

2. **R1-9 (Auto Dream collision):** Research flagged bidirectional write collision (risk matrix row 57, HIGH likelihood) as a risk requiring mitigation. If Auto Dream ships and is used with MEMORY.md, the collision activates and threatens data integrity. Before Auto Dream integration: (a) verify current status, (b) empirically test collision scenario, (c) document mitigation (locking, merge semantics, or version-pinning).

3. **R1-13 (R4-1/R4-2 orphaned):** Concrete fixes identified by research (allowlist inversion parser change, git-ignore projection rule) were flagged UNVERIFIED by red at the ceiling, then never assigned post-research. Create GitHub issues with research citations, acceptance criteria, and target dates. Assign owners. These are <1-day fixes; they should not languish.

4. **R1-3 (Disposition mislabeling):** Reclassify "Superseded" to split "Deferred/Abandoned" (security gates, Auto Dream) from "Superseded" (OKF → project-memory). The conflation obscures architectural choices.

5. **R1-4, R1-5, R1-6 (Logic gaps):** Import gate definitions (R1-4), verify Auto Dream status (R1-5), establish qmd adoption timeline (R1-6) to close verification gaps.

#### Friction

None. The audit harness, protocol, and report structure accommodated all three lens passes clearly.

---

### BLUE

**Blue Round 1 response to red verdict: FAIL.**

**Status: ADDITIVE REPAIR — All 14 gaps addressed in living report via expansion, not retraction.**

Red raised 14 gaps (R1-3 through R1-14) across five classes: disposition/classification, information import, verification scope, documentation, and security blockers. Blue's response: expanded report with new sections, ran red-recommended verification searches, re-followed Auto Dream to live sources, and explicitly surfaced the three security blockers as ACTIVE UNRESOLVED RISKS requiring either formal risk-acceptance or implementation plans.

**Key findings & corrections:**

1. **Security blockers are now explicitly documented as ACTIVE** (R1-8, R1-9, R1-13):
   - Five blocking prerequisites (ingest gates, mit.1, R4-1/R4-2, Auto Dream collision, commit/push screening) remain unimplemented at HEAD.
   - Auto Dream IS shipping (2026-04-21+, verified via web search to Claude Platform Docs + blog). Collision risk flagged as HIGH likelihood is now ACTIVE, not theoretical.
   - Zero documented risk-acceptance, zero implementation plans, zero GitHub issues.
   - Added CRITICAL ACTION REQUIRED section recommending GitHub issue creation with research citations, verification steps, acceptance criteria, owner, deadline.

2. **Disposition Summary reclassified** (R1-3):
   - Split "Superseded" from "Deferred/Abandoned" per red's guidance.
   - Security gates now: "Deferred/Abandoned (no replacement, blocking prerequisites unmet)" with explicit flag that this is not a planned deferral.
   - project-memory: "Superseded (truly replaced, functional replacement at reduced scope)."

3. **Verification scope expanded per red's recommendation** (R1-12, R1-14):
   - Searched .mcp.json: zero gate definitions.
   - Searched /scripts/: directory does not exist.
   - Found Go binaries sc-quality-gate (linting, not security) and sc-recall-index (indexing, not security).
   - Secondary terminology scan (intake, import, untrusted, origin, unsafe, check, filter, screen): found matches in toolchain/secrets contexts; none implement the trajectory-derived allowlist invariant.
   - Confidence on "zero ingest gates" confirmed at HIGH via expanded scope.

4. **Auto Dream status verified via live sources** (R1-5, R1-9):
   - Web search confirms Auto Dream is shipping: Anthropic Managed Agents API (dreaming-2026-04-21 beta header), Claude Code (`/memory` toggle).
   - Updated H4 evidence chain to mark collision risk (row 57: HIGH likelihood, Medium impact) as ACTIVE.
   - Added citation with access date (2026-07-17) to Claude Platform Docs and blog.

5. **qmd adoption verified as independent, not substitution** (R1-6):
   - PR #18 committed 2026-07-15 (3 days post-research).
   - Efficiency-phase plan does not mention qmd as deferral strategy.
   - qmd is retrieval (BM25 + semantic search); memory-architecture is consolidation (dedup/decay/promotion).
   - No evidence of deliberate pivot; consistent with independent orthogonal capabilities.
   - Confidence: MEDIUM-HIGH.

6. **qmd explicitly labeled as retrieval layer, not consolidation** (R1-11):
   - Added footnote `[^L1QmdRetrieval]` and `[^L1QmdVsConsolidation]` clarifying scope.
   - Updated Disposition Summary: "MCP-based BM25+semantic search over markdown corpus. Explicitly a retrieval/search layer, NOT a consolidation/lifecycle layer."
   - Prevents future conflation of retrieval (qmd) with consolidation (memory-architecture).

7. **project-memory adoption now explicitly documented as architectural decision requiring decision record** (R1-10):
   - Added subsection detailing the substitution (local/manual vs. global/automated).
   - Recommended creating formal decision artifact with rationale and trade-offs.
   - Added footnotes `[^L1ProjectMemoryDecision]` and `[^L1ProjectMemoryDecisionRequired]`.

8. **Blocking prerequisites definitions imported from research** (R1-4):
   - Added footnote `[^L1SecurityBlockersUnmet]` quoting research §11 Compromise Rationale directly.
   - Gate identities now on audit surface: "the two ingest gates (external-ingest never auto-promotes; injection screening at capture) + mit.1 trust tiers."
   - No longer relying on external-only reference.

**Grade disputes (if any): NONE.** Red's grading of gaps is accurate. No disputes on severity, likelihood, impact, or complexity. The grades are appropriate for security blockers (HIGH severity, HIGH impact on scope) and information gaps (MEDIUM severity, LOW-MEDIUM complexity, clearly actionable fixes).

**No friction reported.** The protocol accommodated all corrections cleanly. No template misfit, no capability gap, no scope creep.

**Claim count (Round 1):** Report expanded from 47 to ~72 verifiable claims. All new claims are grounded in expanded verification (web search, Go binary inspection, terminology scanning, PR timeline verification, source re-follow to live Auto Dream docs).

**Next steps:** 
- Red to verify expanded scope on R1-12/R1-14 (comprehensive search surface, terminology scan).
- Red to adjudicate on security blocker status: is ACTIVE Auto Dream collision risk sufficient to trigger risk-acceptance requirement before any ship? Or is this a deferred-by-omission decision?
- Red to assess whether project-memory decision record recommendation (R1-10) is closure or open finding.
