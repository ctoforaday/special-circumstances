---
name: gap-live-source-drift
description: Citation figures/mechanisms drift when live web sources move after the report's access date; re-follow to primary, don't trust the footnote's numbers
metadata:
  classes: [live-source-drift]
  type: feedback
---

Gap pattern: **live-source drift**. A footnote's access date freezes a claim, but the cited
source keeps moving. Star counts, "current algorithm" descriptions, and vendor pipelines change;
the footnote silently goes stale even though it looked verified at drafting.

**Why:** In the memory-architecture audit (round 1, lens 3) three findings were only catchable by
re-fetching live sources: mem0 had switched from retrieve-then-classify (ADD/UPDATE/DELETE/NOOP)
to single-pass ADD-only; claude-mem stars had moved 46k→87.1k; a Letta "git-branch" mechanism was
absent from the cited blog entirely (traced only to an unnamed forum). An audit against archived
snapshots would have passed all three.

**How to apply:**
- Always re-follow citations to the *current* primary source, not just confirm the footnote is
  well-formed. Grade corroboration against what the source says *now*.
- Volatile numbers (stars, benchmarks, "latest version") get LOW confidence unless pinned with an
  access date AND the substantive claim survives without the exact number.
- When a "mechanism to steal / adopt" is recommended, verify the vendor still ships it — vendors
  abandon the exact pipeline being praised (mem0 case). A recommendation to adopt an abandoned
  design is a MEDIUM substantive gap, not pedantry.
- Watch for compound footnotes (`blog + docs + community forum`) where the load-bearing detail
  traces only to the vaguest, unfollowable member. Demand the specific source or downgrade the
  claim.
