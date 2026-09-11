# Hook binaries install themselves; plugin content is pinned to its release

Closes the `ideas/backlog.md` item "empty-bin window after a version bump". Rulings (gblock,
2026-09-11): self-heal for ordinary installs, manual updates and unobserved updates; pin the
marketplace to release tags; binaries stay in the plugin cache's `bin/`; FEOV stays broken
until v2 and gets no dedicated work here. After the first audit (FAIL): notification is
event-aware; a CI job proves the fetcher on macOS and Windows; the secret-guard window (R7) is
accepted as stated. After the second audit (FAIL): a displaying event that finds a fetch already
running announces it; the fetcher checks its prerequisites up front and names the missing one.
After the third audit (FAIL): a fetch installs all or nothing; the doctor's gh-based `-fix`
fetch and the new curl-based script both stay.

## I. Summary & Goals

**Problem, measured 2026-09-11 on this machine.**

1. Every plugin version change creates a cache directory whose `bin/` holds only the tracked
   `.gitkeep`, and nothing on any machine fills it: `/plugin update`, marketplace auto-update
   and the rc unit's updater move content only, and `commands/doctor.md` forbids an automatic
   `--fix`. Each hook's guard then prints one stderr line and exits 0 — and Claude Code routes
   stderr on exit 0 to the debug log only. This session's transcript holds 46 such warnings
   (`hook_success`, `exitCode: 0`, stderr only); neither the human nor the agent saw one.
   prosthetic-conscience 0.44.0 and gray-area 0.11.0 ran for hours with every hook a no-op,
   the secret guard included. The doctor's own EMPTY-BIN warning could not fire either: it tests
   `len(ReadDir(bin)) == 0`, and `.gitkeep` makes that false.
2. A relative-path marketplace source copies plugin content from the marketplace's HEAD; only
   the version string is pinned. The 0.44.0 cache was copied from `eb9d0e13` (main) while its
   release binaries were built at `02d9901` (the tag). This is FEOV's breakage — main's
   `hooks.json` calls `feov-subagentstart`, which no release ships — and the doctor's false
   "stale" on all 13 freshly fetched binaries.

**Goals and success criteria.**

| # | Goal | Measured by |
|---|---|---|
| G1 | A fresh install and an update to a new version each end with every `tools/cmd` binary of the plugin present in `bin/`, checksum-equal to the release's `SHA256SUMS`, with no human action | V-R1, V-R2 |
| G2 | While binaries are missing the human is told at the first displaying hook event, then at most once per 10 minutes; a non-displaying event never spends that budget | contract test (event matrix), V-R1 attachment, V-R4 human check |
| G3 | The guard never delays a tool call for the download: a guard invocation with missing binaries returns in < 1 s | V-R1 `durationMs`, CI guard gate, `fetch-bin-platforms` job |
| G4 | For each tagged plugin, installed content is identical to its release tag's tree and the install record's `gitCommitSha` is the tag's commit; `sc-doctor` reports no "stale" on it | V-R5 |
| G5 | A failed fetch (no network, release not yet published, checksum mismatch, asset missing, unsupported platform, `curl` or checksum tool missing) leaves no binary in `bin/`, is retried later, and names its cause and the manual command to the human | contract test, V-R3 |
| G6 | The fetcher installs and its detached child survives its parent's exit on Linux, macOS and Git Bash on Windows | `fetch-bin-platforms` CI job |

**Non-goals.** FEOV behaviour (pinning incidentally hands consumers FEOV 1.64.0 exactly as
released, which is internally consistent). A multicall binary to shrink the download. Moving
binaries to `${CLAUDE_PLUGIN_DATA}` (ruled out: `bin/` is on the Bash PATH and skills call
`telepathy` bare). Blocking tool calls while the secret guard is absent (R7, ruled).

## II. Technical Context

- **Claude Code 2.1.268** (this box). Hook events include no install, update or enable event;
  `Setup` fires only under `--init`, `--init-only` or `--maintenance`.
