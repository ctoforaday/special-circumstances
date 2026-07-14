# Red audit — round 1, lens 1 (leaf-node citation verification, slice 1 of 3)

Scope: report.md §0 "Headline" and §1 "Doubts: confirmed, refuted, needs instrumentation" (1.1
H1, 1.2 H2, 1.3 H3, 1.4 un-frontiered doubts), plus footnotes those sections cite. Sections
2-5 and remaining footnotes are out of scope for this instance (slices 2-3).

## Finding 1 — HIGH severity: the report's own headline and #1 action item were overtaken by
live git state during the debate window (live-source drift, keystone-level)

**Location:** §0, "the fixes exist, but they have not shipped" — quote: "any run invoked against
`main` today — including a hypothetical run 3 — still carries R1-HARNESS-1's exact defect class
and run 2's null-crash class." Also §3 row 1 (out of my slice but same defect): "Review and merge
PR #14 ... **Do first. This is a shipping decision, not a proposal**."

**Verification performed:** `git log --oneline -1 main` in the live `special-circumstances` repo
→ HEAD is `00018a5174a8`, "Merge pull request #14 from ctoforaday/feat/feov-dogfood-round-1".
`gh pr view 14 --json state,mergedAt` → `{"state":"MERGED","mergedAt":"2026-07-14T05:58:54Z"}`
(= 2026-07-13 22:58:54 -0700). Blue's own [^MainVsBranch] verification was performed against
`9ff0fad` at 22:50:27 -0700 — **8 minutes before the merge**. Confirmed directly on current
`main`: `debate.js` (renamed from `workflow.js`, as the report predicts) now contains the args
guard (`const a = typeof args === 'string' ? JSON.parse(args) : args`), the null-guard
(`if (!redEnv) throw new Error(...)` before every `redEnv.verdict`/`.gaps` dereference),
`catechism_template.md` replacing `heilmeier_template.md`, and `tests/simulator/{harness.mjs,
debate.test.mjs}` — i.e., everything the report says is "absent from `main`" is now present on
`main`.

**Why this matters:** this is not a citation-accuracy failure in the leaf-node sense (blue's
verification was true when performed, and is footnoted with an honest timestamp) — it is the
report's central rhetorical frame ("a shipping question, not a research question," "do first")
being falsified by events that happened *during the same debate*. A reader opening this report
even hours after assembly will find its most urgent claim already resolved, and the report
carries no mechanism — no "re-verify against current HEAD before acting" instruction, no
volatility flag on time-sensitive claims — to warn them. This is a recurring risk class for any
retrospective that both audits live repo/PR state *and* issues a "ship this now" recommendation:
the state is a moving target for the whole duration research is being conducted on it.

**Grading:** likelihood — realized, not merely possible (it already happened, verifiably, within
this same session). Impact — high on the framing/headline and the ranked #1 action item, but
**does not touch the substantive engineering analysis** (the guard's design, the simulator's
design, the trimodal taxonomy) which remains correct regardless of merge status. Complexity to
mitigate — low: one line in the headline noting the claims are pinned to a commit SHA and a
build-time timestamp, and a standing instruction (protocol- or red-level) to re-run `git log -1
main` / `gh pr view` immediately before final assembly and before any reader acts on a "ship this"
recommendation. Not asking blue to chase a moving target mid-debate — asking for the volatility to
be named, the same discipline §3 item 10 (access-date-delta recording) already proposes for
external citations; this is the same failure mode for the repo's *own* state.

**Disposition:** raised, not closed. This is squarely the kind of gap that gets risk-accepted only
with an explicit rationale, not silently — the complexity-to-mitigate is low enough that
risk-acceptance would need real justification.

## Finding 2 — MEDIUM-LOW confidence: [^AgentDiversity]'s two headline statistics do not trace to
the cited paper across four independent access attempts

**Location:** §1.1, "Disconfirming evidence against the naive fix" — quote: "Heterogeneous personas
(distinct lenses, not just different starting topics) measurably reduce pairwise error correlation
(~19%) and recover up to ~95% of a fully independent ensemble's gain — but same-base-model agents
remain more correlated than architecturally distinct ones." Same figures recur in §1.4's inventory
of lane-3-only content and (out of slice) §3 row 6's grading rationale — this is a load-bearing,
multiply-cited quantitative claim, not a passing aside.

