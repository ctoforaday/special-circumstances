# Red audit — round 2, lens 2 (leaf-node citation verification, instance 2 of 3)

Scope per the debate script's own slice instruction: divide the report's sections evenly among
the 3 citation-lens instances and take slice 2. Slice 2 = **§2 "Testing strategy: the trimodal
classification and the simulator" (2.1–2.4) and §3 "What should change before run 4" (the graded
table, rows 1–19, plus the risk-accepted list)**, and every footnote those two sections cite.
Full living `blue/report.md` (704 lines) re-read in context first, not a diff. Ledger consulted
first per protocol: claims already graded HIGH in a prior round and whose section is unchanged
per `blue/CHANGELOG.md` were not re-fetched (citationPasses arithmetic, PdfMcp tool existence,
ChangelogR2 quote, most §2.1 Tier-A internal-source rows, SimulatorTests general existence,
JudgeUnguarded/Reverify47ae48d core claims). Everything below is either newly checked this round
or checked at a level of specificity the ledger did not yet cover.

## GAP-R2L2-1 — MEDIUM confidence: the report's headline cost-comparison figures rest on a
measurement method the project's own live backlog now shows undercounts real spend by ~92%,
and the report has no mechanism to have known this or to flag it

**Location:** §2.3, closing line — quote: *"Entire suite: zero tokens, no network, in-memory
filesystem fakes, well under a second — against historical incident costs of 252.9k (run 1) and
~3M tokens (run 2's quota crashes) [L3]."* Same figures re-anchor §2.4's grading — quote: *"Graded:
simulator = high likelihood x high impact (253k–3M tokens per historical incident) x low
complexity."*

**Verification performed:** live re-fetch of `ideas/backlog.md` at current `main` HEAD (`88eb57f`,
2026-07-14T06:34:58Z — two commits past the `47ae48d` SHA the report's own
[^Reverify47ae48d]/[^PinnedRepoState] discipline pinned; both intervening commits are docs-only,
touching only `ideas/backlog.md`, confirmed via `gh api .../commits/{sha} --jq '.files[].filename'`
for both `47ae48d` and `88eb57f`). The current backlog contains a new item (line 28) not present
in either lane draft or any prior round's citation: *"run cost audit — a tool, not a diet (from
run 3's live measurement; explicitly NOT a call to cheapen judgment): ... FINDINGS so far (run 3,
round 1): the panel token counter excludes cache traffic = 92% of real flow (panel said 610K;
transcripts showed 47.7M); blue-synthesize on the session model was the single priciest agent
($10.58 — cache RATES, not output volume); red ≈ $20/round at this corpus size; full run projects
$80-120 at list rates."*

