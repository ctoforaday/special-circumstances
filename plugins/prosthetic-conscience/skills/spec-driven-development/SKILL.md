---
name: spec-driven-development
description: Use when creating, reviewing, or auditing an implementation plan — the mandatory five-section plan structure (numbered I through V, from Summary & Goals to the Verification Plan), the TinySpec loop for small work, and the auditor gate.
---

# spec-driven-development

Intent separates from execution; every change is verified. All technical changes follow this standard.

## The plan structure (I–V)

Every implementation plan MUST contain:

1. **I. Summary & Goals** — objective (what problem) and success criteria (quantitative: "builds in < 2 min", "zero critical CVEs").
2. **II. Technical Context** — language/version, primary dependencies, storage/data model, constraints (security/privacy/network).
3. **III. Proposed Changes (the spec)** — grouped by component, `[NEW]`/`[MODIFY]`/`[DELETE]` tags, directory tree of the proposed structure. A `[MODIFY]`/`[DELETE]` to a **contract** (a function signature, CLI flag set, arg shape, event schema, or a prompt a seat is told) MUST carry a **Consumer census**: the exact search that enumerates its callers, tests, and sibling scripts — *run, with results pasted* — one line per consumer stating whether it changes. Complete when re-running the command surfaces nothing the list omits. An executable artifact that survives, not a prose promise ([[context-efficiency]]).
4. **IV. Risk & Mitigation** — risks graded likelihood × impact × complexity-to-mitigate; each implementation step that mitigates one links to it.
5. **V. Verification Plan (the checklist)** — automated tests (exact commands), manual verification steps, and the auditor gate. This section is the seed of the [[validation-loop]] — it MUST be executable, not aspirational. §V MUST also name at least one **driveable check on real (not synthetic) data** — the change exercised against a real artifact or run, output observed. Fixtures prove logic; only real data surfaces data-shaped defects (fallback collisions, harness sentinels, encoding). Running it before "done" is the author's [[validation-loop]] duty; the gate can confirm §V *names* such a check, not that it was run.

## The TinySpec loop

For small tasks (UI tweaks, minor bugs), a single `tinyspec.md` MAY combine plan and tasks for speed. TinySpec still carries success criteria and a verification checklist — small is not exempt from V.

## The auditor gate

- BEFORE the gate, any **design fork with a behavioral, semantic, cost, or reversibility implication** MUST be resolved with the human. The gate vets ONE design against the standard; it is not where you discover which design is wanted — a fork reopened after PASS wastes every round spent on the discarded branch.
- BEFORE an implementation plan is treated as approved, it MUST pass the auditor gate (`/plan-audit`) on **Alignment**, **Completeness**, and **Safety** — defined in the `plan-auditor` agent, which is what applies them (one definition, where it is used).
- The gate is binary — `VERDICT: PASS` or `FAIL` with actionable gaps. It never soft-passes.
