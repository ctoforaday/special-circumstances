# Cross-Platform sc-doctor: Dual-Harness Support (Claude Code & Antigravity)

**Status:** Draft  
**Author:** Antigravity / Gregory Block  
**Date:** 2026-11-07  
**Rule-Class:** `protocol-reform`  
**Sibling-Sweep:** none  

---

## I. Summary & Goals

### 1. Objective
Enable `sc-doctor` (the environment preflight and binary self-healing tool in `plugins/prosthetic-conscience`) to operate natively and cross-platform across both host harnesses: **Claude Code** and **Google Antigravity (`agy`)**, as well as local development checkouts.

### 2. The Problem
When running `/doctor` under Google Antigravity, `sc-doctor` suffers from four critical harness-isolation defects:
1. **Root Resolution Failure:** `resolveRoot()` only checks the `-root` CLI flag, `CLAUDE_PLUGIN_ROOT`, and executable grandparent. In Antigravity sessions where `CLAUDE_PLUGIN_ROOT` is unset, running `sc-doctor` from outside the binary directory (or via an alias) lacks harness environment awareness (`PLUGIN_ROOT`, `ANTIGRAVITY_APP_DATA_DIR`, or `~/.gemini/config/plugins/`).
2. **Sibling Plugin Invisibility:** Sibling plugin discovery (`siblingBinaries`, `siblingRequirements`, `danceWarnings`) hardcodes Claude Code's two-level marketplace cache layout (`<cacheDir>/<marketplace>/<plugin>/<version>/`). Antigravity installs plugins in a flat layout (`<configRoot>/plugins/<plugin>/` with manifests directly at root and no version subdirectories). Consequently, `sc-doctor` discovers zero sibling plugins under Antigravity: `feov-record`, `gray-area`, `telepathy`, and their required external toolchains (`node`, `jq`, `ripgrep`) are completely invisible. `sc-doctor` prints a false `READY` verdict even when FEOV or Gray Area dependencies are completely missing or unbuilt.
3. **Misidentified Plugin Identity & Staleness Blindness:** `hookBinaries()` derives plugin name and version via `filepath.Base(filepath.Dir(root))` and `filepath.Base(root)`. In Antigravity (and repo checkouts), this incorrectly parses `Plugin = "plugins"` and `Version = "prosthetic-conscience"`. Furthermore, `ReferenceCommit()` only searches for Claude Code's `installed_plugins.json` matching `installPath == root`. In Antigravity, no `installed_plugins.json` exists in `.gemini`, causing `sc-doctor` to report `(11 stamped binaries have no reference commit to check against — staleness NOT MEASURED for them)`.
4. **Harness-Confused Thinking Warnings:** `captureWarnings()` assumes Claude Code's `~/.claude/settings.json` and warns about `showThinkingSummaries: false`. Antigravity records reasoning natively in its trajectory logs (`transcript.jsonl` / `transcript_full.jsonl`) without Claude's display setting.

### 3. Quantitative Success Criteria
1. **100% Sibling Plugin Discovery in Both Layouts:** In both Claude Code versioned cache and Antigravity flat plugin layout fixtures, `sc-doctor` discovers all Special Circumstances plugins (`prosthetic-conscience`, `frank-exchange-of-views`, `gray-area`, `sleeper-service`).
2. **Strictest Requirement Merging Across Harnesses:** Missing required tools in sibling plugins (e.g., `node` for FEOV, `git` for Gray Area) produce `BLOCKED` in both Claude Code and Antigravity.
3. **Multi-Plugin Provisioning via `--fix`:** `sc-doctor -fix` provisions missing hook binaries across all discovered plugins in both directory structures.
4. **Reference Commit Resolution:** Matches commit references via Git HEAD in checkouts, `installed_plugins.json` (by path or plugin+version) when available, and `.fetch/installed` / `.fetch/commit` stamps.
5. **Real CLI Test Coverage:** Real CLI subprocess execution tests (`exec.Command`) testing both Claude Code and Antigravity layouts with zero synthetic mocks for CLI invocation.
6. **Zero Performance or Build Regressions:** Full test suite passes in `< 30s`; all existing `sc-doctor` unit tests remain green.

