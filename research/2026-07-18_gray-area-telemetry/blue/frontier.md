# blue frontier — Trajectory telemetry for agent adjudication: what can be mined from Claude Code transcripts today, what settings or APIs expose a summary of an agent's reasoning rather than only its acts, and which of those are sound enough to carry a citable finding?

## H1: Transcript Acts Are The Only Direct Record; Reasoning Is Reverse-Engineered
**What would be true if right**: Claude Code transcripts capture only tool calls, outputs, and final responses. All reasoning must be reconstructed by examining the sequence of acts. Adjudication would depend entirely on behavioral inference without explicit reasoning traces.

**Disconfirming search**: Find explicit reasoning state, thinking traces, or reasoning summaries in Claude Code transcript formats; discover APIs that expose reasoning directly.

---

## H2: Extended Thinking Tags Provide Sound Reasoning Traces in Transcripts
**What would be true if right**: Claude Haiku 4.5+ extended thinking (enclosed in thinking tags) is captured in Claude Code transcripts and provides verifiable, auditable reasoning traces sufficient for agent adjudication. Thinking content survives transcript recording and is accessible for analysis.

**Disconfirming search**: Verify thinking tags are not captured in transcripts; find thinking is unreliable, inconsistent, or post-hoc; discover thinking is stripped before transcript serialization.

---

## H3: Claude Code Exposes Structured Reasoning Summaries Via Dedicated APIs or Debug Settings
**What would be true if right**: Beyond raw transcripts, Claude Code or the Claude Code plugin system provides dedicated APIs, debug modes, or configuration settings that expose structured reasoning summaries (decision trees, alternative branches, confidence scores) without requiring reverse-engineering from acts.

**Disconfirming search**: Verify no such APIs exist; find all reasoning export is transcript-only; discover any reasoning APIs are internal or undocumented.

---

## H4: Reasoning Reconstruction From Acts Alone Is Insufficient for Citable Adjudication; Auxiliary Telemetry Required
**What would be true if right**: Transcript-based reasoning reconstruction lacks critical context needed for sound adjudication. Citable findings require auxiliary telemetry: timing of decisions, model state transitions, context window allocation, tool-choice alternatives considered, or explicit confidence traces that are not present in act sequences.

**Disconfirming search**: Find act sequences provide sufficient grounds for principled adjudication; discover reasoning can be fully reconstructed without auxiliary telemetry.

---

## H5: Current Claude Code Transcripts Preserve Sufficient Context to Distinguish Good From Bad Agent Decisions
**What would be true if right**: The current Claude Code transcript format (including reasoning-relevant fields, decision points, error handling, and retry logic) preserves enough context to enable principled adjudication of agent quality without auxiliary telemetry or new APIs. Citable findings are reachable from transcript content alone.

**Disconfirming search**: Identify critical decision context missing from transcripts; find adjudication requires data not present in current transcript formats; discover decision alternatives or reasoning branches are not captured.

---
