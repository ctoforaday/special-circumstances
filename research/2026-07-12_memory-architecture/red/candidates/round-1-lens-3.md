# Red audit — Round 1, Lens 3: leaf-node citation verification (slice 3 of 3)

**Slice:** §7 (H5 Alternatives), §8 (changes table), §9 (risk grading), §10 (unverified items),
and the footnotes load-bearing in those sections
([^ClaudeMem], [^BasicMemory], [^MemZero], [^LettaSleep], [^ZepGraphiti], [^AgentsDumber]).
**Method:** followed each citation to its primary source and confirmed corroboration at the leaf
node. Access 2026-07-13 (report access 2026-07-12).

**Lens verdict: CONCERNS — do not pass.** One medium substantive citation drift (mem0), three
low-grade numeric/attribution flags, one informational precision note. Zep citation is fully
clean. None blocking; GAP-2 stays open until blue rebuts or the lead adjudicates.

---

## Graded gaps

### GAP-3.1 — mem0 pipeline description is stale; vendor has moved to ADD-only (MEDIUM)
- **Location:** §7 "**mem0 / Letta / Zep**" — *"**Steal**: mem0's retrieve-then-classify dedup
  pipeline (§2.3a)"*; and §2.2 *"mem0's pipeline embeds each candidate fact, vector-retrieves the
  top-K similar existing memories, then has an LLM classify ADD/UPDATE/DELETE/NOOP against that
  neighborhood."*
- **Leaf check:** the current mem0 README (cited primary source [^MemZero], `mem0ai/mem0`, the
  April 2026 algorithm) states: *"Single-pass ADD-only extraction — one LLM call, no
  UPDATE/DELETE. Memories accumulate; nothing is overwritten,"* with multi-signal
  (semantic + BM25 + entity) retrieval. The ADD/UPDATE/DELETE/NOOP classify pipeline the report
  describes matches the original mem0 *paper*, not the current shipping design of the repo the
  footnote points at.
- **Corroboration confidence:** MEDIUM. Accurate to the mem0 paper; contradicted by the current
  primary source ([^MemZero] cites the GitHub repo, which now describes ADD-only).
- **Impact:** MEDIUM. §7 recommends *stealing* a pipeline the vendor itself abandoned. Worse, the
  omission cuts against blue's own argument: mem0's pivot to ADD-only accumulation is direct
  vendor corroboration of §2.3b's "expansion appends — it never rewrites the claim." Blue leaves
  its strongest external witness on the table while citing the superseded version.
- **Likelihood × complexity:** the drift is real (high); fix is low-cost prose (update the
  description; cite mem0's ADD-only pivot as support for §2.3b).
- **Disposition:** FIX. Either update to mem0's current ADD-only design (and harvest it for
  §2.3b), or explicitly frame the retrieve-then-classify description as "mem0 v1 / the paper,"
  noting the vendor has since moved away from it.

### GAP-3.2 — claude-mem star count does not match the source (LOW)
- **Location:** §7 *"**claude-mem** (46k stars) is the strongest adopt-instead candidate"*
  (also §1.5 "46k-star").
- **Leaf check:** the cited repo (`thedotmack/claude-mem`, [^ClaudeMem]) shows **87.1k stars** on
  access. The "46k" figure matches neither current state nor plausible one-day growth — it was
  stale or wrong at drafting.
- **Corroboration confidence:** LOW for the "46k" figure specifically. HIGH for every other
  claude-mem attribute (SQLite storage, hook-based capture, AI compression, `<private>` tag
  exclusion — all confirmed on the source).
- **Impact:** LOW. The number is decorative; the substantive claim (popular, ecosystem-scale
  plugin) holds either way.
