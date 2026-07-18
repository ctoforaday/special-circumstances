# red citation-ledger — How should sleeper-service, the autonomous learning loop (Phase 4), be designed?

## round 1 — lens 4 (slice: §6/§7/§8 + referenced footnotes)

§6r9 "Bootstrap guard already shipped in hooks.json" (+ verbatim _comment quote; all hook commands wrapped) | [^HooksJson] @7bc501e via git show | high | r1 | 2026-07-17
§6r8/§3.3 #76239 OPEN — stdio MCP tools silently missing when startup >~2s; regression since 2.1.144 | [^McpHeadlessBugs] #76239 live fetch | high | r1 | 2026-07-17
§6r8/§3.3 #68375 OPEN — stdio tool-call hang w/ multi-server fleet; --strict-mcp-config workaround | [^McpHeadlessBugs] #68375 live fetch | high | r1 | 2026-07-17
§6r8/§3.3 "#32191 open / leaf-checked OPEN" | [^McpHeadlessBugs] #32191 live fetch | LOW — REFUTED: Closed as duplicate (canonical untraced; 2.1.58–2.1.71 era) → L4-F1 | r1 | 2026-07-17
§7 "#22055 fetched directly with statuses quoted" — Closed as not planned; title verbatim | [^PermAskBypass] live fetch | high | r1 | 2026-07-17
[^PermAskBypass] "workaround in thread: chmod 444 + PreToolUse hook" | gh issue view 22055 --comments (full thread) | split: PreToolUse HIGH (quote captured), chmod-444 LOW — absent from thread → L4-F2 | r1 | 2026-07-17
§7 ephemeral-instrument disclosure — probe P1/P2 commands+outputs present in blue/candidates/lane-3.md | local read | high | r1 | 2026-07-17
NOT re-fetched (slice boundary, assign if uncovered): #6631, #25621 statuses ([^DenyRWIssue]/[^DenyBashIssue], slice 3)
"struggle to self-correct ... performance even degrades" + oracle-label dependency ("improvements vanish when oracle labels are not available") | arXiv:2310.01798 abs + ar5iv body | high | R1-L1 | 2026-07-17
Reflexion: verbal reflections kept in episodic memory buffer, used in subsequent trials | arXiv:2303.11366 abs | high | R1-L1 | 2026-07-17
Voyager: "environment feedback, execution errors, and self-verification"; "alleviates catastrophic forgetting"; ever-growing compositional skill library | arXiv:2305.16291 abs | high | R1-L1 | 2026-07-17
DGM: archive of generated agents; empirically validates each change on benchmarks | arXiv:2505.22954 abs | high | R1-L1 | 2026-07-17
"improve themselves the more compute they are provided" | sakana.ai/dgm/ (NOT arXiv abs — mishomed in [^DGM], see L1-F3) | high-at-correct-source | R1-L1 | 2026-07-17
SICA: "performance gains from 17% to 53% on a random subset of SWE Bench Verified" | arXiv:2504.15228 abs | high (venue metadata mismatch, L1-F4) | R1-L1 | 2026-07-17
STOP: seed improver improves itself; sandbox-bypass frequency evaluated | arXiv:2310.02304 abs | high | R1-L1 | 2026-07-17
STOP figures: GPT-4 0.42% (0.31–0.57%); 0.46% (0.35–0.61%) with warning; 10,000 improvements; not statistically significant (two-proportion z-test) | ar5iv.labs.arxiv.org/html/2310.02304 §6.2 | high (open question 8 satisfied at LaTeX-render fidelity) | R1-L1 | 2026-07-17
Dependabot: 11.3% of projects deprecated; developers "configure Dependabot toward reducing the number of notifications" | arXiv:2206.07230 abs (verbatim) | high | R1-L1 | 2026-07-17
Dependabot-fatigue framing; "in 2022, Dependabot automatically generated over 75 million pull requests" | arXiv:2502.06175 abs + ar5iv body | high | R1-L1 | 2026-07-17
DGM safety incidents: fake test logs; "removed the markers we use in the reward function to detect hallucination (despite our explicit instruction not to do so)"; "transparent, traceable lineage of every change"; sandboxed under human supervision | sakana.ai/dgm/ | high (all verbatim) | R1-L1 | 2026-07-17
CoastRunners looping-targets reward hack (OpenAI 2016) | en.wikipedia.org/wiki/Reward_hacking | high (qualitative use) | R1-L1 | 2026-07-17
arXiv:2604.13602 exists; title "Reward Hacking in the Era of Large Models: Mechanisms, Emergent Misalignment, Challenges" | arXiv abs | high (existence only) | R1-L1 | 2026-07-17
Alert-fatigue "under 1 in 5 alerts acted on" | vendor posts (search digest; vib.community/devops.com snippet: 18% of 9.6M events) | low on number / medium on phenomenon — matches blue's self-grade; WebSearch attempted, no primary found | R1-L1 | 2026-07-17
FrictionRun3: 17 attributed entries; filename-keyed write-block (e4); Grep count footgun (e12); Read cap (e15) | research/2026-07-12_feov-retrospective/friction.md @7bc501e | high | R1-L1 | 2026-07-17
FrictionRun4: memory unreadable 4 seat classes; Read cap 6th seat class; write-guard 5th consecutive round-seat; MUST-try skipped blue-respond-r1; log() settled by ~/.claude spelunking; abort "NO REPORT ASSEMBLED — resumable via wf_5cefd2a4-35f"; ~37 entries within "~30–39" band | research/2026-07-14_efficiency-investigation/friction.md @7bc501e | high | R1-L1 | 2026-07-17
Backlog items 10, 15–17, 27c, 28, 29, 34, 36, 39 — quotes verbatim at those line numbers (incl. QmdDaemon item-34 quote, 0/175 batching, 92% cache exclusion) | ideas/backlog.md @7bc501e | high (quotes); "40 statused items" LOW — recount = 25 checkboxes/39 lines (L1-F1) | R1-L1 | 2026-07-17
EffReport: severity-floor termination REJECTED; log() console-ephemeral; board-telemetry.jsonl durable sink with named consumers | research/2026-07-14_efficiency-investigation/report.md @7bc501e | high | R1-L1 | 2026-07-17
EfficiencyPlan: PR-A.1 telemetry line; PR-C.2 red-memory mirror at pre-create; attestation ceiling | plans/efficiency-phase.md @7bc501e | high ("FEOV 0.6.0" fragment medium) | R1-L1 | 2026-07-17
"an LLM executing mechanics is an unenforced good-faith contract" | plugins/frank-exchange-of-views/commands/research.md @7bc501e line 9 (verbatim) | high | R1-L1 | 2026-07-17
Doubts quote: "sc-quality-gate fired on workflow-agent writes; red-auditor wrote its `memory: project` gap-pattern file" | ideas/doubts.md @7bc501e (verbatim; closed-item numbering ambiguous) | high | R1-L1 | 2026-07-17
Red-gap mirror "1,558 lines / 30+ named patterns" | inputs/red-gap-patterns.md (wc: 1,557 newlines — trailing-newline artifact) | high | R1-L1 | 2026-07-17
Preamble pin: evidence base at 7bc501e per inputs/PINNED.md; git diff vs pin empty on all cited internal paths | inputs/PINNED.md + git diff | high | R1-L1 | 2026-07-17
--bare skip list + future -p default | [^HeadlessDocs] code.claude.com/docs/en/headless | high | 1 | 2026-07-17
skills/custom commands work in -p | [^HeadlessDocs] | high | 1 | 2026-07-17
bg workflow wait cap 10min v2.1.182+ / CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS / 0=no limit | [^HeadlessDocs] | high | 1 | 2026-07-17
system/init plugins+plugin_errors, fail-CI guidance | [^HeadlessDocs] | high | 1 | 2026-07-17
json output total_cost_usd + per-model breakdown | [^HeadlessDocs] | high | 1 | 2026-07-17
--resume session id, same-directory scope | [^HeadlessDocs] | high | 1 | 2026-07-17
bare skips OAuth/keychain, API-key/apiKeyHelper auth | [^HeadlessDocs] | high | 1 | 2026-07-17
10MB stdin cap | [^HeadlessDocs] | high | 1 | 2026-07-17
dontAsk auto-deny non-allowlisted, locked-down CI | [^HeadlessDocs]+[^CliReference] | high | 1 | 2026-07-17
--max-turns/--max-budget-usd/--strict-mcp-config/--plugin-dir/--fallback-model/--settings/--json-schema flag descriptions | [^CliReference] | high | 1 | 2026-07-17
exit code "0/1" specifically | [^CliReference] | LOW (page has no exit-code docs; L2-F2) | 1 | 2026-07-17
-p workspace trust "no dialog appears and the rules stay ignored" | [^PermissionsDoc] | high | 1 | 2026-07-17
"Permission rules are enforced by Claude Code, not by the model" | [^PermissionsDoc] | high | 1 | 2026-07-17
/loop session-scoped, 7-day expiry, no catch-up for missed fires | [^ScheduledTasks] | high | 1 | 2026-07-17
disable-model-invocation scheduled fires reach Claude as plain text (v2.1.196+) | [^ScheduledTasks] | high | 1 | 2026-07-17
Desktop task per-task permission config / cloud >=1h interval | [^ScheduledTasks] | high | 1 | 2026-07-17
routines: autonomous no-prompts, claude/ push restriction, fresh clone, daily cap + rejected-until-reset, connectors default-included, identity attribution, claude.ai-login-only, green-status warning, research-preview volatile | [^RoutinesDocs] | high | 1 | 2026-07-17
issue 76239 OPEN + ~2s effective wait | [^McpHeadlessBugs] gh+body | high | 1 | 2026-07-17
issue 68375 OPEN, strict-mcp-config workaround | [^McpHeadlessBugs] gh | high | 1 | 2026-07-17
issue 32191 status = CLOSED DUPLICATE (report body said OPEN; L2-F1) | [^McpHeadlessBugs] gh | LOW-as-cited | 1 | 2026-07-17
issue 66395 [DOCS] CLOSED NOT_PLANNED; regression window title-asserted | [^WindowsHang] gh | medium | 1 | 2026-07-17
issue 837 CLOSED COMPLETED; 14246 CLOSED DUPLICATE (supersession story holds) | [^SlashHeadlessIssues] gh | high | 1 | 2026-07-17
issue 23707 exists, CLOSED NOT_PLANNED | [^WebSandbox] gh | high | 1 | 2026-07-17
GHA schedule delayed/dropped under high load, 5-min minimum | [^GhaSchedule] GitHub docs | high | 1 | 2026-07-17
GHA 5-30min typical delay figures | [^GhaSchedule] community | medium (discussions not fetched; labeled community-measured) | 1 | 2026-07-17
systemd Persistent= verbatim | [^MissedRun] man7.org (freedesktop 403) | high | 1 | 2026-07-17
Task Scheduler StartWhenAvailable missed-start semantics | [^MissedRun] learn.microsoft.com | high | 1 | 2026-07-17
anacron interval-since-last-run | [^MissedRun] | medium (not fetched; siblings verified) | 1 | 2026-07-17
probes P1/P2 output figures | [^HeadlessProbe] lane-3 transcript + live claude --version 2.1.212 | medium (ephemeral self-report, blue-acknowledged, fix offered) | 1 | 2026-07-17
smoke mode 1-lane/1-round/1-citation-pass/haiku/~50k | [^ResearchCommand]+[^Backlog 17] pinned | high | 1 | 2026-07-17
keeper runs omit model; judgmentModel inherits session | [^ResearchCommand] pinned | high | 1 | 2026-07-17
script-vs-prose doctrine quote | [^ResearchCommand] pinned | high | 1 | 2026-07-17
stop-and-resume practice + capture step artifacts | [^ResearchCommand] pinned | high | 1 | 2026-07-17
backlog items 10/17/27c/34/36/39 quotes incl. qmd daemon ladder | [^Backlog]/[^QmdDaemon] pinned | high | 1 | 2026-07-17
qmd fallback Grep/Read quote | [^QmdFallback] pinned run-4 friction | high | 1 | 2026-07-17
smoke UNVERIFIED + Catechism template friction | [^SmokeRecord] pinned | high | 1 | 2026-07-17
IdeaStudy novel p<0.05 / weaker feasibility / self-eval+diversity failures / human rerank improved | [^IdeaStudy] arXiv abs+html 2409.04109 | high | 1 | 2026-07-17
Reflexion feedback->verbal reflection->episodic memory->later trials | [^Reflexion] arXiv abs 2303.11366 | high | 1 | 2026-07-17
DGM empirically validates each change; archive admission = compile+edit-ability, all agents benchmark-evaluated | [^DGM] arXiv abs+html 2505.22954 | high | 1 | 2026-07-17
DGM footnote quote "improve themselves the more compute they are provided" | [^DGM] | LOW (absent from abs and html; L2-F3) | 1 | 2026-07-17
DGMSakana fake-logs / markers-removal / lineage / sandbox quotes | [^DGMSakana] sakana.ai/dgm | high | 1 | 2026-07-17
port-plan quotes (human approves each step; Phase-4 verify; decision 6) | [^PortPlan] AgentOrange working tree | high-content / pin-absent (L2-F4: 7bc501e path MISSING, PINNED.md wrong) | 1 | 2026-07-17

