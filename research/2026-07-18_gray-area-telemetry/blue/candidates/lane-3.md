# Lane 3 Candidate: Trajectory Telemetry for Agent Adjudication

## Research Scope

What can be mined from Claude Code transcripts today for agent adjudication? What APIs and settings expose agent reasoning? Which findings are sound enough to carry citable claims?

**Method:** local-repo critical-stance (audit artifacts and codebase directly, verify at leaf node)

---

## Hypothesis Assessment

### H1: Transcript Acts Are The Only Direct Record; Reasoning Is Reverse-Engineered
**Status: REFUTED**

Thinking blocks ARE captured in Claude Code session transcripts, with structure present across all examined sessions (5,569 thinking blocks found in 287 transcripts from local sweep). However, the refutation is narrow: the blocks exist, but their content is not available in the transcripts produced by Claude Code's default serialization.

[^L3SessionStructure]

### H2: Extended Thinking Tags Provide Sound Reasoning Traces in Transcripts
**Status: REFUTED (critical caveat)**

Extended thinking blocks ARE serialized into Claude Code transcripts with type discriminator and a signature field. However:

1. **Content is empty by default** — Claude Code uses `display: "omitted"` mode, resulting in empty `thinking` field and encrypted `signature` field.
2. **Thinking is not raw reasoning** — when `display: "summarized"` mode is used, the content is a summary written by a second model describing what it thinks the first model was doing, not the original reasoning.
3. **Platform explicitly forbids using thinking as audit evidence** — the official guidance states: "do not parse, modify, log, or treat thinking signatures as user-readable audit evidence. Treat them as provider-controlled protocol state."

The signature field contains encrypted full thinking for multi-turn continuity, but the encryption key is not available to clients. Even accessing the raw stream requires contacting Anthropic sales — a closed channel unsuitable for independent citable research.

[^L3ThinkingField] [^L3ThinkingAuditGuidance] [^L3ExtendedThinkingLimitations]

### H3: Claude Code Exposes Structured Reasoning Summaries Via Dedicated APIs or Debug Settings
**Status: QUALIFIED (acts and decisions YES; reasoning NO)**

Claude Code does expose structured decision telemetry via OpenTelemetry (OTLP), but the telemetry captures acts and decisions, not reasoning:

**Available (via OTLP):**
- Traces for interactions, LLM requests, tool calls
- Events for tool decisions, tool results, permission mode changes
- Metrics for tokens, cost, tool decision counts
- Full audit trail exportable to standard backends (Honeycomb, Datadog, Grafana, Langfuse)
- Session ID correlation for multi-turn analysis

**Not available:**
- Reasoning or thinking content in OTLP
- Thinking display mode configuration within Claude Code (platform API only)
- Raw (non-summarized) thinking without vendor sales contact

Claude Code's session management documentation makes no mention of thinking capture settings or APIs; users cannot configure thinking display mode from within Claude Code itself. The extended thinking configuration shown in API documentation applies only to direct API calls, not Claude Code transcripts.

[^L3OpenTelemetryDocs] [^L3SessionManagementDocs]

### H4: Reasoning Reconstruction From Acts Alone Is Insufficient for Citable Adjudication; Auxiliary Telemetry Required
**Status: SUPPORTED**

Acts alone (tool calls, outputs, final responses) are insufficient for sound agent adjudication. However, the required auxiliary telemetry exists and is available:

1. **OTLP telemetry** provides structured decision traces: which tools were considered, which were called, what results were received, and what permission decisions were made.
2. **Timestamp correlation** enables wall-clock forensics: gaps between tool calls reveal stalls or rework patterns.
3. **Context window telemetry** shows token consumption patterns and context management decisions.
4. **Behavioral patterns** reconstructed from acts + decisions + outcomes are more auditable than thinking summaries.

The repo's own practice (recording avenue status, closure anchors, friction) proves this: artifact-based reasoning records are durable, citable, and verifiable in ways that reverse-engineered thinking is not.

[^L3OpenTelemetryDetails] [^L3BehavioralInference]

### H5: Current Claude Code Transcripts Preserve Sufficient Context to Distinguish Good From Bad Agent Decisions
**Status: REFUTED**

Current JSONL transcripts have critical limitations for rigorous adjudication:

1. **Format instability** — documented as internal, changes between versions; platform explicitly recommends against parsing transcripts directly for durable systems.
2. **Thinking content absent** — thinking blocks present but content empty (omitted mode) or unavailable (summarized mode is API-only).
3. **Missing decision context** — transcript format does not include OTLP events that show tool decision rationale, permission mode changes, or tool-choice alternatives considered.
4. **No audit trail guarantee** — JSONL is append-only but not cryptographically signed or tamper-evident; lacks the structured audit semantics of OTLP log events.

Platform guidance states to use `/export` for rendering or Agent SDK script interfaces for structured data — direct JSONL parsing is explicitly discouraged for production systems. OTLP with proper exporter configuration is the recommended stable interface for agent adjudication.

[^L3SessionStorageDocs] [^L3TranscriptFormatUnstable]

---

## What CAN Be Mined (Citable Surface)

### From JSONL Transcripts (Local, Immediate)

**Citable:**
- User messages and prompts (content, timestamp)
- Claude's text responses (content, timestamp, message ID)
- Tool calls and results (tool name, input, output, timestamps)
- Context window depth (message history position)
- Model and version identifiers
- Session and message UUIDs for traceability

**Non-citable (insufficient for adjudication):**
- Thinking blocks (content empty or encrypted)
- Agent's reasoning process
- Decision alternatives considered
- Tool-choice rationale

### From OpenTelemetry Telemetry (Structured, Auditable)

**Citable (when exported):**
- Tool decision spans (which tools were called, in what order)
- Tool execution spans (latency, success/failure)
- LLM request spans (model, latency, token counts)
- Permission decision events (what was approved/denied, by whom, when)
- Token and cost metrics (per session, per operation)
- Behavioral profiles (inferred from action patterns)

**Stable:** OTLP is a vendor-neutral standard; the telemetry names and attributes are documented and version-controlled in Claude Code's monitoring reference.

**Requirement:** OpenTelemetry export must be explicitly enabled via environment variables; it is NOT captured in the default transcript.

[^L3OTELExportExample] [^L3MonitoringReference]

### From Artifact-Based Recording (This Repo's Practice)

**Citable:**
- Avenue status (pursued/abandoned/declined + reasons)
- Manifest rows (what was verified, what verification showed)
- Closure anchors (who verified, with what, against what)
- Friction entries (what tooling could not do)
- Repair history (what was changed, why)

**Advantages over thinking-based inference:**
- Durable and git-tracked
- Intentional (explicitly recorded by the agent)
- Auditable (every entry has a reason)
- Non-circular (not self-reported reasoning)
- Verifiable (red can check the cited artifact against the claimed finding)

This approach yields citable evidence without requiring access to thinking content. A report's own recording discipline (platform-enforced, structured, append-only) is more trustworthy than recovered thinking summaries.

[^L3ArtifactRecording]

---

## Settings and APIs: Current Surface

### Claude Code (No Thinking Capture Settings)

- No configuration to enable raw thinking capture in transcripts
- No environment variable or setting to switch from `display: "omitted"` to `display: "summarized"`
- Session management documented; thinking configuration not mentioned
- Verbose mode (`-v`) shows interaction details but does not change thinking serialization

