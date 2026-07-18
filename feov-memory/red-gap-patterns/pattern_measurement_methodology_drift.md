---
name: pattern-measurement-methodology-drift
description: A cited self-reported number (tokens/cost/etc.) survives leaf-node verification against its named source, but a LATER, more rigorous measurement from the same project quietly proves the counting method was wrong — the citation is accurate and the number is still stale
metadata:
  classes: [measurement-methodology-drift, self-attestation]
  type: feedback
---

Gap pattern: **measurement-methodology drift**, a third sibling to [[gap_live_source_drift]]
(external source moves) and [[pattern_self_referential_repo_drift]] (the audited repo's *code
state* moves). Here neither the citation nor the code changes — a **downstream, more rigorous
audit of the same project** later reveals that the informal self-reported figure being cited was
computed with a flawed method, without ever touching the original artifact.

**Why:** Caught in the FEOV-retrospective audit (round 2, lens 2). The report cites "252.9k tokens
(run 1)" and "~3M tokens (run 2)" as headline historical-incident costs, each footnoted to a
friction file's own self-report — both citations check out fine at the leaf node (the number is in
the file, as quoted). But the project's *live* backlog (checked because it kept moving past the
report's pinned SHA — see [[pattern_self_referential_repo_drift]]) had, by the time of this audit,
accumulated a NEW entry: a formal cost-audit tool built specifically to check the informal panel
token counter, which found it undercounts real spend by ~92% (excludes cache traffic — 610K
reported vs. 47.7M in the raw transcripts, for one round). The original 252.9k/3M figures were
never re-computed under the corrected method, so their comparability to each other, and their
precision as decision-supporting evidence, is now suspect — even though every individual citation
in the report is, in isolation, "accurate" (the number really is in the cited file).

**The trap this evades standard leaf-node checking:** grading corroboration confidence per
statement-reference pair passes this claim at HIGH (source says exactly what's quoted) even though
the number itself is now known-unreliable by the project's own later work. Leaf-node fidelity and
methodological soundness are orthogonal checks; passing the first says nothing about the second.

**How to apply:**
- When a report leans on a *magnitude comparison* between self-reported figures (X tokens vs. Y
  tokens vs. Z tokens) to argue urgency/priority, check whether the project has since built (or
  could build) an independent audit of the counting method itself — not just whether the cited
  number still appears in its source file.
- A live backlog/changelog that keeps drifting past the report's pin (per
  [[pattern_self_referential_repo_drift]]) is exactly where this kind of methodology-correcting
  entry shows up — re-check it for content, not just "did the SHA move."
- Grade this MEDIUM impact, not a verdict-blocker: the qualitative recommendation usually survives
  (the free thing is still free, the expensive thing is still expensive) — the gap is precision and
  comparability of the specific numbers used to argue the case, not the direction of the argument.
- Fix is cheap: one footnote flagging that pre-audit figures are self-reported and likely
  undercounted, pointing at the newer audit mechanism as the path to a comparable number next time.
