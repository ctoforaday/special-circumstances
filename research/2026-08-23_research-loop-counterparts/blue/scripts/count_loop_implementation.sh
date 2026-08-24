#!/usr/bin/env bash
# Settles, by computation, the build-state claim in blue's report:
# FEOV's Go implementation size vs. sleeper-service's scaffold vs. dream/memory's
# zero implementation, all at the pinned commit 2ce929f.
set -euo pipefail
cd /home/user/special-circumstances

echo "== FEOV .go file count at 2ce929f =="
git ls-tree -r --name-only 2ce929f -- plugins/frank-exchange-of-views/tools | grep -c '\.go$'

echo "== sleeper-service file list at 2ce929f =="
git ls-tree -r --name-only 2ce929f -- plugins/sleeper-service

echo "== dream/memory implementation primitives at 2ce929f (expect: none) =="
git ls-tree -r --name-only 2ce929f | grep -iE 'dream\.js|memory-consolidator|memory-curator|knowledge-ingest|\.knowledge\.toml|okf' || echo "NONE FOUND"

echo "== memory-architecture.md status line =="
git show 2ce929f:plans/memory-architecture.md | sed -n '3p'

echo "== sleeper-service README status line =="
git show 2ce929f:plugins/sleeper-service/README.md | grep -i "^Status:"
