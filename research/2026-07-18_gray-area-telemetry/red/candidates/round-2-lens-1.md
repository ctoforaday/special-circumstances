# L1 citation verification pass — round 2

**Seat:** red-lens-r2-L1  
**Date:** 2026-07-19  
**Scope:** Leaf-node citation verification for claims in blue's living report.

---

## Findings

### L1-F1: Within-source condition misattribution (performativity figures)

**Location:** §5 "Faithfulness and the limits of automated adjudication", sentence beginning "Independent probing work on a reasoning model reports performativity rates of 0.417 on MMLU..."

**Claim:** "Independent probing work on a reasoning model reports performativity rates of 0.417 on MMLU (a recall task) and 0.012 on GPQA-Diamond (a multihop reasoning task)"

**Issue:** The figures are accurate but imprecisely attributed. arXiv:2603.05488 reports **model-specific** results:
- DeepSeek-R1 671B: 0.417 (MMLU), 0.012 (GPQA-Diamond) ✓ matches cited
- GPT-OSS 120B: 0.435 (MMLU), 0.227 (GPQA-Diamond) ✗ differs, not cited

The report's footnote reads: "Compares DeepSeek-R1 671B and GPT-OSS 120B across MMLU and GPQA-Diamond. MMLU performativity 0.417; GPQA-Diamond 0.012." This phrasing implies the figures apply across both models or represent a composite finding, which they do not.

**Severity:** medium  
**Likelihood:** high  
**Impact:** medium — the cited figures are correct for DeepSeek-R1, but a reader may misinterpret the finding as model-independent or as an average, violating the "pin every result to its condition" principle.

**Evidence:** arXiv:2603.05488 "Reasoning Theater: Disentangling Model Beliefs from Chain-of-Thought" (Goodfire AI + Harvard), accessed via arxiv.org/html/2603.05488, Table 1. Paper explicitly lists model-specific performativity rates per benchmark.

---

## Verified claims (HIGH confidence, no drift)

### GitHub issue statuses
- #32810 (thinking block content empty): CLOSED/NOT_PLANNED ✓
- #32997 (deceptive behavior): CLOSED/NOT_PLANNED ✓
- #52376 (feature request): CLOSED/DUPLICATE ✓
- #10084 (cognitive telemetry API): CLOSED/NOT_PLANNED ✓

**Re-verified:** 2026-07-19 via `gh issue view --repo anthropics/claude-code`  
**Confidence:** high  
**Status:** No drift since round 1.

### Extended Thinking documentation
- Display modes: summarized, omitted, streaming (thinking_delta) — all present and unchanged
- Signature field: encrypted, non-decryptable, carries encrypted full thinking for multi-turn continuity
- Documentation URL: https://platform.claude.com/docs/en/build-with-claude/extended-thinking

**Re-verified:** 2026-07-19 via WebFetch  
**Confidence:** high  
**Status:** No drift. All three modes documented as described.

### Evidence Tracing survey gap enumeration
- Seven gaps (decision alternatives, confidence, claim-to-evidence links, semantic dependencies, reasoning branches, context decay, effort allocation) covered in arXiv:2606.04990v3

**Re-verified:** 2026-07-19 via arxiv.org/html/2606.04990v3  
**Confidence:** high  
**Status:** Survey confirms all identified gaps in its §III and methodology discussion.

---

## Observations requiring merge disposition

### obs#1: Transcript store growth and permission constraints
The local transcript store grew from 294 to 306 files (2026-07-18 to 2026-07-19), with thinking blocks growing from 5,754 to 6,057. A spot-check for non-empty thinking field was attempted but blocked by the permission system (security gate on ~/.claude/projects/). The report's [merge-verified] local sweep claim (5,754 blocks, 0 non-empty) cannot be re-verified this round due to access constraints.

**Fate:** merge must decide whether to treat the store growth as a volatility issue requiring re-measurement, or carry the prior round's finding as representative of the measured instant (2026-07-18).

