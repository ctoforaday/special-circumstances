---
description: Environment preflight — probe local tool dependencies, report READY / DEGRADED / BLOCKED. `/sc-doctor --fix` builds or fetches the hook binaries (with consent).
---

Run the Special Circumstances environment preflight. Model terse-communication: table + verdict, no filler.

1. Read `${CLAUDE_PLUGIN_ROOT}/requirements.json`. For each entry in `tools[]`, run its `check_cmd`; record present/absent (+ version if present).
2. Detect the platform: `go env GOOS GOARCH` if Go is present, else `uname -sm` / PowerShell `$env:OS`.
3. Check the hook binaries — for each directory under `${CLAUDE_PLUGIN_ROOT}/tools/cmd/` (sc-quality-gate, sc-secrets-gate, sc-toolchain-nudge, …), does `${CLAUDE_PLUGIN_ROOT}/bin/<name>` (`.exe` on Windows) exist?
4. Print a table — tool → ✓/✗ → version, or the `install[GOOS]` command if absent — then a verdict:
   - **READY**: every `required` tool present and every hook binary built (recommended tools may be missing).
   - **DEGRADED**: a `recommended` tool or any hook binary is missing (hooks capability-gate and degrade).
   - **BLOCKED**: a `required` tool is missing.

If the argument is `--fix`: installing mutates the machine, so YOU MUST get explicit confirmation first (Semantic Consent), then provision only what is missing:

- **Hook binaries** — the primary fix; for EACH missing binary `<name>` in `tools/cmd/`:
  - If Go is on PATH: `go build -C "${CLAUDE_PLUGIN_ROOT}/tools" -o "${CLAUDE_PLUGIN_ROOT}/bin/<name>" ./cmd/<name>` (append `.exe` on Windows).
  - Else: download the asset `<name>_{goos}_{goarch}{exe}` from the latest `ctoforaday/special-circumstances` GitHub Release (`gh release download` or the browser URL), verify its SHA256 against the release `SHA256SUMS`, and place it in `${CLAUDE_PLUGIN_ROOT}/bin/`.
- **External tools** (git / gh / qlty) — print the `install[GOOS]` command; run it only on a second explicit confirmation, one tool at a time.

AFTER fixing, re-probe and reprint the verdict. YOU MUST NOT auto-run `--fix` at session start.
