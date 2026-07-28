#!/usr/bin/env bash
# Install the Special Circumstances plugins into ~/.claude and build their hook
# binaries. Idempotent — safe to re-run.
#
# Intended for the setup script of a Claude Code cloud environment, which is the
# only point that runs BEFORE a session resolves its plugins. A SessionStart hook
# is too late: plugin components are already loaded by the time it fires, and a
# fresh-container-per-session environment never reaches a "next session" that
# would see the install.
#
# The marketplace repository is private. Cloning it needs credentials that reach
# GitHub from wherever this runs — in a cloud environment that means the repo
# must be in the session's scope. If it is not, MARKETPLACE_SOURCE can point at
# an already-checked-out copy instead (see below).
#
# Env:
#   MARKETPLACE_SOURCE  what to hand `claude plugin marketplace add`
#                       default: ctoforaday/special-circumstances (GitHub)
#                       may instead be a local path to a checkout, which the CLI
#                       accepts as a `directory` source
#   SKIP_BUILD=1        install plugins but do not build the Go hook binaries

set -euo pipefail

MARKETPLACE_NAME=special-circumstances
MARKETPLACE_SOURCE="${MARKETPLACE_SOURCE:-ctoforaday/special-circumstances}"
PLUGINS=(prosthetic-conscience frank-exchange-of-views sleeper-service)

log() { printf '[bootstrap-plugins] %s\n' "$*"; }
die() { log "ERROR: $*" >&2; exit 1; }

command -v claude >/dev/null || die "claude CLI not on PATH"

# --- marketplace ------------------------------------------------------------
# `marketplace add` errors when the name is already known, so check first rather
# than swallowing the failure — a genuine add failure must not be silent.
if claude plugin marketplace list --json 2>/dev/null | grep -q "\"$MARKETPLACE_NAME\""; then
  log "marketplace $MARKETPLACE_NAME already known"
else
  log "adding marketplace from $MARKETPLACE_SOURCE"
  claude plugin marketplace add "$MARKETPLACE_SOURCE" \
    || die "could not add marketplace from $MARKETPLACE_SOURCE (private repo — are credentials in scope here?)"
fi

# --- plugins ----------------------------------------------------------------
installed="$(claude plugin list 2>/dev/null || true)"
for p in "${PLUGINS[@]}"; do
  if printf '%s' "$installed" | grep -q "$p@$MARKETPLACE_NAME"; then
    log "$p already installed"
  else
    log "installing $p"
    claude plugin install "$p@$MARKETPLACE_NAME" || die "install failed: $p"
  fi
done

# --- hook binaries ----------------------------------------------------------
# Plugin content is COPIED into a versioned cache at install time, so the
# binaries must be built there — ${CLAUDE_PLUGIN_ROOT} resolves to the cache
# directory, never to a repository checkout.
[ "${SKIP_BUILD:-}" = "1" ] && { log "SKIP_BUILD=1, done"; exit 0; }

if ! command -v go >/dev/null; then
  log "go not on PATH — skipping hook binaries."
  log "Hooks degrade to a single stderr line until built; run /prosthetic-conscience:doctor --fix in-session."
  exit 0
fi

cache_root="${CLAUDE_CODE_PLUGIN_CACHE_DIR:-$HOME/.claude/plugins}/cache/$MARKETPLACE_NAME"
[ -d "$cache_root" ] || die "plugin cache not found at $cache_root"

built=0
for plugin_dir in "$cache_root"/*/*/; do
  tools="$plugin_dir/tools"
  [ -d "$tools/cmd" ] || continue
  mkdir -p "$plugin_dir/bin"
  for cmd_dir in "$tools"/cmd/*/; do
    name="$(basename "$cmd_dir")"
    if go build -C "$tools" -o "$plugin_dir/bin/$name" "./cmd/$name"; then
      built=$((built + 1))
    else
      log "WARNING: build failed: $name"
    fi
  done
done
log "built $built hook binaries"

# --- verify -----------------------------------------------------------------
# The suite ships its own preflight; trust it over this script's bookkeeping.
doctor="$(find "$cache_root" -name sc-doctor -type f -perm -u+x 2>/dev/null | head -1)"
if [ -n "$doctor" ]; then
  CLAUDE_PLUGIN_ROOT="$(dirname "$(dirname "$doctor")")" "$doctor" || true
fi
