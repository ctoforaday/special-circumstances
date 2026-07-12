---
name: test-driven-development
description: Use when building or changing behaviour that can be tested — red/green/refactor discipline, test-first contracts, and the ping-pong variant for pairing.
---

# test-driven-development

The test is the specification; write it first.

- BEFORE writing implementation code for a testable behaviour, YOU MUST write the failing test that specifies it — and run it to confirm it fails for the expected reason (red).
- AFTER the test fails correctly, YOU MUST write the minimum implementation that makes it pass (green); YOU MUST NOT gold-plate beyond what the test demands.
- AFTER green, YOU MUST refactor with the suite as the safety net (see [[refactoring-safety]]) — structure improves only while tests stay green.
- During pairing, YOU MUST offer the ping-pong variant: one side writes the failing test, the other makes it pass, then proposes the next test (see [[pair-programming]]).
- During repair of a reported bug, YOU MUST first reproduce it as a failing test — a fix without a pinning test is a regression waiting to return.
- YOU MUST NOT claim test-driven development happened when tests were written after the implementation; test-after is verification, not specification. Both are honest; name them honestly.
