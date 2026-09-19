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
├── skills/                    # 27 skills (22 cognitive + 5 entry points, former commands)
│   └── */SKILL.md             # Added "disable-slash-command: true" for 22 cognitive skills; 5 entry points remain visible
└── tools/
    ├── internal/
    │   └── hookcore/          # Pure-Go normalized domain model (zero external dependencies)
    └── bin/                   # Compiled multi-harness binaries (sc-pretooluse, sc-posttooluse, etc.)
```

---

## 3. Toolchain Normalization (`hookcore`)

### 3.1 Design Principles & Invariants
1. **Zero External Dependencies:** Preserve the strict [`plugins/prosthetic-conscience/tools/go.mod`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/go.mod) invariant (pure Go 1.25 standard library; no Protobuf runtimes, reflection bloat, or external packages).
2. **Single Ingestion Pass:** A single decoder `hookcore.DecodeInvocation(r io.Reader) (*hookcore.Invocation, error)` that detects the payload structure by sniffing discriminator fields:
   - Claude Code: `tool_name`, `tool_input` (nested snake_case), `session_id`.
   - Antigravity: `stepIdx`, `toolCall.name`, `toolCall.args` (nested PascalCase), `conversationId`.
3. **Fail-Closed Security:** Unparseable payloads or unmapped outbound tools default to `Decision: Deny` in `sc-pretooluse`.
4. **Clean Exit Code Mapping:**
   - Claude Code: Exit 0 for allow/deny decisions (`permissionDecision: "deny"` in JSON); Exit 2 with stderr feedback for linter/policy failures in `PostToolUse`.
   - Antigravity: Exit 0 with JSON `{"decision": "deny", "deny_reason": "..."}` for `PreToolUse`; Exit 0 with `{}` for `PostToolUse` while caching findings to inject on `PreInvocation` via `{"injectSteps": [...]}`.

---

### 3.2 Complete Lifecycle Event & Execution Matrix

| Lifecycle Event | Anthropic Claude Code | Google Antigravity (`agy` / Jetski) | Normalization Strategy |
| :--- | :--- | :--- | :--- |
| **`SessionStart`** | Stdin: `session_id`, `transcript_path`, `cwd`. Output: exit 0, `additionalContext`. | Stdin: `conversationId`, `transcriptPath`, `workspacePaths` (array). Output: exit 0, `injectSteps`. | Ingest project root via `hookenv.ProjectDir`. Emit engine-specific context injection schema. |
| **`PreToolUse`** | Stdin: `tool_name`, `tool_input` (snake_case). Output: exit 0, `permissionDecision`. | Stdin: `stepIdx`, `toolCall: { name, args }` (PascalCase). Output: exit 0, `decision: "deny"`, `deny_reason`. | Canonicalize tool name & args into `hookcore.Invocation`. Output dual-target denial schema. |
| **`PostToolUse`** | Stdin: `tool_name`, `tool_input`, `tool_result`. Output: exit 2 (stderr feedback). | Stdin: `toolCall: { name, args }`, `error`. Output: exit 0 (stdout `{}`). | On Antigravity, write lint findings to `.findings_${conversationId}.json`; bridge to model turn via `PreInvocation`. |
| **`PreInvocation`** | *Unsupported / N/A*. | Stdin: `invocationNum`, `initialNumSteps`. Output: `injectSteps: [...]`. | Consumes `.findings_${conversationId}.json` cache and injects ephemeral quality-gate advisories. |
| **`PostInvocation`**| *Unsupported / N/A*. | Stdin: `invocationNum`, `fullyIdle`. Output: `{}`. | Resets per-turn telemetry and caches. |
| **`Stop`** | Stdin: `session_id`, `transcript_path`. Output: exit 0. | Stdin: `conversationId`, `transcriptPath`, `terminationReason`. Output: exit 0. | Route both to session archive and trajectory capture. |
| **`PostToolUseFailure`** | Stdin: `tool_name`, `tool_input`, `error`. | *Merged into `PostToolUse`* (passed as `error: string`). | `sc-strike-counter` reads `error != ""` on Antigravity's `PostToolUse`. |
| **Claude-Only Events** | `SubagentStop`, `PreCompact`, `PostCompact`, `FileChanged`, `SessionEnd`. | *Unsupported by engine*. | Claude hook wrapper runs natively; Antigravity manifest omits them. |

---

### 3.3 Tool Arguments Normalization (PascalCase ↔ snake_case)

Antigravity uses **PascalCase** inside `toolCall.args`, while Claude Code uses flat **snake_case** in `tool_input`:

| Tool Category | Claude Tool | Antigravity Tool | Claude Key (`tool_input`) | Antigravity Key (`toolCall.args`) | Canonical `Invocation` Field |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Shell Command** | `Bash` | `run_command` | `command` | `CommandLine` | `ToolInput.Command` |
| | | | `description` | `Description` / `toolAction` | `ToolInput.Description` |
| | | | *(ambient)* | `Cwd` | `ToolInput.Cwd` |
| **Write File** | `Write` | `write_to_file` | `file_path` / `path` | `TargetFile` | `ToolInput.FilePath` |
| | | | `content` | `CodeContent` | `ToolInput.Content` |
| **Edit File** | `Edit` | `replace_file_content` | `file_path` / `path` | `TargetFile` | `ToolInput.FilePath` |
| | | | `old_string` | `TargetContent` | `ToolInput.OldContent` |
| | | | `new_string` | `ReplacementContent` | `ToolInput.NewContent` |
| | | | `replace_all` | `AllowMultiple` | `ToolInput.AllowMultiple` |
| **Read File** | `Read` | `view_file` | `file_path` | `AbsolutePath` | `ToolInput.FilePath` |
| **Web Search** | `WebSearch` | `search_web` | `query` | `query` / `Query` | `ToolInput.Query` |
| **Web Fetch** | `WebFetch` | `read_url_content` | `url` | `Url` | `ToolInput.URL` |
| **Directory Search**| `LS` / `GlobTool`| `find_by_name` | `path` | `DirectoryPath` / `SearchDirectory` | `ToolInput.DirectoryPath` |
| **Grep Search** | `Grep` | `grep_search` | `path`, `pattern` | `SearchPath`, `Query` | `ToolInput.SearchPath`, `ToolInput.Query` |

---

### 3.4 Output Contracts & Exit Codes

| Action | Platform | Exit Code | Stdout Payload | Stderr Payload |
| :--- | :--- | :--- | :--- | :--- |
| **Allow** | Both | `0` | `{}` | Empty |
| **Deny (`PreToolUse`)** | Claude Code | `0` | `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"<msg>"},"systemMessage":"<msg>"}` | Empty |
| | Antigravity | `0` | `{"decision":"deny","deny_reason":"<msg>","reason":"<msg>"}` | Empty |
| **Feedback (`PostToolUse`)**| Claude Code | `2` | Empty | `<lint/policy message>` (injected directly to model context) |
| | Antigravity | `0` | `{}` (findings written to `.findings_${conversationId}.json`) | Empty |
| **Feedback (`PreInvocation`)**| Claude Code | *N/A* | *N/A* | *N/A* |
| | Antigravity | `0` | `{"injectSteps":[{"ephemeralMessage":"Quality Gate Notice:\n<feedback>"}]}` | Empty |
| **Context Injection (`SessionStart`)**| Claude Code | `0` | `{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"<msg>","watchPaths":[...]}}` | Empty |
| | Antigravity | `0` | `{"injectSteps":[{"ephemeralMessage":"<msg>"}]}` | Empty |

---

### 3.5 Special Circumstances Hook Audit & Root Defects Discovered

1. **`sc-pretooluse`**:
   - *Existing Defect*: `hookunit.NewCtx` read only `tool_name`. Under Antigravity, `c.ToolName` became `""`, causing `secretsgate.Unit().Applies` to evaluate to `false`. **Secrets checking was silently bypassed.**
   - *Fix*: `hookcore.DecodeInvocation` normalizes `toolCall.name` (`run_command` → `Bash`, `search_web` → `WebSearch`, `read_url_content` → `WebFetch`) and extracts `CommandLine` → `Command`.
2. **`sc-posttooluse`**:
   - *Existing Defect*: `qualitygate.Unit()` applies on `Write` or `Edit`. Antigravity's `write_to_file` and `replace_file_content` were ignored, and `TargetFile` was missing from `tool_input`. Gate returned `"no file in payload"`.
   - *Fix*: Normalize tools to `Write`/`Edit`, extract `TargetFile` to `FilePath`. Write findings cache on Antigravity; exit 0.
3. **`sc-sessionstart`**:
   - *Existing Defect*: `hookenv.Explain` expects `CLAUDE_PROJECT_DIR` and payload `cwd`. Under Antigravity, `CLAUDE_PROJECT_DIR` is unset and `cwd` is omitted (`workspacePaths` is provided). Hook exited with 0 output.
   - *Fix*: `hookenv.ProjectDir` inspects `workspacePaths[0]`. Outputs `injectSteps` for Antigravity.
4. **`sc-strike-counter`**:
   - *Existing Defect*: Relied exclusively on Claude's `PostToolUseFailure`.
   - *Fix*: On Antigravity, registers on `PostToolUse` and evaluates `error != ""`.
5. **`go.mod` Invariant Maintained**:
   - Package `tools/internal/hookcore` uses strictly pure Go 1.25 standard library (`encoding/json`, `io`, `strings`, `path/filepath`, `errors`, `fmt`).
   - [`plugins/prosthetic-conscience/tools/go.mod`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/go.mod) remains at 0 external dependencies.

---

### 3.6 Architecture of `tools/internal/hookcore`

```go
package hookcore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Harness string

