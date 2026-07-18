# red candidate pass — round 1, lens 2 (leaf-node citation verification, slice 2 of 4)

Slice: report §2 (H2 — /self-improve + /graduate mechanics) and §3 (H3 — scheduling and
headless reality), plus the footnote definitions those sections reference. Full report
re-read whole (2 consecutive windows, 1111 lines). Citation ledger was empty at start
(round 1); every claim below verified fresh at the leaf. Repo working tree confirmed AT the
pin (`git rev-parse HEAD` = 7bc501e, clean) — pinned files read directly.

## Findings

### L2-F1 — MEDIUM: body claims all three MCP issues "leaf-checked OPEN"; #32191 is CLOSED (duplicate)
- **Location:** §3.3 "Disconfirming pass — where headless breaks", item 2. Quoted sentence:
  "three relevant upstream defects, status leaf-checked OPEN 2026-07-17: stdio MCP tools
  silently missing on the first turn when server startup exceeds ~2s (regression since
  2.1.144, #76239); stdio tool calls hanging with several servers loaded, worked around by
  `--strict-mcp-config` (#68375); HTTP MCP `-p` runs exiting silently (#32191, older)."
- **Verified (gh issue view, 2026-07-17):** #76239 OPEN ✓ (and the ~2s figure is in the
  issue body: "The effective wait window is roughly ~2s" ✓); #68375 OPEN ✓ (title confirms
  2.1.177 regression + strict-mcp-config workaround ✓); **#32191 CLOSED as DUPLICATE** ✗.
- **The defect:** the body sentence asserts leaf-checked-OPEN for all three; the footnote
  [^McpHeadlessBugs] itself disclaims #32191 ("status per search listing; not individually
  re-fetched — grade accordingly"). Body overstates the verification scope AND the status is
  factually wrong. Footnote-lags-body class (red pattern: incomplete-repair footnote lag /
  open-bug-actually-closed).
- **Grading:** likelihood CERTAIN (verified); impact LOW-MEDIUM — the two load-bearing bugs
  (#76239/#68375, which motivate the strict-mcp-config + daemon workarounds) ARE open, so
  the design is unaffected; the defect is a false verification claim in a report whose
  authority rests on verification discipline. Complexity-to-fix TRIVIAL.
- **Required fix:** restate as "two leaf-checked OPEN (#76239, #68375); #32191 CLOSED as
  duplicate (canonical issue not traced) — retained as historical evidence of the class."

### L2-F2 — LOW: "exit code 0/1" attributed to the CLI reference, which carries no exit-code documentation
- **Location:** §3.2, "Exit behavior: code 0 success; code 1 on `--max-turns` exhaustion and
  other errors" + footnote [^CliReference] listing "exit code 0/1" among flags "quoted
  verbatim".
- **Verified:** full cli-reference fetch 2026-07-17 — no exit-code documentation on the page
  (the --max-turns entry says only "Exits with an error when the limit is reached"). Probe
  P1 corroborates 0-on-success only. The specific value "1" is unpinned. Lane 3's
  `claude --help` cross-check may be the real source, but the footnote attributes it to the
  doc page.
- **Grading:** likelihood CERTAIN (misattribution verified); impact LOW (the wrapper should
  test `!= 0`, not `== 1`, regardless — and §8 open question 6 already demands the
  forced-budget exit-semantics test); complexity TRIVIAL.
- **Required fix:** soften to "exits nonzero (doc: 'exits with an error'; exact code
  unpublished — treat any nonzero as failure)" or re-attribute the 0/1 claim to
  `claude --help` 2.1.212 explicitly.

### L2-F3 — LOW: [^DGM] footnote quote "improve themselves the more compute they are provided" not in the cited source
- **Location:** Footnotes, [^DGM] (referenced from §1.2 and §2.4): 'Key quotes: ...
  "improve themselves the more compute they are provided."'
- **Verified:** arXiv abs/2505.22954 fetched (quote ABSENT); arXiv html/2505.22954 fetched
  (quote and close variants ABSENT — closest is a compute-COST remark in the Conclusion).
  Likely a transposition from the Sakana DGM post's framing (footnote over-attribution
  class). The load-bearing DGM claims survive independently: abstract verbatim "empirically
  validates each change using coding benchmarks" ✓; html §3 "Only agents that compile
  successfully and retain the ability to edit a given codebase are added to the DGM
  archive" + "Each newly generated agent is quantitatively evaluated on a chosen coding
  benchmark" ✓ — so §2.4's "only admits a change after empirical validation, never on the
  proposer's say-so" stands (precision note: DGM admits benchmark-EVALUATED agents even
  when scores are lower — gate is validity + evaluation, not success; this parallels, not
  breaks, the stub→graduation analogy).