**Verification performed:** fetched `arxiv.org/abs/2602.03794` (title confirmed: "Understanding
Agent Scaling in LLM-Based Multi-Agent Systems via Diversity"; abstract-level claims like "2
diverse agents can match or exceed 16 homogeneous agents" confirmed), `arxiv.org/html/2602.03794`,
`arxiv.org/html/2602.03794v1` (explicit full-text search requested for "19%" and "95%" — response:
"no specific percentage near 19%... nor any percentage near 95%... What IS present: Table 1 shows
persona diversity gains ranging from +0.4% to +24.7%... no 95% recovery metric appears"), a
targeted web search for `"19%"` + this paper (no corroborating snippet), and `arxiv.org/pdf/2602.03794`
(tool returned raw PDF stream data, uninterpretable). Four independent paths, zero confirmation of
either figure.

**Confidence: LOW.** The paper exists and its general thesis (heterogeneity beats homogeneous
scaling) is real and topically on-point — this is not evidence of fabrication from whole cloth —
but the two specific numbers dressed in false precision ("~19%," "~95%") are unconfirmed and
should not be repeated at HIGH/asserted confidence until traced to a specific table or page. Given
the tooling could not render the full PDF (see friction), I cannot rule out the numbers living in a
table my access path missed — this is "needs more evidence," not a fabrication verdict.

**Grading:** likelihood the figures are miscited or invented — moderate; impact — medium-high,
since §3 row 6 leans on "external evidence ~19% correlation reduction" to grade the per-lane
diversity recommendation's likelihood as High; complexity to fix — low (locate the exact
table/section and cite it precisely, or soften to the qualitative claim the abstract actually
supports: "heterogeneous configurations measurably outperform homogeneous scaling" without the
specific percentages).

## Finding 3 — LOW-MEDIUM confidence: [^DiminishingReturns] bundles four sources under one figure
("plateau at 2-3 rounds/2-4 agents") that none of them individually appears to state

**Location:** §1.1 — quote: "Diversity/committee gains plateau at roughly 2–4 agents; 'just add
lanes' is the scope-creep reading of the finding [L2]."

**Verification performed:** checked all four bundled sources individually.
`arxiv.org/abs/2603.20640` ("Hear Both Sides") is about message-filtering/noise in debate (Diversity-
Aware Retention), not a plateau-count finding. `arxiv.org/abs/2601.19921` ("Demystifying Multi-Agent
Debate") is about diversity-of-initialization and confidence-calibration interventions, no plateau
count in the abstract. VentureBeat (websearch) discusses a *different* diminishing-returns story
(tool-count >10 causing a 2-6x efficiency penalty, a ~45% single-agent-accuracy inflection point) —
real and on-topic for "diminishing returns" generally, but not the "2-4 agents" figure specifically.
arXiv:2605.00914 (already separately verified for its own distinct figures in Finding for
[^IsolatedCorrection], §2.1 out-of-slice) — its confirmed abstract content doesn't include a "2-4
agent" plateau claim either.

**Confidence: LOW-MEDIUM.** This is a footnote-over-attribution pattern: four citations bundled
under one generic claim, and the specific numeric bound isn't independently pinned to any one of
them at the level I could check (abstracts/HTML summaries; full-text tables were not reachable —
see friction). The general "diminishing returns exist" claim is well corroborated in aggregate; the
specific "2-4 agents" bound is not.

**Grading:** likelihood the specific bound is imprecise/unsourced — moderate; impact — medium (used
to support a "do not just add lanes" argument that the report reaches independently via several
other, better-corroborated lines of evidence, so this citation is not solely load-bearing);
complexity to fix — low (pin the figure to one source with a page/table reference, or state it as
a qualitative synthesis rather than an implied-precise number).

## Finding 4 — MEDIUM confidence, low severity: [^WisdomCrowds] URL does not resolve as given

**Location:** §1.1, footnote [^WisdomCrowds]: "Search synthesis across 'The Wisdom of the LLM
Crowd' (alexanderakm.github.io) and related 2026 LLM-ensemble literature."

**Verification performed:** `WebFetch` on `https://alexanderakm.github.io/blog/wisdom-of-llm-crowd`
→ 404. Websearch located the actual resource at
`alexanderakm.github.io/projects/wisdom-of-llm-crowd.pdf` (a PDF, not the blog page implied) —
topically on point per search snippets ("wisdom of the crowds may not straightforwardly transfer to
LLMs... deliberation between LLMs can promote coherence... similar to cross-cultural homogenization
effects"). The report's specific paraphrase ("under independence... under correlation, diversity
collapses and the collective inherits its members' errors") is a plausible restatement of Condorcet-
style reasoning applied to this material but is self-labeled by blue as "search synthesis," not a
direct quote — appropriately hedged already.

**Grading:** low severity — the footnote already discloses its own synthesis nature. Fix is
citation hygiene only: correct the URL to the resolving path. Not blocking.

## Finding 5 — spot-checks that passed (HIGH confidence, no gap)

- [^ResearchCommand] "`--lanes` (blue candidate drafts, default 3)" — verbatim match in
  `commands/research.md` line 6.
- [^Run2Frontier] "lane 1 took H1 to saturation then breadth; lane 2 took H2..." — verbatim match,
  `research/2026-07-12_memory-architecture/blue/frontier.md` line 11.
- [^MainVsBranch] line counts (blue/report.md=2145, report.md=2972, red/findings.md=695) —
  confirmed via `wc -l`, exact.
- Disconfirming budgets "3/14" and "2/4" for this retrospective's own lane-1/lane-2 — confirmed
  verbatim in `blue/candidates/lane-1.md` and `lane-2.md`.
- [^IsolatedCorrection] title, "2.1-3.4x" token multiplier, "85.5%" modal adoption, "32.3" point
  oracle gap — all confirmed via direct fetch of `arxiv.org/abs/2605.00914`.
- R2-10 §12.5 SingleUserLowRisk description, "preserved distinctly-sourced near-duplicates" and
  other CHANGELOG-R0 merge-vocabulary quotes, catechism template questions 4-7's exact wording,
  `red-auditor.md`'s `memory: project` declaration — all confirmed verbatim against the live repo.
- [^ProvenanceSurvey] title and "typed graph of an agent execution" core definition confirmed;
  the secondary "PROV-AGENT extending W3C PROV" detail unconfirmed at abstract level (not the
  load-bearing part of the claim — low stakes).
- Catechism template now lives on `main` (PR #14 merged) with the exact reframed questions the
  report quotes — content claim is true; only the "unmerged branch" framing is stale (Finding 1).

## Friction

- `Read` tool cannot render PDF pages on this Windows environment: "pdftoppm is not installed.
  Install poppler-utils." This blocked direct-page verification of arXiv:2602.03794 and
  arXiv:2604.18005's full text after `WebFetch`'s own PDF handling returned uninterpreted binary
  stream data for both. I would have used page-level PDF rendering to locate the exact table/page
  carrying (or disproving) the "~19%"/"~95%" and "2-4 agents" figures instead of relying on a
  fetch-tool's lossy markdown conversion, which self-reported inability to parse the compressed
  PDF object streams. This is the same lossy-arXiv-fetch class noted in prior rounds' memory, now
  compounded by the local PDF-rendering path also being unavailable.
- `arxiv.org/html/2604.18005` (no version suffix) 404s while the paper is real and has an abstract
  page — arXiv's HTML mirror does not always exist at the un-versioned path; would have tried
  `/html/2604.18005v1` etc. as a matter of routine had time budget allowed a full sweep on both
  disputed papers within this slice.

## Synopsis (for envelope)

Slice 1 (§0, §1) verification: one HIGH-severity live-source-drift keystone finding (PR #14 merged
mid-debate, falsifying the report's headline and #1 action item's present-tense framing, though not
its underlying analysis); two MEDIUM-LOW confidence quantitative-citation findings (AgentDiversity's
19%/95% figures and DiminishingReturns' "2-4 agents" bound do not trace across four independent
access attempts each); one LOW-severity citation-hygiene note (WisdomCrowds URL). Seven other
citations in this slice verified HIGH confidence, verbatim, against live sources.
