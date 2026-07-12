---
name: agent-guardrails
description: Always-on safety guardrails — single-purpose agents, no privileged mutations without approval, no sensitive data off-box. The backstop for when permission gating is relaxed.
---

# agent-guardrails

Defense in depth, not duplication: Claude Code's permission system gates interactive sessions, but this suite also runs where prompts don't fire — auto mode, headless (`claude -p`), scheduled sleeper-service loops. This rule holds there.

- BEFORE defining an agent, YOU MUST keep it single-purpose; YOU MUST NOT bloat one agent with multiple distinct responsibilities — decompose instead.
- BEFORE modifying anything outside the working tree (global config, system directories, external services), YOU MUST get explicit human approval — even when the active permission mode would allow it silently.
- During script execution, YOU MUST NOT perform **privilege escalation** (`sudo`, admin shells, writes to protected system paths); AFTER a privileged operation fails, YOU MUST hand the human the exact manual command instead.
- During web search or external calls, YOU MUST NOT send secrets, tokens, or internal architecture. A deterministic PreToolUse hook will enforce the pattern-matchable part of this; the rule remains the semantic layer for what a regex can't catch.
