# Red round-4, lens-3 (leaf-node citation verification) — slice 3 of 3

**Slice:** §10 (Unverified items) → §12 (Round-1 responses) → §13 (Round-2) → §14 (Round-3) →
**Footnotes** (the whole citation-definition surface). Full report re-read in context (lines
1–1802); CHANGELOG used only as a navigation hint.

**Lens verdict for this slice: CITATION SURFACE MATERIALLY CLEAN.** Every round-3 citation-surface
residual that falls in this slice is now closed and re-verified at the leaf node; the two
load-bearing poisoning figures re-verify HIGH live; footnote-reference integrity is intact. One NEW
low-severity metric-labeling imprecision (R4-L3-1). No blocking citation gap in slice 3. This is a
per-lens/per-slice disposition, not the holistic verdict (design gaps R3-1..R3-9 belong to other
lenses).

---

## Leaf-node re-verifications performed this round (live)

- **`[^EnvInjectedMemory]` (arXiv 2604.02623, "Poison Once, Exploit Forever")** — abstract fetched
  live. Exact match: "eTAMP achieves substantial attack success rates: **up to 32.5% on GPT-5-mini,
  23.4% on GPT-5.2, and 19.5% on GPT-OSS-120B**" and "ASR increasing **up to 8 times** when agents
  struggle with dropped clicks or garbled text." Corroboration **HIGH**. §4/§13.3 figures are exact.
- **`[^Minja]` (arXiv 2503.03704, "Memory Injection Attacks on LLM Agents via Query-Only
  Interaction")** — first abstract/PDF fetches failed to surface numbers (v1 abstract + 7.7MB
  compressed PDF = tool friction, NOT absence). The **v2 HTML abstract carries them verbatim**:
  "achieves a high average success rate of **98.2%** for injecting malicious records into the
  memory, and a high average attack success rate of **76.8%**." Table 1 per-dataset averages bracket
  these (ISR 95.6–100%, ASR 57.0–98.9%). Corroboration **HIGH**. Title/authors match. **R1-28's
  round-3 closure holds** — MINJA is genuinely traceable to its exact digits.

## Round-3 citation-surface residuals in this slice — all CLOSED at the leaf node

- **R2-9(a) — `[^MemoryDocs]` "v2.1.59+" deletion (the last-standing open citation residual from
  round 3):** VERIFIED EXECUTED. `grep "2.1.59"` shows the version string survives only in (i) the
  §1.2 body sentence that explicitly labels it "uncorroborated and is dropped" (line 145) and (ii)
  the `[^MemoryDocs]` removal-note (line 1754). It appears **nowhere as a live descriptive claim**.
  The retract-by-annotation red flagged in round 3 is now an actual deletion. **Closed.**
- **R3-13 — §1.5 "46k-star":** VERIFIED CLOSED. §1.5 (line 232–234) now reads "~87.1k-star"; §7
  (665) and `[^ClaudeMem]` (1765) agree; "46k" survives only in stale-notes. **Closed.**
- **R3-17 — `[^MemoryPoisonCve]` footnote flat vs body medium-confidence:** VERIFIED CLOSED. Line
  1784 now carries the medium-confidence / vendor-blog-only / CVE-id-illustrative tag mirroring the
  §4 body. **Closed.**
- **"80–99%" and "~90% environment-injection":** both survive only as explicit retraction/removal
  notes (lines 450, 457, 1785 for 80–99%; the surviving "90%" hits are env-injection retraction
  prose + the unrelated legitimate mem0 "~90% token / ~91% latency" figure in `[^MemZero]`). No live
  contradicted figure survives. Consistent with red's round-3 R2-8/R1-28 closures.

## Footnote-reference integrity — CLEAN

52 footnote definitions, 52 distinct labels, **zero dangling references** (no label used without a
definition) and **zero orphan definitions** (every defined label is referenced in the body). Blue's
"reference integrity re-checked (no danglers)" changelog claim verified mechanically.

---

## New graded gap

