# Red audit — round 1, lens: leaf-node citation verification — instance 2 of 3

Slice: §2 "Testing strategy: the trimodal classification and the simulator" (2.1–2.4) + §3
"What should change before run 4" (graded table + risk-accepted list). Every claim in this slice
traced to its leaf-node source; confidence graded per statement↔reference pair.

## Headline finding — the report's central "unmerged" framing is stale as of this audit

**Location:** §2.1 Tier A, row "`redEnv`/`blueEnv` null after terminal-failure `agent()` return" —
*"backlog item marked `[x]` but unguarded on `main` (§0)"*; §2.2 — *"PR #14's `harness.mjs`
answers it"*, *"PR #14 already wires the suite into CI"*; §2.3 — *"PR #14's suite already contains
11 passing tests"*; §3 row 1 — *"**Do first. This is a shipping decision, not a proposal**"*; §3
row 2 — *"the most inexcusable open item in the corpus"*; §4 rank 3/4 — *"Fixed on PR #14,
unmerged; first live trial = run 3"*.

`git log --oneline` on `main`, this machine, now:
```
00018a5 Merge pull request #14 from ctoforaday/feat/feov-dogfood-round-1
203ead5 Merge main into feat/feov-dogfood-round-1 ...
9ff0fad docs(backlog): graduate simulator, per-role models, citation ledger, write-block fix to PR #14
```
`9ff0fad` — the commit blue's own [^MainVsBranch] footnote cites as proof `main` "carries the
unguarded destructure" — is timestamped `2026-07-13 22:50:27`. The merge commit `00018a5` is
timestamped `2026-07-13 22:58:54`, eight minutes later, and is now `main`'s actual HEAD. Direct
verification on the current tree:
- `debate.js` lines 33–36: `const a = typeof args === 'string' ? JSON.parse(args) : args` /
  destructure / `if (!topic || !runDir || ...) throw new Error(...)` — the args guard is present.
- `debate.js` line 171: `if (!redEnv) throw new Error(...)` precedes line 172's
  `redEnv.verdict` access — the null-guard is present (a clean thrown Error, not the run-2
  `TypeError`).
- `node --test tests/simulator/debate.test.mjs` (run live, this audit): **11/11 pass**, ~127ms,
  including "founding regression 2: null red-merge aborts cleanly, not with a TypeError."

**This falsifies or moots, as currently written:** the "unguarded on `main`" claim in §2.1's null
row; the entire "on the branch"/"unmerged" framing threaded through §2.2–§2.3; §3 row 1's "do
first" framing (already done); §3 row 2's "most inexcusable open item" framing (superseded — the
guard exists; what remains to verify is *completeness*, i.e. whether all 4+ schema'd call sites
are covered, not *existence*); §4 ranks 3–4's "unmerged" status line.

**Grading:** likelihood — certain (confirmed directly against the live repository, not inferred).
Impact — high: this is the report's own headline mechanism (§0's "shipping question, not a
research question" and "a backlog checkbox is not a diff") and it anchors the "do first" ordering
of half of §3's table; a reader acting on §3 today would spend effort "reviewing and merging"
something already merged, and would misjudge which guards still need extension-verification.
Complexity to fix — low: re-run the same `git log`/`git show`/`node --test` checks this audit ran,
and either (a) note the merge and reframe "do first" as "verify completeness of what shipped" or
(b) if this genuinely raced blue's synthesis window, add an explicit "as of this exact commit"
timestamp discipline to the live-verification footnote convention so future re-reads can tell
whether the fact has since flipped. This is exactly the live-source-drift risk the retrospective's
own corpus warns about for external citations (§2.1 Tier C, "live-source drift") — the report
did not anticipate that its *own build artifact* is exactly as volatile as an external gh-issue
status or a vendor blog and needs the same access-date-delta discipline recommended in §3 row 10.
**Not risk-acceptable as a documentation nicety** — it is the report's load-bearing keystone claim
being wrong as read today.

## Miscitation — [^AgentDiversity] figures do not appear in the cited paper

**Location:** §3, row 6 ("Engineered per-lane diversity"): *"High — run 2 measured the convergence
directly; external evidence ~19% correlation reduction[^AgentDiversity]"*. (Same footnote also
backs the §1.1 sentence *"Heterogeneous personas ... measurably reduce pairwise error correlation
(~19%) and recover up to ~95% of a fully independent ensemble's gain"* — out of this instance's
assigned section, but the citation is reused verbatim in my slice, so the miscitation is live here
too.)