### 4. Value Case
A preflight tool that reports `READY` while sibling tools are uninstalled defeats the foundational tenet of Special Circumstances: **adversarial verification over deference**. Under Antigravity, seats were crashing mid-round when attempting to invoke `feov-record` or `telepathy` because `sc-doctor` silently omitted them. A unified, cross-platform preflight eliminates harness-specific blind spots, ensures consistent verification, and costs negligible additional runtime.

---

## II. Technical Context

- **Language / Version:** Go 1.27 (`GOTOOLCHAIN=local`).
- **Target Harnesses:**
  - Claude Code: Nested directory structure (`<cacheRoot>/<marketplace>/<plugin>/<version>/`). Uses `CLAUDE_PLUGIN_ROOT`, `installed_plugins.json`, and `.claude/settings.json`.
  - Google Antigravity: Flat directory structure (`<configRoot>/plugins/<plugin>/`). Uses `ANTIGRAVITY_AGENT=1`, `ANTIGRAVITY_APP_DATA_DIR`, `PLUGIN_ROOT`, and native trajectory thinking logs.
  - Development Checkout: Flat repository layout (`<repoRoot>/plugins/<plugin>/`). Uses Git HEAD.
- **Dependencies:** Standard library (`os`, `path/filepath`, `encoding/json`, `os/exec`, `debug/buildinfo`) + internal `shared/buildid` and `internal/toolchain`. No external network dependencies for preflight read-only checks.
- **Constraints:**
  - Read-only contract: Plain `sc-doctor` run MUST NOT mutate the filesystem or make network requests.
  - No backward compatibility shims or transition archaeology (`AGENTS.md`).
  - Strict Design-by-Contract: Deterministic outputs, explicit error reporting.

---

## III. Proposed Changes (the spec)

```
plugins/prosthetic-conscience/
├── skills/
│   └── doctor/
│       └── SKILL.md                          [MODIFY] Support dual-harness execution & bootstrap
└── tools/
    └── internal/
        └── doctor/
            ├── main.go                       [MODIFY] Cross-platform root & sibling discovery
            ├── reference.go                  [MODIFY] Cross-harness reference commit resolution
            ├── capture.go                    [MODIFY] Harness-aware reasoning capture check
            ├── main_test.go                  [MODIFY] Add flat Antigravity sibling test
            ├── reference_test.go             [MODIFY] Add cross-manifest & commit stamp tests
            ├── cli_test.go                   [NEW] Real CLI subprocess integration tests for both harnesses
            └── run_test.go                   [MODIFY] Test dual-harness plugin identity & roots
```

### 1. `main.go` Refactoring & Abstractions

#### A. Unified Plugin Identity Resolution
Replace fragile path basename guessing with `pluginIdentity(root string) (name, version string)`:
1. Inspect manifests: `plugin.json`, `.claude-plugin/plugin.json`, `.antigravity-plugin/plugin.json`. If found and non-empty, extract `Name` and `Version`.
2. Inspect `.fetch/installed`: If present, parse version from tag (e.g., `prosthetic-conscience--v0.52.0` -> `0.52.0`).
3. Path heuristics fallback:
   - If `filepath.Base(root)` is semver (`isSemver(base)`: leading digit + numeric dotted pattern), set `name = filepath.Base(filepath.Dir(root))` and `version = base` (Claude Code versioned cache).
   - Otherwise, set `name = filepath.Base(root)` and `version = ""` (flat layout fallback).

