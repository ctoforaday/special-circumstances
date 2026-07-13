# Blue report — memory architecture for Special Circumstances (living, Round 0)

**Scope:** evaluate the OKF-inspired, git-native memory proposal
(`inputs/memory-architecture-proposal.md`): global + per-project stores, native surfaces as
generated projections, trajectory-to-memory extraction, nightly dream consolidation.
**Method:** two research lanes (H1-deep and H2-deep, breadth across H1–H5), 41 searches/fetches
plus leaf-node verification on this machine; disconfirming budget met in both lanes; both lanes
reached saturation (Hindsight, mem0, Letta, basic-memory recurring). This report is the union of
both lane drafts; nothing substantive dropped.

## Verdict

**The architecture is directionally right and better-supported by external evidence than the
proposal itself knows** — Anthropic is independently converging on the same loop natively, the
consolidation literature endorses exactly the proposal's levers, and measured context-rot data
verifies (rather than inherits) the reviewer's "unbounded pile" complaint. But it ships with:

- **two false inherited claims** (a secret-scrub gate and a `docs/scheduling.md` cited as
  "existing" — neither exists in the repo);
- **one missing threat model severe enough to be blocking** (memory poisoning: the store is a
  pipeline from untrusted input to always-on trusted context, with a documented CVE precedent
  against the exact file the proposal adopts as its inbox);
- **one factually wrong mapping row** (§5 agent `memory:` — the harness injects from fixed
  paths, not from arbitrary store paths, and the write is bidirectional);
- **a headless-hooks assumption that current open bugs contradict**; and
- **an unpriced collision with native machinery** (Auto Memory shipped; a native "Auto Dream"
  consolidation is rolling out behind a flag and would be a second writer on `MEMORY.md`).

No surveyed alternative dominates; the bespoke layer remains justified for a *shrunken* remit.
Consolidated required changes are in §8; risk grading in §9.

---

## 1. H1 — Substrate: holds, with corrections

### 1.1 Open Knowledge Format: verified real, verified young

The spec exists as described: OKF v0.1 (explicitly **Draft**), in
`GoogleCloudPlatform/knowledge-catalog`; a directory of markdown files with YAML frontmatter,
`type` the only mandatory field; `title`, `description`, `resource`, `tags`, `timestamp`
recommended; producers may add custom fields and "consumers must tolerate unknown keys" — which
makes the proposal's §3.1 profile (status, confidence, provenance, etc.) **spec-legal by
construction**, a profile rather than a fork.[^OkfSpec][^OkfBlog]

Corrections and cautions:

- **Reserved files carry no frontmatter.** Per the spec, `index.md` and `log.md` are reserved
  *and have no frontmatter* (versioning via `okf_version` in the root `index.md` is the stated
  exception). The proposal's store layout is compatible, but any tooling that assumes frontmatter
  on `index.md` files would be off-spec.[^OkfSpec]
- **The spec is roughly four weeks old** (announced mid-June 2026). Community reception includes
  exactly the skepticism a pragmatist should price in: "markdown files with metadata" rebrand
  critiques, Google-abandonment risk, brittle path-based links on rename, and — notably — an
  independent observation that *an agent-updated OKF bundle is an indirect-prompt-injection
  vector* (see §4).[^OkfSkeptic][^OkfDeepDive] Its "external documented standard" benefit is
  true but currently aspirational — the ecosystem is four weeks old.
- **Abandonment risk is real but cheap.** The format degenerates gracefully to plain
  markdown + frontmatter, so upstream death costs the *citation*, not the *store*. Recommended
  posture: adopt OKF as a documentation convention pinned at v0.1 (`okf_version: "0.1"`), not as
  a dependency — this matches §9.7 of the proposal but should be stated as the design stance,
  not a risk item.

### 1.2 Native-surface mapping (proposal §5): five rows verified, one wrong, one shaky

Verified against current Claude Code documentation and, where possible, this machine:

- **`@`-import**: relative and absolute paths including `@~/...` work; imports recurse to a
  **maximum depth of four hops**; code spans/fenced blocks are skipped; imported files **load at
  launch and consume context** (splitting "helps organization but does not reduce context").
  The first import pointing outside the project triggers a **one-time approval dialog**; if
  declined, imports stay disabled **silently** — the global
  `@~/.claude/knowledge/projections/active.md` import can be dead with no error surface, and a
  headless run that never saw the dialog may silently not load the projection. Phase 0 must
  verify approval-state behavior under `claude -p`; projection health needs a SessionStart
  check.[^MemoryDocs]
- **`MEMORY.md` auto-memory** (native, on by default since v2.1.59): lives at
  `~/.claude/projects/<project>/memory/MEMORY.md` (project path derived from the git repo,
  shared across worktrees); first 200 lines or 25KB load at session start; topic files load on
  demand. Plain markdown, editable — the §5 ingest arrow (dream loop reads, promotes, prunes) is
  mechanically sound. Two levers the proposal misses: **`autoMemoryDirectory`** (a settings key
  that relocates the whole auto-memory directory — it could point *into* the knowledge store's
  short-term area, collapsing the ingest hop entirely) and
  `CLAUDE_CODE_DISABLE_AUTO_MEMORY` / `autoMemoryEnabled` for clean-room testing.[^MemoryDocs]
- **`.claude/rules/` exists natively and the proposal ignores it.** Markdown files in
  `.claude/rules/` (project) and `~/.claude/rules/` (user) load at launch with CLAUDE.md
  priority, support **path-scoped `paths:` frontmatter** (file-type-specific knowledge loads
  only when Claude touches matching files) and symlinks. A generated
  `.claude/rules/knowledge.md` is a *simpler projection target* than
  `@`-import-plus-SessionStart: no import approval dialog, no hop budget, and native precedence
  (user rules load before project rules — exactly the proposal's §8 merge order, for
  free). Projecting `type: rule` concepts to `.claude/rules/knowledge-*.md` with `paths` derived
  from concept tags spends context only when relevant and keeps `CLAUDE.md` untouched by
  generated content.[^MemoryDocs]