const (
	HarnessClaudeCode  Harness = "claude"
	HarnessAntigravity Harness = "antigravity"
)

type CanonicalTool string

const (
	ToolBash      CanonicalTool = "Bash"
	ToolWrite     CanonicalTool = "Write"
	ToolEdit      CanonicalTool = "Edit"
	ToolRead      CanonicalTool = "Read"
	ToolWebSearch CanonicalTool = "WebSearch"
	ToolWebFetch  CanonicalTool = "WebFetch"
	ToolListDir   CanonicalTool = "ListDir"
	ToolGrep      CanonicalTool = "Grep"
	ToolOther     CanonicalTool = "Other"
)

type Invocation struct {
	Harness        Harness
	Event          string
	SessionID      string
	ProjectDir     string
	Tool           CanonicalTool
	RawToolName    string
	Input          ToolInput
	ToolResult     string
	Error          string
	RawPayload     []byte
}

type ToolInput struct {
	Command       string
	FilePath      string
	Content       string
	OldContent    string
	NewContent    string
	Query         string
	URL           string
	DirectoryPath string
	SearchPath    string
	AllowMultiple bool
}
```

---

## 4. Skills Compatibility Audit & Platform Parity

An exhaustive audit of all 37 skills across the repository (`prosthetic-conscience`, `gray-area`, and `frank-exchange-of-views`) was conducted to ensure cross-platform compatibility between Claude Code and Google Antigravity.

### 4.1 Skill Census & Visibility Invariants (37 Skills Total)

1. **Visible Operator Entry Points (12 Total)** — Invocable via `/` slash command palette; carry neither `disable-slash-command` nor `user-invocable`:
   - **10 Migrated Former Commands**:
     - `plugins/prosthetic-conscience`: `/checkpoint`, `/doctor`, `/plan-audit`, `/probe`, `/resume`
     - `plugins/gray-area`: `/audit-checkpoint`, `/audit-pr-body`, `/audit-repetition`, `/audit-seat-coverage`
     - `plugins/frank-exchange-of-views`: `/research`
   - **2 Operator Workflow Skills**:
     - `plugins/frank-exchange-of-views`: `/adversarial-audit`
     - `plugins/gray-area`: `/elicitation-testing`
2. **Hidden Cognitive Runbooks (25 Total)** — Hidden from interactive slash command palette via dual frontmatter flags (`disable-slash-command: true` for Antigravity, `user-invocable: false` for Claude Code):
   - `plugins/prosthetic-conscience` (22 procedural rules): `agent-guardrails`, `anti-spinning`, `complete-the-concept`, `context-checkpointing`, `context-efficiency`, `critical-stance`, `design-by-contract`, `facts-are-fields`, `git-proficiency`, `markdown-proficiency`, `pair-programming`, `plan-act-reflect`, `project-memory`, `qlty-proficiency`, `refactoring-safety`, `scratch-policy`, `semantic-consent`, `spec-driven-development`, `terse-communication`, `test-driven-development`, `think-around-problem`, `validation-loop`.
   - `plugins/frank-exchange-of-views` (1 internal protocol): `research-protocol`.
   - `plugins/gray-area` (2 internal runbooks): `restart-recovery`, `telepathy`.

### 4.2 Structural and Frontmatter Compliance
- **Schema & Formatting**: All 37 skills adhere to pure YAML frontmatter blocks delimited by `---`.
- **Quality Gates**: Every skill passes `scripts/frontmatter` validation (`name` matches parent directory name, descriptions within 1,536-character ceiling).
- **Internal Cross-References**: Wiki-links (e.g., `[[elicitation-testing]]`) resolve cleanly to existing skill directories; zero dead links to `commands/` exist.

### 4.3 Runtime Variable Expansion Variance (`${CLAUDE_PLUGIN_ROOT}`)
Six skills reference `${CLAUDE_PLUGIN_ROOT}`:
- `plugins/frank-exchange-of-views/skills/research/SKILL.md` (references `${CLAUDE_PLUGIN_ROOT}/skills/research-protocol/scripts/debate.js` and Claude's `Workflow` tool).
- `plugins/prosthetic-conscience/skills/doctor/SKILL.md` (references `${CLAUDE_PLUGIN_ROOT}/bin/sc-doctor`).
- `plugins/gray-area/skills/{audit-checkpoint, audit-pr-body, audit-repetition, audit-seat-coverage}/SKILL.md` (references `${CLAUDE_PLUGIN_ROOT}/bin/gray-area`).

**Platform Behavior Differences**:
- **Claude Code**: Interpolates `${CLAUDE_PLUGIN_ROOT}` with the installed plugin's filesystem path before prompt injection into model context.
- **Google Antigravity**: Does not perform `${CLAUDE_PLUGIN_ROOT}` substitution in skill bodies.

**Dual-Target Resolution Strategy**:
- **Compiled Binaries**: `bin/sc-doctor` and `bin/gray-area` are built into the plugin directory. In Antigravity environments, skills and rules invoke binaries via repository-relative paths (`./bin/...`), user PATH (when installed via `/doctor`), or plugin path fallback (`~/.gemini/config/plugins/.../bin`).
- **FEOV Swarm Orchestration**: Claude Code invokes `debate.js` using Claude's proprietary `Workflow` tool and `agent()` JS API. Antigravity does not support the `Workflow` tool; instead, it executes the multi-agent research debate through its native subagent runtime (`invoke_subagent` launching `lead`, `researcher`, `auditor`) or standalone CLI execution.

---

## 5. Implementation Checklist

- [x] **Phase 1: Skill Slash-Command Cleanliness & Visibility Parity** (Landed in #1006 & #1019)
  - All commands migrated to skills (`plugins/*/skills/*/SKILL.md`) ensuring unified platform parity.
  - All 25 non-user-facing cognitive skills annotated with `disable-slash-command: true` and `user-invocable: false`.
  - All 12 operator entry points and workflow skills (the 10 former commands: 5 in `prosthetic-conscience`, 4 in `gray-area`, 1 in `frank-exchange-of-views`; plus `adversarial-audit` and `elicitation-testing`) intentionally kept visible with neither hiding flag. Everything that was a command is intended to be visible.
- [ ] **Phase 2: Hook Config Separation & Symlink**
  - Extract Claude hook definitions to `hooks-claude.json`.
  - Update `.claude-plugin/plugin.json` to reference `"hooks": "./hooks-claude.json"`.
  - Create `hooks-jetski.json` with Antigravity top-level schema.
  - Symlink `hooks.json` -> `hooks-jetski.json`.
- [ ] **Phase 3: `hookcore` Ingestion Engine**
  - Implement `tools/internal/hookcore` package in Go.
  - Update `sc-pretooluse` and `sc-posttooluse` to consume `hookcore.Invocation`.
  - Update `sc-sessionstart` and `sc-strike-counter` for dual-harness payload handling.
  - Verify with unit tests across both mock Claude and Antigravity payloads.
- [ ] **Phase 4: Multi-Harness Validation**
  - Run `scripts/validatejson` and `scripts/pluginparity`.
  - Run `claude plugin validate --strict plugins/prosthetic-conscience`.
  - Run `agy plugin validate plugins/prosthetic-conscience`.
