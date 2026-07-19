# Red Audit — Round 2, Lens L5 (Logic & Completeness)

**Seat:** red-lens-r2-L5 | **Run date:** 2026-07-19 | **Audit scope:** round 1 repairs + logical consistency

## Summary

All 10 repairs from blue's R1 CHANGELOG are implemented in the living report. Repairs R1-1 through R1-10 successfully address their corresponding gap requirements: citation fixes, provenance reading, footnote corrections, scope conditions, hedging, composition rules. Core logic is sound and well-hedged throughout. Three issues found: broken footnote reference (orphaned citation), incomplete dated-specific retirement (Q4 2026 in open question), and undefended tier classification (permission decisions).

---

## Findings

### L5-F1 — Broken footnote reference in Catechism case-against

**Location:** Catechism "4. The case against" (line 92) — bullet point *"Risk-accepted residuals we are choosing to live with, named: silent tool-result truncation with no audit marker;[^ToolTruncation]"*

**Problem:** Footnote reference [^ToolTruncation] is cited but the footnote definition does not exist. R1-4 required retiring the dev.to citation (which did not exist as a published article); the footnote definition was deleted but the reference in the Catechism was not. This leaves a broken link—a reader following the citation will find no definition.

**Required fix:** Either restore the ToolTruncation footnote with a working citation, or remove the reference [^ToolTruncation] from line 92 and reground the truncation claim in an existing footnote (the claim is currently grounded in `[^ToolTruncationLimits]` and `[^BinaryOtelNames]` in §5, which provide the maxResultSizeChars evidence).

**Acceptance check:** Grep report for [^ToolTruncation] — zero matches should result (the reference is removed), OR the footnote definition reappears with a working source (nist.gov, platform.claude.com, or other primary).

**Severity:** low (cosmetic, does not corrupt logic) | **Likelihood:** high (reference clearly dangling) | **Impact:** low (the claim is grounded elsewhere) | **Complexity:** trivial

---

### L5-F2 — Incomplete R1-5 repair: dated NIST reference persists in open question

**Location:** Open question 7 (line 516) — *"If NIST's Q4 2026 interoperability profile mandates reasoning-step logging, what would a compliant implementation look like given the faithfulness problem?"*

**Problem:** R1-5 required retiring dated specifics (2026-02-17 launch date, April listening sessions, Q4 2026 profile) OR re-sourcing to NIST primary. The main narrative in §8 was successfully de-dated ("Industry standards for agent audit logging (NIST and others) are in development…"). However, open question 7 still references "NIST's Q4 2026 interoperability profile" without a source. This is inconsistent: the body text no longer makes dated claims about NIST, but the open question does.

**Required fix:** Either retire the date from open question 7 ("If NIST's reasoning-step logging mandate arrives in future versions…") or re-source the Q4 2026 date to a NIST primary (nist.gov press release, whitepaper, roadmap). The footnote currently sources only "NIST and others" generically, without the specific 2026 date.

**Acceptance check:** Read open question 7; if the Q4 2026 reference is present, confirm a nist.gov-based footnote supports it. If the date is retired, confirm the question is re-phrased without the calendar reference.

**Severity:** low-medium (inconsistency, not factual error) | **Likelihood:** medium (the declination reasoning cited "dated specifics" as the liability) | **Impact:** low-medium (undisclosed unverified date in an open question, not the main narrative) | **Complexity:** trivial

---

### L5-F3 — Logic gap in Tier 1 definition: permission decisions not directly observable via OpenTelemetry

**Location:** §6 "Tier 1 — directly citable (facts of execution)" (line 356–360) — *"Tool-call counts and sequences; tool inputs and outputs; token counts and cost; request latency; permission approvals and denials; retry and rework patterns; context-window depth and pressure; the gap between a user's stated goal and the tool calls that followed."*

**Problem:** The list claims "permission approvals and denials" are Tier 1 (directly citable). However, §4 OpenTelemetry instruments enumerate the available signals: `claude_code.interaction`, `claude_code.tool`, `claude_code.tool.execution`, `claude_code.tool.blocked_on_user`, plus event names `tool_decision`, `tool_result`, `permission_mode_changed`, etc. There is no explicit permission-approval or permission-denial event. Only `permission_mode_changed` (when the user switches the permission mode) appears.

This suggests:
- Permission decisions (approve/deny a specific tool call) are not directly emitted as OpenTelemetry events
- They might be inferrable from the presence/absence of denied-tool-call events (Tier 2 pattern inference: "absence of denial events" per line 365–366)
- But they are not directly observable facts of execution (Tier 1)

The report later contradicts itself (line 365–366), placing "permission-boundary adherence inferred from the absence of denial events" in Tier 2 (pattern inference), not Tier 1. This is correct but inconsistent with Tier 1's claim of direct citeability.

**Required fix:** Remove "permission approvals and denials" from the Tier 1 list, or add a clarifying note that approval/denial is observable only as a Tier 2 absence-of-event inference (denied tool calls did not appear), not as a direct permission-decision instrument. Verify against §4's instrument enumeration that permission decisions ARE available as first-class events, or retract them from Tier 1.

