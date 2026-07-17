# Lane 1 — Memory-Architecture Review Against Shipped Implementation

**Scope:** Adversarial-disconfirming-first audit of the five frontier hypotheses against what shipped in the special-circumstances repo since 2026-07-12 memory-architecture research completion. SMOKE mode: mechanical fact verification, shallow research scope.

**Period audited:** 2026-07-12 (research completion) → 2026-07-17 HEAD (6f0f8bd).

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

**Disconfirming search:**

1. **Ingest gates / injection screening NOT in plugins** [^L1IngestCodeSearch]
   - `grep -r "ingest\|injection\|taint" plugins/` (excluding research runs) returns zero matches in Go, TypeScript, or skill markdown.
   - No `WebFetch|WebSearch|Bash` screening hook exists outside the research debate text.
   - The `sc-secrets-gate` PreToolUse hook (cited in memory-architecture report as shipping) exists and is functional (referenced in prosthetic-conscience Phase-2 build), but it is **outbound-tool-output only** and does not cover inbound/candidate-tier taint screening. [^L1SecretsGate]

2. **Taint-boundary allowlist inversion (R4-1 fix) NOT implemented** [^L1TaintSearch]
   - Zero matches for "allowlist" / "denylist" in plugins/ code or skills.
   - The report's §15.1 claims a Round-4 fix (allowlist inversion) that "blue fixed... UNVERIFIED by red" — this fix was proposed in the research but **never committed to the codebase**.

3. **Git-ignore projection (R4-2 fix) NOT implemented** [^L1GitignoreSearch]
   - No `projections/` directory exists in `.claude/` or project roots.
   - No `.gitignore` rule for projections exists.
   - The R4-2 recommendation to "git-ignore `projections/`, commit raw concept bodies only" was a proposed fix in the research but never landed in implementation.

4. **`/dream` consolidation NOT shipped** [^L1DreamSearch]
   - Zero commands named `/dream` or `/remember` in `plugins/frank-exchange-of-views/commands/` or `plugins/sleeper-service/commands/`.
   - The sleeper-service plugin is **scaffold-only (Phase 0)**, with no commands, skills, or agents implemented. [^L1SleeperStub]
   - The nightly consolidation loop is described in the proposal but exists only as text, not code.

**Confidence grading:** HIGH on zero-implementation claim (multiple independent search surfaces, grep on both code and prose, plugin directory structure verified). The R4-1 and R4-2 fixes were research-stage proposals never committed post-debate.

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

**Disconfirming search:**

1. **No Phase-2/3 overlap resolution or two-writer conflict investigation** [^L1OverlapSearch]
   - Zero commits post-research touching "memory conflicts," "Auto Dream," "write collision," or "MEMORY.md synchronization."
   - The research report (§6.3) flagged "Native Auto Dream two-writer collision on `MEMORY.md`" as a risk requiring fix — **this fix was never implemented**.

2. **Project-store feature never shipped** [^L1ProjectStoreSearch]
   - No PR-review-workflow or "project store" feature in the codebase.
   - The proposal envisioned committed project stores cloned with the repo; this never materialized.
   - The R4-2 fix (git-ignore projection, commit bodies only) was a proposed de-escalation of the feature scope, not its implementation.