- **Grading:** likelihood CERTAIN; impact LOW (decorative quote, not load-bearing);
  complexity TRIVIAL. Fix: drop the quote or re-attribute to [^DGMSakana] after locating it
  there.

### L2-F4 — LOW (infrastructure, for the lead): inputs/PINNED.md pins a path that does not exist at the pin
- **Location:** `inputs/PINNED.md` row "`plans/claude-port-plan.md` | `7bc501e`" vs footnote
  [^PortPlan]: "the pin's `plans/claude-port-plan.md` path does not exist in the
  special-circumstances tree at `7bc501e` (verified by `git show`)".
- **Verified:** `git cat-file -e 7bc501e:plans/claude-port-plan.md` → MISSING; pin's plans/
  tree holds only README.md + efficiency-phase.md. Blue's footnote is CORRECT and honestly
  disclosed; the defect is in the run's pin manifest (setup tooling asserted a nonexistent
  path), and the port-plan citation is therefore snapshot-grade (AgentOrange working tree),
  not pin-grade. The quotes blue carries from it verified verbatim against
  `AgentOrange/docs/claude-port-plan.md` ("human approves each step"; Phase-4 verify line;
  resolved decision 6 daily/opt-in) — content fidelity HIGH, pin integrity absent.
- **Grading:** likelihood CERTAIN; impact LOW today (single-operator box, file verified in
  the working tree at audit time) but this is the cross-corpus-drift class the pinning
  convention exists to kill; complexity LOW (fix the setup script's pin validation, or
  stage the file into inputs/). Not a blue defect — routed to the lead as protocol/tooling.

### L2-F5 — INFO: #66395 ([^WindowsHang]) is a [DOCS] issue, CLOSED as NOT_PLANNED
- **Verified (gh):** title confirms the regression window (v2.1.161–v2.1.168, fixed
  v2.1.169) exactly as the report carries it; but the issue is a docs-gap complaint, closed
  not-planned, and the regression facts are the issue author's title assertion (body not
  audited). Blue's self-grade (MEDIUM details / HIGH existence) is honest and stands. Nice
  to add the closed-not-planned status at next touch; no fix demanded.

## Verification table (claim ↔ reference, confidence, attempt lines for every grade below HIGH)

