# Blue CHANGELOG

## Round 0 — initial synthesis (2026-07-12)

Created `blue/report.md` by structural union of `blue/candidates/lane-1.md` (H1-deep) and
`blue/candidates/lane-2.md` (H2-deep). Also created `blue/frontier.md` reconstructing the
H1–H5 hypotheses as the lanes tested them. No substantive content dropped; both candidate
drafts preserved unmodified.

Merge operations:

- **Reorganized** into shared frame: verdict → H1 substrate (§1) → H2 consolidation (§2) →
  native-harness convergence (§3, lane 2 unique) → memory poisoning (§4, both lanes,
  merged) → H3 cadence (§5) → H4 complexity (§6) → H5 alternatives (§7) → consolidated
  change list (§8) → risk grading (§9) → unverified items (§10) → footnotes.
- **Deduplicated overlapping claims** (kept the more detailed variant, merged unique
  details from both): OKF spec verification (lane-1 §1.1 + lane-2 §4.1); transcript JSONL
  leaf-node verification (lane-1 §1.4 + lane-2 §4.2 — kept lane-1's schema detail plus
  lane-2's version-pinned-contract-with-fallback treatment); `@`-import semantics and
  silent-disable (both); `.claude/rules/` projection alternative (both); agent `memory:`
  fixed-path correction (lane-1) merged with issue #57507 caveat (lane-2); memory-poisoning
  sections (lane-1 §4 + lane-2 §2 — union of required changes: trust tiers,
  permanent ingest gate, injection screening, independent-provenance corroboration,
  de-authorized projection voice); consolidation-loss evidence (lane-1 §2 + lane-2 §1.1);
  bot-review fatigue (lane-1 Dependabot + lane-2 agent-PR 61.4% figures — both kept);
  confidence-float removal (both — union of rationales incl. lane-2's BeliefMemory
  scoping); dedup candidate-retrieval gap (both — kept lane-2's paraphrase/LLM-judge
  evidence, lane-1's ~300–500 trigger); cadence findings (lane-1 threshold-fallback +
  lane-2 RecMem laziness + nightly gate — all kept); alternatives survey (union of
  claude-mem / basic-memory / mem0 / Letta / Zep dispositions and steal-lists).
- **Merged change lists** into one graded 14-item table (§8): lane-1 items 1–10 and lane-2
  items 1–10 map onto merged items with no loss; grades reconciled (highest grade wins on
  overlap).
- **Merged risk tables** (§9): lane-2's grading table extended with lane-1-specific rows
  (agent-memory correctness, secret/PII outbound, projection context-rot already shared)
  and a third risk-accepted row (PR-ratification flow, from lane-1 §5 partial-YAGNI).
- **Footnote union with label reconciliation** (41 → 38 after dedup): merged
  `OkfSpec`/`OKFSpec`, `OkfBlog`/`OKFBlog`, `MemoryDocs`/`ClaudeMemoryDocs`,
  `SubagentDocs`/`SubagentMemory`, `ConsolidationProblem`/`HindsightConsolidation`,
  `MemZero`, `LettaSleep`, `GenerativeAgents`, `FaultyMemories`, `BasicMemory`,
  `MemoryPoisonCve`/`CiscoMemoryCVE`/`OmegamaxCVE`,
  `MemoryPoisonSurvey`/`MemoryPoisoningStudy`. Split the label collision on `ContextRot`
  (different sources) into `ContextRotChroma` (Chroma Research) and `InstructionBudget`
  (tianpan.co). Kept `ZepCritique` (Zep blog) and `ZepGraphiti` (arXiv 2501.13956)
  distinct — different sources.
- **Preserved distinctly-sourced near-duplicates** rather than collapsing them: lane-1's
  Dependabot evidence and lane-2's agent-PR evidence both support §2.4; lane-1's Chroma
  context-rot and lane-2's instruction-budget both support §6.1.
- **Unverified-items section** (§10) unions lane-1's labeled internal-artifact caveats with
  lane-2's ARC-AGI and Auto Dream availability caveats.
