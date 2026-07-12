---
name: context-efficiency
description: Always-on context-budget discipline — shield the context with subagents and files, peek instead of streaming, and phase work so artifacts survive compression.
---

# context-efficiency

Protect the context. Leave artifacts that survive compression.

- BEFORE reading bulk output (test runs, logs, big files, wide sweeps), YOU MUST shield the context: run the noisy process in the background with output to a file, then peek with targeted reads (`Grep`, `tail -n 50`, offset reads) — or delegate the reading to a subagent whose context absorbs the bulk and whose return is only the conclusion plus pointers. YOU MUST NOT stream raw output into the conversation.
- AFTER a large tool result, YOU MUST carry forward the conclusion, not the dump; if the data will be needed again, YOU MUST write it to a file rather than re-read it into context.
- BEFORE a complex operation, YOU MUST break it into sequential phases — design, implement, verify — and AFTER each phase, YOU MUST leave a written artifact (plan, todo list, state file, validation loop). Phases exist to create **recoverable moments**: artifacts survive context compression when reasoning doesn't. Compressed mid-implementation, you work back from the written design; compressed mid-verification, you trust the written validation loop.
