# Plugin commands become skills

Written 2026-09-16T09:28:49Z against `cc6b0f7e`.

## I. Summary & Goals

**Objective.** Every `plugins/<plugin>/commands/<name>.md` moves to `plugins/<plugin>/skills/<name>/SKILL.md`. After the change no plugin has a `commands/` directory, and every gate that walked `commands/` walks the new paths instead.

**Why.** One entry-point concept has two homes. Claude Code merged commands into skills ("A file at `.claude/commands/deploy.md` and a skill at `.claude/skills/deploy/SKILL.md` both create `/deploy` and work the same way … For new work, prefer skills" — code.claude.com/docs/en/skills). The split confused the human locating `/frank-exchange-of-views:research` (2026-09-16).

**Non-goal: Agent Skills spec conformance.** agentskills.io's reference validator (`skills-ref`) rejects any frontmatter field outside `name, description, license, compatibility, metadata, allowed-tools`. `argument-hint` (on `research` and `plan-audit`) fails it. This plan does not chase that validator; our plugins are Claude Code plugins (hooks, agents, binaries).

**Ruled by the human.**
- Migrate all ten.
- Every migrated skill stays model-invocable: no `disable-model-invocation` on any of them. The hook binaries tell the agent to run `/prosthetic-conscience:doctor --fix` (5 sites) and `/prosthetic-conscience:resume` (2 sites), which requires that.

**Success criteria.**
1. `git ls-files 'plugins/*/commands/*'` prints nothing.
2. Each of the ten slash names resolves and runs from a plugin loaded out of this checkout, with `${CLAUDE_PLUGIN_ROOT}` expanded (§V.3).
3. `go -C scripts run ./check` green at `-count=1`.
4. Each gate that covered a command file still covers it at its new path — proven by the delete/invert checks in §V.2, not by a green run.

**Value case.** The naive version is a `git mv` of ten files. It would ship green while silently dropping coverage: `frontmatter` stops requiring a block on those files, the flag-spelling gate loses its operator exemption for `research` (red) or, if the glob is not added, stops reading `research` at all (silent), and `vocabulary_prose_test`'s glob errors on a missing file. The extra cost is about a dozen small edits across four gates, a few comments and one README heading.

## II. Technical Context

- **Harness.** Claude Code plugin loader. A plugin skill's slash name is `/<plugin>:<name>`: the frontmatter `name` sets the last segment, and the directory name applies only when `name` is absent. Nothing in the harness requires the two to agree, so after the move a slash name comes from a hand-typed field — hence the gate in B1. All 27 existing `SKILL.md` files (22+3+2) already have `name` equal to their directory (checked by a shell loop over `plugins/*/skills/*/SKILL.md`); after the move there are 37. `$ARGUMENTS`, `argument-hint`, `${CLAUDE_PLUGIN_ROOT}` all work in a plugin `SKILL.md` (docs, quoted above). Commands and skills share one listing, and the 1,536-character description cap applies to both.
- **Gates** are Go (`scripts/`, Go 1.25), plus FEOV's surface tests under `plugins/frank-exchange-of-views/tools/integration/surface`.
- **Release surface.** Plugin content only, so no `version` bump in this PR (release-boundary rule). Consumers get it at the next release of prosthetic-conscience, gray-area and FEOV.
- **Name clashes.** None of the ten names exists under its plugin's `skills/` today (checked: `[ -e plugins/$p/skills/$n ]` false for all ten). `prosthetic-conscience:probe` is also an agent name; agents and skills are separate namespaces, and the command already coexists with it.

## III. Proposed Changes

### A. The ten files — `[DELETE]` `commands/<name>.md`, `[NEW]` `skills/<name>/SKILL.md` (via `git mv`)

| Plugin | Names |
|---|---|
| prosthetic-conscience | checkpoint, doctor, plan-audit, probe, resume |
| gray-area | audit-checkpoint, audit-pr-body, audit-repetition, audit-seat-coverage |
| frank-exchange-of-views | research |

Per file:
- **Frontmatter.** Add `name: <name>` as the first field; keep `description` and `argument-hint` byte-identical. No `disable-model-invocation`.
- **Body.** Unchanged, except `probe`: its report rows are `skill / agent / command / hook / voice`, and step 1 is "**Command** — confirm this loaded". With no command directory left, the row checks nothing distinct from `skill`. Step 1 becomes "**Entry point** — confirm this skill loaded from `prosthetic-conscience`", and the table columns become `skill / agent / entry point / hook / voice`. Its `description` changes with it: "skill/agent/command/hook wiring" becomes "skill/agent/entry-point/hook wiring" — the only description that is not kept byte-identical.
- **Wording.** "This command reads…" in `audit-checkpoint` stays. To the user it is still a slash command; the word names how it is invoked, not the directory.

