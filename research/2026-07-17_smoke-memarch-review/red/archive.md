# Red Archive — Closed Gaps

## R1-1: Commit 32f13b2 date misattribution (closed_as_reframe)

**Found by:** L1 (leaf-node citation verification)  
**Round:** 1  
**Access date:** 2026-07-17

### Finding

Blue's report cites commit 32f13b2 in the Timeline & Supersession Evidence table (line 159) with date 2026-07-04. Leaf-node verification via `git show 32f13b2` confirms the actual date is 2026-07-11 16:37:46 -0700.

### Verification

```bash
git show 32f13b2 | grep "Date:"
# Output: Date:   Wed Jul 11 16:37:46 2026 -0700
```

The discrepancy shifts the reported deferral window from "2026-07-12 → 2026-07-17, 5 days" (line 161) to the actual "2026-07-11 → 2026-07-17, 6 days."

### Impact

**Severity:** low-medium — factual accuracy loss  
**Likelihood:** certain — error verified  
**Impact:** low-medium — timeline credibility affected  
**Complexity:** trivial — data correction only

The error does NOT affect the core finding (zero implementation) or the deferral inference (plan deleted, zero post-research commits). It reduces the precision of historical accuracy.

### Closure

**Class:** closed_as_reframe — requires precision correction in report  
**Recommended fix:** Update table row 1 to date 2026-07-11; update deferral-window sentence (line 161) to "6 days, ZERO commits" instead of "5 days."

---

## R1-2: .claude/ directory structure claim false (closed_as_reframe)

**Found by:** L1 (leaf-node citation verification)  
**Round:** 1  
**Access date:** 2026-07-17

### Finding

Blue's report (H2, line 83) claims `.claude/` lists "only: `rules/` (CLAUDE.md rule mirror), `projects/` (session transcripts)." Leaf-node verification via `ls -la .claude/` and `find .claude -maxdepth 1 -type d` shows `.claude/` contains NO subdirectories, only 4 files: `prosthetic-conscience-hook.log`, `run-live.json`, `settings.json`, `settings.local.json`.

### Verification

```bash
ls -la .claude/
# Output: total 0 (directory only; no subdirs)

find .claude -maxdepth 1 -type d
# Output: ./.claude (only the directory itself)
```

Neither `rules/` nor `projects/` subdirectory exists.

### Impact

**Severity:** low-medium — factual accuracy loss  
**Likelihood:** certain — error verified  
**Impact:** low — does not affect knowledge/ absence claim (which is correct)  
**Complexity:** trivial — data correction only

The error creates reader confusion about .claude/ structure but does NOT invalidate the absence of `knowledge/` directory (which is the point of the evidence chain). The baseline claim is false; the conclusion is sound.

### Closure

**Class:** closed_as_reframe — requires precision correction in report  
**Recommended fix:** Correct line 83 to accurately describe .claude/ contents. Either remove the false claim about `rules/` and `projects/`, or explicitly state these directories do NOT exist. Retain the correct finding that `knowledge/` is absent.

---

## R1-7: Plan branch narrative (unmerged vs. deleted) (closed_as_reframe)

**Found by:** L3 (dark-side & risk verification)  
**Round:** 1  
**Access date:** 2026-07-17

### Finding

Blue's report narrative (H2, line 56) states "Commit 32f13b2 (2026-07-04) added `plans/memory-architecture.md` (migrated from AgentOrange PR #3)." The implicit narrative: the proposal was brought into the repo, then deleted. Leaf-node verification via git branch queries shows commit 32f13b2 lives on an unmerged feature branch `plans/memory-architecture`, not in main branch history.

### Verification

```bash
git merge-base --is-ancestor 32f13b2 6f0f8bd
# Output: FALSE (commit is not an ancestor of main HEAD)

git branch --contains 32f13b2 -a
# Output: plans/memory-architecture (branch, not merged to main)

git log --all --graph --decorate | grep -A5 32f13b2
# Shows commit on separate trunk, diverging from Phase-0 bootstrap, never merged
```

### Impact

**Severity:** low-medium — narrative/metadata framing  
**Likelihood:** certain — verified in git history  
**Impact:** low-medium — affects reversibility interpretation  
**Complexity:** low — metadata correction only

The error is significant for *interpretation* but not for *fact*. The fact remains: the proposal was never implemented post-research. The framing ("deleted" vs. "unmerged branch") has different implications: a deleted proposal is more final; an unmerged branch suggests reversibility. The deferral signal is same; the reversibility interpretation differs.

### Closure

**Class:** closed_as_reframe — requires narrative correction in report  
**Recommended fix:** Reword H2 evidence chain to clarify: "Proposal added on unmerged feature branch 2026-07-04 (`plans/memory-architecture` branch); never integrated into main branch; zero subsequent commits on the branch since Phase-0 bootstrap (2026-07-11). The branch exists and could be revived if priorities change; the main branch history contains zero memory-architecture code."

---

**Archive Summary:** 3 closed gaps, all as precision-repair/reframe. No regressions, no new gaps minted. All closures preserve core findings while correcting factual accuracy and narrative clarity.
