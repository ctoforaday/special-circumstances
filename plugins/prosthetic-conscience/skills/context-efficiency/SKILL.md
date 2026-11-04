---
name: context-efficiency
description: Always-on context-budget discipline — shield the context with subagents and files, peek instead of streaming, and phase work so artifacts survive compression.
disable-slash-command: true
user-invocable: false
---

# context-efficiency

Protect the context. Leave artifacts that survive compression.

- BEFORE reading bulk output (test runs, logs, big files, wide sweeps), YOU MUST shield the context: run the noisy process in the background with output to a file, then peek with targeted reads (a search, `tail -n 50`, offset reads) — or delegate the reading to a subagent whose context absorbs the bulk and whose return is only the conclusion plus pointers. YOU MUST NOT stream raw output into the conversation.
- During tool use, YOU MUST issue calls that do not depend on each other — reads, searches, status checks — together in one step, never one per step. Every step re-reads the whole context, so a serial chain of independent calls multiplies the cost and adds no information.
- AFTER sending a command to the background, YOU MUST wait for the harness's completion notice — do other independent work, or end the turn — and YOU MUST NOT poll it for status. A peek at its output serves a decision the output changes, such as stopping a wedged run, never the question of whether it has finished. Where the harness asks how long to wait before backgrounding, give the command's expected duration.
- AFTER a large tool result, YOU MUST carry forward the conclusion, not the dump; if the data will be needed again, YOU MUST write it to a file rather than re-read it into context.
- BEFORE a complex operation, YOU MUST break it into sequential phases — design, implement, verify — and AFTER each phase, YOU MUST leave a written artifact (plan, todo list, state file, validation loop). Phases exist to create **recoverable moments**: artifacts survive context compression when reasoning doesn't. Compressed mid-implementation, you work back from the written design; compressed mid-verification, you trust the written validation loop.