```
plugins/prosthetic-conscience/skills/{checkpoint,doctor,plan-audit,probe,resume}/SKILL.md
plugins/gray-area/skills/{audit-checkpoint,audit-pr-body,audit-repetition,audit-seat-coverage}/SKILL.md
plugins/frank-exchange-of-views/skills/research/SKILL.md
```

### B. Gates — `[MODIFY]`

1. **`scripts/frontmatter/main.go` `mustCarryFrontmatter`.** Today: `/agents/` or `/commands/`. New: `/agents/*.md`, or a file named `SKILL.md` directly under `plugins/<p>/skills/<name>/`. This is a class fix: the 22 existing skills were never in the mandatory set either, and a `SKILL.md` with no block loads with no name or description just like an agent. `references/*.md` stay optional. Update the refusal text ("An agent/command without one…" becomes "An agent or skill without one…"). `frontmatter_test.go:146` swaps `plugins/p/commands/c.md` for `plugins/p/skills/s/SKILL.md`, and adds a negative case: `plugins/p/skills/s/references/r.md` is NOT mandatory. The file header comment (`main.go:5`, "Every agent, command and skill") and the `mustCarryFrontmatter` doc comment (`main.go:111`, "a plugin's agents and commands") are updated to match.
   **Name equals directory (human ruling, 2026-09-16).** The same guard refuses a `plugins/<p>/skills/<d>/SKILL.md` whose `name` is absent or differs from `<d>`: "`<file>`: name `<n>` does not match its directory `<d>` — the slash command is `/<p>:<n>`, not `/<p>:<d>`." All 37 skill files, not only the ten. Tests: a mismatch case and an absent-name case fail; a matching case passes.
2. **`scripts/rulesweep/sweep.go:48`.** Remove the `/commands/[^/]+\.md$` pattern; `/skills/[^/]+/SKILL\.md$` (line 45) covers the new paths. `sweep_test.go:139`: `plugins/x/commands/doctor.md` becomes `plugins/x/skills/doctor/SKILL.md`.
3. **`scripts/archaeology/sweep.go:40`.** Remove `plugins/*/commands/*.md`; `plugins/*/skills/*/SKILL.md` covers the new paths. `sweep_test.go:44`: the path becomes `plugins/frank-exchange-of-views/skills/research/SKILL.md`.
4. **FEOV `surface/promptverbs_test.go`.**
   - Line 258: `{"commands","*.md"}` becomes `{"skills","research","SKILL.md"}`. The existing `research-protocol` globs do not reach it.
   - Line 527: the operator exemption `filepath.Base(filepath.Dir(path)) == "commands"` is keyed on a directory name. It becomes an exact match on the path, made repo-relative and `filepath.ToSlash`-normalised (Windows CI) before comparison — `repotree.Glob` returns absolute paths — against `plugins/frank-exchange-of-views/skills/research/SKILL.md`. A dir-name key of `research` would also exempt any future `skills/research/` elsewhere (facts-are-fields).
   - The comment above it keeps its reason (the operator types flags).
   - The `promptCatalogue` ratchet is keyed on basename, so `skills/research/SKILL.md` and `skills/research-protocol/SKILL.md` would share the key `"SKILL.md"`. Re-key the catalogue on the plugin-relative slash path, so each file carries its own pin: `promptverbs_test.go:632` `base := filepath.Base(path)` becomes the plugin-relative `filepath.ToSlash` path; every existing key in `promptCatalogue` (line 596) is rewritten to its path with its pin unchanged, and `skills/research/SKILL.md` gets its own entry at its measured count; the comment on the research-protocol entry stays with that entry. The same test fails when a `promptCatalogue` key names no file `agentFacingFiles` returns — otherwise a mistyped key silently falls back to "untracked, ceiling 0". The other `filepath.Base` uses (lines 299, 306, 791, 803, 811, 924) are message labels, not keys — unchanged.
5. **FEOV `surface/vocabulary_prose_test.go:153`.** `{plugin,"commands","research.md"}` becomes `{plugin,"skills","research","SKILL.md"}`.
6. **`scripts/pluginparity`.** No change. None of the ten descriptions contains `Always-on` (checked), so they are neither flagged nor expected in CLAUDE.md's imports. §V.1 runs it to confirm.

### C. Comments and docs — `[MODIFY]`, present tense

