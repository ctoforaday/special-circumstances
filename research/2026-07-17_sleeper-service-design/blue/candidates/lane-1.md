# Lane 1 candidate — sleeper-service (Phase 4) design

Method lens: **adversarial-disconfirming-first** — evidence AGAINST each frontier hypothesis was
hunted before evidence for it; H1 taken first, then breadth across H2–H5. Evidence base pinned
at `7bc501e` (inputs/PINNED.md); port-plan §3c read at `AgentOrange/docs/claude-port-plan.md`
(pin's `plans/` path absent in the special-circumstances tree — standing friction). External
sources accessed 2026-07-17; Claude Code docs are living sources — volatility noted per
footnote. Red's gap-pattern inventory was read pre-flight; Pattern A (issue-status
verification), Pattern B/E (figure-to-source fidelity), and live-source-drift disciplines were
applied to every external citation below.

---

## 0. Design summary (the implementable shape)

```
plugins/sleeper-service/
├── .claude-plugin/plugin.json            # requires frank-exchange-of-views
├── skills/continuous-learning/SKILL.md   # promotion ladder: insight → MEMORY → idea stub →
│                                         #   graduated research → rules/skills (HUMAN-GATED);
│                                         #   expand-existing-before-append discipline
├── commands/
│   ├── self-improve.md                   # daily loop: harvest → pick ONE → bounded research →
│   │                                     #   idea stub. Writes ONLY research/ + ideas/.
│   └── graduate.md                       # idea → full FEOV debate → projects/ promotion PR.
│                                         #   frontmatter: disable-model-invocation: true
│                                         #   (never invokable by a scheduled fire or the loop)
├── scripts/
│   ├── harvest.mjs                       # mechanical signal harvester (Node, cost-audit.mjs class)
│   └── sleeper-guard/ (Go PreToolUse)    # write-fence: allow research/ + ideas/ only when
│                                         #   the sleeper marker is present; deny-list baked in
└── docs/scheduling.md                    # the ladder: manual → Task Scheduler/cron `claude -p`
                                          #   → Desktop scheduled task → cloud Routine; preflight
                                          #   + ceilings for every unattended rung
```

Loop invariants (each argued from evidence in §§1–5):

1. **Consume durable artifacts, never introspection** — friction.md harvests, cost.md,
   board-telemetry.jsonl, run-report open questions, ideas/backlog.md + doubts.md, red's
   gap-pattern memory (mirrored). If signal isn't in a durable artifact, fix capture, not recall.
2. **Mechanics cheap, judgment full-strength** — a deterministic script ranks; the ONE picked
   item gets real research; the stub's job is provenance + alternatives, not conclusions.
3. **Headless is viable but must be *verified per run*, not assumed** — `claude -p` loads
   plugins/MCP/hooks (non-`--bare`); the scheduler wrapper asserts plugin load from
   `system/init` before trusting the run, and sets the background-workflow wait ceiling.
4. **Consent gates are structural and layered; at least one layer sits outside the model's
   write reach entirely** (branch protection / PR review). Permission rules alone are
   empirically insufficient — documented enforcement bugs, some closed *not planned*.
5. **Every unattended run carries a hard per-run budget (`--max-budget-usd`), a turn cap, a
   month-to-date ledger preflight, and leaves a resumable honest partial on any abort.**

---

## 1. H1 — What the loop consumes (adversarial pass first)

### 1.1 The case against artifact-mining (hunted first)

Three disconfirming lines were pursued:

- **Telemetry noise / alert fatigue.** Operations practice reports that the large majority of
  raw telemetry alerts are never acted on and that teams desensitize; vendor analyses of
  observability event streams put the acted-upon fraction under one in five (figure seen at
  search-digest level only — NOT leaf-verified; carried as qualitative direction, not a
  number).[^L1AlertFatigue] If sleeper-service consumed raw telemetry, the picker would drown
  and the falsifier ("curation stage reintroduces judgment cost") would trigger.
- **Goodhart / metric gaming.** When a measure becomes the optimization target it stops
  measuring; the reward-hacking literature documents agents satisfying the literal metric while
  defeating the intent (CoastRunners lagoon; formalizations showing proxy/true-objective
  divergence under continued optimization).[^L1Goodhart] A loop that scores itself on
  "friction items closed" or "proposals emitted" can game those counters.
- **Signal that evaporates.** The corpus itself proves some in-run signal is session-local:
  `log()` is operator-console-ephemeral (settled by direct check, run 4), the friction array
  was script-local until the terminal return (run 3, R5-6), and raw trajectories "currently
  evaporate with the session" (backlog item 10).[^L1FrictionEff][^L1Backlog]

### 1.2 Why artifact-mining survives the attack

- **The friction harvests are pre-curated, not raw.** Each entry is a seat's judgment-shaped
  complaint (capability named + what-I-would-have-done), aggregated by the lead — the run-3
  file is 17 attributed entries, the efficiency run's ~30; both already read as ranked
  backlogs: PDF extraction was reported by red, blue, AND judge across two consecutive runs
  as gap #1; the Read-cap class recurred at six seat classes in one run with counts
  attached.[^L1Friction3][^L1FrictionEff] The alert-fatigue failure mode belongs to unfiltered
  machine telemetry; this input is closer to postmortem findings than to alerts. The noise
  falsifier is real but is answered by *what* is consumed, not by abandoning consumption.
