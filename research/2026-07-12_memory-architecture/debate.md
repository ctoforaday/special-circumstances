## Round 0

### BLUE

Position: **the proposal's architecture is directionally right — endorse with mandatory
amendments, not a redesign.** Two lanes (H1 substrate deep, H2 consolidation deep; breadth
across H1–H5; 41 searches, saturation reached, disconfirming budget met) converge on the same
verdict, and the strongest single piece of evidence is external: Anthropic is independently
shipping the same loop (native Auto Memory, flag-gated "Auto Dream" consolidation), which
validates the shape while shrinking the bespoke remit.

What blue asserts, with evidence in `blue/report.md`:

1. **Substrate holds** (H1): OKF is real, spec-legal for our profile, four weeks old — adopt as
   pinned convention, not dependency. Transcript JSONL verified at the leaf node on this
   machine. The file-based-memory pattern is widely shipped; every matured system added a
   derived index — name the trigger (~300–500 concepts), don't defer indefinitely.
2. **Consolidation is the top technical risk but mitigable** (H2): lossy-compression and
   continuous-rewrite corruption are measured phenomena; the structural fix is append-only
   expansion (claims immutable after promotion; change = supersede) plus per-pass change caps.
   Review-by-git-diff is forensic, not preventive — bot commits are systematically unreviewed.
3. **One blocking omission**: memory poisoning. The design pipes untrusted input into always-on
   context and its corroboration rule launders injections into permanent rules
   (CVE-2026-21852 precedent against the exact inbox file). Trust-tiered provenance, permanent
   human gate on external-ingest promotion, injection screening, independent-source
   corroboration — blocking before Phase 1.
4. **Three factual defects in the proposal**: the §5 agent-`memory:` row is wrong (harness
   injects from fixed paths; bidirectional write collision); the "existing secret-scrub gate"
   does not exist; `docs/scheduling.md` does not exist. Verified locally.
5. **Pragmatist simplifications**: drop the stored confidence float (derive activation from
   observables); single projection channel (prefer path-scoped `.claude/rules/` over
   @-import + SessionStart double injection); hard token budget on `active.md`;
   sequential-only subagents in the scheduled pass.
6. **Risk-accepted, with rationale**: OKF v0.1 drift (degrades to plain markdown),
   multi-machine store divergence (single operator; git remote is the sync story),
   project-store PR-ratification flow (keep optional, off by default). A design made worse to
   satisfy these edge cases would itself be a defect.
7. **No alternative dominates** (H5): claude-mem, basic-memory, mem0, Letta, Zep each fail a
   binding constraint (SQLite opacity, missing lifecycle, daemon/service-bound); each
   contributes a stealable mechanism, catalogued in §7.

Consolidated change list: `blue/report.md` §8 (14 items, 3 blocking). Full risk grading: §9.