#### B. Sibling Plugin Discovery (`siblingPluginRoots`)
Create a single source of truth for sibling discovery:
```go
type pluginRef struct {
    Name    string
    Version string
    Root    string
}

func siblingPluginRoots(root string) []pluginRef
```
1. Determine `selfName, _ := pluginIdentity(root)`.
2. Determine layout structure by inspecting `root`:
   - Check if `isSemver(filepath.Base(root))`:
     - If true, `root` is a version directory within a versioned marketplace cache (`<marketDir>/<plugin>/<version>`):
       - The container of plugins is `marketDir := filepath.Dir(filepath.Dir(root))`.
       - For each sibling entry `e` in `marketDir` (skipping `selfName` and non-directories):
         Find `v := newestVersionDir(filepath.Join(marketDir, e.Name()))`. If `v != ""`:
         `siblingDir := filepath.Join(marketDir, e.Name(), v)`
         Resolve `name, version := pluginIdentity(siblingDir)` (defaulting version to `v`).
         Add `pluginRef{Name: name, Version: version, Root: siblingDir}`.
   - If `isSemver(filepath.Base(root))` is false, `root` is directly a plugin directory in a flat layout (`<pluginsDir>/<plugin>`, e.g., Antigravity `~/.gemini/config/plugins/<plugin>` or dev repo `plugins/<plugin>`):
     - The container of sibling plugins is `parent := filepath.Dir(root)`.
     - For every child directory `e` in `parent` (skipping `selfName` and non-directories):
       `siblingDir := filepath.Join(parent, e.Name())`
       Verify `siblingDir` is a plugin directory (contains `requirements.json`, `tools/cmd`, or a plugin manifest `plugin.json` / `.claude-plugin/plugin.json` / `.antigravity-plugin/plugin.json`). If not, skip (e.g. non-SC plugins or scratch dirs).
       Resolve `name, version := pluginIdentity(siblingDir)`.
       Add `pluginRef{Name: name, Version: version, Root: siblingDir}`.
3. Refactor `siblingBinaries(root)` and `siblingRequirements(root)` to call `siblingPluginRoots(root)` directly.

#### C. Guarded `danceWarnings`
Ensure `danceWarnings` only runs when `root` is in a versioned cache layout (i.e. `isSemver(filepath.Base(root))` is true). In flat layouts (Antigravity and dev checkouts), `danceWarnings` returns `nil`, avoiding false checks against sibling plugin names.

#### D. Environment Root Resolution
Update `resolveRoot` and `Main()`:
- `Main()` checks `CLAUDE_PLUGIN_ROOT`, then `PLUGIN_ROOT`, passing the resolved env root to `run()`.
- `resolveRoot()` checks:
  1. `rootFlag` (`-root`)
  2. `envRoot` (`CLAUDE_PLUGIN_ROOT` or `PLUGIN_ROOT`)
  3. `executable()` grandparent
  4. Antigravity fallback: if running under Antigravity (`ANTIGRAVITY_AGENT=1` or `ANTIGRAVITY_APP_DATA_DIR` set), check `~/.gemini/config/plugins/prosthetic-conscience`.

### 2. `reference.go` Reference Commit Generalization
Update `ReferenceCommit(b binStatus) string`:
1. Check `HeadCommit(b.Root)` (returns Git HEAD if inside repository).
2. Check `.fetch/commit` in `b.Root` if present (holds the VCS commit SHA).
3. Check `installedCommit(b.Root)` (Claude Code exact path match in `installed_plugins.json`).
4. Check `installedCommitByName(b.Plugin, b.Version)`: walks upward for `installed_plugins.json` or checks `~/.claude/plugins/installed_plugins.json` to find matching plugin name and version, returning `GitCommitSha`.
5. Update `stampInstalled(bins []binStatus)`: when installing or verifying binaries, stamp `.fetch/commit` with `ReadBuildStamp(binPath(b)).Revision` so future checks on any platform immediately find the reference commit. Note that `.fetch/installed` contains the release tag string (e.g. `prosthetic-conscience--v0.52.0`) used for tag matching by `fetch-bin.sh`, while `.fetch/commit` contains the raw 40-character commit SHA used by `ReferenceCommit` and `Compare`.

