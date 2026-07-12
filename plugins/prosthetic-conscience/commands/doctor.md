---
description: Environment preflight — runs the tested sc-doctor binary for a deterministic table + READY / DEGRADED / BLOCKED verdict. `--fix` (with consent) builds or fetches missing hook binaries.
---

Run the Special Circumstances environment preflight. Model terse-communication: relay the binary's output, no filler.

1. Run `${CLAUDE_PLUGIN_ROOT}/bin/sc-doctor` (`.exe` on Windows) and relay its table + verdict verbatim — the check logic lives in the tested binary, not in this prompt.
2. **Bootstrap** (only when the sc-doctor binary itself is missing): if Go is on PATH, build it — `go build -C "${CLAUDE_PLUGIN_ROOT}/tools" -o "${CLAUDE_PLUGIN_ROOT}/bin/sc-doctor" ./cmd/sc-doctor` (append `.exe` on Windows) — then run it. Else download the asset `sc-doctor_{goos}_{goarch}{exe}` from the latest `ctoforaday/special-circumstances` GitHub Release, verify its SHA256 against the release `SHA256SUMS`, place it in `${CLAUDE_PLUGIN_ROOT}/bin/`, and run it.
3. If the argument is `--fix`: fixing mutates the machine, so YOU MUST get explicit human confirmation first (see [[semantic-consent]]), then run `sc-doctor -fix` — it rebuilds (or prints fetch instructions for) every missing hook binary and reprints the verdict.
4. External tools (git / gh / qlty): the binary prints the per-platform install command but never runs it; execute one only on a second explicit confirmation, one tool at a time.

YOU MUST NOT auto-run `--fix` at session start.
