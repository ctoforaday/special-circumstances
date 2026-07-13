# Blue lane 2 — dedup/consolidation as the Achilles heel (H2), then breadth

Lane scope: H2 first (LLM-driven expand-before-append without a semantic index silently loses or
fragments knowledge), then breadth across H1/H3/H4/H5. 20 web searches + 6 primary-source fetches;
disconfirming budget met (searches on file-memory success, LLM-judge dedup adequacy, files-win/YAGNI
criticism, idle-consolidation validation). Saturation: final searches returned already-seen sources.

---

## 1. H2 verdict: consolidation is the top technical risk — real, measured, but mitigable, and the proposal's mitigation is under-specified in two precise places

### 1.1 The failure is documented, not hypothetical

- **Continuous LLM rewriting of stored memories degrades them.** "Useful Memories Become Faulty
  When Continuously Updated by LLMs" documents that repeated LLM update cycles corrupt previously
  useful memories via interference, drift of the stored text's meaning, and loss of specifics; the
  degradation intensifies with update frequency.[^FaultyMemories] Related commentary reports memory
  utility that first rises then falls **below the no-memory baseline** under repeated consolidation
  (one reported figure: a frontier model failing 54% of ARC-AGI problems it had previously solved
  once consolidated memory was attached — reported in secondary commentary, not independently
  re-verified here).[^AgentsDumber]
- **The specific corruption mode is semantic intensification and summarization drift** — "likes
  mild spicy food" becomes "loves very spicy food" over rewrite cycles; each compression discards
  entity-level detail retrieval later depends on.[^MemorySurvey]
- **The OpenClaw "details unavailable" pattern generalizes**: stale, contradictory, and
  near-duplicate facts accumulate and degrade behavior *even when retrieval works*; the three
  drivers are context economics, entity drift, and index precision decay.[^HindsightConsolidation]

### 1.2 But the industry did not respond by abandoning consolidation — it converged on four write-time levers the proposal already has three of

The consolidation problem's standard treatment: **importance filtering at write time, merge with
entity/conflict resolution at write time, confidence decay over time (exponential preferred), and
eviction-by-unretrievability rather than deletion**.[^HindsightConsolidation] Production systems
implement exactly this shape: mem0's pipeline retrieves semantically similar existing memories,
then has an LLM choose ADD/UPDATE/DELETE/NOOP per candidate fact[^MemZero]; Zep/Graphiti compares
new edges against semantically related existing edges with an LLM to detect contradictions, then
*invalidates* (never deletes) superseded facts with validity windows[^ZepGraphiti]; Letta ships
"sleep-time agents" that consolidate, deduplicate, and prune memory blocks in the background while
the primary agent is idle.[^LettaSleep] The proposal's promotion ladder, supersedes-not-delete,
and decay table are the same levers. **H2 does not kill the design.**

### 1.3 The two precise under-specifications

**(a) Candidate retrieval is unspecified.** Expand-existing-before-append requires the consolidator
to *find* the overlapping concept first. The proposal says "search the target bundle" (§6) without
saying how. Evidence: lexical/title matching is systematically weak against paraphrase — semantic
methods beat lexical baselines by 11–20+ points on paraphrase detection, and near-identical
meanings routinely share almost no surface vocabulary (99%+ semantic similarity with single-digit
BLEU overlap).[^ParaphraseGap] LLM pairwise judgment of *given* candidate pairs is reliable at high
similarity but degrades sharply near the decision boundary (at cosine ≥0.95 every flagged pair is a
true duplicate; at 0.85–0.87 only ~1.5% are)[^LLMJudgeDedup] — so the binding constraint is
**recall of candidate pairs**, not judgment quality. At the suite's realistic scale
(hundreds of concepts, single operator), the whole per-type index fits in a consolidator's context,
so "read `index.md` + `description` lines for the whole bundle, then pairwise-judge" is adequate
**now** — but the design must *name* this as the candidate-retrieval mechanism and state the scale
ceiling at which the deferred SQLite/vector index (§11) stops being optional. Precedent for the
upgrade path exists: markdown-store-plus-SQLite-index hybrids are an established pattern
(basic-memory builds a semantic graph over plain markdown; sqlite-memory adds hybrid retrieval over
markdown).[^BasicMemory][^SqliteMemory]

