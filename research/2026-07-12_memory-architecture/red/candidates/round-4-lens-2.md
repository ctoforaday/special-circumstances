# Red — Round 4, Lens 2 (leaf-node citation verification), slice 2 of 3

**Slice:** §3 (native-harness convergence) · §4 (memory poisoning) · §5 (H3 cadence) · §6 (H4
complexity) · §7 (H5 alternatives). Full report re-read in context; every footnote referenced in
these sections followed to the live primary where reachable.

**Slice verdict: PASS-eligible on citation grounds — no blocking or medium citation defect in
slice 2.** The load-bearing figures corroborate at the leaf node this round; the Round-3 repairs
(R3-14/15/16) landed with body↔footnote parity and did **not** regress. One LOW over-attribution
residual (R4-L2-a) is open; two hygiene notes. PASS strictly unavailable only because R4-L2-a is
newly raised and not yet closed — it is a one-line fix, not a design gap.

---

## Leaf-node re-verifications this round (recorded so they are not re-litigated)

- **[^RecMem] arXiv 2605.16045** — title *"RecMem: Recurrence-based Memory Consolidation…"* confirmed;
  abstract states **"reduces the memory construction token cost of three SOTA memory systems by up
  to 87%"** **"while exceeding their accuracy."** §5 body + footnote match. **R3-15 fix verified
  clean, no regression. Corroboration HIGH.**
- **[^EnvInjectedMemory] arXiv 2604.02623** — *"Poison Once, Exploit Forever…"*; abstract carries
  **"up to 32.5% on GPT-5-mini, 23.4% on GPT-5.2, and 19.5% on GPT-OSS-120B"** and **"ASR increasing
  up to 8 times"** under environmental stress — matches §4 verbatim. **R2-8 correction re-confirmed
  live. Corroboration HIGH.**
- **[^Minja] arXiv 2503.03704** — *"Memory Injection Attacks on LLM Agents via Query-Only
  Interaction"*; abstract carries **"high average success rate of 98.2%"** (injection) and **"high
  average attack success rate of 76.8%"** — matches §4. **R1-28 closure re-confirmed at the leaf
  node.** *Note:* first fetch (arxiv.org/abs) surfaced the title but not the numbers; second fetch
  (arxiv.org/html) surfaced both from the abstract — a small-model fetch inconsistency, **not** a
  citation defect. The band's two poles (~32.5% env-only → ~76.8–98.2% MINJA) are now both
  leaf-node-corroborated by red this round. Corroboration HIGH.
- **[^ContextRotChroma]** — 18 LLMs evaluated; **"Even a single distractor reduces performance
  relative to the baseline"**, amplifying with input length — matches §6.1. Corroboration HIGH.
- **[^MemorySurvey] trim (R3-14)** and **[^InstructionBudget] (R3-16)** — body claims and footnote
  claim-lists are consistent (MemorySurvey now backs only summarization-drift; InstructionBudget
  separates the ~150–200 instruction budget from the <100 line budget). Body↔footnote parity holds.
- **R2-9(a) (MemoryDocs "v2.1.59+")** — CLOSED: the version parenthetical is gone from the
  descriptive clause of [^MemoryDocs]; it survives only inside the repair-note explaining the
  removal. (Borderline slice-1/3 footnote; confirmed in passing.)
- **[^GenerativeAgents] arXiv 2304.03442** and **[^BeliefMemory] arXiv 2605.05583** — abstracts
  confirm the qualitative claims (importance-threshold reflection; BeliefMem outperforms ALFWorld
  baselines) but abstract-only fetches do **not** surface the precise `~150`-threshold /
  `59.88→28.71` digits; see R4-L2-c. Both figures already carry blue's own hedges (footnote-only /
  "not re-confirmed at the leaf node, rounded-and-hedged" per R1-30) — **as-disclosed, not new
  gaps.**

---

## Graded gaps (this slice)

### R4-L2-a (lens 2) — OVER-ATTRIBUTION: the confidence-calibration claim's arXiv leg does not carry it [severity LOW]
- **Location:** §6.2 — *"a stored 0.0–1.0 confidence … exhibit calibration failure / 'runaway
  certainty'. (R3-14 scope-trim … the surviving, sourced claim is the calibration/runaway-certainty
  failure mode in [^MemoryEviction].)"* Footnote [^MemoryEviction] bundles *"Agent Memory Eviction:
  8 Policies…"* (Medium, Bhagya Rana) **+** *"Governing Evolving Memory in LLM Agents (SSGM)"*,
  arXiv 2603.11768.
