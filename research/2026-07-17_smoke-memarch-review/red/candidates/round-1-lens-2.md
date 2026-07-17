# Round 1, Lens 2 (L2) — Logic & Completeness Audit

**Audited document:** C:/Users/gbloc/Projects/special-circumstances/research/2026-07-17_smoke-memarch-review/blue/report.md

**Lens scope:** Logic soundness, leaps of faith, unexplored alternatives, template compliance.

**Audit method:** Full-document re-read; anchor each finding to section heading + quoted text; verify premises and counterfactuals.

---

## Findings

### L2-F1: Mislabeled supersession in Disposition Summary (HIGH)

**Location:** "## Disposition Summary" → "### Superseded ✗" section

**Claim:** "Memory-architecture's blocking security gates (mit.1–2, ingest screening, taint-boundary allowlist inversion) → **not implemented; research-stage fixes R4-1/R4-2 never committed**."

**Gap:** Supersession requires that an *alternative* mechanism was built and prioritized in place of the original. The report shows only that the original items were deferred or abandoned—no documented successor. The three items listed under "Superseded" lack evidence of replacement:

1. **Security gates:** Report finds "Zero implementation of `mit.1` (trust-tier schema) or the two ingest gates" (H5). The report then lists this as *superseded*, but names no replacement gate or mechanism.
2. **Auto Dream coordination:** Report finds "no native Auto Dream integration attempted" (H4). No successor layer cited as a replacement for Auto Dream's consolidation function.
3. **OKF schema/promotion ladder:** Report cites "project-memory skill" as lightweight alternative—this IS true supersession. Correct categorization here.

**Impact:** The "Superseded ✗" heading conflates "unimplemented" (items 1–2) with "architecturally swapped" (item 3), obscuring the actual disposition. Items 1–2 should be listed as "Deferred / Abandoned" (no successor named), not "Superseded."

**Recommendation:** Reclassify security gates and Auto Dream coordination under a "Deferred / Abandoned" section; retain project-memory under "Superseded" with OKF as the superseded item.

---

### L2-F2: Undefined reference to "the two ingest-edge gates" (MEDIUM)

**Location:** "### H5: Build proceeded under re-scoped margin..." → Item 1 ("Security gates (mit.1–2, ingest screening) NOT shipped")

**Claim:** "The research report identifies 'blocking core = two ingest gates + mit.1 trust tiers'" (quoting H5 §1).

**Gap:** The report defines "the two ingest gates" by reference to memory-architecture report §6.3, risk matrix row 50, but neither the gate identities nor their definitions are imported into this audit surface. The footnote [^L1Mit1Mit2Search] reads:

> "Memory-architecture report §6.3 risk matrix row 50 identifies 'Blocking core = the two ingest-edge gates (external-ingest never auto-promotes; injection screening at capture) + mit.1 trust tiers...'"

The parenthetical describes *functions* ("never auto-promotes", "screening at capture") but does not establish these as distinct gates with separate names or implementations. Without seeing the original research document, a reader cannot verify:

1. Are "external-ingest" and "injection screening" truly two gates, or are they two features of one gate?
2. Does the original research name them as "gate 1" and "gate 2," or is blue inferring the count?
3. Is the report accurately reflecting the research, or has it mis-parsed a compound recommendation into a two-item checklist?

**Impact:** MEDIUM. The verification surface (this audit) cannot independently confirm whether "two ingest-edge gates" is accurate or if blue has double-counted or misread the original scope. The finding rests on an external document not brought into audit context.

**Recommendation:** Either import the gate definitions and rationales into this report (quote the original research), or flag the dependency: "Dependent on verification in source research document (memory-architecture report §6.3)."

---

### L2-F3: Unverified premise in Auto Dream integration finding (MEDIUM)

**Location:** "### H4: Recommendations superseded by native Auto Dream..." → Item 3 ("No attempt at Auto Dream integration")

**Claim:** "The efficiency-phase plan (PRs #14–22) focuses entirely on debate-engine cost levers... No FEOV, PC, or sleeper-service agent was modified to check Auto Dream's native flag or coordinate with its consolidation pass."

**Grounding issue:** The finding's soundness depends on:

1. ✓ Verified: Integration code is absent (multiple search surfaces, all negative).
2. ✗ Unverified: Auto Dream "reportedly... flag-gated, rolling out" at research time and "apparently becoming available" afterward (per H4 §3, confidence graded MEDIUM due to live-source drift).

