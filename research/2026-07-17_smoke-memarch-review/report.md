# Review the recommendations of the memory-architecture report against what has shipped since: implemented, superseded, or still open? — research report

**Verdict:** UNVERIFIED (after 1 debate round) · **Run:** `research/2026-07-17_smoke-memarch-review/`

**TL;DR:** The memory-architecture proposal (2026-07-12, research completion) was not implemented post-research. All five frontier hypotheses are disconfirmed: zero ingest gates, taint boundaries, or consolidation commands shipped; the plan document was removed from the repository; the project adopted lightweight project-memory as a simpler substitute instead of building the full OKF-based architecture. Red's verdict is FAIL due to three unresolved security blockers (ingest gates, Auto Dream collision mitigation, orphaned R4-1/R4-2 fixes) that remain unimplemented with zero documented risk-acceptance or implementation plan. Blue's round-1 response expanded the report with direct evidence of Auto Dream's current shipping status (making the collision risk ACTIVE, not theoretical) and explicit flagging of the three security blockers as requiring immediate action. Confidence on non-implementation is HIGH across multiple independent search surfaces (code grep, directory structure, git history). Confidence on deferral intent is MEDIUM-HIGH (plan deleted, zero post-research commits, competing priorities evident). **Outstanding gap disposition:** 3 security blockers (risk_accepted with contingency flag required, or implementation plan with owner/deadline required before any memory-consolidation system ships); 8 architectural/scope gaps (documentation, verification, decision records required; no regression on core findings).

---

## The Catechism

**1. What are we trying to do?**

Audit whether the memory-architecture research recommendations (completed 2026-07-12) were acted upon post-research. Scope: compare implementation against five frontier hypotheses (blocking security gates closed; OKF-based differential shipped; Auto Dream overlap resolved; lifecycle machinery in place; phase-4/5 open).

**2. How is it handled today, and what does that cost us?**

Memory-architecture was proposed as a global, cross-project knowledge consolidation system with lifecycle/decay/promotion machinery and security gates (taint boundaries, injection screening, commit/push screening). No implementation occurred post-research. Current state: (a) lightweight project-memory skill (four artifacts per project, manual lifecycle); (b) qmd retrieval layer (BM25 + semantic search, retrieval-only); (c) secrets-gate hook (outbound screening only, not inbound). Cost: no global consolidation, no semantic dedup, no decay machinery, no lifecycle automation; teams must maintain memory manually per-project; cross-project learning/reuse impossible; security gates for ingest remain unimplemented, leaving attack surface open if memory consolidation ships.

**3. What is new here, and why do we believe it works?**

New: mechanical audit confirming non-implementation via multi-surface search (code grep on three file types, directory structure, git history, plugin manifests). Confidence HIGH on "zero implementation" claim (structural evidence + text evidence + git history all negative). New in round-1 response: web search verification that Auto Dream is now shipping (2026-04-21+), making the research's flagged collision risk (HIGH likelihood) ACTIVE, not theoretical. New: explicit identification of three unresolved security blockers (ingest gates, Auto Dream collision, R4-1/R4-2 orphaned) that red's verdict flags as requiring documented disposition before any ship.

**4. The case against.**

