# FEOV: Cross-Model Seats (Per-Seat Model & Harness) — Implementation Plan

> **Issue:** [#1827](https://github.com/ctoforaday/special-circumstances/issues/1827) — *FEOV: cross-model seats (per-seat model and harness), measured before any default*
> **Related Audit Finding:** `~/.claude/scratch/universe-research-audit/FINDINGS.md` Row 13 / Fork N (*Panickssery et al. 2024* — correlated error when author and reviewer share one model tier)
> **Status:** In Progress — Phase 1 landed in `d8e334251`; updated 2026-11-14 for `agy 1.5.9` native `run_workflow` `agent()` orchestration, per-subagent `PreInvocation` delivery & termination (`#1830`), and `\r` spinner stripping in `parseAgyModels` ([#1867](https://github.com/ctoforaday/special-circumstances/issues/1867), [#1871](https://github.com/ctoforaday/special-circumstances/issues/1871))

---

## I. Summary & Goals

### What we are building
1. **Per-seat `{ harness, model }` configuration** replacing the two-tier `model` / `judgmentModel` configuration across `feov-record setup`, `inputs/run-config.json`, `debate.js`, `dashboard`, and all readers/auditors (`modeltier`, `record`, `cost`, `capture`, `cli/tiers.go`).
2. **Dynamic harness & model discovery** at `setup` preflight (`agy models` / `agy agents` for Antigravity seats; Claude's accepted model surface for Claude Code seats), so FEOV follows what each harness can dispatch rather than hardcoding a static list.
3. **Two native workflows — each system native from the POV of the agents:**
   - **Claude Code native workflow (`debate.js`):** Keeps Claude's `Workflow` tool (`agent()`, `SubagentStart`/`SubagentStop`, `StructuredOutput`, `resumeFromRunId`) for Claude runs, updated to read per-seat models from `seats`.
   - **Antigravity native workflow (`debate.py`):** A Python 3 workflow executed via Antigravity's native **`run_workflow` tool** (`agy 1.5.9+`, bundled Python — [#1871](https://github.com/ctoforaday/special-circumstances/issues/1871)), dispatching every seat via `run_workflow`'s native **`agent(prompt=..., type_name=..., role=..., model=..., schema=...)`** primitive (and `parallel(...)`, never top-level `agy -p` / `claude -p` subprocesses), and relying on Antigravity's **native subagent hooks** (`feov-subagentstart`, `feov-sessionstart -event PreInvocation`, `feov-pretooluse`, `feov-posttooluse`, `feov-subagentstop`) for sitting brackets, per-subagent (`conversationId`) work-list / report-text delivery, per-sitting hook termination (`PreInvocation` `{"terminationBehavior": "terminate"}`), tool-call limits, and `finish` tool envelope validation.
4. **Cross-harness measurement & cost reporting:** Extend `seatturn.Parse`, `servedmodel.Read`, `cost.ScanTranscript`, `capture.ModelTierAudit`, and `cli/tiers.go` to read Antigravity `transcript.jsonl` (`PLANNER_RESPONSE`), verify served models per seat against `run-config.json` `seats[seatID]` (honestly reporting `NotMeasured` when Antigravity omits `model`), and report non-Claude token usage cleanly as `Unpriced` without false `fable` tier fallbacks.
5. **Constitution risk disclosure (`FINDINGS.md` Row 13):** State the correlated-error risk of single-model author/reviewer runs in `red-chair` and `lead-judge` constitutions (`scripts/agentgen/src/`).

### Value case
Running all seats on one model tier risks correlated misses between author (`blue-*`) and reviewer (`red-*` / `judge`) (`FINDINGS.md` Row 13), and locks specialized seats like `red-lens-dark-side` out of cyber-permissive or non-Anthropic models available on Antigravity (`agy models`). Keeping `debate.js` native to Claude `Workflow` while adding `debate.py` for Antigravity and mixed-harness runs (with `enginegen` keeping shared roster/schema constants in sync) avoids forcing Antigravity through a Claude wrapper subagent and keeps every seat native to its own harness from the agent's POV.

### Quantitative success criteria
- `unset $(env | grep -o '^FEOV_[A-Z_]*') && go -C scripts run ./check` exits 0 with **0 failures** (`enginegen -check`, `agentgen -check`, `pluginparity`, `versionguard`, `archaeology`, `golden`, and all FEOV Go tests at `-count=1`).
- `HOOKCORE_LIVE=1` Antigravity and Claude hook tests pass with **0 failures**.
- `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` and `python3 -m unittest` for `debate.py` pass with **0 failures**.
- Live `--smoke` runs via `scripts/universe.sh` on Claude-only (`debate.js`), Antigravity-only (`debate.py`), and mixed-harness (`claude` + `antigravity`) configurations complete with **100% of sittings bracketed** in `record.db` and non-failing `feov-record tiers` / `feov-record capture`.

### Non-goals & explicit boundaries
- **No backwards compatibility path:** Per `AGENTS.md` and Issue #1827, `model` and `judgmentModel` are removed from `run-config.json`, `debate.js` args, `dashboard`, and `feov-record setup` with zero aliases, dual readers, or fallbacks.
- **No default tier changes:** Default models for standard runs remain unchanged until real cross-model runs are measured.
- **Constitutions remain primary defence:** Cross-model seats supplement (never replace) the seat constitutions and adversarial protocol.
- **No `EventSchema` bump:** Per-seat configuration lives in `inputs/run-config.json`, and existing `sitting_open` / `sitting_close` / `register` / `seat_turn` tables already carry `agent_id`, `agent_type`, `seat_id`, and `transcript_path`.

---

## II. Current State & Gaps

### Technical baseline
- **Languages & runtimes:** Go `1.27` (`plugins/frank-exchange-of-views/tools/`), Node.js (`debate.js` + `tests/simulator/`), Python 3 (dev scripts in `scripts/`, target runtime for `debate.py`).
- **Primary dependencies:** `modernc.org/sqlite`, `google.golang.org/protobuf`, `github.com/spf13/cobra` / `pflag`, `shared/hookcore`.
- **Storage & data model:** Per-run `record.db` (SQLite event log + views), `inputs/run-config.json` (run terms & config), and `<project>/.claude/run-live.json` (live run marker + `session_id` ownership).
- **Constraints:** `AGENTS.md` forbids backwards-compatibility paths and hand-editing generated carriers (`enginegen`, `agentgen`, `cmdtreegen`); `FEOV_*` env vars must be unset before Go tests.

### Measured gaps

| Component | Current Behaviour | Gap for #1827 |
|---|---|---|
| `setup/run.go` (`Config`, `Run`) & `cli/setup.go` | Requires `--model` (bulk) and `--judgment-model` (judgment); writes `"model"` and `"judgmentModel"` to `inputs/run-config.json`. | Cannot express per-seat models or harnesses (`claude` vs `antigravity`); performs no live `agy models` / `agy agents` discovery. |
| `internal/modeltier/modeltier.go` | `Order = ["haiku", "sonnet", "opus", "fable"]`; `Of(m)` scans substrings and falls back to `"fable"`; `Config(runDir)` reads `"model"` / `"judgmentModel"`. | Any non-Anthropic model (e.g. `gpt-oss-120b-medium`, `gemini-4-argon-high`) silently classifies as `"fable"` (dearest Anthropic tier). |
| `internal/dashboard/model.go` & `cli/dashboard.go` | `dashboard.Config` reads `model` / `judgmentModel` from `inputs/run-config.json` and CLI flags. | Would silently render empty strings once `run-config.json` drops `model`/`judgmentModel`. |
| `skills/research-protocol/scripts/debate.js` | Takes `{ model, judgmentModel }` and builds `TIER = { bulk, judgment }`. Calls `Workflow` `agent()` directly. | Cannot vary model per seat (e.g. `red-lens-dark-side`) and cannot dispatch Antigravity seats (`Workflow` has no shell/SDK access). |
| `skills/research/SKILL.md`, `scripts/universe.sh`, `docs/local-development.md` | Instructs `/research` and smoke runs solely in terms of `--model`, `--judgment-model`, and Claude's `Workflow` tool. | No Python workflow launch path and no `--bulk`/`--judgment`/`--seat` flags. |
| `plugins/frank-exchange-of-views/hooks.json:23` & `internal/hookcmd` | Claude `.claude-plugin/hooks.json` wires `SubagentStart`/`SubagentStop` -> `feov-subagentstart`/`feov-subagentstop`. Antigravity `hooks.json` only wires `PreInvocation`, `PreToolUse`, `PostToolUse` (and line 23 invokes `feov-posttoolusefailure` instead of `feov-posttooluse`), omitting `SessionStart` (`feov-subagentstart`) and `PostInvocation` (`feov-subagentstop`) even though `shared/hookcore` already normalizes Antigravity's per-subagent `conversationId`/`agentName`/`transcriptPath` into `AgentID`/`AgentType`/`AgentTranscriptPath` and formats `injectSteps`. | Antigravity `hooks.json` needs `SessionStart` -> `feov-subagentstart`, `PreInvocation` -> `feov-sessionstart -event PreInvocation`, `PostInvocation` -> `feov-subagentstop`, and `PostToolUse` -> `feov-posttooluse`; `hookcmd` needs per-`conversationId` queueing (`#1867`), `PreInvocation`-time `report_read` recording, and `PostToolUse` -> `PreInvocation` per-sitting termination (`#1830`, `#1871`). |
| `internal/seatturn/seatturn.go` & `internal/servedmodel/servedmodel.go` | Parses only Claude JSONL (`{"message":{"model":...,"usage":...}}` and `fallback` blocks). | Returns 0 turns / `NOT MEASURED` on Antigravity `transcript.jsonl` (`type: "PLANNER_RESPONSE"` with `input_tokens`, `cache_read_tokens`, `output_tokens`). |
| `internal/cost/cost.go`, `record/servedmodel.go`, `capture/capture.go`, `cli/tiers.go` | `TierMismatch` & `recordTierFindings` compare `modeltier.Of(configured)` vs `modeltier.Of(served)` against `bulk`/`judgment`; `cost.Prices` only holds `claude-*`. | Unrecognized models trigger false `FAIL`/`WARN` tier mismatches against `fable` instead of per-seat `seats[seatID]` comparison + non-fatal `Unpriced` cost reporting. |
| `scripts/agentgen/src/red-chair.md` & `lead-judge.md` | Zero mention of correlated-error risk when author and reviewer share a model tier. | Missing Row 13 disclosure from `FINDINGS.md`. |

---

## III. Proposed Design

### 1. Per-Seat Configuration (`setup`, `run-config.json`, `modeltier`, `dashboard`)

Replace `model` and `judgmentModel` everywhere with an explicit per-seat map resolved at `setup`:

- **`inputs/run-config.json` schema:**
  ```json
  {
    "topic": "...",
    "runDir": "...",
    "seats": {
      "frontier": { "harness": "claude", "model": "claude-sonnet-5-5" },
      "blue-lane": { "harness": "claude", "model": "claude-sonnet-5-5" },
      "blue-synthesize": { "harness": "claude", "model": "claude-opus-5-5" },
      "blue-respond": { "harness": "claude", "model": "claude-sonnet-5-5" },
      "red-chair": { "harness": "claude", "model": "claude-opus-5-5" },
      "red-lens-adversary": { "harness": "claude", "model": "claude-sonnet-5-5" },
      "red-lens-architecture": { "harness": "claude", "model": "claude-sonnet-5-5" },
      "red-lens-computation": { "harness": "claude", "model": "claude-sonnet-5-5" },
      "red-lens-dark-side": { "harness": "antigravity", "model": "gpt-oss-120b-medium" },
      "red-lens-evidence": { "harness": "claude", "model": "claude-sonnet-5-5" },
      "red-lens-logic": { "harness": "claude", "model": "claude-sonnet-5-5" },
      "red-lens-voice": { "harness": "claude", "model": "claude-sonnet-5-5" },
      "judge": { "harness": "claude", "model": "claude-opus-5-5" }
    }
  }
  ```
  - Valid `harness` values: `"claude"` | `"antigravity"`.
  - `blue-lane` covers all N lane invocations; `judge` covers petition, position, docket, exit, and assemble sittings.
- **`feov-record setup` CLI flags (`internal/cli/setup.go`, `internal/flags/names.go`, `internal/setup/run.go`):**
  - `--bulk <harness:model>` — default for bulk seats (`frontier`, `blue-lane`, `blue-respond`, `red-lens-*`).
  - `--judgment <harness:model>` — default for judgment seats (`blue-synthesize`, `red-chair`, `judge`).
  - `--seat <seat-id>=<harness:model>` (repeatable `StringArrayVar`) — overrides any single seat in `seatclass.Seats`.
  - Bare `<model>` without `<harness>:` defaults to the launching harness (`antigravity` when `ANTIGRAVITY_CLI` / Antigravity env is present, else `claude`).
  - Fail-fast #111 guard: `setup` resolves every seat in `seatclass.Seats` (plus `blue-lane`) and refuses with exit 2 if any seat remains unresolved.
  - `--skip-model-check` — skips live CLI discovery during offline unit tests.
  - `setup` prints `workflow: claude` (when all seats use `claude` and launched in Claude) or `workflow: python` (when any seat uses `antigravity` or launched in Antigravity).
- **`internal/dashboard/model.go` & `internal/cli/dashboard.go`:**
  - Replace `Config.Model` / `Config.JudgmentModel` with `Config.Seats map[string]modeltier.SeatSpec` read from `inputs/run-config.json` (and summarize per-seat / distinct models in the dashboard header).

### 2. Dynamic Harness & Model Discovery

Per Issue #1827, FEOV must follow what each harness can dispatch rather than fixing a static list:

- **Antigravity (`harness == "antigravity"`):**
  - `internal/setup` invokes `agy models` (parsing the first whitespace-delimited column of model IDs after stripping both ANSI CSI escape sequences and `\r`-delimited progress spinner frames `⠋ Fetching available models...\r`, which otherwise corrupt the first line and drop `gemini-4-argon-medium` on `agy 1.5.9` — [#1867](https://github.com/ctoforaday/special-circumstances/issues/1867)) and `agy agents` (verifying every FEOV agent name exists in Antigravity's loaded agent registry — [#1867](https://github.com/ctoforaday/special-circumstances/issues/1867)).
  - If a configured Antigravity model is absent from `agy models`, `setup` refuses and prints the live list returned by `agy models` (including `gemini-4-carbon-*`, `gemini-4-hydrogen-*`, `gemini-4-argon-*`, `gpt-oss-120b-medium`, `claude-opus-5-5-high`, etc.).
- **Claude Code (`harness == "claude"`):**
  - Accepts Claude aliases (`haiku`, `sonnet`, `opus`, `fable`) and explicit Claude model IDs (`claude-*`). `setup` verifies `claude` is executable on `PATH` when `workflow: python` will spawn `claude -p`.
- **`internal/modeltier` refactoring:**
  - Replace `Config(runDir) (string, string)` with `SeatConfigs(runDir string) map[string]SeatSpec` (`type SeatSpec struct { Harness, Model string }`).
  - `Of(m)` remains strictly an Anthropic tier helper; non-Anthropic models never silently fall back to `"fable"` in substitution or cost audits.

### 3. Two Native Workflows (`debate.js` and `debate.py`) & Hook Attribution

Each seat runs **natively** in its target harness from the POV of the agent:

#### A. Claude Code Native Workflow (`plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js`)
- Used when **all seats** in `run-config.json` have `harness == "claude"` and `/research` is invoked in Claude Code.
- `args` drops `model` and `judgmentModel` and takes `seats` (validated against the roster, every entry having `harness === 'claude'`).
- `dispatchAs(id)` reads `seats[id].model` and passes `{ subagent_type, model: seats[id].model }` to `Workflow` `agent(prompt, opts)`.
- Preserves `SubagentStart`/`SubagentStop` hooks, `StructuredOutput` schema enforcement, and `resumeFromRunId` replay cache unchanged.

#### B. Antigravity Native Workflow (`plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.py`)
- Executed via Antigravity's native **`run_workflow` tool** (`ScriptPath: ".../debate.py"`, shipped in `agy 1.5.9` with bundled Python — symmetrical to Claude's `Workflow` tool running `debate.js` — [#1871](https://github.com/ctoforaday/special-circumstances/issues/1871)), driving the exact same FEOV loop (`feov-record dispatch next`, `Frontier -> Blue -> Red -> Debate -> Assemble`).
- **Native `run_workflow` `agent()` seats (never top-level `agy -p` / `claude -p` subprocesses):**
  - Dispatches each seat via `run_workflow`'s native `agent(prompt=..., type_name=..., role=..., model=..., schema=...)` primitive (and `parallel(...)` for parallel lanes and lenses).
  - Source-verified in `third_party/jetski/cortex/core/tools/runworkflow/run_workflow_agents.go` (`createWorkflowAgent`, line 98):
    - Resolves the custom agent (`resolveAgentItem`, line 352) by `type_name` (e.g. `red-lens-logic`), attaches a synthetic `finish(result=<schema>)` tool (`withWorkflowFinishTool`, line 315) when `schema=` is passed, and calls `uuid.NewRandom` to mint a fresh, distinct `conversationId` for that workflow subagent.
    - Every hook wired on `DeclarativeAgentMixin` (`declarative_agent_mixin.go:194-286`: `SessionStart`, `PreInvocation`, `PreToolUse`, `PostToolUse`, `PostInvocation`, `Stop`) calls `BuildHookArgsCommon` (`hooks.go:167`), passing `conversationId` = that workflow subagent's UUID, `agentName` = `type_name`, and `transcriptPath` = `<appDataDir>/brain/<subagent-conversationId>/.system_generated/logs/transcript_full.jsonl`.
  - **Per-sitting failure & hook termination (`#1830` alignment — never `engine-stopped` on a seat failure):**
    - `DeclarativeAgentMixin`'s `NewDirectPostToolHookDecorator` (`tool_hook_decorator.go:39`) discards `PostToolHookResult.OverwriteResult`.
    - Instead, returning `{"terminationBehavior": "terminate"}` from `PreInvocation` (`prehooks.go:43`) or `PostInvocation` (`posthooks.go:47`) returns `EXECUTOR_TERMINATION_REASON_HOOK_TERMINATE` (`0x9`) for **that subagent's executor loop only**.
    - When a workflow subagent exits without calling `finish`, `runAgentTurn` (`run_workflow_agents.go:151`) -> `agentFinishResult` -> `finishResult` (line 287) returns error `"the agent's finish call has no result"` over the JSON-lines RPC pipe to the Python workflow process (`_resolve`), raising `WorkflowError` on `await seat`.
    - In `debate.py`, wrapping `try: return await agent(...) except WorkflowError: return None` catches that single sitting's termination or failure cleanly while the workflow continues to the next sitting (matching `debate.js`'s `try { return await agent(...) } catch { return null }`).
  - **Native sitting brackets, per-`conversationId` delivery & `report_read` recording via existing hooks:**
    - **Antigravity (`plugins/frank-exchange-of-views/hooks.json`):** Wire `SessionStart` -> `feov-subagentstart -event SessionStart` (opens the sitting and injects the initial work item via `injectSteps`), `PreInvocation` -> `feov-sessionstart -event PreInvocation`, `PreToolUse` -> `feov-pretooluse`, `PostToolUse` -> `feov-posttooluse` (fixing line 23 `feov-posttoolusefailure`), and `PostInvocation` -> `feov-subagentstop -event PostInvocation` (closes the sitting and ingests `transcript.jsonl`) ([#1871](https://github.com/ctoforaday/special-circumstances/issues/1871)).
    - **Per-subagent `PreInvocation` delivery & `report_read` at injection (`#1867`, `#1830`):** In `hookcmd/post.go:82` (`nextItem`), read `req.Command` (not `req.RawToolInput["command"]`, which is `"CommandLine"` on Antigravity `run_command`), and in `hookcmd/hookcmd.go:140` (`respond`), queue the pending delivery item keyed by the subagent's `conversationId` (`req.AgentID` when non-empty). When `PreInvocation` fires immediately after `PostToolUse` (before the next LLM call) for that `conversationId` and emits `{"injectSteps": [{"ephemeralMessage": "<text>"}]}`, `feov-hook` records the `report_read` row right then when handed to `injectSteps` (never when merely queued at `PostToolUse`). If `PostToolUse` sees a stuck delivery that cannot advance, it writes the reason to FEOV's `(TOOL, DEFECT)` log keyed to the sitting (`#1830`) and writes a per-`conversationId` terminate marker so the immediately following `PreInvocation` returns `{"terminationBehavior": "terminate"}` before another LLM call is made.
    - **Envelope gate tool (`hookcmd/hookcmd.go` & `post.go`):** Both `run_workflow` `agent(..., schema=...)` and `agy -p --json-schema` return structured output via the `finish` tool (`withWorkflowFinishTool`), whereas Claude uses `StructuredOutput`. `isEnvelopeTool(req.Engine, in.ToolName)` matches `"StructuredOutput"` on `EngineClaude` and `"finish"` on `EngineAntigravity` so `gateEnvelope` (`PreToolUse`) and `admit` (`PostToolUse`) enforce work-list refusals, `CHAIR_ENVELOPE` plan replacement, and `envelope_admitted` identically on both harnesses.
- **Zero drift via `enginegen`:**
  - `plugins/frank-exchange-of-views/tools/devcmd/enginegen` generates the shared `SEATS` roster, `OCCASION`, `ENVELOPE` JSON schemas, and `PLAN` constants into **both** `debate.js` and `debate.py` (with `-check` gating both).

#### C. `/research` Skill (`plugins/frank-exchange-of-views/skills/research/SKILL.md`)
- Step 2 calls `feov-record setup` with `--bulk`, `--judgment`, and any `--seat <id>=<harness:model>` overrides (`--smoke` sets `--bulk` and `--judgment` to the current harness's cheap smoke model).
- Step 3 branches on `setup`'s `workflow:` line (`Workflow` tool with `debate.js` on Claude Code; Antigravity's native `run_workflow` tool with `ScriptPath: ".../debate.py"` and `Args: {"runDir": ..., "binDir": ...}` on Antigravity `1.5.9+` — [#1871](https://github.com/ctoforaday/special-circumstances/issues/1871)).

### 4. Transcript Ingestion, Served-Model Verification & Cost Reporting

- **`internal/seatturn/seatturn.go`:**
  - Auto-detect Antigravity `transcript.jsonl` lines (`"source": "MODEL"` and `"type": "PLANNER_RESPONSE"`) alongside Claude JSONL lines. Map `input_tokens`, `cache_read_tokens`, `output_tokens`, `created_at`, `thinking != ""`, `len(tool_calls) > 0` into `seatturn.Turn`.
- **`internal/servedmodel/servedmodel.go`:**
  - Read `model` from Claude `message.model` / `fallback` blocks as today.
  - Empirically verified on `agy 1.5.8`: neither `agy -p --output-format json` nor Antigravity's `transcript.jsonl` `PLANNER_RESPONSE` emits a `model` field. Therefore `servedmodel.Observe` honestly returns `NotMeasured` (`"antigravity transcript carries no model field"`) rather than trusting a self-reported or sidecar-copied model.
- **`internal/record/servedmodel.go`, `internal/capture/capture.go`, `internal/cost/cost.go`, `internal/cli/tiers.go`:**
  - Look up each seat's configured `{ Harness, Model }` from `modeltier.SeatConfigs(runDir)[seatID]`.
  - Compare `served` against `configured.Model` (Anthropic tier match when both are recognized Anthropic models; exact normalized ID match otherwise; `NotMeasured` seats are reported without failing `capture`).
  - In `cost.ScanTranscript` / `cost.Report`: keep `cost.Prices` for Anthropic models; for non-Anthropic / un-modelled Antigravity turns, sum tokens accurately and mark `Row.Unpriced` without failing `capture` or applying Anthropic `fable` rates.

### 5. Constitution Disclosure (`FINDINGS.md` Row 13)

- Update `scripts/agentgen/src/red-chair.md` and `scripts/agentgen/src/lead-judge.md` to state the risk of correlated error when author (`blue-*`) and reviewer (`red-*` / `judge`) run on the same model/harness, and regenerate via `go -C scripts run ./agentgen`.

### 6. Consumer Census

Command run in `.claude/worktrees/plan-cross-model-seats`:
```sh
git grep -l -E "judgmentModel|judgment-model|judgment_model|run-config\.json|modeltier"
```
Every returned file accounted for:
- **Production Go & CLI (must update):**
  - `plugins/frank-exchange-of-views/tools/internal/flags/names.go` (`Model`, `JudgmentModel` -> `Bulk`, `Judgment`, `Seat`, `SkipModelCheck`)
  - `plugins/frank-exchange-of-views/tools/internal/cli/setup.go` & `internal/setup/run.go` (`Config`, `Run`)
  - `plugins/frank-exchange-of-views/tools/internal/modeltier/modeltier.go`
  - `plugins/frank-exchange-of-views/tools/internal/dashboard/model.go` & `internal/cli/dashboard.go`
  - `plugins/frank-exchange-of-views/tools/internal/record/servedmodel.go`
  - `plugins/frank-exchange-of-views/tools/internal/cost/cost.go`
  - `plugins/frank-exchange-of-views/tools/internal/capture/capture.go`
  - `plugins/frank-exchange-of-views/tools/internal/cli/tiers.go`
  - `plugins/frank-exchange-of-views/tools/internal/servedmodel/servedmodel.go`
  - `plugins/frank-exchange-of-views/tools/internal/debatejs/debatejs.go` & `internal/seatprobe/production.go`
  - `plugins/frank-exchange-of-views/tools/devcmd/enginegen/main.go`
- **Production Go (reads `run-config.json` for other keys — verify unaffected):**
  - `plugins/frank-exchange-of-views/tools/internal/setup/billing.go`, `internal/hookprovenance/hookprovenance.go`, `internal/sittingcap/sittingcap.go`, `internal/record/params.go`, `internal/record/mintbudget.go`, `internal/record/store.go`, `internal/record/record.go`
- **Scripts, Hooks, Skills & Docs (must update):**
  - `plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js` & `debate.py` (new)
  - `plugins/frank-exchange-of-views/skills/research/SKILL.md`
  - `plugins/frank-exchange-of-views/hooks.json` (line 23 `feov-posttooluse` fix)
  - `scripts/universe.sh` (`--model $MODEL --judgment-model $JUDGMENT_MODEL` -> `--bulk`/`--judgment`/`--seat`)
  - `docs/local-development.md`
  - `scripts/agentgen/src/red-chair.md`, `lead-judge.md`
- **Tests & Goldens (must update / regenerate):**
  - `plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs`, `prompts.test.mjs`, `testdata/*.golden`
  - `plugins/frank-exchange-of-views/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden`
  - `plugins/frank-exchange-of-views/tools/internal/cli/{setup_cli_test.go,takeover_test.go,tiers_test.go,epochlimitexit_test.go,hooklimitenvelope_test.go,readinghelpers_test.go}`
  - `plugins/frank-exchange-of-views/tools/internal/capture/{capture_test.go,lanecoverage_test.go,lensturn_test.go}`
  - `plugins/frank-exchange-of-views/tools/internal/{cost/cost_test.go,dashboard/render_test.go,modeltier/modeltier_test.go,record/tiersubstitution_test.go,hookcmd/limit_test.go,hookcmd/deadreport_test.go}`
  - `plugins/frank-exchange-of-views/tools/releasegate/fuzz/{fuzz_test.go,staleretry_test.go,converseoracle_test.go,failedblue_test.go,coverage_test.go,benchlift_test.go}`
- **Historical / Archived (read-only, not modified):**
  - `ideas/backlog.md`, `plans/*.md`, `research/2026-08-23_*/inputs/run-config.json`, `tools/internal/record/migrate/*` (and `recordpb`, `recordsql` hits from `run-config.json` comments).

---

## IV. Risk & Mitigation

| Risk | Likelihood | Impact | Complexity to Mitigate | Mitigation |
|---|---|---|---|---|
| Antigravity `PostToolUse` queues items under empty `ClaudeAgentID` (`#1867`) or `OverwriteResult` is ignored by `DeclarativeAgentMixin`, breaking work-list / report-text delivery or stuck-sitting termination. | High | High | Low | Key Antigravity's pending queue by subagent `conversationId` (`req.AgentID`), record `report_read` at `PreInvocation` when emitting `injectSteps`, and on a stuck delivery write `(TOOL, DEFECT)` + a per-`conversationId` terminate marker so `PreInvocation` returns `{"terminationBehavior": "terminate"}` ([#1867](https://github.com/ctoforaday/special-circumstances/issues/1867), [#1871](https://github.com/ctoforaday/special-circumstances/issues/1871), `#1830`). |
| Antigravity `PostToolUse` invokes `feov-posttoolusefailure` (`hooks.json:23`), breaking Antigravity seat envelope/next-item hooks. | High | High | Low | Fix `hooks.json:23` to `feov-posttooluse` in Phase 1 and verify with `HOOKCORE_LIVE=1`. |
| `debate.js` and `debate.py` prompt/schema blocks drift over time. | Medium | High | Low | Extend `tools/devcmd/enginegen` to generate shared `SEATS`, `OCCASION`, `ENVELOPE`, and `PLAN` blocks into both files and enforce both with `enginegen -check`. |
| Removing `model`/`judgmentModel` breaks `dashboard`, `universe.sh`, or `releasegate/fuzz` fixtures. | Medium | Medium | Low | Consumer Census above covers every caller; `go -C scripts run ./check` and `release-fuzz` CI catch any missed fixture. |
| Antigravity `transcript.jsonl` omits `model`, causing `servedmodel` to report `NotMeasured` for Antigravity seats. | High (verified) | Low | Low | Explicitly return `NotMeasured` (never fabricate a self-reported sidecar), and treat `NotMeasured` as non-fatal in `capture` & `tiers` while still verifying `agy models` at `setup`. |

### Phased execution
1. **Phase 1 (Go foundation & hooks):** Update `modeltier`, `setup`, `cli/setup.go`, `dashboard`, `seatturn`, `servedmodel`, `cost`, `capture`, `cli/tiers.go`, `hooks.json:23`, `hookcmd` per-`conversationId` delivery & `PreInvocation` termination, `releasegate/fuzz` fixtures, and `universe.sh`/`local-development.md`.
2. **Phase 2 (Workflows & constitutions):** Extend `enginegen` for JS + Python, update `debate.js` & simulator tests, implement `debate.py` (`run_workflow` `agent()`) + Python unit tests, update `skills/research/SKILL.md`, `red-chair.md`, and `lead-judge.md`.
3. **Phase 3 (Live smoke & measurement):** Run `scripts/universe.sh` `--smoke` on Claude and Antigravity runs before any default changes.

---

## V. Verification Plan

1. **Go & Repo Gates (0 failures required):**
   ```sh
   unset $(env | grep -o '^FEOV_[A-Z_]*')
   go -C scripts run ./check
   ```
2. **Live Hook Verification (`AGENTS.md` requirement for `hooks.json` / Antigravity hook changes):**
   ```sh
   HOOKCORE_LIVE=1 go -C shared/hookcore test -count=1 ./...
   ```
3. **Real-Data Antigravity Transcript Checks:**
   - Unit test `seatturn.Parse`, `servedmodel.Read`, and `cost.ScanTranscript` against an actual Antigravity session `transcript.jsonl` (`~/.gemini/antigravity-cli/brain/<id>/.system_generated/logs/transcript.jsonl`), verifying real `PLANNER_RESPONSE` token counts and honest `NotMeasured` model reporting.
   - Verify `feov-record tiers` and `feov-record capture` on a synthetic mixed-harness run containing both a Claude JSONL transcript and an Antigravity `transcript.jsonl`.
4. **Workflow Unit Tests:**
   - `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs`
   - `python3 -m unittest` for `debate.py` (verifying `run_workflow` `agent()` / `parallel()` dispatch, `schema=` envelope handling, and `try ... except WorkflowError: return None` per-sitting recovery — with zero `feov-sitting-write` calls from Python).
5. **Live End-to-End Smoke & Measurement:**
   - `scripts/universe.sh build` + live `--smoke` mixed-harness run (`claude` + `antigravity`) verifying sitting brackets in `record.db`, `feov-record tiers`, `cost.md`, and `feov-record capture`.
