# RED audit — Round 1, Lens 3 (dark-side / risk)

**Scope:** failure modes, likelihood × impact × complexity grading, security blindspots, tradeoff exposure.

**Methodology:** full-context re-read of blue/report.md; leaf-node verification of key claims (qmd shipping, R4 status, original disposition); spot-check original memory-architecture report for lineage.

---

## L3-F1: Partial supersession of item 6 (qmd) creates false closure on durability layer

**Location:** § H3 — Partially Superseded by FEOV/qmd Convergence (Item 6 Only), lines 20-24

**Quoted claim:** "The §8 recommendation for an SQLite/embedding index ceiling (item 6) has been **genuinely superseded** by the qmd recall layer... This solves the context retrieval problem differently — not via durability/schema but via fast searchability."

**Finding:** The report's own footnote (line 69) contradicts the "genuine supersession" claim: "qmd is searchable but does **not solve the consolidation rewrite-corruption problem** (risk matrix line 53: 'append-only rule' = the actual fix, not yet implemented)." The original recommendation (2026-07-12 report, line 60) intended a "~300–500-concept ceiling as the **trigger** for a deferred SQLite/embedding index" — the purpose was to name a consolidation durability boundary, not merely to establish a retrieval ceiling. Confusing layers: qmd solves **retrieval** (fast search); the original risk (line 53 of 2026-07-12 report: "Consolidation rewrite-corruption... High over months... High (silent knowledge loss)") requires **durability/append-only** (prevent silent data loss). Two different failure modes, two different fixes. Calling qmd a replacement is accurate for retrieval speed, **inaccurate for the durability intent** the ceiling was meant to trigger. The report buries this as a caveat rather than flagging it as an incomplete substitution.

**Risk grade:**
- **Likelihood:** HIGH — The report explicitly states the caveat in footnote but frames the body as "genuinely superseded"; readers will trust the body verdict.
- **Impact:** MEDIUM — Item 6 is not a Phase-0/1 blocker (defer is explicit), so operating without it for months doesn't introduce a new acute risk. However, consolidation rewrite-corruption is explicitly HIGH impact ("silent knowledge loss"). The lack of durable consolidation machinery is a slow-burn problem, not mitigated by qmd.
- **Complexity to mitigate:** LOW — The fix is framing: restate "qmd addresses retrieval; consolidation durability remains open" and either accelerate consolidation machinery or accept the risk of silent data loss during Phase 2+.

---

## L3-F2: Unverified R4 blocking-candidates remain unresolved and are load-bearing for security invariant

**Location:** § Status of Unverified Blocking Candidates (R4), lines 50-58

**Quoted claim:** "Both [R4-1 and R4-2] are framed as 'hardening, not redesign' and marked 'blue-fixed in §15' of the report. However, the report's own disposition is **UNVERIFIED** — red never verified them at the leaf node. They remain **open blocking gaps** pending verification and implementation."

**Finding:** The report correctly identifies that R4-1 (taint-boundary allowlist inversion) and R4-2 (git-ignore projections, commit bodies only) are unverified. However, the **severity and load-bearing status are under-emphasized**. The original memory-architecture report (2026-07-12, line 97) states the disposition: "the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the **R4-1/R4-2 structural fixes must be closed and independently verified before implementation proceeds**." These are **keystone security invariants** (original line 5: "the two final blocking-candidates... are the keystone security invariant's load-bearing pieces"). The report identifies them as open but does not emphasize that the entire memory-architecture design's security rests on two **unverified** fixes. This is not a "nice-to-have caveat"; this is the gate.

**Risk grade:**
- **Likelihood:** HIGH — R4-1 requires a parser change to invert from denylist to allowlist; the original report's 4-round history shows "blue's freshly-shipped mitigations carry un-graded next-order failures caught only on adversarial re-examination." No implementation work has shipped (unmerged branch); no red verification has happened. The risk that a new failure is lurking is empirically supported by the debate's pattern.
- **Impact:** HIGH — Memory poisoning (the risk R4-1 and R4-2 close) is "High (persistent context compromise)" per the original risk matrix (line 50). Clone-time injection via R4-2 is "High (zero-click active-authority load of attacker bytes)" (line 52). These are security invariants, not performance optimizations.
- **Complexity to mitigate:** HIGH — Verifying R4-1 requires tracing the allowlist inversion through all channels (Bash, MCP, sidechain, in-repo reads) to confirm no laundering path remains. Verifying R4-2 requires confirming that a committed `active.md` does not auto-import in any clone scenario. Both are adversarial verification tasks, not configuration checks.

---

## L3-F3: No implementation timeline or risk acceptance for Phase 2+ deferral of blocking-candidates

**Location:** § H1 — Deferred as Phase 2+, lines 12-13; § H4 — All Blockers Remain Open, lines 27-29

**Quoted claim:** "All blocking candidates and High-priority items remain unstarted, **awaiting Phase 0 FEOV/port-plan foundation work to complete**."

