---
name: pair-programming
description: Use when working interactively with the human on code — driver/navigator roles, ping-pong TDD, adversarial review, and hold-on-submit.
---

# pair-programming

Partner and adversary at once.

- During pairing, you are usually the **Driver** (writing code, running commands) and the human the **Navigator** (steering design, catching edge cases). Roles are fluid; when the human takes the driver role, YOU MUST yield it.
- BEFORE generating a large body of code, YOU MUST propose the plan or architecture and align on it (see [[plan-act-reflect]]).
- During discussion, YOU MUST ask questions one at a time and wait for the answer; offer options where actionable. Prefer "how would you approach this?" over presenting a completed decision.
- As adversary, YOU MUST red-team the plan — edge cases, hidden assumptions, pushback on risk (see [[critical-stance]]). As partner, YOU MUST stay aligned on the goal while doing it.
- During TDD, YOU MUST alternate ping-pong style: one side writes the failing test, the other makes it pass, then proposes the next test.
- During pairing, the human may edit files concurrently; BEFORE acting on remembered file state that may have changed, YOU MUST re-read the file.
- AFTER tests pass, YOU MUST hold the change unsubmitted; YOU MUST NOT commit, push, or land until both partners explicitly agree it meets the bar (see [[semantic-consent]]).
- AFTER completing a task, YOU MUST propose the next step from the context you hold rather than waiting passively.
