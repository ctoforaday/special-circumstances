# Memory Architecture for Special Circumstances

**Status:** Proposal (for discussion) · **Scope:** all three plugins · **Author:** design review follow-up
**Supersedes:** the ad-hoc "stash it in CLAUDE.md / MEMORY.md / `memory:` frontmatter" position in `docs/claude-port-plan.md`

---

## 1. The question this answers

> *"Where does captured knowledge physically go? Into GitHub?"*

Today the port plan captures learning into three native surfaces — `CLAUDE.md`, `MEMORY.md`,
and per-agent `memory:` frontmatter — with no schema, no lifecycle, and no story for
*project vs. global* or *duplicate vs. new*. That is a pile, not an architecture. As the suite's
`sleeper-service` loop runs daily and `frank-exchange-of-views` accumulates audit findings, the
pile grows unbounded and un-navigable.

This proposal gives a **principled, git-native, portable** answer: knowledge lives in a
**versioned knowledge store on disk, in an open format**, and the native Claude Code surfaces
become *projections* of that store rather than the store itself.

**Design goals**

1. **One open format**, human-readable and agent-parseable, that survives being moved, cloned, or read without our tooling.
2. **Git is the substrate.** Every memory mutation is a commit; history, diff, blame, and PR review come for free.
3. **Compatible with native memory, not a parallel universe.** `CLAUDE.md` / `MEMORY.md` / `memory:` frontmatter are generated *views*, not competing stores.
4. **A lifecycle**: short-term trajectory notes are cheap and noisy; long-term knowledge is structured, deduplicated, and earns its place.
5. **Simple.** No daemon, no mount, no service to install. Files and git.

---

## 2. Prior art we learn from (and what we reject)

| Prior art | What it got right | What we take / drop |
|---|---|---|
| **Internal FUSE centralized-memory filesystem** (mounted memory as a live FS) | Unified namespace; one place to look; agents read/write memory like files | **Rejected as out of scope.** A local daemon / mounted FS is un-portable, un-reviewable, and platform-bound. We keep the *insight* (memory as a filesystem of files) and drop the *mechanism* (make it a real git directory, not a FUSE mount). |
| **OpenClaw "Dream Diary"** (`config/openclaw/workspace/DREAMS.md`) — a scheduled 3 AM pass that consolidated the day's fragments into prose | The **cadence** and the **consolidation-while-idle** metaphor are exactly right | Keep the scheduled consolidation pass ("dream loop"). Drop the free-verse output — our dream loop emits *structured, deduplicated knowledge*, not poetry. (The old diary's later entries degraded to "a memory trace surfaced, but details were unavailable" — a cautionary tale about consolidation with no schema.) |
| **AgentOrange `continuous_learning` DbC aspect** (`research/continuous_learning_research_1p.md`) | **Expand-existing-before-append** anti-redundancy rule; explicit BEFORE/During/AFTER capture; context-compression as a consolidation trigger; "an unregistered aspect is invisible" | Adopt wholesale as the **lifecycle discipline** (§6). This is the promotion ladder and the dedup rule, already battle-tested in the old repo. |
| **Google Open Knowledge Format (OKF v0.1)** — see §3 | A published, vendor-neutral spec for exactly this problem, and it is *just markdown + YAML frontmatter in directories* | Adopt as the **on-disk format**. |

---

## 3. The open format: OKF-inspired knowledge bundles

