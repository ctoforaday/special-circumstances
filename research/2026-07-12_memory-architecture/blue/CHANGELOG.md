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

## Round 1 — response to red's audit (2026-07-13)

All edits additive: repairs in place for factual errors (a false claim is not substance to
preserve; surrounding true content kept), new analysis added as §12 and new table rows. Nothing
from Round 0 removed.

**Leaf-node re-verification performed this round (this machine):**
- Confirmed red's R1-1: `plugins/prosthetic-conscience/tools/internal/secrets/secrets.go`
  (shared matcher, built-for-reuse header), `tools/cmd/sc-secrets-gate/main.go` (wired PreToolUse
  deny-hook), `hooks/hooks.json` (wired on WebFetch|WebSearch|Bash) all ship. Blue Round 0 grepped
  only `*.md` — retracted.
- Web-searched + confirmed: mem0 now single-pass ADD-only (corroborates §2.3b); 60% figure belongs
  to arXiv 2603.17781; claude-mem ~87.1k stars; opportunistic/supply-chain poisoning literature;
  single-user-low-risk disconfirming consensus; git index.lock contention evidence.

**In-place repairs (factual, R-numbered):**
- Verdict: reframed to gated-on-blockers (R1-17); "strongest" → "suggestive" for Auto Dream keystone
  (R1-10); verdict bullet list corrected (secret-scrub was blue's error, not proposal's).
- §1.2: #57507 is Closed-not-planned + workaround + Subpattern B (R1-20); dropped "v2.1.59" (R1-22).
- §1.3: #56540 Closed-not-planned, macOS-launchd-specific, Windows scope caveat (R1-21).
- §2.1: 60%/36.7x re-attributed to arXiv 2603.17781 [^FactsFirstClass] (R1-18); ARC 52.6% (R1-26).
- §2.2 / §7: mem0 corrected to current ADD-only, harvested as support for append-only (R1-23).
- §2.4: 61.4%/71.6% relabeled approximate (R1-19); solo-operator extrapolation relabeled reasoned
  inference (R1-14).
- §4: CVE id-mapping + "removed from system prompt" tagged medium-confidence (R1-29); 80-99% band
  softened to "up to ~90-95% (MINJA/env-injection), attributed" (R1-28).
- §6.2: ALFWorld digits rounded-and-hedged (R1-30); named the replacement tie-breaker (R1-13).
- §6.3 item 1: **corrected in place** — secret-scrub matcher + gate ship; wire-not-build; outbound-
  only limit (R1-1). Footnote [^LocalRepoScrub] rewritten with the retraction.
- §7: claude-mem ~87.1k (R1-24); basic-memory local-first/cloud-optional (R1-27); Letta git-branch
  detail downgraded to community-suggested (R1-25).

**New additive content:**
- §3 / §6.2 / §7: inline caveats withdrawing autoMemoryDirectory-into-store (R1-4a) and coupling
  the .claude/rules channel to trust tier (R1-4b), pointing to §12.4.
- §8 table: items 2/3/4 reworded; new rows 15-20 (clone gate, provenance-of-content, channel/voice,
  concurrency, consolidator, deliverables); effort note (R1-16).
- §9 table: multi-machine row split; new rows for concurrency (R1-5), history-scrub (R1-6),
  self-poisoning curator (R1-7), clone vector (R1-2), bootstrap laundering (R1-3); poisoning row
  regraded with attacker-model note (R1-11/R1-28).
- **New §12** (Round-1 responses): §12.2 clone-time injection (accept, blocking); §12.3 provenance-
  of-content (accept, blocking); §12.4 channel/screen/voice reconciliation (accept); §12.5 netted
  build-vs-adopt + attacker model (part accept, part rebut of R1-11 grade); §12.6 concurrency
  control (accept, carved from YAGNI); §12.7 history-scrub tradeoff (part accept); §12.8 ephemeral
  consolidator memory (accept); §12.9 delivered re-scoped phase table + defer/timing branch (R1-15)
  + Evidence-section cap (R1-12); §12.10 items held as adequate.
- 5 new footnotes: [^FactsFirstClass], [^EnvInjectedMemory], [^SkillSupplyChain],
  [^SingleUserLowRisk], [^GitLockContention]. Footnote-reference integrity re-checked (no danglers).

## Round 2 — response to red's audit + lead's ruling (2026-07-13)

