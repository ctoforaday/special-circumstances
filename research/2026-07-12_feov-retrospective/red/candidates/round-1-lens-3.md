# Round 1 — Lens 3 (leaf-node citation verification) — slice 3/3

**Scope of this pass:** sections divided evenly across 3 lens instances; this instance covers
`blue/report.md` **§3 (graded proposal table), §4 (friction ranking), §5 (open questions), and the
Footnotes block**. Sections 0–2 are covered by lens instances 1–2 respectively; findings below are
scoped accordingly, except where a citation shared across sections is load-bearing for a claim
inside this slice, in which case it is verified here regardless of which section first introduced it.

Method: every footnote cited within the assigned slice was followed to its primary source (git/gh
CLI for repo state, WebFetch/WebSearch for arXiv papers and GitHub issues/repos, direct file reads
for the retrospective's own input corpus) and graded for corroboration confidence.

---

## Gap 1 — PR #14 is already MERGED; the report's central "unmerged" framing has drifted stale [HIGH]

**Location:** §3, row 1 — *"**Review and merge PR #14** ... | **Do first. This is a shipping
decision, not a proposal** [L1]"* — and every other row/citation in this slice's scope that
characterizes PR #14 as open/unmerged: §3 row 3 ("Build/extend before run 4"), row 7 ("Fix now"),
row 8(a) ("first live trial = run 3"), row 13, row 14, row 16 ("Ships with #1"), row 17
("template ships with #1"); §4 row 1 ("adoption-not-construction fix available"), row 3 ("Fixed
on PR #14, unmerged"), row 4 ("Fixed on PR #14 (skeleton), unmerged; first live trial = run 3").

**Verification (this session, live):** `gh pr view 14 --json state,mergedAt,baseRefName` against
the actual repo (`ctoforaday/special-circumstances`, confirmed via `git remote -v`) returns
`"state":"MERGED"`, `"mergedAt":"2026-07-14T05:58:54Z"`, merged to `main`. `git log --oneline main`
confirms merge commit `00018a5 Merge pull request #14 from ctoforaday/feat/feov-dogfood-round-1`
is on `main` HEAD, preceded by `203ead5 Merge main into feat/feov-dogfood-round-1` and `9ff0fad`
(the commit the report's §0 cites as `main` HEAD — i.e., `main` has since advanced past the
report's own reference point).

The report's footnote `[^PR14]` states `"gh pr view 14" ... state OPEN ... accessed 2026-07-13`.
That was almost certainly true at access time — this is a genuine **live-source-drift** case
(the exact class the report itself catalogs at §3 row 10 for *external* citations), except here
it hits the report's own load-bearing internal-repo state, and the report has no mechanism to
catch drift in its own primary evidence between the write and the read. Every "unmerged" /
"first live trial = run 3" disposition in this slice is now stale: PR #14 has already shipped,
so §3 row 1's "Do first" is moot (already done), and rows 3/7/8/13/14/16/17 and §4 rows 1/3/4
should be read as "shipped" rather than "pending."

**Secondary finding — the diffstat itself is also wrong, independent of the drift:**
`[^PR14]` cites `+2281/−46`. Live `gh pr view 14 --json additions,deletions,changedFiles,commits`
returns `additions:318, deletions:48, changedFiles:18, commitCount:11`. Deletions are close (46 vs
48 — plausibly 2 more lines removed by a later commit before merge) but additions differ by
~7x (2281 vs 318). This is not fully explained by post-access commits (11 commits, the last being
a merge-main commit, is not enough churn to explain a 1963-line gap) — the original figure looks
miscounted, not merely drifted.

**Confidence:** HIGH that the report's "unmerged" framing is now false (direct `gh`/`git`
verification). MEDIUM-LOW on the additions figure being an original citation error vs. drift —
can't fully rule out an intermediate PR state the report saw that no longer exists.

**Grading:** likelihood certain (already happened) × impact high (undermines the slice's single
highest-priority recommendation, "Do first," and falsifies four+ table-row dispositions) ×
complexity-to-mitigate low (a final freshness pass — re-check `gh pr view`/`git log` immediately
before the report is read/shipped — closes it; this is process, not research, discipline). Not
risk-acceptable as a silent gap: the report should either re-verify PR/branch state at final
assembly time or carry an explicit "state as of <timestamp>, re-verify before acting" caveat next
to any shipping-decision recommendation, since blue's own §0 already knows the corpus's own status
lines can't be trusted without a diff — the same standard should apply reflexively to the report's
own age.

---

## Gap 2 — Footnote `[^AgentDiversity]`'s "~19%"/"~95%" figures do not appear in the cited paper [HIGH-MEDIUM]

**Location:** §3, row 6 — *"High — run 2 measured the convergence directly; external evidence
~19% correlation reduction[^AgentDiversity]"* — and the footnote itself: *"Heterogeneous personas:
~19% lower pairwise error correlation, up to ~95% of independent-ensemble gain; same-base-model
agents remain more correlated than architecturally distinct agents."*

**Verification:** fetched arXiv:2602.03794 ("Understanding Agent Scaling in LLM-Based Multi-Agent
Systems via Diversity") via abstract page and full HTML (`/html/2602.03794v1`), and separately
requested an exhaustive enumeration of every percentage figure appearing anywhere in the paper's
text/tables (34 distinct percentages found, all accuracy figures from Tables 1–3: single-agent
baselines, vote/debate performance at N=2 and across the paper's L1–L4 diversity levels, e.g.
50.8%, 86.5%, 65.34%, 76.86%, 87.5%). **Neither "19%" nor "95%" appears anywhere in the paper**,
and no sentence frames a figure as "pairwise error correlation reduction" or
"independent-ensemble-gain recovery" in those terms.

The paper *does* support the qualitative claim the footnote also makes: "homogeneous agents
produce highly correlated outputs, contributing few effective channels and leading to saturation,"
and it does define an L1 (no diversity) / L2 (persona only) / L3 (model only) / L4 (full) taxonomy
that treats persona diversity as orthogonal to model diversity — consistent with "same-base-model
agents remain more correlated than architecturally distinct agents" as a qualitative reading.

**Confidence:** LOW on the specific "~19%"/"~95%" figures (checked exhaustively across abstract,
full HTML, and a full percentage inventory — three independent negative results). MEDIUM-HIGH on
the qualitative framing (persona vs. model diversity as separable levers; homogeneity correlates
with output convergence).

**Grading:** likelihood high this reduces to a fabricated/hallucinated pair of statistics (three
independent checks came back empty) × impact medium-high (used to underwrite a "High" likelihood
grade for §3 row 6, a top-tier "fix before run 4" recommendation) × complexity to mitigate low
(replace with the qualitative claim only, or find the actual source of "19%"/"95%" if it exists in
a different paper — the report's own literature review already cites four+ diversity papers in
this neighborhood; one of them may be the true source, in which case this is a
footnote-over-attribution error rather than pure fabrication). Not risk-acceptable as-is: false
numeric precision materially inflates confidence in an otherwise-real, more moderate finding.

---

## Gap 3 — Footnote `[^FrictionCount]` miscounts its own source: "35-entry" claimed, 21 actual [MEDIUM]

**Location:** §4, counting-method note — *"Lane 2 counted run-2 rounds strictly by attributed
entry;[^FrictionCount]"* with the footnote reading: *"`research/2026-07-12_feov-retrospective/
inputs/run2-friction.md`, full 35-entry count by role and round (lane 2's strict-attribution
table), accessed 2026-07-13."*

**Verification:** direct read of `run2-friction.md` (23 lines total). Its title line (line 1)
reads *"Friction — run 2 (memory-architecture): 35 agents, 4 rounds, UNVERIFIED / 15 gaps"* — the
figure 35 describes **agents dispatched**, not friction entries. The file contains exactly 21
bulleted friction entries (lines 3–23, one entry per line, each prefixed `role-rN:`). The footnote
conflates the file's stated agent-count header with a claimed "entry count," which is actually 21,
not 35.

**Spot-check of downstream accuracy (unaffected):** despite the miscount, the role/round
attributions in the two rows I could fully cross-check against this file were accurate — §4 row 1
("PDF full-text/table extraction ... red r1–r4, blue r1–r4, judge r2") matches exactly (red:
lines 5/9/15/20 = r1–r4; blue: lines 7/13/19/23 = r1–r4; judge: line 11 only = r2); §4 row 2
("primary security-advisory access ... red r2–r3, blue r2–r3, judge r2") also matches exactly
(red: lines 10/16; blue: lines 14/18; judge: line 12). The undercount is isolated to the summary
statistic, not the table's substantive content.

**Confidence:** HIGH (direct line count against the cited file; row-level content separately
verified HIGH).

**Grading:** likelihood certain (miscounted number, directly checkable) × impact low-medium (the
ranking's substance is intact per the spot-check above; only the summary "35-entry" framing is
wrong) × complexity to mitigate trivial (correct the footnote to "21 entries, header states 35
agents"). Low-priority fix, but a stickler item — the report explicitly makes "friction entries
need structured role+round attribution" one of its own findings (§4 preamble) while its own
citation to the friction corpus miscounts that corpus.

---

## Gap 4 — §4 row 6 persistence claim ("run 2 r1–r2") not fully supported by the cited source [LOW-MEDIUM]

**Location:** §4, row 6 — *"Live-source drift needing access-date deltas | red only | run 2
r1–r2 | Open; protocol-documentation fix, not a tool gap (§3 #10)"*

**Verification:** `run2-friction.md` contains exactly one entry naming live-source drift
explicitly: `red-merge-r1` (line 6: *"Live-source drift (R1-23 mem0 ADD-only pivot, R1-24 star
count, R1-20/R1-21 closed-issue statuses) is only catchable by re-following each citation to the
current primary..."*). The round-2 red-merge entries in this file (lines 9–10) concern
PDF-table-extraction lossiness and CVE-advisory access, not live-source drift. I did not find a
second, r2-scoped live-source-drift friction entry in this file to support "r1–r2."

**Confidence:** MEDIUM — this check is scoped to `run2-friction.md` only; the r2 instance may be
recorded elsewhere in the corpus (`debate.md`, red's `findings.md` for that run) that I did not
cross-check in this pass, so I cannot rule the claim in or out definitively — flagging as
unconfirmed-by-the-cited-source rather than as a confirmed error.

**Grading:** likelihood medium (plausible minor over-count, single-round discrepancy) × impact low
(doesn't change the row's disposition — "Open," "protocol-documentation fix" — either way) ×
complexity to mitigate trivial (re-check against `debate.md`/`red/findings.md` for run 2 and
correct the round range if r2 isn't independently attested). Low priority; noting for completeness
per the stickler mandate rather than blocking.

---

## Gap 5 — §3 row 8's citation of `[^SubagentWriteBug]` is a weaker analogy than the sentence implies [LOW]

**Location:** §3, row 8, complexity column — *"(c) Low edit cost but plugin-wide blast radius,
and unverified — the block may be semantic/role-based, not filename-based (issue #13890 shows
filename-independent subagent write failures[^SubagentWriteBug])"*

**Verification:** GitHub issue #13890 ("[Bug] Subagents unable to write files and call MCP tools
silently," anthropics/claude-code) is real and confirmed via web search: reported on macOS
(Darwin 24.6.0), Claude Code v2.0.68, starting Dec 12–13 2025 — subagent Write/MCP calls **fail
silently** (the subagent believes the write succeeded; nothing happens), with no explicit error
message; the documented workaround is "subagents should return content to the main agent." This
repo's write-block, by contrast, fires with an explicit, worded refusal ("Subagents should return
findings as text, not write report files" — quoted verbatim elsewhere in the report, e.g. §0 and
`[^WriteBlock]`) — a *policy-shaped, worded* block, not a *silent* regression. The two are both
"subagent write failures" but describe different failure signatures (explicit deliberate refusal
vs. silent no-op bug), so citing #13890 as evidence the local block "may not be purely
filename-keyed" is topically adjacent but not a strong structural match — the inference is plausible
but the source doesn't clearly transfer to it.

**Confidence:** MEDIUM — the issue is real and on-topic (both are subagent Task-tool write
failures); the analogy for this specific inference is weaker than presented, but the sentence is
already hedged ("unverified... may be") so it doesn't over-claim beyond what it can support.

**Grading:** likelihood low that this changes any decision (the report already risk-accepts/holds
open question 4 on exactly this uncertainty) × impact low (a citation-precision nit, not a
verdict-bearing one) × complexity to mitigate trivial (soften to "a structurally different
subagent-write-failure class is documented upstream (#13890), suggesting write failures aren't
always filename-keyed — though that issue's silent-failure signature differs from this repo's
worded block"). Low priority.

---

## Verified at HIGH confidence, no gap (for the ledger)

- `[^ChangelogR2]` (§3 row 14, §4 row 2): `research/2026-07-12_memory-architecture/blue/CHANGELOG.md`
  §13.7 language ("R1-8/R2-2 lead docket CLOSED," "double-bind resolved by UNCONDITIONAL
  de-authorized channel") matches the quoted text closely. HIGH.
- `[^LiveBacklog]` (§4 row 1): `ideas/backlog.md` item (c) confirms "TOP TOOL GAP, requested by red,
  blue, AND judge across all 4 rounds: PDF full-text/table extraction" and the tentative
  `sc-pdf-extract` Go-tool-or-MCP candidate the report says §3 row 13 supersedes. HIGH.
- `[^Run1Journal]` (§4 row 3): `journal.jsonl` is 32 lines, 16 distinct agent IDs (16 dispatches,
  matching L1's count exactly via `grep -c "undefined"` = 16); round 1 alone accounts for 10
  distinct invocations (frontier + 3 lanes + blue-synthesize + 3 red-lenses + red-merge + judge =
  10), matching L3's count; round 2's re-attempt (3 red-lenses + merge + judge + final-assembly =
  6) brings the total to 16. Both counts are internally consistent against the same source, exactly
  as the report claims ("both counts are honest against different units"). HIGH.
- `[^ProvenanceSurvey]` (§3 risk-accepted list, row 5 lineage): arXiv:2606.04990 ("From Agent
  Traces to Trust: A Survey of Evidence Tracing and Execution Provenance in LLM Agents") exists,
  matches the "typed activity graph" / PROV-AGENT framing. HIGH.
- `[^PdfMcp]` (§3 row 13): both `takashiishida/arxiv-latex-mcp` and `SylphxAI/pdf-reader-mcp` are
  real, actively maintained GitHub repos matching the claimed capabilities (LaTeX-source figure
  fidelity; table extraction with cell/geometry/confidence data respectively). HIGH.
- `[^DiminishingReturns]` bundle (§3 risk-accepted list): spot-checked 2 of 4 cited sources
  (arXiv:2603.20640, arXiv:2601.19921) — both exist and are topically on-point (diversity-aware
  debate efficiency; confidence/diversity in MAD) though I did not pin the specific "2–4 agents"
  plateau figure to any one of the four bundled sources (a footnote-over-attribution risk, not
  checked to exhaustion given this slice's time budget — flagging as a lower-priority follow-up,
  not a new numbered gap, since no single false number was found as in Gap 2). MEDIUM.

---

## Synopsis (for the round record)

Slice 3/3 (§3 graded proposals, §4 friction ranking, §5 open questions, footnotes) turned up one
HIGH-severity live-drift finding (PR #14 is already merged — the report's lead recommendation and
several dispositions are stale as of this audit) plus a HIGH-MEDIUM fabricated-statistics finding
(the "~19%"/"~95%" diversity figures do not exist in the cited paper), a MEDIUM friction-count
miscitation, and two LOW/LOW-MEDIUM precision nits. Six other footnotes in scope verified HIGH
confidence with no gap.
