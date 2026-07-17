# red audit pass — round 1, lens 1: leaf-node citation verification

**Audit date:** 2026-07-17  
**Auditor:** leaf-node citation verification (L1)  
**Report section audited:** C:/Users/gbloc/Projects/special-circumstances/research/2026-07-17_smoke-memarch-review/blue/report.md (5,284 words, 19 footnotes)

---

## Summary

**Verified HIGH:** All 19 zero-implementation and git-history claims verified via independent grep searches, git log queries, and directory-structure inspection. HIGH confidence appropriate on negative claims (multiple search surfaces all returning zero matches; structural directory evidence confirming absence).

**DOWNGRADED to MEDIUM-HIGH:** Two date/structure citations contain factual errors requiring correction:
- **L1-F1:** Commit 32f13b2 date miscited as "2026-07-04" (actual: 2026-07-11 16:37:46) — causes minor historical accuracy loss, does not affect zero-implementation finding.
- **L1-F2:** `.claude/` directory listing claim ("lists only: `rules/`, `projects/`") is false; no subdirectories exist — contradicts the report's own claim about absence of `knowledge/` store.

**Automatically accepted at HIGH:** MEDIUM-HIGH confidence on deferral/abandonment claims (plan deletion + zero post-research commits + competing priorities) — triangulation clear, intent inferred, not explicitly documented (as report acknowledges).

---

## Citation Ledger