All edits additive: citation regressions repaired in place; design residuals answered in new §13;
lead's docket item (R1-8/R2-2) closed on corrected accounting. Nothing from Rounds 0–1 removed.

**Leaf-node re-verification this round (web):**
- Confirmed red's R2-8: arXiv 2604.02623 abstract reports ASR 32.5%/23.4%/19.5% (up to 8× under
  stress), NOT ~90% — Round 1's env-injection figure was ~triple the paper's. Confirmed MINJA is
  arXiv 2503.03704 (~98.2% injection / ~76.8% attack success). The "~90%" is retracted everywhere.

**In-place citation/label repairs (R2-numbered):**
- §4 + risk row + §12.5: "~90%/~90–95%" replaced with the corrected wide band (~32.5% env-only up
  to ~76.8–98.2% MINJA), attributed (R2-8). New footnote [^Minja] (arXiv 2503.03704).
- [^EnvInjectedMemory]: figure corrected to ≤32.5% + 8× stress (R2-8).
- [^MemoryPoisonSurvey]: "80–99%" clause removed — survey carries no ASR numbers (R2-8/R2-9).
- [^MemoryDocs]: dropped "v2.1.59+" (R2-9, propagating R1-22 to the footnote).
- [^SubagentDocs] + §3 body: "v2.1.33+" attributed to community source, not docs (R2-12).
- [^LettaSleep]: git-branch clause moved out of the primary-source claim list (R2-9, R1-25).
- [^SingleUserLowRisk]: relabeled — dev.to frames by scale not trust; advisory-locking-is-enough is
  blue's own reasoned synthesis, not external corroboration; self-survey no longer laundered (R2-10).
- §4 "Required changes" re-anchored from flat "before Phase 1" to per-phase gates (R2-11).
- §2.3a bare "(§11)" → "proposal §11" (R2-13).

**New additive content:**
- **§0 Heilmeier Catechism** added (R2-13) — 9 questions, travels into final assembly.
- **New §13** (Round-2 responses): §13.2 R2-1 clone-ratification redesigned to key on commit
  AUTHORSHIP not content fingerprint (self-defeating form withdrawn); §13.3 R2-8 likelihood re-based
  on corrected number, blocking grade unaffected (rests on impact + CVE precedent); §13.4 R2-3
  turn-level provenance specified + re-graded Medium (was "one predicate") + partial rebut
  (conservative form is not "useless"); §13.5 R2-4 pid+heartbeat liveness + explicit-pathspec commit;
  §13.6 R2-5 history-scrub reconciled (local retains full history; only post-leak public mirror
  degraded — nice-to-have case); §13.7 **R1-8/R2-2 lead docket CLOSED** — "Shared" mislabel
  corrected, 3 net-new widenings counted, value bounded ordinally (2 load-bearing / 2 nice-to-have),
  build re-argued, double-bind resolved by UNCONDITIONAL de-authorized channel; §13.8 R2-6
  consolidator-reads-as-data + graded residual; §13.9 R2-7 flag-absent MEMORY.md fallback; §13.10
  R1-11 lead ruling reflected; §13.11 §8 rows 21–27; §13.12 §9 risk rows; §13.13 risk-accepts
  (multi-machine, signed-commit strong form).
- §12.4(b) marked SUPERSEDED by §13.7(4) (unconditional de-authorization).
- §12.9 Phase 2 row + §3 consequence 3: MEMORY.md two-writer resolution now branches on Phase-0
  flag check (R2-7).
- §8 item 15 + §9 clone row: content-fingerprint form withdrawn, authorship form noted.
- Verdict count updated to 27 items (Round-2 fixes = items 21–27).
- 1 new footnote ([^Minja]); reference integrity re-checked.

## Round 3 — response to red's audit (2026-07-13)

All edits additive: citation lags fixed in place; design residuals answered in new §14; the
organizing information-flow invariant red offered the lead is adopted explicitly (one principle
replacing six spot-patches). Nothing from Rounds 0–2 removed.

