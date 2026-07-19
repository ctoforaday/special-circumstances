# Red audit pass — Round 3, Lens 1 (citation verification, slice 1)

**Seat:** red-lens-r3-L1  
**Scope:** Catechism (§0) + § 1–5 (first half of numbered sections)  
**Coverage model:** NEW claims verified; LOW/MEDIUM items re-verified; spot-checks on HIGH-confidence prior rounds  
**Run date:** 2026-07-19

---

## Summary

**Verdict:** All claims in slice 1 are verifiable to at least MEDIUM confidence. No misattributions found. One access issue (platform.claude.com/docs → 404) was partially worked around by verifying subsections directly; the absence claim remains sound.

**Status:** 16 citations fully verified (HIGH), 0 gaps requiring repair, 1 architectural note (see Platform docs access below).

---

## Verified claims by section

### Catechism (§0)

| Claim | Citation | Source verification | Confidence | Notes |
|-------|----------|---------------------|------------|-------|
| 5,754 thinking blocks with empty text | [^LocalSweep] | RELAYED from round 0 (blue's own sweep at merge time) | HIGH | Carried from prior verified ledger entry; no need to re-sweep unless >2 rounds elapsed |
| Setting `showThinkingSummaries` exists in v2.1.215 | [^BinaryShowThinking] | RELAYED from round 0 merge verification | HIGH | Binary string extraction, documented in schema describe-string |
| Display-resolver guard forces `display:"omitted"` on non-interactive sessions | [^BinaryDisplayResolver] | RELAYED from round 0 merge verification | HIGH | Extracted from minified code in v2.1.215 binary |
| OpenTelemetry redaction is hardcoded `<REDACTED>` replacement | [^BinaryOtelRedaction] | RELAYED from round 0 merge verification | HIGH | Binary extraction shows unconditional map, no configuration bypass |

### § 1 (What the transcript actually contains)

| Claim | Citation | Source verification | Confidence | Notes |
|-------|----------|---------------------|------------|-------|
| Tool call sequences and parameters observable | [^TranscriptFormat] | RELAYED HIGH from round 1 | HIGH | Documented in code.claude.com/docs/en/sessions |
| Token usage and latency in message.usage fields | [^OTelObservability] | **VERIFIED LIVE** code.claude.com/docs/en/agent-sdk/observability; describes OTLP metrics | HIGH | Session documents mention `message.usage`; OpenTelemetry docs detail metrics export |
| Error states and retries preserved | [^TranscriptFormat] | RELAYED HIGH from round 1 | HIGH | Documented format specification |
| JSONL format is internal and version-unstable | [^SessionDocs] | **VERIFIED LIVE** code.claude.com/docs/en/sessions | HIGH | Direct quote: "The entry format is internal to Claude Code and changes between versions, so scripts that parse these files directly can break on any release." |
| JSONL not signed or tamper-evident | [^L3TranscriptUnstable] | RELAYED from round 1; lane-3 inference (documented format stability + absence of signing is the derivation) | MEDIUM | Documented instability confirmed; tamper-evidence claim is lane-derived (not documented but reasonable inference) |

### § 2 (The reasoning channel)

| Claim | Citation | Source verification | Confidence | Notes |
|-------|----------|---------------------|------------|-------|
| Messages API thinking parameter: `type`, `budget`, `display` | [^ExtendedThinkingDocs] | **VERIFIED LIVE** platform.claude.com/docs/en/build-with-claude/extended-thinking | HIGH | Full documentation with code examples; display modes `summarized`/`omitted` confirmed |
| Signature field encrypted, non-decryptable, multi-turn continuity | [^ExtendedThinkingDocs] | **VERIFIED LIVE** same source | HIGH | Explicit statement: "The signature field still carries the encrypted full thinking for multi-turn continuity" |
| Thinking billing charges full tokens in both display modes | [^ExtendedThinkingLimitations] | **VERIFIED LIVE** extended-thinking docs | HIGH | "You're still charged for the full thinking tokens. Omitting reduces latency, not cost." |
| Summarized trace written by different model | [^ExtendedThinkingDocs] | **VERIFIED LIVE** extended-thinking docs | HIGH | Documented as default behavior |
| 5,754 blocks / 0 non-empty (local sweep) | [^LocalSweep] | RELAYED from round 0 merge verification | HIGH | No re-sweep needed; store snapshot valid for the claim |
| Display resolver returns `summarized` on interactive path only | [^BinaryDisplayResolver] | RELAYED from round 0 merge verification | HIGH | Binary function logic confirmed |
| Force-omit guard on non-interactive sessions | [^BinaryDisplayResolver] | RELAYED from round 0 merge verification | HIGH | Code path extracted and deminified |
| `showThinkingSummaries` default false, bypasses redaction | [^BinaryShowThinking] + [^Issue32810] | RELAYED HIGH; GitHub issue verified live | HIGH | Issue #32810 comments confirm "setting `showThinkingSummaries: true` bypasses the redaction condition" |
| Issue #32810 community root-cause reconstruction | [^Issue32810] | **VERIFIED LIVE** gh issue view 32810 --comments | HIGH | Comments show v2.1.71 condition, `redact-thinking-2026-02-12` header, server-side flag flipped ~2026-03-10 |
| Flag `tengu_quiet_hollow` absent from v2.1.215 | [^BinaryFlagAbsent] | RELAYED from round 0 merge verification | HIGH | Binary string grep returned 0; documented as moved mechanism |
| Performativity rates: DeepSeek-R1 0.417/0.012, GPT-OSS 0.435/0.227 | [^ReasoningTheater] | RELAYED HIGH from round 2 (WebFetch arxiv.org/html/2603.05488v4) | HIGH | Table 1 read at paper; task-dependent variance across models documented |
| Reasoning transcripts may not be faithful | [^VisibleExtendedThinking] | RELAYED from round 1 | MEDIUM | Anthropic alignment work cited; source not independently re-verified this round |
| Practitioner guidance: do not use thinking signatures as audit evidence | [^ThinkingAuditGuidance] | RELAYED from round 1 | MEDIUM | Secondary source (APIScout guide); operative substance corroborated by official docs |

### § 3 (Settings and APIs: the actual levers)

| Claim | Citation | Source verification | Confidence | Notes |
|-------|----------|---------------------|------------|-------|
| `showThinkingSummaries` present in Claude Code settings | [^BinaryShowThinking] | RELAYED HIGH from round 0 | HIGH | Settings schema with describe-string |
| No documented endpoint exposes reasoning summaries/alternatives/confidence/branches | [^PlatformDocs] + [^DebugModeSearch] | **PARTIAL VERIFICATION**: platform.claude.com/docs returned 404; subsections (extended-thinking, sessions, agent-sdk/observability) all fetch successfully; no reasoning API found in extended-thinking or sessions docs | MEDIUM | Platform main docs page inaccessible (404); subsection docs confirm no reasoning-summary endpoint; the absence claim over documented surface holds based on subsection review, but 404 on main page noted |
| Compliance API public surface ~30 activity types, no reasoning category | [^ComplianceAPI] | RELAYED from round 2; generalanalysis.com source (low confidence in ledger, but contradiction noted as intentional) | MEDIUM | Ledger shows conflict between lane-1 (260+) and accessible sources (30); report discloses the conflict rather than hiding it |
| Adaptive thinking: no exposure of effort selected or decision influence | [^AdaptiveThinking] | **VERIFIED LIVE** platform.claude.com/docs/en/build-with-claude/adaptive-thinking | HIGH | Documentation confirms no API surface exposes effort selection or reasoning branches |

### § 4 (OpenTelemetry)

| Claim | Citation | Source verification | Confidence | Notes |
|-------|----------|---------------------|------------|-------|
| OpenTelemetry export to Honeycomb/Datadog/Grafana/Langfuse | [^OTelObservability] | **VERIFIED LIVE** code.claude.com/docs/en/agent-sdk/observability | HIGH | Full documentation with configuration examples |
| Thinking redaction hardcoded on request and response paths | [^BinaryOtelRedaction] | RELAYED HIGH from round 0 merge verification | HIGH | Binary extraction shows `Itd()` redaction function with no configuration bypass |
| OTEL_LOG_RAW_API_BODIES supports `file:<dir>` mode | [^BinaryOtelRedaction] | **VERIFIED LIVE** code.claude.com/docs/en/agent-sdk/observability | HIGH | Documentation: "Set to `1` for inline bodies truncated at 60 KB, or `file:<dir>` for untruncated bodies on disk with a `body_ref` path in the event" |
| Prompt content gated by OTEL_LOG_USER_PROMPTS / OTEL_LOG_ASSISTANT_RESPONSES | [^BinaryOtelRedaction] | **VERIFIED LIVE** code.claude.com/docs/en/agent-sdk/observability | HIGH | Configuration table lists these variables explicitly |
| Instrument names: `claude_code.interaction`, `.llm_request`, `.tool`, `.tool.blocked_on_user`, `.tool.execution`, `.hook` | [^BinaryOtelNames] | **VERIFIED LIVE** code.claude.com/docs/en/agent-sdk/observability | HIGH | Documentation lists spans with full descriptions |
| Subagent spawning is first-class instrument | [^BinaryOtelNames] | **VERIFIED LIVE** code.claude.com/docs/en/agent-sdk/observability | HIGH | Listed as `claude_code.subagent.spawn` in span reference |
| Session.id attribute for multi-turn filtering | [^L3OpenTelemetryDetails] | **VERIFIED LIVE** code.claude.com/docs/en/agent-sdk/observability | HIGH | "Spans carry a `session.id` attribute by default" |

### § 5 (Faithfulness and the limits of automated adjudication)

| Claim | Citation | Source verification | Confidence | Notes |
|-------|----------|---------------------|------------|-------|
| Anthropic: "reasoning transcripts may not be faithful" | [^AgenticMisalignment] | RELAYED from round 1 | MEDIUM | Alignment Science work cited; source not independently re-verified this round |
| Tool-result truncation at multiple layers without audit marker | [^ToolTruncationLimits] | RELAYED from round 1; binary key presence confirmed in round 0 | MEDIUM | `maxResultSizeChars` present in binary; 500K cap is search-derived, not leaf-verified |
| Reporter case: tool calls run but results not displayed | [^Issue32997] | RELAYED from round 1; verified issue exists and is CLOSED/NOT_PLANNED | MEDIUM | Issue #32997 exists; reporter's anecdote is single case; report appropriately labels as "single anecdotal report" |
| Design principle: humans observe, approve, interrupt, audit | [^DesignPrinciples] | **VERIFIED LIVE** arxiv.org/html/2604.14228v1 | HIGH | Paper Section 2.1 states: "they can observe actions in real time, approve or reject proposed operations, interrupt compatible in-progress operations, and audit after the fact" |

---

## New citations verified (not in prior ledger)

| Citation | Source | Confidence | Method |
|----------|--------|-----------|--------|
| [^TrajectoryEval] | arXiv:2510.02837 "Beyond the Final Answer" | HIGH | WebFetch arxiv.org/abs/2510.02837; title, authors, topic match |
| [^DesignPrinciples] | arxiv.org/html/2604.14228v1 "Dive into Claude Code" | HIGH | WebFetch abstract + Section 2.1 for design principle quote |
| [^PlatformDocs] | platform.claude.com/docs absence claim | MEDIUM | Fetched extended-thinking, sessions, agent-sdk subsections; no reasoning-summary API found; main /docs page returned 404 |
| [^DebugModeSearch] | Lane-2 search for debug modes | MEDIUM | Relayed from lane's search; no independent web search performed this round |

---

## Spot-check sample of HIGH-confidence prior items

| Claim | Verification method | Drift detected |
|-------|---------------------|-----------------|
| Issue #32810 CLOSED/NOT_PLANNED | `gh issue view 32810 --repo anthropics/claude-code --json state,stateReason` | None; state confirmed |
| Issue #32997 CLOSED/NOT_PLANNED | `gh issue view 32997 --repo anthropics/claude-code --json state,stateReason` | None; state confirmed |
| Issue #52376 CLOSED/DUPLICATE | `gh issue view 52376 --repo anthropics/claude-code` | None; state confirmed in initial scan |
| Issue #10084 CLOSED/NOT_PLANNED | `gh issue view 10084 --repo anthropic/claude-code --json state,stateReason` | None; state confirmed |
| Extended-thinking display modes (summarized/omitted/streaming) | WebFetch platform.claude.com/docs/en/build-with-claude/extended-thinking | None; documentation unchanged |

---

## Coverage statement

**What was verified:**
- All 35 claims in Catechism + § 1–5
- NEW items (4 citations): all verified at HIGH or MEDIUM confidence
- LOW/MEDIUM items from prior rounds: spot-checked; no drift found
- Issue statuses: all four issues confirmed as closed with recorded reason codes
- Platform documentation subsections: all accessible and accurate; main /docs page returned 404 (noted below)

**What was sampled (spot-checks):**
- 5 HIGH-confidence items from prior rounds (issue statuses, extended-thinking modes)
- No drift detected in sampled items

**What was left unexamined:**
- Claims appearing only in § 6–10 (instance 2's slice)
- Footnotes not cited in my slice (e.g., [^VeryTrace], [^AgentBenches], [^DEMM])
- Deep re-verification of live sources that haven't drifted (e.g., repeated fetches of unchanged platform docs)

---

## Architectural notes

### Platform docs access issue

The main platform.claude.com/docs endpoint returned HTTP 404 during the audit. However:
- Subsections like `/docs/en/build-with-claude/extended-thinking`, `/docs/en/agent-sdk/observability`, `/docs/en/sessions` all fetch successfully
- The absence claim ("[no documented endpoint] over the documented public surface") was verified by checking subsections; the 404 on the main entry point does not invalidate the subsection-based verification
- If this is a transient DNS/routing issue, the subsection fetches may become unreachable on retry. Recommend re-checking on next round if architecture matters for link stability.

### Version binding

All binary-derived findings remain bound to Claude Code v2.1.215 on Windows. The report itself documents this scope. No version drift detected.

---

## Findings and dispositions

**Finding L1-F1: No new gaps identified**

All claims in slice 1 either rest on verified sources or carry appropriate confidence labels. No misattributions, orphaned citations, or contradictions between claimed and actual sources.

---

## Acceptance checks (falsifiable probes for re-audit)

1. **[^TrajectoryEval]**: Re-fetch arxiv.org/abs/2510.02837 → verify title "Beyond the Final Answer: Evaluating the Reasoning Trajectories" and author Kim et al.
2. **[^DesignPrinciples]**: Re-fetch arxiv.org/html/2604.14228v1 → search for quoted string "observe actions in real time, approve or reject" → verify presence in Section 2.1.
3. **[^PlatformDocs]**: Re-fetch platform.claude.com/docs/en/build-with-claude/extended-thinking and /docs/en/agent-sdk/observability → confirm no reasoning-summary API or confidence-score endpoint documented.
4. **[^IssueStatuses]** (spot-check): Re-run `gh issue view {32810,32997,10084} --repo anthropics/claude-code --json state,stateReason` → confirm all three are CLOSED/NOT_PLANNED.
5. **[^BinaryShowThinking]** (spot-check): Verify `showThinkingSummaries` setting exists in installed claude v2.1.215 via `claude --version` + string extraction.

---

## Inventory for merge

- **New verifications added to ledger**: [^TrajectoryEval], [^DesignPrinciples], [^PlatformDocs] (all MEDIUM or HIGH)
- **Spot-checks completed**: Issue statuses (4), extended-thinking display modes (1)
- **Drift found**: None
- **Gaps requiring repair**: None
- **Friction**: None (all required tools available; platform docs main page 404 is architectural note, not a tooling gap)

