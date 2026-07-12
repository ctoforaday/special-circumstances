# Special Circumstances: Porting AgentOrange to a Claude-Native Plugin Suite

## Context

AgentOrange started as an "Antigravity Meta Brain": a Design-by-Contract rule corpus, an
adversarial research swarm (Lead/Researcher/Auditor), a spec-driven development standard, and an
hourly cron loop that researched *its own configuration*. Around it grew a local-LLM stack
(LiteLLM/oMLX/OpenClaw/Jaeger) that is **out of scope — not ported, not archived**. The old repo
stays behind as the historical record.

The port target is **Special Circumstances**: a new GitHub repo hosting a Claude Code
**marketplace of three plugins**, each named for the Culture:

| Plugin | Named for | Carries |
|---|---|---|
| **prosthetic-conscience** | The drone that keeps you honest | Core adversarial + cowork behavior: DbC rules, pair programming, SDD + plan-audit gate, project memory, proficiency skills, quality hooks |
| **frank-exchange-of-views** | A heated argument, diplomatically put | The research debate engine: blue team (builders), red team (auditors), debate + resolution model, best-of-N, Heilmeier Catechism, full artifact preservation |
| **sleeper-service** | The GSV quietly running vast hidden projects | Autonomous learning: self-improvement loop, graduation pipeline, continuous-learning promotion ladder, scheduling |

Install order: prosthetic-conscience is the base (the other two preload its rule-skills);
frank-exchange-of-views is required by sleeper-service (which invokes it). These are **structural
dependencies**, not just prose — Claude Code plugins support a `dependencies` array in `plugin.json`
(semver-pinned, auto-enabled when the dependent is enabled), so installing sleeper-service pulls in
its deps. One marketplace install gets all three; each remains individually useful.

---

## Part 1 — First principles: how the harnesses differ

### The core question