- `plugins/frank-exchange-of-views/tools/internal/setup/run.go:152,223` and `run_test.go:28,169,291`: `commands/research.md` becomes `skills/research/SKILL.md`. Bare `research.md` at `run_test.go:291` also changes.
- `plugins/frank-exchange-of-views/tools/internal/repotree/repotree.go:122`, `surface/docssurface_test.go:18`, `surface/literalsurface_test.go:16`: drop `commands/` from the lists of what the prompt gates walk.
- `CLAUDE.md:43` (naming rule): "`agents/`, `skills/` and `commands/`" becomes "`agents/` and `skills/`".
- `CLAUDE.md:29` "(skills, agents, commands, hooks, tools)" and `README.md:232` "skills, agents, commands, hooks, Go tools": drop "commands" — both describe what a plugin directory holds.
- `plugins/prosthetic-conscience/README.md:7` "22 skills, 5 commands, 2 agents" becomes "27 skills (5 of them slash commands), 2 agents". `README.md:75` "The base plugin: 22 skills" becomes "27 skills". `README.md:91` "The other eleven load on demand by description: …" becomes "Five are slash commands (`/checkpoint` · `/resume` · `/doctor` · `/plan-audit` · `/probe`); the other eleven load on demand by description: …" — so 27 = 11 always-on + 5 slash commands + 11 on demand. `plugins/prosthetic-conscience/README.md:15` ("Eleven load on every session, the rest by description") stays true — unchanged.
- `scripts/frontmatter/main.go:174` comment "An agent or command with no frontmatter" becomes "An agent or skill with no frontmatter", alongside :5, :111 and the :179 refusal. `frontmatter_test.go:151` — the negative list that asserts `plugins/p/skills/s/SKILL.md` is NOT mandatory — moves that path to the positive list.
- Census for this class (run): `git grep -n -i -E 'eleven|twenty|\b[0-9]+ (skills|commands)\b|agent or command|agent/command|agents and commands' -- README.md CLAUDE.md 'plugins/*/README.md' docs scripts/frontmatter` → `README.md:75,91`, `plugins/prosthetic-conscience/README.md:7,15,36`, `scripts/frontmatter/main.go:111,174,179`. `:36` ("the eleventh binary") counts binaries — unchanged.
- `plugins/prosthetic-conscience/README.md:73` `## Commands` becomes `## Slash commands`, with the same list. It lists what a user types, which is still a command.

### D. Not changed, with reason

- `plans/**`, `ideas/backlog.md`: design history of what shipped at the time, not live surfaces.
- `feov-memory/red-gap-patterns/pattern_doctrine_vs_implementation.md:36` and its `.claude/agent-memory` copy record a sleeper-service R4 finding where the file did ship under `commands/`. The lesson ("check where a file physically lives and what the harness does with that location") holds for `skills/` unchanged.
- `.github/workflows/hooks.yml:1380`, `scripts/frontmatter/main.go:12`: `probe.md` names the historical incident's file. Changing the path in the prose would misreport it, so the reason comment stays.
- Cobra `Commands()` / "Available Commands:" hits: CLI subcommands, a different sense of the word.

### E. Consumer census (run, results pasted)

```
git grep -n -I -E 'commands/|commands directory|/commands\b|"commands"|slash command' -- . ':!run-archive' ':!research' | grep -v -E '^(plans|ideas)/'
```
Hits (each disposition above): `CLAUDE.md:43` C · `feov-memory/…doctrine…:36` + agent-memory copy D · `surface/docssurface_test.go:18` C · `surface/literalsurface_test.go:16` C · `surface/promptverbs_test.go:258,527` B4 · `surface/vocabulary_prose_test.go:153` B5 · `proof/errorsignature_test.go:13` (`ABSENT commands/self-improve.md` fixture text) — unchanged: a synthetic missing-path string in a proof fixture, not a walk of the tree · `repotree.go:122` C · `setup/run.go:152,223` C · `setup/run_test.go:28,169` C · `gray-area/commands/audit-checkpoint.md:28` A (prose, moves) · `scripts/archaeology/sweep.go:40`, `sweep_test.go:44` B3 · `scripts/frontmatter/main.go:121`, `frontmatter_test.go:146` B1 · `scripts/rulesweep/sweep.go:48`, `sweep_test.go:139` B2 · others (`hooks/README.md:11`, `requirements.json:67`, `doctor/main.go:71`, `hookinvocation/invocation_test.go:124`, `qlty CHEATSHEET:53`, `gray-area main_test.go:152`) are "command" in the hook-command / CLI / shell sense — unchanged.

