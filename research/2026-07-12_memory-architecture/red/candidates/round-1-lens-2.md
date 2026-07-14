# Red audit — Round 1, Lens 2: leaf-node citation verification (instance 2 of 3)

**Slice:** §4 (memory poisoning), §5 (H3 cadence), §6 (H4 complexity), §7 (H5 alternatives)
plus their footnotes.
**Method:** followed each load-bearing citation to source; graded corroboration confidence
per statement↔reference pair. Local-repo claims (§6.3) verified on this machine; external
arXiv/CVE/vendor claims verified via web.
**Verdict for this slice:** PASS with 3 low/medium-confidence flags. No falsified citation
found. Every arXiv identifier checked resolves to a real paper with a title matching the
footnote; the CVE is real; the two local "false premise" claims are true.

---

## Citations verified — corroboration HIGH (no action)

- **§6.3 claim #2 — `docs/scheduling.md` does not exist.** Anchor: *"Proposal §7.6 claims
  '`docs/scheduling.md` in sleeper-service **already documents** the recipes'. **The file does
  not exist**; sleeper-service is currently a stub (plugin.json + README)."* Verified on this
  machine: `plugins/sleeper-service/` contains only `.claude-plugin/plugin.json` and
  `README.md`; no `scheduling.md` anywhere in the repo. [^LocalRepoSleeper] corroborated HIGH.
- **§6.3 claim #1 — no secret-scrub gate exists.** Anchor: *"'the port plan's **existing**
  secret-scrub (`git grep` denylist)'. **No such gate exists.**"* Verified: grep for
  `secret-scrub|detect-secrets|gitleaks|denylist` across repo markdown returns only the research
  docs themselves (report/candidates/proposal), no tooling artifact in any plugin.
  [^LocalRepoScrub] corroborated HIGH.
- **§4 — CVE-2026-21852 memory-poisoning phenomenon.** Anchor: *"a malicious npm postinstall
  appended instructions to Claude Code's `MEMORY.md`; the harness loaded the first 200 lines
  with high authority every session."* Cisco disclosure (April 2026) confirms exactly this
  mechanism and persistence; patch in v2.2. [^MemoryPoisonCve] phenomenon corroborated HIGH.
- **§5 — RecMem eager-consolidation waste.** Anchor: *"RecMem shows eager consolidation
  (LLM-processing every incoming item) wastes 77–87% of construction tokens versus
  recurrence-triggered consolidation."* arXiv 2605.16045 is real; source states token cost
  reduced "by up to 87%"; recurrence-trigger framing and "no accuracy gain from eagerness"
  match. [^RecMem] corroborated HIGH (upper bound exact; 77% lower bound is the cross-system
  range, consistent).
- **§7 — claude-mem facts.** Anchor: *"claude-mem (46k stars) … hook-driven session capture,
  AI compression, local storage … `<private>` capture-time redaction."* All confirmed: lifecycle
  hooks, SQLite+FTS5+Chroma, `<private>` tag exclusion. Star count 46.1K is a valid point-in-time
  snapshot (a more recent source shows 65.8K — understated, not wrong). [^ClaudeMem] HIGH.
- **§7 — mem0 / Zep / basic-memory mechanisms.** ADD/UPDATE/DELETE/NOOP classify-against-neighbors
  (mem0), temporal-knowledge-graph invalidate-not-delete (Zep, arXiv 2501.13956), markdown+SQLite+MCP
  (basic-memory) all corroborated. [^MemZero][^ZepGraphiti][^BasicMemory] HIGH.
- **§5 — Generative Agents reflection trigger.** arXiv 2304.03442 real; importance-sum threshold
  (~150) matches the paper's design. [^GenerativeAgents] HIGH.

---

## GAPS (graded; raised, open until closed/rebutted/adjudicated)

### GAP-L2-1 — "80–99% attack success" not pinned to its cited source
- **Location:** §4 NEW BLOCKING RISK — memory poisoning. Quoted: *"Systematic studies report
  attack success rates against LLM agent memory systems of **80–99%**."* Also load-bearing in
  §9 risk table row 1: *"80–99% reported attack success"* justifies the Med likelihood of the
  sole blocking risk.
