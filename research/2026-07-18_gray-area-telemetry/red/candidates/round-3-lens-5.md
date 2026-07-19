# Red audit — Round 3, Lens 5 (Logic and Completeness)

**Seat:** red-lens-r3-L5  
**Scope:** Logic gaps, leaps of faith, unexplored alternatives, template compliance in the blue living report.  
**Report audited:** C:/Users/gbloc/Projects/special-circumstances/research/2026-07-18_gray-area-telemetry/blue/report.md  
**STEELMAN focus:** Declined and abandoned avenues (the case AGAINST the design), not just conclusions.

---

## Findings Summary

Nine logic and completeness gaps identified across the report. Severity ranges from low (scope clarification) to medium-high (blocking the conclusive test). No gaps are found in the METHODOLOGY — the leaf-node verification is rigorous. The gaps are in the FRAMING and REASONING at decision points.

### Summary by category

- **Declined avenues** (1): The showThinkingSummaries test is the single strongest counterexample to the headline, but is declined on consent grounds without explicit operator escalation.
- **Unresolved mechanisms** (1): Serialization vs. display-resolver debate is noted but not closed; the causal finding rests on parsimony, not proof.
- **Stale evidence** (1): Closed-not-planned issues reference v2.1.71 when the mechanism moved in v2.1.215; closure date relative to v2.1.215 release is unverified.
- **Mixed verification levels** (2): Unverified figures (IBM 45/94%, 500K maxResultSizeChars) are cited in risk reasoning; Compliance API count discrepancy (30 vs. 260+) is flagged but unresolved.
- **Generalization beyond target** (1): Faithfulness case rests on vendor self-report and other models (DeepSeek-R1, GPT-OSS), not Claude-specific research.
- **Timing assumption not stated** (1): Artifact recording is recommended as externally checkable, but assumes real-time logging, not end-of-run summary.
- **Scope conflation** (1): Adaptive thinking section mixes "can't measure effort" with "can't adjudicate," conflating reasoning quality with acts adjudication.
- **Enforcement cost underestimated** (1): Tier-discipline enforcement is prescribed as "low complexity," but the report itself conflates tiers in its own reasoning.

---

## Findings (lens-scoped)

### L5-F1: Declined avenue blocks the conclusive test

**Severity:** medium-high  
**Location:** "Declined avenues" section, lines-of-inquiry.md, showThinkingSummaries test  
**Anchor:** "writing to the user's global ~/.claude/settings.json is a state-modifying change outside the working tree and outside this seat's consent"

**Gap:** The test to set `showThinkingSummaries=true` on a non-interactive session is **the single strongest counterexample to the headline** "reasoning is almost not recorded." Declining it because of consent is caution against casual side effects, appropriate for routine work. But a finding that **pivots the entire conclusion's validity** merits explicit operator consent, not silent deferral.

The report defers this as open question 1 (§ 14.1), which is honest. But honesty about a gap should not preclude asking the operator to close it.

**Why this matters:** If `showThinkingSummaries=true` produces non-empty thinking in a non-interactive subagent, the headline is wrong. If it still produces empty blocks, the finding is STRONGER. Either outcome is decisive. The consent rationale protects against casual side effects; this test is not casual.

**Recommendation:** Escalate to the operator with explicit stakes before the round ends: "Setting showThinkingSummaries=true is the single test that could overturn the headline. We declined to run it to avoid modifying your settings without consent. Should we proceed, or carry this as an open question for the next run?"

---

### L5-F2: Competing mechanism for empty blocks is noted but not resolved

**Severity:** medium  
**Location:** Section "Provenance and limitations"  
**Anchor:** "The serialization-vs-resolver question: the display-resolver finding... is more parsimonious... versus the serialization claim... Both do not need to be true; the resolver account holds at the leaf."

**Gap:** The report acknowledges a live competing hypothesis: **serialization** (the API returns thinking, but the Claude Code harness serializes the message structure without its content) vs. **resolver guard** (the client sets `display:omitted` before sending). The report judges the resolver account "more parsimonious" because it requires one guard rather than two bugs.

Parsimony is a tie-breaker, not proof. The report then states in §2: "This is a **causal finding** for the non-interactive branch: the display resolver guard forces `display:"omitted"` on that path."

A finding that rests on preferring one plausible explanation over another is a causal **inference**, not a finding with confidence "verified."

