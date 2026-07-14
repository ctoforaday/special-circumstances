# Red — Round 3, lens 1 (leaf-node citation verification), instance 1 of 3

**Slice:** §0 Heilmeier, §1 (H1 substrate, incl. §1.5), §2 (H2 consolidation), §3 (competitive
landscape), §4 (memory poisoning) — plus the footnotes those sections cite. Method: re-read the
full living `blue/report.md` (1462 lines) in context; followed each citation in-slice to its
source; graded corroboration per statement; re-checked whether the Round-2 repairs landed at the
citation surface (not just in the body prose).

**Lens verdict for this slice: FAIL — two un-propagated repairs.** The substance of §1–§4 verifies
clean at the leaf node, including the freshly-added MINJA citation. But **two previously-closed
repairs did not actually execute at the leaf node**: one footnote still literally carries the
figure its own R2-9 repair note says was dropped, and one body section still carries a figure R1-24
corrected everywhere else. Both are the *repair-note-landed-but-edit-didn't* failure — the reason
this lens re-follows citations rather than trusting the change log.

---

## New / re-opened gaps

### R3-1 (reopens R2-9a) — `[^MemoryDocs]` still carries "(auto memory native v2.1.59+)" as a standing claim; the R2-9 repair note announces a drop that never executed [severity LOW-MEDIUM]
- **Location:** footnote `[^MemoryDocs]` (line 1414) — *"MEMORY.md location and 200-line/25KB load **(auto memory native v2.1.59+)**, `autoMemoryDirectory` …"* immediately followed by *"**R2-9 footnote repair:** the parenthetical \"auto memory native v2.1.59+\" is dropped — the docs carry no version number."*
- **Problem:** the footnote is self-contradictory. `grep -o` confirms **two** occurrences of `v2.1.59+` on line 1414: (1) the standing descriptive claim, never removed; (2) inside the repair note that claims (1) was dropped. R2-9(a) was recorded as the fix for exactly this footnote lag, but the *edit* was never made — only the *note*. A leaf-node reader following `[^MemoryDocs]` still lands on "(auto memory native v2.1.59+)", the uncorroborated version R1-22 retracted. The body (§1.2 line 144, §3 line 375) correctly reads "version unspecified"; the footnote remains the one surface that contradicts them. This is R2-9(a) NOT closed — worse than a plain lag, because the footnote now asserts and retracts the same string in one breath.
- **Required fix:** delete the parenthetical "(auto memory native v2.1.59+)" from the descriptive text of `[^MemoryDocs]`. Leave the repair note or drop it — but the standing string must go.
- **Grade:** likelihood certain (verified in the text) · impact low-medium (citation surface contradicts itself and the repaired body) · complexity-to-fix trivial (one deletion). Corroboration of the footnote as written: contradicted-by-itself. **Pattern: repair-note-without-edit (a recorded repair's note landed, its edit did not).**

### R3-2 (reopens R1-24) — §1.5 still reads "claude-mem (46k-star …"; R1-24's ~87.1k correction never propagated here [severity LOW]
- **Location:** §1.5 (line 230) — *"claude-mem (**46k-star** Claude Code plugin: hook-based session capture → AI compression → local SQLite + full-text search)."*
- **Problem:** R1-24 (recorded CLOSED) corrected the stale "46k" to "~87.1k" in §7 (line 643) and in `[^ClaudeMem]` (line 1425) — both now say ~87.1k and explicitly flag "46k" as stale. §1.5 was missed: `grep -o "46k"` shows line 230 carries a *standing* "46k-star" claim, contradicting §7 and the footnote within the same document. R1-24 was closed on a partial propagation. Decorative (the substantive point — ecosystem-scale, popular — holds, as R1-24 itself noted), but a leaf-node reader gets two different star counts from one report.
- **Required fix:** change §1.5 line 230 "46k-star" to "~87.1k-star" (or drop the count), matching §7 and `[^ClaudeMem]`.
- **Grade:** likelihood certain (verified) · impact low (decorative; internal contradiction only) · complexity trivial. Corroboration: the corrected figure exists elsewhere in-doc; this instance is stale. **Pattern: un-propagated repair (closed on partial application).**

---

## Verified clean this round (recorded so not re-raised)

- **`[^Minja]` (arXiv 2503.03704) — MINJA — HIGH, leaf-node (web) this round.** Confirmed id, title ("Memory Injection Attacks on LLM Agents via Query-Only Interaction" / "A Practical Memory Injection Attack against LLM Agents"), and figures: **~98.2% injection success / ~76.8% attack success** (secondary coverage: ">95% injection / ~70% attack under idealized conditions"). The R2-8 MINJA leg — previously "correct-in-fact but untraced-as-cited" — is now **traced and correct**. Discharges the traceability half of R1-28/R2-8's MINJA concern.
- **§4 body attack-success figures — R2-8 body repair LANDED.** All standing `~90%`/`80–99%` occurrences in §0–§4 are now either (a) mem0 token-reduction context (lines 293, 1434 — different meaning, not attack success) or (b) inside explicit retraction/correction statements (§4 lines 443, 450). No standing attack-success `~90%` regression survives in my slice. The corrected wide band (≤32.5% environment-only up to ~76.8–98.2% MINJA) is consistently stated (§4 lines 448–457) and attributed to `[^EnvInjectedMemory]` + `[^Minja]`.
- **`[^MemoryPoisonSurvey]` footnote (R2-9b) — LANDED.** "80–99%" now appears only inside the removal note, not as a standing claim; footnote correctly disclaims that the survey carries no attack-success numbers.
- **`[^LettaSleep]` footnote (R2-9c) — LANDED.** Git-branch clause moved out of the primary-source claim list; "community best-practices forum" now named in the source line.
- **`[^SubagentDocs]` footnote (R2-12) — LANDED.** "v2.1.33+" no longer stands in the descriptive text; attributed to the community report (shanraisshan), feature doc-confirmed / version community-only. (Note: this is the *correct* execution of the same repair R3-1 shows `[^MemoryDocs]` failed to execute — so the fix pattern is known to blue; the miss is isolated to `[^MemoryDocs]`.)
- **`[^FactsFirstClass]` (arXiv 2603.17781)** — §2.1 60%/252×/exact-match — HIGH (carried from prior rounds; re-read in context, attribution to this footnote is now correct and `[^ConsolidationProblem]` correctly retains only the four-levers/decay claims).
- **§3 Auto Dream (`[^AutoDream]`/`[^DreamSkill]`)** — correctly confined to §10 Unverified; §3 Consequence 1 reads "suggestive, not strongest" (R1-10 holds). Not laundered. Not a gap.
- **§4 CVE `[^MemoryPoisonCve]`** — CVE-2026-21852 / "removed from system prompt" remain medium-confidence-tagged (R1-29 disclosure intact); no new regression. Post-cutoff vendor-blog sourcing still unresolvable from here (carried friction).

---

## Corroboration-confidence summary for the slice

- HIGH: `[^Minja]` (re-verified web), `[^FactsFirstClass]`, `[^MemZero]`, `[^FaultyMemories]`, `[^ZepGraphiti]`, `[^MemorySurvey]`, `[^LocalTranscripts]` (§1.4), `[^OkfSpec]`, §4 corrected band attribution.
- MEDIUM (disclosed as such, accepted): `[^MemoryPoisonCve]` id-vector mapping + system-prompt-removal (R1-29 tags), `[^EnvInjectedMemory]` (abstract-level, R2-8-confirmed ≤32.5%).
- CONTRADICTED-BY-ITSELF / STALE: `[^MemoryDocs]` v2.1.59+ (R3-1); §1.5 "46k-star" (R3-2).

## Notes for consolidation (do not fold into findings.md until instances 2–3 return)
- R3-1 and R3-2 are both the *repair-recorded-but-not-executed* pattern. They argue the round-2 "CLOSED"/"repair landed" ledger was verified from the change log, not re-followed at the leaf node — recommend the consolidated findings re-audit every R-numbered "closed" repair by grep, not by trusting the CHANGELOG.
- No new *substantive* (non-citation) gap found in §0–§4 this round; the design residuals live in §12–§13 (other slices).