- **Finding:** The primary paper in [^MemoryPoisonSurvey] (arXiv 2606.04329, "From Untrusted
  Input to Trusted Memory") is real and on-topic — it builds MPBench and shows aggressive
  memory-writing agents are more exploitable — but I could not confirm the specific **80–99%**
  band in that paper. The nearest concrete figure in the neighborhood literature is MINJA's
  "~95% injection / ~70% attack success." The footnote bundles three sources (arXiv paper +
  Schneider + SpAIware coverage); the headline number is not clearly attributable to the primary.
- **Corroboration confidence:** MEDIUM. The threat class is well-corroborated (HIGH); the
  specific quantitative band is not traced to a specific source.
- **Grade:** likelihood-of-error Med · impact Low (the blocking disposition survives even if the
  number is softer — CVE precedent alone carries the risk) · complexity-to-fix Low.
- **Ask of blue:** pin the 80–99% figure to a single citable source and page/section, or soften
  to "reported success rates up to ~95% (MINJA)" with the exact attribution. Not PASS-blocking.

### GAP-L2-2 — CVE-2026-21852 number may conflate two disclosures
- **Location:** §4. Quoted: *"**CVE-2026-21852** (disclosed April 2026): a malicious npm
  postinstall appended instructions to Claude Code's `MEMORY.md`."*
- **Finding:** The Cisco blog cited (title: "persistent memory compromise") fully supports the
  memory-poisoning narrative. However, several vulnerability databases attach **CVE-2026-21852**
  to a differently-framed issue — GitHub Advisory GHSA-jh7p-qr78-84p7 titles it "Leaks Data via
  Malicious Environment Configuration Before Trust Confirmation"; SentinelOne calls it an
  "Information Disclosure Flaw." Possible that the memory-poisoning writeup and the
  info-disclosure CVE are distinct disclosures being merged under one number. Blue's second
  source (omegamax) does tie the number to memory poisoning, so the pairing is defensible.
- **Corroboration confidence:** phenomenon HIGH; exact CVE-number↔memory-poisoning mapping MEDIUM.
- **Grade:** likelihood Med · impact Low (the argument rests on the mechanism, not the CVE id) ·
  complexity Low.
- **Ask of blue:** either confirm the CVE id maps to the MEMORY.md postinstall vector in the
  primary (Cisco/Anthropic) advisory, or cite the phenomenon by the Cisco blog title and treat
  the CVE number as illustrative. Not PASS-blocking.

### GAP-L2-3 — BeliefMem exact ALFWorld figures unconfirmed
- **Location:** §6.2. Quoted: *"The one strong benchmark win for confidence-bearing memory
  (ALFWorld 59.9 vs 28.7)."* Footnote gives *"59.88 → 28.71."*
- **Finding:** arXiv 2605.05583 is real; the qualitative claim (deterministic collapse of
  probabilistic memory causes self-reinforcing error; BeliefMem wins on ALFWorld + LoCoMo) is
  corroborated. The exact digits 59.88/28.71 were not confirmed at the leaf node from the
  abstract/review sources available. Blue uses the figure carefully — explicitly scoped to
  partial observability and cited *against* adopting a confidence float for this workload — so
  the interpretive use is sound regardless of the precise digits.
- **Corroboration confidence:** exact digits MEDIUM; interpretive use HIGH.
- **Grade:** likelihood Low · impact Low (supports a simplification that has independent
  justification) · complexity trivial.
- **Ask of blue:** confirm the two figures against the paper's results table, or round-and-hedge.
  Not PASS-blocking.

---

## Low-confidence *sources* noted (not gaps — framing already honest)

- **§6.1 instruction budget** — *"frontier models reliably follow roughly 150–200 instructions,
  of which Claude Code's own system prompt consumes ~50 … degradation observable past ~80 dense
  rule-lines."* [^InstructionBudget] is a practitioner blog (tianpan.co); numbers are soft. Blue
  attributes them as "practitioner guidance," and only the *direction* (hard budget needed) is
  load-bearing. Acceptable framing; flagged for the record.
- **§6.1 Context Rot "18-model study"** [^ContextRotChroma] — vendor (Chroma) study; blue already
  states the vendor caveat inline. Model count not independently re-counted; direction-only use.

---

## Items already correctly labeled unverified by blue (no action — good practice)
- Native Auto Dream availability (§10, [^AutoDream][^DreamSkill]).
- ARC-AGI 54% regression (§10, [^AgentsDumber], secondary commentary).
These fall partly outside my slice but confirm blue is not laundering the weakest claims.

---

## Slice verdict
**PASS.** Leaf-node citation integrity for §4–§7 holds: no fabricated identifier, no
title mismatch, no source that contradicts the statement it supports. Three MEDIUM-confidence
flags (specific numbers not pinned to specific sources) — all "needs tighter attribution," none
falsified, none PASS-blocking. The two local-repo defect claims blue raises against the proposal
(§6.3) are independently true.

## Friction
None. Local filesystem and web access were sufficient for leaf-node verification of this slice.
One note: arXiv abstract pages were reachable via search snippets rather than direct fetch;
exact interior figures (BeliefMem table, MemoryPoisonSurvey attack-rate band) would have been
confirmable to the digit with a working PDF fetch — graded MEDIUM rather than HIGH in their
absence.
