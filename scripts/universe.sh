#!/usr/bin/env bash
# universe.sh — build a self-contained Claude Code environment from THIS checkout and run the
# research engine inside it, as one `claude -p` process.
#
# WHY THIS EXISTS, AND WHAT IT REPLACES.
#
# The engine was being exercised by a scratch harness that reimplemented the Workflow tool's
# globals over a pool of `claude -p` processes, one per seat. It measured a different system, and
# every difference cost a day:
#
#   - Seats spawned as TOP-LEVEL processes never fire SubagentStart/SubagentStop, so no sitting
#     span was ever written and `gray-area` capture had no transcript to bind. The per-sitting
#     tool-call cap survived only because sittingcap keys on the register event, which the author
#     chose for exactly this reason.
#   - The harness pasted each seat's JSON schema into the prompt and then accepted any parseable
#     object. The real harness enforces it. An arm died twelve sittings in on a relayed plan
#     missing a field the tool always writes, and the defect was in the instrument.
#   - `parallel()` swallowed a rejected seat into null, which the engine reads as "this lens filed
#     nothing" — identical to a lens that audited and found nothing.
#
# A rig that is easier than the thing it stands for does not give a conservative reading. It runs a
# different experiment. This script removes the rig: the engine is invoked the way a user invokes
# it, and everything the plugins install — hooks, binaries, subagent dispatch — is really there.
#
# TWO UNIVERSES, AND THE DEFAULT IS THE ISOLATED ONE.
#
#   isolated (default)  CLAUDE_CONFIG_DIR and CLAUDE_CODE_PLUGIN_CACHE_DIR point inside WORKDIR,
#                       so nothing touches ~/.claude. Other sessions on this box are unaffected,
#                       and the universe is deletable in one `rm -rf`.
#   --live              install into ~/.claude instead. Use when you want the change in YOUR next
#                       session. It is shared state: other sessions on this box see it.
#
# The plugin content is COPIED into a versioned cache at install time and ${CLAUDE_PLUGIN_ROOT}
# resolves there, never to a checkout — so the binaries are built in the cache, after install.
# That ordering is bootstrap-plugins.sh's and is kept here.
#
# Usage:
#   scripts/universe.sh build                     # make the universe from this checkout
#   scripts/universe.sh run "<topic>" [flags...]  # run /research inside it, one claude process
#   scripts/universe.sh shell                     # interactive claude in the universe
#   scripts/universe.sh watch [secs]              # the BOARD, live, from the record (read-only)
#   scripts/universe.sh doctor                    # what is installed, which binaries, what version
#
# Env:
#   WORKDIR   where the universe lives (default ~/.claude/scratch/universe)
#   MODEL / JUDGMENT_MODEL   default haiku/haiku — the engine's own --smoke tier
#   LANES     default 1
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKDIR="${WORKDIR:-$HOME/.claude/scratch/universe}"
MARKETPLACE_NAME=special-circumstances
PLUGINS=(prosthetic-conscience frank-exchange-of-views sleeper-service gray-area)
MODEL="${MODEL:-haiku}"
JUDGMENT_MODEL="${JUDGMENT_MODEL:-haiku}"
LANES="${LANES:-1}"
# bypassPermissions, AND BOTH ALTERNATIVES FAIL. `dontAsk` does not mean "proceed without asking" —
# it DENIES, so every Bash and Read a seat needs returns "Permission to use Bash has been denied
# because Claude Code is running in don't ask mode" and the run writes no record at all. `--bg`
# refuses bypassPermissions until its disclaimer is accepted interactively, which a universe's own
# config dir has never seen. So the run is a foreground `-p`, which is also sufficient: the
# Workflow the skill dispatches SURVIVES the process exit — measured, events kept landing for nine
# minutes after `-p` returned — so the session ending is not the debate ending.
PERMISSION_MODE="${PERMISSION_MODE:-bypassPermissions}"

LIVE=0

log()  { printf '[universe] %s\n' "$*"; }
die()  { printf '[universe] ERROR: %s\n' "$*" >&2; exit 1; }

# --dir MAKES THE UNIVERSE EXPLICIT. A default that is silently reused is how two experiments end
# up in one directory; naming it is cheap and the name is what every guard below is about.
DIR_GIVEN=0
FORCE=0
ARGS=()
while [ $# -gt 0 ]; do
  case "$1" in
    -d|--dir)   WORKDIR="${2:?--dir needs a path}"; DIR_GIVEN=1; shift 2 ;;
    --dir=*)    WORKDIR="${1#--dir=}"; DIR_GIVEN=1; shift ;;
    --live)     LIVE=1; shift ;;
    --force)    FORCE=1; shift ;;
    *)          ARGS+=("$1"); shift ;;
  esac