**(b) Expand-existing invites the continuous-rewrite failure.** "Expand" as implemented by an LLM
re-emitting the whole concept file is exactly the repeated-rewrite loop that
[^FaultyMemories] shows corrupts memories. Fix is cheap and structural: **expansion appends — it
never rewrites the claim.** The concept body's core claim becomes effectively immutable after
promotion; corroboration appends to the Evidence section and bumps counters in frontmatter;
changing the *claim itself* requires `supersedes` (new file, old one deprecated), mirroring
Zep's invalidate-don't-mutate discipline.[^ZepGraphiti] This turns every consolidation diff into
additions + frontmatter bumps + whole-file supersessions — trivially reviewable, and drift-proof
because prose is never LLM-round-tripped in place.

### 1.4 The review-by-git-diff guard is weaker than the proposal assumes

The proposal leans on "every merge is a git diff a human can review" (§9.4). Measured behavior of
humans around agent-authored changes: in a large OSS sample, **61.4% of agent-authored pull
requests received no recorded review activity at all**, and 71.6% of review comments on them were
authored by other agents[^UnreviewedPRs]; rubber-stamping under queue pressure is the documented
failure mode.[^AIApprovingPRs] A nightly bot commit to a knowledge repo is the least-reviewed
artifact class there is. Consequence: the git-diff guard should be treated as *forensic* (undo
after harm is noticed), not *preventive*. Preventive guards must be structural: the
append-only-expansion rule (1.3b), hard caps on what a single dream pass may change
(e.g. max N supersessions per pass, else halt and flag), and the human gate at the
projection/skill promotion rungs — not diff review of routine passes.

---

## 2. Missing risk (absent from §9 entirely): memory poisoning — the store is an injection-persistence vector

§9.5 covers *outbound* leakage (secrets pushed to GitHub). The *inbound* threat is documented and
recent: **CVE-2026-21852** — a malicious npm postinstall appends instructions to Claude Code's
`MEMORY.md`; the harness loaded the first 200 lines with high authority every session; Anthropic's
fix (v2.1.50/v2.2) **removed user memories from the system prompt** to reduce their
authority.[^CiscoMemoryCVE][^OmegamaxCVE] Systematic studies of memory-poisoning attacks on
stateful agents confirm the pattern class: poison once via any untrusted input channel, exploit
across every future session.[^MemoryPoisoningStudy]

The proposal makes this *worse* than stock Claude Code in three ways:

1. **`/ingest <url|file|dir>` (§7.2/7.3) is a designed pipeline from untrusted external content
   into the store** — and the dream loop then *launders provenance*: a poisoned short-term note
   that survives to `knowledge/` carries a legitimate-looking `provenance` entry and gets projected
   into *every* session via `active.md` → `CLAUDE.md` `@`-import, which (unlike auto memory
   post-fix) still lands in context with instruction-like authority.
2. **`MEMORY.md` as "the inbox" (§5) automatically promotes the exact file the CVE targeted** into
   a durable, cross-session store.
3. **The nightly headless pass runs with no human present** — the one moment an operator might
   notice odd content is skipped by design.

**Required changes:** (i) trust-tier the `provenance.source` — concepts whose provenance chain
includes `url:` or third-party `file:` sources MUST NOT auto-promote to `active` or into any
projection without explicit human confirmation, ever (a permanent gate, not a confidence
threshold); (ii) the dream pass runs an injection-pattern scan (imperative-instruction detection,
"ignore previous", tool-invocation phrasing) over candidates before promotion, symmetric with the
outbound secret-scrub; (iii) projections render concepts as *reference knowledge*, not
instruction-voiced text, wherever possible — reduce the authority of the surface.

---

## 3. The competitive landscape moved: the harness itself is converging on this design (H5, revised)

- **Auto memory is now native and on by default** (v2.1.59+): Claude writes its own
  `~/.claude/projects/<project>/memory/MEMORY.md` index (first 200 lines / 25KB loaded per
  session) plus on-demand topic files; per-project, machine-local; `autoMemoryDirectory` is
  **configurable** in settings.[^ClaudeMemoryDocs]
