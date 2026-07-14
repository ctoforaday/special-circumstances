---
description: Run the research debate engine — additive blue team vs gate-keeping red team, judged termination, full adversarial record preserved.
argument-hint: <topic> [--lanes N] [--max-rounds N] [--model sonnet|haiku|opus] [--judgment-model ...] [--smoke]
---

Run a frank exchange of views on the topic in `$ARGUMENTS` (if no topic given, ask — do not guess). Parse optional flags: `--lanes` (blue candidate drafts, default 3; the engine enforces a floor of 3 unless `laneFloorOverride` gives a reason), `--max-rounds` (safety ceiling only — the real terminator is red-PASS or judged deadlock; default 12), and `--smoke` (pipeline exercise for ~50k tokens: sets `lanes: 1`, `maxRounds: 1`, `model: "haiku"`, `laneFloorOverride: "smoke run — pipeline exercise only"`; use before merging debate-script changes).

1. Create the run directory: `research/<yyyy-mm-dd>_<short-slug>/` in the current project, and PRE-CREATE the full blackboard skeleton so agents only ever append to existing artifacts (the harness write-blocks subagents creating report-like files from scratch — the guard is filename-keyed and path-independent, per run 3's controlled experiment): `blue/candidates/` and `red/candidates/` directories, plus stub files `debate.md`, `report.md`, `friction.md`, `blue/frontier.md`, `blue/report.md`, `blue/CHANGELOG.md`, `red/findings.md`, `red/citation-ledger.md` — each seeded with a one-line `# <name> — <topic>` header.
2. **Pin the corpus.** Record the evidence base's state in `<run directory>/inputs/PINNED.md`: the repo HEAD commit at launch, plus the paths (and, for other run directories, the round number) of every corpus the topic cites. Instruct via the topic text that agents cite the pinned commit/round for repo and cross-corpus references — run 3's red audited commits made *mid-run* and cross-corpus citations drifted (R5-2 class); the evidence base must not move under the report. Freeze your own pushes to cited paths while the run is live.
3. **Refresh recall** (skip silently if `.mcp.json` declares no `qmd` server — the pin is the
   opt-in). Index maintenance is CLI, run via the SAME pinned package the server uses — no
   separate install: `npx -y <pkg from .mcp.json> <subcommand>` (a global `qmd` binary is an
   equivalent fast path if present). Ensure the corpus roots the topic cites are indexed
   collections (`... collection list`; add missing ones with `... collection add <path> --name
   sc-<root>` — one collection per corpus root, e.g. `research/`, `ideas/`, never per run, so
   collections stay bounded), then run `... update` and `... embed` so both lexical and
   semantic layers start the run fresh. During the run the `sc-recall-index` hook keeps FTS
   current on every write; re-run `embed` only between rounds if you are relaying anyway (it
   is incremental). Seats never touch the CLI — their access is the MCP server (see
   research-protocol → Recall).
4. Invoke the **Workflow** tool with `scriptPath` = `${CLAUDE_PLUGIN_ROOT}/skills/research-protocol/scripts/debate.js` and `args` = `{ "topic": "<topic>", "runDir": "<run directory path>", "lanes": N, "maxRounds": N, "model": "<model>", "judgmentModel": "<model>", "laneFloorOverride": "<reason, only if lanes < 3>" }` (omit models unless given). Note the **Transcript dir** path the tool prints — step 7 needs it. **Model guidance:** `model` drives the bulk seats (lanes, lenses, responses); `judgmentModel` drives the judgment seats (blue-synthesize, red-merge, judge, assemble) and defaults to inheriting the session model — NOT to `model` — so a cheap dev run keeps full-strength judgment. **For keeper runs, omit `model` entirely: red's lenses ride the bulk tier, and a cheap bulk model silently discounts the leaf-node verification that is load-bearing (retrospective §3 row 16b) — treat lens-sourced grades from cheap-bulk runs with a confidence discount.** `model: sonnet` for development; `--smoke` for smoke tests. On a RESUME, keep the original run's models — changing either busts the agent cache and re-runs completed rounds.
5. AFTER the workflow returns, relay its envelope verbatim (verdict, rounds, lanes, outstanding gaps) plus the run-directory path — YOU MUST NOT re-summarize the report's content; the report is the deliverable, and it is for the human.
6. If the verdict is UNVERIFIED, say so plainly with the outstanding gap count — the gate never soft-passes, and neither does the relay.
7. **Capture the run record** (the retrospective could not be written for runs whose trails evaporated — §3 row 11):
   - Merge the envelope's `friction` entries into `<run directory>/friction.md` (one line each, attributed; seats also append there directly during the run — deduplicate, don't drop) and say so — friction records are input to the self-improvement loop.
   - Copy the workflow transcript dir's `journal.jsonl` into `<run directory>/trajectories/` (git-tracked) and gzip the `agent-*.jsonl` transcripts alongside as `trajectories/agent-transcripts.tar.gz` (gitignored).
   - Run the cost audit: `node ${CLAUDE_PLUGIN_ROOT}/skills/research-protocol/scripts/cost-audit.mjs <transcript dir> > <run directory>/cost.md` — the run's measured tokens and dollars belong in the record next to its friction.
