# Plugin `hooks.json` holds only documented keys — the prose moves to `hooks/README.md` (#913)

> STATUS: approved by gblock 2026-09-11 ("implement, PR, merge when green"). TinySpec; the five
> sections are kept for the record.

## I. Summary & Goals

Claude Code 2.1.268 prints `hooks.json: unknown key "_comment" … ignored` at the start of every
session for each `_comment` key in a plugin's `hooks/hooks.json` (observed 2026-09-11 in a fresh
session: prosthetic-conscience and gray-area named). The hooks still run; every consumer's every
session opens with the noise.

**Objective:** no plugin `hooks.json` carries a key the client does not document, the reasons the
`_comment`s held survive verbatim next to the wiring, and a gate stops the next one landing.

**Success criteria:** (1) the three `hooks.json` files hold only documented keys; (2) every
`_comment` text is in that plugin's `hooks/README.md`, one section per event; (3)
`scripts/validatejson` fails an undocumented key and a wrongly shaped file, and is loud if it
finds no plugin `hooks.json` to check; (4) a scratch session with the edited plugins prints no
"unknown key" line.

## II. Technical Context

- 15 `_comment` keys on main (5eefa7df): prosthetic-conscience 10 (top + 9 matcher groups),
  frank-exchange-of-views 4 (top + `PreToolUse`, added by #914, + `SubagentStart` +
  `SubagentStop`), gray-area 1 (top); sleeper-service has no `hooks.json`. (#913 counted 14 on the
  tree before #914.)
  No code reads a `_comment` (`grep -rn _comment` over scripts, plugins, .github, excluding the
  files themselves). No `hooks.json` reader decodes strictly, so a top-level `description` breaks
  none of them.
- The hooks reference (https://code.claude.com/docs/en/hooks) documents an optional top-level
  `description`, matcher groups of `matcher` + `hooks`, and hook fields `type, if, timeout,
  statusMessage, once` plus per-type fields; the plugins reference shows only `hooks`. No
  documented per-entry free-text field exists (`statusMessage` is displayed while the hook runs).
- `claude plugin validate --strict` passes all four plugins WITH the 15 `_comment` keys — the
  client's validator does not check `hooks.json`, so the documented key set is kept by hand in
  the gate, with that reason stated where it is defined.
- `jq` rewrites prosthetic-conscience's `—` escapes on 20 command lines, so the `_comment`
  lines are deleted by line, not by a `jq` rewrite; every other byte is unchanged.
- The `archaeology` gate scans agents, commands and skills Markdown and `CLAUDE.md` only; a
  `hooks/README.md` is not an agent-facing surface, so the rationale moves verbatim.

## III. Proposed Changes

- `plugins/{prosthetic-conscience,frank-exchange-of-views,gray-area}/hooks/hooks.json` — delete
  every `_comment`; add a one-line top-level `description`.
- `plugins/*/hooks/README.md` [NEW] — every `_comment` text verbatim under its event. One sentence
  changes: prosthetic-conscience's `Stop` rationale said "the only hooks.json check in CI tests
  bootstrap-guard degradation", which this change makes false; it now names `validatejson`'s key
  check too, and keeps its point (nothing checks that a built binary is registered).
- `scripts/validatejson/{hooks.go,hooks_test.go}` [NEW], `main.go` [MODIFY] — `hookKeyProblems`
  over every tracked `plugins/*/hooks/hooks.json`; refuse undocumented keys at each level and a
  file that is not the documented shape; fail if no such file is tracked.

## IV. Risk & Mitigation

| Risk | Mitigation |
|---|---|
| The docs add a field the gate refuses | One-line widening; the refusal names the key and the reason |
| A reason is lost in the move | Texts copied from the decoded JSON; the review compares each README section to the deleted `_comment` |
| The client rejects `description` too | §V.4 measures it in a scratch session |

## V. Verification Plan

1. `go -C scripts test -count=1 ./validatejson` and `go -C scripts run ./validatejson` (clean on
   the edited tree; re-add one `_comment` locally and it MUST fail naming it).
2. `jq -e` parses each edited `hooks.json`; `git diff` shows only `_comment` deletions and the
   `description` lines.
3. The plugins' own suites that read `hooks.json` (`go -C plugins/<p>/tools test -count=1 ./...`
   for all three) and the repo gates: `pluginparity`, `frontmatter`, `archaeology`, `rulesweep`,
   `golden`.
4. A scratch session, Remote Control off, loading the three edited plugins with `--plugin-dir`:
   its startup prints no "unknown key" line for them (the installed copies, which still carry
   `_comment`, are distinguished by their path in the warning or disabled for the run).
