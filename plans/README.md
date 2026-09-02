# Plans

Planning and design artifacts for Special Circumstances. Each arrives as its own pull request for review; once agreed and built, the relevant pieces graduate into the plugins and the planning doc stays here as historical record. The directory therefore holds both live proposals and shipped plans (~39 files today) — deliberately separated from the repo's "real" documentation (`README.md`, the per-plugin READMEs). **The authoritative state of any plan is the dated `> STATUS` line under its own title** (shipped — historical record / in progress / superseded / not started), not this index.

Entry points:

- **`claude-port-plan.md`** — the master plan (shipped — historical record): a first-principles teardown of the Antigravity origin, the harness comparison, the original three-plugin architecture (now four — gray-area joined 2026-07), the inter-agent contracts, and the phased build plan. Phase 4 (sleeper-service) remains unbuilt.
- **`memory-architecture.md`** — proposal, not started: an Open-Knowledge-Format, git-native memory architecture (trajectory→memory skills, a `/dream` consolidation loop, global + project stores).
- **`context-checkpointing.md`** — shipped in **prosthetic-conscience** as the context-checkpointing skill and its checkpoint hooks; §15 records the corrections and the split from gray-area.
- **`gray-area.md`** — foundations for the fourth plugin, since built: trajectory mining, the 2.1.220 hook surface and what it lets us enforce, and a phased build. §4 records why continuity is *not* here.
- **`reasoning-telemetry.md`** — research record: what is observable in a session, and how to expose thinking summaries — the measured correction to "reasoning is not persisted", the three capture channels graded, and why a summary still cannot promote a finding.

The frank-exchange-of-views reform arc is its own cluster: `constitutional-reform.md` (the design), `change-waves.md` (the tracker it shipped through), `rulebook-audit.md` (the rules the evidence indicted), plus the record-layer plans (`record-*.md`).
