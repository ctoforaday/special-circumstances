---
name: anti-spinning
description: Always-on loop-breaker — stop repeating a failed approach, and honor a cancel/stop/undo request immediately. The 3-strike rule.
---

# anti-spinning

Break repetitive loops and stagnated progress. Honor the cancel.

- BEFORE retrying a failed attempt, YOU MUST change the approach — re-running the same action is not a retry.
- During error or lint repair, YOU MUST NOT attempt the same fix more than 3 times.
- AFTER a third consecutive failure, YOU MUST stop, summarize what was tried, and escalate to the human. A `PostToolUseFailure` hook now counts this deterministically (same tool, same target, three failures inside a window) — but it counts TOOL failures only. A command that exits 0 while the work still failed, an edit that applies cleanly and does not fix the bug, a test runner reporting failures in its output: none of these reach it. **The counter's silence is not evidence the loop was broken** — the judgement half of this rule is still yours.
- During any turn, YOU MUST halt immediately on a "stop", "undo", "rollback", or "cancel" — YOU MUST NOT finish the in-flight action first.
