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
