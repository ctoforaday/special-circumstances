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
# THE STAGED MARKETPLACE IS THE PLUGIN ROOT A SESSION LOADS. It is a `directory` marketplace
# source, and Claude Code runs a directory source's plugins IN PLACE: ${CLAUDE_PLUGIN_ROOT}, the
# skills' base directory, the agent definitions, hooks.json and bin/ all resolve under
# $WORKDIR.marketplace/plugins/<name>. `claude plugin install` still copies each plugin into the
# versioned cache and names that copy in installed_plugins.json, but a session does not load it.
#
# The binaries are built in the cache, after install — bootstrap-plugins.sh's ordering — and then
# mirrored into the staged marketplace. An empty bin/ there is filled BY DESIGN: hooks.json execs
# hooks/fetch-bin.sh when a binary is missing, which downloads that plugin's RELEASE and installs
# it there. With the mirror, every feov-record on disk is this checkout's, and which root a hook or
# a lead reaches stops being a question the result depends on.
#
# Usage:
#   scripts/universe.sh build                     # make the universe from this checkout
#   scripts/universe.sh run "<topic>" [flags...]  # run /research inside it, one DETACHED claude
#                                                 # process, and follow its log. --no-follow returns
#                                                 # at once; killing the follower leaves the run running
#   scripts/universe.sh resume                    # restart a run whose process died with no outcome
#   scripts/universe.sh shell                     # interactive claude in the universe
#   scripts/universe.sh watch [secs]              # the BOARD, live, from the record (read-only), until
#                                                 # the process ends. Exit 3 means TRUNCATED
#   scripts/universe.sh doctor                    # what is installed, which binaries, what version
#
# Env:
#   WORKDIR   where the universe lives (default ~/.claude/scratch/universe)
#   MODEL / JUDGMENT_MODEL   default haiku/haiku — the engine's own --smoke tier
#   LANES     default 1
#   CONCURRENT_AGENTS  default 16 — the Workflow tool's per-run concurrent agent limit. The engine's
#             own default is min(16, cpu_cores - 2), which is 2 on a 4-core box and serialises the
#             seven-lens fan-out into four waves.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKDIR="${WORKDIR:-$HOME/.claude/scratch/universe}"
MARKETPLACE_NAME=special-circumstances
PLUGINS=(prosthetic-conscience frank-exchange-of-views sleeper-service gray-area)
# SMOKE IS THE DEFAULT, AND IT HAS TO BE SAID RATHER THAN IMPLIED.
#
# This script set the tiers and the lane count to smoke values and then left --mint-budget and
# --k-max at the engine's defaults, so it produced something LABELLED a smoke in every log line
# that was not one. Measured across two runs meant to be compared: the 2026-09-19 baseline ran
# --mint-budget 1 --k-max 2 (launched --smoke by hand) and the 2026-09-20 run ran 5 and 6. Their
# gap counts and epoch counts were then read as a result about the change under test.
#
# SMOKE=0 runs the development configuration, which is an order of magnitude more expensive and is
# a deliberate act; either way the mode is printed beside the tiers, so a log says which experiment
# it was.
SMOKE="${SMOKE:-1}"
MODEL="${MODEL:-haiku}"
JUDGMENT_MODEL="${JUDGMENT_MODEL:-haiku}"
LANES="${LANES:-1}"
# THE WORKFLOW ENGINE'S CONCURRENCY IS min(16, cpu_cores - 2), AND THAT HEURISTIC IS WRONG FOR
# THIS WORKLOAD.
#
# Measured on two runs (universe-m7, universe-m8) on a 4-core box: peak concurrency 2, exactly
# min(16, 4-2), with 38 of 38 sitting starts firing within 0.1s of another sitting ENDING — a hard
# semaphore, not scheduling noise. Seven lenses are dispatched per epoch and run as four waves of
# two, so the epoch-1 lens phase took 542s where its longest single lens took 214s.
#
# cpu_cores - 2 is the right denominator for CPU-bound agents. A seat is INFERENCE-bound: it waits
# on the API and spends its own CPU only on the occasional feov-record spawn. Capping fan-out by
# core count prices idle waiting as if it were compute.
#
# 16 is the engine's own ceiling and the operator's call (gblock, 2026-09-24). The widest fan-out the
# debate has is seven lens sittings in one epoch, so 16 is headroom rather than a target: it stops
# the cap being the thing that decides, and leaves the lane count free to grow.
CONCURRENT_AGENTS="${CONCURRENT_AGENTS:-16}"
# bypassPermissions, AND BOTH ALTERNATIVES FAIL. `dontAsk` does not mean "proceed without asking" —
# it DENIES, so every Bash and Read a seat needs returns "Permission to use Bash has been denied
# because Claude Code is running in don't ask mode" and the run writes no record at all. `--bg`
# refuses bypassPermissions until its disclaimer is accepted interactively, which a universe's own
# config dir has never seen. So the run is a `-p`, and the process must be kept alive for the whole
# debate: the Workflow does NOT survive it. Print mode terminates background tasks on a ceiling,
# which killed a run mid-sitting, so `cmd_supervise` sets CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS=0 and
# says why; `launch_detached` keeps the process out of reach of whatever launched it, and says why.
PERMISSION_MODE="${PERMISSION_MODE:-bypassPermissions}"

