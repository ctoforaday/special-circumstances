# red citation-ledger — Review the recommendations of the memory-architecture report against what has shipped since: implemented, superseded, or still open?

## Round 1 Verifications (Lens 1: Leaf-node citation verification)

| Claim | Reference | Confidence | Round | Access Date |
|-------|-----------|------------|-------|-------------|
| Commit 32f13b2 added plans/memory-architecture.md (git log only add) | git log --all -- plans/memory-architecture.md | HIGH | 1 | 2026-07-17 |
| **Commit 32f13b2 date is 2026-07-11, NOT 2026-07-04** (REPORT ERROR) | git show 32f13b2 (Date field) | HIGH | 1 | 2026-07-17 |
| plans/memory-architecture.md absent from HEAD | git show HEAD:plans/memory-architecture.md 2>&1 | HIGH | 1 | 2026-07-17 |
| Zero ingest\|injection\|taint in plugins/ outside research/ | grep -r "ingest\|injection\|taint" plugins/ --include="*.go" --include="*.ts" --include="*.md" | HIGH | 1 | 2026-07-17 |
| Zero /dream\|/remember commands in plugins/ | find plugins -type f \( -name "*.md" -o -name "*.json" \) -exec grep -l "/dream\|/remember" {} \; | HIGH | 1 | 2026-07-17 |
| Zero knowledge/ directories found | find . -maxdepth 2 -type d -name "knowledge" | HIGH | 1 | 2026-07-17 |
| sc-secrets-gate hook exists (hooks.json, requirements.json wired) | grep -r "sc-secrets-gate" plugins/ --include="*.json" --include="*.md" | HIGH | 1 | 2026-07-17 |
| **.claude/ contains NO subdirectories (rules/, projects/ don't exist)** (REPORT ERROR) | ls -la .claude/ and find .claude -maxdepth 1 -type d | HIGH | 1 | 2026-07-17 |
| sleeper-service is Phase-0 scaffold only | ls -la plugins/sleeper-service/ | HIGH | 1 | 2026-07-17 |
| Post-research commits prioritize efficiency-phase, qmd, PDF (2026-07-11 to 2026-07-17) | git log --oneline --since="2026-07-11" | HIGH | 1 | 2026-07-17 |
| plans/ contains only efficiency-phase.md + README.md | ls -la plans/ | HIGH | 1 | 2026-07-17 |
| project-memory skill exists (commit 89a3442, Phase 2) | git show 89a3442 --stat | HIGH | 1 | 2026-07-17 |
| Memory-architecture report §6.3 risk matrix discusses secret/PII gates | grep -n "secret.*outbound\|WebFetch.*WebSearch.*Bash" research/2026-07-12_memory-architecture/report.md | HIGH | 1 | 2026-07-17 |
| Zero allowlist\|denylist schemas in plugins/ | grep -r "allowlist\|denylist" plugins/ --include="*.go" --include="*.ts" --include="*.md" | HIGH | 1 | 2026-07-17 |
| No projections/ directory or .gitignore rules | find . -name "projections" -o -path "*/.gitignore" | HIGH | 1 | 2026-07-17 |
| Zero OKF-schema instantiation (type: rule\|fact) in code | grep -r "type:.*rule\|type:.*fact\|type:.*glossary" plugins/ --include="*.md" | HIGH | 1 | 2026-07-17 |
| efficiency-phase.md scopes to debate-cost levers only (no memory-architecture mention) | grep -n "memory\|dream\|consolidat\|Auto Dream" plans/efficiency-phase.md | HIGH | 1 | 2026-07-17 |
| qmd recall layer is retrieval-only, not consolidation/lifecycle | git show 4a3801c (PR #18) + design scope review | HIGH | 1 | 2026-07-17 |
| Zero mit.1, ingest gates in implementation | grep -r "mit\.1\|ingest.*gate" plugins/ --include="*.go" --include="*.ts" --include="*.md" | HIGH | 1 | 2026-07-17 |
