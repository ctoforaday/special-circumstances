# Red audit — round 1, lens 4: logic and completeness

Lens mandate: leaps of faith, missing counterarguments, unexplored alternatives, template
compliance. Full living report re-read in context (all 554 lines of `blue/report.md`), plus
`inputs/{doubts,backlog,run1-friction}.md`, `blue/{frontier.md,CHANGELOG.md,candidates/lane-3.md}`,
and direct verification against the live `special-circumstances` repo (`git log`, `git status`,
`debate.js`, `tests/simulator/debate.test.mjs`, `ideas/backlog.md`) as of this audit.

## GAP-L4-1 — The report's central headline is now stale: PR #14 merged to `main` during/after synthesis [severity HIGH]

- **Location:** §0, "Headline: the fixes exist, but they have not shipped" — *"any run invoked
  against `main` today — including a hypothetical run 3 — still carries R1-HARNESS-1's exact
  defect class and run 2's null-crash class. This reframes 'what should change before run 4' from
  a research question into, first, **a shipping question**."*
- **Direct verification (this audit, `special-circumstances` repo):** `git log --oneline -3` shows
  `main` HEAD = `00018a5` = **"Merge pull request #14 from ctoforaday/feat/feov-dogfood-round-1"**,
  committed `2026-07-13 22:58:54`. The commit blue's own footnote [^MainVsBranch] cites as `main`
  (`9ff0fad`, 22:50:27) is now three commits behind HEAD; PR #14 merged roughly eight minutes after
  that citation. Reading `plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js`
  (the renamed file) on `main` directly confirms the fixes are live: line 33 `JSON.parse` guard,
  line 35 the `topic`/`runDir` unbound-refusal throw, line 136 `if (!blueEnv) throw`, line 171
  `if (!redEnv) throw`, the citation-ledger clause (line 156), per-role `bulk`/`judgment` split
  (lines 38–39), and `tests/simulator/{harness.mjs,debate.test.mjs}` with the 11 named tests all
  present and matching blue's own description of the branch. `ideas/backlog.md` on `main` now
  shows the write-block, simulator, per-role-model, and rename items as `[x]` and *this is no
  longer a false checkbox* — the code backs it.
