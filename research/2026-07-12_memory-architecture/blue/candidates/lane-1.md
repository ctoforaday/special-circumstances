# Lane 1 — H1 (substrate) to saturation, then breadth

Assignment: hypothesis H1 (git-native markdown-plus-frontmatter is a proven, sufficient
substrate) verified deep; then breadth across H2–H5. 21 searches/fetches plus local
leaf-node verification on this machine; 7 searches spent on disconfirming evidence
(headless hook failures, flat-file scaling criticism, consolidation-loss studies,
Open Knowledge Format skepticism, bot-commit review fatigue, confidence-calibration
criticism). Saturation reached: final searches returned already-seen sources (Hindsight,
mem0, Letta recurring).

Verdict for this lane: **the architecture is directionally right and better-supported by
external evidence than the proposal itself knows — but it ships with two false inherited
claims, one missing threat model (memory poisoning) severe enough to be blocking, one
factually wrong mapping row (§5 agent memory), and a headless-hooks assumption that
current open bugs contradict.**

---

## 1. H1 — Substrate holds, with corrections

### 1.1 Open Knowledge Format: verified real, verified young

The spec exists as described: OKF v0.1 (explicitly **Draft**), a directory of markdown
files with YAML frontmatter, `type` the only mandatory field; `title`, `description`,
`resource`, `tags`, `timestamp` recommended; producers may add custom fields and
"consumers must tolerate unknown keys" — which makes the §3.1 profile (status, confidence,
provenance, etc.) spec-legal by construction.[^OkfSpec][^OkfBlog]

Corrections and cautions the proposal needs:

- **Reserved files carry no frontmatter.** Per the spec, `index.md` and `log.md` are
  reserved *and have no frontmatter* (versioning via `okf_version` in the root `index.md`
  is the stated exception). The proposal's store layout is compatible, but any tooling
  that assumes frontmatter on `index.md` files would be off-spec.[^OkfSpec]
- **The spec is roughly four weeks old** (announced mid-June 2026) and community
  reception includes exactly the skepticism a pragmatist should price in: "markdown files
  with metadata" rebrand critiques, Google-abandonment risk, brittle path-based links on
  rename, and — notably — an independent observation that *an agent-updated OKF bundle is
  an indirect-prompt-injection vector* (see §4).[^OkfSkeptic][^OkfDeepDive]
- **Abandonment risk is real but cheap.** Because the format degenerates gracefully to
  plain markdown + frontmatter, upstream death costs us the *citation*, not the *store*.
  Recommended posture: adopt OKF as a documentation convention pinned at v0.1
  (`okf_version: "0.1"`), not as a dependency. This matches §9.7 but should be stated as
  the design stance, not a risk item.

### 1.2 Native-surface mapping (§5): five rows verified, one wrong, one shaky

Verified against current Claude Code documentation and, where possible, this machine:

- **`@`-import**: relative and absolute paths including `@~/...` work; imports recurse to
  a **maximum depth of four hops**; code spans/fenced blocks are skipped; imported files
  **load at launch and consume context** (splitting "helps organization but does not
  reduce context"). The first external import triggers an **approval dialog**; if
  declined, imports stay disabled silently — a headless run that never saw the dialog may
  silently not load the projection. Phase 0 must verify approval-state behavior under
  `claude -p`.[^MemoryDocs]
- **`MEMORY.md` auto-memory**: lives at `~/.claude/projects/<project>/memory/MEMORY.md`
  (project path derived from the git repo, shared across worktrees); first 200 lines or
  25KB load at session start; topic files load on demand. Plain markdown, editable — the
  §5 ingest arrow (dream loop reads, promotes, prunes) is mechanically sound.
  Two levers the proposal misses: **`autoMemoryDirectory`** (a settings key that relocates
  the whole auto-memory directory — it could point *into* the knowledge store's
  short-term area, collapsing the ingest hop entirely) and
  `CLAUDE_CODE_DISABLE_AUTO_MEMORY` / `autoMemoryEnabled` for clean-room
  testing.[^MemoryDocs]
- **`.claude/rules/` exists natively and the proposal ignores it.** Markdown files in
  `.claude/rules/` (project) and `~/.claude/rules/` (user) load at launch with CLAUDE.md
  priority, support path-scoped `paths:` frontmatter and symlinks. A generated
  `.claude/rules/knowledge.md` is a *simpler projection target* than
  `@`-import-plus-SessionStart: no import approval dialog, no hop budget, native
  precedence (user rules load before project rules — exactly the proposal's §8 merge
  order, for free).[^MemoryDocs]
- **Agent `memory:` frontmatter — the §5 row is wrong as written.** The harness injects
  persistent memory from **fixed paths**: `~/.claude/agent-memory/<agent>/` (scope
  `user`) or `.claude/agent-memory/<agent>/` (scope `project`), not from arbitrary
  store paths. The proposal's `knowledge/agents/<agent>/` sub-bundle cannot be "what the
  harness injects"; the projection must be *written into* the harness path. And because
  the agent itself writes its own memory there mid-session, the dream loop regenerating
  that file is a **bidirectional write collision** — the loop must merge agent-authored
  notes back into the store before regenerating, or it destroys the agent's own learning.
  Also worth noting: `project`-scoped agent memory is **already git-tracked in-repo** —
  the native surface is partially git-native today.[^SubagentDocs]
- **Hooks**: `SessionStart` (fires on startup/resume/clear/compact, can inject
  `additionalContext`), `Stop` (fires when Claude finishes responding, receives
  `last_assistant_message`), and `PreCompact` (manual/auto matchers) all exist as §5
  assumes — **in interactive mode**.[^HooksDocs]

### 1.3 The shaky row: hooks under `claude -p`

Multiple open issues document hooks misbehaving in non-interactive mode: hooks not
executing at all in headless invocations (#20063), a configured Stop hook causing
`claude -p` to emit an **empty result** (#38651), `PreToolUse` not firing under `-p`
(#40506), and SessionEnd unreliability. Documentation confirms `SessionStart` is
*supported* in `-p` (it can even set `initialUserMessage`), but the bug record says
treat every hook-in-headless behavior as unverified until tested on the shipping
version.[^HeadlessHookBugs][^HooksDocs] Consequence: §7.1's capture path (Stop/PreCompact)
is trustworthy for interactive sessions — which is where trajectories worth capturing
mostly happen — but the scheduled `/dream` flow must not *depend* on hooks firing inside
its own headless run, and Phase 0 needs an explicit hook-fire test matrix
(interactive × headless × each event).

### 1.4 Transcript substrate: verified at the leaf node (resolves §9.1)

Inspected directly on this machine: transcripts are per-session JSONL at
`~/.claude/projects/<project-slug>/<session-uuid>.jsonl`. Line schema (version 2.1.207):
typed records (`user`, `assistant`, `system`, `file-history-snapshot`,
`permission-mode`, ...) with `uuid`/`parentUuid` threading, `sessionId`, `cwd`,
`gitBranch`, `version`, ISO timestamps, and Anthropic-API-shaped `message` objects.
Sidechains (subagent transcripts) are flagged `isSidechain`. Parseable today —
**but the schema is undocumented and carries a `version` field for a reason**: treat it
as an unstable interface, isolate all reads behind one parser module, and record the
tested version.[^LocalTranscripts]

### 1.5 File-based memory precedent: proven pattern, with one consistent asterisk

The "markdown files as agent memory source of truth" pattern is widely shipped:
basic-memory (markdown + local SQLite index, MCP, no server), Wuphf ("LLM-native wiki",
local markdown backed by git with BM25 + SQLite on top), memsearch (markdown + Milvus),
and claude-mem (46k-star Claude Code plugin: hook-based session capture → AI compression
→ local SQLite + full-text search).[^BasicMemory][^AgenticDigest][^ClaudeMem] The
asterisk: **essentially every system that matured added a derived index beside the
files** — SQLite FTS, BM25, or embeddings. Files-as-truth survives; files-as-*retrieval*
is where they all outgrew grep. The proposal's "SQLite + vector index: deferred, not
rejected" (§11) is therefore the right call with the wrong precision — it needs a
**named trigger** (e.g., concept count crossing ~300–500, or first observed dedup miss),
not an indefinite deferral.

The counter-literature (flat files "fail at scale": token bloat, no retrieval, no
supersedence) is real but converges on a nuance that favors the proposal: *"early-stage
agents don't have a retrieval problem — they have a curation problem."* The dream loop
**is** the curation mechanism the critics say flat files lack; and the loudest "markdown
is not memory" piece is from Zep, which sells the alternative.[^MemoryMdProblem][^ZepCritique]

**H1 verdict: holds**, conditional on fixing the agent-memory row, testing hooks
headless, and scheduling the retrieval-index trigger.

---

## 2. H2 — Dedup and consolidation: the discipline is right, the mechanism is underspecified

- **Lossy consolidation is empirically real.** Repeated LLM compression measurably
  destroys knowledge: one study storing 2,000 facts and compressing 36.7× found **60% of
  the knowledge base irretrievably lost**; "summarization drift" (each pass discards
  detail until memory no longer matches what happened) is a named failure mode; a 2026
  paper is titled, on the nose, *Useful Memories Become Faulty When Continuously Updated
  by LLMs* — utility rises early in consolidation, then declines.[^ConsolidationProblem][^FaultyMemories]
  The proposal's existing mitigations are the exact ones the literature recommends —
  keep raw evidence linked from consolidated concepts (`provenance`), soft-delete via
  `supersedes` + one-cycle grace, git history as ground truth. Add two hardenings:
  **never rewrite a concept body destructively during merge** (append evidence; edit
  claims only with the diff shown), and **cap fan-in per consolidation pass** so one bad
  dream can't restructure the whole bundle.
- **The mature dedup pipeline has a stage the proposal lacks.** Mem0's operation — the
  closest production analogue of expand-before-append — is: embed candidate fact,
  vector-retrieve top-K similar existing memories, then LLM classifies
  ADD/UPDATE/DELETE/NOOP against that neighborhood.[^MemZero] The proposal specifies the
  classify stage but not candidate retrieval. At current scale this is fine — a few
  hundred small concept files fit in the consolidator's context, so "read the whole
  bundle" is the retrieval — but that *must be stated as the v1 mechanism with its scale
  ceiling*, connecting to the index trigger in §1.5.
- **Review-by-git-diff is a weak sole guard.** Bot-generated commits are systematically
  under-reviewed: Dependabot merges ~54% with heavy noise, and the documented failure
  pair is rubber-stamping or queue abandonment.[^BotReviewFatigue] A single operator
  reviewing nightly dream diffs will decay to LGTM within weeks. Mitigations: bound the
  per-dream diff (candidate cap), one-line dream commit summaries
  (`+3 concepts, 2 merged, 1 pruned` — already in §7.5), a weekly digest instead of a
  daily review expectation, and reserve mandatory human review for the tiers where it
  changes behavior (projection/active.md changes, cross-scope promotion, rule-skill
  promotion) rather than every merge.

---

## 3. H3 — Cadence: the hybrid design is right; the risk is the scheduler, not the clock

- Idle-time consolidation is now a mainstream pattern with a name: Letta's **sleep-time
  compute** — background agents that rewrite/derive memory while the primary agent is
  idle; one implementation even commits reflections to an isolated git branch to avoid
  contention.[^LettaSleep] The Stanford generative-agents architecture triggers
  reflection on an **importance-sum threshold** (~2–3×/day in practice), i.e.
  event-thresholded, not clock-driven.[^GenerativeAgents]
- The proposal is already a hybrid: Stop/PreCompact hooks give event-triggered *capture*;
  nightly `/dream` is a *sweep*, not the sole trigger. That matches the literature.
  Two adjustments: (a) add an **event-threshold fallback** — when pending short-term
  candidates exceed N, surface a "run /dream" nudge (SessionStart already planned to
  carry exactly this line), so consolidation still happens if the scheduler never runs;
  (b) treat headless reliability as the real risk — established guidance is to run a
  workflow interactively until it is boringly predictable *before* scheduling it
  headless, which the phased plan (interactive `/dream` in Phase 2, schedule in Phase 5)
  already respects.[^HeadlessGuide]

---

## 4. NEW BLOCKING RISK — memory poisoning (absent from §9 entirely)

This architecture builds a pipeline from **untrusted input to always-on trusted
context**: web pages read mid-session and `/ingest`ed documents flow into trajectories →
short-term notes → consolidated concepts → `active.md` → *every future session's
context*. That pipeline is a documented attack class:

- **CVE-2026-21852** (disclosed April 2026, patched in Claude Code v2.2): a malicious
  npm package appended instructions to `MEMORY.md` during install; Claude Code then
  treated them as authoritative in every session.[^MemoryPoisonCve] The proposal's store
  reproduces this surface and *widens* it (more files, more writers).
- **SpAIware** demonstrated persistent spyware planted in ChatGPT long-term memory via
  indirect prompt injection in web content — attack and effect temporally
  decoupled.[^MemoryPoisonSurvey]
- Systematic studies report attack success rates against LLM agent memory systems of
  80–99%.[^MemoryPoisonSurvey]
- The dream loop adds a *laundering* mechanism the attacks love: two poisoned
  trajectories = `review_count: 2` = "corroborated" = auto-promoted to `active`. The
  consolidator can convert a one-shot injection into a high-confidence permanent rule.
- Independent corroboration that this is the format's known weak point: OKF community
  discussion flags agent-updated bundles as an indirect-prompt-injection
  vector.[^OkfDeepDive]

Required changes (blocking before Phase 1):

1. **Trust tiers in provenance**: `operator-confirmed` > `trajectory-derived` >
   `external-ingest`. Tier caps the maximum status reachable without a human gate.
2. **External-ingest content never auto-promotes to `active`** — `/ingest` output is
   quarantined at `candidate` until a human confirms.
3. **Injection screening at capture and at promotion** (instruction-shaped content in a
   `fact`, imperative verbs addressed to the agent, tool-use directives inside ingested
   text → flag, don't consolidate).
4. Corroboration must come from **independent provenance** — two notes tracing to the
   same source (same URL, same package, same session) count once.

---

## 5. H4 — Complexity: mostly earns its keep; two simplifications and two false premises

**Evidence that the lifecycle machinery is not gold-plating:**

- **Context rot is measured**: Chroma's 18-model study shows output quality degrades as
  input grows, and *irrelevant* context degrades it fast — even single distractors hurt;
  quality of context beats quantity.[^ContextRot] An unbounded CLAUDE.md/MEMORY.md pile
  is therefore not merely untidy — it is a measured performance regression. The bounded,
  curated `active.md` projection is evidence-backed context engineering. (This verifies,
  rather than inherits, the reviewer's "unbounded pile" claim — with the addition that
  **`active.md` needs an explicit token budget**, which the proposal lacks.)
- **Decay is the under-provisioned lever, not the over-provisioned one**: practitioner
  literature calls decay "the lever most agent memory systems skip, and the one that
  matters most for long-running agents"; half-life decay reinforced by fresh evidence is
  a standard form.[^MemoryEviction]

**Simplifications a pragmatist should take:**

- **Drop the stored `confidence` float in v1.** `review_count`, `last_seen`, `status`,
  and provenance tier are *observable facts*; a stored 0.0–1.0 confidence is a synthetic
  number with admitted-guess thresholds (§9.6) and a known calibration failure mode
  ("runaway certainty").[^MemoryEviction] Derive activation from observables
  (`status: active` AND `review_count ≥ 2` AND `last_seen` within window AND trust tier
  sufficient); keep the schema slot for later. This deletes the worst-tuned arithmetic
  without losing the ladder.
- **Collapse the double injection.** §5 has `active.md` arriving via both `@`-import and
  a SessionStart hook ("belt-and-suspenders") — that is double context cost and two
  failure surfaces where the docs offer a third, simpler native path: a generated file
  under `.claude/rules/` (§1.2). Pick exactly one projection channel.
- Single-operator YAGNI confirmed as *partial*: the project-store PR-ratification flow
  (§7.5 step 5, "optionally open a PR") is latent value for a one-person suite — keep
  optional, off by default. But the global/project split itself costs little and maps to
  native precedence, so it stays.

**Two false premises found by local verification (critical-stance):**

1. §9.5 / Phase 5 cite "the port plan's **existing** secret-scrub (`git grep`
   denylist)". **No such gate exists.** The suite contains a semantic guardrail rule
   whose deterministic hook is future-tense ("a PreToolUse hook *will* enforce…"), a git
   history-scrub cheatsheet, and a "PII-scrubbed" convention note — no scrub tooling.
   The gate must be *built*, and ad-hoc `git grep` is the wrong tool: use a maintained
   scanner (gitleaks / detect-secrets class) plus capture-time redaction — claude-mem's
   `<private>` tag exclusion is a pattern worth stealing.[^LocalRepoScrub][^ClaudeMem]
2. §7.6 claims "`docs/scheduling.md` in sleeper-service **already documents** the
   recipes". **The file does not exist**; sleeper-service is currently a stub
   (plugin.json + README).[^LocalRepoSleeper] The scheduling story is planned, not
   shipped — `/dream` inherits a dependency on unbuilt work, which belongs in the
   phase plan, not the assumptions.

(Also unverified in this lane, labeled as such: the internal FUSE prior art, the
OpenClaw dream-diary degradation anecdote, and the AgentOrange `continuous_learning`
aspect's "battle-tested" status — all internal artifacts cited by the proposal without
independent corroboration here.)

---

## 6. H5 — Alternatives: nothing dominates; three things to steal

- **claude-mem** (46k stars) is the strongest adopt-instead candidate: plugin-native,
  hook-driven session capture, AI compression, local storage, layered retrieval
  (~10× token efficiency claimed). It fails the suite's constraints where they bind:
  storage is SQLite (not human-readable markdown, not git-diffable, not PR-reviewable),
  no project-store-committed-with-code, no promotion ladder to skills, third-party
  dependency for load-bearing infrastructure. **Steal**: `<private>` capture-time
  redaction; proof that hook-based trajectory capture works at ecosystem
  scale.[^ClaudeMem]
- **basic-memory** is the closest philosophical match (markdown source of truth +
  derived SQLite index + MCP, no server) and is the existence proof for §1.5's
  files-plus-index endgame; it lacks lifecycle/decay/promotion and agent-memory
  integration, so it complements rather than replaces.[^BasicMemory]
- **mem0 / Letta / Zep**: all service/DB-bound — the frontier's predicted disqualifier
  holds. **Steal**: mem0's retrieve-then-classify dedup pipeline (§2), Letta's sleep-time
  framing and isolated-branch commits (§3), Zep's fact-supersedence-with-validity-interval
  as the conceptual model behind `supersedes`/`last_seen`.[^MemZero][^LettaSleep][^ZepCritique]
- **Native-surfaces-plus-thin-skill** (H4's thin design) is not a competitor once the
  context-rot and curation evidence is in (§5) — but its best half-idea survives as the
  `.claude/rules/` projection channel and the `autoMemoryDirectory` ingest collapse.

---

## 7. Changes required before implementation (lane-1 consolidated list)

| # | Change | Grade |
|---|---|---|
| 1 | Add memory-poisoning threat model: provenance trust tiers, independent-source corroboration, ingest quarantine, injection screening at capture/promotion | **Blocking** |
| 2 | Fix §5 agent-memory row: project into `.claude/agent-memory/<name>/`; define merge for bidirectional writes | Blocking (correctness) |
| 3 | Build the secret-scrub gate (gitleaks/detect-secrets + capture-time `<private>`-style redaction) — it does not exist to be reused | Blocking for any remote push |
| 4 | Phase 0 adds a hook-fire test matrix (interactive × headless × Stop/PreCompact/SessionStart) and an import-approval-under-`-p` check; isolate transcript parsing behind one versioned module | High |
| 5 | Specify v1 dedup retrieval as whole-bundle-in-context with a named ceiling; define the derived-index trigger (~300–500 concepts or first observed dedup miss) | High |
| 6 | Token budget for `active.md`; single projection channel (prefer generated `.claude/rules/` file over `@`-import + SessionStart double injection) | High |
| 7 | Drop stored `confidence` float in v1; derive activation from `review_count`/`last_seen`/`status`/trust tier | Medium (simplification) |
| 8 | Event-threshold consolidation fallback (pending-candidate count nudge) so the system degrades gracefully without the scheduler | Medium |
| 9 | Consolidation hardening: no destructive body rewrites, fan-in cap per dream pass; weekly digest + tier-gated human review instead of nightly diff review | Medium |
| 10 | Reframe OKF as pinned convention (`okf_version: "0.1"`), correct the §7.6 and §9.5 "already exists" claims, note `index.md`/`log.md` carry no frontmatter | Low |

---

## Footnotes

[^OkfSpec]: *Open Knowledge Format (OKF) Specification*, GoogleCloudPlatform/knowledge-catalog `okf/SPEC.md` (GitHub), accessed 2026-07-12. v0.1 Draft; `type` sole required field; `index.md`/`log.md` reserved without frontmatter; `okf_version` in root index; consumers must tolerate unknown keys.
[^OkfBlog]: *How the Open Knowledge Format can improve data sharing*, Google Cloud Blog, accessed 2026-07-12. "Just markdown, just files, just YAML frontmatter"; hostable in any git repo.
[^OkfSkeptic]: *Google Cloud Introduces Open Knowledge Format (OKF)*, MarkTechPost, June 16 2026, and community adoption commentary (owox.com; dev.to/maskaravivek), accessed 2026-07-12. Announced mid-June 2026; "markdown with metadata" rebrand critique; abandonment risk.
[^OkfDeepDive]: *Is OKF Worth Adopting Yet? A Deep Dive into Google's Open Knowledge Format*, ewandel.de, accessed 2026-07-12. v0.1 breaking-change risk; link brittleness on rename; agent-updated bundles as indirect-prompt-injection vector.
[^MemoryDocs]: *How Claude remembers your project*, Claude Code documentation (code.claude.com/docs/en/memory), accessed 2026-07-12. `@`-import semantics (4-hop max, code-span skip, external-import approval dialog, imports load at launch), MEMORY.md location and 200-line/25KB load, `autoMemoryDirectory`, `.claude/rules/` incl. user-level rules and load order.
[^SubagentDocs]: *Create custom subagents*, Claude Code documentation (code.claude.com/docs/en/sub-agents), accessed 2026-07-12. `memory: user|project|local`; user scope at `~/.claude/agent-memory/<name>/`, project scope at `.claude/agent-memory/<name>/` ("shareable via version control").
[^HooksDocs]: *Hooks reference*, Claude Code documentation (code.claude.com/docs/en/hooks), accessed 2026-07-12. SessionStart sources and `additionalContext`/`initialUserMessage` (the latter explicitly applies in `-p`); Stop `last_assistant_message`; PreCompact matchers.
[^HeadlessHookBugs]: GitHub issues anthropics/claude-code #20063 (hooks don't run in headless mode), #38651 (Stop hook empties `claude -p` result), #40506 (PreToolUse not firing in `-p`), #37559 (hook docs vs. behavior), accessed 2026-07-12. Open bug record for hooks under non-interactive mode.
[^LocalTranscripts]: Local inspection, `~/.claude/projects/C--Users-gbloc-Projects-AgentOrange/*.jsonl`, Claude Code v2.1.207, this machine, 2026-07-12. Primary-source verification of transcript path and line schema (§1.4).
[^BasicMemory]: *basic-memory* (basicmachines-co, GitHub), accessed 2026-07-12. Markdown knowledge graph + local SQLite index, MCP server, no cloud.
[^AgenticDigest]: *Git-based LLM wikis move agent memory into Markdown*, The Agentic Digest, accessed 2026-07-12. Wuphf: local markdown + git + BM25/SQLite index; survey of filesystem-markdown memory family and its cost ("at the cost of scale and automatic semantic search").
[^ClaudeMem]: *claude-mem* (thedotmack, GitHub; docs.claude-mem.ai; Augment Code coverage), accessed 2026-07-12. 46k-star Claude Code plugin: hook-based capture, AI compression, local SQLite + FTS, layered retrieval, `<private>` tag exclusion.
[^MemoryMdProblem]: *The MEMORY.md Problem: Why Local Files Fail at Scale*, DEV Community (anajuliabit), and *memweave* (Towards Data Science), accessed 2026-07-12. Flat-file failure modes (token bloat, no retrieval/supersedence); counter-nuance: "early-stage agents don't have a retrieval problem — they have a curation problem."
[^ZepCritique]: *Markdown is not agent memory*, Zep blog, accessed 2026-07-12. Compounding errors, no fact supersedence, concurrent-writer divergence; vendor of the competing temporal-knowledge-graph product — motivated but substantively argued.
[^ConsolidationProblem]: *The Consolidation Problem in Agent Memory*, Hindsight (vectorize.io), May 2026, accessed 2026-07-12. Consolidation vs. lossy compaction; 2,000-fact/36.7× compression study with 60% irretrievable loss; summarization drift; keep-raw-linked mitigation.
[^FaultyMemories]: *Useful Memories Become Faulty When Continuously Updated by LLMs*, arXiv:2605.12978, accessed 2026-07-12. Memory utility rises then declines under continuous LLM updating.
[^MemZero]: *Mem0: Building Production-Ready AI Agents with Scalable Long-Term Memory* (paper coverage: emergentmind.com; deepwiki mem0ai/mem0), accessed 2026-07-12. Pipeline: embed → vector-retrieve top-K neighbors → LLM classifies ADD/UPDATE/DELETE/NOOP.
[^BotReviewFatigue]: *Reducing Alert Fatigue via AI-Assisted Negotiation: A Case for Dependabot* (arXiv:2502.06175); IEEE TSE study of dependency-bot PRs (arXiv:2206.07230); Pixee merge-rate analysis, accessed 2026-07-12. ~54% Dependabot merge rate; rubber-stamping vs. queue abandonment as the documented failure pair.
[^LettaSleep]: *Sleep-time Compute*, Letta blog + Letta docs (sleeptime architectures) + community best-practices forum, accessed 2026-07-12. Background agents consolidating memory during idle; isolated git-branch commits to avoid contention.
[^GenerativeAgents]: Park et al., *Generative Agents: Interactive Simulacra of Human Behavior* (2023), via architecture summaries (memx.app; subodhjena.com), accessed 2026-07-12. Reflection triggered by importance-sum threshold (~2–3×/day); retrieval = recency (exponential decay) + importance + relevance.
[^HeadlessGuide]: *Claude Code in CI/CD and Headless Automation* (hidekazu-konishi.com) and MindStudio headless-mode guides, accessed 2026-07-12. Headless as the last pattern adopted; run interactively until predictable.
[^MemoryPoisonCve]: *CVE-2026-21852: Agent Memory Poisoning in Your Codebase* (omegamax.co; Cisco disclosure, April 2026), accessed 2026-07-12. MEMORY.md poisoning via npm package; patched Claude Code v2.2.
[^MemoryPoisonSurvey]: *From Untrusted Input to Trusted Memory: A Systematic Study of Memory Poisoning Attacks in LLM Agents* (arXiv:2606.04329); Christian Schneider, *Memory poisoning in AI agents: exploits that wait*; SpAIware coverage, accessed 2026-07-12. 80–99% reported attack success rates; temporal decoupling of attack and effect.
[^MemoryEviction]: *Agent Memory Eviction: 8 Policies That Stop Stale Tool Decisions* (Medium, Bhagya Rana) and *Governing Evolving Memory in LLM Agents (SSGM)* (arXiv:2603.11768), accessed 2026-07-12. Half-life decay reinforced by evidence; inferred memories decay faster; decay as the most-skipped, most-needed lever; confidence calibration / runaway-certainty risk.
[^ContextRot]: *Context Rot: How Increasing Input Tokens Impacts LLM Performance*, Chroma Research, July 2025, accessed 2026-07-12. 18 frontier models degrade with input length; irrelevant distractors degrade sharply; vendor caveat (Chroma sells vector DBs) noted.
[^LocalRepoScrub]: Local verification, special-circumstances repo, 2026-07-12: `grep -i secret|scrub|denylist` across `*.md` — no secret-scrub gate artifact; `plugins/prosthetic-conscience/skills/agent-guardrails/SKILL.md` says a deterministic PreToolUse hook "will enforce" (future tense).
[^LocalRepoSleeper]: Local verification, special-circumstances repo, 2026-07-12: `plugins/sleeper-service/` contains only `.claude-plugin/plugin.json` and `README.md`; no `docs/scheduling.md`.
