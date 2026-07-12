# Frontier hypotheses — memory architecture for Special Circumstances

Topic: is the Open-Knowledge-Format-inspired, git-native design (global + per-project stores,
native CLAUDE.md/MEMORY.md as generated projections, trajectory-to-memory extraction, nightly
dream consolidation) the right memory architecture; risks; alternatives; changes before build.

Each hypothesis states what would be observably true if the candidate answer were right, and
what evidence would confirm or disconfirm it. Searches test these, not wander.

---

## H1 — Substrate holds: the proposal is right because git-native markdown-plus-frontmatter is a proven, sufficient memory substrate for agent suites of this scale

**If true, we would find:**
- The Open Knowledge Format v0.1 actually exists as described (directory of markdown + YAML
  frontmatter, `type` as the only required field, `index.md`/`log.md` reserved names) and is
  stable enough to pin a profile against.
- Working precedents of file/git-based agent memory (no database, no daemon) operating at
  hundreds-to-thousands of concepts without retrieval or navigation collapse — e.g.
  basic-memory, Obsidian-as-agent-memory, git-backed knowledge bases in agent frameworks.
- Claude Code's `@`-import in CLAUDE.md, `MEMORY.md` auto-memory, per-agent `memory:`
  frontmatter, and `Stop`/`PreCompact`/`SessionStart` hooks all behave as the proposal assumes
  (each mapping row in §5 is load-bearing).

**Disconfirmed by:** OKF materially different from the description or already abandoned;
documented scaling ceilings for flat-file memory; any §5 native-surface assumption false
(e.g. MEMORY.md not readable/writable the way §5's ingest arrow needs, hooks unavailable
headless).

## H2 — Dedup is the Achilles heel: the design fails in whole or part because LLM-driven expand-before-append without a semantic index silently loses or fragments knowledge

**If true, we would find:**
- Published failures or measurements of LLM-based memory consolidation (merge/summarize passes)
  degrading content — the OpenClaw "details unavailable" pattern reproduced elsewhere.
- Production agent-memory systems (mem0, Letta/MemGPT, Zep, LangMem…) converging on
  embedding/vector or graph indices for duplicate detection precisely because lexical/title
  matching under-detects overlap.
- Evidence that review-by-git-diff is an inadequate guard in practice (humans don't review
  nightly bot commits; automation bias literature).

**Disconfirmed by:** file-only systems demonstrating acceptable dedup at this scale; evidence
that frontier-model consolidators reliably detect paraphrase-level overlap without embeddings.

## H3 — Cadence is wrong: event-triggered consolidation beats a nightly clock-driven dream loop

**If true, we would find:**
- Memory-architecture literature (generative agents' reflection, MemGPT's context-pressure
  triggers, cognitive-science consolidation models) favoring threshold/event triggers
  (context compaction, N new candidates, session end) over fixed daily batch.
- Failure modes specific to headless scheduled runs: `claude -p` non-interactive workflows
  spawning sub-agents being unreliable/unsupported; scheduled unattended LLM writes to a
  shared store drifting without the human-in-the-loop that interactive triggers get for free.
- Staleness cost: within-day duplicate accumulation between dreams materially degrading the
  active projection.

**Disconfirmed by:** evidence that batch/idle consolidation is standard and adequate (and that
the PreCompact/Stop hooks in §7.1 already cover the event-trigger case, making nightly merely
a sweep, not the sole trigger).

## H4 — Complexity exceeds payoff: a thinner design (native surfaces + a dedup skill + the promotion ladder, no OKF profile, no projections layer, no confidence arithmetic) captures most of the value

**If true, we would find:**
- Evidence that the marginal value of `confidence`/`review_count`/decay arithmetic is
  speculative — no measured benefit in comparable systems, thresholds admitted guesses (§9.6).
- The projection indirection creating real failure surface (generated-file drift, double
  injection via both `@`-import and SessionStart hook, hand-edit collisions) that a direct
  "curated CLAUDE.md section" avoids.
- Single-operator context: suite has one user; multi-collaborator PR-review benefits of the
  project store are mostly latent; YAGNI applies to global-repo push, promote-to-global flow.

**Disconfirmed by:** evidence the lifecycle fields do real work (decay demonstrably prevents
context-rot; confidence gating measurably improves projection quality) or that ungated native
surfaces reliably rot without them (the reviewer's "unbounded pile" claim verified, not
inherited).

## H5 — An existing system beats bespoke: adopting or embedding a maintained memory tool (mem0, Letta, Zep, basic-memory, or Claude Code-native memory alone) dominates building this

**If true, we would find:**
- A maintained open tool already offering file-based or git-compatible stores with
  extraction + consolidation + decay, integrable as an MCP server or CLI, with better dedup
  than title-matching — making phases 1–5 mostly redundant.
- Interop cost of bespoke OKF profile (nobody else reads it) exceeding its portability benefit.

**Disconfirmed by:** surveyed tools all being service/daemon/DB-bound (violating the suite's
no-daemon, git-reviewable, PR-able constraints — the same grounds that rejected the FUSE
prior art) or unmaintained; none offering the project-store-committed-with-code property.

---

**Disconfirming-evidence budget:** ≥1 in 5 searches targets whichever hypothesis currently
looks strongest (expected: H1/H4 tension — searches must actively hunt file-memory scaling
failures and lifecycle-field successes, not just confirmations).
