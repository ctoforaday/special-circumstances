---
name: spec-driven-development
description: Use when creating, reviewing, or auditing an implementation plan — the mandatory SDD structure (I–V), the TinySpec loop for small work, and the auditor gate.
---

# spec-driven-development

Intent separates from execution; every change is verified. All technical changes follow this standard.

## The plan structure (I–V)

Every implementation plan MUST contain:

1. **I. Summary & Goals** — objective (what problem) and success criteria (quantitative: "builds in < 2 min", "zero critical CVEs").
2. **II. Technical Context** — language/version, primary dependencies, storage/data model, constraints (security/privacy/network).
3. **III. Proposed Changes (the spec)** — grouped by component, `[NEW]`/`[MODIFY]`/`[DELETE]` tags, directory tree of the proposed structure.
4. **IV. Risk & Mitigation** — risks graded likelihood × impact × complexity-to-mitigate; each implementation step that mitigates one links to it.
5. **V. Verification Plan (the checklist)** — automated tests (exact commands), manual verification steps, and the auditor gate. This section is the seed of the [[validation-loop]] — it MUST be executable, not aspirational.

## The TinySpec loop

For small tasks (UI tweaks, minor bugs), a single `tinyspec.md` MAY combine plan and tasks for speed. TinySpec still carries success criteria and a verification checklist — small is not exempt from V.

## The auditor gate

- BEFORE an implementation plan is treated as approved, it MUST pass the auditor gate (`/plan-audit`): **Alignment** (solves the researched problem), **Completeness** (all files and dependencies accounted for), **Safety** (violates no guardrails).
- The gate is binary — `VERDICT: PASS` or `FAIL` with actionable gaps. It never soft-passes.
