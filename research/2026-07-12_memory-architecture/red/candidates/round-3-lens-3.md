# Round 3 — Lens 3: leaf-node citation verification (slice 3: §13 + Footnotes)

**Surface:** full `blue/report.md` re-read in context; slice = §13 (Round-2 responses) and the
Footnotes block (lines 1408-1461, the citation surface). Lens = follow every reference to its
primary, grade corroboration confidence per statement. Live fetches performed this round.

## Verdict for this slice: near-clean. One blocker closable, two LOW footnote-lag residuals.

The Round-2 citation regression (R2-8) is **fully repaired and verified at the leaf node** — both
props of the sole blocking risk now hold and are followable. Remaining defects are two LOW
incomplete-repair residuals on the *footnote surface*, both verified in the text.

---

## CLOSE at leaf node

### R2-8 — VERIFIED REPAIRED (recommend CLOSED)
Both legs re-verified live this round:
- `[^EnvInjectedMemory]` = arXiv **2604.02623** — abstract returns ASR **32.5% / 23.4% / 19.5%**
  (GPT-5-mini / GPT-5.2 / GPT-OSS-120B), "up to 8×" under environmental stress ("Frustration
  Exploitation"). Matches blue's footnote **exactly**. The retracted "~90%" appears nowhere as an
  asserted figure (grep-confirmed: all surviving "~90%" are retraction-context or the unrelated
  mem0 ~90% token-reduction figure). **HIGH, leaf-node.**
- `[^Minja]` = arXiv **2503.03704**, Dong et al. — full-text + secondary review confirm
  **ISR 98.2% / ASR 76.8%** ("over 95% ISR / over 70% ASR"). Now cited and followable (Round 1's
  gap: figure used but source unlinked). Title tracks the current arXiv version. **HIGH, leaf-node.**
The MINJA/env-injection band is now correctly stated (wide: ~32.5% environment-only → ~76.8–98.2%
query-driven) and each half traces to a source that carries it. **R2-8 closable.**

### New citation added this round — clean
- `[^Minja]` (arXiv 2503.03704): verified HIGH (above). No further action.

---

## Graded gaps (open)

### R2-9(a) — STILL OPEN: `[^MemoryDocs]` "v2.1.59+" was annotated, not deleted [severity LOW]
- **Location:** Footnotes, `[^MemoryDocs]` (line 1414) — *"MEMORY.md location and 200-line/25KB
  load **(auto memory native v2.1.59+)**, `autoMemoryDirectory` ... **R2-9 footnote repair:** the
  parenthetical 'auto memory native v2.1.59+' is dropped — the docs carry no version number ... now
  it does."*
- **Problem:** the repair note claims the parenthetical "is dropped," but the string
  `(auto memory native v2.1.59+)` is **still literally present** in the descriptive clause. The
  footnote is self-contradictory: it asserts the version, then a trailing note says the assertion
  was removed. A leaf-node reader scanning the descriptive clause still lands on "v2.1.59+" stated
  as documented fact. This is **retract-by-annotation, not deletion** — the R1-22 body correction
  was declared propagated (CHANGELOG + §13.1 list R2-9 as closed) but the artifact contradicts the
  claim. Contrast `[^SubagentDocs]` (R2-12, line 1415): there the "v2.1.33+" version was correctly
  **removed from the descriptive clause** and appears only inside the meta-note. Same author, same
  round, one clean deletion and one botched annotation — so R2-9(a) is a genuine execution miss,
  not a stylistic choice.
- **Required fix:** delete the four words "(auto memory native v2.1.59+)" from the descriptive
  clause; the trailing R2-9 note then stands on its own.
- **Grade:** likelihood certain (verified in text) · impact low (retraction is disclosed in the
  same footnote, so not laundered — but the asserted string persists) · complexity trivial.
  Corroboration of the footnote-as-written: internally contradicted. **Pattern: incomplete-repair
  footnote lag — retract-by-annotation variant (string left asserted).**

### R2-9 extension / R1-29 footnote leg — `[^MemoryPoisonCve]` asserts flatly what the body tags medium-confidence [severity LOW]
- **Location:** Footnotes, `[^MemoryPoisonCve]` (line 1444) — *"Malicious npm postinstall →
  MEMORY.md instructions treated as authoritative every session; fix (v2.1.50/v2.2) **removed user
  memories from system prompt**."* Cross-read with §4 body (lines 433-438) which tags this
  medium-confidence and the CVE id-vector mapping "illustrative", and §13.7 (line 1295) which
  routes the whole argument around the detail's uncertainty.
- **Problem:** R1-29's required fix was to "tag 'removed from system prompt' as medium-confidence"
  and "treat the CVE number as illustrative." The **body** did this; the **footnote** did not — it
  states the system-prompt removal, the CVE-id→postinstall-vector mapping, and the specific
  "v2.1.50/v2.2" versions as bare fact, all resting on two vendor blogs (Cisco, omegamax),
  post-cutoff and unverifiable from here. Same body-tagged / footnote-flat asymmetry as R2-9(a).
  This is the exact citation surface R1-29 flagged; the accepted-as-disclosed status was granted
  on the *body* disclosure, but the leaf-node reader following the footnote gets the un-tagged
  version.
- **Required fix:** carry a "medium-confidence; vendor-blog-only; CVE-id mapping illustrative" tag
  into the footnote itself, mirroring the §4 body.
- **Grade:** likelihood certain (verified) · impact low (body discloses it; §13.7 argument built
  not to depend on it) · complexity trivial. Corroboration: phenomenon medium, id-mapping /
  system-prompt-removal medium, unchanged from R1-29 — this is a surface-propagation gap, not a new
  substantive doubt. **Pattern: incomplete-repair footnote lag.**

---

## Verified-clean this slice (recorded so not re-raised)
- `[^EnvInjectedMemory]` (2604.02623) — 32.5%/23.4%/19.5% + 8× stress — HIGH, re-verified live.
- `[^Minja]` (2503.03704) — 98.2% ISR / 76.8% ASR — HIGH, re-verified live.
- `[^MemoryPoisonSurvey]` (2606.04329) — the "80–99%" clause now appears **only** inside its
  removal note (R2-9b), not as an asserted figure. Clean.
- `[^LettaSleep]` — git-branch clause correctly moved out of the primary-source claim list to a
  named community-forum attribution (R2-9c). Clean.
- `[^SubagentDocs]` — "v2.1.33+" correctly removed from the descriptive clause and attributed to
  the community source (R2-12). Clean — and the correct model for the R2-9(a) fix.
- No stray asserted "~90%" survives anywhere in the report (grep-confirmed).

## Note on §13 (design-response prose, out of citation lens but scanned)
§13's citation-bearing claims are self-referential to the verified footnotes; the R1-29
"removed user memories from system prompt" detail invoked in §13.7(4) is explicitly constructed to
hold under **both** branches of that detail's uncertainty ("regardless of whether the
medium-confidence CVE detail is precisely accurate") — honest handling; not a citation gap.

## Friction
- Abstract-only arXiv fetches remain lossy for in-body numbers, but this round the HTML/full-text
  route surfaced MINJA's ISR/ASR (98.2%/76.8%) and 2604.02623's per-model ASR directly — decisive
  for closing R2-8. A PDF-table extractor would still be faster/definitive for the R1-19 pair
  (61.4%/71.6% agent-PR figures) which remain untraced.