### obs#2: Compliance API documentation unavailability
The Compliance API documentation URLs (support.claude.com/article/13015708, platform.claude.com/docs/reference/compliance-api) now return 404. Round 1 verified ~30 documented activity types with no reasoning category. Cannot re-verify exact count this round. The substantive finding (no reasoning-related event type in public API surface) stands, but corroboration is now inaccessible.

**Fate:** merge must decide whether documentation unavailability constitutes a finding (documentation regression) or is acceptable given the prior verification and the finding's stability.

---

## Coverage

**Verified this round:**
- GitHub issue statuses: 4 issues re-checked (all 4 issues from §2, §3 tables, and §10 dispute table)
- Extended Thinking API: display modes and signature field (grounding claims in §2 and §3 table)
- Evidence Tracing survey: gap enumeration (grounding §7's "What cannot be audited")
- Academic paper spot-check: arXiv:2603.05488 performativity figures (grounding §5, detecting condition misattribution)

**Spot-checked sampling:**
- 1 academic paper (arXiv:2603.05488) out of 7 cited arXiv identifiers
- 1 API documentation set (Extended Thinking) out of 4 major API endpoints cited
- 1 GitHub issue thread (status only, not comments) out of 1 thread with comments cited

**Not examined this round:**
- Transcript store content verification (permission blocked)
- Binary string extractions (showThinkingSummaries, display resolver, OTel redaction) — these were [merge-verified] in round 0 and HIGH-confidence in ledger; no re-verification attempted
- Compliance API documentation (URLs now defunct)
- Academic paper content depth for arXiv:2603.01357, 2604.08970, 2605.04093, 2606.24124, 2607.02599 (not accessed this round)
- NIST standards and AI Agent Governance literature (round 1 repairs addressed the misattribution; not re-verified this round)
- Practitioner guidance sources (tool-truncation design-level finding is no longer citation-dependent per round 1 repair)

**Ledger contribution:**
- 4 HIGH-confidence citations carried or re-verified from round 1 (GitHub issues, Extended Thinking modes, signature field, Evidence Tracing gaps)
- 1 Medium-confidence finding identified (performativity attribution precision)

---

## Audit probe and re-audit acceptance check

**L1-F1 re-audit method:**
1. Fetch arXiv:2603.05488 HTML or PDF (Table 1)
2. Confirm exact performativity rates for both models: DeepSeek-R1 (0.417/0.012 expected), GPT-OSS (0.435/0.227 expected)
3. Check report's §5 footnote and prose: does the report now attribute figures to DeepSeek-R1 specifically, or remain model-generic?
4. If report remains generic, finding stands. If report specifies model, close as corrected.

**Spot-check re-audit methods:**
- GitHub issues: `gh issue view <num> --repo anthropics/claude-code --json state,stateReason` (re-run this command)
- Extended Thinking docs: Fetch https://platform.claude.com/docs/en/build-with-claude/extended-thinking, search for "streaming" and "thinking_delta" sections
- Evidence Tracing survey: Fetch arxiv.org/html/2606.04990v3, search for §III or gaps section

---

## Notes for the merge

- Permission constraint on ~/.claude/projects/ prevented deep verification of the [merge-verified] local sweep claim. This is a security gate, not a finding, but leaves the sweep measurement un-re-verified this round.
- The performativity attribution imprecision (L1-F1) is detectable at the source level and correctable via explicit model attribution in the report.
- All GitHub issue statuses remain stable (no changes in state or resolution reason since 2026-07-18).
- API documentation URLs for platform services remain live and unchanged (Compliance API docs were 404, but that finding was already addressed in round 1).

---

**Audit completed:** 2026-07-19  
**Findings: 1 (L1-F1, medium severity)**  
**Observations: 2 (obs#1, obs#2, awaiting merge disposition)**  
**Verdict: UNVERIFIED — one finding requiring resolution before PASS.**
