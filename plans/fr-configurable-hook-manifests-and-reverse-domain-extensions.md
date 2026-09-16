# Feature Request: Configurable Hook Manifests and Agent Plugins 1.0.0 Reverse-Domain Support

> **Status:** 2026-09-16: Proposal — Upstream Feature Request & Compatibility Analysis  
> **Target Systems:** Google Antigravity (`agy` / Jetski engine) & Anthropic Claude Code (`claude`)  
> **Normative Standard:** [Agent Plugins Specification 1.0.0](https://github.com/agentplugins/agent-plugins-spec/blob/main/spec/1.0.0.md) (§5 Manifest & §8 Client Extensions)  
> **Discovered Friction:** Empirical dual-target validation in [`special-circumstances`](file:///home/gblock_ctoforaday_com/projects/special-circumstances)

---

## 1. Problem Statement & Operational Friction

When building portable, multi-harness agent plugins intended to run concurrently across both Anthropic Claude Code and Google Antigravity, developers encounter two symmetric implementation gaps:

1. **Antigravity Hardcodes `hooks.json` Path:**
   - In `inspect.go` (`scanHooks`) and `hooks_manager.go` of `agy` (v1.2.2), hook discovery is hardcoded:
     ```go
     filepath.Join(pluginDir, "hooks.json")
     ```
   - Antigravity ignores any `"hooks"` path declared in `plugin.json`.
   - If a plugin names its hooks file `hooks-jetski.json` or places it under `com.google/hooks.json`, `agy plugin validate` discovers 0 hooks.
   - **Operational Cost:** Authors must maintain an artificial filesystem symlink (`hooks.json -> hooks-jetski.json`) to distinguish Antigravity hooks from other harness configs on disk.

2. **Claude Code Rejects Root Manifest (`plugin.json`):**
   - Claude Code (`claude` v2.1.270) strictly mandates `.claude-plugin/plugin.json`.
   - Running `claude plugin validate .` on an Agent Plugins 1.0.0-compliant directory containing root `plugin.json` fails with Exit 1:
     ```
     Validation failed:
     Plugin manifest not found at .claude-plugin/plugin.json
     ```
   - **Operational Cost:** Authors must duplicate manifest metadata across two different directory locations (`.claude-plugin/plugin.json` and root `plugin.json`).

3. **Both Engines Lack Agent Plugins 1.0.0 Extension Directories (§8.2):**
   - The Agent Plugins 1.0.0 standard defines top-level reverse-domain extension folders:
     ```text
     my-plugin/
     ├── plugin.json
     ├── skills/
     ├── com.anthropic/
     └── com.google/
     ```
   - Neither engine scans its corresponding reverse-domain folder on disk.

---

## 2. Normative Specification Alignment

The [Agent Plugins Specification 1.0.0](https://github.com/agentplugins/agent-plugins-spec) provides a clear architectural blueprint for multi-client co-existence:

### Section 5.1: Manifest Location
> *"Clients MUST check for a manifest at `plugin.json` in the plugin root. The Agent Plugins core specification defines exactly one portable manifest per plugin. No other file can replace, supplement, or override the core fields in root `plugin.json`."*

### Section 8.1 & 8.2: Client Extensions
> *"Client-specific manifest data MUST be represented under a reverse-domain namespace in `extensions`. Client-specific files MUST be represented under a top-level directory named for that namespace. A client MAY use either representation or both."*

Current implementations lag behind this standard:
- Antigravity adopted the root `plugin.json` location (§5.1), but failed to adopt client extensions (§8) for its lifecycle hooks.
- Claude Code supports custom hook paths via `"hooks": "./path.json"`, but failed to adopt root `plugin.json` discovery (§5.1).

---

## 3. Detailed Upstream Feature Requests

### 3.1 Feature Request for Google Antigravity (`google-antigravity/antigravity-cli`)

#### Title: Support Configurable Hooks Path and Reverse-Domain Extensions in `plugin.json`

#### Proposed Changes:
1. **Configurable Hook Path:**
   Allow `plugin.json` to declare an explicit path to hooks:
   ```json
   {
     "name": "prosthetic-conscience",
     "version": "0.4.0",
     "hooks": "./hooks-jetski.json"
   }
   ```
2. **Reverse-Domain Extension Object (§8.1):**
   Support `extensions["com.google"].hooks`:
   ```json
   {
     "name": "prosthetic-conscience",
     "extensions": {
       "com.google": {
         "hooks": "./hooks-jetski.json"
       }
     }
   }
   ```
3. **Extension Directory Fallback (§8.2):**
   If `hooks.json` is absent at root, probe `com.google/hooks.json` or `com.google/hooks/hooks.json`.
4. **Resolution Priority:**
   1. `extensions["com.google"].hooks` (if declared in `plugin.json`)
   2. Root `"hooks"` key (if declared in `plugin.json`)
   3. `com.google/hooks.json` (if directory exists)
   4. Root `hooks.json` (legacy default fallback for 100% backward compatibility)

---

### 3.2 Feature Request for Anthropic Claude Code (`anthropics/claude-code`)

#### Title: Support Root `plugin.json` Discovery and Agent Plugins 1.0.0 Manifests

#### Proposed Changes:
1. **Root `plugin.json` Discovery:**
   Update plugin discovery to check for `plugin.json` at package root if `.claude-plugin/plugin.json` is not found.
2. **Reverse-Domain Extension Object (§8.1):**
   Read Anthropic-specific keys from `extensions["com.anthropic"]` in root `plugin.json`:
   ```json
   {
     "name": "prosthetic-conscience",
     "extensions": {
       "com.anthropic": {
         "hooks": "./hooks-claude.json"
       }
     }
   }
   ```
3. **Extension Directory Fallback (§8.2):**
   Recognize `com.anthropic/` directory as an alternative to `.claude-plugin/`.
4. **Resolution Priority:**
   1. `.claude-plugin/plugin.json` (legacy default for 100% backward compatibility)
   2. Root `plugin.json` (Agent Plugins 1.0.0 standard)

---

## 4. Interim Workaround: The Validated Symlink Bridge

Until upstream runtimes implement Agent Plugins 1.0.0 Section 8, the following layout provides 100% compliance, zero warnings, and concurrent compatibility:

```
plugins/prosthetic-conscience/
├── .claude-plugin/
│   └── plugin.json                 # Declares "hooks": "./hooks-claude.json"
├── plugin.json                     # Antigravity & Agent Plugins 1.0.0 root manifest
├── hooks-claude.json               # Claude-formatted lifecycle hooks
├── hooks-jetski.json               # Antigravity-formatted lifecycle hooks
└── hooks.json -> hooks-jetski.json # Relative symlink satisfying agy inspect.go
```

### Verification Matrix
- `claude plugin validate --strict .` $\rightarrow$ **PASS (Exit 0)**
- `agy plugin validate .` $\rightarrow$ **PASS (Exit 0)**
