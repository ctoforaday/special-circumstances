# Red round-4 lens-1 — leaf-node citation verification (instance 1/3, slice §0–§5)

**Surface:** full `blue/report.md` re-read in context (1802 lines); footnotes §Footnotes leaf-followed
for the slice. Slice 1 of 3 = §0 Heilmeier, §1 H1 substrate, §2 H2 consolidation, §3 native
convergence, §4 poisoning, §5 H3 cadence. Lens: follow every reference; grade corroboration per
statement.

**Verdict for this pass: FAIL (one new citation gap, R4-1, not closed/rebutted/risk-accepted).**
Severity is LOW-MEDIUM and the affected argument's qualitative direction survives — consistent with
the declining-severity convergence red has recorded, not a regression. Live leaf-node work this round
re-confirmed the R2-8/R3 closure and confirmed all six R3 citation repairs landed in the text.

---

## NEW GAP

### R4-1 (round-4 lens-1) — MISCITED figures: §2.3a's cosine-bin dedup precision numbers are not in `[^LLMJudgeDedup]`, whose actual content is a different methodology [severity LOW-MEDIUM]
- **Location:** §2.3 "(a) Candidate retrieval is unspecified." — *"LLM pairwise judgment of *given* candidate pairs is reliable at high similarity but degrades sharply near the decision boundary (at cosine ≥0.95 every flagged pair is a true duplicate; at 0.85–0.87 only ~1.5% are)[^LLMJudgeDedup] — so the binding constraint is **recall of candidate pairs**, not judgment quality."*
- **Problem:** `[^LLMJudgeDedup]` = arXiv **2604.18835**, *"Semantic Needles in Document Haystacks: Sensitivity Testing of LLM-as-a-Judge Similarity Scoring"* (Aksoy et al., PNNL, Apr 2026). Verified at the leaf node this round via **three independent routes** (abstract fetch, full-text HTML fetch, web-search summary): the paper is a multifactorial sensitivity study of LLM *scoring on a 0–100 scale* under perturbations (negation, conjunction swap, named-entity replacement), reporting **within-document positional bias** and **model-specific scoring fingerprints**. It does **not** use cosine-similarity thresholds and does **not** report true-duplicate precision by cosine bin. The cited specifics — "cosine ≥0.95 → 100% true duplicates; 0.85–0.87 → ~1.5%" — are the signature of an **embedding near-duplicate precision curve**, a different measurement from what this paper performs. A skeptic following the footnote lands on a paper that does not carry the numbers (the "laundered into fact" failure the protocol names; same class as R1-18 figure-real/source-wrong).
- **What survives:** the *qualitative* half of the sentence — "LLM pairwise judgment degrades near the decision boundary" — is actually **supported** by 2604.18835 (its whole point is LLM sensitivity degrading on subtle/near-boundary semantic differences). Only the parenthetical **cosine-bin precision figures** are unsupported by the cited source. And the §2.3a conclusion ("binding constraint is *recall*, not judgment quality; whole-bundle-in-context is adequate at this scale") does not *rest* on the exact 1.5% number — it rests on the qualitative degradation + the paraphrase-recall gap ([^ParaphraseGap]). So the argument is not destroyed; the citation surface is wrong.
- **Required fix:** either (a) re-attribute the cosine-bin figures to the embedding-dedup study that actually carries them and quote the bins, or (b) drop the parenthetical numbers and keep the qualitative "LLM judgment degrades near the boundary" claim, which `[^LLMJudgeDedup]` does support. Do not leave specific cosine-precision statistics hanging on a 0–100-scale sensitivity paper.
- **Grade:** corroboration LOW-as-cited for the cosine-bin figures (HIGH for the qualitative degradation direction) · likelihood-of-miscitation medium-high (3 fetch/search routes agree on scope mismatch; not a single lossy fetch) · impact low-medium (props a specific quantitative claim in the dedup-recall argument; the argument's conclusion survives on the qualitative leg) · complexity-to-fix trivial. **Pattern: footnote over-attribution / figure-source mismatch (specific statistic pinned to a source of different methodology).**

---

## Re-confirmed LIVE this round (recorded so they are not re-raised)

- **`[^Minja]` (arXiv 2503.03704)** — full-text HTML fetch returns **ISR 98.2% / ASR 76.8%** verbatim
  (plus per-dataset spread: eICU 98.5%/90.0%, MIMIC-III 95.6%/57.0%). Matches §4 body and the
  footnote exactly. R2-8/R1-28 MINJA leg re-confirmed at the leaf node — closure stands. HIGH.
- **`[^EnvInjectedMemory]` (2604.02623)** ≤32.5% band — carried from R3 live verification; no standing
  "~90%" survives in slice §0–§5 (grep-clean; the only "~90%" in slice is the mem0 token-reduction
  figure in §2.2, unrelated). Closure stands.

## R3 citation repairs — verified landed in the text this round (slice §0–§5)

- **R3-13** — §1.5 (line 234) now reads "**~87.1k-star**" and flags Round 0's "46k" as stale. Propagated. Closed stands.
- **R2-9(a)** — `[^MemoryDocs]` (line 1754): the "(auto memory native v2.1.59+)" parenthetical is **deleted** from the descriptive clause (no longer retract-by-annotation). The four words are gone; the footnote's descriptive clause carries no version. **R2-9(a) now genuinely CLOSED** — the standing open item from R3 is discharged.
- **R3-14** — `[^MemorySurvey]` (line 1773) claim list trimmed to summarization drift only; ~29-day half-life / "semantic intensification" / cross-version drift withdrawn; §2.1 meaning-drift re-attributed to `[^FaultyMemories]`. Landed.
- **R3-15** — `[^RecMem]` (line 1782) + §5 (lines 519-522) now read "up to ~87% / accuracy maintained or improved." Landed. (Note: §5 is in slice; the 87% upper bound is not independently re-verified this round — abstract fetches remain lossy — but the R3 live verification stands and the body no longer asserts the unsourced 77% lower bound.)
- **R3-16** — `[^InstructionBudget]` (line 1788) + §6.1 separates the 150–200 instruction budget from the <100 line budget. (§6.1 is slice-2 boundary; noted for the instance-2 auditor.)
- **R3-17** — `[^MemoryPoisonCve]` (line 1784) now carries the medium-confidence / vendor-blog-only / CVE-id-illustrative tag mirroring the §4 body. Landed.

## Unable-to-corroborate at leaf node (friction, not a hard gap)

- **`[^ParaphraseGap]` (§2.3a, Springer 10.1007/s10579-023-09715-7 + MDPI):** the Springer primary
  **redirects to an auth wall (HTTP 303 → idp.springer.com)** — unfollowable for a skeptic and
  unverifiable from here. The cited figures ("semantic beats lexical by 11–20+ points"; "99%+
  semantic similarity with single-digit BLEU overlap") are plausible and well-known in the paraphrase
  literature, so this is **unable-to-corroborate, NOT contradicted** (cf. the R1-19 / R3-14 lossy-fetch
  caveat). Followability concern: at least one of the two `[^ParaphraseGap]` sources should be an
  open-access primary (the MDPI co-source may already be) so the leaf node is reachable. LOW.
- **§2.1 / §10 ARC-AGI "52.6% after 10 rounds":** `[^AgentsDumber]` (johnsonlee.io blog) attributes
  the figure to `[^FaultyMemories]` (arXiv 2605.12978); the origin figure is **still not confirmed at
  the primary** this round. This is **honestly labeled unverified in §10** and quarantined there
  (correct handling per R1-26) — recorded as still-open-at-primary, **not a new gap**.

## Verified-clean in slice (spot-checked, no issue)

- `[^Minja]`, `[^EnvInjectedMemory]` (above); `[^MemoryPoisonSurvey]` (no longer backs any ASR figure
  — R2-9(b) closure holds; the "80–99%" lives only in the removal note); `[^FactsFirstClass]` (§2.1
  60%/252× — HIGH, carried from prior live verification); `[^MemZero]` (§2.2 mem0 ADD-only + ~90%
  token/~91% latency — HIGH, carried); `[^ZepGraphiti]`, `[^GenerativeAgents]` (§5 threshold/2–3×-day)
  — on the standing verified-clean list, not re-fetched this pass.

---

## Synopsis
1. One new gap R4-1: §2.3a's cosine-bin dedup precision figures ("≥0.95→100%, 0.85–0.87→~1.5%") are miscited to arXiv 2604.18835, an LLM-as-judge *sensitivity* paper (0–100 scale, positional bias) that carries no cosine-threshold dedup stats — 3 fetch/search routes agree on the scope mismatch; qualitative degradation claim survives, specific numbers do not. LOW-MEDIUM, trivial fix.
2. Live re-confirmation: MINJA ISR 98.2%/ASR 76.8% matches (R2-8/R1-28 closure stands); all six R3 citation repairs verified landed in slice text, and R2-9(a) (v2.1.59) is now *genuinely* deleted — discharged.
3. Verdict FAIL for this pass (R4-1 open); `[^ParaphraseGap]` paywalled (unable-to-corroborate, friction) and the §10 ARC 52.6% attribution still unconfirmed-at-primary but honestly quarantined — neither a new hard gap.