**Why this matters:** The headline claim is "the non-interactive sessions are forced to omitted by a display resolver guard." If serialization is equally plausible, this is inference, not causality. The distinction is important for risk: if it's inference, the mechanism could move or be misunderstood; if it's causal, it's stable.

**Recommendation:** Either (a) definitively rule out serialization by showing the API returns non-empty content in a test, or (b) downgrade §2's causal claim to "the display resolver guard is consistent with observed empty blocks and more parsimonious than the alternative serialization hypothesis."

---

### L5-F3: Closed-not-planned issues are stale relative to v2.1.215 mechanism changes

**Severity:** medium  
**Location:** Section 2, "Provenance of the default"  
**Anchor:** "the flag name `tengu_quiet_hollow` does **not** appear anywhere in the v2.1.215 binary, though the `redact-thinking-2026-02-12` beta registration does. The mechanism described in the thread has moved."

**Gap:** The report uses GitHub issue #32810 (closed not-planned) to establish the v2.1.71 mechanism: "the flag `tengu_quiet_hollow` is on" server-side, flipped ~2026-03-10. Then the report shows that in v2.1.215, `tengu_quiet_hollow` is **absent** — "the mechanism has moved."

But the report cites the closed-not-planned status as evidence: "A closed-not-planned feature request is stronger evidence *against* the capability arriving than an open one."

This is a **temporal gap:** the issue was closed about an old mechanism. If #32810 was closed on v2.1.71 (or before v2.1.215 shipped), then its closure says nothing about v2.1.215's current behavior. A feature request closed for v2.1.71 carries no signal about v2.1.215's design.

**Why this matters:** The report uses issue-closure status as evidence of vendor intent ("we're not going to ship raw thinking"). But the issues describe a mechanism that no longer exists in v2.1.215 in the same form. If the issue was closed before the mechanism moved, the closure is stale evidence.

**Recommendation:** Verify for #32810, #32997, #10084, #52376: (a) closure date and reason, (b) v2.1.215 release date. If issues were closed before v2.1.215 shipped, note that their closure reflects v2.1.71 (or earlier) state, not current intent.

---

### L5-F4: Secondary figures used in risk grading are marked unverified

**Severity:** low-medium  
**Location:** Section 9, Risk matrix  
**Anchor:** "Reasoning-quality claim published on Tier-3 evidence | medium | high | low (tier discipline) | **mitigate** — tier label mandatory on every claim"

**Gap:** The risk matrix prescribes "mitigate — tier label mandatory on every claim" with complexity "**low**." But the report marks several figures as "not leaf-verified" and cites them in risk reasoning:

- IBM judge-vs-deterministic: 45% vs. 94% (footnote 620, marked "Secondary listicle; the IBM figures were **not** traced to a primary source this round and are carried as unverified")
- maxResultSizeChars ceiling: 500K (footnote 612, marked "search-derived and not leaf-verified")

These unverified figures are cited in substantive analyses (e.g., tool-result truncation as a risk factor). The complexity of enforcing tier discipline is not **LOW** if the report itself conflates verified and unverified claims in its own risk reasoning.

**Why this matters:** If tier-label enforcement is LOW complexity, it should be demonstrable in the report itself. But the report uses unverified figures in risk severity estimates, suggesting that even internal tier discipline is non-trivial.

**Recommendation:** Either (a) mark risk-matrix entries that depend on unverified sources with a caveat, or (b) pre-verify all figures cited in risk reasoning before assigning them complexity grades.

---

### L5-F5: Faithfulness argument uses mostly vendor self-report and out-of-domain models

**Severity:** medium  
**Location:** Section 2, "Even when captured, the trace is not the reasoning"  
**Anchor:** "models 'often make decisions based on factors that they don't explicitly discuss in their thinking process'" (citing Anthropic Visible Extended Thinking) and performativity figures for DeepSeek-R1 and GPT-OSS

**Gap:** The report's case that "reasoning may not be faithful" rests on:

1. **Anthropic Alignment Science paper** ("reasoning transcripts may not be faithful") — vendor self-report, discusses transcripts generally, NOT Claude Code's thinking blocks specifically
2. **Performativity research** on DeepSeek-R1 (0.417 on MMLU, 0.012 on GPQA-Diamond) and GPT-OSS (0.435 on MMLU, 0.227 on GPQA-Diamond) — examines **different models**, not Claude 4/3.5/Haiku