LIVE=0
FOLLOW=1

log()  { printf '[universe] %s\n' "$*"; }
die()  { printf '[universe] ERROR: %s\n' "$*" >&2; exit 1; }

# --dir MAKES THE UNIVERSE EXPLICIT. A default that is silently reused is how two experiments end
# up in one directory; naming it is cheap and the name is what every guard below is about.
FORCE=0
ARGS=()
while [ $# -gt 0 ]; do
  case "$1" in
    -d|--dir)   WORKDIR="${2:?--dir needs a path}"; shift 2 ;;
    --dir=*)    WORKDIR="${1#--dir=}"; shift ;;
    --live)     LIVE=1; shift ;;
    --force)    FORCE=1; shift ;;
    --no-follow) FOLLOW=0; shift ;;
    *)          ARGS+=("$1"); shift ;;
  esac
done
set -- "${ARGS[@]+"${ARGS[@]}"}"

# WHAT SAYS WHETHER A RUN IS ALIVE, AND WHAT IT LAUNCHED. Written per launch; a resume repoints the
# pid file and the log, and leaves the origin stream where it is, because that is where the run's
# Workflow call was first made.
PID_FILE="$WORKDIR/run.pid"
CURRENT_LOG="$WORKDIR/run-current.log"
ORIGIN_STREAM="$WORKDIR/run-origin.jsonl"

# IS THE LAST LAUNCH STILL RUNNING: `running`, `dead`, or `unknown` when nothing recorded one. Silence
# cannot answer this: a seat thinking for ten minutes and a killed run write the same nothing.
#
# THE PID IS A SESSION ID, and the question is whether that SESSION has a live process, not whether
# that pid does. The supervisor can be SIGKILLed while claude goes on writing the record, which a
# `kill -0` on the supervisor would report as dead and `resume` would then answer with a second lead
# on one record. It also survives a reboot: a recycled pid is rarely a session leader, so a stale pid
# file reads as dead rather than as alive for as long as some unrelated process holds the number.
process_state() {
  local pid
  pid="$(cat "$PID_FILE" 2>/dev/null)"
  if [ -z "$pid" ]; then echo unknown; return; fi
  if ps -o stat= -s "$pid" 2>/dev/null | grep -qv '^Z'; then echo running; else echo dead; fi
}

# The rc of the run's EXIT line in this log, or nothing when the log has none. The renderer's lines all
# carry a [time] prefix, so nothing else writes a line that starts this way.
exit_rc() {
  sed -n 's/^EXIT \([0-9][0-9]*\)$/\1/p' "$1" 2>/dev/null | tail -n 1
}

process_detail() {
  if [ ! -f "$PID_FILE" ]; then printf 'no pid file at %s' "$PID_FILE"; return; fi
  local rc exit_says="no EXIT line"
  rc="$(exit_rc "$CURRENT_LOG")"
  [ -z "$rc" ] || exit_says="EXIT $rc"
  printf 'pid %s, %s, log %s' "$(cat "$PID_FILE")" "$exit_says" "$CURRENT_LOG"
}

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
  # A LIVE PROCESS IS NOT A QUESTION FOR --force, and it holds before `setup` has written the marker.
  # The marker can go stale and the human has to rule on it; a process in this universe's session is
  # measured, and proceeding would overwrite the pid file that is the only handle on it.
  if [ "$(process_state)" = running ]; then
    printf '[universe] REFUSING to %s\n' "$what" >&2
    printf '  A run process is alive in %s (%s)\n' "$WORKDIR" "$(process_detail)" >&2
    printf '  To end it:   pkill -s %s\n' "$(cat "$PID_FILE")" >&2
    printf '  For a separate experiment:   %s --dir <other> %s\n' "${BASH_SOURCE[0]}" "$what" >&2
    exit 3
  fi
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
  printf '  If the run was TRUNCATED:          %s --dir %s resume\n' "${BASH_SOURCE[0]}" "$WORKDIR" >&2
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
  share_credentials
}

