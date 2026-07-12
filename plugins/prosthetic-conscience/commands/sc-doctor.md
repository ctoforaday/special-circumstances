---
description: Environment preflight — probe local tool dependencies, report READY / DEGRADED / BLOCKED. `/sc-doctor --fix` builds or fetches the sc-quality-gate hook binary (with consent).
---

Run the Special Circumstances environment preflight. Model terse-communication: table + verdict, no filler.

1. Read `${CLAUDE_PLUGIN_ROOT}/requirements.json`. For each entry in `tools[]`, run its `check_cmd`; record present/absent (+ version if present).
2. Detect the platform: `go env GOOS GOARCH` if Go is present, else `uname -sm` / PowerShell `$env:OS`.
3. Check the hook binary: does `${CLAUDE_PLUGIN_ROOT}/bin/sc-quality-gate` (`.exe` on Windows) exist?
4. Print a table — tool → ✓/✗ → version, or the `install[GOOS]` command if absent — then a verdict:
   - **READY**: every `required` tool present (recommended may be missing).
   - **DEGRADED**: a `recommended` tool or the hook binary is missing (hooks capability-gate and degrade).
   - **BLOCKED**: a `required` tool is missing.

If the argument is `--fix`: installing mutates the machine, so YOU MUST get explicit confirmation first (Semantic Consent), then provision only what is missing:

- **Hook binary `sc-quality-gate`** — the primary fix:
  - If Go is on PATH: `go build -C "${CLAUDE_PLUGIN_ROOT}/tools" -o "${CLAUDE_PLUGIN_ROOT}/bin/sc-quality-gate" ./cmd/sc-quality-gate` (append `.exe` on Windows).
  - Else: download the asset `sc-quality-gate_{goos}_{goarch}{exe}` from the latest `ctoforaday/special-circumstances` GitHub Release (`gh release download` or the browser URL), verify its SHA256 against the release `SHA256SUMS`, and place it in `${CLAUDE_PLUGIN_ROOT}/bin/`.
- **External tools** (git / gh / qlty) — print the `install[GOOS]` command; run it only on a second explicit confirmation, one tool at a time.

AFTER fixing, re-probe and reprint the verdict. YOU MUST NOT auto-run `--fix` at session start.
