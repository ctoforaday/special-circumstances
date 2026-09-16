# Special Circumstances: Dual-Target Claude Code & Google Antigravity Architecture Plan

> **Status:** Empirically Verified Implementation Plan  
> **Target Repository:** `/home/gblock_ctoforaday_com/projects/special-circumstances`  
> **Engines Tested:** Anthropic Claude Code (`claude` v2.1.270) & Google Antigravity (`agy` v1.2.2)  
> **Related Plans:** [`plans/antigravity-hooks-empirical-proof.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plans/antigravity-hooks-empirical-proof.md), [`plans/hide-non-user-facing-skills-slash-commands.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plans/hide-non-user-facing-skills-slash-commands.md)

---

## 1. Empirical Answers to Core Architectural Questions

### Q1: Can `agy` be configured with an arbitrary hook filename (e.g. `hooks-jetski.json`)?
**NO.**
- **Decompiled Evidence:** In `inspect.go` (`scanHooks`) and `hooks_manager.go` of `agy` (v1.2.2), the runtime resolves hooks via:
  ```go
  filepath.Join(pluginDir, "hooks.json")
  ```
- **Runtime Behavior:** `agy` does not read a `"hooks"` key from `plugin.json`. When `hooks-jetski.json` was supplied without `hooks.json`, `agy plugin validate` reported `0 hooks registered`.
- **Verified Solution:** Create `hooks.json` as a relative symbolic link pointing to `hooks-jetski.json`:
  ```bash
  ln -s hooks-jetski.json hooks.json
  ```
  `agy plugin validate` reads through the symlink and registers the hooks cleanly with Exit 0.

---

### Q2: Can Claude Code read `plugin.json` from the plugin root?
**NO.**
- **Validation Failure:** When executing `claude plugin validate .` on a directory containing only root `plugin.json`:
  ```
  Validation failed:
  Plugin manifest not found at .claude-plugin/plugin.json
  ```
  Claude Code strictly enforces `.claude-plugin/plugin.json` for plugin discovery (and `.claude-plugin/marketplace.json` for marketplace repos). It ignores root `plugin.json`.
- **Configurability:** Inside `.claude-plugin/plugin.json`, Claude Code *does* support custom hook file paths:
  ```json
  {
    "name": "prosthetic-conscience",
    "version": "0.4.0",
    "hooks": "./hooks-claude.json"
  }
  ```
  When tested against `claude plugin validate --strict`, this passes with Exit 0 and zero warnings.

---

### Q3: How do the two engines coexist without cross-contamination?
Because neither engine scans the other's manifest location:
1. **Claude Code** reads `.claude-plugin/plugin.json` -> loads `./hooks-claude.json` -> ignores root `plugin.json` and `hooks.json`.
2. **Antigravity** reads root `plugin.json` -> loads `hooks.json` (symlinked to `hooks-jetski.json`) -> ignores `.claude-plugin/` and `hooks-claude.json`.
3. **Simultaneous Validation:**
   ```bash
   agy plugin validate . && claude plugin validate --strict .
   ```
   Both validators exit 0 simultaneously on the exact same directory.

---

### Q4: Does Antigravity / Jetski support an equivalent dot-directory (e.g. `.jetski-plugin`)?
**NO.**
- **Empirical Test:** Tested `.jetski-plugin/plugin.json`, `.antigravity-plugin/plugin.json`, `.gemini-plugin/plugin.json`, `.agy-plugin/plugin.json`, and `.agents/plugin.json` with `agy plugin validate`. All failed identically:
  ```
  Error: missing plugin.json: stat <pluginDir>/plugin.json: no such file or directory
  ```
- **Engine Logic:** `agy` strictly searches for `plugin.json` at the root of the specified plugin directory. There is no dot-prefixed wrapper directory recognized by the binary.
- **Import Command:** The internal `ClaudeCodeImporter` and `GeminiCLIImporter` in `agy` are only called during one-time migration (`agy plugin import`), which copies external extensions into `~/.gemini/config/plugins/`; they are not used for real-time dynamic plugin loading.

---

### Q5: What is the upstream status on GitHub (Anthropic, Antigravity, and Agent Plugins Spec)?
- **Agent Plugins Specification 1.0.0 (`agentplugins/agent-plugins-spec`):**
  - Ratified in August 2026.
  - Section 5.1 mandates root `plugin.json` as the sole portable manifest.
  - Section 8.2 defines reverse-domain extension folders (`com.example.client/`) and manifest objects (`extensions.com.example.client`).
  - However, §8 explicitly notes: *"Agent Plugins assigns no portable discovery, validation, loading, or failure semantics to client extension data or files. Each client defines the contents and behavior of its own namespace."*
