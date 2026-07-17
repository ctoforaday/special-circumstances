# Red Ledger — Round 1

**Status:** 11 open gaps (3 pending precision repairs, 8 requiring blue response/verification, 3 security blockers). Verdict: **FAIL** — closing security blockers or risk-acceptance required before pass.

---

## OPEN GAPS

### R1-1: Commit 32f13b2 date misattribution
**Location:** Timeline & Supersession Evidence, table row 1 (line 159)  
**Quoted sentence:** "| 2026-07-04 | Commit 32f13b2: `plans/memory-architecture.md` added (migrated from AgentOrange PR #3) | Proposal introduced to repo |"  
**Problem:** Commit 32f13b2 is dated 2026-07-11 16:37:46, not 2026-07-04. Verified via `git show 32f13b2`. Shifts deferral window from "5 days" to "6 days."  
**Required fix:** Correct commit date in table and update deferral-window claim.  
**Severity:** low-medium  
**Likelihood:** certain  
**Impact:** low-medium  
**Complexity:** trivial  
**Found by:** ["L1"]

---

### R1-2: .claude/ directory structure claim false
**Location:** H2: Memory-architecture build deferred, line 83  
**Quoted sentence:** "`ls -la .claude/` shows only: `rules/` (existing CLAUDE.md rule mirror), `projects/` (session transcripts). No knowledge/ subdirectory."  
**Problem:** .claude/ contains no subdirectories. Verified via `ls -la .claude/` and `find .claude -maxdepth 1 -type d`. Report claims `rules/` and `projects/` subdirs exist; they do not. Knowledge/ absence is correct, but baseline claim is false.  
**Required fix:** Correct line 83 to accurately describe .claude/ contents (4 files, no subdirectories).  
**Severity:** low-medium  
**Likelihood:** certain  
**Impact:** low  
**Complexity:** trivial  
**Found by:** ["L1"]

---

### R1-3: Mislabeled "Superseded" disposition
**Location:** Disposition Summary, "Superseded ✗" section (line 178)  
**Quoted sentence:** "Memory-architecture's blocking security gates (mit.1–2, ingest screening, taint-boundary allowlist inversion) → **not implemented; research-stage fixes R4-1/R4-2 never committed**."  
**Problem:** Supersession requires an alternative mechanism built in place of original. Report lists security gates as "Superseded" without naming a replacement. Gates are unimplemented (deferred), not replaced (superseded). Auto Dream coordination similarly unimplemented, not replaced. Only OKF/project-memory is true supersession. Conflates "unimplemented" with "architecturally swapped."  
**Required fix:** Reclassify security gates and Auto Dream under separate "Deferred / Abandoned" section; retain project-memory under "Superseded."  
**Severity:** medium  
**Likelihood:** certain  
**Impact:** medium  
**Complexity:** low  
**Found by:** ["L2"]

---

### R1-4: Undefined reference to "two ingest-edge gates"
**Location:** H5: Build proceeded under re-scoped margin, item 1 (line 129)  
**Quoted sentence:** "The research report identifies 'blocking core = two ingest gates + mit.1 trust tiers'" (citing H5 §1).  
**Problem:** Report defines "the two ingest gates" by footnote reference to memory-architecture report §6.3, risk matrix row 50, but neither gate identities nor definitions are imported into audit surface. Parenthetical describes *functions* ("never auto-promotes", "screening at capture") but does not establish these as distinct named gates or verify count. Reader cannot verify whether "two ingest-edge gates" is accurate or if blue has double-counted or misread the source.  
**Required fix:** Either import gate definitions/rationales from source research, or flag as external-document dependency requiring source verification.  
**Severity:** medium  
**Likelihood:** medium-high  
**Impact:** medium  
**Complexity:** low  
**Found by:** ["L2"]

---

### R1-5: Unverified counterfactual premise on Auto Dream
**Location:** H4: Recommendations superseded by native Auto Dream, item 3 (line 116)  
**Quoted sentence:** "No FEOV, PC, or sleeper-service agent was modified to check Auto Dream's native flag or coordinate with its consolidation pass."  
**Problem:** Finding's soundness depends on Auto Dream being currently available/feature-complete/compatible (counterfactual premise: "should have integrated"). Report acknowledges live-source drift risk (Auto Dream status "reportedly flag-gated, rolling out" at research time, unverified now) but does not verify current Auto Dream status or whether integration was feasible. Conclusion "should have integrated" rests on unverified counterfactual. Project may have correctly identified Auto Dream as unavailable/unstable post-research.  
**Required fix:** Add follow-up verification: "Auto Dream current status (flag availability, maturity, API compatibility) not verified in this audit. Recommend: query Auto Dream vendor/community status as of HEAD date before concluding 'should have integrated.'"  
**Severity:** medium  
**Likelihood:** medium-high  
**Impact:** medium  
**Complexity:** low-medium  
**Found by:** ["L2"]

