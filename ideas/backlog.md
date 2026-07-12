# Backlog — small findings awaiting a PR

Working list of paper cuts and improvements found while dogfooding. Each item graduates into
whichever PR next touches its plugin; bigger items get promoted to `plans/` proposals instead.

- [ ] **frank-exchange-of-views**: rename `skills/research-protocol/scripts/workflow.js` → `debate.js` (name by function); update the `scriptPath` reference in `commands/research.md`.
- [ ] **prosthetic-conscience**: the doctor command's bootstrap snippet must not clobber Windows `%TMP%` — a bash `TMP=$(mktemp -d)` overwrote the exported variable and broke child processes (`gh`, `go`). Use a non-reserved shell variable name in `commands/doctor.md`.
- [ ] **frank-exchange-of-views**: investigate the workflow-subagent write-block — blue's `report.md` write was refused ("Subagents should return findings as text, not write report files") while `debate.md`/`CHANGELOG.md` writes succeeded. If report-named files are systematically blocked for workflow agents, the blackboard design needs an explicit carve-out or different filenames. Evidence: run wf_40406212-f7d friction + `undefined/` artifacts.
- [ ] **frank-exchange-of-views**: run-1 staged inputs (`inputs/`, seeded `debate.md`) vanished from the run directory during the broken run — find what deleted them (journal has the record) before trusting staged inputs.
- [ ] **frank-exchange-of-views**: formalize trajectory capture in `/research` — after each run, copy the workflow `journal.jsonl` into `<run>/trajectories/` (git-tracked; compact structured returns) and gzip the per-agent transcripts alongside (gitignored; durable home decided by the memory-architecture proposal, PR #2). Raw trajectories are the primary self-learning input and currently evaporate with the session.
