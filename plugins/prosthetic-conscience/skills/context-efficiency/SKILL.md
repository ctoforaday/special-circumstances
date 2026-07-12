---
name: context-efficiency
description: Always-on context-budget discipline — prune aggressively, snapshot rather than stream, and partition work into sequential phases.
---

# context-efficiency

Manage the context budget deliberately.

- BEFORE a new task phase, YOU MUST summarize and prune context no longer in use.
- During a turn, YOU MUST NOT accumulate raw tool output — AFTER reading it, keep the conclusion and discard the dump.
- During inspection, YOU MUST NOT stream (`tail -f`); take a discrete snapshot (`tail -n 50`, `ls -l`) and move on.
- BEFORE a complex operation, YOU MUST break it into sequential phases; YOU MUST NOT design, implement, and verify in one undifferentiated pass.