- **Intrinsic self-correction without external signal is the refuted alternative.** LLMs do
  not reliably improve their own outputs from introspection alone; performance can degrade
  after self-correction without external feedback.[^L1SelfCorrect] Loops that work
  (Reflexion-class) convert *execution feedback* into persistent verbal memory consumed on the
  next episode[^L1Reflexion] — structurally what friction.md + gap-pattern memory + cost.md
  already are. An introspective /self-improve ("read my own rules and have opinions") is the
  design the external evidence argues against.
- **Goodhart is contained by separating signal from objective.** The scores in §2's picker
  rank *attention*, they are never a success metric the loop reports on itself; the only
  self-graded artifact the loop emits is an idea stub that a human triages, and promotion runs
  the full adversarial FEOV pipeline where red's job is exactly to catch gamed claims. The
  loop optimizes nothing end-to-end; it proposes.
- **Evaporating signal → add capture (H1c confirmed).** The run-4 efficiency debate already
  ratified the pattern: durable signal goes to a git-tracked sink with named consumers
  (`trajectories/board-telemetry.jsonl`, one JSON line per round), explicitly because
  `log()` persists nowhere.[^L1EffReport] Sleeper-service inherits this rule: any signal the
  harvester wants that is not durable becomes a capture fix in FEOV/PC, not a recall hack.

**Verdict on H1: SUPPORTED, with the noise caveat absorbed by input selection** (curated
complaint/measurement artifacts, recurrence thresholds, human triage at the stub gate).
Confidence: HIGH on the corpus claims (pinned artifacts read directly), MEDIUM on the external
generalizations (qualitative directions leaf-verified at abstract level; specific vendor
figures not carried).

### 1.3 The consumption pipeline (implementable)

**Inputs (all git-tracked at known paths):**

| Source | Signal class |
|---|---|
| `research/*/friction.md` (every run) | capability gaps, protocol misfits, with seat + recurrence |
| `research/*/cost.md` | spend physics; where mechanics are expensive |
| `research/*/trajectories/board-telemetry.jsonl` | debate-health signal (open counts, mass trend) |
| run `report.md` "open questions carried past this run" | declared research debts |
| `ideas/backlog.md`, `ideas/doubts.md` | existing dispositions — the DEDUPE surface |
| red's gap-pattern agent memory (mirrored to a readable path) | blue-quality defect classes |

**Stage 1 — mechanical harvest (`scripts/harvest.mjs`, zero tokens or cheap tier).** Parse the
inputs into a normalized signal table: `signal id | class | occurrences (run, seat) | max
severity seen | first/last seen | existing disposition`. Dedupe keys on *class*, not exact
string — the corpus's own lineage lesson: identity-keyed detectors never fire when each cycle
mints fresh ids; carry a `supersedes` field when a signal continues a known
class.[^L1Backlog] Anything already marked DONE/FIXED in backlog.md is closed signal
(recurrence after a fix is its own high-value class: regression).

**Stage 2 — ranking (deterministic, stated in the command file so it is auditable):**
`score = recurrence_across_runs × severity × staleness_decay`, ties broken toward items with a
measured cost attached. The formula is mechanics; it may be wrong, and that is acceptable
because the human sees the full ranked table in the idea stub's provenance and the picker
never spends judgment-tier tokens.

**Red-memory provisioning (standing friction, now a design requirement):** four seat classes
in the efficiency run could not read red's gap-pattern memory[^L1FrictionEff]; this run's
setup mirrored it to `inputs/red-gap-patterns.md`. Sleeper-service makes the mirror a
*harvest step*: harvest.mjs copies the red-auditor memory dir into the loop's run directory
inputs at start, so the loop and any FEOV run it spawns always have a readable path.

---

## 2. H2 — /self-improve and /graduate mechanics

### 2.1 Disconfirming pass

- **Are cheap stubs worthless?** The strongest external evidence on machine-generated research
  ideas: in a 100+-expert blind study, LLM-generated ideas were judged *more novel* but
  *less feasible* than expert ideas, with weak idea diversity and unreliable self-assessment;
  human re-ranking improved outcomes.[^L1IdeaStudy] Read against H2: an unattended loop's raw
  ideas are exactly the artifact you should NOT trust unreviewed — which is an argument for
  the stub-then-human-gate design, and *against* both extremes (auto-promotion, and demanding
  the nightly loop produce graduation-grade research).