### 3. `capture.go` Harness Awareness
Update `boxWarnings()`:
- Check if running under Antigravity (`os.Getenv("ANTIGRAVITY_AGENT") != ""` or `os.Getenv("ANTIGRAVITY_APP_DATA_DIR") != ""`).
- If running under Antigravity, skip Claude Code's `showThinkingSummaries` warning, as Antigravity captures reasoning natively into `transcript.jsonl`.
- If running under Claude Code, retain existing `captureWarnings` behavior.

### 4. `SKILL.md` Dual-Harness Invocation & Bootstrap
Update `plugins/prosthetic-conscience/skills/doctor/SKILL.md`:
- Preserve the exact literal token `${CLAUDE_PLUGIN_ROOT}` unnested so Claude Code's prompt injector regex `/\$\{CLAUDE_PLUGIN_ROOT\}/g` reliably matches and interpolates it, while providing fallback to `${PLUGIN_ROOT}` and PATH via shell chaining:
  ```bash
  B="${CLAUDE_PLUGIN_ROOT}/bin/sc-doctor"; [ -x "$B" ] || B="${PLUGIN_ROOT}/bin/sc-doctor"; [ -x "$B" ] || B="sc-doctor"; "$B"
  ```
- For bootstrap: chain directory variables:
  ```bash
  D="${CLAUDE_PLUGIN_ROOT}"; [ -n "$D" ] || D="${PLUGIN_ROOT}"; [ -n "$D" ] || D="${HOME}/.gemini/config/plugins/prosthetic-conscience"
  ```
  and read version from `plugin.json` or `.claude-plugin/plugin.json` or `.antigravity-plugin/plugin.json`.

---

### Consumer Census

Exact search executed:
`git grep -n "resolveRoot\|hookBinaries\|siblingBinaries\|siblingRequirements\|ReferenceCommit\|installedCommit" plugins/prosthetic-conscience/tools/internal/doctor/`

| Symbol | Location | Status |
|---|---|---|
| `resolveRoot` | `internal/doctor/main.go:614`, `internal/doctor/main.go:644`, `internal/doctor/run_test.go:244` | Signature preserved (`rootFlag, envRoot string, executable func() (string, error)`). Internal logic enhanced. |
| `hookBinaries` | `internal/doctor/main.go:94`, `internal/doctor/main.go:663`, `internal/doctor/release_test.go:99`, `internal/doctor/run_test.go:278` | Signature preserved (`hookBinaries(root string) []binStatus`). Uses `pluginIdentity` instead of path basename slicing. |
| `siblingBinaries` | `internal/doctor/main.go:102`, `internal/doctor/main.go:663`, `internal/doctor/main_test.go:101,120`, `internal/doctor/release_test.go:118`, `internal/doctor/run_test.go:311` | Signature preserved (`siblingBinaries(root string) []binStatus`). Delegates to `siblingPluginRoots`. |
| `siblingRequirements` | `internal/doctor/main.go:351`, `internal/doctor/main.go:685`, `internal/doctor/dance_test.go:90`, `internal/doctor/provision_test.go:95`, `internal/doctor/run_test.go:314` | Signature preserved (`siblingRequirements(root string) []pluginReq`). Delegates to `siblingPluginRoots`. |
| `ReferenceCommit` | `internal/doctor/reference.go:24`, `internal/doctor/main.go:187,212,264`, `internal/doctor/reference.go:84`, `internal/doctor/reference_test.go:46,69,81,98` | Signature preserved (`ReferenceCommit(b binStatus) string`). Enhanced with name/version manifest lookup and `.fetch/commit`. |
| `installedCommit` | `internal/doctor/reference.go:45`, `internal/doctor/reference.go:28` | Preserved for exact path match; helper `installedCommitByName` added alongside. |

---

## IV. Risk & Mitigation

