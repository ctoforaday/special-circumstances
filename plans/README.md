# Plans

Planning and design artifacts for Special Circumstances. Each arrives as its own pull request for review; once agreed and built, the relevant pieces graduate into the plugins and the planning doc stays here as historical record. The directory therefore holds both live proposals and shipped plans — deliberately separated from the repo's "real" documentation (`README.md`, the per-plugin READMEs). **The authoritative state of any plan is the dated status marker at the top of its own file** — a `> STATUS` line (shipped — historical record / in progress / superseded / not started), or a retirement banner where a plan was retired outright — never this index, which is an entry point and not a register.

Entry points:

- **`claude-port-plan.md`** — the master plan (shipped — historical record): a first-principles teardown of the Antigravity origin, the harness comparison, the original three-plugin architecture (now four — gray-area joined 2026-07), the inter-agent contracts, and the phased build plan. Phase 4 (sleeper-service) remains unbuilt.
- **`memory-architecture.md`** — proposal, not started: an Open-Knowledge-Format, git-native memory architecture (trajectory→memory skills, a `/dream` consolidation loop, global + project stores).
- **`context-checkpointing.md`** — shipped in **prosthetic-conscience** as the context-checkpointing skill and its checkpoint hooks; §15 records the corrections and the split from gray-area.
- **`gray-area.md`** — foundations for the fourth plugin, since built: trajectory mining, the 2.1.220 hook surface and what it lets us enforce, and a phased build. §4 records why continuity is *not* here.
- **`reasoning-telemetry.md`** — research record: what is observable in a session, and how to expose thinking summaries — the measured correction to "reasoning is not persisted", the three capture channels graded, and why a summary still cannot promote a finding.

Once a plan's design has shipped **and nothing live cites it by path**, it moves to
[`historical/`](historical/) — that directory's README carries the census that decides it. Moving a
plan is how a reader is told it describes a PAST tree; editing it to match today would destroy the
record of what changed, which is the thing a plan is for. A plan that shipped but is still cited as
the design of record stays here — `claude-port-plan.md` is the standing example, named by four
plugin READMEs, the repository README, `MEMORY.md` and a Go source comment.

The frank-exchange-of-views reform arc is its own cluster: `constitutional-reform.md` (the design), `change-waves.md` (the tracker it shipped through), `rulebook-audit.md` (the rules the evidence indicted), plus the record-layer plans (`record-*.md`).
