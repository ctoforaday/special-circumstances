# Setup script for Claude Code on the web

Cloud sessions start from a fresh container, so a plugin installed during a
session is gone the moment that container is reclaimed. The **Setup script**
field on a cloud environment is the only hook that runs *before* Claude Code
launches, which is the only point early enough for a plugin install to be
visible to the session that follows.

Anthropic snapshots the filesystem after the setup script completes and reuses
that snapshot for later sessions in the same environment, so the install
survives without re-running. The script runs again when you change it or the
environment's allowed hosts, and after the cache expires (roughly seven days).

## Where it goes

Open the environment settings dialog and paste the script into the **Setup
script** field. This is web-UI only — the Claude mobile app monitors and steers
sessions but does not edit environment configuration, so this is a task for a
browser. Each environment is configured separately.

## The script

```bash
#!/bin/bash
# Special Circumstances plugins. Never exits non-zero: a failing setup script
# blocks the session from starting, and missing plugins are the lesser problem.
M=special-circumstances
C="${CLAUDE_CODE_PLUGIN_CACHE_DIR:-$HOME/.claude/plugins}/cache/$M"

claude plugin marketplace list --json 2>/dev/null | grep -q "\"$M\"" \
  || claude plugin marketplace add ctoforaday/special-circumstances \
  || { echo "SC: marketplace clone failed — check repository access and the environment's network access level"; exit 0; }

for p in prosthetic-conscience frank-exchange-of-views sleeper-service gray-area; do
  claude plugin install "$p@$M" || echo "SC: install failed: $p"
done

command -v go >/dev/null || { echo "SC: no go, hook binaries unbuilt"; exit 0; }
T=$(find "$C" -maxdepth 3 -type d -name tools -print -quit 2>/dev/null)
[ -n "$T" ] && go build -C "$T" ./... >/dev/null 2>&1   # warm module cache before parallel builds
for d in "$C"/*/*/; do
  [ -d "$d/tools/cmd" ] || continue
  mkdir -p "$d/bin"
  for c in "$d"tools/cmd/*/; do
    n=$(basename "$c"); go build -C "$d/tools" -o "$d/bin/$n" "./cmd/$n" || echo "SC: build failed: $n" &
  done
done
wait
exit 0
```

`scripts/bootstrap-plugins.sh` in this repository is the same logic with fuller
comments and diagnostics. Paste the script above rather than calling that file:
the setup script runs before Claude Code launches, the snapshot is per
environment rather than per repository, and in any environment not scoped to
this repository the checkout is not there to call.

## Why it never exits non-zero

A setup script that exits non-zero makes the whole session fail to start. Every
failure path above degrades to a warning and `exit 0`. A session that comes up
without the plugins is a nuisance; an environment that will not boot is not.

## Repository visibility decides where this works

This repository is private. The clone needs GitHub credentials that reach it,
which in a cloud environment means the repository must be in that session's
scope. Everywhere else the clone fails, the script warns, and the session starts
without the plugins.

Were the repository public, the clone would need no credentials and the script
would work in any environment at Trusted network access. Verified indirectly:
unattached public repositories (`octocat/Hello-World`, `cli/cli`) clone through
the session proxy without credentials. The documented restriction limiting
GitHub requests to session-attached repositories covers API and release-asset
requests, not `git clone` of a public repository.

## Requirements

| Requirement | Detail |
|---|---|
| Network access | Trusted (the default) or higher. **None** blocks the clone and the Go module fetches. |
| Allowed domains | `github.com`, `codeload.github.com`, `proxy.golang.org`, `sum.golang.org` — all on the default Trusted list |
| Go toolchain | Pre-installed in cloud environments. `go.mod` declares 1.24; the toolchain resolves 1.25.0 on first build |
| Runtime budget | Keep the setup script under roughly five minutes so the environment cache can build |

## Verification

Run against a throwaway plugin cache. `CLAUDE_CODE_PLUGIN_CACHE_DIR` redirects
the copied plugin content and the marketplace registry:

```bash
tmp=$(mktemp -d)
CLAUDE_CODE_PLUGIN_CACHE_DIR="$tmp" bash ./scripts/bootstrap-plugins.sh
```

| Path | Result |
|---|---|
| Cold, empty cache | exit 0, ~30s, every plugin's hook binaries built — one per directory under `plugins/*/tools/cmd/`, 12 at the time of writing — and `sc-doctor` verdict DEGRADED |
| Unreachable marketplace | **exit 0**, warns, session startup unaffected |
| Warm, already installed | exit 0, skips marketplace and plugin installs, rebuilds binaries |

The DEGRADED verdict is `qlty` and `qmd` missing, both recommended rather than
required. `gh` is *not* part of it: it is absent by design in cloud sessions,
which reach GitHub through MCP tools, and the manifest declares that with
`not_applicable_in` so the doctor reports it as n/a rather than as a defect
with install advice nobody can act on.

**The isolation is partial, and the part it misses is the part that matters.**
`CLAUDE_CODE_PLUGIN_CACHE_DIR` does not cover `~/.claude/settings.json`, so a
throwaway run still writes the real `extraKnownMarketplaces` entry *and* enables
all four plugins in `enabledPlugins`. What it leaves behind is a session
configured for plugins whose content lives in a temp directory that is about to
be deleted. Re-run the script without the variable afterwards to repopulate the
real cache, or hand-remove both keys from `~/.claude/settings.json`.

## Building the hook binaries by hand

`/prosthetic-conscience:doctor --fix` fetches CI-built release assets and falls
back to a from-source build. Where neither reaches — no release asset for the
platform, or no network — build directly. Plugin content is **copied** into a
versioned cache at install time, so `${CLAUDE_PLUGIN_ROOT}` resolves to that
cache and never to a repository checkout. The binaries must land in the cache:

```bash
root="$HOME/.claude/plugins/cache/special-circumstances/prosthetic-conscience/<version>"
go build -C "$root/tools" -o "$root/bin/sc-doctor" ./cmd/sc-doctor
```

Repeat per command in `tools/cmd/`, and again for
`frank-exchange-of-views`'s `feov-record`. Go is the only prerequisite; `-race`
in the test suite additionally needs CGO.

Until the binaries exist, each hook degrades to a single line on stderr pointing
at `/prosthetic-conscience:doctor --fix` rather than failing the tool call.