**Finding:** The report defers all memory-architecture work to Phase 2+ pending Phase 0/1 completion, but does not quantify the risk of that deferral. Three gaps in risk-acceptance:
1. **No timeline:** Phase 0/1 completion date is unknown; the report does not specify how long the system will operate without R4-1/R4-2 security hardening.
2. **No alternative control:** If Phase 0/1 slips, there is no compensating control (e.g., restricted use of untrusted inputs, offline-only mode, additional review gates).
3. **No explicit risk acceptance:** The report says items "remain unimplemented" and "awaiting" but does not record a risk-acceptance decision. The original disposition (2026-07-12 report, line 97) **requires** R4-1/R4-2 to be "closed and independently verified before implementation proceeds" — this is a gate, not a deferral with an expiration condition.

**Risk grade:**
- **Likelihood:** MEDIUM — Phase 0/1 is explicitly planned, so deferral is known and intentional. However, multi-phase projects often slip, especially when blocked on upstream work. The likelihood of Phase 0/1 taking >3 months is non-trivial in project planning.
- **Impact:** MEDIUM-HIGH — If Phase 0/1 takes 6+ months and memory-poisoning vectors remain unverified/unfixed, the system operates in a known-vulnerable state for an extended period. The impact is the time-window during which an attacker could exploit unpatched poisoning.
- **Complexity to mitigate:** LOW — The fix is explicit: record a risk-acceptance decision in `red/ledger.md` naming the timeline (e.g., "Phase 0/1 must complete by DATE before R4-1/R4-2 implementation") and either commit to that date or escalate if it's missed.

---

## L3-F4: Consolidation machinery is foundational, not peripheral; absence is architectural, not feature-gapped

**Location:** § H4 — All Blockers Remain Open, line 23; § Residual Caveats, line 69

**Quoted claim:** "The memory-architecture's **consolidation machinery remains unbuilt**; qmd provides the retrieval layer for what exists."

**Finding:** This is framed as a "residual caveat" but it is an **architectural gap**, not a feature gap. The original design (2026-07-12 report, Heilmeier §3) premises the entire append-only durability model on consolidation: "a nightly consolidation pass" is the core loop that re-derives tiers, deduplicates, and writes immutable claims. Without it, the system has:
- **No consolidation machinery** — no nightly loop
- **No append-only enforcement** — claims can be rewritten (risk: silent corruption)
- **No tier re-derivation** — fresh clones have no `active.md` to import (R4-2 fix won't re-derive if the derivation loop doesn't exist)

The report says (2026-07-12, line 53): "Consolidation rewrite-corruption: **High over months**... **High** (silent knowledge loss)." This is a **slow-burn failure mode** that will manifest as the memory store grows over months. Deferring it to Phase 2+ means operating without this structural control for an extended period.

**Risk grade:**
- **Likelihood:** HIGH — The failure mode is time-dependent ("over months"); it will trigger if the system accumulates many candidates without consolidation.
- **Impact:** HIGH — "Silent knowledge loss" means the system will lose ground truth without any visible failure signal.
- **Complexity to mitigate:** HIGH — Consolidation is a multi-phase architectural component (Phase 2–5 in the original plan). No workaround exists that provides the same durability guarantees.

---

## L3-F5: Methodology is shallow; verdicts (falsified/validated) are over-confident for smoke-run scope

**Location:** § Methodology, lines 74-76

**Quoted claim:** "This is a **smoke run (shallow, mechanical)**: a pipeline exercise confirming the review process, not a full audit. Research was **targeted to the five hypotheses**; disconfirming evidence was sought first (grep for shipped code contradicting the 'unimplemented' finding)."

**Finding:** The report acknowledges shallow scope but asserts high-confidence verdicts:
- H2 "Falsified" (items 2, 5, 13 exist in infrastructure but not as memory-architecture features)
- H5 "Falsified" (key items not shipped; grep for /dream, /ingest, /knowledge all zero)

Grep absence does not prove functional absence. If consolidation machinery exists under a different command name or module structure (e.g., `consolidateMemory()` vs `/dream`, or in a Go layer only), the grep would miss it. The report grepped only `*.md` and shell commands; it did not audit the implementation layer. A "falsified" verdict should rest on deeper evidence when the cost of being wrong is high. For a smoke run, softer language ("no evidence of shipping in primary channels" vs. "falsified") is more calibrated to the methodology.

**Risk grade:**
- **Likelihood:** MEDIUM — The logic is sound for the grep scope, but the scope is limited. The likelihood of a false-negative is non-negligible if implementation uses different naming.
- **Impact:** LOW-MEDIUM — If H2/H5 are wrong, they just confirm what the report already asserts (nothing shipped). The impact is loss of confidence in the verdict, not a missed security gate.
- **Complexity to mitigate:** MEDIUM — Verify by reading code/architecture, not just grep. A targeted search in the Go layer (if it exists) would strengthen the "falsified" claim.

---

## L3-F6: Unmerged branch with all design work is at risk of divergence without a merge timeline

**Location:** § H4 — All Blockers Remain Open, line 29

**Quoted claim:** "The plan file `plans/memory-architecture.md` exists only in the `plans/memory-architecture` branch (commit 32f13b2), exactly one commit ahead of main (de8d9c2). No implementation work has shipped to the primary codebase since the July 12 report."

