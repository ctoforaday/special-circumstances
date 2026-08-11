#!/usr/bin/env bash
# Install the Special Circumstances plugins into ~/.claude and build their hook
# binaries. Idempotent — safe to re-run.
#
# Intended for the Setup script field of a Claude Code cloud environment, which
# is the only point that runs BEFORE a session resolves its plugins. A
# SessionStart hook is too late: it fires after Claude Code launches, by which
# point plugin components are already loaded.
#
# THIS SCRIPT ALWAYS EXITS 0, DELIBERATELY. A setup script that exits non-zero
# makes the whole session fail to start. Missing plugins are an inconvenience; a
# session that will not boot is not. Every failure below degrades to a warning.
#
# The marketplace repository is private, so the clone needs credentials that
# reach GitHub from wherever this runs — in a cloud environment, that means the
# repository must be in the session's scope. Where it is not, this script warns
# and exits 0, leaving the session to start without the plugins.
#
# Env:
#   MARKETPLACE_SOURCE  what to hand `claude plugin marketplace add`
#                       default: ctoforaday/special-circumstances (GitHub)
#                       may instead be a local path to a checkout, which the CLI
#                       accepts as a `directory` source
#   SKIP_BUILD=1        install plugins but do not build the Go hook binaries

set -uo pipefail   # NO -e: a failed step must warn, never abort the session

MARKETPLACE_NAME=special-circumstances
MARKETPLACE_SOURCE="${MARKETPLACE_SOURCE:-ctoforaday/special-circumstances}"
# Every plugin in .claude-plugin/marketplace.json. A plugin missing from this list
# is silently never installed, which looks exactly like a plugin that does not
# exist yet — so this list and the marketplace manifest must move together.
PLUGINS=(prosthetic-conscience frank-exchange-of-views sleeper-service gray-area)

log()  { printf '[bootstrap-plugins] %s\n' "$*"; }
warn() { printf '[bootstrap-plugins] WARNING: %s\n' "$*" >&2; }
# Bail out of the script without failing the session.
give_up() { warn "$*"; log "continuing without plugins"; exit 0; }

command -v claude >/dev/null || give_up "claude CLI not on PATH"

# --- marketplace ------------------------------------------------------------
# `marketplace add` errors when the name is already known, so check first rather
# than swallowing the failure — a genuine add failure must still be reported.
if claude plugin marketplace list --json 2>/dev/null | grep -q "\"$MARKETPLACE_NAME\""; then
  log "marketplace $MARKETPLACE_NAME already known"
else
  log "adding marketplace from $MARKETPLACE_SOURCE"
  claude plugin marketplace add "$MARKETPLACE_SOURCE" \
    || give_up "could not add marketplace from $MARKETPLACE_SOURCE — the repository is private, so check it is in this environment's GitHub scope"
fi

# --- plugins ----------------------------------------------------------------
installed="$(claude plugin list 2>/dev/null)"
failed=0
for p in "${PLUGINS[@]}"; do
  if printf '%s' "$installed" | grep -q "$p@$MARKETPLACE_NAME"; then
    log "$p already installed"
  else
    log "installing $p"
    claude plugin install "$p@$MARKETPLACE_NAME" || { warn "install failed: $p"; failed=$((failed + 1)); }
  fi
done
[ "$failed" -gt 0 ] && warn "$failed plugin(s) did not install"

# --- hook binaries ----------------------------------------------------------
# Plugin content is COPIED into a versioned cache at install time, so the
# binaries must be built there — ${CLAUDE_PLUGIN_ROOT} resolves to the cache
# directory, never to a repository checkout.
[ "${SKIP_BUILD:-}" = "1" ] && { log "SKIP_BUILD=1, done"; exit 0; }

if ! command -v go >/dev/null; then
  log "go not on PATH — skipping hook binaries."
  log "Hooks degrade to one stderr line until built; run /prosthetic-conscience:doctor --fix in-session."
  exit 0
fi

cache_root="${CLAUDE_CODE_PLUGIN_CACHE_DIR:-$HOME/.claude/plugins}/cache/$MARKETPLACE_NAME"
[ -d "$cache_root" ] || give_up "plugin cache not found at $cache_root"

# Warm the module cache once, serially. Left to the parallel loop below, every
# build races to download the same toolchain and dependencies, which is the part
# most likely to overrun the setup script's ~5 minute budget.
first_tools="$(find "$cache_root" -maxdepth 3 -type d -name tools -print -quit 2>/dev/null)"
[ -n "$first_tools" ] && { log "warming Go module cache"; go build -C "$first_tools" ./... >/dev/null 2>&1; }

log "building hook binaries"
for plugin_dir in "$cache_root"/*/*/; do
  tools="$plugin_dir/tools"
  [ -d "$tools/cmd" ] || continue
  mkdir -p "$plugin_dir/bin"
  for cmd_dir in "$tools"/cmd/*/; do
    name="$(basename "$cmd_dir")"
    # A development harness that Go's internal-package rule forces to live here. The marker
    # is the single fact; scripts/pluginparity reads the same one for its count.
    [ -f "$cmd_dir/DEV-ONLY" ] && continue
    go build -C "$tools" -o "$plugin_dir/bin/$name" "./cmd/$name" \
      || warn "build failed: $name" &
  done
done
wait

built="$(find "$cache_root" -type f -perm -u+x -path '*/bin/*' 2>/dev/null | wc -l)"
log "built $built hook binaries"

# --- verify -----------------------------------------------------------------
# The suite ships its own preflight; trust it over this script's bookkeeping.
# It exits non-zero on a BLOCKED verdict, which must not fail the session.
doctor="$(find "$cache_root" -name sc-doctor -type f -perm -u+x 2>/dev/null | head -1)"
if [ -n "$doctor" ]; then
  CLAUDE_PLUGIN_ROOT="$(dirname "$(dirname "$doctor")")" "$doctor" || true
fi

exit 0
