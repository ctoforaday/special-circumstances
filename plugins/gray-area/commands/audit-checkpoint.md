---
description: Put the checkpoint's validation-loop claims against what this session actually ran. Reports CITED / STALE / NO-EVIDENCE / UNCHECKABLE with provenance on both sides. `--json` for rows, or pass a transcript path to audit a different session.
---

Adjudicate a sealed checkpoint against the trajectory. Model [[terse-communication]]: relay the binary's rows, add no interpretation of your own.

The check logic lives in the tested binary, not in this prompt. Your job is to run it and report what it says.

1. Locate the note — active `projects/<name>/` or `research/<slug>/`, else `.claude/checkpoints/CHECKPOINT.md`. If none exists, say so in one line and stop. Do NOT reconstruct one; an invented note is adjudicated exactly as seriously as a real one.
2. Run `${CLAUDE_PLUGIN_ROOT}/bin/gray-area` (`.exe` on Windows) as `gray-area checkpoint <note>`, passing `--json` straight through if the caller gave it. With no transcript argument the binary resolves this session's from gray-area's own manifest and prints which row it used; pass a transcript path only when auditing a different session.
3. Relay the rows verbatim. Exit is non-zero when anything is `STALE` or `NO-EVIDENCE`.

**Read the verdicts as written, and do not upgrade them.**

- `CITED` — a matching invocation is in the trajectory, with its uuid.
- `STALE` — the claim's own trigger surface was written to after the claimed run time. The pass may have been real; it is not current. **Re-run that check and update the note** — that is the whole point of the row.
- `NO-EVIDENCE` — nothing matched. This is an ABSENCE, not a finding that the check did not run: the command may have been spelled in a form the tokens miss, or run in a different session. The row prints what was searched and how much was searched so you can tell which. YOU MUST NOT report it as "the check was never run".
- `UNCHECKABLE` — the claim names no command. Nothing was searched; nothing is being alleged.

**A `CITED` row confirms the ACT, never the OUTCOME.** The trajectory records what was **RUN**; it does not record what the run **SAID** — result bodies are conversation content and this plugin does not copy them. So a claim reading `` `go run ./check` … last run: pass `` splits into an act this record can check and a pass it cannot see. Every such row prints `NOT MEASURED:` naming the claimed outcome. YOU MUST NOT relay a `CITED` row as confirming the check passed; it confirms the command was invoked.

**What this establishes, and what it does not.** It establishes what the two records say when placed side by side. The transcript is append-only, not signed, so it is evidence of what was recorded, never proof that the record is authentic. YOU MUST NOT present a clean run as proof the work was done correctly — only that the note's claims and the session's acts agree.

**If the binary cannot resolve a transcript**, it says why and exits non-zero. The usual cause is that gray-area's `SessionStart` hook has not run in this project, so no session row exists. Report that; YOU MUST NOT search `~/.claude/projects/` for a likely-looking file — deterministic attribution is the property this plugin is built on, and a guessed transcript produces confident findings about the wrong session.

**Bootstrap** (only when the `gray-area` binary itself is missing): if Go is on PATH, build it — `go build -C "${CLAUDE_PLUGIN_ROOT}/tools" -o "${CLAUDE_PLUGIN_ROOT}/bin/gray-area" ./cmd/gray-area` (append `.exe` on Windows) — then run it. Otherwise report the absence and stop.

This command reads a session transcript, which carries user text, paths, and whatever tool results contained. It is a declared, scoped inspection of this project's own session: nothing leaves the machine, and no snapshot is written. See the plugin README on the line this plugin will not cross.