done
set -- "${ARGS[@]+"${ARGS[@]}"}"

# THE IN-FLIGHT REPORT IS THE THING WORTH PROTECTING, and it is protected proactively rather than
# by a note. `setup` writes .claude/run-live.json and the run's own dashboard treats that marker as
# the signal to keep watching; while it is there, something claims this directory.
#
# Refusing is the DEFAULT because the two failures are not symmetric. Refusing a stale universe
# costs one `--force`. Rebuilding over a live one swaps the hook binaries a running engine spawns
# per tool call, and starting a second run in the same tree writes two debates into one record —
# and a half-overwritten report does not announce itself, it just reads wrong later.
#
# Staleness is REPORTED, never decided here. The engine derives it from the run's own cadence
# (3x the p90 gap between its events); a shell script inventing a rival threshold would be a second
# answer to a question the record already answers. So this prints how long the record has been
# quiet and lets the human rule.
guard_no_live_run() {
  local what="$1"
  local marker="$WORKDIR/src/.claude/run-live.json"
  [ -f "$marker" ] || return 0

  # THE WAL IS WHAT MOVES, NOT THE DATABASE. SQLite in WAL mode appends to record.db-wal and
  # leaves record.db's mtime almost untouched, so a guard that stats the database reports a busy
  # run as long dead. Measured: a run writing events at 16:32:33 showed a record.db mtime of
  # 16:20:04 — twelve minutes of writes invisible — and this guard called it stale on that basis.
  # Every candidate is considered and the NEWEST wins, so the answer cannot depend on which file
  # the engine happened to touch last.
  local newest quiet=""
  newest="$(find "$WORKDIR/src/research" \( -name 'record.db' -o -name 'record.db-wal' -o -name 'record.db-shm' \) -printf '%T@ %p\n' 2>/dev/null | sort -rn | head -1)"
  if [ -n "$newest" ]; then
    local mtime; mtime="${newest%% *}"
    quiet="$(( ( $(date +%s) - ${mtime%.*} ) / 60 )) min since the record was last written"
  else
    quiet="no record written yet"
  fi

  if [ "$FORCE" = "1" ]; then
    log "WARNING: a run claims $WORKDIR ($quiet) — proceeding because --force was given"
    return 0
  fi
  printf '[universe] REFUSING to %s\n' "$what" >&2
  printf '  A run claims %s\n' "$WORKDIR" >&2
  printf '  marker: %s\n' "$marker" >&2
  printf '  %s\n' "$quiet" >&2
  printf '\n  A rebuild swaps the hook binaries a live engine spawns per tool call; a second run\n' >&2
  printf '  writes two debates into one record. Neither failure announces itself.\n\n' >&2
  printf '  If the run is FINISHED or dead:   rm %s\n' "$marker" >&2
  printf '  For a separate experiment:        %s --dir <other> %s\n' "${BASH_SOURCE[0]}" "$what" >&2
  printf '  If you are certain:               add --force\n' >&2
  exit 3
}

# THE ENVIRONMENT IS THE UNIVERSE. Everything below runs under these, and `--live` is the one
# path that does not set them — which is the whole difference between the two modes.
universe_env() {
  if [ "$LIVE" = "1" ]; then return; fi
  export CLAUDE_CONFIG_DIR="$WORKDIR/config"
  export CLAUDE_CODE_PLUGIN_CACHE_DIR="$WORKDIR/plugins"
  mkdir -p "$CLAUDE_CONFIG_DIR" "$CLAUDE_CODE_PLUGIN_CACHE_DIR"
}

# A SEAT'S IDENTITY AND RUN MUST NOT LEAK IN FROM THE PARENT. A shell that has FEOV_RUN set —
# which any session that has touched a run does — makes the tool refuse a run directory it was
# handed, and makes every test in this tree fail in a way that reads like a code defect. Measured:
# a full gate run failed a refusal assertion for exactly this reason and cost an afternoon.
clear_seat_env() {
  unset FEOV_RUN FEOV_RUN_FROM_WRAPPER FEOV_AGENT_ID FEOV_AGENT_TYPE FEOV_HOOK_VERSION
  unset CLAUDE_PROJECT_DIR
}

