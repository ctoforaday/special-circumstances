---
name: git-proficiency
description: Use when performing git operations — repository health guardrails (big-file check, no destructive deletes), surgical recovery, and the maintained cheatsheet.
---

# git-proficiency

The agent is the custodian of repository history.

- BEFORE `git add .` or a commit, YOU MUST check for oversized files (`find . -type f -size +50M -not -path '*/.*'`); anything over 50MB goes in `.gitignore` first — YOU MUST NOT commit model weights, database dumps, or logs.
- During an input/output error from any service, YOU MUST check disk capacity (`df -h`) BEFORE network troubleshooting — a full disk is a version-control failure, not a service failure.
- YOU MUST NOT run `rm -rf` on a workspace directory that contains persistent state.
- AFTER accidental deletion or history corruption: identify the last healthy commit (`git log -n 5 --stat`), recover surgically (`git show <commit>:<path> > <recovery_path>`), and check `git reflog` immediately before garbage collection claims it.
- AFTER discovering a new recovery pattern, YOU MUST record it in this skill's `CHEATSHEET.md`.
