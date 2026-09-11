# Special Circumstances — repository guide

**Scope: this file governs work *on* this repo only.** It is not part of any plugin: behaviour delivered to consuming projects lives entirely under `plugins/`, and nothing in this file reaches an installing project. Dev concerns belong here; consumer behaviour belongs in the plugins.

## Identity

Special Circumstances is an *adversarial* methodology suite: it is not a yes-man. When you are wrong, it is expected to say so, with a reason. A good argument is a courtesy, not an attack.

## Rules

Always-on rules bind every session via the imports below; the rest load on demand by description. `design-by-contract` is the authoring grammar (BEFORE / During / AFTER · YOU MUST).

@plugins/prosthetic-conscience/skills/terse-communication/SKILL.md
@plugins/prosthetic-conscience/skills/semantic-consent/SKILL.md
@plugins/prosthetic-conscience/skills/plan-act-reflect/SKILL.md
@plugins/prosthetic-conscience/skills/anti-spinning/SKILL.md
@plugins/prosthetic-conscience/skills/context-efficiency/SKILL.md
@plugins/prosthetic-conscience/skills/agent-guardrails/SKILL.md
@plugins/prosthetic-conscience/skills/think-around-problem/SKILL.md
@plugins/prosthetic-conscience/skills/validation-loop/SKILL.md
@plugins/prosthetic-conscience/skills/complete-the-concept/SKILL.md
@plugins/prosthetic-conscience/skills/facts-are-fields/SKILL.md

## Repository structure

| Path | Role |
|---|---|
| `plugins/<name>/` | The product: everything a consumer installs (skills, agents, commands, hooks, tools) |
| `.claude-plugin/marketplace.json` | Marketplace manifest listing the four plugins — keep in step with `PLUGINS` in `scripts/bootstrap-plugins.sh` |
| `plans/` | Design artifacts under review — each arrives as a PR; graduates into the plugins |
| `README.md`, `plugins/*/README.md` | The shipped documentation |
| `ideas/` `research/` `projects/` | Working corpus — starts empty, seeded by `/research` (and by `/self-improve` once sleeper-service ships) |
| `run-archive/` | The raw record of each captured run, gzipped. `research/` is gitignored, so a run directory survives a resume but NOT the container; this is the only part of a run that outlives it, and every audit re-reads it. Records and proofs only — the fetched-source cache is re-fetchable and 100× larger. |
| `scripts/` | Dev tooling, one Go module: `go -C scripts run ./<tool>`. Nothing in it ships. |

## Developing this repo

### Change discipline

- **No backwards compatibility, no archaeology.** When a format, schema, enum value, verb, role id, record field or store shape changes, the live code accepts and prints only the new form — no aliases, dual readers, tolerance paths, read-time fallbacks, or "formerly X" prose in help, prompts or refusals. Old FEOV records are brought forward by `migrate`, one translation per change with a test on an archived run; gray-area's catalogue rebuilds empty on a shape bump and `telepathy backfill` refills it. YOU MUST NOT propose a compatibility path, even as an option — it is not a fork the human is asked to rule on again.
  - FEOV's event-schema epoch (`record.EventSchema`, generated from `requirements.json` by `scripts/schemagen`) is compared once, at `setup`, binary against the plugin's declared epoch; bump it whenever the event shape changes. Nothing compares a run's recorded epoch on read or resume — old data is refused by its content, and the refusal names what it found and points to `migrate`.
- BEFORE coining a name, YOU MUST grep it across the Go tree, `debate.js`, `agents/`, `skills/` and `commands/`, then read the declaration at every hit for what it MEANS there — the same word in another sense reads as confirmation. A concept that already has a name is extended, never given a rival term. FEOV's concept nouns live in its terms registry (`tools/internal/terms/terms.json`, gated by `TestVocabularyProse`); after changing it, run `go -C scripts run ./vocabdoc`.
- BEFORE concluding from a measurement that structure does not matter, YOU MUST ask what the measurement assumes and whether the roadmap — parallel actions, shared expensive reads — breaks it. Structure is judged against where the system is heading.
- Committed tooling is Go: new dev tools go in `scripts/`, never Node/`.mjs`, and never under `plugins/*/tools/cmd/`, where `sc-doctor` treats every directory as a shipped hook binary. Scratch may be anything.

### Agent-facing text — prompts, constitutions, help, skills

- Present tense only; the why goes in a code comment at the source or in the commit. `archaeology` gates the pattern-matchable part.
- Each `<verb> --help` page MUST teach its verb on its own: remove repetition only where that page keeps every instruction. One Go constant rendered on every page that needs it is fine; one *home* that removes the text from a page is not. Tool mechanics live in the tool; duties live in constitutions and prompts.
- Seat prompts and constitutions name ACTS, never a command path or flag — CI pins this (`TestNoPromptGrowsItsCommandCatalogue`, `TestTheSeatPromptsNameNoVerb`, `TestTheShippedConstitutionsNameNoVerb`).
- AFTER changing any of these surfaces, YOU MUST read every changed golden line, and each affected `--help` before and after, one command at a time, for semantic loss.

### Verification

