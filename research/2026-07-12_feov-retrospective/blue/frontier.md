# Frontier hypotheses (Round 0)

Grounded in a first pass over `inputs/doubts.md`, `inputs/backlog.md`, `inputs/run1-friction.md`,
`inputs/run2-friction.md`, the run-2 corpus (`blue/candidates/lane-{1,2}.md`, `blue/CHANGELOG.md`,
`blue/frontier.md` from that run), and `frank-exchange-of-views/commands/research.md` +
`skills/research-protocol/scripts/workflow.js`. Restated as falsifiable predictions before deep
research begins.

- **H1 — Lane diversity is unenforced and lanes converge in practice.** If true: run 2's
  `blue/candidates/lane-1.md` (assigned H1-substrate-first) and `lane-2.md` (assigned
  H2-consolidation-first) will show substantial structural and substantive overlap despite
  different starting hypotheses — same missing-risk discovery (memory poisoning, independently
  headlined by both as "absent from §9"), same H1/H2/H4/H5 coverage by the "breadth" phase, and
  a CHANGELOG dominated by dedup-and-merge operations rather than reconciliation of genuinely
  distinct material. Only 2 lanes were dispatched though the command's documented default is 3 —
  worth checking whether lane count itself was under-provisioned this run.
- **H2 — Consensus vs. minority provenance is destroyed at synthesis, not merely under-surfaced.**
  If true: `blue/CHANGELOG.md`'s merge language ("kept both," "union of," "merged X §n + Y §n") will
  contain no vocabulary distinguishing a claim both lanes reached independently (strong convergent
  signal) from a claim only one lane surfaced (minority report) — and `blue/report.md` prose will read
  identically for both classes. Corollary: red's own corroboration grading in `red/findings.md` will
  show no input channel for lane-count-agreement, meaning red is regrading claims blind to whether
  they were 1-lane or 2-lane sourced.
- **H3 — The defect population is bimodal: harness/caller-plumbing bugs are zero-token
  unit-testable by stubbing `agent()`; leaf-node/citation-verification bugs are only observable
  live or in production.** If true: classifying every run-1 and run-2 defect (undefined run-dir,
  filename-write-guard false positives, ENAMETOOLONG heredocs, deadlocked rounds with no inputs,
  ARC/PDF-table figures, live-source drift on gh issue status and star counts) will show a majority
  reproducible against a Node simulator that stubs `agent()` with canned envelopes and asserts on
  the caller's argument-construction, file-path guards, and round/deadlock control flow — while a
  persistent residual (PDF table extraction, primary-advisory access, live citation drift, hook-fire
  verification) requires real tool calls or real filesystem/network state and cannot be pulled into
  the simulator at any complexity cost.
- **H4 — The highest-leverage pre-run-4 changes are structural and cheap, not exploratory.** If
  true: proposals already implied by the friction/backlog corpus — a caller-side arg-shape preflight
  guard, an explicit artifact-filename allowlist/carve-out for the write-block heuristic, and an
  engineered (not accidental) per-lane diversity assignment (distinct lenses/source-classes, not just
  different starting hypotheses) — will each grade high-likelihood x high-impact x low-complexity,
  because run 1's root cause (stringified args) was already caller-side and run 2's write-block was
  already a known filename heuristic, not a hard capability limit.
- **H5 — The friction corpus is dominated by 2-3 systemic capability gaps independently reported
  by multiple distinct agent roles across multiple rounds, not a long tail of one-off complaints.**
  If true: counting distinct-agent-role attributions in `run1-friction.md` + `run2-friction.md` will
  show PDF-full-text/table-extraction and primary-source (security-advisory) access each reported
  by blue, red, AND judge roles across 3+ rounds, while the write-block/ENAMETOOLONG/preflight-guard
  complaints cluster in round 1 only (already fixed) — meaning the *stable, unresolved* ranked list
  is short and dominated by document-fetch fidelity, not tool diversity.