- **Problem:** R3-14 stripped the calibration/runaway-certainty claim off [^MemorySurvey] and
  re-homed it on [^MemoryEviction], and blue explicitly contrasts it against the now-*"inference"*-
  labelled cross-version clause — i.e. it is presented as the **sourced** survivor. But leaf-node,
  the arXiv leg (SSGM 2603.11768) discusses temporal-decay modelling and semantic drift and **does
  not carry "confidence calibration failure" or "runaway certainty."** So after the R3-14 narrowing,
  the claim rests **solely on the Medium listicle**, while drawing citation prestige from the
  co-bundled arXiv primary that does not support it — the footnote-over-attribution pattern (a
  bundle where only the non-primary leg carries the specific claim).
- **Required fix:** either drop the SSGM co-cite for *this* claim (attribute calibration/
  runaway-certainty to the Medium source alone, and grade it as blog-sourced), or relabel it
  "inference / practitioner-reported," parallel to the cross-version clause it is contrasted with.
- **Grade:** corroboration low for the calibration claim as sourced · likelihood-of-error low
  (claim is plausible, merely under-sourced) · **impact low** — the confidence-float-drop
  recommendation stands independently on the observable-facts argument (review_count/last_seen/
  status are observed; a stored float adds a model call and admitted-guess thresholds) and on the
  separately-cited BeliefMem counter-evidence; calibration is supporting colour, not load-bearing ·
  complexity trivial.

### R4-L2-b (lens 2) — HEDGE-LAG: §5 states Auto Dream's exact trigger as fact at the use-site [severity LOW / hygiene]
- **Location:** §5 — *"Native Auto Dream's ~24h + >5-sessions trigger is itself a hybrid
  clock+threshold gate.[^AutoDream]"* (also §3 line ~388 states the same numbers).
- **Problem:** the `~24h + >5-sessions` trigger is sourced only to [^AutoDream]/[^DreamSkill] —
  third-party blogs + a community skill replicating an *unreleased* feature, correctly filed under
  §10 Unverified. §3 carries the hedge ("verified as concept, unverified as a dependable API"); the
  §5 use-site presents the precise trigger as a plain fact with no inline caveat. A reader in §5
  sees a specific numeric trigger with the appearance of a verified figure.
- **Required fix:** at the §5 use-site, tag the trigger "(community-reported, §10 Unverified)" or
  drop the precise numbers, keeping the qualitative "hybrid clock+threshold" point (which the
  synthesis actually relies on).
- **Grade:** corroboration low (community/unreleased) · impact low (§5's synthesis leans on the
  *shape*, not the numbers; §3/§10 carry the hedge) · complexity trivial.

### R4-L2-c (lens 2) — CONFIRM, not re-open: precise digits remain unconfirmable at the leaf node [informational]
- **Location:** §5 [^GenerativeAgents] (`~150` reflection threshold — footnote-only) and §6.2
  [^BeliefMemory] (ALFWorld `59.88 → 28.71`).
- **Status:** abstract-only fetches this round again did not surface either precise figure (they may
  live in body tables). Both are already honestly disclosed by blue — the `~150` appears only in the
  footnote, and R1-30 already labels the ALFWorld digits "not re-confirmed at the leaf node,
  rounded-and-hedged." **R1-30 stands as-disclosed; no new gap.** Recorded only so the unconfirmed
  status is not mistaken for a fresh omission. A PDF-table-extraction capability would close both
  definitively (see friction, carried from R1-19).

---

## Not gaps (checked, clean in this slice)

- §4 [^MemoryPoisonCve] — R3-17 medium-confidence / vendor-blog-only / CVE-id-illustrative tag now
  mirrored in the footnote; body↔footnote parity. Clean.
- §4 [^OkfDeepDive] (OKF-bundle-as-injection-vector) — community blog, used as corroborating colour,
  labelled as such. Clean.
- §7 [^MemZero] (mem0 ADD-only), [^ZepGraphiti] (2501.13956), [^BasicMemory] (local-first/cloud-
  optional, R1-27) — all previously verified HIGH; no drift observed. Clean.
- §5 [^LettaSleep] git-branch clause (R1-25/R2-9c) — remains moved out of the primary-source claim
  list into a "community-suggested pattern" note; footnote consistent. Clean.