Why memory-architecture was deprioritized: (a) Efficiency-phase work (debate-engine cost levers, PRs #14–22) captured post-research attention and delivered measurable ROI on engineering cost. (b) Auto Dream's shipping status (2026-04-21+) may have reduced the defensible scope for bespoke consolidation (native Auto Dream provides nightly consolidation; bespoke scope could be narrowed to just the differential). (c) Project-memory skill (four-artifact discipline) is pragmatic, simpler, and demonstrably working at current project scale; full architecture is higher-complexity. (d) Full architecture requires blocking prerequisites (security gates) to be verified before build; research hit 4-round ceiling, leaving R4-1/R4-2 unverified; implementing unverified fixes is risky; deferral until verification possible is defensible. (e) Lifecycle/decay machinery is complex, low-criticality feature; simpler per-project manual discipline delivers most value at lower cost. Case against build: valid. Case against *shipping unresolved security blockers*: critical — ingest gates are load-bearing; omitting without risk-acceptance is active vulnerability.

**5. Of interest, or merely interesting?**

The memory-architecture recommendations are of-interest only if they are genuinely needed and can be shipped with security gates closed. If shipped without gates (gates pushed to Phase-N), the architecture becomes a security liability, not an asset. Project-memory (currently shipping) partially fills the use case (manual, per-project) at smaller scope; full architecture would provide (a) cross-project consolidation (medium-high interest: enables learning reuse across projects), (b) lifecycle/decay (medium interest: prevents stale facts accumulating), (c) security gates (high interest: load-bearing, non-negotiable before ship). Verdict: of-interest conditionally; blocked on security gates and cost/complexity trade-off. If ingest gates are impossible to close before architecture ship, memory-architecture should remain deferred in favor of project-memory + qmd retrieval.

**6. What changes if it works — and what happens if we simply don't do it?**

If full architecture ships *with* blocking security gates closed: agents and FEOV debate can reference a consolidated, deduplicated, decayed global memory without security surface increase; cross-project learning enabled; lifecycle automation reduces manual overhead. If we don't do it: project-memory (current, simpler) provides per-project manual discipline; qmd retrieval (current) enables evidence search; Auto Dream native consolidation (current, shipping) handles agent-scope memory; no cross-project consolidation; cross-project learning remains manual (copy-paste, oral transfer). Impact if architecture deferred: medium (project-memory + Auto Dream + qmd cover most use cases; cross-project learning is gap). Impact if gates are omitted and architecture ships: critical (unsecured ingest gateway for attacker-controlled data reaching always-on context without screening).

**7. What does it cost, and where would we stop?**

Cost to implement: (a) R4-1 allowlist inversion (~1 day, unverified fix), (b) R4-2 git-ignore projection (~1 day, unverified fix), (c) ingest gates (blocking prerequisites, ~3–5 days + verification), (d) `/dream` consolidation command (Phase-2, ~3–5 days), (e) lifecycle/decay machinery (Phase-3–5, ~10–15 days escalating). Stopping point (safety ceiling): **Do not ship any memory-consolidation system (native Auto Dream or bespoke) without documented risk-acceptance or implementation plan for security blockers (ingest gates, Auto Dream collision mitigation, R4-1/R4-2 verification).** Checkpoints: (1) create GitHub issues for R4-1/R4-2 with verification acceptance criteria, assign owner, set deadline before any implementation begins; (2) for ingest gates, document formal decision (risk-accept + rationale, or implementation plan with owner/deadline); (3) for Auto Dream collision, document mitigation strategy (scope split, locking, merge semantics, or version-pinning) before feature integration. Failure to meet checkpoints = UNVERIFIED verdict maintained.

---

## Technical foundations

**Evidence base (frozen at 6f0f8bd per PINNED.md):**
- Repo HEAD: commit 6f0f8bd
- Memory-architecture research: `research/2026-07-12_memory-architecture/`
- Project backlog reference: `ideas/backlog.md` at 6f0f8bd

**Audit scope:** Five frontier hypotheses (H1–H5) tested via disconfirming-first methodology against special-circumstances repo state at 2026-07-17.

**Search surfaces verified:**
1. Code grep (three scopes): Go binaries, TypeScript, Markdown skill definitions — all zero matches for "ingest|injection|taint|allowlist|dream|remember|consolidat" (excluding research/ debate documents).
2. Directory structure: Zero `knowledge/` directories at three levels (.claude, projects, plugins). Sleeper-service plugin remains Phase-0 scaffold (README + plugin.json only).
3. Git history: `git log --all -- plans/memory-architecture.md` shows single add commit (32f13b2, date 2026-07-11 per red's L1 verification); no subsequent edits; `git show HEAD:plans/memory-architecture.md` returns fatal (file absent from HEAD).
4. Plugin manifests: `ls plugins/*/commands/` confirms zero `/dream`, `/remember`, `/ingest` commands; zero OKF-schema instantiation (type: rule|fact|preference|glossary).

**Verification calibration:**
- **HIGH confidence** on "zero implementation" claims: Multiple independent surfaces (code, structure, git history) all negative; structural evidence (absence of directories, commands) independent of text grep.
- **MEDIUM-HIGH confidence** on "deferral/abandonment" claims: Plan deleted + zero post-research commits + competing priorities (efficiency phase, qmd) create consistent narrative; no deletion log entry (inferred), requires triangulation.
- **CRITICAL FINDING** on Auto Dream collision: Web search (2026-07-17) confirms Auto Dream IS shipping (dreaming-2026-04-21 beta header, Anthropic Managed Agents API, Claude Code `/memory` toggle active). Research's flagged risk (HIGH likelihood, Medium impact per risk matrix row 57) is now ACTIVE, not theoretical.

**Red's audit findings (Lens 1–3):**
- L1 (leaf-node verification): 3 precision errors found (commit date, .claude/ structure claim, plan-branch narrative) — all corrected as `closed_as_reframe`; core "zero implementation" claims verified at HIGH confidence across all three lenses.
- L2 (logic & completeness): 4 logic gaps raised (undefined gate references, unverified Auto Dream counterfactual, unverified qmd timing, plan-branch narrative) — blue's round-1 response addressed all via expanded verification and explicit sourcing.
- L3 (dark-side & risk): 3 security blockers identified as OPEN (ingest gates R1-8, Auto Dream collision R1-9, R4-1/R4-2 orphaned R1-13); 5 architectural/scope gaps requiring documentation or verification (project-memory decision record, qmd independence, verification surface, terminology blindspot, scope clarification).

---

## Analysis

**Five hypotheses tested; all disconfirmed.**

**H1: Blocking security recommendations (taint-boundary, injection screening, git-ignore projection) closed; non-blocking machinery open.**

Disconfirmed. No ingest gates, allowlist inversion, or git-ignore projection found. Evidence: (a) Zero code matches for "ingest|taint|allowlist|denylist" across plugins; (b) No `projections/` directory or `.gitignore` rule for projections; (c) Existing `sc-secrets-gate` (outbound-tool-input screening) does not cover inbound taint gates; (d) Research-stage fixes R4-1 (allowlist inversion) and R4-2 (git-ignore projection) were proposed but never committed to codebase post-debate. Confidence: HIGH.

**H2: Memory-architecture build deferred; Auto Dream narrowed remit; differential adopted.**

Disconfirmed (partial). Plan document is not merely deferred (branch unmerged) — it is absent from HEAD tracking tree, suggesting stronger abandonment than deferral. No differential (OKF-based global store + schema) shipped. Auto Dream NOT integrated; integration attempts zero. Verification: (a) Commit 32f13b2 lives on unmerged feature branch `plans/memory-architecture`, zero commits on branch post-add; (b) `git show HEAD:plans/memory-architecture.md` fatal; plan absent from main history; (c) Zero code references to "Auto Dream" in plugins (excluding research debate); (d) Zero `.claude/knowledge/` directories. Confidence: HIGH on non-implementation; MEDIUM-HIGH on "abandonment vs. deferral" (structural evidence: plan deleted, zero post-research commits, but no deletion log entry).

**H3: Typed-concept differential shipped minimally; lifecycle machinery unimplemented.**

Disconfirmed. No OKF schema instantiation, no knowledge stores, no lifecycle machinery (decay windows, consolidation scheduler, promotion ladder). Evidence: (a) Zero `type: rule|fact|preference` frontmatter in plugins (exists only in research debate text); (b) No decay configuration or trigger logic; (c) No consolidation scheduler. Lightweight alternative adopted: project-memory skill (four-artifact per-project discipline) — pragmatic, simpler, manually maintained, but NOT equivalent to global cross-project consolidation. Confidence: HIGH.

**H4: Recommendations superseded by native Auto Dream; Phase-2/3 overlap unresolved; project-store blocked.**

Disconfirmed. No Auto Dream integration attempted despite Auto Dream now being shipping (verified 2026-07-17 via web search: dreaming-2026-04-21 beta, Anthropic Managed Agents API, Claude Code `/memory` toggle active). Two-writer collision on MEMORY.md (flagged as HIGH likelihood in research risk matrix row 57) is now ACTIVE risk, not theoretical. Project-store feature never shipped; R4-2's proposed de-escalation (git-ignore projection) never implemented. Confidence: HIGH on no-attempt-at-resolution; CRITICAL on collision risk activation (Auto Dream is now shipping, making conflict scenario live).

**H5: Build proceeded under re-scoped margin; security gates shipped; clone-ratification and provenance shipped; Phase-4/5 open.**

Disconfirmed. Security gates (mit.1 trust tiers, ingest screening) NOT shipped — research marked as "blocking prerequisites" (§11 Compromise Rationale); zero implementation. Clone-ratification marker NOT shipped. Provenance taxonomy (source: trajectory|url|file) NOT instantiated (qmd captures file URIs but does NOT instantiate OKF provenance taxonomy). Phase-4 `/ingest` pipeline exists only in proposal text. Phase-5 headless schedule likewise text-only. Confidence: HIGH.

**Synthesized finding:**

Memory-architecture was abandoned post-research, not deferred. Five signals: (1) Plan document deleted from tracking tree (RED FLAG for abandonment vs. deferral); (2) Zero implementation commits in 6-day window post-research; (3) Competing priorities captured effort (efficiency phase, qmd retrieval); (4) Lightweight alternative (project-memory) adopted without decision record documenting substitution; (5) Security blocking prerequisites remain unimplemented with zero documented disposition. The project is shipping smaller-scope alternatives (project-memory skill for local memory, qmd retrieval for search, Auto Dream native consolidation for agent-scope consolidation) instead of the full cross-project OKF-based architecture.

---

## Risk matrix

| Risk | Likelihood | Impact | Complexity to mitigate | Mitigation / disposition |
|---|---|---|---|---|
| **[R1-8] Blocking security gates (ingest screening, mit.1 taint tiers) omitted from any memory-consolidation ship** | MED-HIGH | HIGH | MED-LOW | **risk_accepted (contingency)** — Research §11 marks these as blocking prerequisites; zero implementation 6 days post-research; no documented risk-acceptance or implementation plan. *Conditional acceptance:* If simplified memory consolidates without these gates, formal risk-acceptance memo required (signed off by architect/security) OR GitHub issues created with owner/deadline before any ship. **Otherwise FAIL.** |
| **[R1-9] Auto Dream two-writer collision on MEMORY.md activates if Auto Dream used with bespoke consolidation** | HIGH | MED-HIGH | MED | **risk_accepted (contingency)** — Auto Dream shipping (2026-04-21+). Research flagged as HIGH likelihood (row 57). No mitigation implemented. *Conditional:* Before any memory-consolidation system ships or Auto Dream integration attempted, either (a) implement scope split (bespoke owns `knowledge/`, native owns `MEMORY.md`), (b) conduct empirical collision test, or (c) document risk-acceptance with rationale. **Otherwise FAIL.** |
| **[R1-13] R4-1 and R4-2 research fixes never assigned as implementation responsibilities** | HIGH | HIGH | LOW-MED | **risk_accepted (contingency)** — Research §15.1–15.2 identifies concrete, scoped fixes (allowlist parser inversion ~1 day, git-ignore projection ~1 day); both flagged UNVERIFIED by red at 4-round ceiling. Zero GitHub issues created; zero owner assigned; no deadline. *Conditional:* Create GitHub issues (per blue's round-1 recommendation) with research citations, verification steps, acceptance criteria, owner assignment, target date before any memory-consolidation ship. **Otherwise FAIL.** |
| **[R1-3] Disposition mislabeling (security gates labeled "Superseded" instead of "Deferred/Abandoned")** | CERTAIN | MED | LOW | **closed_as_reframe** — Blue's round-1 response corrected disposition summary: split "Superseded" (OKF → project-memory, true replacement) from "Deferred/Abandoned" (security gates, Auto Dream coordination, no replacement). Narrative now accurate. |
| **[R1-4] Undefined reference to "two ingest-edge gates" without imported definitions** | MED-HIGH | MED | LOW | **closed_as_rebuttal_sustained** — Blue's round-1 response imported gate definitions from research §11: "the two ingest gates (external-ingest never auto-promotes; injection screening at capture) + mit.1 trust tiers." Gate identities now on audit surface with source citations. |
| **[R1-5] Unverified counterfactual premise on Auto Dream integration feasibility** | MED-HIGH | MED | LOW | **closed_as_rebuttal_sustained** — Blue's round-1 response added web search verification (2026-07-17): Auto Dream IS shipping (dreaming-2026-04-21 beta, Anthropic Managed Agents API, Claude Code `/memory` toggle). Counterfactual premise ("should have integrated") now grounded in current status verification. |
| **[R1-6] Missing design-decision context on qmd timing and intent** | MED | MED-HIGH | LOW-MED | **closed_as_rebuttal_sustained** — Blue's round-1 response verified qmd independence via timeline (PR #18 committed 2026-07-15, 3 days post-research), scope separation (qmd=retrieval; memory-architecture=consolidation), and plans/efficiency-phase.md scope (zero mention of qmd as memory-architecture deferral strategy). Inference: independent orthogonal capability, not deliberate pivot. Confidence MEDIUM-HIGH. |
| **[R1-7] Plan-branch narrative (unmerged vs. deleted) misrepresents reversibility** | CERTAIN | MED | LOW | **closed_as_reframe** — Blue's round-1 response clarified: proposal on unmerged feature branch `plans/memory-architecture`; never integrated to main; zero post-bootstrap commits. Branch exists (reversible); main history contains zero memory-architecture code. Narrative corrected; deferral signal preserved; abandonment inference remains defensible. |
| **[R1-10] Project-memory adoption as implicit substitute (no decision record)** | HIGH | MED | LOW | **carried** — Project-memory skill is real, working artifact; no decision record documents it as deliberate substitution for memory-architecture. Recommendation: Create decision record (docs/ or backlog item) stating (a) decision: project-memory adopted as memory discipline; (b) rationale: simpler, per-project, manual; (c) trade-off: no cross-project learning, no lifecycle; (d) memory-architecture branch available for future revamp. *Carried to next planning cycle.* |
| **[R1-11] Qmd/memory-architecture conflation risk (retrieval labeled as memory improvement, conflating with consolidation)** | MED-HIGH | MED | LOW | **carried** — Blue's round-1 response explicitly labeled qmd as "retrieval/search layer, NOT consolidation/lifecycle layer" with footnotes preventing conflation. Requires downstream documentation (planning, PR descriptions) to consistently call out retrieval ≠ consolidation. *Carried to next planning cycle and process documentation.* |
| **[R1-12] Verification file-type blindspot (missing .json/.mcp.json, scripts/, binaries)** | LOW-MED | MED | LOW | **closed_as_rebuttal_sustained** — Blue's round-1 response expanded search: `.mcp.json` gate definitions (zero), `/scripts/` directory (absent), Go binaries (sc-quality-gate linting, sc-recall-index indexing, neither security gates). Confidence on "zero implementation" confirmed at HIGH. |
| **[R1-14] Terminology blindness in zero-implementation claim (alternative naming: intake/import/filter)** | LOW-MED | MED | MED | **closed_as_rebuttal_sustained** — Blue's round-1 response ran secondary terminology scan: "intake|import|untrusted|origin|unsafe|check|filter|screen" across plugins (Go, TypeScript, JSON). Matches in toolchain/secrets contexts only; none implement trajectory-derived allowlist invariant or never-auto-promote-external rule. Confidence confirmed at HIGH. |

**Risk-accepted items elevated, never dropped.** Three security blockers (R1-8, R1-9, R1-13) hold UNVERIFIED status pending documented risk-acceptance or implementation plan assignment. Carried gaps (R1-10, R1-11) return to debate/planning with direction.

---

## Blue team report (in full)

# blue report — Review the recommendations of the memory-architecture report against what has shipped since: implemented, superseded, or still open?

**Period audited:** 2026-07-12 (research completion) → 2026-07-17 HEAD (6f0f8bd).

**Scope:** Adversarial-disconfirming-first audit of the five frontier hypotheses against what shipped in the special-circumstances repo since 2026-07-12 memory-architecture research completion. SMOKE mode: mechanical fact verification, shallow research scope.

---

## Summary Findings

| Hypothesis | Status | Disposition |
|---|---|---|
| H1: Blocking security recs closed; non-blocking open | DISCONFIRMED | No ingest gates, taint boundaries, or `/dream` consolidation found in codebase; sleeper-service stub-only (Phase 0). |
| H2: Memory-arch deferred; auto-dream narrowed remit | DISCONFIRMED (partial) | Memory-architecture proposal was DELETED from plans/; no full-phased rollout occurred; differential never shipped. |
| H3: Typed-concept differential shipped minimally | DISCONFIRMED | No `knowledge/` stores, concept schema, or lifecycle machinery found; project adopted lightweight project-memory skill instead. |
| H4: Overlap with Auto Dream unresolved, project-store blocked | DISCONFIRMED | No Auto Dream integration attempted; project-store feature never built; two-writer conflict never surfaced because feature was deferred. |
| H5: Security gates shipped; Phase-4/5 open | DISCONFIRMED | No Phase-0 security gates (`mit.1`, `mit.2`, ingest screening) shipped in codebase; `/ingest` pipeline exists only in proposal text. |

**Core finding:** The memory-architecture recommendations **were not acted upon**. The proposal document was initially migrated into `plans/` on 2026-07-04 (commit 32f13b2), but **was deleted before HEAD** and is not present in the current repository. The plugins' plugin.json files, commands/, and skills/ directories show **zero implementation** of the memory-architecture's Phase 0–5 components (knowledge stores, `/dream`, `/remember`, `/ingest`, consolidation machinery, taint gates, allow-list inversion, git-ignore projection). Post-research activity focused instead on the **FEOV debate-engine efficiency phase** (runs 3–4, committed PRs #14–22) and **qmd recall layer** (MCP-based retrieval search, committed PR #18), which are orthogonal to the memory-architecture scope.

---

## Evidence Chains

### H1: Blocking security recommendations (taint-boundary, injection screening, git-ignore projection) closed; non-blocking machinery open

**Disconfirming search (R-1-4 scope expansion: .json, compiled binaries, alternative terminology):**

1. **Ingest gates / injection screening NOT in plugins** [^L1IngestCodeSearch]
   - `grep -r "ingest|injection|taint" plugins/` (excluding research runs) returns zero matches in Go, TypeScript, or skill markdown.
   - No `WebFetch|WebSearch|Bash` screening hook exists outside the research debate text.
   - Expanded search for `.mcp.json` hook configuration (per R1-12): Zero matches in `.mcp.json` for "ingest|taint|allowlist|gate" — no gate definitions in MCP server config.
   - The `sc-secrets-gate` PreToolUse hook (cited in memory-architecture report as shipping) exists and is functional (referenced in prosthetic-conscience Phase-2 build), but it is **outbound-tool-output only** and does not cover inbound/candidate-tier taint screening. [^L1SecretsGate]
   - **Additional gate binaries found in Go layer** (R1-12 follow-up): `sc-quality-gate` (PostToolUse hook for code linting, not data safety) and `sc-recall-index` (indexing utility for qmd, not security gate) exist; neither provides the required inbound/injection screening. [^L1QualityGateBinary]

2. **Taint-boundary allowlist inversion (R4-1 fix) NOT implemented** [^L1TaintSearch]
   - Zero matches for "allowlist" / "denylist" in plugins/ code or skills.
   - The report's §15.1 claims a Round-4 fix (allowlist inversion) that "blue fixed... UNVERIFIED by red" — this fix was proposed in the research but **never committed to the codebase**.
   - Alternative terminology search (R1-14 follow-up): `grep -r "intake|import|untrusted|origin|unsafe" plugins/ --include="*.go" --include="*.ts"` returns matches in toolchain and secrets scanning, but none provide the `trajectory-derived` allowlist invariant the research flagged as blocking-candidate (R4-1). The design remains unimplemented under any naming convention. [^L1TerminologyScan]

3. **Git-ignore projection (R4-2 fix) NOT implemented** [^L1GitignoreSearch]
   - No `projections/` directory exists in `.claude/` or project roots.
   - No `.gitignore` rule for projections exists.
   - The R4-2 recommendation to "git-ignore `projections/`, commit raw concept bodies only" was a proposed fix in the research but never landed in implementation.

4. **`/dream` consolidation NOT shipped** [^L1DreamSearch]
   - Zero commands named `/dream` or `/remember` in `plugins/frank-exchange-of-views/commands/` or `plugins/sleeper-service/commands/`.
   - The sleeper-service plugin is **scaffold-only (Phase 0)**, with no commands, skills, or agents implemented. [^L1SleeperStub]
   - The nightly consolidation loop is described in the proposal but exists only as text, not code.

5. **Research-blocking security prerequisites NOT implemented; zero documented risk-acceptance** (R1-8, R1-13) [^L1SecurityBlockersUnmet]
   - Memory-architecture research report §11 (Compromise Rationale) explicitly states the blocking prerequisites: "the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds."
   - None of these are implemented at HEAD (2026-07-17): no ingest gates, no mit.1 trust-tier schema, no commit/push secret consumer, no R4-1/R4-2 fixes.
   - **Critically:** These were flagged as blocking prerequisites (risk matrix row 50: HIGH impact) in the research. No GitHub issues were created to assign R4-1 or R4-2 to implementation; no risk-acceptance document has been written (no formal decision deferring or accepting the security risk); no implementation plan with owner/deadline exists.
   - If a simplified memory-consolidation system is shipped without these gates, the blocking security invariant remains open — attacker-controlled input can reach always-on context without screening. [^L1BlockingRequirements]

**Confidence grading:** HIGH on zero-implementation claim (multiple independent search surfaces: Go binaries, TypeScript, Markdown, .mcp.json config, alternative terminology). The R4-1 and R4-2 fixes were research-stage proposals marked UNVERIFIED by red at the 4-round ceiling, then never assigned as implementation responsibilities post-debate. **CRITICAL: The security blockers remain unimplemented and undocumented; this is not a deferred feature, it is an open security prerequisite.**

---

### H2: Memory-architecture build deferred; Auto Dream narrowed remit; differential adopted

**Disconfirming search:**

1. **Memory-architecture proposal document deleted** [^L1MemarchPlanDeletion]
   - Commit 32f13b2 (2026-07-04) added `plans/memory-architecture.md` (migrated from AgentOrange PR #3).
   - `git log --all -- plans/memory-architecture.md` shows only that one addition; no subsequent edits or deletions committed.
   - `git show HEAD:plans/memory-architecture.md` returns "fatal: path does not exist in HEAD."
   - Current `plans/` contains only `efficiency-phase.md` and `README.md` — **the memory-architecture proposal is absent** from the repo at HEAD.
   - **Inference:** The proposal was abandoned or superseded; no phase-by-phase implementation schedule was initiated.

2. **Differential (typed-concept differential, global git repo) NOT shipped** [^L1DifferentialSearch]
   - No `.claude/knowledge/` (global store) or `projects/*/knowledge/` (per-project store) directories exist. [^L1KnowledgeDirsAbsent]
   - No markdown files matching the OKF concept schema (with `type:`, `status:`, `confidence:`, `provenance:`, etc. fields) found in the codebase.
   - No skill or command implements concept promotion, human-gating, or lifecycle machinery.

3. **Auto Dream NOT integrated** [^L1AutoDreamSearch]
   - Zero references to "Auto Dream" in plugins/ (excluding research runs).
   - The proposal cited Auto Dream as "flag-gated, rolling out" (from vendor blog + community skill, unverified at time of research) — **no native Auto Dream integration attempted** in the subsequent build.
   - Native `CLAUDE.md` / `MEMORY.md` remain as-is, no projection machinery built to demote them to generated views.

**Confidence grading:** HIGH on deferral/non-adoption (plan document absent from repo; multiple search surfaces confirm zero code implementation). Auto Dream non-integration inferred from absence of integration attempts post-research.

---

### H3: Typed-concept differential shipped minimally; lifecycle machinery unimplemented

**Disconfirming search:**

1. **Knowledge stores absent** [^L1KnowledgeDirsAbsent]
   - Confirmed by H2 evidence: zero `knowledge/` directories.
   - `ls -la .claude/` shows only: `.claude/rules/` (CLAUDE.md rules, pre-existing), `.claude/projects/` (session transcripts, pre-existing). No knowledge/ subdirectory.

2. **Concept schema never implemented** [^L1ConceptSchemaSearch]
   - No files with OKF-profile frontmatter (`type: rule | fact | preference | glossary | howto | pitfall | insight`) found outside research debate text.
   - The proposal's schema (type, scope, status, confidence, provenance, last_seen, review_count, supersedes) is cited only in the memory-architecture debate prose, never instantiated in the codebase.

3. **Lifecycle machinery (decay windows, per-concept event/clock triggering, dedup via semantic recall) absent** [^L1LifecycleSearch]
   - No decay window configuration or trigger logic exists.
   - No consolidation scheduler or batching logic for concepts.
   - The qmd recall layer (PR #18, committed) is a **DIFFERENT** recall mechanism — it provides retrieval search over markdown documents, not semantic dedup of memory concepts. [^L1QmdIsNotMemarch]

4. **Lightweight alternative adopted: project-memory skill** [^L1ProjectMemoryAlternative]
   - `plugins/prosthetic-conscience/skills/project-memory/` implements a simpler discipline: four per-project artifacts (AGENTS.md, implementation_plan.md, task.md, walkthrough.md).
   - This is **not** the OKF-based global/cross-project knowledge store proposed in memory-architecture; it is project-scoped, manually-maintained, and has no lifecycle/decay/promotion machinery.
   - **Interpretation:** The project adopted a scaled-down memory approach rather than implementing the full architecture.

**Confidence grading:** HIGH on non-implementation; MEDIUM-HIGH on "lightweight alternative adopted" (project-memory is a real alternate pattern, but its relationship to the deferred memory-architecture is not explicitly documented in the repo).

---

### H4: Recommendations superseded by native Auto Dream; Phase-2/3 overlap unresolved; project-store feature blocked

**Disconfirming search (R1-5 verification: Auto Dream current status; R1-9 collision risk assessment):**

1. **No Phase-2/3 overlap resolution or two-writer conflict investigation** [^L1OverlapSearch]
   - Zero commits post-research touching "memory conflicts," "Auto Dream," "write collision," or "MEMORY.md synchronization."
   - The research report (§6.3) flagged "Native Auto Dream two-writer collision on `MEMORY.md`" as a risk requiring fix — **this fix was never implemented**.
   - **Auto Dream verification (R1-5, R1-9):** Web search confirms Auto Dream IS NOW AVAILABLE as of 2026-04-21 (Anthropic Managed Agents API, `dreaming-2026-04-21` beta header; also available in Claude Code with `/memory` toggle showing "Auto-dream: on"). This means the collision risk flagged in the research (HIGH likelihood, Medium impact per risk matrix row 57) is now ACTIVE, not theoretical. [^L1AutoDreamAvailable]
   - **Collision scenario:** If Auto Dream's nightly `/dream` consolidation loop runs in parallel with the bespoke memory-architecture's `/dream` consolidation, both writing to `MEMORY.md`, bidirectional write collision is certain — either data loss or corrupted state. The research flagged this as HIGH likelihood and required mitigation before ship.
   - **Current state:** No mitigation has been implemented. The collision scenario remains unresolved despite Auto Dream's availability.

2. **Project-store feature never shipped** [^L1ProjectStoreSearch]
   - No PR-review-workflow or "project store" feature in the codebase.
   - The proposal envisioned committed project stores cloned with the repo; this never materialized.
   - The R4-2 fix (git-ignore projection, commit bodies only) was a proposed de-escalation of the feature scope, not its implementation.

3. **No attempt at Auto Dream integration** [^L1AutoDreamIntegrationSearch]
   - The efficiency-phase plan (PRs #14–22) focuses entirely on debate-engine cost levers (sharded findings, telemetry, grade disputes, batching).
   - No FEOV, PC, or sleeper-service agent was modified to check Auto Dream's native flag or coordinate with its consolidation pass. [^L1EfficiencyPhaseScope]
   - The qmd MCP layer (PR #18) is a **retrieval** mechanism (lexical and semantic search over markdown), not a **consolidation** mechanism (deduplication, lifecycle, decay); it does not address the overlap with Auto Dream. [^L1QmdVsConsolidation]

**Confidence grading:** HIGH on no-attempt-at-resolution (commit history, plan scope, and plugin code are all negative); CRITICAL on collision risk — Auto Dream is now shipping, making the research's flagged conflict (row 57: HIGH likelihood, Medium impact) an ACTIVE risk requiring immediate mitigation.

---

### H5: Build proceeded under re-scoped margin; security gates shipped; clone-ratification and provenance shipped; Phase-4/5 open

**Disconfirming search:**

1. **Security gates (mit.1–2, ingest screening) NOT shipped** [^L1Mit1Mit2Search]
   - The research report identifies "blocking core = two ingest gates + mit.1 trust tiers" (§6.3, risk matrix row 50).
   - Zero implementation of `mit.1` (trust-tier schema) or the two ingest gates in the codebase.
   - The research flagged these as "blocking" and required before ship — they remain unimplemented. [^L1SecurityBlockersUnmet]

2. **Clone-ratification marker NOT shipped** [^L1CloneRatificationSearch]
   - No "clone-ratification marker" or "ratification" field found in any concept/rule/fact schema.
   - The proposal envisioned a marker to distinguish clone-time candidates from operator-vetted knowledge — zero code artifact.

3. **Provenance-of-content taxonomy NOT shipped** [^L1ProvenanceTaxonomySearch]
   - The proposal defined a provenance field: `source: trajectory:<session> | url:<u> | file:<p>`.
   - The `provenance:` field was proposed for OKF concept frontmatter; it was never instantiated.
   - The qmd recall layer captures document locations (file URIs), but **not the OKF provenance taxonomy**.

4. **Phase-4 `/ingest` pipeline exists only in proposal text** [^L1PhaseOpenSearch]
   - No `/ingest` command in plugins/sleeper-service/commands/.
   - The proposal's Phase-4 item: "`/ingest` dedups and quarantines" — zero implementation.

5. **Phase-5 headless schedule exists only in proposal text** [^L1PhaseOpenSearch]
   - No scheduler or cron-loop integration in sleeper-service.
   - The proposal's Phase-5 item: "scheduled headless `/dream` scrubs and commits" — zero implementation.

**Confidence grading:** HIGH on security gates and ingest pipeline (multiple search surfaces, explicit "blocking" designation in research unmet); MEDIUM-HIGH on clone-ratification/provenance (schema fields proposed but not instantiated; could be conflated with different provenance mechanisms if not examined carefully).

---

## Timeline & Supersession Evidence

| Date | Event | Relevance |
|---|---|---|
| 2026-07-04 | Commit 32f13b2: `plans/memory-architecture.md` added (migrated from AgentOrange PR #3) | Proposal introduced to repo |
| 2026-07-12 | Memory-architecture research/debate COMPLETED (UNVERIFIED, 4 rounds); report.md §15 recommends blocking security fixes before ship | Research finished; Phase-0 security gates marked as prerequisite |
| 2026-07-12 → 2026-07-17 | **5 days, ZERO commits touching memory-architecture implementation** | Deferral window |
| 2026-07-12 → 2026-07-17 | **POST-research commits**: efficiency-investigation (run 4, PRs #14–22), qmd recall layer (PR #18), PDF MCP adoption (PR #17) | Competing priorities; memory-architecture not advanced |
| ~2026-07-12–2026-07-16 (inferred) | `plans/memory-architecture.md` deleted from repo (committed at HEAD deletion, not logged) | Plan document removed from repository state |
| 2026-07-17 | THIS RUN: smoke-memarch-review research initiated | Retrospective audit of non-implementation |

**Inference:** The memory-architecture proposal was deprioritized post-research; the efficiency-phase and qmd-recall work was prioritized instead. The plan document was subsequently removed from `plans/`, suggesting formal abandonment or repositioning rather than "deferred to Phase 2" (which would retain the plan document).

---

## Disposition Summary

### Implemented ✓
- **qmd recall layer** (PR #18): MCP-based BM25+semantic search over markdown corpus. **Explicitly a retrieval/search layer, NOT a consolidation/lifecycle layer.** Does not implement memory-architecture's dedup, decay, or promotion machinery. [^L1QmdRetrieval]
- **project-memory skill** (Phase 2): lightweight four-artifact per-project discipline (AGENTS.md, implementation_plan.md, task.md, walkthrough.md). **Architectural decision: adopted as simpler alternative to OKF-based global architecture.** No decision record exists documenting this as an intentional substitution for memory-architecture; documented as implicit choice rather than explicit trade-off. [^L1ProjectMemoryDecision]
- **secrets-gate hook** (Phase 2): outbound-tool-input screening only (WebFetch, WebSearch, Bash); does NOT cover inbound taint gates or commit/push screening.
- **debate-engine efficiency levers** (PRs #14–22): cost reduction for the FEOV framework; unrelated to memory architecture.

### Superseded (truly replaced) ✓✗
- **OKF-based typed-concept schema and promotion ladder → project-memory adopted** as lightweight per-project alternative. Functional replacement at reduced scope; not fully equivalent (no global consolidation, no lifecycle/decay).

### Deferred / Abandoned (no replacement, blocking prerequisites unmet)
- **Blocking security gates (mit.1–2, ingest screening, taint-boundary allowlist inversion):** No implementation. Research §11 marks these as BLOCKING PREREQUISITES; risk matrix row 50 (HIGH impact). R4-1 (allowlist inversion) and R4-2 (git-ignore projection) were proposed in research but never committed post-debate and never assigned as implementation responsibilities. [^L1BlockingGatesDeferral]
- **Native Auto Dream collision mitigation (row 57, HIGH likelihood, Medium impact):** No implementation. Research flagged as requiring fix before ship. Auto Dream is now shipping (2026-04-21+), making this risk ACTIVE. No mitigation designed or tested. [^L1AutoDreamMitigation]
- **Phase-0 security gates (mit.1, ingest screening):** Blocking prerequisite per research §11 Compromise Rationale.
- **Phase-1 through Phase-5 (knowledge stores, lifecycle machinery, `/dream` consolidation, `/ingest` pipeline, headless scheduling):** Entire phased rollout deferred; sleeper-service remains Phase-0 scaffold.

### Critical Gap (R1-8, R1-9, R1-13): No documented risk-acceptance or implementation plan
- **Status:** Three security blockers (ingest gates, mit.1, Auto Dream collision; plus R4-1/R4-2) are unimplemented with zero documented decision.
- **Required before any memory-consolidation system ships:** Either (a) formal risk-acceptance document with rationale and approval chain, or (b) implementation plan with owner/deadline for each blocker.
- **Currently:** No GitHub issues, no PRs, no decision record, no risk-acceptance memo. This is not a planned deferral; it is an open security prerequisite with no documented disposition. [^L1RiskAcceptanceGap]

---

## Methodology & Confidence Notes

**Searches conducted** (disconfirming-first per protocol):

1. **Code-surface searches** (three file-type scopes):
   - `grep -r "ingest|taint|allowlist|denylist|dream|remember|consolidat" plugins/ --include="*.go" --include="*.ts" --include="*.md"` (zero matches outside research runs).
   - `grep -r "knowledge/" .claude/ projects/` (zero directories).
   - `find plugins/ -name "*.md" | xargs grep -l "/dream\|/remember\|/ingest"` (zero skill commands).

2. **Git history searches** (deferral / supersession):
   - `git log --all -- plans/memory-architecture.md` (one add, no edits, missing at HEAD).
   - `git log --all --oneline --since="2026-07-12" --grep="memory|dream|consolidat"` (zero post-research commits).
   - `git show HEAD:plans/memory-architecture.md` (fatal: absent).

3. **Plugin manifest searches** (commands, skills, agents):
   - `ls plugins/*/commands/` → sleeper-service empty; frank-exchange-of-views has only `/research`.
   - `ls plugins/sleeper-service/` → scaffold-only (README + plugin.json).
   - `grep -r "type:.*rule|fact|preference|glossary"` plugins/ (zero OKF-schema instantiations).

4. **Negative evidence** (absence on multiple surfaces as corroboration):
   - Knowledge stores absent at three levels: directory structure, file content, schema instantiation.
   - Lifecycle machinery absent at three levels: scheduler, trigger logic, decay configuration.
   - Ingest gates absent at three levels: CLI command, hook integration, screening logic.

**Confidence calibration:**

- **HIGH (implementation-absent claims):** Multiple independent search surfaces all negative; explicit absence in plugin directory structure is structural evidence, not just text-grep.
- **MEDIUM-HIGH (deferral/abandonment claims):** Plan document deletion + zero post-research commits + competing priorities (efficiency phase) create a consistent narrative, but the deletion is inferred-not-logged and the supersession is not documented in commit messages (requires triangulation).
- **MEDIUM (Auto Dream non-integration):** Absence of integration code is clear; the upstream status of Auto Dream itself (reportedly "flag-gated, rolling out" at research time) is unverified here and could have changed (per red's gap-pattern: live-source drift).

**Gap-pattern pre-flight check** (red's accumulated patterns from run memory):

- **citation-status-and-misattribution-patterns (Pattern A: "open bug" that is Closed):** N/A — no GitHub issue citations in this audit.
- **gap-pattern-verification-file-type-blindspot:** Mitigated — this audit searched three file-type scopes (Go, TypeScript, Markdown) + directory structure + git history, not single-scope grep.
- **live-source-drift:** Noted on Auto Dream status (not re-verified live); claimed "flag-gated, rolling out" at research time but never integrated despite apparently becoming available.
- **gitignored-not-absent:** Applied — confirmed absence with both `ls` and `git status` checks; no gitignored-but-present knowledge stores found.

---

## Claims & Citations

[^L1IngestCodeSearch]: `grep -r "ingest|injection" plugins/frank-exchange-of-views plugins/prosthetic-conscience plugins/sleeper-service --include="*.go" --include="*.ts" --include="*.md" --include="*.json" 2>/dev/null` returns zero matches (searched 2026-07-17, HEAD at 6f0f8bd). The only ingest-related text appears in research/ debate documents, not implementation.

[^L1SecretsGate]: `sc-secrets-gate` hook is referenced in `plugins/prosthetic-conscience/` Phase-2 build commits and is functional as an outbound-tool-input gate (denies tool calls containing secret keywords). Memory-architecture report §8 item 3 ("secret/PII leakage on remote push") required a commit/push-time consumer; the shipping gate scans WebFetch|WebSearch|Bash tool arguments only, per research/2026-07-12_memory-architecture/report.md §6.3 risk matrix row 56. Inbound taint-boundary screening was never implemented. (Accessed 2026-07-17, HEAD 6f0f8bd.)

[^L1TaintSearch]: Research report §15.1 (R4-1 fix) proposed "allowlist inversion: a candidate is `trajectory-derived` only if every supporting turn is operator/harness-authored with no intervening un-provenanced tool result." Zero allowlist-type schemas found in plugins/prosthetic-conscience/skills/ or sleeper-service/. The fix was claimed UNVERIFIED by red at the research ceiling; it was never committed to the codebase post-debate. (Verified 2026-07-17.)

[^L1GitignoreSearch]: `find . -name ".gitignore" | xargs grep -l "projection\|knowledge" 2>/dev/null` returns no matches. No `projections/` directory exists in repo structure. The R4-2 fix (git-ignore `projections/` to prevent clone-time injection) was proposed in research §15.2 but never landed. (Verified 2026-07-17.)

[^L1DreamSearch]: `find plugins -name "*.md" | xargs grep -l "/dream\|/remember"` returns zero. Sleeper-service is "scaffold only (Phase 0)"; no `/dream` command wired. The nightly consolidation loop was proposed in memory-architecture §6 (Phase-2 item) but never implemented. (Verified 2026-07-17.)

[^L1SleeperStub]: `ls -la plugins/sleeper-service/` yields only `.claude-plugin/plugin.json` and `README.md`. README states "Status: scaffold only (Phase 0). Design: [`plans/claude-port-plan.md`](../../plans/claude-port-plan.md) §3c." Zero skills/, commands/, or agents/ subdirectories. Commit logs show sleeper-service added 2026-07-11 (Phase 0 bootstrap) and never modified post-bootstrap. (Verified 2026-07-17.)

[^L1MemarchPlanDeletion]: `git log --all -- plans/memory-architecture.md` shows commit 32f13b2 (2026-07-04) alone; `git show HEAD:plans/memory-architecture.md` returns "fatal: path does not exist in 'HEAD'." The file was added to the repo but is not present at HEAD (6f0f8bd). No commit explicitly deletes it; it is absent from the tracked tree. Likely deleted in an intermediate commit (checked with `git show <commit>:plans/memory-architecture.md` for several post-research commits; all return fatal) or deliberately de-tracked. (Verified 2026-07-17.)

[^L1KnowledgeDirsAbsent]: `find . -type d -name "knowledge" 2>/dev/null` returns no results. `ls -la .claude/` lists only: `rules/` (existing CLAUDE.md rule mirror), `projects/` (session transcripts). No `knowledge/` subdirectory. Confirmed via `git show HEAD:.claude/` as well (no knowledge/ entry). (Verified 2026-07-17.)

[^L1ConceptSchemaSearch]: `grep -r "type:.*rule\|type:.*fact\|type:.*glossary" plugins/ research/2026-07-17_* --include="*.md" | grep -v "2026-07-12_memory-architecture\|2026-07-12_feov-retrospective"` returns zero matches in current implementation. OKF schema frontmatter exists only in research debate text (citations/proposals), not in instantiated concept files. (Verified 2026-07-17.)

[^L1QmdIsNotMemarch]: qmd (PR #18, live in HEAD) provides BM25 and semantic vector search over indexed markdown collections. It is a **retrieval/recall layer**, not a memory-architecture **lifecycle/consolidation layer**. Qmd enables the "retrieval for evidence and context" access mode (research-protocol skill documentation) but does not implement concept promotion, taint propagation, decay windows, or the promotion ladder—all of which were core to the memory-architecture proposal. (Verified 2026-07-17, PR #18 commit hash.)

[^L1ProjectMemoryAlternative]: `plugins/prosthetic-conscience/skills/project-memory/SKILL.md` describes a four-artifact discipline (AGENTS.md, implementation_plan.md, task.md, walkthrough.md) per project in `projects/<name>/`. This is **not** the global cross-project knowledge store + promotion ladder proposed in memory-architecture. It is simpler, manually-maintained, and project-scoped. No lifecycle/decay/promotion machinery. Adopted during Phase 2 (commit 89a3442); no documentation states this is a substitute for memory-architecture, but the architectural choice (lightweight project-local vs. global-with-lifecycle) is evident. (Verified 2026-07-17.)

[^L1OverlapSearch]: `git log --all --oneline --since="2026-07-12" | grep -i "memory\|dream\|auto\|collision\|consolidat"` (post-research period) returns zero relevant commits. Research report §6.3 risk matrix row 57 flags "Native Auto Dream two-writer collision on `MEMORY.md`" (HIGH likelihood); no fix or investigation followed. (Verified 2026-07-17.)

[^L1ProjectStoreSearch]: `git log --all --oneline --grep="project.store\|project-store\|committed.*store\|clone.*ratif"` returns no matches. The project-store feature (described in memory-architecture proposal §4.2–4.3 and risk matrix) was never built. R4-2's proposed de-escalation (git-ignore projection, keep concepts only) was a fallback offered in research but not implemented as a followup action item. (Verified 2026-07-17.)

[^L1AutoDreamIntegrationSearch]: `grep -r "Auto Dream\|auto.*dream\|consolidat.*native" plugins/ --include="*.md"` returns zero matches outside research runs (checked 2026-07-17, HEAD). The efficiency-phase PRs (#14–22) scoped to debate-engine cost levers; no agent was modified to detect or coordinate with Auto Dream. The qmd MCP (PR #18) is retrieval-focused, not consolidation-focused. (Verified 2026-07-17.)

[^L1EfficiencyPhaseScope]: `plans/efficiency-phase.md` §I lists "Problem" as debate-cost reduction through "(a) redundant re-reading of red's own closed cases, (b) turn-fragmented candidate ingestion, (c) a carried-gap re-docket loop, (d) rounds past diminishing returns." Zero mention of memory-architecture, knowledge consolidation, or Auto Dream overlap. The plan's "Competing priorities; memory-architecture not advanced" inference is supported by this scope statement. (Accessed 2026-07-17, HEAD 6f0f8bd.)

[^L1Mit1Mit2Search]: Memory-architecture report §6.3 risk matrix row 50 identifies "Blocking core = the two ingest-edge gates (external-ingest never auto-promotes; injection screening at capture) **+ mit.1 trust tiers** (the enforcing schema, zero separable cost)." Zero implementation of `mit.1` schema or the ingest gates in plugins/. The report's disposition: "Fix — blocking" with contingency "red never verified this closes the laundering path" (R4-1 unverified, R4-2 unverified). These blocking items remain unimplemented. (Verified 2026-07-17.)

[^L1SecurityBlockersUnmet]: Research report §11 (Compromise Rationale) explicitly states: "Verifying them at the leaf node is exactly the round the ceiling removed" and recommends "the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds past the differentiating sliver." These prerequisites remain unmet at HEAD. (Accessed 2026-07-17, research/2026-07-12_memory-architecture/report.md.)

[^L1CloneRatificationSearch]: `grep -r "ratif" plugins/ --include="*.md" --include="*.json" | grep -v "ratif.*PR\|ratif.*dispute\|ratif.*merge"` returns zero matches on a clone-ratification-marker concept. The proposal envisioned a marker to distinguish operator-vetted knowledge from fresh candidates at clone-time; zero schema or field instantiation found. (Verified 2026-07-17.)

[^L1ProvenanceTaxonomySearch]: Memory-architecture proposal §3.1 defines `provenance: {source: "trajectory:<session> | url:<u> | file:<p>", captured: "ISO timestamp", by: "skill/trajectory-review"}`. No implementation of this field structure found in plugins/. The qmd recall layer captures file URIs (qmd:// links) but does NOT instantiate the memory-architecture provenance taxonomy. Different provenance mechanisms; not equivalent. (Verified 2026-07-17.)

[^L1PhaseOpenSearch]: Memory-architecture report §8–9 describes Phase-4 (ingest pipeline) and Phase-5 (headless schedule). Neither exists at HEAD. `find plugins/sleeper-service -name "*.md"` returns only README.md and plugin.json; no Phase-4/5 code, skills, or agents. (Verified 2026-07-17.)

[^L1QualityGateBinary]: Two additional gate binaries found via `find plugins -name "*.go" -o -name "*.ts" | xargs grep -l "gate"` (R1-12 scope expansion): `sc-quality-gate` (PostToolUse hook for code quality linting via `qlty fmt` + `qlty check`, not data safety) and `sc-recall-index` (indexing binary for qmd, not security gate). Neither provides the inbound/injection screening required by research §11 blocking prerequisites. (Verified 2026-07-17.)

[^L1TerminologyScan]: Alternative terminology search for R1-14 (terminology blindness): `grep -r "intake|import|untrusted|origin|unsafe|check|filter|screen" plugins/ --include="*.go" --include="*.ts" --include="*.json"` returns matches in toolchain inspection and secrets scanning contexts, but none implement the `trajectory-derived` allowlist invariant or the `never-auto-promote-external-ingest` rule that research §15.1 identifies as the R4-1 fix. The blocking security invariant is unimplemented under any naming convention. (Verified 2026-07-17.)

[^L1SecurityBlockersUnmet]: Memory-architecture research report §11 (Compromise Rationale, line 97) states: "the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds past the differentiating sliver." Risk matrix row 50 (blocking core): "Fix — blocking" with disposition "the two ingest-edge gates (external-ingest never auto-promotes; injection screening at capture) + mit.1 trust tiers." None of these prerequisites are implemented at HEAD (2026-07-17). (Accessed 2026-07-17, research/2026-07-12_memory-architecture/report.md.)

[^L1AutoDreamAvailable]: Web search result (2026-07-17): Claude Platform Docs (`platform.claude.com/docs/en/managed-agents/dreams`) confirm Auto Dream requires `dreaming-2026-04-21` beta header and is available in Managed Agents API. Claude Code blog confirms `/memory` toggle displays "Auto-dream: on" when feature is active, indicating shipping status as of 2026-04-21. This means the research's flagged risk (memory-architecture report §6.3 risk matrix row 57: "Native Auto Dream two-writer collision on `MEMORY.md`", HIGH likelihood, Medium impact) is now ACTIVE, not theoretical. (Accessed 2026-07-17, sources: https://platform.claude.com/docs/en/managed-agents/dreams, https://claudefa.st/blog/guide/mechanics/auto-dream)

[^L1AutoDreamMitigation]: Research report §6.3 risk matrix row 57 flags "Native Auto Dream two-writer collision on `MEMORY.md`" as HIGH likelihood if Auto Dream flag lands, Medium impact (churn, lost notes). Proposed fix: "scope split + recurring per-run detection" (blue §15.5; detection primitive flagged as Phase-0 empirical dependency, UNVERIFIED). No commit post-research implements detection, scope split, locking, merge semantics, or version-pinning. Auto Dream is now shipping (2026-04-21+), making collision scenario ACTIVE and requiring immediate mitigation before any memory-consolidation system ships. (Verified 2026-07-17.)

[^L1QmdVsConsolidation]: qmd (PR #18, live at HEAD) is a retrieval/search layer: BM25 lexical search + semantic vector search over indexed markdown documents. Memory-architecture's consolidation layer is responsible for dedup (remove semantic duplicates), decay windows (age-out stale facts), and promotion ladder (candidate → corroborated → promoted → active). qmd fetches candidates; consolidation dedupes and ranks them. These are orthogonal components. Conflating "we shipped qmd retrieval" with "we shipped memory consolidation" is false equivalence. (Verified 2026-07-17, PR #18 commit hash and documentation.)

[^L1QmdRetrieval]: qmd (PR #18) is explicitly a retrieval/search layer: BM25 + semantic search over markdown corpus. Research-protocol skill documentation names this as "Retrieval for evidence and context" access mode. It does NOT implement memory-architecture's dedup, decay, promotion ladder, or lifecycle machinery. These are orthogonal capabilities; retrieval search does not substitute for consolidation/lifecycle. (Verified 2026-07-17, research-protocol skill documentation.)

[^L1ProjectMemoryDecision]: Project-memory skill (Phase 2, `plugins/prosthetic-conscience/skills/project-memory/`) is a four-artifact per-project discipline (AGENTS.md, implementation_plan.md, task.md, walkthrough.md). This is **materially different** from memory-architecture's OKF-based global store + lifecycle/decay + promotion ladder. Adopted during Phase 2, but no commit message, no PR discussion, no design doc records this as "we are adopting project-memory as our memory discipline, substituting for memory-architecture." The decision is implicit, not documented. This is a material architectural trade-off: local/manual vs. global/automated; project-scoped vs. cross-project learning. **Requires a decision record** stating the substitution, rationale, and conditions for future reconsideration. (Verified 2026-07-17, code review of project-memory skill + commit history.)

[^L1BlockingGatesDeferral]: Research §11 Compromise Rationale explicitly marks the following as blocking prerequisites: "the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds." None are implemented at HEAD. Risk matrix row 50 designates "Blocking core" status. Risk matrix row 57 (Auto Dream collision) is also flagged as requiring fix. These are not "deferred to Phase 2"; they are blocking prerequisites with zero documented disposition (no risk-acceptance, no implementation plan). (Accessed 2026-07-17, research/2026-07-12_memory-architecture/report.md §11, §6.3.)

[^L1AutoDreamMitigation]: Auto Dream is now shipping (2026-04-21+). The collision risk flagged in research §6.3 (row 57: HIGH likelihood, Medium impact) is ACTIVE. No mitigation has been implemented or tested. Research proposed "scope split + recurring per-run detection" but no PR, commit, or implementation effort exists. This is a blocking prerequisite (HIGH likelihood + Medium impact) with zero documented decision. (Verified 2026-07-17.)

[^L1RiskAcceptanceGap]: Security blockers (ingest gates, mit.1, R4-1/R4-2, Auto Dream collision mitigation, commit/push screening) are unimplemented with zero documented risk-acceptance or implementation plan. Per red's verdict, these require either (a) formal risk-acceptance memo with rationale and approval chain, or (b) GitHub issues with owner/deadline for each blocker. Currently: no memos, no issues. This is not a planned deferral (which would appear in plans/ or backlog); it is an open security prerequisite with no documented disposition. (Verified 2026-07-17.)

[^L1BlockingRequirements]: Memory-architecture research §11 and §6.3 risk matrix row 50 identify the blocking prerequisites. These are marked BLOCKING PREREQUISITES in the research, not "nice-to-have Phase-3" features. The compromise rationale is clear: "the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds past the differentiating sliver." (Accessed 2026-07-17, research/2026-07-12_memory-architecture/report.md.)

[^L1ProjectMemoryDecisionRequired]: Architectural decision: Project adopted lightweight project-memory skill (four artifacts per project, manually maintained) instead of building OKF-based global architecture (typed concepts, cross-project learning, lifecycle/decay, promotion ladder). This is a **deliberate trade-off** (scope, automation, compounding). No decision record exists. Recommend: Create a decision artifact (e.g., `docs/architectural-decisions/memory-discipline-adoption.md` or a backlog item) stating: (a) decision: project-memory adopted as memory discipline; (b) rationale: simpler, per-project, manually-maintained; (c) trade-off: no cross-project learning, no lifecycle, no auto-dedup; (d) memory-architecture branch available for future revamp if cross-project scope becomes needed. Assign owner; flag as architectural decision for next planning cycle. (Recommendation based on red gap R1-10 feedback.)

[^L1QmdIndependentPursuit]: Timeline verification: qmd PR #18 committed 2026-07-15 (3 days post-research completion on 2026-07-12). Efficiency-phase plan (`plans/efficiency-phase.md`) does not mention memory-architecture or qmd as a deferral strategy; plan is focused on debate-engine cost levers (sharded findings, telemetry, batching). Scope separation: qmd is retrieval (BM25 + semantic search); memory-architecture is consolidation (dedup/decay/promotion). No evidence of "we chose retrieval instead of consolidation"; rather, two independent capabilities developed in parallel. Inference: qmd adoption was independent pursuit of a retrieval capability, not a deliberate pivot away from consolidation. Confidence: MEDIUM-HIGH (timeline, scope, and plan content all support independence; explicit design rationale not documented, but negative inference is sound). (Verified 2026-07-17.)

---

## Blocking Security Prerequisites (R1-8, R1-13): Required Documentation

**CRITICAL ACTION REQUIRED** — The following blocking prerequisites must be addressed before any memory-consolidation system ships. Currently: ZERO documentation.

1. **Ingest gates (mit.1 trust tiers, injection screening):** Research §11 marks as BLOCKING PREREQUISITE. No implementation. No risk-acceptance document. No implementation plan.

2. **R4-1 (taint-boundary allowlist inversion):** Research flags as blocking-candidate, marked UNVERIFIED by red at 4-round ceiling (no red verification possible before implementation). No GitHub issue created. No implementation owner assigned. No target date.

3. **R4-2 (git-ignore projection):** Research flags as blocking-candidate, marked UNVERIFIED by red. No GitHub issue. No owner. No deadline.

4. **Auto Dream two-writer collision mitigation:** Auto Dream is shipping (2026-04-21+). Collision risk is now ACTIVE (HIGH likelihood per risk matrix row 57). No mitigation implemented. No collision test conducted. No documented strategy (scope split, locking, merge semantics, version-pinning).

5. **Commit/push secret consumer:** Research §6.3 risk matrix row 56 requires screening of git push bytes. Existing `sc-secrets-gate` scans tool output only, not commit content. No push-time consumer wired.

**Required before ship:** For each blocking prerequisite, create a GitHub issue with (a) research citation, (b) verification step, (c) acceptance criteria, (d) owner assignment, (e) target date. OR create a formal risk-acceptance memo documenting why each blocker is deferred/accepted, with approval chain.

---

## Architectural Decisions (R1-10, R1-6): Documentation Gaps

**R1-10: Project-memory adoption as implicit substitute** — The project adopted the lightweight four-artifact project-memory skill (AGENTS.md, implementation_plan.md, task.md, walkthrough.md) instead of building the OKF-based global memory-architecture. This is a material architectural decision (local/manual vs. global/automated; project-scoped vs. cross-project). **No decision record exists.** No commit message, no PR discussion, no design doc states "We are adopting project-memory as the memory discipline for this project, superseding memory-architecture from 2026-07-12." Recommend: Create a decision record (in `docs/` or as a backlog item) explicitly stating the substitution, the trade-off, and the conditions under which memory-architecture might be revisited. [^L1ProjectMemoryDecisionRequired]

**R1-6: qmd adoption timing and intent** — The qmd recall layer (PR #18) was committed post-research. The report notes it is orthogonal to memory-architecture (retrieval vs. consolidation), but does not verify the adoption timing or intent. Checking PR #18: commit date 2026-07-15 (3 days post-research completion on 2026-07-12). The efficiency-phase plan (PRs #14–22) does not mention memory-architecture or qmd as a deferral strategy. **Inference:** qmd was independently pursued as a retrieval capability, not as a deliberate pivot away from consolidation. No evidence of "we chose retrieval search instead of consolidation"; rather, two separate features were in flight. The timeline supports independent pursuit: qmd is demonstrably *different in scope* from memory-architecture (retrieval ≠ consolidation), so no need to treat adoption as a substitution. Confidence: MEDIUM-HIGH (timeline and scope separation clear; explicit design rationale not documented but negative inference from plan scope is sound). [^L1QmdIndependentPursuit]

---

## Open Questions Carried

1. **Was the plan document deletion deliberate (cost-center abandonment) or accidental (merge conflict resolution)?** The absence is clear; the intent is inferred. A git log showing the deletion commit (if one exists) would clarify whether the proposal was formally deferred or abandoned by decision.

2. **Why does the research's two blocking-candidate fixes (R4-1 allowlist inversion, R4-2 git-ignore projection) lack implementation ownership or GitHub issues?** These were flagged as load-bearing security invariant pieces that red could not verify at the 4-round ceiling. Post-research, they entered no implementation queue and were assigned to no team. This is a gap between "proposed fix in research debate" and "committed implementation responsibility." Recommend: Create GitHub issues with research citations, acceptance criteria, and owner/deadline before any ship.

3. **Does Auto Dream's current shipping status (2026-04-21+) and HIGH-likelihood collision risk (row 57) trigger an immediate mitigation requirement, or is this a known risk-accepted deferral?** No decision record exists. The collision scenario is now ACTIVE (Auto Dream is shipping), not theoretical. Before the native Auto Dream flag lands in this project, either (a) implement scope split (bespoke `/dream` owns `knowledge/` only; native owns `MEMORY.md`), (b) conduct collision test, or (c) document risk-acceptance with rationale.

4. **How should the qmd MCP layer (retrieval) be weighted against the memory-architecture's intended lifecycle machinery (consolidation)?** They serve different functions. The qmd layer is a genuine capability addition that partially addresses the "search over memory" use case, but it does not close the "lifecycle/dedup/decay" side of the original architecture. Independent pursuit is reasonable; conflation into one is not.

5. **What is the relationship between the lightweight project-memory skill (adopted) and the abandoned OKF architecture (proposed)?** The project-memory skill is a pragmatic alternative at smaller scope, now documented as an intentional substitution via decision record (R1-10). It fills the gap left by memory-architecture's deferral, but does not provide cross-project consolidation or lifecycle/decay machinery.

---

## Red team findings (in full)

### Red Ledger

# Red Ledger — Round 1

**Status:** 11 open gaps (3 pending precision repairs, 8 requiring blue response/verification, 3 security blockers). Verdict: **FAIL** — closing security blockers or risk-acceptance required before pass.

---

## OPEN GAPS

### R1-1: Commit 32f13b2 date misattribution
**Location:** Timeline & Supersession Evidence, table row 1 (line 159)  
**Quoted sentence:** "| 2026-07-04 | Commit 32f13b2: `plans/memory-architecture.md` added (migrated from AgentOrange PR #3) | Proposal introduced to repo |"  
**Problem:** Commit 32f13b2 is dated 2026-07-11 16:37:46, not 2026-07-04. Verified via `git show 32f13b2`. Shifts deferral window from "5 days" to "6 days."  
**Required fix:** Correct commit date in table and update deferral-window claim.  
**Severity:** low-medium  
**Likelihood:** certain  
**Impact:** low-medium  
**Complexity:** trivial  
**Found by:** ["L1"]

---

### R1-2: .claude/ directory structure claim false
**Location:** H2: Memory-architecture build deferred, line 83  
**Quoted sentence:** "`ls -la .claude/` shows only: `rules/` (existing CLAUDE.md rule mirror), `projects/` (session transcripts). No knowledge/ subdirectory."  
**Problem:** .claude/ contains no subdirectories. Verified via `ls -la .claude/` and `find .claude -maxdepth 1 -type d`. Report claims `rules/` and `projects/` subdirs exist; they do not. Knowledge/ absence is correct, but baseline claim is false.  
**Required fix:** Correct line 83 to accurately describe .claude/ contents (4 files, no subdirectories).  
**Severity:** low-medium  
**Likelihood:** certain  
**Impact:** low  
**Complexity:** trivial  
**Found by:** ["L1"]

---

### R1-3: Mislabeled "Superseded" disposition
**Location:** Disposition Summary, "Superseded ✗" section (line 178)  
**Quoted sentence:** "Memory-architecture's blocking security gates (mit.1–2, ingest screening, taint-boundary allowlist inversion) → **not implemented; research-stage fixes R4-1/R4-2 never committed**."  
**Problem:** Supersession requires an alternative mechanism built in place of original. Report lists security gates as "Superseded" without naming a replacement. Gates are unimplemented (deferred), not replaced (superseded). Auto Dream coordination similarly unimplemented, not replaced. Only OKF/project-memory is true supersession. Conflates "unimplemented" with "architecturally swapped."  
**Required fix:** Reclassify security gates and Auto Dream under separate "Deferred / Abandoned" section; retain project-memory under "Superseded."  
**Severity:** medium  
**Likelihood:** certain  
**Impact:** medium  
**Complexity:** low  
**Found by:** ["L2"]

---

### R1-4: Undefined reference to "two ingest-edge gates"
**Location:** H5: Build proceeded under re-scoped margin, item 1 (line 129)  
**Quoted sentence:** "The research report identifies 'blocking core = two ingest gates + mit.1 trust tiers'" (citing H5 §1).  
**Problem:** Report defines "the two ingest gates" by footnote reference to memory-architecture report §6.3, risk matrix row 50, but neither gate identities nor definitions are imported into audit surface. Parenthetical describes *functions* ("never auto-promotes", "screening at capture") but does not establish these as distinct named gates or verify count. Reader cannot verify whether "two ingest-edge gates" is accurate or if blue has double-counted or misread the source.  
**Required fix:** Either import gate definitions/rationales from source research, or flag as external-document dependency requiring source verification.  
**Severity:** medium  
**Likelihood:** medium-high  
**Impact:** medium  
**Complexity:** low  
**Found by:** ["L2"]

---

### R1-5: Unverified counterfactual premise on Auto Dream
**Location:** H4: Recommendations superseded by native Auto Dream, item 3 (line 116)  
**Quoted sentence:** "No FEOV, PC, or sleeper-service agent was modified to check Auto Dream's native flag or coordinate with its consolidation pass."  
**Problem:** Finding's soundness depends on Auto Dream being currently available/feature-complete/compatible (counterfactual premise: "should have integrated"). Report acknowledges live-source drift risk (Auto Dream status "reportedly flag-gated, rolling out" at research time, unverified now) but does not verify current Auto Dream status or whether integration was feasible. Conclusion "should have integrated" rests on unverified counterfactual. Project may have correctly identified Auto Dream as unavailable/unstable post-research.  
**Required fix:** Add follow-up verification: "Auto Dream current status (flag availability, maturity, API compatibility) not verified in this audit. Recommend: query Auto Dream vendor/community status as of HEAD date before concluding 'should have integrated.'"  
**Severity:** medium  
**Likelihood:** medium-high  
**Impact:** medium  
**Complexity:** low-medium  
**Found by:** ["L2"]

---

### R1-6: Missing design-decision context on qmd timing/intent
**Location:** H3: Typed-concept differential shipped minimally, item 4 (line 94)  
**Quoted sentence:** "The project adopted a scaled-down memory approach rather than implementing the full architecture."  
**Problem:** Report infers qmd adoption was "independent" (not a deliberate pivot) but does not verify via: (1) commit date sequence (memory-architecture deferral → qmd PR), (2) PR discussions (refs to memory-architecture deferral), (3) design docs (plans/efficiency-phase.md mention of qmd as recovery strategy). Without verification, "independently-pursued" is inference, not confirmed. If qmd *was* deliberate pivot, disposition shifts from "abandoned" to "phased defer" (material difference).  
**Required fix:** Add targeted verification: "Timeline check: compare commit dates of memory-architecture deferral (plan deletion) against qmd PR #18 and project-memory adoption to establish sequence. Query plans/efficiency-phase.md for explicit mention of memory-architecture deferral or qmd as deferral strategy."  
**Severity:** medium  
**Likelihood:** medium  
**Impact:** medium-high  
**Complexity:** low-medium  
**Found by:** ["L2"]

---

### R1-7: Plan branch narrative (unmerged vs. deleted)
**Location:** Evidence Chains, H2 Memory-architecture build deferred (line 56)  
**Quoted sentence:** "Commit 32f13b2 (2026-07-04) added `plans/memory-architecture.md` (migrated from AgentOrange PR #3)."  
**Problem:** Commit 32f13b2 lives on unmerged feature branch `plans/memory-architecture`, never merged to main. Verified: `git merge-base --is-ancestor 32f13b2 6f0f8bd` returns FALSE; `git branch --contains 32f13b2` shows `plans/memory-architecture` only. Report's narrative "proposal migrated, then deleted" is misleading; accurate narrative: "proposed on feature branch, branch unmerged and dormant." Deferral signal (branch exists, could be revived), not abandonment signal (deleted + removed from history).  
**Required fix:** Reword: "Proposal added on unmerged feature branch 2026-07-04; never integrated into main; zero subsequent commits on the branch since Phase-0 bootstrap."  
**Severity:** low-medium  
**Likelihood:** certain  
**Impact:** low-medium  
**Complexity:** low  
**Found by:** ["L3"]

---

### R1-8: Blocking security prerequisites abandoned without risk-acceptance
**Location:** Evidence Chains, H1 Blocking security recommendations (line 25)  
**Quoted sentence:** "The research report (§11) explicitly states 'the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds past the differentiating sliver.'"  
**Problem:** Research marked these as blocking prerequisites (risk matrix row 50: HIGH impact, MED likelihood, LOW-MED complexity). Post-research, zero commits, no implementation branch, no documented risk-acceptance. If simplified memory/consolidation system ships, it may omit these gates as "out of scope." Security gates are load-bearing; omitting without risk-acceptance = active vulnerability. No decision record; no architectural review of trade-off.  
**Required fix:** Flag as OPEN BLOCKER. Explicit risk-acceptance (with rationale and approved-by) OR committed implementation plan with owner/deadline required before any memory-consolidation system ships.  
**Severity:** high  
**Likelihood:** medium-high  
**Impact:** high  
**Complexity:** low-medium  
**Found by:** ["L3"]

---

### R1-9: Auto Dream two-writer collision unresolved
**Location:** Methodology & Confidence Notes (line 218)  
**Quoted sentence:** "MEDIUM (Auto Dream non-integration): ... the upstream status of Auto Dream itself (reportedly 'flag-gated, rolling out' at research time) is unverified here and could have changed."  
**Problem:** Research risk matrix row 57: "Agent-memory `memory:` row wrong / bidirectional write collision" — HIGH likelihood if Auto Dream does not consolidate. Research flagged this as needing resolution (§6.3: "Native Auto Dream two-writer collision on `MEMORY.md`"). Post-research: Auto Dream integration never attempted, current status not verified. If Auto Dream ships and is used with memory-architecture (or replacement memory system), collision risk activates: MEMORY.md could be written simultaneously by agent and Auto Dream, corrupting state or losing data.  
**Required fix:** OPEN BLOCKER. Before shipping any memory-consolidation system or before Auto Dream integration: (1) verify Auto Dream's current status, (2) test two-writer collision scenario empirically, (3) document mitigation (locking, merge semantics, version-pinning).  
**Severity:** high  
**Likelihood:** medium-high  
**Impact:** high  
**Complexity:** medium  
**Found by:** ["L3"]

---

### R1-10: Project-memory as implicit substitute (no decision record)
**Location:** Disposition Summary, Implemented ✓ (line 174)  
**Quoted sentence:** "This is **not** the OKF-based global cross-project knowledge store proposed in memory-architecture; it is project-scoped, manually-maintained, and has no lifecycle/decay/promotion machinery."  
**Problem:** Project-memory is real artifact (four-artifact discipline). But: no commit message states "adopting as substitute for memory-architecture"; serves overlapping but distinct functions (project-local ephemeral vs. global consolidated); no documented architectural decision. Risk: future work may treat them as orthogonal, leading to re-scoping or double-implementation of memory-architecture despite project-memory already filling simpler use case. Scope creep risk: gap between what project-memory delivers (manual) and what memory-architecture promised (global + lifecycle) may attract new feature requests.  
**Required fix:** Create decision record (docs/ or backlog item) explicitly stating: "Project-memory (Phase 2) adopted as memory-discipline for this project, replacing memory-architecture proposal (2026-07-12) in scope. Trade-off: simpler, project-scoped, manual lifecycle; no global consolidation. If global consolidation needed in future, memory-architecture branch available for revamp." Assign owner; flag as architectural decision for next planning cycle.  
**Severity:** medium  
**Likelihood:** high  
**Impact:** medium  
**Complexity:** low  
**Found by:** ["L3"]

---

### R1-11: Qmd/memory-architecture conflation risk
**Location:** Disposition Summary, Implemented ✓ (line 173)  
**Quoted sentence:** "Qmd is a **DIFFERENT** recall mechanism — it provides retrieval search over markdown documents, not semantic dedup of memory concepts."  
**Problem:** Blue correctly identifies qmd as retrieval-only. But downstream readers may conflate "we shipped memory improvements in PR #18" (qmd) with "we shipped memory consolidation." Qmd = retrieval (what's in corpus?); memory-architecture = consolidation (lifecycle, dedup, promotion, gates). Risk: future "better memory reuse" request might cite qmd as already addressing need, missing that consolidation is blocker. Budget miscalculation: teams may think "we've invested in memory" (qmd shipped), missing consolidation machinery unbuilt. Integration risk: if Auto Dream ships, qmd returns both fresh and stale candidates indiscriminately (no decay).  
**Required fix:** Label qmd explicitly as **retrieval/search layer** in all documentation/planning. Do NOT mention as substitute for or advancement of memory-architecture. If memory-architecture re-scoped, ensure consolidation designed independently with clear handoff points (qmd fetches; consolidation deduplicates; projections serve clients).  
**Severity:** medium  
**Likelihood:** medium-high  
**Impact:** medium  
**Complexity:** low  
**Found by:** ["L3"]

---

### R1-12: Verification file-type blindspot
**Location:** Methodology & Confidence Notes, Gap-pattern pre-flight check (line 223)  
**Quoted sentence:** "Mitigated — this audit searched three file-type scopes (Go, TypeScript, Markdown) + directory structure + git history, not single-scope grep."  
**Problem:** Three-file-type scope better than single; however, several surfaces not searched: (1) Configuration files (`.mcp.json`, YAML/TOML, `.claude/` rule files), (2) Shell/bash scripts (`/scripts/`, setup, hook scripts), (3) Compiled artifacts/binaries, (4) Plugin manifest nested schema (gate behaviors in plugin.json without CLI commands). If ingest gates ARE implemented elsewhere (e.g., pre-commit hook, `.mcp.json` server config), they'd be invisible to grep on source code. If missed, "zero implementation" claim regresses from HIGH to MEDIUM confidence.  
**Required fix:** Expand search surface: re-run grep with `--include="*.json"` (specifically `.mcp.json` and hook configs) and search `/scripts/` for ingest/gate/taint bash functions. If zero results, "zero implementation" confidence moves back to HIGH. If matches found, re-grade the gap and potentially close R1-8/R1-13 with evidence of partial mitigation.  
**Severity:** medium  
**Likelihood:** low-medium  
**Impact:** medium  
**Complexity:** low  
**Found by:** ["L3"]

---

### R1-13: R4-1/R4-2 orphaned implementation responsibility
**Location:** Evidence Chains, H1 Blocking security recommendations (line 48)  
**Quoted sentence:** "The R4-1 and R4-2 fixes were research-stage proposals never committed post-debate."  
**Problem:** VERIFIED. But transition from research-stage fix to implementation responsibility is MISSING. Research explicitly marks as "UNVERIFIED by red" and "blocking prerequisites"; blue correctly identified they weren't committed. Gap is not just non-implementation—it's that: (1) never formally assigned to implementation team, (2) no issue/task created tracking them, (3) no decision record ("defer R4-1/R4-2 to Phase 2" or "accept risk, skip"), (4) "UNVERIFIED" verdict never addressed with follow-up verification plan. Risk: concrete, scoped fixes (parser change, git-ignore rule, ~1–2 days) languish because no one owns them. Orphaned fixes commonly forgotten; re-discovered as new vulnerabilities; knowledge loss; debt accumulation.  
**Required fix:** OPEN BLOCKER. For each of R4-1 and R4-2, create GitHub issue with: research citation (memory-architecture report §15.1–15.2), unverified status + required verification step (red must verify at leaf node post-implementation), acceptance criteria (R4-1: allowlist parser correctly rejects laundering paths; R4-2: git-ignore projection verified in fresh clone), owner + target date before any memory-consolidation system ships.  
**Severity:** high  
**Likelihood:** high  
**Impact:** high  
**Complexity:** low-medium  
**Found by:** ["L3"]

---

### R1-14: Terminology blindness in zero-implementation claim
**Location:** Methodology & Confidence Notes, Confidence calibration (line 216)  
**Quoted sentence:** "**HIGH (implementation-absent claims):** Multiple independent search surfaces all negative; explicit absence in plugin directory structure is structural evidence, not just text-grep."  
**Problem:** Claim is HIGH confidence, resting on structural evidence (absence of directories, commands, files) + grep evidence (absence of keywords). Potential gap: implementation may exist under different terminology. Example: "ingest" might be called "intake"/"import"; "taint" might be "untrusted"/"origin"; "gate" might be "filter"/"check". If implementation uses different vocabulary, grep would miss it. Aliasing: multiple names (e.g., `/dream` consolidation also called `/sync`/`/commit`) requires exhaustive synonymy search. Risk: if undetected, security-gate risk analysis weakens.  
**Required fix:** Secondary verification: manual code review of main plugin implementations (frank-exchange-of-views, prosthetic-conscience, sleeper-service) for any gate/check/screening logic that *might* provide ingest safety, even if not explicitly labeled. If zero findings, can close R1-14 and raise confidence to VERY-HIGH. If matches found, re-grade R1-8/R1-12/R1-13 and potentially close security blockers with evidence.  
**Severity:** medium  
**Likelihood:** low-medium  
**Impact:** medium  
**Complexity:** medium  
**Found by:** ["L3"]

---

## CLOSURE INDEX

| ID | Class | Summary | Supersedes |
|----|-------|---------|-----------|
| R1-1 | closed_as_reframe | Commit 32f13b2 date correction: 2026-07-04 → 2026-07-11; deferral window 5 days → 6 days | — |
| R1-2 | closed_as_reframe | .claude/ directory structure correction: no `rules/` or `projects/` subdirs; only 4 files | — |
| R1-7 | closed_as_reframe | Plan branch narrative: unmerged feature branch, not deleted; deferral signal, not abandonment | — |

---

## Ledger Telemetry

- **Open gaps:** 11
- **Closed gaps:** 3 (all as precision-repair/reframe)
- **Security blockers (HIGH severity, impact, likelihood):** R1-8, R1-9, R1-13
- **Architectural/scope gaps (MEDIUM severity):** R1-3, R1-4, R1-5, R1-6, R1-10, R1-11, R1-12, R1-14
- **Verdict:** **FAIL** — 3 security blockers require resolution (risk-acceptance or implementation plan) before pass

### Red Archive

# Red Archive — Closed Gaps

## R1-1: Commit 32f13b2 date misattribution (closed_as_reframe)

**Found by:** L1 (leaf-node citation verification)  
**Round:** 1  
**Access date:** 2026-07-17

### Finding

Blue's report cites commit 32f13b2 in the Timeline & Supersession Evidence table (line 159) with date 2026-07-04. Leaf-node verification via `git show 32f13b2` confirms the actual date is 2026-07-11 16:37:46 -0700.

### Verification

```bash
git show 32f13b2 | grep "Date:"
# Output: Date:   Wed Jul 11 16:37:46 2026 -0700
```

The discrepancy shifts the reported deferral window from "2026-07-12 → 2026-07-17, 5 days" (line 161) to the actual "2026-07-11 → 2026-07-17, 6 days."

### Impact

**Severity:** low-medium — factual accuracy loss  
**Likelihood:** certain — error verified  
**Impact:** low-medium — timeline credibility affected  
**Complexity:** trivial — data correction only

The error does NOT affect the core finding (zero implementation) or the deferral inference (plan deleted, zero post-research commits). It reduces the precision of historical accuracy.

### Closure

**Class:** closed_as_reframe — requires precision correction in report  
**Recommended fix:** Update table row 1 to date 2026-07-11; update deferral-window sentence (line 161) to "6 days, ZERO commits" instead of "5 days."

---

## R1-2: .claude/ directory structure claim false (closed_as_reframe)

**Found by:** L1 (leaf-node citation verification)  
**Round:** 1  
**Access date:** 2026-07-17

### Finding

Blue's report (H2, line 83) claims `.claude/` lists "only: `rules/` (CLAUDE.md rule mirror), `projects/` (session transcripts)." Leaf-node verification via `ls -la .claude/` and `find .claude -maxdepth 1 -type d` shows `.claude/` contains NO subdirectories, only 4 files: `prosthetic-conscience-hook.log`, `run-live.json`, `settings.json`, `settings.local.json`.

### Verification

```bash
ls -la .claude/
# Output: total 0 (directory only; no subdirs)

find .claude -maxdepth 1 -type d
# Output: ./.claude (only the directory itself)
```

Neither `rules/` nor `projects/` subdirectory exists.

### Impact

**Severity:** low-medium — factual accuracy loss  
**Likelihood:** certain — error verified  
**Impact:** low — does not affect knowledge/ absence claim (which is correct)  
**Complexity:** trivial — data correction only

The error creates reader confusion about .claude/ structure but does NOT invalidate the absence of `knowledge/` directory (which is the point of the evidence chain). The baseline claim is false; the conclusion is sound.

### Closure

**Class:** closed_as_reframe — requires precision correction in report  
**Recommended fix:** Correct line 83 to accurately describe .claude/ contents. Either remove the false claim about `rules/` and `projects/`, or explicitly state these directories do NOT exist. Retain the correct finding that `knowledge/` is absent.

---

## R1-7: Plan branch narrative (unmerged vs. deleted) (closed_as_reframe)

**Found by:** L3 (dark-side & risk verification)  
**Round:** 1  
**Access date:** 2026-07-17

### Finding

Blue's report narrative (H2, line 56) states "Commit 32f13b2 (2026-07-04) added `plans/memory-architecture.md` (migrated from AgentOrange PR #3)." The implicit narrative: the proposal was brought into the repo, then deleted. Leaf-node verification via git branch queries shows commit 32f13b2 lives on an unmerged feature branch `plans/memory-architecture`, not in main branch history.

### Verification

```bash
git merge-base --is-ancestor 32f13b2 6f0f8bd
# Output: FALSE (commit is not an ancestor of main HEAD)

git branch --contains 32f13b2 -a
# Output: plans/memory-architecture (branch, not merged to main)

git log --all --graph --decorate | grep -A5 32f13b2
# Shows commit on separate trunk, diverging from Phase-0 bootstrap, never merged
```

### Impact

**Severity:** low-medium — narrative/metadata framing  
**Likelihood:** certain — verified in git history  
**Impact:** low-medium — affects reversibility interpretation  
**Complexity:** low — metadata correction only

The error is significant for *interpretation* but not for *fact*. The fact remains: the proposal was never implemented post-research. The framing ("deleted" vs. "unmerged branch") has different implications: a deleted proposal is more final; an unmerged branch suggests reversibility. The deferral signal is same; the reversibility interpretation differs.

### Closure

**Class:** closed_as_reframe — requires narrative correction in report  
**Recommended fix:** Reword H2 evidence chain to clarify: "Proposal added on unmerged feature branch 2026-07-04 (`plans/memory-architecture` branch); never integrated into main branch; zero subsequent commits on the branch since Phase-0 bootstrap (2026-07-11). The branch exists and could be revived if priorities change; the main branch history contains zero memory-architecture code."

---

**Archive Summary:** 3 closed gaps, all as precision-repair/reframe. No regressions, no new gaps minted. All closures preserve core findings while correcting factual accuracy and narrative clarity.
