# Red Lens 6 (Dark-Side Risk) — Round 1 Audit Pass

**Seat:** red-lens-r1-L6  
**Date:** 2026-07-19  
**Scope:** Failure modes, dark-side risks, weak evidence, hidden assumptions in blue's report.

---

## Summary

Blue's report establishes that Claude Code transcripts do not capture readable reasoning content (5,754 thinking blocks, zero non-empty). The headline is **sound in its observed scope** but rests on **untested assumptions** about whether settings can bypass the observed default behavior. Blue's mitigation strategy (artifact-based recording) is honest about limits but creates a **self-referential evidence loop** — the proof that artifacts work is an artifact. Three operational risks are accepted without harness discipline: vendor behavior drift, metric-name divergence, and JSONL integrity. Seven findings recorded; two are high-severity, five are medium.

---

## Findings

### L6-F1: Critical test left unrun — showThinkingSummaries bypass (HIGH SEVERITY)

**Location:** Section 2: "The reasoning channel"; Open questions carried past this round, Q1.

Blue's headline — "the lever does not reach us where we need it" — rests on the assumption that the display resolver's non-interactive branch is insurmountable. The single experiment that could overturn this (testing `showThinkingSummaries: true` on non-interactive sessions) remains untested.

Blue declined to run it citing semantic consent (global state mutation), which is correct. But the result is that **the negative finding rests on unverified code-path analysis**, not empirical behavior. Blue notes this is "the single experiment that could overturn the headline" but does not test it.

**Risk:** If the setting actually works across session types, the entire finding flips. Probability: low (the code path is plausible), consequence: high (headline becomes "thinking *is* available under configuration, not entirely unavailable").

**Confidence:** Low on the negative claim (code-path reasoning without behavioral validation).

**Recommendation:** Run the experiment with explicit operator consent, or qualify the headline to "on a default-configured install without explicitly enabling showThinkingSummaries."

---

### L6-F2: Mechanism inference incomplete — flag regression unverified (MEDIUM-HIGH SEVERITY)

**Location:** Section 2: "The mechanism, read out of the shipped client"; lines 213–220.

