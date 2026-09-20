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
# The plugin content is COPIED into a versioned cache at install time, and the binaries are built
# there, after install — that ordering is bootstrap-plugins.sh's and is kept here.
#
# BUT THE CACHE IS NOT THE ONLY PLUGIN ROOT. The staged marketplace is one too, and an empty bin/
# in it is filled BY DESIGN: hooks.json execs hooks/fetch-bin.sh when a binary is missing, which
# downloads that plugin's RELEASE and installs it there. A build that deleted those binaries had
# them back, from the release, forty seconds later — and the run used them. So the build mirrors
# what it compiled into the staging tree as well: every feov-record on disk is this checkout's,
# and which one gets used stops being a question the result depends on.
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
# bypassPermissions, AND BOTH ALTERNATIVES FAIL. `dontAsk` does not mean "proceed without asking" —
# it DENIES, so every Bash and Read a seat needs returns "Permission to use Bash has been denied
# because Claude Code is running in don't ask mode" and the run writes no record at all. `--bg`
# refuses bypassPermissions until its disclaimer is accepted interactively, which a universe's own
# config dir has never seen. So the run is a foreground `-p`, and the process must be kept alive
# for the whole debate: the Workflow does NOT survive it. Print mode terminates background tasks on
# a ceiling, which killed a run mid-sitting. `cmd_run` sets CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS=0
# and says why.
PERMISSION_MODE="${PERMISSION_MODE:-bypassPermissions}"

LIVE=0

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
  # THE STAGING DIR LIVES BESIDE THE UNIVERSE AND OUTLIVES THE BUILD, and both halves of that are
  # load-bearing in opposite directions.
  #
  # OUTSIDE, because it holds a `plugins` symlink into the checkout. It was $WORKDIR/marketplace,
  # and a run's lead exploring its own WORKDIR with bypassPermissions found it and ran
  # $WORKDIR/marketplace/plugins/frank-exchange-of-views/bin/feov-record: the CHECKOUT's binaries,
  # whatever a previous build left there, instead of the cache this script just built. A universe
  # containing a door back to the working tree is not isolated, and it fails silently — the run
  # works, with the wrong binary.
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

  # THE STAGING TREE GETS THE SAME BINARIES, because it is a plugin root too and the hooks will
  # fill it from the RELEASE if it is empty (see the staging comment above). Copied from what was
  # just built rather than rebuilt, so the two roots cannot diverge even by a recompile: whichever
  # one a hook or a curious lead reaches, it is this checkout's code.
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
}

cmd_doctor() {
  universe_env; clear_seat_env
  local cache_root="${CLAUDE_CODE_PLUGIN_CACHE_DIR:-$HOME/.claude/plugins}/cache/$MARKETPLACE_NAME"
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

  # THE PROMOTED GAP-PATTERN CORPUS, because without it the SECOND run in a universe refuses.
  #
  # `setup` reads the memory sources and indexes the corpus by class, since delivery is
  # class-indexed and an unclassified pattern reaches no seat. A fresh universe's project has no
  # promoted corpus at all — and a run's lenses ACCRUE patterns into
  # .claude/agent-memory/<seat>/, unclassified by default. So run one leaves three unclassified
  # files behind and run two refuses: "3 unclassified pattern(s) against 0 delivered class(es)",
  # red would open the run substantially blind while its memory directory looks full.
  #
  # The refusal is right. What was wrong is that the universe never had the classified corpus this
  # repo's own runs read, so red was being seated with less memory than a real run gives it — which
  # would have made every universe smoke a weaker audit than the thing it stands for, silently.
  # Copied, not symlinked: a run accrues INTO this directory, and the checkout is not a scratch pad.
  #
  # THE WHOLE feov-memory DIRECTORY, not just the corpus. Staging red-gap-patterns/ alone left out
  # class-registry.json, which setup reads from the same directory — and without it setup says
  # "nothing constrains --class, so every mint this run will be REFUSED". Measured: the lead hit
  # that refusal, read setup.go to work out why, searched the filesystem for a class-registry.json,
  # found one under ~/.claude/plugins/marketplaces/ and hand-copied it in. Twenty-one tool calls
  # and 256 SECONDS before the workflow dispatched, against 91s for a universe that never staged a
  # corpus at all. Half a memory directory is slower than none, because none refuses cleanly and
  # half sends the lead looking for the other half.
  if [ -d "$REPO/feov-memory" ]; then
    mkdir -p "$src/feov-memory"
    cp -r "$REPO/feov-memory/." "$src/feov-memory/" 2>/dev/null
    log "staged feov-memory: $(find "$src/feov-memory/red-gap-patterns" -name '*.md' 2>/dev/null | wc -l) pattern(s), registry $([ -f "$src/feov-memory/class-registry.json" ] && echo present || echo MISSING)"
  else
    log "WARNING: no $REPO/feov-memory — red opens with accrued memory only, and mints may be refused"
  fi

  local stamp; stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  local out="$WORKDIR/run-$stamp.log"

  log "topic: $topic"
  # --smoke SETS THE TIERS ITSELF, so asking for it alongside a different tier is two instructions
  # and one of them has to lose silently. Refused instead: the operator says which experiment they
  # meant. (Matching values are not a conflict — the defaults here ARE smoke's.)
  smoke_flag=""
  if [ "$SMOKE" = "1" ]; then
    if [ "$MODEL" != "haiku" ] || [ "$JUDGMENT_MODEL" != "haiku" ] || [ "$LANES" != "1" ]; then
      die "SMOKE=1 sets model=haiku judgment=haiku lanes=1 (and --mint-budget 1 --k-max 2), but this run asks for model=$MODEL judgment=$JUDGMENT_MODEL lanes=$LANES. Pass SMOKE=0 for the development configuration, or drop the overrides."
    fi
    smoke_flag=" --smoke"
  fi
  log "tiers: model=$MODEL judgment=$JUDGMENT_MODEL lanes=$LANES mode=$([ "$SMOKE" = "1" ] && echo smoke || echo development)"
  log "cwd:   $src"
  log "load at start: $(cut -d' ' -f1-3 /proc/loadavg)"

  local raw="$WORKDIR/stream-$stamp.jsonl"
  log "  progress below; raw stream -> $raw"
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
  ( cd "$src" && CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS=0 claude -p \
      "/frank-exchange-of-views:research $topic --model $MODEL --judgment-model $JUDGMENT_MODEL --lanes $LANES$smoke_flag $*" \
      --output-format stream-json --verbose --permission-mode "$PERMISSION_MODE" </dev/null ) \
    | python3 "$REPO/scripts/universe-stream.py" --raw "$raw" | tee "$out"
  # BOTH ARMS, because they fail differently and the second one is the one that matters. claude's
  # status says the process produced an answer; the renderer's says the ENGINE RAN. A universe whose
  # plugins did not load exits 0 on the first and 2 on the second. Reading only PIPESTATUS[0] is the
  # same defect as reading a pipeline's exit code and getting `tee`'s.
  local claude_rc=${PIPESTATUS[0]} render_rc=${PIPESTATUS[1]}
  local rc=$claude_rc
  [ "$claude_rc" -eq 0 ] && rc=$render_rc
  log "exit $rc — $out"
  log "load at end: $(cut -d' ' -f1-3 /proc/loadavg)"
  log "board: $0 --dir $WORKDIR watch   (a run with no outcome row was TRUNCATED, not slow)"
  return "$rc"
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
