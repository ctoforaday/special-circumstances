# Lane 2: Trajectory Telemetry for Agent Adjudication — Primary Literature Survey

## Executive Summary

Claude Code transcripts can be mined for **tool call sequences, token usage, and latency metrics**—sufficient for coarse behavioral inference. However, **reasoning itself is not directly captured**: extended thinking tags are encrypted or summarized, OpenTelemetry redacts thinking content, and APIs expose no structured reasoning summaries (e.g., decision trees, alternatives, confidence scores). **Current telemetry is sound for auditing acts, not reasoning.** Citable findings require explicit layering: reconstruct reasoning from tool sequences as a hypothesis, then verify via independent means (re-execution, formal trace verification frameworks). Solo-transcript adjudication is unsound; auxiliary tracing (evidence chains, semantic provenance) is the load-bearing requirement for citable agent quality judgments.

---

## What Can Be Mined from Claude Code Transcripts Today

### Directly Captured and Reliable

**1. Tool Call Sequences and Parameters**  
Claude Code JSONL transcripts record every tool invocation as a `tool_use` block with name and `input` JSON [^L2TranscriptFormat]. These sequences are auditable:
- Which tools were selected (file reads, writes, Bash commands, web fetches)
- The order of invocations (procedural state)
- Input arguments (what parameters the agent chose)

**Soundness for adjudication:** ★★★★☆  
Tool selection and parameter construction are directly observable. Comparative analysis of two agents' tool sequences can distinguish efficient from wasteful or error-prone patterns [^L2TrajectoryEval]. This is the only reasoning-adjacent data with minimal reconstruction required.

**2. Token Usage and Model Latency**  
OpenTelemetry telemetry and `message.usage` fields in transcripts capture input/output token counts per turn and span duration for each LLM request [^L2OTelObservability]. These are precise operationally.

**Soundness for adjudication:** ★★★★★  
Token counts and latency are facts of execution, not inferences. Comparisons are valid: "Agent A used 2× the tokens to reach the same outcome."

**3. Tool Execution Results**  
`tool_result` blocks contain the output of each tool call: file contents, shell command output, web fetch responses [^L2TranscriptFormat]. Results are unmodified capture of the environment's response.

**Soundness for adjudication:** ★★★★★  
Results are ground truth from the agent's environment. Auditing whether an agent correctly **interpreted** the result is a separate, reasoning-level problem (see below).

**4. Error States and Retries**  
Transcripts preserve failed tool calls, error messages, and retry patterns [^L2TranscriptFormat]. These are directly observable acts.

**Soundness for adjudication:** ★★★☆☆  
Frequency of retries is auditable, but the *reason* for the retry (whether an intelligent recovery or flailing) requires reasoning reconstruction.

### Indirectly Mineable (Reconstruction Required)

**5. Reasoning Patterns from Tool Sequences**  
No reasoning is explicitly logged, but tool call sequences can be reverse-engineered for heuristic signals [^L2TrajectoryEval]:
- Backtracking detection (revisiting the same file multiple times without progress)
- Tool-result consumption (did the agent read and use the tool output it requested?)
- Hypothesis testing (e.g., try-edit-verify-retry patterns)

**Soundness for adjudication:** ★★☆☆☆  
Sequences are observable, but interpretation is speculative. A revisit to the same file could indicate careful verification or confused looping. This is where auxiliary telemetry (explicit reasoning traces, see §III.D) becomes critical.

---

## Settings and APIs That Expose Reasoning Summaries

### 1. Extended Thinking Configuration (Currently the Primary Channel)

**API Surface:**  
Claude's Messages API exposes thinking via the `thinking` parameter in requests:
```json
{
  "thinking": {
    "type": "enabled",
    "budget_tokens": 10000,
    "display": "summarized"  // or "omitted"
  }
}
```
[^L2ExtendedThinkingDocs]

**What Is Returned:**  
Responses include a `thinking` content block with three possible states:
- **`display: "summarized"`** — Returns a text summary of the model's reasoning (not the full thinking, a digest created by Claude itself)
- **`display: "omitted"`** — Returns an empty `thinking` field with a `signature` field (encrypted full thinking for multi-turn continuity)
- **Streaming** — Arrives as `thinking_delta` events during response generation [^L2ExtendedThinkingDocs]