# PLUGINS AND SETTINGS ARE ISOLATED; AUTHENTICATION IS NOT, AND THAT DISTINCTION COST REAL MONEY.
#
# CLAUDE_CONFIG_DIR isolates the whole config, which is right for what this script is FOR — a
# universe must not install its plugins into ~/.claude or be affected by what is already there.
# But the credential lives in that directory too, so isolating it isolated the subscription as
# well, and `claude -p` inside the universe fell through to whatever else it could authenticate
# with: the console/API profile on this box, in a different organization, billed in CREDITS.
#
# MEASURED, and it went unnoticed because the number looked like telemetry rather than a bill.
# Every run reports total_cost_usd whichever way it authenticated, so ~$21 a run read as an
# interesting statistic for months (project memory: "$209/10 runs") and was actual spend. The
# 2026-09-20 smoke then died mid-epoch-3 with "Credit balance is too low", which is the API's
# message and not a subscription's, thirty-five minutes and $7.09 into a run that had to be thrown
# away.
#
# A SYMLINK RATHER THAN A COPY, so a token refreshed in the real session is picked up here and
# there is one source of truth. A copy goes stale at the next refresh and leaves a second credential
# on disk under scratch.
#
# AN ABSENT CREDENTIAL IS LOUD. Falling back silently is the whole defect: the run still works, it
# just bills somewhere else, and the only symptom arrives as an out-of-credit failure long after the
# decision was made.
share_credentials() {
  local src="$HOME/.claude/.credentials.json"
  local dst="$CLAUDE_CONFIG_DIR/.credentials.json"
  if [ ! -e "$src" ]; then
    printf '[universe] WARNING: no %s — the universe cannot see your subscription and will\n' "$src" >&2
    printf '           authenticate however it can, which on this box means API CREDITS.\n' >&2
    return
  fi
  [ -L "$dst" ] && [ "$(readlink "$dst")" = "$src" ] && return
  rm -f "$dst"
  ln -s "$src" "$dst"
}

# A SEAT'S IDENTITY AND RUN MUST NOT LEAK IN FROM THE PARENT. A shell that has FEOV_RUN set —
# which any session that has touched a run does — makes the tool refuse a run directory it was
# handed, and makes every test in this tree fail in a way that reads like a code defect. Measured:
# a full gate run failed a refusal assertion for exactly this reason and cost an afternoon.
clear_seat_env() {
  unset FEOV_RUN FEOV_RUN_FROM_WRAPPER FEOV_AGENT_ID FEOV_AGENT_TYPE FEOV_HOOK_VERSION
  unset CLAUDE_PROJECT_DIR
  hermetic_path
}

