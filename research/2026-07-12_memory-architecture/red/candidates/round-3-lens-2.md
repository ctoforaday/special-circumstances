# Red — Round 3, Lens 2 (leaf-node citation verification), slice 2 of 3

**Slice:** §5 (H3 Cadence) · §6 (H4 Complexity) · §7 (H5 Alternatives) · §8 (Changes required) ·
§9 (Risk grading) · §10 (Unverified items). Full `report.md` re-read in context (1462 lines);
CHANGELOG treated as navigation hint only.

**Method:** followed every external citation reachable from these sections to its primary and
graded corroboration. Re-verified the load-bearing, not-yet-closed footnotes at the leaf node this
round (Chroma context-rot, InstructionBudget, RecMem, BeliefMemory, MemorySurvey). Cross-checked
that Round-2 footnote repairs (R2-9) actually landed on the citation surface for the footnotes my
sections touch.

**Lens verdict: FAIL (CHANGES-REQUIRED) for this slice.** No new *blocking* defect, but one MEDIUM
corroboration gap (MemorySurvey over-attribution — the sole prop for §6.1's "decay guesses are in
the evidenced band") and two LOW citation imprecisions, plus confirmation that R1-30 stays open and
an incidental R2-9(a) incomplete-repair spotted on the citation surface. The standing cumulative
blockers (R2-1, R2-3, R2-8 disposition, R1-8/R2-2 + R1-11 docket) are outside this slice and
unchanged.

---

## Verified-clean this round (leaf-node, recorded so they are not re-raised)

- **`[^ContextRotChroma]`** (Chroma "Context Rot", `trychroma.com/research/context-rot`) — **HIGH,
  re-verified live.** Primary confirms verbatim: "18 LLMs" tested; "model performance degrades as
  input length increases"; "even a single distractor reduces performance relative to the baseline
  … adding four distractors compounds this degradation." §6.1's "18-model study … even single
  distractors hurt" is corroborated exactly. This is the keystone that lets §6.1 claim the
  unbounded pile is a *measured* regression (verifies, not inherits) — the claim stands.
- **`[^InstructionBudget]`** (tianpan.co) — **HIGH for the instruction-count claim.** Primary
  confirms "reliably follow somewhere between 150 and 200 instructions"; "system prompt already
  contains roughly 50 built-in instructions"; degradation framing at 40–80 lines. §6.1's
  "150–200 instructions … ~50 consumed by system prompt … degradation past ~80 dense rule-lines"
  is corroborated. (One sub-claim not supported — see R3-L2-3.)
- **`[^BeliefMemory]`** (arXiv 2605.05583) — **HIGH for the interpretive use.** Abstract confirms
  the self-reinforcing-error mechanism verbatim: "By committing to one conclusion and discarding
  uncertainty, these methods introduce self-reinforcing error." §6.2's scoping to partial
  observability is sound. (Exact digits still unconfirmed — R1-30 residual, below.)
- **`[^RecMem]`** (arXiv 2605.16045) — **HIGH for the headline.** Abstract confirms "reduces the
  memory construction token cost of three SOTA memory systems by up to 87% while exceeding their
  accuracy." Upper bound and direction match §5. (Lower bound / framing — R3-L2-2.)
- **§10 `[^AgentsDumber]`** — footnote and §10 body both now read "52.6% after 10 consolidation
  rounds" (R1-26 landed); correctly quarantined as commentary relaying `[^FaultyMemories]`. Clean.
- **§10 `[^AutoDream]` / `[^DreamSkill]`** — correctly on the Unverified list, not laundered. Clean.
- **§7 `[^LettaSleep]`** — R2-9(c) now **landed**: the footnote moves the "isolated git-branch
  commits" clause out of the primary-source claim list and labels it a community-suggested pattern;
  §5/§7 body reads "community-suggested." R1-25 footnote leg discharged.
- **§7 `[^ClaudeMem]` (~87.1k, R1-24), `[^BasicMemory]` (cloud-optional, R1-27), `[^MemZero]`
  (ADD-only, R1-23), `[^ZepGraphiti]`** — closed in prior rounds; re-read consistent in §7 body.
- **§8 / §9 tables** — internal cross-references only; no external leaf nodes. §9's poisoning row
  ("~32.5% environment-only up to ~76.8–98.2% MINJA") is internally consistent with the R2-8
  corrected figures. No new external claim to follow.

---

## New / updated graded gaps (this slice)

### R3-L2-1 — MEMORYSURVEY OVER-ATTRIBUTION: three of the four things the footnote claims the paper carries are not corroborated at the leaf node — including the ~29-day half-life that is §6.1's *sole* support for "decay guesses are in the evidenced band" [severity MEDIUM]
- **Location:** §6.1 — *"an empirically tuned importance half-life of ~29 days brackets the
  proposal's 14-day short-term / 60-day candidate windows — the guesses are in the evidenced
  band.[^MemoryEviction][^ConsolidationProblem][^MemorySurvey]"*; footnote `[^MemorySurvey]`
  (arXiv 2603.07670) — *"Summarization drift and semantic intensification; importance-score drift
  across model versions; ~29-day empirical half-life."*
- **Problem:** the footnote asserts the survey carries four specifics. A leaf-node fetch of the
  HTML (`2603.07670v1`) confirms only **summarization drift** (plus a generic mention that
  MemoryBank uses Ebbinghaus-curve decay). It could **not** find (a) the **~29-day** half-life
  figure, (b) **semantic intensification** (the "likes mild → loves very spicy" example — that
  example is used in §2.1, outside this slice, also cited here), or (c) **importance-score drift
  across model versions**. The 29-day figure is the load-bearing one: it is attributed *only* to
  `[^MemorySurvey]` (the co-cited `[^MemoryEviction]` and `[^ConsolidationProblem]` footnotes make
  no 29-day claim), and it is the entire evidentiary basis for §6.1's conclusion that the
  proposal's 14/60-day windows are "in the evidenced band" — i.e. the argument that the decay
  machinery is not guesswork. If the figure cannot be pinned, that sub-claim is asserted, not
  evidenced.
- **Caveat (standing friction):** HTML/abstract fetches are lossy for in-body numbers; a survey is
  long and the 29-day figure may sit in a body paragraph the fetch missed. This is
  **"unable-to-corroborate-at-leaf-node," not "contradicted"** (cf. R1-19). But a specific
  quantitative figure used to validate a design parameter must be pinnable.
- **Required fix:** pin the ~29-day half-life to a specific source + section (name whether it is
  MemorySurvey, MemoryEviction, or elsewhere) and quote it; or soften §6.1 to "practitioner decay
  windows are days-to-weeks; the proposal's 14/60-day windows are plausible" without the false
  precision. Trim `[^MemorySurvey]`'s claim list to what the paper demonstrably carries
  (summarization drift), or move semantic-intensification / cross-version-score-drift to a source
  that carries them.
- **Grade:** corroboration LOW-as-cited for the three unconfirmed specifics (HIGH for the
  summarization-drift claim) · likelihood-of-miscitation medium · impact medium (sole prop for the
  "decay is evidenced" sub-argument in the "machinery earns its keep" section) · complexity-to-fix
  low. **Pattern: footnote asserts specifics not surfaced at the primary (leaf-node over-attribution).**

### R3-L2-2 — RECMEM: the "77–87%" lower bound and "no accuracy gain from eagerness" framing are not in the abstract; the abstract states a *stronger* result (accuracy exceeded) [severity LOW]
- **Location:** §5 — *"RecMem shows eager consolidation (LLM-processing every incoming item) wastes
  77–87% of construction tokens versus recurrence-triggered consolidation, with no accuracy gain
  from eagerness.[^RecMem]"*
- **Problem:** the abstract (arXiv 2605.16045) reports cost reduced **"by up to 87% while exceeding
  their accuracy."** Two mismatches: (a) the **77%** lower bound is not in the abstract — it is a
  specific figure that would live in a per-system body table (lossy-fetch: unconfirmed, not
  contradicted); (b) "**no accuracy gain from eagerness**" *understates* the paper — the paper
  claims recurrence-triggered *exceeds* eager accuracy, a stronger result. The understatement is
  harmless to blue's argument (it is conservative), but the "77–87%" band presents an unconfirmed
  lower bound as if pinned — the same range-fabrication shape flagged elsewhere, here benign.
- **Required fix:** state "up to ~87% token reduction (arXiv 2605.16045), with accuracy maintained
  or improved"; drop the unconfirmed 77% lower bound or pin it to the body table.
- **Grade:** corroboration HIGH for the upper bound / LOW for the lower bound · impact low (§5
  cadence recommendation, not verdict-bearing) · complexity-to-fix trivial.

### R3-L2-3 — INSTRUCTIONBUDGET: the "<200 lines per always-loaded file" figure is not in the cited primary (which says <100 / 40–80); likely a conflation of the confirmed "150–200 *instructions*" count [severity LOW]
- **Location:** §6.1 — *"practitioner guidance converges on <200 lines per always-loaded file, with
  degradation observable past ~80 dense rule-lines.[^InstructionBudget]"*
- **Problem:** the tianpan primary says a well-curated CLAUDE.md "should fit in 40–80 lines" and
  "under 100 is a reasonable upper bound" — **not** "<200 lines." The confirmed figure is
  "150–200 *instructions*" (a count of rules, not a line count). "<200 *lines*" appears to transpose
  the instruction-count into a line-count; it is 2× the primary's stated line ceiling. The
  co-bundled "MindStudio context-rot analysis" is not independently followable, so the 200-line
  figure has no confirmable source. The "~80 dense rule-lines" half **is** supported (40–80 band).
- **Required fix:** align the line figure with the primary ("<100 lines, 40–80 well-curated"), and
  keep the "150–200 instructions" count separate from any line count; or pin "<200 lines" to a
  followable source.
- **Grade:** corroboration LOW for "<200 lines" / HIGH for "150–200 instructions" and "~80 lines"
  · impact low · complexity-to-fix trivial.

### R1-30 (residual, confirmed not regressed) — BeliefMem ALFWorld 59.88/28.71 still unconfirmed at leaf node [severity LOW, carried]
- **Location:** §6.2 — *"The one strong benchmark win for confidence-bearing memory (ALFWorld 59.9
  vs 28.7 …)."*; footnote *"ALFWorld 59.88 → 28.71."*