| Risk | Likelihood | Impact | Complexity | Mitigation |
|---|---|---|---|---|
| **R1: False positive sibling detection** in flat directory containing non-plugin directories | Medium | Medium | Low | `siblingPluginRoots` verifies directories contain `requirements.json`, `tools/cmd`, or a plugin manifest before treating them as plugins. Non-plugin directories are skipped. |
| **R2: `danceWarnings` false alarm** on plugin directory names in flat layout | High | High | Low | Explicitly guard `danceWarnings` with `isSemver(filepath.Base(root))`. In flat layout (where base is plugin name), `danceWarnings` returns `nil`. |
| **R3: Breaking existing Claude Code cache tests** | Medium | High | Low | All existing tests in `run_test.go`, `main_test.go`, `dance_test.go`, and `provision_test.go` use the nested versioned directory fixture. The unified discovery explicitly checks flat vs versioned, keeping all existing fixtures green. |
| **R4: Subprocess CLI test environment flakiness** | Low | Medium | Medium | Compile test binary once in `TestMain` or helper, execute with explicit isolated `t.TempDir()` environments for Claude and Antigravity fixtures. |

---

## V. Verification Plan (the checklist)

### 1. Automated Unit & Contract Tests
Run from `plugins/prosthetic-conscience/tools`:
```bash
export PATH="/usr/local/go/bin:/home/gblock_ctoforaday_com/.local/bin:$HOME/go/bin:$PATH"
unset $(env | grep -o '^FEOV_[A-Z_]*')
go test -v -count=1 ./internal/doctor/...
```

### 2. Real CLI Integration Tests (`cli_test.go`)
Automated tests executing the real compiled `sc-doctor` binary via `exec.Command`:
1. **Claude Code Layout Real CLI Execution:**
   - Execute built `sc-doctor` with `CLAUDE_PLUGIN_ROOT` pointing to a versioned fixture (`.../cache/special-circumstances/prosthetic-conscience/0.52.0`).
   - Sibling `gray-area/0.18.0` with `requirements.json`.
   - Assert exit code 0, table contains both plugins and merged obligations.
2. **Antigravity Layout Real CLI Execution:**
   - Execute built `sc-doctor` with `ANTIGRAVITY_AGENT=1` and `PLUGIN_ROOT` (and with `CLAUDE_PLUGIN_ROOT` unset) pointing to a flat fixture (`.../config/plugins/prosthetic-conscience`).
   - Sibling flat directories `gray-area` and `frank-exchange-of-views`.
   - Assert exit code 0, table contains `-- frank-exchange-of-views --` and `-- gray-area --`, verifies tools from all three plugins.
   - Also test invoking `sc-doctor` with no `-root` flag from an arbitrary outside working directory (verifying executable grandparent and Antigravity fallback discovery).
3. **Dual-Harness `--fix` Real CLI Execution:**
   - Test `sc-doctor -fix` in both fixtures. Verify it detects unbuilt binaries across siblings in both layouts.

### 3. Driveable Check on Real Data
Run the built `sc-doctor` on the host machine against the live Antigravity plugin installation:
```bash
export PATH="/usr/local/go/bin:/home/gblock_ctoforaday_com/.local/bin:$HOME/go/bin:$PATH"
unset $(env | grep -o '^FEOV_[A-Z_]*')
go build -C plugins/prosthetic-conscience/tools -o ~/.gemini/config/plugins/prosthetic-conscience/bin/sc-doctor ./cmd/sc-doctor
~/.gemini/config/plugins/prosthetic-conscience/bin/sc-doctor
```
**Expected Real Output:**
- Table lists `prosthetic-conscience` tools and built binaries.
- Table lists `-- frank-exchange-of-views --` and probes `claude`, `node`, `uv`, `jq`.
- Table lists `-- gray-area --` and probes `ripgrep`, `git`.
- Table lists binaries for `feov-record`, `gray-area`, `telepathy`.
- Final verdict incorporates all three plugins.

### 4. Repo CI Gates
```bash
go -C scripts run ./check -only golden,vet,fmt,frontmatter,pluginparity
```
