# red citation-ledger

Verified citations: claim | reference | confidence | round | access-date

Issue statuses: #32810 CLOSED/NOT_PLANNED, #32997 CLOSED/NOT_PLANNED, #52376 CLOSED/DUPLICATE, #10084 CLOSED/NOT_PLANNED | gh issue view --repo anthropics/claude-code | high | 1 | 2026-07-18

Extended thinking display modes (summarized/omitted/streaming) | platform.claude.com/docs/en/build-with-claude/extended-thinking | high | 1 | 2026-07-18

Signature field: encrypted, multi-turn continuity, non-decryptable | platform.claude.com/docs/en/build-with-claude/extended-thinking | high | 1 | 2026-07-18

JSONL format internal and version-unstable | code.claude.com/docs/en/sessions | high | 1 | 2026-07-18

OpenTelemetry: traces/metrics/logs to Honeycomb/Datadog/Grafana/Langfuse | code.claude.com/docs/en/agent-sdk/observability | high | 1 | 2026-07-18

Extended-thinking redaction hardcoded, not config-bypassable | code.claude.com/docs/en/agent-sdk/observability | high | 1 | 2026-07-18

GitHub #32810 comment: tengu_quiet_hollow flag ~2026-03-10, showThinkingSummaries bypass | gh issue view 32810 --comments | high | 1 | 2026-07-18

Adaptive thinking: no API exposure of effort_selected or decision influence | platform.claude.com/docs/en/build-with-claude/adaptive-thinking | high | 1 | 2026-07-18

Span names: claude_code.interaction, .llm_request, .tool, .tool.blocked_on_user, .tool.execution, .hook | code.claude.com/docs/en/agent-sdk/observability | high | 1 | 2026-07-18

Session.id attribute for multi-turn filtering | code.claude.com/docs/en/agent-sdk/observability | high | 1 | 2026-07-18

--- red-merge-r1 leaf verifications, round 1 (2026-07-19) ---

Performativity rate 0.417 (DeepSeek-R1 / MMLU) | goodfire.ai/research/reasoning-theater AS CITED | low — page does not carry the digit | 1 | 2026-07-19
Performativity 0.417 MMLU and 0.012 GPQA-Diamond, DeepSeek-R1 671B + GPT-OSS 120B | arXiv:2603.05488 "Reasoning Theater: Disentangling Model Beliefs from Chain-of-Thought" (Goodfire AI + Harvard) | medium — corroborated via search result summaries, digit not yet read at the paper itself | 1 | 2026-07-19
NIST AI Agent Standards Initiative 2026-02-17, April sessions, Q4 2026 profile | meta-intelligence.tech | low — direct fetch returns a Taiwan consulting site with zero NIST content | 1 | 2026-07-19
"Tool-Result Truncation: The Silent Bug That Makes Agents Lie" | dev.to/gabrielanhaia | low — author reachable, 30 articles indexed, cited title absent | 1 | 2026-07-19
Compliance API 260+ activity types across 33 categories | generalanalysis.com/guides/claude-compliance-api | low — accessible source reports ~30 typed events, contradicting by ~9x | 1 | 2026-07-19
feov-record blue verb list (avenue, closing, confidence, dispute, friction, manifest-row, petition, position, register, render, retire, revision) | `feov-record blue --help`, plugin 0.10.0 | high — command run at merge; `close` is NOT a blue verb | 1 | 2026-07-19
Pinned inputs probe-thinking-persistence.md + mining-substrate-architecture.md exist at cacb736 | `git ls-tree -r cacb736` + `git show` | high — both retrieved, 41 and 40 lines | 1 | 2026-07-19
Repo HEAD 4baf282 at audit; cacb736 is a valid commit object | `git rev-parse --short HEAD`, `git cat-file -t` | high | 1 | 2026-07-19

--- red-merge-r2 leaf verifications, round 2 (2026-07-19) ---

arXiv:2603.05488 exists; title "Reasoning Theater: Disentangling Model Beliefs from Chain-of-Thought"; authors Boppana, Ma, Loeffler, Sarfati, Bigelow, Geiger, Lewis, Merullo; submitted 2026-03-05, v4 | WebFetch arxiv.org/abs/2603.05488 | high | 2 | 2026-07-19
Performativity Table 1, PER MODEL: DeepSeek-R1 671B MMLU 0.417 / GPQA-Diamond 0.012; GPT-OSS 120B MMLU 0.435 / GPQA-Diamond 0.227 | WebFetch arxiv.org/html/2603.05488v4 | high — read at the paper's own table; UPGRADES round-1 line "medium — digit not yet read at the paper itself" | 2 | 2026-07-19
"Performativity collapses across task difficulty" as a property of the phenomenon | arXiv:2603.05488 AS CITED at §2 | low — DeepSeek-R1 falls ~35x, GPT-OSS 120B falls ~1.9x; the paper's second arm does not show a collapse | 2 | 2026-07-19
NIST-attributed quotation "structured audit logs to every agent action, logging the full chain: input, reasoning steps, tool calls, data accessed, output, and human approval" | zylos.ai/research/2026-05-01-ai-agent-governance-compliance-2026/ | low — page exists and is on-topic, but does NOT contain the quoted string and does NOT attribute the audit-log requirement to NIST; presents an Agent Decision Record schema as an emerging INDUSTRY standard | 2 | 2026-07-19
Compliance API "roughly 30 activity types" | support.claude.com/en/articles/13015708 | low — page LOADS (lens L1's 404 used a mistyped path) but enumerates no activity types; it is a navigation page pointing at platform docs | 2 | 2026-07-19
Compliance API "roughly 30 activity types" | platform.claude.com/docs/compliance-api | low — 404, CARRIED from round 1 (red-merge-r1 fetch), undisputed by blue | 2 | 2026-07-19
feov-record blue verb list unchanged: avenue, closing, confidence, dispute, friction, manifest-row, petition, position, register, render, retire, revision | `feov-record blue --help`, plugin 0.10.0 | high — command re-run at red-merge-r2; matches report footnote exactly; `close` remains a red-merge verb | 2 | 2026-07-19
GitHub issue statuses (#32810, #32997, #52376, #10084) — no drift | lens L1 re-run of `gh issue view --repo anthropics/claude-code` | high — RELAYED from reading seat red-lens-r2-L1; not re-run at the merge | 2 | 2026-07-19
Extended-thinking display modes and signature field — no drift | lens L1 WebFetch of platform.claude.com extended-thinking | high — RELAYED from reading seat red-lens-r2-L1; not re-fetched at the merge | 2 | 2026-07-19
Local transcript store "zero non-empty thinking" | ~/.claude/projects sweep | UNRE-VERIFIED this round — lens L1's spot-check was blocked by the permission system; store observed growing 294->306 files / 5,754->6,057 blocks, consistent with the report's own disclosed moving-target caveat | 2 | 2026-07-19
