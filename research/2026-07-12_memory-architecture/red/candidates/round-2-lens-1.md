# Red round-2 candidate — lens 1 (leaf-node citation verification), slice 1 of 3

**Slice:** §1 (H1 substrate) · §2 (H2 consolidation) · §3 (competitive convergence) · §4 (poisoning),
plus the footnotes those sections cite. Method: re-read full `blue/report.md` in context (1010 lines),
then followed each slice-1 reference to its primary and graded corroboration per statement↔reference pair.
Round 1 produced R1-18…R1-30 for citations; Round 2's job is to confirm blue's repairs landed at the
leaf node and to audit the footnotes added in Round 1 that now carry load.

---

## CLOSED this round (repairs verified at the leaf node — recorded so they are not re-raised)

- **R1-18 CLOSED.** §2.1 headline "~60% of facts" re-attributed to `[^FactsFirstClass]` (arXiv 2603.17781,
  *Facts as First Class Objects*). Verified against the paper abstract: "summarization destroys 60% of
  facts"; "100% accuracy across all conditions at 252x lower cost"; "100% exact-match accuracy from 10 to
  7,000 facts." All three verbatim. Blue also **dropped** the Round-0 "2,000 facts / 36.7×" specifics from
  the §2.1 body (they are not in the abstract) and now states only "~60%" + "~252×" — both corroborated.
  Corroboration HIGH.
- **R1-23 CLOSED.** §2.2/§7 mem0 corrected to current single-pass ADD-only. Verified against the mem0
  README (`[^MemZero]`, mem0ai/mem0): "Single-pass ADD-only extraction -- one LLM call, no UPDATE/DELETE.
  Memories accumulate; nothing is overwritten." Verbatim. Blue's harvest of the ADD-only pivot as direct
  production corroboration of §2.3b (append-only) is sound. Corroboration HIGH.
  - Minor, non-gap: `[^MemZero]` still cites "~90% token / ~91% latency reduction"; the README surfaces
    accuracy benchmarks (LoCoMo 92.5, LongMemEval 94.4, BEAM 64.1), not those two percentages — they
    trace to the mem0 *paper*, which the footnote also bundles. Peripheral, paper-sourced, not load-bearing.
    No action.
- **§1.1 OKF — HIGH (was already clean, re-confirmed).** Verified against the spec (`[^OkfSpec]`,
  GoogleCloudPlatform/knowledge-catalog `okf/SPEC.md`): "Version 0.1 — Draft"; `type` sole required field;
  title/description/resource/tags/timestamp recommended; index/log files carry no frontmatter; "Producers
  MAY include any additional keys. Consumers SHOULD preserve unknown keys ... SHOULD NOT reject documents
  with unrecognized fields." All verbatim. The §3.1-profile-is-spec-legal-by-construction claim holds.

## Re-verified holding (from Round 1, not re-fetched — no new evidence disturbs them)

- `[^FaultyMemories]` (arXiv 2605.12978) — inverted-U utility / corruption via drift — HIGH (lens-1 R1).
- `[^ZepGraphiti]` (arXiv 2501.13956) — invalidate-not-delete, bi-temporal windows — HIGH (lens 2/3 R1).
- `[^MemoryDocs]` native-surface mechanics — HIGH (lens-1 R1).
- `[^ConsolidationProblem]` (Hindsight) four-levers/decay — HIGH, but does **not** carry the §2.1 figure
  (that is now correctly on `[^FactsFirstClass]`).

---

## NEW / ESCALATED gap

### R2-1 (lens 1, escalates R1-28) — the R1-28 repair introduced a leaf-node CONTRADICTION: "~90% environment-injection" attack success is not what the cited paper reports [severity MEDIUM]

- **Location:** §4 — *"the nearest concrete, citable figures are **MINJA at ~95% (and ~70% under a harder
  condition)** and **~90% in the environment-injected web-agent setting** (R1-28 repair). Softened to
  **"up to ~90–95% (MINJA / environment-injection), attributed"**"* (cited `[^MemoryPoisonSurvey][^EnvInjectedMemory]`);
  footnote `[^EnvInjectedMemory]` — *"~90% attack success in the web-agent environment-injection setting"*;
  and §9 risk row 1 — *"success-if-attempted up to ~90–95%"*; §12.5 relies on the same figure.
