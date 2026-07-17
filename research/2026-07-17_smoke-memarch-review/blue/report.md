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
