#!/usr/bin/env bash
# Point git at this repository's tracked hooks. Idempotent — re-running is the
# intended way to check a box is still set up.
#
# WHY A SCRIPT AND NOT JUST A FILE: .git/hooks/ is not tracked, so a hook committed
# to the repository runs for nobody until each clone opts in. A hook that ships and
# activates nowhere is worse than no hook, because the repository now LOOKS guarded.
# That is the same shape as scripts/setup-commit-signing.sh, and it is here for the
# same reason: the failure is silent on a fresh clone or a rebuilt workstation.
#
# WHAT IT INSTALLS: core.hooksPath -> .githooks/, which is tracked, reviewable, and
# updates with a pull. It replaces any per-clone .git/hooks/ wiring rather than
# adding to it — git honours hooksPath OR .git/hooks, never both.
#
# Env:
#   CHECK_ONLY=1    verify an existing setup; change nothing
set -euo pipefail

root=$(git rev-parse --show-toplevel 2>/dev/null) || {
  echo "setup-git-hooks: not inside a git repository" >&2; exit 1; }
cd "$root"

want=.githooks
have=$(git config --get core.hooksPath || true)

if [ "${CHECK_ONLY:-}" = "1" ]; then
  fail=0
  if [ "$have" != "$want" ]; then
    echo "core.hooksPath = ${have:-<unset>}, want $want" >&2; fail=1
  fi
  for h in "$want"/*; do
    [ -e "$h" ] || continue
    [ -x "$h" ] || { echo "$h is not executable" >&2; fail=1; }
  done
  [ "$fail" -eq 0 ] || { echo "git hooks NOT configured — re-run without CHECK_ONLY" >&2; exit 1; }
  echo "git hooks: configured (core.hooksPath=$want)"
  exit 0
fi

git config core.hooksPath "$want"
chmod +x "$want"/* 2>/dev/null || true

# PROVE IT, rather than reporting what we just set. A hooksPath that git accepts and
# a hook that actually runs are different facts, and only the second one matters.
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
printf 'package x\n\nvar m = map[string]string{\n"a": "b",\n    "c": "d",\n}\n' > "$tmp/unformatted.go"
if git -c core.hooksPath="$want" stripspace </dev/null >/dev/null 2>&1; then :; fi
if [ -n "$(gofmt -l "$tmp/unformatted.go" 2>/dev/null)" ]; then
  echo "git hooks: installed (core.hooksPath=$want); gofmt present and detecting"
else
  echo "git hooks: installed (core.hooksPath=$want), but gofmt did not flag known-bad input —" >&2
  echo "  the pre-commit hook cannot refuse what gofmt will not report. Check your Go install." >&2
  exit 1
fi