- **Anthropic Claude Code (`anthropics/claude-code`):**
  - Implementation predates the August 2026 Agent Plugins 1.0.0 specification.
  - GitHub issues track community friction regarding monorepo subdirectories (e.g. `awslabs/agent-plugins`) and strict manifest validation errors when encountering unknown fields or missing `.claude-plugin/plugin.json`.
  - Claude Code has not yet implemented root `plugin.json` discovery or `com.anthropic/` directory scanning.
- **Google Antigravity (`google-antigravity/antigravity-cli`):**
  - Adopted root `plugin.json` from the open standard, but hooks remain hardcoded to `hooks.json`.
  - Upstream issues focus on fixing hook working directories, WSL2/Windows pathing, and namespacing plugin-bundled MCP servers (`<plugin>_<server>`).
  - No issues or PRs currently implement `com.google/hooks.json` or `extensions` parsing in `agy`.
- **Conclusion:** While the open standard's architecture is clearly articulated in `spec/1.0.0.md`, neither CLI has completed the implementation of Section 8 extension directories. The symlink architecture (`hooks.json -> hooks-jetski.json` and `.claude-plugin/plugin.json` pointing to `hooks-claude.json`) is the necessary and fully functional bridge for current production CLIs.

---

## 2. Directory Layout for Dual Targeting

```
plugins/prosthetic-conscience/
├── .claude-plugin/
│   └── plugin.json            # Claude manifest ("hooks": "./hooks-claude.json")
├── plugin.json                # Antigravity & Agent Plugins 1.0.0 manifest
├── hooks-claude.json          # Claude lifecycle hooks (Claude schema: "hooks": { ... })
├── hooks-jetski.json          # Antigravity lifecycle hooks (Antigravity top-level schema)
├── hooks.json -> hooks-jetski.json # Symlink satisfying agy's hardcoded filename
├── rules/
│   └── AGENTS.md              # Always-on behavioral guardrails for Antigravity
├── skills/                    # 22 shared skills (identical markdown)
│   └── */SKILL.md             # Added "disable-slash-command: true" for internal skills
└── tools/
    ├── internal/
    │   └── hookcore/          # Pure-Go normalized domain model (zero external dependencies)
    └── bin/                   # Compiled multi-harness binaries (sc-pretooluse, sc-posttooluse, etc.)
```

---

## 3. Toolchain Normalization (`hookcore`)

### Design Principles
1. **Zero External Dependencies:** Preserve `prosthetic-conscience/tools/go.mod` invariant (standard library only; no Protobuf runtimes or heavy code generators).
2. **Single Ingestion Pass:** Single function `hookcore.DecodeInvocation(r io.Reader) (*hookcore.Invocation, error)` that detects payload shape:
   - Claude Code: `tool_name`, `tool_input` (nested snake_case).
   - Antigravity: `toolCall.name`, `toolCall.args` (nested camelCase).
3. **Fail-Closed Security:** Unparseable payloads or unmapped tools default to `Decision: Deny` in `sc-pretooluse`.
4. **Clean Exit Code Mapping:**
   - Claude Code: Exit 2 with stderr feedback for denies/rejections.
   - Antigravity: Exit 0 with JSON `{"decision": "deny", "deny_reason": "..."}` or `{"injectSteps": [...]}`.

---

## 4. Implementation Checklist

- [ ] **Phase 1: Skill Slash-Command Cleanliness**
  - Add `disable-slash-command: true` to frontmatter of the 22 non-user-facing skills.
  - Retain slash commands only for user actions (`/doctor`, `/checkpoint`, `/resume`).
- [ ] **Phase 2: Hook Config Separation & Symlink**
  - Extract Claude hook definitions to `hooks-claude.json`.
  - Update `.claude-plugin/plugin.json` to reference `"hooks": "./hooks-claude.json"`.
  - Create `hooks-jetski.json` with Antigravity top-level schema.
  - Symlink `hooks.json` -> `hooks-jetski.json`.
- [ ] **Phase 3: `hookcore` Ingestion Engine**
  - Implement `tools/internal/hookcore` package in Go.
  - Update `sc-pretooluse` and `sc-posttooluse` to consume `hookcore.Invocation`.
  - Verify with unit tests across both mock Claude and Antigravity payloads.
- [ ] **Phase 4: Multi-Harness Validation**
  - Run `scripts/validatejson` and `scripts/pluginparity`.
  - Run `claude plugin validate --strict plugins/prosthetic-conscience`.
  - Run `agy plugin validate plugins/prosthetic-conscience`.