- **Status:** the abstract confirms the qualitative claim (probabilistic-memory collapse causes
  self-reinforcing error; BeliefMem wins on ALFWorld) but does **not** surface the exact digits.
  Blue already rounds-and-hedges in body ("59.9 vs 28.7") per R1-30, so no regression. The exact
  59.88/28.71 pair remains leaf-node-unconfirmed (lossy fetch; likely a body results table). No
  action beyond the existing hedge; recorded so it is not re-raised as new.
- **Grade:** unchanged from R1-30 (corroboration medium for digits / high for interpretive use).

---

## Cross-slice observation (anchor owned by slice 1; surfaced for the lead)

- **`[^MemoryDocs]` R2-9(a) repair is INCOMPLETE on the citation surface.** The footnote still
  physically contains the parenthetical *"(auto memory native v2.1.59+)"* inline, then appends a
  note: *"the parenthetical 'auto memory native v2.1.59+' is dropped."* The retracted number was
  **annotated, not removed** — a leaf-node reader scanning the footnote still lands on "v2.1.59+."
  §13.1 records R2-9 as "three footnote lags propagated"; for leg (a) the propagation added a
  note but left the offending text in place. This footnote is referenced from §1.2/§1.3/§3
  (slice 1's leaf nodes), so the anchor belongs to instance 1; flagged here because this lens
  caught it while auditing the footnote block. Fix: delete the parenthetical outright (the "keep
  everything additively" discipline does not apply to a *retracted citation figure* — that is the
  exact R2-9 failure mode). Downgrade vs original R2-9(a): now LOW (annotated) rather than MEDIUM
  (silent), but **not closed**.

---

## Nothing-to-flag (followed, clean, no gap)

- §5 `[^GenerativeAgents]` (reflection threshold ~150; ~2–3×/day; recency/importance/relevance
  retrieval) — closed HIGH in prior rounds; §5 body consistent. Not re-litigated.
- §5 `[^HeadlessGuide]`, §6 `[^FilesWin]`/`[^VectorOverkill]`, §7 `[^ZepCritique]` — used
  qualitatively/directionally; framing matches the sources' known positions; no specific figure
  to contradict. Clean.
- §8 effort note and §9 dispositions — internal; consistent with the R2-8-corrected poisoning
  figures and the lead's R1-11 ruling as recorded in §13.10.

---

## Friction (this lens)

- The MEDIUM gap (R3-L2-1) and the LOW ones (R3-L2-2, R1-30) all hinge on figures that *may* sit in
  a body table the HTML/abstract fetch omits — the standing lossy-fetch friction. A full-PDF-text /
  PDF-table-extraction tool would discharge R3-L2-1 (~29-day half-life), R3-L2-2 (RecMem 77%),
  and R1-30 (BeliefMem digits) definitively — currently they are "unable-to-corroborate," which
  under the stickler rule is a graded low-confidence gap, not a pass.