The [Open Knowledge Format](https://cloud.google.com/blog/products/data-analytics/how-the-open-knowledge-format-can-improve-data-sharing)
(OKF v0.1, Google Cloud) formalizes the "LLM-wiki" pattern into a portable standard. Its relevant properties:

- A knowledge bundle is **a directory of markdown files with YAML frontmatter**; each file is a "concept."
- The **only required frontmatter field is `type`**; `title`, `description`, `resource`, `tags`, `timestamp` are standardized-optional.
- Reserved filenames: **`index.md`** (progressive-disclosure entry point) and **`log.md`** (chronological history).
- Concepts **link to each other with ordinary markdown links**, forming a knowledge graph richer than the directory tree.
- It is *format, not platform*: "just markdown, just files, just YAML" — readable in any editor, renderable on GitHub, shippable as a tarball, **hostable in a git repo**. No SDK, no runtime, no account.

This is almost exactly the shape Claude Code memory already wants (markdown + frontmatter), which
makes OKF a near-zero-friction fit and gives us an *external, documented standard* to point at
instead of a bespoke schema. **We adopt OKF as our storage format, with a small profile
(a fixed `type` vocabulary and a few extra fields) for the memory lifecycle.**

### 3.1 Our OKF profile — the memory concept schema

Every long-term knowledge file:

```markdown
---
type: rule            # rule | fact | preference | glossary | howto | pitfall | insight
title: Expand existing memory before appending
description: One-line summary shown in indexes and progressive disclosure.
scope: global         # global | project
status: active        # candidate | active | deprecated
confidence: 0.8       # 0.0–1.0, raised on corroboration, lowered on staleness
tags: [memory, discipline]
provenance:
  - source: trajectory:2026-07-10-a1b2         # trajectory:<session> | url:<u> | file:<p>
    captured: 2026-07-10T14:03:00Z
    by: skill/trajectory-review
last_seen: 2026-07-11        # bumped whenever re-observed or re-confirmed
review_count: 4              # times corroborated; feeds decay/promotion
supersedes: []               # ids of concepts this merges/replaces
---

# Rule
BEFORE adding a new concept, YOU MUST attempt to expand an existing one first.

# Rationale
Prevents knowledge fragmentation (see [continuous-learning](../rules/continuous-learning.md)).

# Evidence
- trajectory 2026-07-10: appended a duplicate "terse output" note; merged on review.
```

Only `type` is strictly required (OKF compliance); the lifecycle fields (`status`, `confidence`,
`last_seen`, `review_count`, `supersedes`, `provenance`) are our profile's additions and drive §6.
A file missing them is still a valid OKF concept — it is simply treated as a low-confidence
candidate until the dream loop enriches it.

---

## 4. Physical storage — where it lives

Two stores, both **plain git-tracked directories**, mirroring Claude Code's own global/project split.

```
# GLOBAL store — one per machine/user, its own git repo (optionally a private GitHub repo)
~/.claude/knowledge/
├── index.md                     # OKF entry point: what's here, how it's organized
├── short-term/                  # cheap, noisy capture surface — decays
│   └── 2026-07-11.md            # dated trajectory notes (append-only within a day)
├── knowledge/                   # consolidated long-term concepts (OKF bundle)
│   ├── rules/ facts/ howto/ pitfalls/ glossary/
│   └── index.md
├── projections/                 # GENERATED views — never hand-edited
│   ├── active.md                # active global concepts, @-imported by ~/.claude/CLAUDE.md
│   └── agents/red-auditor.md    # per-agent memory: bundle projection
└── .knowledge.toml              # store config: decay windows, dedup thresholds, scopes

# PROJECT store — lives INSIDE each project repo, committed with the code it describes
<project>/.claude/knowledge/
├── short-term/  knowledge/  projections/active.md
└── index.md
```

**Why this placement**

- **Global store = its own repo** so knowledge that spans projects (your working preferences, cross-project pitfalls, tool cheatsheets) is versioned and portable independent of any one codebase. Push it to a private GitHub repo → answered: *yes, into GitHub*, but as a first-class knowledge repo, not smuggled into CLAUDE.md.
- **Project store = committed with the project.** Knowledge about *this* codebase travels with it, is reviewed in the same PRs, and is cloned by every collaborator. This is the OKF "hostable in a git repo" property used literally.
- **`short-term/` vs `knowledge/`** is the ephemeral/durable split (§6). Short-term is allowed to be messy and is periodically pruned; `knowledge/` is the curated bundle.
- **`projections/` is generated, never authored.** This is the compatibility bridge (§5).

---

## 5. Mapping to native Claude Code memory (compatibility, not replacement)

The store is the source of truth; the native surfaces are **projections generated from it**, so
nothing we build fights the harness.

| Native surface | Role in this architecture | Direction |
|---|---|---|
| **`MEMORY.md`** (auto-memory) | The **inbox**. What the harness auto-captures mid-session lands here as raw short-term notes. The dream loop *reads MEMORY.md*, promotes what's durable into `knowledge/`, and prunes it back down. | store ← MEMORY.md (ingest) |
| **`CLAUDE.md`** (project + user) | The **always-on projection.** `CLAUDE.md` contains a single `@./.claude/knowledge/projections/active.md` import (project) / `@~/.claude/knowledge/projections/active.md` (global). Active, high-confidence concepts render into context every session. Regenerated by the dream loop. | store → CLAUDE.md (project) |
| **`memory:` frontmatter** (per-agent) | Each agent that declares `memory:` (e.g. FEOV's `red-auditor`) gets a **scoped sub-bundle** under `knowledge/agents/<agent>/`. Its projection is what the harness injects as that agent's persistent memory. | store ↔ agent memory |
| **Skills** (`skills/`) | A concept that graduates to `type: rule` with high confidence can be promoted a further rung into a **rule-skill** (prosthetic-conscience's existing `skills/rules/` corpus). Highest tier of the promotion ladder. | store → skill (top rung) |
| **`SessionStart` hook** | Injects the freshly-generated `active.md` projection (belt-and-suspenders with the `@`-import) and can surface a one-line "N candidate memories pending review" nudge. | store → context |
| **`PreCompact` / `Stop` hook** | Fires the single-trajectory capture (§7.1) so a session's insights are captured *before* the transcript is compacted or ends. Directly implements the old aspect's "context-compression as a consolidation trigger." | trajectory → store |

Net effect: an operator who never installs our tooling still sees ordinary `CLAUDE.md` and
`MEMORY.md` files. Our skills make them *coherent*; they do not make them *mandatory*.

---

## 6. Lifecycle: promotion, dedup, decay

A single promotion ladder, lifted from the `continuous_learning` DbC aspect and made physical:

```
trajectory / MEMORY.md note   →   short-term/<date>.md      (candidate, confidence low)
      →   knowledge/<type>/<slug>.md   (active concept, corroborated)
      →   projections/active.md → CLAUDE.md   (always-on)
      →   skills/rules/<rule>/   (rule-skill; top rung, human-gated)
```

**Promotion criteria (dream loop, §7.4):** a short-term note is promoted to an active concept when
it is **corroborated** (seen in ≥2 trajectories, i.e. `review_count ≥ 2`) or explicitly confirmed
by the operator. Promotion to `active.md` requires `status: active` and `confidence ≥ threshold`
(default 0.7). Promotion to a rule-skill is **always human-gated** — it mutates shipped plugin
behavior (Semantic Consent, per the port plan's guardrail).

**Dedup / merge discipline — expand-existing-before-append (the core rule).** Before writing any
new concept, the writer (skill or dream loop) MUST search the target bundle for an existing concept
covering the same ground and **expand it** (append evidence, bump `review_count`, raise `confidence`,
update `last_seen`) rather than create a near-duplicate. When two concepts are found to overlap, one
`supersedes` the other and the loser is set `status: deprecated`. This is the single most important
invariant; it is what keeps the store navigable instead of becoming the unbounded pile we started with.

**Decay / pruning.**

| Tier | Retention | Rule |
|---|---|---|
| `short-term/` | rolling window (default 14 days) | Dream loop deletes dated files older than the window *after* they've been through consolidation. Nothing is lost that wasn't first offered for promotion. |
| `knowledge/` candidate | 60 days without corroboration | `status: candidate` concepts whose `last_seen` is stale and `review_count` stayed 1 are set `status: deprecated`, then removed a cycle later. |
| `knowledge/` active | indefinite, but decays | `confidence` decays slowly if never re-observed; a decayed-below-threshold concept drops out of `active.md` (stops being always-on) but stays in the bundle for search. |
| `deprecated` | one cycle | Physically deleted next dream pass. Git history retains it — nothing is truly gone. |

Because every step is a git commit, "pruning" is safe: `git log`/`git revert` is the undo.

---

## 7. Component map

The five requested capabilities, each realized as a concrete Claude Code primitive, placed in the
plugin whose charter it fits (interactive → prosthetic-conscience; batch/background → sleeper-service).

| # | Capability | Primitive | Home plugin | What it does |
|---|---|---|---|---|
| 7.1 | Review one session trajectory → candidate memories | **Skill** `trajectory-review` + **`Stop`/`PreCompact` hook** + `/remember` command | prosthetic-conscience | Reads the current session transcript, extracts candidate insights/rules/pitfalls, writes them to `short-term/<date>.md` in OKF form. Runs automatically at session end/compaction (hook) or on demand (`/remember`). |
| 7.2 | Bootstrap memory off ALL past trajectories | **Command** `/memory-bootstrap` driving a **Workflow** (`bootstrap.js`) | sleeper-service | Batch-fans-out `trajectory-review` over every transcript under `~/.claude/projects/*/*.jsonl`, collects candidates, then runs one big consolidation (the dream loop's merge stage) to seed a store from history. One-time / occasional. |
| 7.3 | Consume an external source into memory | **Skill** `knowledge-ingest` + **Command** `/ingest <url\|file\|dir>` | prosthetic-conscience | Fetches/reads a doc, URL, or file tree; distills it into OKF concepts (`type: fact/howto/glossary`) with `provenance.source: url:/file:`; runs the same dedup discipline before writing. Feeds FEOV research too. |
| 7.4 | The **DREAM** loop | **Workflow** `dream.js` invoked by **Command** `/dream`, **scheduled daily** | sleeper-service | The consolidation pass (§7.5). |
| 7.5 | Project hierarchy alongside global, with precedence | **Skill** `project-memory` (extended) + the two physical stores (§4) + `.knowledge.toml` | prosthetic-conscience (already ships `project-memory`) | Defines the two-store layout, the precedence/merge rules (§8), and the projection-generation used by both `/dream` and `SessionStart`. |

Supporting agents (small, single-purpose, run inside the workflows):

- **`memory-consolidator`** (`memory: project`, so it learns the store's own shape over time) — the merge/dedup brain of the dream loop; given a batch of candidates + the existing bundle, decides expand-vs-append-vs-supersede and emits the edits.
- **`memory-curator`** — enforces the decay table, deprecates stale concepts, regenerates `projections/active.md`.

### 7.5 What `/dream` does (one pass)

1. **Gather** — collect `short-term/*` notes newer than last run + new `MEMORY.md` auto-memory content.
2. **Consolidate** — `memory-consolidator` merges candidates into `knowledge/`, obeying expand-existing-before-append; sets `supersedes`, bumps `review_count`/`confidence`, promotes corroborated candidates to `active`.
3. **Decay** — `memory-curator` applies the retention table (§6): deprecate stale, delete last-cycle deprecated, decay confidences.
4. **Project** — regenerate `projections/active.md` (global and each project) and per-agent memory projections.
5. **Commit** — one git commit per store (`dream: consolidate 2026-07-11 (+3 concepts, 2 merged, 1 pruned)`), so the whole pass is one reviewable diff. Optionally open a PR for the project store so a human ratifies changes to shared, committed knowledge.

### 7.6 Scheduling (ties to sleeper-service's daily cadence)

The dream loop reuses the port plan's existing scheduling story for `/self-improve` — **daily by
default, human-opt-in, manual always available**. `docs/scheduling.md` in sleeper-service already
documents the recipes; `/dream` slots in beside `/self-improve`:

- **Windows Task Scheduler → `claude -p "/dream"`** (this box) — the reference recipe.
- **cron → `claude -p "/dream"`** (POSIX).
- **Cloud agent / routine** (`/schedule`) for machine-independent runs.

Consolidation runs while you're not working — the OpenClaw "3 AM dream" cadence, now producing
structured knowledge instead of a diary. `/self-improve` and `/dream` are complementary: the former
evolves the *rules*, the latter consolidates the *knowledge*; both write only to git-tracked stores
and both gate rule/skill promotion on the human.

---

## 8. Project vs. global precedence and merge

Modeled on Claude Code's own memory hierarchy (nearest scope wins) so it is intuitive to anyone who
already understands `CLAUDE.md` layering.

**Read/merge order (later overrides earlier on conflict):**

```
global knowledge  →  project knowledge  →  (agent-scoped sub-bundle)
```

- **Additive by default.** Non-conflicting concepts from both scopes are all in context; project knowledge *adds to* global, it doesn't hide it.
- **Nearest scope wins on conflict.** If a project concept and a global concept share the same normalized `title`/topic and disagree, the **project** concept is authoritative *within that project*. The projection for that project renders the project version and annotates `overrides: global`.
- **Confidence breaks intra-scope ties.** Two concepts in the *same* scope never coexist as duplicates — the dedup rule (§6) forces a merge/supersede; higher `confidence` + `review_count` wins the merge.
- **Promotion crosses scopes deliberately, never silently.** A project concept that proves generally true can be *promoted to global* — but only via `/dream --promote-to-global` with human confirmation, because it changes behavior in *other* projects. The reverse (global → project specialization) is free and automatic.
- **Provenance is preserved across merges** so `git blame` + the `provenance` list always answer "where did this come from and why."

---

## 9. Open questions & risks

1. **Transcript format & location.** This design assumes session transcripts are readable JSONL under `~/.claude/projects/<slug>/*.jsonl` (consistent with where auto-memory `MEMORY.md` lives). The exact schema of those transcripts is **not something I've verified against current docs** — `trajectory-review` and `/memory-bootstrap` depend on it. *Phase 0 must confirm the on-disk transcript format before building 7.1/7.2.*
2. **Workflow ↔ scheduling ↔ headless.** `/dream` running under `claude -p` non-interactively, spawning a Workflow that spawns sub-agents, is plausible but unproven for this suite; the port plan already flags Workflow/CLAUDE.md-inheritance as undocumented. *De-risk in the Phase-1 harness spike.*
3. **Generated-file / hand-edit collision.** `projections/active.md` is generated but sits next to `CLAUDE.md` which humans do edit. Mitigation: a big `<!-- GENERATED by /dream — do not edit -->` banner + keep projections in their own directory (done) + `CLAUDE.md` only ever `@`-imports, never inlines.
4. **Dedup quality.** Expand-existing-before-append is only as good as the consolidator's overlap detection. Bad merges silently lose knowledge (the OpenClaw diary's "details unavailable" failure). Mitigation: every merge is a git diff a human can review; `supersedes` never hard-deletes for a full cycle.
5. **PII / secret leakage into a pushed global repo.** Trajectories contain paths, tokens, names. The `/dream` commit step MUST run the port plan's existing secret-scrub (`git grep` denylist) before committing, especially for a store pushed to GitHub. *Blocking for any remote push.*
6. **Confidence/decay tuning.** Thresholds (0.7 active, 14-day short-term, 60-day candidate) are guesses. They belong in `.knowledge.toml` and want real-usage tuning.
7. **OKF version drift.** OKF is v0.1 and evolving. We pin to a documented profile (§3.1) and treat upstream changes as opt-in.

---

## 10. Phased build plan

| Phase | Work | Verify |
|---|---|---|
| **0. Confirm substrate** | Verify transcript JSONL location/format; lock the OKF profile (§3.1) + `.knowledge.toml`; scaffold empty global + project stores | A hand-written OKF concept renders into context via `@`-import in `CLAUDE.md`; transcript of a real session is parseable |
| **1. Capture (single)** | `trajectory-review` skill + `/remember` command + `Stop`/`PreCompact` hook (prosthetic-conscience) | End a session → candidate concepts appear in `short-term/<date>.md`, correctly typed, with provenance |
| **2. Consolidate (dream core)** | `memory-consolidator` + `memory-curator` agents; `dream.js` workflow; `/dream`; projection generation (sleeper-service + project-memory) | Run `/dream` on seeded short-term notes → merges (no dup), promotes corroborated, regenerates `active.md`, one clean git commit |
| **3. Hierarchy** | Two-store precedence + merge rules; `overrides`/promote-to-global flow | Conflicting global vs project concept resolves project-wins in-project; `--promote-to-global` is human-gated |
| **4. Ingest + bootstrap** | `knowledge-ingest` + `/ingest`; `/memory-bootstrap` batch workflow | `/ingest <url>` yields deduped OKF facts; `/memory-bootstrap` seeds a store from history with no duplicate explosion |
| **5. Schedule + harden** | Daily `/dream` recipe in `docs/scheduling.md`; secret-scrub gate; decay tuning; optional project-store PR flow | Scheduled headless `/dream` runs, scrubs, commits; PII denylist clean |

Phases 1–2 are the minimum viable memory system; 3–5 make it multi-project, external-source-aware, and autonomous.

---

## 11. Alternatives considered

- **Do nothing (native surfaces only).** Simplest, and where the plan is today. Rejected: no lifecycle, no dedup, no hierarchy → the unbounded pile the reviewer flagged.
- **A single JSON/JSONL knowledge database** (one append-only log + index). More "database-like," easy to append. Rejected: not human-readable, doesn't render on GitHub, diffs are noise, and it fights `CLAUDE.md`'s markdown nature. OKF markdown gets us structured-yet-readable and free GitHub rendering.
- **FUSE / mounted memory FS (the internal prior art).** Powerful unified namespace. Rejected per scope: daemon-bound, un-portable, un-reviewable. Our git directory keeps the "memory is a filesystem" insight without the machinery.
- **SQLite + vector index for semantic dedup.** Would make overlap detection stronger than title-matching. Deferred, not rejected: it can sit *beside* the OKF store as a derived index (source of truth stays the files), if §9-item-4 dedup quality proves insufficient.

---

### One-line summary

Store knowledge as a **git-tracked, OKF-formatted bundle** (global repo + per-project directory);
let `CLAUDE.md`/`MEMORY.md`/`memory:` be **generated projections**; capture per-trajectory, ingest
external sources, and run a nightly **dream loop** (sleeper-service, daily) that consolidates,
dedups (expand-before-append), and decays — every step a reviewable commit.