| # | Claim (§) | Reference | Confidence | Note / attempt line if <HIGH |
|---|---|---|---|---|
| 1 | --bare skip list; "will become the default for `-p` in a future release" (§3.2) | [^HeadlessDocs] | HIGH | verbatim match, full-page fetch 2026-07-17 |
| 2 | skills/custom commands work in -p; /login-class excluded (§3.2) | [^HeadlessDocs] | HIGH | verbatim |
| 3 | bg subagent/workflow wait capped 10 min v2.1.182+; env var; 0=no limit (§3.2) | [^HeadlessDocs] | HIGH | verbatim; note: cap applies to BACKGROUND subagents/workflows awaited at exit — a foreground-blocking workflow is not the capped case; report's "MUST raise ceiling" is correct where FEOV runs backgrounded |
| 4 | system/init `plugins`/`plugin_errors`; fail-CI guidance (§3.2) | [^HeadlessDocs] | HIGH | verbatim |
| 5 | json output: total_cost_usd + per-model breakdown (§3.2, §5.1) | [^HeadlessDocs] | HIGH | verbatim |
| 6 | --resume + session id; same-directory scope (§3.2) | [^HeadlessDocs] | HIGH | verbatim |
| 7 | bare skips OAuth/keychain → API-key/apiKeyHelper auth (§3.2) | [^HeadlessDocs] | HIGH | verbatim |
| 8 | 10MB stdin cap (§3.2) | [^HeadlessDocs] | HIGH | verbatim (v2.1.128+) |
| 9 | dontAsk auto-denies non-allowlisted; "locked-down CI" (§3.2/§4.2) | [^HeadlessDocs]/[^CliReference] | HIGH | verbatim both pages |
| 10 | --max-turns / --max-budget-usd / --strict-mcp-config / --plugin-dir / --fallback-model / --settings / --json-schema≥2.1.205 descriptions (§3.2) | [^CliReference] | HIGH | verbatim, full flag-table fetch |
| 11 | exit code "0 success / 1 on exhaustion" (§3.2) | [^CliReference] | **LOW** | ATTEMPT: full cli-reference fetch — exit-code docs ABSENT from page; probe P1 shows 0-success only → L2-F2 |
| 12 | -p workspace trust: "no dialog appears and the rules stay ignored" (§3.2) | [^PermissionsDoc] | HIGH | verbatim (permissions page, "Project allow rules and workspace trust") |
| 13 | /loop session-scoped; 7-day expiry; "No catch-up for missed fires" (§3.3) | [^ScheduledTasks] | HIGH | verbatim |
| 14 | disable-model-invocation → scheduled fire reaches Claude as plain text (§2.4, §4.3 L6) | [^ScheduledTasks] | HIGH | verbatim, v2.1.196+; scope note: doc covers SCHEDULED fires — the -p-session case is exactly blue's open question 3, correctly carried |
| 15 | Desktop per-task permission config, local files, machine-on; cloud ≥1h min interval (§3.4) | [^ScheduledTasks] | HIGH | comparison table verbatim |
| 16 | Routines: autonomous/no prompts; claude/-prefix push restriction; fresh default-branch clone; daily cap + "rejected until the window resets"; connectors default-included; identity attribution; claude.ai-login-only; green-status warning; research-preview volatility (§3.4, §5.1) | [^RoutinesDocs] | HIGH | all ten quotes verbatim, full-page fetch |
| 17 | #76239 OPEN + ~2s window (§3.3) | [^McpHeadlessBugs] | HIGH | gh status + issue body |
| 18 | #68375 OPEN, strict-mcp-config workaround (§3.3) | [^McpHeadlessBugs] | HIGH | gh status + title |
| 19 | #32191 "leaf-checked OPEN" (§3.3 body) | [^McpHeadlessBugs] | **LOW** | ATTEMPT: gh issue view → CLOSED as DUPLICATE → L2-F1 |
| 20 | #66395 regression window/fix version (§3.3) | [^WindowsHang] | MEDIUM | ATTEMPT: gh status fetched (CLOSED NOT_PLANNED, [DOCS] issue); facts are title-asserted, body not audited → L2-F5; blue's own MEDIUM grade honest |
| 21 | #837/#14246 historical, superseded by current doc (§3.3) | [^SlashHeadlessIssues] | HIGH | gh: #837 CLOSED COMPLETED, #14246 CLOSED DUPLICATE — consistent with supersession story; statuses now leaf-fetched (upgrade from blue's flagged-unfetched) |
| 22 | #23707 background agents fail on web sandbox (§3.4) | [^WebSandbox] | HIGH | gh: exists, CLOSED NOT_PLANNED (strengthens: won't fix) |
| 23 | GHA schedule delay/drop language + 5-min minimum (§3.4) | [^GhaSchedule] | HIGH | GitHub's own doc verbatim: "can be delayed during periods of high loads... some queued jobs may be dropped" |
| 24 | GHA "5–30 min typical at :00" numbers (§3.4) | [^GhaSchedule] | MEDIUM | ATTEMPT: GitHub doc fetched (no numbers there); community discussions #52477/#156282 not fetched — carried as community-measured per blue's own label |
| 25 | systemd Persistent= semantics (§3.4) | [^MissedRun] | HIGH | verbatim via man7.org (freedesktop.org returned 403 — noted as friction; alternate source succeeded) |
| 26 | Task Scheduler missed-start catch-up (§3.4) | [^MissedRun] | HIGH | learn.microsoft.com StartWhenAvailable: "can start the task at any time after its scheduled time has passed" — semantics confirmed; report's string is the UI label, API doc corroborates |
| 27 | anacron interval-since-last-run (§3.4) | [^MissedRun] | MEDIUM | ATTEMPT: not individually fetched (both sibling primitives verified; anacron semantics standard) — trivial to pin at build |
| 28 | Probes P1/P2 outputs (exit 0, $0.0246903, terminal_reason completed, hook write BLOCKED, $0.058) (§3.1) | [^HeadlessProbe] | MEDIUM | ATTEMPT: lane-3 transcript grep — all figures match the transcript verbatim; CLI 2.1.212 confirmed live by this audit (`claude --version`); residual = self-reported ephemeral instrument, acknowledged by blue §7 with the re-run-and-commit fix offered — accept the offer at build, no gap minted |
| 29 | smoke mode = 1 lane/1 round/1 citation pass/haiku/~50k (§2.2 step 4) | [^ResearchCommand]+[^Backlog item 17] | HIGH | both pinned files read; "1 citation pass" is backlog-17 wording ✓ |
| 30 | "for keeper runs, omit model entirely"; judgmentModel inherits session (§2.4) | [^ResearchCommand] | HIGH | verbatim, pinned file |
| 31 | script-vs-prose doctrine quote (§1.4/§4.1) | [^ResearchCommand] | HIGH | verbatim |
| 32 | stop-and-resume standing practice; capture emits cost.md + run-record-audit.md (§5.3/§3.4) | [^ResearchCommand] | HIGH | pinned file steps 3/5 |
| 33 | backlog items 10/17/27c/34/36/39 as quoted (§1–§3, footnotes) | [^Backlog]/[^QmdDaemon] | HIGH | pinned file read; item-34 daemon quotes (PID file, /health, :8181/mcp, 36.3s vs 2.9s) verbatim |
| 34 | qmd-unreachable degradation precedent (§3.3 item 2c) | [^QmdFallback] | HIGH | pinned run-4 friction, blue-lane-1 entry verbatim: "fell back to Grep/Read on the local corpus, workable here" |
| 35 | smoke run honest-UNVERIFIED + Catechism friction precedent (§2.1) | [^SmokeRecord] | HIGH | pinned friction.md read — matches exactly |
| 36 | IdeaStudy: novel p<0.05, weaker feasibility, self-eval failures + diversity lack, human rerank improved (§2.1) | [^IdeaStudy] | HIGH | abs verbatim + html: "AI Ideas + Human Rerank" condition confirmed, rerank outscored AI-alone (5.81 vs 5.64 novelty) |
| 37 | Reflexion: feedback → verbal reflection → episodic memory → later trials (§1.2/§2.2) | [^Reflexion] | HIGH | abstract verbatim |
| 38 | DGM: archive + empirical validation of each change (§2.4) | [^DGM] | HIGH | abstract verbatim "empirically validates each change using coding benchmarks"; html §3 admission conditions quoted — precision note in L2-F3 |
| 39 | [^DGM] key-quote "improve themselves the more compute they are provided" | [^DGM] | **LOW** | ATTEMPT: abs fetch + html full-text fetch — phrase absent from both → L2-F3 |
| 40 | DGMSakana: fake test logs; markers-removal quote; lineage quote; sandbox under human supervision (§2.4/§4.1) | [^DGMSakana] | HIGH | all four quotes verbatim from sakana.ai/dgm |
| 41 | Port-plan quotes: "human approves each step"; Phase-4 verify line; decision 6 daily/human-opt-in; §3c structure (§2.4, §3.3, §3.4) | [^PortPlan] | HIGH (content) / pin ABSENT | content verbatim vs AgentOrange working tree; pin-grade impossible — L2-F4 (infrastructure) |

## Envelope

```json
{
  "lens": "L2 — leaf-node citation verification, slice 2/4 (§2–§3 + owned footnotes)",
  "findings": ["L2-F1 (MEDIUM)", "L2-F2 (LOW)", "L2-F3 (LOW)", "L2-F4 (LOW/infra)", "L2-F5 (INFO)"],
  "verdict_input": "no blocking finding in this slice; §2–§3's evidentiary base is unusually sound — 34 of 41 checked pairs HIGH, every sub-HIGH pair carries an attempt line",
  "citation_ledger": "41 entries appended",
  "friction": [
    "freedesktop.org systemd man page returned HTTP 403 to WebFetch — worked around via man7.org mirror; the protocol could name preferred-mirror fallbacks for common man-page sources",
    "permissions-doc fetch overflowed the tool-result cap (54KB persisted to a side file) — grep-on-persisted-output worked but is undocumented workflow; a domain-scoped fetch-extract prompt discipline or page-section fetch would be cleaner",
    "gh issue view proved strictly better than WebFetch for issue-status leaf checks (7 statuses in one call) — worth a protocol note so lanes stop grading down on 'not individually re-fetched' when a one-line gh loop was triable"
  ]
}
```