---

### R1-6: Missing design-decision context on qmd timing/intent
**Location:** H3: Typed-concept differential shipped minimally, item 4 (line 94)  
**Quoted sentence:** "The project adopted a scaled-down memory approach rather than implementing the full architecture."  
**Problem:** Report infers qmd adoption was "independent" (not a deliberate pivot) but does not verify via: (1) commit date sequence (memory-architecture deferral → qmd PR), (2) PR discussions (refs to memory-architecture deferral), (3) design docs (plans/efficiency-phase.md mention of qmd as recovery strategy). Without verification, "independently-pursued" is inference, not confirmed. If qmd *was* deliberate pivot, disposition shifts from "abandoned" to "phased defer" (material difference).  
**Required fix:** Add targeted verification: "Timeline check: compare commit dates of memory-architecture deferral (plan deletion) against qmd PR #18 and project-memory adoption to establish sequence. Query plans/efficiency-phase.md for explicit mention of memory-architecture deferral or qmd as deferral strategy."  
**Severity:** medium  
**Likelihood:** medium  
**Impact:** medium-high  
**Complexity:** low-medium  
**Found by:** ["L2"]

---

### R1-7: Plan branch narrative (unmerged vs. deleted)
**Location:** Evidence Chains, H2 Memory-architecture build deferred (line 56)  
**Quoted sentence:** "Commit 32f13b2 (2026-07-04) added `plans/memory-architecture.md` (migrated from AgentOrange PR #3)."  
**Problem:** Commit 32f13b2 lives on unmerged feature branch `plans/memory-architecture`, never merged to main. Verified: `git merge-base --is-ancestor 32f13b2 6f0f8bd` returns FALSE; `git branch --contains 32f13b2` shows `plans/memory-architecture` only. Report's narrative "proposal migrated, then deleted" is misleading; accurate narrative: "proposed on feature branch, branch unmerged and dormant." Deferral signal (branch exists, could be revived), not abandonment signal (deleted + removed from history).  
**Required fix:** Reword: "Proposal added on unmerged feature branch 2026-07-04; never integrated into main; zero subsequent commits on the branch since Phase-0 bootstrap."  
**Severity:** low-medium  
**Likelihood:** certain  
**Impact:** low-medium  
**Complexity:** low  
**Found by:** ["L3"]

---

### R1-8: Blocking security prerequisites abandoned without risk-acceptance
**Location:** Evidence Chains, H1 Blocking security recommendations (line 25)  
**Quoted sentence:** "The research report (§11) explicitly states 'the two ingest gates + mit.1, the agent-memory correctness fix, the commit/push secret consumer, and the R4-1/R4-2 structural fixes must be closed *and independently verified* before implementation proceeds past the differentiating sliver.'"  
**Problem:** Research marked these as blocking prerequisites (risk matrix row 50: HIGH impact, MED likelihood, LOW-MED complexity). Post-research, zero commits, no implementation branch, no documented risk-acceptance. If simplified memory/consolidation system ships, it may omit these gates as "out of scope." Security gates are load-bearing; omitting without risk-acceptance = active vulnerability. No decision record; no architectural review of trade-off.  
**Required fix:** Flag as OPEN BLOCKER. Explicit risk-acceptance (with rationale and approved-by) OR committed implementation plan with owner/deadline required before any memory-consolidation system ships.  
**Severity:** high  
**Likelihood:** medium-high  
**Impact:** high  
**Complexity:** low-medium  
**Found by:** ["L3"]

---

