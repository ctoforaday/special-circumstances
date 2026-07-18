---
name: pattern-self-referential-repo-drift
description: The audited repo's own git state (not an external citation) moves between blue's verification and red's audit — and a merge landing doesn't mean every related call site got fixed
metadata:
  classes: [live-source-drift, incomplete-repair-propagation]
  type: feedback
---

Gap pattern: **self-referential repo drift**, distinct from [[gap_live_source_drift]] (which is
about external web citations going stale). Here the *subject of the report itself* — the repo
being retrospected — changes state between the report's own verification timestamp and red's
audit pass, because both blue and the codebase live in the same fast-moving repo.

**Why:** Caught in the FEOV-retrospective audit (round 1, lens 5). Blue's report headlined "the
fixes exist but have not shipped," verified against `main` @ a specific commit. The PR in question
merged to `main` ~8 minutes *after* blue's verification commit — a genuine race, not sloppiness —
but by the time red audited, the headline was false on the current primary source (confirmed via
`git log`, `git merge-base --is-ancestor`, and `gh pr view --json state,mergedAt,mergeCommit`
against `origin/main`, not just local `git log`).

**The second-order trap:** once you notice the drift, the *naive* fix ("flip unmerged to merged")
overstates resolution. A merge landing does not mean every location the report's disposition table
depends on actually changed. In this case: the merge shipped 3 of 4 null-guard call sites: a
4th (the judge-adjudication call) still threw unguarded on the merged `main`. Verify each
sub-claim the drift affects individually — don't let "it merged" become "therefore it's fixed."

**How to apply:**
- For any report auditing "has X shipped / is Y merged," re-run the exact git/`gh` commands red
  is auditing *now*, not trust the report's cited commit hash as still being HEAD.
- `git log --oneline -1 origin/main` + `gh pr view <n> --json state,mergedAt,mergeCommit` +
  `git merge-base --is-ancestor <sha> origin/main` — confirms merged AND pushed, not a local-only
  merge commit.
- When drift is found, re-verify every downstream claim the stale premise supports individually
  (grep the actual diff / re-read the actual merged file) before declaring it resolved — a partial
  fix disguised as a full one is worse than the original staleness, because it reads as closed.
- Flag both halves as separate gaps: (a) the stale headline itself, (b) whichever specific
  sub-claim survives even after the correction — since a reader who fixes (a) alone will assume
  (b) is also handled.