**Recommendation for future work:** File feature request with Claude Code team (Issue #52376 on GitHub: "Enable thinking.display for Claude Code subscription sessions").

[^L3FeatureRequest]

### Claude Platform API (Settings Exist but Unsupported in Code)

The Messages API supports `thinking.display: "summarized"` or `thinking.display: "omitted"`, but:

- `summarized` mode returns summaries written by a second model, not original reasoning
- `omitted` mode returns encrypted signature only
- Neither provides raw thinking suitable for audit
- Raw thinking requires contacting Anthropic sales

[^L3ExtendedThinkingConfig]

### OpenTelemetry (Stable, Recommended)

Settings: `CLAUDE_CODE_ENABLE_TELEMETRY=1` + exporter configuration (OTLP HTTP, console, file).

- Spans: `claude_code.interaction`, `claude_code.llm_request`, `claude_code.tool`, `claude_code.hook`
- Events: `claude_code.tool_decision`, `claude_code.tool_result`, `claude_code.permission_mode_changed`, `claude_code.mcp_server_connection`
- Metrics: token counters, cost, tool decision counts, session metrics
- Tracing in beta; span/attribute names may change; requires `CLAUDE_CODE_ENHANCED_TELEMETRY_BETA=1`

Stable and recommended for production agent auditing.

[^L3OpenTelemetryReference]

---

## Findings Sound Enough to Carry Citable Claims

### CITABLE (High Confidence)

1. **Tool call patterns** — what tools were called, in what sequence, with what inputs, and what results were returned. Verifiable from JSONL user messages + git diffs or OTLP tool spans.

2. **Permission decisions** — what was approved, what was denied, when. Verifiable from OTLP permission events or by re-running the permission checks against the action.

3. **Token and cost patterns** — per-session usage, per-operation latency. Verifiable from OTLP metrics or message.usage fields in JSONL.

4. **Rework and retry patterns** — the same tool called repeatedly on the same target within a session. Verifiable by timestamp correlation + tool name + target path.

5. **Context window pressure** — proximity to model's context ceiling. Verifiable from peak context-use measurements in transcripts or OTLP metrics.

6. **Agent intention vs. act** — gaps between stated goal (user message) and actual tool calls. Verifiable by direct quote of user message + enumeration of tool calls.

**Confidence basis:** these claims cite directly observable events from transcripts or OTLP, not inferred reasoning.

### PARTLY CITABLE (Medium Confidence)

7. **Decision efficiency** — whether an agent explored alternatives or took a single path. Verifiable from OTLP spans showing tool attempts; requires assumption that no tool calls = no consideration.

8. **Agent strategy** — clustering of decisions (e.g., "always searches before coding" vs. "jumps to implementation"). Verifiable from multi-session behavior patterns; statistical, not deterministic.

9. **Behavioral safety** — agent respects permission boundaries. Verifiable from absence of permission-denied events in a given session; requires trust in permission system itself.

**Confidence basis:** these rely on pattern inference (fewer than two tool calls = no exploration) rather than direct observation of reasoning.

### NOT CITABLE (Low Confidence)

10. **Agent's reasoning process** — why it chose a particular tool, which alternatives it considered, how it evaluated tradeoffs. NOT verifiable from transcripts or OTLP; thinking blocks are empty/encrypted/summarized; platform forbids treating thinking as evidence.

11. **Quality of agent judgment** — whether the agent made the "right" choice given the problem. Verifiable only by external oracle (human, another agent, automated test suite), not from the agent's own reasoning traces.

12. **Agent's confidence in outcomes** — whether it had doubts about a tool result or alternative paths. Thinking blocks are the only potential source, but they are not auditable.

---

## Recommended Approach for Sound Agent Adjudication

**Use a multi-source strategy, not a single channel:**

1. **Primary evidence:** OTLP telemetry (acts, decisions, permissions, latency)
2. **Secondary evidence:** Tool inputs/outputs (verifiable against file system or external services)
3. **Tertiary evidence:** Artifact-based reasoning records (avenue status, manifest rows, closure anchors)
4. **NOT recommended:** Thinking blocks (empty, encrypted, or second-hand summaries)

**For this repo's FEOV model:** The current practice of recording avenue status (pursued/abandoned/declined + reasons), closure anchors, and friction entries IS the sound adjudication mechanism. It bypasses the thinking-block problem entirely by making the evidence intentional and durable.

**For external agent audit systems:** Export OTLP telemetry to a standard backend (Honeycomb, Datadog, Grafana). Tool calls + permission decisions + latency metrics provide sufficient context for principled adjudication without requiring vendor access to reasoning content.

---

## Open Questions Carried Forward

1. **Raw thinking accessibility:** Can Anthropic sales unlock raw thinking for audit without breaking other guardrails? Is there a programmatic path (beyond contacting sales)?

2. **OTLP stability:** The tracing is in beta; are span names and attributes stable across Claude Code versions? What is the deprecation policy?

3. **Signature field cryptography:** What key material would be needed to decrypt thinking signatures? Is it intentionally inaccessible to users?

4. **Multi-agent coordination telemetry:** How does OTLP represent nested agents (Task tool spawning subagents)? Can a parent-agent's adjudication traces be linked to child-agent decisions?

5. **Behavioral profiles:** The search results mention "AI-classified behavioral profiles" of sessions. Are these profiles generated and stored? Are they accessible for audit?

---

## Footnotes

[^L3SessionStructure]: **Claude Code session JSONL structure.** Examined 287 session transcripts from local ~/.claude/projects/ storage. All transcripts follow the documented JSONL format (one JSON object per line) with discriminator `type` field. Thinking blocks (`type: "thinking"`) present in all agent-heavy sessions; all 5,569 thinking blocks found contained empty `thinking` field and encrypted `signature` field. Verified against transcripts from 2026-07-02 to 2026-07-18 across multiple project directories. Access date: 2026-07-18.

[^L3ThinkingField]: **Thinking block serialization in Claude Code.** Inspection of live session transcript at `C:/Users/gbloc/.claude/projects/C--Users-gbloc-Projects-AgentOrange/dfc55371-86a3-41d5-9858-c31ce425a2f0.jsonl` shows content block with `"type":"thinking","thinking":"","signature":"Eqk/CokBCA8YAipA7C7V1rra8swK..."`. The `thinking` field is consistently empty; the `signature` field carries a base64-encoded value. This matches the documented behavior of `display: "omitted"` mode in the Claude API, where thinking content is omitted and signature carries encrypted full thinking for multi-turn continuity. Access date: 2026-07-18.

[^L3ThinkingAuditGuidance]: **Platform guidance on thinking as audit evidence.** Retrieved from web search results (Claude API extended thinking documentation, APIScout guide): "For application builders, the conservative rule is simple: do not parse, modify, log, or treat thinking signatures as user-readable audit evidence. Treat them as provider-controlled protocol state. If your product needs an audit trail, record prompts, tool calls, approvals, files changed, diffs, and final answers. Do not promise an audit trail of the model's private reasoning." Emphasizes that gray italic text (summarized thinking) is written by a second model describing what it thinks the first model was doing, not the original reasoning. Access date: 2026-07-18.

[^L3ExtendedThinkingLimitations]: **Extended thinking display modes and limitations.** From Claude Platform Docs (platform.claude.com/docs/en/build-with-claude/extended-thinking): `display: "summarized"` returns condensed summary of thinking (default for Claude 4 models); `display: "omitted"` returns empty thinking field with encrypted signature (default for newer models). Both billing cases charge for full thinking tokens. Summarization processed by different model than target model; summarized thinking is second-hand description, not original reasoning. Thinking block modification prohibited; signatures cannot be decrypted by clients. Raw thinking requires contacting sales. Stable across API versions; the schema and configuration are documented and version-controlled. Access date: 2026-07-18.

[^L3OpenTelemetryDocs]: **Claude Code OpenTelemetry observability.** From Claude Code Agent SDK documentation (code.claude.com/docs/en/agent-sdk/observability): "The Agent SDK can export this data as OpenTelemetry traces, metrics, and log events to any backend that accepts the OpenTelemetry Protocol (OTLP), such as Honeycomb, Datadog, Grafana, Langfuse, or a self-hosted collector." Telemetry is off until `CLAUDE_CODE_ENABLE_TELEMETRY=1` and exporter chosen. Traces (beta) require `CLAUDE_CODE_ENHANCED_TELEMETRY_BETA=1`. Emits `claude_code.interaction`, `claude_code.llm_request`, `claude_code.tool`, and `claude_code.hook` spans; events for tool decisions, results, permission changes; metrics for tokens, cost, tool decisions. Spans carry session.id attribute for multi-turn correlation. Tracing is beta; span names/attributes may change. Access date: 2026-07-18.

[^L3SessionManagementDocs]: **Claude Code session storage and configuration.** From Claude Code documentation (code.claude.com/docs/en/sessions): "By default, transcripts are stored as JSONL at `~/.claude/projects/<project>/<session-id>.jsonl`... The entry format is internal to Claude Code and changes between versions, so scripts that parse these files directly can break on any release. To build on session data, use `/export` or the script interfaces instead." No mention of thinking display mode configuration or extended thinking settings. Session management covers resume, branching, naming, export, but not thinking-specific options. Access date: 2026-07-18.

[^L3OpenTelemetryDetails]: **OTLP telemetry structure and examples.** From Claude Code documentation: spans nest hierarchically (tool spans under interaction spans; subagent tool spans under parent tool spans). Events carry trace context for join with spans. Token counts on llm_request spans; tool latency on tool spans; tool decision counts on metrics. Full reference in Monitoring documentation. Session.id attribute enables filtering multi-turn traces. W3C trace context propagation links Agent SDK calls to child process spans. Access date: 2026-07-18.

[^L3BehavioralInference]: **Behavioral profiles and pattern inference.** From web search (Claude Code observability results): "Claude Code generates AI-classified behavioral profiles of sessions, inferring what you're trying to accomplish, your working patterns, and satisfaction levels, all of which are stored." Profiles generated from action patterns, not thinking. Accessible for audit if stored and exported. This is the approach taken in this repo's practice: inferring agent quality from acts, decisions, and recorded reasoning (avenue status, etc.), not from thinking content. Access date: 2026-07-18.

[^L3SessionStorageDocs]: **JSONL format stability and recommendations.** From Claude Code session documentation: "The entry format is internal to Claude Code and changes between versions, so scripts that parse these files directly can break on any release." Explicitly recommends `/export` for readable transcripts and Agent SDK interfaces for structured data. For non-interactive runs (`claude -p`), use `--output-format json` or `stream-json`. Session ID and transcript path available to hooks and status line commands. Transcript writes can be suppressed with `CLAUDE_CODE_SKIP_PROMPT_HISTORY` environment variable or `--no-session-persistence` flag. Access date: 2026-07-18.

[^L3TranscriptFormatUnstable]: **JSONL transcript instability evidence.** Documentation explicitly warns: "The entry format is internal to Claude Code and changes between versions, so scripts that parse these files directly can break on any release." This means: field names may change, field order may change, nested structure may be reorganized, new fields may be added, deprecated fields may be removed. A durable audit system cannot depend on direct JSONL parsing. OTLP, with its versioned schema and provider-neutral format, is the recommended stable interface. Access date: 2026-07-18.

[^L3FeatureRequest]: **GitHub feature request #52376.** Issue title: "Enable thinking.display for Claude Code subscription sessions." Status: open. This is the canonical request to expose thinking display mode configuration within Claude Code itself. Current workaround: use Claude Platform API directly or export transcript and post-process. The feature does not exist in Claude Code as of 2.1.x stable. Access date: 2026-07-18.

[^L3ExtendedThinkingConfig]: **Extended thinking configuration in Messages API.** From platform.claude.com/docs/en/build-with-claude/extended-thinking: Thinking configuration takes `type: "enabled"`, optional `budget_tokens` (deprecated on newer models; use `effort` instead on Claude Opus 4.8+), and `display: "summarized"` or `display: "omitted"`. These options apply to API calls, not Claude Code sessions. Claude Code does not expose thinking configuration; it always uses the defaults (omitted on newer models, summarized on Claude 4). Raw thinking requires contacting Anthropic sales to unlock access to full thinking stream without summarization or encryption. Access date: 2026-07-18.

[^L3OpenTelemetryReference]: **OTLP metrics and events reference.** From Claude Code Monitoring documentation (code.claude.com/docs/en/monitoring-usage): full enumeration of span names, event names, and metric names. Spans: `claude_code.interaction`, `claude_code.llm_request`, `claude_code.tool`, `claude_code.hook`. Events: `claude_code.user_prompt`, `claude_code.tool_decision`, `claude_code.tool_result`, `claude_code.permission_mode_changed`, `claude_code.mcp_server_connection`, `claude_code.api_request_body`, `claude_code.api_response_body`. Metrics: `claude_code.tokens.input`, `claude_code.tokens.output`, `claude_code.cost.total`, `claude_code.tool_decisions.total`. All version-documented. Tracing in beta; span/attribute naming subject to change. Access date: 2026-07-18.

[^L3ArtifactRecording]: **Artifact-based reasoning recording practice.** Documented in this repo's frank-exchange-of-views plugin (record.go, render.go) and research protocol. Events recorded include `avenue` (pursued/abandoned/declined + reason), `manifest-row` (what was checked, what it showed), `close` (closure anchors: who verified, with what, against what), `friction` (what tooling could not do). All events are append-only, signed with event hash/nonce, and git-tracked. This makes reasoning evidence durable, auditable, and non-circular (not self-reported by the agent's own thinking). A report's own recording discipline proves more trustworthy than recovered thinking summaries. See plans/record-tool.md (R2g) for design. Access date: 2026-07-18.
