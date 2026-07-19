# Tool resolution: discover once, shim where PATH already reaches, never rewrite

## I. Summary & Goals

A tool that is installed and working is reported missing because the process checking it
resolves only through `PATH`. This has cost four separate defects in a single day:

| symptom | reality |
|---|---|
| `feov-record ✗ (required)` after a fully successful `sc-doctor --fix` | installed into the plugin's `bin/`, checked on `PATH` |
| `qlty ✗` → verdict pinned at DEGRADED | installed at `~/.qlty/bin` |
| `jq ✗` | installed via winget |
| "`-race` blocked, no C compiler" | MinGW-W64 16.1.0 installed, eight levels deep under `AppData` |

The last one is the instructive one: a subagent searched `C:\` to depth 5, found nothing,
reported the compiler absent, and the lead **believed it and repeated it to another
subagent**, which then skipped work on that basis. A wrong environment fact propagates
exactly like a wrong citation.

**Goals.**

1. Discovery happens ONCE, in one place, and is written down.
2. Our own code never again equates "not on `PATH`" with "not installed".
3. A caller that names an unresolvable tool is TOLD where it actually is.
4. The trajectory record continues to say exactly what ran.

**Non-goal (explicitly rejected — see §II).** Rewriting the operator's or the agent's
commands to fully-qualify binary paths.

## II. Technical Context & Design

### What the harness actually allows

Checked against the documentation and then verified on this machine, because the
documented answer and the observed answer disagreed:

| mechanism | status |
|---|---|
| `PreToolUse` returning `updatedInput` to rewrite a Bash command | Documented and supported |
| A plugin shipping `env` in `settings.json` | **Not supported** — plugin settings accept only `agent` and `subagentStatusLine` |
| A plugin's `bin/` being added to the Bash tool's `PATH` | Documented — **but NOT observed here.** No plugin cache path appears in `$PATH`, and no shipped binary (`sc-doctor`, `sc-push-freeze-guard`, `feov-record`) resolves by bare name, despite all three plugins having populated `bin/` directories |
| `SessionStart` hook writing `CLAUDE_ENV_FILE` | Documented; runs as a shell preamble before Bash commands |

The `bin/` discrepancy is unexplained and MUST be treated as unavailable until someone
reproduces it working. It may require a manifest declaration we do not make, or a newer
CLI. **This plan therefore assumes `PATH` is not mutable by us.**

### Why rewriting is rejected

`PreToolUse` *can* rewrite commands. It should not.

1. **It breaks the record.** The transcript would say one thing while another executed.
   This project has spent a full cycle removing self-report from the evidence chain, and
   the trajectory corpus is the substrate Gray Area is meant to mine for deception. A
   layer that silently alters commands makes "what did this agent actually do"
   unanswerable — and it would be *our* layer doing the lying.
2. **Shell text surgery has no safe bottom.** Quoting, heredocs, pipelines, `VAR=x cmd`,
   subshells, aliases, and `git -C` — the last of which just cost us a blind push-freeze
   guard for exactly this reason. Any rewriter is a parser; an incomplete parser that
   usually works is the worst kind, because it fails rarely and silently.
3. **Documented fragility.** Where several `PreToolUse` hooks return `updatedInput` for
   one tool, the last to finish wins and hook order is non-deterministic. A rewriter is
   therefore not composable with any future hook that touches Bash.
4. **Its failure mode is a silent success.** Qualify the wrong binary and everything
   looks fine.

### The mechanism that actually works: shims in a directory already on PATH

**Operator proposal, verified 2026-07-19 and adopted as primary.** We cannot put a
directory on `PATH`, but one is already there:

| shell | `C:\Users\gbloc\bin` on PATH? |
|---|---|
| Git Bash (the Bash tool) | yes — FIRST entry |
| Windows native (PowerShell tool, and any Go hook Claude Code spawns) | yes |

It is on both, so a shim placed there resolves for every consumer: the Bash tool, the
PowerShell tool, hook binaries, and their child processes. The directory does not exist
yet; creating it is sufficient, because both shells already carry it.

This removes the need for a `PreToolUse` hook entirely. No command is inspected, no
command is rewritten, and the trajectory record is untouched by construction — the shim is
invisible to everything except name resolution, which is exactly the layer where the
defect lives.

`sc-doctor --fix` therefore: discovers real locations, creates the user bin directory if
absent, and writes a thin shim per unresolvable tool (a `.cmd` wrapper on Windows —
symlinks need developer mode or elevation; a symlink on POSIX).

**Constraints this carries, none of them fatal:**

- It writes OUTSIDE the working tree, into the user's home. Per `agent-guardrails` that
  needs explicit human approval, which `sc-doctor`'s existing `--fix` confirmation already
  models. It MUST NOT happen on a bare `sc-doctor` run.
- It is user-global, not project-scoped: these shims affect everything the user runs, not
  only Claude. That is a feature for the operator and a responsibility for us.
- **Shadowing.** A shim named `gcc` would outrank a real `gcc` installed later, since the
  user bin directory is FIRST on PATH. So: only ever create a shim for a tool that does
  not already resolve, and give `sc-doctor` a way to list and remove what it created.
- Shims are machine-specific and regenerable; they are never committed.

### Defense in depth: a resolver and a manifest

Kept, at reduced scope. The shims fix name resolution for everything downstream; the
resolver fixes OUR OWN checks, which is where three of the four failures actually lived —
`sc-doctor` reporting a present tool as missing is a defect even when the shim would have
saved the caller. A check that can only see `PATH` is the root cause, and shims treat the
symptom.

### The rejected design: a resolver, a manifest, and an advisory

**Resolution order**, used by every consumer: `toolpaths.json` → `PATH` → known install
roots → not found.

`sc-doctor` owns discovery and writes `toolpaths.json` (schema-versioned, machine-local,
never committed). Known roots include winget's package directory, `~/.qlty/bin`,
chocolatey, scoop, and msys/mingw — searched by *known location*, not by an unbounded
filesystem sweep, since the sweep is what failed.

Two consumers:

- **Our own binaries and scripts** resolve through the manifest. This alone closes the
  entire class for `sc-doctor`'s own checks and for any hook shelling out to a tool.
- **A `PreToolUse` advisory hook** for Bash: when a command names a tool that is not
  resolvable but IS in the manifest, it returns `permissionDecision: "ask"` (or a warning)
  whose reason carries the qualified path. The agent reissues the command itself, so the
  command that runs is the command that was recorded.

Advise-don't-rewrite keeps the record honest, needs no shell parser (a token scan is
enough to *name a suspect*, and being wrong costs one advisory line rather than a
corrupted command), and composes with other hooks because it changes nothing.

### Staleness

An index that answers confidently from stale data is a worse failure than no index —
already learned twice here, from qmd's cache and from the golden runner reporting
"recorded" while Go's test cache meant it had written nothing. Every manifest entry
carries the resolved path and the version string observed at discovery. A consumer that
finds a manifest path missing from disk MUST fall through to live discovery and mark the
manifest stale, never silently trust it.

## III. Implementation plan (file-by-file)

1. `plugins/prosthetic-conscience/tools/internal/toolpath/` — new package: `Resolve(name)`,
   `Discover()`, `Load()`, `Save()`. Known-roots table lives here, platform-tagged.
2. `plugins/prosthetic-conscience/tools/cmd/sc-doctor/` — call `Discover()`; write
   `toolpaths.json`; report a found-but-off-PATH tool as **`present at <path>, not on
   PATH`** rather than as missing, with the PATH command as the remedy.
3. `plugins/prosthetic-conscience/tools/internal/shim/` — new package: create/list/remove
   shims in the user bin directory. `.cmd` wrapper on Windows, symlink on POSIX. Refuses
   to shim a tool that already resolves (shadowing guard).
4. `sc-doctor --fix` wiring: create the user bin dir if absent, shim each
   present-but-unresolvable tool, and PRINT what it created. No hook is registered — the
   advisory design was dropped once the user-bin PATH entry was verified.
5. `plugins/frank-exchange-of-views/requirements.json` — audit for the same defect that
   put a plugin-shipped binary in an external-tool manifest (already fixed once).
6. `plugins/prosthetic-conscience/.claude-plugin/plugin.json` — version bump.

## IV. Risk & Mitigation (likelihood x impact x complexity-to-mitigate)

| risk | l x i | mitigation |
|---|---|---|
| Advisory fires on a tool that resolves fine (false positive noise) | med x low | Only advise when resolution actually fails; silent otherwise. Noise is self-limiting because a resolvable tool never triggers it |
| Manifest goes stale after an uninstall/upgrade | med x med | Entries carry path + version; a missing path forces live rediscovery and marks the manifest stale rather than reporting the old answer |
| Token scan misidentifies the executable (`VAR=x cmd`, pipelines) | high x **low** | Being wrong costs one advisory line. This is precisely why the design advises rather than rewrites — the same error under a rewriter corrupts the command |
| Discovery cost on every session | low x low | Discovery runs in `sc-doctor`, not per-command. The hook only reads a small JSON file |
| A future hook also targets Bash `updatedInput` | low x med | We register no such hook at all |
| A shim shadows a real tool installed later (user bin is FIRST on PATH) | med x **high** | Only shim what does not already resolve; `sc-doctor` can list and remove its own shims; each shim is stamped with a marker comment identifying it as generated |
| Shims write into the user's home, outside the working tree | high x med | Gated behind `--fix` and its existing explicit confirmation; never on a bare run. Reported line by line so the operator sees every file created |
| Writing a machine-local file into the plugin cache | med x low | Cache dir already receives `--fix`-installed binaries; the file is regenerable and never committed |

## V. Verification plan

1. `go test -count=1 -cover ./...` in `plugins/prosthetic-conscience/tools` — the new
   `toolpath` package covered including: manifest hit, PATH hit, known-root hit, not-found,
   and a stale entry whose recorded path no longer exists.
2. `go test -race -count=1 ./...` — with `gcc` on PATH from the location recorded in
   session memory. This now runs; a prior report that it could not was wrong.
3. **The four historical failures become fixtures.** `gcc`, `qlty`, `jq`, and `feov-record`
   must each resolve through the resolver on this machine, asserted against their real
   locations. That is the honest regression test: these are not hypotheticals, they are
   the four cases that actually shipped broken.
4. Shim behaviour, verified as real processes in BOTH shells: after `--fix`, `qlty` and
   `jq` must resolve by bare name in the Bash tool AND in the PowerShell tool. A tool that
   already resolves must NOT be shimmed. `sc-doctor` must be able to list and remove what
   it created.
5. `sc-doctor` run end-to-end on this machine must report `qlty` and `jq` as
   present-but-off-PATH, and must NOT report them missing. Verified against the binary's
   real output, not a unit test's idea of it.
6. Confirm the transcript record is unchanged: the executed command string for a Bash call
   under the advisory hook is byte-identical to the one issued.
