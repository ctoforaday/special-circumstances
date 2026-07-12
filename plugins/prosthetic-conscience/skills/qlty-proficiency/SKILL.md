---
name: qlty-proficiency
description: Use when formatting, linting, or assessing code quality — operational patterns for the qlty command-line tool, with the maintained cheatsheet.
---

# qlty-proficiency

Uniform quality across languages via one tool.

- BEFORE running quality operations, YOU MUST confirm qlty is present (the shared toolchain probe / `/sc-doctor`); when absent, say so and skip — do not improvise per-language substitutes without being asked.
- During formatting and linting, YOU MUST use the patterns in this skill's `CHEATSHEET.md` (`qlty fmt`, `qlty check`, sample/config commands) rather than ad-hoc tool invocations.
- AFTER discovering a new qlty pattern or trap, YOU MUST record it in `CHEATSHEET.md`.