All harness config answers: *"how does standing instruction-content get into an agent's context —
and does it reach the children?"* Antigravity's answer had a hole that shaped the entire old
codebase: **rules, hooks, and plugin behavior applied ONLY to the main agent.** Subagents
inherited nothing — no rules, no hooks, no context — which is why `prompt_resolver.py` existed
(hand-compile the world into every child's prompt) and why the orchestrator needed SDK hooks to
auto-repair spawns that forgot their system prompts.

### What Claude Code inherits to subagents (verified against current docs)

| Concern | Antigravity | Claude Code (confirmed) |
|---|---|---|
| Project standing context | Main agent only | **Subagents inherit the full CLAUDE.md/memory hierarchy** (user + project + local). Exception: built-in Explore/Plan agents skip it for speed; our custom agents get it. |
| Plugin rules → children | Nothing; hand-compiled per spawn | **No automatic propagation** (same gap) — but the native fix is the **`skills:` frontmatter field**: skills listed there have their full content injected into the subagent's context at startup. Rules become skills; each agent role *declares* its rules in one YAML line. `prompt_resolver.py` reborn as declared inheritance. |
| Skills discovery in children | Not available | **Subagents can discover/invoke all skills** via the Skill tool (unless denied). |
| Hooks on child tool calls | Did not fire | **Verified empirically (Phase 1 spike, 2026-07): the parent's plugin PostToolUse hook DOES fire on a subagent's own Write** — so "the qlty gate holds when a swarm agent writes" is valid, as originally designed. (This overturned a docs-*reading* that claimed otherwise; leaf-node testing won.) Confirmed for Task/Agent subagents; the Workflow-tool swarm is expected to match (direct test pending). A plugin *subagent* still can't declare its **own** `hooks:` frontmatter — but the parent's hooks apply to it. Parent also sees `SubagentStart`/`SubagentStop`. |
| Nesting | Fragile, collision-prone | Subagents can spawn subagents to **depth 5**. |
| Per-agent persistent memory | None | **`memory:` frontmatter field** (user/project/local) — an agent can remember across sessions. |

Design consequences, baked into every agent definition in this suite:
1. Rules ship as **skills** (they're loadable prompt fragments — that's what a skill is).
2. Every agent definition carries `skills: [...]` naming the rule-skills it runs under. No
   compiler, no registry; the filesystem is still the source of truth, natively.
3. Agent bodies are **self-contained** (@-imports in agent files are unsupported — the `skills:`
   mechanism replaces them).
4. The red-team auditor gets `memory: project` so it accumulates knowledge of gap patterns it has
   caught before — continuous learning wired into the adversary itself.
5. **A subagent does not auto-inherit a plugin's skills or injected content** — rule content reaches
   a child *only* via that child agent's `skills:` frontmatter (verified 2026-07: the probe agent
   carried exactly its declared rule-skills and quoted the `critical-stance` probe verbatim). **But
   the parent's plugin hooks DO fire on the child's tool calls** (also verified live) — so the
   quality/verification gate rides PostToolUse on swarm writes as originally designed; no lead-run
   sweep is needed as a fallback. (A child still can't declare its *own* `hooks:`.)
6. **Roles-as-mindsets, stated precisely (settled across two review rounds): mindsets live in
   shared rule-skills; agents are per-plugin DUTY specialists.** The generic thing is the mindset
   (critical-stance is the auditor base class, preloaded cross-plugin via namespaced `skills:`);
   an agent file is that mindset's thin binding to one duty — tools + memory + envelope contracts.
   One agent per duty, not per knowledge domain (plan-auditor = the stateless plan gate;
   red-auditor = the debate participant with living findings and project memory — same inherited
   mindset, different duties, deliberately not merged). Baking a standard into a single-duty agent
   is fine; proliferating agents per knowledge domain is not.

### Rules / skills / discovery mapping (unchanged from prior draft, condensed)

- Antigravity `trigger: always_on` → CLAUDE.md (+SessionStart hook injection for plugin consumers).
- Antigravity `trigger: model_decision` → a Claude **skill** (description-matched, lazy-loaded) —
  literally the same concept, natively.
- Antigravity `trigger: glob` → hooks (deterministic) or file-type-scoped skill descriptions.
- `ensure_skills.py`/`skills.json`/`activate_skill` → obsolete: native auto-discovery,
  progressive disclosure, no build step.
- OpenClaw for contrast: plugins are *code*; content arrives via workspace file conventions
  (persona files + skills dir scanned at session start). Claude plugins are *content bundles*
  (`commands/`, `agents/`, `skills/`, `hooks/`) distributed via git + `marketplace.json`.

### Verdict (unchanged): port the methodology, delete the machinery

Ship: the protocols, gates, templates, and rules. Delete as harness-redundant: the Lead agent
(a workflow script is the lead), watchdogs/heartbeats, `JSON BEGIN/END` A2A markers (schema
outputs), compaction sniffing, shadow workspaces, prompt/skill compilers, spawn toolchain,
anti-loop detectors (787-line orchestrator → ~100-line workflow). Differentiation from built-in
deep-research: **repo-aware, artifact-producing, adversarially gated, debate-preserving**.

---

## Part 2 — Inter-agent communication contracts (the blackboard model)

Antigravity's cross-agent pain: children returned raw contents to the parent, or you hand-built
blackboards/shared filesystems because agents were separate conversations with separate scratch.
Claude Code dissolves the *mechanism* problem — **workflow agents run with isolated contexts but
a shared working directory** (confirmed) — but the *contracts* must be explicit:

**The three principles:**
1. **The filesystem is the blackboard.** All swarm artifacts live in the git-tracked research run
   directory. Consequence: swarm agents must NOT use `isolation: worktree` (it severs the shared
   filesystem).
2. **The payload is the file; the schema is the envelope.** No large content ever travels through
   an agent return value. Agents write documents to the blackboard and return small structured
   handles.
3. **The lead is a script, not a conversation.** Control flow, round-keeping, diff computation,
   and termination are deterministic workflow code; semantic judgment calls (adjudicating
   rebuttals) are delegated to a small lead-judge agent.

**The envelopes:**
- Blue → Lead: `{ path, tldr, claim_count, saturation_reached, open_questions[] }`
- Lead → Red: the blackboard path to the **full living report** (red always re-reads it in
  context). A change-summary since the last audited snapshot MAY ride along as a *navigation hint*
  — never as a replacement for the document (diffs mislead on research prose).
- Red → Lead: `{ verdict: PASS|FAIL, gaps: [{id, location, problem, required_fix, severity,
  likelihood, impact, complexity_cost}], corroboration: [{claim, reference, confidence}],
  citations_checked, notes }` — findings are graded, not binary (see "Red-team judgement" in §3b).
- Lead-judge → transcript: `{ gap_id, resolution: closed|rebuttal_sustained|risk_accepted|carried|unresolved,
  rationale }` — `risk_accepted` = a valid finding blue rejects on a likelihood/impact/complexity
  tradeoff; recorded in the report, never silently dropped.

---

## Part 3 — The three plugins

### 3a. prosthetic-conscience (core: how we work)

The base plugin. Adversarial-partner behavior for interactive work, plus the shared rule
substrate the other plugins preload.

```
plugins/prosthetic-conscience/
├── .claude-plugin/plugin.json
├── skills/
│   ├── rules/                        # the DbC corpus, one skill per rule (preloadable)
│   │   ├── critical-stance/          # "not a yes-man"
│   │   ├── anti-spinning/            # 3-strike + honor-cancel
│   │   ├── semantic-consent/         # intent needs consent; syntax is agent's domain
│   │   ├── context-efficiency/  plan-act-reflect/  refactoring-safety/
│   │   ├── scratch-policy/  think-around-problem/  terse-communication/
│   │   ├── agent-guardrails/  dbc-authoring/
│   ├── pair-programming/             # driver/navigator, ping-pong TDD, hold-on-submit (typos fixed)
│   ├── spec-driven-development/      # SDD I–V + plan_template.md + tinyspec_template.md
│   ├── project-memory/               # 4-artifact projects/<name>/ workspace discipline
│   └── proficiency/{git,markdown,qlty}/   # SKILL.md + agent-maintained CHEATSHEET.md
├── agents/plan-auditor.md            # SDD gate: Alignment/Completeness/Safety (JSON verdict)
├── commands/{plan-audit, doctor}.md  # /plan-audit [file]; /sc-doctor — environment preflight
├── requirements.json                 # declared tool deps: name/purpose/required?/min-ver/per-OS install
└── hooks/hooks.json                  # qlty fmt/check on Edit|Write — CAPABILITY-GATED (no-op + warn if qlty absent); SessionStart probe
```

*Confirmed (decision resolved): SDD + plan-audit live here — they're interactive adversarial
cowork, and FEOV research reports cross-link into SDD plans.*

*Review follow-up — validation must outlive the spec (raised on §3a):* an SDD plan is loaded only
while a task references it, but the **validation discipline** (what the checks are and how to run
them) has to stay available even after work drifts *beyond* the plan — otherwise you lose the loop
exactly when you leave the plan's scope, which is when it matters most. Design answer: the
verification loop ships as its own **always-on rule-skill** (`validation-loop`) that SDD *invokes*
but that also stands alone in the core plugin, so it is discoverable and loaded independent of any
active spec. A separate plugin is overkill here — a core rule-skill already "stays loaded." The
adjacent concern — *surviving a context-compression event* so you don't forget the loop mid-task —
is genuinely separate and is being worked as its own follow-up PR (the Memento checkpoint proposal).

#### 3a′. Environment preflight — the "install problem"

The plan leans on local tools (`qlty` for the quality gate, `git`/`gh` for the blackboard and PR
flow). That toolchain will **not** always exist — qlty is absent on the primary dev box today. A
hook that shells out to a missing binary hard-fails on every Edit/Write, which is strictly worse
than no hook. The suite must **detect, report, and degrade — never assume.** This is Design-by-
Contract at the environment boundary: *BEFORE a tool-dependent hook runs, its binary MUST exist*;
the failure mode is graceful degradation, not a crash. Three layers, all in prosthetic-conscience:

1. **Requirements manifest + shared checker** (`requirements.json`, one per plugin): each tool
   declares `{ name, purpose, tier: required|recommended|optional, min_version, check_cmd,
   install: {windows, macos, linux} }` — the single source of truth. A small shared **toolchain
   probe** (one helper module, e.g. `lib/toolchain.*`) reads it and answers "is tool X present and
   version-OK?" — the *only* implementation of the presence check, imported by every tool-dependent
   hook AND by `/sc-doctor`. Knowledge of our local dependencies lives in exactly one place (per
   review: a central `toolchains` module, not a check hand-rolled into each hook).
2. **`/sc-doctor` command** (environment preflight, à la `flutter doctor`): aggregates every plugin's
   manifest, probes each tool (presence + version), prints a per-tool ✓/✗ report with the exact
   install command for the current OS, and a summary verdict — **READY / DEGRADED / BLOCKED**.
   Read-only by default.
3. **Capability-gated hooks + SessionStart nudge**: every tool-dependent hook calls the shared
   toolchain probe *first*; if the tool is absent it **no-ops with one intelligent warning** (names
   the tool, why it was skipped, and `/sc-doctor`) instead of erroring. This is a **hard blocker for
   porting the real qlty hook** — an ungated hook fails on *every* Edit/Write on a box without the
   formatter (exactly this box today) — so the gate ships **with** the hook, not later. A
   SessionStart hook runs the same probe once and surfaces a single non-blocking line if a
   recommended tool is missing. Fast, non-fatal.

**Installation stays human-gated (Semantic Consent):** `/sc-doctor` may *offer* to run the install
command it printed (`/sc-doctor --fix`, one tool at a time with confirmation), but installing mutates
the machine — explicit go-ahead required, never auto-run at session start. This box (Windows) is
the first test case: qlty ships a PowerShell installer; git/gh are already present. The manifest
carries per-OS install strings so the report is actionable wherever the suite lands. **Separation
of concerns (per review):** graceful hook degradation is *core* and ships with the hooks; the
actual formatter installation and the `/sc-doctor --fix` installer flow are a *separable* concern that
can land as their own PR — the shared toolchain probe is the seam between them.

**Hook implementation & install path — Go, tested, prebuilt.** Hooks fire constantly and *must not*
break (the historical failure mode), so they are **compiled Go binaries**, not shell (untestable)
or interpreted scripts (fragile, runtime-dependent). Each hook is a pure, table-driven-tested Go
command that reuses the shared toolchain probe. A GitHub Actions matrix runs `go vet` / `go test` on
Windows/macOS/Linux on every change — can't merge red, the guarantee shell can't give — and on a
`{plugin}--v{version}` tag (the same tags that drive plugin `dependencies`) cross-compiles all
platforms and publishes static binaries + SHA256s to the GitHub Release. **The target box never
needs Go:** `/sc-doctor --fix` builds via `go build` if Go is present, else fetches the prebuilt
release asset for this GOOS/GOARCH (checksum-verified) into `${CLAUDE_PLUGIN_ROOT}/bin/`, which
`hooks.json` invokes. Graduation: publish to scoop/winget/brew so the per-OS `install` string is a
one-liner and PATH is handled for us. Until provisioned, the hook capability-gates and no-ops with a
warning. The one non-Go seam — the installer (chicken-and-egg: it runs before any binary exists) —
stays `/sc-doctor`-driven (no committed script to rot) and self-verifying (re-probes after fetching).
Proven end to end on the Windows dev box in Phase 1: `sc-quality-gate` builds, `go test` green, and
it degrades cleanly when qlty is absent.

**Hook inventory — every hook the suite ships or plans, tracked here so none get lost in review
threads.** All follow the Go-toolchain rules above (tested, capability-gated, degrade-never-fail):

| Hook binary | Event / matcher | Purpose | Status |
|---|---|---|---|
| `sc-quality-gate` | `PostToolUse(Write\|Edit)` | qlty fmt + check on every write (fires in subagents too — verified); skip+warn when qlty absent | ✅ shipped (Phase 1) |
| `sc-secrets-gate` | `PreToolUse(WebFetch\|WebSearch\|Bash)` | **privilege-boundary for data**: block outbound payloads matching secret/token/key patterns before they leave the box; deterministic layer of agent-guardrails (the rule stays as the semantic layer a regex can't cover) | Phase 2 |
| `sc-toolchain-nudge` | `SessionStart` | one non-blocking line when a recommended tool is missing (runs the shared toolchain probe once) | Phase 2 |
| checkpoint seal/restore | `PreCompact` + `SessionStart(compact\|resume)` | seal the agent-authored CHECKPOINT.md before compaction; re-inject a terse digest after | tracked in the Memento proposal (PR #3) — builds on this toolchain if adopted |

### 3b. frank-exchange-of-views (the debate engine)

Given a topic, produce a **verified research deliverable with the full adversarial record
preserved**. This is the deepest redesign, per review feedback.

**Roles — a strict division of labour (the core correction from review):**
- **Blue team (builders) — ADDITIVE ONLY.** Researcher agents whose synthesis is *union, not
  summary*. Best-of-N: N parallel lanes/candidate drafts; a blue synthesizer **merges by
  inclusion** — deduplicating overlapping claims but never dropping substantive content — so the
  living report *broadens and deepens* each round (more detail, scope, and avenues of research),
  never narrows. Blue may add, dedup, and reorganize; it may **not** subtract substance. Preloads:
  research-protocol, critical-stance, think-around-problem rule-skills.
- **Red team (auditors) — the SUBTRACTIVE/critical voice, and the gate-keeper.** Adversarial
  reviewers. Best-of-N: M parallel audit passes with distinct lenses (leaf-node citation
  verification / logic & completeness / dark-side & risk); a red merger consolidates into the
  living findings report with one binary verdict. **Red — not the lead — decides when blue has met
  the bar** and issues PASS. Every proposal to cut, caveat, or flag weak evidence originates here.
  `memory: project` — the red team remembers gap patterns across sessions.
- **Lead — mechanics + final compromise only.** The workflow script runs control flow, round-
  keeping, and termination; a small **lead-judge** agent is invoked **only at the end** to build
  the compromise across residual red/blue disagreement (rebuttals blue sustains but red still
  contests). The lead does *not* adjudicate round-to-round — passing is red's call.

**Red-team judgement — trust and risk are *graded*, not binary (review directive):** two
dimensions run underneath every finding, and they change what "resolved" means.
- **Corroboration confidence.** Red doesn't just mark a claim true/false — for each
  *statement ↔ reference* pair it assigns a **confidence** that the source actually corroborates the
  statement (facts are rarely black and white). Low confidence → "needs more evidence, blue digs
  further," not an automatic fail. **The human is an untrusted source too:** a claim asserted only
  because the operator said so in a comment is *not proven* — if blue leans on "because you told
  me," red flags it and demands independent corroboration.
- **Likelihood × impact × complexity-cost.** Every gap/risk carries an estimated **probability** of
  being hit, an **impact** if it is, and the **complexity** mitigating it would add. This gates the
  anti-edge-case rule: *interesting ≠ of interest.* Auditors surface low-probability / low-impact
  findings but **must not force** blue to absorb complexity that makes the design strictly worse to
  satisfy them. Security is always a tradeoff (usability, performance, and human-maintainability all
  trade against it); the pair's job is to **elevate** the risk, reason about the tradeoff, and
  propose mitigations — even partial ones — then make a call. **Blue may legitimately reject a valid
  finding** on tradeoff grounds; the risk is still *documented* with its likelihood/impact and the
  rejection rationale, never silently dropped. Complexity has a price, and keeping the code the best
  documentation of the system is itself a goal worth protecting.

**Artifact layout — nothing is summarized away** (one directory per run, all git-tracked):

```
research/<date>_<slug>/
├── report.md              # the final agreed deliverable (assembled last)
├── blue/
│   ├── report.md          # blue's LIVING report — grows every round (additive; never summarized away)
│   ├── CHANGELOG.md       # what blue changed each round (keeps debate.md argument-focused)
│   └── candidates/        # best-of-N raw drafts and lane outputs, preserved
├── red/
│   ├── findings.md        # red's LIVING audit report — updated every round
│   └── candidates/        # per-lens audit passes, preserved
└── debate.md              # the FULL three-party transcript — every round, not just the finale
```

**The debate loop — runs until resolved, not for a fixed count** (review: 3 rounds was too few,
even 5 short for deep research; termination is a *judgement*, not a counter):
1. **Blue** produces/updates `blue/report.md` (round 1: frontier hypotheses → best-of-N lanes →
   additive synthesis; later rounds: expand and repair in response to red's gaps, rebuttals
   allowed) and logs the concrete edits to `blue/CHANGELOG.md`.
2. **Red** audits the **full living report, read in context** — not a bare diff: this is research
   prose, where a decontextualized diff misleads. The lead may hand red a change-summary as a
   *navigation hint*, but red always re-reads the whole document. Red updates `red/findings.md`
   (verdict + gap list) and **owns the PASS/FAIL call**.
3. **Lead** appends both positions to `debate.md`, then checks the exit condition:
   - **Red issues PASS** → assemble the final report.
   - **Deadlock / spinning** (lead-judge sees no *new* substantive gaps, or the same rebuttals
     recycling — the anti-spinning honor-cancel signal) → lead-judge builds the final compromise
     across the standing disagreement; assemble with unresolved items stamped `UNVERIFIED` and
     recorded. The gate never soft-passes.
   - **Otherwise** → another round. A generous safety ceiling (configurable, default high) only
     bounds runaway cost; the real terminator is *red-PASS-or-deadlock*, judged rather than counted.

**Debate transcript (`debate.md`) — the FULL record, appended every round** (paired with
`blue/CHANGELOG.md`, which carries the mechanical edit log so the transcript stays argument-focused):

```markdown
## Round 1
### RED (opening audit)
VERDICT: FAIL — 4 gaps. [gap table]
### BLUE (response)
Gap R1-1: accepted — §3 revised, new primary source added.
Gap R1-2: REBUTTAL — cited source is the RFC itself; challenge misreads §4.2.
### LEAD (resolution)
R1-1 closed. R1-2 rebuttal sustained (judge rationale: ...). R1-3, R1-4 carried.
```

**Final report structure (`report.md`)** — assembled by the lead by **UNION, not summary**: it
concatenates and cross-links the teams' full reports and the debate record. It must never compress
the research into a one-page digest — that was Antigravity's characteristic failure (ask it to
merge three documents and it truncates or summarizes one away). The research is *for the human*;
the summary is not the deliverable. Sections:

1. **TL;DR + Verdict** (VERIFIED/UNVERIFIED stamp, rounds taken)
2. **Heilmeier Catechism** — the agreed answers, after the summary, before the meat: What are we
   trying to do (no jargon)? How is it done today and what are the limits? What is new here and
   why will it succeed? Who cares? If it succeeds, what difference does it make? What are the
   risks? What does it cost? How long? What are the mid-term and final checks for success?
3. **Technical Foundations / Analysis / Risk Matrix** — the meat (per the Gold Standard template);
   the Risk Matrix grades each risk by **likelihood × impact × complexity-to-mitigate**, records
   mitigations (even partial ones), and lists tradeoff-**rejected** risks with rationale (elevated
   and documented, not dropped)
4. **Blue Team report** — included in full detail
5. **Red Team findings** — included in full detail, final verdict and gap dispositions
6. **Debate record** — round summaries + pointer to `debate.md`
7. **Semantic footnotes** — `[^WordLabel]` with Title/Author/URL/AccessDate (numeric deprecated)

```
plugins/frank-exchange-of-views/
├── .claude-plugin/plugin.json
├── agents/{blue-researcher, red-auditor, lead-judge}.md
├── skills/gold-standard-research/
│   ├── SKILL.md                      # protocol: frontier, saturation, ≥20% disconfirming budget
│   ├── report_template.md  debate_template.md  heilmeier_template.md
│   └── workflow.js                   # the debate engine (~100 lines)
└── commands/research.md              # /research <topic> [--lanes N] [--lenses M] [--max-rounds N]
```

### 3c. sleeper-service (autonomous learning)

The system that improves the system, while you sleep — always human-gated at the promotion step.

```
plugins/sleeper-service/
├── .claude-plugin/plugin.json
├── skills/continuous-learning/       # promotion ladder: insight → MEMORY → rule-skill → cheatsheet;
│                                     # expand-existing-before-append; DbC encoding discipline
├── commands/
│   ├── self-improve.md               # enumerate own rules/skills/agents/ideas/research → pick one
│   │                                 # → /research "how should X evolve?" → idea stub w/ alternatives
│   └── graduate.md                   # idea → research → project promotion (human approves each step)
└── docs/scheduling.md                # default cadence: DAILY; recipes: manual, Windows Task Scheduler + `claude -p`, cloud agent
```

Guardrail preserved: the loop writes only `research/` and `ideas/`; promotion into rules/skills
requires the human (Semantic Consent). **Depends on** frank-exchange-of-views (and on
prosthetic-conscience for its rule-skills), declared structurally in `plugin.json`'s `dependencies`
array (semver-pinned via `{plugin}--v{version}` git tags; enabling sleeper-service auto-enables its
deps) — not left to `/sc-doctor` to police.

### Repo layout

```
special-circumstances/
├── .claude-plugin/marketplace.json   # lists the three plugins
├── plugins/{prosthetic-conscience, frank-exchange-of-views, sleeper-service}/
├── .claude/settings.json             # dogfood: repo enables its own plugins
├── CLAUDE.md                         # THIN index: group identity + always-on @-imports only — rules live as discrete skills, never inline
├── README.md                         # the Special Circumstances story
├── ideas/  research/  projects/      # EMPTY scaffolding on publish — seeded fresh via /research (clean start)
├── MEMORY.md                         # design-rationale log, PII scrubbed
└── .qlty/qlty.toml
```

---

## Part 4 — README story (portfolio framing)

1. **Hook**: "Special Circumstances — an adversarial human/AI methodology suite. It argues with
   you on purpose, and it researches its own rules while you sleep."
2. The three plugins, each with its Culture epigraph and its living implementation.
3. **The deletion table** — hand-rolled Antigravity machinery → Claude-native equivalent
   (787-line orchestrator → ~100-line workflow; compilers → native discovery; A2A markers →
   schemas; watchdogs → harness). The engineering-judgment centerpiece.
4. Install (marketplace URL) + dogfood tour (`/research`, watch the debate transcript grow).
5. Origins paragraph linking the old repo. No dead code shipped.

---

## Part 5 — Phases & verification

| Phase | Work | Verify |
|---|---|---|
| **0. Bootstrap — ✅ DONE** | Private `ctoforaday/special-circumstances`; marketplace + three plugin manifests; dogfood `settings.json`; thin CLAUDE.md; README; MEMORY.md; MIT LICENSE; `.qlty`; empty clean-start dirs | ✅ **Verified live**: `/plugin marketplace add ctoforaday/special-circumstances` succeeded and all three plugins installed. (Plan artifacts kept as PRs under `plans/`.) |
| **1. Harness spike — ✅ DONE (verified live on Windows)** | Seed skill (`critical-stance`) + communication model (`terse-communication`, `design-by-contract`) + `probe` agent (preloads them via `skills:`) + `/probe` + `/sc-doctor` + a **tested Go hook toolchain** (`sc-quality-gate` binary + shared toolchain probe + CI matrix + `/sc-doctor` build/fetch) — `go test` green (PR #5, merged) | ✅ **All five green:** `/sc-doctor` + rule-skills load; `skills:` preloads into the subagent (probe quoted `critical-stance` verbatim); the Go hook fires on **both main-agent AND subagent** writes; capability-gates (qlty absent → skip+warn, exit 0); `${CLAUDE_PLUGIN_ROOT}/bin/sc-quality-gate` resolves on Windows. **Finding: plugin hooks DO fire in subagents** (overturned the docs-reading — see Part 1). |
| **2. prosthetic-conscience** | Rule corpus → skills (chunk 1: PR #8); pair-programming, SDD, project-memory, proficiency; plan-auditor + /plan-audit; **hooks per the §3a′ inventory: `sc-secrets-gate` (PreToolUse outbound-secrets block) + `sc-toolchain-nudge` (SessionStart)** | critical-stance probe gets pushback; /plan-audit returns schema verdict on a real plan; **sc-secrets-gate blocks a seeded fake token in an outbound payload (unit + live test)**; on a qlty-absent box edits still succeed (hook no-ops + warns, /sc-doctor reports the install command); qlty clean where present |
| **3. frank-exchange-of-views** | Agents (blue/red/judge), templates (report/debate/Heilmeier), workflow, /research | **E2E dogfood** on a fresh real topic: run dir contains living blue+red reports, candidates, debate.md with 3-party rounds; final report embeds both team reports + Heilmeier; seed a bad citation → FAIL round → diff-based re-audit → resolution recorded in transcript |
| **4. sleeper-service** | continuous-learning skill; /self-improve (daily default); /graduate; scheduling docs | Headless `claude -p "/self-improve"` produces a run dir + idea stub; touches only research/+ideas/ |
| **5. Story + publish** | **Clean start — no corpus import.** README; empty ideas/research/projects scaffolding; PII/secret scrub of skills+docs; LICENSE; publish | Fresh clone → README install steps verbatim; `git grep -iE "jarvis|8070926388|0xDEADC0DE"` clean; final `/research` smoke populates the first real run |

~6–8 working days (up ~1 day for the richer FEOV artifact/debate model). Main risks: (a) exact
plugin mechanics for SessionStart injection and cross-plugin skill preloading — Phase 1 spike
de-risks both; fallback for cross-plugin preloads is vendoring copies of needed rule-skills into
FEOV/SS; (b) Workflow-agent CLAUDE.md inheritance is undocumented — spike verifies; fallback is
explicit `skills:` preloads carrying everything the swarm needs (already the design). **(c) RESOLVED
(Phase 1 spike, verified live): plugin hooks DO fire on subagent writes — the qlty gate rides
PostToolUse on swarm writes as designed. Verified for Task/Agent subagents; the Workflow-tool swarm
is expected to match, a direct test pending.**

---

## Resolved decisions

1. **Structure**: one repo (`special-circumstances`), one marketplace, **three plugins**:
   prosthetic-conscience (core/adversarial/cowork), frank-exchange-of-views (research debate
   engine), sleeper-service (autonomous learning).
2. **Naming**: the group is **Special Circumstances**; plugins carry Culture names above.
3. **Swarm shape**: deterministic workflow lead + best-of-N blue/red teams + lead-judge for
   adjudication; full artifact preservation (living team reports, candidates, debate transcript);
   Heilmeier Catechism in the final report between summary and meat.
4. **Evidence corpus**: **clean start** — ship no historical corpus; `ideas/ research/ projects/`
   are empty scaffolding seeded fresh by `/research` during the dogfood tour. Portfolio-clean,
   zero PII-import risk, nothing carried over from the old repo but the link in the README.
5. **SDD placement**: **prosthetic-conscience** — SDD + `/plan-audit` stay in the core plugin
   (interactive adversarial cowork; FEOV research reports cross-link into SDD plans).
6. **Self-improve cadence**: **daily** default once scheduled (over the old hourly); manual run
   always available, scheduling always human-opt-in.

## Open decisions

None — all resolved. **Phase 0 is complete** (repo bootstrapped; marketplace + all three plugins
install verified). **Phase 1** (harness spike) is in flight as PR #5, pending merge + `/reload-plugins`
to verify live — and to settle the one open harness question (do plugin hooks fire in subagents?).