- **Agent `memory:` frontmatter — the §5 row is wrong as written.** The harness injects
  persistent memory from **fixed paths**: `~/.claude/agent-memory/<agent>/` (scope `user`) or
  `.claude/agent-memory/<agent>/` (scope `project`), not from arbitrary store paths; the first
  200 lines of the agent's MEMORY.md are injected. The proposal's `knowledge/agents/<agent>/`
  sub-bundle cannot be "what the harness injects"; the projection must be *written into* the
  harness path. And because the agent itself writes its own memory there mid-session, the dream
  loop regenerating that file is a **bidirectional write collision** — the loop must merge
  agent-authored notes back into the store before regenerating, or it destroys the agent's own
  learning. Two further notes: `project`-scoped agent memory is **already git-tracked in-repo**
  (the native surface is partially git-native today), and there is an open bug where the
  `memory:` field is **non-functional when a tools allowlist is present** (issue #57507) — the
  row is load-bearing on a currently-flaky feature.[^SubagentDocs][^SubagentMemoryBug]
- **Hooks**: `SessionStart` (fires on startup/resume/clear/compact, can inject
  `additionalContext`), `Stop` (fires when Claude finishes responding, receives
  `last_assistant_message`), and `PreCompact` (manual/auto matchers) all exist as §5 assumes —
  **in interactive mode**.[^HooksDocs]

### 1.3 The shaky row: hooks and fan-out under `claude -p`

Multiple open issues document hooks misbehaving in non-interactive mode: hooks not executing at
all in headless invocations (#20063), a configured Stop hook causing `claude -p` to emit an
**empty result** (#38651), `PreToolUse` not firing under `-p` (#40506), and SessionEnd
unreliability. Documentation confirms `SessionStart` is *supported* in `-p` (it can even set
`initialUserMessage`), but the bug record says treat every hook-in-headless behavior as
unverified until tested on the shipping version.[^HeadlessHookBugs][^HooksDocs]

Separately: `claude -p` waits for background subagents (10-minute default cap, tunable via
`CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS`), but there is an open issue where **parallel Task
fan-out hangs under non-TTY parents** (cron/scheduled contexts) — precisely the dream loop's
runtime.[^HeadlessDocs][^HeadlessHang]

Consequences: §7.1's capture path (Stop/PreCompact) is trustworthy for interactive sessions —
which is where trajectories worth capturing mostly happen — but the scheduled `/dream` flow must
not *depend* on hooks firing inside its own headless run; the scheduled pass should be designed
**sequential-subagent or single-agent**, with parallel fan-out reserved for interactively-invoked
`/dream` and `/memory-bootstrap`; and Phase 0 needs an explicit hook-fire test matrix
(interactive × headless × Stop/PreCompact/SessionStart).

### 1.4 Transcript substrate: verified at the leaf node (resolves proposal §9.1)

Inspected directly on this machine: transcripts are per-session JSONL at
`~/.claude/projects/<project-slug>/<session-uuid>.jsonl`. Line schema (version 2.1.207): typed
records (`user`, `assistant`, `system`, `file-history-snapshot`, `permission-mode`, ...) with
`uuid`/`parentUuid` threading, `sessionId`, `cwd`, `gitBranch`, `version`, ISO timestamps, and
Anthropic-API-shaped `message` objects. Sidechains (subagent transcripts) are flagged
`isSidechain`. Parseable today — **but the schema is undocumented, internal, and changes between
releases** (it carries a `version` field for a reason).[^LocalTranscripts][^TranscriptFormat]
Treatment: isolate all reads behind one parser module; make §9.1's Phase-0 check a
**pinned-version contract with a fallback** (e.g. degrade to `/export`), not a one-time
confirmation; record the tested version.

### 1.5 File-based memory precedent: proven pattern, with one consistent asterisk

The "markdown files as agent memory source of truth" pattern is widely shipped: basic-memory
(markdown + local SQLite index, MCP, no server), Wuphf ("LLM-native wiki", local markdown backed
by git with BM25 + SQLite on top), memsearch (markdown + Milvus), sqlite-memory (markdown +
hybrid retrieval), and claude-mem (46k-star Claude Code plugin: hook-based session capture → AI
compression → local SQLite + full-text
search).[^BasicMemory][^AgenticDigest][^ClaudeMem][^SqliteMemory] The asterisk: **essentially
every system that matured added a derived index beside the files** — SQLite FTS, BM25, or
embeddings. Files-as-truth survives; files-as-*retrieval* is where they all outgrew grep. The
proposal's "SQLite + vector index: deferred, not rejected" (§11) is therefore the right call with
the wrong precision — it needs a **named trigger** (concept count crossing ~300–500, or first
observed dedup miss), not an indefinite deferral.

The counter-literature (flat files "fail at scale": token bloat, no retrieval, no supersedence)
is real but converges on a nuance that favors the proposal: *"early-stage agents don't have a
retrieval problem — they have a curation problem."* The dream loop **is** the curation mechanism
the critics say flat files lack. The loudest "markdown is not memory" piece is from Zep, which
sells the alternative (compounding errors, no fact supersedence, concurrent-writer divergence —
motivated but substantively argued).[^MemoryMdProblem][^ZepCritique] The
"just use files, judgment is the binding constraint" position is now practitioner consensus for
small corpora, and filesystem agents beat vector pipelines on small complex corpora (the
advantage inverts at scale) — which *supports the substrate* and cautions only against the
arithmetic (§6).[^FilesWin][^VectorOverkill]

**H1 verdict: holds**, conditional on fixing the agent-memory row, testing hooks headless,
choosing one projection channel, and naming the retrieval-index trigger.

---

## 2. H2 — Consolidation and dedup: top technical risk; real, measured, mitigable; under-specified in two precise places

### 2.1 The failure is documented, not hypothetical

- **Repeated LLM compression measurably destroys knowledge.** One study storing 2,000 facts and
  compressing 36.7× found **60% of the knowledge base irretrievably lost**; "summarization
  drift" (each pass discards detail until memory no longer matches what happened) is a named
  failure mode.[^ConsolidationProblem]
- **Continuous LLM rewriting of stored memories degrades them.** *Useful Memories Become Faulty
  When Continuously Updated by LLMs* documents that repeated update cycles corrupt previously
  useful memories via interference, drift of the stored text's meaning, and loss of specifics;
  degradation intensifies with update frequency — utility rises early in consolidation, then
  declines.[^FaultyMemories] Related secondary commentary reports memory utility falling **below
  the no-memory baseline** under repeated consolidation (one reported figure: a frontier model
  failing 54% of ARC-AGI problems it had previously solved once consolidated memory was
  attached — reported in commentary, not independently re-verified).[^AgentsDumber]
- **The specific corruption mode is semantic intensification and summarization drift** — "likes
  mild spicy food" becomes "loves very spicy food" over rewrite cycles; each compression
  discards entity-level detail retrieval later depends on.[^MemorySurvey]
- **The OpenClaw "details unavailable" pattern generalizes**: stale, contradictory, and
  near-duplicate facts accumulate and degrade behavior *even when retrieval works*; drivers are
  context economics, entity drift, and index precision decay.[^ConsolidationProblem]

### 2.2 The industry's response validates the proposal's levers

The consolidation problem's standard treatment: **importance filtering at write time, merge with
entity/conflict resolution at write time, confidence decay over time (exponential preferred),
and eviction-by-unretrievability rather than deletion**.[^ConsolidationProblem] Production
systems implement exactly this shape: mem0's pipeline embeds each candidate fact,
vector-retrieves the top-K similar existing memories, then has an LLM classify
ADD/UPDATE/DELETE/NOOP against that neighborhood[^MemZero]; Zep/Graphiti compares new edges
against semantically related existing edges with an LLM to detect contradictions, then
*invalidates* (never deletes) superseded facts with validity windows[^ZepGraphiti]; Letta ships
sleep-time agents that consolidate, deduplicate, and prune memory blocks in the background while
the primary agent is idle.[^LettaSleep] The proposal's promotion ladder, supersedes-not-delete,
and decay table are the same levers; its existing mitigations (provenance links to raw evidence,
`supersedes` + one-cycle grace, git history as ground truth) are the exact ones the literature
recommends. **H2 does not kill the design.**

### 2.3 The two precise under-specifications

**(a) Candidate retrieval is unspecified.** Expand-existing-before-append requires the
consolidator to *find* the overlapping concept first; the proposal says "search the target
bundle" (§6) without saying how — mem0's pipeline has a retrieval stage the proposal lacks.
Evidence that this matters: lexical/title matching is systematically weak against paraphrase —
semantic methods beat lexical baselines by 11–20+ points on paraphrase detection, and
near-identical meanings routinely share almost no surface vocabulary (99%+ semantic similarity
with single-digit BLEU overlap).[^ParaphraseGap] LLM pairwise judgment of *given* candidate
pairs is reliable at high similarity but degrades sharply near the decision boundary (at cosine
≥0.95 every flagged pair is a true duplicate; at 0.85–0.87 only ~1.5% are)[^LLMJudgeDedup] — so
the binding constraint is **recall of candidate pairs**, not judgment quality. At the suite's
realistic scale (hundreds of small concept files, single operator), the whole bundle — or at
least `index.md` plus every `description` line — fits in the consolidator's context, so
"read the whole bundle, then pairwise-judge" is adequate **now**. But the design must *name*
this as the v1 candidate-retrieval mechanism, state its scale ceiling (~300–500 concepts per
store, or first observed dedup miss), and make that ceiling the trigger for the deferred
SQLite/vector index (§11). Precedent for the upgrade path exists: markdown-store-plus-SQLite
hybrids are an established pattern.[^BasicMemory][^SqliteMemory]

**(b) "Expand existing" invites the continuous-rewrite failure.** Expansion implemented as an
LLM re-emitting the whole concept file is exactly the repeated-rewrite loop that corrupts
memories.[^FaultyMemories] The fix is cheap and structural: **expansion appends — it never
rewrites the claim.** A concept body's core claim becomes effectively immutable after promotion;
corroboration appends to the Evidence section and bumps counters in frontmatter; changing the
*claim itself* requires `supersedes` (new file, old one deprecated) — mirroring Zep's
invalidate-don't-mutate discipline.[^ZepGraphiti] This turns every consolidation diff into
additions + frontmatter bumps + whole-file supersessions — trivially reviewable, and drift-proof
because prose is never LLM-round-tripped in place. Add two further hardenings: **cap fan-in per
consolidation pass** so one bad dream can't restructure the whole bundle, and never edit claims
without the diff shown.

### 2.4 Review-by-git-diff is a weak sole guard — demote it from preventive to forensic

The proposal leans on "every merge is a git diff a human can review" (§9.4). Measured behavior
says otherwise: bot-generated commits are systematically under-reviewed — Dependabot PRs merge
~54% amid heavy noise, with rubber-stamping or queue abandonment as the documented failure
pair[^BotReviewFatigue]; in a large OSS sample, **61.4% of agent-authored pull requests received
no recorded review activity at all**, and 71.6% of review comments on them were authored by
other agents.[^UnreviewedPRs][^AIApprovingPRs] A nightly bot commit to a knowledge repo is the
least-reviewed artifact class there is; a single operator reviewing nightly dream diffs will
decay to LGTM within weeks.

Treatment: the git-diff guard is *forensic* (undo after harm is noticed), not *preventive*.
Preventive guards must be structural: the append-only-expansion rule (§2.3b), hard caps on what
a single dream pass may change (max N supersessions/deletions per pass — halt and flag on
breach), bounded per-dream diffs (candidate cap), one-line dream commit summaries
(`+3 concepts, 2 merged, 1 pruned` — already in proposal §7.5), a **weekly digest** instead of a
daily review expectation, and mandatory human review reserved for the tiers where it changes
behavior (projection/`active.md` changes, cross-scope promotion, rule-skill promotion) rather
than every merge.

---

## 3. The competitive landscape moved: the harness itself is converging on this design

- **Auto memory is native and on by default** (v2.1.59+): Claude writes its own
  `~/.claude/projects/<project>/memory/MEMORY.md` index (first 200 lines / 25KB loaded per
  session) plus on-demand topic files; per-project, machine-local; `autoMemoryDirectory`
  configurable.[^MemoryDocs]
- **"Auto Dream" — a native nightly consolidation — is rolling out behind a server-side flag**:
  a four-phase pass (orient → gather signal from session transcripts → consolidate: merge
  duplicates, resolve contradictions, absolutize dates → prune and re-index MEMORY.md under the
  200-line threshold), triggered at ~24h + >5 sessions since last run; community skills already
  replicate it. Availability is flag-gated and not universal — verified as concept, unverified
  as a dependable API.[^AutoDream][^DreamSkill]
- **Per-subagent persistent memory exists natively** (v2.1.33+), per §1.2 — with the #57507
  allowlist bug caveat.[^SubagentDocs][^SubagentMemoryBug]

**Consequences.**

1. *Validation*: Anthropic independently building trajectory-signal-gathering + scheduled
   consolidation is the strongest available evidence that the proposal's core loop is the right
   shape.
2. *Collision*: the proposal assigns `/dream` to read-and-prune `MEMORY.md` while native Auto
   Dream consolidates the same file on its own clock — a **two-writer conflict with no
   coordination story**.
3. *Scope transfer*: native machinery now covers per-project capture and consolidation *without
   building anything*. The bespoke layer's defensible remit shrinks to what native does not do:
   **cross-project global knowledge as a reviewable git repo; typed/schema'd concepts;
   external-source ingest with provenance; human-gated promotion to skills; the project store
   committed with the code**. The build plan should be re-scoped so phases duplicate nothing the
   harness ships. Concretely: consider pointing `autoMemoryDirectory` *into* the store's
   `short-term/` (making native capture the ingest mechanism), or scope `/dream` to `knowledge/`
   only and let native Auto Dream own `MEMORY.md`, consuming its *output* as the inbox.

---

## 4. NEW BLOCKING RISK — memory poisoning (absent from proposal §9 entirely)

Proposal §9.5 covers *outbound* leakage (secrets pushed to GitHub). The *inbound* threat is
undocumented there, and it is the architecture's worst gap: the design builds a pipeline from
**untrusted input to always-on trusted context**. Web pages read mid-session and `/ingest`ed
documents flow into trajectories → short-term notes → consolidated concepts → `active.md` →
*every future session's context*. That pipeline is a documented attack class:

- **CVE-2026-21852** (disclosed April 2026): a malicious npm postinstall appended instructions
  to Claude Code's `MEMORY.md`; the harness loaded the first 200 lines with high authority every
  session. Anthropic's fix (v2.1.50/v2.2) **removed user memories from the system prompt** to
  reduce their authority. The proposal's store reproduces this surface and *widens* it (more
  files, more writers) — and §5's "MEMORY.md as the inbox" automatically promotes the exact file
  the CVE targeted into a durable, cross-session store.[^MemoryPoisonCve]
- **SpAIware** demonstrated persistent spyware planted in ChatGPT long-term memory via indirect
  prompt injection in web content — attack and effect temporally decoupled. Systematic studies
  report attack success rates against LLM agent memory systems of
  **80–99%**.[^MemoryPoisonSurvey]
- **The dream loop adds a laundering mechanism the attacks love**: a poisoned short-term note
  that survives to `knowledge/` carries a legitimate-looking `provenance` entry; two poisoned
  trajectories = `review_count: 2` = "corroborated" = auto-promoted to `active`. The
  consolidator can convert a one-shot injection into a high-confidence permanent rule — and the
  `CLAUDE.md` `@`-import projection (unlike post-fix auto memory) still lands in context with
  instruction-like authority.
- **The nightly headless pass runs with no human present** — the one moment an operator might
  notice odd content is skipped by design.
- Independent corroboration that this is the format's known weak point: OKF community discussion
  flags agent-updated bundles as an indirect-prompt-injection vector.[^OkfDeepDive]

**Required changes (blocking before Phase 1):**

1. **Trust tiers in provenance**: `operator-confirmed` > `trajectory-derived` >
   `external-ingest`. Tier caps the maximum status reachable without a human gate.
2. **External-ingest content never auto-promotes**: concepts whose provenance chain includes
   `url:` or third-party `file:` sources MUST NOT reach `active` or any projection without
   explicit human confirmation — a **permanent gate, not a confidence threshold**; `/ingest`
   output is quarantined at `candidate`.
3. **Injection screening at capture and at promotion**, symmetric with the outbound
   secret-scrub: instruction-shaped content in a `fact`, imperative verbs addressed to the
   agent, "ignore previous" phrasing, tool-use directives inside ingested text → flag, don't
   consolidate.
4. **Corroboration must come from independent provenance** — two notes tracing to the same
   source (same URL, same package, same session) count once.
5. **De-authorize the projection voice**: projections render concepts as *reference knowledge*,
   not instruction-voiced text, wherever possible — reduce the authority of the surface.

---

## 5. H3 — Cadence: hybrid design is right; the risk is the scheduler, not the clock

- Idle-time consolidation is a mainstream production pattern with a name: Letta's **sleep-time
  compute** — background agents rewrite/derive memory while the primary agent is idle; one
  implementation commits reflections to an isolated git branch to avoid
  contention.[^LettaSleep]
- The Stanford generative-agents architecture triggers reflection on an **importance-sum
  threshold** (~2–3×/day in practice) — event-thresholded, not clock-driven; its retrieval
  combines recency (exponential decay), importance, and relevance.[^GenerativeAgents]
- **Eager per-note LLM processing is measurably wasteful**: RecMem shows eager consolidation
  (LLM-processing every incoming item) wastes 77–87% of construction tokens versus
  recurrence-triggered consolidation, *with no accuracy gain from eagerness* — consolidate only
  when an item accumulates enough semantically similar neighbors.[^RecMem]
- Native Auto Dream's ~24h + >5-sessions trigger is itself a hybrid clock+threshold
  gate.[^AutoDream]

**Synthesis**: the proposal is already the right two-level shape — event-driven *capture*
(Stop/PreCompact hooks) plus a nightly *sweep* (not the sole trigger). Adjustments: (a)
**threshold-gate the nightly pass** (skip when fewer than N new candidates — mirrors Auto
Dream's gate, saves cost, avoids no-op commits); (b) add an **event-threshold fallback** — when
pending short-term candidates exceed N, surface a "run /dream" nudge (SessionStart already
planned to carry exactly this line), so consolidation still happens if the scheduler never runs;
(c) inside the pass, promote on recurrence (`review_count ≥ 2` is already the criterion) and
make the *processing* lazy too, per RecMem; (d) treat headless reliability as the real risk —
established guidance is to run a workflow interactively until it is boringly predictable
*before* scheduling it headless, which the phased plan (interactive `/dream` in Phase 2,
schedule in Phase 5) already respects.[^HeadlessGuide] No change to the daily default is
warranted.

---

## 6. H4 — Complexity: mostly earns its keep; simplifications; two false premises

### 6.1 Evidence the lifecycle machinery is not gold-plating

- **Context rot is measured**: Chroma's 18-model study shows output quality degrades as input
  grows, and *irrelevant* context degrades it fast — even single distractors hurt; quality of
  context beats quantity.[^ContextRotChroma] An unbounded CLAUDE.md/MEMORY.md pile is therefore
  not merely untidy — it is a measured performance regression. The bounded, curated `active.md`
  projection is evidence-backed context engineering. (This *verifies*, rather than inherits, the
  reviewer's "unbounded pile" claim.)
- **Instruction adherence has a budget**: frontier models reliably follow roughly 150–200
  instructions, of which Claude Code's own system prompt consumes ~50; practitioner guidance
  converges on <200 lines per always-loaded file, with degradation observable past ~80 dense
  rule-lines.[^InstructionBudget]
- **Decay is the under-provisioned lever, not the over-provisioned one**: practitioner
  literature calls decay "the lever most agent memory systems skip, and the one that matters
  most for long-running agents"; half-life decay reinforced by fresh evidence is a standard
  form, and an empirically tuned importance half-life of ~29 days brackets the proposal's
  14-day short-term / 60-day candidate windows — the guesses are in the evidenced
  band.[^MemoryEviction][^ConsolidationProblem][^MemorySurvey]

### 6.2 Simplifications a pragmatist should take

- **Drop the stored `confidence` float in v1.** `review_count`, `last_seen`, `status`, and
  provenance tier are *observable facts*; a stored 0.0–1.0 confidence is a synthetic number with
  admitted-guess thresholds (proposal §9.6) and known failure modes: LLM-assigned scores drift
  across model versions, add a model call per write, and exhibit "runaway
  certainty".[^MemoryEviction][^MemorySurvey] The one strong benchmark win for
  confidence-bearing memory (ALFWorld 59.9 vs 28.7) is for *belief distributions over uncertain
  conclusions in partially observable environments* — not this workload (curated operator
  knowledge).[^BeliefMemory] Derive activation from observables (`status: active` AND
  `review_count ≥ 2` AND `last_seen` within window AND trust tier sufficient); keep the schema
  slot for later. This deletes the worst-tuned arithmetic without losing the ladder, and deletes
  proposal §9.6 (threshold tuning) as a risk item.
- **The projection needs a hard budget, not just a quality gate.** `active.md` must carry a
  **hard line/entry cap** in `.knowledge.toml`, with rank-by-(`review_count`, recency) eviction
  into the on-demand bundle — otherwise a healthy store eventually poisons its own projection
  with volume.[^ContextRotChroma][^InstructionBudget]
- **Collapse the double injection.** Proposal §5 has `active.md` arriving via both `@`-import
  and a SessionStart hook ("belt-and-suspenders") — double context cost and two failure surfaces
  where the docs offer a third, simpler native path: a generated, path-scoped file under
  `.claude/rules/` (§1.2). Pick exactly one projection channel; `.claude/rules/` is the
  pressure-relief valve for the context budget.
- **Single-operator YAGNI is confirmed as *partial***: the project-store PR-ratification flow
  (proposal §7.5 step 5) is latent value for a one-person suite — keep optional, off by default.
  But the global/project split itself costs little and maps to native precedence, so it stays.
- Against over-thinning (disconfirming blue's own H4 lean): the files-win consensus for small
  corpora *supports the substrate* and cautions only against the arithmetic; the lifecycle
  ladder (capture → corroborate → promote → decay) is precisely the "judgment" layer that
  consensus says matters. Keep the ladder; simplify its numbers.[^FilesWin][^VectorOverkill]

### 6.3 Two false premises found by local verification (critical-stance)

1. Proposal §9.5 / Phase 5 cite "the port plan's **existing** secret-scrub (`git grep`
   denylist)". **No such gate exists.** The suite contains a semantic guardrail rule whose
   deterministic hook is future-tense ("a PreToolUse hook *will* enforce…"), a git history-scrub
   cheatsheet, and a "PII-scrubbed" convention note — no scrub tooling. The gate must be
   *built*, and ad-hoc `git grep` is the wrong tool: use a maintained scanner
   (gitleaks / detect-secrets class) plus capture-time redaction — claude-mem's `<private>` tag
   exclusion is a pattern worth stealing.[^LocalRepoScrub][^ClaudeMem]
2. Proposal §7.6 claims "`docs/scheduling.md` in sleeper-service **already documents** the
   recipes". **The file does not exist**; sleeper-service is currently a stub
   (plugin.json + README). The scheduling story is planned, not shipped — `/dream` inherits a
   dependency on unbuilt work, which belongs in the phase plan, not the
   assumptions.[^LocalRepoSleeper]

---

## 7. H5 — Alternatives: nothing dominates; what to steal from each

- **claude-mem** (46k stars) is the strongest adopt-instead candidate: plugin-native,
  hook-driven session capture, AI compression, local storage, layered retrieval (~10× token
  efficiency claimed). It fails the suite's constraints where they bind: storage is SQLite (not
  human-readable markdown, not git-diffable, not PR-reviewable), no
  project-store-committed-with-code, no promotion ladder to skills, third-party dependency for
  load-bearing infrastructure. **Steal**: `<private>` capture-time redaction; proof that
  hook-based trajectory capture works at ecosystem scale.[^ClaudeMem]
- **basic-memory** is the closest philosophical match (markdown source of truth + derived SQLite
  index + MCP, no server/cloud) and is the existence proof for §1.5's files-plus-index endgame;
  it lacks lifecycle/decay/promotion, PR flow, and agent-memory integration, and adds an MCP
  server dependency — it complements rather than replaces.[^BasicMemory]
- **mem0 / Letta / Zep**: all service/daemon/database-bound — they violate the suite's
  no-daemon, git-reviewable constraints on the same grounds that rejected FUSE. **Steal**:
  mem0's retrieve-then-classify dedup pipeline (§2.3a), Letta's sleep-time framing and
  isolated-branch commits (§5), Zep's fact-supersedence-with-validity-interval as the conceptual
  model behind `supersedes`/`last_seen`.[^MemZero][^LettaSleep][^ZepGraphiti][^ZepCritique]
- **Native-surfaces-plus-thin-skill** (the H4 thin design) is not a competitor once the
  context-rot and curation evidence is in (§6) — but its best half-ideas survive as the
  `.claude/rules/` projection channel and the `autoMemoryDirectory` ingest collapse. And the
  harness's own trajectory (§3) means the thin design's *capture* half is arriving for free;
  nothing surveyed offers project-store-committed-with-code. **Bespoke remains justified for the
  shrunken remit; no external adoption dominates.**

---

## 8. Changes required before implementation (consolidated, both lanes)

| # | Change | Grade |
|---|---|---|
| 1 | Add memory-poisoning threat model (§4): provenance trust tiers; external-ingest permanently human-gated from projections; injection screening at capture and promotion; independent-source corroboration; de-authorized projection voice | **Blocking** |
| 2 | Fix proposal §5 agent-memory row: project into the harness's fixed `agent-memory/` paths; define merge for bidirectional writes; contingent on issue #57507 resolution | **Blocking** (correctness) |
| 3 | Build the secret-scrub gate (gitleaks/detect-secrets class + capture-time `<private>`-style redaction) — it does not exist to be reused | **Blocking** for any remote push |
| 4 | Re-scope against native machinery (§3): resolve the `/dream` vs Auto Dream two-writer conflict explicitly (own `knowledge/` only, or consume Auto Dream output as inbox); evaluate `autoMemoryDirectory`-into-store; drop bespoke work duplicating native capture | High |
| 5 | Phase 0 adds a hook-fire test matrix (interactive × headless × Stop/PreCompact/SessionStart) and an import-approval-under-`-p` check; transcript parsing behind one version-pinned module with a fallback (e.g. `/export`) | High |
| 6 | Specify v1 dedup candidate retrieval as whole-bundle-in-context with a named ceiling (~300–500 concepts or first observed dedup miss) that triggers the deferred SQLite/embedding index | High |
| 7 | Hard token/line budget for `active.md` with rank-based eviction; single projection channel — prefer generated, path-scoped `.claude/rules/` files over `@`-import + SessionStart double injection | High |
| 8 | Append-only expansion: claims immutable after promotion; change = supersede (new file); no destructive body rewrites; fan-in cap per dream pass | High |
| 9 | Scheduled dream pass: sequential subagents only (headless fan-out hang #56540); parallel fan-out reserved for interactive invocation | High |
| 10 | Drop stored `confidence` float in v1; derive activation from `status` + `review_count` + `last_seen` + trust tier; delete threshold-tuning risk | Medium (simplification) |
| 11 | Threshold-gate the nightly run (skip < N candidates); event-threshold fallback nudge so the system degrades gracefully without the scheduler; lazy per-note processing (RecMem) | Medium |
| 12 | Demote review-by-git-diff to forensic control: per-pass change caps (max supersessions/deletions, halt-and-flag on breach); weekly digest + tier-gated human review instead of nightly diff review | Medium |
| 13 | Projection-health check in SessionStart (silent-dead external `@`-import) | Medium |
| 14 | Reframe OKF as pinned convention (`okf_version: "0.1"`); correct the §7.6 and §9.5 "already exists" claims; note `index.md`/`log.md` carry no frontmatter | Low |

## 9. Risk grading (likelihood × impact × complexity-to-fix)

| Risk | L | I | Fix cost | Disposition |
|---|---|---|---|---|
| Memory poisoning via ingest/inbox (§4) | Med (single operator, but npm-CVE precedent; 80–99% reported attack success) | High (persistent compromise) | Med | Fix — **blocking** |
| Consolidation rewrite-corruption (§2.3b) | High over months | High (silent knowledge loss) | Low (append-only rule) | Fix |
| Agent-memory row wrong / bidirectional collision (§1.2) | Certain as written | Med (destroys agent learning) | Low (project into harness path + merge) | Fix — blocking correctness |
| Secret/PII leakage on remote push (§6.3) | Med | High | Med (build scanner gate) | Fix — blocking for push |
| Native Auto Dream two-writer conflict (§3) | High if flag lands | Med (churn, lost notes) | Low (scope split) | Fix |
| Headless hooks/fan-out failures (§1.3) | High in cron context | Med (silent no-op nights) | Low (sequential; test matrix) | Fix |
| Dedup recall shortfall at scale (§2.3a) | Med (scale-dependent) | Med (fragmentation) | Low now / Med later | Fix cheap path now, name ceiling |
| Unreviewed bot commits (§2.4) | High | Med | Low (caps + digest) | Fix |
| Projection context-rot (§6) | Med | Med (adherence loss across all rules) | Low (hard cap) | Fix |
| Confidence-float drift (§6.2) | Med | Low | Negative (removal simplifies) | Fix by deletion |
| Transcript format churn (§1.4) | Med | Low (feature degrades, recoverable) | Low (parser module + fallback) | Fix |
| OKF v0.1 drift / abandonment (§1.1) | Low | Low (profile pinned; custom keys legal; degrades to plain markdown) | — | **Risk-accept** (proposal §9.7 stands, restated as design stance) |
| Multi-machine store divergence | Low (single operator, one box) | Low | Med (sync protocol) | **Risk-accept** — YAGNI; git remote is the sync story if ever needed |
| Project-store PR-ratification flow unused | High (one-person suite) | Low | — | **Risk-accept** — keep optional, off by default |

## 10. Unverified items (labeled, not laundered)

- The internal FUSE prior art, the OpenClaw dream-diary degradation anecdote, and the
  AgentOrange `continuous_learning` aspect's "battle-tested" status — all internal artifacts
  cited by the proposal without independent corroboration in either lane.
- The ARC-AGI 54% regression figure — secondary commentary only.[^AgentsDumber]
- Native Auto Dream availability — verified as concept and community replication, unverified as
  a dependable API (server-side flag).[^AutoDream][^DreamSkill]

---

## Footnotes

[^OkfSpec]: *Open Knowledge Format (OKF) Specification*, GoogleCloudPlatform/knowledge-catalog `okf/SPEC.md` (GitHub), https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md, accessed 2026-07-12. v0.1 Draft; `type` sole required field; `index.md`/`log.md` reserved without frontmatter; `okf_version` in root index; producers MAY add keys, consumers must tolerate unknown keys.
[^OkfBlog]: *How the Open Knowledge Format can improve data sharing*, Google Cloud Blog, https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing/, accessed 2026-07-12. Announced 2026-06-12; "just markdown, just files, just YAML frontmatter"; hostable in any git repo.
[^OkfSkeptic]: *Google Cloud Introduces Open Knowledge Format (OKF)*, MarkTechPost, June 16 2026, and community adoption commentary (owox.com; dev.to/maskaravivek), accessed 2026-07-12. "Markdown with metadata" rebrand critique; abandonment risk.
[^OkfDeepDive]: *Is OKF Worth Adopting Yet? A Deep Dive into Google's Open Knowledge Format*, ewandel.de, accessed 2026-07-12. v0.1 breaking-change risk; link brittleness on rename; agent-updated bundles as indirect-prompt-injection vector.
[^MemoryDocs]: *How Claude remembers your project*, Claude Code documentation, https://code.claude.com/docs/en/memory, accessed 2026-07-12. `@`-import semantics (4-hop max, code-span skip, external-import approval dialog with silent-disable on decline, imports load at launch and consume context), MEMORY.md location and 200-line/25KB load (auto memory native v2.1.59+), `autoMemoryDirectory`, `CLAUDE_CODE_DISABLE_AUTO_MEMORY`/`autoMemoryEnabled`, `.claude/rules/` incl. user-level rules, `paths:` frontmatter, symlinks, load order; CLAUDE.md delivered as user message, not system prompt.
[^SubagentDocs]: *Create custom subagents*, Claude Code documentation, https://code.claude.com/docs/en/sub-agents (plus shanraisshan/claude-code-best-practice agent-memory report), accessed 2026-07-12. `memory: user|project|local` (v2.1.33+); user scope at `~/.claude/agent-memory/<name>/`, project scope at `.claude/agent-memory/<name>/` ("shareable via version control"); first 200 lines of agent MEMORY.md injected.
[^SubagentMemoryBug]: *[BUG] `memory:` field in subagent frontmatter not functional — v2.1.137; tools allowlist appears to override auto-enable*, anthropics/claude-code issue #57507, https://github.com/anthropics/claude-code/issues/57507, accessed 2026-07-12.
[^HooksDocs]: *Hooks reference*, Claude Code documentation, https://code.claude.com/docs/en/hooks, accessed 2026-07-12. SessionStart sources and `additionalContext`/`initialUserMessage` (the latter explicitly applies in `-p`); Stop `last_assistant_message`; PreCompact matchers.
[^HeadlessHookBugs]: GitHub issues anthropics/claude-code #20063 (hooks don't run in headless mode), #38651 (Stop hook empties `claude -p` result), #40506 (PreToolUse not firing in `-p`), #37559 (hook docs vs. behavior), accessed 2026-07-12. Open bug record for hooks under non-interactive mode.
[^HeadlessDocs]: *Run Claude Code programmatically*, Claude Code documentation, https://code.claude.com/docs/en/headless, accessed 2026-07-12. `claude -p` waits for background subagents; 10-min cap via `CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS`.
[^HeadlessHang]: *claude -p headless under non-TTY parent: parallel Task fan-out hangs*, anthropics/claude-code issue #56540, https://github.com/anthropics/claude-code/issues/56540, accessed 2026-07-12.
[^LocalTranscripts]: Local inspection, `~/.claude/projects/C--Users-gbloc-Projects-AgentOrange/*.jsonl`, Claude Code v2.1.207, this machine, 2026-07-12. Primary-source verification of transcript path and line schema (§1.4).
[^TranscriptFormat]: *Claude Code JSONL transcript format explained*, claude-dev.tools + simonw/claude-code-transcripts, https://claude-dev.tools/docs/jsonl-format, accessed 2026-07-12. Path/schema confirmed; "internal to Claude Code and changes between versions."
[^BasicMemory]: *basic-memory*, basicmachines-co (GitHub), https://github.com/basicmachines-co/basic-memory, accessed 2026-07-12. Local-first markdown knowledge graph + local SQLite index, MCP server, no cloud.
[^AgenticDigest]: *Git-based LLM wikis move agent memory into Markdown*, The Agentic Digest, accessed 2026-07-12. Wuphf: local markdown + git + BM25/SQLite index; survey of filesystem-markdown memory family and its cost ("at the cost of scale and automatic semantic search").
[^ClaudeMem]: *claude-mem*, thedotmack (GitHub; docs.claude-mem.ai; Augment Code coverage), accessed 2026-07-12. 46k-star Claude Code plugin: hook-based capture, AI compression, local SQLite + FTS, layered retrieval (~10× token efficiency claimed), `<private>` tag exclusion.
[^SqliteMemory]: *sqlite-memory*, sqliteai (GitHub), https://github.com/sqliteai/sqlite-memory, accessed 2026-07-12. Markdown-based agent memory with semantic search + hybrid retrieval; precedent for the deferred index.
[^MemoryMdProblem]: *The MEMORY.md Problem: Why Local Files Fail at Scale*, DEV Community (anajuliabit), and *memweave* (Towards Data Science), accessed 2026-07-12. Flat-file failure modes (token bloat, no retrieval/supersedence); counter-nuance: "early-stage agents don't have a retrieval problem — they have a curation problem."
[^ZepCritique]: *Markdown is not agent memory*, Zep blog, accessed 2026-07-12. Compounding errors, no fact supersedence, concurrent-writer divergence; vendor of the competing temporal-knowledge-graph product — motivated but substantively argued.
[^ZepGraphiti]: *Zep: A Temporal Knowledge Graph Architecture for Agent Memory*, arXiv 2501.13956 + getzep/graphiti (GitHub), https://arxiv.org/html/2501.13956v1, accessed 2026-07-12. LLM contradiction detection against semantically related edges; invalidate-not-delete with validity windows.
[^ConsolidationProblem]: *The Consolidation Problem in Agent Memory*, Hindsight (Vectorize) blog, 2026-05-21, https://hindsight.vectorize.io/blog/2026/05/21/agent-memory-consolidation, accessed 2026-07-12. Consolidation vs. lossy compaction; 2,000-fact/36.7× compression study with 60% irretrievable loss; summarization drift; keep-raw-linked mitigation; four levers (importance/merge/decay/eviction); "decay is the lever most systems skip"; stale/contradictory/near-duplicate accumulation degrades behavior even when retrieval works.
[^FaultyMemories]: *Useful Memories Become Faulty When Continuously Updated by LLMs*, Zhang et al., arXiv 2605.12978, https://arxiv.org/pdf/2605.12978, accessed 2026-07-12. Repeated LLM update cycles corrupt memories (interference, meaning drift, loss of specifics); utility rises then declines; intensifies with update frequency.
[^AgentsDumber]: *Long-Term Memory Is Making Agents Dumber*, Johnson Lee blog, 2026-05-20, https://johnsonlee.io/2026/05/20/faulty-agent-memory.en/, accessed 2026-07-12. Secondary commentary; ARC-AGI 54% figure reported here, not independently verified.
[^MemorySurvey]: *Memory for Autonomous LLM Agents: Mechanisms, Evaluation, and Emerging Frontiers*, arXiv 2603.07670, https://arxiv.org/html/2603.07670v1, accessed 2026-07-12. Summarization drift and semantic intensification; importance-score drift across model versions; ~29-day empirical half-life.
[^MemZero]: *Mem0: Building Production-Ready AI Agents with Scalable Long-Term Memory*, mem0ai/mem0 (GitHub) + paper coverage (emergentmind.com; deepwiki), https://github.com/mem0ai/mem0, accessed 2026-07-12. Two-phase pipeline: embed → vector-retrieve top-K neighbors → LLM classifies ADD/UPDATE/DELETE/NOOP.
[^ParaphraseGap]: *Semantic search as extractive paraphrase span detection*, Language Resources and Evaluation (Springer), https://link.springer.com/article/10.1007/s10579-023-09715-7, + MDPI *Transformer Models for Paraphrase Detection*, accessed 2026-07-12. Semantic beats lexical by 11–20+ points; high-semantic/low-lexical-overlap gap (99%+ similarity with single-digit BLEU).
[^LLMJudgeDedup]: *Semantic Needles in Document Haystacks: Sensitivity Testing of LLM-as-a-Judge Similarity Scoring*, arXiv 2604.18835, https://arxiv.org/pdf/2604.18835, accessed 2026-07-12. Threshold-dependent pairwise dedup judgment reliability (cosine ≥0.95 all true duplicates; 0.85–0.87 ~1.5%).
[^BotReviewFatigue]: *Reducing Alert Fatigue via AI-Assisted Negotiation: A Case for Dependabot* (arXiv 2502.06175); IEEE TSE study of dependency-bot PRs (arXiv 2206.07230); Pixee merge-rate analysis, accessed 2026-07-12. ~54% Dependabot merge rate; rubber-stamping vs. queue abandonment as the documented failure pair.
[^UnreviewedPRs]: *On the Footprints of Reviewer Bots' Feedback on Agentic Pull Requests in OSS GitHub Repositories*, arXiv 2604.24450, https://arxiv.org/html/2604.24450v1, accessed 2026-07-12. 61.38% of agent PRs no recorded review; 71.58% of review comments agent-authored.
[^AIApprovingPRs]: *AI is approving our pull requests*, fin.ai / Intercom engineering blog, https://ideas.fin.ai/p/ai-is-approving-our-pull-requests, accessed 2026-07-12. Rubber-stamping under queue pressure.
[^LettaSleep]: *Sleep-time Compute*, Letta blog + Letta docs (sleeptime architectures) + community best-practices forum, https://www.letta.com/blog/sleep-time-compute/, accessed 2026-07-12. Background agents consolidate/dedup/prune while primary agent idle; isolated git-branch commits to avoid contention.
[^GenerativeAgents]: Park et al., *Generative Agents: Interactive Simulacra of Human Behavior* (2023), arXiv 2304.03442 (via memx.app; subodhjena.com architecture summaries), accessed 2026-07-12. Reflection triggered when accumulated importance exceeds threshold (~150; ~2–3×/day in practice); retrieval = recency (exponential decay) + importance + relevance.
[^RecMem]: *RecMem: Recurrence-based Memory Consolidation for Efficient and Effective Long-Running LLM Agents*, arXiv 2605.16045, https://arxiv.org/html/2605.16045v1, accessed 2026-07-12. Eager consolidation wastes 77–87% construction tokens with no accuracy gain; recurrence-triggered consolidation.
[^HeadlessGuide]: *Claude Code in CI/CD and Headless Automation* (hidekazu-konishi.com) and MindStudio headless-mode guides, accessed 2026-07-12. Headless as the last pattern adopted; run interactively until predictable.
[^MemoryPoisonCve]: *Identifying and remediating a persistent memory compromise in Claude Code*, Cisco Blogs (CVE-2026-21852 disclosure, April 2026), https://blogs.cisco.com/ai/identifying-and-remediating-a-persistent-memory-compromise-in-claude-code, and *CVE-2026-21852: Agent Memory Poisoning in Your Codebase*, omegamax.co, https://omegamax.co/blog/agent-memory-poisoning-cve-2026, accessed 2026-07-12. Malicious npm postinstall → MEMORY.md instructions treated as authoritative every session; fix (v2.1.50/v2.2) removed user memories from system prompt.
[^MemoryPoisonSurvey]: *From Untrusted Input to Trusted Memory: A Systematic Study of Memory Poisoning Attacks in LLM Agents*, arXiv 2606.04329, https://arxiv.org/pdf/2606.04329; Christian Schneider, *Memory poisoning in AI agents: exploits that wait*; SpAIware coverage, accessed 2026-07-12. 80–99% reported attack success rates; temporal decoupling of attack and effect.
[^MemoryEviction]: *Agent Memory Eviction: 8 Policies That Stop Stale Tool Decisions* (Medium, Bhagya Rana) and *Governing Evolving Memory in LLM Agents (SSGM)* (arXiv 2603.11768), accessed 2026-07-12. Half-life decay reinforced by evidence; inferred memories decay faster; decay as the most-skipped, most-needed lever; confidence calibration / runaway-certainty risk.
[^ContextRotChroma]: *Context Rot: How Increasing Input Tokens Impacts LLM Performance*, Chroma Research, July 2025, accessed 2026-07-12. 18 frontier models degrade with input length; irrelevant distractors degrade sharply; vendor caveat (Chroma sells vector DBs) noted.
[^InstructionBudget]: *Your CLAUDE.md Is Probably Too Long*, tianpan.co, 2026-02-14, https://tianpan.co/blog/2026-02-14-writing-effective-agent-instruction-files (+ MindStudio context-rot analysis), accessed 2026-07-12. ~150–200 instruction adherence budget, ~50 consumed by system prompt; degradation past ~80 dense rule-lines.
[^AutoDream]: *Claude Code Dreams: Anthropic's New Memory Feature*, claudefa.st, https://claudefa.st/blog/guide/mechanics/auto-dream, + *Auto Memory and Auto Dream* (antoniocortes.com, 2026-03-30), accessed 2026-07-12. Four-phase pass (orient/gather/consolidate/prune); ~24h + >5 sessions trigger; server-side flag rollout — availability unverified as stable API.
[^DreamSkill]: *dream-skill*, grandamenium (GitHub), https://github.com/grandamenium/dream-skill, accessed 2026-07-12. "Replicates Anthropic's unreleased auto-dream feature," 4-phase, 24h auto-trigger — evidence of community replication and flag-gated status.
[^FilesWin]: *Forget RAG: The Best AI Agent Memory Is a Plain Text File*, voxos.ai, https://voxos.ai/blog/how-to-give-ai-coding-agents-long-term-m/index.html (+ dev.to *All of Them Use Flat Files*), accessed 2026-07-12. Files-win consensus for small corpora; judgment, not infrastructure, is the binding constraint.
[^VectorOverkill]: *Did Agents Kill Vector Search? The Honest, Scale-Dependent Answer*, thedataexperts.us, https://www.thedataexperts.us/writing/vector-db-vs-files-agents-retrieval.html, accessed 2026-07-12. Filesystem agents beat vector pipelines on small complex corpora; advantage inverts at scale.
[^BeliefMemory]: *Belief Memory: Agent Memory Under Partial Observability*, arXiv 2605.05583, https://arxiv.org/html/2605.05583v1, accessed 2026-07-12. ALFWorld 59.88 → 28.71 when probabilistic memory collapsed to deterministic — the confidence-helps evidence, scoped to partial observability.
[^LocalRepoScrub]: Local verification, special-circumstances repo, 2026-07-12: `grep -i secret|scrub|denylist` across `*.md` — no secret-scrub gate artifact; `plugins/prosthetic-conscience/skills/agent-guardrails/SKILL.md` says a deterministic PreToolUse hook "will enforce" (future tense).
[^LocalRepoSleeper]: Local verification, special-circumstances repo, 2026-07-12: `plugins/sleeper-service/` contains only `.claude-plugin/plugin.json` and `README.md`; no `docs/scheduling.md`.