**Acceptance check:** Read Tier 1 definition; if permission decisions remain, verify that §4 names an explicit OpenTelemetry permission-decision instrument (e.g., `claude_code.permission_decision` or equivalent). If no such instrument exists, remove the claim or move to Tier 2.

**Severity:** low (affects boundary of classification, not the validity of the tiers themselves) | **Likelihood:** low-medium (permission modes ARE tracked, but not per-decision approvals) | **Impact:** low-medium (misclassifies a category used for grading claims, could lead to overconfident citation of permission decisions as Tier 1 facts) | **Complexity:** low (either add the instrument or remove the claim)

---

## Repairs Verified

All 10 round-1 repairs verified implemented:

| Gap | Repair | Status |
|-----|--------|--------|
| R1-1 | Performativity figures re-cited to arXiv:2603.05488, both 0.417 (MMLU) and 0.012 (GPQA-Diamond) with task-dependence | ✓ PASS |
| R1-2 | Pinned files stated as recoverable; serialization hypothesis quoted and adjudicated; independence provisionally retracted | ✓ PASS |
| R1-3 | feov-record blue verb list quoted in full; blue-vs-red seat separation stated | ✓ PASS |
| R1-4 | dev.to citation removed; tool-truncation re-grounded on binary evidence | ✓ PASS (orphaned reference L5-F1) |
| R1-5 | Main narrative de-dated; dated specifics remain in open question 7 (L5-F2) | ✓ PARTIAL |
| R1-6 | Compliance API conflict disclosed: ~30 vs 260+ in table cell and footnote | ✓ PASS |
| R1-7 | Headline and Catechism Q3 include scope conditions (v2.1.215, default-configured, non-interactive); causality stated | ✓ PASS |
| R1-8 | Absolute "ever" hedged to "on this version" (v2.1.215) | ✓ PASS |
| R1-9 | Durability argument added: disconfirmability via external evidence is the key advantage over thinking blocks | ✓ PASS |
| R1-10 | Composition rule for multi-tier claims added: grade at weakest leg | ✓ PASS |

---

## Logic Audit Summary

**Headline logic:** Sound. Properly bounded to scope conditions (v2.1.215, default-configured, non-interactive, `showThinkingSummaries` unset). Negative claim is fenced and open questions identify experiments that could overturn it.

**Recommendation logic:** Sound. Artifact-based recording is recommended as a better medium than thinking blocks on durability and disconfirmability, while explicitly not claiming superiority on sincerity. Cost-benefit clearly stated.

**Case-against:** Comprehensive and honest. Acknowledges limitations: sampling (one machine), version-binding (v2.1.215 + history of server-side changes), stale community citations (closed-not-planned issues from v2.1.71), method failure modes (binary string extraction), and faithfulness limits (reasoning traces unreliable evidence). These are well-integrated into the risk matrix.

**Risk matrix:** Reasonably comprehensive. Dispositions are explicit (mitigate vs risk-accept). One row explicitly risk-accepts "artifact self-reports are post-hoc rationalization" with the rationale that independent corroboration is too expensive; this is disclosed as the report's honest limit (§8).

**Unexplored alternatives:** None flagged. The declinations to re-run experiments (showThinkingSummaries test, OTLP collector setup) are reasoned; the declination to re-fetch secondary sources turned out to miss defects (broken dev.to citation, NIST miscitation) but the reasoning was sound at the time (prioritization call).

**Template compliance:** Catechism structure (7 questions + "of interest" + risk matrix) follows the specified format. Argument is durable (git-tracked), repeatable, and cited with access dates. No claims are laundered into fact; unverified figures are labeled. Minority-lane claims carry attribution. Merge-verified claims are flagged.

---

## Verdict

**PASS (with conditions):** All substantive R1 repairs are implemented and sound. Three low-to-trivial issues require cleanup (broken reference L5-F1, incomplete declination L5-F2, undefended tier classification L5-F3) but none corrupt the core logic. The report's headline, recommendation, and risk posture are well-reasoned and properly hedged.

**Conditions for closure:**
1. Remove or redefine [^ToolTruncation] reference (L5-F1)
2. Retire Q4 2026 date from open question 7 or source it to NIST primary (L5-F2)
3. Verify permission decisions are Tier 1 observable or move them to Tier 2 (L5-F3)

---

## Archive Notes for Future Rounds

- **Partial-fix signal:** Round 1 repair R1-5 was successful on the main narrative (§8) but incomplete across the document (open question 7). Future audits should check that dated-specific retirements are consistent report-wide.
- **Orphaned-reference pattern:** Retirement of a footnote (R1-4, dev.to) left a dangling citation. Add a lint pass or grep check to flag unreferenced footnotes before shipping.
- **Permission-decision ambiguity:** OpenTelemetry signal names are explicit, but the report's tier classification of permission decisions needs reconciliation with the available instruments. Recommend documenting which tier each observable-fact type belongs to in §4 directly (not inferred).