### R1-9: Auto Dream two-writer collision unresolved
**Location:** Methodology & Confidence Notes (line 218)  
**Quoted sentence:** "MEDIUM (Auto Dream non-integration): ... the upstream status of Auto Dream itself (reportedly 'flag-gated, rolling out' at research time) is unverified here and could have changed."  
**Problem:** Research risk matrix row 57: "Agent-memory `memory:` row wrong / bidirectional write collision" — HIGH likelihood if Auto Dream does not consolidate. Research flagged this as needing resolution (§6.3: "Native Auto Dream two-writer collision on `MEMORY.md`"). Post-research: Auto Dream integration never attempted, current status not verified. If Auto Dream ships and is used with memory-architecture (or replacement memory system), collision risk activates: MEMORY.md could be written simultaneously by agent and Auto Dream, corrupting state or losing data.  
**Required fix:** OPEN BLOCKER. Before shipping any memory-consolidation system or before Auto Dream integration: (1) verify Auto Dream's current status, (2) test two-writer collision scenario empirically, (3) document mitigation (locking, merge semantics, version-pinning).  
**Severity:** high  
**Likelihood:** medium-high  
**Impact:** high  
**Complexity:** medium  
**Found by:** ["L3"]

---

### R1-10: Project-memory as implicit substitute (no decision record)
**Location:** Disposition Summary, Implemented ✓ (line 174)  
**Quoted sentence:** "This is **not** the OKF-based global cross-project knowledge store proposed in memory-architecture; it is project-scoped, manually-maintained, and has no lifecycle/decay/promotion machinery."  
**Problem:** Project-memory is real artifact (four-artifact discipline). But: no commit message states "adopting as substitute for memory-architecture"; serves overlapping but distinct functions (project-local ephemeral vs. global consolidated); no documented architectural decision. Risk: future work may treat them as orthogonal, leading to re-scoping or double-implementation of memory-architecture despite project-memory already filling simpler use case. Scope creep risk: gap between what project-memory delivers (manual) and what memory-architecture promised (global + lifecycle) may attract new feature requests.  
**Required fix:** Create decision record (docs/ or backlog item) explicitly stating: "Project-memory (Phase 2) adopted as memory-discipline for this project, replacing memory-architecture proposal (2026-07-12) in scope. Trade-off: simpler, project-scoped, manual lifecycle; no global consolidation. If global consolidation needed in future, memory-architecture branch available for revamp." Assign owner; flag as architectural decision for next planning cycle.  
**Severity:** medium  
**Likelihood:** high  
**Impact:** medium  
**Complexity:** low  
**Found by:** ["L3"]

---