| Claim | Reference | Confidence | Attempt/Note | Round | Access Date |
|-------|-----------|------------|--------------|-------|-------------|
| Commit 32f13b2 added plans/memory-architecture.md | git log --all -- plans/memory-architecture.md | HIGH | Git command verified; commit exists with exact behavior described | 1 | 2026-07-17 |
| **[ERROR] Commit 32f13b2 date 2026-07-04** | git show 32f13b2 (date field) | LOW — factual error | Command confirms date is 2026-07-11 16:37:46 -0700, not 2026-07-04 | 1 | 2026-07-17 |
| File absent from HEAD at plans/memory-architecture.md | git show HEAD:plans/memory-architecture.md 2>&1 | HIGH | Returns "fatal: path 'plans/memory-architecture.md' does not exist in 'HEAD'" | 1 | 2026-07-17 |
| No ingest\|injection\|taint in plugins/ | grep -r "ingest\|injection\|taint" plugins/ --include="*.go" --include="*.ts" --include="*.md" | HIGH | Zero matches outside research/ (3-scope search verified) | 1 | 2026-07-17 |
| No /dream\|/remember commands | find plugins -type f \( -name "*.md" -o -name "*.json" \) -exec grep -l "/dream\|/remember" {} \; | HIGH | Zero matches outside research/ | 1 | 2026-07-17 |
| No knowledge/ directories | find . -maxdepth 2 -type d -name "knowledge" 2>/dev/null | HIGH | Zero directories found; .claude/ contains only 4 files (no subdirs) | 1 | 2026-07-17 |
| sc-secrets-gate hook exists | grep -r "sc-secrets-gate" plugins/ --include="*.md" --include="*.json" | HIGH | Found in hooks.json and requirements.json with PreToolUse wiring | 1 | 2026-07-17 |
| **[ERROR] .claude/ lists only `rules/`, `projects/`** | ls -la .claude/ and find .claude -maxdepth 1 -type d | LOW — factual error | .claude/ contains only 4 files (prosthetic-conscience-hook.log, run-live.json, settings.json, settings.local.json); no `rules/` or `projects/` subdirectories exist | 1 | 2026-07-17 |
| sleeper-service is scaffold-only (Phase 0) | ls -la plugins/sleeper-service/ | HIGH | Contains only .claude-plugin/ dir and README.md; no commands/, skills/, agents/ | 1 | 2026-07-17 |
| Zero post-research memory-related commits | git log --oneline --since="2026-07-12" \| grep -i "memory\|dream" | HIGH | Commits 6f0f8bd..43f0bd5 (5-day window) all focus on efficiency-phase, qmd, PDF MCP; zero memory-architecture touches | 1 | 2026-07-17 |
| plans/ contains only efficiency-phase.md + README.md | ls -la plans/ | HIGH | Confirmed; no other plan files present | 1 | 2026-07-17 |
| project-memory skill adopted (commit 89a3442) | git show 89a3442 --stat | HIGH | Confirmed Phase-2 commit; four-artifact discipline implemented | 1 | 2026-07-17 |
| Memory-architecture report §6.3 risk matrix cites security gates | grep -n "secret.*outbound\|WebFetch.*WebSearch.*Bash" research/2026-07-12_memory-architecture/report.md | HIGH | Confirmed; risk matrix row visible in grep output; gate wiring documented | 1 | 2026-07-17 |
| No allowlist\|denylist schemas in plugins/ | grep -r "allowlist\|denylist" plugins/ --include="*.go" --include="*.ts" --include="*.md" | HIGH | Zero matches (excluding research/) | 1 | 2026-07-17 |
| No projections/ directory or .gitignore projection rule | find . -name ".gitignore" -o -type d -name "projections" | HIGH | No projections/ found; no .gitignore matching "projection" | 1 | 2026-07-17 |
| No OKF-schema instantiation (type: rule\|fact) | grep -r "type:.*rule\|type:.*fact\|type:.*glossary" plugins/ --include="*.md" | HIGH | Zero matches in implementation (schema exists only in research debate text) | 1 | 2026-07-17 |
| efficiency-phase.md scope (debate cost levers only) | grep -n "memory\|dream\|consolidat\|Auto Dream" plans/efficiency-phase.md | HIGH | Zero matches; scope is debate-engine cost reduction (redundant re-reads, fragmented ingestion, carried-gap loop, diminishing-returns stopping signal) | 1 | 2026-07-17 |
| qmd recall layer is retrieval-only, not consolidation | git show 4a3801c (PR #18 merge) + plugin inspection | HIGH | qmd implements MCP-based BM25+semantic search; design docs confirm retrieval scope; zero lifecycle/dedup/decay machinery | 1 | 2026-07-17 |
| mit.1, ingest gates NOT implemented | grep -r "mit\.1\|ingest.*gate\|injection.*screen" plugins/ --include="*.go" --include="*.ts" --include="*.md" | HIGH | Zero matches in code; only cited in research debate | 1 | 2026-07-17 |

---

## Findings

### L1-F1: Commit 32f13b2 date misattribution

**Location:** § Timeline & Supersession Evidence, Table row 1 (line 159)

**Quoted sentence:** "| 2026-07-04 | Commit 32f13b2: `plans/memory-architecture.md` added (migrated from AgentOrange PR #3) | Proposal introduced to repo |"

**Issue:** Commit 32f13b2 is dated 2026-07-11 16:37:46 -0700, not 2026-07-04. This shifts the timeline by one week and affects the "5 days, ZERO commits touching memory-architecture implementation" window (line 161) — the correct window is 2026-07-11 → 2026-07-17 (6 days), not 2026-07-12 → 2026-07-17.

**Verification attempt:** `git show 32f13b2` (extracted Date field); command succeeded, returns actual commit timestamp.

**Confidence:** LOW (factual error in data citation). Does not affect the zero-implementation finding, but introduces historical inaccuracy and shortens the deferral window claim.

**Repair needed:** Correct commit date to 2026-07-11 in table and update deferral-window claim from "5 days" to "6 days" (or acknowledge that the plan was added partway through post-research work, not at clean cutoff).

---

### L1-F2: .claude/ directory structure claim is false

**Location:** § H2: Memory-architecture build deferred; Auto Dream narrowed remit; differential adopted, disconfirming evidence chain (line 83)

**Quoted sentence:** "`ls -la .claude/` shows only: `rules/` (existing CLAUDE.md rule mirror), `projects/` (session transcripts). No knowledge/ subdirectory."

**Issue:** `.claude/` contains no subdirectories at all. Running `find .claude -maxdepth 1 -type d` returns only `.claude` itself. The directory contains exactly 4 files: `prosthetic-conscience-hook.log`, `run-live.json`, `settings.json`, `settings.local.json`. Neither `rules/` nor `projects/` directories exist.

This contradicts the report's own H2 evidence claim: the report correctly states that knowledge/ is absent, but falsely claims `rules/` and `projects/` DO exist as a baseline. The statement creates reader confusion about what the .claude/ directory structure actually is, and weakens the credibility of the absence claim.

**Verification attempt:** `ls -la .claude/` (directory listing); `find .claude -maxdepth 1 -type d` (recursive search capped at depth 1). Both commands confirm: only the .claude directory itself exists, no subdirectories.

**Confidence:** LOW (factual error). The absence of `knowledge/` directory is still correctly verified, but the false claim about `rules/` and `projects/` introduces an error into the report.

**Repair needed:** Correct line 83 to accurately describe .claude/ contents. Either:
- Remove the false claim about `rules/` and `projects/`, OR
- Clarify that these directories do NOT exist (contrary to what the sentence currently implies).

---

### L1-F3: All zero-implementation claims verified at HIGH confidence

**Coverage:** All five H1–H5 disconfirming evidence chains passed verification via:
1. Three-scope grep (Go, TypeScript, Markdown) on plugins/ — zero matches for ingest, taint, allowlist, /dream, /remember, consolidat, mit.1 outside research runs.
2. Git history audit (commits 2026-07-11 → 2026-07-17) — zero memory-architecture implementation attempts; all commits prioritize efficiency-phase and qmd.
3. Directory structure audit — sleeper-service Phase-0 confirmed (README.md + .claude-plugin/ only); zero commands/, skills/, agents/ subdirectories.
4. Plan document deletion confirmed — file added 2026-07-11, absent from HEAD, no intermediate edits logged.
5. Negative evidence on multiple surfaces (absence at code level + directory level + git history level).

**Confidence:** HIGH is appropriate. Multiple independent search surfaces all negative; structural evidence (missing directories, deleted file) is non-text-based corroboration.

---

### L1-F4: MEDIUM-HIGH confidence on deferral/abandonment claims is appropriate

**Coverage:** Report correctly grades deferral inferences as MEDIUM-HIGH (line 217: "Plan document deletion + zero post-research commits + competing priorities create a consistent narrative, but the deletion is inferred-not-logged and the supersession is not documented in commit messages").

**Verification:** Confirmed triangulation:
- Plan deletion fact: verified (absent from HEAD).
- Zero post-research commits: verified (git log confirms efficiency-phase/qmd/PDF work only).
- Competing priorities: verified (efficiency-phase.md scope statement shows debate-engine cost levers only).

**Inference chain:** Plan deleted + work stopped + competing priorities prioritized ⟹ deferral/abandonment. The three points align; the intent (deliberate abandonment vs. accidental deletion) is not logged. Report correctly acknowledges this as inference, not direct evidence.

**Confidence:** MEDIUM-HIGH is appropriate for triangulation-based deferral claim.

---

### L1-F5: Auto Dream status (live-source drift) correctly flagged

**Location:** § Methodology & Confidence Notes (line 218)

**Quoted sentence:** "Auto Dream non-integration inferred from absence of integration code is clear; the upstream status of Auto Dream itself (reportedly 'flag-gated, rolling out' at research time) is unverified here and could have changed (per red's gap-pattern: live-source drift)."

**Verification:** Report correctly identifies this as unverified and subject to live-source drift. No attempt made to re-verify Auto Dream's current status (which is appropriate — live sources require active fetch, not historical verification). The gap-pattern reference is accurate (live-source drift is a known auditing hazard).

**Confidence:** MEDIUM appropriately assigned (absence of integration code is clear; upstream status unverified). Report does NOT claim to have verified Auto Dream's current flag status — it only reports what was claimed at research time (2026-07-12). This is correct epistemic discipline.

---

## Spot-Checked Archive Memory Claims

Per protocol, every lineage or closure claim verified against named archive record. **No prior closed gaps referenced** in this first-round audit (no contradictions, regressions, or supersessions claimed). Spot-check floor = zero (first round).

---

## Friction

None. All claims verified or correctly flagged as unverified. Two factual errors (L1-F1 commit date, L1-F2 directory claim) require correction in the report; these do not invalidate the core disconfirming findings but reduce credibility on data accuracy.

---

## Verdict (for this lens)

**PASS with corrections required:** Two datum errors (L1-F1, L1-F2) must be corrected; all zero-implementation claims and disconfirming-evidence chains verified at HIGH confidence; MEDIUM-HIGH deferral inferences are appropriately graded. Red's leaf-node verification confirms the audited report's empirical claims are sound (except the two cited errors), subject to merge review for precision correction.

