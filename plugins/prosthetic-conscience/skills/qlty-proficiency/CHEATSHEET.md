# Qlty CLI Cheatsheet & Operational Rules

The Qlty CLI is the unified interface for code quality (linting, formatting, metrics) in this repository.

## Operational Rules

1. **Initialization**: The repository must contain a `.qlty/qlty.toml` configuration file. Run `qlty init` only if this is missing.
2. **Verification (Linting)**: You MUST run `qlty check` during the **Verification** phase of any coding task. Resolve all reported issues before final submission.
3. **Formatting**: Formatting is handled automatically via post-edit hooks. You do not need to run `qlty fmt` manually.
4. **Metrics**: Use `qlty metrics` to assess code complexity and technical debt when planning refactors.

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