# THE HOST'S PLUGIN BINARIES MUST NOT BE REACHABLE FROM INSIDE AN ISOLATED UNIVERSE, and they were:
# the plugin system puts every installed plugin's bin/ on PATH, and a universe inherits the parent
# shell's PATH wholesale. CLAUDE_CODE_PLUGIN_CACHE_DIR isolates what the universe INSTALLS; it does
# nothing about what is already reachable by name.
#
# MEASURED, and it cost universe-m11 an entire sitting. A blue seat typed a BARE `feov-record`
# instead of the run's `$RUN/.bin/feov-record`, and PATH handed it
# ~/.claude/plugins/cache/.../frank-exchange-of-views/1.72.0/bin/feov-record — the SAME plugin
# version as the universe's, at event-schema epoch 8 against the run's epoch 14. Its write died on a
# raw `CHECK constraint failed: type (275)`, every call after it was refused with "`register` is your
# first act and it has not happened", and that sitting made 100 tool calls and recorded NOTHING. A
# plugin version is not an epoch, so the matching version made the two look interchangeable.
#
# REMOVED RATHER THAN SHADOWED. Prepending the run's .bin would make a bare invocation work, and a
# bare invocation working is what let this happen unnoticed for a whole run: the seat's prompt names
# the binary by path, so the honest outcome for a name lookup is "command not found" — loud, local to
# the call, and it sends the seat to the path it was given. Shadowing would also be impossible to do
# correctly here, because the run directory does not exist until `setup` runs inside the universe.
#
# ISOLATED ONLY. Under `--live` the universe INSTALLS into ~/.claude, so the host cache IS its cache
# and stripping it would remove the binaries the run needs.
hermetic_path() {
  if [ "$LIVE" = "1" ]; then return; fi
  local host_cache="$HOME/.claude/plugins/cache" out="" p dropped=0
  local IFS=:
  for p in $PATH; do
    case "$p" in
      "$host_cache"/*) dropped=$((dropped + 1)); continue ;;
    esac
    out="${out:+$out:}$p"
  done
  unset IFS
  if [ "$dropped" -gt 0 ]; then
    export PATH="$out"
    log "PATH: dropped $dropped host plugin-cache bin dir(s) — a bare tool name must not reach the host's binaries"
  fi
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
  # THE BUILD RECORDS THE COMMIT IT BUILT, and builds only a commit: a smoke is bound to what it ran.
  [ -z "$(git -C "$REPO" status --porcelain)" ] || die "the checkout has uncommitted changes — a universe is built from a commit"
  rm -f "$WORKDIR/built-from"
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
  # THE STAGING DIR LIVES BESIDE THE UNIVERSE AND OUTLIVES THE BUILD, and both halves of that are
  # load-bearing in opposite directions.
  #
  # OUTSIDE, because a lead exploring its own WORKDIR with bypassPermissions runs whatever
  # feov-record it finds there, and the staging tree is a plugin root with a bin/ of its own.
  #
  # OUTLIVES, because `extraKnownMarketplaces` in the universe's settings.json records this PATH,
  # and Claude re-reads it at every session start. Staged in `mktemp -d` and deleted on return, the
  # config pointed at a directory that no longer existed: the plugins silently failed to load, so
  # `/frank-exchange-of-views:research` was not a command, and the lead — asked to research a topic
  # with a slash command it did not have — simply ANSWERED THE TOPIC. One turn, five cents, exit 0.
  # Deleting it also put it in /tmp, which this repo's scratch rule forbids for exactly this class
  # of "it worked when I ran it" failure.
  local mkt; mkt="${WORKDIR%/}.marketplace"
  rm -rf "$mkt"
  mkdir -p "$mkt/.claude-plugin"

  # A COPY, AND EVERY `bin/` IT CONTAINS IS BUILT FROM THIS CHECKOUT BELOW.
  #
  # Three attempts at this, and the first two were the same mistake in different places. It began
  # as a `plugins` symlink into the checkout, inside WORKDIR and then beside it; both times the
  # lead explored, found $mkt/plugins/frank-exchange-of-views/bin/feov-record, and ran `setup` with
  # it — the CHECKOUT's binary, whatever a previous build happened to leave there. A 16:18 build
  # drove a run against a 17:58 cache and staged the very file that run existed to prove deleted.
  #
  # So the third attempt removed bin/ entirely, and that was WORSE, because it created a vacuum the
  # plugin fills by design. hooks.json runs ${CLAUDE_PLUGIN_ROOT}/bin/<hook>, and when that is
  # missing it execs hooks/fetch-bin.sh, which downloads the binaries FROM THE RELEASE TAG and
  # installs them into that same bin/. Measured: bin/ reappeared 40 seconds after a build that had
  # deleted it, holding `feov-record version babd3ea` — the 1.72.0 release, predating every change
  # the run was built to exercise. The lead's first act was `ls -la bin/`, and it found them.
  #
  # A path a lead can reach is a path a lead will use, and it is right to — it went looking for the
  # tool and found one. Hiding the tool is not the fix and neither is removing it: the fix is that
  # EVERY feov-record on disk is this checkout's, so which one gets used stops being a question
  # the result depends on. A populated bin/ also means the self-heal never fires, so a universe
  # cannot silently downgrade itself to the last release.
  mkdir -p "$mkt/plugins"
  for p in "${PLUGINS[@]}"; do
    cp -r "$REPO/plugins/$p" "$mkt/plugins/$p" || die "could not stage plugin $p"
    rm -rf "$mkt/plugins/$p/bin"
    # .fetch IS THE SELF-HEAL'S STATE, and staging the developer's copy of it is inheriting an
    # accident. This checkout carries one from a release fetch that ran against it; `cp -r` put it
    # in the universe, where its stamp happened to match and so happened to hold the self-heal shut.
    # A universe on a clean clone has no such luck. Stripped here and written deliberately below.
    rm -rf "$mkt/plugins/$p/.fetch"
  done
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

  # Binaries are built IN THE CACHE, where the install put each plugin's tools/, and mirrored into
  # the staged marketplace below, which is where ${CLAUDE_PLUGIN_ROOT} resolves. A build in the
  # checkout leaves both roots with no binaries and hooks degrading to a stderr line — which looks,
  # from inside a run, exactly like hooks that are not configured.
  #
  # They are COMPILED FROM THE CHECKOUT's source and written into the cache, because Go stamps a
  # binary with its commit only when it builds inside a git tree, and the cache is a copy with no
  # .git: compiled there, every register's tool_version and hook_version read `unknown`, so run.md
  # could not tell one build from another. The checkout is the commit the cache was installed from
  # (the clean-tree check above).
  local cache_root="${CLAUDE_CODE_PLUGIN_CACHE_DIR:-${CLAUDE_CONFIG_DIR:-$HOME/.claude}/plugins}/cache/$MARKETPLACE_NAME"
  [ -d "$cache_root" ] || die "plugin cache not found at $cache_root"

  local first_tools
  first_tools="$(find "$REPO/plugins" -maxdepth 2 -type d -name tools -print -quit 2>/dev/null)"
  [ -n "$first_tools" ] && { log "warming the Go module cache"; go build -C "$first_tools" ./... >/dev/null 2>&1; }

  local built=0
  for plugin_dir in "$cache_root"/*/*/; do
    local pname; pname="$(basename "$(dirname "$plugin_dir")")"
    local tools="$REPO/plugins/$pname/tools"
    [ -d "$plugin_dir/tools/cmd" ] || continue
    [ -d "$tools/cmd" ] || die "the cache holds $pname's tools but the checkout has no $tools/cmd"
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

  # THE STAGING TREE GETS THE SAME BINARIES, because it is the root a session loads and the hooks
  # fill it from the RELEASE if it is empty (see the header). Copied from what was just built
  # rather than rebuilt, so the two roots cannot diverge even by a recompile: whichever one a hook
  # or a curious lead reaches, it is this checkout's code.
  local staged=0
  for plugin_dir in "$cache_root"/*/*/; do
    [ -d "$plugin_dir/bin" ] || continue
    local pname; pname="$(basename "$(dirname "$plugin_dir")")"
    [ -d "$mkt/plugins/$pname" ] || continue
    mkdir -p "$mkt/plugins/$pname/bin"
    for b in "$plugin_dir"/bin/*; do
      [ -f "$b" ] || continue
      cp -f "$b" "$mkt/plugins/$pname/bin/" || die "could not mirror $(basename "$b") into the staging tree"
      staged=$((staged + 1))
    done
  done

  # THE SELF-HEAL TRIGGERS ON THE STAMP, NOT ON AN EMPTY bin/ — which is what the previous version
  # of this comment got wrong, and it said "the self-heal never reaches for a release" over code
  # that did not stop it. fetch-bin.sh lists a binary as MISSING when `.fetch/installed` does not
  # name this plugin's tag, however present and executable the file is: "When the stamp names
  # another release, or none, every binary is fetched." A `go build` writes no stamp, so a universe
  # built on a clean clone is one SessionStart away from having its binaries replaced by the
  # release's, mid-run, with `setup`'s version preflight already past.
  #
  # So the stamp is written for every root this script populates, naming that plugin's own version
  # — the same string tag_of() computes — and the assertion below proves `ensure` then does nothing.
  local stamped=0
  for root in "$cache_root"/*/*/ "$mkt/plugins"/*/; do
    [ -d "$root" ] || continue
    local pj="$root/.claude-plugin/plugin.json"
    [ -f "$pj" ] || pj="$root/plugin.json"
    [ -f "$pj" ] || continue
    local nm ver
    nm=$(python3 -c "import json,sys;print(json.load(open(sys.argv[1]))['name'])" "$pj" 2>/dev/null) || continue
    ver=$(python3 -c "import json,sys;print(json.load(open(sys.argv[1]))['version'])" "$pj" 2>/dev/null) || continue
    [ -n "$nm" ] && [ -n "$ver" ] || continue
    mkdir -p "$root/.fetch"
    printf '%s\n' "$nm--v$ver" > "$root/.fetch/installed"
    stamped=$((stamped + 1))
  done
  log "mirrored $staged binaries and stamped $stamped root(s) as installed"

  # ASSERT IT, rather than assert it in a comment. `ensure` prints a systemMessage and spawns a
  # background fetch exactly when it thinks a binary is missing; on a correctly stamped root it
  # says nothing. A universe whose binaries the release would overwrite mid-run is not a universe
  # that proves anything about this checkout, and that is a checkable condition, so it is checked.
  for root in "$cache_root"/*/*/; do
    [ -f "$root/hooks/fetch-bin.sh" ] || continue
    local out
    out=$(CLAUDE_PLUGIN_ROOT="$root" sh "$root/hooks/fetch-bin.sh" ensure SessionStart 2>&1)
    if [ -n "$out" ]; then
      die "the self-heal still fires in $(basename "$(dirname "$root")"): $out
  A stamped root must make \`ensure\` a no-op. If it does not, this universe's binaries can be
  replaced by the release's at any session start, and the run would be measuring the release."
    fi
  done
  log "self-heal checked: silent on every cache root"

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

  # THE CONFIG'S MARKETPLACE PATH MUST STILL EXIST, asserted here rather than discovered by a run.
  # A dangling path does not fail the build, the install, or `doctor` — the cache is fully
  # populated and every binary is there. It fails at SESSION START, by the plugins not loading,
  # which presents as a lead that answers the topic instead of researching it.
  python3 - "$settings" "$mkt" <<'PY' || die "the universe's marketplace path is not the one just staged"
import json, os, sys
settings, staged = sys.argv[1], sys.argv[2]
d = json.load(open(settings))
src = (d.get("extraKnownMarketplaces", {}).get("special-circumstances", {}).get("source") or {})
path = src.get("path")
if path != staged:
    print(f"[universe] settings.json names {path!r}, not the staged {staged!r}")
    raise SystemExit(1)
if not os.path.isdir(path):
    print(f"[universe] settings.json names {path!r}, which does not exist")
    raise SystemExit(1)
print(f"[universe] marketplace source resolves: {path}")
PY
  cmd_doctor
  log "the marketplace source lives at $mkt — removing the universe means removing BOTH paths"
  printf '%s %s\n' "$(git -C "$REPO" rev-parse HEAD)" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$WORKDIR/built-from"
}

cmd_doctor() {
  universe_env; clear_seat_env
  local cache_root="${CLAUDE_CODE_PLUGIN_CACHE_DIR:-${CLAUDE_CONFIG_DIR:-$HOME/.claude}/plugins}/cache/$MARKETPLACE_NAME"
  log "cache: $cache_root"
  for plugin_dir in "$cache_root"/*/*/; do
    [ -d "$plugin_dir" ] || continue
    local n; n=$(find "$plugin_dir/bin" -maxdepth 1 -type f 2>/dev/null | wc -l)
    printf '  %-56s binaries=%s\n' "${plugin_dir#"$cache_root"/}" "$n"
  done
  local rec=("$cache_root/frank-exchange-of-views"/*/bin/feov-record)
  for r in "${rec[@]}"; do
    [ -x "$r" ] && log "feov-record: schema epoch $("$r" --schema 2>/dev/null), build $("$r" --version 2>/dev/null | awk '{print $NF}')"
  done
}

# A RUN IS DETACHED FROM WHATEVER LAUNCHED IT, because the launcher's lifetime is not the debate's.
#
# Measured on m37: `run` held `claude -p` in the foreground of an agent's Bash background task, and
# that task's 7200000 ms ceiling killed the debate in epoch 11 of 12, with three lens sittings open,
# no verdict and no outcome. Nothing warned of it, and no earlier smoke had run long enough to find it.
#
# So the debate gets its own session, with no stdin or stdout shared with the caller, and what the
# caller holds is at most a `tail` of the log. A group kill of the caller cannot reach it, and neither
# can a closed pipe.
#
# THE SUPERVISOR WRITES ITS OWN PID. `setsid` forks when its caller leads a process group, so `$!` is
# not always the session; the session leader is the one process that can say for certain.
launch_detached() {
  local stamp="$1" mode="$2" session="${3:-}"
  command -v setsid >/dev/null || die "setsid not on PATH: a run cannot be detached from its launcher"
  rm -f "$PID_FILE"
  ln -sfn "run-$stamp.log" "$CURRENT_LOG"
  setsid "${BASH:-bash}" "${BASH_SOURCE[0]}" --dir "$WORKDIR" _supervise "$stamp" "$mode" "$session" \
    </dev/null >>"$WORKDIR/run-$stamp.log" 2>&1 &
  local tries=0
  until [ -s "$PID_FILE" ]; do
    tries=$((tries + 1))
    [ "$tries" -le 100 ] || die "the supervisor wrote no pid file in 10s: see $WORKDIR/run-$stamp.log"
    sleep 0.1
  done
}

# The stamp names this launch's log, stream and prompt. It has to be unused: two launches in one
# second would append to one log, and the first one's EXIT line would then answer for the second.
new_stamp() {
  local stamp; stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  while [ -e "$WORKDIR/run-$stamp.log" ]; do
    sleep 1
    stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  done
  printf '%s' "$stamp"
}

# What `run` and `resume` tell the caller once the process is detached, and then the follow. Following
# is a view: the tail ending, for any reason, says nothing about the run until the state is read.
report_launch() {
  local stamp="$1" pid out="$WORKDIR/run-$1.log"
  pid="$(cat "$PID_FILE")"
  log "detached: session $pid. Killing this command leaves the run running"
  log "log:    $out"
  log "stream: $WORKDIR/stream-$stamp.jsonl"
  log "wait:   tail -n +1 -f --pid=$pid $out | grep -m1 '^EXIT '   (no line means it was killed)"
  log "board:  $0 --dir $WORKDIR watch"
  [ "$FOLLOW" = "1" ] || return 0

  tail -n +1 -f --pid="$pid" "$out" &
  local tail_pid=$!
  trap 'kill "$tail_pid" 2>/dev/null; log "follower stopped; the run goes on (session $pid)"; exit 143' TERM INT HUP
  wait "$tail_pid"
  trap - TERM INT HUP

  if [ "$(process_state)" = running ]; then
    log "follower stopped; the run goes on (session $pid)"
    return 0
  fi
  local rc; rc="$(exit_rc "$out")"
  if [ -z "$rc" ]; then
    log "the run's process is gone with no EXIT line: it was killed"
    return 1
  fi
  return "$rc"
}

cmd_run() {
  local topic="${1:-}"; shift || true
  [ -n "$topic" ] || die "usage: universe.sh run \"<topic>\" [extra /research flags]"
  universe_env; clear_seat_env

  guard_no_live_run "run"
  local src="$WORKDIR/src"
  mkdir -p "$src"

  local stamp; stamp="$(new_stamp)"
  local out="$WORKDIR/run-$stamp.log"

  # The header goes to the log only: the follow below prints the log from its first line.
  log "topic: $topic" >> "$out"
  # --smoke SETS THE TIERS ITSELF, so asking for it alongside a different tier is two instructions
  # and one of them has to lose silently. Refused instead: the operator says which experiment they
  # meant. (Matching values are not a conflict — the defaults here ARE smoke's.)
  smoke_flag=""
  if [ "$SMOKE" = "1" ]; then
    if [ "$MODEL" != "haiku" ] || [ "$JUDGMENT_MODEL" != "haiku" ] || [ "$LANES" != "1" ]; then
      die "SMOKE=1 sets model=haiku judgment=haiku lanes=1 (and --mint-budget 3 --k-max 2), but this run asks for model=$MODEL judgment=$JUDGMENT_MODEL lanes=$LANES. Pass SMOKE=0 for the development configuration, or drop the overrides."
    fi
    smoke_flag=" --smoke"
  fi
  {
    log "tiers: model=$MODEL judgment=$JUDGMENT_MODEL lanes=$LANES mode=$([ "$SMOKE" = "1" ] && echo smoke || echo development)"
    # PRINTED BESIDE THE TIERS because it changes the wall clock a run is compared on: two runs at
    # different concurrency are not the same experiment, and the engine's own default depends on the
    # BOX rather than on anything the run records.
    log "concurrency: $CONCURRENT_AGENTS agent(s) (engine default here would be $(( $(nproc) - 2 )) on $(nproc) cores)"
    log "cwd:   $src"
  } >> "$out"

  printf '%s' "/frank-exchange-of-views:research $topic --model $MODEL --judgment-model $JUDGMENT_MODEL --lanes $LANES$smoke_flag $*" \
    > "$WORKDIR/prompt-$stamp.txt"
  ln -sfn "stream-$stamp.jsonl" "$ORIGIN_STREAM"
  launch_detached "$stamp" run
  report_launch "$stamp"
}

# THE SAME DEBATE, RESTARTED BY ITS OWN LEAD. A run with no outcome is resumable: the Workflow tool
# replays what it finished, so the lead's session is resumed and told to invoke it with
# resumeFromRunId and the SAME args. The args, the script, the session and the run id are read from
# the stream's first Workflow call and its result, so none of them is retyped by an operator.
#
# REFUSED UNLESS THE PROCESS IS MEASURED DEAD. A second lead on a live run writes two debates into
# one record, and an unknown process state is not evidence of a dead one.
cmd_resume() {
  universe_env; clear_seat_env
  case "$(process_state)" in
    running)
      printf '[universe] REFUSING to resume: the run process is still alive (%s)\n' "$(process_detail)" >&2
      exit 3 ;;
    unknown)
      printf '[universe] REFUSING to resume: %s, so nothing says the process is dead\n' "$(process_detail)" >&2
      exit 3 ;;
  esac
  [ -f "$ORIGIN_STREAM" ] || { printf '[universe] REFUSING to resume: no stream at %s\n' "$ORIGIN_STREAM" >&2; exit 3; }

  local stamp; stamp="$(new_stamp)"
  local fields
  # The script refuses a stream with no Workflow call and a record with an outcome, and says which.
  fields="$(python3 "$REPO/scripts/universe-resume.py" "$ORIGIN_STREAM")" || exit 3
  local session prompt
  session="${fields%%$'\n'*}"
  prompt="${fields#*$'\n'}"
  printf '%s' "$prompt" > "$WORKDIR/prompt-$stamp.txt"

  log "resuming lead session $session" >> "$WORKDIR/run-$stamp.log"
  launch_detached "$stamp" resume "$session"
  report_launch "$stamp"
}

# THE DETACHED SIDE: one claude and its renderer, in the session launch_detached made. Internal; the
# environment is the launcher's, already the universe's. stdout and stderr are the run log.
#
# THIS FUNCTION ENDS IN `exit`, and that is load-bearing. Bash reads a script as it goes, and a run
# lasts hours while the checkout goes on being edited (see "The universe is a snapshot"). The
# functions were parsed at startup; returning to the case below would read whatever is now at that
# byte offset.
cmd_supervise() {
  local stamp="$1" mode="$2" session="${3:-}"
  local raw="$WORKDIR/stream-$stamp.jsonl" out="$WORKDIR/run-$stamp.log"
  local prompt; prompt="$(cat "$WORKDIR/prompt-$stamp.txt")"
  printf '%s\n' "$$" > "$PID_FILE.tmp"
  mv "$PID_FILE.tmp" "$PID_FILE"
  log "load at start: $(cut -d' ' -f1-3 /proc/loadavg)"
  cd "$WORKDIR/src" || { printf 'EXIT 1\n'; exit 1; }

  local resume_args=() render_args=()
  if [ "$mode" = resume ]; then
    resume_args=(--resume "$session")
    render_args=(--resumed)
  fi

  # THE DEBATE DOES NOT OUTLIVE THIS PROCESS, and believing it did cost a truncated run that read
  # as a finished one. `claude -p` waits a bounded time for background tasks and then TERMINATES
  # them: "Background tasks still running after 600s; terminating". A research run is tens of
  # minutes, so the workflow is killed mid-debate — measured on `is 91 prime`, which died with a
  # lens's `mint` as its last event and no `sitting_close` after it: nine sitting spans opened,
  # eight closed, no chair verdict, no outcome, three gaps left OPEN. Nothing in the stream says
  # "truncated"; the board just stops, which is the same shape as a run that is still thinking.
  #
  # An earlier measurement found the workflow alive nine minutes after exit and concluded it
  # survives. Nine minutes is 540s — INSIDE the ceiling. The measurement never reached the boundary
  # it was taken to have tested.
  #
  # 0 means wait indefinitely, which is what a run that is the whole point of the process needs.
  #
  # CLAUDE WRITES THE RAW STREAM TO A FILE, AND THE RENDERER TAILS THE FILE, so the view cannot harm
  # the capture. Measured on m37's resume: with claude piped into the renderer, the renderer exited on
  # the session's first `result` event and left the debate writing to a pipe with no reader. A seat's
  # tool that outlives a killed claude holds no pipe here either: the tail ends when claude does, so
  # the EXIT line does not wait on it.
  : > "$raw"
  CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS=0 CLAUDE_CODE_WORKFLOW_MAX_CONCURRENT_AGENTS="$CONCURRENT_AGENTS" \
    claude -p "${resume_args[@]+"${resume_args[@]}"}" "$prompt" \
      --output-format stream-json --verbose --permission-mode "$PERMISSION_MODE" </dev/null >>"$raw" &
  local claude_pid=$!
  # A signal to this process ends the run with its line written, rather than leaving claude behind
  # with nobody to report it. `wait` returns early on a trapped signal, hence the loop.
  trap 'kill -TERM "$claude_pid" 2>/dev/null' TERM HUP
  tail -n +1 -f --pid="$claude_pid" "$raw" | python3 "$REPO/scripts/universe-stream.py" "${render_args[@]+"${render_args[@]}"}" &
  local render_pid=$!

  wait "$claude_pid"
  local claude_rc=$?
  while kill -0 "$claude_pid" 2>/dev/null; do
    wait "$claude_pid"
    claude_rc=$?
  done
  wait "$render_pid"
  local render_rc=$?

  # WHAT THE KILLED CLAUDE LEFT BEHIND ENDS WITH IT. A seat's tool still running after its lead died
  # would go on writing to the record under a resume, and would read as a live run to process_state.
  trap '' TERM
  pkill -TERM -s "$$" 2>/dev/null

  # BOTH ARMS, because they fail differently and the second one is the one that matters. claude's
  # status says the process produced an answer; the renderer's says the ENGINE RAN. A universe whose
  # plugins did not load exits 0 on the first and 2 on the second.
  local rc=$claude_rc
  [ "$claude_rc" -eq 0 ] && rc=$render_rc
  log "claude exit $claude_rc, renderer exit $render_rc — $out"
  log "load at end: $(cut -d' ' -f1-3 /proc/loadavg)"
  # THE LAST LINE, AND THE ONE AN OPERATOR WAITS ON. Absent, the supervisor itself was killed.
  printf 'EXIT %d\n' "$rc"
  exit "$rc"
}

# WHAT THE DEBATE ITSELF IS DOING, as against what the harness is doing. The stream shows dispatch
# and verbs; this shows the board — which is the thing a human actually wants to know. Reads the
# record READ-ONLY so it can never disturb a live run.
#
# The process state is read BEFORE the board. An outcome is written before the process ends, so a
# process read as dead cannot be paired with a board read too early to hold its outcome.
#
# It stops when the process has ended, since the board can no longer move, and takes the board's
# status: 3 is TRUNCATED.
cmd_watch() {
  universe_env; clear_seat_env
  local every="${1:-15}" state rc
  while true; do
    state="$(process_state)"
    python3 "$REPO/scripts/universe-board.py" "$WORKDIR/src" --process "$state" --detail "$(process_detail)" \
      --resume "$0 --dir $WORKDIR resume"
    rc=$?
    [ "$state" = dead ] && return "$rc"
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
  resume) shift; cmd_resume "$@" ;;
  _supervise) shift; cmd_supervise "$@" ;;
  shell)  shift; cmd_shell "$@" ;;
  watch)  shift; cmd_watch "$@" ;;
  doctor) shift; cmd_doctor "$@" ;;
  *) sed -n '/^# Usage:/,/^set -uo/{/^set -uo/d;p;}' "${BASH_SOURCE[0]}"; exit 2 ;;
esac