- **Disposition:** FIX cheap — correct the figure or drop the precise count ("a widely-adopted
  Claude Code plugin"). Star counts are volatile; pin with access date or omit.

### GAP-3.3 — Letta "isolated git-branch commits" not supported by the cited blog (LOW–MED)
- **Location:** §7 *"Letta's sleep-time framing and isolated-branch commits (§5)"*; §5
  *"one implementation commits reflections to an isolated git branch to avoid contention."*
- **Leaf check:** the primary Letta sleep-time blog ([^LettaSleep], letta.com) contains **no
  mention of git, branches, or any version-control contention mechanism**. It corroborates the
  sleep-time concept (background agent manages memory asynchronously while primary is idle) but
  not the git-branch detail. The footnote's compound source list ends in "community
  best-practices forum" — the load-bearing mechanistic detail traces only to that unnamed forum,
  which a skeptic cannot follow.
- **Corroboration confidence:** HIGH for the sleep-time concept; LOW for "isolated git branch to
  avoid contention."
- **Impact:** LOW–MED. It is cited as a concrete thing to "steal" and as §5 precedent; the
  specific mechanism should be traceable if it is going to seed a design decision.
- **Disposition:** FIX — name the forum/source for the git-branch pattern, downgrade it to
  "a community-suggested pattern," or drop the git-branch specificity and keep only the verified
  sleep-time framing.

### GAP-3.4 — ARC-AGI figure misquoted (LOW; already labeled unverified)
- **Location:** §10 *"The ARC-AGI 54% regression figure — secondary commentary only."*; §2.1
  *"a frontier model failing 54% of ARC-AGI problems it had previously solved once consolidated
  memory was attached."*
- **Leaf check:** the cited source ([^AgentsDumber], johnsonlee.io) states accuracy dropped to
  **52.6% after 10 rounds** (a ~47.4-point fall from 100%). "Failing 54%" matches neither the
  fail rate (47.4%) nor the solved rate (52.6%). Separately, the blog attributes the figure to
  the *Useful Memories Become Faulty* paper ([^FaultyMemories], arXiv 2605.12978) — a primary
  source the report already cites — so "secondary commentary only" undersells its provenance.
- **Corroboration confidence:** LOW for the exact "54%" number.
- **Impact:** LOW — the item is explicitly quarantined in §10 as unverified, which is correct
  handling. The blast radius is a mildly-wrong number in a caveated list.
- **Disposition:** FIX cheap — quote the source's actual "52.6% after 10 rounds," and note the
  figure originates in [^FaultyMemories] rather than being pure secondary commentary.

### GAP-3.5 — basic-memory "no server/cloud" is imprecise (LOW / informational)
- **Location:** §7 *"basic-memory ... (markdown source of truth + derived SQLite index + MCP,
  no server/cloud)"*.
- **Leaf check:** the source ([^BasicMemory]) confirms local mode is serverless
  ("No servers required"), but an **optional paid cloud** ($15/mo, cross-device sync, Postgres
  alternative) exists. "No server/cloud" as an absolute is slightly off.
- **Corroboration confidence:** HIGH for the substantive point (can run fully local; files +
  SQLite + MCP; files are source of truth). Imprecise only on the absolute "no cloud."
- **Impact:** LOW. Does not change the §7 conclusion (basic-memory complements, adds MCP-server
  dependency).
- **Disposition:** informational — tighten to "local-first; cloud optional."

---

## Verified clean (positive corroboration recorded)

- **[^ZepGraphiti] — HIGH.** §7/§2.2 claim (LLM contradiction detection against semantically
  related edges; invalidate-not-delete via validity windows) is confirmed verbatim in arXiv
  2501.13956 §2.2.3 (Temporal Extraction and Edge Invalidation), including the bi-temporal
  `t_invalid`/`t_valid` mechanism. No gap.
- **[^ClaudeMem] feature claims — HIGH.** SQLite storage, five lifecycle hooks, AI compression,
  `<private>` tag exclusion all confirmed on source (only the star count is off; see GAP-3.2).
- **[^BasicMemory] architecture — HIGH** (modulo the cloud precision note, GAP-3.5).
- **§10 internal-artifact labeling** (internal FUSE prior art, OpenClaw dream-diary anecdote,
  AgentOrange `continuous_learning` "battle-tested") — **correctly labeled unverified, not
  laundered.** This is proper handling per the protocol, not a gap. Recorded so it is not
  re-raised.

## Internal-consistency spot check (§8/§9 tables)
- Issue numbers consistent across body and tables: headless fan-out hang **#56540**
  ([^HeadlessHang], §1.3, §8 row 9, §9); agent-memory allowlist bug **#57507**
  ([^SubagentMemoryBug], §1.2, §8 row 2). No cross-reference drift found in this slice.
- §8/§9 rows recapitulate earlier-section citations; no new leaf claims introduced that were not
  verified in their home sections.

## Friction
None for this task — all slice-3 sources were web-fetchable and the local report was fully
readable. Note: three findings (GAP-3.1 mem0, GAP-3.2 star count) are *drift* discoveries
enabled only because live sources moved since the report's access date; a citation-verification
lens run against archived snapshots would have missed them. Recommend the protocol record
access-date deltas explicitly.