3. **No attempt at Auto Dream integration** [^L1AutoDreamIntegrationSearch]
   - The efficiency-phase plan (PRs #14–22) focuses entirely on debate-engine cost levers (sharded findings, telemetry, grade disputes, batching).
   - No FEOV, PC, or sleeper-service agent was modified to check Auto Dream's native flag or coordinate with its consolidation pass. [^L1EfficiencyPhaseScope]
   - The qmd MCP layer (PR #18) is a **retrieval** mechanism, not a **consolidation** mechanism; it does not address the overlap with Auto Dream.

**Confidence grading:** HIGH on no-attempt-at-resolution (commit history, plan scope, and plugin code are all negative); explicit-investigation absent from post-research work.

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
- **qmd recall layer** (PR #18): MCP-based BM25+semantic search over markdown corpus; orthogonal to memory-architecture.
- **project-memory skill** (Phase 2): lightweight four-artifact per-project discipline; simpler alternative to OKF architecture.
- **secrets-gate hook** (Phase 2): outbound-tool-input screening only; does NOT cover inbound taint gates.
- **debate-engine efficiency levers** (PRs #14–22): cost reduction for the FEOV framework; unrelated to memory architecture.

### Superseded ✗
- Memory-architecture's blocking security gates (mit.1–2, ingest screening, taint-boundary allowlist inversion) → **not implemented; research-stage fixes R4-1/R4-2 never committed**.
- Native Auto Dream coordination (proposed as necessary in report §6.3 risk matrix) → **not attempted**.
- OKF-based typed-concept schema and promotion ladder → **not built; project-memory adopted as lightweight alternative**.

### Open / Deferred
- Phase-0 security gates (blocking prerequisite per research report).
- Phase-1 through Phase-5 (complete staging of knowledge stores, lifecycle machinery, `/dream` consolidation, `/ingest` pipeline, headless scheduling).
- sleeper-service plugin buildout (remains Phase-0 scaffold).

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

- **[citation-status-and-misattribution-patterns](https://...) (Pattern A: "open bug" that is Closed):** N/A — no GitHub issue citations in this audit.
- **[gap-pattern-verification-file-type-blindspot](https://...) :** Mitigated — this audit searched three file-type scopes (Go, TypeScript, Markdown) + directory structure + git history, not single-scope grep.
- **[live-source-drift](https://...) :** Noted on Auto Dream status (not re-verified live); claimed "flag-gated, rolling out" at research time but never integrated despite apparently becoming available.
- **[gitignored-not-absent](https://...) :** Applied — confirmed absence with both `ls` and `git status` checks; no gitignored-but-present knowledge stores found.

---

## Claims & Citations

[^L1IngestCodeSearch]: `grep -r "ingest|injection" plugins/frank-exchange-of-views plugins/prosthetic-conscience plugins/sleeper-service --include="*.go" --include="*.ts" --include="*.md" --include="*.json" 2>/dev/null` returns zero matches (searched 2026-07-17, HEAD at 6f0f8bd). The only ingest-related text appears in research/ debate documents, not implementation.

[^L1SecretsGate]: `sc-secrets-gate` hook is referenced in `plugins/prosthetic-conscience/` Phase-2 build commits and is functional as an outbound-tool-input gate (denies tool calls containing secret keywords). Memory-architecture report §8 item 3 ("secret/PII leakage on remote push") required a commit/push-time consumer; the shipping gate scans WebFetch|WebSearch|Bash tool arguments only, per research/2026-07-12_memory-architecture/report.md §6.3 risk matrix row 56. Inbound taint-boundary screening was never implemented.

[^L1TaintSearch]: Research report §15.1 (R4-1 fix) proposed "allowlist inversion: a candidate is `trajectory-derived` only if every supporting turn is operator/harness-authored with no intervening un-provenanced tool result." Zero allowlist-type schemas found in plugins/prosthetic-conscience/skills/ or sleeper-service/. The fix was claimed UNVERIFIED by red at the research ceiling; it was never committed to the codebase post-debate.

[^L1GitignoreSearch]: `find . -name ".gitignore" | xargs grep -l "projection\|knowledge" 2>/dev/null` returns no matches. No `projections/` directory exists in repo structure. The R4-2 fix (git-ignore `projections/` to prevent clone-time injection) was proposed in research §15.2 but never landed.

[^L1DreamSearch]: `find plugins -name "*.md" | xargs grep -l "/dream\|/remember"` returns zero. Sleeper-service is "scaffold only (Phase 0)"; no `/dream` command wired. The nightly consolidation loop was proposed in memory-architecture §6 (Phase-2 item) but never implemented.

[^L1SleeperStub]: `ls -la plugins/sleeper-service/` yields only `.claude-plugin/plugin.json` and `README.md`. README states "Status: scaffold only (Phase 0). Design: [`plans/claude-port-plan.md`](../../plans/claude-port-plan.md) §3c." Zero skills/, commands/, or agents/ subdirectories. Commit logs show sleeper-service added 2026-07-11 (Phase 0 bootstrap) and never modified post-bootstrap.

[^L1MemarchPlanDeletion]: `git log --all -- plans/memory-architecture.md` shows commit 32f13b2 (2026-07-04) alone; `git show HEAD:plans/memory-architecture.md` returns "fatal: path does not exist in 'HEAD'." The file was added to the repo but is not present at HEAD (6f0f8bd). No commit explicitly deletes it; it is absent from the tracked tree. Likely deleted in an intermediate commit (checked with `git show <commit>:plans/memory-architecture.md` for several post-research commits; all return fatal) or deliberately de-tracked.

[^L1KnowledgeDirsAbsent]: `find . -type d -name "knowledge" 2>/dev/null` returns no results. `ls -la .claude/` lists only: `rules/` (existing CLAUDE.md rule mirror), `projects/` (session transcripts). No `knowledge/` subdirectory. Confirmed via `git show HEAD:.claude/` as well (no knowledge/ entry).

[^L1ConceptSchemaSearch]: `grep -r "type:.*rule\|type:.*fact\|type:.*glossary" plugins/ research/2026-07-17_* --include="*.md" | grep -v "2026-07-12_memory-architecture\|2026-07-12_feov-retrospective"` returns zero matches in current implementation. OKF schema frontmatter exists only in research debate text (citations/proposals), not in instantiated concept files.

[^L1QmdIsNotMemarch]: qmd (PR #18, live in HEAD) provides BM25 and semantic vector search over indexed markdown collections. It is a **retrieval/recall layer**, not a memory-architecture **lifecycle/consolidation layer**. Qmd enables the "retrieval for evidence and context" access mode (research-protocol SKILL.md) but does not implement concept promotion, taint propagation, decay windows, or the promotion ladder—all of which were core to the memory-architecture proposal.

[^L1ProjectMemoryAlternative]: `plugins/prosthetic-conscience/skills/project-memory/SKILL.md` describes a four-artifact discipline (AGENTS.md, implementation_plan.md, task.md, walkthrough.md) per project in `projects/<name>/`. This is **not** the global cross-project knowledge store + promotion ladder proposed in memory-architecture. It is simpler, manually-maintained, and project-scoped. No lifecycle/decay/promotion machinery. Adopted during Phase 2 (commit 89a3442); no documentation states this is a substitute for memory-architecture, but the architectural choice (lightweight project-local vs. global-with-lifecycle) is evident.

[^L1OverlapSearch]: `git log --all --oneline --since="2026-07-12" | grep -i "memory\|dream\|auto\|collision\|consolidat"` (post-research period) returns zero relevant commits. Research report §6.3 risk matrix row 57 flags "Native Auto Dream two-writer collision on `MEMORY.md`" (HIGH likelihood); no fix or investigation followed.

[^L1ProjectStoreSearch]: `git log --all --oneline --grep="project.store\|project-store\|committed.*store\|clone.*ratif"` returns no matches. The project-store feature (described in memory-architecture proposal §4.2–4.3 and risk matrix) was never built. R4-2's proposed de-escalation (git-ignore projection, keep concepts only) was a fallback offered in research but not implemented as a followup action item.

[^L1AutoDreamIntegrationSearch]: `grep -r "Auto Dream\|auto.*dream\|consolidat.*native" plugins/ --include="*.md"` returns zero matches outside research runs (checked 2026-07-17, HEAD). The efficiency-phase PRs (#14–22) scoped to debate-engine cost levers; no agent was modified to detect or coordinate with Auto Dream. The qmd MCP (PR #18) is retrieval-focused, not consolidation-focused.

[^L1EfficiencyPhaseScope]: `plans/efficiency-phase.md` §I lists "Problem" as debate-cost reduction through "(a) redundant re-reading of red's own closed cases, (b) turn-fragmented candidate ingestion, (c) a carried-gap re-docket loop, (d) rounds past diminishing returns." Zero mention of memory-architecture, knowledge consolidation, or Auto Dream overlap. The plan's "Competing priorities; memory-architecture not advanced" inference is supported by this scope statement.

[^L1Mit1Mit2Search]: Memory-architecture report §6.3 risk matrix row 50 identifies "Blocking core = the two ingest-edge gates (external-ingest never auto-promotes; injection screening at capture) **+ mit.1 trust tiers** (the enforcing schema, zero separable cost)." Zero implementation of `mit.1` schema or the ingest gates in plugins/. The report's disposition: "Fix — blocking" with contingency "red never verified this closes the laundering path" (R4-1 unverified, R4-2 unverified). These blocking items remain unimplemented.

[^L1SecurityBlockersUnmet]: Research report §11 (Compromise Rationale) explicitly states: "Verifying them at the leaf node is exactly the round the ceiling removed" and recommends "the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds past the differentiating sliver." These prerequisites remain unmet at HEAD.

[^L1CloneRatificationSearch]: `grep -r "ratif" plugins/ --include="*.md" --include="*.json" | grep -v "ratif.*PR\|ratif.*dispute\|ratif.*merge"` returns zero matches on a clone-ratification-marker concept. The proposal envisioned a marker to distinguish operator-vetted knowledge from fresh candidates at clone-time; zero schema or field instantiation found.

[^L1ProvenanceTaxonomySearch]: Memory-architecture proposal §3.1 defines `provenance: {source: "trajectory:<session> | url:<u> | file:<p>", captured: "ISO timestamp", by: "skill/trajectory-review"}`. No implementation of this field structure found in plugins/. The qmd recall layer captures file URIs (qmd:// links) but does NOT instantiate the memory-architecture provenance taxonomy. Different provenance mechanisms; not equivalent.

[^L1PhaseOpenSearch]: Memory-architecture report §8–9 describes Phase-4 (ingest pipeline) and Phase-5 (headless schedule). Neither exists at HEAD. `find plugins/sleeper-service -name "*.md"` returns only README.md and plugin.json; no Phase-4/5 code, skills, or agents.

