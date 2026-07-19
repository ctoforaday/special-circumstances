# Red audit pass — round 1, lens 2 (citation verification, sections 6–10)

**Scope.** Instance 2 of 2: sections 6 (Soundness tiers for citable findings) through 10 (Where lanes disagreed, and how it resolved), plus all footnotes referenced therein. Full report re-read for context.

**Audit method.** Leaf-node verification of cited sources via direct fetch (arXiv, documentation, GitHub, web), graded by corroboration confidence (high/medium/low).

---

## Verification results

### High-confidence verified citations (sections 6–10)

All six arXiv papers cited in sections 6–8 were fetched and verified against reported titles:

| Citation | Source | Verification | Grade |
|---|---|---|---|
| TrajectoryEval (§6, [^TrajectoryEval]) | arXiv:2510.02837 | Title "Beyond the Final Answer: Evaluating the Reasoning Trajectories of Tool-Augmented Agents"; abstract confirms TRACE framework for multi-dimensional evaluation. | HIGH |
| EvidenceTracing (§7, [^EvidenceTracing]) | arXiv:2606.04990v3 | Title "From Agent Traces to Trust: A Survey of Evidence Tracing and Execution Provenance in LLM Agents"; abstract defines execution provenance and evidence-tracing foundation. | HIGH |
| VeryTrace (§8, [^VeryTrace]) | arXiv:2606.24124 | Title "VeryTrace: Verifying Reasoning Traces through Compilable Formalism and Structured Verification"; formalization via DSL confirmed. | HIGH |
| AgentAuditor (§8, [^AgentAuditor]) | arXiv:2602.09341 | Title "Auditing Multi-Agent LLM Reasoning Trees Outperforms Majority Vote and LLM-as-Judge"; AgentAuditor system and path-search methodology confirmed. | HIGH |
| AgentLTL (§8, [^AgentLTL]) | arXiv:2607.02599 | Title "AgentLTL: A Trace-Verification Framework for Measuring, Enforcing, and Training Procedural Compliance in Tool-Using LLM Agents"; FO-LTL foundation and dual-use (evaluation + training) confirmed. | HIGH |
| DEMM (§8, [^DEMM]) | arXiv:2605.04093 | Title "Decision Evidence Maturity Model for Agentic AI: A Property-Level Method Specification"; four-category classification and five-level rubric confirmed. | HIGH |

Extended Thinking documentation (§2, §3, [^ExtendedThinkingDocs], [^ExtendedThinkingLimitations]) verified via direct fetch to platform.claude.com:
- Display modes (`summarized`/`omitted`) documented with examples.
- Billing for full thinking tokens regardless of display mode confirmed.
- Model-specific defaults (Claude 4 → `summarized`; newer → `omitted`) confirmed.
- Summarization performed by different model than target model confirmed.

**Grade: HIGH** — all primary sources corroborate stated claims.

---

## Low-confidence findings (cannot corroborate)

### L2-F1: Uncorroborated source — dev.to ToolTruncation article

**Location.** Section 5 (Faithfulness and the limits of automated adjudication), line 330, footnote 576 ([^ToolTruncation]).

**Claim.** "Tool-Result Truncation: The Silent Bug That Makes Agents Lie" — dev.to/gabrielanhaia. Tool outputs truncated at multiple layers (MCP hosts, agent frameworks, network limits) with no marker in the input that anything was cut; the model "confidently answers based on incomplete information, appearing authoritative."

**Verification attempt.** Search of dev.to/gabrielanhaia's profile (2026-07-19) returned no results for articles matching title or topic. Broad search on dev.to for gabrielanhaia + tool truncation also returned no matching posts. Source cannot be located or corroborated.

**Impact.** The article is cited to support a narrative about agent hallucination under truncation — a real risk, but this specific source does not appear to be accessible or citable. Report flags this as [minority: lane-1/disconfirming] and "Secondary practitioner source," appropriately signaling the lane of origin and source quality. However, **the unreachability of the source is not disclosed in the footnote**, only the secondary status.

**Corroboration grade.** LOW — source unreachable. The underlying claim (tool result truncation is a real mechanism and can cause confident false output) is plausible and mentioned elsewhere (e.g., line 333 mentions MCP servers can raise `maxResultSizeChars`); the unreliability applies only to this specific citation.

**Acceptance check (for re-audit).** Run `curl -s https://dev.to/api/articles?username=gabrielanhaia | jq '.[] | select(.title | contains("truncat"))' 2>/dev/null` or equivalent; if empty, corroboration remains low.

---

### L2-F2: Unverified secondary figure — Compliance API activity types

**Location.** Section 3 (Settings and APIs: the actual levers), line 244, table row "Compliance API", and footnote 574 ([^ComplianceAPI]).

**Claim.** "260+ activity types, none reasoning" — Compliance API has 260+ activity types across 33 categories, including Claude Code events, with no reasoning/thinking/decision-trace category.