**Why this matters:** the report's §2/§3/§4 cost argument for the simulator and for de-prioritizing
things like the claim manifest and lane-count floor leans on a specific numeric comparison — "zero
tokens" vs. "252.9k" vs. "~3M" vs. "~50k for smoke" — presented as commensurable magnitudes
supporting urgency ("historical incident costs," "253k–3M tokens per historical incident"). The
live backlog item above is a primary-source admission, from the same project, that its own
in-band token counter (the mechanism presumably informing at least the "~3M" figure, which traces
to a friction self-report, not an audited transcript count) undercounts real spend by roughly an
order of magnitude when cache traffic is excluded. Two things are unverified as a result: (1)
whether 252.9k and ~3M were computed by the same undercounting method as the one the cost-audit
tool just caught, or by a more complete transcript-level count (the report's own footnotes
[^Run1Journal] and the run-2 friction file don't say which); (2) whether the *relative* comparison
between the simulator's real "zero tokens" and the historical figures survives a ~92%
methodology correction applied unevenly (if 3M becomes ~37M under the audited method but 252.9k
was already transcript-derived per [^Run1Journal]'s `journal.jsonl` sourcing, the two "historical
incident" figures cited side-by-side in the same sentence are not apples-to-apples even before
this round's discovery, and are now confirmed to potentially be off by an order of magnitude in
one direction only).

**Does not overturn any recommendation** — the simulator is still free and the historical incidents
still cost real money either way, so the disposition-level argument holds up under either counting
method. This is a precision/citation-freshness gap, not a verdict-level one.

**Grading:** likelihood — high (the undercount is a confirmed, dated, primary-source fact as of
this round, not speculative); impact — medium (weakens the specific numeric framing used
repeatedly across §2.3/§2.4/§4 to argue urgency and priority ordering, without changing which
build items are prioritized); complexity to fix — low (one footnote noting that pre-cost-audit
token/dollar figures throughout §2–§4 are self-reported and likely undercounted, pointing at the
now-existing cost-audit tool, backlog item 28, as the mechanism to obtain a comparable figure
before quoting these numbers again in a future round or run 4's own retrospective).

**Disposition recommended:** raise, don't block — this is a live-source-drift finding of the same
class the report's own §0 Round 1 correction and [^PinnedRepoState] discipline were built to catch,
just recurring one layer further out (the *evidence backing the evidence* moved, not the PR itself).

## GAP-R2L2-2 — LOW confidence, low severity: the `--smoke`-absence claim is narrower than its own
footnote actually checked, and is literally false for one of the two files it names

**Location:** §3, row 4 — quote: *"**[OPEN]** `/research --smoke` mode — confirmed absent: no
`--smoke` flag in `commands/research.md` or `debate.js` as of `main` @ `47ae48d`[^SmokeAbsent]."*

**Verification performed:** direct fetch of both files at current `main` (`88eb57f`; `debate.js`
unchanged since `47ae48d` — only `ideas/backlog.md` changed in the intervening two commits, so the
`47ae48d`-era read still applies). `commands/research.md`: no occurrence of the string `--smoke`
anywhere in the file (argument-hint or body) — this half of the claim holds, confirmed.
`debate.js`: the string `--smoke` **does** appear, in the header comment block (lines 17–18):
*"Behavior needs live agents — but --smoke (1 lane, 1 round, model=haiku) exercises the pipeline
for ~50k tokens."* The footnote [^SmokeAbsent] reads "See [^Reverify47ae48d]", and
[^Reverify47ae48d]'s own verification trail states only *"`commands/research.md` argument-hint and
body text checked for `--smoke` — no match"* — it never claims to have grepped `debate.js` for the
term, yet row 4's prose asserts absence "in `commands/research.md` **or** `debate.js`," which
over-claims the footnote's actual scope and is imprecise about `debate.js` specifically (the
functional claim — no code path parses a `smoke` argument — is still correct; the literal-absence
claim is not).

**Why this is low-severity:** the disposition ("still not built") is unaffected — there is no
parsed `--smoke` flag anywhere, the comment is descriptive/aspirational, not a shipped mode. This
is a footnote-scope/wording precision issue, not a substantive error.

**Grading:** likelihood — medium (a real, checkable wording overreach); impact — low (no
downstream conclusion depends on the literal string's absence from the comment); complexity to fix
— low (reword to "no functional `--smoke` argument-parsing path exists; the term appears only in
`debate.js`'s descriptive header comment, not as parsed input, and not at all in
`commands/research.md`").

## Spot-checks that passed (HIGH confidence, no gap) — new or newly-specific this round

- **Doctrine quote, exact match.** `debate.js` lines 22–24 (live, current `main`): *"Per-role split
  (efficiency doctrine: cheapen redundancy and mechanics, never judgment or the adversary)"* —
  verbatim match to §3 row 16b's quote.
- **Red-lens/red-merge tier routing, exact match.** `debate.js` line 164 dispatches each
  `red-lens-*` pass with `{ ...bulk, ... }`; line 168 dispatches `red-merge` with
  `{ ...judgment, ... }` — confirms row 16b's "routing table sends each red-lens pass ... to the
  cheap bulk tier and only red-merge ... to judgment" exactly, and confirms the named
  doctrine-vs-routing tension is real (this pass ran at bulk tier itself, per the same mechanism).
- **Judge call site still unguarded, reconfirmed at current HEAD.** `debate.js` line 181
  `const judge = await agent(...)`, line 184 `for (const r of judge.resolutions)` — no null check
  between them, on `main` @ `88eb57f` (two commits past the report's `47ae48d` pin; both
  intervening commits are docs-only per direct diff check, so the code claim is unchanged).
- **`citationPasses` still computed once outside the loop, reconfirmed.** Line 139
  `const citationPasses = ...` precedes the `while` loop at line 148; never reassigned inside it —
  matches row 2b exactly, reconfirmed at current HEAD.
- **Lane-count floor and missing `lanes` return field, reconfirmed.** `lanes = 3` default (line 34)
  with no minimum check anywhere in the dispatch (line 128); the `return` object (lines 210–218)
  has no `lanes` key — matches row 7 exactly.
- **Simulator injection mechanism, exact match.** `tests/simulator/harness.mjs`: `const
  AsyncFunction = Object.getPrototypeOf(async function () {}).constructor`, and
  `loadDebateScript` wraps the real file's source in `new AsyncFunction(...)` — confirms §2.2's
  "wrap the real script body in `new AsyncFunction(...)`" claim precisely. `parallel`'s
  `try { out.push(await t()) } catch { out.push(null) }` confirms §2.2/§2.3's "a throwing thunk
  resolves to null rather than rejecting the batch" precisely.
- **11-test count and content, exact match, in order.** `tests/simulator/debate.test.mjs` contains
  exactly 11 `test(...)` blocks, matching §2.3's enumerated list 1–11 in the same order (stringified
  args; unbound-topic refusal; null red-merge; null blue-synthesize; happy path with phase order +
  lane count; per-role model routing; contested docket; deadlock; safety ceiling; citation-pass
  scaling 1↔4; friction attribution).
- **[^AgentTestTiers] figure, exact match.** Direct fetch of the OpenHelm blog post confirms the
  explicit stated rule: *"95% of tests should mock LLMs, 5% use real LLM calls
  (integration/E2E tests)"* — matches the report's "~95%/5% mock-heavy to ~5% real-call split"
  precisely.
- **[^PR14Description] quote, exact match.** `gh pr view 14`'s body text contains, verbatim:
  *"`/research` now pre-creates the blackboard skeleton so subagents only append (dodges the
  harness write-block on fresh report-like files; red's own recommended fix)."*
- **Backlog items cited in §2/§3/§4's footnotes all still present, verbatim, on current `main`**
  despite two further commits to the file since the report's last check: the CRLF `.gitattributes`
  item, the trajectory-capture proposal (still unchecked `[ ]`, wording unchanged), the claim
  manifest item (5) (wording unchanged, "one artifact, five wins"), the blue pre-flight
  self-audit item (4) (wording unchanged), the PDF-extraction "TOP TOOL GAP" item (wording
  unchanged), and the round-scoped-audit item (3) ("human-gated ... trades against the full-re-read
  principle" — matches row 18's framing exactly).
- **PDF-extraction MCP tools' "active maintenance, passing test suites" claim (§3 row 13) — checked
  live, holds up on inspection, not merely asserted.** `takashiishida/arxiv-latex-mcp`: not
  archived, pushed 2026-06-30, 5 open issues, most recent CI runs `success`. `SylphxAI/pdf-reader-mcp`:
  not archived, pushed 2026-07-13 (same day as this check), most recent CI runs on feature/dependabot
  branches show some `failure` conclusions, but every run against the `main`/default branch in the
  visible history is `success` (CI, Deploy Docs, Release, Publish Docker Image, Code Quality all
  green). The feature-branch failures do not undermine the row's claim as worded (scoped to
  adoption from `main`); noted here as a check that was performed, not assumed.

## Friction

- None specific to this slice this round — `gh api` access to the live `special-circumstances`
  repo and its GitHub Actions run history was available and sufficient for every check attempted
  (repo metadata, file contents, commit diffs, workflow-run history). No PDF-rendering or lossy-fetch
  friction recurred in this slice, unlike round 1's lens passes over §1.

## Synopsis (for envelope)

Slice 2 (§2 testing strategy + §3 graded-change table) verification: one MEDIUM-confidence,
medium-impact finding that the report's own headline cost-comparison figures (252.9k / ~3M / ~50k
tokens) rest on a measurement method the project's live backlog now shows undercounts real spend
by ~92% (cache traffic excluded) — a live-source-drift recurrence one layer out from the PR-merge
class round 1 already caught, not verdict-changing but worth a footnote. One LOW-severity wording
finding: the `--smoke`-absent claim's scope exceeds what its own footnote checked and is literally
imprecise about `debate.js`'s comment text, though the functional conclusion is correct. Everything
else checked this round — the doctrine quote, the bulk/judgment routing, the judge-unguarded and
`citationPasses` defects reconfirmed at a HEAD two commits further on, the simulator's injection
mechanism and its 11 tests matching the report's list exactly in order, the AgentTestTiers figure,
the PR14Description quote, every cited backlog item, and the PDF-tool-adoption maintenance claim —
verified HIGH confidence, verbatim or functionally exact, against live current sources.
