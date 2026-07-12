---
name: design-by-contract
description: Use when writing rules, defining agent roles, or authoring multi-step operational instructions — anchor every instruction to verifiable execution states.
---

# design-by-contract

All complex instructions MUST be state-based logical contracts.

- Format every rule with the contract keywords: `BEFORE [action], YOU MUST [behaviour]`; `During [state], YOU MUST NOT [behaviour]`; `AFTER [event], YOU MUST [behaviour]`.
- YOU MUST NOT use vague "Always / Never" imperatives for multi-step logic.
- **Semantic Precision**: YOU MUST NOT sacrifice intent or critical context for brevity; every contract MUST be unambiguous.
