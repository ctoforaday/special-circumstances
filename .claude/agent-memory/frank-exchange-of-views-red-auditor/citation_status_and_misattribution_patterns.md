---
name: citation-status-and-misattribution-patterns
description: Recurring blue citation defects — closed GitHub issues cited as "open", and real figures miscited to the wrong source
metadata:
  type: feedback
---

Two leaf-node citation gap patterns seen in the memory-architecture research (round 1). Check for
both on every citation-verification lens.

**Pattern A — "open bug" that is actually Closed (not planned).** Blue cited GitHub issues
(#57507, #56540) as "open bugs" when both are *Closed as not planned*. This is not cosmetic:
- It inverts the fix story (a not-planned issue will NOT be resolved upstream → design must own the
  workaround, not wait).
- It can create an unsatisfiable plan dependency (a *blocking* change was made "contingent on
  issue resolution" — but it will never resolve).
**How to apply:** whenever a claim rests on a GitHub issue, WebFetch the issue and confirm
open/closed + closure reason. "Closed not planned" ≠ "bug doesn't exist" (the report still
corroborates the phenomenon) but the status and any plan-dependency wording must be corrected.

**Pattern B — real figure, wrong source.** A striking quantitative claim ("60% loss / 36.7×
compression / 2,000 facts") was cited to a blog that does not contain it; the number actually
comes from a different paper (arXiv 2603.17781, "Facts as First Class Objects"). The claim is
true but the footnote is unfollowable — exactly the "laundered into fact" failure.
**How to apply:** for headline numbers, fetch the *cited* source and confirm the number appears
*there*, not merely that the number exists somewhere. Grade statement-as-cited LOW even when the
underlying fact is true.

**Pattern C — a Round-N repair introduces a FRESH contradicted figure.** When blue softens a
challenged number ("80–99%" → "up to ~90–95% (MINJA / environment-injection)"), it may attach the
softened number to a *specific* source that does not support it. Round 2: the environment-injection
"~90%" was pinned to arXiv 2604.02623, whose real max ASR is ~32.5% (up to 8× under stress from a
low base). The repair swapped an untraced band for a *contradicted* attribution — worse, not
better. **How to apply:** never treat a Round-N citation repair as trustworthy because it was a
"fix." Re-follow every repaired footnote to its named primary; a repair that pins a number to a
source is a new statement↔reference pair to verify from scratch. Watch especially for repairs that
split a bundled band across named sources — each named source must now individually carry its half.

**Pattern D — disconfirming/"consensus" citation that rests on an unfollowable self-survey.** Blue
cited a dev.to article + "practitioner consensus surveyed <date>" for a quote-shaped disconfirming
claim; the article framed the topic differently and did not carry the claim, so the load fell on
the unfollowable self-survey. The human/agent's own survey is an untrusted, unfollowable source.
**How to apply:** when a footnote bundles a real source with "consensus surveyed" / "practitioner
consensus," follow the real source and check it carries the *specific* quoted claim; if not, the
claim rests only on the self-survey — flag it, especially when it is a disconfirming leg blue uses
to weaken its own grade.

**Pattern E — the self-audit's verification-limit EXCUSE misstates the source's accessibility.**
Blue honestly labels a figure "from search digest, not leaf-verified (paywalled)" — but the cited
source is arXiv-open, one WebFetch from adjudication (efficiency run, round 1: the "~34% NVD-vs-CNA"
figure excused as paywalled was cited to open arXiv 2508.13644, which does not contain the figure
at all — the paper compares scoring *systems*, never NVD vs CNA). An honest-looking hedge can hide
a misattribution: the hedge implies "the figure is in this paper, just unchecked," when the figure
is not in the paper. **How to apply:** on every "not leaf-verified because <reason>" label, test the
reason — if the source is actually open, do the fetch yourself and adjudicate; grade misattribution
(not mere non-verification) when the figure is absent, even for claims blue fenced as
non-load-bearing.

**Related caveat — HTML-arXiv fetch is lossy on numbers in tables.** When a fetch can't find a
cited statistic in an arXiv HTML page, grade "uncorroborated at leaf node," not "false" — the
small-model fetch often can't read tables. Flag for re-verification against the PDF. (Seen with
arXiv 2604.24450 / 61.38% / 71.58%.)

**Related caveat — WebFetch of a GitHub issue thread is lossy on COMMENTS (false absence-claims).**
Sleeper-service run round 1: WebFetch of issue #22055 confidently reported "no PreToolUse/chmod
workaround in thread" while `gh issue view 22055 --comments` showed a full PreToolUse workaround
comment (GitHub collapses/paginates comments; the page fetch never sees them). Status checks via
WebFetch are fine; any claim about what a thread DOES or DOES NOT contain must go through
`gh issue view <n> --comments` + targeted grep before grading. Same round, same issue: the thread
check also caught a half-false footnote ("chmod 444 + PreToolUse in thread" — PreToolUse yes,
chmod 444 absent) — split bundled workaround claims and verify each half.

**Related caveat — the fetch model's SUMMARY can contradict its own ENUMERATION.** Efficiency run
round 3: asked to count Table 2 rows of arXiv 2510.12697, the fetch model listed all 22 rows
correctly, then asserted "18 reported configurations" in its summary line. Had the summary been
trusted, a *correct* blue claim ("22 configurations") would have been falsely flagged. **How to
apply:** never cite the fetch model's aggregate/count/summary statement as the verification
result — recount from its quoted enumeration by hand; if it enumerates and summarizes
inconsistently, the enumeration is the evidence and the discrepancy goes to friction, not to a
finding against blue.