- **Hook output.** On exit 0, stdout and stderr go to the debug log, except plain stdout on
  SessionStart and UserPromptSubmit, which becomes agent context. A TOP-LEVEL
  `{"systemMessage": …}` is displayed (a `hook_system_message` transcript attachment) — measured
  per event in step 0:

  | Event | systemMessage | Wired to a guard in |
  |---|---|---|
  | SessionStart | displayed (S2, S3) | prosthetic-conscience, gray-area |
  | PreToolUse | displayed (S2) | prosthetic-conscience, FEOV |
  | PostToolUse | displayed (S3) | prosthetic-conscience |
  | PostToolUseFailure | displayed (S3) | prosthetic-conscience |
  | Stop | displayed (S3) | prosthetic-conscience, gray-area |
  | SubagentStart, SubagentStop | attachment confined to the subagent's own transcript, never the human-facing one (S3) — not displayed | all three |
  | SessionEnd, PreCompact, PostCompact, FileChanged | not measured (SessionEnd fired in neither spike transcript; the others cannot be driven headless) — treated as not displayed | prosthetic-conscience, gray-area (SessionEnd) |

  The same text inside `hookSpecificOutput` produced no attachment. Command hooks accept
  `"async": true`.
- **Detach.** A child started with `nohup … &` from a SessionStart or PreToolUse hook outlived
  both the hook and the session (S1, Linux, headless).
- **Windows:** hook commands run in Git Bash when installed, otherwise PowerShell. The current
  guards are already POSIX shell, so Git Bash is already required; unchanged here.
- **Marketplace sources.** A relative path is copied from the marketplace clone at HEAD;
  `git-subdir` takes `url`, `path`, `ref`, `sha`. Probed 2026-09-11 in an isolated
  `CLAUDE_CONFIG_DIR` (`~/.claude/scratch/pin-probe`): pinned to `prosthetic-conscience--v0.44.0`,
  the install is identical to `git archive` of the tag (only the client's `.in_use` marker
  differs; main differs from the tag in `commands/doctor.md`, so the probe discriminates), and
  the record's sha is `02d9901`. Moving `ref` to v0.43.0 and back updated both ways. A `ref`
  naming a tag that does not exist fails the update loudly ("Remote branch … not found") and
  keeps the installed version.
- **Releases.** A tag push runs the release job, which builds six pairs
  `{windows,darwin,linux} × {amd64,arm64}` as `<name>_<os>_<arch>[.exe]` plus `SHA256SUMS`
  (`sha256sum` format). The repository is public; an unauthenticated download of an asset
  returns 200. linux/amd64 totals: prosthetic-conscience 11 assets, 28 MB; gray-area 3 assets,
  17 MB.
- **Tracked contents of `bin/`:** `.gitkeep` in prosthetic-conscience, FEOV and sleeper-service.
- **Consumer prerequisites for the fetcher:** `sh`, `curl`, `uname`, `mkdir`, `mv`, and one of
  `sha256sum` / `shasum -a 256`. No `gh`, no Go.