## round 2 — lens 4 (slice: §6/§7/§8 + [^HooksJson])

§6 row 8 #76239 OPEN — stdio MCP tools silently missing, startup >~2s, regression since 2.1.144 | [^McpHeadlessBugs] live WebFetch | high (drift re-check, still OPEN) | r2 | 2026-07-17
§6 row 8 #68375 OPEN — stdio tool-call hang w/ full MCP fleet, --strict-mcp-config workaround; title notes regression 2.1.177 | [^McpHeadlessBugs] live WebFetch | high (drift re-check, still OPEN) | r2 | 2026-07-17
§6 row 9 "Bootstrap guard already shipped in hooks.json" | [^HooksJson] @7bc501e | high (round-1 pin unchanged, not re-fetched) | r2 | 2026-07-17
§7 Pattern B/E "pricing figures graded MEDIUM" | vs §5.2/[^Pricing] R1-11 upgrade to leaf-HIGH | stale self-report → L4-F1 (LOW): §7 self-audit not repaired for pricing MEDIUM→HIGH | r2 | 2026-07-17
§8 OQ8 STOP figures (0.42%/0.46%, insignificantly higher, 10,000, syntactic) | [^STOP] ar5iv (round-1 HIGH) | high (unchanged, not re-fetched) | r2 | 2026-07-17

## round 1 — lens 3 (slice: §4/§5 + referenced footnotes)