Blue establishes that `tengu_quiet_hollow` (the flag mentioned in issue #32810) is absent from v2.1.215 and concludes "the mechanism moved." But blue does not verify that the *new* mechanism produces the same behavior.

The causal chain is: issue #32810 describes condition X involving flag → blue shows flag is gone → blue assumes mechanism moved → therefore empty blocks are expected. The inference of the moved mechanism is sound, but untested. Did the mechanism that replaces it actually cause the empty blocks blue observed?

**Risk:** The mechanism described in the issue (header + server-side flag) may have been replaced with a fundamentally different redaction path. The observation (empty blocks) is consistent with either the old mechanism or a new one. Without testing whether the setting change has any effect, the mechanism explanation is incomplete.

**Confidence:** Medium-high that there's a causal-chain gap; low-medium that it's consequential (the observation stands; the mechanism explanation is the weak part).

**Recommendation:** Trace the header's role in v2.1.215 empirically, or reframe the finding as "observed: all-empty blocks" rather than "explained by: mechanism X."

---

### L6-F3: Operational discipline missing — "cheap sweep" unsystematized (MEDIUM SEVERITY)

**Location:** Section 9: Risk matrix, row 8 (line 461).

Blue risk-accepts the likelihood of vendor behavior change ("medium" on the scale) because "the sweep is cheap; schedule it, do not engineer around it." This delegates an ongoing validation duty to unspecified future work.

In practice, the local sweep involved 5,754 blocks across 294 files, 287 sessions, leaf extraction. It took hours of focused work. The assumption that re-running is "cheap" conflates the individual operation (shell command: yes, cheap) with the operational discipline (detecting drift, scheduling re-runs, integrating results).

**Risk:** If the vendor changes behavior in v2.1.216, no monitoring detects the drift, and a harness built on v2.1.215's behavior fails silently. The "cheap" mitigation only works if someone runs it.

**Complexity to mitigate:** Medium-high. Requires: version monitoring, re-run scheduling, result integration.

**Recommendation:** Establish a repeatable sweep recipe in the harness; trigger on version-upgrade events, or document version-pinning assumptions in any code that parses transcripts.

---

### L6-F4: Audit log lacks integrity guarantee (MEDIUM SEVERITY)

**Location:** Section 1: "What the transcript actually contains"; lines 156–160.

The JSONL format is "append-only but is not signed or tamper-evident." If a malicious agent or operator gains filesystem access, it can edit prior lines without audit trail.

In an adversarial audit context—where agents are being tested for compliance—unsigned logs create a blindspot: silent evidence tampering.

**Risk:** Medium in practice (requires filesystem access), high in consequence (evidence loss, undetectable).

**Complexity to mitigate:** Low. HMAC-SHA256 per line, appended at write-time.

**Recommendation:** Add cryptographic signatures to JSONL lines, or document the assumption that the local filesystem is trusted. If filesystem-level trust is already assumed, this gap is superseded by the larger security boundary.

---

### L6-F5: Self-referential evidence loop — artifact findings grounded in artifacts (MEDIUM-HIGH SEVERITY)

**Location:** Section 8: "What to do instead"; lines 403–410.

Blue's recommended solution is artifact-based recording (avenue status, manifest rows, closure anchors, friction). The evidence that this works is *this report*, which is an artifact-based record. The findings are grounded in the very method they justify.

Blue acknowledges the limit (line 409): artifacts "buy durability and non-circularity, not sincerity." So blue is transparent that artifacts are self-reports. But the circularity is real: any agent running under this harness will use artifact recording to justify its own findings, creating mutual validation.

**Risk:** High that this pattern recurs. Every agent will corroborate its own findings via its own artifacts. This looks like evidence convergence from the outside but is consensus on method, not on truth.

**Complexity to mitigate:** High. Requires independent external validation (human audit, second agent with different method, automated testing).

**Recommendation:** Explicitly require mixed-method evidence for high-stakes findings (e.g., compliance violations must not rest on agent's own artifacts alone). Treat artifact findings as "corroborated by agent's own record" rather than "verified."

---

### L6-F6: Pinned input contract unvalidated (MEDIUM SEVERITY)

**Location:** Provenance and limitations; lines 600–604.

The run's pinned evidence base (inputs/PINNED.md) names two files that do not exist: `probe-thinking-persistence.md` and `mining-substrate-architecture.md`. The pinned repo HEAD (cacb736) does not match actual HEAD (4baf282).

Blue asserts "no claim rests on either file" but this assertion is unverified — no dependency trace was run to check whether lane reasoning inherited from them.

**Risk:** If those files were supposed to be part of the evidence chain, blue's findings may rest on dropped reasoning. A pinned-input mismatch is a run-start validation failure.

**Complexity to mitigate:** Medium. Requires: run-start validation (do files exist?), dependency tracing if missing.

**Recommendation:** At next run-start, validate that all pinned files exist before synthesis begins. If missing, treat as run-blocker or explicitly scope-cut affected claims.

---

### L6-F7: Metric names drift silently (MEDIUM SEVERITY)

**Location:** Section 4: OpenTelemetry; lines 289–293.

Lane-3's metric enumeration (`claude_code.tokens.input`, `claude_code.cost.total`, `claude_code.tool_decisions.total`) does not match the binary's actual names (`claude_code.token.usage`, `claude_code.cost.usage`, `claude_code.code_edit_tool.decision`).

Blue notes this is a "naming divergence, not capability dispute" and accepts it as version-bound risk. But if anyone built monitoring on lane-3's names, the parser silently breaks—the metric is not found, and failures go undetected.

**Risk:** Medium in practice (requires downstream dependence on lane-3 names), high in consequence (silent failures).

**Complexity to mitigate:** Low. Validate metric names against binary at harness-start.

**Recommendation:** Include a metric-name validation step in any harness using OpenTelemetry parsing. Document names as version-bound.

---

## Gaps Rebutted: Why Blue's Findings Still Stand

- **"Acts aren't directly observable either"** — The report is honest about acts: tool calls, parameters, and results are captured in transcript blocks. The chain is direct (tool_use → tool_result in the JSONL) without reconstruction. This is the strongest data blue has.

- **"Binary extraction can be fooled by runtime construction"** — True, but blue's findings (the beta header exists, showThinkingSummaries exists, redaction map exists) are corroborated by vendor documentation and the issue tracker. The binary method is fallible but over-corroborated on the key claims.

- **"The sweep could be wrong (one machine, one user)"** — True. But blue is explicit about scope ("one default-configured install") and the finding (empty blocks) is a direct measurement, not an inference. Reproducible on the same machine; generalizability is blue's stated limitation, not a hidden one.

- **"OTLP redaction could have configuration I missed"** — The extracted map is hardcoded with no config argument upstream. This is a strong claim, verified against documentation. If a future version adds a config, the finding becomes "was hardcoded in v2.1.215" and the headline adjusts.

---

## What Is Verified to Hold

1. **Binary extraction is methodical.** Blue extracted 19 instrument names, string-present checks, and redaction maps. The method is reproducible; the scope is limited to v2.1.215 on Windows, explicitly stated.

2. **Issue statuses are accurate.** Blue verified #32810, #32997, #52376, #10084 against the GitHub issue tracker. The status corrections (closed/not-planned, duplicate) are factual.

3. **Scope limitations are transparent.** Blue states "absence claim over the documented public surface" and "single-machine binding" and "version binding." These are honest bracketing, not hidden assumptions.

4. **Risk matrix is graded clearly.** Blue assigns likelihood/impact/complexity to each risk and states the disposition (mitigate, accept, risk-accept-with-disclosure). The grading is traceable.

5. **Acts are sound.** Tool sequences, parameters, results, and token usage are captured exhaustively. This is the strongest finding.

---

## Contested Gaps: Three Items for Dispute Resolution

### Contested Severity of L6-F1 (Untested Bypass)

**Blue's position:** The decline to test is defensible (semantic consent around global state). The finding is carried as an open question.

**Red's position:** It should be elevated from open question to high-severity finding if the headline is to rest on it.

**Reconciliation:** If the headline is rephrased to "on default configuration, thinking is empty," the finding drops in severity. If the headline stands as "thinking is unavailable," the test is prerequisite.

### Contested Responsibility for L6-F3 (Sweep Discipline)

**Blue's position:** The sweep is cheap; operators can schedule it. This is a process issue, not a technical one.

**Red's position:** Delegating to unspecified future work is an operational gap, not a mitigated risk.

**Reconciliation:** A repeatable recipe in code (bash script or tool verb) that can be run on every version-upgrade event clarifies the responsibility and cost.

### Contested Significance of L6-F5 (Self-Referential Loop)

**Blue's position:** Artifact recording is honest about limits and better than reversed-engineered thinking. Yes, it's self-report, but durable and intentional.

**Red's position:** The proof-of-concept is itself an artifact, which is circular.

**Reconciliation:** The finding is that artifact-based findings need corroboration by independent method (human audit, second agent, tests). This is not a show-stopper; it's a design constraint.

---

## Risk-Acceptance Audit: What Blue Is Choosing to Live With

| Risk | Mitigated | Accepted | Status |
|---|---|---|---|
| Harness parses empty blocks, misinterprets silence as "no reasoning" | ✓ | | Stated mitigation: assert non-empty before trusting |
| JSONL field/shape change breaks parser | | ✓ | Read-only forensics accepted; durable paths (OTLP) preferred |
| Silent tool-result truncation corrupts finding | | ✓ | Accepted with disclosure in tier grading |
| Metric names drift between versions | | ✓ | Accepted; re-enumerate per version |
| Reasoning-quality claim on Tier-3 evidence | ✓ | | Stated mitigation: tier label mandatory |
| Artifact self-reports are post-hoc | | ✓ | Accepted; disclose the limit; require corroboration for high-stakes |
| Vendor changes behavior without release | | ✓ | Accepted; sweep is cheap, schedule it (gap: no harness yet) |

Blue's risk-acceptance is **honest and graded**. The outstanding question is whether the operational discipline to execute the accepted mitigations (re-enumerate, re-sweep, corroborate) will exist when needed.

---

## Recommendation: Verdict

**CONDITIONAL PASS.** 

The report's findings about acts and artifacts are sound. The headline about reasoning unavailability is **plausible but rests on untested assumptions**. The operational plan (artifact recording) is well-grounded but creates a self-referential validation loop that needs external corroboration for high-stakes claims.

**Conditions:**
1. Rerun L6-F1 (showThinkingSummaries bypass test) with operator consent, or qualify the headline.
2. Establish repeatable harness discipline for L6-F3 (sweep scheduling, metric validation).
3. Document the self-referential risk of artifact-based findings (L6-F5) and require mixed-method evidence for compliance-level claims.
4. Validate pinned inputs at run-start for future rounds (L6-F6).

If these conditions are met or explicitly risk-accepted by the operator, the report graduates to **VERIFIED.**
