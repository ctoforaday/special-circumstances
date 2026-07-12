# Backlog — small findings awaiting a PR

Working list of paper cuts and improvements found while dogfooding. Each item graduates into
whichever PR next touches its plugin; bigger items get promoted to `plans/` proposals instead.

- [x] **frank-exchange-of-views**: rename `skills/research-protocol/scripts/workflow.js` → `debate.js` (name by function); update the `scriptPath` reference in `commands/research.md`.
- [x] **prosthetic-conscience**: the doctor command's bootstrap snippet must not clobber Windows `%TMP%` — a bash `TMP=$(mktemp -d)` overwrote the exported variable and broke child processes (`gh`, `go`). Use a non-reserved shell variable name in `commands/doctor.md`.