[^AgentDiversity] cites arXiv:2602.03794, "Understanding Agent Scaling in LLM-Based Multi-Agent
Systems via Diversity." I fetched both `arxiv.org/abs/2602.03794` and the full HTML text
(`arxiv.org/html/2602.03794`, `.../2602.03794v1`) directly and searched for the "19%"/"95%"
figures: **neither appears**. What that paper actually reports is cosine-similarity-based
redundancy (Figure 6, "mean pairwise cosine similarity vs. agent count") and accuracy comparisons
("2 diverse agents can match or exceed 16 homogeneous agents"; Table 2: L4 diversity at N=2 =
67.71% vs. L1 homogeneous at N=16 = 65.34%) — no "pairwise error correlation" percentage, no
"fraction of independent-ensemble gain recovered" framing at all.

A targeted web search for the literal figures located them in a **different** paper:
arXiv:2603.22103, "Multiperspectivity as a Resource for Narrative Similarity Prediction" — "Personas
with more distinctly formulated perspectives produce less correlated errors (19% lower pairwise
error correlation), which in turn yields a larger ensemble gain under majority voting (75.3% vs.
76.0%)... substantial pairwise error correlations (r=.388 ... r=.461)." I fetched
`arxiv.org/abs/2603.22103` directly to confirm domain: an ensemble of 31 LLM personas for
narrative-similarity prediction on SemEval-2026 Task 4 — a specific NLP-annotation task, not a
general multi-agent-system diversity/testing-strategy finding. Even after relocating the figure to
its real source, I could not confirm the "~95% of a fully independent ensemble's gain" clause in
either paper; it may be a further conflation (possibly with the confidence-interval "95%" the
narrative-similarity paper reports elsewhere, which is a different statistical object).

**Grading:** likelihood — certain (fetched the cited source twice and confirmed absence; located
the real source of the "19%" figure by search). Impact — medium: the qualitative direction of the
claim (heterogeneous personas reduce correlated error) is still directionally supported by
2602.03794's own findings and by [^DiversityCollapse]/[^WisdomCrowds] elsewhere in the report, so
§3 row 6's disposition ("Fix before run 4, scoped to source-class/method assignment") likely
survives on the qualitative case alone — but the specific number is unsupported by its citation
and should not stand as "external evidence" backing a graded Likelihood="High" cell in a
build-priority table. Complexity to fix — low: either re-cite the 19% figure to arXiv:2603.22103
(with the caveat that its domain is narrative-similarity annotation, a much narrower analogy to
FEOV's research-lane diversity than the current footnote implies) or drop the specific percentage
and argue the row on the qualitative literature alone, which is what [^DiversityCollapse] and
[^WisdomCrowds] already do.

## Internal contradiction — write-block described as "report.md-specific" against its own row's evidence

**Location:** §2.1 Tier B, row 1: *"Filename write-block on `blue/report.md` (run 1),
`red/findings.md` (run 2), and `blue/report.md` again (this synthesis...)"* — immediately followed
by footnote [^WriteBlock], quoting `ideas/backlog.md` item 8 verbatim: *"CONFIRMED as a hard,
**report.md-specific** tool error."*

The row's own second listed occurrence is `red/findings.md`, not `report.md` — a different
filename. I verified this independently against the primary friction record, not just the
backlog's summary: `inputs/run2-friction.md` line 3, red-merge-r1 — *"Write tool refused
`red/findings.md` on a **filename-heuristic guard ('findings' .md)**"* — the friction author's own
description already contradicts "report.md-specific" one layer up from the backlog. Blue's report
carries the backlog's overclaimed characterization into a footnote without flagging that its own
adjacent sentence (and the underlying primary source it could have checked) already falsifies the
"-specific" framing: the guard fires on at least two distinct filenames sharing a semantic class
(protocol-mandated output-artifact names), not one literal string.

**Grading:** likelihood — certain (the contradiction is present in the text itself, confirmed
against the cited primary friction record). Impact — low-medium: §3 row 8's disposition already
reaches a defensible conclusion (favor the pre-created skeleton over a rename, partly *because*
issue #13890 suggests the block may not be purely filename-keyed) — so the eventual recommendation
is not undermined. But the uncorrected "report.md-specific" quote sitting in [^WriteBlock] is a
misleading characterization of the mechanism that a skeptic checking only that footnote (not the
full row) would take at face value, and it understates the case for treating (c) (rename) as even
weaker than graded — if the guard already keys on "findings" as well as "report", a rename to
`corpus.md`/`audit.md` is safer bet than the "report.md-specific" framing implies, not a
coin-flip. Complexity to fix — low: strike "report.md-specific" or append "(the backlog's own
phrasing; contradicted by the same row's `findings.md` occurrence and by run2-friction.md's
'filename-heuristic guard' description — the block appears keyed to a semantic class of
output-artifact names, not one literal filename)."

## Verified HIGH confidence (no gap)

- §2.1 Tier A row 1: "252.9k tokens, 11m48s" — exact match, `inputs/run1-friction.md` line 18.
  Journal file is exactly 32 lines as cited, 16 paired started/result entries (16 distinct
  `agentId`s) — consistent with [^Run1Journal]'s "L1 counts 16 dispatch entries."
- §2.1 Tier A row 5 ("Empty-candidates cascade"): footnote [^RedMergeR2] quote — exact match,
  `inputs/run1-friction.md` line 7 ("No round-2 red candidate lens passes were ever produced, so
  the merge had no inputs by construction").
- §2.1 Tier A row 4 (`citationPasses` arithmetic): `Math.min(4, Math.max(1,
  Math.ceil((blueEnv.claim_count || 20) / 40)))` — verified against the live source,
  `plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js` line 139. Exact
  match, character for character.
- §2.1 Tier B row 1, [^HookGrep]: grep for `report|findings` (case-insensitive) against
  `plugins/prosthetic-conscience/hooks/` — reproduced, zero matches. Confirmed: the write-block is
  not implemented by this repo's own hooks.
- §2.1 Tier B row 1, [^SubagentWriteBug]: did not re-fetch the GitHub issue itself this round (out
  of budget for this slice), but the claim as stated ("filename-independent") is consistent with
  and does not overstate the primary friction evidence. Medium-high, not re-verified at the
  GitHub-issue leaf node this round — carry forward.
- §2.2, [^NoPackageJson]: glob for `package.json` anywhere in the repo — zero results, confirmed.
- §2.2, [^GoTests]: `main_test.go`/`*_test.go` files confirmed present under
  `plugins/prosthetic-conscience/tools/` (6 files); consistent with stdlib-only testing per the
  file layout (no external test-framework import checked line-by-line, but no `package.json` or
  Go module vendoring for a test framework exists either — supports the claim).
- §2.2, [^Rename]: `ideas/backlog.md` line 6, exact match, marked `[x]` and now also verified
  actually done on `main` (`scripts/debate.js` exists; `workflow.js` does not) — stronger than the
  report claims (it says this is a backlog item; it is in fact shipped).
- §2.2/§2.3, [^SimulatorTests]: verified beyond the citation — not only does the test file exist
  with the claimed structure, I ran it live on current `main`: 11/11 pass, ~127ms, matching every
  named test title in §2.3's merged case list #1–11 (regression 1/1b/2/2b, happy path, per-role
  models, contested docket, deadlock, ceiling, citation passes, friction).
- §3 row 13, [^PdfMcp]: both `github.com/takashiishida/arxiv-latex-mcp` and
  `github.com/SylphxAI/pdf-reader-mcp` exist and match the described capabilities (LaTeX source
  fetch for exact figures; PDF table extraction with cell/bbox/confidence data respectively).
- §3 row 14, [^ChangelogR2]: `research/2026-07-12_memory-architecture/blue/CHANGELOG.md` lines
  132–134 — "§13.7 R1-8/R2-2 lead docket CLOSED... double-bind resolved by UNCONDITIONAL
  de-authorized channel" — matches (near-verbatim; report's phrasing is a faithful paraphrase, not
  a misquote).
