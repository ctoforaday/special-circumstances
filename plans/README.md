# Plans

Planning and design artifacts for Special Circumstances. Each arrives as its own pull request for review; once agreed and built, the relevant pieces graduate into the plugins and the planning doc stays here as historical record. The directory therefore holds both live proposals and shipped plans — deliberately separated from the repo's "real" documentation (`README.md`, the per-plugin READMEs). **The authoritative state of any plan is the dated status marker at the top of its own file** — a `> STATUS` line (shipped — historical record / in progress / superseded / not started), or a retirement banner where a plan was retired outright — never this index, which is an entry point and not a register.

Entry points:

- **`claude-port-plan.md`** — the master plan (shipped — historical record): a first-principles teardown of the Antigravity origin, the harness comparison, the original three-plugin architecture (now four — gray-area joined 2026-07), the inter-agent contracts, and the phased build plan. Phase 4 (sleeper-service) remains unbuilt.
- **`memory-architecture.md`** — proposal, not started: an Open-Knowledge-Format, git-native memory architecture (trajectory→memory skills, a `/dream` consolidation loop, global + project stores).
- **`context-checkpointing.md`** — the live design, shipped in **prosthetic-conscience** as the context-checkpointing skill and its seal/restore/observe/re-arm hooks; Phase 4's sleeper-service wiring and R11 are the open threads. The archaeology — the corrections, the gray-area retarget and its reversal, and the acceptance and measurement runs — is in [`historical/context-checkpointing.md`](historical/context-checkpointing.md), which kept its original section numbers so existing §-citations still resolve.
- **`gray-area.md`** — foundations for the fourth plugin, **partly** built (v0.9.0 ships Phases 0-2 and 5; Phase 3 instrumented reasoning and Phase 4 bench symmetry remain): trajectory mining, the 2.1.220 hook surface and what it lets us enforce, and a phased build. §4 records why continuity is *not* here; §11.7-§11.11 are the #189 investigation the shipped capture schema is built on.
- **`reasoning-telemetry.md`** — research record: what is observable in a session, and how to expose thinking summaries — the measured correction to "reasoning is not persisted", the three capture channels graded, and why a summary still cannot promote a finding.

Once a plan's design has shipped **and nothing live cites it by path**, it moves to
[`historical/`](historical/) — that directory's README carries the census that decides it. Moving a
plan is how a reader is told it describes a PAST tree; editing it to match today would destroy the
record of what changed, which is the thing a plan is for. A plan that shipped but is still cited as
the design of record stays here — `claude-port-plan.md` is the standing example, named by three
plugin READMEs, the repository README, `MEMORY.md`, a Go source comment, `requirements.json`,
`.qlty/qlty.toml` and three sibling plans. **And citations are not only by filename:** live code and
sibling plans reach into these documents by SECTION (`§3a′`), by wave id (`W1.13`), by item number
and by risk id, so a filename census is the floor of the question and not the whole of it.

- **`record-protobuf.md`** — the live design: `Event.Payload`'s map became a generated protobuf
  schema, one message per event type under a `oneof`, with the enums, the requiredness annotations
  and the key census that gate it. Shipped; three small items remain. Archaeology — the textproto
  wire format that shipped and was then retired by the SQLite store, the five-stage read rule, the
  six audit rounds and the pre-change censuses — in
  [`historical/record-protobuf.md`](historical/record-protobuf.md).
- **`record-sqlite.md`** — the live design: the record IS a SQLite database, append-only, with the
  DDL derived from the protobuf descriptors. Shipped and in production; the remaining Go folds
  (`BoardState` above all) are the open thread. Archaeology — the sharded-JSONL storage it replaced,
  the cutover's defect autopsies and the decisions the cutover reversed — in
  [`historical/record-sqlite.md`](historical/record-sqlite.md).

- **`bench-rulings-first-class.md`** — in progress: the bench's disposition of a gap becomes a
  `docket` motion, id'd and joined to what it settled (#681). Scope 1 (#695) and Scope 3 (#702)
  shipped; the delivered halves and the audit record are in
  [`historical/bench-rulings-first-class.md`](historical/bench-rulings-first-class.md).

The frank-exchange-of-views reform arc is its own cluster: `constitutional-reform.md` (the design), `change-waves.md` (the tracker it shipped through), `rulebook-audit.md` (the rules the evidence indicted), plus the record-layer plans (`record-*.md`).