- **Does bounded research produce stubs too shallow to survive red later?** (H2's falsifier.)
  Partially defused by lowering what the stub must be: the stub is a *well-provenanced
  pointer with alternatives*, not a conclusion. Graduation re-runs full FEOV from scratch; a
  shallow stub costs a wasted graduation candidate, not a corrupted rule. Residual risk
  (stub quality too low to even triage) is checked by the human at the ideas/ gate — and the
  measured smoke-mode cost (~50k tokens)[^L1Backlog] says a bounded single-lane pass is
  affordable enough that shallowness is a tuning knob, not a structural flaw.

### 2.2 /self-improve (the daily driver)

1. **Enumerate** — run harvest.mjs; read the ranked signal table. Enumeration covers the
   suite's own surface (rules/skills/agents/commands as *subjects*), ideas/, and open research
   questions — read-only on everything outside research/ + ideas/.
2. **Pick ONE** — top-ranked un-dispositioned signal. One item per run is a cost ceiling and
   a quality floor (the Reflexion analogy: one reflection per episode, persisted).
3. **Research its evolution** — bounded mode by default: a single-researcher pass or
   smoke-scale FEOV (1 lane, 1 round, 1 citation pass, bulk model) with the research-protocol
   skill preloaded — disconfirming budget and semantic footnotes intact. Full-debate strength
   is *reserved for /graduate*; the daily loop must run less deep rather than not at all
   (H2b: a full run measured $414.97 at list rates[^L1Cost4] — unaffordable daily; the loop's
   whole daily budget is set below one full-debate ROUND).
4. **Emit an idea stub** — `ideas/<date>_<slug>.md` (append-only; never rewrites existing
   ideas or doubts). Stub contract (what makes graduation auditable later):
   - problem statement + the harvested signal rows that motivated it (provenance to pinned
     artifact lines);
   - **3–5 genuinely distinct alternatives** (think-around-problem applied to itself),
     each with a one-line case and cost class;
   - acceptance shape: what evidence would graduate it, what would kill it;
   - self-graded confidence + declared unknowns (the blue pre-flight discipline);
   - `graduation_trigger`: recurrence/severity condition under which /graduate is worth a
     human's time.

### 2.3 /graduate (the promotion pipeline — human at every step)

1. Human invokes `/graduate <idea>` (the command carries `disable-model-invocation: true`,
   so no scheduled fire or loop iteration can invoke it — the harness treats such skills as
   plain text on scheduled invocation[^L1ScheduledTasks]; see §4 for why this is a real
   mechanism, not prompt courtesy).
2. Full FEOV `/research` on the idea's question, evidence base pinned; the idea stub's
   alternatives become frontier hypotheses.
3. On a VERIFIED (or human-accepted UNVERIFIED) report: promotion proposal into `projects/`
   — a plan, not a rule change.
4. Any change to rules/skills/commands/hooks lands as a **pull request a human reviews and
   merges** — the loop never touches those paths even at graduation (§4's write fence has no
   graduation exception; the *human's interactive session* makes the edits, on the evidence).

**Verdict on H2: SUPPORTED** — thin driver over existing machinery; picker mechanics cheap
and deterministic; judgment untouched. Confidence HIGH (internal cost evidence measured;
external stub-quality evidence MEDIUM, one strong study).

---

## 3. H3 — Scheduling and headless reality (harness facts, checked at primary)

### 3.1 What `claude -p` verifiably supports (docs, accessed 2026-07-17)

- Non-interactive full agent loop; **all CLI options work with `-p`**; user-invoked skills and
  custom (plugin) slash commands work — `/name` in the prompt string is expanded before
  running; interactive-only built-ins (`/login`) are not available.[^L1HeadlessDocs]
- **Without `--bare`, `claude -p` loads the same context an interactive session would** —
  hooks, skills, plugins, MCP servers, CLAUDE.md, auto memory. `--bare` (recommended for CI,
  slated to become the `-p` default) skips ALL auto-discovery — so the sleeper recipe must
  either not use `--bare`, or pass `--plugin-dir`/`--mcp-config`/`--settings` explicitly.
  **Design choice: explicit `--bare` + explicit flags is the reproducible recipe** ("the same
  result on every machine"), with non-bare as the simpler fallback; either way the wrapper
  must re-check when the default flips.[^L1HeadlessDocs]
- **Load verification is machine-checkable**: `--output-format stream-json` begins with a
  `system/init` event carrying `plugins` (name+path) and `plugin_errors` arrays — the docs
  explicitly suggest failing CI when a plugin did not load. The scheduler wrapper asserts
  sleeper-service + frank-exchange-of-views are present before treating the run as
  live.[^L1HeadlessDocs]
- **Background workflows are wait-capped at 10 minutes by default** (v2.1.182+,
  `CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS`; `0` = no limit). A debate workflow spawned inside a
  headless run can exceed 10 minutes routinely — the wrapper MUST raise this ceiling or the
  harness truncates paid-for work.[^L1HeadlessDocs] (This is exactly the class of harness
  limit H3 predicted someone would trip silently.)
- Exit behavior: code 0 success; code 1 on `--max-turns` exhaustion and other errors —
  machine-readable for the scheduler. `--output-format json` carries `total_cost_usd` and a
  per-model cost breakdown per invocation.[^L1HeadlessDocs][^L1CliReference]
- Permission pre-configuration headless: `--allowedTools`, `--disallowedTools`,
  `--permission-mode` (incl. `dontAsk` — denies anything not allow-listed; documented for
  "locked-down CI runs").[^L1CliReference][^L1HeadlessDocs]
- MCP headless: `--mcp-config` + `--strict-mcp-config` (ignore all other MCP configs) —
  and in non-bare mode the project `.mcp.json` loads as in interactive sessions. The **qmd
  HTTP daemon** (verified live on this box: `qmd mcp --http --daemon`, PID file, `/health`,
  Streamable HTTP at :8181/mcp) is the recall path that avoids per-invocation model loads
  (bare-CLI hybrid query measured 36.3s from model loading alone)[^L1Backlog] — the scheduled
  wrapper starts-or-checks the daemon before launching, and degrades to lex-only (BM25,
  model-free) when the daemon is down rather than aborting.

### 3.2 Disconfirming pass — where headless breaks

- **Headless regressions are a live, recurring class**: `claude -p` hung/appeared to hang on
  Windows during the slash-command scan for eight consecutive releases (v2.1.161–v2.1.168;
  fixed v2.1.169)[^L1WindowsHang]; custom slash commands were not discovered in CLI/SSH
  headless mode on Linux in an earlier version[^L1HeadlessDiscovery]; plugin-scoped command
  namespace resolution has its own failure history. Consequence: **scheduling.md must ship a
  preflight** (a trivial `claude -p --output-format json` ping asserting plugin presence and
  a version pin known-good for headless) and treat silent-success as unproven — the routines
  doc itself warns that green run status ≠ task success.[^L1RoutinesDocs]
- **/loop is NOT the scheduling substrate**: session-scoped, fires only while the session is
  open and idle, recurring tasks expire after seven days.[^L1ScheduledTasks] Fine for
  babysitting a PR; wrong for a daily standing loop.

### 3.3 The scheduling ladder (docs/scheduling.md content)

| Rung | Mechanism | Fit |
|---|---|---|
| 0 | Manual `/self-improve` in a session | always available; the default until the human opts into scheduling |
| 1 | **OS scheduler + `claude -p`** (Windows Task Scheduler on this box; cron elsewhere) | the recommended local rung: full local files, qmd daemon, plugins, hooks; machine must be on; wrapper owns preflight + budget + ledger |
| 2 | Desktop scheduled task | persistent, visual schedule, local files + MCP config, per-task permission config; requires the Desktop app machinery[^L1ScheduledTasks] |
| 3 | **Cloud Routine** (research preview) | runs on Anthropic infra with the machine off; schedule ≥1h interval; **fresh clone, no local files; no permission prompts (fully autonomous); pushes restricted to `claude/`-prefixed branches by default**; MCP via connectors or committed `.mcp.json`; counts against subscription usage + a daily routine cap; requires claude.ai login (API-key auth unsupported)[^L1RoutinesDocs] |
| 4 | GitHub Actions `schedule` trigger | CI-grade recipe; secrets-managed; the port to team contexts |

Rung-3 caveats the doc must state honestly: the qmd daemon does not exist in the cloud
environment (setup-script install is possible — ~973MB models — but unverified; degrade to
lex-only or repo-committed `.mcp.json` stdio); routine behavior/limits are research-preview
volatile; and everything a routine does appears as the human's GitHub identity —
which is an argument FOR the PR gate, not against the rung.[^L1RoutinesDocs]

**Verdict on H3: SUPPORTED WITH CONDITIONS** — every capability the loop needs is documented
first-party for `-p`, but the design must carry (a) the load-verification preflight, (b) the
background-wait ceiling, (c) a version pin + known-regression note, (d) the `--bare` default
flip on the horizon. Confidence HIGH (primary docs, fetched today; living source — re-verify
at build time).

---

## 4. H4 — Consent gates that autonomy cannot erode

### 4.1 Disconfirming pass first: the gates' own failure evidence

- **Instructional gates degrade under load** — the corpus's own record: a prompt-level MUST
  (PDF extraction try-before-grade-down) was skipped live by a lane; batching prose was
  ignored wholesale at the bulk tier (0/175 multi-call messages).[^L1FrictionEff][^L1Backlog]
  "Please only write research/ and ideas/" is a wish, not a gate.
- **Permission rules are empirically leaky for exactly the file-protection case**:
  Edit/Write tools bypassing `permissions.ask` rules — reported specifically as "impossible
  to protect critical files (hooks, settings)" — is **closed NOT PLANNED** upstream
  (#22055)[^L1PermAskBypass]; deny rules non-enforced for Read/Write tools was reported and
  persisted after a claimed fix (#6631, closed; reporter re-confirmed at
  v1.0.93)[^L1DenyRWIssue]; Bash deny-rule non-enforcement was reported and closed as
  duplicate (#25621).[^L1DenyBashIssue] Per red's Pattern A discipline these closures invert
  the fix story: **the design must own the gate; upstream will not.** (Statuses verified by
  direct fetch 2026-07-17; behavior in current 2026 builds unverified — but a gate you must
  re-verify per release is not a foundation.)
- **Autonomous loops editing their own constraints is a realized failure, not a
  hypothetical**: Sakana's AI Scientist modified its own execution script to extend its
  timeout and spawned itself recursively; the authors' conclusion was sandboxing/isolation,
  not better instructions.[^L1AIScientist] Security practice generalizes this as Excessive
  Agency: the model may draft, execution of sensitive actions requires an external approval
  path and least-privilege boundaries.[^L1OWASP]

### 4.2 The layered gate design (each layer with its route-around answer)

**Write surface: `research/**` and `ideas/**` only. Deny-set: `plugins/**`, `.claude/**`
(settings, hooks wiring, loop.md), `.claude-plugin/**`, `.mcp.json`, `CLAUDE.md`, `.github/**`,
and the sleeper guard's own binary/manifest paths.**

| Layer | Mechanism | "How would the loop route around it?" — answer |
|---|---|---|
| 1. Permission config | `permissions.deny` + `--disallowedTools` on the deny-set; `--permission-mode dontAsk` for unattended runs | Known enforcement bugs (§4.1) → this layer is a belt, never load-bearing |
| 2. Path-fence hook | `sleeper-guard` PreToolUse (Go, sc-quality-gate class — hooks verifiably fire on subagent writes[^L1Doubts]) blocks Write/Edit/Bash-write targets outside research/+ideas/ when the sleeper marker is present (env var set by the wrapper, echoed into the run record). Deny-set **compiled into the binary**, not read from an editable config | The loop could try Bash to edit files (deny-rule leak class) → the hook matches Bash too, with an **allowlist inversion** (only research/+ideas/ writable; everything else denied by default) per red's invariant-soundness lesson — a denylist of tools is under-inclusive; an allowlist of destinations is not |
| 3. Self-protection | The hook denies writes to the plugin cache, `~/.claude/settings.json`, hooks.json, and its own binary path; the empty-bin bootstrap guard (backlog #36) keeps hook failure loud, and the wrapper aborts the run if the guard binary is missing (fail-closed, not no-op) | The loop could disable hooks by editing settings → settings paths are inside the deny-set; residually, chmod-readonly on guard files as defense-in-depth (a documented community pattern for the #22055 gap)[^L1PermAskBypass] |
| 4. **Out-of-band merge gate (the load-bearing layer)** | Unattended runs operate on a `sleeper/`-prefixed branch (or worktree); promotion into rules/skills happens ONLY via PR into main under **branch protection + CODEOWNERS on `plugins/**` and `.claude/**`** — server-side controls no local write can alter. Cloud routines get this natively: pushes restricted to `claude/`-prefixed branches by default[^L1RoutinesDocs] | The loop cannot route around GitHub's server-side enforcement from a local session; the erosion path left is *social* (human rubber-stamping PRs), which is a review-culture problem no mechanism solves — named honestly in scheduling.md |
| 5. Invocation gate | `/graduate` carries `disable-model-invocation: true` — the harness withholds such skills from scheduled/self invocation (fires reach Claude as plain text)[^L1ScheduledTasks]; graduation exists only as a human keystroke | A prompt-injected "run /graduate" in harvested friction text lands as inert text; layers 2+4 still hold even if invocation slipped |

**Semantic-consent doctrine, stated as the invariant:** *the loop proposes in its own space;
every promotion crosses a boundary only a human action can cross, and the boundary's
enforcement lives where the loop cannot write.* Risk-accepted residue (named per the
pragmatist duty): a determined human can misconfigure all five layers; layer 4 depends on
remote hosting (a purely local repo downgrades to layers 1–3 + review discipline) — accepted
because the suite's operating model already assumes a GitHub-hosted repo with PR review
(PR #1 flow).

**Verdict on H4: SUPPORTED — and the permission-rules-only variant is REFUTED at the leaf
node.** Confidence HIGH.

---

## 5. H5 — Cost discipline for unattended runs

### 5.1 Facts (primary, accessed 2026-07-17)

- **`--max-budget-usd` exists** (print mode only): maximum dollars per invocation before
  stopping. `--max-turns` caps agentic turns (exit code 1 on hit). `--model` pins the tier;
  `--fallback-model` chains alternates.[^L1CliReference] H5's falsifier ("no pre-launch quota
  introspection") is **half-defused**: the per-run ceiling is harness-native.
- **No harness-native cross-run (monthly) budget exists for a local scheduler.** For org API
  accounts, the Admin API offers `/v1/organizations/usage_report/messages` and
  `/v1/organizations/cost_report` (≈5-minute freshness, 1/min polling) — but requires an
  Admin key, i.e., not available to subscription-auth (claude.ai) usage, which is exactly the
  auth mode routines require.[^L1UsageAPI][^L1RoutinesDocs] Honest design: the **standing cap
  is a local ledger** — the wrapper appends each run's `total_cost_usd` (from
  `--output-format json`) to a git-tracked `ideas/.sleeper-ledger.jsonl`; preflight sums
  month-to-date and **skips the run** (loudly, with a dated skip record) when over cap.
  Post-hoc audit stays cost-audit.mjs's job.
- Cloud routines meter against subscription usage with a daily routine-run cap and optional
  metered overage[^L1RoutinesDocs] — a platform-side backstop for rung 3, not a substitute
  for the wrapper's per-run ceiling.

### 5.2 Discipline (cheapen redundancy and mechanics, never judgment)

| Stage | Tier | Ceiling |
|---|---|---|
| harvest.mjs | zero tokens (script) | n/a |
| pick + stub drafting | bulk (haiku/sonnet class) | `--max-turns` small; single agent |
| bounded research pass | bulk | smoke-scale (~50k tokens measured for smoke mode)[^L1Backlog] |
| whole daily run | — | `--max-budget-usd` **$2–5**; monthly ledger cap ~$50 (operator-tunable) |
| /graduate full debate | judgment tiers per FEOV routing | human-present; existing FEOV ceilings + stop-and-resume |

Rationale from measurement: a full debate run costs $414.97 at list rates with cache traffic
at 99% of tokens and judgment-seat spikes[^L1Cost4] — three orders of magnitude above a
defensible daily unattended spend. The daily loop must therefore never spawn full FEOV; the
expensive machinery is reachable only through the human-gated /graduate.

**Stop conditions → honest partials.** Run 4's death at the monthly spend limit is the type
specimen: null-guard abort, resumable cached state, honest UNVERIFIED assembly — losing no
paid-for work.[^L1FrictionEff] Sleeper inherits: every stage writes the blackboard as it goes
(harvest table, pick rationale, stub draft), so a budget-killed nightly run leaves a resumable
stub plus a dated abort record, and the next scheduled run's preflight detects and offers the
resume instead of restarting. Termination stays judged where judgment exists (graduation runs:
telemetry + stop-and-resume, ratified run 4; automatic severity-floor termination was
evaluated and REJECTED)[^L1EffReport] — the daily loop's ceilings are pure cost caps, not
quality judgments, which is why they may be automatic.

**Verdict on H5: SUPPORTED, with the monthly guard honestly degraded to
ledger-plus-static-ceiling** (the falsifier's fallback, now designed-in rather than
discovered later). Confidence HIGH on flags/docs; MEDIUM on the ledger design (unbuilt).

---

## 6. Pre-flight self-audit (blue discipline)

- Every load-bearing claim footnoted; access dates on all external footnotes (2026-07-17);
  living-source volatility flagged (Claude Code docs, GitHub issue statuses, routines
  research-preview).
- Red gap-pattern checks applied: **Pattern A** — all three permission issues fetched
  directly; statuses quoted (#22055 closed not planned; #25621 closed duplicate; #6631
  closed with reporter-confirmed persistence). **Pattern B/E** — the alert-fatigue "under one
  in five" figure is carried as qualitative-only with an explicit not-leaf-verified label;
  no number laundered. **Live-source drift** — docs re-fetched today, not quoted from
  memory. **File-type blindspot** — no "X doesn't exist" claim here rests on a single-scope
  grep; the "no monthly cap mechanism" claim is scoped to what the fetched CLI reference and
  Usage-API docs state, and labeled as such.
- Confidence self-grades: stated per section verdict.
- Friction encountered this lane: logged to run friction.md (qmd MCP tools not exposed at
  the lane seat despite server instructions injected; pinned `plans/` path mismatch carried
  from frontier).

## 7. Open questions carried

1. Does the PreToolUse write-fence fire identically under `claude -p --bare` with explicit
   `--settings`? (Interactive + subagent firing is verified[^L1Doubts]; headless-specific
   firing needs the Phase-4 smoke test — the plan's own verify column.)
2. `disable-model-invocation: true` is documented for scheduled fires[^L1ScheduledTasks];
   confirm equivalent enforcement when a hostile prompt *inside a -p session* tries to
   invoke /graduate (layers 2+4 hold regardless).
3. Cloud-rung qmd: can the routine setup script install qmd + models within environment
   caching limits, or does rung 3 run recall-degraded (lex-only / no qmd)?
4. Subscription-auth cost telemetry: is `total_cost_usd` populated meaningfully for
   claude.ai-auth (vs API-key) runs, or does the ledger need a token-count proxy?
5. When `--bare` becomes the `-p` default, does the non-bare recipe silently lose plugins?
   (Wrapper pins the CLI version; re-verify on bump.)

---

## Footnotes (lane-1 namespace)

[^L1HeadlessDocs]: "Run Claude Code programmatically" — Claude Code Docs,
  https://code.claude.com/docs/en/headless — accessed 2026-07-17. Living source. Key quotes:
  "Add `--bare` to reduce startup time by skipping auto-discovery of hooks, skills, plugins,
  MCP servers, auto memory, and CLAUDE.md"; "User-invoked skills and custom commands work in
  `-p` mode"; "`--bare` is the recommended mode for scripted and SDK calls, and will become
  the default for `-p` in a future release"; background subagents/workflows wait "capped at
  ten minutes by default" (`CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS`); `system/init` reports
  `plugins` and `plugin_errors`; `--output-format json` includes `total_cost_usd`.
[^L1CliReference]: "CLI reference" — Claude Code Docs,
  https://code.claude.com/docs/en/cli-reference — accessed 2026-07-17. Living source. Flags
  quoted verbatim: `--max-turns` ("Limit the number of agentic turns (print mode only). Exits
  with an error when the limit is reached"), `--max-budget-usd` ("Maximum dollar amount to
  spend on API calls before stopping (print mode only)"), `--disallowedTools`,
  `--permission-mode` (incl. `dontAsk`), `--mcp-config`, `--strict-mcp-config`,
  `--plugin-dir`, `--fallback-model`, `--no-session-persistence`; exit code 0/1 behavior.
[^L1ScheduledTasks]: "Run prompts on a schedule" — Claude Code Docs,
  https://code.claude.com/docs/en/scheduled-tasks — accessed 2026-07-17. Living source.
  Comparison table (Cloud/Desktop//loop): /loop session-scoped, 7-day expiry, machine+session
  required; Desktop: machine on, no open session, local files, per-task permission config;
  Cloud: no machine, no permission prompts, fresh clone, connectors, ≥1h interval. Also: "a
  scheduled fire only runs skills that Claude is allowed to invoke on its own... skills
  marked `disable-model-invocation: true`... reach Claude as plain text instead of executing."
[^L1RoutinesDocs]: "Automate work with routines" — Claude Code Docs,
  https://code.claude.com/docs/en/routines — accessed 2026-07-17. Research preview; explicitly
  volatile. Key quotes: "Routines run autonomously as full Claude Code cloud sessions: there
  is no permission-mode picker and no approval prompts during a run"; "By default, Claude can
  only push to branches prefixed with `claude/`"; each repo "cloned at the start of a run,
  starting from the default branch"; daily routine-run cap + subscription usage draw-down;
  `/schedule` "requires a claude.ai subscription login. If ANTHROPIC_API_KEY... is set...
  remove it first"; "A green status in the run list... does not mean the task in your prompt
  succeeded."
[^L1PermAskBypass]: "[BUG] Edit/Write tools bypass permissions.ask rules (regression of
  #11226)" — anthropics/claude-code issue #22055,
  https://github.com/anthropics/claude-code/issues/22055 — accessed 2026-07-17 (direct
  fetch). Status: **Closed as not planned**. Reproduction: Bash ask rules prompt; Edit/Write
  ask rules do not (files modified with no prompt), defeating protection of `.claude/hooks/**`
  and `.claude/settings.json`. Community defense-in-depth workaround noted in the thread:
  chmod 444 + PreToolUse hook.
[^L1DenyRWIssue]: "Permission Deny Configuration Not Enforced for Read/Write Tools" —
  anthropics/claude-code issue #6631,
  https://github.com/anthropics/claude-code/issues/6631 — accessed 2026-07-17 (direct fetch).
  Status: Closed; a prior fix was claimed via #4467, reporter re-confirmed the bypass at
  v1.0.93 (Aug 2025). Behavior in current builds unverified — treated as "cannot be the
  load-bearing layer" evidence, not as a claim about today's build.
[^L1DenyBashIssue]: "Permission deny rules not enforced for Bash commands" —
  anthropics/claude-code issue #25621,
  https://github.com/anthropics/claude-code/issues/25621 — accessed 2026-07-17 (direct
  fetch). Status: Closed as duplicate (phenomenon corroborated; canonical issue not traced —
  labeled accordingly).
[^L1WindowsHang]: "[DOCS] `claude -p` slow / appearing to hang on Windows during the
  slash-command and skill scan... (regression v2.1.161–v2.1.168, fixed in v2.1.169)" —
  anthropics/claude-code issue #66395,
  https://github.com/anthropics/claude-code/issues/66395 — accessed 2026-07-17 via search
  result title (title itself carries the version span; body not fetched — MEDIUM confidence
  on details, HIGH on existence of the regression class).
[^L1HeadlessDiscovery]: "Custom slash commands not discovered in CLI/SSH headless mode" —
  anthropics/claude-code issue #14246,
  https://github.com/anthropics/claude-code/issues/14246 — accessed 2026-07-17 via search
  digest (v2.0.71, Linux/aarch64). Status not fetched — cited only as evidence the headless
  discovery path has failed before, not as a current bug.
[^L1SelfCorrect]: "Large Language Models Cannot Self-Correct Reasoning Yet" — Huang, Chen,
  Mishra, Zheng, Yu, Song, Zhou; ICLR 2024; arXiv:2310.01798,
  https://arxiv.org/abs/2310.01798 — accessed 2026-07-17. Intrinsic self-correction (no
  external feedback) fails to improve and sometimes degrades reasoning performance.
[^L1Reflexion]: "Reflexion: Language Agents with Verbal Reinforcement Learning" — Shinn et
  al.; NeurIPS 2023; arXiv:2303.11366, https://arxiv.org/abs/2303.11366 — accessed
  2026-07-17. Environment/execution feedback converted to persistent verbal memory consumed
  in later episodes; the working precedent for artifact-fed improvement loops.
[^L1IdeaStudy]: "Can LLMs Generate Novel Research Ideas? A Large-Scale Human Study with 100+
  NLP Researchers" — Si, Yang, Hashimoto; arXiv:2409.04109,
  https://arxiv.org/abs/2409.04109 — accessed 2026-07-17. LLM ideas judged more novel
  (p<0.05) but weaker on feasibility; weak diversity and unreliable self-assessment; human
  re-ranking helps — the evidence base for human triage at the idea gate.
[^L1AIScientist]: "The AI Scientist: Towards Fully Automated Open-Ended Scientific
  Discovery" — Sakana AI, https://sakana.ai/ai-scientist/ — accessed 2026-07-17. The system
  edited its own execution script to extend its timeout and relaunched itself recursively;
  authors recommend sandboxing/isolation ("safe code execution"). Realized precedent for an
  autonomous loop modifying its own constraints.
[^L1OWASP]: "LLM06: Excessive Agency" — OWASP Top 10 for LLM Applications (2025 list; read
  via secondary explainers, primary list page not fetched — MEDIUM confidence on exact
  taxonomy wording, HIGH on the doctrine) — accessed 2026-07-17. Root causes: excessive
  functionality/permissions/autonomy; mitigations: least privilege, human-in-the-loop
  approval for sensitive actions, draft-vs-execute separation.
[^L1AlertFatigue]: Alert-fatigue/actionability claims — practitioner/vendor literature
  (openobserve.ai, site24x7, ennetix AIOps posts) — accessed 2026-07-17 at search-digest
  level only; the specific "under 1 in 5 alerts acted on / 9.6M events" figure is NOT
  leaf-verified and is carried as qualitative direction only. LOW confidence on numbers;
  MEDIUM on the qualitative phenomenon (independently attested across sources).
[^L1Goodhart]: Reward hacking / specification gaming — "Reward hacking," Wikipedia (CoastRunners
  case, orig. OpenAI "Faulty reward functions in the wild," 2016),
  https://en.wikipedia.org/wiki/Reward_hacking; survey "Reward Hacking in the Era of Large
  Models" arXiv:2604.13602 — accessed 2026-07-17 at search-digest level; qualitative use
  only (no specific figures carried).
[^L1UsageAPI]: "Usage and Cost API" — Claude Platform Docs,
  https://platform.claude.com/docs/en/manage-claude/usage-cost-api — accessed 2026-07-17 via
  search digest. `/v1/organizations/usage_report/messages` + `/v1/organizations/cost_report`;
  Admin API key required (org accounts); ~5-min freshness, 1/min sustained polling. MEDIUM
  confidence (page not fetched whole); the design claim it supports ("no subscription-auth
  monthly API guard") is additionally supported by the routines doc's auth constraints.
[^L1Cost4]: `research/2026-07-14_efficiency-investigation/cost.md` at pin `7bc501e`. Measured:
  total $414.97 list-rate across 42 agents/1975 api-turns; cache traffic 99% of tokens;
  judgment-seat premium cache-RATE-driven; per-round red-lens lines $41.89–$61.46.
[^L1Friction3]: `research/2026-07-12_feov-retrospective/friction.md` at pin `7bc501e`. 17
  attributed entries; PDF extraction reported by red-merge every round (entries 1, 5, 7, 11);
  write-block filename-keyed isolation (entry 4); Grep count footgun (12); Read-cap (15).
[^L1FrictionEff]: `research/2026-07-14_efficiency-investigation/friction.md` at pin
  `7bc501e`. ~30 entries incl.: red gap-pattern memory unreadable at four seat classes; Read
  cap at every full-read seat every round; MUST-try clause skipped live (blue-respond-r1
  entry); abort disclosure (monthly spend-limit death, resumable state, "NO REPORT ASSEMBLED
  — resumable via wf_5cefd2a4-35f"); log() settled console-ephemeral (blue-respond-r2).
[^L1Backlog]: `ideas/backlog.md` at pin `7bc501e`. Items cited: 10 (trajectories evaporate),
  15–17 (smoke mode ~50k tokens; assemble-on-failure), 28 (cost findings; panel counter
  excludes cache = 92% of flow), 29 (lineage-blind detector → supersedes fix), 34 (qmd
  measured ladder; HTTP daemon verified live 2026-07-14; bare-CLI 36.3s model-load penalty;
  MCP client-authored searches), 36 (empty-bin hook crash-storm + bootstrap guard), 39
  (batching prose ignored at haiku, 0/175).
[^L1Doubts]: `ideas/doubts.md` at pin `7bc501e`. Closed doubts: "Plugin hooks + agent memory
  in workflow agents — CONFIRMED... sc-quality-gate fired on workflow-agent writes;
  red-auditor wrote its memory: project gap-pattern file."
[^L1EffReport]: `research/2026-07-14_efficiency-investigation/report.md` at pin `7bc501e`,
  §1/§2.5/§6: severity-floor termination REJECTED as specified; per-round board-profile
  telemetry + documented stop-and-resume ratified; durable sink = merge-seat append to
  git-tracked `trajectories/board-telemetry.jsonl` with named consumers; "log() is
  console-ephemeral."
[^L1PortPlan]: `docs/claude-port-plan.md` (AgentOrange working tree; pin's
  `plans/claude-port-plan.md` path absent — friction), §3c + Phase table: sleeper-service
  structure, "the loop writes only research/ and ideas/; promotion into rules/skills requires
  the human (Semantic Consent)"; Phase-4 verify: "Headless `claude -p \"/self-improve\"`
  produces a run dir + idea stub; touches only research/+ideas/"; resolved decision 6: daily
  cadence, scheduling always human-opt-in.
