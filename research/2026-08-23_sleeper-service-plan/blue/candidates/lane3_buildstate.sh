#!/bin/sh
# lane-3 leaf audit: is the #429 subject (sleeper-service loop) actually built?
# Re-arms: run from repo root (pin 2ce929f). Prints the plugin's real file tree,
# the artifacts the plan (claude-port-plan.md 3c) says it must carry, and PASS/FAIL
# on each. Settles the verification-status question mechanically: an absent command
# cannot have produced a green Phase-4 E2E.
set -eu
root="${1:-.}"
ss="$root/plugins/sleeper-service"
echo "== actual sleeper-service tree =="
find "$ss" -type f | sort
echo
echo "== planned artifacts (claude-port-plan.md 3c) present? =="
for a in \
  "skills/continuous-learning/SKILL.md" \
  "commands/self-improve.md" \
  "commands/graduate.md" \
  "docs/scheduling.md"; do
  if [ -e "$ss/$a" ]; then echo "PRESENT  $a"; else echo "ABSENT   $a"; fi
done
echo
echo "== plugin.json dependencies array (plan 3c: must declare frank-exchange-of-views)? =="
if grep -q '"dependencies"' "$ss/.claude-plugin/plugin.json"; then echo "PRESENT"; else echo "ABSENT"; fi
echo
echo "== any hook enforcing write-confinement to research/+ideas/ across the suite? =="
if grep -rniE 'ideas/|research/' "$root"/plugins/*/hooks/*.json 2>/dev/null | grep -qiE 'deny|block|confine'; then
  echo "FOUND enforcement"; else echo "NONE — confinement is prose-only, not mechanized"; fi
