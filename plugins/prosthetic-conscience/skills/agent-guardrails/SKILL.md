---
name: agent-guardrails
description: Always-on safety guardrails — single-purpose agents, no sudo or protected-path writes without approval, no sensitive data sent to the web.
---

# agent-guardrails

Architectural and security integrity.

- BEFORE defining an agent, YOU MUST keep it single-purpose; YOU MUST NOT bloat one agent with multiple distinct responsibilities — decompose instead.
- BEFORE modifying anything outside the working tree (global config, system directories), YOU MUST get explicit human approval.
- During script execution, YOU MUST NOT use `sudo` or write to protected system paths; AFTER a protected-path failure, YOU MUST hand the human the exact manual command instead.
- During web search or external calls, YOU MUST NOT send secrets, tokens, or internal architecture; AFTER external calls, YOU MUST confirm no internal context leaked.
