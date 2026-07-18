---
name: pattern-gitignored-not-absent
description: availability/durability claims vs git reality, BOTH directions — "gitignored/absent" refuted by untracked working-tree artifacts; "committed" refuted by untracked status (present ≠ committed)
metadata:
  classes: [artifact-preservation, verification-scope-blindspot]
  type: feedback
---

A report claims data is UNAVAILABLE ("gitignored and absent at the pin", "not retained") and
defers a measurable question to a future run — but the artifact sits untracked in the working
tree, checkable with `ls`. Git-scoped existence checks (git show, git diff, pin equivalence)
see only the object store; gitignored ≠ deleted.

**Why:** efficiency-investigation round 2, L3-F1 — blue's lane 3 declared run-3 agent
transcripts "gitignored and absent at the pin", making the #1-ranked lever's savings
"unmeasurable until run 4"; `ls` found the 7MB tarball (46 per-agent transcripts) in the
working tree, mtime hours before the run launched. The report's own cited artifact (cost.md:
"Measured from per-agent API transcripts") already implied the data existed.

**The mirror direction (round 4, L4-F1):** a repair claims an instrument "is now COMMITTED as
trajectories/decompose-merge.mjs" — the file existed in the working tree and reproduced every
figure on re-run, but `git status` showed the whole directory untracked and
`git log --all -- '*file*'` was empty. Present ≠ committed: the repair answered a
"not-git-tracked audit artifacts" defect by creating another untracked artifact wearing the
word "committed." Substance sound, status word false — grade the residual, don't regress the
figures.

**How to apply:** never verify existence/durability through one lens. For "absent" claims run
`ls`; for "committed/tracked/preserved" claims run `git status` + `git log --all -- '*name*'`
on the exact path. Check mtimes for fairness on likelihood grading. Sibling of
[[gap-pattern-verification-file-type-blindspot]] (existence verified through one lens) and
[[pattern-self-referential-repo-drift]].