dontAsk auto-denies non-allowlisted; deny→ask→allow; "If a tool is denied at any level, no other level can allow it"; CLI args below managed settings | [^PermissionsDoc] | high | r1 | 2026-07-17
Bash rule syntax `Bash(git push *)` valid (doc's own deny example); `:*` suffix = trailing-space-star equivalent | [^PermissionsDoc] | high | r1 | 2026-07-17
Write(path) rules accepted-never-matched + v2.1.210 startup warning; Edit rules cover all file-editing tools | [^PermissionsDoc] | high | r1 | 2026-07-17
Read deny blocks Edit incl. file creation, v2.1.208 (Write/NotebookEdit NOT covered — doc says add Edit deny) | [^PermissionsDoc] | high | r1 | 2026-07-17
File rules don't cover arbitrary subprocesses; sandbox = OS-level complement; blocking hook precedence over allow (exit-2 before permission eval); hooks can't loosen deny/ask | [^PermissionsDoc]+[^HooksDoc] | high | r1 | 2026-07-17
disableBypassPermissionsMode is a `permissions.`-scoped key ("permissions.disableBypassPermissionsMode"), any scope, self-lockout quote verbatim — report §4.2 JSON places it at TOP level → silent no-op (L3-F1, MEDIUM) | [^PermissionsDoc] | high (report JSON defective) | r1 | 2026-07-17
Windows /c/... normalization; `//path` absolute anchor; bare-relative path rules anchor at CURRENT DIRECTORY, `/path` at settings source — §4.2 deny set is cwd-dependent as written (L3-F6, LOW-MEDIUM) | [^PermissionsDoc] | high | r1 | 2026-07-17
#6631 closed; fix claimed via #4467 ("fix merged" ant-kurt); reporter re-confirmed bypass at v1.0.93 Aug 2025 — report account exact | [^DenyRWIssue] live fetch | high | r1 | 2026-07-17
#25621 closed as duplicate, canonical not named on page (report labels accordingly) | [^DenyBashIssue] live fetch | high | r1 | 2026-07-17
AI Scientist incidents: "edited the code to perform a system call to run itself... endlessly calling itself"; "modify its own code to extend the timeout period"; "mitigated by sandboxing" | [^AIScientist] sakana.ai/ai-scientist | high (upgrades blue's MEDIUM) | r1 | 2026-07-17
OWASP Excessive Agency LLM06:2025: excessive functionality/permissions/autonomy; least privilege; "human-in-the-loop control to require a human to approve high-impact actions"; "Implement authorization in downstream systems rather than relying on an LLM"; logging; rate-limiting | [^OWASP] genai.owasp.org | high | r1 | 2026-07-17
AI Control 2312.06942: Greenblatt/Shlegeris/Sachan/Roger; protocols evaluated "if the model is itself intentionally trying to subvert them" | [^AIControl] arXiv abs | high | r1 | 2026-07-17
Hooks: agent_id/agent_type subagent fields; PreToolUse exit-2 + permissionDecision:"deny"; settings hook edits picked up by file watcher; plugin hooks/hooks.json merge | [^HooksDoc] | high | r1 | 2026-07-17
Usage/Cost Admin API: usage_report/messages + cost_report endpoints; Admin key required; "unavailable for individual accounts"; ~5-min freshness; 1/min sustained polling | [^UsageAPI] leaf fetch (upgrades blue's MEDIUM search-digest grade) | high | r1 | 2026-07-17
Workspace SPEND limits: Console Limits tab only, no API read/set documented; BUT Rate Limits API (/v1/organizations/rate_limits + workspaces/{id}/rate_limits) READS org+workspace rate limits, read-only, Admin key — §5.1 flat "no endpoint to read" is stale for rate limits (L3-F2, MEDIUM) | [^ConsoleLimits] + platform.claude.com/docs/en/manage-claude/rate-limits-api | high | r1 | 2026-07-17
Pricing leaf-verified: Haiku 4.5 $1/$5; Sonnet 4.5/4.6 $3/$15 (Sonnet 5 intro $2/$10 → $3/$15 from 2026-09-01); Fable/Mythos 5 $10/$50; Opus 4.5–4.8 $5/$25 (report "frontier ~$10/$50" true only of Fable-class → L3-F4); Batch flat 50%; cache read 0.1x; Fable/Mythos/Sonnet-5 tokenizer ~+30% tokens | [^Pricing] platform pricing page | high (Batch ≤24h window NOT on page — medium, batch-processing page unfetched) | r1 | 2026-07-17
--fallback-model NOT print-only in current cli-reference (persistent fallbackModel setting documented) — report footnote label wrong (L3-F3, LOW) | [^CliReference] | high | r1 | 2026-07-17
cost.md: $414.97 / 42 agents / 1975 api-turns / cache 99% / judgment premium cache-RATE-driven — all verbatim | [^CostRecord] @7bc501e | high | r1 | 2026-07-17
efficiency-phase.md: run-3 $149.95 (line 6); "dwarfing every lever" (line 46); attestation ceiling; revisit triggers; PR-C mirror | [^EfficiencyPlan] @7bc501e | high | r1 | 2026-07-17
semantic-consent final clause verbatim (report truncates trailing parenthetical; meaning intact) | [^SemanticConsent] @7bc501e | high | r1 | 2026-07-17
push-freeze-guard contract comment verbatim ("It NEVER blocks — the freeze is a commitment the human may consciously override...") | [^PushGuard] @7bc501e | high | r1 | 2026-07-17
friction.md run-4 §4/§5 leans: "The MUST-try clause has no observable" (line 14, round-0 false-paywall specimen); ABORT DISCLOSURE line 48 verbatim | [^FrictionRun4] @7bc501e | high | r1 | 2026-07-17
STOP §6.2/Table 2 re-verified independently at ar5iv: 0.42% (0.31–0.57%) / 0.46% (0.35–0.61%) / 10,000 / syntactic `use_sandbox=False`|`exec(` — concurs with lens 1/lens 4 lines | [^STOP] ar5iv | high | r1 | 2026-07-17
#22055 thread: PreToolUse exit-2 protected-files workaround verbatim in thread; chmod seen only inside a user's settings snippet, no chmod-444 recommendation found — concurs with L4-F2 direction | [^PermAskBypass] gh --comments | high (workaround); chmod-444 sub-claim low | r1 | 2026-07-17

## round 2 — lens 1 (slice: preamble/§0/§1/§2 + referenced footnotes)

§1.2 "31h ... two consecutive runs" (R1-2 repair) — line 31 sub-item (h) verbatim; line-scheme validated by anchors 27c=line 27, item 39=line 39 | git show 7bc501e:ideas/backlog.md | high | r2 | 2026-07-17
§1.3/[^IdeasCorpus] 25 statused checkbox items / 39 lines (R1-1 repair) — recounted | git show 7bc501e:ideas/backlog.md | high | r2 | 2026-07-17
[^Backlog] "15–17 ... assemble-on-failure" — (c) is line 18 at pin; content verbatim, range short by one → L1-F2 | git show 7bc501e:ideas/backlog.md | high-content / low-as-pinned | r2 | 2026-07-17
§1.3/[^RedPatterns] 1,557 lines (R1-30 repair) — wc: 1557 lines / 119,418 bytes | local wc | high | r2 | 2026-07-17
[^SICA] Comments "Submitted as a preprint to NeurIPS 2025" (R1-4 repair) + 17–53% SWE Bench Verified | arXiv abs 2504.15228 live fetch | high | r2 | 2026-07-17
§1.2 STOP "~page-of-code seed improver" — paraphrase; "simply prompts the language model...", "minimizing initial prompt complexity" corroborate; literal "page" absent | ar5iv 2310.02304 live fetch | medium (color) | r2 | 2026-07-17
[^PortPlan] decision 6 ("daily default once scheduled ... scheduling always human-opt-in"), "human approves each step" (line 287), guardrail lines 291–292, Phase-4 verify line 331 — all verbatim | AgentOrange docs/claude-port-plan.md working tree = 6df52af (clean) | high-content / snapshot-grade (pin defect R1-7 standing) | r2 | 2026-07-17
Propagation-grep spot-check: `40 statused`/`1,558`/`Bash(node`/`two consecutive runs` — only correction/mention contexts standing | grep report.md | high | r2 | 2026-07-17
CHANGELOG claim_count arithmetic (124+18=142; 8 steps + 10 stub fields = 18) — recounted | report §2.2/§2.3 | high | r2 | 2026-07-17

## round 2 — lens 2 (slice: §2/§3 + referenced footnotes)

§3.2 "--json-schema ... invalid schema exits with error ≥2.1.205" | [^CliReference] live fetch (verbatim: "exits with an error on an invalid schema"; "Before v2.1.205 ... no error"; min-version 2.1.205 marker) | high | r2 | 2026-07-17
§3.3 #66395 regression window v2.1.161–v2.1.168, fixed v2.1.169 | [^WindowsHang] gh issue body (quotes v2.1.169 changelog: "regression in 2.1.161"; span "between v2.1.161 (June 2, 2026) and v2.1.168 (June 6, 2026)") | high — UPGRADE from r1 medium (body now fetched; footnote's "body not fetched" caveat retirable) | r2 | 2026-07-17
§3.4 rung-3 "~973MB models" | [^QmdDaemon] backlog item 34 @7bc501e via git show ("Total model footprint if daemon-only: 973MB (embed+rerank), not 2.2GB") | high | r2 | 2026-07-17
§3.4 web sessions "isolated Anthropic-managed VM per session, credentials held outside the sandbox" | [^WebSandbox] code.claude.com/docs/en/sandbox-environments live fetch (verbatim: "isolated, Anthropic-managed virtual machine"; "a separate proxy holds your GitHub token outside the sandbox") | high — UPGRADE from surveyed; doc names the GitHub token specifically | r2 | 2026-07-17
§3.4 anacron interval-since-last-run | [^MissedRun] man7.org anacron(8) live fetch ("checks whether this job has been executed in the last n days"; "does not assume that the machine is running continuously") | high — UPGRADE from r1 medium | r2 | 2026-07-17
§3.4 GHA 5–30min community delay figure | [^GhaSchedule] github.com/orgs/community/discussions/52477 live fetch (delay/drop docs language confirmed in-thread; one measured report "close to an hour" at 00:00 — worse than the report's range, strengthening the design conclusion) | medium on the specific figure (community-grade by nature; report labels it so — no gap); high on the phenomenon | r2 | 2026-07-17
§2.4 "DGM only admits a change to its archive after empirical validation against a benchmark" + "analogy is exact" | [^DGM] abs (r1 leaf: abstract "empirically validates each change"; r1 ledger: archive admission = compile+edit-ability) | MEDIUM — ordering true, gating implication overstated; DGM retains low scorers → L2-F1 | r2 | 2026-07-17
§3.3 R1-5 repair text (two-open/one-closed restatement) + §3.2 R1-6 repair text (any-nonzero-is-failure) | re-checked vs live cli-reference + ledger statuses | high (repairs faithful, no regression) | r2 | 2026-07-17
Carried without re-fetch (same-day access, claims verified HIGH r1): [^HeadlessDocs] §3.2 set, [^CliReference] flag texts, [^PermissionsDoc] trust-dialog, [^ScheduledTasks] /loop + disable-model-invocation, [^RoutinesDocs] rung-3 set, [^McpHeadlessBugs] #76239/#68375, [^SlashHeadlessIssues], [^QmdDaemon] ladder, [^QmdFallback], [^PortPlan] (pin-absent caveat stands, lead-docketed), [^IdeaStudy], [^Reflexion], [^SmokeRecord], [^Backlog] smoke shape, [^EfficiencyPlan], [^HeadlessProbe] (medium, disposition-of-record: re-run at build) | r2 | 2026-07-17

## round 2 — lens 3 (slice: §5/§6 + referenced footnotes)

§5.2 pricing re-fetched (no drift, same access date): Haiku 4.5 $1/$5; Sonnet 4.5/4.6 $3/$15; Sonnet 5 intro $2/$10→$3/$15 (2026-09-01); Opus 4.5–4.8 $5/$25; Fable/Mythos 5 $10/$50; Batch flat 50%; cache read 0.1×; +30% tokenizer note verbatim (Sonnet 4.6 = legacy boundary) | [^Pricing] platform pricing page (live) | high | r2 | 2026-07-17
§5.2 Batch ≤24h async window RESOLVED — "Batches expire if processing does not complete within 24 hours"; "the 24-hour processing window" | batch-processing page (live) | high (round-1 MEDIUM sub-claim → HIGH available) | r2 | 2026-07-17
§5.2 $414.97 (cost.md) + run-3 $149.95 (efficiency-phase.md §I) bundled under single [^CostRecord] marker → L3-F2 LOW (footnote discloses split; both figures verified) | [^CostRecord]+[^EfficiencyPlan] | high (figures); low (bundling nit) | r2 | 2026-07-17
§6 row 5 "Certain (no API)" retains pre-R1-9 flat shorthand; §5.1 rate-limit-API-readable requalification (R1-9) did not propagate to the cell → L3-F1 LOW (conclusion unaffected) | [^RateLimitsAPI]/§5.1 internal | high (underlying claim) | r2 | 2026-07-17
§5/§6 already-HIGH doc/issue leaves NOT re-fetched (verified HIGH r1, byte-stable section, same access date, no drift possible): [^CliReference] budget/turn flags, [^HeadlessDocs] json cost, [^UsageAPI]/[^RateLimitsAPI]/[^ConsoleLimits] quota set, [^RoutinesDocs] cloud caps, [^McpHeadlessBugs] row-8 statuses, [^HooksJson] bootstrap guard, [^PushGuard], [^QmdFallback], [^EffReport] severity-floor-rejected, [^FrictionRun4] run-4 death, [^HeadlessProbe] P1 cost | high (carried) | r2 | 2026-07-17

## round 3 — lens 1 (slice: preamble/§0/§1 + referenced footnotes)

preamble/§0 "invariant 7 ... per the lead's direction" | debate.md round-2 ### LEAD (line 344: "derive the four fixes from it, and add it to §0's invariants") | high | r3 | 2026-07-17
§0 R2-19 artifact enumeration total over printed tree | recount: tree 8 entries = 3 code + 2 commands + scheduling doc + skill file + manifest | high | r3 | 2026-07-17
§1.5 R2-5 "the wrapper itself now creates [the sub-run dir] when it runs setup-research-run.mjs wrapper-side" | setup-research-run.mjs at plugins/frank-exchange-of-views/skills/research-protocol/scripts/ — takes `<runDir>` argv, mkdirSync recursive skeleton; git diff vs 7bc501e empty at that path | high (stamp-at-creation buildable as claimed) | r3 | 2026-07-17
§1.5 R2-5 run-window fallback "(ledger timestamps)" | internal cross-read vs §2.2 steps 0/7 (ledger record is step-7-only; aborted/DEAD runs are DEFINED by its absence) | LOW — fallback void for dead runs → L1-F1 | r3 | 2026-07-17
§1.5 R2-6 "every infrastructure class ... ALSO surfaces independently on the doctor/dead-man line" | internal cross-read vs §3.4 (surface carries LAST reason only; intermittent infra events overwritten by next success) | MEDIUM — overbroad for intermittent classes → L1-F3 | r3 | 2026-07-17
§1.4 R2-18 "expected ~$0.10–0.50/night; 30 × expected ≈ $3–15/mo; cap-trip = anomaly" | arithmetic recomputed; P2 $0.058 consistent with §3.1/[^HeadlessProbe] (P2 figure stays MEDIUM, ephemeral-instrument disposition-of-record: re-run at build) | high (arithmetic) / medium (P2 anchor) | r3 | 2026-07-17
§1.4 R2-11 graduation-queued exemption | internal consistency vs §2.3 status enum + §6 row 3 — all three sites consistent | high | r3 | 2026-07-17
§1.3 telemetry row "(shipping in FEOV 0.6.0 per the ratified efficiency plan)" | this run's own trajectories/board-telemetry.jsonl EXISTS (FEOV 0.7.0 shipped it); plans/efficiency-phase.md does say 0.6.0 (lines 65/72) — plan quote faithful, tense stale | LOW as-written (drift) → L1-F2 | r3 | 2026-07-17
[^AlertFatigue] "under 1 in 5 acted on" | fresh WebSearch r3: 2026 survey (n=1,039 SRE/DevOps, Feb 2026) "57% report fewer than 30% of alerts actionable"; ACM Computing Surveys 10.1145/3723158 (SOC alert fatigue) | LOW on report's specific number (unchanged, blue self-grade honest); phenomenon upgradeable to HIGH via citation swap → L1-F4 note | r3 | 2026-07-17
Carried without re-fetch (verified HIGH r1/r2, sections unchanged r2, immutable pins or same-day access, ≤2 rounds elapsed): [^SelfCorrect], [^Reflexion], [^Voyager], [^DGM]+[^DGMSakana], [^SICA], [^STOP] (the "~page-of-code" paraphrase stays MEDIUM color per r2), [^Dependabot], [^DependabotFatigue], [^Goodhart] (qualitative-only, blue-labeled), [^FrictionRun3], [^FrictionRun4], [^Backlog], [^IdeasCorpus], [^EffReport], [^EfficiencyPlan], [^ResearchCommand], [^RedPatterns], [^PortPlan] (snapshot-grade; pin-absent defect standing, lead-docketed; AgentOrange working tree still 6df52af-clean per session git status) | r3 | 2026-07-17

## round 3 — lens 2 (slice: §2/§3 + referenced footnotes)

§2.2 step 4/§4.3 R2-3 locus: step 2 session-Bash node setup, step 3 Workflow scriptPath=debate.js, step 5 session-Bash node capture; shipped 0.7.0 copy byte-identical to pin | plugin cache research.md + git show 7bc501e | high | r3 | 2026-07-17
[^ResearchCommand] --smoke 1-lane/1-round/haiku/~50k; keeper-omit-model; stop-and-resume; capture emits cost.md+run-record-audit.md; script-vs-prose quote | same file (verbatim) | high | r3 | 2026-07-17
§2.2 step 0/§3.2 R2-1: --input-format <format> "only works with --print", values text/stream-json | claude --help live, CLI 2.1.212 | high | r3 | 2026-07-17
§2.4 R2-21 DGM repair: admission = compile+retain-edit-ability; low scorers retained ("stepping stones"); parent selection ~proportional to performance; every agent benchmark-evaluated | arxiv.org/html/2505.22954 live fetch | high (repair clean, no regression) | r3 | 2026-07-17
§2.2 step 3 "--json-schema leg" of the phase drive | headless docs live: structured_output documented under --output-format json ONLY; stream-json/mid-turn composition undocumented; --help silent | LOW as-asserted → L2-F1 | r3 | 2026-07-17
§3.2 dontAsk "auto-denies anything not allow-listed" | headless docs live ("...or the read-only command set") + permissions doc #read-only-commands (ls, cat, echo, pwd, head, tail, grep, find, wc, which, diff, stat, du, cd, read-only git — "in every mode", not configurable; deny rules can re-cover) | REFUTED as stated → L2-F2 | r3 | 2026-07-17
§3.2 [^HeadlessDocs] re-fetch: --bare future default; 10MB stdin; bg wait 10min v2.1.182+; system/init plugins+plugin_errors; total_cost_usd; --json-schema invalid-schema error (pre-2.1.205 silent) | headless docs live | high (zero drift) | r3 | 2026-07-17
§3.4 R2-10 "k×$5/night... ~3 nights at the ceiling" | internal recompute vs §3.4 one-resume-per-nightly-fire semantics (≤$5/night; ~10 nights to cap) | LOW → L2-F3 | r3 | 2026-07-17
Carried without re-fetch (HIGH r1/r2, same-day access, claim text unchanged): [^IdeaStudy], [^SmokeRecord], [^ScheduledTasks], [^McpHeadlessBugs] #76239/#68375, [^WindowsHang], [^SlashHeadlessIssues], [^WebSandbox], [^GhaSchedule], [^MissedRun], [^QmdDaemon], [^QmdFallback], [^HeadlessProbe] (medium, re-run at build) | r3 | 2026-07-17

## round 3 — lens 3 (slice: §4/§5 + referenced footnotes; no blue round-3 revision — spot-check volatile/load-bearing leaves)

§5.2 pricing zero-drift re-fetch (VOLATILE, footnote self-flags "re-fetch at citation-verification"): Haiku 4.5 $1/$5; Sonnet 4.5/4.6 $3/$15; Sonnet 5 intro $2/$10→$3/$15 (2026-09-01); Opus 4.5–4.8 $5/$25; Fable/Mythos 5 $10/$50; Batch flat 50%; cache read 0.1× | [^Pricing] platform pricing page live | high (unchanged) | r3 | 2026-07-17
§5.2 tokenizer "+30%" scope: page names Opus 4.7+ AND Fable/Mythos/Sonnet-5; report lists only Fable/Mythos/Sonnet-5 → L3-F1 (LOW, completeness not fidelity; claim-as-written TRUE, immaterial to $414.97/$149.95 approximate-anchor caveat) | [^Pricing] | high-as-written / low-completeness | r3 | 2026-07-17
§4.1 AI Control 2312.06942: title + authors (Greenblatt/Shlegeris/Sachan/Roger) + "intentionally trying to subvert them" quote; page notes ICML version (OpenReview) — venue "ICML 2024" now corroborated (r1 verified only authors+quote) | [^AIControl] arXiv abs live | high (venue upgrade) | r3 | 2026-07-17
§4.1 #22055 CLOSED NOT PLANNED, title verbatim; WebFetch page-only missed comments (INSUFFICIENT alone) | [^PermAskBypass] WebFetch | high (status) | r3 | 2026-07-17
§4.3 layer 3 #22055 PreToolUse exit-2 protected-files workaround verbatim in thread ("hooks run at the process level … exit 2 = tool call rejected"); chmod present only in a `Bash(chmod:*)` allow snippet, NOT a chmod-444 rec → R1-13 stands | [^PermAskBypass] gh issue view 22055 --comments | high | r3 | 2026-07-17
§4/§5 already-HIGH leaves carried (verified r1/r2, ≤2 rounds, same-day, sections unchanged): [^PermissionsDoc] enforced-not-by-model + deny-precedence, [^DenyRWIssue] #6631, [^DenyBashIssue] #25621, [^STOP] 0.42%/0.46%, [^AIScientist], [^DGMSakana], [^OWASP] (self-graded MEDIUM taxonomy), [^HooksDoc], [^IdeasCorpus] hooks-fire-on-subagent, [^ScheduledTasks] disable-model-invocation, [^SemanticConsent], [^PushGuard], [^CostRecord] $414.97/$149.95, [^CliReference] budget/turn/fallback flags, [^UsageAPI], [^RateLimitsAPI], [^ConsoleLimits], [^EfficiencyPlan] attestation ceiling, [^HeadlessProbe] P1/P2 | high (carried) | r3 | 2026-07-17

## round 3 — lens 4 (slice: §6/§7/§8 + [^HooksJson])

§6 row 8/§7 #76239 OPEN (regression since 2.1.144) | [^McpHeadlessBugs] live gh | high (drift re-check r3, still OPEN) | r3 | 2026-07-17
§6 row 8/§7 #68375 OPEN (full MCP fleet hang, regression 2.1.177, --strict-mcp-config workaround) | [^McpHeadlessBugs] live gh | high (drift re-check r3, still OPEN) | r3 | 2026-07-17
§4.1/§7 #22055 CLOSED NOT_PLANNED | [^PermAskBypass] live gh | high (drift re-check r3) | r3 | 2026-07-17
§7 #66395 CLOSED NOT_PLANNED ([DOCS], span 2.1.161-2.1.168 fixed 2.1.169) | [^WindowsHang] live gh | high (drift re-check r3) | r3 | 2026-07-17
§3.3/§6/§7 #32191 CLOSED DUPLICATE | [^McpHeadlessBugs] live gh | high (drift re-check r3) | r3 | 2026-07-17
§3.4/§7 #23707 CLOSED NOT_PLANNED | [^WebSandbox] live gh | high (drift re-check r3) | r3 | 2026-07-17
§7 #837 CLOSED COMPLETED / #14246 CLOSED DUPLICATE | [^SlashHeadlessIssues] | high-carried from r1 (live gh classifier-blocked x3; closed=low-volatility, story unaffected) | r3 | 2026-07-17
§6 row 4/§4.3 layer 4(i) "pinned read-only git argv set / no rule grants argv choosing a subprocess write target" | Bash(git log *) allow rule vs CLI behavior | LOW — REFUTED at leaf: `git log --output=<path>` wrote /tmp file exit 0; write gadget present w/o any bug; not covered by OQ18 → L4-F1 | r3 | 2026-07-17
§7 R2-14 Pattern B/E stale-grade fix (round-2 lens-4 L4-F1) | §7 bullet + §5.2/[^Pricing] | high — CLOSED (bullet now "upgraded to leaf-verified HIGH round 1, R1-11") | r3 | 2026-07-17
§6 row 9 [^HooksJson] bootstrap guard | @7bc501e pin | high (pin-immutable, carried r1) | r3 | 2026-07-17
§7/§8 OQ8 STOP figures 0.42%/0.46%/10,000/syntactic | [^STOP] ar5iv | high (non-volatile academic, carried r1/r2) | r3 | 2026-07-17
§7 R2 upgrades [^WindowsHang]/[^WebSandbox]/[^MissedRun]/[^Pricing] Batch 24h/--json-schema>=2.1.205/~973MB | per-footnote | high (carried r2 lens 2/3, 1 round elapsed, stable) | r3 | 2026-07-17
[^ResearchCommand] --input-format stream-json print-only + mixed locus | claude --help 2.1.212 + doubts.md @7bc501e | high (incidental re-confirm) | r3 | 2026-07-17

## round 4 — lens 2 (slice: §2/§3 + referenced footnotes)

§3.2/[^PermissionsDoc] R3-14 carve-out quote — full 15-command set incl. read-only git, "without a permission prompt in every mode", "not configurable", ask/deny-per-command remedy | permissions doc live WebFetch | high (verbatim, zero drift vs r3) | r4 | 2026-07-17
[^PermissionsDoc] unquoted-glob rule ("every flag is read-only"; find/git still prompt on unquoted globs — doc also names sort, sed) | permissions doc live | high | r4 | 2026-07-17
[^PermissionsDoc] exec wrappers watch/setsid/ionice/flock + find -exec/-delete always prompt | permissions doc live | high | r4 | 2026-07-17
[^PermissionsDoc] deny→ask→allow order + "a deny rule can't carry allowlist exceptions" + "If a tool is denied at any level, no other level can allow it" | permissions doc live | high | r4 | 2026-07-17
[^PermissionsDoc] dontAsk "Auto-denies tools unless pre-approved via /permissions or permissions.allow rules" | permissions doc live | high | r4 | 2026-07-17
NEW LEAF: "Read and Edit deny rules apply ... to file commands Claude Code recognizes in Bash, such as cat, head, tail, and sed" — grounds L2-F2: §4.2/§6-row-13 "cat .claude/projects would have been AUTO-APPROVED under the round-2 profile" REFUTED for that named target (round-2 profile carried the Read deny; exposure was real only for allow-scoped-not-denied paths) | permissions doc live | high (doc clause) / LOW as-cited in report | r4 | 2026-07-17
§2.2 step 3 R3-1 repair (json-schema mid-drive demoted, fenced-block fallback) | vs r3 lens-2 leaf (headless docs) | high (repair faithful, no regression) | r4 | 2026-07-17
§2.2 step 0 R3-1 degrade-note "named readers" (§2.3 confidence field + doctor line) | internal cross-read: neither reader site specifies it → L2-F1 (LOW) | low (reader-lag) | r4 | 2026-07-17
§3.4 R3-12 arithmetic (≤$5/night; cap ~10 nights; HALT night 12; cap-first in ceiling case) | recomputed | high | r4 | 2026-07-17
§3.4 R3-2 rung-0 row + gate-table R0 cells vs §2.2 --manual clause | internal | high | r4 | 2026-07-17
§2.3 R3-13 status enum vs §1.4/§6 row 3 | internal, three sites | high | r4 | 2026-07-17
§3.4 R3-11 signature-normalization spec (exit class + templated line + placeholders; zero-firings telemetry) | internal, spec-level | high | r4 | 2026-07-17
Carried without re-fetch (HIGH r1–r3, same-day access, claim text unchanged): [^McpHeadlessBugs] #76239/#68375, [^HeadlessDocs], [^CliReference] --input-format + flags, [^ScheduledTasks], [^RoutinesDocs], [^WindowsHang], [^SlashHeadlessIssues], [^WebSandbox], [^GhaSchedule], [^MissedRun], [^QmdDaemon], [^QmdFallback], [^ResearchCommand], [^Backlog], [^IdeaStudy], [^SmokeRecord], [^Reflexion], [^DGM], [^IdeasCorpus], [^HooksDoc], [^PortPlan] (pin-defect standing), [^HeadlessProbe] (medium, re-run at build) | r4 | 2026-07-17

## round 4 — lens 1 (slice: preamble/§0/§1 + referenced footnotes)

§1.2 "struggle to self-correct ... performance even degrades" | [^SelfCorrect] arXiv abs 2310.01798 live re-fetch (verbatim, zero drift) | high | r4 | 2026-07-17
§1.2 oracle-feedback dependency | [^SelfCorrect] ar5iv body (r1 leaf; immutable source, abstract re-confirmed r4) | high (carried) | r4 | 2026-07-17
§1.2 Reflexion verbal reflections → episodic memory buffer → subsequent trials | [^Reflexion] arXiv abs 2303.11366 live ("maintain their own reflective text in an episodic memory buffer to induce better decision-making in subsequent trials") | high | r4 | 2026-07-17
§1.2 Voyager "environment feedback, execution errors, and self-verification" + ever-growing skill library + "alleviates catastrophic forgetting" | [^Voyager] arXiv abs 2305.16291 live | high (all verbatim) | r4 | 2026-07-17
§1.2/§2.4 DGMSakana five quotes incl. "improve themselves the more compute they are provided" + markers-removal + lineage + sandboxed-under-supervision | [^DGMSakana] sakana.ai/dgm live | high (zero drift) | r4 | 2026-07-17
§1.1 Dependabot 11.3% deprecated + configure-toward-fewer-notifications | [^Dependabot] arXiv abs 2206.07230 live | high (verbatim) | r4 | 2026-07-17
§1.1 DependabotFatigue alert-fatigue framing | [^DependabotFatigue] arXiv abs 2502.06175 live | high | r4 | 2026-07-17
[^DependabotFatigue] ">75M PRs generated in 2022" | NOT in abstract; MUST-try /html/2502.06175 → exact sentence in Introduction (paper's own source = Forbes footnote, press-derived — nuance noted, no gap) | high fidelity | r4 | 2026-07-17
§1.3 R3-16 "SHIPPED as of FEOV 0.7.0 — present in this run's own trajectories/" | Glob: trajectories/board-telemetry.jsonl EXISTS | high | r4 | 2026-07-17
preamble/§0 "invariant 8 ... added at the lead's direction" | debate.md round-3 ### LEAD line 540 (verbatim direction) | high | r4 | 2026-07-17
preamble "all 17 round-3 gaps addressed" + CHANGELOG claim_count 151 (49+38+46=133 + 18) | red/ledger.md OPEN GAPS recount = 17; arithmetic recomputed; §2.3 stub fields recount = 10 | high | r4 | 2026-07-17
§1.5 R3-3/R3-5 + R3-6-leg repair texts | internal cross-read vs §2.2 steps 0/7 + §3.4 (consistent, present) | high | r4 | 2026-07-17
§1.5 R3-4 repair (START at step 0, END at any observed exit, DEAD-mark sweep) | internal cross-read | high as-printed; residual → L1-F1 (unobserved wrapper exit ⇒ multi-day window over-sweeps human dirs) | r4 | 2026-07-17
§0 R3-10 enumeration extension | direct read; count headline "exactly THREE new code artifacts" unreconciled with enumerated SessionStart executable → L1-F2 | high (items present) / low (count claim) | r4 | 2026-07-17
§1.4 R3-2 leg + R3-13 leg | cross-read vs §3.4 row 0 + §2.3 status enum | high | r4 | 2026-07-17
Carried (immutable pins or ≤2-round leaf checks): [^FrictionRun3], [^FrictionRun4], [^Backlog], [^IdeasCorpus], [^EffReport], [^EfficiencyPlan], [^ResearchCommand], [^RedPatterns], [^SmokeRecord] (@7bc501e), [^SICA] (r2), [^STOP] (r1×3+r2; "~page-of-code" stays MEDIUM color), [^Goodhart] (qualitative-only, blue-labeled), [^AlertFatigue] (LOW-on-number standing; r3 swap offer unbanked), [^HeadlessProbe] (MEDIUM, disposition-of-record), [^PortPlan] (snapshot-grade per adjudicated R1-7; AgentOrange HEAD 6df52af clean per session git status) | r4 | 2026-07-17

## round 4 — lens 3 (slice: §4/§5 + referenced footnotes; R3-14/R3-15/R3-9/R3-17 changed sections + >2-round re-fetch dues)

R3-14 carve-out quote + command set + not-configurable + ask/deny remedy (§4.2/§4.3/[^PermissionsDoc]) | [^PermissionsDoc] live fetch | high (verbatim; concurs lens 2) | r4 | 2026-07-17
"a deny rule can't carry allowlist exceptions" + "If a tool is denied at any level…" (§4.2) | [^PermissionsDoc] live | high (both verbatim) | r4 | 2026-07-17
"A blocking hook … stops the tool call before permission rules are evaluated" (§4.3 layers 2/4) | [^PermissionsDoc] live, hooks-interaction section | high (verbatim) | r4 | 2026-07-17
permissions.disableBypassPermissionsMode "disable" inside permissions (§4.2 JSON); same doc sentence documents permissions.disableAutoMode → OQ17 resolvable now → L3-F3 | [^PermissionsDoc] live + [^CliReference] `auto` mode listed | high | r4 | 2026-07-17
Carve-out deny-enumeration completeness: doc list is "These include…" (non-exhaustive; classifier not configurable) vs §4.2 "full enumeration in the shipped file" → L3-F1 MEDIUM (bare-Bash deny is the doc-verified structural close: "A bare tool name like Bash removes the tool from Claude's context entirely") | [^PermissionsDoc] live | high (doc reading) | r4 | 2026-07-17
§4.2/§7 "no prompt ⇒ carve-out classifier treats git log --output as read-only" — blue seat's own broad-Bash session confounds the probe; not isolable from any seat (incl. this one) → L3-F2 LOW | internal; isolation requires clean dontAsk profile (OQ18 build test) | n/a | r4 | 2026-07-17
Read/Edit denies bind recognized Bash file commands (cat/head/tail/sed) — strengthens row-13 named-target closure → L3-F4 (converges with lens 2's sharper L2 reading) | [^PermissionsDoc] live | high | r4 | 2026-07-17
§5.2/[^Pricing] zero-drift re-fetch; R3-17 tokenizer set VERBATIM (Opus 4.7+ Opus models, Fable 5, Mythos 5, Mythos Preview, Sonnet 5; ~30%; Sonnet 4.6-and-earlier legacy); all price rows + Batch flat 50% + cache 0.1× unchanged | [^Pricing] live | high | r4 | 2026-07-17
§4.1 #6631 closed; #4467 "fix merged" (ant-kurt); reporter re-confirmed bypass v1.0.93 Aug 2025 | [^DenyRWIssue] live | high (r4 drift re-check, zero drift) | r4 | 2026-07-17
§4.1 #25621 CLOSED duplicate, canonical unnamed on page | [^DenyBashIssue] live | high (r4 drift re-check) | r4 | 2026-07-17
§5.1 --max-budget-usd verbatim; --max-turns exits-with-error; --fallback-model NO print-only + persistent fallbackModel setting (R1-10 stands); no exit-code table (R1-6 stands) | [^CliReference] live | high (r4 re-fetch, zero drift) | r4 | 2026-07-17
§5.1 Usage/Cost Admin API: both endpoints; Admin key; "unavailable for individual accounts"; ~5-min freshness; 1/min polling | [^UsageAPI] live | high (r4 re-fetch) | r4 | 2026-07-17
§5.1 Rate Limits API: org + workspace endpoints; read-only ("Can I update rate limits with this API? No — Console Limits tab"); Admin key; no spend-limit API documented | [^RateLimitsAPI] live | high (r4 re-fetch; §5.1 requalified shape fully corroborated) | r4 | 2026-07-17
§4.1 STOP figures: 0.42% (Wilson 0.31–0.57) / 0.46% (0.35–0.61) / 10,000 / not significant (two-proportion z-test α=0.05) / syntactic use_sandbox=False|exec( | [^STOP] ar5iv live | high (r4 re-fetch, exact) | r4 | 2026-07-17
§4.1/§2.4 DGMSakana five quotes (fake log; markers-removal incl. parenthetical; lineage; sandboxed-under-human-supervision; improve-with-more-compute) | [^DGMSakana] live | high (all verbatim; concurs lens 1) | r4 | 2026-07-17
§4.1 AIScientist: sandboxing quote verbatim live; timeout-modification fragment verbatim live; run-itself fragment behavior-confirmed live (verbatim captured r1) | [^AIScientist] live | high | r4 | 2026-07-17
§4.1 OWASP LLM06: 3 root causes + least-privilege + HITL-high-impact + downstream-authorization + logging + rate-limiting quoted; draft-vs-execute gloss leaf-confirmed (manual-review/"hit send" mitigation) | [^OWASP] genai.owasp.org live | high (draft-vs-execute upgraded from unpinned) | r4 | 2026-07-17
§5.2 arithmetic recomputed: $3–15/mo; ≥3× headroom; $60–150/mo; $4.5k–12k/mo; 0.1×≈90% | internal | high | r4 | 2026-07-17
Carried on pin immutability (no drift possible): [^SemanticConsent], [^PushGuard], [^CostRecord] $414.97/$149.95, [^EfficiencyPlan], [^HooksJson], [^IdeasCorpus] hooks-fire-on-subagent, [^HeadlessProbe] (medium, disposition-of-record: re-run at build) | @7bc501e | high (carried) | r4 | 2026-07-17

## round 4 — lens 4 (slice: §6/§7/§8 + [^IdeasCorpus]/[^HooksJson])

§6 row 8/§3.3 #76239 OPEN (title verbatim) | live gh r4 | high (drift re-check, still OPEN) | r4 | 2026-07-17
§6 row 8/§3.3 #68375 OPEN (title verbatim: regression 2.1.177, --strict-mcp-config workaround) | live gh r4 | high (drift re-check, still OPEN) | r4 | 2026-07-17
§7/§4.2/[^PermissionsDoc] carve-out passage verbatim (full 14-command set + read-only git; "in every mode"; "not configurable"; ask/deny remedy) + "a deny rule can't carry allowlist exceptions" + unquoted-glob rule (doc names find/sort/sed/git) + exec-wrapper/find -exec always-prompt | live WebFetch permissions doc (independent of lens 2/3 fetches — three-way concurrence) | high (verbatim, zero drift) | r4 | 2026-07-17
§4.2 carve-out deny enumeration recount: all 14 doc-listed commands deny-covered (ls bare+starred, pwd bare) | internal recount vs live doc | high — NOTE lens 3 L3-F1: doc's "These include" is non-exhaustive, so enumeration-completeness is against the doc's LIST, not the classifier's SET | r4 | 2026-07-17
§7 "red offered risk-accept" (R3-17) + "R3-9 recommend-not-block" + "all 17 addressed" | red/ledger.md R3-17 grading + R3-9 heading + LEAD docket recount | high (faithful to red's record) | r4 | 2026-07-17
§7 "BROADER than the round-3 gap summary (adds 9)" | debate.md ### RED r3 (abbreviated cat/grep/git) vs red/ledger.md R3-14 problem block (full set enumerated) | high-as-scoped — true of the summary channel; red's ledger was complete; lossy-summary friction already logged | r4 | 2026-07-17
§7/§4.2/§6 row 4 "showing the carve-out classifier itself passes --output" (attribution sub-claim; gadget itself r3-verified twice, stands) | isolation probe attempted ×2 (nested claude -p --permission-mode dontAsk, fresh untrusted temp repo, ~/.claude/settings.json verified: no Bash allows, defaultMode "auto") — both DENIED by this seat's auto-mode classifier | LOW-MEDIUM — layer-masked: auto-mode seats' approving layer is the AUTO classifier → L4-F1 (converges lens 3 L3-F2) | r4 | 2026-07-17
Lens-2 new leaf CONCURRED at my own fetch: "Read and Edit deny rules apply to both the native tools and to file commands Claude Code recognizes in Bash, such as cat, head, tail, and sed" — affects §6 row 13 (my slice): "would have been AUTO-APPROVED under the round-2 profile" over-claims for the Read-denied transcript target | persisted WebFetch output, direct quote | high (doc clause; row-13 sentence LOW as-written — defer merge to L2-F2 lineage) | r4 | 2026-07-17
§6 row 8 #32191 / §7 #66395/#23707/#837/#14246/#22055 closed statuses | carried r3 live checks (closed = low-volatility); #6631/#25621 re-checked live by lens 3 this round | high (carried) | r4 | 2026-07-17
§6 row 9 [^HooksJson] + row 4 [^IdeasCorpus] | @7bc501e pin-immutable | high (carried r1–r3) | r4 | 2026-07-17
§8 OQ8 STOP figures | [^STOP] ar5iv (lens 3 re-fetched exact this round) | high | r4 | 2026-07-17

## round 5 — lens 4 (slice: §6/§7/§8 + referenced footnotes)

§6 row 8/§3.3/§7 #76239 OPEN (title verbatim, regression since 2.1.144) | [^McpHeadlessBugs] live WebFetch r5 | high (drift re-check, still OPEN) | r5 | 2026-07-17
§6 row 8/§3.3/§7 #68375 OPEN (title verbatim, regression 2.1.177, --strict-mcp-config workaround) | [^McpHeadlessBugs] live WebFetch r5 | high (drift re-check, still OPEN) — NOTE now carries GitHub `stale` label (bot auto-close = future drift risk; still OPEN today) | r5 | 2026-07-17
§6 row 4/§8 OQ18(c) "git format-patch -1 -o <path> → exit 0 arbitrary out-of-repo patch (R4-2 leaf-verified)" | live `git format-patch -1 -o /tmp/l4probe_r5 HEAD` → exit 0, patch written out-of-repo | high (blue's absorbed leaf faithful, re-run) | r5 | 2026-07-17
§6 row 13/§7 R4-5 deny-reach clause ("Read/Edit denies extend to Bash cat/head/tail/sed") | [^PermissionsDoc] leaf-verified r4 (L2+L4), doc re-fetched zero-drift ×4 seats r4; blue R4-5 fix quotes faithfully | high (carried, ≤2 rounds, doc stable) | r5 | 2026-07-17
Carried HIGH (verified r1–r4, ≤2 rounds, non-volatile/pin-immutable, section text stable): §6 row 7 [^STOP]/[^DGMSakana]/[^AIScientist], §6 row 1 [^SlashHeadlessIssues] #837/#14246 closed, §6 row 8 #32191 closed, §6 row 3 [^Dependabot], §6 row 9 [^HooksJson] @pin, §6 rows 11/14 [^RoutinesDocs], §8 OQ4 ~973MB [^QmdDaemon] @pin, §8 OQ8 STOP figures, §7 Pattern A roster (#22055/#25621/#6631/#66395/#23707 closed) | per-footnote | high (carried) | r5 | 2026-07-17
Archive spot-checks r5: R2-13 (§6 row 5 cell) / R2-14 (§7 pricing-lag bullet) / R3-16 (telemetry SHIPPED — Glob re-confirms board-telemetry.jsonl exists) — all archive records match report+ledger | targeted archive read + Glob | high | r5 | 2026-07-17

## round 5 — lens 2 (slice: §2/§3 + referenced footnotes)

§3.3 /loop session-scoped, 7-day recurring expiry, "No catch-up for missed fires", fires only while open+idle | [^ScheduledTasks] live WebFetch (re-fetch due, last leaf r1) | high (all verbatim, zero drift) | r5 | 2026-07-17
§2.4/§0 disable-model-invocation scheduled fires reach Claude as plain text (v2.1.196 marker on page) | [^ScheduledTasks] live | high (verbatim) | r5 | 2026-07-17
§3.4 rung-2 Desktop per-task permission config / machine-on / local files / 1-min minimum | [^ScheduledTasks] live compare table | high | r5 | 2026-07-17
§3.4 rung-3 full set: no-prompts autonomous; claude/-branch default; fresh clone default branch; ≥1h; connectors default-included; identity attribution; claude.ai-login-only + API accounts unsupported; daily cap "rejected until the window resets"; green-status warning; research-preview volatile | [^RoutinesDocs] live WebFetch (re-fetch due, last leaf r1) | high (all verbatim, zero drift) | r5 | 2026-07-17
§3.4 systemd Persistent= stored-on-disk + trigger-on-activation | [^MissedRun] man7.org systemd.timer live | high (verbatim) | r5 | 2026-07-17
§3.4 anacron executed-in-last-n-days + no continuous-running assumption | [^MissedRun] man7.org anacron(8) live | high (verbatim) | r5 | 2026-07-17
§3.4 Task Scheduler missed-start catch-up | [^MissedRun] learn.microsoft.com StartWhenAvailable property page ("can start the task at any time after its scheduled time has passed"); r1 settings URL now 404 | high on mechanism / MEDIUM exact-UI-string leaf (page moved; attempts: task-scheduler-2-settings 404 → tasksettings-startwhenavailable ok) | r5 | 2026-07-17
§3.4 GHA schedule delayed at high load / dropped / 5-min minimum | [^GhaSchedule] docs.github.com live | high (verbatim) | r5 | 2026-07-17
§3.4 web sessions isolated Anthropic-managed VM + separate proxy holds GitHub token outside sandbox | [^WebSandbox] sandbox-environments live | high (verbatim, zero drift vs r2) | r5 | 2026-07-17
§3.3 #837 CLOSED (live page; reason COMPLETED carried r1 gh — WebFetch renders status not reason; gh classifier-blocked from seats) | [^SlashHeadlessIssues] WebFetch | high | r5 | 2026-07-17
§3.3 #14246 CLOSED as Duplicate | [^SlashHeadlessIssues] WebFetch | high | r5 | 2026-07-17
§2.1 IdeaStudy novel p<0.05 / feasibility slightly weaker / self-eval + diversity failures | [^IdeaStudy] arXiv abs 2409.04109 live | high (verbatim; "human re-ranking improved" carried on r1 /html leaf — not in abstract) | r5 | 2026-07-17
§3.3 qmd ladder 973MB / 36.3s vs 2.9s / :8181/mcp + /health + PID | [^QmdDaemon] git show 7bc501e:ideas/backlog.md:34 pin spot-check | high (verbatim) | r5 | 2026-07-17
§2.4/§3.4 PortPlan "human approves each step" (l.287) / "scheduling always human-opt-in" (l.357) / Phase-4 "touches only research/+ideas/" (l.331) | [^PortPlan] AgentOrange tree, HEAD 6df52af clean | high-content / snapshot-grade (R1-7 standing) | r5 | 2026-07-17
§3.4 R4-10 cap/HALT arithmetic (deaths n4/n8; cap-skip ~n11; HALT ~$10 into month 2; ~$55-60 total) | internal recompute | high (correct) | r5 | 2026-07-17
§2.3/§3.4 R4-7 repair (degrade-note readers now specified both sites) — closes r4 L2-F1 | internal cross-read | high | r5 | 2026-07-17
§3.4 R4-14 four SessionStart print states complete (incl. R3-9 enabled-healthy timestamp) | internal | high | r5 | 2026-07-17
§3.4 rung-0 ladder cell "command markdown is the wrapper's phase-1 prompt payload" vs R4-1 trampoline (payload in skill file) | internal contradiction, same section | LOW as-written → L2-F1 (r5) | r5 | 2026-07-17
§2.3 status enum omits regression flag domain (3 of 4 R4-9 re-surface flags are enum values; graduated-state rule unpinned) | internal | LOW → L2-F2 (r5) | r5 | 2026-07-17
Carried without re-fetch (within carry rules): [^HeadlessDocs] (r3 live), [^CliReference] (r3 help + r4 live), [^PermissionsDoc] carve-out (r4 x3 live), [^McpHeadlessBugs] #76239/#68375 (r4 live gh), [^WindowsHang] (r3 live), pin-immutables [^ResearchCommand]/[^Backlog]/[^SmokeRecord]/[^EfficiencyPlan]/[^QmdFallback]/[^IdeasCorpus], [^Reflexion]/[^DGMSakana] (r4 live), [^DGM] html (r3 live), [^HeadlessProbe] (medium, re-run at build) | r5 | 2026-07-17

## round 5 — lens 1 (slice: preamble/§0/§1 + referenced footnotes)

§1.3 red-memory mirror "1,557 lines" | run-dir inputs/red-gap-patterns.md wc = 1557 (NOT committed at 7bc501e — run-local input, frozen) | high (>2-round re-verify; pinned input, no drift) | r5 | 2026-07-17
§1.3 backlog "25 statused checkbox / 39 lines" | git show 7bc501e:ideas/backlog.md = 25 checkbox, 39 lines | high | r5 | 2026-07-17
§1.3 telemetry "SHIPPED FEOV 0.7.0 — present in trajectories/" | Glob run-dir board-telemetry.jsonl EXISTS | high (R3-16 carried) | r5 | 2026-07-17
§1.4 [^PortPlan] "human approves each step"(287)/Phase-4 "touches only research/+ideas/"(331)/decision 6 daily default scheduling human-opt-in(356-357) | AgentOrange docs/claude-port-plan.md @6df52af, working tree clean | high-content/snapshot-grade (pin defect R1-7 adjudicated); >2-round re-verify zero drift | r5 | 2026-07-17
§1.4 R4-12 est_complexity "NAMED source = human-recorded complexity note in matching backlog entry" | git show 7bc501e:ideas/backlog.md grep complex/est-/effort/difficulty → no structured complexity field on any of 25 items | LOW-as-implied: named source vacuous vs pinned corpus, factor inert default-1 → L1-F1 | r5 | 2026-07-17
Carried HIGH without re-fetch (verified r4, 1 round elapsed, §1.1/§1.2 UNCHANGED in round-4 revision, immutable arXiv/same-day): [^SelfCorrect], [^Reflexion], [^Voyager], [^DGM]+[^DGMSakana], [^SICA], [^STOP] (page-of-code MEDIUM color), [^Dependabot], [^DependabotFatigue] (75M via /html), [^Goodhart] (qualitative), [^AlertFatigue] (LOW-on-number, honestly labeled, replacement banked r3), [^FrictionRun3], [^FrictionRun4], [^Backlog], [^EffReport], [^EfficiencyPlan], [^ResearchCommand], [^IdeasCorpus], [^HeadlessProbe] (P2 $0.058 MEDIUM ephemeral, disposition-of-record) | r5 | 2026-07-17

## round 5 — lens 3 (slice: §4/§5 + referenced footnotes; R4-2/R4-3/R4-5/R4-11 changed §4 sites re-fetched live)

§4.2/§4.3 "a bare tool name like `Bash` removes the tool from Claude's context entirely" (R4-3 basis) + `Bash(*)`≡`Bash` both remove as deny | [^PermissionsDoc] live WebFetch | high (verbatim, zero drift) | r5 | 2026-07-17
§4.2 read-only carve-out set verbatim + "These **include** …" (non-exhaustive, R4-3 premise sound) + "not configurable; add an `ask` or `deny` rule" | [^PermissionsDoc] live | high | r5 | 2026-07-17
§4.2 sort/sed doc-named (line 198, unquoted-glob-prompt context); file/readlink/strings/less correctly labeled "likely siblings" not doc-named | [^PermissionsDoc] live | high | r5 | 2026-07-17
§4.2/§6-row13 "Read and Edit deny rules apply … and to file commands Claude Code recognizes in Bash, such as cat/head/tail/sed" (R4-5 basis) | [^PermissionsDoc] live | high | r5 | 2026-07-17
§4.2 "If a tool is denied at any level, no other level can allow it" (line 487) + "enforced by Claude Code, not by the model" | [^PermissionsDoc] live | high | r5 | 2026-07-17
§4.2 belt denies `Bash(* --output=*)`/`Bash(* -o *)` doc-legal: "Wildcards can appear at any position" + worked `Bash(* --version)` | [^PermissionsDoc] live | high | r5 | 2026-07-17
§4.2 dontAsk "Auto-denies tools unless pre-approved via /permissions or permissions.allow rules"; carve-out is Bash-only (does not extend tool-level Read) | [^PermissionsDoc] live | high | r5 | 2026-07-17
§4.2 disableBypassPermissionsMode/disableAutoMode "disable" in any settings file (OQ17 answerable; blue keeps open = conservative, not defect) | [^PermissionsDoc] live | high | r5 | 2026-07-17
§4.2 git allow-rule comment "carve-out auto-approves read-only git regardless" CONTRADICTED by same footnote's bare-Bash-removes-tool fact under shipped profile; git allow rules dead under profile → L3-F1 | [^PermissionsDoc] live | LOW as-retained (reference refutes) | r5 | 2026-07-17
§4.3 layer2/4 PreToolUse exit-2 deny + permissionDecision:deny + "blocking hook … stops the tool call before permission rules are evaluated" + "Hook decisions don't bypass permission rules" + agent_id/agent_type + plugin hooks.json | [^HooksDoc] live | high | r5 | 2026-07-17
§5.2 pricing full grid zero-drift (Haiku $1/$5; Sonnet 4.5/4.6 $3/$15; Sonnet 5 intro $2/$10 thru 2026-08-31 → $3/$15; Opus 4.5–4.8 $5/$25; Fable/Mythos5 $10/$50; Batch 50%; cache 0.1×; tokenizer Opus4.7+/Fable5/Mythos5/Preview/Sonnet5 ~+30%, Sonnet4.6-earlier legacy) | [^Pricing] live | high | r5 | 2026-07-17
§4.1 issue statuses live: #22055 CLOSED NOT_PLANNED; #6631 CLOSED COMPLETED (closure≠fixed framing consistent); #25621 CLOSED DUPLICATE | gh issue view | high (drift re-check) | r5 | 2026-07-17
§4.3/§6-row4 R4-2 gadget reproduced live: `git format-patch -1 -o /tmp/…` exit 0 out-of-repo patch; sticked short `-o/tmp/…` exit 0 too | live Bash this box | high | r5 | 2026-07-17
Carried HIGH (≤2 rounds, immutable pins/stable, sections unchanged): [^STOP], [^AIScientist], [^DGMSakana], [^OWASP], [^AIControl], [^CostRecord] $414.97/$149.95, [^EfficiencyPlan], [^EffReport], [^FrictionRun4], [^ResearchCommand], [^SemanticConsent], [^PushGuard], [^HooksJson], [^IdeasCorpus] hooks-fire-on-subagent, [^UsageAPI], [^RateLimitsAPI], [^ConsoleLimits], [^RoutinesDocs] claude/-branch; [^HeadlessProbe] MEDIUM (ephemeral, re-run at build) | r5 | 2026-07-17

## round 5 — lens re-fetch dues + merge probes (leaf verification of the round-4 revision)

§3.3.3 /loop session-scoped; idle-only; 7-day recurring expiry; "No catch-up for missed fires" | [^ScheduledTasks] live | high (zero drift) | r5 | 2026-07-17
§2.4/§0 disable-model-invocation scheduled fires reach Claude as plain text (v2.1.196 marker) | [^ScheduledTasks] live | high | r5 | 2026-07-17
§3.4 rung-2 Desktop per-task permission config / machine-on / local files | [^ScheduledTasks] live compare table | high | r5 | 2026-07-17
§3.4 rung-3 full set: no-prompt autonomy; claude/-branch push default; fresh clone; >=1h; connectors default-included; identity attribution; claude.ai-login-only; daily cap "rejected until the window resets"; green-status warning; research-preview volatility | [^RoutinesDocs] live | high (zero drift, all verbatim) | r5 | 2026-07-17
§3.4 systemd Persistent= catch-up + anacron interval-since-last-run | [^MissedRun] man7.org live | high | r5 | 2026-07-17
§3.4 Task Scheduler missed-start catch-up | [^MissedRun] learn.microsoft.com — r1 settings URL now 404; rerouted to TaskSettings.StartWhenAvailable property page | high on mechanism / MEDIUM on exact MMC UI string (live-source drift; suggest footnote re-point) | r5 | 2026-07-17
§3.4 rung-4 GHA schedule delayed at high load / dropped under saturation / 5-min minimum | [^GhaSchedule] docs.github.com live | high | r5 | 2026-07-17
§3.4 web sessions: isolated Anthropic-managed VM; credentials outside sandbox | [^WebSandbox] live | high | r5 | 2026-07-17
§3.3.1 #837 CLOSED (status live; COMPLETED reason carried from r1 gh leaf — WebFetch renders status only); #14246 Closed as Duplicate | [^SlashHeadlessIssues] live | high | r5 | 2026-07-17
§6 row 8 #76239 OPEN, title verbatim; #68375 OPEN, title verbatim — NOW CARRIES GitHub `stale` LABEL (volatility signal: bot auto-close risk; keep re-checking) | [^McpHeadlessBugs] live | high | r5 | 2026-07-17
§2.1 IdeaStudy: more novel p<0.05; feasibility weaker; self-eval + diversity failures (rerank sub-claim via r1 /html leaf, immutable version) | [^IdeaStudy] arXiv abs live | high | r5 | 2026-07-17
§4 [^PermissionsDoc] full quote set: bare tool name "removes the tool from Claude's context entirely" (+ Bash(*)≡Bash as deny); carve-out "include[s]" 14 + read-only git, not configurable, per-command ask/deny remedy; deny-reach "Read and Edit deny rules apply … to file commands Claude Code recognizes in Bash (cat/head/tail/sed)"; deny-at-any-level supremacy; "enforced by Claude Code, not by the model"; wildcards at any position; dontAsk auto-deny; disableBypassPermissionsMode/disableAutoMode; //-absolute anchoring | live WebFetch | high (all verbatim, zero drift) | r5 | 2026-07-17
§4.2 retained comment "the built-in read-only-git carve-out auto-approves read-only git regardless (R3-14)" | [^PermissionsDoc] | LOW — REFUTED for the shipped profile: the bare Bash deny removes the tool, making the within-Bash carve-out vacuous; comment is an un-reconciled R3-14-era survivor -> R5-1 sub-leg | r5 | 2026-07-17
§4.3 [^HooksDoc]: PreToolUse exit-2 blocks; permissionDecision deny; blocking hook precedes allow rules; hook decisions don't bypass permission rules; agent_id/agent_type; plugin hooks.json source | live | high | r5 | 2026-07-17
§5.2 [^Pricing] full grid: Haiku $1/$5; Sonnet 4.5/4.6 $3/$15; Sonnet 5 intro $2/$10 -> $3/$15 from 2026-09-01; Opus 4.5-4.8 $5/$25; Fable/Mythos 5 $10/$50; Batch flat 50%; cache 0.1x; tokenizer +30% set | live (VOLATILE) | high (zero drift) | r5 | 2026-07-17
§4.1 #22055 CLOSED NOT_PLANNED; #6631 CLOSED (reporter re-confirmed at v1.0.93 framing consistent); #25621 CLOSED DUPLICATE | live gh (lens 3) | high | r5 | 2026-07-17
§6 row 4 / §4.2 git format-patch -1 -o /tmp/<dir> → exit 0, OUT-OF-REPO patch (spaced form) | local leaf re-run (L3, L4) | high | r5 | 2026-07-17
§4.2 belt gap: git format-patch -1 -o/tmp/<dir> (ATTACHED form, no space) → exit 0, out-of-repo patch; matches NO belt deny incl. the new Bash(* -o *) | local leaf run (L6 + merge re-run) | high -> R5-5 | r5 | 2026-07-17
§1.3 mirror 1,557 lines / 30+ patterns; backlog 25 checkboxes / 39 lines; telemetry jsonl EXISTS | run-dir wc + git show 7bc501e + Glob | high | r5 | 2026-07-17
§1.4/§2.4/§3.4 [^PortPlan] quotes ("human approves each step" l.287; Phase-4 verify l.331; decision 6 l.356-357) | AgentOrange working tree at HEAD 6df52af, file clean | high-content / snapshot-grade (R1-7 adjudicated) — zero drift over 3 rounds | r5 | 2026-07-17
§1.4 est_complexity "NAMED source" (backlog human-recorded complexity note) | git show 7bc501e:ideas/backlog.md exhaustive grep | LOW as-implied — NO parseable complexity field on any of 25 items; conditional is well-formed but the source is empty -> R5-10 | r5 | 2026-07-17
§3.4 R4-10 cap/HALT arithmetic (deaths nights 4/8; cap-skip ~night 11; HALT ~$5-10 into month 2; ~$55-60 two-month worst case) | independent recompute x3 (L2, L5, L6) | high | r5 | 2026-07-17
