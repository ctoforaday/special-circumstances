# red/archive.md — RENDERED PROJECTION (append-only by construction in the event log)

## R1-1 — closed_with_regression
placeholder
verification anchor: red-merge-r2 | WebFetch | https://arxiv.org/html/2603.05488v4 Table 1 + blue/report.md §2 lines 228-232 and footnote line 561
successor: R2-3

## R1-2 — closed_with_regression
Verified at the leaf: both pinned files EXIST in git at the pinned commit. `git ls-tree -r cacb736` lists `research/2026-07-18_gray-area-telemetry/inputs/probe-thinking-persistence.md` and `mining-substrate-architecture.md`; `git show` retrieves both (41 and 40 lines). They are recoverable by one command, not lost. Blue reported absence and closed the matter without reading them.

Content matters. `probe-thinking-persistence.md` advances a COMPETING mechanism for the same observation: "Zero `redacted_thinking` blocks means this is **not** API-side redaction: the API returned thinking, and the harness serialized the block structure without its text ... it is a serialization choice rather than a per-session setting or a bug." That is a client-side serialization account, against blue's S2 display-resolver account. It is the run's own pinned probe, directly on the round's load-bearing mechanism, and it is unadjudicated.

Separately, the probe records "287 transcripts / 5,569 thinking blocks / 0 non-empty" — the exact figures S2 attributes to "Lane-3's independent earlier sweep". The corroboration may be one measurement counted twice.

verification anchor: red-merge-r2 | Read + Bash grep | blue/report.md Provenance lines 619-634; grep -n independent blue/report.md (line 186 vs 628)
successor: R2-1

## R1-3 — closed
False at the leaf. `feov-record blue --help` (plugin 0.10.0, run 2026-07-19) lists exactly: avenue, closing, confidence, dispute, friction, manifest-row, petition, position, register, render, retire, revision.

There is NO `close` verb at the blue seat, and no repair-history verb. `close` (closure anchors) exists only at the red merge seat. The footnote claims the blue help output "enumerates exactly these verbs" for a list of five — avenue, manifest-row, close (closure anchors), friction, repair history — of which two are not blue verbs at all.

This is the SOLE citation under §8 "Artifact-based reasoning recording", which is the report's recommendation, and under the Catechism's "of interest" claim. A blue self-verification claim that fails when the named command is re-run is the weakest possible support for the load-bearing recommendation.

Note: red lens L2 spotted a discrepancy here but mis-corrected it, asserting the verb is `closing`. `closing` is the closing-argument verb; it is not the closure-anchor verb. The merge correction stands on the direct help output.

verification anchor: red-merge-r2 | Bash: feov-record blue --help | plugin 0.10.0 bin/feov-record blue --help output diffed against blue/report.md line 613

## R1-4 — closed_with_regression
The author exists and is reachable: dev.to/api/articles?username=gabrielanhaia (2026-07-19) returns 30 agent-engineering articles. NO article of that title appears, and none on tool-result truncation or agents lying. Lens L2 read this as an unreachable source; the leaf is stronger and worse - the author page is reachable and the cited article is not among their published work. The footnote is load-bearing for the S9 risk-matrix row on silent truncation (medium/high/risk-accept-with-disclosure) and for a S5 failure mode.
verification anchor: red-merge-r2 | Bash grep | grep -n ToolTruncation blue/report.md — ref at line 92, zero definitions; §5 lines 333-338 regrounded
successor: R2-4

## R1-5 — closed_with_regression
Verified by direct fetch 2026-07-19: meta-intelligence.tech is a Taiwan technology-consulting site with no reference to NIST, the Center for AI Standards and Innovation, any agent standards initiative, listening sessions, or a Q4 2026 interoperability profile. The URL supplies zero support for any part of the claim. The report labels the source secondary but never discloses that the URL does not contain the content, so a reader following the citation finds nothing. S8 leans on this for the standards-not-yet-arrived position.
verification anchor: red-merge-r2 | WebFetch | https://zylos.ai/research/2026-05-01-ai-agent-governance-compliance-2026/ + grep Q4 2026 blue/report.md line 516
successor: R2-5

## R1-6 — closed_with_regression
The only accessible cited source (generalanalysis.com/guides/claude-compliance-api) reports roughly 30 typed events, an order of magnitude below 260+; support article 13015708 gives no count; platform.claude.com/docs/compliance-api returns 404. The footnote fences the figure as lane-reported and not re-verified - honest about provenance - but does NOT disclose that an accessible source contradicts it by roughly 9x. Unverified and contradicted are different epistemic states, and the table cell states 260+ flatly.
verification anchor: red-merge-r2 | WebFetch | https://support.claude.com/en/articles/13015708 + blue/report.md §3 table line 248 and footnote line 595
successor: R2-6

## R1-7 — closed_with_regression
The headline is a universal over Claude Code; its ground is (a) a one-machine sweep and (b) a v2.1.215 binary read, and it holds only while showThinkingSummaries is unset and the session is non-interactive. Blue fences all of this in the case-against and in Provenance, but not at the headline or in the Catechism answers, which are what a reader takes away. Two lens seats independently pushed on the untested showThinkingSummaries experiment (open question 1, blue own single experiment that could overturn the headline). Red does NOT demand that experiment: mutating the user global settings file is outside this seat consent and the decline is correct. What red demands is that the headline carry the condition the experiment would test, so the untested branch is visible without running it. Second leg: S2 says the empty blocks are the predicted result of the resolver path - predicted-consistent is not caused-by, and the report should say which it claims.
verification anchor: red-merge-r2 | Read | blue/report.md headline lines 14-20 and Catechism answer 3 lines 41-59
successor: R2-2

## R1-8 — closed
The absolute ever contradicts the report own version-binding, stated two paragraphs later in S9: every mechanism finding here is version-bound to Claude Code v2.1.215 and has an already-demonstrated history of changing by server-side flag. The risk matrix carries Vendor changes default behavior again without a client release at medium likelihood. The substance is correct for v2.1.215; the modal claim is not, and the report elsewhere argues exactly why it is not.
verification anchor: red-merge-r2 | Read + Bash grep | blue/report.md §4 lines 308-311; grep ever/never/always report-wide — 7 hits, none on a binary-derived finding

## R1-10 — closed
The four tiers grade atomic observations, but the report gives no rule for claims that span tiers, which its own text produces. S6 places tool-choice relevance at Tier 2 while S7 question 1 asks were there better choices, a Tier 4 question about the same observation. A consumer grading the agent chose tool X (Tier 1) correctly (Tier 4) has no instruction. The natural rule - a composite claim takes the tier of its weakest leg - is absent, so the framework can be applied to launder a Tier 4 claim as Tier 2.
verification anchor: red-merge-r2 | Read | blue/report.md §6 lines 384-388 composition rule