- **"Auto Dream" — a native nightly consolidation — is rolling out behind a server-side flag**:
  a four-phase pass (orient → gather signal from session transcripts → consolidate: merge
  duplicates, resolve contradictions, absolutize dates → prune and re-index MEMORY.md under the
  200-line threshold), triggered at ~24h + >5 sessions since last run; community skills already
  replicate it.[^AutoDream][^DreamSkill] Availability is flag-gated and not universal — treat as
  direction-of-travel, verified as concept, unverified as a dependable API.
- **Per-subagent persistent memory exists natively** (v2.1.33+): `memory: user|project|local`
  frontmatter, stored at `~/.claude/agent-memory/<agent>/` or `.claude/agent-memory/<agent>/`,
  first 200 lines of the agent's MEMORY.md injected — but there is an open bug where the field is
  non-functional when a tools allowlist is present (issue #57507), so §5's agent-memory row is
  load-bearing on a currently-flaky feature.[^SubagentMemory][^SubagentMemoryBug]

**Consequences.** (a) *Validation*: Anthropic independently building trajectory-signal-gathering +
scheduled consolidation is the strongest available evidence that the proposal's core loop is the
right shape. (b) *Collision*: the proposal assigns `/dream` to read-and-prune `MEMORY.md` while
native Auto Dream consolidates the same file on its own clock — a two-writer conflict with no
coordination story. (c) *Scope transfer*: native machinery now covers per-project capture and
consolidation *without building anything*. The bespoke layer's defensible remit shrinks to what
native does not do: **cross-project global knowledge as a reviewable git repo; typed/schema'd
concepts; external-source ingest with provenance; human-gated promotion to skills; the project
store committed with the code**. The build plan should be re-scoped so phases duplicate nothing the
harness ships. Concretely: consider pointing `autoMemoryDirectory` *into* the store's
`short-term/` (making native capture the ingest mechanism), or scope `/dream` to `knowledge/` only
and let native Auto Dream own `MEMORY.md`, consuming its *output* as the inbox.

Survey of external alternatives (H5 remainder): mem0, Zep, Letta are service/daemon/database-bound
— they violate the suite's no-daemon, git-reviewable constraints on the same grounds that rejected
FUSE.[^MemZero][^ZepGraphiti][^LettaSleep] basic-memory is the closest external fit (local-first
markdown + MCP + semantic graph) but has no lifecycle/promotion/PR flow and adds an MCP server
dependency.[^BasicMemory] Nothing surveyed offers project-store-committed-with-code. **Bespoke
remains justified for the shrunken remit; no external adoption dominates.**

---

## 4. Substrate verification (H1): holds, with four leaf-level corrections

1. **OKF is real and as described**: v0.1 Draft in `GoogleCloudPlatform/knowledge-catalog`;
   `type` is the only required frontmatter field; `index.md`/`log.md` reserved; **custom
   frontmatter keys are explicitly permitted** ("Producers MAY include any additional keys"), so
   the §3.1 profile is spec-compliant, not a fork. Announced June 2026 — it is *four weeks old*;
   "external documented standard" is true but its ecosystem benefit is currently
   aspirational.[^OKFSpec][^OKFBlog]
2. **Transcripts**: confirmed JSONL at `~/.claude/projects/<encoded-path>/<session-id>.jsonl`,
   one JSON object per line, parseable — but the entry format **is internal and changes between
   releases**; parsers break on updates. §9.1's Phase-0 check must become a *pinned-version
   contract with a fallback* (e.g. degrade to `/export`), not a one-time
   confirmation.[^TranscriptFormat]
3. **`@`-imports**: confirmed, max depth 4; **imports pointing outside the project trigger a
   one-time approval dialog, and if declined, stay disabled silently** — the global
   `@~/.claude/knowledge/projections/active.md` import can be dead without any error surface.
   Projection health needs a SessionStart check, which the proposal's belt-and-suspenders hook can
   absorb.[^ClaudeMemoryDocs]
4. **A better projection surface exists that §5 never considers: `.claude/rules/`** — markdown
   rule files loaded with CLAUDE.md priority, **path-scoped via `paths:` frontmatter** so
   file-type-specific knowledge loads only when Claude touches matching files, and symlink-friendly
   for sharing.[^ClaudeMemoryDocs] Projecting `type: rule` concepts to
   `.claude/rules/knowledge-*.md` (with `paths` derived from concept tags) spends context only when
   relevant — directly addressing the context-budget problem (§6 below) — and keeps `CLAUDE.md`
   untouched by generated content.

Headless execution (feeds §9.2): `claude -p` waits for background subagents (10-min default cap,
tunable), but there is an open issue where **parallel Task fan-out hangs under non-TTY parents
(cron/scheduled contexts)** — precisely the dream loop's runtime.[^HeadlessDocs][^HeadlessHang]
The scheduled dream pass should be designed **sequential-subagent or single-agent**, with parallel
fan-out reserved for interactively-invoked `/dream` and `/memory-bootstrap`.

---

## 5. Cadence (H3): nightly-as-sweep is defensible; eager per-note LLM processing is not

- Generative-agents reflection is **importance-threshold-triggered** (accumulated importance >
  threshold), not clock-driven.[^GenerativeAgents]
- RecMem shows **eager consolidation** (LLM-processing every incoming item) wastes 77–87% of
  construction tokens versus recurrence-triggered consolidation, *with no accuracy gain from
  eagerness* — consolidate only when an item accumulates enough semantically similar
  neighbors.[^RecMem]
- Letta's sleep-time agents validate idle-time batch consolidation as a production
  pattern[^LettaSleep]; native Auto Dream's ~24h + >5-sessions trigger is itself a hybrid
  clock+threshold gate.[^AutoDream]

**Synthesis**: the proposal already has event-driven *capture* (Stop/PreCompact hooks) and
clock-driven *consolidation* — the right two-level shape. Two adjustments: (i) gate the nightly
pass on a threshold (skip when fewer than N new candidates — mirrors Auto Dream's >5-sessions gate,
saves cost and avoids no-op commits); (ii) inside the pass, do not LLM-elaborate every short-term
note — promote on recurrence (`review_count ≥ 2` is already the criterion; make the *processing*
lazy too, per RecMem). No change to the daily default is warranted.

---

## 6. Lifecycle arithmetic (H4): keep decay, keep counts, drop the confidence float

- **Decay earns its place**: it is "the lever most agent memory systems skip yet the one that
  matters most for long-running agents"[^HindsightConsolidation]; an empirically tuned importance
  half-life of ~29 days[^MemorySurvey] brackets the proposal's 14-day short-term / 60-day candidate
  windows — the guesses are in the evidenced band.
- **Confidence-as-LLM-assigned-float does not**: LLM-rated importance/confidence scores drift
  across model versions and add a model call per write[^MemorySurvey]; the one strong benchmark
  win for confidence-bearing memory (ALFWorld 59.9 vs 28.7) is for *belief distributions over
  uncertain conclusions in partially observable environments*[^BeliefMemory] — not this workload
  (curated operator knowledge). Replace `confidence: 0.0–1.0` with what is already observable and
  stable: `status` (ordinal) + `review_count` (count) + `last_seen` (date). Promotion rule becomes
  "`review_count ≥ 2` and seen within window" — same ladder, no pseudo-precise float to tune,
  nothing to drift when the underlying model changes. This also deletes §9.6 (threshold tuning) as
  a risk.
- **The projection needs a hard budget, not just a quality gate.** Instruction adherence degrades
  as always-on context grows — frontier models reliably follow roughly 150–200 instructions, of
  which Claude Code's own system prompt consumes ~50; practitioner guidance converges on <200
  lines per always-loaded file with degradation observable past ~80 dense
  rule-lines.[^ContextRot][^ClaudeMemoryDocs] `active.md` must carry a **hard line/entry cap** in
  `.knowledge.toml`, with rank-by-(`review_count`, recency) eviction into the on-demand bundle —
  otherwise a healthy store eventually poisons its own projection with volume. Path-scoped
  `.claude/rules/` projection (§4.4) is the pressure-relief valve.

Against over-thinning (disconfirming my own H4 lean): the "just use files, judgment is the binding
constraint" position is now the practitioner consensus for small corpora[^FilesWin][^VectorOverkill]
— which *supports the substrate* and cautions only against the arithmetic. The lifecycle ladder
itself (capture → corroborate → promote → decay) is precisely the "judgment" layer that consensus
says matters. Keep the ladder; simplify its numbers.

---

## 7. Changes required before implementation (lane-2 consolidated list)

1. **Specify candidate retrieval for dedup** (§1.3a): whole-index + description scan with pairwise
   LLM judgment now; named scale ceiling (~500 concepts/store) that triggers the deferred
   SQLite/embedding index.
2. **Append-only expansion; claims immutable after promotion; change = supersede** (§1.3b) —
   structural defense against continuous-rewrite corruption.
3. **Add memory poisoning to §9 as blocking** (§2): provenance trust tiers (external-source
   concepts permanently human-gated from projections), injection scan in the dream pass, projection
   voice de-authorized.
4. **Re-scope against native machinery** (§3): resolve the `/dream` vs Auto Dream two-writer
   conflict explicitly (own `knowledge/` only, or ingest Auto Dream's output); evaluate
   `autoMemoryDirectory`-into-store; drop bespoke work that duplicates native capture.
5. **Replace confidence float with status + review_count + last_seen** (§6); delete threshold
   tuning as a risk item.
6. **Hard cap `active.md`**; prefer path-scoped `.claude/rules/` projection for `type: rule`
   concepts (§4.4, §6).
7. **Scheduled dream pass: sequential subagents only** (headless fan-out hang, §4); threshold-gate
   the nightly run (skip < N candidates, §5).
8. **Transcript parsing gets a version-pinned contract + fallback**, not a one-time Phase-0 check
   (§4.2); per-agent `memory:` rows contingent on issue #57507 resolution (§3).
9. **Projection-health check in SessionStart** (silent-dead external `@`-import, §4.3).
10. **Demote review-by-git-diff from preventive to forensic control** (§1.4): add per-pass change
    caps (max supersessions/deletions per dream pass, halt-and-flag on breach).

## 8. Lane-2 risk grading (likelihood × impact × complexity-to-fix)

| Risk | L | I | Fix cost | Disposition |
|---|---|---|---|---|
| Consolidation rewrite-corruption (§1.3b) | High over months | High (silent knowledge loss) | Low (append-only rule) | Fix |
| Dedup recall shortfall at scale (§1.3a) | Med (scale-dependent) | Med (fragmentation) | Low now / Med later | Fix cheap path now, name ceiling |
| Memory poisoning via ingest/inbox (§2) | Med (single operator, but npm-CVE precedent) | High (persistent compromise) | Med | Fix — blocking |
| Native Auto Dream two-writer conflict (§3) | High if flag lands | Med (churn, lost notes) | Low (scope split) | Fix |
| Headless fan-out hang (§4) | High in cron context | Med (silent no-op nights) | Low (sequential) | Fix |
| Unreviewed bot commits (§1.4) | High | Med | Low (caps) | Fix |
| Projection context-rot (§6) | Med | Med (adherence loss across all rules) | Low (cap) | Fix |
| Confidence-float drift (§6) | Med | Low | Negative (removal simplifies) | Fix by deletion |
| Transcript format churn (§4.2) | Med | Low (feature degrades, recoverable) | Low | Fix |
| OKF v0.1 drift (§4.1) | Low | Low (profile pinned; custom keys legal) | — | Risk-accept (proposal §9.7 stands) |
| Multi-machine store divergence | Low (single operator, one box) | Low | Med (sync protocol) | Risk-accept — YAGNI; git remote is the sync story if ever needed |

---

## Footnotes

[^FaultyMemories]: "Useful Memories Become Faulty When Continuously Updated by LLMs" — Zhang et al., arXiv 2605.12978. Accessed 2026-07-12. https://arxiv.org/pdf/2605.12978
[^AgentsDumber]: "Long-Term Memory Is Making Agents Dumber" — Johnson Lee blog, 2026-05-20 (secondary commentary; ARC-AGI 54% figure reported here, not independently verified). Accessed 2026-07-12. https://johnsonlee.io/2026/05/20/faulty-agent-memory.en/
[^MemorySurvey]: "Memory for Autonomous LLM Agents: Mechanisms, Evaluation, and Emerging Frontiers" — arXiv 2603.07670 (summarization drift, semantic intensification, importance-score drift across model versions, ~29-day empirical half-life). Accessed 2026-07-12. https://arxiv.org/html/2603.07670v1
[^HindsightConsolidation]: "The Consolidation Problem in Agent Memory" — Hindsight (Vectorize) blog, 2026-05-21 (four levers: importance/merge/decay/eviction; "decay is the lever most systems skip"). Accessed 2026-07-12. https://hindsight.vectorize.io/blog/2026/05/21/agent-memory-consolidation
[^MemZero]: "mem0ai/mem0" — GitHub + Mem0 architecture breakdowns (two-phase extract/update; retrieve-similar-then-ADD/UPDATE/DELETE/NOOP). Accessed 2026-07-12. https://github.com/mem0ai/mem0
[^ZepGraphiti]: "Zep: A Temporal Knowledge Graph Architecture for Agent Memory" — arXiv 2501.13956 + getzep/graphiti GitHub (LLM contradiction detection against semantically related edges; invalidate-not-delete with validity windows). Accessed 2026-07-12. https://arxiv.org/html/2501.13956v1
[^LettaSleep]: "Sleep-time Compute" — Letta blog + Letta forum "Sleeptime Agents for Memory Consolidation" (background agents consolidate/dedup/prune while primary agent idle). Accessed 2026-07-12. https://www.letta.com/blog/sleep-time-compute/
[^ParaphraseGap]: "Semantic search as extractive paraphrase span detection" — Language Resources and Evaluation (Springer), + MDPI "Transformer Models for Paraphrase Detection" (semantic beats lexical by 11–20+ points; high-semantic/low-lexical overlap gap). Accessed 2026-07-12. https://link.springer.com/article/10.1007/s10579-023-09715-7
[^LLMJudgeDedup]: "Semantic Needles in Document Haystacks: Sensitivity Testing of LLM-as-a-Judge Similarity Scoring" — arXiv 2604.18835 (threshold-dependent pairwise dedup judgment reliability). Accessed 2026-07-12. https://arxiv.org/pdf/2604.18835
[^BasicMemory]: "basicmachines-co/basic-memory" — GitHub (local-first markdown knowledge graph for LLMs, MCP server). Accessed 2026-07-12. https://github.com/basicmachines-co/basic-memory
[^SqliteMemory]: "sqliteai/sqlite-memory" — GitHub (markdown-based agent memory with semantic search + hybrid retrieval; precedent for the deferred index). Accessed 2026-07-12. https://github.com/sqliteai/sqlite-memory
[^UnreviewedPRs]: "On the Footprints of Reviewer Bots' Feedback on Agentic Pull Requests in OSS GitHub Repositories" — arXiv 2604.24450 (61.38% of agent PRs no recorded review; 71.58% of review comments agent-authored). Accessed 2026-07-12. https://arxiv.org/html/2604.24450v1
[^AIApprovingPRs]: "AI is approving our pull requests" — fin.ai / Intercom engineering blog (rubber-stamping under queue pressure). Accessed 2026-07-12. https://ideas.fin.ai/p/ai-is-approving-our-pull-requests
[^CiscoMemoryCVE]: "Identifying and remediating a persistent memory compromise in Claude Code" — Cisco Blogs (CVE-2026-21852; npm postinstall → MEMORY.md; fix removed user memories from system prompt, v2.1.50). Accessed 2026-07-12. https://blogs.cisco.com/ai/identifying-and-remediating-a-persistent-memory-compromise-in-claude-code
[^OmegamaxCVE]: "CVE-2026-21852: Agent Memory Poisoning in Your Codebase" — Omegamax blog. Accessed 2026-07-12. https://omegamax.co/blog/agent-memory-poisoning-cve-2026
[^MemoryPoisoningStudy]: "From Untrusted Input to Trusted Memory: A Systematic Study of Memory Poisoning Attacks in LLM Agents" — arXiv 2606.04329. Accessed 2026-07-12. https://arxiv.org/pdf/2606.04329
[^ClaudeMemoryDocs]: "How Claude remembers your project" — Claude Code official docs (auto memory v2.1.59+, 200-line/25KB MEMORY.md load, autoMemoryDirectory setting, @-import depth 4 + external-import approval dialog, .claude/rules/ with paths frontmatter, CLAUDE.md as user message not system prompt). Accessed 2026-07-12. https://code.claude.com/docs/en/memory
[^AutoDream]: "Claude Code Dreams: Anthropic's New Memory Feature" — claudefa.st + "Auto Memory and Auto Dream" (antoniocortes.com, 2026-03-30) (four-phase pass; ~24h + >5 sessions trigger; server-side flag rollout — availability unverified as stable API). Accessed 2026-07-12. https://claudefa.st/blog/guide/mechanics/auto-dream
[^DreamSkill]: "grandamenium/dream-skill" — GitHub ("replicates Anthropic's unreleased auto-dream feature," 4-phase, 24h auto-trigger — evidence of community replication and of the feature's flag-gated status). Accessed 2026-07-12. https://github.com/grandamenium/dream-skill
[^SubagentMemory]: "Create custom subagents" — Claude Code docs + shanraisshan/claude-code-best-practice agent-memory report (memory: user|project|local, v2.1.33+, ~/.claude/agent-memory/<name>/, first 200 lines injected). Accessed 2026-07-12. https://code.claude.com/docs/en/sub-agents
[^SubagentMemoryBug]: "[BUG] `memory:` field in subagent frontmatter not functional — v2.1.137; tools allowlist appears to override auto-enable" — anthropics/claude-code issue #57507. Accessed 2026-07-12. https://github.com/anthropics/claude-code/issues/57507
[^OKFSpec]: "knowledge-catalog/okf/SPEC.md" — GoogleCloudPlatform GitHub (v0.1 Draft; type sole required field; index.md/log.md reserved; custom keys permitted). Accessed 2026-07-12. https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md
[^OKFBlog]: "How the Open Knowledge Format can improve data sharing" — Google Cloud blog (announced 2026-06-12). Accessed 2026-07-12. https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing/
[^TranscriptFormat]: "Claude Code JSONL transcript format explained" — claude-dev.tools + simonw/claude-code-transcripts (path/schema confirmed; "internal to Claude Code and changes between versions"). Accessed 2026-07-12. https://claude-dev.tools/docs/jsonl-format
[^HeadlessDocs]: "Run Claude Code programmatically" — Claude Code docs (claude -p waits for background subagents; 10-min cap via CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS). Accessed 2026-07-12. https://code.claude.com/docs/en/headless
[^HeadlessHang]: "claude -p headless under non-TTY parent: parallel Task fan-out hangs" — anthropics/claude-code issue #56540. Accessed 2026-07-12. https://github.com/anthropics/claude-code/issues/56540
[^GenerativeAgents]: "Generative Agents: Interactive Simulacra of Human Behavior" — Park et al., arXiv 2304.03442 (reflection triggered when accumulated importance exceeds threshold ~150). Accessed 2026-07-12. https://arxiv.org/abs/2304.03442
[^RecMem]: "RecMem: Recurrence-based Memory Consolidation for Efficient and Effective Long-Running LLM Agents" — arXiv 2605.16045 (eager consolidation wastes 77–87% construction tokens with no accuracy gain; recurrence-triggered consolidation). Accessed 2026-07-12. https://arxiv.org/html/2605.16045v1
[^BeliefMemory]: "Belief Memory: Agent Memory Under Partial Observability" — arXiv 2605.05583 (ALFWorld 59.88 → 28.71 when probabilistic memory collapsed to deterministic — the confidence-helps evidence, scoped to partial observability). Accessed 2026-07-12. https://arxiv.org/html/2605.05583v1
[^ContextRot]: "Your CLAUDE.md Is Probably Too Long" — tianpan.co, 2026-02-14 (+ MindStudio context-rot analysis) (~150–200 instruction adherence budget, ~50 consumed by system prompt; degradation past ~80 dense lines). Accessed 2026-07-12. https://tianpan.co/blog/2026-02-14-writing-effective-agent-instruction-files
[^FilesWin]: "Forget RAG: The Best AI Agent Memory Is a Plain Text File" — voxos.ai (+ dev.to "All of Them Use Flat Files") (files-win consensus for small corpora; judgment, not infrastructure, is the binding constraint). Accessed 2026-07-12. https://voxos.ai/blog/how-to-give-ai-coding-agents-long-term-m/index.html
[^VectorOverkill]: "Did Agents Kill Vector Search? The Honest, Scale-Dependent Answer" — thedataexperts.us (filesystem agents beat vector pipelines on small complex corpora; advantage inverts at scale). Accessed 2026-07-12. https://www.thedataexperts.us/writing/vector-db-vs-files-agents-retrieval.html
