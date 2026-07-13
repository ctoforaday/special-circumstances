# Frontier hypotheses (reconstructed from lane drafts, Round 0)

Recorded before lane research began; restated here as the lanes tested them.

- **H1 — Substrate.** Git-native markdown-plus-YAML-frontmatter is a proven, sufficient substrate for agent memory; the Open Knowledge Format and the native Claude Code surfaces (`CLAUDE.md`, `MEMORY.md`, `memory:` frontmatter, hooks, `@`-imports) behave as the proposal's §3–§5 assume.
- **H2 — Consolidation.** LLM-driven expand-existing-before-append without a semantic index silently loses or fragments knowledge; dedup/consolidation is the design's Achilles heel.
- **H3 — Cadence.** Clock-driven nightly consolidation is the wrong trigger; the literature favors event-thresholded consolidation.
- **H4 — Complexity.** The lifecycle arithmetic (confidence floats, decay windows, promotion thresholds) is over-provisioned for a single-operator suite; a thinner design on native surfaces would win.
- **H5 — Alternatives.** An existing system (claude-mem, basic-memory, mem0, Letta, Zep, or the harness's own emerging machinery) dominates the bespoke design in part or whole.

Lane assignments: lane 1 took H1 to saturation then breadth; lane 2 took H2 to saturation then breadth. Both met the disconfirming-evidence budget (lane 1: 7 of 21 searches; lane 2: explicit disconfirming searches on file-memory success, LLM-judge dedup adequacy, files-win/YAGNI criticism, idle-consolidation validation).
