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

## Common Commands

- `qlty check`: Unified linting/quality check across all supported languages.
- `qlty fmt`: Auto-format changed files.
- `qlty fmt --all`: Force auto-format across the entire codebase.
- `qlty plugins list`: Show available and enabled plugins.
- `qlty metrics`: Output code quality metrics (complexity, duplication, etc.).

## Supported Languages & Tools

- **Go**: `gofmt`, `golangci-lint`
- **Markdown/HTML/CSS/YAML**: `prettier`
- **Python**: `ruff`
- **Shell**: `shellcheck`, `shfmt`
- **SQL**: `sqlfluff`