- **Why this is a logic gap, not just a stale fact:** the report's own epistemic move in §0 was to
  distrust self-reported status ("a backlog checkbox is not a diff") and verify against the
  shipping ref directly — a sound method. But the method was applied once, at a single point in
  time, to a document whose subject (a live git repo, mid-review) was changing under it during the
  same working session (the merge landed ~8 minutes after blue's own cited HEAD, likely while blue
  was still drafting §0–§5). The report never flags its central claim as time-boxed or names a
  recheck-before-final-assembly step, despite the corpus's own H5/§4 teaching exactly this lesson
  about "already fixed" parentheticals going stale (§4: *"the frontier's '(already fixed)'
  parenthetical for the round-1 cluster is false for all three items as of `main`"*) — the report
  applies that lesson backward (to run 2) but not forward (to its own drafting window).
- **Cascading scope — this is not a single-sentence fix.** The premise "unmerged"/"not on `main`"
  loads every one of: §0's entire headline and "Live addendum"; §3 row 1's disposition ("**Do
  first. This is a shipping decision, not a proposal**" — the shipping already happened, so "do
  first" is now backwards); row 2's "PR #14 covers 2 of 4+" (still true, but the framing "verify
  the guards on the shipping ref" now reads as already-partially-done, not prospective); rows 3, 4,
  7, 9, 16 which describe PR #14 content as future work; §4 rank 3's status ("**Fixed on PR #14,
  unmerged** — not safely retired from this ranking until it ships") and rank 4's status ("**Fixed
  on PR #14 (skeleton), unmerged; first live trial = run 3**"); and the shape-verdict correction at
  the end of §4 (*"the write-block fix is unmerged, ENAMETOOLONG has only improvised workarounds,
  and the preflight guard was never on the shipping ref"* — two of those three clauses are now
  false).
- **Grading:** likelihood — certain (directly verified, not inferential); impact — high (the
  report's own top-line thesis and the #1-ranked action item in its graded table are inverted by
  the current state of the world; a reader acting on §3 row 1 today would be told to do something
  already done, and would misjudge urgency on rows 3/4/7/9/16 and §4 ranks 3/4); complexity to fix
  — low (one revision pass: re-verify current `main` HEAD, update §0's headline to "shipped,
  pending its first live trial (run 3)," and thread the status change through the ~6 dependent
  table rows and two friction-ranking entries named above). Not proposing "recheck the repo on a
  timer" as a process fix — that is disproportionate machinery for a gap whose actual fix is "the
  next blue revision round re-verifies before final assembly," which the debate loop already
  provides for free.
- **Disconfirming check performed:** considered whether this is unfalsifiable timing noise (i.e.,
  maybe the merge is coincidental and unrelated to this run) — ruled out: the merge commit message
  explicitly references the same branch (`feat/feov-dogfood-round-1`) and PR (#14) that blue's own
  footnotes [^PR14], [^PR14Description], [^SimulatorTests] discuss at length, and the merge	 landed
  within the same session's timestamp window blue itself cites.

## GAP-L4-2 — "Null at every schema'd call site" overstates a uniform failure mode; only one site currently dereferences unguarded [severity MEDIUM]

- **Location:** §2.3, item 1 — *"**Null at every schema'd call site, not two** — frontier, each
  red-lens, judge, blue-respond, final assembly; each null crashes on a different subsequent
  dereference. 'A suite that only guards the observed site is not founding — it is anecdotal.'"*
- **Direct verification (`debate.js` on `main`, post-merge):** of the five call sites named, only
  **`red-merge`** (guarded, line 171) and **`blue-synthesize`** (guarded, line 136) have their
  return values dereferenced *and* guarded. Of the three claimed additional crash sites:
  - `frontier`'s `agent()` return (line 122) is never assigned to a variable — a null return
    cannot crash on "a subsequent dereference" because there is no dereference; it would silently
    mean `blue/frontier.md` was never written, a *different* failure class (silent missing-file,
    not `TypeError`).
  - each `red-lens` pass's return inside `parallel(...)` (line 162) is likewise discarded, not
    dereferenced; same silent-omission class, not a crash.
  - `judge`'s return (line 181) **is** dereferenced unguarded — `judge.resolutions` at line 184 —
    and this is a real, currently-live `TypeError` risk on `main` today, in the same failure class
    as run 2's `redEnv` crash. This is the one site the claim gets right.
  - `blue-respond`'s reassignment to `blueEnv` (line 194) is not read again until the final return
    statement, which is already guarded (`blueEnv ? blueEnv.claim_count : null`, line 216) — so a
    null there degrades silently rather than crashing.
  - `final assembly`'s return (line 206) is never assigned at all.
- **Why this matters for the lens (leap of faith / overgeneralized claim):** the report treats five
  sites as symmetric ("a different subsequent dereference" implies each one throws) when direct
  reading shows one real crash site (`judge`), one already-fixed pair (`blue-synthesize`,
  `red-merge`), and three silent-degradation sites requiring a *different* kind of test assertion
  (assert the artifact/phase was skipped and the run continues, not assert-throws). The §2.3
  founding-suite recommendation inherits this imprecision: a test suite built to "assert no crash"
  uniformly across all five sites would under-test the `judge` site (needs assert-does-not-throw
  AND assert-continues-safely) and over-test the other three (asserting non-crash proves nothing,
  since they were never going to crash).
- **Grading:** likelihood — high (verified directly against current source); impact — medium (the
  underlying priority — guard every call site — is still correct and cheap; what's wrong is the
  taxonomy used to write the test cases, which would waste a few of the "12 additions" on the wrong
  assertion shape); complexity to fix — low (differentiate the additions list: "assert-throws-then-
  recovers" for `judge`; "assert-silent-degrade-and-continue" for `frontier`/`red-lens`/`assembly`;
  keep `blue-respond` as already covered by the guarded final return).

## GAP-L4-3 — PDF-extraction MCP adoption recommendation doesn't address the supply-chain risk the report itself just raised [severity MEDIUM]

- **Location:** §3, row 13 — *"**PDF full-text/table extraction — adopt, don't build**... off-the-
  shelf MCP servers exist (`arxiv-latex-mcp` for exact LaTeX figures; `pdf-reader-mcp` for tables
  with cell data/confidence)[^PdfMcp] — no bespoke `sc-pdf-extract` Go tool needed."*
- **Missing counterargument:** the same report's §1.1 treats CVE-2026-21852 (a malicious npm
  postinstall script poisoning Claude Code's memory file) as a headline, independently-replicated,
  load-bearing finding about the danger of installing unaudited third-party packages into an agent
  pipeline — *"the memory-poisoning finding was independently triangulated by two agents with
  different search strategies"* is treated as strong signal about supply-chain risk elsewhere in
  the same corpus this retrospective is built from. Row 13 recommends adopting two unaudited,
  small third-party MCP servers into the same pipeline (one giving arbitrary code an entry point
  into every future citation-verification pass) without naming or dismissing that risk class at
  all — no maintainer/provenance check, no mention of running the MCP server sandboxed or pinned,
  no acknowledgment that this is precisely the attack surface the report elsewhere spent a whole
  refinement establishing matters. This is not a request to block adoption — it is a gap in the
  report's own argument: the disposition is graded on cost/complexity only ("Low once scoped as
  adoption"), never on the risk axis its own §1.1 evidence base would apply to any other new
  dependency.
- **Grading:** likelihood — medium (third-party MCP servers with malicious or compromised releases
  are a documented, not hypothetical, class per the report's own CVE citation); impact — medium-
  high if realized (an MCP server sits in the same trust boundary as the agent's tool calls);
  complexity to mitigate — low (one sentence: name the maintainer/provenance check or pin-and-audit
  step before adoption, or explicitly risk-accept with rationale as the report does elsewhere for
  comparable gaps). Recommend closing by addition, not by blocking the adoption itself — forcing a
  full vendor-security review before using two file-parsing MCP servers would be disproportionate
  to the actual risk; naming the omission and a one-line mitigation is proportionate.

## GAP-L4-4 — Cross-provider (architecturally distinct) model diversity is never named as an alternative to lens/method-assignment, despite the report's own cited evidence favoring it [severity LOW-MEDIUM]

- **Location:** §1.1 / §3 row 6 — *"**Engineered per-lane diversity** — assign distinct
  method/source-class lenses..., not persona text and not more headcount"* — disposition: *"Fix
  before run 4, scoped to source-class/method assignment."*
- **The report's own citation contradicts the silent omission:** §1.1 quotes
  [^AgentDiversity] directly: *"same-base-model agents remain more correlated than architecturally
  distinct ones, so lens-diversity is a measured improvement, not a guarantee."* The report
  correctly uses this to caveat its own recommended fix, but never names — to adopt, or to
  explicitly dismiss — the alternative the citation is actually pointing at: assigning at least one
  lane to an architecturally distinct model/provider (not just a distinct prompt/lens on the same
  Claude model family), which its own source says reduces pairwise error correlation more than
  persona/method assignment alone. If this is infeasible given the harness (the `Workflow`/`Task`
  tool dispatches agents through a single model family and `model`/`judgmentModel` only select
  Claude aliases per `debate.js`), the report should say so and dismiss the alternative with that
  reason — it currently never surfaces the option at all, for or against.
- **Grading:** likelihood — medium (the harness constraint plausibly makes this infeasible today,
  which would fully justify not building it — but that justification is never stated); impact —
  low-medium (method/lens assignment is still graded a real improvement on its own terms; this is a
  missing counterargument, not a wrong conclusion); complexity to mitigate — low (one sentence in
  §1.1 or §3 row 6 naming the alternative and why it's out of scope, e.g. harness/tool-access
  constraint, cost of a second provider's API, or deliberate scope limit).

## GAP-L4-5 — The "Live addendum" write-block recurrence is asserted with no artifact evidence, unlike every other friction claim in the corpus [severity LOW-MEDIUM]

- **Location:** §0, "Live addendum" — *"the write-block fired *again* on this very report — the
  synthesizer's Write of `blue/report.md` was refused with the same message recorded in both runs'
  friction..., and a first chunked-heredoc workaround attempt then failed on shell parsing..., forcing
  a third path (scratchpad Write + copy). Third write-block occurrence, third distinct run, at the
  synthesis seat."*
- **Verification attempted, not corroborated:** every other write-block citation in the report
  traces to a concrete artifact — `inputs/run1-friction.md`'s verbatim envelope text, or
  `ideas/backlog.md`'s forensic note (`is_error: True`). This run's own occurrence has no
  corresponding artifact: there is no `friction.md` and no `trajectories/journal.jsonl` for
  `2026-07-12_feov-retrospective` itself (checked directly — the only journal in the run directory
  is `inputs/run1-defect-record/trajectories/journal.jsonl`, which is *input* material from the
  historical run, not this run's own record). The claim is self-reported by the same synthesis pass
  that is making the argument it supports (*"the defect class is alive on the current environment;
  §3 item 8's fix is not speculative hardening"*) — exactly the "operator said so" pattern the
  report itself warns against elsewhere (§0: *"'the operator said so' is not corroboration"*,
  attributed to lane 3). This is not a claim I can disconfirm either — it is plausible and
  consistent with the two prior documented occurrences — but as written it is an uncorroborated
  self-report functioning as load-bearing evidence for a document-wide conclusion, with no footnote
  and no artifact trail, in a report that otherwise holds itself to citing forensic evidence for
  every other instance of the same defect.
- **Grading:** likelihood — high that *something like this* happened (consistent with the pattern);
  corroboration confidence — **low** as currently written (no artifact); impact — low-medium (the
  report's disposition for item 8 doesn't actually change if this specific occurrence is dropped —
  two independently-documented occurrences already establish the pattern); complexity to mitigate —
  low (either add a footnote pointing to whatever transcript/tool-log evidence exists for this
  session, or soften the language to "self-observed, not yet logged to this run's own friction
  file" and recommend §3 item 11's trajectory-capture fix would have prevented this exact
  evidentiary gap — a nice self-demonstrating tie-back the report doesn't currently make).

## GAP-L4-6 — Lane 3's disconfirming-search budget is reported qualitatively while lanes 1 and 2 report a quantified ratio, though all three are presented as equivalently meeting the protocol floor [severity LOW]

- **Location:** report header (lines 1–10) — *"~22 web searches/fetches across the three lanes
  (disconfirming budget met in each lane: 3/14, 2/4, and per-claim source checks respectively)."*
- **Verification:** `blue/candidates/lane-3.md` was read in full for disconfirming-evidence
  accounting; it contains no numeric search-count ratio anywhere (confirmed by direct read and
  targeted search) — only qualitative statements like "every claim below was checked against the
  artifact, not asserted from the doubt's framing" and a per-claim disconfirming caveat on the
  diversity-fix recommendation (§1.1's AgentDiversity caveat). "Per-claim source checks" is a
  different practice than "1 search in 5 devoted to hunting disconfirming evidence" (the protocol's
  actual floor) — checking a claim's source is baseline citation discipline, not the same thing as
  budgeting search effort toward disconfirmation. The header sentence presents the three
  accountings as parallel/equivalent ("respectively") when only two of the three are actually
  checkable against the stated floor.
- **Grading:** likelihood — high (directly verified by absence); impact — low (does not appear to
  change any downstream finding — lane 3's actual content does include disconfirming material,
  e.g. the AgentDiversity caveat and the §0 false-premise catches); complexity to mitigate — low
  (either quantify lane 3's ratio if recoverable from its own search record, or rephrase the header
  to not imply parity of accounting method across lanes).

---

## Round-1 lens-4 synopsis

Six gaps, none disputing the report's substantive conclusions on H1–H5: one HIGH (the central §0
headline and ~8 dependent disposition/status cells are now stale — PR #14 merged to `main` at
`00018a5` during/after this synthesis, verified directly), three MEDIUM (an overstated uniform
null-crash claim across 5 call sites when only 1 currently dereferences unguarded; a missing
supply-chain counterargument against the report's own PDF-MCP adoption recommendation; an unnamed
cross-provider-model alternative its own cited source favors), two LOW-MEDIUM/LOW (an uncorroborated
self-reported live occurrence functioning as load-bearing evidence; a qualitative-vs-quantitative
mismatch in the three lanes' disconfirming-budget accounting). Recommend blue revise §0 and the
dependent table/ranking cells before this goes to the judge — everything else is additive-fix-sized,
not blocking.