### R4-L3-1 (lens 3 R4) — IMPRECISE metric-labeling: the MINJA "success band" conflates injection-success with attack-success [severity LOW]
- **Location:** §9 risk table row 1 — *"success-if-attempted ~32.5% environment-only up to
  ~76.8–98.2% for query-driven MINJA"*; §12.5 — *"up to ~76.8–98.2% for direct query-driven
  MINJA"*; §13.3 — *"The direct query-driven MINJA variant succeeds ~76.8–98.2%."*
- **Problem:** the MINJA paper reports **two distinct metrics**: 98.2% is the **injection** success
  rate (ISR — malicious records planted into memory) and 76.8% is the **attack** success rate (ASR —
  malicious behavior actually triggered). §4 (line 456) states them correctly and separately
  ("~98.2% injection success / ~76.8% attack success"). But §9/§12.5/§13.3 collapse them into a
  single "succeeds ~76.8–98.2%" **range**, whose two endpoints are different measurements — the
  upper bound is not a higher *attack*-success observation, it is a *different quantity* (injection).
  The honest attack-success figure for MINJA is a point (~76.8% avg; 57.0–98.9% across datasets),
  not a "76.8–98.2%" band. A skeptic reading §13.3 would infer attack success reaches 98.2%, which
  the paper does not claim.
- **Required fix:** in §9/§12.5/§13.3, state MINJA as "~76.8% attack success (98.2% injection
  success)" or an ASR range "~57–99% depending on task," not a merged "76.8–98.2%" band — matching
  the correct §4 phrasing. Trivial relabel.
- **Grade:** corroboration HIGH for both numbers (leaf-node verified) · likelihood-of-misread medium
  · impact LOW (does not touch the blocking disposition, which rests on impact + CVE precedent, not
  the headline rate) · complexity trivial.

---

## Verified-clean in this slice (recorded so they are not re-raised)

- `[^EnvInjectedMemory]` (2604.02623) — 32.5/23.4/19.5% + 8× — HIGH, leaf-node re-verified live.
- `[^Minja]` (2503.03704) — 98.2% ISR / 76.8% ASR — HIGH, leaf-node re-verified live (v2 HTML).
- `[^SkillSupplyChain]`, `[^GitLockContention]`, `[^MemZero]`, `[^FaultyMemories]`,
  `[^FactsFirstClass]` (all referenced in §10/§12/§13) — carried forward from red's prior HIGH
  verified-clean list; nothing in slice-3 usage contradicts them.
- `[^SingleUserLowRisk]` — §12.5 usage matches the R2-10 relabel (blue's own synthesis, not
  external corroboration; scale-claim only). Correctly disclosed, not laundered. Not a gap.
- `[^AutoDream]`/`[^DreamSkill]` — §10 correctly labels Auto Dream unverified-as-API. Not a gap.
- `[^AgentsDumber]`/`[^FaultyMemories]` — §10 ARC "52.6% after 10 rounds" consistent with the
  footnotes and red's R1-26 closure. Not re-contested.
- CVE detail "removed user memories from system prompt" — §13.7(4) is explicitly engineered to hold
  "regardless of whether the medium-confidence CVE detail is precisely accurate," so the design no
  longer depends on the unverifiable vendor-blog claim (R1-29). Defensive framing, not a gap.

## Friction

- **PDF-table / abstract-version friction (recurring, matches R1-19's friction).** The MINJA figures
  were invisible to (i) the default arXiv `/abs/` abstract fetch (v1 abstract lacks them) and (ii)
  the 7.7MB PDF fetch (compressed tables unreadable), and only surfaced via the `/html/…v2` route. A
  reliable **PDF-table-extraction / version-pinned-HTML** capability would have closed the MINJA
  leaf-node check in one call instead of three. With it I would also have definitively confirmed the
  MINJA per-dataset ASR spread without the v2-HTML fallback. Pattern to watch: a citation "closed as
  traceable" in a prior round can rest on paper-level traceability while its *specific number* still
  sits behind table friction — traceability of the paper is not verification of the digit. Here the
  digit did check out; the friction only delayed it.