cmd_build() {
  command -v claude >/dev/null || die "claude CLI not on PATH"
  guard_no_live_run "build"
  command -v go >/dev/null     || die "go not on PATH — the hook binaries cannot be built"
  universe_env; clear_seat_env

  if [ "$LIVE" = "1" ]; then
    log "LIVE: installing into ~/.claude — this is shared with every session on this box"
  else
    log "isolated universe at $WORKDIR"
  fi
  log "source: $REPO (this checkout, at $(git -C "$REPO" rev-parse --short HEAD))"

  # A STAGED MANIFEST, BECAUSE THE REAL ONE PINS RELEASES. Adding $REPO directly looks right and
  # is not: three of the four plugins declare a `git-subdir` source against the GitHub URL with a
  # `ref` pinned to a release tag, so `plugin install` fetches the PUBLISHED content and the
  # working tree is never consulted. Measured the first time this script ran — the universe came
  # up carrying released prompts, with none of the checkout's changes in it, and reported success.
  #
  # So the universe gets its own manifest whose every source is a local path, over a `plugins`
  # symlink into this checkout. Same four plugins, same names; the only difference is that the
  # content resolves to the working tree, which is the entire point of building one.
  # EVERY `claude plugin` CALL RUNS FROM OUTSIDE THE CHECKOUT, and this is not fussiness. Run with
  # cwd inside the repo, `marketplace remove` resolved the PROJECT's .claude/settings.json — a
  # TRACKED file — and stripped its extraKnownMarketplaces and enabledPlugins. A tool meant to leave
  # the tree alone silently edited committed state. $WORKDIR is outside the repo, so project
  # settings are not found and only the universe's own config is touched.
  mkdir -p "$WORKDIR"
  # THE STAGING DIR LIVES OUTSIDE THE UNIVERSE AND IS REMOVED AFTER INSTALL. It was $WORKDIR/
  # marketplace, holding a `plugins` symlink into the checkout — and a run's lead, exploring its own
  # WORKDIR with bypassPermissions, found it and ran
  # $WORKDIR/marketplace/plugins/frank-exchange-of-views/bin/feov-record: the CHECKOUT's binaries,
  # which are whatever a previous build left there, instead of the cache this script just built.
  # A universe that contains a door back to the working tree is not isolated, and the failure is
  # silent — the run works, with the wrong binary.
  local mkt; mkt="$(mktemp -d)"
  trap 'rm -rf "$mkt"' RETURN
  mkdir -p "$mkt/.claude-plugin"
  ln -sfn "$REPO/plugins" "$mkt/plugins"
  python3 "$REPO/scripts/universe-manifest.py" \
    "$REPO/.claude-plugin/marketplace.json" "$mkt/.claude-plugin/marketplace.json" \
    || die "could not stage the local-path manifest"

  if ( cd "$WORKDIR" && claude plugin marketplace list --json 2>/dev/null ) | grep -q "\"$MARKETPLACE_NAME\""; then
    log "marketplace known — refreshing it from this checkout"
    ( cd "$WORKDIR" && claude plugin marketplace remove "$MARKETPLACE_NAME" ) >/dev/null 2>&1
  fi
  ( cd "$WORKDIR" && claude plugin marketplace add "$mkt" ) || die "could not add marketplace from $mkt"

  for p in "${PLUGINS[@]}"; do
    log "installing $p"
    ( cd "$WORKDIR" && claude plugin install "$p@$MARKETPLACE_NAME" ) >/dev/null || log "WARNING: install failed: $p"
  done

  # Binaries are built IN THE CACHE, because that is where ${CLAUDE_PLUGIN_ROOT} resolves. A build
  # in the checkout would leave the installed plugin with no binaries and hooks degrading to a
  # stderr line — which looks, from inside a run, exactly like hooks that are not configured.
  local cache_root="${CLAUDE_CODE_PLUGIN_CACHE_DIR:-$HOME/.claude/plugins}/cache/$MARKETPLACE_NAME"
  [ -d "$cache_root" ] || die "plugin cache not found at $cache_root"

  local first_tools
  first_tools="$(find "$cache_root" -maxdepth 3 -type d -name tools -print -quit 2>/dev/null)"
  [ -n "$first_tools" ] && { log "warming the Go module cache"; go build -C "$first_tools" ./... >/dev/null 2>&1; }

  local built=0
  for plugin_dir in "$cache_root"/*/*/; do
    local tools="$plugin_dir/tools"
    [ -d "$tools/cmd" ] || continue
    mkdir -p "$plugin_dir/bin"
    for cmd_dir in "$tools"/cmd/*/; do
      local name; name="$(basename "$cmd_dir")"
      if go build -C "$tools" -o "$plugin_dir/bin/$name" "./cmd/$name" 2>/dev/null; then
        built=$((built + 1))
      else
        log "WARNING: could not build $name"
      fi
    done
  done
  log "built $built hook binaries into the cache"

  # Enable the plugins for this universe. Without this the content is installed and nothing loads
  # it, which presents as a session with no skills and no hooks.
  local settings="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/settings.json"
  python3 - "$settings" <<'PY'
import json, os, sys
p = sys.argv[1]
d = json.load(open(p)) if os.path.exists(p) else {}
d.setdefault("enabledPlugins", {})
for name in ("prosthetic-conscience", "frank-exchange-of-views", "sleeper-service", "gray-area"):
    d["enabledPlugins"][f"{name}@special-circumstances"] = True
os.makedirs(os.path.dirname(p), exist_ok=True)
json.dump(d, open(p, "w"), indent=2)
print(f"[universe] enabled {len(d['enabledPlugins'])} plugin(s) in {p}")
PY
  cmd_doctor
}

cmd_doctor() {
  universe_env; clear_seat_env
  local cache_root="${CLAUDE_CODE_PLUGIN_CACHE_DIR:-$HOME/.claude/plugins}/cache/$MARKETPLACE_NAME"
  log "cache: $cache_root"
  for plugin_dir in "$cache_root"/*/*/; do
    [ -d "$plugin_dir" ] || continue
    local n; n=$(ls "$plugin_dir/bin" 2>/dev/null | wc -l)
    printf '  %-56s binaries=%s\n' "${plugin_dir#"$cache_root"/}" "$n"
  done
  local rec="$cache_root/frank-exchange-of-views"/*/bin/feov-record
  # shellcheck disable=SC2086
  for r in $rec; do
    [ -x "$r" ] && log "feov-record: schema epoch $("$r" --schema 2>/dev/null), build $("$r" --version 2>/dev/null | awk '{print $NF}')"
  done
}

# ONE BACKGROUND SESSION, BECAUSE THE ENGINE OUTLIVES A TURN.
#
# `claude -p` exits when the model ends its turn. The skill dispatches the debate through the
# Workflow tool, which RETURNS IMMEDIATELY and runs in the background — so the lead invokes it,
# says "I'll relay the envelope when the debate returns", ends its turn, and the process exits
# taking the workflow with it. Measured: the record stopped at eight sitting spans and went stale
# thirteen minutes before anyone noticed, with `subtype: success` and no error anywhere.
#
# `--bg` starts a session that persists and prints an id. The workflow runs under it to completion;
# `claude logs <id>` shows its output, `claude stop <id>` ends it, and the board view below reads
# the record while it works.
cmd_run() {
  local topic="${1:-}"; shift || true
  [ -n "$topic" ] || die "usage: universe.sh run \"<topic>\" [extra /research flags]"
  universe_env; clear_seat_env

  guard_no_live_run "run"
  local src="$WORKDIR/src"
  mkdir -p "$src"
  local stamp; stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  local out="$WORKDIR/run-$stamp.log"

  log "topic: $topic"
  log "tiers: model=$MODEL judgment=$JUDGMENT_MODEL lanes=$LANES"
  log "cwd:   $src"
  log "load at start: $(cut -d' ' -f1-3 /proc/loadavg)"

  local raw="$WORKDIR/stream-$stamp.jsonl"
  log "  progress below; raw stream -> $raw"
  ( cd "$src" && claude -p \
      "/frank-exchange-of-views:research $topic --model $MODEL --judgment-model $JUDGMENT_MODEL --lanes $LANES $*" \
      --output-format stream-json --verbose --permission-mode "$PERMISSION_MODE" </dev/null ) \
    | python3 "$REPO/scripts/universe-stream.py" --raw "$raw" | tee "$out"
  local rc=${PIPESTATUS[0]}
  log "exit $rc — $out"
  log "load at end: $(cut -d' ' -f1-3 /proc/loadavg)"
  log "the Workflow outlives this process: watch it with  $0 --dir $WORKDIR watch"
  return $rc
}

# WHAT THE DEBATE ITSELF IS DOING, as against what the harness is doing. The stream shows dispatch
# and verbs; this shows the board — which is the thing a human actually wants to know. Reads the
# record READ-ONLY so it can never disturb a live run.
cmd_watch() {
  universe_env; clear_seat_env
  local every="${1:-15}"
  while true; do
    python3 "$REPO/scripts/universe-board.py" "$WORKDIR/src" || true
    sleep "$every"
  done
}

cmd_shell() {
  universe_env; clear_seat_env
  local src="$WORKDIR/src"; mkdir -p "$src"
  ( cd "$src" && exec claude )
}

case "${1:-}" in
  build)  shift; cmd_build "$@" ;;
  run)    shift; cmd_run "$@" ;;
  shell)  shift; cmd_shell "$@" ;;
  watch)  shift; cmd_watch "$@" ;;
  doctor) shift; cmd_doctor "$@" ;;
  *) sed -n '1,45p' "${BASH_SOURCE[0]}"; exit 2 ;;
esac
