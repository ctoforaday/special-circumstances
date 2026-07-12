# Doubts — hypotheses about our own design, validated against dogfood runs

Running list. Each doubt gets checked against real run artifacts; confirmed doubts graduate to
the backlog or a plan; refuted ones get closed with the evidence. (Doubting our own design is
the same discipline we impose on research — think-around-problem applied to ourselves.)

## Open — check against run 2026-07-12_memory-architecture

- [ ] **Blue lane diversity is unenforced.** The only differentiation is "lane N takes hypothesis N
  first, then breadth" — no distinct search strategies, no assigned perspectives, no forbidden
  overlap. Prediction to test: significant convergence between `blue/candidates/lane-*.md`
  (same sources, same structure, same conclusions). If confirmed: lanes need engineered diversity
  (distinct lenses/methods/source-classes per lane), not just different starting hypotheses.
- [ ] **The Heilmeier is DARPA-shaped.** Cost/duration/milestone questions fit proposals and
  programs; they map awkwardly onto explainer or survey questions. Watch how the catechism reads
  for an architecture-evaluation topic. If strained: template variants by question type
  (decision / feasibility / survey), or applicability-marked sections.
- [ ] **The risk matrix presumes a design under evaluation.** Same variant concern as Heilmeier.
- [ ] **Duplication between "the meat" and blue-in-full.** The analytical core derives from blue's
  report; union-not-summary doesn't tell the assembler what belongs where. Intended division:
  meat = the agreed post-debate answers; blue = the evidentiary record. Check whether the
  assembler repeats blue verbatim.
- [ ] **Blue's `open_questions` have no home in the report template** — the envelope carries them,
  the template drops them. Check whether they survive anywhere in report.md.
- [ ] **Consensus vs. minority reports are indistinguishable after synthesis.** The union merge
  erases claim provenance: a claim both lanes found independently (strong convergent signal) reads
  identically to a single lane's leaf-node discovery (minority report — possibly gold, possibly
  noise; red should weigh them differently). Measure on this run: compare `blue/candidates/lane-*`
  against `blue/report.md` and classify claims consensus/minority. If the distinction matters:
  the synthesizer tags claims with lane provenance during the union, and red's corroboration
  grading gets provenance as an input.

## Closed

- [x] **Cross-plugin `skills:` preloading reaches workflow agents** — CONFIRMED (run-1 forensics):
  all 16 transcripts carry the literal injected content of `frank-exchange-of-views:research-protocol`
  AND `prosthetic-conscience:critical-stance`. The mindset-inheritance mechanism is real.
- [x] **Plugin hooks + agent memory in workflow agents** — CONFIRMED (hook log + agent-memory file):
  sc-quality-gate fired on workflow-agent writes; red-auditor wrote its `memory: project` gap-pattern
  file. Every harness mechanism the suite depends on is now empirically verified.

(none yet)