- **Problem:** two coupled defects, both born of the R1-28 repair (which replaced an unpinnable "80–99%"
  band with two *specific* attributions):
  1. **Contradicted at the leaf node.** `[^EnvInjectedMemory]` = arXiv 2604.02623, *"Poison Once, Exploit
     Forever: Environment-Injected Memory Poisoning Attacks on Web Agents"*. Its abstract reports ASR
     **"up to 32.5% on GPT-5-mini, 23.4% on GPT-5.2, and 19.5% on GPT-OSS-120B"** — with an "up to 8×
     under environmental stress" multiplier noted but not stated to reach 90%. The "~90% in the
     environment-injected web-agent setting" is **not ~90%; it is ≤32.5% baseline** in the very paper
     cited for it. This is the "laundered into fact" failure the protocol names — a skeptic following the
     footnote lands on a paper that reports roughly one-third the claimed rate.
  2. **MINJA leg is correct-in-fact but untraceable through the cited footnotes.** MINJA's ">95% injection
     / >70% attack success" is real, but it lives in **arXiv 2503.03704** (*Memory Injection Attacks on LLM
     Agents via Query-Only Interaction*), which is **not** cited in either `[^MemoryPoisonSurvey]` or
     `[^EnvInjectedMemory]`. And `[^MemoryPoisonSurvey]` (arXiv 2606.04329) abstract explicitly "provides
     no numerical attack success rates." So the MINJA ~95%/~70% figures — the one accurate half of the
     band — are attributed to footnotes that do not carry them.
- **Interaction with R1-11:** the *disposition* still survives — the blocking core is the two ingest-edge
  gates, which do not depend on the exact success rate (blue's own §12.5 concedes this; R1-11 established
  it). So impact is bounded. But this is a fresh leaf-node contradiction in a Round-1 *repair*, not a
  residual round-0 issue — the fix regressed the citation rather than closing it. Red does not let a
  contradicted number stand just because the verdict does not rest on it.
- **Required fix:** (a) drop the "environment-injection ~90%" claim or re-attribute it to a source that
  reports ~90% (2604.02623 does not); the env-injection paper supports the *opportunistic/untargeted
  attacker model* (§12.5) at its real, lower rates — keep it for *that*, not for a 90% number. (b) Cite
  MINJA to arXiv 2503.03704 directly for the ~95%/~70% figures, or state "~95% (MINJA, arXiv 2503.03704)"
  so the number is followable. (c) Since `[^MemoryPoisonSurvey]` carries no ASR numbers, stop using it to
  back any success-rate figure — keep it for the taxonomy/"aggressive read-write = more exploitable" claim
  it does support.
- **Grade:** corroboration — MINJA figure high-in-fact / untraceable-as-cited; env-injection figure
  **contradicted at leaf node** · likelihood-of-error certain (verified) · impact LOW-MEDIUM (a
  success-rate figure the disposition does not rest on, per R1-11) · complexity-to-fix LOW (re-cite / drop).

---

## Items examined and NOT re-escalated (graded-and-disclosed; blue's handling accepted)

- **R1-29 (CVE-2026-21852 id-mapping + "removed from system prompt").** Blue tagged both medium-confidence
  in §4 and built the differential-authority argument to stand even if the detail is imprecise. That is the
  correct disclosure for a post-cutoff, vendor-blog-sourced claim I cannot verify from here. Not a fresh
  gap — remains a labeled medium-confidence item, honestly flagged. No re-escalation.
- **R1-24/R1-25/R1-26/R1-27** (claude-mem stars, Letta git-branch, ARC-AGI 52.6%, basic-memory cloud) — the
  slice-1-adjacent ones live in §7/§10 (slices 2–3); Round-1 repairs are reflected in the footnotes I read
  and appear consistent. Deferred to the slices that own those sections.

## Friction

- HTML/abstract-only arXiv fetches remain lossy for in-body numbers: I can *confirm a contradiction* when
  the abstract reports a materially different figure (as with 2604.02623: ≤32.5% vs claimed ~90%), and I
  can *confirm a match* when the abstract carries the number (2603.17781: 60%/252×), but I **cannot rule a
  figure out** when it might sit in a body table the abstract omits (e.g. whether 2606.04329 quotes MINJA
  internally). A full-PDF-text-search or PDF-table-extraction tool would let me discharge the MINJA-in-
  survey question definitively rather than grading it "untraceable-as-cited."

## Synopsis of slice-1 verdict contribution

R1-18 and R1-23 verified closed at the leaf node (FactsFirstClass 60%/252×; mem0 ADD-only — both verbatim);
§1.1 OKF and mem0 harvest HIGH. One new gap R2-1: the R1-28 repair regressed — its "~90% environment-
injection" figure is contradicted by the cited paper (≤32.5%) and the accurate MINJA ~95% figure is
attributed to footnotes that do not carry it. Slice-1 citations are otherwise sound; R2-1 does not block
(disposition survives per R1-11) but must be corrected, not left standing.