The report does **NOT** cite published work on Claude models' reasoning faithfulness. The leap from "other models show performativity gaps" + "vendor says reasoning may be unfaithful" to "Claude Code's thinking blocks are untrustworthy" is **not empirically grounded on Claude**.

**Why this matters:** Claude's thinking behavior may differ from DeepSeek-R1 or GPT-OSS. The faithfulness problem is real, but the case for Claude-specific untrustworthiness is vendor-reported, not empirically validated on the target model.

**Recommendation:** Either (a) cite published work on Claude thinking faithfulness, or (b) explicitly state: "The recommendation rests on vendor guidance (Anthropic's position that reasoning may not be faithful) rather than on empirical validation of Claude's reasoning quality. The performativity research on other models is illustrative, not definitive."

---

### L5-F6: Artifact recording is recommended as non-circular, but timing assumption is not stated

**Severity:** low-medium  
**Location:** Section 8, "Artifact-based reasoning recording"  
**Anchor:** "it is still self-report, so it buys durability and non-circularity, not sincerity... the tracing is *adversary-checkable*: a judge can follow a cited avenue to the tool call it names and to the file it changed, and verify or refute the agent's self-report."

**Gap:** The report argues artifact recording is superior to thinking blocks because "it buys durability and non-circularity" and is "adversary-checkable." But the advantage **assumes real-time recording**, not end-of-run summary.

If an agent records "I declined path X because Y" **after** finishing the run, the record is still **post-hoc rationalization**, just externally checkable. The report does not discuss the **TIMING** of artifact recording:
- Real-time (enforced by hooks during work) — circumvents post-hoc rationalization
- End-of-run summary — vulnerable to the same post-hoc rationalization as thinking blocks

The feov-record discipline is mentioned (avenue, manifest-row, friction, closure anchors), but the report does not state whether these are enforced real-time or permitted end-of-run.

**Why this matters:** If artifacts are end-of-run summary, they are just as subject to post-hoc rationalization as thinking blocks. The advantage only holds for real-time, untamperable recording.

**Recommendation:** State explicitly: "Artifact recording is superior to thinking blocks only if recording happens DURING work and is untamperable. If feov-record permits end-of-run summary, the post-hoc rationalization problem persists, and the advantage is durability (git-tracked, not ephemeral) rather than non-circularity."

---

### L5-F7: Adaptive thinking section conflates "can't measure effort" with "can't adjudicate"

**Severity:** low  
**Location:** Section 3, table entry for "Adaptive thinking"  
**Anchor:** "The API does not expose which effort level the model selected or how effort shaped the decision, so identical prompts under different latency conditions may produce different reasoning and identical outputs — making reasoning-quality adjudication impossible without controlled re-execution."

**Gap:** The report states: (a) effort level is not exposed, therefore (b) reasoning cannot be audited. But the report's own recommendation is to audit **ACTS**, not reasoning.

If two agents produce identical tool sequences (identical acts) under different effort levels, does this prevent **acts-level** adjudication? The report doesn't explore this. The effort-exposure gap matters for reasoning **quality**, not for acts adjudication.

This is a minor conflation of scopes (reasoning quality vs. acts adjudication), but it weakens the argument by mixing concerns.

**Why this matters:** The report recommends auditing acts (Tier 1 facts). The adaptive-thinking section raises a concern about reasoning quality (Tier 3 inference). These are distinct failure modes. The section should clarify: "Adaptive thinking makes reasoning-quality claims non-reproducible. This does not affect acts-level adjudication (Tier 1), but does affect reasoning-soundness claims (Tier 3)."

**Recommendation:** Split the concern: acts adjudication is unaffected; reasoning-quality adjudication requires controlled re-execution. Clarify which level the finding applies to.

---

### L5-F8: "No reasoning API" claim searches public surface only (minor, but loose thread)

**Severity:** low  
**Location:** Section 3, "No dedicated reasoning API exists"  
**Anchor:** "This is an absence claim over the documented public surface as searched on 2026-07-18, not a proof of non-existence; undocumented and enterprise surfaces are outside what we checked."

**Gap:** The report properly scopes and honestly caveat the claim. BUT: the Compliance API section (§3, table) flags a discrepancy — lane-1 reported 260+ activity types, but the documented ~30 count does not list any reasoning category.