```
git grep -n -I -E '\b(checkpoint|doctor|plan-audit|probe|resume|audit-checkpoint|audit-pr-body|audit-repetition|audit-seat-coverage|research)\.md\b' -- plugins scripts README.md CLAUDE.md .claude-plugin .github feov-memory
```
Hits: `hooks.yml:1380` D · `frontmatter/main.go:12`, `frontmatter_test.go:39,41` D (the `probe.md` fixture name is a test input string) · `setup/run_test.go:291` C · plus the B/C hits above.

```
git grep -n -E '"skills", ?"\*"|skills/\*' -- '*.go' '*.js' '*.sh' '*.yml' '*.json'
```
Enumerators of `skills/*`: `archaeology/sweep.go:41-43` (gains ten files: must stay green, §V.1) · `pluginparity/main.go:279` (B6).

Slash names emitted by hook/tool code (`git grep -ohE '/(prosthetic-conscience|gray-area|frank-exchange-of-views):[a-z-]+' -- 'plugins/*/tools/**.go' | sort | uniq -c`): `doctor` ×5, `resume` ×2. This is why no skill gets `disable-model-invocation`.

Every other site naming one of the ten slash names (`git grep -n -E '/(prosthetic-conscience|gray-area|frank-exchange-of-views):(checkpoint|doctor|plan-audit|probe|resume|audit-[a-z-]+|research)\b' -- ':!plans' ':!ideas' ':!run-archive' ':!plugins/*/tools/**.go' | cut -d: -f1 | sort | uniq -c`): `.claude-plugin/marketplace.json` 1 · `.gitignore` 1 · `.qlty/qlty.toml` 1 · `CLAUDE.md` 1 · `README.md` 9 · `docs/setup-script.md` 2 · `plugins/frank-exchange-of-views/{README.md 1, bin/.gitkeep 1, hooks/fetch-bin.sh 1, hooks/hooks.json 3, requirements.json 1}` · `plugins/gray-area/{README.md 1, bin/.gitkeep 1, commands/audit-checkpoint.md 1, hooks/fetch-bin.sh 1, hooks/hooks.json 4, requirements.json 2}` · `plugins/prosthetic-conscience/{.claude-plugin/plugin.json 1, README.md 2, bin/.gitkeep 1, hooks/fetch-bin.sh 1, hooks/hooks.json 10, requirements.json 2, skills/qlty-proficiency/SKILL.md 1}` · `plugins/sleeper-service/bin/.gitkeep 1` · `scripts/bootstrap-plugins.sh 1` · `scripts/fetchbingen/fetch-bin.sh 1`. **All unchanged**: each names a slash name, and B1's name gate plus §V.3 invoking all ten prove every name survives.

Bare-word sweep (`git grep -n -i -w -E 'commands?' -- README.md CLAUDE.md 'plugins/*/README.md'`, read by hand; CLI, shell, hook-command and "run a command" senses set aside): `CLAUDE.md:29`, `README.md:75` (count, found via "22 skills"), `README.md:232`, `plugins/prosthetic-conscience/README.md:7,73` → §III.C. `README.md:111` (`| Command | Question it answers |`, a table of slash commands and CLI verbs) and `plugins/gray-area/README.md:288` (`## The command`, the audit slash command) → unchanged: invocation sense.

**Completion test for the census.** After the edit, re-running the first command with `commands/` restricted to directory paths (`git grep -n -E 'plugins/[^ ]*/commands/|"commands", ?"|/commands/'`) outside `plans/`, `ideas/`, and §D's list prints nothing.

## IV. Risk & Mitigation

| # | Risk | L × I × C | Mitigation |
|---|---|---|---|
| R1 | `${CLAUDE_PLUGIN_ROOT}` not expanded in a plugin SKILL.md; doctor/gray-area audits/research run a literal path | low (docs say it expands) × high × low | §V.3 runs `doctor` and `research`'s path line from a checkout-loaded plugin before merge |
| R2 | A gate silently stops covering a moved file (glob misses, exemption misses) | medium × high × low | B1–B5; §V.2 delete/invert each |
| R3 | Skill auto-invocation now triggers on description match more eagerly than a command did (e.g. `research` launching a debate unasked) | unknown × medium × medium | Human ruled model-invocable. §V.3 driveable check records whether a plain prompt ("where is the research skill file") invokes it. Reported, not gated |
| R4 | Goldens that render the skill listing or the research prompt change | low × low × low | §V.1 `check` at `-count=1`; read any changed golden line |
| R5 | `frontmatter` widening fails on an existing SKILL.md with a broken block | low × low × low | That would be a real latent defect it surfaces; fix in the same PR |
| R6 | A stacked or open PR edits a `commands/*.md` file and conflicts | medium × low × low | `gh pr list --state open --json number,files --jq '.[] | select(any(.files[]; .path | test("/commands/"))) | .number'` before starting; rebase-on-main rule |
| R7 | A `name:` typo changes a slash name silently | low × high × low | B1 name gate; §V.2f; §V.3 invokes all ten |