**Critical Limitation:**  
The summarized thinking is **not raw reasoning** — it is Claude's own abstract summary of its thinking process. Anthropic's documentation and research both note: "we don't know for certain that what's in the thought process truly represents what's going on in the model's mind" [^L2VisibleExtendedThinking]. The signature preserves the encrypted full thinking for inference continuity, but is not accessible to adjudicators.

**Soundness for adjudication:** ★★☆☆☆  
Summaries are post-hoc narrations, not transcripts of reasoning. They may elide or misrepresent the actual decision-making. For **citable findings**, summaries require independent verification (e.g., does the tool sequence align with the claimed reasoning?).

### 2. Claude Code JSONL Thinking Blocks (Encrypted by Default)

**What Is Captured:**  
Claude Code's JSONL transcripts include thinking blocks when extended thinking is enabled [^L2TranscriptFormat]:
```json
{
  "type": "assistant",
  "message": {
    "content": [
      {
        "type": "thinking",
        "thinking": "... thinking content ..."
      },
      {
        "type": "text",
        "text": "..."
      }
    ]
  }
}
```

**The Capture Problem:**  
As of February 12, 2026, the `redact-thinking-2026-02-12` header suppresses thinking rendering in the Claude Code UI [^L2WhyHidingThinking]. More critically, on-disk transcripts store thinking as either:
- **Encrypted signature** (~600 characters) — the actual thinking is inaccessible
- **Summarized text** (if `display: summarized` is set) — subject to the same post-hoc narration limits as §II.A.1
- **Empty field** (if `display: omitted`) — no reasoning content at all, only the signature

[^L2TranscriptLimitations]

**Soundness for adjudication:** ★★☆☆☆  
Encrypted thinking cannot be audited. Summarized thinking is unreliable for citable findings. The Feb 2026 change means **most production Claude Code sessions produce transcripts with no readable reasoning at all**.

### 3. OpenTelemetry Telemetry (Reasoning Explicitly Redacted)

**What Is Exposed:**  
Claude Code Agent SDK exports OpenTelemetry traces, metrics, and log events when configured with `CLAUDE_CODE_ENABLE_TELEMETRY=1` [^L2OTelObservability]:
- **Metrics:** token counts, latency, tool decisions (accept/reject), lines of code changed
- **Log events:** structured records for prompts, API requests, tool results
- **Traces:** spans for interactions, LLM requests, tool calls, hooks

**Reasoning Redaction Policy:**  
The documentation explicitly states: "Importantly, the thinking text itself stays private. Even when you enable raw API body logging, extended-thinking content is redacted from the exported bodies" [^L2ReasoningAPITelemetry]. OpenTelemetry is designed for operational visibility, not reasoning audit.

**Soundness for adjudication:** ★★★★☆  
Token latency and tool call patterns are reliable. The deliberate redaction of thinking means **reasoning reconstruction from telemetry alone is unsound**. OTel is suitable for "did the agent call tools efficiently?" not "was the reasoning sound?"

### 4. Adaptive Reasoning (Claude 4.6, Successor to Extended Thinking)

**Configuration:**  
The newer adaptive thinking feature (Feb 2026+) allows setting effort levels:
```json
{
  "thinking": {
    "type": "adaptive"
  },
  "output_config": {
    "effort": "medium"  // or "low", "high", "max"
  }
}
```
[^L2AdaptiveThinking]

**What It Exposes:**  
Responses return `thinking` blocks with the same structure as extended thinking (summarized or omitted). The API does not expose which effort level was chosen by the model, internal reasoning branches, or how the effort parameter influenced the decision.

**Soundness for adjudication:** ★★☆☆☆  
Adaptive reasoning hides the "thought budget" used. Reproducibility is degraded: the same prompt under different latency constraints may produce different reasoning traces but identical final outputs, making adjudication of reasoning quality impossible without controlled re-execution.

