---
name: refactoring-safety
description: Use when refactoring, restructuring, or resolving a fault — behavior-preserving, green build and tests, class-of-problem fixes, uniform design over hacks.
---

# refactoring-safety

Refactors preserve behavior and raise structure.

- BEFORE finalizing a refactor, YOU MUST diff the result against the original and confirm no behavior, contract, or directive silently disappeared — and the result MUST build and pass its tests. A refactor that doesn't compile is not a refactor; a green suite is the floor, not the goal.
- When you find a fault, YOU MUST fix the **class of problem** everywhere it occurs — the postmortem discipline: prevent the class, don't patch the instance. "Everywhere" spans every carrier of the concept, not just the file in front of you ([[complete-the-concept]]).
- During restructuring, YOU MUST drive toward uniform abstractions and recognizable patterns: no hacks unless architecturally unavoidable; no third copy of anything (the third duplication is the refactor trigger); no methods so long they can't be read without scrolling — size is a safety property, not a style preference.