**Finding:** An unmerged branch carrying all the memory-architecture design work is at risk of becoming stale. If main evolves (especially in the memory layer, native Auto Dream rollout, or hooks infrastructure), the branch will diverge and become harder to integrate. The report does not mention:
1. When the branch will be merged (if at all)
2. Whether the branch is being kept in sync with main
3. The risk of the branch becoming obsolete if Phase 0/1 work on main changes the target architecture

This is a long-term integration risk, especially for a design that claims to be "narrowed to the differentiating sliver" (original report, line 38).

**Risk grade:**
- **Likelihood:** MEDIUM — Long-lived branches in multi-phase projects often diverge. Native Auto Dream rollout or qmd adoption changes to MCP infrastructure could affect branch integration.
- **Impact:** MEDIUM — If the branch diverges, rework is needed to re-integrate; design decisions may need to be revisited.
- **Complexity to mitigate:** LOW — Merge the branch or commit to a merge timeline within Phase 0. If merge is deferred, establish a "keep-in-sync" policy (rebase against main on Phase-boundary commits).

---

## L3-F7: Smoke-run caveat footnoted but not carried into confidence grading

**Location:** § Residual Caveats, line 69; § Methodology, line 75

**Quoted claim:** "This run does not **audit the quality of the qmd adoption** relative to the item-6 recommendation, only its existence."

**Finding:** The report acknowledges it did not verify whether qmd actually solves the consolidation-complexity runaway problem or carries new risks (e.g., MCP server failure modes, refresh latency for embeddings, scale limits). However, this caveat is isolated to a footnote and does not lower the confidence in the H3 "Partially Validated" verdict. A "partially validated" verdict should note the unexplored quality risks. If qmd has a failure mode (e.g., MCP server crashes, losing the search index), the report's "validated" claim becomes misleading.

**Risk grade:**
- **Likelihood:** MEDIUM — qmd is new infrastructure (shipped 2026-07-14); production failure modes may not yet be known.
- **Impact:** MEDIUM — If qmd fails silently, the retrieval layer is lost but the system continues to operate; this is not a security gate but a usability/observability risk.
- **Complexity to mitigate:** MEDIUM — Audit qmd's failure modes and add telemetry/recovery logic if necessary.

---

## L3-F8: Report structure buries lineage of R4 fixes to the original disposition requirement

**Location:** § Status of Unverified Blocking Candidates (R4), lines 50-58; cross-reference original 2026-07-12 report line 97

**Finding:** The report correctly identifies R4-1 and R4-2 as unverified but does not explicitly tie them back to the original disposition requirement: "the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the **R4-1/R4-2 structural fixes must be closed and independently verified before implementation proceeds**" (original line 97). This is not a "nice-to-have caveat" that can be deferred; it is a **gate** that blocks implementation. The smoke report's framing ("remain open blocking gaps pending verification") is correct but could emphasize: "**This blocks implementation per the original disposition (2026-07-12 report, line 97).**"

**Risk grade:**
- **Likelihood:** HIGH — The original report is explicit; the risk is that the gate requirement is overlooked during Phase 0/1 planning.
- **Impact:** HIGH — If Phase 0/1 proceeds without closing R4-1/R4-2, the project has shipped unverified security-critical fixes to production.
- **Complexity to mitigate:** LOW — Escalate R4-1/R4-2 to a Phase-0 blocker or explicitly document that they are Phase 2/3 gates.

---

## Summary

**Open gaps (graded):**

| Gap | L | I | C | Status |
|---|---|---|---|---|
| L3-F1: Partial qmd supersession (retrieval ≠ durability) | HIGH | MEDIUM | LOW | False closure; reframe required |
| L3-F2: Unverified R4 security invariants (allowlist inversion, git-ignore projections) | HIGH | HIGH | HIGH | Load-bearing; gate blocks implementation |
| L3-F3: Phase 2+ deferral lacks timeline and risk acceptance | MEDIUM | MEDIUM-HIGH | LOW | Gate requirement not carried forward |
| L3-F4: Consolidation machinery absent (foundational, not peripheral) | HIGH | HIGH | HIGH | Slow-burn silent-data-loss risk over months |
| L3-F5: Smoke-run verdicts over-confident (falsified based on grep absence) | MEDIUM | LOW-MEDIUM | MEDIUM | Methodology not calibrated to confidence |
| L3-F6: Unmerged branch at risk of divergence without timeline | MEDIUM | MEDIUM | LOW | Integration risk; rebase or merge policy needed |
| L3-F7: qmd quality not audited; failure modes unexplored | MEDIUM | MEDIUM | MEDIUM | Caveat footnoted but not carried into verdict |
| L3-F8: Original disposition gate (R4-1/R4-2 block implementation) not emphasized | HIGH | HIGH | LOW | Lineage risk; gate requirement unclear |

**Verdict:** FAIL pending gaps L3-F1 (reframe), L3-F2 (critical blocker), L3-F3 (timeline + acceptance), L3-F4 (timeline + risk acceptance), L3-F8 (escalate to Phase 0 gate).

---

## Friction

None. The protocol and template fit the work cleanly.