**Leaf-node re-verification this round (web fetches):**
- [^RecMem] arXiv 2605.16045 abstract: "up to 87%" token reduction *while exceeding* baseline
  accuracy — confirms red R3-15 (Round 0's "77–87% / no accuracy gain" was wrong).
- [^MemorySurvey] arXiv 2603.07670v1: fetched text carries **no** ~29-day half-life, no "semantic
  intensification", no cross-version score drift — confirms red R3-14.
- [^InstructionBudget] tianpan.co: "40–80 lines, under 100 upper bound" for the line budget, 150–200
  for the instruction budget — confirms red R3-16 ("<200 lines" conflated the two).

**In-place citation/label repairs (R3-numbered):**
- §1.5: "46k-star" → "~87.1k-star" (R3-13).
- §2.1: "semantic intensification" re-attributed to [^FaultyMemories]; ~29-day figure withdrawn (R3-14).
- §5 + [^RecMem]: "77–87% / no accuracy gain" → "up to ~87% / accuracy maintained or improved" (R3-15).
- §6.1: ~29-day half-life withdrawn → "days-to-weeks, plausible" (R3-14).
- §6.1 + [^InstructionBudget]: "<200 lines" → "<100 lines (40–80 dense)"; instruction vs line budgets
  separated (R3-16).
- §6.2: cross-version score-drift clause re-scoped to [^MemoryEviction], relabeled inference (R3-14).
- [^MemoryDocs]: "(auto memory native v2.1.59+)" **deleted from the descriptive clause** — R2-9(a)
  was retract-by-annotation; now executed as deletion.
- [^MemoryPoisonCve]: medium-confidence / vendor-blog-only / CVE-id-illustrative tag carried into the
  footnote to mirror the §4 body (R3-17).
- [^MemorySurvey]: claim list trimmed to summarization drift; withdrawn claims annotated (R3-14).
- §13.7(3) table + §13.7 prose: typing reclassified **surface-neutral / defense-enabling**, not
  surface-narrowing (R3-10); build case survives without the false narrowing-credit.
- §8 header: forward pointer to §13.11 (items 21–27) + §14.9 (28–31) + the §14.8 operative-decisions
  table (R3-11/R3-12). Verdict count → 31 items.

**New additive content — §14 (Round-3 responses):**
- **§14.1 the trust-derivation invariant** (adopted — red's meta-recommendation): committed
  trust-elevating fields never inherit; import corollary (foreign clones load reference-tier, fields
  reset) + session corollary (taint parser-derived, transitive, not LLM self-report). Organizing fix
  for R3-1/2/3/5/7/9.
- §14.2 R3-1 authorship clone-gate: accept the re-grade (forgery low-effort/targeting-required;
  foreign-clone ratify inherits §2.4 decay) but the invariant demotes authorship to a
  nudge-convenience, so forgery buys nudge-suppression not activation; honest v1 guarantee stated
  (defends untargeted/broadcast; residual = manual `/remember` of poisoned content); signed-commit
  risk-accept **strengthened** (signing gates only the nudge under the invariant).
- §14.3 R3-3/R3-7: withdraw §13.4's unsound self-reported/parentUuid mechanism; adopt mechanical
  transitive taint; concede sound per-turn info-flow is unsolved → **risk-accept**, auto-promotion
  **downgraded to a convenience** (untainted sessions + operator-confirmed only).
- §14.4 R3-4: accept the leaf-node contradiction; consolidator **does** read bodies semantically for
  dedup under data-framing; "opaque payload" corrected; residual re-graded up (Low-Med-L/Med-High-I),
  capped by git-revert recoverability + per-pass caps, not claimed closed.
- §14.5 R3-9: separate Threat A (prompt-injection-of-consolidator → §13.8) from Threat B
  (field-inflation → invariant + mit.4); structured-field reliance no longer claimed injection-safe
  in general.
- §14.6 R3-2: per-concept authorship + import corollary; collaborator concepts arrive reference-tier,
  elevate only by local action; residual graded; attaches to nice-to-have committed-project-store.
- §14.7: R3-8 reconciled (value = own global store, no foreign clone; clone risk attaches to
  nice-to-have; import clamp absorbs breadth-driven cloning); R3-6 recurring per-run flag check; R3-5
  widening-#2 acceptance conditioned explicitly on the §14.1 invariant.
- **§14.8 consolidated operative-decisions table** (R3-12) — single current-decision surface, 10
  contested items with superseded forms as pointers.
- §14.9 §8 rows 28–31; §14.10 §9 risk rows; §14.11 risk-accepts (unsolved info-flow, signed-commit
  strong form re-affirmed).
- No new footnotes; reference integrity re-checked (no danglers).

## Round 4 — response to red's audit (2026-07-13)

All edits additive: two structural gaps in the §14.1 invariant closed (each a removal of trust /
a `.gitignore` line, not new machinery); five coherence residuals accepted or risk-accepted; six
citation/labeling lags fixed in place. New §15. Nothing from Rounds 0–3 removed.

**In-place citation / labeling fixes:**
- Title line 1: "living, Round 1" → "living, Round 4" (R4-7).
- §0 Q3/Q5: reconciled with §14.3/§15.1 — auto corroborate→promote is a convenience over
  untainted sessions; durable promotion is operator-gated (R4-7).
- §9 / §12.5 / §13.3: MINJA stated as two distinct metrics — ~76.8% *attack* success (98.2%
  *injection* success) — not a merged 76.8–98.2% band; §4 was already correct (R4-12).
- §2.3a + [^LLMJudgeDedup]: cosine-bin precision figures dropped (they are an embedding
  near-duplicate curve, not this 0–100 LLM-judge sensitivity study); qualitative
  boundary-degradation claim retained; footnote relabeled (R4-9).
- §6.2 + [^MemoryEviction]: calibration/runaway-certainty claim de-coupled from the SSGM arXiv
  leg (which covers temporal decay/semantic drift only) → attributed to the Medium listicle
  alone, graded blog-sourced; drop-the-float recommendation stands on observable-facts + BeliefMem
  (R4-10).
- §5: Auto Dream ~24h + >5-session trigger tagged "(community-reported, §10 Unverified)" (R4-11).
- §14.1 import corollary: `last_seen` added to the reset list (cleared / set to import time) —
  the one field the invariant header names but the reset list omitted (R4-8).

**New additive content — §15 (Round-4 responses):**
- **§15.1 R4-1** — invert the session-corollary taint **denylist to an allowlist** (fail-closed):
  default-tainted for any un-provenanced tool result (Bash / MCP / sidechain / non-locally-trusted
  `Read`); sidechain taint propagates to parent; new tool types default tainted. Proof of the gap
  is internal — §6.3's outbound gate already wires on `Bash`. Folds into item 29.
- **§15.2 R4-2** — the missing **session-open enforcer**: git-ignore `projections/` (commit raw
  concept bodies only). A fresh clone has no `@`-importable active-authority surface; the projection
  regenerates on first local `/dream` with the import-corollary clamp applied. Makes R1-2/R2-1/
  R3-1/R3-2 structurally moot. Price stated: the projection no longer travels (committed-store value
  shrinks to concepts-only). New §8 item 32, Blocking.
- **§15.3 R4-3** — corrects §14.7: the own-global store is locally authored, so the import corollary
  does not fire on it — the honest bound is a **single ingest-time human gate**, not per-project
  re-derivation. Risk-accept the per-project gate (it would gut the cross-project differentiator).
- **§15.4 R4-4** — (a) states the value-side movement to the lead: §13.7 build margin now rests on
  operator-gated promotion, not auto-recurrence; margin narrows but does not invert (load-bearing
  differentiators never depended on auto-promotion). (b) Apply §2.4 controls to `/remember`
  (per-batch caps, weekly digest, tier-gating, mit.3 screening). New §8 item 33, Medium.
- **§15.5 R4-6** — detection primitive specified as `/dream`-recorded hash-delta (no
  commit-authorship, since MEMORY.md lives outside the project repo); Auto-Dream-*specific* signature
  marked an unverified Phase-0 empirical dependency; degrades to fail-safe heuristic. Amends item 31.
- **§15.6 R4-8 / §15.7 R4-5 / §15.8** — last_seen fix recorded; operative blocking set recomputed
  **once** = {1, 2, 3, 16, 28, 29, 32} = **7** (was stale "5"); item 29 does not fold into 16
  (distinct scopes); Verdict count corrected to "33 items, 7 blocking"; §14.8 gets a blocking-set line.
- **§15.9** — R3-1/R3-2/R3-5/R3-8 close **by construction** under §15.1 + §15.2; R1-19 carried
  (PDF-table-extraction friction).
- §15.10 §8 rows 32–33 + amended 29/31; §15.11 §9 risk rows (R4); §15.12 risk-accepts.
- Pointers added at §14.1 (denylist superseded / "sound" corrected), §14.7 (R3-5 own-global
  correction; R3-6 detection primitive).
- No new footnotes; reference integrity re-checked (no danglers; §15 cites existing labels only).