### 5. Structured Outputs API (Not Reasoning-Specific)

**What It Is:**  
The Structured Outputs API compiles JSON schemas into token-generation constraints, guaranteeing response schema validity [^L2StructuredOutputs]. This is a contract on output format, not reasoning exposure.

```json
{
  "anthropic-beta": "structured-outputs-2025-11-13",
  "output_format": {
    "type": "json_schema",
    "json_schema": {...}
  }
}
```

**Relevance to Reasoning Capture:**  
Zero. Structured outputs constrain what is returned, not what is reasoned. An agent returning valid JSON may have arrived at the answer through unsound reasoning.

**Soundness for adjudication:** ☆☆☆☆☆ (not applicable)

### 6. Debug and Introspection APIs (Not Found)

**Search Finding:**  
No public Claude Code API or configuration exposes structured reasoning summaries, decision trees, alternative branches, or confidence scores [^L2DebugModeSearch]. GitHub feature requests exist (#10084, "Expose Claude Code Cognitive Telemetry States via API"), indicating this capability is desired but not shipped.

**Soundness for adjudication:** N/A — capability does not exist

---

## What Is Sound Enough for Citable Findings

### Tier 1: Direct Acts (High Confidence)

**Citable Statements:**
- "Agent A made X tool calls; Agent B made Y tool calls." [^L2TranscriptFormat]
- "Agent A's run consumed N input tokens and M output tokens." [^L2TranscriptFormat]
- "Agent A called the file-write tool with arguments {file: X, content: Y}." [^L2TranscriptFormat]
- "Agent A's request latency was T milliseconds." [^L2OTelObservability]

**Citation Mechanism:**  
Cite the run's transcript path or OpenTelemetry export ID. Provide the exact turn number and message UUID. Auditors can reproduce by replaying the JSONL.

**Confidence Grade:** VERIFIED (facts of execution)

---

### Tier 2: Tool Sequence Patterns (Medium Confidence, Requires Method Disclosure)

**Citable Statements:**
- "Agent A exhibits a backtracking pattern: it revisited the same file path in 5 of 12 tool calls without code changes between visits." [^L2TrajectoryEval]
- "Agent A's tool selection aligns with the task context: 8 of 8 Read calls targeted relevant files; 7 of 8 Bash calls were valid test commands." [^L2TrajectoryEval]

**Citation Mechanism:**  
Disclose the heuristic used to detect the pattern (e.g., "file path equality check across tool_use input JSON"). Provide the transcript range. Note that this is inference, not observation.

**Confidence Grade:** PLAUSIBLE (pattern-based, requires independent verification)

**Disconfirming Evidence Needed:**  
Verify the pattern against the tool results. If Agent A revisited the same file but each read returned different content or the tool failed, the "backtracking" interpretation is wrong. An auditor must re-read the tool results to confirm the pattern interpretation.

---

### Tier 3: Reasoning Reconstruction (Low Confidence Without Auxiliary Evidence)

**Not Citable Alone:**
- "Agent A reasoned that file X was the root cause."  
- "Agent A decided on a two-phase approach: retrieval then synthesis."

**Why:**  
Tool sequences can be consistent with multiple reasoning hypotheses. The agent may have called the retrieval tool first by accident, coincidentally followed by synthesis, creating an illusory "two-phase" pattern.

**Path to Citable Status:**  
Pair transcript analysis with auxiliary evidence:
1. **Re-execution trace:** Run the same prompt under the same configuration and compare tool sequences. If both runs follow the same pattern, the reasoning hypothesis is more plausible [^L2TrajectoryEval].
2. **Evidence chain reconstruction:** Map each tool call to a claimed inference ("Agent needed file X to infer Y"). Verify the inference against the tool result. If the tool result contradicts the claimed inference, the reasoning hypothesis is false [^L2EvidentTracing].
3. **Formal trace verification:** Use a framework like VeryTrace to compile the tool sequence into a formal model, then check it for logical consistency [^L2VeryTrace].

---

### Tier 4: Reasoning Quality (Not Citable from Transcripts Alone)

**Cannot Be Concluded from Transcripts:**
- "Agent A's reasoning was sound / flawed."  
- "Agent A avoided hallucination / committed hallucination."  
- "Agent A's confidence was calibrated / overconfident."

**Why:**  
Reasoning traces are not captured (encrypted or redacted). Hallucination requires comparing the agent's claims against ground truth, which transcripts do not carry. Confidence is not logged.

**What Researchers Do Instead:**  
The recent literature (2025–2026) on agent auditing converges on **evidence-chain verification** and **trajectory analysis frameworks**:
- **AgentAuditor** (arXiv:2602.09341): Resolves reasoning conflicts by comparing branches at critical divergence points [^L2AgentAuditor].
- **VeryTrace** (arXiv:2606.24124): Compiles natural-language reasoning into formal state transitions and checks them against formal specifications [^L2VeryTrace].
- **AgentLTL** (arXiv:2607.02599): Specifies procedural compliance inline with JSONL traces, then checks trace compliance against the spec [^L2AgentLTL].
- **From Agent Traces to Trust** (arXiv:2606.04990): Surveys evidence tracing and execution provenance, identifying the gap: "No single system spans trace sources, fine granularity, runtime timing, an explicit representation, and multiple trust functions at once" [^L2TracesSurvey].

The consensus: **transcripts alone are insufficient for reasoning adjudication**. Sound adjudication requires:
1. **Complete provenance traces** (why each claim was made, what evidence supports it)
2. **Explicit reasoning formalization** (converting natural-language reasoning into verifiable logical form)
3. **Multi-source evidence** (tool results, retrieved documents, memory reads, semantic links)

---

## Critical Gaps: What Cannot Be Audited from Current Telemetry

1. **Decision Alternatives** — Transcripts do not record what tools the agent *considered but rejected*. Adjudication cannot ask, "Were there better choices?" [^L2TrajectoryEval]

2. **Confidence and Uncertainty** — No field captures the agent's confidence in each decision. Internal confidence is not exposed [^L2OTelObservability].

3. **Claim-to-Evidence Links** — When an agent makes a factual claim in natural language, transcripts do not tag which tool result (if any) supports it. Hallucination is undetectable [^L2TracesSurvey].

4. **Semantic Dependencies** — Tool results are independent blocks. The transcript does not represent how one result causally influenced the next decision [^L2TracesSurvey].

5. **Reasoning Branches** — Extended thinking summaries are single-threaded narration. Agents often consider multiple hypotheses; transcripts capture only the chosen path [^L2TrajectoryEval].

6. **Context Decay** — Transcripts are append-only. The agent cannot annotate which earlier observations it relied on for a decision, or which it ignored [^L2TracesSurvey].

---

## Practical Guidance for Citable Adjudication

### For Individual Agent Runs

**Tier-1 (Directly Citable):**
- Count tool calls by type. Measure latency and token usage. Report with transcript path.

**Tier-2 (Conditional on Method Disclosure):**
- Analyze tool-call sequences for patterns (e.g., test-driven retry frequency). Disclose the pattern-detection heuristic. Prepare to re-audit with independent verification.

**Tier-3+ (Multi-Run or Experimental Setup Required):**
- For reasoning quality claims, implement comparative re-execution: run the same prompt on both agents under fixed conditions, analyze trajectory agreement. Cite the experimental protocol.
- For hallucination detection, overlay transcripts with ground-truth data (e.g., correct API outputs vs. claimed outputs) and check for divergence.

### For Comparative Adjudication

**Tier-1 Example:**  
"Agent A used 47% fewer tokens than Agent B to complete the same task" — cite token counts from both transcripts' message.usage fields.

**Tier-2 Example:**  
"Agent A exhibits fewer failed retries (3 vs. 12). The tool error patterns differ: Agent A's errors were transient (Bash timeouts); Agent B's were logical (wrong file path)." — cite error messages from tool_result blocks; disclose the classification heuristic.

**Tier-3 Example:**  
"Agent A's reasoning appears more coherent: when re-run on the same seed prompt, 8 of 10 re-runs follow similar tool sequences (tool-call overlap > 70%), suggesting stable reasoning. Agent B's re-runs diverge (overlap < 40%), suggesting brittle or random reasoning." — requires multiple runs, statistical analysis, explicit reproducibility details.

---

## Recommendation for Citable Findings

**Standard Practice (Adopted in Recent Agent-Auditing Research):**  
Disclose three elements in any claim:
1. **What was observed** (e.g., tool counts, error rates) — directly from transcripts or telemetry.
2. **How it was measured** — the heuristic, script, or algorithm used.
3. **What it does NOT tell you** — e.g., "Tool count reflects efficiency, not correctness." [^L2TrajectoryEval]

This three-part structure is used in arXiv:2510.02837 (trajectory evaluation) and is the standard in agent-auditing benchmarks (ASTRA-bench, Litmus, AgentLTL) [^L2AgentBenches].

---

## Summary Table: Telemetry Soundness by Use Case

| Use Case | Data Source | Direct Citable? | With Auxiliary Evidence? | Sound? | Notes |
|---|---|---|---|---|---|
| Token efficiency comparison | JSONL `message.usage`, OTel metrics | ✓ | — | ✓✓✓✓✓ | Facts of execution |
| Tool call frequency | JSONL tool_use blocks | ✓ | — | ✓✓✓✓✓ | Observable sequence |
| Error recovery patterns | JSONL tool_result error fields | ✓* | ✓ | ✓✓☆☆☆ | Requires interpretation; check results |
| Reasoning path comparison | JSONL thinking blocks | ✗ | ✓ | ✓☆☆☆☆ | Thinking is encrypted/summarized; needs re-execution |
| Reasoning quality judgment | Transcripts alone | ✗ | ✗ | ✗☆☆☆☆ | Requires formal verification frameworks |
| Hallucination detection | Transcripts vs. ground truth | ✗ | ✓ | ✓✓☆☆☆ | Requires overlay with external facts |

*Directly observable, but requires tool_result interpretation for full soundness.

---

## Footnotes

[^L2TranscriptFormat]: Claude Code JSONL transcripts store one JSON object per line under `~/.claude/projects/<encoded-path>/<session-id>.jsonl`. Each line is an event: `type` field discriminates (user/assistant/system); assistant lines include `message.content` array with content blocks (text, thinking, tool_use, tool_result). Specification: claude-dev.tools/docs/jsonl-format. Accessed 2026-07-18.

[^L2ExtendedThinkingDocs]: Anthropic Claude Platform documentation, "Extended thinking" (https://platform.claude.com/docs/en/build-with-claude/extended-thinking). Accessed 2026-07-18. Describes `thinking` parameter with `type: "enabled"`, `budget_tokens`, and `display` options (summarized, omitted).

[^L2VisibleExtendedThinking]: Anthropic news (https://www.anthropic.com/news/visible-extended-thinking). Quotes: "we don't know for certain that what's in the thought process truly represents what's going on in the model's mind" and "models often make decisions based on factors that they don't explicitly discuss in their thinking process."

[^L2TranscriptLimitations]: HypoGray, "Why Claude Code hides its thinking — and how to turn it back on" (https://hypogray.com/stories/claude-code-hides-thinking). Describes redact-thinking-2026-02-12 header (Feb 12, 2026) and encrypted signature (~600 chars) in on-disk JSONL. Accessed 2026-07-18.

[^L2TrajectoryEval]: Kim et al., "Beyond the Final Answer: Evaluating the Reasoning Trajectories of Tool-Augmented Agents" (arXiv:2510.02837, October 2025). Argues that trajectory-level evaluation (tool sequences, intermediate outputs, decision paths) is necessary to distinguish lucky guesses from sound reasoning. Accessed via https://arxiv.org/pdf/2510.02837. Access date 2026-07-18.

[^L2OTelObservability]: Anthropic Claude Code documentation, "Observability with OpenTelemetry" (https://code.claude.com/docs/en/agent-sdk/observability). Describes spans (claude_code.interaction, claude_code.llm_request, claude_code.tool), metrics (token counts, latency, tool decisions), and log events. Specifies thinking redaction: "extended-thinking content is redacted from the exported bodies." Accessed 2026-07-18.

[^L2ReasoningAPITelemetry]: Anthropic research and documentation on telemetry. States: "Importantly, the thinking text itself stays private. Even when you enable raw API body logging, extended-thinking content is redacted from the exported bodies." From various sources including observability.docs and API specifications. Accessed 2026-07-18.

[^L2AdaptiveThinking]: Anthropic Claude Platform documentation, "Adaptive thinking" and "Effort" (https://platform.claude.com/docs/en/build-with-claude/adaptive-thinking, https://platform.claude.com/docs/en/build-with-claude/effort). Describes `type: "adaptive"` and `output_config.effort` levels (low, medium, high, max). No API for inspecting internal reasoning branches or effort allocation. Accessed 2026-07-18.

[^L2StructuredOutputs]: Anthropic Claude Platform documentation, "Structured outputs" (https://platform.claude.com/docs/en/build-with-claude/structured-outputs). Describes JSON schema constraints via anthropic-beta header. This is an output format contract, not reasoning exposure. Accessed 2026-07-18.

[^L2DebugModeSearch]: Web search for Claude Code debug modes, reasoning APIs, and cognitive telemetry (query: ""Claude Code" debug mode reasoning summary API settings"). Found feature request GitHub #10084 ("Expose Claude Code Cognitive Telemetry States via API") indicating this capability is desired but not publicly available. Search conducted 2026-07-18.

[^L2WhyHidingThinking]: Anthropic blog and documentation on redact-thinking-2026-02-12 header. Cites "latency reduction and user-experience clarity" for suppressing thinking renders in Claude Code terminal UI. Accessed from HypoGray and developer documentation, 2026-07-18.

[^L2EvidentTracing]: Chen et al., "From Agent Traces to Trust: A Survey of Evidence Tracing and Execution Provenance in LLM Agents" (arXiv:2606.04990v3, June 2026). Defines evidence tracing as connecting evidence units to claims and introduces provenance relations (Support, Derive, Depend-on, Contradict, Invalidate, Trigger, Update). Accessed via https://arxiv.org/html/2606.04990v3. Access date 2026-07-18.

[^L2TracesSurvey]: Chen et al. arXiv:2606.04990v3, §III. Key quote: "No single system spans trace sources, fine granularity, runtime timing, an explicit representation, and multiple trust functions at once." Identifies fragmentation across retrieval attribution systems, safety tool-parameter checks, and observability frameworks. Access date 2026-07-18.

[^L2VeryTrace]: Xu et al., "VeryTrace: Verifying Reasoning Traces through Compilable Formalism and Structured Verification" (arXiv:2606.24124, June 2026). Proposes converting natural-language reasoning into formal logical systems for automated verification. Addresses gap that current reasoning outputs lack machine-checkable structure. Accessed via https://arxiv.org/pdf/2606.24124. Access date 2026-07-18.

[^L2AgentAuditor]: Jiao et al., "Auditing Multi-Agent LLM Reasoning Trees Outperforms Majority Vote and LLM-as-Judge" (arXiv:2602.09341, February 2026). Describes AgentAuditor method: resolves reasoning conflicts by comparing branches at critical divergence points, turning global adjudication into localized verification. Shows advantage over consensus methods. Accessed via https://arxiv.org/pdf/2602.09341. Access date 2026-07-18.

[^L2AgentLTL]: Lemos et al., "AgentLTL: A Trace-Verification Framework for Measuring, Enforcing, and Training Procedural Compliance in Tool-Using LLM Agents" (arXiv:2607.02599, July 2026). Proposes LTL specification language for both offline eval and online enforcement of agent traces. Accessed via https://arxiv.org/pdf/2607.02599. Access date 2026-07-18.

[^L2AgentBenches]: References to ASTRA-bench (arXiv:2603.01357), Litmus benchmark (arXiv:2604.08970), and other 2025–2026 agent evaluation frameworks that adopt trajectory-level analysis. All use method-disclosure as standard practice. Accessed 2026-07-18.
