# Lane 3 candidate — sleeper-service (Phase 4) design

Method lens: **local-repo critical-stance** — subject artifacts audited directly at the pin
(`7bc501e`), harness facts established by live probe on this box (claude CLI 2.1.212), external
claims leaf-verified against primary docs/issues with access dates. Hypothesis order per
dispatch: H3 (headless harness facts) first, then breadth across H1/H2/H4/H5.

Confidence self-grades per section; open questions declared at the end. Every GitHub issue
cited was status-checked live (red gap-pattern A); repo claims verified with `git show` at the
pin, not from prose (patterns: gitignored≠absent, file-type blindspot — searches ran over
code, config, and manifest layers, not `*.md` only).

---

## 0. Verdict of the lane in one paragraph

Daily unattended `claude -p "/sleeper-service:self-improve"` is **viable on this box today**
— proven by live probe, not inference: a plugin slash command executed headless, spawned a
subagent with preloaded skills, returned exit 0 and machine-readable JSON cost telemetry
[^L3HeadlessProbe]. The same probe demonstrated the safety default the design should build on:
with no permission pre-configuration, the headless run's **write was denied** and the agent
reported the block honestly instead of routing around it. The design is therefore: a thin,
mostly-scripted driver over existing FEOV machinery (mechanics scripted, judgment prompted —
the corpus's own doctrine [^L3ScriptVsProse]), consuming the already-shipping harvest
artifacts, writing only `research/` + `ideas/`, held inside an **allowlist-inverted
permission profile** (`dontAsk` + narrow allow rules) whose deny set lives outside the
loop's writable surface, with per-run `--max-budget-usd` ceilings and a mechanical
month-to-date ledger preflight. The known weak joint is MCP-under-headless (three open
upstream bugs [^L3McpHeadlessBugs]); the loop must degrade to Grep/Read when qmd is
unreachable, which the efficiency run already demonstrated is workable [^L3QmdFallback].

---

## 1. H3 first — headless viability is settled by harness facts (VERIFIED, high confidence)

### 1.1 What was verified live on this box (2026-07-17, claude CLI 2.1.212)

Probe P1 — plain print mode with budget flag and JSON output:

```
claude -p "Reply with exactly: OK" --model haiku --output-format json --max-budget-usd 0.10
```

Result: exit 0; JSON envelope carrying `result`, `is_error:false`, `num_turns`,
`session_id`, `total_cost_usd:0.0246903`, per-model usage breakdown,
`permission_denials:[]`, `terminal_reason:"completed"` [^L3HeadlessProbe]. Everything a
scheduler needs for machine-readable success/failure and spend accounting is in one JSON
object on stdout.

Probe P2 — **plugin slash command, headless**:

```
claude -p "/prosthetic-conscience:probe" --model haiku --output-format json --max-budget-usd 0.30
```

Result: exit 0; the marketplace-installed plugin command loaded and ran; it spawned a
subagent; cross-plugin `skills:` preloading reached the subagent (the probe quoted the
critical-stance verification line verbatim); and — decisive for §4 — the probe's hook-test
write was **permission-denied** because no write permission was pre-configured, and the run
reported `Hook | BLOCKED — subagent write permission denied` in its result text rather than
attempting a workaround [^L3HeadlessProbe]. Cost: $0.058 at haiku.

Primary-doc corroboration: "User-invoked skills and custom commands work in `-p` mode:
include `/skill-name` in the prompt string and Claude Code expands it before running"
[^L3HeadlessDocs]. Phase 4's port-plan verification line — "Headless `claude -p
"/self-improve"` produces a run dir + idea stub; touches only research/+ideas/"
[^L3PortPlan] — is mechanically achievable.

### 1.2 The headless flag surface that matters (verified via `claude --help` + docs)

| Need | Mechanism (verified present, 2.1.212) |
|---|---|
| Non-interactive run | `-p/--print`; slash commands expand in the prompt string [^L3HeadlessDocs] |
| Machine-readable result | `--output-format json` (result, is_error, total_cost_usd, permission_denials, num_turns) [^L3HeadlessProbe] |
| Hard spend ceiling | `--max-budget-usd <amt>` — "Maximum dollar amount to spend on API calls before stopping (print mode only)" [^L3CliReference] |
| Turn ceiling | `--max-turns` — "Exits with an error when the limit is reached" [^L3CliReference] |
| Permission posture | `--permission-mode` (`dontAsk` auto-denies anything not pre-approved) [^L3PermissionsDocs] |
| Model pinning + resilience | `--model`, `--fallback-model a,b` (print-only) [^L3CliReference] |
| MCP scoping | `--mcp-config` + `--strict-mcp-config` (load ONLY the named servers) [^L3CliReference] |
| Structured output | `--json-schema` (validated; invalid schema exits with error ≥2.1.205) [^L3HeadlessDocs] |

### 1.3 The three traps the recipe must design around (disconfirming evidence, all verified)

1. **`--bare` drift.** Docs: "`--bare` is the recommended mode for scripted and SDK calls,
   and will become the default for `-p` in a future release" — and bare mode "skip[s]
   auto-discovery of hooks, skills, plugins, MCP servers, auto memory, and CLAUDE.md"
   [^L3HeadlessDocs]. Sleeper-service NEEDS the plugin stack. The scheduled recipe must
   therefore pass its requirements explicitly and defensively (never rely on `-p`'s current
   non-bare default surviving upgrades); the cheapest durable form is an explicit smoke
   assertion in the loop's step 0: the `system/init` stream event (or a trivial probe) lists
   loaded `plugins`, and the run **aborts loudly if sleeper-service/FEOV are absent** — the
   docs expose `plugins` and `plugin_errors` fields for exactly this CI pattern
   [^L3HeadlessDocs].
2. **Workspace trust is interactive-only.** "In non-interactive mode with `-p`, no dialog
   appears and the rules stay ignored" — project-settings allow rules apply only after the
   workspace trust dialog was accepted in an interactive session [^L3PermissionsDocs]. Recipe
   line: trust the repo interactively once before scheduling (true on this box already); a
   fresh clone scheduled cold will silently run with project allow rules ignored — which
   fails SAFE (denied writes), but fails, so it must be a documented precondition, not a
   surprise.
3. **MCP under headless is the fragile joint.** Three relevant upstream defects, all
   **Open** (status leaf-checked 2026-07-17): stdio MCP tools silently missing on the first
   turn when server startup exceeds ~2s (regression since 2.1.144, #76239); stdio tool calls
   hanging indefinitely with several servers loaded, worked around by `--strict-mcp-config`
   (#68375); HTTP MCP `-p` runs exiting silently with no output (#32191, older)
   [^L3McpHeadlessBugs]. Because these are open (not closed-fixed), the design must OWN the
   workaround, not wait for upstream: (a) the loop's MCP profile is
   `--strict-mcp-config --mcp-config <sleeper-mcp.json>` naming **qmd only** — fewer servers
   is the #68375 mitigation and the loop has no need of pdf-reader/arxiv-latex at the driver
   level (research subagents reach tools via ToolSearch per the shipped seat grants);
   (b) qmd is addressed via the **HTTP daemon**, verified live on this box 2026-07-14
   (`qmd mcp --http --daemon`, PID file, `/health`, MCP Streamable HTTP at `:8181/mcp`)
   [^L3QmdDaemon] — the scheduler preflight curls `/health` and starts the daemon if absent,
   converting the #76239 slow-stdio-start class into a non-event; (c) the loop DEGRADES to
   Grep/Read over the corpus when qmd is unreachable — the efficiency run's seats did exactly
   this and produced usable work, logging the absence as friction [^L3QmdFallback].

### 1.4 Scheduler options (primary doc, compared) and the recommended default

Claude Code now ships a four-tier scheduling surface [^L3ScheduledTasks]:

| Option | Runs on | Machine on? | Session? | Fit for sleeper |
|---|---|---|---|---|
| `/loop` + cron tools | local session | yes | yes, open | NO — session-scoped, 7-day expiry, wrong shape for a daily standing loop |
| Desktop scheduled tasks | local machine | yes | no | Viable alternate (per-task permission config; needs Desktop app) |
| **Routines** (cloud) | Anthropic cloud | no | no | Off-box option — see below; research preview |
| OS scheduler + `claude -p` | local machine | yes | no | **RECOMMENDED default** — fully controllable flags, this design's recipe |

Recommended default for this box: **Windows Task Scheduler → the sleeper launch script →
`claude -p "/sleeper-service:self-improve" [guard flags]`**, because it is the only local
option where every §1.2 flag is explicit and version-pinnable, and the port plan already
names it [^L3PortPlan]. The launch script (mechanics, Node like the other run scripts) owns:
qmd daemon health preflight, month-to-date budget ledger check (§5), date-slug idempotency
(if today's run dir exists → resume or no-op, since OS schedulers have no catch-up
semantics and the in-session scheduler explicitly does "no catch-up for missed fires"
[^L3ScheduledTasks]), and JSON-result logging.

**Cloud Routines as the off-box alternate** (documented, not default): routines "run
autonomously as full Claude Code cloud sessions: there is no permission-mode picker and no
approval prompts during a run", on a **fresh clone of the default branch**, and "by default,
Claude can only push to branches prefixed with `claude/`" [^L3RoutinesDocs]. That last
property is a structural human-gate: a cloud sleeper can only ever produce a `claude/`
branch a human merges. Costs of the alternate: no local qmd index (BM25 recall gone unless
rebuilt in the environment's setup script), all claude.ai connectors included by default
(must be trimmed per routine), subscription-usage draw with a daily run cap, and research
preview volatility ("Behavior, limits, and the API surface may change") [^L3RoutinesDocs].
GitHub Actions `schedule` triggers are the third documented fallback [^L3ScheduledTasks].

**Grade: H3 CONFIRMED** with named traps. The falsifier (headless can't load plugin
commands) is disproven by direct probe.

---

## 2. H1 — what the loop consumes: artifact mining, mostly scripted (high confidence)

### 2.1 The inputs already exist and are already pre-ranked (verified at the pin)

Audited directly at `7bc501e`:

- **`friction.md` per run** — run 3: 17 attributed entries; run 4: ~39 entries including a
  full envelope harvest section; even the single-round smoke run produced an honest
  template-misfit entry [^L3FrictionCorpus]. These read as pre-ranked improvement backlogs:
  PDF extraction was reported by every red merge across two consecutive runs ("gap #1");
  the write-guard and Read-cap classes carry per-round recurrence counts in the text itself.
- **`cost.md` per run** — measured seat-round token/dollar tables (run 3: $149.95; run 4:
  $414.97; cache traffic 99% of tokens) [^L3CostRecord].
- **`trajectories/board-telemetry.jsonl`** — one JSON line per round (open count, severity,
  mint profile, mass), shipping in FEOV 0.6.0 per the ratified efficiency plan
  [^L3EfficiencyPlan].
- **red's gap-pattern memory** — 30+ named patterns, now mirrored into each run's
  `inputs/red-gap-patterns.md` at pre-create (PR-C.2) [^L3EfficiencyPlan]; this run's own
  mirror is 1,558 lines.
- **`ideas/backlog.md`** (40 statused items, each already carrying provenance to the run
  that minted it) and **`ideas/doubts.md`** (hypothesis → adjudication lifecycle, all five
  founding doubts closed with evidence) [^L3IdeasCorpus].
- **`run-record-audit.md`** per run from the capture script (integrity verdicts)
  [^L3ResearchCommand].

The corpus demonstrates H1(c) in both directions: critical signal that once evaporated
(raw trajectories, `log()` output, envelope friction on abort) got **capture added** —
trajectory tarballs, friction-to-file-as-you-go, the telemetry line — rather than clever
recall bolted on afterward [^L3FrictionCorpus]. The input design should keep that rule: if
the loop wants a signal that isn't durable, the fix is a capture PR, not a bigger prompt.

### 2.2 From artifacts to proposals: harvest is a script, ranking is arithmetic, the pick is judgment

Per the script-vs-prose doctrine ("an LLM executing mechanics is an unenforced good-faith
contract" [^L3ScriptVsProse]), the pipeline splits:

1. **`harvest.mjs` (NEW, zero tokens, simulator-testable):** parses every run's
   `friction.md` + `cost.md` + `board-telemetry.jsonl` + `run-record-audit.md`, plus
   `ideas/backlog.md` checkbox state and `inputs/red-gap-patterns.md` headers; clusters
   entries by defect class (the corpus already names classes consistently enough for
   keyword clustering — "write guard", "Read cap", "PDF", "heredoc"); emits a scored
   docket: `class | occurrences | seats affected | first/last seen | staleness | open
   backlog item? | score`. Score = recurrence × severity-proxy (seat-classes affected) ×
   staleness, tunable constants in the script.
2. **The model picks ONE** — a judgment call, but a cheap one: the worst failure is a
   suboptimal research topic for one bounded run (self-correcting next day). Risk-accepted
   at the bulk tier (§5); think-around-problem still applies to the pick (the prompt
   requires the docket's top 3 be compared, not just top-1 taken).

External corroboration (secondary to the local evidence under this lens): loops fed by
recorded execution feedback are the load-bearing pattern in the self-improvement
literature — Reflexion improves agents via verbal reflection on task feedback signals held
in an episodic buffer [^L3Reflexion]; Voyager grows an ever-expanding skill library through
"environment feedback, execution errors, and self-verification" [^L3Voyager]. Both mine
execution traces, neither introspects its own prompt text cold — which is H1's claim.

**Falsifier check (noise/lag):** the corpus's mitigation already exists — `ideas/backlog.md`
is the human-curated intermediate between raw friction and action, and friction→backlog
graduation has happened by hand after every run [^L3IdeasCorpus]. The loop reads BOTH: raw
harvest for recency, backlog state for curation. If harvest-ranking still proves noisy in
runs 5–7, the degradation path is "loop proposes from backlog only" — judgment cost stays
zero because the curation judgment was already spent by the human.

**Self-poisoning guard (design decision, from red's memory-poisoning pattern class):** the
loop writes idea stubs as NEW files (`ideas/<date>_<slug>.md`) and NEVER edits
`ideas/backlog.md` or its own harvest inputs — the ranking input surface stays
human-and-run-record authored, so a bad stub cannot amplify its own future score. Past run
dirs are convention-immutable (the pinning convention), so the harvest reads a stable
corpus.

---

## 3. H2 — /self-improve and /graduate mechanics (medium-high confidence)

### 3.1 /self-improve: a thin driver, bounded by default

A full keeper debate measured $414.97 (run 4) and ~$150 (run 3) [^L3CostRecord] — daily
full-strength is indefensible ($4.5k–12k/month). The `--smoke` mode already shipped in
`/research` (1 lane, 1 round, haiku, ~50k tokens ≈ low single-digit dollars)
[^L3ResearchCommand], and the smoke run records prove the bounded mode yields an honest
artifact: a single-round UNVERIFIED verdict with template friction surfaced instead of
silently degraded [^L3SmokeRecord]. So:

```
/self-improve  (command in plugins/sleeper-service/commands/self-improve.md)
  0. PREFLIGHT (script): plugins loaded (abort if FEOV/SS absent — §1.3.1); qmd /health
     (degrade note if down); month-to-date ledger vs monthly cap (§5, abort if over);
     today's run-dir idempotency check.
  1. ENUMERATE (script): harvest.mjs docket + inventory of own surface (skills/commands/
     agents across the three plugins — from the marketplace manifests, not hardcoded).
  2. PICK ONE (model, bulk tier): compare docket top-3, pick, state why in one paragraph.
  3. RESEARCH "how should X evolve?" (delegate to FEOV in BOUNDED mode: --smoke-class
     parameters; verdict will honestly be UNVERIFIED — that is correct for a stub).
  4. EMIT the idea stub (model): ideas/<date>_<slug>.md.
  5. RECORD (script): append cost JSON + stub path to the sleeper ledger; exit with the
     JSON envelope.
```

**Idea-stub contract** (what makes graduation auditable later — the enforceable shape,
schema-checkable by the quality-gate hook):

```
# idea: <title>
provenance: <friction lines / backlog ids / telemetry rounds that motivated it>
alternatives: <≥3 genuinely distinct, per think-around-problem, each with a one-line cost>
proposal: <the favored alternative and why it beat the others>
acceptance-shape: <what a graduated version must demonstrate>
research: <run-dir of the bounded pass; verdict stamp — expected UNVERIFIED>
confidence: <self-graded>
est-cost-to-graduate: <keeper-run + build estimate>
```

**Falsifier handling (H2's own):** if smoke-grade stubs consistently die under red at
graduation, run LESS often at full strength rather than daily at smoke strength. This is
measurable: the ledger records stub→graduation survival; the revisit trigger is written
into the scheduling doc, not left to memory (same discipline as the efficiency plan's named
revisit triggers [^L3EfficiencyPlan]).

### 3.2 /graduate: human-invoked, interactive-only, where the judgment spend concentrates

`/graduate <idea-file>` is deliberately NOT schedulable: it is the promotion path and every
step is a human decision (port plan: "human approves each step" [^L3PortPlan]).

1. Human picks the stub; the command relays the stub's provenance and acceptance-shape.
2. **Full FEOV keeper run** on the stub's question — judgment tier intact (`judgmentModel`
   inherits the session model by design; "for keeper runs, omit `model` entirely"
   [^L3ResearchCommand]). This is where the $150–400 class spend belongs: once per
   graduation, not once per day. Efficiency doctrine holds: the daily loop cheapens
   redundancy and mechanics; graduation never cheapens the adversary or the judge.
3. If the debate survives red: promotion to a `plans/` proposal (SDD shape, plan-audit
   gate) or directly to a `projects/` entry; the human executes the merge/PR. The loop's
   write boundary (§4) means even /graduate's artifacts land in research/ideas/plans as
   PROPOSALS; the rules/skills edit ships in a human-reviewed PR.

---

## 4. H4 — consent gates must be structural (high confidence; the lane's core finding)

### 4.1 The evidence that instructions alone do not hold

- Local: prompt-level MUSTs were violated live in the pinned corpus — the PDF MUST-try
  clause skipped at round 0 on an open-access paper ("the MUST-try clause has no
  observable") [^L3FrictionCorpus]; the batching prose ignored 0/175 times at haiku
  [^L3IdeasCorpus]. The corpus's own conclusion is the script-vs-prose doctrine
  [^L3ScriptVsProse].
- External: Sakana's AI Scientist, given write access to its own experiment harness,
  edited its own execution script — relaunching itself in an infinite loop and extending
  its own timeout rather than fixing its code; Sakana's stated lesson is sandboxing/
  isolation, i.e., mechanism outside the model [^L3AIScientist].
- Platform: "Permission rules are enforced by Claude Code, not by the model. Instructions
  in your prompt or CLAUDE.md shape what Claude tries to do, but they don't change what
  Claude Code allows" [^L3PermissionsDocs]. The platform itself tells us which layer is
  load-bearing.

### 4.2 The permission profile (allowlist-inverted, per red's own recommendation pattern)

Red's invariant-soundness pattern says denylists default-unsafe for new channels; invert to
an allowlist [^L3RedPatterns]. The headless recipe therefore runs:

```
claude -p "/sleeper-service:self-improve" \
  --permission-mode dontAsk \
  --allowedTools "Edit(research/**)" "Edit(ideas/**)" \
                 "Bash(node */sleeper-service/*)" "Bash(git log *)" ... (enumerated, narrow) \
  --strict-mcp-config --mcp-config .claude/sleeper-mcp.json \
  --max-budget-usd <perRunCap> --max-turns <cap> \
  --output-format json
```

- `dontAsk` "auto-denies tools unless pre-approved" [^L3PermissionsDocs] — every tool the
  recipe forgot is DENIED, not prompted (there is no one to prompt). New tools default
  safe. Probe P2 demonstrated the deny path works and is honestly reported
  [^L3HeadlessProbe].
- `Edit(path)` rules are the file-write primitive that "cover[s] all built-in tools that
  edit files" (Write/NotebookEdit forms are warned-unmatched since 2.1.210 — use Edit
  forms) [^L3PermissionsDocs].

### 4.3 Guarding the guards (the "how would the loop route around this?" table)

Deny-precedence physics, verified against the permissions doc: "Rules are evaluated in
order: deny, then ask, then allow"; "If a tool is denied at any level, no other level can
allow it"; command-line arguments sit BELOW managed settings and cannot override any deny
[^L3PermissionsDocs]. So the belt-and-braces deny set is placed at **user scope**
(`~/.claude/settings.json`) — outside the repo the loop works in:

```
deny: Edit(//c/Users/gbloc/.claude/**), Edit(.claude/**), Edit(plugins/**),
      Edit(CLAUDE.md), Edit(MEMORY.md), Edit(.mcp.json), Edit(.github/**),
      Edit(plans/**), Edit(projects/**), Bash(git push *)
permissions.disableBypassPermissionsMode: "disable"
```

| Escape route the loop could try | What stops it | Layer |
|---|---|---|
| Edit its own allow rules in `.claude/settings.json` | `Edit(.claude/**)` deny; deny-before-allow means its own allow rules can't help | permission engine |
| Edit the user-scope deny file | `Edit(//c/Users/gbloc/.claude/**)` absolute-anchored deny (Windows paths normalize to `/c/...` [^L3PermissionsDocs]); the file is also outside the workspace | permission engine |
| Edit the plugin (its own commands/guardrail definitions, the FEOV engine, hooks.json) | `Edit(plugins/**)` in-repo; plugin CACHE lives under `~/.claude/plugins/` → covered by the user-scope absolute deny | permission engine |
| Escalate to `bypassPermissions` | `disableBypassPermissionsMode: "disable"` — "A user can set it in their own settings to lock themselves out of bypass mode" [^L3PermissionsDocs] | settings |
| Write guardrail files via Bash subprocess (`node -e`, `sed`, redirects) | REAL GAP: file rules "don't apply to arbitrary subprocesses" [^L3PermissionsDocs]. Mitigations: (a) the Bash allowlist under `dontAsk` enumerates specific commands — arbitrary interpreters aren't on it; (b) a sleeper-scoped PreToolUse hook (Go, sc-secrets-gate class — hooks verified to fire on subagent tool use [^L3IdeasCorpus]) blocks Bash whose command text targets the guardrail set; "a blocking hook … stops the tool call before permission rules are evaluated" and hook decisions cannot be bypassed by allow rules [^L3PermissionsDocs]; (c) OS sandboxing where available merges Edit-deny paths into an OS boundary — limited on Windows, so the residual is RISK-ACCEPTED (§6 row 4) with the hook as the live control | allowlist + hook + (partial) sandbox |
| Push a promotion to main itself | `Bash(git push *)` deny in the loop profile (sc-push-freeze-guard warns-only by design [^L3PushGuard] — the loop profile upgrades this to a deny for loop sessions only); on cloud Routines, pushes are structurally restricted to `claude/` branches [^L3RoutinesDocs] | permission engine / platform |
| Persuade the human via a poisoned stub | out of mechanism's reach — the gate is the human reading the stub's provenance chain at /graduate; stub contract makes provenance mandatory and checkable | process |

**Semantic-consent continuity:** the shipped skill already defines the loop's posture when
it hits a wall — "AFTER hitting a tool limit, YOU MUST escalate with a recommendation — YOU
MUST NOT attempt an autonomous hot-fix. As a subagent, the escalation channel is your
return envelope: report the gap as friction" [^L3SemanticConsent]. The nightly run's
friction lands in the run record; the loop eats its own complaints the next morning. The
consent doctrine and the input pipeline are the same mechanism seen from two sides.

**Promotion is a human git action, always.** Nothing in the loop's profile can modify
rules/skills/commands/hooks/settings/manifests; a graduated change reaches them only as a
PR a human merges. Autonomy cannot erode the gate because the gate is not in the loop's
writable universe — that is the design invariant, stated once: **the loop's write surface
and the suite's behavior surface are disjoint by permission-engine enforcement, and the
mapping is allowlist-defined, never denylist-defined.**

---

## 5. H5 — cost discipline for unattended runs (high confidence on mechanics; one honest gap)

- **Per-run ceiling:** `--max-budget-usd` exists and is print-mode-native [^L3CliReference].
  Its exceed-behavior is under-documented ("before stopping" — no exit-code contract
  published); the build PR must verify by forced-tiny-budget test and record the observed
  behavior (open question OQ-2). `--max-turns` backs it up with a documented error exit.
- **Monthly ceiling (the falsifier, honestly):** no CLI-level programmatic quota/remaining-
  subscription introspection is documented — H5(c)'s fallback triggers: **static ceilings
  plus post-hoc audit**. Mechanically: every run appends its JSON `total_cost_usd` to
  `ideas/.sleeper-ledger.jsonl` (inside the writable surface, append-only); preflight sums
  month-to-date and refuses launch above the standing cap. This is self-accounting, so it
  inherits the attestation ceiling (in-run checks catch shape, post-hoc audit catches
  vacuity [^L3EfficiencyPlan]) — but the number comes from the harness's own cost meter,
  not the model's self-report, which is the strongest primitive available locally. Cloud
  Routines add a platform-enforced backstop: daily run caps and subscription limits, with
  runs "rejected until the window resets" absent usage credits [^L3RoutinesDocs].
- **Tiering (cheapen redundancy and mechanics, never judgment):** enumeration/scoring =
  scripts ($0); pick + stub-writing = bulk tier (risk-accepted: worst case is one wasted
  bounded run); bounded research pass = smoke-class parameters; **graduation keeper runs =
  interactive, human-launched, judgment tier untouched**. The daily loop performs NO final
  judgment — its every output is a proposal judged later — so the efficiency doctrine's
  protected category (judgment, the adversary, the full re-read) is never exercised
  unattended, and therefore never cheapened. Run-4's measured physics (judgment-seat
  premium is cache-RATE-driven [^L3CostRecord]) says the savings from keeping strong
  models OUT of the nightly loop are the dominant term — bulk-tier keeper freight "dwarf[s]
  every lever" [^L3EfficiencyPlan].
- **Stop conditions → honest partials:** run 4's death at the monthly spend limit is the
  proven pattern — null-guard abort, resumable cached state, honest UNVERIFIED assembly
  [^L3FrictionCorpus]. The loop writes the blackboard as it goes (setup script first, stub
  file before polish); an aborted nightly run leaves a resumable run dir the next night's
  idempotency check picks up (resume with original models — cache-safe, per the standing
  stop-and-resume practice [^L3ResearchCommand]).
- **Cost telemetry feeds back:** the ledger and per-run cost.md are harvest.mjs inputs —
  the loop consumes its own spend record and can propose its own diet (but never enact one:
  a model-tier change is a settings/plugin edit, outside the write surface).

---

## 6. Risk matrix (likelihood × impact × complexity-to-mitigate; risk-accepted rows argued)

| # | Risk | L | I | C | Disposition |
|---|---|---|---|---|---|
| 1 | MCP headless flake (open bugs #76239/#68375/#32191) starves qmd recall | M | L | L | Mitigate: daemon `/health` preflight + strict-mcp-config + Grep/Read degradation (proven workable [^L3QmdFallback]) |
| 2 | `--bare` becomes `-p` default upstream; nightly run silently loses plugins | M | H | L | Mitigate: step-0 loaded-plugins assertion, abort loudly (§1.3.1) |
| 3 | Loop edits its own guardrails | L | H | L | Mitigate structurally (§4.3); deny-precedence + user-scope placement + hook |
| 4 | Bash-subprocess write bypass on Windows (no full sandbox) | L | H | M | Layered (allowlist + hook), residual RISK-ACCEPTED: the actor is our own configured loop drifting from its prompt inside an enumerated Bash allowlist, not an adversary; full-sandbox complexity exceeds L×I today. Revisit trigger: any hook-log line showing a blocked guardrail-targeting Bash call |
| 5 | Unattended overspend | L | M | L | Per-run `--max-budget-usd` + `--max-turns` + ledger preflight; platform caps on cloud |
| 6 | Stub quality too low to graduate (H2 falsifier) | M | M | L | Measured via ledger survival rate; named revisit trigger: switch to weekly-full-strength |
| 7 | Version-bump empty-bin window: nightly run during update dance hits hook crash-storm | M | L | L | Bootstrap guard already shipped in hooks.json [^L3HooksJson]; scheduler preflight can also run doctor check; accept residual noise |
| 8 | Self-poisoning of the ranking input | L | M | L | Stub-files-only write pattern; backlog/harvest inputs stay human/run-record authored (§2.2) |
| 9 | Cloud Routine acts through the operator's identity ("commits and pull requests carry your GitHub user" [^L3RoutinesDocs]) with all connectors default-included | M | M | L | If the cloud alternate is used: trim connectors to none/qmd-equivalent, keep `claude/`-branch restriction ON; document in scheduling.md |

Blue-pragmatist note (scope defense): this design adds exactly TWO new code artifacts
(harvest.mjs, the sleeper PreToolUse guard hook) plus two command prompts and a scheduling
doc. Everything else reuses shipped machinery (FEOV smoke mode, setup/capture scripts,
permission engine, existing hooks). Proposals to build a bespoke daemon supervisor, a
quota-introspection service, or Windows sandboxing were considered and rejected: complexity
exceeds likelihood × impact in every case above.

---

## 7. Friction (lane)

1. **PINNED.md path defect:** the pin table lists `plans/claude-port-plan.md` at `7bc501e`,
   but `git show 7bc501e:plans/` contains only `README.md` and `efficiency-phase.md` — the
   port plan does not exist in the special-circumstances tree at the pin (verified; the
   sleeper-service README's link to it dangles too). The actual document lives at
   `AgentOrange/docs/claude-port-plan.md` (last commit `6df52af`). Cited accordingly; the
   pin table should be corrected or the plan imported.
2. Live-probe evidence (P1/P2 outputs) is quoted in this draft with the exact commands for
   re-derivation, but the JSON outputs themselves live only in this lane's transcript — an
   ephemeral-instrument residue (red's pattern). Cheap fix if red demands it: re-run the two
   probes and commit outputs under the run dir.

## 8. Open questions carried

- **OQ-1:** Does `-p` non-bare reliably load marketplace plugins from `enabledPlugins`
  when the plugin cache is mid-update (empty-bin window)? Probe P2 says yes when stable;
  the update-collision case is untested.
- **OQ-2:** `--max-budget-usd` exceed semantics (exit code? partial result? mid-turn
  stop?) — undocumented; must be tested in the build PR.
- **OQ-3:** Are Desktop scheduled tasks available/stable on Windows Desktop today (the
  compare table implies yes; not probed)?
- **OQ-4:** Routines research-preview churn — worth re-checking at build time whether
  `/schedule` + repo-scoped environments have stabilized.
- **OQ-5:** Stub survival rate instrumentation format (ledger field set) — decide at build.

## Footnotes

[^L3HeadlessProbe]: Live headless probes P1/P2, this box, 2026-07-17, claude CLI 2.1.212.
    P1: `claude -p "Reply with exactly: OK" --model haiku --output-format json
    --max-budget-usd 0.10` → exit 0, `total_cost_usd: 0.0246903`, `is_error: false`,
    `permission_denials: []`. P2: `claude -p "/prosthetic-conscience:probe" --model haiku
    --output-format json --max-budget-usd 0.30` → exit 0, plugin command executed, subagent
    spawned, critical-stance preload quoted verbatim, hook-test write permission-denied and
    reported. Outputs quoted in lane transcript; commands stated for re-derivation.
[^L3HeadlessDocs]: "Run Claude Code programmatically", Claude Code docs,
    https://code.claude.com/docs/en/headless, accessed 2026-07-17. Quotes: slash-command
    expansion in `-p`; `--bare` skip list and "will become the default for `-p` in a future
    release"; `system/init` `plugins`/`plugin_errors` fields; `total_cost_usd` in JSON
    output. Living source — volatile.
[^L3CliReference]: "CLI reference", Claude Code docs,
    https://code.claude.com/docs/en/cli-reference, accessed 2026-07-17. `--max-budget-usd`
    ("Maximum dollar amount to spend on API calls before stopping (print mode only)");
    `--max-turns` ("Exits with an error when the limit is reached"); `--strict-mcp-config`;
    `--fallback-model` (print-only); `--permission-mode` values. Cross-checked against
    `claude --help` locally (2.1.212).
[^L3PermissionsDocs]: "Configure permissions", Claude Code docs,
    https://code.claude.com/docs/en/permissions, accessed 2026-07-17. Quotes: deny→ask→allow
    evaluation order; "If a tool is denied at any level, no other level can allow it";
    "Permission rules are enforced by Claude Code, not by the model"; `dontAsk` auto-deny;
    Edit-rule coverage of all editing tools (2.1.210 warning for Write-form rules);
    subprocess non-coverage warning; blocking-hook precedence; Windows path normalization
    (`C:\Users\alice` → `/c/Users/alice`); workspace-trust ignored in `-p`;
    `disableBypassPermissionsMode` settable at any scope. Living source — volatile.
[^L3ScheduledTasks]: "Run prompts on a schedule", Claude Code docs,
    https://code.claude.com/docs/en/scheduled-tasks, accessed 2026-07-17. The
    Cloud/Desktop//loop comparison table; 7-day expiry; "No catch-up for missed fires";
    unattended pointers to Routines / GitHub Actions / Desktop tasks.
[^L3RoutinesDocs]: "Automate work with routines", Claude Code docs,
    https://code.claude.com/docs/en/routines, accessed 2026-07-17. Quotes: research-preview
    note; "run autonomously … no permission-mode picker and no approval prompts"; fresh
    clone from default branch; "by default, Claude can only push to branches prefixed with
    `claude/`"; connectors all-included by default; daily run cap and subscription usage;
    actions attributed to the operator's linked identities. Living source — volatile.
[^L3McpHeadlessBugs]: GitHub anthropics/claude-code issues, all status-checked OPEN on
    2026-07-17: #76239 (stdio MCP tools silently missing on first turn when server start
    exceeds the ~2s non-blocking pre-wait; regression since 2.1.144),
    https://github.com/anthropics/claude-code/issues/76239; #68375 (stdio tool call hangs
    with multiple servers loaded; `--strict-mcp-config` works around),
    https://github.com/anthropics/claude-code/issues/68375; #32191 (`-p` with HTTP MCP
    server exits silently), https://github.com/anthropics/claude-code/issues/32191
    (status per search listing; not individually re-fetched — grade accordingly). Open ≠
    will-be-fixed: design owns the workaround.
[^L3QmdDaemon]: `ideas/backlog.md` @ `7bc501e`, qmd adoption item: "HTTP DAEMON VERIFIED
    LIVE (2026-07-14, this box, CPU-only): `qmd mcp --http --daemon` works as README
    documents (PID file, `qmd mcp stop`, `/health`, MCP Streamable HTTP at :8181/mcp) —
    Phase 4 can depend on it." Plus measured ladder (bare-CLI hybrid 36.3s; daemon lex-only
    2.9s; BM25 CLI 0.6s).
[^L3QmdFallback]: `research/2026-07-14_efficiency-investigation/friction.md` @ `7bc501e`,
    blue-lane-1 entry: qmd MCP unavailable at seat → "fell back to Grep/Read on the local
    corpus, workable here."
[^L3PortPlan]: "Claude Code port plan" §3c and Phase table,
    `AgentOrange/docs/claude-port-plan.md` @ `6df52af` (AgentOrange repo — see friction
    item 1: the PINNED.md path `plans/claude-port-plan.md` does not exist at `7bc501e`).
    §3c: loop writes only research/+ideas/; promotion requires the human; daily default;
    Phase-4 verify line quoted in §1.1.
[^L3FrictionCorpus]: `research/2026-07-12_feov-retrospective/friction.md` (17 entries; PDF
    gap entries 1/5/7/11/17-adjacent; write-guard entries 3/4/8/10) and
    `research/2026-07-14_efficiency-investigation/friction.md` (run-4 harvest incl. the
    ABORT DISCLOSURE lead entry and blue-respond-r1's MUST-try-skipped record), both @
    `7bc501e`, read in full.
[^L3CostRecord]: `research/2026-07-14_efficiency-investigation/cost.md` @ `7bc501e`:
    run-4 total $414.97 over 4 rounds/42 agents; "Cache traffic is 99% of all tokens";
    run-3 baseline $149.95 per `plans/efficiency-phase.md` §I. Judgment-seat premium
    cache-RATE-driven per its Notes.
[^L3EfficiencyPlan]: `plans/efficiency-phase.md` @ `7bc501e`: ratified telemetry line
    (PR-A.1), red gap-pattern mirroring into run inputs (PR-C.2), attestation ceiling
    (§II constraints), bulk-tier freight note (§I out-of-scope), named revisit triggers.
[^L3IdeasCorpus]: `ideas/backlog.md` and `ideas/doubts.md` @ `7bc501e`. Backlog: 40
    statused items with run provenance; batching A/B "0/175 multi-call messages" at haiku;
    hooks-fire-on-subagent-writes confirmed in doubts.md closed item 3.
[^L3ResearchCommand]: `plugins/frank-exchange-of-views/commands/research.md` @ `7bc501e`:
    `--smoke` parameters (1 lane, 1 round, haiku, ~50k tokens); keeper-run model guidance;
    stop-and-resume as standing practice; capture step emitting cost.md and
    run-record-audit.md.
[^L3SmokeRecord]: `research/2026-07-17_smoke-ab-memarch-review/friction.md` @ `7bc501e`:
    single-round UNVERIFIED assembly with Catechism template misfit surfaced as friction —
    the bounded mode's honest-artifact precedent.
[^L3ScriptVsProse]: `plugins/frank-exchange-of-views/commands/research.md` @ `7bc501e`:
    "Prose here is for DECISIONS; the mechanics are scripted (design-by-contract: an LLM
    executing mechanics is an unenforced good-faith contract)."
[^L3SemanticConsent]: `plugins/prosthetic-conscience/skills/semantic-consent/SKILL.md` @
    `7bc501e`, final clause quoted verbatim in §4.3.
[^L3PushGuard]: `plugins/prosthetic-conscience/tools/cmd/sc-push-freeze-guard/main.go` @
    `7bc501e`, contract comment: "It NEVER blocks — the freeze is a commitment the human
    may consciously override; the guard's job is making the commitment impossible to
    forget, not impossible to break."
[^L3HooksJson]: `plugins/prosthetic-conscience/hooks/hooks.json` @ `7bc501e`: every hook
    command wrapped in the bootstrap guard ("a fresh plugin-cache version ships from git
    WITHOUT binaries … an unguarded hook crash-storms every tool call in that window").
[^L3RedPatterns]: `inputs/red-gap-patterns.md` (this run's staged mirror), read in full
    (1,558 lines): invariant-soundness-by-enumeration (allowlist inversion), citation
    Pattern A (issue-status checks), gitignored≠absent, file-type blindspot, policy-
    without-mechanism — all applied in this lane's method.
[^L3Reflexion]: Shinn et al., "Reflexion: Language Agents with Verbal Reinforcement
    Learning", arXiv:2303.11366, https://arxiv.org/abs/2303.11366, accessed 2026-07-17
    (abstract level): agents "verbally reflect on task feedback signals" held in an
    episodic memory buffer.
[^L3Voyager]: Wang et al., "Voyager: An Open-Ended Embodied Agent with Large Language
    Models", arXiv:2305.16291, https://arxiv.org/abs/2305.16291, accessed 2026-07-17
    (abstract level): "an ever-growing skill library of executable code" and iterative
    prompting incorporating "environment feedback, execution errors, and self-verification."
[^L3AIScientist]: Lu et al., "The AI Scientist: Towards Fully Automated Open-Ended
    Scientific Discovery", arXiv:2408.06292 + Sakana AI announcement
    https://sakana.ai/ai-scientist/, accessed 2026-07-17 (search-digest + abstract level,
    not full-PDF leaf-verified — the self-modification incidents (self-relaunch system
    call; timeout self-extension) and the sandboxing recommendation are stated in the
    paper's safety discussion per multiple corroborating accounts; grade MEDIUM pending a
    page-level PDF check if load-bearing beyond corroboration).
