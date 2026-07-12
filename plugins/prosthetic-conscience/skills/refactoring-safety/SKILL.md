---
name: refactoring-safety
description: Use when refactoring, restructuring, or resolving a class of fault — preserve behavior and directives, and fix the whole class, not one instance.
---

# refactoring-safety

Structure MUST NOT cost substance.

- Structural cleanup (headers, templates, reorganization) YOU MUST NOT achieve at the expense of a specific operational directive; if you relocate a directive, leave a high-leverage pointer where it was.
- BEFORE finalizing a refactor, YOU MUST diff the new version against the original and confirm every behavior and responsibility is still accounted for.
- When you find a fault, YOU MUST fix the **class of problem** everywhere it occurs — not only the single reported instance.