The report cites "live-source drift" as a gap-pattern (per "Gap-pattern pre-flight check" §3) but does not attempt to verify the current status of Auto Dream. The conclusion "should have integrated" rests on the counterfactual premise that Auto Dream is currently available, feature-complete, and compatible with the special-circumstances architecture. Without verifying this premise, the finding is incomplete: the project may have correctly identified Auto Dream as unavailable or unstable post-research.

**Impact:** MEDIUM. The severity of "missing Auto Dream integration" depends on whether Auto Dream is actually available now. The report acknowledges this gap (MEDIUM confidence) but closes the loop via the absence of integration code alone, not by verifying whether integration was even feasible.

**Recommendation:** Add a follow-up verification: "Auto Dream current status (flag availability, maturity, API compatibility) not verified in this audit. Recommendation for next round: query Auto Dream vendor/community status as of HEAD date."

---

### L2-F4: Missing design-decision counterfactual on qmd vs. memory-architecture consolidation (MEDIUM)

**Location:** "### H3: Typed-concept differential shipped minimally..." → Item 3 ("Lifecycle machinery (decay windows, per-concept event/clock triggering, dedup via semantic recall) absent")

**Claim:** "The qmd recall layer (PR #18, committed) is a **DIFFERENT** recall mechanism — it provides retrieval search over markdown documents, not semantic dedup of memory concepts."

**Completeness gap:** The report correctly distinguishes qmd (retrieval) from memory-architecture (consolidation/lifecycle). But it doesn't explore whether qmd represents a deliberate architectural *pivot*—i.e., "defer lifecycle machinery, ship retrieval first"—or a collision of independent work. The report concludes (H3 §5, "Interpretation" paragraph):

> "The project adopted a scaled-down memory approach rather than implementing the full architecture."

And then in "Open Questions Carried" §4:

> "Qmd provides retrieval; memory-architecture provided consolidation and lifecycle. They are orthogonal scopes. No evidence that qmd was intended as a substitute for the consolidation layer."

And later (§5):

> "The project-memory skill is a pragmatic alternative at smaller scope. No documentation explains this as a deliberate substitution; it appears to be a independently-pursued simpler approach that filled the gap left by memory-architecture's deferral."

**Issue:** The report infers that qmd adoption was "independent" (not a deliberate pivot) but does not verify this via:

- Commit dates: Did project-memory ship *after* memory-architecture deferral, suggesting causation?
- PR discussions: Do any commits or PRs reference memory-architecture deferral as a reason for choosing qmd?
- Design docs: Does `plans/efficiency-phase.md` discuss qmd as a recovery strategy or a separate initiative?

Without these checks, the "independently-pursued" claim is an inference, not verified. If qmd *was* a deliberate pivot, the disposition shifts from "abandoned" to "phased defer" (a design choice), which is a material difference.

**Impact:** MEDIUM. The finding that memory-architecture consolidation machinery is missing is not wrong, but the *interpretation* of why it's missing (abandonment vs. phased deferral) is unverified. This affects how seriously the absence should be weighted.

**Recommendation:** Add targeted verification: "Timeline check: compare commit dates of memory-architecture deferral (plan deletion) against qmd PR #18 and project-memory adoption to establish sequence. Query `plans/efficiency-phase.md` for explicit mention of memory-architecture deferral or qmd as a deferral strategy."

---

## Summary

| Finding | Category | Severity | Status |
|---------|----------|----------|--------|
| L2-F1 | Mislabeled disposition | HIGH | Reclassify "Superseded" items into "Deferred" and "Superseded" buckets separately. |
| L2-F2 | Undefined reference (external scope) | MEDIUM | Import gate definitions or flag external dependency for source verification. |
| L2-F3 | Unverified counterfactual premise | MEDIUM | Verify Auto Dream availability/maturity as of 2026-07-17 before concluding "should have integrated." |
| L2-F4 | Missing design-decision context | MEDIUM | Verify qmd adoption timing and intent against memory-architecture deferral to distinguish "pivot" from "coincidence." |

---

## No Template Friction

The audit structure (disconfirming hypotheses, evidence chains, disposition summary, open questions, methodology notes) fits the SMOKE audit scope well. No capability gaps or protocol misfits encountered.

