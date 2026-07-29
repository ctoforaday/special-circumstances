# Qlty CLI Cheatsheet & Operational Rules

The Qlty CLI is the unified interface for code quality (linting, formatting, metrics) in this repository.

## Operational Rules

1. **Initialization**: The repository must contain a `.qlty/qlty.toml` configuration file. Run `qlty init` only if this is missing. Outside an initialized repository every qlty subcommand fails with a Rust backtrace — presence of the binary is not sufficient, presence of the config is the real precondition.
2. **Verification (Linting)**: You MUST run `qlty check` during the **Verification** phase of any coding task. Resolve all reported issues before final submission.
3. **Formatting**: Formatting is handled automatically via post-edit hooks. You do not need to run `qlty fmt` manually.
4. **Metrics**: Use `qlty metrics` to assess code complexity and technical debt when planning refactors.

## The post-edit hook (`sc-quality-gate`)

After every `Write`/`Edit`, the hook runs `qlty fmt --trigger agent <file>` then
`qlty check --no-fix <file>`, scoped to the one file that changed.

- Findings come back on stderr; the write is **not** reverted.
- When `qlty fmt` rewrites the file, the hook says so — **re-read the file before your next `Edit`**, or the edit fails on a string mismatch against your stale copy.
- Missing qlty, an uninitialized repository, a timeout, or a qlty crash never costs the tool call: the hook exits 0 and records the reason in `.claude/prosthetic-conscience-hook.log`.

| Environment variable | Effect |
|---|---|
| `SC_QUALITY_GATE=off` | Disable the hook entirely |
| `SC_QUALITY_GATE=check-only` | Lint but never rewrite files |
| `SC_QUALITY_GATE_TIMEOUT_SECONDS` | Budget per firing (default 30, max 600) |

## Timings (measured, qlty 0.639.0)

| Run | Cost |
|---|---|
| Cold — qlty installs its linters | ~29s |
| Warm, cached | ~1s |
| Warm, `--no-cache` | ~8s |

The cold run is why the hook has a timeout at all. Warm the cache with one
manual `qlty check` after cloning rather than paying it inside an edit.

> [!WARNING]
> **qlty caches issues hard, and a stale cache lies.** A sweep re-run after changing
> `.qlty/qlty.toml` and adding a plugin config returned a byte-identical count —
> 34,768 issues, same rule breakdown — when the true figure was 5,633. Treat any count
> from a cached run as unverified; re-run with `--no-cache` before believing a number
> or concluding that a config change had no effect.

## Choosing which plugins to enable

The post-edit hook runs `qlty check` **per file, on every write**, so a plugin's cost is
paid per edit rather than per CI run. Measure before enabling — a plugin returning dozens
of findings per touch is one that gets the gate switched off entirely.

Three traps found by measuring rather than reasoning:

- **`mode = "monitor"` does not quiet the command line.** A monitor-mode plugin still
  exits 1 and still prints its findings from `qlty check`; the mode suppresses the Qlty
  Cloud gate only. It is not a way to keep a noisy linter out of the hook's feedback.
- **A plugin cannot be scoped to a subset of file types.** Per-plugin `exclude_patterns`
  had no measurable effect (367 findings before and after), and a tool's own ignore file
  is largely bypassed because qlty passes explicit paths (367 → 365). Only the top-level
  `exclude_patterns` works, and it applies to every plugin at once.
- **A `[[source]]` block is mandatory.** Without it, every `name = "..."` fails the whole
  run with `Plugin definition not found` — the plugin list alone is inert.

`osv-scanner` reads the `go` directive as though it were a built toolchain, so the form
of the directive decides the verdict: `go 1.25` (open-ended) reports clean, while
`go 1.25.0` (an exact release) reports its accumulated CVEs. `go mod tidy` may force the
exact form, in which case the findings are unavoidable rather than negligent.

## Common Commands

- `qlty check`: Unified linting/quality check across all supported languages.
- `qlty fmt`: Auto-format changed files.
- `qlty fmt --all`: Force auto-format across the entire codebase.
- `qlty plugins list`: Show available and enabled plugins.
- `qlty metrics`: Output code quality metrics (complexity, duplication, etc.).

## Supported Languages & Tools

What qlty *can* run, which is not the same as what a given repository *should* enable:

- **Go**: `gofmt`, `golangci-lint`, `radarlint-go`
- **Markdown**: `markdownlint`, `prettier`, `vale`
- **HTML/CSS/YAML/JSON**: `prettier`, `yamllint`
- **JavaScript/TypeScript**: `eslint`, `biome`, `prettier`
- **Python**: `ruff`, `bandit`, `radarlint-python`
- **Shell**: `shellcheck`, `shfmt`
- **SQL**: `sqlfluff`
- **GitHub Actions**: `actionlint`, `zizmor`
- **Dependencies / secrets**: `osv-scanner`, `trufflehog`

`qlty init` autodetects by config-file presence, so it will not enable `markdownlint` or
`prettier` on a repository full of markdown that carries no config for them. Absence from
its output is not a judgment that the plugin does not apply.