## V. Verification Plan

1. `go -C scripts run ./check` → all gates green at `-count=1` · re-armed by any file in §III.
   Includes frontmatter, rulesweep, archaeology, pluginparity, golden, FEOV surface tests.
   `rulesweep` reads commit trailers: run after committing. The commit touches protocol surfaces at old and new paths, so it carries `Rule-Class:` and `Sibling-Sweep:` trailers (classes from `feov-memory/protocol-class-registry.md`).
2. **Coverage-survives checks, by hand** · re-armed by B1–B5:
   - a. Strip the frontmatter from `prosthetic-conscience/skills/doctor/SKILL.md` → `go -C scripts run ./frontmatter` fails naming it. Restore.
   - b. Add a `--lanes` token to a seat-facing file (e.g. `skills/research-protocol/SKILL.md`) → the flag gate fails. The same token in `skills/research/SKILL.md` → it passes (exemption holds). Revert both.
   - c. Rename `skills/research/SKILL.md` temporarily → `vocabulary_prose_test` and the promptverbs glob fail loudly (repotree.Glob refuses the empty set). Restore.
   - d. Add a "formerly" line to a moved SKILL.md → `archaeology` fails naming it. Revert.
   - e. Add one seat-verb invocation to `skills/research/SKILL.md` → the command ratchet fails naming `skills/research/SKILL.md` (not `SKILL.md`). Rename a `promptCatalogue` key to a nonexistent path → fails. Revert both.
   - f. Change `name: resume` to `name: resumes` → `frontmatter` fails naming the file and both slash names. Delete the `name:` line → fails. Restore.
3. **Driveable check on real data** · re-armed by §III.A:
   - Preconditions: no `.claude/run-live.json` in this worktree (else SubagentStart/Stop attribute to that run). Expected, gitignored side effects: `hooks/fetch-bin.sh` under `--plugin-dir` may download release binaries into each plugin's `bin/`, and the prosthetic-conscience seal hooks may write under the worktree's `.claude/checkpoints/`. Before the runs, build the hook binaries into each plugin's `bin/` (`go -C plugins/<p>/tools build -o ../bin/ ./cmd/...`) so `doctor` does not need its `go build` bootstrap, which the narrow `--allowedTools` denies.
   - Launch `claude -p` headless from this worktree with `--plugin-dir` for each of the three plugins (the docs say `--plugin-dir` takes precedence over the installed plugin of the same name). Permissions: `--allowedTools "Bash(*/bin/sc-doctor*) Bash(*/bin/gray-area*) Read"` — never `--dangerously-skip-permissions`, so the no-topic `research` invocation cannot launch a Workflow.
   - Invoke all ten slash names, one `claude -p` each. For each, the transcript shows the skill loaded under its original `/<plugin>:<stem>` name (not "unknown command").
   - `doctor`, the four gray-area audits, `research`: the Bash `tool_use` input (or the loaded skill text) holds the checkout's absolute path — not the cache path and not a literal `${CLAUDE_PLUGIN_ROOT}`.
   - `/frank-exchange-of-views:research` with no topic → asks for a topic (reads `$ARGUMENTS` as empty) without launching.
   - `/prosthetic-conscience:plan-audit plans/commands-to-skills.md` → the path reaches the body through `$ARGUMENTS`.
   - `/prosthetic-conscience:doctor --fix` (the form the hooks emit) and `/prosthetic-conscience:checkpoint --show` → the transcript shows the harness-appended `ARGUMENTS: --fix` / `ARGUMENTS: --show` line for these bodies, which have no `$ARGUMENTS` placeholder.
   - Bare `/checkpoint` and `/probe` (the names the plugin README lists) → record what each resolves to; `/resume` and `/doctor` stay the built-ins by precedence. Recorded, not gated.
   - Transcript of each kept under `~/.claude/scratch/commands-to-skills/`.
   - Also: a plain prompt "where is the research skill file" — record whether `research` auto-invokes (R3).
4. Census completion command (§III.E) → no output.
5. `claude plugin validate plugins/<p>` for the three plugins → exit 0, before and after (a pre-existing failure is recorded, not attributed to this change).
6. `git ls-files 'plugins/*/commands/*'` → empty.
7. Auditor gate: `/plan-audit plans/commands-to-skills.md`, before implementation.