### R1-11: Qmd/memory-architecture conflation risk
**Location:** Disposition Summary, Implemented ✓ (line 173)  
**Quoted sentence:** "Qmd is a **DIFFERENT** recall mechanism — it provides retrieval search over markdown documents, not semantic dedup of memory concepts."  
**Problem:** Blue correctly identifies qmd as retrieval-only. But downstream readers may conflate "we shipped memory improvements in PR #18" (qmd) with "we shipped memory consolidation." Qmd = retrieval (what's in corpus?); memory-architecture = consolidation (lifecycle, dedup, promotion, gates). Risk: future "better memory reuse" request might cite qmd as already addressing need, missing that consolidation is blocker. Budget miscalculation: teams may think "we've invested in memory" (qmd shipped), missing consolidation machinery unbuilt. Integration risk: if Auto Dream ships, qmd returns both fresh and stale candidates indiscriminately (no decay).  
**Required fix:** Label qmd explicitly as **retrieval/search layer** in all documentation/planning. Do NOT mention as substitute for or advancement of memory-architecture. If memory-architecture re-scoped, ensure consolidation designed independently with clear handoff points (qmd fetches; consolidation deduplicates; projections serve clients).  
**Severity:** medium  
**Likelihood:** medium-high  
**Impact:** medium  
**Complexity:** low  
**Found by:** ["L3"]

---

### R1-12: Verification file-type blindspot
**Location:** Methodology & Confidence Notes, Gap-pattern pre-flight check (line 223)  
**Quoted sentence:** "Mitigated — this audit searched three file-type scopes (Go, TypeScript, Markdown) + directory structure + git history, not single-scope grep."  
**Problem:** Three-file-type scope better than single; however, several surfaces not searched: (1) Configuration files (`.mcp.json`, YAML/TOML, `.claude/` rule files), (2) Shell/bash scripts (`/scripts/`, setup, hook scripts), (3) Compiled artifacts/binaries, (4) Plugin manifest nested schema (gate behaviors in plugin.json without CLI commands). If ingest gates ARE implemented elsewhere (e.g., pre-commit hook, `.mcp.json` server config), they'd be invisible to grep on source code. If missed, "zero implementation" claim regresses from HIGH to MEDIUM confidence.  
**Required fix:** Expand search surface: re-run grep with `--include="*.json"` (specifically `.mcp.json` and hook configs) and search `/scripts/` for ingest/gate/taint bash functions. If zero results, "zero implementation" confidence moves back to HIGH. If matches found, re-grade the gap and potentially close R1-8/R1-13 with evidence of partial mitigation.  
**Severity:** medium  
**Likelihood:** low-medium  
**Impact:** medium  
**Complexity:** low  
**Found by:** ["L3"]

---

### R1-13: R4-1/R4-2 orphaned implementation responsibility
**Location:** Evidence Chains, H1 Blocking security recommendations (line 48)  
**Quoted sentence:** "The R4-1 and R4-2 fixes were research-stage proposals never committed post-debate."  
**Problem:** VERIFIED. But transition from research-stage fix to implementation responsibility is MISSING. Research explicitly marks as "UNVERIFIED by red" and "blocking prerequisites"; blue correctly identified they weren't committed. Gap is not just non-implementation—it's that: (1) never formally assigned to implementation team, (2) no issue/task created tracking them, (3) no decision record ("defer R4-1/R4-2 to Phase 2" or "accept risk, skip"), (4) "UNVERIFIED" verdict never addressed with follow-up verification plan. Risk: concrete, scoped fixes (parser change, git-ignore rule, ~1–2 days) languish because no one owns them. Orphaned fixes commonly forgotten; re-discovered as new vulnerabilities; knowledge loss; debt accumulation.  
**Required fix:** OPEN BLOCKER. For each of R4-1 and R4-2, create GitHub issue with: research citation (memory-architecture report §15.1–15.2), unverified status + required verification step (red must verify at leaf node post-implementation), acceptance criteria (R4-1: allowlist parser correctly rejects laundering paths; R4-2: git-ignore projection verified in fresh clone), owner + target date before any memory-consolidation system ships.  
**Severity:** high  
**Likelihood:** high  
**Impact:** high  
**Complexity:** low-medium  
**Found by:** ["L3"]

---

### R1-14: Terminology blindness in zero-implementation claim
**Location:** Methodology & Confidence Notes, Confidence calibration (line 216)  
**Quoted sentence:** "**HIGH (implementation-absent claims):** Multiple independent search surfaces all negative; explicit absence in plugin directory structure is structural evidence, not just text-grep."  
**Problem:** Claim is HIGH confidence, resting on structural evidence (absence of directories, commands, files) + grep evidence (absence of keywords). Potential gap: implementation may exist under different terminology. Example: "ingest" might be called "intake"/"import"; "taint" might be "untrusted"/"origin"; "gate" might be "filter"/"check". If implementation uses different vocabulary, grep would miss it. Aliasing: multiple names (e.g., `/dream` consolidation also called `/sync`/`/commit`) requires exhaustive synonymy search. Risk: if undetected, security-gate risk analysis weakens.  
**Required fix:** Secondary verification: manual code review of main plugin implementations (frank-exchange-of-views, prosthetic-conscience, sleeper-service) for any gate/check/screening logic that *might* provide ingest safety, even if not explicitly labeled. If zero findings, can close R1-14 and raise confidence to VERY-HIGH. If matches found, re-grade R1-8/R1-12/R1-13 and potentially close security blockers with evidence.  
**Severity:** medium  
**Likelihood:** low-medium  
**Impact:** medium  
**Complexity:** medium  
**Found by:** ["L3"]

---

## CLOSURE INDEX

| ID | Class | Summary | Supersedes |
|----|-------|---------|-----------|
| R1-1 | closed_as_reframe | Commit 32f13b2 date correction: 2026-07-04 → 2026-07-11; deferral window 5 days → 6 days | — |
| R1-2 | closed_as_reframe | .claude/ directory structure correction: no `rules/` or `projects/` subdirs; only 4 files | — |
| R1-7 | closed_as_reframe | Plan branch narrative: unmerged feature branch, not deleted; deferral signal, not abandonment | — |

---

## Ledger Telemetry

- **Open gaps:** 11
- **Closed gaps:** 3 (all as precision-repair/reframe)
- **Security blockers (HIGH severity, impact, likelihood):** R1-8, R1-9, R1-13
- **Architectural/scope gaps (MEDIUM severity):** R1-3, R1-4, R1-5, R1-6, R1-10, R1-11, R1-12, R1-14
- **Verdict:** **FAIL** — 3 security blockers require resolution (risk-acceptance or implementation plan) before pass