The report does not resolve: if 260+ is real, could reasoning categories be **undocumented**? The report flags the discrepancy ("lane-1 reported 260+, a count not corroborated by publicly accessible sources") then leaves the Compliance API conclusion ("no reasoning event category") resting on the ~30 public count.

This is not a deep gap (the caveat is there), but it's a loose thread.

**Why this matters:** If the Compliance API has undocumented activity types, reasoning categories could be among them. The conclusion "no reasoning API exists" should either verify the 260+ claim or explicitly state: "On publicly documented surfaces, no reasoning category exists."

**Recommendation:** Either verify the 260+ figure via enterprise access or publicly documented evidence, or explicitly state the conclusion applies to documented public surface only.

---

### L5-F9: Tier-discipline enforcement cost is underestimated

**Severity:** low-medium  
**Location:** Section 6, "Soundness tiers for citable findings" and Section 9 Risk matrix  
**Anchor:** "Composition rule for claims spanning tiers... Grade such claims at the tier of their weakest leg" and "Reasoning-quality claim published on Tier-3 evidence | medium | high | low (tier discipline) | **mitigate** — tier label mandatory on every claim"

**Gap:** The risk matrix prescribes: "mitigate — tier label mandatory on every claim" with complexity "**low**." But enforcing tier discipline is **non-trivial** — the report itself demonstrates this.

The report uses Tier 2 pattern inference (backtracking detection, tool-choice relevance) in the artifact-recording recommendation without explicitly labeling these as Tier 2. For example, "the agent chose this tool for a reason" is Tier 2 inference (pattern inference), not Tier 1 fact. But in Section 8's recommendations, this reasoning is embedded without a tier label.

If the report itself conflates tiers in its own reasoning, then the enforcement cost is not **LOW**.

**Why this matters:** Tier discipline is the gate against laundering Tier 3 (requires external oracle) conclusions under Tier 2 (plausible inference) labels. If the report itself conflates tiers, the enforcement mechanism is weaker than prescribed.

**Recommendation:** Either (a) upgrade the enforcement cost estimate (e.g., "medium — requires disciplined labeling of every claim, and self-enforcement is non-trivial"), or (b) demonstrate proper tier labeling in the report itself as a model, then re-estimate the cost.

---

## Verdict

The report's methodology is rigorous at the leaf node (binary string extraction, transcript sweep, issue-tracker verification). The logic gaps are in the **framing** and **reasoning at decision points** — the choices about which avenues to pursue, how to handle competing explanations, and how to generalize findings.

**None of these gaps overturn the central finding** ("reasoning is not recorded on non-interactive sessions by default"). But they affect:
- **Confidence** in the causal mechanism (L5-F2: resolver vs. serialization)
- **Decisiveness** of the evidence (L5-F3: staleness of issue references)
- **Generalizability** of the faithfulness concern (L5-F5: vendor self-report + other models)
- **Enforceability** of the recommendations (L5-F6, L5-F9)

**Strongest gap:** L5-F1 (blocked conclusive test) — the showThinkingSummaries experiment is the single strongest counterexample and should be escalated to the operator for explicit consent, not silently deferred.

---

## Spot-check acceptance criteria

For re-audit at merge:
1. **L5-F1**: Operator explicitly asked about showThinkingSummaries test; decision recorded (approved, deferred, or executed).
2. **L5-F2**: Either (a) serialization ruled out by API inspection, or (b) causal claim downgraded to inference in §2.
3. **L5-F3**: Issue closure dates and v2.1.215 release date compared; staleness noted if applicable.
4. **L5-F4, L5-F9**: Risk matrix reviewed; unverified figures marked or pre-verified; enforcement costs re-estimated if needed.
5. **L5-F5**: Faithfulness claim explicitly attributed to vendor guidance rather than Claude-specific empirical validation.
6. **L5-F6**: Artifact recording timing (real-time vs. end-of-run) stated; post-hoc rationalization risk acknowledged.
7. **L5-F7**: Adaptive thinking gap scoped to reasoning quality (Tier 3), not acts adjudication (Tier 1).
8. **L5-F8**: Compliance API count (260+ vs. ~30) verified or explicitly scoped to documented public surface.

---

**Rendered:** 2026-07-19 (red-lens-r3-L5)  
**Seat:** red-lens-r3-L5