- **CI runners:** `windows-latest` already runs three jobs; no job runs on macOS.
- **Constraints:** the fetch reads only the plugin's own pinned tag, never "latest"; every file
  is checksum-verified before it is made executable; writes stay inside `${CLAUDE_PLUGIN_ROOT}`
  (`bin/` and the fetcher's state directory `.fetch/`); nothing is sent anywhere but GitHub's
  release download URL.

## III. Proposed Changes

### Step 0 — spikes (done 2026-09-11; no shipped code)

Scratch: `~/.claude/scratch/selfheal-spike/` and `~/.claude/scratch/selfheal-spike2/`; each is a
throwaway plugin loaded with `--plugin-dir` into `claude -p`, with its transcript under
`~/.claude/projects/-home-gblock-ctoforaday-com--claude-scratch-selfheal-spike{,2}-work/`.

- **S1, detach survival.** SessionStart and PreToolUse hooks each started
  `nohup sh -c 'sleep 25; touch $MARK' >/dev/null 2>&1 &` and exited. The session ended at
  17:31:46; the marks appeared at 17:32:06 and 17:32:08. The guard starts the fetch itself.
- **S2, systemMessage placement.** Top level → `hook_system_message` on SessionStart and
  PreToolUse; nested in `hookSpecificOutput` → nothing.
- **S3, systemMessage per event.** The table in §II. A first S3 run emitted invalid JSON through a
  quoting error; each event then recorded a `hook_non_blocking_error` with the parse error, which
  says nothing about display. The recorded run is the second.

### A. The fetcher `[NEW]`

```
scripts/fetchbingen/
  fetch-bin.sh          [NEW] the one authored copy
  main.go               [NEW] copies it into every plugin that has tools/cmd; -check fails on drift
                              and on any hooks.json guard whose event argument is not its event key
  main_test.go          [NEW]
plugins/prosthetic-conscience/hooks/fetch-bin.sh     [NEW, generated]
plugins/gray-area/hooks/fetch-bin.sh                 [NEW, generated]
plugins/frank-exchange-of-views/hooks/fetch-bin.sh   [NEW, generated]
```

Three copies because a plugin cannot reach another plugin's cache; generated, per `buildidgen`'s
precedent, so there is one authored source and a staleness gate rather than a guard over three
hand-kept files.

**State lives in `${CLAUDE_PLUGIN_ROOT}/.fetch/`, never in `bin/`.** `bin/` holds binaries and
`.gitkeep` only, so nothing that lists `bin/` (the doctor, FEOV's `binariesIn`, the Bash PATH)
sees the fetcher's files. `.fetch/` sits on the same filesystem as `bin/`, so the final `mv` is
atomic. Files: `.fetch/lock/` (directory lock), `.fetch/failed` (cause + time), `.fetch/notified`
(mtime = last message), `.fetch/log`, `.fetch/tmp.<pid>/`.

`sh fetch-bin.sh hook <Event>` (called by a guard whose binary is missing; `<Event>` is the
literal event name the guard is wired under):
- **One speaking rule governs every branch below.** An invocation speaks only when `<Event>` is
  in the **displaying set** (§II table: SessionStart, PreToolUse, PostToolUse,
  PostToolUseFailure, Stop) and `.fetch/notified` is absent or older than 10 minutes. Speaking
  prints one top-level `{"systemMessage": …}` — on SessionStart with the same text as
  `hookSpecificOutput.additionalContext`, so the agent sees it too — and touches
  `.fetch/notified`. A non-displaying event never prints to stdout and never touches
  `.fetch/notified`, whichever branch it takes. Where `.fetch/` cannot be created, a displaying
  event speaks without a throttle: a root the fetcher cannot write is a state the human must see.
- No readable `version` in `${CLAUDE_PLUGIN_ROOT}/.claude-plugin/plugin.json` → the message
  naming `/prosthetic-conscience:doctor --fix` goes to stderr always and is spoken under the
  rule. Exit 0 (the CI empty-root case).
- `.fetch/failed` younger than 5 minutes → speak its recorded cause and the manual command. Older
  → treated as absent, so the fetch is retried.
- Lock free → take it (`mkdir .fetch/lock`; a lock older than 10 minutes is broken), start
  `fetch-bin.sh fetch` detached (`nohup … &`, or `( … & )` where `nohup` is absent), and speak
  "installing hook binaries for <plugin> <version> in the background".
- Lock held → leave the lock alone and speak the same "installing…" text. A fetch started by a
  non-displaying event is therefore announced by the first displaying event after it.
- Always exit 0; never wait for the network. Stdout is empty or exactly one JSON object.

`sh fetch-bin.sh fetch`:
- Prerequisites first: `command -v curl`, then `sha256sum` or else `shasum`. A missing one fails
  at once with a named cause ("curl not found"; "neither sha256sum nor shasum found") before any
  download.
- Tag `<plugin>--v<version>`, both from `plugin.json`'s top-level fields. (Built without the
  planned cross-check against the cache path: under `--plugin-dir` the parent directory is
  `plugins/`, and the path is not the record.)
- `uname -s` → `linux` | `darwin` | `windows` (MINGW*, MSYS*, CYGWIN*); `uname -m` → `amd64`
  (x86_64, amd64) | `arm64` (arm64, aarch64). Anything else fails with "unsupported platform
  <s>/<m>".
- Download `SHA256SUMS`, then each `tools/cmd/<name>` asset, into `.fetch/tmp.<pid>/`. A
  `tools/cmd` binary absent from `SHA256SUMS` fails loudly, naming it — the FEOV class, caught
  rather than half-installed.
- **All or nothing.** Verify every downloaded digest first; only when all verify, `chmod 0755`
  and `mv` each into `bin/`. One bad digest voids the batch: nothing is moved, and
  `.fetch/failed` names the asset. `bin/` therefore never mixes binaries from a failed fetch
  with binaries from an earlier one. An interruption between two `mv`s leaves verified binaries
  of this one release; the guards of the missing ones fetch again.
- Any failure writes `.fetch/failed`, removes the temp dir, releases the lock. Success removes
  `.fetch/failed`. Log to `.fetch/log`.
- Base URL `https://github.com/ctoforaday/special-circumstances/releases/download`, overridable
  by `SC_RELEASE_BASE_URL` so the contract test and the CI jobs can point it at a local `file://`
  release.

### B. Hook wiring `[MODIFY]`

`plugins/*/hooks/hooks.json`, all 17 guards (prosthetic-conscience 10, gray-area 4, FEOV 3). The
present-binary branch is unchanged; the missing-binary branch becomes:

```sh
F="${CLAUDE_PLUGIN_ROOT}/hooks/fetch-bin.sh"; if [ -f "$F" ]; then exec sh "$F" hook <Event>; fi; echo '<plugin>: hook binaries missing — run /prosthetic-conscience:doctor --fix' >&2
```

The existence check keeps a root without the script (CI's empty root, a damaged cache) on an
exit-0 path that writes to stderr only. Built differently from the round-2 text, which had this
fallback print a `systemMessage` on every event: FEOV's hooks/README.md records that a
SubagentStart/Stop hook that talks re-invokes the seat (nine firings measured), and the fallback
cannot tell its event apart without the script. A root with `hooks.json` but no `fetch-bin.sh`
cannot come from an install; both ship in the same tree. `fetchbingen -check` fails when a guard's `<Event>` differs from the event
key it sits under, or when a guard does not reference `fetch-bin.sh` — the literal is a copy of
the key, and a copy-paste slip would silently move a guard out of the displaying set.

### C. Marketplace pinned to release tags `[MODIFY]` + generator `[NEW]`

- `.claude-plugin/marketplace.json`: prosthetic-conscience, gray-area and FEOV take
  `{"source":"git-subdir","url":"https://github.com/ctoforaday/special-circumstances.git",
  "path":"plugins/<name>","ref":"<name>--v<version>"}`. sleeper-service keeps
  `./plugins/sleeper-service`: it is never tagged.
- `scripts/marketplacegen` `[NEW]`: writes `url`, `path` and `ref` for every plugin that has
  `tools/cmd` (the predicate that makes a plugin taggable), from its `plugin.json` `version`;
  `-check` fails on drift.
- Release flow: the release PR's version bump fails the gate until `marketplacegen` rewrites the
  `ref`, so the bump and the pin land together. Between merge and tag push, client updates fail
  and keep the old version (probed). Between tag push and published assets, the fetcher's
  failure is retried after 5 minutes.
- `CLAUDE.md` Releases section: the release PR carries the regenerated `ref`.

### D. Doctor `[MODIFY]`

- `danceWarnings` (`main.go:384-399`): the EMPTY-BIN predicate stops inferring "no binaries" from
  directory emptiness — already false today because of `.gitkeep`. It becomes: any binary
  `binariesOf(newest)` lists is not built. The text reports `.fetch/failed`'s cause when present,
  otherwise says the hooks are installing the binaries and points at `.fetch/log`.
  `dance_test.go`'s fixture gains the real shape: `bin/` with `.gitkeep` only, with and without
  a `.fetch/` state directory.
- `commands/doctor.md`: "YOU MUST NOT auto-run `--fix` at session start" becomes: hook binaries
  install themselves from the plugin's pinned release; `--fix` is the manual path for when that
  failed; installing an external tool still needs explicit confirmation. Step 2 (bootstrap when
  `sc-doctor` itself is missing, pinned tag since #947) stays: it is the synchronous path for a
  `/doctor` run before the fetch lands.
- `ReferenceCommit`, `fetchRelease` and the rest of `-fix`: unchanged. On a pinned install the
  record's sha is the tag commit, which the release binaries are stamped with (V-R5). Two fetch
  implementations stay by ruling: the curl-based `fetch-bin.sh` is the automatic route; `-fix`'s
  gh-based Go fetch is the manual one a human runs, including when the script reports `curl` or a
  checksum tool missing. Both are pinned to the plugin's tag and both are contract-tested against
  `AssetName`, the name the release job publishes.
- Built after the fourth audit: `plugins/gray-area/bin/.gitkeep` (gray-area shipped no `bin/`, so
  a fresh install had none on PATH at enable), and `fetchbingen -check` requires `bin/.gitkeep`
  in every plugin that ships binaries.
- Package doc `main.go:6-8` says `-fix` builds first and prints fetch instructions otherwise; it
  fetches first and builds as the fallback (`main.go:524-525`). Corrected in the same edit.
- `[NEW] tools/internal/doctor/fetchbin_contract_test.go` beside `release_contract_test.go`, so the
  script is tied to the same `AssetName` the release job is tested against. It drives the
  generated `hooks/fetch-bin.sh` the way a guard does (`sh fetch-bin.sh hook <Event>` with
  `CLAUDE_PLUGIN_ROOT` set, then waits for the detached child) against a `file://` release laid
  out with `AssetName` names and a real `SHA256SUMS`. Cases: install end to end; wrong digest →
  no binary in `bin/`, `.fetch/failed` names it; two assets, one good and one with a wrong digest
  → `bin/` gains neither; asset missing from `SHA256SUMS`; unsupported
  platform (`uname` stubbed on PATH); `curl` absent, and both checksum tools absent (PATH
  stubbed) → `.fetch/failed` holds the named cause; lock older than 10 min → broken; unreadable
  version → stderr always, JSON only on a displaying event; every event name in the §II table —
  displaying ones emit one top-level `systemMessage` JSON object and touch `.fetch/notified`, the
  others print nothing and leave it untouched. The speaking rule is driven on BOTH branches that
  can follow a non-displaying event: (a) it takes the lock, then a displaying event arrives while
  the fetch runs → speaks "installing…" once, and a second displaying event inside 10 minutes is
  silent; (b) it finds `.fetch/failed`, then a displaying event arrives → speaks the cause.

### E. CI `[MODIFY]` + `[NEW]`

- `.github/workflows/hooks.yml` "Hook bootstrap guards degrade gracefully": on an empty plugin
  root every guard exits 0, names `doctor --fix` on stderr, and prints nothing on stdout (the
  guard's own fallback); a second pass on a populated root with
  `SC_RELEASE_BASE_URL` at an empty `file://` directory: each guard exits 0 within 2 s and its
  stdout is empty or one JSON object.
- `[NEW]` job `fetch-bin-platforms`, matrix `ubuntu-latest`, `macos-latest`, `windows-latest`
  (`shell: bash`, which is Git Bash on Windows): builds a fixture plugin root with two
  `tools/cmd` entries and a `file://` release named by the runner's `uname` mapping, runs a real
  guard command from it, asserts the guard returns within 2 s (CI-runner slack on G3's 1 s, which
  V-R1 measures on a real session), then waits up to 60 s for both
  binaries in `bin/` with matching digests after the guard's shell has exited. This proves the
  script's platform mapping, checksum tool choice, `.exe` naming and detach survival per OS. It
  cannot prove how Claude Code itself treats a hook's children on macOS or Windows (R1).
- `scripts/check` gates `fetchbingen`, `marketplacegen` and `fetch-bin-platforms` (the last runs
  the host-platform leg locally); `parity_test.go` enforces the pairing with `hooks.yml`.

### F. Prose carriers `[MODIFY]`

Each states self-heal as the primary path and `doctor --fix` as the manual one:
`README.md:184`; `plugins/prosthetic-conscience/README.md:38`;
`plugins/prosthetic-conscience/hooks/README.md:11-13`; `plugins/gray-area/hooks/README.md:10-12`;
`plugins/frank-exchange-of-views/hooks/README.md:36-38,68-70,88-90`;
`plugins/prosthetic-conscience/requirements.json:67` and `plugins/gray-area/requirements.json:29`
(`_hook_binaries._comment`); `plugins/gray-area/commands/audit-checkpoint.md:26` (bootstrap: the
binary may still be arriving); `docs/setup-script.md:38-50` (the Go build becomes an optional
warm-up; without Go the hooks fetch on first use) and `:166-184`; `CLAUDE.md:71`; code comments
describing the bootstrap window or the guard's missing-binary behaviour —
`doctor/builtfrom.go:5-6`, `doctor/main.go:6-8` (D),
`frank-exchange-of-views/tools/internal/setup/hookprovenance.go:50-52,107,125-126`,
`hookprovenance_test.go:71-72`, `sittinghook/sitting.go:188`, `seatenv/identity.go:94`;
`ideas/backlog.md:36` (checked off).

### Consumer census (run 2026-09-11; saved at `~/.claude/scratch/doctor-bootstrap/census.txt`)

Re-run:
1. `grep -c 'hook binaries missing' plugins/*/hooks/hooks.json`
2. `grep -rln marketplace.json --include=*.go --include=*.sh --include=*.yml --include=*.mjs --include=*.js .`
3. `grep -rn -iE 'doctor (-|--)fix|release asset|go build.*bin/|ships (from git )?without binaries|arrive via' --include=*.json --include=*.md --include=*.yml --include=*.go plugins docs README.md CLAUDE.md scripts .github | grep -v -E '_test\.go|/plans/|/ideas/'`
4. `grep -rn -E '\b(fetchRelease|downloadArgs|describeFetchFailure|verifySHA256|releaseRepo|fixWith|danceWarnings)\b' --include=*.go plugins scripts`
5. `grep -rn -E 'ReadDir|len\(entries\)' plugins/*/tools/internal/doctor/*.go plugins/frank-exchange-of-views/tools/internal/setup/*.go | grep -v _test` and `git ls-files 'plugins/*/bin/*'`
6. `grep -rn -iE 'stderr line|one line (of|on) stderr|degrades? to (one|a single|ONE)|crash[- ]storm is unmissable|guard warning|guard (finds|prints|degrades)' --include=*.go --include=*.md --include=*.json --include=*.yml plugins docs README.md CLAUDE.md scripts .github | grep -v -E '/plans/|/ideas/'`

| Consumer | Changes? |
|---|---|
| `plugins/prosthetic-conscience/hooks/hooks.json` (10 guards) | yes, B |
| `plugins/gray-area/hooks/hooks.json` (4 guards) | yes, B |
| `plugins/frank-exchange-of-views/hooks/hooks.json` (3 guards) | yes, B — same concept; no FEOV behaviour work |
| `.github/workflows/hooks.yml:1354-1380` guard gate | yes, E |
| `scripts/check` gate list + `parity_test.go` | yes, E |
| `scripts/bootstrap-plugins.sh` (adds the marketplace by repo; reads no source) | no |
| `scripts/validatejson` (JSON validity only) | no; must pass on the new shape |
| `scripts/pluginparity` (reads only `name`, `main.go:42`; magic bytes exclude `#!`) | no |
| `commands/doctor.md` steps 2 and 3, closing line | yes, D (step 2 text unchanged) |
| `doctor/main.go` `danceWarnings` predicate + text, `dance_test.go`, `run_test.go:313` | yes, D |
| `doctor/main.go:74` `binariesOf`, `:105`, `:332`, `:354` (list `tools/cmd`, marketplace and version dirs — not "is bin empty") | no |
| FEOV `setup/hookprovenance.go:110` `binariesIn` (lists every file in `bin/`) | no — `.fetch/` keeps the fetcher's files out of `bin/`; comment at `:107` is F |
| FEOV `setup/setup.go:606,624` (`*.md` listings) | no |
| `doctor/main.go` `fix`/`fixWith`/`fetchRelease`/`downloadArgs`/`verifySHA256`/`describeFetchFailure` and their tests | no |
| `doctor/builtfrom.go:144`, `builtfrom_test.go` ("run doctor --fix to rebuild") | no |
| FEOV `setup.go:544,556,567`, `setup_test.go:218` (remedies naming the manual path) | no |
| `toolchainnudge`, `qualitygate`, `requirements.json:17` (external tools and Go's role in `-fix`) | no |
| `pluginparity/main.go:321,360`, `versionguard/main.go:154,193`, `buildid.go:33-35` ×4, `buildidgen/main.go:7` | no — still true |
| Prose and comments listed in F | yes, F |
| `hooks.yml:1378,1380` ("guard warning" in the gate) | yes, E |
| `doctor/main.go:396` ("hooks degrade to guard warnings") | yes, D |
| FEOV `record/tiersubstitution_test.go:83` (a gate's stderr), gray-area `telecli/find.go:114` (a read verb's stderr) | no — other messages |
| gray-area `cmd/gray-area-capture/catalogue.go:86` (capture failures go to one stderr line) | no here — same silence class, carried by the tracked stderr issue |
| `plans/*` historical mentions | no |

### Tracked beyond this plan

- Hook messages on stderr at exit 0 reach nobody, and the guard is not the only one: gray-area's
  `catalogue/open.go:309` ("rebuilt empty — run `telepathy backfill`") went unseen this session
  too. One issue: census every hook binary's stderr-on-exit-0 output and move the actionable
  ones to `systemMessage` on displaying events.
- After the next prosthetic-conscience and gray-area release, repeat V-R1 through the pinned
  marketplace (the tags pinned today predate `fetch-bin.sh`).
- FEOV's own preflight (`setup.go:544,556,567`) still sends a missing `feov-record` to
  `doctor --fix`. A FEOV hook that fires installs `feov-record` with its siblings, but a run that
  fires no hook first is not healed. Left for v2, per the ruling.

## IV. Risk & Mitigation

| # | Risk | Likelihood | Impact | Mitigation cost | Mitigated by |
|---|---|---|---|---|---|
| R1 | A hook's detached child is killed, so the fetch never completes | low on Linux (S1); Claude Code's handling on macOS/Windows unmeasured | high | medium | S1; `fetch-bin-platforms` proves the script-level detach on all three OSes (E); a killed fetch leaves a stale lock that is broken after 10 min and retried, and a failure is reported; the Claude-Code-level residue on macOS/Windows is stated |
| R2 | `systemMessage` is discarded — the same silence this plan exists to end | low (S2, S3) | high | low | Top-level shape; event-aware emission (A); contract-test event matrix (D); V-R1; V-R4 |
| R3 | Each release has a window where the tag exists but assets do not | high | low | low | Failure recorded, retried after 5 min, message says so (A) |
| R4 | `plugin.json`'s version, read by pattern, misses after a shape change and folds into a guess | low | medium | low | A miss is a loud failure, never a default; contract test |
| R5 | Concurrent hooks race the download or leave a stale lock | high | medium | low | `mkdir` lock, 10-min break, per-pid temp dir, atomic `mv`; contract test |
| R6 | A partial or tampered file becomes executable | low | high | low | Digest checked before `chmod`/`mv`; wrong-digest case |
| R7 | The secret guard is off until binaries arrive | certain | medium | — | Ruled accepted: the window shrinks from "until a human notices" to the download time |
| R8 | Dev sessions in this repo get released content, not main | certain | low | — | Ruled acceptable; smoke runs refresh the cache by hand (`~/scratch/install-b4-cache.sh`) or use `--plugin-dir` |
| R9 | A release PR forgets the `ref` | medium | high | low | `marketplacegen -check` in `check` and CI (C) |
| R10 | Windows without Git Bash runs hooks in PowerShell | low | high | — | Unchanged from today: the current guards already need Git Bash |
| R11 | A guard's event literal drifts from its key, moving it in or out of the displaying set | medium | medium | low | `fetchbingen -check` (B) |
| R12 | SessionEnd, PreCompact, PostCompact and FileChanged may display after all | unknown | low | — | Treated as non-displaying: they still start the fetch; the message comes from the next displaying event |
| R13 | Two displaying events in the same instant both pass the `.fetch/notified` check and the human sees the message twice | low | low | — | Accepted: a duplicate is cosmetic, and G2 bounds the silence, not the count |

## V. Verification Plan

Re-armed by any change to `scripts/fetchbingen/fetch-bin.sh`, a generated `fetch-bin.sh`, any
`hooks.json`, `.claude-plugin/marketplace.json`, a `plugin.json` version, or
`tools/internal/doctor/`.

**Automated.**
1. `go -C scripts run ./check` — every local gate, including `fetchbingen`, `marketplacegen` and
   the host leg of `fetch-bin-platforms`, tests at `-count=1`.
2. `go -C plugins/prosthetic-conscience/tools test -count=1 ./internal/doctor/` — the fetch-bin
   contract test and the updated dance tests.
3. The `hooks.yml` guard gate's shell block, run locally against an empty root and against a
   populated root with `SC_RELEASE_BASE_URL=file:///nonexistent`.
4. Delete-the-row checks: invert the digest comparison in `fetch-bin.sh` → 2 fails; drop
   `PostToolUse` from the displaying set → 2 fails; make the held-lock branch silent → 2 fails;
   delete the `curl` precheck → 2 fails; move each asset as soon as it verifies → 2's mixed-batch
   case fails; change one guard's `<Event>` literal →
   `fetchbingen -check` fails; restore each.
5. CI: `fetch-bin-platforms` green on all three runners.

**Driveable checks on real data** (isolated `CLAUDE_CONFIG_DIR` under
`~/.claude/scratch/selfheal/`, real GitHub release assets):
- **V-R1 fresh install.** A local directory marketplace whose relative sources are this branch's
  prosthetic-conscience and gray-area (versions 0.44.0 / 0.11.0, so the real releases are
  fetched). `claude -p` a prompt that runs one Bash command. Observe: a `hook_system_message`
  attachment; the guard's `durationMs` < 1000; after the session, `bin/` holds every
  `tools/cmd` binary and each `sha256sum` equals its release `SHA256SUMS` line; no `.fetch/`
  file inside `bin/`; a second `claude -p` shows the hooks running their binaries (`hook_success`
  without the guard message).
- **V-R2 update.** Remove the binaries from that version directory's `bin/` (leaving `.gitkeep`,
  the state a new version directory is created in) and repeat; same observations; `sc-doctor`
  from that cache reports the EMPTY-BIN state while `.fetch/lock` is held.
- **V-R3 failure.** `SC_RELEASE_BASE_URL=https://example.invalid` for one run: the tool call
  proceeds, `bin/` gains no binary, `.fetch/failed` names the cause, the displayed message names
  `/prosthetic-conscience:doctor --fix`, and `sc-doctor` repeats the cause.
- **V-R4 human.** gblock runs one interactive session against V-R1's setup and confirms the
  message appeared in the terminal.
- **V-R5 pin.** Install this branch's `marketplace.json` in a fresh isolated config; for each of
  the three tagged plugins, `diff -r` the cache directory against `git archive <tag>` of its
  subdirectory (only `.in_use` may differ) and compare the record's `gitCommitSha` with
  `git rev-parse <tag>^{commit}`; after self-heal, that cache's `sc-doctor` reports no "stale".

**Gate.** `/plan-audit` on this file before any code; `rulesweep` needs `Rule-Class:` and
`Sibling-Sweep:` trailers on the commit touching `commands/doctor.md` and `hooks.json`.
`FEOV_RELEASE_GATE` is not re-armed: no record, flag or enum changes.