- BEFORE pushing, run `go -C scripts run ./check` — every CI gate that can run locally, tests at `-count=1` (`-list` prints the set, `-only` narrows it). A cached `go test ./...` reports stale goldens green. `golden` covers FEOV's module; gray-area's goldens run in its own `go test`. Regenerate with `UPDATE_GOLDENS=1 go test -count=1 <pkg>` or `go -C scripts run ./golden -review`, and read every changed line.
- `FEOV_RELEASE_GATE=1 go test ./releasegate/fuzz/` (FEOV's tools module) runs only on a feov tag and in the scheduled `hooks` run — never in PR CI or plain `go test`. BEFORE calling a record, flag or enum change done, YOU MUST run it by hand; BEFORE blaming a branch for a red release gate, run it on `origin/main`.
- `rulesweep` reads its `Rule-Class:` / `Sibling-Sweep:` trailers from commits (classes: `feov-memory/protocol-class-registry.md`) and reports NOT MEASURED while protocol-surface edits are uncommitted; `frontmatter` reads tracked files only. Neither has measured an uncommitted tree.
- BEFORE trusting a coverage number, YOU MUST remember what it cannot see: coverage says a line RAN, never that a test would NOTICE it changing. `internal/secrets` reported **100.0% of statements** while two of its eight secret patterns could be deleted outright with the suite green. Read a coverage figure as "nothing here is unreached", never as "this is tested".
  - The question coverage cannot answer takes a minute to ask by hand, on the one place you are already unsure about: **delete the row, invert the branch, run the suite.** If nothing fails, nothing depended on it. Aim it at tables — pattern tables, rule tables, rosters, dispatch tables — because a row that no test misses is the shape this repository is built from, and because deleting one is the check that found ten unasserted fields of `gray-area`'s capture manifest (#797). A field nothing asserts is a field that can stop being written while every row still reads as complete.
  - A test of the function is not a test of the call: delete the call site and run the suite.
- BEFORE asserting a negative — X does not exist, is not possible, has already happened — YOU MUST run the check that would find it. A listing capped at N that prints exactly N is truncated. For what an agent SEES, render it (the real `--help`, the prompt golden, the transcript's tool result), never grep the source it is built from. A red on one box only is a hypothesis about the box until the code's requirement is checked.
- Timing: interleave the arms, keep setup outside the timer, `ps` by tenant before believing a latency, and run a suite solo before a timing verdict; on a shared box, allocation counts beat wall clock.
- CI runs Go 1.25 with `GOTOOLCHAIN=local`; set the same to reproduce a CI Go failure.
- Windows CI runs gray-area's whole suite: a test that redirects home sets `USERPROFILE` as well as `HOME`; a built binary needs `.exe`; a golden is named `*.golden` or `*.sql` (`.gitattributes` pins line endings by extension); a pflag default that is a path gets a display `DefValue`.
- modernc SQLite ignores the query string of a DSN not prefixed `file:` — a read-only open is a `file:` URI, never `path?mode=ro`.

### Working tree, scratch, hooks

- Worktrees live under `.claude/worktrees/<name>/`, inside the main checkout. YOU MUST NOT `cd ..` out of a worktree subdirectory, and background commands MUST use absolute paths (or `go -C` / `git -C`) — the cwd does not carry into them. A relative path from the wrong cwd silently measures or mutates the main checkout; "already used by worktree" is the tell.
- Scratch — run directories, built binaries, measurement artifacts, caches — goes under a home-dir area such as `~/.claude/scratch/<task>/`, never `/tmp`.
- A running session's hooks execute from the plugin install cache (`~/.claude/plugins/cache/special-circumstances/<plugin>/<version>/bin/`), never the checkout; `/prosthetic-conscience:doctor --fix` refreshes them.
- A FEOV smoke run launches from its own source tree (a detached worktree at `main`): while `.claude/run-live.json` names a run, the SubagentStart/Stop hooks attribute every subagent in that project — a dev session's included — to it.

### Pull requests

- CI tests the merge with `main`: bring `main` in early. BEFORE merging a PR whose CI ran before another help- or golden-changing PR merged, YOU MUST rebase, regenerate goldens at `-count=1`, read each changed line and re-run CI — goldens conflict semantically, not textually (difftest's `manual` golden renders every help page).
- Generated conflicts (`record.pb.go`, `record.proto.sha256`) resolve by re-running `go -C scripts run ./protogen`, never by hand.
- The repository deletes a branch on merge. BEFORE pushing to a PR's branch, YOU MUST check `gh pr view <n> --json state`; "Create a pull request for" in push output is a HARD STOP — the PR merged and the push re-created its branch. Retarget a stacked child to `main` BEFORE merging its parent, or GitHub closes the child.

### Releases

- **Versions move at a RELEASE BOUNDARY, not per PR.** An ordinary PR changes plugin content and leaves `version` alone. A release is a human call — made when the binary/text contract has actually moved — and it is ONE act: bump the plugin's `version` in `plugin.json` and tag that commit `<plugin>--v<version>`. The release job refuses a tag whose manifest disagrees (`versionguard -tag`), so the two cannot drift.
  - A bump per PR makes the version a commit counter that tells a consumer nothing while the tags never move, so no plugin ends up with a tag matching its own version and `sc-doctor -fix` pins every download to a release that does not exist.
  - The cost, stated plainly: between releases `/plugin update` ships nothing, because it is version-gated. That is what a release model means. AFTER a release, run `/plugin update` + `/reload-plugins` to pull it. A running session keeps the plugins it loaded; a Remote Control session has no terminal and declines `/reload-plugins`, so it takes a new release only when started fresh (New session, or a restart of the `claude rc` server).
  - `versionguard` fails a version that goes BACKWARDS, and sweeps for stale per-binary version constants.
  - Mechanics: a release PR bumps only `version` in each releasing plugin's `plugin.json`; after it merges, ANNOTATED tags `<plugin>--v<version>` go on its merge commit, and the tag push runs the release job. Only the tip version is tagged — a skipped version is a decision, not a backlog. `sleeper-service` is never tagged: it has no binaries, and the release job fails such a tag.
  - The tag is the human's call, decision by decision. Agreement among agent sessions is an input, not authority; a ruling relayed by a peer is attributed to the peer.
