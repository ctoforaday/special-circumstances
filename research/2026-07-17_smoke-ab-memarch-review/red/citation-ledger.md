# red citation-ledger — Review the recommendations of the memory-architecture report against what has shipped since: implemented, superseded, or still open?

| Claim | Reference | Confidence | Round | Access-date |
|-------|-----------|-----------|-------|------------|
| H1: Phase 2+ items explicitly deferred | frontier.md §H1 | HIGH | 1 | 2026-07-17 |
| H2: Items 2,5,13 not as memory-arch features | blue/report.md line 17; git grep /dream /ingest OKF | HIGH | 1 | 2026-07-17 |
| H3: Item 6 (SQLite ceiling) superseded by qmd adoption | git show 70e35d1; PR #18 merge 4a3801c; frank-exchange-of-views 0.5.0; prosthetic-conscience 0.7.0 | HIGH | 1 | 2026-07-17 |
| Commit 70e35d1 dated 2026-07-14; feat: qmd recall layer | git log --all \| grep qmd; git show 70e35d1 timestamp | HIGH | 1 | 2026-07-17 |
| qmd MCP server, SKILL.md three-access-modes doctrine, sc-recall-index hook all in 70e35d1 | git show 70e35d1 .mcp.json .diff; frank-exchange-of-views/skills/research-protocol/SKILL.md §Recall; sc-recall-index/main.go | HIGH | 1 | 2026-07-17 |
| H4: All blockers remain open; branch unmerged; no impl shipped since July 12 | git log main..plans/memory-architecture --oneline (32f13b2 only); git grep dream ingest okf main (excl. research/plans) | HIGH | 1 | 2026-07-17 |
| H5: Disconfirming greps zero /dream /ingest knowledge OKF .okf in shipping code | git grep exclusions on research/plans directories | HIGH | 1 | 2026-07-17 |
| R4-1 (taint allowlist inversion) remains unverified design text | memory-architecture/report.md lines 85–87; no allowlist code in main | HIGH | 1 | 2026-07-17 |
| R4-2 (git-ignore projections) remains unverified design text | memory-architecture/report.md lines 85–87; no .gitignore projection pattern in main | HIGH | 1 | 2026-07-17 |
| Memory-architecture report verdict UNVERIFIED after 4 rounds (safety ceiling, not soft-pass) | memory-architecture/report.md line 1; compromise rationale line 97 | HIGH | 1 | 2026-07-17 |
| Item 6 caveat: qmd solves retrieval/dedup, not consolidation rewrite-corruption (append-only fix unbuilt) | memory-architecture/report.md risk matrix lines 53, 60; blue/report.md line 69 | HIGH | 1 | 2026-07-17 |
| Efficiency-phase PRs #19–24 address debate-engine, not memory-architecture | git log --all --oneline --since=2026-07-12; commit messages | HIGH | 1 | 2026-07-17 |