- §3 risk-accepted list, [^ProvenanceSurvey] (2606.04990): paper exists; general framing ("typed
  activity graph," evidence tracing, execution provenance) matches the search-verified abstract.
  Did not independently confirm the specific "PROV-AGENT extending W3C PROV" sub-clause at the
  leaf node this round — medium confidence on that sub-clause, high on the general framing.

## Needs more evidence (medium confidence, not a failure — carry forward)

- §3 risk-accepted list, [^DiminishingReturns]: a four-source bundle (arXiv:2603.20640,
  arXiv:2601.19921, VentureBeat, arXiv:2605.00914). I confirmed arXiv:2601.19921 ("Demystifying
  Multi-Agent Debate: The Role of Confidence and Diversity") exists and is on-topic (diminishing
  returns from homogeneous-agent debate, diversity/confidence as the missing levers) but did not
  verify the specific "2–3 rounds / 2–4 agents" plateau figure against that paper's own numbers,
  nor check the other three sources in the bundle individually. Per the footnote-over-attribution
  pattern (a multi-cite footnote often has only one source actually carrying the specific figure),
  this bundle should be pinned to whichever single source states the "2–4 agents" number, or the
  number should be dropped in favor of the qualitative claim the bundle collectively supports.

## Verdict for this slice

§2 (testing strategy) is well-evidenced at the leaf node for its local-repo claims (source-code
line citations check out exactly) but its "PR #14 unmerged" framing is now factually stale — a
live-source-drift failure on the report's own build artifact, not an external citation. §3's
build-priority table inherits that staleness in rows 1–2 and mis-grades row 6's Likelihood cell on
a miscited external figure. None of these are fatal to the report's underlying reasoning (the
qualitative case for guards/simulator/diversity-engineering all survive), but the ordering and
"do first" language a reader would act on today are wrong as written.