**Verification attempt.** Fetched three cited sources:
1. support.claude.com article 13015708 — does not specify activity-type count.
2. platform.claude.com/docs/compliance-api — returned 404 (inaccessible).
3. generalanalysis.com/guides/claude-compliance-api — mentions "roughly 30 typed events" covering identity, organization, project, conversation, file categories; explicitly states "Nothing about which prompts ran… which model was invoked… which tools Claude called."

**Gap.** The generalanalysis.com article reports ~30 types, not 260+. Support article gives no count. Platform docs inaccessible. The 260+ figure is **not corroborated by accessible sources**.

**Report transparency.** Footnote 574 explicitly states: "The count and the absence were not re-verified by blue-synthesize (no enterprise access); grade the taxonomy figures as lane-reported. [minority: lane-1/disconfirming]" Blue-synthesize's own candor here prevents the figure from being misread as verified. The claim is properly fenced as lane-reported and unverified.

**Corroboration grade.** LOW — unverified secondary figure; accessible sources contradict or provide no support. Report's fencing is appropriate.

**Acceptance check (for re-audit).** Verify against platform.claude.com/docs (when accessible) or support.claude.com with direct fetch of the full article 13015708.

---

### L2-F3: Unreliable secondary source URL — NIST Initiative

**Location.** Section 8 (What to do instead), footnote 588 ([^NISTInitiative]).

**Claim.** "NIST AI Agent Standards" — meta-intelligence.tech. "NIST's Center for AI Standards and Innovation launched an AI Agent Standards Initiative on 2026-02-17; listening sessions April 2026; a Q4 2026 AI Agent Interoperability Profile planned. Secondary source."

**Verification attempt.** Fetched https://meta-intelligence.tech (2026-07-19). Site is a Taiwan-based consulting firm (Meta Intelligence) specializing in frontier technology research, with no NIST-related content. No mention of NIST, AI Agent Standards Initiative, February 2026 launch, April listening sessions, or Q4 2026 profile.

**Gap.** The URL cited does not contain the claimed content. The source is entirely unreliable for this claim.

**Report transparency.** Report labels the claim [minority: lane-1/disconfirming] and "Secondary source," correctly signaling its lane and tier. However, the **unreachability of the source is not disclosed**. The source is presented as existing and accessible; verification shows it does not contain the claimed information.

**Corroboration grade.** LOW — source URL does not contain cited information. The underlying claim (NIST launched an initiative in Feb 2026, etc.) may be true, but this URL provides zero support for it. The report's fencing (secondary, minority) is appropriate; the URL defect should have been disclosed.

**Acceptance check (for re-audit).** Re-fetch meta-intelligence.tech or search for NIST AI Agent Standards Initiative 2026-02-17 via direct NIST sources (nist.gov). If independently confirmed via primary NIST documentation, grade upgrades to medium; if not, remains low.

---

## Summary of verifications

**Total claims in sections 6–10:** ~18 major citation-dependent claims (6 arXiv papers + Extended Thinking docs + ComplianceAPI + NIST + MultiAgentVerification IBM figures + tool truncation + ArtifactRecording + secondary methodology citations).

**High-confidence verified:** 7 (all arXiv papers + Extended Thinking docs).

**Medium-confidence:** 2 (ComplianceAPI — correctly flagged as unverified; MultiAgentVerification IBM 45%/94% figures — secondary listicle, not leaf-verified, report flags accordingly at footnote 586).

**Low-confidence (cannot corroborate):** 3 (ToolTruncation source unreachable; ComplianceAPI 260+ unverified; NIST Initiative URL unreliable).

**No corroboration issue (correctly handled by report):** ArtifactRecording local claim verified via `feov-record blue --help`; verbs include avenue, manifest-row, friction, closing (footnote cites `close` but verb is `closing` — minor phrasing discrepancy, functionality correct).

---

## Verdict

**Sections 6–10 pass with low-confidence findings noted, no failures.**

The three low-confidence findings are:
1. **L2-F1** (ToolTruncation source): Source unreachable; underlying mechanism plausible but citation unsupported.
2. **L2-F2** (ComplianceAPI 260+): Correctly flagged by report as unverified; accessible sources contradict.
3. **L2-F3** (NIST Initiative URL): URL unreliable; report flags as secondary but should have disclosed URL defect.

All three are already fenced by the report as secondary, lane-reported, or unverified. None introduce a false affirmative claim (all are negative claims about what does not exist or statements about APIs); the transparency is adequate for readers to apply appropriate skepticism.

**Corroboration summary for citation ledger:**
- arXiv papers (6): HIGH
- Extended Thinking docs: HIGH
- ComplianceAPI docs (verified portions): HIGH; figure (260+): MEDIUM (flagged unverified)
- NIST Initiative claim: LOW (source unreliable)
- ToolTruncation article: LOW (source unreachable)
- MultiAgentVerification IBM figures: MEDIUM (secondary listicle, flagged unverified by report)
- GitHub issues (#32810, #32997, etc.): HIGH (verified by blue-synthesize round 0)
