---
name: pattern-adjacent-defect-count-bleed
description: A repair re-grades defect A's recurrence count using an event that actually belongs to structurally-adjacent defect B — check the cited section literally contains the claimed event, not just that a plausible-sounding event exists somewhere nearby
metadata:
  classes: [citation-figure-misattribution, figure-recount-fails]
  type: feedback
---

When two defect classes are narratively paired throughout a report (e.g. "write-block,
ENAMETOOLONG" as the recurring Tier-B pair), a round's repair to one class's recurrence count
can silently import an occurrence that actually belongs to the *other* class — because both
use the same language ("third occurrence," "this round," "the merge seat") and appear in the
same sentences/paragraphs throughout the corpus.

**Why:** In the FEOV retrospective round 2 audit, blue's R1-13 fix re-graded ENAMETOOLONG's
likelihood Medium→High citing "a third occurrence... at the red-merge seat, per debate.md's
round-1 merge-seat friction." Direct read of that exact cited section showed it contains only
a PDF-fetch-depth note and a process-misfit note — zero mentions of ENAMETOOLONG, heredocs, or
shell-parse failures. The actual documented ENAMETOOLONG occurrences were two (run 2, and this
retrospective's own round-0 synthesis), not three. The write-block defect — described in
adjacent sentences throughout the same report, also using "third occurrence" language, also
citing the round-0 synthesis and a round-1 red-merge hit — is what the round count actually
tracked. The repair correctly copied a *number* from the surrounding prose without checking
that the number's *source citation* named the right defect class.

**How to apply:** When a repair changes a likelihood/recurrence grade citing "per [section X]",
open section X and confirm the specific defect name/keyword appears there — do not accept that
*a* recurrence event of *some* kind is described nearby as sufficient. This is a distinct failure
mode from [[pattern_repair_regression_citation]] (new source contradicts or omits a figure): here
the cited section is real and about the same *general topic* (this report's Tier-B live-smoke
defects) but the specific claimed event is simply absent — a total-absence miscitation, not a
contradicting-number miscitation. Watch especially for shared adjectives ("third occurrence,"
"this round," "the merge seat") reused across two co-located defect rows in a graded table — the
copy-paste boundary between rows is where the bleed happens.

Related: [[pattern_repair_regression_citation]], [[citation_status_and_misattribution_patterns]].
