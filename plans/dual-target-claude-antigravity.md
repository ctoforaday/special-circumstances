# Special Circumstances: Dual-Target Claude Code & Google Antigravity Architecture Plan (#1473)

> **Status:** Empirically Verified Implementation Plan (v2.2.0 — [#1474](https://github.com/ctoforaday/special-circumstances/issues/1474), Umbrella [#1473](https://github.com/ctoforaday/special-circumstances/issues/1473), Plan Audit [#1467](https://github.com/ctoforaday/special-circumstances/issues/1467))  
> **Target Repository:** `/home/gblock_ctoforaday_com/projects/special-circumstances` (`origin/main` @ `61107ba4`)  
> **Engines Tested:** Anthropic Claude Code (`claude` v2.1.309) & Google Antigravity (`agy` v1.4.12)  
> **Primary Source Evidence:** [`plans/antigravity-hooks-empirical-proof.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plans/antigravity-hooks-empirical-proof.md) (§1–§7), [`plans/hide-non-user-facing-skills-slash-commands.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plans/hide-non-user-facing-skills-slash-commands.md), [`plans/fr-configurable-hook-manifests-and-reverse-domain-extensions.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plans/fr-configurable-hook-manifests-and-reverse-domain-extensions.md)

---

## I. Summary & Goals

**Objective.** Enable the four Special Circumstances plugins (`prosthetic-conscience`, `frank-exchange-of-views`, `gray-area`, `sleeper-service`) and all 15 shipped hook binaries across the three Go plugin modules to execute concurrently under both **Anthropic Claude Code (`claude 2.1.309`)** and **Google Antigravity (`agy 1.4.12`)** from a single repository branch, with zero regressions to Claude Code behavior, zero filesystem symlinks, zero external Go dependencies, and zero additional binaries in `tools/cmd/`.

**Rulings this plan implements (settled — 2026-10-18, updated 2026-11-01 for `agy 1.4.12` / `claude 2.1.309`).**

1. **Manifests:** Place `.antigravity-plugin/plugin.json` beside `.claude-plugin/plugin.json` in all four plugins, locked byte-for-byte on `name` and `version` by `scripts/versionguard` and `scripts/pluginparity` (omitting Claude-only `"dependencies"` in `frank-exchange-of-views/.antigravity-plugin/plugin.json` so `(*pluginValidator).checkManifestKeys` emits `0` warnings) ([#1468](https://github.com/ctoforaday/special-circumstances/issues/1468)).
2. **Hook Manifest Coexistence (No Symlinks, Zero Validator Warnings):** Relocate Claude Code's hook manifest from `hooks/hooks.json` to `.claude-plugin/hooks.json` (declared via `"hooks": "./.claude-plugin/hooks.json"` in `.claude-plugin/plugin.json`) while keeping `hooks/fetch-bin.sh` and `hooks/README.md` in `hooks/`, and place Antigravity's named-hook group manifest at `<plugin_root>/hooks.json`, validated by `scripts/validatejson` ([#1468](https://github.com/ctoforaday/special-circumstances/issues/1468)). This eliminates the unconditional `hooks/hooks.json: never loaded` warning in `agy 1.4.12` (`(*pluginValidator).checkHooks` at `0x98ca0c5`) and gives each harness its own unambiguous hook manifest. Future migration to Agent Plugins `1.1.0+` §8.2 reverse-domain extension directories (`com.google*/hooks/hooks.json` and `com.anthropic*/hooks/hooks.json`) is tracked in [#1456](https://github.com/ctoforaday/special-circumstances/issues/1456).
3. **One Common Wire-Adapter Library (`hookcore`):** Abstract all four hook process boundaries (`env`, `stdin`, `stdout`, `stderr` + `exitCode`) in one pure-stdlib package authored at `scripts/internal/hookcore` and propagated byte-identically into `plugins/{prosthetic-conscience,frank-exchange-of-views,gray-area}/tools/internal/hookcore` via `scripts/hookcoregen` ([#1470](https://github.com/ctoforaday/special-circumstances/issues/1470), [#1471](https://github.com/ctoforaday/special-circumstances/issues/1471), [#1472](https://github.com/ctoforaday/special-circumstances/issues/1472)).
4. **Skills & Agents Parity:** Convert `tools:` in `scripts/agentgen/src/fragments/seat-tools.md` (regenerating all 11 `frank-exchange-of-views` agents via `scripts/agentgen`) and `plugins/prosthetic-conscience/agents/{plan-auditor,probe}.md` to dual-target YAML sequences (including both classic Antigravity tool names and `agy 1.4.10+` `oneharness` tool names `read_file`, `write_file`, `edit_file`) enforced by `scripts/frontmatter`, withholding `WebFetch`/`read_url_content`/`read_browser_page` in `frank-exchange-of-views` to preserve its evidence-cache invariant; update the 6 `${CLAUDE_PLUGIN_ROOT}` skills to fall back to bare binary names on `PATH` ([#1469](https://github.com/ctoforaday/special-circumstances/issues/1469)).

**Success criteria.**

1. **Dual CLI Validation:** `claude plugin validate` (`2.1.309`) and `agy plugin validate --strict` (`1.4.12`) both exit `0` with `0` warnings across all four plugins, reporting `agents 2, hooks 1, skills 27` on `prosthetic-conscience`, `agents 11, hooks 1, skills 3` on `frank-exchange-of-views`, and `hooks 1, skills 7` on `gray-area` (§V.1).
2. **Zero Symlinks & Unchanged Binary Count:** `git ls-files -s | awk '$1 == "120000"'` returns `0` tracked symlinks; `plugins/*/tools/cmd/*` holds the exact `19` command directories (`15` hook binaries + `4` CLI/helper binaries) enforced by `scripts/pluginparity` (§V.2).
3. **Zero External Dependencies in `hookcore`:** `plugins/prosthetic-conscience/tools/go.mod` retains `0` `require` directives, and `TestHookBinaryLinksNothingExpensive` (`plugins/frank-exchange-of-views/tools/integration/surface/hookgraph_test.go`) passes with `"internal/hookcore"` linking only Go standard library packages with zero `init()` functions (§V.3).
4. **Preserved Security & Root Invariants:** `TestTheShippedBinariesHonourTheIOContract` (`invocation_test.go`), `TestNoHookGuessesAProjectRootFromItsWorkingDirectory` (`projectroot_test.go`), and `TestNoHookReadsTheEnvironmentVariableWithoutResolvingThroughThisPackage` (`hookenv_test.go`) pass under both Claude and Antigravity payloads (§V.4).
5. **Driveable Real-Client Checks:** Headless `agy -p` and `claude -p` smoke runs (§V.6) verify:
   - `sc-pretooluse` blocks a leaked secret on `run_command` / `Bash` and surfaces the denial reason to the model;
   - `sc-posttooluse` (`qualitygate`) queues lint/format feedback on Antigravity `write_to_file` / `write_file` into `<PluginDataDir>/pending_<SessionID>.jsonl` and `sc-sessionstart -event PreInvocation` injects it on the next turn;
   - `feov-pretooluse` rewrites `CommandLine` via `PreToolHookResult.overwrite` on Antigravity and `updatedInput.command` on Claude.
6. **Repository Gates:** `go -C scripts run ./check` (including `hookcoregen -check`, `hookfailgen -check`, `buildidgen -check`, `fetchbingen -check`, `agentgen -check`, `validatejson`, `frontmatter`, `versionguard`, `pluginparity`, `deadcode`) and `FEOV_RELEASE_GATE=1` fuzz pass with `0` failures (§V.8).

**Value case.** Against the naive alternatives — (a) per-binary `if isAntigravity` branches scattered across 15 hook entrypoints in three Go modules, or (b) symlinked `hooks.json -> hooks-jetski.json` manifests with separate Antigravity shim binaries in `tools/cmd/` — this architecture:
- Centralizes all wire translation (`env`, `stdin`, `stdout`, `stderr`, `exitCode`, tool/arg canonicalization, and the `PostTool` → `PreInvocation` queue) into one ~250-line stdlib-only package (`scripts/internal/hookcore`), propagated by `scripts/hookcoregen` using the exact pattern already proven by `scripts/buildidgen` and `scripts/hookfailgen`.
- Leaves every `hookunit.Unit` (`secretsgate`, `pushfreezeguard`, `qualitygate`, `toolchainnudge`, `checkpointrestore`) and `feov`'s `hookgate` untouched by synthesizing canonical snake_case keys inside `req.ToolInput`.
- Avoids Windows git symlink failures (`core.symlinks=false` checkouts turning `hooks.json` into a plain-text filename string) and `agy plugin validate --strict` warnings by pairing `.antigravity-plugin/plugin.json` + root `hooks.json` with `.claude-plugin/plugin.json` + `.claude-plugin/hooks.json`.
- Avoids adding a 16th hook binary directory in `tools/cmd/` by handling `PreInvocation` via `sc-sessionstart -event PreInvocation` (matching `gray-area-capture -event <Event>`) and folding `strikecounter` into `sc-posttooluse` when `req.ToolError != ""` on Antigravity (preserving `fetchbingen`'s one-hook-per-event invariant).

**Split (five independently green PRs, tracked in [#1468](https://github.com/ctoforaday/special-circumstances/issues/1468)–[#1472](https://github.com/ctoforaday/special-circumstances/issues/1472)).**
> **Upstream Sequencing Gate (2026-11-01):** Full migration execution and `debate.js` runtime orchestration on Antigravity are gated on upstream `/run-workflow` and Antigravity dynamic workflows launching. Only the `hookcore` library ([#1470](https://github.com/ctoforaday/special-circumstances/issues/1470)) can be built out in advance; actual plugin migration and `debate.js` execution wait for dynamic workflows to ship.

1. **PR 1 ([#1469](https://github.com/ctoforaday/special-circumstances/issues/1469) — Skills & Agents Parity):** Cheap, immediately shippable text + frontmatter gate update (`seat-tools.md`, `plan-auditor.md`, `probe.md`, `scripts/agentgen`, `scripts/frontmatter`, and the 6 `SKILL.md` binary fallbacks). Immediately fixes `agents 0` in `agy plugin validate`.
2. **PR 2 ([#1468](https://github.com/ctoforaday/special-circumstances/issues/1468) — Dual-Target Manifests, `.claude-plugin/hooks.json`, Root `hooks.json`, & `fetch-bin.sh`):** `.antigravity-plugin/plugin.json` (`versionguard` + `pluginparity`), `git mv plugins/*/hooks/hooks.json plugins/*/.claude-plugin/hooks.json` (`"hooks": "./.claude-plugin/hooks.json"`), root `hooks.json` (`validatejson`), and dual-harness `fetch-bin.sh` + `hookfailures.Ensure`.
3. **PR 3 ([#1470](https://github.com/ctoforaday/special-circumstances/issues/1470) — `scripts/internal/hookcore` & `scripts/hookcoregen`):** The common wire-adapter library, its unit tests over Claude and `.exa.hooks_pb` fixtures, and the `hookcoregen` gate in `scripts/check` and `.github/workflows/hooks.yml`.
4. **PR 4 ([#1471](https://github.com/ctoforaday/special-circumstances/issues/1471) — `prosthetic-conscience` Hook Migration):** Wire all 10 `prosthetic-conscience` hook binaries onto `hookcore`.
5. **PR 5 ([#1472](https://github.com/ctoforaday/special-circumstances/issues/1472) — `frank-exchange-of-views` & `gray-area` Hook Migration):** Wire `feov-*` (updating `hookgraph_test.go`) and `gray-area-capture` onto `hookcore`.

---

## II. Technical Context

- **Toolchain & Modules:** Go 1.25+, `GOTOOLCHAIN=local` in CI. Four Go modules: `scripts/`, `plugins/prosthetic-conscience/tools` (0 external dependencies), `plugins/frank-exchange-of-views/tools`, and `plugins/gray-area/tools`. Each plugin's `tools/` module compiles standalone so `sc-doctor --fix` can build from an installed plugin directory without the monorepo.
- **Primary Source Verification:** All binary offsets, `.gopclntab` symbols (`0x1625570`, `text_start` `0x6819000` in `agy 1.4.12`), `.exa.hooks_pb` protobuf descriptors (`0x5f93cc8` in `agy 1.4.12`), and `claude 2.1.309` decompiled JS functions referenced below are documented with reproducible extraction commands in [`plans/antigravity-hooks-empirical-proof.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plans/antigravity-hooks-empirical-proof.md) §6–§7.

### II.1 Empirical Answers to Core Architectural Questions (Q1–Q5)

#### Q1: Can `agy` be configured with an arbitrary hook filename or `.antigravity-plugin/hooks.json`?
**NO — `agy 1.4.12` strictly requires `<plugin_dir>/hooks.json` at the plugin root AND warns if `hooks/hooks.json` exists.**
- `(*pluginValidator).checkHooks` (`0x98ca020`) resolves hooks strictly from `filepath.Join(pluginDir, "hooks.json")`.
- At `0x98ca0c5`, `(*pluginValidator).checkHooks` calls `v.exists("hooks/hooks.json")` (`0x5c24e49`) and **unconditionally** emits `SeverityWarning`: `hooks/hooks.json: never loaded: plugin hooks are read only from hooks.json at the plugin root` (`0x5dc2d05`), even when root `<plugin_dir>/hooks.json` exists and loads cleanly. That warning causes `agy plugin validate --strict` to exit `1`.
- Declaring `"hooks": "./hooks-agy.json"` in `.antigravity-plugin/plugin.json` emits `warning: unknown field "hooks" is ignored` (`(*pluginValidator).checkManifestKeys` at `0x98d7300`) and registers `hooks: 0`.
- Placing `hooks.json` inside `.antigravity-plugin/hooks.json` emits `warning: .antigravity-plugin/hooks.json: never loaded: components are read from the plugin root, not .antigravity-plugin. Move it next to .antigravity-plugin` (`hooks: 0`).
- **Coexistence without symlinks or validator warnings:** Move Claude Code's `hooks/hooks.json` to `.claude-plugin/hooks.json` (referenced via `"hooks": "./.claude-plugin/hooks.json"` in `.claude-plugin/plugin.json`) and place Antigravity's named-hook manifest at `<plugin_dir>/hooks.json`.

#### Q2: Can Claude Code read `plugin.json` from the plugin root?
**NO.** Running `claude plugin validate .` on a directory containing only root `plugin.json` fails with `Plugin manifest not found at .claude-plugin/plugin.json`. Claude Code reads `.claude-plugin/plugin.json` (which supports `"hooks": "./.claude-plugin/hooks.json"`), ignoring `.antigravity-plugin/plugin.json` and `<plugin_dir>/hooks.json`.

#### Q3: How do the two engines coexist without cross-contamination?
1. **Claude Code (`2.1.309`)** reads `.claude-plugin/plugin.json` → loads `.claude-plugin/hooks.json` (via `"hooks": "./.claude-plugin/hooks.json"`) → ignores `.antigravity-plugin/plugin.json` and root `hooks.json`.
2. **Antigravity (`1.4.12`)** reads `.antigravity-plugin/plugin.json` → loads root `hooks.json` → ignores `.claude-plugin/` (so `.claude-plugin/hooks.json` never triggers `checkHooks`).
3. Both `agy plugin validate --strict .` and `claude plugin validate .` exit `0` with `0` warnings simultaneously on the same directory with zero symlinks.

#### Q4: Does Antigravity support an equivalent dot-directory (`/.antigravity-plugin/plugin.json`)?
**YES (Added in `agy 1.3.2`, verified in `agy 1.4.12` at `inspect.findManifestPath` `0x98c5360`).** Only `plugin.json` lives inside `.antigravity-plugin/`; all plugin components (`skills/`, `agents/`, `bin/`, `hooks.json`) remain at the plugin root.

#### Q5: What is the upstream status on GitHub (Anthropic, Antigravity, and Agent Plugins Spec)?
- **Agent Plugins Specification (`agentplugins/agent-plugins-spec` v1.0.0 & v1.1.0 draft):**
  - Hooks are excluded from the v1.0.0 and v1.1.0 portable core (`spec/1.0.0.md` §7). Instead, §8.1 and §8.2 define reverse-domain client extensions (`extensions["com.example.client"]` in `plugin.json` and `<plugin_root>/com.example.client/hooks/hooks.json`).
  - `agy 1.4.8` removed the experimental `enable-agent-plugins-spec` flag and enabled Agent Plugins 1.0 extension directories (`com.google/` and `com.google.antigravity/`) by default (`inspect.findExtensionDirectory` at `0x98c9ba0` in `1.4.12`) for `skills/`, `agents/`, and `.mcp.json`, but `(*pluginValidator).checkHooks` (`0x98ca020`) still reads `<plugin_root>/hooks.json`. `claude 2.1.309` has not adopted Agent Plugins 1.0.0.
  - Future migration to §8.2 extension directories once upstream CLIs load hooks from `<namespace>/hooks/hooks.json` is tracked in [#1456](https://github.com/ctoforaday/special-circumstances/issues/1456).

---

## III. Proposed Changes

### III.1 Directory Layout for Dual Targeting

```text
plugins/prosthetic-conscience/
├── .claude-plugin/
│   ├── plugin.json            # [MODIFY] Claude manifest (adds "hooks": "./.claude-plugin/hooks.json")
│   └── hooks.json             # [MOVE]   Moved from hooks/hooks.json (read by claude, fetchbingen, invocation_test)
├── .antigravity-plugin/
│   └── plugin.json            # [NEW]    Antigravity manifest (locked to .claude-plugin/plugin.json via versionguard)
├── hooks/
│   ├── fetch-bin.sh           # [MODIFY] Generated by scripts/fetchbingen (dual-harness ROOT + engine-aware speak())
│   └── README.md              # [MODIFY] Documents .claude-plugin/hooks.json + root hooks.json coexistence
├── hooks.json                 # [NEW]    Antigravity lifecycle hooks (named-hook group schema at plugin root)
├── agents/
│   ├── plan-auditor.md        # [MODIFY] Dual-target YAML sequence tools: [...]
│   └── probe.md               # [MODIFY] Dual-target YAML sequence tools: [...]
├── skills/
│   └── doctor/SKILL.md        # [MODIFY] Dual-target binary invocation fallback
└── tools/
    └── internal/
        └── hookcore/          # [NEW]    Generated byte-identical copy from scripts/internal/hookcore
```

(The same `.claude-plugin/hooks.json`, `.antigravity-plugin/plugin.json`, root `hooks.json`, and `tools/internal/hookcore` layout applies to `frank-exchange-of-views` and `gray-area`; `sleeper-service` receives `.antigravity-plugin/plugin.json` only, as it ships no hooks or binaries.)

---

### III.2 Dual-Target Manifests, `.claude-plugin/hooks.json`, Root `hooks.json`, & Self-Healing Bootstrap ([#1468](https://github.com/ctoforaday/special-circumstances/issues/1468))

- **`[NEW]` `plugins/{prosthetic-conscience,frank-exchange-of-views,gray-area,sleeper-service}/.antigravity-plugin/plugin.json`:**
  Contains `name`, `version`, `description`, `author`, `repository`, and `license` matching `.claude-plugin/plugin.json`. For `frank-exchange-of-views`, omit the Claude-only `"dependencies"` key (which `(*pluginValidator).checkManifestKeys` at `0x98d7300` in `agy 1.4.12` flags with `warning: unknown field "dependencies" is ignored`).
- **`[MOVE]` `plugins/{prosthetic-conscience,frank-exchange-of-views,gray-area}/hooks/hooks.json` → `plugins/{...}/.claude-plugin/hooks.json` & `[MODIFY]` `.claude-plugin/plugin.json`:**
  Relocate Claude's hook manifest into `.claude-plugin/hooks.json` and declare `"hooks": "./.claude-plugin/hooks.json"` in `.claude-plugin/plugin.json` so `(*pluginValidator).checkHooks` (`0x98ca0c5`) never sees `hooks/hooks.json`. Update all 8 repository consumers (detailed in §III.7 C6).
- **`[MODIFY]` `scripts/versionguard/carriers.go` & `scripts/versionguard/versionguard_test.go`:**
  Add `.antigravity-plugin/plugin.json` as a version carrier per plugin so any tag or version bump checks both `.claude-plugin/plugin.json` and `.antigravity-plugin/plugin.json`.
- **`[MODIFY]` `scripts/pluginparity/main.go` & `scripts/pluginparity/pluginparity_test.go`:**
  Assert that every plugin in `.claude-plugin/marketplace.json` carries `.antigravity-plugin/plugin.json` with identical `name` and `version` to `.claude-plugin/plugin.json`.
- **`[NEW]` `plugins/{prosthetic-conscience,frank-exchange-of-views,gray-area}/hooks.json`:**
  Antigravity named-hook group manifests at the plugin root. Because `jsonhook.executeCommandModeHook` (`0x80aff00`) sets `cmd.Dir` to `<plugin_root>` and does not set `PLUGIN_ROOT` or `PLUGIN_DATA` (and `(*pluginValidator).checkHookCommand` at `0x98cc640` warns on `$PLUGIN_ROOT`, `$PLUGIN_DATA`, and `$CLAUDE_PLUGIN_ROOT`), each entry uses `R="${PWD}"` so `B="$R/bin/<binary>"` and `F="$R/hooks/fetch-bin.sh"` resolve cleanly and bypass `checkHookCommand`'s static file-existence check when `bin/` is not yet populated:
  - `plugins/prosthetic-conscience/hooks.json`: registers `SessionStart` (`sc-sessionstart`), `PreInvocation` (`sc-sessionstart -event PreInvocation`), `PreTool` (`sc-pretooluse`, Antigravity-only matcher `read_url_content|read_browser_page|search_web|run_command` so `(*pluginValidator).checkHookMatcher` at `0x98ca820` emits `0` warnings), `PostTool` (`sc-posttooluse`, matcher `.*`), and `PostInvocation` (`sc-stop`).
  - `plugins/frank-exchange-of-views/hooks.json`: registers `PreInvocation` (`fetch-bin.sh ensure SessionStart`), `PreTool` (`feov-pretooluse`, matcher `.*`), and `PostTool` (`feov-posttoolusefailure`, matcher `.*`).
  - `plugins/gray-area/hooks.json`: registers `PreInvocation` (`gray-area-capture -event SessionStart`) and `PostInvocation` (`gray-area-capture -event Stop`).
- **`[MODIFY]` `scripts/validatejson/hooks.go`, `scripts/validatejson/hooks_test.go`, & `scripts/validatejson/main.go`:**
  Update `isPluginHooks` to match `plugins/*/.claude-plugin/hooks.json` (Claude schema) and add `isAntigravityPluginHooks` for `plugins/*/hooks.json` validating the top-level named-hook group schema and known Antigravity events (`SessionStart`, `PreInvocation`, `PreTool`, `PreToolUse`, `PostTool`, `PostToolUse`, `PostInvocation`, `Stop`).
- **`[MODIFY]` `scripts/fetchbingen/fetch-bin.sh`, `scripts/fetchbingen/main.go`, & `scripts/internal/hookfailures/failures.go`:**
  - In `scripts/fetchbingen/fetch-bin.sh`: resolve `ROOT=${CLAUDE_PLUGIN_ROOT:-${PLUGIN_ROOT:-$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)}}`; in `field()`, read `$ROOT/.claude-plugin/plugin.json` (falling back to `$ROOT/.antigravity-plugin/plugin.json`); in `speak()`, when `[ -n "${ANTIGRAVITY_CONVERSATION_ID:-}" ] || [ -n "${PLUGIN_ROOT:-}" ]` and `CLAUDE_PLUGIN_ROOT` is unset (Antigravity) and `SC_ENSURE_INTERNAL` is not `1`, emit `{"injectSteps":[{"systemMessage":{"systemMessage":"%s","metadata":{"user_markdown":"%s"}}}]}` on `SessionStart`/`PreInvocation` and nothing on `stdout` for `PreTool`/`PreToolUse`/`PostTool`/`PostToolUse`/`PostInvocation`/`Stop` (because `protojson.Unmarshal` with `DiscardUnknown: false` rejects `systemMessage` on those result messages).
  - In `scripts/fetchbingen/main.go`: check both `plugins/<p>/.claude-plugin/hooks.json` (`${CLAUDE_PLUGIN_ROOT}` guard) and `plugins/<p>/hooks.json` (`${PWD}` guard).
  - In `scripts/internal/hookfailures/failures.go` (`Ensure`): fall back from `os.Getenv("CLAUDE_PLUGIN_ROOT")` to `os.Getenv("PLUGIN_ROOT")` to `os.Executable()` grandparent / `os.Getwd()` and pass `SC_ENSURE_INTERNAL=1` in the child environment so `Ensure` continues parsing `{"systemMessage":"..."}` uniformly. Regenerate via `go -C scripts run ./fetchbingen` and `go -C scripts run ./hookfailgen`.

---

### III.3 Agent & Skill Cross-Harness Compatibility ([#1469](https://github.com/ctoforaday/special-circumstances/issues/1469))

- **`[MODIFY]` `scripts/agentgen/src/fragments/seat-tools.md` & Regenerate `plugins/frank-exchange-of-views/agents/*.md` (11 files):**
  Change `tools:` from a CSV string to a YAML sequence containing Claude, classic Antigravity, and `agy 1.4.10+` `oneharness` tool names (`read_file`, `write_file`, `edit_file`) while preserving FEOV's no-`WebFetch` / no-`read_url_content` / no-`read_browser_page` invariant:
  ```yaml
  tools: ["Read", "Write", "Edit", "Glob", "Grep", "Bash", "WebSearch", "ToolSearch", "view_file", "read_file", "write_to_file", "write_file", "replace_file_content", "edit_file", "find_by_name", "grep_search", "run_command", "search_web"]
  ```
  Run `go -C scripts run ./agentgen` to update all 11 generated `plugins/frank-exchange-of-views/agents/*.md` files.
- **`[MODIFY]` `plugins/prosthetic-conscience/agents/plan-auditor.md` & `plugins/prosthetic-conscience/agents/probe.md`:**
  - `plan-auditor.md`: `tools: ["Read", "Grep", "Glob", "Bash", "WebSearch", "WebFetch", "view_file", "read_file", "grep_search", "find_by_name", "run_command", "search_web", "read_url_content"]`.
  - `probe.md`: `tools: ["Read", "Write", "Bash", "view_file", "read_file", "write_to_file", "write_file", "run_command"]`.
- **`[MODIFY]` `scripts/frontmatter/main.go` & `scripts/frontmatter/frontmatter_test.go`:**
  Enforce that `tools` and `skills` in `plugins/*/agents/*.md` decode as YAML sequences (`[]any`), failing on scalar strings so `agents 0` cannot recur.
- **`[MODIFY]` Six CLI-Invoking `SKILL.md` Files:**
  Update `plugins/prosthetic-conscience/skills/doctor/SKILL.md`, `plugins/gray-area/skills/{audit-checkpoint,audit-pr-body,audit-repetition,audit-seat-coverage}/SKILL.md`, and `plugins/frank-exchange-of-views/skills/research/SKILL.md` to invoke compiled plugin binaries via:
  ```sh
  B="${CLAUDE_PLUGIN_ROOT}/bin/<bin>"; [ -x "$B" ] || B="<bin>"; "$B"
  ```

---

### III.4 Unified Hook Wire-Adapter Library (`scripts/internal/hookcore` & `scripts/hookcoregen` — [#1470](https://github.com/ctoforaday/special-circumstances/issues/1470))

#### III.4.1 Process Boundary Normalization Matrix (`env`, `stdin`, `stdout`, `stderr`, `exit`)

| Process Boundary | Claude Code (`2.1.309`) | Google Antigravity (`agy 1.4.12` — `exa.hooks_pb`) | `hookcore` Normalization Contract |
| :--- | :--- | :--- | :--- |
| **`env` (Paths)** | `CLAUDE_PROJECT_DIR`, `CLAUDE_PLUGIN_ROOT`, `CLAUDE_PLUGIN_DATA` | `ANTIGRAVITY_CONVERSATION_ID` (set on all hook subprocesses at `0x80b0216`); `cmd.Dir` is `<plugin_root>` (`0x80b0120`); `PLUGIN_ROOT` and `PLUGIN_DATA` are set only for MCP / agent `run_command` | `PluginRoot`: `CLAUDE_PLUGIN_ROOT` → `PLUGIN_ROOT` → `os.Executable()` grandparent → `os.Getwd()` (when `.antigravity-plugin/plugin.json` exists).<br>`PluginDataDir`: `CLAUDE_PLUGIN_DATA` → `PLUGIN_DATA` → `filepath.Join(filepath.Dir(filepath.Dir(artifactDirectoryPath)), "plugin-data", pluginName)` → `<ProjectDir>/.claude`.<br>`ProjectDir`: `CLAUDE_PROJECT_DIR` → `cwd` → `workspacePaths[0]` → git root of tool arg `Cwd`/`Path`/`TargetFile`/`AbsolutePath` (never `os.Getwd()`). |
| **`stdin` (Envelope)** | snake_case: `session_id`, `cwd`, `transcript_path`, `agent_id`, `agent_type`, `agent_transcript_path`, `hook_event_name`, `stop_hook_active`, `is_interrupt`, `prompt` | camelCase (`HookArgsCommon` at `0x5f93cc8`): `conversationId`, `workspacePaths`, `transcriptPath`, `artifactDirectoryPath`, `parentConversationId`, `rootConversationId`, `agentName`, `executionNum`, `terminationReason`, `lastUserInput` (`field 8`, `1.4.7+`) | Sniff `ANTIGRAVITY_CONVERSATION_ID` env var or JSON keys (`conversationId`/`toolCall`/`workspacePaths` vs `session_id`/`tool_name`/`cwd`) to set `EngineClaude` or `EngineAntigravity` and populate canonical `Request` fields. |
| **`stdin` (Tool Call)** | `tool_name` (`Bash`, `Write`, `Edit`, `Read`, `WebSearch`, `WebFetch`), `tool_input` (snake_case), `tool_response` (`PostToolUse`), `error` (`PostToolUseFailure`) | `toolCall.name` (`run_command`, `write_to_file`/`write_file`, `replace_file_content`/`multi_replace_file_content`/`edit_file`, `view_file`/`read_file`, `search_web`, `read_url_content`/`read_browser_page`), `toolCall.args` (PascalCase), `result` (`field 4`, `1.4.12+`) and `error` (`field 3`) on `PostTool` | Map `toolCall.name` to canonical `ToolName`, populate `req.ToolResult` (`tool_response` \| `result`), and synthesize canonical snake_case keys (`command`, `file_path`, `content`, `old_string`, `new_string`, `query`, `url`) inside `req.ToolInput` alongside original fields. |
| **`stdout` (`PreTool` / `PreToolUse`)** | `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"\|"deny"\|"ask","permissionDecisionReason":"...","updatedInput":{...}},"systemMessage":"..."}` | `PreToolHookResult`: `{"decision":"allow"\|"deny"\|"ask","reason":"...","deny_reason":"...","overwrite":{...}}` | Encode `Decision`, `Reason`, and `UpdatedCommand` (`updatedInput.command` on Claude; `overwrite.CommandLine` on Antigravity). On Antigravity, if `Say` or `rec.Settle()` is non-empty, enqueue it to `PreInvocation`. |
| **`stdout` / `stderr` / `exit` (`PostTool` / `PostToolUse` & `PostToolUseFailure`)** | `sc-posttooluse` (`qualitygate`): `Feedback` to **`stderr`** + **`exit 2`**, `{"systemMessage":"..."}` on `stdout`.<br>`sc-strike-counter`: `hookSpecificOutput.additionalContext` + `systemMessage` on `stdout`, `exit 0`. | `PostToolHookResult` only accepts `{}` (or empty `stdout`), and `exit 2` is treated as a hook execution failure rather than model feedback. | **Claude:** preserve `stderr` + `exit 2` for `Feedback` and `hookSpecificOutput` for `AdditionalContext`.<br>**Antigravity:** append `Feedback`, `AdditionalContext`, and `Say`/`rec.Settle()` to `<PluginDataDir>/pending_<SessionID>.jsonl`, write diagnostics to `stderr`, emit nothing/`{}` on `stdout`, and return **`exit 0`**. |
| **`stdout` (`SessionStart` & `PreInvocation`)** | `SessionStart`: `{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"...","watchPaths":[...]},"systemMessage":"..."}` | `SessionStartHookResult` / `PreInvocationHookResult`: `{"injectSteps":[{"ephemeralMessage":"..."},{"systemMessage":{"systemMessage":"...","metadata":{"user_markdown":"..."}}},{"errorMessage":{"modelMessage":"...","userMessage":"..."}}],"terminationBehavior":"terminate"\|"force_continue"}` | On Antigravity `SessionStart`, emit `AdditionalContext` (`ephemeralMessage`) + `Say` (`systemMessage` with `metadata.user_markdown`). On `PreInvocation` (`sc-sessionstart -event PreInvocation`), atomically read and unlink `<PluginDataDir>/pending_<SessionID>.jsonl` and emit `injectSteps` (plus `terminationBehavior: "terminate"` + `errorMessage` when `resp.TerminateTurn` is set). |
| **`stdout` (`PostInvocation` / `Stop`)** | `{"hookSpecificOutput":{"hookEventName":"Stop","additionalContext":"..."},"systemMessage":"..."}` | `PostInvocationHookResult`: `{"decision":"continue","reason":"...","terminationBehavior":"terminate"\|"force_continue"}` (or empty `stdout` when silent) | Translate `AdditionalContext` on `Stop` into `{"decision":"continue","reason":"..."}` on Antigravity; when silent, write nothing to `stdout`. |

#### III.4.2 Tool Name & Argument Normalization Table

| Tool Category | Claude `tool_name` | Antigravity `toolCall.name` (Classic + `oneharness`) | Claude `tool_input` Key | Antigravity `toolCall.args` Key | Synthesized Canonical `req.ToolInput` Key |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Shell Command** | `Bash` | `run_command` | `command`, `description` | `CommandLine`, `Cwd`, `toolAction` / `toolSummary` | `"command"`, `"cwd"`, `"description"` |
| **Write File** | `Write` | `write_to_file`, `write_file` | `file_path`, `content` | `TargetFile` / `Path`, `CodeContent` / `Content` | `"file_path"`, `"content"` |
| **Edit File** | `Edit` | `replace_file_content`, `multi_replace_file_content`, `edit_file` | `file_path`, `old_string`, `new_string`, `replace_all` | `TargetFile` / `Path`, `TargetContent` / `Target`, `ReplacementContent` / `Replacement`, `AllowMultiple` | `"file_path"`, `"old_string"`, `"new_string"`, `"replace_all"` |
| **Read File** | `Read` | `view_file`, `read_file` | `file_path` | `AbsolutePath` / `Path` | `"file_path"` |
| **Web Search** | `WebSearch` | `search_web` | `query` | `query` / `Query` | `"query"` |
| **Web Fetch** | `WebFetch` | `read_url_content`, `read_browser_page` | `url` | `Url` | `"url"` |
| **File Search** | `Glob` | `find_by_name` | `path`, `pattern` | `SearchDirectory`, `Pattern` | `"path"`, `"pattern"` |
| **Grep Search** | `Grep` | `grep_search` | `path`, `pattern` | `SearchPath`, `Query` | `"path"`, `"pattern"` |

#### III.4.3 Pure-Stdlib Interface of `scripts/internal/hookcore`

Because `scripts/hookcoregen` copies `hookcore.go` byte-identically into `prosthetic-conscience`, `frank-exchange-of-views`, and `gray-area`, `hookcore` imports **only the Go standard library** and accepts `*hookfailures.Recorder` via a structural `Settler` interface (zero module-local imports):

```go
package hookcore

import (
	"encoding/json"
	"io"
)

type Engine string

const (
	EngineClaude      Engine = "claude"
	EngineAntigravity Engine = "antigravity"
)

type Decision string

const (
	DecisionNone  Decision = ""
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
	DecisionAsk   Decision = "ask"
)

// Settler is satisfied by *hookfailures.Recorder in each plugin module without importing it.
type Settler interface {
	Settle() string
	Persist()
}

type Request struct {
	Engine              Engine
	Event               string
	Raw                 []byte          // Exact stdin bytes (preserved for secretsgate #211 raw scan)
	Parsed              bool            // Whether stdin parsed as valid JSON
	ProjectDir          string          // CLAUDE_PROJECT_DIR -> cwd -> workspacePaths[0] -> tool-arg git root
	PluginRoot          string          // CLAUDE_PLUGIN_ROOT -> PLUGIN_ROOT -> os.Executable() ../.. -> os.Getwd()
	PluginDataDir       string          // CLAUDE_PLUGIN_DATA -> PLUGIN_DATA -> <appDataDir>/plugin-data/<plugin> -> <ProjectDir>/.claude
	SessionID           string          // session_id | conversationId
	TranscriptPath      string          // transcript_path | transcriptPath
	ArtifactDirPath     string          // artifactDirectoryPath (Antigravity)
	AgentID             string          // agent_id | (conversationId when parentConversationId != "")
	AgentType           string          // agent_type | agentName
	AgentTranscriptPath string          // agent_transcript_path | (transcriptPath when subagent)
	ToolName            string          // Canonical ("Bash", "Write", "Edit", "Read", "WebSearch", "WebFetch", "Glob", "Grep")
	RawToolName         string          // Original wire name ("run_command", "write_to_file", "write_file", ...)
	ToolInput           json.RawMessage // Normalized JSON carrying canonical snake_case keys + original fields
	ToolResult          json.RawMessage // tool_response (Claude) | PostToolHookArgs.result (Antigravity 1.4.12+)
	FilePath            string          // Extracted file path (file_path | path | Path | TargetFile | AbsolutePath)
	Command             string          // Extracted shell command (command | CommandLine)
	ToolError           string          // error (PostToolUseFailure on Claude; PostTool.error on Antigravity)
	LastUserInput       string          // prompt (Claude) | lastUserInput (Antigravity 1.4.7+)
	IsInterrupt         bool            // is_interrupt | (terminationReason == "USER_CANCELED")
	StopHookActive      bool            // stop_hook_active | (executionNum > 0)
	TerminationReason   string          // PostInvocation.terminationReason on Antigravity
	Trigger             string          // PreCompact/PostCompact trigger
	CustomInstructions  string          // PreCompact custom_instructions
	CompactSummary      string          // PostCompact compact_summary
}

type Response struct {
	Decision          Decision
	Reason            string
	UpdatedCommand    string
	Feedback          string
	AdditionalContext string
	WatchPaths        []string
	PreCompactSteer   string
	Say               string
	Warnings          string
	TerminateTurn     bool // PreInvocation: emits terminationBehavior="terminate" + errorMessage on Antigravity
}

func Read(event string, stdin io.Reader) Request
func ProjectDir(env, cwd string, workspacePaths []string, toolDir string) string
func Respond(req Request, resp Response, stdout, stderr io.Writer, rec Settler) int
```

- **`[NEW]` `scripts/hookcoregen/main.go` & `main_test.go`:** Copies `scripts/internal/hookcore/*.go` (with `// Code generated by scripts/hookcoregen from scripts/internal/hookcore/<name>. DO NOT EDIT.`) into all three plugin modules; `-check` verifies zero drift.
- **`[MODIFY]` `scripts/check/gates.go`, `scripts/check/parity_test.go`, & `.github/workflows/hooks.yml`:** Register `hookcoregen` (`go run ./hookcoregen -check`, `ciJob: "debate-sim"`). `scripts/deadcode/main.go:142` already matches `^// Code generated by \S+ from (\S+\.go)\. DO NOT EDIT\.$` with no change needed.

---

### III.5 `prosthetic-conscience` Hook Migration ([#1471](https://github.com/ctoforaday/special-circumstances/issues/1471))

- **`[MODIFY]` `plugins/prosthetic-conscience/tools/internal/hookenv/hookenv.go` & `plugins/prosthetic-conscience/tools/internal/hookinvocation/projectroot_test.go`:**
  `hookenv.ProjectDir` delegates to `hookcore.ProjectDir` while preserving its existing 2-arg caller signature (`ProjectDir(env, payloadDir)` where `req.ProjectDir` supplies the normalized payload directory), keeping `TestNoHookReadsTheEnvironmentVariableWithoutResolvingThroughThisPackage` (`hookenv_test.go`) green. Update `withoutProjectRoot` in `projectroot_test.go` to strip `PLUGIN_ROOT=`, `PLUGIN_DATA=`, and `ANTIGRAVITY_CONVERSATION_ID=`.
- **`[MODIFY]` `plugins/prosthetic-conscience/tools/internal/sessionstart/main.go`:**
  Support `-event SessionStart | PreInvocation` (default `SessionStart`). Pass `req.PluginRoot` into `Units(req.PluginRoot)` on `SessionStart` (fixing `StageManifest` when `CLAUDE_PLUGIN_ROOT` is unset); on `PreInvocation`, flush `<PluginDataDir>/pending_<SessionID>.jsonl` and due `hookfailures` entries via `hookcore.Respond`.
- **`[MODIFY]` `plugins/prosthetic-conscience/tools/internal/pretooluse/main.go`:**
  Decode via `hookcore.Read("PreToolUse", stdin)` into `hookunit.Ctx` (preserving `req.Raw` and `req.Parsed` for `secretsgate` `#211`) and emit via `hookcore.Respond`.
- **`[MODIFY]` `plugins/prosthetic-conscience/tools/internal/posttooluse/main.go` & `internal/strikecounter/main.go`:**
  Decode via `hookcore.Read`. When `req.Engine == EngineAntigravity && req.ToolError != ""`, `sc-posttooluse` also runs `strikecounter` for the failed tool call and passes any 3-strike `AdditionalContext` + `qualitygate` `Feedback` to `hookcore.Respond` (which queues them to `<PluginDataDir>/pending_<SessionID>.jsonl` and exits `0`).
- **`[MODIFY]` `internal/stopnudge`, `internal/checkpointseal`, `internal/filechangedrearm`, `internal/postcompactobserve`:**
  Decode via `hookcore.Read` and emit via `hookcore.Respond`.

---

### III.6 `frank-exchange-of-views` & `gray-area` Hook Migration ([#1472](https://github.com/ctoforaday/special-circumstances/issues/1472))

- **`[MODIFY]` `plugins/frank-exchange-of-views/tools/internal/hookcmd/hookcmd.go`:**
  - `Pre` (`feov-pretooluse`): read `hookcore.Read("PreToolUse", stdin)` (normalizing `run_command` → `Bash` and `CommandLine` → `command`) and emit `Allow` + `UpdatedCommand` (`updatedInput.command` on Claude; `overwrite.CommandLine` on Antigravity), `Ask`, or `Deny` via `hookcore.Respond`.
  - `PostFailure` (`feov-posttoolusefailure`): read `hookcore.Read(event, stdin)`; on Antigravity `PostTool` / `PostToolUse`, return immediately if `req.ToolError == ""` and record the failure when `req.ToolError != ""`.
  - `feov-subagentstart` & `feov-subagentstop`: read normalized `AgentID`, `AgentType`, `AgentTranscriptPath`, and `ProjectDir` via `hookcore.Read`.
- **`[MODIFY]` `plugins/frank-exchange-of-views/tools/integration/surface/hookgraph_test.go`:**
  Add `"internal/hookcore"` (with its zero-`init()`, stdlib-only cost statement) to `hookLocalPackages` for `./cmd/feov-pretooluse`, `./cmd/feov-posttoolusefailure`, `./cmd/feov-subagentstart`, and `./cmd/feov-subagentstop`.
- **`[MODIFY]` `plugins/gray-area/tools/cmd/gray-area-capture/main.go`:**
  Read `hookcore.Read(*event, stdin)` and emit `healed` + `r.Settle()` via `hookcore.Respond`.

---

### III.7 Consumer Censuses (`origin/main` @ `61107ba4`)

<details><summary>C1 — <code>git grep -nE '\.claude-plugin/plugin\.json|marketplace\.json' -- scripts/ plugins/ .github/ docs/</code> (31 lines)</summary>

```
.github/workflows/hooks.yml:973:          printf '{\n  "name": "fixture",\n  "version": "1.2.3"\n}\n' > "$root/.claude-plugin/plugin.json"
docs/local-development.md:52:`.claude-plugin/marketplace.json` pins three of the four plugins to a **release tag**, via a
plugins/frank-exchange-of-views/hooks/fetch-bin.sh:14:# The release is the tag <name>--v<version>, both read from .claude-plugin/plugin.json — never
plugins/frank-exchange-of-views/hooks/fetch-bin.sh:33:	[ -n "$ROOT" ] && [ -f "$ROOT/.claude-plugin/plugin.json" ] || return 0
plugins/frank-exchange-of-views/hooks/fetch-bin.sh:34:	sed -n "s/^  \"$1\"[[:space:]]*:[[:space:]]*\"\([^\"]*\)\".*/\1/p" "$ROOT/.claude-plugin/plugin.json" | head -n 1
plugins/frank-exchange-of-views/requirements.json:4:  "_eventSchema_comment": "... It lives here rather than in .claude-plugin/plugin.json ..."
plugins/frank-exchange-of-views/tools/internal/cli/versionsync_test.go:52:// The plugin manifest must NOT carry the epoch as well: .claude-plugin/plugin.json has a
plugins/frank-exchange-of-views/tools/internal/cli/versionsync_test.go:70:			t.Errorf("%s is in .claude-plugin/plugin.json. It belongs in requirements.json alone: "+
plugins/gray-area/hooks/fetch-bin.sh:14,33,34: (generated from scripts/fetchbingen/fetch-bin.sh)
plugins/prosthetic-conscience/hooks/fetch-bin.sh:14,33,34: (generated from scripts/fetchbingen/fetch-bin.sh)
plugins/prosthetic-conscience/skills/doctor/SKILL.md:9:2. **Bootstrap** ... `${CLAUDE_PLUGIN_ROOT}/.claude-plugin/plugin.json` ...
scripts/bootstrap-plugins.sh:30:# Every plugin in .claude-plugin/marketplace.json.
scripts/fetchbingen/fetch-bin.sh:13,32,33: (authored fetch-bin.sh)
scripts/marketplacegen/main.go:2,30:	manifest = ".claude-plugin/marketplace.json"
scripts/pluginparity/main.go:67,68,107:	if err := json.Unmarshal([]byte(read(".claude-plugin/marketplace.json")), &m); err != nil {
scripts/universe.sh:349,419:    local pj="$root/.claude-plugin/plugin.json"
scripts/validatejson/main.go:3:// Dev tooling for this repository only. A broken plugin.json / marketplace.json /
scripts/versionguard/versionguard_test.go:87,124,164,235:	write(t, dir, "plugins/demo/.claude-plugin/plugin.json", ...)
```
</details>

- **Changes:** `scripts/fetchbingen/fetch-bin.sh` (and the 3 generated `plugins/*/hooks/fetch-bin.sh` copies), `scripts/pluginparity/main.go`, `scripts/versionguard/carriers.go` + `versionguard_test.go`, and `plugins/*/.claude-plugin/plugin.json` (adding `"hooks": "./.claude-plugin/hooks.json"` in the 3 hook-bearing plugins) (§III.2).
- **Unchanged:** `.github/workflows/hooks.yml:973`, `docs/local-development.md:52`, `requirements.json:4`, `versionsync_test.go:52,70`, `doctor/SKILL.md:9` (`.claude-plugin/plugin.json` remains present in every plugin checkout), `bootstrap-plugins.sh:30`, `marketplacegen/main.go`, `universe.sh`.

<details><summary>C2 — <code>git grep -nE 'CLAUDE_PLUGIN_ROOT|CLAUDE_PROJECT_DIR|CLAUDE_PLUGIN_DATA' -- plugins/ scripts/</code> (62 lines)</summary>

```
plugins/frank-exchange-of-views/hooks/fetch-bin.sh:15,21,169
plugins/frank-exchange-of-views/hooks/hooks.json:8,18,28,38,48
plugins/frank-exchange-of-views/skills/research/SKILL.md:25
plugins/frank-exchange-of-views/tools/internal/hookfailures/failures.go:269,271
plugins/frank-exchange-of-views/tools/releasegate/fuzz/hookinvocation_test.go:134,178
plugins/gray-area/hooks/fetch-bin.sh:15,21,169
plugins/gray-area/hooks/hooks.json:8,18,28,38
plugins/gray-area/requirements.json:30
plugins/gray-area/skills/{audit-checkpoint,audit-pr-body,audit-repetition,audit-seat-coverage}/SKILL.md:11 (and 20/21/27)
plugins/gray-area/tools/cmd/gray-area-capture/main.go:304,335,382
plugins/gray-area/tools/cmd/gray-area-capture/main_test.go:189,190
plugins/gray-area/tools/internal/hookfailures/failures.go:269,271
plugins/prosthetic-conscience/hooks/fetch-bin.sh:15,21,169
plugins/prosthetic-conscience/hooks/hooks.json:9,20,30,40,50,60,70,80,90,100
plugins/prosthetic-conscience/requirements.json:63,64
plugins/prosthetic-conscience/skills/doctor/SKILL.md:8,9
plugins/prosthetic-conscience/tools/internal/checkpointseal/main.go:599
plugins/prosthetic-conscience/tools/internal/doctor/fetchbin_contract_test.go:68,78
plugins/prosthetic-conscience/tools/internal/doctor/main.go:735
plugins/prosthetic-conscience/tools/internal/filechangedrearm/main.go:279
plugins/prosthetic-conscience/tools/internal/hookenv/hookenv.go:4,17,69
plugins/prosthetic-conscience/tools/internal/hookenv/hookenv_test.go:63,81,109,124,126,130
plugins/prosthetic-conscience/tools/internal/hookfailures/failures.go:271,273
plugins/prosthetic-conscience/tools/internal/hookinvocation/invocation_test.go:30,33,70,119,206,207
plugins/prosthetic-conscience/tools/internal/hookinvocation/projectroot_test.go:30,90
plugins/prosthetic-conscience/tools/internal/postcompactobserve/main.go:260
plugins/prosthetic-conscience/tools/internal/posttooluse/main.go:111
plugins/prosthetic-conscience/tools/internal/pretooluse/main.go:128
plugins/prosthetic-conscience/tools/internal/qualitygate/main.go:123
plugins/prosthetic-conscience/tools/internal/qualitygate/main_test.go:57
plugins/prosthetic-conscience/tools/internal/sessionstart/main.go:160
plugins/prosthetic-conscience/tools/internal/stopnudge/machinereader_test.go:123
plugins/prosthetic-conscience/tools/internal/stopnudge/main.go:206
plugins/prosthetic-conscience/tools/internal/stopnudge/main_test.go:304,325
plugins/prosthetic-conscience/tools/internal/strikecounter/failures_test.go:68
plugins/prosthetic-conscience/tools/internal/strikecounter/main.go:230
plugins/prosthetic-conscience/tools/internal/toolchainnudge/failures_test.go:76
plugins/prosthetic-conscience/tools/internal/toolchainnudge/main.go:113
scripts/bootstrap-plugins.sh:68,108
scripts/fetchbingen/fetch-bin.sh:15,21,169
scripts/fetchbingen/main.go:36,44,146,176
scripts/fetchbingen/main_test.go:12,13,64,108
scripts/internal/hookfailures/failures.go:269,271
scripts/internal/hookfailures/failures_test.go:451,465
scripts/universe.sh:224,327,363,439
```
</details>

- **Changes:** `scripts/fetchbingen/fetch-bin.sh` (+ 3 generated copies), `scripts/internal/hookfailures/failures.go` (+ 3 generated copies), the 6 `SKILL.md` files (`doctor`, 4 `gray-area/audit-*`, `feov/research`), `gray-area-capture/main.go`, `sessionstart/main.go`, `pretooluse/main.go`, `posttooluse/main.go`, `strikecounter/main.go`, `stopnudge/main.go`, `checkpointseal/main.go`, `filechangedrearm/main.go`, `postcompactobserve/main.go`, `hookenv/hookenv.go`, `hookinvocation/projectroot_test.go`, and the 8 `hooks/hooks.json` consumers listed in C6 below (§III.2–§III.6).
- **Unchanged:** `doctor/main.go:735` (already falls back to `os.Executable()` grandparent at `doctor/main.go:341-365`), `bootstrap-plugins.sh`, `universe.sh`.

<details><summary>C3 — <code>git grep -nE '^tools:|^skills:' -- plugins/*/agents/ scripts/agentgen/src/</code> (32 lines)</summary>

```
plugins/frank-exchange-of-views/agents/{blue-researcher,blue-synthesizer,lead-judge,red-chair,red-lens-adversary,red-lens-architecture,red-lens-computation,red-lens-dark-side,red-lens-evidence,red-lens-logic,red-lens-voice}.md:4:tools: Read, Write, Edit, Glob, Grep, Bash, WebSearch, ToolSearch
plugins/frank-exchange-of-views/agents/{...}.md:5:skills: [...]
plugins/prosthetic-conscience/agents/plan-auditor.md:4:tools: Read, Grep, Glob, Bash, WebSearch, WebFetch
plugins/prosthetic-conscience/agents/plan-auditor.md:5:skills: [critical-stance, design-by-contract, spec-driven-development, terse-communication, complete-the-concept]
plugins/prosthetic-conscience/agents/probe.md:4:tools: Read, Write, Bash
plugins/prosthetic-conscience/agents/probe.md:5:skills: [critical-stance, terse-communication, design-by-contract]
scripts/agentgen/src/blue-researcher.md:5:skills: [...]
scripts/agentgen/src/blue-synthesizer.md:5:skills: [...]
scripts/agentgen/src/fragments/lens-frontmatter.md:2:skills: [...]
scripts/agentgen/src/fragments/seat-tools.md:1:tools: Read, Write, Edit, Glob, Grep, Bash, WebSearch, ToolSearch
scripts/agentgen/src/lead-judge.md:5:skills: [...]
scripts/agentgen/src/red-chair.md:5:skills: [...]
```
</details>

- **Changes:** `scripts/agentgen/src/fragments/seat-tools.md` (regenerating all 11 `plugins/frank-exchange-of-views/agents/*.md` via `go -C scripts run ./agentgen`), `plugins/prosthetic-conscience/agents/plan-auditor.md`, and `plugins/prosthetic-conscience/agents/probe.md` (§III.3).

<details><summary>C4 — <code>git grep -nE 'hookenv\.ProjectDir|hookfailures\.(New|Emit|Ensure)' -- plugins/ scripts/</code> (54 lines)</summary>

```
plugins/frank-exchange-of-views/tools/internal/hookcmd/hookcmd.go:63:	rec := hookfailures.New("frank-exchange-of-views", hookName, event, time.Now(), os.Stderr)
plugins/frank-exchange-of-views/tools/internal/hookcmd/hookcmd.go:81:	hookfailures.Emit(stdout, rec.Settle())
plugins/gray-area/tools/cmd/gray-area-capture/main.go:281:		msg, err := ensure("SessionStart")
plugins/gray-area/tools/cmd/gray-area-capture/main.go:290:	hookfailures.Emit(stdout, hookfailures.JoinMessages(healed, r.Settle()))
plugins/prosthetic-conscience/tools/internal/checkpointseal/main.go:539:	projectDir = hookenv.ProjectDir(projectDir, in.CWD)
plugins/prosthetic-conscience/tools/internal/checkpointseal/main.go:542:	rec := hookfailures.New("prosthetic-conscience", "sc-checkpoint-seal", in.HookEventName, now, stderr)
plugins/prosthetic-conscience/tools/internal/filechangedrearm/main.go:191:	projectDir = hookenv.ProjectDir(projectDir, in.CWD)
plugins/prosthetic-conscience/tools/internal/filechangedrearm/main.go:193:	rec := hookfailures.New("prosthetic-conscience", "sc-filechanged-rearm", "FileChanged", now, stderr)
plugins/prosthetic-conscience/tools/internal/postcompactobserve/main.go:193,195
plugins/prosthetic-conscience/tools/internal/posttooluse/main.go:84,85,98
plugins/prosthetic-conscience/tools/internal/pretooluse/main.go:70,104,105
plugins/prosthetic-conscience/tools/internal/qualitygate/main.go:316
plugins/prosthetic-conscience/tools/internal/sessionstart/main.go:66,105,115
plugins/prosthetic-conscience/tools/internal/stopnudge/main.go:128,129
plugins/prosthetic-conscience/tools/internal/strikecounter/main.go:169,170
scripts/fetchbingen/main.go:180,196,206,208
```
</details>

- **Changes:** Every non-test hook caller above routes through `hookcore.Read` and `hookcore.Respond` while keeping `hookenv.ProjectDir` and `hookfailures.New`/`Ensure` intact (§III.5–§III.6).

<details><summary>C5 — <code>git grep -nE 'hookfailgen|buildidgen|fetchbingen|agentgen' -- scripts/ .github/</code> (52 lines)</summary>

```
.github/workflows/hooks.yml:1508:        run: go run ./buildidgen -check
.github/workflows/hooks.yml:1511:        run: go run ./fetchbingen -check
.github/workflows/hooks.yml:1514:        run: go run ./hookfailgen -check
.github/workflows/hooks.yml:1526:        run: go run ./agentgen -check
scripts/check/gates.go:249,251,253,261
scripts/check/parity_test.go:78
scripts/deadcode/main.go:21,142
```
</details>

- **Changes:** `.github/workflows/hooks.yml`, `scripts/check/gates.go`, and `scripts/check/parity_test.go` add `hookcoregen` alongside `buildidgen` and `hookfailgen` (§III.4.3).

<details><summary>C6 — <code>git grep -nE 'hooks/hooks\.json|"hooks\.json"' -- plugins/ scripts/ .github/</code> (18 lines — A2 Option 1 relocation to <code>.claude-plugin/hooks.json</code>)</summary>

```
.github/workflows/hooks.yml:1548:            const files = fs.globSync('plugins/*/hooks/hooks.json');
plugins/frank-exchange-of-views/tools/integration/surface/hookgraph_test.go:260:	manifest := filepath.Join(root, "plugins", "frank-exchange-of-views", "hooks", "hooks.json")
plugins/frank-exchange-of-views/tools/internal/setup/hookprovenance.go:129:	b, err := os.ReadFile(filepath.Join(versionDir, "hooks", "hooks.json"))
plugins/frank-exchange-of-views/tools/internal/setup/hookprovenance_test.go:39:	if err := os.WriteFile(filepath.Join(dir, "hooks", "hooks.json"), []byte(hooks), 0o644); err != nil {
plugins/frank-exchange-of-views/tools/releasegate/fuzz/hookinvocation_test.go:50:	hooksPath, err := repotree.Plugin("hooks", "hooks.json")
plugins/prosthetic-conscience/tools/internal/hookinvocation/invocation_test.go:47:	b, err := os.ReadFile(filepath.Join("..", "..", "..", "hooks", "hooks.json"))
scripts/fetchbingen/main.go:89:		hj, err := os.ReadFile(filepath.Join(root, "plugins", p, "hooks", "hooks.json"))
scripts/fetchbingen/main.go:91,156,212: (error messages naming plugins/%s/hooks/hooks.json)
scripts/rulesweep/sweep.go:57:	regexp.MustCompile(`/hooks/hooks\.json$`),
scripts/rulesweep/sweep_test.go:141:		"plugins/x/hooks/hooks.json",
scripts/validatejson/hooks.go:46:	ok, _ := path.Match("plugins/*/hooks/hooks.json", filepath.ToSlash(p))
scripts/validatejson/hooks_test.go:62-67:		"plugins/gray-area/hooks/hooks.json": true, ...
scripts/validatejson/main.go:68:		fmt.Fprintln(os.Stderr, "validate-json: no plugins/*/hooks/hooks.json is tracked — the pattern is wrong, not the tree")
```
</details>

- **Changes (8 sites):**
  1. `scripts/fetchbingen/main.go` & `main_test.go`: check both `plugins/<p>/.claude-plugin/hooks.json` (`${CLAUDE_PLUGIN_ROOT}` guard) and `plugins/<p>/hooks.json` (`${PWD}` guard).
  2. `scripts/validatejson/hooks.go`, `hooks_test.go`, & `main.go`: match `plugins/*/.claude-plugin/hooks.json` (Claude schema) and `plugins/*/hooks.json` (Antigravity named-hook group schema).
  3. `scripts/rulesweep/sweep.go` & `sweep_test.go`: change `/hooks/hooks\.json$` → `/hooks\.json$` so both `.claude-plugin/hooks.json` and root `hooks.json` remain gated as protocol surfaces.
  4. `.github/workflows/hooks.yml:1548`: update `fs.globSync` to `['plugins/*/.claude-plugin/hooks.json', 'plugins/*/hooks.json']` and run both `claude plugin validate` and `agy plugin validate --strict`.
  5. `plugins/prosthetic-conscience/tools/internal/hookinvocation/invocation_test.go:47`: read `../../../.claude-plugin/hooks.json` (plus companion test over `../../../hooks.json`).
  6. `plugins/frank-exchange-of-views/tools/integration/surface/hookgraph_test.go:260`: read `.claude-plugin/hooks.json` (and `hooks.json`).
  7. `plugins/frank-exchange-of-views/tools/releasegate/fuzz/hookinvocation_test.go:50`: `repotree.Plugin(".claude-plugin", "hooks.json")`.
  8. `plugins/frank-exchange-of-views/tools/internal/setup/hookprovenance.go:129` & `hookprovenance_test.go:39`: read `filepath.Join(versionDir, ".claude-plugin", "hooks.json")` first and fall back to `filepath.Join(versionDir, "hooks", "hooks.json")` for older cached release versions in `~/.claude/plugins/cache/`.

---

## IV. Risk & Mitigation

| # | Risk | Likelihood | Impact | Complexity | Mitigation |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **R1** | `sc-pretooluse` fails closed on unparseable non-secret JSON or unmapped tools, breaking `#211` and `TestTheShippedBinariesHonourTheIOContract` | Medium | High (blocks valid tool calls) | Low | `hookcore.Request` preserves `Raw []byte` and `Parsed bool`; `secretsgate` scans `req.Raw` when `!req.Parsed` and denies only on secret match (§III.4.3, §V.4). |
| **R2** | Headless `agy -p` sends `"workspacePaths": []` (empty slice), causing index panic or `os.Getwd()` guessing | High | High (panic or cross-project state corruption) | Low | `hookcore.ProjectDir` length-checks `workspacePaths`, falls back to git root of tool arg `Cwd`/`Path`/`TargetFile`/`AbsolutePath`, and never calls `os.Getwd()` (§III.4.1, §V.4). |
| **R3** | Antigravity `protojson.Unmarshal` (`DiscardUnknown: false`) rejects Claude `systemMessage` or `hookSpecificOutput` fields on `PreTool` / `PostTool` / `PostInvocation` | High | High (hook execution error in `agy`) | Low | `hookcore.Respond` and `fetch-bin.sh` emit only `.exa.hooks_pb`-compliant fields on `EngineAntigravity` and empty `stdout` when silent (§III.2, §III.4.1). |
| **R4** | Multiple `write_to_file` / `write_file` / `replace_file_content` / `edit_file` calls in one turn overwrite each other's lint feedback or re-inject stale findings on later turns | Medium | Medium (lost or looping lint advisories) | Low | Append JSON lines to `<PluginDataDir>/pending_<SessionID>.jsonl` and atomically rename-and-unlink on `PreInvocation` flush (§III.4.1, §V.5). |
| **R5** | Adding `internal/hookcore` to FEOV hooks pulls in heavy transitive dependencies or fails `TestHookBinaryLinksNothingExpensive` | Low | High (3× per-call hook latency regression) | Low | `hookcore` uses Go stdlib only (`Settler` structural interface, zero `init()`); `hookgraph_test.go` allowlists `"internal/hookcore"` and forbids `record`/`sqlite`/`proto`/`cobra` (§III.4.3, §III.6). |
| **R6** | Version drift between `.claude-plugin/plugin.json` and `.antigravity-plugin/plugin.json` | Medium | Medium (stale release tag / broken `fetch-bin.sh`) | Low | Register `.antigravity-plugin/plugin.json` in both `scripts/versionguard/carriers.go` and `scripts/pluginparity/main.go` (§III.2, §V.2). |
| **R7** | Editing `plugins/frank-exchange-of-views/agents/*.md` directly instead of `scripts/agentgen/src/fragments/seat-tools.md` fails `agentgen -check` | High | Medium (CI gate failure) | Low | Edit `scripts/agentgen/src/fragments/seat-tools.md` and regenerate via `go -C scripts run ./agentgen` (§III.3, §V.2). |
| **R8** | `agy 1.4.12` `plugin validate --strict` rejects `hooks/hooks.json`, Claude tool names in `hooks.json` matchers, `$PLUGIN_ROOT`/`$PLUGIN_DATA`/`$CLAUDE_PLUGIN_ROOT` in `hooks.json` commands, or `"dependencies"` in `.antigravity-plugin/plugin.json` | High | High (CI `agy plugin validate --strict` exits `1`) | Low | Move Claude hooks to `.claude-plugin/hooks.json` (`"hooks": "./.claude-plugin/hooks.json"`), list only Antigravity tool names in root `hooks.json` matchers, use `R="${PWD}"` in root `hooks.json`, and omit `"dependencies"` from `frank-exchange-of-views/.antigravity-plugin/plugin.json` (§II.1, §III.2). |

---

## V. Verification Plan

1. **Dual-Engine Plugin Validators (PR 1 & PR 2):**
   ```bash
   for p in plugins/*; do
     ~/.local/bin/claude plugin validate "$p"
     ~/.local/bin/agy plugin validate --strict "$p"
   done
   ```
   Confirm `0` errors and `0` warnings across all four plugins, `agents 2` on `prosthetic-conscience`, and `agents 11` on `frank-exchange-of-views`.
2. **Generators & Static Repository Gates (Every PR):**
   ```bash
   go -C scripts run ./agentgen -check
   go -C scripts run ./fetchbingen -check
   go -C scripts run ./buildidgen -check
   go -C scripts run ./hookfailgen -check
   go -C scripts run ./hookcoregen -check
   go -C scripts run ./check
   ```
   Confirm `git ls-files -s | awk '$1 == "120000"'` outputs `0` lines (zero symlinks).
3. **`hookcore` Unit & Wire-Descriptor Contract Tests (PR 3):**
   ```bash
   go -C scripts test -count=1 -v ./internal/hookcore/... ./hookcoregen/...
   ```
   Table-driven tests over captured Claude (`2.1.309`) and Antigravity (`1.4.12` `.exa.hooks_pb`) payloads covering `SessionStart`, `PreInvocation` (`terminationBehavior: "terminate"`, `metadata.user_markdown`), `PreTool` / `PreToolUse` (allow, deny, ask, `UpdatedCommand` → `overwrite.CommandLine`, classic + `oneharness` tool names and `Path` arg, malformed non-secret raw bytes, malformed secret-containing raw bytes), `PostTool` / `PostToolUse` (`Feedback` queue write on Antigravity vs `stderr` + `exit 2` on Claude, `PostToolHookArgs.result` field 4, `ToolError != ""` strike counting), and `PostInvocation` / `Stop`.
4. **Module Test Suites & Invariant Guards (PR 4 & PR 5):**
   ```bash
   go -C plugins/prosthetic-conscience/tools test -count=1 ./...
   go -C plugins/frank-exchange-of-views/tools test -count=1 ./...
   go -C plugins/gray-area/tools test -count=1 ./...
   FEOV_RELEASE_GATE=1 go -C plugins/frank-exchange-of-views/tools test -count=1 ./releasegate/fuzz/
   ```
   Specifically verify `TestTheShippedBinariesHonourTheIOContract`, `TestNoHookGuessesAProjectRootFromItsWorkingDirectory`, `TestNoHookReadsTheEnvironmentVariableWithoutResolvingThroughThisPackage`, and `TestHookBinaryLinksNothingExpensive`.
5. **Mutation Pass (`CLAUDE.md` Discipline):**
   Run `go -C scripts run ./mutate -module plugins/prosthetic-conscience/tools` (and over `scripts/internal/hookcore`): invert engine detection discriminator, delete `overwrite.CommandLine` assignment, delete `pending_<SessionID>.jsonl` unlink on `PreInvocation`, and verify each mutant is killed by a test.
6. **Driveable Check on Real Data (Both Live Clients — `claude 2.1.309` & `agy 1.4.12`):**
   With the built binaries installed into a scratch plugin directory:
   - **Antigravity (`~/.local/bin/agy -p`):**
     1. Run `agy -p "run the shell command echo 'ghp_1234567890abcdefghijklmnopqrstuvwxyzABCD'" --dangerously-skip-permissions` → verify `sc-pretooluse` denies `run_command` via `PreToolHookResult` and the command never executes.
     2. Run `agy -p "write unformatted Go code to scratch/bad.go and then run echo done" --dangerously-skip-permissions` → verify `sc-posttooluse` queues `qlty` feedback into `<PluginDataDir>/pending_<conversationId>.jsonl` on `PostTool` (exit `0`) and `sc-sessionstart -event PreInvocation` injects the feedback via `injectSteps` on the next turn, unlinking the queue file.
   - **Claude Code (`~/.local/bin/claude -p`):**
     1. Run `claude -p` with `--plugin-dir plugins/prosthetic-conscience` on the same secret probe and unformatted file probe → verify byte-identical denial on `PreToolUse` and `stderr` + `exit 2` feedback on `PostToolUse`.
7. **Auditor Gate ([#1467](https://github.com/ctoforaday/special-circumstances/issues/1467)):**
   Run `/prosthetic-conscience:plan-audit plans/dual-target-claude-antigravity.md` under Claude Code (`plan-auditor`) to `VERDICT: PASS`.

---

## Forks — rulings (gblock, 2026-10-18; updated 2026-11-01)

- **F1 — Manifest location:** `.antigravity-plugin/plugin.json` alongside `.claude-plugin/plugin.json` (no root `plugin.json` clutter; omit `"dependencies"` in `frank-exchange-of-views/.antigravity-plugin/plugin.json`): **RULED.**
- **F2 — Hook manifest coexistence (updated 2026-11-01 for `agy 1.4.12`):** Relocate Claude's hook manifest from `hooks/hooks.json` to `.claude-plugin/hooks.json` (declared via `"hooks": "./.claude-plugin/hooks.json"` in `.claude-plugin/plugin.json`) and place Antigravity's named-hook group manifest at `<plugin_root>/hooks.json` with zero symlinks and zero `agy plugin validate --strict` warnings; track future Agent Plugins `1.1.0+` §8.2 reverse-domain extension directories in [#1456](https://github.com/ctoforaday/special-circumstances/issues/1456): **RULED.**
- **F3 — Common wire-adapter library (`hookcore`):** Single-source `scripts/internal/hookcore` abstracting `env`, `stdin`, `stdout`, `stderr`, and `exitCode` (including `ANTIGRAVITY_CONVERSATION_ID`, `artifactDirectoryPath` fallback, `oneharness` tool names `read_file`/`write_file`/`edit_file` + `Path`, `PostToolHookArgs.result`, `PreInvocation` `terminationBehavior`, and `HookSystemMessage.metadata.user_markdown`), propagated byte-identically into all three plugin Go modules via `scripts/hookcoregen`: **RULED.**
- **F4 — Zero new binaries in `tools/cmd/`:** Handle Antigravity `PreInvocation` via `sc-sessionstart -event PreInvocation` and fold Antigravity `PostTool` (`error != ""`) strike counting into `sc-posttooluse`: **RULED.**

