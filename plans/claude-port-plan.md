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
frank-exchange-of-views is required by sleeper-service (which invokes it). One marketplace install
gets all three; each remains individually useful.

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
| Hooks on child tool calls | Did not fire | **PreToolUse/PostToolUse fire within subagents** — the qlty gate holds when a swarm agent writes. Parent additionally sees `SubagentStart`/`SubagentStop`. |
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
- Red → Lead: `{ verdict: PASS|FAIL, gaps: [{id, location, problem, required_fix, severity}],
  citations_checked, notes }`
- Lead-judge → transcript: `{ gap_id, resolution: closed|rebuttal_sustained|carried|unresolved,
  rationale }`

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
├── commands/{plan-audit, doctor}.md  # /plan-audit [file]; /doctor — environment preflight
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

1. **Requirements manifest** (`requirements.json`, one per plugin): each tool declares
   `{ name, purpose, tier: required|recommended|optional, min_version, check_cmd,
   install: {windows, macos, linux} }`. Single source of truth for both the doctor and the hooks.
2. **`/doctor` command** (environment preflight, à la `flutter doctor`): aggregates every plugin's
   manifest, probes each tool (presence + version), prints a per-tool ✓/✗ report with the exact
   install command for the current OS, and a summary verdict — **READY / DEGRADED / BLOCKED**.
   Read-only by default.
3. **Capability-gated hooks + SessionStart nudge**: the qlty hook checks `Get-Command qlty`
   (PowerShell) / `command -v qlty` (POSIX) before running; if absent it **no-ops with one warning
   line** pointing at `/doctor`, instead of erroring. A SessionStart hook runs the same probe once
   and surfaces a single non-blocking line if a recommended tool is missing. Fast, non-fatal.

**Installation stays human-gated (Semantic Consent):** `/doctor` may *offer* to run the install
command it printed (`/doctor --fix`, one tool at a time with confirmation), but installing mutates
the machine — explicit go-ahead required, never auto-run at session start. This box (Windows) is
the first test case: qlty ships a PowerShell installer; git/gh are already present. The manifest
carries per-OS install strings so the report is actionable wherever the suite lands.

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
3. **Technical Foundations / Analysis / Risk Matrix** (the meat, per the Gold Standard template)
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
requires the human (Semantic Consent). Requires frank-exchange-of-views.

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
| **0. Bootstrap** | New `special-circumstances` repo; marketplace + three plugin manifests; .qlty; LICENSE | Skeleton installs from git URL into a scratch project; all three plugins listed |
| **1. Harness spike** | One trivial skill+agent+command+hook in prosthetic-conscience; verify `skills:` preloading reaches a subagent; verify PostToolUse fires on subagent Write; SessionStart rule injection | Spawn test agent → confirm preloaded rule text ("not a yes-man" probe) + hook firing on its file write |
| **2. prosthetic-conscience** | Rule corpus → skills; pair-programming, SDD, project-memory, proficiency; plan-auditor + /plan-audit; **environment preflight (requirements.json + /doctor + capability-gated hooks)** | critical-stance probe gets pushback; /plan-audit returns schema verdict on a real plan; **on this box (qlty absent) edits still succeed — hook no-ops + warns — and /doctor reports it with the Windows install command**; qlty clean where present |
| **3. frank-exchange-of-views** | Agents (blue/red/judge), templates (report/debate/Heilmeier), workflow, /research | **E2E dogfood** on a fresh real topic: run dir contains living blue+red reports, candidates, debate.md with 3-party rounds; final report embeds both team reports + Heilmeier; seed a bad citation → FAIL round → diff-based re-audit → resolution recorded in transcript |
| **4. sleeper-service** | continuous-learning skill; /self-improve (daily default); /graduate; scheduling docs | Headless `claude -p "/self-improve"` produces a run dir + idea stub; touches only research/+ideas/ |
| **5. Story + publish** | **Clean start — no corpus import.** README; empty ideas/research/projects scaffolding; PII/secret scrub of skills+docs; LICENSE; publish | Fresh clone → README install steps verbatim; `git grep -iE "jarvis|8070926388|0xDEADC0DE"` clean; final `/research` smoke populates the first real run |

~6–8 working days (up ~1 day for the richer FEOV artifact/debate model). Main risks: (a) exact
plugin mechanics for SessionStart injection and cross-plugin skill preloading — Phase 1 spike
de-risks both; fallback for cross-plugin preloads is vendoring copies of needed rule-skills into
FEOV/SS; (b) Workflow-agent CLAUDE.md inheritance is undocumented — spike verifies; fallback is
explicit `skills:` preloads carrying everything the swarm needs (already the design).

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

None — all resolved. Ready to move from plan to **Phase 0 (Bootstrap)**.
