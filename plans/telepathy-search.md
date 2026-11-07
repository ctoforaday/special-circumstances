# `telepathy search`: fast transcript query and coordination guidance (#1776)

> STATUS: approved by gblock 2026-11-06; implemented and under review in PR #1776.

## I. Summary & Goals

**The defect.**
Agents coordinating across harnesses (Claude Code Remote Control and Antigravity Remote Control) repeatedly bypassed fast transcript lookup in favor of ad-hoc, slow raw SQL queries via `telepathy sql`.

Root causes:
1. `AGENTS.md` explicitly mandated `telepathy agents` / `telepathy sql`, failing to mention transcript lookup.
2. `plugins/gray-area/skills/telepathy/SKILL.md` Section 1 provided an invalid internal-table SQL recipe (`SELECT session_id, bridge_session_id, cwd, project_dir FROM session WHERE closed_at IS NULL;`), which breached the published view contract (`v_*`) and provided no guidance for locating sessions by PR, branch, or issue.
3. Vocabulary mismatch: Agents intuitively reached for `telepathy search <term>`. Because Cobra registered only `telepathy find`, the command failed with exit code 2 (`unknown command "search"`).

**Objective.**
Cleanly rename the verb `find` to `search` across the codebase, test suites, golden outputs, and documentation in strict accordance with the repository's change discipline (`AGENTS.md`: *No backwards compatibility, no archaeology*). Overhaul cross-harness coordination guidance in `AGENTS.md` and `plugins/gray-area/skills/telepathy/SKILL.md` to establish fast transcript search (`telepathy search`) and file queries (`telepathy touched`) as the mandatory first-line coordination mechanisms.

**Success criteria.**
1. `telepathy search <term>` executes ripgrep over transcripts in ~50ms, returning attributed hits by speaker channel.
2. `telepathy --help` lists and documents `search`, with zero references to `find`.
3. `AGENTS.md` mandates `telepathy search <term>` and `telepathy touched <path>` before raw SQL.
4. `plugins/gray-area/skills/telepathy/SKILL.md` provides copy-pasteable recipes for finding session owners by PR, branch, issue number, and touched files, and corrects the SQL fallback to use published view `v_session`.
5. All CI check gates (`golden`, `vet`, `fmt`, `frontmatter`, `pluginparity`, `rulesweep`) pass cleanly.

## II. Technical Context

- Language: Go 1.27 (`GOTOOLCHAIN=local`).
- CLI framework: `github.com/spf13/cobra`.
- Dependencies: `ripgrep` (required by gray-area).
- Store: SQLite (`$XDG_STATE_HOME/special-circumstances/catalogue/catalogue.db`).
- Trajectory source: local harness transcripts (`~/.claude/projects`, `~/.gemini/antigravity-cli/brain`).

## III. Proposed Changes (the spec)

```
plugins/gray-area/tools/internal/telecli/
├── search.go                      [RENAME from find.go; MODIFY to searchCmd]
├── search_test.go                 [RENAME from find_test.go; MODIFY]
├── root.go                        [MODIFY: register newSearchCmd, example]
├── root_test.go                   [MODIFY: sorted verbs test]
├── golden_test.go                 [MODIFY: golden test verbs slice]
├── session.go                     [MODIFY: references from find to search]
└── testdata/golden/
    ├── help-search.golden         [RENAME from help-find.golden]
    ├── search-*.golden            [RENAME 12 find-*.golden files]
    ├── help-root.golden           [MODIFY: regenerated]
    └── help-session.golden        [MODIFY: regenerated]

plugins/gray-area/
├── README.md                      [MODIFY: verb table and explanation]
├── requirements.json              [MODIFY: ripgrep requirement description]
└── skills/
    ├── telepathy/SKILL.md         [MODIFY: overhaul Section 1, rename find -> search]
    ├── elicitation-testing/SKILL.md [MODIFY: telepathy search reference]
    └── restart-recovery/SKILL.md  [MODIFY: telepathy search reference]

AGENTS.md                          [MODIFY: line 84 coordination rule]
README.md                          [MODIFY: switchboard verb table & requirements]
plans/README.md                    [MODIFY: index this plan]
```

### Consumer Census

Exact search: `git grep -E '\btelepathy find\b|\bnewFindCmd\b|\bfind-.*\.golden\b'`

| Consumer Site | Action Taken |
|---|---|
| `plugins/gray-area/tools/internal/telecli/find.go` | Renamed to `search.go`; `newFindCmd` -> `newSearchCmd`, `find <term>` -> `search <term>` |
| `plugins/gray-area/tools/internal/telecli/find_test.go` | Renamed to `search_test.go`; updated test calls |
| `plugins/gray-area/tools/internal/telecli/root.go` | Updated verb registration to `newSearchCmd(&env)` |
| `plugins/gray-area/tools/internal/telecli/golden_test.go` | Updated `verbs` list to replace `"find"` with `"search"` |
| `plugins/gray-area/tools/internal/telecli/session.go` | Updated error/help string referencing `find` to `search` |
| `plugins/gray-area/tools/internal/telecli/testdata/golden/help-find.golden` | Renamed to `help-search.golden` |
| `plugins/gray-area/tools/internal/telecli/testdata/golden/find-*.golden` (12 files) | Renamed to `search-*.golden` |
| `AGENTS.md` | Updated line 84: mandate `telepathy search` and `telepathy touched` before raw SQL |
| `plugins/gray-area/skills/telepathy/SKILL.md` | Replaced all `find` occurrences with `search`; updated Section 1 coordination recipes |
| `plugins/gray-area/skills/elicitation-testing/SKILL.md` | Replaced `telepathy find` with `telepathy search` |
| `plugins/gray-area/skills/restart-recovery/SKILL.md` | Replaced `telepathy find` with `telepathy search` |
| `plugins/gray-area/README.md` | Updated command listings and explanation |
| `plugins/gray-area/requirements.json` | Updated `_tier_comment` mentioning `find` to `search` |
| `README.md` | Updated switchboard command table |

## IV. Risk & Mitigation

| Risk | Assessment | Mitigation |
|---|---|---|
| Broken consumer scripts relying on `find` | low × low × low | Change discipline strictly mandates no aliases or backwards compatibility. Scripts use new verb. |
| Stale goldens fail CI | certain × high × low | Run `go test ./internal/telecli -update` and verify diff with `git diff`. |
| Protocol rulesweep fails on uncommitted skill changes | high × medium × low | Commit skill changes with `Rule-Class: contract-changed-caller-unchanged` and `Sibling-Sweep:` trailer. |

## V. Verification Plan

1. **Unit & Golden Tests**:
   ```bash
   go -C plugins/gray-area/tools test -count=1 ./...
   ```
2. **Check Gates**:
   ```bash
   go -C scripts run ./check -only rulesweep,golden,vet,fmt,frontmatter,pluginparity
   ```
3. **Live CLI Smoke Test**:
   ```bash
   go -C plugins/gray-area/tools run ./cmd/telepathy --help
   go -C plugins/gray-area/tools run ./cmd/telepathy search "1024"
   ```
   Verify `search` is displayed alphabetically between `link` and `send`, help text is accurate, and search runs against real transcripts.
